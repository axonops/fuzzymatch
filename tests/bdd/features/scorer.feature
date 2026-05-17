# Phase 8 BDD coverage: composite weighted Scorer.
#
# Mandatory scenario classes per 08-CONTEXT.md §7 (all 12 represented):
#   1. Default scorer happy path                    -> "Default scorer matches identifier-style pair"
#   2. Default scorer below threshold               -> "Default scorer rejects dissimilar inputs"
#   3. Custom 1-algorithm Scorer                    -> "Single-algorithm Scorer composes correctly"
#   4. Custom 2-algorithm weighted Scorer           -> "Two-algorithm weighted Scorer composes correctly"
#   5. WithoutAlgorithm composition                 -> "WithoutAlgorithm removes the algorithm"
#   6. Last-write-wins for duplicate AlgoID         -> "Duplicate WithAlgorithm calls use the latest weight"
#   7. WithoutNormalisation behaviour               -> "WithoutNormalisation preserves raw bytes"
#   8. ErrMissingThreshold                          -> "NewScorer without WithThreshold returns ErrMissingThreshold"
#   9. ErrEmptyScorer                               -> "NewScorer without algorithms returns ErrEmptyScorer"
#  10. ErrInvalidWeight                             -> "WithAlgorithm with negative weight returns ErrInvalidWeight"
#  11. Concurrent safety                            -> "Concurrent Score calls return identical results"
#  12. ScoreAll AlgoID keys                         -> "ScoreAll returns map keyed by AlgoID"
#
# goleak.VerifyTestMain in tests/bdd/bdd_test.go is the goroutine-leak
# detector that gates scenario 11. The concurrent step uses sync.WaitGroup
# (deterministic completion); no goroutine leaks if all 100 goroutines
# call wg.Done().

