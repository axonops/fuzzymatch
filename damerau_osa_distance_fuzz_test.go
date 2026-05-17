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

// damerau_osa_distance_fuzz_test.go runs native Go fuzzing against
// DamerauLevenshteinOSADistance (Boytsov 2011 / Damerau 1964 OSA).
// Asserts symmetry + identity.
//
// Note on symmetry: OSA distance IS symmetric (the recurrence is
// symmetric in the swap of i and j; only the no-re-edit restriction
// distinguishes OSA from Full DL). The "ca"/"abc" discriminating
// vector returns 3 in BOTH directions: d("ca","abc") == d("abc","ca") == 3.
//
// Threat model: T-08.5-24 (D - DoS via fuzz-discovered panic).

package fuzzymatch_test

import (
	"testing"

	"github.com/axonops/fuzzymatch"
)

// FuzzDamerauLevenshteinOSADistance asserts symmetry + identity for
// the OSA distance.
func FuzzDamerauLevenshteinOSADistance(f *testing.F) {
	for _, pair := range []struct{ a, b string }{
		{"ca", "abc"},    // Boytsov 2011 §3.1 discriminating vector (returns 3)
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
		dab := fuzzymatch.DamerauLevenshteinOSADistance(a, b)
		dba := fuzzymatch.DamerauLevenshteinOSADistance(b, a)
		if dab != dba {
			t.Fatalf("DamerauLevenshteinOSADistance not symmetric for (%q, %q): %d vs %d", a, b, dab, dba)
		}
		daa := fuzzymatch.DamerauLevenshteinOSADistance(a, a)
		if daa != 0 {
			t.Fatalf("DamerauLevenshteinOSADistance identity violated for %q: distance to self = %d", a, daa)
		}
	})
}
