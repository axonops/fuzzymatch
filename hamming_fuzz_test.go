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

// hamming_fuzz_test.go runs native Go fuzzing against HammingScore.
// Two properties for any input:
//
//  1. Never panics (implicit — any panic is reported as a fuzz crash).
//  2. Score is in [0.0, 1.0]; never NaN; never Inf.
//
// Programmatic seeds include the canonical Hamming 1950 reference vectors,
// invalid UTF-8 sequences (to exercise the byte-level path's resilience — the
// byte variant operates on bytes directly so invalid UTF-8 cannot panic), and
// a length-mismatched pair to exercise the silent-zero path.
//
// The on-disk corpus lives in testdata/fuzz/FuzzHammingScore/. CI's
// nightly fuzz job runs the fuzzer for 60s; locally run
// `go test -fuzz=FuzzHammingScore -fuzztime=30s ./...` for a quick smoke test.

package fuzzymatch_test

import (
	"math"
	"testing"

	"github.com/axonops/fuzzymatch"
)

// FuzzHammingScore asserts panic-free + score in [0,1] for all inputs,
// including invalid UTF-8 and length-mismatched pairs.
func FuzzHammingScore(f *testing.F) {
	// Programmatic seed entries — canonical reference vectors plus edge cases.
	for _, pair := range []struct{ a, b string }{
		{"karolin", "kathrin"},    // Hamming 1950 reference pair — distance 3
		{"1011101", "1001001"},    // Hamming 1950 reference pair — distance 2
		{"abc", "abc"},            // identical
		{"", ""},                  // both-empty → score 1.0
		{"abc", "ab"},             // unequal-length → silent-zero policy
		{"ab", "abc"},             // unequal-length, reversed
		{"", "abc"},               // one-empty
		{"\xff\xfe", "abc"},       // invalid UTF-8 (byte path must not panic)
		{"\xc0\x80", "\xc0\x81"}, // invalid UTF-8 overlong sequences
		{"café", "cafè"},          // equal-rune-count multi-byte (rune path)
	} {
		f.Add(pair.a, pair.b)
	}

	f.Fuzz(func(t *testing.T, a, b string) {
		// Property 1: must not panic (implicit — propagates to fuzz harness as crash).
		got := fuzzymatch.HammingScore(a, b)

		// Property 2: score must be in [0.0, 1.0].
		if math.IsNaN(got) {
			t.Errorf("HammingScore(%q, %q) = NaN; want a value in [0,1]", a, b)
		}
		if math.IsInf(got, 0) {
			t.Errorf("HammingScore(%q, %q) = Inf; want a value in [0,1]", a, b)
		}
		if got < 0.0 || got > 1.0 {
			t.Errorf("HammingScore(%q, %q) = %g; want in [0,1]", a, b, got)
		}
	})
}
