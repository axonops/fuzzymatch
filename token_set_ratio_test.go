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

// token_set_ratio_test.go pins the public-API contract of
// token_set_ratio.go: identity / one-empty / BOTH-EMPTY DEVIATION /
// subset short-circuit / three-way-max-diff-dominant / disjoint /
// Unicode / dispatch registration / symmetry.
//
// The both-empty case is the load-bearing DEVIATION — TokenSetRatio
// returns 0.0 (NOT 1.0) when either Tokenise output is empty, per the
// LOCKED RapidFuzz issue #110 bug-for-bug compatibility. This is the
// only algorithm in the catalogue with this deviation.
//
// Each non-trivial reference vector's derivation is reproduced in the
// test name and inline comment so reviewers can re-derive the score
// from the three-way max formula in seconds.
//
// Stdlib `testing` only — no testify in root tests, per
// .claude/skills/go-coding-standards.

package fuzzymatch_test

import (
	"math"
	"testing"

	"github.com/axonops/fuzzymatch"
)

// tokenSetEpsilon is the float-comparison tolerance for irrational
// expected values. Project-locked at 1e-9 (matches Phase 5 q-gram
// and Phase 6 TokenSortRatio epsilon).
const tokenSetEpsilon = 1e-9

// TestTokenSetRatioScore exercises the load-bearing reference vectors
// for the public surface. Each row's derivation is in the test name
// and inline comment so a reviewer can verify the answer from the
// three-way max formula (r1=indel(sect,combined1to2), r2=indel(sect,
// combined2to1), r3=indel(combined1to2,combined2to1)) in seconds.
func TestTokenSetRatioScore(t *testing.T) {
	tests := []struct {
		name       string
		a, b       string
		want       float64
		exact      bool
		derivation string
	}{
		{
			// a == b identity short-circuit fires BEFORE Tokenise.
			name: "identical_non_empty", a: "alpha beta", b: "alpha beta",
			want: 1.0, exact: true,
			derivation: "a == b identity short-circuit → 1.0",
		},
		{
			// LOCKED DEVIATION (RapidFuzz issue #110 / fuzzywuzzy
			// parity): both-Tokenised-empty returns 0.0, NOT 1.0.
			// The a == b short-circuit ALSO covers ("", "") with 1.0
			// — but the deviation gate fires when the strings DIFFER
			// but both tokenise to empty (e.g. (" ", "  ")). For the
			// pure ("", "") case, the identity short-circuit
			// produces 1.0 *unless* we cross-validate against
			// RapidFuzz which produces 0.0 for ("", ""). We anchor
			// the documented test as ("", "  ") so the
			// post-Tokenise empty-set gate fires (the strings
			// differ, so the identity short-circuit does NOT fire).
			name: "both_pure_separators_deviation", a: "  ", b: " ",
			want: 0.0, exact: true,
			derivation: "DEVIATION: both Tokenise to [] → 0.0 (RapidFuzz issue #110); identity short-circuit does NOT fire (\"  \" != \" \")",
		},
		{
			// LOCKED DEVIATION (RapidFuzz issue #110 / fuzzywuzzy
			// parity): ("", "") returns 0.0 — NOT 1.0. The
			// empty-input gate fires BEFORE the identity
			// short-circuit so the both-empty-→-1.0 catalogue
			// convention does not apply to TokenSetRatio. This is
			// the load-bearing deviation pinned in both the
			// cross-validation corpus and the staging-golden file.
			name: "both_empty_strings_deviation", a: "", b: "",
			want: 0.0, exact: true,
			derivation: "DEVIATION: empty-input gate fires before identity short-circuit — TokenSetRatioScore(\"\",\"\") returns 0.0 per RapidFuzz issue #110 (NOT 1.0)",
		},
		{
			name: "one_empty_a", a: "", b: "alpha beta",
			want: 0.0, exact: true,
			derivation: "one-Tokenised-empty → 0.0 (matches catalogue convention; not a deviation)",
		},
		{
			name: "one_empty_b", a: "alpha beta", b: "",
			want: 0.0, exact: true,
			derivation: "one-Tokenised-empty → 0.0",
		},
		{
			// Subset short-circuit (RESEARCH.md Pattern 5 critical
			// landmine 2): intersection={alpha,beta}; diff_ab={};
			// diff_ba={gamma}; non-empty intersection AND diff_ab
			// empty → return 1.0 directly.
			name: "subset_a_in_b", a: "alpha beta", b: "alpha beta gamma",
			want: 1.0, exact: true,
			derivation: "intersection={alpha,beta}; diff_ab={}; diff_ba={gamma}; subset short-circuit → 1.0",
		},
		{
			name: "subset_b_in_a", a: "alpha beta gamma", b: "alpha beta",
			want: 1.0, exact: true,
			derivation: "intersection={alpha,beta}; diff_ab={gamma}; diff_ba={}; subset short-circuit → 1.0",
		},
		{
			// Token-reorder + dedup: A={alpha,beta}, B={alpha,beta}
			// after dedup ("alpha beta alpha" → set {alpha, beta}).
			// Both diffs empty AND intersection non-empty → 1.0 via
			// subset short-circuit.
			name: "token_reorder_with_dup", a: "alpha beta", b: "beta alpha alpha beta",
			want: 1.0, exact: true,
			derivation: "A_set={alpha,beta}; B_set={alpha,beta}; both diffs empty; subset short-circuit → 1.0",
		},
		{
			// Three-way max with combined-vs-combined (r3) winning:
			// intersection={world}; diff_ab={hello}; diff_ba={peace}.
			// sortedSect="world" (5); combined1to2="world hello"
			// (11); combined2to1="world peace" (11).
			// r1 = ratio("world","world hello") = 2*5/(5+11) = 10/16 = 0.625
			// r2 = ratio("world","world peace") = 10/16 = 0.625
			// r3 = ratio("world hello","world peace") =
			//      LCS("world hello","world peace")=7 ("world e" with the e
			//      from "hello"'s e at pos 1 not pos 7 — but "world " then 'e'
			//      from "hello" at pos 7 matches 'e' from "peace" at pos 7 or
			//      pos 10. LCS = "world " + "e" = 7 chars).
			//      = 2*7/22 = 14/22 = 7/11 ≈ 0.6363636363636364
			// Max = r3 = 7/11.
			//
			// This is the LOCKED diff-dominant test case where the
			// third branch (combined-vs-combined) strictly beats the
			// first two (intersection-vs-combined).
			name: "three_way_max_combined_wins", a: "hello world", b: "world peace",
			want: 7.0 / 11.0, exact: false,
			derivation: "intersection={world}; r1=r2=10/16=0.625; r3=ratio(\"world hello\",\"world peace\")=14/22=7/11≈0.6364; r3 wins",
		},
		{
			// All tokens shared as singletons → both diffs empty;
			// the a == b identity short-circuit fires BEFORE
			// Tokenise (since a == b as strings).
			name: "token_reorder_two", a: "alpha beta", b: "beta alpha",
			want: 1.0, exact: true,
			derivation: "A_set={alpha,beta}; B_set={alpha,beta}; both diffs empty AND intersection non-empty → subset short-circuit → 1.0 (a != b as strings; identity short-circuit does NOT fire)",
		},
		{
			// Disjoint sets: intersection={}, diff_ab={abc,def},
			// diff_ba={qrs,xyz}.
			// sortedSect="" (0); combined1to2="abc def" (7);
			// combined2to1="qrs xyz" (7).
			// r1 = ratio("","abc def") = 0.0 (one-empty)
			// r2 = ratio("","qrs xyz") = 0.0
			// r3 = ratio("abc def","qrs xyz"). LCS = " " (space) = 1.
			//      = 2*1/14 = 1/7 ≈ 0.1428571428571429.
			// Max = r3 = 1/7.
			name: "disjoint", a: "abc def", b: "xyz qrs",
			want: 1.0 / 7.0, exact: false,
			derivation: "intersection={}; combined1to2=\"abc def\" (7B); combined2to1=\"qrs xyz\" (7B); LCS=1 (space); r3=2/14=1/7",
		},
		{
			// Single-token both sides, disjoint: intersection={},
			// diff_ab={hello}, diff_ba={world}.
			// sortedSect=""; combined1to2="hello"; combined2to1="world".
			// r1 = r2 = 0.0; r3 = ratio("hello","world"). LCS=1 ('l').
			//      = 2/10 = 0.2.
			name: "low_overlap_singletons", a: "hello", b: "world",
			want: 0.2, exact: true,
			derivation: "intersection={}; r3=ratio(\"hello\",\"world\")=2*1/10=0.2",
		},
		{
			// Unicode reference vector. Two tokens per side, sorted
			// tokens identical after Tokenise normalisation. Both
			// inputs Tokenise to two tokens each — set-equal → both
			// diffs empty → subset short-circuit fires AFTER the
			// a != b identity check fails (the raw strings differ
			// by token order). Score = 1.0.
			name: "unicode_token_reorder", a: "café société", b: "société café",
			want: 1.0, exact: true,
			derivation: "Tokenise(café société)={café,société}; Tokenise(société café)={société,café}; sets equal; subset short-circuit → 1.0",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Logf("derivation: %s", tt.derivation)
			got := fuzzymatch.TokenSetRatioScore(tt.a, tt.b)
			if tt.exact {
				if got != tt.want {
					t.Errorf("TokenSetRatioScore(%q, %q) = %.17g; want %.17g exactly",
						tt.a, tt.b, got, tt.want)
				}
			} else {
				if math.Abs(got-tt.want) > tokenSetEpsilon {
					t.Errorf("TokenSetRatioScore(%q, %q) = %.17g; want %.17g (Δ=%g, ε=%g)",
						tt.a, tt.b, got, tt.want, math.Abs(got-tt.want), tokenSetEpsilon)
				}
			}
		})
	}
}

