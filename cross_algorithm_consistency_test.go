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

// TestCrossAlgorithm_Tversky_JaccardEquivalence is the Phase 5 cross-algorithm
// pin for the Tversky → Jaccard degenerate case. At α = β = 1.0 the Tversky
// formula reduces to the multiset Jaccard coefficient — see Tversky 1977 §2 and
// the RV-T3 reference vector in Phase 5 RESEARCH.md §1.4 / §2.4.
//
// Algebra (multiset, |A∩B| = c, |A\B| = a, |B\A| = b):
//
//	tversky = c / (c + α·a + β·b)
//	         = c / (c + 1·a + 1·b)        // α = β = 1.0
//	         = c / (c + a + b)
//	         = |A∩B| / |A∪B|              // = Jaccard
//
// The assertion is BIT-EXACT (math.Float64bits equality) — any non-equality
// indicates a divergent reduction order between the Tversky and Jaccard hot
// paths, which would also threaten the Cosine cross-platform determinism gate.
//
// Cross-algorithm defence-in-depth layered atop the unit test in tversky_test.go
// and the property test in props_test.go.
func TestCrossAlgorithm_Tversky_JaccardEquivalence(t *testing.T) {
	pairs := []struct {
		a, b string
		n    int
	}{
		{"abcd", "abce", 2},     // RV-T3 canonical
		{"AGCT", "AGCTAGCT", 2}, // RV-J1 from Ukkonen 1992
		{"hello", "world", 2},   // no-overlap edge case
		{"café", "cafe", 2},     // multi-byte input (byte path)
		{"abcdef", "bcdefg", 3}, // larger overlap, n=3
	}
	for _, p := range pairs {
		gotTversky := fuzzymatch.TverskyScore(p.a, p.b, p.n, 1.0, 1.0)
		gotJaccard := fuzzymatch.QGramJaccardScore(p.a, p.b, p.n)
		if math.Float64bits(gotTversky) != math.Float64bits(gotJaccard) {
			t.Errorf("Tversky/Jaccard equivalence broke for (%q,%q,n=%d):\n  Tversky(α=β=1.0) = %.17g (bits=0x%016x)\n  Jaccard          = %.17g (bits=0x%016x)",
				p.a, p.b, p.n,
				gotTversky, math.Float64bits(gotTversky),
				gotJaccard, math.Float64bits(gotJaccard))
		}
	}
}

// TestCrossAlgorithm_Tversky_DiceEquivalence is the Phase 5 cross-algorithm
// pin for the Tversky → Sørensen-Dice degenerate case. At α = β = 0.5 the
// Tversky formula reduces to the Dice coefficient — see Tversky 1977 §2 and
// the RV-T4 reference vector in Phase 5 RESEARCH.md §1.4 / §2.4.
//
// Algebra (multiset, |A∩B| = c, |A\B| = a, |B\A| = b):
//
//	tversky = c / (c + α·a + β·b)
//	         = c / (c + 0.5·a + 0.5·b)            // α = β = 0.5
//	         = c / (c + 0.5·(a + b))
//	         = 2c / (2c + a + b)
//	         = 2·|A∩B| / (|A| + |B|)              // = Sørensen-Dice
//
// The assertion is BIT-EXACT — defence-in-depth layered atop the unit test
// and the property test for Tversky's Dice-equivalence regime.
func TestCrossAlgorithm_Tversky_DiceEquivalence(t *testing.T) {
	pairs := []struct {
		a, b string
		n    int
	}{
		{"abcd", "abce", 2}, // RV-T4 canonical
		{"AGCT", "AGCTAGCT", 2},
		{"hello", "world", 2},
		{"café", "cafe", 2},
		{"abcdef", "bcdefg", 3},
	}
	for _, p := range pairs {
		gotTversky := fuzzymatch.TverskyScore(p.a, p.b, p.n, 0.5, 0.5)
		gotDice := fuzzymatch.SorensenDiceScore(p.a, p.b, p.n)
		if math.Float64bits(gotTversky) != math.Float64bits(gotDice) {
			t.Errorf("Tversky/Dice equivalence broke for (%q,%q,n=%d):\n  Tversky(α=β=0.5) = %.17g (bits=0x%016x)\n  Dice             = %.17g (bits=0x%016x)",
				p.a, p.b, p.n,
				gotTversky, math.Float64bits(gotTversky),
				gotDice, math.Float64bits(gotDice))
		}
	}
}

