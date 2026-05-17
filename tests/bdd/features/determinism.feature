# Phase 8.5 Plan 17b — Gap 3 BDD coverage: determinism contract.
#
# Six scenarios covering the determinism guarantees documented in
# docs/requirements.md §13:
#
#   1. Score determinism — two consecutive Score calls with the same
#      inputs produce byte-identical float64 output.
#   2. ScoreAll determinism — two consecutive ScoreAll calls produce
#      the same key set with the same per-key float64 values.
#   3. Algorithm-function determinism — direct LevenshteinScore calls
#      produce byte-identical output across invocations.
#   4. Jaro-Winkler determinism (load-bearing for Phase 5 float-
#      reduction-order check) — repeated calls produce identical
#      output.
#   5. Map-iteration discipline — Algorithms() returns a slice in
#      stable AlgoID-ascending order (no map iteration on output).
#   6. Cosine determinism (load-bearing — the cross-platform golden-
#      file gate algorithm per docs/requirements.md §13.3) — repeated
#      calls on the same inputs produce byte-identical output.
#
# Step definitions live in tests/bdd/steps/determinism_steps.go.
#
# Golden-file byte-identity across the 5-platform CI matrix is
# verified by TestGolden_* tests in algorithms_golden_test.go;
# this BDD feature exercises the consumer-facing same-process
# determinism guarantee.

@determinism
Feature: Deterministic output (docs/requirements.md §13)
  As a fuzzymatch consumer
  I want repeated identical scoring calls to return identical results
  So that downstream caches, dashboards, and reproducible computations
  can rely on the library producing stable output.

  @determinism @scorer
  Scenario: DefaultScorer.Score is deterministic across repeat calls
    # docs/requirements.md §13.2 — composite score is bitwise stable
    # across calls within a process AND across the 5-platform CI
    # matrix (the latter is verified by the golden-file test).
    Given the determinism input pair "user_create" and "userCreate"
    When I call DefaultScorer.Score twice
    Then both scores should be byte-identical

  @determinism @scoreall
  Scenario: DefaultScorer.ScoreAll is deterministic across repeat calls
    # docs/requirements.md §13.4 — map contents are deterministic
    # (the same keys with the same values); iteration order is
    # randomised per Go map semantics but consumer-visible contents
    # are stable.
    Given the determinism input pair "user_create" and "userCreate"
    When I call DefaultScorer.ScoreAll twice
    Then both per-algorithm score maps should contain the same key-value pairs

  @determinism @algorithm
  Scenario: LevenshteinScore is deterministic across repeat calls
    # docs/requirements.md §13.1 — Algorithm score stability. Direct
    # algorithm-function calls must produce byte-identical output
    # across runs.
    Given the determinism input pair "kitten" and "sitting"
    When I call LevenshteinScore twice
    Then both scores should be byte-identical

  @determinism @jarowinkler
  Scenario: JaroWinklerScore is deterministic across repeat calls
    # docs/requirements.md §13.1 — Jaro-Winkler is a load-bearing
    # float-reduction-order check (the prefix-bonus multiplication
    # path exercises the (x*y) + z parenthesisation discipline).
    Given the determinism input pair "MARTHA" and "MARHTA"
    When I call JaroWinklerScore twice
    Then both scores should be byte-identical

  @determinism @algorithms-order
  Scenario: Scorer.Algorithms returns slice in AlgoID-ascending order
    # docs/requirements.md §13.4 — no map iteration on output paths.
    # Algorithms() returns a fresh slice in AlgoID-ascending order
    # so consumers can rely on stable iteration without sorting.
    When I call DefaultScorer.Algorithms
    Then the returned slice should be in AlgoID-ascending order

  @determinism @cosine
  Scenario: CosineScore is deterministic across repeat calls
    # docs/requirements.md §13.3 — Cosine is the LOAD-BEARING cross-
    # platform float-determinism algorithm. The intersection-key sort
    # plus the math.Sqrt-only normalisation produce byte-identical
    # output across linux/amd64, linux/arm64, darwin/amd64, darwin/
    # arm64, and windows/amd64 (verified by the golden file).
    Given the determinism input pair "abcdef" and "abcdgh"
    When I call CosineScore twice
    Then both scores should be byte-identical
