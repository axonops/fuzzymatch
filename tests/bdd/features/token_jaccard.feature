# Primary source: Jaccard, P. (1912). "The distribution of the flora in
# the alpine zone." New Phytologist 11(2):37-50, p. 43 — the set
# coefficient |A ∩ B| / |A ∪ B|. Applied here to the deduplicated SET of
# tokens produced by Tokenise(s, DefaultTokeniseOptions()).
#
# Cross-validation: hand-derived RV-TJ1..RV-TJ6 reference vectors per
# CONTEXT.md §1b LOCKED — no RapidFuzz toolchain (the RapidFuzz
# cross-validation corpus at testdata/cross-validation/token-ratios/
# vectors.json covers only the four Indel-based ratios; TokenJaccard is
# NOT in the corpus per CONTEXT §1b LOCKED).
#
# DISTINCT from Q-Gram Jaccard (qgram_jaccard.feature): TokenJaccard uses
# SET semantics on the deduplicated token set; Q-Gram Jaccard uses
# MULTISET semantics on q-gram counts. RESEARCH.md Pattern 8 establishes
# the semantic divergence is intentional — token presence is a binary
# signal, q-gram presence is a multiplicity signal.
#
# Surface: TokenJaccardScore is the dispatched byte-path score. There is
# no rune-path variant: Tokenise is UTF-8-aware so the rune semantic is
# already preserved at the tokenisation layer (per token_sort_ratio.go's
# CONTEXT §6 LOCKED).

@token
Feature: TokenJaccard (set-Jaccard over Tokenise output, distinct from Q-Gram Jaccard's multiset semantics)
  TokenJaccard tokenises each side via Tokenise(s, DefaultTokeniseOptions()),
  deduplicates each token list to a set, and computes the Jaccard
  similarity |A ∩ B| / |A ∪ B| over the two sets. The result lies in
  [0.0, 1.0]. Symmetric across argument order. Identical inputs (or
  inputs whose token sets are identical after dedup) score 1.0. Token
  multiplicity is COLLAPSED — TokenJaccard does NOT count repeated
  tokens (distinct from Q-Gram Jaccard which counts q-gram
  multiplicities).

  @token @token-jaccard
  Scenario Outline: Canonical reference vectors
    When I compute the TokenJaccard score between "<a>" and "<b>"
    Then the score should be approximately <score> within 0.0001

    Examples:
      | a                     | b                     | score  |
      | a b c                 | b c d                 | 0.5000 |
      | a b                   | a b c                 | 0.6667 |
      | a a b                 | a b                   | 1.0000 |
      | a b c                 | x y z                 | 0.0000 |
      | alpha beta gamma delta | alpha beta epsilon zeta | 0.3333 |

  @token @token-jaccard
  Scenario: identical strings score 1.0
    When I compute the TokenJaccard score between "hello world" and "hello world"
    Then the score should be exactly 1

  @token @token-jaccard
  Scenario: both-empty strings score 1.0 (STANDARD catalogue convention)
    # TokenJaccard follows the STANDARD catalogue convention for both-empty
    # inputs — returns 1.0 (vacuous identity match). DOES NOT deviate like
    # TokenSetRatio (which returns 0.0 per the LOCKED RapidFuzz issue #110
    # bug-for-bug compatibility). The a == b identity short-circuit fires
    # before Tokenise.
    When I compute the TokenJaccard score between "" and ""
    Then the score should be exactly 1

  @token @token-jaccard
  Scenario Outline: one-empty string scores 0.0
    When I compute the TokenJaccard score between "<a>" and "<b>"
    Then the score should be exactly 0

    Examples:
      | a     | b     |
      | hello |       |
      |       | hello |

  @token @token-jaccard
  Scenario: score is symmetric across argument order
    # Set construction is order-independent (map[string]struct{}); the
    # integer-counter intersection cardinality is invariant under
    # argument swap; the single division produces identical float64
    # output regardless of argument order. Bit-for-bit equality —
    # no float tolerance needed.
    When I compute the TokenJaccard score between "a b c" and "b c d"
    And I compute the second TokenJaccard score between "b c d" and "a b c"
    Then both TokenJaccard scores should be equal

  @token @token-jaccard
  Scenario: set semantics deduplicate repeated tokens (distinct from Q-Gram Jaccard multiset)
    # The KEYSTONE set-vs-multiset distinction LOCKED in plan 06-04.
    # Tokenise("a a b") yields ["a","a","b"]; the set deduplication
    # collapses this to {a, b}. Tokenise("a b") yields ["a","b"]; set
    # {a, b}. Intersection {a, b}; union {a, b}; J = 2/2 = 1.0. A
    # multiset implementation would yield 2/3 ≈ 0.667 instead (counts
    # min(2,1)+min(1,1) = 2; sum 3+2-2 = 3). The set semantics are
    # intentional per RESEARCH.md Pattern 8: token presence is a binary
    # signal.
    When I compute the TokenJaccard score between "a a b" and "a b"
    Then the score should be exactly 1
