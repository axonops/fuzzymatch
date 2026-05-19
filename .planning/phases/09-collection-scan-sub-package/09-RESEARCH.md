# Phase 9: Collection Scan Sub-package — Research

**Researched:** 2026-05-19
**Domain:** Collection-scan layer (Layer 3) — `github.com/axonops/fuzzymatch/scan`
**Confidence:** HIGH

---

<user_constraints>
## User Constraints (from 09-CONTEXT.md)

> Verbatim copy of the locked decisions, discretion areas, and deferred ideas. The planner MUST treat the Decisions block as immutable input and MUST NOT explore alternatives to D-01..D-08.

### Locked Decisions

**§1. Warning type alignment with Phase 8 — LOCKED (SPEC OVERRIDE × 2)**

- **D-01:** `scan.Warning.Scores` is **`map[AlgoID]float64`**, overriding spec §12.1 line 1337's `map[string]float64`. **Rationale:** `Warning.Scores` is populated directly from `Scorer.ScoreAll`, which Phase 8 LOCKED to `map[AlgoID]float64` (08-CONTEXT.md §1, SPEC OVERRIDE applied to `docs/requirements.md` §8.3 + §8.6 in lockstep with Plan 08-03). The same override applies here for the same reason: typed enum keys give compile-time safety; `map[string]float64` would force consumers to re-encode the snake_case namespace. Spec amendment lands in the planner's first commit per the Phase 8 precedent.

