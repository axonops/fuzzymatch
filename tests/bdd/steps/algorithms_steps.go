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

// Package steps contains the godog step definitions for Phase 2's six
// character-based algorithms. The AlgorithmContext struct holds state between
// steps within a scenario. Step functions return error rather than calling
// t.Errorf, as required by the godog API.
//
// Wave 2 plans (02-02 through 02-06) extend this file by:
//   - Adding their algorithm-specific step functions to AlgorithmContext.
//   - Registering those step regexes in InitializeScenario.
//   - Adding any additional state fields to AlgorithmContext.
//
// testify is available in tests/bdd/ (isolated sub-module) but is not required
// for this initial harness; step functions return error directly.

package steps

import (
	"fmt"
	"math"
	"strings"

	"github.com/cucumber/godog"

	"github.com/axonops/fuzzymatch"
)

// AlgorithmContext holds state between BDD steps within a scenario. Each
// scenario instantiates a fresh AlgorithmContext to ensure isolation.
//
// Wave 2 plans may add additional state fields if their step patterns require
// more than two scores. The two-score design covers the symmetry scenario
// ("compute A and compute B, then assert both equal").
type AlgorithmContext struct {
	lastScore    float64 // populated by "I compute the Xxx score between" steps
	lastScore2   float64 // populated by "I compute the second Xxx score between" steps
	lastDistance int     // populated by "I compute the Xxx distance between" steps (plan 02-02+)
}

// iComputeTheLevenshteinScoreBetween computes LevenshteinScore(a, b) and
// stores the result in lastScore.
func (ctx *AlgorithmContext) iComputeTheLevenshteinScoreBetween(a, b string) error {
	ctx.lastScore = fuzzymatch.LevenshteinScore(a, b)
	return nil
}

// iComputeTheSecondLevenshteinScoreBetween computes LevenshteinScore(a, b) and
// stores the result in lastScore2. Used by the symmetry scenario to capture
// a second score for comparison.
func (ctx *AlgorithmContext) iComputeTheSecondLevenshteinScoreBetween(a, b string) error {
	ctx.lastScore2 = fuzzymatch.LevenshteinScore(a, b)
	return nil
}

// theScoreShouldBeApproximately asserts that lastScore is within tolerance of
// expected. Uses absolute difference (math.Abs) to avoid sign-sensitive
// comparisons.
func (ctx *AlgorithmContext) theScoreShouldBeApproximately(expected, tolerance float64) error {
	if math.Abs(ctx.lastScore-expected) > tolerance {
		return fmt.Errorf("expected score %f ± %f, got %f", expected, tolerance, ctx.lastScore)
	}
	return nil
}

// theScoreShouldBeExactly asserts that lastScore equals expected exactly.
// Used for edge cases (both-empty = 1.0, one-empty = 0.0) where the result
// must be the exact IEEE-754 value without floating-point rounding.
func (ctx *AlgorithmContext) theScoreShouldBeExactly(expected float64) error {
	if ctx.lastScore != expected {
		return fmt.Errorf("expected score exactly %f, got %f", expected, ctx.lastScore)
	}
	return nil
}

// bothLevenshteinScoresShouldBeEqual asserts that lastScore == lastScore2. Used
// by the symmetry scenario after computing Score(a,b) and Score(b,a).
func (ctx *AlgorithmContext) bothLevenshteinScoresShouldBeEqual() error {
	if ctx.lastScore != ctx.lastScore2 {
		return fmt.Errorf("scores not equal: %f != %f", ctx.lastScore, ctx.lastScore2)
	}
	return nil
}

// ---------------------------------------------------------------------------
// Hamming step definitions (plan 02-02)
// ---------------------------------------------------------------------------

// iComputeTheHammingScoreBetween computes HammingScore(a, b) and stores the
// result in lastScore. Demonstrates the locked unequal-length silent-zero
// policy: Score("abc", "ab") == 0.0 silently.
func (ctx *AlgorithmContext) iComputeTheHammingScoreBetween(a, b string) error {
	ctx.lastScore = fuzzymatch.HammingScore(a, b)
	return nil
}

// iComputeTheSecondHammingScoreBetween computes HammingScore(a, b) and stores
// the result in lastScore2. Used by the symmetry scenario.
func (ctx *AlgorithmContext) iComputeTheSecondHammingScoreBetween(a, b string) error {
	ctx.lastScore2 = fuzzymatch.HammingScore(a, b)
	return nil
}

// iComputeTheHammingDistanceBetween computes HammingDistance(a, b) and stores
// the result in lastDistance. Used by the distance-equals-max-length scenario.
func (ctx *AlgorithmContext) iComputeTheHammingDistanceBetween(a, b string) error {
	ctx.lastDistance = fuzzymatch.HammingDistance(a, b)
	return nil
}

// bothHammingScoresShouldBeEqual asserts lastScore == lastScore2.
func (ctx *AlgorithmContext) bothHammingScoresShouldBeEqual() error {
	if ctx.lastScore != ctx.lastScore2 {
		return fmt.Errorf("hamming scores not equal: %f != %f", ctx.lastScore, ctx.lastScore2)
	}
	return nil
}

// theDistanceShouldBe asserts lastDistance == expected.
//
// This step is INTENTIONALLY algorithm-agnostic: it matches the value
// written by whichever *Distance* step ran most recently in the current
// scenario (HammingDistance, DamerauLevenshteinOSADistance, or
// DamerauLevenshteinFullDistance). Per-scenario AlgorithmContext isolation
// means cross-scenario bleed is impossible, but if a single scenario chains
// two distance computations the assertion applies to the LAST one. If a
// scenario ever needs to assert on a specific algorithm's distance after a
// later distance step has run, introduce an algorithm-suffixed step
// (e.g. theHammingDistanceShouldBe). Closes IN-06 from 02-REVIEW.md.
func (ctx *AlgorithmContext) theDistanceShouldBe(expected int) error {
	if ctx.lastDistance != expected {
		return fmt.Errorf("expected distance %d, got %d", expected, ctx.lastDistance)
	}
	return nil
}

// ---------------------------------------------------------------------------
// Jaro step definitions (plan 02-03)
// ---------------------------------------------------------------------------

// iComputeTheJaroScoreBetween computes JaroScore(a, b) and stores the result
// in lastScore. The Jaro formula is symmetric and operates on bytes for ASCII
// inputs (zero allocation on inputs <= 256 bytes).
func (ctx *AlgorithmContext) iComputeTheJaroScoreBetween(a, b string) error {
	ctx.lastScore = fuzzymatch.JaroScore(a, b)
	return nil
}

// iComputeTheSecondJaroScoreBetween computes JaroScore(a, b) and stores the
// result in lastScore2. Used by the symmetry scenario to capture a second score
// for comparison.
func (ctx *AlgorithmContext) iComputeTheSecondJaroScoreBetween(a, b string) error {
	ctx.lastScore2 = fuzzymatch.JaroScore(a, b)
	return nil
}

// bothJaroScoresShouldBeEqual asserts lastScore == lastScore2.
func (ctx *AlgorithmContext) bothJaroScoresShouldBeEqual() error {
	if ctx.lastScore != ctx.lastScore2 {
		return fmt.Errorf("jaro scores not equal: %f != %f", ctx.lastScore, ctx.lastScore2)
	}
	return nil
}

