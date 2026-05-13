---
plan_id: 01-03
phase: 01-foundation-infrastructure
plan: 03
type: execute
wave: 3
depends_on:
  - 01-02
autonomous: true
objective: >
  Land the full release apparatus — GoReleaser v2 config, tag-push-only
  release workflow with cosign keyless signing of checksums.txt, Syft SBOM,
  GitHub OIDC build-provenance attestation, conventional-commit lint, and
  CLA Assistant — so the v1.0.0 release in Phase 11 can ship via CI only.
  No local-tagging escape hatch.
files_modified:
  - .goreleaser.yml
  - .github/workflows/release.yml
  - .github/workflows/commitlint.yml
  - .github/workflows/cla.yml
  - .commitlintrc.yml
requirements:
  - REL-01
  - REL-02
  - REL-03
  - REL-04
  - REL-05
  - CI-09
  - CI-10
must_haves:
  truths:
    - "Releases happen exclusively on tag push (`refs/tags/v*`); no `workflow_dispatch` publishes a release"
    - "No local `git tag` or `goreleaser release` path produces a release — verified by absence of any non-tag-triggered goreleaser release step in workflows"
    - "`checksums.txt` is signed by cosign keyless via GitHub OIDC with `--bundle`"
    - "Each release publishes a Syft SPDX-JSON SBOM"
    - "Each release publishes a build-provenance attestation via `actions/attest-build-provenance@v2`"
    - "Conventional-commit lint runs on every PR title and commit"
    - "CLA Assistant runs on every PR (Dependabot + bots allowlisted)"
  artifacts:
    - path: .goreleaser.yml
      provides: GoReleaser v2 library-only release configuration
      contains: 'version: 2'
    - path: .github/workflows/release.yml
      provides: Tag-push-triggered release pipeline with cosign + SBOM + attestation
      contains: "startsWith(github.ref, 'refs/tags/v')"
    - path: .github/workflows/commitlint.yml
      provides: Conventional-commit lint on PRs
      contains: "wagoid/commitlint-github-action"
    - path: .github/workflows/cla.yml
      provides: CLA Assistant workflow
      contains: "contributor-assistant/github-action"
    - path: .commitlintrc.yml
      provides: Conventional-commit ruleset
      contains: "@commitlint/config-conventional"
  key_links:
    - from: .github/workflows/release.yml
      to: .goreleaser.yml
      via: goreleaser-action invokes the config
      pattern: "goreleaser/goreleaser-action"
    - from: .github/workflows/release.yml
      to: cosign + actions/attest-build-provenance
      via: signing + attestation post-build steps
      pattern: "cosign sign-blob.*--bundle"
    - from: .github/workflows/commitlint.yml
      to: .commitlintrc.yml
      via: lint action reads config
      pattern: "wagoid/commitlint-github-action"
---

<objective>
Wire the full v1.0.0-ready release pipeline so signed, attested, SBOM-bundled
releases are tag-push automated and locally unforgeable.

Purpose: Establish CI-only release discipline now (before any release happens),
making it structurally impossible to ship from a workstation. Required for
Phase 11's v1.0.0 freeze.

Output:
  - .goreleaser.yml (v2 schema, library-only)
  - .github/workflows/release.yml (tag-trigger only, cosign + SBOM + attestation)
  - .github/workflows/commitlint.yml
  - .github/workflows/cla.yml
  - .commitlintrc.yml

REL-07 (algorithm deprecation policy) documents in CONTRIBUTING.md (plan
01-08). It is listed in this plan's requirements because the release-policy
constraint (within-major add but not remove; scoring changes require minor
bump) is part of release discipline conceptually — but the documentation lands
in plan 01-08. This plan establishes the technical pipeline; plan 01-08
records the policy text.
</objective>

<execution_context>
@$HOME/.claude/get-shit-done/workflows/execute-plan.md
@$HOME/.claude/get-shit-done/templates/summary.md
</execution_context>

<context>
@.planning/PROJECT.md
@.planning/REQUIREMENTS.md
@.planning/phases/01-foundation-infrastructure/01-CONTEXT.md
@.planning/research/STACK.md
@CLAUDE.md
@.claude/skills/commit-standards/SKILL.md
@.claude/skills/algorithm-licensing-standards/SKILL.md
@Makefile
@.github/workflows/ci.yml
@CHANGELOG.md
@LICENSE
@NOTICE

