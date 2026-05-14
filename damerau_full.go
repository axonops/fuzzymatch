// Copyright 2026 AxonOps Limited
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

// damerau_full.go implements the Damerau-Levenshtein Full (Lowrance-Wagner)
// edit-distance similarity for the fuzzymatch catalogue.
//
// Source: Lowrance, R., Wagner, R. A. (1975). "An extension of the
// string-to-string correction problem." Journal of the ACM, 22(2):177-183.
//
// DL-Full (the Lowrance-Wagner formulation) is the TRUE metric variant
// of Damerau-Levenshtein distance. It permits unrestricted transpositions:
// any pair of adjacent characters may be transposed, and the characters
// may subsequently be edited. Compare with DamerauLevenshteinOSA (the
// Optimal String Alignment variant), which restricts each substring to
// at most one transposition and is NOT a metric.
//
// Triangle inequality holds for DL-Full unconditionally. Use this
// variant when correctness > speed and metric properties matter.
//
// Recurrence (1-indexed; Lowrance-Wagner 1975, adapted for 0-based strings):
//
//	Initialisation:
//	  D[-1, j] = ∞         (phantom sentinel row; stores large value)
//	  D[ i,-1] = ∞         (phantom sentinel column)
//	  D[ 0, j] = j         (cost of inserting j characters)
//	  D[ i, 0] = i         (cost of deleting i characters)
//	  da[c]   = 0          (last row where character c was seen; 0 = never seen)
//
//	For i = 1..m, j = 1..n:
//	  cost = 0 if a[i-1] == b[j-1], else 1
//	  l = da[b[j-1]]   // last row where b[j-1] appeared
//	  k = db            // last col in this row where a[i-1] appeared
//	  D[i,j] = min(
//	              D[i-1, j-1] + cost,                               // substitution
//	              D[i,   j-1] + 1,                                  // insertion
//	              D[i-1, j  ] + 1,                                  // deletion
//	              D[l-1, k-1] + (i-l-1) + 1 + (j-k-1)             // transposition
//	          )
//	  if a[i-1] == b[j-1]: db = j     // update last column match for a[i-1]
//	After row i: da[a[i-1]] = i       // update last row where a[i-1] was seen
//
// The transposition term D[l-1, k-1] + (i-l-1) + 1 + (j-k-1) represents:
// the edit distance up to the anchor positions (l-1, k-1), plus the cost
// of deleting the intervening characters of a (i-l-1 deletions), plus the
// transposition itself (cost 1), plus the cost of deleting the intervening
// characters of b (j-k-1 insertions).
//
// Because D[l-1, k-1] may reference rows far outside a two-row window, this
// implementation uses a heap-allocated full (m+2)×(n+2) DP table for all
// inputs. The two-row + auxiliary-table optimisation is deferred to a v1.x
// performance follow-up (see SUMMARY for details).
//
// Score normalisation: DamerauLevenshteinFullScore(a, b) = 1 - D/max(len(a), len(b)).
// Both-empty → distance 0, score 1.0 exactly. One-empty → distance = max(len),
// score 0.0 exactly. Worst-case time O(m·n), space O(m·n) — full DP table.
//
// Implementation discipline:
//
//   - The full (m+2)×(n+2) DP table is heap-allocated for all inputs. The
//     two-row + auxiliary-anchor-table optimisation (which would achieve 0
//     allocations on ASCII inputs ≤ 64 bytes) is a v1.x follow-up.
//   - da array (ASCII path: [256]int) maps each byte to the last row index
//     where it appeared. Map LOOKUP only: da[a[i-1]] is a point lookup, never
//     iterated to produce output (DET-03). The rune path uses map[rune]int —
//     also lookup-only.
//   - NO init()-time table builds (per docs/requirements.md §5(12)).
//   - NO map iteration on output paths (DET-03). Map lookups (da[c]) are
//     fine — the prohibition is on range-over-map for output construction.
//   - NO transcendental float operations (DET-06): only +, -, *, / and
//     float64() conversions.
//   - NO goroutines, channels, or mutexes (D-09).
//   - NO []byte(string) conversion on the byte hot path — bytes are accessed
//     directly via a[i-1] and b[j-1] string indexing.
//   - Rune variants eagerly convert via []rune(a), []rune(b) (two allocations).
//   - The 0-alloc budget is NOT met by this v1.0 implementation — the full DP
//     table allocates O(m·n) ints. This is documented and tracked for v1.x.

package fuzzymatch

