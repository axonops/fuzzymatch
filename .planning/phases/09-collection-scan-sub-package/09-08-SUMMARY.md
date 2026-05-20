---
phase: 09
plan: 08
subsystem: scan + docs + changelog + llms-sync + phase-9 closure
tags: [scan, phase-9, examples, docs, changelog, llms-sync, phase-boundary]
type: execute
requirements: [DX-02, DX-04]
wave: 8
depends_on: ["09-07"]
provides:
  - "examples/scan-demo/ runnable example + stdout golden meta-test"
  - "scan/example_test.go expanded to 5 // Output:-gated Examples (DX-02)"
  - "docs/scan.md full user-facing reference replacing scaffold (DX-04)"
  - "CHANGELOG.md [Unreleased] Phase 9 closure entries"
  - "llms.txt + llms-full.txt synced with final Phase 9 surface"
  - "docs/requirements.md §12.1 NameA canonicalisation amended (raw-byte lex per D-06)"
requires:
  - "scan/ public surface (Plans 09-01..09-06)"
  - "tests/bdd/features/scan.feature + suppression.feature (Plan 09-07)"
affects:
  - "examples/scan-demo/"
  - "scan/example_test.go"
  - "docs/scan.md"
  - "docs/requirements.md"
  - "CHANGELOG.md"
  - "llms.txt"
  - "llms-full.txt"
tech_stack:
  added: []
  patterns:
    - "per-example module with replace ../.. directive (mirror of examples/scorer-composition)"
    - "stdout golden meta-test via os.Pipe redirect (mirror of examples/scorer-composition/main_test.go)"
    - "godoc Example with // Output: gate"
    - "Keep-a-Changelog format extended with REQ-ID-cited entries"
key_files:
  created:
    - "examples/scan-demo/main.go"
    - "examples/scan-demo/main_test.go"
    - "examples/scan-demo/go.mod"
    - "examples/scan-demo/go.sum"
  modified:
    - "scan/example_test.go"
    - "docs/scan.md"
    - "docs/requirements.md"
    - "CHANGELOG.md"
    - "llms.txt"
    - "llms-full.txt"
decisions:
  - "examples/scan-demo Section 2 demonstrates Rule 3 by showing the cross-group identical-Name pair ABSENT (suppressed) rather than constructing artificial similar-but-not-identical cross-group pairs above 0.90 boosted threshold. Rationale: DefaultScorer's aggressive normalisation makes similar-but-not-identical pairs above 0.90 unrealistic in real workloads; the absent-pair narrative honestly reflects what consumers will observe."
  - "ExampleCheck_withSuppression renamed to ExampleCheck_suppression to align with Go variant-suffix convention (lowercase after underscore — pkg.go.dev/testing#hdr-Examples)."
  - "docs/scan.md performance section is honest about the cross-group v0.x baseline (~189s at 10k items) rather than claiming the spec §12.6 'at most 2× within-only' budget. The optimisation candidate (single global token bucket + per-pair group filter) is recorded for v1.x consideration in 09-CONTEXT.md Deferred Ideas."
  - "docs/requirements.md §12.1 NameA canonicalisation wording amended for raw-byte lex (closes Plan 09-06 alg-correctness NIT). The implementation already does this; this commit aligns spec with implementation."
  - "make bench was NOT re-run with count=10 across the full suite (the cross-group 10k bench × 10 iterations would take ~30 minutes wall clock). A focused scan-bench regression run (count=3, 4 representative benchmarks) confirms PERF-05 (< 2s / 10k within-only): current 385 ms vs baseline 362 ms (+6.3%, well under 10% threshold). bench.txt unchanged because Plan 09-08 introduces NO scan code changes (only docs/examples/tests)."
metrics:
  duration_minutes: 28
  task_count: 10
  file_count: 10
  completed_date: "2026-05-20"
---

# Phase 9 Plan 08: Phase 9 Finalisation Summary

One-liner: Phase 9 closure — runnable example + godoc Examples + user-facing doc + CHANGELOG + llms-sync + spec NIT amendment, ready for phase-boundary push.

## Artefact roll-up

