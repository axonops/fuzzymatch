# Tuning the Scorer

This document covers the calibration loop for the Phase 8 Scorer:
how to pick a threshold, how to allocate weights, and how to choose
the algorithm subset for a particular data domain.

The right configuration depends heavily on the consumer's data —
identifier matching for code/config taxonomies differs from
personal-name record linkage differs from product-catalogue fuzzy
search. The patterns documented here are heuristics, not laws; the
final calibration always comes from labelled data in the consumer's
own corpus.

See also:

- [`docs/scorer.md`](scorer.md) — the Scorer API reference.
- [`docs/algorithms.md`](algorithms.md) — per-algorithm intended-use
  notes.
- [`docs/faq.md`](faq.md) — common questions including "which
  algorithm for which data".
- [`docs/requirements.md`](requirements.md) §8 — the formal Scorer
  spec.

## How to pick a threshold

The threshold is the boundary at which `Scorer.Match` flips from
`false` to `true`. There is no universally safe default — the right
value depends on the false-positive / false-negative trade-off the
consumer is willing to accept.

The recommended calibration loop:

1. **Label a sample.** Pick a representative sample of pairs from
   your data. For each pair, decide whether you (the human) would
   consider it a match. Aim for 100-500 pairs split roughly 50/50
   between expected matches and expected non-matches; smaller corpora
   work too but the threshold estimate gets noisier.

2. **Score the sample.** Run `DefaultScorer().Score(a, b)` over every
   pair (or your chosen custom composition). Record the composite
   score and the human label.

3. **Inspect the distributions.** Plot the score histogram for the
   matches vs the non-matches (or print a sorted table). Most pairs
   will be unambiguous — matches clustered near `1.0` and
   non-matches near `0.0`. The interesting region is the middle band
   where the two distributions overlap.

