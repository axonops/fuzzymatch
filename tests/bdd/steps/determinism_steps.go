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

// determinism_steps.go contains the godog step definitions for the
// Phase 8.5 Plan 17b Gap 3 determinism.feature scenarios.
//
// The DeterminismContext struct holds state between steps within a
// scenario; each scenario instantiates a fresh DeterminismContext
// via InitDeterminismSteps (called from algorithms_steps.go's
// InitializeScenario).
//
// All determinism scenarios live under the @determinism Gherkin tag
// in features/determinism.feature. Six scenarios cover the
// determinism guarantees documented in docs/requirements.md §13.

package steps

import (
	"fmt"
	"math"

	"github.com/cucumber/godog"

	"github.com/axonops/fuzzymatch"
)

// DeterminismContext holds state between BDD steps within a single
// determinism scenario. inputA / inputB hold the two strings; score1
// / score2 hold the two successive scalar scores; scoreMap1 /
// scoreMap2 hold the two successive ScoreAll maps; algorithms holds
// the slice returned by DefaultScorer.Algorithms.
type DeterminismContext struct {
	inputA, inputB string
	score1, score2 float64
	scoreMap1      map[fuzzymatch.AlgoID]float64
	scoreMap2      map[fuzzymatch.AlgoID]float64
	algorithms     []fuzzymatch.ScorerAlgorithm
}

// theDeterminismInputPair sets inputA and inputB from the Gherkin
// step.
// Step regex: `^the determinism input pair "([^"]*)" and "([^"]*)"$`
func (dc *DeterminismContext) theDeterminismInputPair(a, b string) error {
	dc.inputA = a
	dc.inputB = b
	dc.score1 = math.NaN()
	dc.score2 = math.NaN()
	dc.scoreMap1 = nil
	dc.scoreMap2 = nil
	dc.algorithms = nil
	return nil
}

// iCallDefaultScorerScoreTwice runs DefaultScorer().Score(a, b)
// twice and stores both results.
// Step regex: `^I call DefaultScorer\.Score twice$`
func (dc *DeterminismContext) iCallDefaultScorerScoreTwice() error {
	s := fuzzymatch.DefaultScorer()
	dc.score1 = s.Score(dc.inputA, dc.inputB)
	dc.score2 = s.Score(dc.inputA, dc.inputB)
	return nil
}

// iCallDefaultScorerScoreAllTwice runs DefaultScorer().ScoreAll(a, b)
// twice and stores both maps.
// Step regex: `^I call DefaultScorer\.ScoreAll twice$`
func (dc *DeterminismContext) iCallDefaultScorerScoreAllTwice() error {
	s := fuzzymatch.DefaultScorer()
	dc.scoreMap1 = s.ScoreAll(dc.inputA, dc.inputB)
	dc.scoreMap2 = s.ScoreAll(dc.inputA, dc.inputB)
	return nil
}

// iCallLevenshteinScoreTwice runs LevenshteinScore(a, b) twice and
// stores both scalar results.
// Step regex: `^I call LevenshteinScore twice$`
func (dc *DeterminismContext) iCallLevenshteinScoreTwice() error {
	dc.score1 = fuzzymatch.LevenshteinScore(dc.inputA, dc.inputB)
	dc.score2 = fuzzymatch.LevenshteinScore(dc.inputA, dc.inputB)
	return nil
}

// iCallJaroWinklerScoreTwice runs JaroWinklerScore(a, b) twice.
// Step regex: `^I call JaroWinklerScore twice$`
func (dc *DeterminismContext) iCallJaroWinklerScoreTwice() error {
	dc.score1 = fuzzymatch.JaroWinklerScore(dc.inputA, dc.inputB)
	dc.score2 = fuzzymatch.JaroWinklerScore(dc.inputA, dc.inputB)
	return nil
}

// iCallCosineScoreTwice runs CosineScore(a, b, 3) twice — the
// default trigram n value matching the dispatch wrapper.
// Step regex: `^I call CosineScore twice$`
func (dc *DeterminismContext) iCallCosineScoreTwice() error {
	const n = 3
	dc.score1 = fuzzymatch.CosineScore(dc.inputA, dc.inputB, n)
	dc.score2 = fuzzymatch.CosineScore(dc.inputA, dc.inputB, n)
	return nil
}

// iCallDefaultScorerAlgorithms invokes DefaultScorer().Algorithms()
// and stores the result for the order-check assertion.
// Step regex: `^I call DefaultScorer\.Algorithms$`
func (dc *DeterminismContext) iCallDefaultScorerAlgorithms() error {
	dc.algorithms = fuzzymatch.DefaultScorer().Algorithms()
	return nil
}

