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

// nysiis.go implements the NYSIIS (New York State Identification and
// Intelligence System) phonetic algorithm for the fuzzymatch catalogue.
//
// Source-Origin Statement:
//
//	Algorithmic origin:  Taft, R. L. (1970). Name search techniques.
//	                     New York State Identification and Intelligence
//	                     System, Special Report No. 1. Albany, NY.
//	Canonical algorithm description (primary source for fresh transcription):
//	                     Knuth, D. E. (1973). The Art of Computer Programming,
//	                     Vol. 3, §6.4. Addison-Wesley.
//	Note: Taft 1970 is a NY State Special Report not available through
//	academic publishers; Knuth's secondary description in TAOCP Vol. 3
//	§6.4 is the authoritative algorithm description used for this
//	implementation.
//	Code consulted for reference vectors only:  jellyfish==1.2.1 (BSD-2-Clause)
//
// Variant chosen: original NYSIIS-1970, 6-character truncation, each rule
// applied once. Modified-NYSIIS-1991 / "wonderland" / iterate-to-fixed-point
// variants REJECTED per CONTEXT.md §2 LOCKED.
//
// MIT-licensed Go ports NOT consulted: github.com/UjjwalAyyangar/go-jellyfish
//
// Algorithm (Taft 1970 via Knuth TAOCP Vol. 3 §6.4 description):
//
// The 9-step NYSIIS procedure applied once (no iteration):
//
//  1. Prefix transliterations (applied to first characters):
//     MAC → MCC, KN → NN, K (initial) → C, PH → FF, PF → FF,
//     initial vowels (A,E,I,O,U) → A (collapse to single A)
//
//  2. Suffix transliterations (applied to last characters):
//     EE, IE → Y; DT, RT, RD, NT, ND → D
//
//  3. First character of result is kept from the prefix-translated name.
//
//  4. Body (iterative) rules applied left-to-right after the first character:
//     EV → AF; vowels (A,E,I,O,U) → A; Q → G; Z → S; M → N;
//     KN → N (if K at pos i and N at pos i+1); K → C;
//     SCH → S (if S at pos i and CH at pos i+1,i+2);
//     PH → FF; H (if surrounded by non-vowel/non-H/W on both sides: remove);
//     W (if preceded by vowel: remove)
//
//  5. Remove trailing S (unless the only character).
//
//  6. If last two characters are AY → replace with Y.
//
//  7. Remove trailing A (unless the only character).
//
//  8. Collapse consecutive identical characters (dedup).
//
//  9. Truncate to 6 characters (LOAD-BEARING per CONTEXT.md §2 LOCKED).
//
// Variant gate (LOAD-BEARING):
//
//	NYSIISCode("Catherine") == "CATARA"   (6 chars — NOT "CATARAN" which is
//	the jellyfish/modified-NYSIIS output)
//	NYSIISCode("Brown") == NYSIISCode("Browne") == "BRAN"
//	NYSIISCode("Robert") == "RABAD"
//
// Non-ASCII input handling: this algorithm operates on ASCII letters
// [A-Za-z] only. Non-ASCII runes (accented characters, emoji,
// combining marks) are dropped silently before encoding. For
// Unicode-aware similarity on non-ASCII input, compose with
// Normalise + diacritic stripping before calling this function, or
// use a character-based algorithm (e.g. Levenshtein, Jaro-Winkler).
//
// Note: NYSIIS is tuned for English-language names. Non-ASCII and
// non-English names produce codes with limited phonetic fidelity.
//
// Public surface (two functions):
//
//   - NYSIISCode(s string) string    — ≤ 6 uppercase ASCII letters
//   - NYSIISScore(a, b string) float64 — binary 0.0 / 1.0
//
// Only NYSIISScore is registered in the dispatch table (slot 25 —
// AlgoNYSIIS — see algoid.go and dispatch_nysiis.go).

package fuzzymatch

// nysiisIsVowel reports whether b (uppercase ASCII) is a vowel (A,E,I,O,U).
// Y is not treated as a vowel in NYSIIS.
func nysiisIsVowel(b byte) bool {
	return b == 'A' || b == 'E' || b == 'I' || b == 'O' || b == 'U'
}

