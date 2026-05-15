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

// monge_elkan.go implements the Monge-Elkan similarity for the fuzzymatch
// catalogue. Monge-Elkan is the only Phase 6 algorithm that takes a
// pluggable INNER metric (specified as an AlgoID): it tokenises each side
// and computes, for every token in A, the maximum inner-metric similarity
// against every token in B, then averages those per-token maxima. With a
// fixed inner metric the function is direction-sensitive — swapping the
// inputs generally yields a different score because the per-token-max
// reduction is taken over A's tokens, not B's.
//
// Two public surfaces ship:
//
//   - MongeElkanScore           — the asymmetric direct surface
//   - MongeElkanScoreSymmetric  — the arithmetic mean of MongeElkanScore
//                                 in the two directions; symmetric in
//                                 (a, b) by construction
//
// The dispatch slot AlgoMongeElkan wraps MongeElkanScoreSymmetric with
// inner = AlgoJaroWinkler and DefaultNormalisationOptions() per
// 06-CONTEXT.md §4 LOCKED — the dispatch surface is the symmetric variant
// so AlgoMongeElkan participates in the standard symmetric property-test
// set without exemption. Consumers needing genuine asymmetric Monge-Elkan
// semantics call MongeElkanScore directly or, in Phase 8, use the Scorer
// option WithMongeElkanAlgorithm(weight, inner) which forwards the inner
// AlgoID to MongeElkanScore.
//
// Source: Monge, A. E., & Elkan, C. P. (1996). "The field matching
// problem: algorithms and applications." Proceedings of the Second
// International Conference on Knowledge Discovery and Data Mining
// (KDD'96): 267-270, §3 — the "Smith-Waterman-based field matching"
// algorithm. The per-token-max-mean construction is the canonical form
// adopted by SecondString (Cohen, Ravikumar & Fienberg 2003) and the
// modern Python "py_stringmatching" / Java "SecondString" libraries.
// Equation: for strings A and B with token vectors tA = tokens(A),
// tB = tokens(B), and inner similarity function sim:
//
//	ME(A, B, sim) = (1 / |tA|) · Σ_{i=1..|tA|} max_{j=1..|tB|} sim(tA[i], tB[j])
//
// The function is direction-sensitive because the OUTER sum is over A's
// tokens and the INNER max is over B's tokens — swapping arguments swaps
// the reduction order, producing a different value when |tA| ≠ |tB| or
// when the inner-metric matrix is not symmetric across token positions.
//
// The symmetric variant is the arithmetic mean of the two directions:
//
//	ME_sym(A, B, sim) = (ME(A, B, sim) + ME(B, A, sim)) / 2.0
//
// which is invariant under (a, b) swap (the sum of two terms swapped is
// the same sum) — this is the LOCKED default per CONTEXT.md §4.
//
// Algorithm — MongeElkanScore(a, b, inner, opts):
//
//  1. Allow-list gate: if !permittedMongeElkanInner[inner] panic with
//     the documented message (per CONTEXT.md §3 LOCKED — Pitfall 4
//     self-recursion gate).
//  2. Identity short-circuit: if a == b, return 1.0 directly (saves
//     Tokenise allocations on identical inputs; ME(x, x, sim) is always
//     1.0 for any well-behaved sim — every token's max-against-self
//     pair is sim(t, t) = 1.0).
//  3. Tokenise both sides using DefaultTokeniseOptions() (Lowercase: true,
//     SplitCamelCase: true, SplitConsecutiveUpper: true, SeparatorChars:
//     "_-.:/ \t\n\r"). See the tokeniser-divergence note below.
//  4. Both-Tokenised-empty guard: return 1.0 (vacuous match — STANDARD
//     catalogue convention, mirrors TokenSortRatio / TokenJaccard).
//  5. One-Tokenised-empty guard: return 0.0.
//  6. Look up the inner-metric function via dispatch[inner] (safe: the
//     allow-list above guarantees inner ∈ permittedMongeElkanInner AND
//     each of the 14 dispatch slots has been registered by the time
//     Phase 6 plan 06-05 lands).
//  7. Outer loop over tA: for each token tokA, take the max sim(tokA, tokB)
//     over all tokB in tB; accumulate into sumOfMax.
//  8. Return sumOfMax / float64(len(tA)) — single division on
//     float-derived value.
//
// Algorithm — MongeElkanScoreSymmetric(a, b, inner, opts):
//
//	return (MongeElkanScore(a, b, inner, opts) +
//	        MongeElkanScore(b, a, inner, opts)) / 2.0
//
// The allow-list panic surfaces from the FIRST call (a, b) before the
// second call runs — invalid inner triggers the panic exactly once,
// regardless of which direction surfaces it.
//
// Conventions (mirror the Phase 5/6 short-circuit pattern; STANDARD
// catalogue convention — distinct from TokenSetRatio's RapidFuzz #110
// deviation):
//
//   - both-empty                → 1.0  (covered by a == b short-circuit
//                                       AND post-Tokenise both-empty guard)
//   - identical                 → 1.0  (a == b short-circuit; inner irrelevant)
//   - one-empty                 → 0.0
//   - both-tokens-pure-separators → 1.0 (post-Tokenise both-empty guard)
//
// Tokeniser divergence (OQ-1 — same as TokenSortRatio / TokenJaccard):
// fuzzymatch's Tokenise is identifier-aware (camelCase / snake_case /
// kebab-case / dot-case + lowercasing). The reference vectors RV-ME1..
// RV-ME6 use whitespace-only lowercase ASCII inputs so the per-token
// derivation is unambiguous; for mixed identifier-style inputs the
// project tokenisation produces semantically richer splits.
//
// Inner-metric allow-list (14 entries — OQ-4 RESOLUTION LOCKED 2026-05-15
// includes AlgoRatcliffObershelp as a character-tier algorithm):
//
//   - AlgoLevenshtein              — Levenshtein 1965
//   - AlgoDamerauLevenshteinOSA    — Damerau 1964 OSA variant (Boytsov 2011)
//   - AlgoDamerauLevenshteinFull   — Lowrance-Wagner 1975
//   - AlgoHamming                  — Hamming 1950
//   - AlgoJaro                     — Jaro 1989
//   - AlgoJaroWinkler              — Winkler 1990 (DEFAULT inner per
//                                    CONTEXT §4 LOCKED)
//   - AlgoStrcmp95                 — Winkler 1994
//   - AlgoSmithWatermanGotoh       — Smith-Waterman 1981 + Gotoh 1982
//   - AlgoLCSStr                   — Wagner-Fischer 1974 (substring variant)
//   - AlgoQGramJaccard             — Jaccard 1912 over q-grams (Ukkonen 1992)
//   - AlgoSorensenDice             — Sørensen 1948 / Dice 1945 over q-grams
//   - AlgoCosine                   — Salton & McGill 1983 over q-grams
//   - AlgoTversky                  — Tversky 1977 over q-grams
//   - AlgoRatcliffObershelp        — Ratcliff & Metzener 1988 (OQ-4 LOCKED)
//
// EXPLICITLY NOT permitted (verified by exhaustive panic test
// TestMongeElkan_PanicsOnNonPermittedInner — walks all 23 AlgoIDs and
// asserts the 9 rejected ones panic with the documented message):
//
//   - AlgoMongeElkan: self-recursion would cause infinite recursion
//     (RESEARCH.md Pitfall 4). The dispatch wrapper for AlgoMongeElkan
//     binds an inner ≠ AlgoMongeElkan; direct calls with inner =
//     AlgoMongeElkan are programmer-error and surface as panic per
//     CONTEXT §3 LOCKED.
//   - AlgoTokenSortRatio / AlgoTokenSetRatio / AlgoPartialRatio /
//     AlgoTokenJaccard: token-on-token is meaningless (the inner metric
//     receives SINGLE tokens from the outer Tokenise step; re-tokenising
//     a single token is a no-op or identity-equivalent at best, and
//     recursive in spirit — these algorithms are themselves token-tier
//     compositions).
//   - Phase 7 phonetic (AlgoSoundex / AlgoDoubleMetaphone / AlgoNYSIIS /
//     AlgoMRA): these are reserved for Phase 7's ADDITIVE allow-list
//     expansion. When Phase 7 lands, planners ADD 4 entries here AND
//     update the panic-test fixture (rejected: 9 → 5; permitted: 14 → 18).
//
// Direct-call validation (CONTEXT.md §3 LOCKED):
//
//   - Non-permitted inner panics with the message
//     "fuzzymatch: AlgoID <name> not permitted as Monge-Elkan inner metric"
//     where <name> is the canonical AlgoID String() output (e.g.
//     "MongeElkan", "TokenSortRatio", "Soundex"). The exact message
//     format is pinned by TestMongeElkan_PanicMessageFormat to gate
//     against regressions.
//   - The Phase 8 Scorer option WithMongeElkanAlgorithm(weight, inner)
//     returns ErrInvalidAlgoID (declared in errors.go Phase 1) for the
//     same input — direct-call panic discipline applies only to the
//     direct surface where programmer error must fail loudly.
//
// Asymmetry-discriminating reference vector pair (RV-ME6 from
// RESEARCH.md §"Specific Ideas" — load-bearing regression gate):
//
//	MongeElkanScore("alpha", "alpha beta gamma", AlgoLevenshtein, opts) =
//	  tokens(A) = ["alpha"]; tokens(B) = ["alpha","beta","gamma"]
//	  max_inner(alpha, *) = max(Lev(alpha,alpha)=1, Lev(alpha,beta)=0.2,
//	                            Lev(alpha,gamma)=0.2) = 1.0
//	  ME(A, B) = 1.0 / 1 = 1.0
//
//	MongeElkanScore("alpha beta gamma", "alpha", AlgoLevenshtein, opts) =
//	  tokens(A) = ["alpha","beta","gamma"]; tokens(B) = ["alpha"]
//	  max_inner(alpha, *) = Lev(alpha,alpha) = 1.0
//	  max_inner(beta,  *) = Lev(beta,alpha)  = 0.2
//	  max_inner(gamma, *) = Lev(gamma,alpha) = 0.2
//	  ME(B, A) = (1.0 + 0.2 + 0.2) / 3 = 0.466666…
//
//	1.0 ≠ 0.466666… — the input swap with the same inner produces a
//	different score. The asymmetric direct-call surface preserves this
//	direction-sensitivity; MongeElkanScoreSymmetric averages the two
//	directions to (1.0 + 0.466666…) / 2 = 0.733333… which is symmetric.
//	The TestMongeElkanScore RV-ME6 row, the
//	TestProp_MongeElkanScore_AsymmetricWhenTokenCountAsymmetric property
//	test, AND the BDD asymmetry scenario together form the three-layer
//	defence against direction-aggregation regressions.
//
// Source-origin discipline (per algorithm-licensing-standards):
//
//   - Primary source:        Monge & Elkan 1996 KDD'96 §3.
//   - Cross-validation:      NONE — hand-derived RV-ME1..RV-ME6 reference
//                            vectors in monge_elkan_test.go per
//                            CONTEXT.md §1b LOCKED. The RapidFuzz
//                            cross-validation corpus does NOT include
//                            Monge-Elkan entries (RapidFuzz's default
//                            inner metric may not match this project's
//                            JaroWinkler default; the corpus is for the
//                            four Indel-based ratios only).
//   - Tie-break:             none (Monge-Elkan is unambiguous given a
//                            fixed inner metric).
//   - GPL/LGPL provenance:   none.
//   - Code copied verbatim:  none.
//
// Implementation discipline:
//
//   - NO init()-time table builds (per docs/requirements.md §5(12) and
//     determinism-standards DET-13). The permittedMongeElkanInner
//     allow-list is declared at PACKAGE SCOPE as a map literal with
//     compile-time initialiser — no init() function.
//   - NO map iteration on output paths (DET-03). The outer/inner loops
//     iterate token SLICES (deterministically ordered by Tokenise);
//     the allow-list map is only LOOKED UP (boolean membership test),
//     never iterated.
//   - NO transcendental float operations (DET-06): only float64 max
//     comparisons, additive accumulation, and a single final division
//     with explicit parenthesisation per DET-06.
//   - Identity short-circuit `if a == b { return 1.0 }` AFTER the
//     allow-list gate (so invalid inner panics even on identical
//     inputs — programmer error fails loudly per CONTEXT §3) but
//     BEFORE Tokenise (avoids the make([]string, 0, 4) allocation).
//   - Left-to-right reduction with explicit parenthesisation per DET-06:
//     `sumOfMax / float64(len(tokensA))` for asymmetric;
//     `(asymmetricAB + asymmetricBA) / 2.0` for symmetric.
//   - Inner-loop max uses explicit if/else (NOT builtin max) per the
//     project canonical pattern for determinism-reviewer auditability
//     (same precedent as q_gram.go's tverskyFromQGramMaps).
//
// Public surface (TWO functions per CONTEXT §4 LOCKED):
//
//   - MongeElkanScore(a, b string, inner AlgoID, opts NormalisationOptions) float64
//   - MongeElkanScoreSymmetric(a, b string, inner AlgoID, opts NormalisationOptions) float64
//
// Only the dispatch slot AlgoMongeElkan is registered (via
// dispatch_monge_elkan.go binding the symmetric variant + AlgoJaroWinkler
// + DefaultNormalisationOptions per CONTEXT §4 LOCKED). Phase 8's
// `WithMongeElkanAlgorithm(weight, inner)` Scorer option will forward the
// user-supplied inner AlgoID; Phase 7 ADDS phonetic AlgoIDs to the
// allow-list ADDITIVELY (4 new entries → 18 total).
//
// NormalisationOptions parameter: the opts argument is accepted for
// forward-compatibility with the Phase 8 Scorer option
// WithMongeElkanAlgorithm; MongeElkanScore itself does NOT call
// Normalise — Tokenise(s, DefaultTokeniseOptions()) lowercases via
// TokeniseOptions.Lowercase = true. Pre-Normalise inputs explicitly if
// you want NFC / diacritic stripping before tokenisation.
//
// Complexity:
//
//	O(|tA|·|tB|·cost(inner))
//
//	where |tA|, |tB| are the post-tokenisation token counts and
//	cost(inner) is the inner metric's per-comparison complexity
//	(e.g. Jaro-Winkler is O(min(m,n)) per token comparison).
//
// DoS notice:
//
//	On inputs with > 1,000 tokens per side this performs ~10^6 inner-
//	metric comparisons. With Jaro-Winkler inner (the default), each
//	comparison is O(token_length) — total cost on 1000-token inputs
//	approximates 10^7 character operations. In untrusted-input
//	contexts (HTTP request body, file uploads, user-submitted
//	identifiers), pre-validate token-count ceilings before calling.
//	See BenchmarkMongeElkan_Pathological_1000Tokens for measured
//	timings on 1000×1000-token inputs.

