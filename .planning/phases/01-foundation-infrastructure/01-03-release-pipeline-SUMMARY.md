---
phase: 01-foundation-infrastructure
plan: 03
subsystem: infra
tags: [release-pipeline, goreleaser, cosign, sbom, attestation, commitlint, cla, oidc]

# Dependency graph
requires:
  - phase: 01-foundation-infrastructure
    provides: ci.yml baseline + Makefile `release-check` target (plan 01-02); LICENSE / NOTICE / CHANGELOG / README (plan 01-01)
provides:
  - .goreleaser.yml — GoReleaser v2 library-only release config (source tarball + checksums.txt + Syft SPDX-JSON SBOM + GitHub Release with Keep-a-Changelog extraction)
  - .github/workflows/release.yml — tag-push-only release pipeline (cosign keyless sign + SBOM + actions/attest-build-provenance + verify-blob sanity check)
  - .github/workflows/commitlint.yml — Conventional-commit lint via wagoid/commitlint-github-action@v6
  - .github/workflows/cla.yml — CLA Assistant via contributor-assistant/github-action@v2.6.1, signatures stored in-repo
  - .commitlintrc.yml — Conventional-commit ruleset extending @commitlint/config-conventional
affects: every future plan (commit messages now lint-gated on every PR; releases are tag-push-only and structurally local-unforgeable)

# Tech tracking
tech-stack:
  added: []  # No runtime deps; only CI/release tooling configuration
  patterns:
    - "GoReleaser library mode: `builds:` entry with `skip: true` produces no binaries; the source tarball + checksums.txt + SBOM are the entire artefact set. Pin via `version: \"~> v2\"` in the action; the goreleaser-action@v7 default matches."
    - "Cosign keyless signing of checksums.txt with `--bundle` (v3 default + requirement). OIDC issuer is `https://token.actions.githubusercontent.com`. Bundle uploaded to the GitHub Release as `dist/checksums.txt.bundle`. Verification happens in-workflow via `cosign verify-blob` before exit."
    - "Tag-push-only release trigger (`on.push.tags: ['v*']`). No `workflow_dispatch`, no `repository_dispatch`, no `schedule`, no `pull_request` — local releases are structurally impossible. Every publishing step also carries `if: startsWith(github.ref, 'refs/tags/v')` as defence-in-depth."
    - "Least-privilege workflow permissions: workflow-level `contents: read`; per-job escalation only where required (`contents: write` for release upload, `id-token: write` for OIDC keyless and attest, `attestations: write` for build-provenance, `packages: read` for cosign installer cache)."
    - "CLA Assistant in-repo signatures pattern (mask's pattern per D-08): `path-to-signatures: signatures/version1/cla.json`, `branch: main`. Hard `if:` condition on the job restricts execution to legitimate CLA flows. The action does NOT check out PR code, so `pull_request_target` cannot be tricked into running malicious PR code with target-branch permissions."

key-files:
  created:
    - .goreleaser.yml
    - .github/workflows/release.yml
    - .github/workflows/commitlint.yml
    - .github/workflows/cla.yml
    - .commitlintrc.yml
  modified: []

key-decisions:
  - "Conventional-commit type allowlist landed at 11 types: feat, fix, test, docs, chore, refactor, perf, build, ci, style, revert. The plan listed 11 explicitly; this matches @commitlint/config-conventional's default set plus the explicit additions."
  - "Cosign installer pinned at `cosign-release: 'v3.0.1'` even though `sigstore/cosign-installer@v3` is the major-version pin. v3.0.1 is the spec-locked floor per STACK.md; the major pin lets Dependabot bump within v3 patches while the action body resolves the cosign binary version explicitly."
  - "CLA workflow stores signatures in-repo at `signatures/version1/cla.json` rather than in a separate org-controlled repo. Mirrors mask's pattern (D-08); avoids the operational overhead of a cross-repo signatures store. The `CLA_PAT` secret enables the CLA Assistant action to write to this repo without falling foul of `GITHUB_TOKEN`'s self-write limit."
  - "GoReleaser `extra_files` includes `dist/checksums.txt.bundle` even though `gh release upload` also attaches the same bundle. This is defence-in-depth: if a future refactor of release.yml drops the `gh release upload` step, the asset set remains consistent. Both paths are idempotent on the GitHub Release (the upload uses `--clobber`)."

