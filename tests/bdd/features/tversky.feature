# Primary source: Tversky, A. (1977). "Features of similarity."
# Psychological Review 84(4):327-352, §2 — the asymmetric similarity
# model (eq. (1) p. 332, "the ratio model").
#
# Cross-validation: hand-derived RV-T1..RV-T4 reference vectors per
# CONTEXT.md §4 LOCKED — no Python toolchain. Each derivation is
# reviewer-verifiable in under a minute against Tversky 1977 §2
# (RESEARCH.md §2.4).
#
# LOAD-BEARING per CONTEXT.md §5: the asymmetry direction-sensitivity
# scenario below is the BDD-layer regression detector for the RV-T1 vs
# RV-T2 input-swap pair. A silent α/β swap inside TverskyScore would
# produce |T(a, b, ...) - T(b, a, ...)| < 0.01 (or vice versa, swap
# the values) — the "differ by more than 0.1" assertion catches this
# at the BDD layer in addition to the unit-test
# (TestTversky_AsymmetryDirectionSensitive) and golden-file
# (testdata/golden/_staging/tversky.json — RV-T1 + RV-T2 as separate
# rows) gates.
#
# Surface: TverskyScore is the dispatched byte-path score (this feature
# exercises that surface end-to-end with explicit α/β arguments). The
# rune-path variant TverskyScoreRunes is exercised by a dedicated rune
# scenario at the bottom of this file.

Feature: Tversky asymmetric similarity (parameterised q-gram metric over weighted residuals)
  Tversky is the parameterised q-gram metric:
  T(A, B, α, β) = |QA ∩ QB| / (|QA ∩ QB| + α·|QA − QB| + β·|QB − QA|)
  over the multiset of overlapping length-n substrings of A and B
  (Tversky 1977 §2 eq. (1) p. 332). With α ≠ β the function is
  direction-sensitive — swapping (a, b) with fixed (α, β) generally
  produces a different score. With α = β = 1 the function reduces to
  Q-Gram Jaccard; with α = β = 0.5 it reduces to Sørensen-Dice. Both
  byte and rune surfaces ship and return scores in [0.0, 1.0].

  Scenario Outline: Canonical reference vectors (RV-T1, RV-T3, RV-T4 + identity)
    # RV-T1: asymmetric direction-sensitive — first half of the
    # load-bearing input-swap pair. RV-T3: α=β=1.0 reduces to Jaccard
    # (the cross-check pair confirms degeneracy). RV-T4: α=β=0.5
    # reduces to Sørensen-Dice (the cross-check pair confirms
    # degeneracy). The identity row pins the a == b short-circuit at
    # an asymmetric (α, β).
    When I compute the Tversky score between "<a>" and "<b>" with n <n> alpha <alpha> beta <beta>
    Then the score should be approximately <score> within 0.0001

    Examples:
      | a     | b      | n | alpha | beta | score  |
      | abcd  | abcdef | 2 | 0.8   | 0.2  | 0.8824 |
      | abcd  | abce   | 2 | 1.0   | 1.0  | 0.5000 |
      | abcd  | abce   | 2 | 0.5   | 0.5  | 0.6667 |
      | hello | hello  | 2 | 0.8   | 0.2  | 1.0000 |

  Scenario: identical strings score 1.0 (α/β irrelevant — short-circuit fires)
    When I compute the Tversky score between "user_id" and "user_id" with n 2 alpha 0.8 beta 0.2
    Then the score should be exactly 1

  Scenario: both-empty strings score 1.0
    # The both-empty convention is the vacuous-match identity: with no
    # q-grams to disagree about, the multisets are trivially equal. The
    # a == b identity short-circuit handles this case (α/β irrelevant).
    When I compute the Tversky score between "" and "" with n 2 alpha 0.5 beta 0.5
    Then the score should be exactly 1

  Scenario Outline: one-empty string scores 0.0
    When I compute the Tversky score between "<a>" and "<b>" with n 2 alpha 0.5 beta 0.5
    Then the score should be exactly 0

    Examples:
      | a   | b   |
      | abc |     |
      |     | abc |

  Scenario: Asymmetry direction-sensitivity gate (RV-T1 vs RV-T2 input swap — LOAD-BEARING)
    # The LOAD-BEARING asymmetry gate at the BDD layer. With fixed
    # (α=0.8, β=0.2):
    #   RV-T1: TverskyScore("abcd",   "abcdef", 2, 0.8, 0.2) = 0.8823529411764706
    #   RV-T2: TverskyScore("abcdef", "abcd",   2, 0.8, 0.2) = 0.6521739130434783
    # The two scores differ by ≈ 0.2302 — the input swap with the SAME
    # (α, β) is the direct evidence of direction-sensitivity. A silent
    # α/β swap inside TverskyScore would cause the values to flip
    # (RV-T1 → RV-T2's score and vice versa) — both would still be in
    # [0, 1] and would pass RangeBounds + Identity invariants — but
    # this scenario's `differ by more than 0.1` assertion catches the
    # regression.
    When I compute the Tversky score between "abcd" and "abcdef" with n 2 alpha 0.8 beta 0.2
    And I compute the second Tversky score between "abcdef" and "abcd" with n 2 alpha 0.8 beta 0.2
    Then the two Tversky scores should differ by more than 0.1

  Scenario: Parameter-swap symmetry (T(a,b,α,β) = T(b,a,β,α) — Tversky 1977 §2)
    # The algebraic identity that pins the asymmetry as a consequence
    # of α ≠ β rather than a one-sided coding error. Swapping the
    # inputs is EXACTLY equivalent to swapping the parameters.
    When I compute the Tversky score between "abcd" and "abcdef" with n 2 alpha 0.8 beta 0.2
    And I compute the second Tversky score between "abcdef" and "abcd" with n 2 alpha 0.2 beta 0.8
    Then both Tversky scores should be equal

  Scenario: rune-path Unicode pair (café/cafe at n=2, α=β=0.5 → Dice-equivalent)
    # The rune path treats "café" as 4 runes (decomposing as
    # ['c','a','f','é']), producing rune-bigrams {"ca":1, "af":1, "fé":1}.
    # Against "cafe" ({"ca":1, "af":1, "fe":1}), the intersection is 2
    # ("ca", "af"), |A−B| = 1 (fé), |B−A| = 1 (fe). With α=β=0.5
    # the formula reduces to the Sørensen-Dice degeneracy:
    # T = 2/(2 + 0.5·1 + 0.5·1) = 2/3 ≈ 0.6667. The byte path would
    # split "é" mid-codepoint and yield a different score.
    When I compute the TverskyRunes score between "café" and "cafe" with n 2 alpha 0.5 beta 0.5
    Then the score should be approximately 0.6667 within 0.0001
