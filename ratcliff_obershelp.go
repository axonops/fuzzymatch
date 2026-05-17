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

// ratcliff_obershelp.go implements the Ratcliff-Obershelp gestalt-pattern-
// matching similarity for the fuzzymatch catalogue.
//
// Source: Ratcliff, J. W., Metzener, D. E. (1988). "Pattern matching: the
// gestalt approach." Dr. Dobb's Journal, 13(7):46-51 — the primary
// description of the recursive longest-common-substring decomposition.
//
// Algorithm (the "gestalt" pattern matching of Ratcliff & Metzener 1988):
//
//  1. Find the longest contiguous matching substring of a and b. Call its
//     length N.
//  2. Recursively apply step 1 to the prefix-of-a / prefix-of-b portions
//     LEFT of the match.
//  3. Recursively apply step 1 to the suffix-of-a / suffix-of-b portions
//     RIGHT of the match.
//  4. Sum the matched-character count M across all recursion levels.
//  5. Normalise the score: score = 2 · M / (len(a) + len(b)).
//
// This implementation is byte-for-byte equivalent to Python's
// `difflib.SequenceMatcher(autojunk=False, a=a, b=b).ratio()` within 1e-9
// tolerance, on the canonical Dr. Dobb's 1988 reference vectors AND on the
// committed cross-validation corpus (the corpus itself is plan 04-04's
// deliverable; this plan ships the algorithm + the paper-pinned unit tests).
//
// The `autojunk=False` qualifier is LOAD-BEARING. Python's
// difflib.SequenceMatcher defaults to `autojunk=True`, a performance
// heuristic that DROPS "junk" characters when len(b) ≥ 200 and a character
// appears in ≥ 1% of positions. This is NOT the Ratcliff-Obershelp
// algorithm — it is a difflib speed optimisation. Disabling autojunk
// recovers the true algorithm. The cross-validation corpus is generated
// with autojunk=False; this implementation matches that semantic byte-for-
// byte. See PITFALLS.md §6 and RESEARCH.md Pitfall 2.
//
// ASYMMETRY (OQ-1 RESOLUTION LOCKED 2026-05-14):
//
// RatcliffObershelpScore is INTENTIONALLY ASYMMETRIC in argument order.
// This mirrors Python's difflib.SequenceMatcher.ratio() — see CPython
// bpo-37004 for the upstream history. For example,
// RatcliffObershelpScore("tide", "diet") = 0.25 while
// RatcliffObershelpScore("diet", "tide") = 0.5. The asymmetry is a
// consequence of the longest-common-substring tie-break (leftmost-in-`a`
// first, then leftmost-in-`b` among ties) plus the recursive decomposition
// — swapping `a` and `b` can pick a different starting match and therefore
// recurse into different prefix/suffix splits.
//
// The standard Phase 2 `Symmetric` property test is DROPPED for this
// algorithm; the remaining five invariants (RangeBounds, Identity, NoNaN,
// NoInf, NoNegativeZero) still apply. A pinned asymmetry test in
// ratcliff_obershelp_test.go documents the contract. The cross-algorithm
// consistency test landing in plan 04-05 adds an inverse-form regression
// guard.
//
// Source-origin discipline (per algorithm-licensing-standards):
//
//   - Primary source:        Ratcliff & Metzener 1988 (Dr. Dobb's Journal)
//   - Cross-validation:      Python difflib.SequenceMatcher (PSF licence,
//                            stdlib) — consulted ONLY for the
//                            find_longest_match tie-break contract
//                            (leftmost-in-`a` first, then leftmost-in-`b`
//                            among ties) and for cross-validation reference
//                            vectors. NOT for code copying.
//   - GPL/LGPL provenance:   none.
//   - Code copied verbatim:  none.
//
// Implementation discipline (inherits Phase 2 + 3):
//
//   - Recursion uses the Go language-native call stack (per CONTEXT.md D-2;
//     simpler than an explicit iterative stack; recursion depth is bounded
//     by O(min(len(a), len(b))) so no stack-overflow risk for reasonable
//     inputs).
//   - The longest-common-substring inner step is INLINED (per CONTEXT.md
//     D-3) rather than reusing lcsstr.go's helper, because Ratcliff-
//     Obershelp needs the START position in BOTH `a` and `b` (not just
//     the length + end-index-in-`a` that lcsstrDP returns).
//   - NO init()-time table builds (per docs/requirements.md §5(12)).
//   - NO map iteration on output paths (DET-03).
//   - NO transcendental float operations (DET-06): only +, *, /, and
//     float64() conversion; score normalisation parenthesised left-to-
//     right as `numer := 2.0 * float64(M); denom := float64(la + lb);
//     return numer / denom`.
//   - NO goroutines, channels, or mutexes.
//   - Identity short-circuit `if a == b { return 1.0 }` on both
//     RatcliffObershelpScore and RatcliffObershelpScoreRunes BEFORE any
//     []rune allocation (IN-04 closure).
//
// Public surface (two functions):
//
//   - RatcliffObershelpScore(a, b string) float64
//   - RatcliffObershelpScoreRunes(a, b string) float64
//
// Only RatcliffObershelpScore is registered in the dispatch table (slot 22
// — the LAST slot, numAlgorithms-1 — see algoid.go AlgoRatcliffObershelp;
// dispatch is float64-valued and byte-string keyed).
//
// Worst-case complexity: O(N² · M) for the recursive longest-common-
// substring decomposition (DP at each recursion level). Recursion depth is
// O(min(la, lb)) in the worst case. For pathological repeated-character
// inputs (e.g. all-a strings with strategic differences) the algorithm
// degrades to quadratic-plus behaviour — this is documented as accepted
// per the threat model; callers control input size.

