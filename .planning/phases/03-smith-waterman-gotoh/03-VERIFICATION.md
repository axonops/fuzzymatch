---
phase: 03-smith-waterman-gotoh
verified: 2026-05-14T14:30:00Z
status: passed
score: 4/4 must-haves verified
overrides_applied: 0
---

# Phase 3: Smith-Waterman-Gotoh Verification Report

**Phase Goal:** Implement Smith-Waterman-Gotoh local alignment with configurable affine gap penalty, **isolated into its own phase** because the published Gotoh 1982 affine-gap recurrence contains a known erratum (initialisation step and an indexing flip — a biorxiv survey found 8 of 31 lecture-slide reproductions inherit the bug) and primary-source citation alone is insufficient. The implementation must cross-validate against EMBOSS or biopython reference vectors, the erratum must be called out explicitly in the file's block comment, and `algorithm-correctness-reviewer` review is gated on the cross-validation evidence appearing in the PR description.

**Verified:** 2026-05-14T14:30:00Z
**Status:** passed
**Re-verification:** No — initial verification

## Goal Achievement

### Observable Truths (ROADMAP Success Criteria)

| # | Truth | Status | Evidence |
|---|-------|--------|----------|
| 1 | `SmithWatermanGotoh` produces scores byte-identical to biopython on documented reference-vector set including long-gap canary that exercises Gotoh 1982 initialisation step | VERIFIED | `TestSWG_CrossValidation` runs 16 sub-tests against `testdata/cross-validation/swg/vectors.json` (biopython 1.87 corpus). All 16 entries pass with `\|got - biopython_normalised\| <= 1e-9`. The `one_long_gap_canary` entry (`abc________def` / `abcdef`, default params) records biopython_score=3.0, biopython_normalised=0.5; our implementation matches at delta=0 (exact float equality). Confirmed via live test run. |
| 2 | Implementation file's block comment cites Gotoh 1982 AND explicitly names corrected formulation source (Flouri et al. biorxiv survey), with erratum and correction documented inline | VERIFIED | `swg.go` lines 18-35 cite Smith-Waterman 1981, Gotoh 1982 (J. Mol. Biol. 162:705-708), and Flouri et al. 2015 (biorxiv 031500). Lines 28-35 explicitly state the Gotoh 1982 erratum and document the corrected formulation: "every border cell of M, Ix, Iy initialises to 0 for LOCAL alignment (NOT -Inf, NOT the global-alignment gap-open ladder)". Recurrence shown verbatim in godoc block lines 37-57. |
| 3 | Configurable affine gap penalty (gap-open + gap-extend) exposed via per-algorithm option; default values documented with rationale; property tests verify identity, range, and non-negativity invariants | VERIFIED | `SWGParams` struct (swg.go:93) exposes `Match`, `Mismatch`, `GapOpen`, `GapExtend` as exported float64 fields. `NewSWGParams()` (swg.go:123) returns documented defaults (1.0, -1.0, -1.5, -0.5). Six public functions accept params via `*WithParams` variants. Property tests in props_test.go: `TestProp_SmithWatermanGotohScore_RangeBounds` (line 739), `_Identity` (751), `_Symmetric` (765), `_NoNaN` (778), `_NoInf` (790), `_NoNegativeZero` (802), `_GapSplitInvariance` (822), `_RawNeverExceedsMatchTimesMinLen` (847), `_MonotonicWithMatchReward` (871), `_Symmetric` rune variant (956). All pass under testing/quick (default 100 iterations). |
| 4 | Allocation budget enforced via benchmark; two-row DP variant; cross-platform golden file entry added; BDD scenario covers canonical long-gap reference case | VERIFIED | Two-row three-matrix DP: swg.go:317 uses `var buf [(maxStackInputLen + 1) * 6]float64` (3120 bytes stack buffer); no `[m+1][n+1]float64` allocated. Six benchmarks in swg_bench_test.go (ASCII_Short, ASCII_Medium, ASCII_Long, Unicode_Short, WithParams_ASCII_Short, RawScore_ASCII_Short); four 0-alloc gates achieve 0 B/op, 0 allocs/op per bench.txt. Runtime alloc tests: `TestSmithWatermanGotohScore_ZeroAllocs_ASCII_Short/_Medium` (swg_test.go:366, 380) assert via `testing.AllocsPerRun`. Golden file entry `SmithWatermanGotoh_one_long_gap_canary` (a=`abc________def`, b=`abcdef`, expected_score=0.5) present in testdata/golden/algorithms.json. BDD `swg.feature` line 43-46 includes scenario "gap-split canary — symmetric long-gap pair scores equally". `make verify-determinism` exits 0 confirming cross-platform byte-identical merged algorithms.json. |