<interfaces>
Tooling version pins (STACK.md spec-locks):

| Tool                              | Pinned version                    |
|-----------------------------------|-----------------------------------|
| GoReleaser                        | v2.15.4 — pin via `version: "~> v2"` |
| goreleaser/goreleaser-action      | v7 (defaults to `~> v2`)          |
| Cosign                            | v3.0.1 (keyless default; `--bundle` required in v3) |
| sigstore/cosign-installer         | v3 (action major)                 |
| actions/attest-build-provenance   | v2                                |
| Syft                              | invoked by GoReleaser `sboms:` (built-in hook) |
| wagoid/commitlint-github-action   | v6 (latest major)                 |
| contributor-assistant/github-action | v2.6.1 (pin via dependabot)     |
| anchore/sbom-action/download-syft | v0 (Syft binary download helper)  |

Trigger rules (NON-NEGOTIABLE):
- release.yml triggers EXCLUSIVELY on `push.tags: 'v*'`. No `workflow_dispatch`
  block. No `repository_dispatch`. No `schedule`. No `pull_request`.
- Every `goreleaser release` step in release.yml is defensively wrapped in
  `if: startsWith(github.ref, 'refs/tags/v')`.

Permissions baseline for release.yml (least-privilege):
- workflow-level: `contents: read`
- job-level `release`: `contents: write`, `id-token: write`,
  `attestations: write`, `packages: read`

OIDC issuer for cosign:
  `https://token.actions.githubusercontent.com`

For a LIBRARY (no binaries), GoReleaser:
1. Generates `checksums.txt` over the source tarball.
2. Generates a Syft SPDX-JSON SBOM.
3. Publishes the GitHub Release with the extracted CHANGELOG section.
Release.yml then:
4. Cosign keyless-signs `dist/checksums.txt` with `--bundle`.
5. Uploads the bundle to the release.
6. Calls `actions/attest-build-provenance@v2` against the dist artefacts.
7. Verifies the bundle (`cosign verify-blob`) before exit.
</interfaces>
</context>

<tasks>