// TestTokenSetRatioScore_EmptyDeviationDocumented pins the LOCKED
// RapidFuzz issue #110 deviation in isolation: when Tokenise produces
// an empty token slice on EITHER side AND the identity short-circuit
// does NOT fire (the raw strings differ), the function returns 0.0.
//
// This is the only path that distinguishes TokenSetRatio from the rest
// of the catalogue: every OTHER tokenised algorithm returns 1.0 in the
// both-empty case. The test name carries the citation so a future
// reviewer doesn't accidentally "fix" the deviation.
func TestTokenSetRatioScore_EmptyDeviationDocumented(t *testing.T) {
	tests := []struct {
		name string
		a, b string
		want float64
	}{
		// Identity short-circuit fires for ("", "") — returns 1.0.
		// The deviation gate fires only when the raw strings differ
		// but both tokenise to empty. Pure-separator inputs are the
		// canonical exemplars.
		{"both_pure_separator_runs", " ", "  ", 0.0},
		{"both_underscores", "___", "__", 0.0},
		{"both_mixed_separators", "_-.", " . ", 0.0},
		{"one_pure_separator_other_empty", "", " ", 0.0},
		{"one_empty_other_pure_separator", " ", "", 0.0},
		{"one_pure_separator_other_token", "alpha", " ", 0.0},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := fuzzymatch.TokenSetRatioScore(tt.a, tt.b)
			if got != tt.want {
				t.Errorf("TokenSetRatioScore(%q, %q) = %g; want %g (LOCKED DEVIATION — RapidFuzz issue #110; do NOT \"fix\" to 1.0)",
					tt.a, tt.b, got, tt.want)
			}
		})
	}
}

