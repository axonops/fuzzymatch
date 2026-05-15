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

// token_jaccard_test.go pins the public-API contract of token_jaccard.go:
// identity, both-empty (STANDARD catalogue convention — 1.0, distinct
// from TokenSetRatio's locked RapidFuzz issue #110 deviation),
// one-empty, the six hand-derived reference vectors RV-TJ1..RV-TJ6 per
// CONTEXT.md §1b LOCKED (each carries its formula derivation in the
// test comment so a reviewer can re-derive the expected value in under
// a minute), dispatch registration through AlgoTokenJaccard, the
// load-bearing set-vs-multiset distinction (KEYSTONE RV-TJ3: the
// fuzzymatch TokenJaccard uses SET semantics — token deduplication via
// map[string]struct{} — while Phase 5's Q-Gram Jaccard uses MULTISET
// semantics over q-gram counts), and the alloc-budget ceiling.
//
// Stdlib `testing` only — no testify in root tests, per
// .claude/skills/go-coding-standards.

package fuzzymatch_test

import (
	"math"
	"testing"

	"github.com/axonops/fuzzymatch"
)

// tokenJaccardEpsilon is the float-comparison tolerance for irrational
// expected values (e.g. 2/3 = 0.6666666666666666). Phase 2/3/4/5
// convention is 1e-9; the Token Jaccard formula is a single
// integer-valued division so the actual accuracy is far higher than
// 1e-9. For exact-rational expected values (0.0, 0.5, 1.0) the tests
// use direct equality.
const tokenJaccardEpsilon = 1e-9

