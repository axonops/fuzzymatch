# Phase 8: Composite Scorer - Research

**Researched:** 2026-05-16
**Domain:** Functional-options weighted Scorer, float-determinism composite reduction, golden-file harness, BDD goleak integration
**Confidence:** HIGH

---

<user_constraints>
## User Constraints (from CONTEXT.md)

### Locked Decisions

1. **ScoreAll key type — `map[AlgoID]float64`** (SPEC OVERRIDE over `docs/requirements.md` §8.3's `map[string]float64`). Typed enum keys for compile-time safety. Non-deterministic iteration order is documented; internal computation iterates AlgoID-sorted order.
2. **Threshold is MANDATORY for `NewScorer`** — returns `ErrMissingThreshold` if no `WithThreshold` option is passed. `DefaultScorer()` bakes 0.85 and cannot fail. Missing-threshold check fires FIRST in the validation pipeline.
3. **Normalisation flow** — pre-normalise once at the Scorer boundary; pass `na, nb` to ALL algorithms. Token-based algorithms keep their internal `Tokenise(s, DefaultTokeniseOptions())` calls unchanged. ME's vestigial `opts NormalisationOptions` parameter remains accepted-but-ignored (`_ = opts`). No token-based algorithm files are touched in Phase 8.
4. **Four plans in strict linear sequence** — 08-01 (errors+options), 08-02 (NewScorer+Score+Match), 08-03 (ScoreAll+Threshold+Algorithms+DefaultScorer+property/concurrent), 08-04 (golden+BDD+docs+examples). No parallel waves.
5. **Float-determinism reduction** — explicit `acc = acc + (entry.weight * score)` parenthesisation, AlgoID-sorted iteration, no transcendentals. This is the load-bearing correctness pattern from Phase 5 Cosine.
6. **Golden corpus** — reuse identifier-similarity 14-row corpus plus 8-12 Scorer-specific entries (22-26 total). Schema uses `AlgoID.String()` for `scoreAll` keys. Generated via `go test -run TestGolden_ -update ./...`.
7. **BDD coverage** — 8-12 scenarios in `tests/bdd/features/scorer.feature`. 12 required scenario classes specified exactly in CONTEXT.md §7.
8. **ME opts stays as no-op** — Phase 8 does NOT modify `monge_elkan.go` or any token-based algorithm file.

### Claude's Discretion

- Exact `scorerConfig` internal layout (slice-of-entries vs map-keyed-by-AlgoID)
- Closure-capture mechanism for parameterised algorithms
- Exact godoc wording beyond SPEC-OVERRIDE notice and ErrMissingThreshold paragraph
- `scorer-default.json` exact entry count between 22 and 26
- BDD scenario count between 8 and 12
- Whether concurrent test uses `t.Parallel` + `sync.WaitGroup` (must be `sync.WaitGroup` — root is stdlib-only)
- `docs/scorer.md` and `docs/tuning.md` exact prose structure (docs-writer + user-guide-reviewer have authority)
- Whether `examples/scorer-composition/main.go` is single-file or split
- Whether to introduce `scorer_internal_test.go` for unexported `scorerConfig` invariants (recommended yes)
- Allocation budget actual count recorded in plan 08-04 bench fixture (spec says ≤ 8 allocs)

### Deferred Ideas (OUT OF SCOPE)

- Drop ME's vestigial `opts` parameter (Phase 11 API-freeze decision)
- Amend `docs/requirements.md` §8.3 to say `map[AlgoID]float64` (lands in plan 08-04 docs commit, not a Phase 8 scope expansion)
- Amend REQUIREMENTS.md SCORER-08 to fix `WithCustomNormalisation` → `WithNormalisation` (plan 08-04 docs commit)
- Pooled `transform.Transformer` for Normalise (v1.x perf revisit)
- `DefaultScorer` composition v1.x revisit (Phase 11 integration shakedown)
- Scorer-level allocation reuse via `sync.Pool` (v1.x perf opportunity)
- `Scorer.MarshalJSON` (out of v1.0 scope)
- Threshold-edge BDD scenarios on exhaustive option combinations (v1.x)
</user_constraints>

---

<phase_requirements>
## Phase Requirements

| ID | Description | Research Support |
|----|-------------|------------------|
| SCORER-01 | `NewScorer(opts ...ScorerOption)` — immutable after construction, concurrent-safe | Confirmed: no mutexes needed; struct is read-only after `NewScorer` returns. dispatch[] is populated at package-load via `var _ = func()` idiom before any Scorer call. |
| SCORER-02 | `DefaultScorer()` + `DefaultScorerOptions()` | Confirmed: `AlgoIDs()` canonical order and dispatch table are the building blocks. 6-algorithm composition per spec §8.5. |
| SCORER-03 | Auto-normalised weights (sum-to-1 invariant) via `WithAlgorithm(AlgoID, weight)` | Confirmed: divide each weight by sum at NewScorer time when `WithNormaliseWeights(true)` (default). |
| SCORER-04 | `Score(a, b) float64` — composite weighted score in `[0.0, 1.0]` | Confirmed: `acc = acc + (entry.weight * score)` reduction loop over AlgoID-sorted entries. Pre-normalise once before the loop. |
| SCORER-05 | `ScoreAll(a, b) map[AlgoID]float64` — typed AlgoID keys (SPEC OVERRIDE confirmed) | Confirmed: fresh map per call; non-deterministic iteration order is documented. Internal computation uses sorted order regardless. |
| SCORER-06 | `Match(a, b) bool` + `Threshold()` accessor | Confirmed: `Match` delegates to `Score` and compares against the stored threshold. `Threshold()` is a plain accessor returning the float64. |
| SCORER-07 | `Algorithms() []ScorerAlgorithm` — configured AlgoID set in stable order | Confirmed: returns a fresh slice per call, sorted by AlgoID ascending. `ScorerAlgorithm{ID AlgoID, Weight float64}` struct (post-normalisation weight). |
| SCORER-08 | Normalisation control via `WithoutNormalisation()` / `WithNormalisation(opts)` | Confirmed: `applyNormalisation bool` + `normaliseOpts NormalisationOptions` fields in `scorerConfig`. `WithoutNormalisation()` sets `applyNormalisation = false`; `WithNormalisation(opts)` sets `applyNormalisation = true` + stores `opts`. |
</phase_requirements>

---

## Summary

Phase 8 ships the composite weighted Scorer (Layer 2 of the three-layer fuzzymatch architecture). The design is thoroughly specified in `08-CONTEXT.md` — eight locked decisions leave the planner with implementation choices (config struct layout, closure mechanics) rather than design choices. The research confirms every architectural assumption against the actual codebase and extracts the concrete file:line patterns the planner needs to produce well-specified tasks.

The most important technical detail is the **float-determinism reduction loop**. The Phase 5 Cosine implementation (`cosine.go:343`) established the load-bearing pattern: `dot = (float64(qa[k]) * float64(qb[k])) + dot` with a sorted-keys pre-pass. The Scorer's `Score` method applies the identical discipline at a higher level: `acc = acc + (entry.weight * score)` with `s.algorithmsAlgoIDSorted` iterated in AlgoID-ascending order (sort once at `NewScorer` time).

The second critical detail is the **golden file harness mechanics**. The `assertGolden` helper in `golden_test.go` + `WriteGoldenFile` in `golden_canonical.go` are the established pattern. Plan 08-04 introduces a new `scorer_golden_test.go` (not merging into `algorithms_golden_test.go`) that calls `assertGolden(t, "scorer-default.json", scorerGoldenFile{...})`. The `-update` flag is already declared at the package level in `golden_test.go` and is shared across all `TestGolden_*` functions.

The **BDD harness** is fully operational at `tests/bdd/bdd_test.go` with `goleak.VerifyTestMain(m)` already wired. Adding `scorer.feature` to `tests/bdd/features/` is sufficient for godog to discover it; step definitions go in `tests/bdd/steps/`. The existing `AlgorithmContext` struct in `algorithms_steps.go` needs extension (or a new `ScorerContext` type) to hold `*Scorer`, last score, and last error state across BDD steps.

**Primary recommendation:** The planner should model the `scorerConfig` as a `[]scorerEntry` slice (not a map), where `WithoutAlgorithm` linearly scans and removes the matching entry. This matches the last-write-wins semantic naturally — later `WithAlgorithm(id, w)` calls overwrite earlier ones because `NewScorer` iterates the applied-options slice and merges by AlgoID. At `NewScorer` time, deduplicate to a map internally, then convert to a sorted slice — this gives O(1) dispatch during `Score` with zero allocation on the hot path.

---

## Architectural Responsibility Map

| Capability | Primary Tier | Secondary Tier | Rationale |
|------------|-------------|----------------|-----------|
| `NewScorer` construction + validation | Root package (scorer.go) | — | Pure function; no I/O; no goroutines per project constraint |
| Weight auto-normalisation | Root package (scorer.go, NewScorer) | — | One-time computation at construction; immutable after |
| `Score` composite reduction | Root package (scorer.go, Score method) | dispatch[] array (algoid.go) | Hot path; accesses dispatch table directly |
| `ScoreAll` per-algorithm breakdown | Root package (scorer.go, ScoreAll method) | — | Allocates fresh `map[AlgoID]float64` per call; documented |
| `Match` threshold comparison | Root package (scorer.go, Match method) | — | Delegates to `Score`; no additional allocation |
| Pre-normalisation step | Root package (normalise.go, Normalise function) | — | Already implemented; Scorer calls it at boundary |
| Parameterised algorithm options | Root package (scorer_options.go, With*Algorithm fns) | — | Closures capturing params; dispatch happens at Score time |
| Golden file gate | `scorer_golden_test.go` (new, plan 08-04) | `golden_test.go` (assertGolden infrastructure) | Separate file from algorithms_golden_test.go |
| BDD scenarios | `tests/bdd/features/scorer.feature` + `tests/bdd/steps/` | `tests/bdd/bdd_test.go` (goleak harness) | BDD sub-module isolation enforced by go.mod |
| Documentation | `docs/scorer.md`, `docs/tuning.md` (populated from scaffold) | `examples/scorer-composition/main.go` (new) | Plan 08-04 |

---

## Standard Stack

### Core (all in existing root go.mod — zero new deps needed)

| Library | Version | Purpose | Why Standard |
|---------|---------|---------|--------------|
| `encoding/json` | stdlib (Go 1.26.3) | Golden file serialisation via `canonicalMarshal` | Already used by `golden_canonical.go` |
| `sort` | stdlib (Go 1.26.3) | `sort.Slice` for AlgoID-sorted entry slice | Zero allocation on the hot path (sort is in-place on existing slice) |
| `errors` | stdlib (Go 1.26.3) | New sentinel errors (`ErrEmptyScorer`, etc.) | Established pattern in `errors.go` |
| `sync` | stdlib (Go 1.26.3) | `sync.WaitGroup` for concurrent safety test | Root-module-only; no `errgroup` (that is `golang.org/x/sync`) |

### Test Stack

| Library | Module | Purpose | Notes |
|---------|--------|---------|-------|
| `testing` | stdlib | Unit tests, property tests (`testing/quick`), benchmarks | Root module; no testify per spec-lock |
| `github.com/cucumber/godog` | `tests/bdd/go.mod` | BDD scenarios for scorer.feature | Already present |
| `go.uber.org/goleak` | `tests/bdd/go.mod` | Goroutine leak detection via `goleak.VerifyTestMain` | Already present; hooked in `bdd_test.go` |
| `github.com/stretchr/testify` | `tests/bdd/go.mod` | Step definition assertions | Permitted only in `tests/bdd/`; already present |

**No new dependencies required** — Phase 8 adds only files to the existing root module. The BDD sub-module's dependencies are already pinned.

---

## Architecture Patterns

### System Architecture Diagram

```
Consumer code
    │
    ▼
NewScorer(opts ...ScorerOption)  ←── With*Algorithm / WithThreshold /
    │                                 WithNormalisation / WithoutAlgorithm
    │ validation pipeline
    │  1. ErrMissingThreshold (first)
    │  2. ErrEmptyScorer
    │  3. ErrInvalidWeight
    │  4. ErrInvalidThreshold
    │  5. ErrInvalidAlgorithm (per entry)
    │  6. per-param validation (q-gram, Tversky)
    │  7. weight normalisation (if WithNormaliseWeights=true)
    │  8. sort entries by AlgoID → s.algorithmsAlgoIDSorted
    ▼
*Scorer (immutable, concurrent-safe)
    │
    ├──▶ Score(a, b) float64
    │       │
    │       ├─ if applyNormalisation: na=Normalise(a,opts), nb=Normalise(b,opts)
    │       │
    │       └─ for entry in algorithmsAlgoIDSorted:
    │               score := entry.scoreFn(na, nb)     ← dispatch[id] or closure
    │               acc = acc + (entry.weight * score)  ← DET-06 explicit parens
    │          return acc
    │
    ├──▶ Match(a, b) bool  →  Score(a,b) >= s.threshold
    │
    ├──▶ ScoreAll(a, b) map[AlgoID]float64
    │       │  (same normalisation + per-algo dispatch, fresh map per call)
    │       └─ documented: map iteration order is non-deterministic
    │
    ├──▶ Threshold() float64
    ├──▶ Algorithms() []ScorerAlgorithm   ← fresh slice, sorted by AlgoID
    │
DefaultScorer() *Scorer  →  6-algorithm composition per spec §8.5, threshold 0.85
DefaultScorerOptions() []ScorerOption  →  same composition as mutable option slice

                        ┌─────────────────────────────────────┐
                        │  dispatch[AlgoID] (algoid.go:324)   │
                        │  All 23 slots populated after Ph 7  │
                        │  dispatch[AlgoCosine](a,b) → n=3    │
                        │  dispatch[AlgoMongeElkan](a,b) →    │
                        │    MongeElkanScoreSymmetric(JW dflt) │
                        └─────────────────────────────────────┘
```

### Recommended File Structure

```
fuzzymatch/
├── scorer.go                   # NEW: Scorer struct, NewScorer, Score, Match methods
├── scorer_options.go           # NEW: ScorerOption type, scorerConfig, all With* functions
├── scorer_options_test.go      # NEW: unit tests for every option happy path + error path
├── scorer_test.go              # NEW: NewScorer, Score, Match unit + concurrent tests
├── scorer_internal_test.go     # NEW: scorerConfig unexported invariants (last-write-wins)
├── scorer_bench_test.go        # NEW: BenchmarkDefaultScorer_Score_ASCII_Short/Medium/Long
├── errors.go                   # EXTEND: add ErrEmptyScorer, ErrInvalidWeight, ErrInvalidThreshold, ErrMissingThreshold
├── algoid.go                   # READ-ONLY in Phase 8 (AlgoID, dispatch, AlgoIDs already complete)
├── testdata/golden/
│   └── scorer-default.json     # NEW in plan 08-04
├── tests/bdd/features/
│   └── scorer.feature          # NEW in plan 08-04
├── tests/bdd/steps/
│   └── algorithms_steps.go     # EXTEND: add ScorerContext + scorer step definitions
├── examples/identifier-similarity/main.go   # EXTEND: add Score+Match columns (plan 08-04)
└── examples/scorer-composition/main.go      # NEW in plan 08-04
```

### Pattern 1: `scorerEntry` Slice Design (Claude's Discretion)

**What:** A `[]scorerEntry` slice (not a map) stores the applied algorithms at option-accumulation time. `NewScorer` deduplicates and sorts the slice at construction time.

**When to use:** Always in Phase 8. Map would require iteration on the hot path; slice gives O(N) dispatch with zero allocation.

**Example:**

```go
// Source: derived from Phase 5 Cosine and Phase 6 ME dispatch patterns
type scorerEntry struct {
    id      AlgoID
    weight  float64
    scoreFn func(a, b string) float64
}

type scorerConfig struct {
    entries          []scorerEntry  // options accumulate here during construction
    threshold        float64
    thresholdSet     bool
    normaliseWeights bool
    applyNorm        bool
    normOpts         NormalisationOptions
}

// Inside NewScorer — dedup by last-write-wins, then sort:
seen := make(map[AlgoID]scorerEntry)
for _, e := range cfg.entries {
    seen[e.id] = e  // last write wins for duplicate AlgoIDs
}
// Convert to sorted slice for deterministic iteration at Score time:
sorted := make([]scorerEntry, 0, len(seen))
for _, id := range AlgoIDs() {  // canonical order from algoid.go:283-307
    if e, ok := seen[id]; ok {
        sorted = append(sorted, e)
    }
}
// sorted is now AlgoID-ascending (AlgoIDs() returns canonical order)
```

### Pattern 2: Float-Determinism Reduction Loop (LOCKED)

**What:** Explicit parenthesisation + sorted iteration. Carries forward from `cosine.go:343`.

**Source:** `cosine.go` lines 341-344 — the load-bearing determinism precedent.

```go
// Inside Scorer.Score (LOCKED form — see CONTEXT.md §5):
var acc float64
for _, entry := range s.algorithmsAlgoIDSorted {
    score := entry.scoreFn(na, nb)
    acc = acc + (entry.weight * score)  // explicit parens per DET-06
}
return acc
```

**Note:** `sort.Strings(intersectionKeys)` in `cosine.go` is needed there because Cosine iterates a map of q-gram strings. The Scorer equivalent is simpler: the slice is pre-sorted at `NewScorer` time — no per-call sort needed on the hot path.

### Pattern 3: Parameterised Closure Capture

**What:** `WithQGramJaccardAlgorithm(weight, n)` stores a closure that calls `QGramJaccardScore(a, b, n)` directly (ignoring `dispatch[AlgoQGramJaccard]`).

**Source:** `dispatch_cosine.go:42-44` shows the default-n-3 dispatch wrapper. Parameterised options bypass the dispatch table entirely.

```go
// Source: dispatch_cosine.go:42-44 (default wrapper pattern to invert)
func WithQGramJaccardAlgorithm(weight float64, n int) ScorerOption {
    return func(cfg *scorerConfig) error {
        if n < 1 {
            return ErrInvalidQGramSize
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

**For `WithMongeElkanAlgorithm(weight, inner)`:**

```go
func WithMongeElkanAlgorithm(weight float64, inner AlgoID) ScorerOption {
    return func(cfg *scorerConfig) error {
        // Validate inner is a permitted ME inner metric
        // (dispatch nil-check or explicit allow-list cross-reference)
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

### Pattern 4: dispatch[] Access for Non-Parameterised `WithAlgorithm`

**What:** `WithAlgorithm(id, weight)` stores `dispatch[id]` as the `scoreFn`. The dispatch slot is guaranteed populated at package-load time (all 23 `var _ = func() bool {}()` registrations fire before `NewScorer` can be called).

```go
// Source: algoid.go:324 — dispatch array declaration
// All slots populated after Phase 7 via dispatch_*.go var-init idiom

func WithAlgorithm(algo AlgoID, weight float64) ScorerOption {
    return func(cfg *scorerConfig) error {
        if weight <= 0 {
            return ErrInvalidWeight
        }
        // Bounds check for out-of-range AlgoID:
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

### Pattern 5: Golden File Harness Mechanics

**What:** The `assertGolden(t, filename, v)` helper in `golden_test.go:66` is the standard entry point. Plan 08-04 introduces `scorer_golden_test.go` with `TestGolden_ScorerDefault` that calls `assertGolden(t, "scorer-default.json", scorerGoldenFile{...})`.

**Source:** `golden_test.go:66-88` — `assertGolden` helper; `golden_canonical.go:61-73` — `canonicalMarshal`.

**The schema for `scorer-default.json` (from CONTEXT.md §6):**

```go
type scorerGoldenEntry struct {
    A          string             `json:"a"`
    B          string             `json:"b"`
    Score      float64            `json:"score"`
    Match      bool               `json:"match"`
    ScoreAll   map[string]float64 `json:"scoreAll"`  // string keys = AlgoID.String()
    ScorerConf string             `json:"scorer_config"`
}

type scorerGoldenFile struct {
    Metadata struct {
        Phase           int    `json:"phase"`
        GeneratedAt     string `json:"generated_at"`
        ScorerSignature string `json:"scorer_signature"`
    } `json:"_metadata"`
    Entries []scorerGoldenEntry `json:"entries"`
}
```

**Critical:** `scoreAll` in the JSON uses `AlgoID.String()` as string keys (not the typed `AlgoID` — JSON keys must be strings). Internally, `ScoreAll` returns `map[AlgoID]float64`; the golden serialisation converts to `map[string]float64` for the JSON fixture only.

**Regen command:** `go test -run TestGolden_ -update ./...`

### Pattern 6: BDD Context Extension

**What:** The existing `AlgorithmContext` in `tests/bdd/steps/algorithms_steps.go` needs a `ScorerContext` state block, or a new separate context type. Given Phase 8 is the first heavy Scorer BDD use, the planner should introduce a `ScorerContext` struct alongside `AlgorithmContext` and register its steps in `InitializeScenario`.

**Source:** `tests/bdd/steps/algorithms_steps.go:43-63` — `AlgorithmContext` struct pattern; `tests/bdd/bdd_test.go:37-39` — `goleak.VerifyTestMain` hook (already wired; Phase 8 adds no new goroutines so goleak should pass trivially).

```go
// In tests/bdd/steps/ (new or extended file):
type ScorerContext struct {
    scorer    *fuzzymatch.Scorer
    lastScore float64
    lastMatch bool
    lastErr   error
    scoreAll  map[fuzzymatch.AlgoID]float64
}
```

### Anti-Patterns to Avoid

- **Map iteration in Score's reduction loop:** the `s.algorithmsAlgoIDSorted` slice is pre-sorted at `NewScorer` time; never iterate a `map[AlgoID]scorerEntry` on the Score hot path.
- **Per-call weight normalisation:** weights are normalised once at `NewScorer` time; `Score` should never divide or accumulate sums — just use stored normalised weights.
- **`sync.Mutex` in `*Scorer`:** the struct is immutable after construction; no mutex needed. Confirmed: concurrent safety comes from read-only access to fields set at construction time.
- **Calling `sort.Slice` inside `Score`:** AlgoID-sorted slice is built once in `NewScorer`; calling sort again on each `Score` invocation is a ~30-40ns per-call regression on the < 30µs budget.
- **`init()` for dispatch population:** all 23 dispatch slots use `var _ = func() bool {}()` idiom. Phase 8 must not introduce any `init()`.
- **testify in root test files:** `scorer_test.go`, `scorer_options_test.go`, `scorer_internal_test.go`, `scorer_bench_test.go` are root module files — stdlib `testing` only.
- **errgroup in concurrent test:** `golang.org/x/sync/errgroup` is not in root `go.mod`. Use `sync.WaitGroup` instead.

---

## Don't Hand-Roll

| Problem | Don't Build | Use Instead | Why |
|---------|-------------|-------------|-----|
| Float determinism across platforms | Custom float-rounding or interval arithmetic | `acc = acc + (entry.weight * score)` explicit-parens pattern from `cosine.go:343` | Same code already passing the CI matrix golden gate |
| JSON golden file serialisation | Custom JSON builder | `canonicalMarshal` / `WriteGoldenFile` / `assertGolden` already in `golden_canonical.go` + `golden_test.go` | Canonical form locked at Phase 1 D-12 |
| AlgoID canonical ordering | Custom sort | `AlgoIDs()` returns canonical slice; use it as dedup-merge iteration order | Spec stability contract; adds new algorithm = new entry at end only |
| BDD goleak wiring | New TestMain | `tests/bdd/bdd_test.go:37-39` already wires `goleak.VerifyTestMain(m)` | Already present; adding `scorer.feature` is sufficient |
| Dispatch table bounds check | Ad-hoc `if id > 22` | `int(algo) >= numAlgorithms || dispatch[algo] == nil` | `numAlgorithms` is the compile-time constant from `algoid.go:187` |
| Concurrent-safe Scorer | `sync.RWMutex` | Immutable struct — zero synchronisation needed | Construction sets all fields; `Score` / `ScoreAll` / `Match` are read-only |

**Key insight:** Phase 8 is almost entirely an integration exercise — the hard primitives (dispatch table, Normalise, golden harness, BDD runner, float-determinism patterns) are already in the codebase. The Scorer is new wiring code, not new algorithmic complexity.

---

## Common Pitfalls

### Pitfall 1: FMA Fusion in the Composite Reduction Loop

**What goes wrong:** On arm64, the Go compiler may fuse `a + (b*c)` into a single FMA instruction, which changes the rounding. The parenthesisation `acc + (w*s)` does NOT prevent FMA on arm64 per `golang/go#17895`.

**Why it happens:** Go 1.26 on arm64 emits FMA for patterns the compiler identifies as fuseable. Unlike IEEE-754's required "faithfully rounded" constraint, FMA changes the intermediate rounding.

**How to avoid:** The `cosine.go` godoc (lines 288-297) documents that the empirical observation is that integer-derived float products in the current algorithms are small enough that FMA-vs-non-FMA divergence falls below the byte-diff threshold of the golden gate. The Scorer faces the same situation: algorithm scores in [0, 1] × weights in [0, 1] are small enough to be safe empirically. If `scorer-default.json` ever diverges across CI matrix platforms, insert an explicit double cast: `score := float64(float64(entry.weight) * float64(rawScore))`.

**Warning signs:** `scorer-default.json` byte-mismatch in CI between linux/amd64 and linux/arm64 or darwin/arm64 entries.

**Source:** `cosine.go:288-297` [VERIFIED: source file read]

### Pitfall 2: Map Iteration in ScoreAll Building

**What goes wrong:** If `ScoreAll` builds the result map then ranges over an intermediate map to populate it, Go's map iteration order is randomised. The map contents are correct but the order of operations affects floating-point accumulation in the composite score (if ScoreAll were to re-sum — but it shouldn't). The real risk is the test asserting specific map contents: always use `result[fuzzymatch.AlgoLevenshtein]` not range-based iteration in tests.

**Why it happens:** Go maps are intentionally non-deterministic since Go 1.0.

**How to avoid:** `ScoreAll` iterates `s.algorithmsAlgoIDSorted` (slice, deterministic) to populate the map; it never iterates the map itself. Tests use direct key access or sort the keys before asserting.

**Warning signs:** `PropScorer_ScoreAllValues_Deterministic` failing intermittently.

### Pitfall 3: `ErrMissingThreshold` Check Order

**What goes wrong:** If the empty-scorer check fires before the missing-threshold check, a user who writes `NewScorer(WithAlgorithm(AlgoLevenshtein, 1.0))` (threshold missing) gets `ErrEmptyScorer` instead of `ErrMissingThreshold`, which is incorrect and confusing.

**Why it happens:** The natural coding order is "check if there are any algorithms first."

**How to avoid:** The validation pipeline MUST check `!cfg.thresholdSet` first and return `ErrMissingThreshold` before any other validation. Per CONTEXT.md §2 Action item: "gates on the missing-threshold check FIRST."

**Warning signs:** A unit test `NewScorer(WithAlgorithm(AlgoLevenshtein, 1.0))` returning `ErrEmptyScorer`.

### Pitfall 4: `WithoutAlgorithm` No-Op on Absent ID

**What goes wrong:** `WithoutAlgorithm(AlgoDoubleMetaphone)` called on a `scorerConfig` that doesn't contain `AlgoDoubleMetaphone` must silently no-op. If the implementation panics or returns an error, `append(DefaultScorerOptions(), WithoutAlgorithm(AlgoDM))` breaks for any DefaultScorer composition that excludes a non-default algorithm.

**Why it happens:** Linear scan finds no entry; a naive implementation returns an error.

**How to avoid:** `WithoutAlgorithm` scans `cfg.entries` for the matching `id`; if not found, returns `nil` (no-op). Specifically: `for i, e := range cfg.entries { if e.id == id { cfg.entries = append(cfg.entries[:i], cfg.entries[i+1:]...) } }` — remove-by-index without returning error.

**Warning signs:** BDD scenario 5 (`WithoutAlgorithm` composition) returning an error.

### Pitfall 5: Double Weight Normalisation

**What goes wrong:** If `DefaultScorerOptions()` returns options with pre-normalised weights (already summing to 1.0) and `NewScorer` normalises again, the weights are incorrect (each weight gets divided by 1.0 again, which is a no-op numerically — actually safe — but if the weights sum to slightly less than 1.0 due to float rounding, renormalisation amplifies the drift).

**Why it happens:** `DefaultScorerOptions()` returns the SAME option functions as a user calling `WithAlgorithm(AlgoX, rawWeight)`. Weights are raw (e.g. `1.0` each for equal-weight composition). `NewScorer` normalises them. Normalising already-normalised weights is harmless mathematically but is still the correct flow.

**How to avoid:** `DefaultScorerOptions()` returns raw (unnormalised) weights. `NewScorer` always normalises when `WithNormaliseWeights(true)`. This is correct. The pitfall is adding normalisation inside `DefaultScorerOptions()` itself — don't.

### Pitfall 6: `scorer-default.json` ScoreAll Key Format

**What goes wrong:** The golden file's `scoreAll` uses `AlgoID.String()` as map keys (e.g. `"DamerauLevenshteinOSA"`) not the AlgoID integer values. If the `scorerGoldenEntry` struct's `ScoreAll` field is `map[AlgoID]float64`, `encoding/json` will try to marshal integer keys as JSON object keys — which works in Go 1.26 (`json.Marshal` converts int keys to string), but produces `"0"`, `"1"`, ... not the readable names.

**How to avoid:** The golden file serialisation struct uses `map[string]float64` for `ScoreAll` (converting from `map[AlgoID]float64` via a for-range over `AlgoIDs()` + `id.String()`). The public API still returns `map[AlgoID]float64` — only the golden file uses string keys for human readability.

**Warning signs:** `scorer-default.json` containing `"0": 1.0, "1": 0.98...` instead of `"Levenshtein": 1.0, "DamerauLevenshteinOSA": 0.98...`.

---

## Code Examples

Verified patterns from codebase source files:

### dispatch[] Access Pattern (non-parameterised algorithms)

```go
// Source: algoid.go:324 — dispatch array declaration
// Source: dispatch_cosine.go:41-46 — canonical var-init population idiom

// Read-only access in Scorer hot path — zero allocation:
score := dispatch[entry.id](na, nb)
// Bounds guard (use at NewScorer validation time only):
if int(algo) >= numAlgorithms || dispatch[algo] == nil {
    return ErrInvalidAlgorithm
}
```

### Canonical Sorted Iteration

```go
// Source: algoid.go:282-308 — AlgoIDs() returns stable canonical order
// Use at NewScorer dedup step:
for _, id := range AlgoIDs() {
    if e, ok := seen[id]; ok {
        sortedEntries = append(sortedEntries, e)
    }
}
// sortedEntries is now AlgoID-ascending per canonical iota order.
```

### Float-Determinism Reduction

```go
// Source: cosine.go:341-344 — load-bearing precedent
var acc float64
for _, entry := range s.algorithmsAlgoIDSorted {
    score := entry.scoreFn(na, nb)
    acc = acc + (entry.weight * score)  // DET-06 explicit parens
}
```

### Sentinel Error Declaration (extend errors.go)

```go
// Source: errors.go:48-106 — established godoc + errors.New pattern
// Phase 8 appends these after ErrEmptyInput:

var ErrEmptyScorer = errors.New("fuzzymatch: scorer has no algorithms (pass at least one WithAlgorithm option or use DefaultScorer)")

var ErrInvalidWeight = errors.New("fuzzymatch: invalid algorithm weight (must be > 0)")

var ErrInvalidThreshold = errors.New("fuzzymatch: invalid threshold (must be in [0.0, 1.0])")

var ErrMissingThreshold = errors.New("fuzzymatch: scorer threshold required (pass WithThreshold or use DefaultScorer)")
```

### assertGolden Usage

```go
// Source: golden_test.go:66-88 — assertGolden helper
// Plan 08-04 adds scorer_golden_test.go with:
func TestGolden_ScorerDefault(t *testing.T) {
    s := fuzzymatch.DefaultScorer()
    entries := buildScorerGoldenEntries(t, s)  // helper builds all 22-26 rows
    assertGolden(t, "scorer-default.json", scorerGoldenFile{
        Metadata: scorerGoldenMetadata{
            Phase:           8,
            GeneratedAt:     "<filled by -update pass>",
            ScorerSignature: "DefaultScorer-2026-05-16",
        },
        Entries: entries,
    })
}
```

### BDD goleak — already wired

```go
// Source: tests/bdd/bdd_test.go:37-39 — existing TestMain
func TestMain(m *testing.M) {
    goleak.VerifyTestMain(m)  // already present; Phase 8 adds scorer.feature only
}
```

---

## State of the Art

| Old Approach | Current Approach | When Changed | Impact |
|--------------|------------------|--------------|--------|
| `dispatch[id]` not nil-checked in Scorer | `int(id) >= numAlgorithms \|\| dispatch[id] == nil` gate in `NewScorer` | Phase 8 (plan 08-01) | Prevents panic on AlgoID(999) input |
| No composite Scorer (Layer 2 absent) | `NewScorer` + `DefaultScorer` with 6-algorithm mix | Phase 8 | Layer 2 complete |
| `map[string]float64` for ScoreAll (spec §8.3) | `map[AlgoID]float64` (SPEC OVERRIDE) | Phase 8 §1 | Typed keys; api-ergonomics-reviewer sign-off required in plan 08-03 PR |
| ME's `opts NormalisationOptions` unused | Accepted-but-ignored (`_ = opts`) per Phase 6 | Phase 6 | Phase 8 confirms no change; deferred to Phase 11 |

**Deprecated/outdated:**
- `docs/requirements.md §8.3's `map[string]float64``: superseded by CONTEXT.md §1 SPEC OVERRIDE. Will be updated in plan 08-04's docs commit.
- `REQUIREMENTS.md SCORER-08's `WithCustomNormalisation()``: typo; corrected to `WithNormalisation(opts)` per spec §8.2. Will be fixed in plan 08-04.

---

## Runtime State Inventory

> Not applicable — Phase 8 is greenfield feature code added to an existing library. No rename, refactor, or migration. No stored data, live service config, OS-registered state, secrets, or stale build artifacts involved.

---

## Open Questions

1. **`scorer-default.json` — GeneratedAt field stability**
   - What we know: The `_metadata.generated_at` field is an ISO timestamp written at `-update` time. If two platforms run `-update` at different wall-clock times, `generated_at` will differ, breaking byte-equality.
   - What's unclear: Should `generated_at` be omitted from the golden file comparison, or should it be a fixed string?
   - Recommendation: Either (a) exclude `_metadata` from the JSON comparison (compare only `entries`), or (b) use a fixed sentinel value like `"GOLDEN"` for the timestamp at assertion time (injected by the test, not from `time.Now()`). Option (b) is simpler. The planner should pick one approach consistently with how Phase 5's `algorithms.json` handles any timestamp fields. **Check: `algorithms.json` has no timestamp field** — the simpler schema has no `_metadata.generated_at`. Recommendation: omit `generated_at` from the fixture or use a fixed string `"PINNED"` that never changes.

2. **`WithSmithWatermanGotohAlgorithm` — `SWGParams` type location**
   - What we know: The parameterised SWG option takes `SWGParams` (gap-open and gap-extend penalties). This type was defined in Phase 3.
   - What's unclear: Whether `SWGParams` is the exact exported type name.
   - Recommendation: Before writing `scorer_options.go`, grep for `SWGParams` in `swg.go` to confirm the exact type name. [ASSUMED — not verified in this session; low risk as Phase 3 is complete.]

3. **BDD step definitions — new file vs extend `algorithms_steps.go`**
   - What we know: `algorithms_steps.go` contains `AlgorithmContext` and all 23 algorithm step definitions. Adding Scorer steps there would make the file very large.
   - What's unclear: Whether godog requires all steps in one file or supports split step files.
   - Recommendation: Introduce `scorer_steps.go` in `tests/bdd/steps/`. Godog's `ScenarioInitializer` accepts multiple registration functions; `InitializeScenario` can call both `algorithms_steps.go:InitAlgorithmSteps` and `scorer_steps.go:InitScorerSteps`. [ASSUMED: godog supports split step registration — confirmed by design pattern, not verified against godog v0.15.0 docs.]

---

## Environment Availability

> Step 2.6 SKIPPED — Phase 8 is purely code/config additions to an existing Go module. No new external tools, services, CLIs, runtimes, databases, or package managers are introduced. All dependencies (`encoding/json`, `sort`, `errors`, `sync`, `testing/quick`) are stdlib. BDD dependencies (godog, goleak, testify) are already present in `tests/bdd/go.mod`.

---

## Validation Architecture

### Test Framework

| Property | Value |
|----------|-------|
| Framework | Go stdlib `testing` (root), godog v0.15.0 (BDD sub-module) |
| Config file | None (root); `tests/bdd/go.mod` (BDD) |
| Quick run command | `go test ./... -run TestScorer` |
| Full suite command | `go test ./... && cd tests/bdd && go test ./...` |
| Race detector | `go test -race ./...` |
| Golden gate | `go test -run TestGolden_ ./...` |
| BDD | `cd tests/bdd && go test ./...` |

### Phase Requirements → Test Map

| Req ID | Behavior | Test Type | Automated Command | File |
|--------|----------|-----------|-------------------|------|
| SCORER-01 | `NewScorer` validates and returns immutable `*Scorer` | unit + property | `go test -run TestNewScorer ./...` | `scorer_test.go` Wave 08-02 |
| SCORER-01 | Concurrent `Score`/`ScoreAll`/`Match` on same `*Scorer` | concurrent | `go test -race -run TestScorer_ConcurrentSafety ./...` | `scorer_test.go` Wave 08-03 |
| SCORER-02 | `DefaultScorer()` matches spec §8.5 composition | unit | `go test -run TestDefaultScorer ./...` | `scorer_test.go` Wave 08-03 |
| SCORER-02 | `DefaultScorerOptions()` produces same Scorer as `DefaultScorer()` | unit | `go test -run TestDefaultScorerOptions ./...` | `scorer_test.go` Wave 08-03 |
| SCORER-03 | Weight auto-normalisation sums to 1.0 | property (`testing/quick`) | `go test -run TestProp_WeightSumOne ./...` | `scorer_test.go` Wave 08-03 |
| SCORER-03 | Last-write-wins for duplicate AlgoID | unit | `go test -run TestScorer_LastWriteWins ./...` | `scorer_internal_test.go` Wave 08-02 |
| SCORER-04 | `Score` in [0,1] with normalised weights | property (`testing/quick`) | `go test -run TestProp_ScoreInRange ./...` | `scorer_test.go` Wave 08-03 |
| SCORER-04 | `PropScorer_DeterministicAcrossRuns` | property | `go test -run TestProp_DeterministicAcrossRuns ./...` | `scorer_test.go` Wave 08-03 |
| SCORER-04 | `scorer-default.json` byte-identical across CI matrix | golden | `go test -run TestGolden_ScorerDefault ./...` | `scorer_golden_test.go` Wave 08-04 |
| SCORER-05 | `ScoreAll` returns correct per-algorithm scores | unit | `go test -run TestScoreAll ./...` | `scorer_test.go` Wave 08-03 |
| SCORER-06 | `Match` returns true iff `Score >= Threshold` | unit | `go test -run TestMatch ./...` | `scorer_test.go` Wave 08-02 |
| SCORER-07 | `Algorithms()` returns fresh sorted slice | unit | `go test -run TestAlgorithms ./...` | `scorer_test.go` Wave 08-03 |
| SCORER-08 | `WithoutNormalisation` disables pre-normalisation | unit + BDD | `go test -run TestScorer_WithoutNormalisation ./...` | `scorer_test.go` Wave 08-02 |
| All | BDD scenarios (8-12) in `scorer.feature` | BDD (godog) | `cd tests/bdd && go test ./...` | `scorer.feature` Wave 08-04 |
| All | `goleak.VerifyTestMain` confirms no goroutine leaks | goleak (BDD) | `cd tests/bdd && go test ./...` | `tests/bdd/bdd_test.go` (already wired) |
| All | Allocation budget ≤ 8 allocs on ASCII ≤ 50 chars | benchmark | `go test -bench=BenchmarkDefaultScorer_Score -benchmem ./...` | `scorer_bench_test.go` Wave 08-04 |

### Sampling Rate

- **Per task commit:** `go test -run TestScorer -race ./...`
- **Per wave merge:** `go test ./... -race && cd tests/bdd && go test ./...`
- **Phase gate:** `make check` (full quality gate including golangci-lint, golden files, coverage) before `/gsd-verify-work`

### Wave 0 Gaps

- [ ] `scorer_test.go` — covers SCORER-01, SCORER-03, SCORER-04, SCORER-06 (Wave 08-02)
- [ ] `scorer_options_test.go` — covers all option happy + error paths (Wave 08-01)
- [ ] `scorer_internal_test.go` — covers last-write-wins `scorerConfig` invariants (Wave 08-02)
- [ ] `scorer_bench_test.go` — covers < 30µs / ≤ 8 allocs budget (Wave 08-04)
- [ ] `scorer_golden_test.go` — covers cross-platform determinism via `scorer-default.json` (Wave 08-04)
- [ ] `tests/bdd/features/scorer.feature` — covers 8-12 BDD scenarios (Wave 08-04)
- [ ] `tests/bdd/steps/scorer_steps.go` — step definitions for scorer.feature (Wave 08-04)

*(No new test framework installation needed — all frameworks are already present.)*

---

## Security Domain

> `security_enforcement` is not explicitly `false` in `.planning/config.json` — treating as enabled.

### Applicable ASVS Categories

| ASVS Category | Applies | Standard Control |
|---------------|---------|-----------------|
| V2 Authentication | No | Library has no auth surface |
| V3 Session Management | No | Pure-function library; no sessions |
| V4 Access Control | No | No access control concepts |
| V5 Input Validation | Yes | Algorithmic inputs validated in `NewScorer` via `ErrInvalid*` sentinels; `Normalise` handles invalid UTF-8 via `U+FFFD` replacement (never panics) |
| V6 Cryptography | No | No cryptographic operations |

### Known Threat Patterns for fuzzymatch Scorer

| Pattern | STRIDE | Standard Mitigation |
|---------|--------|---------------------|
| Adversarial regex / algorithmic complexity DoS via pathological string input | DoS | Scorer calls pure-function algorithms with guaranteed linear/polynomial worst-case complexity (RE2-safe by design); spec §14 enforces per-algorithm time budgets; fuzz tests assert no panics on arbitrary input |
| Weight parameter manipulation producing non-finite scores (NaN/Inf from 0-weight normalisation) | Tampering | `ErrInvalidWeight` rejects weight ≤ 0; normalisation divides by sum-of-positive-weights (guaranteed > 0 after validation); `PropScorer_ScoreInRange` property test asserts [0,1] bounds |
| Panic on out-of-range AlgoID (AlgoID(999)) | DoS | `NewScorer` validation gates on `int(algo) >= numAlgorithms || dispatch[algo] == nil` and returns `ErrInvalidAlgorithm` instead of panicking; direct algorithm functions panic loudly (programmer error) per CONTEXT.md §5 locked |
| BDD/test goroutine leaks introducing Scheduler contention in production | DoS | `goleak.VerifyTestMain(m)` in BDD runner; library has no goroutines (confirmed: no `go` statements in root module) |

---

## Assumptions Log

| # | Claim | Section | Risk if Wrong |
|---|-------|---------|---------------|
| A1 | `SWGParams` is the exact exported type name for Smith-Waterman-Gotoh parameters (from Phase 3) | Standard Stack / Pattern 3 | Low — only affects option function signature; easily corrected by reading `swg.go` before writing `scorer_options.go` |
| A2 | godog v0.15.0 supports split step registration across multiple `*.go` files in `tests/bdd/steps/` | Common Pitfalls / Pitfall 6 / BDD section | Low — godog's `ScenarioInitializer` is a single function but can call multiple sub-registrations from multiple files; this is standard Go |
| A3 | Phase 7 populated all 4 phonetic dispatch slots (Soundex, DoubleMetaphone, NYSIIS, MRA) at `dispatch[AlgoSoundex]` etc. | Standard Stack (dispatch table) | Medium — if any phonetic dispatch slot is still nil, `WithAlgorithm(AlgoSoundex, w)` would return `ErrInvalidAlgorithm`. Verify with `grep -r "dispatch\[AlgoSoundex\]" .` before plan 08-02. Per ROADMAP.md Phase 7 is marked completed (✓) so this is expected to be safe. |

**If this table is empty:** All claims in this research were verified or cited — no user confirmation needed.

The three assumptions above are low-to-medium risk. A1 and A2 are easily verified during implementation. A3 has a direct verification command in the note.

---

## Project Constraints (from CLAUDE.md)

The following directives from `./CLAUDE.md` are actionable for Phase 8:

| Directive | Implication for Phase 8 |
|-----------|------------------------|
| Zero runtime deps in root `go.mod` | `scorer.go` and `scorer_options.go` must use stdlib only; no new `require` lines |
| No cgo anywhere | Confirmed: no native bindings in Scorer |
| No testify in root test files | `scorer_test.go`, `scorer_options_test.go`, `scorer_internal_test.go` use stdlib `testing` only |
| Conventional commits with issue references | Every commit: `feat(scorer): <description> (#<issue>)` |
| No goroutines in library code | `Scorer.Score` / `ScoreAll` / `Match` are synchronous; no goroutines |
| No map iteration on output paths | `ScoreAll` iterates `s.algorithmsAlgoIDSorted` (slice) to populate the map; never iterates the map to compute the composite score |
| No `math.Pow`, `math.Log`, `math.Exp`, `math.FMA` | The `(w*s)+acc` reduction uses only multiplication and addition |
| No `init()` functions | `scorer.go` and `scorer_options.go` have no `init()`; all dispatch registration is already in `dispatch_*.go` via `var _ = func() bool {}()` |
| api-ergonomics-reviewer gates every API surface | Plan 08-01 PR: option shapes. Plan 08-03 PR: ScoreAll SPEC OVERRIDE sign-off (explicit required). |
| commit-message-reviewer before every commit | Applies to all 4 plan PRs |
| algorithm-correctness-reviewer on composite reduction | Plan 08-02 PR specifically |
| determinism-reviewer on golden file | Plan 08-04 PR |
| Releases via CI only | Not directly applicable in Phase 8 (no release) |

---

## Sources

### Primary (HIGH confidence)
- `08-CONTEXT.md` (2026-05-16) — 8 locked decisions, phase domain description, all implementation patterns [VERIFIED: file read]
- `algoid.go` lines 1-325 — AlgoID enum, `dispatch[]` array, `AlgoIDs()` canonical order, `numAlgorithms` constant [VERIFIED: file read]
- `errors.go` lines 1-107 — existing sentinel errors, godoc pattern, `errors.New` form [VERIFIED: file read]
- `normalise.go` lines 1-453 — `Normalise`, `NormalisationOptions`, `DefaultNormalisationOptions` [VERIFIED: file read]
- `cosine.go` lines 288-360 — FMA risk documentation, `(x*y)+z` explicit-parens pattern, sorted-intersection-keys reduction [VERIFIED: file read]
- `dispatch_cosine.go` — default-n-3 dispatch wrapper pattern [VERIFIED: file read]
- `dispatch_monge_elkan.go` — symmetric variant dispatch, JaroWinkler default inner [VERIFIED: file read]
- `golden_canonical.go` — `canonicalMarshal`, `WriteGoldenFile` implementation [VERIFIED: file read]
- `golden_test.go` — `assertGolden` helper, `-update` flag, `updateGolden` package-level flag [VERIFIED: file read]
- `algorithms_golden_test.go` lines 1-100 — `assertGoldenStaging`, `goldenAlgorithmsFile` schema [VERIFIED: file read]
- `monge_elkan.go` lines 377-406 — vestigial `opts` parameter (`_ = opts`), identity short-circuit, dispatch table usage [VERIFIED: file read]
- `tests/bdd/bdd_test.go` — `goleak.VerifyTestMain(m)` hook, `godog.TestSuite` wiring [VERIFIED: file read]
- `tests/bdd/steps/algorithms_steps.go` (header) — `AlgorithmContext` struct pattern [VERIFIED: file read]
- `.planning/REQUIREMENTS.md` lines 59-68 — SCORER-01..08 with `map[AlgoID]float64` on SCORER-05 [VERIFIED: file read]
- `.claude/skills/determinism-standards/SKILL.md` — no-map-iteration rule, float stability, golden file discipline [VERIFIED: file read]
- `.claude/skills/performance-standards/SKILL.md` — Scorer budgets (< 30µs / ≤ 8 allocs), benchmark file pattern [VERIFIED: file read]

### Secondary (MEDIUM confidence)
- `.claude/skills/go-testing-standards/SKILL.md` — BDD step patterns, coverage targets, `testing/quick` property tests [VERIFIED: file read]
- `.planning/ROADMAP.md` Phase 8 section — plan structure, success criteria [VERIFIED: file read]

### Tertiary (LOW confidence)
- godog v0.15.0 multi-file step registration behaviour — standard Go but not directly verified against godog docs this session [ASSUMED]

---

## Metadata

**Confidence breakdown:**
- Standard stack: HIGH — all existing libraries; no new deps; verified from source files
- Architecture patterns: HIGH — derived directly from `algoid.go`, `cosine.go`, `dispatch_*.go`, `golden_test.go`, `bdd_test.go`
- Pitfalls: HIGH — most are documented in existing source file godoc comments with line references
- Validation architecture: HIGH — test framework, flags, and harness all verified from existing files

**Research date:** 2026-05-16
**Valid until:** 2026-06-16 (30 days — stable library with locked decisions)