// DamerauLevenshteinFullDistance returns the Damerau-Levenshtein Full
// (Lowrance-Wagner 1975) edit distance between a and b — the minimum
// number of single-character insertions, deletions, substitutions, or
// adjacent transpositions (with no restriction on subsequent editing)
// required to transform a into b.
//
// This is the TRUE metric variant of Damerau-Levenshtein distance. The
// triangle inequality holds unconditionally. Compare with
// DamerauLevenshteinOSADistance, which restricts each substring to at
// most one transposition and is NOT a metric.
//
// Discriminating vector:
//   - DamerauLevenshteinFullDistance("ca", "abc") == 2  (Full DL)
//   - DamerauLevenshteinOSADistance("ca", "abc") == 3   (OSA — different!)
//
// Edge cases:
//   - DamerauLevenshteinFullDistance("", "") == 0
//   - DamerauLevenshteinFullDistance("", "abc") == 3
//   - DamerauLevenshteinFullDistance("abc", "abc") == 0
//   - DamerauLevenshteinFullDistance("ab", "ba") == 1 (single transposition)
//
// Worst-case time: O(m·n) where m = len(a), n = len(b).
// Space: O(m·n) — full DP table (v1.0 implementation; see file godoc for
// v1.x two-row optimisation plan).
//
// This function operates on bytes. For multi-byte UTF-8 inputs, use
// DamerauLevenshteinFullDistanceRunes to obtain the rune-aware distance.
func DamerauLevenshteinFullDistance(a, b string) int {
	if a == b {
		return 0 // fast identity (covers both-empty and identical inputs)
	}
	m, n := len(a), len(b)
	if m == 0 {
		return n
	}
	if n == 0 {
		return m
	}
	return damerauFullDP(a, b, m, n)
}

// DamerauLevenshteinFullDistanceRunes returns the Damerau-Levenshtein Full
// (Lowrance-Wagner 1975) distance between a and b, treating each string as
// a sequence of Unicode code points (runes) rather than bytes.
//
// This produces correct results for multi-byte UTF-8 strings. For ASCII
// inputs, prefer DamerauLevenshteinFullDistance.
//
// The rune variant allocates two []rune slices. The auxiliary last-occurrence
// table uses a map[rune]int (heap allocation — unavoidable for Unicode). The
// map is QUERIED via point lookup only, never iterated to produce output (DET-03).
func DamerauLevenshteinFullDistanceRunes(a, b string) int {
	ra := []rune(a) // 1 alloc
	rb := []rune(b) // 1 alloc
	return damerauFullDistanceRuneSlices(ra, rb)
}

// DamerauLevenshteinFullScore returns the Damerau-Levenshtein Full
// (Lowrance-Wagner 1975) similarity between a and b as a value in [0.0, 1.0],
// where 1.0 means identical and 0.0 means maximally dissimilar.
//
// Normalisation: score = 1 - distance / max(len(a), len(b)).
//
// Edge cases:
//   - DamerauLevenshteinFullScore("", "") == 1.0 exactly (both-empty identity)
//   - DamerauLevenshteinFullScore("abc", "") == 0.0 exactly (one-empty)
//   - DamerauLevenshteinFullScore(a, b) == DamerauLevenshteinFullScore(b, a) (symmetric)
//   - DamerauLevenshteinFullScore("ca", "abc") ≈ 0.3333 (1 - 2/3; discriminating vector)
//
// This function operates on bytes. For multi-byte UTF-8 inputs, use
// DamerauLevenshteinFullScoreRunes to obtain the rune-aware similarity.
func DamerauLevenshteinFullScore(a, b string) float64 {
	maxLen := len(a)
	if len(b) > maxLen {
		maxLen = len(b)
	}
	if maxLen == 0 {
		return 1.0 // both-empty → identity; guards NaN (0/0) and -0 (Inf/Inf)
	}
	dist := DamerauLevenshteinFullDistance(a, b)
	return 1.0 - float64(dist)/float64(maxLen)
}

// DamerauLevenshteinFullScoreRunes returns the Damerau-Levenshtein Full
// (Lowrance-Wagner 1975) similarity treating a and b as sequences of Unicode
// code points (runes) rather than bytes. The score is in [0.0, 1.0], where
// 1.0 means identical and 0.0 means maximally dissimilar.
//
// Normalisation uses the rune count: score = 1 - dist / max(runeLen(a),
// runeLen(b)). For example, "café" vs "cafe" gives rune distance 1 out of 4
// runes → score 0.75.
//
// The rune variant allocates two []rune slices and a map[rune]int auxiliary
// last-occurrence table. The map is QUERIED via point lookup only, never
// iterated to produce output (DET-03). For ASCII inputs, prefer
// DamerauLevenshteinFullScore.
func DamerauLevenshteinFullScoreRunes(a, b string) float64 {
	ra := []rune(a) // 1 alloc
	rb := []rune(b) // 1 alloc
	maxLen := len(ra)
	if len(rb) > maxLen {
		maxLen = len(rb)
	}
	if maxLen == 0 {
		return 1.0 // both-empty → identity; guards NaN (0/0)
	}
	dist := damerauFullDistanceRuneSlices(ra, rb)
	return 1.0 - float64(dist)/float64(maxLen)
}

