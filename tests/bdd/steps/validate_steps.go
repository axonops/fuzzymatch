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

// validate_steps.go contains the godog step definitions for the
// Phase 8.5 Plan 13 Validate surface. The ValidateContext struct holds
// state between steps within a scenario; each scenario instantiates a
// fresh ValidateContext via InitValidateSteps (called from
// algorithms_steps.go's InitializeScenario).
//
// All Validate scenarios live under the @validate Gherkin tag in
// features/validate.feature. The five WarnKind constants each get a
// scenario; plus a clean-input scenario (nil-return) and a determinism
// scenario (two-call byte-equality, T-08.5-26 mitigation gate).
//
// testify is permitted in tests/bdd/ but is not required here — the
// helper assertions are plain returns of fmt.Errorf, matching the
// pattern in algorithms_steps.go and scorer_steps.go.

package steps

import (
	"fmt"
	"reflect"
	"strings"

	"github.com/cucumber/godog"

	"github.com/axonops/fuzzymatch"
)

// ValidateContext holds state between BDD steps within a Validate
// scenario. Each scenario instantiates a fresh ValidateContext.
//
// inputA / inputB hold the two strings under test. warnings is the
// result of the first Validate call; secondWarnings is the result of
// the second call (used by the determinism scenario).
type ValidateContext struct {
	inputA, inputB string
	warnings       []fuzzymatch.Warning
	secondWarnings []fuzzymatch.Warning
}

// theValidationInputPair sets inputA and inputB from the Gherkin step.
// Step regex: `^the validation input pair "([^"]*)" and "([^"]*)"$`
func (vc *ValidateContext) theValidationInputPair(a, b string) error {
	vc.inputA = a
	vc.inputB = b
	vc.warnings = nil
	vc.secondWarnings = nil
	return nil
}

// aValidationInputPairOfLengthEach constructs two strings of the
// requested byte length (using 'a' as filler) and stores them in the
// context. Used by the WarnPathologicallyLargeInput scenario.
// Step regex: `^a validation input pair of length (\d+) each$`
func (vc *ValidateContext) aValidationInputPairOfLengthEach(n int) error {
	if n < 0 {
		return fmt.Errorf("invalid length %d", n)
	}
	vc.inputA = strings.Repeat("a", n)
	vc.inputB = strings.Repeat("b", n)
	vc.warnings = nil
	vc.secondWarnings = nil
	return nil
}

// iCallValidate runs Validate against the context's input pair and
// stores the result in warnings.
// Step regex: `^I call Validate$`
func (vc *ValidateContext) iCallValidate() error {
	vc.warnings = fuzzymatch.Validate(vc.inputA, vc.inputB)
	return nil
}

// iCallValidateTwice runs Validate against the context's input pair
// twice and stores both results.
// Step regex: `^I call Validate twice$`
func (vc *ValidateContext) iCallValidateTwice() error {
	vc.warnings = fuzzymatch.Validate(vc.inputA, vc.inputB)
	vc.secondWarnings = fuzzymatch.Validate(vc.inputA, vc.inputB)
	return nil
}

// theWarningsIncludeKind asserts that at least one Warning in the
// stored warnings has the named WarnKind.
// Step regex: `^the warnings include (Warn[A-Za-z]+)$`
func (vc *ValidateContext) theWarningsIncludeKind(kindName string) error {
	want, err := parseWarnKind(kindName)
	if err != nil {
		return err
	}
	for _, w := range vc.warnings {
		if w.Kind == want {
			return nil
		}
	}
	return fmt.Errorf(
		"warnings do not include %s; got %d warnings: %s",
		kindName, len(vc.warnings), renderWarnings(vc.warnings),
	)
}

// theWarningsIncludeKindScopedTo asserts that at least one Warning has
// both the named WarnKind AND is scoped to the named AlgoID (matching
// AlgoID.String()).
// Step regex: `^the warnings include (Warn[A-Za-z]+) scoped to ([A-Za-z]+)$`
func (vc *ValidateContext) theWarningsIncludeKindScopedTo(kindName, algoName string) error {
	wantKind, err := parseWarnKind(kindName)
	if err != nil {
		return err
	}
	wantAlgo, err := parseAlgoID(algoName)
	if err != nil {
		return err
	}
	for _, w := range vc.warnings {
		if w.Kind == wantKind && w.Algorithm == wantAlgo {
			return nil
		}
	}
	return fmt.Errorf(
		"warnings do not include {%s, %s}; got %d warnings: %s",
		algoName, kindName, len(vc.warnings), renderWarnings(vc.warnings),
	)
}

