# Extending fuzzymatch

This document covers patterns for extending fuzzymatch — building
domain-specific Scorers, composing phonetic algorithms with edit
distance, customising the inner metric for Monge-Elkan, and (in the
rare cases where the curated catalogue is insufficient) writing a
custom algorithm that integrates cleanly with the public API.

This document is a scaffold. The patterns documented here mature as
each phase lands; Phase 8 (Scorer) and Phase 6 (token-based, including
Monge-Elkan's parametric inner-metric dispatch) are the primary
contributors.

See also:

- [`docs/scorer.md`](scorer.md) — Scorer composition (Phase 8).
- [`docs/tuning.md`](tuning.md) — threshold calibration.
- [`docs/algorithms.md`](algorithms.md) — per-algorithm reference.
- [`docs/requirements.md`](requirements.md) §6, §8 — public API and
  Scorer spec.
- [`CONTRIBUTING.md`](../CONTRIBUTING.md) — algorithm-proposal flow for
  consumers wanting a new algorithm in the upstream catalogue.

## Composing domain-specific Scorers

TBD. The Scorer ships in Phase 8. Until then, see `docs/requirements.md`
§8.2 for the option-set spec.

## Composing phonetic algorithms with edit distance

TBD. Phonetic algorithms (Soundex, Double Metaphone, NYSIIS, MRA)
expose both their underlying encoder (e.g. `SoundexCode`) and a binary
score (`SoundexScore` — 1.0/0.0). Consumers wanting continuous
similarity over phonetic codes apply Levenshtein (or any other
character-based metric) to the encoder output. See
`docs/requirements.md` §11.

## Custom inner metric for Monge-Elkan

TBD. Monge-Elkan accepts an `AlgoID` parameter at call time, so any
catalogue algorithm with a [0.0, 1.0] score can be the inner metric.
See `docs/requirements.md` §7.3.1.

## Custom algorithms outside the catalogue

The curated 23-algorithm catalogue is the public v1.x contract; new
algorithms enter the catalogue through the algorithm-proposal flow
documented in [`CONTRIBUTING.md`](../CONTRIBUTING.md). For consumers
whose use case is outside the spec, the recommended pattern is to
write the algorithm in the consumer's own package and use the Scorer's
weight set to combine it with the catalogue. The library does NOT
expose a plug-in interface for runtime algorithm registration — the
typed `AlgoID` enum and the array-backed dispatch table are
intentional design choices for zero-allocation hot-path access (see
`docs/requirements.md` §6 and the `go-coding-standards` skill).

## Determinism for custom algorithms

Any custom algorithm composed into a Scorer (via consumer code, since
no plug-in API exists) MUST satisfy the same determinism contract as
the library's algorithms:

- No map iteration on output paths.
- No transcendental float operations on output paths (DET-06).
- No `init()`-time table builds.
- Byte-identical output across the five supported platforms.

See `docs/requirements.md` §11 and `.claude/skills/determinism-standards/`
for the full discipline.
