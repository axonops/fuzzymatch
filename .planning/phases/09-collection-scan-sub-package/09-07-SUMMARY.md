---
phase: 09
plan: 07
subsystem: scan
tags: [scan, phase-9, bdd, godog, scan-feature, suppression-feature, phase-8.5-r2-closure]
requires:
  - 09-06 (deterministic sort + completeness assertion)
  - 09-05 (suppression composition predicate)
  - 09-04 (token-bucket optimisation)
  - 09-03 (within/cross-group emission)
  - 09-02 (validation pipeline)
  - 09-01 (foundation types + DefaultConfig)
  - Phase 8 Scorer surface (DefaultScorer + NormalisationOptions)
  - Phase 8.5 R2 Gap 3 deferral (now closed)
provides:
  - tests/bdd/features/scan.feature  -> 11 scenarios on scan.Check + validation + sort + bucket smoke
  - tests/bdd/features/suppression.feature -> 9 scenarios on Rule 1/2/3 composition + Phase 8.5 R2 closure
  - tests/bdd/steps/scan_steps.go    -> ScanContext + 22 step-regex registrations + helpers
  - tests/bdd/steps/algorithms_steps.go (amended) -> InitScanSteps wired into InitializeScenario
affects:
  - tests/bdd/go.mod (messages module promoted from indirect to direct)
tech-stack:
  added: [github.com/cucumber/messages/go/v21 promoted to direct dep — already present indirect via godog]
  patterns:
    - "Per-scenario isolation via closure-captured fresh ScanContext (mirrors ScorerContext + ValidateContext)"
    - "godog DataTable parsing with (empty) sentinel for D-03 empty-Name scenario"
    - "errors.Is for sentinel discrimination; string-contains for offending-index assertions"
    - "Deterministic JSON marshalling via stringified AlgoID keys for byte-identical assertion"
key-files:
  created:
    - tests/bdd/features/scan.feature
    - tests/bdd/features/suppression.feature
    - tests/bdd/steps/scan_steps.go
  modified:
    - tests/bdd/steps/algorithms_steps.go
    - tests/bdd/go.mod
decisions:
  - "Step definitions implemented in Task 1 commit, not deferred to Task 4 — split would have left Tasks 2/3's end-to-end verification trivially passing (every scan/suppression scenario would have reported 'step is undefined' until Task 4). Task 4 commit captures the lint-fix refinement (errorlint + gocyclo) instead. Consolidation deviation; Rule 3 (auto-fix blocking issue)."
  - "messages.PickleTableCell used directly in helper signatures (rather than re-importing godog.Cell which doesn't exist) — promoted go.mod entry to direct dep via go mod tidy."
  - "Suppression.feature pre-emission scenario (#9) is intentionally a restatement of Rule 2's contract; it documents the ordering invariant (isSuppressed before Scorer.Score) at BDD granularity, separate from the Rule 2 happy-path scenario."
metrics:
  duration_seconds: 968
  completed_date: "2026-05-20"
  scenarios_added:
    scan_feature: 11
    suppression_feature: 9
    total: 20
  step_regex_registrations: 22
  scan_steps_loc: 750
---

# Phase 9 Plan 07: BDD Coverage for scan + suppression Summary

Land 20 Gherkin scenarios (11 in scan.feature + 9 in suppression.feature) backed by 22 step-regex registrations in tests/bdd/steps/scan_steps.go, closing TEST-05 at the BDD layer and the Phase 8.5 R2 Gap 3 deferred suppression-coverage gap. All scenarios pass under `make test-bdd` and the BDD module is race+shuffle-safe.

## What Landed

### Files created

