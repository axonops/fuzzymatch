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

// nysiis_test.go verifies the NYSIISCode and NYSIISScore functions against
// literature reference vectors from Knuth TAOCP Vol. 3 §6.4 and
// jellyfish==1.2.1 (BSD-2-Clause) supplementary vectors.
//
// LOAD-BEARING gate: TestNYSIISCode_TruncationGate asserts
//   len(NYSIISCode("Catherine")) == 6
// which is the discriminating invariant for original Taft-1970 vs.
// modified-NYSIIS (jellyfish emits "CATARAN", length 7).
//
// Reference: Taft, R. L. (1970). Name search techniques.
// New York State Identification and Intelligence System, Special Report No. 1.
// Canonical algorithm description: Knuth, D. E. (1973). TAOCP Vol. 3, §6.4.

package fuzzymatch_test

import (
	"math"
	"regexp"
	"testing"

	"github.com/axonops/fuzzymatch"
)

// TestNYSIIS_BothEmpty verifies both-empty returns 1.0 (identity convention
// per algorithm-correctness-standards: both inputs empty → return 1.0).
func TestNYSIIS_BothEmpty(t *testing.T) {
	got := fuzzymatch.NYSIISScore("", "")
	if got != 1.0 {
		t.Errorf("NYSIISScore(\"\", \"\") = %v; want 1.0 (both-empty convention)", got)
	}
	// NYSIISCode("") must be "".
	code := fuzzymatch.NYSIISCode("")
	if code != "" {
		t.Errorf("NYSIISCode(\"\") = %q; want \"\"", code)
	}
}

// TestNYSIIS_OneEmpty verifies one-empty returns 0.0.
func TestNYSIIS_OneEmpty(t *testing.T) {
	cases := []struct{ a, b string }{
		{"", "Brown"},
		{"Brown", ""},
	}
	for _, tc := range cases {
		got := fuzzymatch.NYSIISScore(tc.a, tc.b)
		if got != 0.0 {
			t.Errorf("NYSIISScore(%q, %q) = %v; want 0.0 (one-empty)", tc.a, tc.b, got)
		}
	}
}

// TestNYSIISCode_KnuthReferenceVectors exercises the 12 literature reference
// vectors from RESEARCH.md §3.3, all derived from Knuth TAOCP Vol. 3 §6.4
// and supplemented by jellyfish/testdata/nysiis.csv for confirmation.
//
// RV-N1..RV-N12 are listed in the table in RESEARCH.md §3.3. All codes that
// differ from jellyfish (>6 chars) are the TRUNCATED Taft-1970 values.
//
// Derivation of NYSIIS rules applied here (Wikipedia/Taft 9-step procedure):
//   - First 2+ chars: MAC→MCC, KN→NN, initial K→C, etc.
//   - Last chars: EE/IE→Y, DT/RT/RD/NT/ND→D
//   - Body: EV→AF; vowels→A; Q→G; Z→S; M→N; KN→N; K→C; SCH→S;
//     PH→FF; H adjacent to non-H/W removed; W after vowel removed
//   - Dedup consecutive same-letter sequences
//   - Truncate to 6 chars
func TestNYSIISCode_KnuthReferenceVectors(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  string
		// derivation documents how the algorithm produces the code
		derivation string
	}{
		{
			name:       "RV-N1: Brown",
			input:      "Brown",
			want:       "BRAN",
			derivation: "B→B; R→R; OW→... body rules: O→A(vowel), W after vowel removed; N→N → BRAN",
		},
		{
			name:       "RV-N2: Browne (same as Brown)",
			input:      "Browne",
			want:       "BRAN",
			derivation: "Browne→BRAN: same body as Brown; trailing E: last-2 = NE, not a special suffix",
		},
		{
			name:       "RV-N3: Robert",
			input:      "Robert",
			want:       "RABAD",
			derivation: "R→R; O→A(vowel); B→B; E→A(vowel); R→R; T→terminal suffix: RT→D → RABARD→dedup? →RABAD",
		},
		{
			name:       "RV-N4: Catherine (truncated from CATARAN)",
			input:      "Catherine",
			want:       "CATARA",
			derivation: "Truncated from 7-char CATARAN to 6 chars (Taft-1970 truncation gate)",
		},
		{
			name:       "RV-N5: Katherine (same as Catherine after truncation)",
			input:      "Katherine",
			want:       "CATARA",
			derivation: "KN- prefix→NN... but K alone at start→C; same as Catherine after initial processing",
		},
		{
			name:       "RV-N6: Johnathan (truncated from JANATAN)",
			input:      "Johnathan",
			want:       "JANATA",
			derivation: "Truncated from 7-char JANATAN to 6 chars (Taft-1970 truncation gate)",
		},
		{
			name:       "RV-N7: Jonathan (same as Johnathan after truncation)",
			input:      "Jonathan",
			want:       "JANATA",
			derivation: "Truncated from JANATAN (7) to JANATA (6)",
		},
		{
			name:       "RV-N8: John",
			input:      "John",
			want:       "JAN",
			derivation: "J→J; O→A(vowel); H→removed (adjacent H/W rules); N→N → JAN",
		},
		{
			name:       "RV-N9: Teresa",
			input:      "Teresa",
			want:       "TARAS",
			derivation: "T→T; E→A(vowel); R→R; E→A(vowel); S→S; A→... last char A, trailing vowel handling→Y? No; body: →TARAS",
		},
		{
			name:       "RV-N10: Theresa (same as Teresa)",
			input:      "Theresa",
			want:       "TARAS",
			derivation: "TH at start: PH/TH initial rule; same final code as Teresa",
		},
		{
			name:       "RV-N11: montgomery (truncated from MANTGANARY)",
			input:      "montgomery",
			want:       "MANTGA",
			derivation: "Truncated from 10-char MANTGANARY to 6 chars (Taft-1970 truncation gate)",
		},
		{
			name:       "RV-N12: empty",
			input:      "",
			want:       "",
			derivation: "Empty input → empty output",
		},
		// Regression: CR-01 (REVIEW.md) — RD→D and ND→D suffix rules
		// previously dropped the D and left R/N as the trailing char. Both
		// surnames below exercise the path:
		{
			name:       "RV-N13: Byrd (RD suffix → D)",
			input:      "Byrd",
			want:       "BYD",
			derivation: "B→B; Y→Y; trailing RD → D (Taft 1970 step 3): BYRD → BYD",
		},
		{
			name:       "RV-N14: Bond (ND suffix → D)",
			input:      "Bond",
			want:       "BAD",
			derivation: "B→B; O→A (vowel transliteration); trailing ND → D: BAND-truncated-by-step3 → BAD",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := fuzzymatch.NYSIISCode(tc.input)
			if got != tc.want {
				t.Errorf("NYSIISCode(%q) = %q; want %q\n  derivation: %s",
					tc.input, got, tc.want, tc.derivation)
			}
		})
	}
}

