---
phase: 08-composite-scorer
reviewed: 2026-05-17T00:00:00Z
depth: standard
files_reviewed: 23
files_reviewed_list:
  - errors.go
  - errors_test.go
  - scorer_options.go
  - scorer_options_test.go
  - scorer_options_internal_test.go
  - scorer.go
  - scorer_test.go
  - scorer_internal_test.go
  - scorer_bench_test.go
  - scorer_golden_test.go
  - testdata/golden/scorer-default.json
  - examples/identifier-similarity/main.go
  - examples/identifier-similarity/main_test.go
  - examples/scorer-composition/main.go
  - examples/scorer-composition/main_test.go
  - examples/scorer-composition/go.mod
  - examples/scorer-composition/go.sum
  - tests/bdd/features/scorer.feature
  - tests/bdd/steps/scorer_steps.go
  - tests/bdd/steps/algorithms_steps.go
  - docs/scorer.md
  - docs/tuning.md
  - docs/requirements.md
findings:
  critical: 2
  warning: 5
  info: 5
  total: 12
status: issues_found
---

# Phase 8: Code Review Report

**Reviewed:** 2026-05-17
**Depth:** standard
**Files Reviewed:** 23
**Status:** issues_found

## Summary

Phase 8 introduces the composite weighted Scorer (Layer 2 of three). The
implementation is well-structured: production code is goroutine-free,
init-free, third-party-import-free, and uses the LOCKED determinism
discipline (AlgoID-sorted iteration, left-to-right reduction with
explicit parens). Concurrency safety is solid — the Scorer is constructed
once and the methods are pure reads of immutable fields. The
no-map-iteration-on-output discipline is preserved (`ScoreAll` writes
into a map but iterates a sorted slice to populate it; `Algorithms()`
returns a fresh slice in AlgoID order).

Two correctness defects sit at the option layer:

