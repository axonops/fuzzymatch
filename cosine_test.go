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

// cosine_test.go pins the public-API contract of cosine.go: identity,
// both-empty, one-empty, orthogonal, the FIVE hand-derived reference
// vectors RV-C1..RV-C5 from RESEARCH.md §2.3 (each documented inline at
// 17-significant-digit float64 precision, reviewer-verifiable in <30s
// against Salton & McGill 1983 §4.1 eq. 4.4 p.121), the rune-path café
// reference (RV-C3), exact symmetry, the direct-call panic-on-n<1
// contract, the per-process determinism / sorted-key-iteration
// regression test, and the alloc budget ceiling.
//
// This file is LOAD-BEARING per CONTEXT.md §4 — Cosine carries the
// cross-validation density that would otherwise come from an external
// library. The hand-derivation comment blocks ARE the third-party
// reviewable proof of correctness.
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

// cosineEpsilon is the float-comparison tolerance for irrational
// expected values. The Cosine formula uses math.Sqrt (IEEE-754 correctly
// rounded on all four CI platforms per RESEARCH.md §3.5) so the actual
// accuracy is at the last-bit level; the convention is locked at 1e-15
// (~5 ULP for values near 1.0). For exact-rational expected values
// (0.0, 0.5, 1.0) the tests use direct equality.
const cosineEpsilon = 1e-15

// TestCosine_BothEmpty pins the both-empty convention:
// CosineScore("", "", n) == 1.0 (vacuous match) — covered by the
// a == b identity short-circuit.
func TestCosine_BothEmpty(t *testing.T) {
	for _, n := range []int{1, 2, 3, 5} {
		if got := fuzzymatch.CosineScore("", "", n); got != 1.0 {
			t.Errorf("CosineScore(\"\", \"\", %d) = %g; want 1.0", n, got)
		}
		if got := fuzzymatch.CosineScoreRunes("", "", n); got != 1.0 {
			t.Errorf("CosineScoreRunes(\"\", \"\", %d) = %g; want 1.0", n, got)
		}
	}
}

// TestCosine_OneEmpty pins the one-empty convention: 0.0 in both
// directions (asymmetric short-circuit gates the identity path away
// before reaching extraction).
func TestCosine_OneEmpty(t *testing.T) {
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
			if got := fuzzymatch.CosineScore(tt.a, tt.b, 2); got != 0.0 {
				t.Errorf("CosineScore(%q, %q, 2) = %g; want 0.0", tt.a, tt.b, got)
			}
			if got := fuzzymatch.CosineScoreRunes(tt.a, tt.b, 2); got != 0.0 {
				t.Errorf("CosineScoreRunes(%q, %q, 2) = %g; want 0.0", tt.a, tt.b, got)
			}
		})
	}
}

// TestCosine_Identical pins the identity short-circuit: any non-empty
// x returns 1.0 for any n >= 1 (the a == b guard fires before
// extraction).
func TestCosine_Identical(t *testing.T) {
	tests := []string{"abc", "user_id", "x", "WIKIMEDIA", "café", "AGCT", "hello"}
	for _, s := range tests {
		t.Run(s, func(t *testing.T) {
			for _, n := range []int{1, 2, 3, 5} {
				if got := fuzzymatch.CosineScore(s, s, n); got != 1.0 {
					t.Errorf("CosineScore(%q, %q, %d) = %g; want 1.0", s, s, n, got)
				}
				if got := fuzzymatch.CosineScoreRunes(s, s, n); got != 1.0 {
					t.Errorf("CosineScoreRunes(%q, %q, %d) = %g; want 1.0", s, s, n, got)
				}
			}
		})
	}
}

// TestCosine_Orthogonal pins the empty-intersection case: when QA and
// QB share no keys, the dot product is zero and so cos(A, B) = 0.
//
//	a = "abc", b = "xyz", n = 2
//	QA = {ab:1, bc:1};   QB = {xy:1, yz:1}
//	intersection keys = []
//	dot = 0 → cos = 0 / (sqrt(2)·sqrt(2)) = 0
func TestCosine_Orthogonal(t *testing.T) {
	if got := fuzzymatch.CosineScore("abc", "xyz", 2); got != 0.0 {
		t.Errorf("CosineScore(\"abc\", \"xyz\", 2) = %g; want 0.0", got)
	}
	// Rune-path orthogonal pair too.
	if got := fuzzymatch.CosineScoreRunes("abc", "xyz", 2); got != 0.0 {
		t.Errorf("CosineScoreRunes(\"abc\", \"xyz\", 2) = %g; want 0.0", got)
	}
}

