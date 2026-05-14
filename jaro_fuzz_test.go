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

// jaro_fuzz_test.go runs native Go fuzzing against JaroScore.
// Two properties for any input:
//
//  1. Never panics (implicit — any panic is reported as a fuzz crash).
//  2. Score is in [0.0, 1.0]; never NaN; never Inf.
//
// Programmatic seeds include the canonical Jaro 1989 / Winkler 1990 reference
// vectors, invalid UTF-8 sequences (to exercise the byte-level path's
// resilience — T-02-03-02 mitigation), and a length-mismatched pair.
//
// The on-disk corpus lives in testdata/fuzz/FuzzJaroScore/. CI's nightly fuzz
// job runs the fuzzer for 60s; locally run
// `go test -fuzz=FuzzJaroScore -fuzztime=30s ./...` for a quick smoke test.

package fuzzymatch_test

import (
	"math"
	"testing"

	"github.com/axonops/fuzzymatch"
)

// FuzzJaroScore asserts panic-free + score in [0,1] + no NaN + no Inf for all
// inputs, mitigating T-02-03-02 (malformed UTF-8 causing panic).
func FuzzJaroScore(f *testing.F) {
	// Programmatic seed entries — canonical reference vectors plus edge cases.
	for _, pair := range []struct{ a, b string }{
		{"MARTHA", "MARHTA"},                // Winkler 1990 reference pair
		{"DIXON", "DICKSONX"},               // Winkler 1990 reference pair
		{"JELLYFISH", "SMELLYFISH"},         // Jaro 1989 reference pair
		{"ABC", "ABC"},                      // identical
		{"", "ABC"},                         // one-empty
		{"", ""},                            // both-empty
		{"\xff\xfe", "abc"},                 // invalid UTF-8 (high bytes without continuation)
		{"\xc0\x80", "abc"},                 // invalid UTF-8 (overlong NUL encoding)
		{"hello world", "hello"},            // common prefix, length mismatch
		{"MARTHA", "MARHTA EXTRA PADDING"},  // length mismatch
	} {
		f.Add(pair.a, pair.b)
	}

	f.Fuzz(func(t *testing.T, a, b string) {
		// Property 1: must not panic. (Implicit — any panic from JaroScore
		// propagates to the fuzz harness and is reported as a crash.)
		got := fuzzymatch.JaroScore(a, b)

		// Property 2: score must not be NaN.
		if math.IsNaN(got) {
			t.Errorf("JaroScore(%q, %q) = NaN; want a value in [0,1]", a, b)
		}
		// Property 3: score must not be Inf.
		if math.IsInf(got, 0) {
			t.Errorf("JaroScore(%q, %q) = Inf; want a value in [0,1]", a, b)
		}
		// Property 4: score must be in [0.0, 1.0].
		if got < 0.0 || got > 1.0 {
			t.Errorf("JaroScore(%q, %q) = %g; want in [0,1]", a, b, got)
		}
	})
}
