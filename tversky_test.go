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

// tversky_test.go pins the public-API contract of tversky.go: identity,
// both-empty, one-empty, the four canonical hand-derived reference
// vectors RV-T1..RV-T4 from RESEARCH.md §2.4 (covering the load-bearing
// asymmetric direction-sensitive pair RV-T1/RV-T2 plus the Jaccard
// (RV-T3) and Dice (RV-T4) cross-check pairs), the rune-path café
// reference, the LOAD-BEARING TestTversky_AsymmetryDirectionSensitive
// gate that asserts RV-T1 ≠ RV-T2, the bit-exact JaccardCrossCheck and
// DiceCrossCheck cross-algorithm consistency tests, the parameter-swap
// symmetry (T(a,b,α,β) = T(b,a,β,α)) and the α=β symmetry (T(a,b,α,α) =
// T(b,a,α,α)) algebraic identity tests, the direct-call panic-on-n<1
// AND panic-on-invalid-α/β contracts, and the alloc budget ceiling.
//
// Stdlib `testing` only — no testify in root tests, per
// .claude/skills/go-coding-standards.

package fuzzymatch_test

import (
	"errors"
	"math"
	"strconv"
	"strings"
	"testing"

	"github.com/axonops/fuzzymatch"
)

// tverskyEpsilon is the float-comparison tolerance for irrational
// expected values (e.g. 3/3.4 = 0.8823529411764706). The Tversky
// formula is a single multiplication-then-addition-then-division on
// integer-valued counts so the actual accuracy is far higher than 1e-15;
// the 1e-15 tolerance per the plan is the locked convention. For
// exact-rational expected values (0.0, 0.5, 1.0) the tests use direct
// equality.
const tverskyEpsilon = 1e-15

// TestTversky_BothEmpty pins the both-empty convention:
// TverskyScore("", "", n, α, β) == 1.0 (vacuous match) — covered by
// the a == b identity short-circuit. Tested across multiple α/β
// configurations to confirm the identity short-circuit fires
// REGARDLESS of α/β (including α=β=0.5 — which would otherwise be
// the most ambiguous default).
func TestTversky_BothEmpty(t *testing.T) {
	type ab struct{ alpha, beta float64 }
	configs := []ab{
		{0.5, 0.5},
		{0.8, 0.2},
		{0.2, 0.8},
		{1.0, 1.0},
	}
	for _, n := range []int{1, 2, 3, 5} {
		for _, c := range configs {
			if got := fuzzymatch.TverskyScore("", "", n, c.alpha, c.beta); got != 1.0 {
				t.Errorf("TverskyScore(\"\", \"\", %d, %g, %g) = %g; want 1.0", n, c.alpha, c.beta, got)
			}
			if got := fuzzymatch.TverskyScoreRunes("", "", n, c.alpha, c.beta); got != 1.0 {
				t.Errorf("TverskyScoreRunes(\"\", \"\", %d, %g, %g) = %g; want 1.0", n, c.alpha, c.beta, got)
			}
		}
	}
}

// TestTversky_OneEmpty pins the one-empty convention: 0.0 in both
// directions (asymmetric short-circuit gates the identity path away
// before reaching extraction). Tested across multiple α/β
// configurations.
func TestTversky_OneEmpty(t *testing.T) {
	tests := []struct{ a, b string }{
		{"", "abc"},
		{"abc", ""},
		{"", "x"},
		{"x", ""},
		{"", "café"},
		{"café", ""},
	}
	configs := []struct{ alpha, beta float64 }{
		{0.5, 0.5},
		{0.8, 0.2},
		{0.0, 0.5}, // α=0 with β>0 is valid — exercises the boundary
	}
	for _, tt := range tests {
		for _, c := range configs {
			t.Run(tt.a+"_"+tt.b, func(t *testing.T) {
				if got := fuzzymatch.TverskyScore(tt.a, tt.b, 2, c.alpha, c.beta); got != 0.0 {
					t.Errorf("TverskyScore(%q, %q, 2, %g, %g) = %g; want 0.0", tt.a, tt.b, c.alpha, c.beta, got)
				}
				if got := fuzzymatch.TverskyScoreRunes(tt.a, tt.b, 2, c.alpha, c.beta); got != 0.0 {
					t.Errorf("TverskyScoreRunes(%q, %q, 2, %g, %g) = %g; want 0.0", tt.a, tt.b, c.alpha, c.beta, got)
				}
			})
		}
	}
}

// TestTversky_Identical pins the identity short-circuit: any
// non-empty x returns 1.0 for any n >= 1 and any valid (α, β) — the
// a == b guard fires before extraction (so α/β are irrelevant to the
// outcome on identical inputs).
func TestTversky_Identical(t *testing.T) {
	tests := []string{"abc", "user_id", "x", "WIKIMEDIA", "café", "AGCT", "hello"}
	configs := []struct{ alpha, beta float64 }{
		{0.5, 0.5},
		{0.8, 0.2},
		{0.2, 0.8},
		{1.0, 1.0},
		{0.0, 0.5}, // boundary: α=0 still valid when β>0
	}
	for _, s := range tests {
		t.Run(s, func(t *testing.T) {
			for _, n := range []int{1, 2, 3, 5} {
				for _, c := range configs {
					if got := fuzzymatch.TverskyScore(s, s, n, c.alpha, c.beta); got != 1.0 {
						t.Errorf("TverskyScore(%q, %q, %d, %g, %g) = %g; want 1.0",
							s, s, n, c.alpha, c.beta, got)
					}
					if got := fuzzymatch.TverskyScoreRunes(s, s, n, c.alpha, c.beta); got != 1.0 {
						t.Errorf("TverskyScoreRunes(%q, %q, %d, %g, %g) = %g; want 1.0",
							s, s, n, c.alpha, c.beta, got)
					}
				}
			}
		})
	}
}

