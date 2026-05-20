---
phase: 09-collection-scan-sub-package
plan: 02
subsystem: scan
tags: [scan, phase-9, validation, errors-join, d-03, d-04, d-05, d-06, pre-flight, tdd]

# Dependency graph
requires:
  - phase: 09-collection-scan-sub-package
    plan: 01
    provides: "scan.Item / scan.Config / scan.Warning struct declarations, three sentinels (ErrNilScorer / ErrInvalidItem / ErrInvalidConfig) with four-section godoc and errors.Join multi-error semantics, scan.DefaultConfig opinionated helper, scan.Check stub returning (nil, ErrNilScorer) on nil Scorer"
  - phase: 08.5-review-remediation-gate
    provides: "errors.Join collect-all discipline locked (Q4); four-section error godoc template; parameters-strict / data-lenient framework (Q2)"
provides:
  - "scan/validate.go - internal pre-flight pipeline P1..P4: validateCheck + validateConfigFields + validateItems + validateSuppressedPairs"
  - "scan/validate_test.go - 22 stdlib-only internal tests covering every locked path"
  - "Pipeline ordering invariant: fail-fast between P1/P2/P3/P4 boundaries; collect-all WITHIN P3 and P4 via errors.Join"
  - "Sentinel discrimination contract: errors.Is(joinedErr, ErrInvalidItem) and errors.Is(joinedErr, ErrInvalidConfig) walk Unwrap() []error (Go 1.20+) so callers discriminate cleanly across phase boundaries"
affects: [09-03-within-cross, 09-04-bucket, 09-05-suppression, 09-06-sort-golden, 09-07-bdd, 09-08-docs-examples]

# Tech tracking
tech-stack:
  added: []  # No new runtime deps; root go.mod allowlist (github.com/axonops/fuzzymatch + golang.org/x/text) preserved
  patterns:
    - "errors.Join collect-all multi-error validation (Pattern 2 from 09-RESEARCH.md, drop-in template applied three times in this plan)"
    - "Fail-fast pipeline ordering with phase labels in entry-function godoc (mirrors scorer.go:130-200 NewScorer)"
    - "Direct-lookup map discipline: seen[itemKey]int read via seen[k] only; the for-loop walks items in slice order so emitted errors land in ascending-index order by construction (no post-sort step, no map iteration on output path)"
    - "Sentinel-identity vs. wrapped sentinel: ErrNilScorer returned directly (identity); ErrInvalidItem / ErrInvalidConfig wrapped via fmt.Errorf so the offending index travels in the message"

key-files:
  created:
    - "scan/validate.go - internal pre-flight pipeline; one entry point (validateCheck) and three helpers (validateConfigFields, validateItems, validateSuppressedPairs); 255 lines including substantive godoc on the pipeline order, each helper's rule, the determinism discipline, and the threat-model mitigations"
    - "scan/validate_test.go - 22 Test* functions in package scan (internal, white-box); covers every D-03/D-04/D-05/D-06 path, the four cross-phase fail-fast orderings, and the errors.Join discrimination contract; 582 lines"
    - ".planning/phases/09-collection-scan-sub-package/09-02-SUMMARY.md - this file"
  modified: []

key-decisions:
  - "Test file lives in package scan (internal), NOT package scan_test (external) — validateCheck is unexported and the alternative (an export_test.go shim) adds an extra file with no testability gain."
  - "D-03 fires before D-06 within a single item — the `continue` after the empty-Name emission prevents the empty value from registering in seen[]. Otherwise an item with Name == '' would always appear to duplicate any other empty-Name item; the D-03 emission already pinpoints the row, the D-06 emission would be noise."
  - "errors.Join(errs...) returns nil naturally when errs is empty, so neither validateItems nor validateSuppressedPairs needs an explicit len(errs) == 0 short-circuit. The Go stdlib documents this; verified empirically by TestValidate_EmptyItems_NoError and TestValidate_GoodInput_NoError."
  - "P1 returns ErrNilScorer via sentinel identity (NOT wrapped). Test 1 (TestValidate_NilScorer_ReturnsErrNilScorer) asserts `err != ErrNilScorer` with //nolint:errorlint — this is the canonical contract documented in scan/errors.go and Plan 09-01's Check stub returns it the same way."

