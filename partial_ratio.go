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

// partial_ratio.go implements the Partial Ratio similarity for the
// fuzzymatch catalogue. Partial Ratio is the third Indel-formula
// consumer of the shared LCS-subsequence kernel in token_indel.go
// (Wagner-Fischer 1974 J. ACM 21(1):168-173) and the only Phase 6
// algorithm with BOTH byte and rune surfaces per spec lines 609-610.
//
// Unlike TokenSortRatio (plan 06-01) and TokenSetRatio (plan 06-02),
// Partial Ratio does NOT tokenise its inputs — it operates at the
// character level over the entire string. The algorithm slides the
// shorter input across every alignment of the longer input (with three
// distinct iteration regions per the RapidFuzz reference) and returns
// the maximum Indel-formula similarity. The three-region iteration
// (left tail / middle / right tail) is non-obvious — a naive
// single-loop implementation misses Region 1 and Region 3 alignments
// (06-RESEARCH.md Pitfall 3).
//
// Sources:
//
//   - Engineering provenance: RapidFuzz documentation,
//     https://rapidfuzz.github.io/RapidFuzz/Usage/fuzz.html#partial-ratio
//     — the canonical modern reference for the algorithm (per
//     algorithm-correctness-standards "For algorithms without an
//     academic primary source ... cite the canonical modern reference").
//     RapidFuzz in turn descends from SeatGeek's fuzzywuzzy (2014 —
//     superseded by RapidFuzz which fixed several scoring
//     inconsistencies). The three-region iteration pattern + s1_char_set
//     early-skip pattern were transcribed structurally from RapidFuzz's
//     MIT-licensed `fuzz_py.py::_partial_ratio_impl` — see 06-RESEARCH.md
//     Pattern 6 and Example 4 for the verbatim Go structural transcription
//     template. No code was copied — the implementation is a fresh Go
//     transcription from the Python source's logical structure only.
//   - Underlying DP source: Wagner, R. A., & Fischer, M. J. (1974). "The
//     string-to-string correction problem." Journal of the ACM 21(1):
//     168-173 — the LCS-subsequence dynamic-programming recurrence used
//     by the indelRatio / indelRatioRunes kernels.
//   - Indel-formula equivalence: see 06-RESEARCH.md Pattern 3 for the
//     proof that 2·LCS / (|a|+|b|) equals 1 - IndelDistance / (|a|+|b|)
//     where IndelDistance is the Levenshtein distance restricted to
//     insertions and deletions (the RapidFuzz "Indel" similarity).
//
// Algorithm — PartialRatioScore(a, b) (byte path):
//
//  1. Identity short-circuit: if a == b, return 1.0 immediately. This
//     fires BEFORE any byte slicing or charSet construction.
//  2. Both-empty guard: if len(a) == 0 && len(b) == 0, return 1.0
//     (both-empty identity — already covered by step 1, but kept for
//     clarity and parity with the catalogue convention).
//  3. One-empty guard: if len(a) == 0 || len(b) == 0, return 0.0.
//  4. Determine `shorter` / `longer` ([]byte) by length comparison.
//     For unequal-length inputs the swap is value-preserving; for
//     equal-length inputs an additional symmetric tie-break runs the
//     three-region iteration twice with the roles swapped — see the
//     "Equal-length symmetric tie-break" note below.
//  5. Build `s1_char_set` as a [256]bool — set membership of bytes of
//     `shorter`. Used to early-skip alignments whose last (Region 1 or
//     Region 2) or first (Region 3) byte does not appear in `shorter`
//     — such alignments cannot improve the running maximum.
//  6. Region 1 (left tail): for i := 1; i < m; i++ — substrings
//     `longer[:i]` (i = 1..m-1, length 1..m-1). The left edge of the
//     shorter string "hangs off" the start of the longer string. Skip
//     if charSet[longer[i-1]] is false.
//  7. Region 2 (middle): for i := 0; i <= n-m; i++ — substrings
//     `longer[i:i+m]` (length m). The full m-length window slides
//     across the longer string. Skip if charSet[longer[i+m-1]] is
//     false. Early-exit when best == 1.0 (perfect match found).
//  8. Region 3 (right tail): for i := n-m; i < n; i++ — substrings
//     `longer[i:]` (length n-i, decreasing from m down to 1). The
//     right edge of the shorter string "hangs off" the end of the
//     longer string. Skip if charSet[longer[i]] is false.
//     IMPORTANT: when n == m, this region iterates i = 0..n-1
//     (covering the full alignment at i=0 plus all right-suffix
//     alignments) — these would otherwise be missed because Region 2
//     iterates only i = 0 when n-m == 0. Region 3 always runs
//     unconditionally; when n > m there is a single trivial overlap
//     at i = n-m with Region 2 (one redundant indelRatio call,
//     harmless and matches the RapidFuzz reference behaviour).
//  9. Return best.
//
// Algorithm — PartialRatioScoreRunes(a, b) (rune path):
//
// Mirrors the byte path with rune slices and map[rune]struct{} charSet.
// The identity short-circuit fires BEFORE the `[]rune(a)` / `[]rune(b)`
// conversion (saves 2 heap allocations on identical inputs — same
// pattern as `LongestCommonSubstringRunes` in lcsstr.go lines 173-178).
//
// PITFALL 3 — three-region iteration: a naive single-loop
// `for i := 0; i <= n-m; i++` implementation MISSES Regions 1 and 3
// (left and right tails). Fixtures `("abc", "bc") → 1.0` (Region 3
// right-tail wins) and `("abc", "ab") → 1.0` (Region 1 left-tail
// wins) are the KEYSTONE regression gates. Both are pinned in
// (a) unit tests `partial_ratio_test.go`, (b) BDD scenarios
// `tests/bdd/features/partial_ratio.feature` with explicit named
// scenarios "Region 1 left-tail alignment wins" and "Region 3
// right-tail alignment wins", (c) cross-validation corpus entries
// `partial_left_tail_wins` and `partial_right_tail_wins`.
//
// Equal-length symmetric tie-break (matches RapidFuzz
// `partial_ratio_alignment` lines 328-333): when `len(a) == len(b)`
// AND the first pass did not saturate at 1.0, the three-region
// iteration is run a SECOND time with the roles of shorter/longer
// swapped, then the max is taken. Region 1 / Region 3 of the
// three-region implementation are asymmetric in the role of
// `shorter` vs `longer` (they iterate prefixes / suffixes of `longer`,
// never `shorter`); when both inputs are the same length there are
// two valid "shorter, longer" orderings and the optimal alignment
// may live in only one of them. RapidFuzz's wrapper covers this by
// re-invoking the implementation with arguments swapped; we mirror
// that here. This is the LOCKED equal-length symmetry tie-break.
//
// Complexity:
//
//	O(|s|·|l|·max(|s|,|l|))
//
//	where |s| = len(shorter), |l| = len(longer). The three-region
//	iteration calls indelRatio O(|l|) times, each call is O(|s|·|s|)
//	in the worst case (with the inner-loop-over-shorter-side swap in
//	lcsLen). The s1_char_set early-skip prunes alignments whose
//	last/first character does not appear in shorter — this is
//	load-bearing for the pathological budget.
//
// DoS notice:
//
//	On long-vs-short mismatches (~10 chars vs ~10,000 chars) the
//	middle region performs ~10,000 indelRatio calls, each ~100 LCS
//	DP cell updates — the char-set early-skip prunes ~95% of these
//	calls when the shared alphabet is small. In untrusted-input
//	contexts (HTTP request body, file uploads, user-submitted
//	identifiers), pre-validate input length before calling. See
//	BenchmarkPartialRatio_Pathological_LongShortMismatch for measured
//	timings.
//
// Future optimisation (spec-deferred to v1.x):
//
//	TODO(#TBD): implement sliding-window DP per Bachmann RapidFuzz
//	docs — spec line 612 explicitly defers the O(|s|·|l|) sliding-
//	window variant to v1.x. Phase 6 ships the straightforward
//	loop-over-alignments + indelRatio approach with the s1_char_set
//	early-skip. A future GitHub issue will track the sliding-window
//	DP implementation; this TODO will be updated with the issue
//	number once it is created.
//
// Edge cases (mirror Q-Gram Jaccard / LCSStr / TokenSortRatio):
//
//   - PartialRatioScore("",      "")        == 1.0   (both-empty identity; vacuous match)
//   - PartialRatioScore("hello", "hello")   == 1.0   (identity)
//   - PartialRatioScore("hello", "")        == 0.0   (one-empty)
//   - PartialRatioScore("",      "hello")   == 0.0   (one-empty)
//   - PartialRatioScore("abc",   "bc")      == 1.0   (Region 3 right-tail wins — KEYSTONE Pitfall-3 fixture)
//   - PartialRatioScore("abc",   "ab")      == 1.0   (Region 1 left-tail wins — KEYSTONE Pitfall-3 fixture)
//   - PartialRatioScore("YANKEES", "NEW YORK YANKEES") == 1.0 (Region 2 middle wins)
//
// PartialRatio does NOT inherit TokenSetRatio's RapidFuzz issue #110
// deviation. PartialRatio follows the catalogue's standard both-empty
// → 1.0 convention because the algorithm has a meaningful interpretation
// at the character level even when both inputs are empty (the empty
// substring matches the empty string trivially).
//
// Source-origin discipline (per algorithm-licensing-standards):
//
//   - Primary source:        RapidFuzz docs (engineering provenance) +
//                            Wagner & Fischer 1974 (underlying DP).
//   - Cross-validation:      RapidFuzz 3.14.5 via the corpus at
//                            testdata/cross-validation/token-ratios/vectors.json
//                            — every PartialRatio entry asserts byte-stable
//                            agreement within epsilon = 1e-9 for both the
//                            byte and rune surfaces.
//   - Tie-break:             none (the max over float64 ratios; identity
//                            short-circuit + early-exit on best == 1.0
//                            cover the saturation case).
//   - GPL/LGPL provenance:   none.
//   - Code copied verbatim:  none — fresh transcription from RapidFuzz's
//                            `_partial_ratio_impl` Python structure only;
//                            the Indel kernel is this project's own
//                            Wagner-Fischer 1974 implementation in
//                            token_indel.go.
//
// Implementation discipline:
//
//   - NO init()-time table builds (per docs/requirements.md §5(12)).
//   - NO map iteration on output paths (DET-03). The rune charSet is
//     a `map[rune]struct{}` used ONLY for O(1) membership lookups;
//     never iterated.
//   - NO transcendental float operations (DET-06): the score comparisons
//     use `>` and `==`; the indelRatio kernel does the final division
//     with explicit left-to-right parenthesisation.
//   - NO goroutines, channels, or mutexes.
//   - PartialRatio is character-level — NO Tokenise call.
//   - Identity short-circuit `if a == b { return 1.0 }` BEFORE any byte
//     slicing or charSet construction (byte path) and BEFORE `[]rune`
//     conversion (rune path) — saves 2 heap allocations on identical
//     inputs in the rune path (IN-04 closure pattern from Phase 4 /
//     mirrors LongestCommonSubstringRunes lines 173-178).
//   - The byte-path charSet is a stack-allocated `[256]bool`; the
//     rune-path charSet is a `map[rune]struct{}` (cheap — shorter input
//     is typically small).
//   - Inner-loop-over-shorter-side swap is performed via a length
//     comparison so the indelRatio kernel always sees `m <= n`. The
//     swap is value-preserving because the optimal alignment is
//     symmetric.
//
// Public surface (two functions — both surfaces; byte path is dispatched):
//
//   - PartialRatioScore(a, b string) float64        (byte path; dispatched)
//   - PartialRatioScoreRunes(a, b string) float64   (rune path; NOT dispatched)
//
// Only PartialRatioScore is registered in dispatch[AlgoPartialRatio]
// (slot 16 — see algoid.go AlgoPartialRatio). The rune-path
// PartialRatioScoreRunes is public but NOT dispatched (the dispatch
// table signature is the byte-path one — same convention as LCSStr's
// rune variants in lcsstr.go).