// TestTversky_ReferenceVectors pins the four canonical hand-derived
// reference vectors (RV-T1..RV-T4) from RESEARCH.md §2.4. Each row's
// derivation is reproduced in the test sub-name and the in-line comment
// so a reviewer can re-derive the expected value from Tversky 1977 §2
// in under a minute.
//
// RV-T1 + RV-T2 are the LOAD-BEARING asymmetric direction-sensitive
// pair (input swap with fixed α=0.8, β=0.2 produces different scores
// — see TestTversky_AsymmetryDirectionSensitive below for the explicit
// inequality assertion).
//
// RV-T3 (α=β=1.0 → Jaccard) and RV-T4 (α=β=0.5 → Dice) are the
// degeneracy cross-check vectors. The bit-exact equality with
// QGramJaccardScore / SorensenDiceScore is asserted by the dedicated
// TestTversky_JaccardCrossCheck / TestTversky_DiceCrossCheck tests
// below.
func TestTversky_ReferenceVectors(t *testing.T) {
	tests := []struct {
		name        string
		a, b        string
		n           int
		alpha, beta float64
		want        float64
		exact       bool // exact equality (rational) vs. epsilon (irrational)
		derivation  string
	}{
		{
			// RV-T1: Tversky 1977 §2 prototype/variant — asymmetric
			// direction-sensitive. The first half of the load-bearing
			// asymmetry pair.
			// QA = bigrams("abcd")   = {ab:1, bc:1, cd:1}            — total 3
			// QB = bigrams("abcdef") = {ab:1, bc:1, cd:1, de:1, ef:1} — total 5
			// |A∩B| = 3; |A−B| = 0; |B−A| = 2
			// T = 3/(3 + 0.8·0 + 0.2·2) = 3/3.4 = 0.8823529411764706
			name:       "RV-T1_abcd_abcdef_n2_a0.8_b0.2",
			a:          "abcd",
			b:          "abcdef",
			n:          2,
			alpha:      0.8,
			beta:       0.2,
			want:       0.8823529411764706,
			exact:      false,
			derivation: "QA={ab,bc,cd}; QB={ab,bc,cd,de,ef}; |∩|=3; |A−B|=0; |B−A|=2; T=3/(3+0.8·0+0.2·2)=3/3.4",
		},
		{
			// RV-T2: input swap of RV-T1 — LOAD-BEARING asymmetry-
			// discriminating pair. Same (α, β); inputs swapped; score
			// differs.
			// QA = bigrams("abcdef") = {ab:1, bc:1, cd:1, de:1, ef:1} — total 5
			// QB = bigrams("abcd")   = {ab:1, bc:1, cd:1}            — total 3
			// |A∩B| = 3; |A−B| = 2; |B−A| = 0
			// T = 3/(3 + 0.8·2 + 0.2·0) = 3/4.6 = 0.6521739130434783
			name:       "RV-T2_abcdef_abcd_n2_a0.8_b0.2",
			a:          "abcdef",
			b:          "abcd",
			n:          2,
			alpha:      0.8,
			beta:       0.2,
			want:       0.6521739130434783,
			exact:      false,
			derivation: "QA={ab,bc,cd,de,ef}; QB={ab,bc,cd}; |∩|=3; |A−B|=2; |B−A|=0; T=3/(3+0.8·2+0.2·0)=3/4.6",
		},
		{
			// RV-T3: α=β=1.0 reduces to Jaccard. Cross-check pair.
			// QA = bigrams("abcd") = {ab:1, bc:1, cd:1}; |QA|=3
			// QB = bigrams("abce") = {ab:1, bc:1, ce:1}; |QB|=3
			// |A∩B| = 2 (ab + bc); |A−B| = 1 (cd); |B−A| = 1 (ce)
			// T = 2/(2 + 1·1 + 1·1) = 2/4 = 0.5
			// Jaccard: |∩|/|∪| = 2/(3+3-2) = 2/4 = 0.5 ✓
			name:       "RV-T3_abcd_abce_n2_a1.0_b1.0",
			a:          "abcd",
			b:          "abce",
			n:          2,
			alpha:      1.0,
			beta:       1.0,
			want:       0.5,
			exact:      true,
			derivation: "QA={ab,bc,cd}; QB={ab,bc,ce}; |∩|=2; |A−B|=1; |B−A|=1; T=2/(2+1+1)=0.5; Jaccard=2/4=0.5",
		},
		{
			// RV-T4: α=β=0.5 reduces to Sørensen-Dice. Cross-check pair.
			// Same q-grams as RV-T3.
			// T = 2/(2 + 0.5·1 + 0.5·1) = 2/3 = 0.6666666666666666
			// Dice: 2·|∩|/(|QA|+|QB|) = 2·2/(3+3) = 4/6 ≈ 0.6666666666666666 ✓
			name:       "RV-T4_abcd_abce_n2_a0.5_b0.5",
			a:          "abcd",
			b:          "abce",
			n:          2,
			alpha:      0.5,
			beta:       0.5,
			want:       2.0 / 3.0,
			exact:      false,
			derivation: "Same q-grams as RV-T3; T=2/(2+0.5·1+0.5·1)=2/3; Dice=4/6=2/3",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Logf("derivation: %s", tt.derivation)
			got := fuzzymatch.TverskyScore(tt.a, tt.b, tt.n, tt.alpha, tt.beta)
			if tt.exact {
				if got != tt.want {
					t.Errorf("TverskyScore(%q, %q, %d, %g, %g) = %g; want %g exactly",
						tt.a, tt.b, tt.n, tt.alpha, tt.beta, got, tt.want)
				}
			} else {
				if math.Abs(got-tt.want) > tverskyEpsilon {
					t.Errorf("TverskyScore(%q, %q, %d, %g, %g) = %.17g; want %.17g (Δ=%g, ε=%g)",
						tt.a, tt.b, tt.n, tt.alpha, tt.beta, got, tt.want, math.Abs(got-tt.want), tverskyEpsilon)
				}
			}
		})
	}
}

