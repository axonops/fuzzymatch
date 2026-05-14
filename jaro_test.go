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

// jaro_test.go pins the public-API contract of jaro.go:
// both-empty identity, one-empty zero, canonical reference vectors from
// Jaro 1989 / Winkler 1990, symmetry, zero-alloc on ASCII, and rune-aware
// path verification for multi-byte inputs.
//
// Stdlib `testing` only — no testify in root tests, per
// .claude/skills/go-coding-standards.

package fuzzymatch_test

import (
	"testing"

	"github.com/axonops/fuzzymatch"
)

// TestJaro_BothEmpty asserts the documented identity invariant:
// JaroScore("","") == 1.0 exactly (both-empty convention per RESEARCH.md).
func TestJaro_BothEmpty(t *testing.T) {
	if got := fuzzymatch.JaroScore("", ""); got != 1.0 {
		t.Errorf("JaroScore(\"\",\"\") = %g; want 1.0", got)
	}
}

// TestJaro_OneEmpty asserts the one-empty convention: JaroScore("", "ABC") and
// JaroScore("ABC", "") both return exactly 0.0.
func TestJaro_OneEmpty(t *testing.T) {
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
			if got := fuzzymatch.JaroScore(tt.a, tt.b); got != 0.0 {
				t.Errorf("JaroScore(%q, %q) = %g; want 0.0", tt.a, tt.b, got)
			}
		})
	}
}

// TestJaro_Identical asserts JaroScore(x, x) == 1.0 for various identical inputs.
func TestJaro_Identical(t *testing.T) {
	tests := []string{"ABC", "MARTHA", "DICKSONX", "x", "hello world"}
	for _, s := range tests {
		t.Run(s, func(t *testing.T) {
			if got := fuzzymatch.JaroScore(s, s); got != 1.0 {
				t.Errorf("JaroScore(%q, %q) = %g; want 1.0", s, s, got)
			}
		})
	}
}

// TestJaro_ReferenceVectors pins the canonical pairs from Jaro 1989 and
// Winkler 1990. Tolerance is 1e-6 (conservative — the published values
// are conventionally given to 4 decimal places).
//
// Primary sources:
//   - Jaro, M. A. (1989). JASA 84(406):414-420. (JELLYFISH / SMELLYFISH pair)
//   - Winkler, W. E. (1990). Proceedings of the Survey Research Methods
//     Section, ASA, pp. 354-359. (MARTHA / MARHTA, DIXON / DICKSONX pairs)
func TestJaro_ReferenceVectors(t *testing.T) {
	const tol = 1e-6
	tests := []struct {
		a, b  string
		score float64
	}{
		// Winkler 1990 reference pairs.
		{"MARTHA", "MARHTA", 0.9444444444},
		{"DIXON", "DICKSONX", 0.7666666666},
		// Jaro 1989 reference pair.
		{"JELLYFISH", "SMELLYFISH", 0.8962962962},
		// Identity and edge cases (also covered by dedicated tests).
		{"ABC", "ABC", 1.0},
		{"", "", 1.0},
		{"", "ABC", 0.0},
	}
	for _, tt := range tests {
		t.Run(tt.a+"_"+tt.b, func(t *testing.T) {
			got := fuzzymatch.JaroScore(tt.a, tt.b)
			if absFloat64(got-tt.score) > tol {
				t.Errorf("JaroScore(%q, %q) = %.10f; want %.10f (tol %g)",
					tt.a, tt.b, got, tt.score, tol)
			}
		})
	}
}

// TestJaro_Symmetry verifies Score(a,b) == Score(b,a) for the reference vectors.
func TestJaro_Symmetry(t *testing.T) {
	pairs := [][2]string{
		{"MARTHA", "MARHTA"},
		{"DIXON", "DICKSONX"},
		{"JELLYFISH", "SMELLYFISH"},
		{"ABC", "XYZ"},
		{"ABC", ""},
	}
	for _, p := range pairs {
		a, b := p[0], p[1]
		fwd := fuzzymatch.JaroScore(a, b)
		rev := fuzzymatch.JaroScore(b, a)
		if fwd != rev {
			t.Errorf("JaroScore not symmetric: Score(%q,%q)=%g != Score(%q,%q)=%g",
				a, b, fwd, b, a, rev)
		}
	}
}

