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

// swg_fuzz_test.go runs native Go fuzzing against SmithWatermanGotohScore.
// Three properties for any input:
//
//  1. Never panics (implicit — any panic is reported as a fuzz crash).
//  2. Score is in [0.0, 1.0]; never NaN; never Inf.
//  3. The byte path tolerates invalid UTF-8 without panic.
//
// Programmatic seeds cover all CONTEXT.md §1 + RESEARCH.md required-case
// categories: canonical reference vector, substring containment, identity,
// both-empty, one-empty, the Gotoh-erratum gap-split canary, invalid UTF-8,
// and Cyrillic multi-byte UTF-8.
//
// The on-disk corpus lives in testdata/fuzz/FuzzSmithWatermanGotohScore/. CI's
// nightly fuzz job runs the fuzzer for 60s+; locally run
// `go test -fuzz=FuzzSmithWatermanGotohScore -fuzztime=30s ./...` for smoke.

package fuzzymatch_test

import (
	"math"
	"testing"

	"github.com/axonops/fuzzymatch"
)

// FuzzSmithWatermanGotohScore asserts panic-free + score in [0,1] for all
// inputs.
func FuzzSmithWatermanGotohScore(f *testing.F) {
	// Programmatic seed entries — canonical reference vectors plus the
	// Phase-3-specific Gotoh-erratum canaries.
	for _, pair := range []struct{ a, b string }{
		{"kitten", "sitting"},                          // Wagner-Fischer canonical
		{"http_request", "http_request_header_fields"}, // substring containment
		{"abc", "abc"},                                 // identical
		{"", "abc"},                                    // one-empty
		{"", ""},                                       // both-empty
		{"abc________def", "abcdef"},                   // one-long-gap canary (Gotoh)
		{"abc____def____", "abcdef"},                   // gap-split symmetry case
		{"\xff\xfe", "abc"},                            // invalid UTF-8
		{"\xc0\x80", "abc"},                            // invalid UTF-8 (overlong NUL)
		{"Привет", "привет"},                           // Cyrillic multi-byte UTF-8
		{"café", "cafe"},                               // Latin supplement multi-byte
		{"hello world", "hello"},                       // partial substring
		{"qqqq", "zzzz"},                               // no overlap
	} {
		f.Add(pair.a, pair.b)
	}

	f.Fuzz(func(t *testing.T, a, b string) {
		// Property 1: must not panic. Any panic propagates to the fuzz
		// harness and is reported as a crash.
		got := fuzzymatch.SmithWatermanGotohScore(a, b)

		// Property 2: score must be a finite value in [0.0, 1.0].
		if math.IsNaN(got) {
			t.Errorf("SmithWatermanGotohScore(%q, %q) = NaN; want a value in [0,1]", a, b)
		}
		if math.IsInf(got, 0) {
			t.Errorf("SmithWatermanGotohScore(%q, %q) = Inf; want a value in [0,1]", a, b)
		}
		if got < 0.0 || got > 1.0 {
			t.Errorf("SmithWatermanGotohScore(%q, %q) = %g; want in [0,1]", a, b, got)
		}
	})
}
