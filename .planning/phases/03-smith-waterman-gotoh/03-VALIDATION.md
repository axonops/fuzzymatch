---
phase: 3
slug: smith-waterman-gotoh
status: draft
nyquist_compliant: false
wave_0_complete: false
created: 2026-05-14
---

# Phase 3 — Validation Strategy

> Per-phase validation contract for feedback sampling during execution.

---

## Test Infrastructure

| Property | Value |
|----------|-------|
| **Framework** | `go test` (stdlib `testing` + `testing/quick`) — root module; godog + goleak + testify — `tests/bdd` module |
| **Config file** | `go.mod` (root), `tests/bdd/go.mod` (BDD), `Makefile` |
| **Quick run command** | `go test ./... -short -count=1 -run "Test.*SWG\|TestProp_SWG\|TestExample_"` |
| **Full suite command** | `make check` |
| **Estimated runtime** | ~120 seconds (full `make check`); ~10 seconds (quick SWG-scoped) |

---

## Sampling Rate

- **After every task commit:** Run `go test ./... -count=1` (root module) for any task touching root .go files; `cd tests/bdd && go test ./... -count=1` for any task touching BDD steps/features
- **After every plan wave:** Run `make check` (full quality gate: unit + property + BDD + lint + vet + fmt + deps allowlist + determinism golden)
- **Before `/gsd-verify-work`:** Full `make check` must be green AND `make bench-compare` shows no > 10% regression
- **Max feedback latency:** 30 seconds for quick run; 120 seconds for full suite

---

## Per-Task Verification Map

