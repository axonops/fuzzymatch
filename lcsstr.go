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

// lcsstr.go implements the Longest Common Substring (LCSStr) similarity for
// the fuzzymatch catalogue.
//
// Source: Wagner, R. A., & Fischer, M. J. (1974). "The string-to-string
// correction problem." Journal of the ACM, 21(1):168-173 — the canonical
// dynamic-programming formulation for longest common substring (the
// "substring" variant of the broader LCS family — contiguous matches only,
// in contrast to LCS-subsequence which allows gaps).
//
// Recurrence (0-indexed; cost = 1 if a[i-1] == b[j-1], else recurrence resets
// to 0; tracks max and ending index):
//
//	D[0, j] = 0    (boundary)
//	D[i, 0] = 0    (boundary)
//	D[i, j] = D[i-1, j-1] + 1   if a[i-1] == b[j-1]
//	         0                  otherwise
//	max_len, end_i = max over all (i, j) of D[i, j], tracking first-found.
//
// Score normalisation (SPEC-PINNED at docs/requirements.md §7.1.9):
//
//	score = 2 · len(lcs) / (len(a) + len(b))   (Sørensen-Dice form)
//
// Edge cases:
//
//   - LongestCommonSubstring("", "")    == "" AND LCSStrScore("", "")    == 1.0
//     (both-empty convention — the 2·n/(n+n)=1 identity holds vacuously)
//   - LongestCommonSubstring("", "abc") == "" AND LCSStrScore("", "abc") == 0.0
//     (one-empty)
//   - LongestCommonSubstring("abc", "xyz") == "" AND LCSStrScore("abc", "xyz") == 0.0
//     (no overlap — the empty-string return IS documented behaviour, NOT a
//     bug; consumers disambiguate via the score)
//   - LongestCommonSubstring(x, x) == x AND LCSStrScore(x, x) == 1.0 (identity)
//
// Tie-break (LOCKED per CONTEXT.md §3): when multiple longest common
// substrings of equal length exist, the LEFTMOST occurrence in `a` is
// returned. This is the natural left-to-right DP iteration order — the
// max-update is written with STRICT-GREATER-THAN (`>`, NOT `>=`) so that the
// first-found-leftmost wins; subsequent equal-length matches do NOT override.
// This is the load-bearing regression test for RESEARCH.md Pitfall 4 — any
// drift to `>=` is caught by TestProp_LongestCommonSubstring_LeftmostTieBreak.
//
// Documented behaviour — LongestCommonSubstring returning "" is ambiguous:
// both the both-empty case AND the no-overlap case return "". Consumers use
// LCSStrScore to disambiguate (1.0 vs 0.0). See RESEARCH.md Pitfall 6.
//
// Implementation discipline (inherits Phase 2):
//
//   - ASCII fast path operates on bytes directly when the shorter dimension
//     n <= maxStackInputLen && isASCII(a) && isASCII(b); a stack-allocated
//     [(maxStackInputLen+1)*2]int buffer holds the two rolling rows.
//     (maxStackInputLen is defined in levenshtein.go — do NOT redeclare here.
//     isASCII is declared in normalise.go — referenced by name, not duplicated.)
//   - Heap path: two make([]int, n+1) calls; 2 allocs on ASCII Long.
//   - Rune path: 2 []rune allocations + 2 rolling-row allocations = 4 minimum.
//   - NO init()-time table builds (per docs/requirements.md §5(12) and
//     .claude/skills/determinism-standards): no var-level side effects.
//   - NO map iteration on output paths (DET-03).
//   - NO transcendental float operations (DET-06): only +, *, /, and float64()
//     conversion; score normalisation parenthesised left-to-right as
//     `(2.0 * float64(n)) / float64(la+lb)` per DET-06.
//   - NO goroutines, channels, or mutexes.
//   - Identity short-circuit `if a == b { return ... }` on both LCSStrScore
//     and LCSStrScoreRunes BEFORE any []rune allocation (IN-04 closure).
//
// Public surface (four functions, spec-pinned at docs/requirements.md §7.1.9):
//
//   - LongestCommonSubstring(a, b string) string
//   - LongestCommonSubstringRunes(a, b string) string
//   - LCSStrScore(a, b string) float64
//   - LCSStrScoreRunes(a, b string) float64
//
// Only LCSStrScore is registered in the dispatch table (slot 8 — see
// algoid.go AlgoLCSStr; dispatch is float64-valued so the string-returning
// surfaces are not dispatched).

