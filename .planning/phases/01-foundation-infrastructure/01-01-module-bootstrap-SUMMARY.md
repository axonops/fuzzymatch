---
phase: 01-foundation-infrastructure
plan: 01
subsystem: infra
tags: [go-module, apache-2.0, godog, goleak, testify, golang.org/x/text, bash, ci-gate]

# Dependency graph
requires:
  - phase: 01-foundation-infrastructure
    provides: phase context (CONTEXT.md), authoritative spec (PROJECT.md, docs/requirements.md), research stack (STACK.md)
provides:
  - root Go module at the frozen freeze point (Go 1.26.3 + golang.org/x/text v0.37.0 only)
  - isolated tests/bdd Go sub-module (godog v0.15.0, goleak v1.3.0, testify v1.10.0) with replace directive
  - scripts/verify-license-headers.sh CI gate (idempotent, dual-mode discovery via git or find)
  - canonical Apache-2.0 file-header convention demonstrated in doc.go and tests/bdd/doc.go
  - empty directory scaffolding: testdata/golden, testdata/fuzz, tests/bdd/features, tests/bdd/steps (.gitkeep)
affects: 01-02-quality-gates, 01-03-release-pipeline, 01-04-determinism-infra, 01-05-primitives-algoid-errors, 01-06-primitives-normalise, every subsequent plan

# Tech tracking
tech-stack:
  added:
    - golang.org/x/text v0.37.0 (sole runtime dep; blank-imported via unicode/norm to pin direct status)
    - github.com/cucumber/godog v0.15.0 (BDD framework, tests/bdd only)
    - go.uber.org/goleak v1.3.0 (goroutine-leak detector, tests/bdd only)
    - github.com/stretchr/testify v1.10.0 (assertions, tests/bdd only)
  patterns:
    - "Module freeze point: root go.mod immutable after plan 01-01; new runtime deps require allowlist amendment + algorithm-licensing-reviewer sign-off"
    - "Isolated BDD sub-module via replace directive at ../..; test deps never leak into root go.mod"
    - "Canonical Apache-2.0 header on every .go file, verified by scripts/verify-license-headers.sh"
    - "Blank-import freeze technique: doc.go carries `import _ \"<dep>\"` to keep otherwise-unused requires direct (not // indirect) until functional code arrives"
    - "Bash CI scripts: shebang `#!/usr/bin/env bash`, `set -euo pipefail`, dual-mode discovery (git ls-files preferred, find fallback), idempotent, exit-code semantics documented in header comment"

key-files:
  created:
    - go.mod (root: module path, Go 1.26.3 directive, single golang.org/x/text require)
    - go.sum (root: x/text checksums only)
    - doc.go (root package fuzzymatch godoc + Apache-2.0 header + blank import pinning x/text)
    - tests/bdd/go.mod (sub-module with godog/goleak/testify + replace ../.. + placeholder fuzzymatch)
    - tests/bdd/go.sum (BDD-only checksums; never reachable from root)
    - tests/bdd/doc.go (package bdd godoc + Apache-2.0 header + blank-import pinning of test deps)
    - tests/bdd/features/.gitkeep (Gherkin feature-file directory placeholder)
    - tests/bdd/steps/.gitkeep (step-definition directory placeholder)
    - testdata/golden/.gitkeep (placeholder for canonical-form golden files, used from plan 01-04 onwards)
    - testdata/fuzz/.gitkeep (placeholder for fuzz corpora, used from Phase 2 onwards)
    - scripts/verify-license-headers.sh (executable; CI gate)
  modified: []

key-decisions:
  - "golang.org/x/text pinned at v0.37.0 (latest stable at execution time per `go list -m -versions`); plan permitted any newer-than-v0.30.0 stable tag, v0.37.0 is current"
  - "Spec-locked versions godog v0.15.0 / goleak v1.3.0 / testify v1.10.0 retained (downgraded from go-mod-tidy-resolved godog v0.15.1 / testify v1.11.1) to honour the plan and STACK.md spec-lock"
  - "Blank import `import _ \"golang.org/x/text/unicode/norm\"` in root doc.go to pin x/text as a DIRECT (not // indirect) require at the freeze point; same technique used in tests/bdd/doc.go to pin the three test-only deps"
  - "Module freeze point now in effect: subsequent plans MUST NOT modify the root require block. Any future runtime-dep additions require explicit allowlist amendment + algorithm-licensing-reviewer sign-off (per CLAUDE.md and PROJECT.md Constraints)"

