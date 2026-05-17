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

// damerau_osa.go implements the Damerau-Levenshtein Optimal String Alignment
// (OSA) distance similarity for the fuzzymatch catalogue.
//
// Source: Boytsov, L. (2011). "Indexing methods for approximate dictionary
// searching: comparative analysis." ACM Journal of Experimental Algorithmics,
// 16, Article 1. (OSA formulation cited from §3.1.)
//
// Historical source: Damerau, F. J. (1964). "A technique for computer detection
// and correction of spelling errors." Communications of the ACM, 7(3):171-176.
// (Original transposition paper; does not name "OSA" explicitly.)
//
// Recurrence (0-indexed, cost = 0 if a[i-1]==b[j-1], else 1):
//
//	D[0, j] = j      (insert j characters)
//	D[i, 0] = i      (delete i characters)
//	D[i, j] = min(
//	             D[i-1, j]   + 1,           // deletion
//	             D[i,   j-1] + 1,           // insertion
//	             D[i-1, j-1] + cost,        // substitution
//	             D[i-2, j-2] + 1            // transposition
//	             (when i>=2 && j>=2 && a[i-1]==b[j-2] && a[i-2]==b[j-1])
//	           )
//
// OSA constraint: each substring participates in at most one
// transposition. After a transposition, the affected characters
// cannot be edited again. This makes DL-OSA NOT a strict metric —
// triangle inequality may fail on contrived inputs. Use
// DamerauLevenshteinFull for the metric variant.
//
// Discriminating vector: DamerauLevenshteinOSADistance("ca", "abc") == 3,
// while DamerauLevenshteinFull returns 2 for the same pair. This is the
// locked gate that distinguishes the two variants.
//
// Score normalisation: DamerauLevenshteinOSAScore(a, b) = 1 - D/max(len(a), len(b)).
// Both-empty → distance 0, score 1.0 exactly. One-empty → score 0.0 exactly.
// Worst-case time O(m·n), space O(n) via three-row rolling DP.
//
// Implementation discipline:
//
//   - ASCII fast path operates on bytes directly for inputs whose every byte
//     is strictly less than 0x80 and whose shorter dimension fits within
//     maxStackInputLen; a stack-allocated [(maxStackInputLen+1)*3]int buffer
//     (1560 bytes) avoids heap allocation.
//   - maxStackInputLen is defined in levenshtein.go — do NOT redeclare here.
//   - NO init()-time table builds (per docs/requirements.md §5(12) and
//     .claude/skills/determinism-standards): no var-level side effects.
//   - NO map iteration on output paths (DET-03).
//   - NO transcendental float operations (DET-06): only +, -, *, / and
//     float64() conversions. No math.Pow, math.Log, math.Exp, math.Sqrt, math.FMA.
//   - NO goroutines, channels, or mutexes (D-09).
//   - NO []byte(string) conversion on the hot path — bytes are accessed
//     directly via a[i-1] and b[j-1] string indexing.
//   - Rune variants allocate two []rune slices — documented per Pattern 8.
//   - The 0-alloc budget applies only to the byte path on ASCII ≤ 64 bytes.

package fuzzymatch

// DamerauLevenshteinOSADistance returns the Damerau-Levenshtein OSA
// (Optimal String Alignment) edit distance between a and b — the minimum
// number of single-character insertions, deletions, substitutions, or
// adjacent transpositions required to transform a into b, with the
// restriction that each substring participates in at most one transposition.
//
// The OSA variant is NOT a strict metric: the triangle inequality may fail
// on contrived inputs (Boytsov 2011 §3.1). For the true metric variant, use
// DamerauLevenshteinFull.
//
// Discriminating vector: DamerauLevenshteinOSADistance("ca", "abc") == 3
// (NOT 2 — that is DamerauLevenshteinFull's value for the same pair).
//
// Edge cases:
//   - DamerauLevenshteinOSADistance("", "") == 0
//   - DamerauLevenshteinOSADistance("", "abc") == 3
//   - DamerauLevenshteinOSADistance("abc", "abc") == 0
//   - DamerauLevenshteinOSADistance("ab", "ba") == 1 (single transposition)
//
// Worst-case time: O(m·n) where m = len(a), n = len(b).
// Space: O(n) — three-row rolling DP, no full m×n table.
//
// This function operates on bytes. For multi-byte UTF-8 inputs, use
// DamerauLevenshteinOSADistanceRunes to obtain the rune-aware distance.
func DamerauLevenshteinOSADistance(a, b string) int {
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
	// Ensure b is the shorter string so the inner-loop dimension n is minimal.
	if m < n {
		a, b = b, a
		m, n = n, m
	}
	// ASCII fast path: stack-allocate three rows when the shorter dimension
	// fits within the stack-buffer threshold. The three slice headers point
	// into buf; buf itself stays on the stack per Go's escape analysis.
	// buf size = (maxStackInputLen+1)*3 = 65*3 = 195 ints = 1560 bytes.
	if n <= maxStackInputLen && isASCII(a) && isASCII(b) {
		var buf [(maxStackInputLen + 1) * 3]int
		prevprev := buf[:n+1]
		prev := buf[n+1 : 2*(n+1)]
		curr := buf[2*(n+1) : 3*(n+1)]
		return damerauOSADP(a, b, m, n, prevprev, prev, curr)
	}
	// Heap path for inputs whose shorter dimension exceeds the threshold
	// or contain non-ASCII bytes.
	prevprev := make([]int, n+1)
	prev := make([]int, n+1)
	curr := make([]int, n+1)
	return damerauOSADP(a, b, m, n, prevprev, prev, curr)
}

