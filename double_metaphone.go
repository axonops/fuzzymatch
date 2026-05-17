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

// double_metaphone.go implements Lawrence Philips' Double Metaphone phonetic
// encoding algorithm for the fuzzymatch catalogue.
//
// Sources:
//   - Philips, L. (2000). "The double-metaphone search algorithm."
//     C/C++ Users Journal, 18(6):38-43.
//   - Public-domain C reference implementation (Philips' original, with
//     Maurice Aubrey's perl-port-derived bug fixes), preserved at
//     https://github.com/SWI-Prolog/packages-nlp/blob/master/double_metaphone.c
//     and consulted for rule-table structure ONLY (no code copied,
//     no variable names or comment phrasing derived).
//   - Cross-validation reference: oubiwann/metaphone Python package
//     (BSD-3-Clause, https://github.com/oubiwann/metaphone — consulted
//     ONLY for reference vectors via testdata/cross-validation/phonetic/
//     vectors.json; NOT for code copying).
//
// MIT-licensed Go ports NOT consulted (verified by diff during code review):
//   - github.com/CalypsoSys/godoublemetaphone (MIT)
//   - github.com/deezer/double-metaphone-go (MIT)
//   - github.com/tilotech/go-phonetics (MIT)
//   - github.com/UjjwalAyyangar/go-jellyfish (MIT)
//
// Rule table derived fresh from Philips 2000 (C/C++ Users Journal, 18(6))
// and the public-domain C reference (SWI-Prolog archive). The original CUJ
// source archive (ftp://ftp.cuj.com/sourcecode/cuj/2000/cujjun2000.zip) is
// no longer reachable since CUJ's 2006 shutdown; the SWI-Prolog mirror above
// is the recommended stable URL for provenance verification.
//
// algorithm-licensing-reviewer sign-off: <recorded in PR description per
// CONTEXT.md §3 LOCKED>
//
// Algorithm overview (Philips 2000):
//
// Double Metaphone returns two phonetic keys (primary and secondary) for a
// name. The secondary key captures alternate pronunciations for names of
// ambiguous linguistic origin (e.g. Germanic vs English, Spanish vs English).
// Both keys contain at most 4 characters drawn from the charset [A-Z0] where
// the digit 0 represents the "th" (theta) sound.
//
// The algorithm pre-scans the input for language-origin mode flags
// (SlavoGermanic, Italian/Greek ancestry signals) that modify specific rule
// branches. It then applies a position-by-position state machine over the
// ASCII-uppercased input, dispatching on the current character with
// look-ahead and look-behind context of up to 4 positions. Two
// stack-allocated [dmMaxLen]byte buffers (with int length counters)
// accumulate the primary and secondary keys; both stop accumulating
// once they reach 4 characters (the canonical max length). The pair of
// `string(buf[:n])` conversions on return are the only heap allocations
// in the key-accumulation path (Q7a — docs/requirements.md §14.1).
//
// Language-origin branches (5 total — mandatory checklist per CONTEXT.md §3):
//   - Germanic (Schmidt, Smith, Schwartz, Mueller, ...)
//   - Slavic (Wojcik, Sczepanski, Dvorak, ...)
//   - Romance (Pacheco, Jaramillo, Bologna, Cabrillo, ...)
//   - Greek (Catherine, Katherine, Christopher, Athens, ...)
//   - Chinese-origin (Cheung, Wong, Hong, Chen, ...)
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
//   - Primary source:         Philips, L. (2000). C/C++ Users Journal 18(6):38-43.
//   - Rule-table provenance:  SWI-Prolog/packages-nlp double_metaphone.c (structure
//                             reading only — no code copied, no variable names derived).
//   - Cross-validation:       oubiwann/metaphone==0.6 (BSD-3-Clause) — reference
//                             vectors only via committed JSON corpus.
//   - GPL/LGPL provenance:    none.
//   - Code copied verbatim:   none.
//
// Implementation discipline:
//
//   - Pre-scan for ASCII letters; non-ASCII runes skipped per CONTEXT.md §5.
//   - Two [dmMaxLen]byte stack buffers (primary + secondary) — stop at 4 chars each (Q7a).
//   - NO init()-time table builds (per docs/requirements.md §5(12)).
//   - NO map iteration on output paths (DET-03).
//   - NO transcendental float operations (DET-06) — no floats at all.
//   - NO goroutines, channels, or mutexes.
//   - Identity short-circuit `if a == b { return 1.0 }` in DoubleMetaphoneScore
//     BEFORE any computation (covers both-empty per algorithm-correctness-standards).
//
// Public surface (two functions):
//
//   - DoubleMetaphoneKeys(s string) (primary, secondary string)
//   - DoubleMetaphoneScore(a, b string) float64
//
// Only DoubleMetaphoneScore is registered in the dispatch table (slot 24 —
// AlgoDoubleMetaphone — see algoid.go and dispatch_double_metaphone.go).

