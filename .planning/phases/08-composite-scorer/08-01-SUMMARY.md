---
phase: 08-composite-scorer
plan: 08-01
subsystem: scorer
tags: [functional-options, sentinel-errors, scorer, validation, phase-8]

# Dependency graph
requires:
  - phase: 01-foundation-infrastructure
    provides: errors.go sentinel pattern, AlgoID enum, dispatch[] array, numAlgorithms constant, NormalisationOptions struct, AlgoIDs() canonical order
  - phase: 02-character-algorithms
    provides: LevenshteinScore, JaroWinklerScore, Strcmp95Score, SmithWatermanGotohScoreWithParams, NewSWGParams, SWGParams
  - phase: 03-character-algorithms-cont
    provides: SWGParams type (validated)
  - phase: 05-q-gram-algorithms
    provides: QGramJaccardScore, SorensenDiceScore, CosineScore, TverskyScore, ErrInvalidQGramSize, ErrInvalidTverskyParam
  - phase: 06-token-based-algorithms
    provides: MongeElkanScoreSymmetric with permitted-inner allow-list
  - phase: 07-phonetic-algorithms
    provides: all 23 dispatch slots populated
provides:
  - Four sentinel errors for Scorer construction (ErrEmptyScorer, ErrInvalidWeight, ErrInvalidThreshold, ErrMissingThreshold)
  - ScorerOption public type (func(*scorerConfig) error)
  - Unexported scorerConfig + scorerEntry accumulator structs
  - 12 With* option functions covering CONTEXT.md §4 complete list
  - applyOptionForProbe / applyOptionsForProbe internal-test probe helpers for unexported scorerConfig inspection
affects: 08-02-NewScorer, 08-03-DefaultScorer-ScoreAll, 08-04-golden-BDD-docs

# Tech tracking
tech-stack:
  added: []  # zero new runtime deps — stdlib only
  patterns:
    - "Sentinel-error append (errors.go) — godoc parity with existing errors.go:38-107 pattern"
    - "Functional-options first-use — ScorerOption = func(*scorerConfig) error with first-error short-circuit"
    - "Slice-of-entries scorerConfig (NOT map-keyed) — last-write-wins resolved by NewScorer (plan 08-02); WithoutAlgorithm scan-and-compact removes ALL matching entries"
    - "Parameterised closure capture — every With*Algorithm closure captures consumer-supplied params by value (n, alpha, beta, inner, SWGParams)"
    - "In-package test file with probe helper (test-only) — chosen over external package fuzzymatch_test + extensive accessors because scorerConfig is unexported and the simpler path avoids growing a public-test-surface tower"

key-files:
  created:
    - scorer_options.go (393 lines — type ScorerOption, scorerEntry, scorerConfig, 12 With* functions)
    - scorer_options_test.go (510 lines — 40+ subtests covering happy + error paths for every option)
    - scorer_options_internal_test.go (110 lines — applyOptionForProbe and probe accessors)
  modified:
    - errors.go (4 new sentinels appended after ErrEmptyInput with godoc parity)
    - errors_test.go (sentinelCases() extended; new TestSentinels_Identity; TestSentinels_LowercaseAndNoTrailingPunctuation relaxed to first-rune check)
    - llms.txt (10 new lines under Errors + new Scorer construction options section)
    - llms-full.txt (Scorer construction options section + 4 new sentinel godoc blocks)

