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

// jarowinkler_fuzz_test.go provides coverage-guided fuzz testing for
// JaroWinklerScore. The harness asserts: no panic, no NaN, no Inf, and
// score in [0.0, 1.0] for arbitrary byte inputs (including invalid UTF-8).
//
// Corpus seeds are stored in testdata/fuzz/FuzzJaroWinklerScore/.
//
// Run with:
//
//	go test -fuzz=FuzzJaroWinklerScore -fuzztime=30s -run=^$ ./...

package fuzzymatch_test

import (
	"math"
	"testing"

	"github.com/axonops/fuzzymatch"
)

// FuzzJaroWinklerScore is the fuzz harness for JaroWinklerScore. It seeds the
// corpus with the canonical Winkler 1990 reference pairs, edge cases, and an
// invalid-UTF-8 input to exercise the byte-level path robustness.
//
// Invariants checked per run:
//   - No panic
//   - !math.IsNaN(score)
//   - !math.IsInf(score, 0)
//   - score >= 0.0 && score <= 1.0
func FuzzJaroWinklerScore(f *testing.F) {
	// Programmatic seeds: Winkler 1990 canonical pairs.
	f.Add("MARTHA", "MARHTA")
	f.Add("DIXON", "DICKSONX")
	f.Add("DWAYNE", "DUANE")
	// Edge cases.
	f.Add("", "ABC")
	f.Add("", "")
	f.Add("ABC", "ABC")
	// Below-threshold pair.
	f.Add("abc", "xyz")
	// Invalid UTF-8 bytes — the byte-level path must not panic.
	f.Add("\xff\xfe", "\xfe\xff")

	f.Fuzz(func(t *testing.T, a, b string) {
		score := fuzzymatch.JaroWinklerScore(a, b)
		if math.IsNaN(score) {
			t.Errorf("JaroWinklerScore(%q, %q) = NaN", a, b)
		}
		if math.IsInf(score, 0) {
			t.Errorf("JaroWinklerScore(%q, %q) = Inf", a, b)
		}
		if score < 0.0 || score > 1.0 {
			t.Errorf("JaroWinklerScore(%q, %q) = %g; want in [0.0, 1.0]", a, b, score)
		}
	})
}
