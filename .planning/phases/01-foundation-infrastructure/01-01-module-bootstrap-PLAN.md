---
plan_id: 01-01
phase: 01-foundation-infrastructure
plan: 01
type: execute
wave: 1
depends_on: []
autonomous: true
objective: >
  Bootstrap the Go module with Go 1.26.3, a single permitted runtime dep
  (golang.org/x/text), Apache-2.0 licensing apparatus, file-header verifier,
  the isolated tests/bdd sub-module (godog + goleak + testify with a replace
  directive), and the empty directory scaffolding the rest of Phase 1 fills in.
  This plan is the load-bearing freeze point for the root go.mod: after this
  plan, no subsequent plan may modify the root require block.
files_modified:
  - go.mod
  - go.sum
  - doc.go
  - tests/bdd/go.mod
  - tests/bdd/go.sum
  - tests/bdd/doc.go
  - tests/bdd/features/.gitkeep
  - tests/bdd/steps/.gitkeep
  - testdata/golden/.gitkeep
  - testdata/fuzz/.gitkeep
  - scripts/verify-license-headers.sh
requirements:
  - FOUND-01
must_haves:
  truths:
    - "Root go.mod declares Go 1.26.3 and exactly one require line: golang.org/x/text"
    - "tests/bdd/go.mod is an isolated sub-module containing godog, goleak, testify with a replace directive pointing at ../.."
    - "Every existing and future .go file carries the AxonOps Apache-2.0 license header"
    - "scripts/verify-license-headers.sh exits 0 on a clean tree and non-zero if any .go file is missing the header"
    - "doc.go provides a package-level godoc comment for the root fuzzymatch package"
  artifacts:
    - path: go.mod
      provides: Module declaration, Go version, golang.org/x/text dependency
      contains: "go 1.26.3"
    - path: tests/bdd/go.mod
      provides: Isolated BDD sub-module with godog/goleak/testify
      contains: "replace github.com/axonops/fuzzymatch =>"
    - path: scripts/verify-license-headers.sh
      provides: CI gate ensuring every .go file has the Apache-2.0 header
    - path: doc.go
      provides: Root package godoc
  key_links:
    - from: tests/bdd/go.mod
      to: ../..
      via: replace directive
      pattern: 'replace github\.com/axonops/fuzzymatch => \.\./\.\.'
    - from: scripts/verify-license-headers.sh
      to: every *.go file
      via: header pattern match
      pattern: "Licensed under the Apache License, Version 2.0"
---

<objective>
Bootstrap the fuzzymatch Go module so that subsequent plans can build/test/lint
against a frozen, single-runtime-dep root module and an isolated BDD sub-module.

Purpose: Establish the single source of truth for the module identity, the
runtime-dep allowlist, the file-header convention, and the BDD module
boundary. Every subsequent plan inherits these decisions.

Output:
  - go.mod / go.sum at repo root with `go 1.26.3` and exactly one require
    (`golang.org/x/text`); no toolchain line.
  - doc.go containing the root package godoc.
  - tests/bdd/go.mod / go.sum with godog v0.15.0, goleak v1.3.0, testify v1.10.0
    and a `replace github.com/axonops/fuzzymatch => ../..` directive.
  - scripts/verify-license-headers.sh — idempotent bash gate that fails if any
    `.go` file under the working tree is missing the Apache-2.0 header.
  - Empty directory scaffolding (with .gitkeep) for testdata/golden,
    testdata/fuzz, tests/bdd/features, tests/bdd/steps.
</objective>

<execution_context>
@$HOME/.claude/get-shit-done/workflows/execute-plan.md
@$HOME/.claude/get-shit-done/templates/summary.md
</execution_context>

<context>
@.planning/PROJECT.md
@.planning/ROADMAP.md
@.planning/STATE.md
@.planning/REQUIREMENTS.md
@.planning/phases/01-foundation-infrastructure/01-CONTEXT.md
@.planning/research/STACK.md
@CLAUDE.md
@.claude/skills/go-coding-standards/SKILL.md
@.claude/skills/algorithm-licensing-standards/SKILL.md
@.claude/skills/commit-standards/SKILL.md

