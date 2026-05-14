# Primary source: Hamming, R. W. (1950). "Error detecting and error correcting
# codes." Bell System Technical Journal, 29(2):147-160.
#
# Score normalisation: 1 - distance / max(len(a), len(b)).
# Unequal-length policy (LOCKED): returns 0.0 silently — no error, no panic.
# Reference vectors from Hamming 1950.

Feature: Hamming similarity algorithm
  The Hamming distance counts positions where two equal-length strings differ.
  For unequal-length inputs, the project-wide silent-zero policy applies:
  HammingDistance returns max(len(a), len(b)) and HammingScore returns 0.0.

  Scenario Outline: canonical reference vectors
    When I compute the Hamming score between "<a>" and "<b>"
    Then the score should be approximately <score> within 0.0001

    Examples:
      | a       | b       | score  |
      | karolin | kathrin | 0.5714 |
      | abc     | abc     | 1.0000 |

  Scenario: both-empty strings score 1.0
    When I compute the Hamming score between "" and ""
    Then the score should be exactly 1.0

  Scenario: unequal length returns silent zero
    When I compute the Hamming score between "abc" and "ab"
    Then the score should be exactly 0.0

  Scenario: unequal length distance equals max length
    When I compute the Hamming distance between "abc" and "ab"
    Then the distance should be 3

  Scenario: score is symmetric for equal-length inputs
    When I compute the Hamming score between "karolin" and "kathrin"
    And I compute the second Hamming score between "kathrin" and "karolin"
    Then both Hamming scores should be equal

  Scenario: score is symmetric for unequal-length inputs
    When I compute the Hamming score between "abc" and "ab"
    And I compute the second Hamming score between "ab" and "abc"
    Then both Hamming scores should be equal
