# Primary source: Salton, G., McGill, M. J. (1983). "Introduction to Modern
# Information Retrieval" McGraw-Hill, §4.1 eq. 4.4, p.121 — vector-space
# cosine measure over q-gram frequency vectors.
#
# Cross-validation: hand-derived RV-C1..RV-C5 reference vectors per
# CONTEXT.md §4 LOCKED — no Python toolchain. Each derivation is
# reviewer-verifiable in under 30 seconds against Salton & McGill 1983
# §4.1 (RESEARCH.md §2.3). High-precision scenarios pin the actual
# IEEE-754 factorised-form output (1 ULP from the rational limit per
# RESEARCH.md "Pitfall 2") at 17-significant-digit precision.
#
# LOAD-BEARING per CONTEXT.md §1: Cosine is the cross-platform
# float-determinism gate for Phase 5. testdata/golden/_staging/
# cosine.json (merged in plan 05-05) is the byte-identical CI matrix
# detector.
#
# Surface: CosineScore is the dispatched byte-path score (this feature
# exercises that surface end-to-end). The rune-path variant
# CosineScoreRunes is exercised by a dedicated rune scenario at the
# bottom of this file (RV-C3).

Feature: Cosine n-gram similarity (vector-space cosine over q-gram frequency vectors)
  Cosine is the textbook information-retrieval similarity measure:
  cos(A, B) = (A · B) / (‖A‖ × ‖B‖) over the multiset of overlapping
  length-n substrings of A and B (Salton & McGill 1983 §4.1 eq. 4.4
  p.121). The byte path operates on bytes (multi-byte UTF-8 splits at
  byte boundaries); the rune path operates on Unicode code points.
  Both surfaces are symmetric and return scores in [0.0, 1.0].

  Scenario Outline: Hand-derived reference vectors with float64 precision
    # RV-C1 / RV-C2 / RV-C5 pin the actual IEEE-754 factorised-form
    # output at 17-digit precision (1 ULP from the rational limit; see
    # cosine_test.go for the full derivation). RV-C4 is approximately
    # 0.5 (1 ULP shortfall — sqrt(2)*sqrt(2) = 2.0000000000000004).
    # Identity / orthogonal use the loose 0.0001 tolerance.
    When I compute the Cosine score between "<a>" and "<b>" with n <n>
    Then the score should be approximately <score> within <tol>

    Examples:
      | a        | b        | n | score              | tol    |
      | abc      | abcd     | 2 | 0.8164965809277259 | 0.0001 |
      | abcdefgh | abcdefgi | 3 | 0.8333333333333335 | 0.0001 |
      | abcde    | abcdf    | 4 | 0.4999999999999999 | 0.0001 |
      | ab       | abcd     | 2 | 0.5773502691896258 | 0.0001 |
      | hello    | hello    | 2 | 1.0000             | 0.0001 |
      | abc      | xyz      | 2 | 0.0000             | 0.0001 |

  Scenario: identical strings score 1.0
    When I compute the Cosine score between "user_id" and "user_id" with n 2
    Then the score should be exactly 1

  Scenario: both-empty strings score 1.0
    # The both-empty convention is the vacuous-match identity: with no
    # q-grams to disagree about, the multisets are trivially equal. The
    # a == b identity short-circuit handles this case.
    When I compute the Cosine score between "" and "" with n 2
    Then the score should be exactly 1

  Scenario Outline: one-empty string scores 0.0
    When I compute the Cosine score between "<a>" and "<b>" with n 2
    Then the score should be exactly 0

    Examples:
      | a   | b   |
      | abc |     |
      |     | abc |

  Scenario: orthogonal strings score 0.0
    # Empty intersection → dot product = 0 → cos = 0 (regardless of the
    # individual norms). Exercises the empty-intersection branch in
    # cosineFromQGramMaps.
    When I compute the Cosine score between "abc" and "xyz" with n 2
    Then the score should be exactly 0

  Scenario: score is symmetric
    # Cosine is symmetric — cos(A, B) = cos(B, A) bit-for-bit, NOT
    # within tolerance. Sorted-key iteration is canonical regardless of
    # input argument order; the dot-product reduction visits the same
    # intersection keys in the same sorted order, producing
    # bit-identical float64 output.
    When I compute the Cosine score between "abcdefgh" and "abcdefgi" with n 3
    And I compute the second Cosine score between "abcdefgi" and "abcdefgh" with n 3
    Then both Cosine scores should be equal

  Scenario: rune-path Unicode pair (RV-C3 café/cafe at n=2)
    # The rune path treats "café" as 4 runes (decomposing as
    # ['c','a','f','é']), producing rune-bigrams {"ca":1, "af":1, "fé":1}.
    # Against "cafe" ({"ca":1, "af":1, "fe":1}), the intersection is
    # ["af", "ca"] (sorted byte-lex; "fé" sorts AFTER ASCII pairs,
    # "fe" not in intersection) and the per-side norms are both
    # sqrt(3), so cos = 2/(sqrt(3)·sqrt(3)) ≈ 0.6666666666666667 (the
    # IEEE-754 actual is 1 ULP above the 2/3 rational limit; see
    # RESEARCH.md "Pitfall 2"). The byte path would split "é"
    # mid-codepoint and yield a different score.
    When I compute the CosineRunes score between "café" and "cafe" with n 2
    Then the score should be approximately 0.6666666666666667 within 0.0001
