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

// token_indel.go provides the unexported LCS-subsequence + Indel-formula
// helpers shared by TokenSortRatio (this plan, 06-01), TokenSetRatio
// (plan 06-02), and PartialRatio (plan 06-03). The public surface for the
// token tier is the per-algorithm Score / ScoreRunes functions — these
// kernel helpers are package-internal per 06-CONTEXT.md §2 LOCKED.
//
// Source: Wagner, R. A., & Fischer, M. J. (1974). "The string-to-string
// correction problem." Journal of the ACM 21(1):168-173 — the canonical
// dynamic-programming formulation of the longest-common-SUBSEQUENCE (LCS)
// problem. The Indel-formula normalisation (2·LCS / (|a|+|b|)) is the
// RapidFuzz-canonical "Indel" similarity, documented at
// https://rapidfuzz.github.io/RapidFuzz/Usage/distance/Indel.html and
// mathematically equivalent to `1 - Indel-distance(a, b) / (|a|+|b|)`
// where Indel-distance is the Levenshtein distance restricted to
// insertions and deletions only (no substitutions) — see 06-RESEARCH.md
// Pattern 3 for the proof.
//
// Algorithm — lcsLen(a, b):
//
//	D[0, j] = 0      (boundary)
//	D[i, 0] = 0      (boundary)
//	D[i, j] = D[i-1, j-1] + 1                if a[i-1] == b[j-1]
//	         max(D[i-1, j], D[i, j-1])       otherwise
//
//	lcsLen(a, b) = D[|a|, |b|]
//
// IMPORTANT — divergence from lcsstr.go's substring DP (PITFALL 6):
// this kernel computes the longest common SUBSEQUENCE — matches MAY HAVE
// GAPS. `lcsstr.go`'s `LongestCommonSubstring` computes the longest
// common SUBSTRING — matches MUST BE CONTIGUOUS. The two are NOT
// interchangeable:
//
//   - lcsLen([]byte("abc"), []byte("axc"))                == 2  (subsequence "ac")
//   - len(LongestCommonSubstring("abc", "axc"))           == 1  (substring "a" or "c")
//
// `token_indel_test.go::TestLCSLen_DistinctFromLCSStr` pins this
// divergence as a regression gate. Any future refactoring that
// accidentally routes an LCS-subsequence call through the substring
// kernel (or vice versa) is caught at unit-test time.
//
// Score normalisation — indelRatio(a, b) (Sørensen-Dice form over LCS):
//
//	indelRatio = 2 · lcsLen(a, b) / (|a| + |b|)
//
// Edge cases:
//
//   - indelRatio("", "")    == 1.0   (both-empty identity; vacuous match)
//   - indelRatio("", "abc") == 0.0   (one-empty)
//   - indelRatio("abc", "") == 0.0   (one-empty)
//   - indelRatio("abc", "abc") == 1.0   (identity; 2·3/6 = 1)
//   - indelRatio("abc", "axc") == 2·2/6 ≈ 0.6666666666666666
//
// Determinism (DET-03 + DET-06):
//
//   - NO map iteration on any output path — the DP fills two integer
//     arrays in left-to-right column order; the output is a single
//     scalar derived from integer cardinalities.
//   - NO transcendental floats: integer arithmetic in the kernel,
//     plus a single `2.0 * float64(lcs) / float64(la+lb)` division in
//     indelRatio with explicit left-to-right parenthesisation.
//   - NO goroutines, channels, or mutexes.
//
// Source-origin discipline (per algorithm-licensing-standards):
//
//   - Primary source:        Wagner & Fischer 1974 J. ACM 21(1):168-173.
//   - Cross-validation:      RapidFuzz 3.14.5 via the corpus at
//                            testdata/cross-validation/token-ratios/vectors.json
//                            (committed by plan 06-01 task 3). The
//                            corpus pins TokenSortRatio (this plan) plus
//                            placeholder entries for TokenSetRatio and
//                            PartialRatio scheduled for plans 06-02 / 06-03.
//   - Tie-break:             none (LCS-subsequence length is a scalar
//                            integer; no candidate ordering to break).
//   - GPL/LGPL provenance:   none.
//   - Code copied verbatim:  none.
//
// Implementation discipline:
//
//   - NO init()-time table builds (per docs/requirements.md §5(12)).
//   - The two DP rows use explicit `if/else` selection of the row-max
//     rather than the Go 1.26 builtin `max`. This mirrors the canonical
//     determinism-reviewer-auditable pattern from qgram_jaccard.go
//     (lines 240-246) and ensures the float-or-integer choice is
//     obvious in code review.
//   - The inner loop always iterates over the SHORTER side. Inputs are
//     swapped (`a, b = b, a`) so that `m <= n` before allocating rows.
//     This keeps the stack buffer small and the heap row size minimal.
//   - `maxStackInputLen` is the shared constant from levenshtein.go
//     (= 64). DO NOT redeclare it here. Inputs with `m <= 64` use a
//     stack-allocated `[maxStackInputLen+1]int` buffer pair (zero
//     heap allocations); longer inputs fall back to `make([]int, m+1)`
//     for each row (2 heap allocations total).
//
// Allocation budget (RESEARCH.md Pattern 2 / docs/requirements.md §14.1):
//
//   - lcsLen on inputs with min(|a|, |b|) <= 64: 0 heap allocations.
//   - lcsLen on inputs with min(|a|, |b|) > 64: 2 heap allocations
//     (one per DP row; the row pointers are swapped in place each
//     outer iteration so no per-row reallocation occurs).
//   - indelRatio inherits lcsLen's budget plus one float-arithmetic
//     reduction (no allocation).
//   - The rune-aware variants (`lcsLenRunes` / `indelRatioRunes`) do
//     not allocate the input rune slice — the caller is expected to
//     have done that work. They inherit the same row-allocation
//     budget as the byte variants.
//
// Plan-introducer / consumer-plan godoc note (per export_test.go
// pattern lines 115-129): plan 06-01 introduces this file; plans
// 06-02 (TokenSetRatio), 06-03 (PartialRatio), and 06-05 (MongeElkan
// — indirectly via the PartialRatio dependency) consume these helpers
// without modifying them.