package fuzzymatch

// PartialRatioScore returns the Partial Ratio similarity between a and
// b in [0.0, 1.0]: the maximum Indel-formula similarity ratio over
// alignments of the shorter input against substrings of the longer
// input, computed via the three-region iteration (Region 1 left tail /
// Region 2 middle / Region 3 right tail) with the `s1_char_set`
// early-skip optimisation. RapidFuzz-canonical normalisation
// 2·LCS / (|substr_a|+|substr_b|) per RapidFuzz `_partial_ratio_impl`
// reference.
//
// Conventions (mirror Q-Gram Jaccard / LCSStr / TokenSortRatio — does
// NOT inherit TokenSetRatio's RapidFuzz issue #110 deviation):
//
//   - PartialRatioScore("",      "")           == 1.0  (both-empty / identity)
//   - PartialRatioScore("hello", "hello")      == 1.0  (identity)
//   - PartialRatioScore("hello", "")           == 0.0  (one-empty)
//   - PartialRatioScore("",      "hello")      == 0.0  (one-empty)
//   - PartialRatioScore("abc",   "bc")         == 1.0  (Region 3 right-tail wins per Pitfall 3)
//   - PartialRatioScore("abc",   "ab")         == 1.0  (Region 1 left-tail wins per Pitfall 3)
//   - PartialRatioScore("YANKEES", "NEW YORK YANKEES") == 1.0 (Region 2 middle wins)
//
// Symmetric across argument order — PartialRatioScore(a, b) ==
// PartialRatioScore(b, a) — because the shorter-longer swap is
// internal and the indelRatio kernel is symmetric. The byte-equality
// is exact (no float tolerance needed); see
// TestProp_PartialRatioScore_Symmetric for the quick.Check property.
//
// PartialRatio is character-level (no Tokenise call). For multi-byte
// UTF-8 inputs where rune-boundary alignment matters (e.g. comparing
// "café" with "caf"), use PartialRatioScoreRunes instead — the byte
// path would split "é" mid-codepoint and compute a different score.
//
// Reference vector (cross-validated against RapidFuzz 3.14.5):
//
//	PartialRatioScore("YANKEES", "NEW YORK YANKEES") = 1.0
//	PartialRatioScore("abc", "bc")                   = 1.0  (Region 3 right-tail wins)
//	PartialRatioScore("abc", "ab")                   = 1.0  (Region 1 left-tail wins)
//
// Worst-case time: O(|s|·|l|·max(|s|,|l|)) — the three-region iteration
// calls indelRatio O(|l|) times; each call is O(|s|·|s|) in the worst
// case. The s1_char_set early-skip prunes alignments whose last/first
// byte does not appear in shorter; this is load-bearing for the
// pathological budget per RESEARCH.md Pitfall 3. See the file-header
// godoc DoS notice for the explicit failure-mode warning.
//
// Allocation budget: 2 []byte allocations (one per side from
// `string([]byte)` semantics; Go's compiler may elide these on
// short inputs); 1 [256]bool stack-allocated charSet; the indelRatio
// kernel inherits its zero-alloc / two-alloc budget per the
// stack-buffer / heap-fallback split.
func PartialRatioScore(a, b string) float64 {
	// Step 1: identity short-circuit BEFORE any byte slicing or charSet
	// construction. Covers both the literal identical case AND the
	// both-empty case (a == "" == b).
	if a == b {
		return 1.0
	}
	// Step 2: both-empty (already covered by step 1, but kept for
	// clarity / parity with the catalogue convention; this branch is
	// effectively dead because step 1 fires first when a == "" == b).
	if len(a) == 0 && len(b) == 0 {
		return 1.0
	}
	// Step 3: one-empty.
	if len(a) == 0 || len(b) == 0 {
		return 0.0
	}
	// Step 4: shorter / longer determination. The swap is value-
	// preserving because the optimal alignment is symmetric in (a, b)
	// for unequal-length inputs.
	ab := []byte(a)
	bb := []byte(b)
	var shorter, longer []byte
	if len(ab) <= len(bb) {
		shorter, longer = ab, bb
	} else {
		shorter, longer = bb, ab
	}
	res := partialRatioThreeRegionMax(shorter, longer)
	// Equal-length symmetric tie-break (matches RapidFuzz
	// `partial_ratio_alignment` lines 328-333): when len1 == len2 AND
	// the first pass did not saturate at 1.0, run the algorithm again
	// with the role of shorter/longer swapped, then take the max.
	// Region 1 / Region 3 of `_partial_ratio_impl` are asymmetric in
	// the role of `shorter` vs `longer` (they iterate prefixes /
	// suffixes of `longer`, never `shorter`); when both inputs are
	// the same length there are two valid "shorter, longer"
	// orderings and the optimal alignment may live in only one of
	// them. RapidFuzz's wrapper covers this by re-invoking the
	// implementation with arguments swapped; we mirror that here.
	if res < 1.0 && len(ab) == len(bb) {
		if r := partialRatioThreeRegionMax(longer, shorter); r > res {
			res = r
		}
	}
	return res
}

