---
phase: 08-composite-scorer
plan: 08-02
subsystem: scorer
tags: [scorer, validation, float-determinism, immutable, concurrent-safe, phase-8]

# Dependency graph
requires:
  - phase: 08-composite-scorer
    plan: 08-01
    provides: ScorerOption, scorerEntry, scorerConfig, 12 With* options, 4 new sentinel errors, applyOptionForProbe
  - phase: 01-foundation-infrastructure
    provides: AlgoID enum, AlgoIDs() canonical order, dispatch[] array, numAlgorithms, Normalise / NormalisationOptions / DefaultNormalisationOptions
  - phase: 02-character-algorithms
    provides: LevenshteinScore (used as the canonical reference in external tests), JaroWinkler (used in the AlgoID-sorting test)
  - phase: 05-q-gram-algorithms
    provides: CONTEXT.md §5 / cosine.go:341-344 float-determinism reduction precedent
provides:
  - Scorer struct (immutable, concurrent-safe after NewScorer)
  - NewScorer(opts ...ScorerOption) (*Scorer, error) — LOCKED 9-step validation pipeline
  - (*Scorer).Score(a, b string) float64 — pre-normalisation + AlgoID-sorted reduction
  - (*Scorer).Match(a, b string) bool — Score >= threshold (boundary inclusive)
  - Last-write-wins dedup invariant for duplicate WithAlgorithm(SameID, _)
  - Weight auto-normalisation (sum-to-1) invariant
  - AlgoID-ascending internal entry order invariant
affects: 08-03-ScoreAll-Threshold-Algorithms-DefaultScorer, 08-04-golden-BDD-docs

# Tech tracking
tech-stack:
  added: []  # zero new runtime deps — stdlib only, no new imports vs plan 08-01
  patterns:
    - "Immutable type with unexported fields set once in constructor — concurrent-safe without locks (no sync.Mutex, no sync/atomic)"
    - "Validation pipeline with LOCKED ordering (CONTEXT.md §2): missing-threshold FIRST → empty → defensive AlgoID bounds → dedup → normalise"
    - "Last-write-wins dedup via map[AlgoID]scorerEntry merge pass (08-RESEARCH.md Pattern 1)"
    - "AlgoIDs() canonical-order materialisation for deterministic iteration"
    - "Float-determinism reduction loop carried forward verbatim from cosine.go:341-344 (DET-06 explicit parens + left-to-right + sorted iteration)"
    - "Pre-normalisation single conditional + two Normalise calls at the Scorer boundary (CONTEXT.md §3)"
    - "Defensive nolint:gocritic on the three explicit-arithmetic determinism lines with DET-06 rationale (assignOp would obscure the locked contract)"

key-files:
  created:
    - scorer.go (388 lines — Scorer struct, NewScorer, Score, Match)
    - scorer_internal_test.go (185 lines — package fuzzymatch internal invariant tests)
    - scorer_test.go (305 lines — package fuzzymatch_test external happy + error paths)
  modified:
    - llms.txt (Scorer + NewScorer + Score + Match added under the Scorer construction section)
    - llms-full.txt (full godoc-summary block for Scorer + NewScorer + Score + Match appended before the scorer_options section)

key-decisions:
  - "scorer.go imports nothing from stdlib. The plan's acceptance criterion bounded imports to 'sort only'; the actual implementation uses AlgoIDs() (which already returns canonical order) so no sort call is needed. Empty import set is strictly within the plan bound. Plan 08-03's Algorithms() accessor (which sorts a fresh slice copy) will be the import that introduces 'sort'."
  - "//nolint:gocritic on the three reduction-loop arithmetic lines (sum = sum + …, weight = weight / sum, acc = acc + (entry.weight * score)) with explicit DET-06 / CONTEXT.md §5 rationale. The assignOp shorthand (+=, /=) is observationally equivalent but would obscure the determinism contract that the plan's verification grep explicitly tests for (`grep -F \"(entry.weight * score)\" scorer.go` and `grep -F \"acc = acc +\" scorer.go`)."
  - "//nolint:gocyclo on NewScorer with rationale citing the 9-step LOCKED pipeline. Splitting into sub-helpers would scatter the locked-order contract across files; the linear in-function form is the contract."
  - "scorerConfig defaults (normaliseWeights=true, applyNorm=true, normOpts=DefaultNormalisationOptions()) are initialised in the FIRST line of NewScorer, BEFORE applying options. This guarantees that options that don't touch normalisation (WithAlgorithm, WithThreshold, …) produce the documented default behaviour per spec §9.4. The plan's Step 1 specifies this; recording for forward-compat."
  - "Defensive per-entry AlgoID bounds + dispatch nil-check (Step 5) is defence-in-depth. The option layer already gates on the same condition, but re-checking here guarantees the final *Scorer's invariants hold regardless of how the option slice was assembled (e.g. a future caller building options via reflection or a config-file decoder)."
  - "Single-entry post-normalisation weight is exactly 1.0 because the weight is divided by itself; verified by TestScorer_LastWriteWins."
  - "2-algorithm normalised weights (1.0, 3.0) → (0.25, 0.75) sum to exactly 1.0 with == (no tolerance). 0.25 and 0.75 are dyadic fractions and exactly representable in IEEE-754. Recording the EXACT-1.0 assertion choice (vs tolerance) for the next plan's property tests to extend."

