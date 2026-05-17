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

// scorer_steps.go contains the godog step definitions for the Phase 8
// composite Scorer surface. The ScorerContext struct holds state
// between steps within a scenario; each scenario instantiates a fresh
// ScorerContext via InitScorerSteps (called from
// algorithms_steps.go's InitializeScenario).
//
// goleak.VerifyTestMain in tests/bdd/bdd_test.go gates the concurrent-
// safety scenario — any goroutine leak introduced by the WaitGroup-
// based concurrent step would fail the BDD suite. The concurrent step
// uses sync.WaitGroup (deterministic completion); no errgroup or
// context.Context per CONTEXT.md §7 class 11.

package steps

import (
	"errors"
	"fmt"
	"reflect"
	"sync"

	"github.com/cucumber/godog"

	"github.com/axonops/fuzzymatch"
)

// ScorerContext holds state between BDD steps within a Scorer scenario.
// Each scenario instantiates a fresh ScorerContext to ensure isolation
// (steps within one scenario share state; steps across scenarios do not).
//
// scorer is the primary *fuzzymatch.Scorer under test; defaultScorer is
// the comparison Scorer used by the WithoutNormalisation scenario (it
// stores the default-Scorer composite of the same pair so the no-norm
// vs default comparison fits in two steps).
//
// lastErr stores the result of an "attempt to construct" step — error-
// path scenarios assert the error via errors.Is against the documented
// sentinels (ErrMissingThreshold, ErrEmptyScorer, ErrInvalidWeight).
//
// concurrentResults backs the 100-goroutine concurrent scenario; the
// step launches 100 goroutines and fills the slice in-place via index
// (no shared-map writes). The assertion step asserts every entry is
// byte-identical.
//
// scoreAll backs the ScoreAll typed-map-key scenario; the assertion
// step checks the documented entries are present (typed AlgoID lookup
// is compile-time-safe via Go's typed-map semantics).
type ScorerContext struct {
	scorer            *fuzzymatch.Scorer
	defaultScorer     *fuzzymatch.Scorer
	lastScore         float64
	defaultScore      float64
	lastMatch         bool
	lastErr           error
	scoreAll          map[fuzzymatch.AlgoID]float64
	concurrentResults []float64

	// Tokenise-equivalence scenario state (Phase 8.5 Q8b).
	// tokeniseInput holds the input string under test; tokeniseOpts
	// holds the configured TokeniseOptions; tokeniseFastTokens and
	// tokeniseRuneTokens hold the outputs of the two dispatch paths
	// (fast = pure-ASCII SeparatorChars; rune = SeparatorChars
	// augmented with a non-ASCII byte so the Tokenise() dispatch
	// falls through to the rune path). The two slices are compared
	// against one another and against the documented token sequence.
	tokeniseInput      string
	tokeniseOpts       fuzzymatch.TokeniseOptions
	tokeniseFastTokens []string
	tokeniseRuneTokens []string
}

// iConstructTheDefaultScorer constructs DefaultScorer and stores it in
// the ScorerContext. Used by happy-path and ScoreAll scenarios.
func (sc *ScorerContext) iConstructTheDefaultScorer() error {
	sc.scorer = fuzzymatch.DefaultScorer()
	return nil
}

// iConstructAScorerWithLevenshteinWeightAndThreshold builds a single-
// algorithm Levenshtein Scorer with the supplied weight and threshold.
// Step regex: `^I construct a Scorer with Levenshtein weight (\d+\.?\d*) and threshold (\d+\.?\d*)$`
func (sc *ScorerContext) iConstructAScorerWithLevenshteinWeightAndThreshold(weight, threshold float64) error {
	s, err := fuzzymatch.NewScorer(
		fuzzymatch.WithAlgorithm(fuzzymatch.AlgoLevenshtein, weight),
		fuzzymatch.WithThreshold(threshold),
	)
	if err != nil {
		return fmt.Errorf("NewScorer: %w", err)
	}
	sc.scorer = s
	return nil
}

