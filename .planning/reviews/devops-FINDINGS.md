---
status: issues_found
agent: devops
scope: CI/CD + release + scripts (phases 1-8)
reviewed: 2026-05-17T08:30:00Z
finding_counts:
  critical: 0
  important: 9
  improvement: 17
  total: 26
---

# DevOps Review — fuzzymatch (Phases 1–8)

Scope: `.github/workflows/{ci,release,security,codeql,commitlint,cla,license-headers}.yml`, `.github/dependabot.yml`, `.github/CODEOWNERS`, `.github/PULL_REQUEST_TEMPLATE.md`, `Makefile`, `.goreleaser.yml`, `.golangci.yml`, `.markdownlint-cli2.yaml`, `.commitlintrc.yml`, `scripts/verify-*.sh`, `.gitignore`.

Overall posture is strong. Release discipline is sound: tag-push-only trigger, cosign keyless with `--bundle`, build-provenance attestation, post-release verify step, Syft SBOM via GoReleaser. The 5-platform matrix is complete. There are no Critical findings — no local-tag patterns, no missing signing on release artefacts, no unpinned destructive actions on release.

The Important findings cluster around (a) one unpinned action tag, (b) absence of an explicit `make check`/CI nightly workflow promised by CLAUDE.md, (c) Dependabot grouping gaps, (d) commit-message prefix divergence from Conventional Commits, (e) `release.yml` not gating on CI green, (f) gosec `-no-fail` masking failures from blocking, (g) coverage script's `_test.go` files counted against per-file floor, (h) no concurrency group on CI workflow, and (i) absent `nightly.yml` long-form fuzz + benchstat-regression job that CLAUDE.md describes.

Improvements are mostly hardening: SHA-pinning vs tag-pinning, narrower step-level conditionals, missing `make` targets enumerated in CLAUDE.md, modest gaps in the Makefile coverage of CI parity, and a number of micro-robustness opportunities in the bash scripts.

Findings follow.

---

### [Important] `DavidAnson/markdownlint-cli2-action@latest-stable` is the only floating reference in the entire workflow tree
- **File:** `/Users/johnny/Development/fuzzymatch/.github/workflows/ci.yml`
- **Phase introduced:** Phase 1 (01-02 quality gates)
- **Issue:** Every other action across all seven workflows is pinned to a specific major (`@v6`, `@v7`, `@v4`, `@v3`, `@v0`) or exact tag (`@v2.6.1`, `@v2.25.0`). `DavidAnson/markdownlint-cli2-action@latest-stable` floats — `latest-stable` resolves to whichever release the maintainer last tagged that way and can change underfoot. This breaks reproducibility of the markdownlint job and is the single failure-of-pinning across the workflow tree.
- **Standard:** CLAUDE.md "Workflow pinning (action versions pinned, no `@latest`)"; STACK.md "markdownlint-cli2 v0.22.1".
- **Action:** Code fix.
- **Rationale:** A floating tag is functionally `@latest` for supply-chain analysis. STACK.md already pins markdownlint-cli2 to v0.22.1; the action wrapper should be pinned to a corresponding major (current is `@v20` series for `markdownlint-cli2-action`).
- **Suggested fix:** Pin to `DavidAnson/markdownlint-cli2-action@v20` (or whichever current major matches the v0.22.1 CLI). For maximum hardening pin to a SHA with a comment recording the corresponding tag.

### [Important] `release.yml` does not require CI to be green before publishing
- **File:** `/Users/johnny/Development/fuzzymatch/.github/workflows/release.yml`
- **Phase introduced:** Phase 1 (01-03 release pipeline)
- **Issue:** The release workflow triggers on tag push and runs GoReleaser without a `needs:` dependency on the CI quality gate or any pre-release reachability check that the tag's commit passed CI. CLAUDE.md "Release workflow that doesn't require CI checks to pass first" is listed as a BLOCKING violation. The current model relies on branch-protection on `main` plus a maintainer's discipline ("only tag commits that merged green to main"), but the workflow itself does not encode that.
- **Standard:** CLAUDE.md `## CI Workflow Requirements > release.yml — "Runs full CI checks before release"`.
- **Action:** Code fix (recommend) OR Discuss-phase needed (decision).
- **Rationale:** A tag pushed to any commit triggers release regardless of CI state. Branch protection on `main` reduces but does not eliminate the risk: e.g. a maintainer with `Allow specified actors to bypass required pull requests` could push a tag to a non-`main` commit. The defence-in-depth pattern is to add a job-level guard that verifies the tagged SHA has a green CI run (via `gh api /repos/$REPO/commits/$SHA/check-runs`) or to add an explicit `needs:` if CI is restructured to be a reusable workflow.
- **Suggested fix:** Add a pre-release job that calls `gh api repos/${{ github.repository }}/commits/${{ github.sha }}/status` (or `check-runs`) and fails if combined status is not `success`. Alternative: convert CI into a reusable workflow and `needs: ci-required` from release.yml.

