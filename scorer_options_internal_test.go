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

// scorer_options_internal_test.go exposes a single probe helper
// (applyOptionForProbe) that the external package fuzzymatch_test
// tests in scorer_options_test.go use to drive a ScorerOption against
// a fresh scorerConfig and inspect both the returned error AND the
// resulting unexported configuration. The probe is the chosen mechanism
// (per CONTEXT.md §"Claude's Discretion") for plan 08-01 to unit-test
// the option layer in isolation — plan 08-02 will introduce NewScorer
// and at that point the external tests in scorer_options_test.go can
// also exercise the end-to-end path.
//
// Living in package fuzzymatch (no _test suffix) is the conventional Go
// pattern for exposing package-internal state to external test files:
// the build-tag _test.go suffix ensures this file never ships in the
// public artifact.

package fuzzymatch

// applyOptionForProbe runs a single ScorerOption against a fresh
// scorerConfig and returns the resulting state plus any error. The
// caller (scorer_options_test.go) can then assert on the error via
// errors.Is and on the resulting state via the helper accessors.
//
// This is a test-only helper. Production code constructs the
// scorerConfig inside NewScorer (plan 08-02) by iterating all options
// and short-circuiting on the first error.
func applyOptionForProbe(opt ScorerOption) (scorerConfig, error) {
	cfg := scorerConfig{}
	err := opt(&cfg)
	return cfg, err
}

// applyOptionsForProbe runs a sequence of ScorerOptions against a
// single fresh scorerConfig, returning the accumulated state plus the
// FIRST error encountered (mirroring NewScorer's short-circuit
// semantic).
func applyOptionsForProbe(opts ...ScorerOption) (scorerConfig, error) {
	cfg := scorerConfig{}
	for _, opt := range opts {
		if err := opt(&cfg); err != nil {
			return cfg, err
		}
	}
	return cfg, nil
}

// probeEntryCount reports the number of accumulated algorithm entries
// in a scorerConfig. Used by scorer_options_test.go to assert
// WithAlgorithm append behaviour and WithoutAlgorithm removal
// behaviour without exposing the slice itself.
func probeEntryCount(cfg scorerConfig) int {
	return len(cfg.entries)
}

// probeEntryAt returns the (id, weight) of the entry at the given
// index. Returns (-1, 0) for out-of-range indices so the test can
// distinguish a missing entry from a present zero-id entry.
func probeEntryAt(cfg scorerConfig, i int) (AlgoID, float64) {
	if i < 0 || i >= len(cfg.entries) {
		return AlgoID(-1), 0
	}
	return cfg.entries[i].id, cfg.entries[i].weight
}

// probeEntryHasScoreFn reports whether the entry at index i has a
// non-nil scoreFn closure. Production code requires every entry's
// scoreFn to be non-nil before Score is invoked (plan 08-02).
func probeEntryHasScoreFn(cfg scorerConfig, i int) bool {
	if i < 0 || i >= len(cfg.entries) {
		return false
	}
	return cfg.entries[i].scoreFn != nil
}

// probeThreshold returns (threshold, thresholdSet). NewScorer (plan
// 08-02) gates on thresholdSet == false → ErrMissingThreshold before
// any other validation; this helper lets the option-layer tests
// confirm WithThreshold sets BOTH fields atomically.
func probeThreshold(cfg scorerConfig) (float64, bool) {
	return cfg.threshold, cfg.thresholdSet
}

// probeNormalisation returns (applyNorm, normOpts). WithNormalisation
// sets applyNorm = true AND stores opts; WithoutNormalisation sets
// applyNorm = false without touching normOpts.
func probeNormalisation(cfg scorerConfig) (bool, NormalisationOptions) {
	return cfg.applyNorm, cfg.normOpts
}

// probeNormaliseWeights returns the normaliseWeights flag. Default
// at NewScorer time is true; WithNormaliseWeights(false) opts out.
func probeNormaliseWeights(cfg scorerConfig) bool {
	return cfg.normaliseWeights
}

// probeScoreFnInvoke runs the entry's scoreFn with the supplied
// inputs. Used by parameterised-option tests in Task 3 to confirm
// each closure dispatches to the correct underlying score function
// with the consumer-supplied parameters captured.
func probeScoreFnInvoke(cfg scorerConfig, i int, a, b string) float64 {
	if i < 0 || i >= len(cfg.entries) || cfg.entries[i].scoreFn == nil {
		return -1
	}
	return cfg.entries[i].scoreFn(a, b)
}