patterns-established:
  - "AxonOps Apache-2.0 header (Copyright 2026 AxonOps Limited + standard 11-line Apache-2.0 boilerplate) is the canonical .go file header; the literal substring 'Licensed under the Apache License, Version 2.0' is the verifier needle"
  - "doc.go pattern: header → blank line → // Package <name> godoc paragraph → package <name> → optional blank imports for module-freeze-point dependency pinning"
  - "BDD sub-module isolation pattern: tests/bdd/go.mod with `replace github.com/axonops/fuzzymatch => ../..` and a placeholder version `v0.0.0-00010101000000-000000000000`"
  - "CI verifier scripts: bash with `set -euo pipefail`, `git ls-files`-preferred discovery, `find` fallback for tarball CI, offending paths to stderr one-per-line, success message on stdout, exit codes 0/1 with the script's own header documenting the contract"

requirements-completed:
  - FOUND-01

# Metrics
duration: ~15min
completed: 2026-05-13
---

# Phase 1 Plan 1: Module Bootstrap Summary

**Frozen Go module at Go 1.26.3 + golang.org/x/text v0.37.0 with isolated tests/bdd sub-module (godog/goleak/testify) and an executable Apache-2.0 license-header CI gate.**

## Performance

- **Duration:** ~15 min
- **Started:** 2026-05-13T11:14:00Z (approximate; agent spawn)
- **Completed:** 2026-05-13T11:20:51Z
- **Tasks:** 3 of 3
- **Files created:** 11 (go.mod, go.sum, doc.go, tests/bdd/go.mod, tests/bdd/go.sum, tests/bdd/doc.go, tests/bdd/features/.gitkeep, tests/bdd/steps/.gitkeep, testdata/golden/.gitkeep, testdata/fuzz/.gitkeep, scripts/verify-license-headers.sh)

## Accomplishments

- Root `go.mod` initialised at the freeze point: module `github.com/axonops/fuzzymatch`, directive `go 1.26.3`, exactly one direct require (`golang.org/x/text v0.37.0`), no `toolchain` line. `go build`, `go vet`, `go mod tidy` all green and idempotent.
- `doc.go` carries the canonical AxonOps Apache-2.0 header and the root package godoc (multi-paragraph overview covering the three-layer architecture and pointing at `docs/requirements.md` as authoritative).
- `tests/bdd/` sub-module created with its own `go.mod` (godog v0.15.0, goleak v1.3.0, testify v1.10.0, placeholder fuzzymatch overridden by `replace ../..`). Root `go.mod` is unchanged by this step.
- `scripts/verify-license-headers.sh` is the CI gate that fails any .go file missing the Apache-2.0 header within its first 25 lines; idempotent, dual-mode discovery (git ls-files preferred / find fallback), passes on the current tree, validated interactively against a header-less scratch file (exit 1 + offending path on stderr).
- Empty directory scaffolding committed via `.gitkeep` for `testdata/golden`, `testdata/fuzz`, `tests/bdd/features`, `tests/bdd/steps` so subsequent plans inherit the layout.

## Task Commits

Each task was committed atomically on `worktree-agent-a2481f6dd2beb4823`:

1. **Task 1: Initialise root go.mod with Go 1.26.3 and golang.org/x/text** — `dc5880e` (feat)
2. **Task 2: Create tests/bdd sub-module with isolated test dependencies** — `ded9ed1` (chore)
3. **Task 3: Create license-header verifier and scaffold remaining directories** — `e9ffaec` (chore)

## Files Created/Modified

Created:
- `go.mod` — module declaration, `go 1.26.3`, single direct require `golang.org/x/text v0.37.0`
- `go.sum` — checksums for x/text only
- `doc.go` — Apache-2.0 header + root package godoc + blank import of `golang.org/x/text/unicode/norm` (freeze-point pin)
- `tests/bdd/go.mod` — sub-module declaration, `go 1.26.3`, replace directive `github.com/axonops/fuzzymatch => ../..`, spec-locked test deps
- `tests/bdd/go.sum` — checksums for godog/goleak/testify and their transitives (yaml.v3, cucumber-messages, gherkin, go-spew, go-difflib, etc.)
- `tests/bdd/doc.go` — Apache-2.0 header + `package bdd` godoc + blank-import freeze of the three test deps and the parent module
- `tests/bdd/features/.gitkeep` — empty marker for Gherkin feature files (populated from plan 01-?? onwards)
- `tests/bdd/steps/.gitkeep` — empty marker for godog step definitions
- `testdata/golden/.gitkeep` — empty marker for canonical-form golden files (used by determinism golden test from plan 01-04)
- `testdata/fuzz/.gitkeep` — empty marker for fuzz corpora (used from Phase 2 onwards)
- `scripts/verify-license-headers.sh` — executable bash CI gate (chmod +x at creation)