// TestCrossAlgorithm_QGramJaccard_AtMostSorensenDice asserts the algebraic
// hierarchy invariant: for every input pair, QGramJaccardScore ≤ SorensenDiceScore.
//
// Derivation: Sørensen-Dice and Jaccard are related by DSC = 2·J / (1 + J).
// For J ∈ [0, 1] we have DSC ≥ J because:
//
//	DSC ≥ J  ⟺  2·J / (1 + J) ≥ J        // (1 + J) > 0 throughout [0, 1]
//	         ⟺  2·J ≥ J·(1 + J)
//	         ⟺  2·J ≥ J + J²
//	         ⟺  J ≥ J²
//	         ⟺  J·(1 − J) ≥ 0             // true for J ∈ [0, 1] ✓
//
// Equality holds iff J ∈ {0, 1}. The hand-pinned table covers RV-J1..J4 and
// RV-D1..D4 plus boundary cases; if a regression silently swaps the two
// algorithms or applies the wrong normalisation, this test fires.
func TestCrossAlgorithm_QGramJaccard_AtMostSorensenDice(t *testing.T) {
	pairs := []struct {
		a, b string
		n    int
	}{
		{"AGCT", "AGCTAGCT", 2}, // RV-J1
		{"abcd", "abxy", 2},     // partial overlap
		{"hello", "world", 2},   // no overlap → J = D = 0 (boundary)
		{"hello", "hello", 2},   // identical → J = D = 1.0 (boundary)
		{"abcdef", "bcdefg", 3}, // RV-D-style overlap, n=3
		{"night", "nacht", 2},   // canonical Dice example
		{"café", "cafe", 2},     // multi-byte (byte path)
		{"abcdef", "abcXef", 3}, // RV-D3 single-substitution, n=3
	}
	for _, p := range pairs {
		gotJ := fuzzymatch.QGramJaccardScore(p.a, p.b, p.n)
		gotD := fuzzymatch.SorensenDiceScore(p.a, p.b, p.n)
		if !(gotJ <= gotD) {
			t.Errorf("Algorithm-hierarchy invariant violated for (%q,%q,n=%d): QGramJaccard = %.17g > SorensenDice = %.17g (DSC = 2·J/(1+J) ≥ J for J ∈ [0,1])",
				p.a, p.b, p.n, gotJ, gotD)
		}
	}
}