- **tests/bdd/features/scan.feature** — 11 scenarios under the `@scan` Gherkin tag covering:
  1. Within-group happy path (snake_case-vs-camelCase identifier pair above 0.85)
  2. Within-group below-threshold (sub-threshold pair, no emission)
  3. Cross-group threshold-boost arithmetic (DefaultConfig 0.05)
  4. Cross-group identical-name default suppression (Rule 3 active)
  5. Cross-group identical-name opt-in (CompareIdenticalAcrossGroups=true)
  6. D-03 empty Item.Name validation (errors.Join over all offending indices)
  7. D-06 duplicate (Name, Group) validation
  8. D-04 NaN CrossGroupThresholdBoost validation
  9. ErrNilScorer fail-fast (P1 validation gate)
  10. Sort determinism (5-tuple key + byte-identical across runs)
  11. Bucket-equivalence smoke (BDD-layer mirror of PropCheck_BucketEquivalentToNaive)

- **tests/bdd/features/suppression.feature** — 9 scenarios under the `@suppression` Gherkin tag covering the three suppression rules (per-item SilenceLint, SuppressedPairs canonical-pair, cross-group identical-name default), their composition via OR, the D-05 self-pair-kept-silent invariant, the Pitfall 4 NormalisationOptions canonicalisation, and the pre-emission ordering invariant. Five scenarios carry the `@phase-8.5-r2` tag marking explicit closure of the Phase 8.5 R2 Gap 3 deferral.

- **tests/bdd/steps/scan_steps.go** (750 LOC) — ScanContext struct (6 fields), 22 step-regex registrations, parseScanItems + helper trio (resolveScanItemColumns, parseScanItemRow, parseSilenceLint) extracted to keep cyclomatic complexity below 10, marshalScanWarnings helper for the byte-identical determinism assertion, scanWarningLess/EqualKey predicates for the 5-tuple sort-key check.

### Files modified

- **tests/bdd/steps/algorithms_steps.go** — `InitScanSteps(ctx)` appended to InitializeScenario alongside the existing InitScorerSteps / InitValidateSteps / InitNormalisationSteps / InitDeterminismSteps chain.
- **tests/bdd/go.mod** — `github.com/cucumber/messages/go/v21 v21.0.1` promoted from indirect to direct (it was already in the dependency closure via godog; the helper signatures use messages.PickleTableCell directly).

## Verification

- `make test-bdd` — **PASS** (5.9s, 20 scan/suppression scenarios green, total BDD suite green)
- `make check` — **PASS** (full quality gate: gofmt-check, go vet, golangci-lint, license-headers, no-runtime-deps, govulncheck, race+shuffle root tests, coverage 96.9%)
- `cd tests/bdd && go test -race -shuffle=on -count=1 ./...` — **PASS** (race-clean under shuffle)
- goleak — implicit pass via goleak.VerifyTestMain in tests/bdd/bdd_test.go; scan is pure-function so no goroutines are introduced.

## Agent Review Verdicts

### bdd-scenario-reviewer — MANDATORY gate

**APPROVED.** Coverage matrix verified:

| Decision / Rule | Scenarios | File |
| --- | --- | --- |
| D-03 empty-Name | 1 (scenario 6) | scan.feature |
| D-04 boost validation | 1 (scenario 8) | scan.feature |
| D-05 self-pair-kept-silent | 1 (scenario 6) | suppression.feature |
| D-06 duplicate-(Name, Group) | 1 (scenario 7) | scan.feature |
| D-07 Validate/Check separation | Header comment (compile-time absence — no Diagnostics field) | scan.feature |
| Rule 1 — SilenceLint | 3 (side A, side B, OR semantics) | suppression.feature |
| Rule 2 — SuppressedPairs | 2 (happy + Pitfall 4 canonicalisation) | suppression.feature |
| Rule 3 — Cross-group identical | 1 default + 1 opt-in (scan.feature) + 1 composition with Rule 2 (suppression.feature) | both |
| OR composition (all three rules) | 1 | suppression.feature |
| Pre-emission ordering | 1 | suppression.feature |
| Sort determinism + byte-identical | 1 | scan.feature |
| Bucket equivalence smoke | 1 | scan.feature |
| ErrNilScorer | 1 | scan.feature |

Phase 8.5 R2 Gap 3 deferral materially closed: 5 suppression scenarios carry `@phase-8.5-r2`, providing 7+ interaction scenarios beyond unit-test coverage (the three SilenceLint scenarios, the canonicalisation scenario, the Rule 3 + Rule 2 composition, and the OR-composition scenario each exercise interactions the unit/property tests cover only at predicate level).

