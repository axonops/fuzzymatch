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

// mra.go implements the MRA (Match Rating Approach) phonetic algorithm for the
// fuzzymatch catalogue.
//
// Source-Origin Statement:
//
//	Primary source:
//	  Moore, G. B., Kuhns, J. L., Trefftzs, J. L., Montgomery, C. A. (1977).
//	  Accessing individual records from personal data files using non-unique
//	  identifiers. National Bureau of Standards (now NIST), Technical Note 943.
//	  Available at: https://nvlpubs.nist.gov/nistpubs/Legacy/TN/nbstechnicalnote943.pdf
//	  (HTTP 200 verified 2026-05-15 per RESEARCH.md §1.4)
//	Cross-validation: jellyfish==1.2.1 (BSD-2-Clause) — consulted ONLY for
//	  reference-vector cross-validation via testdata/cross-validation/phonetic/vectors.json.
//	  NOT for code copying.
//	MIT-licensed Go ports NOT consulted: github.com/UjjwalAyyangar/go-jellyfish (MIT).
//	GPL/LGPL provenance: none.
//	Code copied verbatim: none.
//
// Per spec line 691 and CONTEXT.md §6 LOCKED: the (bool, int) return on
// MRACompare is faithful to NBS-943 steps 5 and 6, exposing the raw 0-6
// similarity counter for downstream consumers. MRACompare is the ONLY public
// function in the fuzzymatch catalogue with a non-float64 return shape.
//
// Algorithm (NBS Tech Note 943, encoding rules 1-3 + comparison steps 1-6):
//
// Encoding (MRACode):
//  1. Delete all vowels unless the vowel begins the word.
//  2. Remove the second consonant of any double consonants.
//  3. Reduce codex to 6 letters by joining first 3 + last 3 if len > 6.
//
// Comparison (MRACompare):
//  1. Reject if |len(codex_a) - len(codex_b)| >= 3 → return (false, 0).
//  2. Determine minimum threshold from Table A (see mraThresholdTable).
//  3. Process L→R: remove identical characters from both codexes.
//  4. Process R→L: remove identical characters from remaining.
//  5. similarity = 6 - max(unmatched_left, unmatched_right).
//  6. Match iff similarity >= threshold.
//
// Non-ASCII input handling: this algorithm operates on ASCII letters
// [A-Za-z] only. Non-ASCII runes (accented characters, emoji,
// combining marks) are dropped silently before encoding. For
// Unicode-aware similarity on non-ASCII input, compose with
// Normalise + diacritic stripping before calling this function, or
// use a character-based algorithm (e.g. Levenshtein, Jaro-Winkler).
//
// Note: MRA is tuned for English-language names. Non-ASCII and
// non-English names produce codes with limited phonetic fidelity.
//
// Public surface (three functions — per CONTEXT.md §6 LOCKED):
//
//   - MRACode(s string) string                             — ≤ 6 uppercase ASCII letters
//   - MRACompare(a, b string) (matched bool, simScore int) — raw 0-6 NBS similarity counter
//   - MRAScore(a, b string) float64                        — binary 0.0 / 1.0
//
// MRACompare is the ONLY function in the catalogue with a (bool, int) return.
// Only MRAScore is registered in the dispatch table (slot 26 — AlgoMRA —
// see algoid.go and dispatch_mra.go). MRACode and MRACompare are public but
// not dispatched (the dispatch table is (a, b string) float64 valued).

package fuzzymatch

// mraIsVowel reports whether b (uppercase ASCII) is a vowel (A, E, I, O, U).
// Y is NOT treated as a vowel in MRA encoding — only AEIOU are removed.
func mraIsVowel(b byte) bool {
	return b == 'A' || b == 'E' || b == 'I' || b == 'O' || b == 'U'
}

// mraThresholdTable maps sum-of-codex-lengths to minimum-rating per NBS Tech
// Note 943 Table A. Values for sum in [0, 12]; sum > 12 clamps to 2 per the
// table's monotonic-decrease semantics — see mraThreshold() (RESEARCH.md
// Pitfall 7.C — the sum>12 clamp is often omitted from Wikipedia-style
// summaries).
//
// Index is sum-of-lengths (0-12); value is minimum threshold:
//
//	sum ≤ 4:      threshold 5
//	4 < sum ≤ 7:  threshold 4
//	7 < sum ≤ 11: threshold 3
//	sum = 12:     threshold 2
//	sum > 12:     threshold 2 (clamped — same as sum=12)
var mraThresholdTable = [13]int{
	// sum:  0  1  2  3  4  5  6  7  8  9  10 11 12
	5, 5, 5, 5, 5, 4, 4, 4, 3, 3, 3, 3, 2,
}

