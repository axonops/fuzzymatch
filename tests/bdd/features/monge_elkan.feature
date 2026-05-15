# Primary source: Monge, A. E., & Elkan, C. P. (1996). "The field
# matching problem: algorithms and applications." Proceedings of the
# Second International Conference on Knowledge Discovery and Data Mining
# (KDD'96): 267-270, §3 — the per-token-max-mean field matching
# algorithm.
#
# Cross-validation: hand-derived RV-ME1..RV-ME6 reference vectors per
# 06-CONTEXT.md §1b LOCKED — no RapidFuzz toolchain (the RapidFuzz
# cross-validation corpus does NOT include Monge-Elkan entries because
# RapidFuzz's default inner metric may not match this project's
# JaroWinkler default; the corpus is for the four Indel-based ratios
# only).
#
# LOAD-BEARING per 06-CONTEXT.md §3 LOCKED: the asymmetry direction-
# sensitivity scenario below is the BDD-layer regression detector for
# the RV-ME4 vs RV-ME6 input-swap pair. A silent direction swap inside
# MongeElkanScore would cause both calls to return the same value —
# both would still be in [0, 1] and still pass RangeBounds + Identity —
# but the "differ by more than 0.1" assertion catches the regression
# alongside the unit-test (TestMongeElkanScore_AsymmetryDirectionSensitive)
# and property-test (TestProp_MongeElkanScore_AsymmetricWhenTokenCountAsymmetric)
# gates.
#
# Surface: MongeElkanScore is the ASYMMETRIC direct surface;
# MongeElkanScoreSymmetric is the SYMMETRIC (arithmetic mean of the two
# directions) variant; the dispatch slot AlgoMongeElkan binds the
# SYMMETRIC variant with AlgoJaroWinkler default inner per
# CONTEXT.md §4 LOCKED.