<task type="auto">
  <name>Task 1: Write .goreleaser.yml (v2 schema, library-only)</name>
  <files>.goreleaser.yml</files>
  <read_first>
    - .planning/research/STACK.md (GoReleaser v2.15.4, `~> v2` pin row; SBOM hook row)
    - .planning/phases/01-foundation-infrastructure/01-CONTEXT.md (D-06 plan 03 — library-only)
    - CHANGELOG.md (Keep-a-Changelog — release notes extraction source)
    - https://goreleaser.com/customization/sbom/ (Syft integration reference)
    - https://goreleaser.com/customization/builds/ (library mode: `builds: skip: true`)
  </read_first>
  <action>
    Write `.goreleaser.yml` for a library release (no binaries). Concrete shape:

    1. First line: `version: 2`.
    2. `project_name: fuzzymatch`.
    3. `before:` hooks: `go mod tidy`, `go mod download`.
    4. `builds:` set to a single entry with `skip: true` (library — no Go
       binary to compile).
    5. `archives:` block producing a source tarball:
       - id: `source`
       - format: `tar.gz`
       - files: whitelist `LICENSE`, `NOTICE`, `README.md`, `CHANGELOG.md`,
         `docs/**/*`, `**/*.go`, `go.mod`, `go.sum`, `Makefile`, `.golangci.yml`
       - Explicitly omit `.planning/`, `.claude/`, `.github/`, `dist/`,
         `vendor/`, `tests/bdd/` from the source archive (these are dev/CI
         artefacts not meant to ship in the source tarball).
       - wrap_in_directory: `fuzzymatch_{{ .Version }}`
    6. `checksum:` block:
       - name_template: `checksums.txt`
       - algorithm: `sha256`
    7. `sboms:` block (Syft, SPDX-JSON):
       - id: `default`
       - artifacts: `archive`
       - cmd: `syft`
       - args: `[ "$artifact", "--output", "spdx-json=$document", "--source-name", "fuzzymatch", "--source-version", "{{ .Version }}" ]`
       - documents: `[ "{{ .ArtifactName }}.sbom.spdx.json" ]`
    8. `changelog:` block:
       - use: `git`
       - sort: `asc`
       - groups (conventional-commit grouping; order ascending):
         - Features (feat), Fixes (fix), Performance (perf), Refactor
           (refactor), Tests (test), Docs (docs), Chore (chore)
       - filters.exclude: `["^Merge ", "^test:", "^chore: bump version"]`
    9. `release:` block:
       - github: `{ owner: axonops, name: fuzzymatch }`
       - draft: false
       - prerelease: auto (matches v0.x pre-1.0 convention)
       - name_template: `v{{ .Version }}`
       - mode: `keep-existing`
       - extra_files: include `dist/checksums.txt.bundle` (the cosign
         signature bundle written by the release workflow's cosign step
         — goreleaser uploads it via extra_files so the release artefact
         set is consistent).
    10. `snapshot:` block:
        - name_template: "{{ .Tag }}-next"

    Verify with `goreleaser check`. This MUST exit 0; the Makefile target
    `release-check` (from plan 01-02) now invokes goreleaser check against
    this file and exits non-zero if it fails — that is, the no-op behaviour
    from plan 01-02 disappears now.

    Concrete identifiers:
      - File `.goreleaser.yml`
      - Schema marker `version: 2`
      - Project name `fuzzymatch`
      - Owner `axonops`
      - Source archive directory template `fuzzymatch_{{ .Version }}`
      - SBOM document filename pattern `{{ .ArtifactName }}.sbom.spdx.json`
      - Checksums filename `checksums.txt`
  </action>
  <verify>
    <automated>head -n 5 .goreleaser.yml | grep -q '^version:\s*2' &amp;&amp; { command -v goreleaser &amp;&amp; goreleaser check; } || echo "goreleaser not installed locally; CI runs goreleaser check"</automated>
  </verify>
  <acceptance_criteria>
    - `.goreleaser.yml` exists
    - First non-comment line is `version: 2`
    - Contains `project_name: fuzzymatch`
    - Contains a `builds:` block with `skip: true` (library mode)
    - Contains an `archives:` block with `format: tar.gz`
    - Contains a `checksum:` block with `name_template: checksums.txt` and `algorithm: sha256`
    - Contains an `sboms:` block invoking `syft` with `spdx-json` output
    - Contains a `changelog:` block with at least 6 conventional-commit groups
    - Contains a `release:` block with `github.owner: axonops` and `github.name: fuzzymatch`
    - `goreleaser check` (if installed) exits 0
    - `make release-check` exits 0
    - `.planning/`, `.claude/`, `.github/`, `tests/bdd/` are excluded from the source archive (visible in the `archives` block file rules)
  </acceptance_criteria>
  <done>
    .goreleaser.yml is a v2 schema, library-only release config that produces
    a source tarball, checksums.txt, and a Syft SPDX-JSON SBOM. `goreleaser
    check` exits 0; `make release-check` no longer prints the pending message.
  </done>
</task>