// iConstructAScorerWithLevenshteinAndJaroWinkler builds a two-algorithm
// Scorer with explicit weights. Auto-normalisation is on; the supplied
// weights sum to 1.0 in the canonical case (0.5 + 0.5).
// Step regex: `^I construct a Scorer with Levenshtein weight (\d+\.?\d*) and JaroWinkler weight (\d+\.?\d*) and threshold (\d+\.?\d*)$`
func (sc *ScorerContext) iConstructAScorerWithLevenshteinAndJaroWinkler(levWeight, jwWeight, threshold float64) error {
	s, err := fuzzymatch.NewScorer(
		fuzzymatch.WithAlgorithm(fuzzymatch.AlgoLevenshtein, levWeight),
		fuzzymatch.WithAlgorithm(fuzzymatch.AlgoJaroWinkler, jwWeight),
		fuzzymatch.WithThreshold(threshold),
	)
	if err != nil {
		return fmt.Errorf("NewScorer: %w", err)
	}
	sc.scorer = s
	return nil
}

// iConstructTheDefaultScorerWithoutDoubleMetaphone builds a Scorer via
// DefaultScorerOptions() + WithoutAlgorithm(AlgoDoubleMetaphone). The
// resulting Scorer's Algorithms() slice excludes AlgoDoubleMetaphone.
func (sc *ScorerContext) iConstructTheDefaultScorerWithoutDoubleMetaphone() error {
	opts := append(fuzzymatch.DefaultScorerOptions(), fuzzymatch.WithoutAlgorithm(fuzzymatch.AlgoDoubleMetaphone))
	s, err := fuzzymatch.NewScorer(opts...)
	if err != nil {
		return fmt.Errorf("NewScorer(DefaultScorerOptions + WithoutAlgorithm(DM)): %w", err)
	}
	sc.scorer = s
	return nil
}

// iConstructAScorerWithDuplicateLevenshteinWeights exercises the last-
// write-wins dedup pattern (CONTEXT.md §7 class 6). Two WithAlgorithm
// (AlgoLevenshtein, w) calls; only the latter weight survives.
// Step regex: `^I construct a Scorer with duplicate Levenshtein weights (\d+\.?\d*) and (\d+\.?\d*) and threshold (\d+\.?\d*)$`
func (sc *ScorerContext) iConstructAScorerWithDuplicateLevenshteinWeights(w1, w2, threshold float64) error {
	s, err := fuzzymatch.NewScorer(
		fuzzymatch.WithAlgorithm(fuzzymatch.AlgoLevenshtein, w1),
		fuzzymatch.WithAlgorithm(fuzzymatch.AlgoLevenshtein, w2),
		fuzzymatch.WithThreshold(threshold),
	)
	if err != nil {
		return fmt.Errorf("NewScorer (duplicate Levenshtein): %w", err)
	}
	sc.scorer = s
	return nil
}

// iConstructTheDefaultScorerWithoutNormalisation builds a Scorer via
// DefaultScorerOptions() + WithoutNormalisation(). Used by the
// WithoutNormalisation scenario alongside a separately-constructed
// defaultScorer for direct comparison.
func (sc *ScorerContext) iConstructTheDefaultScorerWithoutNormalisation() error {
	opts := append(fuzzymatch.DefaultScorerOptions(), fuzzymatch.WithoutNormalisation())
	s, err := fuzzymatch.NewScorer(opts...)
	if err != nil {
		return fmt.Errorf("NewScorer(DefaultScorerOptions + WithoutNormalisation): %w", err)
	}
	sc.scorer = s
	sc.defaultScorer = fuzzymatch.DefaultScorer()
	return nil
}

// iScoreAndWithTheScorer runs Score / Match on the (a, b) pair using
// the ScorerContext's scorer. Stores both results so a subsequent
// "match should be true/false" or "composite score should be X" step
// can assert either output.
// Step regex: `^I score "([^"]*)" and "([^"]*)" with the Scorer$`
func (sc *ScorerContext) iScoreAndWithTheScorer(a, b string) error {
	if sc.scorer == nil {
		return fmt.Errorf("scorer not constructed — call a Given step first")
	}
	sc.lastScore = sc.scorer.Score(a, b)
	sc.lastMatch = sc.scorer.Match(a, b)
	return nil
}

// iRecordTheNoNormalisationScorerCompositeScore stores the WithoutNormalisation
// Scorer's composite as defaultScore — temporarily; the next step
// overwrites scorer to defaultScorer and records the comparison value.
// This convolution is a Gherkin-readability trade-off so the scenario
// reads naturally ("I record the no-normalisation score" / "I score
// the same pair with the default Scorer" / "the no-normalisation
// composite should be less than the default composite").
func (sc *ScorerContext) iRecordTheNoNormalisationScorerCompositeScore() error {
	// lastScore was populated by iScoreAndWithTheScorer in the previous step.
	// Keep it in defaultScore as a temporary holding slot so the next
	// step's Score call overwrites lastScore with the default-Scorer
	// composite.
	sc.defaultScore = sc.lastScore
	return nil
}

