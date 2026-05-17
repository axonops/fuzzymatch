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

// tversky.go implements the Tversky asymmetric similarity index for the
// fuzzymatch catalogue. Tversky is the parameterised q-gram metric where
// the weights α and β control how strongly each side's residual elements
// (|A−B| and |B−A|) penalise the score; with α ≠ β the function is
// direction-sensitive — swapping the inputs generally yields a different
// score. With α = β = 1 the function reduces to Q-Gram Jaccard; with
// α = β = 0.5 it reduces to Sørensen-Dice. These two equivalences are
// pinned bit-for-bit by cross-check unit tests AND by property tests
// (RV-T3 / RV-T4 + TestProp_TverskyScore_JaccardCrossCheck +
// TestProp_TverskyScore_DiceCrossCheck).
//
// Source: Tversky, A. (1977). "Features of similarity." Psychological
// Review 84(4):327-352, §2 — the asymmetric similarity model. Equation
// (1) on p. 332 ("the ratio model") is the canonical formulation; the
// worked examples on pp. 331-332 motivate the asymmetric prototype /
// variant interpretation.
//
// Algorithm (multiset / weighted form, per docs/requirements.md §7.2.4):
//
//   For strings A, B, q-gram size n, and weights α, β with α ≥ 0, β ≥ 0,
//   α + β > 0:
//
//     QA = multiset of overlapping length-n substrings of A
//     QB = multiset of overlapping length-n substrings of B
//
//     |QA ∩ QB| = Σ_{k ∈ keys(QA) ∩ keys(QB)} min(countA[k], countB[k])
//     |QA − QB| = Σ countA[k] − |QA ∩ QB|         (multiset difference)
//     |QB − QA| = Σ countB[k] − |QA ∩ QB|
//
//     T(A, B, α, β) = |QA ∩ QB|
//                     ─────────────────────────────────────────
//                     |QA ∩ QB| + α·|QA − QB| + β·|QB − QA|
//
// Properties (Tversky 1977 §2):
//
//   - Asymmetry:                T(A, B, α, β) ≠ T(B, A, α, β) in general
//                               when α ≠ β AND |A−B| ≠ |B−A|.
//   - Parameter-swap symmetry:  T(A, B, α, β) = T(B, A, β, α) ALWAYS.
//   - α = β symmetry:           T(A, B, α, α) = T(B, A, α, α).
//   - Jaccard degeneracy:       T(A, B, 1, 1) = J(A, B) (Q-Gram Jaccard).
//   - Dice degeneracy:          T(A, B, 0.5, 0.5) = DSC(A, B) (Sørensen-Dice).
//
// The Jaccard / Dice degeneracies are EXACT bit-for-bit on the same
// q-gram multisets — the arithmetic for both reduces to the same single
// integer-valued division (Jaccard) or to (2·int)/(int+int) (Dice) once
// the constant-weight α and β are folded in. This is the load-bearing
// cross-validation gate for Tversky correctness; without an external
// reference library, these algebraic equivalences plus the RV-T1 vs
// RV-T2 input-swap discrimination ARE the third-party-verifiable proof.
//
// Conventions (mirror the Phase 2/3/4 + plan 05-01/05-02/05-03
// short-circuit pattern):
//
//   - both-empty   → 1.0  (a == b identity short-circuit; vacuous match)
//   - identical    → 1.0  (a == b short-circuit, α/β irrelevant)
//   - one-empty    → 0.0
//
// Direct-call validation (Phase 8.5 Q2 data-vs-parameter framework,
// docs/requirements.md §6.A — supersedes the original CONTEXT.md §5
// lock by adding NaN/Inf rejection and typed-error panic values):
//
//   - n < 1 panics with `fmt.Errorf("%w: …", ErrInvalidQGramSize, …)`.
//     The string portion still contains "fuzzymatch: invalid q-gram
//     size" (the sentinel's Error() text) so log scraping remains
//     unaffected; the wrapping enables `errors.Is(panicValue.(error),
//     ErrInvalidQGramSize)` discrimination on recover().
//   - α or β is NaN or ±Inf panics with
//     `fmt.Errorf("%w: …", ErrInvalidTverskyParam, …)`.
//   - α < 0, β < 0, or α + β == 0 panics with the same typed error
//     wrapping `ErrInvalidTverskyParam`. The α + β > 0 constraint
//     is implemented as `α == 0 && β == 0` rather than `α + β == 0` to
//     avoid any float-comparison anxiety on slightly-negative sums
//     (the upstream α < 0 / β < 0 rejection means we know both are
//     non-negative by the time the joint-zero test runs).
//   - The Phase 8 Scorer option WithTverskyAlgorithm returns
//     ErrInvalidQGramSize / ErrInvalidTverskyParam as typed errors on
//     the same inputs — the panic is reserved for the direct-call
//     surface where programmer error must fail loudly.
//
// Asymmetry-discriminating reference vector pair (RV-T1 / RV-T2 from
// RESEARCH.md §2.4 — load-bearing regression gate):
//
//   TverskyScore("abcd", "abcdef", 2, 0.8, 0.2) =
//
//     QA = bigrams("abcd")   = {ab:1, bc:1, cd:1}            — total 3
//     QB = bigrams("abcdef") = {ab:1, bc:1, cd:1, de:1, ef:1} — total 5
//     |A∩B| = 3; |A−B| = 0; |B−A| = 2;
//     T = 3 / (3 + 0.8·0 + 0.2·2) = 3 / 3.4 = 0.8823529411764706
//
//   TverskyScore("abcdef", "abcd", 2, 0.8, 0.2) =
//
//     QA = bigrams("abcdef") = {ab:1, bc:1, cd:1, de:1, ef:1} — total 5
//     QB = bigrams("abcd")   = {ab:1, bc:1, cd:1}            — total 3
//     |A∩B| = 3; |A−B| = 2; |B−A| = 0;
//     T = 3 / (3 + 0.8·2 + 0.2·0) = 3 / 4.6 = 0.6521739130434783
//
//   0.8823529411764706 ≠ 0.6521739130434783 — the input swap with the
//   SAME (α, β) yields a different score. A silent α/β swap inside the
//   implementation would also cause RV-T1 to return 0.6521 and RV-T2 to
//   return 0.8823 — both still in [0, 1], both still passing
//   RangeBounds + Identity invariants — but they'd swap. The
//   TestTversky_AsymmetryDirectionSensitive unit test, the
//   TestProp_TverskyScore_AsymmetricWhenAlphaNeqBeta property test,
//   AND the BDD asymmetry scenario together form the three-layer
//   defence against parameter-order regressions.
//
// Source-origin discipline (per algorithm-licensing-standards):
//
//   - Primary source:        Tversky 1977 §2 eq. (1) p. 332 (Psychological
//                            Review 84(4):327-352).
//   - Cross-validation:      hand-derived RV-T1..RV-T4 reference vectors
//                            in tversky_test.go, each with the formula
//                            derivation embedded in the test comment.
//                            Plus bit-exact algebraic cross-checks
//                            against QGramJaccardScore (α=β=1.0) and
//                            SorensenDiceScore (α=β=0.5) at both the
//                            unit-test and property-test levels. Per
//                            CONTEXT.md §4 LOCKED, no Python toolchain
//                            is used.
//   - Tie-break:             none (Tversky is unambiguous; the asymmetry
//                            gate is the parameter-order regression
//                            detector, not a tie-break).
//   - GPL/LGPL provenance:   none.
//   - Code copied verbatim:  none.
//
// Implementation discipline:
//
//   - NO init()-time table builds (per docs/requirements.md §5(12)).
//   - NO map iteration on output paths (DET-03). All three multiset
//     cardinalities (|A∩B|, |A−B|, |B−A|) are integer counters derived
//     via map-length arithmetic and per-key lookups; the OUTPUT is a
//     single float64, not an ordered slice, so map-iteration order does
//     not affect the result.
//   - NO transcendental float operations (DET-06): only integer
//     arithmetic, float64() casts, and a single multiplication +
//     addition + division. No math.Pow / math.Log / math.Exp / math.FMA
//     / math.Sqrt.
//   - T formula uses explicit left-to-right parenthesisation
//     `float64(intersection) / (float64(intersection) + (alpha *
//     float64(aMinusB)) + (beta * float64(bMinusA)))` per DET-06 — no
//     associativity ambiguity for cross-platform determinism.
//   - Identity short-circuit a == b BEFORE any extraction work — both
//     covers the both-empty case AND avoids two map allocations on
//     identical inputs (regardless of α/β).
//   - The rune-path variant calls extractQGramsRunes which performs
//     the []rune conversion once per side; the byte-path variant
//     performs zero rune-slice allocations.
//
// Public surface (two functions; no rune-Distance variant — Tversky
// does not expose a distance, only a similarity score):
//
//   - TverskyScore(a, b string, n int, alpha, beta float64) float64
//   - TverskyScoreRunes(a, b string, n int, alpha, beta float64) float64
//
// Only TverskyScore is registered in the dispatch table (slot
// AlgoTversky — see algoid.go). The dispatch table maps AlgoID to
// (a, b string) float64 and has no place for n, α, or β; per CONTEXT.md
// "Claude's Discretion" the dispatch wrapper binds n=3, α=β=1.0 — the
// Jaccard fallback. Callers needing genuine asymmetric Tversky semantics
// invoke TverskyScore directly or, in Phase 8, use
// WithTverskyAlgorithm(weight, alpha, beta) on the Scorer.
//
// Worst-case complexity: O(la + lb) time + O(la + lb) space for the
// two multiset maps. Pure-function library — caller controls input
// size; the algorithm has no input-validation rejection on long input.