**Score:** 4/4 truths verified

### Required Artifacts

| Artifact | Expected | Status | Details |
|----------|----------|--------|---------|
| `swg.go` | Public surface + DP kernel; cites Gotoh 1982 + Flouri 2015; documents erratum | VERIFIED | 18712 bytes; all 3 primary references cited; erratum statement at lines 28-35; recurrence in godoc; 6 public functions + SWGParams + NewSWGParams; two-row DP with 3120-byte stack buffer; no init(); no transcendental math (grep confirmed) |
| `dispatch_swg.go` | Registers `SmithWatermanGotohScore` into `dispatch[AlgoSmithWatermanGotoh]` | VERIFIED | 1477 bytes; uses `var _ = func() bool { dispatch[AlgoSmithWatermanGotoh] = SmithWatermanGotohScore; return true }()` pattern at line 30-33; no init(). TestDispatch_SmithWatermanGotohRegistered (algoid_test.go:287) confirms slot 6 non-nil. |
| `swg_test.go` | Unit tests + cross-validation gate + alloc gates | VERIFIED | 20498 bytes; 16 test functions including `TestSmithWatermanGotoh_GapSplitCanary` (line 273), `TestSmithWatermanGotohScore_ZeroAllocs_ASCII_Short/_Medium` (366, 380), and `TestSWG_CrossValidation` (411) with 16 t.Run sub-tests. |
| `swg_bench_test.go` | Six benchmarks (ASCII_Short/Medium/Long, Unicode_Short, WithParams, RawScore) | VERIFIED | All 6 benchmarks declared at lines 55, 69, 85, 102, 118, 134; each uses `b.ReportAllocs()`. bench.txt records 60 rows (6 series × 10 samples). |
| `props_test.go` (extended) | 9 SWG property tests + rune symmetry | VERIFIED | 10 property tests appended for SWG covering RangeBounds, Identity, Symmetric byte+rune, NoNaN, NoInf, NoNegativeZero, GapSplitInvariance, RawNeverExceedsMatchTimesMinLen, MonotonicWithMatchReward. |
| `testdata/cross-validation/swg/vectors.json` | 16 entries with one_long_gap_canary, biopython 1.87 | VERIFIED | 4577 bytes; version=1; biopython_version=1.87; 16 entries (8 mandatory + 8 additional); one_long_gap_canary at biopython_score=3.0, biopython_normalised=0.5; canonical byte form (2-space indent, trailing LF). |
| `testdata/golden/algorithms.json` | Merged 7-algorithm golden file with SWG entries | VERIFIED | 6143 bytes; 38 entries across 7 algorithms (DamerauLevenshteinFull, DamerauLevenshteinOSA, Hamming, Jaro, JaroWinkler, Levenshtein, SmithWatermanGotoh); 6 SWG entries including `SmithWatermanGotoh_one_long_gap_canary` (expected_score=0.5); byte-stable on re-run; cross-platform-identical via `make verify-determinism`. |
| `testdata/golden/_staging/swg.json` | Per-algorithm staging file with 6 entries | VERIFIED | 1067 bytes; 6 entries sorted alphabetically by Name; byte-stable. |
| `tests/bdd/features/swg.feature` | Gherkin scenarios including gap-split canary | VERIFIED | 2310 bytes; 6 scenarios: canonical reference vectors, identity, both-empty, one-empty, symmetry, gap-split canary (long-gap reference case at line 43-46). BDD test run confirms feature loaded with all scenarios passing. |
| `docs/requirements.md` §7.1.8 | Lists 6 SWG functions + SWGParams + NewSWGParams; Flouri 2015 citation | VERIFIED | §7.1.8 at line 427 cites all 3 primary references (incl. Flouri 2015 erratum source); lists `SWGParams`, `NewSWGParams`, 3 normalised + 3 raw functions; documents Raw* surface rationale, default params, no-validation policy, complexity O(min(m,n)) space, cross-validation note. |
| `llms.txt` | SWG section with 8 exported symbols | VERIFIED | Line 85 `### Smith-Waterman-Gotoh local-alignment similarity`; lines 87-94 list type SWGParams + NewSWGParams + 6 functions. TestAIFriendly passes. |
| `scripts/gen-swg-cross-validation.py` | Developer-only biopython generator | VERIFIED | 8004 bytes; executable; uses `Bio.Align.PairwiseAligner` mode="local"; deterministic byte-stable output; 16 cases in CASES list; BSD-3-Clause attribution. |
| `Makefile` (regen target) | `regen-swg-cross-validation` target | VERIFIED | Target present; in .PHONY list; tolerant `command -v python3` gate. |

