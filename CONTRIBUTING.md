# Contributing to fuzzymatch

## Welcome / Scope

fuzzymatch is in pre-release (v0.x). During pre-release we welcome
**issues** — bug reports, algorithm proposals, API ergonomics
feedback, documentation gaps. **Pull requests from external
contributors** are accepted after the v1.0.0 API freeze; until then,
all code changes route through the AxonOps maintainer team to keep
the catalogue, scorer, and scan sub-package converging on the
locked-in v1.0.0 contract.

The canonical workflow source is [`CLAUDE.md`](CLAUDE.md). Anything in
this file that conflicts with CLAUDE.md is a bug — CLAUDE.md wins,
please open an issue so we can reconcile.

## Pre-PR Checklist

Before opening a PR (post-v1.0.0):

1. `make check` exits 0 on a clean tree (lint, vet, race-shuffled
   tests, govulncheck, license headers, deps allowlist, tidy, coverage
   floors).
2. Commit messages follow [Conventional Commits](https://www.conventionalcommits.org/)
   — see [`.claude/skills/commit-standards/SKILL.md`](.claude/skills/commit-standards/SKILL.md).
3. The PR is signed via the CLA Assistant workflow
   ([`.github/workflows/cla.yml`](.github/workflows/cla.yml)). Signing
   is automatic — the bot will comment on the PR.
4. A CHANGELOG entry has been added under `## [Unreleased]` in
   [`CHANGELOG.md`](CHANGELOG.md).
5. If the PR touches algorithm code, `make bench-compare` has been
   run locally and any regression > 10% is explained in the PR
   description.
6. The PR template's fields are filled out — including the Source
   Origin Statement section for algorithm PRs.

## Local Development Setup

Required tools (versions pinned to match CI):

| Tool | Version | Install |
|------|---------|---------|
| Go | 1.26.3 (or newer 1.26.x) | https://go.dev/dl |
| `golangci-lint` | v2.12.2 | `go install github.com/golangci/golangci-lint/v2/cmd/golangci-lint@v2.12.2` |
| `govulncheck` | latest | `go install golang.org/x/vuln/cmd/govulncheck@latest` |
| `goimports` | latest | `go install golang.org/x/tools/cmd/goimports@latest` |
| `benchstat` (optional) | latest | `go install golang.org/x/perf/cmd/benchstat@latest` |
| `markdownlint-cli2` (optional) | v0.22.x | `npm install -g markdownlint-cli2@0.22` |

The full stack rationale is in [`.planning/research/STACK.md`](.planning/research/STACK.md).

Clone, verify, and run the test suite:

```bash
git clone https://github.com/axonops/fuzzymatch.git
cd fuzzymatch
make check
```

## Make Targets

Every target listed here is enforced by the `makefile_targets_test.go`
meta-test — new targets MUST be added to this section AND to the
canonical list inside the test. Removing a target requires a matching
removal here.

- `make check` — full quality gate (fmt-check, vet, lint,
  verify-license-headers, verify-deps-allowlist, tidy-check, security,
  test-race, coverage, verify-determinism).
- `make test` — unit + property + meta tests (`go test -race -shuffle=on -count=1`).
- `make test-bdd` — godog BDD scenarios from the `tests/bdd/` sub-module.
- `make test-fuzz` — short fuzz run (60s per fuzzer).
- `make lint` — `golangci-lint` v2 (standard preset + project opt-ins).
- `make vet` — `go vet ./...` (kept separate from `make lint` for
  clean signal attribution).
- `make fmt` — auto-fix formatting (`gofmt -s` + `goimports -local
  github.com/axonops/fuzzymatch`).
- `make fmt-check` — check-only formatting; fails on any drift.
- `make bench` — run all benchmarks; writes to stdout.
- `make bench-compare` — compare current bench results against the
  committed `bench.txt` via `benchstat`; flags any regression > 10%.
- `make coverage` — generate coverage profile + HTML report.
- `make tidy` — `go mod tidy` for root and `tests/bdd/`.
- `make tidy-check` — verify `tidy` produces no diff (CI gate).
- `make security` — `govulncheck` + `gosec`.
- `make verify-deps-allowlist` — verify root `go.mod` requires only
  `golang.org/x/text` (the single curated runtime dep).
- `make verify-determinism` — run the determinism golden test on the
  current platform; compares against committed fixtures.
- `make verify-license-headers` — verify every `.go` file carries
  the Apache-2.0 + AxonOps copyright header.
- `make regen-swg-cross-validation` — developer-only; regenerate
  `testdata/cross-validation/swg/vectors.json` from biopython's
  `Bio.Align.PairwiseAligner` via `scripts/gen-swg-cross-validation.py`.
  Requires `python3 -m pip install --user biopython` (1.85+). NOT
  part of `make check`; CI consumes the committed JSON via
  `TestSWG_CrossValidation` and does not require Python.
- `make regen-ratcliff-obershelp-cross-validation` — developer-only;
  regenerate `testdata/cross-validation/ratcliff-obershelp/vectors.json`
  from Python stdlib `difflib.SequenceMatcher(autojunk=False).ratio()`
  via `scripts/gen-ratcliff-obershelp-cross-validation.py`. Requires
  Python 3.7+ (difflib is stdlib — no `pip install` needed). NOT
  part of `make check`; CI consumes the committed JSON via
  `TestRatcliffObershelp_CrossValidation` and does not require Python.
- `make release-check` — validate `.goreleaser.yml` parses; never
  invokes a release locally (releases ship via CI only — see below).
- `make clean` — clear test cache and coverage outputs.

## Conventional Commits

All commits MUST follow the Conventional Commits 1.0 specification —
see [`.claude/skills/commit-standards/SKILL.md`](.claude/skills/commit-standards/SKILL.md)
for the exact rules and our `(scope)` conventions (per-algorithm
scopes for algorithm work; `(scorer)`, `(scan)`, `(normalise)` etc.
for primitives; `(ci)`, `(docs)`, `(chore)` for repo-wide changes).

**No AI attribution in commit messages.** Commits represent the
project's work; do not mention Claude, GPT, Copilot, or any other
AI tool in the commit message body or footer.

The `commitlint.yml` GitHub Actions workflow enforces this on PR
titles and on every commit in the PR.

## Algorithm Contribution Flow

Algorithm additions go through extra gates because we ship from
primary academic sources and screen for patent encumbrance.

1. Open an issue using the
   [`algorithm-proposal.yml`](.github/ISSUE_TEMPLATE/algorithm-proposal.yml)
   template. Include the primary academic source citation, patent
   screen status, and at least 3 input → expected-output reference
   vectors with attribution.
2. The `algorithm-licensing-reviewer` agent screens the proposal for
   patent encumbrance and licence compatibility BEFORE implementation
   begins.
3. Once licence-cleared, the `algorithm-correctness-reviewer` agent
   verifies the primary-source citation, the recurrence/formula in the
   issue's reference vectors, and the proposed constants.
4. Implementation is written **fresh from the primary source**.
   MIT/BSD-licensed Go implementations MAY be consulted for
   reference-vector cross-validation but **never** for code copying.
   GPL/LGPL implementations MUST NOT be consulted at all.
5. The PR includes the Source Origin Statement section of the PR
   template, listing the primary source, any implementations studied
   for vectors (with their licences), and explicit confirmations of
   no-code-copying and no-GPL-consultation.
6. All review gates pass — see [`CLAUDE.md`](CLAUDE.md) "Workflow —
   Agent Gates" for the full list (algorithm-correctness,
   algorithm-performance, determinism, api-ergonomics,
   code-reviewer).

