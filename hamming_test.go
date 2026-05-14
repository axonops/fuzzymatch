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

// hamming_test.go pins the public-API contract of hamming.go:
// both-empty identity, identical strings, canonical reference vectors
// from Hamming 1950, the LOCKED unequal-length silent-zero policy, symmetry,
// rune-level multi-byte distance, and the 0-alloc ASCII guarantee.
//
// Stdlib `testing` only — no testify in root tests, per
// .claude/skills/go-coding-standards.

package fuzzymatch_test

import (
	"math"
	"testing"

	"github.com/axonops/fuzzymatch"
)

// TestHamming_BothEmpty asserts the documented identity invariant:
// Distance("","") = 0; Score("","") = 1.0.
func TestHamming_BothEmpty(t *testing.T) {
	if got := fuzzymatch.HammingDistance("", ""); got != 0 {
		t.Errorf("HammingDistance(\"\",\"\") = %d; want 0", got)
	}
	if got := fuzzymatch.HammingScore("", ""); got != 1.0 {
		t.Errorf("HammingScore(\"\",\"\") = %g; want 1.0", got)
	}
}

// TestHamming_Identical asserts distance 0, score 1.0 for identical strings.
func TestHamming_Identical(t *testing.T) {
	tests := []string{"abc", "karolin", "x", "hello world"}
	for _, s := range tests {
		t.Run(s, func(t *testing.T) {
			if got := fuzzymatch.HammingDistance(s, s); got != 0 {
				t.Errorf("HammingDistance(%q, %q) = %d; want 0", s, s, got)
			}
			if got := fuzzymatch.HammingScore(s, s); got != 1.0 {
				t.Errorf("HammingScore(%q, %q) = %g; want 1.0", s, s, got)
			}
		})
	}
}

// TestHamming_ReferenceVectors pins the canonical pairs from
// Hamming 1950 (Bell System Technical Journal, 29(2):147-160).
func TestHamming_ReferenceVectors(t *testing.T) {
	const tol = 1e-9
	tests := []struct {
		a, b  string
		dist  int
		score float64
	}{
		// Hamming 1950 — canonical reference vectors.
		{"karolin", "kathrin", 3, 1.0 - 3.0/7.0},
		// Second canonical pair (binary strings, 7 positions, 2 differences).
		{"1011101", "1001001", 2, 1.0 - 2.0/7.0},
		// Identity.
		{"abc", "abc", 0, 1.0},
		// Both-empty.
		{"", "", 0, 1.0},
	}
	for _, tt := range tests {
		t.Run(tt.a+"_"+tt.b, func(t *testing.T) {
			if got := fuzzymatch.HammingDistance(tt.a, tt.b); got != tt.dist {
				t.Errorf("HammingDistance(%q, %q) = %d; want %d", tt.a, tt.b, got, tt.dist)
			}
			if got := fuzzymatch.HammingScore(tt.a, tt.b); math.Abs(got-tt.score) > tol {
				t.Errorf("HammingScore(%q, %q) = %.10f; want %.10f (tol %g)", tt.a, tt.b, got, tt.score, tol)
			}
		})
	}
}

// TestHamming_UnequalLength_SilentZero explicitly tests the LOCKED
// unequal-length silent-zero policy from CONTEXT.md: HammingDistance returns
// max(len(a), len(b)) and HammingScore returns exactly 0.0 — no error, no panic.
func TestHamming_UnequalLength_SilentZero(t *testing.T) {
	tests := []struct {
		a, b     string
		wantDist int
		wantScr  float64
	}{
		// abc (3) vs ab (2): max=3, score=0.0
		{"abc", "ab", 3, 0.0},
		// ab (2) vs abc (3): symmetric — max=3, score=0.0
		{"ab", "abc", 3, 0.0},
		// empty vs non-empty
		{"", "abc", 3, 0.0},
		{"abc", "", 3, 0.0},
		// single vs empty
		{"x", "", 1, 0.0},
		{"", "x", 1, 0.0},
		// longer on left
		{"abcde", "ab", 5, 0.0},
		// longer on right
		{"ab", "abcde", 5, 0.0},
	}
	for _, tt := range tests {
		t.Run(tt.a+"_"+tt.b, func(t *testing.T) {
			if got := fuzzymatch.HammingDistance(tt.a, tt.b); got != tt.wantDist {
				t.Errorf("HammingDistance(%q, %q) = %d; want %d (unequal-length policy: max(len))", tt.a, tt.b, got, tt.wantDist)
			}
			if got := fuzzymatch.HammingScore(tt.a, tt.b); got != tt.wantScr {
				t.Errorf("HammingScore(%q, %q) = %g; want %g exactly (unequal-length silent-zero policy)", tt.a, tt.b, got, tt.wantScr)
			}
		})
	}
}