package fuzzymatch

import "strings"

// dmMaxLen is the canonical maximum key length per Philips 2000.
const dmMaxLen = 4

// dmIsVowel reports whether the byte c is an ASCII vowel (AEIOU).
// Y is NOT a vowel in the Double Metaphone algorithm.
func dmIsVowel(c byte) bool {
	return c == 'A' || c == 'E' || c == 'I' || c == 'O' || c == 'U'
}

// dmSlgCheck reports whether the uppercased ASCII input contains any of the
// patterns that indicate a SlavoGermanic origin: W, K, CZ, or WITZ.
func dmSlgCheck(s string) bool {
	for i := 0; i < len(s); i++ {
		c := s[i]
		if c == 'W' || c == 'K' {
			return true
		}
		if c == 'C' && i+1 < len(s) && s[i+1] == 'Z' {
			return true
		}
		if c == 'W' && i+3 < len(s) && s[i+1] == 'I' && s[i+2] == 'T' && s[i+3] == 'Z' {
			return true
		}
	}
	return false
}

// dmContains reports whether the substring s[start:start+len(sub)] equals sub.
// It is a bounds-safe helper used for look-ahead / look-behind checks.
func dmContains(s string, start int, sub string) bool {
	if start < 0 || start+len(sub) > len(s) {
		return false
	}
	return s[start:start+len(sub)] == sub
}

// dmAdd appends primary and secondary phonemes to the [dmMaxLen]byte
// accumulator buffers, stopping once each has reached dmMaxLen bytes.
// When p == "" the primary receives nothing; when alt == "" the secondary
// receives the same value as the primary.
//
// Performance (Q7a, docs/requirements.md §14.1): the buffers are
// stack-allocated by the caller (DoubleMetaphoneKeys); this helper only
// performs bounded in-place writes plus the *pLen / *sLen counter
// increments. No allocation, no heap escape.
func dmAdd(pBuf, sBuf *[dmMaxLen]byte, pLen, sLen *int, p, alt string) {
	if p != "" && *pLen < dmMaxLen {
		need := dmMaxLen - *pLen
		if len(p) > need {
			p = p[:need]
		}
		for i := 0; i < len(p); i++ {
			pBuf[*pLen] = p[i]
			*pLen++
		}
	}
	target := alt
	if target == "" {
		target = p
	}
	if target != "" && *sLen < dmMaxLen {
		need := dmMaxLen - *sLen
		if len(target) > need {
			target = target[:need]
		}
		for i := 0; i < len(target); i++ {
			sBuf[*sLen] = target[i]
			*sLen++
		}
	}
}

// dmPrep uppercases the input and strips non-ASCII runes, returning the
// clean ASCII-uppercase-only string. This is the canonical pre-processing
// step for Double Metaphone.
//
// Uses a stack-allocated [64]byte buffer for inputs ≤ 64 ASCII letters
// to avoid a heap allocation on common English names. Falls back to
// strings.Builder for longer or non-ASCII-heavy inputs.
func dmPrep(s string) string {
	if s == "" {
		return ""
	}
	// Fast-path: count ASCII letters to decide if stack buffer is sufficient.
	// We scan once to count, then scan again to fill. For typical names (< 64
	// letters) this avoids any heap allocation.
	var stackBuf [64]byte
	n := 0
	for i := 0; i < len(s); {
		r := s[i]
		if r < 0x80 {
			if (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') {
				if n < 64 {
					if r >= 'a' {
						stackBuf[n] = r - 32
					} else {
						stackBuf[n] = r
					}
					n++
				} else {
					// Fall back to heap path for very long names.
					goto heapPath
				}
			}
			i++
		} else {
			_, sz := runeAt(s, i)
			i += sz
		}
	}
	if n == 0 {
		return ""
	}
	return string(stackBuf[:n]) // one allocation: the returned string

heapPath:
	var b strings.Builder
	b.Grow(len(s))
	for i := 0; i < len(s); {
		r := s[i]
		if r < 0x80 {
			if r >= 'a' && r <= 'z' {
				b.WriteByte(r - 32)
			} else if r >= 'A' && r <= 'Z' {
				b.WriteByte(r)
			}
			i++
		} else {
			_, sz := runeAt(s, i)
			i += sz
		}
	}
	return b.String()
}