// partialRatioThreeRegionMax runs the three-region iteration over
// (shorter, longer) byte slices with the s1_char_set early-skip and
// returns the maximum Indel-formula similarity across all alignments.
//
// Caller guarantees: len(shorter) > 0; len(longer) > 0;
// len(shorter) <= len(longer). Extracted from PartialRatioScore to
// keep that function under the gocyclo ceiling (mirrors the
// helper-extraction pattern from token_set_ratio.go's
// buildTokenSetPartitions / tokenSetThreeWayMax — see plan 06-02
// SUMMARY.md deviation #3). Each region is further factored into a
// per-region helper to keep the orchestrator under gocyclo=10.
func partialRatioThreeRegionMax(shorter, longer []byte) float64 {
	m, n := len(shorter), len(longer)

	// Build s1_char_set for early-skip. Stack-allocated [256]bool —
	// zero heap allocation regardless of input length.
	var charSet [256]bool
	for _, ch := range shorter {
		charSet[ch] = true
	}

	best := partialRatioRegion1Bytes(shorter, longer, m, &charSet, 0.0)
	var perfect bool
	best, perfect = partialRatioRegion2Bytes(shorter, longer, m, n, &charSet, best)
	if perfect {
		return 1.0
	}
	best = partialRatioRegion3Bytes(shorter, longer, m, n, &charSet, best)
	return best
}

