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

// token_jaccard_fuzz_test.go runs native Go fuzzing against the
// TokenJaccardScore public surface. Properties checked per input:
//
//   1. Never panics (implicit — any panic propagates as a fuzz crash).
//   2. Score never returns NaN.
//   3. Score never returns ±Inf.
//   4. Score returns a value in [0.0, 1.0].
//   5. Identity short-circuit holds — TokenJaccardScore(x, x) == 1.0 for
//      all x. The single-string fuzz target is achieved by re-running
//      the dual-string body with a == b after the main body, providing
//      load-bearing coverage for the IN-04 closure pattern.
//
// On-disk seed corpus: programmatic f.Add seeds in this file cover the
// six hand-derived reference vectors (RV-TJ1..RV-TJ6), the set-vs-
// multiset keystone (RV-TJ3), identity, both-empty, one-empty,
// orthogonal sets, invalid UTF-8, and multi-byte UTF-8 cases.
//
// CI's nightly fuzz job runs each fuzzer for 60s+; locally run
// `go test -fuzz=FuzzTokenJaccardScore -fuzztime=10s ./...` for a
// smoke check.

package fuzzymatch_test

import (
	"math"
	"strings"
	"testing"

	"github.com/axonops/fuzzymatch"
)

// FuzzTokenJaccardScore exercises the public surface with programmatic
// seeds covering the six reference vectors and a broad set of edge
// cases. The fuzz body asserts the four invariants (no panic, no NaN,
// no ±Inf, score in [0, 1]) plus the identity short-circuit.
func FuzzTokenJaccardScore(f *testing.F) {
	// Programmatic seeds covering the load-bearing fixtures.
	for _, seed := range []struct {
		a, b string
	}{
		{"a b c", "b c d"}, // RV-TJ1 partial overlap
		{"a b", "a b c"},   // RV-TJ2 subset
		{"a a b", "a b"},   // RV-TJ3 set-vs-multiset keystone
		{"a b c", "x y z"}, // RV-TJ4 disjoint
		{"a b c", "a b c"}, // RV-TJ5 identity
		{"alpha beta gamma delta", "alpha beta epsilon zeta"}, // RV-TJ6 partial
		{"", ""},                         // both-empty STANDARD
		{"hello", ""},                    // one-empty
		{"", "hello"},                    // one-empty
		{"hello world", "world hello"},   // token-reorder
		{"\xff\xfe", "abc"},              // invalid UTF-8
		{"\xc0\x80", "abc"},              // overlong NUL
		{"café münchen", "münchen wien"}, // multi-byte UTF-8
		{strings.Repeat("a ", 100), strings.Repeat("a ", 50)}, // dedup-heavy
		{"___", "..."},        // pure separators (both → empty token sets → 1.0)
		{"userID", "user_id"}, // tokeniser-divergence (identifier-style)
	} {
		f.Add(seed.a, seed.b)
	}

	f.Fuzz(func(t *testing.T, a, b string) {
		got := fuzzymatch.TokenJaccardScore(a, b)
		if math.IsNaN(got) {
			t.Fatalf("TokenJaccardScore(%q, %q) = NaN", a, b)
		}
		if math.IsInf(got, 0) {
			t.Fatalf("TokenJaccardScore(%q, %q) = Inf", a, b)
		}
		if got < 0.0 || got > 1.0 {
			t.Fatalf("TokenJaccardScore(%q, %q) = %g; want in [0, 1]", a, b, got)
		}
		// Identity short-circuit regression check: TokenJaccardScore(a, a)
		// must return exactly 1.0 (the a == b short-circuit fires
		// before Tokenise). This is load-bearing for the IN-04 closure
		// pattern: any future refactor that drops the short-circuit
		// would surface here because Tokenise of pure-separator
		// strings would produce an empty token set on both sides and
		// still return 1.0 — but for strings with mixed content the
		// short-circuit is the only guarantee of exact 1.0.
		if idScore := fuzzymatch.TokenJaccardScore(a, a); idScore != 1.0 {
			t.Fatalf("TokenJaccardScore(%q, %q) = %g; identity must be exactly 1.0", a, a, idScore)
		}
	})
}