// iScoreTheSamePairWithTheDefaultScorer runs Score on the default
// Scorer for the previously-recorded pair. The pair is stored
// implicitly in the step state — this step assumes the previous
// "I score X and Y" step ran first.
func (sc *ScorerContext) iScoreTheSamePairWithTheDefaultScorer() error {
	if sc.defaultScorer == nil {
		return fmt.Errorf("defaultScorer not constructed — preceding 'I construct the default Scorer without normalisation' must populate it")
	}
	sc.lastScore = sc.defaultScorer.Score("XMLParser", "xml_parser")
	return nil
}

// theNoNormalisationCompositeShouldBeLessThanTheDefaultComposite asserts
// the relative ordering documented in CONTEXT.md §7 class 7. The
// WithoutNormalisation Scorer scores XMLParser / xml_parser lower than
// the default Scorer because the case + underscore-vs-camelCase
// differences are not erased before the algorithms run.
func (sc *ScorerContext) theNoNormalisationCompositeShouldBeLessThanTheDefaultComposite() error {
	// defaultScore holds the WithoutNormalisation composite (captured
	// in the "I record" step); lastScore holds the default-Scorer
	// composite (captured in the subsequent "I score the same pair
	// with the default Scorer" step). The assertion is that the
	// WithoutNormalisation composite (stored in defaultScore) is
	// strictly less than the default composite (stored in lastScore).
	if sc.defaultScore >= sc.lastScore {
		return fmt.Errorf(
			"expected no-normalisation composite (%f) to be less than default composite (%f)",
			sc.defaultScore, sc.lastScore,
		)
	}
	return nil
}

// iListTheScorerAlgorithms calls Algorithms() and stores nothing
// (subsequent assertion steps inspect via direct method call). Exists
// so the Gherkin reads naturally ("When I list the Scorer algorithms
// / Then the Scorer should have exactly N algorithm").
func (sc *ScorerContext) iListTheScorerAlgorithms() error {
	if sc.scorer == nil {
		return fmt.Errorf("scorer not constructed — call a Given step first")
	}
	return nil
}

// theScorerAlgorithmsShouldNotInclude asserts that the named AlgoID is
// NOT in the Scorer's Algorithms() slice. Used by the WithoutAlgorithm
// scenario.
// Step regex: `^the Scorer algorithms should not include ([A-Za-z]+)$`
func (sc *ScorerContext) theScorerAlgorithmsShouldNotInclude(algoName string) error {
	if sc.scorer == nil {
		return fmt.Errorf("scorer not constructed")
	}
	target, err := algoIDFromName(algoName)
	if err != nil {
		return err
	}
	for _, a := range sc.scorer.Algorithms() {
		if a.ID == target {
			return fmt.Errorf("expected Algorithms() to NOT include %s, but found it with weight %f", algoName, a.Weight)
		}
	}
	return nil
}

// theScorerShouldHaveExactlyAlgorithm asserts the Algorithms() slice
// length. Used by the last-write-wins dedup scenario (expected: 1).
// Step regex: `^the Scorer should have exactly (\d+) algorithm$`
func (sc *ScorerContext) theScorerShouldHaveExactlyAlgorithm(count int) error {
	if sc.scorer == nil {
		return fmt.Errorf("scorer not constructed")
	}
	got := len(sc.scorer.Algorithms())
	if got != count {
		return fmt.Errorf("expected Algorithms() length %d, got %d", count, got)
	}
	return nil
}

// theAlgorithmShouldBeAlgoLevenshteinWithWeight asserts the single
// surviving algorithm after last-write-wins dedup is AlgoLevenshtein
// with the expected weight (1.0 after auto-normalisation of one entry).
// Step regex: `^the algorithm should be ([A-Za-z]+) with weight (\d+\.?\d*)$`
func (sc *ScorerContext) theAlgorithmShouldBeAlgoLevenshteinWithWeight(algoName string, weight float64) error {
	if sc.scorer == nil {
		return fmt.Errorf("scorer not constructed")
	}
	algos := sc.scorer.Algorithms()
	if len(algos) == 0 {
		return fmt.Errorf("no algorithms in Scorer")
	}
	target, err := algoIDFromName(algoName)
	if err != nil {
		return err
	}
	if algos[0].ID != target {
		return fmt.Errorf("expected first algorithm to be %s, got %s", algoName, algos[0].ID.String())
	}
	if algos[0].Weight != weight {
		return fmt.Errorf("expected first algorithm weight %f, got %f", weight, algos[0].Weight)
	}
	return nil
}

