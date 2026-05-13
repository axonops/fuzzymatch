---
status: partial
phase: 01-foundation-infrastructure
source: [01-VERIFICATION.md]
started: 2026-05-13T16:35:00Z
updated: 2026-05-13T16:35:00Z
---

## Current Test

[awaiting human testing]

## Tests

### 1. Simulated tag-push to a fork produces a signed release
expected: GoReleaser runs in CI; checksums.txt is cosign-signed; Syft SPDX-JSON SBOM is uploaded; actions/attest-build-provenance@v2 publishes an OIDC attestation; cosign verify-blob succeeds inside the workflow before exit
result: [pending]

### 2. CI matrix runs green on all five GitHub-hosted runners with byte-identical golden file
expected: All five matrix jobs (linux-amd64, linux-arm64, darwin-amd64 on macos-15-intel, darwin-arm64 on macos-latest, windows-amd64) pass; testdata/golden/normalisation.json diffs cleanly on every entry across all five platforms
result: [pending]

### 3. GitHub branch protection on `main` requires canonical CI checks
expected: Branch protection is configured in repository settings so a PR cannot merge with red checks on ci / security / codeql / license-headers / commitlint / CLA Assistant
result: [pending]

### 4. CLA Assistant signing flow works end-to-end
expected: CLA.md exists in the repo; CLA_PAT secret is configured; a non-allowlisted contributor comments "I have read the CLA Document..." on a PR and CLA Assistant commits the signature to signatures/version1/cla.json
result: [pending]

### 5. markdownlint-cli2 passes on the docs locally and in CI
expected: `npx markdownlint-cli2 "**/*.md"` exits 0 outside `.planning/` and `.claude/`; the license-headers + linting workflows go green on a clean push
result: [pending]

## Summary

total: 5
passed: 0
issues: 0
pending: 5
skipped: 0
blocked: 0

## Gaps
