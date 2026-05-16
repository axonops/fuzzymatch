---
phase: 8
slug: composite-scorer
status: draft
nyquist_compliant: false
wave_0_complete: false
created: 2026-05-16
---

# Phase 8 — Validation Strategy

> Per-phase validation contract for feedback sampling during execution.

---

## Test Infrastructure

| Property | Value |
|----------|-------|
| **Framework** | Go stdlib `testing` (root) + `testing/quick` (property) + native `go test -fuzz`; godog v0.15.0 + goleak v1.3.0 + testify v1.10.0 (BDD sub-module at `tests/bdd/go.mod`) |
| **Config file** | None at root (zero-config stdlib); `tests/bdd/go.mod` for BDD sub-module |
| **Quick run command** | `go test -race -run TestScorer ./...` |
| **Full suite command** | `go test -race ./... && cd tests/bdd && go test -race ./...` |
| **Estimated runtime** | ~25 seconds (root) + ~15 seconds (BDD); ~40 seconds total |

---

## Sampling Rate

- **After every task commit:** Run `go test -race -run TestScorer ./...` (root) — catches the option/Scorer/Score unit regressions in < 5 seconds.
- **After every plan wave:** Run `go test -race ./... && cd tests/bdd && go test -race ./...` — full root + BDD suite.
- **Before `/gsd-verify-work`:** `make check` (golangci-lint v2 + go vet + golden file gate + coverage ≥ 95% overall, ≥ 90% per file, 100% public API surface) must be green.
- **Max feedback latency:** 5 seconds for the per-task command on a warm cache.

---

## Per-Task Verification Map

> Tasks are listed by the four-plan decomposition locked in CONTEXT.md §4 (08-01 → 08-02 → 08-03 → 08-04, strict linear sequence — no parallel waves). Plan-level Task IDs follow the standard `{phase}-{plan}-{task}` pattern; the gsd-planner will subdivide each plan into individual tasks during planning. The table below is the requirement-to-test contract — the planner attaches `<acceptance_criteria>` blocks that reference these commands.

