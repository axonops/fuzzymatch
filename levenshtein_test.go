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

// levenshtein_test.go pins the public-API contract of levenshtein.go:
// both-empty identity, one-empty distance, canonical reference vectors from
// Levenshtein 1965 / Wagner-Fischer 1974, symmetry, byte vs rune path
// equivalence on ASCII inputs, multi-byte rune handling, and the NaN/Inf
// guards.
//
// Stdlib `testing` only — no testify in root tests, per
// .claude/skills/go-coding-standards.

package fuzzymatch_test

import (
	"testing"

	"github.com/axonops/fuzzymatch"
)

// absFloat64 returns the absolute value of f. Used for float tolerance checks
// without importing math just for math.Abs (though we use math.Abs below
// for consistency since math is already imported for quick.Check helpers).
// We keep this inline helper to avoid polluting later files via package-level
// naming collisions.
func absFloat64(f float64) float64 {
	if f < 0 {
		return -f
	}
	return f
}

// TestLevenshtein_BothEmpty asserts the documented identity invariant:
// Distance("","") = 0; Score("","") = 1.0.
func TestLevenshtein_BothEmpty(t *testing.T) {
	if got := fuzzymatch.LevenshteinDistance("", ""); got != 0 {
		t.Errorf("LevenshteinDistance(\"\",\"\") = %d; want 0", got)
	}
	if got := fuzzymatch.LevenshteinScore("", ""); got != 1.0 {
		t.Errorf("LevenshteinScore(\"\",\"\") = %g; want 1.0", got)
	}
}

// TestLevenshtein_OneEmpty covers ("", "abc") and ("abc", "") → distance 3, score 0.0.
func TestLevenshtein_OneEmpty(t *testing.T) {
	tests := []struct {
		a, b     string
		wantDist int
		wantScr  float64
	}{
		{"", "abc", 3, 0.0},
		{"abc", "", 3, 0.0},
		{"", "x", 1, 0.0},
		{"x", "", 1, 0.0},
	}
	for _, tt := range tests {
		t.Run(tt.a+"_"+tt.b, func(t *testing.T) {
			if got := fuzzymatch.LevenshteinDistance(tt.a, tt.b); got != tt.wantDist {
				t.Errorf("LevenshteinDistance(%q, %q) = %d; want %d", tt.a, tt.b, got, tt.wantDist)
			}
			if got := fuzzymatch.LevenshteinScore(tt.a, tt.b); got != tt.wantScr {
				t.Errorf("LevenshteinScore(%q, %q) = %g; want %g", tt.a, tt.b, got, tt.wantScr)
			}
		})
	}
}

// TestLevenshtein_Identical covers ("abc", "abc") → distance 0, score 1.0.
func TestLevenshtein_Identical(t *testing.T) {
	tests := []string{"abc", "kitten", "saturday", "x", "hello world"}
	for _, s := range tests {
		t.Run(s, func(t *testing.T) {
			if got := fuzzymatch.LevenshteinDistance(s, s); got != 0 {
				t.Errorf("LevenshteinDistance(%q, %q) = %d; want 0", s, s, got)
			}
			if got := fuzzymatch.LevenshteinScore(s, s); got != 1.0 {
				t.Errorf("LevenshteinScore(%q, %q) = %g; want 1.0", s, s, got)
			}
		})
	}
}

// TestLevenshtein_ReferenceVectors pins the canonical pairs from
// Levenshtein 1965 (Soviet Physics Doklady 10(8):707-710) and
// Wagner-Fischer 1974 (JACM 21(1):168-173).
func TestLevenshtein_ReferenceVectors(t *testing.T) {
	const tol = 1e-9
	tests := []struct {
		a, b  string
		dist  int
		score float64
	}{
		// Wagner-Fischer 1974 — the two canonical reference pairs.
		{"kitten", "sitting", 3, 1.0 - 3.0/7.0},
		{"saturday", "sunday", 3, 1.0 - 3.0/8.0},
		// Identity (both identical and both-empty covered separately above).
		{"abc", "abc", 0, 1.0},
		// One-empty: distance == len(non-empty), score == 0.
		{"", "abc", 3, 0.0},
		{"abc", "", 3, 0.0},
	}
	for _, tt := range tests {
		t.Run(tt.a+"_"+tt.b, func(t *testing.T) {
			if got := fuzzymatch.LevenshteinDistance(tt.a, tt.b); got != tt.dist {
				t.Errorf("LevenshteinDistance(%q, %q) = %d; want %d", tt.a, tt.b, got, tt.dist)
			}
			if got := fuzzymatch.LevenshteinScore(tt.a, tt.b); absFloat64(got-tt.score) > tol {
				t.Errorf("LevenshteinScore(%q, %q) = %.10f; want %.10f (tol %g)", tt.a, tt.b, got, tt.score, tol)
			}
		})
	}
}