// NYSIISCode returns the NYSIIS phonetic code for s, truncated to a maximum
// of 6 uppercase ASCII letters using the original Taft-1970 algorithm per
// Knuth TAOCP Vol. 3 §6.4.
//
// Variant choice (LOAD-BEARING per CONTEXT.md §2 LOCKED):
//   - 6-character truncation applied at end (original Taft-1970).
//   - Each rule applied once (no iterate-to-fixed-point).
//   - NYSIISCode("Catherine") == "CATARA" (NOT "CATARAN" which jellyfish returns).
//
// Non-ASCII input handling: this algorithm operates on ASCII letters
// [A-Za-z] only. Non-ASCII runes (accented characters, emoji,
// combining marks) are dropped silently before encoding. For
// Unicode-aware similarity on non-ASCII input, compose with
// Normalise + diacritic stripping before calling this function, or
// use a character-based algorithm (e.g. Levenshtein, Jaro-Winkler).
func NYSIISCode(s string) string {
	if len(s) == 0 {
		return ""
	}

	// Step 0: Decode input to uppercase ASCII letters into a stack-allocated
	// working buffer. Non-ASCII runes are silently skipped per CONTEXT.md §5.
	// name[0..nLen-1] holds the uppercase ASCII letters.
	// Maximum supported input: 128 ASCII letters (longer inputs are truncated
	// at the buffer boundary — adequate for any reasonable name input).
	var nameBuf [128]byte
	nLen := 0
	for i := 0; i < len(s); {
		b := s[i]
		if b >= 0x80 {
			// Non-ASCII: skip full UTF-8 sequence.
			_, sz := runeAt(s, i)
			i += sz
			continue
		}
		i++
		if b >= 'a' && b <= 'z' {
			b -= 32
		}
		if b < 'A' || b > 'Z' {
			continue // non-letter ASCII dropped
		}
		if nLen < 128 {
			nameBuf[nLen] = b
			nLen++
		}
	}
	if nLen == 0 {
		return ""
	}
	name := nameBuf[:nLen]

	// Step 1: Prefix transliterations.
	// Rules applied in priority order (longer patterns first).
	// We use a separate output buffer to avoid aliasing issues on insert/replace.
	var pfxBuf [130]byte
	pfxLen := 0
	switch {
	case nLen >= 3 && name[0] == 'M' && name[1] == 'A' && name[2] == 'C':
		// MAC → MCC
		pfxBuf[0], pfxBuf[1], pfxBuf[2] = 'M', 'C', 'C'
		pfxLen = 3 + copy(pfxBuf[3:], name[3:])
	case nLen >= 2 && name[0] == 'K' && name[1] == 'N':
		// KN → NN
		pfxBuf[0], pfxBuf[1] = 'N', 'N'
		pfxLen = 2 + copy(pfxBuf[2:], name[2:])
	case nLen >= 2 && name[0] == 'P' && name[1] == 'H':
		// initial PH → FF
		pfxBuf[0], pfxBuf[1] = 'F', 'F'
		pfxLen = 2 + copy(pfxBuf[2:], name[2:])
	case nLen >= 2 && name[0] == 'P' && name[1] == 'F':
		// initial PF → FF
		pfxBuf[0], pfxBuf[1] = 'F', 'F'
		pfxLen = 2 + copy(pfxBuf[2:], name[2:])
	default:
		pfxLen = copy(pfxBuf[:], name)
		if pfxBuf[0] == 'K' {
			pfxBuf[0] = 'C'
		} else if nysiisIsVowel(pfxBuf[0]) {
			pfxBuf[0] = 'A'
		}
	}
	work := pfxBuf[:pfxLen]

	// Step 2: Suffix transliterations.
	n := pfxLen
	switch {
	case n >= 2 && work[n-2] == 'E' && work[n-1] == 'E':
		// EE → Y
		work = work[:n-2]
		work = append(work, 'Y')
		n = len(work)
	case n >= 2 && work[n-2] == 'I' && work[n-1] == 'E':
		// IE → Y
		work = work[:n-2]
		work = append(work, 'Y')
		n = len(work)
	case n >= 2 && work[n-2] == 'D' && work[n-1] == 'T':
		// DT → D (remove trailing T)
		work = work[:n-1]
		n--
	case n >= 2 && work[n-2] == 'R' && work[n-1] == 'T':
		// RT → D
		work[n-2] = 'D'
		work = work[:n-1]
		n--
	case n >= 2 && work[n-2] == 'R' && work[n-1] == 'D':
		// RD → D (replace R with D, drop trailing D-slot)
		work[n-2] = 'D'
		work = work[:n-1]
		n--
	case n >= 2 && work[n-2] == 'N' && work[n-1] == 'T':
		// NT → D
		work[n-2] = 'D'
		work = work[:n-1]
		n--
	case n >= 2 && work[n-2] == 'N' && work[n-1] == 'D':
		// ND → D (replace N with D, drop trailing D-slot)
		work[n-2] = 'D'
		work = work[:n-1]
		n--
	}

	if n == 0 {
		return ""
	}

	// Step 4: Body transliterations applied left-to-right after the first char.
	// We build the result into a stack-allocated result buffer.
	var resBuf [128]byte
	res := resBuf[:0]
	res = append(res, work[0]) // first char unchanged

	body := work[1:]
	i := 0
	for i < len(body) {
		c := body[i]

		// Multi-char patterns: check longest first (SCH before shorter patterns).
		if i+2 < len(body) && c == 'S' && body[i+1] == 'C' && body[i+2] == 'H' {
			// SCH → S
			res = append(res, 'S')
			i += 3
			continue
		}
		if i+1 < len(body) {
			switch {
			case c == 'E' && body[i+1] == 'V':
				// EV → AF
				res = append(res, 'A', 'F')
				i += 2
				continue
			case c == 'K' && body[i+1] == 'N':
				// KN → N
				res = append(res, 'N')
				i += 2
				continue
			case c == 'P' && body[i+1] == 'H':
				// PH → FF
				res = append(res, 'F', 'F')
				i += 2
				continue
			}
		}

		// Single-char rules.
		switch c {
		case 'A', 'E', 'I', 'O', 'U':
			// Vowel → A
			res = append(res, 'A')
		case 'Q':
			res = append(res, 'G')
		case 'Z':
			res = append(res, 'S')
		case 'M':
			res = append(res, 'N')
		case 'K':
			res = append(res, 'C')
		case 'H':
			// H is removed from the body (NYSIIS body rule — H does not map
			// to any phonetic code in the original Taft-1970 algorithm).
			// Produces: John → JAN (O→A, H removed, N→N);
			//           Catherine → CATARA (T, H removed, E→A, R, I→A, N, E→A → truncated).
			// Do nothing: H is silently dropped.
		case 'W':
			// W is removed if preceded by a vowel (after translation to A).
			var prev byte
			if len(res) > 0 {
				prev = res[len(res)-1]
			}
			if nysiisIsVowel(prev) {
				// Remove W (skip it).
			} else {
				res = append(res, 'W')
			}
		default:
			res = append(res, c)
		}
		i++
	}

	// Step 5: Remove trailing S (unless it is the only character).
	if len(res) > 1 && res[len(res)-1] == 'S' {
		res = res[:len(res)-1]
	}

	// Step 6: If last two chars are AY → replace with Y.
	if len(res) >= 2 && res[len(res)-2] == 'A' && res[len(res)-1] == 'Y' {
		res[len(res)-2] = 'Y'
		res = res[:len(res)-1]
	}

	// Step 7: Remove trailing A (unless it is the only character).
	if len(res) > 1 && res[len(res)-1] == 'A' {
		res = res[:len(res)-1]
	}

	// Step 8: Collapse consecutive identical characters (dedup) in-place.
	// dst tracks the write position; src walks forward.
	dst := 1
	for src := 1; src < len(res); src++ {
		if res[src] != res[dst-1] {
			res[dst] = res[src]
			dst++
		}
	}
	res = res[:dst]

	// Step 9: Truncate to 6 characters (LOAD-BEARING Taft-1970 truncation).
	if len(res) > 6 {
		res = res[:6]
	}

	return string(res)
}

// NYSIISScore returns 1.0 if a and b share the same non-empty NYSIIS code
// (original Taft-1970 variant, 6-char truncation), and 0.0 otherwise.
//
// Both-empty convention: both-empty returns 1.0 (covered by the identity
// short-circuit per algorithm-correctness-standards).
// One-empty: returns 0.0 (empty code → no phonetic match).
//
// Non-ASCII input handling: this algorithm operates on ASCII letters
// [A-Za-z] only. Non-ASCII runes (accented characters, emoji,
// combining marks) are dropped silently before encoding. For
// Unicode-aware similarity on non-ASCII input, compose with
// Normalise + diacritic stripping before calling this function, or
// use a character-based algorithm (e.g. Levenshtein, Jaro-Winkler).
func NYSIISScore(a, b string) float64 {
	// Identity short-circuit covers both-empty (1.0) and same-string (1.0).
	if a == b {
		return 1.0
	}
	ca := NYSIISCode(a)
	cb := NYSIISCode(b)
	// Both codes must be non-empty and equal for a phonetic match.
	if ca != "" && ca == cb {
		return 1.0
	}
	return 0.0
}
