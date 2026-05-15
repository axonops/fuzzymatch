# Primary sources: Dice, L. R. (1945). "Measures of the amount of ecologic
# association between species." Ecology 26(3):297-302, §3 — original
# formulation. Sørensen, T. (1948). "A method of establishing groups of
# equal amplitude in plant sociology based on similarity of species
# content." Kongelige Danske Videnskabernes Selskab 5(4):1-34, §3 —
# independent rediscovery of the same coefficient.
#
# Cross-validation: hand-derived RV-D1..RV-D4 reference vectors per
# CONTEXT.md §4 LOCKED — no Python toolchain. Each derivation is
# reviewer-verifiable in under a minute against Dice 1945 §3 (RESEARCH.md
# §2.2).
#
# Surface: SorensenDiceScore is the dispatched byte-path score (this feature
# exercises that surface end-to-end). The rune-path variant
# SorensenDiceScoreRunes is exercised by a dedicated rune scenario at the
# bottom of this file.

Feature: Sørensen-Dice similarity (multiset Dice coefficient over q-grams)
  Sørensen-Dice is the textbook q-gram metric that weights the
  intersection more heavily than Jaccard:
  DSC(A, B) = 2·|QA ∩ QB| / (|QA| + |QB|) over the multiset of overlapping
  length-n substrings of A and B. The byte path operates on bytes
  (multi-byte UTF-8 splits at byte boundaries); the rune path operates on
  Unicode code points. Both surfaces are symmetric and return scores in
  [0.0, 1.0].

  Scenario Outline: Canonical reference vectors
    When I compute the SorensenDice score between "<a>" and "<b>" with n <n>
    Then the score should be approximately <score> within 0.0001

    Examples:
      | a      | b      | n | score  |
      | night  | nacht  | 2 | 0.2500 |
      | abcdef | bcdefg | 2 | 0.8000 |
      | abcdef | abcXef | 3 | 0.2500 |
      | hello  | hello  | 2 | 1.0000 |

  Scenario: identical strings score 1.0
    When I compute the SorensenDice score between "user_id" and "user_id" with n 2
    Then the score should be exactly 1

  Scenario: both-empty strings score 1.0
    # The both-empty convention is the vacuous-match identity: with no
    # q-grams to disagree about, the multisets are trivially equal. The
    # a == b identity short-circuit handles this case.
    When I compute the SorensenDice score between "" and "" with n 2
    Then the score should be exactly 1

  Scenario Outline: one-empty string scores 0.0
    When I compute the SorensenDice score between "<a>" and "<b>" with n 2
    Then the score should be exactly 0

    Examples:
      | a   | b   |
      | abc |     |
      |     | abc |

  Scenario: score is symmetric
    # Sørensen-Dice is symmetric — DSC(A, B) = DSC(B, A) bit-for-bit, NOT
    # within tolerance. The integer-derived single
    # multiplication-then-division produces identical float64 output
    # regardless of argument order.
    When I compute the SorensenDice score between "night" and "nacht" with n 2
    And I compute the second SorensenDice score between "nacht" and "night" with n 2
    Then both SorensenDice scores should be equal

  Scenario: rune-path Unicode pair (café/cafe at n=2)
    # The rune path treats "café" as 4 runes (decomposing as
    # ['c','a','f','é']), producing rune-bigrams {"ca":1, "af":1, "fé":1}.
    # Against "cafe" ({"ca":1, "af":1, "fe":1}), the intersection is 2
    # ("ca", "af") and the per-side multiset totals are both 3, so
    # DSC = 2·2/(3+3) = 4/6 ≈ 0.6667. The byte path would split "é"
    # mid-codepoint and yield a different score.
    When I compute the SorensenDiceRunes score between "café" and "cafe" with n 2
    Then the score should be approximately 0.6667 within 0.0001