key-decisions:
  - "scorerConfig is a slice-of-entries (not map-keyed). Last-write-wins resolution and WithoutAlgorithm scan-and-compact are simplest natively; the dedup step is NewScorer's job in plan 08-02."
  - "WithoutAlgorithm removes ALL matching entries (not just the most recent). Confirmed by TestWithoutAlgorithm_RemovesAllMatchingEntries — duplicate WithAlgorithm(AlgoLevenshtein, _) twice then remove → zero entries. This matches Pitfall 4 from 08-RESEARCH.md and ensures NewScorer's later dedup sees a clean slice."
  - "Probe-helper approach for unit testing the option layer. scorer_options_internal_test.go (package fuzzymatch, _test suffix on the file makes it test-only) exposes applyOptionForProbe / applyOptionsForProbe and read-only accessors over the unexported scorerConfig. The tests in scorer_options_test.go (package fuzzymatch, in-package) drive options against fresh configs and inspect both error returns and accumulated state."
  - "TverskyScore parameter order is (a, b, n, alpha, beta) — n FIRST after the strings. The plan's prose said (alpha, beta, n) which is the option's parameter order; the inner closure passes them in the actual TverskyScore order: TverskyScore(a, b, n, alpha, beta). Verified by reading tversky.go:230."
  - "SmithWatermanGotohScoreWithParams is the params-accepting entry point (not SmithWatermanGotohScore which uses NewSWGParams defaults). The closure calls SmithWatermanGotohScoreWithParams."
  - "WithMongeElkanAlgorithm rejects inner == AlgoMongeElkan explicitly with ErrInvalidAlgorithm. The full 18-entry permitted-inner allow-list remains enforced at Score time inside MongeElkanScoreSymmetric (Phase 6 + 7 locked behaviour); the option layer adds only the trivial-recursion + bounds check."
  - "ME's vestigial opts NormalisationOptions parameter is passed as DefaultNormalisationOptions() from every WithMongeElkanAlgorithm closure. Per CONTEXT.md §8 the opts is currently a no-op (accepted-but-ignored on monge_elkan.go:393); Phase 8 plumbs it for forward-compat without modifying monge_elkan.go."

patterns-established:
  - "Pattern 1: Slice-of-entries config struct — accumulates options in application order; dedup + sort happens at constructor time (NewScorer in plan 08-02), not at option-application time."
  - "Pattern 2: Validate-at-option-application-time — each option returns the appropriate sentinel error immediately. NewScorer short-circuits on the first error so the user sees the first malformed option, not a cascading list."
  - "Pattern 3: Capture closure params by value — every parameterised option (n, alpha, beta, inner, params SWGParams) stores its config inside the closure by value so the same option may be applied to multiple NewScorer calls without state aliasing."
  - "Pattern 4: In-package test probes for unexported state — scorer_options_internal_test.go (test-only via _test.go suffix) exposes applyOptionForProbe and read-only accessors. External tests (when needed) can call these helpers without inflating the public API surface with getters."

requirements-completed:
  - SCORER-01  # Foundation: option-application-time validation, option layer is the substrate
  - SCORER-03  # Auto-normalised weights — WithNormaliseWeights option + ErrInvalidWeight gate land here (sum-to-1 lands in plan 08-02)
  - SCORER-08  # Normalisation control — WithNormalisation / WithoutNormalisation options land here (wiring through Score lands in plan 08-02)

# Metrics
duration: 13m
completed: 2026-05-17
---

# Phase 8 Plan 01: Scorer Errors + Functional-Options Surface Summary

**Twelve `With*` functional-option constructors plus four new sentinel errors land the Phase 8 substrate so plan 08-02's `NewScorer` validation pipeline can consume a frozen option contract — zero new runtime dependencies, stdlib `testing` only, no `init()`, no goroutines, no third-party imports.**

## Performance

- **Duration:** 13m 43s
- **Started:** 2026-05-17T04:25:04Z
- **Completed:** 2026-05-17T04:38:47Z
- **Tasks:** 3 (TDD red→green per task)
- **Files modified:** 7 (4 new test+source files; 4 existing files appended: errors.go, errors_test.go, llms.txt, llms-full.txt)
- **Commits:** 3 atomic feat commits

## Accomplishments

