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

// lcsstr_fuzz_test.go runs native Go fuzzing against the FULL LCSStr public
// surface. One fuzzer exercises ALL FOUR public functions per input pair
// (Phase 3 WR-02 closure — the fuzz harness exercises the full public surface,
// not just the dispatched byte-path score function):
//
//   - LCSStrScore                (byte path, score)
//   - LCSStrScoreRunes           (rune path, score)
//   - LongestCommonSubstring     (byte path, substring)
//   - LongestCommonSubstringRunes (rune path, substring)
//
// Properties checked per surface, per input:
//
//  1. Never panics (implicit — any panic propagates as a fuzz crash).
//  2. Score surfaces never return NaN.
//  3. Score surfaces never return ±Inf.
//  4. Score surfaces return a value in [0.0, 1.0].
//  5. Substring surfaces tolerate invalid UTF-8 / multi-byte UTF-8 / Cyrillic
//     without panic.
//
// Programmatic seeds cover all RESEARCH.md required-case categories: canonical
// reference vector (kitten/sitting), substring containment (http_request),
// identity, both-empty, one-empty, leftmost-tie-break (abcXYZabc/abc — load-
// bearing for Pitfall 4), no-overlap (abc/xyz — Pitfall 6 disambiguation),
// invalid UTF-8 (\xff\xfe), multi-byte UTF-8 (café/cafe), and Cyrillic
// (Привет/привет).
//
// The on-disk corpus lives in testdata/fuzz/FuzzLCSStrScore/. CI's nightly
// fuzz job runs the fuzzer for 60s+; locally run
// `go test -fuzz=FuzzLCSStrScore -fuzztime=10s ./...` for a smoke check.

package fuzzymatch_test

import (
	"math"
	"testing"

	"github.com/axonops/fuzzymatch"
)

// FuzzLCSStrScore asserts panic-free + finite-and-in-range across all four
// public LCSStr functions for every input pair (byte + rune paths,
// score + substring surfaces).
func FuzzLCSStrScore(f *testing.F) {
	// Programmatic seed entries — canonical reference vectors plus the
	// Pitfall-4-specific leftmost-tie-break canary and Pitfall-6 no-overlap
	// disambiguation case.
	for _, pair := range []struct{ a, b string }{
		{"kitten", "sitting"},                          // Wagner-Fischer canonical
		{"http_request", "http_request_header_fields"}, // substring containment
		{"abc", "abc"},                                 // identical
		{"", "abc"},                                    // one-empty
		{"abc", ""},                                    // one-empty (reverse)
		{"", ""},                                       // both-empty
		{"abcXYZabc", "abc"},                           // LEFTMOST tie-break (Pitfall 4)
		{"abc", "xyz"},                                 // no-overlap disambiguation (Pitfall 6)
		{"\xff\xfe", "abc"},                            // invalid UTF-8
		{"\xc0\x80", "abc"},                            // invalid UTF-8 (overlong NUL)
		{"café", "cafe"},                               // Latin supplement multi-byte
		{"Привет", "привет"},                           // Cyrillic multi-byte UTF-8
		{"mississippi", "issi"},                        // tied 4-char overlap
		{"qqqq", "zzzz"},                               // no overlap (4 chars each)
	} {
		f.Add(pair.a, pair.b)
	}

	f.Fuzz(func(t *testing.T, a, b string) {
		// Property: substring-returning surfaces must not panic. Returns are
		// allowed to be empty (both-empty and no-overlap cases) but must not
		// panic on any byte/rune content. We use `_ = ...` to assert
		// computation completes without panic.
		_ = fuzzymatch.LongestCommonSubstring(a, b)
		_ = fuzzymatch.LongestCommonSubstringRunes(a, b)

		// Property: score-returning surfaces. Each value must be finite and
		// in [0.0, 1.0]. Two surfaces per input.
		scores := []struct {
			name string
			val  float64
		}{
			{"LCSStrScore", fuzzymatch.LCSStrScore(a, b)},
			{"LCSStrScoreRunes", fuzzymatch.LCSStrScoreRunes(a, b)},
		}
		for _, s := range scores {
			if math.IsNaN(s.val) {
				t.Errorf("%s(%q, %q) = NaN; want a finite value", s.name, a, b)
			}
			if math.IsInf(s.val, 0) {
				t.Errorf("%s(%q, %q) = Inf; want a finite value", s.name, a, b)
			}
			if s.val < 0.0 || s.val > 1.0 {
				t.Errorf("%s(%q, %q) = %g; want in [0,1]", s.name, a, b, s.val)
			}
		}
	})
}
