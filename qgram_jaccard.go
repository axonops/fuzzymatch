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

// qgram_jaccard.go implements the Q-Gram Jaccard similarity for the
// fuzzymatch catalogue. Q-Gram Jaccard is the textbook q-gram metric
// and the simplest consumer of the shared q-gram extraction
// infrastructure in q_gram.go (Ukkonen 1992 §3).
//
// Source: Ukkonen, E. (1992). "Approximate string-matching with q-grams
// and maximal matches." Theoretical Computer Science 92(1):191-211, §3
// — the multiset Jaccard formulation over q-gram counts. Underlying
// set-coefficient origin: Jaccard, P. (1912). "The distribution of the
// flora in the alpine zone." New Phytologist 11(2):37-50, p. 43.
//
// Algorithm (multiset / weighted form, per docs/requirements.md §7.2.1):
//
//   For strings A, B and q-gram size n:
//
//     QA = multiset of overlapping length-n substrings of A
//     QB = multiset of overlapping length-n substrings of B
//
//     |QA ∩ QB| = Σ_{k ∈ keys(QA) ∩ keys(QB)} min(countA[k], countB[k])
//     |QA ∪ QB| = Σ_{k ∈ keys(QA) ∪ keys(QB)} max(countA[k], countB[k])
//                = Σ countA[k] + Σ countB[k] - |QA ∩ QB|
//                = totalA + totalB - intersection
//
//     J(A, B) = |QA ∩ QB| / |QA ∪ QB|
//
// Conventions (mirror the Phase 2/3/4 short-circuit pattern):
//
//   - both-empty   → 1.0 (covered by the a == b identity short-circuit)
//   - identical    → 1.0 (a == b short-circuit)
//   - one-empty    → 0.0
//
// Direct-call validation (CONTEXT.md §5 LOCKED):
//
//   - n < 1 panics with the message "fuzzymatch: invalid q-gram size".
//     The Phase 8 Scorer option WithQGramJaccardAlgorithm returns
//     ErrInvalidQGramSize instead — the panic is reserved for the
//     direct-call surface where programmer error must fail loudly.
//
// Ukkonen 1992 §3 worked-example reference vector (RV-J1):
//
//   QGramJaccardScore("AGCT", "AGCTAGCT", 2) =
//
//     QA = bigrams("AGCT")     = {AG:1, GC:1, CT:1}        — total 3
//     QB = bigrams("AGCTAGCT") = {AG:2, GC:2, CT:2, TA:1}  — total 7
//     |QA ∩ QB| = min(1,2)+min(1,2)+min(1,2)+min(0,1) = 3
//     |QA ∪ QB| = 3 + 7 - 3 = 7
//     J = 3/7 = 0.42857142857142855
//
// Source-origin discipline (per algorithm-licensing-standards):
//
//   - Primary source:        Ukkonen 1992 §3 (multiset Jaccard
//                            formulation); Jaccard 1912 p. 43 (set
//                            coefficient origin).
//   - Cross-validation:      none — hand-derived RV-J1..RV-J6
//                            reference vectors in qgram_jaccard_test.go,
//                            each with the formula derivation embedded
//                            in the test comment. Per CONTEXT.md §4
//                            LOCKED, no Python toolchain is used.
//   - Tie-break:             none (set-Jaccard is unambiguous; the
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
//     arithmetic, float64() casts, and a single division. No
//     math.Pow / math.Log / math.Exp / math.FMA / math.Sqrt.
//   - Identity short-circuit a == b BEFORE any extraction work — both
//     covers the both-empty case AND avoids two map allocations on
//     identical inputs.
//   - The rune-path variant calls extractQGramsRunes which performs
//     the []rune conversion once per side; the byte-path variant
//     performs zero rune-slice allocations.
//
// Public surface (two functions; no rune-Distance variant — Q-Gram
// Jaccard does not expose a distance, only a similarity score):
//
//   - QGramJaccardScore(a, b string, n int) float64
//   - QGramJaccardScoreRunes(a, b string, n int) float64
//
// Only QGramJaccardScore is registered in the dispatch table (slot 9
// — see algoid.go AlgoQGramJaccard) with a default n = 3 wrapper —
// the dispatch table maps AlgoID to (a, b string) float64 and has no
// place for the n parameter. Specific n overrides happen via the
// Phase 8 Scorer option WithQGramJaccardAlgorithm(weight, n).
//
// Worst-case complexity: O(la + lb) time + O(la + lb) space for the
// two multiset maps. Pure-function library — caller controls input
// size; the algorithm has no input-validation rejection on long input.

package fuzzymatch

