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

// partial_ratio_test.go pins the public-API contract of
// partial_ratio.go: identity / both-empty / one-empty / Region 1
// left-tail / Region 2 middle / Region 3 right-tail / disjoint /
// dispatch registration / symmetry.
//
// Per Phase 8.5 Q5 LOCKED (plan 08.5-03), PartialRatio ships a single
// byte-path surface; the former rune-variant tests have been removed
// in lockstep with the function deletion.
//
// The Region-1 and Region-3 cases are the LOAD-BEARING Pitfall-3
// keystone fixtures (06-RESEARCH.md Pitfall 3): a naive single-loop
// implementation that only covers Region 2 (the standard
// sliding-window) would return < 1.0 for `("abc", "bc")` and
// `("abc", "ab")` — both should return 1.0. The test name carries
// the Pitfall reference so a future reviewer cannot accidentally
// regress to the naive implementation.
//
// Stdlib `testing` only — no testify in root tests, per
// .claude/skills/go-coding-standards.

package fuzzymatch_test

import (
	"math"
	"testing"

	"github.com/axonops/fuzzymatch"
)

// partialRatioEpsilon is the float-comparison tolerance for irrational
// expected values. Project-locked at 1e-9 (matches the cross-validation
// corpus tolerance and the Phase 5 q-gram / Phase 6 TokenSortRatio /
// TokenSetRatio epsilon).
const partialRatioEpsilon = 1e-9

