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

// ratcliff_obershelp_runes_fuzz_test.go runs native Go fuzzing
// against RatcliffObershelpScoreRunes — the rune-aware variant of the
// Ratcliff-Metzener 1988 gestalt score. Pattern mirrors
// levenshtein_runes_fuzz_test.go.
//
// Note: Ratcliff-Obershelp is asymmetric by design (Phase 8.5 Q6a;
// docs/requirements.md §7.24); the fuzz harness asserts range
// bounds and panic-freedom only — symmetry is NOT asserted.
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

// FuzzRatcliffObershelpScoreRunes asserts panic-free + score in
// [0,1] for the rune-variant Ratcliff-Obershelp surface.
func FuzzRatcliffObershelpScoreRunes(f *testing.F) {
	for _, pair := range []struct{ a, b string }{
		{"Pennsylvania", "Pencilvaneya"}, // Ratcliff 1988 canonical
		{"WIKIMEDIA", "WIKIMANIA"},        // difflib-aligned canonical
		{"café", "cafe"},                  // Latin-supplement
		{"Привет", "привет"},              // Cyrillic, case-only
		{"", ""},                          // both-empty
		{"", "abc"},                       // one-empty
		{"aaa", "bbb"},                    // no overlap
		{"\xff\xfe", "abc"},               // invalid UTF-8
		{"日本語", "日本語"},                   // identity multi-byte
	} {
		f.Add(pair.a, pair.b)
	}
	f.Fuzz(func(t *testing.T, a, b string) {
		got := fuzzymatch.RatcliffObershelpScoreRunes(a, b)
		if math.IsNaN(got) {
			t.Fatalf("RatcliffObershelpScoreRunes(%q, %q) = NaN", a, b)
		}
		if math.IsInf(got, 0) {
			t.Fatalf("RatcliffObershelpScoreRunes(%q, %q) = Inf", a, b)
		}
		if got < 0.0 || got > 1.0 {
			t.Fatalf("RatcliffObershelpScoreRunes(%q, %q) = %g; want in [0,1]", a, b, got)
		}
	})
}
