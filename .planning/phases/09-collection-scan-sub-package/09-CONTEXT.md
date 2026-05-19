# Phase 9: Collection Scan Sub-package - Context

**Gathered:** 2026-05-19
**Status:** Ready for planning

<domain>
## Phase Boundary

Ship Layer 3 of the three-layer fuzzymatch architecture — the **`github.com/axonops/fuzzymatch/scan`** sub-package: a turnkey collection-scan layer over the Phase 8 Scorer that finds pairs of similar names in a collection with grouping semantics, suppression composition, deterministic output, and a token-bucket optimisation property-test-verified equivalent to the naive O(N²) implementation.

**Public surface delivered (per spec §12.1, modulo decisions below):**

- `scan.Item` struct (`Name`, `Group`, `SilenceLint`, `Tag any`)
- `scan.Kind` enum (locked over spec's `WarningKind` — see D-02) + `KindWithinGroup` / `KindAcrossGroups` constants + `String()` returning CamelCase
- `scan.Warning` struct (`Kind`, `NameA/NameB`, `GroupA/GroupB`, `TagA/TagB any`, `Score float64`, `Scores map[AlgoID]float64` — see D-01)
- `scan.Config` struct (`Scorer`, `CompareAcrossGroups`, `CrossGroupThresholdBoost`, `CompareIdenticalAcrossGroups`, `SuppressedPairs`)
- `scan.DefaultConfig(s *fuzzymatch.Scorer) Config` helper baking the opinionated `CrossGroupThresholdBoost = 0.05` + `CompareIdenticalAcrossGroups = false` defaults (see D-04)
- `scan.Check(items []Item, cfg Config) ([]Warning, error)`
- Sentinel errors: `scan.ErrNilScorer`, `scan.ErrInvalidItem`, `scan.ErrInvalidConfig`
- Determinism artefacts: `testdata/golden/scan-default.json` (cross-platform byte-identical), `PropCheck_DeterministicAcrossRuns`, `PropCheck_BucketEquivalentToNaive`
- BDD: `tests/bdd/features/scan.feature` + `tests/bdd/features/suppression.feature` (the Phase 8.5 R2 deferral lands here)
- Performance: < 2s for 10,000 items / 500 groups (PERF-05); benchstat regression detection
- Docs: `docs/scan.md` populated (currently scaffold)
- Example: `examples/scan-demo/main.go` runnable, meta-tested

**Closes:** SCAN-01..06 + PERF-05.

**Out of scope:**
- `Extract` / `ExtractOne` one-to-many API — Phase 10
- Streaming `iter.Seq[Warning]` variant — v2 (V2-API-03)
- Parallel goroutine-fanned scan — v2 (V2-API-04)
- Any new algorithms — catalogue is complete at 23 after Phase 7

</domain>

<decisions>
## Implementation Decisions

### §1. Warning type alignment with Phase 8 — LOCKED (SPEC OVERRIDE × 2)

- **D-01:** `scan.Warning.Scores` is **`map[AlgoID]float64`**, overriding spec §12.1 line 1337's `map[string]float64`. **Rationale:** `Warning.Scores` is populated directly from `Scorer.ScoreAll`, which Phase 8 LOCKED to `map[AlgoID]float64` (08-CONTEXT.md §1, SPEC OVERRIDE applied to `docs/requirements.md` §8.3 + §8.6 in lockstep with Plan 08-03). The same override applies here for the same reason: typed enum keys give compile-time safety; `map[string]float64` would force consumers to re-encode the snake_case namespace. Spec amendment lands in the planner's first commit per the Phase 8 precedent.

- **D-02:** The within/cross discriminator type is **`scan.Kind`** (NOT `scan.WarningKind` as in spec §12.1, NOT `scan.WarnKind` for accidental symmetry with root's `fuzzymatch.WarnKind`). Constants unchanged: `KindWithinGroup`, `KindAcrossGroups`. `String()` returns CamelCase (`"WithinGroup"`, `"AcrossGroups"`) per the §6 + Phase 8.5 Q6b convention. **Rationale:** package-scoped `Kind` is unambiguous at the call site (`scan.KindWithinGroup`); the `WarnKind` / `WarningKind` pair would mislead consumers into expecting a shared inheritance. Spec amendment renames `WarningKind` → `Kind` throughout §12.1, §12.4, §15 references.

**Action for the planner:**
- The first plan in Phase 9 (most likely the foundation plan landing `scan/scan.go` + types) amends `docs/requirements.md` §12.1 to: `Kind` type rename, `Scores map[AlgoID]float64`, line 1389 phrasing update, line 1359 default-0.05 statement migrated to `DefaultConfig` godoc.
- Mention both overrides explicitly in `scan.Warning.Scores` and `scan.Kind` godoc with a back-reference to this CONTEXT.md §1.
- Cite api-ergonomics-reviewer sign-off on the foundation plan's PR (same pattern as Phase 7 Plan 07-02 / Phase 8 Plan 08-03).

### §2. Config validation strictness — LOCKED (Phase 8.5 Q2 framework applied to scan)

Phase 8.5 Q2 locks: **parameters strict, comparison data lenient.** Config fields and Item.Name are parameters in scan's surface (Item.Name has no defined similarity semantics when empty, so it's a parameter, not data). Three coupled decisions:

- **D-03:** **Empty `Item.Name` → `ErrInvalidItem`, all offending indices collected via `errors.Join`.** Pre-flight pass over `items[]` walks the full slice, accumulates every empty-Name index, and returns a single wrapped error. Caller fixes the whole batch in one round-trip. **Spec §12.1 line 1389** ("the offending item's index") is amended to: "every offending index, joined via `errors.Join`."
  ```go
  // items[7].Name == "" and items[42].Name == ""
  _, err := scan.Check(items, cfg)
  // err: errors.Join of two ErrInvalidItem wraps; errors.Is(err, scan.ErrInvalidItem) == true
  ```

- **D-04:** **`Config.CrossGroupThresholdBoost` zero-value = `0.0` (no boost). The opinionated default `0.05` lives in `scan.DefaultConfig(s *fuzzymatch.Scorer) Config`**, mirroring Phase 8's `Scorer.DefaultScorer` / `DefaultScorerOptions` pattern. Strict validation per Q2: NaN, ±Inf, < 0, > 1 → `ErrInvalidConfig` at `Check` call time (NOT panic — Config is a parameter the consumer might construct programmatically). Spec §12.1 line 1359 default-0.05 statement migrates to `DefaultConfig` godoc.
  ```go
  cfg := scan.DefaultConfig(s)                       // boost = 0.05, identical-cross = false
  cfg.CompareAcrossGroups = true                     // still opt-in
  ```

- **D-05:** **`Config.SuppressedPairs` validation: empty strings → `ErrInvalidConfig`, all offending indices collected via `errors.Join`.** Self-pairs (`a == b` post-normalisation) are **silently kept** — they're harmless because `Check` never emits a self-warning. Same `errors.Join` pattern as D-03 / D-06 for consistency across all three input-validation paths.

**Action for the planner:**
- The validation order is: `cfg.Scorer == nil` (cheapest, fail-first) → `Config` field validation (boost range, NaN/Inf) → `Items[]` validation (empty-Name AND duplicate-(Name, Group) per D-06) → `SuppressedPairs` validation. Within each phase, accumulate **all** errors and return a single `errors.Join`. Across phases, fail-fast — once `Scorer` is nil there's nothing else to validate.
- All three `ErrInvalid*` wraps carry the originating index in the error message (e.g. `"scan: invalid item at index 7: empty name"`) so the caller can pinpoint the failure without re-walking the slice.

### §3. Sort-key completeness — LOCKED (closes SCAN-05 in-line assertion)

- **D-06:** **Duplicate `(Name, Group)` pairs in `Items[]` → `ErrInvalidItem` at validation time, all duplicate indices collected via `errors.Join`.** With duplicates rejected at the door, the sort key `(Kind, NameA, NameB, GroupA, GroupB)` is **provably unique by construction**. The "in-line completeness assertion" required by SCAN-05 becomes a hard invariant (`if dup found in sorted output → panic with ErrInternalInvariantViolated`), not a tiebreaker. Consistent `errors.Join` pattern with D-03 / D-05.

  ```go
  items := []scan.Item{
      {Name: "user_id", Group: "login"},
      {Name: "user_id", Group: "login"},  // duplicate
      {Name: "userId",  Group: "login"},
  }
  _, err := scan.Check(items, cfg)
  // err: scan: invalid item at index 1: duplicate (Name, Group) of index 0
  // errors.Is(err, scan.ErrInvalidItem) == true
  ```

**Action for the planner:**
- The duplicate-detection pass uses a `map[struct{Name, Group string}]int` from canonical-pair → first-seen index. Map iteration is **never exposed on the output path** — the map is consumed only to populate the error list, which is then sorted by ascending duplicate-index before being joined.
- The in-line completeness assertion at the end of `Check` (after sorting) verifies no two adjacent warnings share `(Kind, NameA, NameB, GroupA, GroupB)`. A violation is a library bug, not user error — panic with the Phase 8.5 Q4 `fuzzymatch.ErrInternalInvariantViolated` sentinel (Gap 5 from Phase 8.5).
- `scan.feature` BDD covers the duplicate-detection case (both single and multi-duplicate, exercising the `errors.Join` accumulation).

### §4. Validate / scan integration — LOCKED

- **D-07:** **Separate concerns.** `scan.Check` does similarity computation only. Consumers wanting input-quality diagnostics call `fuzzymatch.Validate(a, b)` explicitly, either upstream of `scan.Check` (per-item pre-flight) or downstream (per-emitted-warning post-hoc). `scan.Check` does NOT internally invoke `Validate`; `scan.Warning` does NOT carry a `Diagnostics []fuzzymatch.Warning` field; `Check` returns `([]scan.Warning, error)` (NOT `([]scan.Warning, []fuzzymatch.Warning, error)`).

**Rationale:** matches Phase 8.5 Q4 intent — `Validate` is a deliberate per-pair diagnostic surface for consumers who want pre-flight input checks; folding it into `Check` would couple two diagnostic systems with different cardinalities (Validate: O(N) per item; Check: O(N²) before bucketing). Clean two-API boundary is the principled split.

**Action for the planner:**
- `docs/scan.md` quickstart shows the recommended pattern: call `fuzzymatch.Validate` over each item's Name in a precondition step, then call `scan.Check` for similarity.
- `examples/scan-demo/main.go` may demonstrate this composition explicitly (one snippet using Validate-then-Check) to make the boundary visible.
- No new entry on the `fuzzymatch.Warning` / `scan.Warning` cross-type story — they remain distinct types in distinct packages.

### §5. Bucket threshold — LOCKED as private constant

- **D-08:** **`bucketThreshold = 50` is a private package constant in `scan/bucket.go`, NOT a `Config.BucketThreshold` field.** Naive nested loops are used when a group's item count is `≤ bucketThreshold`; the token-bucket optimisation kicks in above that.

**Rationale (user-locked principle; saved to user-memory `feedback-yagni-for-tuning-knobs`):**
1. **YAGNI for optimisation knobs.** Consumers don't reason about "should my bucket threshold be 30 or 70?" — they reason about "is this pair flagged or not?"
2. **API expansion is cheap, contraction is impossible.** Promoting to `Config.BucketThreshold` in v1.x is a non-breaking minor-version addition; going the other direction is breaking.
3. **Phase 8.5 was tightening, not expanding.** Q3 / Q4 / Q5 all removed surface; adding a new knob no consumer has demanded violates that direction.
4. **Benchmarking doesn't need a public field.** Maintainers verify the threshold via internal tests where the private constant is accessible.

**Action for the planner — empirical validation BEFORE v1.0 ships:**
- Add `BenchmarkScanCheck_BucketVsNaive_GroupSize` to `scan/scan_bench_test.go` sweeping group sizes `10 / 25 / 50 / 75 / 100 / 200` with the same total item count.
- If the empirical crossover (where bucket overtakes naive on the benchmarked hardware) differs materially from `50`, **update the constant** before Phase 9 closes. The spec's stated `50` is a starting hypothesis; the benchmark is load-bearing.
- Commit the chosen constant value with a one-line `// bucketThreshold is the group-size cutoff at which the token-bucket optimisation overtakes the naive nested-loop comparison. Empirically validated on darwin/arm64 by BenchmarkScanCheck_BucketVsNaive_GroupSize at $TIME / Plan 09-NN.` comment, including the wall-clock crossover from the benchmark.

### Claude's Discretion

- **Where Item validation surfaces report errors** (single-error format vs structured ErrInvalidItem with index field) — Claude can choose between `fmt.Errorf("scan: invalid item at index %d: %w", idx, ErrInvalidItem)` (Go idiomatic) vs a dedicated `*InvalidItemError` struct exposing `.Index int` field, as long as the error implements `errors.Is(err, scan.ErrInvalidItem)` cleanly and the message embeds the index.
- **Token-bucket internal data layout** (`map[string][]int` per spec §12.5 vs a different structure that achieves the same algorithmic complexity) — Claude / the planner decides. The property test `PropCheck_BucketEquivalentToNaive` is the load-bearing correctness gate.
- **Whether to expose a precomputed `tokeniseAll` helper** in `scan/bucket.go` vs inline it inside `Check` — Claude's call. Either way, every Item.Name is tokenised at most once per `Check` invocation.
- **Example program shape for `examples/scan-demo/main.go`** — must demonstrate (a) within-group only with the default Scorer, (b) within + cross-group with `DefaultConfig`, (c) at least one of each suppression mode (per-item `SilenceLint`, `SuppressedPairs`, identical-cross suppression). Beyond that, Claude / docs-writer can shape the narrative.
- **`scan.feature` and `suppression.feature` scenario count** — bdd-scenario-reviewer's call, but minimum coverage: every D-03..D-07 decision has at least one scenario, plus the Phase 8.5 R2 deferred suppression scenarios.

</decisions>

<canonical_refs>
## Canonical References

**Downstream agents MUST read these before planning or implementing.**

### Authoritative specs
- `docs/requirements.md` §12 (lines 1266–1473) — Scan sub-package: public API, within/cross passes, suppression composition, determinism guarantees, token-bucket optimisation, performance budget, repository layout.
- `docs/requirements.md` §13 — Determinism guarantees (algorithm score stability, Scorer composite stability, scan output sort, no-map-iteration-on-output rule, golden file coverage).
- `docs/requirements.md` §15 — BDD coverage discipline (scan.feature + suppression.feature land here per §15.6).
- `docs/requirements.md` §6 + §11.5 — Algorithm identifiers naming convention; `Validate` / `Warning` / `WarnKind` surface in root package (informs the `scan.Kind` vs `WarningKind` naming decision).
- `.planning/REQUIREMENTS.md` — SCAN-01..06, PERF-05, DET-04 (NaN/Inf/-0 handling for scan), TEST-05 (BDD scope including scan), DX-02 / DX-04 (godoc Example + docs/scan.md).
- `.planning/ROADMAP.md` Phase 9 section — phase goal, dependencies, success criteria.

### Prior-phase decisions that flow forward (MUST read)
- `.planning/phases/08-composite-scorer/08-CONTEXT.md` §1 — Phase 8 SPEC OVERRIDE for `ScoreAll` returning `map[AlgoID]float64`. **D-01 in this CONTEXT.md extends the same override to `scan.Warning.Scores`.** Required reading; the rationale for the override applies verbatim here.
- `.planning/phases/08-composite-scorer/08-CONTEXT.md` §2 — Mandatory `WithThreshold` rationale (informs scan's thinking on opinionated-vs-baked defaults — `DefaultConfig` follows the same "explicit construction is the default, helper bakes opinions" pattern).
- `.planning/phases/08.5-review-remediation-gate/08.5-CONTEXT.md` Q2 — data-vs-parameter framework. **D-04 and D-05 in this CONTEXT.md apply Q2 to scan's `Config` and `SuppressedPairs`.** Required reading.
- `.planning/phases/08.5-review-remediation-gate/08.5-CONTEXT.md` Q4 — `Validate` public surface, `ErrInternalInvariantViolated` sentinel (used by D-06's in-line completeness assertion). Required reading.
- `.planning/phases/08.5-review-remediation-gate/08.5-CONTEXT.md` Q6b — CamelCase String() convention (applied to `scan.Kind.String()` per D-02).

### Standards skills
- `.claude/skills/algorithm-correctness-standards/SKILL.md` — primary-source citation, mathematical invariants. Scan is not a single-algorithm phase but the in-line completeness assertion + property tests follow the same discipline.
- `.claude/skills/determinism-standards/SKILL.md` — no-map-iteration-on-output, float stability, golden files. Load-bearing for D-06's invariant and the `scan-default.json` golden.
- `.claude/skills/go-coding-standards/SKILL.md` — error conventions (`errors.Is` / `errors.Join` patterns), no-cgo, dependency rules.
- `.claude/skills/go-testing-standards/SKILL.md` — coverage targets (≥95% overall / ≥90% per file / 100% public API), property test discipline, BDD coverage rule.
- `.claude/skills/performance-standards/SKILL.md` — allocation budgets, benchstat regression detection (load-bearing for D-08's empirical bucket-threshold validation).
- `.claude/skills/documentation-standards/SKILL.md` — godoc Example per public function, docs/scan.md scaffolding.

### Existing code the planner should map BEFORE writing plans
- `scorer.go` — Scorer surface that scan consumes. `Scorer.ScoreAll` signature determines D-01 directly. `Scorer.Threshold()` informs cross-group boost arithmetic.
- `validate.go` + `warn_kind.go` — root-package Warning / WarnKind types. Confirms the rename motivation in D-02; reference for the `scan.Kind` shape.
- `errors.go` — sentinel error catalogue (post-Phase-8.5 inventory). New scan-scoped sentinels in `scan/errors.go` follow the same four-section godoc template.
- `algoid.go` — `AlgoID` enum used in D-01's `map[AlgoID]float64`. `AlgoID.String()` returns CamelCase per Phase 8.5 Q6b.
- `tokenise.go` — used by D-08's token-bucket extraction. The Phase 8.5 Q8b ASCII fast path is load-bearing for the < 2s budget.
- `tests/bdd/features/scorer.feature` — pattern for scan.feature / suppression.feature structure.
- `examples/scorer-composition/` — pattern for examples/scan-demo/ shape and meta-test wiring.

</canonical_refs>

<code_context>
## Existing Code Insights

### Reusable Assets
- **`Scorer.Score`, `Scorer.ScoreAll`, `Scorer.Match`, `Scorer.Threshold`** — scan composes against these directly. `Scorer.Match` is the primary call inside the inner loop (returns the threshold-applied bool); `Scorer.ScoreAll` populates `Warning.Scores` only when an emission is happening (so the per-algorithm breakdown is paid for only on the rare positive case).
- **`Tokenise(s, DefaultTokeniseOptions())`** — used in D-08's bucket extraction to derive per-Item token sets. Phase 8.5 Q8b ASCII fast path returns substrings via `s[lo:hi]` zero-copy when `opts.Lowercase == false` — load-bearing for the < 2s/10k-items budget.
- **`Normalise(s, opts)`** — applied to NameA / NameB before comparison (the Scorer already normalises internally per Phase 8 §3, so scan must NOT double-normalise); the bucket extractor uses the Scorer's normalisation options to ensure SuppressedPairs and Items normalise consistently.
- **`fuzzymatch.ErrInternalInvariantViolated`** (Phase 8.5 Gap 5) — typed panic value for library-internal bugs; used by D-06's in-line completeness assertion panic path.
- **`AlgoID.String()`** — CamelCase rendering for Scores map keys when consumers need human-readable display.

### Established Patterns
- **Functional-options-with-helper pattern (Phase 8 `DefaultScorer` / `DefaultScorerOptions`)** — D-04's `DefaultConfig(s)` follows the same shape: bake opinionated defaults into a helper; zero-value of the struct remains a valid (minimal) configuration.
- **errors.Join for collected validation errors** — new pattern locked in Phase 9 (D-03, D-05, D-06). Three input-validation paths all use it for consistency. Replaces the spec's implicit "fail-fast" wording.
- **In-line completeness assertion with `ErrInternalInvariantViolated` panic** — Phase 8.5 Gap 5 established the typed-panic vocabulary; D-06 uses it for the SCAN-05 sort-key invariant.
- **SPEC OVERRIDE pattern (Phase 8 §1)** — the planner's first commit amends `docs/requirements.md` in lockstep, with a `SPEC OVERRIDE (Phase 9)` note in the affected section and a back-reference to this CONTEXT.md. D-01 and D-02 follow the same pattern.
- **api-ergonomics-reviewer sign-off recording** — Phase 7 Plan 07-02 (Double Metaphone) and Phase 8 Plan 08-03 (ScoreAll override) both record the reviewer sign-off in the PR description. Phase 9's foundation plan does the same for D-01 + D-02.
- **No-map-iteration-on-output rule** — token-bucket internals use `map[string][]int` but never iterate it on output paths; the bucket is consumed in input-order to build candidate sets, which are then sorted before scoring.

### Integration Points
- **Sub-package layout** — `scan/` lives under the root module (`github.com/axonops/fuzzymatch/scan`), no separate `go.mod`. The root package has zero dependency on `scan` (Layer 3 → Layer 2 is one-directional).
- **Golden file** — `testdata/golden/scan-default.json` joins the existing five-file golden inventory (`algorithms.json`, `scorer-default.json`, `normalisation.json`, plus the future `extract-default.json` in Phase 10). Cross-platform CI matrix exercises it the same way as the others.
- **BDD module** — `tests/bdd/go.mod` already imports godog + goleak + testify (Phase 8). Phase 9 adds `tests/bdd/steps/scan_steps.go` for scan + suppression scenario step definitions; goleak verifies zero goroutine leaks (scan is pure-function, no concurrency).
- **Makefile** — `make check` covers scan via `go test ./...`; `make test-bdd` covers the new feature files; `make bench` includes the empirical bucket-threshold benchmark from D-08; `make verify-determinism` covers the new golden file.
- **llms.txt + llms-full.txt** — every new public symbol synced per the Phase 5+ per-plan llms-sync discipline. `make verify-llms-sync` (Phase 8.5 Q13) catches drift.
- **CHANGELOG.md** — Phase 9 entries under `[Unreleased]`: Added (scan sub-package, scan.Kind, scan.Warning, scan.Config, scan.Check, scan.DefaultConfig, three scan sentinels), plus a one-line spec-amendment entry for §12.1 (D-01 / D-02 lockstep amendments).

</code_context>

<specifics>
## Specific Ideas

- **Empirical bucket-threshold validation is load-bearing.** The user explicitly noted that the spec's `50` is a starting hypothesis, not a settled value. The `BenchmarkScanCheck_BucketVsNaive_GroupSize` sweep (10 / 25 / 50 / 75 / 100 / 200) is required before v1.0 ships, and if the crossover differs materially, the constant updates. Cite the wall-clock crossover in the in-source comment alongside the constant.
- **errors.Join used uniformly across all three input-validation paths.** D-03 (empty Item.Name), D-05 (empty SuppressedPairs entry), D-06 (duplicate `(Name, Group)` Items) all collect every offending index and return a single joined error. The user corrected the initial fail-fast framing twice during discussion — uniform collect-all is the locked pattern.
- **Spec drift the planner's first commit must fix:**
  - §12.1 line 1306–1322: `WarningKind` → `Kind` rename
  - §12.1 line 1337: `Scores map[string]float64` → `Scores map[AlgoID]float64`
  - §12.1 line 1359: default-0.05 statement migrates to `DefaultConfig` godoc
  - §12.1 line 1389: "the offending item's index" → "every offending index, joined via errors.Join"
  - §12.4 / §12.6 / §15: every `WarningKind` reference updated to `Kind`
- **api-ergonomics-reviewer sign-off** must be recorded in the foundation plan's PR description for D-01 + D-02 (the SPEC OVERRIDEs). Match the Phase 8 Plan 08-03 / Phase 7 Plan 07-02 precedent.

</specifics>

<deferred>
## Deferred Ideas

- **`Config.BucketThreshold` public field** — D-08 explicitly locked the threshold as a private constant for v1.0. If post-v1.0 consumers surface a real tuning need (adversarial workloads, distributions where the empirical crossover varies wildly), promote to a Config field in a non-breaking minor-version addition. Track via a future GitHub issue if/when the need materialises.
- **Side-channel `Validate` integration** — D-07 ruled this out for v1.0. If consumers report repeated boilerplate of "call Validate on every Item.Name, then call Check," consider a `scan.CheckWithDiagnostics` variant in v1.x that wraps the two. Until that signal exists, the two-API boundary stands.
- **Streaming output (`iter.Seq[Warning]` Check variant)** — already noted as V2-API-03 in `.planning/REQUIREMENTS.md`. Not in scope for Phase 9.
- **Parallel goroutine-fanned scan** — already noted as V2-API-04. Not in scope for Phase 9 (the < 2s/10k-items budget is met by token-bucket optimisation + ASCII fast paths in the underlying algorithms; goroutines would re-open the determinism story).
- **Dedicated `scan.Item` constructor / builder** — not raised during discussion. The plain struct literal pattern is fine for v1.0. Revisit if consumers report awkward construction patterns.

</deferred>

---

*Phase: 9-Collection-Scan-Sub-package*
*Context gathered: 2026-05-19*
