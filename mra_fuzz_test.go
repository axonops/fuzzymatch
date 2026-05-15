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

// mra_fuzz_test.go provides coverage-guided fuzz testing for MRACode,
// MRACompare, and MRAScore. Seeds cover both ASCII (the encoded regime)
// and mixed non-ASCII (the silent-skip regime) per CONTEXT.md §5.
//
// MRA-specific invariants verified (in addition to no-panic):
//  1. MRACode charset: matches ^[A-Z]{0,6}$ (uppercase ASCII letters, max 6 chars).
//  2. MRACompare integer range: 0 <= simScore <= 6.
//  3. MRAScore binary: score ∈ {0.0, 1.0} strictly.
//  4. MRAScore-Compare consistency: (MRAScore(a, b) == 1.0) == MRACompare(a, b).matched.

package fuzzymatch_test

import (
	"math"
	"regexp"
	"testing"

	"github.com/axonops/fuzzymatch"
)

// FuzzMRA exercises MRACode, MRACompare, and MRAScore on arbitrary byte inputs.
// The seed corpus covers the ASCII-only regime and the non-ASCII-silent-skip regime
// per CONTEXT.md §5, plus threshold-edge cases and length-difference cases.
func FuzzMRA(f *testing.F) {
	// ASCII regime seeds — literature reference vectors and typical English names.
	f.Add("Byrne", "Boern")
	f.Add("Smith", "Smyth")
	f.Add("Catherine", "Katherine")
	f.Add("William", "Willyam")
	f.Add("Robert", "Robin")
	f.Add("Brown", "Browne")
	f.Add("Kathrynoglin", "Kathrynoglin")
	f.Add("", "")
	f.Add("A", "A")
	f.Add("Smith", "Smith")
	f.Add("John", "Jon")
	f.Add("James", "Jameson")

	// Length-difference >= 3 pairs (auto-mismatch gate).
	f.Add("Ad", "ZachariahMontgomery")
	f.Add("A", "Kathrynoglin")

	// Threshold-edge pairs (sum_len near table boundaries: 4, 7, 11, 12).
	f.Add("AB", "CD")       // sum_len=4 → threshold 5
	f.Add("ABC", "DEFG")    // sum_len=7 → threshold 4
	f.Add("ABCDE", "FGHIJ") // sum_len=10 → threshold 3

	// Non-ASCII silent-skip regime — per CONTEXT.md §5.
	f.Add("Müller", "Miller")
	f.Add("Café", "Cafe")
	f.Add("中文", "Chinese")
	f.Add("🎉hello", "hello")
	f.Add("\xff\xfe", "") // invalid UTF-8

	re := regexp.MustCompile(`^[A-Z]{0,6}$`)

	f.Fuzz(func(t *testing.T, a, b string) {
		// Invariant 1: MRACode charset — no panic, output matches ^[A-Z]{0,6}$.
		codeA := fuzzymatch.MRACode(a)
		codeB := fuzzymatch.MRACode(b)
		if !re.MatchString(codeA) {
			t.Errorf("MRACode(%q) = %q; want match ^[A-Z]{0,6}$", a, codeA)
		}
		if !re.MatchString(codeB) {
			t.Errorf("MRACode(%q) = %q; want match ^[A-Z]{0,6}$", b, codeB)
		}

		// Invariant 2: MRACompare integer range — 0 <= simScore <= 6.
		matched, sim := fuzzymatch.MRACompare(a, b)
		if sim < 0 || sim > 6 {
			t.Errorf("MRACompare(%q, %q).simScore = %d; want 0 <= sim <= 6", a, b, sim)
		}

		// Invariant 3: MRAScore binary — result must be exactly 0.0 or 1.0.
		score := fuzzymatch.MRAScore(a, b)
		if score != 0.0 && score != 1.0 {
			t.Errorf("MRAScore(%q, %q) = %v; want exactly 0.0 or 1.0", a, b, score)
		}
		if math.IsNaN(score) || math.IsInf(score, 0) {
			t.Errorf("MRAScore(%q, %q) = %v; want finite", a, b, score)
		}

		// Invariant 4: MRAScore-Compare consistency.
		// (MRAScore(a, b) == 1.0) must equal (MRACompare(a, b).matched).
		scoreIs1 := score == 1.0
		if scoreIs1 != matched {
			t.Errorf("consistency violation: MRAScore(%q, %q)=%v (is1.0=%v) but MRACompare.matched=%v",
				a, b, score, scoreIs1, matched)
		}
	})
}
