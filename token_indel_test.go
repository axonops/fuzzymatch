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
//
// token_indel_test.go pins the unexported lcsLen / indelRatio /
// lcsLenRunes / indelRatioRunes helpers from token_indel.go
// (re-exported via LCSLenForTest / IndelRatioForTest /
// LCSLenRunesForTest / IndelRatioRunesForTest in export_test.go).
//
// The kernel is shared infrastructure for plans 06-01 (TokenSortRatio),
// 06-02 (TokenSetRatio), and 06-03 (PartialRatio). Its contract is the
// Wagner-Fischer 1974 LCS-SUBSEQUENCE recurrence (NOT the substring
// recurrence — see TestLCSLen_DistinctFromLCSStr for the PITFALL 6
// regression gate) and the Indel-formula normalisation
// 2·LCS / (|a|+|b|).
//
// Stdlib `testing` only — no testify in root tests, per
// .claude/skills/go-coding-standards.

package fuzzymatch_test

import (
	"math"
	"testing"

	"github.com/axonops/fuzzymatch"
)

// tokenIndelEpsilon is the float-comparison tolerance for irrational
// expected values. Phase 2/3/4/5 convention is 1e-9; the Indel formula
// reduces to a single integer-valued division so the actual accuracy
// is far higher than 1e-9, but the convention is locked. For
// exact-rational expected values (0.0, 0.5, 0.75, 1.0) tests use
// direct equality.
const tokenIndelEpsilon = 1e-9

// TestLCSLen exercises the LCS-SUBSEQUENCE byte-path kernel against a
// table of canonical fixtures including the Wagner-Fischer §3-style
// worked example (subsequence "AGCT" in "AGCTAGCT" → 4) and the
// PITFALL 6 fixture ("abc"/"axc" → 2) that pins the divergence from
// the LCS-SUBSTRING kernel.
func TestLCSLen(t *testing.T) {
	tests := []struct {
		name string
		a, b string
		want int
	}{
		{"both_empty", "", "", 0},
		{"one_empty_a", "", "abc", 0},
		{"one_empty_b", "abc", "", 0},
		{"identity", "abc", "abc", 3},
		{"identity_long", "AGCTAGCT", "AGCTAGCT", 8},
		// PITFALL 6 — subsequence "ac" exists in both (gap at index 1).
		// The substring kernel would return 1.
		{"pitfall6_abc_axc", "abc", "axc", 2},
		// Wagner-Fischer §3-style: subsequence "AGCT" of length 4 is
		// present in both (the second occurrence in "AGCTAGCT" or any
		// of the contiguous AGCT positions).
		{"wagner_fischer_AGCT_AGCTAGCT", "AGCT", "AGCTAGCT", 4},
		// Asymmetric-length: the inner-loop-over-shorter-side swap
		// must produce the same result regardless of argument order.
		{"asymmetric_short_first", "ace", "abcde", 3}, // subsequence "ace"
		{"asymmetric_long_first", "abcde", "ace", 3},  // mirror
		{"disjoint_letters", "abc", "xyz", 0},         // no common chars
		{"single_char_match", "a", "abc", 1},          // one shared
		{"single_char_no_match", "a", "bcd", 0},       // disjoint singleton
		{"interleaved", "abcabc", "acbacb", 4},        // subsequence "abab" or "acac" or "abac" length 4
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := fuzzymatch.LCSLenForTest([]byte(tt.a), []byte(tt.b))
			if got != tt.want {
				t.Errorf("lcsLen(%q, %q) = %d; want %d", tt.a, tt.b, got, tt.want)
			}
		})
	}
}

