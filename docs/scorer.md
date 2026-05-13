# Scorer (Phase 8)

The `Scorer` is the second layer of the three-layer architecture: it
composes any subset of the 23 catalogue algorithms into a weighted
composite score with caller-controlled threshold and normalisation.

This document is a scaffold. The Scorer lands in Phase 8 of the roadmap.
Until it ships, see [`docs/requirements.md`](requirements.md) §8 for the
authoritative spec — that section pins construction shape, default
algorithm set, weight semantics, threshold behaviour, the `Score` /
`Match` / `ScoreAll` method signatures, and edge-case contracts.

> Phase 1 ships the Phase-8 prerequisites: `AlgoID` enum + dispatch table
> (algoid.go), sentinel errors (errors.go), and the `Normalise` and
> `Tokenise` primitives the Scorer applies before each algorithm
> invocation.

## Construction

TBD. See `docs/requirements.md` §8.1 for the spec.

## Defaults

TBD. See `docs/requirements.md` §8.5 for the default-Scorer composition.

## Composition

TBD. See `docs/requirements.md` §8.2 for the option set.

## Threshold

TBD. See `docs/requirements.md` §8.3 for `Match` behaviour and threshold
semantics.

## ScoreAll

TBD. See `docs/requirements.md` §8.6 for per-algorithm-score retrieval
and the deterministic-map-contents / non-deterministic-iteration
contract.

## Match

TBD. See `docs/requirements.md` §8.3 for the boolean match contract.

## Thread Safety

TBD. The Scorer will be **immutable after construction** and safe for
concurrent use by construction. Callers wanting a different
configuration construct a fresh `Scorer`. See `docs/requirements.md` §8
and the project Thread safety note in [`README.md`](../README.md).
