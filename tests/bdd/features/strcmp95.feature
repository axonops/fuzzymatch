# Primary source: Winkler, W. E. (1994). "Advanced methods for record linkage."
# Proceedings of the Section on Survey Research Methods, ASA: 467-472, §3.
#
# Cross-validation reference: U.S. Census Bureau (1995) strcmp95.c
# (public domain — U.S. Government work; consulted ONLY for reference
# vectors per .claude/skills/algorithm-licensing-standards).
#
# Score normalisation: Strcmp95 layers four adjustments atop a Jaro pass —
# similar-character credit (Winkler 1994 §3 table), Winkler prefix boost,
# and the long-string adjustment. Score in [0.0, 1.0].

Feature: Strcmp95 similarity
  Strcmp95 is Winkler's 1994 enhancement of Jaro-Winkler for record-linkage
  and surname matching. It adds a similar-character credit pass over
  unmatched positions and a long-string adjustment on top of the
  Jaro-Winkler prefix boost. ASCII-only; for Unicode input, normalise via
  fuzzymatch.Normalise first.

  Scenario Outline: canonical reference vectors
    When I compute the Strcmp95 score between "<a>" and "<b>"
    Then the score should be approximately <score> within 0.0001

    Examples:
      | a         | b          | score  |
      | MARTHA    | MARHTA     | 0.9676 |
      | DWAYNE    | DUANE      | 0.8925 |
      | DIXON     | DICKSONX   | 0.8517 |
      | HAMINGTON | HAMMINGTON | 0.9820 |

  Scenario: identical strings score 1.0
    When I compute the Strcmp95 score between "user_id" and "user_id"
    Then the score should be exactly 1

  Scenario: both-empty strings score 1.0
    When I compute the Strcmp95 score between "" and ""
    Then the score should be exactly 1

  Scenario: one-empty string scores 0.0
    When I compute the Strcmp95 score between "abc" and ""
    Then the score should be exactly 0

  Scenario: score is symmetric
    When I compute the Strcmp95 score between "MARTHA" and "MARHTA"
    And I compute the second Strcmp95 score between "MARHTA" and "MARTHA"
    Then both Strcmp95 scores should be equal

  Scenario: similar-character table fires on DWAYNE / DUANE
    # The W/U pair is in the Winkler 1994 §3 similar-character table.
    # Strcmp95 (0.8925) should strictly exceed JaroWinkler (0.8400 — pinned
    # against the published Winkler 1990 value for this pair) — confirming
    # RESEARCH.md Pitfall 1 warning sign #2: similar-character table fires.
    When I compute the Strcmp95 score between "DWAYNE" and "DUANE"
    Then the score should be approximately 0.8925 within 0.0001
    When I compute the JaroWinkler score between "DWAYNE" and "DUANE"
    Then the score should be approximately 0.8400 within 0.0001
