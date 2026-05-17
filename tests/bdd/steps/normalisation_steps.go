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

// normalisation_steps.go contains the godog step definitions for the
// Phase 8.5 Plan 17b Gap 3 normalisation.feature scenarios.
//
// The NormalisationContext struct holds state between steps within a
// scenario; each scenario instantiates a fresh NormalisationContext
// via InitNormalisationSteps (called from algorithms_steps.go's
// InitializeScenario).
//
// All normalisation scenarios live under the @normalisation Gherkin
// tag in features/normalisation.feature. Six scenarios cover the
// Normalise contract documented in docs/requirements.md §9.

package steps

import (
	"fmt"

	"github.com/cucumber/godog"

	"github.com/axonops/fuzzymatch"
)

// NormalisationContext holds state between BDD steps within a single
// normalisation scenario. inputA is the primary normalisation input;
// inputB is used only by the precomposed/decomposed equivalence
// scenario. outputA / outputB hold the corresponding Normalise
// results.
type NormalisationContext struct {
	inputA, inputB   string
	outputA, outputB string
}

// theNormalisationInput sets inputA from the Gherkin step.
// Step regex: `^the normalisation input "([^"]*)"$`
func (nc *NormalisationContext) theNormalisationInput(s string) error {
	nc.inputA = s
	nc.inputB = ""
	nc.outputA = ""
	nc.outputB = ""
	return nil
}

// thePrecomposedAndDecomposedInputs sets inputA (precomposed) and
// inputB (decomposed) for the NFC-equivalence scenario.
// Step regex: `^the precomposed input "([^"]*)" and the decomposed input "([^"]*)"$`
func (nc *NormalisationContext) thePrecomposedAndDecomposedInputs(a, b string) error {
	nc.inputA = a
	nc.inputB = b
	nc.outputA = ""
	nc.outputB = ""
	return nil
}

// iNormaliseUsingDefaultNormalisationOptions runs
// fuzzymatch.Normalise(inputA, DefaultNormalisationOptions()) and
// stores the result in outputA. If inputB is set, normalises it too
// into outputB.
// Step regex: `^I normalise using DefaultNormalisationOptions$`
func (nc *NormalisationContext) iNormaliseUsingDefaultNormalisationOptions() error {
	opts := fuzzymatch.DefaultNormalisationOptions()
	nc.outputA = fuzzymatch.Normalise(nc.inputA, opts)
	return nil
}

// iNormaliseBothWithDefaultNormalisationOptions normalises both
// inputA and inputB with DefaultNormalisationOptions; used by the
// precomposed/decomposed NFC-equivalence scenario.
// Step regex: `^I normalise both with DefaultNormalisationOptions$`
func (nc *NormalisationContext) iNormaliseBothWithDefaultNormalisationOptions() error {
	opts := fuzzymatch.DefaultNormalisationOptions()
	nc.outputA = fuzzymatch.Normalise(nc.inputA, opts)
	nc.outputB = fuzzymatch.Normalise(nc.inputB, opts)
	return nil
}

// iNormaliseUsingDefaultNormalisationOptionsWithStripDiacritics runs
// fuzzymatch.Normalise with StripDiacritics: true and stores the
// result in outputA.
// Step regex: `^I normalise using DefaultNormalisationOptions with StripDiacritics$`
func (nc *NormalisationContext) iNormaliseUsingDefaultNormalisationOptionsWithStripDiacritics() error {
	opts := fuzzymatch.DefaultNormalisationOptions()
	opts.StripDiacritics = true
	nc.outputA = fuzzymatch.Normalise(nc.inputA, opts)
	return nil
}

// iNormaliseTheOutputAgainUsingDefaultNormalisationOptions runs
// Normalise on outputA and stores the result in outputB; used by the
// idempotence scenario.
// Step regex: `^I normalise the output again using DefaultNormalisationOptions$`
func (nc *NormalisationContext) iNormaliseTheOutputAgainUsingDefaultNormalisationOptions() error {
	opts := fuzzymatch.DefaultNormalisationOptions()
	nc.outputB = fuzzymatch.Normalise(nc.outputA, opts)
	return nil
}

// theNormalisedOutputShouldBe asserts outputA equals the expected
// string.
// Step regex: `^the normalised output should be "([^"]*)"$`
func (nc *NormalisationContext) theNormalisedOutputShouldBe(expected string) error {
	if nc.outputA != expected {
		return fmt.Errorf(
			"normalised output mismatch: got %q, want %q (input: %q)",
			nc.outputA, expected, nc.inputA,
		)
	}
	return nil
}

// theTwoOutputsShouldBeByteIdentical asserts outputA == outputB.
// Step regex: `^the two outputs should be byte-identical$`
func (nc *NormalisationContext) theTwoOutputsShouldBeByteIdentical() error {
	if nc.outputA != nc.outputB {
		return fmt.Errorf(
			"outputs differ:\n  outputA: %q (from input %q)\n  outputB: %q (from input %q)",
			nc.outputA, nc.inputA, nc.outputB, nc.inputB,
		)
	}
	return nil
}

// InitNormalisationSteps registers all normalisation step regexes
// with the supplied godog ScenarioContext. Each scenario gets a fresh
// NormalisationContext (one per scenario, not one per step) keyed off
// the closure-captured `nc` variable.
//
// Called from algorithms_steps.go's InitializeScenario alongside
// InitScorerSteps and InitValidateSteps.
func InitNormalisationSteps(ctx *godog.ScenarioContext) {
	nc := &NormalisationContext{}

	// Given steps — input setup.
	ctx.Step(
		`^the normalisation input "([^"]*)"$`,
		nc.theNormalisationInput,
	)
	ctx.Step(
		`^the precomposed input "([^"]*)" and the decomposed input "([^"]*)"$`,
		nc.thePrecomposedAndDecomposedInputs,
	)

	// When steps — Normalise invocation.
	ctx.Step(
		`^I normalise using DefaultNormalisationOptions$`,
		nc.iNormaliseUsingDefaultNormalisationOptions,
	)
	ctx.Step(
		`^I normalise using DefaultNormalisationOptions with StripDiacritics$`,
		nc.iNormaliseUsingDefaultNormalisationOptionsWithStripDiacritics,
	)
	ctx.Step(
		`^I normalise both with DefaultNormalisationOptions$`,
		nc.iNormaliseBothWithDefaultNormalisationOptions,
	)
	ctx.Step(
		`^I normalise the output again using DefaultNormalisationOptions$`,
		nc.iNormaliseTheOutputAgainUsingDefaultNormalisationOptions,
	)

	// Then steps — assertions.
	ctx.Step(
		`^the normalised output should be "([^"]*)"$`,
		nc.theNormalisedOutputShouldBe,
	)
	ctx.Step(
		`^the two outputs should be byte-identical$`,
		nc.theTwoOutputsShouldBeByteIdentical,
	)
}
