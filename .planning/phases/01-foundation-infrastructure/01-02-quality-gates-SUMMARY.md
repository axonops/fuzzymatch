---
phase: 01-foundation-infrastructure
plan: 02
subsystem: infra
tags: [golangci-lint, makefile, ci, dependabot, govulncheck, gosec, codeql, markdownlint, quality-gate]

# Dependency graph
requires:
  - phase: 01-foundation-infrastructure
    provides: root go.mod, tests/bdd go.mod, scripts/verify-license-headers.sh, canonical Apache-2.0 header convention (plan 01-01)
provides:
  - .golangci.yml — golangci-lint v2 schema configuration (curated linter set; misspell UK; gocyclo 10; formatters block)
  - .markdownlint-cli2.yaml — markdownlint configuration excluding planning artefacts
  - Makefile — 20 canonical quality-gate targets including aggregate `check`
  - .github/workflows/ci.yml — runs `make check` + markdownlint on PR / push to main
  - .github/workflows/security.yml — govulncheck + gosec (SARIF) on PR / push / weekly Mondays
  - .github/workflows/codeql.yml — Go semantic analysis on PR / push / weekly Tuesdays
  - .github/workflows/license-headers.yml — wraps scripts/verify-license-headers.sh
  - .github/dependabot.yml — daily PRs for gomod (root + tests/bdd) and github-actions, grouped
affects: 01-03-release-pipeline (consumes CI baseline + adds commitlint + CLA + release workflows), 01-04-determinism-infra (expands ci matrix to 5 platforms + lands verify-deps-allowlist.sh / determinism golden / benchstat harness), every subsequent plan (gated by `make check` on every PR)

# Tech tracking
tech-stack:
  added: []  # No runtime deps; only CI/local tooling configuration
  patterns:
    - "golangci-lint v2 schema: `version: \"2\"` as first non-comment line; formatters block split from linters block per v2 layout; misspell locale UK matches documentation-standards British English"
    - "Makefile aggregate pattern: `check` chains fmt-check → vet → lint → verify-license-headers → verify-deps-allowlist → tidy-check → security → test → coverage → coverage-check, with later-plan-dependent targets (verify-deps-allowlist, release-check, bench, bench-compare) as tolerant no-ops printing pending messages"
    - "CI least-privilege: workflow-level `permissions: contents: read`; per-job escalation only where required (security-events: write for gosec SARIF; actions: read for CodeQL)"
    - "Dependabot grouped-PRs: gomod direct/indirect splits; single github-actions group; `commit-message.prefix` per ecosystem (chore / chore(test-bdd) / chore(ci)) keeps the conventional-commits stream clean"
    - "goimports `-local github.com/axonops/fuzzymatch` enforces a three-group import layout (stdlib | third-party | local), wired through the formatters block of golangci-lint and the Makefile fmt/fmt-check targets"

key-files:
  created:
    - .golangci.yml
    - .markdownlint-cli2.yaml
    - Makefile
    - .github/workflows/ci.yml
    - .github/workflows/security.yml
    - .github/workflows/codeql.yml
    - .github/workflows/license-headers.yml
    - .github/dependabot.yml
  modified:
    - tests/bdd/doc.go (Rule 3 fix — reorder blank imports for goimports -local)

key-decisions:
  - "golangci-lint v2.11.4 used locally; CI pins v2.12.2 per STACK.md. v2 schema is forwards-compatible between v2.11 and v2.12, so the same .golangci.yml drives both."
  - "Aggregated `make check` is the local single-entrypoint; CI runs the exact same target rather than re-listing the individual sub-commands. This guarantees CI cannot diverge from the local gate."
  - "Targets depending on later-plan artefacts (`verify-deps-allowlist`, `release-check`, `bench`, `bench-compare`) are tolerant no-ops printing a `pending plan 01-NN` message, NOT errors. This keeps `make check` green on the frozen tree and lets every later plan flip those messages off by adding the missing artefact."
  - "`coverage-check` parses `go tool cover -func` against COVERAGE_FLOOR=95.0% but treats an empty coverage profile (no test files yet — the current Phase 1 state) as `pending`, skipping the floor check. Plan 01-04 lands per-file and public-API enforcement."
  - "Cross-platform matrix expansion (linux/amd64, linux/arm64, darwin/amd64, darwin/arm64, windows/amd64) is intentionally deferred to plan 01-04 (determinism infra) so the matrix lands together with the determinism golden test that justifies it. ci.yml here runs ubuntu-latest only."
  - "gosec runs in two places: (a) informationally inside golangci-lint via `make lint` for early developer feedback; (b) definitively in .github/workflows/security.yml with SARIF upload to the GitHub Security tab. STACK.md mandates the security.yml job; the lint pass is a low-friction supplement."