- **Four new sentinel errors** appended to `errors.go` with godoc parity to the existing `ErrInvalidInput`/`ErrInvalidQGramSize`/`ErrInvalidTverskyParam` pattern (description → blank doc line → `Discriminate via errors.Is(err, fuzzymatch.ErrX); never match the error message string.`). Canonical messages locked per CONTEXT.md §2: `ErrEmptyScorer`, `ErrInvalidWeight`, `ErrInvalidThreshold`, `ErrMissingThreshold`.
- **ScorerOption + scorerConfig + scorerEntry types** declared in the new `scorer_options.go`. `ScorerOption func(*scorerConfig) error` is the public surface; `scorerConfig` and `scorerEntry` are unexported package-private accumulators.
- **All 12 `With*` option functions** implemented and unit-tested. Six non-parameterised (`WithAlgorithm`, `WithoutAlgorithm`, `WithNormalisation`, `WithoutNormalisation`, `WithThreshold`, `WithNormaliseWeights`) plus six parameterised algorithm options (`WithQGramJaccardAlgorithm`, `WithSorensenDiceAlgorithm`, `WithCosineAlgorithm`, `WithTverskyAlgorithm`, `WithMongeElkanAlgorithm`, `WithSmithWatermanGotohAlgorithm`).
- **40+ unit tests** in `scorer_options_test.go` covering every option's happy path AND every error path. Closure capture is verified by invoking the stored `scoreFn` directly through a probe accessor and comparing against direct `XxxScore(...)` calls with the same parameters — confirms `n`, `alpha`, `beta`, `inner`, `SWGParams` are captured correctly.
- **llms.txt + llms-full.txt synced in lockstep** with every new exported symbol so `TestAIFriendly_LLMSTxtReferencesEveryExportedSymbol` stays green.

## Task Commits

Each task was committed atomically:

1. **Task 1: Append four sentinel errors to errors.go and sync llms.txt/llms-full.txt** — `bbc12f9` (feat)
2. **Task 2: Declare ScorerOption + scorerConfig + scorerEntry and six non-parameterised options** — `1ea99e2` (feat)
3. **Task 3: Implement six parameterised algorithm options** — `7d14777` (feat)

## Files Created/Modified

### Created

- `scorer_options.go` — public `ScorerOption` type, unexported `scorerEntry` + `scorerConfig` accumulator structs, 12 `With*` option function implementations with full godoc.
- `scorer_options_test.go` — 40+ unit tests covering happy + error paths for every option. Uses `errors.Is` discrimination (never string match), and exercises closure capture for the parameterised options by invoking the stored `scoreFn` and comparing against the direct `XxxScore(...)` reference.
- `scorer_options_internal_test.go` — package-internal probe helpers (`applyOptionForProbe`, `applyOptionsForProbe`, `probeEntryCount`, `probeEntryAt`, `probeEntryHasScoreFn`, `probeThreshold`, `probeNormalisation`, `probeNormaliseWeights`, `probeScoreFnInvoke`). Test-only via `_test.go` suffix; never ships in the public artifact.

### Modified

- `errors.go` — appended 4 sentinel `var Err* = errors.New(...)` declarations after `ErrEmptyInput`. Now contains 10 sentinels total (6 pre-existing + 4 new).
- `errors_test.go` — extended `sentinelCases()` to include the 6 existing parameterised sentinels plus 4 new Phase 8 sentinels; added `TestSentinels_Identity` pinning the locked message strings from CONTEXT.md §2; relaxed `TestSentinels_LowercaseAndNoTrailingPunctuation` to gate only the FIRST-rune capitalisation per Effective Go's actual convention (embedded public-API identifier names in remediation hints are permitted).
- `llms.txt` — added a new `### Scorer construction options` section with 13 new symbol lines (ScorerOption + 12 With* functions); appended 4 new sentinel lines under `### Sentinel errors`.
- `llms-full.txt` — added Scorer-construction-options section with full godoc blocks for the 13 symbols; appended 4 new sentinel godoc blocks after the existing `ErrInvalidTverskyParam` entry.

## Sentinel Error Contracts (locked)

The four new sentinels carry these exact canonical message strings (verified by `TestSentinels_Identity`):