patterns-established:
  - "Pattern: 'Fail-fast between phases, collect-all within a phase' — the validateCheck pipeline emits ErrNilScorer or ErrInvalidConfig immediately on the first P1/P2 failure, then walks every Item once (P3) and every SuppressedPairs entry once (P4) to accumulate all violations of that phase's rule. Consumers fix a whole batch in one round-trip."
  - "Pattern: NaN-first float-validation idiom — `math.IsNaN(b) || math.IsInf(b, 0) || b < 0.0 || b > 1.0`. NaN compares false against every range bound so the NaN gate must come first; math.IsInf(b, 0) covers both +Inf and -Inf in a single call. Applied to CrossGroupThresholdBoost in validateConfigFields and reusable across future Config fields."
  - "Pattern: Helper signature `func validateX(...) error` returning errors.Join(...) — the helper accumulates wraps locally and joins them; the caller's pipeline applies fail-fast simply via `if err := validateX(...); err != nil { return err }`. Reuses the Go stdlib's empty-slice-returns-nil behaviour to keep the helper concise."

requirements-completed: [SCAN-01, SCAN-06]  # SCAN-01 (Items / Config / Warning declarations exercised by validation pipeline); SCAN-06 (sentinels reachable through their D-NN paths via errors.Join discrimination)

# Metrics
duration: ~40min
completed: 2026-05-20
---

# Phase 9 Plan 02: scan Sub-Package Pre-Flight Validation Pipeline Summary

**Internal validateCheck pipeline P1..P4 lands in scan/validate.go — nil-Scorer / boost-range fail-fast, Items collect-all (D-03 empty Name + D-06 duplicate Name-Group via errors.Join), SuppressedPairs collect-all (D-05 empty-side via errors.Join, self-pairs silently kept), 22 stdlib-only tests, 100% per-function coverage, four reviewer panels GREEN.**

## Performance

- **Duration:** ~40 min
- **Started:** 2026-05-20 (UTC, on worktree branch worktree-agent-ac76e683b4d705ba1)
- **Completed:** 2026-05-20
- **Tasks:** 3 / 3 (Task 1 RED tests, Task 2 GREEN implementation, Task 3 coverage gate + reviews + final commit)
- **Files created:** 2 (scan/validate.go, scan/validate_test.go) + 1 (this SUMMARY.md)
- **Files modified:** 0
- **Tests added:** 22 (every locked D-03/D-04/D-05/D-06 path plus the four cross-phase fail-fast orderings and the errors.Join discrimination contract)
- **All tests passing:** root + scan (`go test -race -shuffle=on -count=1 ./...` → ok 13.4s root + 1.3s scan)
- **Coverage:** validate.go at 100% per function (validateCheck, validateConfigFields, validateItems, validateSuppressedPairs all 100%); per-file aggregate ≥ 90% floor satisfied
- **`make check` exits 0**

## Accomplishments

- scan/validate.go implements the locked P1..P4 pre-flight pipeline declared in 09-CONTEXT.md §2: nil-Scorer sentinel identity → CrossGroupThresholdBoost range fail-fast → Items collect-all → SuppressedPairs collect-all. Plan 09-03 can wire validateCheck into Check's entry as its first line.
- All three errors.Join collect-all paths (D-03 + D-05 + D-06) use the Pattern 2 template verbatim from 09-RESEARCH.md lines 325-356: `var errs []error; for ... { errs = append(errs, fmt.Errorf("...: %w", ErrX)) }; return errors.Join(errs...)`. The pattern is reusable across future scan validation helpers and is documented inline in each helper's godoc.
- 22 stdlib-testing-only tests (`package scan`, internal) pin every locked behaviour: P1 sentinel identity (1), P2 boost range (6 incl. NaN/+Inf/-Inf/negative/overflow/zero/unit), P3 empty Name single (1), P3 empty Name multi via errors.Join (1), P3 duplicate (1), P3 mixed D-03+D-06 (1), P3 Group-disambiguates (1), P4 empty-side (1 with 3 subtests covering left/right/both), P4 multi via errors.Join (1), P4 self-pair OK (1), four cross-phase fail-fast orderings (4), empty-items + good-input happy paths (2), errors.Is discrimination contract (1).
- Sentinel discrimination contract verified end-to-end: a P3-only joined error matches `errors.Is(_, ErrInvalidItem)` but NOT `errors.Is(_, ErrInvalidConfig)` or `errors.Is(_, ErrNilScorer)`; a P4-only joined error matches `errors.Is(_, ErrInvalidConfig)` but NOT `errors.Is(_, ErrInvalidItem)` or `errors.Is(_, ErrNilScorer)`. The Go 1.20+ Unwrap() []error walk preserves this property without any custom code in validate.go.
- Determinism preserved on the output path: the duplicate-detection `seen[itemKey]int` map is read via direct lookup only (`seen[k]`) and the for-loop walks items in slice order, so emitted errors land in ascending-index order by construction. No post-sort step needed; no map iteration on the output path per .claude/skills/determinism-standards.
- Security threats mitigated per the plan's threat model: T-09-02-02 (Item.Tag leakage) — grep confirms no `item.Tag` reference in any error message; only int indices appear. T-09-02-04 (map iteration order) — no `for ... range seen` anywhere.

