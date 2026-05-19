---
phase: 9
slug: collection-scan-sub-package
status: ready
nyquist_compliant: true
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

Every task in Plans 09-01 through 09-08 maps to an automated verify command (or a Wave 0 dependency closing in Plan 09-01). Source: each plan's `<verify><automated>...</automated></verify>` block per task.

| Task | Plan | Wave | Requirement | Threat Ref | Test Type | Automated Command | Status |
|------|------|------|-------------|------------|-----------|-------------------|--------|
| 09-01-01 — scan/ skeleton (Item, Kind, Warning, Config, sentinels, DefaultConfig) | 01 | 1 | SCAN-01, SCAN-06 | — | unit | `go test -race -shuffle=on -count=1 ./scan/...` | ⬜ pending |
| 09-01-02 — Scorer.NormalisationOptions() accessor (Open Question 1) | 01 | 1 | SCAN-01 | — | unit | `go test -race -shuffle=on -count=1 ./...` | ⬜ pending |
| 09-01-03 — Spec amendments (§12.1 SPEC OVERRIDE) + llms/CHANGELOG sync | 01 | 1 | SCAN-01 | — | doc | `markdownlint-cli2 docs/requirements.md llms.txt llms-full.txt CHANGELOG.md && grep -v '^#' docs/requirements.md \| grep -c WarningKind \| grep -q '^0$'` | ⬜ pending |
| 09-01-04 — Final commit + 5-reviewer panel | 01 | 1 | SCAN-01, SCAN-06 | — | gate | `make check` | ⬜ pending |
| 09-02-01 — validateCheck pipeline (P1 Scorer-nil fail-fast) | 02 | 2 | SCAN-01, SCAN-06 | — | unit | `go test -race -shuffle=on -count=1 -run TestValidateCheck ./scan/...` | ⬜ pending |
| 09-02-02 — P2 Config validation (CrossGroupThresholdBoost NaN/Inf/range) | 02 | 2 | SCAN-06 | — | unit | `go test -race -shuffle=on -count=1 -run TestValidateConfigFields ./scan/...` | ⬜ pending |
| 09-02-03 — P3+P4 collect-all errors.Join (D-03/D-05/D-06) + final gate | 02 | 2 | SCAN-01, SCAN-06 | — | unit | `make check` | ⬜ pending |
| 09-03-01 — Check body: normalise pass + within-group naive scan | 03 | 3 | SCAN-01, SCAN-04 | — | unit | `go test -race -shuffle=on -count=1 -run TestCheck_WithinGroup ./scan/...` | ⬜ pending |
| 09-03-02 — Cross-group naive pass + threshold-boost clamp + identical-name default | 03 | 3 | SCAN-01, SCAN-04 | — | unit | `go test -race -shuffle=on -count=1 -run TestCheck_CrossGroup ./scan/...` | ⬜ pending |
| 09-03-03 — Final commit + gate | 03 | 3 | SCAN-01, SCAN-04 | — | gate | `make check` | ⬜ pending |
| 09-04-01 — bucket.go + bucketThreshold private constant + tokeniseAll | 04 | 4 | SCAN-02, PERF-05 | — | unit | `go test -race -shuffle=on -count=1 -run TestBucket ./scan/...` | ⬜ pending |
| 09-04-02 — PropCheck_BucketEquivalentToNaive + PropCheck_DeterministicAcrossRuns | 04 | 4 | SCAN-02 | — | property | `go test -race -shuffle=on -count=1 -run PropCheck_Bucket ./scan/...` | ⬜ pending |
| 09-04-03 — BenchmarkScanCheck_BucketVsNaive_GroupSize sweep + DefaultScorer_10k | 04 | 4 | PERF-05 | — | benchmark | `go test -bench=BenchmarkScanCheck_ -benchmem -count=1 -run=^$ ./scan/...` | ⬜ pending |
| 09-04-04 — Empirical bucketThreshold validation (autonomous: false) | 04 | 4 | PERF-05, D-08 | — | manual-inspection | benchstat read + constant update if crossover differs from 50 | ⬜ pending |
| 09-04-05 — Final commit + 6-reviewer panel (incl. algorithm-licensing-reviewer) | 04 | 4 | SCAN-02, PERF-05 | — | gate | `make check` | ⬜ pending |
| 09-05-01 — suppress.go: isSuppressed + canonicalPair + suppressionCtx | 05 | 5 | SCAN-03 | — | unit | `go test -race -shuffle=on -count=1 -run TestSuppress ./scan/...` | ⬜ pending |
| 09-05-02 — Wire suppression into Check inner loop pre-emission | 05 | 5 | SCAN-03, SCAN-04 | — | unit | `go test -race -shuffle=on -count=1 -run TestCheck_Suppression ./scan/...` | ⬜ pending |
| 09-05-03 — Final commit + gate | 05 | 5 | SCAN-03 | — | gate | `make check` | ⬜ pending |
| 09-06-01 — sort.SliceStable on (Kind, NameA, NameB, GroupA, GroupB) | 06 | 6 | SCAN-05 | — | unit | `go test -race -shuffle=on -count=1 -run TestSort ./scan/...` | ⬜ pending |
| 09-06-02 — In-line completeness assertion + ErrInternalInvariantViolated panic | 06 | 6 | SCAN-05 | T-INVARIANT | unit | `go test -race -shuffle=on -count=1 -run TestCompletenessAssertion ./scan/...` | ⬜ pending |
| 09-06-03 — testdata/golden/scan-default.json staging + merge | 06 | 6 | DET-04 | — | golden | `go test -race -shuffle=on -count=1 -run TestGolden_Scan ./scan/...` | ⬜ pending |
| 09-06-04 — FuzzCheck + PropCheck_NoSelfWarnings + PropCheck_NoNaN/NoInf | 06 | 6 | DET-04 | — | property+fuzz | `go test -race -count=1 -run PropCheck_No ./scan/... && go test -fuzz=FuzzCheck -fuzztime=30s ./scan/...` | ⬜ pending |
| 09-06-05 — Final commit + determinism-reviewer gate | 06 | 6 | SCAN-05, DET-04 | — | gate | `make verify-determinism && make check` | ⬜ pending |
| 09-07-01 — scan_steps.go skeleton + InitScanSteps registration | 07 | 7 | TEST-05 | — | bdd-infra | `cd tests/bdd && go build ./... && go vet ./...` | ⬜ pending |
| 09-07-02 — scan.feature scenarios (D-03/D-04/D-06/D-07 + sort + bucket smoke) | 07 | 7 | SCAN-03, TEST-05 | — | bdd | `make test-bdd` | ⬜ pending |
| 09-07-03 — suppression.feature scenarios (D-05 rules + Phase 8.5 R2 closure) | 07 | 7 | SCAN-03, TEST-05 | — | bdd | `make test-bdd` | ⬜ pending |
| 09-07-04 — step definitions implementation | 07 | 7 | TEST-05 | — | bdd | `make test-bdd` | ⬜ pending |
| 09-07-05 — Final commit + bdd-scenario-reviewer gate | 07 | 7 | SCAN-03, TEST-05 | — | gate | `make test-bdd && make check` | ⬜ pending |
| 09-08-01 — examples/scan-demo/main.go runnable example | 08 | 8 | DX-02 | — | example | `go run ./examples/scan-demo` | ⬜ pending |
| 09-08-02 — examples/scan-demo/main_test.go meta-test | 08 | 8 | DX-02 | — | meta-test | `go test ./examples/scan-demo` | ⬜ pending |
| 09-08-03 — scan/example_test.go godoc Examples (one per public func) | 08 | 8 | DX-02 | — | example | `go test -run Example ./scan/...` | ⬜ pending |
| 09-08-04 — docs/scan.md user-facing reference | 08 | 8 | DX-04 | — | doc | `markdownlint-cli2 docs/scan.md` | ⬜ pending |
| 09-08-05 — CHANGELOG.md [Unreleased] Phase 9 entries | 08 | 8 | DX-04 | — | doc | `markdownlint-cli2 CHANGELOG.md` | ⬜ pending |
| 09-08-06 — llms.txt + llms-full.txt sync | 08 | 8 | DX-04 | — | doc | `make verify-llms-sync` | ⬜ pending |
| 09-08-07 — Full quality gate (make check + make bench + benchstat) | 08 | 8 | PERF-05, DET-04, TEST-05 | — | gate | `make check && make bench` | ⬜ pending |
| 09-08-08 — Final agent panel (licensing, api-ergonomics, determinism, user-guide, docs-writer, code, security, go-quality, commit-message) | 08 | 8 | DX-02, DX-04 | — | review-gate | manual reviewer sign-offs in 09-08-SUMMARY.md | ⬜ pending |
| 09-08-09 — Phase-boundary commit cluster | 08 | 8 | DX-02, DX-04 | — | gate | `git status` clean after commits | ⬜ pending |
| 09-08-10 — Phase-boundary push (autonomous: false per user-memory) | 08 | 8 | — | — | manual-push | `git push origin main` after CI green on prior commits | ⬜ pending |

