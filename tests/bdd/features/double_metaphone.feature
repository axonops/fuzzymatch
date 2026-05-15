# Primary source: Philips, L. (2000). "The double-metaphone search algorithm."
# C/C++ Users Journal, 18(6):38-43.
# Public-domain C reference (rule-table provenance):
#   https://github.com/SWI-Prolog/packages-nlp/blob/master/double_metaphone.c
# Cross-validation: oubiwann/metaphone==0.6 (BSD-3-Clause).
#
# Variant choice (LOAD-BEARING): Philips 2000 canonical 4-char keys, charset [A-Z0].
#   - DoubleMetaphoneKeys("Schmidt") = ("XMT", "SMT")
#   - DoubleMetaphoneKeys("Smith") = ("SM0", "XMT")
#   - DoubleMetaphoneScore("Schmidt", "Smith") = 1.0 (XMT cross-match)
#   - DoubleMetaphoneKeys("Catherine") = DoubleMetaphoneKeys("Katherine") = ("K0RN", "KTRN")
#   - DoubleMetaphoneKeys("Pacheco") contains "PXK" (Spanish Romance gate)
#
# Language-branch checklist (CONTEXT.md §3 LOCKED — all 5 branches must pass):
#   Germanic / Slavic / Romance / Greek / Chinese-origin
#
# Non-ASCII input: silently dropped per CONTEXT.md §5.

Feature: Double Metaphone phonetic similarity (Philips 2000)
  Double Metaphone encodes names into two phonetic keys (primary and secondary),
  each at most 4 characters from the charset [A-Z0] where 0 represents the theta
  sound. DoubleMetaphoneScore returns 1.0 if any of the four key combinations match
  (primary-primary, primary-secondary, secondary-primary, secondary-secondary) and
  the matched key is non-empty; otherwise 0.0.

  Scenario: identical strings score 1.0
    # The identity short-circuit fires before any DoubleMetaphoneKeys computation.
    When I compute the Double Metaphone score between "Schmidt" and "Schmidt"
    Then the score should be exactly 1

  Scenario: both-empty strings score 1.0
    # Both-empty convention: the identity short-circuit returns 1.0.
    When I compute the Double Metaphone score between "" and ""
    Then the score should be exactly 1

  Scenario: Germanic branch — Schmidt/Smith XMT cross-match (LOAD-BEARING)
    # MANDATORY GATE: Schmidt primary = "XMT"; Smith secondary = "XMT".
    # The shared XMT key yields score 1.0 via the primary-a == secondary-b branch.
    When I compute the Double Metaphone keys of "Schmidt"
    Then the keys should be "XMT" and "SMT"
    When I compute the Double Metaphone keys of "Smith"
    Then the keys should be "SM0" and "XMT"
    When I compute the Double Metaphone score between "Schmidt" and "Smith"
    Then the score should be exactly 1

  Scenario: Greek branch — Catherine equals Katherine (LOAD-BEARING)
    # MANDATORY GATE: both names encode identically to ("K0RN", "KTRN").
    # TH → theta "0" primary, T secondary — Greek origin pattern.
    When I compute the Double Metaphone keys of "Catherine"
    Then the keys should be "K0RN" and "KTRN"
    When I compute the Double Metaphone keys of "Katherine"
    Then the keys should be "K0RN" and "KTRN"
    When I compute the Double Metaphone score between "Catherine" and "Katherine"
    Then the score should be exactly 1

  Scenario: Romance/Spanish branch — Pacheco PXK gate
    # Spanish CH after vowel produces X/K sounds.
    When I compute the Double Metaphone keys of "Pacheco"
    Then the primary key should contain "PXK"

  Scenario: Slavic branch — Sczepanski non-empty keys
    # Slavic SZC compound produces phonetic key.
    When I compute the Double Metaphone keys of "Sczepanski"
    Then both keys should be non-empty

  Scenario: Chinese-origin branch — Cheung non-empty keys
    # Initial CHE- pattern (Chinese-origin) triggers X/K phoneme.
    When I compute the Double Metaphone keys of "Cheung"
    Then both keys should be non-empty

  Scenario: non-matching names score 0.0
    When I compute the Double Metaphone score between "Smith" and "Garcia"
    Then the score should be exactly 0

  Scenario: one-empty string scores 0.0
    When I compute the Double Metaphone score between "Schmidt" and ""
    Then the score should be exactly 0
