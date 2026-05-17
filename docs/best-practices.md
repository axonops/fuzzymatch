# Best Practices

This document covers production patterns for fuzzymatch consumers:
when to use which surface, how to structure the input-quality
audit, how to pin a calibrated configuration, and how to compose the
library safely under concurrency.

See also:

- [`docs/scorer.md`](scorer.md) — Scorer API reference.
- [`docs/tuning.md`](tuning.md) — threshold and weight calibration.
- [`docs/algorithms.md`](algorithms.md) — per-algorithm reference.
- [`docs/extending.md`](extending.md) — composing custom scorers.
- [`docs/faq.md`](faq.md) — frequently asked questions.

## Input validation with Validate

`fuzzymatch.Validate(a, b string) []Warning` is the consumer-facing
input-quality diagnostic surface. It inspects the two input strings
and returns a deterministically-sorted slice of `Warning` values
describing problematic-but-non-fatal input shapes.

### When to use Validate

Call `Validate` whenever the inputs are untrusted or low-quality:

- **User-supplied identifiers** that may be empty, partial, or
  malformed.
- **Schema field names** authored by multiple teams where the input
  shape varies.
- **External data feeds** where input quality is not enforced
  upstream.
- **Audit-style telemetry** where consumers want to log the reason a
  score is low (empty input vs unequal length vs no tokens after
  normalisation).

Calling `Validate` on every pair before every `Score` call is wasteful
— the algorithms themselves accept any input and produce a sensible
value (the lenient comparison-data contract from
[`docs/requirements.md`](requirements.md) §6.A). `Validate` is the
optional companion that says whether the value will be meaningful.

### When NOT to use Validate

- **High-throughput hot paths** where every microsecond matters and
  the inputs are known-good upstream. The algorithms are correct
  without `Validate`; the diagnostic is purely additive.
- **Inside tight loops** where the warnings would be discarded
  unread. If you are not going to act on the warnings, do not call
  `Validate`.

### What Validate does NOT do

`Validate` is intentionally narrow:

- **Does NOT modify the inputs.** It is a pure read-only inspection.
- **Does NOT return an error.** It returns `[]Warning` (or `nil`);
  there is no error channel.
- **Does NOT panic.** Even on extreme inputs (gigabyte strings,
  invalid UTF-8, empty bytes); the function is defensive.
- **Does NOT block the score.** `Validate` is independent of `Score`;
  calling one does not affect the other. They can be invoked in
  either order, or only one, or neither.
- **Does NOT mutate global state.** Pure function; safe for
  concurrent use from any number of goroutines.

### The Validate-then-Score idiom

The recommended pattern for code paths that audit input quality:

```go
warnings := fuzzymatch.Validate(a, b)
for _, w := range warnings {
    log.Printf("input-quality warning: %s (%s): %s",
        w.Kind, w.Algorithm, w.Detail)
}
score := fuzzymatch.DefaultScorer().Score(a, b)
```

The warnings can be:

- **Logged** for post-hoc analysis (the canonical case).
- **Aggregated** as telemetry counters keyed by `Kind` and
  `Algorithm`.
- **Filtered** to a specific algorithm of interest (e.g. a Hamming-
  heavy workload pays close attention to `WarnUnequalLength` scoped
  to `AlgoHamming`).
- **Used to early-exit** before scoring, when the warnings indicate
  the score will be useless (e.g. both inputs all-non-ASCII for a
  phonetic-heavy Scorer).

### Per-WarnKind semantics

| `WarnKind`                     | Scope (`Algorithm` field)       | Triggered when                                                                                  |
|--------------------------------|---------------------------------|--------------------------------------------------------------------------------------------------|
| `WarnEmptyInput`               | `AlgoIDAny` (cross-cutting)     | `a == ""` or `b == ""`. Every algorithm sees a degenerate case.                                  |
| `WarnUnequalLength`            | `AlgoHamming` (per-algorithm)   | `len(a) != len(b)`. Hamming silent-max policy applies — score is `0.0`.                          |
| `WarnNoTokensAfterNormalise`   | 5 token-tier algorithms         | `Tokenise(s, DefaultTokeniseOptions())` returns an empty slice for at least one input.           |
| `WarnAllNonASCIIDropped`       | 5 ASCII-only algorithms         | One input contains characters but every rune is non-ASCII. ASCII-only encoders drop every rune. |
| `WarnPathologicallyLargeInput` | `AlgoIDAny` (cross-cutting)     | `max(len(a), len(b)) > 65536`. Quadratic algorithms allocate proportionally.                     |