*Status: ⬜ pending · ✅ green · ❌ red · ⚠️ flaky*

---

## Wave 0 Requirements

Wave 0 lands in Plan 09-01 — the foundation plan's Task 1 creates the scan/ package skeleton that all subsequent plans compile against.

- [ ] `scan/scan.go` — Item, Warning, Config struct skeletons + Check stub (returns ErrNilScorer until 09-03 body lands)
- [ ] `scan/kind.go` — Kind type + KindWithinGroup/KindAcrossGroups constants + String()
- [ ] `scan/errors.go` — ErrNilScorer, ErrInvalidItem, ErrInvalidConfig sentinels
- [ ] `scan/doc.go` — package godoc
- [ ] `scan/example_test.go` — ExampleKind_String (one godoc Example to verify go test runs Examples)
- [ ] `scan/scan_test.go` — type-assertion-only smoke tests so the package has > 0% coverage from day one
- [ ] `scorer.go` — Scorer.NormalisationOptions() accessor added (Open Question 1 resolution)
- [ ] `tests/bdd/steps/scan_steps.go` skeleton — landed in Plan 09-07 Task 1 (Wave 0 for the BDD sub-module)

Once Plan 09-01 commits, `wave_0_complete: true` can flip in this file's frontmatter and dependent plans 09-02..09-08 are unblocked.