// ---------------------------------------------------------------------------
// Jaro-Winkler step definitions (plan 02-04)
// ---------------------------------------------------------------------------

// iComputeTheJaroWinklerScoreBetween computes JaroWinklerScore(a, b) and
// stores the result in lastScore. JaroWinklerScore delegates to JaroScore
// and applies the Winkler 1990 prefix boost when Jaro >= 0.7.
func (ctx *AlgorithmContext) iComputeTheJaroWinklerScoreBetween(a, b string) error {
	ctx.lastScore = fuzzymatch.JaroWinklerScore(a, b)
	return nil
}

// iComputeTheSecondJaroWinklerScoreBetween computes JaroWinklerScore(a, b) and
// stores the result in lastScore2. Used by the symmetry scenario to capture
// a second score for comparison.
func (ctx *AlgorithmContext) iComputeTheSecondJaroWinklerScoreBetween(a, b string) error {
	ctx.lastScore2 = fuzzymatch.JaroWinklerScore(a, b)
	return nil
}

// bothJaroWinklerScoresShouldBeEqual asserts lastScore == lastScore2.
func (ctx *AlgorithmContext) bothJaroWinklerScoresShouldBeEqual() error {
	if ctx.lastScore != ctx.lastScore2 {
		return fmt.Errorf("jaro-winkler scores not equal: %f != %f", ctx.lastScore, ctx.lastScore2)
	}
	return nil
}

// bothJaroWinklerAndJaroScoresShouldBeEqual asserts that lastScore (the
// JaroWinklerScore computed in the current scenario) equals lastScore2 (the
// JaroScore computed in the same scenario). Used by the boost-gate scenario
// to verify JW == J for below-threshold pairs.
func (ctx *AlgorithmContext) bothJaroWinklerAndJaroScoresShouldBeEqual() error {
	if ctx.lastScore != ctx.lastScore2 {
		return fmt.Errorf("JaroWinklerScore (%f) != JaroScore (%f) for below-threshold pair; boost should not be applied",
			ctx.lastScore, ctx.lastScore2)
	}
	return nil
}

// ---------------------------------------------------------------------------
// Damerau-Levenshtein OSA step definitions (plan 02-05)
// ---------------------------------------------------------------------------

// iComputeTheDamerauLevenshteinOSAScoreBetween computes
// DamerauLevenshteinOSAScore(a, b) and stores the result in lastScore.
// The ASCII fast path uses a stack-allocated three-row DP buffer for zero
// heap allocations on inputs ≤ 64 bytes.
func (ctx *AlgorithmContext) iComputeTheDamerauLevenshteinOSAScoreBetween(a, b string) error {
	ctx.lastScore = fuzzymatch.DamerauLevenshteinOSAScore(a, b)
	return nil
}

// iComputeTheSecondDamerauLevenshteinOSAScoreBetween computes
// DamerauLevenshteinOSAScore(a, b) and stores the result in lastScore2.
// Used by the symmetry scenario to capture a second score for comparison.
func (ctx *AlgorithmContext) iComputeTheSecondDamerauLevenshteinOSAScoreBetween(a, b string) error {
	ctx.lastScore2 = fuzzymatch.DamerauLevenshteinOSAScore(a, b)
	return nil
}

// iComputeTheDamerauLevenshteinOSADistanceBetween computes
// DamerauLevenshteinOSADistance(a, b) and stores the result in lastDistance.
// Used by the discriminating-vector scenario to gate the distance == 3 contract.
func (ctx *AlgorithmContext) iComputeTheDamerauLevenshteinOSADistanceBetween(a, b string) error {
	ctx.lastDistance = fuzzymatch.DamerauLevenshteinOSADistance(a, b)
	return nil
}

// bothDamerauLevenshteinOSAScoresShouldBeEqual asserts lastScore == lastScore2.
func (ctx *AlgorithmContext) bothDamerauLevenshteinOSAScoresShouldBeEqual() error {
	if ctx.lastScore != ctx.lastScore2 {
		return fmt.Errorf("damerau OSA scores not equal: %f != %f", ctx.lastScore, ctx.lastScore2)
	}
	return nil
}

// ---------------------------------------------------------------------------
// Damerau-Levenshtein Full step definitions (plan 02-06)
// ---------------------------------------------------------------------------

// iComputeTheDamerauLevenshteinFullScoreBetween computes
// DamerauLevenshteinFullScore(a, b) and stores the result in lastScore.
// The full DP table is heap-allocated for all inputs (v1.0 implementation;
// see damerau_full.go for the v1.x two-row optimisation plan).
func (ctx *AlgorithmContext) iComputeTheDamerauLevenshteinFullScoreBetween(a, b string) error {
	ctx.lastScore = fuzzymatch.DamerauLevenshteinFullScore(a, b)
	return nil
}

// iComputeTheSecondDamerauLevenshteinFullScoreBetween computes
// DamerauLevenshteinFullScore(a, b) and stores the result in lastScore2.
// Used by the symmetry scenario to capture a second score for comparison.
func (ctx *AlgorithmContext) iComputeTheSecondDamerauLevenshteinFullScoreBetween(a, b string) error {
	ctx.lastScore2 = fuzzymatch.DamerauLevenshteinFullScore(a, b)
	return nil
}

// iComputeTheDamerauLevenshteinFullDistanceBetween computes
// DamerauLevenshteinFullDistance(a, b) and stores the result in lastDistance.
// Used by the discriminating-vector scenario to gate the distance == 2 contract.
func (ctx *AlgorithmContext) iComputeTheDamerauLevenshteinFullDistanceBetween(a, b string) error {
	ctx.lastDistance = fuzzymatch.DamerauLevenshteinFullDistance(a, b)
	return nil
}

// bothDamerauLevenshteinFullScoresShouldBeEqual asserts lastScore == lastScore2.
func (ctx *AlgorithmContext) bothDamerauLevenshteinFullScoresShouldBeEqual() error {
	if ctx.lastScore != ctx.lastScore2 {
		return fmt.Errorf("damerau Full scores not equal: %f != %f", ctx.lastScore, ctx.lastScore2)
	}
	return nil
}

// ---------------------------------------------------------------------------
// Smith-Waterman-Gotoh step definitions (plan 03-01)
// ---------------------------------------------------------------------------

// iComputeTheSmithWatermanGotohScoreBetween computes
// SmithWatermanGotohScore(a, b) and stores the result in lastScore.
// SWG returns the local-alignment similarity in [0, 1] using the documented
// default params (Match=1.0, Mismatch=-1.0, GapOpen=-1.5, GapExtend=-0.5).
func (ctx *AlgorithmContext) iComputeTheSmithWatermanGotohScoreBetween(a, b string) error {
	ctx.lastScore = fuzzymatch.SmithWatermanGotohScore(a, b)
	return nil
}

// iComputeTheSecondSmithWatermanGotohScoreBetween computes
// SmithWatermanGotohScore(a, b) and stores the result in lastScore2. Used by
// symmetry / gap-split-canary scenarios to capture a second score.
func (ctx *AlgorithmContext) iComputeTheSecondSmithWatermanGotohScoreBetween(a, b string) error {
	ctx.lastScore2 = fuzzymatch.SmithWatermanGotohScore(a, b)
	return nil
}

