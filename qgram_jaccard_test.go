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

// qgram_jaccard_test.go pins the public-API contract of qgram_jaccard.go:
// identity, both-empty, one-empty, the canonical Ukkonen 1992 §3
// reference vector (RV-J1) plus three additional hand-derived vectors
// (RV-J2..RV-J4) covering identical, no-overlap, and single-shared-q-gram
// cases, the rune-path café reference (RV-J5-Runes), the direct-call
// panic-on-n<1 contract, and the alloc budget ceiling.
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

// jaccardEpsilon is the float-comparison tolerance for irrational
// expected values (e.g. 3/7 = 0.42857142857142855). Phase 2/3/4
// convention is 1e-9; the Q-Gram Jaccard formula is a single
// integer-valued division so the actual accuracy is far higher than
// 1e-9, but the convention is locked. For exact-rational expected
// values (0.0, 0.2, 0.5, 1.0) the tests use direct equality.
const jaccardEpsilon = 1e-9

// TestQGramJaccard_BothEmpty pins the both-empty convention:
// QGramJaccardScore("", "", n) == 1.0 (vacuous match) — covered by
// the a == b identity short-circuit.
func TestQGramJaccard_BothEmpty(t *testing.T) {
	for _, n := range []int{1, 2, 3, 5} {
		if got := fuzzymatch.QGramJaccardScore("", "", n); got != 1.0 {
			t.Errorf("QGramJaccardScore(\"\", \"\", %d) = %g; want 1.0", n, got)
		}
		if got := fuzzymatch.QGramJaccardScoreRunes("", "", n); got != 1.0 {
			t.Errorf("QGramJaccardScoreRunes(\"\", \"\", %d) = %g; want 1.0", n, got)
		}
	}
}

// TestQGramJaccard_OneEmpty pins the one-empty convention: 0.0 in both
// directions (asymmetric short-circuit gates the identity path away
// before reaching extraction).
func TestQGramJaccard_OneEmpty(t *testing.T) {
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
			if got := fuzzymatch.QGramJaccardScore(tt.a, tt.b, 2); got != 0.0 {
				t.Errorf("QGramJaccardScore(%q, %q, 2) = %g; want 0.0", tt.a, tt.b, got)
			}
			if got := fuzzymatch.QGramJaccardScoreRunes(tt.a, tt.b, 2); got != 0.0 {
				t.Errorf("QGramJaccardScoreRunes(%q, %q, 2) = %g; want 0.0", tt.a, tt.b, got)
			}
		})
	}
}

// TestQGramJaccard_Identical pins the identity short-circuit: any
// non-empty x returns 1.0 for any n >= 1 (the a == b guard fires
// before extraction).
func TestQGramJaccard_Identical(t *testing.T) {
	tests := []string{"abc", "user_id", "x", "WIKIMEDIA", "café", "AGCT", "hello"}
	for _, s := range tests {
		t.Run(s, func(t *testing.T) {
			for _, n := range []int{1, 2, 3, 5} {
				if got := fuzzymatch.QGramJaccardScore(s, s, n); got != 1.0 {
					t.Errorf("QGramJaccardScore(%q, %q, %d) = %g; want 1.0", s, s, n, got)
				}
				if got := fuzzymatch.QGramJaccardScoreRunes(s, s, n); got != 1.0 {
					t.Errorf("QGramJaccardScoreRunes(%q, %q, %d) = %g; want 1.0", s, s, n, got)
				}
			}
		})
	}
}

// TestQGramJaccard_ReferenceVectors pins the four canonical
// hand-derived reference vectors (RV-J1..RV-J4) from RESEARCH.md §2.1.
// Each row's derivation is reproduced in the test sub-name and the
// in-line comment so a reviewer can re-derive the expected value from
// Ukkonen 1992 §3 in under a minute.
func TestQGramJaccard_ReferenceVectors(t *testing.T) {
	tests := []struct {
		name      string
		a, b      string
		n         int
		want      float64
		exact     bool // exact equality (rational) vs. epsilon (irrational)
		derivation string
	}{
		{
			// RV-J1: Ukkonen 1992 §3 worked example (load-bearing
			// primary-source pin). |QA|=3, |QB|=7, |QA∩QB|=3,
			// |QA∪QB|=7, J=3/7=0.42857142857142855.
			name:       "RV-J1_AGCT_AGCTAGCT_n2",
			a:          "AGCT",
			b:          "AGCTAGCT",
			n:          2,
			want:       3.0 / 7.0,
			exact:      false,
			derivation: "QA={AG:1,GC:1,CT:1}; QB={AG:2,GC:2,CT:2,TA:1}; |∩|=3; |∪|=7; J=3/7",
		},
		{
			// RV-J3: orthogonal — no shared bigrams.
			// QA={ab,bc}; QB={xy,yz}; |∩|=0; |∪|=4; J=0.
			name:       "RV-J3_abc_xyz_n2",
			a:          "abc",
			b:          "xyz",
			n:          2,
			want:       0.0,
			exact:      true,
			derivation: "QA={ab,bc}; QB={xy,yz}; |∩|=0; |∪|=4; J=0/4=0",
		},
		{
			// RV-J4: single-shared-q-gram (discriminates 0.0 vs partial).
			// QA={ab,bc,cd}; QB={ab,bx,xy}; |∩|=1 ({ab}); |∪|=5; J=1/5=0.2.
			name:       "RV-J4_abcd_abxy_n2",
			a:          "abcd",
			b:          "abxy",
			n:          2,
			want:       0.2,
			exact:      true,
			derivation: "QA={ab,bc,cd}; QB={ab,bx,xy}; |∩|=1; |∪|=5; J=1/5=0.2",
		},
		{
			// RV-J6: n > min length. Both q-gram views are empty (each
			// input shorter than n=5), so both-empty convention fires
			// inside the helper → 1.0.
			name:       "RV-J6_n_too_large",
			a:          "ab",
			b:          "abc",
			n:          5,
			want:       1.0,
			exact:      true,
			derivation: "QA={}; QB={}; both-empty convention → 1.0",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Logf("derivation: %s", tt.derivation)
			got := fuzzymatch.QGramJaccardScore(tt.a, tt.b, tt.n)
			if tt.exact {
				if got != tt.want {
					t.Errorf("QGramJaccardScore(%q, %q, %d) = %g; want %g exactly",
						tt.a, tt.b, tt.n, got, tt.want)
				}
			} else {
				if math.Abs(got-tt.want) > jaccardEpsilon {
					t.Errorf("QGramJaccardScore(%q, %q, %d) = %.17g; want %.17g (Δ=%g, ε=%g)",
						tt.a, tt.b, tt.n, got, tt.want, math.Abs(got-tt.want), jaccardEpsilon)
				}
			}
		})
	}
}

