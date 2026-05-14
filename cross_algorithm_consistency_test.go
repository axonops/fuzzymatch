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

// cross_algorithm_consistency_test.go pins relationships across the seven
// Phase 2 + Phase 3 character-based algorithms:
//
//   - DIVERGENCE: DL-OSA and DL-Full produce DIFFERENT distances for
//     the canonical discriminating vector "ca"/"abc" (3 vs 2). This
//     is a load-bearing claim of the phase (Phase 2 ROADMAP success
//     criterion #2); the file pins it cross-algorithm rather than only
//     inside each algorithm's own _test.go file.
//
//   - LOCAL-VS-GLOBAL DIVERGENCE: Smith-Waterman-Gotoh (local
//     alignment) scores STRICTLY HIGHER than Levenshtein (global edit
//     distance) on a substring-containment input. This is the
//     load-bearing Phase 3 claim: SWG finds the substring while
//     Levenshtein counts every uncovered position as an edit.
//
//   - CONVERGENCE: every algorithm's score on identical inputs is
//     exactly 1.0 (identity); on both-empty inputs is 1.0; on one-empty
//     inputs is 0.0. Pinned across all seven algorithms.
//
//   - SINGLE-SUBSTITUTION AGREEMENT: Levenshtein, DL-OSA, DL-Full,
//     and Hamming all return distance == 1 for a single-character
//     substitution between equal-length strings (e.g. "a"/"b"). SWG
//     is excluded — it has no Distance variant (CONTEXT.md §7 / IN-06).
//
// Stdlib `testing` only — no testify in root tests, per
// .claude/skills/go-coding-standards.

package fuzzymatch_test

import (
	"math"
	"testing"

	"github.com/axonops/fuzzymatch"
)

// TestCrossAlgorithm_OSA_Full_Divergence asserts the canonical discriminating
// vector that distinguishes DL-OSA from DL-Full in a single test body:
//
//   - DamerauLevenshteinOSADistance("ca", "abc") == 3
//   - DamerauLevenshteinFullDistance("ca", "abc") == 2
//
// This divergence is the proof-of-correctness contract for Phase 2
// (ROADMAP success criterion #2). The OSA restriction forbids re-editing
// characters after a transposition (Boytsov 2011 §3.1 / Damerau 1964);
// Full DL (Lowrance-Wagner 1975) allows unrestricted transpositions with
// no re-edit restriction, yielding a smaller distance on this pair.
//
// If this test fails with got_osa==2, the OSA recurrence has collapsed
// into Full DL semantics. If it fails with got_full==3, the Full DL
// recurrence is applying OSA's restriction.
func TestCrossAlgorithm_OSA_Full_Divergence(t *testing.T) {
	// Boytsov 2011 §3.1 / Damerau 1964 discriminating vector:
	// "ca" vs "abc" must return 3 under OSA and 2 under Full DL.
	gotOSA := fuzzymatch.DamerauLevenshteinOSADistance("ca", "abc")
	if gotOSA != 3 {
		t.Errorf("DamerauLevenshteinOSADistance(\"ca\",\"abc\") = %d, want 3 (Boytsov 2011 §3.1)", gotOSA)
	}

	gotFull := fuzzymatch.DamerauLevenshteinFullDistance("ca", "abc")
	if gotFull != 2 {
		t.Errorf("DamerauLevenshteinFullDistance(\"ca\",\"abc\") = %d, want 2 (Lowrance-Wagner 1975)", gotFull)
	}

	// Cross-check the divergence itself: OSA must be strictly greater than Full
	// for this pair (the discriminating-vector property).
	if gotOSA == gotFull {
		t.Errorf("OSA and Full DL agree on \"ca\"/\"abc\" (both = %d); expected divergence (OSA=3 != Full=2)", gotOSA)
	}
}

// TestCrossAlgorithm_IdentityConvergence asserts that all seven Phase 2 + 3
// algorithms return Score(input, input) == 1.0 for a non-empty identical
// input. This is the identity invariant — a fundamental requirement of any
// correct similarity function.
func TestCrossAlgorithm_IdentityConvergence(t *testing.T) {
	type scoreFunc struct {
		name string
		fn   func(a, b string) float64
	}

	funcs := []scoreFunc{
		{"LevenshteinScore", fuzzymatch.LevenshteinScore},
		{"DamerauLevenshteinOSAScore", fuzzymatch.DamerauLevenshteinOSAScore},
		{"DamerauLevenshteinFullScore", fuzzymatch.DamerauLevenshteinFullScore},
		{"HammingScore", fuzzymatch.HammingScore},
		{"JaroScore", fuzzymatch.JaroScore},
		{"JaroWinklerScore", fuzzymatch.JaroWinklerScore},
		{"SmithWatermanGotohScore", fuzzymatch.SmithWatermanGotohScore},
	}

	input := "abc"
	for _, f := range funcs {
		t.Run(f.name, func(t *testing.T) {
			got := f.fn(input, input)
			if got != 1.0 {
				t.Errorf("%s(%q, %q) = %v, want 1.0 (identity invariant)", f.name, input, input, got)
			}
		})
	}
}