// bothSmithWatermanGotohScoresShouldBeEqual asserts lastScore == lastScore2.
func (ctx *AlgorithmContext) bothSmithWatermanGotohScoresShouldBeEqual() error {
	if ctx.lastScore != ctx.lastScore2 {
		return fmt.Errorf("swg scores not equal: %f != %f", ctx.lastScore, ctx.lastScore2)
	}
	return nil
}

// ---------------------------------------------------------------------------
// Strcmp95 step definitions (plan 04-01)
// ---------------------------------------------------------------------------

// iComputeTheStrcmp95ScoreBetween computes Strcmp95Score(a, b) and stores the
// result in lastScore. Strcmp95 layers four Winkler 1994 adjustments atop
// Jaro: similar-character credit, prefix boost, long-string adjustment.
// ASCII-only; no Runes variant per CONTEXT.md §2.
func (ctx *AlgorithmContext) iComputeTheStrcmp95ScoreBetween(a, b string) error {
	ctx.lastScore = fuzzymatch.Strcmp95Score(a, b)
	return nil
}

// iComputeTheSecondStrcmp95ScoreBetween computes Strcmp95Score(a, b) and
// stores the result in lastScore2. Used by symmetry scenarios to capture a
// second score for comparison.
func (ctx *AlgorithmContext) iComputeTheSecondStrcmp95ScoreBetween(a, b string) error {
	ctx.lastScore2 = fuzzymatch.Strcmp95Score(a, b)
	return nil
}

// bothStrcmp95ScoresShouldBeEqual asserts lastScore == lastScore2.
func (ctx *AlgorithmContext) bothStrcmp95ScoresShouldBeEqual() error {
	if ctx.lastScore != ctx.lastScore2 {
		return fmt.Errorf("strcmp95 scores not equal: %f != %f", ctx.lastScore, ctx.lastScore2)
	}
	return nil
}

// ---------------------------------------------------------------------------
// Ratcliff-Obershelp step definitions (plan 04-03)
//
// Only the single Score step is needed — the symmetry scenario is OMITTED
// per OQ-1 resolution (LOCKED 2026-05-14), so no "second" / "equal" steps
// exist. The asymmetric-by-design semantics are verified by unit tests
// (TestRatcliffObershelp_AsymmetryPin) and cross-algorithm consistency
// tests in plan 04-05.
// ---------------------------------------------------------------------------

// roAutojunkA and roAutojunkB are the 205-char constructed inputs used by
// the autojunk-sensitive BDD scenario. Constructed as:
//
//	roAutojunkA = "a"*100 + "x"*5 + "a"*100   (205 chars)
//	roAutojunkB = "a"*50  + "y"*5 + "a"*150   (205 chars)
//
// These trigger Python difflib's autojunk heuristic when len(b) >= 200 and
// a character appears in >= 1% of positions — both true here. Pinning the
// expected score (~0.7317 from difflib(autojunk=False).ratio()) proves the
// Go implementation does NOT have an autojunk-like heuristic enabled
// (RESEARCH.md Pitfall 2 closure).
//
// Computed via strings.Repeat at package-init time so the character counts
// are arithmetic — no hand-counting required. (The values are package-
// scoped var bindings, not const, because strings.Repeat is a function call
// and therefore not a constant expression in Go.)
var (
	roAutojunkA = strings.Repeat("a", 100) + strings.Repeat("x", 5) + strings.Repeat("a", 100) // 205 chars
	roAutojunkB = strings.Repeat("a", 50) + strings.Repeat("y", 5) + strings.Repeat("a", 150)  // 205 chars
)

// iComputeTheRatcliffObershelpScoreBetween computes RatcliffObershelpScore
// (a, b) and stores the result in lastScore. Ratcliff-Obershelp is the
// difflib-equivalent (matches Python difflib.SequenceMatcher(autojunk=False).
// ratio() within 1e-9). Only the dispatched byte-path score function is
// exercised via BDD; the rune-path variant RatcliffObershelpScoreRunes is
// covered by unit tests.
func (ctx *AlgorithmContext) iComputeTheRatcliffObershelpScoreBetween(a, b string) error {
	ctx.lastScore = fuzzymatch.RatcliffObershelpScore(a, b)
	return nil
}

// iComputeTheRatcliffObershelpScoreForTheAutojunkSensitivePair computes
// RatcliffObershelpScore on the 205-char autojunk-sensitive constructed
// inputs (roAutojunkA / roAutojunkB) and stores the result in lastScore.
// Used by the 200+-char autojunk-sensitive scenario in the BDD feature
// file. The constructed pair lives as a Go constant rather than in the
// Gherkin file because 205-character literals in Examples tables are
// hard to read.
func (ctx *AlgorithmContext) iComputeTheRatcliffObershelpScoreForTheAutojunkSensitivePair() error {
	ctx.lastScore = fuzzymatch.RatcliffObershelpScore(roAutojunkA, roAutojunkB)
	return nil
}

// ---------------------------------------------------------------------------
// LCSStr step definitions (plan 04-02)
// ---------------------------------------------------------------------------

// iComputeTheLCSStrScoreBetween computes LCSStrScore(a, b) and stores the
// result in lastScore. LCSStr is the Sørensen-Dice-normalised longest common
// substring similarity; only the dispatched byte-path score function is
// exercised via BDD (the substring-returning surface
// LongestCommonSubstring and the rune variants are covered by unit tests).
func (ctx *AlgorithmContext) iComputeTheLCSStrScoreBetween(a, b string) error {
	ctx.lastScore = fuzzymatch.LCSStrScore(a, b)
	return nil
}

// iComputeTheSecondLCSStrScoreBetween computes LCSStrScore(a, b) and stores
// the result in lastScore2. Used by symmetry scenarios to capture a second
// score for comparison.
func (ctx *AlgorithmContext) iComputeTheSecondLCSStrScoreBetween(a, b string) error {
	ctx.lastScore2 = fuzzymatch.LCSStrScore(a, b)
	return nil
}

// bothLCSStrScoresShouldBeEqual asserts lastScore == lastScore2.
func (ctx *AlgorithmContext) bothLCSStrScoresShouldBeEqual() error {
	if ctx.lastScore != ctx.lastScore2 {
		return fmt.Errorf("lcsstr scores not equal: %f != %f", ctx.lastScore, ctx.lastScore2)
	}
	return nil
}

// ---------------------------------------------------------------------------
// Q-Gram Jaccard step definitions (plan 05-01)
//
// Q-Gram Jaccard takes an additional `n int` parameter that the existing
// Phase 2/3/4 score-step grammar does not accommodate. New step regexes
// are registered with a `with n (\d+)` suffix to capture the q-gram size.
// Both surfaces (byte + rune) ship together; the symmetry scenario uses
// the standard "second score" / "both equal" pattern from earlier phases.
// ---------------------------------------------------------------------------