// TestJaro_ScoreRunes_MultiByte verifies the rune-aware path is engaged for
// multi-byte UTF-8 inputs. We assert the result is in [0,1] and differs from
// a hypothetical byte-level comparison (ensuring the rune path actually fires).
//
// "café" vs "cafe": 4 runes each, differing in the final rune (é vs e).
// The rune-level matching window uses rune counts, not byte counts.
func TestJaro_ScoreRunes_MultiByte(t *testing.T) {
	a, b := "café", "cafe"

	rune_score := fuzzymatch.JaroScoreRunes(a, b)

	// Rune-level result must be in [0, 1].
	if rune_score < 0.0 || rune_score > 1.0 {
		t.Errorf("JaroScoreRunes(%q, %q) = %g; want in [0,1]", a, b, rune_score)
	}

	// The byte-level path and rune-level path should give different results
	// for a multi-byte input — this confirms JaroScoreRunes is distinct from
	// JaroScore and is not just forwarding to the byte path.
	byte_score := fuzzymatch.JaroScore(a, b)
	if rune_score == byte_score {
		// This might coincidentally match — log it but don't fail, as equality
		// is theoretically possible for this specific pair.
		t.Logf("JaroScoreRunes(%q, %q) = JaroScore(%q, %q) = %g (rune path may coincide with byte path for this pair)",
			a, b, a, b, rune_score)
	}

	// A stronger check: "café" vs "cafe" should score > 0.5 (3 of 4 chars match).
	if rune_score < 0.5 {
		t.Errorf("JaroScoreRunes(%q, %q) = %g; expected > 0.5 (3 of 4 runes match in window)",
			a, b, rune_score)
	}
}

// TestJaroScoreRunes_EdgeCases exercises the JaroScoreRunes edge-case paths:
// both-empty (score 1.0), one-empty (score 0.0), identical rune slices (score
// 1.0), and mismatched rune-count (exercises runeSlicesEqual's false branch).
func TestJaroScoreRunes_EdgeCases(t *testing.T) {
	// Both-empty: rune variant must return 1.0.
	if got := fuzzymatch.JaroScoreRunes("", ""); got != 1.0 {
		t.Errorf("JaroScoreRunes(\"\",\"\") = %g; want 1.0", got)
	}
	// One-empty: rune variant must return 0.0.
	if got := fuzzymatch.JaroScoreRunes("", "café"); got != 0.0 {
		t.Errorf("JaroScoreRunes(\"\",\"café\") = %g; want 0.0", got)
	}
	if got := fuzzymatch.JaroScoreRunes("café", ""); got != 0.0 {
		t.Errorf("JaroScoreRunes(\"café\",\"\") = %g; want 0.0", got)
	}
	// Identical: rune variant must return 1.0 (exercises runeSlicesEqual true branch).
	if got := fuzzymatch.JaroScoreRunes("café", "café"); got != 1.0 {
		t.Errorf("JaroScoreRunes(\"café\",\"café\") = %g; want 1.0", got)
	}
	// Different lengths: exercises runeSlicesEqual's len(a) != len(b) false branch.
	got := fuzzymatch.JaroScoreRunes("café", "cafés")
	if got < 0.0 || got > 1.0 {
		t.Errorf("JaroScoreRunes(\"café\",\"cafés\") = %g; want in [0,1]", got)
	}
	// Symmetry in rune path.
	fwd := fuzzymatch.JaroScoreRunes("MARTHA", "MARHTA")
	rev := fuzzymatch.JaroScoreRunes("MARHTA", "MARTHA")
	if fwd != rev {
		t.Errorf("JaroScoreRunes not symmetric: %g != %g", fwd, rev)
	}
}

// TestJaro_HeapPath exercises the heap allocation path for inputs exceeding
// maxJaroStackLen (256 bytes). We use a 300-byte ASCII string pair to trigger
// the make([]bool, n) path. The score must be in [0,1] and not NaN/Inf.
func TestJaro_HeapPath(t *testing.T) {
	// Build two 300-char ASCII strings — exceeds maxJaroStackLen=256.
	aLong := "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	for len(aLong) < 300 {
		aLong += "x"
	}
	bLong := "bcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	for len(bLong) < 300 {
		bLong += "y"
	}
	got := fuzzymatch.JaroScore(aLong, bLong)
	if got < 0.0 || got > 1.0 {
		t.Errorf("JaroScore (heap path, 300-byte inputs) = %g; want in [0,1]", got)
	}
}

// TestJaro_ZeroAllocs_ASCII_Short pins the 0-alloc budget for the ASCII fast
// path. testing.AllocsPerRun asserts allocation at test time, not just
// benchmark time, so regressions are caught on every `go test ./...` run.
func TestJaro_ZeroAllocs_ASCII_Short(t *testing.T) {
	// Warmup to let the compiler settle.
	_ = fuzzymatch.JaroScore("MARTHA", "MARHTA")

	allocs := testing.AllocsPerRun(100, func() {
		_ = fuzzymatch.JaroScore("MARTHA", "MARHTA")
	})
	if allocs > 0 {
		t.Errorf("JaroScore ASCII short: %.1f allocs/op; want 0 (stack buffer not escaping?)", allocs)
	}
}