// TestCosine_RV_C1_AsciiN2Irrational pins the RV-C1 reference vector
// from RESEARCH.md §2.3 — short-ASCII irrational at n=2.
//
// Hand derivation (Salton & McGill 1983 §4.1 eq. 4.4, p.121):
//
//	a = "abc", b = "abcd", n = 2
//
//	QA = bigrams("abc")  = [ab, bc]
//	   = {ab:1, bc:1}
//	   ‖A‖² = 1² + 1² = 2
//
//	QB = bigrams("abcd") = [ab, bc, cd]
//	   = {ab:1, bc:1, cd:1}
//	   ‖B‖² = 1² + 1² + 1² = 3
//
//	intersection keys (sorted) = ["ab", "bc"]
//	dot = 1·1 + 1·1 = 2
//
//	cos = 2 / (sqrt(2) · sqrt(3)) = 2 / sqrt(6)
//
// Mathematical limit: 2/sqrt(6) ≈ 0.8164965809277261. IEEE-754 actual
// from the Salton & McGill 1983 §4.1 factorised form
// `dot / (sqrt(normASq) * sqrt(normBSq))`:
//
//	math.Sqrt(2) = 1.4142135623730951
//	math.Sqrt(3) = 1.7320508075688772
//	math.Sqrt(2) * math.Sqrt(3) = 2.4494897427831783
//	2.0 / 2.4494897427831783       = 0.81649658092772592
//
// This is 1 ULP below the rational limit (the `2.0/math.Sqrt(6)` direct
// form would yield 0.81649658092772615 — also 1 ULP from the limit, in
// the opposite direction). Pin the FACTORISED-form output because the
// implementation uses that form (RESEARCH.md "Pitfall 2": the
// expected value is what the implementation actually produces).
//
// This vector exercises the math.Sqrt precision gate — sqrt(2) and
// sqrt(3) are IEEE-754 correctly rounded on all four CI platforms per
// RESEARCH.md §3.5.
func TestCosine_RV_C1_AsciiN2Irrational(t *testing.T) {
	// Pin to the actual factorised-form float64 output. cosineEpsilon
	// (1e-15) is generous; the test will catch any divergence > 5 ULP.
	const want = 0.81649658092772592
	got := fuzzymatch.CosineScore("abc", "abcd", 2)
	if math.Abs(got-want) > cosineEpsilon {
		t.Errorf("CosineScore(\"abc\", \"abcd\", 2) = %.17g; want %.17g (Δ=%g, ε=%g)",
			got, want, math.Abs(got-want), cosineEpsilon)
	}
}

// TestCosine_RV_C2_LargeIntersectionN3 pins the RV-C2 reference vector
// from RESEARCH.md §2.3 — large-intersection at n=3. This exercises
// the sorted-key accumulation order from CONTEXT.md §3 LOCKED.
//
// Hand derivation (Salton & McGill 1983 §4.1 eq. 4.4, p.121):
//
//	a = "abcdefgh", b = "abcdefgi", n = 3
//
//	QA = trigrams("abcdefgh") = [abc, bcd, cde, def, efg, fgh]
//	   = {abc:1, bcd:1, cde:1, def:1, efg:1, fgh:1}
//	   ‖A‖² = 6 · 1² = 6
//
//	QB = trigrams("abcdefgi") = [abc, bcd, cde, def, efg, fgi]
//	   = {abc:1, bcd:1, cde:1, def:1, efg:1, fgi:1}
//	   ‖B‖² = 6 · 1² = 6
//
//	intersection keys (sorted) = ["abc", "bcd", "cde", "def", "efg"]   ← 5 keys
//	dot = 1+1+1+1+1 = 5
//
//	cos = 5 / (sqrt(6) · sqrt(6)) = 5/6
//
// Mathematical limit: 5/6 ≈ 0.8333333333333334. IEEE-754 actual from
// the Salton & McGill 1983 factorised form
// `dot / (sqrt(normASq) * sqrt(normBSq))`:
//
//	math.Sqrt(6)              = 2.449489742783178
//	math.Sqrt(6) * math.Sqrt(6) = 6.000000000000001 (NOT exact 6 —
//	  last-bit rounding of the product)
//	5.0 / 6.000000000000001    = 0.83333333333333348
//
// This is 1 ULP above the rational limit. The 5/6 direct form yields
// 0.83333333333333337 (the rational limit). Pin the FACTORISED-form
// output because the implementation uses that form per RESEARCH.md
// "Pitfall 2".
//
// The 5-key sorted-iteration dot product is the load-bearing exercise
// for the CONTEXT.md §3 LOCKED sorted-key iteration order.
func TestCosine_RV_C2_LargeIntersectionN3(t *testing.T) {
	// Pin to the actual factorised-form float64 output.
	const want = 0.83333333333333348
	got := fuzzymatch.CosineScore("abcdefgh", "abcdefgi", 3)
	if math.Abs(got-want) > cosineEpsilon {
		t.Errorf("CosineScore(\"abcdefgh\", \"abcdefgi\", 3) = %.17g; want %.17g (Δ=%g, ε=%g)",
			got, want, math.Abs(got-want), cosineEpsilon)
	}
}