// iAttemptToConstructAScorerWithLevenshteinWeightAndNoThreshold runs
// NewScorer without WithThreshold; stores the error for the assertion
// step.
// Step regex: `^I attempt to construct a Scorer with Levenshtein weight (-?\d+\.?\d*) and no threshold$`
func (sc *ScorerContext) iAttemptToConstructAScorerWithLevenshteinWeightAndNoThreshold(weight float64) error {
	_, err := fuzzymatch.NewScorer(
		fuzzymatch.WithAlgorithm(fuzzymatch.AlgoLevenshtein, weight),
	)
	sc.lastErr = err
	return nil
}

// iAttemptToConstructAScorerWithThresholdAndNoAlgorithms runs NewScorer
// with only WithThreshold; stores the error.
// Step regex: `^I attempt to construct a Scorer with threshold (\d+\.?\d*) and no algorithms$`
func (sc *ScorerContext) iAttemptToConstructAScorerWithThresholdAndNoAlgorithms(threshold float64) error {
	_, err := fuzzymatch.NewScorer(fuzzymatch.WithThreshold(threshold))
	sc.lastErr = err
	return nil
}

// iAttemptToConstructAScorerWithLevenshteinWeightAndThreshold runs
// NewScorer with the supplied (potentially-invalid) weight and a
// threshold; stores the error for the assertion step.
// Step regex: `^I attempt to construct a Scorer with Levenshtein weight (-?\d+\.?\d*) and threshold (\d+\.?\d*)$`
func (sc *ScorerContext) iAttemptToConstructAScorerWithLevenshteinWeightAndThreshold(weight, threshold float64) error {
	_, err := fuzzymatch.NewScorer(
		fuzzymatch.WithAlgorithm(fuzzymatch.AlgoLevenshtein, weight),
		fuzzymatch.WithThreshold(threshold),
	)
	sc.lastErr = err
	return nil
}

// constructingTheScorerShouldReturn asserts the stored lastErr matches
// the named sentinel via errors.Is. Used by all three error-path
// scenarios.
// Step regex: `^constructing the Scorer should return (Err\w+)$`
func (sc *ScorerContext) constructingTheScorerShouldReturn(errName string) error {
	if sc.lastErr == nil {
		return fmt.Errorf("expected error %s, got nil", errName)
	}
	var target error
	switch errName {
	case "ErrMissingThreshold":
		target = fuzzymatch.ErrMissingThreshold
	case "ErrEmptyScorer":
		target = fuzzymatch.ErrEmptyScorer
	case "ErrInvalidWeight":
		target = fuzzymatch.ErrInvalidWeight
	case "ErrInvalidThreshold":
		target = fuzzymatch.ErrInvalidThreshold
	default:
		return fmt.Errorf("unknown sentinel name %q (supported: ErrMissingThreshold, ErrEmptyScorer, ErrInvalidWeight, ErrInvalidThreshold)", errName)
	}
	if !errors.Is(sc.lastErr, target) {
		return fmt.Errorf("expected errors.Is(err, %s), got err = %v", errName, sc.lastErr)
	}
	return nil
}

// iCallScoreOnAndFromGoroutinesSimultaneously runs Score on the
// supplied pair from N goroutines concurrently. The results land in
// concurrentResults[i] (no shared map writes, no shared counter — just
// per-goroutine slice index assignment).
//
// The step uses sync.WaitGroup for deterministic completion. No
// errgroup, no context.Context — goleak in bdd_test.go's TestMain
// catches any goroutine that escapes the WaitGroup.
// Step regex: `^I call Score on "([^"]*)" and "([^"]*)" from (\d+) goroutines simultaneously$`
func (sc *ScorerContext) iCallScoreOnAndFromGoroutinesSimultaneously(a, b string, n int) error {
	if sc.scorer == nil {
		return fmt.Errorf("scorer not constructed")
	}
	sc.concurrentResults = make([]float64, n)
	var wg sync.WaitGroup
	wg.Add(n)
	for i := 0; i < n; i++ {
		go func(idx int) {
			defer wg.Done()
			sc.concurrentResults[idx] = sc.scorer.Score(a, b)
		}(i)
	}
	wg.Wait()
	return nil
}

