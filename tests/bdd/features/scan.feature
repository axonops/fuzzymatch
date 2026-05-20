# Phase 9 BDD coverage: collection-scan sub-package (scan.Check).
#
# Mandatory scenario classes per 09-CONTEXT.md (one per locked decision +
# sort determinism + bucket smoke). The numbered list below is the
# bdd-scenario-reviewer gate: every locked decision (D-03, D-04, D-06)
# and every public-API surface has at least one scenario.
#
#   1. Within-group happy path        -> "Within-group scan finds similar names"
#   2. Within-group below threshold   -> "Within-group scan respects threshold"
#   3. Cross-group boost arithmetic   -> "Cross-group scan applies threshold boost from DefaultConfig"
#   4. Cross-group identical default  -> "Cross-group identical-name pairs are suppressed by default"
#   5. Cross-group identical opt-in   -> "Cross-group identical-name pairs are emitted when CompareIdenticalAcrossGroups is true"
#   6. D-03 empty-Name validation     -> "Empty Item.Name returns ErrInvalidItem with errors.Join across all offending indices"
#   7. D-06 duplicate validation      -> "Duplicate (Name, Group) returns ErrInvalidItem with errors.Join across all duplicate indices"
#   8. D-04 boost validation          -> "NaN CrossGroupThresholdBoost returns ErrInvalidConfig"
#   9. ErrNilScorer surface           -> "Nil Scorer returns ErrNilScorer"
#  10. Sort determinism (5-tuple key) -> "Output is sorted by (Kind, NameA, NameB, GroupA, GroupB) deterministically across runs"
#  11. Bucket equivalence smoke       -> "Bucket optimisation produces output identical to a second invocation for the same input"
#
# D-07 boundary note: scan.Check does NOT internally invoke
# fuzzymatch.Validate — consumers compose Validate-then-Check explicitly.
# The separation is documented in docs/scan.md and surfaces here as the
# absence of any Diagnostics field on scan.Warning (no runtime test
# required — the absence is a compile-time guarantee).

