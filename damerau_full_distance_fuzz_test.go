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

// damerau_full_distance_fuzz_test.go runs native Go fuzzing against
// DamerauLevenshteinFullDistance (Lowrance-Wagner 1975 unrestricted
// transpositions). Asserts symmetry + identity.
//
// Threat model: T-08.5-24 (D - DoS via fuzz-discovered panic).

package fuzzymatch_test

import (
	"testing"

	"github.com/axonops/fuzzymatch"
)

// FuzzDamerauLevenshteinFullDistance asserts symmetry + identity for
// the Lowrance-Wagner unrestricted-transposition distance.
func FuzzDamerauLevenshteinFullDistance(f *testing.F) {
	for _, pair := range []struct{ a, b string }{
		{"ca", "abc"},    // Lowrance-Wagner discriminating vector (returns 2)
		{"abcd", "abdc"}, // single adjacent transposition
		{"", ""},
		{"", "abc"},
		{"abc", "abc"},
		{"café", "caéf"},
		{"\xff\xfe", "abc"},
	} {
		f.Add(pair.a, pair.b)
	}
	f.Fuzz(func(t *testing.T, a, b string) {
		dab := fuzzymatch.DamerauLevenshteinFullDistance(a, b)
		dba := fuzzymatch.DamerauLevenshteinFullDistance(b, a)
		if dab != dba {
			t.Fatalf("DamerauLevenshteinFullDistance not symmetric for (%q, %q): %d vs %d", a, b, dab, dba)
		}
		daa := fuzzymatch.DamerauLevenshteinFullDistance(a, a)
		if daa != 0 {
			t.Fatalf("DamerauLevenshteinFullDistance identity violated for %q: distance to self = %d", a, daa)
		}
	})
}
