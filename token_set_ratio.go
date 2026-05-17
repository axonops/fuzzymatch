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

// token_set_ratio.go implements the Token Set Ratio similarity for the
// fuzzymatch catalogue. Token Set Ratio is the second Indel-formula
// consumer of the shared LCS-subsequence kernel in token_indel.go
// (Wagner-Fischer 1974 J. ACM 21(1):168-173) — the most algorithmically
// complex of the three Indel-based ratios in Phase 6.
//
// Sources:
//
//   - Engineering provenance: RapidFuzz documentation,
//     https://rapidfuzz.github.io/RapidFuzz/Usage/fuzz.html#token-set-ratio
//     — the canonical modern reference for the algorithm (per
//     algorithm-correctness-standards "For algorithms without an
//     academic primary source ... cite the canonical modern reference").
//     RapidFuzz in turn descends from SeatGeek's fuzzywuzzy (2014 —
//     superseded by RapidFuzz which fixed several scoring
//     inconsistencies; the empty-set-returns-0.0 behaviour is preserved
//     bug-for-bug per RapidFuzz issue #110 — see DEVIATION below).
//   - Underlying DP source: Wagner, R. A., & Fischer, M. J. (1974). "The
//     string-to-string correction problem." Journal of the ACM 21(1):
//     168-173 — the LCS-subsequence dynamic-programming recurrence used
//     by the indelRatio kernel.
//   - Indel-formula equivalence: see 06-RESEARCH.md Pattern 3 for the
//     proof that 2·LCS / (|a|+|b|) equals 1 - IndelDistance / (|a|+|b|)
//     where IndelDistance is the Levenshtein distance restricted to
//     insertions and deletions (the RapidFuzz "Indel" similarity).
//
// Algorithm — TokenSetRatioScore(a, b):
//
//   1. Empty-input gate: if a == "" or b == "" (including
//      ("", "")), return 0.0 immediately. This is the LOCKED
//      bug-for-bug RapidFuzz issue #110 / fuzzywuzzy parity — the
//      catalogue's standard both-empty → 1.0 convention is overridden
//      for TokenSetRatio per the DEVIATION below. The gate fires
//      BEFORE the identity short-circuit because the deviation
//      requires ("", "") → 0.0, not 1.0.
//   2. Identity short-circuit (non-empty only): if a == b AND a != "",
//      return 1.0 immediately. This avoids the Tokenise allocation
//      on identical inputs.
//   3. Tokenise both sides using DefaultTokeniseOptions() — see OQ-1
//      RESOLUTION below for the tokeniser-divergence note.
//   4. Empty-token-set DEVIATION (post-Tokenise): if EITHER Tokenise
//      output is empty (zero tokens), return 0.0 — catches the case
//      where the raw strings are non-empty but tokenise to nothing
//      (e.g. (" ", "  ")). The pre-Tokenise empty-input gate above
//      catches the cheaper case where a literal empty string is
//      passed.
//   5. Build deduplicated sorted slices: intersectKeys (tokens in BOTH
//      sets), diffABKeys (tokens in A but not B), diffBAKeys (tokens in
//      B but not A). Each slice is sorted byte-lex ascending so all
//      downstream joins are deterministic.
//   6. Subset short-circuit: if intersectKeys is non-empty AND
//      (diffABKeys is empty OR diffBAKeys is empty) — i.e. one token set
//      is a subset of the other — return 1.0 directly without computing
//      the three ratios.
//   7. Three-way max construction. Join each sorted slice with a single
//      ASCII space:
//        sortedSect      = strings.Join(intersectKeys, " ")
//        sortedDiffAB    = strings.Join(diffABKeys,    " ")
//        sortedDiffBA    = strings.Join(diffBAKeys,    " ")
//      Build the two combined-and-sorted-with-intersection strings:
//        combined1to2 = sortedSect (+ " " + sortedDiffAB if both non-empty,
//                                                       else concatenation)
//        combined2to1 = sortedSect (+ " " + sortedDiffBA if both non-empty,
//                                                       else concatenation)
//      Compute three Indel ratios and take the max:
//        r1 = indelRatio(sortedSect,   combined1to2)  // intersection vs intersection+diff_ab
//        r2 = indelRatio(sortedSect,   combined2to1)  // intersection vs intersection+diff_ba
//        r3 = indelRatio(combined1to2, combined2to1)  // intersection+diff_ab vs intersection+diff_ba
//        return max(r1, r2, r3)
//      The max is taken via explicit if-chain (NOT builtin `max`) to
//      mirror the determinism-reviewer-auditable pattern from
//      token_indel.go (qgram_jaccard.go).
//
// DEVIATION from the catalogue's both-empty-→-1.0 convention (LOCKED
// in plan 06-02): TokenSetRatioScore returns 0.0 when EITHER Tokenise
// output is zero-length — even when both inputs would otherwise be
// considered identical (e.g. both empty strings, both pure-separator
// strings). This is bug-for-bug compatibility with RapidFuzz's
// token_set_ratio (rapidfuzz issue #110) which itself mirrors
// fuzzywuzzy. Other tokenised algorithms in the catalogue
// (TokenJaccard, MongeElkan) follow the standard both-empty → 1.0
// convention; TokenSetRatio is the documented exception. The
// deviation is necessary because the algorithm's three-way construction
// has no meaningful interpretation when there are no tokens to
// intersect.
//
// Note: the identity short-circuit `if a == b { return 1.0 }` fires
// ONLY for non-empty identical strings — so TokenSetRatioScore(
// "hello", "hello") returns 1.0. TokenSetRatioScore("", "") returns
// 0.0 because the empty-input gate (which is checked FIRST) fires
// before the identity short-circuit. The post-Tokenise empty-set
// gate handles the remaining edge: raw strings non-empty but both
// tokenise to empty (e.g. (" ", "  ")) → 0.0.
//
// Three-way max construction note (LOCKED in plan 06-02): the third
// branch is `indelRatio(combined1to2, combined2to1)` — the LCS between
// `intersection + " " + diffAB` and `intersection + " " + diffBA` —
// NOT `indelRatio(diffAB, diffBA)` (which would be just the
// differences). The first two branches (r1, r2) reduce analytically to
// 2·sectLen / (sectLen + sectLen + separator + diffLen) because the
// intersection is a prefix of each combined string and the LCS equals
// sectLen exactly. Hence the implementation could substitute integer
// arithmetic for r1 and r2, but the code computes them via indelRatio
// for clarity and DRY (the analytical-vs-DP equivalence is a property
// test in props_test.go, not a code-side optimisation).
//
// OQ-1 RESOLUTION (tokeniser-divergence handling — LOCKED in plan
// 06-01): RapidFuzz tokenises via Python `str.split()` —
// whitespace-only, case-preserving. fuzzymatch's `Tokenise(s,
// DefaultTokeniseOptions())` is camelCase / snake_case / kebab-case /
// dot-case aware AND lowercasing. For inputs without identifier-style
// boundaries (pure whitespace-separated lowercase ASCII text), the two
// tokenisations agree and the scores match. For inputs with mixed
// identifier styles — e.g. "userID" vs "user_id" — the project
// tokenisation produces semantically richer splits ([user, id] vs [user,
// i, d] under RapidFuzz's str.split which leaves "userID" as one token).
// The cross-validation corpus at
// testdata/cross-validation/token-ratios/vectors.json is restricted to
// whitespace-only lowercase ASCII inputs so cross-validation against
// RapidFuzz 3.14.5 is byte-stable.
//
// Source-origin discipline (per algorithm-licensing-standards):
//
//   - Primary source:        RapidFuzz docs (engineering provenance) +
//                            Wagner & Fischer 1974 (underlying DP).
//   - Cross-validation:      RapidFuzz 3.14.5 via the corpus at
//                            testdata/cross-validation/token-ratios/vectors.json
//                            — every TokenSet entry asserts byte-stable
//                            agreement within epsilon = 1e-9.
//   - Tie-break:             none (Tokenise is deterministic;
//                            sort.Strings is stable byte-lex; the three
//                            ratios are symmetric in their respective
//                            argument orders — see
//                            TestProp_TokenSetRatioScore_Symmetric).
//   - GPL/LGPL provenance:   none.
//   - Code copied verbatim:  none — fresh transcription from RapidFuzz's
//                            fuzz_py.py token_set_ratio Python source
//                            structure only; the Indel kernel is this
//                            project's own Wagner-Fischer 1974
//                            implementation in token_indel.go.
//
// Implementation discipline:
//
//   - NO init()-time table builds (per docs/requirements.md §5(12)).
//   - NO map iteration on output paths (DET-03). The intersection /
//     difference key slices are populated by iterating tokens slices
//     (sorted output order is independent of map iteration order
//     because we sort.Strings the result before any consumption). The
//     map[string]struct{} sets are used only for O(1) membership tests;
//     no string is built from set iteration.
//   - NO transcendental float operations (DET-06): integer arithmetic
//     in the kernel plus one final division per indelRatio call with
//     explicit left-to-right parenthesisation.
//   - NO goroutines, channels, or mutexes.
//   - Identity short-circuit `if a == b { return 1.0 }` BEFORE Tokenise
//     to avoid the Tokenise allocation on identical inputs.
//   - Three-way max via explicit if-chain (NOT builtin `max`) to
//     mirror the determinism-reviewer-auditable pattern from
//     token_indel.go and qgram_jaccard.go.
//   - NO public *Runes variant — Tokenise handles UTF-8 internally per
//     06-CONTEXT.md §6 LOCKED.
//
// Public surface (one function — the dispatched byte-path score):
//
//   - TokenSetRatioScore(a, b string) float64
//
// Registered in dispatch table slot AlgoTokenSetRatio (slot 15 — see
// algoid.go). The dispatch table maps AlgoID to (a, b string) float64
// with no place for parameters; TokenSetRatio has no parameters so the
// wrapper is the function value directly (no closure needed — mirrors
// dispatch_lcsstr.go and dispatch_token_sort_ratio.go).
//
// Complexity:
//
//   O((|sorted_combined_1to2| + |sorted_combined_2to1|)^2)
//
//   where |sorted_combined_1to2| / |sorted_combined_2to1| are the
//   lengths of the sorted-joined intersection-plus-diff strings. Worst
//   case is asymmetric set cardinalities — a small intersection plus
//   large differences on both sides drives the indelRatio cost on the
//   third (combined-vs-combined) branch. The first two branches
//   (intersection vs intersection-plus-diff) are O(|intersection| ·
//   |combined|) but the LCS equals |intersection| exactly (prefix
//   match), so the inner loop terminates early on the contiguous-prefix
//   path.
//
// DoS notice:
//
//   On inputs where Tokenise produces > 100 tokens per side with a
//   small intersection (≤ 5 tokens) and large diffs (~95 tokens each),
//   the combined-vs-combined indelRatio call performs ~10^4 LCS DP
//   cell updates. In untrusted-input contexts (HTTP request body, file
//   uploads, user-submitted identifiers), pre-validate token-count
//   ceilings before calling. See
//   BenchmarkTokenSetRatio_Pathological_AsymmetricSetCardinalities for
//   measured timings on representative pathological shapes.

