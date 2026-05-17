---
phase: 08
plan: 08-04
subsystem: composite-scorer
tags: [scorer, phase-8, finalisation, golden, bdd, benchmark, docs, requirements]
one_liner: "Phase 8 finalisation — cross-platform golden + BDD coverage + bench fixture + extended example + new composition example + populated docs + spec amendments"
status: complete
requires:
  - Phase 8 plans 08-01 (option layer), 08-02 (NewScorer + Score + Match), 08-03 (ScoreAll + accessors + DefaultScorer)
  - Phase 1-7 (23-algorithm catalogue + dispatch + Normalise + Tokenise)
provides:
  - testdata/golden/scorer-default.json — 22-entry cross-platform determinism gate
  - tests/bdd/features/scorer.feature + tests/bdd/steps/scorer_steps.go — 12 BDD scenarios
  - scorer_bench_test.go — 6 allocation-aware benchmarks
  - examples/scorer-composition/ — new runnable example program
  - examples/identifier-similarity Score + Match columns
  - docs/scorer.md (340 lines) — Scorer API guide
  - docs/tuning.md (208 lines) — threshold + weight calibration guidance
  - docs/requirements.md §8.3/§8.6/§13.4 — SPEC OVERRIDE annotations
  - .planning/REQUIREMENTS.md SCORER-01..08 Complete + SCORER-08 typo fix
affects: []
tech-stack:
  added: []
  patterns:
    - "stdout golden test pattern (os.Pipe redirect + line-by-line diff) — second usage in examples/"
    - "BDD step context per feature (ScorerContext separate from AlgorithmContext)"
    - "InitScorerSteps called from InitializeScenario — extension pattern"
key-files:
  created:
    - scorer_golden_test.go
    - scorer_bench_test.go
    - testdata/golden/scorer-default.json
    - tests/bdd/features/scorer.feature
    - tests/bdd/steps/scorer_steps.go
    - examples/scorer-composition/main.go
    - examples/scorer-composition/main_test.go
    - examples/scorer-composition/go.mod
    - examples/scorer-composition/go.sum
  modified:
    - examples/identifier-similarity/main.go
    - examples/identifier-similarity/main_test.go
    - docs/scorer.md
    - docs/tuning.md
    - docs/requirements.md
    - .planning/REQUIREMENTS.md
    - tests/bdd/steps/algorithms_steps.go
decisions:
  - Golden file omits _metadata.generated_at to match algorithms.json byte-stability pattern (RESEARCH.md Open Question 1)
  - scoreAll keys in JSON use AlgoID.String() not integer enum values (RESEARCH.md Pitfall 6)
  - InitScorerSteps wired from existing InitializeScenario (single TestMain preserved)
  - Allocation budget exceedance recorded as manual-review item, not test failure
  - llms.txt + llms-full.txt already complete from prior plans — no update needed
metrics:
  duration: "21 min 5 sec"
  completed: 2026-05-17T05:37Z
---

# Phase 8 Plan 08-04: Finalisation Summary

Plan 08-04 is the four-pillar finalisation of Phase 8: cross-platform
determinism gate (golden file), consumer contract (BDD), performance
fixture (benchmark), and the consumer-facing surfaces (examples + docs
+ spec amendments). All six tasks executed atomically; six commits
referencing 08-04 land on the worktree branch.

## Files

### Created (9)

| File                                            | Purpose                                                            |
| ----------------------------------------------- | ------------------------------------------------------------------ |
| `scorer_golden_test.go`                         | Cross-platform float-determinism golden test for DefaultScorer + 4 variants |
| `testdata/golden/scorer-default.json`           | 22-entry corpus, byte-stable across CI matrix                      |
| `scorer_bench_test.go`                          | 6 allocation-aware benchmarks (Score/ScoreAll/Match × Short/Med/Long/Unicode) |
| `tests/bdd/features/scorer.feature`             | 12 BDD scenarios covering all mandatory classes from CONTEXT.md §7 |
| `tests/bdd/steps/scorer_steps.go`               | ScorerContext + InitScorerSteps registration                       |
| `examples/scorer-composition/main.go`           | DefaultScorerOptions + WithoutAlgorithm + WithThreshold demo       |
| `examples/scorer-composition/main_test.go`      | stdout golden gate                                                 |
| `examples/scorer-composition/go.mod`            | per-example module isolation                                       |
| `examples/scorer-composition/go.sum`            | per-example module isolation                                       |

### Modified (7)