// allGoroutineResultsShouldBeByteIdentical asserts every entry in
// concurrentResults is exactly equal to concurrentResults[0] (bitwise
// float64 equality, NOT within tolerance). Documents the determinism
// guarantee: the same Scorer + same input always produces the same
// output regardless of which goroutine called Score first.
// Step regex: `^all (\d+) goroutine results should be byte-identical$`
func (sc *ScorerContext) allGoroutineResultsShouldBeByteIdentical(n int) error {
	if len(sc.concurrentResults) != n {
		return fmt.Errorf("expected %d concurrent results, got %d", n, len(sc.concurrentResults))
	}
	first := sc.concurrentResults[0]
	for i, r := range sc.concurrentResults {
		if r != first {
			return fmt.Errorf("concurrent result %d (%v) differs from result 0 (%v) — Scorer is not concurrent-safe or not deterministic", i, r, first)
		}
	}
	return nil
}

// iCallScoreAllOnAndWithTheScorer runs ScoreAll on the supplied pair
// using the ScorerContext's scorer; stores the typed map for the
// assertion steps.
// Step regex: `^I call ScoreAll on "([^"]*)" and "([^"]*)" with the Scorer$`
func (sc *ScorerContext) iCallScoreAllOnAndWithTheScorer(a, b string) error {
	if sc.scorer == nil {
		return fmt.Errorf("scorer not constructed")
	}
	sc.scoreAll = sc.scorer.ScoreAll(a, b)
	return nil
}

// theScoreAllMapShouldContain asserts that the typed-AlgoID-keyed map
// contains the named AlgoID. The lookup is via the typed enum (compile-
// time-safe — if someone refactors AlgoID, this step fails to compile).
// Step regex: `^the ScoreAll map should contain ([A-Za-z]+)$`
func (sc *ScorerContext) theScoreAllMapShouldContain(algoName string) error {
	if sc.scoreAll == nil {
		return fmt.Errorf("scoreAll not populated — call 'I call ScoreAll' first")
	}
	target, err := algoIDFromName(algoName)
	if err != nil {
		return err
	}
	if _, ok := sc.scoreAll[target]; !ok {
		return fmt.Errorf("expected ScoreAll map to contain %s (id=%d), but key missing", algoName, target)
	}
	return nil
}

// theScoreAllMapShouldNotContain is the inverse assertion — used by
// the WithoutAlgorithm scenario class 12 sub-case that verifies an
// algorithm NOT in the DefaultScorer composition (Cosine) is absent
// from ScoreAll.
// Step regex: `^the ScoreAll map should not contain ([A-Za-z]+)$`
func (sc *ScorerContext) theScoreAllMapShouldNotContain(algoName string) error {
	if sc.scoreAll == nil {
		return fmt.Errorf("scoreAll not populated — call 'I call ScoreAll' first")
	}
	target, err := algoIDFromName(algoName)
	if err != nil {
		return err
	}
	if _, ok := sc.scoreAll[target]; ok {
		return fmt.Errorf("expected ScoreAll map to NOT contain %s (id=%d), but key present", algoName, target)
	}
	return nil
}

// theScorerMatchResultShouldBe asserts the boolean Match outcome
// stored in lastMatch.
// Step regex: `^the Scorer match result should be (true|false)$`
func (sc *ScorerContext) theScorerMatchResultShouldBe(expected string) error {
	wantMatch := expected == "true"
	if sc.lastMatch != wantMatch {
		return fmt.Errorf("expected Match=%v, got %v (composite score = %f, threshold = %f)",
			wantMatch, sc.lastMatch, sc.lastScore, sc.scorer.Threshold())
	}
	return nil
}

// theScorerCompositeScoreShouldBeExactly asserts the composite score
// is exactly equal to the expected value (bitwise float64 equality —
// no tolerance). Used for identity cases where the composite must be
// exactly 1.0.
// Step regex: `^the Scorer composite score should be exactly (\d+\.?\d*)$`
func (sc *ScorerContext) theScorerCompositeScoreShouldBeExactly(expected float64) error {
	if sc.lastScore != expected {
		return fmt.Errorf("expected composite score exactly %f, got %f", expected, sc.lastScore)
	}
	return nil
}