// TestCrossAlgorithm_Cosine_GeometricMeanBound is a defence-in-depth sanity
// test for Cosine. It asserts:
//
//   - all hand-pinned pairs return values in [0.0, 1.0]
//   - identity pairs return BIT-EXACT 1.0 (math.Float64bits comparison)
//   - orthogonal pairs return BIT-EXACT 0.0
//
// The LOAD-BEARING gate for Cosine determinism is the cross-platform CI matrix
// running `make verify-determinism` against testdata/golden/algorithms.json (per
// Phase 5 CONTEXT.md §1). This in-package test catches local regressions earlier
// in the dev loop without depending on the multi-platform CI surface.
//
// Inputs span ASCII + Unicode at multiple n values to exercise the sorted-key
// dot-product loop on multiple intersection sizes.
func TestCrossAlgorithm_Cosine_GeometricMeanBound(t *testing.T) {
	type cosCase struct {
		a, b string
		n    int
		// kind discriminates the assertion class.
		kind string // "range" | "identity" | "orthogonal"
	}
	cases := []cosCase{
		{"abc", "abcd", 2, "range"},          // RV-C1 irrational
		{"abcdefgh", "abcdefgi", 3, "range"}, // RV-C2 large intersection
		{"abcde", "abcdf", 4, "range"},       // RV-C3 n=4
		{"café", "cafe", 2, "range"},         // RV-C4 multi-byte
		{"héllo", "hello", 3, "range"},       // RV-C5 multi-byte
		{"hello", "hello", 2, "identity"},    // identity (byte)
		{"abcdef", "abcdef", 3, "identity"},  // identity (longer)
		{"abc", "xyz", 2, "orthogonal"},      // orthogonal
		{"hello", "world", 2, "orthogonal"},  // orthogonal
	}
	for _, c := range cases {
		got := fuzzymatch.CosineScore(c.a, c.b, c.n)
		// Range invariant fires for every case.
		if !(got >= 0.0 && got <= 1.0) {
			t.Errorf("CosineScore(%q,%q,n=%d) = %.17g; want value in [0.0, 1.0]",
				c.a, c.b, c.n, got)
		}
		switch c.kind {
		case "identity":
			if math.Float64bits(got) != math.Float64bits(1.0) {
				t.Errorf("CosineScore(%q,%q,n=%d) identity = %.17g (bits=0x%016x); want 1.0 bit-exact",
					c.a, c.b, c.n, got, math.Float64bits(got))
			}
		case "orthogonal":
			if math.Float64bits(got) != math.Float64bits(0.0) {
				t.Errorf("CosineScore(%q,%q,n=%d) orthogonal = %.17g (bits=0x%016x); want 0.0 bit-exact",
					c.a, c.b, c.n, got, math.Float64bits(got))
			}
		}
	}
}

// TestCrossAlgorithm_Tversky_AsymmetryPin is the Phase 5 INVERSE-FORM
// regression guard for Tversky's intentional asymmetry. Tversky 1977 §2
// defines tversky(a, b) = |A∩B| / (|A∩B| + α·|A\B| + β·|B\A|); at α ≠ β the
// formula is direction-sensitive — swapping the operands yields a different
// score because the α-weighted and β-weighted set differences trade roles.
//
// Hand-pinned canonical pair (matches the RV-T1 / RV-T2 staging-golden entries
// in testdata/golden/_staging/tversky.json):
//
//	TverskyScore("abcd",   "abcdef", 2, 0.8, 0.2) = 0.8823529411764706
//	TverskyScore("abcdef", "abcd",   2, 0.8, 0.2) = 0.6521739130434783
//	|fwd − rev| ≈ 0.2302
//
// This is the FOURTH layer of asymmetry coverage (after the unit test, the
// property test, and the BDD scenario added in Phase 5 plan 05-04). If a
// future refactor accidentally:
//
//   - canonicalises argument order (silent symmetry workaround)
//   - swaps the α and β parameters internally
//   - halves both α and β to preserve their sum
//
// the fwd != rev assertion fails. The 0.1 magnitude floor surfaces near-
// equality regressions early.
func TestCrossAlgorithm_Tversky_AsymmetryPin(t *testing.T) {
	const a, b = "abcd", "abcdef"
	const n = 2
	const alpha, beta = 0.8, 0.2
	fwd := fuzzymatch.TverskyScore(a, b, n, alpha, beta)
	rev := fuzzymatch.TverskyScore(b, a, n, alpha, beta)
	if fwd == rev {
		t.Errorf("TverskyScore is INTENTIONALLY asymmetric for %q/%q with α=%v, β=%v — got fwd=%g == rev=%g (regression to symmetric behaviour or silent α/β swap)",
			a, b, alpha, beta, fwd, rev)
	}
	if math.Abs(fwd-rev) <= 0.1 {
		t.Errorf("TverskyScore asymmetry magnitude collapsed for %q/%q with α=%v, β=%v — |fwd − rev| = %g ≤ 0.1 (expected ≈ 0.2302; near-equality regression)",
			a, b, alpha, beta, math.Abs(fwd-rev))
	}
}