- **NaN slips through `WithThreshold`** (CR-01). `t < 0.0 || t > 1.0`
  is false for NaN, so `WithThreshold(math.NaN())` silently succeeds.
  `Match` then becomes `score >= NaN`, which is always false — silent
  no-match malfunction. The error-table docs in `docs/scorer.md` claim
  NaN is rejected (it isn't).
- **`WithTverskyAlgorithm` does not validate the `α + β > 0`
  constraint** (CR-02). The option lets `(0, 0)` pass and the panic
  fires later inside `TverskyScore` at Score-time. This is inconsistent
  with the rest of the With\*Algorithm validation contract (every
  other option validates fully up front and returns a typed sentinel)
  and breaks the documented "fail loudly at construction" pattern.

Beyond the two critical option-layer defects, there is a flaky property
test (WR-01 — `uint16` overflow produces zero weight and trips the
test's own "constructor failure is a bug" branch), a brittle hardcoded
BDD step (WR-02), and the algorithm-performance-reviewer follow-up
already flagged in 08-04-SUMMARY (WR-04 — allocations 12 short / 34
medium versus the ≤ 8 budget). The other items are godoc / comment
inaccuracies and minor coverage gaps.

## Critical Issues

### CR-01: `WithThreshold` does not reject NaN

**File:** `scorer_options.go:257-266`
**Issue:** `WithThreshold(t)` only checks `t < 0.0 || t > 1.0`. Both
comparisons evaluate to `false` for `t = math.NaN()`, so the option
accepts NaN, stores it in `cfg.threshold`, and sets `thresholdSet =
true`. `NewScorer` then freezes the Scorer with `threshold = NaN`.
`(*Scorer).Match` returns `s.Score(a, b) >= s.threshold` — and `x >=
NaN` is always `false`, so the Scorer silently never matches anything
regardless of input quality.

The documented contract in `docs/scorer.md:283` explicitly claims NaN
is rejected ("Returned when `WithThreshold` receives a value outside
`[0.0, 1.0]`, or a NaN."). The implementation does not match the
documentation, and the documented behaviour is the correct one — a NaN
threshold is a programmer error that should surface immediately.

**Fix:**
```go
import "math"

func WithThreshold(t float64) ScorerOption {
    return func(cfg *scorerConfig) error {
        if math.IsNaN(t) || t < 0.0 || t > 1.0 {
            return ErrInvalidThreshold
        }
        cfg.threshold = t
        cfg.thresholdSet = true
        return nil
    }
}
```

Add a unit test in `scorer_options_test.go` (`TestWithThreshold_RejectsNaN`)
asserting `errors.Is(err, ErrInvalidThreshold)` and that `thresholdSet`
stays false on rejection.

---

### CR-02: `WithTverskyAlgorithm` does not enforce α + β > 0; consumer call panics at Score time

**File:** `scorer_options.go:381-399`
**Issue:** `WithTverskyAlgorithm` validates only `alpha < 0 || beta < 0`
but allows the combination `alpha == 0 && beta == 0`. The Tversky
denominator is `intersection + α·|A\B| + β·|B\A|`; with both weights
zero this collapses to the bare intersection count which can be zero
(no shared q-grams) and is undefined as a similarity. `tversky.go:241`
explicitly panics on this case:

```go
if alpha < 0 || beta < 0 || (alpha == 0 && beta == 0) {
    panic("fuzzymatch: invalid tversky parameter")
}
```

So a consumer who calls:

```go
s, _ := fuzzymatch.NewScorer(
    fuzzymatch.WithTverskyAlgorithm(0.5, 0, 0, 3),
    fuzzymatch.WithThreshold(0.5),
)
s.Score("abc", "abc") // PANIC
```

succeeds at NewScorer time (no error) and panics on the first Score
call. This breaks the documented Phase 8 contract — every other
With\*Algorithm option fully validates at option-application time and
returns the appropriate `ErrInvalid*` sentinel. The godoc at
`scorer_options.go:376-380` actually acknowledges the gap and waves it
off ("this option does not re-check it because either α or β being >
0 is satisfied by the typical use cases"), but "typical use" is not a
defence — the option is the documented validation boundary, and a
panic at Score time defeats the "construction-time error" contract
that the rest of the Scorer surface upholds.

**Fix:**
```go
func WithTverskyAlgorithm(weight, alpha, beta float64, n int) ScorerOption {
    return func(cfg *scorerConfig) error {
        if weight <= 0 {
            return ErrInvalidWeight
        }
        if n < 1 {
            return ErrInvalidQGramSize
        }
        if alpha < 0 || beta < 0 || (alpha == 0 && beta == 0) {
            return ErrInvalidTverskyParam
        }
        cfg.entries = append(cfg.entries, scorerEntry{
            id:      AlgoTversky,
            weight:  weight,
            scoreFn: func(a, b string) float64 { return TverskyScore(a, b, n, alpha, beta) },
        })
        return nil
    }
}
```

Add a unit test (`TestWithTverskyAlgorithm_RejectsBothZero`) asserting
`errors.Is(err, ErrInvalidTverskyParam)` for `alpha = 0, beta = 0`.

## Warnings

### WR-01: `TestProp_Scorer_WeightSumOne` random-vector branch is flaky on `uint16(65535)`

**File:** `scorer_test.go:820-853`
**Issue:** The random-vector wrapper inside `TestProp_Scorer_WeightSumOne`
intends to bias inputs into `(0, 100]` so weights never hit zero:

```go
toPositive := func(u uint16) float64 {
    // Avoid zero by adding 1 in the numerator; range becomes
    // (0, 100].
    return float64(u+1) / float64(uint32(1)<<16) * 100.0
}
```

The `u+1` expression is computed in `uint16` arithmetic (the constant
`1` is implicitly typed `uint16` because `u` is `uint16`). When `u =
65535`, `u + 1` overflows to `0` and `toPositive` returns `0.0`. The
zero-weight then trips `ErrInvalidWeight` at the option layer, and the
property function returns `false` ("constructor failure on positive
weights is itself a bug"), failing the property test.

For three independent uint16 inputs over 100 quick.Check iterations
the probability of at least one hitting 65535 is approximately
`1 - (65535/65536)^300 ≈ 0.46%` per test run — a flake-once-per-200
runs pattern that will surface intermittently in CI nightly fuzz runs
and confuse downstream debugging.

**Fix:** Map into `uint32` before the addition so the +1 cannot
overflow:

```go
toPositive := func(u uint16) float64 {
    return float64(uint32(u)+1) / float64(uint32(1)<<16) * 100.0
}
```

Or, more clearly, clamp into a half-open range that excludes zero:

```go
toPositive := func(u uint16) float64 {
    // Map [0, 65535] → (0.0, 100.0] — never zero by construction.
    return (float64(u) + 1.0) / 65536.0 * 100.0
}
```

---

### WR-02: BDD step `iScoreTheSamePairWithTheDefaultScorer` hardcodes its inputs

**File:** `tests/bdd/steps/scorer_steps.go:189-195`
**Issue:** The step hardcodes the input pair:

```go
func (sc *ScorerContext) iScoreTheSamePairWithTheDefaultScorer() error {
    if sc.defaultScorer == nil { ... }
    sc.lastScore = sc.defaultScorer.Score("XMLParser", "xml_parser")
    return nil
}
```

The Gherkin step is `^I score the same pair with the default Scorer$`
— a step whose phrasing implies the pair was captured in a preceding
step. The pair is not captured; the step's natural-language contract
("the same pair") is enforced only by convention in the one feature
scenario that currently uses it. Any future scenario that reuses this
step phrase with a different pair will silently score `"XMLParser"` /
`"xml_parser"` and produce nonsense.

**Fix:** Capture the input pair in `ScorerContext` when the preceding
"I score X and Y" step runs, then reuse those captured values:

```go
type ScorerContext struct {
    ...
    lastA, lastB string
}

func (sc *ScorerContext) iScoreAndWithTheScorer(a, b string) error {
    if sc.scorer == nil { ... }
    sc.lastA, sc.lastB = a, b
    sc.lastScore = sc.scorer.Score(a, b)
    sc.lastMatch = sc.scorer.Match(a, b)
    return nil
}

func (sc *ScorerContext) iScoreTheSamePairWithTheDefaultScorer() error {
    if sc.defaultScorer == nil { ... }
    if sc.lastA == "" && sc.lastB == "" {
        return fmt.Errorf("no pair captured; preceding 'I score X and Y' step must run first")
    }
    sc.lastScore = sc.defaultScorer.Score(sc.lastA, sc.lastB)
    return nil
}
```

---

### WR-03: `WithoutAlgorithm` godoc claims reverse iteration, code iterates forward

**File:** `scorer_options.go:188-194`
**Issue:** The implementation comment states:

```go
// Linear scan-and-compact. The option slice is typically small
// (default Scorer has 6 entries) so the O(n) cost is
// negligible. Iterate in reverse so removals do not shift
// later indices.
filtered := cfg.entries[:0]
for _, e := range cfg.entries {  // forward iteration
```

The code iterates forward, not in reverse. The forward `[:0]`-then-
append-kept-entries compaction is the canonical Go idiom and is
correct; the comment is stale and confuses readers debugging the
option layer. A future reader looking for a reverse-iteration bug will
chase a phantom.

**Fix:** Update the comment to match the implementation:

```go
// Linear scan-and-compact in place: alias filtered to the
// existing backing array via cfg.entries[:0], then append every
// non-matching entry. The kept-entries write index never
// outruns the read index, so the compaction is safe in-place
// (standard Go slice-filter idiom).
```

---

### WR-04: `DefaultScorer.Score` exceeds the ≤ 8 allocs/op budget on ASCII Short and Medium

**File:** `scorer_bench_test.go:71-98` + `.planning/phases/08-composite-scorer/08-04-SUMMARY.md:164-187`
**Issue:** This is the pre-flagged follow-up from plan 08-04 SUMMARY,
surfaced here per the orchestrator's request so it is tracked in the
review artifact:

- `BenchmarkDefaultScorer_Score_ASCII_Short` — 12 allocs/op (budget ≤ 8;
  50% over)
- `BenchmarkDefaultScorer_Score_ASCII_Medium` — 34 allocs/op (budget ≤ 8;
  4× over)

The wall-time budget (`< 30 µs`) is met on both inputs; only the
allocation budget is breached. Root causes per 08-04-SUMMARY:

1. Six per-call algorithm closures each allocate independently.
2. Map-keyed q-gram counting (QGram/Sørensen/TokenJaccard) compounds
   small allocations.
3. `Normalise(a, opts)` and `Normalise(b, opts)` allocate two new
   strings per call (DefaultNormalisationOptions has Lowercase +
   StripSeparators + SplitCamelCase + NFC).

**Fix (recommendation per 08-04-SUMMARY):** Defer to the
`algorithm-performance-reviewer` agent to design the allocation-
reduction strategy. Likely candidates:

- `sync.Pool` for the per-algorithm scratch maps.
- A `[]byte` reusable buffer pattern for `Normalise` on stack-friendly
  inputs (the ASCII fast path is already in place at the per-algorithm
  level; the Scorer-boundary `Normalise` call needs the same).
- Drop `applyNormalisation` ASCII-no-op fast path: when the input is
  pure-ASCII and `DefaultNormalisationOptions()` would not change the
  string, skip the `Normalise` allocation.

This is **not blocking** for plan 08-04 acceptance (the wall-time
budget passes and the allocation budget is documented as an
algorithm-performance-reviewer follow-up in the SUMMARY), but it
should be tracked as a v1.0 release blocker per the performance-
standards skill.

---

### WR-05: `WithoutNormalisation` godoc misleads on `normOpts` reuse

**File:** `scorer_options.go:232-240`
**Issue:** The godoc states:

> applyNorm becomes false but the previously-stored normOpts value is
> intentionally not cleared (cheap; harmless; allows a subsequent
> WithNormalisation to inspect-and-reuse if it wishes).

But `WithNormalisation` unconditionally **overwrites** `cfg.normOpts`:

```go
func WithNormalisation(opts NormalisationOptions) ScorerOption {
    return func(cfg *scorerConfig) error {
        cfg.applyNorm = true
        cfg.normOpts = opts   // OVERWRITES — no inspect-and-reuse path
        return nil
    }
}
```

No call site in the codebase inspects-and-reuses the previously stored
normOpts. The godoc claim is fiction. The actual reason `normOpts`
stays populated is that clearing it would require a sentinel-value
check elsewhere; leaving it is harmless because `applyNormalisation =
false` gates the read in `Score` / `ScoreAll`. The godoc should say
that, not invent an "inspect-and-reuse" pattern.

**Fix:**
```go
// WithoutNormalisation returns a ScorerOption that disables pre-
// comparison normalisation in the resulting Scorer. ... Passing this
// option AFTER WithNormalisation in the same option slice disables
// normalisation (later option wins); applyNorm becomes false. The
// previously-stored normOpts value is left untouched — it is unused
// while applyNorm is false (Score/ScoreAll gate the Normalise call on
// applyNormalisation), and clearing it would require a sentinel-value
// check at every read site for no behavioural gain.
```

## Info

### IN-01: `docs/scorer.md` method-count text contradicts its own table

**File:** `docs/scorer.md:104,259`
**Issue:** Line 105 says "All four methods are pure functions" but the
table above (lines 97-103) lists five methods (`Score`, `Match`,
`ScoreAll`, `Threshold`, `Algorithms`). Line 259 repeats the error
("All four methods (`Score`, `Match`, `ScoreAll`, `Threshold`,
`Algorithms`)" — names five inside a "four" sentence). Minor reader
confusion.

**Fix:** Change "four" to "five" in both locations, or rephrase as
"All methods on `*Scorer`".

---

### IN-02: `docs/scorer.md` error table claims `ErrInvalidThreshold` rejects NaN, but it doesn't

**File:** `docs/scorer.md:283`
**Issue:** The error-table row reads:

> `ErrInvalidThreshold` | Returned when `WithThreshold` receives a
> value outside `[0.0, 1.0]`, or a NaN.

The implementation does not reject NaN (see CR-01). The documentation
is aspirational; the fix for CR-01 will bring the code in line with
this documentation, at which point this Info finding is automatically
resolved.

**Fix:** Fix CR-01 (the code), not this doc — the documentation states
the intended behaviour correctly.

---

### IN-03: `WithMongeElkanAlgorithm` allow-list panic at Score time, not validation-time

**File:** `scorer_options.go:425-446`
**Issue:** `WithMongeElkanAlgorithm` validates `inner` against
`numAlgorithms` and the trivial-recursion case (`inner == AlgoMongeElkan`)
but does NOT validate against `MongeElkanScoreSymmetric`'s own
18-entry allow-list. An inner AlgoID outside that allow-list (e.g.
`AlgoSoundex` if Soundex is excluded by the Phase 6 ME contract) is
accepted at construction and panics at the first Score call. This
mirrors the inconsistency in CR-02 (TverskyScore panic semantics) but
the godoc is explicit about the deferral ("The full 18-entry inner
allow-list is enforced inside MongeElkanScoreSymmetric (Phase 6 +
Phase 7 locked behaviour)").

Promoting this to Warning would force a v1.x API decision (does the
Scorer mirror ME's allow-list, or does the option layer remain
deliberately thin?). Tracking as Info for now; revisit during the
api-ergonomics-reviewer's v1.0 surface freeze.

**Fix (deferred decision):** Either (a) introspect ME's allow-list
from the Scorer option layer and return `ErrInvalidAlgorithm` at
construction-time, or (b) accept the deferred-panic behaviour and add
a BDD scenario that pins the panic-at-Score-time contract via godog's
recover mechanism.

---

### IN-04: BDD coverage gap — no `ErrInvalidThreshold` scenario

**File:** `tests/bdd/features/scorer.feature:104-127`
**Issue:** The error-path scenarios cover `ErrMissingThreshold`,
`ErrEmptyScorer`, and `ErrInvalidWeight`, but not `ErrInvalidThreshold`.
The `constructingTheScorerShouldReturn` step (scorer_steps.go:339) has
a case branch for `ErrInvalidThreshold`, suggesting it was intended.
The step regex
`^I attempt to construct a Scorer with Levenshtein weight (-?\d+\.?\d*) and threshold (\d+\.?\d*)$`
does not allow negative thresholds (`(\d+\.?\d*)` has no `-?` prefix).
Add a scenario:

```gherkin
@scorer @errors
Scenario: WithThreshold with out-of-range value returns ErrInvalidThreshold
  When I attempt to construct a Scorer with Levenshtein weight 1.0 and threshold 1.5
  Then constructing the Scorer should return ErrInvalidThreshold
```

(And widen the regex to `(-?\d+\.?\d*)` for the threshold parameter to
support the negative-threshold sub-case.)

**Fix:** Add the scenario and a NaN scenario once CR-01 is resolved.

---

### IN-05: `scorer_bench_test.go` Match-benchmark sink gate is unreachable for compiler-elimination purposes

**File:** `scorer_bench_test.go:171-173`
**Issue:** The `var sink int` + `if s.Match(...) { sink++ }` loop has
a post-loop guard `if sink < -1`. Because `sink` is `int` and only
ever incremented from 0, `sink < -1` is provably false at compile time
— a sufficiently aggressive optimiser could elide the entire post-loop
branch (and, with profile-guided optimisation, infer the loop is
dead). The current Go compiler does not perform this elimination, so
the benchmark works in practice, but the locked-pattern docstring at
the file header claims this idiom defeats dead-code elimination — and
this particular instance is the weakest defence in the file.

**Fix:** Use the result inside the loop more directly:

```go
func BenchmarkDefaultScorer_Match_ASCII_Short(b *testing.B) {
    s := fuzzymatch.DefaultScorer()
    b.ReportAllocs()
    b.ResetTimer()
    var sink bool
    for i := 0; i < b.N; i++ {
        sink = s.Match("abc", "abcd")
    }
    if sink && !sink {
        b.Fatal("sink unexpectedly toggled — compiler folded the benchmark away")
    }
}
```

Or alternatively keep the int counter but compare against `b.N` after
the loop (the compiler cannot fold `b.N`).

---

_Reviewed: 2026-05-17_
_Reviewer: Claude (gsd-code-reviewer)_
_Depth: standard_