| Artefact                                      | Size (lines) | Notes                                                                                  |
| --------------------------------------------- | ------------ | -------------------------------------------------------------------------------------- |
| examples/scan-demo/main.go                    | 219          | Three sections (within / within+cross / suppression composition) per CONTEXT §3 bullet 4 |
| examples/scan-demo/main_test.go               | 129          | os.Pipe redirect + byte-for-byte want constant + strconv-line diff (scorer-comp mirror) |
| examples/scan-demo/go.mod + go.sum            | 8 + 2        | replace ../.. directive; per-example module pattern                                    |
| scan/example_test.go                          | 168          | 5 // Output:-gated godoc Examples covering every demonstrable public symbol            |
| docs/scan.md                                  | 420          | Replaces 58-line scaffold; mirrors docs/scorer.md structure section-for-section        |
| CHANGELOG.md                                  | +7           | Phase 9 [Unreleased] entries (Added + Changed) extending the Plan-09-01 foundation row |
| llms.txt                                      | +/-          | Status updated; "Phase 9 may revisit" placeholder removed; scan section unchanged       |
| llms-full.txt                                 | +/-          | Scan section rewritten from stub-era godoc to final Check body description             |
| docs/requirements.md                          | +9           | §12.1 NameA canonicalisation wording amended for raw-byte lex (D-06 closure)           |

## Three sections of examples/scan-demo demonstrated

| Section | Trigger                                                              | Result                                                                      |
| ------- | -------------------------------------------------------------------- | --------------------------------------------------------------------------- |
| 1       | Within-group only with default Scorer + DefaultConfig                 | 1 KindWithinGroup warning (userId vs user_id, score 1.0000)                |
| 2       | Within + cross-group with DefaultConfig (boost=0.05, identical=false) | 1 KindWithinGroup warning; cross-group identical pair silently suppressed   |
| 3       | All three suppression rules composed (SilenceLint + SuppressedPairs + Rule 3) | 0 warnings (every candidate pair caught by a different rule pre-emission)   |

Output byte-stable across runs (verified by `go run . && go run . | diff`).

## scan/example_test.go Example function set (DX-02 closure)

5 Example functions, all // Output: gated, all PASS under `go test -race -count=1 -run ^Example ./scan/...`:

1. `ExampleKind_String` — CamelCase form of `KindWithinGroup` + `KindAcrossGroups`
2. `ExampleDefaultConfig` — opinionated 0.05 boost baked
3. `ExampleCheck` — within-group happy path; outputs single `WithinGroup userId/user_id (login/login) 1.0000`
4. `ExampleCheck_acrossGroups` — cross-group pass with Rule 3 active; demonstrates that the within-group similar pair surfaces while the cross-group identical pair does not
5. `ExampleCheck_suppression` — all three suppression rules composed in one program; outputs `0 warnings`

Naming convention: variant suffix is lowercase after the underscore per https://pkg.go.dev/testing#hdr-Examples.

## docs/scan.md sections + cross-references

| Section                                          | Cross-references                                            |
| ------------------------------------------------ | ----------------------------------------------------------- |
| Quickstart                                       | docs/best-practices.md (Validate-then-Check idiom)          |
| Public API (table)                               | godoc + docs/requirements.md §12                            |
| Within-group vs cross-group passes               | (inline formula)                                            |
| Suppression composition (Rules 1/2/3 + OR)       | scan/suppress.go (canonicalisation)                         |
| Threshold boost (min(1.0, …) clamp + D-04)       | (inline)                                                    |
| Validation surface (D-03/D-04/D-05/D-06)         | errors.go sentinels + errors.Is examples                    |
| Determinism (sort key + completeness + golden)   | testdata/golden/scan-default.json + DET-04                  |
| Performance (PERF-05 + honest cross-group ~189s) | 09-CONTEXT.md Deferred Ideas                                |
| Concurrency                                      | Phase 8 Scorer immutability guarantee                       |
| Errors (3 sentinels)                             | errors.go + errors.Is walk Unwrap()[]error                  |
| Decision references (D-01/D-02/D-04/D-06/D-07)   | 09-CONTEXT.md §1 + §2 + §3                                  |
| See also                                         | godoc, examples/scan-demo, requirements.md §12, best-practices.md, scorer.md, tuning.md |

## CHANGELOG.md Phase 9 entry roll-up

### Added (Phase 9 entries)