package fuzzymatch

// lcsLen returns the length of the longest common SUBSEQUENCE of a and b
// (Wagner & Fischer 1974). Operates on byte slices.
//
// Edge cases (asserted by token_indel_test.go::TestLCSLen):
//
//   - lcsLen(nil, nil)               == 0
//   - lcsLen([]byte(""), []byte("")) == 0
//   - lcsLen([]byte("abc"), []byte("axc")) == 2 (subsequence "ac")
//   - lcsLen([]byte("abc"), []byte("abc")) == 3 (identity)
//
// Worst-case time: O(m·n) where m = min(len(a), len(b)), n = max(len(a), len(b)).
// Space: O(min(m, n)) — two-row DP, no full m×n table allocated.
// Allocation budget: 0 heap allocs when min(m, n) <= maxStackInputLen
// (= 64); 2 heap allocs otherwise (one per DP row).
//
// The inner loop iterates over the SHORTER side; inputs are swapped
// before allocating to keep the stack buffer small. The swap does not
// affect the result (LCS-subsequence length is symmetric in (a, b)).
//
// PITFALL 6 — this is the SUBSEQUENCE length, NOT the substring length.
// See the file-header godoc and TestLCSLen_DistinctFromLCSStr for the
// divergence regression gate.
func lcsLen(a, b []byte) int {
	la, lb := len(a), len(b)
	if la == 0 || lb == 0 {
		return 0
	}
	// Iterate the inner loop over the SHORTER side. Subsequence length
	// is symmetric, so the swap is value-preserving.
	if la > lb {
		a, b = b, a
		la, lb = lb, la
	}
	// `la` is now min(|a|, |b|); the DP rows have length la+1.
	if la <= maxStackInputLen {
		var prevArr, currArr [maxStackInputLen + 1]int
		return lcsLenDP(a, b, la, lb, prevArr[:la+1], currArr[:la+1])
	}
	return lcsLenDP(a, b, la, lb, make([]int, la+1), make([]int, la+1))
}

// lcsLenDP runs the two-row LCS-subsequence DP and returns
// D[lb][la] = lcsLen(a, b) where `a` is the inner-loop side
// (length la) and `b` is the outer-loop side (length lb). prev and curr
// MUST each have length la+1.
//
// Caller guarantees: la > 0, lb > 0, la <= lb. The buffers arrive
// zero-initialised (stack arrays are zero; make([]int, ...) is zero).
//
// The row-max selection (`max(prev[i], curr[i-1])`) is written as an
// explicit if/else rather than builtin `max` per the canonical
// determinism-reviewer-auditable pattern from qgram_jaccard.go
// (lines 240-246). This makes the recurrence step explicit in code
// review and matches the broader project style.
func lcsLenDP(a, b []byte, la, lb int, prev, curr []int) int {
	// The outer loop walks the longer side b. The inner loop walks the
	// shorter side a (length la). This gives a column-major fill: each
	// outer iteration j computes D[j][1..la] from D[j-1][...].
	for j := 1; j <= lb; j++ {
		// curr[0] is the "no-a-consumed" prefix; always 0. Initialised
		// implicitly by the zero buffer on the first iteration; on
		// subsequent iterations the swap below makes curr[0] from the
		// previous row's curr[0] (still 0 because we never write it).
		// No explicit reset is needed.
		bj := b[j-1]
		for i := 1; i <= la; i++ {
			if a[i-1] == bj {
				curr[i] = prev[i-1] + 1
			} else {
				// Row-max — explicit if/else (NOT builtin `max`) per
				// determinism-reviewer auditability.
				if prev[i] > curr[i-1] {
					curr[i] = prev[i]
				} else {
					curr[i] = curr[i-1]
				}
			}
		}
		// Roll: the freshly-computed curr becomes prev for the next j;
		// the now-stale prev becomes the curr scratch row (its contents
		// are overwritten unconditionally in the inner loop above, so
		// no zero-reset is needed).
		prev, curr = curr, prev
	}
	// After the final swap, prev holds the last-computed row.
	return prev[la]
}