// partialRatioRegion1Bytes iterates Region 1 (left tail) of the byte
// path: substrings `longer[:i]` for i = 1..m-1. Skip if
// charSet[longer[i-1]] is false (the last byte of the candidate
// substring is not in shorter — that alignment cannot produce a match
// at the right edge).
func partialRatioRegion1Bytes(shorter, longer []byte, m int, charSet *[256]bool, best float64) float64 {
	for i := 1; i < m; i++ {
		if !charSet[longer[i-1]] {
			continue
		}
		if r := indelRatio(shorter, longer[:i]); r > best {
			best = r
		}
	}
	return best
}

// partialRatioRegion2Bytes iterates Region 2 (middle) of the byte
// path: full m-length windows `longer[i:i+m]` for i = 0..n-m. Skip if
// charSet[longer[i+m-1]] is false. Returns (best, perfect) where
// perfect == true signals best == 1.0 (caller should early-exit
// without iterating Region 3).
func partialRatioRegion2Bytes(shorter, longer []byte, m, n int, charSet *[256]bool, best float64) (float64, bool) {
	for i := 0; i <= n-m; i++ {
		if !charSet[longer[i+m-1]] {
			continue
		}
		if r := indelRatio(shorter, longer[i:i+m]); r > best {
			best = r
		}
		if best == 1.0 {
			return best, true
		}
	}
	return best, false
}