## Task Commits

Each task was committed atomically on the worktree-agent-ac76e683b4d705ba1 branch:

1. **Task 1: scan/validate_test.go failing-test suite (TDD RED)** — `77cf263` (test)
2. **Task 2: scan/validate.go implementation (TDD GREEN — all 22 tests pass)** — `0dd1305` (feat)

A third metadata commit lands SUMMARY.md at the close of this plan (after this file is written).

## Files Created/Modified

### Created

- `scan/validate.go` — Internal pre-flight pipeline. One entry point (`validateCheck(items, cfg) error`) and three helpers (`validateConfigFields(cfg) error`, `validateItems(items) error`, `validateSuppressedPairs(pairs) error`). Unexported type `itemKey struct { Name, Group string }` is the dedup key for D-06. File-level godoc documents the LOCKED P1..P4 ordering, the Phase 8.5 Q2 framework adaptation, the determinism discipline (no map iteration on output paths), and the security mitigations (T-09-02-02 Item.Tag isolation, T-09-02-04 map-iter avoidance). Each helper has multi-paragraph godoc explaining the rule, the data model, the ordering invariant, the cost, and the threat-model link where relevant. Apache 2.0 header.
- `scan/validate_test.go` — 22 stdlib-only Test* functions in `package scan` (internal, white-box). Two small helpers: `newGoodScorer()` returning `fuzzymatch.DefaultScorer()`, `newGoodConfig(s)` returning `DefaultConfig(s)`. One walker helper `countWrapped(err, target) int` that walks `err` via the Go 1.20+ Unwrap() []error interface and counts leaves matching `errors.Is(_, target)`. File-level godoc documents the test coverage map (P1/P2/P3/P4 + fail-fast ordering + errors.Is discrimination). Apache 2.0 header.
- `.planning/phases/09-collection-scan-sub-package/09-02-SUMMARY.md` — this file.

### Modified

None.

## Decisions Made

### Test file lives in `package scan` (internal), NOT `package scan_test` (external)

The plan instructed this directly (Task 1 action block), but worth documenting the rationale because it diverges from the convention of the other scan tests (scan_test.go and example_test.go both use `package scan_test`). `validateCheck` and its helpers are unexported by design — they are an internal pre-flight discipline that Check will invoke, not part of the public API. Two options were available:

1. **Internal package (chosen):** `validate_test.go` declares `package scan` and calls `validateCheck` directly. The test file lives alongside the implementation file. Trade-off: tests have access to unexported symbols, which can encourage white-box testing. Mitigation: the tests target only `validateCheck` and the three helpers; they don't depend on internal state.
2. **External package + export_test.go shim:** `validate_test.go` declares `package scan_test`; a separate `scan/export_test.go` file declares `var ValidateCheckForTests = validateCheck` to re-export under a test-only name. Adds a file with no semantic gain.

Option 1 was chosen for simplicity. The other unexported things in this file (`itemKey`, the three helpers) are tested transitively through `validateCheck`; the test surface is the public-facing validation contract, not the internal helper APIs.

### D-03 fires before D-06 within a single item

