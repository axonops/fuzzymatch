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

// validate_test.go covers the Validate public surface (validate.go) +
// WarnKind enum (warn_kind.go). Tests close VALIDATE-01..VALIDATE-05
// from .planning/REQUIREMENTS.md.
//
// Coverage breakdown:
//   - TestWarnKind_String — VALIDATE-03 (CamelCase String() per Q6b)
//   - TestWarnKind_Constants_Distinct — VALIDATE-04 (5 distinct values)
//   - TestWarnKinds_OrderAndContents — slice-helper determinism
//   - TestValidate_EmptyInput — WarnEmptyInput
//   - TestValidate_HammingUnequalLength — WarnUnequalLength
//   - TestValidate_TokenTierEmpty — WarnNoTokensAfterNormalise
//   - TestValidate_ASCIIOnlyAlgosWithUnicode — WarnAllNonASCIIDropped
//   - TestValidate_LargeInput — WarnPathologicallyLargeInput
//   - TestValidate_NoWarnings_ReturnsNil — nil-vs-empty-slice contract
//   - TestValidate_DeterministicOrdering — sort.SliceStable gate
//   - TestValidate_NeverPanics_PerAlgorithm — VALIDATE-01 panic-free contract
//   - TestValidate_PerAlgorithm_RuleTable — VALIDATE-05 explicit per-algo dispatch
//   - TestAlgoIDAny_String — sentinel renders cleanly

package fuzzymatch_test

import (
	"reflect"
	"sort"
	"strings"
	"testing"
	"unicode/utf8"

	"github.com/axonops/fuzzymatch"
)

// TestWarnKind_String pins the CamelCase contract per VALIDATE-03 +
// Q6b. The labels must match docs/requirements.md §11.5 exactly so
// log lines and golden files stay stable across patch releases.
func TestWarnKind_String(t *testing.T) {
	tests := []struct {
		kind fuzzymatch.WarnKind
		want string
	}{
		{fuzzymatch.WarnEmptyInput, "EmptyInput"},
		{fuzzymatch.WarnUnequalLength, "UnequalLength"},
		{fuzzymatch.WarnNoTokensAfterNormalise, "NoTokensAfterNormalise"},
		{fuzzymatch.WarnAllNonASCIIDropped, "AllNonASCIIDropped"},
		{fuzzymatch.WarnPathologicallyLargeInput, "PathologicallyLargeInput"},
	}
	for _, tt := range tests {
		t.Run(tt.want, func(t *testing.T) {
			got := tt.kind.String()
			if got != tt.want {
				t.Errorf("WarnKind(%d).String() = %q; want %q", int(tt.kind), got, tt.want)
			}
		})
	}
}

// TestWarnKind_String_OutOfRange exercises the fallback branch and
// pins the WarnKind(0) zero-value behaviour (which must produce a
// debug-only stringification, not one of the named labels).
func TestWarnKind_String_OutOfRange(t *testing.T) {
	tests := []struct {
		kind fuzzymatch.WarnKind
		want string
	}{
		{fuzzymatch.WarnKind(0), "WarnKind(0)"}, // zero value reserved as unset
		{fuzzymatch.WarnKind(99), "WarnKind(99)"},
		{fuzzymatch.WarnKind(-1), "WarnKind(-1)"},
	}
	for _, tt := range tests {
		t.Run(tt.want, func(t *testing.T) {
			got := tt.kind.String()
			if got != tt.want {
				t.Errorf("WarnKind(%d).String() = %q; want %q", int(tt.kind), got, tt.want)
			}
		})
	}
}

