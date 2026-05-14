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

// jarowinkler_test.go pins the public-API contract of jarowinkler.go:
// both-empty identity, one-empty zero, canonical reference vectors from
// Winkler 1990, boost threshold gate, prefix cap at 4, symmetry, zero-alloc
// on ASCII, and constants traceability.
//
// Stdlib `testing` only — no testify in root tests, per
// .claude/skills/go-coding-standards.

package fuzzymatch_test

import (
	"testing"

	"github.com/axonops/fuzzymatch"
)

// TestJaroWinkler_BothEmpty asserts the documented identity invariant:
// JaroWinklerScore("","") == 1.0 exactly (both-empty convention).
func TestJaroWinkler_BothEmpty(t *testing.T) {
	if got := fuzzymatch.JaroWinklerScore("", ""); got != 1.0 {
		t.Errorf("JaroWinklerScore(\"\",\"\") = %g; want 1.0", got)
	}
}

// TestJaroWinkler_OneEmpty asserts the one-empty convention:
// JaroWinklerScore("", "ABC") and JaroWinklerScore("ABC", "") both return
// exactly 0.0.
func TestJaroWinkler_OneEmpty(t *testing.T) {
	tests := []struct {
		a, b string
	}{
		{"", "ABC"},
		{"ABC", ""},
		{"", "x"},
		{"x", ""},
		{"", "MARTHA"},
	}
	for _, tt := range tests {
		t.Run(tt.a+"_"+tt.b, func(t *testing.T) {
			if got := fuzzymatch.JaroWinklerScore(tt.a, tt.b); got != 0.0 {
				t.Errorf("JaroWinklerScore(%q, %q) = %g; want 0.0", tt.a, tt.b, got)
			}
		})
	}
}

// TestJaroWinkler_Identical asserts JaroWinklerScore(x, x) == 1.0 for various
// identical inputs.
func TestJaroWinkler_Identical(t *testing.T) {
	tests := []string{"ABC", "MARTHA", "DICKSONX", "x", "hello world"}
	for _, s := range tests {
		t.Run(s, func(t *testing.T) {
			if got := fuzzymatch.JaroWinklerScore(s, s); got != 1.0 {
				t.Errorf("JaroWinklerScore(%q, %q) = %g; want 1.0", s, s, got)
			}
		})
	}
}

// TestJaroWinkler_ReferenceVectors pins the canonical pairs from Winkler 1990
// p. 357. Tolerance is 1e-6 for MARTHA/MARHTA and DIXON/DICKSONX (precision
// pairs); 1e-3 for DWAYNE/DUANE (where the published value is given less
// precisely).
//
// Primary source:
//   - Winkler, W. E. (1990). "String comparator metrics and enhanced decision
//     rules in the Fellegi-Sunter model of record linkage." Proceedings of the
//     Section on Survey Research Methods, ASA: 354-359. (p. 357)
func TestJaroWinkler_ReferenceVectors(t *testing.T) {
	tests := []struct {
		a, b  string
		score float64
		tol   float64
	}{
		// Winkler 1990 p. 357 canonical reference pairs — tight tolerance.
		{"MARTHA", "MARHTA", 0.9611111111, 1e-6},
		{"DIXON", "DICKSONX", 0.8133333333, 1e-6},
		// DWAYNE/DUANE — looser tolerance (value given to ~4 decimal places).
		{"DWAYNE", "DUANE", 0.8400, 1e-3},
		// Edge cases.
		{"ABC", "ABC", 1.0, 0.0},
		{"", "", 1.0, 0.0},
		{"", "ABC", 0.0, 0.0},
	}
	for _, tt := range tests {
		t.Run(tt.a+"_"+tt.b, func(t *testing.T) {
			got := fuzzymatch.JaroWinklerScore(tt.a, tt.b)
			diff := got - tt.score
			if diff < 0 {
				diff = -diff
			}
			if diff > tt.tol {
				t.Errorf("JaroWinklerScore(%q, %q) = %.10f; want %.10f (tol %g)",
					tt.a, tt.b, got, tt.score, tt.tol)
			}
		})
	}
}