// TestTversky_AsymmetryDirectionSensitive is the LOAD-BEARING
// asymmetry gate. Without this test, the implementation could silently
// swap α and β (and still pass RangeBounds + Identity property tests)
// because the swap simply reflects the score from RV-T1 to RV-T2 — both
// values are still in [0, 1], both still satisfy identity on equal
// inputs, and only this dedicated input-swap discrimination catches
// the bug.
//
// Computes RV-T1 = TverskyScore("abcd", "abcdef", 2, 0.8, 0.2) and
// RV-T2 = TverskyScore("abcdef", "abcd", 2, 0.8, 0.2) and asserts
// |RV-T1 − RV-T2| > 0.1. The actual difference is ≈ 0.2302
// (0.8823529411764706 − 0.6521739130434783 = 0.2301790281330...).
//
// On failure: a t.Errorf message about parameter-order regression
// (asymmetry-gate failure → silent α/β swap is the most likely root
// cause).
func TestTversky_AsymmetryDirectionSensitive(t *testing.T) {
	rvT1 := fuzzymatch.TverskyScore("abcd", "abcdef", 2, 0.8, 0.2)
	rvT2 := fuzzymatch.TverskyScore("abcdef", "abcd", 2, 0.8, 0.2)
	delta := math.Abs(rvT1 - rvT2)
	const minDelta = 0.1
	if delta <= minDelta {
		t.Errorf(`Tversky asymmetry gate FAILED — input swap with fixed (α=0.8, β=0.2) did not produce a meaningfully different score:
  TverskyScore("abcd", "abcdef", 2, 0.8, 0.2) = %.17g  (RV-T1; expected 0.8823529411764706)
  TverskyScore("abcdef", "abcd", 2, 0.8, 0.2) = %.17g  (RV-T2; expected 0.6521739130434783)
  |RV-T1 − RV-T2| = %g; want > %g
This test is the load-bearing regression detector for parameter-order
correctness. The most likely cause of failure is a silent α/β swap
inside TverskyScore — verify the formula is
  T = |A∩B| / (|A∩B| + α·|A−B| + β·|B−A|)
NOT
  T = |A∩B| / (|A∩B| + β·|A−B| + α·|B−A|)
where α weighs the FIRST argument's residuals (|A−B|) and β weighs the
SECOND (|B−A|).`,
			rvT1, rvT2, delta, minDelta)
	}
	// Pin both endpoint values inline (defence-in-depth — if the asymmetry
	// gate fires AND the values happen to land outside the inequality, the
	// per-value pin still surfaces the regression).
	const wantT1 = 0.8823529411764706
	const wantT2 = 0.6521739130434783
	if math.Abs(rvT1-wantT1) > tverskyEpsilon {
		t.Errorf("RV-T1 endpoint pin: got %.17g; want %.17g (Δ=%g, ε=%g)",
			rvT1, wantT1, math.Abs(rvT1-wantT1), tverskyEpsilon)
	}
	if math.Abs(rvT2-wantT2) > tverskyEpsilon {
		t.Errorf("RV-T2 endpoint pin: got %.17g; want %.17g (Δ=%g, ε=%g)",
			rvT2, wantT2, math.Abs(rvT2-wantT2), tverskyEpsilon)
	}
}

