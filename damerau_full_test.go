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

// damerau_full_test.go pins the public-API contract of damerau_full.go:
// the full reference-vector suite for DamerauLevenshteinFullDistance /
// DamerauLevenshteinFullScore (Lowrance-Wagner 1975).
//
// The Task 1 TDD canary (TestDamerauLevenshteinFull_DiscriminatingVector_Stub)
// lives in damerau_full_discriminator_test.go and is NOT duplicated here.
// This file ships the FULL reference-vector suite that builds on the stub.
//
// Primary source: Lowrance, R., Wagner, R. A. (1975). "An extension of the
// string-to-string correction problem." Journal of the ACM, 22(2):177-183.
//
// Discriminating vector: DamerauLevenshteinFullDistance("ca", "abc") == 2.
// This DIFFERS from DL-OSA (plan 02-05), which returns 3 for the same pair.
// See .planning/phases/02-core-character-algorithms-six/02-RESEARCH.md
// §DL-OSA vs DL-Full Divergence and the ROADMAP success criterion #2.
//
// Stdlib testing only — no testify in root tests, per
// .claude/skills/go-coding-standards.

package fuzzymatch_test

import (
	"math"
	"testing"

	"github.com/axonops/fuzzymatch"
)

// TestDamerauLevenshteinFull_BothEmpty asserts the documented identity invariant:
// Distance("","") = 0; Score("","") = 1.0.
func TestDamerauLevenshteinFull_BothEmpty(t *testing.T) {
	if got := fuzzymatch.DamerauLevenshteinFullDistance("", ""); got != 0 {
		t.Errorf("DamerauLevenshteinFullDistance(\"\",\"\") = %d; want 0", got)
	}
	if got := fuzzymatch.DamerauLevenshteinFullScore("", ""); got != 1.0 {
		t.Errorf("DamerauLevenshteinFullScore(\"\",\"\") = %g; want 1.0", got)
	}
}

// TestDamerauLevenshteinFull_OneEmpty covers ("", "abc") and ("abc", "") →
// distance 3, score 0.0.
func TestDamerauLevenshteinFull_OneEmpty(t *testing.T) {
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
			if got := fuzzymatch.DamerauLevenshteinFullDistance(tt.a, tt.b); got != tt.wantDist {
				t.Errorf("DamerauLevenshteinFullDistance(%q, %q) = %d; want %d", tt.a, tt.b, got, tt.wantDist)
			}
			if got := fuzzymatch.DamerauLevenshteinFullScore(tt.a, tt.b); got != tt.wantScr {
				t.Errorf("DamerauLevenshteinFullScore(%q, %q) = %g; want %g", tt.a, tt.b, got, tt.wantScr)
			}
		})
	}
}

// TestDamerauLevenshteinFull_Identical covers identical string pairs →
// distance 0, score 1.0.
func TestDamerauLevenshteinFull_Identical(t *testing.T) {
	tests := []string{"abc", "ab", "ba", "kitten", "saturday", "x", "hello world"}
	for _, s := range tests {
		t.Run(s, func(t *testing.T) {
			if got := fuzzymatch.DamerauLevenshteinFullDistance(s, s); got != 0 {
				t.Errorf("DamerauLevenshteinFullDistance(%q, %q) = %d; want 0", s, s, got)
			}
			if got := fuzzymatch.DamerauLevenshteinFullScore(s, s); got != 1.0 {
				t.Errorf("DamerauLevenshteinFullScore(%q, %q) = %g; want 1.0", s, s, got)
			}
		})
	}
}

