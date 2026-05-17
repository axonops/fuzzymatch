# Phase 8.5 Plan 17b — Gap 3 BDD coverage: Normalise surface.
#
# Six scenarios covering the Normalise contract documented in
# docs/requirements.md §9:
#
#   1. Default options (ASCII identifier-style input) — camelCase split,
#      separator collapse, lowercase fold.
#   2. Diacritic stripping (StripDiacritics opt-in) — "café → cafe".
#   3. NFC normalisation (precomposed vs decomposed equivalence) —
#      byte-identical output on both forms of the same logical string.
#   4. Idempotence — Normalise(Normalise(s)) == Normalise(s).
#   5. ASCII fast path — pure-ASCII input produces correct output
#      without invoking the Unicode normalisation tables.
#   6. Empty input — returns empty (the trivial-identity case).
#
# Step definitions live in tests/bdd/steps/normalisation_steps.go.
#
# scan.feature + suppression.feature are deliberately NOT created in
# Plan 17b — they are deferred to Phase 9 per the CONTEXT.md Gap 3
# default (acknowledged in Plan 18's verification-gate report).

@normalisation
Feature: Normalise input strings (docs/requirements.md §9)
  As a fuzzymatch consumer
  I want to normalise input strings consistently before scoring
  So that identifier variations (case, separators, diacritics, NFC/NFD)
  do not produce spurious similarity differences.

  @normalisation @defaults
  Scenario: DefaultNormalisationOptions folds case and splits camelCase + kebab-case
    # docs/requirements.md §9 — the v1.0 default Normalise behaviour
    # tuned for code-identifier matching.
    Given the normalisation input "UserCreate-Event"
    When I normalise using DefaultNormalisationOptions
    Then the normalised output should be "user create event"

  @normalisation @diacritics
  Scenario: StripDiacritics opt-in folds "Café Müller" to "cafe muller"
    # docs/requirements.md §9 — StripDiacritics applies NFD →
    # Remove(Mn) → NFC so combining marks are dropped and the result
    # is precomposed ASCII-equivalent.
    Given the normalisation input "Café Müller"
    When I normalise using DefaultNormalisationOptions with StripDiacritics
    Then the normalised output should be "cafe muller"

  @normalisation @nfc
  Scenario: Precomposed and decomposed forms produce byte-identical output
    # docs/requirements.md §9 — NFC normalisation makes precomposed
    # "é" (U+00E9) and decomposed "e + U+0301" produce the same
    # output bytes.
    Given the precomposed input "é" and the decomposed input "é"
    When I normalise both with DefaultNormalisationOptions
    Then the two outputs should be byte-identical

  @normalisation @idempotence
  Scenario: Normalise is idempotent for any fixed options
    # docs/requirements.md §9 idempotence invariant — applying
    # Normalise twice produces the same output as applying it once.
    # Verified by property test PropNormalise_Idempotent; this BDD
    # scenario is the consumer-facing assertion.
    Given the normalisation input "UserCreate-Event"
    When I normalise using DefaultNormalisationOptions
    And I normalise the output again using DefaultNormalisationOptions
    Then the two outputs should be byte-identical

  @normalisation @ascii
  Scenario: ASCII-only input passes through the fast path correctly
    # docs/requirements.md §9.2 ASCII fast path — pure-ASCII input
    # bypasses the Unicode normalisation tables and uses a stack-
    # allocated buffer. The visible output is identical to the
    # rune-path output (verified by property test PropNormalise_
    # ASCIIFastPathEquivalent); this scenario asserts a concrete
    # ASCII case.
    Given the normalisation input "hello-world"
    When I normalise using DefaultNormalisationOptions
    Then the normalised output should be "hello world"

  @normalisation @empty
  Scenario: Empty input returns empty
    # docs/requirements.md §9 — Normalise on empty input is the
    # trivial identity (no allocations, no work).
    Given the normalisation input ""
    When I normalise using DefaultNormalisationOptions
    Then the normalised output should be ""