patterns-established:
  - "Apache-2.0 file-header gate, CI-wired: license-headers.yml runs scripts/verify-license-headers.sh on every PR. Future plans MUST keep new .go files carrying the canonical header — the gate fails CI if they do not."
  - "Aggregate `make check` is the local pre-PR gate; CI mirrors it via .github/workflows/ci.yml. Contributors run `make check` before pushing; the same command runs in CI. No diff between local and CI semantics."
  - "Dependabot watches three ecosystems; PRs flow through the same CI gate as human PRs. No bypass for bot PRs."

requirements-completed:
  - CI-01
  - CI-02
  - CI-03
  - CI-04
  - CI-07
  - CI-08
  - CI-11
  - TEST-08

# Metrics
duration: ~20min
completed: 2026-05-13
---

# Phase 1 Plan 2: Quality Gates Summary

**Wired the canonical `make check` aggregate quality gate (lint, vet, fmt, race, vuln, license-headers, tidy, coverage) locally and in CI, with parallel govulncheck/gosec/CodeQL security workflows and Dependabot daily dependency PRs grouped by ecosystem.**

## Performance

- **Duration:** ~20 min
- **Started:** 2026-05-13T11:24Z (approximate; agent spawn)
- **Completed:** 2026-05-13T11:30Z
- **Tasks:** 3 of 3
- **Files created:** 8
- **Files modified:** 1 (tests/bdd/doc.go — Rule 3 fix)

## Accomplishments

- **`.golangci.yml` (v2 schema)** — `version: "2"` on the first non-comment line, curated linter set (errcheck, govet, staticcheck, revive, gocyclo, ineffassign, unused, unparam, misspell, gocritic, bodyclose, errorlint, exhaustive, gosec), misspell locale UK per documentation-standards, gocyclo min-complexity 10, dedicated `formatters` block (gofmt + gofumpt + goimports with `local-prefixes: github.com/axonops/fuzzymatch`), test-file exclusions for gocyclo/exhaustive/gosec. `golangci-lint run ./...` clean on the current tree.
- **`.markdownlint-cli2.yaml`** — relaxes line-length (MD013), inline-HTML (MD033), and first-line (MD041) rules; excludes `.planning/**`, `.claude/**`, `node_modules/**`, `vendor/**`, and `tests/bdd/**/vendor/**`.
- **`Makefile`** — all 20 canonical targets from CLAUDE.md "Makefile Targets" present. `make check` is the aggregate gate; later-plan-dependent targets (`verify-deps-allowlist`, `release-check`, `bench`, `bench-compare`) are tolerant no-ops printing a `pending plan 01-NN` message until those plans land their artefacts. `make check` exits 0 end-to-end on a clean tree.
- **`.github/workflows/ci.yml`** — workflow-level `permissions: contents: read`; `quality` job runs `make check` after installing pinned tools (golangci-lint v2.12.2, govulncheck@latest, goimports@latest); `markdownlint` job runs `DavidAnson/markdownlint-cli2-action@latest-stable`. Triggers: pull_request + push to main.
- **`.github/workflows/security.yml`** — govulncheck job + gosec job (uploading SARIF to the GitHub Security tab via `github/codeql-action/upload-sarif@v4`). Cron `0 6 * * 1` (Mondays 06:00 UTC) plus PR / push.
- **`.github/workflows/codeql.yml`** — `github/codeql-action@v4` (init + autobuild + analyze) with `build-mode: autobuild` and `languages: [go]`. Cron `0 6 * * 2` (Tuesdays 06:00 UTC) plus PR / push.
- **`.github/workflows/license-headers.yml`** — wraps `scripts/verify-license-headers.sh` from plan 01-01.
- **`.github/dependabot.yml`** — three ecosystems (gomod at `/`, gomod at `/tests/bdd`, github-actions at `/`) with grouped PRs (direct/indirect for gomod; single `actions` group for github-actions). Each ecosystem has a distinct `commit-message.prefix` (`chore` / `chore(test-bdd)` / `chore(ci)`) so the conventional-commits stream remains clean.

