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

// tversky_fuzz_test.go runs native Go fuzzing against the two public
// Tversky surfaces (byte + rune). Properties checked per surface, per
// input:
//
//   1. Never panics (implicit — any panic propagates as a fuzz crash).
//   2. Score never returns NaN.
//   3. Score never returns ±Inf.
//   4. Score returns a value in [0.0, 1.0].
//
// The n parameter is coerced into [1, 8] (via fuzzCoerceN from the
// Q-Gram Jaccard fuzz harness — the project-wide n-coerce helper) AND
// the α, β weights are coerced into [0.0, 1.0] with α + β > 0
// enforced in the fuzz body so the documented panic paths are not
// exercised by the harness — those contracts are unit-tested
// separately by TestTversky_PanicsOnInvalidN and
// TestTversky_PanicsOnInvalidParams.
//
// Future hardening (deferred): NaN α / β inputs are not directly
// exercised by this harness because the squashing function maps NaN
// to a safe interior point. If a NaN-input fuzz finding emerges in
// practice, the fuzz body can be loosened to pass NaN through and the
// public-API gate updated to detect it.
//
// On-disk seed corpus: testdata/fuzz/FuzzTverskyScore/seed-001 and
// testdata/fuzz/FuzzTverskyScoreRunes/seed-001 in the `go test fuzz
// v1` literal format. The seed encodes the canonical RV-T1 pair
// ("abcd", "abcdef", n=2, α=0.8, β=0.2) — the load-bearing asymmetric
// configuration; rerunning the harness with the seed exercises the
// asymmetric arithmetic path on the exact canonical input.
//
// Programmatic f.Add seeds cover the full reference-vector slate
// (RV-T1..RV-T4), identity, both-empty, one-empty, orthogonal,
// invalid UTF-8, multi-byte UTF-8, long-input, and edge α/β values
// (0.0, 1.0).
//
// CI's nightly fuzz job runs each fuzzer for 60s+; locally run
// `go test -fuzz=FuzzTverskyScore -fuzztime=10s ./...` for a smoke
// check.

package fuzzymatch_test

import (
	"math"
	"strings"
	"testing"

	"github.com/axonops/fuzzymatch"
)

// fuzzCoerceTverskyParam coerces an arbitrary float64 into [0.0, 1.0]
// inclusive. NaN inputs map to 0.5 (a safe interior point); ±Inf maps
// to 1.0; otherwise the squash `|x| / (|x| + 1)` is used which is
// monotonic in |x| and bounded by [0, 1).
//
// This is a separate helper from props_test.go's tverskyAlpha because
// the fuzz harness cannot use test-only helpers from a different
// _test.go file IN PRINCIPLE (Go test files SHARE the same package, so
// they CAN share helpers — but keeping the harness self-contained
// helps when reading a single fuzz file in isolation).
func fuzzCoerceTverskyParam(p float64) float64 {
	if math.IsNaN(p) {
		return 0.5
	}
	if math.IsInf(p, 0) {
		return 1.0
	}
	p = math.Abs(p)
	return p / (p + 1.0)
}

// fuzzCoerceTverskyParams coerces (α, β) into the valid-parameter
// region: both in [0, 1], α + β > 0. If both squash to 0.0, force
// (α, β) = (1.0, 0.0) — a documented valid boundary.
func fuzzCoerceTverskyParams(alpha, beta float64) (float64, float64) {
	a := fuzzCoerceTverskyParam(alpha)
	b := fuzzCoerceTverskyParam(beta)
	if a == 0.0 && b == 0.0 {
		return 1.0, 0.0
	}
	return a, b
}

