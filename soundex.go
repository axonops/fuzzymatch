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

// soundex.go implements the Soundex phonetic algorithm (Knuth/Census variant)
// for the fuzzymatch catalogue.
//
// Source: Knuth, D. E. (1973). The Art of Computer Programming, Vol. 3:
// Sorting and Searching, Section 6.4. Addison-Wesley. — canonical algorithm
// description for the Knuth/Census (American Soundex) variant.
//
// Algorithmic origin: Russell, R. C., Odell, M. K. (1918, 1922). U.S. Patents
// 1,261,167 and 1,435,663. Now public domain; cited for completeness. The
// authoritative algorithm description used for this implementation is Knuth's
// TAOCP Vol. 3 §6.4.
//
// Algorithm (Knuth/Census variant — NOT the SQL/MySQL variant):
//
//  1. Retain the first letter of the name (uppercased).
//  2. Replace consonants with digits (after the first letter):
//       B, F, P, V          → 1
//       C, G, J, K, Q, S, X, Z → 2
//       D, T               → 3
//       L                  → 4
//       M, N               → 5
//       R                  → 6
//  3. H and W are SKIPPED (they do NOT break adjacent-same-group consonant
//     collapse). This is the defining distinction of the Knuth/Census
//     variant vs. the SQL/MySQL variant where H/W ARE separators.
//  4. Vowels (A, E, I, O, U) and Y act as separators — they reset the
//     "last group" to a sentinel value so the next consonant is not
//     suppressed even if it shares a group with the previous consonant.
//     Vowels themselves are NOT output.
//  5. Same-group adjacent consonants (after H/W skip) collapse to one digit.
//  6. If two adjacent consonants are separated only by H or W AND they
//     are in the same group, they collapse to one digit. If separated by
//     a vowel, they produce two separate digits.
//  7. Pad with zeros to 4 characters (1 letter + 3 digits) if needed.
//  8. Empty input returns "".
//
// Variant gate (LOAD-BEARING):
//   SoundexCode("Tymczak") == "T522"   (NOT "T520" which the SQL variant returns)
//   SoundexCode("Ashcraft") == SoundexCode("Ashcroft") == "A261"
//   (H/W in the middle do NOT reset the adjacent-group collapse)
//
// Non-ASCII input handling: this algorithm operates on ASCII letters
// [A-Za-z] only. Non-ASCII runes (accented characters, emoji,
// combining marks) are dropped silently before encoding. For
// Unicode-aware similarity on non-ASCII input, compose with
// Normalise + diacritic stripping before calling this function, or
// use a character-based algorithm (e.g. Levenshtein, Jaro-Winkler).
//
// Source-origin discipline (per algorithm-licensing-standards):
//
//   - Primary source:       Knuth, D. E. (1973). TAOCP Vol. 3, §6.4
//                           (Algorithmic origin: Russell & Odell 1918/1922)
//   - Cross-validation:     jellyfish==1.2.1 (BSD-2-Clause) — consulted
//                           ONLY for reference-vector cross-validation.
//                           NOT for code copying. jellyfish 1.2.1 also uses
//                           the Knuth/Census H/W-skip variant (confirmed by
//                           direct read of jellyfish/src/soundex.rs).
//   - GPL/LGPL provenance:  none.
//   - Code copied verbatim: none.
//   - MIT-licensed Go ports NOT consulted for code:
//       github.com/xrash/smetrics (MIT) — Soundex implementation
//       github.com/tilotech/go-phonetics (MIT) — Soundex + Metaphone
//       github.com/UjjwalAyyangar/go-jellyfish (MIT) — phonetic algorithms
//
// Implementation discipline:
//
//   - ASCII byte-scan — input processed as []byte; rune scan for non-ASCII
//     skip per CONTEXT.md §5 silent-skip discipline.
//   - Stack-allocated [4]byte result buffer — zero allocations for any input.
//   - NO init()-time table builds (per docs/requirements.md §5(12)).
//   - NO map iteration on output paths (DET-03).
//   - NO transcendental float operations (DET-06) — no floats at all.
//   - NO goroutines, channels, or mutexes.
//   - Identity short-circuit `if a == b { return 1.0 }` in SoundexScore
//     BEFORE any computation (covers both-empty per algorithm-correctness-
//     standards both-empty → 1.0 convention).
//
// Public surface (two functions):
//
//   - SoundexCode(s string) string
//   - SoundexScore(a, b string) float64
//
// Only SoundexScore is registered in the dispatch table (slot 23 —
// AlgoSoundex — see algoid.go and dispatch_soundex.go).