package fuzzymatch

import (
	"sort"
	"strings"
)

// TokenSetRatioScore returns the Token Set Ratio similarity between a
// and b in [0.0, 1.0]: tokenise both sides using DefaultTokeniseOptions(),
// compute the sorted intersection, sorted diff_a_to_b, and sorted
// diff_b_to_a token slices, then take the maximum of three Indel ratios
// over the constructed-string forms (intersection vs intersection+diff_ab,
// intersection vs intersection+diff_ba, intersection+diff_ab vs
// intersection+diff_ba). See the file-header godoc for the full
// algorithm and the LOCKED RapidFuzz issue #110 empty-set deviation.
//
// For programmatic input-quality checks before scoring (including
// WarnNoTokensAfterNormalise scoped to AlgoTokenSetRatio),
// see [fuzzymatch.Validate].
//
// Conventions:
//
//   - TokenSetRatioScore("hello",       "hello")            == 1.0  (identity short-circuit; non-empty)
//   - TokenSetRatioScore("",            "")                 == 0.0  (DEVIATION — RapidFuzz issue #110;
//     empty-input returns 0.0, NOT 1.0, before identity check fires)
//   - TokenSetRatioScore("hello",       "")                 == 0.0  (one-empty)
//   - TokenSetRatioScore("",            "hello")            == 0.0  (one-empty)
//   - TokenSetRatioScore(" ",           "  ")               == 0.0  (DEVIATION — both Tokenise to [];
//     post-Tokenise empty-set gate)
//   - TokenSetRatioScore("alpha beta",  "alpha beta gamma") == 1.0  (subset short-circuit: A ⊆ B)
//   - TokenSetRatioScore("alpha beta gamma", "alpha beta")  == 1.0  (subset short-circuit: B ⊆ A)
//   - TokenSetRatioScore("alpha beta",  "beta alpha")       == 1.0  (set equality after dedup+sort)
//
// Symmetric across argument order — TokenSetRatioScore(a, b) ==
// TokenSetRatioScore(b, a) — because Tokenise is deterministic, set
// construction is order-independent, and the three branches are
// individually symmetric (r1 and r2 are paired symmetric variants and
// the max(r1, r2, r3) operator is order-insensitive; r3 is symmetric
// in its argument order via indelRatio's own symmetry). See
// TestProp_TokenSetRatioScore_Symmetric for the quick.Check property.
//
// Tokeniser divergence from RapidFuzz (OQ-1 RESOLUTION LOCKED — see
// file-header godoc): fuzzymatch's Tokenise is identifier-aware
// (camelCase / snake_case / etc.), unlike RapidFuzz's
// whitespace-only Python str.split. For whitespace-only lowercase ASCII
// inputs the two agree; for identifier-style inputs the project
// tokenisation produces semantically richer splits.
//
// Empty-input / empty-set DEVIATION (LOCKED — see file-header
// godoc): if either input is empty OR either Tokenise output is
// empty, the function returns 0.0 (NOT 1.0) per RapidFuzz issue
// #110 bug-for-bug compatibility. The empty-input gate fires BEFORE
// the identity short-circuit so TokenSetRatioScore("", "") returns
// 0.0 (matching RapidFuzz) — NOT 1.0 (the identity short-circuit
// only fires for non-empty identical strings).
//
// Reference vector (cross-validated against RapidFuzz 3.14.5):
//
//	TokenSetRatioScore("hello world", "world peace") = 7/11 ≈ 0.6364
//	  (the three-way max where the combined-vs-combined branch wins:
//	   ratio("hello world", "peace world") = 14/22 = 7/11)
//
// Worst-case time: O((|combined1to2|+|combined2to1|)^2) for the third
// indelRatio call (the combined-vs-combined branch). The first two
// branches reduce to O(|intersection| · |combined|) thanks to the
// contiguous-prefix LCS structure.
//
// This function operates on bytes (the joined sorted strings are
// compared byte-by-byte by the LCS-subsequence DP). For multi-byte
// UTF-8 token contents, Tokenise still produces well-formed UTF-8
// tokens; the byte-level Indel kernel operates on those bytes
// directly. There is no rune-path variant: Tokenise is UTF-8-aware so
// the rune semantic is already preserved at the tokenisation layer.
func TokenSetRatioScore(a, b string) float64 {
	// Empty-input fast path: at least one input is empty → return
	// 0.0 directly. This matches RapidFuzz's behaviour bit-for-bit
	// per issue #110, INCLUDING the both-empty case ("", "") which
	// RapidFuzz returns 0.0 for despite the inputs being trivially
	// identical (fuzzywuzzy parity). This gate fires BEFORE the
	// identity short-circuit because Indel's both-empty → 1.0
	// convention is overridden in TokenSetRatio by the LOCKED
	// deviation.
	if a == "" || b == "" {
		return 0.0
	}
	// Identity short-circuit for non-empty identical strings —
	// avoids the Tokenise allocation on identical inputs. The empty
	// case is handled above (a == b == "" returns 0.0 via the
	// deviation, not 1.0 via identity).
	if a == b {
		return 1.0
	}
	tokensA := Tokenise(a, DefaultTokeniseOptions())
	tokensB := Tokenise(b, DefaultTokeniseOptions())
	// LOCKED DEVIATION per RapidFuzz issue #110 / fuzzywuzzy compat:
	// when EITHER Tokenise output is empty (zero tokens), return 0.0
	// (NOT 1.0). This catches the case where the raw strings are
	// non-empty but tokenise to nothing — e.g. (" ", "  ") — where
	// the empty-input gate above would not fire.
	if len(tokensA) == 0 || len(tokensB) == 0 {
		return 0.0
	}

	intersectKeys, diffABKeys, diffBAKeys := buildTokenSetPartitions(tokensA, tokensB)

	// Subset short-circuit (RESEARCH.md Pattern 5 critical landmine 2):
	// when the intersection is non-empty AND one of the diffs is empty
	// — i.e. one token set is a subset of the other — the algorithm
	// returns 1.0 directly. This matches RapidFuzz's behaviour exactly
	// (`if intersect and (not diff_ab or not diff_ba): return 100`).
	if len(intersectKeys) > 0 && (len(diffABKeys) == 0 || len(diffBAKeys) == 0) {
		return 1.0
	}

	return tokenSetThreeWayMax(intersectKeys, diffABKeys, diffBAKeys)
}

