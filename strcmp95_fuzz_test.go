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

// strcmp95_fuzz_test.go runs native Go fuzzing against Strcmp95Score.
//
// Properties checked per input:
//
//  1. Never panics (implicit — any panic is reported as a fuzz crash).
//  2. Never returns NaN.
//  3. Never returns +/-Inf.
//  4. Returns a value in [0.0, 1.0].
//
// Single surface (no Runes, no Params, no Raw per CONTEXT.md §2) — the
// multi-surface pattern from swg_fuzz_test.go collapses to a single direct
// call.
//
// Programmatic seeds cover RESEARCH.md required-case categories: canonical
// Winkler 1994 / Census Bureau reference vectors, identity, both-empty,
// one-empty, invalid UTF-8 (Strcmp95 is byte-only — invalid sequences must
// not panic), the long-string adjustment trigger pair, and a length-≤4
// pair where the long-string adjustment must NOT fire.
//
// The on-disk corpus lives in testdata/fuzz/FuzzStrcmp95Score/. CI's
// nightly fuzz job runs the fuzzer for 60s+; locally run
// `go test -fuzz=FuzzStrcmp95Score -fuzztime=30s ./...` for smoke.

package fuzzymatch_test

import (
	"math"
	"testing"

	"github.com/axonops/fuzzymatch"
)

// FuzzStrcmp95Score asserts panic-free + finite-and-in-range across the
// Strcmp95Score surface for every input pair.
func FuzzStrcmp95Score(f *testing.F) {
	for _, pair := range []struct{ a, b string }{
		{"MARTHA", "MARHTA"},        // Winkler 1990 canonical pair
		{"DWAYNE", "DUANE"},         // Census Bureau — similar-char table fires (W/U)
		{"DIXON", "DICKSONX"},       // Census Bureau — similar-char table fires (C/K)
		{"abc", "abc"},              // identical (identity short-circuit)
		{"", "abc"},                 // one-empty
		{"abc", ""},                 // one-empty (reverse)
		{"", ""},                    // both-empty
		{"\xff\xfe", "abc"},         // invalid UTF-8 — Strcmp95 is byte-only
		{"\xc0\x80", "abc"},         // invalid UTF-8 (overlong NUL)
		{"HAMINGTON", "HAMMINGTON"}, // long-string-adjustment trigger (Pitfall 5)
		{"AB", "AC"},                // length ≤ 4 (long-string adj should NOT fire)
		{"dwayne", "duane"},         // case-folded similar-pair (verifies ASCII case-fold in lookup)
		{"qqqq", "zzzz"},            // no overlap
	} {
		f.Add(pair.a, pair.b)
	}
	f.Fuzz(func(t *testing.T, a, b string) {
		got := fuzzymatch.Strcmp95Score(a, b)
		if math.IsNaN(got) {
			t.Errorf("Strcmp95Score(%q, %q) = NaN; want a finite value", a, b)
		}
		if math.IsInf(got, 0) {
			t.Errorf("Strcmp95Score(%q, %q) = Inf; want a finite value", a, b)
		}
		if got < 0.0 || got > 1.0 {
			t.Errorf("Strcmp95Score(%q, %q) = %g; want in [0,1]", a, b, got)
		}
	})
}