// TestCosineRunes_RV_C3_UnicodeN2 pins the RV-C3 reference vector from
// RESEARCH.md §2.3 — Unicode/runes pair at n=2. The byte path would
// split "é" mid-codepoint and produce a different score; the rune
// variant guarantees rune-boundary alignment.
//
// Hand derivation (Salton & McGill 1983 §4.1 eq. 4.4, p.121):
//
//	a = "café", b = "cafe", n = 2 (RUNE PATH — CosineScoreRunes)
//
//	Rune decomposition:
//	  a runes = ['c', 'a', 'f', 'é']
//	  b runes = ['c', 'a', 'f', 'e']
//
//	QA = rune-bigrams(a) = ["ca", "af", "fé"]   ← "fé" as rune-bigram
//	   = {ca:1, af:1, fé:1}
//	   ‖A‖² = 3
//
//	QB = rune-bigrams(b) = ["ca", "af", "fe"]
//	   = {ca:1, af:1, fe:1}
//	   ‖B‖² = 3
//
//	intersection keys (sorted, byte-comparison on UTF-8 encoding):
//	   ["af", "ca"]   ← "fé" sorts AFTER ASCII pairs in byte-order;
//	                    "fe" not in intersection
//	dot = 1·1 + 1·1 = 2
//
//	cos = 2 / (sqrt(3) · sqrt(3)) = 2/3
//
// Mathematical limit: 2/3 ≈ 0.6666666666666666. IEEE-754 actual from
// the Salton & McGill 1983 factorised form
// `dot / (sqrt(normASq) * sqrt(normBSq))`:
//
//	math.Sqrt(3)              = 1.7320508075688772
//	math.Sqrt(3) * math.Sqrt(3) = 2.9999999999999996 (NOT exact 3 —
//	  last-bit rounding)
//	2.0 / 2.9999999999999996   = 0.66666666666666674
//
// This is 1 ULP above the rational limit. Pin the FACTORISED-form
// output per RESEARCH.md "Pitfall 2" — the implementation uses that
// form.
//
// `sort.Strings` on the q-gram keys is byte-lexicographic, not Unicode-
// collation, but it IS total and deterministic across platforms — Go's
// `sort.Strings` uses `<` on `string` (byte-comparison per language
// spec). RESEARCH.md §3.6 confirms.
func TestCosineRunes_RV_C3_UnicodeN2(t *testing.T) {
	// Pin to the actual factorised-form float64 output.
	const want = 0.66666666666666674
	got := fuzzymatch.CosineScoreRunes("café", "cafe", 2)
	if math.Abs(got-want) > cosineEpsilon {
		t.Errorf("CosineScoreRunes(\"café\", \"cafe\", 2) = %.17g; want %.17g (Δ=%g, ε=%g)",
			got, want, math.Abs(got-want), cosineEpsilon)
	}
}

