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

// FuzzCheckConfig fuzzes the Config-shape surface: the
// CrossGroupThresholdBoost float and the CompareAcrossGroups /
// CompareIdenticalAcrossGroups bools. Closes the test-analyst gap
// from the Phase 9 final panel: the original FuzzCheck only fuzzes
// (a, b string) with a fixed Config; this harness exercises the
// validation pipeline P2 (config-field strict-range) and the
// boost-arithmetic-clamp boundary under random caller input.
//
// Invariant: panic-freedom. ErrInvalidConfig on NaN/±Inf/out-of-range
// boost is the expected error; other errors propagate.
func FuzzCheckConfig(f *testing.F) {
	// Seed corpus — documented boundary values for the boost float
	// plus the four bool-combination corners.
	seeds := []struct {
		boost                                 float64
		compareAcross, compareIdenticalAcross bool
	}{
		{0.0, false, false}, // zero-value baseline
		{0.05, true, false}, // DefaultConfig posture
		{1.0, true, true},   // upper-boundary
		{0.5, true, false},  // mid-range
	}
	for _, s := range seeds {
		f.Add(s.boost, s.compareAcross, s.compareIdenticalAcross)
	}

	f.Fuzz(func(t *testing.T, boost float64, compareAcross, compareIdenticalAcross bool) {
		items := []scan.Item{
			{Name: "user_id", Group: "g1"},
			{Name: "userId", Group: "g2"},
		}
		cfg := scan.Config{
			Scorer:                       fuzzymatch.DefaultScorer(),
			CompareAcrossGroups:          compareAcross,
			CrossGroupThresholdBoost:     boost,
			CompareIdenticalAcrossGroups: compareIdenticalAcross,
		}
		// Discard result; assert no panic.
		_, _ = scan.Check(items, cfg)
	})
}

// FuzzCheckSuppressedPairs fuzzes the SuppressedPairs [][2]string
// shape under a fixed corpus. Closes the test-analyst gap on
// suppression-validation panic-freedom: empty-string entries trigger
// D-05 ErrInvalidConfig, canonicalised pair lookups exercise the
// Pitfall 4 normalisation path, and adversarial UTF-8 / null bytes
// flow through canonicalPair.
//
// Invariant: panic-freedom. ErrInvalidConfig on any empty-string
// pair entry is expected; other errors propagate.
func FuzzCheckSuppressedPairs(f *testing.F) {
	// Seed corpus — boundary pair entries (empty, self-pair, mixed
	// case, Unicode, invalid UTF-8).
	for _, seed := range []struct{ a, b string }{
		{"user_id", "userId"},  // canonical match-able pair
		{"USER_ID", "user_id"}, // case-insensitive via normalisation
		{"user_id", "user_id"}, // self-pair (D-05 silently kept)
		{"", "userId"},         // empty side → ErrInvalidConfig
		{"\xff", "userId"},     // invalid UTF-8 in pair
		{"a\x00b", "user_id"},  // embedded null
	} {
		f.Add(seed.a, seed.b)
	}

	f.Fuzz(func(t *testing.T, a, b string) {
		items := []scan.Item{
			{Name: "user_id", Group: "g1"},
			{Name: "userId", Group: "g2"},
		}
		cfg := scan.DefaultConfig(fuzzymatch.DefaultScorer())
		cfg.CompareAcrossGroups = true
		cfg.SuppressedPairs = [][2]string{{a, b}}
		// Discard result; assert no panic.
		_, _ = scan.Check(items, cfg)
	})
}

// FuzzCheckItems fuzzes the []Item shape (Name + Group + SilenceLint)
// under a fixed Config. Closes the test-analyst gap on
// items-validation panic-freedom: empty Name (D-03), duplicate (Name,
// Group) (D-06), and SilenceLint OR-semantics all exercise the
// pre-emission predicate paths.
//
// Invariant: panic-freedom. ErrInvalidItem on empty/duplicate is
// expected; other errors propagate.
func FuzzCheckItems(f *testing.F) {
	for _, seed := range []struct {
		nameA, groupA string
		silenceA      bool
		nameB, groupB string
		silenceB      bool
	}{
		{"user_id", "g1", false, "userId", "g2", false},  // canonical
		{"user_id", "g1", true, "userId", "g2", false},   // SilenceLint A
		{"user_id", "g1", false, "userId", "g1", false},  // same group
		{"user_id", "g1", false, "user_id", "g1", false}, // D-06 duplicate
		{"", "g1", false, "userId", "g2", false},         // D-03 empty
		{"\xff", "g1", false, "\xff", "g2", false},       // invalid UTF-8
	} {
		f.Add(seed.nameA, seed.groupA, seed.silenceA, seed.nameB, seed.groupB, seed.silenceB)
	}

	f.Fuzz(func(t *testing.T, nameA, groupA string, silenceA bool, nameB, groupB string, silenceB bool) {
		items := []scan.Item{
			{Name: nameA, Group: groupA, SilenceLint: silenceA},
			{Name: nameB, Group: groupB, SilenceLint: silenceB},
		}
		cfg := scan.DefaultConfig(fuzzymatch.DefaultScorer())
		cfg.CompareAcrossGroups = true
		// Discard result; assert no panic.
		_, _ = scan.Check(items, cfg)
	})
}