// DamerauLevenshteinOSADistanceRunes returns the Damerau-Levenshtein OSA
// distance between a and b, treating each string as a sequence of Unicode
// code points (runes) rather than bytes.
//
// This produces correct results for multi-byte UTF-8 strings. For ASCII
// inputs, prefer DamerauLevenshteinOSADistance (zero allocations on inputs ≤ 64 bytes).
func DamerauLevenshteinOSADistanceRunes(a, b string) int {
	if a == b {
		return 0 // fast identity — saves two []rune allocations
	}
	ra := []rune(a) // 1 alloc
	rb := []rune(b) // 1 alloc
	return damerauOSADistanceRuneSlices(ra, rb)
}

// DamerauLevenshteinOSAScore returns the Damerau-Levenshtein OSA similarity
// between a and b as a value in [0.0, 1.0], where 1.0 means identical and
// 0.0 means maximally dissimilar.
//
// For programmatic input-quality checks before scoring,
// see [fuzzymatch.Validate].
//
// Normalisation: score = 1 - distance / max(len(a), len(b)).
//
// Edge cases:
//   - DamerauLevenshteinOSAScore("", "") == 1.0 exactly (both-empty identity)
//   - DamerauLevenshteinOSAScore("abc", "") == 0.0 exactly (one-empty)
//   - DamerauLevenshteinOSAScore(a, b) == DamerauLevenshteinOSAScore(b, a) (symmetric)
//   - DamerauLevenshteinOSAScore("ca", "abc") == 0.0 exactly (discriminating vector)
//
// This function operates on bytes. For multi-byte UTF-8 inputs, use
// DamerauLevenshteinOSAScoreRunes to obtain the rune-aware similarity.
func DamerauLevenshteinOSAScore(a, b string) float64 {
	maxLen := len(a)
	if len(b) > maxLen {
		maxLen = len(b)
	}
	if maxLen == 0 {
		return 1.0 // both-empty → identity; guards NaN (0/0) and -0 (Inf/Inf)
	}
	dist := DamerauLevenshteinOSADistance(a, b)
	return 1.0 - float64(dist)/float64(maxLen)
}

// DamerauLevenshteinOSAScoreRunes returns the Damerau-Levenshtein OSA
// similarity treating a and b as sequences of Unicode code points (runes)
// rather than bytes. The score is in [0.0, 1.0], where 1.0 means identical
// and 0.0 means maximally dissimilar.
//
// Normalisation uses the rune count: score = 1 - dist / max(runeLen(a),
// runeLen(b)). For example, "café" vs "cafe" gives rune distance 1 out of 4
// runes → score 0.75.
//
// The rune variant allocates two []rune slices. For ASCII inputs, prefer
// DamerauLevenshteinOSAScore (zero allocations on inputs ≤ 64 bytes).
func DamerauLevenshteinOSAScoreRunes(a, b string) float64 {
	if a == b {
		return 1.0 // fast identity — covers both-empty and identical inputs without []rune alloc
	}
	ra := []rune(a) // 1 alloc
	rb := []rune(b) // 1 alloc
	maxLen := len(ra)
	if len(rb) > maxLen {
		maxLen = len(rb)
	}
	if maxLen == 0 {
		return 1.0 // both-empty → identity; guards NaN (0/0)
	}
	dist := damerauOSADistanceRuneSlices(ra, rb)
	return 1.0 - float64(dist)/float64(maxLen)
}