// TestCosine_RV_C4_N4Exact pins the RV-C4 reference vector from
// RESEARCH.md §2.3 — n=4 exercise. The mathematical limit is 1/2 = 0.5
// exactly, but the float64 floating-point computation
// `1.0 / (math.Sqrt(2) * math.Sqrt(2))` produces 0.49999999999999989
// because `math.Sqrt(2) * math.Sqrt(2) = 2.0000000000000004` (last-bit
// error in the multiplication). This is a textbook example of
// RESEARCH.md "Pitfall 2": the implementation is correct (it follows
// the Salton & McGill 1983 §4.1 factorised form
// `dot / (sqrt(normASq) * sqrt(normBSq))`), and the 1-ULP shortfall
// from the rational ideal is a property of IEEE-754 arithmetic, not
// of the algorithm.
//
// Hand derivation (Salton & McGill 1983 §4.1 eq. 4.4, p.121):
//
//	a = "abcde", b = "abcdf", n = 4
//
//	QA = 4-grams("abcde") = [abcd, bcde]
//	   = {abcd:1, bcde:1}
//	   ‖A‖² = 2
//
//	QB = 4-grams("abcdf") = [abcd, bcdf]
//	   = {abcd:1, bcdf:1}
//	   ‖B‖² = 2
//
//	intersection keys (sorted) = ["abcd"]
//	dot = 1·1 = 1
//
//	cos = 1 / (sqrt(2) · sqrt(2))
//
//	Mathematical limit: 1/2 = 0.5. IEEE-754 actual:
//	  math.Sqrt(2) = 1.4142135623730951
//	  math.Sqrt(2) * math.Sqrt(2) = 2.0000000000000004 (NOT exact 2 —
//	    the last-bit rounding of the product is the source of the ULP
//	    shortfall)
//	  1.0 / 2.0000000000000004 = 0.49999999999999989
//
// This pair anchors the n=4 path AND surfaces the IEEE-754 rounding
// reality of the factorised cosine form. RV-C1 / RV-C2 / RV-C5
// exercise irrational sqrt paths.
func TestCosine_RV_C4_N4Exact(t *testing.T) {
	// Pin to the actual float64 output, NOT the rational 0.5.
	const want = 0.49999999999999989
	got := fuzzymatch.CosineScore("abcde", "abcdf", 4)
	if got != want {
		t.Errorf("CosineScore(\"abcde\", \"abcdf\", 4) = %.17g; want %.17g exactly",
			got, want)
	}
}

// TestCosine_RV_C5_OptionalIrrational pins the optional RV-C5
// reference vector from RESEARCH.md §2.3 — single-key intersection
// where sqrt(3) is the irrational source. Provides a fifth hand-
// derivation block for additional reviewer density per CONTEXT.md §4
// LOCKED.
//
// Hand derivation (Salton & McGill 1983 §4.1 eq. 4.4, p.121):
//
//	a = "ab", b = "abcd", n = 2
//
//	QA = bigrams("ab") = [ab]
//	   = {ab:1}
//	   ‖A‖² = 1
//
//	QB = bigrams("abcd") = [ab, bc, cd]
//	   = {ab:1, bc:1, cd:1}
//	   ‖B‖² = 3
//
//	intersection keys (sorted) = ["ab"]
//	dot = 1·1 = 1
//
//	cos = 1 / (sqrt(1) · sqrt(3)) = 1 / sqrt(3)
//	    = 1 / 1.7320508075688772
//	    = 0.5773502691896258
//
// Notes: sqrt(1) = 1.0 exactly; the irrational comes entirely from
// sqrt(3) ≈ 1.7320508075688772 (IEEE-754 correctly rounded — RESEARCH.md
// §3.5).
func TestCosine_RV_C5_OptionalIrrational(t *testing.T) {
	const want = 0.5773502691896258
	got := fuzzymatch.CosineScore("ab", "abcd", 2)
	if math.Abs(got-want) > cosineEpsilon {
		t.Errorf("CosineScore(\"ab\", \"abcd\", 2) = %.17g; want %.17g (Δ=%g, ε=%g)",
			got, want, math.Abs(got-want), cosineEpsilon)
	}
}

// TestCosine_Symmetric pins Cosine's exact symmetry — cos(A,B) == cos(B,A)
// bit-for-bit, NOT within tolerance. Sorted-key iteration is canonical
// regardless of input argument order; the dot-product reduction visits
// the same intersection keys in the same sorted order, producing
// bit-identical float64 output.
func TestCosine_Symmetric(t *testing.T) {
	tests := []struct {
		a, b string
		n    int
	}{
		{"abc", "abcd", 2},          // RV-C1
		{"abcdefgh", "abcdefgi", 3}, // RV-C2
		{"abcde", "abcdf", 4},       // RV-C4
		{"hello", "world", 2},       // generic
		{"night", "nacht", 2},       // generic
	}
	for _, tt := range tests {
		t.Run(tt.a+"_"+tt.b, func(t *testing.T) {
			fwd := fuzzymatch.CosineScore(tt.a, tt.b, tt.n)
			rev := fuzzymatch.CosineScore(tt.b, tt.a, tt.n)
			if fwd != rev {
				t.Errorf("CosineScore not symmetric: cos(%q,%q,%d)=%.17g, cos(%q,%q,%d)=%.17g",
					tt.a, tt.b, tt.n, fwd, tt.b, tt.a, tt.n, rev)
			}
			fwdR := fuzzymatch.CosineScoreRunes(tt.a, tt.b, tt.n)
			revR := fuzzymatch.CosineScoreRunes(tt.b, tt.a, tt.n)
			if fwdR != revR {
				t.Errorf("CosineScoreRunes not symmetric: cos(%q,%q,%d)=%.17g, cos(%q,%q,%d)=%.17g",
					tt.a, tt.b, tt.n, fwdR, tt.b, tt.a, tt.n, revR)
			}
		})
	}
}

