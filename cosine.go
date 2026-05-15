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

// cosine.go implements the Cosine similarity over q-gram frequency
// vectors for the fuzzymatch catalogue. Cosine is the textbook
// vector-space measure from information retrieval (Salton & McGill 1983
// §4.1 eq. 4.4 p.121) and is the LOAD-BEARING cross-platform float-
// determinism algorithm for Phase 5: the algorithms.json golden file
// Cosine entries are the gate that detects any drift in float reduction
// order across the four CI platforms (linux/amd64, linux/arm64,
// darwin/arm64, windows/amd64).
//
// Source: Salton, G., & McGill, M. J. (1983). "Introduction to Modern
// Information Retrieval." McGraw-Hill, Chapter 4 §4.1, equation 4.4
// (Cosine measure) on p.121 — the canonical vector-space-model cosine
// similarity over term-frequency vectors. The q-gram-based
// formulation (frequency vectors over the multiset of overlapping
// length-n substrings) is the textbook adaptation widely cited in
// approximate string matching since the 1990s.
//
// Algorithm (per docs/requirements.md §7.2.3):
//
//	For strings A, B and q-gram size n:
//
//	  QA = multiset of overlapping length-n substrings of A
//	  QB = multiset of overlapping length-n substrings of B
//
//	  A · B  = Σ_{k ∈ keys(QA) ∩ keys(QB)} countA[k] · countB[k]
//	  ‖A‖²   = Σ countA[k]²            (sum of squared multiset counts)
//	  ‖B‖²   = Σ countB[k]²
//
//	  cos(A, B) = (A · B) / (‖A‖ × ‖B‖)
//	            = (A · B) / (sqrt(‖A‖²) × sqrt(‖B‖²))
//
// Conventions (mirror the Phase 2/3/4 short-circuit pattern; identical
// to the Q-Gram Jaccard / Sørensen-Dice conventions for consistency):
//
//   - both-empty   → 1.0 (covered by the a == b identity short-circuit)
//   - identical    → 1.0 (a == b short-circuit)
//   - one-empty    → 0.0
//   - orthogonal   → 0.0 (empty intersection → dot=0 → cos=0)
//
// Direct-call validation (CONTEXT.md §5 LOCKED):
//
//   - n < 1 panics with the message "fuzzymatch: invalid q-gram size".
//     The Phase 8 Scorer option WithCosineAlgorithm returns
//     ErrInvalidQGramSize instead — the panic is reserved for the
//     direct-call surface where programmer error must fail loudly.
//
// Salton & McGill 1983 §4.1 reference vector (RV-C1):
//
//	CosineScore("abc", "abcd", 2) =
//
//	  QA = bigrams("abc")  = {ab:1, bc:1};       ‖A‖² = 2
//	  QB = bigrams("abcd") = {ab:1, bc:1, cd:1}; ‖B‖² = 3
//	  intersection (sorted) = ["ab", "bc"];      dot = 2
//	  cos = 2 / (sqrt(2)·sqrt(3)) = 2/sqrt(6) = 0.8164965809277261
//
// LOAD-BEARING DETERMINISM (CONTEXT.md §3 LOCKED):
//
//	The dot-product reduction MUST iterate intersection keys in
//	SORTED ORDER. The implementation builds the intersection key
//	slice, calls sort.Strings on it, then iterates the sorted slice
//	with explicit (x*y) + z parenthesisation per DET-06.
//
//	Sum-of-squares norm reductions iterate map values in any order
//	(integer addition is exactly associative; the float-determinism
//	risk is on the dot-product reduction only).
//
//	math.Sqrt is the ONLY math.* call used; it is IEEE-754 correctly
//	rounded on all four CI platforms per RESEARCH.md §3.5. NO math.Pow,
//	NO math.Log, NO math.Exp, NO math.FMA.
//
// Source-origin discipline (per algorithm-licensing-standards):
//
//   - Primary source:        Salton & McGill 1983 §4.1 eq. 4.4 p.121
//                            ("Introduction to Modern Information
//                            Retrieval", McGraw-Hill).
//   - Cross-validation:      none — hand-derived RV-C1..RV-C5
//                            reference vectors in cosine_test.go with
//                            17-significant-digit float64 derivations
//                            in test comments per CONTEXT.md §4 LOCKED.
//                            No external library reference; no Python
//                            toolchain.
//   - Tie-break:             none (Cosine over q-gram frequency
//                            vectors is unambiguous; the algorithm is
//                            a single inner-product / norm computation).
//   - GPL/LGPL provenance:   none.
//   - Code copied verbatim:  none.
//
// Implementation discipline:
//
//   - NO init()-time table builds (per docs/requirements.md §5(12)).
//   - NO map iteration on output paths for the dot product (DET-03):
//     the intersection key slice is built then sort.Strings'd before
//     iteration; the sum-of-squares norm reductions iterate map
//     values (not output keys) and are safe because integer addition
//     is exactly associative.
//   - NO transcendental float operations except math.Sqrt (DET-06):
//     only +, -, *, /, comparisons, float64() casts, and math.Sqrt.
//   - Identity short-circuit a == b BEFORE any extraction work — both
//     covers the both-empty case AND avoids two map allocations and
//     the sort on identical inputs.
//   - Defensive empty-multiset guard inside cosineFromQGramMaps
//     handles len(qa)==0 && len(qb)==0 (e.g. n > min(len(a), len(b))
//     with non-identical inputs) and len(qa)==0 || len(qb)==0
//     (asymmetric short-circuit) without ever producing 0/0 NaN.
//   - The rune-path variant calls extractQGramsRunes which performs
//     the []rune conversion once per side; the byte-path variant
//     performs zero rune-slice allocations.
//
// Public surface (two functions; no rune-Distance variant — Cosine
// does not expose a distance, only a similarity score):
//
//   - CosineScore(a, b string, n int) float64
//   - CosineScoreRunes(a, b string, n int) float64
//
// Only CosineScore is registered in the dispatch table (slot
// AlgoCosine — see algoid.go) with a default n = 3 wrapper — the
// dispatch table maps AlgoID to (a, b string) float64 and has no
// place for the n parameter. Specific n overrides happen via the
// Phase 8 Scorer option WithCosineAlgorithm(weight, n).
//
// Worst-case complexity: O(la + lb) extraction + O(k log k) for
// sort.Strings on the intersection slice (k = |intersection| ≤
// min(distinct q-grams in QA, QB)) + O(k) dot-product reduction +
// O(la + lb) norm computation. Pure-function library — caller
// controls input size; the algorithm has no input-validation rejection
// on long input.
//
// Allocation budget (RESEARCH.md §4.1, docs/requirements.md §14.1):
// two map allocations from extractQGrams + one []string slice for the
// intersection keys + cap-hint backing arrays. Realistic ceiling on
// short inputs: ≤ 5 allocs (Sørensen-Dice's 4 + 1 sorted-key slice).

