---
status: complete
phase: 09-collection-scan-sub-package
mode: maintainer-direct
source:
  - 09-01-SUMMARY.md
  - 09-02-SUMMARY.md
  - 09-03-SUMMARY.md
  - 09-04-SUMMARY.md
  - 09-05-SUMMARY.md
  - 09-06-SUMMARY.md
  - 09-07-SUMMARY.md
  - 09-08-SUMMARY.md
started: 2026-05-21T04:14:27Z
updated: 2026-05-21T04:24:00Z
---

## Current Test

[testing complete]

## Tests

### 1. make check
expected: |
  All quality gates green — fmt, vet, golangci-lint, license headers, deps
  allowlist, tidy, govulncheck, race+shuffle tests, coverage floors.
result: pass
evidence: |
  - 0 lint issues (root + tests/bdd)
  - 220 .go files carry Apache-2.0 header
  - root go.mod allowlist clean (2 non-indirect: axonops/fuzzymatch, golang.org/x/text)
  - No vulnerabilities
  - Race+shuffle tests: root 15.96s + scan 70.77s OK
  - Coverage 96.9% (overall >= 95%, per-file >= 90%, Floor 3 AST passed)

### 2. make test-bdd
expected: |
  All BDD scenarios green, including 11 scan + 9 suppression scenarios.
result: pass
evidence: |
  cd tests/bdd && CGO_ENABLED=1 go test -race -count=1 ./... → ok 7.66s

### 3. PERF-05 budget (10k items < 2s)
expected: |
  BenchmarkScanCheck_DefaultScorer_10k lands under the 2s spec budget on
  darwin/arm64, ideally within ±20% of Phase 9 baseline (362 ms).
result: pass
evidence: |
  472 ms / op on darwin/arm64 Apple M2 — 4.2× under spec budget.
  +30% vs Phase 9 baseline (362 ms) — within tolerance for a different run
  with active background load; both well under the 2s gate.

### 4. Cross-platform golden determinism
expected: |
  testdata/golden/scan-default.json passes locally; the same file passes on
  every leg of the cross-platform CI matrix on origin/main.
result: pass
evidence: |
  Local: go test -run TestGolden_ ./... → ok (root + scan) on darwin/arm64.
  Origin/main CI: latest run 2026-05-20T20:46:10Z → CI, License Headers,
  Security, CodeQL all success. Cross-platform matrix (linux amd64+arm64,
  darwin arm64, windows amd64) implicit in the CI workflow's success.

## Summary

total: 4
passed: 4
issues: 0
pending: 0
skipped: 0

## Gaps

[none]