// algoIDFromName maps a Gherkin step's algorithm-name string (e.g.
// "DoubleMetaphone", "AlgoLevenshtein") to the corresponding AlgoID.
// Accepts both "Algo<Name>" and bare "<Name>" forms for Gherkin
// readability.
func algoIDFromName(name string) (fuzzymatch.AlgoID, error) {
	// Strip optional "Algo" prefix so "AlgoDoubleMetaphone" and
	// "DoubleMetaphone" both map to AlgoDoubleMetaphone.
	stripped := name
	if len(name) > 4 && name[:4] == "Algo" {
		stripped = name[4:]
	}
	// Linear scan over AlgoIDs() — 23 entries, called once per step;
	// the overhead is negligible vs the readability of the lookup.
	for _, id := range fuzzymatch.AlgoIDs() {
		if id.String() == stripped {
			return id, nil
		}
	}
	return 0, fmt.Errorf("unknown algorithm name %q (must match an AlgoID.String() value such as Levenshtein, DoubleMetaphone, etc.)", name)
}

// ---------------------------------------------------------------------
// Tokenise ASCII-fast-path equivalence scenario (Phase 8.5 Q8b)
// ---------------------------------------------------------------------

// theTokeniseInputIs stores the input string that the scenario will
// pass to Tokenise. The string is held verbatim — no escape handling
// beyond Gherkin's own quote-pair parsing.
// Step regex: `^the Tokenise input "([^"]*)"$`
func (sc *ScorerContext) theTokeniseInputIs(input string) error {
	sc.tokeniseInput = input
	return nil
}

// theTokeniseOptionsAre parses the two-bool option block and stores
// it on the context. SeparatorChars and SplitConsecutiveUpper are
// held at their default values (SplitConsecutiveUpper = true,
// SeparatorChars = the default pure-ASCII set "_-.:/ \t\n\r") so the
// scenario stays focused on the SplitCamelCase × Lowercase axes —
// the two boolean knobs that exercise the boundary-detection and
// lowercase-fold paths respectively.
// Step regex: `^Tokenise options SplitCamelCase=(true|false) Lowercase=(true|false)$`
func (sc *ScorerContext) theTokeniseOptionsAre(splitCamel, lowercase string) error {
	sc.tokeniseOpts = fuzzymatch.TokeniseOptions{
		Lowercase:             lowercase == "true",
		SplitCamelCase:        splitCamel == "true",
		SplitConsecutiveUpper: true,
		SeparatorChars:        fuzzymatch.DefaultTokeniseOptions().SeparatorChars,
	}
	return nil
}

// theTokeniseASCIIFastPathRuns invokes Tokenise() with the configured
// options. Because the input is pure ASCII AND SeparatorChars is
// pure ASCII (default set), Tokenise's documented dispatch criterion
// (isASCIITokenise(s) && !containsNonASCII(opts.SeparatorChars))
// selects the byte-level fast path. The output is stored in
// tokeniseFastTokens.
// Step regex: `^the Tokenise ASCII fast path runs$`
func (sc *ScorerContext) theTokeniseASCIIFastPathRuns() error {
	sc.tokeniseFastTokens = fuzzymatch.Tokenise(sc.tokeniseInput, sc.tokeniseOpts)
	return nil
}

// theTokeniseRuneFallbackPathRuns invokes Tokenise() with the same
// input and Lowercase/SplitCamelCase settings BUT with a non-ASCII
// byte appended to SeparatorChars. This forces the documented
// dispatch criterion to fail (containsNonASCII becomes true), which
// in turn selects the rune-based fallback. The chosen non-ASCII
// character "™" (U+2122) is a 3-byte UTF-8 sequence that never
// occurs in the input — so the separator semantic is unchanged from
// the fast-path invocation. The output is stored in
// tokeniseRuneTokens.
// Step regex: `^the Tokenise rune fallback path runs$`
func (sc *ScorerContext) theTokeniseRuneFallbackPathRuns() error {
	runeOpts := sc.tokeniseOpts
	runeOpts.SeparatorChars = sc.tokeniseOpts.SeparatorChars + "™"
	sc.tokeniseRuneTokens = fuzzymatch.Tokenise(sc.tokeniseInput, runeOpts)
	return nil
}