| Sentinel | Message |
|----------|---------|
| `ErrEmptyScorer` | `fuzzymatch: scorer has no algorithms (pass at least one WithAlgorithm option or use DefaultScorer)` |
| `ErrInvalidWeight` | `fuzzymatch: invalid algorithm weight (must be > 0)` |
| `ErrInvalidThreshold` | `fuzzymatch: invalid threshold (must be in [0.0, 1.0])` |
| `ErrMissingThreshold` | `fuzzymatch: scorer threshold required (pass WithThreshold or use DefaultScorer)` |

## Option Function Signatures (final)

| Function | Signature | Returns |
|----------|-----------|---------|
| `WithAlgorithm` | `(algo AlgoID, weight float64) ScorerOption` | ErrInvalidWeight, ErrInvalidAlgorithm |
| `WithoutAlgorithm` | `(id AlgoID) ScorerOption` | nil (silent no-op if absent) |
| `WithNormalisation` | `(opts NormalisationOptions) ScorerOption` | nil |
| `WithoutNormalisation` | `() ScorerOption` | nil |
| `WithThreshold` | `(t float64) ScorerOption` | ErrInvalidThreshold |
| `WithNormaliseWeights` | `(normalise bool) ScorerOption` | nil |
| `WithQGramJaccardAlgorithm` | `(weight float64, n int) ScorerOption` | ErrInvalidWeight, ErrInvalidQGramSize |
| `WithSorensenDiceAlgorithm` | `(weight float64, n int) ScorerOption` | ErrInvalidWeight, ErrInvalidQGramSize |
| `WithCosineAlgorithm` | `(weight float64, n int) ScorerOption` | ErrInvalidWeight, ErrInvalidQGramSize |
| `WithTverskyAlgorithm` | `(weight, alpha, beta float64, n int) ScorerOption` | ErrInvalidWeight, ErrInvalidQGramSize, ErrInvalidTverskyParam |
| `WithMongeElkanAlgorithm` | `(weight float64, inner AlgoID) ScorerOption` | ErrInvalidWeight, ErrInvalidAlgorithm (including trivial-self-recursion rejection) |
| `WithSmithWatermanGotohAlgorithm` | `(weight float64, params SWGParams) ScorerOption` | ErrInvalidWeight |

## Internal `scorerConfig` Shape (frozen for plan 08-02)

```go
type scorerEntry struct {
    id      AlgoID
    weight  float64
    scoreFn func(a, b string) float64
}

type scorerConfig struct {
    entries          []scorerEntry            // appended-in-order; deduped + sorted by NewScorer
    threshold        float64                  // set by WithThreshold
    thresholdSet     bool                     // distinguishes "never set" from "set to 0.0"
    normaliseWeights bool                     // controls sum-to-1 step at NewScorer time
    applyNorm        bool                     // controls Normalise call in Scorer.Score
    normOpts         NormalisationOptions     // stored when WithNormalisation applied
}
```

