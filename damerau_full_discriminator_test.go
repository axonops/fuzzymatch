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

// damerau_full_discriminator_test.go is the Task 1 TDD canary for the
// Damerau-Levenshtein Full implementation (plan 02-06 §B-1). Its sole purpose
// is to gate the Lowrance-Wagner recurrence at Task 1 BEFORE the full
// reference-vector suite in damerau_full_test.go runs at Task 2.
//
// The test pins the discriminating-vector contract: "ca"/"abc" must return
// distance 2 under DL-Full (Lowrance-Wagner 1975). This value is the locked
// gate that distinguishes DL-Full (this plan) from DL-OSA (plan 02-05, which
// returns 3 for the same pair). If this test fails with got==3, the recurrence
// has collapsed into OSA semantics and must be corrected before proceeding to
// Task 2.
//
// The full reference-vector suite lives in damerau_full_test.go (Task 2).
// Stdlib testing only — no testify.

package fuzzymatch_test

import (
	"testing"

	"github.com/axonops/fuzzymatch"
)

// TestDamerauLevenshteinFull_DiscriminatingVector_Stub asserts the locked
// discriminating-vector contract: DamerauLevenshteinFullDistance("ca", "abc")
// must return 2 (NOT 3 — that is DL-OSA's value for the same pair).
//
// Lowrance-Wagner 1975 discriminating vector:
// DL-Full permits unrestricted transpositions — any pair of adjacent
// characters may be transposed and subsequently edited. Transforming
// "ca" → "abc" under Full DL requires only 2 edits (one transposition +
// one insertion), whereas OSA requires 3 because the OSA restriction
// forbids re-editing characters after a transposition.
//
// If this test fails with got==3, the recurrence collapsed into OSA semantics.
func TestDamerauLevenshteinFull_DiscriminatingVector_Stub(t *testing.T) {
	// Lowrance-Wagner 1975 discriminating vector: "ca" vs "abc" must
	// return 2. OSA (plan 02-05) returns 3 for the same pair. If this
	// test fails with got==3, the recurrence collapsed into OSA semantics.
	got := fuzzymatch.DamerauLevenshteinFullDistance("ca", "abc")
	if got != 2 {
		t.Fatalf("DamerauLevenshteinFullDistance(\"ca\",\"abc\") = %d, want 2 (NOT 3 — that is OSA's value)", got)
	}
	if got == 3 {
		t.Fatalf("recurrence collapsed into OSA semantics")
	}
}