// tokenSetThreeWayMax computes the three-way max over the three
// Indel-ratio branches that define Token Set Ratio. Extracted out of
// TokenSetRatioScore so the parent function stays under the
// cyclomatic-complexity ceiling.
//
//	r1: intersection vs intersection+diff_ab    (analytical: 2·sect/(sect+sect_ab))
//	r2: intersection vs intersection+diff_ba    (analytical: 2·sect/(sect+sect_ba))
//	r3: intersection+diff_ab vs intersection+diff_ba
//
// Branches 1 and 2 reduce to a closed-form expression because the
// intersection is a contiguous prefix of the combined string, so
// LCS(sortedSect, combined) == len(sortedSect). The code uses
// indelRatio for both branches to keep the Indel-kernel surface
// area small and the algorithm reviewer-auditable.
//
// The max is taken via explicit if-chain (NOT builtin `max`) to
// mirror the determinism-reviewer-auditable pattern from
// token_indel.go (qgram_jaccard.go).
func tokenSetThreeWayMax(intersectKeys, diffABKeys, diffBAKeys []string) float64 {
	sortedSect := strings.Join(intersectKeys, " ")
	sortedDiffAB := strings.Join(diffABKeys, " ")
	sortedDiffBA := strings.Join(diffBAKeys, " ")
	combined1to2 := joinSectAndDiff(sortedSect, sortedDiffAB)
	combined2to1 := joinSectAndDiff(sortedSect, sortedDiffBA)
	r1 := indelRatio([]byte(sortedSect), []byte(combined1to2))
	r2 := indelRatio([]byte(sortedSect), []byte(combined2to1))
	r3 := indelRatio([]byte(combined1to2), []byte(combined2to1))
	best := r1
	if r2 > best {
		best = r2
	}
	if r3 > best {
		best = r3
	}
	return best
}

