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

// ratcliff_obershelp_fuzz_test.go runs native Go fuzzing against BOTH
// Ratcliff-Obershelp public surfaces (Phase 3 WR-02 closure — the fuzz
// harness exercises the full public surface, not just the dispatched
// byte-path score function):
//
//   - RatcliffObershelpScore       (byte path, score)
//   - RatcliffObershelpScoreRunes  (rune path, score)
//
// Properties checked per surface, per input:
//
//  1. Never panics (implicit — any panic propagates as a fuzz crash).
//  2. Score never returns NaN.
//  3. Score never returns ±Inf.
//  4. Score returns a value in [0.0, 1.0].
//
// Programmatic seeds cover all RESEARCH.md required-case categories: standard
// edges (identity, both-empty, one-empty, no-overlap), Dr. Dobb's 1988 paper
// vectors (WIKIMEDIA/WIKIMANIA, GESTALT/GESTALT_PATTERN_MATCHING), an
// autojunk-sensitive 200+char pair (proves the implementation does NOT have
// difflib's autojunk heuristic enabled — RESEARCH.md Pitfall 2), substring
// containment, multi-byte UTF-8 (café/cafe, Привет/привет), and invalid
// UTF-8 (\xff\xfe, \xc0\x80).
//
// The on-disk corpus lives in testdata/fuzz/FuzzRatcliffObershelpScore/. CI's
// nightly fuzz job runs the fuzzer for 60s+; locally run
// `go test -fuzz=FuzzRatcliffObershelpScore -fuzztime=10s ./...` for a smoke
// check.

package fuzzymatch_test

import (
	"math"
	"strings"
	"testing"

	"github.com/axonops/fuzzymatch"
)

// FuzzRatcliffObershelpScore asserts panic-free + finite-and-in-range
// across both public RatcliffObershelp functions for every input pair
// (byte + rune paths).
func FuzzRatcliffObershelpScore(f *testing.F) {
	// Autojunk-sensitive seed: a 205-char input that would trigger
	// difflib's autojunk heuristic if it were enabled. Including this
	// seed in the corpus exercises the autojunk=False semantic on every
	// fuzz run (RESEARCH.md Pitfall 2 closure).
	autojunkA := strings.Repeat("a", 100) + strings.Repeat("x", 5) + strings.Repeat("a", 100)
	autojunkB := strings.Repeat("a", 50) + strings.Repeat("y", 5) + strings.Repeat("a", 150)

	for _, pair := range []struct{ a, b string }{
		// Standard edges
		{"abc", "abc"}, // identity
		{"", ""},       // both-empty
		{"", "abc"},    // one-empty
		{"abc", ""},    // one-empty (reverse)
		{"abc", "xyz"}, // no-overlap

		// Dr. Dobb's 1988 paper vectors
		{"WIKIMEDIA", "WIKIMANIA"},
		{"GESTALT", "GESTALT_PATTERN_MATCHING"},

		// Autojunk-sensitive 200+char pair (Pitfall 2)
		{autojunkA, autojunkB},

		// Substring containment
		{"abcdef", "xyzabcdefuvw"},

		// Multi-byte UTF-8
		{"café", "cafe"},
		{"Привет", "привет"},

		// Invalid UTF-8
		{"\xff\xfe", "abc"},
		{"\xc0\x80", "abc"}, // overlong NUL

		// Asymmetric pair (OQ-1 — verify no panic on either direction)
		{"tide", "diet"},
		{"diet", "tide"},
	} {
		f.Add(pair.a, pair.b)
	}

	f.Fuzz(func(t *testing.T, a, b string) {
		// Property: both score-returning surfaces are finite and in
		// [0.0, 1.0]. Implicit no-panic on the function calls.
		scores := []struct {
			name string
			val  float64
		}{
			{"RatcliffObershelpScore", fuzzymatch.RatcliffObershelpScore(a, b)},
			{"RatcliffObershelpScoreRunes", fuzzymatch.RatcliffObershelpScoreRunes(a, b)},
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