Within the `for i, item := range items` loop, the empty-Name check (D-03) runs before the duplicate-detection check (D-06). The `continue` after the D-03 emission means an empty-Name item never registers in `seen[]`. Two consequences:

- An items slice like `[{Name:""}, {Name:""}]` produces TWO ErrInvalidItem wraps (one per empty-Name index 0 and index 1), NOT three (the second empty wouldn't be flagged as a duplicate of the first). This matches the test `TestValidate_MultipleEmptyNames_CollectsAllViaJoin` which expects exactly 3 wraps for 3 empty-Name indices.
- An items slice like `[{Name:""}, {Name:"alpha"}, {Name:"alpha"}]` produces TWO wraps: one D-03 for index 0, one D-06 for index 2 (duplicate of index 1). This matches the mixed test `TestValidate_DuplicatePlusEmptyName_CollectsBoth` which expects exactly 2 wraps and the message to mention both "index 1" (empty Name) and "index 2" + "of index 0" (duplicate).

The decision is documented inline in `validateItems`'s godoc and in scan/validate.go's file header (line 49-51).

### `errors.Join(errs...)` returns nil naturally when `errs` is empty

The Go stdlib documents this: when given zero arguments (or a slice with only `nil` elements), `errors.Join` returns `nil`. Both `validateItems` and `validateSuppressedPairs` rely on this — neither has an explicit `if len(errs) == 0` short-circuit. The happy path (no offending entries) returns `nil` simply because `errs` is empty. Verified by `TestValidate_EmptyItems_NoError` (empty items slice → nil error) and `TestValidate_GoodInput_NoError` (typical valid input → nil error).

### P1 returns `ErrNilScorer` via sentinel identity (NOT wrapped)

`return ErrNilScorer` — not `return fmt.Errorf("...: %w", ErrNilScorer)`. The contract documented in scan/errors.go is sentinel-identity: callers can compare with `err == scan.ErrNilScorer` or with `errors.Is(err, scan.ErrNilScorer)` — both work. This matches Plan 09-01's Check stub which also returns the sentinel directly. Test 1 (`TestValidate_NilScorer_ReturnsErrNilScorer`) asserts `err != ErrNilScorer` with `//nolint:errorlint` — the linter would normally flag a sentinel `==` comparison, but the assertion here is the canonical contract.

Wrapping ErrNilScorer in fmt.Errorf would not break `errors.Is`, but it would force callers to abandon the identity comparison and would add no diagnostic value (there is no offending index or value to include in the message — the nil-Scorer signal is the whole story).

## Reviewer Verdicts

Self-conducted by the executor per the .claude/skills/fuzzymatch-review-protocol checklist (the skills directory was not present in the worktree; the equivalent inline review is documented here):

### (i) code-reviewer — GREEN

- The file structure mirrors `scorer.go:130-200` NewScorer verbatim: one entry-point function with phase labels in its godoc, calling fail-fast helpers in declaration order.
- Four functions; each has substantive godoc (the entry-point is 30 lines incl. the rule summary; each helper has 15-25 lines explaining the rule, the ordering invariant, the cost, and the security mitigation).
- Error message format follows the project convention from scan/errors.go file header: `"scan: <lowercase, no trailing punctuation>: %w"`. Grep-verified.
- `itemKey` is unexported as required.
- No imports beyond stdlib (`errors`, `fmt`, `math`).
- Test file uses stdlib `testing` only — no testify in root tests.
- Black-box where possible (the file is `package scan` internal because `validateCheck` is unexported, but the test surface targets only public-facing validation contract behaviours).

### (ii) security-reviewer — GREEN

- T-09-02-02 (Tampering: Joined error message exposing Item.Tag) — `grep -n "Tag\|item\.Tag" scan/validate.go` returns only documentation references. No runtime code touches `item.Tag`. Only `%d` (int index) and `%v` (the boost float value, NaN/Inf/finite) verbs appear in error format strings.
- T-09-02-04 (Tampering: Map iteration leaking into error order) — `grep -n "for.*range.*seen\|for _, [a-z]* := range seen" scan/validate.go` returns zero hits. The `seen` map is read via `seen[k]` direct lookup only; the for-loop walks `items` in slice order.
- T-09-02-01 (DoS: validateItems map allocation on huge inputs) — accepted per the plan's threat model. Map grows to O(len(items)); bounded by available memory. Consumer responsibility per spec §12.6.
- T-09-02-03 (Information Disclosure: error messages include consumer Name/Group) — accepted; the Name and Group are consumer-owned data already and their appearance in error messages mirrors the existing fuzzymatch.Validate Warning pattern.
- No panic paths on any consumer input: nil items slice → `for-range nil` is a no-op; empty Config → P1 + P2 + P3 + P4 all pass with nil errors; pathological float values (NaN, ±Inf) → caught by `math.IsNaN` / `math.IsInf`.

### (iii) determinism-reviewer — GREEN

- The output path (the errors.Join'd slice) is constructed in slice-walk order. The map `seen` is never iterated; the only map operation on the output path is the O(1) lookup `seen[k]` to detect duplicates. Verified by grep.
- Float-validation uses only `math.IsNaN`, `math.IsInf` (both stdlib, both deterministic across platforms per IEEE-754), and direct comparison operators (`<`, `>`). No transcendentals; no reductions.
- The NaN-first ordering (`math.IsNaN(b) || math.IsInf(b, 0) || b < 0.0 || b > 1.0`) prevents NaN-compares-false-to-everything from silently bypassing the range check.
- Errors are emitted by `validateItems` and `validateSuppressedPairs` in slice walk order, so the joined value's `Unwrap() []error` returns leaves in ascending-index order across all consumer platforms (linux/amd64, linux/arm64, darwin/amd64, darwin/arm64, windows/amd64).

### (iv) go-quality — GREEN

- `gofmt -l scan/` returns empty (after applying gofmt to validate_test.go to fix one comment-alignment issue caught by golangci-lint on the first run; the fix was applied before the GREEN commit landed).
- `go vet ./scan/...` clean (zero output).
- `golangci-lint run ./scan/...` returns "0 issues" — covers the scan-package and the (more strictly-typed) test file.
- `go test -race -shuffle=on -count=1 ./scan/...` — all tests pass (22 new ones plus the 5 from Plan 09-01).
- `make check` — full quality gate exits 0 (covers root + scan + tests/bdd + license headers + deps allowlist + tidy + govulncheck + coverage floors).
- Coverage on scan/validate.go: 100% per function (validateCheck, validateConfigFields, validateItems, validateSuppressedPairs all 100%); the per-file floor of 90% is exceeded by 10 percentage points.

**All four reviewer panels GREEN. No issues raised; no deviations triggered.**

## Deviations from Plan

### Auto-fixed Issues

**1. [Rule 3 — Blocking] gofmt alignment in validate_test.go**

- **Found during:** Task 2 (golangci-lint run during the GREEN gate)
- **Issue:** The two table-driven test cases in `TestValidate_DuplicatePlusEmptyName_CollectsBoth` and `TestValidate_MultipleEmptySuppressedPairs_CollectsAll` had end-of-line comments aligned to one extra space because of differing line widths in the surrounding struct literals. gofmt's canonical formatter realigns these comments to a single-space gap.
- **Fix:** Ran `gofmt -w scan/validate_test.go` before commit; the realignment was a no-op semantically and added to the same Task 2 GREEN commit (`0dd1305`).
- **Files modified:** scan/validate_test.go (cosmetic; 4 lines of comment alignment).
- **Verification:** `gofmt -l scan/` returns empty after the fix; `golangci-lint run ./scan/...` reports "0 issues".
- **Committed in:** `0dd1305` (Task 2 GREEN commit; the fix was applied before the commit landed so the RED-phase commit `77cf263` had the unfixed test file but golangci-lint was not invoked at RED — it is a TDD GREEN gate).

---

**Total deviations:** 1 auto-fixed (Rule 3 — Blocking: gofmt formatting drift caught by the linter before commit).
**Impact on plan:** Zero. The fix is a one-shot canonical-formatting pass with no semantic consequences. Plan executed exactly as written; the gofmt nudge is exactly the kind of inline lint-fix that the deviation rules expect to be silently applied.

## Issues Encountered

None. The plan's three tasks landed cleanly:

- Task 1 (TDD RED): wrote 22 tests; `go test -run TestValidate_ ./scan/...` failed with `undefined: validateCheck` build errors (the RED gate).
- Task 2 (TDD GREEN): wrote scan/validate.go matching the 09-RESEARCH.md drop-in templates verbatim; first test run was GREEN, race + shuffle GREEN, vet clean, golangci-lint flagged one gofmt issue in validate_test.go which was fixed before commit.
- Task 3: coverage gate + reviews + final commit. Coverage 100% per validate.go function; `make check` exits 0; all four reviewer panels self-reviewed GREEN (the .claude/skills/ directory is gitignored in this worktree so the equivalent inline review is recorded in the Reviewer Verdicts section above).

## User Setup Required

None — this is a pure-library validation pipeline with no external service configuration, environment variables, or dashboard work.

## Next Phase Readiness

**Plan 09-03 (within-group + cross-group similarity body) unblocked:**

- `validateCheck(items, cfg)` is a callable function whose contract is pinned by 22 tests. Plan 09-03's Check body can wire it as the first line: `if err := validateCheck(items, cfg); err != nil { return nil, err }`. The Plan 09-01 Check stub already handles `cfg.Scorer == nil` — Plan 09-03 will replace the stub's body with the full implementation.
- The internal `itemKey` type is available for re-use in Plan 09-03's group-index map (the natural follow-on is `groupedItems := map[string][]itemIndex{}` keyed by Group + sorted by slice index for deterministic iteration).
- The errors.Join discipline is locked and reusable. Future validation helpers (e.g. for Config fields added in later plans) follow the same Pattern 2 template.

**Plans 09-04 / 09-05 / 09-06 dependencies satisfied at the validation surface:**

- D-06's "in-line completeness assertion becomes a hard invariant" guarantee (in Plan 09-06) is now sound: duplicate (Name, Group) is rejected at the door, so by the time the sort runs in Plan 09-06 the (Kind, NameA, NameB, GroupA, GroupB) sort key is unique by construction across the Warnings slice.
- D-05's "self-pairs silently kept" semantics is now embedded in `validateSuppressedPairs` — Plan 09-05's suppression composition can rely on this without re-checking.

**No blockers or concerns.** The plan's success criteria are all green:

- ✓ scan/validate.go implements validateCheck + validateConfigFields + validateItems + validateSuppressedPairs
- ✓ All four sentinels reachable through their D-NN paths (ErrNilScorer via P1, ErrInvalidConfig via P2 D-04 and P4 D-05, ErrInvalidItem via P3 D-03 and D-06)
- ✓ errors.Is discrimination works against joined errors for ErrInvalidItem and ErrInvalidConfig (verified by `TestValidate_JoinedError_DiscriminatesAllSentinels`)
- ✓ Self-pairs in SuppressedPairs silently accepted (D-05; verified by `TestValidate_SelfPairInSuppressed_OK`)
- ✓ Pipeline order matches CONTEXT.md §2: P1 → P2 → P3 → P4 (verified by the four `TestValidate_FailFastOrder_*` tests)
- ✓ Plan 09-03 can call validateCheck as its first step in Check

## Self-Check: PASSED

**Created files exist:**

- `scan/validate.go` — FOUND
- `scan/validate_test.go` — FOUND
- `.planning/phases/09-collection-scan-sub-package/09-02-SUMMARY.md` — FOUND (this file)

**Commits exist:**

- `77cf263` — FOUND (test(scan): add failing tests for validation pipeline P1..P4 (TDD RED))
- `0dd1305` — FOUND (feat(scan): pre-flight validation pipeline P1..P4 with errors.Join collect-all)

**Gates pinned:**

- `make check` exits 0
- `go test -race -shuffle=on -count=1 ./...` all green (root + scan/)
- `go tool cover -func=/tmp/scan_cover.out | grep validate.go` shows 100% per function on every validate.go function
- `gofmt -l scan/` returns empty
- `go vet ./scan/...` clean
- `golangci-lint run ./scan/...` returns "0 issues"
- `make verify-deps-allowlist` clean (2 non-indirect modules: github.com/axonops/fuzzymatch + golang.org/x/text)
- `make verify-license-headers` clean (208 .go files)

---

*Phase: 09-collection-scan-sub-package*
*Plan: 02*
*Completed: 2026-05-20*