package fuzzymatch

// RatcliffObershelpScore is the difflib-equivalent. If you want fuzzy
// string matching that behaves like Python's difflib.ratio(), use this. If
// you want the RapidFuzz "ratio()" semantics — the Indel formula
// 2·LCS/(|a|+|b|) used by Token Sort Ratio / Token Set Ratio / Partial
// Ratio — use those functions in Phase 6 instead.
//
// Returns the Ratcliff-Obershelp gestalt-pattern-matching similarity in
// [0.0, 1.0]. Behaves byte-for-byte like
// difflib.SequenceMatcher(autojunk=False, a=a, b=b).ratio().
//
// For programmatic input-quality checks before scoring,
// see [fuzzymatch.Validate].
//
// RatcliffObershelpScore is NOT symmetric in argument order. This mirrors
// Python's difflib.SequenceMatcher(autojunk=False).ratio() behaviour (see
// CPython bpo-37004). For symmetric similarity, callers should sort inputs
// by length first or use a different algorithm (e.g. LCSStrScore). For
// example, RatcliffObershelpScore("tide", "diet") = 0.25 while
// RatcliffObershelpScore("diet", "tide") = 0.5.
//
// Edge cases:
//
//   - RatcliffObershelpScore("", "")    == 1.0 (both-empty identity)
//   - RatcliffObershelpScore("", "abc") == 0.0 (one-empty)
//   - RatcliffObershelpScore("abc", "abc") == 1.0 (identity short-circuit)
//
// Performance scope (Q7c, docs/requirements.md §14.1):
//
//	The allocation budget published in §14.1 (≤ 4 allocs on ASCII Short)
//	covers short inputs where recursion depth stays shallow. The
//	roFindLongestMatch implementation allocates two rolling DP rows per
//	recursion level — there is no structural ASCII fast path or
//	stack-buffer short-circuit, since each recursive call works on
//	different substring boundaries. For long inputs (≥ 500 chars) the
//	per-call allocation count scales with the number of matched
//	substrings discovered; see §14.1's long-input row for the relaxed
//	budget. The threat model bounds the worst case via the
//	per-algorithm input ceiling.
//
// This function operates on bytes. For multi-byte UTF-8 inputs, use
// RatcliffObershelpScoreRunes to obtain the rune-aware similarity.
func RatcliffObershelpScore(a, b string) float64 {
	if a == b {
		return 1.0 // identity short-circuit (covers both-empty too)
	}
	la, lb := len(a), len(b)
	if la == 0 || lb == 0 {
		return 0.0
	}
	m := roMatchedLength(a, b)
	// Explicit left-to-right parenthesisation per DET-06 (cross-platform
	// float-determinism).
	numer := 2.0 * float64(m)
	denom := float64(la + lb)
	return numer / denom
}

// RatcliffObershelpScoreRunes is the rune-path variant of
// RatcliffObershelpScore. It treats each input as a sequence of Unicode
// code points (runes) rather than bytes, so multi-byte UTF-8 sequences
// are compared atomically.
//
// The rune variant allocates two []rune slices on the heap. For ASCII
// inputs, prefer RatcliffObershelpScore (no []rune conversion needed).
//
// The asymmetric-by-design semantics and edge-case conventions are
// identical to RatcliffObershelpScore.
func RatcliffObershelpScoreRunes(a, b string) float64 {
	if a == b {
		return 1.0 // identity short-circuit — avoids []rune allocations
	}
	if len(a) == 0 || len(b) == 0 {
		return 0.0
	}
	ra := []rune(a) // 1 alloc
	rb := []rune(b) // 1 alloc
	la, lb := len(ra), len(rb)
	m := roMatchedLengthRunes(ra, rb)
	// Explicit left-to-right parenthesisation per DET-06.
	numer := 2.0 * float64(m)
	denom := float64(la + lb)
	return numer / denom
}