Modified: none (no pre-existing files touched).

## Decisions Made

1. **`golang.org/x/text` version: v0.37.0** — the plan permits "whatever is newest stable at execution time". `go list -m -versions golang.org/x/text` reports the current latest stable as v0.37.0; pinned there.
2. **Spec-locked test-dep versions retained** — `go mod tidy` initially auto-resolved to godog v0.15.1 and testify v1.11.1 (the latest stable). Per the plan's explicit version pins and `.planning/research/STACK.md`'s spec-lock, downgraded to godog v0.15.0 and testify v1.10.0 via `go get @v0.15.0` / `go get @v1.10.0`.
3. **Blank-import freeze technique** — to honour both "x/text must be a direct require in go.mod after this plan" and "doc.go is solely the package-godoc anchor and `go mod tidy` must be idempotent", added a single blank import (`import _ "golang.org/x/text/unicode/norm"`) at the bottom of `doc.go`. This keeps the require direct (not `// indirect`) and survives `go mod tidy`. The same technique is used in `tests/bdd/doc.go` for the three test deps and the parent-module placeholder.

## Deviations from Plan

### Auto-fixed Issues

**1. [Rule 3 - Blocking] Blank-import freeze in doc.go to honour internally-conflicting acceptance criteria**

- **Found during:** Task 1 (Initialise root go.mod) and Task 2 (BDD sub-module)
- **Issue:** The plan simultaneously requires (a) `go.mod` carrying a direct require for `golang.org/x/text` (must-have plus acceptance-criterion `grep -E '^\s*golang\.org/x/text v' go.mod` matches exactly one line); (b) `go mod tidy` produces no diff (idempotent); (c) `doc.go` has "NO declarations beyond the package clause". These three together are only simultaneously satisfiable if at least one Go source file in the package transitively imports the dependency — otherwise `go mod tidy` removes the unused require (verified empirically: a tidy of a doc-comment-only root deleted the x/text require line). The same problem applies, three-fold, to the tests/bdd sub-module for godog/goleak/testify.
- **Fix:** Added a single blank import to each `doc.go`:
  - Root `doc.go`: `import _ "golang.org/x/text/unicode/norm"`
  - `tests/bdd/doc.go`: blank imports for `github.com/axonops/fuzzymatch`, `github.com/cucumber/godog`, `github.com/stretchr/testify/assert`, `go.uber.org/goleak`.
  These are imports (not type/var/func/const declarations), preserve the godoc placement, are gofmt-clean, do not change runtime behaviour beyond running each package's `init()` (intentional: x/text/unicode/norm tables are exactly what later plans need ready), and survive `go mod tidy`. The blank-import block in each file is preceded by a comment explaining the freeze-point rationale.
- **Files modified:** `doc.go`, `tests/bdd/doc.go`
- **Verification:** `go mod tidy` is now idempotent in both modules; `grep '^require'` returns the expected lines; `go build ./...` and `go vet ./...` are green; `gofmt -l` reports no changes needed.
- **Committed in:** `dc5880e` (root) and `ded9ed1` (tests/bdd) — part of the Task 1 and Task 2 commits respectively.

**2. [Rule 3 - Blocking] `go list -m all` returns 5 modules, not the plan's stated "exactly two"**

- **Found during:** Task 1 verification
- **Issue:** The plan's acceptance criterion states `go list -m all` should return exactly two modules (main + x/text). In practice, x/text v0.37.0's own `go.mod` declares require lines for `golang.org/x/tools`, `golang.org/x/mod`, and `golang.org/x/sync` (for its internal code generators — `tagx:ignore` markers in upstream confirm these are build-time tooling, not runtime). With Go 1.17+ lazy module loading, these phantom modules appear in `go list -m all` even though they are NOT downloaded, NOT in our `go.sum`, NOT in our compiled output, and `go mod why <each>` returns "main module does not need package <each>".
- **Fix:** Accepted as a property of x/text's upstream `go.mod` rather than a defect of this plan. The plan's spirit (no direct runtime deps beyond x/text; no transitive runtime deps in compiled artefact) is satisfied:
  - `go list -m -f '{{if not .Indirect}}{{.Path}} {{.Version}}{{end}}' all` lists exactly: main module + x/text.
  - Our `go.sum` carries only `golang.org/x/text v0.37.0` hashes (verified by inspection).
  - The phantom modules are never fetched (verified with `go mod why` returning "not needed").
