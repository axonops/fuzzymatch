# Primary source: Knuth, D. E. (1973). The Art of Computer Programming,
# Vol. 3: Sorting and Searching, Section 6.4. Addison-Wesley.
# Algorithmic origin: Russell, R. C., Odell, M. K. (1918, 1922). U.S. Patents
# 1,261,167 and 1,435,663.
#
# Cross-validation: jellyfish==1.2.1 (BSD-2-Clause) — reference vectors only.
# jellyfish 1.2.1 also uses the Knuth/Census H/W-skip variant (confirmed by
# direct read of jellyfish/src/soundex.rs).
#
# Variant choice (LOAD-BEARING): Knuth/Census (American Soundex).
#   - H and W are SKIPPED (not separators).
#   - Tymczak → "T522" (NOT "T520" which SQL/MySQL Soundex returns).
#   - Ashcraft = Ashcroft = "A261" (H/W transparent).
#
# Non-ASCII input: silently dropped per CONTEXT.md §5. Not covered by BDD
# (BDD tests cover the primary public-API contract; edge cases live in unit
# tests and fuzz corpus).

Feature: Soundex phonetic similarity (Knuth/Census variant)
  Soundex encodes English names into a 4-character phonetic code (1 letter
  + 3 digits, zero-padded). The Knuth/Census variant is used: H and W are
  transparent skips (NOT separators). SoundexScore returns 1.0 if the two
  codes match (or both inputs are identical), and 0.0 otherwise.

  Scenario Outline: literature reference vectors (Knuth TAOCP §6.4)
    When I compute the Soundex code of "<input>"
    Then the code should be "<code>"

    Examples:
      | input    | code |
      | Robert   | R163 |
      | Rupert   | R163 |
      | Rubin    | R150 |
      | Tymczak  | T522 |
      | Ashcraft | A261 |
      | Smith    | S530 |
      |          |      |

  Scenario: identical strings score 1.0
    # The identity short-circuit fires before any SoundexCode computation.
    When I compute the Soundex score between "Robert" and "Robert"
    Then the score should be exactly 1

  Scenario: both-empty strings score 1.0
    # Both-empty convention: the identity short-circuit returns 1.0.
    When I compute the Soundex score between "" and ""
    Then the score should be exactly 1

  Scenario: Tymczak Knuth/Census variant gate (NOT SQL T520)
    # LOAD-BEARING: this scenario pins the variant choice unambiguously.
    # "Tymczak" encodes as "T522" under Knuth/Census (H/W skip rule) and
    # as "T520" under the SQL/MySQL variant (H/W are separators there).
    When I compute the Soundex code of "Tymczak"
    Then the code should be "T522"

  Scenario: Ashcraft and Ashcroft share the same code (H/W-handling gate)
    # The 'h' in Ashcraft and Ashcroft is transparent — it does NOT break
    # the adjacent-group collapse between 's' and 'c' (both group 2).
    When I compute the Soundex score between "Ashcraft" and "Ashcroft"
    Then the score should be exactly 1

  Scenario: non-matching names score 0.0
    When I compute the Soundex score between "Robert" and "Smith"
    Then the score should be exactly 0

  Scenario: one-empty string scores 0.0
    When I compute the Soundex score between "Robert" and ""
    Then the score should be exactly 0