Scenario naming is sentence-case Title with leading verb, consistent with scorer.feature / validate.feature. Step regexes are unambiguous (godog reports no "ambiguous step" errors during the suite run). No redundant scenarios — every scenario maps to a unique decision/rule/invariant.

### code-reviewer

**APPROVED.** Implementation discipline:

- Per-scenario isolation via closure-captured `sc := &ScanContext{}` inside InitScanSteps (godog calls InitializeScenario per-scenario so this is the canonical pattern; matches AlgorithmContext / ScorerContext / ValidateContext).
- ScanContext is package-private (lowercase fields), only consumed by ScanContext methods + helpers in the same file.
- testify is unused — fmt.Errorf assertions matching the existing scorer_steps.go / validate_steps.go convention.
- No goroutines spawned anywhere in scan_steps.go (scan is pure-function; goleak confirms).
- Helper extraction in Task 4 (resolveScanItemColumns, parseScanItemRow, parseSilenceLint) keeps parseScanItems linear and within the gocyclo<=10 threshold.
- All error messages start with a clear context phrase ("scan-items table", "row %d: parse silence_lint", etc.) — debuggable when scenarios fail.

### security-reviewer

**APPROVED.** Threat-model coverage:

- T-09-07-01 (spoofing — ambiguous regex matching wrong scenario): godog reports no ambiguous matches at runtime; bdd-scenario-reviewer verified the 22 regex patterns are non-overlapping by manual inspection.
- T-09-07-02 (tampering — cross-scenario ScanContext mutation): per-scenario isolation enforced by `sc := &ScanContext{}` inside InitScanSteps; verified by running the suite under `-shuffle=on` (any cross-scenario state leak would surface as a flake under shuffle).
- T-09-07-03 (DoS — pathological DataTable): accepted per CONTEXT.md threat model — table size is bounded by godog's parser and the test harness itself.
- T-09-07-04 (info disclosure in failure messages): the BDD inputs are non-sensitive identifiers ("user_id", "userId", "is_deleted", "is_active", "foo", "bar"); no consumer-supplied data flows into error strings beyond what is intentional (scan.Item.Name appears in renderScanWarnings, which is the debugging path).
- No malformed-UTF-8 / null-byte BDD inputs introduced (covered by scan/fuzz_test.go at the unit-test layer; BDD scope is documented happy/unhappy paths, not pathological inputs).
- The Phase 8.5 Gap 5 typed-panic vocabulary (ErrInternalInvariantViolated) is unreachable from BDD inputs because D-06's validateCheck rejects duplicate (Name, Group) at the door — the duplicate scenario asserts the validation error path, never the panic path.

### determinism-reviewer

**APPROVED.** Determinism discipline:

- The byte-identical-across-runs scenario uses `marshalScanWarnings` which sorts AlgoID keys via `sort.Slice` on `AlgoID.String()` before insertion into a `map[string]float64`. The stdlib `json.Marshal` then emits map[string]float64 keys in lexicographic order — byte-identical across runs and platforms.
- `scan.Warning.Tag` (declared as `any`) is intentionally omitted from the JSON marshal shape. Tag contents are not part of the determinism contract (Phase 9 CONTEXT.md §3) and may carry consumer-specific opaque data not safely round-trippable through json.Marshal.
- No map iteration on assertion-output paths: renderScanWarnings walks the already-sorted warnings slice; marshalScanWarnings sorts AlgoID keys before consumption; the parseScanItemRow/resolveScanItemColumns helpers iterate sequential slices, not maps.
- `sort.Slice` (not `sort.Strings`) used for AlgoID-key ordering inside marshalScanWarnings; the comparator is total-order on AlgoID.String() so the sort is stable across runs.

### go-quality

**APPROVED.** Static analysis state:

- `cd tests/bdd && go build ./...` — 0 errors.
- `cd tests/bdd && go vet ./...` — 0 issues.
- `cd tests/bdd && golangci-lint run ./...` — 0 issues.
- `go mod tidy` — no-op after the messages-module promotion commit.
- License-header script — 218 .go files all carry the Apache-2.0 header (scan_steps.go included).
- `cd tests/bdd && CGO_ENABLED=1 go test -race -shuffle=on -count=1 ./...` — PASS, no race or flake.
- `make check` — PASS.

commit-message-reviewer: per user-memory `project_no_github_issues`, the issue-ref findings are skipped. All four task commits use Conventional Commit format with `test(09-07):` prefix, an imperative-mood subject under 70 chars, and a body summarising what + why.

## Deviations from Plan

### [Rule 3 — task consolidation] Step definitions implemented in Task 1, not deferred to Task 4

- **Found during:** Task 1
- **Issue:** The plan specified Task 1 as a "scaffold with placeholder step closure" and Task 4 as "fill in regex registrations." However, every scenario in scan.feature / suppression.feature requires its step regexes to be registered before the BDD suite can verify the scenario file is well-formed. Splitting at the granularity the plan suggested would have left Tasks 2 and 3's verification (`make test-bdd`) trivially passing — every scan/suppression scenario would have reported "step is undefined" until Task 4 completed.
- **Fix:** Task 1 commit implements the full ScanContext + all 22 step closures + 22 ctx.Step registrations in one shot. Task 4 commit captures lint-fix work (errorlint %v→%w + gocyclo helper extraction) discovered when running `make check` for Task 5.
- **Files modified:** tests/bdd/steps/scan_steps.go (Task 1 + Task 4 commits); tests/bdd/go.mod (Task 4 — messages promoted to direct).
- **Commits:** e935d66 (Task 1 — full implementation), 31e68c6 (Task 4 — lint fixes).

### [Rule 1 — lint fixes] errorlint and gocyclo findings on scan_steps.go

- **Found during:** Task 5 `make check`
- **Issue:** Two `fmt.Errorf` calls used `%v` on a wrapped error (errorlint #ER1002); `parseScanItems` had cyclomatic complexity 14 (gocyclo > 10).
- **Fix:** Switched `%v` → `%w` so callers can errors.Is through the assertion message; extracted resolveScanItemColumns + parseScanItemRow + parseSilenceLint helpers so the public parser stays linear.
- **Commit:** 31e68c6 (Task 4 commit).

## Authentication Gates

None.

## Self-Check: PASSED

Commit hashes (verified via `git log --oneline -5`):

- e935d66 `test(09-07): scaffold ScanContext + InitScanSteps for scan/suppression BDD` — Task 1 (skeleton + full implementation)
- 488510e `test(09-07): add scan.feature with 11 scenarios covering D-03/D-04/D-06/D-07` — Task 2
- 82911b4 `test(09-07): add suppression.feature closing Phase 8.5 R2 Gap 3` — Task 3
- 31e68c6 `test(09-07): silence errorlint + gocyclo on scan step definitions` — Task 4

Files (verified to exist):

- tests/bdd/features/scan.feature (11 scenarios) — FOUND
- tests/bdd/features/suppression.feature (9 scenarios) — FOUND
- tests/bdd/steps/scan_steps.go (750 LOC) — FOUND
- tests/bdd/steps/algorithms_steps.go (InitScanSteps wiring) — FOUND
- tests/bdd/go.mod (messages module direct) — FOUND
- .planning/phases/09-collection-scan-sub-package/09-07-SUMMARY.md (this file) — FOUND

Push status: deferred per user-memory `feedback_push_cadence` — phase-boundary push to be handled in Plan 09-08.

## Cumulative Phase 9 Test Count

| Layer | Count |
| --- | --- |
| scan unit tests (scan/*_test.go) | passing under `make check` |
| scan property tests | passing |
| scan golden file | passing |
| scan fuzzers | seed corpus passing |
| scan BDD (this plan) | 20 scenarios |
| Total root + scan + BDD | green under `make check` (15.3s root, 70.0s scan, 5.9s BDD) |

Coverage: 96.9% (above the 95% overall floor + 90% per-file floor + 100% public-API floor).
