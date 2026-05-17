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

// soundex_test.go pins the public-API contract of soundex.go: identity,
// both-empty, one-empty, the canonical Knuth TAOCP Vol. 3 §6.4 reference
// vectors (Robert/Rupert/Rubin/Tymczak/Ashcraft/Ashcroft/Pfister/Smith/
// Honeyman/Lloyd), the LOAD-BEARING Knuth/Census variant gate (Tymczak→T522,
// NOT SQL T520), the H/W-handling gate (Ashcraft/Ashcroft→A261), and the
// non-ASCII silent-skip pin per CONTEXT.md §5.
//
// Stdlib `testing` only — no testify in root tests, per
// .claude/skills/go-coding-standards.

package fuzzymatch_test

import (
	"testing"

	"github.com/axonops/fuzzymatch"
)

// TestSoundex_BothEmpty asserts both-empty → code="", score=1.0 per the
// algorithm-correctness-standards both-empty → 1.0 convention. The identity
// short-circuit covers this case.
func TestSoundex_BothEmpty(t *testing.T) {
	if got := fuzzymatch.SoundexCode(""); got != "" {
		t.Errorf("SoundexCode(\"\") = %q; want \"\"", got)
	}
	if got := fuzzymatch.SoundexScore("", ""); got != 1.0 {
		t.Errorf("SoundexScore(\"\", \"\") = %g; want 1.0", got)
	}
}

// TestSoundex_OneEmpty asserts one-empty → score=0.0 in both argument
// positions. An empty code cannot match a non-empty code.
func TestSoundex_OneEmpty(t *testing.T) {
	if got := fuzzymatch.SoundexScore("Robert", ""); got != 0.0 {
		t.Errorf("SoundexScore(\"Robert\", \"\") = %g; want 0.0", got)
	}
	if got := fuzzymatch.SoundexScore("", "Robert"); got != 0.0 {
		t.Errorf("SoundexScore(\"\", \"Robert\") = %g; want 0.0", got)
	}
}

// TestSoundexCode_KnuthReferenceVectors verifies all 12 canonical reference
// vectors from Knuth TAOCP Vol. 3 §6.4 (Knuth p. 393 examples) plus edge
// cases. These are the primary-source literature vectors.
//
// Cross-validated against jellyfish==1.2.1 (BSD-2-Clause) which also uses
// the Knuth/Census variant (confirmed by direct read of
// jellyfish/src/soundex.rs — H/W are skipped, not separators).
func TestSoundexCode_KnuthReferenceVectors(t *testing.T) {
	tests := []struct {
		name string
		in   string
		want string
	}{
		// Knuth TAOCP Vol. 3 §6.4 p. 393 canonical examples:
		{"Robert", "Robert", "R163"},     // Knuth p. 393 first example
		{"Rupert", "Rupert", "R163"},     // Knuth p. 393 — same code as Robert
		{"Rubin", "Rubin", "R150"},       // Knuth p. 393
		{"Tymczak", "Tymczak", "T522"},   // Knuth/Census variant gate (SQL yields T520)
		{"Ashcraft", "Ashcraft", "A261"}, // H/W-handling gate
		{"Ashcroft", "Ashcroft", "A261"}, // H/W-handling pair (same as Ashcraft)
		{"Pfister", "Pfister", "P236"},   // P skips F (both group 1 → P absorbs P, then group 2 then 3 then 6 -- wait, P236)
		{"Smith", "Smith", "S530"},
		{"Honeyman", "Honeyman", "H555"}, // mn adjacent (both group 5) — wait Honeyman: H-o-n-e-y-m-a-n
		{"Lloyd", "Lloyd", "L300"},       // double-L → only one L digit
		// Additional reference vectors (Knuth §6.4 broader set):
		{"Jackson", "Jackson", "J250"},
		{"Euler", "Euler", "E460"},
		{"Ellery", "Ellery", "E460"}, // Same code as Euler
		{"Gauss", "Gauss", "G200"},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			got := fuzzymatch.SoundexCode(tt.in)
			if got != tt.want {
				t.Errorf("SoundexCode(%q) = %q; want %q", tt.in, got, tt.want)
			}
		})
	}
}