<task type="auto">
  <name>Task 2: Write release workflow with cosign keyless + SBOM + attestation</name>
  <files>.github/workflows/release.yml</files>
  <read_first>
    - .planning/research/STACK.md (cosign v3.0.1 `--bundle` required; goreleaser-action v7 + `~> v2`; actions/checkout@v6, setup-go@v6; actions/attest-build-provenance@v2; OIDC issuer)
    - .planning/phases/01-foundation-infrastructure/01-CONTEXT.md (D-07 — release exclusively via tag push; D-09 — github-hosted runners only)
    - CLAUDE.md ("Releases — CI Only" section — non-negotiable)
    - .goreleaser.yml (just created — release.yml invokes goreleaser-action which reads this file)
    - https://github.com/goreleaser/example-supply-chain (canonical cosign + SBOM + attest pattern reference)
  </read_first>
  <action>
    Write `.github/workflows/release.yml` triggering EXCLUSIVELY on tag push.

    1. Trigger block: only `on.push.tags: ['v*']`. NO `workflow_dispatch`,
       `repository_dispatch`, `schedule`, or `pull_request`.

    2. Workflow-level permissions: `contents: read`.

    3. Single job `release`:
       - runs-on: `ubuntu-latest`
       - Concurrency: `group: release-${{ github.ref }}`, `cancel-in-progress: false`
       - Job-level permissions (least-privilege escalation):
           `contents: write` — create GitHub Release
           `id-token: write` — OIDC for cosign keyless + attest
           `attestations: write` — actions/attest-build-provenance
           `packages: read`
       - Defensive guard: each step that publishes (goreleaser release,
         cosign sign-blob, attest-build-provenance, gh release upload)
         carries `if: startsWith(github.ref, 'refs/tags/v')`.
       - Steps:

         a. `uses: actions/checkout@v6` with `fetch-depth: 0`

         b. `uses: actions/setup-go@v6` with `go-version-file: go.mod`

         c. `uses: sigstore/cosign-installer@v3` with
            `cosign-release: 'v3.0.1'`

         d. `uses: anchore/sbom-action/download-syft@v0`
            (so GoReleaser can invoke `syft` from PATH per the sboms hook)

         e. `uses: goreleaser/goreleaser-action@v7` with:
              `distribution: goreleaser`
              `version: "~> v2"`
              `args: release --clean`
              `env: GITHUB_TOKEN: ${{ secrets.GITHUB_TOKEN }}`

         f. Cosign keyless sign of checksums.txt (after goreleaser so
            dist/checksums.txt exists):
            ```
            run: |
              cosign sign-blob --yes \
                --bundle dist/checksums.txt.bundle \
                --oidc-issuer https://token.actions.githubusercontent.com \
                dist/checksums.txt
            ```

         g. Upload the cosign bundle to the GitHub Release:
            ```
            run: gh release upload "${GITHUB_REF_NAME}" dist/checksums.txt.bundle --clobber
            env:
              GH_TOKEN: ${{ secrets.GITHUB_TOKEN }}
            ```

         h. Build-provenance attestation:
              `uses: actions/attest-build-provenance@v2`
              `with: subject-path: 'dist/*.tar.gz'`

         i. Post-release sanity check — verify the cosign bundle:
            ```
            run: |
              cosign verify-blob \
                --bundle dist/checksums.txt.bundle \
                --certificate-identity-regexp 'https://github.com/axonops/fuzzymatch/.+' \
                --certificate-oidc-issuer https://token.actions.githubusercontent.com \
                dist/checksums.txt
            ```

    Concrete identifiers:
      - File `.github/workflows/release.yml`
      - Trigger `push.tags: ['v*']` (no other triggers)
      - Cosign command `cosign sign-blob --yes --bundle dist/checksums.txt.bundle --oidc-issuer https://token.actions.githubusercontent.com dist/checksums.txt`
      - Attestation action `actions/attest-build-provenance@v2`
      - Cosign installer `sigstore/cosign-installer@v3` with `cosign-release: 'v3.0.1'`
      - GoReleaser action `goreleaser/goreleaser-action@v7` with `version: "~> v2"`
  </action>
  <verify>
    <automated>python3 -c "import yaml; cfg=yaml.safe_load(open('.github/workflows/release.yml')); on=cfg.get(True) or cfg.get('on'); assert isinstance(on, dict), 'on must be a dict'; assert 'push' in on and 'tags' in on['push'] and 'v*' in on['push']['tags'], 'tag trigger missing'; assert 'workflow_dispatch' not in on and 'repository_dispatch' not in on and 'pull_request' not in on and 'schedule' not in on, 'forbidden trigger present'; print('OK')" &amp;&amp; grep -q 'cosign sign-blob' .github/workflows/release.yml &amp;&amp; grep -q -- '--bundle' .github/workflows/release.yml &amp;&amp; grep -q 'actions/attest-build-provenance@v2' .github/workflows/release.yml &amp;&amp; grep -q 'sigstore/cosign-installer@v3' .github/workflows/release.yml &amp;&amp; grep -q 'goreleaser/goreleaser-action@v7' .github/workflows/release.yml &amp;&amp; test "$(grep -c 'workflow_dispatch' .github/workflows/release.yml)" = "0"</automated>
  </verify>
  <acceptance_criteria>
    - `.github/workflows/release.yml` exists and is valid YAML
    - The ONLY trigger is `push.tags: ['v*']` — no `workflow_dispatch`, `repository_dispatch`, `schedule`, or `pull_request`
    - `grep -c 'workflow_dispatch' .github/workflows/release.yml` returns 0
    - Workflow-level `permissions: contents: read`
    - Job-level `permissions:` includes `id-token: write`, `contents: write`, `attestations: write`
    - Uses `actions/checkout@v6` with `fetch-depth: 0`
    - Uses `actions/setup-go@v6` with `go-version-file: go.mod`
    - Uses `sigstore/cosign-installer@v3` with `cosign-release: 'v3.0.1'`
    - Uses `goreleaser/goreleaser-action@v7` with `version: "~> v2"`
    - Contains the literal `cosign sign-blob`, `--yes`, `--bundle dist/checksums.txt.bundle`, `--oidc-issuer https://token.actions.githubusercontent.com`, and `dist/checksums.txt` (in any whitespace arrangement)
    - Uses `actions/attest-build-provenance@v2` with `subject-path: 'dist/*.tar.gz'`
    - Contains a `cosign verify-blob` step proving the signature verifies before workflow exit
    - Every release-publishing step carries `if: startsWith(github.ref, 'refs/tags/v')`
    - `actionlint` (if installed) reports no errors
  </acceptance_criteria>
  <done>
    Release workflow triggers exclusively on tag push, performs goreleaser
    release, cosign-signs checksums.txt with --bundle, attaches the bundle
    to the release, generates the build-provenance attestation, and verifies
    the signature before exit.
  </done>
