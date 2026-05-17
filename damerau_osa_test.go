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

// damerau_osa_test.go pins the public-API contract of damerau_osa.go:
// the full reference-vector suite for DamerauLevenshteinOSADistance /
// DamerauLevenshteinOSAScore.
//
// The Task 1 TDD canary (TestDamerauLevenshteinOSA_DiscriminatingVector_Stub)
// lives in damerau_osa_discriminator_test.go and is NOT duplicated here.
// This file ships the FULL reference-vector suite that builds on the stub.
//
// Primary source: Boytsov, L. (2011). "Indexing methods for approximate
// dictionary searching." ACM Journal of Experimental Algorithmics, 16, §3.1.
// Historical source: Damerau, F. J. (1964). Communications of the ACM, 7(3):171-176.
//
// Stdlib testing only — no testify in root tests, per
// .claude/skills/go-coding-standards.

package fuzzymatch_test

import (
	"testing"

	"github.com/axonops/fuzzymatch"
)

// TestDamerauLevenshteinOSA_BothEmpty asserts the documented identity invariant:
// Distance("","") = 0; Score("","") = 1.0.
func TestDamerauLevenshteinOSA_BothEmpty(t *testing.T) {
	if got := fuzzymatch.DamerauLevenshteinOSADistance("", ""); got != 0 {
		t.Errorf("DamerauLevenshteinOSADistance(\"\",\"\") = %d; want 0", got)
	}
	if got := fuzzymatch.DamerauLevenshteinOSAScore("", ""); got != 1.0 {
		t.Errorf("DamerauLevenshteinOSAScore(\"\",\"\") = %g; want 1.0", got)
	}
}

// TestDamerauLevenshteinOSA_OneEmpty covers ("", "abc") and ("abc", "") →
// distance 3, score 0.0.
func TestDamerauLevenshteinOSA_OneEmpty(t *testing.T) {
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
			if got := fuzzymatch.DamerauLevenshteinOSADistance(tt.a, tt.b); got != tt.wantDist {
				t.Errorf("DamerauLevenshteinOSADistance(%q, %q) = %d; want %d", tt.a, tt.b, got, tt.wantDist)
			}
			if got := fuzzymatch.DamerauLevenshteinOSAScore(tt.a, tt.b); got != tt.wantScr {
				t.Errorf("DamerauLevenshteinOSAScore(%q, %q) = %g; want %g", tt.a, tt.b, got, tt.wantScr)
			}
		})
	}
}

// TestDamerauLevenshteinOSA_Identical covers identical string pairs →
// distance 0, score 1.0.
func TestDamerauLevenshteinOSA_Identical(t *testing.T) {
	tests := []string{"abc", "ab", "ba", "kitten", "saturday", "x", "hello world"}
	for _, s := range tests {
		t.Run(s, func(t *testing.T) {
			if got := fuzzymatch.DamerauLevenshteinOSADistance(s, s); got != 0 {
				t.Errorf("DamerauLevenshteinOSADistance(%q, %q) = %d; want 0", s, s, got)
			}
			if got := fuzzymatch.DamerauLevenshteinOSAScore(s, s); got != 1.0 {
				t.Errorf("DamerauLevenshteinOSAScore(%q, %q) = %g; want 1.0", s, s, got)
			}
			// Exercise the byte-identity fast path in the Runes variants:
			// when a == b, both variants short-circuit before the []rune
			// conversion, returning distance 0 / score 1.0 without
			// allocating rune slices.
			if got := fuzzymatch.DamerauLevenshteinOSADistanceRunes(s, s); got != 0 {
				t.Errorf("DamerauLevenshteinOSADistanceRunes(%q, %q) = %d; want 0 (fast identity)", s, s, got)
			}
			if got := fuzzymatch.DamerauLevenshteinOSAScoreRunes(s, s); got != 1.0 {
				t.Errorf("DamerauLevenshteinOSAScoreRunes(%q, %q) = %g; want 1.0 (fast identity)", s, s, got)
			}
		})
	}
}

// TestDamerauLevenshteinOSA_ReferenceVectors pins canonical pairs from
// Boytsov 2011 §3.1 and Damerau 1964.
func TestDamerauLevenshteinOSA_ReferenceVectors(t *testing.T) {
	const tol = 1e-9
	tests := []struct {
		a, b  string
		dist  int
		score float64
	}{
		// Transposition: "ab" → "ba" costs 1 edit.
		{"ab", "ba", 1, 0.5},
		// Discriminating vector: OSA returns 3, Full DL returns 2.
		// Boytsov 2011 §3.1 — the OSA constraint forbids re-editing after
		// transposition, making "ca"→"abc" cost 3 under OSA.
		{"ca", "abc", 3, 0.0},
		// Identity.
		{"abc", "abc", 0, 1.0},
		// One-empty.
		{"", "abc", 3, 0.0},
		{"abc", "", 3, 0.0},
	}
	for _, tt := range tests {
		t.Run(tt.a+"_"+tt.b, func(t *testing.T) {
			if got := fuzzymatch.DamerauLevenshteinOSADistance(tt.a, tt.b); got != tt.dist {
				t.Errorf("DamerauLevenshteinOSADistance(%q, %q) = %d; want %d", tt.a, tt.b, got, tt.dist)
			}
			got := fuzzymatch.DamerauLevenshteinOSAScore(tt.a, tt.b)
			if absFloat64(got-tt.score) > tol {
				t.Errorf("DamerauLevenshteinOSAScore(%q, %q) = %.10f; want %.10f (tol %g)", tt.a, tt.b, got, tt.score, tol)
			}
		})
	}
}

