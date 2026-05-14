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
	lastScore    float64 // populated by "I compute the Xxx score between" steps
	lastScore2   float64 // populated by "I compute the second Xxx score between" steps
	lastDistance int     // populated by "I compute the Xxx distance between" steps (plan 02-02+)
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

// ---------------------------------------------------------------------------
// Hamming step definitions (plan 02-02)
// ---------------------------------------------------------------------------

// iComputeTheHammingScoreBetween computes HammingScore(a, b) and stores the
// result in lastScore. Demonstrates the locked unequal-length silent-zero
// policy: Score("abc", "ab") == 0.0 silently.
func (ctx *AlgorithmContext) iComputeTheHammingScoreBetween(a, b string) error {
	ctx.lastScore = fuzzymatch.HammingScore(a, b)
	return nil
}

// iComputeTheSecondHammingScoreBetween computes HammingScore(a, b) and stores
// the result in lastScore2. Used by the symmetry scenario.
func (ctx *AlgorithmContext) iComputeTheSecondHammingScoreBetween(a, b string) error {
	ctx.lastScore2 = fuzzymatch.HammingScore(a, b)
	return nil
}

// iComputeTheHammingDistanceBetween computes HammingDistance(a, b) and stores
// the result in lastDistance. Used by the distance-equals-max-length scenario.
func (ctx *AlgorithmContext) iComputeTheHammingDistanceBetween(a, b string) error {
	ctx.lastDistance = fuzzymatch.HammingDistance(a, b)
	return nil
}

// bothHammingScoresShouldBeEqual asserts lastScore == lastScore2.
func (ctx *AlgorithmContext) bothHammingScoresShouldBeEqual() error {
	if ctx.lastScore != ctx.lastScore2 {
		return fmt.Errorf("hamming scores not equal: %f != %f", ctx.lastScore, ctx.lastScore2)
	}
	return nil
}

// theDistanceShouldBe asserts lastDistance == expected. Used by the
// unequal-length distance-equals-max-length contract scenario.
func (ctx *AlgorithmContext) theDistanceShouldBe(expected int) error {
	if ctx.lastDistance != expected {
		return fmt.Errorf("expected distance %d, got %d", expected, ctx.lastDistance)
	}
	return nil
}

// ---------------------------------------------------------------------------
// Jaro step definitions (plan 02-03)
// ---------------------------------------------------------------------------

// iComputeTheJaroScoreBetween computes JaroScore(a, b) and stores the result
// in lastScore. The Jaro formula is symmetric and operates on bytes for ASCII
// inputs (zero allocation on inputs <= 256 bytes).
func (ctx *AlgorithmContext) iComputeTheJaroScoreBetween(a, b string) error {
	ctx.lastScore = fuzzymatch.JaroScore(a, b)
	return nil
}

// iComputeTheSecondJaroScoreBetween computes JaroScore(a, b) and stores the
// result in lastScore2. Used by the symmetry scenario to capture a second score
// for comparison.
func (ctx *AlgorithmContext) iComputeTheSecondJaroScoreBetween(a, b string) error {
	ctx.lastScore2 = fuzzymatch.JaroScore(a, b)
	return nil
}

// bothJaroScoresShouldBeEqual asserts lastScore == lastScore2.
func (ctx *AlgorithmContext) bothJaroScoresShouldBeEqual() error {
	if ctx.lastScore != ctx.lastScore2 {
		return fmt.Errorf("jaro scores not equal: %f != %f", ctx.lastScore, ctx.lastScore2)
	}
	return nil
}

// ---------------------------------------------------------------------------
// Damerau-Levenshtein OSA step definitions (plan 02-05)
// ---------------------------------------------------------------------------