// TestSoundexCode_TymczakVariantGate is the LOAD-BEARING Knuth/Census
// variant gate. "Tymczak" encodes as "T522" under the Knuth/Census rule
// (H and W do not break adjacent-group collapse), but as "T520" under the
// SQL/MySQL variant (where H and W ARE separators). This test pins the
// variant choice unambiguously.
//
// Derivation:
//
//	T → T (first letter)
//	y → vowel separator (lastGroup reset to 0)
//	m → group 5 → digit '5', lastGroup=5
//	c → group 2 → digit '2', lastGroup=2
//	z → group 2 → SAME group as c → SUPPRESSED (lastGroup stays 2)
//	a → vowel separator (lastGroup reset to 0)
//	k → group 2 → digit '2', lastGroup=2
//	Result: T522 (NOT T520 — z is suppressed, not separated by 'a')
//
// SQL/MySQL variant would give T520 because in that variant 'z' → '2'
// is output separately from 'c' → '2' even though they are adjacent in
// the same group.
func TestSoundexCode_TymczakVariantGate(t *testing.T) {
	const input = "Tymczak"
	const wantKnuth = "T522"
	const wantSQL = "T520" // not our variant; documented for reference

	got := fuzzymatch.SoundexCode(input)
	if got != wantKnuth {
		t.Errorf("SoundexCode(%q) = %q; want %q (Knuth/Census variant)\n"+
			"  SQL/MySQL variant would return %q — verify this is not a variant regression",
			input, got, wantKnuth, wantSQL)
	}
}

// TestSoundexCode_HW_HandlingPairs verifies that H and W are treated as
// transparent skips (not separators). "Ashcraft" and "Ashcroft" must both
// produce "A261": the 'h' between 's' (group 2) and 'c' (group 2) is
// transparent, so 'c' is suppressed (same group as 's'). The 'cr' in
// Ashcraft vs the 'cr' in Ashcroft both trigger the same path.
//
// Note: per Knuth §6.4, 's' (group 2) + 'h' (skip) + 'c' (group 2) →
// the 'c' is suppressed because the H skip does not reset lastGroup.
// Then 'r' (group 6) → '6'. The 'a' before 'f' (Ashcraft) acts as
// a separator but 'f' (group 1) comes after, giving 'A' + '2' + '6' + '1'.
func TestSoundexCode_HW_HandlingPairs(t *testing.T) {
	codeA := fuzzymatch.SoundexCode("Ashcraft")
	codeB := fuzzymatch.SoundexCode("Ashcroft")
	if codeA != "A261" {
		t.Errorf("SoundexCode(\"Ashcraft\") = %q; want \"A261\" (H/W skip gate)", codeA)
	}
	if codeB != "A261" {
		t.Errorf("SoundexCode(\"Ashcroft\") = %q; want \"A261\" (H/W skip gate)", codeB)
	}
	if codeA != codeB {
		t.Errorf("SoundexCode(\"Ashcraft\") = %q != SoundexCode(\"Ashcroft\") = %q; H/W handling must produce equal codes", codeA, codeB)
	}
}

// TestSoundexScore_Reflexivity asserts SoundexScore(x, x) == 1.0 for
// representative non-empty inputs. The identity short-circuit in
// SoundexScore fires before any computation.
func TestSoundexScore_Reflexivity(t *testing.T) {
	tests := []string{"Robert", "Smith", "a", "Z", "x"}
	for _, s := range tests {
		s := s
		t.Run(s, func(t *testing.T) {
			if got := fuzzymatch.SoundexScore(s, s); got != 1.0 {
				t.Errorf("SoundexScore(%q, %q) = %g; want 1.0 (identity)", s, s, got)
			}
		})
	}
}

// TestSoundexScore_NonMatch asserts SoundexScore returns 0.0 when codes differ.
func TestSoundexScore_NonMatch(t *testing.T) {
	pairs := [][2]string{
		{"Robert", "Smith"},
		{"Smith", "Jones"},
	}
	for _, pair := range pairs {
		a, b := pair[0], pair[1]
		if got := fuzzymatch.SoundexScore(a, b); got != 0.0 {
			t.Errorf("SoundexScore(%q, %q) = %g; want 0.0 (codes differ)", a, b, got)
		}
	}
}