package fuzzymatch

// permittedMongeElkanInner enumerates the AlgoIDs valid as Monge-Elkan
// inner metrics. Declared at PACKAGE SCOPE (per DET-13 / Phase 5 §5
// LOCKED — NO init()-time table builds). Phase 7 ADDITIVELY adds phonetic
// AlgoIDs (1 per plan: 07-01 adds AlgoSoundex; 07-02..07-04 add the rest).
// The allow-list is the single source of truth — when a Phase 7 plan lands,
// it ADDS its entry here AND updates the panic-test fixture in
// monge_elkan_test.go in the SAME COMMIT (per CONTEXT.md §4 LOCKED).
//
// 16 entries (9 character-tier + 4 q-gram tier + 1 gestalt + 2 phonetic-tier
// — plan 07-01 adds AlgoSoundex; plan 07-02 adds AlgoDoubleMetaphone;
// plans 07-03..07-04 add NYSIIS/MRA additively → 18 total after Phase 7 completes).
//
// EXPLICITLY NOT permitted (verified by exhaustive panic test):
//
//   - AlgoMongeElkan: self-recursion infinite loop (RESEARCH.md Pitfall 4)
//   - AlgoTokenSortRatio / AlgoTokenSetRatio / AlgoPartialRatio /
//     AlgoTokenJaccard: token-on-token meaningless (the inner metric
//     receives single tokens from the outer Tokenise; re-tokenising
//     single tokens is a no-op / identity-equivalent at best, recursive
//     at worst)
//   - Phase 7 phonetic remaining (AlgoNYSIIS / AlgoMRA): added in
//     Phase 7 plans 07-03..07-04 ADDITIVELY
var permittedMongeElkanInner = map[AlgoID]bool{
	// Character tier (9):
	AlgoLevenshtein:            true, // Levenshtein 1965
	AlgoDamerauLevenshteinOSA:  true, // Damerau 1964 OSA variant (Boytsov 2011)
	AlgoDamerauLevenshteinFull: true, // Lowrance-Wagner 1975
	AlgoHamming:                true, // Hamming 1950
	AlgoJaro:                   true, // Jaro 1989
	AlgoJaroWinkler:            true, // Winkler 1990 (DEFAULT inner)
	AlgoStrcmp95:               true, // Winkler 1994
	AlgoSmithWatermanGotoh:     true, // Smith-Waterman 1981 + Gotoh 1982
	AlgoLCSStr:                 true, // Wagner-Fischer 1974 (substring variant)

	// Q-gram tier (4):
	AlgoQGramJaccard: true, // Jaccard 1912 / Ukkonen 1992
	AlgoSorensenDice: true, // Sørensen 1948 / Dice 1945
	AlgoCosine:       true, // Salton & McGill 1983
	AlgoTversky:      true, // Tversky 1977

	// Gestalt tier (1) — OQ-4 RESOLUTION LOCKED 2026-05-15:
	AlgoRatcliffObershelp: true, // Ratcliff & Metzener 1988 — character-tier per OQ-4

	// Phonetic tier (Phase 7) — additive per CONTEXT.md §4 LOCKED:
	AlgoSoundex:         true, // Russell 1918 / Knuth TAOCP §6.4 — plan 07-01
	AlgoDoubleMetaphone: true, // Philips 2000 — plan 07-02
}

