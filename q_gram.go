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

// q_gram.go provides the unexported q-gram extraction helpers shared by
// Q-Gram Jaccard (this plan), Sørensen-Dice (plan 05-02), Cosine
// (plan 05-03), and Tversky (plan 05-04). The public surface for the
// q-gram tier is exactly the four algorithms' Score / ScoreRunes
// functions — these helpers are package-internal per CONTEXT.md §2
// LOCKED.
//
// Source: Ukkonen, E. (1992). "Approximate string-matching with q-grams
// and maximal matches." Theoretical Computer Science 92(1):191-211,
// §2 (q-gram definition) and §3 (set / multiset Jaccard formulation
// over q-gram counts).
//
// Algorithm — extractQGrams(s, n):
//
//   1. Reject n < 1 and len(s) < n by returning a non-nil empty
//      map[string]int{}. This frees the algorithm-level callers from
//      writing a separate guard for the degenerate input shapes; the
//      direct-call panic on n < 1 is performed by the algorithm
//      function itself per CONTEXT.md §5.
//   2. Allocate the result map with a capacity hint sized to the
//      maximum possible distinct q-gram count: len(s)-n+1 for the byte
//      path, runeCount-n+1 for the rune path. This avoids one or two
//      map-rehash allocations on medium-to-long inputs (RESEARCH.md §4.2).
//   3. Walk the byte / rune slice with a sliding window of length n,
//      incrementing the count for each window. Multiset semantics:
//      repeated q-grams accumulate (e.g. extractQGrams("AAAA", 2) ==
//      {"AA": 3}).
//
// Determinism (DET-03):
//
//   The returned map MUST NOT be iterated by callers on any output
//   path. Every consumer of these helpers either (a) reads only map
//   lengths and per-key lookups (the multiset cardinality and
//   intersection/difference computations in Q-Gram Jaccard,
//   Sørensen-Dice, and Tversky), or (b) extracts the keys and sorts
//   them before iteration (the dot-product loop in Cosine — landing in
//   plan 05-03). Map-iteration discipline is the load-bearing
//   determinism gate for the q-gram tier.
//
// Source-origin discipline (per algorithm-licensing-standards):
//
//   - Primary source:        Ukkonen 1992 §2-3 (TCS 92(1):191-211).
//   - Cross-validation:      none — hand-derived RV-J1..RV-J4
//                            reference vectors in qgram_jaccard_test.go.
//   - Tie-break:             none (q-gram extraction is unambiguous;
//                            sliding-window left-to-right is the
//                            canonical Ukkonen 1992 enumeration).
//   - GPL/LGPL provenance:   none.
//   - Code copied verbatim:  none.
//
// Implementation discipline:
//
//   - NO init()-time table builds (per docs/requirements.md §5(12)).
//   - NO map iteration on output paths (DET-03; see godoc above).
//   - NO transcendental float operations (DET-06): these helpers
//     return integer-valued maps; floats enter only at the algorithm
//     normalisation step.
//   - NO goroutines, channels, or mutexes.
//   - Empty inputs produce empty (non-nil) maps; callers MAY rely on
//     the result being non-nil for safe range-iteration in test code.
//   - The substring s[i:i+n] is a slice header into the input string
//     — no per-key string copy on the heap; only the map's internal
//     hash bookkeeping allocates.
//
// Allocation budget (RESEARCH.md §4.1, docs/requirements.md §14.1):
//
//   - Two map allocations per algorithm call (one per side).
//   - Up to ~2 additional map-growth allocations on medium-to-long
//     inputs (capacity hint mitigates this).
//   - No stack-buffer fast path (RESEARCH.md §4.3 — the map allocation
//     dominates regardless; reducing peripheral allocations would not
//     change the dominant cost).

package fuzzymatch

// extractQGrams returns the multiset of overlapping length-n BYTE
// q-grams of s. Returns a non-nil empty map[string]int{} when n < 1,
// when s is empty, or when len(s) < n.
//
// Multiset semantics: repeated q-grams accumulate; for example,
// extractQGrams("AGCTAGCT", 2) returns {"AG": 2, "GC": 2, "CT": 2,
// "TA": 1} (the four distinct keys with total multiset cardinality 7,
// matching Ukkonen 1992 §3's worked example).
//
// The returned map MUST NOT be iterated by callers on any output path
// (DET-03). Callers compute intersection / union / difference
// cardinalities by reading map lengths and per-key counts; the
// dot-product / sorted-iteration variants extract and sort keys
// explicitly (Cosine, plan 05-03).
func extractQGrams(s string, n int) map[string]int {
	if n < 1 || len(s) < n {
		return map[string]int{}
	}
	// Capacity hint: at most len(s)-n+1 distinct q-grams (when every
	// window contributes a new key). Sizing the map up front avoids
	// one or two rehash allocations on medium-to-long inputs.
	m := make(map[string]int, len(s)-n+1)
	for i := 0; i <= len(s)-n; i++ {
		// s[i:i+n] is a slice header into the input — no string copy.
		m[s[i:i+n]]++
	}
	return m
}

// extractQGramsRunes returns the multiset of overlapping length-n RUNE
// q-grams of s. Returns a non-nil empty map[string]int{} when n < 1,
// when s is empty, or when the rune count of s is less than n.
//
// Keys are UTF-8 encoded strings produced from windows of the rune
// slice (string(runes[i:i+n])). For pure-ASCII inputs the byte and
// rune paths produce semantically equivalent maps, but the keys differ
// in identity — the rune path's keys come from a freshly-allocated
// rune-slice conversion, so callers should not rely on key identity
// across the two paths.
//
// Multiset semantics and the no-iteration discipline mirror
// extractQGrams (see godoc above).
func extractQGramsRunes(s string, n int) map[string]int {
	if n < 1 {
		return map[string]int{}
	}
	runes := []rune(s)
	if len(runes) < n {
		return map[string]int{}
	}
	// Capacity hint: at most len(runes)-n+1 distinct q-grams.
	m := make(map[string]int, len(runes)-n+1)
	for i := 0; i <= len(runes)-n; i++ {
		// string(runes[i:i+n]) allocates a fresh UTF-8 string for the
		// map key; the rune-slice itself is a sub-slice of `runes`
		// (no extra allocation).
		m[string(runes[i:i+n])]++
	}
	return m
}