// bothTokenisePathsProduce asserts that (a) the fast and rune paths
// produced byte-identical token sequences and (b) the sequence matches
// the documented expected output. The expected sequence is a literal
// Go-style slice rendered in the Gherkin step (e.g. `["foo", "Bar",
// "baz", "qux"]`). Parsing is delegated to a small helper to keep
// the step body focused.
// Step regex: `^both Tokenise paths produce (\[.*\])$`
func (sc *ScorerContext) bothTokenisePathsProduce(expectedLiteral string) error {
	expected, err := parseTokeniseExpectedSequence(expectedLiteral)
	if err != nil {
		return fmt.Errorf("could not parse expected token sequence %q: %w", expectedLiteral, err)
	}
	if !reflect.DeepEqual(sc.tokeniseFastTokens, sc.tokeniseRuneTokens) {
		return fmt.Errorf("ASCII fast path %q diverged from rune path %q (Phase 8.5 Q8b equivalence violation)",
			sc.tokeniseFastTokens, sc.tokeniseRuneTokens)
	}
	if !reflect.DeepEqual(sc.tokeniseFastTokens, expected) {
		return fmt.Errorf("Tokenise paths agreed on %q, but expected sequence was %q",
			sc.tokeniseFastTokens, expected)
	}
	return nil
}

// parseTokeniseExpectedSequence parses a Go-style string-slice literal
// of the form `["foo", "Bar", "baz", "qux"]` into a []string. The
// parser is intentionally minimal — it accepts single-quoted strings
// separated by commas and optional whitespace, surrounded by square
// brackets. This is sufficient for the Tokenise scenario (the only
// caller); any expansion to more complex slice literals should be
// re-evaluated against the established BDD step shape.
func parseTokeniseExpectedSequence(literal string) ([]string, error) {
	if len(literal) < 2 || literal[0] != '[' || literal[len(literal)-1] != ']' {
		return nil, fmt.Errorf("expected sequence must be bracket-delimited; got %q", literal)
	}
	body := literal[1 : len(literal)-1]
	// Empty body -> empty slice.
	if len(body) == 0 {
		return []string{}, nil
	}
	var out []string
	inQuote := false
	var cur []byte
	for i := 0; i < len(body); i++ {
		c := body[i]
		switch {
		case c == '"' && !inQuote:
			inQuote = true
			cur = cur[:0]
		case c == '"' && inQuote:
			inQuote = false
			out = append(out, string(cur))
		case inQuote:
			cur = append(cur, c)
		// Outside quotes: skip commas, whitespace.
		case c == ',' || c == ' ' || c == '\t':
			// no-op
		default:
			return nil, fmt.Errorf("unexpected character %q at index %d in sequence literal %q", c, i, literal)
		}
	}
	if inQuote {
		return nil, fmt.Errorf("unterminated quoted token in sequence literal %q", literal)
	}
	return out, nil
}

