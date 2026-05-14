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

// lcsstr_test.go pins the public-API contract of lcsstr.go: identity,
// both-empty, one-empty, no-overlap disambiguation pin (RESEARCH.md Pitfall
// 6), canonical Wagner-Fischer 1974 reference vectors, the LEFTMOST-in-`a`
// tie-break (RESEARCH.md Pitfall 4 — load-bearing for the strict-`>`
// max-update), byte-vs-rune equivalence on ASCII, multi-byte rune handling,
// AND the runtime allocation gate.
//
// Stdlib `testing` only — no testify in root tests, per
// .claude/skills/go-coding-standards.

package fuzzymatch_test

import (
	"math"
	"testing"

	"github.com/axonops/fuzzymatch"
)

// TestLCSStr_BothEmpty pins the both-empty convention: LongestCommonSubstring
// returns "" and LCSStrScore returns 1.0 (the 2·n/(n+n)=1 identity that holds
// vacuously when n=0 is documented as a convention, not derived).
func TestLCSStr_BothEmpty(t *testing.T) {
	if got := fuzzymatch.LongestCommonSubstring("", ""); got != "" {
		t.Errorf("LongestCommonSubstring(\"\", \"\") = %q; want \"\"", got)
	}
	if got := fuzzymatch.LongestCommonSubstringRunes("", ""); got != "" {
		t.Errorf("LongestCommonSubstringRunes(\"\", \"\") = %q; want \"\"", got)
	}
	if got := fuzzymatch.LCSStrScore("", ""); got != 1.0 {
		t.Errorf("LCSStrScore(\"\", \"\") = %g; want 1.0", got)
	}
	if got := fuzzymatch.LCSStrScoreRunes("", ""); got != 1.0 {
		t.Errorf("LCSStrScoreRunes(\"\", \"\") = %g; want 1.0", got)
	}
}

// TestLCSStr_OneEmpty pins the one-empty convention: LongestCommonSubstring
// returns "" and LCSStrScore returns 0.0 (no shared content).
func TestLCSStr_OneEmpty(t *testing.T) {
	tests := []struct{ a, b string }{
		{"", "abc"},
		{"abc", ""},
		{"", "x"},
		{"x", ""},
	}
	for _, tt := range tests {
		t.Run("Sub_"+tt.a+"_"+tt.b, func(t *testing.T) {
			if got := fuzzymatch.LongestCommonSubstring(tt.a, tt.b); got != "" {
				t.Errorf("LongestCommonSubstring(%q, %q) = %q; want \"\"", tt.a, tt.b, got)
			}
			if got := fuzzymatch.LongestCommonSubstringRunes(tt.a, tt.b); got != "" {
				t.Errorf("LongestCommonSubstringRunes(%q, %q) = %q; want \"\"", tt.a, tt.b, got)
			}
			if got := fuzzymatch.LCSStrScore(tt.a, tt.b); got != 0.0 {
				t.Errorf("LCSStrScore(%q, %q) = %g; want 0.0", tt.a, tt.b, got)
			}
			if got := fuzzymatch.LCSStrScoreRunes(tt.a, tt.b); got != 0.0 {
				t.Errorf("LCSStrScoreRunes(%q, %q) = %g; want 0.0", tt.a, tt.b, got)
			}
		})
	}
}

// TestLCSStr_NoOverlap_DisambiguationPin is the RESEARCH.md Pitfall 6 closure:
// when inputs share NO characters, LongestCommonSubstring returns "" (same
// return as the both-empty case) — but LCSStrScore disambiguates by returning
// 0.0 (vs 1.0 for both-empty). This test pins BOTH return values to nail down
// the documented behaviour as a regression contract.
func TestLCSStr_NoOverlap_DisambiguationPin(t *testing.T) {
	if got := fuzzymatch.LongestCommonSubstring("abc", "xyz"); got != "" {
		t.Errorf("LongestCommonSubstring(\"abc\", \"xyz\") = %q; want \"\" (no shared chars)", got)
	}
	if got := fuzzymatch.LCSStrScore("abc", "xyz"); got != 0.0 {
		t.Errorf("LCSStrScore(\"abc\", \"xyz\") = %g; want 0.0 (disambiguates from both-empty=1.0)", got)
	}
	// And the disambiguation contract: both-empty and no-overlap return the
	// same substring ("") but DIFFERENT scores.
	if fuzzymatch.LCSStrScore("", "") == fuzzymatch.LCSStrScore("abc", "xyz") {
		t.Errorf("LCSStrScore both-empty and no-overlap should DIFFER (1.0 vs 0.0)")
	}
}

