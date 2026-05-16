# Phase 8: Composite Scorer — Pattern Map

**Mapped:** 2026-05-16
**Files analysed:** 17 (new or modified)
**Analogs found:** 17 / 17

---

## File Classification

| New / Modified File | Role | Data Flow | Closest Analog | Match Quality |
|---------------------|------|-----------|----------------|---------------|
| `scorer.go` | service / type | transform (composite reduction) | `cosine.go` (reduction loop) + `algoid.go` (dispatch access) | role-match + data-flow-match |
| `scorer_options.go` | config / builder | transform (option accumulation) | `dispatch_cosine.go` + `dispatch_monge_elkan.go` | role-match |
| `errors.go` (append) | utility | n/a | `errors.go:48-107` (existing sentinels) | exact |
| `scorer_test.go` | test | request-response | `cosine_test.go` (table-driven stdlib) | exact |
| `scorer_options_test.go` | test | request-response | `cosine_test.go` (table-driven stdlib) | exact |
| `scorer_internal_test.go` | test (package-internal) | request-response | no existing `_internal_test.go`; pattern from `normalise_test.go` (`package fuzzymatch`) | role-match |
| `scorer_golden_test.go` | test (golden gate) | transform | `algorithms_golden_test.go` + `golden_test.go` | exact |
| `scorer_bench_test.go` | test (benchmark) | transform | `cosine_bench_test.go` | exact |
| `tests/bdd/features/scorer.feature` | test (BDD) | event-driven | `cosine.feature` / `tversky.feature` | exact |
| `tests/bdd/steps/scorer_steps.go` | test (BDD steps) | event-driven | `tests/bdd/steps/algorithms_steps.go` | exact |
| `testdata/golden/scorer-default.json` | data | n/a | `testdata/golden/algorithms.json` (schema) | role-match |
| `examples/scorer-composition/main.go` | example | request-response | `examples/identifier-similarity/main.go` | exact |
| `examples/scorer-composition/main_test.go` | test (example stdout gate) | request-response | `examples/identifier-similarity/main_test.go` | exact |
| `examples/identifier-similarity/main.go` (extend) | example | request-response | itself (append pattern) | exact |
| `docs/scorer.md` (populate) | docs | n/a | `docs/algorithms.md` (scaffold → prose pattern) | role-match |
| `docs/tuning.md` (populate) | docs | n/a | `docs/algorithms.md` | role-match |
| `llms.txt` / `llms-full.txt` (extend) | docs | n/a | themselves (extend-in-lockstep pattern from Phase 5+) | exact |

---

## Pattern Assignments

---

### `scorer.go` (service/type, composite reduction)

**Analogs:**
- `cosine.go:303-344` — float-determinism reduction loop (explicit parens, sorted keys)
- `algoid.go:310-324` — dispatch array declaration and access pattern
- `algoid.go:282-308` — `AlgoIDs()` for canonical sort order

**File header pattern** (`cosine.go:1-14`):
```go
// Copyright 2026 AxonOps Limited
//
// Licensed under the Apache License, Version 2.0 (the "License");
// ...
```

**Imports pattern** — scorer.go will use only stdlib. The analog `cosine.go` imports:
```go
import (
    "math"
    "sort"
)
```
`scorer.go` will import only `"sort"` (no `math` — the composite reduction uses only `+` and `*`).

**Core reduction loop pattern** (`cosine.go:341-344`):
```go
var dot float64
for _, k := range intersectionKeys {
    dot = (float64(qa[k]) * float64(qb[k])) + dot
}
```
Scorer analog (from CONTEXT.md §5 LOCKED):
```go
var acc float64
for _, entry := range s.algorithmsAlgoIDSorted {
    score := entry.scoreFn(na, nb)
    acc = acc + (entry.weight * score)  // DET-06 explicit parens
}
return acc
```
Same discipline: explicit `(a * b) + acc` parenthesisation; pre-sorted slice iterated left-to-right; no transcendentals; no per-call sort.

**dispatch table access pattern** (`algoid.go:310-324`):
```go
var dispatch [numAlgorithms]func(a, b string) float64
```
In `NewScorer` validation, the bounds check is:
```go
if int(algo) >= numAlgorithms || dispatch[algo] == nil {
    return ErrInvalidAlgorithm
}
```
`numAlgorithms` is the compile-time constant from `algoid.go:187`. Direct array index, zero allocation.

**AlgoID-sorted dedup pattern** (`algoid.go:282-308`):
```go
func AlgoIDs() []AlgoID {
    return []AlgoID{
        AlgoLevenshtein,
        AlgoDamerauLevenshteinOSA,
        // ... 23 entries in canonical iota order
    }
}
```
Use `AlgoIDs()` as the merge-iteration order in `NewScorer`'s dedup step:
```go
seen := make(map[AlgoID]scorerEntry)
for _, e := range cfg.entries {
    seen[e.id] = e  // last-write-wins: later option overwrites earlier
}
sorted := make([]scorerEntry, 0, len(seen))
for _, id := range AlgoIDs() {  // canonical order from algoid.go:282-307
    if e, ok := seen[id]; ok {
        sorted = append(sorted, e)
    }
}
```

**What's the same vs different:** The cosine reduction loop is byte-for-byte the model for `Score`'s inner loop. The only difference is that cosine's sorted keys are q-gram strings (sorted at call time from a map); the Scorer's sorted entries are `[]scorerEntry` pre-sorted once at `NewScorer` time — so there is no per-call sort on the hot path.

