# Primary source: Boytsov, L. (2011). "Indexing methods for approximate
# dictionary searching: comparative analysis." ACM Journal of Experimental
# Algorithmics, 16, Article 1. (OSA formulation cited from §3.1.)
#
# Historical source: Damerau, F. J. (1964). "A technique for computer
# detection and correction of spelling errors." Communications of the ACM,
# 7(3):171-176.
#
# Score normalisation: 1 - distance / max(len(a), len(b)).
# OSA constraint: each substring participates in at most one transposition.
# DL-OSA is NOT a strict metric — triangle inequality may fail on contrived inputs.
# Reference vectors from Boytsov 2011 §3.1.

Feature: Damerau-Levenshtein OSA similarity algorithm
  The Damerau-Levenshtein OSA (Optimal String Alignment) distance extends
  Levenshtein by counting adjacent transpositions as a single edit, with
  the restriction that each substring participates in at most one
  transposition. OSA is NOT a strict metric.

  Scenario: OSA discriminating reference vector
    # This vector proves OSA != Full DL: Full DL returns distance 2 (score ~0.3333)
    # for the same pair. OSA returns distance 3 (score 0.0) because the OSA
    # restriction forbids re-editing characters after a transposition.
    # Boytsov 2011 §3.1 — the canonical discriminating vector.
    When I compute the DamerauLevenshteinOSA distance between "ca" and "abc"
    Then the distance should be 3

  Scenario Outline: canonical reference vectors
    When I compute the DamerauLevenshteinOSA score between "<a>" and "<b>"
    Then the score should be approximately <score> within 0.0001

    Examples:
      | a   | b   | score  |
      | ab  | ba  | 0.5000 |
      | ca  | abc | 0.0000 |
      | abc | abc | 1.0000 |

  Scenario: both-empty strings score 1.0
    When I compute the DamerauLevenshteinOSA score between "" and ""
    Then the score should be exactly 1.0

  Scenario: score is symmetric
    When I compute the DamerauLevenshteinOSA score between "ab" and "ba"
    And I compute the second DamerauLevenshteinOSA score between "ba" and "ab"
    Then both DamerauLevenshteinOSA scores should be equal
