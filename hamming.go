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

// hamming.go implements the Hamming distance/score similarity for the
// fuzzymatch catalogue.
//
// Source: Hamming, R. W. (1950). "Error detecting and error correcting
// codes." Bell System Technical Journal, 29(2):147-160.
//
// The Hamming distance between two equal-length strings is the number of
// positions at which the corresponding characters (bytes, for the byte
// variant; runes, for the rune variant) differ. In Hamming's original
// formulation, the inputs must be equal length — the metric is defined
// only for strings of the same length.
//
// Inputs of unequal length are not an error: HammingDistance returns
// max(len(a), len(b)) and HammingScore returns 0.0. Callers wanting
// strict Hamming-1950 equal-length semantics should length-check
// before calling.
//
// Score normalisation: HammingScore(a, b) = 1 - distance / max(len(a), len(b)).
// Both-empty → distance 0, score 1.0 exactly. Unequal length → distance
// max(len(a), len(b)), score 0.0 exactly (the normalisation resolves cleanly:
// 1 - max(len)/max(len) = 0.0).
//
// Implementation discipline:
//
//   - Single O(min(m,n)) counting loop — no DP table, no allocation.
//   - NO init()-time side effects (per docs/requirements.md §5(12) and
//     .claude/skills/determinism-standards): dispatch registered via var init.
//   - NO map iteration on output paths (DET-03).
//   - NO transcendental float operations (DET-06): only +, -, *, / and
//     float64() conversions.
//   - NO goroutines, channels, or mutexes (D-09).
//   - NO []byte(s) conversion — byte indexing into string is zero-allocation.
//   - Rune variants allocate two []rune slices — documented per Pattern 8.
//   - The 0-alloc budget applies only to the byte path (any length — Hamming
//     is a single loop with no DP buffer).

package fuzzymatch

// HammingDistance returns the Hamming distance between a and b — the number
// of byte positions at which the two strings differ.
//
// Inputs of unequal length are not an error: HammingDistance returns
// max(len(a), len(b)). This is the project-wide convention for handling
// length-mismatched inputs (see file-level godoc for the rationale). Callers
// wanting strict Hamming-1950 equal-length semantics should length-check
// before calling.
//
// Edge cases:
//   - HammingDistance("", "") == 0
//   - HammingDistance("abc", "abc") == 0
//   - HammingDistance("abc", "ab") == 3 (max(3, 2) per the unequal-length policy)
//
// This function operates on bytes. For multi-byte UTF-8 inputs, use
// HammingDistanceRunes to obtain the rune-aware distance.
func HammingDistance(a, b string) int {
	if a == b {
		return 0 // fast identity (covers both-empty and identical inputs)
	}
	m, n := len(a), len(b)
	if m != n {
		// Unequal-length policy: return max(len(a), len(b)) per CONTEXT.md.
		// This makes HammingScore normalise cleanly to 0.0.
		if m > n {
			return m
		}
		return n
	}
	// Equal-length: count byte-position mismatches.
	var dist int
	for i := 0; i < m; i++ {
		if a[i] != b[i] {
			dist++
		}
	}
	return dist
}

// HammingDistanceRunes returns the Hamming distance between a and b, treating
// each string as a sequence of Unicode code points (runes) rather than bytes.
//
// This produces correct results for multi-byte UTF-8 strings. For example,
// "café" is 4 runes but 5 bytes; "cafè" is also 4 runes — the rune-level
// distance is 1 (the final rune differs), whereas the byte-level distance
// would count the UTF-8 byte sequences individually.
//
// Inputs of unequal rune-count are not an error: HammingDistanceRunes returns
// max(runeLen(a), runeLen(b)) per the same unequal-length convention.
//
// The rune variant allocates two []rune slices. For ASCII inputs, prefer
// HammingDistance (zero allocations at any length).
func HammingDistanceRunes(a, b string) int {
	if a == b {
		return 0 // fast identity — saves two []rune allocations
	}
	ra := []rune(a) // 1 alloc
	rb := []rune(b) // 1 alloc
	m, n := len(ra), len(rb)
	if m != n {
		if m > n {
			return m
		}
		return n
	}
	var dist int
	for i := 0; i < m; i++ {
		if ra[i] != rb[i] {
			dist++
		}
	}
	return dist
}

// HammingScore returns the Hamming similarity between a and b as a value in
// [0.0, 1.0], where 1.0 means identical and 0.0 means maximally dissimilar.
//
// For programmatic input-quality checks before scoring (including detection
// of the unequal-length condition that triggers the silent-zero return),
// see [fuzzymatch.Validate] — it emits WarnUnequalLength scoped to
// AlgoHamming.
//
// Normalisation: score = 1 - distance / max(len(a), len(b)).
//
// Inputs of unequal length are not an error: HammingScore returns 0.0
// silently. Callers wanting strict Hamming-1950 equal-length semantics should
// length-check before calling (see file-level godoc for the full rationale).
//
// Edge cases:
//   - HammingScore("", "") == 1.0 exactly (both-empty identity)
//   - HammingScore("abc", "abc") == 1.0 exactly (identical strings)
//   - HammingScore("abc", "ab") == 0.0 exactly (unequal-length: score is 0.0,
//     while the underlying HammingDistance returns max(len)=3, not 0)
//   - HammingScore(a, b) == HammingScore(b, a) (symmetric)
//
// This function operates on bytes. For multi-byte UTF-8 inputs, use
// HammingScoreRunes to obtain the rune-aware similarity.
func HammingScore(a, b string) float64 {
	maxLen := len(a)
	if len(b) > maxLen {
		maxLen = len(b)
	}
	if maxLen == 0 {
		return 1.0 // both-empty → identity; guards NaN (0/0)
	}
	dist := HammingDistance(a, b)
	return 1.0 - float64(dist)/float64(maxLen)
}

// HammingScoreRunes returns the Hamming similarity treating a and b as
// sequences of Unicode code points (runes) rather than bytes. The score is
// in [0.0, 1.0], where 1.0 means identical and 0.0 means maximally
// dissimilar.
//
// Normalisation uses the rune count: score = 1 - dist / max(runeLen(a),
// runeLen(b)).
//
// Inputs of unequal rune-count are not an error: HammingScoreRunes returns
// 0.0 silently per the same unequal-length convention as HammingScore.
//
// The rune variant allocates two []rune slices. For ASCII inputs, prefer
// HammingScore (zero allocations at any length).
func HammingScoreRunes(a, b string) float64 {
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
	m, n := len(ra), len(rb)
	if m != n {
		// Unequal rune-count: distance = max(m, n), score = 0.0.
		return 0.0
	}
	var dist int
	for i := 0; i < m; i++ {
		if ra[i] != rb[i] {
			dist++
		}
	}
	return 1.0 - float64(dist)/float64(maxLen)
}