// TestTversky_JaccardCrossCheck pins the algebraic identity
// TverskyScore(a, b, n, 1.0, 1.0) == QGramJaccardScore(a, b, n)
// bit-for-bit (via math.Float64bits comparison). Tversky 1977 §2 with
// α=β=1 reduces to the Jaccard coefficient on multisets — the
// equivalence holds because |A∪B| = |A| + |B| − |A∩B| for multisets
// where |A−B| = |A| − |A∩B| (and similarly for |B−A|), which collapses
// the Tversky denominator to exactly the Jaccard union.
//
// The bit-exact assertion guards against any future drift in either
// algorithm's reduction order. A single ULP difference would surface
// here; an algorithm-correctness reviewer can re-derive the equivalence
// from the formulae above.
func TestTversky_JaccardCrossCheck(t *testing.T) {
	tests := []struct {
		a, b string
		n    int
	}{
		{"abcd", "abce", 2},     // RV-T3 anchor
		{"AGCT", "AGCTAGCT", 2}, // Ukkonen 1992 §3 worked example (RV-J1)
		{"abc", "xyz", 2},       // orthogonal
		{"abcd", "abxy", 2},     // single-shared bigram (RV-J4)
		{"abcdef", "bcdefg", 2}, // high-overlap (RV-D2 anchor; works for Jaccard too)
		{"hello", "world", 2},
		{"abcdefgh", "abcdefgi", 3}, // n=3 (large intersection)
		{"x", "y", 1},               // n=1 unigram
	}
	for _, tt := range tests {
		t.Run(tt.a+"_"+tt.b, func(t *testing.T) {
			tv := fuzzymatch.TverskyScore(tt.a, tt.b, tt.n, 1.0, 1.0)
			jc := fuzzymatch.QGramJaccardScore(tt.a, tt.b, tt.n)
			if math.Float64bits(tv) != math.Float64bits(jc) {
				t.Errorf("Tversky(α=β=1) ≠ Jaccard for (%q, %q, n=%d): Tversky=%.17g (bits=%x), Jaccard=%.17g (bits=%x)",
					tt.a, tt.b, tt.n,
					tv, math.Float64bits(tv),
					jc, math.Float64bits(jc))
			}
			// Mirror on the rune surface.
			tvR := fuzzymatch.TverskyScoreRunes(tt.a, tt.b, tt.n, 1.0, 1.0)
			jcR := fuzzymatch.QGramJaccardScoreRunes(tt.a, tt.b, tt.n)
			if math.Float64bits(tvR) != math.Float64bits(jcR) {
				t.Errorf("TverskyRunes(α=β=1) ≠ JaccardRunes for (%q, %q, n=%d): Tversky=%.17g, Jaccard=%.17g",
					tt.a, tt.b, tt.n, tvR, jcR)
			}
		})
	}
}

// TestTversky_DiceCrossCheck pins the algebraic identity
// TverskyScore(a, b, n, 0.5, 0.5) == SorensenDiceScore(a, b, n)
// bit-for-bit. With α=β=0.5 the Tversky denominator becomes
// |A∩B| + 0.5·(|A−B| + |B−A|) = (|A| + |B|)/2; multiplying through:
// T = 2·|A∩B|/(|A| + |B|) = DSC.
//
// As with TestTversky_JaccardCrossCheck, bit-exact equality is the
// gate — any drift surfaces immediately.
func TestTversky_DiceCrossCheck(t *testing.T) {
	tests := []struct {
		a, b string
		n    int
	}{
		{"abcd", "abce", 2},     // RV-T4 anchor
		{"night", "nacht", 2},   // RV-D1 canonical NLP-textbook bigram pair
		{"abcdef", "bcdefg", 2}, // RV-D2 high-overlap
		{"abcdef", "abcXef", 3}, // RV-D3 trigram variant
		{"abc", "xyz", 2},       // orthogonal
		{"hello", "world", 2},
		{"x", "y", 1},               // n=1 unigram
		{"abcdefgh", "abcdefgi", 3}, // n=3
	}
	for _, tt := range tests {
		t.Run(tt.a+"_"+tt.b, func(t *testing.T) {
			tv := fuzzymatch.TverskyScore(tt.a, tt.b, tt.n, 0.5, 0.5)
			dc := fuzzymatch.SorensenDiceScore(tt.a, tt.b, tt.n)
			if math.Float64bits(tv) != math.Float64bits(dc) {
				t.Errorf("Tversky(α=β=0.5) ≠ Dice for (%q, %q, n=%d): Tversky=%.17g (bits=%x), Dice=%.17g (bits=%x)",
					tt.a, tt.b, tt.n,
					tv, math.Float64bits(tv),
					dc, math.Float64bits(dc))
			}
			// Mirror on the rune surface.
			tvR := fuzzymatch.TverskyScoreRunes(tt.a, tt.b, tt.n, 0.5, 0.5)
			dcR := fuzzymatch.SorensenDiceScoreRunes(tt.a, tt.b, tt.n)
			if math.Float64bits(tvR) != math.Float64bits(dcR) {
				t.Errorf("TverskyRunes(α=β=0.5) ≠ DiceRunes for (%q, %q, n=%d): Tversky=%.17g, Dice=%.17g",
					tt.a, tt.b, tt.n, tvR, dcR)
			}
		})
	}
}