patterns-established:
  - "Tag-push-only release discipline. Future plans MUST NOT add `workflow_dispatch`, `repository_dispatch`, `schedule`, or any other trigger to release.yml. Branch protection on `main` (manual user setup) should additionally restrict tag creation to maintainers."
  - "Conventional-commit lint is now mandatory on every PR. Future contributions (human or Dependabot) MUST follow the type-enum allowlist and the 72-char header limit. Existing commits in the worktree (this plan's commits use `chore(01-03):`) comply."
  - "Cosign keyless + SBOM + attestation triad for releases. Every release artefact set must include: signed checksums.txt bundle, Syft SPDX-JSON SBOM, build-provenance attestation, and a verify-blob sanity check inside the workflow."

requirements-completed:
  - REL-01
  - REL-02
  - REL-03
  - REL-04
  - REL-05
  - CI-09
  - CI-10

# Metrics
duration: ~12min
completed: 2026-05-13
---

# Phase 1 Plan 3: Release Pipeline Summary

**Landed the full v1.0.0-ready release apparatus — GoReleaser v2 library-only config, tag-push-only release workflow with cosign keyless signing, Syft SPDX-JSON SBOM, build-provenance attestation, conventional-commit lint, and CLA Assistant — making local releases structurally impossible before any release ships.**

## Performance

- **Duration:** ~12 min
- **Tasks:** 3 of 3
- **Files created:** 5
- **Files modified:** 0

## Accomplishments

- **`.goreleaser.yml` (v2 schema, library-only)** — `version: 2` on the first non-comment line; `project_name: fuzzymatch`; `builds:` entry with `skip: true` (no Go binaries); `archives:` produces a `tar.gz` source tarball with explicit file whitelist (LICENSE, NOTICE, README.md, CHANGELOG.md, go.mod, go.sum, Makefile, .golangci.yml, docs/**, **/*.go); `checksum:` sha256 → `checksums.txt`; `sboms:` Syft SPDX-JSON over the archive; `changelog:` block with 7 conventional-commit groups (Features / Fixes / Performance / Refactor / Tests / Docs / Chore) and filters excluding merges, test commits, and version-bump commits; `release:` block scoped to `axonops/fuzzymatch` with `prerelease: auto` (auto-flips at v1.0.0) and `mode: keep-existing`; `snapshot:` template `{{ .Tag }}-next`. `goreleaser check` exits 0; `make release-check` no longer prints the pending message.

- **`.github/workflows/release.yml`** — triggers EXCLUSIVELY on `push.tags: ['v*']`. Workflow-level permissions `contents: read`; single `release` job escalates to `contents: write`, `id-token: write`, `attestations: write`, `packages: read`. Steps: `actions/checkout@v6` (fetch-depth 0), `actions/setup-go@v6` (go-version-file go.mod), `sigstore/cosign-installer@v3` (cosign-release v3.0.1), `anchore/sbom-action/download-syft@v0`, `goreleaser/goreleaser-action@v7` (version `~> v2`, args `release --clean`), `cosign sign-blob --yes --bundle dist/checksums.txt.bundle --oidc-issuer https://token.actions.githubusercontent.com dist/checksums.txt`, `gh release upload "${GITHUB_REF_NAME}" dist/checksums.txt.bundle --clobber`, `actions/attest-build-provenance@v2` over `dist/*.tar.gz`, and finally `cosign verify-blob` proving the signature verifies before workflow exit. Every publishing step also carries `if: startsWith(github.ref, 'refs/tags/v')`. Concurrency group `release-${{ github.ref }}` with `cancel-in-progress: false` to prevent simultaneous releases on the same tag.

- **`.commitlintrc.yml`** — extends `@commitlint/config-conventional`; `type-enum` restricted to exactly 11 types (feat, fix, test, docs, chore, refactor, perf, build, ci, style, revert); `header-max-length: 72`; `subject-case: lower-case`; `subject-empty: never`; `subject-full-stop: never .`; `scope-case: lower-case`. All severities are 2 (error).

- **`.github/workflows/commitlint.yml`** — `pull_request` trigger; runs `wagoid/commitlint-github-action@v6` against `.commitlintrc.yml`. Workflow + job permissions `contents: read`, `pull-requests: read` (least-privilege; the action only needs to read PR metadata).

- **`.github/workflows/cla.yml`** — triggers on `issue_comment` (created) and `pull_request_target` (opened/closed/synchronize). Job-level permissions `actions: write`, `contents: write`, `pull-requests: write`, `statuses: write` (each justified inline). Hard `if:` condition restricts execution to legitimate CLA flows. Runs `contributor-assistant/github-action@v2.6.1` with `path-to-signatures: signatures/version1/cla.json`, `branch: main`, allowlist of `dependabot[bot], renovate[bot], github-actions[bot]`. `CLA_PAT` secret env var is consumed; until it is set, signature commits will fail visibly (the maintainer is then prompted to add it). The CLA document itself (`CLA.md`) lands in plan 01-08; the action's `path-to-document` URL 404s until then, which is acceptable as no external PRs are expected before plan 01-08 ships.

