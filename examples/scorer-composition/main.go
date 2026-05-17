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

// Package main demonstrates Phase 8 Scorer composition patterns:
//
//   - DefaultScorer — the opinionated six-algorithm composition with
//     baked-in 0.85 threshold; one-line construction for casual users.
//   - DefaultScorerOptions + WithoutAlgorithm(AlgoDoubleMetaphone) +
//     WithThreshold(0.80) — the canonical "default minus phonetic"
//     use case for numeric-identifier data where phonetic similarity
//     between digits like "user_123" and "user_124" is misleading.
//
// The example prints both Scorers' composite scores and match decisions
// side-by-side for several input pairs so the cost of removing one
// algorithm is visible: pairs that score high under DefaultScorer often
// score lower under DefaultScorer-minus-DoubleMetaphone, but the gap is
// small for non-phonetic data and large for phonetic data
// (Smith / Schmidt). The threshold drops from 0.85 → 0.80 because one
// signal is removed and the consumer compensates.
//
// Run with:
//
//	go run ./examples/scorer-composition/
package main

import (
	"fmt"

	"github.com/axonops/fuzzymatch"
)

// pairs is the ordered list of input pairs used to exercise both
// Scorers. The pairs are chosen to highlight the difference between
// "full default Scorer" and "default minus phonetic":
//
//   - user_id / userId       — identifier-style, mostly orthographic
//     similarity; phonetic match is incidental.
//   - Smith / Schmidt        — classic phonetic-but-not-orthographic
//     pair (DoubleMetaphone returns 1.0; edit distance is high).
//   - customer / customers   — pure suffix-stripping case.
//   - XMLParser / xml_parser — case + separator drift; Normalise
//     erases the difference.
//   - org_id / organisation_id — abbreviation expansion; the longer
//     side dominates Levenshtein but token-similarity catches it.
var pairs = [][2]string{
	{"user_id", "userId"},
	{"Smith", "Schmidt"},
	{"customer", "customers"},
	{"XMLParser", "xml_parser"},
	{"org_id", "organisation_id"},
}

// defaultS is the opinionated six-algorithm Phase 8 default Scorer.
// Constructed once at package init time; immutable; safe for concurrent
// use. Used by both the "Default" column of the rendered table and
// as the comparison baseline for the customised "MinusDM" Scorer below.
var defaultS = fuzzymatch.DefaultScorer()

// minusDMS is the "default minus DoubleMetaphone" Scorer constructed
// via `append(DefaultScorerOptions(), WithoutAlgorithm(AlgoDoubleMetaphone),
// WithThreshold(0.80))`. The threshold drops from 0.85 → 0.80 because
// removing one of six signals reduces the composite ceiling; the lower
// threshold compensates.
//
// The error from NewScorer is intentionally ignored because the
// composition is statically valid: DefaultScorerOptions() is known to
// produce a valid 6-algorithm base, WithoutAlgorithm is a no-op
// silencer (it cannot fail), and WithThreshold(0.80) is in the [0, 1]
// range that scorer.go validates. Production code SHOULD check the
// error — this is example code.
var minusDMS = newMinusDMScorer()

// newMinusDMScorer builds the "default minus DoubleMetaphone" Scorer.
// Extracted into a function so the var initializer above stays a
// single-line expression and so the error-handling pattern (panic on
// programmer error, since the composition is statically valid) lives
// in one named place that future readers can grep for.
func newMinusDMScorer() *fuzzymatch.Scorer {
	opts := append(fuzzymatch.DefaultScorerOptions(),
		fuzzymatch.WithoutAlgorithm(fuzzymatch.AlgoDoubleMetaphone),
		fuzzymatch.WithThreshold(0.80),
	)
	s, err := fuzzymatch.NewScorer(opts...)
	if err != nil {
		// This is unreachable because the composition is statically
		// valid (DefaultScorerOptions + remove one algorithm + set
		// threshold). Panic so a programmer error during refactor is
		// surfaced immediately rather than silently producing a nil
		// Scorer.
		panic(fmt.Sprintf("scorer-composition example: NewScorer failed unexpectedly: %v", err))
	}
	return s
}

func main() {
	// Column widths: 32 chars for the pair column, 12 chars for each
	// of the four data columns (Default, MinusDM, MDef, MMDM). Header
	// + separator + data rows all use the same widths.
	const pairWidth = 32
	const dataWidth = 12

	// Header row.
	fmt.Printf("%-*s", pairWidth, "Pair (a / b)")
	for _, name := range []string{"Default", "MinusDM", "MDef", "MMDM"} {
		fmt.Printf("%*s", dataWidth, name)
	}
	fmt.Println()

	// Separator line.
	for i := 0; i < pairWidth+4*dataWidth; i++ {
		fmt.Print("-")
	}
	fmt.Println()

	// Data rows.
	for _, p := range pairs {
		a, b := p[0], p[1]
		label := fmt.Sprintf("%s / %s", a, b)
		fmt.Printf("%-*s", pairWidth, label)
		fmt.Printf("%*s", dataWidth, fmt.Sprintf("%.4f", defaultS.Score(a, b)))
		fmt.Printf("%*s", dataWidth, fmt.Sprintf("%.4f", minusDMS.Score(a, b)))
		fmt.Printf("%*s", dataWidth, fmt.Sprintf("%t", defaultS.Match(a, b)))
		fmt.Printf("%*s", dataWidth, fmt.Sprintf("%t", minusDMS.Match(a, b)))
		fmt.Println()
	}
}
