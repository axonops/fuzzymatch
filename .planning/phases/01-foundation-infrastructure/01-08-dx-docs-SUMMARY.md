---
status: complete
phase: 01-foundation-infrastructure
plan: 01-08-dx-docs
date: 2026-05-13
tasks_completed: 3
tasks_total: 3
deviations: 3
---

# Plan 01-08: DX Docs — Summary

## Outcome

Consumer-facing and contributor-facing documentation now sits on top
of the foundation built in plans 01-01 through 01-07. The project is
shippable as a v0.x pre-release artefact: a developer landing on the
repository can read the README, navigate to per-area docs scaffolds,
follow CONTRIBUTING for local setup, file a structured issue, open a
PR that carries a Source Origin Statement, and verify a future v1.x
release with cosign — all without reading the source.

The two meta-tests added in Task 2 keep this documentation honest in
CI: `TestAIFriendly_LLMSContainsAllExports` parses the root package
AST and asserts every exported symbol is referenced in `llms.txt`;
`TestMakefile_HasCanonicalTargets` + `TestMakefile_TargetsDocumentedInContributing`
enforce bi-directional Makefile ↔ CONTRIBUTING.md coverage.

## Commits

- `3da1309` — `docs(01-08): mask-polish README, add llms.txt + llms-full.txt, ai_friendly_test.go`
- `5817a0e` — `docs(01-08): add docs/* scaffolds (23-algo catalogue, FAQ, Scorer/scan/tuning/extending/performance) + makefile_targets_test.go`
- `9087a31` — `docs(01-08): add SECURITY, CoC, CONTRIBUTING, CODEOWNERS, issue + PR templates`
- `74b028f` — `fix(01-08): gofmt list-comment indentation and silence SA1019 in ai_friendly_test`

## Files Created or Modified

### Task 1 (consumer documentation)
- `README.md` — full rewrite: 23-row algorithm catalogue table grouped
  by category with per-row primary-source citation and
  `docs/algorithms.md` anchor; Quick start; Configuration; Thread
  safety; Why-this-library-exists; For AI Assistants block.
- `llms.txt` — concise AI-friendly index of every exported root-package
  symbol.
- `llms-full.txt` — full per-symbol signatures, per-AlgoID primary-source
  citation table, per-error message text, per-options-struct field
  tables, sample programs, determinism contract, threat surfaces.
- `ai_friendly_test.go` — documentation-drift gate; parses root
  package via `go/parser.ParseDir` + `go/ast`, asserts each exported
  symbol appears verbatim in `llms.txt`.

### Task 2 (docs/* scaffolds + makefile meta-test)
- `docs/algorithms.md` — H1 + 23 H2 sections (one per AlgoID), each
  with name + algorithm class + primary-source citation + target phase.
- `docs/scorer.md`, `docs/scan.md`, `docs/extending.md`,
  `docs/tuning.md`, `docs/performance.md`, `docs/faq.md` — scaffolds
  with H2 anchors. `faq.md` carries the six DX-06 mandatory entries
  (Needleman-Wunsch, Metaphone 3, embeddings/ML, phonetic-as-binary,
  generics, x/text).
- `makefile_targets_test.go` — Makefile ↔ CONTRIBUTING bi-directional
  drift gate.

### Task 3 (policy + templates)
- `SECURITY.md` — H2 sections per plan; `cosign verify-blob` recipe;
  90-day coordinated disclosure; out-of-scope list.
- `CODE_OF_CONDUCT.md` — Contributor Covenant 2.1 adoption by
  reference (deviation 1; see below).
- `CONTRIBUTING.md` — pre-PR checklist, local setup matrix, all 19
  canonical Makefile targets, conventional commits with no-AI-attribution
  rule, algorithm contribution flow, REL-07 deprecation policy,
  CI-only release process, benchmark discipline.
- `.github/CODEOWNERS` — single `@axonops/fuzzymatch-maintainers`
  routing.
- `.github/ISSUE_TEMPLATE/bug.yml`, `feature.yml`, `algorithm-proposal.yml`
  — GitHub issue-form schema; algorithm-proposal forces explicit
  patent-screen status and reference vectors with attribution.
- `.github/PULL_REQUEST_TEMPLATE.md` — Source Origin Statement section
  for algorithm PRs; mathematical-invariants checklist; review-gate
  awareness list per CLAUDE.md.
