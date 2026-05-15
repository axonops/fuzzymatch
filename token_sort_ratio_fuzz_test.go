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

// token_sort_ratio_fuzz_test.go runs native Go fuzzing against
// TokenSortRatioScore. Properties checked per input:
//
//   1. Never panics (implicit — any panic propagates as a fuzz crash).
//   2. Score never returns NaN.
//   3. Score never returns ±Inf.
//   4. Score returns a value in [0.0, 1.0].
//   5. Identical inputs always score 1.0 (load-bearing for the identity
//      short-circuit).
//
// Programmatic f.Add seeds cover: canonical token-reorder pair,
// identity, both-empty, one-empty, orthogonal, invalid UTF-8, multi-byte
// UTF-8, long inputs, and pathological pure-separator inputs.
//
// CI's nightly fuzz job runs each fuzzer for 60s+; locally run
// `go test -fuzz=FuzzTokenSortRatioScore -fuzztime=10s ./...` for a
// smoke check.

package fuzzymatch_test

import (
	"math"
	"strings"
	"testing"

	"github.com/axonops/fuzzymatch"
)

// FuzzTokenSortRatioScore exercises the TokenSortRatioScore public
// surface. Identical-input regression check is the load-bearing
// property — without it a future refactor could route identical
// inputs through the post-Tokenise path and lose the short-circuit's
// 1.0 guarantee.
func FuzzTokenSortRatioScore(f *testing.F) {
	for _, seed := range []struct {
		a, b string
	}{
		{"fuzzy wuzzy was a bear", "wuzzy fuzzy was a bear"}, // canonical reorder → 1.0
		{"hello", "hello"},               // identity
		{"", ""},                         // both-empty
		{"alpha beta", ""},               // one-empty
		{"", "alpha beta"},               // one-empty mirror
		{"abc", "xyz"},                   // orthogonal
		{"\xff\xfe", "alpha"},            // invalid UTF-8
		{"\xc0\x80", "alpha"},            // overlong NUL
		{"café société", "société café"}, // multi-byte reorder
		{strings.Repeat("a", 200), strings.Repeat("ab", 100)}, // long input
		{" _.- ", "_-.: "}, // pure separators on both sides
	} {
		f.Add(seed.a, seed.b)
	}

	f.Fuzz(func(t *testing.T, a, b string) {
		got := fuzzymatch.TokenSortRatioScore(a, b)
		if math.IsNaN(got) {
			t.Fatalf("TokenSortRatioScore(%q, %q) = NaN", a, b)
		}
		if math.IsInf(got, 0) {
			t.Fatalf("TokenSortRatioScore(%q, %q) = Inf", a, b)
		}
		if got < 0.0 || got > 1.0 {
			t.Fatalf("TokenSortRatioScore(%q, %q) = %g; want in [0, 1]", a, b, got)
		}
		// Identical-input pin: the a == b short-circuit MUST return
		// exactly 1.0 for every identical input including all-separator
		// strings. This is the regression gate against a future
		// refactor that strips the short-circuit.
		if same := fuzzymatch.TokenSortRatioScore(a, a); same != 1.0 {
			t.Fatalf("TokenSortRatioScore(%q, %q) = %g; want 1.0 (identity short-circuit)", a, a, same)
		}
	})
}
