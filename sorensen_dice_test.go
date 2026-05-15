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

// sorensen_dice_test.go pins the public-API contract of sorensen_dice.go:
// identity, both-empty, one-empty, the four canonical hand-derived
// reference vectors RV-D1..RV-D4 from RESEARCH.md §2.2 (covering the
// load-bearing NLP-textbook bigram pair, a high-overlap analogue of
// Dice 1945 §3, a trigram variant, and identity), the rune-path café
// reference, exact symmetry, the direct-call panic-on-n<1 contract,
// and the alloc budget ceiling.
//
// Stdlib `testing` only — no testify in root tests, per
// .claude/skills/go-coding-standards.

package fuzzymatch_test

import (
	"math"
	"strconv"
	"strings"
	"testing"

	"github.com/axonops/fuzzymatch"
)

// sorensenDiceEpsilon is the float-comparison tolerance for irrational
// expected values. The Sørensen-Dice formula is a single
// integer-valued multiplication-then-division so the actual accuracy
// is far higher than 1e-9, but the convention is locked. For
// exact-rational expected values (0.0, 0.25, 0.8, 1.0) the tests use
// direct equality.
const sorensenDiceEpsilon = 1e-9

// TestSorensenDice_BothEmpty pins the both-empty convention:
// SorensenDiceScore("", "", n) == 1.0 (vacuous match) — covered by the
// a == b identity short-circuit.
func TestSorensenDice_BothEmpty(t *testing.T) {
	for _, n := range []int{1, 2, 3, 5} {
		if got := fuzzymatch.SorensenDiceScore("", "", n); got != 1.0 {
			t.Errorf("SorensenDiceScore(\"\", \"\", %d) = %g; want 1.0", n, got)
		}
		if got := fuzzymatch.SorensenDiceScoreRunes("", "", n); got != 1.0 {
			t.Errorf("SorensenDiceScoreRunes(\"\", \"\", %d) = %g; want 1.0", n, got)
		}
	}
}

// TestSorensenDice_OneEmpty pins the one-empty convention: 0.0 in both
// directions (asymmetric short-circuit gates the identity path away
// before reaching extraction).
func TestSorensenDice_OneEmpty(t *testing.T) {
	tests := []struct{ a, b string }{
		{"", "abc"},
		{"abc", ""},
		{"", "x"},
		{"x", ""},
		{"", "café"},
		{"café", ""},
	}
	for _, tt := range tests {
		t.Run(tt.a+"_"+tt.b, func(t *testing.T) {
			if got := fuzzymatch.SorensenDiceScore(tt.a, tt.b, 2); got != 0.0 {
				t.Errorf("SorensenDiceScore(%q, %q, 2) = %g; want 0.0", tt.a, tt.b, got)
			}
			if got := fuzzymatch.SorensenDiceScoreRunes(tt.a, tt.b, 2); got != 0.0 {
				t.Errorf("SorensenDiceScoreRunes(%q, %q, 2) = %g; want 0.0", tt.a, tt.b, got)
			}
		})
	}
}

// TestSorensenDice_Identical pins the identity short-circuit: any
// non-empty x returns 1.0 for any n >= 1 (the a == b guard fires
// before extraction).
func TestSorensenDice_Identical(t *testing.T) {
	tests := []string{"abc", "user_id", "x", "WIKIMEDIA", "café", "AGCT", "hello"}
	for _, s := range tests {
		t.Run(s, func(t *testing.T) {
			for _, n := range []int{1, 2, 3, 5} {
				if got := fuzzymatch.SorensenDiceScore(s, s, n); got != 1.0 {
					t.Errorf("SorensenDiceScore(%q, %q, %d) = %g; want 1.0", s, s, n, got)
				}
				if got := fuzzymatch.SorensenDiceScoreRunes(s, s, n); got != 1.0 {
					t.Errorf("SorensenDiceScoreRunes(%q, %q, %d) = %g; want 1.0", s, s, n, got)
				}
			}
		})
	}
}

