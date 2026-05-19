---
phase: 9
slug: collection-scan-sub-package
status: draft
nyquist_compliant: false
wave_0_complete: false
created: 2026-05-19
---

# Phase 9 — Validation Strategy

> Per-phase validation contract for feedback sampling during execution.

---

## Test Infrastructure

| Property | Value |
|----------|-------|
| **Framework** | go test (stdlib) — root module; godog + goleak + testify in `tests/bdd/` |
| **Config file** | `Makefile` (canonical targets) |
| **Quick run command** | `go test ./scan/...` |
| **Full suite command** | `make check` |
| **Estimated runtime** | ~30s quick / ~3 min full (excluding fuzz) |

---

## Sampling Rate

- **After every task commit:** Run `go test ./scan/...` (quick — covers new scan code)
- **After every plan wave:** Run `make check` (full — covers root + scan + determinism golden + dependency allowlist)
- **Before `/gsd-verify-work`:** Full `make check` AND `make test-bdd` AND `make verify-determinism` must be green
- **Max feedback latency:** ~30 seconds for quick, ~3 min for full

---

## Per-Task Verification Map

> Populated by gsd-planner during plan creation. Each task in each plan must map to an automated command (or be flagged as Wave 0 dependency).

| Task ID | Plan | Wave | Requirement | Threat Ref | Secure Behavior | Test Type | Automated Command | File Exists | Status |
|---------|------|------|-------------|------------|-----------------|-----------|-------------------|-------------|--------|
| _populated by planner_ | | | | | | | | | ⬜ pending |

*Status: ⬜ pending · ✅ green · ❌ red · ⚠️ flaky*

---

## Wave 0 Requirements

> Populated by gsd-planner. Likely Wave 0 items for Phase 9:

- [ ] `scan/scan.go` — public types skeleton (Item, Kind, Warning, Config) compilable
- [ ] `scan/errors.go` — sentinel errors declared so dependent plans can `errors.Is`
- [ ] `tests/bdd/steps/scan_steps.go` — godog step skeleton so BDD scenarios can compile
- [ ] `testdata/golden/scan-default.json` — placeholder so golden-check has a file to compare against

*Final list confirmed by planner.*

---

## Manual-Only Verifications

| Behavior | Requirement | Why Manual | Test Instructions |
|----------|-------------|------------|-------------------|
| api-ergonomics-reviewer sign-off recording in foundation plan PR | D-01 + D-02 (SPEC OVERRIDE) | Reviewer sign-off is a human gate documented in PR description per Phase 7/8 precedent | Spawn `api-ergonomics-reviewer` agent on foundation plan PR; capture verdict in PR description |
| Empirical bucket-threshold crossover (D-08) | PERF-05, SCAN-02 | Crossover wall-clock differs by hardware — value must be inspected and committed by a human | Run `BenchmarkScanCheck_BucketVsNaive_GroupSize`; inspect benchstat output; update `bucketThreshold` constant + in-source comment with date/wall-clock crossover |

---

## Validation Sign-Off

- [ ] All tasks have `<automated>` verify or Wave 0 dependencies
- [ ] Sampling continuity: no 3 consecutive tasks without automated verify
- [ ] Wave 0 covers all MISSING references
- [ ] No watch-mode flags
- [ ] Feedback latency < 30s for quick, < 3 min for full
- [ ] `nyquist_compliant: true` set in frontmatter

**Approval:** pending
