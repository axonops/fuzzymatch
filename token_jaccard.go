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

// token_jaccard.go implements the TokenJaccard similarity for the
// fuzzymatch catalogue. TokenJaccard is the simplest Phase 6 algorithm:
// set-Jaccard |A ∩ B| / |A ∪ B| over the SET of tokens produced by
// Tokenise(s, DefaultTokeniseOptions()).
//
// Source: Jaccard, P. (1912). "The distribution of the flora in the
// alpine zone." New Phytologist 11(2):37-50, p. 43 — the set
// coefficient. This is the SAME primary source as Phase 5's Q-Gram
// Jaccard (qgram_jaccard.go); the divergence is in WHAT is being
// compared: TokenJaccard compares the deduplicated SET of TOKENS;
// Q-Gram Jaccard compares the MULTISET of overlapping length-n
// q-grams.
//
// Algorithm — TokenJaccardScore(a, b):
//
//   For strings A, B:
//
//     SA = SET of tokens of Tokenise(A, DefaultTokeniseOptions())
//          — multiplicity collapsed via map[string]struct{}
//     SB = SET of tokens of Tokenise(B, DefaultTokeniseOptions())
//
//     |SA ∩ SB| = #{ k : k ∈ SA AND k ∈ SB }   — scalar int
//     |SA ∪ SB| = |SA| + |SB| - |SA ∩ SB|       — set inclusion-exclusion
//
//     J(A, B) = |SA ∩ SB| / |SA ∪ SB|
//
// The integer-counter intersection cardinality is computed by iterating
// the SMALLER set and probing the larger set — DET-03 satisfied (output
// is a scalar int, not a slice from set iteration; integer addition is
// associative; map iteration order does not affect the result). This is
// the canonical RESEARCH.md Pattern 8 implementation pattern.
//
// Set vs Multiset distinction (LOCKED 2026-05-15):
//
//   TokenJaccard uses SET semantics (deduplicated tokens — token-presence
//   is a binary signal). Phase 5's Q-Gram Jaccard
//   (qgram_jaccard.go::QGramJaccardScore) uses MULTISET semantics over
//   q-gram counts. The semantic divergence is intentional per
//   06-RESEARCH.md Pattern 8: a token appearing twice in the input
//   doesn't make it "more present"; q-gram presence is a multiplicity
//   signal because longer overlapping runs of identical q-grams indicate
//   stronger similarity. RV-TJ3 ("a a b" vs "a b" → 1.0) is the keystone
//   regression gate for this distinction; the same input under multiset
//   Jaccard would yield 2/3.
//
// Conventions (mirror the Phase 5 q-gram Jaccard pattern):
//
//   - both-empty       → 1.0 (covered by a == b identity short-circuit
//                              AND the post-Tokenise both-empty guard).
//                              This is the STANDARD catalogue convention
//                              (LOCKED 2026-05-15) — TokenJaccard does
//                              NOT deviate like TokenSetRatio (which
//                              returns 0.0 per the LOCKED RapidFuzz
//                              issue #110 bug-for-bug compatibility).
//   - identical        → 1.0 (a == b short-circuit)
//   - one-empty        → 0.0
//   - identical tokens → 1.0 (a != b but Tokenise outputs produce
//                              identical sets after dedup)
//
// Source-origin discipline (per algorithm-licensing-standards):
//
//   - Primary source:        Jaccard 1912 p. 43 (same paper as
//                            Q-Gram Jaccard; applied to TOKEN sets here
//                            instead of q-gram multisets).
//   - Cross-validation:      NONE — hand-derived RV-TJ1..RV-TJ6
//                            reference vectors in token_jaccard_test.go
//                            per CONTEXT.md §1b LOCKED. The RapidFuzz
//                            cross-validation corpus does NOT include
//                            TokenJaccard entries. The set-Jaccard
//                            formula is unambiguous from Jaccard 1912
//                            and the integer-counter intersection
//                            cardinality has no implementation choices
//                            left to ambiguate.
//   - Tie-break:             none (set-Jaccard is unambiguous; the
//                            cardinalities are associative integer
//                            arithmetic).
//   - GPL/LGPL provenance:   none.
//   - Code copied verbatim:  none.
//
// Implementation discipline:
//
//   - NO init()-time table builds (per docs/requirements.md §5(12)).
//   - NO map iteration on output paths (DET-03). The intersection
//     cardinality is computed by iterating the SMALLER set and counting
//     hits in the LARGER set — output is a scalar int, NOT an ordered
//     slice (LOCKED 2026-05-15 — integer-counter intersection
//     cardinality satisfies DET-03 without any sort).
//   - NO transcendental float operations (DET-06): only integer
//     arithmetic, float64() casts, and a single division. No math.X
//     beyond the project allowlist (none used here).
//   - Identity short-circuit `if a == b { return 1.0 }` BEFORE Tokenise
//     to avoid the `make([]string, 0, 4)` allocation on identical
//     inputs (matches qgram_jaccard.go's IN-04 closure pattern).
//   - Tokeniser-divergence note (OQ-1 — same as TokenSortRatio):
//     fuzzymatch's Tokenise is identifier-aware (camelCase / snake_case
//     / kebab-case / dot-case + lowercasing). For whitespace-only
//     lowercase ASCII inputs the behaviour is unambiguous (every
//     reference vector uses such inputs to keep the derivation
//     reviewer-verifiable). For mixed identifier-style inputs the
//     project tokenisation produces semantically richer splits.
//   - NO public *Runes variant — Tokenise handles UTF-8 internally per
//     06-CONTEXT.md §6 LOCKED (same convention as TokenSortRatio /
//     TokenSetRatio).
//
// Allocation budget:
//
//   - ≤ 4 baseline allocations per call on short inputs:
//       * 2 for the two Tokenise outputs (one []string per side)
//       * 2 for the two map[string]struct{} sets
//     The integer-counter intersection cardinality is zero-allocation
//     (only an int counter + a single division).
//     Realistic ceiling for short-input ASCII is set at 8 in
//     token_jaccard_test.go's TestTokenJaccardScore_AllocsBudget to
//     accommodate Go's per-token string allocation in Tokenise's
//     lowercase fold path.
//
// Public surface (one function — the dispatched byte-path score):
//
//   - TokenJaccardScore(a, b string) float64
//
// Registered in dispatch table slot AlgoTokenJaccard (slot 17 — see
// algoid.go). The dispatch table maps AlgoID to (a, b string) float64
// with no place for parameters; TokenJaccard has no parameters so the
// wrapper is the function value directly (no closure needed — mirrors
// dispatch_lcsstr.go and dispatch_token_sort_ratio.go).
//
// Worst-case complexity: O(|a|+|b|) time for Tokenise + O(|tA|+|tB|)
// time for set construction + O(min(|setA|,|setB|)) time for the
// intersection scan, where t* is the token count on each side. Space:
// O(|setA|+|setB|) for the two maps. Pure-function library — caller
// controls input size; the algorithm has no input-validation rejection
// on long input.

