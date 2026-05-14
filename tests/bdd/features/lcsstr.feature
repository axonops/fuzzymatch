# Primary source: Wagner, R. A., & Fischer, M. J. (1974). "The string-to-string
# correction problem." Journal of the ACM, 21(1):168-173. The canonical
# dynamic-programming formulation for the longest-common-substring family.
#
# Score normalisation (SPEC-PINNED at docs/requirements.md §7.1.9):
#   score = 2 · len(lcs) / (len(a) + len(b))   (Sørensen-Dice form)
#
# Surface: LCSStrScore is the dispatched byte-path score (this feature
# exercises that surface end-to-end). The substring-returning surface
# (LongestCommonSubstring / *Runes) is covered by unit tests, not BDD —
# BDD scenarios assert score numerics only.

Feature: LCSStr similarity (longest common substring)
  LCSStr (Longest Common Substring) similarity is the Sørensen-Dice-normalised
  length of the longest contiguous segment shared by two strings. The
  substring-returning surface (LongestCommonSubstring) is non-standard but
  consumer-justified for the schema-similarity use case ("which substring
  is driving the match?"). When multiple longest matches of equal length
  exist, the leftmost in a is returned.

  Scenario Outline: canonical reference vectors
    When I compute the LCSStr score between "<a>" and "<b>"
    Then the score should be approximately <score> within 0.0001

    Examples:
      | a            | b                          | score  |
      | kitten       | sitting                    | 0.4615 |
      | http_request | http_request_header_fields | 0.6316 |
      | abcdef       | zabcdefuvw                 | 0.7500 |
      | banana       | ananas                     | 0.8333 |

  Scenario: identical strings score 1.0
    When I compute the LCSStr score between "user_id" and "user_id"
    Then the score should be exactly 1

  Scenario: both-empty strings score 1.0
    # The both-empty convention: LCSStrScore returns 1.0 vacuously; the
    # 2·0/(0+0) division is short-circuited by an explicit both-empty guard.
    When I compute the LCSStr score between "" and ""
    Then the score should be exactly 1

  Scenario: one-empty string scores 0.0
    When I compute the LCSStr score between "abc" and ""
    Then the score should be exactly 0

  Scenario: no-overlap disambiguation pin (RESEARCH.md Pitfall 6)
    # LongestCommonSubstring returns "" both for both-empty AND for no-overlap.
    # LCSStrScore disambiguates: 1.0 for both-empty, 0.0 for no-overlap.
    # This scenario pins the no-overlap arm of that contract.
    When I compute the LCSStr score between "abc" and "xyz"
    Then the score should be exactly 0

  Scenario: score is symmetric
    # The substring relation is symmetric and 2·len(lcs)/(la+lb) is symmetric
    # in (a, b) — verified on the canonical Wagner-Fischer 1974 pair.
    When I compute the LCSStr score between "kitten" and "sitting"
    And I compute the second LCSStr score between "sitting" and "kitten"
    Then both LCSStr scores should be equal

  Scenario: leftmost tie-break preserves score symmetry
    # The leftmost-in-a tie-break (strict-`>` max-update — RESEARCH.md
    # Pitfall 4) affects WHICH substring is returned, but the SCORE itself
    # is symmetric in argument order: both directions return 2·3/(9+3) = 0.5.
    When I compute the LCSStr score between "abcXYZabc" and "abc"
    Then the score should be exactly 0.5
    When I compute the LCSStr score between "abc" and "abcXYZabc"
    Then the score should be exactly 0.5

  Scenario: substring containment scores < 1.0 (vs SWG which would saturate)
    # The substring containment case where 'a' is fully present in 'b' — the
    # LCS is 'a' itself, but the score normalisation by (la+lb) keeps the
    # result well below 1.0 (in contrast to SWG which clamps to 1.0 here).
    When I compute the LCSStr score between "http_request" and "http_request_header_fields"
    Then the score should be approximately 0.6316 within 0.0001