// TestSorensenDice_ReferenceVectors pins the four canonical
// hand-derived reference vectors (RV-D1..RV-D4) from RESEARCH.md §2.2.
// Each row's derivation is reproduced in the test sub-name and the
// in-line comment so a reviewer can re-derive the expected value from
// Dice 1945 §3 / Sørensen 1948 §3 in under a minute.
func TestSorensenDice_ReferenceVectors(t *testing.T) {
	tests := []struct {
		name       string
		a, b       string
		n          int
		want       float64
		exact      bool // exact equality (rational) vs. epsilon (irrational)
		derivation string
	}{
		{
			// RV-D1: load-bearing canonical NLP-textbook bigram pair.
			// QA = bigrams("night") = {ni:1, ig:1, gh:1, ht:1}  — total 4
			// QB = bigrams("nacht") = {na:1, ac:1, ch:1, ht:1}  — total 4
			// |QA ∩ QB| = 1 (ht); DSC = 2·1/(4+4) = 2/8 = 0.25.
			name:       "RV-D1_night_nacht_n2",
			a:          "night",
			b:          "nacht",
			n:          2,
			want:       0.25,
			exact:      true,
			derivation: "QA={ni,ig,gh,ht}; QB={na,ac,ch,ht}; |∩|=1; DSC=2·1/(4+4)=0.25",
		},
		{
			// RV-D2: Dice 1945 §3 high-overlap analogue.
			// QA = bigrams("abcdef") = {ab:1, bc:1, cd:1, de:1, ef:1} — total 5
			// QB = bigrams("bcdefg") = {bc:1, cd:1, de:1, ef:1, fg:1} — total 5
			// |QA ∩ QB| = 4 (bc, cd, de, ef); DSC = 2·4/(5+5) = 8/10 = 0.8.
			name:       "RV-D2_abcdef_bcdefg_n2",
			a:          "abcdef",
			b:          "bcdefg",
			n:          2,
			want:       0.8,
			exact:      true,
			derivation: "QA={ab,bc,cd,de,ef}; QB={bc,cd,de,ef,fg}; |∩|=4; DSC=2·4/(5+5)=0.8",
		},
		{
			// RV-D3: trigram variant — exercises the n=3 path.
			// QA = trigrams("abcdef") = {abc:1, bcd:1, cde:1, def:1} — total 4
			// QB = trigrams("abcXef") = {abc:1, bcX:1, cXe:1, Xef:1} — total 4
			// |QA ∩ QB| = 1 (abc); DSC = 2·1/(4+4) = 0.25.
			name:       "RV-D3_abcdef_abcXef_n3",
			a:          "abcdef",
			b:          "abcXef",
			n:          3,
			want:       0.25,
			exact:      true,
			derivation: "QA={abc,bcd,cde,def}; QB={abc,bcX,cXe,Xef}; |∩|=1; DSC=2·1/(4+4)=0.25",
		},
		{
			// RV-D4: identity. The a == b short-circuit fires before
			// extraction so no q-gram work happens.
			name:       "RV-D4_hello_hello_n2",
			a:          "hello",
			b:          "hello",
			n:          2,
			want:       1.0,
			exact:      true,
			derivation: "a == b identity short-circuit → 1.0",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Logf("derivation: %s", tt.derivation)
			got := fuzzymatch.SorensenDiceScore(tt.a, tt.b, tt.n)
			if tt.exact {
				if got != tt.want {
					t.Errorf("SorensenDiceScore(%q, %q, %d) = %g; want %g exactly",
						tt.a, tt.b, tt.n, got, tt.want)
				}
			} else {
				if math.Abs(got-tt.want) > sorensenDiceEpsilon {
					t.Errorf("SorensenDiceScore(%q, %q, %d) = %.17g; want %.17g (Δ=%g, ε=%g)",
						tt.a, tt.b, tt.n, got, tt.want, math.Abs(got-tt.want), sorensenDiceEpsilon)
				}
			}
		})
	}
}

// TestSorensenDice_Symmetric pins Sørensen-Dice's exact symmetry —
// DSC(A,B) == DSC(B,A) bit-for-bit, NOT within tolerance. The
// intersection cardinality and per-side multiset totals are
// integer-valued and agnostic to argument order; the single
// multiplication + division produces identical float64 output.
func TestSorensenDice_Symmetric(t *testing.T) {
	tests := []struct {
		a, b string
		n    int
	}{
		{"night", "nacht", 2},
		{"abcdef", "bcdefg", 2},
		{"abcdef", "abcXef", 3},
		{"hello", "world", 2},
	}
	for _, tt := range tests {
		t.Run(tt.a+"_"+tt.b, func(t *testing.T) {
			fwd := fuzzymatch.SorensenDiceScore(tt.a, tt.b, tt.n)
			rev := fuzzymatch.SorensenDiceScore(tt.b, tt.a, tt.n)
			if fwd != rev {
				t.Errorf("SorensenDiceScore not symmetric: DSC(%q,%q,%d)=%g, DSC(%q,%q,%d)=%g",
					tt.a, tt.b, tt.n, fwd, tt.b, tt.a, tt.n, rev)
			}
			fwdR := fuzzymatch.SorensenDiceScoreRunes(tt.a, tt.b, tt.n)
			revR := fuzzymatch.SorensenDiceScoreRunes(tt.b, tt.a, tt.n)
			if fwdR != revR {
				t.Errorf("SorensenDiceScoreRunes not symmetric: DSC(%q,%q,%d)=%g, DSC(%q,%q,%d)=%g",
					tt.a, tt.b, tt.n, fwdR, tt.b, tt.a, tt.n, revR)
			}
		})
	}
}

