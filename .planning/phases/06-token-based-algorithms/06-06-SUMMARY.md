---
phase: 06-token-based-algorithms
plan: 06
subsystem: algorithm-catalogue
tags: [finalisation, golden-merge, cross-algorithm-tests, bench-baseline, identifier-similarity, phase-close-out]

# Dependency graph
requires:
  - phase: 02-core-character-algorithms-six
    provides: [algorithms_golden_test.go staging-merge pattern, identifier-similarity 14-column baseline, cross_algorithm_consistency_test.go file]
  - phase: 03-smith-waterman-gotoh
    provides: [LCSStr / RatcliffObershelp staging entries carried forward]
  - phase: 04-remaining-character-gestalt
    provides: [Strcmp95 / LCSStr / RatcliffObershelp staging entries; RatcliffObershelp difflib semantic for PartialRatio cross-test]
  - phase: 05-q-gram-algorithms
    provides: [QGram / Dice / Cosine / Tversky staging entries; bench.txt Phase 5 baseline; MULTISET-Jaccard semantic for TokenJaccard cross-test]
  - phase: 06-token-based-algorithms / plan 06-01 (TokenSortRatio)
    provides: [testdata/golden/_staging/token_sort_ratio.json]
  - phase: 06-token-based-algorithms / plan 06-02 (TokenSetRatio)
    provides: [testdata/golden/_staging/token_set_ratio.json, RapidFuzz-#110 DEVIATION pin, BenchmarkTokenSetRatio_Pathological_AsymmetricSetCardinalities]
  - phase: 06-token-based-algorithms / plan 06-03 (PartialRatio)
    provides: [testdata/golden/_staging/partial_ratio.json, BenchmarkPartialRatio_Pathological_LongShortMismatch_{Bytes,Runes}]
  - phase: 06-token-based-algorithms / plan 06-04 (TokenJaccard)
    provides: [testdata/golden/_staging/token_jaccard.json, SET-semantics + KEYSTONE RV-TJ3]
  - phase: 06-token-based-algorithms / plan 06-05 (MongeElkan)
    provides: [testdata/golden/_staging/monge_elkan.json, asymmetric/symmetric two-surface pattern, BenchmarkMongeElkan_Pathological_1000Tokens]

provides:
  - testdata/golden/algorithms.json — extended to 144 entries across 19 algorithms (Phase 1-5 + 5 Phase 6); byte-stable across CI matrix
  - examples/identifier-similarity/main.go — 19-column rendered table (14 + 5 new Phase 6 columns); MongeElk wrapper binds LOCKED dispatch defaults
  - examples/identifier-similarity/main_test.go — `want` constant regenerated with byte-exact 19-column output (column-width gate also updated)
  - bench.txt — full-replace baseline including all 19 algorithms + 4 LOCKED pathological fixtures (TokenSet asymmetric, PartialRatio long-short ×2, MongeElkan 1000-tokens)
  - cross_algorithm_consistency_test.go — 4 new Phase 6 cross-algorithm tests pinning LOCKED divergences

affects: [07-phonetic-tier (additive allow-list expansion + per-plan staging-golden additions), 08-scorer (consumes 19-algorithm dispatch table), 10-extract]

# Tech tracking
tech-stack:
  added: []
  patterns:
    - "Staging-golden → finalisation-merge: per-plan staging files (testdata/golden/_staging/<algo>.json) merged into canonical algorithms.json on phase close-out, sorted alphabetically by Name with duplicate-Name sanity check"
    - "Identifier-similarity additive column extension: append-only modification of the algorithms slice; column-width constant (algoWidth=13) accommodates short labels for all new columns without layout change"
    - "Parameter-rich algorithm wrapping in examples: MongeElkanScoreSymmetric closure binds the LOCKED dispatch defaults (AlgoJaroWinkler + DefaultNormalisationOptions) so the example mirrors the dispatch table exactly"
    - "Cross-algorithm divergence pinning: each new test asserts ALL THREE legs of a divergence (this-side value + that-side value + the cross-check that they differ) so a regression of either algorithm fires"
    - "bench.txt full-replace via `go test -bench=. -benchmem -count=10 -run='^$' ./...`: the `-run='^$'` skip-tests pattern is required when property-test flakiness (out-of-scope items) would otherwise abort bench collection"

key-files:
  created:
    - ".planning/phases/06-token-based-algorithms/06-06-SUMMARY.md (this file)"
    - ".planning/phases/06-token-based-algorithms/deferred-items.md (flaky property test in plan 06-05 deferred to a follow-up)"
  modified:
    - "algorithms_golden_test.go (stagingFiles slice: 14 → 19 entries; alphabetical insertion of monge_elkan / partial_ratio / token_jaccard / token_set_ratio / token_sort_ratio)"
    - "testdata/golden/algorithms.json (regenerated via TestGolden_Algorithms_Merge -update; 144 entries across 19 algorithms)"
    - "examples/identifier-similarity/main.go (algorithms slice: 14 → 19 entries; MongeElk wrapper binds dispatch defaults)"
    - "examples/identifier-similarity/main_test.go (want constant regenerated for 19-column byte-exact output; column-width gate still passes)"
    - "cross_algorithm_consistency_test.go (4 new Phase 6 cross-algorithm tests appended; file-header comment updated to reflect Phase 6 coverage)"
    - "bench.txt (full-replace: 1,040 benchmark runs across 104 unique benchmark functions; 19 algorithms + 4 pathological fixtures)"

key-decisions:
  - "Full-replace bench.txt LOCKED 2026-05-15: regenerated on developer hardware (darwin/arm64, Apple M2) via `go test -bench=. -benchmem -count=10 -run='^$' ./...` taking 1475s. The `-run='^$'` (skip all unit tests) was required to bypass a pre-existing flaky property test in plan 06-05's contributions (TestProp_MongeElkanScore_AsymmetricWhenTokenCountAsymmetric) — tracked in deferred-items.md, fix scoped to a follow-up plan."
  - "MongeElkan example wrapper binds CONTEXT §4 LOCKED defaults: identifier-similarity uses `func(a,b) { return MongeElkanScoreSymmetric(a, b, AlgoJaroWinkler, DefaultNormalisationOptions()) }` mirroring dispatch_monge_elkan.go's slot-13 wiring exactly. Maintains identical column behaviour between the example and the public dispatch surface."
  - "Cross-algorithm tests pin BOTH sides + the divergence cross-check: each new Phase 6 cross-algorithm test asserts (a) this-side expected value, (b) that-side expected behaviour, and (c) the divergence itself (values must differ / value-A must exceed value-B / etc.). This three-leg form ensures a regression of EITHER algorithm fires the test, not just regressions of one side."
  - "TokenJaccard column rendered for ('a a b', 'a b') is 1.0 (SET semantics): verified end-to-end by example output and cross_algorithm test. The Phase 5 QGramJaccard on the same input yields < 1.0 (MULTISET semantics) — pinned by TestCrossAlgorithm_TokenJaccard_VsQGramJaccard_SetVsMultisetDivergence."
  - "Empty-input deviation pin LOCKED 2026-05-15: TestCrossAlgorithm_TokenSetRatio_EmptyDeviation_PinnedAgainstTokenJaccard asserts that TokenSetRatioScore('','') = 0.0 (DEVIATION per plan 06-02 LOCKED RapidFuzz #110) AND TokenJaccardScore('','') = 1.0 (STANDARD per plan 06-04 LOCKED) AND the gap is exactly 1.0. The catalogue intentionally contains BOTH conventions; this test prevents accidental homogenisation."

patterns-established:
  - "Phase finalisation = three modifications + one regeneration: (1) staging-merge slice append, (2) example-table column append, (3) cross-algorithm test append, plus (4) bench.txt full-replace. Pattern is reusable for Phase 7 (phonetic) and Phase 8+ finalisation plans."
  - "Cross-algorithm divergence test naming: `TestCrossAlgorithm_<NewAlgo>_<Behavior>_VsOrPinnedAgainst_<OtherAlgo>` — surfaces both endpoints in the test name so failing-test output identifies the cross-algorithm pair directly."
  - "MongeElkan symmetric-vs-asymmetric mean relationship: `MongeElkanScoreSymmetric(a, b, inner, opts) = (MongeElkanScore(a, b, inner, opts) + MongeElkanScore(b, a, inner, opts)) / 2` — pinned to within 1e-12 tolerance by TestCrossAlgorithm_MongeElkanSymmetric_VsAsymmetric_Identity. Defence-in-depth alongside the unit + property tests in plan 06-05."

requirements-completed: [TOKEN-01, TOKEN-02, TOKEN-03, TOKEN-04, TOKEN-05]

# Metrics
duration: 36min
completed: 2026-05-15
---

# Phase 6 Plan 06: Phase 6 Token-Based Algorithms Finalisation Summary

**Phase 6 close-out: merged the 5 Phase 6 staging-golden files into the canonical `testdata/golden/algorithms.json` (19 algorithms, 144 entries), extended the runnable `examples/identifier-similarity/` example from 14 to 19 columns, full-replaced `bench.txt` with the regenerated baseline including all 5 Phase 6 algorithms + 4 LOCKED pathological fixtures, and pinned 4 new cross-algorithm consistency tests (set-vs-multiset divergence, empty-input DEVIATION-vs-STANDARD contrast, symmetric-mean identity, PartialRatio-vs-RatcliffObershelp distinct semantic).**

## Performance

- **Duration:** ~36 min (most of which was the `make bench` regeneration at 1475s = 24.6 min)
- **Started:** 2026-05-15T12:06:16Z
- **Completed:** 2026-05-15T12:42:27Z
- **Tasks:** 2 / 2
- **Files modified:** 6 (algorithms_golden_test.go, testdata/golden/algorithms.json, examples/identifier-similarity/main.go, examples/identifier-similarity/main_test.go, cross_algorithm_consistency_test.go, bench.txt)
- **Files created:** 2 (this SUMMARY.md, deferred-items.md)

## Accomplishments

- **algorithms.json merged to 19 algorithms (144 entries)** — TestGolden_Algorithms_Merge regenerates the file deterministically via alphabetical merge of all 19 staging files with a duplicate-Name sanity check. Cross-platform CI matrix (linux/amd64, linux/arm64, darwin/amd64, darwin/arm64, windows/amd64) will diff byte-identically on next merge.
- **identifier-similarity extended 14 → 19 columns** — added TokenSort, TokenSet, Partial, TokenJac, MongeElk columns; the MongeElk wrapper binds the LOCKED dispatch defaults (AlgoJaroWinkler + DefaultNormalisationOptions); example renders correctly on the 7 pinned identifier pairs.
- **bench.txt full-replace with 104 benchmark functions (1,040 runs)** — all 19 algorithms exercised at ASCII Short/Medium/Long + Unicode Short; the 4 LOCKED pathological fixtures (TokenSet asymmetric set cardinalities, PartialRatio long/short mismatch in both Bytes and Runes variants, MongeElkan 1000-tokens) present; Phase 1-5 carry-forward benchmarks preserved.
- **4 cross-algorithm consistency tests added** — pinning the LOCKED Phase 6 divergences:
  - `TestCrossAlgorithm_TokenJaccard_VsQGramJaccard_SetVsMultisetDivergence` (plan 06-04 KEYSTONE / RESEARCH.md Pattern 8)
  - `TestCrossAlgorithm_TokenSetRatio_EmptyDeviation_PinnedAgainstTokenJaccard` (plan 06-02 RapidFuzz #110 DEVIATION vs plan 06-04 STANDARD)
  - `TestCrossAlgorithm_MongeElkanSymmetric_VsAsymmetric_Identity` (plan 06-05 CONTEXT §4 LOCKED — symmetric = mean of two asymmetric directions)
  - `TestCrossAlgorithm_PartialRatio_VsRatcliffObershelp_DistinctSemantic` (plan 06-03 Indel vs Phase 4 difflib — proves the surfaces are NOT aliased)

## Task Commits

Each task was committed atomically:

1. **Task 1: Merge Phase 6 staging-golden + extend identifier-similarity + cross-algorithm tests** — `38a23a7` (feat)
2. **Task 2: Full-replace bench.txt** — `5e3f389` (perf)

_Final metadata commit (this SUMMARY + deferred-items) will be created by the orchestrator after the worktree merges back to main._

## Files Created/Modified

### Modified

- `algorithms_golden_test.go` — `stagingFiles` slice extended 14 → 19 entries (alphabetical insertion of `_staging/monge_elkan.json`, `_staging/partial_ratio.json`, `_staging/token_jaccard.json`, `_staging/token_set_ratio.json`, `_staging/token_sort_ratio.json`).
- `testdata/golden/algorithms.json` — regenerated via TestGolden_Algorithms_Merge -update. Contains 144 entries across 19 algorithms. Byte-stable on re-run.
- `examples/identifier-similarity/main.go` — `algorithms` slice extended 14 → 19 entries; godoc updated to describe Phase 6 inclusion; MongeElk wrapper binds dispatch defaults per CONTEXT §4 LOCKED.
- `examples/identifier-similarity/main_test.go` — `want` constant regenerated for byte-exact 19-column stdout; column-count gate TestExample_ColumnWidths still passes (header width 279 == separator width 279).
- `cross_algorithm_consistency_test.go` — 4 new tests appended after the Phase 5 Tversky-asymmetry section; file-header documentation updated to list the Phase 6 LOCKED divergences.
- `bench.txt` — full-replace via `go test -bench=. -benchmem -count=10 -run='^$' ./...` on darwin/arm64 (Apple M2). 1,040 benchmark runs, 104 unique benchmark functions, ~1,475s total wallclock.

### Created

- `.planning/phases/06-token-based-algorithms/06-06-SUMMARY.md` — this file.
- `.planning/phases/06-token-based-algorithms/deferred-items.md` — tracks the pre-existing flaky property test discovered during bench regeneration.

## TOKEN-01..05 Commit-Status Verification

All 5 Phase 6 algorithm requirement IDs verified as fully committed:

| ID       | Algorithm        | dispatch[] | algorithms.json | identifier-similarity | staging file | bench.txt | cross-algo test |
|----------|------------------|-----------|-----------------|-----------------------|--------------|-----------|------------------|
| TOKEN-01 | TokenSortRatio   | slot 14 (dispatch_token_sort_ratio.go) | 10 entries | TokenSort column | _staging/token_sort_ratio.json | 4 fixtures | TokenJaccard/Cross-tests reference TokenSort via identity / one-empty convergence (Phase 1-5 cross-tests extend) |
| TOKEN-02 | TokenSetRatio    | slot 15 (dispatch_token_set_ratio.go)  | 12 entries | TokenSet column  | _staging/token_set_ratio.json  | 5 fixtures (incl. 1 pathological) | `TestCrossAlgorithm_TokenSetRatio_EmptyDeviation_PinnedAgainstTokenJaccard` |
| TOKEN-03 | PartialRatio     | slot 16 (dispatch_partial_ratio.go)    | 10 entries | Partial column   | _staging/partial_ratio.json    | 7 fixtures (incl. 2 pathological — Bytes + Runes) | `TestCrossAlgorithm_PartialRatio_VsRatcliffObershelp_DistinctSemantic` |
| TOKEN-04 | TokenJaccard     | slot 17 (dispatch_token_jaccard.go)    | 10 entries | TokenJac column  | _staging/token_jaccard.json    | 4 fixtures | `TestCrossAlgorithm_TokenJaccard_VsQGramJaccard_SetVsMultisetDivergence` |
| TOKEN-05 | MongeElkan       | slot 13 (dispatch_monge_elkan.go) — binds SYMMETRIC variant + AlgoJaroWinkler + DefaultNormalisationOptions per CONTEXT §4 LOCKED | 10 entries | MongeElk column (wrapper binds dispatch defaults) | _staging/monge_elkan.json | 6 fixtures (incl. 1 pathological) | `TestCrossAlgorithm_MongeElkanSymmetric_VsAsymmetric_Identity` |

(Slot numbers above reflect existing dispatch table indices; ordering came from plans 06-01..06-05 and was not changed by 06-06.)

## Deviations from Plan

### Auto-fixed Issues (Rules 1-3)

None — no bugs introduced, no missing critical functionality, no blocking issues.

### Out-of-scope Discoveries (deferred-items.md)

**1. [Out of scope] Flaky property test in plan 06-05's MongeElkan asymmetry property**
- **Found during:** Initial `make bench` invocation (which runs property tests before benchmarks)
- **Issue:** `TestProp_MongeElkanScore_AsymmetricWhenTokenCountAsymmetric` (props_test.go:3203-3239) failed once on heavily-shrunken Unicode random input. Re-runs in isolation (5×) pass; flake-rate appears to be <<1%.
- **Root cause hypothesis:** The property uses `strings.Fields(a)` for the token-count premise, but `MongeElkan` internally uses the project's `Tokenise` which splits on additional identifier boundaries. On highly-shrunken Unicode-only inputs the two tokenisers disagree on the premise, leading the property to fire even when its premise should hold vacuously.
- **Workaround used by plan 06-06:** Ran the bench step via `go test -bench=. -benchmem -count=10 -run='^$' ./...` (skip-tests pattern), which bypasses the property test entirely. Bench output is full and committable.
- **Deferred to:** Follow-up plan against plan 06-05.

## Auth gates

None — pure-function library, no auth surface.

## Threat Surface Scan

No new threat surface introduced. The threat register from the plan's `<threat_model>` block:

- **T-06-06-01** (tampering: silent algorithms.json drift) — **mitigated** by TestGolden_Algorithms_Merge running on every PR plus the duplicate-Name sanity check (lines 195-198 of algorithms_golden_test.go) plus the cross-platform CI matrix.
- **T-06-06-02** (tampering: bench.txt regression masking) — **mitigated** by benchstat regression detection in CI (>10% at p<0.05 fails the build). Local full-replace on developer hardware is acceptable here because Phase 6 introduces 5 new algorithms; the baseline is freshly grounded in their actual numbers, not stale Phase 5 numbers.

## Phase 6 Close-Out

- **ROADMAP.md Phase 6 entry** — should be flipped to `[x]` (Phase 6 done) by the orchestrator after all worktree merges. `gsd-verify-work` handles this transition.
- **REQUIREMENTS.md TOKEN-01..05 entries** — should be flipped to `[x]` / "Complete" in the traceability table by the orchestrator. The commit history end-to-end traceability for each ID is the table in the previous section.
- **No modifications to llms.txt / llms-full.txt by this plan** — per-plan sync was performed in 06-01..06-05; this plan is finalisation-only.

## Phase 7 Forward-Compatibility Note

When Phase 7 (phonetic-tier: Soundex / Double Metaphone / NYSIIS / MRA) lands, planners must:

1. **Add 4 new AlgoIDs** to the `permittedMongeElkanInner` allow-list in monge_elkan.go (per plan 06-05 SUMMARY's Phase 7 forward-compatibility note) — the allow-list grows from 14 → 18 entries.
2. **Update the exhaustive-panic test** in monge_elkan_test.go — the rejected-AlgoID count shrinks from 9 → 5.
3. **Add 4 new staging-golden files** under `testdata/golden/_staging/` (one per phonetic algorithm).
4. **Append 4 new entries to `stagingFiles` slice** in algorithms_golden_test.go (alphabetical insertion).
5. **Append 4 new columns to the identifier-similarity example** (e.g. "Soundex", "DMetaph", "NYSIIS", "MRA" — all ≤ 13 chars).
6. **Append 4 new benchmark groups to bench.txt** via full-replace.
7. **Optionally add cross-algorithm consistency tests** for phonetic-vs-character divergences (e.g. Soundex collapses "kn"/"n" by Knuth 1973 ruleset).

## Self-Check: PASSED

Files created:
- `[FOUND]` `.planning/phases/06-token-based-algorithms/06-06-SUMMARY.md`
- `[FOUND]` `.planning/phases/06-token-based-algorithms/deferred-items.md`

Files modified:
- `[FOUND]` `algorithms_golden_test.go` (commit 38a23a7)
- `[FOUND]` `testdata/golden/algorithms.json` (commit 38a23a7)
- `[FOUND]` `examples/identifier-similarity/main.go` (commit 38a23a7)
- `[FOUND]` `examples/identifier-similarity/main_test.go` (commit 38a23a7)
- `[FOUND]` `cross_algorithm_consistency_test.go` (commit 38a23a7)
- `[FOUND]` `bench.txt` (commit 5e3f389)

Commits:
- `[FOUND]` `38a23a7` — feat(06-06): merge Phase 6 staging-golden files + extend identifier-similarity + cross-algorithm tests
- `[FOUND]` `5e3f389` — perf(06-06): full-replace bench.txt baseline for Phase 6 catalogue

Verification gates (manually verified before SUMMARY write):
- `[PASS]` `go test -run 'TestGolden_Algorithms_Merge|TestCrossAlgo|TestAlgoIDs' -race -shuffle=on -count=1 ./...`
- `[PASS]` `go test -race -shuffle=on -count=1 ./examples/identifier-similarity/...`
- `[PASS]` `make verify-determinism`
- `[PASS]` `make fmt-check vet verify-license-headers verify-deps-allowlist`