// partialRatioRegion3Bytes iterates Region 3 (right tail) of the byte
// path: substrings `longer[i:]` for i = n-m..n-1. Skip if
// charSet[longer[i]] is false (the first byte of the candidate
// substring is not in shorter). When n == m this iterates i = 0..n-1
// (covering both the full alignment at i=0 and the right-suffix
// alignments at i=1..n-1) — these would otherwise be missed because
// Region 2 evaluates only i=0. When n > m there is a single trivial
// overlap at i = n-m with Region 2 (one redundant indelRatio call;
// harmless and matches the RapidFuzz reference behaviour).
func partialRatioRegion3Bytes(shorter, longer []byte, m, n int, charSet *[256]bool, best float64) float64 {
	for i := n - m; i < n; i++ {
		if !charSet[longer[i]] {
			continue
		}
		if r := indelRatio(shorter, longer[i:]); r > best {
			best = r
		}
	}
	return best
}

// PartialRatioScoreRunes is the rune-path variant of PartialRatioScore.
// It treats each input as a sequence of Unicode code points (runes)
// rather than bytes, so multi-byte UTF-8 sequences are compared
// atomically. For example, PartialRatioScoreRunes("café", "caf") = 1.0
// (the 3-rune subsequence of the 4-rune input matches "caf" exactly);
// the byte-path equivalent would compute differently because "café"
// occupies 5 bytes while "caf" occupies 3 bytes.
//
// The rune variant allocates two []rune slices (one per side) and one
// map[rune]struct{} for the charSet. The identity short-circuit fires
// BEFORE `[]rune(a)` / `[]rune(b)` conversion, saving 2 heap
// allocations on identical inputs — same pattern as
// LongestCommonSubstringRunes (lcsstr.go lines 173-178).
//
// The leftmost-alignment semantic and edge-case conventions are
// identical to PartialRatioScore.
//
// PartialRatioScoreRunes is NOT registered in the dispatch table
// (dispatch[AlgoPartialRatio] holds the byte-path PartialRatioScore
// only — dispatch table signature is byte-path).
//
// Reference vector:
//
//	PartialRatioScoreRunes("café", "caf") = 1.0
//	  (3-rune subsequence "caf" matches the entirety of "caf")
//	PartialRatioScoreRunes("café", "café") = 1.0  (identity)
//
// Worst-case time: O(|s|·|l|·max(|s|,|l|)) over rune counts (NOT byte
// counts). Allocation budget: 2 []rune allocations + 1 map[rune]
// struct{} + indelRatioRunes kernel allocations.
func PartialRatioScoreRunes(a, b string) float64 {
	// Identity short-circuit BEFORE []rune conversion — saves 2 heap
	// allocations on identical inputs. Mirrors LongestCommonSubstringRunes
	// in lcsstr.go.
	if a == b {
		return 1.0
	}
	if len(a) == 0 && len(b) == 0 {
		return 1.0
	}
	if len(a) == 0 || len(b) == 0 {
		return 0.0
	}
	ra := []rune(a) // 1 alloc
	rb := []rune(b) // 1 alloc
	var shorter, longer []rune
	if len(ra) <= len(rb) {
		shorter, longer = ra, rb
	} else {
		shorter, longer = rb, ra
	}
	res := partialRatioThreeRegionMaxRunes(shorter, longer)
	// Equal-length symmetric tie-break (matches RapidFuzz behaviour;
	// see PartialRatioScore for the rationale and citation).
	if res < 1.0 && len(ra) == len(rb) {
		if r := partialRatioThreeRegionMaxRunes(longer, shorter); r > res {
			res = r
		}
	}
	return res
}