// iComputeTheQGramJaccardScoreBetweenWithN computes
// QGramJaccardScore(a, b, n) and stores the result in lastScore. The
// dispatched byte-path surface; multi-byte UTF-8 splits q-grams at
// byte boundaries.
func (ctx *AlgorithmContext) iComputeTheQGramJaccardScoreBetweenWithN(a, b string, n int) error {
	ctx.lastScore = fuzzymatch.QGramJaccardScore(a, b, n)
	return nil
}

// iComputeTheSecondQGramJaccardScoreBetweenWithN computes
// QGramJaccardScore(a, b, n) and stores the result in lastScore2. Used by
// the symmetry scenario to capture a second score for J(A, B) == J(B, A).
func (ctx *AlgorithmContext) iComputeTheSecondQGramJaccardScoreBetweenWithN(a, b string, n int) error {
	ctx.lastScore2 = fuzzymatch.QGramJaccardScore(a, b, n)
	return nil
}

// iComputeTheQGramJaccardRunesScoreBetweenWithN computes
// QGramJaccardScoreRunes(a, b, n) and stores the result in lastScore.
// The rune path; multi-byte UTF-8 windows are compared atomically.
func (ctx *AlgorithmContext) iComputeTheQGramJaccardRunesScoreBetweenWithN(a, b string, n int) error {
	ctx.lastScore = fuzzymatch.QGramJaccardScoreRunes(a, b, n)
	return nil
}

// bothQGramJaccardScoresShouldBeEqual asserts lastScore == lastScore2.
// Used by the symmetry scenario after computing J(A, B) and J(B, A).
func (ctx *AlgorithmContext) bothQGramJaccardScoresShouldBeEqual() error {
	if ctx.lastScore != ctx.lastScore2 {
		return fmt.Errorf("qgram jaccard scores not equal: %f != %f", ctx.lastScore, ctx.lastScore2)
	}
	return nil
}

// ---------------------------------------------------------------------------
// Sørensen-Dice step methods (plan 05-02)
//
// Same `with n <n>` shape introduced by plan 05-01 — the q-gram-tier
// algorithms all carry the n parameter. Both byte and rune surfaces ship;
// the symmetry scenario uses the standard "second score" / "both equal"
// pattern from earlier phases.
// ---------------------------------------------------------------------------

// iComputeTheSorensenDiceScoreBetweenWithN computes
// SorensenDiceScore(a, b, n) and stores the result in lastScore. The
// dispatched byte-path surface; multi-byte UTF-8 splits q-grams at byte
// boundaries.
func (ctx *AlgorithmContext) iComputeTheSorensenDiceScoreBetweenWithN(a, b string, n int) error {
	ctx.lastScore = fuzzymatch.SorensenDiceScore(a, b, n)
	return nil
}

// iComputeTheSecondSorensenDiceScoreBetweenWithN computes
// SorensenDiceScore(a, b, n) and stores the result in lastScore2. Used
// by the symmetry scenario to capture a second score for
// DSC(A, B) == DSC(B, A).
func (ctx *AlgorithmContext) iComputeTheSecondSorensenDiceScoreBetweenWithN(a, b string, n int) error {
	ctx.lastScore2 = fuzzymatch.SorensenDiceScore(a, b, n)
	return nil
}

// iComputeTheSorensenDiceRunesScoreBetweenWithN computes
// SorensenDiceScoreRunes(a, b, n) and stores the result in lastScore.
// The rune path; multi-byte UTF-8 windows are compared atomically.
func (ctx *AlgorithmContext) iComputeTheSorensenDiceRunesScoreBetweenWithN(a, b string, n int) error {
	ctx.lastScore = fuzzymatch.SorensenDiceScoreRunes(a, b, n)
	return nil
}

// bothSorensenDiceScoresShouldBeEqual asserts lastScore == lastScore2.
// Used by the symmetry scenario after computing DSC(A, B) and DSC(B, A).
func (ctx *AlgorithmContext) bothSorensenDiceScoresShouldBeEqual() error {
	if ctx.lastScore != ctx.lastScore2 {
		return fmt.Errorf("sorensen dice scores not equal: %f != %f", ctx.lastScore, ctx.lastScore2)
	}
	return nil
}

// ---------------------------------------------------------------------------
// Cosine step methods (plan 05-03)
//
// Same `with n <n>` shape introduced by plan 05-01 — the q-gram-tier
// algorithms all carry the n parameter. Both byte and rune surfaces ship;
// the symmetry scenario uses the standard "second score" / "both equal"
// pattern from earlier phases. The high-precision Examples in
// cosine.feature pin actual IEEE-754 factorised-form output (1 ULP from
// the rational limit per RESEARCH.md "Pitfall 2").
// ---------------------------------------------------------------------------

// iComputeTheCosineScoreBetweenWithN computes CosineScore(a, b, n) and
// stores the result in lastScore. The dispatched byte-path surface;
// multi-byte UTF-8 splits q-grams at byte boundaries.
func (ctx *AlgorithmContext) iComputeTheCosineScoreBetweenWithN(a, b string, n int) error {
	ctx.lastScore = fuzzymatch.CosineScore(a, b, n)
	return nil
}

// iComputeTheSecondCosineScoreBetweenWithN computes
// CosineScore(a, b, n) and stores the result in lastScore2. Used by
// the symmetry scenario to capture a second score for cos(A, B) ==
// cos(B, A).
func (ctx *AlgorithmContext) iComputeTheSecondCosineScoreBetweenWithN(a, b string, n int) error {
	ctx.lastScore2 = fuzzymatch.CosineScore(a, b, n)
	return nil
}

// iComputeTheCosineRunesScoreBetweenWithN computes
// CosineScoreRunes(a, b, n) and stores the result in lastScore. The
// rune path; multi-byte UTF-8 windows are compared atomically.
func (ctx *AlgorithmContext) iComputeTheCosineRunesScoreBetweenWithN(a, b string, n int) error {
	ctx.lastScore = fuzzymatch.CosineScoreRunes(a, b, n)
	return nil
}

// bothCosineScoresShouldBeEqual asserts lastScore == lastScore2. Used
// by the symmetry scenario after computing cos(A, B) and cos(B, A).
func (ctx *AlgorithmContext) bothCosineScoresShouldBeEqual() error {
	if ctx.lastScore != ctx.lastScore2 {
		return fmt.Errorf("cosine scores not equal: %f != %f", ctx.lastScore, ctx.lastScore2)
	}
	return nil
}

// ---------------------------------------------------------------------------
// Tversky step methods (plan 05-04)
//
// Carries the `with n <n> alpha <alpha> beta <beta>` shape — Tversky
// is the only q-gram-tier algorithm that ships an asymmetric (α/β)
// surface, so the BDD step grammar adds two float captures. The
// asymmetry direction-sensitivity scenario uses a dedicated
// "two scores should differ by more than X" step that asserts a
// MINIMUM separation between lastScore and lastScore2 — the
// LOAD-BEARING regression detector for parameter-order correctness
// (RV-T1 vs RV-T2 input swap). The parameter-swap symmetry scenario
// reuses the standard "both equal" pattern from earlier phases.
// ---------------------------------------------------------------------------