// TestWarnKind_Constants_Distinct asserts each named constant has a
// distinct int value (the iota+1 contract). Two constants sharing a
// value would break programmatic dispatch in consumers.
func TestWarnKind_Constants_Distinct(t *testing.T) {
	all := fuzzymatch.WarnKinds()
	seen := make(map[fuzzymatch.WarnKind]bool, len(all))
	for _, k := range all {
		if seen[k] {
			t.Errorf("WarnKind value %d (%s) duplicated", int(k), k)
		}
		seen[k] = true
	}
	// Pin the iota+1 contract — WarnEmptyInput must be 1, not 0.
	if fuzzymatch.WarnEmptyInput != 1 {
		t.Errorf("WarnEmptyInput = %d; want 1 (iota+1 contract)", int(fuzzymatch.WarnEmptyInput))
	}
}

// TestWarnKinds_OrderAndContents pins the deterministic enumeration
// order of WarnKinds(). Two successive calls must return the same
// slice contents in the same order.
func TestWarnKinds_OrderAndContents(t *testing.T) {
	first := fuzzymatch.WarnKinds()
	second := fuzzymatch.WarnKinds()
	if !reflect.DeepEqual(first, second) {
		t.Fatalf("WarnKinds() not stable: %v vs %v", first, second)
	}
	want := []fuzzymatch.WarnKind{
		fuzzymatch.WarnEmptyInput,
		fuzzymatch.WarnUnequalLength,
		fuzzymatch.WarnNoTokensAfterNormalise,
		fuzzymatch.WarnAllNonASCIIDropped,
		fuzzymatch.WarnPathologicallyLargeInput,
	}
	if !reflect.DeepEqual(first, want) {
		t.Errorf("WarnKinds() = %v; want %v", first, want)
	}
	if len(first) != 5 {
		t.Errorf("WarnKinds() len = %d; want 5", len(first))
	}
}

// TestAlgoIDAny_String asserts the cross-cutting sentinel renders as
// "Any". The label is part of the v1.x stability contract — consumers
// log Warning.Algorithm.String() expecting this exact value.
func TestAlgoIDAny_String(t *testing.T) {
	got := fuzzymatch.AlgoIDAny.String()
	if got != "Any" {
		t.Errorf("AlgoIDAny.String() = %q; want %q", got, "Any")
	}
}

// TestValidate_EmptyInput asserts WarnEmptyInput fires for every empty-
// input shape and is scoped to AlgoIDAny. Validates VALIDATE-01 +
// VALIDATE-02 + VALIDATE-05's WarnEmptyInput rule.
func TestValidate_EmptyInput(t *testing.T) {
	tests := []struct {
		name string
		a, b string
	}{
		{"a_empty", "", "abc"},
		{"b_empty", "abc", ""},
		{"both_empty", "", ""},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			warnings := fuzzymatch.Validate(tt.a, tt.b)
			if !containsKind(warnings, fuzzymatch.WarnEmptyInput) {
				t.Errorf("Validate(%q, %q) missing WarnEmptyInput; got %v", tt.a, tt.b, warnings)
			}
			// WarnEmptyInput must be scoped to AlgoIDAny per the rule
			// table.
			found := false
			for _, w := range warnings {
				if w.Kind == fuzzymatch.WarnEmptyInput {
					if w.Algorithm != fuzzymatch.AlgoIDAny {
						t.Errorf("WarnEmptyInput scoped to %s; want AlgoIDAny", w.Algorithm)
					}
					found = true
				}
			}
			if !found {
				t.Errorf("WarnEmptyInput not present in %v", warnings)
			}
		})
	}
}

// TestValidate_HammingUnequalLength asserts the Hamming-specific
// WarnUnequalLength rule. Both characters are ASCII so no other
// per-algorithm warning fires.
func TestValidate_HammingUnequalLength(t *testing.T) {
	warnings := fuzzymatch.Validate("a", "ab")
	found := false
	for _, w := range warnings {
		if w.Kind == fuzzymatch.WarnUnequalLength && w.Algorithm == fuzzymatch.AlgoHamming {
			found = true
		}
	}
	if !found {
		t.Errorf("Validate(\"a\", \"ab\") missing {AlgoHamming, WarnUnequalLength}; got %v", warnings)
	}
}