## Task Commits

Each task was committed atomically on `worktree-agent-a0a6f4eb70fe1bb90`:

1. **Task 1: Write .goreleaser.yml (v2 schema, library-only)** — `ff40ddc` (chore)
2. **Task 2: Write release workflow with cosign keyless + SBOM + attestation** — `0623cfc` (chore)
3. **Task 3: Write commitlint + CLA Assistant workflows + .commitlintrc.yml** — `d643672` (chore)

## GoReleaser Archive File Whitelist

The source tarball ships exactly these globs (recorded for forward reference; future plans modifying the source-tarball contents update `.goreleaser.yml`'s `archives[0].files` block):

```yaml
- LICENSE
- NOTICE
- README.md
- CHANGELOG.md
- go.mod
- go.sum
- Makefile
- .golangci.yml
- src: docs
  dst: docs
- src: "**/*.go"
  dst: .
```

Explicitly OMITTED from the source archive (because they are dev / CI / planning artefacts that do not belong in a release tarball): `.planning/`, `.claude/`, `.github/`, `dist/`, `vendor/`, `tests/bdd/`. The whitelist enforces these omissions by their absence from the `files:` block.

The wrap directory is `fuzzymatch_{{ .Version }}` (a release tagged `v0.1.0` produces `fuzzymatch_0.1.0/` as the top-level directory inside `fuzzymatch_0.1.0_source.tar.gz`).

## Action Versions (recorded for STACK.md traceability)

| Action / Tool                                | Pinned version          | STACK.md compliant |
|----------------------------------------------|--------------------------|--------------------|
| `actions/checkout`                           | `@v6` + `fetch-depth: 0` | ✅                 |
| `actions/setup-go`                           | `@v6` + `go-version-file: go.mod` | ✅        |
| `sigstore/cosign-installer`                  | `@v3` + `cosign-release: 'v3.0.1'` | ✅       |
| `anchore/sbom-action/download-syft`          | `@v0`                    | ✅                 |
| `goreleaser/goreleaser-action`               | `@v7` + `version: "~> v2"` | ✅              |
| `actions/attest-build-provenance`            | `@v2`                    | ✅                 |
| `wagoid/commitlint-github-action`            | `@v6`                    | ✅                 |
| `contributor-assistant/github-action`        | `@v2.6.1`                | ✅                 |

**No STACK.md drift.** Every pinned major matches STACK.md "Build & Quality Tooling" and "Release Stack" tables.

## Decisions Made

1. **Library-only release** — `.goreleaser.yml`'s `builds:` entry uses `skip: true` instead of an empty list. An empty `builds:` block would cause GoReleaser to attempt to detect a default build; the explicit skip is unambiguous.
2. **`extra_files` defence-in-depth** — `dist/checksums.txt.bundle` is uploaded both by GoReleaser (via `extra_files`) and by an explicit `gh release upload` step in release.yml. The duplication is intentional: the `gh release upload --clobber` path is the primary upload (it runs after the cosign sign-blob step that writes the bundle, so it is guaranteed to have content), while the `extra_files` entry is a safety net if a future refactor reorders or drops the `gh release upload` step. Both paths are idempotent on the GitHub Release.
3. **Cosign cert-identity regexp** — `cosign verify-blob` uses `--certificate-identity-regexp 'https://github.com/axonops/fuzzymatch/.+'` rather than a pinned exact identity. This accommodates the workflow path changing between runs (e.g., `.github/workflows/release.yml@refs/tags/v0.1.0` vs `@refs/heads/main`) without weakening the OIDC issuer check (which is pinned to `https://token.actions.githubusercontent.com`).
4. **CLA workflow `if:` guard placement** — the `if:` lives on the JOB, not on individual steps. This means the entire job (including the implicit checkout-less setup) is skipped if neither the comment matches nor the event is `pull_request_target`. Simpler and stricter than step-level guards.

## Deviations from Plan

### Auto-fixed Issues

**1. [Rule 1 — Bug] `archives.format` is deprecated in GoReleaser v2; use `formats:` (list).**

- **Found during:** Task 1 verification (`goreleaser check`).
- **Issue:** Plan specified `format: tar.gz` (singular) in the archives block. GoReleaser v2 schema deprecates the singular form in favour of `formats: [tar.gz]` (list). Initial draft used the singular form per the plan; `goreleaser check` warned and would fail under strict mode in a future v2.x release.
- **Fix:** Used `formats: [tar.gz]` (the v2 idiom) from the start in the final file. The `archives:` block still produces a tar.gz; the plan's acceptance criterion "Contains an `archives:` block with `format: tar.gz`" is satisfied by the literal token `tar.gz` appearing in the archives block (verified by `grep -c 'tar.gz' .goreleaser.yml`).
- **Files modified:** `.goreleaser.yml` (forward-compatible idiom).
- **Verification:** `goreleaser check` exits 0; `make release-check` exits 0.
- **Committed in:** `ff40ddc` (the initial Task 1 commit; no separate fix commit needed because this was caught before commit).

**2. [Rule 1 — Bug] `snapshot.name_template` is deprecated; use `version_template`.**

- **Found during:** Task 1 verification (`goreleaser check`).
- **Issue:** Plan specified `snapshot: name_template: "{{ .Tag }}-next"` (the v1 idiom). GoReleaser v2 renamed this to `version_template`. The v1 form still parses but produces a deprecation warning.
- **Fix:** Used `version_template` (the v2 idiom) in the final file.
- **Files modified:** `.goreleaser.yml`.
- **Verification:** `goreleaser check` exits 0 with no deprecation warnings.
- **Committed in:** `ff40ddc`.

**3. [Rule 2 — Critical] `workflow_dispatch` literal in a comment failed the literal `grep -c 'workflow_dispatch' .github/workflows/release.yml == 0` acceptance criterion.**

- **Found during:** Task 2 verification.
- **Issue:** The header comment in release.yml originally listed the forbidden triggers by name ("no `workflow_dispatch`, no `repository_dispatch`, ..."). The plan's acceptance criterion is a literal `grep -c` returning 0. The comment satisfies the *spirit* (the trigger is not present in YAML) but failed the literal check.
- **Fix:** Rewrote the comment to describe the trigger discipline without naming the forbidden tokens. The functional content is unchanged; the comment still warns devops/devs that manual dispatch / repository dispatch / schedule / pull-request triggers are forbidden.
- **Files modified:** `.github/workflows/release.yml`.
- **Verification:** `grep -c 'workflow_dispatch' .github/workflows/release.yml` returns `0`; the python yaml parse still confirms exclusive tag trigger.
- **Committed in:** `0623cfc` (Task 2 commit; fix bundled into the initial commit because it was caught before commit).

**Total deviations:** 3 auto-fixed. Two were GoReleaser v2-vs-v1 idiom updates surfaced by `goreleaser check`; one was a comment-vs-literal-grep mismatch surfaced by the plan's own automated verify script.

## Issues Encountered

- **Locally installed `goreleaser` is v2.15.3, STACK.md pins v2.15.4.** Schema is forwards-compatible between v2.15.3 and v2.15.4 — the same `.goreleaser.yml` drives both. CI uses goreleaser-action@v7 with `version: "~> v2"` which resolves to the latest v2 patch at release time. No action needed.
- **`coverage.out` artefact left behind by `make check`.** Plan 01-02's Makefile produces this and there is no `.gitignore` in the repo yet. Cleaned up post-verification; the file should be added to `.gitignore` in a future plan (likely plan 01-08's docs/contributing pass).
- **`commitlint` cannot be exercised locally.** The wagoid action runs in CI only (Node-based). The `.commitlintrc.yml` content was hand-validated against `@commitlint/config-conventional`'s rule schema; CI will exercise it on the first PR.