## Task Commits

Each task was committed atomically on `worktree-agent-ab93d11d6e2710893`:

1. **Task 1: Write .golangci.yml (v2 schema) and .markdownlint-cli2.yaml** — `972bef1` (chore)
2. **Task 2: Write the canonical Makefile** — `f7fe3b2` (chore; bundles the Rule 3 fix to tests/bdd/doc.go)
3. **Task 3: Write CI workflows + Dependabot** — `9317d76` (chore)

## Files Created/Modified

Created:
- `.golangci.yml` — golangci-lint v2 configuration (87 lines incl. header comment)
- `.markdownlint-cli2.yaml` — markdownlint v2 configuration (18 lines)
- `Makefile` — 20 canonical targets, ~180 lines including header comments and per-target documentation
- `.github/workflows/ci.yml` — quality + markdownlint jobs
- `.github/workflows/security.yml` — govulncheck + gosec jobs
- `.github/workflows/codeql.yml` — Go semantic analysis
- `.github/workflows/license-headers.yml` — license-header verifier wrapper
- `.github/dependabot.yml` — daily grouped PRs for three ecosystems

Modified:
- `tests/bdd/doc.go` — reordered blank imports so the local-prefix `github.com/axonops/fuzzymatch` import sits in its own group, satisfying `goimports -local github.com/axonops/fuzzymatch`. This was a pre-existing fmt drift from plan 01-01 surfaced by the new fmt-check target.

## Linter Inventory

The 14 linters enabled beyond golangci-lint v2's `default: standard` set:

| Linter         | Purpose                                                 |
|----------------|---------------------------------------------------------|
| errcheck       | Unchecked errors                                        |
| govet          | go vet pass list                                        |
| staticcheck    | Comprehensive correctness + style                       |
| revive         | golint replacement, configurable                        |
| gocyclo        | Cyclomatic complexity ≤ 10                              |
| ineffassign    | Ineffectual assignments                                 |
| unused         | Unused symbols                                          |
| unparam        | Unused function parameters                              |
| misspell       | British English (locale UK)                             |
| gocritic       | Bugs / performance / style diagnostics                  |
| bodyclose      | HTTP response body close (defensive — no net/http use yet) |
| errorlint      | Go 1.13 error wrapping correctness                      |
| exhaustive     | Enum switch exhaustiveness                              |
| gosec          | Informational security pass (definitive run in security.yml) |

Plus the formatters block: `gofmt -s`, `gofumpt`, `goimports` (with `local-prefixes: github.com/axonops/fuzzymatch`).

## Makefile Target Inventory

| Target                   | Behaviour                                                                 |
|--------------------------|---------------------------------------------------------------------------|
| `check` (aggregate)      | fmt-check → vet → lint → verify-license-headers → verify-deps-allowlist → tidy-check → security → test → coverage → coverage-check |
| `test`                   | `go test -race -shuffle=on -count=1 ./...`                                |
| `test-bdd`               | `cd tests/bdd && go test -race -count=1 ./...`                            |
| `test-fuzz`              | Discovers fuzzers; runs `go test -fuzz` for 60s each; no-op if none found |
| `lint`                   | `golangci-lint run ./...` + same in tests/bdd                             |
| `vet`                    | `go vet ./...` + same in tests/bdd                                        |
| `fmt`                    | `gofmt -s -w .` + `goimports -local github.com/axonops/fuzzymatch -w .`  |
| `fmt-check`              | Asserts both `gofmt -s -l` and `goimports -l` produce no output           |
| `bench`                  | `go test -bench=. -benchmem -count=10` if any benchmark exists; else no-op |
| `bench-compare`          | `benchstat bench.txt bench.txt.new` if both present; else pending message |
| `coverage`               | `go test -race -coverprofile=coverage.out -covermode=atomic ./...`       |
| `coverage-check`         | Floor 95.0%; tolerant of empty profile (pending Phase 2 algorithms)       |
| `tidy`                   | `go mod tidy` + same in tests/bdd                                         |
| `tidy-check`             | Runs tidy then asserts `git diff --exit-code` on the four mod files       |
| `security`               | govulncheck if installed; tolerant if missing                             |
| `verify-deps-allowlist`  | Runs `scripts/verify-no-runtime-deps.sh` if present (plan 01-04 lands it) |
| `verify-determinism`     | `go test -run TestGolden_ ./...` (plan 01-04 lands matching tests)        |
| `verify-license-headers` | `bash scripts/verify-license-headers.sh`                                  |
| `release-check`          | `goreleaser check` if `.goreleaser.yml` present (plan 01-03 lands it)     |
| `clean`                  | `go clean ./...` + remove coverage.out + bench.txt.new                    |