// TestDamerauLevenshteinOSA_DiscriminatingVector is the full (richer) assertion
// of the discriminating-vector contract, complementing the Task 1 stub.
//
// Pins: DamerauLevenshteinOSADistance("ca", "abc") == 3 (RESEARCH.md and
// ROADMAP success criterion #2). DL-Full (plan 02-06) returns 2 for the same
// pair. The score 0.0 follows from normalisation: 1 - 3/max(2,3) = 1-3/3 = 0.
func TestDamerauLevenshteinOSA_DiscriminatingVector(t *testing.T) {
	// Forward direction.
	if got := fuzzymatch.DamerauLevenshteinOSADistance("ca", "abc"); got != 3 {
		t.Errorf("DamerauLevenshteinOSADistance(\"ca\",\"abc\") = %d; want 3 (NOT 2 — that is DL-Full's value, Boytsov 2011 §3.1)", got)
	}
	// Reverse direction: symmetric.
	if got := fuzzymatch.DamerauLevenshteinOSADistance("abc", "ca"); got != 3 {
		t.Errorf("DamerauLevenshteinOSADistance(\"abc\",\"ca\") = %d; want 3 (symmetry of OSA distance)", got)
	}
	// Score: 1 - 3/3 = 0.0 exactly.
	if got := fuzzymatch.DamerauLevenshteinOSAScore("ca", "abc"); got != 0.0 {
		t.Errorf("DamerauLevenshteinOSAScore(\"ca\",\"abc\") = %g; want 0.0 (distance 3 / maxLen 3)", got)
	}
}

// TestDamerauLevenshteinOSA_Symmetry verifies Score(a,b) == Score(b,a) for
// all reference-vector pairs.
func TestDamerauLevenshteinOSA_Symmetry(t *testing.T) {
	pairs := [][2]string{
		{"ab", "ba"},
		{"ca", "abc"},
		{"abc", "abc"},
		{"abc", ""},
		{"kitten", "sitting"},
	}
	for _, p := range pairs {
		a, b := p[0], p[1]
		fwd := fuzzymatch.DamerauLevenshteinOSAScore(a, b)
		rev := fuzzymatch.DamerauLevenshteinOSAScore(b, a)
		if fwd != rev {
			t.Errorf("DamerauLevenshteinOSAScore not symmetric: Score(%q,%q)=%g != Score(%q,%q)=%g", a, b, fwd, b, a, rev)
		}
	}
}

// TestDamerauLevenshteinOSA_DistanceRunes_MultiByte asserts the rune variant
// returns the rune-aware distance for multi-byte UTF-8 inputs.
//
// "café" is 4 runes but 5 bytes; "cafe" is 4 runes and 4 bytes.
// Rune-level: 1 substitution (é→e). Byte-level: 2 (2-byte é vs 1-byte e).
func TestDamerauLevenshteinOSA_DistanceRunes_MultiByte(t *testing.T) {
	a, b := "café", "cafe"
	gotRune := fuzzymatch.DamerauLevenshteinOSADistanceRunes(a, b)
	if gotRune != 1 {
		t.Errorf("DamerauLevenshteinOSADistanceRunes(%q, %q) = %d; want 1 (rune-level distance)", a, b, gotRune)
	}
	const wantScore = 0.75
	const tol = 1e-9
	gotScore := fuzzymatch.DamerauLevenshteinOSAScoreRunes(a, b)
	if absFloat64(gotScore-wantScore) > tol {
		t.Errorf("DamerauLevenshteinOSAScoreRunes(%q, %q) = %g; want %g", a, b, gotScore, wantScore)
	}
}

// TestDamerauLevenshteinOSAScore_ZeroAllocs_ASCII_Short pins the 0-alloc budget
// for the ASCII fast path. testing.AllocsPerRun asserts allocation at test time,
// not just benchmark time, so regressions are caught on every `go test ./...` run.
func TestDamerauLevenshteinOSAScore_ZeroAllocs_ASCII_Short(t *testing.T) {
	allocs := testing.AllocsPerRun(100, func() {
		_ = fuzzymatch.DamerauLevenshteinOSAScore("ab", "ba")
	})
	if allocs != 0 {
		t.Errorf("DamerauLevenshteinOSAScore(\"ab\",\"ba\") allocates %g heap objects; want 0 (ASCII fast path must use stack buffer)", allocs)
	}
}
