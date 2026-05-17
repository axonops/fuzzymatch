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

// levenshtein_distance_fuzz_test.go runs native Go fuzzing against
// LevenshteinDistance. The distance variants of the four character
// algorithms (Levenshtein, DL-Full, DL-OSA, Hamming) return an int
// rather than a [0, 1] float, so the fuzz harness asserts the
// mathematical invariants the function promises:
//
//  1. Never panics on arbitrary input.
//  2. Symmetry: Distance(a, b) == Distance(b, a).
//  3. Identity: Distance(a, a) == 0.
//
// Threat model: T-08.5-24 (D - DoS via fuzz-discovered panic).

package fuzzymatch_test

import (
	"testing"

	"github.com/axonops/fuzzymatch"
)

// FuzzLevenshteinDistance asserts symmetry + identity for the
// Wagner-Fischer 1974 edit distance.
func FuzzLevenshteinDistance(f *testing.F) {
	for _, pair := range []struct{ a, b string }{
		{"kitten", "sitting"},
		{"saturday", "sunday"},
		{"", ""},
		{"", "abc"},
		{"abc", ""},
		{"café", "cafe"},
		{"\xff\xfe", "abc"},
		{"abc", "abc"},
	} {
		f.Add(pair.a, pair.b)
	}
	f.Fuzz(func(t *testing.T, a, b string) {
		dab := fuzzymatch.LevenshteinDistance(a, b)
		dba := fuzzymatch.LevenshteinDistance(b, a)
		if dab != dba {
			t.Fatalf("LevenshteinDistance not symmetric for (%q, %q): %d vs %d", a, b, dab, dba)
		}
		daa := fuzzymatch.LevenshteinDistance(a, a)
		if daa != 0 {
			t.Fatalf("LevenshteinDistance identity violated for %q: distance to self = %d", a, daa)
		}
	})
}