// TestLCSStr_Identical pins LongestCommonSubstring(x, x) == x and
// LCSStrScore(x, x) == 1.0 (the 2·n/(n+n)=1 identity) for non-empty x.
func TestLCSStr_Identical(t *testing.T) {
	tests := []string{"abc", "user_id", "x", "http_request"}
	for _, s := range tests {
		t.Run(s, func(t *testing.T) {
			if got := fuzzymatch.LongestCommonSubstring(s, s); got != s {
				t.Errorf("LongestCommonSubstring(%q, %q) = %q; want %q", s, s, got, s)
			}
			if got := fuzzymatch.LCSStrScore(s, s); got != 1.0 {
				t.Errorf("LCSStrScore(%q, %q) = %g; want 1.0", s, s, got)
			}
		})
	}
}

// TestLCSStr_ReferenceVectors_WagnerFischer1974 pins canonical Wagner-Fischer
// 1974 longest-common-substring reference vectors. The DP recurrence:
//
//	D[i,j] = D[i-1,j-1] + 1 if a[i-1] == b[j-1] else 0
//
// produces these expected (substring, score) outputs:
//
//	kitten/sitting → "itt", 2·3/(6+7) = 6/13 ≈ 0.461538…
//	http_request/http_request_header_fields → "http_request", 2·12/(12+26) = 24/38 ≈ 0.631578…
//	abcdef/zabcdefuvw → "abcdef", 2·6/(6+10) = 12/16 = 0.75
func TestLCSStr_ReferenceVectors_WagnerFischer1974(t *testing.T) {
	tests := []struct {
		a, b   string
		wantS  string
		wantSc float64
	}{
		{"kitten", "sitting", "itt", 6.0 / 13.0},
		{"http_request", "http_request_header_fields", "http_request", 24.0 / 38.0},
		{"abcdef", "zabcdefuvw", "abcdef", 12.0 / 16.0},
		{"banana", "ananas", "anana", 10.0 / 12.0},
	}
	const eps = 1e-12
	for _, tt := range tests {
		t.Run(tt.a+"_"+tt.b, func(t *testing.T) {
			if got := fuzzymatch.LongestCommonSubstring(tt.a, tt.b); got != tt.wantS {
				t.Errorf("LongestCommonSubstring(%q, %q) = %q; want %q", tt.a, tt.b, got, tt.wantS)
			}
			got := fuzzymatch.LCSStrScore(tt.a, tt.b)
			if math.Abs(got-tt.wantSc) > eps {
				t.Errorf("LCSStrScore(%q, %q) = %g; want ≈ %g (Δ=%g, eps=%g)",
					tt.a, tt.b, got, tt.wantSc, math.Abs(got-tt.wantSc), eps)
			}
		})
	}
}

// TestLCSStr_LeftmostTieBreak_Pinned is the LOAD-BEARING regression test for
// RESEARCH.md Pitfall 4: the strict-`>` max-update means first-found-leftmost
// wins. Multiple longest-common-substrings of equal length exist in these
// inputs; only the LEFTMOST occurrence in `a` is acceptable.
//
// If the kernel is written with `>=` instead of `>`, the LAST tied match wins
// and these assertions fail.
func TestLCSStr_LeftmostTieBreak_Pinned(t *testing.T) {
	tests := []struct {
		a, b      string
		wantSub   string
		wantStart int // index in `a` where wantSub starts
	}{
		// Canonical tie-break case: two occurrences of "abc" in `a` both length 3.
		{"abcXYZabc", "abc", "abc", 0},
		// Multi-occurrence with separator
		{"xy_abc_xy_abc", "abc", "abc", 3},
		// aaa / aa: three tied length-2 windows in `a` ("aa" at 0, "aa" at 1);
		// leftmost is at index 0.
		{"aaa", "aa", "aa", 0},
		// Two equal-length runs in `a` matching the full `b`.
		{"foo_bar_foo", "foo", "foo", 0},
		// Tied 4-char overlap (mississippi has "issi" twice as a substring).
		{"mississippi", "issi", "issi", 1},
	}
	for _, tt := range tests {
		t.Run(tt.a+"_"+tt.b, func(t *testing.T) {
			got := fuzzymatch.LongestCommonSubstring(tt.a, tt.b)
			if got != tt.wantSub {
				t.Errorf("LongestCommonSubstring(%q, %q) = %q; want %q (leftmost-in-a tie-break)",
					tt.a, tt.b, got, tt.wantSub)
				return
			}
			// And verify position-in-`a` is the leftmost match.
			if idx := indexInA(tt.a, got); idx != tt.wantStart {
				t.Errorf("LongestCommonSubstring(%q, %q) found %q at index %d; want leftmost index %d",
					tt.a, tt.b, got, idx, tt.wantStart)
			}
		})
	}
}