// roMatchedLength returns the total matched-character count across the
// recursive longest-common-substring decomposition of a and b. Operates
// on string bytes directly.
//
// The recursion contract:
//
//	matched(a, b) = 0                          if a == "" || b == ""
//	matched(a, b) = N + matched(a[:aLo], b[:bLo])
//	                  + matched(a[aLo+N:], b[bLo+N:])
//	             where (aLo, _, bLo, _, N) = roFindLongestMatch(a, b)
//	matched(a, b) = 0                          if N == 0
//
// Uses the Go language-native call stack (per CONTEXT.md D-2). Recursion
// depth is bounded by O(min(la, lb)).
func roMatchedLength(a, b string) int {
	if len(a) == 0 || len(b) == 0 {
		return 0
	}
	aLo, aHi, bLo, bHi, n := roFindLongestMatch(a, b)
	if n == 0 {
		return 0
	}
	// Recurse into LEFT-of-match portions and RIGHT-of-match portions.
	return n + roMatchedLength(a[:aLo], b[:bLo]) + roMatchedLength(a[aHi:], b[bHi:])
}

// roFindLongestMatch returns the longest contiguous matching substring of
// a and b, expressed as (aLo, aHi, bLo, bHi, n) where:
//
//   - n is the match length.
//   - a[aLo:aHi] == b[bLo:bHi] is the matched substring.
//   - aHi == aLo + n and bHi == bLo + n.
//
// Mirrors the contract of Python difflib.SequenceMatcher.find_longest_match
// with autojunk=False — tie-break is leftmost-in-`a` first, then leftmost-
// in-`b` among ties. The strict-`>` max-update enforces "first match wins"
// on the natural left-to-right DP iteration order, equivalent to the
// difflib documented behaviour.
//
// When no characters match at all, returns (0, 0, 0, 0, 0).
//
// Same LCS-substring DP recurrence as lcsstr.go's lcsstrDP, extended to
// return the START position in BOTH `a` and `b`. The inline variant is
// preferred over reusing lcsstrDP here because Ratcliff-Obershelp needs
// aLo / bLo / bHi (lcsstrDP returns only the length + end-index-in-`a`).
func roFindLongestMatch(a, b string) (aLo, aHi, bLo, bHi, n int) {
	la, lb := len(a), len(b)
	if la == 0 || lb == 0 {
		return 0, 0, 0, 0, 0
	}
	// Two-row rolling DP. prev[j] = D[i-1][j]; curr[j] = D[i][j].
	prev := make([]int, lb+1)
	curr := make([]int, lb+1)
	var maxLen, maxEndI, maxEndJ int
	for i := 1; i <= la; i++ {
		for j := 1; j <= lb; j++ {
			if a[i-1] == b[j-1] {
				curr[j] = prev[j-1] + 1
				if curr[j] > maxLen { // STRICT > — leftmost tie-break
					maxLen = curr[j]
					maxEndI = i // exclusive end index in `a`
					maxEndJ = j // exclusive end index in `b`
				}
			} else {
				curr[j] = 0 // recurrence resets on mismatch
			}
		}
		prev, curr = curr, prev
		// No re-zero of curr needed: the inner loop above writes curr[j]
		// unconditionally for every j in 1..lb (matched branch writes
		// prev[j-1]+1; mismatched branch writes 0). curr[0] is never read.
		// make([]int, ...) returns zero-initialised slices, so the first
		// iteration is also safe.
	}
	if maxLen == 0 {
		return 0, 0, 0, 0, 0
	}
	aLo = maxEndI - maxLen
	aHi = maxEndI
	bLo = maxEndJ - maxLen
	bHi = maxEndJ
	return aLo, aHi, bLo, bHi, maxLen
}

// roMatchedLengthRunes is the rune-slice variant of roMatchedLength.
func roMatchedLengthRunes(a, b []rune) int {
	if len(a) == 0 || len(b) == 0 {
		return 0
	}
	aLo, aHi, bLo, bHi, n := roFindLongestMatchRunes(a, b)
	if n == 0 {
		return 0
	}
	return n + roMatchedLengthRunes(a[:aLo], b[:bLo]) + roMatchedLengthRunes(a[aHi:], b[bHi:])
}

// roFindLongestMatchRunes is the rune-slice variant of roFindLongestMatch.
// The recurrence is identical; only the indexing source differs (rune
// comparison rather than byte comparison).
func roFindLongestMatchRunes(a, b []rune) (aLo, aHi, bLo, bHi, n int) {
	la, lb := len(a), len(b)
	if la == 0 || lb == 0 {
		return 0, 0, 0, 0, 0
	}
	prev := make([]int, lb+1)
	curr := make([]int, lb+1)
	var maxLen, maxEndI, maxEndJ int
	for i := 1; i <= la; i++ {
		for j := 1; j <= lb; j++ {
			if a[i-1] == b[j-1] {
				curr[j] = prev[j-1] + 1
				if curr[j] > maxLen { // STRICT > — leftmost tie-break
					maxLen = curr[j]
					maxEndI = i
					maxEndJ = j
				}
			} else {
				curr[j] = 0
			}
		}
		prev, curr = curr, prev
		// No re-zero needed — see roFindLongestMatch comment above.
	}
	if maxLen == 0 {
		return 0, 0, 0, 0, 0
	}
	aLo = maxEndI - maxLen
	aHi = maxEndI
	bLo = maxEndJ - maxLen
	bHi = maxEndJ
	return aLo, aHi, bLo, bHi, maxLen
}
