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

// double_metaphone_fuzz_test.go runs native Go fuzzing against
// DoubleMetaphoneKeys and DoubleMetaphoneScore. Properties checked per input:
//
//  1. Never panics (implicit — any panic propagates as a fuzz crash).
//  2. Score never returns NaN.
//  3. Score never returns ±Inf.
//  4. Score is exactly 0.0 or 1.0 (binary, not arbitrary float).
//  5. Identity: DoubleMetaphoneScore(x, x) == 1.0 for all x.
//  6. Key charset BOTH keys: ^[A-Z0]{0,4}$ (each key ≤ 4 chars, only [A-Z0]).
//
// Seed corpus covers both the ASCII regime (literature reference vectors)
// and the non-ASCII silent-skip regime per CONTEXT.md §5.
//
// CI runs each fuzzer for 60s per build; locally run
// `go test -fuzz=FuzzDoubleMetaphone -fuzztime=10s ./...` for a smoke check.

package fuzzymatch_test

import (
	"math"
	"testing"

	"github.com/axonops/fuzzymatch"
)

// FuzzDoubleMetaphone exercises the public surface with programmatic seeds
// covering all 5 language-branch fixtures and non-ASCII silent-skip cases.
func FuzzDoubleMetaphone(f *testing.F) {
	// ASCII regime — all 5 language-branch mandatory fixtures.
	for _, seed := range []struct{ a, b string }{
		// Germanic (CONTEXT.md §3 mandatory):
		{"Schmidt", "Smith"}, // XMT cross-match
		{"Schwartz", "Schwartz"},
		// Slavic:
		{"Sczepanski", "Dvorak"},
		{"Wojcik", "Wojcik"},
		// Romance:
		{"Pacheco", "Pacheco"}, // PXK gate
		{"Jaramillo", "Bologna"},
		// Greek (CONTEXT.md §3 mandatory):
		{"Catherine", "Katherine"}, // K0RN match
		{"Christopher", "Christopher"},
		// Chinese-origin:
		{"Cheung", "Wong"},
		{"Chen", "Hong"},
		// Edge:
		{"", ""},
		{"Schmidt", ""},
		{"", "Smith"},
		{"Caesar", "Caesar"},
		{"Knock", "Knock"},
		// Non-ASCII silent-skip regime (CONTEXT.md §5):
		{"Müller", "Mueller"},
		{"Café", "Cafe"},
		{"中文", ""},
		{"🎉hello", "hello"},
		{"\xff\xfe", "abc"},
	} {
		f.Add(seed.a, seed.b)
	}

	f.Fuzz(func(t *testing.T, a, b string) {
		score := fuzzymatch.DoubleMetaphoneScore(a, b)

		// Invariant 2: no NaN
		if math.IsNaN(score) {
			t.Fatalf("DoubleMetaphoneScore(%q, %q) = NaN", a, b)
		}
		// Invariant 3: no Inf
		if math.IsInf(score, 0) {
			t.Fatalf("DoubleMetaphoneScore(%q, %q) = Inf", a, b)
		}
		// Invariant 4: binary
		if score != 0.0 && score != 1.0 {
			t.Fatalf("DoubleMetaphoneScore(%q, %q) = %g; must be 0.0 or 1.0 (binary)", a, b, score)
		}
		// Invariant 5: identity
		if idScore := fuzzymatch.DoubleMetaphoneScore(a, a); idScore != 1.0 {
			t.Fatalf("DoubleMetaphoneScore(%q, %q) = %g; identity must be 1.0", a, a, idScore)
		}

		// Invariant 6: key charset [A-Z0]{0,4} for BOTH keys
		for _, input := range []string{a, b} {
			primary, secondary := fuzzymatch.DoubleMetaphoneKeys(input)
			for _, key := range []string{primary, secondary} {
				if len(key) > 4 {
					t.Fatalf("DoubleMetaphoneKeys(%q) key %q has len %d; want ≤ 4", input, key, len(key))
				}
				for j := 0; j < len(key); j++ {
					c := key[j]
					if (c < 'A' || c > 'Z') && c != '0' {
						t.Fatalf("DoubleMetaphoneKeys(%q) key %q contains invalid char %q (must be [A-Z0])",
							input, key, string(c))
					}
				}
			}
		}

		// Range invariant (redundant with binary, kept for defence-in-depth):
		if score < 0.0 || score > 1.0 {
			t.Fatalf("DoubleMetaphoneScore(%q, %q) = %g; want in [0.0, 1.0]", a, b, score)
		}
	})
}
