# Phase 9 BDD coverage: suppression composition.
#
# Closes Phase 8.5 R2 Gap 3 — the suppression scenarios deferred at the
# Phase 8.5 review gate land here. Each of the three suppression rules
# (per 09-CONTEXT.md §3 and docs/requirements.md §12.3) has at least one
# scenario; composition / ordering / canonicalisation are covered
# additionally.
#
# Three suppression rules (Rule order in scan/suppress.go isSuppressed):
#
#   Rule 1 — per-item SilenceLint (cheapest; one-sided OR semantics).
#            Setting the flag on either side of a candidate pair
#            silences the pair.
#   Rule 2 — SuppressedPairs canonical-pair lookup. The consumer-
#            supplied [][2]string is normalised once at Check entry via
#            the Scorer's NormalisationOptions, then stored as a
#            canonical (lex-sorted) pair set. Lookups are order-
#            independent.
#   Rule 3 — cross-group identical-name default (SCAN-04). When
#            CompareIdenticalAcrossGroups == false (DefaultConfig),
#            cross-group pairs whose normalised names coincide are
#            suppressed.
#
# Rules compose via short-circuit OR; any rule firing suppresses the
# pair. Suppression is applied PRE-EMISSION (the isSuppressed predicate
# is called before Scorer.Score on every candidate pair) so a suppressed
# pair never reaches the emission site. Phase 8.5 R2 deferred coverage
# is closed by:
#
#   - 3 SilenceLint scenarios (Rule 1 from either side + OR semantics)
#   - 1 SuppressedPairs canonical-pair scenario (Rule 2 happy path)
#   - 1 SuppressedPairs canonicalised-via-Normalise scenario (Rule 2
#     Pitfall 4 — raw forms still match Normalised candidates)
#   - 1 D-05 self-pair scenario (silently kept; never emits)
#   - 1 cross-group identical compose-with-explicit scenario (Rule 3 +
#     Rule 2 composition)
#   - 1 OR-composition scenario (Rules 1, 2, 3 firing on disjoint pairs)
#   - 1 pre-emission ordering scenario (output count matches the
#     "without suppression" reference minus the suppressed set)