| File                                            | Change                                                              |
| ----------------------------------------------- | ------------------------------------------------------------------- |
| `examples/identifier-similarity/main.go`        | Added Score + Match columns from DefaultScorer at end of table      |
| `examples/identifier-similarity/main_test.go`   | Regenerated `want` constant with 2 extra columns (25 total)         |
| `tests/bdd/steps/algorithms_steps.go`           | Added `InitScorerSteps(ctx)` call at end of `InitializeScenario`    |
| `docs/scorer.md`                                | Scaffold → full 340-line consumer guide                             |
| `docs/tuning.md`                                | Scaffold → full 208-line calibration guide                          |
| `docs/requirements.md`                          | §8.3 + §8.6 + §13.4 SPEC OVERRIDE annotations (map[AlgoID]float64) |
| `.planning/REQUIREMENTS.md`                     | SCORER-08 typo fix + all 8 SCORER-* flipped to Complete             |

## Commits

| # | Hash    | Subject                                                                  |
| - | ------- | ------------------------------------------------------------------------ |
| 1 | 57302d6 | test(08-04): add scorer-default.json cross-platform determinism gate     |
| 2 | dfb5a59 | test(08-04): add Phase 8 Scorer allocation-aware benchmark fixture       |
| 3 | 5932788 | test(08-04): add Scorer BDD scenarios and step definitions               |
| 4 | b7867ba | feat(08-04): extend identifier-similarity + add scorer-composition examples |
| 5 | 7a15dc5 | docs(08-04): populate docs/scorer.md and docs/tuning.md from scaffold    |
| 6 | a4b49c9 | docs(08-04): amend §8.3 ScoreAll spec + flip SCORER-* to Complete        |

## Golden File Composition

`testdata/golden/scorer-default.json` contains **22 entries** spanning
**5 Scorer compositions**:

| scorer_config                          | Entry count | Purpose                                                  |
| -------------------------------------- | ----------- | -------------------------------------------------------- |
| `DefaultScorer`                        | 14          | Identifier-similarity corpus + threshold edges + Unicode + phonetic + identity / empty / dissimilar |
| `DefaultScorer-WithoutNormalisation`   | 2           | Raw-byte comparison for XMLParser / xml_parser + café / cafe |
| `DefaultScorer-MinusDoubleMetaphone`   | 1           | Smith / Schmidt without phonetic signal                  |
| `Levenshtein-Only-Threshold-0.5`       | 2           | Minimum viable single-algorithm Scorer                   |
| `Raw-Weights-Lev-1-JW-3-NoNorm`        | 3           | WithNormaliseWeights(false) — composite may exceed 1.0   |

The file is **byte-stable** by construction:

- `_metadata.generated_at` is INTENTIONALLY OMITTED (matches algorithms.json
  pattern; without this the file would diff on every `-update` regen).
- `scoreAll` keys use `AlgoID.String()` strings (e.g. `"DamerauLevenshteinOSA"`)
  not integer enum values — grep-friendly without sacrificing the
  underlying typed-key public API.
- `json.MarshalIndent` with two-space indent sorts map keys
  alphabetically, so the on-disk byte form is deterministic.

Cross-platform CI matrix gate is **pending** — load-bearing manual
review item #10 per VALIDATION.md.

## BDD Scenarios

`tests/bdd/features/scorer.feature` contains **12 scenarios** (within
the [8, 12] target band). Each maps to one or more of the mandatory
classes from CONTEXT.md §7:

| Scenario                                                             | CONTEXT.md §7 class                |
| -------------------------------------------------------------------- | ---------------------------------- |
| Default scorer matches identifier-style pair                         | 1. Default happy path              |
| Default scorer rejects dissimilar inputs                             | 2. Default below threshold         |
| Single-algorithm Scorer composes correctly                           | 3. Custom 1-algorithm              |
| Two-algorithm weighted Scorer composes correctly                     | 4. Custom 2-algorithm weighted     |
| WithoutAlgorithm removes the algorithm                               | 5. WithoutAlgorithm                |
| Duplicate WithAlgorithm calls use the latest weight                  | 6. Last-write-wins                 |
| WithoutNormalisation preserves raw bytes                             | 7. WithoutNormalisation            |
| NewScorer without WithThreshold returns ErrMissingThreshold          | 8. ErrMissingThreshold             |
| NewScorer without algorithms returns ErrEmptyScorer                  | 9. ErrEmptyScorer                  |
| WithAlgorithm with negative weight returns ErrInvalidWeight          | 10. ErrInvalidWeight               |
| Concurrent Score calls return identical results                      | 11. Concurrent (goleak gate)       |
| ScoreAll returns map keyed by AlgoID                                 | 12. ScoreAll AlgoID keys           |