<interfaces>
No code interfaces yet — this plan creates the module skeleton. The only
"interface" is the module identity itself:

- Module path:  github.com/axonops/fuzzymatch
- Go directive: go 1.26.3   (NO `toolchain` line — leave it implicit)
- Sole runtime require: golang.org/x/text  (latest stable; executor verifies
  via `go list -m -versions golang.org/x/text` and pins the highest stable
  semver tag — at planning time this is v0.30.0; if newer is stable, use that)

BDD sub-module:
- Module path:  github.com/axonops/fuzzymatch/tests/bdd
- Go directive: go 1.26.3
- require:      github.com/cucumber/godog v0.15.0
                go.uber.org/goleak       v1.3.0
                github.com/stretchr/testify v1.10.0
                github.com/axonops/fuzzymatch v0.0.0-00010101000000-000000000000
- replace:      github.com/axonops/fuzzymatch => ../..

License-header pattern (single canonical form, used by every .go file in this
repo from this plan forward):

  // Copyright 2026 AxonOps Limited
  //
  // Licensed under the Apache License, Version 2.0 (the "License");
  // you may not use this file except in compliance with the License.
  // You may obtain a copy of the License at
  //
  //     http://www.apache.org/licenses/LICENSE-2.0
  //
  // Unless required by applicable law or agreed to in writing, software
  // distributed under the License is distributed on an "AS IS" BASIS,
  // WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
  // See the License for the specific language governing permissions and
  // limitations under the License.
</interfaces>
</context>

<tasks>

