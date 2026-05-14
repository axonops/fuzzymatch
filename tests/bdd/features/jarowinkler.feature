# Primary source: Winkler, W. E. (1990). "String comparator metrics and
# enhanced decision rules in the Fellegi-Sunter model of record linkage."
# Proceedings of the Section on Survey Research Methods, American Statistical
# Association: 354-359. (See especially p. 357 for constants and reference pairs.)
#
# Jaro-Winkler formula: JW = J + L * p * (1 - J)
#   where J = JaroScore(a, b), L = common-prefix length (capped at 4),
#   p = 0.1 (prefix scale factor), applied only when J >= 0.7 (boost gate).
#
# Constants (LOCKED by REQUIREMENTS.md CHAR-06 and CONTEXT.md):
#   winklerPrefixScale    = 0.1   (Winkler 1990 p. 357)
#   winklerMaxPrefix      = 4     (Winkler 1990 p. 357)
#   winklerBoostThreshold = 0.7   (Winkler 1990 p. 357)
#
# Note on non-metric property: Jaro-Winkler is NOT a metric; the triangle
# inequality does not hold. BDD scenarios for "JW is not a metric" are not
# included because the claim is verified by the omission of
# TestProp_JaroWinklerScore_TriangleInequality in props_test.go (with
# documented rationale) and by the explicit godoc paragraph in jarowinkler.go.

Feature: Jaro-Winkler similarity algorithm
  The Jaro-Winkler similarity extends the Jaro similarity with a common-prefix
  bonus that rewards strings sharing a common prefix. The bonus is proportional
  to the prefix length (capped at 4) and only applied when the underlying Jaro
  score is at least 0.7 (the boost threshold). Higher scores indicate greater
  similarity, with 1.0 meaning identical and 0.0 meaning maximally dissimilar.

  The both-empty convention: JaroWinklerScore("", "") == 1.0 exactly (identity).
  The one-empty convention: JaroWinklerScore("", "any") == 0.0 exactly.

  Scenario Outline: canonical reference vectors (Winkler 1990 p. 357)
    When I compute the JaroWinkler score between "<a>" and "<b>"
    Then the score should be approximately <score> within <tol>

    Examples:
      | a        | b        | score  | tol    |
      | MARTHA   | MARHTA   | 0.9611 | 0.0001 |
      | DIXON    | DICKSONX | 0.8133 | 0.0001 |
      | DWAYNE   | DUANE    | 0.8400 | 0.0010 |
      | ABC      | ABC      | 1.0000 | 0.0000 |

  Scenario: both-empty strings score 1.0
    When I compute the JaroWinkler score between "" and ""
    Then the score should be exactly 1.0

  Scenario: one-empty strings score 0.0
    When I compute the JaroWinkler score between "" and "ABC"
    Then the score should be exactly 0.0

  Scenario: boost gated by threshold — below-threshold pair returns Jaro unchanged
    # "abc" vs "xyz" has no character matches → Jaro = 0.0 (well below 0.7).
    # JaroWinklerScore must equal JaroScore for the same pair (no boost applied).
    When I compute the JaroWinkler score between "abc" and "xyz"
    And I compute the Jaro score between "abc" and "xyz"
    Then both JaroWinkler and Jaro scores should be equal

  Scenario: prefix length capped at 4 characters
    # "TESTABCD" vs "TESTABCE" share a 7-character common prefix "TESTABC",
    # but winklerMaxPrefix = 4 means L is capped at 4.
    # Expected JW = J + 4 * 0.1 * (1 - J) where J = JaroScore("TESTABCD","TESTABCE").
    # The score must be within tolerance of the L=4 expectation, not the L=7 expectation.
    When I compute the JaroWinkler score between "TESTABCD" and "TESTABCE"
    Then the score should be approximately 0.9500 within 0.0001

  Scenario: JaroWinkler is symmetric
    When I compute the JaroWinkler score between "MARTHA" and "MARHTA"
    And I compute the second JaroWinkler score between "MARHTA" and "MARTHA"
    Then both JaroWinkler scores should be equal