// TestJaroWinkler_BoostThresholdGate verifies that pairs with underlying Jaro
// score below the boost threshold (0.7) return the unmodified Jaro score.
//
// "abc" vs "xyz" — no character matches → Jaro = 0.0 (well below 0.7).
// JaroWinklerScore must equal JaroScore for the same pair.
func TestJaroWinkler_BoostThresholdGate(t *testing.T) {
	pairs := [][2]string{
		{"abc", "xyz"}, // Jaro = 0.0 (no matches)
		{"aaa", "bbb"}, // Jaro = 0.0 (no matches)
		{"ab", "yz"},   // Jaro = 0.0 (no matches)
	}
	for _, p := range pairs {
		a, b := p[0], p[1]
		jw := fuzzymatch.JaroWinklerScore(a, b)
		j := fuzzymatch.JaroScore(a, b)
		if jw != j {
			t.Errorf("BoostThresholdGate: JaroWinklerScore(%q,%q)=%g != JaroScore(%q,%q)=%g (below-threshold pair must return Jaro unchanged)",
				a, b, jw, a, b, j)
		}
	}
}

// TestJaroWinkler_PrefixCapAt4 verifies that a common prefix longer than 4
// characters is capped at 4. The pair "TESTABCD" / "TESTABCE" shares a 7-char
// common prefix "TESTABC", but L must be capped at 4.
func TestJaroWinkler_PrefixCapAt4(t *testing.T) {
	a, b := "TESTABCD", "TESTABCE"
	j := fuzzymatch.JaroScore(a, b)

	// Expected JW with L=4 (cap):
	expectedJW := j + float64(4)*0.1*(1.0-j)
	// Wrongly-uncapped JW would use L=7:
	wrongJW := j + float64(7)*0.1*(1.0-j)

	got := fuzzymatch.JaroWinklerScore(a, b)

	const tol = 1e-9
	diff := got - expectedJW
	if diff < 0 {
		diff = -diff
	}
	if diff > tol {
		t.Errorf("JaroWinklerScore(%q,%q) = %.10f; want %.10f (L=4 capped, not L=7 = %.10f)",
			a, b, got, expectedJW, wrongJW)
	}
	// Sanity: assert it's not equal to the wrongly-uncapped value.
	if got == wrongJW {
		t.Errorf("JaroWinklerScore(%q,%q) = %.10f which equals the uncapped (L=7) value — prefix cap not applied",
			a, b, got)
	}
}

// TestJaroWinkler_Symmetry verifies Score(a,b) == Score(b,a) for the
// canonical reference vectors and additional pairs.
func TestJaroWinkler_Symmetry(t *testing.T) {
	pairs := [][2]string{
		{"MARTHA", "MARHTA"},
		{"DIXON", "DICKSONX"},
		{"DWAYNE", "DUANE"},
		{"ABC", "XYZ"},
		{"ABC", ""},
		{"TESTABCD", "TESTABCE"},
	}
	for _, p := range pairs {
		a, b := p[0], p[1]
		fwd := fuzzymatch.JaroWinklerScore(a, b)
		rev := fuzzymatch.JaroWinklerScore(b, a)
		if fwd != rev {
			t.Errorf("JaroWinklerScore not symmetric: Score(%q,%q)=%g != Score(%q,%q)=%g",
				a, b, fwd, b, a, rev)
		}
	}
}

// TestJaroWinkler_ConstantsTraceable pins the three Winkler 1990 constants
// against accidental drift. Each constant is re-exported via export_test.go
// (WinklerPrefixScaleForTest, WinklerMaxPrefixForTest, WinklerBoostThresholdForTest).
//
// If any constant is modified, this test fails immediately — the constants are
// LOCKED by REQUIREMENTS.md CHAR-06 and CONTEXT.md per Winkler 1990 p. 357.
func TestJaroWinkler_ConstantsTraceable(t *testing.T) {
	if fuzzymatch.WinklerPrefixScaleForTest != 0.1 {
		t.Errorf("winklerPrefixScale = %g; want exactly 0.1 (Winkler 1990 p. 357)",
			fuzzymatch.WinklerPrefixScaleForTest)
	}
	if fuzzymatch.WinklerMaxPrefixForTest != 4 {
		t.Errorf("winklerMaxPrefix = %d; want exactly 4 (Winkler 1990 p. 357 L_max)",
			fuzzymatch.WinklerMaxPrefixForTest)
	}
	if fuzzymatch.WinklerBoostThresholdForTest != 0.7 {
		t.Errorf("winklerBoostThreshold = %g; want exactly 0.7 (Winkler 1990 p. 357 boost gate)",
			fuzzymatch.WinklerBoostThresholdForTest)
	}
}