// MongeElkanScore returns the asymmetric Monge-Elkan similarity between
// a and b under the given inner metric: tokenise both sides, then
// compute
//
//	ME(A, B, sim) = (1 / |tA|) · Σ_{i=1..|tA|} max_{j=1..|tB|} sim(tA[i], tB[j])
//
// where sim is the inner-metric function bound to inner via the
// dispatch table. Returns a value in [0.0, 1.0].
//
// The function is direction-sensitive: MongeElkanScore(a, b, inner, opts)
// generally ≠ MongeElkanScore(b, a, inner, opts) when |tokens(a)| ≠
// |tokens(b)| or when the inner-metric matrix is asymmetric across the
// token positions. For a symmetric variant call MongeElkanScoreSymmetric.
//
// The inner AlgoID MUST be one of the 14 permitted inner metrics (see
// permittedMongeElkanInner in monge_elkan.go). Passing AlgoMongeElkan
// (self-reference), any token-tier AlgoID (AlgoTokenSortRatio /
// AlgoTokenSetRatio / AlgoPartialRatio / AlgoTokenJaccard), or any
// phonetic AlgoID (AlgoSoundex / AlgoDoubleMetaphone / AlgoNYSIIS /
// AlgoMRA — reserved for Phase 7's additive allow-list expansion) panics
// with the message
//
//	"fuzzymatch: AlgoID <name> not permitted as Monge-Elkan inner metric"
//
// where <name> is the canonical String() output of the AlgoID. The
// Phase 8 Scorer option WithMongeElkanAlgorithm returns ErrInvalidAlgoID
// instead — direct-call panic discipline per CONTEXT.md §3 LOCKED.
//
// The opts NormalisationOptions parameter is accepted for
// forward-compatibility with the Phase 8 Scorer option
// WithMongeElkanAlgorithm; MongeElkanScore itself does NOT call
// Normalise — Tokenise(s, DefaultTokeniseOptions()) lowercases via
// TokeniseOptions.Lowercase = true. Pre-Normalise inputs explicitly if
// you want NFC / diacritic stripping before tokenisation.
//
// Conventions (STANDARD catalogue both-empty → 1.0; distinct from
// TokenSetRatio's LOCKED RapidFuzz issue #110 deviation):
//
//   - MongeElkanScore("",        "",        inner, opts) == 1.0  (both-empty)
//   - MongeElkanScore("hello",   "hello",   inner, opts) == 1.0  (identity)
//   - MongeElkanScore("",        "hello",   inner, opts) == 0.0  (one-empty)
//   - MongeElkanScore("hello",   "",        inner, opts) == 0.0  (one-empty)
//
// Reference vector (RV-ME1 — canonical asymmetric example):
//
//	MongeElkanScore("user create", "usr creating", AlgoJaroWinkler, opts) ≈ 0.9125
//	  tokens(A) = ["user","create"]; tokens(B) = ["usr","creating"]
//	  max_inner(user,   *) = max(JW(user,usr)=0.9333, JW(user,creating)=0.4167) = 0.9333
//	  max_inner(create, *) = max(JW(create,usr)=0.5, JW(create,creating)=0.8917) = 0.8917
//	  ME(A, B) = (0.9333 + 0.8917) / 2 ≈ 0.9125
//
// Reference vector (RV-ME6 — KEYSTONE asymmetry gate):
//
//	MongeElkanScore("alpha", "alpha beta gamma", AlgoLevenshtein, opts) = 1.0
//	MongeElkanScore("alpha beta gamma", "alpha", AlgoLevenshtein, opts) ≈ 0.466666…
//	  → 1.0 ≠ 0.466666… — the input swap with the same inner produces
//	    direction-sensitive scores (load-bearing regression gate).
//
// See the file-level godoc for the full inner-metric allow-list, the
// DoS notice on 1000-token inputs, and the source-origin discipline.
func MongeElkanScore(a, b string, inner AlgoID, opts NormalisationOptions) float64 {
	// Allow-list gate — per CONTEXT §3 LOCKED + RESEARCH.md Pitfall 4.
	// Fires FIRST so invalid inner panics even on identical inputs
	// (programmer error fails loudly).
	if !permittedMongeElkanInner[inner] {
		panic("fuzzymatch: AlgoID " + inner.String() + " not permitted as Monge-Elkan inner metric")
	}
	// Identity short-circuit — avoids Tokenise allocations. ME(x, x, sim)
	// = 1.0 for any well-behaved sim because every token's max-against-
	// self pair is sim(t, t) = 1.0 and the average of |tA| ones is 1.0.
	if a == b {
		return 1.0
	}
	// opts is intentionally unused in the function body — accepted for
	// Phase 8 Scorer compatibility; Tokenise's Lowercase handles
	// case-folding at the tokeniser layer.
	_ = opts
	tokensA := Tokenise(a, DefaultTokeniseOptions())
	tokensB := Tokenise(b, DefaultTokeniseOptions())
	// Both-Tokenised-empty: vacuous match (STANDARD catalogue
	// convention; mirrors TokenSortRatio / TokenJaccard — distinct from
	// TokenSetRatio's RapidFuzz #110 deviation).
	if len(tokensA) == 0 && len(tokensB) == 0 {
		return 1.0
	}
	// One-Tokenised-empty: 0.0.
	if len(tokensA) == 0 || len(tokensB) == 0 {
		return 0.0
	}
	// Inner-metric function lookup via dispatch. Safe: the allow-list
	// above guarantees inner ∈ permittedMongeElkanInner AND each of the
	// 14 permitted dispatch slots has been registered by package load
	// time (the var _ = func() bool { ... }() idiom in each dispatch_*.go
	// file fires before any consumer can read dispatch).
	innerFn := dispatch[inner]
	// Outer reduction: for each token in tA, take max sim against tB;
	// accumulate the maxima for averaging. The double loop is the
	// canonical Monge & Elkan 1996 §3 form. Explicit if/else for max
	// per the project canonical pattern (NOT builtin max — keeps
	// determinism-reviewer audit trail uniform across phases).
	var sumOfMax float64
	for _, tokA := range tokensA {
		var maxSim float64
		for _, tokB := range tokensB {
			s := innerFn(tokA, tokB)
			if s > maxSim {
				maxSim = s
			}
		}
		sumOfMax += maxSim
	}
	// Single division on integer-derived float64. IEEE-754 correctly-
	// rounded division produces byte-identical output across all four
	// CI platforms (DET-06).
	return sumOfMax / float64(len(tokensA))
}

