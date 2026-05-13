---
plan_id: 01-02
phase: 01-foundation-infrastructure
plan: 02
type: execute
wave: 2
depends_on:
  - 01-01
autonomous: true
objective: >
  Land the full local + CI quality apparatus: golangci-lint v2 config,
  Makefile, markdownlint config, ci/security/codeql workflows, dependabot,
  license-header CI gate, tidy-check, and coverage enforcement. After this
  plan, `make check` is the canonical local "is my branch shippable" command,
  and CI reproduces it on every PR. Conventional-commit lint and CLA Assistant
  workflows are owned by plan 01-03 (release pipeline) where they fit
  topically.
files_modified:
  - .golangci.yml
  - Makefile
  - .markdownlint-cli2.yaml
  - .github/workflows/ci.yml
  - .github/workflows/security.yml
  - .github/workflows/codeql.yml
  - .github/workflows/license-headers.yml
  - .github/dependabot.yml
requirements:
  - CI-01
  - CI-02
  - CI-03
  - CI-04
  - CI-07
  - CI-08
  - CI-11
  - TEST-08
must_haves:
  truths:
    - "`make check` exits 0 on a clean tree and is the local pre-PR gate"
    - "`.golangci.yml` declares `version: \"2\"` on the first non-comment line"
    - "`make tidy-check` fails CI on any go.mod / go.sum drift"
    - "`make verify-deps-allowlist` is wired as a Makefile target placeholder; the underlying script lands in plan 01-04"
    - "govulncheck, gosec, codeql, markdownlint, license-header verification all run in CI"
    - "Dependabot watches gomod (root), gomod (tests/bdd), and github-actions"
    - "tests/bdd will use goleak in TestMain to catch goroutine leaks (TEST-08 — the BDD test harness lands when the first BDD feature lands; until then the goleak convention is documented in this plan and enforced when scaffold lands)"
    - "No CI workflow uses `permissions: write-all`; every workflow follows least-privilege"
  artifacts:
    - path: .golangci.yml
      provides: golangci-lint v2 configuration with formatters block and curated linter set
      contains: 'version: "2"'
    - path: Makefile
      provides: 18 canonical local quality targets per CLAUDE.md
      contains: "^check:"
    - path: .github/workflows/ci.yml
      provides: Lint+vet+test+race+coverage job (matrix expanded in plan 01-04)
      contains: "golangci-lint"
    - path: .github/workflows/security.yml
      provides: govulncheck + gosec scheduled weekly + on PR
      contains: "govulncheck"
    - path: .github/workflows/codeql.yml
      provides: CodeQL Go analysis
      contains: "github/codeql-action"
    - path: .github/workflows/license-headers.yml
      provides: CI gate invoking scripts/verify-license-headers.sh from plan 01-01
      contains: "verify-license-headers.sh"
    - path: .github/dependabot.yml
      provides: Daily dependency update PRs for gomod (root, tests/bdd) and github-actions
      contains: "package-ecosystem"
    - path: .markdownlint-cli2.yaml
      provides: Markdown lint configuration
  key_links:
    - from: .github/workflows/ci.yml
      to: Makefile
      via: workflow invokes `make check`
      pattern: "make.*check"
    - from: .github/workflows/license-headers.yml
      to: scripts/verify-license-headers.sh
      via: workflow runs the script
      pattern: "verify-license-headers.sh"
    - from: Makefile
      to: golangci-lint
      via: `lint` target invokes golangci-lint
      pattern: "golangci-lint run"
---

<objective>
Wire the full quality apparatus so every PR is gated by lint, vet, race, vuln,
security, codeql, markdownlint, tidy-check, license-header check, and coverage
on every commit.

Purpose: Catch correctness and hygiene regressions automatically before they
land. `make check` is the single local entry point; CI mirrors it.

Output:
  - .golangci.yml (v2 schema)
  - Makefile (18 targets including `check`)
  - .markdownlint-cli2.yaml
  - .github/workflows/{ci,security,codeql,license-headers}.yml
  - .github/dependabot.yml

Out of scope (plan 01-03):
  - .github/workflows/commitlint.yml + .commitlintrc.yml (CI-09 — plan 01-03)
  - .github/workflows/cla.yml (CI-10 — plan 01-03)
  - .github/workflows/release.yml (REL-* — plan 01-03)
</objective>