// TestJaroWinklerScore_ZeroAllocs_ASCII_Short pins the 0-alloc budget for the
// ASCII fast path. JaroWinklerScore adds only a constant-bounded prefix loop
// over JaroScore — no additional allocations.
func TestJaroWinklerScore_ZeroAllocs_ASCII_Short(t *testing.T) {
	// Warmup.
	_ = fuzzymatch.JaroWinklerScore("MARTHA", "MARHTA")

	allocs := testing.AllocsPerRun(100, func() {
		_ = fuzzymatch.JaroWinklerScore("MARTHA", "MARHTA")
	})
	if allocs > 0 {
		t.Errorf("JaroWinklerScore ASCII short: %.1f allocs/op; want 0 (stack buffer not escaping?)", allocs)
	}
}

// TestJaroWinklerScoreRunes_MultiByte verifies the rune-aware path handles
// multi-byte UTF-8 inputs correctly. The score must be in [0,1] and > 0.5
// for strings that share most runes.
func TestJaroWinklerScoreRunes_MultiByte(t *testing.T) {
	a, b := "café", "café!"
	got := fuzzymatch.JaroWinklerScoreRunes(a, b)
	if got < 0.0 || got > 1.0 {
		t.Errorf("JaroWinklerScoreRunes(%q,%q) = %g; want in [0,1]", a, b, got)
	}
	// Strings sharing a common rune prefix should score > 0.5.
	if got < 0.5 {
		t.Errorf("JaroWinklerScoreRunes(%q,%q) = %g; expected > 0.5", a, b, got)
	}
	// Symmetry.
	rev := fuzzymatch.JaroWinklerScoreRunes(b, a)
	if got != rev {
		t.Errorf("JaroWinklerScoreRunes not symmetric: %g != %g", got, rev)
	}
}

// TestJaroWinklerScoreRunes_BelowBoostThreshold verifies that the rune-aware
// path returns the raw Jaro score (no prefix bonus) when Jaro < 0.7
// (winklerBoostThreshold). "abc" vs "xyz" have zero common characters so
// JaroScore == 0.0, which is below the 0.7 gate — the boost branch is skipped.
func TestJaroWinklerScoreRunes_BelowBoostThreshold(t *testing.T) {
	// "abc" vs "xyz": no common characters → Jaro = 0 → below 0.7 gate.
	got := fuzzymatch.JaroWinklerScoreRunes("abc", "xyz")
	jaroScore := fuzzymatch.JaroScoreRunes("abc", "xyz")
	if got != jaroScore {
		t.Errorf("JaroWinklerScoreRunes(\"abc\",\"xyz\") = %g; expected raw Jaro (%g) because Jaro < 0.7 gate", got, jaroScore)
	}
	// Must be exactly 0 since strings share no characters.
	if got != 0.0 {
		t.Errorf("JaroWinklerScoreRunes(\"abc\",\"xyz\") = %g; want 0.0 (no matches)", got)
	}
}

// TestJaroWinklerScoreRunes_IdentityAndEmpty verifies the rune path on edge
// cases: identical non-empty strings (score 1.0) and both-empty (score 1.0).
func TestJaroWinklerScoreRunes_IdentityAndEmpty(t *testing.T) {
	if got := fuzzymatch.JaroWinklerScoreRunes("ABC", "ABC"); got != 1.0 {
		t.Errorf("JaroWinklerScoreRunes(\"ABC\",\"ABC\") = %g; want 1.0 (identity)", got)
	}
	if got := fuzzymatch.JaroWinklerScoreRunes("", ""); got != 1.0 {
		t.Errorf("JaroWinklerScoreRunes(\"\",\"\") = %g; want 1.0 (both-empty)", got)
	}
	if got := fuzzymatch.JaroWinklerScoreRunes("", "ABC"); got != 0.0 {
		t.Errorf("JaroWinklerScoreRunes(\"\",\"ABC\") = %g; want 0.0 (one-empty)", got)
	}
}