// iComputeTheTverskyScoreBetweenWithNAlphaBeta computes
// TverskyScore(a, b, n, alpha, beta) and stores the result in
// lastScore. The dispatched byte-path surface; multi-byte UTF-8
// splits q-grams at byte boundaries.
func (ctx *AlgorithmContext) iComputeTheTverskyScoreBetweenWithNAlphaBeta(a, b string, n int, alpha, beta float64) error {
	ctx.lastScore = fuzzymatch.TverskyScore(a, b, n, alpha, beta)
	return nil
}

// iComputeTheSecondTverskyScoreBetweenWithNAlphaBeta computes
// TverskyScore(a, b, n, alpha, beta) and stores the result in
// lastScore2. Used by the asymmetry direction-sensitivity scenario
// (RV-T1 vs RV-T2 input swap with same α/β) AND by the parameter-
// swap symmetry scenario (input swap WITH α/β swap).
func (ctx *AlgorithmContext) iComputeTheSecondTverskyScoreBetweenWithNAlphaBeta(a, b string, n int, alpha, beta float64) error {
	ctx.lastScore2 = fuzzymatch.TverskyScore(a, b, n, alpha, beta)
	return nil
}

// iComputeTheTverskyRunesScoreBetweenWithNAlphaBeta computes
// TverskyScoreRunes(a, b, n, alpha, beta) and stores the result in
// lastScore. The rune path; multi-byte UTF-8 windows are compared
// atomically.
func (ctx *AlgorithmContext) iComputeTheTverskyRunesScoreBetweenWithNAlphaBeta(a, b string, n int, alpha, beta float64) error {
	ctx.lastScore = fuzzymatch.TverskyScoreRunes(a, b, n, alpha, beta)
	return nil
}

// theTwoTverskyScoresShouldDifferByMoreThan asserts |lastScore −
// lastScore2| > threshold. The LOAD-BEARING asymmetry direction-
// sensitivity step: with the canonical RV-T1 vs RV-T2 input-swap pair
// at α=0.8/β=0.2, the actual difference is ≈ 0.2302; threshold 0.1
// gates against silent α/β swap regressions while leaving headroom
// for any future tightening of the IEEE-754 rounding behaviour.
func (ctx *AlgorithmContext) theTwoTverskyScoresShouldDifferByMoreThan(threshold float64) error {
	delta := math.Abs(ctx.lastScore - ctx.lastScore2)
	if delta <= threshold {
		return fmt.Errorf("tversky asymmetry gate FAILED: lastScore=%g, lastScore2=%g, |Δ|=%g; want > %g (the input swap with fixed α/β should produce direction-sensitive scores)",
			ctx.lastScore, ctx.lastScore2, delta, threshold)
	}
	return nil
}

// ---------------------------------------------------------------------------
// TokenSortRatio step methods (plan 06-01)
//
// Token Sort Ratio is parameter-free (no n, no α/β): the step grammar
// reverts to the Phase 2/3/4 score-step shape (no `with n …` suffix).
// Both the second-score and both-equal helpers mirror the
// Levenshtein / Jaro pattern from earlier phases.
// ---------------------------------------------------------------------------

// iComputeTheTokenSortRatioScoreBetween computes
// TokenSortRatioScore(a, b) and stores the result in lastScore.
func (ctx *AlgorithmContext) iComputeTheTokenSortRatioScoreBetween(a, b string) error {
	ctx.lastScore = fuzzymatch.TokenSortRatioScore(a, b)
	return nil
}

// iComputeTheSecondTokenSortRatioScoreBetween computes
// TokenSortRatioScore(a, b) and stores the result in lastScore2. Used
// by the symmetry scenario to capture a second score for
// T(A, B) == T(B, A).
func (ctx *AlgorithmContext) iComputeTheSecondTokenSortRatioScoreBetween(a, b string) error {
	ctx.lastScore2 = fuzzymatch.TokenSortRatioScore(a, b)
	return nil
}

// bothTokenSortRatioScoresShouldBeEqual asserts lastScore == lastScore2.
// Used by the symmetry scenario after computing T(A, B) and T(B, A).
func (ctx *AlgorithmContext) bothTokenSortRatioScoresShouldBeEqual() error {
	if ctx.lastScore != ctx.lastScore2 {
		return fmt.Errorf("token sort ratio scores not equal: %f != %f", ctx.lastScore, ctx.lastScore2)
	}
	return nil
}

// ---------------------------------------------------------------------------
// TokenSetRatio step methods (plan 06-02)
//
// Token Set Ratio is parameter-free (no n, no α/β): the step grammar
// reverts to the Phase 2/3/4 score-step shape (no `with n …` suffix),
// mirroring TokenSortRatio from plan 06-01.
// ---------------------------------------------------------------------------

// iComputeTheTokenSetRatioScoreBetween computes
// TokenSetRatioScore(a, b) and stores the result in lastScore.
func (ctx *AlgorithmContext) iComputeTheTokenSetRatioScoreBetween(a, b string) error {
	ctx.lastScore = fuzzymatch.TokenSetRatioScore(a, b)
	return nil
}

// iComputeTheSecondTokenSetRatioScoreBetween computes
// TokenSetRatioScore(a, b) and stores the result in lastScore2. Used
// by the symmetry scenario to capture a second score for
// T(A, B) == T(B, A).
func (ctx *AlgorithmContext) iComputeTheSecondTokenSetRatioScoreBetween(a, b string) error {
	ctx.lastScore2 = fuzzymatch.TokenSetRatioScore(a, b)
	return nil
}

// bothTokenSetRatioScoresShouldBeEqual asserts lastScore == lastScore2.
// Used by the symmetry scenario after computing T(A, B) and T(B, A).
func (ctx *AlgorithmContext) bothTokenSetRatioScoresShouldBeEqual() error {
	if ctx.lastScore != ctx.lastScore2 {
		return fmt.Errorf("token set ratio scores not equal: %f != %f", ctx.lastScore, ctx.lastScore2)
	}
	return nil
}

// bothTverskyScoresShouldBeEqual asserts lastScore == lastScore2.
// Used by the parameter-swap symmetry scenario after computing
// T(a, b, α, β) and T(b, a, β, α) — the algebraic identity from
// Tversky 1977 §2 that holds bit-for-bit.
func (ctx *AlgorithmContext) bothTverskyScoresShouldBeEqual() error {
	if ctx.lastScore != ctx.lastScore2 {
		return fmt.Errorf("tversky scores not equal: %f != %f (parameter-swap symmetry violated — T(a,b,α,β) should equal T(b,a,β,α))",
			ctx.lastScore, ctx.lastScore2)
	}
	return nil
}

// ---------------------------------------------------------------------------
// PartialRatio step methods (plan 06-03)
//
// Partial Ratio is parameter-free (no n, no α/β) and ships BOTH byte
// (PartialRatioScore — dispatched) and rune (PartialRatioScoreRunes —
// public but NOT dispatched) surfaces per 06-CONTEXT.md §6 LOCKED.
// Each surface gets its own step verb so feature authors can target
// the byte path with `PartialRatio` and the rune path with
// `PartialRatioRunes`.
// ---------------------------------------------------------------------------