This shape is the frozen contract for plan 08-02's `NewScorer` constructor. Plan 08-02 will:
1. Iterate `opts` in order, short-circuiting on the first non-nil error.
2. Check `!cfg.thresholdSet` → `ErrMissingThreshold` FIRST (per Pitfall 3, CONTEXT.md §2).
3. Check `len(cfg.entries) == 0` → `ErrEmptyScorer`.
4. Dedup entries via `seen := map[AlgoID]scorerEntry` (last-write-wins on duplicates).
5. Normalise weights if `cfg.normaliseWeights` (default true at NewScorer; this plan's option default at zero-value time is FALSE — plan 08-02 sets it true before iterating options).
6. Sort the resulting `[]scorerEntry` by `AlgoIDs()` canonical order.
7. Freeze the sorted slice into `*Scorer.algorithmsAlgoIDSorted` (read-only after NewScorer returns).

## Probe Helper Mechanism

Because `scorerConfig` is unexported and plan 08-01 does not yet provide `NewScorer`, the option layer is unit-tested via the in-package probe helpers in `scorer_options_internal_test.go`:

```go
// scorer_options_internal_test.go (package fuzzymatch — test-only)
func applyOptionForProbe(opt ScorerOption) (scorerConfig, error)
func applyOptionsForProbe(opts ...ScorerOption) (scorerConfig, error)
func probeEntryCount(cfg scorerConfig) int
func probeEntryAt(cfg scorerConfig, i int) (AlgoID, float64)
func probeEntryHasScoreFn(cfg scorerConfig, i int) bool
func probeThreshold(cfg scorerConfig) (float64, bool)
func probeNormalisation(cfg scorerConfig) (bool, NormalisationOptions)
func probeNormaliseWeights(cfg scorerConfig) bool
func probeScoreFnInvoke(cfg scorerConfig, i int, a, b string) float64
```

These helpers will remain useful in plan 08-02 for `scorerConfig` last-write-wins invariants that NewScorer cannot directly expose. They are package-internal (and test-only via the `_test.go` filename suffix), so the public artifact remains free of any test-helper surface.

## Deviations from Plan

### Auto-fixed Issues

**1. [Rule 1 - Bug] Loosened `TestSentinels_LowercaseAndNoTrailingPunctuation` to first-rune check**

- **Found during:** Task 1 (RED→GREEN transition for `ErrEmptyScorer` and `ErrMissingThreshold`)
- **Issue:** The pre-existing convention test iterated every rune of the error message body and rejected any uppercase rune. The four new sentinels carry locked message strings (per CONTEXT.md §2 and PATTERNS.md) that include embedded Go identifier names like `WithAlgorithm`, `DefaultScorer`, `WithThreshold` — these are required to point the consumer at the correct remediation API. The Go convention from Effective Go ("Error strings should not be capitalized") constrains the FIRST word of the message, not every embedded rune. Standard library examples like `"io: read/write on closed pipe"` embed capitalised identifier-name-equivalents inside the message body.
- **Fix:** Relaxed the test to inspect only the first rune of the body after the `fuzzymatch:` prefix. Updated the test godoc to explain the rationale ("constrains the FIRST-rune capitalisation only — embedded identifier names ... are permitted because they refer to public API symbols and disambiguate the actionable next step for the consumer").
- **Files modified:** `errors_test.go` (lines 174-199 — `TestSentinels_LowercaseAndNoTrailingPunctuation`)
- **Commit:** `bbc12f9`
- **Verification:** All 6 pre-existing sentinels still pass this gate (their bodies are all lowercase first-rune); the 4 new Phase 8 sentinels pass.

### Plan Interpretation Choices

**2. [Plan interpretation] `scorer_options_test.go` uses `package fuzzymatch` (in-package), not `package fuzzymatch_test`**

- **Plan wording:** The plan's Task 2 `<action>` block described the test file as `package fuzzymatch_test` with a probe helper from an in-package `scorer_options_internal_test.go`.
- **Why deviated:** `applyOptionForProbe` returns `(scorerConfig, error)` — but `scorerConfig` is unexported. An external `package fuzzymatch_test` file cannot accept that return type directly; it would either need a large surface of exported per-field accessors OR the tests would have to be in-package.
- **Choice:** In-package `scorer_options_test.go` with read-only probe accessors in `scorer_options_internal_test.go`. The test file does NOT call any unexported function directly except via the documented probe helpers, so it remains a thin compositional gate.
- **Acceptance-criterion impact:** The acceptance criterion (`Probe helper file (if used) is named scorer_options_internal_test.go with package fuzzymatch and contains func applyOptionForProbe(`) is satisfied. The criterion did not require `scorer_options_test.go` to be `package fuzzymatch_test` strictly.
- **Forward implication:** Plan 08-02 can introduce end-to-end `NewScorer(opts...)` tests in either package without conflict; the probes remain as a thin invariant-gate over unexported state.

**3. [Plan parameter ordering — corrected]** `TverskyScore` signature is `(a, b, n, alpha, beta)`, not `(a, b, alpha, beta, n)`

- **Plan wording:** Task 3 `<behavior>` block said "`WithTverskyAlgorithm(0.5, 1.0, 1.0, 3)` produces an option storing scoreFn closure over (alpha=1, beta=1, n=3)" and the action block said "closure calls TverskyScore(a, b, alpha, beta, n)".
- **Actual function signature:** Reading `tversky.go:230` — `func TverskyScore(a, b string, n int, alpha, beta float64) float64` — `n` comes FIRST, before alpha/beta.
- **Fix:** The closure calls `TverskyScore(a, b, n, alpha, beta)` (correct order). The option's PARAMETER order remains as the plan stated: `WithTverskyAlgorithm(weight, alpha, beta float64, n int)` — the consumer-facing API ergonomic does NOT match the underlying function's parameter order, which is fine; the closure translates.
- **Verified:** `TestWithTverskyAlgorithm_CapturesParams` invokes the closure and compares against `TverskyScore(a, b, c.n, c.alpha, c.beta)` (the correct argument order); the test passes.

### No Other Deviations

All other plan instructions were followed verbatim. The 12 `With*` function signatures match the plan exactly. The 4 sentinel error names and canonical messages match CONTEXT.md §2 exactly. The validation pipeline order is set up for plan 08-02 (threshold-set FIRST, then empty-scorer, then weight, then per-entry checks). No `init()`, no goroutines, no third-party imports, no testify in root.

## Threat Surface Scan

No threats outside the plan's `<threat_model>` register. The two registered threats are both mitigated:

| Threat ID | Mitigation status | Verifying test |
|-----------|-------------------|----------------|
| T-08-01 (weight + numeric param tampering) | mitigated | `TestWithAlgorithm_InvalidWeight`, `TestWithThreshold_OutOfRange`, `TestWithQGramJaccardAlgorithm_InvalidN`, `TestWithTverskyAlgorithm_InvalidAlpha`, `TestWithTverskyAlgorithm_InvalidBeta` |
| T-08-02 (out-of-range AlgoID / trivial-recursion DoS) | mitigated | `TestWithAlgorithm_InvalidAlgoID`, `TestWithMongeElkanAlgorithm_RejectsSelf`, `TestWithMongeElkanAlgorithm_InvalidInner` |

## Verification Run Log

| Command | Result |
|---------|--------|
| `go vet ./...` | clean (no output) |
| `gofmt -l scorer_options.go scorer_options_test.go scorer_options_internal_test.go errors.go errors_test.go` | clean (no output) |
| `go test -race -run "TestSentinels\|TestWith" ./...` | green |
| `go test -race -run TestAIFriendly ./...` | green (llms sync) |
| `go test -race ./... -count=1` | full suite green |
| `grep -c "^var Err" errors.go` | `10` (6 existing + 4 new) ✓ |
| `grep -c "^func With" scorer_options.go` | `12` ✓ |
| `grep -c "^require" go.mod` | `1` (only the pre-existing `golang.org/x/text`) ✓ |

## Self-Check: PASSED

- ✓ `scorer_options.go` exists
- ✓ `scorer_options_test.go` exists
- ✓ `scorer_options_internal_test.go` exists
- ✓ `errors.go` contains 4 new sentinels (`grep -c "^var Err"` returns 10)
- ✓ `scorer_options.go` contains 12 `With*` functions
- ✓ `bbc12f9` exists (`feat(08-01): add four Scorer sentinel errors`)
- ✓ `1ea99e2` exists (`feat(08-01): declare ScorerOption surface and six non-parameterised options`)
- ✓ `7d14777` exists (`feat(08-01): implement six parameterised algorithm options`)
- ✓ Zero new runtime deps (root go.mod unchanged besides existing `golang.org/x/text`)
- ✓ Full test suite green under `-race -count=1`
- ✓ llms.txt + llms-full.txt synced; `TestAIFriendly_LLMSTxtReferencesEveryExportedSymbol` passes