// TestDamerauLevenshteinFull_ReferenceVectors is a table-driven test over the
// canonical Lowrance-Wagner 1975 reference vectors plus the discriminating
// vector. Float tolerance is 1e-9.
func TestDamerauLevenshteinFull_ReferenceVectors(t *testing.T) {
	tests := []struct {
		a, b      string
		wantDist  int
		wantScore float64
	}{
		// Lowrance-Wagner 1975 — simple transposition (same as OSA for this case)
		{"ab", "ba", 1, 0.5},
		// Lowrance-Wagner 1975 discriminating vector — DIFFERENT from OSA (which returns 3)
		{"ca", "abc", 2, 1.0 - float64(2)/float64(3)},
		// Identity
		{"abc", "abc", 0, 1.0},
		// One-empty
		{"", "abc", 3, 0.0},
		// Both-empty
		{"", "", 0, 1.0},
		// Single-char substitution
		{"a", "b", 1, 0.0},
		// Single-char insertion
		{"ab", "abc", 1, 1.0 - float64(1)/float64(3)},
		// Longer pair
		{"kitten", "sitting", 3, 1.0 - float64(3)/float64(7)},
	}
	const tol = 1e-9
	for _, tt := range tests {
		t.Run(tt.a+"_"+tt.b, func(t *testing.T) {
			gotDist := fuzzymatch.DamerauLevenshteinFullDistance(tt.a, tt.b)
			if gotDist != tt.wantDist {
				t.Errorf("DamerauLevenshteinFullDistance(%q, %q) = %d; want %d", tt.a, tt.b, gotDist, tt.wantDist)
			}
			gotScore := fuzzymatch.DamerauLevenshteinFullScore(tt.a, tt.b)
			if math.Abs(gotScore-tt.wantScore) > tol {
				t.Errorf("DamerauLevenshteinFullScore(%q, %q) = %.10f; want %.10f (tolerance %e)", tt.a, tt.b, gotScore, tt.wantScore, tol)
			}
		})
	}
}

// TestDamerauLevenshteinFull_DiscriminatingVector is the full assertion
// (richer than the Task 1 stub) that DamerauLevenshteinFullDistance("ca","abc")
// == 2 AND DamerauLevenshteinFullScore("ca","abc") ≈ 0.3333.
//
// This vector proves DL-Full differs from DL-OSA. The DL-OSA variant (plan
// 02-05) returns distance 3 (score 0.0) for the same pair because the OSA
// restriction forbids re-editing characters after a transposition.
//
// Reference: .planning/phases/02-core-character-algorithms-six/02-RESEARCH.md
// §DL-OSA vs DL-Full Divergence; ROADMAP success criterion #2 (cross-algorithm
// discriminating-vector contract).
func TestDamerauLevenshteinFull_DiscriminatingVector(t *testing.T) {
	const a, b = "ca", "abc"
	wantDist := 2
	wantScore := 1.0 - float64(2)/float64(3) // ≈ 0.3333...

	gotDist := fuzzymatch.DamerauLevenshteinFullDistance(a, b)
	if gotDist != wantDist {
		t.Errorf("DamerauLevenshteinFullDistance(%q,%q) = %d; want %d (NOT 3 — that is OSA's value)", a, b, gotDist, wantDist)
	}

	gotScore := fuzzymatch.DamerauLevenshteinFullScore(a, b)
	const tol = 1e-9
	if math.Abs(gotScore-wantScore) > tol {
		t.Errorf("DamerauLevenshteinFullScore(%q,%q) = %.10f; want %.10f (≈0.3333…, NOT 0.0 — that is OSA's value)", a, b, gotScore, wantScore)
	}

	// Explicitly assert the OSA value is different (cross-check):
	// DL-OSA returns 3 for this pair; DL-Full returns 2.
	osaDist := fuzzymatch.DamerauLevenshteinOSADistance(a, b)
	if osaDist != 3 {
		t.Errorf("Cross-check: DamerauLevenshteinOSADistance(%q,%q) = %d; want 3 (OSA discriminating vector must be 3, not 2)", a, b, osaDist)
	}
	if gotDist == osaDist {
		t.Errorf("DL-Full distance (%d) == DL-OSA distance (%d) for (%q,%q); they must diverge on this vector", gotDist, osaDist, a, b)
	}
}

