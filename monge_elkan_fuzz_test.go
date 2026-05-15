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

// monge_elkan_fuzz_test.go runs native Go fuzzing against the two
// public Monge-Elkan surfaces. Properties checked per input:
//
//   1. Never panics (implicit — any panic propagates as a fuzz crash).
//   2. Score never returns NaN.
//   3. Score never returns ±Inf.
//   4. Score returns a value in [0.0, 1.0].
//   5. Identity short-circuit holds — MongeElkanScore(x, x, inner) ==
//      1.0 for all x and all permitted inner.
//
// The inner AlgoID is COERCED via fuzzCoerceMongeElkanInner so the fuzz
// harness never exercises the documented panic path — that contract is
// unit-tested separately by TestMongeElkan_PanicsOnNonPermittedInner.
// The harness's job is to exercise the score-computation surface, not
// the panic surface.
//
// On-disk seed corpus: programmatic f.Add seeds in this file cover the
// hand-derived reference vectors, identity, both-empty, one-empty,
// orthogonal sets, invalid UTF-8, and multi-byte UTF-8 cases.
//
// CI's nightly fuzz job runs each fuzzer for 60s+; locally run
// `go test -fuzz=FuzzMongeElkanScore -fuzztime=10s ./...` for a smoke
// check.

package fuzzymatch_test

import (
	"math"
	"strings"
	"testing"

	"github.com/axonops/fuzzymatch"
)

// fuzzCoerceMongeElkanInner coerces an arbitrary int input from the
// fuzz harness into one of the 14 permitted Monge-Elkan inner AlgoIDs.
// The harness never produces a non-permitted AlgoID — the panic
// contract is unit-tested separately. Mirrors the q-gram tier's
// fuzzCoerceN pattern (modular reduction over the valid range).
func fuzzCoerceMongeElkanInner(n int) fuzzymatch.AlgoID {
	permitted := []fuzzymatch.AlgoID{
		fuzzymatch.AlgoLevenshtein,
		fuzzymatch.AlgoDamerauLevenshteinOSA,
		fuzzymatch.AlgoDamerauLevenshteinFull,
		fuzzymatch.AlgoHamming,
		fuzzymatch.AlgoJaro,
		fuzzymatch.AlgoJaroWinkler,
		fuzzymatch.AlgoStrcmp95,
		fuzzymatch.AlgoSmithWatermanGotoh,
		fuzzymatch.AlgoLCSStr,
		fuzzymatch.AlgoQGramJaccard,
		fuzzymatch.AlgoSorensenDice,
		fuzzymatch.AlgoCosine,
		fuzzymatch.AlgoTversky,
		fuzzymatch.AlgoRatcliffObershelp,
	}
	if n < 0 {
		n = -n
	}
	return permitted[n%len(permitted)]
}

