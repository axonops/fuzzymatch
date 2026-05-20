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

// fuzz_test.go runs native Go fuzzing against scan.Check. The
// invariant is panic-freedom on consumer-supplied input (per 09-RESEARCH.md
// §Security Domain): scan.Check NEVER panics on caller-supplied data.
// Errors are acceptable (ErrInvalidItem on empty Name, ErrInvalidConfig
// on bad SuppressedPairs, etc.). The completeness-assertion panic is
// unreachable for fuzz inputs because D-06 rejects duplicate
// (Name, Group) at validateCheck and the synthetic Items below carry
// distinct Group values by construction.
//
// Seed corpus covers boundary inputs explicitly: empty string,
// self-name pairing, snake_case vs camelCase, Unicode (Greek + Cyrillic
// + Latin supplement), null bytes, invalid UTF-8 sequences.
//
// CI's nightly fuzz job runs the fuzzer for an extended window; locally
// run:
//
//	go test -fuzz=FuzzCheck -fuzztime=30s -run=^$ ./scan/...

package scan_test

import (
	"testing"

	"github.com/axonops/fuzzymatch"
	"github.com/axonops/fuzzymatch/scan"
)

// FuzzCheck asserts panic-freedom for any consumer input. The single
// hard invariant is "no panic". Errors are acceptable; the recovered
// completeness-assertion panic is unreachable under D-06 validation
// for the two-Item shape used here.
func FuzzCheck(f *testing.F) {
	// Seed corpus — boundary inputs covering the documented edges.
	for _, pair := range []struct{ a, b string }{
		{"", "x"},             // ErrInvalidItem boundary (empty Name)
		{"a", "a"},            // self-name pair (distinct groups, distinct (Name, Group))
		{"user_id", "userId"}, // canonical identifier-style pair
		{"κόσμε", "kosme"},    // Greek vs ASCII transliteration
		{"Привет", "privet"},  // Cyrillic vs ASCII
		{"café", "cafe"},      // Latin supplement (é = U+00E9)
		{"a\x00b", "a\x00b"},  // embedded null byte
		{"\xff\xfe", "\xff"},  // invalid UTF-8 (high bytes without continuation)
		{"\xc0\x80", "abc"},   // invalid UTF-8 (overlong NUL encoding)
		{"verylongidentifiername123", "veryLongIdentifierName123"},
	} {
		f.Add(pair.a, pair.b)
	}

	f.Fuzz(func(t *testing.T, a, b string) {
		items := []scan.Item{
			{Name: a, Group: "g1"},
			{Name: b, Group: "g2"},
		}
		cfg := scan.DefaultConfig(fuzzymatch.DefaultScorer())
		cfg.CompareAcrossGroups = true

		// Invariant: Check must NEVER panic on consumer-supplied
		// input. Errors are acceptable (e.g. ErrInvalidItem on empty
		// Name); the completeness-assertion panic is unreachable for
		// these inputs because:
		//   - The two items live in distinct Groups, so D-06
		//     duplicate-(Name, Group) detection is satisfied.
		//   - Even when a == b, the within-group pass produces no
		//     pairs (each group has one item) and the cross-group
		//     pass produces at most one warning per direction.
		//
		// We discard the result and only assert no panic. Any panic
		// here surfaces as a fuzz crash report with the offending
		// (a, b) corpus entry.
		_, _ = scan.Check(items, cfg)
	})
}
