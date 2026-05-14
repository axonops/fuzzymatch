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

// damerau_full_fuzz_test.go runs native Go fuzzing against
// DamerauLevenshteinFullScore. Two properties for any input:
//
//  1. Never panics (implicit — any panic is reported as a fuzz crash).
//  2. Score is in [0.0, 1.0]; never NaN; never Inf.
//
// Programmatic seeds include the canonical Lowrance-Wagner 1975 reference
// vectors, the discriminating vector "ca"/"abc" (Full DL returns 2, not 3),
// invalid UTF-8 sequences (to exercise the byte-level path's resilience),
// and Cyrillic.
//
// The on-disk corpus lives in testdata/fuzz/FuzzDamerauLevenshteinFullScore/.
// CI's nightly fuzz job runs the fuzzer for 60s; locally run:
// `go test -fuzz=FuzzDamerauLevenshteinFullScore -fuzztime=30s ./...`

package fuzzymatch_test

import (
	"math"
	"testing"

	"github.com/axonops/fuzzymatch"
)

// FuzzDamerauLevenshteinFullScore asserts panic-free + score in [0,1] for all
// inputs, including invalid UTF-8 (the byte path is byte-safe by construction;
// U+FFFD replacement happens at the []rune conversion level for the rune path).
func FuzzDamerauLevenshteinFullScore(f *testing.F) {
	// Programmatic seed entries — canonical reference vectors plus edge cases.
	for _, pair := range []struct{ a, b string }{
		{"ab", "ba"},          // Lowrance-Wagner — transposition costs 1 (same as OSA)
		{"ca", "abc"},         // Discriminating vector — Full DL returns 2 (NOT 3)
		{"abc", "abc"},        // identical
		{"", "abc"},           // one-empty
		{"", ""},              // both-empty
		{"\xff\xfe", "abc"},   // invalid UTF-8 (high bytes without continuation)
		{"\xc0\x80", "abc"},   // invalid UTF-8 (overlong NUL encoding)
		{"Привет", "привет"},  // Cyrillic (multi-byte UTF-8)
		{"café", "cafe"},      // Latin supplement (é = U+00E9, 2 bytes)
		{"kitten", "sitting"}, // Levenshtein canonical pair (cross-check)
	} {
		f.Add(pair.a, pair.b)
	}

	f.Fuzz(func(t *testing.T, a, b string) {
		// Property 1: must not panic. (Implicit — any panic from
		// DamerauLevenshteinFullScore propagates to the fuzz harness.)
		got := fuzzymatch.DamerauLevenshteinFullScore(a, b)

		// Property 2: score must be in [0.0, 1.0].
		if math.IsNaN(got) {
			t.Errorf("DamerauLevenshteinFullScore(%q, %q) = NaN; want a value in [0,1]", a, b)
		}
		if math.IsInf(got, 0) {
			t.Errorf("DamerauLevenshteinFullScore(%q, %q) = Inf; want a value in [0,1]", a, b)
		}
		if got < 0.0 || got > 1.0 {
			t.Errorf("DamerauLevenshteinFullScore(%q, %q) = %g; want in [0,1]", a, b, got)
		}
	})
}