// joinSectAndDiff returns the RapidFuzz-canonical concatenation of the
// sorted intersection and a sorted diff slice. When BOTH sides are
// non-empty, a single ASCII space separates them; otherwise the
// concatenation degrades to whichever side is non-empty (or "" if both
// are empty — that case never arises in practice because the empty-set
// gate and subset short-circuit fire first).
//
// This helper exists so the conditional separator logic is in one
// place and the call sites in TokenSetRatioScore stay readable.
func joinSectAndDiff(sect, diff string) string {
	switch {
	case sect == "":
		return diff
	case diff == "":
		return sect
	default:
		return sect + " " + diff
	}
}

// buildTokenSetPartitions returns the sorted deduplicated key slices
// for the intersection, diff_a_to_b, and diff_b_to_a of two token
// slices.
//
// The intersection / diff_a_to_b sweep iterates tokensA once,
// classifying each unique token into intersection (if present in B)
// or diff_ab (otherwise). The diff_b_to_a sweep iterates tokensB once,
// classifying each unique token absent from A. Each result slice is
// sorted byte-lex ascending so all downstream joins are deterministic
// byte-for-byte across all four CI platforms.
//
// DET-03 compliance: although the helper uses map[string]struct{} for
// O(1) membership and dedup tests, the OUTPUT slices are built by
// iterating the input slices (deterministic order) and SORTED before
// return — no string is built from set iteration. The function
// satisfies the no-map-iteration-on-output-paths rule.
//
// Capacity hints come from the input slices' lengths (upper bounds on
// unique-token counts), avoiding map growth on the typical < 50-token
// fast path.
func buildTokenSetPartitions(tokensA, tokensB []string) (intersectKeys, diffABKeys, diffBAKeys []string) {
	setA := make(map[string]struct{}, len(tokensA))
	for _, t := range tokensA {
		setA[t] = struct{}{}
	}
	setB := make(map[string]struct{}, len(tokensB))
	for _, t := range tokensB {
		setB[t] = struct{}{}
	}
	intersectKeys = make([]string, 0, len(tokensA))
	diffABKeys = make([]string, 0, len(tokensA))
	seenIntersect := make(map[string]struct{}, len(tokensA))
	seenDiffAB := make(map[string]struct{}, len(tokensA))
	for _, t := range tokensA {
		if _, inB := setB[t]; inB {
			if _, seen := seenIntersect[t]; !seen {
				seenIntersect[t] = struct{}{}
				intersectKeys = append(intersectKeys, t)
			}
			continue
		}
		if _, seen := seenDiffAB[t]; !seen {
			seenDiffAB[t] = struct{}{}
			diffABKeys = append(diffABKeys, t)
		}
	}
	diffBAKeys = make([]string, 0, len(tokensB))
	seenDiffBA := make(map[string]struct{}, len(tokensB))
	for _, t := range tokensB {
		if _, inA := setA[t]; inA {
			continue
		}
		if _, seen := seenDiffBA[t]; !seen {
			seenDiffBA[t] = struct{}{}
			diffBAKeys = append(diffBAKeys, t)
		}
	}
	sort.Strings(intersectKeys)
	sort.Strings(diffABKeys)
	sort.Strings(diffBAKeys)
	return intersectKeys, diffABKeys, diffBAKeys
}