// mraThreshold returns the minimum match-rating threshold for a given
// sum-of-codex-lengths per NBS Tech Note 943 Table A.
// For sum > 12: returns 2 (clamp — RESEARCH.md Pitfall 7.C).
func mraThreshold(sumLen int) int {
	if sumLen > 12 {
		// sum > 12 → threshold 2 (clamp; not always stated in Wikipedia summaries)
		return 2
	}
	return mraThresholdTable[sumLen]
}

// MRACode returns the MRA (Match Rating Approach) phonetic code for s per
// NBS Tech Note 943 encoding rules:
//
//  1. Delete all vowels unless the vowel begins the word.
//  2. Remove the second consonant of any double consonants.
//  3. Reduce codex to 6 letters by joining first 3 + last 3 if len > 6.
//
// Output is always ≤ 6 uppercase ASCII letters matching ^[A-Z]{0,6}$.
// Empty input returns "". Non-ASCII runes are dropped silently per CONTEXT.md §5.
//
// Reference vectors (NBS Tech Note 943 / jellyfish==1.2.1 cross-validation):
//
//	MRACode("Byrne")        = "BYRN"
//	MRACode("Smith")        = "SMTH"
//	MRACode("Smyth")        = "SMYTH"   (Y is not a vowel in MRA)
//	MRACode("Catherine")    = "CTHRN"
//	MRACode("Kathrynoglin") = "KTHGLN"  (first-3 + last-3 truncation gate)
//
// Non-ASCII input handling: this algorithm operates on ASCII letters
// [A-Za-z] only. Non-ASCII runes (accented characters, emoji,
// combining marks) are dropped silently before encoding. For
// Unicode-aware similarity on non-ASCII input, compose with
// Normalise + diacritic stripping before calling this function, or
// use a character-based algorithm (e.g. Levenshtein, Jaro-Winkler).
func MRACode(s string) string { //nolint:gocyclo // canonical NBS-943 three-step encoding (input filter + vowel removal + double-consonant dedup); each step contributes a small branch count, totalling 16
	if len(s) == 0 {
		return ""
	}

	// Step 0: Decode input to uppercase ASCII letters, silently skipping
	// non-ASCII runes per CONTEXT.md §5. Stack-allocated intermediate buffer.
	var rawBuf [64]byte
	rawLen := 0
	for i := 0; i < len(s); {
		b := s[i]
		if b >= 0x80 {
			// Non-ASCII: skip full UTF-8 sequence.
			sz := runeSizeAt(s, i)
			i += sz
			continue
		}
		i++
		if b >= 'a' && b <= 'z' {
			b -= 32 // uppercase
		}
		if b < 'A' || b > 'Z' {
			continue // non-letter ASCII dropped
		}
		if rawLen < 64 {
			rawBuf[rawLen] = b
			rawLen++
		}
	}
	if rawLen == 0 {
		return ""
	}

	// Step 1: Delete all vowels unless the vowel begins the word.
	// Keep rawBuf[0] unconditionally; for subsequent positions drop vowels.
	var step1Buf [64]byte
	step1Len := 0
	step1Buf[step1Len] = rawBuf[0]
	step1Len++
	for i := 1; i < rawLen; i++ {
		c := rawBuf[i]
		if !mraIsVowel(c) {
			step1Buf[step1Len] = c
			step1Len++
		}
	}

	// Step 2: Remove the second consonant of any double consonants.
	// Scan left-to-right; only add a character if it differs from the last
	// accepted character (deduplication of adjacent identical consonants).
	// We operate on the result of step 1, so vowels (except leading) are
	// already removed; all remaining characters (except possibly leading) are
	// consonants.
	var step2Buf [64]byte
	step2Len := 0
	if step1Len > 0 {
		step2Buf[0] = step1Buf[0]
		step2Len = 1
		for i := 1; i < step1Len; i++ {
			if step1Buf[i] != step2Buf[step2Len-1] {
				step2Buf[step2Len] = step1Buf[i]
				step2Len++
			}
		}
	}

	// Step 3: Reduce codex to 6 letters by joining first 3 + last 3 if len > 6.
	var result [6]byte
	if step2Len <= 6 {
		copy(result[:], step2Buf[:step2Len])
		return string(result[:step2Len])
	}
	// step2Len > 6 here (the <=6 branch above returned). Every index in
	// [step2Len-3, step2Len-1] is strictly positive (>= 4) and < step2Len
	// (which is <= 64 because step2Buf has length 64 and the writes to it
	// in steps 1-2 above are bounded by `if step2Len < 64` / `if step1Len
	// < 64` increments). gosec's G602 bounds-prover does not propagate
	// the `step2Len <= 6` early-return into the negation, so these three
	// indexed reads trigger false positives — silenced explicitly.
	// first 3 + last 3:
	result[0] = step2Buf[0]
	result[1] = step2Buf[1]
	result[2] = step2Buf[2]
	result[3] = step2Buf[step2Len-3] //nolint:gosec // G602 false positive: step2Len > 6 here (early-return above)
	result[4] = step2Buf[step2Len-2] //nolint:gosec // G602 false positive: step2Len > 6 here (early-return above)
	result[5] = step2Buf[step2Len-1] //nolint:gosec // G602 false positive: step2Len > 6 here (early-return above)
	return string(result[:])
}

