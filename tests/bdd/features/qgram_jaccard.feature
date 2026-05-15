# Primary source: Ukkonen, E. (1992). "Approximate string-matching with q-grams
# and maximal matches." Theoretical Computer Science 92(1):191-211, §3 — the
# multiset Jaccard formulation over q-gram counts. Underlying set-coefficient
# origin: Jaccard, P. (1912). "The distribution of the flora in the alpine
# zone." New Phytologist 11(2):37-50, p. 43.
#
# Cross-validation: hand-derived RV-J1..RV-J6 reference vectors per CONTEXT.md
# §4 LOCKED — no Python toolchain. Each derivation is reviewer-verifiable in
# under a minute against Ukkonen 1992 §3 (RESEARCH.md §2.1).
#
# Surface: QGramJaccardScore is the dispatched byte-path score (this feature
# exercises that surface end-to-end). The rune-path variant
# QGramJaccardScoreRunes is exercised by a dedicated rune scenario at the
# bottom of this file.

Feature: Q-Gram Jaccard similarity (multiset Jaccard over q-grams)
  Q-Gram Jaccard is the textbook q-gram metric: J(A, B) = |QA ∩ QB| / |QA ∪ QB|
  over the multiset of overlapping length-n substrings of A and B. The byte
  path operates on bytes (multi-byte UTF-8 splits at byte boundaries); the
  rune path operates on Unicode code points. Both surfaces are symmetric and
  return scores in [0.0, 1.0].

  Scenario Outline: Canonical reference vectors
    When I compute the QGramJaccard score between "<a>" and "<b>" with n <n>
    Then the score should be approximately <score> within 0.0001

    Examples:
      | a    | b        | n | score  |
      | AGCT | AGCTAGCT | 2 | 0.4286 |
      | hello | hello   | 2 | 1.0000 |
      | abc  | xyz      | 2 | 0.0000 |
      | abcd | abxy     | 2 | 0.2000 |

  Scenario: identical strings score 1.0
    When I compute the QGramJaccard score between "user_id" and "user_id" with n 2
    Then the score should be exactly 1

  Scenario: both-empty strings score 1.0
    # The both-empty convention is the vacuous-match identity: with no
    # q-grams to disagree about, the multisets are trivially equal. The
    # a == b identity short-circuit handles this case.
    When I compute the QGramJaccard score between "" and "" with n 2
    Then the score should be exactly 1

  Scenario Outline: one-empty string scores 0.0
    When I compute the QGramJaccard score between "<a>" and "<b>" with n 2
    Then the score should be exactly 0

    Examples:
      | a   | b   |
      | abc |     |
      |     | abc |

  Scenario: score is symmetric
    # Set-Jaccard is symmetric — J(A, B) = J(B, A) bit-for-bit, NOT within
    # tolerance. The integer-derived single division produces identical
    # float64 output regardless of argument order.
    When I compute the QGramJaccard score between "AGCT" and "AGCTAGCT" with n 2
    And I compute the second QGramJaccard score between "AGCTAGCT" and "AGCT" with n 2
    Then both QGramJaccard scores should be equal

  Scenario: rune-path Unicode pair (RV-J5-Runes)
    # The rune path treats "café" as 4 runes (decomposing as ['c','a','f','é']),
    # producing rune-bigrams {"ca":1, "af":1, "fé":1}. Against "cafe"
    # ({"ca":1, "af":1, "fe":1}), the intersection is 2 ("ca", "af") and
    # the union is 4, so J = 2/4 = 0.5. The byte path would split "é"
    # mid-codepoint and yield a different score.
    When I compute the QGramJaccardRunes score between "café" and "cafe" with n 2
    Then the score should be approximately 0.5000 within 0.0001