patterns-established:
  - "Pattern 1: Immutable struct + read-only methods for concurrent-safe APIs. No sync.Mutex; all fields set once in the constructor. Verified by the deterministic-across-calls external test (1000 calls; identical float64 each time)."
  - "Pattern 2: Float-determinism reduction loop reused at Scorer scope. The cosine.go:343 precedent is now the project's canonical shape for any per-element weighted accumulation; future composites (Extract / scan) should follow the same `acc = acc + (entry.weight * score)` form."
  - "Pattern 3: LOCKED validation pipeline gate ordering. Missing-threshold FIRST is now precedent for any future construction-time validator (e.g. plan 08-04's example program; phase 9 scan options). The disambiguation rationale (clear diagnostic on multi-error inputs) generalises."
  - "Pattern 4: //nolint:gocritic with DET-06 rationale on explicit-arithmetic determinism lines. Establishes the project's stance: locked-pattern comments are silenced explicitly with rationale; never silently rewritten by linter."

requirements-completed:
  - SCORER-01  # Foundation construction + concurrent-safe — verified by 1000-call determinism test
  - SCORER-03  # Auto-normalised weights — sum-to-1 invariant proven by TestScorer_WeightNormalisation_SumsToOne
  - SCORER-04  # Score (composite weighted score) — full external test coverage including identity + AsciiPair + determinism
  - SCORER-06  # Match (partial — Match method lands here; Threshold() accessor lands in plan 08-03)

# Metrics
duration: 12m
completed: 2026-05-17
---

# Phase 8 Plan 02: Scorer Struct + NewScorer + Score + Match Summary

**The Phase 8 composite weighted Scorer's executable surface lands — immutable struct, locked 9-step validation pipeline (missing-threshold FIRST), last-write-wins dedup, AlgoID-ascending internal order, weight auto-normalisation, float-determinism-locked reduction loop carrying forward the cosine.go:343 explicit-arithmetic pattern. Zero new runtime deps, no new imports, full suite green under `-race -count=1` and the determinism test green under `-count=10`.**

## Performance

- **Duration:** 12m
- **Started:** 2026-05-17T04:43:16Z
- **Completed:** 2026-05-17T04:55:03Z
- **Tasks:** 2 (TDD red→green per task)
- **Files modified:** 5 (3 new source/test files + 2 llms files synced)
- **Commits:** 2 atomic `feat(08-02)` commits

## Task Commits

Each task was committed atomically with the conventional-commit `feat(08-02):` prefix:

1. **Task 1: Scorer struct + NewScorer validation pipeline + weight normalisation + AlgoID-sorted entry slice** — `949e116`
2. **Task 2: Scorer.Score (determinism-locked reduction) + Scorer.Match (boundary-inclusive threshold)** — `8b39bc4`

## Files Created/Modified

### Created

- **`scorer.go` (388 lines)** — Phase 8's executable Scorer surface:
  - Package-header godoc block locking the design notes (validation pipeline order, dedup semantics, weight normalisation, float-determinism reduction, concurrency, FMA-fusion caveat) all citing CONTEXT.md / 08-RESEARCH.md / cosine.go.
  - `type Scorer struct` with four unexported fields (`algorithmsAlgoIDSorted`, `threshold`, `applyNormalisation`, `normaliseOpts`).
  - `func NewScorer(opts ...ScorerOption) (*Scorer, error)` — the 9-step LOCKED pipeline.
  - `func (s *Scorer) Score(a, b string) float64` — pre-normalisation + reduction loop.
  - `func (s *Scorer) Match(a, b string) bool` — `Score(a, b) >= threshold`.

