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

// q_gram_test.go pins the unexported extractQGrams + extractQGramsRunes
// helpers from q_gram.go (re-exported via ExtractQGramsForTest /
// ExtractQGramsRunesForTest in export_test.go).
//
// The extractor is shared infrastructure for plans 05-01..05-04; its
// contract is multiset semantics with overlapping windows, non-nil
// empty maps for degenerate inputs, and the Ukkonen 1992 §3 worked
// example as the canonical reference vector.
//
// Stdlib `testing` only — no testify in root tests, per
// .claude/skills/go-coding-standards.

package fuzzymatch_test

import (
	"reflect"
	"testing"

	"github.com/axonops/fuzzymatch"
)

// TestExtractQGrams_Empty pins the degenerate-input contract: empty
// string, n < 1, and n > len all return a non-nil empty map. The
// non-nil guarantee lets callers (the four q-gram algorithms) range
// over the result without a separate nil check.
func TestExtractQGrams_Empty(t *testing.T) {
	tests := []struct {
		name string
		s    string
		n    int
	}{
		{"empty_string_n2", "", 2},
		{"empty_string_n1", "", 1},
		{"n_zero", "hello", 0},
		{"n_negative", "hello", -1},
		{"n_very_negative", "hello", -100},
		{"n_greater_than_len", "ab", 5},
		{"n_equals_len_plus_one", "abc", 4},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := fuzzymatch.ExtractQGramsForTest(tt.s, tt.n)
			if got == nil {
				t.Fatalf("extractQGrams(%q, %d) = nil; want non-nil empty map", tt.s, tt.n)
			}
			if len(got) != 0 {
				t.Errorf("extractQGrams(%q, %d) = %v; want empty map", tt.s, tt.n, got)
			}
		})
	}
}

// TestExtractQGrams_AGCTAGCT pins the Ukkonen 1992 §3 worked example
// — the load-bearing primary-source reference vector for the q-gram
// tier. Multiset cardinality 7 (sum of values) across 4 distinct keys.
func TestExtractQGrams_AGCTAGCT(t *testing.T) {
	got := fuzzymatch.ExtractQGramsForTest("AGCTAGCT", 2)
	want := map[string]int{
		"AG": 2,
		"GC": 2,
		"CT": 2,
		"TA": 1,
	}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("extractQGrams(\"AGCTAGCT\", 2) = %v; want %v", got, want)
	}
	// Multiset cardinality (sum of values) must be len(s)-n+1 = 7.
	var totalCard int
	for _, c := range got {
		totalCard += c
	}
	if totalCard != 7 {
		t.Errorf("multiset cardinality = %d; want 7 (len(s)-n+1)", totalCard)
	}
}

// TestExtractQGrams_AGCT pins the smaller half of the Ukkonen 1992 §3
// pair — the |QA| = 3 side.
func TestExtractQGrams_AGCT(t *testing.T) {
	got := fuzzymatch.ExtractQGramsForTest("AGCT", 2)
	want := map[string]int{
		"AG": 1,
		"GC": 1,
		"CT": 1,
	}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("extractQGrams(\"AGCT\", 2) = %v; want %v", got, want)
	}
}

// TestExtractQGrams_Multiset pins the overlapping-window accumulation
// rule: repeated q-grams increment the count rather than collapsing
// to a set.
func TestExtractQGrams_Multiset(t *testing.T) {
	tests := []struct {
		name string
		s    string
		n    int
		want map[string]int
	}{
		{
			name: "AAAA_n2_three_AA",
			s:    "AAAA",
			n:    2,
			want: map[string]int{"AA": 3},
		},
		{
			name: "AAAAA_n3_three_AAA",
			s:    "AAAAA",
			n:    3,
			want: map[string]int{"AAA": 3},
		},
		{
			name: "ababab_n2_alternating",
			s:    "ababab",
			n:    2,
			want: map[string]int{"ab": 3, "ba": 2},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := fuzzymatch.ExtractQGramsForTest(tt.s, tt.n)
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("extractQGrams(%q, %d) = %v; want %v", tt.s, tt.n, got, tt.want)
			}
		})
	}
}

