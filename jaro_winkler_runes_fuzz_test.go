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

// jaro_winkler_runes_fuzz_test.go runs native Go fuzzing against
// JaroWinklerScoreRunes — the rune-aware variant of the Winkler 1990
// score (Jaro + prefix bonus, capped 4 chars, ℓ scale 0.1). Pattern
// mirrors jaro_runes_fuzz_test.go.
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

// FuzzJaroWinklerScoreRunes asserts panic-free + score in [0,1] for
// the rune-variant Jaro-Winkler surface.
func FuzzJaroWinklerScoreRunes(f *testing.F) {
	for _, pair := range []struct{ a, b string }{
		{"MARTHA", "MARHTA"},   // Winkler 1990 canonical
		{"DIXON", "DICKSONX"},  // Winkler 1990 canonical
		{"DWAYNE", "DUANE"},    // Winkler 1990 canonical
		{"café", "café"},       // identity multi-byte
		{"Привет", "привет"},   // Cyrillic, case-only
		{"", ""},               // both-empty
		{"", "abc"},            // one-empty
		{"aaa", "bbb"},         // no common characters
		{"\xff\xfe", "abc"},    // invalid UTF-8
	} {
		f.Add(pair.a, pair.b)
	}
	f.Fuzz(func(t *testing.T, a, b string) {
		got := fuzzymatch.JaroWinklerScoreRunes(a, b)
		if math.IsNaN(got) {
			t.Fatalf("JaroWinklerScoreRunes(%q, %q) = NaN", a, b)
		}
		if math.IsInf(got, 0) {
			t.Fatalf("JaroWinklerScoreRunes(%q, %q) = Inf", a, b)
		}
		if got < 0.0 || got > 1.0 {
			t.Fatalf("JaroWinklerScoreRunes(%q, %q) = %g; want in [0,1]", a, b, got)
		}
	})
}