// TestHamming_Symmetry verifies Score(a,b) == Score(b,a) for reference
// vectors AND for the unequal-length silent-zero case.
func TestHamming_Symmetry(t *testing.T) {
	pairs := [][2]string{
		{"karolin", "kathrin"},
		{"1011101", "1001001"},
		{"abc", "abc"},
		{"abc", "ab"}, // unequal-length: both directions should return 0.0
		{"ab", "abc"},
		{"abc", ""},
		{"", "abc"},
	}
	for _, p := range pairs {
		a, b := p[0], p[1]
		fwd := fuzzymatch.HammingScore(a, b)
		rev := fuzzymatch.HammingScore(b, a)
		if fwd != rev {
			t.Errorf("HammingScore not symmetric: Score(%q,%q)=%g != Score(%q,%q)=%g", a, b, fwd, b, a, rev)
		}
	}
}

// TestHamming_DistanceRunes_MultiByte asserts that the rune variant returns
// the rune-aware distance for multi-byte UTF-8 inputs.
//
// "café" is 4 runes (c, a, f, é) but 5 bytes.
// "cafè" is also 4 runes (c, a, f, è) but also 5 bytes.
// Rune-level distance: 1 (only the final rune differs: é ≠ è).
//
// Source: Unicode Standard §3.9 — é (U+00E9) and è (U+00E8) each encode as
// 2 bytes in UTF-8, both are distinct rune values.
func TestHamming_DistanceRunes_MultiByte(t *testing.T) {
	a, b := "café", "cafè"
	// Rune-level: 1 mismatch (é vs è at position 3).
	gotDist := fuzzymatch.HammingDistanceRunes(a, b)
	if gotDist != 1 {
		t.Errorf("HammingDistanceRunes(%q, %q) = %d; want 1 (rune-level distance)", a, b, gotDist)
	}
	// Score: 1 - 1/4 = 0.75
	const wantScore = 0.75
	const tol = 1e-9
	gotScore := fuzzymatch.HammingScoreRunes(a, b)
	if math.Abs(gotScore-wantScore) > tol {
		t.Errorf("HammingScoreRunes(%q, %q) = %g; want %g", a, b, gotScore, wantScore)
	}
}

// TestHamming_DistanceRunes_UnequalLength verifies the rune variant's
// unequal-rune-count path: HammingDistanceRunes returns max(runeCount(a),
// runeCount(b)) and HammingScoreRunes returns 0.0.
func TestHamming_DistanceRunes_UnequalLength(t *testing.T) {
	tests := []struct {
		a, b     string
		wantDist int
		wantScr  float64
	}{
		// Purely ASCII unequal-length inputs.
		{"abc", "ab", 3, 0.0},
		{"ab", "abc", 3, 0.0},
		// Multi-byte rune: "café" (4 runes) vs "caf" (3 runes).
		{"café", "caf", 4, 0.0},
		{"caf", "café", 4, 0.0},
		// Empty vs non-empty.
		{"", "abc", 3, 0.0},
		{"abc", "", 3, 0.0},
	}
	for _, tt := range tests {
		t.Run(tt.a+"_"+tt.b, func(t *testing.T) {
			if got := fuzzymatch.HammingDistanceRunes(tt.a, tt.b); got != tt.wantDist {
				t.Errorf("HammingDistanceRunes(%q, %q) = %d; want %d", tt.a, tt.b, got, tt.wantDist)
			}
			if got := fuzzymatch.HammingScoreRunes(tt.a, tt.b); got != tt.wantScr {
				t.Errorf("HammingScoreRunes(%q, %q) = %g; want %g (unequal-length silent-zero)", tt.a, tt.b, got, tt.wantScr)
			}
		})
	}
}

// TestHamming_ScoreRunes_BothEmpty asserts HammingScoreRunes("","") == 1.0.
func TestHamming_ScoreRunes_BothEmpty(t *testing.T) {
	if got := fuzzymatch.HammingScoreRunes("", ""); got != 1.0 {
		t.Errorf("HammingScoreRunes(\"\",\"\") = %g; want 1.0", got)
	}
	if got := fuzzymatch.HammingDistanceRunes("", ""); got != 0 {
		t.Errorf("HammingDistanceRunes(\"\",\"\") = %d; want 0", got)
	}
}

// TestHammingScore_ZeroAllocs pins the 0-alloc budget for the ASCII byte
// path at any length. Hamming is a single counting loop with no DP buffer,
// so it is trivially zero-alloc for all ASCII inputs.
func TestHammingScore_ZeroAllocs(t *testing.T) {
	// Warmup to let any first-call initialisation settle.
	_ = fuzzymatch.HammingScore("karolin", "kathrin")

	allocs := testing.AllocsPerRun(100, func() {
		_ = fuzzymatch.HammingScore("karolin", "kathrin")
	})
	if allocs > 0 {
		t.Errorf("HammingScore ASCII: %.1f allocs/op; want 0", allocs)
	}
}
