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

// scorer_options_test.go covers happy + error paths for every Phase 8
// ScorerOption introduced in plan 08-01. The tests use the
// applyOptionForProbe / applyOptionsForProbe helpers from
// scorer_options_internal_test.go (package fuzzymatch) to exercise the
// option layer in isolation BEFORE plan 08-02's NewScorer constructor
// exists. Once plan 08-02 lands, end-to-end NewScorer-based tests will
// supersede the probe-based ones; the probes remain as a thin internal
// gate for unexported scorerConfig invariants.
//
// Stdlib testing only — no testify in root (per CLAUDE.md and
// .claude/skills/go-coding-standards/SKILL.md).

package fuzzymatch

import (
	"errors"
	"testing"
)

// --- Non-parameterised options (Task 2) -----------------------------------

func TestWithAlgorithm_HappyPath(t *testing.T) {
	cfg, err := applyOptionForProbe(WithAlgorithm(AlgoLevenshtein, 0.5))
	if err != nil {
		t.Fatalf("WithAlgorithm(AlgoLevenshtein, 0.5) returned err = %v; want nil", err)
	}
	if got := probeEntryCount(cfg); got != 1 {
		t.Fatalf("entry count = %d; want 1", got)
	}
	id, w := probeEntryAt(cfg, 0)
	if id != AlgoLevenshtein {
		t.Errorf("entry[0].id = %v; want AlgoLevenshtein", id)
	}
	if w != 0.5 {
		t.Errorf("entry[0].weight = %g; want 0.5", w)
	}
	if !probeEntryHasScoreFn(cfg, 0) {
		t.Errorf("entry[0].scoreFn is nil; want dispatch[AlgoLevenshtein]")
	}
}

func TestWithAlgorithm_InvalidWeight(t *testing.T) {
	cases := []struct {
		name   string
		weight float64
	}{
		{"zero", 0},
		{"negative", -0.1},
		{"large_negative", -1.5},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			_, err := applyOptionForProbe(WithAlgorithm(AlgoLevenshtein, c.weight))
			if !errors.Is(err, ErrInvalidWeight) {
				t.Errorf("WithAlgorithm(AlgoLevenshtein, %g) err = %v; want ErrInvalidWeight", c.weight, err)
			}
		})
	}
}

func TestWithAlgorithm_InvalidAlgoID(t *testing.T) {
	cases := []struct {
		name string
		id   AlgoID
	}{
		{"out_of_range_high", AlgoID(999)},
		{"out_of_range_negative", AlgoID(-1)},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			_, err := applyOptionForProbe(WithAlgorithm(c.id, 1.0))
			if !errors.Is(err, ErrInvalidAlgoID) {
				t.Errorf("WithAlgorithm(%v, 1.0) err = %v; want ErrInvalidAlgoID", c.id, err)
			}
		})
	}
}

func TestWithoutAlgorithm_NoOpOnAbsent(t *testing.T) {
	cfg, err := applyOptionForProbe(WithoutAlgorithm(AlgoLevenshtein))
	if err != nil {
		t.Fatalf("WithoutAlgorithm on fresh config returned err = %v; want nil (silent no-op)", err)
	}
	if got := probeEntryCount(cfg); got != 0 {
		t.Errorf("entry count = %d; want 0 (config was empty before and after)", got)
	}
}

func TestWithoutAlgorithm_RemovesPresent(t *testing.T) {
	cfg, err := applyOptionsForProbe(
		WithAlgorithm(AlgoLevenshtein, 0.5),
		WithAlgorithm(AlgoJaro, 0.3),
		WithoutAlgorithm(AlgoLevenshtein),
	)
	if err != nil {
		t.Fatalf("applyOptionsForProbe returned err = %v; want nil", err)
	}
	if got := probeEntryCount(cfg); got != 1 {
		t.Fatalf("entry count = %d; want 1 (Levenshtein removed)", got)
	}
	id, w := probeEntryAt(cfg, 0)
	if id != AlgoJaro {
		t.Errorf("remaining entry id = %v; want AlgoJaro", id)
	}
	if w != 0.3 {
		t.Errorf("remaining entry weight = %g; want 0.3", w)
	}
}

