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

// jaro.go implements the Jaro similarity for the fuzzymatch catalogue.
//
// Source: Jaro, M. A. (1989). "Advances in record-linkage methodology as
// applied to matching the 1985 census of Tampa, Florida." Journal of the
// American Statistical Association, 84(406):414-420.
//
// Supporting reference for canonical reference vectors: Winkler, W. E.
// (1990). "String comparator metrics and enhanced decision rules in the
// Fellegi-Sunter model of record linkage." Proceedings of the Survey
// Research Methods Section of the American Statistical Association,
// pp. 354-359.
//
// # Formula
//
// Given two strings s1 (length la) and s2 (length lb):
//
//  1. Compute the matching window:
//     w = floor(max(la, lb) / 2) - 1   (clamped to >= 0)
//
//  2. First pass — matching: for each position i in s1, scan the window
//     [max(0, i-w), min(lb, i+w+1)] in s2 for an unmatched character
//     equal to s1[i]. Mark both matchA[i] and matchB[j] true. Count m
//     (total matches).
//
//  3. If m == 0, return 0.0.
//
//  4. Second pass — transpositions: walk positions matched in s1 (in
//     order), walk positions matched in s2 (in order), count mismatches
//     between the two character sequences, then halve to get t.
//
//  5. Jaro score (left-to-right, explicit parenthesisation per DET-06):
//     J = (m/la + m/lb + (m-t)/m) / 3.0
//
// # Reference Vector (Winkler 1990)
//
// MARTHA vs MARHTA:
//   - w = floor(6/2) - 1 = 2
//   - m = 6 (all six characters matched)
//   - Matched s1 order: M A R T H A (positions 0,1,2,3,4,5)
//   - Matched s2 order: M A R H T A (positions 0,1,2,4,3,5 in s2)
//   - Transpositions: T≠H, H≠T → 2 mismatches → t = 1
//   - J = (6/6 + 6/6 + 5/6) / 3 = (1 + 1 + 0.8333…) / 3 = 0.9444…
//
// # Jaro is NOT a metric
//
// Jaro is NOT a metric: the triangle inequality does not hold for
// Jaro similarity. Callers reasoning about distances should use
// Levenshtein, Damerau-Levenshtein OSA, or Damerau-Levenshtein Full
// instead.
//
// Because Jaro is not a metric, no JaroDistance function is provided —
// the formula yields a similarity in [0.0, 1.0] directly with no
// distance-to-similarity inversion.
//
// # Implementation discipline
//
//   - ASCII fast path uses stack-allocated [256]bool × 2 match-flag arrays
//     for inputs where both strings are ≤ maxJaroStackLen bytes and both
//     are pure ASCII (isASCII from normalise.go). maxJaroStackLen = 256 is
//     distinct from maxStackInputLen = 64 (the DP threshold in levenshtein.go)
//     because Jaro's match-flag arrays are O(n) booleans, not O(n) ints — 256
//     booleans fit on the stack with negligible cost, whereas the DP buffer
//     must be kept smaller to avoid excessive stack pressure.
//   - NO init()-time table builds (per docs/requirements.md §5(12) and
//     .claude/skills/determinism-standards): dispatch registered via var init.
//   - NO map iteration on output paths (DET-03).
//   - NO transcendental float operations (DET-06): only +, -, * and / with
//     float64() conversions. No math.Pow, math.Log, math.Exp, math.Sqrt,
//     or math.FMA.
//   - NO goroutines, channels, or mutexes (D-09).
//   - NO []byte(s) conversion — byte indexing into string is zero-allocation.
//   - Rune variants allocate two []rune slices — documented per Pattern 8.
//   - The 0-alloc budget applies only to the ASCII byte path on inputs ≤
//     maxJaroStackLen bytes.
//   - Worst-case time complexity: O(la · w) ≈ O(la · lb). See threat model
//     T-02-03-01 in the plan for DoS analysis.

package fuzzymatch