---

### `scorer_options.go` (config/builder, option accumulation)

**Analogs:**
- `dispatch_cosine.go:41-46` — default-parameter closure pattern
- `dispatch_monge_elkan.go:68-73` — multi-parameter closure with inner-AlgoID default

**Closure capture pattern for default dispatch** (`dispatch_cosine.go:41-46`):
```go
var _ = func() bool {
    dispatch[AlgoCosine] = func(a, b string) float64 {
        return CosineScore(a, b, 3)
    }
    return true
}()
```
`WithAlgorithm(id, weight)` uses the same slot directly rather than a wrapper:
```go
func WithAlgorithm(algo AlgoID, weight float64) ScorerOption {
    return func(cfg *scorerConfig) error {
        if weight <= 0 {
            return ErrInvalidWeight
        }
        if int(algo) >= numAlgorithms || dispatch[algo] == nil {
            return ErrInvalidAlgorithm
        }
        cfg.entries = append(cfg.entries, scorerEntry{
            id:      algo,
            weight:  weight,
            scoreFn: dispatch[algo],
        })
        return nil
    }
}
```

**Parameterised closure pattern** — inversion of the `dispatch_cosine.go` default wrapper. The dispatch wrapper binds `n=3`; the parameterised option overrides with consumer-supplied `n`:
```go
func WithQGramJaccardAlgorithm(weight float64, n int) ScorerOption {
    return func(cfg *scorerConfig) error {
        if n < 1 {
            return ErrInvalidQGramSize
        }
        if weight <= 0 {
            return ErrInvalidWeight
        }
        cfg.entries = append(cfg.entries, scorerEntry{
            id:      AlgoQGramJaccard,
            weight:  weight,
            scoreFn: func(a, b string) float64 { return QGramJaccardScore(a, b, n) },
        })
        return nil
    }
}
```

**Monge-Elkan inner dispatch pattern** (`dispatch_monge_elkan.go:68-73`):
```go
var _ = func() bool {
    dispatch[AlgoMongeElkan] = func(a, b string) float64 {
        return MongeElkanScoreSymmetric(a, b, AlgoJaroWinkler, DefaultNormalisationOptions())
    }
    return true
}()
```
`WithMongeElkanAlgorithm(weight, inner)` inverts this — consumer supplies the inner AlgoID instead of the default:
```go
func WithMongeElkanAlgorithm(weight float64, inner AlgoID) ScorerOption {
    return func(cfg *scorerConfig) error {
        // ... weight/inner validation ...
        cfg.entries = append(cfg.entries, scorerEntry{
            id:      AlgoMongeElkan,
            weight:  weight,
            scoreFn: func(a, b string) float64 {
                return MongeElkanScoreSymmetric(a, b, inner, DefaultNormalisationOptions())
            },
        })
        return nil
    }
}
```

**`SWGParams` type is confirmed** (`swg.go:97`): the exported type is `SWGParams` (a struct with `MatchScore`, `MismatchScore`, `GapOpen`, `GapExtend` fields; default constructor is `NewSWGParams()`).

**What's the same vs different:** The dispatch files show the unparameterised and parameterised default wrappers. `scorer_options.go` inverts the pattern — instead of the library binding defaults, the consumer supplies values. The `ScorerOption func(*scorerConfig) error` return type follows the standard functional-options pattern; no existing file in the codebase uses this exact shape (first use in Phase 8), but the closure-capture mechanics are identical to the dispatch wrappers.

---

### `errors.go` (append 4 new sentinels)

**Analog:** `errors.go:48-107` — the existing six sentinel declarations.

**Exact pattern to copy** (`errors.go:48-107`):
```go
// ErrInvalidInput indicates a caller-provided string fails an
// algorithm's documented input constraints ...
//
// Discriminate via errors.Is(err, fuzzymatch.ErrInvalidInput); never
// match the error message string.
var ErrInvalidInput = errors.New("fuzzymatch: invalid input")
```

Four new sentinels follow the identical form. From RESEARCH.md §Code Examples:
```go
// ErrEmptyScorer indicates NewScorer was called with no algorithm options.
// Pass at least one WithAlgorithm option or use DefaultScorer().
//
// Discriminate via errors.Is(err, fuzzymatch.ErrEmptyScorer); never
// match the error message string.
var ErrEmptyScorer = errors.New("fuzzymatch: scorer has no algorithms (pass at least one WithAlgorithm option or use DefaultScorer)")

// ErrInvalidWeight indicates an algorithm weight was ≤ 0. Weights must
// be strictly positive.
//
// Discriminate via errors.Is(err, fuzzymatch.ErrInvalidWeight); never
// match the error message string.
var ErrInvalidWeight = errors.New("fuzzymatch: invalid algorithm weight (must be > 0)")

// ErrInvalidThreshold indicates a WithThreshold value was outside [0.0, 1.0].
//
// Discriminate via errors.Is(err, fuzzymatch.ErrInvalidThreshold); never
// match the error message string.
var ErrInvalidThreshold = errors.New("fuzzymatch: invalid threshold (must be in [0.0, 1.0])")

// ErrMissingThreshold indicates NewScorer was called without WithThreshold.
// The threshold is a calibration parameter with no universally-safe default.
// Pass WithThreshold(t) with t ∈ [0.0, 1.0], or use DefaultScorer() which
// bakes 0.85 in.
//
// Returned by NewScorer when no WithThreshold option is present.
//
// Discriminate via errors.Is(err, fuzzymatch.ErrMissingThreshold); never
// match the error message string.
var ErrMissingThreshold = errors.New("fuzzymatch: scorer threshold required (pass WithThreshold or use DefaultScorer)")
```