| Task ID | Plan | Wave | Requirement | Threat Ref | Secure Behavior | Test Type | Automated Command | File Exists | Status |
|---------|------|------|-------------|------------|-----------------|-----------|-------------------|-------------|--------|
| 03-01-01 | 03-01 | 1 | CHAR-08 | — | DP kernel returns deterministic float for all valid string pairs | unit + property | `go test ./... -count=1 -run "TestSmithWatermanGotoh\|TestProp_SmithWatermanGotoh"` | ❌ W0 | ⬜ pending |
| 03-01-02 | 03-01 | 1 | CHAR-08 | — | Both-empty / one-empty / identity edge cases return documented constants | unit | `go test ./... -count=1 -run "TestSmithWatermanGotoh_EdgeCases"` | ❌ W0 | ⬜ pending |
| 03-01-03 | 03-01 | 1 | CHAR-08 | — | ASCII fast path triggers for inputs where `n <= 64 && isASCII(a) && isASCII(b)` — 0 B/op, 0 allocs/op | benchmark | `go test ./... -count=1 -bench "BenchmarkSmithWatermanGotohScore_ASCII_Short" -benchmem -run '^$'` | ❌ W0 | ⬜ pending |
| 03-01-04 | 03-01 | 1 | CHAR-08 | — | Score function never produces NaN / +Inf / -Inf / -0 for any input (DET-04) | property | `go test ./... -count=1 -run "TestProp_SmithWatermanGotoh_NoNaN\|_NoInf\|_NoNegativeZero"` | ❌ W0 | ⬜ pending |
| 03-01-05 | 03-01 | 1 | CHAR-08 | — | Score symmetry: `Score(a,b) == Score(b,a)` for byte path AND rune path | property | `go test ./... -count=1 -run "TestProp_SmithWatermanGotohScore_Symmetric\|TestProp_SmithWatermanGotohScoreRunes_Symmetric"` | ❌ W0 | ⬜ pending |
| 03-01-06 | 03-01 | 1 | CHAR-08 | — | Score range `[0.0, 1.0]` for all inputs (clamped normalisation) | property | `go test ./... -count=1 -run "TestProp_SmithWatermanGotohScore_RangeBounds"` | ❌ W0 | ⬜ pending |
| 03-01-07 | 03-01 | 1 | CHAR-08 | — | Identity: `Score(x,x) == 1.0` for non-empty x | property | `go test ./... -count=1 -run "TestProp_SmithWatermanGotohScore_Identity"` | ❌ W0 | ⬜ pending |
| 03-01-08 | 03-01 | 1 | CHAR-08 | — | Gotoh-erratum canary: gap-split-invariance property test fails on erratum-bug implementation, passes on correct implementation | property | `go test ./... -count=1 -run "TestProp_SmithWatermanGotoh_GapSplitInvariance"` | ❌ W0 | ⬜ pending |
| 03-01-09 | 03-01 | 1 | CHAR-08 | — | Raw upper-bound invariant: `RawScore <= Match * min(len(a), len(b))` always | property | `go test ./... -count=1 -run "TestProp_SmithWatermanGotoh_RawNeverExceedsMatchTimesMinLen"` | ❌ W0 | ⬜ pending |
| 03-01-10 | 03-01 | 1 | CHAR-08 | — | Match-reward monotonicity: increasing `Match` cannot decrease `RawScore` | property | `go test ./... -count=1 -run "TestProp_SmithWatermanGotoh_MonotonicWithMatchReward"` | ❌ W0 | ⬜ pending |
| 03-01-11 | 03-01 | 1 | CHAR-08 | — | Fuzz harness: no panics on random `[]byte` inputs over 60s | fuzz | `go test -fuzz "FuzzSmithWatermanGotohScore" -fuzztime 60s ./...` | ❌ W0 | ⬜ pending |
| 03-01-12 | 03-01 | 1 | CHAR-08 | — | BDD: `swg.feature` scenarios pass (identity, both-empty, one-empty, two-substring, custom-params, raw-vs-normalised) | BDD | `cd tests/bdd && go test ./... -count=1 -run "TestFeatures"` | ❌ W0 | ⬜ pending |
| 03-01-13 | 03-01 | 1 | CHAR-08 | — | Dispatch table populated: `dispatch[AlgoSmithWatermanGotoh] == SmithWatermanGotohScore` | unit | `go test ./... -count=1 -run "TestAlgoID_DispatchPopulated"` | ❌ W0 | ⬜ pending |
| 03-02-01 | 03-02 | 1 | CHAR-08 | — | Cross-validation corpus exists: `testdata/cross-validation/swg/vectors.json` parses, contains required edge-case entries (identity, both-empty, one-empty, two-substring, no-overlap, one-long-gap-canary, non-default-params) | unit | `go test ./... -count=1 -run "TestSWG_CrossValidation"` | ❌ W0 | ⬜ pending |
| 03-02-02 | 03-02 | 1 | CHAR-08 | — | Cross-validation green: `|our_normalised - biopython_normalised| <= 1e-9` for every entry in the corpus | unit | `go test ./... -count=1 -run "TestSWG_CrossValidation"` | ❌ W0 | ⬜ pending |
| 03-02-03 | 03-02 | 1 | CHAR-08 | — | Cross-validation corpus regenerable from `scripts/gen-swg-cross-validation.py` (developer-only target) | manual | `make regen-swg-cross-validation` | ❌ W0 | ⬜ pending |
| 03-03-01 | 03-03 | 1 | CHAR-08 | — | Cross-algorithm consistency: SWG scores `http_request` vs `http_request_header_fields` strictly higher than Levenshtein on the same pair (local-alignment-vs-edit-distance divergence) | unit | `go test ./... -count=1 -run "TestCrossAlgorithm_SWG_Levenshtein_SubstringDivergence"` | ❌ W0 | ⬜ pending |
| 03-03-02 | 03-03 | 1 | CHAR-08 | — | Golden file canonical: `testdata/golden/algorithms.json` contains SWG entries; byte-identical across linux/amd64, linux/arm64, darwin/arm64, windows/amd64 | unit (CI) | `go test ./... -count=1 -run "TestGolden_AlgorithmsCanonical"` + cross-platform CI matrix | ❌ W0 | ⬜ pending |
| 03-03-03 | 03-03 | 1 | CHAR-08 | — | `examples/identifier-similarity/main.go` shows SWG column; `TestExample_Output` and `TestExample_ColumnWidths` pass | unit | `go test ./... -count=1 -run "TestExample_Output\|TestExample_ColumnWidths"` | ❌ W0 | ⬜ pending |
| 03-03-04 | 03-03 | 1 | CHAR-08 | — | `bench.txt` updated with SWG benchmark rows; `make bench-compare` shows no > 10% regression on existing benchmarks | benchmark | `make bench && benchstat bench.txt new-bench.txt` | ❌ W0 | ⬜ pending |
| 03-03-05 | 03-03 | 1 | CHAR-08 | — | `llms.txt` lists all 8 new SWG public symbols (SWGParams + NewSWGParams + 6 functions; AlgoSmithWatermanGotoh was added in Phase 1); `TestAIFriendly_AllExportedSymbolsListed` passes | unit | `go test ./... -count=1 -run "TestAIFriendly"` | ❌ W0 | ⬜ pending |
| 03-03-06 | 03-03 | 1 | CHAR-08 | — | `docs/requirements.md` §7.1.8 updated to list all 6 SWG public functions (3 normalised + 3 Raw) | docs | `grep -c "SmithWatermanGotoh.*Score" docs/requirements.md` returns ≥ 6 | ❌ W0 | ⬜ pending |
| 03-03-07 | 03-03 | 1 | CHAR-08 | — | Coverage: ≥ 95% overall, ≥ 90% on `swg.go`, 100% on public API surface | unit | `make coverage` | ❌ W0 | ⬜ pending |
| 03-03-08 | 03-03 | 1 | CHAR-08 | — | Determinism golden: cross-platform byte-identical via CI matrix | unit (CI) | `make verify-determinism` | ❌ W0 | ⬜ pending |