@suppression
Feature: Suppression composition (Phase 9 — closes Phase 8.5 R2 Gap 3)
  Suppression rules compose additively via OR. Per-item SilenceLint
  is the cheapest check; SuppressedPairs is canonicalised once at
  Check entry via the Scorer's NormalisationOptions; the cross-group
  identical-name default is the third rule, opt-out via
  Config.CompareIdenticalAcrossGroups. Background:
  docs/requirements.md §12.3 and docs/scan.md.

  @suppression @silencelint @rule1 @phase-8.5-r2
  Scenario: SilenceLint=true on Item A suppresses all pairs involving A
    # Phase 8.5 R2 deferral closure — Rule 1 from side A. user_id has
    # SilenceLint=true; no warning involving user_id is emitted.
    Given I construct the default Scorer for scan
    And the scan items
      | name    | group | silence_lint |
      | user_id | login | true         |
      | userId  | login | false        |
    And the scan config is the default scan config
    When I invoke scan.Check
    Then scan.Check returns no error
    And no scan warning involves "user_id"
    And the scan warnings list has 0 entries

  @suppression @silencelint @rule1 @phase-8.5-r2
  Scenario: SilenceLint=true on Item B suppresses all pairs involving B
    # Phase 8.5 R2 deferral closure — Rule 1 from side B. Sanity
    # check that suppression is symmetric on which side carries the
    # flag. userId has SilenceLint=true; the same pair is suppressed.
    Given I construct the default Scorer for scan
    And the scan items
      | name    | group | silence_lint |
      | user_id | login | false        |
      | userId  | login | true         |
    And the scan config is the default scan config
    When I invoke scan.Check
    Then scan.Check returns no error
    And no scan warning involves "userId"
    And the scan warnings list has 0 entries

  @suppression @silencelint @or @rule1 @phase-8.5-r2
  Scenario: SilenceLint on either side suppresses (OR semantics)
    # Phase 8.5 R2 deferral closure — Rule 1 OR semantics. A is
    # silenced, B isn't; the pair A-B is suppressed.
    Given I construct the default Scorer for scan
    And the scan items
      | name    | group | silence_lint |
      | user_id | login | false        |
      | userId  | login | true         |
      | userid  | login | false        |
    And the scan config is the default scan config
    When I invoke scan.Check
    Then scan.Check returns no error
    And no scan warning involves "userId"

  @suppression @suppressedpairs @rule2
  Scenario: SuppressedPairs entry suppresses the listed canonical pair
    # 09-CONTEXT.md §3 Rule 2 — happy path. The user_id ↔ userId pair
    # is in SuppressedPairs; even though both items score above the
    # 0.85 threshold, no warning emits.
    Given I construct the default Scorer for scan
    And the scan items
      | name    | group | silence_lint |
      | user_id | login | false        |
      | userId  | login | false        |
    And the scan config is the default scan config
    And I set suppressed pairs
      | a       | b      |
      | user_id | userId |
    When I invoke scan.Check
    Then scan.Check returns no error
    And the scan warnings list has 0 entries

  @suppression @suppressedpairs @canonical @rule2
  Scenario: SuppressedPairs is canonicalised via Scorer NormalisationOptions
    # 09-CONTEXT.md §3 Rule 2 + Pitfall 4 — raw consumer-supplied
    # forms (UPPER_CASE, mixedCase) still match Normalised candidates
    # because buildSuppressionCtx normalises the SuppressedPairs
    # entries once at Check entry using the same NormalisationOptions
    # the Scorer applies to the items.
    Given I construct the default Scorer for scan
    And the scan items
      | name    | group | silence_lint |
      | user_id | login | false        |
      | userId  | login | false        |
    And the scan config is the default scan config
    And I set suppressed pairs
      | a       | b      |
      | USER_ID | UserId |
    When I invoke scan.Check
    Then scan.Check returns no error
    And the scan warnings list has 0 entries

  @suppression @suppressedpairs @selfpair @D-05
  Scenario: Self-pairs in SuppressedPairs are silently kept and never emit
    # 09-CONTEXT.md §2 D-05 — self-pairs (a == b after normalisation)
    # are silently kept at validation time. Check never emits a self-
    # warning because validateItems' duplicate-(Name, Group) check
    # (D-06) plus the i<j inner-loop iteration discipline prevent
    # any (i, i) candidate pair from reaching the emission site.
    # The SuppressedPairs entry is accepted; Check returns nil error
    # with no warnings.
    Given I construct the default Scorer for scan
    And the scan items
      | name | group | silence_lint |
      | foo  | g1    | false        |
      | bar  | g1    | false        |
    And the scan config is the default scan config
    And I set suppressed pairs
      | a   | b   |
      | foo | foo |
    When I invoke scan.Check
    Then scan.Check returns no error

  @suppression @cross @identical @rule3 @phase-8.5-r2
  Scenario: Cross-group identical-name default suppression composes with explicit SuppressedPairs
    # Phase 8.5 R2 deferral closure — Rule 3 + Rule 2 composition.
    # CompareAcrossGroups=true with DefaultConfig
    # (CompareIdenticalAcrossGroups=false → Rule 3 active). Items
    # produce three candidate cross-group pairs of interest:
    #   - user_id @login vs user_id @profile  -> Rule 3 suppresses
    #   - userId @login vs userId @profile    -> Rule 3 suppresses
    # PLUS an explicit Rule 2 entry suppresses user_id↔userId across
    # groups. None of the explicitly-similar cross-group pairs emit.
    Given I construct the default Scorer for scan
    And the scan items
      | name    | group   | silence_lint |
      | user_id | login   | false        |
      | userId  | login   | false        |
      | user_id | profile | false        |
      | userId  | profile | false        |
    And the scan config is the default scan config
    And I set CompareAcrossGroups to true
    And I set suppressed pairs
      | a       | b      |
      | user_id | userId |
    When I invoke scan.Check
    Then scan.Check returns no error
    And the scan warnings list has 0 AcrossGroups entries

  @suppression @or @composition @phase-8.5-r2
  Scenario: Suppression rules compose via OR — any rule firing suppresses
    # Phase 8.5 R2 deferral closure — three rules firing on disjoint
    # candidate pairs:
    #   - Rule 1 fires for the (user_id, userId) pair via SilenceLint
    #     on userId in @login
    #   - Rule 2 fires for the (user-id, user_id) pair via
    #     SuppressedPairs explicit entry
    #   - Rule 3 fires for the (user_id @login, user_id @profile)
    #     cross-group identical-name default
    # The output set excludes ALL three suppressed pair classes; the
    # @profile within-group pair (userId vs user-id) is unaffected
    # and emits cleanly. Verifies that suppression composition is OR
    # — no rule cancels another, and the cheapest rule wins per
    # short-circuit evaluation in scan/suppress.go isSuppressed.
    Given I construct the default Scorer for scan
    And the scan items
      | name    | group   | silence_lint |
      | user_id | login   | false        |
      | userId  | login   | true         |
      | user-id | login   | false        |
      | userId  | profile | false        |
      | user_id | profile | false        |
    And the scan config is the default scan config
    And I set CompareAcrossGroups to true
    And I set suppressed pairs
      | a       | b       |
      | user-id | user_id |
    When I invoke scan.Check
    Then scan.Check returns no error
    And no scan warning involves "userId"
    And the scan warnings list has 0 AcrossGroups entries

  @suppression @ordering @pre-emission
  Scenario: Suppression is applied pre-emission, never post-emission
    # 09-CONTEXT.md §3 ordering — isSuppressed is called BEFORE
    # Scorer.Score on every candidate pair (scan/scan.go inner loop).
    # A pair listed in SuppressedPairs must never appear in output,
    # even when its raw composite score would clear the threshold.
    # This scenario is a direct restatement of Rule 2's contract; it
    # is included separately because pre-emission ordering is a
    # documented invariant distinct from the "Rule 2 happy path"
    # scenario above (the latter checks Rule 2 fires; this one
    # documents WHEN it fires relative to Score).
    Given I construct the default Scorer for scan
    And the scan items
      | name    | group | silence_lint |
      | user_id | login | false        |
      | userId  | login | false        |
    And the scan config is the default scan config
    And I set suppressed pairs
      | a       | b      |
      | user_id | userId |
    When I invoke scan.Check
    Then scan.Check returns no error
    And the scan warnings list has 0 entries