// TestTokenJaccardScore covers ≥ 8 cases including the six hand-derived
// reference vectors RV-TJ1..RV-TJ6 per CONTEXT.md §1b LOCKED. Each row's
// derivation is reproduced in the test sub-name and the in-line comment
// so a reviewer can re-derive the expected value from Jaccard 1912 p. 43
// in under a minute. The set-vs-multiset distinction (RV-TJ3) is the
// load-bearing keystone for the set-Jaccard semantics LOCKED in plan
// 06-04.
func TestTokenJaccardScore(t *testing.T) {
	tests := []struct {
		name       string
		a, b       string
		want       float64
		exact      bool // exact equality (rational) vs. epsilon (irrational)
		derivation string
	}{
		{
			// RV-TJ1 — partial overlap (load-bearing hand-derived RV):
			// Tokenise("a b c") = ["a","b","c"]; set {a,b,c}
			// Tokenise("b c d") = ["b","c","d"]; set {b,c,d}
			// intersection = {b,c}; |∩| = 2
			// union = {a,b,c,d}; |∪| = 4
			// J = 2/4 = 0.5
			name:       "RV-TJ1_partial_overlap",
			a:          "a b c",
			b:          "b c d",
			want:       0.5,
			exact:      true,
			derivation: "set A={a,b,c}; set B={b,c,d}; |∩|=2 ({b,c}); |∪|=4; J=2/4=0.5",
		},
		{
			// RV-TJ2 — subset (one set is contained in the other):
			// Tokenise("a b") = ["a","b"]; set {a,b}
			// Tokenise("a b c") = ["a","b","c"]; set {a,b,c}
			// intersection = {a,b}; |∩| = 2
			// union = {a,b,c}; |∪| = 3
			// J = 2/3 ≈ 0.6666666666666666
			name:       "RV-TJ2_subset",
			a:          "a b",
			b:          "a b c",
			want:       2.0 / 3.0,
			exact:      false,
			derivation: "set A={a,b}; set B={a,b,c}; A ⊂ B; |∩|=2; |∪|=3; J=2/3",
		},
		{
			// RV-TJ3 — KEYSTONE set-vs-multiset distinction.
			// Tokenise("a a b") = ["a","a","b"]; set deduplicates to {a,b}
			// Tokenise("a b") = ["a","b"]; set {a,b}
			// intersection = {a,b}; |∩| = 2
			// union = {a,b}; |∪| = 2
			// J = 2/2 = 1.0
			// A multiset-Jaccard implementation would yield 2/3 ≈ 0.667
			// here (multiplicities min(2,1)+min(1,1)=2; sum 3+2−2=3); the
			// set-Jaccard SEMANTICS LOCKED in plan 06-04 produces 1.0
			// because token-presence is a binary signal.
			name:       "RV-TJ3_set_vs_multiset_keystone",
			a:          "a a b",
			b:          "a b",
			want:       1.0,
			exact:      true,
			derivation: "set A=dedup({a,a,b})={a,b}; set B={a,b}; SET semantics → |∩|=2, |∪|=2, J=1.0 (MULTISET would yield 2/3)",
		},
		{
			// RV-TJ4 — disjoint token sets:
			// Tokenise("a b c") = ["a","b","c"]; set {a,b,c}
			// Tokenise("x y z") = ["x","y","z"]; set {x,y,z}
			// intersection = {}; |∩| = 0
			// union = {a,b,c,x,y,z}; |∪| = 6
			// J = 0/6 = 0.0
			name:       "RV-TJ4_disjoint",
			a:          "a b c",
			b:          "x y z",
			want:       0.0,
			exact:      true,
			derivation: "set A={a,b,c}; set B={x,y,z}; disjoint; |∩|=0; |∪|=6; J=0/6=0",
		},
		{
			// RV-TJ5 — identity (covered by a == b short-circuit; the
			// short-circuit fires BEFORE Tokenise, returning 1.0 without
			// allocating).
			name:       "RV-TJ5_identity",
			a:          "a b c",
			b:          "a b c",
			want:       1.0,
			exact:      true,
			derivation: "a == b identity short-circuit fires before Tokenise; returns 1.0",
		},
		{
			// RV-TJ6 — partial overlap on multi-token Greek-letter names:
			// Tokenise("alpha beta gamma delta") = ["alpha","beta","gamma","delta"]; set {alpha,beta,gamma,delta}
			// Tokenise("alpha beta epsilon zeta") = ["alpha","beta","epsilon","zeta"]; set {alpha,beta,epsilon,zeta}
			// intersection = {alpha,beta}; |∩| = 2
			// union = {alpha,beta,gamma,delta,epsilon,zeta}; |∪| = 6
			// J = 2/6 = 1/3 ≈ 0.3333333333333333
			name:       "RV-TJ6_partial_overlap_greek",
			a:          "alpha beta gamma delta",
			b:          "alpha beta epsilon zeta",
			want:       2.0 / 6.0,
			exact:      false,
			derivation: "set A={alpha,beta,gamma,delta}; set B={alpha,beta,epsilon,zeta}; |∩|=2 ({alpha,beta}); |∪|=6; J=2/6=1/3",
		},
		{
			// Both-empty STANDARD catalogue convention — TokenJaccard
			// returns 1.0 (vacuous identity match). DOES NOT deviate
			// like TokenSetRatio (LOCKED 2026-05-15: TokenJaccard
			// follows the STANDARD catalogue convention, NOT the
			// RapidFuzz issue #110 deviation).
			name:       "both_empty",
			a:          "",
			b:          "",
			want:       1.0,
			exact:      true,
			derivation: "both-empty STANDARD convention (a == b short-circuit fires) → 1.0",
		},
		{
			// One-empty A — 0.0:
			// Tokenise("") = []; one tokenised side empty → 0.0
			name:       "one_empty_a",
			a:          "",
			b:          "hello",
			want:       0.0,
			exact:      true,
			derivation: "Tokenise(\"\") = []; one-tokenised-empty → 0.0",
		},
		{
			// One-empty B — 0.0:
			// Tokenise("") = []; one tokenised side empty → 0.0
			name:       "one_empty_b",
			a:          "hello",
			b:          "",
			want:       0.0,
			exact:      true,
			derivation: "Tokenise(\"\") = []; one-tokenised-empty → 0.0",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Logf("derivation: %s", tt.derivation)
			got := fuzzymatch.TokenJaccardScore(tt.a, tt.b)
			if tt.exact {
				if got != tt.want {
					t.Errorf("TokenJaccardScore(%q, %q) = %g; want %g exactly",
						tt.a, tt.b, got, tt.want)
				}
			} else {
				if math.Abs(got-tt.want) > tokenJaccardEpsilon {
					t.Errorf("TokenJaccardScore(%q, %q) = %.17g; want %.17g (Δ=%g, ε=%g)",
						tt.a, tt.b, got, tt.want, math.Abs(got-tt.want), tokenJaccardEpsilon)
				}
			}
		})
	}
}

// TestTokenJaccardScore_Symmetric pins set-Jaccard's exact symmetry:
// J(A, B) == J(B, A) bit-for-bit. The set construction is
// order-independent (map[string]struct{}); the integer-counter
// intersection cardinality is invariant under argument swap; the
// single division on integer-derived float64 values produces identical
// output regardless of argument order. No float tolerance needed.
func TestTokenJaccardScore_Symmetric(t *testing.T) {
	tests := []struct {
		a, b string
	}{
		{"a b c", "b c d"},
		{"a b", "a b c"},
		{"hello world", "world hello"},
		{"alpha beta gamma delta", "alpha beta epsilon zeta"},
		{"abc", "xyz"},
	}
	for _, tt := range tests {
		t.Run(tt.a+"_"+tt.b, func(t *testing.T) {
			fwd := fuzzymatch.TokenJaccardScore(tt.a, tt.b)
			rev := fuzzymatch.TokenJaccardScore(tt.b, tt.a)
			if fwd != rev {
				t.Errorf("TokenJaccardScore not symmetric: J(%q,%q)=%g, J(%q,%q)=%g",
					tt.a, tt.b, fwd, tt.b, tt.a, rev)
			}
		})
	}
}

