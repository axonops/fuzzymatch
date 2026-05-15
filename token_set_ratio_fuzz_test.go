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

// token_set_ratio_fuzz_test.go runs native Go fuzzing against
// TokenSetRatioScore. Properties checked per input:
//
//   1. Never panics (implicit — any panic propagates as a fuzz crash).
//   2. Score never returns NaN.
//   3. Score never returns ±Inf.
//   4. Score returns a value in [0.0, 1.0].
//   5. Symmetric: TokenSetRatioScore(a, b) == TokenSetRatioScore(b, a).
//
// Programmatic f.Add seeds cover: canonical token-reorder pair,
// identity, both-empty (where the deviation gate fires), one-empty,
// orthogonal, invalid UTF-8, multi-byte UTF-8, long inputs, and
// pathological pure-separator inputs.
//
// CI's nightly fuzz job runs each fuzzer for 60s+; locally run
// `go test -fuzz=FuzzTokenSetRatioScore -fuzztime=10s ./...` for a
// smoke check.

package fuzzymatch_test

import (
	"math"
	"strings"
	"testing"

	"github.com/axonops/fuzzymatch"
)

// FuzzTokenSetRatioScore exercises the TokenSetRatioScore public
// surface. Symmetric-pair property is the load-bearing assertion:
// without it a future refactor could break the symmetric invariant
// (e.g. by swapping operand order in only one of the three indelRatio
// calls).
func FuzzTokenSetRatioScore(f *testing.F) {
	for _, seed := range []struct {
		a, b string
	}{
		{"alpha beta", "beta alpha gamma"}, // three-way max non-trivial
		{"hello world", "world peace"},     // r3 (combined-vs-combined) wins
		{"hello", "hello"},                 // identity
		{"", ""},                           // both-empty
		{"alpha beta", ""},                 // one-empty
		{"", "alpha beta"},                 // one-empty mirror
		{"abc def", "xyz qrs"},             // disjoint
		{"\xff\xfe", "alpha"},              // invalid UTF-8
		{"\xc0\x80", "alpha"},              // overlong NUL
		{"café société", "société café"},   // multi-byte reorder (subset)
		{strings.Repeat("a ", 100), strings.Repeat("b ", 100)}, // long input
		{" _.- ", "_-.: "}, // pure separators on both sides (DEVIATION gate)
	} {
		f.Add(seed.a, seed.b)
	}

	f.Fuzz(func(t *testing.T, a, b string) {
		got := fuzzymatch.TokenSetRatioScore(a, b)
		if math.IsNaN(got) {
			t.Fatalf("TokenSetRatioScore(%q, %q) = NaN", a, b)
		}
		if math.IsInf(got, 0) {
			t.Fatalf("TokenSetRatioScore(%q, %q) = Inf", a, b)
		}
		if got < 0.0 || got > 1.0 {
			t.Fatalf("TokenSetRatioScore(%q, %q) = %g; want in [0, 1]", a, b, got)
		}
		// Symmetric-pair pin: TokenSetRatioScore(a, b) must equal
		// TokenSetRatioScore(b, a) bit-for-bit. Tokenise is
		// deterministic; set construction is order-independent;
		// the three-way max operator is order-insensitive; r1 and
		// r2 swap when (a, b) → (b, a) but the max(r1, r2, r3) is
		// invariant; r3 is symmetric in its argument order via
		// indelRatio's own symmetry.
		if rev := fuzzymatch.TokenSetRatioScore(b, a); rev != got {
			t.Fatalf("TokenSetRatioScore not symmetric: T(%q,%q)=%g, T(%q,%q)=%g",
				a, b, got, b, a, rev)
		}
	})
}