4. **Pick the threshold.** For each candidate threshold (start at
   `0.85`, the `DefaultScorer` value), count the false positives
   (non-match labelled, score >= threshold) and false negatives
   (match labelled, score < threshold). Pick the threshold that
   optimises your domain-specific cost function:
   - **Precision-critical** (don't surface false matches at any
     cost): pick a higher threshold (`0.90`–`0.95`).
   - **Recall-critical** (don't miss any real match): pick a lower
     threshold (`0.70`–`0.80`).
   - **Balanced** (F1, accuracy): start at `DefaultScorer`'s `0.85`
     and adjust in `0.05` increments.

5. **Adjust in `0.05` increments.** Coarser increments (`0.10`) miss
   the inflection point; finer increments (`0.01`) overfit to your
   sample. `0.05` is the sweet spot for a 100-500 pair calibration.

6. **Cross-validate on held-out data.** If your sample is large
   enough, split it into a tuning set and a held-out validation set.
   Pick the threshold on the tuning set and confirm the chosen value
   produces similar false-positive / false-negative rates on the
   held-out set.

When Phase 9's `scan.Check` ships, this loop can run end-to-end over
a whole corpus without writing pair-iteration boilerplate. Until then,
the loop is a short Go program calling `Scorer.Score` in a `for` loop.

## How to pick weights

If `DefaultScorer` already works for your data, leave the weights
alone. The calibration cost of weight tuning is high, and the gain
over a balanced composition is usually small.

When custom weights are warranted (some algorithm is systematically
misleading on your data, or you have strong domain knowledge), the
recommended pattern:

1. **Start with equal weights via `DefaultScorerOptions`.** Build the
   custom Scorer on top of the default composition so the per-
   algorithm baseline is the same as `DefaultScorer`:

   ```go
   opts := append(fuzzymatch.DefaultScorerOptions(),
       fuzzymatch.WithThreshold(0.85), // re-state to override default
   )
   s, _ := fuzzymatch.NewScorer(opts...)
   ```

2. **Profile each algorithm via `ScoreAll`.** Run `s.ScoreAll(a, b)`
   over your labelled sample. For each algorithm, compute the
   correlation between its per-algorithm score and the human label.
   Algorithms that correlate well are good signals; algorithms that
   correlate poorly are noise (or worse, signal in the wrong
   direction).

3. **Adjust weights up for high-correlation algorithms, down for
   low-correlation.** A useful rule of thumb: weight ∝ correlation².
   Algorithms with negative correlation should be removed entirely
   via `WithoutAlgorithm`, not down-weighted.

4. **Re-run the threshold calibration.** Weight changes shift the
   composite-score distribution, so the threshold tuned for the
   default composition is unlikely to be optimal for the custom
   composition. Re-run "How to pick a threshold" above.

5. **Stop early.** Each additional tuning iteration provides
   diminishing returns. A two-iteration calibration (weights → re-
   threshold) usually captures most of the available improvement.

## DefaultScorer composition rationale

`DefaultScorer` combines six algorithms at equal raw weight with
threshold `0.85`:

| AlgoID                      | Why                                                                                                       |
| --------------------------- | --------------------------------------------------------------------------------------------------------- |
| `AlgoDamerauLevenshteinOSA` | Edit distance + adjacent transposition — covers typos and common keyboard errors in one algorithm.        |
| `AlgoJaroWinkler`           | Prefix bonus — strong signal for left-anchored matches (identifiers, names).                              |
| `AlgoTokenJaccard`          | Set similarity over tokens — handles snake_case / camelCase / kebab-case where token order shouldn't matter. |
| `AlgoQGramJaccard`          | Character trigram set similarity — language-agnostic per-character coverage that complements edit distance. |
| `AlgoSorensenDice`          | Same trigram space, different normalisation — paired with QGramJaccard for false-positive reduction.       |
| `AlgoDoubleMetaphone`       | Binary phonetic match — adds a signal for phonetically-equivalent names (Smith / Schmidt).                |

The composition is the result of empirical calibration on the
identifier-similarity test corpus. The threshold `0.85` is the value
where the false-positive and false-negative rates balance on the
labelled subset. For data domains substantially different from
identifier matching (e.g. street addresses, product names),
re-calibrate via the loop above.

## Choosing the right algorithm subset

As a starting heuristic — refine via the calibration loop:

- **Identifier matching (snake_case / camelCase / kebab-case names):**
  start with `DefaultScorer`. Consider removing `AlgoDoubleMetaphone`
  for numeric-heavy identifiers.
- **Free-form names with prefix significance (person names, request
  IDs):** keep `AlgoJaroWinkler` (already in default) and consider
  adding `AlgoStrcmp95`.
- **Pronunciation-equivalent matching (English-language names):**
  `AlgoDoubleMetaphone` or `AlgoNYSIIS` as a binary signal in the
  composite.
- **Substring-of containment (one name "contains" another):** add
  `AlgoSmithWatermanGotoh` or `AlgoPartialRatio`.
- **Token-reordering tolerance (multi-word phrases):** add
  `AlgoTokenSortRatio` or `AlgoTokenSetRatio`.
- **Long-string fuzzy match (descriptions, documents):** raise the
  q-gram size with `WithQGramJaccardAlgorithm(weight, 4)` or higher;
  trigrams (the default n=3) under-discriminate on long inputs.

## Performance / accuracy trade-off

Adding more algorithms to the composition increases accuracy but also
per-comparison cost. Each algorithm has its own complexity profile;
see [`docs/performance.md`](performance.md) for per-algorithm cost
numbers and [`docs/requirements.md`](requirements.md) §14 for the
per-algorithm allocation budgets.

A rough rule of thumb: `DefaultScorer.Score` on a 30-character ASCII
pair runs in under 30 µs on a modern Apple Silicon laptop. Doubling
the algorithm count roughly doubles the wall-time cost. Removing
algorithms cuts the cost proportionally.

For high-throughput callers (millions of pairs per second), profile
with `go test -bench` over a representative sample and consider:

1. Removing algorithms that contribute little signal in your data.
2. Pre-normalising upstream and passing `WithoutNormalisation` to skip
   the per-call Normalise.
3. Using the one-to-many `Extract` API (Phase 10) instead of
   pair-iteration for nearest-neighbour search.

## Pinning a calibrated configuration

Once a configuration is calibrated, pin it as a constant in the
consumer codebase:

```go
// CalibratedScorer is the production Scorer for identifier matching.
// Calibrated on the labelled identifier corpus at commit abc1234;
// threshold 0.82 balances precision and recall on the held-out set.
var CalibratedScorer = func() *fuzzymatch.Scorer {
    s, err := fuzzymatch.NewScorer(
        fuzzymatch.WithAlgorithm(fuzzymatch.AlgoDamerauLevenshteinOSA, 0.30),
        fuzzymatch.WithAlgorithm(fuzzymatch.AlgoJaroWinkler, 0.25),
        fuzzymatch.WithAlgorithm(fuzzymatch.AlgoTokenJaccard, 0.20),
        fuzzymatch.WithAlgorithm(fuzzymatch.AlgoQGramJaccard, 0.15),
        fuzzymatch.WithAlgorithm(fuzzymatch.AlgoSorensenDice, 0.10),
        fuzzymatch.WithThreshold(0.82),
    )
    if err != nil {
        panic("CalibratedScorer construction failed: " + err.Error())
    }
    return s
}()
```

The `*Scorer` returned by `NewScorer` is immutable and safe for
concurrent use, so a single package-level `var` is the canonical
production pattern.