// TestNYSIISCode_TruncationGate is the LOAD-BEARING discriminator test for the
// original Taft-1970 variant vs. the modified/jellyfish NYSIIS variant.
//
// jellyfish.nysiis("Catherine") = "CATARAN" (7 chars — modified NYSIIS, no truncation).
// Our impl MUST return "CATARA" (6 chars — original Taft-1970, truncated).
//
// If this test fails with len == 7, the modified-NYSIIS variant was accidentally
// shipped. See CONTEXT.md §2 LOCKED and RESEARCH.md §7 Pitfall 7.B.
func TestNYSIISCode_TruncationGate(t *testing.T) {
	code := fuzzymatch.NYSIISCode("Catherine")
	if len(code) != 6 {
		t.Errorf(
			"NYSIISCode(\"Catherine\") = %q (length %d); want length 6 exactly.\n"+
				"  If length is 7 (%q), the MODIFIED-NYSIIS variant (no truncation) was shipped instead of\n"+
				"  the original Taft-1970 6-char-truncation variant locked by CONTEXT.md §2.\n"+
				"  jellyfish.nysiis(\"Catherine\") = \"CATARAN\" (7 chars); Taft-1970 truncated = \"CATARA\" (6 chars).",
			code, len(code), "CATARAN",
		)
	}
	if code != "CATARA" {
		t.Errorf("NYSIISCode(\"Catherine\") = %q; want \"CATARA\" (Taft-1970 6-char truncation)", code)
	}
}

// TestNYSIISScore_BrownBrowneCanonicalMatch verifies the canonical pair from
// Knuth TAOCP Vol. 3 §6.4: Brown and Browne both encode to "BRAN".
func TestNYSIISScore_BrownBrowneCanonicalMatch(t *testing.T) {
	// Both should encode to BRAN.
	codeB := fuzzymatch.NYSIISCode("Brown")
	codeBe := fuzzymatch.NYSIISCode("Browne")
	if codeB != "BRAN" {
		t.Errorf("NYSIISCode(\"Brown\") = %q; want \"BRAN\"", codeB)
	}
	if codeBe != "BRAN" {
		t.Errorf("NYSIISCode(\"Browne\") = %q; want \"BRAN\"", codeBe)
	}
	// Score should be 1.0.
	score := fuzzymatch.NYSIISScore("Brown", "Browne")
	if score != 1.0 {
		t.Errorf("NYSIISScore(\"Brown\", \"Browne\") = %v; want 1.0 (canonical BRAN pair)", score)
	}
}