package fuzzymatch

// TokenJaccardScore returns the TokenJaccard similarity between a and b
// in [0.0, 1.0]: tokenise both sides using DefaultTokeniseOptions(),
// deduplicate each token list to a set, then compute the Jaccard
// coefficient |A ∩ B| / |A ∪ B| over the two sets.
//
// Conventions (the STANDARD catalogue both-empty convention; distinct
// from TokenSetRatio's LOCKED RapidFuzz issue #110 deviation):
//
//   - TokenJaccardScore("",        "")              == 1.0  (both-empty STANDARD)
//   - TokenJaccardScore("hello",   "hello")         == 1.0  (identity)
//   - TokenJaccardScore("hello",   "")              == 0.0  (one-empty)
//   - TokenJaccardScore("",        "hello")         == 0.0  (one-empty)
//   - TokenJaccardScore("a b c",   "b c d")         == 0.5   (RV-TJ1)
//   - TokenJaccardScore("a b",     "a b c")         == 2.0/3 (RV-TJ2 subset)
//   - TokenJaccardScore("a a b",   "a b")           == 1.0  (RV-TJ3 SET dedup)
//   - TokenJaccardScore("a b c",   "x y z")         == 0.0   (RV-TJ4 disjoint)
//
// Symmetric across argument order — TokenJaccardScore(a, b) ==
// TokenJaccardScore(b, a) — because the set construction is
// order-independent and the integer-counter intersection cardinality is
// invariant under argument swap. The byte-equality is exact (no float
// tolerance needed); see TestProp_TokenJaccardScore_Symmetric for the
// quick.Check property.
//
// Set vs Multiset distinction: TokenJaccard uses SET semantics on the
// deduplicated token list (per RESEARCH.md Pattern 8). It is DISTINCT
// from Phase 5's Q-Gram Jaccard (qgram_jaccard.go::QGramJaccardScore)
// which uses MULTISET semantics over q-gram counts. RV-TJ3 ("a a b" vs
// "a b" → 1.0) is the keystone regression gate. The same inputs under
// Q-Gram Jaccard's multiset semantics yield a different score (≠ 1.0)
// because q-gram presence is a multiplicity signal.
//
// Tokeniser divergence: fuzzymatch's Tokenise is identifier-aware
// (camelCase / snake_case / etc.); for inputs with mixed identifier
// styles the project tokenisation produces semantically richer splits
// than a pure whitespace split. For whitespace-only lowercase ASCII
// inputs (which all six hand-derived reference vectors use) the
// behaviour is unambiguous.
//
// Reference vector (hand-derived per CONTEXT.md §1b LOCKED):
//
//	TokenJaccardScore("a b c", "b c d") = 0.5
//	  set A = {a,b,c}; set B = {b,c,d}; |∩| = 2; |∪| = 4; J = 2/4 = 0.5
//
// This function operates on the deduplicated SET of tokens (presence is
// a binary signal). For multi-byte UTF-8 token contents, Tokenise still
// produces well-formed UTF-8 tokens; map[string]struct{} keys compare
// the full UTF-8 byte sequence. There is no rune-path variant: Tokenise
// is UTF-8-aware so the rune semantic is already preserved at the
// tokenisation layer.
func TokenJaccardScore(a, b string) float64 {
	if a == b {
		return 1.0 // identity short-circuit — avoids Tokenise allocations
	}
	tokensA := Tokenise(a, DefaultTokeniseOptions())
	tokensB := Tokenise(b, DefaultTokeniseOptions())
	// Both-Tokenised-empty: vacuous match. STANDARD catalogue convention
	// (LOCKED 2026-05-15) — TokenJaccard does NOT deviate like
	// TokenSetRatio. This branch fires for inputs that are pure
	// separators on both sides (e.g. "  " vs "___") because the a == b
	// short-circuit above doesn't catch those when the raw strings differ.
	if len(tokensA) == 0 && len(tokensB) == 0 {
		return 1.0
	}
	// One-Tokenised-empty: 0.0. Matches the catalogue convention for
	// asymmetric empty-vs-non-empty inputs.
	if len(tokensA) == 0 || len(tokensB) == 0 {
		return 0.0
	}
	setA := tokensToSet(tokensA)
	setB := tokensToSet(tokensB)
	intersection := setIntersectionCardinality(setA, setB)
	// Set inclusion-exclusion: |A ∪ B| = |A| + |B| - |A ∩ B|.
	union := len(setA) + len(setB) - intersection
	if union == 0 {
		// Defensive: both sets are empty. The len(tokens*)==0 guards
		// above already cover this, but keep the explicit fall-through
		// to avoid a 0/0 NaN if invariants change.
		return 1.0
	}
	// Single division on integer-derived float64 values. Both numerator
	// and denominator fit exactly in float64 for any input where the
	// total distinct-token count is below 2^53 (~9e15) — well above any
	// realistic input. IEEE-754 correctly-rounded division produces
	// byte-identical output across all four CI platforms (DET-06).
	return float64(intersection) / float64(union)
}