// TestTversky_ParameterSwapSymmetry pins the Tversky 1977 §2 algebraic
// identity TverskyScore(a, b, n, α, β) == TverskyScore(b, a, n, β, α)
// bit-for-bit. This is a STRONGER claim than asymmetry-when-α≠β: it
// asserts that swapping the inputs is EXACTLY equivalent to swapping
// the parameters. If the implementation silently swapped α and β
// internally, this property would still hold (vacuously); but combined
// with TestTversky_AsymmetryDirectionSensitive (which pins the actual
// numeric values), the two together rule out parameter-order bugs.
func TestTversky_ParameterSwapSymmetry(t *testing.T) {
	tests := []struct {
		a, b        string
		n           int
		alpha, beta float64
	}{
		{"abcd", "abcdef", 2, 0.8, 0.2}, // RV-T1/T2 anchor
		{"abcd", "abce", 2, 1.0, 1.0},   // RV-T3 (degenerate; α=β so trivially holds)
		{"abcd", "abce", 2, 0.5, 0.5},   // RV-T4 (degenerate)
		{"abc", "xyz", 2, 0.7, 0.3},     // orthogonal
		{"hello", "world", 2, 0.9, 0.1}, // short asymmetric
		{"AGCT", "AGCTAGCT", 2, 0.3, 0.7},
		{"abcdef", "bcdefg", 3, 0.4, 0.6},
		{"x", "y", 1, 0.0, 0.5}, // boundary: α=0 with β>0
		{"x", "y", 1, 0.5, 0.0}, // boundary: β=0 with α>0
	}
	for _, tt := range tests {
		t.Run(tt.a+"_"+tt.b, func(t *testing.T) {
			fwd := fuzzymatch.TverskyScore(tt.a, tt.b, tt.n, tt.alpha, tt.beta)
			rev := fuzzymatch.TverskyScore(tt.b, tt.a, tt.n, tt.beta, tt.alpha)
			if math.Float64bits(fwd) != math.Float64bits(rev) {
				t.Errorf("Tversky parameter-swap symmetry violated: T(%q,%q,%d,%g,%g)=%.17g (bits=%x); T(%q,%q,%d,%g,%g)=%.17g (bits=%x)",
					tt.a, tt.b, tt.n, tt.alpha, tt.beta, fwd, math.Float64bits(fwd),
					tt.b, tt.a, tt.n, tt.beta, tt.alpha, rev, math.Float64bits(rev))
			}
			// Mirror on the rune surface.
			fwdR := fuzzymatch.TverskyScoreRunes(tt.a, tt.b, tt.n, tt.alpha, tt.beta)
			revR := fuzzymatch.TverskyScoreRunes(tt.b, tt.a, tt.n, tt.beta, tt.alpha)
			if math.Float64bits(fwdR) != math.Float64bits(revR) {
				t.Errorf("TverskyRunes parameter-swap symmetry violated: T(%q,%q,%d,%g,%g)=%.17g; T(%q,%q,%d,%g,%g)=%.17g",
					tt.a, tt.b, tt.n, tt.alpha, tt.beta, fwdR,
					tt.b, tt.a, tt.n, tt.beta, tt.alpha, revR)
			}
		})
	}
}

// TestTversky_SymmetricWhenAlphaEqBeta pins the corollary of the
// parameter-swap symmetry: when α = β, TverskyScore(a, b, α, α) ==
// TverskyScore(b, a, α, α) bit-for-bit. The function is symmetric in
// the inputs whenever the weights are equal — this is the "Tversky
// degenerates to a symmetric coefficient" property used implicitly by
// the Jaccard / Dice cross-checks.
func TestTversky_SymmetricWhenAlphaEqBeta(t *testing.T) {
	tests := []struct {
		a, b  string
		n     int
		alpha float64
	}{
		{"abcd", "abcdef", 2, 0.5},
		{"abcd", "abce", 2, 1.0},
		{"abcd", "abce", 2, 0.5},
		{"abc", "xyz", 2, 0.7},
		{"hello", "world", 2, 0.9},
		{"AGCT", "AGCTAGCT", 2, 0.3},
		{"abcdef", "bcdefg", 3, 1.0},
	}
	for _, tt := range tests {
		t.Run(tt.a+"_"+tt.b, func(t *testing.T) {
			fwd := fuzzymatch.TverskyScore(tt.a, tt.b, tt.n, tt.alpha, tt.alpha)
			rev := fuzzymatch.TverskyScore(tt.b, tt.a, tt.n, tt.alpha, tt.alpha)
			if math.Float64bits(fwd) != math.Float64bits(rev) {
				t.Errorf("Tversky α=β symmetry violated: T(%q,%q,%d,%g,%g)=%.17g; T(%q,%q,%d,%g,%g)=%.17g",
					tt.a, tt.b, tt.n, tt.alpha, tt.alpha, fwd,
					tt.b, tt.a, tt.n, tt.alpha, tt.alpha, rev)
			}
		})
	}
}