20 targets total (CLAUDE.md "Makefile Targets" lists 18 nominally but the canonical list extends to `coverage-check` and `verify-license-headers`).

## Action Versions (recorded for STACK.md traceability)

| Action / Tool                              | Pinned version           | STACK.md compliant |
|--------------------------------------------|--------------------------|--------------------|
| `actions/checkout`                         | `@v6`                    | ✅                 |
| `actions/setup-go`                         | `@v6` + `go-version-file: go.mod` | ✅          |
| `github/codeql-action/init`/autobuild/analyze | `@v4`                  | ✅                 |
| `github/codeql-action/upload-sarif`        | `@v4`                    | ✅                 |
| `DavidAnson/markdownlint-cli2-action`      | `@latest-stable`         | ✅ (STACK row 7)   |
| `securego/gosec`                           | `@v2.25.0`               | ✅                 |
| `golangci-lint` (installed in CI)          | `v2.12.2`                | ✅                 |
| `govulncheck` (installed in CI)            | `@latest`                | ✅                 |
| `goimports` (installed in CI)              | `@latest`                | ✅                 |

**No STACK.md drift.** Every pinned major matches STACK.md "Build & Quality Tooling" and "Release Stack" tables.

## Dependabot Group Configuration

| Ecosystem         | Directory        | Schedule | Open-PR cap | Groups                                       | commit-prefix    |
|-------------------|------------------|----------|-------------|----------------------------------------------|------------------|
| gomod             | `/`              | daily    | 5           | `direct` (dependency-type direct), `indirect` (indirect) | `chore`          |
| gomod             | `/tests/bdd`     | daily    | 5           | same                                         | `chore(test-bdd)` (labels: `test-only`) |
| github-actions    | `/`              | daily    | 5           | `actions` (patterns `"*"`)                   | `chore(ci)`      |

## Decisions Made

1. **golangci-lint v2 schema** — `version: "2"` confirmed as first non-comment line (line 12, after the explanatory header comment). The plan's stated probe `head -n 5 .golangci.yml | grep -qE '^version:\s*"?2"?'` is a heuristic; the acceptance criterion "First non-comment line is exactly `version: \"2\"`" is the binding contract and is satisfied.
2. **Aggregated `make check` as CI entrypoint** — CI runs the exact target rather than re-listing sub-commands. Guarantees zero local-vs-CI semantic drift.
3. **Tolerant no-op pattern for future-plan-dependent targets** — `verify-deps-allowlist`, `release-check`, `bench`, `bench-compare`, and (for the empty-profile case) `coverage-check` print a `pending plan 01-NN` message and exit 0. Each later plan flips the message off by landing the missing artefact.
4. **Cross-platform matrix deferred to plan 01-04** — ci.yml here runs `ubuntu-latest` only. Plan 01-04 (determinism infra) lands the linux/amd64 + linux/arm64 + darwin/amd64 + darwin/arm64 + windows/amd64 matrix together with the determinism golden test that justifies it.
5. **gosec runs in two places** — informationally inside golangci-lint (`make lint` for early feedback) and definitively in `security.yml` (with SARIF upload). Belt-and-braces; the CI definitive run is the gating one.

## Deviations from Plan

### Auto-fixed Issues

**1. [Rule 3 — Blocking] `goimports -local github.com/axonops/fuzzymatch` failed on `tests/bdd/doc.go`**