@scan
Feature: Collection scan (Phase 9)
  The scan sub-package finds pairs of similar names in a collection
  with within-group and cross-group passes, deterministic output
  ordering, a validation surface (D-03 empty-Name, D-04 boost,
  D-06 duplicate-(Name, Group)), and a token-bucket optimisation
  proven equivalent to the naive O(N²) implementation. Background
  is documented in docs/requirements.md §12 and docs/scan.md.

  @scan @within @happy
  Scenario: Within-group scan finds similar names
    # 09-CONTEXT.md happy-path class — snake_case-vs-camelCase
    # identifier pair scoring above the DefaultScorer 0.85 threshold
    # in the within-group pass. Sort canonicalises NameA / NameB so
    # for the (user_id, userId) pair NameA is "userId" (lex-smaller).
    Given I construct the default Scorer for scan
    And the scan items
      | name    | group | silence_lint |
      | user_id | login | false        |
      | userId  | login | false        |
      | other   | login | false        |
    And the scan config is the default scan config
    When I invoke scan.Check
    Then scan.Check returns no error
    And the scan warnings include a WithinGroup pair with names "userId" and "user_id"

  @scan @within @threshold
  Scenario: Within-group scan respects threshold
    # 09-CONTEXT.md threshold class — two within-group items whose
    # composite similarity sits below 0.85 (mirrors the scorer.feature
    # "is_deleted vs is_active" sub-threshold pair). No warnings emit.
    Given I construct the default Scorer for scan
    And the scan items
      | name       | group | silence_lint |
      | is_deleted | login | false        |
      | is_active  | login | false        |
    And the scan config is the default scan config
    When I invoke scan.Check
    Then scan.Check returns no error
    And the scan warnings list has 0 entries

  @scan @cross @boost
  Scenario: Cross-group scan applies threshold boost from DefaultConfig
    # 09-CONTEXT.md cross-group class — DefaultConfig bakes
    # CrossGroupThresholdBoost = 0.05 (D-04). With DefaultScorer
    # Threshold = 0.85, the effective cross-group threshold is
    # min(1.0, 0.85 + 0.05) = 0.90. The step asserts the documented
    # boost arithmetic; runtime gating is exercised by scan_test.go.
    Given I construct the default Scorer for scan
    And the scan items
      | name    | group   | silence_lint |
      | user_id | login   | false        |
      | userId  | profile | false        |
    And the scan config is the default scan config
    And I set CompareAcrossGroups to true
    When I invoke scan.Check
    Then scan.Check returns no error
    And the effective cross-group threshold equals min(1.0, Threshold + Boost)

  @scan @cross @identical @default-suppressed
  Scenario: Cross-group identical-name pairs are suppressed by default
    # 09-CONTEXT.md SCAN-04 + Rule 3 — DefaultConfig sets
    # CompareIdenticalAcrossGroups = false. Identical names across
    # groups never emit a cross-group warning.
    Given I construct the default Scorer for scan
    And the scan items
      | name    | group   | silence_lint |
      | user_id | login   | false        |
      | user_id | profile | false        |
    And the scan config is the default scan config
    And I set CompareAcrossGroups to true
    When I invoke scan.Check
    Then scan.Check returns no error
    And the scan warnings list has 0 AcrossGroups entries

  @scan @cross @identical @opt-in
  Scenario: Cross-group identical-name pairs are emitted when CompareIdenticalAcrossGroups is true
    # 09-CONTEXT.md SCAN-04 opt-in — flipping the toggle disables
    # Rule 3, so identical-name pairs DO emit cross-group warnings.
    Given I construct the default Scorer for scan
    And the scan items
      | name    | group   | silence_lint |
      | user_id | login   | false        |
      | user_id | profile | false        |
    And the scan config is the default scan config
    And I set CompareAcrossGroups to true
    And I set CompareIdenticalAcrossGroups to true
    When I invoke scan.Check
    Then scan.Check returns no error
    And the scan warnings list has 1 AcrossGroups entries
    And the scan warnings include a AcrossGroups pair with names "user_id" and "user_id"

  @scan @validation @empty-name @D-03
  Scenario: Empty Item.Name returns ErrInvalidItem with errors.Join across all offending indices
    # 09-CONTEXT.md D-03 — empty Item.Name is rejected; every
    # offending index is collected via errors.Join. The error string
    # mentions both index 1 and index 3. errors.Is discrimination
    # against the sentinel still resolves because Go 1.20+ errors.Is
    # walks Unwrap() []error.
    Given I construct the default Scorer for scan
    And the scan items
      | name    | group | silence_lint |
      | ok      | g1    | false        |
      | (empty) | g1    | false        |
      | ok2     | g1    | false        |
      | (empty) | g2    | false        |
    And the scan config is the default scan config
    When I invoke scan.Check
    Then scan.Check returns an error matching scan.ErrInvalidItem
    And the error mentions index 1 and index 3

  @scan @validation @duplicate @D-06
  Scenario: Duplicate (Name, Group) returns ErrInvalidItem with errors.Join across all duplicate indices
    # 09-CONTEXT.md D-06 — duplicate (Name, Group) pairs collide
    # with the SCAN-05 sort-key uniqueness invariant. Validation
    # collects every duplicate index via errors.Join; the error
    # string mentions both index 1 and index 2 (the first-seen
    # index 0 is referenced as the colliding original).
    Given I construct the default Scorer for scan
    And the scan items
      | name    | group | silence_lint |
      | user_id | login | false        |
      | user_id | login | false        |
      | user_id | login | false        |
    And the scan config is the default scan config
    When I invoke scan.Check
    Then scan.Check returns an error matching scan.ErrInvalidItem
    And the error mentions index 1 and index 2

  @scan @validation @boost @D-04
  Scenario: NaN CrossGroupThresholdBoost returns ErrInvalidConfig
    # 09-CONTEXT.md D-04 — Config.CrossGroupThresholdBoost is
    # strict-validated at Check entry: NaN, ±Inf, < 0, > 1 all
    # reject with ErrInvalidConfig.
    Given I construct the default Scorer for scan
    And the scan items
      | name    | group | silence_lint |
      | user_id | login | false        |
      | userId  | login | false        |
    And the scan config is the default scan config
    And the scan config has CrossGroupThresholdBoost set to NaN
    When I invoke scan.Check
    Then scan.Check returns an error matching scan.ErrInvalidConfig

  @scan @validation @nil-scorer
  Scenario: Nil Scorer returns ErrNilScorer
    # 09-CONTEXT.md validation pipeline P1 — the cheapest, fail-fast
    # gate. cfg.Scorer == nil short-circuits before Config / Items /
    # SuppressedPairs validation. The error matches ErrNilScorer by
    # sentinel identity (not errors.Join — there's nothing to collect).
    Given the scan config has a nil Scorer
    And the scan items
      | name    | group | silence_lint |
      | user_id | login | false        |
      | userId  | login | false        |
    When I invoke scan.Check
    Then scan.Check returns an error matching scan.ErrNilScorer

  @scan @determinism @sort
  Scenario: Output is sorted by (Kind, NameA, NameB, GroupA, GroupB) deterministically across runs
    # 09-CONTEXT.md determinism + SCAN-05 — Check sorts before
    # returning and the result is identical across runs.
    # Two invocations on the same input must marshal to byte-
    # identical JSON; the slice is also lex-ordered on the 5-tuple
    # sort key.
    Given I construct the default Scorer for scan
    And the scan items
      | name    | group | silence_lint |
      | user_id | login | false        |
      | userId  | login | false        |
      | userid  | login | false        |
      | user-id | login | false        |
    And the scan config is the default scan config
    When I invoke scan.Check twice
    Then the two scan warnings slices are byte-identical
    And the scan warnings are sorted by (Kind, NameA, NameB, GroupA, GroupB)

  @scan @bucket @smoke
  Scenario: Bucket optimisation produces output identical to a second invocation for the same input
    # 09-CONTEXT.md SCAN-02 / D-08 — the load-bearing
    # PropCheck_BucketEquivalentToNaive property test in
    # scan/props_test.go proves bucket == naive across random inputs;
    # this BDD scenario smokes idempotence at the public-API layer
    # (Check returns identical output on repeated invocation of the
    # same input, regardless of whether the bucket or naive path
    # ran). Small group size (≤ bucketThreshold) — naive path runs
    # on both invocations.
    Given I construct the default Scorer for scan
    And the scan items
      | name    | group | silence_lint |
      | user_id | login | false        |
      | userId  | login | false        |
      | userid  | login | false        |
    And the scan config is the default scan config
    When I invoke scan.Check
    Then scan.Check returns no error
    And the scan warnings match the naive-loop reference output