All 12 classes represented. The full BDD suite runs **271 scenarios
green** (including the 12 new Scorer scenarios). `goleak.VerifyTestMain`
in `tests/bdd/bdd_test.go` reports **zero goroutine leaks** — the
100-goroutine concurrent scenario uses `sync.WaitGroup` for
deterministic completion (no errgroup, no context.Context).

## Performance Budget

Measured on **darwin/arm64 (Apple M2)** at `go test -bench -count=3`:

| Benchmark                                       | ns/op    | B/op   | allocs/op | Budget          | Pass? |
| ----------------------------------------------- | -------: | -----: | --------: | --------------- | ----- |
| BenchmarkDefaultScorer_Score_ASCII_Short        |    925   |   192  |       12  | < 30 µs / ≤ 8   | wall ✓, allocs ✗ |
| BenchmarkDefaultScorer_Score_ASCII_Medium       |   8101   |  6208  |       34  | < 30 µs / ≤ 8   | wall ✓, allocs ✗ |
| BenchmarkDefaultScorer_Score_ASCII_Long         | 946742   | 137024 |       37  | informational   | n/a   |
| BenchmarkDefaultScorer_Score_Unicode_Short      |   1428   |   872  |       21  | informational   | n/a   |
| BenchmarkDefaultScorer_ScoreAll_ASCII_Short     |   1025   |   384  |       14  | informational   | n/a   |
| BenchmarkDefaultScorer_Match_ASCII_Short        |    925   |   192  |       12  | match ≈ score   | ✓     |

### Performance Budget Investigation Required

**The `< 30 µs / ≤ 8 allocs` budget for ASCII Short/Medium is met on
wall-time but exceeded on allocations:**

- ASCII Short: 12 allocs/op (budget ≤ 8) — 50% over budget
- ASCII Medium: 34 allocs/op (budget ≤ 8) — 4× over budget

**Root cause hypothesis (to be confirmed by algorithm-performance-reviewer):**

- DefaultScorer composes 6 algorithms; each algorithm's per-call
  allocations (1-3 each for QGram/Sørensen/TokenJaccard maps) compound
  to ~12 allocs at the smallest input size.
- Normalise allocates the pre-normalised strings once per Score call
  (1-2 allocs).
- For ASCII Medium (30 chars), the QGram maps grow beyond their
  initial capacity hint → re-hash allocations dominate.

**Recommended escalation:** spawn `algorithm-performance-reviewer`
against this plan's PR with the benchmark output. The reviewer's
verdict determines whether:
1. The budget is revised (currently set to a 6-algorithm composite —
   may need to relax for the n-of-6 ensemble case), or
2. The implementation needs allocation-pooling (sync.Pool for the
   per-algorithm intermediate maps), or
3. Some algorithms' per-call allocation patterns can be reduced (e.g.
   stack-allocated buffers for short ASCII inputs).

Phase 8 ships with the observed numbers documented; the budget
investigation is a separate review track per VALIDATION.md manual
gate #4. The benchmark does NOT enforce the budget as `t.Fatal` —
that would couple the test to specific runner hardware.

## Documentation

### docs/scorer.md (340 lines)

Sections:

- Overview (immutability + concurrency)
- Quickstart (5-line `DefaultScorer` code block from CONTEXT.md §Specific Ideas)
- Custom Composition (with the `WithThreshold` REQUIRED comment)
- Default minus one algorithm (canonical `WithoutAlgorithm(AlgoDoubleMetaphone)` pattern)
- Parameterised algorithm options (n=4 QGram, α=β=0.5 Tversky, etc.)
- Method Reference table (Score / Match / ScoreAll / Threshold / Algorithms)
- Threshold rationale (why no library-wide default)
- Default Composition table (6 algorithms + threshold rationale)
- Normalisation Control (WithNormalisation / WithoutNormalisation)
- Note on ME's vestigial opts parameter (CONTEXT.md §8 — no-op when invoked through Scorer)
- ScoreAll Method (typed-AlgoID-key behaviour + SPEC OVERRIDE annotation)
- Algorithms accessor (fresh slice + AlgoID-sorted order)
- Concurrency guarantees (immutable + race-clean)
- Errors table (ErrMissingThreshold / ErrEmptyScorer / ErrInvalidWeight / ErrInvalidThreshold)
- Weight Normalisation (auto-normalise default + raw-weights override)
- Last-write-wins for duplicate AlgoIDs
- See also (links to tuning.md / algorithms.md / requirements.md / examples)