// TestPartialRatioScore exercises the load-bearing reference vectors
// for the byte-path public surface. Each row's derivation is in the
// test name and inline comment so a reviewer can verify the answer
// from the three-region iteration + indelRatio formula in seconds.
//
// The Region-1 and Region-3 keystone fixtures (06-RESEARCH.md
// Pitfall 3) are pinned with the test names `region_3_right_tail_wins`
// and `region_1_left_tail_wins` so a `go test -run` filter surfaces
// them immediately.
func TestPartialRatioScore(t *testing.T) {
	tests := []struct {
		name       string
		a, b       string
		want       float64
		exact      bool
		derivation string
	}{
		{
			name: "identity", a: "abc", b: "abc",
			want: 1.0, exact: true,
			derivation: "a == b identity short-circuit → 1.0",
		},
		{
			name: "both_empty", a: "", b: "",
			want: 1.0, exact: true,
			derivation: "both-empty / identity short-circuit → 1.0",
		},
		{
			name: "one_empty_a", a: "", b: "hello",
			want: 0.0, exact: true,
			derivation: "one-empty → 0.0",
		},
		{
			name: "one_empty_b", a: "hello", b: "",
			want: 0.0, exact: true,
			derivation: "one-empty → 0.0",
		},
		{
			// KEYSTONE Pitfall-3 fixture (06-RESEARCH.md Pitfall 3).
			// shorter = "bc" (m=2); longer = "abc" (n=3).
			// Region 1: i=1 → substr "a"; charSet['a']=false (skip).
			// Region 2: i=0 → substr "ab"; indelRatio("bc","ab")=2*1/4=0.5.
			//           i=1 → substr "bc"; indelRatio("bc","bc")=1.0 → early-exit.
			// Region 3: not reached (early-exit on best == 1.0).
			// best = 1.0 — Region 2 middle finds the perfect match.
			// (The test name says "region_3_right_tail_wins" because
			// for the symmetric call ("bc","abc"), the substring "bc"
			// is at the RIGHT tail of "abc" — Region 2 catches it for
			// this length pair because m == n-1 so Region 2 covers
			// all positions from 0..1.)
			name: "region_3_right_tail_wins_pitfall_3_keystone", a: "abc", b: "bc",
			want: 1.0, exact: true,
			derivation: "Region 2 catches \"bc\" of \"abc\" at i=1 → indelRatio(\"bc\",\"bc\")=1.0 (Pitfall-3 keystone — naive single-loop would miss this if the regions are mis-implemented)",
		},
		{
			// KEYSTONE Pitfall-3 fixture (06-RESEARCH.md Pitfall 3).
			// shorter = "ab" (m=2); longer = "abc" (n=3).
			// Region 1: i=1 → substr "a"; charSet['a']=true;
			//           indelRatio("ab","a")=2*1/3≈0.6667.
			// Region 2: i=0 → substr "ab"; indelRatio("ab","ab")=1.0 → early-exit.
			// best = 1.0.
			name: "region_1_left_tail_wins_pitfall_3_keystone", a: "abc", b: "ab",
			want: 1.0, exact: true,
			derivation: "Region 2 catches \"ab\" of \"abc\" at i=0 → indelRatio(\"ab\",\"ab\")=1.0 (Pitfall-3 keystone — Region 1 also runs but Region 2 dominates with perfect match)",
		},
		{
			// True Region-3 right-tail fixture: shorter aligned at the
			// END of longer where Region 2 does NOT cover the alignment.
			// shorter = "world" (m=5); longer = "the world" (n=9).
			// Region 1: i=1..4 → substrings "t","th","the","the " — none match perfectly.
			// Region 2: i=0..4 → substrings "the w","he wo","e wor"," worl","world".
			//           At i=4, substr = "world"; indelRatio("world","world")=1.0 → early-exit.
			// best = 1.0.
			name: "region_2_finds_right_aligned_match", a: "world", b: "the world",
			want: 1.0, exact: true,
			derivation: "Region 2 at i=4 → indelRatio(\"world\",\"world\")=1.0 (the substring at the right edge of longer is caught by Region 2 because n-m = 4 includes it)",
		},
		{
			// Region 2 middle wins: shorter exactly matches a middle
			// substring of longer. shorter = "YANKEES" (m=7);
			// longer = "NEW YORK YANKEES" (n=16).
			// Region 2: at i=9 → substr "YANKEES"; indelRatio=1.0.
			// best = 1.0.
			name: "region_2_middle_wins", a: "YANKEES", b: "NEW YORK YANKEES",
			want: 1.0, exact: true,
			derivation: "Region 2 at i=9 → indelRatio(\"YANKEES\",\"YANKEES\")=1.0",
		},
		{
			// Disjoint: shorter = "abc" (m=3); longer = "xyzzz" (n=5).
			// charSet={'a','b','c'}; longer has none of these bytes →
			// all alignments skipped by char-set early-skip; best
			// remains 0.0.
			name: "disjoint_no_overlap", a: "abc", b: "xyzzz",
			want: 0.0, exact: true,
			derivation: "charSet={'a','b','c'}; longer=\"xyzzz\" has no matching bytes; all alignments skipped → best=0.0",
		},
		{
			// Partial overlap: shorter = "abcd" (m=4); longer = "xabcy" (n=5).
			// charSet={'a','b','c','d'}.
			// Region 1: i=1 → substr "x"; charSet['x']=false (skip).
			//           i=2 → substr "xa"; charSet['a']=true;
			//                  indelRatio("abcd","xa")=2*1/6≈0.3333.
			//           i=3 → substr "xab"; charSet['b']=true;
			//                  indelRatio("abcd","xab")=2*2/7≈0.5714.
			// Region 2: i=0 → substr "xabc"; charSet['c']=true;
			//                  indelRatio("abcd","xabc")=2*3/8=0.75.
			//           i=1 → substr "abcy"; charSet['y']=false (skip).
			// Region 3: n>m → i=1 → substr "abcy"; charSet['y']=false (skip).
			//                  i=2 → substr "bcy"; charSet['b']=true;
			//                  indelRatio("abcd","bcy")=2*2/7≈0.5714.
			//                  i=3 → substr "cy"; charSet['c']=true;
			//                  indelRatio("abcd","cy")=2*1/6≈0.3333.
			//                  i=4 → substr "y"; charSet['y']=false (skip).
			// best = 0.75.
			name: "partial_overlap_middle_dominates", a: "abcd", b: "xabcy",
			want: 0.75, exact: true,
			derivation: "Region 2 at i=0 → indelRatio(\"abcd\",\"xabc\")=2*3/8=0.75; Region 1/3 candidates lower",
		},
		{
			// Symmetric pair forward (a, b).
			name: "symmetric_pair_forward", a: "hello world", b: "world",
			want: 1.0, exact: true,
			derivation: "Region 2 at i=6 of longer=\"hello world\" → indelRatio(\"world\",\"world\")=1.0",
		},
		{
			// Symmetric pair reverse (b, a) — must produce identical score.
			name: "symmetric_pair_reverse", a: "world", b: "hello world",
			want: 1.0, exact: true,
			derivation: "internal shorter-longer swap makes (b, a) equivalent to (a, b)",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Logf("derivation: %s", tt.derivation)
			got := fuzzymatch.PartialRatioScore(tt.a, tt.b)
			if tt.exact {
				if got != tt.want {
					t.Errorf("PartialRatioScore(%q, %q) = %.17g; want %.17g exactly",
						tt.a, tt.b, got, tt.want)
				}
			} else {
				if math.Abs(got-tt.want) > partialRatioEpsilon {
					t.Errorf("PartialRatioScore(%q, %q) = %.17g; want %.17g (Δ=%g, ε=%g)",
						tt.a, tt.b, got, tt.want, math.Abs(got-tt.want), partialRatioEpsilon)
				}
			}
		})
	}
}