<execution_context>
@$HOME/.claude/get-shit-done/workflows/execute-plan.md
@$HOME/.claude/get-shit-done/templates/summary.md
</execution_context>

<context>
@.planning/PROJECT.md
@.planning/ROADMAP.md
@.planning/REQUIREMENTS.md
@.planning/phases/01-foundation-infrastructure/01-CONTEXT.md
@.planning/research/STACK.md
@CLAUDE.md
@.claude/skills/go-coding-standards/SKILL.md
@.claude/skills/go-testing-standards/SKILL.md
@.claude/skills/determinism-standards/SKILL.md
@scripts/verify-license-headers.sh
@go.mod
@tests/bdd/go.mod

<interfaces>
Tooling version pins from STACK.md and CONTEXT.md:

| Tool                          | Pinned version            | Where used               |
|-------------------------------|---------------------------|--------------------------|
| Go                            | 1.26.3                    | actions/setup-go go-version-file: go.mod |
| golangci-lint                 | v2.12.2                   | .golangci.yml schema v2 + `golangci/golangci-lint-action@v8` |
| govulncheck                   | latest stable             | `go install golang.org/x/vuln/cmd/govulncheck@latest` |
| gosec                         | v2.25.0                   | `securego/gosec@v2.25.0` Action |
| CodeQL action                 | github/codeql-action@v4   | init + autobuild + analyze |
| markdownlint-cli2             | v0.22.1                   | `DavidAnson/markdownlint-cli2-action@latest-stable` |
| actions/checkout              | v6                        | every workflow |
| actions/setup-go              | v6                        | every workflow |

Makefile target taxonomy (canonical names per CLAUDE.md "Makefile Targets"):
  check, test, test-bdd, test-fuzz, lint, vet, fmt, fmt-check, bench,
  bench-compare, coverage, tidy, tidy-check, security, verify-deps-allowlist,
  verify-determinism, verify-license-headers, release-check, clean

`make check` is the aggregate: fmt-check → vet → lint → verify-license-headers
  → verify-deps-allowlist → tidy-check → security → test → coverage → coverage-check

`verify-deps-allowlist`, `verify-determinism`, `bench`, `bench-compare`
targets land in this Makefile as Make rules NOW, but the underlying scripts
(scripts/verify-no-runtime-deps.sh, golden harness, bench.txt) are created
in plan 01-04. Until then those targets tolerate the missing scripts and
print a "pending plan 01-04" message, exiting 0.

`release-check` invokes `goreleaser check` against .goreleaser.yml from plan
01-03 — until that file exists, the target prints a "pending plan 01-03"
message and exits 0.
</interfaces>
</context>

<tasks>

