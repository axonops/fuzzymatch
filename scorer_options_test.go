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
			if !errors.Is(err, ErrInvalidAlgorithm) {
				t.Errorf("WithAlgorithm(%v, 1.0) err = %v; want ErrInvalidAlgorithm", c.id, err)
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
