# Primary source: Moore, G. B., Kuhns, J. L., Trefftzs, J. L., Montgomery, C. A.
# (1977). Accessing individual records from personal data files using non-unique
# identifiers. National Bureau of Standards (now NIST), Technical Note 943.
# Available at: https://nvlpubs.nist.gov/nistpubs/Legacy/TN/nbstechnicalnote943.pdf
#
# Cross-validation: jellyfish==1.2.1 (BSD-2-Clause) — reference vectors only.
# MIT-licensed Go ports NOT consulted: github.com/UjjwalAyyangar/go-jellyfish.
#
# MRA is the only algorithm in the fuzzymatch catalogue with a (bool, int)
# public return shape: MRACompare(a, b) → (matched bool, simScore int) per
# CONTEXT.md §6 LOCKED and spec line 691. The int is the raw 0-6 NBS
# similarity counter.

@phonetic
Feature: MRA (Moore 1977 / NBS Tech Note 943) — Match Rating Approach
  MRACode encodes an ASCII name per NBS-943 rules (vowel-deletion except
  leading + double-consonant dedup + first-3-last-3 truncation if len > 6).
  MRACompare returns (matched bool, simScore int) where simScore ∈ [0,6] is
  the raw NBS similarity counter. MRAScore returns binary 1.0/0.0 = matched.

  @phonetic @mra
  Scenario Outline: Byrne / Boern canonical encoding — literature reference vectors RV-M1 and RV-M2
    # RV-M1: Byrne → BYRN (B leading, y→Y consonant, r→R, n→N, e→dropped non-leading vowel).
    # RV-M2: Boern → BRN  (B leading, o→dropped, e→dropped, r→R, n→N).
    # Source: jellyfish==1.2.1 testdata (lines 1–2); NBS Tech Note 943 encoding rules.
    When I compute the MRA code of "<name>"
    Then the code should be "<code>"

    Examples:
      | name  | code |
      | Byrne | BYRN |
      | Boern | BRN  |

  @phonetic @mra
  Scenario Outline: Smith / Smyth canonical encoding — literature reference vectors RV-M3 and RV-M4
    # RV-M3: Smith → SMTH (i removed as vowel).
    # RV-M4: Smyth → SMYTH (Y is NOT a vowel in MRA; kept as consonant).
    When I compute the MRA code of "<name>"
    Then the code should be "<code>"

    Examples:
      | name  | code  |
      | Smith | SMTH  |
      | Smyth | SMYTH |

  @phonetic @mra
  Scenario: Smith/Smyth phonetic match — RV-M7 — threshold gate
    # MRACompare("Smith","Smyth"): SMTH (len 4) vs SMYTH (len 5).
    # sum_len=9 → threshold=3 (Table A: 7 < sum ≤ 11 → threshold 3).
    # After L→R + R→L elimination: unmatched_A=0, unmatched_B=1.
    # similarity = 6 - max(0,1) = 5. 5 >= 3 → match.
    When I compare with MRA "Smith" and "Smyth"
    Then the MRA match should be true
    And the MRA similarity should be 5

  @phonetic @mra
  Scenario: length-difference >= 3 auto-mismatch gate — RV-M8
    # MRACode("Ad")="AD" (len 2); MRACode("ZachariahMontgomery")="ZCHMRY" (len 6).
    # |2 - 6| = 4 >= 3 → automatic mismatch per NBS Tech Note 943 step 1.
    # fuzzymatch returns (false, 0) not an error per CONTEXT.md §6 LOCKED.
    When I compare with MRA "Ad" and "ZachariahMontgomery"
    Then the MRA match should be false
    And the MRA similarity should be 0

  @phonetic @mra
  Scenario: MRAScore Smith/Smyth returns 1.0 — wraps MRACompare — RV-M12
    # MRAScore(a, b) == 1.0 iff MRACompare(a, b).matched.
    # Smith/Smyth: MRACompare returns (true, 5) → MRAScore returns 1.0.
    When I compute the MRA score between "Smith" and "Smyth"
    Then the score should be exactly 1

  @phonetic @mra
  Scenario: identity — any string scores 1.0 against itself
    When I compute the MRA score between "Byrne" and "Byrne"
    Then the score should be exactly 1

  @phonetic @mra
  Scenario: both-empty strings score 1.0 — RV-M11
    # Empty → empty code. MRACompare("","")=(true,6) per algorithm-correctness-standards.
    # sum_len=0 → threshold=5; no chars to eliminate; sim=6-0=6. 6>=5 → match.
    When I compute the MRA score between "" and ""
    Then the score should be exactly 1

  @phonetic @mra
  Scenario: Kathrynoglin first-3-last-3 truncation gate — RV-M6
    # Pre-truncation form KTHRYNGLN (len 9 > 6) → first 3 "KTH" + last 3 "GLN" = "KTHGLN".
    # This test fails with a 9-char code if the truncation rule is not applied.
    When I compute the MRA code of "Kathrynoglin"
    Then the code should be "KTHGLN"

  @phonetic @mra
  Scenario: length-difference auto-mismatch returns 0.0 via MRAScore
    When I compute the MRA score between "Ad" and "ZachariahMontgomery"
    Then the score should be exactly 0

  @phonetic @mra
  Scenario Outline: one-empty input scores 0.0
    When I compute the MRA score between "<a>" and "<b>"
    Then the score should be exactly 0

    Examples:
      | a     | b     |
      | Byrne |       |
      |       | Byrne |
