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

// soundex_fuzz_test.go runs native Go fuzzing against SoundexCode and
// SoundexScore. Properties checked per input:
//
//  1. Never panics (implicit — any panic propagates as a fuzz crash).
//  2. Score never returns NaN.
//  3. Score never returns ±Inf.
//  4. Score is exactly 0.0 or 1.0 (binary, not arbitrary float).
//  5. Identity: SoundexScore(x, x) == 1.0 for all x.
//  6. Code charset: SoundexCode(s) is empty OR matches [A-Z][0-9]{3} (4 chars).
//
// Seed corpus covers both the ASCII regime (literature reference vectors)
// and the non-ASCII silent-skip regime per CONTEXT.md §5.
//
// CI runs each fuzzer for 60s per build; locally run
// `go test -fuzz=FuzzSoundex -fuzztime=10s ./...` for a smoke check.

package fuzzymatch_test

import (
	"math"
	"testing"

	"github.com/axonops/fuzzymatch"
)

// FuzzSoundex exercises the public surface with programmatic seeds covering
// the load-bearing fixtures and a broad set of edge cases.
func FuzzSoundex(f *testing.F) {
	// ASCII regime seeds — literature reference vectors.
	for _, seed := range []struct{ a, b string }{
		{"Robert", "Rupert"},     // same code R163
		{"Tymczak", "Tymczak"},   // identity + Knuth/Census gate
		{"Ashcraft", "Ashcroft"}, // H/W-handling pair
		{"Smith", "Jones"},       // different codes
		{"", ""},                 // both-empty
		{"Robert", ""},           // one-empty
		{"", "Smith"},            // one-empty
		{"Lloyd", "Lloyd"},       // double-L collapse
		{"Pfister", "Pfister"},   // Pf pair
		// Non-ASCII silent-skip regime (CONTEXT.md §5).
		{"Müller", "Mueller"}, // ü dropped → Mller / Mueller
		{"Café", "Cafe"},      // é dropped → Cf / Cafe
		{"中文", ""},            // all non-ASCII → ""
		{"🎉hello", "hello"},   // emoji prefix skipped
		{"\xff\xfe", "abc"},   // invalid UTF-8
	} {
		f.Add(seed.a, seed.b)
	}

	f.Fuzz(func(t *testing.T, a, b string) {
		score := fuzzymatch.SoundexScore(a, b)
		if math.IsNaN(score) {
			t.Fatalf("SoundexScore(%q, %q) = NaN", a, b)
		}
		if math.IsInf(score, 0) {
			t.Fatalf("SoundexScore(%q, %q) = Inf", a, b)
		}
		if score != 0.0 && score != 1.0 {
			t.Fatalf("SoundexScore(%q, %q) = %g; Soundex score must be 0.0 or 1.0 (binary)", a, b, score)
		}

		// Identity: SoundexScore(a, a) must be exactly 1.0.
		if idScore := fuzzymatch.SoundexScore(a, a); idScore != 1.0 {
			t.Fatalf("SoundexScore(%q, %q) = %g; identity must be exactly 1.0", a, a, idScore)
		}

		// Code charset: output must be empty or exactly [A-Z][0-9]{3}.
		codeA := fuzzymatch.SoundexCode(a)
		if codeA != "" {
			if len(codeA) != 4 {
				t.Fatalf("SoundexCode(%q) = %q; len=%d; want 4 (1 letter + 3 digits)", a, codeA, len(codeA))
			}
			if codeA[0] < 'A' || codeA[0] > 'Z' {
				t.Fatalf("SoundexCode(%q) = %q; first char %q not in [A-Z]", a, codeA, codeA[0:1])
			}
			for i := 1; i < 4; i++ {
				if codeA[i] < '0' || codeA[i] > '9' {
					t.Fatalf("SoundexCode(%q) = %q; digit %d char %q not in [0-9]", a, codeA, i, codeA[i:i+1])
				}
			}
		}

		// Range: score in [0.0, 1.0] (redundant with binary check above, kept for defence-in-depth).
		if score < 0.0 || score > 1.0 {
			t.Fatalf("SoundexScore(%q, %q) = %g; want in [0.0, 1.0]", a, b, score)
		}
	})
}