// tokensToSet returns a SET of the tokens (deduplicated). The
// map[string]struct{} value type is the canonical zero-byte set marker
// in Go. Pre-sizing the map with len(tokens) avoids growth allocs on
// short identifier-style inputs. Empty input is supported (returns an
// empty map) but the caller has already gated empty token slices via the
// post-Tokenise guards in TokenJaccardScore — this helper is only
// reached for non-empty token slices in the score path.
func tokensToSet(tokens []string) map[string]struct{} {
	set := make(map[string]struct{}, len(tokens))
	for _, t := range tokens {
		set[t] = struct{}{}
	}
	return set
}

// setIntersectionCardinality returns |A ∩ B| computed by iterating the
// SMALLER set and probing the larger set. DET-03 satisfied: the output
// is a scalar int (associative integer addition; map iteration order
// does not affect the result). Walking the smaller side keeps the
// lookup count to len(min(setA, setB)). The selection of "smaller side"
// uses a simple len-comparison swap so the helper is deterministic in
// its arithmetic regardless of which argument is the smaller set on a
// given call.
func setIntersectionCardinality(setA, setB map[string]struct{}) int {
	small, large := setA, setB
	if len(setB) < len(setA) {
		small, large = setB, setA
	}
	var intersection int
	for k := range small {
		if _, ok := large[k]; ok {
			intersection++
		}
	}
	return intersection
}