- **`scorer_internal_test.go` (185 lines)** — package `fuzzymatch` invariants the external test cannot observe:
  - `TestScorer_LastWriteWins` — duplicate AlgoIDs collapse to ONE entry with the LATER weight; single-entry post-normalisation weight = 1.0 exactly.
  - `TestScorer_WeightNormalisation_SumsToOne` — 2-algo (1.0, 3.0) → (0.25, 0.75) with EXACT-1.0 sum.
  - `TestScorer_WeightNormalisation_DisabledRawPreserved` — `WithNormaliseWeights(false)` preserves raw weights.
  - `TestScorer_EntriesSorted_AlgoIDAscending` — options in reverse AlgoID order produce ascending slice; pairwise strict-ascending check.

- **`scorer_test.go` (305 lines)** — package `fuzzymatch_test` external coverage:
  - `TestNewScorer_MissingThreshold` (2 sub-cases) — discriminates the FIRST gate.
  - `TestNewScorer_EmptyScorer` — second gate.
  - `TestNewScorer_InvalidWeightPropagates` — option-layer short-circuit propagation.
  - `TestScorer_Score_Identity` — Score(x, x) = 1.0 exactly.
  - `TestScorer_Score_AsciiPair` — 1-algorithm composite equals raw `LevenshteinScore` byte-for-byte.
  - `TestScorer_Score_DeterministicAcrossCalls` — 1000 calls byte-identical; reinforced by `go test -count=10`.
  - `TestScorer_Match_ThresholdInclusive` — `Score == threshold → Match true`.
  - `TestScorer_Match_BelowThreshold` — `Score < threshold → Match false`.
  - `TestScorer_WithoutNormalisation` — `WithoutNormalisation()` causes the raw-bytes score on `XMLParser`/`xml_parser` to be strictly less than the normalised score; identity invariant holds regardless.

### Modified

- **`llms.txt`** — `Scorer` (type), `NewScorer` (function), `(s *Scorer) Score`, `(s *Scorer) Match` added under the Scorer construction section. Section heading retitled to reflect plan 08-02's contribution ("plan 08-02 lands NewScorer + Score + Match; ScoreAll / Threshold / Algorithms / DefaultScorer land in plan 08-03").
- **`llms-full.txt`** — new "Scorer type and NewScorer constructor (scorer.go — Phase 8 plan 08-02)" section with full godoc-summary blocks for `Scorer`, `NewScorer`, `Score`, `Match`. Inserted BEFORE the existing scorer_options section so the type comes before its consumers.

## Final Scorer Struct Field Set

```go
type Scorer struct {
    algorithmsAlgoIDSorted []scorerEntry  // dedup'd + AlgoID-ascending; reduction-loop iteration order
    threshold              float64        // from WithThreshold; required at construction
    applyNormalisation     bool           // gate for pre-Score Normalise calls
    normaliseOpts          NormalisationOptions  // stored when applyNormalisation == true
}
```

## NewScorer Step Ordering (final)

The 9-step LOCKED pipeline matching CONTEXT.md §2 + 08-RESEARCH.md Pitfall 3 verbatim:

1. **Initialise** `cfg := scorerConfig{normaliseWeights: true, applyNorm: true, normOpts: DefaultNormalisationOptions()}` BEFORE applying options so unset options produce documented defaults.
2. **Apply options** in supplied order with **first-error short-circuit** — `for _, opt := range opts { if err := opt(&cfg); err != nil { return nil, err } }`.
3. **Missing-threshold check FIRST** — `if !cfg.thresholdSet { return nil, ErrMissingThreshold }`.
4. **Empty-algorithms** — `if len(cfg.entries) == 0 { return nil, ErrEmptyScorer }`.
5. **Defensive per-entry AlgoID bounds + dispatch nil-check** — defence-in-depth over the option layer's same gate.
6. **Dedup via last-write-wins** — `seen[e.id] = e` over a `map[AlgoID]scorerEntry`.
7. **AlgoID-ascending materialisation** — iterate `AlgoIDs()` and append matching entries from `seen` into `sorted`.
8. **Weight auto-normalisation** (when `cfg.normaliseWeights == true`) — left-to-right sum (`sum = sum + sorted[i].weight`), defensive `sum == 0 → ErrInvalidWeight` guard, then divide each entry.
9. **Freeze** — return `&Scorer{algorithmsAlgoIDSorted: sorted, threshold: cfg.threshold, applyNormalisation: cfg.applyNorm, normaliseOpts: cfg.normOpts}`.