// TestTverskyRunes_CafeReference pins the rune-path reference vector:
// "café" / "cafe" / n=2 / α=β=0.5 → 2/3 ≈ 0.6666666666666666.
//
// QA = rune-bigrams("café") = {"ca":1, "af":1, "fé":1}  — total 3
// QB = rune-bigrams("cafe") = {"ca":1, "af":1, "fe":1}  — total 3
// |A∩B| = 2 (ca, af); |A−B| = 1 (fé); |B−A| = 1 (fe);
// T = 2 / (2 + 0.5·1 + 0.5·1) = 2 / 3 ≈ 0.6666666666666666
//
// Pinning this load-bearing rune-path vector ensures the rune
// extractor and the algorithm wire up correctly — a regression where
// the byte path is silently called would yield a different score (the
// byte path treats "café" as 5 bytes and produces different bigrams).
func TestTverskyRunes_CafeReference(t *testing.T) {
	got := fuzzymatch.TverskyScoreRunes("café", "cafe", 2, 0.5, 0.5)
	want := 2.0 / 3.0
	if math.Abs(got-want) > tverskyEpsilon {
		t.Errorf("TverskyScoreRunes(\"café\", \"cafe\", 2, 0.5, 0.5) = %.17g; want %.17g (Δ=%g, ε=%g)",
			got, want, math.Abs(got-want), tverskyEpsilon)
	}
	// Cross-check: at α=β=0.5 this should equal Sørensen-Dice on the
	// same rune bigrams.
	dice := fuzzymatch.SorensenDiceScoreRunes("café", "cafe", 2)
	if math.Float64bits(got) != math.Float64bits(dice) {
		t.Errorf("TverskyScoreRunes(α=β=0.5) ≠ SorensenDiceScoreRunes on (café, cafe, n=2): Tversky=%.17g, Dice=%.17g",
			got, dice)
	}
}

// TestTversky_PanicsOnInvalidN pins the direct-call panic-on-n<1
// contract per the Phase 8.5 Q2 data-vs-parameter framework
// (docs/requirements.md §6.A): direct calls panic with a TYPED-ERROR
// value `fmt.Errorf("%w: …", ErrInvalidQGramSize, …)` so recover()
// callers can discriminate via `errors.Is(panicValue.(error),
// ErrInvalidQGramSize)`. The Error() text of the wrapping error still
// contains "invalid q-gram size" (the sentinel's own message) so
// log-scraping consumers are unaffected.
//
// Note the n-validation happens AFTER the a == b identity short-circuit
// (so identical inputs at any n return 1.0 without panicking) but
// BEFORE any extraction work.
func TestTversky_PanicsOnInvalidN(t *testing.T) {
	tests := []int{0, -1, -100, math.MinInt32}
	for _, n := range tests {
		t.Run("n_"+strconv.Itoa(n), func(t *testing.T) {
			// Byte path. Use distinct inputs so a == b short-circuit
			// does not fire (the panic gate only triggers for differing
			// inputs).
			func() {
				defer func() {
					r := recover()
					if r == nil {
						t.Errorf("TverskyScore(\"abc\", \"abd\", %d, 0.5, 0.5) did not panic", n)
						return
					}
					err, ok := r.(error)
					if !ok {
						t.Errorf("panic value type = %T (%v); want error wrapping ErrInvalidQGramSize", r, r)
						return
					}
					if !errors.Is(err, fuzzymatch.ErrInvalidQGramSize) {
						t.Errorf("panic err = %v; want errors.Is(_, ErrInvalidQGramSize)", err)
					}
					if !strings.Contains(err.Error(), "invalid q-gram size") {
						t.Errorf("panic message %q does not contain \"invalid q-gram size\"", err.Error())
					}
				}()
				_ = fuzzymatch.TverskyScore("abc", "abd", n, 0.5, 0.5)
			}()
			// Rune path.
			func() {
				defer func() {
					r := recover()
					if r == nil {
						t.Errorf("TverskyScoreRunes(\"abc\", \"abd\", %d, 0.5, 0.5) did not panic", n)
						return
					}
					err, ok := r.(error)
					if !ok {
						t.Errorf("panic value type = %T (%v); want error wrapping ErrInvalidQGramSize", r, r)
						return
					}
					if !errors.Is(err, fuzzymatch.ErrInvalidQGramSize) {
						t.Errorf("panic err = %v; want errors.Is(_, ErrInvalidQGramSize)", err)
					}
					if !strings.Contains(err.Error(), "invalid q-gram size") {
						t.Errorf("panic message %q does not contain \"invalid q-gram size\"", err.Error())
					}
				}()
				_ = fuzzymatch.TverskyScoreRunes("abc", "abd", n, 0.5, 0.5)
			}()
		})
	}
}