// damerauOSADP is the inner three-row rolling DP kernel for byte-level
// Damerau-Levenshtein OSA. prevprev, prev, and curr must each have length n+1.
//
// The kernel operates on string bytes directly via a[i-1] / b[j-1] indexing —
// no string-to-byte-slice conversion, zero allocation on the byte path.
//
// Three-row rolling pattern: after computing each row i, the rows are
// rotated: prevprev, prev, curr = prev, curr, prevprev. The slice that
// was prevprev (oldest row) becomes curr and is fully overwritten on the
// next iteration before any reads occur.
//
// After the outer loop completes, the answer is in prev[n]. This is because
// the final rotation moves the last computed row from curr into prev.
//
// The high cyclomatic complexity is inherent to the OSA recurrence: deletion,
// insertion, substitution, and transposition each require a branch, and the
// three-way minimum of the first three requires two conditional comparisons.
// Extracting sub-functions would obscure the recurrence and hurt inlining.
func damerauOSADP(a, b string, m, n int, prevprev, prev, curr []int) int { //nolint:gocyclo // OSA DP kernel — four-operation recurrence is inherently complex; see godoc above
	// Initialise prev (row 0): D[0, j] = j (cost of inserting j characters).
	// prevprev is initialised to zeros (Go zero-value) — row "-1" is notional
	// and only accessed when i>=2 in the transposition check.
	for j := 0; j <= n; j++ {
		prev[j] = j
	}
	for i := 1; i <= m; i++ {
		curr[0] = i // D[i, 0] = i (cost of deleting i characters)
		for j := 1; j <= n; j++ {
			cost := 1
			if a[i-1] == b[j-1] {
				cost = 0
			}
			// Three-way minimum: deletion, insertion, substitution.
			v := prev[j] + 1 // deletion: D[i-1, j] + 1
			if w := curr[j-1] + 1; w < v {
				v = w // insertion: D[i, j-1] + 1
			}
			if w := prev[j-1] + cost; w < v {
				v = w // substitution: D[i-1, j-1] + cost
			}
			// Transposition: D[i-2, j-2] + 1
			// Available when i >= 2, j >= 2, and adjacent characters are swapped.
			if i >= 2 && j >= 2 && a[i-1] == b[j-2] && a[i-2] == b[j-1] {
				if w := prevprev[j-2] + 1; w < v {
					v = w
				}
			}
			curr[j] = v
		}
		// Three-way rotate: oldest row (prevprev) is recycled as the new curr.
		// After this: the row we just computed (curr) becomes prev,
		// the previous prev becomes prevprev,
		// and the old prevprev is handed back as the new curr (to be overwritten).
		prevprev, prev, curr = prev, curr, prevprev
	}
	// After the final rotation, the last computed row is in prev.
	return prev[n]
}

// damerauOSADistanceRuneSlices computes the DL-OSA distance between two rune
// slices using the same three-row rolling DP algorithm as the byte variant.
// It is called by DamerauLevenshteinOSADistanceRunes and
// DamerauLevenshteinOSAScoreRunes after the []rune(string) conversion.
//
// The high cyclomatic complexity mirrors damerauOSADP — the OSA four-operation
// recurrence (deletion, insertion, substitution, transposition) is inherently
// branchy. The rune variant is a separate function rather than a generic to
// avoid the allocation overhead of interface boxing on the hot path.
func damerauOSADistanceRuneSlices(ra, rb []rune) int { //nolint:gocyclo // OSA DP kernel — four-operation recurrence mirrors damerauOSADP; see godoc above
	if len(ra) == 0 {
		return len(rb)
	}
	if len(rb) == 0 {
		return len(ra)
	}
	m, n := len(ra), len(rb)
	// Ensure rb is the shorter slice for the inner-loop dimension.
	if m < n {
		ra, rb = rb, ra
		m, n = n, m
	}
	prevprev := make([]int, n+1)
	prev := make([]int, n+1)
	curr := make([]int, n+1)
	for j := 0; j <= n; j++ {
		prev[j] = j
	}
	for i := 1; i <= m; i++ {
		curr[0] = i
		for j := 1; j <= n; j++ {
			cost := 1
			if ra[i-1] == rb[j-1] {
				cost = 0
			}
			v := prev[j] + 1 // deletion
			if w := curr[j-1] + 1; w < v {
				v = w // insertion
			}
			if w := prev[j-1] + cost; w < v {
				v = w // substitution
			}
			if i >= 2 && j >= 2 && ra[i-1] == rb[j-2] && ra[i-2] == rb[j-1] {
				if w := prevprev[j-2] + 1; w < v {
					v = w // transposition
				}
			}
			curr[j] = v
		}
		prevprev, prev, curr = prev, curr, prevprev
	}
	return prev[n]
}