<task type="auto">
  <name>Task 1: Write .golangci.yml (v2 schema) and .markdownlint-cli2.yaml</name>
  <files>.golangci.yml, .markdownlint-cli2.yaml</files>
  <read_first>
    - .planning/research/STACK.md (golangci-lint v2.12.2 row, markdownlint-cli2 v0.22.1 row; alternatives table)
    - .planning/phases/01-foundation-infrastructure/01-CONTEXT.md (D-10 — CI gates)
    - .claude/skills/go-coding-standards/SKILL.md (Quality Gate section — lint clean, vet clean, race clean, tidy clean, vuln clean; misspell locale UK per documentation-standards)
    - .claude/skills/documentation-standards/SKILL.md (British English for prose)
    - https://golangci-lint.run/docs/configuration/file/ (v2 config schema reference)
  </read_first>
  <action>
    Write the golangci-lint v2 configuration and markdownlint-cli2 configuration.

    1. `.golangci.yml` — v2 schema. First non-comment line MUST be
       `version: "2"`.
       Structure:
       - `version: "2"`
       - `run:`
           `timeout: 5m`
           `tests: true`
       - `linters:` block:
           `default: standard` (golangci-lint v2 vocabulary — enables the
           curated default-recommended set)
           `enable:` explicit list adding to defaults: errcheck, govet,
            staticcheck, revive, gocyclo, ineffassign, unused, unparam,
            misspell, gocritic, bodyclose, errorlint, exhaustive, gosec
           `disable:` empty (or omit)
           `settings:` (or `linters-settings:` per v2 layout):
             `gocyclo: { min-complexity: 10 }` (matches go-coding-standards)
             `revive: { confidence: 0.8 }`
             `misspell: { locale: "UK" }` (British English per documentation-standards)
             `gosec: { severity: low, confidence: low }`  (informational in
               lint; security.yml runs gosec definitively)
       - `formatters:` block (v2-specific — split from linters):
           `enable: [gofmt, gofumpt, goimports]`
           `settings:`
             `gofmt: { simplify: true }`
             `goimports: { local-prefixes: github.com/axonops/fuzzymatch }`
       - `issues:` block:
           `exclude-rules:`
             - path: `_test\.go`
               linters: [gocyclo, exhaustive, gosec]
           `max-issues-per-linter: 0`  (no cap — report everything)
           `max-same-issues: 0`

    2. `.markdownlint-cli2.yaml`:
       - `config:` map with rules:
           `MD013: false` (line-length — long URLs and code lines)
           `MD033: false` (allow inline HTML for badges)
           `MD041: false` (first line can be a badge/frontmatter)
       - `globs: ["**/*.md"]`
       - `ignores:
           - "node_modules/**"
           - "vendor/**"
           - ".planning/**"
           - ".claude/**"
           - "tests/bdd/**/vendor/**"`

    Concrete identifiers:
      - File `.golangci.yml`, schema marker `version: "2"`.
      - File `.markdownlint-cli2.yaml`.
      - Enabled linters: errcheck, govet, staticcheck, revive, gocyclo,
        ineffassign, unused, unparam, misspell, gocritic, bodyclose, errorlint,
        exhaustive, gosec.
      - misspell locale: UK.
      - gocyclo min-complexity: 10.

    Note (per CLAUDE.md Design Principle 13): the exact enabled-linter set
    may be adjusted by go-quality / code-reviewer at execution time if a
    specific linter produces excessive false positives. The structure (v2
    schema, formatters block split, misspell UK, gocyclo 10) is locked.
  </action>
  <verify>
    <automated>head -n 5 .golangci.yml | grep -qE '^version:\s*"?2"?' &amp;&amp; python3 -c "import yaml; yaml.safe_load(open('.golangci.yml'))" &amp;&amp; test -f .markdownlint-cli2.yaml &amp;&amp; python3 -c "import yaml; yaml.safe_load(open('.markdownlint-cli2.yaml'))" &amp;&amp; { command -v golangci-lint &amp;&amp; golangci-lint run --no-config-cache ./...; } || echo "golangci-lint not installed locally; rely on CI"</automated>
  </verify>
  <acceptance_criteria>
    - `.golangci.yml` exists and is valid YAML
    - First non-comment line is exactly `version: "2"`
    - Contains a `linters:` top-level block
    - Contains a `formatters:` top-level block (v2-specific)
    - Enables at minimum: errcheck, govet, staticcheck, revive, gocyclo, ineffassign, unused
    - misspell locale set to UK
    - gocyclo min-complexity set to 10
    - `golangci-lint run ./...` exits 0 on the current tree (only doc.go exists — clean)
    - `.markdownlint-cli2.yaml` exists and is valid YAML
    - `.markdownlint-cli2.yaml` excludes `.planning/**` and `.claude/**`
    - Running `markdownlint-cli2 "**/*.md"` exits 0 on the current tree
  </acceptance_criteria>
  <done>
    golangci-lint v2 schema is committed and passes; markdownlint config
    excludes planning artifacts and passes on current docs.
  </done>
</task>