// TestTversky_PanicsOnInvalidParams pins the direct-call panic-on-
// invalid-α/β contract per the Phase 8.5 Q2 data-vs-parameter framework
// (docs/requirements.md §6.A). Direct calls panic with a TYPED-ERROR
// value `fmt.Errorf("%w: …", ErrInvalidTverskyParam, …)` so recover()
// callers can discriminate via `errors.Is(panicValue.(error),
// ErrInvalidTverskyParam)`. The Error() text still contains "invalid
// tversky parameter" (the sentinel's own message) so log-scraping
// consumers are unaffected.
//
// Failure modes (all share the same sentinel):
//
//   - α or β is NaN (Phase 8.5 Q2 NaN guard)
//   - α or β is ±Inf (Phase 8.5 Q2 Inf guard)
//   - α < 0 (any β ≥ 0)
//   - β < 0 (any α ≥ 0)
//   - α == 0 AND β == 0 (denominator-zero risk if intersection==0 too)
//
// The α == 0 case with β > 0 (and vice versa) is VALID and must NOT
// panic — covered by TestTversky_ZeroAlphaWithPositiveBeta below.
//
// As with the n-panic test, distinct inputs are used so the a == b
// short-circuit does not gate away the panic.
func TestTversky_PanicsOnInvalidParams(t *testing.T) {
	nan := math.NaN()
	pInf := math.Inf(+1)
	nInf := math.Inf(-1)
	tests := []struct {
		name        string
		alpha, beta float64
	}{
		// Existing negative + both-zero cases (preserved).
		{"alpha_neg", -0.1, 0.5},
		{"alpha_very_neg", -100.0, 0.5},
		{"beta_neg", 0.5, -0.1},
		{"beta_very_neg", 0.5, -100.0},
		{"both_neg", -0.5, -0.5},
		{"both_zero", 0.0, 0.0},
		// Phase 8.5 Q2 — NaN guard cases.
		{"alpha_nan", nan, 0.5},
		{"beta_nan", 0.5, nan},
		{"both_nan", nan, nan},
		// Phase 8.5 Q2 — Inf guard cases.
		{"alpha_pos_inf", pInf, 0.5},
		{"alpha_neg_inf", nInf, 0.5},
		{"beta_pos_inf", 0.5, pInf},
		{"beta_neg_inf", 0.5, nInf},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Byte path.
			func() {
				defer func() {
					r := recover()
					if r == nil {
						t.Errorf("TverskyScore(\"abc\", \"abd\", 2, %g, %g) did not panic", tt.alpha, tt.beta)
						return
					}
					err, ok := r.(error)
					if !ok {
						t.Errorf("panic value type = %T (%v); want error wrapping ErrInvalidTverskyParam", r, r)
						return
					}
					if !errors.Is(err, fuzzymatch.ErrInvalidTverskyParam) {
						t.Errorf("panic err = %v; want errors.Is(_, ErrInvalidTverskyParam)", err)
					}
					if !strings.Contains(err.Error(), "invalid tversky parameter") {
						t.Errorf("panic message %q does not contain \"invalid tversky parameter\"", err.Error())
					}
				}()
				_ = fuzzymatch.TverskyScore("abc", "abd", 2, tt.alpha, tt.beta)
			}()
			// Rune path.
			func() {
				defer func() {
					r := recover()
					if r == nil {
						t.Errorf("TverskyScoreRunes(\"abc\", \"abd\", 2, %g, %g) did not panic", tt.alpha, tt.beta)
						return
					}
					err, ok := r.(error)
					if !ok {
						t.Errorf("panic value type = %T (%v); want error wrapping ErrInvalidTverskyParam", r, r)
						return
					}
					if !errors.Is(err, fuzzymatch.ErrInvalidTverskyParam) {
						t.Errorf("panic err = %v; want errors.Is(_, ErrInvalidTverskyParam)", err)
					}
					if !strings.Contains(err.Error(), "invalid tversky parameter") {
						t.Errorf("panic message %q does not contain \"invalid tversky parameter\"", err.Error())
					}
				}()
				_ = fuzzymatch.TverskyScoreRunes("abc", "abd", 2, tt.alpha, tt.beta)
			}()
		})
	}
}

// TestTverskyScore_DirectCall is the Phase 8.5 Q2 named contract test
// for the data-vs-parameter framework's direct-call panic discipline.
// Asserts that a direct TverskyScore call with α=β=0 panics with a
// typed-error value wrapping ErrInvalidTverskyParam.
//
// This is the canonical failure-mode example the framework documents:
// previously WithTverskyAlgorithm(_, 0, 0, _) constructed successfully
// and the panic only surfaced on the first Scorer.Score call (or on a
// direct TverskyScore call). Plan 04 closes both legs: the option-time
// guard rejects α+β==0 at construction; the direct-call surface
// panics with a discriminable typed error.
func TestTverskyScore_DirectCall(t *testing.T) {
	cases := []struct {
		name        string
		alpha, beta float64
	}{
		{"both_zero", 0.0, 0.0},
		{"alpha_nan", math.NaN(), 1.0},
		{"alpha_pos_inf", math.Inf(+1), 1.0},
		{"beta_neg_inf", 1.0, math.Inf(-1)},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			defer func() {
				r := recover()
				if r == nil {
					t.Fatalf("TverskyScore(\"a\", \"b\", 2, %g, %g) did not panic", c.alpha, c.beta)
				}
				err, ok := r.(error)
				if !ok {
					t.Fatalf("panic value type = %T (%v); want error wrapping ErrInvalidTverskyParam", r, r)
				}
				if !errors.Is(err, fuzzymatch.ErrInvalidTverskyParam) {
					t.Fatalf("expected errors.Is(_, ErrInvalidTverskyParam); got %v", err)
				}
			}()
			_ = fuzzymatch.TverskyScore("a", "b", 2, c.alpha, c.beta)
		})
	}
}