### docs/tuning.md (208 lines)

Sections:

- Overview (heuristics, not laws)
- How to pick a threshold (6-step calibration loop with 0.05 increments)
- How to pick weights (correlation-based heuristic via ScoreAll profiling)
- DefaultScorer composition rationale (per-algorithm justification table)
- Choosing the right algorithm subset (per-domain heuristics for identifiers / names / pronunciation / substring / token-order / long-string)
- Performance / accuracy trade-off (forward-reference to docs/performance.md)
- Pinning a calibrated configuration (package-level `var` pattern)

### markdownlint outcome

`markdownlint-cli2` was **not on PATH** at execution time; lint
verification deferred to `make check` / CI. The v2 markdown ruleset is
configured in the project's `.markdownlint-cli2.yaml`.

## Spec Amendments

### docs/requirements.md

- **§8.3** `ScoreAll` signature changed from `map[string]float64` →
  `map[AlgoID]float64` with a SPEC OVERRIDE block citing 08-CONTEXT.md
  §1 + the api-ergonomics-reviewer sign-off recorded in plan 08-03's PR.
- **§8.6** prose updated to describe the typed-AlgoID-keyed map.
- **§13.4** determinism note updated to match.
- The two remaining `map[string]float64` references in §8.3/§8.6 are
  **inside** the SPEC OVERRIDE annotations (citing the historical form
  for context) — not live specifications.

### .planning/REQUIREMENTS.md

- **SCORER-08** description amended from `WithCustomNormalisation()`
  → `WithNormalisation(opts)` (per CONTEXT.md Deferred Ideas — typo
  fix carried into Phase 8 finalisation).
- **SCORER-01..SCORER-08** status flipped from `Pending` → `Complete`
  in both the checkbox list (lines 61-68) and the Traceability table
  (lines 229-236).

## Threat Flags

No new threat-flag entries — all threats listed in plan 08-04's
`<threat_model>` (T-08-03 examples DoS, T-08-04 cross-platform float
divergence, T-08-05 BDD goroutine leak) are mitigated by the
shipped artefacts:

- T-08-03: example programs use hardcoded input pairs; no consumer-
  supplied data path.