| Task ID | Plan | Wave | Requirement | Threat Ref | Secure Behavior | Test Type | Automated Command | File Exists | Status |
|---------|------|------|-------------|------------|-----------------|-----------|-------------------|-------------|--------|
| 08-01-* | 01 | 1 | SCORER-01, SCORER-03 | T-08-01 (NaN/Inf weight rejection) | `WithAlgorithm(_, ≤ 0)` returns `ErrInvalidWeight`; `WithAlgorithm(invalidAlgoID, _)` returns `ErrInvalidAlgorithm` | unit | `go test -run TestWith ./...` | ❌ W0 (Wave 08-01) | ⬜ pending |
| 08-02-* | 02 | 2 | SCORER-01, SCORER-04, SCORER-06, SCORER-08 | T-08-02 (composite reduction determinism) | `NewScorer` returns `ErrMissingThreshold` FIRST; `Score` uses `(w*s)+acc` parenthesisation in AlgoID-sorted order | unit + property | `go test -race -run TestNewScorer\|TestScorer_Score\|TestMatch ./...` | ❌ W0 (Wave 08-02) | ⬜ pending |
| 08-02-LWW | 02 | 2 | SCORER-03 | — | Duplicate `WithAlgorithm(AlgoX, _)` calls — later weight wins; `WithoutAlgorithm(AlgoX)` no-ops on absent ID | internal unit | `go test -run TestScorer_LastWriteWins ./...` | ❌ W0 (`scorer_internal_test.go`) | ⬜ pending |
| 08-03-* | 03 | 3 | SCORER-02, SCORER-05, SCORER-07 | — | `DefaultScorer()` cannot fail; `ScoreAll` returns `map[AlgoID]float64` (SPEC OVERRIDE — typed keys); `Algorithms()` returns fresh sorted slice | unit + property | `go test -race -run TestDefaultScorer\|TestScoreAll\|TestAlgorithms ./...` | ❌ W0 (Wave 08-03) | ⬜ pending |
| 08-03-PROP | 03 | 3 | SCORER-03, SCORER-04 | — | `PropScorer_WeightSumOne` (when normalised); `PropScorer_ScoreInRange [0,1]` (when normalised); `PropScorer_DeterministicAcrossRuns` | property (`testing/quick`) | `go test -run TestProp_Scorer ./...` | ❌ W0 (Wave 08-03) | ⬜ pending |
| 08-03-CONC | 03 | 3 | SCORER-01 | T-08-03 (data race on shared Scorer) | N goroutines call `Score`/`ScoreAll`/`Match` on same `*Scorer` — `go test -race` clean; identical results across goroutines | concurrent (`sync.WaitGroup` — stdlib only) | `go test -race -run TestScorer_ConcurrentSafety ./...` | ❌ W0 (Wave 08-03) | ⬜ pending |
| 08-04-GOLD | 04 | 4 | SCORER-04 | T-08-04 (cross-platform float drift) | `scorer-default.json` (22-26 entries; mandatory rows per CONTEXT.md §6) diffs byte-identically across linux/amd64, linux/arm64, darwin/amd64, darwin/arm64, windows/amd64 | golden file | `go test -run TestGolden_ScorerDefault ./...` | ❌ W0 (`scorer_golden_test.go`) | ⬜ pending |
| 08-04-BDD | 04 | 4 | SCORER-01..SCORER-08 | T-08-05 (goroutine leak in BDD harness) | 8-12 scenarios in `tests/bdd/features/scorer.feature` covering the 12 mandatory scenario classes (CONTEXT.md §7); `goleak.VerifyTestMain` confirms zero goroutine leaks | BDD (godog) + goleak | `cd tests/bdd && go test -race ./...` | ❌ W0 (`scorer.feature` + `scorer_steps.go`) | ⬜ pending |
| 08-04-BENCH | 04 | 4 | SCORER-04 | — | `BenchmarkDefaultScorer_Score` on ASCII ≤ 50 chars: < 30 µs wall time, ≤ 8 allocations per call | benchmark | `go test -bench=BenchmarkDefaultScorer_Score -benchmem ./...` | ❌ W0 (`scorer_bench_test.go`) | ⬜ pending |
| 08-04-DOCS | 04 | 4 | SCORER-08 | — | `docs/scorer.md` quickstart compiles via meta-test (the existing `ai_friendly_test.go` pattern); `docs/tuning.md` "How to pick a threshold" + "How to pick weights" sections present; llms.txt + llms-full.txt sync verified by existing meta-test | meta-test | `go test -run TestLLMSync ./... && go test -run TestDocsCompile ./...` | ❌ W0 (extends existing meta-tests) | ⬜ pending |
| 08-04-EXAMPLES | 04 | 4 | SCORER-02, SCORER-08 | — | `examples/identifier-similarity/main.go` gains a final Score+Match column; new `examples/scorer-composition/main.go` (+ companion `main_test.go` with golden stdout) demonstrates `DefaultScorerOptions() + WithoutAlgorithm + WithThreshold` composition | example + golden stdout | `go test ./examples/... -update` then `go test ./examples/...` | ❌ W0 (Wave 08-04) | ⬜ pending |

*Status: ⬜ pending · ✅ green · ❌ red · ⚠️ flaky*

**Sampling continuity audit:** the four plans are sequential; each plan has at least one automated test command above. No 3 consecutive tasks lack an automated `<verify>` hook.

---

## Wave 0 Requirements

The following test files do NOT yet exist on disk and MUST be created during each plan's execution. The gsd-planner attaches Wave 0 dependencies on these files where appropriate (a unit test cannot run before its `_test.go` file exists).

