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

// Package main demonstrates all fourteen Phase 2 + 3 + 4 + 5 character-based,
// gestalt, and q-gram similarity algorithms from github.com/axonops/fuzzymatch
// side-by-side on database column-name identifier pairs.
//
// The example was designed for axonops/audit, the primary downstream
// consumer of fuzzymatch, where "semantic equivalence detection" across
// Cassandra table schemas requires comparing column names that may differ
// in naming convention (snake_case vs camelCase), casing, separators,
// or abbreviation style.
//
// Each row in the printed table represents a pair of database identifiers;
// each column represents one of the fourteen algorithms (Phase 2 six +
// Smith-Waterman-Gotoh + Strcmp95 + LCSStr + Ratcliff-Obershelp + the four
// Phase 5 q-gram algorithms QGramJ / Dice / Cos / Tversky at default n=3).
// Cell values are similarity scores in [0.0, 1.0] rounded to 4 decimal places.
//
// Note: CONTEXT.md <deferred> identifier-similarity format spec'd
// `ERR` for Hamming length-mismatch BEFORE the Hamming silent-zero
// policy was locked (commit 1e25e31). The locked Hamming policy
// supersedes that earlier illustrative format — the example shows
// `0.0000` and never `ERR`. This resolution is a documentation
// supersession, not a scope reduction.
//
// Run with:
//
//	go run ./examples/identifier-similarity/
package main

import (
	"fmt"

	"github.com/axonops/fuzzymatch"
)

// pairs is the ordered list of database column-name identifier pairs used
// to demonstrate the six Phase 2 algorithms. The pairs are chosen to cover
// the full range of similarity scenarios:
//
//   - Naming-convention drift (snake_case ↔ camelCase)
//   - Semantic equivalence with different forms
//   - Short-token semantic synonyms
//   - Separator-only differences
//   - Abbreviation expansion
//   - Equal-length, content-different (where Hamming is informative)
//   - Same shape, opposite meaning (a known failure case — teaches limitations)
var pairs = []struct{ a, b string }{
	{"user_id", "userId"},
	{"created_at", "creationTimestamp"},
	{"status", "state"},
	{"email", "e_mail"},
	{"org_id", "organisation_id"},
	{"latitude", "longitude"},
	{"is_deleted", "is_active"},
}

// algorithms is the ordered list of fourteen Phase 2 + 3 + 4 + 5 scoring
// functions with their display names. The order matches the column layout in
// the printed table: Levenshtein, DL-OSA, DL-Full, Hamming, Jaro, Jaro-Winkler,
// SWG (Smith-Waterman-Gotoh), Strcmp95, LCSStr, RO (Ratcliff-Obershelp),
// QGramJ (q-gram Jaccard), Dice (Sørensen-Dice), Cos (Cosine), Tversky.
//
// "RO" is the short label for Ratcliff-Obershelp — the function name
// "RatcliffObershelpScore" overflows the algoWidth=13 column budget by
// several characters, so the column header uses the conventional
// abbreviation. The Go reference (godoc / llms.txt) carries the full name.
//
// The four Phase 5 q-gram algorithms have a `(a, b string, n int)` signature;
// the dispatch slot here takes `(a, b string) float64`, so each entry wraps
// the algorithm with the default `n = 3`. Tversky additionally takes
// (alpha, beta float64); the wrapper passes α=β=1.0 so the value reduces to
// the Jaccard coefficient. This is the same dispatch convention used by the
// algoid.go dispatch table per Phase 5 CONTEXT.md §5 LOCKED.
var algorithms = []struct {
	name string
	fn   func(a, b string) float64
}{
	{"Levenshtein", fuzzymatch.LevenshteinScore},
	{"DL-OSA", fuzzymatch.DamerauLevenshteinOSAScore},
	{"DL-Full", fuzzymatch.DamerauLevenshteinFullScore},
	{"Hamming", fuzzymatch.HammingScore},
	{"Jaro", fuzzymatch.JaroScore},
	{"Jaro-Winkler", fuzzymatch.JaroWinklerScore},
	{"SWG", fuzzymatch.SmithWatermanGotohScore},
	{"Strcmp95", fuzzymatch.Strcmp95Score},
	{"LCSStr", fuzzymatch.LCSStrScore},
	{"RO", fuzzymatch.RatcliffObershelpScore}, // RO = Ratcliff-Obershelp (column-width compact label)
	{"QGramJ", func(a, b string) float64 { return fuzzymatch.QGramJaccardScore(a, b, 3) }},
	{"Dice", func(a, b string) float64 { return fuzzymatch.SorensenDiceScore(a, b, 3) }},
	{"Cos", func(a, b string) float64 { return fuzzymatch.CosineScore(a, b, 3) }},
	{"Tversky", func(a, b string) float64 { return fuzzymatch.TverskyScore(a, b, 3, 1.0, 1.0) }}, // α=β=1.0 → Jaccard fallback (CONTEXT.md §5 LOCKED)
}

func main() {
	// Column widths: 32 chars for the pair column, 13 chars per algorithm.
	const pairWidth = 32
	const algoWidth = 13

	// Header row.
	fmt.Printf("%-*s", pairWidth, "Pair (a / b)")
	for _, algo := range algorithms {
		fmt.Printf("%*s", algoWidth, algo.name)
	}
	fmt.Println()

	// Separator line.
	for i := 0; i < pairWidth+len(algorithms)*algoWidth; i++ {
		fmt.Print("-")
	}
	fmt.Println()

	// Data rows: one per pair.
	for _, p := range pairs {
		label := fmt.Sprintf("%s / %s", p.a, p.b)
		fmt.Printf("%-*s", pairWidth, label)
		for _, algo := range algorithms {
			score := algo.fn(p.a, p.b)
			fmt.Printf("%*s", algoWidth, fmt.Sprintf("%.4f", score))
		}
		fmt.Println()
	}
}