**Validation pipeline order** (from CONTEXT.md §2 LOCKED): the check order in `NewScorer` must be:
1. `!cfg.thresholdSet` → `ErrMissingThreshold` (first, before everything else)
2. `len(cfg.entries) == 0` → `ErrEmptyScorer`
3. Individual weight validation → `ErrInvalidWeight`
4. `cfg.threshold` range → `ErrInvalidThreshold`
5. Per-entry AlgoID validity → `ErrInvalidAlgorithm`
6. Per-entry parameter validation (q-gram n, Tversky α/β) → `ErrInvalidQGramSize` / `ErrInvalidTverskyParam`

**What's the same vs different:** Identical `var ErrXxx = errors.New(...)` pattern with the same godoc structure (description + `Discriminate via errors.Is` guidance). The new sentinels append after `ErrEmptyInput` (line 107). No new import needed — `errors` is already imported at `errors.go:36`.

---

### `scorer_test.go` (test, stdlib testing)

**Analog:** `cosine_test.go:1-42` — file header, package declaration, import block.

**Package declaration and imports pattern** (`cosine_test.go:36-41`):
```go
package fuzzymatch_test

import (
    "math"
    "testing"

    "github.com/axonops/fuzzymatch"
)
```
`scorer_test.go` uses `package fuzzymatch_test`. No testify. No `testing/quick` in the file header imports (property tests go in a separate section or the same file but imported separately as needed).

**Table-driven test pattern** (`cosine_test.go:54-63`):
```go
func TestCosine_BothEmpty(t *testing.T) {
    for _, n := range []int{1, 2, 3, 5} {
        if got := fuzzymatch.CosineScore("", "", n); got != 1.0 {
            t.Errorf("CosineScore(\"\", \"\", %d) = %g; want 1.0", n, got)
        }
    }
}
```

**Property test pattern** (`normalise_test.go:283-293`):
```go
func TestProp_Normalise_Idempotent(t *testing.T) {
    opts := fuzzymatch.DefaultNormalisationOptions()
    f := func(s string) bool {
        once := fuzzymatch.Normalise(s, opts)
        twice := fuzzymatch.Normalise(once, opts)
        return once == twice
    }
    if err := quick.Check(f, nil); err != nil {
        t.Errorf("Normalise is not idempotent under default options: %v", err)
    }
}
```
Scorer property tests follow the identical shape:
```go
func TestProp_Scorer_DeterministicAcrossRuns(t *testing.T) {
    s := fuzzymatch.DefaultScorer()
    f := func(a, b string) bool {
        return s.Score(a, b) == s.Score(a, b)
    }
    if err := quick.Check(f, nil); err != nil {
        t.Errorf("PropScorer_DeterministicAcrossRuns: %v", err)
    }
}
```

**Concurrent safety test pattern** — no existing file uses `sync.WaitGroup` for concurrent Scorer tests (first use in Phase 8), but the project constraint is explicit: `sync.WaitGroup` only (no `errgroup` — `golang.org/x/sync` not in root `go.mod`):
```go
func TestScorer_ConcurrentSafety(t *testing.T) {
    s := fuzzymatch.DefaultScorer()
    const n = 100
    var wg sync.WaitGroup
    wg.Add(n)
    results := make([]float64, n)
    for i := 0; i < n; i++ {
        go func(idx int) {
            defer wg.Done()
            results[idx] = s.Score("user_id", "userId")
        }(i)
    }
    wg.Wait()
    for i, got := range results {
        if got != results[0] {
            t.Errorf("goroutine %d: got %g; want %g (first result)", i, got, results[0])
        }
    }
}
```

**What's the same vs different:** Identical file structure (header, `package fuzzymatch_test`, stdlib imports). Tests use `t.Errorf`/`t.Fatalf` with `%g` for floats, `%q` for strings. No testify. Property tests use `testing/quick`. Concurrent test is a new pattern but uses only `sync.WaitGroup` (stdlib).

---

### `scorer_options_test.go` (test, options)

**Analog:** `cosine_test.go` (pattern — same package, same table-driven style).

**Error-path test pattern** — the cosine test for the `n < 1` panic is the closest existing pattern. For options, tests assert returned `error` values instead of panics:
```go
func TestWithAlgorithm_InvalidWeight(t *testing.T) {
    _, err := fuzzymatch.NewScorer(
        fuzzymatch.WithAlgorithm(fuzzymatch.AlgoLevenshtein, -0.5),
        fuzzymatch.WithThreshold(0.7),
    )
    if !errors.Is(err, fuzzymatch.ErrInvalidWeight) {
        t.Errorf("NewScorer with weight=-0.5: got %v; want ErrInvalidWeight", err)
    }
}
```

**What's the same vs different:** Same `package fuzzymatch_test`, same `t.Errorf` pattern. The errors are returned values (not panics), so tests use `errors.Is` rather than `recover()`. The panic-recovery pattern from `algorithms_steps.go` is for direct algorithm calls, not for option constructors.

---

### `scorer_internal_test.go` (test, package-internal)