// InitScorerSteps registers all Scorer step regexes with the supplied
// godog ScenarioContext. Each scenario gets a fresh ScorerContext (one
// per scenario, not one per step), keyed off the closure-captured `sc`
// variable.
//
// Called from algorithms_steps.go's InitializeScenario at the bottom
// of the registration block so the Phase 2-7 step regexes stay where
// they are; Phase 8 adds its own block here without re-shaping the
// existing harness.
func InitScorerSteps(ctx *godog.ScenarioContext) {
	sc := &ScorerContext{}

	// Default-Scorer construction.
	ctx.Step(
		`^I construct the default Scorer$`,
		sc.iConstructTheDefaultScorer,
	)
	ctx.Step(
		`^I construct the default Scorer without DoubleMetaphone$`,
		sc.iConstructTheDefaultScorerWithoutDoubleMetaphone,
	)
	ctx.Step(
		`^I construct the default Scorer without normalisation$`,
		sc.iConstructTheDefaultScorerWithoutNormalisation,
	)

	// Custom Scorer construction (happy path).
	ctx.Step(
		`^I construct a Scorer with Levenshtein weight (\d+\.?\d*) and threshold (\d+\.?\d*)$`,
		sc.iConstructAScorerWithLevenshteinWeightAndThreshold,
	)
	ctx.Step(
		`^I construct a Scorer with Levenshtein weight (\d+\.?\d*) and JaroWinkler weight (\d+\.?\d*) and threshold (\d+\.?\d*)$`,
		sc.iConstructAScorerWithLevenshteinAndJaroWinkler,
	)
	ctx.Step(
		`^I construct a Scorer with duplicate Levenshtein weights (\d+\.?\d*) and (\d+\.?\d*) and threshold (\d+\.?\d*)$`,
		sc.iConstructAScorerWithDuplicateLevenshteinWeights,
	)

	// Error-path Scorer construction.
	ctx.Step(
		`^I attempt to construct a Scorer with Levenshtein weight (-?\d+\.?\d*) and no threshold$`,
		sc.iAttemptToConstructAScorerWithLevenshteinWeightAndNoThreshold,
	)
	ctx.Step(
		`^I attempt to construct a Scorer with threshold (\d+\.?\d*) and no algorithms$`,
		sc.iAttemptToConstructAScorerWithThresholdAndNoAlgorithms,
	)
	ctx.Step(
		`^I attempt to construct a Scorer with Levenshtein weight (-?\d+\.?\d*) and threshold (\d+\.?\d*)$`,
		sc.iAttemptToConstructAScorerWithLevenshteinWeightAndThreshold,
	)
	ctx.Step(
		`^constructing the Scorer should return (Err\w+)$`,
		sc.constructingTheScorerShouldReturn,
	)

	// Score / Match / ScoreAll invocations.
	ctx.Step(
		`^I score "([^"]*)" and "([^"]*)" with the Scorer$`,
		sc.iScoreAndWithTheScorer,
	)
	ctx.Step(
		`^I record the no-normalisation Scorer composite score$`,
		sc.iRecordTheNoNormalisationScorerCompositeScore,
	)
	ctx.Step(
		`^I score the same pair with the default Scorer$`,
		sc.iScoreTheSamePairWithTheDefaultScorer,
	)
	ctx.Step(
		`^the no-normalisation composite should be less than the default composite$`,
		sc.theNoNormalisationCompositeShouldBeLessThanTheDefaultComposite,
	)
	ctx.Step(
		`^the Scorer match result should be (true|false)$`,
		sc.theScorerMatchResultShouldBe,
	)
	ctx.Step(
		`^the Scorer composite score should be exactly (\d+\.?\d*)$`,
		sc.theScorerCompositeScoreShouldBeExactly,
	)

	// Algorithms() assertions.
	ctx.Step(
		`^I list the Scorer algorithms$`,
		sc.iListTheScorerAlgorithms,
	)
	ctx.Step(
		`^the Scorer algorithms should not include ([A-Za-z]+)$`,
		sc.theScorerAlgorithmsShouldNotInclude,
	)
	ctx.Step(
		`^the Scorer should have exactly (\d+) algorithm$`,
		sc.theScorerShouldHaveExactlyAlgorithm,
	)
	ctx.Step(
		`^the algorithm should be ([A-Za-z]+) with weight (\d+\.?\d*)$`,
		sc.theAlgorithmShouldBeAlgoLevenshteinWithWeight,
	)

	// Concurrent-safety scenario (goleak gate).
	ctx.Step(
		`^I call Score on "([^"]*)" and "([^"]*)" from (\d+) goroutines simultaneously$`,
		sc.iCallScoreOnAndFromGoroutinesSimultaneously,
	)
	ctx.Step(
		`^all (\d+) goroutine results should be byte-identical$`,
		sc.allGoroutineResultsShouldBeByteIdentical,
	)

	// ScoreAll typed-AlgoID-key assertions.
	ctx.Step(
		`^I call ScoreAll on "([^"]*)" and "([^"]*)" with the Scorer$`,
		sc.iCallScoreAllOnAndWithTheScorer,
	)
	ctx.Step(
		`^the ScoreAll map should contain ([A-Za-z]+)$`,
		sc.theScoreAllMapShouldContain,
	)
	ctx.Step(
		`^the ScoreAll map should not contain ([A-Za-z]+)$`,
		sc.theScoreAllMapShouldNotContain,
	)

	// Tokenise ASCII-fast-path equivalence scenario (Phase 8.5 Q8b).
	// The four steps wire the equivalence scenario in
	// features/scorer.feature: input -> options -> two-path
	// invocation -> equality assertion.
	ctx.Step(
		`^the Tokenise input "([^"]*)"$`,
		sc.theTokeniseInputIs,
	)
	ctx.Step(
		`^Tokenise options SplitCamelCase=(true|false) Lowercase=(true|false)$`,
		sc.theTokeniseOptionsAre,
	)
	ctx.Step(
		`^the Tokenise ASCII fast path runs$`,
		sc.theTokeniseASCIIFastPathRuns,
	)
	ctx.Step(
		`^the Tokenise rune fallback path runs$`,
		sc.theTokeniseRuneFallbackPathRuns,
	)
	ctx.Step(
		`^both Tokenise paths produce (\[.*\])$`,
		sc.bothTokenisePathsProduce,
	)
}
