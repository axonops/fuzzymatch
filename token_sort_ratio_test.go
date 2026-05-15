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

// token_sort_ratio_test.go pins the public-API contract of
// token_sort_ratio.go: identity / both-empty / one-empty conventions,
// token-reorder invariance (the load-bearing property), subset, disjoint,
// Unicode (multi-byte UTF-8), and dispatch registration.
//
// Each non-trivial reference vector's derivation is reproduced in the
// test name so reviewers can re-derive the score from the formula
// 2·LCS / (|joinedA|+|joinedB|) in seconds.
//
// Stdlib `testing` only — no testify in root tests, per
// .claude/skills/go-coding-standards.

package fuzzymatch_test

import (
	"math"
	"testing"

	"github.com/axonops/fuzzymatch"
)

// tokenSortEpsilon is the float-comparison tolerance for irrational
// expected values. Project-locked at 1e-9 (matches Phase 5 q-gram
// epsilon). Exact-rational values (1.0, 0.0, 0.2) use direct equality.
const tokenSortEpsilon = 1e-9

// TestTokenSortRatioScore exercises the load-bearing reference vectors
// for the public surface. Each row's derivation is in the test name
// and inline comment so a reviewer can verify the answer from the
// formula 2·LCS / (|joinedA|+|joinedB|) in seconds.
func TestTokenSortRatioScore(t *testing.T) {
	tests := []struct {
		name       string
		a, b       string
		want       float64
		exact      bool
		derivation string
	}{
		{
			name: "both_empty_identity", a: "", b: "",
			want: 1.0, exact: true,
			derivation: "a == b identity short-circuit covers both-empty → 1.0",
		},
		{
			name: "identical_non_empty", a: "hello", b: "hello",
			want: 1.0, exact: true,
			derivation: "a == b identity short-circuit → 1.0",
		},
		{
			name: "one_empty_a", a: "", b: "hello",
			want: 0.0, exact: true,
			derivation: "one-empty convention → 0.0",
		},
		{
			name: "one_empty_b", a: "hello", b: "",
			want: 0.0, exact: true,
			derivation: "one-empty convention → 0.0",
		},
		{
			name: "token_reorder_two_tokens", a: "alpha beta", b: "beta alpha",
			want: 1.0, exact: true,
			derivation: "sorted(A)=sorted(B)=[alpha,beta]; joined identical; indelRatio=1.0",
		},
		{
			name: "token_reorder_canonical_fuzzy", a: "fuzzy wuzzy was a bear", b: "wuzzy fuzzy was a bear",
			want: 1.0, exact: true,
			derivation: "sorted tokens identical on both sides → joined identical → 1.0",
		},
		{
			name: "subset_alpha_in_alpha_beta", a: "alpha", b: "alpha beta",
			want: 2.0 * 5.0 / 15.0, exact: false,
			derivation: "joinedA=\"alpha\" (5B); joinedB=\"alpha beta\" (10B); lcs=5 (\"alpha\"); 2·5/15 ≈ 0.6667",
		},
		{
			name: "subset_alpha_beta_in_alpha_beta_gamma", a: "alpha beta", b: "alpha beta gamma",
			want: 2.0 * 10.0 / 26.0, exact: false,
			derivation: "joinedA=\"alpha beta\" (10B); joinedB=\"alpha beta gamma\" (16B); lcs=10; 2·10/26 ≈ 0.7692",
		},
		{
			name: "disjoint_abc_xyz", a: "abc", b: "xyz",
			want: 0.0, exact: true,
			derivation: "joinedA=\"abc\"; joinedB=\"xyz\"; lcs=0; 2·0/6 = 0.0",
		},
		{
			name: "low_overlap_hello_world", a: "hello", b: "world",
			want: 0.2, exact: true,
			derivation: "joinedA=\"hello\"; joinedB=\"world\"; lcs=1 (\"l\"); 2·1/10 = 0.2",
		},
		{
			// Unicode reference vector. Sorted-joined:
			//   A = "café société" — c(1)a(1)f(1)é(2) + " " + s(1)o(1)c(1)i(1)é(2)t(1)é(2) = 15 bytes
			//   B = "cafe societe"  — cafe(4) + " " + societe(7) = 12 bytes
			// Byte-level LCS = 9 (cross-validated against the Wagner-Fischer
			// kernel; the implementation produces 0.66666... = 2·9/27).
			// This vector is locked here for cross-platform float-determinism
			// in addition to its RapidFuzz cross-validation counterpart.
			name: "unicode_cafe_societe", a: "café société", b: "cafe societe",
			want: 2.0 * 9.0 / 27.0, exact: false,
			derivation: "joinedA UTF-8 len=15B; joinedB len=12B; byte-LCS=9; 2·9/27 ≈ 0.6667",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Logf("derivation: %s", tt.derivation)
			got := fuzzymatch.TokenSortRatioScore(tt.a, tt.b)
			if tt.exact {
				if got != tt.want {
					t.Errorf("TokenSortRatioScore(%q, %q) = %.17g; want %.17g exactly",
						tt.a, tt.b, got, tt.want)
				}
			} else {
				if math.Abs(got-tt.want) > tokenSortEpsilon {
					t.Errorf("TokenSortRatioScore(%q, %q) = %.17g; want %.17g (Δ=%g, ε=%g)",
						tt.a, tt.b, got, tt.want, math.Abs(got-tt.want), tokenSortEpsilon)
				}
			}
		})
	}
}

