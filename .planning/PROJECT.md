# fuzzymatch

## What This Is

A standalone Go library for fuzzy string matching — a pluggable catalogue of 23 string-similarity algorithms, a weighted composite `Scorer`, an optional collection-scan sub-package, and a one-to-many `Extract` search API. Stdlib-only with a single curated exception (`golang.org/x/text` for Unicode normalisation), no cgo, Apache 2.0. Built for Go developers who need a correctness-first, deterministic, production-grade fuzzy-matching toolkit they can drop into any project.

## Core Value

**A developer can compare two strings (or scan a collection) with a known-correct algorithm and trust the resulting similarity score is mathematically sound, deterministic across platforms, and stable across patch releases.**

If everything else fails, this must work: a single algorithm function call returns a correct, reproducible score in `[0.0, 1.0]`.

## Requirements

### Validated

<!-- Shipped and confirmed valuable. -->

- [x] Unicode normalisation in `Normalise` — NFC/NFD + diacritic stripping via `golang.org/x/text/unicode/norm` (validated in Phase 1)
- [x] Apache-2.0 release plumbing — goreleaser v2, cosign keyless, Syft SBOM, OIDC attestation, CLA Assistant, commitlint, Dependabot, CodeQL, govulncheck, gosec (validated in Phase 1)
- [x] OSS-first developer experience from v0.1.0 — README, CONTRIBUTING, CHANGELOG, SECURITY, CoC, llms.txt + llms-full.txt, docs/*.md scaffolds, issue + PR templates (validated in Phase 1)
- [x] Cross-platform determinism plumbing — 5-platform CI matrix, golden-file canonical byte form, `make verify-determinism`, locked normalisation.json golden fixture (validated in Phase 1; cross-platform diff is a deferred UAT)
- [x] Algorithm correctness discipline scaffolding — patent-screen gate, fresh-implementation rule, AlgoID dispatch primitive, sentinel-errors vocabulary, the 9 standards skills + 17 agent gates ready for Phase 2 onwards (validated in Phase 1)
- [x] Phonetic algorithm coverage — Soundex (Knuth/Census), Double Metaphone (Philips 2000), NYSIIS (Taft 1970, 6-char truncation), MRA (NBS Tech Note 943) (validated in Phase 7)
- [x] Input-quality diagnostics — `fuzzymatch.Validate(a, b string) []Warning` + `Warning` struct + `WarnKind` enum (5 constants: `WarnEmptyInput`, `WarnUnequalLength`, `WarnNoTokensAfterNormalise`, `WarnAllNonASCIIDropped`, `WarnPathologicallyLargeInput`) with per-algorithm rule table; documented across the 6 required surfaces (README Quick Start + Common Patterns, `docs/algorithms.md`, `docs/best-practices.md`, per-algorithm godoc cross-references, `llms.txt` + `llms-full.txt`, runnable `examples/validate-input-quality/` programme) (VALIDATE-01..06; validated in Phase 8.5)
- [x] Phase 8 review-finding remediation — 74 Critical + 195 Important findings from the 14-agent comprehensive review CLOSED across 19 implementation plans implementing 14 Q-decisions and 7 gap resolutions; pre-v1.0 breaking-change consolidation landed (MongeElkan symmetric-by-default + opts removal, `PartialRatioScoreRunes` removal, Hamming silent-max, `ErrInvalidAlgorithm` → `ErrInvalidAlgoID` rename, 3 unused sentinels removed); data-vs-parameter validation framework (NaN/Inf/(α+β≤0) guards) + Q11b FMA-defeating double-cast + Q7a/Q8b performance optimisations + Q12a/Q12b test-infrastructure tightening + Q13 devops cluster + Q10 cross-validation corpora (jellyfish + py_stringmatching pinned generators) + Q11c paper-anchored Philips 2000 worked-examples test (validated in Phase 8.5)

### Active

<!-- Current scope. All hypotheses until shipped. Detailed technical scope lives in docs/requirements.md (the authoritative spec). -->

- [ ] 23-algorithm catalogue, each implemented from primary academic source with literature-reference unit tests, mathematical-invariant property tests, and BDD scenarios (see `docs/requirements.md` §7)
- [ ] Weighted composite `Scorer` — immutable, concurrent-safe, configurable threshold and normalisation (see `docs/requirements.md` §8)
- [ ] Optional `scan` sub-package — turnkey collection-scan layer over the Scorer with suppression semantics (see `docs/requirements.md` §9)
- [ ] One-to-many `Extract` / `ExtractOne` search API — `process.extract`-equivalent (RapidFuzz-inspired) for "find best matches in a candidate list" workflows (added to v1.0 scope 2026-05-13)
- [ ] Per-algorithm performance discipline — allocation budgets, ASCII fast paths, two-row DP, benchstat regression detection (see `docs/requirements.md` §12). Scaffolded in Phase 1; exercised against algorithms in Phase 2+.

### Out of Scope

<!-- Explicit boundaries with reasoning. -->

- **Needleman-Wunsch** — superseded by Smith-Waterman-Gotoh for our use cases (see `docs/requirements.md` §4)
- **Soft-TFIDF** — requires a corpus model; out of scope for a pure-function library (see `docs/requirements.md` §4)
- **Metaphone 3** — U.S. Patent 7440941; AxonOps declines to ship patent-encumbered algorithms even where unenforced
- **cgo / native bindings** — zero-cgo is a hard constraint for portability
- **Embedding-based or learned similarity** — pure-function, stdlib-only library; ML lives in downstream consumers
- **Windows support quirks beyond determinism** — windows/amd64 must pass the determinism gate, but no Windows-specific tooling investment
- **Initial dogfooding for audit-event taxonomy** — that consumer lives downstream of v1.0.0; the library is built spec-first, then consumed

## Context

- **Trigger / primary downstream consumer:** an audit-event taxonomy project needs a "this field looks similar to that field — are you sure you want to add it?" warning system. fuzzymatch is the library that powers those checks. The taxonomy project is downstream of v1.0.0 — it shapes the API surface but does not distort early-phase prioritisation.
- **Authoritative technical spec:** `docs/requirements.md` (1812 lines, 23 algorithms, full per-algorithm spec, mathematical invariants, performance budgets, release phasing, acceptance criteria). PROJECT.md is the high-level lens; the spec is the deep technical contract.
- **Prior-art survey:** `docs/prior-art-research.md` — Go ecosystem audit and algorithm taxonomy. Source material for `gsd-project-researcher` agents.
- **Reference project — `mask`:** existing AxonOps Go library on GitHub. fuzzymatch mirrors its DX patterns (goreleaser, CLA, DCO, NOTICE, conventional-commit CI, llms.txt, godoc examples) deliberately. Divergences from mask need rationale.
- **Tooling already in place:** 17 Claude Code agents (`.claude/agents/`) and 11 reusable skills (`.claude/skills/`) implementing the review gates — algorithm-correctness, algorithm-licensing, algorithm-performance, determinism, api-ergonomics, code-reviewer, security-reviewer, test-writer, BDD-scenario-reviewer, go-quality, docs-writer, user-guide-reviewer, devops, issue-writer/closer, test-analyst, commit-message-reviewer. `BOOTSTRAP.md` documents the wiring sequence.
- **No production code exists yet** — repo is at bootstrap commit. Phase 1 will lay the foundation (module init, Apache-2.0 headers, CI scaffolding, `Normalise`/`Tokenise`/`AlgoID`).

## Constraints

- **Tech stack:** Go 1.26+, stdlib + **a single curated runtime dep: `golang.org/x/text`** (Unicode normalisation only). No other runtime deps. No cgo. Test-only dependencies (godog, goleak, testify) isolated in `tests/bdd/go.mod` — never in root `go.mod`. Root tests use stdlib `testing` only. The runtime-dep allowlist is enforced by `make verify-deps-allowlist` in CI; any PR proposing to extend the allowlist requires explicit user approval and `algorithm-licensing-reviewer` sign-off.
- **Licence:** Apache 2.0. No GPL/LGPL-derived code anywhere. No patent-encumbered algorithms.
- **Performance:** per-algorithm allocation budgets, ASCII fast paths where applicable, two-row DP for `O(mn)` algorithms, benchstat-tracked regression detection (see `docs/requirements.md` §12).
- **Determinism:** cross-platform byte-identical output verified by golden-file test in CI matrix (linux amd64+arm64, darwin arm64, windows amd64). No map iteration on output paths. NaN/Inf/-0 handled explicitly. (see `docs/requirements.md` §11)
- **Release discipline:** releases happen exclusively via CI on tag push. No local tagging, no local goreleaser invocations, no `--no-verify` shortcuts.
- **Coverage targets:** ≥ 95% overall, ≥ 90% per file, 100% on public API surface (see `.claude/skills/go-testing-standards/SKILL.md`).
- **Correctness discipline:** every algorithm is fresh-implemented from its primary academic source, with the citation inline at the top of the file, the formula in the file's godoc block, literature reference vectors in unit tests, and mathematical invariants in property tests.
- **API surface authority:** the `api-ergonomics-reviewer` and `user-guide-reviewer` agents have final say on function names, signatures, option shapes, and error names. Code blocks in `docs/requirements.md` are illustrative; agents have veto authority (see CLAUDE.md, Design Principle 13).

## Key Decisions

| Decision | Rationale | Outcome |
|----------|-----------|---------|
| `docs/requirements.md` is the authoritative spec; PROJECT.md is a thin synthesis pointing at it | The spec is comprehensive (1812 lines) and the result of substantial design work; duplicating it would create drift | — Pending |
| OSS-first from v0.1.0 (full DX polish ships with the first algorithm) | Mirroring `mask`; community-facing posture from day 1 lowers contribution friction | — Pending |
| Mirror `mask` DX patterns broadly — goreleaser, CLA, DCO, NOTICE, llms.txt, godoc examples — divergences need rationale | `mask` is a working reference; consistency across AxonOps Go libraries lowers cognitive load for contributors | — Pending |
| Spec-first phasing; do NOT distort v0.1.0 around the audit-event taxonomy use case | Taxonomy work is a downstream consumer, not a v0.1.0 driver; building spec-first keeps the library general | — Pending |
| §19 phasing is the default but the roadmapper may restructure if it sees a better shape | §19 reflects considered design but isn't load-bearing; user will review and approve the final shape | — Pending |
| Zero runtime dependencies, no cgo, no testify in root tests | Maximises portability and supply-chain safety; stricter than mask | ⚠️ Revised 2026-05-13 |
| Narrow runtime-dep allowlist: stdlib + `golang.org/x/text` only (added 2026-05-13) | Unicode NFC/NFD + diacritic stripping is table-stakes for the audit-event taxonomy consumer; inlining a maintained NFC implementation is too much surface area; `x/text` is Go-team curated and supply-chain auditable. All other non-stdlib runtime deps remain forbidden. | — Pending |
| `Extract` / `ExtractOne` one-to-many search API in v1.0 scope (added 2026-05-13) | Most-requested feature in comparable libraries (RapidFuzz `process.extract`); shipping it in v1.0 differentiates fuzzymatch and removes the most common reason consumers reach for other libraries. May add a phase. | — Pending |
| Releases happen via CI only; no local tagging | Reproducibility, supply-chain integrity, prevents accidental releases | — Pending |
| Patent screen before every algorithm implementation (algorithm-licensing-reviewer is a gating agent) | Apache-2.0 hygiene; AxonOps declines patent-encumbered algorithms even where unenforced | — Pending |

## Evolution

This document evolves at phase transitions and milestone boundaries.

**After each phase transition** (via `/gsd-transition`):
1. Requirements invalidated? → Move to Out of Scope with reason
2. Requirements validated? → Move to Validated with phase reference
3. New requirements emerged? → Add to Active
4. Decisions to log? → Add to Key Decisions
5. "What This Is" still accurate? → Update if drifted

**After each milestone** (via `/gsd-complete-milestone`):
1. Full review of all sections
2. Core Value check — still the right priority?
3. Audit Out of Scope — reasons still valid?
4. Update Context with current state

---
*Last updated: 2026-05-19 — Phase 8.5 (Review Remediation Gate) complete: 19 implementation plans + Plan 18 verification gate, 20/20 SUMMARYs landed. Closed all 74 Critical + 195 Important findings from the Phase 8 14-agent comprehensive review (465 findings total). Q1–Q14 decisions + Gap 1–7 resolutions implemented; new Validate public surface (VALIDATE-01..06) ships across 6 documentation surfaces; pre-v1.0 breaking changes consolidated (MongeElkan symmetric default + opts removal, PartialRatioScoreRunes removal, Hamming silent-max, ErrInvalidAlgorithm→ErrInvalidAlgoID rename, 3 unused sentinels removed); data-vs-parameter validation framework lands NaN/Inf/(α+β≤0) guards across WithThreshold/WithAlgorithm/WithTverskyAlgorithm + typed-panic discipline at direct-call sites; Q11b FMA-defeating double-cast at cosine.go:359 + scorer.go:386; Q7a DoubleMetaphone [dmMaxLen]byte (-33% bytes, -18% time) + Q8b Tokenise ASCII fast path (-96% allocs on Long); Q10 3 new cross-validation corpora (character via jellyfish==1.2.1, qgram + monge-elkan via py_stringmatching==0.4.7); Q11c paper-anchored Philips 2000 worked-examples test; Q12a/Q12b test-infrastructure tightening (AST-based Floor 3 helper, mixed-shape property generators); Q13 devops cluster (36 SHA-pinned actions, release.yml needs:[ci], nightly.yml, make verify-llms-sync, bench.txt regen, HashiCorp licence audit); docs/algorithms.md rewritten from 239-line TBD scaffold to 1587-line per-algorithm reference; docs/best-practices.md NEW (216 LOC consumer guide); 23 algorithm files carry [fuzzymatch.Validate] godoc cross-refs. Plan 18 14-agent panel returned Critical 0 / Important 0 / Improvement 2 (R1 Q7d q-gram perf regression scheduled v1.x; R2 scan/suppression BDD deferred to Phase 9 — both by design). Post-verification cross-platform CI matrix surfaced + closed 4 Important findings the panel missed (golangci-lint v2 module path, Windows golden-file CRLF, markdownlint scope, -race requires cgo) plus a pre-existing flaky MongeElkan property test (replaced with constructive strict-subset universal property) and Floor 1/2/3 coverage closure (overall 92.1% → 96.9%; DefaultScorer + NewSWGParams defence-in-depth panics refactored into testable unexported helpers). 33 STRIDE threats (T-08.5-01..T-08.5-33) verified SECURED by gsd-security-auditor: 26 mitigate / 7 accept / 0 transfer / 0 open. UAT 10/10 pass. CI green across all 5 platforms (linux/amd64+arm64, darwin/amd64+arm64, windows/amd64) plus markdownlint, bench-compare, License Headers, Security, CodeQL, Code Quality workflows. Lesson recorded as user-memory feedback_ci_before_verification_gate.md: future phase verifications must run on the full cross-platform CI matrix before final approval.*

*Last updated: 2026-05-15 — Phase 6 (Token-based Algorithms) complete: 6 plans / 6 SUMMARYs, TOKEN-01..TOKEN-05 delivered. Shared `token_indel.go` (lcsLen / indelRatio + rune variants per Wagner-Fischer 1974 LCS-subsequence DP, two-row optimisation, stack-allocated `[65]int` rows when `min(m,n) ≤ 64`) plus Token Sort Ratio (Indel over space-sorted tokens), Token Set Ratio (three-way Indel max — bug-for-bug RapidFuzz issue #110 empty-set behaviour), Partial Ratio (byte + rune surfaces, three-region iteration with `s1_char_set` early-skip, equal-length symmetric tie-break), Token Jaccard (set-Jaccard with KEYSTONE RV-TJ3 set-vs-multiset divergence vs Q-Gram Jaccard), and Monge-Elkan (asymmetric + symmetric, 14-entry `permittedMongeElkanInner` allow-list at package scope, exhaustive panic test for every non-permitted AlgoID). RapidFuzz 3.14.5 cross-validation infrastructure (pinned generator, 20-entry × 4-surface corpus, Go loader test) — all 80 sub-tests pass within ε=1e-9. `algorithms.json` extended to 19 algorithms × 144 entries; bench.txt full-replaced with 4 LOCKED pathological fixtures; 4 new cross-algorithm consistency tests; `examples/identifier-similarity` extended to 19 columns. `make check` / `make test-bdd` / `make verify-determinism` all green. 27/27 must-haves met. Code review: 0 critical, 0 blocker, 5 warnings, 6 info (all advisory; one flaky `TestProp_MongeElkanScore_AsymmetricWhenTokenCountAsymmetric` deferred to follow-up).*

*Last updated: 2026-05-16 — Phase 7 (Phonetic Algorithms) complete: 5 plans / 5 SUMMARYs, PHON-01..PHON-04 delivered. Shipped Soundex (Knuth/Census variant; `Tymczak → T522` gate, H/W-skip rule, `Ashcraft = Ashcroft → A261`), Double Metaphone (Philips 2000, ~430 LOC state machine across 5 mandatory language branches: Germanic `Schmidt → (XMT, SMT)`, Greek `Catherine = Katherine → (K0RN, KTRN)`, Romance `Pacheco` contains PXK, Slavic `Sczepanski`, Chinese-origin `Cheung`; algorithm-licensing-reviewer sign-off recorded; 4 MIT-Go-port negative-attribution lines), NYSIIS (Taft 1970 with mandatory 6-char truncation per CONTEXT §2 LOCKED; Knuth TAOCP §6.4 cited as canonical algorithm description; 40% `variant_divergence` corpus rate against jellyfish 1.2.1 extended variant), and MRA (NBS Tech Note 943 with three public surfaces — `MRACode` / `MRACompare` (the catalogue's only non-float64 public return) / `MRAScore` — and package-level `var mraThresholdTable` with explicit `sum > 12 → 2` clamp). `permittedMongeElkanInner` extended 14 → 18 entries in lockstep with `monge_elkan_test.go` rejected-slice 9 → 5 shrink. Dual-pin `jellyfish==1.2.1 + Metaphone==0.6` Python cross-validation generator + 104-entry `vectors.json` corpus + new `phonetic-codes.json` string-equality determinism golden file. `testdata/golden/algorithms.json` extended 19 → 23 algorithms × 184 entries; `examples/identifier-similarity` extended 19 → 23 columns; new `examples/phonetic-keys` educational program; `bench.txt` full-replace with 4 new phonetic benchmarks (no >10% regression on Phase 6 carry-forward); new `monge_elkan_phonetic_inner.feature` BDD covering ME-over-{Soundex, DM, NYSIIS, MRA}. Code review surfaced 3 critical findings: CR-01 (NYSIIS RD/ND suffix dropped trailing D) FIXED in 81632f0 with RV-N13/RV-N14 regression tests; CR-02 (DM SC+i+=3 stride) INVESTIGATED and REJECTED in bdcf662 — original behaviour matches the oubiwann metaphone==0.6 reference corpus that `Sczepanski → SKPN` depends on; CR-03 (MRA one-empty returned 1.0 for short inputs) FIXED in c404634 with explicit `lenA==0 || lenB==0` guard + `TestMRACompare_OneEmpty`. 4/4 must-haves verified. `go test ./...` + BDD all green.*

*Phase 5 (Q-gram Algorithms) complete: 5 plans / 5 SUMMARYs, QGRAM-01..QGRAM-05 delivered. Shared `q_gram.go` (extractQGrams + extractQGramsRunes per Ukkonen 1992) plus Q-Gram Jaccard (Jaccard 1912), Sørensen-Dice (Dice 1945 + Sørensen 1948), Cosine (Salton & McGill 1983 eq.4.4 — load-bearing cross-platform float-determinism gate with sorted-key iteration + explicit `(x*y)+z` parenthesisation + `math.Sqrt`-only), and Tversky (Tversky 1977 — three-layer asymmetry gate: unit RV-T1≠RV-T2 + property + BDD). Cross-algorithm consistency tests pin Tversky/Jaccard equivalence at α=β=1, Tversky/Dice equivalence at α=β=0.5, and the QGramJaccard ≤ SorensenDice algebraic hierarchy. `make check` / `make test-bdd` / `make verify-determinism` all green, coverage 97.1% (per-file ≥ 90%), 4/4 verification must-haves met. Code review: 0 critical, 5 warnings, 7 info (advisory; Tversky NaN α/β hardening flagged for v1.x).*

*Phase 4 (Remaining Character & Gestalt) complete: 5 plans / 5 SUMMARYs, CHAR-07 + CHAR-09 + GESTALT-01 delivered. Strcmp95 (Winkler 1994), LCSStr (Wagner-Fischer 1974), Ratcliff-Obershelp (Dr. Dobb's 1988 — difflib-equivalent, asymmetric per OQ-1) shipped with full unit + property + fuzz + bench + BDD + golden coverage. Cross-validation against Python `difflib.SequenceMatcher(autojunk=False)` corpus committed. `make check` / `make test-bdd` / `make verify-determinism` all green, coverage 97.3%, 24/24 verification must-haves met. Code review: 0 blockers, 4 warnings, 6 info (advisory).*

*Phase 1 (Foundation & Infrastructure) complete: 8 plans / 8 SUMMARYs, 38/38 requirements accounted for, coverage 96.7%, five deferred items in `01-HUMAN-UAT.md`.*
