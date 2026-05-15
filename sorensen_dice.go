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

// sorensen_dice.go implements the Sørensen-Dice coefficient for the
// fuzzymatch catalogue. Sørensen-Dice is the textbook q-gram metric
// that weights the intersection more heavily than Jaccard
// (DSC = 2·|∩|/(|QA|+|QB|) vs Jaccard's |∩|/|∪|), making it the
// canonical default for many fuzzy-name-matching workloads.
//
// Source: Dice, L. R. (1945). "Measures of the amount of ecologic
// association between species." Ecology 26(3):297-302, §3 — the
// original formulation. Independently rediscovered by:
// Sørensen, T. (1948). "A method of establishing groups of equal
// amplitude in plant sociology based on similarity of species
// content." Kongelige Danske Videnskabernes Selskab 5(4):1-34, §3.
// The coefficient is symmetric and identical under both derivations;
// the modern community names it Sørensen-Dice to acknowledge both
// authors.
//
// Algorithm (multiset / weighted form, per docs/requirements.md §7.2.2):
//
//   For strings A, B and q-gram size n:
//
//     QA = multiset of overlapping length-n substrings of A
//     QB = multiset of overlapping length-n substrings of B
//
//     |QA ∩ QB| = Σ_{k ∈ keys(QA) ∩ keys(QB)} min(countA[k], countB[k])
//     |QA|      = Σ countA[k]            (sum of multiset counts)
//     |QB|      = Σ countB[k]
//
//     DSC(A, B) = 2·|QA ∩ QB| / (|QA| + |QB|)
//
// Conventions (mirror the Phase 2/3/4 short-circuit pattern; identical
// to the Q-Gram Jaccard conventions in plan 05-01 for consistency):
//
//   - both-empty   → 1.0 (covered by the a == b identity short-circuit)
//   - identical    → 1.0 (a == b short-circuit)
//   - one-empty    → 0.0
//
// Direct-call validation (CONTEXT.md §5 LOCKED):
//
//   - n < 1 panics with the message "fuzzymatch: invalid q-gram size".
//     The Phase 8 Scorer option WithSorensenDiceAlgorithm returns
//     ErrInvalidQGramSize instead — the panic is reserved for the
//     direct-call surface where programmer error must fail loudly.
//
// Canonical NLP-textbook reference vector (RV-D1):
//
//   SorensenDiceScore("night", "nacht", 2) =
//
//     QA = bigrams("night") = {ni:1, ig:1, gh:1, ht:1} — total 4
//     QB = bigrams("nacht") = {na:1, ac:1, ch:1, ht:1} — total 4
//     |QA ∩ QB| = 1 (only "ht" shared)
//     DSC = 2·1 / (4 + 4) = 2/8 = 0.25
//
// Source-origin discipline (per algorithm-licensing-standards):
//
//   - Primary source:        Dice 1945 §3 (Ecology 26(3):297-302) +
//                            Sørensen 1948 §3 (Kgl. Danske Videnskab.
//                            Selskab 5(4):1-34) — independent
//                            rediscoveries of the same coefficient.
//   - Cross-validation:      none — hand-derived RV-D1..RV-D4
//                            reference vectors in sorensen_dice_test.go,
//                            each with the formula derivation embedded
//                            in the test comment. Per CONTEXT.md §4
//                            LOCKED, no Python toolchain is used.
//   - Tie-break:             none (Sørensen-Dice is unambiguous; the
//                            multiset cardinality is associative
//                            integer arithmetic).
//   - GPL/LGPL provenance:   none.
//   - Code copied verbatim:  none.
//
// Implementation discipline:
//
//   - NO init()-time table builds (per docs/requirements.md §5(12)).
//   - NO map iteration on output paths (DET-03). The intersection
//     cardinality is computed by iterating the smaller of the two
//     multisets and summing min(countA[k], countB[k]) into an INTEGER
//     counter — the OUTPUT is the integer count, NOT an ordered
//     slice, so map-iteration order does not affect the result.
//   - NO transcendental float operations (DET-06): only integer
//     arithmetic, float64() casts, and a single multiplication +
//     addition + division. No math.Pow / math.Log / math.Exp /
//     math.FMA / math.Sqrt.
//   - DSC formula uses explicit left-to-right parenthesisation
//     `(2.0 * float64(intersection)) / (float64(lenA) + float64(lenB))`
//     per DET-06 — no associativity ambiguity for cross-platform
//     determinism.
//   - Identity short-circuit a == b BEFORE any extraction work — both
//     covers the both-empty case AND avoids two map allocations on
//     identical inputs.
//   - The rune-path variant calls extractQGramsRunes which performs
//     the []rune conversion once per side; the byte-path variant
//     performs zero rune-slice allocations.
//
// Public surface (two functions; no rune-Distance variant —
// Sørensen-Dice does not expose a distance, only a similarity score):
//
//   - SorensenDiceScore(a, b string, n int) float64
//   - SorensenDiceScoreRunes(a, b string, n int) float64
//
// Only SorensenDiceScore is registered in the dispatch table (slot
// AlgoSorensenDice — see algoid.go) with a default n = 3 wrapper —
// the dispatch table maps AlgoID to (a, b string) float64 and has no
// place for the n parameter. Specific n overrides happen via the
// Phase 8 Scorer option WithSorensenDiceAlgorithm(weight, n).
//
// Worst-case complexity: O(la + lb) time + O(la + lb) space for the
// two multiset maps. Pure-function library — caller controls input
// size; the algorithm has no input-validation rejection on long input.