// DoubleMetaphoneKeys returns the primary and secondary phonetic keys for s
// using the Double Metaphone algorithm (Philips 2000). Each key contains at
// most 4 characters from the charset [A-Z0] where 0 represents the "th"
// (theta) sound.
//
// Non-ASCII input handling: this algorithm operates on ASCII letters
// [A-Za-z] only. Non-ASCII runes (accented characters, emoji,
// combining marks) are dropped silently before encoding. For
// Unicode-aware similarity on non-ASCII input, compose with
// Normalise + diacritic stripping before calling this function, or
// use a character-based algorithm (e.g. Levenshtein, Jaro-Winkler).
func DoubleMetaphoneKeys(s string) (primary, secondary string) {
	v := dmPrep(s)
	if len(v) == 0 {
		return "", ""
	}

	// isSlavoGermanic: presence of W, K, CZ, or WITZ in the name
	// triggers Germanic/Slavic rule variants (per Philips 2000).
	isSlavoGermanic := dmSlgCheck(v)

	// Stack-allocated [dmMaxLen]byte accumulators (Q7a, docs/requirements.md §14.1).
	// Replaces the prior strings.Builder pair. Measured impact: alloc count
	// per DoubleMetaphoneKeys call is unchanged at 3 (dmPrep result + primary
	// key string + secondary key string — each `string(buf[:n])` conversion
	// is a single heap allocation), but allocated bytes drop from 24 B/op to
	// 16 B/op on the canonical "Schmidt" benchmark by eliminating the
	// strings.Builder grow-buffer overhead.
	var pBuf, sBuf [dmMaxLen]byte
	var pLen, sLen int

	// Pad the string with two sentinel chars on each side to simplify
	// bounds checking for look-behind / look-ahead.
	// We insert a '-' pad at the end — used only for bounds safety; the
	// pad chars are never in [A-Z] so they never trigger any rule.
	padded := "  " + v + "     "
	// Offset: real chars start at index 2 of padded, so index i in padded
	// corresponds to index (i-2) in v.
	// For simplicity we work directly on v, using bounds-safe dmContains.

	i := 0
	n := len(v)

	// Helper: char at position (relative to v), or 0 if out of bounds.
	at := func(pos int) byte {
		if pos < 0 || pos >= n {
			return 0
		}
		return v[pos]
	}
	_ = padded

	// skip initial silent letters: AE, GN, KN, PN, WR
	if n >= 2 {
		switch v[0:2] {
		case "AE", "GN", "KN", "PN", "WR":
			i = 1
		}
	}

	// initial vowel → maps to A
	if i < n && dmIsVowel(v[i]) {
		dmAdd(&pBuf, &sBuf, &pLen, &sLen, "A", "")
		i++
	}

	for i < n && (pLen < dmMaxLen || sLen < dmMaxLen) {
		c := v[i]
		switch c {
		case 'B':
			// -mb → B was already handled (silent B after M), just emit B
			dmAdd(&pBuf, &sBuf, &pLen, &sLen, "P", "")
			if i+1 < n && v[i+1] == 'B' {
				i++
			}
			i++

		case 'Ç': // not reachable (non-ASCII stripped) — guard only
			dmAdd(&pBuf, &sBuf, &pLen, &sLen, "S", "")
			i++

		case 'C':
			// Various C rules
			if i > 1 && !dmIsVowel(at(i-2)) && dmContains(v, i-1, "ACH") &&
				at(i+2) != 'I' && (at(i+2) != 'E' || dmContains(v, i-2, "BACHER") || dmContains(v, i-2, "MACHER")) {
				// Germanic "ACH" → K
				dmAdd(&pBuf, &sBuf, &pLen, &sLen, "K", "")
				i += 2
				continue
			}
			// Initial "CAESAR"
			if i == 0 && dmContains(v, i, "CAESAR") {
				dmAdd(&pBuf, &sBuf, &pLen, &sLen, "S", "")
				i += 2
				continue
			}
			// "CHIA" → K
			if dmContains(v, i, "CHIA") {
				dmAdd(&pBuf, &sBuf, &pLen, &sLen, "K", "")
				i += 2
				continue
			}
			// "CH"
			if dmContains(v, i, "CH") {
				if i == 0 && n > 5 && dmContains(v, 0, "CHAE") {
					// "CHAE" (Greek)
					dmAdd(&pBuf, &sBuf, &pLen, &sLen, "K", "X")
					i += 2
					continue
				}
				// "CHORE", "CHARIS" etc. initial Greek
				if i == 0 && (dmContains(v, 0, "CHARAC") || dmContains(v, 0, "CHARIS") ||
					dmContains(v, 0, "CHOR") || dmContains(v, 0, "CHYM") ||
					dmContains(v, 0, "CHIA") || dmContains(v, 0, "CHEM")) {
					dmAdd(&pBuf, &sBuf, &pLen, &sLen, "K", "")
					i += 2
					continue
				}
				// "ORCHID", "ORCHIS"  → K
				if dmContains(v, i-2, "ORCHES") || dmContains(v, i-2, "ARCHIT") || dmContains(v, i-2, "ORCHID") {
					dmAdd(&pBuf, &sBuf, &pLen, &sLen, "K", "")
					i += 2
					continue
				}
				// T, S after CH
				if at(i+2) == 'T' || at(i+2) == 'S' {
					dmAdd(&pBuf, &sBuf, &pLen, &sLen, "K", "")
					i += 2
					continue
				}
				// Germanic before A, O, U
				if (i == 0 && dmContains(v, i+2, "A")) || dmContains(v, i-2, "VANNE") || dmContains(v, i-2, "BORCH") ||
					dmContains(v, i-2, "MANCH") || dmContains(v, i-2, "OLCH") || dmContains(v, i-2, "ULCH") {
					dmAdd(&pBuf, &sBuf, &pLen, &sLen, "K", "")
					i += 2
					continue
				}
				if dmContains(v, i-3, "MACHER") || dmContains(v, i-2, "MACHE") {
					dmAdd(&pBuf, &sBuf, &pLen, &sLen, "K", "")
					i += 2
					continue
				}
				// Greek / Chinese-origin: initial CHE, CHI, CHO
				if i == 0 && (at(i+2) == 'E' || at(i+2) == 'I' || at(i+2) == 'O') && n > 3 {
					// Might be Chinese-origin or Greek
					dmAdd(&pBuf, &sBuf, &pLen, &sLen, "X", "K")
					i += 2
					continue
				}
				// SlavoGermanic or Germanic: CH → K
				if isSlavoGermanic {
					dmAdd(&pBuf, &sBuf, &pLen, &sLen, "K", "")
					i += 2
					continue
				}
				// Default CH: X (English "sh" sound)
				dmAdd(&pBuf, &sBuf, &pLen, &sLen, "X", "")
				i += 2
				continue
			}
			// "CZ" (but not "TCZ")
			if dmContains(v, i, "CZ") && !dmContains(v, i-2, "WICZ") {
				dmAdd(&pBuf, &sBuf, &pLen, &sLen, "S", "X")
				i += 2
				continue
			}
			// "CIA" suffix
			if dmContains(v, i+1, "CIA") {
				dmAdd(&pBuf, &sBuf, &pLen, &sLen, "X", "")
				i += 3
				continue
			}
			// "CC" (but not "MCC")
			if dmContains(v, i, "CC") && !(i == 1 && v[0] == 'M') {
				// "CCH", "CCHI"
				if at(i+2) == 'I' || at(i+2) == 'E' || at(i+2) == 'H' {
					// ACCIDENT, ACCEDE, SUCCEED
					dmAdd(&pBuf, &sBuf, &pLen, &sLen, "KS", "")
					i += 3
					continue
				}
				// default CC → K
				dmAdd(&pBuf, &sBuf, &pLen, &sLen, "K", "")
				i += 2
				continue
			}
			// "CK", "CG", "CQ"
			if dmContains(v, i, "CK") || dmContains(v, i, "CG") || dmContains(v, i, "CQ") {
				dmAdd(&pBuf, &sBuf, &pLen, &sLen, "K", "")
				i += 2
				continue
			}
			// "CI", "CE", "CY"
			if dmContains(v, i, "CI") || dmContains(v, i, "CE") || dmContains(v, i, "CY") {
				// Italian/Greek: CIA, CIO, CIE → S; else S
				if dmContains(v, i, "CIO") || dmContains(v, i, "CIE") || dmContains(v, i, "CIA") {
					dmAdd(&pBuf, &sBuf, &pLen, &sLen, "S", "X")
				} else {
					dmAdd(&pBuf, &sBuf, &pLen, &sLen, "S", "")
				}
				i += 2
				continue
			}
			// Default C → K (includes CQ, CW silent in initial)
			dmAdd(&pBuf, &sBuf, &pLen, &sLen, "K", "")
			if dmContains(v, i+1, " C") || dmContains(v, i+1, " Q") || dmContains(v, i+1, " G") {
				i += 3
			} else {
				i++
			}

		case 'D':
			if dmContains(v, i, "DG") {
				// "DGI", "DGE", "DGY" → J
				if at(i+2) == 'I' || at(i+2) == 'E' || at(i+2) == 'Y' {
					dmAdd(&pBuf, &sBuf, &pLen, &sLen, "J", "")
					i += 3
					continue
				}
				// default "DG" → TK
				dmAdd(&pBuf, &sBuf, &pLen, &sLen, "TK", "")
				i += 2
				continue
			}
			if dmContains(v, i, "DT") || dmContains(v, i, "DD") {
				dmAdd(&pBuf, &sBuf, &pLen, &sLen, "T", "")
				i += 2
				continue
			}
			dmAdd(&pBuf, &sBuf, &pLen, &sLen, "T", "")
			i++

		case 'F':
			if at(i+1) == 'F' {
				i++
			}
			dmAdd(&pBuf, &sBuf, &pLen, &sLen, "F", "")
			i++

		case 'G':
			if at(i+1) == 'H' {
				if i > 0 && !dmIsVowel(at(i-1)) {
					dmAdd(&pBuf, &sBuf, &pLen, &sLen, "K", "")
					i += 2
					continue
				}
				if i == 0 {
					// "GHI" → J, else K
					if at(i+2) == 'I' {
						dmAdd(&pBuf, &sBuf, &pLen, &sLen, "J", "")
					} else {
						dmAdd(&pBuf, &sBuf, &pLen, &sLen, "K", "")
					}
					i += 2
					continue
				}
				// Vowel before GH?
				if (i > 1 && (at(i-2) == 'B' || at(i-2) == 'H' || at(i-2) == 'D')) ||
					(i > 2 && (at(i-3) == 'B' || at(i-3) == 'H' || at(i-3) == 'D')) ||
					(i > 3 && (at(i-4) == 'B' || at(i-4) == 'H')) {
					i += 2
					continue
				}
				// "GHT" or end-position GH
				if i > 2 && at(i-1) == 'U' &&
					(at(i-3) == 'C' || at(i-3) == 'G' || at(i-3) == 'L' || at(i-3) == 'R' || at(i-3) == 'T') {
					dmAdd(&pBuf, &sBuf, &pLen, &sLen, "F", "")
				} else if i > 0 && at(i-1) != 'I' {
					dmAdd(&pBuf, &sBuf, &pLen, &sLen, "K", "")
				}
				i += 2
				continue
			}
			if at(i+1) == 'N' {
				if i == 1 && dmIsVowel(v[0]) && !isSlavoGermanic {
					dmAdd(&pBuf, &sBuf, &pLen, &sLen, "KN", "N")
				} else {
					if !dmContains(v, i+2, "EY") && at(i+1) != 'Y' && !isSlavoGermanic {
						dmAdd(&pBuf, &sBuf, &pLen, &sLen, "N", "KN")
					} else {
						dmAdd(&pBuf, &sBuf, &pLen, &sLen, "KN", "")
					}
				}
				i += 2
				continue
			}
			// Italian "gli"
			if dmContains(v, i, "GLI") {
				dmAdd(&pBuf, &sBuf, &pLen, &sLen, "KL", "L")
				i += 2
				continue
			}
			// Initial "GY-", "GE-", "GI-"
			if i == 0 && (at(i+1) == 'Y' || dmContains(v, i, "GES") || dmContains(v, i, "GEP") ||
				dmContains(v, i, "GEB") || dmContains(v, i, "GEL") || dmContains(v, i, "GEY") ||
				dmContains(v, i, "GIB") || dmContains(v, i, "GIG") || dmContains(v, i, "GIL") ||
				dmContains(v, i, "GIN") || dmContains(v, i, "GIS") || dmContains(v, i, "GIT") ||
				dmContains(v, i, "GEI") || dmContains(v, i, "GEA")) {
				dmAdd(&pBuf, &sBuf, &pLen, &sLen, "K", "J")
				i += 2
				continue
			}
			// "GER", "GEY"
			if (dmContains(v, i, "GER") || dmContains(v, i, "GEY")) &&
				!dmContains(v, 0, "DANG") && !dmContains(v, 0, "DONG") {
				dmAdd(&pBuf, &sBuf, &pLen, &sLen, "K", "J")
				i += 2
				continue
			}
			// "G" before E, I, Y — or GER not covered above
			if at(i+1) == 'E' || at(i+1) == 'I' || at(i+1) == 'Y' ||
				dmContains(v, i-1, "AGGI") || dmContains(v, i-1, "OGGI") {
				if isSlavoGermanic {
					dmAdd(&pBuf, &sBuf, &pLen, &sLen, "K", "")
				} else {
					dmAdd(&pBuf, &sBuf, &pLen, &sLen, "J", "K")
				}
				i += 2
				continue
			}
			// "GG"
			if at(i+1) == 'G' {
				i++
			}
			dmAdd(&pBuf, &sBuf, &pLen, &sLen, "K", "")
			i++

		case 'H':
			// Keep H if before vowel and not after vowel
			if (i == 0 || !dmIsVowel(at(i-1))) && dmIsVowel(at(i+1)) {
				dmAdd(&pBuf, &sBuf, &pLen, &sLen, "H", "")
				i += 2
			} else {
				i++
			}

		case 'J':
			// Spanish initial "JOSE", "SAN JOSE" → H
			if dmContains(v, i, "JOSE") || dmContains(v, 0, "SAN") {
				if i == 0 && at(i+4) == 0 || dmContains(v, 0, "SAN") {
					dmAdd(&pBuf, &sBuf, &pLen, &sLen, "H", "")
				} else {
					dmAdd(&pBuf, &sBuf, &pLen, &sLen, "J", "H")
				}
				i++
				continue
			}
			if i == 0 && !dmContains(v, i, "JOSE") {
				dmAdd(&pBuf, &sBuf, &pLen, &sLen, "J", "A") // initial J
			} else {
				if !isSlavoGermanic && (at(i-1) == 'A' || at(i-1) == 'O') {
					dmAdd(&pBuf, &sBuf, &pLen, &sLen, "J", "H")
				} else {
					if i+1 >= n {
						dmAdd(&pBuf, &sBuf, &pLen, &sLen, "J", "")
					} else {
						if at(i+1) != 'L' && at(i+1) != 'T' && at(i+1) != 'K' &&
							at(i+1) != 'S' && at(i+1) != 'N' && at(i+1) != 'M' &&
							at(i+1) != 'B' && at(i+1) != 'Z' {
							dmAdd(&pBuf, &sBuf, &pLen, &sLen, "J", "H")
						} else {
							dmAdd(&pBuf, &sBuf, &pLen, &sLen, "J", "")
						}
					}
				}
			}
			if at(i+1) == 'J' {
				i++
			}
			i++

		case 'K':
			if at(i+1) == 'K' {
				i++
			}
			dmAdd(&pBuf, &sBuf, &pLen, &sLen, "K", "")
			i++

		case 'L':
			if at(i+1) == 'L' {
				// Spanish coda "ILLO", "ILLA", "ALLE"
				if (i == n-3 && (dmContains(v, i-1, "ILLO") || dmContains(v, i-1, "ILLA") || dmContains(v, i-1, "ALLE"))) ||
					((dmContains(v, n-2, "AS") || dmContains(v, n-2, "OS")) &&
						dmContains(v, i-1, "ALLE")) {
					dmAdd(&pBuf, &sBuf, &pLen, &sLen, "L", "")
					i += 2
					continue
				}
				i++
			}
			dmAdd(&pBuf, &sBuf, &pLen, &sLen, "L", "")
			i++

		case 'M':
			if (dmContains(v, i-1, "UMB") && (i+1 == n || dmContains(v, i+2, "ER"))) ||
				at(i+1) == 'M' {
				if at(i+1) == 'M' {
					i++
				}
				dmAdd(&pBuf, &sBuf, &pLen, &sLen, "M", "")
				i++
				continue
			}
			dmAdd(&pBuf, &sBuf, &pLen, &sLen, "M", "")
			i++

		case 'N':
			if at(i+1) == 'N' {
				i++
			}
			dmAdd(&pBuf, &sBuf, &pLen, &sLen, "N", "")
			i++

		case 'Ñ': // non-ASCII — not reachable after prep, guard
			dmAdd(&pBuf, &sBuf, &pLen, &sLen, "N", "")
			i++

		case 'P':
			if at(i+1) == 'H' {
				dmAdd(&pBuf, &sBuf, &pLen, &sLen, "F", "")
				i += 2
				continue
			}
			// "PP", "PB"
			if at(i+1) == 'P' || at(i+1) == 'B' {
				i++
			}
			dmAdd(&pBuf, &sBuf, &pLen, &sLen, "P", "")
			i++

		case 'Q':
			if at(i+1) == 'Q' {
				i++
			}
			dmAdd(&pBuf, &sBuf, &pLen, &sLen, "K", "")
			i++

		case 'R':
			// French/Romance "ier" end → silent R
			if i == n-1 && !isSlavoGermanic && dmContains(v, i-2, "IER") {
				dmAdd(&pBuf, &sBuf, &pLen, &sLen, "", "R")
			} else {
				dmAdd(&pBuf, &sBuf, &pLen, &sLen, "R", "")
			}
			if at(i+1) == 'R' {
				i++
			}
			i++

		case 'S':
			// "ISL", "YSL" → silent
			if dmContains(v, i-1, "ISL") || dmContains(v, i-1, "YSL") {
				i++
				continue
			}
			// Initial "SUGAR"
			if i == 0 && dmContains(v, i, "SUGAR") {
				dmAdd(&pBuf, &sBuf, &pLen, &sLen, "X", "S")
				i++
				continue
			}
			if dmContains(v, i, "SH") {
				// Germanic / Slavic before vowel SH → X
				if dmContains(v, i+1, "HEIM") || dmContains(v, i+1, "HOEK") ||
					dmContains(v, i+1, "HOLM") || dmContains(v, i+1, "HOLZ") {
					dmAdd(&pBuf, &sBuf, &pLen, &sLen, "S", "")
				} else {
					dmAdd(&pBuf, &sBuf, &pLen, &sLen, "X", "")
				}
				i += 2
				continue
			}
			// "SION", "SIAN" → X
			if dmContains(v, i, "SIO") || dmContains(v, i, "SIA") {
				if !isSlavoGermanic {
					dmAdd(&pBuf, &sBuf, &pLen, &sLen, "S", "X")
				} else {
					dmAdd(&pBuf, &sBuf, &pLen, &sLen, "S", "")
				}
				i += 3
				continue
			}
			// Germanic "SM", "SN", "SL", "SW"
			if i == 0 && (dmContains(v, i+1, "M") || dmContains(v, i+1, "N") ||
				dmContains(v, i+1, "L") || dmContains(v, i+1, "W")) {
				dmAdd(&pBuf, &sBuf, &pLen, &sLen, "S", "X")
				i++
				continue
			}
			// "SCH"
			if dmContains(v, i, "SCH") {
				// Preceding vowel Germanic: "SCHER", "SCHEN" → X, SK
				if dmContains(v, i+3, "ER") || dmContains(v, i+3, "EN") {
					dmAdd(&pBuf, &sBuf, &pLen, &sLen, "X", "SK")
				} else if i == 0 && !dmIsVowel(at(3)) && at(3) != 'W' {
					// Initial SCH + consonant (not W) — Germanic names like Schmidt.
					// Primary is X (sh-sound); secondary is S (Germanic hard SCH).
					dmAdd(&pBuf, &sBuf, &pLen, &sLen, "X", "S")
				} else {
					dmAdd(&pBuf, &sBuf, &pLen, &sLen, "X", "")
				}
				i += 3
				continue
			}
			// "SC" + E/I/Y → S/SK. The SC branch always advances by 3 (not 2):
			// the post-SC char is consumed as part of the SC trigram regardless of
			// the I/E/Y test outcome. This is the Philips 2000 / oubiwann metaphone==0.6
			// reference behaviour — verified by the Slavic "Sczepanski" corpus entry
			// (produces SKPN, not SKSP).
			if dmContains(v, i, "SC") {
				if at(i+2) == 'I' || at(i+2) == 'E' || at(i+2) == 'Y' {
					dmAdd(&pBuf, &sBuf, &pLen, &sLen, "S", "")
					i += 3
					continue
				}
				dmAdd(&pBuf, &sBuf, &pLen, &sLen, "SK", "")
				i += 3
				continue
			}
			// Q9 simplification: the original Philips C++ port had an if/else
			// here that discriminated the French SAIS-end rule
			// (i == n-1 && (at(i-1) == 'A' || at(i-1) == 'I')), but both
			// branches produced the identical ("S", "") emission. The locked
			// spec (docs/requirements.md §7.21) confirms both paths collapse to
			// the same dmAdd call; behaviour preservation is verified by
			// TestDoubleMetaphone_PaperWorkedExamples (Sais case, Plan 05 Gap 6
			// gate). British English: deliberate behaviour-preserving cleanup.
			dmAdd(&pBuf, &sBuf, &pLen, &sLen, "S", "")
			if at(i+1) == 'S' || at(i+1) == 'Z' {
				i++
			}
			i++

		case 'T':
			if dmContains(v, i, "TION") {
				dmAdd(&pBuf, &sBuf, &pLen, &sLen, "X", "")
				i += 3
				continue
			}
			if dmContains(v, i, "TIA") || dmContains(v, i, "TCH") {
				dmAdd(&pBuf, &sBuf, &pLen, &sLen, "X", "")
				i += 3
				continue
			}
			if dmContains(v, i, "TH") || dmContains(v, i, "TTH") {
				// Germanic / Romance TH → T (both keys T)
				// Condition: VAN or VON prefix (Dutch/Germanic), or SCH prefix,
				// or TH followed by OM/AM (Thomas, Thame — English Germanic).
				// Note: isSlavoGermanic is NOT used for the TH rule in Philips 2000;
				// Greek names like Katherine carry K (SlavoGermanic flag) but still
				// produce theta "0" for the TH sound.
				if dmContains(v, i+2, "OM") || dmContains(v, i+2, "AM") ||
					dmContains(v, 0, "VAN ") || dmContains(v, 0, "VON ") ||
					dmContains(v, 0, "SCH") {
					dmAdd(&pBuf, &sBuf, &pLen, &sLen, "T", "")
				} else {
					// Default TH → theta "0" (primary), T (secondary)
					dmAdd(&pBuf, &sBuf, &pLen, &sLen, "0", "T")
				}
				if dmContains(v, i, "TTH") {
					i += 3
				} else {
					i += 2
				}
				continue
			}
			if at(i+1) == 'T' || at(i+1) == 'D' {
				i++
			}
			dmAdd(&pBuf, &sBuf, &pLen, &sLen, "T", "")
			i++

		case 'V':
			if at(i+1) == 'V' {
				i++
			}
			dmAdd(&pBuf, &sBuf, &pLen, &sLen, "F", "")
			i++

		case 'W':
			// Initial "WR" → R
			if i == 0 && dmContains(v, i, "WR") {
				dmAdd(&pBuf, &sBuf, &pLen, &sLen, "R", "")
				i += 2
				continue
			}
			if i == 0 && (dmIsVowel(at(i+1)) || dmContains(v, i, "WH")) {
				// Initial W + vowel → two sounds A, F
				if dmIsVowel(at(i + 1)) {
					dmAdd(&pBuf, &sBuf, &pLen, &sLen, "A", "F")
				} else {
					dmAdd(&pBuf, &sBuf, &pLen, &sLen, "A", "")
				}
			}
			// Slavic: "WICZ", "WITZ" → TS or FX
			if dmContains(v, i, "WICZ") || dmContains(v, i, "WITZ") {
				dmAdd(&pBuf, &sBuf, &pLen, &sLen, "TS", "FX")
				i += 4
				continue
			}
			// Else silent W
			i++

		case 'X':
			if !(i == n-1 && (dmContains(v, i-3, "IAU") || dmContains(v, i-3, "EAU") ||
				dmContains(v, i-2, "AU") || dmContains(v, i-2, "OU"))) {
				dmAdd(&pBuf, &sBuf, &pLen, &sLen, "KS", "")
			}
			if at(i+1) == 'C' || at(i+1) == 'X' {
				i++
			}
			i++

		case 'Z':
			// Greek initial "ZA", "ZI", "ZO" or double "ZZ"
			if at(i+1) == 'H' {
				// ZH → J sound
				dmAdd(&pBuf, &sBuf, &pLen, &sLen, "J", "")
				i += 2
				continue
			}
			if dmContains(v, i+1, "ZO") || dmContains(v, i+1, "ZI") || dmContains(v, i+1, "ZA") ||
				(isSlavoGermanic && i > 0 && at(i-1) != 'T') {
				dmAdd(&pBuf, &sBuf, &pLen, &sLen, "S", "TS")
			} else {
				dmAdd(&pBuf, &sBuf, &pLen, &sLen, "S", "")
			}
			if at(i+1) == 'Z' {
				i++
			}
			i++

		default:
			// Vowels A, E, I, O, U: only code the initial vowel (already handled).
			// Subsequent vowels are skipped in Double Metaphone.
			i++
		}
	}

	// Convert the stack-buffer slices to strings. Each `string(buf[:n])`
	// is the single heap allocation per non-empty key (Q7a, §14.1).
	// The dmAdd write loops already cap pLen / sLen at dmMaxLen, so no
	// additional bounds clamp is required here.
	return string(pBuf[:pLen]), string(sBuf[:sLen])
}

