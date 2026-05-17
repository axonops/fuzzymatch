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

// validate_fuzz_test.go runs native Go fuzzing against the Validate
// public surface. Properties checked per input:
//
//   1. Never panics (implicit — any panic propagates as a fuzz crash).
//   2. Every returned Warning has Kind in the documented WarnKind
//      enum range (1..len(WarnKinds())).
//   3. Every returned Warning has Algorithm either in the AlgoIDs()
//      catalogue OR equal to AlgoIDAny (the cross-cutting sentinel).
//   4. Every returned Warning's Detail is valid UTF-8.
//   5. The returned slice is sorted by (Algorithm, Kind) (determinism
//      regression gate, mirrors TestValidate_DeterministicOrdering).
//
// Seed corpus: programmatic f.Add covers (empty, empty), ("abc", "abc"),
// (short, long), (Unicode, ASCII), (invalid UTF-8, valid UTF-8). The
// fuzzer extends from this seed via the engine's coverage-guided
// mutation.
//
// Locally run `go test -fuzz=FuzzValidate -fuzztime=10s ./...` for a
// smoke check; CI's nightly fuzz job runs 60s+ per fuzzer.

package fuzzymatch_test

import (
	"sort"
	"strings"
	"testing"
	"unicode/utf8"

	"github.com/axonops/fuzzymatch"
)

// FuzzValidate is the seed-corpus + property fuzzer for the Validate
// public surface.
func FuzzValidate(f *testing.F) {
	// Programmatic seeds — canonical edge cases and the smoke set from
	// TestValidate_NeverPanics_PerAlgorithm.
	seeds := []struct {
		a, b string
	}{
		{"", ""},
		{"abc", "abc"},
		{"hello", "world"},
		{"", "abc"},
		{"abc", ""},
		{"a", "ab"},
		{"hello", "---"},
		{"中文", "日本語"},
		{"\xff\xfe", "abc"},                    // invalid UTF-8 vs ASCII
		{"\xc0\x80", "abc"},                    // overlong NUL
		{"𝕳𝖊𝖑𝖑𝖔", "Hello"},                    // 4-byte UTF-8 vs ASCII
		{strings.Repeat("a", 70_000), "short"}, // pathologically large
		{"--", "--"},                           // separator-only both sides
		{"a\x00b", "a\x00b"},                   // embedded NUL
	}
	for _, s := range seeds {
		f.Add(s.a, s.b)
	}

	// Pre-compute the valid WarnKind range so the fuzz body avoids
	// the slice allocation on every iteration.
	maxKind := fuzzymatch.WarnKind(0)
	for _, k := range fuzzymatch.WarnKinds() {
		if k > maxKind {
			maxKind = k
		}
	}

	// Pre-compute the valid AlgoID catalogue as a set for O(1) lookup.
	validAlgos := make(map[fuzzymatch.AlgoID]bool, 24)
	for _, id := range fuzzymatch.AlgoIDs() {
		validAlgos[id] = true
	}
	validAlgos[fuzzymatch.AlgoIDAny] = true

	f.Fuzz(func(t *testing.T, a, b string) {
		warnings := fuzzymatch.Validate(a, b)

		// Property 5: sorted by (Algorithm, Kind).
		if !sort.SliceIsSorted(warnings, func(i, j int) bool {
			if warnings[i].Algorithm != warnings[j].Algorithm {
				return warnings[i].Algorithm < warnings[j].Algorithm
			}
			return warnings[i].Kind < warnings[j].Kind
		}) {
			t.Errorf("Validate(%q, %q) returned unsorted warnings: %v", a, b, warnings)
		}

		for _, w := range warnings {
			// Property 2: Kind is in the documented enum range.
			if w.Kind < 1 || w.Kind > maxKind {
				t.Errorf("Validate(%q, %q) returned out-of-range Kind %d (%s)", a, b, int(w.Kind), w.Kind)
			}
			// Property 3: Algorithm is either AlgoIDAny or in the
			// AlgoIDs() catalogue.
			if !validAlgos[w.Algorithm] {
				t.Errorf("Validate(%q, %q) returned out-of-range Algorithm %d (%s)", a, b, int(w.Algorithm), w.Algorithm)
			}
			// Property 4: Detail is valid UTF-8.
			if !utf8.ValidString(w.Detail) {
				t.Errorf("Validate(%q, %q) returned invalid-UTF-8 Detail: %q", a, b, w.Detail)
			}
		}

		// Determinism: two calls return identical output.
		second := fuzzymatch.Validate(a, b)
		if len(warnings) != len(second) {
			t.Errorf("Validate(%q, %q) non-deterministic: len differs %d vs %d", a, b, len(warnings), len(second))
			return
		}
		for i := range warnings {
			if warnings[i] != second[i] {
				t.Errorf("Validate(%q, %q) non-deterministic at index %d: %v vs %v", a, b, i, warnings[i], second[i])
			}
		}
	})
}
