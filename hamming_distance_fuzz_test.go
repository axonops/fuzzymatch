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

// hamming_distance_fuzz_test.go runs native Go fuzzing against
// HammingDistance (Hamming 1950). Asserts symmetry + identity.
//
// Phase 8.5 Q1 LOCKED silent-max policy: on unequal-length inputs,
// HammingDistance returns max(len(a), len(b)) rather than panicking
// or returning an error tuple. The symmetry assertion still holds:
// max(len(a), len(b)) == max(len(b), len(a)).
//
// Threat model: T-08.5-24 (D - DoS via fuzz-discovered panic).

package fuzzymatch_test

import (
	"testing"

	"github.com/axonops/fuzzymatch"
)

// FuzzHammingDistance asserts symmetry + identity for the Hamming
// distance, including the unequal-length silent-max path.
func FuzzHammingDistance(f *testing.F) {
	for _, pair := range []struct{ a, b string }{
		{"karolin", "kathrin"},   // Hamming 1950 canonical (dist 3)
		{"1011101", "1001001"},   // Hamming 1950 canonical (dist 2)
		{"abc", "abc"},           // identity
		{"", ""},                 // both-empty
		{"abc", "ab"},            // unequal-length (silent-max: returns max=3)
		{"ab", "abc"},            // reversed silent-max (returns max=3)
		{"\xff\xfe", "\xff\xff"}, // invalid UTF-8, equal byte count
	} {
		f.Add(pair.a, pair.b)
	}
	f.Fuzz(func(t *testing.T, a, b string) {
		dab := fuzzymatch.HammingDistance(a, b)
		dba := fuzzymatch.HammingDistance(b, a)
		if dab != dba {
			t.Fatalf("HammingDistance not symmetric for (%q, %q): %d vs %d", a, b, dab, dba)
		}
		daa := fuzzymatch.HammingDistance(a, a)
		if daa != 0 {
			t.Fatalf("HammingDistance identity violated for %q: distance to self = %d", a, daa)
		}
	})
}