@token
Feature: Monge-Elkan (asymmetric per-token-max-mean with 17 permitted inner AlgoIDs + symmetric variant)
  Monge-Elkan tokenises each side and, for each token in A, takes the
  maximum inner-metric similarity over every token in B; then averages
  those per-token maxima. With a fixed inner metric the function is
  direction-sensitive — swapping (a, b) with the same inner generally
  produces a different score. The symmetric variant takes the arithmetic
  mean of the two directions and is order-independent by construction.
  The result lies in [0.0, 1.0]. Both surfaces ship; the dispatch slot
  AlgoMongeElkan binds the symmetric variant with AlgoJaroWinkler as
  the default inner metric per CONTEXT.md §4 LOCKED.

  @token @monge-elkan
  Scenario Outline: Canonical reference vectors (asymmetric, AlgoJaroWinkler inner)
    # RV-ME1: canonical asymmetric example with JaroWinkler inner.
    # RV-ME3: disjoint Greek-letter pair (combines several mid-scoring
    #         JaroWinkler comparisons).
    # The conventions row pins the identity short-circuit at an
    # asymmetric inner.
    When I compute the MongeElkan score between "<a>" and "<b>" with inner AlgoJaroWinkler
    Then the score should be approximately <score> within 0.0001

    Examples:
      | a            | b            | score  |
      | user create  | usr creating | 0.9125 |
      | alpha beta   | gamma delta  | 0.6917 |
      | hello world  | hello world  | 1.0000 |

  @token @monge-elkan
  Scenario Outline: Canonical reference vectors (asymmetric, AlgoLevenshtein inner)
    # RV-ME4: subset case |tA| < |tB|; the single A-token matches one
    # of the three B-tokens exactly so the max is 1.0.
    # RV-ME5: Unicode tokens (the byte path; Tokenise lowercases via
    # DefaultTokeniseOptions; Lev(café, cafe) on bytes is 0.6 with
    # max-len 5).
    When I compute the MongeElkan score between "<a>" and "<b>" with inner AlgoLevenshtein
    Then the score should be approximately <score> within 0.0001

    Examples:
      | a     | b                | score  |
      | alpha | alpha beta gamma | 1.0000 |
      | café  | cafe             | 0.6000 |

  @token @monge-elkan
  Scenario Outline: Symmetric variant canonical reference vectors (AlgoJaroWinkler inner)
    # The symmetric variant averages the two directions; for the
    # RV-ME1 pair both directions are close enough that the symmetric
    # average is near 0.9125 as well.
    When I compute the MongeElkanSymmetric score between "<a>" and "<b>" with inner AlgoJaroWinkler
    Then the score should be approximately <score> within 0.0001

    Examples:
      | a            | b            | score  |
      | hello world  | hello world  | 1.0000 |
      | alpha beta   | gamma delta  | 0.6917 |

  @token @monge-elkan
  Scenario: identical strings score 1.0 (inner irrelevant — short-circuit fires)
    When I compute the MongeElkan score between "user_id" and "user_id" with inner AlgoJaroWinkler
    Then the score should be exactly 1

  @token @monge-elkan
  Scenario: both-empty strings score 1.0 (STANDARD catalogue convention)
    # MongeElkan follows the STANDARD catalogue convention for both-empty
    # inputs — returns 1.0 (vacuous identity match). DOES NOT deviate
    # like TokenSetRatio (which returns 0.0 per the LOCKED RapidFuzz
    # issue #110 bug-for-bug compatibility). The a == b identity
    # short-circuit fires before Tokenise (inner irrelevant).
    When I compute the MongeElkan score between "" and "" with inner AlgoJaroWinkler
    Then the score should be exactly 1

  @token @monge-elkan
  Scenario Outline: one-empty string scores 0.0
    When I compute the MongeElkan score between "<a>" and "<b>" with inner AlgoJaroWinkler
    Then the score should be exactly 0

    Examples:
      | a     | b     |
      | hello |       |
      |       | hello |

  @token @monge-elkan
  Scenario: MongeElkanScore is direction-sensitive (RV-ME4 vs RV-ME6 input swap — LOAD-BEARING)
    # The LOAD-BEARING asymmetry gate at the BDD layer. With fixed
    # inner = AlgoLevenshtein:
    #   RV-ME4: MongeElkanScore("alpha",            "alpha beta gamma", Lev) = 1.0
    #   RV-ME6: MongeElkanScore("alpha beta gamma", "alpha",            Lev) ≈ 0.4667
    # The two scores differ by ≈ 0.5333 — the input swap with the SAME
    # inner is the direct evidence of direction-sensitivity. A silent
    # direction swap inside MongeElkanScore would cause both calls to
    # return the same value (both 0.4667 or both 1.0) — both would still
    # be in [0, 1] and would pass RangeBounds + Identity invariants —
    # but this scenario's `differ by more than 0.1` assertion catches
    # the regression.
    When I compute the MongeElkan score between "alpha" and "alpha beta gamma" with inner AlgoLevenshtein
    And I compute the second MongeElkan score between "alpha beta gamma" and "alpha" with inner AlgoLevenshtein
    Then the two MongeElkan scores should differ by more than 0.1

  @token @monge-elkan
  Scenario: MongeElkanScoreSymmetric is order-independent (the arithmetic mean of two directions)
    # The symmetric variant's construction (ME(A,B) + ME(B,A))/2 is
    # invariant under (a, b) swap. The exact bit-for-bit symmetry is
    # the load-bearing property that lets AlgoMongeElkan participate
    # in the standard symmetric property-test set without exemption
    # (the dispatch wrapper binds the symmetric variant per CONTEXT §4).
    When I compute the MongeElkanSymmetric score between "alpha" and "alpha beta gamma" with inner AlgoLevenshtein
    And I compute the second MongeElkanSymmetric score between "alpha beta gamma" and "alpha" with inner AlgoLevenshtein
    Then both MongeElkan scores should be equal

  @token @monge-elkan
  Scenario Outline: non-permitted inner AlgoIDs panic with the documented message
    # RESEARCH.md Pitfall 4 + CONTEXT.md §3 LOCKED — rejection of non-permitted
    # inner AlgoIDs (FINAL Phase 7 state — 5 rejected entries):
    #   - AlgoMongeElkan: self-recursion (infinite loop guard)
    #   - AlgoTokenSortRatio / AlgoTokenSetRatio / AlgoPartialRatio /
    #     AlgoTokenJaccard: token-on-token meaningless
    # All 4 phonetic AlgoIDs are PERMITTED as of Phase 7:
    #   - AlgoSoundex: plan 07-01 (14→15 entries)
    #   - AlgoDoubleMetaphone: plan 07-02 (15→16 entries)
    #   - AlgoNYSIIS: plan 07-03 (16→17 entries)
    #   - AlgoMRA: plan 07-04 (17→18 entries, FINAL)
    # The exhaustive permitted/rejected coverage lives in
    # TestMongeElkan_PanicsOnNonPermittedInner.
    When I attempt to compute the MongeElkan score between "a b" and "c d" with inner Algo<inner>
    Then the call should panic with "<phrase>"

    Examples:
      | inner           | phrase                                            |
      | MongeElkan      | not permitted as Monge-Elkan inner metric         |
      | TokenSortRatio  | not permitted as Monge-Elkan inner metric         |
      | TokenJaccard    | not permitted as Monge-Elkan inner metric         |