// TestTversky_ZeroAlphaWithPositiveBeta pins the boundary-VALID case:
// α = 0 with β > 0 (and the symmetric β = 0 with α > 0) MUST NOT panic
// — only the joint α==0 AND β==0 case is invalid. On identical inputs
// the identity short-circuit fires and returns 1.0 regardless of α/β.
func TestTversky_ZeroAlphaWithPositiveBeta(t *testing.T) {
	tests := []struct {
		name        string
		alpha, beta float64
	}{
		{"alpha_zero_beta_pos", 0.0, 0.5},
		{"alpha_pos_beta_zero", 0.5, 0.0},
		{"alpha_zero_beta_one", 0.0, 1.0},
		{"alpha_one_beta_zero", 1.0, 0.0},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Identical inputs — should return 1.0 via short-circuit.
			func() {
				defer func() {
					if r := recover(); r != nil {
						t.Errorf("TverskyScore(\"hello\", \"hello\", 2, %g, %g) panicked unexpectedly: %v",
							tt.alpha, tt.beta, r)
					}
				}()
				if got := fuzzymatch.TverskyScore("hello", "hello", 2, tt.alpha, tt.beta); got != 1.0 {
					t.Errorf("TverskyScore(\"hello\", \"hello\", 2, %g, %g) = %g; want 1.0",
						tt.alpha, tt.beta, got)
				}
			}()
			// Distinct inputs — should compute a non-panicking score.
			func() {
				defer func() {
					if r := recover(); r != nil {
						t.Errorf("TverskyScore(\"abcd\", \"abce\", 2, %g, %g) panicked unexpectedly: %v",
							tt.alpha, tt.beta, r)
					}
				}()
				got := fuzzymatch.TverskyScore("abcd", "abce", 2, tt.alpha, tt.beta)
				if math.IsNaN(got) || math.IsInf(got, 0) {
					t.Errorf("TverskyScore(\"abcd\", \"abce\", 2, %g, %g) = %g; want finite", tt.alpha, tt.beta, got)
				}
				if got < 0.0 || got > 1.0 {
					t.Errorf("TverskyScore(\"abcd\", \"abce\", 2, %g, %g) = %g; want in [0, 1]", tt.alpha, tt.beta, got)
				}
			}()
		})
	}
}

// TestTversky_DispatchEqualsJaccard pins the dispatch-table contract:
// the registered AlgoTversky wrapper invokes TverskyScore(a, b, 3,
// 1.0, 1.0) per dispatch_tversky.go, which equals QGramJaccardScore(a,
// b, 3) bit-for-bit (the Jaccard-fallback compromise per CONTEXT.md
// "Claude's Discretion" — see dispatch_tversky.go for the rationale).
//
// This test exercises the dispatch table indirectly via the algorithm
// equivalence; the dispatch-slot-non-nil check lives in algoid_test.go
// (TestDispatch_TverskyRegistered).
func TestTversky_DispatchEqualsJaccard(t *testing.T) {
	tests := []struct {
		a, b string
	}{
		{"abcd", "abcdef"},
		{"AGCT", "AGCTAGCT"},
		{"abc", "xyz"},
		{"hello", "world"},
		{"abcdef", "bcdefg"},
	}
	for _, tt := range tests {
		t.Run(tt.a+"_"+tt.b, func(t *testing.T) {
			tv := fuzzymatch.TverskyScore(tt.a, tt.b, 3, 1.0, 1.0)
			jc := fuzzymatch.QGramJaccardScore(tt.a, tt.b, 3)
			if math.Float64bits(tv) != math.Float64bits(jc) {
				t.Errorf("dispatch[AlgoTversky] equivalence broken for (%q, %q): TverskyScore(...,3,1,1)=%.17g, QGramJaccardScore(...,3)=%.17g",
					tt.a, tt.b, tv, jc)
			}
		})
	}
}

// TestTversky_AllocsBudget asserts the per-call allocation count
// stays within the documented RESEARCH.md §4.1 budget of <= 6 allocs
// for short inputs (two map allocations per side + map-growth ancillary
// allocations). The exact alloc count depends on Go's map
// implementation and is platform-stable; the assertion is a CEILING
// rather than an exact pin, so future Go map-growth tweaks do not
// fail the test.
//
// CONTEXT.md §5 / RESEARCH.md §4.1 budget is "≤ 4 allocations" but
// the realistic table in RESEARCH.md §4.1 acknowledges 4–11 across
// input sizes. Tversky has no sort.Strings (the dot-product loop in
// Cosine is the only q-gram algorithm that sorts), so the alloc count
// is effectively the same as Q-Gram Jaccard / Sørensen-Dice — two
// extractor maps + capacity hints.
func TestTversky_AllocsBudget(t *testing.T) {
	const ceiling = 6.0
	got := testing.AllocsPerRun(100, func() {
		_ = fuzzymatch.TverskyScore("abcd", "abcdef", 2, 0.8, 0.2)
	})
	if got > ceiling {
		t.Errorf("TverskyScore allocs/op = %g; want <= %g (RESEARCH.md §4.1 budget)", got, ceiling)
	}
}