**Analog:** There are no existing `_internal_test.go` files in the codebase. The closest pattern is any `_test.go` file that uses `package fuzzymatch` (not `package fuzzymatch_test`) — the `normalise_test.go` uses `package fuzzymatch_test`; the internal package approach is new in Phase 8.

**Package declaration for internal tests:**
```go
package fuzzymatch
```
(No `_test` suffix — gives access to unexported `scorerConfig`, `scorerEntry`, etc.)

**Last-write-wins test shape:**
```go
// TestScorerConfig_LastWriteWins verifies that when WithAlgorithm is called
// twice with the same AlgoID, only the latter weight survives in the
// constructed Scorer.
func TestScorerConfig_LastWriteWins(t *testing.T) {
    s, err := NewScorer(
        WithAlgorithm(AlgoLevenshtein, 0.3),
        WithAlgorithm(AlgoLevenshtein, 0.7),  // wins
        WithThreshold(0.5),
    )
    if err != nil {
        t.Fatalf("NewScorer: %v", err)
    }
    algos := s.Algorithms()
    if len(algos) != 1 {
        t.Fatalf("want 1 algorithm; got %d", len(algos))
    }
    // Weight is normalised (1 algo → weight = 1.0 after normalisation),
    // but the entry must correspond to the second WithAlgorithm call.
    if algos[0].ID != AlgoLevenshtein {
        t.Errorf("want AlgoLevenshtein; got %v", algos[0].ID)
    }
}
```

**What's the same vs different:** The file uses `package fuzzymatch` (no `_test` suffix) to access unexported types. The test structure (`t.Fatalf`/`t.Errorf`, no testify) is identical to all other root module tests.

---

### `scorer_golden_test.go` (test, golden gate)

**Analog:** `algorithms_golden_test.go:1-100` + `golden_test.go:66-88`.

**`assertGolden` call pattern** (`golden_test.go:66-88`):
```go
func assertGolden(t *testing.T, filename string, v any) {
    t.Helper()
    got, err := fuzzymatch.CanonicalMarshalForTest(v)
    // ... reads testdata/golden/<filename>, compares bytes ...
}
```

**Golden file schema struct pattern** (`algorithms_golden_test.go:51-64`):
```go
type goldenAlgorithmEntry struct {
    Name          string  `json:"name"`
    Algorithm     string  `json:"algorithm"`
    A             string  `json:"a"`
    B             string  `json:"b"`
    ExpectedScore float64 `json:"expected_score"`
}
type goldenAlgorithmsFile struct {
    Version int                    `json:"version"`
    Entries []goldenAlgorithmEntry `json:"entries"`
}
```
`scorer_golden_test.go` defines analogous structs (from CONTEXT.md §6, noting Pitfall 6: `ScoreAll` in the JSON uses `map[string]float64` with `AlgoID.String()` keys, not `map[AlgoID]float64`):
```go
type scorerGoldenEntry struct {
    A          string             `json:"a"`
    B          string             `json:"b"`
    Score      float64            `json:"score"`
    Match      bool               `json:"match"`
    ScoreAll   map[string]float64 `json:"scoreAll"`  // AlgoID.String() keys
    ScorerConf string             `json:"scorer_config"`
}
type scorerGoldenMetadata struct {
    Phase           int    `json:"phase"`
    ScorerSignature string `json:"scorer_signature"`
}
type scorerGoldenFile struct {
    Metadata scorerGoldenMetadata `json:"_metadata"`
    Entries  []scorerGoldenEntry  `json:"entries"`
}
```
Note: `generated_at` is omitted from the schema (per RESEARCH.md Open Question 1 — `algorithms.json` has no timestamp, so `scorer-default.json` should follow the same pattern for byte-identical CI gate stability).

**`TestGolden_ScorerDefault` invocation:**
```go
func TestGolden_ScorerDefault(t *testing.T) {
    s := fuzzymatch.DefaultScorer()
    entries := buildScorerGoldenEntries(t, s)
    assertGolden(t, "scorer-default.json", scorerGoldenFile{
        Metadata: scorerGoldenMetadata{
            Phase:           8,
            ScorerSignature: "DefaultScorer-2026-05-16",
        },
        Entries: entries,
    })
}
```
The `-update` flag is already declared at package level in `golden_test.go:52` — `scorer_golden_test.go` shares it automatically (same package `fuzzymatch_test`).

**What's the same vs different:** Identical use of `assertGolden` (same helper, same package). The schema struct differs (Scorer fields vs algorithm fields). The `ScoreAll` map in the JSON uses `string` keys for human-readability, while the public API returns `map[AlgoID]float64`.

---

### `scorer_bench_test.go` (test, benchmark)

**Analog:** `cosine_bench_test.go:1-119` — exact structural template.

**File header and package** (`cosine_bench_test.go:1-45`):
```go
// cosine_bench_test.go runs allocation-aware benchmarks ...
//
// Performance budget per RESEARCH.md §4.1 ...
//   - ASCII Short  (~5 chars):    ≤ 5 allocs/op
//   - ASCII Medium (~50 chars):   ≤ 7 allocs/op

package fuzzymatch_test

import (
    "strings"
    "testing"

    "github.com/axonops/fuzzymatch"
)
```