## User Setup Required (Follow-ups)

The following manual configuration steps remain. None block this plan from completing.

1. **`CLA_PAT` repository secret** — NOT YET ADDED. Required by `.github/workflows/cla.yml` for the CLA Assistant action to commit new signatures to `signatures/version1/cla.json`. Until the secret is set, signature commits will fail visibly (the workflow runs but the action errors when attempting to write). The maintainer should create a fine-scoped PAT with `contents: write` on this repo and add it as `CLA_PAT` before the first external-contribution PR.

2. **`CLA.md` document** — NOT YET CREATED. Referenced by `cla.yml` at `https://github.com/axonops/fuzzymatch/blob/main/CLA.md`. The URL 404s until the file lands. Tracked for plan 01-08 (CONTRIBUTING / governance docs); this is acceptable because no external PRs are expected before plan 01-08 ships.

3. **Branch protection on `main`** — NOT YET CONFIGURED. Per the threat-model (T-01-03-02), branch protection should restrict tag creation to maintainers. Recommended required checks: `CI / make check`, `CI / markdownlint`, `Security / govulncheck`, `Security / gosec`, `CodeQL / Analyze (go)`, `License Headers / verify-license-headers`, `Commit Lint / commitlint`, `CLA Assistant / cla-check`. This is a manual GitHub-UI configuration step out-of-band of code.

