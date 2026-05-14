---
phase: 2
slug: 02-core-character-algorithms-six
status: draft
nyquist_compliant: false
wave_0_complete: false
created: 2026-05-14
---

# Phase 2 — Validation Strategy

> Per-phase validation contract for feedback sampling during execution.
> Source: `02-RESEARCH.md` § Validation Architecture (committed at `ecc309a`).

---

## Test Infrastructure

| Property | Value |
|----------|-------|
| **Framework** | stdlib `testing` + `testing/quick` (root); godog v0.15.0 (BDD sub-module at `tests/bdd/`) |
| **Config file** | `.golangci.yml` (already in place from Phase 1); `Makefile`; `tests/bdd/go.mod` |
| **Quick run command** | `go test -race -shuffle=on -count=1 ./...` |
| **Full suite command** | `make check` |
| **Estimated runtime** | ~30-60s for `go test ./...`; ~3-5 min for `make check` |

---

## Sampling Rate

- **After every task commit:** Run `go test -race -shuffle=on -count=1 -run <TestPattern> ./...` (scoped to the algorithm just touched)
- **After every plan wave:** Run `make check` (full quality gate: vet, lint, test, golden verify, benchstat regression check)
- **Before `/gsd-verify-work`:** Full suite green on all five CI platforms (linux/amd64, linux/arm64, darwin/amd64, darwin/arm64, windows/amd64)
- **Max feedback latency:** ~60s for per-task; ~5 min for per-wave

---

## Per-Task Verification Map

> Populated by the planner during `/gsd-plan-phase`. Each row maps a PLAN.md task to the
> automated command that proves its acceptance criteria. Filled after `## PLANNING COMPLETE`.

| Task ID | Plan | Wave | Requirement | Threat Ref | Secure Behavior | Test Type | Automated Command | File Exists | Status |
|---------|------|------|-------------|------------|-----------------|-----------|-------------------|-------------|--------|
| _to be populated_ | _by planner_ | — | — | — | — | — | — | — | ⬜ pending |

*Status: ⬜ pending · ✅ green · ❌ red · ⚠️ flaky*

### Phase-level Requirement → Test Map (from RESEARCH.md)

| Req ID | Behavior | Test Type | Automated Command |
|--------|----------|-----------|-------------------|
| CHAR-01 | Levenshtein distance/score byte+rune correct | unit | `go test -run TestLevenshtein ./...` |
| CHAR-02 | DL-OSA correct + discriminating vector `"ca"`/`"abc"` → 3 | unit | `go test -run TestDamerauLevenshteinOSA ./...` |
| CHAR-03 | DL-Full correct + discriminating vector `"ca"`/`"abc"` → 2 | unit | `go test -run TestDamerauLevenshteinFull ./...` |
| CHAR-04 | Hamming byte+rune + unequal-length silent zero | unit | `go test -run TestHamming ./...` |
| CHAR-05 | Jaro byte+rune correct | unit | `go test -run TestJaro$ ./...` |
| CHAR-06 | Jaro-Winkler byte+rune + Winkler 1990 constants | unit | `go test -run TestJaroWinkler ./...` |
| PERF-01 | 0 allocs on ASCII ≤ 50 chars per algorithm | benchmark | `go test -bench=. -benchmem -run=^$ ./...` |
| PERF-02 | ASCII fast path proven by escape analysis | code review | `go build -gcflags="-m=2" ./... 2>&1` |
| PERF-03 | Two-row DP, no full DP table allocation | code review | manual + `gcflags=-m` |
| TEST-01 | Reference vectors from primary sources cited inline | unit | `go test -run TestXxx_ReferenceVectors ./...` |
| TEST-02 | Property tests (range, identity, symmetry, triangle, NaN, Inf) | property | `go test -run TestProp ./...` |
| TEST-04 | Benchmarks with `b.ReportAllocs()` + alloc assertions | benchmark | `go test -bench=. -benchmem -run=^$ ./...` |
| TEST-05 | BDD scenarios per algorithm | BDD | `cd tests/bdd && go test ./...` |
| DET-02 | Scores stable in golden file `testdata/golden/algorithms.json` | golden | `make verify-determinism` |
| DET-04 | No NaN, Inf, or negative-zero on any path | property | `go test -run TestProp_.*NoNaN ./... && go test -run TestProp_.*NoInf ./...` |
| DX-02 | godoc + Example function per algorithm | unit | `go test -run ExampleXxx ./...` |
| DX-05 | `examples/identifier-similarity/` runs and meta-test passes | meta | `go test ./examples/identifier-similarity/...` |

---

## Wave 0 Requirements

All of the following must be created — none exist yet (per RESEARCH.md inventory):