package fuzzymatch

// LongestCommonSubstring returns the longest substring common to a and b.
// Returns the empty string when (a) both inputs are empty, OR (b) the inputs
// share no characters. Consumers wanting to disambiguate these two cases
// should call LCSStrScore: LCSStrScore("", "") == 1.0 whereas
// LCSStrScore("abc", "xyz") == 0.0.
//
// When multiple longest common substrings of equal length exist, the LEFTMOST
// occurrence in a is returned (the natural left-to-right DP iteration order;
// the max-update uses strict `>` so first-found-leftmost wins).
//
// Edge cases:
//   - LongestCommonSubstring("", "")    == ""
//   - LongestCommonSubstring("", "abc") == ""
//   - LongestCommonSubstring("abc", "") == ""
//   - LongestCommonSubstring("abc", "xyz") == "" (no shared characters)
//   - LongestCommonSubstring("abc", "abc") == "abc"
//   - LongestCommonSubstring("abcXYZabc", "abc") == "abc" (leftmost)
//
// Worst-case time: O(m·n) where m = len(a), n = len(b).
// Space: O(min(m,n)) — two-row DP, no full m×n table allocated.
//
// Performance scope (Q7c, docs/requirements.md §14.1):
//
//	The allocation budget published in §14.1 reflects the ASCII fast path:
//	when both inputs are pure ASCII AND the shorter dimension n ≤
//	maxStackInputLen, the two DP rows live in a stack-allocated buffer
//	(zero heap allocations). For long inputs (n > maxStackInputLen, ≥ 500
//	chars on either side) the algorithm falls back to a heap-allocated
//	two-row DP — per-call allocation count rises to 2 (one make per row)
//	and byte-count scales as O(n)·sizeof(int). See §14.1's long-input
//	row for the relaxed budget; the threat model bounds the worst case
//	via the per-algorithm input ceiling.
//
// This function operates on bytes. For multi-byte UTF-8 inputs, use
// LongestCommonSubstringRunes to obtain the rune-aware substring.
//
// # Substring escape (shared backing storage)
//
// The non-empty return value shares its backing storage with a — it is a
// slice header into a's underlying bytes, returned without a copy. Callers
// that retain the result across the lifetime of a will keep a's backing
// storage alive. For typical small-string usage this is benign. For a
// consumer that constructs an ephemeral mega-string and extracts a tiny
// shared segment with LongestCommonSubstring, defensively copy the result
// (string([]byte(result))) before discarding the source. The phrasing
// mirrors the Go stdlib convention for substring-returning functions such
// as strings.SplitN and strings.Index.
func LongestCommonSubstring(a, b string) string {
	if a == b {
		return a // identity short-circuit (covers identical and both-empty=="")
	}
	if len(a) == 0 || len(b) == 0 {
		return ""
	}
	// We want the inner DP loop to iterate over the SHORTER dimension so
	// the stack buffer stays small. However, the tie-break invariant is
	// defined as "leftmost in `a`" — swapping a and b would yield the
	// leftmost in the SWAPPED `a` (i.e. the original `b`). So we must NOT
	// swap when we need to report the substring; instead we make the inner
	// loop range over b directly with prev/curr sized to len(b)+1, which
	// preserves the directional invariant.
	m, n := len(a), len(b)
	if n <= maxStackInputLen && isASCII(a) && isASCII(b) {
		var buf [(maxStackInputLen + 1) * 2]int
		maxLen, endI := lcsstrDP(a, b, m, n, buf[:n+1], buf[n+1:2*(n+1)])
		if maxLen == 0 {
			return ""
		}
		return a[endI-maxLen : endI]
	}
	maxLen, endI := lcsstrDP(a, b, m, n, make([]int, n+1), make([]int, n+1))
	if maxLen == 0 {
		return ""
	}
	return a[endI-maxLen : endI]
}