// TestDamerauLevenshteinFull_Symmetry asserts Score(a,b) == Score(b,a) on
// the canonical reference vectors. The DL-Full distance D(a,b) == D(b,a) and
// max(len) is also symmetric, so the score is symmetric.
func TestDamerauLevenshteinFull_Symmetry(t *testing.T) {
	pairs := []struct{ a, b string }{
		{"ab", "ba"},
		{"ca", "abc"},
		{"abc", "abc"},
		{"kitten", "sitting"},
	}
	for _, p := range pairs {
		t.Run(p.a+"_"+p.b, func(t *testing.T) {
			fwd := fuzzymatch.DamerauLevenshteinFullScore(p.a, p.b)
			rev := fuzzymatch.DamerauLevenshteinFullScore(p.b, p.a)
			if fwd != rev {
				t.Errorf("DamerauLevenshteinFullScore(%q,%q) = %g != DamerauLevenshteinFullScore(%q,%q) = %g; symmetry violated", p.a, p.b, fwd, p.b, p.a, rev)
			}
		})
	}
}

// TestDamerauLevenshteinFull_DistanceRunes_MultiByte asserts the rune variant
// handles multi-byte UTF-8 correctly. "café" vs "cafe" differ by one rune
// (é → e), giving distance 1 and score 0.75 (1 - 1/4).
func TestDamerauLevenshteinFull_DistanceRunes_MultiByte(t *testing.T) {
	tests := []struct {
		a, b      string
		wantDist  int
		wantScore float64
	}{
		{"café", "cafe", 1, 0.75},
		{"naïve", "naive", 1, 1.0 - 1.0/5.0},
		{"ab", "ba", 1, 0.5}, // rune path, same as byte path for ASCII
		{"", "", 0, 1.0},
	}
	const tol = 1e-9
	for _, tt := range tests {
		t.Run(tt.a+"_"+tt.b, func(t *testing.T) {
			gotDist := fuzzymatch.DamerauLevenshteinFullDistanceRunes(tt.a, tt.b)
			if gotDist != tt.wantDist {
				t.Errorf("DamerauLevenshteinFullDistanceRunes(%q, %q) = %d; want %d", tt.a, tt.b, gotDist, tt.wantDist)
			}
			gotScore := fuzzymatch.DamerauLevenshteinFullScoreRunes(tt.a, tt.b)
			if math.Abs(gotScore-tt.wantScore) > tol {
				t.Errorf("DamerauLevenshteinFullScoreRunes(%q, %q) = %.10f; want %.10f", tt.a, tt.b, gotScore, tt.wantScore)
			}
		})
	}
}

// TestDamerauLevenshteinFullScore_ShortAllocBudget_ASCII pins the allocation
// floor for the DL-Full ASCII path on short inputs. Per the Q8a budget lock
// (CONTEXT.md §Q8a, docs/requirements.md §14.1), the budget is ≤ 1 alloc per
// call (the full DP table heap allocation; a fully zero-alloc stack-buffer
// path would require a ~34 KB stack frame at the 64-byte input ceiling — judged
// too fragile against Go's escape-analysis quirks for the v1.0 release).
//
// Q11e un-skip: previously skipped under the "ZeroAllocs" name; now asserts
// the actual Q8a-locked floor (≤ 1 alloc/op).
func TestDamerauLevenshteinFullScore_ShortAllocBudget_ASCII(t *testing.T) {
	// Pre-warm: invoke once outside AllocsPerRun to ensure any one-shot
	// init (none expected, but defensive) is not counted in the budget.
	_ = fuzzymatch.DamerauLevenshteinFullScore("ab", "ba")
	avg := testing.AllocsPerRun(100, func() {
		_ = fuzzymatch.DamerauLevenshteinFullScore("ab", "ba")
	})
	const budget = 1.0
	if avg > budget {
		t.Errorf("DamerauLevenshteinFullScore(\"ab\",\"ba\") averaged %.2f allocs/op over 100 runs; want ≤ %.0f (Q8a budget)", avg, budget)
	}
}