<task type="auto">
  <name>Task 2: Write the canonical Makefile</name>
  <files>Makefile</files>
  <read_first>
    - CLAUDE.md (project root — "Makefile Targets" section)
    - .planning/phases/01-foundation-infrastructure/01-CONTEXT.md (D-09 — bench targets local-driven; verify-deps-allowlist invokes script from plan 01-04)
    - .claude/skills/go-testing-standards/SKILL.md ("Running Tests" — `go test -race -shuffle=on -v -count=1 ./...`)
    - .claude/skills/go-coding-standards/SKILL.md (Quality Gate section)
    - .golangci.yml (just created)
    - scripts/verify-license-headers.sh (from plan 01-01)
    - go.mod, tests/bdd/go.mod (from plan 01-01)
  </read_first>
  <action>
    Create the Makefile with the 18 canonical targets from CLAUDE.md. Use
    `.PHONY` for every target.

    Canonical target list (exact names):
      check, test, test-bdd, test-fuzz, lint, vet, fmt, fmt-check,
      bench, bench-compare, coverage, coverage-check, tidy, tidy-check,
      security, verify-deps-allowlist, verify-determinism,
      verify-license-headers, release-check, clean

    Target definitions:
      - `test` — `go test -race -shuffle=on -count=1 ./...`
      - `test-bdd` — `cd tests/bdd && go test -race -count=1 ./...`
      - `test-fuzz` — runs `go test -fuzz=. -fuzztime=60s ./...` if any fuzzer
         is registered; otherwise prints "no fuzzers found" and exits 0
      - `lint` — `golangci-lint run ./...` AND `cd tests/bdd && golangci-lint run ./...`
      - `vet` — `go vet ./...` AND `cd tests/bdd && go vet ./...`
      - `fmt` — `gofmt -s -w .` AND `goimports -local github.com/axonops/fuzzymatch -w .`
      - `fmt-check` — asserts no diff after `gofmt -s -d .` AND
         `goimports -local github.com/axonops/fuzzymatch -d .`
      - `bench` — `go test -bench=. -benchmem -count=10 ./... | tee bench.txt.new`
         (no-op-tolerant if no benchmarks exist yet — print message exit 0)
      - `bench-compare` — `benchstat bench.txt bench.txt.new` if both exist;
         otherwise print "pending: bench.txt or bench.txt.new not present;
         run `make bench` first" and exit 0 (plan 01-04 lands the real
         comparison harness)
      - `coverage` — `go test -race -coverprofile=coverage.out -covermode=atomic ./...`
      - `coverage-check` — runs `go tool cover -func=coverage.out`, parses
         the `total:` line's percentage, asserts ≥ 95.0%, exits non-zero
         otherwise. Per-file and public-API enforcement is deferred to plan
         01-04 (requires a parser script). Tolerant: if coverage.out absent,
         prints message exit 0 so a fresh-clone `make coverage-check` doesn't
         spuriously fail.
      - `tidy` — `go mod tidy` AND `cd tests/bdd && go mod tidy`
      - `tidy-check` — `go mod tidy && cd tests/bdd && go mod tidy && cd ..
         && git diff --exit-code go.mod go.sum tests/bdd/go.mod tests/bdd/go.sum`
      - `security` — `govulncheck ./...` (security workflow runs gosec
         definitively; the Makefile security target focuses on govulncheck
         for local use). Tolerant if govulncheck not installed: prints
         install hint and exits 0.
      - `verify-deps-allowlist` — `if test -x scripts/verify-no-runtime-deps.sh;
         then bash scripts/verify-no-runtime-deps.sh; else echo "pending plan
         01-04: scripts/verify-no-runtime-deps.sh not yet present"; fi`
      - `verify-determinism` — `go test -run TestGolden_ ./...` (currently
         finds no tests; will exercise plan 01-04's golden harness)
      - `verify-license-headers` — `bash scripts/verify-license-headers.sh`
      - `release-check` — `if test -f .goreleaser.yml; then goreleaser check;
         else echo "pending plan 01-03: .goreleaser.yml not yet present"; fi`
      - `clean` — `go clean ./...; rm -f coverage.out bench.txt.new`

    `check` target chains:
      `fmt-check vet lint verify-license-headers verify-deps-allowlist
       tidy-check security test coverage coverage-check`

    Use POSIX-portable shell where practical; GNU-Make idioms (pattern rules,
    automatic variables) are acceptable since CI runs Linux/macOS with GNU
    make available.

    Concrete identifiers:
      - Makefile path `Makefile`
      - Aggregate target name `check`
      - Coverage floor: `95.0` (parsed from `go tool cover -func=coverage.out`)
      - Coverage profile filename `coverage.out`
      - Bench output filename `bench.txt.new`

    Note (CLAUDE.md Makefile section requires every target to be documented
    in CONTRIBUTING/README or carry a `## suppress: <reason>` comment): plan
    01-08 writes CONTRIBUTING documenting these targets. For now, all 19
    canonical targets are expected to appear in CONTRIBUTING; no suppression
    is needed.
  </action>
  <verify>
    <automated>make -n check &gt; /dev/null &amp;&amp; make fmt-check &amp;&amp; make vet &amp;&amp; make tidy-check &amp;&amp; make verify-license-headers &amp;&amp; make lint &amp;&amp; make test &amp;&amp; make verify-deps-allowlist &amp;&amp; make verify-determinism &amp;&amp; make release-check &amp;&amp; make clean</automated>
  </verify>
  <acceptance_criteria>
    - `Makefile` exists at the repo root
    - `grep -c '^\.PHONY:' Makefile` is at least 1
    - `make -n check` exits 0
    - `make fmt-check` exits 0
    - `make vet` exits 0
    - `make tidy-check` exits 0
    - `make verify-license-headers` exits 0
    - `make lint` exits 0
    - `make test` exits 0 (no test files yet — Go reports "no test files" which is exit 0)
    - `make verify-deps-allowlist` exits 0 (prints pending message)
    - `make verify-determinism` exits 0 (no TestGolden_ tests yet)
    - `make release-check` exits 0 (prints pending-plan-01-03 message until .goreleaser.yml exists)
    - Each of the 19 canonical target names from CLAUDE.md appears in `grep -E '^[a-z][a-z-]*:' Makefile`
  </acceptance_criteria>
  <done>
    Makefile defines all 19 canonical targets; `make check` is the aggregate
    pre-PR gate; all "pending" targets are no-op-safe until later plans
    fill in their underlying scripts.
  </done>
</task>

<task type="auto">
  <name>Task 3: Write CI workflows (ci, security, codeql, license-headers) + Dependabot</name>
  <files>.github/workflows/ci.yml, .github/workflows/security.yml, .github/workflows/codeql.yml, .github/workflows/license-headers.yml, .github/dependabot.yml</files>
  <read_first>
    - .planning/research/STACK.md (Build & Quality Tooling table — action version pins)
    - .planning/phases/01-foundation-infrastructure/01-CONTEXT.md (D-09 — github-hosted ubuntu-latest only; D-10 — mask-style branch protection)
    - CLAUDE.md ("Workflow — Agent Gates")
    - Makefile (just created — `make check` is the canonical entrypoint)
    - .golangci.yml (just created)
    - .markdownlint-cli2.yaml (just created)
    - scripts/verify-license-headers.sh (plan 01-01)
    - tests/bdd/go.mod (plan 01-01)
    - go.mod (plan 01-01)
  </read_first>
  <action>
    Write the four CI workflows and dependabot config. All workflows use:
      - `actions/checkout@v6`
      - `actions/setup-go@v6` with `go-version-file: go.mod`
      - `permissions:` block at workflow level defaulting `contents: read`,
        escalating only what jobs need (least-privilege)

    Note: cross-platform matrix is NOT added here — plan 01-04 owns the
    determinism matrix expansion. This plan's ci.yml runs on `ubuntu-latest`
    only; plan 01-04 expands the matrix.

    1. `.github/workflows/ci.yml`:
       - Name: `CI`
       - Triggers: `pull_request`, `push` to main
       - Workflow-level `permissions: contents: read`
       - Job `quality`:
           - runs-on: ubuntu-latest
           - steps:
             a. `actions/checkout@v6` with `fetch-depth: 0`
             b. `actions/setup-go@v6` with `go-version-file: go.mod`
             c. Install tools (pinned where mandated by STACK.md):
                 - `go install github.com/golangci/golangci-lint/cmd/golangci-lint@v2.12.2`
                 - `go install golang.org/x/vuln/cmd/govulncheck@latest`
                 - `go install golang.org/x/tools/cmd/goimports@latest`
             d. `make check`
       - Job `markdownlint`:
           - runs-on: ubuntu-latest
           - Uses `DavidAnson/markdownlint-cli2-action@latest-stable`
           - globs: `**/*.md`
           - config: `.markdownlint-cli2.yaml`

    2. `.github/workflows/security.yml`:
       - Name: `Security`
       - Triggers: `pull_request`, `push` to main, `schedule: { cron: '0 6 * * 1' }`
       - Workflow-level `permissions: contents: read`
       - Job `govulncheck`:
           - runs-on: ubuntu-latest
           - permissions: `contents: read`
           - install govulncheck, run `govulncheck ./...`, fail on HIGH/CRITICAL
       - Job `gosec`:
           - runs-on: ubuntu-latest
           - permissions: `contents: read`, `security-events: write` (SARIF upload)
           - Uses `securego/gosec@v2.25.0`
           - args: `-no-fail -fmt sarif -out gosec.sarif ./...`
           - Upload SARIF: `github/codeql-action/upload-sarif@v4` with
             `sarif_file: gosec.sarif`

    3. `.github/workflows/codeql.yml`:
       - Name: `CodeQL`
       - Triggers: `pull_request`, `push` to main, `schedule: { cron: '0 6 * * 2' }`
       - Workflow-level `permissions: contents: read`
       - Job `analyze`:
           - runs-on: ubuntu-latest
           - permissions: `actions: read`, `contents: read`, `security-events: write`
           - matrix: `language: ['go']`
           - steps:
             a. `actions/checkout@v6`
             b. `github/codeql-action/init@v4` with `languages: go`, `build-mode: autobuild`
             c. `github/codeql-action/autobuild@v4`
             d. `github/codeql-action/analyze@v4`

    4. `.github/workflows/license-headers.yml`:
       - Name: `License Headers`
       - Triggers: `pull_request`, `push` to main
       - Workflow-level `permissions: contents: read`
       - Job `verify`:
           - runs-on: ubuntu-latest
           - steps:
             a. `actions/checkout@v6`
             b. `bash scripts/verify-license-headers.sh`

    5. `.github/dependabot.yml`:
       - `version: 2`
       - `updates:`
         - package-ecosystem: `gomod`, directory: `/`, schedule: `daily`,
           open-pull-requests-limit: 5, commit-message.prefix: `chore`,
           groups:
             `indirect: { dependency-type: indirect }`,
             `direct: { dependency-type: direct }`
         - package-ecosystem: `gomod`, directory: `/tests/bdd`, schedule:
           `daily`, open-pull-requests-limit: 5, labels: `[test-only]`,
           commit-message.prefix: `chore(test-bdd)`, groups: same shape as
           root
         - package-ecosystem: `github-actions`, directory: `/`, schedule:
           `daily`, open-pull-requests-limit: 5, commit-message.prefix: `chore(ci)`,
           groups:
             `actions: { patterns: ["*"] }`

    Concrete identifiers:
      - Action versions per STACK.md: `actions/checkout@v6`, `actions/setup-go@v6`,
        `github/codeql-action@v4`, `securego/gosec@v2.25.0`,
        `DavidAnson/markdownlint-cli2-action@latest-stable`
      - golangci-lint binary pinned: `v2.12.2`
      - Cron schedule for security: `0 6 * * 1` (Mondays 06:00 UTC)
      - Cron schedule for codeql: `0 6 * * 2` (Tuesdays 06:00 UTC)
      - Workflow permissions baseline: `contents: read`
  </action>
  <verify>
    <automated>test -f .github/workflows/ci.yml &amp;&amp; test -f .github/workflows/security.yml &amp;&amp; test -f .github/workflows/codeql.yml &amp;&amp; test -f .github/workflows/license-headers.yml &amp;&amp; test -f .github/dependabot.yml &amp;&amp; for f in .github/workflows/ci.yml .github/workflows/security.yml .github/workflows/codeql.yml .github/workflows/license-headers.yml .github/dependabot.yml; do python3 -c "import yaml; yaml.safe_load(open('$f'))" || exit 1; done &amp;&amp; grep -q 'actions/checkout@v6' .github/workflows/ci.yml &amp;&amp; grep -q 'go-version-file:\s*go.mod' .github/workflows/ci.yml &amp;&amp; grep -q 'verify-license-headers.sh' .github/workflows/license-headers.yml &amp;&amp; grep -q 'github/codeql-action' .github/workflows/codeql.yml &amp;&amp; ! grep -rq 'write-all' .github/workflows/</automated>
  </verify>
  <acceptance_criteria>
    - All five files exist and are valid YAML
    - `ci.yml` references `actions/checkout@v6` and `actions/setup-go@v6`
    - `ci.yml` uses `go-version-file: go.mod` (single source of Go-version truth)
    - `ci.yml` invokes `make check`
    - `ci.yml` includes a markdownlint job using `DavidAnson/markdownlint-cli2-action@latest-stable`
    - `security.yml` has a `schedule` block with cron `'0 6 * * 1'`
    - `security.yml` job `gosec` uses `securego/gosec@v2.25.0`
    - `security.yml` job `gosec` has job-level `security-events: write`
    - `security.yml` job `govulncheck` runs `govulncheck ./...`
    - `codeql.yml` uses `github/codeql-action@v4` (not v3)
    - `codeql.yml` has `build-mode: autobuild` and `languages: go`
    - `license-headers.yml` invokes `scripts/verify-license-headers.sh`
    - `dependabot.yml` declares `version: 2`
    - `dependabot.yml` watches three ecosystems: gomod at `/`, gomod at `/tests/bdd`, github-actions at `/`
    - Each gomod ecosystem entry has a `groups:` block with at least `indirect` and `direct`
    - No workflow uses `permissions: write-all` (grep finds no occurrences)
    - `actionlint` (if installed) reports no errors on any workflow file
  </acceptance_criteria>
  <done>
    Four CI workflows + dependabot config exist, parse as valid YAML, pin
    actions at the STACK.md-mandated versions, and run on PR + scheduled
    cadences. Matrix expansion to 5 platforms is deferred to plan 01-04.
  </done>
</task>

</tasks>

<threat_model>
## Trust Boundaries

| Boundary | Description |
|----------|-------------|
| Pull request author → CI | Untrusted PR code runs in CI; workflows must use least-privilege permissions. |
| Third-party GitHub Action → repository | Each action runs with the permissions the workflow grants; pinning to major versions (v4, v6) balances security patches and reproducibility. |
| Dependabot bot → repository | Dependabot raises PRs mutating go.mod / go.sum; verify-deps-allowlist (plan 01-04) blocks any PR introducing a non-x/text runtime dep, even from Dependabot. |

## STRIDE Threat Register

| Threat ID | Category | Component | Disposition | Mitigation Plan |
|-----------|----------|-----------|-------------|-----------------|
| T-01-02-01 | Elevation of Privilege | CI workflow permissions | mitigate | Workflow-level `permissions: contents: read`; specific jobs escalate only what they need. No workflow uses `permissions: write-all`. |
| T-01-02-02 | Tampering | Linter / lint config drift | mitigate | `.golangci.yml` v2 schema enforced by `make lint` in `make check`; CI runs `make check`; PRs with lint violations cannot merge under mask-style branch protection (manual configuration). |
| T-01-02-03 | Information Disclosure | Vulnerability findings | mitigate | govulncheck runs on every PR + weekly; gosec uploads SARIF to GitHub Security tab. Both fail CI on HIGH/CRITICAL. |
| T-01-02-04 | Spoofing | Dependabot PRs with malicious dep updates | mitigate | All Dependabot PRs go through full CI including verify-deps-allowlist (plan 01-04 wires this). A malicious dep that tries to add itself to root go.mod is rejected by the allowlist gate. |
| T-01-02-05 | Tampering | Third-party action compromised | mitigate | All actions pinned to specific majors; dependabot watches `github-actions` ecosystem. SHA pinning is a follow-up if mask-style strict pinning is later mandated. |
| T-01-02-06 | Repudiation | Lint findings missed locally before PR | mitigate | `make check` is the local entrypoint; CONTRIBUTING (plan 01-08) instructs contributors. CI re-runs `make check` independently. |
</threat_model>

<verification>
1. `make check` exits 0 on a clean tree.
2. `make fmt-check`, `make vet`, `make lint`, `make tidy-check`,
   `make verify-license-headers` all exit 0.
3. Each .yml workflow file is valid YAML.
4. `actionlint` (if installed) reports no errors on any workflow file.
5. Each workflow references actions at STACK.md-mandated versions.
6. dependabot.yml watches gomod (root + tests/bdd) and github-actions.
7. No workflow uses `permissions: write-all`.
</verification>

<success_criteria>
- `.golangci.yml` is v2 schema, passes on the current tree.
- `Makefile` defines all 19 canonical targets from CLAUDE.md.
- `make check` is the aggregate pre-PR gate.
- Four GitHub Actions workflows: ci, security, codeql, license-headers.
- Dependabot watches three ecosystems with grouped PRs.
- `.markdownlint-cli2.yaml` excludes planning artifacts.
- Every workflow follows least-privilege permissions.
</success_criteria>

<output>
After completion, create
`.planning/phases/01-foundation-infrastructure/01-02-quality-gates-SUMMARY.md`
listing the linters enabled, the Makefile target inventory, the chosen
action versions (recording any STACK.md drift), and the dependabot group
configuration.
</output>