// TestValidate_HammingEqualLength asserts WarnUnequalLength is NOT
// emitted when inputs are the same length.
func TestValidate_HammingEqualLength(t *testing.T) {
	warnings := fuzzymatch.Validate("abc", "xyz")
	for _, w := range warnings {
		if w.Kind == fuzzymatch.WarnUnequalLength {
			t.Errorf("Validate(\"abc\", \"xyz\") unexpectedly emitted WarnUnequalLength: %v", w)
		}
	}
}

// TestValidate_TokenTierEmpty asserts WarnNoTokensAfterNormalise is
// emitted once per token-tier algorithm when an input tokenises to
// nothing (e.g. pure-separator input like "---" — every byte is a
// separator under DefaultTokeniseOptions).
func TestValidate_TokenTierEmpty(t *testing.T) {
	warnings := fuzzymatch.Validate("hello", "---")
	tokenAlgos := map[fuzzymatch.AlgoID]bool{
		fuzzymatch.AlgoMongeElkan:     false,
		fuzzymatch.AlgoTokenSortRatio: false,
		fuzzymatch.AlgoTokenSetRatio:  false,
		fuzzymatch.AlgoPartialRatio:   false,
		fuzzymatch.AlgoTokenJaccard:   false,
	}
	for _, w := range warnings {
		if w.Kind == fuzzymatch.WarnNoTokensAfterNormalise {
			if _, ok := tokenAlgos[w.Algorithm]; !ok {
				t.Errorf("WarnNoTokensAfterNormalise scoped to non-token-tier algo %s", w.Algorithm)
			}
			tokenAlgos[w.Algorithm] = true
		}
	}
	for algo, seen := range tokenAlgos {
		if !seen {
			t.Errorf("missing WarnNoTokensAfterNormalise for token-tier algo %s", algo)
		}
	}
}

// TestValidate_ASCIIOnlyAlgosWithUnicode asserts WarnAllNonASCIIDropped
// is emitted once per ASCII-only algorithm when inputs contain only
// non-ASCII runes.
func TestValidate_ASCIIOnlyAlgosWithUnicode(t *testing.T) {
	warnings := fuzzymatch.Validate("中文", "日本語")
	asciiAlgos := map[fuzzymatch.AlgoID]bool{
		fuzzymatch.AlgoStrcmp95:        false,
		fuzzymatch.AlgoSoundex:         false,
		fuzzymatch.AlgoDoubleMetaphone: false,
		fuzzymatch.AlgoNYSIIS:          false,
		fuzzymatch.AlgoMRA:             false,
	}
	for _, w := range warnings {
		if w.Kind == fuzzymatch.WarnAllNonASCIIDropped {
			if _, ok := asciiAlgos[w.Algorithm]; !ok {
				t.Errorf("WarnAllNonASCIIDropped scoped to non-ASCII-only algo %s", w.Algorithm)
			}
			asciiAlgos[w.Algorithm] = true
		}
	}
	for algo, seen := range asciiAlgos {
		if !seen {
			t.Errorf("missing WarnAllNonASCIIDropped for ASCII-only algo %s", algo)
		}
	}
}

// TestValidate_LargeInput asserts WarnPathologicallyLargeInput fires
// above the 64 KiB threshold and is scoped to AlgoIDAny. Uses a
// 70_000-byte input so the threshold is comfortably exceeded.
func TestValidate_LargeInput(t *testing.T) {
	big := strings.Repeat("a", 70_000)
	warnings := fuzzymatch.Validate(big, big)
	found := false
	for _, w := range warnings {
		if w.Kind == fuzzymatch.WarnPathologicallyLargeInput {
			if w.Algorithm != fuzzymatch.AlgoIDAny {
				t.Errorf("WarnPathologicallyLargeInput scoped to %s; want AlgoIDAny", w.Algorithm)
			}
			found = true
		}
	}
	if !found {
		t.Errorf("missing WarnPathologicallyLargeInput on 70_000-byte input; got %v", warnings)
	}
}