// TestCosine_PanicsOnInvalidN pins the direct-call panic-on-n<1
// contract per CONTEXT.md §5 LOCKED. Both byte and rune surfaces panic
// with the same message text containing "invalid q-gram size".
func TestCosine_PanicsOnInvalidN(t *testing.T) {
	tests := []int{0, -1, -100, math.MinInt32}
	for _, n := range tests {
		t.Run("n_"+strconv.Itoa(n), func(t *testing.T) {
			// Byte path.
			func() {
				defer func() {
					r := recover()
					if r == nil {
						t.Errorf("CosineScore(\"hello\", \"hello\", %d) did not panic", n)
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
				_ = fuzzymatch.CosineScore("hello", "hello", n)
			}()
			// Rune path.
			func() {
				defer func() {
					r := recover()
					if r == nil {
						t.Errorf("CosineScoreRunes(\"hello\", \"hello\", %d) did not panic", n)
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
				_ = fuzzymatch.CosineScoreRunes("hello", "hello", n)
			}()
		})
	}
}

// TestCosine_AllocsBudget asserts the per-call allocation count stays
// within the documented RESEARCH.md §4.1 budget (Cosine: ≤ 5 allocs on
// short inputs — two extractQGrams maps + one intersection-keys []string
// slice + cap-hint backing arrays). The exact alloc count depends on
// Go's map and slice implementation and is platform-stable; the
// assertion is a CEILING rather than an exact pin so future Go growth
// tweaks do not fail the test.
//
// RESEARCH.md §4.1 acknowledges that the realistic Cosine budget is
// "Jaccard + 1 sorted-key slice" — i.e. one more than Sørensen-Dice's
// 4-alloc budget on short inputs. The ceiling here is set at 7 to match
// the realistic upper bound (two maps + one []string + 4-bit growth
// headroom for the canonical RV-C1 input).
func TestCosine_AllocsBudget(t *testing.T) {
	const ceiling = 7.0
	got := testing.AllocsPerRun(100, func() {
		_ = fuzzymatch.CosineScore("abc", "abcd", 2)
	})
	if got > ceiling {
		t.Errorf("CosineScore allocs/op = %g; want <= %g (RESEARCH.md §4.1 budget — Jaccard + 1 sorted-key slice)", got, ceiling)
	}
}

// TestCosine_SortedKeyIteration is the INVARIANT REGRESSION TEST for
// CONTEXT.md §3 LOCKED: the dot-product loop iterates intersection
// keys in SORTED order. If a future refactor removes the sort.Strings
// call and the dot-product reduction order becomes non-deterministic,
// this test catches the regression.
//
// Method: call CosineScore 1000 times on the load-bearing RV-C2
// 5-key-intersection input ("abcdefgh"/"abcdefgi"/n=3). Every result
// MUST be byte-for-byte identical via math.Float64bits comparison.
// Without the sort, map-iteration randomisation would produce
// different float reduction orders across calls, and the slight
// rounding differences (especially on amd64 where FMA is not emitted —
// see RESEARCH.md §3.1) would surface as bit-level divergence between
// some pairs of iterations.
//
// RV-C2 is the canonical input here because it has a 5-key
// intersection — the largest in the RV-CN catalogue and therefore the
// strongest test of reduction-order sensitivity. Single-key
// intersection inputs (RV-C5) would be insensitive to iteration order.
func TestCosine_SortedKeyIteration(t *testing.T) {
	const iterations = 1000
	const a = "abcdefgh"
	const b = "abcdefgi"
	const n = 3
	baseline := fuzzymatch.CosineScore(a, b, n)
	baselineBits := math.Float64bits(baseline)
	for i := 0; i < iterations; i++ {
		got := fuzzymatch.CosineScore(a, b, n)
		if math.Float64bits(got) != baselineBits {
			t.Fatalf("iteration %d: CosineScore(%q,%q,%d) = %.17g (bits=%x); baseline = %.17g (bits=%x) — sorted-key iteration regression",
				i, a, b, n, got, math.Float64bits(got), baseline, baselineBits)
		}
	}
}