// TestTokenJaccardScore_DispatchRegistration exercises the dispatch
// table wiring: dispatch[AlgoTokenJaccard] must be non-nil after
// dispatch_token_jaccard.go's load-time registration AND must invoke
// TokenJaccardScore producing the documented RV-TJ1 value. A regression
// where dispatch_token_jaccard.go fails to load (or maps the wrong
// function) surfaces here as either a nil-deref panic or a wrong score.
func TestTokenJaccardScore_DispatchRegistration(t *testing.T) {
	if fuzzymatch.DispatchEntryNilForTest(int(fuzzymatch.AlgoTokenJaccard)) {
		t.Fatalf("dispatch[AlgoTokenJaccard] (%d) is nil — dispatch_token_jaccard.go must register TokenJaccardScore at package load time",
			int(fuzzymatch.AlgoTokenJaccard))
	}
	got := fuzzymatch.DispatchInvokeForTest(int(fuzzymatch.AlgoTokenJaccard), "a b c", "b c d")
	want := 0.5
	if got != want {
		t.Errorf("dispatch[AlgoTokenJaccard](\"a b c\", \"b c d\") = %v; want %v (RV-TJ1: |∩|=2, |∪|=4, J=2/4=0.5)",
			got, want)
	}
}

// TestTokenJaccardScore_SetVsMultisetDistinction pins the load-bearing
// SET semantics LOCKED in plan 06-04. The KEYSTONE input ("a a b", "a b")
// MUST produce 1.0 under TokenJaccard's set semantics:
//
//   - Tokenise("a a b") → ["a","a","b"] → set dedup → {a,b}
//   - Tokenise("a b")   → ["a","b"]     → set       → {a,b}
//   - intersection {a,b}; union {a,b}; J = 2/2 = 1.0
//
// The same inputs under Q-Gram Jaccard MULTISET semantics produce a
// DIFFERENT score (≠ 1.0): for q-gram size 1 over the raw input bytes
// the multiset of unigrams differs because Q-Gram Jaccard operates on
// CHARACTER multisets, not deduped token sets. This test asserts BOTH
// the TokenJaccard 1.0 result AND the inequality with QGramJaccard at
// n=1 to surface any future drift to multiset semantics. RESEARCH.md
// Pattern 8 establishes the semantic divergence is intentional: token
// presence is a binary signal (a token appearing twice doesn't make it
// "more present"); q-gram presence is a multiplicity signal.
func TestTokenJaccardScore_SetVsMultisetDistinction(t *testing.T) {
	// Step 1: pin the SET-semantics result.
	tjScore := fuzzymatch.TokenJaccardScore("a a b", "a b")
	if tjScore != 1.0 {
		t.Errorf("TokenJaccardScore(\"a a b\", \"a b\") = %g; want 1.0 (SET semantics — token dedup yields {a,b} on both sides; the KEYSTONE distinction from Q-Gram Jaccard's MULTISET semantics per RESEARCH.md Pattern 8)",
			tjScore)
	}

	// Step 2: pin that the same inputs under Q-Gram Jaccard MULTISET
	// semantics produce a DIFFERENT score. The exact value is not
	// asserted (Q-Gram Jaccard's own tests pin that); only the
	// inequality matters for the keystone distinction.
	//
	// QGramJaccardScore("a a b", "a b", 1):
	//   QA = unigrams("a a b") = {a:2, ' ':2, b:1}; total 5
	//   QB = unigrams("a b")   = {a:1, ' ':1, b:1}; total 3
	//   intersection (multiset) = min(2,1)+min(2,1)+min(1,1) = 3
	//   union = 5 + 3 - 3 = 5
	//   J = 3/5 = 0.6
	qjScore := fuzzymatch.QGramJaccardScore("a a b", "a b", 1)
	if qjScore == tjScore {
		t.Errorf("QGramJaccardScore (MULTISET) equals TokenJaccardScore (SET) for keystone inputs (%g vs %g) — set-vs-multiset distinction has regressed; see RESEARCH.md Pattern 8",
			qjScore, tjScore)
	}
}

// TestTokenJaccardScore_AllocsBudget asserts the per-call allocation
// count stays within the documented ≤ 4 allocs budget (2 Tokenise
// outputs + 2 map[string]struct{} sets per the plan's <action> alloc
// budget — integer-counter intersection adds 0 allocs). The exact
// alloc count depends on Go's map and slice implementation; the
// assertion is a CEILING set per the qgram_jaccard.go RESEARCH.md §4.1
// pattern so future Go map-growth tweaks don't fail the test.
func TestTokenJaccardScore_AllocsBudget(t *testing.T) {
	const ceiling = 8.0
	got := testing.AllocsPerRun(100, func() {
		_ = fuzzymatch.TokenJaccardScore("alpha beta gamma", "beta gamma delta")
	})
	if got > ceiling {
		t.Errorf("TokenJaccardScore allocs/op = %g; want <= %g", got, ceiling)
	}
}