// maxJaroStackLen is the stack-buffer threshold for the Jaro match-flag arrays.
// When both inputs satisfy len(a) <= maxJaroStackLen and len(b) <= maxJaroStackLen
// AND both are pure ASCII (as determined by isASCII from normalise.go), the two
// [256]bool match-flag arrays are allocated on the stack, producing zero heap
// allocations.
//
// This constant is distinct from maxStackInputLen (= 64, owned by levenshtein.go
// for DP-table algorithms). Jaro's match-flag arrays are booleans rather than
// ints: 256 booleans costs 256 bytes of stack, which is well within the Go
// goroutine stack budget and avoids escaping to the heap.
//
// Wave 2 plans 02-04 (JaroWinkler) through 02-06 that build on JaroScore
// MUST NOT redeclare maxJaroStackLen — they reference it here.
const maxJaroStackLen = 256

// JaroScore returns the Jaro similarity between a and b as a value in
// [0.0, 1.0], where 1.0 means identical and 0.0 means maximally dissimilar
// (or one of the two strings is empty while the other is non-empty).
//
// For programmatic input-quality checks before scoring,
// see [fuzzymatch.Validate].
//
// Edge cases:
//   - JaroScore("", "") == 1.0 exactly (both-empty identity convention)
//   - JaroScore("", "abc") == 0.0 exactly (one-empty)
//   - JaroScore("abc", "abc") == 1.0 exactly (identical strings)
//   - JaroScore(a, b) == JaroScore(b, a) (symmetric)
//
// Reference vectors (Jaro 1989 / Winkler 1990):
//   - JaroScore("MARTHA", "MARHTA") ≈ 0.9444
//   - JaroScore("DIXON", "DICKSONX") ≈ 0.7667
//   - JaroScore("JELLYFISH", "SMELLYFISH") ≈ 0.8963
//
// This function operates on bytes. For multi-byte UTF-8 inputs, use
// JaroScoreRunes to obtain the rune-aware similarity.
//
// Time complexity: O(la · w) where w = floor(max(la,lb)/2) - 1. For equal-
// length inputs this is effectively O(la²/2). Space: O(la + lb) for match
// flags (stack-allocated for ASCII ≤ 256 bytes; heap for longer inputs).
func JaroScore(a, b string) float64 {
	if a == b {
		return 1.0 // fast identity — covers both-empty and identical inputs
	}
	la, lb := len(a), len(b)
	if la == 0 || lb == 0 {
		return 0.0 // one-empty → zero similarity
	}

	// Compute matching window: w = max(la,lb)/2 - 1, clamped to >= 0.
	maxLen := la
	if lb > maxLen {
		maxLen = lb
	}
	w := maxLen/2 - 1
	if w < 0 {
		w = 0
	}

	// ASCII fast path: use stack-allocated [256]bool arrays when both inputs
	// fit within maxJaroStackLen bytes and are pure ASCII.
	if la <= maxJaroStackLen && lb <= maxJaroStackLen && isASCII(a) && isASCII(b) {
		var matchA [maxJaroStackLen]bool
		var matchB [maxJaroStackLen]bool
		return jaroBytes(a, b, la, lb, w, matchA[:la], matchB[:lb])
	}

	// Heap path for inputs exceeding maxJaroStackLen or containing non-ASCII.
	return jaroBytes(a, b, la, lb, w, make([]bool, la), make([]bool, lb))
}

// JaroScoreRunes returns the Jaro similarity between a and b, treating each
// string as a sequence of Unicode code points (runes) rather than bytes.
//
// This produces correct results for multi-byte UTF-8 strings. For example,
// "café" and "cafe" differ in one rune (é vs e); the rune-level Jaro
// similarity reflects the rune-level matching window, while the byte-level
// similarity would treat the 2-byte UTF-8 encoding of é as two separate
// bytes.
//
// The rune variant allocates two []rune slices (2 allocs/op). For ASCII
// inputs, prefer JaroScore (zero allocations on inputs ≤ 256 bytes).
func JaroScoreRunes(a, b string) float64 {
	if a == b {
		return 1.0 // fast identity — covers both-empty and identical inputs without []rune alloc
	}
	ra := []rune(a) // 1 alloc
	rb := []rune(b) // 1 alloc
	return jaroRunes(ra, rb)
}