// bothScoresShouldBeByteIdentical asserts score1 == score2 as raw
// float64 bits. NaN sentinel detection: if both values are NaN we
// treat the bit-pattern check as authoritative (math.Float64bits)
// because Go's == operator returns false for two NaN operands.
// Step regex: `^both scores should be byte-identical$`
func (dc *DeterminismContext) bothScoresShouldBeByteIdentical() error {
	b1 := math.Float64bits(dc.score1)
	b2 := math.Float64bits(dc.score2)
	if b1 != b2 {
		return fmt.Errorf(
			"score determinism violated: first=%.17f (0x%x) second=%.17f (0x%x)",
			dc.score1, b1, dc.score2, b2,
		)
	}
	return nil
}

// bothPerAlgorithmScoreMapsShouldContainTheSameKeyValuePairs asserts
// that scoreMap1 and scoreMap2 have the same key set and the same
// per-key bit-identical values. Map iteration order is irrelevant.
// Step regex: `^both per-algorithm score maps should contain the same key-value pairs$`
func (dc *DeterminismContext) bothPerAlgorithmScoreMapsShouldContainTheSameKeyValuePairs() error {
	if len(dc.scoreMap1) != len(dc.scoreMap2) {
		return fmt.Errorf(
			"ScoreAll determinism violated: first map has %d entries, second has %d",
			len(dc.scoreMap1), len(dc.scoreMap2),
		)
	}
	for id, v1 := range dc.scoreMap1 {
		v2, ok := dc.scoreMap2[id]
		if !ok {
			return fmt.Errorf(
				"ScoreAll determinism violated: second map missing AlgoID %s present in first",
				id,
			)
		}
		if math.Float64bits(v1) != math.Float64bits(v2) {
			return fmt.Errorf(
				"ScoreAll determinism violated for AlgoID %s: first=%.17f second=%.17f",
				id, v1, v2,
			)
		}
	}
	return nil
}

// theReturnedSliceShouldBeInAlgoIDAscendingOrder asserts the
// algorithms slice is sorted in AlgoID-ascending order. Strict
// less-than comparison fires on the first out-of-order pair.
// Step regex: `^the returned slice should be in AlgoID-ascending order$`
func (dc *DeterminismContext) theReturnedSliceShouldBeInAlgoIDAscendingOrder() error {
	if len(dc.algorithms) < 2 {
		return nil // vacuously sorted
	}
	for i := 1; i < len(dc.algorithms); i++ {
		if dc.algorithms[i-1].ID >= dc.algorithms[i].ID {
			return fmt.Errorf(
				"Algorithms() not in ascending AlgoID order at index %d: %s (%d) followed by %s (%d)",
				i-1,
				dc.algorithms[i-1].ID, int(dc.algorithms[i-1].ID),
				dc.algorithms[i].ID, int(dc.algorithms[i].ID),
			)
		}
	}
	return nil
}

// InitDeterminismSteps registers all determinism step regexes with
// the supplied godog ScenarioContext. Each scenario gets a fresh
// DeterminismContext (one per scenario, not one per step) keyed off
// the closure-captured `dc` variable.
//
// Called from algorithms_steps.go's InitializeScenario alongside
// InitScorerSteps, InitValidateSteps, and InitNormalisationSteps.
func InitDeterminismSteps(ctx *godog.ScenarioContext) {
	dc := &DeterminismContext{}

	// Given steps — input setup.
	ctx.Step(
		`^the determinism input pair "([^"]*)" and "([^"]*)"$`,
		dc.theDeterminismInputPair,
	)

	// When steps — invocation.
	ctx.Step(
		`^I call DefaultScorer\.Score twice$`,
		dc.iCallDefaultScorerScoreTwice,
	)
	ctx.Step(
		`^I call DefaultScorer\.ScoreAll twice$`,
		dc.iCallDefaultScorerScoreAllTwice,
	)
	ctx.Step(
		`^I call DefaultScorer\.Algorithms$`,
		dc.iCallDefaultScorerAlgorithms,
	)
	ctx.Step(
		`^I call LevenshteinScore twice$`,
		dc.iCallLevenshteinScoreTwice,
	)
	ctx.Step(
		`^I call JaroWinklerScore twice$`,
		dc.iCallJaroWinklerScoreTwice,
	)
	ctx.Step(
		`^I call CosineScore twice$`,
		dc.iCallCosineScoreTwice,
	)

	// Then steps — assertions.
	ctx.Step(
		`^both scores should be byte-identical$`,
		dc.bothScoresShouldBeByteIdentical,
	)
	ctx.Step(
		`^both per-algorithm score maps should contain the same key-value pairs$`,
		dc.bothPerAlgorithmScoreMapsShouldContainTheSameKeyValuePairs,
	)
	ctx.Step(
		`^the returned slice should be in AlgoID-ascending order$`,
		dc.theReturnedSliceShouldBeInAlgoIDAscendingOrder,
	)
}