The full reference lives in
[`docs/algorithms.md#input-validation-with-fuzzymatchvalidate`](algorithms.md#input-validation-with-fuzzymatchvalidate).

### CamelCase output

`WarnKind.String()` returns CamelCase labels matching the constant
suffix:

- `WarnEmptyInput.String()` returns `"EmptyInput"`.
- `WarnUnequalLength.String()` returns `"UnequalLength"`.
- `WarnNoTokensAfterNormalise.String()` returns `"NoTokensAfterNormalise"`.
- `WarnAllNonASCIIDropped.String()` returns `"AllNonASCIIDropped"`.
- `WarnPathologicallyLargeInput.String()` returns `"PathologicallyLargeInput"`.

This matches the `AlgoID.String()` convention (`AlgoLevenshtein` →
`"Levenshtein"`, `AlgoJaroWinkler` → `"JaroWinkler"`, etc.). Consumers
SHOULD treat the string forms as stable across patch releases; the
labels are pinned by `WarnKind_StringConvention_test.go`.

### Determinism

Two successive `Validate(a, b)` calls with identical `(a, b)` inputs
return byte-identical slices. The sort key is `(Algorithm, Kind)`
applied via `sort.SliceStable`, so insertion order within a single
`(Algorithm, Kind)` bucket is preserved. The determinism contract is
verified by `TestValidate_DeterministicOrdering` and the BDD scenario
"Two calls return identical warnings".

## Pinning a calibrated configuration

Once a Scorer is calibrated against a domain corpus, pin it as a
package-level `var` in the consumer codebase:

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
production pattern. Document the calibration provenance (commit,
sample size, held-out validation results) in the comment block above
the `var`.

## Concurrency

All exported functions in the root package are pure: no goroutines,
no channels, no mutexes, no shared mutable state.

- **Algorithm functions** (`LevenshteinScore`, `JaroWinklerScore`,
  etc.) are safe to call concurrently from any number of goroutines.
- **`Normalise` and `Tokenise`** are safe to call concurrently.
- **`Validate`** is safe to call concurrently.
- **`*Scorer`** is immutable after `NewScorer` / `DefaultScorer`
  returns. `Score`, `Match`, `ScoreAll`, `Threshold`, and
  `Algorithms` are all safe to call concurrently.

The library does not internally spawn goroutines. Consumers wanting
parallel scoring (e.g. comparing one input against many candidates
in parallel) construct a single `*Scorer` at start-up and share it
across worker goroutines.

## Error handling

Public functions distinguish two classes of "bad input" with
different policies (documented in
[`docs/requirements.md`](requirements.md) §6.A):

- **Comparison-data inputs** (the strings being compared) are
  handled leniently. Algorithm functions never panic, never return
  errors, and always produce a sensible value. `Validate` is the
  optional companion that reports problematic shapes as warnings.
- **Parameter inputs** (Scorer options, algorithm parameters like
  q-gram `n`, Tversky `α`/`β`, threshold values) are handled
  strictly. Invalid parameters return a typed error sentinel
  (`ErrInvalidThreshold`, `ErrInvalidQGramSize`, etc.) from option
  functions and panic with a wrapped sentinel from direct algorithm
  calls.

Discriminate parameter errors with `errors.Is(err, fuzzymatch.ErrX)`;
never match the error message string. The full sentinel surface lives
in `errors.go` and is documented per-symbol in
[`llms-full.txt`](../llms-full.txt) and `pkg.go.dev`.

## Choosing between layers

Three layers; pick the deepest that matches the question:

| Question                                            | Layer | Surface                                    |
|-----------------------------------------------------|-------|--------------------------------------------|
| "How similar are these two strings?"                | 1     | `LevenshteinScore(a, b)` etc.              |
| "How similar are these two strings overall?"        | 2     | `DefaultScorer().Score(a, b)`              |
| "Which pairs in this collection are similar?"      | 3     | `scan.Check(items, cfg)`                   |

For one-off pair comparisons in tight loops, prefer the algorithm
function directly — no composition overhead, no normalisation cost
unless explicitly invoked. For audit-quality scoring on
heterogeneous inputs, prefer the Scorer. For deduplication passes
over an entire corpus, use `scan.Check` which handles iteration,
suppression, and grouping internally.
