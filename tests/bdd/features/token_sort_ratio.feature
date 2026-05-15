# Engineering source: RapidFuzz documentation,
# https://rapidfuzz.github.io/RapidFuzz/Usage/fuzz.html#token-sort-ratio
# — the canonical modern reference for the Token Sort Ratio algorithm.
# Underlying DP source: Wagner, R. A., & Fischer, M. J. (1974). "The
# string-to-string correction problem." Journal of the ACM 21(1):168-173
# — the LCS-subsequence dynamic-programming recurrence consumed via the
# shared token_indel.go kernel.
#
# Cross-validation: RapidFuzz 3.14.5 via the corpus at
# testdata/cross-validation/token-ratios/vectors.json. The byte-stable
# agreement is asserted by token_ratio_cross_validation_test.go.
#
# Surface: TokenSortRatioScore is the dispatched byte-path score.
# Tokeniser-divergence note (OQ-1 RESOLUTION LOCKED in plan 06-01):
# fuzzymatch's Tokenise is identifier-aware (camelCase / snake_case /
# kebab-case / dot-case + lowercasing); RapidFuzz tokenises via Python
# str.split (whitespace-only, case-preserving). For whitespace-only
# lowercase ASCII inputs (the BDD scenarios below) the two agree.

@token
Feature: Token Sort Ratio (Indel-formula similarity over sorted-joined tokens)
  The Token Sort Ratio tokenises each side, sorts the token slices,
  joins with a single space, then applies the RapidFuzz Indel formula
  2·LCS / (|joinedA|+|joinedB|) over the resulting byte strings. The
  result lies in [0.0, 1.0]. Symmetric across argument order. Identical
  inputs (or inputs whose sorted tokens are identical) score 1.0.

  @token @token-sort-ratio
  Scenario Outline: Canonical reference vectors
    When I compute the TokenSortRatio score between "<a>" and "<b>"
    Then the score should be approximately <score> within 0.0001

    Examples:
      | a                      | b                      | score  |
      | fuzzy wuzzy was a bear | wuzzy fuzzy was a bear | 1.0000 |
      | alpha beta             | beta alpha             | 1.0000 |
      | alpha                  | alpha beta             | 0.6667 |
      | alpha beta             | alpha beta gamma       | 0.7692 |
      | hello                  | world                  | 0.2000 |
      | abc                    | xyz                    | 0.0000 |

  @token @token-sort-ratio
  Scenario: identical strings score 1.0
    When I compute the TokenSortRatio score between "hello world" and "hello world"
    Then the score should be exactly 1

  @token @token-sort-ratio
  Scenario: both-empty strings score 1.0
    # The both-empty convention is the vacuous-match identity: the
    # a == b identity short-circuit fires before Tokenise.
    When I compute the TokenSortRatio score between "" and ""
    Then the score should be exactly 1

  @token @token-sort-ratio
  Scenario Outline: one-empty string scores 0.0
    When I compute the TokenSortRatio score between "<a>" and "<b>"
    Then the score should be exactly 0

    Examples:
      | a           | b           |
      | hello world |             |
      |             | hello world |

  @token @token-sort-ratio
  Scenario: token-reorder produces identical sorted-joined strings
    # The defining property of Token Sort Ratio: argument-order
    # reordering of tokens leaves the score at 1.0 because the
    # sorted-joined representations are byte-identical on both sides.
    When I compute the TokenSortRatio score between "alpha beta gamma" and "gamma alpha beta"
    Then the score should be exactly 1

  @token @token-sort-ratio
  Scenario: score is symmetric across argument order
    # Tokenise is deterministic; sort.Strings is byte-lex stable;
    # indelRatio is symmetric — so the composite TokenSortRatio is
    # symmetric bit-for-bit (no float tolerance needed).
    When I compute the TokenSortRatio score between "alpha beta" and "alpha beta gamma"
    And I compute the second TokenSortRatio score between "alpha beta gamma" and "alpha beta"
    Then both TokenSortRatio scores should be equal