func TestWithoutAlgorithm_RemovesAllMatchingEntries(t *testing.T) {
	// If the same AlgoID was added twice (last-write-wins is resolved by
	// NewScorer in plan 08-02, not by the option layer), WithoutAlgorithm
	// must remove ALL matching entries so subsequent NewScorer dedup
	// sees zero of them.
	cfg, err := applyOptionsForProbe(
		WithAlgorithm(AlgoLevenshtein, 0.3),
		WithAlgorithm(AlgoLevenshtein, 0.7),
		WithoutAlgorithm(AlgoLevenshtein),
	)
	if err != nil {
		t.Fatalf("applyOptionsForProbe returned err = %v; want nil", err)
	}
	if got := probeEntryCount(cfg); got != 0 {
		t.Errorf("entry count = %d; want 0 (both Levenshtein entries removed)", got)
	}
}

func TestWithThreshold_Inclusive(t *testing.T) {
	cases := []struct {
		name string
		t    float64
	}{
		{"lower_bound", 0.0},
		{"upper_bound", 1.0},
		{"midpoint", 0.5},
		{"just_above_zero", 0.001},
		{"just_below_one", 0.999},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			cfg, err := applyOptionForProbe(WithThreshold(c.t))
			if err != nil {
				t.Fatalf("WithThreshold(%g) err = %v; want nil", c.t, err)
			}
			thr, set := probeThreshold(cfg)
			if !set {
				t.Errorf("thresholdSet = false; want true")
			}
			if thr != c.t {
				t.Errorf("threshold = %g; want %g", thr, c.t)
			}
		})
	}
}

func TestWithThreshold_OutOfRange(t *testing.T) {
	cases := []struct {
		name string
		t    float64
	}{
		{"negative", -0.1},
		{"above_one", 1.5},
		{"large_negative", -100},
		{"large_positive", 100},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			cfg, err := applyOptionForProbe(WithThreshold(c.t))
			if !errors.Is(err, ErrInvalidThreshold) {
				t.Errorf("WithThreshold(%g) err = %v; want ErrInvalidThreshold", c.t, err)
			}
			// On rejection, thresholdSet must stay false so NewScorer's
			// missing-threshold check still fires.
			_, set := probeThreshold(cfg)
			if set {
				t.Errorf("thresholdSet = true after rejected WithThreshold; want false")
			}
		})
	}
}

func TestWithNormalisation_StoresOpts(t *testing.T) {
	want := DefaultNormalisationOptions()
	cfg, err := applyOptionForProbe(WithNormalisation(want))
	if err != nil {
		t.Fatalf("WithNormalisation err = %v; want nil", err)
	}
	apply, got := probeNormalisation(cfg)
	if !apply {
		t.Errorf("applyNorm = false; want true")
	}
	if got != want {
		t.Errorf("normOpts = %+v; want %+v", got, want)
	}
}

func TestWithoutNormalisation_SetsFlag(t *testing.T) {
	// First enable, then disable, to confirm the flag is updated.
	cfg, err := applyOptionsForProbe(
		WithNormalisation(DefaultNormalisationOptions()),
		WithoutNormalisation(),
	)
	if err != nil {
		t.Fatalf("applyOptionsForProbe err = %v; want nil", err)
	}
	apply, _ := probeNormalisation(cfg)
	if apply {
		t.Errorf("applyNorm = true after WithoutNormalisation; want false")
	}
}

func TestWithNormaliseWeights_StoresValue(t *testing.T) {
	t.Run("true", func(t *testing.T) {
		cfg, err := applyOptionForProbe(WithNormaliseWeights(true))
		if err != nil {
			t.Fatalf("WithNormaliseWeights(true) err = %v; want nil", err)
		}
		if !probeNormaliseWeights(cfg) {
			t.Errorf("normaliseWeights = false; want true")
		}
	})
	t.Run("false", func(t *testing.T) {
		cfg, err := applyOptionForProbe(WithNormaliseWeights(false))
		if err != nil {
			t.Fatalf("WithNormaliseWeights(false) err = %v; want nil", err)
		}
		if probeNormaliseWeights(cfg) {
			t.Errorf("normaliseWeights = true; want false")
		}
	})
}

// --- Parameterised options (Task 3) -----------------------------------------

