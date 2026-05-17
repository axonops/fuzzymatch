# Scorer

The `Scorer` is the second layer of fuzzymatch's three-layer
architecture. It composes any subset of the 23 catalogue algorithms
into a single weighted similarity score in `[0.0, 1.0]`, with a
configurable threshold for the boolean `Match` shortcut.

A `Scorer` is **immutable after construction** and **safe for
concurrent use** without external locks. Callers wanting a different
configuration build a fresh `Scorer`. There are no global flags, no
mutable internal state, and no init-time work.

The authoritative formal specification lives in
[`docs/requirements.md`](requirements.md) §8. This document is the
consumer-facing guide.

## Quickstart

The opinionated default Scorer is one line:

```go
package main

import "github.com/axonops/fuzzymatch"

func main() {
    s := fuzzymatch.DefaultScorer()
    if s.Match("user_id", "userId") {
        // similar
    }
}
```

`DefaultScorer()` cannot fail. It composes six algorithms at equal
weight with a baked-in threshold of `0.85`. See
[Default Composition](#default-composition) below.

## Custom Composition

For consumer-tuned weight allocations and threshold values, use
`NewScorer`:

```go
s, err := fuzzymatch.NewScorer(
    fuzzymatch.WithAlgorithm(fuzzymatch.AlgoLevenshtein, 0.6),
    fuzzymatch.WithAlgorithm(fuzzymatch.AlgoJaroWinkler, 0.4),
    fuzzymatch.WithThreshold(0.75), // REQUIRED — NewScorer returns ErrMissingThreshold otherwise
)
if err != nil {
    return fmt.Errorf("build scorer: %w", err)
}
```

`WithThreshold` is **mandatory** for `NewScorer`. The threshold is a
calibration parameter with no universally safe default, so the library
refuses to guess. See [Threshold](#threshold).

### Default minus one algorithm

The `WithoutAlgorithm` option composes with `DefaultScorerOptions()` to
produce a tuned variant — useful when the default composition includes
one algorithm that misleads on the consumer's data. The canonical
example is removing `AlgoDoubleMetaphone` for numeric-identifier
workloads where phonetic similarity is irrelevant:

```go
opts := append(fuzzymatch.DefaultScorerOptions(),
    fuzzymatch.WithoutAlgorithm(fuzzymatch.AlgoDoubleMetaphone),
    fuzzymatch.WithThreshold(0.80), // lower because one signal removed
)
s, err := fuzzymatch.NewScorer(opts...)
```

`WithoutAlgorithm` silently no-ops if the supplied `AlgoID` is not
already in the option slice, so it is safe to layer on top of any
composition without prior knowledge of what is present.

### Parameterised algorithm options

Most algorithms register via `WithAlgorithm(AlgoID, weight)` and pick
up their dispatch-default parameters (n=3 for q-gram algorithms, α=β=1
for Tversky, JaroWinkler inner for Monge-Elkan, default `SWGParams`
for Smith-Waterman-Gotoh). For consumer-controlled parameters, use the
dedicated With\*Algorithm options:

```go
s, err := fuzzymatch.NewScorer(
    fuzzymatch.WithQGramJaccardAlgorithm(0.4, 4), // n=4 instead of default 3
    fuzzymatch.WithTverskyAlgorithm(0.3, 0.5, 0.5, 3), // α=β=0.5 (Dice-like)
    fuzzymatch.WithMongeElkanAlgorithm(0.3, fuzzymatch.AlgoLevenshtein), // Lev inner
    fuzzymatch.WithThreshold(0.70),
)
```

## Method Reference

| Method                              | Returns                                     | Description                                                                                              |
| ----------------------------------- | ------------------------------------------- | -------------------------------------------------------------------------------------------------------- |
| `Score(a, b string)`                | `float64`                                   | Composite weighted similarity in `[0.0, 1.0]` (when weights are auto-normalised — the default).          |
| `Match(a, b string)`                | `bool`                                      | True when `Score(a, b) >= Threshold()` (boundary inclusive).                                             |
| `ScoreAll(a, b string)`             | `map[AlgoID]float64`                        | Fresh per-algorithm breakdown keyed by the typed `AlgoID` enum (see [ScoreAll](#scoreall-method) below). |
| `Threshold()`                       | `float64`                                   | The threshold configured at construction time.                                                           |
| `Algorithms()`                      | `[]ScorerAlgorithm`                         | Fresh slice of `{ID, Weight}` entries in `AlgoID`-ascending order (see [Algorithms](#algorithms) below). |

All four methods are pure functions on the `*Scorer` receiver. They
allocate no shared state, perform no goroutine work, and are safe for
concurrent use without locks.

## Threshold

`WithThreshold(t float64)` is the only mandatory option for
`NewScorer`. `t` must lie in `[0.0, 1.0]`; values outside this range
return `ErrInvalidThreshold`.

The library refuses a default threshold for `NewScorer` because no
value is universally safe:

- `1.0` (exact match only) silently produces "no matches found" for
  consumers who forget to set it.
- `0.0` silently makes every comparison a match.
- Inheriting `DefaultScorer`'s `0.85` is arbitrary for non-default
  compositions — `0.85` is calibrated for the SPECIFIC 6-algorithm
  mix, not for a Levenshtein-only Scorer.

Requiring `WithThreshold` forces an explicit calibration step at
construction time. For guidance on picking a value, see
[`docs/tuning.md`](tuning.md) "How to pick a threshold".

`DefaultScorer()` bakes `0.85` in, so casual consumers using the
default are unaffected.

## Default Composition

`DefaultScorer()` composes the following algorithms at equal raw
weight (auto-normalised to `1/6 ≈ 0.1667` each) with `WithThreshold(0.85)`:

| AlgoID                          | Category                       | Why included                                                                                                  |
| ------------------------------- | ------------------------------ | ------------------------------------------------------------------------------------------------------------- |
| `AlgoDamerauLevenshteinOSA`     | Edit distance + transposition  | Catches typo-style mutations and adjacent transpositions in a single algorithm.                               |
| `AlgoJaroWinkler`               | Name-matching with prefix bonus | Strong signal for left-anchored matches (typical of identifiers and proper names).                            |
| `AlgoTokenJaccard`              | Token set                      | Handles tokenisable input (snake_case / camelCase) where token reordering should not affect similarity.       |
| `AlgoQGramJaccard`              | Character n-gram               | Per-character coverage that complements edit distance on longer strings (default n=3).                        |
| `AlgoSorensenDice`              | Character n-gram (set sim)     | Similar profile to QGramJaccard but with a different normalisation; the two together reduce false positives.  |
| `AlgoDoubleMetaphone`           | Phonetic                       | Binary phonetic match — adds a signal for phonetically-equivalent inputs (Smith / Schmidt).                   |

The `0.85` threshold is calibrated for this specific mix; it is the
empirically-derived boundary where the false-positive and false-
negative rates are balanced on the identifier-similarity test corpus.
Consumers can override either the algorithm set or the threshold by
appending options:

```go
s, _ := fuzzymatch.NewScorer(append(
    fuzzymatch.DefaultScorerOptions(),
    fuzzymatch.WithThreshold(0.90), // tighter — fewer matches, higher precision
)...)
```

## Normalisation Control

By default, the Scorer applies `Normalise(s, DefaultNormalisationOptions())`
to both inputs **once** at the Scorer boundary before dispatching to
each algorithm. The pre-normalised strings are passed to every
algorithm — character-based AND token-based. Token-based algorithms
still tokenise the pre-normalised input via their internal `Tokenise`
call (their behaviour is unchanged).

To opt out of pre-normalisation:

```go
s, err := fuzzymatch.NewScorer(append(
    fuzzymatch.DefaultScorerOptions(),
    fuzzymatch.WithoutNormalisation(),
)...)
```

With `WithoutNormalisation`, the raw input bytes are passed to every
algorithm. This is the right choice when the consumer has already
canonicalised the inputs upstream or when raw-byte differences are
significant (e.g. comparing serialised UUIDs).

To customise the normalisation pipeline (different diacritic stripping,
different case folding, etc.):

```go
opts := fuzzymatch.DefaultNormalisationOptions()
opts.StripDiacritics = false // example: keep accents in
s, err := fuzzymatch.NewScorer(append(
    fuzzymatch.DefaultScorerOptions(),
    fuzzymatch.WithNormalisation(opts),
)...)
```

See [`docs/requirements.md`](requirements.md) §6 for the full
`NormalisationOptions` field set.

### Note on Monge-Elkan's `opts` parameter

`MongeElkanScore(a, b, inner, opts)` accepts a
`NormalisationOptions` parameter that is currently a **no-op**
(`_ = opts` inside the function body). This is a vestigial parameter
preserved for API stability — when invoked through the Scorer
(`WithMongeElkanAlgorithm` or via `WithAlgorithm(AlgoMongeElkan, w)`),
the Scorer's own Normalisation pipeline runs first and the parameter
is unused. A future major release may either remove the parameter or
wire it through; this is tracked as a deferred decision in the Phase 8
context.

## ScoreAll Method

`ScoreAll(a, b string)` returns a freshly-allocated
`map[AlgoID]float64` containing the per-algorithm score for every
algorithm configured on the Scorer. Useful for tuning and for the
calibration loop documented in [`docs/tuning.md`](tuning.md).

Note: `docs/requirements.md` §8.3 originally specified
`map[string]float64`; the implementation uses `map[AlgoID]float64`
(typed enum keys) for compile-time type safety. This is a deliberate
SPEC OVERRIDE per the Phase 8 design discussion; the requirements
document has been amended to match.

Go map iteration order is non-deterministic. The map **contents** are
deterministic (the same inputs always produce the same key set with
the same values) but the **iteration order** is randomised. Consumers
that need stable ordering should extract the keys and sort them:

```go
scores := s.ScoreAll(a, b)
ids := make([]fuzzymatch.AlgoID, 0, len(scores))
for id := range scores {
    ids = append(ids, id)
}
sort.Slice(ids, func(i, j int) bool { return ids[i] < ids[j] })
for _, id := range ids {
    fmt.Printf("%s: %.4f\n", id, scores[id])
}
```

Internal computation iterates in `AlgoID`-ascending order regardless,
so the float-determinism guarantee (Phase 5 carry-forward — the
composite score is bitwise stable across platforms) holds independently
of how the consumer iterates the returned map.

## Algorithms

`Algorithms()` returns a fresh slice of `ScorerAlgorithm{ID, Weight}`
entries in `AlgoID`-ascending order. The slice is freshly allocated on
every call, so the caller may freely mutate, sort, or filter it
without affecting other callers.

The weights reflect the post-construction state: after auto-
normalisation (default), weights sum to `1.0`; under
`WithNormaliseWeights(false)`, weights are raw and may sum to any
positive value.

## Concurrency

A `*Scorer` is **immutable after construction**. All four methods
(`Score`, `Match`, `ScoreAll`, `Threshold`, `Algorithms`) are safe to
call concurrently from any number of goroutines. There is no internal
mutex, no atomic operations, and no goroutine creation.

The BDD scenario `Concurrent Score calls return identical results`
(in `tests/bdd/features/scorer.feature`) verifies the concurrency
guarantee end-to-end: 100 goroutines call `Score` on the same pair and
the results are byte-identical.

Consumers that want a different configuration construct a fresh
`Scorer`. There is no "modify in place" path.

## Errors

`NewScorer` returns one of four sentinel errors when its input is
malformed. All are typed `error` values exported from the package;
discriminate via `errors.Is`, never by matching the error message
string.

| Sentinel               | When it fires                                                                                                                                                                                                                                                       |
| ---------------------- | ------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| `ErrMissingThreshold`  | Returned when no `WithThreshold` option was supplied. The threshold is mandatory at construction time — see [Threshold](#threshold) above for the rationale.                                                                                                       |
| `ErrEmptyScorer`       | Returned when no `WithAlgorithm` (or any other With\*Algorithm) option was supplied. A Scorer with zero algorithms has no meaningful Score function.                                                                                                                |
| `ErrInvalidWeight`     | Returned when `WithAlgorithm` (or any parameterised variant) receives a non-positive weight. Weights must be `> 0`; negative or zero weights cannot meaningfully contribute to a composite.                                                                          |
| `ErrInvalidThreshold`  | Returned when `WithThreshold` receives a value outside `[0.0, 1.0]`, or a NaN. The threshold is a probability-like quantity and must lie in the unit interval.                                                                                                       |

The validation order is LOCKED: `ErrMissingThreshold` fires first so a
user who forgets `WithThreshold` AND has another option problem sees
the clear "you forgot the threshold" message rather than a cascading
sentinel from later validation.

## Weight Normalisation

By default, `NewScorer` auto-normalises the supplied weights to sum to
`1.0` (each weight divided by the sum of all weights). This guarantees
the composite `Score` lies in `[0.0, 1.0]`.

To preserve raw weights:

```go
s, _ := fuzzymatch.NewScorer(
    fuzzymatch.WithAlgorithm(fuzzymatch.AlgoLevenshtein, 1.0),
    fuzzymatch.WithAlgorithm(fuzzymatch.AlgoJaroWinkler, 3.0),
    fuzzymatch.WithThreshold(0.5),
    fuzzymatch.WithNormaliseWeights(false), // raw weights — composite may exceed 1.0
)
```

Under `WithNormaliseWeights(false)`, the composite `Score` is the raw
weighted sum and the `[0.0, 1.0]` guarantee is waived; consumers take
responsibility for the weight semantics. This is rarely the right
choice — `WithoutNormaliseWeights` exists for advanced use cases where
the consumer is implementing their own normalisation upstream.

## Last-write-wins for duplicate AlgoIDs

Two `WithAlgorithm(AlgoLevenshtein, w)` calls in the same option list
do NOT compose; only the latter weight survives. The dedup pass in
`NewScorer` iterates the option list in order, building a
`map[AlgoID]scorerEntry` where each assignment overwrites the previous
entry for the same `AlgoID`. This is the documented semantic — the
order of options matters, and a later option always supersedes an
earlier one for the same algorithm.

The behaviour generalises to the parameterised options: a later
`WithQGramJaccardAlgorithm(weight, n)` overrides an earlier
`WithAlgorithm(AlgoQGramJaccard, weight)` (or vice versa) — both
register the same `AlgoID`, and last-write-wins applies.

## See also

- [`docs/tuning.md`](tuning.md) — how to pick a threshold and weights
  for a consumer's specific data.
- [`docs/algorithms.md`](algorithms.md) — per-algorithm intended-use
  notes and primary-source citations.
- [`docs/requirements.md`](requirements.md) §8 — the authoritative
  spec.
- [`examples/identifier-similarity/`](../examples/identifier-similarity/) —
  runnable demo using the default Scorer alongside the 23 raw
  algorithms.
- [`examples/scorer-composition/`](../examples/scorer-composition/) —
  runnable demo of `DefaultScorerOptions() + WithoutAlgorithm + WithThreshold`.