### [Important] No `nightly.yml` workflow — CLAUDE.md describes one with long-form fuzz + benchstat-regression + auto-PR corpus
- **File:** missing (`/Users/johnny/Development/fuzzymatch/.github/workflows/nightly.yml`)
- **Phase introduced:** N/A — never landed
- **Issue:** CLAUDE.md `## CI Workflow Requirements > nightly.yml (daily 03:00 UTC):` enumerates three deliverables — long-form fuzz (5 min per fuzzer), benchstat regression detection against last tagged release (>10% auto-opens an issue), and a self-hosted runner shared with mask/audit. None of these run today. `make test-fuzz` exists but only runs 60s ("short fuzz") via `make check` parity. The benchstat regression gate is `continue-on-error: true` per D-09 and would be re-enabled when a self-hosted runner becomes available — but the nightly workflow is the missing carrier that would enable that gate even without a self-hosted runner (informationally against `ubuntu-latest` over a longer baseline).
- **Standard:** CLAUDE.md `nightly.yml` section (three bullet points).
- **Action:** Discuss-phase needed.
- **Rationale:** Long-form fuzz catches state-space corners that a 60s budget can miss; nightly regression-detection on `ubuntu-latest` would be informational (per D-09) but still valuable as drift-radar. The auto-PR-new-corpus pattern is the standard go-fuzz hygiene loop. Project may be deferring this until the self-hosted runner lands; needs an explicit decision recorded.
- **Suggested fix:** Either land `nightly.yml` (workflow_dispatch + cron daily 03:00 UTC) or formally defer in 01-CONTEXT.md with a tracked Deferred Item linking the nightly enablement to the self-hosted-runner availability.

### [Important] Dependabot groups direct + indirect but no test-only group for the root module
- **File:** `/Users/johnny/Development/fuzzymatch/.github/dependabot.yml`
- **Phase introduced:** Phase 1
- **Issue:** CLAUDE.md `Dependabot` section: "Go modules (PRs grouped: indirect / direct / test-only)". The root `gomod` ecosystem block groups `indirect` and `direct` but has no `test-only` group. Practically the root module has only one runtime dep (`golang.org/x/text`) and no test-only deps because testify is forbidden in root — but a future Phase 5+ may introduce test-helpers and the grouping rule should mirror CLAUDE.md verbatim or the project should record the intentional divergence.
- **Standard:** CLAUDE.md `Dependabot: Go modules (PRs grouped: indirect / direct / test-only)`.
- **Action:** Code fix OR documentation in 01-CONTEXT.md.
- **Rationale:** CLAUDE.md compliance + future-proofing. The `tests/bdd` ecosystem block also lacks the test-only grouping despite all its deps being test-only (godog, goleak, testify); a `test-only` group there with `dependency-type: development` would be more semantically accurate than the current `indirect`/`direct` partition that has no semantic meaning for a test-only module.
- **Suggested fix:** Either (a) add `test-only` group to both `gomod` blocks for parity with CLAUDE.md, or (b) update CLAUDE.md to reflect the simplified two-group model with a one-line rationale (root has no test-only deps by design; `tests/bdd` is wholly test-only so grouping is structural not categorical).

### [Important] No auto-merge configuration for Dependabot patch-only updates
- **File:** `/Users/johnny/Development/fuzzymatch/.github/dependabot.yml` (and absent `.github/workflows/dependabot-automerge.yml`)
- **Phase introduced:** Phase 1
- **Issue:** CLAUDE.md `Dependabot: Daily check; auto-merge enabled for patch updates that pass CI`. No `dependabot-automerge.yml` workflow exists. Daily PR open is configured, but they require manual review-and-merge.
- **Standard:** CLAUDE.md `Dependabot: auto-merge enabled for patch updates that pass CI`.
- **Action:** Code fix.
- **Rationale:** Maintainer time-economy + faster security-patch landing. The standard pattern is a separate workflow that watches `pull_request` events, checks `github.event.pull_request.user.login == 'dependabot[bot]'`, parses the update-type from the Dependabot metadata action, and runs `gh pr merge --auto --squash` for `version-update:semver-patch`.
- **Suggested fix:** Add `.github/workflows/dependabot-auto-merge.yml` with `dependabot/fetch-metadata@v2` and `gh pr merge --auto --squash` on patch updates. Permissions: `contents: write`, `pull-requests: write`.