// LongestCommonSubstringRunes is the rune-path variant of
// LongestCommonSubstring. It treats each input as a sequence of Unicode code
// points (runes) rather than bytes, so multi-byte UTF-8 sequences are
// compared atomically. For example, "café" and "cafe" share the leftmost
// rune-substring "caf" (length 3 runes).
//
// The rune variant allocates two []rune slices plus two DP rows on the heap.
// For ASCII inputs, prefer LongestCommonSubstring (zero allocations on inputs
// ≤ 64 bytes).
//
// The leftmost-in-a tie-break and edge-case conventions are identical to
// LongestCommonSubstring.
//
// The return value is a freshly-allocated string built from a []rune
// segment of the converted a — unlike the byte path it does NOT share
// backing storage with the caller's a.
func LongestCommonSubstringRunes(a, b string) string {
	if a == b {
		return a // identity short-circuit — avoids []rune allocations
	}
	if len(a) == 0 || len(b) == 0 {
		return ""
	}
	ra := []rune(a) // 1 alloc
	rb := []rune(b) // 1 alloc
	m, n := len(ra), len(rb)
	maxLen, endI := lcsstrDPRunes(ra, rb, m, n, make([]int, n+1), make([]int, n+1))
	if maxLen == 0 {
		return ""
	}
	return string(ra[endI-maxLen : endI])
}

// LCSStrScore returns the Longest-Common-Substring similarity between a and b
// in [0.0, 1.0] using the Sørensen-Dice normalisation:
//
//	score = 2 · len(lcs) / (len(a) + len(b))
//
// For programmatic input-quality checks before scoring,
// see [fuzzymatch.Validate].
//
// Edge cases:
//   - LCSStrScore("", "")    == 1.0 (both-empty convention)
//   - LCSStrScore("", "abc") == 0.0 (one-empty)
//   - LCSStrScore("abc", "abc") == 1.0 (identity; 2·3/6 = 1)
//   - LCSStrScore("abc", "xyz") == 0.0 (no overlap)
//
// This function operates on bytes. For multi-byte UTF-8 inputs, use
// LCSStrScoreRunes to obtain the rune-aware similarity.
func LCSStrScore(a, b string) float64 {
	if a == b {
		return 1.0 // identity short-circuit (covers both-empty too)
	}
	la, lb := len(a), len(b)
	if la == 0 || lb == 0 {
		return 0.0
	}
	n := lcsstrLengthOnly(a, b, la, lb)
	// Explicit left-to-right parenthesisation per DET-06 (cross-platform
	// float-determinism).
	numer := 2.0 * float64(n)
	denom := float64(la + lb)
	return numer / denom
}

// LCSStrScoreRunes is the rune-path variant of LCSStrScore. It treats each
// input as a sequence of Unicode code points (runes) rather than bytes.
//
// The score is normalised by the SUM of the two rune lengths (NOT byte
// lengths), so for example LCSStrScoreRunes("café", "cafe") = 2·3/(4+4) =
// 0.75 (3 runes match; 4-rune length each).
//
// The rune variant allocates two []rune slices plus two DP rows on the heap.
// For ASCII inputs, prefer LCSStrScore (zero allocations on inputs ≤ 64
// bytes).
func LCSStrScoreRunes(a, b string) float64 {
	if a == b {
		return 1.0 // identity short-circuit — avoids []rune allocations
	}
	if len(a) == 0 || len(b) == 0 {
		return 0.0
	}
	ra := []rune(a) // 1 alloc
	rb := []rune(b) // 1 alloc
	la, lb := len(ra), len(rb)
	n := lcsstrLengthOnlyRunes(ra, rb, la, lb)
	// Explicit left-to-right parenthesisation per DET-06.
	numer := 2.0 * float64(n)
	denom := float64(la + lb)
	return numer / denom
}