*Status: ⬜ pending · ✅ green · ❌ red · ⚠️ flaky*

---

## Wave 0 Requirements

Phase 3 inherits all test infrastructure from Phase 1 + Phase 2. No new Wave 0 framework setup required. New file creation per task — all listed above as `❌ W0` indicating the file will be created during the task's execution, not pre-staged.

New files Phase 3 creates (none pre-staged at Wave 0):
- [ ] `swg.go` — implementation
- [ ] `dispatch_swg.go` — dispatch registration
- [ ] `swg_test.go` — unit + cross-validation tests
- [ ] `swg_bench_test.go` — benchmark
- [ ] `swg_fuzz_test.go` — fuzz harness
- [ ] `tests/bdd/features/swg.feature` — BDD scenarios
- [ ] `testdata/golden/_staging/swg.json` — staging golden
- [ ] `testdata/fuzz/FuzzSmithWatermanGotohScore/seed-001` — fuzz seed
- [ ] `testdata/cross-validation/swg/vectors.json` — biopython cross-validation corpus
- [ ] `scripts/gen-swg-cross-validation.py` — corpus generator (developer-only)

Existing files Phase 3 extends (append-only per Phase 2 lock):
- `props_test.go` — append SWG property tests
- `example_test.go` — append SWG godoc example
- `algoid_test.go` — append SWG dispatch test
- `tests/bdd/steps/algorithms_steps.go` — append SWG step bindings
- `testdata/golden/algorithms.json` — merged from staging in 03-03
- `cross_algorithm_consistency_test.go` — append SWG-vs-Levenshtein divergence test
- `examples/identifier-similarity/main.go` — add SWG column (7 → 8)
- `bench.txt` — append SWG benchmark rows
- `llms.txt` — append SWG public symbols (9 lines)
- `docs/requirements.md` §7.1.8 — record Raw* surface expansion
- `Makefile` — add `regen-swg-cross-validation` target

---

## Manual-Only Verifications

| Behavior | Requirement | Why Manual | Test Instructions |
|----------|-------------|------------|-------------------|
| biopython cross-validation corpus regeneration | CHAR-08 | Requires Python + biopython 1.85+ runtime; not available in CI | `pip install biopython && make regen-swg-cross-validation && git diff testdata/cross-validation/swg/vectors.json` |
| Cross-platform determinism (linux/amd64, linux/arm64, darwin/arm64, windows/amd64) | CHAR-08 (DET-02) | Requires CI matrix; not runnable on a single developer machine | Push branch to PR — CI matrix runs `make verify-determinism` on all four targets |

---

## Validation Sign-Off

- [ ] All tasks have `<automated>` verify or Wave 0 dependencies
- [ ] Sampling continuity: no 3 consecutive tasks without automated verify
- [ ] Wave 0 covers all MISSING references (Phase 3 has no Wave 0 — all new files created during tasks)
- [ ] No watch-mode flags
- [ ] Feedback latency < 30s (quick) / 120s (full)
- [ ] `nyquist_compliant: true` set in frontmatter (after planner finishes — planner certifies dimension 8 coverage)

**Approval:** pending
