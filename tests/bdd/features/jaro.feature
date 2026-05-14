# Primary source: Jaro, M. A. (1989). "Advances in record-linkage methodology
# as applied to matching the 1985 census of Tampa, Florida." Journal of the
# American Statistical Association, 84(406):414-420.
#
# Supporting reference for canonical reference vectors: Winkler, W. E. (1990).
# "String comparator metrics and enhanced decision rules in the Fellegi-Sunter
# model of record linkage." Proceedings of the Survey Research Methods Section
# of the American Statistical Association, pp. 354-359.
#
# Score formula: J = (m/|s1| + m/|s2| + (m-t)/m) / 3.0
#   where m = matched character count, t = transpositions/2,
#   matching window w = floor(max(|s1|,|s2|)/2) - 1.
#
# Note on non-metric property: Jaro is NOT a metric; the triangle inequality
# does not hold for Jaro similarity. BDD scenarios for "Jaro is not a metric"
# would require three-string comparisons that might not deterministically violate
# the inequality — the claim is verified instead by the omission of
# TestProp_JaroScore_TriangleInequality in props_test.go (with documented
# rationale) and by the explicit godoc paragraph in jaro.go.

Feature: Jaro similarity algorithm
  The Jaro similarity measures character overlap with a matching window and
  transposition penalty. Higher scores indicate greater similarity, with 1.0
  meaning identical and 0.0 meaning maximally dissimilar.

  The both-empty convention (locked per RESEARCH.md §Score Normalisation):
  JaroScore("", "") == 1.0 exactly (identity).

  The one-empty convention: JaroScore("", "any") == 0.0 exactly.

  Scenario Outline: canonical reference vectors
    When I compute the Jaro score between "<a>" and "<b>"
    Then the score should be approximately <score> within 0.0001

    Examples:
      | a          | b          | score  |
      | MARTHA     | MARHTA     | 0.9444 |
      | DIXON      | DICKSONX   | 0.7667 |
      | JELLYFISH  | SMELLYFISH | 0.8963 |
      | ABC        | ABC        | 1.0000 |

  Scenario: both-empty strings score 1.0
    When I compute the Jaro score between "" and ""
    Then the score should be exactly 1.0

  Scenario: one-empty strings score 0.0
    When I compute the Jaro score between "" and "ABC"
    Then the score should be exactly 0.0

  Scenario: Jaro is symmetric
    When I compute the Jaro score between "MARTHA" and "MARHTA"
    And I compute the second Jaro score between "MARHTA" and "MARTHA"
    Then both Jaro scores should be equal
