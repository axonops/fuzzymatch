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

// example_test.go contains runnable godoc examples for the Phase 2
// character-based algorithms. Each ExampleXxx function appears on
// pkg.go.dev alongside the function it documents. Wave 2 plans append their
// ExampleXxx functions to this SAME file.
//
// The Output: blocks are verified byte-for-byte by `go test -run ExampleXxx
// ./...` — any drift in score computation causes a test failure.

package fuzzymatch_test

import (
	"fmt"

	"github.com/axonops/fuzzymatch"
)

// ExampleLevenshteinScore demonstrates the Levenshtein similarity on the
// canonical Wagner-Fischer 1974 reference pair. The score is
// 1 - 3/7 ≈ 0.5714 (distance 3, max length 7).
func ExampleLevenshteinScore() {
	fmt.Printf("%.4f\n", fuzzymatch.LevenshteinScore("kitten", "sitting"))
	// Output:
	// 0.5714
}
