# Tuning fuzzymatch

This document covers threshold calibration, algorithm-subset selection,
and weight-set tuning against a domain corpus. The right algorithm
choice and weight allocation depends heavily on the consumer's data —
identifier matching for code/config taxonomies differs from
personal-name record linkage differs from product-catalogue fuzzy
search.

This document is a scaffold. The tuning patterns documented here mature
alongside Phase 8 (Scorer) and the integration shakedown in Phase 7.

See also:

- [`docs/scorer.md`](scorer.md) — Scorer composition and threshold API.
- [`docs/algorithms.md`](algorithms.md) — per-algorithm intended-use
  notes.
- [`docs/faq.md`](faq.md) — common questions including "which algorithm
  for which data".
- [`docs/requirements.md`](requirements.md) §8 — Scorer spec.

## Calibrating thresholds against a domain corpus

TBD. The recommended pattern is to:

1. Label a held-out sample of expected matches and expected non-matches
   from the consumer's data.
2. Run the Scorer over the labelled sample at varying thresholds.
3. Pick the threshold maximising the consumer's preferred metric
   (precision, recall, F1, or domain-specific cost function).

Worked example: pending Phase 8 + Phase 7 integration shakedown.

## Choosing the right algorithm subset

TBD. As a starting heuristic:

- **Identifier matching (snake_case / camelCase / kebab-case names):**
  start with `AlgoLevenshtein` + `AlgoTokenJaccard` + `AlgoJaroWinkler`.
- **Free-form names with prefix significance (person names, request
  IDs):** add `AlgoJaroWinkler` and `AlgoStrcmp95`.
- **Pronunciation-equivalent matching (English-language names):** add
  `AlgoDoubleMetaphone` or `AlgoSoundex` as a binary signal in the
  composite.
- **Substring-of containment (one name "contains" another):** add
  `AlgoSmithWatermanGotoh` or `AlgoPartialRatio`.
- **Token-reordering tolerance:** add `AlgoTokenSortRatio` or
  `AlgoTokenSetRatio`.

The default `Scorer` (Phase 8) ships with a curated set covering the
common cases; tuning is for consumers who want a different bias.

## Weight semantics

TBD. See `docs/requirements.md` §8.4 for weight normalisation rules.

## Performance / accuracy trade-off

TBD. Adding more algorithms increases accuracy but also per-comparison
cost. See `docs/performance.md` for per-algorithm cost numbers and
`docs/requirements.md` §14 for the per-algorithm allocation budgets.