// TestValidate_LargeInput_BelowThreshold asserts the warning does NOT
// fire for inputs at or below the threshold.
func TestValidate_LargeInput_BelowThreshold(t *testing.T) {
	mid := strings.Repeat("a", 1024) // well below 64 KiB
	warnings := fuzzymatch.Validate(mid, mid)
	for _, w := range warnings {
		if w.Kind == fuzzymatch.WarnPathologicallyLargeInput {
			t.Errorf("Validate(short, short) unexpectedly emitted WarnPathologicallyLargeInput: %v", w)
		}
	}
}

// TestValidate_NoWarnings_ReturnsNil asserts VALIDATE-01's nil-vs-
// empty-slice contract. Two short well-formed ASCII inputs of equal
// length with valid tokens produce no warnings — the return value
// MUST be nil, not []Warning{}.
func TestValidate_NoWarnings_ReturnsNil(t *testing.T) {
	warnings := fuzzymatch.Validate("hello", "world")
	if warnings != nil {
		t.Errorf("Validate(\"hello\", \"world\") = %v; want nil", warnings)
	}
}

// TestValidate_DeterministicOrdering is the load-bearing T-08.5-26
// mitigation gate: two calls with identical inputs return slices that
// are DeepEqual. Any reordering (e.g. accidental map iteration on the
// output path) trips this test.
func TestValidate_DeterministicOrdering(t *testing.T) {
	inputs := []struct {
		a, b string
	}{
		{"", "abc"},
		{"中文", "日本語"},
		{"a", "ab"},
		{"hello", "---"},
		{strings.Repeat("a", 70_000), strings.Repeat("b", 70_000)},
	}
	for _, in := range inputs {
		t.Run(in.a+"|"+in.b, func(t *testing.T) {
			first := fuzzymatch.Validate(in.a, in.b)
			second := fuzzymatch.Validate(in.a, in.b)
			if !reflect.DeepEqual(first, second) {
				t.Errorf("Validate not deterministic: first=%v second=%v", first, second)
			}
			// Also assert the output is sorted by (Algorithm, Kind).
			if !sort.SliceIsSorted(first, func(i, j int) bool {
				if first[i].Algorithm != first[j].Algorithm {
					return first[i].Algorithm < first[j].Algorithm
				}
				return first[i].Kind < first[j].Kind
			}) {
				t.Errorf("Validate output not sorted by (Algorithm, Kind): %v", first)
			}
		})
	}
}

// TestValidate_NeverPanics_PerAlgorithm exercises a table of
// pathological inputs with a deferred recover. Validate is the
// consumer's safety net — it MUST NOT panic on any byte input.
func TestValidate_NeverPanics_PerAlgorithm(t *testing.T) {
	pathological := []struct {
		a, b string
	}{
		{"", ""},
		{"\x00", "\x00"},
		{"\xff\xfe", "\xff\xfe"}, // invalid UTF-8
		{"\xc0\x80", "abc"},      // overlong NUL
		{strings.Repeat("\xff", 200), strings.Repeat("\xff", 100)}, // long invalid UTF-8
		{"a\x00b", "a\x00b"}, // embedded NUL
		{"𝕳𝖊𝖑𝖑𝖔", "𝕳𝖊𝖑𝖑𝖔"},   // 4-byte UTF-8 only
		{"--", "--"}, // separator-only
	}
	for _, in := range pathological {
		t.Run(in.a+"|"+in.b, func(t *testing.T) {
			defer func() {
				if r := recover(); r != nil {
					t.Errorf("Validate(%q, %q) panicked: %v", in.a, in.b, r)
				}
			}()
			_ = fuzzymatch.Validate(in.a, in.b)
		})
	}
}

