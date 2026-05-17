# Phase 8.5 Plan 13 BDD coverage: Validate public surface.
#
# One scenario per WarnKind constant (5 total) per VALIDATE-01..VALIDATE-05
# and docs/requirements.md §11.5 Per-WarnKind semantics:
#
#   WarnEmptyInput              -> "Empty input produces WarnEmptyInput"
#   WarnUnequalLength           -> "Unequal length input produces WarnUnequalLength for Hamming"
#   WarnNoTokensAfterNormalise  -> "Token-tier input that normalises to no tokens produces WarnNoTokensAfterNormalise"
#   WarnAllNonASCIIDropped      -> "Non-ASCII input to ASCII-only algorithms produces WarnAllNonASCIIDropped"
#   WarnPathologicallyLargeInput -> "Pathologically large input produces WarnPathologicallyLargeInput"
#
# Plus a clean-input scenario asserting nil-return (the VALIDATE-01
# nil-vs-empty-slice contract) and a determinism scenario asserting
# two-call byte-equality (T-08.5-26 mitigation gate).

@validate
Feature: Input-quality validation (Phase 8.5 Q4)
  As a fuzzymatch consumer
  I want to detect input-quality problems before scoring
  So that I can degrade gracefully on low-information inputs
  Background is documented in docs/requirements.md §11.5 and
  .planning/REQUIREMENTS.md VALIDATE-01..VALIDATE-06.

  @validate @empty
  Scenario: Empty input produces WarnEmptyInput
    # docs/requirements.md §11.5 — WarnEmptyInput affects every
    # algorithm; emitted once with AlgoIDAny scope (cross-cutting).
    Given the validation input pair "" and "abc"
    When I call Validate
    Then the warnings include WarnEmptyInput

  @validate @hamming
  Scenario: Unequal length input produces WarnUnequalLength for Hamming
    # docs/requirements.md §11.5 — WarnUnequalLength documents the
    # silent-max policy for Hamming; emitted scoped to AlgoHamming.
    Given the validation input pair "abc" and "abcd"
    When I call Validate
    Then the warnings include WarnUnequalLength scoped to Hamming

  @validate @tokens
  Scenario: Token-tier input that normalises to no tokens produces WarnNoTokensAfterNormalise
    # docs/requirements.md §11.5 — separator-only input ("---") tokenises
    # to an empty list under DefaultTokeniseOptions; every token-tier
    # algorithm (MongeElkan, TokenSortRatio, TokenSetRatio, PartialRatio,
    # TokenJaccard) emits the warning scoped to itself.
    Given the validation input pair "hello" and "---"
    When I call Validate
    Then the warnings include WarnNoTokensAfterNormalise scoped to TokenSortRatio

  @validate @ascii
  Scenario: Non-ASCII input to ASCII-only algorithms produces WarnAllNonASCIIDropped
    # docs/requirements.md §11.5 — pure non-ASCII input collapses to
    # empty after ASCII filtering on Strcmp95, Soundex, DoubleMetaphone,
    # NYSIIS, MRA. Five warnings emitted, one scoped to each.
    Given the validation input pair "中文" and "日本語"
    When I call Validate
    Then the warnings include WarnAllNonASCIIDropped scoped to Soundex

  @validate @large
  Scenario: Pathologically large input produces WarnPathologicallyLargeInput
    # docs/requirements.md §11.5 — input above the 64 KiB threshold
    # risks O(m·n) DP-table cost on the quadratic algorithms. Emitted
    # once with AlgoIDAny scope (cross-cutting).
    Given a validation input pair of length 70000 each
    When I call Validate
    Then the warnings include WarnPathologicallyLargeInput

  @validate @clean
  Scenario: Clean input returns nil warnings
    # VALIDATE-01 nil-vs-empty-slice contract: two short well-formed
    # ASCII inputs of equal length with valid tokens produce no
    # warnings; the return value is nil (not []Warning{}).
    Given the validation input pair "hello" and "world"
    When I call Validate
    Then the warnings should be nil

  @validate @determinism
  Scenario: Two calls return identical warnings
    # T-08.5-26 mitigation gate — sort.SliceStable with complete sort
    # key (Algorithm, Kind) guarantees byte-equality across calls.
    Given the validation input pair "" and "abc"
    When I call Validate twice
    Then both warnings slices should be identical
