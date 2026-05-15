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

// partial_ratio_fuzz_test.go runs native Go fuzzing against the two
// public Partial Ratio surfaces (byte + rune). Properties checked per
// surface, per input:
//
//   1. Never panics (implicit — any panic propagates as a fuzz crash).
//   2. Score never returns NaN.
//   3. Score never returns ±Inf.
//   4. Score returns a value in [0.0, 1.0].
//   5. Symmetric across argument order — PR(a,b) == PR(b,a) (bit-for-bit).
//
// Programmatic f.Add seeds cover the canonical reference vectors
// (RV), identity, both-empty, one-empty, Pitfall-3 keystone fixtures,
// disjoint, invalid UTF-8, multi-byte UTF-8, and the pathological
// length-mismatch shape.
//
// CI's nightly fuzz job runs each fuzzer for 60s+; locally run
// `go test -fuzz=FuzzPartialRatioScore -fuzztime=10s ./...` for a
// smoke check.

package fuzzymatch_test

import (
	"math"
	"strings"
	"testing"

	"github.com/axonops/fuzzymatch"
)

// FuzzPartialRatioScore exercises the byte-path public surface.
func FuzzPartialRatioScore(f *testing.F) {
	// Programmatic seeds — 11 cases covering identity, both-empty,
	// one-empty, the two Pitfall-3 keystones, Region 2 middle,
	// disjoint, invalid UTF-8, multi-byte UTF-8, and the pathological
	// length-mismatch shape (truncated to fit in a fuzz seed).
	for _, seed := range []struct{ a, b string }{
		{"abc", "abc"},                  // identity
		{"", ""},                        // both-empty
		{"abc", ""},                     // one-empty
		{"", "abc"},                     // one-empty
		{"abc", "bc"},                   // Pitfall-3 keystone (Region 3 right tail)
		{"abc", "ab"},                   // Pitfall-3 keystone (Region 1 left tail)
		{"YANKEES", "NEW YORK YANKEES"}, // Region 2 middle wins
		{"abc", "xyzzz"},                // disjoint
		{"\xff\xfe", "abc"},             // invalid UTF-8
		{"café", "cafe"},                // multi-byte UTF-8
		{"abcdefghij", strings.Repeat("xyz", 100) + "j"}, // pathological-shape (smaller for fuzz speed)
	} {
		f.Add(seed.a, seed.b)
	}

	f.Fuzz(func(t *testing.T, a, b string) {
		got := fuzzymatch.PartialRatioScore(a, b)
		if math.IsNaN(got) {
			t.Fatalf("PartialRatioScore(%q, %q) = NaN", a, b)
		}
		if math.IsInf(got, 0) {
			t.Fatalf("PartialRatioScore(%q, %q) = Inf", a, b)
		}
		if got < 0.0 || got > 1.0 {
			t.Fatalf("PartialRatioScore(%q, %q) = %g; want in [0, 1]", a, b, got)
		}
		// Symmetry regression gate — PR(a, b) == PR(b, a) bit-for-bit.
		rev := fuzzymatch.PartialRatioScore(b, a)
		if got != rev {
			t.Fatalf("PartialRatioScore not symmetric on fuzzed input: PR(%q,%q)=%g, PR(%q,%q)=%g",
				a, b, got, b, a, rev)
		}
	})
}

// FuzzPartialRatioScoreRunes exercises the rune-path public surface.
// The rune path processes invalid UTF-8 by replacing malformed
// sequences with U+FFFD per Go's []rune conversion semantics.
func FuzzPartialRatioScoreRunes(f *testing.F) {
	for _, seed := range []struct{ a, b string }{
		{"abc", "abc"},       // identity ASCII
		{"café", "café"},     // identity Unicode
		{"", ""},             // both-empty
		{"abc", ""},          // one-empty
		{"", "abc"},          // one-empty
		{"abc", "bc"},        // Pitfall-3 keystone
		{"abc", "ab"},        // Pitfall-3 keystone
		{"café", "caf"},      // Unicode-specific keystone (rune-aware)
		{"αβγδ", "βγ"},       // pure non-ASCII (Greek)
		{"\xff\xfe", "abc"},  // invalid UTF-8 (FFFD-replaced)
		{"Привет", "привет"}, // Cyrillic mixed case
	} {
		f.Add(seed.a, seed.b)
	}

	f.Fuzz(func(t *testing.T, a, b string) {
		got := fuzzymatch.PartialRatioScoreRunes(a, b)
		if math.IsNaN(got) {
			t.Fatalf("PartialRatioScoreRunes(%q, %q) = NaN", a, b)
		}
		if math.IsInf(got, 0) {
			t.Fatalf("PartialRatioScoreRunes(%q, %q) = Inf", a, b)
		}
		if got < 0.0 || got > 1.0 {
			t.Fatalf("PartialRatioScoreRunes(%q, %q) = %g; want in [0, 1]", a, b, got)
		}
		rev := fuzzymatch.PartialRatioScoreRunes(b, a)
		if got != rev {
			t.Fatalf("PartialRatioScoreRunes not symmetric on fuzzed input: PR(%q,%q)=%g, PR(%q,%q)=%g",
				a, b, got, b, a, rev)
		}
	})
}
