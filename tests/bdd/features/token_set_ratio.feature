# Engineering source: RapidFuzz documentation,
# https://rapidfuzz.github.io/RapidFuzz/Usage/fuzz.html#token-set-ratio
# — the canonical modern reference for the Token Set Ratio algorithm.
# Underlying DP source: Wagner, R. A., & Fischer, M. J. (1974). "The
# string-to-string correction problem." Journal of the ACM 21(1):168-173
# — the LCS-subsequence dynamic-programming recurrence consumed via the
# shared token_indel.go kernel.
#
# Empty-set DEVIATION: RapidFuzz issue #110
# (https://github.com/rapidfuzz/RapidFuzz/issues/110) — TokenSetRatio
# returns 0.0 (NOT 1.0) when either tokenised input is empty, mirroring
# fuzzywuzzy bug-for-bug. Other tokenised algorithms in the catalogue
# (TokenJaccard, MongeElkan) follow the standard both-empty → 1.0
# convention; TokenSetRatio is the documented exception.
#
# Cross-validation: RapidFuzz 3.14.5 via the corpus at
# testdata/cross-validation/token-ratios/vectors.json. The byte-stable
# agreement is asserted by token_ratio_cross_validation_test.go.
#
# Surface: TokenSetRatioScore is the dispatched byte-path score.
# Tokeniser-divergence note (OQ-1 RESOLUTION LOCKED in plan 06-01):
# fuzzymatch's Tokenise is identifier-aware (camelCase / snake_case /
# kebab-case / dot-case + lowercasing); RapidFuzz tokenises via Python
# str.split (whitespace-only, case-preserving). For whitespace-only
# lowercase ASCII inputs (the BDD scenarios below) the two agree.

@token
Feature: Token Set Ratio (three-way Indel max with bug-for-bug RapidFuzz empty-set behaviour)
  Token Set Ratio tokenises both sides, deduplicates each tokens list
  to a set, computes the sorted intersection / diff_a_to_b / diff_b_to_a
  slices, then either short-circuits to 1.0 (when one set is a subset
  of the other AND the intersection is non-empty) or takes the maximum
  of three Indel ratios over the constructed-string forms. Returns
  scores in [0.0, 1.0]. Symmetric across argument order.

  DEVIATION (LOCKED RapidFuzz issue #110): when either tokenised input
  is empty, the function returns 0.0 — NOT 1.0. This is the only
  algorithm in the catalogue with this deviation.

  @token @token-set-ratio
  Scenario Outline: Canonical reference vectors
    When I compute the TokenSetRatio score between "<a>" and "<b>"
    Then the score should be approximately <score> within 0.0001

    Examples:
      | a                      | b                      | score  |
      | alpha beta             | alpha beta gamma       | 1.0000 |
      | alpha beta gamma       | alpha beta             | 1.0000 |
      | hello world            | world peace            | 0.6364 |
      | abc def                | xyz qrs                | 0.1429 |
      | hello                  | world                  | 0.2000 |
      | alpha beta             | beta alpha alpha beta  | 1.0000 |

  @token @token-set-ratio
  Scenario: identical strings score 1.0
    # The a == b identity short-circuit fires before Tokenise.
    When I compute the TokenSetRatio score between "hello world" and "hello world"
    Then the score should be exactly 1

  @token @token-set-ratio
  Scenario: both-empty strings return 0.0 (RapidFuzz issue #110 deviation from catalogue convention)
    # LOCKED DEVIATION: TokenSetRatioScore("", "") returns 0.0
    # — NOT 1.0. The empty-input gate fires BEFORE the identity
    # short-circuit per the LOCKED bug-for-bug RapidFuzz issue #110
    # / fuzzywuzzy parity. Other tokenised algorithms in the
    # catalogue (TokenJaccard, MongeElkan) follow the standard
    # both-empty → 1.0 convention; TokenSetRatio is the documented
    # exception.
    When I compute the TokenSetRatio score between "" and ""
    Then the score should be exactly 0

  @token @token-set-ratio
  Scenario: pure-separator inputs return 0.0 (post-Tokenise empty-set deviation)
    # LOCKED DEVIATION post-Tokenise: when both inputs Tokenise to
    # an empty slice (pure-separator runs) AND the raw strings
    # differ (so the identity short-circuit does NOT fire),
    # TokenSetRatio returns 0.0 — NOT 1.0. Same RapidFuzz issue
    # #110 deviation as the both-empty-strings case, but reached via
    # the post-Tokenise gate rather than the pre-Tokenise gate.
    When I compute the TokenSetRatio score between " " and "  "
    Then the score should be exactly 0

  @token @token-set-ratio
  Scenario Outline: one-empty string scores 0.0
    When I compute the TokenSetRatio score between "<a>" and "<b>"
    Then the score should be exactly 0

    Examples:
      | a           | b           |
      | alpha beta  |             |
      |             | alpha beta  |

  @token @token-set-ratio
  Scenario Outline: subset short-circuit returns 1.0
    # When the intersection is non-empty AND one of the token sets is
    # a subset of the other (one diff is empty), TokenSetRatio
    # short-circuits to 1.0 directly without computing the three
    # Indel ratios. This is RESEARCH.md Pattern 5 critical landmine 2.
    When I compute the TokenSetRatio score between "<a>" and "<b>"
    Then the score should be exactly 1

    Examples:
      | a                      | b                      |
      | alpha beta             | alpha beta gamma       |
      | alpha beta gamma       | alpha beta             |
      | alpha                  | alpha beta gamma       |
      | alpha beta gamma       | alpha                  |

  @token @token-set-ratio
  Scenario: token-set score is symmetric
    # Tokenise is deterministic; set construction is order-independent;
    # the three-way max operator is order-insensitive — so the
    # composite TokenSetRatio is symmetric bit-for-bit (no float
    # tolerance needed).
    When I compute the TokenSetRatio score between "hello world" and "world peace"
    And I compute the second TokenSetRatio score between "world peace" and "hello world"
    Then both TokenSetRatio scores should be equal