### [Important] `gosec` runs with `-no-fail` — security findings never block CI
- **File:** `/Users/johnny/Development/fuzzymatch/.github/workflows/security.yml`
- **Phase introduced:** Phase 1
- **Issue:** Line 56: `args: '-no-fail -fmt sarif -out gosec.sarif ./...'`. With `-no-fail`, gosec exits 0 regardless of findings; the SARIF is uploaded to the Security tab but no PR check is failed. CLAUDE.md `security.yml (weekly + on PR): gosec ./...` implies blocking behaviour, and `.golangci.yml` notes "the definitive gosec pass happens in .github/workflows/security.yml with SARIF upload" — but the definitive pass is not actually definitive if findings cannot fail CI.
- **Standard:** CLAUDE.md `security.yml: gosec ./...`; `.golangci.yml:11` comment claims definitive behaviour in security.yml.
- **Action:** Code fix.
- **Rationale:** Findings sit in the GitHub Security tab unattended until a maintainer notices them. Blocking on MEDIUM+ findings (with a documented allowlist mechanism for false positives) is the standard pattern. Equivalent to govulncheck which DOES fail CI today (line 36: `run: govulncheck ./...`).
- **Suggested fix:** Remove `-no-fail` and instead use `-severity medium` (or `high`) to scope blocking severity. Keep SARIF upload as a separate step that runs `if: always()` so the upload happens regardless of gosec exit. Add a `.gosec.yml` config (or `// #nosec G123 -- rationale` annotations) for documented false positives.

### [Important] Coverage-floor script lints `_test.go` files against the 90% per-file floor
- **File:** `/Users/johnny/Development/fuzzymatch/scripts/verify-coverage-floors.sh`
- **Phase introduced:** Phase 1 (01-04 determinism infra)
- **Issue:** The per-file floor loop at lines 101–137 iterates over every file appearing in `coverage.out`. The `go test -coverpkg=./...` profile includes `_test.go` files when they contain executable statements (table init, helper funcs). Test files have variable coverage — a helper that is only called from one specific test will appear with 0–100% coverage on the helper depending on inclusion. The script does not filter `_test.go` from the per-file floor, which means CI can fail on the test-file coverage of an unused test helper. This has not bitten yet because `_test.go` files typically all-execute on the same `go test` invocation, but the gap is real.
- **Standard:** `.claude/skills/go-testing-standards/SKILL.md` per-file ≥90% floor — intent is "production code", not test helpers.
- **Action:** Code fix.
- **Rationale:** Defensive correctness. Even if the current test files all hit 100%, the floor should not police `_test.go`.
- **Suggested fix:** In the awk pipeline at lines 116–137, skip files matching `_test\.go$`. Also consider skipping `doc.go` and `example_test.go` if they appear (the script already handles zero-statement files, which covers most of this).

### [Important] CI workflow has no `concurrency:` block — duplicate runs on rapid pushes
- **File:** `/Users/johnny/Development/fuzzymatch/.github/workflows/ci.yml`
- **Phase introduced:** Phase 1
- **Issue:** Only `release.yml` declares `concurrency:` (line 42). `ci.yml`, `security.yml`, `codeql.yml`, `commitlint.yml`, `license-headers.yml` will all spawn duplicate runs on rapid pushes (force-push to a PR branch, repeated `git push` while CI is mid-run). Each redundant run consumes runner-minutes and clutters PR check status.
- **Standard:** CLAUDE.md `Concurrency groups defined where appropriate` — not algorithmically specified but reviewer-judgment.
- **Action:** Code fix.
- **Rationale:** Runner-minute economy + UI clarity. The standard pattern is `concurrency: { group: ${{ github.workflow }}-${{ github.ref }}, cancel-in-progress: true }` on PR-triggered workflows (cancel-in-progress safe because the new run supersedes), and `cancel-in-progress: false` on push-to-main (cancellation could leave coverage data partially populated).
- **Suggested fix:** Add a workflow-level `concurrency:` block to ci.yml, security.yml, codeql.yml, commitlint.yml, license-headers.yml. For PRs, cancel-in-progress: true; for push-to-main, scope the group differently to allow main to always complete.

