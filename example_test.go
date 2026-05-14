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

// ExampleHammingScore demonstrates the Hamming similarity. The first call uses
// the canonical Hamming 1950 reference pair (equal-length, score ≈ 0.5714).
// The second call demonstrates the LOCKED unequal-length policy: when len(a)
// != len(b), HammingDistance returns max(len(a), len(b)) (NOT zero) and
// HammingScore therefore returns 0.0 silently — no error, no panic. The "zero"
// is in the score normalisation 1 - max/max, not in the underlying distance.
func ExampleHammingScore() {
	// Equal-length: 3 mismatches in 7 positions → 1 - 3/7 ≈ 0.5714.
	fmt.Printf("%.4f\n", fuzzymatch.HammingScore("karolin", "kathrin"))
	// Unequal-length: HammingDistance("abc","ab") = max(3,2) = 3,
	// HammingScore = 1 - 3/max(3,2) = 1 - 3/3 = 0.0 (zero score, not zero distance).
	fmt.Printf("%.4f\n", fuzzymatch.HammingScore("abc", "ab"))
	// Output:
	// 0.5714
	// 0.0000
}

// ExampleJaroScore demonstrates the Jaro similarity on the canonical Winkler
// 1990 reference pair. The score is (6/6 + 6/6 + 5/6) / 3 ≈ 0.9444.
func ExampleJaroScore() {
	fmt.Printf("%.4f\n", fuzzymatch.JaroScore("MARTHA", "MARHTA"))
	// Output:
	// 0.9444
}

// ExampleDamerauLevenshteinOSAScore demonstrates the Damerau-Levenshtein OSA
// similarity. The first call shows the canonical transposition pair "ab"/"ba"
// (distance 1, score 0.5). The second call shows the discriminating vector
// "ca"/"abc" that distinguishes OSA from Full DL — OSA returns 0.0 (distance 3)
// while Full DL would return 0.3333 (distance 2) for the same pair.
func ExampleDamerauLevenshteinOSAScore() {
	fmt.Printf("%.4f\n", fuzzymatch.DamerauLevenshteinOSAScore("ab", "ba"))
	fmt.Printf("%.4f\n", fuzzymatch.DamerauLevenshteinOSAScore("ca", "abc"))
	// Output:
	// 0.5000
	// 0.0000
}

// ExampleDamerauLevenshteinFullScore demonstrates the Damerau-Levenshtein Full
// (Lowrance-Wagner 1975) similarity. The "ca"/"abc" pair demonstrates DL-Full's
// divergence from DL-OSA: DL-OSA returns 0.0000 (distance 3) for the same pair,
// while DL-Full returns 0.3333 (distance 2) because Full DL permits unrestricted
// transpositions with subsequent editing.
func ExampleDamerauLevenshteinFullScore() {
	// The "ca"/"abc" pair demonstrates DL-Full's divergence from DL-OSA:
	// DL-OSA returns 0.0000 (distance 3); DL-Full returns 0.3333 (distance 2).
	fmt.Printf("%.4f\n", fuzzymatch.DamerauLevenshteinFullScore("ca", "abc"))
	fmt.Printf("%.4f\n", fuzzymatch.DamerauLevenshteinFullScore("ab", "ba"))
	// Output:
	// 0.3333
	// 0.5000
}

// ExampleJaroWinklerScore demonstrates the Jaro-Winkler similarity on the
// canonical Winkler 1990 reference pair. The underlying Jaro score is 0.9444
// (MARTHA / MARHTA share a 3-char common prefix "MAR"); JW adds the prefix
// bonus: 0.9444 + 3 * 0.1 * (1 - 0.9444) ≈ 0.9611.
func ExampleJaroWinklerScore() {
	fmt.Printf("%.4f\n", fuzzymatch.JaroWinklerScore("MARTHA", "MARHTA"))
	// Output:
	// 0.9611
}

// ExampleSmithWatermanGotohScore demonstrates the SWG local-alignment
// similarity on a substring-containment pair. The shorter input is fully
// contained in the longer; the local alignment finds the full match, so the
// normalised score clamp(raw / min(len), 0, 1) = clamp(12.0 / 12, 0, 1) =
// 1.0000.
func ExampleSmithWatermanGotohScore() {
	fmt.Printf("%.4f\n", fuzzymatch.SmithWatermanGotohScore("http_request", "http_request_header_fields"))
	// Output:
	// 1.0000
}

// ExampleSmithWatermanGotohRawScore demonstrates the unclamped raw alignment
// score. For the same substring-containment pair, the raw score equals
// Match × min(len) = 1.0 × 12 = 12.0 (twelve match positions, no gap
// penalty). Contrast with the normalised *Score variant which clamps to 1.0.
func ExampleSmithWatermanGotohRawScore() {
	fmt.Printf("%.1f\n", fuzzymatch.SmithWatermanGotohRawScore("http_request", "http_request_header_fields"))
	// Output:
	// 12.0
}

// ExampleStrcmp95Score demonstrates Winkler's Strcmp95 similarity on the
// canonical Winkler 1990 reference pair. Strcmp95 layers four adjustments
// atop Jaro: similar-character credit (Winkler 1994 §3 table), Winkler
// prefix boost, and the long-string adjustment. MARTHA/MARHTA contains no
// similar-character pair (T/H is not in the table) but DOES trigger the
// long-string adjustment, so the score lifts above JaroWinkler's 0.9611.
func ExampleStrcmp95Score() {
	fmt.Printf("%.4f\n", fuzzymatch.Strcmp95Score("MARTHA", "MARHTA"))
	// Output:
	// 0.9676
}