// MRACompare performs the full MRA comparison per NBS Tech Note 943:
//
//  1. Reject if |len(codex_a) - len(codex_b)| >= 3 → return (false, 0).
//  2. Look up minimum threshold from mraThresholdTable (sum > 12 clamps to 2).
//  3. Process L→R: remove identical characters from both codexes.
//  4. Process R→L: remove identical characters from remaining.
//  5. similarity = 6 - max(unmatched_left, unmatched_right). Always in [0, 6].
//  6. Match iff similarity >= threshold.
//
// Returns (matched bool, simScore int) where simScore is the raw 0-6 NBS
// similarity counter. MRACompare is the ONLY public function in the fuzzymatch
// catalogue with a non-float64 return shape (per CONTEXT.md §6 LOCKED and
// spec line 691).
//
// Return-shape rationale (Phase 8.5 Gap 7):
//
// The (bool, int) tuple is INTENTIONAL and is part of the v1.x stability
// contract. NBS Tech Note 943 steps 5 and 6 produce two distinct pieces of
// information per comparison — the raw 0-to-6 similarity counter (step 5)
// AND the threshold-gated match decision (step 6) — and a downstream
// consumer doing record linkage needs both:
//
//   - The decision (matched bool) drives the binary accept/reject branch.
//   - The counter (simScore int) drives downstream ranking, scoring
//     ensembles, and confidence display (a similarity of 5 is meaningfully
//     stronger than a similarity of 3 even when both pass the threshold).
//
// Collapsing the return to a single value would force every consumer to
// reconstruct the counter via a second pass through the algorithm,
// defeating the work step 5 already does. Wrapping into a struct was
// considered and rejected: the result type would appear in exactly one
// place in the catalogue (this function) and a two-field tuple is
// idiomatic in Go.
//
// MRAScore wraps MRACompare to provide the dispatch-table-compatible
// float64 surface (1.0 on match, 0.0 otherwise). See docs/algorithms.md
// and docs/scorer.md for the consumer-facing documentation of this
// surface decision.
//
// Special cases:
//   - Both empty: MRACompare("", "") = (true, 6) per algorithm-correctness-standards.
//   - Identity short-circuit: if a == b, returns (true, 6) immediately.
//   - Length-difference gate: if |len(MRACode(a)) - len(MRACode(b))| >= 3,
//     returns (false, 0) per docs/requirements.md §7.4.4 line 696.
//     Unlike jellyfish which returns an error here, fuzzymatch returns (false, 0)
//     per CONTEXT.md §6 LOCKED (Open Question 2 resolution).
//
// Non-ASCII input handling: this algorithm operates on ASCII letters
// [A-Za-z] only. Non-ASCII runes are dropped silently before encoding.
func MRACompare(a, b string) (matched bool, simScore int) { //nolint:gocyclo // canonical NBS-943 six-step comparison (length gate + threshold lookup + L→R + R→L + unmatched count + threshold compare); the branches mirror the algorithm shape
	// Identity short-circuit covers both-empty (true, 6) and same-string (true, 6).
	if a == b {
		return true, 6
	}

	codexA := MRACode(a)
	codexB := MRACode(b)

	lenA := len(codexA)
	lenB := len(codexB)

	// One-empty guard (catalogue convention — algorithm-correctness-standards §"Edge cases"):
	// empty vs non-empty must never match. The NBS Tech Note 943 length-difference gate
	// fires only at diff >= 3; without this guard, (0,1) and (0,2) length pairs would
	// produce spurious matches because similarity = 6 - max(0, lenB) meets the threshold.
	if lenA == 0 || lenB == 0 {
		return false, 0
	}

	// Step 1: Length-difference gate.
	// |lenA - lenB| >= 3 → automatic mismatch per docs/requirements.md §7.4.4 line 696.
	diff := lenA - lenB
	if diff < 0 {
		diff = -diff
	}
	if diff >= 3 {
		return false, 0
	}

	// Step 2: Determine minimum threshold from Table A.
	sumLen := lenA + lenB
	threshold := mraThreshold(sumLen)

	// Steps 3 and 4: Common-character elimination.
	// Work on byte slices derived from the codex strings. Use small stack-allocated
	// boolean arrays to mark characters as matched (consumed) so we can process
	// L→R then R→L without modifying the underlying slices in place.
	var matchedA [6]bool
	var matchedB [6]bool

	// Step 3: L→R pass — scan from left; for each unmatched char in A, find the
	// first unmatched occurrence in B at the same or later position.
	// NBS Tech Note 943 step 3: "process L→R, remove identical chars from both".
	for i := 0; i < lenA; i++ {
		for j := 0; j < lenB; j++ {
			if !matchedA[i] && !matchedB[j] && codexA[i] == codexB[j] {
				matchedA[i] = true
				matchedB[j] = true
				break
			}
		}
	}

	// Step 4: R→L pass on remaining unmatched characters.
	for i := lenA - 1; i >= 0; i-- {
		if matchedA[i] {
			continue
		}
		for j := lenB - 1; j >= 0; j-- {
			if !matchedA[i] && !matchedB[j] && codexA[i] == codexB[j] {
				matchedA[i] = true
				matchedB[j] = true
				break
			}
		}
	}

	// Count unmatched characters in each codex.
	unmatchedA := 0
	for i := 0; i < lenA; i++ {
		if !matchedA[i] {
			unmatchedA++
		}
	}
	unmatchedB := 0
	for j := 0; j < lenB; j++ {
		if !matchedB[j] {
			unmatchedB++
		}
	}

	// Step 5: similarity = 6 - max(unmatched_a, unmatched_b). Always in [0, 6].
	maxUnmatched := unmatchedA
	if unmatchedB > maxUnmatched {
		maxUnmatched = unmatchedB
	}
	similarity := 6 - maxUnmatched
	if similarity < 0 {
		similarity = 0
	}

	// Step 6: Match iff similarity >= threshold.
	return similarity >= threshold, similarity
}

// MRAScore returns 1.0 if a and b are an MRA match (MRACompare returns matched=true),
// and 0.0 otherwise. This is the binary dispatch-table wrapper around MRACompare.
//
// For programmatic input-quality checks before scoring (including
// WarnAllNonASCIIDropped scoped to AlgoMRA), see [fuzzymatch.Validate].
//
// Strict consistency invariant: MRAScore(a, b) == 1.0 iff MRACompare(a, b).matched.
// (Property-tested by PropMRA_ScoreCompareConsistency.)
//
// Both-empty: MRAScore("", "") = 1.0 (identity short-circuit).
// Identity: MRAScore(s, s) = 1.0 for any s.
//
// Non-ASCII input handling: this algorithm operates on ASCII letters
// [A-Za-z] only. Non-ASCII runes are dropped silently before encoding.
func MRAScore(a, b string) float64 {
	// Identity short-circuit covers both-empty (1.0) and same-string (1.0).
	if a == b {
		return 1.0
	}
	matched, _ := MRACompare(a, b)
	if matched {
		return 1.0
	}
	return 0.0
}