// TestValidate_PerAlgorithm_RuleTable is the VALIDATE-05 acceptance
// gate: every one of the 23 AlgoIDs has an explicit assertion of
// which WarnKind values its degraded input triggers. The test
// constructs targeted inputs and asserts that each (Algorithm,
// WarnKind) pair from the documented rule table actually appears.
//
// Cross-cutting Kinds (WarnEmptyInput, WarnPathologicallyLargeInput)
// are scoped to AlgoIDAny and validated separately in the dedicated
// TestValidate_EmptyInput and TestValidate_LargeInput tests; this
// table covers the per-algorithm rules.
func TestValidate_PerAlgorithm_RuleTable(t *testing.T) {
	// rule table — AlgoID → set of WarnKind values it can emit
	// (per-algorithm only; cross-cutting Kinds covered separately).
	expected := map[fuzzymatch.AlgoID][]fuzzymatch.WarnKind{
		fuzzymatch.AlgoHamming:         {fuzzymatch.WarnUnequalLength},
		fuzzymatch.AlgoStrcmp95:        {fuzzymatch.WarnAllNonASCIIDropped},
		fuzzymatch.AlgoSoundex:         {fuzzymatch.WarnAllNonASCIIDropped},
		fuzzymatch.AlgoDoubleMetaphone: {fuzzymatch.WarnAllNonASCIIDropped},
		fuzzymatch.AlgoNYSIIS:          {fuzzymatch.WarnAllNonASCIIDropped},
		fuzzymatch.AlgoMRA:             {fuzzymatch.WarnAllNonASCIIDropped},
		fuzzymatch.AlgoMongeElkan:      {fuzzymatch.WarnNoTokensAfterNormalise},
		fuzzymatch.AlgoTokenSortRatio:  {fuzzymatch.WarnNoTokensAfterNormalise},
		fuzzymatch.AlgoTokenSetRatio:   {fuzzymatch.WarnNoTokensAfterNormalise},
		fuzzymatch.AlgoPartialRatio:    {fuzzymatch.WarnNoTokensAfterNormalise},
		fuzzymatch.AlgoTokenJaccard:    {fuzzymatch.WarnNoTokensAfterNormalise},
	}
	// Algorithms with no per-algorithm warnings (only cross-cutting
	// Kinds via AlgoIDAny) — they must NOT appear as Warning.Algorithm
	// in any per-algorithm rule output.
	silent := []fuzzymatch.AlgoID{
		fuzzymatch.AlgoLevenshtein,
		fuzzymatch.AlgoDamerauLevenshteinOSA,
		fuzzymatch.AlgoDamerauLevenshteinFull,
		fuzzymatch.AlgoJaro,
		fuzzymatch.AlgoJaroWinkler,
		fuzzymatch.AlgoSmithWatermanGotoh,
		fuzzymatch.AlgoLCSStr,
		fuzzymatch.AlgoQGramJaccard,
		fuzzymatch.AlgoSorensenDice,
		fuzzymatch.AlgoCosine,
		fuzzymatch.AlgoTversky,
		fuzzymatch.AlgoRatcliffObershelp,
	}

	// Compose an input that triggers every per-algorithm rule at
	// once: non-ASCII (ASCII-only algos), unequal length (Hamming),
	// separators-only on side b (token-tier algos).
	a := "中文"
	b := "日本語" // different length from a (so Hamming trips too)

	// First sub-test: ASCII-only + Hamming rules together.
	t.Run("non_ascii_unequal_length", func(t *testing.T) {
		warnings := fuzzymatch.Validate(a, b)
		seen := make(map[fuzzymatch.AlgoID]map[fuzzymatch.WarnKind]bool)
		for _, w := range warnings {
			if seen[w.Algorithm] == nil {
				seen[w.Algorithm] = make(map[fuzzymatch.WarnKind]bool)
			}
			seen[w.Algorithm][w.Kind] = true
		}
		// Assert every documented (algo, kind) for ASCII-only + Hamming
		// surfaced.
		for algo, kinds := range expected {
			// Skip token-tier algos in this sub-test — their input
			// pair must tokenise to empty, which "中文"/"日本語" doesn't.
			if algo == fuzzymatch.AlgoMongeElkan ||
				algo == fuzzymatch.AlgoTokenSortRatio ||
				algo == fuzzymatch.AlgoTokenSetRatio ||
				algo == fuzzymatch.AlgoPartialRatio ||
				algo == fuzzymatch.AlgoTokenJaccard {
				continue
			}
			for _, kind := range kinds {
				if !seen[algo][kind] {
					t.Errorf("rule (%s, %s) not surfaced; got warnings %v", algo, kind, warnings)
				}
			}
		}
	})

	// Second sub-test: token-tier rules via separator-only input.
	t.Run("separator_only_input", func(t *testing.T) {
		warnings := fuzzymatch.Validate("hello", "---")
		seen := make(map[fuzzymatch.AlgoID]map[fuzzymatch.WarnKind]bool)
		for _, w := range warnings {
			if seen[w.Algorithm] == nil {
				seen[w.Algorithm] = make(map[fuzzymatch.WarnKind]bool)
			}
			seen[w.Algorithm][w.Kind] = true
		}
		for _, algo := range []fuzzymatch.AlgoID{
			fuzzymatch.AlgoMongeElkan,
			fuzzymatch.AlgoTokenSortRatio,
			fuzzymatch.AlgoTokenSetRatio,
			fuzzymatch.AlgoPartialRatio,
			fuzzymatch.AlgoTokenJaccard,
		} {
			if !seen[algo][fuzzymatch.WarnNoTokensAfterNormalise] {
				t.Errorf("rule (%s, WarnNoTokensAfterNormalise) not surfaced; got warnings %v", algo, warnings)
			}
		}
	})

	// Third sub-test: silent algorithms never appear in the per-
	// algorithm rule output (only via AlgoIDAny cross-cutting).
	t.Run("silent_algorithms_never_scoped_per_algo", func(t *testing.T) {
		// Compose a maximum-trigger input: empty + non-ASCII + large.
		// Each silent algo should appear at most via AlgoIDAny cross-
		// cutting Kinds, never via per-algorithm Kinds.
		warnings := fuzzymatch.Validate("中文", "")
		for _, w := range warnings {
			// AlgoIDAny is cross-cutting — exempt.
			if w.Algorithm == fuzzymatch.AlgoIDAny {
				continue
			}
			// Per-algorithm scope must not target a silent algo.
			for _, silent := range silent {
				if w.Algorithm == silent {
					t.Errorf("silent algorithm %s unexpectedly scoped per-algorithm: %v", silent, w)
				}
			}
		}
	})
}