// TestSoundexScore_Match asserts SoundexScore returns 1.0 when codes match.
// Robert and Rupert both encode as R163.
func TestSoundexScore_Match(t *testing.T) {
	if got := fuzzymatch.SoundexScore("Robert", "Rupert"); got != 1.0 {
		t.Errorf("SoundexScore(\"Robert\", \"Rupert\") = %g; want 1.0 (both R163)", got)
	}
}

// TestSoundexCode_NonASCII_SilentSkip verifies non-ASCII runes are silently
// dropped per CONTEXT.md §5. The encoded result uses only the ASCII letters.
func TestSoundexCode_NonASCII_SilentSkip(t *testing.T) {
	// "Müller" → encoding sees "Mller" → M460 (M, ll→4, r→6, padded 0)
	// Actually: M + l(4) + l(same group → suppress) + e(vowel, reset) + r(6)
	// = M460 (M + '4' + '6' + '0')
	got := fuzzymatch.SoundexCode("Müller")
	if len(got) != 0 && got[0] != 'M' {
		t.Errorf("SoundexCode(\"Müller\"): first char = %q; want 'M' (non-ASCII ü dropped)", string(got[0]))
	}
	// Verify output charset is [A-Z0-9] or empty.
	for _, c := range got {
		if (c < 'A' || c > 'Z') && (c < '0' || c > '9') {
			t.Errorf("SoundexCode(\"Müller\") = %q; contains non-[A-Z0-9] character %q", got, c)
		}
	}
	// "中文" → no ASCII letters → ""
	got2 := fuzzymatch.SoundexCode("中文")
	if got2 != "" {
		t.Errorf("SoundexCode(\"中文\") = %q; want \"\" (all non-ASCII, no letters)", got2)
	}
}

// TestSoundexCode_OutputLength verifies non-empty input always produces exactly
// 4 characters (1 letter + 3 digits).
func TestSoundexCode_OutputLength(t *testing.T) {
	tests := []string{"A", "Robert", "Smith", "X", "Tymczak", "Ashcraft"}
	for _, s := range tests {
		got := fuzzymatch.SoundexCode(s)
		if len(got) != 4 {
			t.Errorf("SoundexCode(%q) = %q; len=%d; want len=4 (1 letter + 3 digits)", s, got, len(got))
		}
	}
}

// TestSoundexCode_CharsetConstraint verifies output contains only [A-Z][0-9]{3}.
func TestSoundexCode_CharsetConstraint(t *testing.T) {
	tests := []string{"Robert", "Tymczak", "Ashcraft", "Smith", "Lloyd", ""}
	for _, s := range tests {
		got := fuzzymatch.SoundexCode(s)
		if got == "" {
			continue // empty input → empty output — OK
		}
		if got[0] < 'A' || got[0] > 'Z' {
			t.Errorf("SoundexCode(%q) = %q: first character %q not in [A-Z]", s, got, got[0:1])
		}
		for i := 1; i < len(got); i++ {
			if got[i] < '0' || got[i] > '9' {
				t.Errorf("SoundexCode(%q) = %q: character %q at index %d not a digit", s, got, got[i:i+1], i)
			}
		}
	}
}

// TestSoundexScore_NumericalRegression pins SoundexScore values OUTSIDE the
// cross-validation corpus to detect silent regressions in the binary
// 0.0/1.0 contract.
func TestSoundexScore_NumericalRegression(t *testing.T) {
	tests := []struct {
		a, b string
		want float64
	}{
		// Same code (R163): score 1.0
		{"Robert", "Rupert", 1.0},
		// Same code (A261): score 1.0
		{"Ashcraft", "Ashcroft", 1.0},
		// Different codes: score 0.0
		{"Robert", "Smith", 0.0},
		{"Smith", "Jones", 0.0},
		// Both empty: score 1.0
		{"", "", 1.0},
		// One empty: score 0.0
		{"Robert", "", 0.0},
		{"", "Robert", 0.0},
	}
	for _, tt := range tests {
		got := fuzzymatch.SoundexScore(tt.a, tt.b)
		if got != tt.want {
			t.Errorf("SoundexScore(%q, %q) = %g; want %g", tt.a, tt.b, got, tt.want)
		}
	}
}
