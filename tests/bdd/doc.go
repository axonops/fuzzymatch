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

// Package bdd hosts the godog-based behaviour-driven test scenarios for
// fuzzymatch.
//
// This is an isolated Go sub-module (its own go.mod under tests/bdd) so that
// godog, goleak, and testify never appear in the root fuzzymatch module.
// A replace directive points github.com/axonops/fuzzymatch at ../.. so the
// scenarios exercise the local source tree.
//
// Feature files live under tests/bdd/features. Step definitions live under
// tests/bdd/steps. The authoritative testing strategy is documented in
// docs/requirements.md §15.
package bdd

// Blank imports: pin the spec-locked test-only dependencies (godog v0.15.0,
// goleak v1.3.0, testify v1.10.0) and the parent fuzzymatch module at the
// BDD sub-module's freeze point (plan 01-01). Subsequent plans add the actual
// step-definition Go files under tests/bdd/steps that import these
// concretely; until then these blank imports keep the require lines pinned
// and make `go mod tidy` idempotent.
import (
	_ "github.com/axonops/fuzzymatch"
	_ "github.com/cucumber/godog"
	_ "github.com/stretchr/testify/assert"
	_ "go.uber.org/goleak"
)