</task>

<task type="auto">
  <name>Task 3: Write commitlint + CLA Assistant workflows + .commitlintrc.yml</name>
  <files>.github/workflows/commitlint.yml, .github/workflows/cla.yml, .commitlintrc.yml</files>
  <read_first>
    - .planning/research/STACK.md (wagoid/commitlint-github-action; contributor-assistant/github-action CLA Assistant Lite)
    - .planning/phases/01-foundation-infrastructure/01-CONTEXT.md (D-08 — CLA Assistant via contributor-assistant; NOT DCO; mirror mask)
    - .claude/skills/commit-standards/SKILL.md (Conventional-commit types — feat, fix, test, docs, chore, refactor, perf)
    - CLAUDE.md ("Branching & Commits" — conventional commits, no AI attribution)
  </read_first>
  <action>
    Wire conventional-commit enforcement and CLA signing.

    1. `.commitlintrc.yml`:
       - `extends: ["@commitlint/config-conventional"]`
       - rules:
         - `type-enum: [2, always, [feat, fix, test, docs, chore, refactor, perf, build, ci, style, revert]]`
         - `subject-case: [2, always, lower-case]`
         - `subject-empty: [2, never]`
         - `subject-full-stop: [2, never, "."]`
         - `header-max-length: [2, always, 72]`
         - `scope-case: [2, always, lower-case]`
       Severity levels: 0=disabled, 1=warning, 2=error.

    2. `.github/workflows/commitlint.yml`:
       - Name: `Commit Lint`
       - Triggers: `pull_request` (lint PR title + all commits in PR)
       - Workflow-level permissions: `contents: read`, `pull-requests: read`
       - Single job `lint`:
         - runs-on: ubuntu-latest
         - steps:
           a. `actions/checkout@v6` with `fetch-depth: 0`
           b. `uses: wagoid/commitlint-github-action@v6` with:
                `configFile: .commitlintrc.yml`
                `helpURL: https://www.conventionalcommits.org/`
                `failOnWarnings: false`

    3. `.github/workflows/cla.yml`:
       - Name: `CLA Assistant`
       - Triggers:
         - `issue_comment:` types: `[created]`
         - `pull_request_target:` types: `[opened, closed, synchronize]`
       - Workflow-level permissions:
           `actions: write`
           `contents: write` (commit signatures file)
           `pull-requests: write` (post status comment on PRs)
           `statuses: write` (set CLA status)
       - Single job `cla-check`:
         - runs-on: ubuntu-latest
         - if: `(github.event_name == 'issue_comment' && contains(github.event.comment.body, 'I have read the CLA Document and I hereby sign the CLA')) || github.event_name == 'pull_request_target'`
         - steps:
           a. `uses: contributor-assistant/github-action@v2.6.1` with:
                `path-to-signatures: signatures/version1/cla.json`
                `path-to-document: https://github.com/axonops/fuzzymatch/blob/main/CLA.md`
                `branch: main`
                `allowlist: dependabot[bot],renovate[bot],github-actions[bot]`
                `lock-pullrequest-aftermerge: false`
              env:
                `GITHUB_TOKEN: ${{ secrets.GITHUB_TOKEN }}`
                `PERSONAL_ACCESS_TOKEN: ${{ secrets.CLA_PAT }}`

    Decision: signatures stored IN the fuzzymatch repo at
    `signatures/version1/cla.json` (mask's pattern per D-08). No remote
    storage org/repo configured.

    Note: CLA.md (the CLA document) is NOT created by this plan. It is a
    content artefact landing in plan 01-08 or added by the user before the
    first external contribution. Until CLA.md exists the
    `path-to-document` URL 404s; the workflow still gates on the
    comment-based signature acquisition. The CLA workflow gracefully
    degrades if `CLA_PAT` is unset — actual signature commits require a
    PAT with write access since `GITHUB_TOKEN`'s default scope is
    insufficient for self-write. The maintainer adds `CLA_PAT` as a repo
    secret after this plan ships; until then the workflow runs but cannot
    commit new signatures (the user's first external-contribution PR will
    fail the CLA gate visibly, prompting the maintainer to add CLA_PAT).

    Concrete identifiers:
      - File `.commitlintrc.yml`
      - Allowed commit types: feat, fix, test, docs, chore, refactor, perf, build, ci, style, revert
      - Header max length: 72
      - CLA bot allowlist: dependabot[bot], renovate[bot], github-actions[bot]
      - CLA action `contributor-assistant/github-action@v2.6.1`
      - Commitlint action `wagoid/commitlint-github-action@v6`
  </action>
  <verify>
    <automated>test -f .commitlintrc.yml &amp;&amp; test -f .github/workflows/commitlint.yml &amp;&amp; test -f .github/workflows/cla.yml &amp;&amp; for f in .commitlintrc.yml .github/workflows/commitlint.yml .github/workflows/cla.yml; do python3 -c "import yaml; yaml.safe_load(open('$f'))" || exit 1; done &amp;&amp; grep -q '@commitlint/config-conventional' .commitlintrc.yml &amp;&amp; grep -q 'wagoid/commitlint-github-action' .github/workflows/commitlint.yml &amp;&amp; grep -q 'contributor-assistant/github-action' .github/workflows/cla.yml</automated>
  </verify>
  <acceptance_criteria>
    - `.commitlintrc.yml` exists, is valid YAML, extends `@commitlint/config-conventional`
    - `.commitlintrc.yml` declares `type-enum` allowing exactly: feat, fix, test, docs, chore, refactor, perf, build, ci, style, revert
    - `.commitlintrc.yml` enforces `header-max-length: 72` at severity 2 (error)
    - `.github/workflows/commitlint.yml` exists, is valid YAML, triggers on `pull_request`
    - `commitlint.yml` references `wagoid/commitlint-github-action@v6` (or latest stable v6.x)
    - `.github/workflows/cla.yml` exists, is valid YAML
    - `cla.yml` triggers on both `issue_comment` and `pull_request_target`
    - `cla.yml` references `contributor-assistant/github-action@v2.6.1`
    - `cla.yml` allowlist includes `dependabot[bot]`, `renovate[bot]`, `github-actions[bot]`
    - `cla.yml` job-level permissions include `contents: write`, `pull-requests: write`, `statuses: write`
    - No workflow file uses `permissions: write-all`
  </acceptance_criteria>
  <done>
    Conventional-commit lint runs on every PR title and commit; CLA Assistant
    runs on PRs with Dependabot/bots allowlisted. Both follow mask's pattern.
  </done>
</task>

</tasks>

<threat_model>
## Trust Boundaries

| Boundary | Description |
|----------|-------------|
| Git tag push → release pipeline | A tag push triggers a fully automated signed release; only maintainers can push tags to main. Branch protection (configured manually by the user post-plan) restricts tag creation to maintainers. |
| GitHub OIDC issuer → cosign / attest | The release relies on GitHub's OIDC issuer for keyless signing identity; trust delegated to GitHub. |
| External contributor PRs → repo | PRs run against PR-target permissions; CLA Assistant gates new contributors; conventional-commit lint enforces commit hygiene. |
| `pull_request_target` trigger in cla.yml | This trigger runs with target-branch permissions (write access). Locked down by allowlist + `if:` condition so only legitimate CLA flows execute. |

## STRIDE Threat Register

| Threat ID | Category | Component | Disposition | Mitigation Plan |
|-----------|----------|-----------|-------------|-----------------|
| T-01-03-01 | Spoofing | Unsigned or maliciously-signed release | mitigate | Cosign keyless via GitHub OIDC; signing identity is `https://github.com/axonops/fuzzymatch/.github/workflows/release.yml`; `cosign verify-blob` in-workflow proves the signature verifies before workflow exit. Consumers can verify with `cosign verify-blob --certificate-identity-regexp ...`. |
| T-01-03-02 | Tampering | Local release bypass (workstation `goreleaser release`) | mitigate | release.yml triggers EXCLUSIVELY on `push.tags`; no `workflow_dispatch`; no manual publish path. Branch protection (manual setup) blocks tag creation by non-maintainers. CLAUDE.md "Releases — CI Only" documents this discipline. devops agent flags any local-tagging pattern. |
| T-01-03-03 | Information Disclosure | Unsigned SBOM | mitigate | SBOM ships via GoReleaser sboms hook; cosign signing covers `checksums.txt` which is the sha256 manifest covering the SBOM file. Consumers verifying the checksum chain transitively verify the SBOM. |
| T-01-03-04 | Repudiation | Build provenance disputed | mitigate | `actions/attest-build-provenance@v2` produces a verifiable OIDC-signed provenance statement linked to the GitHub Actions workflow run. Stored in GitHub's attestation store. |
| T-01-03-05 | Elevation of Privilege | `pull_request_target` privilege abuse in CLA workflow | mitigate | Hard `if:` condition restricts execution to comment-based signatures matching the exact phrase or to PR open/close/sync events. The CLA Assistant action itself does not check out PR code (no `actions/checkout`), so it cannot be tricked into running malicious PR code with target-branch permissions. |
| T-01-03-06 | Tampering | Compromised cosign installer | mitigate | `sigstore/cosign-installer@v3` is a Sigstore-maintained action; pinned via major version + `cosign-release: 'v3.0.1'` for reproducibility. Dependabot watches for updates. |
| T-01-03-07 | Tampering | Conventional-commit lint bypass | accept | Conventional-commit lint is a quality gate, not a security gate; bypass merely produces inconsistent commit history. Branch protection (manual setup) configures commitlint as a required check. |
| T-01-03-08 | Spoofing | Compromised goreleaser-action | mitigate | Pinned at `@v7` major; the action is GoReleaser-maintained. Dependabot watches for updates. |
</threat_model>

<verification>
1. `.goreleaser.yml` is valid YAML and `goreleaser check` exits 0.
2. `.github/workflows/release.yml` is valid YAML, triggers only on
   `push.tags: ['v*']`, and has zero `workflow_dispatch` occurrences.
3. release.yml contains the cosign sign-blob command with `--bundle` and
   the verify-blob sanity check.
4. Commitlint and CLA workflow files are valid YAML; the action versions
   match STACK.md.
5. No workflow uses `permissions: write-all`.
6. `make release-check` exits 0.
</verification>

<success_criteria>
- GoReleaser v2 schema config produces a source tarball + checksums.txt + SBOM.
- Release workflow triggers ONLY on tag push; cosign keyless signs the
  checksum; attestation is generated; verify-blob passes in-workflow.
- Conventional-commit lint and CLA Assistant gate every PR.
- All actions pinned to STACK.md-mandated versions.
- No local-tagging path produces a release.
</success_criteria>

<output>
After completion, create
`.planning/phases/01-foundation-infrastructure/01-03-release-pipeline-SUMMARY.md`
recording:
  - The exact GoReleaser v2 archive file glob set used (since the file
    whitelist is a place where future plans may need to update).
  - Whether `CLA_PAT` was added as a repo secret (track as a follow-up
    if not).
  - Whether CLA.md exists (track as a follow-up for plan 01-08 if not).
  - Branch-protection settings configured by the user (manual step out of
    band; this plan documents that they need to be configured).
</output>
