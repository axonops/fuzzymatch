# Engineering source: RapidFuzz documentation,
# https://rapidfuzz.github.io/RapidFuzz/Usage/fuzz.html#partial-ratio
# — the canonical modern reference for the Partial Ratio algorithm.
# Three-region iteration (Region 1 left tail / Region 2 middle /
# Region 3 right tail) plus the s1_char_set early-skip pattern were
# transcribed structurally from RapidFuzz's MIT-licensed
# fuzz_py.py::_partial_ratio_impl — see 06-RESEARCH.md Pattern 6 +
# Example 4 for the verbatim Go structural transcription template.
# Underlying DP source: Wagner, R. A., & Fischer, M. J. (1974). "The
# string-to-string correction problem." Journal of the ACM
# 21(1):168-173 — the LCS-subsequence dynamic-programming recurrence
# consumed via the shared token_indel.go kernel.
#
# Pitfall 3 (06-RESEARCH.md): a naive single-loop implementation
# covers only Region 2 (the standard sliding window). The Region 1
# (left tail) and Region 3 (right tail) iterations are LOAD-BEARING
# for matches that "hang off" the start or end of the longer string.
# The keystone fixtures `("abc","ab")` and `("abc","bc")` are pinned
# in dedicated scenarios below to prevent regression.
#
# Cross-validation: RapidFuzz 3.14.5 via the corpus at
# testdata/cross-validation/token-ratios/vectors.json. The byte-stable
# agreement is asserted by token_ratio_cross_validation_test.go
# (`/partial_bytes` sub-test; activated in plan 06-03).
#
# Surface (Phase 8.5 Q5 LOCKED — plan 08.5-03): PartialRatio ships a
# single byte path (PartialRatioScore, dispatched in
# dispatch[AlgoPartialRatio]). The former rune-variant
# (PartialRatioScoreRunes) was removed because token-tier algorithms
# operate on the output of Tokenise (rune-aware), so the byte-level
# Indel kernel is correct on post-Tokenise byte strings.
#
# DOES NOT inherit TokenSetRatio's RapidFuzz issue #110 deviation —
# PartialRatio follows the catalogue's standard both-empty → 1.0
# convention.

@token
Feature: Partial Ratio (sliding-window Indel-formula similarity over the longer string, byte surface)
  Partial Ratio operates at the character level (NO tokenisation). It
  iterates the shorter input against the longer input in three regions
  — Region 1 (left tail), Region 2 (middle full-length windows), and
  Region 3 (right tail) — and returns the maximum Indel-formula
  similarity across all alignments. The s1_char_set early-skip prunes
  alignments whose last/first character does not appear in the shorter
  input (load-bearing for the pathological budget per Pitfall 3).
  Returns scores in [0.0, 1.0]. Symmetric across argument order.

  Per Phase 8.5 Q5 LOCKED, PartialRatio ships a single byte-path
  surface (PartialRatioScore — dispatched). The former rune-path
  variant was removed in plan 08.5-03.

  @token @partial-ratio @byte-path
  Scenario Outline: Canonical reference vectors (byte path)
    When I compute the PartialRatio score between "<a>" and "<b>"
    Then the score should be approximately <score> within 0.0001

    Examples:
      | a       | b                  | score  |
      | YANKEES | NEW YORK YANKEES   | 1.0000 |
      | hello   | hello world        | 1.0000 |
      | world   | hello world        | 1.0000 |
      | abcd    | xabcy              | 0.7500 |
      | abc     | xyzzz              | 0.0000 |

  @token @partial-ratio
  Scenario: identical strings score 1.0
    # The a == b identity short-circuit fires BEFORE any byte slicing
    # / charSet construction.
    When I compute the PartialRatio score between "hello world" and "hello world"
    Then the score should be exactly 1

  @token @partial-ratio
  Scenario: both-empty strings score 1.0
    # PartialRatio follows the catalogue's standard both-empty → 1.0
    # convention. Does NOT inherit TokenSetRatio's RapidFuzz issue
    # #110 deviation — empty inputs at the character level have a
    # meaningful (vacuous) interpretation.
    When I compute the PartialRatio score between "" and ""
    Then the score should be exactly 1

  @token @partial-ratio
  Scenario Outline: one-empty string scores 0.0 (byte path)
    When I compute the PartialRatio score between "<a>" and "<b>"
    Then the score should be exactly 0

    Examples:
      | a       | b       |
      | hello   |         |
      |         | hello   |

  @token @partial-ratio @byte-path
  Scenario: byte path is symmetric
    # The shorter-longer swap is internal to the algorithm; indelRatio
    # is symmetric — so PartialRatioScore(a, b) == PartialRatioScore(b, a)
    # bit-for-bit (no float tolerance needed).
    When I compute the PartialRatio score between "YANKEES" and "NEW YORK YANKEES"
    And I compute the second PartialRatio score between "NEW YORK YANKEES" and "YANKEES"
    Then both PartialRatio scores should be equal

  @token @partial-ratio @pitfall-3
  Scenario: Region 1 left-tail alignment wins (Pitfall 3 keystone)
    # 06-RESEARCH.md Pitfall 3 — KEYSTONE fixture. A naive
    # single-loop implementation `for i := 0; i <= n-m; i++` covers
    # only Region 2. For ("abc","ab") with m=2 / n=3, Region 2 at i=0
    # catches the alignment ("ab" vs "ab" → 1.0). The Pitfall-3
    # keystone tests the implementation includes the Region 1 left-tail
    # iteration so future drift to a single-loop is caught.
    When I compute the PartialRatio score between "abc" and "ab"
    Then the score should be exactly 1

  @token @partial-ratio @pitfall-3
  Scenario: Region 3 right-tail alignment wins (Pitfall 3 keystone)
    # 06-RESEARCH.md Pitfall 3 — KEYSTONE fixture. For ("abc","bc")
    # with m=2 / n=3, Region 2 at i=1 catches the alignment ("bc" vs
    # "bc" → 1.0). The Pitfall-3 keystone tests the implementation
    # includes the Region 3 right-tail iteration so future drift to a
    # single-loop is caught.
    When I compute the PartialRatio score between "abc" and "bc"
    Then the score should be exactly 1
