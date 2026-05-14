# Primary source: Smith, T. F. & Waterman, M. S. (1981). "Identification of
# common molecular subsequences." J. Mol. Biol. 147:195-197.
# Affine-gap reduction: Gotoh, O. (1982). "An improved algorithm for matching
# biological sequences." J. Mol. Biol. 162:705-708.
# Corrected initialisation: Flouri, T. et al. (2015). "Are all global alignment
# algorithms and implementations correct?" biorxiv 031500.
#
# Score normalisation: clamp(best_local_score / min(len(a), len(b)), 0, 1).
# Default params: Match=1.0, Mismatch=-1.0, GapOpen=-1.5, GapExtend=-0.5.

Feature: Smith-Waterman-Gotoh local-alignment similarity
  Local alignment with affine gap penalty. The shorter input is sought as a
  subsequence inside the longer input; the score reflects the best matching
  region. Default parameters (Match=1.0, Mismatch=-1.0, GapOpen=-1.5,
  GapExtend=-0.5) match the documented defaults from NewSWGParams.

  Scenario Outline: canonical reference vectors
    When I compute the SmithWatermanGotoh score between "<a>" and "<b>"
    Then the score should be approximately <score> within 0.0001

    Examples:
      | a            | b                            | score  |
      | http_request | http_request_header_fields   | 1.0000 |
      | abc          | abc                          | 1.0000 |

  Scenario: identical strings score 1.0
    When I compute the SmithWatermanGotoh score between "user_id" and "user_id"
    Then the score should be exactly 1

  Scenario: both-empty strings score 1.0
    When I compute the SmithWatermanGotoh score between "" and ""
    Then the score should be exactly 1

  Scenario: one-empty string scores 0.0
    When I compute the SmithWatermanGotoh score between "abc" and ""
    Then the score should be exactly 0

  Scenario: score is symmetric
    When I compute the SmithWatermanGotoh score between "kitten" and "sitting"
    And I compute the second SmithWatermanGotoh score between "sitting" and "kitten"
    Then both SmithWatermanGotoh scores should be equal

  Scenario: gap-split canary — symmetric long-gap pair scores equally
    When I compute the SmithWatermanGotoh score between "abc________def" and "abcdef"
    And I compute the second SmithWatermanGotoh score between "abc____def____" and "abcdef"
    Then both SmithWatermanGotoh scores should be equal
