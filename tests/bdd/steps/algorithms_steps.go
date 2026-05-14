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

// Package steps contains the godog step definitions for Phase 2's six
// character-based algorithms. The AlgorithmContext struct holds state between
// steps within a scenario. Step functions return error rather than calling
// t.Errorf, as required by the godog API.
//
// Wave 2 plans (02-02 through 02-06) extend this file by:
//   - Adding their algorithm-specific step functions to AlgorithmContext.
//   - Registering those step regexes in InitializeScenario.
//   - Adding any additional state fields to AlgorithmContext.
//
// testify is available in tests/bdd/ (isolated sub-module) but is not required
// for this initial harness; step functions return error directly.

package steps

import (
	"fmt"
	"math"

	"github.com/cucumber/godog"

	"github.com/axonops/fuzzymatch"
)

// AlgorithmContext holds state between BDD steps within a scenario. Each
// scenario instantiates a fresh AlgorithmContext to ensure isolation.
//
// Wave 2 plans may add additional state fields if their step patterns require
// more than two scores. The two-score design covers the symmetry scenario
// ("compute A and compute B, then assert both equal").
type AlgorithmContext struct {
	lastScore  float64 // populated by "I compute the Xxx score between" steps
	lastScore2 float64 // populated by "I compute the second Xxx score between" steps
}

// iComputeTheLevenshteinScoreBetween computes LevenshteinScore(a, b) and
// stores the result in lastScore.
func (ctx *AlgorithmContext) iComputeTheLevenshteinScoreBetween(a, b string) error {
	ctx.lastScore = fuzzymatch.LevenshteinScore(a, b)
	return nil
}

// iComputeTheSecondLevenshteinScoreBetween computes LevenshteinScore(a, b) and
// stores the result in lastScore2. Used by the symmetry scenario to capture
// a second score for comparison.
func (ctx *AlgorithmContext) iComputeTheSecondLevenshteinScoreBetween(a, b string) error {
	ctx.lastScore2 = fuzzymatch.LevenshteinScore(a, b)
	return nil
}

// theScoreShouldBeApproximately asserts that lastScore is within tolerance of
// expected. Uses absolute difference (math.Abs) to avoid sign-sensitive
// comparisons.
func (ctx *AlgorithmContext) theScoreShouldBeApproximately(expected, tolerance float64) error {
	if math.Abs(ctx.lastScore-expected) > tolerance {
		return fmt.Errorf("expected score %f ± %f, got %f", expected, tolerance, ctx.lastScore)
	}
	return nil
}

// theScoreShouldBeExactly asserts that lastScore equals expected exactly.
// Used for edge cases (both-empty = 1.0, one-empty = 0.0) where the result
// must be the exact IEEE-754 value without floating-point rounding.
func (ctx *AlgorithmContext) theScoreShouldBeExactly(expected float64) error {
	if ctx.lastScore != expected {
		return fmt.Errorf("expected score exactly %f, got %f", expected, ctx.lastScore)
	}
	return nil
}

// bothLevenshteinScoresShouldBeEqual asserts that lastScore == lastScore2. Used
// by the symmetry scenario after computing Score(a,b) and Score(b,a).
func (ctx *AlgorithmContext) bothLevenshteinScoresShouldBeEqual() error {
	if ctx.lastScore != ctx.lastScore2 {
		return fmt.Errorf("scores not equal: %f != %f", ctx.lastScore, ctx.lastScore2)
	}
	return nil
}

// InitializeScenario wires step definitions into the godog suite. Each call
// creates a fresh AlgorithmContext bound to the scenario, ensuring per-scenario
// isolation. Wave 2 plans append their algorithm's step regexes here.
//
// Step regexes use the godog-standard pattern: literal text with capture
// groups for variable parts. String captures use `([^"]*)` to exclude the
// surrounding quotes; numeric captures use `(\d+\.\d+)`.
func InitializeScenario(ctx *godog.ScenarioContext) {
	a := &AlgorithmContext{}

	// Levenshtein step definitions.
	ctx.Step(
		`^I compute the Levenshtein score between "([^"]*)" and "([^"]*)"$`,
		a.iComputeTheLevenshteinScoreBetween,
	)
	ctx.Step(
		`^I compute the second Levenshtein score between "([^"]*)" and "([^"]*)"$`,
		a.iComputeTheSecondLevenshteinScoreBetween,
	)
	ctx.Step(
		`^the score should be approximately (\d+\.\d+) within (\d+\.\d+)$`,
		a.theScoreShouldBeApproximately,
	)
	ctx.Step(
		`^the score should be exactly (\d+\.\d+)$`,
		a.theScoreShouldBeExactly,
	)
	ctx.Step(
		`^both Levenshtein scores should be equal$`,
		a.bothLevenshteinScoresShouldBeEqual,
	)
}
