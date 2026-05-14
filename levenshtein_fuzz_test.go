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

// levenshtein_fuzz_test.go runs native Go fuzzing against LevenshteinScore.
// Two properties for any input:
//
//  1. Never panics (implicit — any panic is reported as a fuzz crash).
//  2. Score is in [0.0, 1.0]; never NaN; never Inf.
//
// Programmatic seeds include the canonical Wagner-Fischer reference vectors,
// invalid UTF-8 sequences (to exercise the byte-level path's resilience), and
// Cyrillic (to exercise the rune path).
//
// The on-disk corpus lives in testdata/fuzz/FuzzLevenshteinScore/. CI's
// nightly fuzz job (per Makefile `test-fuzz`) runs the fuzzer for 60s;
// locally run `go test -fuzz=FuzzLevenshteinScore -fuzztime=30s ./...` for a
// quick smoke test.

package fuzzymatch_test

import (
	"math"
	"testing"

	"github.com/axonops/fuzzymatch"
)

// FuzzLevenshteinScore asserts panic-free + score in [0,1] for all inputs.
func FuzzLevenshteinScore(f *testing.F) {
	// Programmatic seed entries — canonical reference vectors plus edge cases.
	for _, pair := range []struct{ a, b string }{
		{"kitten", "sitting"},    // Wagner-Fischer 1974 reference pair
		{"saturday", "sunday"},   // Wagner-Fischer 1974 reference pair
		{"abc", "abc"},           // identical
		{"", "abc"},              // one-empty
		{"", ""},                 // both-empty
		{"\xff\xfe", "abc"},      // invalid UTF-8 (high bytes without continuation)
		{"\xc0\x80", "abc"},      // invalid UTF-8 (overlong NUL encoding)
		{"Привет", "привет"},     // Cyrillic (multi-byte UTF-8)
		{"café", "cafe"},         // Latin supplement (é = U+00E9, 2 bytes)
		{"hello world", "hello"}, // common prefix
	} {
		f.Add(pair.a, pair.b)
	}

	f.Fuzz(func(t *testing.T, a, b string) {
		// Property 1: must not panic. (Implicit — any panic from LevenshteinScore
		// propagates to the fuzz harness and is reported as a crash.)
		got := fuzzymatch.LevenshteinScore(a, b)

		// Property 2: score must be in [0.0, 1.0].
		if math.IsNaN(got) {
			t.Errorf("LevenshteinScore(%q, %q) = NaN; want a value in [0,1]", a, b)
		}
		if math.IsInf(got, 0) {
			t.Errorf("LevenshteinScore(%q, %q) = Inf; want a value in [0,1]", a, b)
		}
		if got < 0.0 || got > 1.0 {
			t.Errorf("LevenshteinScore(%q, %q) = %g; want in [0,1]", a, b, got)
		}
	})
}