- **Files modified:** none (no fix; this is a documentation-only deviation).
- **Verification:** see `go.sum` (4-line file: x/text h1 + x/text go.mod hash). The phantom-modules behaviour is intrinsic to x/text v0.37.0 and would only change if x/text itself dropped its tool-only requires upstream.
- **Committed in:** n/a (documented here for traceability; no code change).

---

**Total deviations:** 2 auto-fixed (both Rule 3 — blocking issues with the plan's internal consistency).
**Impact on plan:** Both fixes are necessary, conservative, and preserve every concrete must-have. The blank-import freeze pattern is the standard Go idiom for "module depends on X even though no functional code uses X yet"; it does not alter the freeze-point contract — the root require block remains a single line, and the test deps remain isolated to tests/bdd. The `go list -m all` clarification is a pure documentation refinement; no observable behaviour changes.

## Issues Encountered

- Local Go toolchain is 1.26.1; plan requires `go 1.26.3` directive. With `GOTOOLCHAIN=auto` (the default), `go get` automatically downloaded `go1.26.3` (darwin/arm64) and ran the subsequent commands under it. No user intervention needed.
- Initial `go mod tidy` in the root module stripped the `golang.org/x/text` require (no importer). Resolved by the blank-import freeze documented above.

## User Setup Required

None — no external service configuration, no environment variables, no manual dashboard steps.

## Threat Surface Scan

No new security-relevant surface introduced beyond the trust-boundaries documented in the plan's `<threat_model>`:
- T-01-01 (runtime allowlist) — baseline established; `make verify-deps-allowlist` (plan 01-04) is the enforcing gate.
- T-01-02 (license-header drift) — `scripts/verify-license-headers.sh` is now in place; plan 01-02 wires it into CI.
- T-01-03 (test-dep leakage) — tests/bdd is structurally isolated; root `go.mod` carries exactly one direct require.

No threat_flags to raise.

## Next Plan Readiness

- Plan 01-02 (quality gates) inherits a working, gofmt-clean, lint-able module skeleton with `scripts/verify-license-headers.sh` ready to wire into the Makefile and CI workflow.
- Plan 01-03 (release pipeline) inherits the canonical license-header convention to embed in goreleaser archive boilerplate and per-file checks.
- Plan 01-04 (determinism infra) inherits `testdata/golden/.gitkeep` and the BDD sub-module ready for godog scenarios.
- All subsequent plans inherit the **module freeze point**: the root `go.mod` require block is frozen. Any plan that needs an additional runtime dep MUST go through allowlist amendment review.

## Self-Check: PASSED

Files claimed to exist (verified with `[ -f ... ]`):
- `go.mod` — FOUND
- `go.sum` — FOUND
- `doc.go` — FOUND
- `tests/bdd/go.mod` — FOUND
- `tests/bdd/go.sum` — FOUND
- `tests/bdd/doc.go` — FOUND
- `tests/bdd/features/.gitkeep` — FOUND
- `tests/bdd/steps/.gitkeep` — FOUND
- `testdata/golden/.gitkeep` — FOUND
- `testdata/fuzz/.gitkeep` — FOUND
- `scripts/verify-license-headers.sh` — FOUND (executable)

Commits claimed to exist (verified with `git log`):
- `dc5880e` — FOUND (Task 1)
- `ded9ed1` — FOUND (Task 2)
- `e9ffaec` — FOUND (Task 3)

Plan-level verification (from `<verification>` section):
- `go build ./...` exit 0 — PASS
- `go vet ./...` exit 0 — PASS
- `grep -c '^require' go.mod` = `1` — PASS
- `cd tests/bdd && go build ./... && go vet ./...` exit 0 — PASS
- `bash scripts/verify-license-headers.sh` exit 0 — PASS
- Failure-path validation (scratch `.go` file rejected with exit 1) — PASS (interactively confirmed)

---
*Phase: 01-foundation-infrastructure*
*Plan: 01 (module-bootstrap)*
*Completed: 2026-05-13*