// DoubleMetaphoneScore returns 1.0 if a and b share at least one non-empty
// phonetic key (primary or secondary) under the Double Metaphone algorithm
// (Philips 2000), and 0.0 otherwise.
//
// The four-way key match rule (per docs/requirements.md §7.4.2):
//
//	1.0 if any of {primary_a == primary_b, primary_a == secondary_b,
//	               secondary_a == primary_b, secondary_a == secondary_b}
//	where the matched key is non-empty.
//
// Empty strings: both-empty returns 1.0 (via identity short-circuit);
// one-empty returns 0.0 (one side always has empty keys).
//
// Non-ASCII input handling: this algorithm operates on ASCII letters
// [A-Za-z] only. Non-ASCII runes (accented characters, emoji,
// combining marks) are dropped silently before encoding. For
// Unicode-aware similarity on non-ASCII input, compose with
// Normalise + diacritic stripping before calling this function, or
// use a character-based algorithm (e.g. Levenshtein, Jaro-Winkler).
func DoubleMetaphoneScore(a, b string) float64 {
	// Identity short-circuit: covers both-empty (1.0) and same-string (1.0).
	if a == b {
		return 1.0
	}
	pa, sa := DoubleMetaphoneKeys(a)
	pb, sb := DoubleMetaphoneKeys(b)
	// Four-way key match — each matched key must be non-empty.
	// pp: a's primary matches b's primary
	if pa != "" && pa == pb {
		return 1.0
	}
	// ps: a's primary matches b's secondary
	if pa != "" && sb != "" && pa == sb {
		return 1.0
	}
	// sp: a's secondary matches b's primary
	if sa != "" && pb != "" && sa == pb {
		return 1.0
	}
	// ss: a's secondary matches b's secondary
	if sa != "" && sb != "" && sa == sb {
		return 1.0
	}
	return 0.0
}