package fuzzymatch

import (
	"fmt"
	"math"
)

// TverskyScore returns the Tversky asymmetric similarity of the q-gram
// multisets of a and b: T(A, B, α, β) = |QA ∩ QB| / (|QA ∩ QB| +
// α·|QA − QB| + β·|QB − QA|), in [0.0, 1.0]. Operates on bytes —
// multi-byte UTF-8 inputs split q-grams at byte boundaries, which can
// produce different results than TverskyScoreRunes on non-ASCII input.
//
// Tversky is symmetric when α = β. When α ≠ β, swapping (a, b) generally
// produces a different score: TverskyScore(a, b, n, α, β) ≠
// TverskyScore(b, a, n, α, β). The parameter-swap symmetry
// TverskyScore(a, b, n, α, β) = TverskyScore(b, a, n, β, α) holds
// always.
//
// Special cases:
//
//   - α = β = 1.0 reduces to Q-Gram Jaccard:
//     TverskyScore(a, b, n, 1.0, 1.0) == QGramJaccardScore(a, b, n)
//     bit-for-bit (Tversky 1977 §2; the multiset arithmetic collapses
//     identically in both formulae).
//   - α = β = 0.5 reduces to Sørensen-Dice:
//     TverskyScore(a, b, n, 0.5, 0.5) == SorensenDiceScore(a, b, n)
//     bit-for-bit.
//
// The q-gram size n MUST be >= 1; n < 1 panics with a value wrapping
// ErrInvalidQGramSize (string message "fuzzymatch: invalid q-gram
// size"). Discriminate via errors.Is on a recovered panic value.
//
// The weights α and β MUST satisfy: not NaN, finite (not ±Inf),
// α >= 0, β >= 0, AND α + β > 0. Equivalently: both weights are
// finite, non-negative, and at least one is strictly positive.
// Violations panic with a value wrapping ErrInvalidTverskyParam
// (string message "fuzzymatch: invalid tversky parameter") on the
// following inputs:
//
//   - α or β is NaN
//   - α or β is ±Inf
//   - α < 0 or β < 0
//   - α + β == 0 (the joint α==0 && β==0 case)
//
// Per Phase 8.5 Q2 (data-vs-parameter framework, docs/requirements.md
// §6.A): comparison-data inputs (a, b) are lenient; parameter inputs
// (n, α, β) are strict. Direct calls panic with a typed-error value
// (`fmt.Errorf("%w: …", ErrInvalidTverskyParam, …)`) so consumers
// using `recover()` can discriminate via `errors.Is`. The Phase 8
// Scorer option WithTverskyAlgorithm returns ErrInvalidQGramSize /
// ErrInvalidTverskyParam as typed return values on the same inputs.
//
// Conventions:
//
//   - TverskyScore("",   "",   n, α, β) == 1.0  (both-empty identity;
//     covered by the a == b short-circuit; α/β irrelevant)
//   - TverskyScore("",   "abc", n, α, β) == 0.0 (one-empty)
//   - TverskyScore("abc", "",   n, α, β) == 0.0 (one-empty)
//   - TverskyScore("hello", "hello", n, α, β) == 1.0 (identity; α/β irrelevant)
//
// When both inputs are non-empty AND each is shorter than n, the q-gram
// extraction returns empty multisets on both sides; the both-empty
// guard does not fire (the input strings differ even though their
// q-gram views are empty), so the function falls through and the
// both-extractions-empty guard inside tverskyFromQGramMaps returns 1.0
// by convention (a vacuous match — both q-gram views are empty).
//
// Reference vectors (RV-T1 / RV-T2 — load-bearing asymmetry pair):
//
//	TverskyScore("abcd", "abcdef", 2, 0.8, 0.2) = 0.8823529411764706
//	TverskyScore("abcdef", "abcd", 2, 0.8, 0.2) = 0.6521739130434783
//
//	The two scores differ — the input swap with the SAME (α, β) is the
//	direct evidence of direction-sensitivity. See the file-level godoc
//	above for the full derivation.
func TverskyScore(a, b string, n int, alpha, beta float64) float64 { //nolint:gocyclo // Q2 strict-parameter framework — five NaN/Inf/sign guards before the core Tversky computation, each a distinct typed-error panic
	if a == b {
		return 1.0 // identity short-circuit (covers both-empty too)
	}
	if n < 1 {
		// Phase 8.5 Q2 — typed-error panic so recover() callers can
		// discriminate via errors.Is(panicValue.(error), sentinel).
		panic(fmt.Errorf("%w: TverskyScore requires n >= 1 (got n=%d)", ErrInvalidQGramSize, n))
	}
	// Phase 8.5 Q2 — strict-parameter discipline on direct calls.
	// NaN/±Inf must be tested FIRST because NaN compares false to
	// every range comparison and would otherwise sneak past the
	// α < 0 / β < 0 / α+β==0 tests. ±Inf in α or β would propagate
	// through α·|A−B| + β·|B−A| as ±Inf, breaking the [0, 1]
	// composite guarantee. The α + β > 0 invariant is implemented
	// as (α == 0 && β == 0) to dodge any float-comparison anxiety
	// on α + β being slightly less than 0; the upstream α < 0 / β < 0
	// rejection means we know both are non-negative at this point.
	if math.IsNaN(alpha) || math.IsNaN(beta) || math.IsInf(alpha, 0) || math.IsInf(beta, 0) {
		panic(fmt.Errorf("%w: TverskyScore requires finite, non-NaN alpha and beta (got alpha=%v, beta=%v)", ErrInvalidTverskyParam, alpha, beta))
	}
	if alpha < 0 || beta < 0 || (alpha == 0 && beta == 0) {
		panic(fmt.Errorf("%w: TverskyScore requires alpha >= 0, beta >= 0, and alpha+beta > 0 (got alpha=%v, beta=%v)", ErrInvalidTverskyParam, alpha, beta))
	}
	if a == "" || b == "" {
		return 0.0
	}
	qa := extractQGrams(a, n)
	qb := extractQGrams(b, n)
	return tverskyFromQGramMaps(qa, qb, alpha, beta)
}