**Benchmark function pattern** (`cosine_bench_test.go:61-71`):
```go
func BenchmarkCosineScore_ASCII_Short(b *testing.B) {
    b.ReportAllocs()
    b.ResetTimer()
    var sink float64
    for i := 0; i < b.N; i++ {
        sink = fuzzymatch.CosineScore("abc", "abcd", 2)
    }
    if sink < 0 {
        b.Fatal("sink unexpectedly negative — compiler folded the benchmark away")
    }
}
```
`scorer_bench_test.go` follows byte-for-byte:
```go
func BenchmarkDefaultScorer_Score_ASCII_Short(b *testing.B) {
    s := fuzzymatch.DefaultScorer()
    b.ReportAllocs()
    b.ResetTimer()
    var sink float64
    for i := 0; i < b.N; i++ {
        sink = s.Score("abc", "abcd")
    }
    if sink < 0 {
        b.Fatal("sink unexpectedly negative — compiler folded the benchmark away")
    }
}
```
Scorer-specific addition: `s := fuzzymatch.DefaultScorer()` allocated once before `b.ResetTimer()` (construction cost is not part of the Score budget). Budget is `< 30µs / ≤ 8 allocs` per spec §14.2.

**What's the same vs different:** Identical `b.ReportAllocs()` + `b.ResetTimer()` + `var sink` + anti-dead-code-elimination gate. Only difference is that Scorer.Score pre-normalises (potentially 1-2 allocations for ASCII via Normalise's fast path) and dispatches to 6 algorithms; the Cosine benchmark dispatches to a single algorithm. The 4 input-size variants (Short/Medium/Long plus Unicode if applicable) follow the same naming convention.

---

### `tests/bdd/features/scorer.feature` (test, BDD)

**Analog:** `tests/bdd/features/cosine.feature:1-97` — the most structurally complete existing feature file (has Scenario Outline, Examples tables, standalone Scenarios, and symmetry scenario).

**Feature preamble pattern** (`cosine.feature:22-29`):
```gherkin
Feature: Cosine n-gram similarity (vector-space cosine over q-gram frequency vectors)
  Cosine is the textbook information-retrieval similarity measure:
  cos(A, B) = (A · B) / (‖A‖ × ‖B‖) ...
```

**Scenario pattern for error paths** (`cosine.feature:59-64` one-empty):
```gherkin
Scenario Outline: one-empty string scores 0.0
  When I compute the Cosine score between "<a>" and "<b>" with n 2
  Then the score should be exactly 0

  Examples:
    | a   | b   |
    | abc |     |
    |     | abc |
```

**`scorer.feature` mandatory scenarios (CONTEXT.md §7):**
1. Default scorer happy path (identifier match → `Match` true)
2. Default scorer below threshold (dissimilar → `Match` false)
3. Custom 1-algorithm Scorer (Levenshtein only + `WithThreshold`)
4. Custom 2-algorithm weighted Scorer
5. `WithoutAlgorithm` composition removes the AlgoID
6. Last-write-wins for duplicate `WithAlgorithm` calls
7. `WithoutNormalisation` → no match on `XMLParser` / `xml_parser`
8. `ErrMissingThreshold` — `NewScorer` without `WithThreshold`
9. `ErrEmptyScorer` — `NewScorer` without any algorithm
10. `ErrInvalidWeight` — negative weight
11. Concurrent safety (N goroutines, goleak gate)
12. `ScoreAll` AlgoID keys display via `String()`

**What's the same vs different:** Same Gherkin syntax, same `Scenario Outline` / `Examples` pattern. The Scorer feature does not need floating-point tolerance scenarios for most cases (composite scores will be asserted with `approximately X within Y` where the exact value is pinned after golden generation). Error-path scenarios use exact string matching on error type names.

---

### `tests/bdd/steps/scorer_steps.go` (test, BDD steps)

**Analog:** `tests/bdd/steps/algorithms_steps.go:40-58` — `AlgorithmContext` struct + step function signatures + `InitializeScenario` registration.

**Context struct pattern** (`algorithms_steps.go:46-58`):
```go
type AlgorithmContext struct {
    lastScore    float64
    lastScore2   float64
    lastDistance int
    lastPanicMsg string
    lastCode     string
    // Phase 7 fields...
}
```
`scorer_steps.go` introduces a parallel `ScorerContext`:
```go
type ScorerContext struct {
    scorer    *fuzzymatch.Scorer
    lastScore float64
    lastMatch bool
    lastErr   error
    scoreAll  map[fuzzymatch.AlgoID]float64
}
```

**Step function pattern** (`algorithms_steps.go:62-65`):
```go
func (ctx *AlgorithmContext) iComputeTheLevenshteinScoreBetween(a, b string) error {
    ctx.lastScore = fuzzymatch.LevenshteinScore(a, b)
    return nil
}
```