## Algorithm Deprecation Policy

> Within a major version, algorithms may be **ADDED** but never
> **REMOVED**. Score-changing edits to an existing algorithm require
> a minor version bump and a CHANGELOG entry. Bug fixes that change
> scores are minor; intentional algorithm-formula changes are minor;
> algorithm removals are major (v2.x.y or later).

This is requirement REL-07 from
[`docs/requirements.md`](docs/requirements.md) §11.2. Reviewers
enforce this during PR review; the release process refuses tags
that violate it.

## Release Process

**Releases happen exclusively via CI on tag push.** Maintainers do
not run `git tag` locally and do not invoke `goreleaser release`
locally.

The supported flows are:

1. **GitHub UI release creation** — Maintainer creates a new release
   in the GitHub UI, GitHub creates the tag, and the
   [`.github/workflows/release.yml`](.github/workflows/release.yml)
   workflow runs goreleaser, cosign, Syft, and attest-build-provenance.
2. **Tag-push from a release PR merge** — A release PR bumps version
   and CHANGELOG; merging it triggers a tag-push workflow that
   creates the tag, which in turn triggers `release.yml`.

Either flow guarantees the release is produced inside CI with the
signing key from Fulcio, the provenance attestation from GitHub OIDC,
and the SBOM from Syft. `make release-check` only validates the
goreleaser config; it never publishes.

## Benchmark Discipline

`bench.txt` is committed at the repo root and is the
benchstat-comparable baseline for the current `main`. Contributors
touching algorithm code MUST:

1. Run `make bench` locally before submitting the PR.
2. Run `make bench-compare` locally and report any regression > 10%
   in the PR description, with a rationale (e.g. "added correctness
   fix that traded 5% throughput for 100% IEEE-754 correctness").

CI runs `make bench-compare` informationally until a self-hosted
runner with stable hardware is provisioned. The 10%-regression
threshold is documented in
[`docs/performance.md`](docs/performance.md) and in
[`.claude/skills/performance-standards/SKILL.md`](.claude/skills/performance-standards/SKILL.md).