---

## Manual-Only Verifications

| Behavior | Requirement | Why Manual | Test Instructions |
|----------|-------------|------------|-------------------|
| api-ergonomics-reviewer sign-off recording in foundation plan PR | D-01 + D-02 (SPEC OVERRIDE) | Reviewer sign-off is a human gate documented in PR description per Phase 7/8 precedent | Spawn `api-ergonomics-reviewer` agent on foundation plan PR; capture verdict in 09-01-SUMMARY.md |
| Empirical bucket-threshold crossover (D-08) | PERF-05, SCAN-02 | Crossover wall-clock differs by hardware — value must be inspected and committed by a human (autonomous: false on Task 09-04-04) | Run `BenchmarkScanCheck_BucketVsNaive_GroupSize`; inspect benchstat output; update `bucketThreshold` constant + in-source comment with date/wall-clock crossover |
| Phase-boundary push to origin | user-memory `feedback_push_cadence` | Push to shared remote is a human action per project policy (autonomous: false on Task 09-08-10) | `git push origin main` only after CI confirms green on the prior commits |

---

## Validation Sign-Off

- [x] All tasks have `<automated>` verify or Wave 0 dependencies
- [x] Sampling continuity: no 3 consecutive tasks without automated verify
- [x] Wave 0 covers all MISSING references
- [x] No watch-mode flags
- [x] Feedback latency < 30s for quick, < 3 min for full
- [x] `nyquist_compliant: true` set in frontmatter

**Approval:** approved 2026-05-19 (plan-checker gate iteration 1)