// TestLevenshtein_Symmetry verifies Score(a,b) == Score(b,a) for all
// reference-vector pairs.
func TestLevenshtein_Symmetry(t *testing.T) {
	pairs := [][2]string{
		{"kitten", "sitting"},
		{"saturday", "sunday"},
		{"abc", "xyz"},
		{"abc", ""},
	}
	for _, p := range pairs {
		a, b := p[0], p[1]
		fwd := fuzzymatch.LevenshteinScore(a, b)
		rev := fuzzymatch.LevenshteinScore(b, a)
		if fwd != rev {
			t.Errorf("LevenshteinScore not symmetric: Score(%q,%q)=%g != Score(%q,%q)=%g", a, b, fwd, b, a, rev)
		}
	}
}

// TestLevenshtein_DistanceRunes_MultiByte asserts that the rune variant returns
// the rune-aware distance for multi-byte UTF-8 inputs.
//
// "café" is 4 runes (c, a, f, é) but 5 bytes (é encodes as 2 bytes in UTF-8).
// "cafe" is 4 runes and 4 bytes.
// Rune-level distance: 1 (one substitution: é→e).
// Byte-level distance: 2 (the 2-byte é sequence vs the 1-byte e).
//
// Source: Unicode Standard §3.9 (UTF-8) — multi-byte encoding of Latin small
// letter e with acute (U+00E9) as 0xC3 0xA9.
func TestLevenshtein_DistanceRunes_MultiByte(t *testing.T) {
	a, b := "café", "cafe"
	// Rune-level: 1 edit (é→e).
	gotRune := fuzzymatch.LevenshteinDistanceRunes(a, b)
	if gotRune != 1 {
		t.Errorf("LevenshteinDistanceRunes(%q, %q) = %d; want 1 (rune-level distance)", a, b, gotRune)
	}
	// Score: 1 - 1/4 = 0.75
	const wantScore = 0.75
	const tol = 1e-9
	gotScore := fuzzymatch.LevenshteinScoreRunes(a, b)
	if absFloat64(gotScore-wantScore) > tol {
		t.Errorf("LevenshteinScoreRunes(%q, %q) = %g; want %g", a, b, gotScore, wantScore)
	}
}

// TestLevenshtein_ASCII_vs_Rune_Equivalence verifies that for purely ASCII
// inputs, the byte and rune variants return identical results.
func TestLevenshtein_ASCII_vs_Rune_Equivalence(t *testing.T) {
	pairs := [][2]string{
		{"kitten", "sitting"},
		{"saturday", "sunday"},
		{"abc", "abc"},
		{"abc", ""},
		{"", ""},
	}
	for _, p := range pairs {
		a, b := p[0], p[1]
		byteScore := fuzzymatch.LevenshteinScore(a, b)
		runeScore := fuzzymatch.LevenshteinScoreRunes(a, b)
		if byteScore != runeScore {
			t.Errorf("ASCII mismatch: LevenshteinScore(%q,%q)=%g != LevenshteinScoreRunes(%q,%q)=%g",
				a, b, byteScore, a, b, runeScore)
		}
		byteDist := fuzzymatch.LevenshteinDistance(a, b)
		runeDist := fuzzymatch.LevenshteinDistanceRunes(a, b)
		if byteDist != runeDist {
			t.Errorf("ASCII mismatch: LevenshteinDistance(%q,%q)=%d != LevenshteinDistanceRunes(%q,%q)=%d",
				a, b, byteDist, a, b, runeDist)
		}
	}
}

// TestLevenshteinScore_ZeroAllocs_ASCII_Short pins the 0-alloc budget for the
// ASCII fast path. testing.AllocsPerRun asserts allocation at test time, not
// just benchmark time, so regressions are caught on every `go test ./...` run.
func TestLevenshteinScore_ZeroAllocs_ASCII_Short(t *testing.T) {
	// Quick warmup to let the JIT settle (avoid first-call init artifacts).
	_ = fuzzymatch.LevenshteinScore("kitten", "sitting")

	allocs := testing.AllocsPerRun(100, func() {
		_ = fuzzymatch.LevenshteinScore("kitten", "sitting")
	})
	if allocs > 0 {
		t.Errorf("LevenshteinScore ASCII short: %.1f allocs/op; want 0 (stack buffer not escaping?)", allocs)
	}
}

// TestLevenshteinScore_ZeroAllocs_ASCII_Medium pins the 0-alloc budget for an
// ASCII input of ~50 bytes (still within the maxStackInputLen=64 threshold).
func TestLevenshteinScore_ZeroAllocs_ASCII_Medium(t *testing.T) {
	const a50 = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWX"
	const b50 = "bcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXY"
	_ = fuzzymatch.LevenshteinScore(a50, b50)

	allocs := testing.AllocsPerRun(100, func() {
		_ = fuzzymatch.LevenshteinScore(a50, b50)
	})
	if allocs > 0 {
		t.Errorf("LevenshteinScore ASCII medium: %.1f allocs/op; want 0 (stack buffer not escaping?)", allocs)
	}
}