// TestTokenSortRatioScore_BothPureSeparators pins the both-Tokenised-empty
// branch: inputs that differ as raw strings but tokenise to empty slices
// on both sides should return 1.0 (vacuous match — the same convention
// as a == b but reached via the post-Tokenise len-check rather than the
// identity short-circuit).
//
// This is the only path that exercises the both-Tokenised-empty branch
// in TokenSortRatioScore (line `if len(tokensA) == 0 && len(tokensB) == 0`);
// without this test the branch would be uncovered.
func TestTokenSortRatioScore_BothPureSeparators(t *testing.T) {
	tests := []struct{ a, b string }{
		{" ", "  "},        // single space vs two spaces
		{"___", "..."},     // different separators
		{"\t\n", "  "},     // whitespace separators
		{" _.- ", "_-.: "}, // mixed separator runs
	}
	for _, tt := range tests {
		t.Run(tt.a+"_vs_"+tt.b, func(t *testing.T) {
			got := fuzzymatch.TokenSortRatioScore(tt.a, tt.b)
			if got != 1.0 {
				t.Errorf("TokenSortRatioScore(%q, %q) = %g; want 1.0 (both Tokenise to empty)",
					tt.a, tt.b, got)
			}
		})
	}
}

// TestTokenSortRatioScore_OneSidePureSeparator pins the
// one-Tokenised-empty branch: one side tokenises to a non-empty slice
// and the other to empty. Score must be 0.0.
func TestTokenSortRatioScore_OneSidePureSeparator(t *testing.T) {
	tests := []struct{ a, b string }{
		{"hello", " "},
		{" ", "hello"},
		{"alpha beta", "___"},
	}
	for _, tt := range tests {
		t.Run(tt.a+"_vs_"+tt.b, func(t *testing.T) {
			got := fuzzymatch.TokenSortRatioScore(tt.a, tt.b)
			if got != 0.0 {
				t.Errorf("TokenSortRatioScore(%q, %q) = %g; want 0.0 (one side Tokenises to empty)",
					tt.a, tt.b, got)
			}
		})
	}
}

// TestTokenSortRatioScore_Symmetric pins the algorithm's symmetry
// across argument order in a tabular form (the quick.Check property
// test in props_test.go provides random-input coverage).
func TestTokenSortRatioScore_Symmetric(t *testing.T) {
	tests := []struct{ a, b string }{
		{"alpha beta", "beta alpha"},
		{"alpha", "alpha beta"},
		{"alpha beta gamma", "gamma alpha"},
		{"hello world", "world hello"},
		{"abc", "xyz"},
		{"café société", "cafe societe"},
	}
	for _, tt := range tests {
		t.Run(tt.a+"_vs_"+tt.b, func(t *testing.T) {
			fwd := fuzzymatch.TokenSortRatioScore(tt.a, tt.b)
			rev := fuzzymatch.TokenSortRatioScore(tt.b, tt.a)
			if fwd != rev {
				t.Errorf("TokenSortRatioScore not symmetric: T(%q,%q)=%g, T(%q,%q)=%g",
					tt.a, tt.b, fwd, tt.b, tt.a, rev)
			}
		})
	}
}

// TestTokenSortRatioScore_DispatchRegistration pins that
// dispatch[AlgoTokenSortRatio] is populated after package load AND
// that invoking the dispatched function returns the same score as a
// direct call. Exercises the dispatch-table side of plan 06-01 task 2.
func TestTokenSortRatioScore_DispatchRegistration(t *testing.T) {
	if fuzzymatch.DispatchEntryNilForTest(int(fuzzymatch.AlgoTokenSortRatio)) {
		t.Fatalf("dispatch[AlgoTokenSortRatio] (%d) is nil — dispatch_token_sort_ratio.go must register TokenSortRatioScore at package load time",
			int(fuzzymatch.AlgoTokenSortRatio))
	}
	got := fuzzymatch.DispatchInvokeForTest(int(fuzzymatch.AlgoTokenSortRatio), "alpha beta", "beta alpha")
	want := fuzzymatch.TokenSortRatioScore("alpha beta", "beta alpha")
	if got != want {
		t.Errorf("dispatch[AlgoTokenSortRatio](\"alpha beta\",\"beta alpha\") = %.17g; want %.17g",
			got, want)
	}
	if got != 1.0 {
		t.Errorf("dispatch[AlgoTokenSortRatio](\"alpha beta\",\"beta alpha\") = %.17g; want 1.0 (token-reorder invariant)",
			got)
	}
}
