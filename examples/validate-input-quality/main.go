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

// Package main demonstrates the fuzzymatch.Validate input-quality
// diagnostic surface (Phase 8.5 Q4 / VALIDATE-06 surface 6).
//
// fuzzymatch.Validate(a, b) inspects two input strings and returns a
// deterministically-sorted slice of Warning values describing
// problematic-but-non-fatal input shapes. It is pure, never panics,
// and returns nil (not an empty slice) when no warning applies.
//
// The example walks four representative input pairs, prints any
// warnings each pair triggers, and then prints the DefaultScorer's
// score for the same pair — showing that the algorithms always
// produce a value (the lenient comparison-data contract) and the
// Validate surface is the optional companion that says whether the
// value will be meaningful.
//
// Pairs chosen to exercise distinct WarnKind values:
//
//   - "user_id" / ""        — WarnEmptyInput (cross-cutting) plus
//     WarnUnequalLength (Hamming) plus
//     WarnNoTokensAfterNormalise (5 token-tier algos)
//   - "abc" / "abcd"        — WarnUnequalLength scoped to AlgoHamming
//   - "中文" / "日本語"        — WarnUnequalLength (Hamming) plus
//     WarnAllNonASCIIDropped (5 ASCII-only algos)
//   - "user_id" / "userId"  — WarnUnequalLength only (Hamming silent-max)
//
// A clean nil-return scenario (no warnings emitted) requires two
// equal-length ASCII inputs that both tokenise to non-empty lists —
// e.g. fuzzymatch.Validate("hello", "world"). The exercise pairs
// above are chosen to surface the most common diagnostic shapes.
//
// Run with:
//
//	go run ./examples/validate-input-quality/
package main

import (
	"fmt"

	"github.com/axonops/fuzzymatch"
)

// pairs is the ordered list of input pairs the example exercises.
// Each pair is chosen to trigger a distinct WarnKind (or none, to
// demonstrate the nil-return contract).
var pairs = [][2]string{
	{"user_id", ""},
	{"abc", "abcd"},
	{"中文", "日本語"},
	{"user_id", "userId"},
}

// defaultS is the opinionated six-algorithm DefaultScorer used to
// show the score alongside the warnings. Constructed once at package
// init time; immutable; safe for concurrent use.
var defaultS = fuzzymatch.DefaultScorer()

func main() {
	for _, p := range pairs {
		a, b := p[0], p[1]
		fmt.Printf("Pair: %q / %q\n", a, b)

		warnings := fuzzymatch.Validate(a, b)
		if warnings == nil {
			fmt.Println("  warnings: <none>")
		} else {
			fmt.Printf("  warnings: %d\n", len(warnings))
			for _, w := range warnings {
				fmt.Printf("    - %s (%s)\n", w.Kind, w.Algorithm)
			}
		}

		score := defaultS.Score(a, b)
		fmt.Printf("  score:    %.4f\n", score)
		fmt.Println()
	}
}