package fuzzymatch

import (
	"math"
	"sort"
)

// CosineScore returns the Cosine similarity of the q-gram frequency
// vectors of a and b: cos(A, B) = (A · B) / (‖A‖ × ‖B‖), in [0.0, 1.0]
// (Salton & McGill 1983 §4.1 eq. 4.4 p.121). Operates on bytes —
// multi-byte UTF-8 inputs split q-grams at byte boundaries, which can
// produce different results than CosineScoreRunes on non-ASCII input.
//
// Iteration order over the intersection keys is SORTED (sort.Strings)
// per CONTEXT.md §3 LOCKED; the dot-product reduction uses explicit
// (x*y) + z parenthesisation per DET-06. See cosine.go inline footnote
// for the FMA-risk surface (RESEARCH.md §3 OQ-1).
//
// The q-gram size n MUST be >= 1; n < 1 panics with the message
// "fuzzymatch: invalid q-gram size" (CONTEXT.md §5 LOCKED — direct
// calls fail loudly on programmer error; the Phase 8 Scorer option
// WithCosineAlgorithm returns ErrInvalidQGramSize instead).
//
// Conventions:
//
//   - CosineScore("",   "",   n) == 1.0  (both-empty identity)
//   - CosineScore("",   "abc", n) == 0.0 (one-empty)
//   - CosineScore("abc", "",   n) == 0.0 (one-empty)
//   - CosineScore("hello", "hello", n) == 1.0 (identity)
//   - CosineScore("abc", "xyz", 2) == 0.0 (orthogonal — empty
//     intersection → dot=0 → cos=0)
//
// When both inputs are non-empty AND each is shorter than n, the
// q-gram extraction returns empty multisets on both sides; the
// both-empty guard does not fire (the input strings differ even
// though their q-gram views are empty), so the function falls through
// and the both-extractions-empty guard inside cosineFromQGramMaps
// returns 1.0 by convention (a vacuous match — both q-gram views are
// empty).
//
// Reference vector (RV-C1 — load-bearing hand-derived irrational at n=2):
//
//	CosineScore("abc", "abcd", 2) = 2/sqrt(6) ≈ 0.8164965809277261
//
//	QA = {ab:1, bc:1}; ‖A‖² = 2;
//	QB = {ab:1, bc:1, cd:1}; ‖B‖² = 3;
//	intersection (sorted) = ["ab", "bc"]; dot = 2;
//	cos = 2/(sqrt(2)·sqrt(3)) = 2/sqrt(6) = 0.8164965809277261.
func CosineScore(a, b string, n int) float64 {
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
	return cosineFromQGramMaps(qa, qb)
}