package fuzzymatch

// soundexNoGroup is the sentinel "no previous group" value for the lastGroup
// tracking variable. Initialised to 0 which is not a valid Soundex group
// digit (groups are 1-6).
const soundexNoGroup byte = 0

// soundexGroup returns the Soundex group digit (1-6) for a given uppercase
// ASCII letter, or soundexNoGroup (0) if the letter is not coded (vowels,
// H, W, Y). H and W return 0 here; their special handling (skip without
// resetting lastGroup) is in SoundexCode.
func soundexGroup(c byte) byte {
	switch c {
	case 'B', 'F', 'P', 'V':
		return 1
	case 'C', 'G', 'J', 'K', 'Q', 'S', 'X', 'Z':
		return 2
	case 'D', 'T':
		return 3
	case 'L':
		return 4
	case 'M', 'N':
		return 5
	case 'R':
		return 6
	}
	// A, E, I, O, U, Y → vowel/separator (returns 0 = soundexNoGroup)
	// H, W → skip without group change (caller handles this case)
	return soundexNoGroup
}

// SoundexCode returns the 4-character Soundex phonetic code for s using the
// Knuth/Census variant (Knuth TAOCP Vol. 3 §6.4). The result is always
// exactly 4 characters (1 uppercase letter + 3 digits, zero-padded) for
// non-empty ASCII input, or the empty string for empty input.
//
// Variant choice (LOAD-BEARING):
//   - H and W are SKIPPED — they do NOT break adjacent-consonant collapse.
//     SoundexCode("Ashcraft") == "A261" (H is transparent).
//   - Tymczak → "T522" (NOT "T520" which SQL/MySQL Soundex returns).
//
// Non-ASCII input handling: this algorithm operates on ASCII letters
// [A-Za-z] only. Non-ASCII runes (accented characters, emoji,
// combining marks) are dropped silently before encoding. For
// Unicode-aware similarity on non-ASCII input, compose with
// Normalise + diacritic stripping before calling this function, or
// use a character-based algorithm (e.g. Levenshtein, Jaro-Winkler).
//
// Performance scope (Q7b, docs/requirements.md §14.1):
//
//	The published budget is ≤ 1 alloc per call. The implementation uses a
//	stack-allocated [4]byte result buffer; the lone heap allocation is the
//	`string(result[:4])` conversion on return — structurally unavoidable
//	without `unsafe.String`, which the project policy declines. The 0-alloc
//	target documented in earlier drafts was unachievable for the same reason.
func SoundexCode(s string) string { //nolint:gocyclo // canonical Russell-Knuth 6-rule encoder; each rule contributes a distinct branch for letter-class dispatch + adjacency dedup + Y/H pass-through + leading-letter preservation
	if len(s) == 0 {
		return ""
	}

	// Stack-allocated result buffer: [0] = first letter, [1..3] = digits.
	var result [4]byte
	digits := 1 // next write position in result (1-based: 1,2,3)

	// Scan bytes for the first ASCII letter to use as the initial letter.
	var firstByte byte
	var startIdx int
	found := false
	for i := 0; i < len(s); {
		b := s[i]
		if b < 0x80 {
			// ASCII
			if (b >= 'A' && b <= 'Z') || (b >= 'a' && b <= 'z') {
				if b >= 'a' {
					b -= 32 // to uppercase
				}
				firstByte = b
				startIdx = i + 1
				found = true
			}
			i++
		} else {
			// Non-ASCII: skip the full UTF-8 sequence.
			sz := runeSizeAt(s, i)
			i += sz
		}
		if found {
			break
		}
	}
	if !found {
		// No ASCII letters in input.
		return ""
	}

	result[0] = firstByte

	// The last group encoded; initialised to the group of the first letter
	// so that if the second letter is in the same group it is suppressed.
	// Per Knuth §6.4: "for instance, the name Pfister is coded P236 not P126."
	lastGroup := soundexGroup(firstByte)

	// Scan remaining bytes.
	for i := startIdx; i < len(s) && digits < 4; {
		b := s[i]
		if b < 0x80 {
			// ASCII
			i++
			if b >= 'a' && b <= 'z' {
				b -= 32 // to uppercase
			}
			if b < 'A' || b > 'Z' {
				// Non-letter ASCII (digit, punctuation, space) — skip,
				// treating like a vowel separator to reset lastGroup.
				lastGroup = soundexNoGroup
				continue
			}
			// H and W: skip transparently without changing lastGroup.
			if b == 'H' || b == 'W' {
				continue
			}
			// Vowels (A, E, I, O, U) and Y: reset lastGroup (separator).
			if b == 'A' || b == 'E' || b == 'I' || b == 'O' || b == 'U' || b == 'Y' {
				lastGroup = soundexNoGroup
				continue
			}
			// Consonant: compute its group.
			g := soundexGroup(b)
			if g == soundexNoGroup {
				continue // should not happen for a-z after H/W/vowel handled above
			}
			// Suppress if same group as last encoded consonant.
			if g == lastGroup {
				continue
			}
			// Encode the digit.
			result[digits] = '0' + g
			digits++
			lastGroup = g
		} else {
			// Non-ASCII: skip the full UTF-8 sequence silently.
			sz := runeSizeAt(s, i)
			i += sz
		}
	}

	// Zero-pad to 4 characters.
	for digits < 4 {
		result[digits] = '0'
		digits++
	}

	return string(result[:])
}