@scorer
Feature: Composite weighted Scorer (Phase 8)
  The Scorer composes any subset of the 23 dispatch-registered
  algorithms into a single weighted similarity score in [0.0, 1.0].
  Threshold-driven Match. Auto-normalised weights by default. Immutable
  after construction; safe for concurrent use without locks. Background
  is documented in docs/requirements.md §8 and docs/scorer.md.

  @scorer @default
  Scenario: Default scorer matches identifier-style pair
    # CONTEXT.md §7 class 1: happy path on snake_case-vs-camelCase
    # identifier pair. DefaultScorer uses the 6-algorithm composition
    # with 0.85 threshold; user_id / userId scores well above 0.85.
    Given I construct the default Scorer
    When I score "user_id" and "userId" with the Scorer
    Then the Scorer match result should be true

  @scorer @default
  Scenario: Default scorer rejects dissimilar inputs
    # CONTEXT.md §7 class 2: below-threshold pair. is_deleted /
    # is_active are similar in shape but opposite in meaning;
    # composite score sits below 0.85.
    Given I construct the default Scorer
    When I score "is_deleted" and "is_active" with the Scorer
    Then the Scorer match result should be false

  @scorer @custom
  Scenario: Single-algorithm Scorer composes correctly
    # CONTEXT.md §7 class 3: minimum viable composition — one
    # algorithm + WithThreshold. Levenshtein-only at threshold 0.5
    # over kitten / sitting (Levenshtein distance 3 / max-len 7 ≈
    # 0.571 — just above the 0.5 threshold).
    Given I construct a Scorer with Levenshtein weight 1.0 and threshold 0.5
    When I score "kitten" and "sitting" with the Scorer
    Then the Scorer match result should be true

  @scorer @custom
  Scenario: Two-algorithm weighted Scorer composes correctly
    # CONTEXT.md §7 class 4: weighted composite of two algorithms.
    # 50/50 split of Levenshtein and JaroWinkler. The composite is
    # the weighted average; for an identical pair both algorithms
    # return 1.0 so the composite is exactly 1.0.
    Given I construct a Scorer with Levenshtein weight 0.5 and JaroWinkler weight 0.5 and threshold 0.7
    When I score "hello" and "hello" with the Scorer
    Then the Scorer composite score should be exactly 1.0
    And the Scorer match result should be true

  @scorer @custom
  Scenario: WithoutAlgorithm removes the algorithm
    # CONTEXT.md §7 class 5: DefaultScorerOptions composed with
    # WithoutAlgorithm. The resulting Scorer's Algorithms() slice
    # excludes the removed AlgoID; ScoreAll's map does not contain
    # the removed key.
    Given I construct the default Scorer without DoubleMetaphone
    When I list the Scorer algorithms
    Then the Scorer algorithms should not include DoubleMetaphone

  @scorer @custom
  Scenario: Duplicate WithAlgorithm calls use the latest weight
    # CONTEXT.md §7 class 6: last-write-wins for duplicate AlgoID.
    # Two WithAlgorithm(AlgoLevenshtein, w) calls — only the latter
    # weight survives. After auto-normalisation (one entry), the
    # surviving entry's weight is 1.0.
    Given I construct a Scorer with duplicate Levenshtein weights 0.3 and 0.7 and threshold 0.5
    When I list the Scorer algorithms
    Then the Scorer should have exactly 1 algorithm
    And the algorithm should be AlgoLevenshtein with weight 1.0

  @scorer @custom
  Scenario: WithoutNormalisation preserves raw bytes
    # CONTEXT.md §7 class 7: XMLParser / xml_parser score lower under
    # WithoutNormalisation than under the default Scorer (case and
    # underscore differences are not erased before the algorithms
    # run). The default Scorer matches this pair; the no-norm Scorer
    # may or may not depending on the composite score relative to
    # threshold — we assert the relative ordering.
    Given I construct the default Scorer without normalisation
    When I score "XMLParser" and "xml_parser" with the Scorer
    And I record the no-normalisation Scorer composite score
    And I score the same pair with the default Scorer
    Then the no-normalisation composite should be less than the default composite

  @scorer @errors
  Scenario: NewScorer without WithThreshold returns ErrMissingThreshold
    # CONTEXT.md §7 class 8: threshold is mandatory at construction
    # time per §2 LOCKED. A consumer who passes algorithms but no
    # WithThreshold gets ErrMissingThreshold.
    When I attempt to construct a Scorer with Levenshtein weight 1.0 and no threshold
    Then constructing the Scorer should return ErrMissingThreshold

  @scorer @errors
  Scenario: NewScorer without algorithms returns ErrEmptyScorer
    # CONTEXT.md §7 class 9: at least one WithAlgorithm option is
    # required. A consumer who passes WithThreshold but no algorithms
    # gets ErrEmptyScorer.
    When I attempt to construct a Scorer with threshold 0.5 and no algorithms
    Then constructing the Scorer should return ErrEmptyScorer

  @scorer @errors
  Scenario: WithAlgorithm with negative weight returns ErrInvalidWeight
    # CONTEXT.md §7 class 10: WithAlgorithm validates weight > 0.
    # A negative weight returns ErrInvalidWeight at option-
    # application time (short-circuits NewScorer before any other
    # validation runs).
    When I attempt to construct a Scorer with Levenshtein weight -0.5 and threshold 0.5
    Then constructing the Scorer should return ErrInvalidWeight

  @scorer @concurrency
  Scenario: Concurrent Score calls return identical results
    # CONTEXT.md §7 class 11: the Scorer is immutable after
    # construction and safe for concurrent use without locks. 100
    # goroutines call Score on the same pair; all return the same
    # value (deterministic computation on shared input). goleak in
    # bdd_test.go's TestMain catches any goroutine leak.
    Given I construct the default Scorer
    When I call Score on "user_id" and "userId" from 100 goroutines simultaneously
    Then all 100 goroutine results should be byte-identical

  @scorer @scoreall
  Scenario: ScoreAll returns map keyed by AlgoID
    # CONTEXT.md §7 class 12: ScoreAll returns map[AlgoID]float64
    # (typed enum keys, per CONTEXT.md §1 SPEC OVERRIDE). The
    # scenario is a runtime sanity check; the compile-time type
    # safety is asserted by the Go compiler at scorer_steps.go
    # compile time.
    Given I construct the default Scorer
    When I call ScoreAll on "user_id" and "userId" with the Scorer
    Then the ScoreAll map should contain AlgoDamerauLevenshteinOSA
    And the ScoreAll map should contain AlgoDoubleMetaphone
    And the ScoreAll map should not contain AlgoCosine