// QGramJaccardScore returns the Jaccard similarity of the q-gram
// multisets of a and b: J(A, B) = |QA ∩ QB| / |QA ∪ QB|, in
// [0.0, 1.0]. Operates on bytes — multi-byte UTF-8 inputs split q-grams
// at byte boundaries, which can produce different results than
// QGramJaccardScoreRunes on non-ASCII input.
//
// The q-gram size n MUST be >= 1; n < 1 panics with the message
// "fuzzymatch: invalid q-gram size" (CONTEXT.md §5 LOCKED — direct
// calls fail loudly on programmer error; the Phase 8 Scorer option
// WithQGramJaccardAlgorithm returns ErrInvalidQGramSize instead).
//
// Conventions:
//
//   - QGramJaccardScore("",   "",   n) == 1.0  (both-empty identity)
//   - QGramJaccardScore("",   "abc", n) == 0.0 (one-empty)
//   - QGramJaccardScore("abc", "",   n) == 0.0 (one-empty)
//   - QGramJaccardScore("hello", "hello", n) == 1.0 (identity)
//
// When both inputs are non-empty AND each is shorter than n, the q-gram
// extraction returns empty multisets on both sides; the both-empty
// guard does not fire (the input strings differ even though their
// q-gram views are empty), so the function falls through and the
// computation produces 0/0 — handled by the explicit empty-multiset
// guard which returns 1.0 by convention (a vacuous match — both q-gram
// views are empty).
//
// Reference vector (Ukkonen 1992 §3 worked example):
//
//	QGramJaccardScore("AGCT", "AGCTAGCT", 2) = 3/7 ≈ 0.42857142857142855
func QGramJaccardScore(a, b string, n int) float64 {
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
	return jaccardFromQGramMaps(qa, qb)
}

// QGramJaccardScoreRunes returns the Jaccard similarity of the q-gram
// multisets of a and b under rune semantics. Multi-byte UTF-8 sequences
// are compared atomically: extractQGramsRunes("café", 2) yields three
// rune-bigrams {"ca", "af", "fé"} where "fé" is the multi-byte UTF-8
// encoding of the two-rune window.
//
// Allocates two []rune slices on the heap (one per side) plus the two
// multiset maps. For pure-ASCII inputs prefer QGramJaccardScore.
//
// Direct-call validation, edge cases, and conventions are identical
// to QGramJaccardScore (see godoc above). The rune-path divergence
// from the byte path is that windows are over runes, not bytes —
// "café" has 4 runes, so QGramJaccardScoreRunes("café", x, 2) is
// well-defined for any x, while the byte path treats "café" as 5 bytes.
//
// Reference vector:
//
//	QGramJaccardScoreRunes("café", "cafe", 2) = 0.5
//
//	QA = rune-bigrams("café") = {"ca":1, "af":1, "fé":1}
//	QB = rune-bigrams("cafe") = {"ca":1, "af":1, "fe":1}
//	intersection = 2 (ca + af); union = 4 (ca + af + fé + fe);
//	J = 2/4 = 0.5
func QGramJaccardScoreRunes(a, b string, n int) float64 {
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
	return jaccardFromQGramMaps(qa, qb)
}

// jaccardFromQGramMaps computes J(A, B) = |QA ∩ QB| / |QA ∪ QB| from
// two pre-extracted q-gram multisets. The multiset cardinalities are:
//
//	totalA       = Σ countA[k]            (sum of map values for QA)
//	totalB       = Σ countB[k]
//	intersection = Σ_{k ∈ keys(QA) ∩ keys(QB)} min(countA[k], countB[k])
//	union        = totalA + totalB - intersection
//
// Helper invariant: the OUTPUT is a single float64 derived from
// integer cardinalities. The internal map iterations to compute
// `total*` and `intersection` produce INTEGER counters whose values
// do not depend on iteration order (integer addition is associative);
// no ordered output is constructed. DET-03 satisfied.
//
// When both qa and qb are empty, returns 1.0 by the both-empty
// convention (a vacuous match — no q-grams to disagree about).
func jaccardFromQGramMaps(qa, qb map[string]int) float64 {
	if len(qa) == 0 && len(qb) == 0 {
		return 1.0
	}
	// Sum the per-side multiset cardinalities. Map iteration order is
	// randomised but integer addition is associative, so the SUM is
	// deterministic regardless of order. DET-03 satisfied: the output
	// is a scalar int, not an ordered slice.
	var totalA, totalB int
	for _, c := range qa {
		totalA += c
	}
	for _, c := range qb {
		totalB += c
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
	union := totalA + totalB - intersection
	if union == 0 {
		// Defensive: both totals are zero. The len(qa)==0 && len(qb)==0
		// guard above already covers this, but keep the explicit
		// fall-through to avoid a 0/0 NaN if invariants change.
		return 1.0
	}
	// Single division on integer-derived float64 values. Both
	// numerator and denominator fit exactly in float64 for any input
	// where len(a)+len(b) < 2^53 (~9e15) — well above any realistic
	// input. IEEE-754 correctly-rounded division produces byte-identical
	// output across all four CI platforms (DET-06 satisfied).
	return float64(intersection) / float64(union)
}