<task type="auto">
  <name>Task 1: Initialise root go.mod with Go 1.26.3 and golang.org/x/text</name>
  <files>go.mod, go.sum, doc.go</files>
  <read_first>
    - .planning/phases/01-foundation-infrastructure/01-CONTEXT.md (D-01, D-07 — go.mod freeze rule, x/text as sole permitted runtime dep)
    - .planning/PROJECT.md (Constraints, Key Decisions)
    - .planning/research/STACK.md (Go 1.26.3 rationale; x/text rationale)
    - .claude/skills/go-coding-standards/SKILL.md (Dependencies section — zero non-stdlib deps except x/text; no testify in root; file-header rule)
    - .claude/skills/documentation-standards/SKILL.md (Godoc — doc.go conventions)
    - LICENSE (Apache-2.0 — already present, do NOT modify)
    - NOTICE (already present, do NOT modify)
  </read_first>
  <action>
    Initialise the root Go module per the locked decisions in D-01 and D-07. Concrete steps:

    1. Run `go mod init github.com/axonops/fuzzymatch` to create go.mod.
    2. Edit go.mod to set the directive line to `go 1.26.3` (per D-09 and STACK.md spec-lock).
       Do NOT add a `toolchain` line — leave it implicit so the directive alone
       governs the minimum.
    3. Add the sole permitted runtime dependency: `golang.org/x/text` at the
       latest stable semver tag. Confirm the latest tag with
       `go list -m -versions golang.org/x/text` and pin to the highest stable
       release (no -rc, no pre-release). The exact version is whatever is
       newest stable at execution time; at planning time the most recent stable
       is v0.30.0 — if a newer stable exists at execution time, use that.
       Run `go get golang.org/x/text@<chosen-version>`.
    4. Run `go mod tidy` and verify:
         - go.mod has exactly ONE `require` block
         - That block contains EXACTLY `golang.org/x/text v0.X.Y` and nothing else
         - No `// indirect` lines slip in (x/text has zero transitive deps)
    5. Create `doc.go` at the repo root with the canonical Apache-2.0 header
       (see <interfaces> for the exact comment block) followed by a package
       godoc:
         - First non-comment line: `// Package fuzzymatch ... `  (the godoc
           description — a 1–3 paragraph overview matching the README's
           "Three layers" framing, citing docs/requirements.md as the
           authoritative spec).
         - Last line: `package fuzzymatch`
       The doc.go file has NO declarations beyond the package clause — it
       exists solely to host the package godoc per documentation-standards
       SKILL.md.
    6. Verify `go build ./...` succeeds (it will succeed with no .go files
       beyond doc.go because doc.go is just a package declaration).
    7. Verify `go vet ./...` succeeds.

    Concrete identifiers in this task: module path `github.com/axonops/fuzzymatch`;
    Go directive `go 1.26.3`; require `golang.org/x/text v0.30.0` (or newer
    stable); root package name `fuzzymatch`; doc file `doc.go`.

    Note (per CLAUDE.md Design Principle 13): exact wording of the package
    godoc prose may be adjusted by user-guide-reviewer at execution time;
    the structure (Apache-2.0 header → blank line → // Package fuzzymatch
    prose → package fuzzymatch) is non-negotiable.
  </action>
  <verify>
    <automated>go build ./... &amp;&amp; go vet ./... &amp;&amp; test "$(grep -cE '^require' go.mod)" = "1" &amp;&amp; grep -qE '^\s*golang\.org/x/text v' go.mod &amp;&amp; ! grep -qE '^\s*[^[:space:]/]+\.[^[:space:]/]+/' go.mod | grep -v 'golang.org/x/text\|^module\|^go\|^require'</automated>
  </verify>
  <acceptance_criteria>
    - `go.mod` contains the literal line `module github.com/axonops/fuzzymatch`
    - `go.mod` contains the literal line `go 1.26.3`
    - `go.mod` has NO `toolchain` line
    - `grep -c '^require' go.mod` returns `1` (single require block)
    - `grep -E '^\s*golang\.org/x/text v' go.mod` matches exactly one line
    - No `require` line other than `golang.org/x/text` appears in go.mod
    - `go list -m all` returns exactly two modules: the root module and `golang.org/x/text`
    - `go build ./...` exits 0
    - `go vet ./...` exits 0
    - `go mod tidy` produces no diff (idempotent)
    - `doc.go` exists at the repo root
    - `doc.go` first line is exactly `// Copyright 2026 AxonOps Limited`
    - `doc.go` contains the literal string `Licensed under the Apache License, Version 2.0`
    - `doc.go` ends with `package fuzzymatch` (after the godoc block)
    - `doc.go` contains a comment line starting with `// Package fuzzymatch`
  </acceptance_criteria>
  <done>
    Root module is initialised; go.mod is frozen at Go 1.26.3 + x/text only;
    doc.go carries the canonical license header and the root package godoc;
    `go build` and `go vet` are green; tidy is idempotent.
  </done>
</task>