// TestNYSIISScore_Identity verifies that identical non-empty inputs return 1.0.
func TestNYSIISScore_Identity(t *testing.T) {
	inputs := []string{"Brown", "Robert", "Catherine", "John", "A"}
	for _, s := range inputs {
		got := fuzzymatch.NYSIISScore(s, s)
		if got != 1.0 {
			t.Errorf("NYSIISScore(%q, %q) = %v; want 1.0 (identity)", s, s, got)
		}
	}
}

// TestNYSIISScore_NonMatchPair verifies that inputs with different codes return 0.0.
func TestNYSIISScore_NonMatchPair(t *testing.T) {
	// Brown (BRAN) vs Robert (RABAD) must return 0.0.
	got := fuzzymatch.NYSIISScore("Brown", "Robert")
	if got != 0.0 {
		t.Errorf("NYSIISScore(\"Brown\", \"Robert\") = %v; want 0.0 (BRAN != RABAD)", got)
	}
}

// TestNYSIISCode_OutputLength verifies that NYSIISCode never returns a code
// longer than 6 characters for any ASCII input.
func TestNYSIISCode_OutputLength(t *testing.T) {
	longInputs := []string{
		"Johnathan", "Jonathan", "Katherine", "Catherine",
		"montgomery", "martincevic", "Christopher", "Bartholomew",
		"Alexandersson", "Konstantinopoulou",
	}
	for _, s := range longInputs {
		code := fuzzymatch.NYSIISCode(s)
		if len(code) > 6 {
			t.Errorf("NYSIISCode(%q) = %q (length %d); want length <= 6 (Taft-1970 truncation)",
				s, code, len(code))
		}
	}
}

// TestNYSIISCode_OutputCharset verifies that NYSIISCode only produces uppercase
// ASCII letters (no digits, no punctuation, no lowercase) matching [A-Z]{0,6}.
func TestNYSIISCode_OutputCharset(t *testing.T) {
	re := regexp.MustCompile(`^[A-Z]{0,6}$`)
	inputs := []string{
		"Brown", "Browne", "Robert", "Catherine", "John", "Teresa",
		"montgomery", "Johnathan", "", "X", "AZ",
	}
	for _, s := range inputs {
		code := fuzzymatch.NYSIISCode(s)
		if !re.MatchString(code) {
			t.Errorf("NYSIISCode(%q) = %q; does not match [A-Z]{0,6}", s, code)
		}
	}
}

// TestNYSIISCode_NonASCIISilentSkip verifies that non-ASCII input is silently
// dropped, consistent with CONTEXT.md §5.
func TestNYSIISCode_NonASCIISilentSkip(t *testing.T) {
	// Non-ASCII runes are dropped; only ASCII [A-Za-z] participates.
	// The result must still match [A-Z]{0,6}.
	re := regexp.MustCompile(`^[A-Z]{0,6}$`)
	nonASCII := []string{"Müller", "Café", "中文", "🎉hello", "\xff\xfe"}
	for _, s := range nonASCII {
		code := fuzzymatch.NYSIISCode(s)
		if !re.MatchString(code) {
			t.Errorf("NYSIISCode(%q) = %q; non-ASCII input produced non-[A-Z]{0,6} output", s, code)
		}
	}
}

// TestNYSIISScore_RangeInvariant verifies Score always returns 0.0 or 1.0.
func TestNYSIISScore_RangeInvariant(t *testing.T) {
	pairs := []struct{ a, b string }{
		{"Brown", "Browne"},
		{"Brown", "Robert"},
		{"", ""},
		{"", "Brown"},
		{"Catherine", "Katherine"},
		{"John", "Joan"},
	}
	for _, p := range pairs {
		got := fuzzymatch.NYSIISScore(p.a, p.b)
		if got != 0.0 && got != 1.0 {
			t.Errorf("NYSIISScore(%q, %q) = %v; want 0.0 or 1.0 (binary score)", p.a, p.b, got)
		}
		if math.IsNaN(got) || math.IsInf(got, 0) {
			t.Errorf("NYSIISScore(%q, %q) = %v; want finite 0.0 or 1.0", p.a, p.b, got)
		}
	}
}

// TestNYSIISScore_CatherineKatherineMatch verifies that the Taft-1970 truncation
// still produces a phonetic match between Catherine and Katherine (both → CATARA).
func TestNYSIISScore_CatherineKatherineMatch(t *testing.T) {
	score := fuzzymatch.NYSIISScore("Catherine", "Katherine")
	if score != 1.0 {
		t.Errorf("NYSIISScore(\"Catherine\", \"Katherine\") = %v; want 1.0 (both encode to CATARA)", score)
	}
}
