---
gsd_state_version: 1.0
milestone: v1.0.0
milestone_name: "**Goal**: Final phase — re-scope `axonops/audit`"
status: executing
stopped_at: Completed Plan 08.5-08 (Q8b — Tokenise ASCII fast path with zero-copy substrings, equivalence property test, dedicated benchmarks, and BDD scenario)
last_updated: "2026-05-17T14:36:57.576Z"
last_activity: 2026-05-17
progress:
  total_phases: 1
  completed_phases: 0
  total_plans: 0
  completed_plans: 0
  percent: 0
---

# Project State

## Project Reference

See: .planning/PROJECT.md (updated 2026-05-13)

**Core value:** A developer can compare two strings (or scan a collection) with a known-correct algorithm and trust the resulting similarity score is mathematically sound, deterministic across platforms, and stable across patch releases.
**Current focus:** Phase 08.5 — review-remediation-gate

## Current Position

Phase: 08.5 (review-remediation-gate) — EXECUTING
Plan: 12 of 20 (08.5-08 just completed; 10 plans done: 01, 02, 03, 04, 05, 06, 07, 08, 09, 10)
Status: Ready to execute
Last activity: 2026-05-17

Progress: [░░░░░░░░░░] 0%

## Performance Metrics

**Velocity:**

- Total plans completed: 39
- Average duration: —
- Total execution time: 0.0 hours

**By Phase:**

| Phase | Plans | Total | Avg/Plan |
|-------|-------|-------|----------|
| 01 | 8 | - | - |
| 02 | 7 | - | - |
| 3 | 3 | - | - |
| 04 | 5 | - | - |
| 5 | 5 | - | - |
| 06 | 6 | - | - |
| 07 | 5 | - | - |

**Recent Trend:**

- Last 5 plans: —
- Trend: — (no execution yet)

*Updated after each plan completion*
| Phase 08.5 P01 | 6 | 1 tasks | 9 files |
| Phase 08.5 P05 | 15min | 1 tasks | 1 files |
| Phase 08.5 P07 | 30min | 2 tasks | 12 files |
| Phase 08.5 P09 | 25min | 1 tasks | 11 files |
| Phase 08.5 P10 | 50min | - tasks | - files |
| Phase 08.5 P02 | 80min | 1 tasks | 20 files |
| Phase 08.5 P04 | 12min | 1 tasks | 4 files |
| Phase 08.5 P06 | 25min | - tasks | - files |
| Phase 08.5 P03 | 35min | 1 tasks | 14 files |
| Phase 08.5 P08 | 25min | 2 tasks | 6 files |

## Accumulated Context

### Decisions

Decisions are logged in PROJECT.md Key Decisions table.
Roadmap-shaping decisions recorded at roadmap creation:

- Phase 1 lands ALL foundation infrastructure (CI matrix, golden files, property-test harness, benchstat, release pipeline, AlgoID dispatch, Normalise with Unicode NFC/NFD, Tokenise, errors) BEFORE any algorithm — 9 of 20 inventoried pitfalls are infrastructure-gated
- Smith-Waterman-Gotoh is isolated into its own phase (Phase 3) due to the documented Gotoh 1982 erratum requiring EMBOSS/biopython cross-validation
- Phonetic algorithms get their own phase (Phase 7) due to unique licence-discipline and primary-source-sourcing characteristics (NYSIIS Taft 1970 hard to obtain; Double Metaphone's ~200-branch rule table)
- Scorer placement: AFTER all 23 algorithms (Phase 8) — spec's default choice; bringing it forward would churn default-Scorer composition with every algorithm addition
- AlgoID dispatch table established in Phase 1 (per FOUND-02) so Monge-Elkan (Phase 6) can take an inner AlgoID parameter without waiting for the full Scorer
- Extract API isolated into its own phase (Phase 10) after Scan, layering atop both single algorithms and the Scorer
- [Phase ?]: Phase 8.5 Plan 01: atomic rename ErrInvalidAlgorithm -> ErrInvalidAlgoID across 15 call sites; add ErrInvalidInnerAlgo (Q4 follow-up) + ErrInternalInvariantViolated (Gap 5 typed-panic); remove 3 unused sentinels; apply 4-section godoc template to every remaining sentinel.
- [Phase ?]: Phase 08.5 Plan 05: Q11c paper-anchored Philips 2000 worked-examples test (10 cases) passes green; Gap 6 resolved — Sais produces (SS, SS) confirming Q9 dupBranchBody removal at double_metaphone.go:744 is behaviour-preserving; Plan 11 unblocked.
- [Phase ?]: Phase 08.5 Plan 07: Q7a [dmMaxLen]byte refactor lands; alloc count unchanged at 3 (structural floor) but byte-count drops 33% and wall time drops 18% on Schmidt benchmark. Q7d 25% capacity hint on q-gram maps. Q11e DL-Full ZeroAllocs un-skipped under Q8a ≤ 1 alloc budget. Q7c/Q7b heap-fallback scope notes added across 5 algorithm files.
- [Phase ?]: Phase 08.5 Plan 09 (Q13): 9 missing TestGolden_*_Staging functions added for DoubleMetaphone/MongeElkan/MRA/NYSIIS/PartialRatio/Soundex/TokenJaccard/TokenSetRatio/TokenSortRatio. Catalogue now at 23/23 producing tests. 9 staging JSONs regenerated to alphabetical-by-Name; informational note fields dropped (rationale lives in test godocs). Phase 8.5 success criterion 9 SATISFIED.
- [Phase ?]: Plan 08.5-10 (Q10): 3 new cross-validation corpora added — character (jellyfish==1.2.1, 33 cases), q-gram (py_stringmatching==0.4.7, 32 cases at q=2+q=3), monge-elkan (py_stringmatching==0.4.7, 32 cases with JW-inner + Lev-inner, asym + sym surfaces). Documented 4 fuzzymatch-vs-reference divergences via variant_divergence pattern + q-gram-uniqueness constraint + relaxed JW-inner tolerance (1e-6 for fp32-JW).
- [Phase ?]: Plan 08.5-02 (Q3): atomic rename MongeElkanScore↔MongeElkanScoreSymmetric — the unsuffixed name is now the symmetric default; the v0.x directional surface is now MongeElkanScoreAsymmetric; the inert NormalisationOptions parameter is removed from both surfaces. Breaking pre-v1.0. 20 .go/.txt files modified in a single feat! commit (672 insertions, 634 deletions). Example Output lines bit-stable. Cross-validation corpus JSON unchanged. BDD Gherkin step language preserved (Go-side mappings only updated).
- [Phase ?]: Phase 08.5 Plan 04 (Q2): atomic single-commit landing of NaN/Inf/(α+β≤0) strict-parameter guards on WithThreshold + WithAlgorithm + WithTverskyAlgorithm; TverskyScore + TverskyScoreRunes direct-call panics upgraded to typed-error values wrapping ErrInvalidTverskyParam / ErrInvalidQGramSize. Data-vs-parameter framework (docs/requirements.md §6.A) is now uniformly enforced at every parameter entry point.
- [Phase ?]: Plan 08.5-06: Applied Q11b FMA-defeating double-cast at cosine.go:343 and scorer.go:380 per docs/requirements.md §14.4; regenerated testdata/golden/scorer-default.json (4 entries × 1 ULP); algorithms.json byte-identical
- [Phase ?]: Phase 08.5 Plan 03 (Q5): atomic single-commit removal of the PartialRatio rune-variant across 14 Go files (138 insertions / 656 deletions). Acceptance grep gate satisfied (PartialRatioScoreRunes Go-source references = 0). Third of four breaking pre-v1.0 surface changes landed (Q1 pending; Q3/Q5 complete; Q4 sentinel-removal pending). Doc-tier residuals (1 docs/requirements.md + 1 llms.txt + 4 llms-full.txt) deferred to Plan 15.
- [Phase ?]: Plan 08.5-08 (Q8b): Tokenise ASCII fast path with byte-level dispatch. Lowercase=false emits zero-copy substrings (158→6 allocs on Long benchmark, 96% reduction). Lowercase=true uses single scratch buffer + per-token string conversion (unsafe.String excluded per m11 LOCKED 2026-05-17). Non-ASCII falls back to preserved rune path. TestProp_Tokenise_ASCIIFastPathEquivalent (500 seeds × 8 option-bitfield) is the load-bearing T-08.5-17 mitigation gate. Memory-retention scope documented in public godoc (T-08.5-16 mitigation). Token-tier algorithms inherit alloc savings.

### Pending Todos

[From .planning/todos/pending/ — ideas captured during sessions]

None yet.

### Blockers/Concerns

[Issues that affect future work]

- Phase 1 readiness: confirm self-hosted bench runner availability for benchstat regression detection (falls back to `ubuntu-latest` informationally if unavailable)
- Phase 1 readiness: confirm CLA-vs-DCO choice (recommendation: CLA Assistant, mirroring `axonops/mask`)
- Phase 7 readiness: NYSIIS primary source (Taft 1970, NY State Special Report No. 1) is hard to obtain — may need to cite Knuth or a secondary review article during planning

## Deferred Items

Items acknowledged and carried forward from previous milestone close:

| Category | Item | Status | Deferred At |
|----------|------|--------|-------------|
| *(none)* | | | |

## Session Continuity

Last session: 2026-05-17T14:36:03.797Z
Stopped at: Completed Plan 08.5-08 (Q8b — Tokenise ASCII fast path with zero-copy substrings, equivalence property test, dedicated benchmarks, and BDD scenario)
Resume file: 
None