// FuzzTverskyScore exercises the byte-path public surface.
func FuzzTverskyScore(f *testing.F) {
	// Programmatic seeds covering the canonical reference-vector
	// slate plus edge cases.
	for _, seed := range []struct {
		a, b        string
		n           int
		alpha, beta float64
	}{
		{"abcd", "abcdef", 2, 0.8, 0.2},                                    // RV-T1 canonical asymmetric
		{"abcdef", "abcd", 2, 0.8, 0.2},                                    // RV-T2 input-swap pair
		{"abcd", "abce", 2, 1.0, 1.0},                                      // RV-T3 Jaccard cross-check
		{"abcd", "abce", 2, 0.5, 0.5},                                      // RV-T4 Dice cross-check
		{"hello", "hello", 2, 0.8, 0.2},                                    // identity
		{"", "", 2, 0.5, 0.5},                                              // both-empty
		{"abc", "", 2, 0.5, 0.5},                                           // one-empty
		{"abc", "xyz", 2, 0.5, 0.5},                                        // orthogonal
		{"abcd", "abxy", 2, 0.7, 0.3},                                      // single-shared
		{"\xff\xfe", "abc", 2, 0.5, 0.5},                                   // invalid UTF-8
		{"\xc0\x80", "abc", 2, 0.5, 0.5},                                   // overlong NUL
		{"café", "cafe", 2, 0.5, 0.5},                                      // multi-byte UTF-8
		{strings.Repeat("a", 200), strings.Repeat("ab", 100), 3, 0.4, 0.6}, // long input
		{"x", "x", 1, 0.8, 0.2},                                            // n=1 unigram identity
		{"abcdefgh", "abcdefgi", 8, 1.0, 0.0},                              // n=8 max + α-only
		{"abcdefgh", "abcdefgi", 3, 0.0, 1.0},                              // β-only edge (α=0 valid)
	} {
		f.Add(seed.a, seed.b, seed.n, seed.alpha, seed.beta)
	}

	f.Fuzz(func(t *testing.T, a, b string, n int, alpha, beta float64) {
		nn := fuzzCoerceN(n)
		al, be := fuzzCoerceTverskyParams(alpha, beta)
		got := fuzzymatch.TverskyScore(a, b, nn, al, be)
		if math.IsNaN(got) {
			t.Fatalf("TverskyScore(%q, %q, %d, %g, %g) = NaN", a, b, nn, al, be)
		}
		if math.IsInf(got, 0) {
			t.Fatalf("TverskyScore(%q, %q, %d, %g, %g) = Inf", a, b, nn, al, be)
		}
		if got < 0.0 || got > 1.0 {
			t.Fatalf("TverskyScore(%q, %q, %d, %g, %g) = %g; want in [0, 1]", a, b, nn, al, be, got)
		}
	})
}

// FuzzTverskyScoreRunes exercises the rune-path public surface.
func FuzzTverskyScoreRunes(f *testing.F) {
	// Programmatic seeds — same shape as the byte-path harness; the
	// rune extractor processes invalid UTF-8 by replacing malformed
	// sequences with U+FFFD per Go's []rune conversion semantics.
	for _, seed := range []struct {
		a, b        string
		n           int
		alpha, beta float64
	}{
		{"abcd", "abcdef", 2, 0.8, 0.2},                                    // RV-T1 (ASCII; both paths align)
		{"abcdef", "abcd", 2, 0.8, 0.2},                                    // RV-T2 input swap
		{"hello", "hello", 2, 0.8, 0.2},                                    // identity
		{"", "", 2, 0.5, 0.5},                                              // both-empty
		{"abc", "", 2, 0.5, 0.5},                                           // one-empty
		{"abc", "xyz", 2, 0.5, 0.5},                                        // orthogonal
		{"café", "cafe", 2, 0.5, 0.5},                                      // load-bearing rune-path canary
		{"Привет", "привет", 2, 0.5, 0.5},                                  // Cyrillic
		{"\xff\xfe", "abc", 2, 0.5, 0.5},                                   // invalid UTF-8 (FFFD-replaced)
		{strings.Repeat("a", 200), strings.Repeat("ab", 100), 3, 0.4, 0.6}, // long input
		{"x", "x", 1, 0.8, 0.2},                                            // n=1
		{"abcdefgh", "abcdefgi", 8, 1.0, 0.0},                              // n=8 max + β=0 boundary
	} {
		f.Add(seed.a, seed.b, seed.n, seed.alpha, seed.beta)
	}

	f.Fuzz(func(t *testing.T, a, b string, n int, alpha, beta float64) {
		nn := fuzzCoerceN(n)
		al, be := fuzzCoerceTverskyParams(alpha, beta)
		got := fuzzymatch.TverskyScoreRunes(a, b, nn, al, be)
		if math.IsNaN(got) {
			t.Fatalf("TverskyScoreRunes(%q, %q, %d, %g, %g) = NaN", a, b, nn, al, be)
		}
		if math.IsInf(got, 0) {
			t.Fatalf("TverskyScoreRunes(%q, %q, %d, %g, %g) = Inf", a, b, nn, al, be)
		}
		if got < 0.0 || got > 1.0 {
			t.Fatalf("TverskyScoreRunes(%q, %q, %d, %g, %g) = %g; want in [0, 1]", a, b, nn, al, be, got)
		}
	})
}