## Reduction Loop (textual form, copy-paste from scorer.go)

```go
var acc float64
for _, entry := range s.algorithmsAlgoIDSorted {
    score := entry.scoreFn(na, nb)
    // DET-06 explicit parens — see CONTEXT.md §5 LOCKED. The
    // `acc = acc + (entry.weight * score)` form (not `acc +=` and
    // not `acc + (entry.weight*score)` without parens) is the
    // locked determinism contract carrying forward from
    // cosine.go:343. The reduction is left-to-right, AlgoID-sorted,
    // and uses only + and * — no transcendentals, no FMA-defeating
    // double cast (see scorer.go header for the FMA-fusion
    // remediation pattern if cross-platform divergence ever
    // appears in the plan-08-04 golden gate).
    acc = acc + (entry.weight * score) //nolint:gocritic // DET-06 explicit-arithmetic locked pattern from cosine.go:343 / CONTEXT.md §5; assignOp shorthand would obscure the determinism contract
}
return acc
```

Plan's verification greps confirmed:

- `grep -F "(entry.weight * score)" scorer.go` → 7 matches (including the load-bearing line at the reduction site).
- `grep -F "acc = acc +" scorer.go` → 4 matches (including the load-bearing line).
- `grep -E "math\.(Pow|Log|Exp|FMA)" scorer.go` (excluding comment lines) → 0 matches.

## Weight Normalisation Correctness Proof (EXACT 1.0)

**Choice: assert EXACT 1.0 (not tolerance).** Both 0.25 and 0.75 are dyadic fractions (1/4 and 3/4) and exactly representable in IEEE-754 binary64. The sum is therefore byte-exactly 1.0; no `math.Abs(sum - 1.0) < 1e-12` tolerance is needed.

`TestScorer_WeightNormalisation_SumsToOne` asserts `sum != 1.0 → error` with `==`. Spot-checks also assert the individual values: `weight[0] != 0.25 → error` and `weight[1] != 0.75 → error`. These hold across the CI matrix because IEEE-754 division by a sum that is an integer (4.0 here) of dyadic-friendly fractions produces exact dyadic results.

This is the strongest form of the invariant. Property-test extensions in plan 08-03 (random weight vectors) will need a small tolerance because non-dyadic weights produce last-bit drift on the division — that is the correct boundary to draw.

## Deviations from Plan

### Auto-fixed Issues

**1. [Rule 3 - Blocking issue] golangci-lint gocritic assignOp findings on the determinism-locked reduction lines**

- **Found during:** Task 2 (post-implementation verification of `golangci-lint run ./...`)
- **Issue:** The plan's load-bearing acceptance criterion requires the EXACT textual form `acc = acc + (entry.weight * score)` in scorer.go (verified by `grep -F`). golangci-lint's gocritic linter flags this as `assignOp: replace with acc += (entry.weight * score)`. The same applies to `sum = sum + sorted[i].weight` and `sorted[i].weight = sorted[i].weight / sum` in the normalisation step.
- **Fix:** Added `//nolint:gocritic` directives on all three lines with explicit DET-06 / CONTEXT.md §5 rationale. The `+=` and `/=` shorthand are observationally equivalent but would (a) violate the plan's grep-verified contract and (b) obscure the determinism discipline at the read site. cosine.go:343 (the canonical precedent) avoids the assignOp trigger because its RHS (`(float64(qa[k]) * float64(qb[k])) + dot`) doesn't begin with `dot` — but the Scorer's reduction must begin with `acc =` for readability of the canonical "fold into accumulator" form, and gocritic doesn't distinguish.
- **Files modified:** `scorer.go` (3 lines with nolint directives + rationale comments)
- **Commit:** `8b39bc4`
- **Verification:** Zero golangci-lint findings on scorer.go, scorer_test.go, scorer_internal_test.go after the directives. Plan's `grep -F "(entry.weight * score)"` and `grep -F "acc = acc +"` both still pass.

**2. [Rule 1 - Style] gocyclo on NewScorer (complexity 16, threshold 10)**

