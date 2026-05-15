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

// nysiis_fuzz_test.go provides coverage-guided fuzz testing for
// NYSIISCode and NYSIISScore. Seeds cover both ASCII (the encoded regime)
// and mixed non-ASCII (the silent-skip regime) per CONTEXT.md §5.
//
// Invariants verified:
//  1. No panic on any input.
//  2. Output charset: matches ^[A-Z]{0,6}$ (uppercase ASCII letters only, max 6 chars).
//  3. Output length: len(NYSIISCode(s)) <= 6 for all s (Taft-1970 truncation — LOAD-BEARING).
//  4. Score range: NYSIISScore(a, b) ∈ {0.0, 1.0}.
//  5. No NaN, no Inf.
//  6. Identity: NYSIISScore(s, s) == 1.0 for any s.

package fuzzymatch_test

import (
	"math"
	"regexp"
	"testing"

	"github.com/axonops/fuzzymatch"
)

// FuzzNYSIIS exercises NYSIISCode and NYSIISScore on arbitrary byte inputs.
// The seed corpus covers the ASCII-only regime and the non-ASCII-silent-skip regime.
func FuzzNYSIIS(f *testing.F) {
	// ASCII regime seeds — literature reference vectors and typical English names.
	f.Add("Brown")
	f.Add("Browne")
	f.Add("Robert")
	f.Add("Catherine")
	f.Add("Katherine")
	f.Add("Johnathan")
	f.Add("Jonathan")
	f.Add("John")
	f.Add("Teresa")
	f.Add("Theresa")
	f.Add("montgomery")
	f.Add("")
	f.Add("A")
	f.Add("ZZ")
	f.Add("Smith")
	f.Add("Jones")

	// Non-ASCII silent-skip regime — per CONTEXT.md §5.
	f.Add("Müller")
	f.Add("Café")
	f.Add("中文")
	f.Add("🎉hello")
	f.Add("\xff\xfe") // invalid UTF-8

	re := regexp.MustCompile(`^[A-Z]{0,6}$`)

	f.Fuzz(func(t *testing.T, s string) {
		// Invariant 1, 2, 3: no panic; charset; length.
		code := fuzzymatch.NYSIISCode(s)
		if !re.MatchString(code) {
			t.Errorf("NYSIISCode(%q) = %q; want match ^[A-Z]{0,6}$", s, code)
		}
		if len(code) > 6 {
			t.Errorf("NYSIISCode(%q) = %q (len %d); want len <= 6 (Taft-1970 truncation gate)",
				s, code, len(code))
		}

		// Invariant 4, 5, 6: score range, no NaN/Inf, identity.
		score := fuzzymatch.NYSIISScore(s, s)
		if score != 1.0 {
			t.Errorf("NYSIISScore(%q, %q) = %v; want 1.0 (identity invariant)", s, s, score)
		}
		if math.IsNaN(score) || math.IsInf(score, 0) {
			t.Errorf("NYSIISScore(%q, %q) = %v; want finite", s, s, score)
		}
	})
}