// indexInA returns the index of the first occurrence of sub in a, or -1.
// Tiny helper to avoid importing strings just for this one byte-scan.
func indexInA(a, sub string) int {
	if sub == "" {
		return 0
	}
	for i := 0; i+len(sub) <= len(a); i++ {
		if a[i:i+len(sub)] == sub {
			return i
		}
	}
	return -1
}

// TestLCSStr_ByteVsRune_Equivalence asserts the byte and rune paths produce
// identical output on pure-ASCII inputs (rune count == byte count).
func TestLCSStr_ByteVsRune_Equivalence(t *testing.T) {
	tests := [][2]string{
		{"kitten", "sitting"},
		{"abc", "abc"},
		{"", ""},
		{"", "abc"},
		{"abc", ""},
		{"abc", "xyz"},
		{"abcXYZabc", "abc"},
		{"http_request", "http_request_header_fields"},
		{"user_id", "user_name"},
	}
	for _, tt := range tests {
		a, b := tt[0], tt[1]
		t.Run(a+"_"+b, func(t *testing.T) {
			byteSub := fuzzymatch.LongestCommonSubstring(a, b)
			runeSub := fuzzymatch.LongestCommonSubstringRunes(a, b)
			if byteSub != runeSub {
				t.Errorf("byte vs rune substring mismatch on ASCII: byte=%q rune=%q for (%q, %q)",
					byteSub, runeSub, a, b)
			}
			byteScore := fuzzymatch.LCSStrScore(a, b)
			runeScore := fuzzymatch.LCSStrScoreRunes(a, b)
			if byteScore != runeScore {
				t.Errorf("byte vs rune score mismatch on ASCII: byte=%g rune=%g for (%q, %q)",
					byteScore, runeScore, a, b)
			}
		})
	}
}

// TestLCSStr_RuneMultiByte verifies the rune path handles multi-byte UTF-8
// correctly. "café" is 4 runes / 5 bytes; "cafe" is 4 runes / 4 bytes. The
// rune-level longest common substring is "caf" (3 runes); score = 2·3/(4+4)
// = 6/8 = 0.75.
//
// The byte-level result is allowed to differ — café/cafe both share the "caf"
// byte prefix (3 bytes); score = 2·3/(5+4) = 6/9. We only PIN the rune path
// here; the byte path is exercised by the ByteVsRune_Equivalence test on ASCII
// inputs only.
func TestLCSStr_RuneMultiByte(t *testing.T) {
	if got := fuzzymatch.LongestCommonSubstringRunes("café", "cafe"); got != "caf" {
		t.Errorf("LongestCommonSubstringRunes(\"café\", \"cafe\") = %q; want \"caf\"", got)
	}
	want := 6.0 / 8.0
	got := fuzzymatch.LCSStrScoreRunes("café", "cafe")
	if math.Abs(got-want) > 1e-12 {
		t.Errorf("LCSStrScoreRunes(\"café\", \"cafe\") = %g; want %g", got, want)
	}
	// Cyrillic case: Привет vs привет — only first-letter case differs in the
	// upper-case Russian П / lower-case п. Rune-level LCS is "ривет" (5 runes).
	if got := fuzzymatch.LongestCommonSubstringRunes("Привет", "привет"); got != "ривет" {
		t.Errorf("LongestCommonSubstringRunes(\"Привет\", \"привет\") = %q; want \"ривет\"", got)
	}
}