// TestCrossAlgorithm_BothEmptyConvergence asserts that all seven Phase 2 + 3
// algorithms return Score("", "") == 1.0. Both-empty identity is the
// project-wide convention documented in RESEARCH.md §Score Normalisation:
// two equal strings (even if empty) have similarity 1.0.
func TestCrossAlgorithm_BothEmptyConvergence(t *testing.T) {
	type scoreFunc struct {
		name string
		fn   func(a, b string) float64
	}

	funcs := []scoreFunc{
		{"LevenshteinScore", fuzzymatch.LevenshteinScore},
		{"DamerauLevenshteinOSAScore", fuzzymatch.DamerauLevenshteinOSAScore},
		{"DamerauLevenshteinFullScore", fuzzymatch.DamerauLevenshteinFullScore},
		{"HammingScore", fuzzymatch.HammingScore},
		{"JaroScore", fuzzymatch.JaroScore},
		{"JaroWinklerScore", fuzzymatch.JaroWinklerScore},
		{"SmithWatermanGotohScore", fuzzymatch.SmithWatermanGotohScore},
	}

	for _, f := range funcs {
		t.Run(f.name, func(t *testing.T) {
			got := f.fn("", "")
			if got != 1.0 {
				t.Errorf("%s(\"\", \"\") = %v, want 1.0 (both-empty convention)", f.name, got)
			}
		})
	}
}

// TestCrossAlgorithm_SingleSubstitution_DistanceAgreement asserts that the
// four distance-bearing Phase 2 algorithms all return distance == 1 for a
// single-character substitution between equal-length strings ("a" vs "b").
//
// Jaro and Jaro-Winkler are excluded: they have no Distance variant and
// their Score values for "a"/"b" are not directly comparable to edit
// distances.
func TestCrossAlgorithm_SingleSubstitution_DistanceAgreement(t *testing.T) {
	type distFunc struct {
		name string
		fn   func(a, b string) int
	}

	funcs := []distFunc{
		{"LevenshteinDistance", fuzzymatch.LevenshteinDistance},
		{"DamerauLevenshteinOSADistance", fuzzymatch.DamerauLevenshteinOSADistance},
		{"DamerauLevenshteinFullDistance", fuzzymatch.DamerauLevenshteinFullDistance},
		{"HammingDistance", fuzzymatch.HammingDistance},
	}

	for _, f := range funcs {
		t.Run(f.name, func(t *testing.T) {
			got := f.fn("a", "b")
			if got != 1 {
				t.Errorf("%s(\"a\", \"b\") = %d, want 1 (single-character substitution)", f.name, got)
			}
		})
	}
}

// TestCrossAlgorithm_OneEmpty_ScoreAgreement asserts that all seven Phase 2 + 3
// algorithms return Score("", "abc") == 0.0. A non-empty string has zero
// similarity to the empty string — maximally dissimilar.
func TestCrossAlgorithm_OneEmpty_ScoreAgreement(t *testing.T) {
	type scoreFunc struct {
		name string
		fn   func(a, b string) float64
	}

	funcs := []scoreFunc{
		{"LevenshteinScore", fuzzymatch.LevenshteinScore},
		{"DamerauLevenshteinOSAScore", fuzzymatch.DamerauLevenshteinOSAScore},
		{"DamerauLevenshteinFullScore", fuzzymatch.DamerauLevenshteinFullScore},
		{"HammingScore", fuzzymatch.HammingScore},
		{"JaroScore", fuzzymatch.JaroScore},
		{"JaroWinklerScore", fuzzymatch.JaroWinklerScore},
		{"SmithWatermanGotohScore", fuzzymatch.SmithWatermanGotohScore},
	}

	for _, f := range funcs {
		t.Run(f.name, func(t *testing.T) {
			got := f.fn("", "abc")
			if got != 0.0 {
				t.Errorf("%s(\"\", \"abc\") = %v, want 0.0 (one-empty convention)", f.name, got)
			}
		})
	}
}