// iComputeTheDamerauLevenshteinOSAScoreBetween computes
// DamerauLevenshteinOSAScore(a, b) and stores the result in lastScore.
// The ASCII fast path uses a stack-allocated three-row DP buffer for zero
// heap allocations on inputs ≤ 64 bytes.
func (ctx *AlgorithmContext) iComputeTheDamerauLevenshteinOSAScoreBetween(a, b string) error {
	ctx.lastScore = fuzzymatch.DamerauLevenshteinOSAScore(a, b)
	return nil
}

// iComputeTheSecondDamerauLevenshteinOSAScoreBetween computes
// DamerauLevenshteinOSAScore(a, b) and stores the result in lastScore2.
// Used by the symmetry scenario to capture a second score for comparison.
func (ctx *AlgorithmContext) iComputeTheSecondDamerauLevenshteinOSAScoreBetween(a, b string) error {
	ctx.lastScore2 = fuzzymatch.DamerauLevenshteinOSAScore(a, b)
	return nil
}

// iComputeTheDamerauLevenshteinOSADistanceBetween computes
// DamerauLevenshteinOSADistance(a, b) and stores the result in lastDistance.
// Used by the discriminating-vector scenario to gate the distance == 3 contract.
func (ctx *AlgorithmContext) iComputeTheDamerauLevenshteinOSADistanceBetween(a, b string) error {
	ctx.lastDistance = fuzzymatch.DamerauLevenshteinOSADistance(a, b)
	return nil
}

// bothDamerauLevenshteinOSAScoresShouldBeEqual asserts lastScore == lastScore2.
func (ctx *AlgorithmContext) bothDamerauLevenshteinOSAScoresShouldBeEqual() error {
	if ctx.lastScore != ctx.lastScore2 {
		return fmt.Errorf("damerau OSA scores not equal: %f != %f", ctx.lastScore, ctx.lastScore2)
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

	// Hamming step definitions (plan 02-02).
	ctx.Step(
		`^I compute the Hamming score between "([^"]*)" and "([^"]*)"$`,
		a.iComputeTheHammingScoreBetween,
	)
	ctx.Step(
		`^I compute the second Hamming score between "([^"]*)" and "([^"]*)"$`,
		a.iComputeTheSecondHammingScoreBetween,
	)
	ctx.Step(
		`^I compute the Hamming distance between "([^"]*)" and "([^"]*)"$`,
		a.iComputeTheHammingDistanceBetween,
	)
	ctx.Step(
		`^both Hamming scores should be equal$`,
		a.bothHammingScoresShouldBeEqual,
	)
	ctx.Step(
		`^the distance should be (\d+)$`,
		a.theDistanceShouldBe,
	)

	// Jaro step definitions (plan 02-03).
	ctx.Step(
		`^I compute the Jaro score between "([^"]*)" and "([^"]*)"$`,
		a.iComputeTheJaroScoreBetween,
	)
	ctx.Step(
		`^I compute the second Jaro score between "([^"]*)" and "([^"]*)"$`,
		a.iComputeTheSecondJaroScoreBetween,
	)
	ctx.Step(
		`^both Jaro scores should be equal$`,
		a.bothJaroScoresShouldBeEqual,
	)

	// Damerau-Levenshtein OSA step definitions (plan 02-05).
	ctx.Step(
		`^I compute the DamerauLevenshteinOSA score between "([^"]*)" and "([^"]*)"$`,
		a.iComputeTheDamerauLevenshteinOSAScoreBetween,
	)
	ctx.Step(
		`^I compute the second DamerauLevenshteinOSA score between "([^"]*)" and "([^"]*)"$`,
		a.iComputeTheSecondDamerauLevenshteinOSAScoreBetween,
	)
	ctx.Step(
		`^I compute the DamerauLevenshteinOSA distance between "([^"]*)" and "([^"]*)"$`,
		a.iComputeTheDamerauLevenshteinOSADistanceBetween,
	)
	ctx.Step(
		`^both DamerauLevenshteinOSA scores should be equal$`,
		a.bothDamerauLevenshteinOSAScoresShouldBeEqual,
	)
}