- [ ] `scorer_options_test.go` — option-function happy + error paths (Wave 08-01)
- [ ] `scorer_test.go` — `NewScorer`/`Score`/`Match`/`ScoreAll`/`DefaultScorer`/`Algorithms` unit + property tests (Waves 08-02, 08-03)
- [ ] `scorer_internal_test.go` — `scorerConfig` last-write-wins, weight normalisation invariants (Wave 08-02; package-internal `package fuzzymatch` to access unexported state)
- [ ] `scorer_golden_test.go` — `TestGolden_ScorerDefault` using existing `assertGolden` harness from `golden_test.go:66` (Wave 08-04)
- [ ] `scorer_bench_test.go` — `BenchmarkDefaultScorer_Score` measuring < 30 µs / ≤ 8 allocs (Wave 08-04)
- [ ] `testdata/golden/scorer-default.json` — generated via `go test -run TestGolden_ScorerDefault -update ./...` (Wave 08-04; not hand-edited)
- [ ] `tests/bdd/features/scorer.feature` — 8-12 Gherkin scenarios (Wave 08-04)
- [ ] `tests/bdd/steps/scorer_steps.go` — step definitions for `scorer.feature` (Wave 08-04)
- [ ] `examples/scorer-composition/main.go` + `examples/scorer-composition/main_test.go` (Wave 08-04)

The `goleak.VerifyTestMain` hook is **already wired** at `tests/bdd/bdd_test.go:37-39` (Phase 6 inheritance) — no new TestMain is required. New scorer steps register via the existing `InitializeScenario` extension point.

*Existing infrastructure that Phase 8 reuses without modification:* `dispatch[]` (algoid.go), `Normalise`/`NormalisationOptions` (normalise.go), `errors.go` sentinel pattern, `golden_test.go` `-update` flag, `tests/bdd/bdd_test.go` goleak hook.

---

## Manual-Only Verifications

| Behavior | Requirement | Why Manual | Test Instructions |
|----------|-------------|------------|-------------------|
| api-ergonomics-reviewer sign-off on `ScoreAll` SPEC OVERRIDE (`map[AlgoID]float64` instead of `map[string]float64`) | SCORER-05 | The SPEC OVERRIDE must be explicitly acknowledged by the api-ergonomics-reviewer agent in plan 08-03's PR description (same pattern as Phase 7 plan 07-02's algorithm-licensing-reviewer sign-off) — this is a process gate, not a runtime test | Run `Agent(api-ergonomics-reviewer)` against plan 08-03 PR; verify the sign-off appears in the PR description; confirm `Scorer.ScoreAll` godoc carries the SPEC-OVERRIDE notice |
| Cross-platform `scorer-default.json` byte-identical verification | SCORER-04 | The cross-platform matrix runs in CI on linux/amd64, linux/arm64, darwin/amd64, darwin/arm64, windows/amd64 — no local single-machine test can confirm cross-platform stability | After plan 08-04 merges, observe the CI matrix run on the merge commit; confirm no platform shows a diff in `scorer-default.json` |
| `algorithm-correctness-reviewer` review of composite reduction loop | SCORER-04 | Pattern-conformance check against Phase 5 Cosine's `(x*y)+z` parenthesisation — process gate, not runtime test | Spawn `Agent(algorithm-correctness-reviewer)` on plan 08-02 PR with focus on `Scorer.Score` reduction loop |
| `determinism-reviewer` review of golden file harness | SCORER-04 | Process gate verifying no `math.Pow`/`math.FMA`/parallel-reduction patterns slip in | Spawn `Agent(determinism-reviewer)` on plan 08-04 PR with focus on `scorer_golden_test.go` and the `Scorer.Score` reduction |
| `user-guide-reviewer` review of `docs/scorer.md` + `docs/tuning.md` | SCORER-08 | Consumer-perspective review cannot be automated — verifies a developer unfamiliar with the library can use the Scorer from the docs alone | Spawn `Agent(user-guide-reviewer)` on plan 08-04 PR |

---

## Validation Sign-Off

- [ ] All tasks have `<automated>` verify commands OR explicit Wave 0 file-creation dependencies
- [ ] Sampling continuity: no 3 consecutive tasks without an automated verify (audit above: ✓)
- [ ] Wave 0 covers all MISSING test files and golden artefacts
- [ ] No watch-mode flags (root uses `go test` one-shot; BDD uses `godog` one-shot — no watchers)
- [ ] Feedback latency < 5 seconds for the per-task command on a warm cache
- [ ] `nyquist_compliant: true` set in frontmatter once plans are written and Wave 0 file creation is scheduled

**Approval:** pending — flips to `approved 2026-05-16` once gsd-plan-checker passes.
