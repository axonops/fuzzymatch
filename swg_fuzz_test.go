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

// swg_fuzz_test.go runs native Go fuzzing against the full SWG public surface.
// One fuzzer exercises all six public functions per input pair:
//
//   - SmithWatermanGotohScore           (byte path, default params, clamped)
//   - SmithWatermanGotohScoreRunes      (rune path, default params, clamped)
//   - SmithWatermanGotohScoreWithParams (byte path, custom params, clamped)
//   - SmithWatermanGotohRawScore        (byte path, default params, unclamped)
//   - SmithWatermanGotohRawScoreRunes   (rune path, default params, unclamped)
//   - SmithWatermanGotohRawScoreWithParams (byte path, custom params, unclamped)
//
// Properties checked per surface, per input:
//
//  1. Never panics (implicit — any panic is reported as a fuzz crash).
//  2. Never returns NaN.
//  3. Never returns +/-Inf.
//  4. Normalised surfaces (Score / ScoreRunes / ScoreWithParams) return a
//     value in [0.0, 1.0].
//  5. The byte path tolerates invalid UTF-8 without panic; the rune path
//     accepts the resulting utf8.RuneError without panic.
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

// FuzzSmithWatermanGotohScore asserts panic-free + finite-and-in-range across
// all six public SWG functions for every input pair (default + custom params,
// byte + rune paths, normalised + raw surfaces).
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

	// custom is a non-default parameter set exercising the *WithParams kernel.
	// Values mirror the "non_default_params" entry in the cross-validation
	// corpus so any kernel transcription bug surfaces on the same numerical
	// gate as the corpus.
	custom := fuzzymatch.SWGParams{
		Match:     2.0,
		Mismatch:  -2.0,
		GapOpen:   -3.0,
		GapExtend: -1.0,
	}

	f.Fuzz(func(t *testing.T, a, b string) {
		// Property 1: none of the six public surfaces may panic. Any panic
		// propagates to the fuzz harness and is reported as a crash.
		surfaces := []struct {
			name   string
			val    float64
			bounded bool // true → must be in [0,1]; false → finite-only
		}{
			{"Score", fuzzymatch.SmithWatermanGotohScore(a, b), true},
			{"ScoreRunes", fuzzymatch.SmithWatermanGotohScoreRunes(a, b), true},
			{"ScoreWithParams", fuzzymatch.SmithWatermanGotohScoreWithParams(a, b, custom), true},
			{"RawScore", fuzzymatch.SmithWatermanGotohRawScore(a, b), false},
			{"RawScoreRunes", fuzzymatch.SmithWatermanGotohRawScoreRunes(a, b), false},
			{"RawScoreWithParams", fuzzymatch.SmithWatermanGotohRawScoreWithParams(a, b, custom), false},
		}

		for _, s := range surfaces {
			// Property 2: never NaN.
			if math.IsNaN(s.val) {
				t.Errorf("%s(%q, %q) = NaN; want a finite value",
					s.name, a, b)
			}
			// Property 3: never Inf.
			if math.IsInf(s.val, 0) {
				t.Errorf("%s(%q, %q) = Inf; want a finite value",
					s.name, a, b)
			}
			// Property 4: normalised surfaces must be in [0.0, 1.0].
			if s.bounded && (s.val < 0.0 || s.val > 1.0) {
				t.Errorf("%s(%q, %q) = %g; want in [0,1]",
					s.name, a, b, s.val)
			}
		}
	})
}