package fuzzymatch

// SorensenDiceScore returns the Sørensen-Dice coefficient of the
// q-gram multisets of a and b: DSC(A, B) = 2·|QA ∩ QB| / (|QA| + |QB|),
// in [0.0, 1.0]. Operates on bytes — multi-byte UTF-8 inputs split
// q-grams at byte boundaries, which can produce different results
// than SorensenDiceScoreRunes on non-ASCII input.
//
// The q-gram size n MUST be >= 1; n < 1 panics with the message
// "fuzzymatch: invalid q-gram size" (CONTEXT.md §5 LOCKED — direct
// calls fail loudly on programmer error; the Phase 8 Scorer option
// WithSorensenDiceAlgorithm returns ErrInvalidQGramSize instead).
//
// Conventions:
//
//   - SorensenDiceScore("",   "",   n) == 1.0  (both-empty identity)
//   - SorensenDiceScore("",   "abc", n) == 0.0 (one-empty)
//   - SorensenDiceScore("abc", "",   n) == 0.0 (one-empty)
//   - SorensenDiceScore("hello", "hello", n) == 1.0 (identity)
//
// When both inputs are non-empty AND each is shorter than n, the
// q-gram extraction returns empty multisets on both sides; the
// both-empty guard does not fire (the input strings differ even
// though their q-gram views are empty), so the function falls through
// and the both-extractions-empty guard inside diceFromQGramMaps
// returns 1.0 by convention (a vacuous match — both q-gram views are
// empty).
//
// Reference vector (RV-D1 — load-bearing canonical NLP-textbook
// bigram pair):
//
//	SorensenDiceScore("night", "nacht", 2) = 0.25
//
//	QA = {ni:1, ig:1, gh:1, ht:1}; QB = {na:1, ac:1, ch:1, ht:1};
//	|∩| = 1; DSC = 2·1/(4+4) = 0.25.
func SorensenDiceScore(a, b string, n int) float64 {
	if n < 1 {
		panic("fuzzymatch: invalid q-gram size")
	}
	if a == b {
		return 1.0 // identity short-circuit (covers both-empty too)
	}
	if a == "" || b == "" {
		return 0.0
	}
	qa := extractQGrams(a, n)
	qb := extractQGrams(b, n)
	return diceFromQGramMaps(qa, qb)
}

