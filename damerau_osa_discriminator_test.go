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

// damerau_osa_discriminator_test.go is the Task 1 TDD canary for the
// Damerau-Levenshtein OSA implementation (plan 02-05 §B-1). Its sole purpose
// is to gate the OSA recurrence at Task 1 BEFORE the full reference-vector
// suite in damerau_osa_test.go runs at Task 2.
//
// The test pins the discriminating-vector contract: "ca"/"abc" must return
// distance 3 under OSA. This value is the locked gate that distinguishes
// DL-OSA (this plan) from DL-Full (plan 02-06, which returns 2 for the same
// pair). If this test fails with got==2, the recurrence has collapsed into
// Full DL semantics and must be corrected before proceeding to Task 2.
//
// The full reference-vector suite lives in damerau_osa_test.go (Task 2).
// Stdlib testing only — no testify.

package fuzzymatch_test

import (
	"testing"

	"github.com/axonops/fuzzymatch"
)

// TestDamerauLevenshteinOSA_DiscriminatingVector_Stub asserts the locked
// discriminating-vector contract: DamerauLevenshteinOSADistance("ca", "abc")
// must return 3 (NOT 2 — that is DL-Full's value for the same pair).
//
// Boytsov 2011 §3.1 / Damerau 1964 discriminating vector:
// The OSA restriction forbids re-editing characters after a transposition.
// Transforming "ca" → "abc" under OSA requires 3 edits; Full DL requires
// only 2 because it can re-use the transposed characters in further edits.
func TestDamerauLevenshteinOSA_DiscriminatingVector_Stub(t *testing.T) {
	// Boytsov 2011 §3.1 / Damerau 1964 discriminating vector:
	// "ca" vs "abc" must return 3 under OSA (because the OSA
	// restriction forbids re-editing after transposition). DL-Full
	// (plan 02-06) returns 2 for the same pair. If this test fails
	// with got==2, the recurrence collapsed into Full DL semantics.
	got := fuzzymatch.DamerauLevenshteinOSADistance("ca", "abc")
	if got != 3 {
		t.Fatalf("DamerauLevenshteinOSADistance(\"ca\",\"abc\") = %d, want 3 (NOT 2 — that is DL-Full's value)", got)
	}
}