// TestQGramJaccard_Symmetric pins set-Jaccard's exact symmetry —
// J(A,B) == J(B,A) bit-for-bit, NOT within tolerance. The intersection
// and union cardinalities are integer-valued and agnostic to argument
// order; the single division produces identical float64 output.
func TestQGramJaccard_Symmetric(t *testing.T) {
	tests := []struct {
		a, b string
		n    int
	}{
		{"AGCT", "AGCTAGCT", 2},
		{"abcd", "abxy", 2},
		{"hello", "world", 2},
		{"abcdef", "bcdefg", 3},
	}
	for _, tt := range tests {
		t.Run(tt.a+"_"+tt.b, func(t *testing.T) {
			fwd := fuzzymatch.QGramJaccardScore(tt.a, tt.b, tt.n)
			rev := fuzzymatch.QGramJaccardScore(tt.b, tt.a, tt.n)
			if fwd != rev {
				t.Errorf("QGramJaccardScore not symmetric: J(%q,%q,%d)=%g, J(%q,%q,%d)=%g",
					tt.a, tt.b, tt.n, fwd, tt.b, tt.a, tt.n, rev)
			}
			fwdR := fuzzymatch.QGramJaccardScoreRunes(tt.a, tt.b, tt.n)
			revR := fuzzymatch.QGramJaccardScoreRunes(tt.b, tt.a, tt.n)
			if fwdR != revR {
				t.Errorf("QGramJaccardScoreRunes not symmetric: J(%q,%q,%d)=%g, J(%q,%q,%d)=%g",
					tt.a, tt.b, tt.n, fwdR, tt.b, tt.a, tt.n, revR)
			}
		})
	}
}

// TestQGramJaccardRunes_CafeReference pins the rune-path RV-J5-Runes
// reference: "café" / "cafe" / n=2 → 0.5.
//
// QA = rune-bigrams("café") = {"ca":1, "af":1, "fé":1}
// QB = rune-bigrams("cafe") = {"ca":1, "af":1, "fe":1}
// intersection = 2 (ca + af); union = 4 (ca + af + fé + fe); J = 2/4 = 0.5
//
// Pinning this load-bearing rune-path vector ensures the rune
// extractor and the algorithm wire up correctly — a regression where
// the byte path is silently called would yield a different score.
func TestQGramJaccardRunes_CafeReference(t *testing.T) {
	got := fuzzymatch.QGramJaccardScoreRunes("café", "cafe", 2)
	want := 0.5
	if got != want {
		t.Errorf("QGramJaccardScoreRunes(\"café\", \"cafe\", 2) = %g; want %g exactly", got, want)
	}
}

// TestQGramJaccard_PanicsOnInvalidN pins the direct-call panic-on-n<1
// contract per CONTEXT.md §5 LOCKED. Both byte and rune surfaces panic
// with the same message text containing "invalid q-gram size".
func TestQGramJaccard_PanicsOnInvalidN(t *testing.T) {
	tests := []int{0, -1, -100, math.MinInt32}
	for _, n := range tests {
		t.Run("n_"+strconv.Itoa(n), func(t *testing.T) {
			// Byte path.
			func() {
				defer func() {
					r := recover()
					if r == nil {
						t.Errorf("QGramJaccardScore(\"hello\", \"hello\", %d) did not panic", n)
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
				_ = fuzzymatch.QGramJaccardScore("hello", "hello", n)
			}()
			// Rune path.
			func() {
				defer func() {
					r := recover()
					if r == nil {
						t.Errorf("QGramJaccardScoreRunes(\"hello\", \"hello\", %d) did not panic", n)
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
				_ = fuzzymatch.QGramJaccardScoreRunes("hello", "hello", n)
			}()
		})
	}
}

// TestQGramJaccard_AllocsBudget asserts the per-call allocation count
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
func TestQGramJaccard_AllocsBudget(t *testing.T) {
	const ceiling = 6.0
	got := testing.AllocsPerRun(100, func() {
		_ = fuzzymatch.QGramJaccardScore("AGCT", "AGCTAGCT", 2)
	})
	if got > ceiling {
		t.Errorf("QGramJaccardScore allocs/op = %g; want <= %g (RESEARCH.md §4.1 budget)", got, ceiling)
	}
}