// CosineScoreRunes returns the Cosine similarity of the q-gram
// frequency vectors of a and b under rune semantics. Multi-byte UTF-8
// sequences are compared atomically: extractQGramsRunes("café", 2)
// yields three rune-bigrams {"ca", "af", "fé"} where "fé" is the
// multi-byte UTF-8 encoding of the two-rune window.
//
// Allocates two []rune slices on the heap (one per side) plus the
// two multiset maps and the intersection-keys []string slice. For
// pure-ASCII inputs prefer CosineScore.
//
// Direct-call validation, edge cases, conventions, and determinism
// guarantees are identical to CosineScore (see godoc above). The
// rune-path divergence from the byte path is that windows are over
// runes, not bytes — "café" has 4 runes, so CosineScoreRunes("café",
// x, 2) is well-defined for any x, while the byte path treats "café"
// as 5 bytes.
//
// Reference vector (RV-C3 — load-bearing rune-path canary):
//
//	CosineScoreRunes("café", "cafe", 2) = 2/3 ≈ 0.6666666666666666
//
//	QA = rune-bigrams("café") = {"ca":1, "af":1, "fé":1}; ‖A‖² = 3;
//	QB = rune-bigrams("cafe") = {"ca":1, "af":1, "fe":1}; ‖B‖² = 3;
//	intersection (sorted byte-lex) = ["af", "ca"]; dot = 2;
//	cos = 2/(sqrt(3)·sqrt(3)) = 2/3 = 0.6666666666666666.
func CosineScoreRunes(a, b string, n int) float64 {
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
	return cosineFromQGramMaps(qa, qb)
}

