# Security Policy

fuzzymatch is a pure-Go library with no network exposure, no I/O, and no
shared mutable state. The realistic threat surface is:

- Algorithmic complexity DoS via pathological input (caught by fuzz tests
  and per-algorithm allocation/time budgets).
- Supply-chain attacks against `golang.org/x/text` (the only non-stdlib
  runtime dependency; mitigated by Dependabot, `govulncheck`,
  `gosec`, and the `verify-deps-allowlist` CI gate).
- Compromised maintainer credentials (mitigated by signed releases via
  cosign keyless + OIDC, plus GitHub branch protection on `main`).

Vulnerabilities **outside** that surface — e.g. RCE in the `unicode/norm`
transformer, undisclosed Go runtime bugs affecting determinism — are
relevant for downstream consumers and are reported as below.

## Supported Versions

| Version | Status | Notes |
|---------|--------|-------|
| v0.x.y | pre-release | no SLA; security fixes ship with the next pre-release tag |
| v1.x.y | (future) | will be supported per `docs/requirements.md` deprecation policy when v1.0.0 ships |

The deprecation policy for v1.x is documented in
[`CONTRIBUTING.md`](CONTRIBUTING.md) — within a major version,
algorithms may be added but not removed; scoring-changing edits require
a minor bump and a CHANGELOG entry.

## Reporting a Vulnerability

Report security issues **privately** to **security@axonops.com**.

Do NOT open public issues for vulnerabilities. The issue templates in
`.github/ISSUE_TEMPLATE/` deliberately omit a "security" type; the
private email is the only supported channel.

When reporting, please include:

- A concise description of the vulnerability.
- The fuzzymatch version (commit SHA, tag, or module version) affected.
- A minimal reproducer (Go program demonstrating the issue) if
  available.
- Your proposed CVSS score, if you have one.
- Whether you are willing to be credited, and the name / handle to use.

## Disclosure Timeline

We follow a 90-day coordinated-disclosure standard:

- **Within 2 business days** of report receipt: acknowledgement.
- **Within 7 business days**: initial assessment and severity triage.
- **Within 30 days**: fix or documented workaround for confirmed
  vulnerabilities. We may negotiate longer timelines for unusually
  complex issues; we will communicate explicitly if so.
- **90 days after initial report** (or 7 days after a fix ships,
  whichever is later): public disclosure via the project CHANGELOG and
  a GitHub Security Advisory.

If the vulnerability is being actively exploited in the wild, we
shorten the timeline to whatever is necessary to protect users.

## Verification of Released Artefacts

Released versions (v1.0.0 onwards) are signed via cosign keyless OIDC
through the release pipeline (see `.github/workflows/release.yml`).
Verify a release as follows:

```bash
cosign verify-blob \
    --bundle checksums.txt.bundle \
    --certificate-identity-regexp 'https://github.com/axonops/fuzzymatch/.+' \
    --certificate-oidc-issuer https://token.actions.githubusercontent.com \
    checksums.txt
```

The signed `checksums.txt` covers the source tarball and the Syft
SPDX-JSON SBOM. Both are GitHub release assets.

For consumers who pin via `go get`, the Go module proxy and checksum
database provide an independent integrity guarantee — the cosign
signature adds non-repudiation against the release pipeline.

## Security Tooling

The CI pipeline runs the following security tools on every PR and on a
weekly schedule:

- **`govulncheck`** (Go vulnerability database) — runs on every PR and
  weekly via `.github/workflows/security.yml`.
- **`gosec`** (static security analysis with SARIF upload to the
  GitHub Security tab) — runs on every PR.
- **CodeQL** (semantic security analysis) — runs on every PR and
  weekly via `.github/workflows/codeql.yml`.
- **Dependabot** — daily checks for the `gomod` and `github-actions`
  ecosystems with grouped PRs.

Findings are triaged via GitHub Security; critical findings block
merge.

## Out-of-scope Reports

The following are explicitly NOT vulnerabilities in fuzzymatch:

- "Algorithm X produces a low similarity score for strings that look
  similar to me." — That is a tuning issue, not a security issue. See
  [`docs/tuning.md`](docs/tuning.md).
- "The library does not implement Metaphone 3." — Deliberate exclusion
  due to U.S. Patent 7,440,941. See
  [`docs/faq.md`](docs/faq.md#why-no-metaphone-3).
- "The library has a non-stdlib runtime dep." — `golang.org/x/text` is
  the only one, locked by `verify-deps-allowlist`. See
  [`docs/faq.md`](docs/faq.md#why-xtext-but-no-other-deps).

If you are uncertain whether your finding is in scope, send it to
`security@axonops.com` and we will triage.