- **Found during:** Task 2 verification (`make fmt-check`).
- **Issue:** Plan 01-01's `tests/bdd/doc.go` grouped the four blank imports together (godog, testify, goleak, parent-module fuzzymatch). The new `formatters: goimports: local-prefixes: github.com/axonops/fuzzymatch` rule in `.golangci.yml` (and the matching `goimports -local …` invocation in the new Makefile `fmt` / `fmt-check` targets) requires the local `github.com/axonops/fuzzymatch` import to sit in its own group, separated by a blank line. The pre-existing layout violated this, blocking `make fmt-check` and therefore `make check`.
- **Fix:** Reordered the blank-import block in `tests/bdd/doc.go` to place the third-party deps (godog / testify / goleak) first, a blank line, then the local-prefix import (`github.com/axonops/fuzzymatch`). No semantic change — only ordering / grouping. `gofmt -s` clean; `go vet ./...` clean; `go mod tidy` idempotent; `make fmt-check` exits 0.
- **Files modified:** `tests/bdd/doc.go`
- **Verification:** `make fmt-check && make lint && make vet && make tidy-check` all exit 0.
- **Committed in:** `f7fe3b2` (bundled with the Makefile commit because the fmt-check target is what surfaces the issue).

**2. [Rule 1 — Bug] `coverage-check` produced spurious 0.0% failure on a tree with no tests yet**

- **Found during:** Task 2 final verification (`make check` end-to-end).
- **Issue:** First-cut `coverage-check` parsed `go tool cover -func=coverage.out` for the `total:` line and asserted `≥ 95.0%`. When the tree has no test files (current Phase 1 state), `go test -coverprofile=coverage.out` writes only the `mode: atomic` header line, and `go tool cover -func` still emits `total: (statements) 0.0%`, causing the floor check to fail spuriously. The plan explicitly requires the target to be "tolerant" — interpreted strictly as "absent profile" originally; the empty-but-present profile is the actual common case on a freshly-frozen module.
- **Fix:** Added a profiled-lines count step (`awk '!/^mode:/' coverage.out`) before the floor check. If the profile contains no profiled lines, print a `pending Phase 2` message and exit 0. The floor check only fires when there is actual coverage data to compare against. Plan 01-04 lands per-file and public-API enforcement, which will subsume this temporary tolerance.
- **Files modified:** `Makefile` (coverage-check recipe).
- **Verification:** `make coverage && make coverage-check` exits 0 with the pending message; `make check` end-to-end exits 0.
- **Committed in:** `f7fe3b2` (final-pass fix bundled with the Makefile commit).

**3. [Rule 1 — Bug] `grep -cv` returning exit 1 on zero-match polluted the coverage-check expression**

- **Found during:** First attempt at fix (2) above.
- **Issue:** Used `grep -cv '^mode:' coverage.out || echo 0` to count profiled lines. POSIX `grep -c` exits 1 when the count is zero, so the `|| echo 0` path triggered, producing the multi-line string `"0\n0"` and breaking the subsequent `-eq` test.
- **Fix:** Replaced the grep pipeline with `awk 'BEGIN{n=0} !/^mode:/{n++} END{print n}'` which always exits 0 and produces a single integer.
- **Files modified:** `Makefile` (coverage-check recipe — second-pass fix).
- **Verification:** `make coverage-check` returns the expected pending message on an empty profile.
- **Committed in:** `f7fe3b2` (consolidated into the Makefile commit).

---

**Total deviations:** 3 auto-fixed (1 × Rule 3 — pre-existing fmt drift; 2 × Rule 1 — first-pass bugs in the new coverage-check recipe surfaced by integration testing). All fixes are surgical, preserve the plan's acceptance criteria, and were validated by re-running `make check` end-to-end.

## Issues Encountered

- **Locally installed `golangci-lint` is v2.11.4, plan/STACK pin v2.12.2.** The v2 schema is forwards-compatible between v2.11 and v2.12, so the same `.golangci.yml` drives both. CI installs v2.12.2 explicitly in `ci.yml`; local users get whatever they have on `PATH`. If a future linter is renamed or removed between minors, the contract reads from the CI-installed version, not the local one. No action needed for this plan.
- **`markdownlint-cli2` not installed locally.** CI exercises it via the GitHub Action. The configuration is valid YAML; markdown content on the tree (README, CHANGELOG, BOOTSTRAP, NOTICE) was not actively scanned but does not introduce new content in this plan.

## User Setup Required

None — no external service configuration. After this plan merges to main the maintainer should manually configure GitHub branch protection on `main` to require: `quality / make check`, `markdownlint / markdownlint`, `Security / govulncheck`, `Security / gosec`, `CodeQL / Analyze (go)`, `License Headers / verify-license-headers`. The threat-model row T-01-02-02 explicitly acknowledges this as manual configuration; the workflows themselves are correct and ready.