// TestValidate_DetailStrings_AreValidUTF8 asserts every emitted
// Warning's Detail field is valid UTF-8. Detail strings are surfaced
// to logs and audit trails — invalid UTF-8 would corrupt downstream
// log pipelines.
func TestValidate_DetailStrings_AreValidUTF8(t *testing.T) {
	inputs := []struct {
		a, b string
	}{
		{"", "abc"},
		{"中文", "日本語"},
		{"a", "ab"},
		{"hello", "---"},
		{strings.Repeat("a", 70_000), strings.Repeat("b", 70_000)},
	}
	for _, in := range inputs {
		warnings := fuzzymatch.Validate(in.a, in.b)
		for _, w := range warnings {
			if !isValidUTF8(w.Detail) {
				t.Errorf("Warning.Detail is not valid UTF-8: %q (Algorithm=%s, Kind=%s)", w.Detail, w.Algorithm, w.Kind)
			}
		}
	}
}

// containsKind reports whether any Warning in s has the given Kind.
func containsKind(s []fuzzymatch.Warning, kind fuzzymatch.WarnKind) bool {
	for _, w := range s {
		if w.Kind == kind {
			return true
		}
	}
	return false
}

// isValidUTF8 reports whether s is well-formed UTF-8. Wraps
// utf8.ValidString from the stdlib.
func isValidUTF8(s string) bool {
	return utf8.ValidString(s)
}