- T-08-04: golden file's cross-platform CI gate is the divergence
  detector (manual review #10 below).
- T-08-05: goleak.VerifyTestMain in tests/bdd/bdd_test.go reports
  zero goroutine leaks across all 271 BDD scenarios.

## Deviations from Plan

### Rule 2 (auto-add missing critical functionality)

**1. Per-example go.mod / go.sum for scorer-composition**
- **Found during:** Task 4
- **Issue:** The plan did not specify creating `go.mod` / `go.sum` for
  the new `examples/scorer-composition/` directory, but the existing
  `examples/identifier-similarity/` and `examples/phonetic-keys/` both
  have their own modules with a `replace` directive pointing at the
  root. Without these the new example would not compile under `go run
  ./examples/scorer-composition/`.
- **Fix:** Created `go.mod` + `go.sum` mirroring the
  identifier-similarity per-example module pattern.
- **Files modified:** `examples/scorer-composition/go.mod`,
  `examples/scorer-composition/go.sum`
- **Commit:** b7867ba

### Rule 1/2 — None applied beyond the above.

### Performance budget exceedance (NOT a deviation)

The allocs/op > 8 observation on ASCII Short/Medium benchmarks is
**not** a Rule 1/2 fix. The plan explicitly states: "If allocs/op > 8
on ASCII Short OR µs/op > 30 on ASCII Medium, the executor flags the
regression and spawns algorithm-performance-reviewer... Do NOT fail
this plan's automated verify on the budget — the budget verification
is a manual gate per VALIDATION.md." The exceedance is recorded above
under "Performance Budget Investigation Required" for the manual
reviewer.

## Manual Review Gates Status

Per VALIDATION.md "Manual-Only Verifications":

| # | Reviewer / Gate                              | State    | Notes                                                                                            |
| - | -------------------------------------------- | -------- | ------------------------------------------------------------------------------------------------ |
| 1 | api-ergonomics-reviewer (ScoreAll SPEC OVERRIDE) | Pending | Re-confirm sign-off persists from plan 08-03's PR (originally recorded there)                    |
| 2 | algorithm-correctness-reviewer (composite reduction loop) | Pending | Re-confirm sign-off persists from plan 08-02's PR                                                  |
| 3 | determinism-reviewer (scorer_golden_test + scorer-default.json) | Pending | First review for plan 08-04 artefact                                                              |
| 4 | algorithm-performance-reviewer (bench fixture) | **Required** | allocs/op exceeds ≤ 8 on ASCII Short (12) + Medium (34) — see "Performance Budget" above        |
| 5 | user-guide-reviewer (docs/scorer.md + docs/tuning.md) | Pending | First review for plan 08-04 artefact                                                              |
| 6 | bdd-scenario-reviewer (scorer.feature + scorer_steps.go) | Pending | First review for plan 08-04 artefact                                                              |
| 7 | test-writer + test-analyst (coverage audit) | Pending | Pre-milestone audit; ≥ 95% overall / ≥ 90% per file / 100% public API                              |
| 8 | code-reviewer (all changed files)             | Pending | General review                                                                                   |
| 9 | commit-message-reviewer (all 08-04 commits)   | Pending | Every commit in this plan's PR                                                                   |
| 10| Cross-platform CI matrix gate                 | Pending | Post-merge CI on linux/{amd64,arm64} + darwin/{amd64,arm64} + windows/amd64 must show byte-identical `scorer-default.json` |

## Verification Results

| Check                                                             | Result |
| ----------------------------------------------------------------- | ------ |
| `go test -race -count=1 ./...` (root suite)                       | PASS   |
| `cd tests/bdd && go test -race -count=1 ./...` (BDD)              | PASS — 271 scenarios green |
| `go test -run TestGolden_ScorerDefault ./...`                     | PASS   |
| `go test -run TestAIFriendly ./...` (llms.txt sync meta-test)     | PASS   |
| `go test -count=1 ./examples/identifier-similarity/...`           | PASS   |
| `go test -count=1 ./examples/scorer-composition/...`              | PASS   |
| `go test -bench=BenchmarkDefaultScorer_Score -benchmem -run=^$`   | PASS (benchmark runs cleanly; numbers recorded above) |
| `gofmt -l .` (excluding .claude/.planning)                        | clean  |
| `go vet ./...`                                                    | clean  |
| `jq '.entries | length' testdata/golden/scorer-default.json`      | 22 (within [22, 26])  |
| `grep -cE "Scenario:" tests/bdd/features/scorer.feature`          | 12 (within [8, 12])   |
| `awk '/^\| SCORER-0[1-8] \|/' .planning/REQUIREMENTS.md \| grep -c "Complete"` | 8     |

## Phase 8 Declaration

**Phase 8 complete — ready for /gsd-verify-work and /gsd-complete-phase.**

All four plans (08-01 → 08-04) have landed. All 8 SCORER-* requirements
flipped to Complete in `.planning/REQUIREMENTS.md`. The composite
Scorer surface is shipped end-to-end with cross-platform determinism
gate + BDD contract + populated documentation + two runnable example
programs + spec amendments aligned with implementation.

The 10 manual review gates above are the remaining work; they are
review-tracked, not implementation work. Once the gates clear,
Phase 8 advances to `complete` in STATE.md.

## Self-Check: PASSED

- `scorer_golden_test.go`: FOUND
- `testdata/golden/scorer-default.json`: FOUND (22 entries verified via jq)
- `scorer_bench_test.go`: FOUND (6 benchmark functions verified via grep)
- `tests/bdd/features/scorer.feature`: FOUND (12 scenarios verified via grep)
- `tests/bdd/steps/scorer_steps.go`: FOUND
- `examples/scorer-composition/main.go`: FOUND
- `examples/scorer-composition/main_test.go`: FOUND
- `examples/scorer-composition/go.mod`: FOUND
- `examples/scorer-composition/go.sum`: FOUND
- `docs/scorer.md`: 340 lines (≥ 80 budget)
- `docs/tuning.md`: 208 lines (≥ 40 budget)
- Commit `57302d6`: FOUND in `git log --oneline`
- Commit `dfb5a59`: FOUND in `git log --oneline`
- Commit `5932788`: FOUND in `git log --oneline`
- Commit `b7867ba`: FOUND in `git log --oneline`
- Commit `7a15dc5`: FOUND in `git log --oneline`
- Commit `a4b49c9`: FOUND in `git log --oneline`
- All tests green (root + BDD + examples).
- All 8 SCORER-* status entries show `Complete` in .planning/REQUIREMENTS.md.
- llms.txt + llms-full.txt already complete from prior plans (TestAIFriendly green).
