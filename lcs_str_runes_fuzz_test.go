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

// lcs_str_runes_fuzz_test.go runs native Go fuzzing against
// LCSStrScoreRunes — the rune-aware variant of the Wagner-Fischer 1974
// longest-common-substring score. Pattern mirrors levenshtein_runes_fuzz_test.go.
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

// FuzzLCSStrScoreRunes asserts panic-free + score in [0,1] for the
// rune-variant LCSStr surface.
func FuzzLCSStrScoreRunes(f *testing.F) {
	for _, pair := range []struct{ a, b string }{
		{"ABABC", "BABCA"},        // Wagner-Fischer canonical (substring "BABC")
		{"hello", "world"},        // no shared substring of length > 1
		{"café", "macaroni café"}, // multi-byte substring containment
		{"Привет", "привет"},      // Cyrillic, case-only
		{"", ""},                  // both-empty
		{"", "abc"},               // one-empty
		{"abc", "abc"},            // identity
		{"\xff\xfe", "abc"},       // invalid UTF-8
		{"日本語", "日本"},             // multi-byte substring
	} {
		f.Add(pair.a, pair.b)
	}
	f.Fuzz(func(t *testing.T, a, b string) {
		got := fuzzymatch.LCSStrScoreRunes(a, b)
		if math.IsNaN(got) {
			t.Fatalf("LCSStrScoreRunes(%q, %q) = NaN", a, b)
		}
		if math.IsInf(got, 0) {
			t.Fatalf("LCSStrScoreRunes(%q, %q) = Inf", a, b)
		}
		if got < 0.0 || got > 1.0 {
			t.Fatalf("LCSStrScoreRunes(%q, %q) = %g; want in [0,1]", a, b, got)
		}
	})
}