### [Important] No `cache:` directive on `actions/setup-go@v6` — module download repeats every job
- **File:** `/Users/johnny/Development/fuzzymatch/.github/workflows/{ci,release,security,codeql}.yml`
- **Phase introduced:** Phase 1
- **Issue:** `setup-go@v6` enables build/module caching by default when `go.sum` is present (Go's setup-go has had this since v5), so caching IS happening. But the workflows do not declare `cache: true` explicitly, nor do they declare `cache-dependency-path` — for the BDD sub-module (`tests/bdd/go.sum`), the cache key derives only from the root `go.sum`, which means a change to `tests/bdd/go.sum` does not invalidate the cache and CI may use stale dependencies.
- **Standard:** CLAUDE.md `Cache strategies (Go modules, build cache)`.
- **Action:** Code fix.
- **Rationale:** Correctness of cache invalidation. Default behaviour caches against the root `go.sum` only; the BDD sub-module needs its `go.sum` in the cache key so dependency drift in `tests/bdd` correctly invalidates.
- **Suggested fix:** Add to each `setup-go` step:
  ```yaml
  cache-dependency-path: |
    go.sum
    tests/bdd/go.sum
  ```

---

### [Improvement] Action tags pinned to majors not SHAs
- **File:** all workflow files
- **Phase introduced:** Phase 1
- **Issue:** Actions are pinned to majors (`actions/checkout@v6`, `actions/setup-go@v6`, `goreleaser/goreleaser-action@v7`) rather than SHAs. Major-tag pinning is the GitHub recommended baseline; SHA-pinning is the supply-chain-hardened option (used by `axonops/mask` in some places per STACK.md). For a security-conscious library being prepared for downstream consumption in Cassandra-adjacent workloads (per BOOTSTRAP context), SHA-pinning the release workflow specifically is justifiable defence-in-depth.
- **Standard:** No project-specific rule mandates SHA-pinning; CLAUDE.md `Workflow pinning (action versions pinned, no @latest)` is satisfied by majors. This is a hardening recommendation.
- **Action:** Discuss-phase needed.
- **Rationale:** Major-tag pinning allows the action author to retroactively swap the SHA the tag points to. SHA-pinning eliminates that. Trade-off: SHA-pinned actions don't auto-receive security patches; Dependabot for github-actions handles refresh.
- **Suggested fix:** SHA-pin actions invoked in `release.yml` (the highest-trust workflow) at minimum: `goreleaser/goreleaser-action`, `sigstore/cosign-installer`, `actions/attest-build-provenance`, `anchore/sbom-action/download-syft`. Leave CI workflow actions on major tags. Document the policy in `.github/dependabot.yml` comments.

### [Improvement] `anchore/sbom-action/download-syft@v0` — `@v0` is unusual major pinning
- **File:** `/Users/johnny/Development/fuzzymatch/.github/workflows/release.yml`
- **Phase introduced:** Phase 1
- **Issue:** Line 63 pins to `@v0`. The anchore/sbom-action repo's tagging is `v0.x.y` semver-zero. `@v0` resolves to whatever the latest `v0.x.y` is — which is a wider window than `@v1` or `@v2`-style pins. Project STACK.md notes "latest stable (current ~v1.42)" for syft itself but the action is at v0.
- **Standard:** CLAUDE.md `Workflow pinning`.
- **Action:** Code fix.
- **Rationale:** Tighten the resolution window. The action's current latest is around v0.18.x; `@v0.18` (or SHA-pin) is a tighter bound than `@v0`.
- **Suggested fix:** Pin to the latest minor (`@v0.18`) or SHA.

### [Improvement] `sigstore/cosign-installer@v3` major-tag pin + explicit `cosign-release` is redundant-but-safe
- **File:** `/Users/johnny/Development/fuzzymatch/.github/workflows/release.yml`
- **Phase introduced:** Phase 1
- **Issue:** Lines 58–60: `uses: sigstore/cosign-installer@v3` with `cosign-release: "v3.0.1"`. The action major v3 is pinned AND the installed cosign binary version is pinned. This is correct defence-in-depth (the action ref controls "what code runs"; `cosign-release` controls "what binary it installs"). Just noting it for completeness — no action needed. Same pattern is missing from goreleaser-action where only `version: "~> v2"` controls the binary; the action ref is `@v7`. Both are fine.
- **Standard:** N/A.
- **Action:** No action needed (informational).
- **Rationale:** Strong pinning.

### [Improvement] `.golangci.yml` `gosec` linter is enabled in lint AND in security.yml — duplicated runs
- **File:** `/Users/johnny/Development/fuzzymatch/.golangci.yml`
- **Phase introduced:** Phase 1
- **Issue:** Line 34 enables `gosec` as part of `golangci-lint run`, and `.github/workflows/security.yml` ALSO runs gosec via `securego/gosec@v2.25.0`. Two configurations, two runs, two possible signal sources for the same finding. golangci-lint's gosec is bundled and version-locked; the standalone security.yml run uses v2.25.0. These can disagree.
- **Standard:** No project-specific rule. `.golangci.yml:11` comment acknowledges that the definitive gosec runs in security.yml.
- **Action:** Code fix.
- **Rationale:** Avoid signal duplication and version-skew. Pick one as canonical.
- **Suggested fix:** Remove `gosec` from `.golangci.yml linters.enable`; leave the standalone `security.yml` as the definitive source. The `.golangci.yml` comment already implies this is intended.

### [Improvement] Two scheduled workflows both run at 06:00 UTC — capacity contention
- **File:** `/Users/johnny/Development/fuzzymatch/.github/workflows/{security,codeql}.yml`
- **Phase introduced:** Phase 1
- **Issue:** `security.yml` schedules `0 6 * * 1` (Mondays 06:00 UTC); `codeql.yml` schedules `0 6 * * 2` (Tuesdays 06:00 UTC). They are different days so no contention in practice, but a future scheduled workflow may collide. Also, 06:00 UTC is high-traffic on the GHA runner pool (start-of-EU-business). 03:00 UTC (start of low-traffic window) is the standard schedule choice and matches what CLAUDE.md specifies for `nightly.yml`.
- **Standard:** CLAUDE.md `nightly.yml (daily 03:00 UTC)`.
- **Action:** Code fix.
- **Rationale:** Avoid queue-time spikes; align with CLAUDE.md's nightly schedule.
- **Suggested fix:** Move both scheduled crons to `0 3 * * 1` / `0 3 * * 2` for consistency with CLAUDE.md's nightly cadence.

### [Improvement] No `make verify-llms-sync` target — CLAUDE.md lists one
- **File:** `/Users/johnny/Development/fuzzymatch/Makefile`
- **Phase introduced:** Phase 1
- **Issue:** CLAUDE.md `## CI Workflow Requirements > ci.yml: LLMs sync check (scripts/verify-llms-sync.sh) — llms.txt references every exported symbol`. The check exists as a Go test (`ai_friendly_test.go`) not as `scripts/verify-llms-sync.sh`. That is structurally fine (the test runs as part of `go test ./...`) but the script-target listed in CLAUDE.md does not exist. CLAUDE.md `Branch protection on main: Required status checks: ... llms-sync` references it as a named status check.
- **Standard:** CLAUDE.md `ci.yml` step list; `Branch protection on main` required-status list.
- **Action:** Discuss-phase needed.
- **Rationale:** Either (a) add a thin `scripts/verify-llms-sync.sh` that runs `go test -run TestAIFriendly_LLMSTxtReferencesEveryExportedSymbol ./...` for parity with CLAUDE.md, or (b) update CLAUDE.md to reflect that the check is a Go test inside the unit test suite and named under `make check`. Option (b) is structurally cleaner.
- **Suggested fix:** Update CLAUDE.md or add a wrapper script that invokes the existing Go test.

### [Improvement] CLAUDE.md lists status check names that don't exist as separate jobs (build, test, coverage, fuzz-short, etc.)
- **File:** N/A — branch protection policy
- **Phase introduced:** Phase 1
- **Issue:** CLAUDE.md `Branch protection on main: Required status checks: lint, vet, markdown-lint, build (5 platforms), test, coverage, fuzz-short, determinism, cross-platform-determinism, BDD, vulncheck, license-check, no-runtime-deps, llms-sync`. None of these (except `markdownlint`, `verify-license-headers`, `govulncheck`, `gosec`, `Analyze (go)`) exist as individual GitHub-check-run names. `make check` consolidates lint+vet+fmt-check+verify-license-headers+verify-deps-allowlist+tidy-check+security+test+coverage+coverage-check into a single named check `make check (linux-amd64)` (etc per matrix). This may diverge from the branch-protection rules a maintainer would want to configure literally.
- **Standard:** CLAUDE.md `Branch protection on main`.
- **Action:** Discuss-phase needed.
- **Rationale:** Either decompose CI jobs into per-step check-runs (clearer signals, more maintenance) or update CLAUDE.md to list the actual check-run names (`make check (linux-amd64)`, `make check (linux-arm64)`, `make check (darwin-amd64)`, `make check (darwin-arm64)`, `make check (windows-amd64)`, `markdownlint`, `verify-license-headers`, `commitlint`, `govulncheck`, `gosec`, `Analyze (go)`).
- **Suggested fix:** Update CLAUDE.md `Branch protection on main` to list actual check-run names, OR decompose CI workflow into one job per category. The aggregate-job approach is simpler operationally; CLAUDE.md should match reality.

### [Improvement] `Makefile` `release-check` target is tolerant no-op when `.goreleaser.yml` is absent — but `.goreleaser.yml` IS present
- **File:** `/Users/johnny/Development/fuzzymatch/Makefile`
- **Phase introduced:** Phase 1
- **Issue:** Lines 272–281: tolerant no-op for `.goreleaser.yml` absence. Plan 01-03 has landed `.goreleaser.yml`, so the tolerance is now stale. The fallthrough branch `echo "goreleaser not installed; install per docs/CONTRIBUTING (plan 01-08)."` references plan 01-08 — that planning reference will become noise post-Phase 1.
- **Standard:** CLAUDE.md `make release-check validates goreleaser config without releasing`.
- **Action:** Code fix.
- **Rationale:** Stale tolerance. Simplify to: if goreleaser absent, print clear install hint; if `.goreleaser.yml` absent, hard-fail (an absent goreleaser config in Phase 2+ is a bug).
- **Suggested fix:** Remove the `if [ -f .goreleaser.yml ]` guard; `.goreleaser.yml` exists now.

### [Improvement] `Makefile` `bench` and `bench-compare` tolerance referring to "Phase 2" is stale
- **File:** `/Users/johnny/Development/fuzzymatch/Makefile`
- **Phase introduced:** Phase 1
- **Issue:** Lines 101–113 (bench), 123–136 (bench-compare): tolerance handlers reference "pending Phase 2" / "pending plan 01-04". Phases 2–7 have shipped (per `.planning/phases/` listing), so benchmarks exist now and the tolerance branch is dead code. Same applies to `coverage-check` (line 153: "pending Phase 2") and `verify-determinism` (line 193: "Plan 01-04 lands the golden-form determinism harness. Until then this target runs `go test -run TestGolden_ ./...` which currently matches no tests" — `algorithms_golden_test.go` exists and runs now).
- **Standard:** N/A — code hygiene.
- **Action:** Code fix.
- **Rationale:** Stale comments mislead future readers. The tolerant-no-op branches still work but communicate the wrong state.
- **Suggested fix:** Update or remove the tolerance-comment blocks for `bench`, `bench-compare`, `coverage-check`, `verify-determinism`. Either delete the bootstrap-state branches entirely or annotate with "kept for re-bootstrap safety" if intentional.

### [Improvement] `verify-no-runtime-deps.sh` uses `-mod=mod` which can rewrite `go.mod`
- **File:** `/Users/johnny/Development/fuzzymatch/scripts/verify-no-runtime-deps.sh`
- **Phase introduced:** Phase 1 (01-04)
- **Issue:** Line 61: `go list -m -mod=mod -f ... all`. `-mod=mod` enables `go list` to add missing requirements to `go.mod` if needed. In CI on a checkout, `go.mod` and `go.sum` are present; `-mod=readonly` (the default in Go 1.16+) would suffice and is safer — it errors instead of mutating.
- **Standard:** Go module hygiene.
- **Action:** Code fix.
- **Rationale:** Reduce blast radius. The script should be read-only.
- **Suggested fix:** Change to `-mod=readonly` (or remove the flag — the default is readonly).

### [Improvement] `verify-coverage-floors.sh` parses `go doc -short .` which can include non-root deferred symbols
- **File:** `/Users/johnny/Development/fuzzymatch/scripts/verify-coverage-floors.sh`
- **Phase introduced:** Phase 1 (01-04)
- **Issue:** Line 166: `go doc -short . 2>/dev/null`. `go doc` operates against the current working directory's main package; this is correct for the root package but does NOT cover the `scan/` sub-package or any future sub-package. CLAUDE.md `Coverage report (must hit 95% overall, 90% per file, 100% public API)` is silent on per-sub-package public-API enforcement, but the spirit is "every exported symbol". `scan/` will land in Phase 9; the script will need updating then.
- **Standard:** CLAUDE.md `100% public API`.
- **Action:** Discuss-phase needed (Phase 9 prep).
- **Rationale:** Phase 9 (`scan/`) is approaching; the public-API floor needs to extend to it.
- **Suggested fix:** When `scan/` lands, extend the script to iterate `for pkg in . ./scan; do go doc -short "$pkg"; done`. Track in 09-CONTEXT.md as a Deferred Item.

### [Improvement] `verify-license-headers.sh` does not check for the AxonOps copyright year update
- **File:** `/Users/johnny/Development/fuzzymatch/scripts/verify-license-headers.sh`
- **Phase introduced:** Phase 1 (01-01)
- **Issue:** The header signature checked is the literal `Licensed under the Apache License, Version 2.0`. This passes regardless of the copyright year line (`Copyright 2026 AxonOps Limited`). A file with `Copyright 2024 SomeOther Org` plus the Apache-2.0 line would pass the check. The Apache header is the legally important part; the copyright line is project-hygiene. The current scope is acceptable but a stricter check (signature must include `Copyright YYYY AxonOps Limited`) is justified.
- **Standard:** `.claude/skills/documentation-standards.md` (project conventions).
- **Action:** Improvement.
- **Rationale:** Strictness. The current check is fit-for-purpose for the spirit of "Apache header present"; a stricter check enforces copyright-attribution consistency.
- **Suggested fix:** Optional — add a second header check requiring `Copyright [0-9]{4} AxonOps Limited`. Track tradeoff: more brittle (year rollovers, BSL-style headers) vs more accurate attribution check.

### [Improvement] `coverage-check` lacks tolerance flag for non-statement `_test.go` files generated by Go tooling
- **File:** `/Users/johnny/Development/fuzzymatch/scripts/verify-coverage-floors.sh`
- **Phase introduced:** Phase 1
- **Issue:** Lines 116–137: the awk parser splits coverage lines by `: . ,` and assumes filename is everything-before-first-colon. For files in subdirectories the filename will include the directory portion correctly, but for filenames containing colons (rare, but legal on POSIX) the parse breaks. Also: files like `export_test.go` are part of the root package's `package fuzzymatch` (white-box) and will appear in the coverage profile. If they have low coverage they will fail the per-file floor.
- **Standard:** N/A.
- **Action:** Code fix combined with finding [Important] above.
- **Rationale:** Combined with the `_test.go` exclusion above, this would be addressed by filtering `_test.go` files from the per-file loop.
- **Suggested fix:** See [Important] above.

### [Improvement] No commit-status check that the `bench.txt` baseline is committed when a regression PR lands
- **File:** N/A — process gap
- **Phase introduced:** Phase 1
- **Issue:** `bench.txt` is checked in (per CLAUDE.md `bench.txt IS committed`). The PR template line 69 reminds contributors to run `make bench-compare` and explain regressions, but there is no automated check that `bench.txt` has been updated when a perf-affecting change lands. `bench-compare-informational` runs but per D-09 is non-blocking.
- **Standard:** CLAUDE.md `benchstat regression > 10% fails CI` (planned, post-self-hosted-runner).
- **Action:** Discuss-phase needed.
- **Rationale:** The post-self-hosted-runner plan covers this; until then a manual gate via PR review is fit-for-purpose. Worth recording explicitly as a Deferred Item in CONTEXT.
- **Suggested fix:** Track in 01-CONTEXT.md Deferred Items (likely already covered by D-09 — verify).

### [Improvement] `release.yml` `if: startsWith(github.ref, 'refs/tags/v')` is redundant defensive guard
- **File:** `/Users/johnny/Development/fuzzymatch/.github/workflows/release.yml`
- **Phase introduced:** Phase 1
- **Issue:** The workflow's `on:` is already `push: tags: ["v*"]` so the workflow only ever runs on `v*` tags. Per-step `if: startsWith(github.ref, 'refs/tags/v')` (lines 66, 76, 84, 91, 97) is defensive redundancy. Harmless but verbose.
- **Standard:** N/A.
- **Action:** Improvement (style).
- **Rationale:** Either remove the guards (the trigger already enforces it) or keep them as documentation. Current preference reads as "belt and braces" against a future trigger expansion; that is a reasonable defensive posture but the comment block at the top of the file already promises "TAG PUSH ONLY" with a strong note.
- **Suggested fix:** Optional — remove the `if:` guards or keep them with a brief inline comment ("redundant defence-in-depth against future trigger drift").

### [Improvement] `release.yml` does not pin the `cosign verify-blob` certificate-identity to the release workflow path
- **File:** `/Users/johnny/Development/fuzzymatch/.github/workflows/release.yml`
- **Phase introduced:** Phase 1
- **Issue:** Lines 99–103: post-release sanity check uses `--certificate-identity-regexp 'https://github.com/axonops/fuzzymatch/.+'`. This regex matches any workflow in the repo, including non-release workflows that might (in future, hypothetically) also sign blobs. The tighter constraint is to match `.../release.yml@refs/tags/v.*` only.
- **Standard:** Sigstore best practice — narrowest workable identity match.
- **Action:** Code fix.
- **Rationale:** Defence-in-depth against a future workflow that signs different blobs being mistakenly trusted as the release signature.
- **Suggested fix:** Change regex to `--certificate-identity-regexp 'https://github.com/axonops/fuzzymatch/\.github/workflows/release\.yml@refs/tags/v.*'`. Same regex should be advertised in SECURITY.md (line 71) for consumer verification.

### [Improvement] `Makefile check` does not run `verify-determinism`
- **File:** `/Users/johnny/Development/fuzzymatch/Makefile`
- **Phase introduced:** Phase 1
- **Issue:** Line 34: `check: fmt-check vet lint verify-license-headers verify-deps-allowlist tidy-check security test coverage coverage-check`. `verify-determinism` is run separately in CI (ci.yml line 80), not part of `make check`. CLAUDE.md `Makefile Review: make check mirrors CI exactly` — but CI runs `make verify-determinism` then `make check`, so `make check` alone is not a complete mirror.
- **Standard:** CLAUDE.md `make check mirrors CI exactly`.
- **Action:** Code fix.
- **Rationale:** Local-development parity. A developer running `make check` should get the same exit-state signal that CI provides on the same SHA. Including `verify-determinism` in `check` adds ~5s and provides much stronger local feedback.
- **Suggested fix:** Add `verify-determinism` to the `check` target's dependency list.

### [Improvement] `Makefile` `tidy-check` runs `go mod tidy` (which mutates) before diffing
- **File:** `/Users/johnny/Development/fuzzymatch/Makefile`
- **Phase introduced:** Phase 1
- **Issue:** Lines 165–168: `tidy-check` runs `go mod tidy` in both modules, then `git diff --exit-code` — this works but mutates the working tree mid-build. If `tidy-check` fails, the working tree is left dirty (the tidy edits are not reverted). A pure-read alternative is `go mod tidy -diff` (Go 1.23+) which prints the diff without applying it.
- **Standard:** Go module hygiene.
- **Action:** Improvement.
- **Rationale:** Read-only check matches the spirit better and leaves no side-effects on failure.
- **Suggested fix:** Use `go mod tidy -diff` in both modules and fail on non-empty diff. Go 1.26 supports this.

---

## Files reviewed (paths)

- `/Users/johnny/Development/fuzzymatch/.github/workflows/ci.yml`
- `/Users/johnny/Development/fuzzymatch/.github/workflows/release.yml`
- `/Users/johnny/Development/fuzzymatch/.github/workflows/security.yml`
- `/Users/johnny/Development/fuzzymatch/.github/workflows/codeql.yml`
- `/Users/johnny/Development/fuzzymatch/.github/workflows/commitlint.yml`
- `/Users/johnny/Development/fuzzymatch/.github/workflows/cla.yml`
- `/Users/johnny/Development/fuzzymatch/.github/workflows/license-headers.yml`
- `/Users/johnny/Development/fuzzymatch/.github/dependabot.yml`
- `/Users/johnny/Development/fuzzymatch/.github/CODEOWNERS`
- `/Users/johnny/Development/fuzzymatch/.github/PULL_REQUEST_TEMPLATE.md`
- `/Users/johnny/Development/fuzzymatch/Makefile`
- `/Users/johnny/Development/fuzzymatch/.goreleaser.yml`
- `/Users/johnny/Development/fuzzymatch/.golangci.yml`
- `/Users/johnny/Development/fuzzymatch/.markdownlint-cli2.yaml`
- `/Users/johnny/Development/fuzzymatch/.commitlintrc.yml`
- `/Users/johnny/Development/fuzzymatch/.gitignore`
- `/Users/johnny/Development/fuzzymatch/scripts/verify-license-headers.sh`
- `/Users/johnny/Development/fuzzymatch/scripts/verify-no-runtime-deps.sh`
- `/Users/johnny/Development/fuzzymatch/scripts/verify-coverage-floors.sh`
- `/Users/johnny/Development/fuzzymatch/SECURITY.md`
- `/Users/johnny/Development/fuzzymatch/ai_friendly_test.go`
- `/Users/johnny/Development/fuzzymatch/go.mod`
- `/Users/johnny/Development/fuzzymatch/tests/bdd/go.mod`
- `/Users/johnny/Development/fuzzymatch/llms.txt`
- `/Users/johnny/Development/fuzzymatch/CHANGELOG.md`

## Notable positives (not findings — recorded for completeness)

- Release flow is well-designed: tag-push-only trigger; explicit comment block forbidding any other trigger; cosign `--bundle` (required in v3); separate `gh release upload` for the bundle as defence-in-depth against goreleaser's `extra_files` block; post-release `cosign verify-blob` sanity check; OIDC build-provenance attestation via `actions/attest-build-provenance@v2`.
- `.goreleaser.yml` is library-aware: `builds: [{skip: true}]` correctly tells GoReleaser there are no binaries to compile; archives produce a source tarball only; SBOM via Syft is wired into the `sboms:` block correctly.
- `permissions: contents: read` at workflow level on every workflow except `cla.yml` (which needs writes for legitimate reasons that are well-documented inline).
- Per-job permission escalation in `release.yml` lists each scope with a justification inline.
- `verify-no-runtime-deps.sh` is well-thought-through: it filters indirect modules out of the check (with a strong comment block explaining why), includes a positive "must contain these allowlist entries" check to catch accidental removal of `x/text`, and is bash-3.2 compatible for macOS-default-shell parity.
- `verify-coverage-floors.sh` enforces the three CLAUDE.md floors (95% overall, 90% per file, 100% public-API funcs via a parser of `go doc -short`) with a bootstrap-tolerance branch.
- `ai_friendly_test.go` enforces llms.txt sync via `go/parser.ParseDir` walking the root package — durable structural check.
- `golangci-lint` v2 configuration with the formatters/linters split; `goimports` configured with `local-prefixes: github.com/axonops/fuzzymatch` for the canonical import grouping.
- `commitlintrc.yml` enforces Conventional Commits with subject-case=lower-case, subject-empty=never, header-max-length=72 — strong baseline.
- CLA workflow uses `pull_request_target` correctly with the comment-text guard and a documented allowlist for bot accounts.
- Dependabot configured for three ecosystems (root gomod, tests/bdd gomod, github-actions) with sensible grouping.
- Markdown linting active on `**/*.md` with `.planning/` and `.claude/` correctly excluded.
- The 5-platform CI matrix is complete and correctly handles the macOS Intel runner deprecation (using `macos-15-intel` after `macos-13` retirement).