// indelRatio returns the Indel-formula similarity between byte slices a
// and b:
//
//	indelRatio(a, b) = 2 · lcsLen(a, b) / (|a| + |b|)
//
// Edge cases:
//
//   - indelRatio(empty, empty)   == 1.0   (both-empty identity)
//   - indelRatio(empty, b≠empty) == 0.0   (one-empty)
//   - indelRatio(a≠empty, empty) == 0.0   (one-empty)
//   - indelRatio(x, x)           == 1.0   (identity; 2·|x|/(2|x|) = 1)
//
// The both-empty identity is the vacuous-match convention shared with
// QGramJaccard / LCSStr / RatcliffObershelp (returns 1.0 because the
// two empty bags trivially agree). One-empty returns 0.0 to match
// RapidFuzz's documented Indel ratio behaviour.
//
// Score normalisation is computed with explicit left-to-right
// parenthesisation per DET-06 (cross-platform float-determinism):
// `(2.0 * float64(lcs)) / float64(la+lb)`. For inputs where
// la + lb <= 2^53 (~9 × 10^15 — far above any realistic input),
// the numerator and denominator fit exactly in float64; IEEE-754
// correctly-rounded division produces byte-identical output across
// all four CI platforms.
func indelRatio(a, b []byte) float64 {
	la, lb := len(a), len(b)
	sum := la + lb
	if sum == 0 {
		return 1.0 // both-empty identity (vacuous match)
	}
	if la == 0 || lb == 0 {
		return 0.0 // one-empty
	}
	lcs := lcsLen(a, b)
	// Explicit left-to-right parenthesisation per DET-06 (mirrors
	// lcsstr.go LCSStrScore lines 213-216).
	numer := 2.0 * float64(lcs)
	denom := float64(sum)
	return numer / denom
}

// lcsLenRunes returns the LCS-subsequence length over rune slices.
// Operates on []rune so multi-byte UTF-8 sequences are compared
// atomically: for "café" / "cafe" the rune-LCS is 3 ("caf"), not 4
// (the byte-path would split "é" mid-codepoint and compute differently).
//
// The two-row DP and inner-loop-over-shorter-side swap mirror
// lcsLen(byte) exactly; only the comparison source changes (rune
// equality rather than byte equality).
//
// Allocation budget: 0 heap allocs when min(|a|, |b|) <= maxStackInputLen;
// 2 heap allocs otherwise. The caller is responsible for the rune-slice
// allocations themselves (typically one `[]rune(s)` conversion per side).
func lcsLenRunes(a, b []rune) int {
	la, lb := len(a), len(b)
	if la == 0 || lb == 0 {
		return 0
	}
	if la > lb {
		a, b = b, a
		la, lb = lb, la
	}
	if la <= maxStackInputLen {
		var prevArr, currArr [maxStackInputLen + 1]int
		return lcsLenDPRunes(a, b, la, lb, prevArr[:la+1], currArr[:la+1])
	}
	return lcsLenDPRunes(a, b, la, lb, make([]int, la+1), make([]int, la+1))
}

// lcsLenDPRunes is the rune-slice variant of lcsLenDP. The recurrence
// is identical; only the comparison source differs (rune equality
// rather than byte equality). See lcsLenDP godoc for caller contract
// and reasoning behind the explicit if/else row-max selection.
func lcsLenDPRunes(a, b []rune, la, lb int, prev, curr []int) int {
	for j := 1; j <= lb; j++ {
		bj := b[j-1]
		for i := 1; i <= la; i++ {
			if a[i-1] == bj {
				curr[i] = prev[i-1] + 1
			} else {
				if prev[i] > curr[i-1] {
					curr[i] = prev[i]
				} else {
					curr[i] = curr[i-1]
				}
			}
		}
		prev, curr = curr, prev
	}
	return prev[la]
}

// indelRatioRunes returns the Indel-formula similarity over rune
// slices. Same edge-case conventions as indelRatio (byte path):
// both-empty → 1.0, one-empty → 0.0, identity → 1.0.
//
// The denominator uses the SUM OF RUNE LENGTHS (not byte lengths). For
// example, indelRatioRunes(["c","a","f","é"], ["c","a","f","e"]) =
// 2·3/(4+4) = 0.75 — three matching runes ("caf") with total length 4
// runes on each side.
//
// Explicit left-to-right parenthesisation per DET-06 mirrors
// indelRatio.
func indelRatioRunes(a, b []rune) float64 {
	la, lb := len(a), len(b)
	sum := la + lb
	if sum == 0 {
		return 1.0
	}
	if la == 0 || lb == 0 {
		return 0.0
	}
	lcs := lcsLenRunes(a, b)
	numer := 2.0 * float64(lcs)
	denom := float64(sum)
	return numer / denom
}