// lcsstrDP runs the two-row DP recurrence and returns (length, endIndexInA).
// prev and curr must each have length n+1. Operates on string bytes directly.
//
// Leftmost-in-a tie-break is established by the STRICT-GREATER-THAN (`>`,
// NOT `>=`) max-update at curr[j] > maxLen: first-found-leftmost wins
// because subsequent equal-length matches do NOT override the recorded
// (maxLen, endI). Changing this `>` to `>=` flips the tie-break to
// rightmost-in-a, breaking RESEARCH.md Pitfall 4's contract.
//
// Caller guarantees: m > 0, n > 0; prev and curr are zero-initialised
// (stack arrays are zero; make([]int, ...) is zero).
func lcsstrDP(a, b string, m, n int, prev, curr []int) (length, endI int) {
	var maxLen, maxEnd int
	for i := 1; i <= m; i++ {
		for j := 1; j <= n; j++ {
			if a[i-1] == b[j-1] {
				curr[j] = prev[j-1] + 1
				if curr[j] > maxLen { // STRICT > — leftmost tie-break (Pitfall 4)
					maxLen = curr[j]
					maxEnd = i // exclusive end index in `a`
				}
			} else {
				curr[j] = 0 // recurrence resets on mismatch
			}
		}
		prev, curr = curr, prev
		// No re-zero of curr needed: the inner loop above writes curr[j]
		// unconditionally for every j in 1..n (matched branch writes
		// prev[j-1]+1; mismatched branch writes 0). curr[0] is never read.
		// The initial buffers arrive zero-initialised (stack arrays are
		// zero; make([]int, ...) is zero), so the first iteration is also
		// safe.
	}
	return maxLen, maxEnd
}

// lcsstrDPRunes is the rune-slice variant of lcsstrDP. The recurrence is
// identical; only the indexing source differs (rune comparison rather than
// byte comparison).
func lcsstrDPRunes(a, b []rune, m, n int, prev, curr []int) (length, endI int) {
	var maxLen, maxEnd int
	for i := 1; i <= m; i++ {
		for j := 1; j <= n; j++ {
			if a[i-1] == b[j-1] {
				curr[j] = prev[j-1] + 1
				if curr[j] > maxLen { // STRICT > — leftmost tie-break
					maxLen = curr[j]
					maxEnd = i
				}
			} else {
				curr[j] = 0
			}
		}
		prev, curr = curr, prev
		// No re-zero needed — see lcsstrDP comment above.
	}
	return maxLen, maxEnd
}

// lcsstrLengthOnly computes the longest-common-substring LENGTH without
// returning the substring itself. It uses the same two-row DP as lcsstrDP
// but allocates from the appropriate buffer based on the ASCII fast-path
// gate.
//
// Caller guarantees: la > 0, lb > 0.
func lcsstrLengthOnly(a, b string, la, lb int) int {
	if lb <= maxStackInputLen && isASCII(a) && isASCII(b) {
		var buf [(maxStackInputLen + 1) * 2]int
		n, _ := lcsstrDP(a, b, la, lb, buf[:lb+1], buf[lb+1:2*(lb+1)])
		return n
	}
	n, _ := lcsstrDP(a, b, la, lb, make([]int, lb+1), make([]int, lb+1))
	return n
}

// lcsstrLengthOnlyRunes is the rune-slice variant of lcsstrLengthOnly.
// Always heap-allocates the two rolling rows (no stack fast path for the
// rune slice variant — matches Phase 2/3 convention).
//
// Caller guarantees: la > 0, lb > 0.
func lcsstrLengthOnlyRunes(a, b []rune, la, lb int) int {
	n, _ := lcsstrDPRunes(a, b, la, lb, make([]int, lb+1), make([]int, lb+1))
	return n
}