// assertParameterisedHappy is shared by the four q-gram options: they all
// share the (weight float64, n int) signature and the same accumulator
// invariants — single entry appended with the expected id / weight /
// non-nil scoreFn. The scoreFn is then exercised on a deterministic
// (a, b) pair to confirm the closure captures n correctly.
func assertParameterisedHappy(t *testing.T, opt ScorerOption, wantID AlgoID, wantWeight float64) {
	t.Helper()
	cfg, err := applyOptionForProbe(opt)
	if err != nil {
		t.Fatalf("option err = %v; want nil", err)
	}
	if got := probeEntryCount(cfg); got != 1 {
		t.Fatalf("entry count = %d; want 1", got)
	}
	id, w := probeEntryAt(cfg, 0)
	if id != wantID {
		t.Errorf("entry[0].id = %v; want %v", id, wantID)
	}
	if w != wantWeight {
		t.Errorf("entry[0].weight = %g; want %g", w, wantWeight)
	}
	if !probeEntryHasScoreFn(cfg, 0) {
		t.Errorf("entry[0].scoreFn is nil")
	}
}

// --- WithQGramJaccardAlgorithm ----------------------------------------------

func TestWithQGramJaccardAlgorithm_HappyPath(t *testing.T) {
	assertParameterisedHappy(t, WithQGramJaccardAlgorithm(0.5, 3), AlgoQGramJaccard, 0.5)
}

func TestWithQGramJaccardAlgorithm_CapturesN(t *testing.T) {
	// Confirm n is captured into the closure (not hardcoded to dispatch
	// default n=3). Compute the option's closure score against the
	// expected QGramJaccardScore(a, b, n) for n=2 and n=5.
	const a, b = "kitten", "sitting"
	for _, n := range []int{2, 5} {
		cfg, err := applyOptionForProbe(WithQGramJaccardAlgorithm(1.0, n))
		if err != nil {
			t.Fatalf("n=%d: option err = %v", n, err)
		}
		got := probeScoreFnInvoke(cfg, 0, a, b)
		want := QGramJaccardScore(a, b, n)
		if got != want {
			t.Errorf("n=%d: closure score = %g; want QGramJaccardScore(_,_,%d) = %g", n, got, n, want)
		}
	}
}

func TestWithQGramJaccardAlgorithm_InvalidWeight(t *testing.T) {
	for _, w := range []float64{0, -0.1, -1.0} {
		_, err := applyOptionForProbe(WithQGramJaccardAlgorithm(w, 3))
		if !errors.Is(err, ErrInvalidWeight) {
			t.Errorf("weight=%g err = %v; want ErrInvalidWeight", w, err)
		}
	}
}

func TestWithQGramJaccardAlgorithm_InvalidN(t *testing.T) {
	for _, n := range []int{0, -1, -100} {
		_, err := applyOptionForProbe(WithQGramJaccardAlgorithm(1.0, n))
		if !errors.Is(err, ErrInvalidQGramSize) {
			t.Errorf("n=%d err = %v; want ErrInvalidQGramSize", n, err)
		}
	}
}

// --- WithSorensenDiceAlgorithm ----------------------------------------------

func TestWithSorensenDiceAlgorithm_HappyPath(t *testing.T) {
	assertParameterisedHappy(t, WithSorensenDiceAlgorithm(0.4, 2), AlgoSorensenDice, 0.4)
}

func TestWithSorensenDiceAlgorithm_CapturesN(t *testing.T) {
	const a, b = "kitten", "sitting"
	for _, n := range []int{2, 5} {
		cfg, err := applyOptionForProbe(WithSorensenDiceAlgorithm(1.0, n))
		if err != nil {
			t.Fatalf("n=%d: option err = %v", n, err)
		}
		got := probeScoreFnInvoke(cfg, 0, a, b)
		want := SorensenDiceScore(a, b, n)
		if got != want {
			t.Errorf("n=%d: closure score = %g; want SorensenDiceScore(_,_,%d) = %g", n, got, n, want)
		}
	}
}

func TestWithSorensenDiceAlgorithm_InvalidWeight(t *testing.T) {
	_, err := applyOptionForProbe(WithSorensenDiceAlgorithm(0, 3))
	if !errors.Is(err, ErrInvalidWeight) {
		t.Errorf("weight=0 err = %v; want ErrInvalidWeight", err)
	}
}