// FuzzMongeElkanScore exercises the ASYMMETRIC direct surface with
// programmatic seeds covering the six reference vectors and a broad
// set of edge cases. The fuzz body asserts the four invariants (no
// panic, no NaN, no ±Inf, score in [0, 1]) plus the identity
// short-circuit.
func FuzzMongeElkanScore(f *testing.F) {
	for _, seed := range []struct {
		a, b string
		inn  int
	}{
		{"user create", "usr creating", 5},                       // RV-ME1 (JaroWinkler=5 in our coercer)
		{"alpha beta", "alpha beta", 5},                          // RV-ME2 identity
		{"alpha beta", "gamma delta", 5},                         // RV-ME3 disjoint
		{"alpha", "alpha beta gamma", 0},                         // RV-ME4 subset (Levenshtein=0)
		{"café", "cafe", 0},                                      // RV-ME5 Unicode
		{"alpha beta gamma", "alpha", 0},                         // RV-ME6 asymmetry keystone
		{"", "", 5},                                              // both-empty
		{"hello", "", 5},                                         // one-empty
		{"", "hello", 5},                                         // one-empty
		{"\xff\xfe", "abc", 5},                                   // invalid UTF-8
		{"\xc0\x80", "abc", 5},                                   // overlong NUL
		{"café münchen", "münchen wien", 5},                      // multi-byte UTF-8
		{strings.Repeat("a ", 50), strings.Repeat("a ", 100), 0}, // dedup-heavy
		{"___", "...", 5},                                        // pure separators (both Tokenise empty)
		{"userID", "user_id", 5},                                 // identifier-style
	} {
		f.Add(seed.a, seed.b, seed.inn)
	}

	f.Fuzz(func(t *testing.T, a, b string, innRaw int) {
		inner := fuzzCoerceMongeElkanInner(innRaw)
		opts := fuzzymatch.DefaultNormalisationOptions()
		got := fuzzymatch.MongeElkanScore(a, b, inner, opts)
		if math.IsNaN(got) {
			t.Fatalf("MongeElkanScore(%q, %q, %s) = NaN", a, b, inner)
		}
		if math.IsInf(got, 0) {
			t.Fatalf("MongeElkanScore(%q, %q, %s) = Inf", a, b, inner)
		}
		if got < 0.0 || got > 1.0 {
			t.Fatalf("MongeElkanScore(%q, %q, %s) = %g; want in [0, 1]", a, b, inner, got)
		}
		// Identity short-circuit regression check.
		if idScore := fuzzymatch.MongeElkanScore(a, a, inner, opts); idScore != 1.0 {
			t.Fatalf("MongeElkanScore(%q, %q, %s) = %g; identity must be exactly 1.0", a, a, inner, idScore)
		}
	})
}

// FuzzMongeElkanScoreSymmetric exercises the SYMMETRIC variant with
// the same seeds and invariants as the asymmetric fuzz target, plus
// the symmetry property: score(a, b) == score(b, a) for all inputs.
func FuzzMongeElkanScoreSymmetric(f *testing.F) {
	for _, seed := range []struct {
		a, b string
		inn  int
	}{
		{"user create", "usr creating", 5},
		{"alpha beta", "alpha beta", 5},
		{"alpha beta", "gamma delta", 5},
		{"alpha", "alpha beta gamma", 0},
		{"café", "cafe", 0},
		{"alpha beta gamma", "alpha", 0},
		{"", "", 5},
		{"hello", "", 5},
		{"", "hello", 5},
		{"\xff\xfe", "abc", 5},
		{"café münchen", "münchen wien", 5},
		{"a b c", "b c d", 5},
		{"___", "...", 5},
	} {
		f.Add(seed.a, seed.b, seed.inn)
	}

	f.Fuzz(func(t *testing.T, a, b string, innRaw int) {
		inner := fuzzCoerceMongeElkanInner(innRaw)
		opts := fuzzymatch.DefaultNormalisationOptions()
		fwd := fuzzymatch.MongeElkanScoreSymmetric(a, b, inner, opts)
		if math.IsNaN(fwd) {
			t.Fatalf("MongeElkanScoreSymmetric(%q, %q, %s) = NaN", a, b, inner)
		}
		if math.IsInf(fwd, 0) {
			t.Fatalf("MongeElkanScoreSymmetric(%q, %q, %s) = Inf", a, b, inner)
		}
		if fwd < 0.0 || fwd > 1.0 {
			t.Fatalf("MongeElkanScoreSymmetric(%q, %q, %s) = %g; want in [0, 1]", a, b, inner, fwd)
		}
		// Symmetry property (load-bearing — this is what distinguishes
		// the symmetric variant from the asymmetric direct surface).
		rev := fuzzymatch.MongeElkanScoreSymmetric(b, a, inner, opts)
		if fwd != rev {
			t.Fatalf("MongeElkanScoreSymmetric(%q, %q, %s) = %g but (%q, %q) = %g — symmetric variant must be order-independent", a, b, inner, fwd, b, a, rev)
		}
	})
}