// MongeElkanScoreSymmetric returns the SYMMETRIC Monge-Elkan similarity
// between a and b — the arithmetic mean of the two directional
// MongeElkanScore values:
//
//	ME_sym(A, B, sim) = (ME(A, B, sim) + ME(B, A, sim)) / 2.0
//
// This is symmetric in (a, b) by construction (the sum of two terms
// swapped under (a, b) → (b, a) is the same sum) and is the surface
// bound to dispatch[AlgoMongeElkan] per CONTEXT.md §4 LOCKED, so
// AlgoMongeElkan participates in the standard PropAlgorithmScore_Symmetric
// property test set without exemption.
//
// The inner AlgoID is subject to the same 14-entry allow-list as
// MongeElkanScore — invalid inner panics with the same message format
// per CONTEXT §3 LOCKED. The panic surfaces from the FIRST
// MongeElkanScore call below (the (a, b) direction); the (b, a) call
// never runs on the panic path.
//
// The opts NormalisationOptions parameter is accepted for
// forward-compatibility with the Phase 8 Scorer option; the actual
// Tokenise step uses DefaultTokeniseOptions() internally per CONTEXT §4.
//
// Conventions: same as MongeElkanScore (both-empty → 1.0; one-empty →
// 0.0; identity → 1.0). The symmetric average of identical short-circuit
// returns is still 1.0; the symmetric average of (0.0, 0.0) is 0.0.
//
// Reference vector (symmetric average of RV-ME6 asymmetric pair):
//
//	MongeElkanScoreSymmetric("alpha", "alpha beta gamma", AlgoLevenshtein, opts) ≈ 0.733333…
//	  ME(A, B) = 1.0; ME(B, A) = 0.466666…; mean = 0.733333…
//
// See the file-level godoc for the inner-metric allow-list, DoS notice,
// and source-origin discipline.
func MongeElkanScoreSymmetric(a, b string, inner AlgoID, opts NormalisationOptions) float64 {
	// The allow-list panic surfaces from the FIRST MongeElkanScore call
	// — invalid inner triggers the panic exactly once per call.
	// Explicit parenthesisation per DET-06; the divide-by-2 on a sum is
	// exact in IEEE-754 for any finite operands (the float64 division
	// by 2.0 is bitwise-exact).
	return (MongeElkanScore(a, b, inner, opts) + MongeElkanScore(b, a, inner, opts)) / 2.0
}
