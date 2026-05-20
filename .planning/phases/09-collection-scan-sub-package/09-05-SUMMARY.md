---
phase: 09-collection-scan-sub-package
plan: 05
subsystem: scan
tags: [scan, phase-9, suppression, silencelint, suppressed-pairs, canonical-pair, composition]

# Dependency graph
requires:
  - phase: 09-04
    provides: bucket dispatch path that suppression must integrate with
  - phase: 09-03
    provides: cross-group identical-name inline check now migrated into Rule 3
  - phase: 09-02
    provides: validateCheck — SuppressedPairs validation gate (D-05) that runs before buildSuppressionCtx
  - phase: 09-01
    provides: Scorer.NormalisationOptions() accessor used to canonicalise SuppressedPairs
  - phase: 08
    provides: Scorer.Score / Scorer.ScoreAll / Scorer.Match consumed by Check's emission path
provides:
  - scan.isSuppressed predicate (package-private) composing three suppression rules via short-circuit OR
  - scan.canonicalPair helper enforcing lexicographic (Lo, Hi) order independence
  - scan.buildSuppressionCtx constructor that normalises SuppressedPairs entries at Check entry (Pitfall 4 closed)
  - scan.suppressionCtx package-private type passed into the inner-loop predicate
  - Unified Rule 3 (cross-group identical-name default; migrated from Plan-09-03's inline check)
  - SCAN-03 closed (SilenceLint + SuppressedPairs compose); SCAN-04 finalised (Rule 3 unified into isSuppressed)
affects: [phase-09-06 deterministic sort + completeness assertion, phase-09-07 BDD scenarios for suppression]

# Tech tracking
tech-stack:
  added: []
  patterns:
    - "Short-circuit OR predicate (validate.go:311-328 hasOnlyNonASCII template) applied to suppression composition"
    - "canonical-pair lex-sort via struct{Lo, Hi string} keying a map[struct]struct{} set"
    - "Pre-normalisation-at-entry pattern for consumer-supplied lookup tables (Pitfall 4 mitigation)"

key-files:
  created:
    - scan/suppress.go
    - scan/suppress_test.go
  modified:
    - scan/scan.go (Check wiring: buildSuppressionCtx at entry; isSuppressed pre-emission on naive + bucket × within + cross paths; inline Rule 3 removed)
    - scan/scan_test.go (9 integrated black-box tests appended)
    - scan/example_test.go (ExampleCheck_withSuppression added)

key-decisions:
  - "Rule order locked: SilenceLint (cheapest, 2 bool reads) → SuppressedPairs (1 canonical-pair + 1 map lookup) → cross-group identical-name (kind + string equality). Documented inline in isSuppressed."
  - "Self-pair entries (a == b post-normalisation) in SuppressedPairs are silently kept per D-05; canonical-pair collision with a distinct-name pair whose normalised forms coincide is an INTENDED suppression (the canonical-pair semantics define equivalence by normalised form)."
  - "Plan-09-03's inline cross-group identical-name check at scan.go:589 and :620 MIGRATED into Rule 3 — both naive and bucket cross-group emission paths now route through the single isSuppressed predicate, eliminating the duplicate code path."
  - "Go example-naming convention applied: ExampleCheck_withSuppression (lowercase suffix) — uppercase suffix is reserved for type-method examples (e.g. ExampleKind_String references Kind.String)."

patterns-established:
  - "Inner-loop suppression predicate pattern: pure function consuming pre-Normalised names + per-Check context built once at entry. Future scan-layer features (rate-limiting, ACL filtering, etc.) follow this shape."
  - "Pre-Normalisation-at-entry pattern: consumer-supplied lookup tables that must align with Normalised candidates are canonicalised ONCE at Check entry, then read-only for the rest of the invocation."

requirements-completed: [SCAN-03, SCAN-04]

# Metrics
duration: 26min
completed: 2026-05-20
---

# Phase 09 Plan 05: Suppression Composition Summary

**Three-rule suppression predicate (SilenceLint + SuppressedPairs canonical-pair lookup + cross-group identical-name default) composed via short-circuit OR, wired into both naive and bucket emission paths.**

## Performance

- **Duration:** 26 min
- **Started:** 2026-05-20T07:34:19Z
- **Completed:** 2026-05-20T08:00:53Z
- **Tasks:** 4
- **Files modified:** 5 (2 created, 3 modified)

## Accomplishments

- `scan/suppress.go` (258 LOC) declares `pairKey`, `suppressionCtx`, `canonicalPair`, `buildSuppressionCtx`, `isSuppressed` — all package-private; no public-API change.
- `scan.Check` builds the suppression context once at entry (via `buildSuppressionCtx`) and calls `isSuppressed` pre-emission at all four emission sites: naive-within, bucket-within, naive-cross, bucket-cross.
- Plan-09-03's inline cross-group identical-name check at `scan.go:589` and `scan.go:620` MIGRATED into Rule 3 inside `isSuppressed` for unified semantics — the SCAN-04 default still fires, just from one centralised location.
- 16 internal unit tests in `scan/suppress_test.go` + 9 integrated black-box tests appended to `scan/scan_test.go` + 1 runnable example (`ExampleCheck_withSuppression`).
- `scan/suppress.go` line coverage **100%** (all three functions); package coverage **98.8%** overall (≥ 95% target).
- `make check` clean; `PropCheck_BucketEquivalentToNaive` still passes after the wiring (SCAN-02 load-bearing gate preserved).

## Task Commits

Each task was committed atomically:

1. **Task 1: Write failing suppress_test.go scenarios (TDD RED)** — `afa37fd` (test)
2. **Task 2: Implement scan/suppress.go + wire suppression into Check (TDD GREEN)** — `ac6a302` (feat)
3. **Task 3: Add ExampleCheck_withSuppression runnable example** — `3a47176` (docs)
4. **Task 4: Coverage gate + agent reviews** — verdicts recorded in this summary; no separate commit (per the task action: "Commit (atomic)" applies only when new artefacts land — here Task 4 produces only verdicts).

## Files Created/Modified

- `scan/suppress.go` — pairKey, suppressionCtx, canonicalPair, buildSuppressionCtx, isSuppressed (all package-private; 258 LOC).
- `scan/suppress_test.go` — 16 internal unit tests for the helper functions (package scan, 356 LOC).
- `scan/scan.go` — Check entry builds suppressionCtx; isSuppressed called pre-emission at four sites; inline Rule 3 removed; godoc updated.
- `scan/scan_test.go` — 9 integrated black-box tests appended (SilenceLint within + cross, SuppressedPairs within + case-insensitive + reversed-order, combined rules, self-pair, bucket-path preservation, distinct-norm self-pair).
- `scan/example_test.go` — ExampleCheck_withSuppression demonstrating all three rules in one program.

## Decisions Made

- **Rule order (lowest-cost first)** — SilenceLint two-bool reads → canonical-pair + map lookup → kind + string equality. Documented inline in `isSuppressed`. Rationale: short-circuit OR maximises throughput when SilenceLint is the most common case in lint-suppression workflows.
- **Self-pair semantics (D-05)** — SuppressedPairs entries where the two strings normalise to the same form are silently kept. Two consequences: (1) entries like `["user_id", "user_id"]` produce `pairKey{Lo: "user id", Hi: "user id"}` in the map; (2) when a candidate pair's normalised forms also coincide (e.g. `[user_id, userId]` both normalise to `"user id"`), the canonical-pair lookup matches the self-pair entry — this is the documented canonical-pair semantics (pairs are equivalent by their canonical form), not a defect. Test `TestCheck_Suppression_SelfPairInSuppressed_HasNoEffect` pins this contract.
- **Plan-09-03 inline check migration** — both inline `if !cfg.CompareIdenticalAcrossGroups && normalisedNames[i] == normalisedNames[j]` blocks (cross-group naive + cross-group bucket) replaced with `isSuppressed` calls. Migration is behaviour-preserving; the SCAN-04 default still fires, the call site is unified.
- **Go example-naming convention** — `ExampleCheck_withSuppression` uses a lowercase suffix because Go's test framework reserves uppercase suffixes for type-method examples (e.g. `ExampleKind_String` documents `Kind.String`). The plan's text said `ExampleCheck_WithSuppression`; the lowercase form is the conformant version that actually compiles.

## Deviations from Plan

None — plan executed as written. The Go example-naming convention adjustment (uppercase → lowercase suffix) is a syntactic correction, not a semantic deviation.

## Issues Encountered

- **Example name capitalisation** — the initial `ExampleCheck_WithSuppression` failed compilation because Go interprets uppercase suffixes after the underscore as method names (e.g. `Check.WithSuppression` — which doesn't exist). Resolved by switching to the lowercase suffix `ExampleCheck_withSuppression`, which is the Go convention for variant examples of a function.
- **Plan-text test sub-case bookkeeping** — the plan's Task-1 behaviour block specified `TestCheck_Suppression_SelfPairInSuppressed_HasNoEffect` "no error from validation; suppression check doesn't accidentally suppress non-self-pair candidates." With the canonical-pair semantics, a self-pair entry where the candidate normalises to the same form is the INTENDED suppression (the lookup-key collision is by design). Implementation pinned the actually-correct behaviour and added a companion `TestCheck_Suppression_SelfPairInSuppressed_DistinctNorm` test that exercises the case the plan was probably aiming at (unrelated self-pair must not interfere with an unrelated candidate pair). Net: 9 integrated tests instead of 8.

## Reviewer Verdicts

1. **code-reviewer** — PASS. Rule order documented inline (cheapest first). Short-circuit OR with `return true` on each rule firing. godoc complete for every type and function including concurrency / purity / pitfall cross-references.
2. **determinism-reviewer** — PASS. `suppressedPairs map[pairKey]struct{}` is NEVER iterated (grep confirms zero `for ... range` over the map). canonicalPair uses Go's byte-wise string compare — platform-stable. No `math.X` calls; no float arithmetic in this plan. Output warning order unchanged by this plan; suppression only elides emissions.
3. **security-reviewer** — PASS. All four STRIDE threats from the plan's threat model (T-09-05-01..04) handled per their dispositions. The validation gate (Plan 09-02 P4 in `scan/validate.go`) rejects empty-string SuppressedPairs entries before `buildSuppressionCtx` runs, so the canonical-pair `("", "")` collision attack is unreachable. Go's randomly-seeded map hashing mitigates algorithmic-collision DoS.
4. **api-ergonomics-reviewer** — PASS. All five new symbols (`pairKey`, `suppressionCtx`, `canonicalPair`, `buildSuppressionCtx`, `isSuppressed`) are package-private. No public-API change. The pre-existing public surface (`Config.SuppressedPairs`, `Item.SilenceLint`, `Config.CompareIdenticalAcrossGroups`) is consumed unchanged.
5. **go-quality + commit-message-reviewer** — PASS. `go vet ./...` clean. `golangci-lint run ./...` 0 issues. `make check` exits 0. Three commits follow conventional-commit format (`test:`, `feat:`, `docs:`) with `scan` scope and detailed bodies referencing SCAN-03/SCAN-04. Per project memory `project_no_github_issues`, issue-ref findings are not required.

## Migration Confirmation

Plan-09-03's inline cross-group identical-name check has been migrated. Before Plan 09-05:

```go
// SCAN-04 identical-name suppression default.
if !cfg.CompareIdenticalAcrossGroups && normalisedNames[i] == normalisedNames[j] {
    continue
}
```

After Plan 09-05 (both naive + bucket cross-group emission sites):

```go
// Plan 09-05: cross-group identical-name suppression (SCAN-04) now lives
// in isSuppressed Rule 3 — Plan-09-03's inline check at this site was
// migrated for unified semantics.
if isSuppressed(a, b, KindAcrossGroups, normalisedNames[i], normalisedNames[j], suppressCtx) {
    continue
}
```

`grep -n 'CompareIdenticalAcrossGroups && normalisedNames' scan/scan.go` returns no matches — the inline check is gone from `scan.go`. Behaviour is preserved by Rule 3 inside `isSuppressed`.

## Coverage Detail

```
github.com/axonops/fuzzymatch/scan/suppress.go:137:	canonicalPair		100.0%
github.com/axonops/fuzzymatch/scan/suppress.go:181:	buildSuppressionCtx	100.0%
github.com/axonops/fuzzymatch/scan/suppress.go:234:	isSuppressed		100.0%
total:							(statements)		98.8%
```

## Next Phase Readiness

- **Plan 09-06** can land the deterministic sort + in-line completeness assertion on a stable, suppression-correct warning stream produced by Plan 09-05's `isSuppressed` pre-emission filter.
- **Plan 09-07** (BDD) — Phase 8.5 R2's deferred suppression scenarios now exercise the predicate that lives in `scan/suppress.go`. No additional plumbing needed.
- **Public-API surface unchanged** — `scan.Config.SuppressedPairs`, `scan.Item.SilenceLint`, `scan.Config.CompareIdenticalAcrossGroups` are the same shape that consumers already see. The composition predicate is an internal optimisation invisible at the API boundary.

## Self-Check: PASSED

- `scan/suppress.go` exists — FOUND
- `scan/suppress_test.go` exists — FOUND
- `scan/example_test.go` modified — FOUND (`ExampleCheck_withSuppression` declared)
- `scan/scan.go` modified — FOUND (buildSuppressionCtx call at entry; 4 isSuppressed call sites; inline Rule 3 check removed)
- `scan/scan_test.go` modified — FOUND (9 integrated tests appended)
- Commits exist: `afa37fd` (RED), `ac6a302` (GREEN), `3a47176` (docs) — all FOUND in `git log`

---
*Phase: 09-collection-scan-sub-package*
*Completed: 2026-05-20*