// TestLCSLen_DistinctFromLCSStr is the LOAD-BEARING PITFALL 6
// regression test. It asserts in the SAME test function that
//
//	lcsLen([]byte("abc"), []byte("axc")) == 2          (subsequence "ac")
//	len(LongestCommonSubstring("abc", "axc")) == 1     (substring "a" or "c")
//
// so the LCS-SUBSEQUENCE vs LCS-SUBSTRING divergence cannot regress
// silently. Any future refactoring that routes an LCS-subsequence call
// through `lcsstr.go`'s substring kernel (or vice versa) is caught here
// before it can corrupt downstream Indel scores.
//
// The fixture is the smallest possible — three characters with a single
// internal mismatch — chosen so a code reviewer can verify the answer
// by inspection in seconds.
func TestLCSLen_DistinctFromLCSStr(t *testing.T) {
	const a, b = "abc", "axc"

	gotLCS := fuzzymatch.LCSLenForTest([]byte(a), []byte(b))
	if gotLCS != 2 {
		t.Errorf("lcsLen(%q, %q) = %d; want 2 (subsequence \"ac\")", a, b, gotLCS)
	}

	gotSubstr := fuzzymatch.LongestCommonSubstring(a, b)
	if len(gotSubstr) != 1 {
		t.Errorf("len(LongestCommonSubstring(%q, %q)) = %d (substring %q); want 1 (substring kernel) — PITFALL 6 regression: substring length unexpectedly matches subsequence length",
			a, b, len(gotSubstr), gotSubstr)
	}

	// The divergence itself: subsequence MUST be strictly greater than
	// substring on this fixture. If they happened to coincide on this
	// input, the test logic above would still catch the absolute values
	// drifting from (2, 1), but the explicit divergence assertion makes
	// the keystone gate visible in code review.
	if gotLCS <= len(gotSubstr) {
		t.Errorf("PITFALL 6 regression: lcsLen (%d) <= len(LongestCommonSubstring) (%d) on (%q, %q) — the SUBSEQUENCE kernel should report a longer match than the SUBSTRING kernel on inputs with internal mismatches",
			gotLCS, len(gotSubstr), a, b)
	}
}

// TestIndelRatio exercises the Indel-formula byte-path kernel against
// hand-derived expected values. Each non-trivial value's derivation is
// reproduced in the test name so reviewers can re-derive the score
// from the formula 2·LCS / (|a|+|b|) in seconds.
func TestIndelRatio(t *testing.T) {
	tests := []struct {
		name       string
		a, b       string
		want       float64
		exact      bool // true → expect bit-equality; false → within tokenIndelEpsilon
		derivation string
	}{
		{
			name: "both_empty_identity", a: "", b: "",
			want: 1.0, exact: true,
			derivation: "vacuous match — both empty",
		},
		{
			name: "one_empty_a", a: "", b: "abc",
			want: 0.0, exact: true,
			derivation: "one-empty convention → 0.0",
		},
		{
			name: "one_empty_b", a: "abc", b: "",
			want: 0.0, exact: true,
			derivation: "one-empty convention → 0.0",
		},
		{
			name: "identity_abc", a: "abc", b: "abc",
			want: 1.0, exact: true,
			derivation: "lcs=3, sum=6, 2·3/6 = 1.0",
		},
		{
			name: "pitfall6_abc_axc", a: "abc", b: "axc",
			want: 2.0 * 2.0 / 6.0, exact: false,
			derivation: "lcs=2 (subsequence \"ac\"), sum=6, 2·2/6 ≈ 0.6667",
		},
		{
			name: "disjoint_abc_xyz", a: "abc", b: "xyz",
			want: 0.0, exact: true,
			derivation: "lcs=0, sum=6, 2·0/6 = 0.0",
		},
		{
			name: "subset_ace_abcde", a: "ace", b: "abcde",
			want: 2.0 * 3.0 / 8.0, exact: true,
			derivation: "lcs=3 (\"ace\"), sum=8, 2·3/8 = 0.75",
		},
		{
			name: "asymmetric_swap_consistent", a: "abcde", b: "ace",
			want: 2.0 * 3.0 / 8.0, exact: true,
			derivation: "symmetric mirror of subset_ace_abcde",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Logf("derivation: %s", tt.derivation)
			got := fuzzymatch.IndelRatioForTest([]byte(tt.a), []byte(tt.b))
			if tt.exact {
				if got != tt.want {
					t.Errorf("indelRatio(%q, %q) = %.17g; want %.17g exactly",
						tt.a, tt.b, got, tt.want)
				}
			} else {
				if math.Abs(got-tt.want) > tokenIndelEpsilon {
					t.Errorf("indelRatio(%q, %q) = %.17g; want %.17g (Δ=%g, ε=%g)",
						tt.a, tt.b, got, tt.want, math.Abs(got-tt.want), tokenIndelEpsilon)
				}
			}
		})
	}
}