## Threat Surface Scan

Reviewed every workflow's permissions block and action-version pinning. No threat flags to raise:

- **T-01-02-01 (Elevation of Privilege — workflow permissions):** Every workflow uses `permissions: contents: read` at the workflow level. Per-job escalation is minimal: `security-events: write` on the gosec job for SARIF upload; `actions: read` + `security-events: write` on the CodeQL `analyze` job (required by the CodeQL action). `grep -rn 'write-all' .github/workflows/` returns nothing.
- **T-01-02-02 (Tampering — linter config drift):** `.golangci.yml` is enforced by `make lint` in `make check`; CI runs `make check`. Branch protection (manual config above) prevents merging without it.
- **T-01-02-03 (Information Disclosure — vulnerability findings):** govulncheck fails CI on HIGH/CRITICAL; gosec uploads SARIF to the GitHub Security tab. Both run weekly even without PR traffic.
- **T-01-02-04 (Spoofing — malicious Dependabot PR):** All Dependabot PRs flow through full CI; plan 01-04's `verify-deps-allowlist` will block any non-x/text runtime dep addition.
- **T-01-02-05 (Tampering — compromised third-party action):** All actions pinned to specific majors. Dependabot watches `github-actions` daily. SHA pinning is a follow-up if mask-style strict pinning is later mandated.
- **T-01-02-06 (Repudiation — local-CI drift):** `make check` is identical locally and in CI. Plan 01-08 documents the contract in CONTRIBUTING.

## Next Plan Readiness

- **Plan 01-03 (release pipeline)** inherits the four CI workflows and adds `commitlint.yml`, `cla.yml`, `release.yml`. The `.golangci.yml`, Makefile, and dependabot.yml are already in place — 01-03 only adds the release-discipline workflows (conventional-commit linting, CLA Assistant, goreleaser, cosign, SBOM, OIDC attestation).
- **Plan 01-04 (determinism infra)** inherits the ci.yml ubuntu-latest baseline and expands the matrix to linux/amd64 + linux/arm64 + darwin/amd64 + darwin/arm64 + windows/amd64. It also lands `scripts/verify-no-runtime-deps.sh` (which `make verify-deps-allowlist` already invokes if present) and the determinism golden test (which `make verify-determinism` already invokes via `go test -run TestGolden_`).
- **Plan 01-05 onwards** inherits the full `make check` gate. Every new .go file MUST carry the Apache-2.0 header; every PR runs through lint + vet + race + vuln + license-headers automatically.

## Self-Check: PASSED

Files claimed to exist (verified with `[ -f ... ]`):
- `.golangci.yml` — FOUND
- `.markdownlint-cli2.yaml` — FOUND
- `Makefile` — FOUND
- `.github/workflows/ci.yml` — FOUND
- `.github/workflows/security.yml` — FOUND
- `.github/workflows/codeql.yml` — FOUND
- `.github/workflows/license-headers.yml` — FOUND
- `.github/dependabot.yml` — FOUND

Commits claimed to exist (verified with `git log --oneline`):
- `972bef1` — FOUND (Task 1: golangci-lint + markdownlint configs)
- `f7fe3b2` — FOUND (Task 2: Makefile + tests/bdd/doc.go fmt fix)
- `9317d76` — FOUND (Task 3: workflows + dependabot)

Plan-level verification (from `<verification>` section):
- `make check` exits 0 on a clean tree — PASS
- `make fmt-check` exits 0 — PASS
- `make vet` exits 0 — PASS
- `make lint` exits 0 (golangci-lint clean on root and tests/bdd) — PASS
- `make tidy-check` exits 0 — PASS
- `make verify-license-headers` exits 0 — PASS
- Each `.yml` workflow file parses as valid YAML — PASS
- `actionlint .github/workflows/*.yml` exits 0 (no errors) — PASS
- Each workflow references actions at STACK.md-mandated versions — PASS
- `dependabot.yml` watches gomod (root + tests/bdd) and github-actions with grouped PRs — PASS
- No workflow uses `permissions: write-all` (`grep -rn 'write-all' .github/workflows/` returns nothing) — PASS

---
*Phase: 01-foundation-infrastructure*
*Plan: 02 (quality-gates)*
*Completed: 2026-05-13*