// partialRatioThreeRegionMaxRunes is the rune-slice variant of
// partialRatioThreeRegionMax. The three-region iteration is identical;
// only the comparison source and the charSet container differ.
//
// Caller guarantees: len(shorter) > 0; len(longer) > 0;
// len(shorter) <= len(longer). Extracted from PartialRatioScoreRunes
// to keep that function under the gocyclo ceiling (mirrors the
// helper-extraction pattern in token_set_ratio.go). Each region is
// further factored into a per-region helper to keep the orchestrator
// under gocyclo=10.
func partialRatioThreeRegionMaxRunes(shorter, longer []rune) float64 {
	m, n := len(shorter), len(longer)

	// Rune charSet — map[rune]struct{} for O(1) membership lookups.
	// Map is queried (never iterated) on the output path so DET-03 is
	// preserved. Capacity hint = m avoids growth allocations.
	charSet := make(map[rune]struct{}, m)
	for _, r := range shorter {
		charSet[r] = struct{}{}
	}

	best := partialRatioRegion1Runes(shorter, longer, m, charSet, 0.0)
	var perfect bool
	best, perfect = partialRatioRegion2Runes(shorter, longer, m, n, charSet, best)
	if perfect {
		return 1.0
	}
	best = partialRatioRegion3Runes(shorter, longer, m, n, charSet, best)
	return best
}

// partialRatioRegion1Runes iterates Region 1 (left tail) of the rune
// path: substrings `longer[:i]` for i = 1..m-1.
func partialRatioRegion1Runes(shorter, longer []rune, m int, charSet map[rune]struct{}, best float64) float64 {
	for i := 1; i < m; i++ {
		if _, ok := charSet[longer[i-1]]; !ok {
			continue
		}
		if r := indelRatioRunes(shorter, longer[:i]); r > best {
			best = r
		}
	}
	return best
}

// partialRatioRegion2Runes iterates Region 2 (middle) of the rune
// path: full m-length windows `longer[i:i+m]` for i = 0..n-m. Returns
// (best, perfect) where perfect == true signals best == 1.0.
func partialRatioRegion2Runes(shorter, longer []rune, m, n int, charSet map[rune]struct{}, best float64) (float64, bool) {
	for i := 0; i <= n-m; i++ {
		if _, ok := charSet[longer[i+m-1]]; !ok {
			continue
		}
		if r := indelRatioRunes(shorter, longer[i:i+m]); r > best {
			best = r
		}
		if best == 1.0 {
			return best, true
		}
	}
	return best, false
}

// partialRatioRegion3Runes iterates Region 3 (right tail) of the rune
// path: substrings `longer[i:]` for i = n-m..n-1. When n == m this
// iterates i = 0..n-1; when n > m there is a single trivial overlap
// at i = n-m with Region 2. See partialRatioRegion3Bytes for the full
// rationale.
func partialRatioRegion3Runes(shorter, longer []rune, m, n int, charSet map[rune]struct{}, best float64) float64 {
	for i := n - m; i < n; i++ {
		if _, ok := charSet[longer[i]]; !ok {
			continue
		}
		if r := indelRatioRunes(shorter, longer[i:]); r > best {
			best = r
		}
	}
	return best
}
