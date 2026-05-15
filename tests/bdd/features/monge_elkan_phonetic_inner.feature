# Primary source: Monge & Elkan (1996), KDD'96 §3 — binary-inner composition.
# Cross-reference: CONTEXT.md §4 LOCKED — permittedMongeElkanInner at FINAL
# 18 entries including the four phonetic AlgoIDs (AlgoSoundex,
# AlgoDoubleMetaphone, AlgoNYSIIS, AlgoMRA) added by plans 07-01..07-04.
#
# These scenarios cover the binary-inner-composition behaviour: phonetic
# algorithms return exactly 0.0 or 1.0 per token comparison. When ME's
# per-token-max-mean accumulates one 1.0 match and one 0.0 non-match (over
# two A-tokens), the result is exactly 0.5.
#
# Fixture rationale:
#   "alpha" and "gamma" are phonetically unrelated — all four phonetic
#   algorithms return 0.0 for this pair. "alpha" and "alpha" are
#   identical — the identity short-circuit fires, returning 1.0, before
#   any phonetic encoding. This fixture is independent of the specific
#   phonetic encoding rules; it exercises the ME accumulation logic.
#
# TestMongeElkanScore_BinaryInner_{Soundex,DoubleMetaphone,NYSIIS,MRA} in
# monge_elkan_test.go are the unit-test analogs. These BDD scenarios are the
# BDD-layer regression detectors.

@token @monge-elkan @phonetic
Feature: Monge-Elkan composition with phonetic binary inner algorithms
  When a phonetic algorithm (Soundex, Double Metaphone, NYSIIS, MRA) is used
  as the inner metric for Monge-Elkan, each per-token comparison returns
  exactly 0.0 or 1.0. The per-token-max-mean then produces exact fractional
  scores: 1.0 when all token pairs match, 0.0 when none match, and 0.5 when
  exactly one of two A-tokens finds a matching B-token.

  @token @monge-elkan @phonetic @soundex
  Scenario Outline: Monge-Elkan over Soundex — binary inner composition
    # "alpha beta" has two tokens; "alpha" matches "alpha" (identity → 1.0)
    # but not "gamma" (phonetically unrelated → 0.0); ME averages 1.0 and 0.0
    # to produce exactly 0.5.
    When I compute the MongeElkan score between "<a>" and "<b>" with inner AlgoSoundex
    Then the score should be approximately <score> within 0.0001

    Examples:
      | a           | b           | score |
      | alpha beta  | alpha gamma | 0.5   |
      | alpha beta  | alpha beta  | 1.0   |
      | alpha       | gamma       | 0.0   |

  @token @monge-elkan @phonetic @double-metaphone
  Scenario Outline: Monge-Elkan over Double Metaphone — binary inner composition
    # Double Metaphone uses two keys (primary + secondary); a match on either
    # key counts as 1.0. "alpha" matches "alpha" (identity); "gamma" does not
    # match "alpha" under Double Metaphone.
    When I compute the MongeElkan score between "<a>" and "<b>" with inner AlgoDoubleMetaphone
    Then the score should be approximately <score> within 0.0001

    Examples:
      | a           | b           | score |
      | alpha beta  | alpha gamma | 0.5   |
      | alpha beta  | alpha beta  | 1.0   |
      | alpha       | gamma       | 0.0   |

  @token @monge-elkan @phonetic @nysiis
  Scenario Outline: Monge-Elkan over NYSIIS — binary inner composition
    # NYSIIS encodes ASCII names to a 6-char key (Taft 1970 / Knuth TAOCP §6.4
    # truncation). "alpha" and "gamma" produce different codes; "alpha" and
    # "alpha" are identical (identity short-circuit → 1.0).
    When I compute the MongeElkan score between "<a>" and "<b>" with inner AlgoNYSIIS
    Then the score should be approximately <score> within 0.0001

    Examples:
      | a           | b           | score |
      | alpha beta  | alpha gamma | 0.5   |
      | alpha beta  | alpha beta  | 1.0   |
      | alpha       | gamma       | 0.0   |

  @token @monge-elkan @phonetic @mra
  Scenario Outline: Monge-Elkan over MRA — binary inner composition
    # MRA (NBS Tech Note 943) applies a length-difference gate before encoding:
    # if |len(code_a) - len(code_b)| > 3, the comparison returns 0.0 directly.
    # "alpha" vs "gamma" produces different MRA codes; "alpha" vs "alpha" is
    # the identity short-circuit (1.0).
    When I compute the MongeElkan score between "<a>" and "<b>" with inner AlgoMRA
    Then the score should be approximately <score> within 0.0001

    Examples:
      | a           | b           | score |
      | alpha beta  | alpha gamma | 0.5   |
      | alpha beta  | alpha beta  | 1.0   |
      | alpha       | gamma       | 0.0   |

  @token @monge-elkan @phonetic @soundex
  Scenario: Monge-Elkan over Soundex — Robert and Rupert share code R163
    # Robert → R163, Rupert → R163: Soundex match → score 1.0.
    # Demonstrates that the binary inner captures real phonetic similarity.
    When I compute the MongeElkan score between "Robert" and "Rupert" with inner AlgoSoundex
    Then the score should be approximately 1.0 within 0.0001

  @token @monge-elkan @phonetic @double-metaphone
  Scenario: Monge-Elkan over Double Metaphone — Schmidt and Smith share XMT key
    # Schmidt → ("XMT", "SMT"); Smith → ("SM0", "XMT"): the shared XMT key
    # matches → score 1.0. Demonstrates the Germanic language-branch fixture
    # from CONTEXT.md §3.
    When I compute the MongeElkan score between "Schmidt" and "Smith" with inner AlgoDoubleMetaphone
    Then the score should be approximately 1.0 within 0.0001
