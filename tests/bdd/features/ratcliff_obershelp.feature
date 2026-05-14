# Primary source: Ratcliff, J. W., & Metzener, D. E. (1988). "Pattern matching:
# the gestalt approach." Dr. Dobb's Journal, 13(7):46-51 — the recursive
# longest-common-substring decomposition for gestalt pattern matching.
#
# Cross-validation: Python difflib.SequenceMatcher(autojunk=False).ratio()
# (PSF licence, stdlib) — reference vectors only, no code copying. The
# autojunk=False qualifier is LOAD-BEARING: difflib's default autojunk=True
# is a performance heuristic that drops "junk" characters when len(b) >= 200
# and a character appears in >= 1% of positions; this is NOT the Ratcliff-
# Obershelp algorithm. RESEARCH.md Pitfall 2.
#
# Surface: RatcliffObershelpScore is the dispatched byte-path score (this
# feature exercises that surface end-to-end). The rune-path variant
# RatcliffObershelpScoreRunes is covered by unit tests, not BDD.
#
# NB: NO symmetry scenario per OQ-1 resolution (LOCKED 2026-05-14). Ratcliff-
# Obershelp is asymmetric by design to preserve byte-for-byte difflib
# equivalence. The asymmetric-by-design contract is verified by unit tests
# (TestRatcliffObershelp_AsymmetryPin) and the cross-algorithm consistency
# test in plan 04-05.

Feature: Ratcliff-Obershelp similarity (difflib-equivalent)
  Ratcliff-Obershelp ("gestalt pattern matching") is the recursive longest-
  common-substring decomposition. Behaves byte-for-byte like Python's
  difflib.SequenceMatcher(autojunk=False).ratio(). NOT symmetric in
  argument order — mirrors difflib per CPython bpo-37004.

  Scenario Outline: canonical Dr. Dobb's 1988 reference vectors
    When I compute the Ratcliff-Obershelp score between "<a>" and "<b>"
    Then the score should be approximately <score> within 0.0001

    Examples:
      | a         | b                        | score  |
      | WIKIMEDIA | WIKIMANIA                | 0.7778 |
      | GESTALT   | GESTALT_PATTERN_MATCHING | 0.4516 |

  Scenario: identical strings score 1.0
    When I compute the Ratcliff-Obershelp score between "user_id" and "user_id"
    Then the score should be exactly 1

  Scenario: both-empty strings score 1.0
    # The both-empty convention: RatcliffObershelpScore returns 1.0
    # vacuously (mirrors difflib(autojunk=False).ratio() on two empty
    # strings). The identity short-circuit handles this case.
    When I compute the Ratcliff-Obershelp score between "" and ""
    Then the score should be exactly 1

  Scenario: one-empty string scores 0.0
    When I compute the Ratcliff-Obershelp score between "abc" and ""
    Then the score should be exactly 0

  Scenario: 200+ character autojunk-sensitive input proves autojunk=False
    # This scenario uses a 205-character input pair that would trigger
    # Python difflib's autojunk heuristic if it were enabled (len(b) >= 200
    # and a character appears in >= 1% of positions). The pair is constructed
    # in the Go step as:
    #   a = "a"*100 + "x"*5 + "a"*100   (205 chars)
    #   b = "a"*50  + "y"*5 + "a"*150   (205 chars)
    # The expected score 0.7317 matches Python
    # difflib.SequenceMatcher(autojunk=False, a=a, b=b).ratio() — verifying
    # this implementation does NOT have an autojunk-like character-dropping
    # heuristic. RESEARCH.md Pitfall 2 closure.
    When I compute the Ratcliff-Obershelp score for the autojunk-sensitive pair
    Then the score should be approximately 0.7317 within 0.0001