### Key Link Verification

| From | To | Via | Status | Details |
|------|-----|-----|--------|---------|
| `dispatch_swg.go` | `algoid.go` dispatch[6] | `var _ = func() bool { dispatch[AlgoSmithWatermanGotoh] = SmithWatermanGotohScore; return true }()` | WIRED | TestDispatch_SmithWatermanGotohRegistered passes; slot 6 non-nil confirmed. |
| `swg.go SmithWatermanGotohScore` | `SmithWatermanGotohScoreWithParams` | Delegation via `NewSWGParams()` | WIRED | swg.go:158-168 confirms delegation pattern (single source of truth). |
| `swg.go fast path` | `normalise.go isASCII` | Same-package call | WIRED | swg.go:313 `if lb <= maxStackInputLen && isASCII(a) && isASCII(b)` — uses inherited Phase 1/2 helper. |
| `swg.go fast path` | Stack buffer `[(maxStackInputLen+1)*6]float64` | Conditional gate on ASCII + length | WIRED | swg.go:317 declares stack buffer; bench results confirm 0 allocs/op on ASCII_Short. |
| `swg_test.go TestSWG_CrossValidation` | `testdata/cross-validation/swg/vectors.json` | `os.ReadFile` + `json.Unmarshal` + per-entry `t.Run` | WIRED | 16 sub-tests run with delta=0 assertion against biopython_normalised; passes. |
| `swg_test.go TestSWG_CrossValidation` | `swg.go SmithWatermanGotohScoreWithParams` | Constructs `fuzzymatch.SWGParams` from entry, calls ScoreWithParams | WIRED | All 16 sub-tests pass with exact float equality. |
| `Makefile regen-swg-cross-validation` | `scripts/gen-swg-cross-validation.py` | `python3 scripts/gen-swg-cross-validation.py` | WIRED | Target present with command-v gate. |
| `algorithms_golden_test.go TestGolden_Algorithms_Merge` | All 7 `_staging/*.json` incl. swg.json | stagingFiles slice + merge loop | WIRED | swg.json at line 170 of stagingFiles slice; merged algorithms.json has 38 entries from all 7 algorithms. |
| `cross_algorithm_consistency_test.go` | `swg.go` + `levenshtein.go` divergence | `TestCrossAlgorithm_SWG_Levenshtein_SubstringDivergence` | WIRED | Test at line 218 asserts SWG > Lev on `http_request`/`http_request_header_fields`; passes. |
| `examples/identifier-similarity/main.go` | `fuzzymatch.SmithWatermanGotohScore` | Algorithms slice entry `{"SWG", fuzzymatch.SmithWatermanGotohScore}` | WIRED | Line 83; example builds and runs; TestExample_Output passes. |
| `tests/bdd/features/swg.feature` | `tests/bdd/steps/algorithms_steps.go` | godog step regex matching `I compute the SmithWatermanGotoh score between` | WIRED | BDD suite runs all 6 SWG scenarios; all pass. |
| `llms.txt` SWG section | `swg.go` exported symbols | ai_friendly_test.go meta-test parses go/ast | WIRED | TestAIFriendly passes; all 8 SWG symbols listed in llms.txt. |

### Data-Flow Trace (Level 4)

Phase 3 produces a pure-function library — no dynamic data fetching; algorithm functions consume their inputs directly. Data flow trace not applicable beyond Level 3 wiring (already covered above). The cross-validation corpus flows as static file → JSON unmarshal → SWGParams → SmithWatermanGotohScoreWithParams → float64 comparison; this flow is fully exercised by the 16 TestSWG_CrossValidation sub-tests.

### Behavioral Spot-Checks