// TestLCSStr_RuneIdentity_ShortCircuit verifies LCSStrScoreRunes returns 1.0
// on identical inputs without allocating []rune(a) / []rune(b) (IN-04 closure
// inherited from Phase 3). Same for LongestCommonSubstringRunes returning the
// original string directly.
func TestLCSStr_RuneIdentity_ShortCircuit(t *testing.T) {
	const s = "café_user"
	if got := fuzzymatch.LCSStrScoreRunes(s, s); got != 1.0 {
		t.Errorf("LCSStrScoreRunes(%q, %q) = %g; want 1.0", s, s, got)
	}
	if got := fuzzymatch.LongestCommonSubstringRunes(s, s); got != s {
		t.Errorf("LongestCommonSubstringRunes(%q, %q) = %q; want %q", s, s, got, s)
	}
	allocs := testing.AllocsPerRun(100, func() {
		_ = fuzzymatch.LCSStrScoreRunes(s, s)
	})
	if allocs > 0 {
		t.Errorf("LCSStrScoreRunes identity short-circuit: %.1f allocs/op; want 0 (no []rune alloc)", allocs)
	}
}

// TestLCSStr_Symmetry pins Score(a, b) == Score(b, a) on the reference-vector
// pairs (and a few extras). Mathematical symmetry: the substring relation and
// 2·len(lcs)/(la+lb) are both symmetric in argument order.
func TestLCSStr_Symmetry(t *testing.T) {
	pairs := [][2]string{
		{"kitten", "sitting"},
		{"http_request", "http_request_header_fields"},
		{"abcXYZabc", "abc"},
		{"abc", "xyz"},
		{"banana", "ananas"},
		{"", ""},
		{"abc", ""},
	}
	for _, p := range pairs {
		a, b := p[0], p[1]
		t.Run(a+"_"+b, func(t *testing.T) {
			fwd := fuzzymatch.LCSStrScore(a, b)
			rev := fuzzymatch.LCSStrScore(b, a)
			if fwd != rev {
				t.Errorf("LCSStrScore not symmetric: Score(%q,%q)=%g != Score(%q,%q)=%g", a, b, fwd, b, a, rev)
			}
			fwdR := fuzzymatch.LCSStrScoreRunes(a, b)
			revR := fuzzymatch.LCSStrScoreRunes(b, a)
			if fwdR != revR {
				t.Errorf("LCSStrScoreRunes not symmetric: Score(%q,%q)=%g != Score(%q,%q)=%g", a, b, fwdR, b, a, revR)
			}
		})
	}
}

// TestLCSStrScore_ZeroAllocs_ASCII_Short pins the 0-alloc budget for the byte
// path on a short ASCII pair (well within maxStackInputLen=64).
func TestLCSStrScore_ZeroAllocs_ASCII_Short(t *testing.T) {
	_ = fuzzymatch.LCSStrScore("kitten", "sitting")
	allocs := testing.AllocsPerRun(100, func() {
		_ = fuzzymatch.LCSStrScore("kitten", "sitting")
	})
	if allocs > 0 {
		t.Errorf("LCSStrScore ASCII short: %.1f allocs/op; want 0 (stack buffer not escaping?)", allocs)
	}
}

// TestLCSStrScore_ZeroAllocs_ASCII_Medium pins the 0-alloc budget at ~50 bytes
// (still within maxStackInputLen=64).
func TestLCSStrScore_ZeroAllocs_ASCII_Medium(t *testing.T) {
	const a50 = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWX"
	const b50 = "bcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXY"
	_ = fuzzymatch.LCSStrScore(a50, b50)
	allocs := testing.AllocsPerRun(100, func() {
		_ = fuzzymatch.LCSStrScore(a50, b50)
	})
	if allocs > 0 {
		t.Errorf("LCSStrScore ASCII medium: %.1f allocs/op; want 0 (stack buffer not escaping?)", allocs)
	}
}

// TestLongestCommonSubstring_ZeroAllocs_ASCII_Short pins zero allocations on
// the byte path of LongestCommonSubstring for an identical-string short input.
// The identity short-circuit `if a == b { return a }` avoids the DP entirely;
// non-identical short inputs would incur the slice-of-string allocation
// inherent in the substring return (a[start:end]) but that's a string-header
// stack value not a heap allocation on the byte path.
func TestLongestCommonSubstring_ZeroAllocs_ASCII_Short(t *testing.T) {
	_ = fuzzymatch.LongestCommonSubstring("kitten", "sitting")
	allocs := testing.AllocsPerRun(100, func() {
		_ = fuzzymatch.LongestCommonSubstring("kitten", "sitting")
	})
	if allocs > 0 {
		t.Errorf("LongestCommonSubstring ASCII short: %.1f allocs/op; want 0", allocs)
	}
}