// damerauFullDP is the Lowrance-Wagner 1975 full DP kernel for byte-level
// Damerau-Levenshtein Full distance.
//
// The algorithm uses a full (m+2)×(n+2) DP table with a phantom sentinel
// row/column at index 0 (corresponding to "row/col -1" in the paper's
// 1-based indexing). The sentinel value bigVal = m+n ensures that any
// transposition term involving a "never-seen" character position is large
// enough to never be selected as the minimum.
//
// da[c] tracks the last row index (1-based in the DP table, 1..m) where
// byte c was last seen. da[c] == 0 means "never seen in a". This is
// READ-ONLY on the output path — da is written only at the end of each
// outer row (da[a[i-1]] = i), and READ only for the transposition lookup
// (l = da[b[j-1]]). No map iteration (DET-03).
//
// High cyclomatic complexity is inherent to the Lowrance-Wagner recurrence:
// substitution, insertion, deletion, and transposition each require a
// branch, and the transposition term involves two additional range guards
// (l > 0 && k > 0). Extraction would obscure the recurrence.
func damerauFullDP(a, b string, m, n int) int { //nolint:gocyclo // Lowrance-Wagner recurrence — four-operation DP with transposition guards; see godoc
	// bigVal is large enough to prevent phantom-sentinel transpositions from
	// being selected: any transposition involving a "never seen" row (da[c]==0)
	// or column (db==0) will produce a cost ≥ bigVal, which exceeds any real
	// edit distance ≤ m+n.
	bigVal := m + n

	// Allocate the full (m+2)×(n+2) DP table as a flat slice for cache
	// locality. Index [i][j] is at d[i*(n+2)+j].
	// Row 0 is the phantom sentinel row (D[-1,*] in the paper).
	// Column 0 is the phantom sentinel column (D[*,-1] in the paper).
	// Row 1 corresponds to D[0,*] (initial costs for 0 characters of a processed).
	// Column 1 corresponds to D[*,0].
	size := (m + 2) * (n + 2)
	d := make([]int, size)
	stride := n + 2

	// Initialise phantom sentinel row (row 0): all bigVal.
	for j := 0; j <= n+1; j++ {
		d[0*stride+j] = bigVal
	}
	// Initialise phantom sentinel column (col 0): all bigVal.
	for i := 0; i <= m+1; i++ {
		d[i*stride+0] = bigVal
	}
	// Initialise row 1 (D[0,*]): D[0,j] = j (cost of inserting j characters).
	// In the table, row index 1 ↔ D[0,*], column index j+1 ↔ D[*,j].
	for j := 0; j <= n; j++ {
		d[1*stride+(j+1)] = j
	}
	// Initialise col 1 (D[*,0]): D[i,0] = i (cost of deleting i characters).
	for i := 0; i <= m; i++ {
		d[(i+1)*stride+1] = i
	}

	// da[c] = last row (1-indexed in the paper, translated to table row i+1)
	// where byte c appeared in a. 0 means "never seen".
	// READ-ONLY on output path: only written at end of each outer iteration,
	// read only for the transposition lookup. No map iteration (DET-03).
	var da [256]int // stack-allocated auxiliary last-occurrence array (2048 bytes)

	for i := 1; i <= m; i++ {
		// db = last column (1-indexed in the paper, i.e. j value 1..n) in this
		// row where a[i-1] matched a character of b. 0 means "no match yet
		// in this row".
		db := 0

		for j := 1; j <= n; j++ {
			// l = last row where b[j-1] appeared in a (0 = never seen).
			l := da[b[j-1]] // map LOOKUP only — not iteration (DET-03)
			// k = last column in this row (i's row) where a[i-1] matched b[*].
			k := db

			cost := 1
			if a[i-1] == b[j-1] {
				cost = 0
				db = j // record this column as the last match for a[i-1]
			}

			// Table indices: paper D[i,j] → table d[(i+1)*stride+(j+1)].
			// Paper D[i-1,j-1] → table d[i*stride+j].
			// Paper D[i-1,j]   → table d[i*stride+(j+1)].
			// Paper D[i,j-1]   → table d[(i+1)*stride+j].
			// Paper D[l-1,k-1] → table d[l*stride+k]  (l,k are 1-based paper values;
			//                    l-1 paper row → l table row; k-1 paper col → k table col).

			// Three standard Levenshtein options.
			sub := d[i*stride+j] + cost         // D[i-1,j-1] + cost (substitution)
			ins := d[(i+1)*stride+j] + 1         // D[i,j-1] + 1 (insertion)
			del := d[i*stride+(j+1)] + 1         // D[i-1,j] + 1 (deletion)

			v := sub
			if ins < v {
				v = ins
			}
			if del < v {
				v = del
			}

			// Transposition option (Lowrance-Wagner 1975).
			// Only valid when both l > 0 and k > 0 (i.e. both characters have
			// been seen before in their respective sequences).
			if l > 0 && k > 0 {
				// trans = D[l-1, k-1] + (i-l-1) + 1 + (j-k-1)
				// In table coordinates: D[l-1,k-1] → d[l*stride+k].
				// (i-l-1) = deletions of intervening a-chars (between row l and row i)
				// 1 = the transposition itself
				// (j-k-1) = deletions of intervening b-chars (between col k and col j)
				trans := d[l*stride+k] + (i - l - 1) + 1 + (j - k - 1)
				if trans < v {
					v = trans
				}
			}

			d[(i+1)*stride+(j+1)] = v
		}

		// After processing row i, record that a[i-1] was last seen at row i.
		// da is only written here; lookups happen only in the j-loop above.
		da[a[i-1]] = i
	}

	// The answer is D[m,n] → table d[(m+1)*stride+(n+1)].
	return d[(m+1)*stride+(n+1)]
}