// iComputeThePartialRatioScoreBetween computes PartialRatioScore(a, b)
// (byte path) and stores the result in lastScore.
func (ctx *AlgorithmContext) iComputeThePartialRatioScoreBetween(a, b string) error {
	ctx.lastScore = fuzzymatch.PartialRatioScore(a, b)
	return nil
}

// iComputeTheSecondPartialRatioScoreBetween computes
// PartialRatioScore(a, b) and stores the result in lastScore2. Used
// by the symmetry scenario to capture a second score for
// PR(A, B) == PR(B, A).
func (ctx *AlgorithmContext) iComputeTheSecondPartialRatioScoreBetween(a, b string) error {
	ctx.lastScore2 = fuzzymatch.PartialRatioScore(a, b)
	return nil
}

// bothPartialRatioScoresShouldBeEqual asserts lastScore == lastScore2.
// Used by the symmetry scenario after computing PR(A, B) and PR(B, A).
func (ctx *AlgorithmContext) bothPartialRatioScoresShouldBeEqual() error {
	if ctx.lastScore != ctx.lastScore2 {
		return fmt.Errorf("partial ratio scores not equal: %f != %f", ctx.lastScore, ctx.lastScore2)
	}
	return nil
}

// iComputeThePartialRatioRunesScoreBetween computes
// PartialRatioScoreRunes(a, b) (rune path) and stores the result in
// lastScore.
func (ctx *AlgorithmContext) iComputeThePartialRatioRunesScoreBetween(a, b string) error {
	ctx.lastScore = fuzzymatch.PartialRatioScoreRunes(a, b)
	return nil
}

// iComputeTheSecondPartialRatioRunesScoreBetween computes
// PartialRatioScoreRunes(a, b) and stores the result in lastScore2.
// Used by the rune-path symmetry scenario.
func (ctx *AlgorithmContext) iComputeTheSecondPartialRatioRunesScoreBetween(a, b string) error {
	ctx.lastScore2 = fuzzymatch.PartialRatioScoreRunes(a, b)
	return nil
}

// bothPartialRatioRunesScoresShouldBeEqual asserts lastScore ==
// lastScore2 on the rune-path symmetry scenario.
func (ctx *AlgorithmContext) bothPartialRatioRunesScoresShouldBeEqual() error {
	if ctx.lastScore != ctx.lastScore2 {
		return fmt.Errorf("partial ratio runes scores not equal: %f != %f", ctx.lastScore, ctx.lastScore2)
	}
	return nil
}

// ---------------------------------------------------------------------------
// TokenJaccard step methods (plan 06-04)
//
// TokenJaccard is parameter-free (no n, no α/β): the step grammar
// reverts to the Phase 2/3/4 score-step shape (no `with n …` suffix),
// mirroring TokenSortRatio from plan 06-01 / TokenSetRatio from plan
// 06-02 / PartialRatio from plan 06-03. Only one surface ships (no
// rune-path variant — Tokenise is UTF-8-aware so the rune semantic is
// preserved at the tokenisation layer; same convention as TokenSortRatio
// and TokenSetRatio).
// ---------------------------------------------------------------------------

// iComputeTheTokenJaccardScoreBetween computes
// TokenJaccardScore(a, b) and stores the result in lastScore.
func (ctx *AlgorithmContext) iComputeTheTokenJaccardScoreBetween(a, b string) error {
	ctx.lastScore = fuzzymatch.TokenJaccardScore(a, b)
	return nil
}

// iComputeTheSecondTokenJaccardScoreBetween computes
// TokenJaccardScore(a, b) and stores the result in lastScore2. Used by
// the symmetry scenario to capture a second score for
// T(A, B) == T(B, A).
func (ctx *AlgorithmContext) iComputeTheSecondTokenJaccardScoreBetween(a, b string) error {
	ctx.lastScore2 = fuzzymatch.TokenJaccardScore(a, b)
	return nil
}

// bothTokenJaccardScoresShouldBeEqual asserts lastScore == lastScore2.
// Used by the symmetry scenario after computing T(A, B) and T(B, A).
func (ctx *AlgorithmContext) bothTokenJaccardScoresShouldBeEqual() error {
	if ctx.lastScore != ctx.lastScore2 {
		return fmt.Errorf("token jaccard scores not equal: %f != %f", ctx.lastScore, ctx.lastScore2)
	}
	return nil
}