- `Makefile` — `## suppress:` comment on internal `coverage-check`
  target (CI-only helper; not user-facing).

## Verification

`make check` exits 0 end-to-end on the plan-level verification:

```
fmt-check, vet, golangci-lint (root + tests/bdd), verify-license-headers,
verify-no-runtime-deps, mod tidy, govulncheck, test -race -shuffle=on,
coverage 96.7% with all floors green, verify-coverage-floors.
```

Plan acceptance checks individually verified:

- All 8 Task-3 files exist; issue templates parse as valid YAML.
- `SECURITY.md` has the 5 required H2 sections; `cosign verify-blob`
  command is present verbatim.
- `CODE_OF_CONDUCT.md` references Contributor Covenant 2.1 (by URL
  rather than full text — see deviation 1).
- `CONTRIBUTING.md` references `make check`, `make bench-compare`,
  Conventional Commits, the algorithm deprecation policy, and the
  CI-only release rule.
- `TestMakefile_HasCanonicalTargets` + `TestMakefile_TargetsDocumentedInContributing` pass.
- `TestAIFriendly_LLMSContainsAllExports` passes.
- 23-row algorithm catalogue table present in README.
- `docs/faq.md` has 6 H2 sections covering the DX-06 mandatory list.

## Deviations

### 1. Reference-style CoC instead of verbatim text — Rule 3 (auto-fix)

The plan called for the Contributor Covenant 2.1 text verbatim. The
verbatim text contains an enumeration of unacceptable behaviours
(harassment, abuse, etc.) that triggered the API content filter
during executor output. **Resolution:** `CODE_OF_CONDUCT.md` adopts
Contributor Covenant 2.1 by reference, links to both the HTML and
plain-text canonical URLs, names the conduct channel
(`conduct@axonops.com`), and includes the Attribution clause per
Covenant's licence terms. This is a recognised pattern (many large
projects, including Kubernetes for years, used a reference-style
CoC). The behavioural contract is identical; only the in-repo
verbatim copy is replaced by a link.

This deviation is documented here for downstream consumers; the
substantive licence (CC BY 4.0) obligations of the Covenant are
satisfied because the project explicitly cites and links the upstream.

### 2. `parser.ParseDir` deprecation silenced with rationale — Rule 1

The `ai_friendly_test.go` meta-test (committed in Task 1) used
`go/parser.ParseDir`, which staticcheck flags as SA1019 since Go
1.25. The recommended replacement (`golang.org/x/tools/go/packages`)
would add a build-tools dep to test-only code. **Resolution:** added
`//nolint:staticcheck` with an inline rationale citing the
single-package, no-build-tag scope. The meta-test continues to do
what it was designed to do.

### 3. `coverage-check` Makefile target suppressed from CONTRIBUTING — Rule 1

`coverage-check` is invoked from `make check` and not directly by
contributors; the canonical 19-target list in CLAUDE.md does not
include it. **Resolution:** added `## suppress: internal helper`
comment above its rule in the Makefile so the bi-directional
meta-test passes without forcing CONTRIBUTING.md to expose an
internal CI-only target.

## Issues Encountered

- **API content-filter on Contributor Covenant verbatim text.**
  Captured as deviation 1. The reference-style alternative was the
  pragmatic and policy-equivalent fix.
- **Executor agent terminated mid-Task-3 with content-filter error.**
  The first two Tasks committed cleanly; Task 3 was completed
  inline by the orchestrator from the executor's draft SECURITY.md
  plus the remaining files written against the plan spec.

## Follow-ups

- Replace `@axonops/fuzzymatch-maintainers` in CODEOWNERS with the
  real GitHub team handle when the team is provisioned.
- Confirm `security@axonops.com` and `conduct@axonops.com` are
  monitored mailboxes; update both files if the canonical addresses
  differ.
- Branch protection on `main` is a one-time GitHub UI setup; not
  automatable from this plan.
- `CLA_PAT` repository secret is required for the CLA Assistant
  workflow (added in plan 01-03) to commit signatures; not set yet.
- The remaining `docs/*.md` scaffolds (`scorer.md`, `scan.md`,
  `extending.md`, `tuning.md`, `performance.md`) are placeholder-only
  bodies; phases 5, 6, and follow-ons populate them.

## Self-Check: PASSED
