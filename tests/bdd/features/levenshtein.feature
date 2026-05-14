# Primary source: Levenshtein, V. I. (1965). "Binary codes capable of correcting
# deletions, insertions, and reversals." Soviet Physics Doklady, 10(8):707-710.
#
# Two-row DP formulation: Wagner, R. A., & Fischer, M. J. (1974). "The
# string-to-string correction problem." Journal of the ACM (JACM), 21(1):168-173.
#
# Score normalisation: 1 - distance / max(len(a), len(b)).
# Reference vectors from Wagner-Fischer 1974.

Feature: Levenshtein similarity algorithm
  The Levenshtein distance measures the minimum number of single-character edits
  (insertions, deletions, or substitutions) required to change one string into
  another. The similarity score normalises this to [0.0, 1.0].

  Scenario Outline: canonical reference vectors
    When I compute the Levenshtein score between "<a>" and "<b>"
    Then the score should be approximately <score> within 0.0001

    Examples:
      | a        | b        | score  |
      | kitten   | sitting  | 0.5714 |
      | saturday | sunday   | 0.6250 |
      | abc      | abc      | 1.0000 |

  Scenario: identical strings score 1.0
    When I compute the Levenshtein score between "user_id" and "user_id"
    Then the score should be exactly 1.0

  Scenario: both-empty strings score 1.0
    When I compute the Levenshtein score between "" and ""
    Then the score should be exactly 1.0

  Scenario: one-empty string scores 0.0
    When I compute the Levenshtein score between "abc" and ""
    Then the score should be exactly 0.0

  Scenario: score is symmetric
    When I compute the Levenshtein score between "kitten" and "sitting"
    And I compute the second Levenshtein score between "sitting" and "kitten"
    Then both Levenshtein scores should be equal