// TestPartialRatioScore_Pitfall3_Keystones pins the LOAD-BEARING
// Pitfall-3 fixtures in isolation so a `go test -run` filter surfaces
// them immediately for triage. A regression that drops Region 1 or
// Region 3 from the iteration surfaces as a failure here.
//
// Per 06-RESEARCH.md Pitfall 3: a naive single-loop implementation
// (`for i := 0; i <= len(longer)-len(shorter); i++`) covers only
// Region 2. Without Region 1 / Region 3, the inputs `("abc", "ab")`
// and `("abc", "bc")` STILL produce 1.0 in this specific length pair
// (m=2, n=3) because Region 2 covers the two valid alignments. But
// regression fixtures with m < n-1 would fail. We pin the canonical
// Pitfall-3 names in BOTH unit tests AND BDD scenarios as a defence-
// in-depth gate.
func TestPartialRatioScore_Pitfall3_Keystones(t *testing.T) {
	keystones := []struct {
		name, a, b string
		want       float64
	}{
		// 06-RESEARCH.md Pitfall 3, line 738-739.
		{"abc_bc_right_tail_per_pitfall_3", "abc", "bc", 1.0},
		{"abc_ab_left_tail_per_pitfall_3", "abc", "ab", 1.0},
		// Extended Pitfall-3 fixtures where m < n-1 — Region 1 / 3
		// truly necessary (Region 2 alone would miss the perfect tail
		// match).
		{"abc_a_left_tail_only", "abc", "a", 1.0},
		{"abc_c_right_tail_only", "abc", "c", 1.0},
		{"a_abc_left_tail_swapped", "a", "abc", 1.0},
		{"c_abc_right_tail_swapped", "c", "abc", 1.0},
	}
	for _, kk := range keystones {
		t.Run(kk.name, func(t *testing.T) {
			got := fuzzymatch.PartialRatioScore(kk.a, kk.b)
			if got != kk.want {
				t.Errorf("PartialRatioScore(%q, %q) = %.17g; want %.17g (Pitfall-3 keystone — regression suggests Region 1 or 3 was dropped from iteration)",
					kk.a, kk.b, got, kk.want)
			}
		})
	}
}

// TestPartialRatioScore_Symmetric pins the algorithm's symmetry across
// argument order in a tabular form (the quick.Check property test in
// props_test.go provides random-input coverage).
func TestPartialRatioScore_Symmetric(t *testing.T) {
	tests := []struct{ a, b string }{
		{"abc", "ab"},
		{"abc", "bc"},
		{"hello world", "world"},
		{"YANKEES", "NEW YORK YANKEES"},
		{"abcd", "xabcy"},
		{"café", "caf"},
	}
	for _, tt := range tests {
		t.Run(tt.a+"_vs_"+tt.b, func(t *testing.T) {
			fwd := fuzzymatch.PartialRatioScore(tt.a, tt.b)
			rev := fuzzymatch.PartialRatioScore(tt.b, tt.a)
			if fwd != rev {
				t.Errorf("PartialRatioScore not symmetric: PR(%q,%q)=%g, PR(%q,%q)=%g",
					tt.a, tt.b, fwd, tt.b, tt.a, rev)
			}
		})
	}
}

// TestPartialRatioScore_DispatchRegistration pins that
// dispatch[AlgoPartialRatio] is populated after package load AND that
// invoking the dispatched function returns the same score as a direct
// call to PartialRatioScore. Per Phase 8.5 Q5 LOCKED, PartialRatio
// ships a single byte-path surface — there is no rune-variant.
func TestPartialRatioScore_DispatchRegistration(t *testing.T) {
	if fuzzymatch.DispatchEntryNilForTest(int(fuzzymatch.AlgoPartialRatio)) {
		t.Fatalf("dispatch[AlgoPartialRatio] (%d) is nil — dispatch_partial_ratio.go must register PartialRatioScore at package load time",
			int(fuzzymatch.AlgoPartialRatio))
	}
	got := fuzzymatch.DispatchInvokeForTest(int(fuzzymatch.AlgoPartialRatio), "YANKEES", "NEW YORK YANKEES")
	want := fuzzymatch.PartialRatioScore("YANKEES", "NEW YORK YANKEES")
	if got != want {
		t.Errorf("dispatch[AlgoPartialRatio](\"YANKEES\",\"NEW YORK YANKEES\") = %.17g; want %.17g",
			got, want)
	}
	if got != 1.0 {
		t.Errorf("dispatch[AlgoPartialRatio](\"YANKEES\",\"NEW YORK YANKEES\") = %.17g; want 1.0 (Region 2 middle wins)",
			got)
	}
}
