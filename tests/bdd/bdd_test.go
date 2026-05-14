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

// bdd_test.go is the godog test runner for the fuzzymatch BDD suite.
// It wires godog's TestMain to goleak's goroutine-leak detector and registers
// all scenario step definitions via the steps package.
//
// Feature files are discovered from tests/bdd/features/*.feature. Adding a
// new feature file is sufficient to register new scenarios — no changes to
// this runner are required.

package bdd_test

import (
	"testing"

	"github.com/cucumber/godog"
	"go.uber.org/goleak"

	"github.com/axonops/fuzzymatch/tests/bdd/steps"
)

// TestMain integrates goleak with godog's test runner.
// goleak.VerifyTestMain catches goroutine leaks introduced by any scenario;
// this is the canonical BDD harness pattern per the test-writer skill.
func TestMain(m *testing.M) {
	goleak.VerifyTestMain(m)
}

// TestBDDSuite runs all Gherkin scenarios discovered under tests/bdd/features/.
// Each scenario gets a fresh AlgorithmContext via InitializeScenario.
func TestBDDSuite(t *testing.T) {
	suite := godog.TestSuite{
		Name:                "fuzzymatch",
		ScenarioInitializer: steps.InitializeScenario,
		Options: &godog.Options{
			Format:   "pretty",
			Paths:    []string{"features"},
			TestingT: t,
		},
	}
	if suite.Run() != 0 {
		t.Fatal("BDD suite failed — see godog output above for failing scenarios")
	}
}