**`InitializeScenario` registration pattern** (`algorithms_steps.go:1131-1155`):
```go
func InitializeScenario(ctx *godog.ScenarioContext) {
    a := &AlgorithmContext{}
    ctx.Step(
        `^I compute the Levenshtein score between "([^"]*)" and "([^"]*)"$`,
        a.iComputeTheLevenshteinScoreBetween,
    )
    // ... more ctx.Step registrations ...
}
```
`scorer_steps.go` adds an `InitScorerSteps(ctx *godog.ScenarioContext)` function (note: same `InitializeScenario` is already registered in `bdd_test.go`; the planner must extend `InitializeScenario` to call `InitScorerSteps(ctx)` or register the `ScorerContext` steps directly in `InitializeScenario`). The simpler approach per RESEARCH.md is a new function `InitScorerSteps` called from within `InitializeScenario` in a second `var sc = &ScorerContext{}` block.

**What's the same vs different:** Identical `package steps`, same `godog.ScenarioContext` step registration, same `return error` step function signature, same `([^"]*)` regex for string captures. The `ScorerContext` holds a `*fuzzymatch.Scorer` (construction happens in a "Given" step) and tracks `lastErr` (for error-path scenarios — `AlgorithmContext` uses `lastPanicMsg` for panics, but Scorer construction errors are returned values not panics).

---

### `testdata/golden/scorer-default.json` (data file)

**Analog:** `testdata/golden/algorithms.json` — schema shape.

**Schema from `algorithms_golden_test.go:51-64`:**
```go
type goldenAlgorithmsFile struct {
    Version int                    `json:"version"`
    Entries []goldenAlgorithmEntry `json:"entries"`
}
```

**`scorer-default.json` schema** (from CONTEXT.md §6):
```json
{
  "_metadata": {
    "phase": 8,
    "scorer_signature": "DefaultScorer-2026-05-16"
  },
  "entries": [
    {
      "a": "user_id",
      "b": "userId",
      "score": 0.987,
      "match": true,
      "scoreAll": {
        "DamerauLevenshteinOSA": 1.0,
        "JaroWinkler": 1.0,
        "TokenJaccard": 1.0,
        "QGramJaccard": 1.0,
        "SorensenDice": 1.0,
        "DoubleMetaphone": 1.0
      },
      "scorer_config": "DefaultScorer"
    }
  ]
}
```

**Canonical form requirements** (`golden_canonical.go:62-73`):
```go
func canonicalMarshal(v any) ([]byte, error) {
    body, err := json.MarshalIndent(v, "", "  ")  // two-space indent
    // ... appends single trailing "\n" ...
}
```
The file is regenerated via `go test -run TestGolden_ -update ./...`.

**Critical: `scoreAll` uses `AlgoID.String()` keys** (not integer AlgoID values) so the JSON is human-readable. The serialisation struct uses `map[string]float64`; the conversion iterates `AlgoIDs()` (canonical order) to build the string-keyed map — this avoids map iteration on an `AlgoID`-keyed map which would produce non-deterministic JSON key order.

**What's the same vs different:** Same `json.MarshalIndent` two-space canonical form. Different schema — Scorer entries include `score`, `match`, `scoreAll`, and `scorer_config` fields. No `generated_at` timestamp (omitted for byte-stability, mirroring `algorithms.json`).

---

### `examples/scorer-composition/main.go` (example program)

**Analog:** `examples/identifier-similarity/main.go:44-168` — full structure.

**Package declaration and imports pattern** (`identifier-similarity/main.go:44-50`):
```go
package main

import (
    "fmt"

    "github.com/axonops/fuzzymatch"
)
```

**Program structure pattern** (`identifier-similarity/main.go:140-168`):
```go
func main() {
    const pairWidth = 32
    const algoWidth = 13
    // header row ...
    // separator line ...
    // data rows ...
}
```

`examples/scorer-composition/main.go` demonstrates the `DefaultScorerOptions() + WithoutAlgorithm + WithThreshold` pattern from CONTEXT.md §Specific Ideas:
```go
// Demonstrating: default scorer minus phonetic, for numeric-identifier data
opts := append(fuzzymatch.DefaultScorerOptions(),
    fuzzymatch.WithoutAlgorithm(fuzzymatch.AlgoDoubleMetaphone),
    fuzzymatch.WithThreshold(0.80),
)
s, _ := fuzzymatch.NewScorer(opts...)
```

**What's the same vs different:** Same `package main`, same `import "fmt"` + `"github.com/axonops/fuzzymatch"`, same `fmt.Printf` table output. The example is narrower (demonstrates one composition pattern) rather than showing all 23 algorithms.

---

### `examples/scorer-composition/main_test.go` (example stdout gate)

**Analog:** `examples/identifier-similarity/main_test.go:28-135` — exact structural template.

**stdout capture pattern** (`identifier-similarity/main_test.go:60-87`):
```go
func TestExample_Output(t *testing.T) {
    origStdout := os.Stdout
    r, w, err := os.Pipe()
    if err != nil { t.Fatalf("...") }
    os.Stdout = w
    defer func() { os.Stdout = origStdout }()

    main()

    w.Close()
    var buf bytes.Buffer
    if _, err := io.Copy(&buf, r); err != nil { t.Fatalf("...") }
    r.Close()

    got := buf.String()
    if got == want { return }
    // ... line-by-line diff ...
}
```

**`want` constant** — committed byte-stable expected stdout. Updated by running `go run .` and pasting:
```go
const want = `...committed output...`
```

**What's the same vs different:** Identical `os.Pipe()` redirect mechanism. The `want` constant will be shorter (fewer pairs, one composition pattern). The test is in `package main` (same file as the example, no `_test` suffix) to call `main()` directly.

---

### `examples/identifier-similarity/main.go` (extend — append Score + Match columns)

**Analog:** itself, lines 108-168. The extension appends two new entries to the `algorithms` slice and extends the column format.

**Current `algorithms` slice tail** (`main.go:133-138`):
```go
{"Soundex", fuzzymatch.SoundexScore},
{"DblMetaph", fuzzymatch.DoubleMetaphoneScore},
{"NYSIIS", fuzzymatch.NYSIISScore},
{"MRA", fuzzymatch.MRAScore},
```
Phase 8 appends after `MRA`:
```go
{"Score", func(a, b string) float64 { return fuzzymatch.DefaultScorer().Score(a, b) }},
{"Match", func(a, b string) float64 {
    if fuzzymatch.DefaultScorer().Match(a, b) { return 1.0 }
    return 0.0
}},
```
(Or using a pre-constructed `DefaultScorer()` at package level to avoid constructing it per-cell.)

**`want` constant in `main_test.go`** must also be updated to include the two new columns — that constant is the byte-stable gate for the extended output.

**What's the same vs different:** Same slice-append pattern. The `main_test.go` `want` constant requires regeneration (run `go run .`, paste output). The table-format constants (`pairWidth`, `algoWidth`) remain unchanged.

---

### `docs/scorer.md` (populate scaffold)

**Analog:** `docs/algorithms.md:1-43` — the header and per-section scaffold structure; contrast with `docs/requirements.md` for a fully populated reference doc.

**Current scaffold state** (`docs/scorer.md:1-12`):
```markdown
# Scorer (Phase 8)

The `Scorer` is the second layer ...

This document is a scaffold. The Scorer lands in Phase 8 ...

## Construction
TBD. See `docs/requirements.md` §8.1 ...
```

The populated doc replaces each `TBD` section with prose, the mandatory quickstart code block (CONTEXT.md §Specific Ideas), the `WithThreshold` required-parameter explanation, the `WithoutNormalisation` semantics, and the `DefaultScorer` composition table.

**What's the same vs different:** The file structure (H1, H2 section headers, code fences) follows `docs/algorithms.md`'s prose pattern. Phase 8 is the first heavy docs-writer use — `docs/scorer.md` and `docs/tuning.md` are the first fully-populated non-requirement docs.

---

### `docs/tuning.md` (populate scaffold)

**Analog:** `docs/algorithms.md` — same scaffold-to-prose pattern.

Mandatory sections per CONTEXT.md §Specific Ideas:
- "How to pick a threshold" — recommends scanning consumer's data corpus, recording false-positive / false-negative trade-off, adjusting in 0.05 increments, forward reference to Phase 9 scan layer.
- "How to pick weights" — start equal-weight, use golden corpus to identify which algorithms best discriminate the domain.

**What's the same vs different:** Same Markdown structure. First time `docs/tuning.md` is populated.

---

### `llms.txt` / `llms-full.txt` (extend)

**Analog:** themselves — Phase 5+ per-plan llms-sync discipline. Each new exported symbol gets a line in `llms.txt` and a full entry in `llms-full.txt` in the same plan that adds the symbol.

New exported symbols in Phase 8:
- `NewScorer`, `Scorer`, `DefaultScorer`, `DefaultScorerOptions`
- `Score`, `ScoreAll`, `Match`, `Threshold`, `Algorithms`
- `ScorerAlgorithm` struct
- `ScorerOption` type
- All `With*` option functions (12 total)
- `ErrEmptyScorer`, `ErrInvalidWeight`, `ErrInvalidThreshold`, `ErrMissingThreshold`

**What's the same vs different:** Identical append pattern. The `ai_friendly_test.go` meta-test (`go/ast` scan) will verify all exported symbols are listed.

---

## Shared Patterns

### File header (copyright + licence)
**Source:** Any existing `.go` file, e.g. `errors.go:1-13`
**Apply to:** All new `.go` files (scorer.go, scorer_options.go, scorer_golden_test.go, scorer_bench_test.go, scorer_test.go, scorer_options_test.go, scorer_internal_test.go, tests/bdd/steps/scorer_steps.go, examples/scorer-composition/main.go, examples/scorer-composition/main_test.go)
```go
// Copyright 2026 AxonOps Limited
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
```

### Sentinel error pattern
**Source:** `errors.go:48-107`
**Apply to:** `errors.go` (4 new sentinels appended), and any `errors.Is` checks in `scorer_test.go` and `scorer_options_test.go`
```go
var ErrXxx = errors.New("fuzzymatch: <lowercase message without trailing punctuation>")
// godoc: Discriminate via errors.Is(err, fuzzymatch.ErrXxx); never match the error message string.
```

### dispatch table bounds check
**Source:** `algoid.go:187` (`numAlgorithms`), `algoid.go:310-324` (`dispatch` array)
**Apply to:** `NewScorer` validation in `scorer.go` (or `scorer_options.go` — in `WithAlgorithm`)
```go
if int(algo) >= numAlgorithms || dispatch[algo] == nil {
    return ErrInvalidAlgorithm
}
```

### No-init dispatch registration idiom
**Source:** `dispatch_cosine.go:41-46` (and all other `dispatch_*.go` files)
**Apply to:** `scorer.go` / `scorer_options.go` — Phase 8 does NOT add new dispatch entries; this pattern is cited as a negative (do not use `init()`, do not add new `dispatch_*.go` files)
```go
var _ = func() bool {
    dispatch[AlgoX] = func(a, b string) float64 { ... }
    return true
}()
```

### Float-determinism reduction
**Source:** `cosine.go:341-344`
**Apply to:** `scorer.go` `Score` method
```go
var acc float64
for _, entry := range s.algorithmsAlgoIDSorted {
    score := entry.scoreFn(na, nb)
    acc = acc + (entry.weight * score)  // DET-06 explicit parens — mirrors cosine.go:343
}
```

### AlgoIDs() canonical sort
**Source:** `algoid.go:282-308`
**Apply to:** `NewScorer` dedup step in `scorer.go`, `Algorithms()` method, `ScoreAll` serialisation in `scorer_golden_test.go`
```go
for _, id := range AlgoIDs() {
    if e, ok := seen[id]; ok {
        sorted = append(sorted, e)
    }
}
```

### Golden file `assertGolden` helper
**Source:** `golden_test.go:66-88`; `golden_canonical.go:61-73` (`canonicalMarshal`)
**Apply to:** `scorer_golden_test.go`
```go
assertGolden(t, "scorer-default.json", scorerGoldenFile{...})
// -update flag is shared from golden_test.go:52 (package-level var)
```

### BDD `InitializeScenario` extension
**Source:** `tests/bdd/steps/algorithms_steps.go:1131-1180`
**Apply to:** `tests/bdd/steps/scorer_steps.go` — new `ScorerContext` + step functions, registered via a new `InitScorerSteps` called from within `InitializeScenario`
```go
func InitializeScenario(ctx *godog.ScenarioContext) {
    a := &AlgorithmContext{}
    // ... existing algorithm steps ...
    InitScorerSteps(ctx)  // NEW: scorer steps
}
```

### Example stdout gate (`os.Pipe` redirect)
**Source:** `examples/identifier-similarity/main_test.go:60-87`
**Apply to:** `examples/scorer-composition/main_test.go`
```go
origStdout := os.Stdout
r, w, _ := os.Pipe()
os.Stdout = w
defer func() { os.Stdout = origStdout }()
main()
w.Close()
// ... io.Copy + compare to committed `want` constant ...
```

---

## No Analog Found

All 17 files have close analogs. There are no files with no match.

The two partial-match cases worth noting for the planner:

| File | Role | Data Flow | Reason |
|------|------|-----------|--------|
| `scorer_internal_test.go` | test (package-internal) | request-response | No existing `package fuzzymatch` (non-`_test`) test file; the pattern is inferred from stdlib conventions and the `package fuzzymatch_test` files |
| `scorer_options.go` functional-options constructor | config/builder | option accumulation | No existing `ScorerOption func(*scorerConfig) error` functional-options constructor in the codebase; the individual `With*` closures map to the `dispatch_*.go` patterns but the composing constructor is new |

---

## Metadata

**Analog search scope:** `/Users/johnny/Development/fuzzymatch/` (root module), `tests/bdd/`, `examples/`
**Files read:** `errors.go`, `algoid.go` (full), `cosine.go` (lines 280-370), `dispatch_cosine.go`, `dispatch_monge_elkan.go`, `golden_test.go`, `golden_canonical.go`, `algorithms_golden_test.go` (lines 1-100), `monge_elkan.go` (grep), `swg.go` (grep), `tests/bdd/bdd_test.go`, `tests/bdd/steps/algorithms_steps.go` (lines 1-120, 1121-1180), `tests/bdd/features/cosine.feature`, `cosine_test.go` (lines 1-100), `cosine_bench_test.go`, `normalise_test.go` (lines 270-347), `examples/identifier-similarity/main.go`, `examples/identifier-similarity/main_test.go`, `docs/scorer.md` (lines 1-30), `docs/algorithms.md` (lines 1-43)
**Pattern extraction date:** 2026-05-16

---

## PATTERN MAPPING COMPLETE

**Phase:** 08 - composite-scorer
**Files classified:** 17
**Analogs found:** 17 / 17

### Coverage
- Files with exact analog: 13 (`errors.go` append, `scorer_test.go`, `scorer_options_test.go`, `scorer_golden_test.go`, `scorer_bench_test.go`, `scorer.feature`, `scorer_steps.go`, `scorer-default.json`, `scorer-composition/main.go`, `scorer-composition/main_test.go`, `identifier-similarity/main.go` extend, `llms.txt`/`llms-full.txt` extend, `bdd_test.go` reuse)
- Files with role-match analog: 4 (`scorer.go`, `scorer_options.go`, `docs/scorer.md`, `docs/tuning.md`)
- Files with no analog: 0 (but 2 partial-match cases noted above)

### Key Patterns Identified
- All new root-module test files use `package fuzzymatch_test` (or `package fuzzymatch` for `scorer_internal_test.go`), stdlib `testing` only, `t.Errorf`/`t.Fatalf`, no testify
- The `Score` method's inner loop copies `cosine.go:341-344`'s `acc = acc + (entry.weight * score)` form verbatim — explicit parens, left-to-right, pre-sorted slice
- All `With*` option functions use the `func(cfg *scorerConfig) error` closure pattern; parameterised options capture consumer-supplied params; non-parameterised options store `dispatch[algo]` directly
- The golden file harness (`assertGolden` + `-update` flag + `canonicalMarshal`) is fully operational; `scorer_golden_test.go` reuses without modification
- `tests/bdd/bdd_test.go` is reused unchanged; adding `scorer.feature` to `tests/bdd/features/` is sufficient for godog discovery; new step definitions go in `tests/bdd/steps/scorer_steps.go` with a new `InitScorerSteps` function registered from within the existing `InitializeScenario`

### File Created
`/Users/johnny/Development/fuzzymatch/.planning/phases/08-composite-scorer/08-PATTERNS.md`

### Ready for Planning
Pattern mapping complete. Planner can now reference analog patterns in PLAN.md files.
