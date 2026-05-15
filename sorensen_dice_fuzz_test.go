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

// sorensen_dice_fuzz_test.go runs native Go fuzzing against the two
// public Sørensen-Dice surfaces (byte + rune). Properties checked
// per surface, per input:
//
//   1. Never panics (implicit — any panic propagates as a fuzz crash).
//   2. Score never returns NaN.
//   3. Score never returns ±Inf.
//   4. Score returns a value in [0.0, 1.0].
//
// The n parameter is coerced into [1, 8] in the fuzz body so the
// documented n < 1 panic path is not exercised by the harness — the
// panic contract is unit-tested separately by
// TestSorensenDice_PanicsOnInvalidN.
//
// On-disk seed corpus: testdata/fuzz/FuzzSorensenDiceScore/seed-001
// and testdata/fuzz/FuzzSorensenDiceScoreRunes/seed-001 in the
// `go test fuzz v1` literal format. Programmatic f.Add seeds cover the
// canonical reference vectors (RV-D1..RV-D4), identity, both-empty,
// one-empty, invalid UTF-8, multi-byte UTF-8, and long-input cases.
//
// CI's nightly fuzz job runs each fuzzer for 60s+; locally run
// `go test -fuzz=FuzzSorensenDiceScore -fuzztime=10s ./...` for a
// smoke check.

package fuzzymatch_test

import (
	"math"
	"strings"
	"testing"

	"github.com/axonops/fuzzymatch"
)

// FuzzSorensenDiceScore exercises the byte-path public surface.
func FuzzSorensenDiceScore(f *testing.F) {
	// Programmatic seeds.
	for _, seed := range []struct {
		a, b string
		n    int
	}{
		{"night", "nacht", 2},   // RV-D1 canonical
		{"abcdef", "bcdefg", 2}, // RV-D2 high-overlap
		{"abcdef", "abcXef", 3}, // RV-D3 trigram
		{"hello", "hello", 2},   // RV-D4 identity
		{"", "", 2},             // both-empty
		{"abc", "", 2},          // one-empty
		{"abc", "xyz", 2},       // orthogonal
		{"\xff\xfe", "abc", 2},  // invalid UTF-8
		{"\xc0\x80", "abc", 2},  // overlong NUL
		{"café", "cafe", 2},     // multi-byte UTF-8
		{strings.Repeat("a", 200), strings.Repeat("ab", 100), 3}, // long input
		{"x", "x", 1},               // n=1 unigram
		{"abcdefgh", "abcdefgi", 8}, // n=8 max
	} {
		f.Add(seed.a, seed.b, seed.n)
	}

	f.Fuzz(func(t *testing.T, a, b string, n int) {
		nn := fuzzCoerceN(n)
		got := fuzzymatch.SorensenDiceScore(a, b, nn)
		if math.IsNaN(got) {
			t.Fatalf("SorensenDiceScore(%q, %q, %d) = NaN", a, b, nn)
		}
		if math.IsInf(got, 0) {
			t.Fatalf("SorensenDiceScore(%q, %q, %d) = Inf", a, b, nn)
		}
		if got < 0.0 || got > 1.0 {
			t.Fatalf("SorensenDiceScore(%q, %q, %d) = %g; want in [0, 1]", a, b, nn, got)
		}
	})
}

// FuzzSorensenDiceScoreRunes exercises the rune-path public surface.
func FuzzSorensenDiceScoreRunes(f *testing.F) {
	// Programmatic seeds — same shape as the byte-path harness; the
	// rune extractor processes invalid UTF-8 by replacing malformed
	// sequences with U+FFFD per Go's []rune conversion semantics.
	for _, seed := range []struct {
		a, b string
		n    int
	}{
		{"night", "nacht", 2},   // RV-D1 (ASCII; both paths align)
		{"abcdef", "bcdefg", 2}, // RV-D2 (ASCII)
		{"hello", "hello", 2},   // RV-D4 identity
		{"", "", 2},             // both-empty
		{"abc", "", 2},          // one-empty
		{"abc", "xyz", 2},       // orthogonal
		{"café", "cafe", 2},     // load-bearing rune-path canary
		{"Привет", "привет", 2}, // Cyrillic
		{"\xff\xfe", "abc", 2},  // invalid UTF-8 (FFFD-replaced)
		{strings.Repeat("a", 200), strings.Repeat("ab", 100), 3}, // long input
		{"x", "x", 1},               // n=1
		{"abcdefgh", "abcdefgi", 8}, // n=8
	} {
		f.Add(seed.a, seed.b, seed.n)
	}

	f.Fuzz(func(t *testing.T, a, b string, n int) {
		nn := fuzzCoerceN(n)
		got := fuzzymatch.SorensenDiceScoreRunes(a, b, nn)
		if math.IsNaN(got) {
			t.Fatalf("SorensenDiceScoreRunes(%q, %q, %d) = NaN", a, b, nn)
		}
		if math.IsInf(got, 0) {
			t.Fatalf("SorensenDiceScoreRunes(%q, %q, %d) = Inf", a, b, nn)
		}
		if got < 0.0 || got > 1.0 {
			t.Fatalf("SorensenDiceScoreRunes(%q, %q, %d) = %g; want in [0, 1]", a, b, nn, got)
		}
	})
}