func TestWithSorensenDiceAlgorithm_InvalidN(t *testing.T) {
	_, err := applyOptionForProbe(WithSorensenDiceAlgorithm(1.0, 0))
	if !errors.Is(err, ErrInvalidQGramSize) {
		t.Errorf("n=0 err = %v; want ErrInvalidQGramSize", err)
	}
}

// --- WithCosineAlgorithm ----------------------------------------------------

func TestWithCosineAlgorithm_HappyPath(t *testing.T) {
	assertParameterisedHappy(t, WithCosineAlgorithm(0.6, 3), AlgoCosine, 0.6)
}

func TestWithCosineAlgorithm_CapturesN(t *testing.T) {
	const a, b = "kitten", "sitting"
	for _, n := range []int{2, 4} {
		cfg, err := applyOptionForProbe(WithCosineAlgorithm(1.0, n))
		if err != nil {
			t.Fatalf("n=%d: option err = %v", n, err)
		}
		got := probeScoreFnInvoke(cfg, 0, a, b)
		want := CosineScore(a, b, n)
		if got != want {
			t.Errorf("n=%d: closure score = %g; want CosineScore(_,_,%d) = %g", n, got, n, want)
		}
	}
}

func TestWithCosineAlgorithm_InvalidWeight(t *testing.T) {
	_, err := applyOptionForProbe(WithCosineAlgorithm(-1, 3))
	if !errors.Is(err, ErrInvalidWeight) {
		t.Errorf("weight=-1 err = %v; want ErrInvalidWeight", err)
	}
}

func TestWithCosineAlgorithm_InvalidN(t *testing.T) {
	_, err := applyOptionForProbe(WithCosineAlgorithm(1.0, -1))
	if !errors.Is(err, ErrInvalidQGramSize) {
		t.Errorf("n=-1 err = %v; want ErrInvalidQGramSize", err)
	}
}

// --- WithTverskyAlgorithm ---------------------------------------------------

func TestWithTverskyAlgorithm_HappyPath(t *testing.T) {
	assertParameterisedHappy(t, WithTverskyAlgorithm(0.7, 1.0, 1.0, 3), AlgoTversky, 0.7)
}

func TestWithTverskyAlgorithm_CapturesParams(t *testing.T) {
	const a, b = "abcdef", "abcxyz"
	cases := []struct {
		name        string
		alpha, beta float64
		n           int
	}{
		{"jaccard_like", 1.0, 1.0, 3},
		{"asymmetric", 0.5, 2.0, 2},
		{"large_alpha", 10.0, 0.1, 4},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			cfg, err := applyOptionForProbe(WithTverskyAlgorithm(1.0, c.alpha, c.beta, c.n))
			if err != nil {
				t.Fatalf("option err = %v", err)
			}
			got := probeScoreFnInvoke(cfg, 0, a, b)
			want := TverskyScore(a, b, c.n, c.alpha, c.beta)
			if got != want {
				t.Errorf("closure score = %g; want TverskyScore(a, b, %d, %g, %g) = %g", got, c.n, c.alpha, c.beta, want)
			}
		})
	}
}

func TestWithTverskyAlgorithm_InvalidWeight(t *testing.T) {
	_, err := applyOptionForProbe(WithTverskyAlgorithm(0, 1.0, 1.0, 3))
	if !errors.Is(err, ErrInvalidWeight) {
		t.Errorf("weight=0 err = %v; want ErrInvalidWeight", err)
	}
}

func TestWithTverskyAlgorithm_InvalidN(t *testing.T) {
	_, err := applyOptionForProbe(WithTverskyAlgorithm(1.0, 1.0, 1.0, 0))
	if !errors.Is(err, ErrInvalidQGramSize) {
		t.Errorf("n=0 err = %v; want ErrInvalidQGramSize", err)
	}
}

func TestWithTverskyAlgorithm_InvalidAlpha(t *testing.T) {
	_, err := applyOptionForProbe(WithTverskyAlgorithm(1.0, -0.1, 1.0, 3))
	if !errors.Is(err, ErrInvalidTverskyParam) {
		t.Errorf("alpha=-0.1 err = %v; want ErrInvalidTverskyParam", err)
	}
}