- **Found during:** Task 1 (post-implementation lint pass)
- **Issue:** NewScorer's 9-step LOCKED validation pipeline produces a cyclomatic complexity of 16 (range loop + 7 conditional checks + 2 nested normalise loops). The project's `.golangci.yml` sets `gocyclo.min-complexity: 10`.
- **Fix:** Added `//nolint:gocyclo` on the function signature with rationale citing the 9-step locked pipeline and CONTEXT.md §2 + 08-RESEARCH.md Pitfall 3. Splitting into `validateThreshold` + `validateEntries` + `dedupAndSort` + `normaliseWeights` sub-helpers would scatter the LOCKED-order contract across files, harming reviewability.
- **Files modified:** `scorer.go` (1 line — function signature with trailing nolint)
- **Commit:** `949e116`
- **Verification:** Function reads linearly top-to-bottom matching the documented step list 1-9; reviewer can verify the pipeline order against CONTEXT.md §2 in <30s.

**3. [Rule 1 - Style] misspell "artifact" → "artefact" in scorer_internal_test.go godoc**

- **Found during:** Task 2 (post-implementation lint pass)
- **Issue:** Project's `.golangci.yml` configures misspell with `locale: UK`. The word "artifact" appeared in scorer_internal_test.go's package-header godoc ("ensures it never ships in the public artifact"); the UK spelling is "artefact".
- **Fix:** `s/artifact/artefact/` in scorer_internal_test.go's godoc.
- **Files modified:** `scorer_internal_test.go` (1 occurrence in godoc).
- **Commit:** `8b39bc4`
- **Note:** scorer_options_internal_test.go (created in plan 08-01) has the same misspelling. That file is NOT in plan 08-02's scope — leaving it for a future commit (or for plan 08-03's verification pass to surface alongside its own additions). Recorded here for traceability.

**4. [Rule 1 - Style] staticcheck QF1001 De Morgan's law rewrite in TestScorer_EntriesSorted_AlgoIDAscending**

- **Found during:** Task 2 lint pass
- **Issue:** The strict-ascending pairwise check used `if !(prev < curr) { … }`. staticcheck flags this as QF1001 "could apply De Morgan's law".
- **Fix:** Rewritten as `if prev >= curr { … }`. Semantically identical; more readable; clears the QF1001 finding.
- **Files modified:** `scorer_internal_test.go` (1 line).
- **Commit:** `8b39bc4`

### Plan Interpretation Choices

**5. [Plan interpretation] scorer.go has ZERO imports (plan said "imports ONLY \"sort\"")**

- **Plan wording:** Task 1 acceptance criterion says `scorer.go imports ONLY "sort" from stdlib`.
- **Why deviated:** The dedup-and-sort step uses `AlgoIDs()` (already in canonical order) to materialise the sorted slice; no `sort.Slice` or `sort.Sort` call is needed. Adding `import "sort"` and not using it would fail to compile in Go.
- **Choice:** Empty import set. This is STRICTLY within the plan bound — the criterion is an upper bound on stdlib imports, not a mandate to import sort. The "no math, no third-party" half of the criterion is satisfied. plan 08-03's Algorithms() accessor (which sorts a fresh slice copy via `sort.Slice` or similar) will be the introduction point for `import "sort"`.
- **Acceptance-criterion impact:** Plan's verification grep `grep -A5 "^import" scorer.go | grep -v "^import\|^)\|^(\|\".*\"" | wc -l` returns 0 (no extra import lines) so the criterion is satisfied. The "OR simpler: verify the import block contains only \"sort\"" branch is degenerate — an empty import set trivially contains "only sort" (vacuously).

### No Other Deviations

All other plan instructions were followed verbatim. Both tasks were implemented in strict TDD order (RED then GREEN). Every acceptance criterion in both task blocks is satisfied. All sentinel error values are reused from plan 08-01 (no new sentinels added). The Apache 2.0 file header is present on all three new files. No `init()`, no goroutines, no testify in root, no math.Pow/Log/Exp/FMA, no third-party imports, no `go.mod` change.

## Note for Plan 08-03

The following surfaces are NOT yet implemented and land in plan 08-03:

- `(s *Scorer) ScoreAll(a, b string) map[AlgoID]float64` (SPEC OVERRIDE — typed AlgoID keys)
- `(s *Scorer) Threshold() float64` accessor
- `(s *Scorer) Algorithms() []ScorerAlgorithm` (fresh slice per call, AlgoID-ascending)
- `type ScorerAlgorithm struct { ID AlgoID; Weight float64 }`
- `func DefaultScorer() *Scorer`
- `func DefaultScorerOptions() []ScorerOption`
- Property tests: `PropScorer_DeterministicAcrossRuns`, `PropScorer_WeightSumOne`, `PropScorer_ScoreInRange`
- Concurrent test invoking Score/ScoreAll/Match from N goroutines on the same `*Scorer` with `-race`