- **D-02:** The within/cross discriminator type is **`scan.Kind`** (NOT `scan.WarningKind` as in spec §12.1, NOT `scan.WarnKind` for accidental symmetry with root's `fuzzymatch.WarnKind`). Constants unchanged: `KindWithinGroup`, `KindAcrossGroups`. `String()` returns CamelCase (`"WithinGroup"`, `"AcrossGroups"`) per the §6 + Phase 8.5 Q6b convention. **Rationale:** package-scoped `Kind` is unambiguous at the call site (`scan.KindWithinGroup`); the `WarnKind` / `WarningKind` pair would mislead consumers into expecting a shared inheritance. Spec amendment renames `WarningKind` → `Kind` throughout §12.1, §12.4, §15 references.

**§2. Config validation strictness — LOCKED (Phase 8.5 Q2 framework applied to scan)**

- **D-03:** Empty `Item.Name` → `ErrInvalidItem`, all offending indices collected via `errors.Join`. Spec §12.1 line 1389 amended to "every offending index, joined via `errors.Join`."

- **D-04:** `Config.CrossGroupThresholdBoost` zero-value = `0.0` (no boost). The opinionated default `0.05` lives in `scan.DefaultConfig(s *fuzzymatch.Scorer) Config`. Strict validation: NaN / ±Inf / < 0 / > 1 → `ErrInvalidConfig` at `Check` call time. Spec §12.1 line 1359 default-0.05 statement migrates to `DefaultConfig` godoc.

- **D-05:** `Config.SuppressedPairs` validation: empty strings → `ErrInvalidConfig`, all offending indices collected via `errors.Join`. Self-pairs (`a == b` post-normalisation) silently kept.

**§3. Sort-key completeness — LOCKED**

- **D-06:** Duplicate `(Name, Group)` pairs in `Items[]` → `ErrInvalidItem` at validation time, all duplicate indices collected via `errors.Join`. With duplicates rejected at the door, the sort key `(Kind, NameA, NameB, GroupA, GroupB)` is provably unique by construction. The SCAN-05 in-line completeness assertion becomes a hard invariant (`if dup found in sorted output → panic with fuzzymatch.ErrInternalInvariantViolated`).

**§4. Validate / scan integration — LOCKED**

- **D-07:** Separate concerns. `scan.Check` does similarity computation only. `scan.Check` does NOT internally invoke `Validate`; `scan.Warning` does NOT carry a `Diagnostics []fuzzymatch.Warning` field; `Check` returns `([]scan.Warning, error)`.

**§5. Bucket threshold — LOCKED as private constant**

- **D-08:** `bucketThreshold = 50` is a private package constant in `scan/bucket.go`, NOT a `Config.BucketThreshold` field. **Empirical validation required:** `BenchmarkScanCheck_BucketVsNaive_GroupSize` sweep (10/25/50/75/100/200) — if the crossover differs materially from 50, update the constant before Phase 9 closes.

### Claude's Discretion

- Item validation surfaces: `fmt.Errorf("scan: invalid item at index %d: %w", idx, ErrInvalidItem)` vs dedicated `*InvalidItemError` struct — either acceptable as long as `errors.Is(err, scan.ErrInvalidItem)` works and the index is in the message.
- Token-bucket internal data layout: `map[string][]int` per spec §12.5 vs alternative — `PropCheck_BucketEquivalentToNaive` is the load-bearing gate.
- Whether to expose a precomputed `tokeniseAll` helper in `scan/bucket.go` vs inline it inside `Check` — Claude's call. Each Item.Name tokenised at most once per `Check` invocation.
- `examples/scan-demo/main.go` narrative shape — must demonstrate (a) within-group default Scorer, (b) within + cross-group with `DefaultConfig`, (c) at least one of each suppression mode.
- `scan.feature` / `suppression.feature` scenario count — bdd-scenario-reviewer's call. Minimum: one scenario per D-03..D-07, plus the Phase 8.5 R2 deferred suppression scenarios.

### Deferred Ideas (OUT OF SCOPE)

- `Config.BucketThreshold` public field — revisit post-v1.0 if tuning need surfaces.
- Side-channel `Validate` integration (e.g. `scan.CheckWithDiagnostics`) — revisit if boilerplate signal emerges.
- Streaming `iter.Seq[Warning]` Check variant — V2-API-03.
- Parallel goroutine-fanned scan — V2-API-04.
- Dedicated `scan.Item` constructor / builder — not raised; plain struct literal pattern is fine for v1.0.
</user_constraints>

<phase_requirements>
## Phase Requirements

| ID | Description | Research Support |
|----|-------------|------------------|
| SCAN-01 | `scan.Check(items, cfg) []Warning` — within-group + cross-group passes with separate thresholds | §"Plan Decomposition" Plans 09-01..03; §"Within-group vs cross-group pass" |
| SCAN-02 | Token-bucket optimisation — property-test verified equivalent to naive O(N²) via `PropCheck_BucketEquivalentToNaive` | §"Token-bucket optimisation strategy"; §"Property test design" |
| SCAN-03 | Per-item `SilenceLint` flag + global `SuppressedPairs` list composing additively | §"Suppression composition semantics" |
| SCAN-04 | Cross-group identical-name suppression default (configurable) | §"Suppression composition semantics" |
| SCAN-05 | Deterministic output sort by `(Kind, NameA, NameB, GroupA, GroupB)` with in-line completeness assertion | §"Sort key completeness assertion"; §"Determinism golden file" |
| SCAN-06 | Sentinel error hierarchy for scan-specific failures | §"Sentinel errors"; §"Validation order + errors.Join discipline" |
| PERF-05 | `scan.Check` performance budget — < 2s for 10,000 items | §"Performance budget"; §"Benchmark suite" |

Cross-cutting requirements implicitly addressed:
- **DET-04** — NaN/Inf/-0 handling propagates from Scorer; `Config.CrossGroupThresholdBoost` validation rejects NaN/Inf at Check entry.
- **TEST-05** — BDD: `scan.feature` + `suppression.feature` (the Phase 8.5 R2 deferral).
- **DX-02** — godoc `Example` per public function (Item, Config, Warning, Kind, Check, DefaultConfig, three sentinels).
- **DX-04** — `docs/scan.md` populated from current scaffold.
</phase_requirements>

## Project Constraints (from CLAUDE.md)

- **Zero non-stdlib runtime deps in root `go.mod`** (only `golang.org/x/text`). The `scan/` sub-package lives under the root module — adding any non-stdlib dep here breaks `make verify-deps-allowlist`. Enforced.
- **No cgo anywhere.** `CGO_ENABLED=0 go build ./...` must pass for the scan sub-package.
- **No testify in root tests.** Permitted only in `tests/bdd/` (already imported there). `scan/*_test.go` uses stdlib `testing` only.
- **No goroutines / channels / mutexes in library code.** Scan is a pure-function library (D-07 + spec §12.4). `goleak.VerifyTestMain` (already wired in `tests/bdd/bdd_test.go`) catches regressions.
- **No `init()` doing non-trivial work** (DET-05). `bucketThreshold` is a `const`; any internal lookup table built via `var = literal`.
- **No map iteration on output paths** (DET-03). `map[string][]int` allowed internally; output assembled into a sorted slice and sorted again at the end.
- **No transcendental float ops** (`math.Pow`/`Log`/`Exp`/`FMA`). Scan does no float arithmetic of its own beyond `min(1.0, scorerThreshold + boost)` and `>=` comparisons; the Scorer handles all scoring.
- **CI-only releases.** Phase 9 ships nothing tagged; the v1.0.0 tag is Phase 11.
- **Mandatory agent gates per phase:** `code-reviewer`, `security-reviewer`, `go-quality`, `test-writer`, `bdd-scenario-reviewer`, `docs-writer`, `user-guide-reviewer`, `devops`, `algorithm-performance-reviewer` (for D-08 bench validation), `determinism-reviewer` (golden-file gate), `api-ergonomics-reviewer` (foundation plan SPEC OVERRIDE sign-off — D-01 + D-02). Per user-memory `project_no_github_issues.md`: skip `commit-message-reviewer` issue-ref findings and `issue-writer`/`issue-closer` (no GitHub issues in use).
- **British English in `.go` files and docs** (per Phase 8.5 Plan 17b sweep; "normalise" not "normalize", "behaviour" not "behavior").
- **Serial phase execution** (user-memory `feedback_phase_execution_serial.md`) — Phase 9 runs one plan at a time, no parallel worktrees.
- **Push to origin at every phase boundary** (user-memory `feedback_push_cadence.md`) — Phase 9 close pushes to `origin/main`; verify CI green before the verification gate (user-memory `feedback_ci_before_verification_gate.md`).
- **SDK wave-mapping caveat** (user-memory `project_sdk_wave_mapping_bug.md`) — when reading plan wave structure, read `wave:` from each plan's frontmatter directly; do not trust `phase-plan-index` aggregate.

## Summary

Phase 9 ships the `scan/` sub-package — Layer 3 of the fuzzymatch architecture — as a thin, pure-function turnkey collection-scan over Phase 8's Scorer. The spec already pins the surface shape (`Item`, `Warning`, `Config`, `Check`, three sentinels, within/cross passes, suppression composition, deterministic sort, token-bucket optimisation, performance budget); CONTEXT.md §1–§5 has further refined eight load-bearing decisions on top of the spec.

The work is **largely additive** — no algorithm implementation, no float-determinism battles, no primary-source citations. The risk surface is concentrated in three areas:

1. **Spec-amendment lockstep** (D-01 + D-02): the foundation plan must amend `docs/requirements.md` §12.1 + §12.4 + §15 references + line 2010 + line 2061 in the same commit as the type declarations land. Failure to do this in lockstep is the Phase 8 precedent for "SPEC OVERRIDE" PR-review confusion.
2. **`errors.Join` discipline across three validation paths** (D-03/D-05/D-06): uniform collect-all-then-join semantics, accumulating offending indices, with `errors.Is` discrimination preserved across the join. The user corrected the initial fail-fast framing twice in discussion (09-DISCUSSION-LOG.md line 49, line 70, line 85) — this pattern is non-negotiable.
3. **Empirical bucket-threshold validation** (D-08): the `BenchmarkScanCheck_BucketVsNaive_GroupSize` sweep is load-bearing for the < 2s/10k-items PERF-05 budget. The spec's `50` is a hypothesis; the benchmark is the gate.

**Primary recommendation:** Decompose into **8 plans across 6 waves**, foundation-first (types + spec amendments + sentinels in one atomic commit so the SPEC OVERRIDE PR-review record is clean), validation pre-flight as a separate plan (because three coupled errors.Join paths in one plan is too much surface area), then within/naive-cross pass, then token-bucket optimisation + property test, then suppression composition, then sort + golden + completeness assertion, then BDD + finalisation. The empirical bucket benchmark lives in the optimisation plan; if it falsifies `50`, the constant updates in that same plan.

## Architectural Responsibility Map

| Capability | Primary Tier | Secondary Tier | Rationale |
|------------|-------------|----------------|-----------|
| Item validation (empty Name, duplicate (Name,Group)) | scan pre-flight pass | — | Pure data validation; uses no Scorer logic |
| Config validation (Scorer nil, boost range, SuppressedPairs entries) | scan pre-flight pass | — | Pure parameter validation per Q2 framework |
| Per-pair similarity computation | Scorer (root pkg) | scan inner loop | scan calls `Scorer.Match` / `Scorer.ScoreAll`; the float arithmetic is Phase 8 territory |
| Normalisation of Names and SuppressedPairs | Scorer + scan | — | scan reads Scorer's `applyNormalisation` + `normaliseOpts` (currently unexported — see Open Question 1); scan applies the same Normalise to SuppressedPairs |
| Tokenisation for bucket | scan (`scan/bucket.go`) | root pkg Tokenise | scan calls `fuzzymatch.Tokenise(name, DefaultTokeniseOptions())` once per Item; the bucket is scan-private |
| Within-group pass | scan core loop | — | Group-keyed iteration; sub-bucket-threshold uses naive nested loops |
| Cross-group pass | scan core loop | — | Off when `CompareAcrossGroups == false`; threshold = `min(1.0, scorer.Threshold() + boost)` |
| Suppression composition | scan suppression module | — | Three rules compose via OR; per-pair check at emission time |
| Output sort + completeness assertion | scan post-pass | — | `sort.SliceStable` on the five-tuple key; in-line linear scan asserts no adjacent duplicates |
| Determinism (cross-platform golden) | scan + CI matrix | — | `testdata/golden/scan-default.json` joins the existing 4-file inventory |
| Performance < 2s / 10k items | scan + Scorer + Tokenise | — | scan controls bucketing; Scorer/Tokenise contribute ASCII fast paths from Phase 8.5 Q8b |

## Standard Stack

### Core (already locked at project scope)

| Library | Version | Purpose | Why Standard |
|---------|---------|---------|--------------|
| Go stdlib `sort` | 1.26.3 bundled | `sort.SliceStable` for output ordering | DET-03 + project convention (Validate uses sort.SliceStable) |
| Go stdlib `errors` | 1.26.3 bundled | `errors.New` for sentinels, `errors.Join` for D-03/D-05/D-06 accumulation, `errors.Is`/`errors.As` for consumer discrimination | Spec §6 + go-coding-standards |
| Go stdlib `fmt` | 1.26.3 bundled | `fmt.Errorf("...: %w", ErrX)` index-wrapping for the three validation paths | Idiomatic Go |
| `github.com/axonops/fuzzymatch` | local | `*Scorer`, `AlgoID`, `Normalise`, `Tokenise`, `ErrInternalInvariantViolated` | Phase 8 + 8.5 |

### Test deps (already in `tests/bdd/go.mod`)

| Library | Version | Purpose | When to Use |
|---------|---------|---------|-------------|
| `github.com/cucumber/godog` | v0.15.0 | BDD scenarios (scan.feature + suppression.feature) | TEST-05; per-Phase-8 pattern |
| `github.com/stretchr/testify` | v1.10.0 | Assertion sugar inside `tests/bdd/steps/scan_steps.go` | Permitted ONLY in `tests/bdd/`; root `scan/*_test.go` uses stdlib `testing` |
| `go.uber.org/goleak` | v1.3.0 | Confirms no goroutine leaks (scan is pure-function) | Already wired via `goleak.VerifyTestMain` in `tests/bdd/bdd_test.go` — inherits automatically |

### Alternatives Considered

| Instead of | Could Use | Tradeoff |
|------------|-----------|----------|
| `map[string][]int` token-bucket | Sorted slice + binary search | Trade-off mentioned in CONTEXT.md Claude's Discretion: `map[string][]int` per spec §12.5 is the established choice; equivalence to naive is the load-bearing gate. Plan stays with spec. |
| `fmt.Errorf` index-wrapped errors | `*InvalidItemError` struct with `.Index int` field | CONTEXT.md Claude's Discretion — either acceptable. Recommendation: `fmt.Errorf` form (simpler, idiomatic Go, no extra type to document with 4-section sentinel godoc); the structured form is over-engineering for an internal index. |

**Installation:** Nothing to install — all dependencies are stdlib or already pinned. The new sub-package directory:

```bash
mkdir -p scan
# No new go.mod — scan/ lives under the root module per spec §12.7
```

**Version verification:** N/A — no new external versions to pin. Confirmed `cucumber/godog@v0.15.0`, `testify@v1.10.0`, `goleak@v1.3.0` are current Phase 8 versions and remain in `tests/bdd/go.mod`.

## Architecture Patterns

### System Architecture Diagram

```
                            ┌────────────────────────────┐
  consumer code             │  scan.Check(items, cfg)    │
  ──────────────────────────►  (single public entry)     │
                            └──────────────┬─────────────┘
                                           │
                                           ▼
                         ┌─────────────────────────────────┐
                         │  Pre-flight validation (P1..P4) │
                         │  P1: cfg.Scorer == nil          │  Fail-fast
                         │  P2: Config field validation    │  ──────►  ErrNilScorer
                         │      (boost range, NaN/Inf)     │           ErrInvalidConfig
                         │  P3: Items[] validation         │  Collect-all
                         │      - empty Name (D-03)        │  ──────►  errors.Join of
                         │      - dup (Name,Group) (D-06)  │           ErrInvalidItem wraps
                         │  P4: SuppressedPairs validation │  Collect-all
                         │      (D-05)                     │  ──────►  errors.Join of
                         └──────────────┬──────────────────┘           ErrInvalidConfig wraps
                                        │ (validated)
                                        ▼
                         ┌─────────────────────────────────┐
                         │  Normalise + Tokenise pass      │
                         │  - Normalise each Item.Name once│
                         │    using cfg.Scorer's opts      │
                         │  - Normalise each SuppressedPairs│
                         │    entry once                   │
                         │  - Tokenise each Item.Name once │
                         │    using DefaultTokeniseOptions │
                         └──────────────┬──────────────────┘
                                        │
                                        ▼
                         ┌─────────────────────────────────┐
                         │  Build group → items map        │
                         │  group → []itemIndex (sorted)   │
                         └──────────────┬──────────────────┘
                                        │
                                        ▼
                         ┌─────────────────────────────────┐
                         │  Within-group pass              │  per group:
                         │  for g in groups (sorted):      │  - len ≤ bucketThreshold:
                         │    if len(items[g]) ≤ 50:       │      naive O(N²) nested loops
                         │      naive nested loop          │  - len > bucketThreshold:
                         │    else:                        │      token-bucket candidates
                         │      bucket-filtered candidates │
                         │    for each candidate pair:     │  scorer.Match against
                         │      apply suppression checks   │  scorer.Threshold()
                         │      scorer.Match → emit?       │
                         └──────────────┬──────────────────┘
                                        │
                                        ▼ (when CompareAcrossGroups)
                         ┌─────────────────────────────────┐
                         │  Cross-group pass               │
                         │  effective threshold =          │
                         │    min(1.0, scorer.Threshold()  │
                         │           + cfg.CrossGroupBoost)│  cross-group threshold
                         │  bucket-filtered cross-group    │  applied to scorer.Score
                         │    candidates                   │  (cannot call scorer.Match
                         │  for each candidate pair:       │   directly — that uses the
                         │    apply suppression checks     │   within-group threshold)
                         │    if score >= effThreshold:    │
                         │      emit warning               │
                         └──────────────┬──────────────────┘
                                        │
                                        ▼
                         ┌─────────────────────────────────┐
                         │  Sort warnings by               │
                         │  (Kind, NameA, NameB,           │  sort.SliceStable
                         │   GroupA, GroupB)               │
                         └──────────────┬──────────────────┘
                                        │
                                        ▼
                         ┌─────────────────────────────────┐
                         │  In-line completeness assertion │
                         │  linear scan of adjacent pairs; │  panic with
                         │  if same 5-tuple key:           │  fuzzymatch.ErrInternal
                         │    panic(ErrInternalInvariant…) │  InvariantViolated
                         └──────────────┬──────────────────┘
                                        │
                                        ▼
                                  []scan.Warning
```

### Recommended Project Structure

Per spec §12.7 + adjustments for CONTEXT.md decisions:

```
scan/
├── scan.go                       # public surface: Item, Config, Warning, Kind, Check, DefaultConfig
├── kind.go                       # Kind type + KindWithinGroup + KindAcrossGroups + String()
├── errors.go                     # 3 sentinels with 4-section godoc per Phase 8.5 Q4 template
├── validate.go                   # internal pre-flight (P1..P4) — all errors.Join paths
├── bucket.go                     # token-bucket implementation + bucketThreshold private const
├── suppress.go                   # suppression composition (D-07 rules; precedence engine)
├── doc.go                        # package documentation (godoc landing page)
├── scan_test.go                  # external/black-box unit tests (package scan_test)
├── scan_internal_test.go         # bucket optimisation correctness, unexported helpers
├── scan_bench_test.go            # benchmarks incl. BenchmarkScanCheck_BucketVsNaive_GroupSize
├── example_test.go               # godoc runnable examples (one per public func per DX-02)
├── props_test.go                 # property tests (PropCheck_BucketEquivalentToNaive,
│                                 #                 PropCheck_DeterministicAcrossRuns)
└── fuzz_test.go                  # FuzzCheck — opaque-input fuzzer
```

**Why a `kind.go` split:** `scan.Kind` is a 3-line type + 5-line String() + 5-line two-constant block. Splitting it into a dedicated file mirrors `warn_kind.go` (root pkg) and isolates the SPEC OVERRIDE rename target for grep-ability ("where does `Kind` live?" → `kind.go`).

**Why a `validate.go` split:** the pre-flight has three coupled `errors.Join` accumulators (D-03 / D-05 / D-06). Keeping it in one file with one entry point (`validateCheck(items, cfg)`) makes the "uniform collect-all" pattern legible and testable in isolation.

**Why a `suppress.go` split:** SCAN-03 + SCAN-04 + the Phase 8.5 R2 deferred BDD scenarios touch a self-contained predicate (`isSuppressed(pair, scopeKind, cfg)`). Carving it out keeps `scan.go`'s main loop focused on iteration and lets the BDD step definitions in `scan_steps.go` exercise the predicate directly.

### Pattern 1: SPEC OVERRIDE in lockstep (foundation commit)

**What:** When CONTEXT.md decisions diverge from `docs/requirements.md`, the planner's first plan amends both the spec and the code in the same commit, with a `SPEC OVERRIDE (Phase 9)` note in the affected spec section and a back-reference to 09-CONTEXT.md.

**When to use:** Any plan that introduces a new public type / signature that contradicts the spec.

**Example (carry-forward from Phase 8 Plan 08-03 / Phase 7 Plan 07-02):**

```go
// scan/scan.go
// Warning carries the metadata for a detected similar-name pair.
//
// SPEC OVERRIDE (Phase 9): Scores is map[fuzzymatch.AlgoID]float64;
// docs/requirements.md §12.1 originally specified map[string]float64.
// See 09-CONTEXT.md §1 D-01 for the rationale (typed enum keys give
// compile-time safety; the Phase 8 ScoreAll override at §8.3 + §8.6
// applies verbatim here). The spec section was amended in the same
// commit; api-ergonomics-reviewer sign-off is recorded in the PR.
type Warning struct {
    Kind                   Kind
    NameA, NameB           string
    GroupA, GroupB         string
    TagA, TagB             any
    Score                  float64
    Scores                 map[fuzzymatch.AlgoID]float64
}
```

```text
# docs/requirements.md §12.1 (after amendment)
type Warning struct {
    Kind           Kind
    NameA, NameB   string
    GroupA, GroupB string
    TagA, TagB     any
    Score          float64

    // SPEC OVERRIDE (Phase 9): Scores is map[AlgoID]float64; see
    // 09-CONTEXT.md §1 D-01. The Phase 8 ScoreAll override at §8.3
    // + §8.6 applies verbatim here for the same typed-enum-keys
    // rationale.
    Scores map[AlgoID]float64
}
```

### Pattern 2: `errors.Join` collect-all validation

**What:** Walk the full input slice, accumulate every offending index into a slice of wrapped errors, return a single `errors.Join(errs...)`. Caller's `errors.Is(err, scan.ErrInvalidItem)` returns true because `errors.Is` walks `Unwrap() []error` on the joined value.

**When to use:** D-03 (empty Item.Name), D-05 (empty SuppressedPairs entry), D-06 (duplicate (Name,Group) Items).

**Example:**

```go
// scan/validate.go
func validateItems(items []Item) error {
    var errs []error
    seen := make(map[itemKey]int, len(items))

    for i, item := range items {
        // D-03: empty Name
        if item.Name == "" {
            errs = append(errs, fmt.Errorf(
                "scan: invalid item at index %d: empty name: %w",
                i, ErrInvalidItem,
            ))
            continue // don't double-flag a same-index duplicate
        }
        // D-06: duplicate (Name, Group)
        k := itemKey{Name: item.Name, Group: item.Group}
        if first, ok := seen[k]; ok {
            errs = append(errs, fmt.Errorf(
                "scan: invalid item at index %d: duplicate (Name, Group) of index %d: %w",
                i, first, ErrInvalidItem,
            ))
            continue
        }
        seen[k] = i
    }

    // Sort errs by index ASCENDING before joining so the joined message
    // walks indices in order — map iteration is contained within the
    // duplicate-detection pass and never reaches the output path.
    return errors.Join(errs...)
}
```

**Critical:** `errors.Is(err, scan.ErrInvalidItem)` works against the joined error because the standard library's `errors.Is` checks `Unwrap() []error` (added in Go 1.20). Documented in the sentinel's 4-section godoc.

### Pattern 3: Suppression composition predicate

**What:** A pure boolean predicate that takes a candidate pair, the pair's Kind, and the (already-normalised) Config, returning true if any of the three suppression rules apply.

**When to use:** Called once per candidate pair, after the threshold check passes but before emission. Cheap (3 boolean OR-ed terms; the most expensive is map lookup against the pre-normalised SuppressedPairs set).

**Example:**

```go
// scan/suppress.go
type suppressionCtx struct {
    suppressedPairs map[pairKey]struct{} // canonical-pair → struct{} — built once at Check entry
    compareIdenticalAcrossGroups bool
}

func isSuppressed(a, b Item, kind Kind, normalisedNameA, normalisedNameB string, sc suppressionCtx) bool {
    // Rule 1: per-item SilenceLint on either side (one-side suppression)
    if a.SilenceLint || b.SilenceLint {
        return true
    }

    // Rule 2: pair in SuppressedPairs (canonical-pair lookup)
    if _, ok := sc.suppressedPairs[canonicalPair(normalisedNameA, normalisedNameB)]; ok {
        return true
    }

    // Rule 3: cross-group identical-name default
    if kind == KindAcrossGroups && normalisedNameA == normalisedNameB && !sc.compareIdenticalAcrossGroups {
        return true
    }

    return false
}

// canonicalPair returns the lexicographically-sorted tuple (lo, hi) so
// SuppressedPairs lookups are order-independent.
func canonicalPair(a, b string) pairKey {
    if a <= b {
        return pairKey{a, b}
    }
    return pairKey{b, a}
}
```

### Pattern 4: Naive ↔ bucket dispatch by group size

**What:** Per-group decision tree; below `bucketThreshold` (50) use naive nested loops; above, use the token-bucket optimisation. Decision lives in one place (`scan.go`), bucket logic in `bucket.go`.

**When to use:** Both the within-group pass and the cross-group pass (cross-group treats the entire item set as one effective group for bucketing).

```go
// scan/scan.go — within-group dispatch
for _, group := range sortedGroups {
    indices := groupIndices[group]
    if len(indices) <= bucketThreshold {
        emitFromNaive(indices, normalisedNames, scorer, threshold, &warnings, ...)
    } else {
        emitFromBucket(indices, tokenBucket, normalisedNames, scorer, threshold, &warnings, ...)
    }
}
```

### Pattern 5: In-line completeness assertion (post-sort)

**What:** Linear scan of adjacent sorted warnings; if any pair shares the full 5-tuple sort key, panic with `fuzzymatch.ErrInternalInvariantViolated`. With D-06 rejecting duplicate (Name,Group) at validation time, this panic is unreachable in correct usage; it exists as a defence-in-depth gate per SCAN-05.

```go
// scan/scan.go — after sort.SliceStable
for i := 1; i < len(warnings); i++ {
    if warningsEqualOnSortKey(warnings[i-1], warnings[i]) {
        panic(fmt.Errorf(
            "%w: scan output has duplicate sort key at index %d (Kind=%s, NameA=%q, NameB=%q, GroupA=%q, GroupB=%q) — Items[] validation should have prevented this",
            fuzzymatch.ErrInternalInvariantViolated,
            i, warnings[i].Kind, warnings[i].NameA, warnings[i].NameB,
            warnings[i].GroupA, warnings[i].GroupB,
        ))
    }
}
```

### Anti-Patterns to Avoid

- **`map[string]float64` for Warning.Scores** — violates D-01. Use `map[fuzzymatch.AlgoID]float64`.
- **`scan.WarningKind` or `scan.WarnKind` type name** — violates D-02. Use `scan.Kind`.
- **Fail-fast on first invalid index** — violates D-03 / D-05 / D-06. Accumulate every offending index into a slice, return `errors.Join`.
- **Magic-zero default for `CrossGroupThresholdBoost`** — violates D-04. Zero-value = 0.0 (no boost); the 0.05 default lives in `DefaultConfig`.
- **`Config.BucketThreshold` public field** — violates D-08. `bucketThreshold` is a private package constant.
- **`scan.Check` calling `fuzzymatch.Validate` internally** — violates D-07. Two clean APIs; consumer composes them.
- **Map iteration on output path** — violates DET-03 + spec §12.4. The token-bucket map is consumed in input-order to build candidate sets; candidate sets are sorted before scoring.
- **Re-normalising inside `scan.Check`** — the Scorer already normalises internally per Phase 8 §3. `scan.Check` must call `fuzzymatch.Normalise(name, opts)` ONCE to build the bucket / suppression-pair index, then pass the **raw** name strings to `scorer.Score` / `scorer.Match` (which re-normalises). Double-normalisation is wasteful, not incorrect — but it's still wrong. See Open Question 1 for the access path to the Scorer's normalisation opts.
- **Direct `scorer.Match` for cross-group pass** — `Match` uses the within-group threshold. Cross-group must call `scorer.Score(a, b) >= effectiveCrossGroupThreshold` instead, where `effectiveCrossGroupThreshold = min(1.0, scorer.Threshold() + cfg.CrossGroupThresholdBoost)`. See spec §12.2 line 1413.

## Don't Hand-Roll

| Problem | Don't Build | Use Instead | Why |
|---------|-------------|-------------|-----|
| Validation-error accumulation | Custom multi-error type with custom `Is()` walker | `errors.Join` from stdlib | Go 1.20+ idiom; `errors.Is`/`errors.As` walk `Unwrap() []error` automatically |
| Stable sort of `[]Warning` | hand-rolled mergesort or quicksort | `sort.SliceStable` (already used in `validate.go`) | Stable, well-tested, deterministic across platforms; the project's existing pattern |
| Canonical-pair normalisation | Custom hashing or canonical-form encoding | Lexicographic min/max + struct key in `map[struct{A, B string}]struct{}` | Two string compares + a struct hash; one allocation per SuppressedPairs entry at Check entry |
| Token-bucket candidate-set union | hand-rolled union-find or bloom filter | `map[int]struct{}` per source item; iterate the source item's tokens and add `bucket[token]` to the set | Spec §12.5 explicitly endorses this; equivalence proven by `PropCheck_BucketEquivalentToNaive` |
| Tokenisation | re-implement camelCase / snake_case splitting | `fuzzymatch.Tokenise(name, fuzzymatch.DefaultTokeniseOptions())` | Phase 8.5 Q8b gave us the ASCII fast path; reusing is mandatory for the < 2s/10k-items PERF-05 budget |
| Normalisation | re-implement case-fold / NFD / diacritic strip | `fuzzymatch.Normalise(name, opts)` with the Scorer's options | See Open Question 1 — the access path is the open issue, not whether to reuse |
| Float-determinism reduction | hand-rolled scoring loop | `scorer.Score(a, b)` / `scorer.ScoreAll(a, b)` | Phase 8's FMA-defeating double-cast lives in `scorer.go:380`; scan defers to it |

**Key insight:** The scan sub-package is **scaffolding around Phase 8** — pre-flight validation, candidate enumeration, suppression filtering, output ordering. Every line of float arithmetic or string-similarity logic should be a delegated call into the root package. The risk of hand-rolling is highest in the candidate-enumeration tier (the bucket) where the temptation to "optimise just a little more" can introduce determinism bugs; the property test gate is non-negotiable here.

## Common Pitfalls

### Pitfall 1: Map iteration leaking into output ordering

**What goes wrong:** A naïve implementation iterates `map[group][]int` directly to emit warnings group-by-group. Go map iteration is randomised, so `scan.Check` returns warnings in a different order each run, breaking `PropCheck_DeterministicAcrossRuns` and `testdata/golden/scan-default.json`.

**Why it happens:** The token-bucket and the group-index are natural map data structures. Forgetting to sort the keys (or sort the final output) is the universal Go determinism mistake.

**How to avoid:** Build a `groups []string` sorted slice at Check entry; iterate that. Build candidate-set slices from map values, sort them before the scoring loop. Final `sort.SliceStable` on warnings before return.

**Warning signs:** `PropCheck_DeterministicAcrossRuns` produces different outputs on consecutive runs; CI golden file diff fails on linux/arm64 vs darwin/arm64.

### Pitfall 2: Sort-key incompleteness leading to non-deterministic stable-sort ties

**What goes wrong:** Two warnings end up with identical 5-tuple sort keys. `sort.SliceStable` preserves *input order* on ties, but the input order depends on map iteration (back to Pitfall 1). Determinism evaporates.

**Why it happens:** Edge cases: same-name items in different positions of the input slice; cross-group pass producing a (X-in-G1, X-in-G2) warning where another consumer has (X-in-G1, X-in-G3) sharing the same NameA / NameB — sort key collision.

**How to avoid:** D-06 rejects duplicate (Name, Group) at validation time — provably uniquifies the (Name, Group) keys per item. Cross-group sort-key collisions are still possible across DIFFERENT name pairs because GroupA / GroupB participate in the key; the sort key `(Kind, NameA, NameB, GroupA, GroupB)` is complete by construction once D-06 holds.

**Warning signs:** The in-line completeness assertion (Pattern 5) panics with `ErrInternalInvariantViolated`. The user-memory `feedback_run_tests_locally_first.md` applies — exercise the assertion locally before pushing.

### Pitfall 3: `scorer.Match` for cross-group pass (wrong threshold)

**What goes wrong:** `scorer.Match(a, b)` uses the within-group threshold. The cross-group pass must use `min(1.0, threshold + boost)`. Calling `Match` for cross-group emits noise (over-matching).

**Why it happens:** `Match` is the obvious "is this a hit?" API. The two-threshold dual is a scan-specific concept the spec introduces.

**How to avoid:** Cross-group pass always computes `scorer.Score(a, b)` (raw composite) and compares against the locally-computed `effectiveCrossGroupThreshold`. Document in the inner-loop comment block.

**Warning signs:** `scan.feature` "cross-group threshold boost is applied" scenario fails; identifier-style cross-group pairs that should be silenced still emit warnings.

### Pitfall 4: SuppressedPairs normalisation mismatch

**What goes wrong:** Consumer passes `SuppressedPairs: [["user_id", "userId"]]`. Inside `Check`, the canonical-pair lookup is built from the post-normalisation Items, but `SuppressedPairs` entries are stored raw. The lookup misses because `"user_id"` (raw) != `"userid"` (normalised).

**Why it happens:** Forgetting that the Items go through normalisation but the SuppressedPairs surface is raw consumer input.

**How to avoid:** At the top of `Check`, after validation, normalise every SuppressedPairs entry using the Scorer's normalisation options, then build the canonical-pair set from the normalised forms. Document in `SuppressedPairs` godoc + `docs/scan.md`. The spec covers this at line 1423: "SuppressedPairs entries are normalised at the start of Check using the Scorer's normalisation options."

**Warning signs:** BDD scenario "SuppressedPairs silences canonical pair regardless of input casing" fails. Consumer reports "I added the pair to SuppressedPairs but it still emits."

### Pitfall 5: Double-normalisation cost

**What goes wrong:** `scan.Check` calls `fuzzymatch.Normalise(item.Name, opts)` to build the bucket, then passes the **normalised** name to `scorer.Score`, which calls `Normalise` again internally (Phase 8 §3). Doubles the normalisation cost.

**Why it happens:** Natural reflex — "I have the normalised string already; pass it." But the Scorer's normalisation is a contract: callers pass raw input, the Scorer normalises. Bypassing breaks the boundary.

**How to avoid:** Inside `scan.Check`, keep two parallel arrays: `rawNames[i] = items[i].Name` (passed to `scorer.Score`) and `normalisedNames[i] = Normalise(items[i].Name, opts)` (used for bucket keys, SuppressedPairs lookup, identical-name suppression check). Call sites are explicit; no normalised string ever crosses the Scorer boundary.

**Warning signs:** Benchmark `BenchmarkScanCheck_DefaultScorer_10k` runs ~2× slower than expected; profile shows `Normalise` consuming ~50% of CPU.

### Pitfall 6: Forgetting D-04's boost clamp at 1.0

**What goes wrong:** Consumer sets `cfg.CrossGroupThresholdBoost = 0.5` against a Scorer with `Threshold() = 0.85`. Naive arithmetic gives `0.85 + 0.5 = 1.35`. If `effectiveCrossGroupThreshold` exceeds 1.0, no pair ever matches (composite is in [0,1]) — effectively disabling cross-group emission silently. Spec §12.2 line 1413 says this is *intentional* and documented behaviour: "clamped to 1.0 (meaning only byte-identical matches pass)."

**Why it happens:** The validation pass (D-04) rejects boost > 1.0, but valid (boost = 0.5) + Scorer threshold (0.85) can still cross 1.0.

**How to avoid:** `effectiveCrossGroupThreshold = math.Min(1.0, scorer.Threshold() + cfg.CrossGroupThresholdBoost)`. Document the clamp in `Config.CrossGroupThresholdBoost` godoc. BDD scenario covers the boundary case.

**Warning signs:** Cross-group pass emits zero warnings even on obviously-similar items.

### Pitfall 7: Tokeniser allocation on every Check call

**What goes wrong:** `Tokenise` is called inside the per-pair loop instead of pre-computed once per Item.

**Why it happens:** Forgetting that bucket construction is a one-time cost.

**How to avoid:** Build `tokenisedNames [][]string` once at Check entry, indexed by item-index. The bucket build pass reads from this slice. Documented in CONTEXT.md "Claude's Discretion": each Item.Name tokenised at most once per Check invocation.

**Warning signs:** `BenchmarkScanCheck_*` shows alloc count linear in pair-count rather than item-count.

### Pitfall 8: Self-pair emission

**What goes wrong:** The inner loop emits warnings for (i, i) pairs.

**Why it happens:** Naive nested loop forgets the `j > i` constraint.

**How to avoid:** Inner loop is always `for j := i+1; j < N; j++` (within-group) or restricted to cross-group candidate sets that exclude self by construction. D-05's silent-keep-of-self-SuppressedPairs is harmless because self-pairs never reach the suppression check.

**Warning signs:** Property test `PropCheck_NoSelfWarnings` fails (add this property test alongside the existing ones).

## Runtime State Inventory

Phase 9 is greenfield code addition (no rename, refactor, migration, or string-replacement). **Section omitted** — no stored data, live service config, OS-registered state, secrets/env vars, or build artefacts carry the names of new types. Spec amendments in `docs/requirements.md` are normal in-tree text edits.

## Validation Architecture

`workflow.nyquist_validation == true` in `.planning/config.json` — this section is REQUIRED.

### Test Framework

| Property | Value |
|----------|-------|
| Framework | stdlib `testing` (root tests); `cucumber/godog v0.15.0` + `testify v1.10.0` + `goleak v1.3.0` (BDD) |
| Config file | none — `go test ./scan/...` is sufficient at root; `tests/bdd/go.mod` is the BDD harness module |
| Quick run command | `go test -race -shuffle=on -count=1 ./scan/...` (root scan unit + property + fuzz seed) |
| Full suite command | `make check && make test-bdd && make verify-determinism` |

### Phase Requirements → Test Map

| Req ID | Behavior | Test Type | Automated Command | File Exists? |
|--------|----------|-----------|-------------------|-------------|
| SCAN-01 | `Check(items, cfg) []Warning` within + cross passes | unit + BDD | `go test -run TestCheck_WithinGroup ./scan/... && cd tests/bdd && go test -run TestBDDSuite -godog.tags @scan ./...` | ❌ Wave 0 |
| SCAN-02 | Bucket equivalent to naive | property + bench | `go test -run TestProp_Check_BucketEquivalentToNaive ./scan/...` + `go test -bench BenchmarkScanCheck_BucketVsNaive_GroupSize ./scan/...` | ❌ Wave 0 |
| SCAN-03 | SilenceLint + SuppressedPairs composition | unit + BDD | `go test -run TestCheck_Suppression ./scan/... && cd tests/bdd && go test -run TestBDDSuite -godog.tags @suppression ./...` | ❌ Wave 0 |
| SCAN-04 | Cross-group identical-name default | unit + BDD | `go test -run TestCheck_IdenticalCross ./scan/... && cd tests/bdd && go test -run TestBDDSuite -godog.tags @suppression ./...` | ❌ Wave 0 |
| SCAN-05 | Deterministic sort + completeness | property + golden | `go test -run "TestProp_Check_DeterministicAcrossRuns\|TestCheck_SortKeyCompleteness\|TestGolden_ScanDefault" ./scan/...` | ❌ Wave 0 |
| SCAN-06 | Sentinel error hierarchy | unit (errors.Is) | `go test -run TestCheck_ErrInvalidItem.*\|TestCheck_ErrNilScorer\|TestCheck_ErrInvalidConfig ./scan/...` | ❌ Wave 0 |
| PERF-05 | < 2s for 10,000 items | benchmark + visual | `go test -bench BenchmarkScanCheck_DefaultScorer_10k -benchtime=3x ./scan/...` (recorded; not a CI gate per Makefile bench-informational policy) | ❌ Wave 0 |
| DET-04 | NaN/Inf/-0 handling | property | `go test -run "TestProp_Check_NoNaN\|TestProp_Check_NoInf" ./scan/...` (delegates to Scorer; trivially holds because scan does no float arithmetic) | ❌ Wave 0 |
| TEST-05 | BDD scope incl. scan | godog | `cd tests/bdd && go test ./...` (entire suite incl. scan.feature + suppression.feature) | ❌ Wave 0 (feature files) |
| DX-02 | godoc Example per public func | example test | `go test -run "Example.*" ./scan/...` | ❌ Wave 0 |
| DX-04 | docs/scan.md populated | manual review | `markdownlint-cli2 docs/scan.md && grep -L 'TBD' docs/scan.md` | scaffold present; needs population |

### Sampling Rate

- **Per task commit:** `go test -race -shuffle=on -count=1 ./scan/...` (quick run); the changed file's test target if granular.
- **Per wave merge:** `make check && make test-bdd` (full root + BDD suite).
- **Phase gate:** `make check && make test-bdd && make verify-determinism` ALL green on origin/main CI before the verification gate (per user-memory `feedback_ci_before_verification_gate.md`). Verify via `gh run list` post-push; do NOT approve the gate while origin CI is red.

### Wave 0 Gaps

The scan sub-package does not yet exist. ALL test infrastructure is created by Phase 9 plans:

- [ ] `scan/scan.go` — public surface (Item, Config, Warning, Kind, Check, DefaultConfig)
- [ ] `scan/kind.go` — Kind type + constants + String()
- [ ] `scan/errors.go` — three sentinels with 4-section godoc
- [ ] `scan/validate.go` — internal pre-flight (P1..P4) collect-all errors.Join
- [ ] `scan/bucket.go` — token-bucket + bucketThreshold const
- [ ] `scan/suppress.go` — suppression composition predicate
- [ ] `scan/doc.go` — package documentation
- [ ] `scan/scan_test.go` — black-box unit tests
- [ ] `scan/scan_internal_test.go` — bucket correctness + suppress predicate
- [ ] `scan/scan_bench_test.go` — bench incl. `BenchmarkScanCheck_BucketVsNaive_GroupSize` (D-08) + `BenchmarkScanCheck_DefaultScorer_10k` (PERF-05)
- [ ] `scan/example_test.go` — runnable godoc examples (≥ 6: Check_WithinGroup, Check_CrossGroup, Check_WithSuppression, Check_WithDefaultConfig, DefaultConfig_DocsExample, Kind_String)
- [ ] `scan/props_test.go` — property tests (`PropCheck_BucketEquivalentToNaive`, `PropCheck_DeterministicAcrossRuns`, `PropCheck_NoSelfWarnings`, `PropCheck_NoNaN`, `PropCheck_NoInf`)
- [ ] `scan/fuzz_test.go` — `FuzzCheck` with seed corpus
- [ ] `testdata/golden/scan-default.json` — golden corpus
- [ ] `testdata/golden/_staging/scan.json` — staging file pattern (per existing staging convention)
- [ ] `scan_golden_test.go` (root pkg OR `scan/scan_golden_test.go`) — golden-file diff test
- [ ] `tests/bdd/features/scan.feature` — BDD scenarios (within/cross/threshold-boost/sort)
- [ ] `tests/bdd/features/suppression.feature` — BDD scenarios (SilenceLint, SuppressedPairs, identical-cross default, composition)
- [ ] `tests/bdd/steps/scan_steps.go` — step definitions; `InitScanSteps(ctx)` wired into `algorithms_steps.go`'s `InitializeScenario`
- [ ] `examples/scan-demo/main.go` + `main_test.go` + `go.mod` — runnable demo + stdout meta-test
- [ ] `docs/scan.md` — populated from scaffold (currently 5-section TBD scaffold)
- [ ] `bench.txt` — extended with scan benchmarks
- [ ] `llms.txt` + `llms-full.txt` — synced for every new exported symbol; `make verify-llms-sync` green

## Security Domain

### Applicable ASVS Categories

| ASVS Category | Applies | Standard Control |
|---------------|---------|-----------------|
| V2 Authentication | no | scan is a pure-function library; no identity context |
| V3 Session Management | no | no session state |
| V4 Access Control | no | no access boundaries; consumer enforces |
| V5 Input Validation | yes | D-03 (empty Name), D-05 (empty SuppressedPairs), D-06 (duplicate (Name,Group)), D-04 (boost NaN/Inf/range). Three errors.Join paths + one fail-fast Scorer-nil check. |
| V6 Cryptography | no | no crypto |
| V7 Error Handling | yes | Three sentinels with 4-section godoc; `errors.Is`/`errors.Join` discrimination; in-line completeness panic with typed error wrapping ErrInternalInvariantViolated; no panic on consumer-supplied input |
| V9 Communication | no | no I/O |
| V12 Resource Files | no | no file I/O |

### Known Threat Patterns for scan sub-package

| Pattern | STRIDE | Standard Mitigation |
|---------|--------|---------------------|
| Adversarial input causing O(N²) cost (10,000 identically-tokenised items would defeat the bucket and trigger naive comparison via the bucket's degenerate "all candidates" case) | Denial of Service | (a) `validatePathologicalThreshold = 65,536` already applies at the per-pair `Validate` surface (root pkg); scan inherits this implicitly. (b) PERF-05 budget < 2s/10k is the empirical ceiling; document worst-case complexity in `Check` godoc + `docs/scan.md`. (c) Adversarial-input warning lives in `docs/scan.md` consumer guidance section. |
| Resource exhaustion via SuppressedPairs explosion (consumer passes 1M-entry SuppressedPairs) | Denial of Service | Lookup is O(1) per candidate via map; build cost is O(N) — linear in SuppressedPairs len. Document the linear-build cost in `Config.SuppressedPairs` godoc. No hard limit imposed; consumer responsibility. |
| Tag field carrying sensitive data leaking into errors | Information Disclosure | `Tag any` is consumer-supplied; scan never stringifies it inside error messages or panics. Error wraps include `Items[i].Name` (already consumer-supplied; same surface) and indices — Tag never appears. Documented in `Item.Tag` godoc. |
| Panic in inner loop killing the consumer's process | Tampering / DoS | scan never panics on consumer-supplied input. The single panic site (in-line completeness assertion) is gated by D-06 + the validation pre-flight; reachable only on a library bug (`ErrInternalInvariantViolated`). Consumer can recover and report via `errors.Is(err, fuzzymatch.ErrInternalInvariantViolated)`. |
| Integer overflow on duplicate-index error message | Tampering | Indices are `int`; bounded by `len(items) <= MaxInt`. Practical inputs cannot exceed `int` range. No mitigation needed. |
| Map iteration leaking input position via output order | Information Disclosure (weak) | DET-03 + spec §12.4 forbid map iteration on output paths. Sorted output. Pre-emptive control. |

**STRIDE summary:** scan is a deterministic pure function over consumer-supplied data. The threat surface is dominated by DoS (input size + complexity) and one Information Disclosure pattern (Tag in error messages, ruled out by design). The existing project-wide threat model from Phase 8 / Phase 8.5 fully covers the Scorer side of the cost; scan adds the candidate-enumeration layer with the bucket + naive dispatch.

## Code Examples

Verified patterns from existing project code:

### Sentinel error declaration (4-section godoc per Phase 8.5 Q4 template)

```go
// scan/errors.go — pattern copied verbatim from errors.go (root pkg)
//
// ErrNilScorer indicates Config.Scorer was nil when Check was invoked.
// A nil Scorer has no algorithms, no threshold, and no normalisation
// options — there is nothing for Check to do.
//
// Common causes: forgetting to set Config.Scorer before calling Check;
// passing a freshly-allocated zero-value Config struct.
//
// Resolution: set cfg.Scorer to a non-nil *fuzzymatch.Scorer (typically
// from fuzzymatch.DefaultScorer() or fuzzymatch.NewScorer(...)). For
// the opinionated default Config + Scorer composition, use
// scan.DefaultConfig(fuzzymatch.DefaultScorer()).
//
// Example:
//
//	_, err := scan.Check(items, scan.Config{Scorer: nil})
//	if errors.Is(err, scan.ErrNilScorer) { /* diagnostic */ }
var ErrNilScorer = errors.New("scan: Config.Scorer is required")
```

### Kind type + CamelCase String (D-02 + Phase 8.5 Q6b)

```go
// scan/kind.go — pattern copied from warn_kind.go (root pkg)

package scan

import "fmt"

// Kind classifies a scan.Warning's pair scope. SPEC OVERRIDE (Phase 9):
// type renamed from "WarningKind" to "Kind" per 09-CONTEXT.md §1 D-02
// — the package-scoped form is unambiguous at the call site
// (scan.KindWithinGroup).
type Kind int

const (
    KindWithinGroup Kind = iota + 1
    KindAcrossGroups
)

// String returns the CamelCase form. Matches the AlgoID.String() +
// WarnKind.String() convention locked in §6 + Phase 8.5 Q6b.
func (k Kind) String() string {
    switch k {
    case KindWithinGroup:
        return "WithinGroup"
    case KindAcrossGroups:
        return "AcrossGroups"
    default:
        return fmt.Sprintf("Kind(%d)", int(k))
    }
}
```

### DefaultConfig helper (D-04 — opinionated default lives here)

```go
// scan/scan.go — pattern modelled on DefaultScorerOptions in scorer.go:549

// DefaultConfig returns an opinionated Config bound to the supplied
// Scorer. The defaults are:
//
//   - CompareAcrossGroups: false (within-group only)
//   - CrossGroupThresholdBoost: 0.05 — applied only when the consumer
//     subsequently sets CompareAcrossGroups = true
//   - CompareIdenticalAcrossGroups: false (suppress identical names
//     across groups by default)
//   - SuppressedPairs: nil
//
// SPEC OVERRIDE (Phase 9): the 0.05 default lives here, NOT as the
// zero-value of Config.CrossGroupThresholdBoost; see 09-CONTEXT.md §2
// D-04. Spec §12.1 line 1359 was amended to reflect this.
//
// Typical usage:
//
//	s := fuzzymatch.DefaultScorer()
//	cfg := scan.DefaultConfig(s)
//	cfg.CompareAcrossGroups = true // opt in
//	warnings, err := scan.Check(items, cfg)
func DefaultConfig(s *fuzzymatch.Scorer) Config {
    return Config{
        Scorer:                       s,
        CompareAcrossGroups:          false,
        CrossGroupThresholdBoost:     0.05,
        CompareIdenticalAcrossGroups: false,
        SuppressedPairs:              nil,
    }
}
```

### Validation collect-all via errors.Join (D-03 / D-05 / D-06)

```go
// scan/validate.go

package scan

import (
    "errors"
    "fmt"
    "math"

    "github.com/axonops/fuzzymatch"
)

func validateCheck(items []Item, cfg Config) error {
    // P1: fail-fast on nil Scorer (cheapest check)
    if cfg.Scorer == nil {
        return ErrNilScorer
    }

    // P2: fail-fast on Config field violations
    if err := validateConfigFields(cfg); err != nil {
        return err
    }

    // P3: collect-all on Items (D-03 + D-06)
    if err := validateItems(items); err != nil {
        return err
    }

    // P4: collect-all on SuppressedPairs (D-05)
    if err := validateSuppressedPairs(cfg.SuppressedPairs); err != nil {
        return err
    }

    return nil
}

func validateConfigFields(cfg Config) error {
    b := cfg.CrossGroupThresholdBoost
    if math.IsNaN(b) || math.IsInf(b, 0) || b < 0.0 || b > 1.0 {
        return fmt.Errorf(
            "scan: CrossGroupThresholdBoost=%v is invalid (must be in [0.0, 1.0], finite, non-NaN): %w",
            b, ErrInvalidConfig,
        )
    }
    return nil
}

// (validateItems and validateSuppressedPairs follow Pattern 2.)
```

### BDD step registration (pattern from scorer_steps.go)

```go
// tests/bdd/steps/scan_steps.go

package steps

import (
    "github.com/cucumber/godog"
)

type ScanContext struct {
    // … per-scenario state …
    items    []scan.Item
    cfg      scan.Config
    warnings []scan.Warning
    err      error
}

func InitScanSteps(ctx *godog.ScenarioContext) {
    sc := &ScanContext{}
    ctx.Step(`^I scan items with the default config$`, sc.iScanWithDefaultConfig)
    ctx.Step(`^the scan should produce (\d+) warnings$`, sc.theScanShouldProduceNWarnings)
    // … etc …
}

// tests/bdd/steps/algorithms_steps.go — extend InitializeScenario at line ~1593:
//   InitScanSteps(ctx)
```

## State of the Art

| Old Approach | Current Approach | When Changed | Impact |
|--------------|------------------|--------------|--------|
| Spec §12.1's `WarningKind` type | `scan.Kind` (D-02) | Phase 9 CONTEXT 2026-05-19 | Spec amendment in foundation commit; cleaner type name; no functional change |
| Spec §12.1's `Scores map[string]float64` | `Scores map[fuzzymatch.AlgoID]float64` (D-01) | Phase 9 CONTEXT 2026-05-19 (extending Phase 8 Plan 08-03 ScoreAll override) | Compile-time-safe map keys; spec amendment in foundation commit |
| Spec §12.1's "wrap the offending item's index" (singular) | "every offending index, joined via errors.Join" (D-03/D-05/D-06) | Phase 9 CONTEXT 2026-05-19 | Consumer fixes whole batch in one round-trip; consistent across three validation paths |
| Spec §12.1's `CrossGroupThresholdBoost` "Default: 0.05" in struct godoc | Zero-value = 0.0; opinionated default in `scan.DefaultConfig()` (D-04) | Phase 9 CONTEXT 2026-05-19 | Mirrors Phase 8 `DefaultScorer` pattern; explicit construction defaults are visible |
| Spec §12.5's `bucketThreshold` = 50 (hardcoded, no override) | Private const + empirical validation gate (D-08) | Phase 9 CONTEXT 2026-05-19 | Same value, but the value is now load-bearing on a benchmark sweep; if benchmark falsifies 50, constant updates in the same plan |

**Deprecated / outdated:**

- Spec §12.1 line 1306–1322 — `WarningKind` type and its `String()` signature need rename to `Kind` (D-02).
- Spec §12.1 line 1337 — `Scores map[string]float64` needs replacement with `Scores map[AlgoID]float64` (D-01).
- Spec §12.1 line 1359 — "Default: 0.05" inside `Config.CrossGroupThresholdBoost` godoc needs migration to `DefaultConfig` godoc (D-04).
- Spec §12.1 line 1389 — "the offending item's index" needs to become "every offending index, joined via errors.Join" (D-03).
- Spec line 2010 (REQUIREMENTS row in §19 phase plan) and line 2061 (Scan sub-package acceptance criteria) — `WarningKind` references need rename to `Kind`.
- `docs/scan.md` — TBD scaffold needs full population (DX-04).

## Assumptions Log

| # | Claim | Section | Risk if Wrong |
|---|-------|---------|---------------|
| A1 | The Scorer's normalisation options (`s.normaliseOpts` + `s.applyNormalisation`) are required to normalise `SuppressedPairs` entries and to drive the bucket's token extraction. The Scorer's fields are unexported (`scorer.go:128`), so scan needs either (a) a new `Scorer.NormalisationOptions() NormalisationOptions` accessor, (b) the consumer to also pass the same opts via Config, or (c) scan always uses `DefaultNormalisationOptions()` (which would diverge from a Scorer constructed with `WithoutNormalisation()` or `WithNormalisation(custom)`). | Open Question 1; Pattern 3; Pattern 4 | Suppression composition diverges from Scorer normalisation, leading to silent SuppressedPairs misses. **HIGH risk.** See Open Question 1 below — this needs an explicit decision in Plan 09-01 before code lands. **[ASSUMED]** until the user / api-ergonomics-reviewer rules. |
| A2 | The Scorer's normalisation options are stable across the lifetime of a Check call (they are — Scorer is immutable post-NewScorer per Phase 8 §3) and reading them is safe from any goroutine (they are — same immutability). | Open Question 1 | None — Phase 8 LOCKED this. **[VERIFIED: scorer.go:60-66 godoc + scorer.go:115-128 struct fields]** |
| A3 | `Scorer.Threshold()` returns the within-group threshold; cross-group pass adds the boost in scan's local arithmetic. | Pattern 4; Pitfall 3 | None — Phase 8 LOCKED this. **[VERIFIED: scorer.go:447-449]** |
| A4 | `Scorer.Score(a, b) float64` is the right call for cross-group pass (must compute composite and apply local effective threshold, not call `Match` which uses the within-group threshold). | Pitfall 3 | Cross-group emits noise. **[VERIFIED: scorer.go:352-389 + scorer.go:398-400]** |
| A5 | `fuzzymatch.Normalise(s, opts)` returns the canonical lowercased / separator-stripped form ready for bucket keying and SuppressedPairs comparison. | Pattern 3; Pitfall 4 | Canonical-pair mismatch. **[VERIFIED: tokenise.go + Phase 8.5 Q8b ASCII fast path in tokenise.go]** |
| A6 | `fuzzymatch.Tokenise(s, fuzzymatch.DefaultTokeniseOptions())` is the right call for token-bucket extraction (matches the spec §12.5 step 1 "Tokenise every name once at the start of Check, using the Scorer's normalisation options"). | Pattern 4; Open Question 2 | Bucket keys mismatch consumer's expected tokens. **[ASSUMED]** — spec says "using the Scorer's normalisation options" but Tokenise takes `TokeniseOptions`, not `NormalisationOptions`. The two are separate option types; the spec is ambiguous. See Open Question 2. |
| A7 | The cross-platform CI matrix (linux/{amd64,arm64}, darwin/{amd64,arm64}, windows/amd64) covers the new `testdata/golden/scan-default.json` automatically because the existing `verify-determinism` Makefile target already iterates `testdata/golden/*.json`. | Determinism golden file; Validation Architecture | If the Makefile target hard-codes file names, scan-default.json won't be checked. Need to verify in Plan 09-06 / 09-08. **[ASSUMED]** until inspected. Risk: low — easy fix. |
| A8 | The empirical bucket-threshold sweep (D-08) runs on darwin/arm64 locally per the user's standard machine and the benchstat self-hosted runner is unavailable (per `.planning/STATE.md` Blockers/Concerns line 125). The crossover value committed should be the darwin/arm64 result; CI re-runs on the matrix are informational only. | Benchmark suite; Open Question 3 | If the crossover is genuinely architecture-specific (e.g. 30 on amd64, 70 on arm64), a single committed constant will under- or over-bucket on the other platform. The < 2s budget should hold either way; this is a tuning question, not a correctness question. **[ASSUMED]** — see Open Question 3. |
| A9 | The 8-plan decomposition below fits the file-ownership constraint that no two plans modify the same file. The foundation plan owns scan.go + kind.go + errors.go + spec + llms; validation owns validate.go; cross/within passes own scan.go's main loop body (Plan 09-03 creates Check, Plan 09-04 extends it for buckets, Plan 09-05 extends for suppression); golden + completeness owns scan.go's sort/assert tail. There's file-overlap on scan.go across Plans 09-03/04/05/06 — flag in the plan-checker review. | Plan Decomposition | If file-overlap is unacceptable per project convention, plans 09-03..06 must merge into a single larger plan (acceptable; reduces wave count from 6 to 5). **[ASSUMED]** — Phase 8's plan 08-02 / 08-03 / 08-04 all touched scorer.go (different methods), so partial file-overlap appears to be acceptable when methods are clearly disjoint. |

## Open Questions

### 1. Scorer's normalisation options access path — BLOCKS Plan 09-01

**What we know:** Spec §12.5 line 1437 says "Tokenise every name once at the start of `Check`, using the Scorer's normalisation options." Spec §12.3 line 1423 says "`SuppressedPairs` entries are normalised at the start of `Check` using the Scorer's normalisation options." Both statements require `scan.Check` to read the Scorer's `applyNormalisation` flag + `normaliseOpts NormalisationOptions` value. These fields are unexported (`scorer.go:115` + `scorer.go:127`).

**What's unclear:** How does `scan.Check` access them? Three options:

| Option | Description | Tradeoff |
|--------|-------------|----------|
| A | Add a public accessor: `func (s *Scorer) NormalisationOptions() (opts NormalisationOptions, applied bool)` | Mirrors the existing `Threshold()` and `Algorithms()` accessors. Single new exported method on the Phase 8 Scorer — minor surface addition. Most idiomatic. Spec amendment to `docs/requirements.md` §8 to add the accessor. |
| B | scan always uses `DefaultNormalisationOptions()` | Simpler scan implementation but DIVERGES from a Scorer constructed with `WithoutNormalisation()` or `WithNormalisation(custom)`. Suppression rules and bucket keys produced from `Normalise(s, defaultOpts)` would mismatch the Scorer's actual scoring — silent semantic bug. **REJECT.** |
| C | Require consumer to pass normalisation opts via `Config` (e.g. `Config.NormalisationOptions NormalisationOptions`) | Couples Scorer + Config; consumer has to keep them in sync. Violates DRY. Adds a Config field nobody asked for. **REJECT.** |

**Recommendation:** Option A — add `(s *Scorer) NormalisationOptions() (NormalisationOptions, bool)` (or `NormaliseOptions()` — `api-ergonomics-reviewer` picks the name). Land in Plan 09-01 alongside the type declarations + spec amendment. The accessor is a 5-line method returning two struct values; no behavioural change.

**Risk if wrong:** HIGH. Without this accessor, scan must either fork the spec (unilaterally choose `DefaultNormalisationOptions()`, silently diverging from `WithoutNormalisation` Scorers) or invent a coupling consumer field. Both choices propagate into BDD scenarios and the golden file. Get this resolved before any code lands.

**Action for the planner:** Surface as the FIRST item in the foundation plan's task list. If api-ergonomics-reviewer wants to take this through a mini-discuss-phase, do that before Plan 09-01 starts.

### 2. Spec §12.5 line 1437 "Tokenise … using the Scorer's normalisation options" — Tokenise takes TokeniseOptions, not NormalisationOptions

**What we know:** Spec §12.5 step 1 says "Tokenise every name once at the start of `Check`, using the Scorer's normalisation options." `fuzzymatch.Tokenise(s string, opts TokeniseOptions) []string` (per tokenise.go:179) takes `TokeniseOptions`, not `NormalisationOptions`. The two option types are separate (NormalisationOptions controls Normalise; TokeniseOptions controls Tokenise).

**What's unclear:** What the spec actually means by "using the Scorer's normalisation options" in the context of Tokenise. Two readings:

- (a) **Normalise first, then Tokenise with defaults.** Apply `Normalise(name, s.normaliseOpts)` to each Item.Name; then `Tokenise(normalised, DefaultTokeniseOptions())`. Bucket keys are tokens of the normalised form.
- (b) **Tokenise directly with default opts.** `Tokenise(name, DefaultTokeniseOptions())`. Bucket keys are tokens of the raw form. The "Scorer's normalisation options" reference is loose wording for "the standard pipeline."

**Recommendation:** **(a)** — Normalise first, Tokenise the normalised form with `DefaultTokeniseOptions()`. This matches the spec's intent (bucket keys mirror what the Scorer actually sees) and what consumers would expect. Document explicitly in Plan 09-01 godoc and in `docs/scan.md`. If Q1 resolves to Option A (accessor), this also resolves: scan calls `s.NormalisationOptions()` → `Normalise(name, normOpts)` → `Tokenise(normalised, DefaultTokeniseOptions())`.

**Risk if wrong:** MEDIUM. Wrong choice gives a semantically valid but consumer-surprising bucket (`Tokenise("user_id")` vs `Tokenise(Normalise("user_id"))` — these produce different token sets when `Normalise` strips separators). The PropCheck_BucketEquivalentToNaive property test catches correctness regardless (the naive comparison also uses normalised input via the Scorer), but the empirical bucket-threshold validation might land on a different crossover.

### 3. Empirical bucket-threshold sweep — single-platform value vs per-platform values

**What we know:** D-08 commits to `bucketThreshold = 50` as a starting hypothesis subject to empirical falsification on darwin/arm64 (the user's machine; the benchstat self-hosted runner is unavailable per STATE.md line 125). The cross-platform CI matrix runs on linux/{amd64,arm64} + darwin/{amd64,arm64} + windows/amd64.

**What's unclear:** Whether to commit a single constant (and accept it may be sub-optimal on other platforms) or expose a build-tagged per-platform value (`// +build darwin,arm64` files).

**Recommendation:** Single constant. Per-platform tuning is an over-engineering trap; D-08's user rationale (YAGNI for tuning knobs) applies to per-platform tuning too. The < 2s/10k budget is the gate, not the absolute crossover. If a single constant produces a > 2s violation on a CI-matrix platform, raise via a follow-up.

**Risk if wrong:** LOW. The < 2s budget has substantial slack (back-of-envelope: 10k items × ~50 tokens-per-item × tokenisation cost ~500ns/item ≈ 5ms tokenisation; bucket build ≈ 1-2ms; per-pair Scorer.Score at ≤ 30µs × (10k × ~50 candidates / 50 group-size) ≈ ~300ms — well within budget). Choosing 30 vs 70 won't break PERF-05.

### 4. Plan-file overlap on scan.go across plans 09-03 / 09-04 / 09-05 / 09-06

**What we know:** The Check function body needs to be assembled incrementally: 09-03 lands the within-group + naive cross-group passes; 09-04 extends with bucket dispatch; 09-05 extends with suppression composition; 09-06 extends with sort + completeness assertion. All four edit scan.go.

**What's unclear:** Whether the project convention forbids file-overlap across plans (the SDK wave-mapping bug note suggests this is sometimes enforced).

**Recommendation:** Inspect Phase 8's plan 08-02 / 08-03 file-modification lists; if they overlapped on scorer.go (they likely did — 08-02 lands NewScorer/Score/Match, 08-03 lands ScoreAll/Threshold/Algorithms/DefaultScorer), then partial overlap is acceptable for clearly-disjoint methods. If not, merge plans 09-03..06 into a single (larger) plan, reducing the wave count from 6 to 4.

**Risk if wrong:** LOW. Either decomposition produces the same correct result; this is purely a plan-checker / file-ownership convention question.

## Plan Decomposition

Recommended **8 plans across 6 waves** (serial execution per user-memory):

| Plan | Wave | Depends on | Files Created | Files Modified | Requirements | Notes |
|------|------|------------|---------------|----------------|--------------|-------|
| **09-01 Foundation** | 1 | — | `scan/scan.go` (Item, Config, Warning structs + Check stub), `scan/kind.go` (Kind + String), `scan/errors.go` (3 sentinels), `scan/doc.go`, `scan/example_test.go` (Kind_String example only), `scan/scan_test.go` (type assertions only) | `docs/requirements.md` (§12.1 + §15 + line 2010 + line 2061 rename + SPEC OVERRIDE notes), `scorer.go` (add `NormalisationOptions()` accessor — Open Question 1), `CHANGELOG.md`, `llms.txt`, `llms-full.txt` | SCAN-06 (sentinels) + foundation for SCAN-01 | api-ergonomics-reviewer sign-off recorded in PR; SPEC OVERRIDE for D-01 + D-02 lockstep with type declarations |
| **09-02 Validation** | 2 | 09-01 | `scan/validate.go` (validateCheck + validateConfigFields + validateItems + validateSuppressedPairs), `scan/validate_test.go` | — | SCAN-06 (errors.Join discipline for D-03/D-05/D-06) | Three errors.Join paths; one BDD scenario per validation gate added in Plan 09-07; coverage gates ≥ 95% / ≥ 90% per file |
| **09-03 Within + naive cross-group** | 3 | 09-02 | — | `scan/scan.go` (Check body: normalise + tokenise pass; group-keyed within pass; naive cross-group pass; cross-group effective-threshold arithmetic with clamp) | SCAN-01 | Naive cross-group ONLY — no bucketing yet; no suppression yet. Establishes the inner loop and the dual-threshold pattern. Initial pass without bucketing keeps the diff small and the PropCheck_BucketEquivalentToNaive gate clean (no bucket = no equivalence question yet). |
| **09-04 Token-bucket optimisation** | 4 | 09-03 | `scan/bucket.go` (bucketThreshold const + tokeniseAll + bucket build + candidate-set union), `scan/props_test.go` (`PropCheck_BucketEquivalentToNaive`, `PropCheck_DeterministicAcrossRuns` initial), `scan/scan_bench_test.go` (`BenchmarkScanCheck_BucketVsNaive_GroupSize` sweep 10/25/50/75/100/200) | `scan/scan.go` (naive ↔ bucket dispatch based on group size) | SCAN-02 + D-08 empirical validation | The D-08 empirical-validation pass — if the benchmark falsifies 50, update the constant in this same plan. algorithm-performance-reviewer sign-off recorded in PR. |
| **09-05 Suppression composition** | 5 | 09-04 | `scan/suppress.go` (isSuppressed + canonicalPair + suppressionCtx), `scan/suppress_test.go` | `scan/scan.go` (suppression wired into inner loop pre-emission), `scan/example_test.go` (Check_WithSuppression added) | SCAN-03 + SCAN-04 | Three suppression rules compose via OR. SuppressedPairs canonicalised at Check entry using Normalise. |
| **09-06 Sort + completeness + golden** | 6 | 09-05 | `testdata/golden/_staging/scan.json` (staging file), `testdata/golden/scan-default.json` (merged), `scan/scan_golden_test.go` (golden-file diff test), `scan/props_test.go` extensions (`PropCheck_NoSelfWarnings`, `PropCheck_NoNaN`, `PropCheck_NoInf`), `scan/fuzz_test.go` (`FuzzCheck`) | `scan/scan.go` (sort.SliceStable + in-line completeness assertion + panic with ErrInternalInvariantViolated), `Makefile` (verify-determinism inclusion if needed per Open Question 7) | SCAN-05 + DET-04 | determinism-reviewer sign-off recorded in PR. Cross-platform CI matrix must show byte-identical scan-default.json. |
| **09-07 BDD coverage** | 6 | 09-06 | `tests/bdd/features/scan.feature` (within/cross/threshold-boost/sort/duplicate-(Name,Group)/empty-Name/Scorer-nil/Config-invalid), `tests/bdd/features/suppression.feature` (SilenceLint × {within,cross} + SuppressedPairs × {with-suppression,without} + CompareIdenticalAcrossGroups × {true,false} + composition: SilenceLint + SuppressedPairs together), `tests/bdd/steps/scan_steps.go` (ScanContext + step regex registrations + InitScanSteps) | `tests/bdd/steps/algorithms_steps.go` (add `InitScanSteps(ctx)` call at line ~1593) | TEST-05 (BDD coverage incl. scan + Phase 8.5 R2 deferred suppression scenarios) | bdd-scenario-reviewer's call on scenario count; minimum 1 per D-03..D-07 + Phase 8.5 R2. goleak via existing `goleak.VerifyTestMain`. |
| **09-08 Examples + docs + bench + finalisation** | 6 | 09-07 | `examples/scan-demo/main.go` + `examples/scan-demo/main_test.go` + `examples/scan-demo/go.mod` (runnable demo + stdout meta-test), `docs/scan.md` (full population from scaffold) | `bench.txt` (extended with all scan benchmarks at the chosen bucketThreshold value), `llms.txt` + `llms-full.txt` (final sync), `.planning/REQUIREMENTS.md` (flip SCAN-01..06 + PERF-05 to Complete), `CHANGELOG.md` (Added section finalised) | DX-02 (godoc Examples) + DX-04 (docs/scan.md) + PERF-05 (10k items bench committed) | docs-writer + user-guide-reviewer sign-off recorded. `make verify-llms-sync` green. |

**Plans 09-06 / 09-07 / 09-08 share Wave 6** because they have no file overlap with each other (09-06 owns golden + scan.go's tail; 09-07 owns BDD files; 09-08 owns examples + docs). Per user-memory `feedback_phase_execution_serial.md`, they execute in serial regardless — but the wave grouping reflects logical "we are now in the finalisation cluster" boundary.

**Validate this decomposition:**

- **Missing plans:** None identified. All seven REQ-IDs covered.
- **Merge candidates:** Plans 09-03 / 09-04 could merge if file-overlap on scan.go is unacceptable (Open Question 4) — but two-step landing (naive first, then bucket) is clearer for the api-ergonomics-reviewer and the property test review. Plans 09-07 / 09-08 could merge into "finalisation" if 8 plans feels too many, but BDD step infrastructure (09-07) is structurally different from the docs/example/bench surface (09-08) and warrants its own plan.
- **Ordering issues:** None — strict dependency chain enforces correct serialisation.
- **Dependency mismatches:** None.

**V2-API decisions to explicitly mark out of scope in plan godoc / docs/scan.md:**

- **V2-API-03** (streaming `iter.Seq[Warning]`) — documented as "out of scope for v1.0; tracked as V2-API-03."
- **V2-API-04** (parallel goroutine-fanned scan) — documented as "out of scope for v1.0; tracked as V2-API-04." Also explains why scan is pure synchronous (the < 2s/10k-items budget is met by token-bucket + ASCII fast paths; goroutines would re-open the determinism story).

Both noted in `docs/scan.md` "Out of scope" section (Plan 09-08).

## Sources

### Primary (HIGH confidence)

- `docs/requirements.md` §12 (lines 1266–1473) — Scan sub-package authoritative spec
- `docs/requirements.md` §13 (lines 1477–1518) — Determinism guarantees
- `docs/requirements.md` §14.1 / §14.2 / §14.4 — Performance budgets + FMA-defeating double-cast
- `docs/requirements.md` §15 (lines 1572+) — Testing strategy incl. BDD feature file inventory
- `.planning/REQUIREMENTS.md` — SCAN-01..06, PERF-05, DET-04, TEST-05, DX-02, DX-04
- `.planning/ROADMAP.md` Phase 9 section (lines 204–212)
- `.planning/phases/09-collection-scan-sub-package/09-CONTEXT.md` — eight locked decisions D-01..D-08 + Claude's Discretion + Deferred Ideas
- `.planning/phases/09-collection-scan-sub-package/09-DISCUSSION-LOG.md` — audit trail confirming user corrections (errors.Join discipline reapplied twice; AskUserQuestion friction)
- `.planning/phases/08-composite-scorer/08-CONTEXT.md` §1 — Phase 8 SPEC OVERRIDE for `ScoreAll` returning `map[AlgoID]float64` (D-01 mirrors)
- `.planning/phases/08-composite-scorer/08-CONTEXT.md` §2 — Mandatory `WithThreshold` rationale (informs D-04 DefaultConfig pattern)
- `.planning/phases/08.5-review-remediation-gate/08.5-CONTEXT.md` Q2 — data-vs-parameter framework (D-03/D-04/D-05/D-06 apply this)
- `.planning/phases/08.5-review-remediation-gate/08.5-CONTEXT.md` Q4 — `ErrInternalInvariantViolated` sentinel (used by D-06)
- `.planning/phases/08.5-review-remediation-gate/08.5-CONTEXT.md` Q6b — CamelCase String() convention (D-02)
- `scorer.go` (lines 75–308 NewScorer, 352–389 Score, 447–449 Threshold, 466–472 Algorithms, 503–523 ScoreAll, 549–559 DefaultScorerOptions, 592–594 DefaultScorer) — Scorer surface scan composes against
- `validate.go` + `warn_kind.go` — root-pkg Warning / WarnKind / AlgoIDAny pattern reference for D-02's CamelCase String()
- `errors.go` — sentinel error catalogue + 4-section godoc template
- `algoid.go` — AlgoID enum used in D-01's `map[AlgoID]float64`
- `tokenise.go` (lines 105–112 DefaultTokeniseOptions, 179+ Tokenise) — used by D-08 bucket extraction; Phase 8.5 Q8b ASCII fast path
- `tests/bdd/features/scorer.feature` — pattern for scan.feature / suppression.feature structure
- `tests/bdd/steps/scorer_steps.go` (lines 1–61 ScorerContext + InitScorerSteps pattern)
- `tests/bdd/steps/algorithms_steps.go` (lines 1576–1605 — `InitScorerSteps(ctx)` / `InitValidateSteps(ctx)` / `InitNormalisationSteps(ctx)` pattern for `InitScanSteps(ctx)`)
- `tests/bdd/bdd_test.go` — `goleak.VerifyTestMain` already wired; scan inherits zero-goroutine-leak gate
- `examples/scorer-composition/main.go` + `main_test.go` + `go.mod` — example program pattern (DX-02 + DX-05 — though DX-05 is Phase 2 scope)
- `testdata/golden/scorer-default.json` — golden file format reference
- `Makefile` (lines 26–60) — make targets the new sub-package wires into
- `.planning/config.json` — `workflow.nyquist_validation: true` confirmed; Validation Architecture section required

### Secondary (HIGH-MEDIUM confidence)

- `.claude/skills/determinism-standards/SKILL.md` — No-Map-Iteration Rule + Sort Key Completeness + Float Stability sections
- `.claude/skills/performance-standards/SKILL.md` — Per-Algorithm Budgets + ASCII Fast Path + Benchstat Regression Detection
- `.claude/skills/algorithm-correctness-standards/SKILL.md` — Mathematical invariants discipline (informs property test design)
- `.claude/skills/algorithm-licensing-standards/SKILL.md` — N/A for scan (no algorithm implementation), but algorithm-licensing-reviewer's pattern of "sign-off recorded in PR" applies to api-ergonomics-reviewer's SPEC OVERRIDE sign-off in Plan 09-01
- `.claude/skills/research-guidance/SKILL.md` — Cite primary sources, capture formulas, flag edge cases
- `CLAUDE.md` (root) — Project Skills directory; agent gates; British English; CI-only releases
- User-memory `feedback_phase_execution_serial.md` — Serial plan execution
- User-memory `project_no_github_issues.md` — Skip commit-message-reviewer issue-ref findings
- User-memory `feedback_ci_before_verification_gate.md` — Origin CI green before verification gate
- User-memory `feedback_yagni_for_tuning_knobs.md` — D-08 private-constant rationale
- User-memory `project_sdk_wave_mapping_bug.md` — Read wave from plan frontmatter directly

### Tertiary (MEDIUM-LOW confidence)

- The cross-platform behaviour of `sort.SliceStable` on identical input on Go 1.26.3 across linux/{amd64,arm64} + darwin/{amd64,arm64} + windows/amd64 — **assumed deterministic from Go stdlib semantics** (`sort.SliceStable` is documented stable). No project-specific verification yet beyond the existing scorer-default.json / algorithms.json / normalisation.json golden files passing on the CI matrix; the same will apply to scan-default.json.
- Whether the Makefile `verify-determinism` target enumerates `testdata/golden/*.json` automatically or hard-codes filenames — **needs inspection in Plan 09-06 / 09-08**. If hard-coded, add `scan-default.json` to the list in the same plan.

## Metadata

**Confidence breakdown:**

- **Spec amendments (D-01 / D-02 + line 1359 + line 1389 + line 2010 + line 2061):** HIGH — exact line numbers verified via grep on docs/requirements.md
- **Plan decomposition (8 plans across 6 waves):** HIGH — modelled directly on Phase 8's 4-plan structure scaled to Phase 9's ~2× surface area; serial execution per user-memory
- **`errors.Join` discipline (Pattern 2):** HIGH — verified against Go stdlib `errors.Is` traversal of `Unwrap() []error` (Go 1.20+); pattern already used implicitly via validate.go sort.SliceStable
- **Cross-group threshold arithmetic + clamp (Pitfall 3 / Pitfall 6):** HIGH — verified against spec §12.2 line 1413
- **Suppression precedence + canonical-pair normalisation (Pattern 3 / Pitfall 4):** HIGH — verified against spec §12.3 line 1417–1423
- **Sort key completeness + ErrInternalInvariantViolated panic (Pattern 5):** HIGH — verified against Phase 8.5 Gap 5 (errors.go:213) + spec §12.4 line 1429
- **Bucket-threshold empirical validation methodology (D-08 benchmark sweep):** MEDIUM — group sizes are CONTEXT.md-fixed at 10/25/50/75/100/200, but total item count, input distribution, and platform-specific reporting are research-determined. Recommendation: 10,000 total items per benchmark variant (matches PERF-05); input distribution = ASCII identifier-like strings (e.g. `field_001..field_N` with low edit-distance overlap clusters); reporter = `bench.txt` committed at the Plan 09-04 close.
- **Open Question 1 (Scorer normalisation accessor):** MEDIUM-LOW — needs explicit decision before Plan 09-01 starts. HIGH risk if mishandled.
- **Open Question 2 (Tokenise + spec ambiguity):** MEDIUM — spec wording is loose; recommendation is documented; risk is low.
- **Cross-platform `testdata/golden/scan-default.json` determinism:** MEDIUM — extrapolated from existing scorer-default.json / algorithms.json passing on the matrix; first-pass run on Plan 09-06 will confirm.

**Research date:** 2026-05-19
**Valid until:** 2026-06-18 (30-day stable horizon; spec drift improbable in <30 days because Phase 9 is the next phase to execute)

---

*Phase: 9 — Collection Scan Sub-package*
*Research synthesised: 2026-05-19*