// cosineFromQGramMaps computes cos(A, B) = (A · B) / (‖A‖ × ‖B‖) from
// two pre-extracted q-gram multisets. Salton & McGill 1983 §4.1 eq. 4.4
// factorised form: dot / (sqrt(normASq) * sqrt(normBSq)).
//
// LOAD-BEARING DETERMINISM (CONTEXT.md §3 LOCKED):
//
//	The intersection key slice is built by iterating one map and
//	checking membership in the other (slice contents identical
//	regardless of which side is iterated; the smaller map is walked
//	to minimise lookups). The slice is then sort.Strings'd in place
//	BEFORE the dot-product loop iterates it. Map-iteration order
//	does not affect the BUILT slice (a set), and sort.Strings makes
//	the iteration order canonical (byte-lexicographic, total,
//	deterministic across platforms — Go's sort.Strings uses `<` on
//	`string`, which is byte-comparison per the language spec).
//
//	The dot-product reduction uses explicit `(x*y) + z`
//	parenthesisation per DET-06 — the project-canonical form (matches
//	lcsstr.go line 215, swg.go, jaro.go).
//
//	Sum-of-squares norms iterate map VALUES in any order — integer
//	addition is exactly associative, so iteration-order independence
//	is automatic; this iteration does NOT touch the output path,
//	since the OUTPUT is a single integer (then a single math.Sqrt
//	per side).
//
//	math.Sqrt is the only math.* call used. It is IEEE-754 correctly
//	rounded on all four CI platforms per RESEARCH.md §3.5 — produces
//	byte-identical output for the same int input across linux/amd64,
//	linux/arm64, darwin/arm64, windows/amd64.
//
//	Final cosine: dot / (normA * normB) — single division on
//	IEEE-754 correctly rounded floats; explicit (normA * normB)
//	parenthesisation matches the Salton & McGill 1983 factorised
//	form and avoids the int64-overflow trap of computing
//	math.Sqrt(float64(normASq * normBSq)) (see RESEARCH.md
//	"Pitfall 6").
//
// FMA risk surface (RESEARCH.md §3, OQ-1): Go 1.26 may emit FMA on
// arm64 for the (x*y)+z pattern; parentheses do NOT defeat FMA fusion
// per golang/go#17895. The cross-platform CI matrix gate
// (testdata/golden/algorithms.json) is the load-bearing detector. If
// matrix divergence ever appears on Cosine entries, remediate by
// inserting an explicit float64 cast: dot = float64(float64(qa[k]) *
// float64(qb[k])) + dot — the explicit cast forces intermediate
// rounding and defeats FMA fusion. See RESEARCH.md §3.4 for the full
// remediation pattern; this is documentation only — no code change is
// required while the CI matrix passes.
//
// When both qa and qb are empty (e.g. n > min(len(a), len(b)) on
// non-identical inputs), returns 1.0 by the both-empty convention (a
// vacuous match — no q-grams to disagree about). When exactly one is
// empty, returns 0.0 (the asymmetric short-circuit guard).
func cosineFromQGramMaps(qa, qb map[string]int) float64 {
	if len(qa) == 0 && len(qb) == 0 {
		return 1.0
	}
	if len(qa) == 0 || len(qb) == 0 {
		return 0.0
	}
	// Build the intersection key slice. Walk the SMALLER map and check
	// membership in the larger — the slice content is identical
	// regardless of which side is iterated (intersection is symmetric).
	// Walking the smaller side keeps the lookup count to len(min(qa, qb)).
	small, large := qa, qb
	if len(qb) < len(qa) {
		small, large = qb, qa
	}
	intersectionKeys := make([]string, 0, len(small))
	for k := range small {
		if _, ok := large[k]; ok {
			intersectionKeys = append(intersectionKeys, k)
		}
	}
	// CONTEXT.md §3 LOCKED — sort intersection keys before iteration
	// to make the dot-product reduction order canonical and
	// cross-platform deterministic. Without this sort, Go's randomised
	// map iteration would visit keys in different orders across runs,
	// producing slight float-rounding divergences on the dot-product
	// reduction. sort.Strings is byte-lexicographic, total, and
	// platform-stable (RESEARCH.md §3.6).
	sort.Strings(intersectionKeys)
	// Dot product: explicit (x*y) + z parenthesisation per DET-06 —
	// the project-canonical form that matches lcsstr.go / swg.go /
	// jaro.go. Per the FMA-risk footnote in this function's godoc
	// above, parentheses do NOT prevent FMA fusion on arm64 (see
	// golang/go#17895); the empirical observation is that the
	// (integer-derived) values of qa[k] and qb[k] are small enough
	// that any FMA-vs-non-FMA divergence falls below the byte-diff
	// threshold of the algorithms.json gate. If matrix divergence
	// ever appears, remediate per RESEARCH.md §3.4.
	var dot float64
	for _, k := range intersectionKeys {
		dot = (float64(qa[k]) * float64(qb[k])) + dot
	}
	// Sum of squares for each norm. Map iteration order is randomised
	// but integer addition is exactly associative, so the per-side SUM
	// is deterministic regardless of order. DET-03 satisfied: the
	// output of THIS loop is a scalar int, not an ordered slice — the
	// determinism risk lives on the dot-product reduction only.
	var sumSquaresA, sumSquaresB int
	for _, c := range qa {
		sumSquaresA += c * c
	}
	for _, c := range qb {
		sumSquaresB += c * c
	}
	// math.Sqrt is the ONLY math.* call. IEEE-754 correctly rounded on
	// all four CI platforms per RESEARCH.md §3.5 — the Go compiler
	// intrinsifies to SQRTSD on amd64 and FSQRTD on arm64, both
	// IEEE-754 §5.4.1 conformant. Same int input → byte-identical
	// float64 output across platforms.
	normA := math.Sqrt(float64(sumSquaresA))
	normB := math.Sqrt(float64(sumSquaresB))
	// Final cosine: single division on IEEE-754 correctly rounded
	// floats; explicit (normA * normB) parenthesisation matches the
	// Salton & McGill 1983 factorised form. NOT
	// math.Sqrt(float64(sumSquaresA * sumSquaresB)) — that pattern
	// risks int64 overflow on pathological long inputs (RESEARCH.md
	// "Pitfall 6") and changes the rounding sequence.
	return dot / (normA * normB)
}
