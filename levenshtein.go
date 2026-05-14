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

// levenshtein.go implements the Levenshtein edit-distance similarity for the
// fuzzymatch catalogue.
//
// Source: Levenshtein, V. I. (1965). "Binary codes capable of correcting
// deletions, insertions, and reversals." Soviet Physics Doklady,
// 10(8):707-710.
//
// Two-row DP formulation: Wagner, R. A., & Fischer, M. J. (1974). "The
// string-to-string correction problem." Journal of the ACM (JACM),
// 21(1):168-173.
//
// Recurrence (0-indexed, cost = 0 if a[i]==b[j], 1 otherwise):
//
//	D[0, j] = j      (insert j characters)
//	D[i, 0] = i      (delete i characters)
//	D[i, j] = min(
//	             D[i-1, j]   + 1,        // deletion
//	             D[i,   j-1] + 1,        // insertion
//	             D[i-1, j-1] + cost      // substitution
//	           )
//
// Score normalisation: LevenshteinScore(a, b) = 1 - D/max(len(a), len(b)).
// Both-empty → distance 0, score 1.0 exactly. One-empty → distance = max(len),
// score 0.0 exactly. Worst-case time O(m·n), space O(min(m,n)).
//
// Implementation discipline:
//
//   - ASCII fast path operates on bytes directly for inputs whose every byte
//     is strictly less than 0x80 and whose shorter dimension fits within
//     maxStackInputLen; a stack-allocated [(maxStackInputLen+1)*2]int buffer
//     avoids heap allocation.
//   - NO init()-time table builds (per docs/requirements.md §5(12) and
//     .claude/skills/determinism-standards): no var-level side effects.
//   - NO map iteration on output paths (DET-03).
//   - NO transcendental float operations (DET-06): only +, -, *, / and
//     float64() conversions.
//   - NO goroutines, channels, or mutexes (D-09).
//   - Rune variants allocate two []rune slices — documented per Pattern 8.
//   - The 0-alloc budget applies only to the byte path on ASCII ≤ 64 bytes.

package fuzzymatch

// maxStackInputLen is the stack-buffer threshold for two-row DP algorithms.
// When the shorter input dimension n satisfies n <= maxStackInputLen, the two
// DP rows are allocated from a [(maxStackInputLen+1)*2]int stack buffer,
// producing zero heap allocations for ASCII inputs within this bound.
//
// This constant is SHARED across Phase 2 DP-based algorithms (Levenshtein,
// Damerau-Levenshtein OSA, Damerau-Levenshtein Full). Wave 2 plans MUST NOT
// redeclare it — they reference maxStackInputLen defined here.
//
// The value 64 is the initial allocation-budget threshold; benchstat tuning
// may revisit in v1.x if profiling reveals a more optimal threshold.
const maxStackInputLen = 64

// LevenshteinDistance returns the Levenshtein edit distance between a and b —
// the minimum number of single-character insertions, deletions, or
// substitutions required to transform a into b.
//
// Edge cases:
//   - LevenshteinDistance("", "") == 0
//   - LevenshteinDistance("", "abc") == 3
//   - LevenshteinDistance("abc", "abc") == 0
//
// Worst-case time: O(m·n) where m = len(a), n = len(b).
// Space: O(min(m,n)) — two-row DP, no full m×n table allocated.
//
// This function operates on bytes. For multi-byte UTF-8 inputs, use
// LevenshteinDistanceRunes to obtain the rune-aware distance.
func LevenshteinDistance(a, b string) int {
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
	// Ensure b is the shorter string so the inner-loop dimension is minimal.
	if m < n {
		a, b = b, a
		m, n = n, m
	}
	// ASCII fast path: stack-allocate two rows when the shorter dimension
	// fits within the stack-buffer threshold AND both inputs are pure ASCII.
	// The slice headers point into buf; buf itself stays on the stack per Go's
	// escape analysis (confirmed via go build -gcflags="-m=2" — buf does not
	// escape).
	//
	// The isASCII gate preserves the documented "ASCII fast path is reserved
	// for ASCII-only inputs" invariant shared with DamerauLevenshteinOSA and
	// JaroScore (see 02-PATTERNS.md → Pattern: ASCII Fast Path Gate). The byte
	// DP works correctly on any byte content, so this is a style/budgeting
	// invariant, not a correctness one — short non-ASCII inputs go through the
	// heap path to keep the 0-alloc budget documented per algorithm matching
	// "ASCII Short/Medium" only.
	if n <= maxStackInputLen && isASCII(a) && isASCII(b) {
		var buf [(maxStackInputLen + 1) * 2]int
		return levenshteinDP(a, b, m, n, buf[:n+1], buf[n+1:2*(n+1)])
	}
	// Heap path for inputs whose shorter dimension exceeds the threshold or
	// contain non-ASCII bytes.
	return levenshteinDP(a, b, m, n, make([]int, n+1), make([]int, n+1))
}

