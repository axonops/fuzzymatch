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

// cross_algorithm_consistency_test.go pins relationships across the six
// Phase 2 character-based algorithms:
//
//   - DIVERGENCE: DL-OSA and DL-Full produce DIFFERENT distances for
//     the canonical discriminating vector "ca"/"abc" (3 vs 2). This
//     is a load-bearing claim of the phase (ROADMAP success criterion #2);
//     the file pins it cross-algorithm rather than only inside each
//     algorithm's own _test.go file.
//
//   - CONVERGENCE: every algorithm's score on identical inputs is
//     exactly 1.0 (identity).
//
//   - SINGLE-SUBSTITUTION AGREEMENT: Levenshtein, DL-OSA, DL-Full,
//     and Hamming all return distance == 1 for a single-character
//     substitution between equal-length strings (e.g. "a"/"b").
//
// Stdlib `testing` only — no testify in root tests, per
// .claude/skills/go-coding-standards.

package fuzzymatch_test

import (
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

// TestCrossAlgorithm_IdentityConvergence asserts that all six Phase 2
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

// TestCrossAlgorithm_BothEmptyConvergence asserts that all six Phase 2
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

// TestCrossAlgorithm_OneEmpty_ScoreAgreement asserts that all six Phase 2
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