// InitializeScenario wires step definitions into the godog suite. Each call
// creates a fresh AlgorithmContext bound to the scenario, ensuring per-scenario
// isolation. Wave 2 plans append their algorithm's step regexes here.
//
// Step regexes use the godog-standard pattern: literal text with capture
// groups for variable parts. String captures use `([^"]*)` to exclude the
// surrounding quotes; numeric captures use `(\d+\.?\d*)` which accepts both
// integer-form (`0`, `1`) and decimal-form (`0.0`, `0.9444`) scores. The
// fractional part is optional so feature authors can write
// `the score should be exactly 0` as well as `... 0.0` (IN-03 closure).
func InitializeScenario(ctx *godog.ScenarioContext) {
	a := &AlgorithmContext{}

	// Levenshtein step definitions.
	ctx.Step(
		`^I compute the Levenshtein score between "([^"]*)" and "([^"]*)"$`,
		a.iComputeTheLevenshteinScoreBetween,
	)
	ctx.Step(
		`^I compute the second Levenshtein score between "([^"]*)" and "([^"]*)"$`,
		a.iComputeTheSecondLevenshteinScoreBetween,
	)
	ctx.Step(
		`^the score should be approximately (\d+\.?\d*) within (\d+\.?\d*)$`,
		a.theScoreShouldBeApproximately,
	)
	ctx.Step(
		`^the score should be exactly (\d+\.?\d*)$`,
		a.theScoreShouldBeExactly,
	)
	ctx.Step(
		`^both Levenshtein scores should be equal$`,
		a.bothLevenshteinScoresShouldBeEqual,
	)

	// Hamming step definitions (plan 02-02).
	ctx.Step(
		`^I compute the Hamming score between "([^"]*)" and "([^"]*)"$`,
		a.iComputeTheHammingScoreBetween,
	)
	ctx.Step(
		`^I compute the second Hamming score between "([^"]*)" and "([^"]*)"$`,
		a.iComputeTheSecondHammingScoreBetween,
	)
	ctx.Step(
		`^I compute the Hamming distance between "([^"]*)" and "([^"]*)"$`,
		a.iComputeTheHammingDistanceBetween,
	)
	ctx.Step(
		`^both Hamming scores should be equal$`,
		a.bothHammingScoresShouldBeEqual,
	)
	ctx.Step(
		`^the distance should be (\d+)$`,
		a.theDistanceShouldBe,
	)

	// Jaro step definitions (plan 02-03).
	ctx.Step(
		`^I compute the Jaro score between "([^"]*)" and "([^"]*)"$`,
		a.iComputeTheJaroScoreBetween,
	)
	ctx.Step(
		`^I compute the second Jaro score between "([^"]*)" and "([^"]*)"$`,
		a.iComputeTheSecondJaroScoreBetween,
	)
	ctx.Step(
		`^both Jaro scores should be equal$`,
		a.bothJaroScoresShouldBeEqual,
	)

	// Jaro-Winkler step definitions (plan 02-04).
	ctx.Step(
		`^I compute the JaroWinkler score between "([^"]*)" and "([^"]*)"$`,
		a.iComputeTheJaroWinklerScoreBetween,
	)
	ctx.Step(
		`^I compute the second JaroWinkler score between "([^"]*)" and "([^"]*)"$`,
		a.iComputeTheSecondJaroWinklerScoreBetween,
	)
	ctx.Step(
		`^both JaroWinkler scores should be equal$`,
		a.bothJaroWinklerScoresShouldBeEqual,
	)
	ctx.Step(
		`^both JaroWinkler and Jaro scores should be equal$`,
		a.bothJaroWinklerAndJaroScoresShouldBeEqual,
	)

	// Damerau-Levenshtein OSA step definitions (plan 02-05).
	ctx.Step(
		`^I compute the DamerauLevenshteinOSA score between "([^"]*)" and "([^"]*)"$`,
		a.iComputeTheDamerauLevenshteinOSAScoreBetween,
	)
	ctx.Step(
		`^I compute the second DamerauLevenshteinOSA score between "([^"]*)" and "([^"]*)"$`,
		a.iComputeTheSecondDamerauLevenshteinOSAScoreBetween,
	)
	ctx.Step(
		`^I compute the DamerauLevenshteinOSA distance between "([^"]*)" and "([^"]*)"$`,
		a.iComputeTheDamerauLevenshteinOSADistanceBetween,
	)
	ctx.Step(
		`^both DamerauLevenshteinOSA scores should be equal$`,
		a.bothDamerauLevenshteinOSAScoresShouldBeEqual,
	)

	// Damerau-Levenshtein Full step definitions (plan 02-06).
	ctx.Step(
		`^I compute the DamerauLevenshteinFull score between "([^"]*)" and "([^"]*)"$`,
		a.iComputeTheDamerauLevenshteinFullScoreBetween,
	)
	ctx.Step(
		`^I compute the second DamerauLevenshteinFull score between "([^"]*)" and "([^"]*)"$`,
		a.iComputeTheSecondDamerauLevenshteinFullScoreBetween,
	)
	ctx.Step(
		`^I compute the DamerauLevenshteinFull distance between "([^"]*)" and "([^"]*)"$`,
		a.iComputeTheDamerauLevenshteinFullDistanceBetween,
	)
	ctx.Step(
		`^both DamerauLevenshteinFull scores should be equal$`,
		a.bothDamerauLevenshteinFullScoresShouldBeEqual,
	)

	// Smith-Waterman-Gotoh step definitions (plan 03-01).
	ctx.Step(
		`^I compute the SmithWatermanGotoh score between "([^"]*)" and "([^"]*)"$`,
		a.iComputeTheSmithWatermanGotohScoreBetween,
	)
	ctx.Step(
		`^I compute the second SmithWatermanGotoh score between "([^"]*)" and "([^"]*)"$`,
		a.iComputeTheSecondSmithWatermanGotohScoreBetween,
	)
	ctx.Step(
		`^both SmithWatermanGotoh scores should be equal$`,
		a.bothSmithWatermanGotohScoresShouldBeEqual,
	)

	// Strcmp95 step definitions (plan 04-01).
	ctx.Step(
		`^I compute the Strcmp95 score between "([^"]*)" and "([^"]*)"$`,
		a.iComputeTheStrcmp95ScoreBetween,
	)
	ctx.Step(
		`^I compute the second Strcmp95 score between "([^"]*)" and "([^"]*)"$`,
		a.iComputeTheSecondStrcmp95ScoreBetween,
	)
	ctx.Step(
		`^both Strcmp95 scores should be equal$`,
		a.bothStrcmp95ScoresShouldBeEqual,
	)

	// LCSStr step definitions (plan 04-02).
	ctx.Step(
		`^I compute the LCSStr score between "([^"]*)" and "([^"]*)"$`,
		a.iComputeTheLCSStrScoreBetween,
	)
	ctx.Step(
		`^I compute the second LCSStr score between "([^"]*)" and "([^"]*)"$`,
		a.iComputeTheSecondLCSStrScoreBetween,
	)
	ctx.Step(
		`^both LCSStr scores should be equal$`,
		a.bothLCSStrScoresShouldBeEqual,
	)

	// Ratcliff-Obershelp step definitions (plan 04-03). NO "second" / "equal"
	// steps — the symmetry scenario is OMITTED per OQ-1 (RO is asymmetric
	// by design). The autojunk-sensitive scenario uses a dedicated step that
	// references 205-char Go constants rather than embedding the long inputs
	// in the Gherkin Examples table.
	ctx.Step(
		`^I compute the Ratcliff-Obershelp score between "([^"]*)" and "([^"]*)"$`,
		a.iComputeTheRatcliffObershelpScoreBetween,
	)
	ctx.Step(
		`^I compute the Ratcliff-Obershelp score for the autojunk-sensitive pair$`,
		a.iComputeTheRatcliffObershelpScoreForTheAutojunkSensitivePair,
	)

	// Q-Gram Jaccard step definitions (plan 05-01). Adds the `with n <n>`
	// suffix to capture the q-gram size that the Phase 2/3/4 grammar does
	// not carry. Both byte and rune surfaces ship; the symmetry scenario
	// uses the second-score / both-equal pattern from earlier phases.
	ctx.Step(
		`^I compute the QGramJaccard score between "([^"]*)" and "([^"]*)" with n (\d+)$`,
		a.iComputeTheQGramJaccardScoreBetweenWithN,
	)
	ctx.Step(
		`^I compute the second QGramJaccard score between "([^"]*)" and "([^"]*)" with n (\d+)$`,
		a.iComputeTheSecondQGramJaccardScoreBetweenWithN,
	)
	ctx.Step(
		`^I compute the QGramJaccardRunes score between "([^"]*)" and "([^"]*)" with n (\d+)$`,
		a.iComputeTheQGramJaccardRunesScoreBetweenWithN,
	)
	ctx.Step(
		`^both QGramJaccard scores should be equal$`,
		a.bothQGramJaccardScoresShouldBeEqual,
	)

	// Sørensen-Dice step definitions (plan 05-02). Same `with n <n>`
	// shape as plan 05-01's QGramJaccard. Both byte and rune surfaces
	// ship; the symmetry scenario uses the second-score / both-equal
	// pattern from earlier phases.
	ctx.Step(
		`^I compute the SorensenDice score between "([^"]*)" and "([^"]*)" with n (\d+)$`,
		a.iComputeTheSorensenDiceScoreBetweenWithN,
	)
	ctx.Step(
		`^I compute the second SorensenDice score between "([^"]*)" and "([^"]*)" with n (\d+)$`,
		a.iComputeTheSecondSorensenDiceScoreBetweenWithN,
	)
	ctx.Step(
		`^I compute the SorensenDiceRunes score between "([^"]*)" and "([^"]*)" with n (\d+)$`,
		a.iComputeTheSorensenDiceRunesScoreBetweenWithN,
	)
	ctx.Step(
		`^both SorensenDice scores should be equal$`,
		a.bothSorensenDiceScoresShouldBeEqual,
	)

	// Cosine step definitions (plan 05-03). Same `with n <n>` shape
	// as plan 05-01's QGramJaccard / plan 05-02's SorensenDice. Both
	// byte and rune surfaces ship; the symmetry scenario uses the
	// second-score / both-equal pattern from earlier phases. The
	// existing approximately-step regex `(\d+\.?\d*)` accepts the
	// 17-digit IEEE-754 form used by the high-precision examples in
	// cosine.feature (IN-03 closure carried forward).
	ctx.Step(
		`^I compute the Cosine score between "([^"]*)" and "([^"]*)" with n (\d+)$`,
		a.iComputeTheCosineScoreBetweenWithN,
	)
	ctx.Step(
		`^I compute the second Cosine score between "([^"]*)" and "([^"]*)" with n (\d+)$`,
		a.iComputeTheSecondCosineScoreBetweenWithN,
	)
	ctx.Step(
		`^I compute the CosineRunes score between "([^"]*)" and "([^"]*)" with n (\d+)$`,
		a.iComputeTheCosineRunesScoreBetweenWithN,
	)
	ctx.Step(
		`^both Cosine scores should be equal$`,
		a.bothCosineScoresShouldBeEqual,
	)

	// Tversky step definitions (plan 05-04). Adds the `alpha <alpha>
	// beta <beta>` suffix to the q-gram-tier `with n <n>` shape —
	// Tversky is the only q-gram algorithm that ships an asymmetric
	// (α/β) surface. The `differ by more than` step is the LOAD-BEARING
	// asymmetry direction-sensitivity gate at the BDD layer; the
	// `both equal` step is reused by the parameter-swap symmetry
	// scenario. The float captures use `(\d+\.?\d*)` (same shape as
	// the existing approximately-step regex) — Tversky requires α >= 0
	// and β >= 0, so unsigned floats suffice for the BDD grammar
	// (the panic-on-negative-α/β contract is unit-tested separately).
	ctx.Step(
		`^I compute the Tversky score between "([^"]*)" and "([^"]*)" with n (\d+) alpha (\d+\.?\d*) beta (\d+\.?\d*)$`,
		a.iComputeTheTverskyScoreBetweenWithNAlphaBeta,
	)
	ctx.Step(
		`^I compute the second Tversky score between "([^"]*)" and "([^"]*)" with n (\d+) alpha (\d+\.?\d*) beta (\d+\.?\d*)$`,
		a.iComputeTheSecondTverskyScoreBetweenWithNAlphaBeta,
	)
	ctx.Step(
		`^I compute the TverskyRunes score between "([^"]*)" and "([^"]*)" with n (\d+) alpha (\d+\.?\d*) beta (\d+\.?\d*)$`,
		a.iComputeTheTverskyRunesScoreBetweenWithNAlphaBeta,
	)
	ctx.Step(
		`^the two Tversky scores should differ by more than (\d+\.?\d*)$`,
		a.theTwoTverskyScoresShouldDifferByMoreThan,
	)
	ctx.Step(
		`^both Tversky scores should be equal$`,
		a.bothTverskyScoresShouldBeEqual,
	)

	// TokenSortRatio step definitions (plan 06-01). Parameter-free shape
	// (no `with n …` suffix) — Token Sort Ratio takes only (a, b string)
	// per CONTEXT.md §6 LOCKED. Both the score, second-score, and
	// both-equal helpers mirror the Levenshtein/Jaro Phase 2/3/4 pattern.
	ctx.Step(
		`^I compute the TokenSortRatio score between "([^"]*)" and "([^"]*)"$`,
		a.iComputeTheTokenSortRatioScoreBetween,
	)
	ctx.Step(
		`^I compute the second TokenSortRatio score between "([^"]*)" and "([^"]*)"$`,
		a.iComputeTheSecondTokenSortRatioScoreBetween,
	)
	ctx.Step(
		`^both TokenSortRatio scores should be equal$`,
		a.bothTokenSortRatioScoresShouldBeEqual,
	)

	// TokenSetRatio step definitions (plan 06-02). Parameter-free shape
	// (no `with n …` suffix) — Token Set Ratio takes only (a, b string)
	// per CONTEXT.md §6 LOCKED. Both the score, second-score, and
	// both-equal helpers mirror the TokenSortRatio pattern from plan
	// 06-01.
	ctx.Step(
		`^I compute the TokenSetRatio score between "([^"]*)" and "([^"]*)"$`,
		a.iComputeTheTokenSetRatioScoreBetween,
	)
	ctx.Step(
		`^I compute the second TokenSetRatio score between "([^"]*)" and "([^"]*)"$`,
		a.iComputeTheSecondTokenSetRatioScoreBetween,
	)
	ctx.Step(
		`^both TokenSetRatio scores should be equal$`,
		a.bothTokenSetRatioScoresShouldBeEqual,
	)

	// PartialRatio step definitions (plan 06-03) — BOTH byte and rune
	// surfaces. Parameter-free shape (no `with n …` suffix). Per
	// CONTEXT.md §6 LOCKED, PartialRatio ships both surfaces; the BDD
	// feature covers both with distinct step verbs (`PartialRatio` for
	// the byte path, `PartialRatioRunes` for the rune path).
	ctx.Step(
		`^I compute the PartialRatio score between "([^"]*)" and "([^"]*)"$`,
		a.iComputeThePartialRatioScoreBetween,
	)
	ctx.Step(
		`^I compute the second PartialRatio score between "([^"]*)" and "([^"]*)"$`,
		a.iComputeTheSecondPartialRatioScoreBetween,
	)
	ctx.Step(
		`^both PartialRatio scores should be equal$`,
		a.bothPartialRatioScoresShouldBeEqual,
	)
	ctx.Step(
		`^I compute the PartialRatioRunes score between "([^"]*)" and "([^"]*)"$`,
		a.iComputeThePartialRatioRunesScoreBetween,
	)
	ctx.Step(
		`^I compute the second PartialRatioRunes score between "([^"]*)" and "([^"]*)"$`,
		a.iComputeTheSecondPartialRatioRunesScoreBetween,
	)
	ctx.Step(
		`^both PartialRatioRunes scores should be equal$`,
		a.bothPartialRatioRunesScoresShouldBeEqual,
	)

	// TokenJaccard step definitions (plan 06-04). Parameter-free shape
	// (no `with n …` suffix) — TokenJaccard takes only (a, b string)
	// per CONTEXT.md §6 LOCKED. Both the score, second-score, and
	// both-equal helpers mirror the TokenSortRatio / TokenSetRatio
	// pattern from plans 06-01 / 06-02. No rune-path variant — Tokenise
	// is UTF-8-aware so the rune semantic is preserved at the
	// tokenisation layer.
	ctx.Step(
		`^I compute the TokenJaccard score between "([^"]*)" and "([^"]*)"$`,
		a.iComputeTheTokenJaccardScoreBetween,
	)
	ctx.Step(
		`^I compute the second TokenJaccard score between "([^"]*)" and "([^"]*)"$`,
		a.iComputeTheSecondTokenJaccardScoreBetween,
	)
	ctx.Step(
		`^both TokenJaccard scores should be equal$`,
		a.bothTokenJaccardScoresShouldBeEqual,
	)
}