// TverskyScoreRunes returns the Tversky asymmetric similarity of the
// q-gram multisets of a and b under rune semantics. Multi-byte UTF-8
// sequences are compared atomically: extractQGramsRunes("café", 2)
// yields three rune-bigrams {"ca", "af", "fé"} where "fé" is the
// multi-byte UTF-8 encoding of the two-rune window.
//
// Allocates two []rune slices on the heap (one per side) plus the
// two multiset maps. For pure-ASCII inputs prefer TverskyScore.
//
// Direct-call validation, edge cases, conventions, and asymmetry
// guarantees are identical to TverskyScore (see godoc above). The
// rune-path divergence from the byte path is that windows are over
// runes, not bytes — "café" has 4 runes, so TverskyScoreRunes("café",
// x, 2, α, β) is well-defined for any x, while the byte path treats
// "café" as 5 bytes.
//
// Reference vector:
//
//	TverskyScoreRunes("café", "cafe", 2, 0.5, 0.5) = 2/3 ≈ 0.6666666666666666
//
//	QA = rune-bigrams("café") = {"ca":1, "af":1, "fé":1}; total 3
//	QB = rune-bigrams("cafe") = {"ca":1, "af":1, "fe":1}; total 3
//	|A∩B| = 2 (ca + af); |A−B| = 1 (fé); |B−A| = 1 (fe);
//	T = 2 / (2 + 0.5·1 + 0.5·1) = 2 / 3 = 0.6666666666666666
func TverskyScoreRunes(a, b string, n int, alpha, beta float64) float64 { //nolint:gocyclo // rune-surface mirror of TverskyScore — same Q2 strict-parameter guards before the core computation
	if a == b {
		return 1.0 // identity short-circuit (covers both-empty too)
	}
	if n < 1 {
		// Phase 8.5 Q2 — typed-error panic (see TverskyScore for the
		// full rationale; the rune surface mirrors the byte surface).
		panic(fmt.Errorf("%w: TverskyScoreRunes requires n >= 1 (got n=%d)", ErrInvalidQGramSize, n))
	}
	if math.IsNaN(alpha) || math.IsNaN(beta) || math.IsInf(alpha, 0) || math.IsInf(beta, 0) {
		panic(fmt.Errorf("%w: TverskyScoreRunes requires finite, non-NaN alpha and beta (got alpha=%v, beta=%v)", ErrInvalidTverskyParam, alpha, beta))
	}
	if alpha < 0 || beta < 0 || (alpha == 0 && beta == 0) {
		panic(fmt.Errorf("%w: TverskyScoreRunes requires alpha >= 0, beta >= 0, and alpha+beta > 0 (got alpha=%v, beta=%v)", ErrInvalidTverskyParam, alpha, beta))
	}
	if a == "" || b == "" {
		return 0.0
	}
	qa := extractQGramsRunes(a, n)
	qb := extractQGramsRunes(b, n)
	return tverskyFromQGramMaps(qa, qb, alpha, beta)
}

