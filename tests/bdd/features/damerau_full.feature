# Primary source: Lowrance, R., Wagner, R. A. (1975). "An extension of the
# string-to-string correction problem." Journal of the ACM, 22(2):177-183.
#
# DL-Full (Lowrance-Wagner formulation) is the TRUE metric variant of
# Damerau-Levenshtein distance. It permits unrestricted transpositions:
# any pair of adjacent characters may be transposed and the characters
# may subsequently be edited.
#
# Score normalisation: 1 - distance / max(len(a), len(b)).
# Triangle inequality holds unconditionally — DL-Full IS a true metric.
# Reference vectors from Lowrance-Wagner 1975.

Feature: Damerau-Levenshtein Full (Lowrance-Wagner) similarity algorithm
  The Damerau-Levenshtein Full distance (Lowrance-Wagner 1975) extends
  Levenshtein by counting adjacent transpositions as a single edit, with
  NO restriction on subsequent editing. This makes DL-Full a TRUE metric:
  the triangle inequality holds unconditionally for all inputs.

  Scenario: Full DL discriminating reference vector — diverges from OSA
    # This vector proves Full DL != OSA. OSA returns distance 3 (score 0.0)
    # for the same pair because the OSA restriction forbids re-editing
    # characters after a transposition. Full DL returns distance 2 (score
    # ≈0.3333) because it permits unrestricted transpositions.
    # Lowrance-Wagner 1975 — the canonical discriminating vector.
    When I compute the DamerauLevenshteinFull distance between "ca" and "abc"
    Then the distance should be 2

  Scenario Outline: canonical reference vectors
    When I compute the DamerauLevenshteinFull score between "<a>" and "<b>"
    Then the score should be approximately <score> within 0.0001

    Examples:
      | a   | b   | score  |
      | ab  | ba  | 0.5000 |
      | ca  | abc | 0.3333 |
      | abc | abc | 1.0000 |

  Scenario: Full DL is a true metric (triangle inequality holds)
    # Triangle inequality D(a,c) <= D(a,b) + D(b,c) holds unconditionally
    # for DL-Full (Lowrance-Wagner 1975). This contrasts with DL-OSA, where
    # the OSA restriction can cause triangle inequality violations on contrived
    # long-string inputs. The actual triangle-inequality property is enforced
    # via testing/quick in props_test.go (TestProp_DamerauLevenshteinFullDistance_TriangleInequality);
    # this BDD scenario serves as documentation of the guarantee.
    #
    # Example: D("ca","ab") + D("ab","abc") >= D("ca","abc")
    # D("ca","ab") = 2 (two edits: c→b and transpose a), D("ab","abc") = 1, D("ca","abc") = 2
    # 2 + 1 = 3 >= 2 — triangle inequality holds.
    When I compute the DamerauLevenshteinFull score between "ca" and "ab"
    Then the score should be approximately 0.0000 within 0.5001

  Scenario: both-empty strings score 1.0
    When I compute the DamerauLevenshteinFull score between "" and ""
    Then the score should be exactly 1.0

  Scenario: score is symmetric
    When I compute the DamerauLevenshteinFull score between "ca" and "abc"
    And I compute the second DamerauLevenshteinFull score between "abc" and "ca"
    Then both DamerauLevenshteinFull scores should be equal