<task type="auto">
  <name>Task 2: Create tests/bdd sub-module with isolated test dependencies</name>
  <files>tests/bdd/go.mod, tests/bdd/go.sum, tests/bdd/doc.go, tests/bdd/features/.gitkeep, tests/bdd/steps/.gitkeep</files>
  <read_first>
    - .planning/phases/01-foundation-infrastructure/01-CONTEXT.md (D-06 plan 01 — tests/bdd boundary; D-07 — root go.mod freeze must not be polluted by test deps)
    - .planning/research/STACK.md (godog v0.15.0, goleak v1.3.0, testify v1.10.0 spec-locked version pins)
    - .claude/skills/go-coding-standards/SKILL.md ("Stdlib testing only in root tests")
    - .claude/skills/go-testing-standards/SKILL.md (BDD section — godog + goleak + testify isolated to tests/bdd/)
    - go.mod (just created in Task 1 — confirm the module path used in the replace directive)
  </read_first>
  <action>
    Create the isolated BDD sub-module that hosts godog/goleak/testify so the
    root module stays free of test dependencies (D-07, go-coding-standards).
    Concrete steps:

    1. Create directory `tests/bdd/`.
    2. From that directory, run `go mod init github.com/axonops/fuzzymatch/tests/bdd`.
    3. Edit `tests/bdd/go.mod`:
         - Directive line: `go 1.26.3` (match root)
         - Add the `replace` directive:
             `replace github.com/axonops/fuzzymatch => ../..`
         - Use `go get` to add the three test-only deps at their exact spec-locked versions:
             - `github.com/cucumber/godog@v0.15.0`
             - `go.uber.org/goleak@v1.3.0`
             - `github.com/stretchr/testify@v1.10.0`
         - Also `go get github.com/axonops/fuzzymatch@v0.0.0-00010101000000-000000000000`
           (the placeholder version that the replace directive overrides — this
           is the canonical Go-module pattern for `replace`-overridden local
           deps).
    4. Create `tests/bdd/doc.go` with the canonical Apache-2.0 header and the
       package godoc (`// Package bdd hosts the godog-based BDD scenarios for
       fuzzymatch ...` followed by `package bdd`). Same structure as the root
       doc.go; different package name.
    5. Create the empty subdirectories `tests/bdd/features/` and
       `tests/bdd/steps/` and add `.gitkeep` files so they're committed.
    6. Run `go mod tidy` inside `tests/bdd/` and confirm:
         - Three direct require lines exactly: godog, goleak, testify
         - One placeholder fuzzymatch require line that is overridden by replace
         - Indirect deps (godog pulls cucumber-messages, gherkin, testify pulls
           pmezard/go-difflib, davecgh/go-spew, yaml.v3) are present and pinned
    7. Run `cd tests/bdd && go build ./...` to confirm the sub-module compiles
       cleanly. There are no .go files beyond doc.go yet, so this just
       resolves the module graph.

    Critical guardrail: NONE of these test dependencies appear in the root
    go.mod (verified by Task 4's verifier script). The replace directive
    is the ONLY place where fuzzymatch's root module is referenced from
    tests/bdd.

    Concrete identifiers: sub-module path
    `github.com/axonops/fuzzymatch/tests/bdd`; package name `bdd`; replace target
    `../..`; placeholder version
    `v0.0.0-00010101000000-000000000000`.
  </action>
  <verify>
    <automated>cd tests/bdd &amp;&amp; go build ./... &amp;&amp; go vet ./... &amp;&amp; grep -qE '^replace github\.com/axonops/fuzzymatch =&gt; \.\./\.\.' go.mod &amp;&amp; grep -qE '^\s*github\.com/cucumber/godog v0\.15\.0' go.mod &amp;&amp; grep -qE '^\s*go\.uber\.org/goleak v1\.3\.0' go.mod &amp;&amp; grep -qE '^\s*github\.com/stretchr/testify v1\.10\.0' go.mod</automated>
  </verify>
  <acceptance_criteria>
    - `tests/bdd/go.mod` exists
    - First non-comment line of `tests/bdd/go.mod`: `module github.com/axonops/fuzzymatch/tests/bdd`
    - `tests/bdd/go.mod` contains `go 1.26.3`
    - `tests/bdd/go.mod` contains `replace github.com/axonops/fuzzymatch => ../..` exactly
    - `tests/bdd/go.mod` requires `github.com/cucumber/godog v0.15.0`
    - `tests/bdd/go.mod` requires `go.uber.org/goleak v1.3.0`
    - `tests/bdd/go.mod` requires `github.com/stretchr/testify v1.10.0`
    - `tests/bdd/go.mod` requires `github.com/axonops/fuzzymatch v0.0.0-00010101000000-000000000000`
    - `cd tests/bdd && go build ./...` exits 0
    - `cd tests/bdd && go vet ./...` exits 0
    - `cd tests/bdd && go mod tidy` produces no diff
    - `tests/bdd/doc.go` carries the Apache-2.0 header and declares `package bdd`
    - `tests/bdd/features/.gitkeep` exists
    - `tests/bdd/steps/.gitkeep` exists
    - Root go.mod is UNCHANGED — it still contains exactly one require (`golang.org/x/text`)
  </acceptance_criteria>
  <done>
    BDD sub-module is initialised with godog/goleak/testify at the spec-locked
    versions and a replace directive pointing at the parent module; root go.mod
    remains frozen with only x/text.
  </done>
</task>

<task type="auto">
  <name>Task 3: Create license-header verifier script and scaffold remaining directories</name>
  <files>scripts/verify-license-headers.sh, testdata/golden/.gitkeep, testdata/fuzz/.gitkeep</files>
  <read_first>
    - .planning/phases/01-foundation-infrastructure/01-CONTEXT.md (D-07 — Apache-2.0 file header is a CI gate; the verifier is the gate)
    - .claude/skills/go-coding-standards/SKILL.md (Files section — every .go file gets the AxonOps Apache-2.0 header)
    - .claude/skills/algorithm-licensing-standards/SKILL.md (Apache-2.0 compatibility, fresh-implementation discipline)
    - LICENSE (Apache-2.0 — the verifier checks for the canonical header summary, not the full license text)
    - doc.go (just created — example of the canonical header form)
    - tests/bdd/doc.go (just created — example of the canonical header form)
  </read_first>
  <action>
    Create the license-header verifier and the remaining directory scaffolding.
    The verifier is plan 01-02's CI gate dependency, so it must be idempotent
    and unambiguous now.

    Concrete steps:

    1. Create `scripts/` directory if absent.
    2. Create `scripts/verify-license-headers.sh` with:
         - Shebang `#!/usr/bin/env bash`
         - `set -euo pipefail`
         - A `usage` comment at the top explaining: "Exits 0 if every .go file
           under the repository (excluding vendored or generated paths) carries
           the AxonOps Apache-2.0 header; exits non-zero with a list of
           offending files otherwise."
         - Discovery loop using `git ls-files '*.go'` (so only tracked files
           are checked — untracked scratch files do not cause spurious failures)
           OR a `find` fallback if not in a git repo (defensive — script must
           also work in a freshly-cloned tarball CI checkout).
         - For each file: assert the file contains the exact literal string
           `Licensed under the Apache License, Version 2.0` somewhere within
           the first 25 lines. Use `head -n 25 "$file" | grep -qF '...'`.
         - Collect failures in an array. After the loop, if the array is
           non-empty, print each offending path one-per-line to stderr and
           exit 1; else print "OK: <count> .go files carry the Apache-2.0
           header." to stdout and exit 0.
         - The script must be safe to re-run (no temp files left behind, no
           side effects).
    3. `chmod +x scripts/verify-license-headers.sh`.
    4. Create empty directories with `.gitkeep` placeholders:
         - `testdata/golden/.gitkeep` (canonical-form golden files land here in
           plans 01-04 and 01-06)
         - `testdata/fuzz/.gitkeep` (fuzz corpora land here from Phase 2
           onwards; placeholder ensures the directory is tracked)
    5. Run the verifier against the current tree. It MUST pass — the only
       .go files at this point are `doc.go` and `tests/bdd/doc.go`, both of
       which carry the canonical header from Task 1 and Task 2.

    Concrete identifiers: script path `scripts/verify-license-headers.sh`;
    grep target string `Licensed under the Apache License, Version 2.0`;
    header search depth `head -n 25`; failure exit code `1`; success message
    prefix `OK:`.

    Note: This verifier is invoked from the Makefile target
    `verify-license-headers` in plan 01-02 and the CI workflow in plan 01-02.
    Plan 01-02 wires it; this plan creates the executable.
  </action>
  <verify>
    <automated>test -x scripts/verify-license-headers.sh &amp;&amp; bash scripts/verify-license-headers.sh &amp;&amp; test -f testdata/golden/.gitkeep &amp;&amp; test -f testdata/fuzz/.gitkeep</automated>
  </verify>
  <acceptance_criteria>
    - `scripts/verify-license-headers.sh` exists and is executable
    - `scripts/verify-license-headers.sh` first line is `#!/usr/bin/env bash`
    - `scripts/verify-license-headers.sh` contains `set -euo pipefail`
    - `scripts/verify-license-headers.sh` greps for the literal `Licensed under the Apache License, Version 2.0`
    - Running `bash scripts/verify-license-headers.sh` on the current tree exits 0
    - To prove the failure path: creating a temporary `_scratch.go` with no header, running the script, observing exit code 1, and removing the scratch file (the executor should perform this validation interactively and confirm the failure-path message lists `_scratch.go`)
    - `testdata/golden/.gitkeep` exists
    - `testdata/fuzz/.gitkeep` exists
    - The script does not write to any path outside `/tmp` or stdout/stderr
  </acceptance_criteria>
  <done>
    The license-header verifier is in place, executable, idempotent, and
    passes on the current tree. Empty directory scaffolding for golden files
    and fuzz corpora is committed.
  </done>
</task>

</tasks>

<threat_model>
## Trust Boundaries

| Boundary | Description |
|----------|-------------|
| External Go module resolver → root go.mod | Module downloads cross from upstream proxy.golang.org into the build; pinning specific versions limits supply-chain risk. |
| Future contributor PRs → file-header convention | Contributors may add .go files without the Apache-2.0 header; the verifier script is the enforcing gate (wired into CI in plan 01-02). |
| tests/bdd sub-module → root module | The replace directive resolves the parent module locally; this is intentional and constrained to the test sub-module. |

## STRIDE Threat Register

| Threat ID | Category | Component | Disposition | Mitigation Plan |
|-----------|----------|-----------|-------------|-----------------|
| T-01-01 | Tampering | go.mod runtime allowlist | mitigate | Only `golang.org/x/text` permitted; `make verify-deps-allowlist` (plan 01-04) enforces in CI; this plan establishes the baseline. Any future PR adding a runtime dep fails the CI gate. |
| T-01-02 | Information Disclosure | License-header drift | mitigate | `scripts/verify-license-headers.sh` enforces the AxonOps Apache-2.0 header on every .go file; CI gate (plan 01-02) blocks PRs that omit it. |
| T-01-03 | Tampering | Test-dependency leakage into root module | mitigate | tests/bdd is an isolated sub-module with its own go.mod; root go.mod has zero non-x/text requires; `make verify-deps-allowlist` (plan 01-04) and `make tidy-check` (plan 01-02) verify on every PR. |
| T-01-04 | Repudiation | Module version drift across patches | accept | Go module proxy + go.sum provide cryptographic pinning; `go mod tidy -diff` catches unintended drift. No additional mitigation needed at this layer. |
| T-01-05 | Spoofing | Compromised upstream `golang.org/x/text` release | accept | x/text is maintained by the Go team under their key infrastructure; trust delegated to Go module proxy. Risk lowered by pinning exact semver. |
</threat_model>

<verification>
Phase-level checks for plan 01-01:

1. `go build ./...` and `go vet ./...` both exit 0.
2. `grep -c '^require' go.mod` returns `1` (exactly one require block).
3. `go list -m all` returns exactly two modules.
4. `cd tests/bdd && go build ./...` and `go vet ./...` exit 0.
5. `bash scripts/verify-license-headers.sh` exits 0.
6. Manual sanity: introducing a header-less .go file and re-running the
   verifier produces exit 1 with the offending path on stderr.
</verification>

<success_criteria>
- Root go.mod is frozen at Go 1.26.3 + golang.org/x/text only
- tests/bdd sub-module is isolated with godog v0.15.0, goleak v1.3.0, testify v1.10.0 and a replace directive
- scripts/verify-license-headers.sh exists, is executable, and is idempotent
- All ScaffOlded directories (testdata/golden, testdata/fuzz, tests/bdd/features, tests/bdd/steps) are committed via .gitkeep
- Phase 1's subsequent plans inherit a working, lintable module skeleton
</success_criteria>

<output>
After completion, create `.planning/phases/01-foundation-infrastructure/01-01-module-bootstrap-SUMMARY.md`
summarising what was created, the chosen `golang.org/x/text` version, the
header verifier's exit semantics, and any deviations from this plan (none
expected; record nothing if none).
</output>