// theWarningsShouldBeNil asserts the stored warnings slice is nil
// (not an empty []Warning{}). This is the VALIDATE-01 nil-vs-empty
// contract.
// Step regex: `^the warnings should be nil$`
func (vc *ValidateContext) theWarningsShouldBeNil() error {
	if vc.warnings != nil {
		return fmt.Errorf(
			"expected nil warnings; got %d: %s",
			len(vc.warnings), renderWarnings(vc.warnings),
		)
	}
	return nil
}

// bothWarningsSlicesShouldBeIdentical asserts the two stored result
// slices are DeepEqual — the determinism contract (T-08.5-26).
// Step regex: `^both warnings slices should be identical$`
func (vc *ValidateContext) bothWarningsSlicesShouldBeIdentical() error {
	if !reflect.DeepEqual(vc.warnings, vc.secondWarnings) {
		return fmt.Errorf(
			"warnings differ between calls: first=%s second=%s",
			renderWarnings(vc.warnings), renderWarnings(vc.secondWarnings),
		)
	}
	return nil
}

// parseWarnKind maps a string label ("WarnEmptyInput", etc.) to the
// corresponding fuzzymatch.WarnKind constant. Accepts both the
// "Warn"-prefixed constant name and the canonical String() form.
func parseWarnKind(s string) (fuzzymatch.WarnKind, error) {
	switch s {
	case "WarnEmptyInput", "EmptyInput":
		return fuzzymatch.WarnEmptyInput, nil
	case "WarnUnequalLength", "UnequalLength":
		return fuzzymatch.WarnUnequalLength, nil
	case "WarnNoTokensAfterNormalise", "NoTokensAfterNormalise":
		return fuzzymatch.WarnNoTokensAfterNormalise, nil
	case "WarnAllNonASCIIDropped", "AllNonASCIIDropped":
		return fuzzymatch.WarnAllNonASCIIDropped, nil
	case "WarnPathologicallyLargeInput", "PathologicallyLargeInput":
		return fuzzymatch.WarnPathologicallyLargeInput, nil
	default:
		return 0, fmt.Errorf("unknown WarnKind label: %q", s)
	}
}

// parseAlgoID maps a string label (the AlgoID.String() canonical form)
// to the corresponding fuzzymatch.AlgoID constant. Includes the
// AlgoIDAny sentinel ("Any") for cross-cutting warning assertions.
func parseAlgoID(s string) (fuzzymatch.AlgoID, error) {
	for _, id := range fuzzymatch.AlgoIDs() {
		if id.String() == s {
			return id, nil
		}
	}
	if s == "Any" {
		return fuzzymatch.AlgoIDAny, nil
	}
	return 0, fmt.Errorf("unknown AlgoID label: %q", s)
}

// renderWarnings produces a stable, human-readable rendering of a
// warnings slice for error messages. Sorted-by-construction so the
// output is deterministic regardless of map iteration.
func renderWarnings(ws []fuzzymatch.Warning) string {
	if len(ws) == 0 {
		return "<empty>"
	}
	var b strings.Builder
	b.WriteString("[")
	for i, w := range ws {
		if i > 0 {
			b.WriteString(", ")
		}
		fmt.Fprintf(&b, "{%s, %s}", w.Algorithm, w.Kind)
	}
	b.WriteString("]")
	return b.String()
}

// InitValidateSteps registers all Validate step regexes with the
// supplied godog ScenarioContext. Each scenario gets a fresh
// ValidateContext (one per scenario, not one per step) keyed off the
// closure-captured `vc` variable.
//
// Called from algorithms_steps.go's InitializeScenario at the bottom
// of the registration block, alongside InitScorerSteps.
func InitValidateSteps(ctx *godog.ScenarioContext) {
	vc := &ValidateContext{}

	// Given steps — input setup.
	ctx.Step(
		`^the validation input pair "([^"]*)" and "([^"]*)"$`,
		vc.theValidationInputPair,
	)
	ctx.Step(
		`^a validation input pair of length (\d+) each$`,
		vc.aValidationInputPairOfLengthEach,
	)

	// When steps — Validate invocation.
	ctx.Step(
		`^I call Validate$`,
		vc.iCallValidate,
	)
	ctx.Step(
		`^I call Validate twice$`,
		vc.iCallValidateTwice,
	)

	// Then steps — assertions.
	ctx.Step(
		`^the warnings include (Warn[A-Za-z]+)$`,
		vc.theWarningsIncludeKind,
	)
	ctx.Step(
		`^the warnings include (Warn[A-Za-z]+) scoped to ([A-Za-z]+)$`,
		vc.theWarningsIncludeKindScopedTo,
	)
	ctx.Step(
		`^the warnings should be nil$`,
		vc.theWarningsShouldBeNil,
	)
	ctx.Step(
		`^both warnings slices should be identical$`,
		vc.bothWarningsSlicesShouldBeIdentical,
	)
}