// jaroBytes is the inner kernel for byte-level Jaro similarity.
// matchA and matchB are pre-allocated slices of length la and lb respectively.
// They must be zero-initialised before the call (stack vars are zero; heap
// slices from make are zero).
//
// The cyclomatic complexity is structurally mandated by the Jaro algorithm:
// two nested loops (matching pass and transposition pass), each with bounds
// clamping and conditional branching. No refactoring can reduce the decision
// count without harming readability or correctness.
func jaroBytes(a, b string, la, lb, w int, matchA, matchB []bool) float64 { //nolint:gocyclo // Jaro match-flag algorithm is inherently complex — see godoc above
	// First pass: find matches.
	m := 0
	for i := 0; i < la; i++ {
		// Compute the window bounds for position i in b.
		lo := i - w
		if lo < 0 {
			lo = 0
		}
		hi := i + w + 1
		if hi > lb {
			hi = lb
		}
		for j := lo; j < hi; j++ {
			if !matchB[j] && a[i] == b[j] {
				matchA[i] = true
				matchB[j] = true
				m++
				break
			}
		}
	}

	if m == 0 {
		return 0.0 // division guard — prevents 0/0 = NaN on the (m-t)/m term
	}

	// Second pass: count transpositions.
	// Walk matched positions in s1 and matched positions in s2 in order;
	// count positions where the matched characters differ; halve to get t.
	t := 0
	j := 0
	for i := 0; i < la; i++ {
		if !matchA[i] {
			continue
		}
		// Advance j to the next matched position in b.
		for j < lb && !matchB[j] {
			j++
		}
		if j < lb {
			if a[i] != b[j] {
				t++
			}
			j++
		}
	}
	t /= 2 // Jaro transposition count = (mismatched matched-pairs) / 2 — Jaro 1989 canonical T/2 halving

	// Three-term Jaro formula with explicit parenthesisation and left-to-right
	// float reduction (determinism-standards DET-06; docs/requirements.md §13).
	fm := float64(m)
	return (fm/float64(la) + fm/float64(lb) + float64(m-t)/fm) / 3.0
}

// jaroRunes is the inner kernel for rune-level Jaro similarity.
// The cyclomatic complexity mirrors jaroBytes — see its godoc for the rationale.
func jaroRunes(ra, rb []rune) float64 { //nolint:gocyclo // Jaro match-flag algorithm is inherently complex — mirrors jaroBytes
	if len(ra) == 0 && len(rb) == 0 {
		return 1.0 // both-empty identity
	}
	la, lb := len(ra), len(rb)
	if la == 0 || lb == 0 {
		return 0.0
	}
	if runeSlicesEqual(ra, rb) {
		return 1.0 // identity fast path after rune conversion
	}

	maxLen := la
	if lb > maxLen {
		maxLen = lb
	}
	w := maxLen/2 - 1
	if w < 0 {
		w = 0
	}

	matchA := make([]bool, la)
	matchB := make([]bool, lb)

	m := 0
	for i := 0; i < la; i++ {
		lo := i - w
		if lo < 0 {
			lo = 0
		}
		hi := i + w + 1
		if hi > lb {
			hi = lb
		}
		for j := lo; j < hi; j++ {
			if !matchB[j] && ra[i] == rb[j] {
				matchA[i] = true
				matchB[j] = true
				m++
				break
			}
		}
	}

	if m == 0 {
		return 0.0
	}

	t := 0
	j := 0
	for i := 0; i < la; i++ {
		if !matchA[i] {
			continue
		}
		for j < lb && !matchB[j] {
			j++
		}
		if j < lb {
			if ra[i] != rb[j] {
				t++
			}
			j++
		}
	}
	t /= 2 // Jaro transposition count = (mismatched matched-pairs) / 2 — Jaro 1989 canonical T/2 halving

	fm := float64(m)
	return (fm/float64(la) + fm/float64(lb) + float64(m-t)/fm) / 3.0
}

// runeSlicesEqual reports whether two rune slices are element-wise equal.
// Extracted to keep jaroRunes below the gocyclo threshold.
func runeSlicesEqual(a, b []rune) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}
