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

// damerau_full_runes_fuzz_test.go runs native Go fuzzing against
// DamerauLevenshteinFullScoreRunes — the rune-aware variant of the
// Lowrance-Wagner 1975 unrestricted transposition score. Pattern
// mirrors levenshtein_runes_fuzz_test.go; the seed corpus exercises
// transposition cases plus multi-byte Unicode.
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

// FuzzDamerauLevenshteinFullScoreRunes asserts panic-free + score in
// [0,1] for the rune-variant DL-Full surface.
func FuzzDamerauLevenshteinFullScoreRunes(f *testing.F) {
	for _, pair := range []struct{ a, b string }{
		{"ca", "abc"},        // Lowrance-Wagner discriminating vector
		{"abcd", "abdc"},     // single adjacent transposition
		{"café", "caéf"},     // multi-byte transposition
		{"Привет", "Прывет"}, // Cyrillic substitution
		{"", ""},             // both-empty
		{"", "abc"},          // one-empty
		{"aaa", "bbb"},       // orthogonal
		{"\xff\xfe", "abc"},  // invalid UTF-8
		{"日本語", "本日語"},       // multi-byte transposition
	} {
		f.Add(pair.a, pair.b)
	}
	f.Fuzz(func(t *testing.T, a, b string) {
		got := fuzzymatch.DamerauLevenshteinFullScoreRunes(a, b)
		if math.IsNaN(got) {
			t.Fatalf("DamerauLevenshteinFullScoreRunes(%q, %q) = NaN", a, b)
		}
		if math.IsInf(got, 0) {
			t.Fatalf("DamerauLevenshteinFullScoreRunes(%q, %q) = Inf", a, b)
		}
		if got < 0.0 || got > 1.0 {
			t.Fatalf("DamerauLevenshteinFullScoreRunes(%q, %q) = %g; want in [0,1]", a, b, got)
		}
	})
}