**Algorithm files (one .go per algorithm):**
- [ ] `levenshtein.go`
- [ ] `damerau_osa.go`
- [ ] `damerau_full.go`
- [ ] `hamming.go`
- [ ] `jaro.go`
- [ ] `jarowinkler.go` *(final filename TBD by planner per RESEARCH.md Open Question #1)*

**Dispatch registration (separate file per algorithm to avoid Wave 2 merge conflicts):**
- [ ] `dispatch_levenshtein.go`
- [ ] `dispatch_damerau_osa.go`
- [ ] `dispatch_damerau_full.go`
- [ ] `dispatch_hamming.go`
- [ ] `dispatch_jaro.go`
- [ ] `dispatch_jarowinkler.go`

**Tests (unit + property + bench + fuzz per algorithm):**
- [ ] `levenshtein_test.go`, `levenshtein_bench_test.go`, `levenshtein_fuzz_test.go`
- [ ] `damerau_osa_test.go`, `damerau_osa_bench_test.go`, `damerau_osa_fuzz_test.go`
- [ ] `damerau_full_test.go`, `damerau_full_bench_test.go`, `damerau_full_fuzz_test.go`
- [ ] `hamming_test.go`, `hamming_bench_test.go`, `hamming_fuzz_test.go`
- [ ] `jaro_test.go`, `jaro_bench_test.go`, `jaro_fuzz_test.go`
- [ ] `jarowinkler_test.go`, `jarowinkler_bench_test.go`, `jarowinkler_fuzz_test.go`
- [ ] `props_test.go` — phase-2 property tests (range, identity, symmetry, triangle inequality, NaN/Inf/-0 absence)

**Cross-cutting test infrastructure:**
- [ ] `algorithms_golden_test.go` + `testdata/golden/algorithms.json` (extends Phase 1 golden infra)
- [ ] `example_test.go` — `ExampleXxx` runnable godoc examples for all six algorithms
- [ ] `tests/bdd/features/algorithms.feature` (or split per-algorithm `.feature` files — planner decides)
- [ ] `tests/bdd/steps/algorithms_steps.go` (new or extended steps wiring)

**Runnable example with meta-test:**
- [ ] `examples/identifier-similarity/main.go`
- [ ] `examples/identifier-similarity/main_test.go` (meta-test asserting stable output)

**First benchstat baseline:**
- [ ] `bench.txt` committed at repo root (output of `make bench` after Wave 1 lands)

**Existing infrastructure inherited from Phase 1 (no gaps):**
- `golden_test.go` (`assertGolden`, `-update` flag)
- `golden_canonical.go` (`canonicalMarshal`, `WriteGoldenFile`)
- `export_test.go` (`CanonicalMarshalForTest`, `DispatchLenForTest`, `DispatchEntryNilForTest`)
- `tests/bdd/go.mod` (godog + goleak + testify, replace directive to root module)
- `Makefile` quality targets (`check`, `verify-determinism`, `bench`, `bench-compare`, `verify-deps-allowlist`)
- CI workflows (cross-platform matrix, security, codeql, release)

---

## Manual-Only Verifications

| Behavior | Requirement | Why Manual | Test Instructions |
|----------|-------------|------------|-------------------|
| Two-row DP confirmed (no full DP table) | PERF-03 | Structural property of source code, not runtime-observable beyond allocation count (which PERF-01 covers indirectly) | Reviewer reads each DP algorithm file; confirms inner loop maintains exactly two `[N]int` rows (or one row + scalar) and never allocates `[m+1][n+1]int` |
| ASCII fast path exists in source | PERF-02 | Verified via escape-analysis output, not pure runtime test | Run `go build -gcflags="-m=2" ./... 2>&1 \| grep -E "(does not escape\|escapes to heap)"` and confirm hot path stack-allocates the buffer |
| Primary-source citation present at top of each `.go` file | TEST-01 / algorithm-correctness skill | Source-code convention, not runtime | Reviewer + `algorithm-correctness-reviewer` agent confirm `// Source:` block with paper, year, page/section at top of each algorithm file |
| Patent screen negative finding for each algorithm | algorithm-licensing skill | Licensing audit, not runtime | `algorithm-licensing-reviewer` agent runs before implementation begins; documents negative finding |

---

## Validation Sign-Off

- [ ] All tasks have `<automated>` verify or Wave 0 dependencies (populated after planner runs)
- [ ] Sampling continuity: no 3 consecutive tasks without automated verify
- [ ] Wave 0 covers all MISSING references
- [ ] No watch-mode flags
- [ ] Feedback latency < 60s per task, < 5 min per wave
- [ ] `nyquist_compliant: true` set in frontmatter (after planner populates per-task map)

**Approval:** pending