// damerauFullDistanceRuneSlices computes the Damerau-Levenshtein Full
// (Lowrance-Wagner 1975) distance between two rune slices. It is called by
// DamerauLevenshteinFullDistanceRunes and DamerauLevenshteinFullScoreRunes
// after the []rune(string) conversion.
//
// The auxiliary last-occurrence table uses map[rune]int (heap allocation).
// The map is QUERIED via point lookup only (da[r]), never iterated to produce
// output (DET-03). This is why the rune variant is not 0-alloc.
//
// High cyclomatic complexity mirrors damerauFullDP — the Lowrance-Wagner
// four-operation recurrence (deletion, insertion, substitution, transposition)
// is inherently branchy.
func damerauFullDistanceRuneSlices(ra, rb []rune) int { //nolint:gocyclo // Lowrance-Wagner recurrence mirrors damerauFullDP; see godoc
	if len(ra) == 0 {
		return len(rb)
	}
	if len(rb) == 0 {
		return len(ra)
	}
	m, n := len(ra), len(rb)
	bigVal := m + n

	size := (m + 2) * (n + 2)
	d := make([]int, size)
	stride := n + 2

	// Phantom sentinel row (row 0): all bigVal.
	for j := 0; j <= n+1; j++ {
		d[0*stride+j] = bigVal
	}
	// Phantom sentinel column (col 0): all bigVal.
	for i := 0; i <= m+1; i++ {
		d[i*stride+0] = bigVal
	}
	// Row 1: D[0,j] = j.
	for j := 0; j <= n; j++ {
		d[1*stride+(j+1)] = j
	}
	// Col 1: D[i,0] = i.
	for i := 0; i <= m; i++ {
		d[(i+1)*stride+1] = i
	}

	// da[r] = last row (1-based) where rune r appeared in ra.
	// Map LOOKUP only — not iterated to produce output (DET-03).
	da := make(map[rune]int)

	for i := 1; i <= m; i++ {
		db := 0
		for j := 1; j <= n; j++ {
			l := da[rb[j-1]] // map LOOKUP only (DET-03)
			k := db

			cost := 1
			if ra[i-1] == rb[j-1] {
				cost = 0
				db = j
			}

			sub := d[i*stride+j] + cost
			ins := d[(i+1)*stride+j] + 1
			del := d[i*stride+(j+1)] + 1

			v := sub
			if ins < v {
				v = ins
			}
			if del < v {
				v = del
			}

			if l > 0 && k > 0 {
				trans := d[l*stride+k] + (i - l - 1) + 1 + (j - k - 1)
				if trans < v {
					v = trans
				}
			}

			d[(i+1)*stride+(j+1)] = v
		}

		da[ra[i-1]] = i // map WRITE only at end of row (DET-03)
	}

	return d[(m+1)*stride+(n+1)]
}