// LevenshteinDistanceRunes returns the Levenshtein edit distance between a and
// b, treating each string as a sequence of Unicode code points (runes) rather
// than bytes.
//
// This produces correct results for multi-byte UTF-8 strings. For example,
// "café" is 4 runes but 5 bytes; the byte-level distance from "cafe" is 2
// (the é encodes as 2 bytes), whereas the rune-level distance is 1 (one
// substitution: é→e).
//
// The rune variant allocates two []rune slices. For ASCII inputs, prefer
// LevenshteinDistance (zero allocations on inputs ≤ 64 bytes).
func LevenshteinDistanceRunes(a, b string) int {
	ra := []rune(a) // 1 alloc
	rb := []rune(b) // 1 alloc
	return levenshteinDistanceRuneSlices(ra, rb)
}

// LevenshteinScore returns the Levenshtein similarity between a and b as a
// value in [0.0, 1.0], where 1.0 means identical and 0.0 means maximally
// dissimilar.
//
// Normalisation: score = 1 - distance / max(len(a), len(b)).
//
// Edge cases:
//   - LevenshteinScore("", "") == 1.0 exactly (both-empty identity)
//   - LevenshteinScore("abc", "") == 0.0 exactly (one-empty)
//   - LevenshteinScore(a, b) == LevenshteinScore(b, a) (symmetric)
//
// This function operates on bytes. For multi-byte UTF-8 inputs, use
// LevenshteinScoreRunes to obtain the rune-aware similarity.
func LevenshteinScore(a, b string) float64 {
	maxLen := len(a)
	if len(b) > maxLen {
		maxLen = len(b)
	}
	if maxLen == 0 {
		return 1.0 // both-empty → identity; guards NaN (0/0)
	}
	dist := LevenshteinDistance(a, b)
	return 1.0 - float64(dist)/float64(maxLen)
}

// LevenshteinScoreRunes returns the Levenshtein similarity treating a and b as
// sequences of Unicode code points (runes) rather than bytes. The score is in
// [0.0, 1.0], where 1.0 means identical and 0.0 means maximally dissimilar.
//
// Normalisation uses the rune count: score = 1 - dist / max(runeLen(a),
// runeLen(b)). For example, "café" vs "cafe" gives rune distance 1 out of 4
// runes → score 0.75.
//
// The rune variant allocates two []rune slices. For ASCII inputs, prefer
// LevenshteinScore (zero allocations on inputs ≤ 64 bytes).
func LevenshteinScoreRunes(a, b string) float64 {
	ra := []rune(a) // 1 alloc
	rb := []rune(b) // 1 alloc
	maxLen := len(ra)
	if len(rb) > maxLen {
		maxLen = len(rb)
	}
	if maxLen == 0 {
		return 1.0 // both-empty → identity; guards NaN (0/0)
	}
	dist := levenshteinDistanceRuneSlices(ra, rb)
	return 1.0 - float64(dist)/float64(maxLen)
}

// levenshteinDP is the inner two-row DP kernel for byte-level Levenshtein.
// prev and curr must each have length n+1.
//
// The kernel operates on string bytes directly via a[i-1] / b[j-1] indexing —
// no string-to-byte-slice conversion, zero allocation.
//
// After the loop, the answer is in prev[n] (rows are swapped at the end of
// each outer iteration, so after the last swap the row that was curr becomes
// prev).
func levenshteinDP(a, b string, m, n int, prev, curr []int) int {
	// Initialise the previous row: D[0,j] = j (cost of inserting j characters).
	for j := 0; j <= n; j++ {
		prev[j] = j
	}
	for i := 1; i <= m; i++ {
		curr[0] = i // D[i,0] = i (cost of deleting i characters)
		for j := 1; j <= n; j++ {
			cost := 1
			if a[i-1] == b[j-1] {
				cost = 0
			}
			// Three-way minimum (inlined for performance — avoids function call overhead).
			v := prev[j] + 1 // deletion
			if w := curr[j-1] + 1; w < v {
				v = w // insertion
			}
			if w := prev[j-1] + cost; w < v {
				v = w // substitution
			}
			curr[j] = v
		}
		prev, curr = curr, prev // swap: prev becomes the current completed row
	}
	return prev[n] // after the final swap, the answer is in prev[n]
}

// levenshteinDistanceRuneSlices computes the Levenshtein distance between two
// rune slices using the same two-row DP algorithm as the byte variant. It is
// called by LevenshteinDistanceRunes and LevenshteinScoreRunes after the
// []rune(string) conversion.
func levenshteinDistanceRuneSlices(ra, rb []rune) int {
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
			v := prev[j] + 1
			if w := curr[j-1] + 1; w < v {
				v = w
			}
			if w := prev[j-1] + cost; w < v {
				v = w
			}
			curr[j] = v
		}
		prev, curr = curr, prev
	}
	return prev[n]
}
