# Primary source: Taft, R. L. (1970). Name search techniques.
# New York State Identification and Intelligence System,
# Special Report No. 1. Albany, NY. (Algorithmic origin.)
#
# Canonical algorithm description:
# Knuth, D. E. (1973). The Art of Computer Programming, Vol. 3, §6.4.
# Addison-Wesley. (Primary source for fresh transcription.)
#
# Note: Taft 1970 is a NY State Special Report not available through
# academic publishers; Knuth's description in TAOCP Vol. 3 §6.4 is the
# authoritative algorithm description used for this implementation.
#
# Variant: original NYSIIS-1970, 6-character truncation, each rule once.
# Modified-NYSIIS (jellyfish variant, no truncation) REJECTED per
# CONTEXT.md §2 LOCKED. The truncation gate test below (Catherine → CATARA)
# is the LOAD-BEARING discriminator — jellyfish emits "CATARAN" (7 chars).
#
# Cross-validation: jellyfish==1.2.1 (BSD-2-Clause) — reference vectors only.
# MIT-licensed Go ports NOT consulted: github.com/UjjwalAyyangar/go-jellyfish.

@phonetic
Feature: NYSIIS (Taft 1970 / Knuth TAOCP §6.4) — original 6-char truncation variant
  NYSIISCode encodes an ASCII name to a maximum 6-character uppercase code
  using the original Taft-1970 procedure (each rule applied once) and truncated
  to 6 characters. NYSIISScore returns binary 1.0 if both inputs produce the
  same non-empty code, 0.0 otherwise.

  @phonetic @nysiis
  Scenario Outline: Brown/Browne canonical pair — literature reference vectors RV-N1 and RV-N2
    # RV-N1: Brown → BRAN (canonical Knuth TAOCP Vol. 3 §6.4 pair, part 1).
    # RV-N2: Browne → BRAN (matches RV-N1 — OW + W-after-vowel removal).
    # NYSIISScore("Brown", "Browne") must be 1.0.
    When I compute the NYSIIS code of "<name>"
    Then the code should be "<code>"

    Examples:
      | name   | code |
      | Brown  | BRAN |
      | Browne | BRAN |

  @phonetic @nysiis
  Scenario: Brown/Browne phonetic match scores 1.0
    When I compute the NYSIIS score between "Brown" and "Browne"
    Then the score should be exactly 1

  @phonetic @nysiis
  Scenario: Robert encodes to RABAD — literature reference vector RV-N3
    # RV-N3: Robert → RABAD (Knuth TAOCP Vol. 3 §6.4 reference).
    # O→A, B stays, E→A, RT suffix → D → RABAD.
    When I compute the NYSIIS code of "Robert"
    Then the code should be "RABAD"

  @phonetic @nysiis
  Scenario: Catherine 6-char truncation gate — LOAD-BEARING Taft-1970 discriminator
    # RV-N4 LOAD-BEARING: Catherine → CATARA (6 chars, Taft-1970 truncated).
    # jellyfish.nysiis("Catherine") = "CATARAN" (7 chars — modified NYSIIS).
    # This test fails with length 7 if the modified-NYSIIS variant is shipped.
    # See RESEARCH.md §7 Pitfall 7.B and CONTEXT.md §2 LOCKED.
    When I compute the NYSIIS code of "Catherine"
    Then the code should be "CATARA"

  @phonetic @nysiis
  Scenario: Katherine encodes to CATARA — same as Catherine (RV-N5)
    When I compute the NYSIIS code of "Katherine"
    Then the code should be "CATARA"

  @phonetic @nysiis
  Scenario: Catherine/Katherine phonetic match scores 1.0
    When I compute the NYSIIS score between "Catherine" and "Katherine"
    Then the score should be exactly 1

  @phonetic @nysiis
  Scenario: identity — any string scores 1.0 against itself
    When I compute the NYSIIS score between "Brown" and "Brown"
    Then the score should be exactly 1

  @phonetic @nysiis
  Scenario: both-empty inputs score 1.0 (standard catalogue convention)
    # Empty → empty code. NYSIISScore("", "") = 1.0 per algorithm-correctness-
    # standards both-empty convention (identity short-circuit fires).
    When I compute the NYSIIS score between "" and ""
    Then the score should be exactly 1

  @phonetic @nysiis
  Scenario Outline: one-empty input scores 0.0
    When I compute the NYSIIS score between "<a>" and "<b>"
    Then the score should be exactly 0

    Examples:
      | a     | b    |
      | Brown |      |
      |       | Brown |

  @phonetic @nysiis
  Scenario: non-matching codes score 0.0
    # Brown (BRAN) and Robert (RABAD) have different codes → 0.0.
    When I compute the NYSIIS score between "Brown" and "Robert"
    Then the score should be exactly 0