## Threat Surface Scan

Reviewed every workflow's permissions block and trigger surface. Threat-register dispositions all match the plan's `<threat_model>` section:

- **T-01-03-01 (Spoofing — unsigned release):** Mitigated by cosign keyless via GitHub OIDC; in-workflow `cosign verify-blob` proves the signature verifies before workflow exit.
- **T-01-03-02 (Tampering — local release bypass):** Mitigated by exclusive tag-push trigger; `grep -c 'workflow_dispatch' .github/workflows/release.yml` returns 0.
- **T-01-03-03 (Information Disclosure — unsigned SBOM):** Mitigated transitively — `checksums.txt` covers the SBOM file, and the bundle signs `checksums.txt`.
- **T-01-03-04 (Repudiation — disputed build provenance):** Mitigated by `actions/attest-build-provenance@v2`.
- **T-01-03-05 (Elevation of Privilege — `pull_request_target` abuse):** Mitigated by hard `if:` condition + no `actions/checkout` in the CLA job; the CLA Assistant action cannot run PR code with target-branch permissions.
- **T-01-03-06 (Tampering — compromised cosign installer):** Mitigated by Sigstore-maintained action pinned at `@v3` + explicit `cosign-release: 'v3.0.1'`; Dependabot watches for updates.
- **T-01-03-07 (Tampering — commitlint bypass):** Accepted (quality gate, not security gate). Branch protection (manual setup above) configures commitlint as a required check.
- **T-01-03-08 (Spoofing — compromised goreleaser-action):** Mitigated by `@v7` major pin + Dependabot.

No new threat surfaces introduced beyond the plan's threat register. No threat flags to raise.

## Next Plan Readiness

- **Plan 01-04 (determinism infra)** — inherits the full release pipeline and adds the cross-platform CI matrix + determinism golden test + `scripts/verify-no-runtime-deps.sh` + benchstat harness. The CI matrix expansion is independent of release.yml; the release workflow stays single-runner (`ubuntu-latest`) since the source tarball is platform-neutral.
- **Plan 01-08 (docs / governance)** — must land `CLA.md` (referenced by `cla.yml`) and `CONTRIBUTING.md` (documenting the conventional-commit policy, REL-07 algorithm deprecation policy, and the release-discipline contract). Should also add `.gitignore` covering `coverage.out`, `dist/`, `bench.txt.new`, and IDE/OS artefacts.
- **Phase 11 (v1.0.0)** — first cosign-signed release flows entirely through `.github/workflows/release.yml`. No local tagging path exists; the maintainer pushes a `v1.0.0` tag from main (or merges a release PR that creates the tag) and the workflow does the rest.

## Self-Check: PASSED

Files claimed to exist (verified with `[ -f ... ]`):

- `.goreleaser.yml` — FOUND
- `.github/workflows/release.yml` — FOUND
- `.github/workflows/commitlint.yml` — FOUND
- `.github/workflows/cla.yml` — FOUND
- `.commitlintrc.yml` — FOUND

Commits claimed to exist (verified with `git log --oneline`):

- `ff40ddc` — FOUND (Task 1: .goreleaser.yml)
- `0623cfc` — FOUND (Task 2: release.yml)
- `d643672` — FOUND (Task 3: commitlint.yml + cla.yml + .commitlintrc.yml)

Plan-level verification (from `<verification>` section):

- `.goreleaser.yml` is valid YAML and `goreleaser check` exits 0 — PASS
- `.github/workflows/release.yml` is valid YAML, triggers only on `push.tags: ['v*']`, and has zero `workflow_dispatch` occurrences — PASS
- release.yml contains the cosign sign-blob command with `--bundle` and the verify-blob sanity check — PASS
- Commitlint and CLA workflow files are valid YAML; action versions match STACK.md — PASS
- No workflow uses `permissions: write-all` — PASS (`grep -rc 'write-all' .github/workflows/` returns 0 across all six workflow files)
- `make release-check` exits 0 — PASS
- `make check` exits 0 end-to-end on the new tree — PASS
- `actionlint .github/workflows/*.yml` exits 0 — PASS

---
*Phase: 01-foundation-infrastructure*
*Plan: 03 (release-pipeline)*
*Completed: 2026-05-13*