The Score/Match surface lands in plan 08-02 specifically so that the property + concurrent tests in plan 08-03 have a working composite reduction to exercise. The 1000-call `TestScorer_Score_DeterministicAcrossCalls` is the per-process precursor; the cross-platform byte-identity is gated by plan 08-04's golden file.

When plan 08-03 introduces `Algorithms()` and `Threshold()`, no changes to scorer.go beyond appending the two accessors are expected. The internal `algorithmsAlgoIDSorted` and `threshold` fields are already laid out for read-only accessor access.

## Threat Surface Scan

No threats outside the plan's `<threat_model>` register. The three registered threats are all mitigated:

| Threat ID | Mitigation status | Verifying test |
|-----------|-------------------|----------------|
| T-08-01 (weight tampering) | mitigated | Defensive `sum == 0 → ErrInvalidWeight` guard in NewScorer Step 8; `TestScorer_WeightNormalisation_SumsToOne` asserts the post-normalisation invariant; PropScorer_ScoreInRange lands in plan 08-03 |
| T-08-02 (out-of-range AlgoID DoS) | mitigated | Defensive `int(e.id) < 0 || int(e.id) >= numAlgorithms || dispatch[e.id] == nil → ErrInvalidAlgorithm` in NewScorer Step 5 (defence-in-depth over the option-layer's same gate) |
| T-08-03 (adversarial string DoS via Score) | mitigated (carried) | All 23 underlying algorithms are RE2-safe / linear-or-polynomial worst-case; Normalise handles invalid UTF-8 via U+FFFD; existing per-algorithm fuzz harnesses cover input panics; this plan adds no new surface that amplifies per-algorithm worst-case beyond the trivial N-times-1-to-6 factor of the reduction loop |

## Verification Run Log

| Command | Result |
|---------|--------|
| `go build ./...` | clean (no output) |
| `go vet ./...` | clean (no output) |
| `gofmt -l scorer.go scorer_test.go scorer_internal_test.go` | clean (no output) |
| `go test -race -run "TestScorer_LastWriteWins\|TestScorer_WeightNormalisation\|TestScorer_EntriesSorted" ./...` | green |
| `go test -race -run "TestNewScorer\|TestScorer_Score\|TestScorer_Match" ./...` | green |
| `go test -race -count=10 -run TestScorer_Score_DeterministicAcrossCalls ./...` | green (10 iterations) |
| `go test -race -count=1 ./...` | full root suite green |
| `go test -race -run TestAIFriendly ./...` | green (llms sync invariant satisfied) |
| `golangci-lint run ./...` (filtered to scorer files) | zero findings on scorer.go, scorer_test.go, scorer_internal_test.go |
| `grep -F "(entry.weight * score)" scorer.go` | 7 matches (load-bearing line + 6 documentation references) |
| `grep -F "acc = acc +" scorer.go` | 4 matches (load-bearing line + 3 documentation references) |
| `grep -E "math\.(Pow\|Log\|Exp\|FMA)" scorer.go` (code only, comments excluded) | 0 matches |
| `grep -c "^require" go.mod` | `1` (only the pre-existing `golang.org/x/text`) ✓ |

## Self-Check: PASSED

- ✓ `scorer.go` exists with `type Scorer struct` and `func NewScorer` + `func (s *Scorer) Score` + `func (s *Scorer) Match`
- ✓ `scorer_internal_test.go` exists with 4 invariant tests (all passing)
- ✓ `scorer_test.go` exists with 9 external tests (all passing)
- ✓ `949e116` (feat(08-02): declare Scorer struct and NewScorer validation pipeline) — task 1 commit found in git log
- ✓ `8b39bc4` (feat(08-02): add Scorer.Score and Scorer.Match with determinism-locked reduction) — task 2 commit found in git log
- ✓ Zero new runtime deps; `go.mod` unchanged from plan 08-01
- ✓ Full test suite green under `-race -count=1`
- ✓ Determinism test green under `-count=10`
- ✓ TestAIFriendly_LLMSTxtReferencesEveryExportedSymbol green (llms.txt + llms-full.txt synced)
- ✓ STATE.md and ROADMAP.md NOT modified (orchestrator owns those writes)