// tverskyFromQGramMaps computes T(A, B, α, β) = |QA ∩ QB| /
// (|QA ∩ QB| + α·|QA − QB| + β·|QB − QA|) from two pre-extracted q-gram
// multisets and the user-supplied weights α, β. The multiset cardinalities
// are:
//
//	totalA       = Σ countA[k]            (sum of map values for QA)
//	totalB       = Σ countB[k]
//	intersection = Σ_{k ∈ keys(QA) ∩ keys(QB)} min(countA[k], countB[k])
//	aMinusB      = totalA - intersection  (multiset difference; equiv. to
//	                                       Σ max(0, countA[k] - countB[k])
//	                                       by the identity min+max=x for
//	                                       non-negative x, y)
//	bMinusA      = totalB - intersection
//
// Helper invariant: the OUTPUT is a single float64 derived from
// integer cardinalities. The internal map iterations to compute
// `total*` and `intersection` produce INTEGER counters whose values
// do not depend on iteration order (integer addition is associative);
// no ordered output is constructed. DET-03 satisfied.
//
// When both qa and qb are empty, returns 1.0 by the both-empty
// convention (a vacuous match — no q-grams to disagree about). This
// branch ALSO defends against the unreachable-but-defensive 0/0 case:
// if intersection = 0 AND aMinusB = 0 AND bMinusA = 0 then both
// multisets are empty, which means totalA = totalB = 0, which is the
// len(qa)==0 && len(qb)==0 early-return path — the divide-by-zero is
// thus structurally impossible by the time we reach the division.
//
// The T formula uses explicit left-to-right parenthesisation
// `float64(intersection) / (float64(intersection) + (alpha *
// float64(aMinusB)) + (beta * float64(bMinusA)))` per DET-06 — no
// associativity ambiguity across CI platforms.
//
// Algebraic cross-check (RV-T3 / RV-T4 from RESEARCH.md §2.4):
//
//   - α = β = 1.0: denom = intersection + (aMinusB + bMinusA) =
//     totalA + totalB − intersection = union. T = intersection/union = J. ✓
//   - α = β = 0.5: denom = intersection + 0.5·(aMinusB + bMinusA) =
//     intersection + 0.5·(totalA + totalB − 2·intersection) = (totalA +
//     totalB)/2. T = 2·intersection/(totalA + totalB) = DSC. ✓
//
// These equivalences hold bit-for-bit with QGramJaccardScore /
// SorensenDiceScore on the same q-gram multisets — pinned by both unit
// tests AND property tests as the load-bearing correctness gate.
func tverskyFromQGramMaps(qa, qb map[string]int, alpha, beta float64) float64 { //nolint:gocyclo // empty-multiset short-circuits + smaller-side intersection iteration + asymmetric α/β weighting + 0/0 denominator guard folded for hot-path locality; same precedent as damerau_osa.go::damerauOSADP per godoc above
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
	// Note `small` / `large` were swapped purely as an iteration-cost
	// optimisation; the resulting `intersection` is symmetric in (qa, qb)
	// regardless of which side was walked, because min(x, y) = min(y, x).
	// The asymmetry of T comes from |A−B| ≠ |B−A| below — those use
	// totalA / totalB which preserve the original argument order.
	aMinusB := totalA - intersection
	bMinusA := totalB - intersection
	// Defensive 0/0 guard: when intersection == 0 AND α·aMinusB +
	// β·bMinusA == 0, the denominator collapses. This happens when one
	// side's q-gram multiset is empty (len(s) < n) AND the other
	// side's residual is weighted by 0 (e.g. α=0 with qb empty so the
	// only non-zero residual is α·totalA which becomes 0). The
	// semantically correct answer is 0.0 — the inputs share no q-grams
	// AND the partial-empty-multiset case mirrors QGramJaccardScore
	// (which returns 0/totalA = 0.0 on the same input).
	//
	// The full-both-empty case (len(qa) == 0 && len(qb) == 0) returned
	// 1.0 above; this guard only fires when ONE side is empty AND the
	// non-empty side's residual is α/β-zeroed.
	if intersection == 0 {
		denom := (alpha * float64(aMinusB)) + (beta * float64(bMinusA))
		if denom == 0.0 {
			return 0.0
		}
		return 0.0 // 0 / positive = 0; explicit short-circuit avoids the division.
	}
	denom := float64(intersection) + (alpha * float64(aMinusB)) + (beta * float64(bMinusA))
	// Single division on integer-derived float64 values (with the user-
	// supplied float64 weights α, β). IEEE-754 correctly-rounded
	// multiplication, addition, and division produce byte-identical
	// output across all four CI platforms (DET-06 satisfied).
	return float64(intersection) / denom
}