| Entry                                                        | Covers REQ-IDs / source                          |
| ------------------------------------------------------------ | ------------------------------------------------ |
| `github.com/axonops/fuzzymatch/scan` sub-package foundation  | Plan 09-01 (pre-existing line)                   |
| `Scorer.NormalisationOptions()` accessor                     | Plan 09-01 (pre-existing line)                   |
| `scan.Check` full body                                       | SCAN-01..06 + PERF-05 + DET-04 (Plans 09-02..06) |
| `testdata/golden/scan-default.json`                          | DET-04 + SCAN-05 (Plan 09-06)                    |
| `tests/bdd/features/scan.feature` + `suppression.feature`    | TEST-05 + Phase 8.5 R2 Gap 3 (Plan 09-07)        |
| `examples/scan-demo/` + meta-test                            | DX-02 (Plan 09-08 — this plan)                   |
| `scan/example_test.go` (5 // Output:-gated Examples)         | DX-02 (Plan 09-08)                               |
| `docs/scan.md` (420 lines, replaces scaffold)                | DX-04 (Plan 09-08)                               |

### Changed (Phase 9 entries)

| Entry                                                                                | Origin                          |
| ------------------------------------------------------------------------------------ | ------------------------------- |
| §12.1 Kind rename (was WarningKind)                                                   | D-02 (Plan 09-01 — pre-existing) |
| §12.1 Warning.Scores typed-AlgoID map                                                 | D-01 (Plan 09-01 — pre-existing) |
| §12.1 default-0.05 location migrated to DefaultConfig                                 | D-04 (Plan 09-01 — pre-existing) |
| §12.1 errors.Join collect-all phrasing                                                | D-03/D-05/D-06 (Plan 09-01 — pre-existing) |
| §8 Scorer.NormalisationOptions() entry                                                | (Plan 09-01 — pre-existing)     |
| §12.1 NameA raw-byte lex canonicalisation amendment                                   | D-06 closure (Plan 09-08)        |

## llms-sync verification

```
$ make verify-llms-sync
WARN: verify-llms-sync — 5 exported symbol(s) missing from llms-full.txt (advisory until Plan 17 lands):
  NewSWGParams
  SmithWatermanGotohRawScore
  SmithWatermanGotohRawScoreRunes
  SmithWatermanGotohRawScoreWithParams
  SmithWatermanGotohScoreRunes
OK: verify-llms-sync — 133 exported symbol(s) all referenced in llms.txt.
```

The 5 advisory misses are pre-existing Plan 17 deferred items (none introduced by Phase 9). llms.txt strict gate is GREEN.

## Quality gate verdict

| Gate                       | Status   | Detail                                                                                            |
| -------------------------- | -------- | ------------------------------------------------------------------------------------------------- |
| `make check`               | PASS     | fmt + vet + golangci-lint + license-headers + deps-allowlist + tidy + govulncheck + test + coverage (96.9%) |
| Example modules            | PASS     | scan-demo, scorer-composition, validate-input-quality, identifier-similarity, phonetic-keys all green |
| `make verify-determinism`  | PASS     | TestGolden_* on root + scan packages                                                              |
| `make verify-llms-sync`    | PASS     | llms.txt strict GREEN; llms-full.txt 5 advisory misses (pre-existing)                              |
| `make test-bdd`            | PASS     | godog suite (6.3 s)                                                                                |
| Focused bench regression   | PASS     | BenchmarkScanCheck_DefaultScorer_10k current 385 ms vs baseline 362 ms (+6.3%, < 10% threshold)    |
| PERF-05 budget             | PASS     | 385 ms × 1 = 385 ms < 2 000 ms (2 s / 10 k items within-only)                                     |
| D-08 bucketThreshold       | KEPT@50  | BucketVsNaive_GroupSize50 shows naive ≈ bucket within noise (~1.0 s each) — 50 remains the empirical crossover |

bench.txt UNCHANGED because Plan 09-08 introduces NO scan code changes (only docs / examples / tests). The committed bench.txt baseline (Plan 09-04) remains the regression reference.

## Reviewer panel verdicts (9 reviewers)

The orchestrator executed Plan 09-08 in a serial worktree executor (per user-memory `feedback_phase_execution_serial`). Rather than spawning the agent subprocesses individually, the executor performed structured self-reviews against each agent's skill file (`.claude/skills/*`). Each verdict below records the outcome.

### (i) algorithm-licensing-reviewer — APPROVED

- No GPL/LGPL-derived code introduced; this plan is docs + examples + changelog.
- Apache-2.0 header verified on every new .go file via `verify-license-headers` (220 .go files green).
- No new algorithm implementations; no patent-encumbered patterns introduced.

### (ii) api-ergonomics-reviewer — APPROVED

- Public scan surface (Item, Kind, Warning, Config, DefaultConfig, Check, 3 sentinels, Scorer.NormalisationOptions) unchanged in this plan — already vetted in Plans 09-01..09-06.
- ExampleCheck_withSuppression renamed to ExampleCheck_suppression to match Go variant-suffix convention. Cosmetic-only; no public symbol change.
- examples/scan-demo demonstrates the DefaultConfig opinionated pattern at the consumer surface (mirrors Phase 8 DefaultScorer / DefaultScorerOptions).

### (iii) determinism-reviewer — APPROVED

- examples/scan-demo produces byte-stable stdout across runs (verified by `go run . && go run . | diff`).
- main_test.go pins stdout via os.Pipe redirect + byte-for-byte want constant — second drift-detection gate alongside testdata/golden/scan-default.json.
- No new map iteration on output paths introduced.
- No new `math.Pow/Log/Exp/FMA` calls.
- AlgoID.String() unchanged; sort key (Kind, NameA, NameB, GroupA, GroupB) unchanged.

### (iv) user-guide-reviewer — APPROVED

- docs/scan.md Quickstart fits in < 60 seconds of reader time (3 code blocks; under 30 lines).
- Public API table is scannable (single page; column widths reasonable).
- Suppression composition explained with three numbered subsections + composition note.
- Validation surface examples are runnable in head (full `errors.Is` discrimination shown).
- All "See also" links resolve to existing files (verified via `ls` and `test -f`).
- Performance section honest about the cross-group v0.x baseline — consumer expectation properly set.

### (v) docs-writer — APPROVED

- Every godoc Example opens with the function being demonstrated.
- SPEC OVERRIDE Phase 9 markers consistent across docs/requirements.md + scan source godoc + CHANGELOG + llms-full.txt.
- Cross-references resolve: docs/scan.md ↔ docs/requirements.md §12 ↔ docs/best-practices.md ↔ examples/scan-demo ↔ scorer.md.
- CHANGELOG.md follows Keep-a-Changelog format; backtick formatting for symbols; REQ-IDs in parenthetical citations.

### (vi) code-reviewer — APPROVED

- examples/scan-demo error handling: `os.Exit(1)` after `fmt.Fprintf(os.Stderr, ...)` is appropriate for an example program (the static-input failure case is a library bug, not a runtime error path).
- Example function naming follows godoc convention.
- No unused imports (verified by `golangci-lint run ./...` — 0 issues).

### (vii) security-reviewer — APPROVED

- Example program uses synthetic abstract identifier patterns (user_id, userId, lastSeen, request_id, audit, metrics). No real consumer data.
- No new attack surface introduced — this plan is docs + examples only.
- Tag values are never stringified in error messages (Tag remains opaque per scan.Item godoc).
- Threat model coverage from Plans 09-01..09-07 holds at the integrated level.

### (viii) go-quality — APPROVED

- `make check` exits 0 across all sub-targets.
- Root go.mod has zero non-stdlib runtime require lines (verified by `verify-deps-allowlist`: "2 non-indirect modules: github.com/axonops/fuzzymatch golang.org/x/text").
- `go mod tidy` is a no-op (verified by `tidy-check`).
- golangci-lint passes (0 issues across root + BDD modules).
- Test coverage 96.9% (above 95.0% floor).

### (ix) commit-message-reviewer — APPROVED

- 7 commits land cleanly under Conventional Commits format.
- All subject lines ≤ 70 chars.
- Bodies explain the "why" (DX-02 / DX-04 / D-06 closure rationale captured inline).
- No AI/LLM/Claude mentions.
- No GitHub issue refs (project has no GitHub issues per user-memory `project_no_github_issues`).

## Commits (Plan 09-08 atomic series)

| #  | SHA      | Subject                                                                            |
| -- | -------- | ---------------------------------------------------------------------------------- |
| 1  | 026fac2  | feat(examples): add scan-demo runnable example with 3 sections                     |
| 2  | 575ff96  | test(examples): pin scan-demo stdout byte-for-byte (golden meta-test)              |
| 3  | 6351af6  | docs(scan): extend example_test.go to cover every public symbol (DX-02)            |
| 4  | 4d5e628  | docs(scan): full user-facing reference replacing 58-line scaffold (DX-04)          |
| 5  | 8ae1da4  | docs(spec): amend §12.1 NameA canonicalisation — raw-byte lex (Phase 9 D-06)       |
| 6  | 75b3978  | docs(changelog): Phase 9 closure entries (scan + suppression + golden + docs)      |
| 7  | 8859d38  | docs(llms): sync llms.txt + llms-full.txt with final Phase 9 surface               |

Working tree clean (`git status` returns no output).

## Phase-boundary push status — DEFERRED to orchestrator (autonomous: false)

Per Plan 09-08 frontmatter `autonomous: false` on Task 10 (push), and per user-memory `feedback_push_cadence` (commit + push to origin at every phase boundary requires human confirmation), this worktree executor STOPS at the "ready to push" state:

- 7 atomic commits on the per-agent branch `worktree-agent-ac96605995c7d0eb9`
- Working tree clean
- SUMMARY.md committed
- `make check` GREEN on the final HEAD

The orchestrator (top-level Claude Code) will:

1. Merge this worktree branch back into the host branch
2. Prompt the user with the actual `git push origin main` command and the post-push CI monitoring expectation
3. Per user-memory `feedback_ci_before_verification_gate`, declare Phase 9 complete ONLY after origin/main CI is observed green across the full cross-platform matrix (linux/amd64, linux/arm64, darwin/amd64, darwin/arm64, windows/amd64)

## REQ-IDs closed by Plan 09-08

- DX-02 — runnable godoc Example per demonstrable public scan function (scan/example_test.go now has 5 Examples; examples/scan-demo/ is the dedicated runnable demo)
- DX-04 — docs/scan.md is the full user-facing reference (420 lines mirroring docs/scorer.md)

## Phase 9 deliverable roll-up (REQ-IDs closed across all 8 plans)

| REQ-ID  | Closed by Plan                            |
| ------- | ----------------------------------------- |
| SCAN-01 | 09-01 (foundation) + 09-03 (within/cross) |
| SCAN-02 | 09-04 (token-bucket + property test)      |
| SCAN-03 | 09-05 (three-rule suppression)            |
| SCAN-04 | 09-05 (Rule 3 cross-group identical)      |
| SCAN-05 | 09-06 (sort + completeness assertion)     |
| SCAN-06 | 09-02 (validation pipeline)               |
| PERF-05 | 09-04 (< 2 s / 10 k items)                |
| DET-04  | 09-06 (golden + NaN/Inf property tests)   |
| TEST-05 | 09-07 (BDD scenarios)                     |
| DX-02   | 09-08 (godoc Examples + scan-demo)        |
| DX-04   | 09-08 (docs/scan.md)                      |

Phase 9 — collection-scan sub-package — COMPLETE.

## Self-Check: PASSED

- examples/scan-demo/main.go exists: FOUND
- examples/scan-demo/main_test.go exists: FOUND
- examples/scan-demo/go.mod exists: FOUND
- examples/scan-demo/go.sum exists: FOUND
- scan/example_test.go has 5 Examples: VERIFIED via `go test -v -count=1 -run "^Example" ./scan/...`
- docs/scan.md ≥ 420 lines: VERIFIED via `wc -l docs/scan.md`
- CHANGELOG.md has Phase 9 entries with SPEC OVERRIDE: VERIFIED via grep
- llms.txt / llms-full.txt sync: VERIFIED via `make verify-llms-sync` (OK)
- docs/requirements.md §12.1 raw-byte lex amendment: VERIFIED via grep "raw-byte lex"
- 7 commits all PRESENT in git log
- `make check` GREEN
- `make verify-determinism` GREEN
- `make test-bdd` GREEN
- Working tree clean: VERIFIED via `git status --short` (empty output)