func TestWithTverskyAlgorithm_InvalidBeta(t *testing.T) {
	_, err := applyOptionForProbe(WithTverskyAlgorithm(1.0, 1.0, -0.1, 3))
	if !errors.Is(err, ErrInvalidTverskyParam) {
		t.Errorf("beta=-0.1 err = %v; want ErrInvalidTverskyParam", err)
	}
}

// --- WithMongeElkanAlgorithm ------------------------------------------------

func TestWithMongeElkanAlgorithm_HappyPath(t *testing.T) {
	assertParameterisedHappy(t, WithMongeElkanAlgorithm(0.8, AlgoJaroWinkler), AlgoMongeElkan, 0.8)
}

func TestWithMongeElkanAlgorithm_CapturesInner(t *testing.T) {
	const a, b = "alpha beta", "alpha gamma"
	cases := []struct {
		name  string
		inner AlgoID
	}{
		{"JaroWinkler", AlgoJaroWinkler},
		{"Levenshtein", AlgoLevenshtein},
		{"Jaro", AlgoJaro},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			cfg, err := applyOptionForProbe(WithMongeElkanAlgorithm(1.0, c.inner))
			if err != nil {
				t.Fatalf("inner=%v option err = %v", c.inner, err)
			}
			got := probeScoreFnInvoke(cfg, 0, a, b)
			want := MongeElkanScore(a, b, c.inner)
			if got != want {
				t.Errorf("inner=%v: closure score = %g; want MongeElkanScore (symmetric default) = %g", c.inner, got, want)
			}
		})
	}
}

func TestWithMongeElkanAlgorithm_InvalidWeight(t *testing.T) {
	_, err := applyOptionForProbe(WithMongeElkanAlgorithm(0, AlgoJaroWinkler))
	if !errors.Is(err, ErrInvalidWeight) {
		t.Errorf("weight=0 err = %v; want ErrInvalidWeight", err)
	}
}

func TestWithMongeElkanAlgorithm_RejectsSelf(t *testing.T) {
	// Trivial recursion guard — MongeElkan as its own inner is
	// infinite-loop-equivalent at Score time. The option layer
	// short-circuits here (not at runtime).
	_, err := applyOptionForProbe(WithMongeElkanAlgorithm(1.0, AlgoMongeElkan))
	if !errors.Is(err, ErrInvalidAlgoID) {
		t.Errorf("inner=AlgoMongeElkan err = %v; want ErrInvalidAlgoID", err)
	}
}

func TestWithMongeElkanAlgorithm_InvalidInner(t *testing.T) {
	for _, inner := range []AlgoID{AlgoID(999), AlgoID(-1)} {
		_, err := applyOptionForProbe(WithMongeElkanAlgorithm(1.0, inner))
		if !errors.Is(err, ErrInvalidAlgoID) {
			t.Errorf("inner=%v err = %v; want ErrInvalidAlgoID", inner, err)
		}
	}
}

// --- WithSmithWatermanGotohAlgorithm ----------------------------------------

func TestWithSmithWatermanGotohAlgorithm_HappyPath(t *testing.T) {
	assertParameterisedHappy(t, WithSmithWatermanGotohAlgorithm(0.5, NewSWGParams()), AlgoSmithWatermanGotoh, 0.5)
}

func TestWithSmithWatermanGotohAlgorithm_CapturesParams(t *testing.T) {
	const a, b = "kitten", "sitting"
	params := NewSWGParams()
	params.Match = 2.0
	params.Mismatch = -0.5
	cfg, err := applyOptionForProbe(WithSmithWatermanGotohAlgorithm(1.0, params))
	if err != nil {
		t.Fatalf("option err = %v", err)
	}
	got := probeScoreFnInvoke(cfg, 0, a, b)
	want := SmithWatermanGotohScoreWithParams(a, b, params)
	if got != want {
		t.Errorf("closure score = %g; want SmithWatermanGotohScoreWithParams = %g", got, want)
	}
}

func TestWithSmithWatermanGotohAlgorithm_InvalidWeight(t *testing.T) {
	_, err := applyOptionForProbe(WithSmithWatermanGotohAlgorithm(-1, NewSWGParams()))
	if !errors.Is(err, ErrInvalidWeight) {
		t.Errorf("weight=-1 err = %v; want ErrInvalidWeight", err)
	}
}