| Behavior | Command | Result | Status |
|----------|---------|--------|--------|
| Full quality gate | `make check` | exit 0; fmt, vet, lint (0 issues), license-headers (65 files), no-runtime-deps, tidy, govulncheck (no vulns), race tests (2.526s), coverage 97.1% ≥ 95% floor | PASS |
| Cross-validation against biopython | `go test -race -run TestSWG_CrossValidation -v ./...` | 16 sub-tests all PASS; one_long_gap_canary at delta=0 | PASS |
| Golden file merge byte-stability | `go test -race -run TestGolden_Algorithms_Merge ./...` | PASS — zero diff without `-update`; 38 entries across 7 algorithms | PASS |
| Cross-algorithm divergence pin | `go test -race -run TestCrossAlgorithm_SWG_Levenshtein_SubstringDivergence ./...` | PASS — SWG=1.0 > Lev≈0.46 on http_request pair | PASS |
| Gap-split canary (PITFALLS §3) | `go test -race -run TestSmithWatermanGotoh_GapSplitCanary ./...` | PASS | PASS |
| Allocation gates | `go test -race -run TestSmithWatermanGotohScore_ZeroAllocs ./...` | PASS — 0 allocations on ASCII Short/Medium under `testing.AllocsPerRun(100, ...)` | PASS |
| Cross-platform determinism | `make verify-determinism` | exit 0; merged algorithms.json byte-stable | PASS |
| BDD swg.feature scenarios | `cd tests/bdd && go test -race -v ./...` | All 6 SWG scenarios pass (canonical reference vectors x N, identity, both-empty, one-empty, symmetric, gap-split canary) | PASS |
| `corpus version + biopython version` | `python3 -c "import json; print(json.load(open('testdata/cross-validation/swg/vectors.json'))['biopython_version'])"` | 1.87 | PASS |

### Probe Execution

Not applicable — fuzzymatch is a pure-function library; no `scripts/*/tests/probe-*.sh` infrastructure declared by the plans.

### Requirements Coverage

| Requirement | Source Plan | Description | Status | Evidence |
|-------------|------------|-------------|--------|----------|
| CHAR-08 | 03-01, 03-02, 03-03 | Smith-Waterman-Gotoh with configurable affine gap penalty; MUST cross-validate against EMBOSS or biopython reference vectors due to documented Gotoh 1982 erratum | SATISFIED | (1) Implementation in swg.go cites Smith-Waterman 1981, Gotoh 1982, Flouri et al. 2015 (the corrected formulation source); erratum statement inline. (2) Cross-validation against biopython 1.87 via testdata/cross-validation/swg/vectors.json (16 entries, all delta=0). (3) Configurable affine gap via SWGParams + 3 *WithParams variants. (4) Property tests for identity/range/symmetry/non-negativity in props_test.go (10 properties). (5) Two-row DP with 3120-byte stack buffer; 0 allocs/op gates. (6) Cross-platform golden file via testdata/golden/algorithms.json. (7) BDD `swg.feature` gap-split canary scenario. (8) docs/requirements.md §7.1.8 lists all 6 functions + SWGParams + NewSWGParams + Flouri 2015 citation. |

No orphaned requirements — REQUIREMENTS.md line 212 maps CHAR-08 to Phase 3, and the plans claim it.

Note: REQUIREMENTS.md line 29 shows CHAR-08 as `[ ]` (unticked) and line 212 lists status "Pending". This is a status-tracking artifact that the orchestrator typically updates at phase commit time; the actual implementation work is complete and verified. Recording as informational; not a blocker.

### Anti-Patterns Found

| File | Line | Pattern | Severity | Impact |
|------|------|---------|----------|--------|
| (none) | — | — | — | No `TBD`, `FIXME`, or `XXX` markers found in any Phase 3 modified file (`grep -n -E "TBD\|FIXME\|XXX" swg.go dispatch_swg.go swg_test.go swg_bench_test.go swg_fuzz_test.go` returns no debt markers). |

Two `//nolint:gocyclo` directives are present on swgDPRaw and swgDPRawRunes (swg.go:352, 450) — these are inherited Phase 2 pattern (matching damerau_full.go and damerau_osa.go) for the inherently branchy three-matrix recurrence. Each carries an inline justification comment. golangci-lint reports 0 issues with these directives in place.

### Human Verification Required

None. All goal criteria are programmatically verifiable and have been verified:

- Cross-validation against biopython: automated via TestSWG_CrossValidation (16 sub-tests; all PASS).
- Erratum citation: verifiable via grep on swg.go (Smith-Waterman 1981, Gotoh 1982, Flouri 2015 all present, erratum statement inline).
- Configurable affine gap + property tests: verifiable via grep + test execution.
- Allocation budget + two-row DP + golden file + BDD: verifiable via bench.txt, code grep, `make verify-determinism`, and BDD run.

### Gaps Summary

**No gaps found.**

All four ROADMAP success criteria are demonstrably satisfied; all required artifacts exist with substantive content and proper wiring; key links are wired and tested; the full quality gate (`make check`) exits 0; the load-bearing biopython cross-validation gate is closed with all 16 corpus entries matching at delta=0 (exact float equality); the Gotoh-erratum gate is closed at four layers (file-level godoc, unit test, property test, BDD scenario, and now cross-validation against biopython).

Phase 3 is shippable.

---

_Verified: 2026-05-14T14:30:00Z_
_Verifier: Claude (gsd-verifier)_