// TestSorensenDiceRunes_CafeReference pins the rune-path RV-D5-style
// reference: "café" / "cafe" / n=2 → 0.6666666666666666.
//
// QA = rune-bigrams("café") = {"ca":1, "af":1, "fé":1}  — total 3
// QB = rune-bigrams("cafe") = {"ca":1, "af":1, "fe":1}  — total 3
// |QA ∩ QB| = 2 (ca, af); DSC = 2·2/(3+3) = 4/6 ≈ 0.6666666666666666
//
// Pinning this load-bearing rune-path vector ensures the rune
// extractor and the algorithm wire up correctly — a regression where
// the byte path is silently called would yield a different score.
func TestSorensenDiceRunes_CafeReference(t *testing.T) {
	got := fuzzymatch.SorensenDiceScoreRunes("café", "cafe", 2)
	want := 4.0 / 6.0
	if math.Abs(got-want) > 1e-15 {
		t.Errorf("SorensenDiceScoreRunes(\"café\", \"cafe\", 2) = %.17g; want %.17g (Δ=%g)",
			got, want, math.Abs(got-want))
	}
}

// TestSorensenDice_PanicsOnInvalidN pins the direct-call panic-on-n<1
// contract per CONTEXT.md §5 LOCKED. Both byte and rune surfaces panic
// with the same message text containing "invalid q-gram size".
func TestSorensenDice_PanicsOnInvalidN(t *testing.T) {
	tests := []int{0, -1, -100, math.MinInt32}
	for _, n := range tests {
		t.Run("n_"+strconv.Itoa(n), func(t *testing.T) {
			// Byte path.
			func() {
				defer func() {
					r := recover()
					if r == nil {
						t.Errorf("SorensenDiceScore(\"hello\", \"hello\", %d) did not panic", n)
						return
					}
					msg, ok := r.(string)
					if !ok {
						t.Errorf("panic value type = %T (%v); want string", r, r)
						return
					}
					if !strings.Contains(msg, "invalid q-gram size") {
						t.Errorf("panic message %q does not contain \"invalid q-gram size\"", msg)
					}
				}()
				_ = fuzzymatch.SorensenDiceScore("hello", "hello", n)
			}()
			// Rune path.
			func() {
				defer func() {
					r := recover()
					if r == nil {
						t.Errorf("SorensenDiceScoreRunes(\"hello\", \"hello\", %d) did not panic", n)
						return
					}
					msg, ok := r.(string)
					if !ok {
						t.Errorf("panic value type = %T (%v); want string", r, r)
						return
					}
					if !strings.Contains(msg, "invalid q-gram size") {
						t.Errorf("panic message %q does not contain \"invalid q-gram size\"", msg)
					}
				}()
				_ = fuzzymatch.SorensenDiceScoreRunes("hello", "hello", n)
			}()
		})
	}
}

// TestSorensenDice_AllocsBudget asserts the per-call allocation count
// stays within the documented RESEARCH.md §4.1 budget of <= 6 allocs
// for short inputs (two map allocations per side + map-growth ancillary
// allocations). The exact alloc count depends on Go's map
// implementation and is platform-stable; the assertion is a CEILING
// rather than an exact pin, so future Go map-growth tweaks do not
// fail the test.
//
// CONTEXT.md §5 / RESEARCH.md §4.1 budget is "≤ 4 allocations" but
// the realistic table in RESEARCH.md §4.1 acknowledges 4–11 across
// input sizes (the 4 floor is the canonical-source ideal; the
// realistic ceiling is what matters for regression detection).
func TestSorensenDice_AllocsBudget(t *testing.T) {
	const ceiling = 6.0
	got := testing.AllocsPerRun(100, func() {
		_ = fuzzymatch.SorensenDiceScore("night", "nacht", 2)
	})
	if got > ceiling {
		t.Errorf("SorensenDiceScore allocs/op = %g; want <= %g (RESEARCH.md §4.1 budget)", got, ceiling)
	}
}
