# fuzzymatch

## What This Is

A standalone Go library for fuzzy string matching — a pluggable catalogue of 23 string-similarity algorithms, a weighted composite `Scorer`, an optional collection-scan sub-package, and a one-to-many `Extract` search API. Stdlib-only with a single curated exception (`golang.org/x/text` for Unicode normalisation), no cgo, Apache 2.0. Built for Go developers who need a correctness-first, deterministic, production-grade fuzzy-matching toolkit they can drop into any project.

## Core Value

**A developer can compare two strings (or scan a collection) with a known-correct algorithm and trust the resulting similarity score is mathematically sound, deterministic across platforms, and stable across patch releases.**

If everything else fails, this must work: a single algorithm function call returns a correct, reproducible score in `[0.0, 1.0]`.

## Requirements

### Validated

<!-- Shipped and confirmed valuable. -->

(None yet — pre-release, nothing shipped)

### Active

<!-- Current scope. All hypotheses until shipped. Detailed technical scope lives in docs/requirements.md (the authoritative spec). -->

- [ ] 23-algorithm catalogue, each implemented from primary academic source with literature-reference unit tests, mathematical-invariant property tests, and BDD scenarios (see `docs/requirements.md` §7)
- [ ] Weighted composite `Scorer` — immutable, concurrent-safe, configurable threshold and normalisation (see `docs/requirements.md` §8)
- [ ] Optional `scan` sub-package — turnkey collection-scan layer over the Scorer with suppression semantics (see `docs/requirements.md` §9)
- [ ] One-to-many `Extract` / `ExtractOne` search API — `process.extract`-equivalent (RapidFuzz-inspired) for "find best matches in a candidate list" workflows (added to v1.0 scope 2026-05-13)
- [ ] Unicode normalisation in `Normalise` — NFC/NFD + diacritic stripping via `golang.org/x/text/unicode/norm` (the single permitted runtime dep)
- [ ] Cross-platform determinism — byte-identical output on linux/amd64, linux/arm64, darwin/arm64, windows/amd64 (see `docs/requirements.md` §11)
- [ ] OSS-first developer experience from v0.1.0 — README, CONTRIBUTING, CHANGELOG, godoc with examples, llms.txt, CI badges (mirrors mask)
- [ ] Apache-2.0 release plumbing — goreleaser, CLA, DCO, NOTICE, conventional-commit linting in CI (mirrors mask)
- [ ] Per-algorithm performance discipline — allocation budgets, ASCII fast paths, two-row DP, benchstat regression detection (see `docs/requirements.md` §12)
- [ ] Algorithm correctness discipline — every algorithm fresh-implemented from primary source with inline citation, patent-screened before implementation, no GPL/LGPL derivation

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
*Last updated: 2026-05-13 after initialization and post-research scope adjustments (Unicode normalisation in v1.0, runtime-dep allowlist of `golang.org/x/text`, `Extract` API in v1.0)*