// TestCrossAlgorithm_SWG_Levenshtein_SubstringDivergence asserts the
// load-bearing local-vs-global-alignment claim: on a substring-
// containment input, Smith-Waterman-Gotoh (LOCAL alignment) scores
// STRICTLY HIGHER than Levenshtein (GLOBAL edit distance), because SWG
// finds the substring while Levenshtein counts every uncovered position
// as an edit.
//
// "http_request" is fully contained in "http_request_header_fields":
//   - SmithWatermanGotohScore = 1.0   (full local match found; clamp
//     returns 1 because raw == min(len))
//   - LevenshteinScore        ≈ 0.46  (≈ 14 edits over max-length 26)
//
// This test will fail if either algorithm regresses on this contract.
func TestCrossAlgorithm_SWG_Levenshtein_SubstringDivergence(t *testing.T) {
	a, b := "http_request", "http_request_header_fields"
	gotSWG := fuzzymatch.SmithWatermanGotohScore(a, b)
	gotLev := fuzzymatch.LevenshteinScore(a, b)
	if !(gotSWG > gotLev) {
		t.Errorf("SWG (%v) must score STRICTLY higher than Levenshtein (%v) on substring-containment pair %q/%q (local-vs-global divergence)",
			gotSWG, gotLev, a, b)
	}
}

// TestCrossAlgorithm_Strcmp95_AtLeastJaroWinkler asserts the algorithm-
// hierarchy invariant from CONTEXT.md §2: Strcmp95 is Jaro-Winkler PLUS
// similar-character credit, prefix boost, and long-string adjustment. Each
// adjustment can only ADD to the underlying Jaro-Winkler score — it must
// never subtract. Therefore Strcmp95(a, b) >= JaroWinkler(a, b) on every
// input pair.
//
// This cross-algorithm test is the catalog-wide counterpart to the
// in-package property test TestProp_Strcmp95Score_AtLeastJaroWinkler (which
// uses testing/quick over arbitrary inputs). The hand-pinned cases below
// fire the adjustments deterministically:
//   - MARTHA/MARHTA  — Winkler's 1990 canonical pair; prefix boost fires
//   - DWAYNE/DUANE   — similar-character credit fires (W↔U)
//   - DIXON/DICKSONX — similar-character credit fires (C↔K) AND
//     long-string adjustment fires (len > 4 after match)
//
// RESEARCH.md Pitfall 1 warning sign #3 closure.
func TestCrossAlgorithm_Strcmp95_AtLeastJaroWinkler(t *testing.T) {
	pairs := [][2]string{
		{"MARTHA", "MARHTA"},
		{"DWAYNE", "DUANE"},
		{"DIXON", "DICKSONX"},
	}
	for _, p := range pairs {
		a, b := p[0], p[1]
		gotStrcmp := fuzzymatch.Strcmp95Score(a, b)
		gotJW := fuzzymatch.JaroWinklerScore(a, b)
		if !(gotStrcmp >= gotJW) {
			t.Errorf("Strcmp95Score(%q, %q) = %v < JaroWinklerScore(%q, %q) = %v (Strcmp95 adjustments must only ADD)",
				a, b, gotStrcmp, a, b, gotJW)
		}
	}
}

// TestCrossAlgorithm_LCSStr_AtLeastLevenshtein_SubstringContainment asserts
// the LCSStr-vs-Levenshtein divergence on a substring-containment input:
// LCSStr finds the contained substring and credits its full length; Levenshtein
// must pay the deletion cost of every uncovered character.
//
//   - LCSStrScore("http_request", "http_request_header_fields") ≈ 0.6316
//     (2 × 12 contained chars / (12 + 26))
//   - LevenshteinScore(...) ≈ 0.46  (14 edits over max-length 26)
//
// LCSStr therefore scores AT LEAST as high as Levenshtein on inputs where one
// is a contiguous substring of the other. The relation may not hold on
// non-containment inputs (Levenshtein can credit interleaved single-character
// matches that LCSStr ignores), so the assertion is hand-pinned to a
// containment pair rather than checked via testing/quick.
func TestCrossAlgorithm_LCSStr_AtLeastLevenshtein_SubstringContainment(t *testing.T) {
	a, b := "http_request", "http_request_header_fields"
	gotLCS := fuzzymatch.LCSStrScore(a, b)
	gotLev := fuzzymatch.LevenshteinScore(a, b)
	if !(gotLCS >= gotLev) {
		t.Errorf("LCSStrScore(%q, %q) = %v < LevenshteinScore(%q, %q) = %v (LCSStr must credit substring containment at least as much as Levenshtein)",
			a, b, gotLCS, a, b, gotLev)
	}
}