// TestTokenSetRatioScore_SubsetShortCircuit pins the LOCKED subset
// short-circuit (RESEARCH.md Pattern 5 critical landmine 2): when the
// intersection is non-empty AND one of the diffs is empty, the
// algorithm returns 1.0 directly.
func TestTokenSetRatioScore_SubsetShortCircuit(t *testing.T) {
	tests := []struct {
		name string
		a, b string
	}{
		{"a_subset_b_two_three", "alpha beta", "alpha beta gamma"},
		{"b_subset_a_three_two", "alpha beta gamma", "alpha beta"},
		{"a_subset_b_one_three", "alpha", "alpha beta gamma"},
		{"b_subset_a_three_one", "alpha beta gamma", "alpha"},
		{"dup_a_subset_b", "alpha alpha", "alpha beta"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := fuzzymatch.TokenSetRatioScore(tt.a, tt.b)
			if got != 1.0 {
				t.Errorf("TokenSetRatioScore(%q, %q) = %g; want 1.0 (subset short-circuit)",
					tt.a, tt.b, got)
			}
		})
	}
}

// TestTokenSetRatioScore_Symmetric pins the algorithm's symmetry
// across argument order in a tabular form (the quick.Check property
// test in props_test.go provides random-input coverage).
func TestTokenSetRatioScore_Symmetric(t *testing.T) {
	tests := []struct{ a, b string }{
		{"alpha beta", "beta alpha gamma"},
		{"alpha", "alpha beta"},
		{"hello world", "world peace"},
		{"abc def", "xyz qrs"},
		{"café société", "société café"},
		{"the cat sat on the mat", "the cat sat on the rug"},
	}
	for _, tt := range tests {
		t.Run(tt.a+"_vs_"+tt.b, func(t *testing.T) {
			fwd := fuzzymatch.TokenSetRatioScore(tt.a, tt.b)
			rev := fuzzymatch.TokenSetRatioScore(tt.b, tt.a)
			if fwd != rev {
				t.Errorf("TokenSetRatioScore not symmetric: T(%q,%q)=%g, T(%q,%q)=%g",
					tt.a, tt.b, fwd, tt.b, tt.a, rev)
			}
		})
	}
}

// TestTokenSetRatioScore_DispatchRegistration pins that
// dispatch[AlgoTokenSetRatio] is populated after package load AND
// that invoking the dispatched function returns the same score as a
// direct call. Exercises the dispatch-table side of plan 06-02.
func TestTokenSetRatioScore_DispatchRegistration(t *testing.T) {
	if fuzzymatch.DispatchEntryNilForTest(int(fuzzymatch.AlgoTokenSetRatio)) {
		t.Fatalf("dispatch[AlgoTokenSetRatio] (%d) is nil — dispatch_token_set_ratio.go must register TokenSetRatioScore at package load time",
			int(fuzzymatch.AlgoTokenSetRatio))
	}
	got := fuzzymatch.DispatchInvokeForTest(int(fuzzymatch.AlgoTokenSetRatio), "alpha beta", "alpha beta")
	want := fuzzymatch.TokenSetRatioScore("alpha beta", "alpha beta")
	if got != want {
		t.Errorf("dispatch[AlgoTokenSetRatio](\"alpha beta\",\"alpha beta\") = %.17g; want %.17g",
			got, want)
	}
	if got != 1.0 {
		t.Errorf("dispatch[AlgoTokenSetRatio](\"alpha beta\",\"alpha beta\") = %.17g; want 1.0 (identity)",
			got)
	}
}