// TestLCSLenRunes exercises the rune-path LCS-subsequence kernel
// against the canonical multi-byte UTF-8 fixture ("café" / "cafe" → 3
// runes shared: "caf"). The byte path would not match the multi-byte
// "é" atomically, so the rune kernel's correctness on multi-byte input
// is the load-bearing property here.
func TestLCSLenRunes(t *testing.T) {
	tests := []struct {
		name string
		a, b string
		want int
	}{
		{"both_empty", "", "", 0},
		{"one_empty", "", "café", 0},
		{"identity_cafe", "café", "café", 4},
		{"cafe_cafe_three_match", "café", "cafe", 3}, // "caf" subsequence
		{"identity_short_unicode", "λα", "λα", 2},    // Greek lambda + alpha
		{"disjoint_unicode", "αβ", "γδ", 0},
		{"ascii_inputs_match_byte_path", "abc", "axc", 2},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := fuzzymatch.LCSLenRunesForTest([]rune(tt.a), []rune(tt.b))
			if got != tt.want {
				t.Errorf("lcsLenRunes(%q, %q) = %d; want %d", tt.a, tt.b, got, tt.want)
			}
		})
	}
}

// TestIndelRatioRunes exercises the rune-path Indel-formula kernel.
// The "café"/"cafe" fixture (3 shared runes out of 4+4) produces the
// canonical 2·3/8 = 0.75 rune-path reference vector.
func TestIndelRatioRunes(t *testing.T) {
	tests := []struct {
		name       string
		a, b       string
		want       float64
		exact      bool
		derivation string
	}{
		{
			name: "both_empty", a: "", b: "",
			want: 1.0, exact: true,
			derivation: "vacuous match",
		},
		{
			name: "one_empty", a: "café", b: "",
			want: 0.0, exact: true,
			derivation: "one-empty convention",
		},
		{
			name: "identity_cafe", a: "café", b: "café",
			want: 1.0, exact: true,
			derivation: "lcs=4 runes, sum=8 runes, 2·4/8 = 1.0",
		},
		{
			name: "cafe_three_share", a: "café", b: "cafe",
			want: 0.75, exact: true,
			derivation: "lcs=3 (\"caf\"), sum=8 runes (4+4), 2·3/8 = 0.75",
		},
		{
			name: "disjoint_unicode", a: "αβ", b: "γδ",
			want: 0.0, exact: true,
			derivation: "lcs=0, 2·0/4 = 0.0",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Logf("derivation: %s", tt.derivation)
			got := fuzzymatch.IndelRatioRunesForTest([]rune(tt.a), []rune(tt.b))
			if tt.exact {
				if got != tt.want {
					t.Errorf("indelRatioRunes(%q, %q) = %.17g; want %.17g exactly",
						tt.a, tt.b, got, tt.want)
				}
			} else {
				if math.Abs(got-tt.want) > tokenIndelEpsilon {
					t.Errorf("indelRatioRunes(%q, %q) = %.17g; want %.17g (Δ=%g, ε=%g)",
						tt.a, tt.b, got, tt.want, math.Abs(got-tt.want), tokenIndelEpsilon)
				}
			}
		})
	}
}

// TestLCSLen_LongInputs_HeapPath exercises the heap-row allocation
// branch (min(|a|, |b|) > maxStackInputLen = 64). Identity on a
// 100-byte input gives lcs = 100 deterministically; this is the
// smallest meaningful check that the heap-path code is reached and
// produces the correct answer.
func TestLCSLen_LongInputs_HeapPath(t *testing.T) {
	a := make([]byte, 100)
	for i := range a {
		a[i] = byte('a' + i%26)
	}
	got := fuzzymatch.LCSLenForTest(a, a)
	if got != 100 {
		t.Errorf("lcsLen on identical 100-byte input = %d; want 100 (heap path identity)", got)
	}
}
