# Extending fuzzymatch

This document covers patterns for extending fuzzymatch — building
domain-specific Scorers, composing phonetic algorithms with edit
distance, customising the inner metric for Monge-Elkan, and (in the
rare cases where the curated catalogue is insufficient) writing a
custom algorithm that integrates cleanly with the public API.

See also:

- [`docs/scorer.md`](scorer.md) — Scorer composition.
- [`docs/tuning.md`](tuning.md) — threshold calibration.
- [`docs/algorithms.md`](algorithms.md) — per-algorithm reference.
- [`docs/best-practices.md`](best-practices.md) — production patterns
  including Validate-then-Score.
- [`docs/requirements.md`](requirements.md) §6, §8 — public API and
  Scorer spec.
- [`CONTRIBUTING.md`](../CONTRIBUTING.md) — algorithm-proposal flow for
  consumers wanting a new algorithm in the upstream catalogue.

## Composing domain-specific Scorers

The canonical pattern is to start from `DefaultScorerOptions()` (the
balanced six-algorithm baseline) and layer domain-specific tweaks on
top via `append`:

```go
opts := append(fuzzymatch.DefaultScorerOptions(),
    // Drop phonetic — useless for numeric identifiers.
    fuzzymatch.WithoutAlgorithm(fuzzymatch.AlgoDoubleMetaphone),
    // Lower the threshold to compensate for the removed signal.
    fuzzymatch.WithThreshold(0.80),
)
s, err := fuzzymatch.NewScorer(opts...)
if err != nil {
    return fmt.Errorf("build scorer: %w", err)
}
```

`DefaultScorerOptions()` returns a fresh slice on every call, so layering
is safe and does not mutate the default for other consumers.

For a fully custom composition (no defaults), use `NewScorer` directly
with `WithAlgorithm` / `WithQGramJaccardAlgorithm` / `WithMongeElkanAlgorithm`
/ etc. plus `WithThreshold`. See [`docs/scorer.md`](scorer.md) for the
full option surface and [`docs/tuning.md`](tuning.md) for the
calibration loop.

## Composing phonetic algorithms with edit distance

Phonetic algorithms (Soundex, Double Metaphone, NYSIIS, MRA) expose
both their underlying encoder (e.g. `SoundexCode`) and a binary score
(`SoundexScore` — 1.0/0.0). For continuous similarity over phonetic
codes — useful when "Smith" and "Smithe" should produce a non-binary
similarity — apply Levenshtein (or any other character-based metric)
to the encoder output in consumer code:

```go
// Continuous Soundex similarity via Levenshtein over the code string.
codeA := fuzzymatch.SoundexCode("Smith")    // "S530"
codeB := fuzzymatch.SoundexCode("Smithe")   // "S530"
score := fuzzymatch.LevenshteinScore(codeA, codeB) // 1.0 (codes match)

// Continuous Double Metaphone similarity via Jaro-Winkler over the
// primary keys.
pA, _ := fuzzymatch.DoubleMetaphoneKeys("Schmidt") // "XMT", "SMT"
pB, _ := fuzzymatch.DoubleMetaphoneKeys("Smith")   // "SM0", "XMT"
score := fuzzymatch.JaroWinklerScore(pA, pB)       // partial
```

The binary `*Score` surface is the default because the canonical
phonetic algorithms are not designed for graded similarity — the
continuous form is a consumer-side composition with whatever inner
metric the consumer chooses. See [`docs/faq.md`](faq.md#why-phonetic-as-binary-in-the-scorer) for the rationale.

## Custom inner metric for Monge-Elkan

Monge-Elkan accepts an `AlgoID` parameter at call time, so any of the
18 permitted catalogue algorithms with a `[0.0, 1.0]` score can be the
inner metric:

```go
// Use Levenshtein as the inner metric (precise edit-distance scoring).
score := fuzzymatch.MongeElkanScore("alpha beta", "alpha beta gamma",
    fuzzymatch.AlgoLevenshtein)

// Use Jaro-Winkler as the inner metric (the dispatch-table default).
score := fuzzymatch.MongeElkanScore("alpha beta", "alpha beta gamma",
    fuzzymatch.AlgoJaroWinkler)
```

The permitted inner allow-list covers the 9 character-tier algorithms,
the 4 q-gram-tier algorithms, `AlgoRatcliffObershelp`, and the 4
phonetic algorithms (18 entries total). Passing `AlgoMongeElkan`
(self-recursion), any token-tier `AlgoID`, or any AlgoID outside the
allow-list panics with `ErrInvalidInnerAlgo` (recover-discriminable).
See `docs/algorithms.md#monge-elkan` for the full list.

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

See `docs/requirements.md` §13 (Determinism Guarantees) and
`.claude/skills/determinism-standards/` for the full discipline.