// SorensenDiceScoreRunes returns the Sørensen-Dice coefficient of the
// q-gram multisets of a and b under rune semantics. Multi-byte UTF-8
// sequences are compared atomically: extractQGramsRunes("café", 2)
// yields three rune-bigrams {"ca", "af", "fé"} where "fé" is the
// multi-byte UTF-8 encoding of the two-rune window.
//
// Allocates two []rune slices on the heap (one per side) plus the
// two multiset maps. For pure-ASCII inputs prefer SorensenDiceScore.
//
// Direct-call validation, edge cases, and conventions are identical
// to SorensenDiceScore (see godoc above). The rune-path divergence
// from the byte path is that windows are over runes, not bytes —
// "café" has 4 runes, so SorensenDiceScoreRunes("café", x, 2) is
// well-defined for any x, while the byte path treats "café" as 5
// bytes.
//
// Reference vector:
//
//	SorensenDiceScoreRunes("café", "cafe", 2) = 4/6 ≈ 0.6666666666666666
//
//	QA = rune-bigrams("café") = {"ca":1, "af":1, "fé":1}
//	QB = rune-bigrams("cafe") = {"ca":1, "af":1, "fe":1}
//	|∩| = 2 (ca + af); DSC = 2·2/(3+3) = 4/6 ≈ 0.6666...
func SorensenDiceScoreRunes(a, b string, n int) float64 {
	if n < 1 {
		panic("fuzzymatch: invalid q-gram size")
	}
	if a == b {
		return 1.0 // identity short-circuit (covers both-empty too)
	}
	if a == "" || b == "" {
		return 0.0
	}
	qa := extractQGramsRunes(a, n)
	qb := extractQGramsRunes(b, n)
	return diceFromQGramMaps(qa, qb)
}

// diceFromQGramMaps computes DSC(A, B) = 2·|QA ∩ QB| / (|QA| + |QB|)
// from two pre-extracted q-gram multisets. The multiset cardinalities
// are:
//
//	lenA         = Σ countA[k]            (sum of map values for QA)
//	lenB         = Σ countB[k]
//	intersection = Σ_{k ∈ keys(QA) ∩ keys(QB)} min(countA[k], countB[k])
//
// Helper invariant: the OUTPUT is a single float64 derived from
// integer cardinalities. The internal map iterations to compute
// `lenA` / `lenB` and `intersection` produce INTEGER counters whose
// values do not depend on iteration order (integer addition is
// associative); no ordered output is constructed. DET-03 satisfied.
//
// When both qa and qb are empty, returns 1.0 by the both-empty
// convention (a vacuous match — no q-grams to disagree about).
//
// The DSC formula uses explicit left-to-right parenthesisation
// `(2.0 * float64(intersection)) / (float64(lenA) + float64(lenB))`
// per DET-06 — no associativity ambiguity across CI platforms.
func diceFromQGramMaps(qa, qb map[string]int) float64 {
	if len(qa) == 0 && len(qb) == 0 {
		return 1.0
	}
	// Sum the per-side multiset cardinalities. Map iteration order is
	// randomised but integer addition is associative, so the SUM is
	// deterministic regardless of order. DET-03 satisfied: the output
	// is a scalar int, not an ordered slice.
	var lenA, lenB int
	for _, c := range qa {
		lenA += c
	}
	for _, c := range qb {
		lenB += c
	}
	// Intersection: walk the SMALLER multiset and sum
	// min(countA[k], countB[k]) for keys present in both. Walking the
	// smaller side keeps the lookup count to len(min(qa, qb)).
	small, large := qa, qb
	if len(qb) < len(qa) {
		small, large = qb, qa
	}
	var intersection int
	for k, cs := range small {
		if cl, ok := large[k]; ok {
			// min(cs, cl) — Go 1.26 stdlib has builtin min, but the
			// project canonical pattern is the explicit if/else for
			// determinism-reviewer auditability.
			if cs < cl {
				intersection += cs
			} else {
				intersection += cl
			}
		}
	}
	if lenA+lenB == 0 {
		// Defensive: both totals are zero. The len(qa)==0 && len(qb)==0
		// guard above already covers this, but keep the explicit
		// fall-through to avoid a 0/0 NaN if invariants change.
		return 1.0
	}
	// Single multiplication + addition + division on integer-derived
	// float64 values. Both numerator and denominator fit exactly in
	// float64 for any input where len(a)+len(b) < 2^53 (~9e15) — well
	// above any realistic input. IEEE-754 correctly-rounded operations
	// produce byte-identical output across all four CI platforms
	// (DET-06 satisfied). Explicit left-to-right parenthesisation:
	// `(2.0 * float64(intersection)) / (float64(lenA) + float64(lenB))`.
	return (2.0 * float64(intersection)) / (float64(lenA) + float64(lenB))
}
