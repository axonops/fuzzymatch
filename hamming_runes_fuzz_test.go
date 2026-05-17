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

// hamming_runes_fuzz_test.go runs native Go fuzzing against
// HammingScoreRunes — the rune-aware variant of the Hamming 1950
// score. Pattern mirrors hamming_fuzz_test.go; the rune variant
// counts rune-level mismatches rather than byte-level, so multi-byte
// inputs of equal rune count exercise the rune-counting path
// directly.
//
// Properties:
//
//  1. Never panics on arbitrary input.
//  2. Score is in [0.0, 1.0]; never NaN; never Inf.
//
// Threat model: T-08.5-24 (D - DoS via fuzz-discovered panic).

package fuzzymatch_test

import (
	"math"
	"testing"

	"github.com/axonops/fuzzymatch"
)

// FuzzHammingScoreRunes asserts panic-free + score in [0,1] for the
// rune-variant Hamming surface, including unequal-length inputs
// (Q1 silent-zero policy: HammingScore returns 0.0 on unequal length).
func FuzzHammingScoreRunes(f *testing.F) {
	for _, pair := range []struct{ a, b string }{
		{"karolin", "kathrin"},   // Hamming 1950 canonical (ASCII, dist 3)
		{"1011101", "1001001"},   // Hamming 1950 canonical (bit-string, dist 2)
		{"café", "cafè"},         // equal-rune-count multi-byte
		{"Привет", "Прывет"},     // Cyrillic, single rune diff
		{"abc", "abc"},           // identity
		{"", ""},                 // both-empty
		{"abc", "ab"},            // unequal-length (silent-zero policy)
		{"\xff\xfe", "\xff\xff"}, // invalid UTF-8, equal byte count
		{"日本語", "日本語"},          // identity multi-byte
	} {
		f.Add(pair.a, pair.b)
	}
	f.Fuzz(func(t *testing.T, a, b string) {
		got := fuzzymatch.HammingScoreRunes(a, b)
		if math.IsNaN(got) {
			t.Fatalf("HammingScoreRunes(%q, %q) = NaN", a, b)
		}
		if math.IsInf(got, 0) {
			t.Fatalf("HammingScoreRunes(%q, %q) = Inf", a, b)
		}
		if got < 0.0 || got > 1.0 {
			t.Fatalf("HammingScoreRunes(%q, %q) = %g; want in [0,1]", a, b, got)
		}
	})
}