// TestExtractQGrams_NEqualsOne pins the n=1 path (unigrams). Each byte
// becomes a single-byte key.
func TestExtractQGrams_NEqualsOne(t *testing.T) {
	got := fuzzymatch.ExtractQGramsForTest("hello", 1)
	want := map[string]int{
		"h": 1,
		"e": 1,
		"l": 2,
		"o": 1,
	}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("extractQGrams(\"hello\", 1) = %v; want %v", got, want)
	}
}

// TestExtractQGrams_NEqualsLen pins the n == len(s) corner case: a
// single q-gram covering the entire input.
func TestExtractQGrams_NEqualsLen(t *testing.T) {
	got := fuzzymatch.ExtractQGramsForTest("abc", 3)
	want := map[string]int{"abc": 1}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("extractQGrams(\"abc\", 3) = %v; want %v", got, want)
	}
}

// TestExtractQGramsRunes_MultiByte pins the rune-path semantics:
// "café" has 4 runes, so rune-bigrams produce 3 entries with the
// multi-byte UTF-8 encoding of "fé" as a string key.
func TestExtractQGramsRunes_MultiByte(t *testing.T) {
	got := fuzzymatch.ExtractQGramsRunesForTest("café", 2)
	want := map[string]int{
		"ca": 1,
		"af": 1,
		"fé": 1,
	}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("extractQGramsRunes(\"café\", 2) = %v; want %v", got, want)
	}
	// Verify the multi-byte key is present as the proper UTF-8 string
	// (the byte-path counterpart would NOT have "fé" as a key — it
	// would have "fÃ" or similar from byte-window splitting).
	if _, ok := got["fé"]; !ok {
		t.Errorf("rune-bigram \"fé\" missing from extractQGramsRunes output: %v", got)
	}
}

// TestExtractQGramsRunes_DivergesFromBytes pins the byte vs rune
// divergence on multi-byte UTF-8 input: the byte path treats "é"
// (U+00E9) as a 2-byte sequence (0xC3 0xA9) and at n=1 produces two
// keys; the rune path treats it as a single rune.
func TestExtractQGramsRunes_DivergesFromBytes(t *testing.T) {
	const eAcute = "é" // 2 bytes UTF-8: 0xC3 0xA9
	bytePath := fuzzymatch.ExtractQGramsForTest(eAcute, 1)
	runePath := fuzzymatch.ExtractQGramsRunesForTest(eAcute, 1)
	// Byte path: 2 single-byte keys (the two UTF-8 bytes of "é").
	if len(bytePath) != 2 {
		t.Errorf("byte path: extractQGrams(%q, 1) has len %d; want 2 (one per UTF-8 byte)", eAcute, len(bytePath))
	}
	// Rune path: 1 single-rune key.
	if len(runePath) != 1 {
		t.Errorf("rune path: extractQGramsRunes(%q, 1) has len %d; want 1 (single rune)", eAcute, len(runePath))
	}
	if _, ok := runePath["é"]; !ok {
		t.Errorf("rune path: extractQGramsRunes(%q, 1) missing \"é\" key: %v", eAcute, runePath)
	}
}

// TestExtractQGramsRunes_Empty pins the degenerate-input contract for
// the rune path. Same shape as the byte-path equivalent.
func TestExtractQGramsRunes_Empty(t *testing.T) {
	tests := []struct {
		name string
		s    string
		n    int
	}{
		{"empty_string_n2", "", 2},
		{"n_zero", "hello", 0},
		{"n_negative", "hello", -1},
		{"n_greater_than_runes", "café", 5}, // 4 runes < 5
		{"n_greater_than_len_byte_path_passes", "ab", 5},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := fuzzymatch.ExtractQGramsRunesForTest(tt.s, tt.n)
			if got == nil {
				t.Fatalf("extractQGramsRunes(%q, %d) = nil; want non-nil empty map", tt.s, tt.n)
			}
			if len(got) != 0 {
				t.Errorf("extractQGramsRunes(%q, %d) = %v; want empty map", tt.s, tt.n, got)
			}
		})
	}
}
