---
gsd_state_version: 1.0
milestone: v1.0.0
milestone_name: "**Goal**: Final phase — re-scope `axonops/audit`"
status: planning
stopped_at: Phase 9 context gathered
last_updated: "2026-05-19T17:52:04.599Z"
last_activity: 2026-05-19
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

Phase: 9
Plan: Not started
Status: Ready to plan
Last activity: 2026-05-19

Progress: [░░░░░░░░░░] 0%

## Performance Metrics

**Velocity:**

- Total plans completed: 60
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
| 08.5 | 21 | - | - |

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
| Phase 08.5 P11 | 40min | - tasks | - files |
| Phase 08.5 P12 | 35min | 4 tasks tasks | 20 files files |
| Phase 08.5 P13 | 40min | 4 tasks tasks | 10 files files |
| Phase 08.5 P14 | 10min | 1 tasks | 1 files |
| Phase 08.5 P15a | 9min | 2 tasks tasks | 28 files files |
| Phase 08.5 P15b | 22min | 3 tasks tasks | 43 files files |
| Phase 08.5 P16 | 55min | 3 tasks | 17 files |
| Phase 08.5 P17a | 30min | 3 tasks tasks | 25 files files |
| Phase 08.5 P17b | 50min | 2 tasks tasks | 22 files files |

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
- [Phase ?]: Phase 08.5 Plan 11 (Q11a/Q11d/Q12a/Q12b): Cluster 6 test-infrastructure bundle. Deleted stale bench.txt.new (Q11a — untracked, gitignored). Rewrote partial_ratio.go TODO(#TBD) as plain spec-deferred note (Q11d — no GH issues per memory). Shipped scripts/cmd/verify-exported-coverage/main.go AST helper enumerating exported symbols via go/parser.ParseDir; cross-references against 'go tool cover -func' for Floor 3a (>= 90% per func); AST identifier scan of *_test.go for Floor 3b. scripts/verify-coverage-floors.sh rewritten to invoke helper. mixedShapeStringGenerator() with 5 shapes per docs/requirements.md §15.3 lock. TestProp_Scorer_WeightSumOne uint16 overflow fixed via uint32(u)+1. Residue: 6 Floor-3 violations (5 funcs + 1 type) — input to Plan 13 improvement sweep.
- [Phase ?]: Phase 08.5 Plan 12 (Cluster 6 test-infrastructure bundle large-grain): 18 new test files. 3 meta-tests (internal_coverage_test build-tag gated, readme_shop_front, documentation_test 35 blocks/34 verified/1 skipped). 3 FuzzScorer_* harnesses. 9 rune-variant + 4 distance-variant + 1 phonetic-code fuzz harnesses. 3 Phase-7 phonetic convergence cross-algorithm tests (Smith/Smithe universal; Schmidt/Schmit + Knight/Night partial — empirically-derived pairs replacing plan-suggested NYSIIS-divergent pairs). 2 Scorer property tests. docs:skip-compile marker convention established.
- [Phase ?]: Plan 08.5-13 (Q4): Validate public surface lands. 10 new exported identifiers (Validate, Warning, WarnKind, WarnKinds, AlgoIDAny, 5 WarnKind constants). AlgoIDAny=-2 sentinel (-1 reserved). WarnKind iota+1 so zero is unset. Per-algorithm fanout for token-tier + ASCII-only Kinds. 64 KiB threshold. sort.SliceStable on (Algorithm, Kind) — T-08.5-26 mitigation. 13 unit tests, 4 benchmarks, FuzzValidate (14 seeds, 28k execs/0 crashes in 5s), 7 BDD scenarios. llms.txt indexed; remaining 6 doc surfaces deferred to Plan 17 as planned.
- [Phase ?]: Plan 08.5-14 (Q9, Gap 6): DoubleMetaphone dupBranchBody at the French SAIS-end rule (double_metaphone.go formerly lines 766-771) collapsed to a single dmAdd(&pBuf, &sBuf, &pLen, &sLen, "S", "") call with an 8-line inline comment citing the docs/requirements.md §7.21 spec lock and the TestDoubleMetaphone_PaperWorkedExamples Sais verification gate. Behaviour preservation verified by Plan 05's paper-anchored test (Gap 6 plan-DAG gate, all 10 cases green pre-edit). Single atomic refactor commit (9 insertions / 6 deletions). gocritic dupBranchBody Critical finding for double_metaphone.go closes. Benchmark allocs/bytes unchanged.
- [Phase ?]: Phase 08.5 Plan 15a (Q14b + Gap 1): mechanical refactors landed atomically — WriteGoldenFile unexport with test-only re-export via export_test.go (production helper named writeGoldenFile, wrapper-prefix string updated); 23 dispatch_*.go files refactored from var _ = func() bool {...}() to explicit func init() (Q14b option A, T-08.5-28 mitigated by unique-slot writes); scorer_options_test.go merged into scorer_options_internal_test.go (Gap 1 rename) via in-place content merge (collision Rule 3) — both probe helpers and Test* functions now live in one _internal_test.go. Plan 15b retains the non-mechanical residue (Gap 2 BDD, Gap 5 companion property test, Gap 7 outcomes, 30+ Code-fix lint sweep).
- [Phase ?]: Phase 08.5 Plan 15b (Q14b residue + Gap 2 + Gap 5 companion + Gap 7): 3 BDD scenarios for ErrInvalidAlgoID/ErrInvalidQGramSize/ErrInvalidTverskyParam landed in scorer.feature; TestDefaultScorer_NeverPanics_PropertyTest at quick.Check MaxCount=100 covers DefaultScorer + all 6 default-removal subsets + AlgoCosine silent-no-op. Gap 7 outcomes per CONTEXT.md defaults: MRACompare tuple documented (no code change); (SWGParams).Validate() added with ErrInvalidSWGParam sentinel + NewSWGParams self-test wrapping ErrInternalInvariantViolated; WithoutAlgorithm silent no-op godoc + BDD scenario; AlgoIDs() hot-path-caching godoc. Q14b lint sweep 63→0 issues across 10 linter classes (30 inline fixes + 28 documented //nolint annotations in golangci_residue.md). British English clean across .go files. runeAt → runeSizeAt rename across 6 phonetic-encoder call sites.
- [Phase ?]: Plan 08.5-16 (Q13): full devops cluster landed — 36 SHA-pinned action references across 8 workflow files; release.yml gated on a GitHub Checks API verify-ci-green pre-flight job; nightly.yml long-form fuzz + benchstat regression + auto-PR for fuzz corpus; make verify-llms-sync as a CI gate (strict on llms.txt, advisory on llms-full.txt until Plan 17); scripts/internal/astwalk shared AST helper; HashiCorp/MPL-2.0 audit PASS (distributed Apache-2.0 surface clean); docs/algorithms.md 9 H2 anchors standardised to hyphenated form; bench.txt regenerated (1296→1419 lines); Q7d 25% q-gram capacity hint surfaced as +15-19% time / +90-107% bytes regression on Medium/Long ASCII — flagged as out-of-scope follow-up.
- [Phase ?]: Plan 08.5-17a (Q12c, Q14a, Q6a, VALIDATE-06 surfaces 2+3): docs/algorithms.md rewritten from 239-line TBD scaffold to 1587-line per-algorithm reference (closes docs-writer Critical 100% TBD scaffold finding). 23 algorithm files carry [fuzzymatch.Validate] godoc cross-references with algorithm-scoped WarnKind callouts. Q12c panic-surface table enumerates 8 public functions + 4 typed-error counter-list. Q14a 5 MB/call ceiling documented with forward reference to GH issue #2. Q6a Ratcliff-Obershelp asymmetric-by-design exception added to .claude/skills/algorithm-correctness-standards/SKILL.md (local-only — .claude/ globally gitignored). markdownlint config extended with 4 additional disables tuned to doc structure.
- [Phase ?]: Phase 08.5 Plan 17b (VALIDATE-06 surfaces 1+4+5+6 + Gap 3 BDD + 30+ docs-writer Important findings): docs/best-practices.md NEW (216 LOC consumer-guide for Validate); examples/validate-input-quality/ NEW (90+122 LOC main + stdout-pinning meta-test); normalisation.feature + determinism.feature NEW (6 + 6 BDD scenarios — scan.feature + suppression.feature deferred to Phase 9 per CONTEXT.md Gap 3 default); README Quick Start refreshed to v1.0 with Common Patterns / Validate-then-Score sub-section; llms-full.txt gained Character-tier foundation section (closes Hamming gap) + Input validation section; Phase 8/9/10 stale framing scrubbed from README + llms.txt + scorer.md + tuning.md + scan.md + extending.md; CONTRIBUTING.md wrong §11.2 → §5 + §13.1 (REL-07 was fictitious); doc.go platform list darwin/amd64 added; faq.md Soft-TFIDF + Validate entries added; British English sweep — 1 hit fixed in prior-art-research.md. 284 BDD scenarios all green (was 272). Phase 8.5 success criterion 3 (6 Validate doc surfaces) FULLY SATISFIED.

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

Last session: 2026-05-19T17:52:04.589Z
Stopped at: Phase 9 context gathered
Resume file: 
.planning/phases/09-collection-scan-sub-package/09-CONTEXT.md