// TestCrossAlgorithm_RatcliffObershelp_PinnedDrDobbs pins the canonical
// Dr. Dobb's 1988 reference vector WIKIMEDIA/WIKIMANIA at the
// difflib(autojunk=False).ratio() value 0.7777777777777778 (= 14/18). This
// is the Phase 3 WR-03 closure: a numerical-regression pin OUTSIDE the
// cross-validation corpus, so a regression that silently mutates the corpus
// alone cannot mask drift.
func TestCrossAlgorithm_RatcliffObershelp_PinnedDrDobbs(t *testing.T) {
	const want = 0.7777777777777778 // difflib.SequenceMatcher(autojunk=False, a="WIKIMEDIA", b="WIKIMANIA").ratio()
	const tol = 1e-9
	got := fuzzymatch.RatcliffObershelpScore("WIKIMEDIA", "WIKIMANIA")
	if math.Abs(got-want) > tol {
		t.Errorf("RatcliffObershelpScore(\"WIKIMEDIA\", \"WIKIMANIA\") = %.17f; want %.17f within %g",
			got, want, tol)
	}
}

// TestCrossAlgorithm_RatcliffObershelp_PinnedAgainstDifflib is the
// load-bearing divergence-pin: it picks a pair where RatcliffObershelp
// VISIBLY differs from both Levenshtein AND Jaro-Winkler, then asserts
// byte-for-byte agreement with difflib(autojunk=False).ratio(). This proves
// the RO surface is the difflib-equivalent and not a near-clone of another
// catalogue algorithm.
//
// For "GESTALT"/"GESTALT_PATTERN_MATCHING":
//   - RatcliffObershelpScore ≈ 0.45161290322580644 (= 14/31; matches difflib)
//   - LevenshteinScore       ≈ 0.29166666666666663 (≈ 17 edits / 24)
//   - JaroWinklerScore       ≈ 0.85833333333333339 (strong prefix-boost)
//
// The divergences are wide (RO differs from Levenshtein by ~0.16 and from
// Jaro-Winkler by ~0.40), well above any sub-ULP tolerance.
func TestCrossAlgorithm_RatcliffObershelp_PinnedAgainstDifflib(t *testing.T) {
	const (
		// Pinned difflib value: difflib.SequenceMatcher(autojunk=False,
		//   a="GESTALT", b="GESTALT_PATTERN_MATCHING").ratio() = 14/31.
		wantDifflib = 0.45161290322580644
		tolDifflib  = 1e-9
		tolDiverge  = 1e-6
	)
	a, b := "GESTALT", "GESTALT_PATTERN_MATCHING"
	gotRO := fuzzymatch.RatcliffObershelpScore(a, b)
	gotLev := fuzzymatch.LevenshteinScore(a, b)
	gotJW := fuzzymatch.JaroWinklerScore(a, b)

	// 1. Divergence: RO must differ from BOTH Levenshtein and Jaro-Winkler.
	if math.Abs(gotRO-gotLev) <= tolDiverge {
		t.Errorf("RatcliffObershelpScore(%q, %q) = %v converges with LevenshteinScore = %v (expected visible divergence)",
			a, b, gotRO, gotLev)
	}
	if math.Abs(gotRO-gotJW) <= tolDiverge {
		t.Errorf("RatcliffObershelpScore(%q, %q) = %v converges with JaroWinklerScore = %v (expected visible divergence)",
			a, b, gotRO, gotJW)
	}

	// 2. difflib equivalence: RO must match the pinned difflib(autojunk=False) value.
	if math.Abs(gotRO-wantDifflib) > tolDifflib {
		t.Errorf("RatcliffObershelpScore(%q, %q) = %.17f; want %.17f within %g (difflib(autojunk=False).ratio() byte-for-byte)",
			a, b, gotRO, wantDifflib, tolDifflib)
	}
}

// TestCrossAlgorithm_RatcliffObershelp_AsymmetryPin is the OQ-1 resolution
// regression guard, locked 2026-05-14 (CONTEXT.md §4). Ratcliff-Obershelp
// is INTENTIONALLY asymmetric in argument order — it mirrors Python
// difflib.SequenceMatcher.ratio(), which is noncommutative because the
// recursive longest-common-substring decomposition is biased toward the
// first input.
//
// This test is the INVERSE-FORM regression guard: it asserts INEQUALITY
// rather than ordering. If a future refactor accidentally introduces a
// symmetric workaround (e.g. canonicalising the input order), fwd == rev
// will trip this test and surface the regression.
//
// "tide"/"diet" is the canonical asymmetric pair documented in
// ratcliff_obershelp.go's godoc:
//   - RatcliffObershelpScore("tide", "diet") = 0.25
//   - RatcliffObershelpScore("diet", "tide") = 0.5
func TestCrossAlgorithm_RatcliffObershelp_AsymmetryPin(t *testing.T) {
	fwd := fuzzymatch.RatcliffObershelpScore("tide", "diet")
	rev := fuzzymatch.RatcliffObershelpScore("diet", "tide")
	if fwd == rev {
		t.Errorf("RatcliffObershelpScore is INTENTIONALLY asymmetric for tide/diet — got fwd=%g==rev=%g (regression to symmetric behaviour)",
			fwd, rev)
	}
}