// SoundexScore returns 1.0 if a and b share the same non-empty Soundex code
// (Knuth/Census variant), and 0.0 otherwise. Empty strings: both-empty returns
// 1.0 (covered by the identity short-circuit); one-empty returns 0.0.
//
// For programmatic input-quality checks before scoring (including
// WarnAllNonASCIIDropped scoped to AlgoSoundex when the input collapses
// to empty after the ASCII-only path), see [fuzzymatch.Validate].
//
// Non-ASCII input handling: this algorithm operates on ASCII letters
// [A-Za-z] only. Non-ASCII runes (accented characters, emoji,
// combining marks) are dropped silently before encoding. For
// Unicode-aware similarity on non-ASCII input, compose with
// Normalise + diacritic stripping before calling this function, or
// use a character-based algorithm (e.g. Levenshtein, Jaro-Winkler).
func SoundexScore(a, b string) float64 {
	// Identity short-circuit covers both-empty (1.0) and same-string (1.0).
	if a == b {
		return 1.0
	}
	ca := SoundexCode(a)
	cb := SoundexCode(b)
	// Both codes must be non-empty and equal for a match.
	if ca != "" && ca == cb {
		return 1.0
	}
	return 0.0
}

// runeSizeAt returns the UTF-8 byte length of the rune starting at
// position i in s. Used by SoundexCode and the other phonetic encoders
// for non-ASCII skip without importing unicode/utf8 (the rune itself
// is never inspected on these skip paths — only the byte size matters).
// The byte-size dispatch matches utf8.DecodeRuneInString's rune-size
// semantics: invalid lead bytes / truncated sequences return 1 so the
// caller advances at least one byte.
func runeSizeAt(s string, i int) int {
	b0 := s[i]
	switch {
	case b0 < 0x80:
		return 1
	case b0 < 0xC0:
		// Continuation byte out of context — treat as 1-byte invalid.
		return 1
	case b0 < 0xE0:
		if i+1 >= len(s) {
			return 1
		}
		return 2
	case b0 < 0xF0:
		if i+2 >= len(s) {
			return 1
		}
		return 3
	}
	if i+3 >= len(s) {
		return 1
	}
	return 4
}
