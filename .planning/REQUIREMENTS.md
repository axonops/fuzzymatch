# Requirements: fuzzymatch

**Defined:** 2026-05-13
**Core Value:** A developer can compare two strings (or scan a collection) with a known-correct algorithm and trust the resulting similarity score is mathematically sound, deterministic across platforms, and stable across patch releases.

**Scope note:** This document is **high-level scope**. The deep technical specification — every algorithm's formula, edge cases, mathematical invariants, performance budgets, normalisation rules, suppression semantics — lives in [`docs/requirements.md`](../docs/requirements.md) (1812 lines, authoritative). Each requirement below references the relevant section of the spec.

---

## v1 Requirements

### Foundation (FOUND)

- [ ] **FOUND-01**: Module bootstrap — `go.mod` (Go 1.26+, root module with `golang.org/x/text` as the sole runtime dep), Apache-2.0 LICENSE + NOTICE, `tests/bdd/go.mod` sub-module with `replace` directive (`docs/requirements.md` §1, §5)
- [ ] **FOUND-02**: `AlgoID` typed enum with `String()`, `AlgoIDs()`, and dispatch-table backing (no `Algorithm` interface — avoids hot-path boxing) (`docs/requirements.md` §6, `.claude/skills/go-coding-standards/SKILL.md`)
- [ ] **FOUND-03**: `Normalise` pipeline — case-fold, separator-strip, ASCII fast path, **Unicode NFC/NFD normalisation + diacritic stripping via `golang.org/x/text/unicode/norm`** (`docs/requirements.md` §9)
- [ ] **FOUND-04**: `Tokenise` — camelCase / snake_case / PascalCase / kebab-case splitting with stable ordering (`docs/requirements.md` §10)
- [x] **FOUND-05**: Sentinel error hierarchy composing well with `errors.Is` / `errors.As` (`docs/requirements.md` §6, §6.A). Canonical v1.0 sentinels: `ErrEmptyScorer`, `ErrInvalidWeight`, `ErrInvalidThreshold`, `ErrInvalidAlgoID` (renamed from `ErrInvalidAlgorithm` in Phase 8.5 Gap 4 resolution), `ErrInvalidQGramSize`, `ErrInvalidTverskyParam`, `ErrInvalidInnerAlgo` (added Phase 8.5 Q4 follow-up), `ErrInternalInvariantViolated` (added Phase 8.5 Gap 5 resolution — typed panic value for library-internal bugs). The pre-8.5 sentinels `ErrInvalidInput`, `ErrInvalidConfiguration`, `ErrEmptyInput` are removed (Phase 8.5 Q4 — none had call sites).

### Character-based algorithms (CHAR)

- [ ] **CHAR-01**: **Levenshtein** edit distance with byte + rune variants, two-row DP, ASCII fast path (`docs/requirements.md` §7.1)
- [ ] **CHAR-02**: **Damerau-Levenshtein OSA** (Optimal String Alignment / restricted transposition) — distinct AlgoID from Full (`docs/requirements.md` §7.2)
- [ ] **CHAR-03**: **Damerau-Levenshtein Full** (Lowrance-Wagner unrestricted transposition) — distinct AlgoID from OSA (`docs/requirements.md` §7.3)
- [ ] **CHAR-04**: **Hamming** distance (equal-length strings; defined behaviour for unequal-length) (`docs/requirements.md` §7.4)
- [ ] **CHAR-05**: **Jaro** similarity (`docs/requirements.md` §7.5)
- [x] **CHAR-06**: **Jaro-Winkler** with configurable prefix boost (`docs/requirements.md` §7.6)
- [ ] **CHAR-07**: **Strcmp95** (Winkler 1994) with similar-character table (`docs/requirements.md` §7.7)
- [ ] **CHAR-08**: **Smith-Waterman-Gotoh** with configurable affine gap penalty — implementation MUST cross-validate against EMBOSS or biopython reference vectors due to documented Gotoh 1982 erratum (`docs/requirements.md` §7.8, research/PITFALLS.md #3)
- [ ] **CHAR-09**: **LCSStr** (longest common substring) similarity (`docs/requirements.md` §7.9)

### Q-gram / N-gram algorithms (QGRAM)

- [x] **QGRAM-01**: Shared q-gram extraction infrastructure (`q_gram.go`) consumed by Jaccard / Sørensen-Dice / Cosine / Tversky (`docs/requirements.md` §7.10)
- [x] **QGRAM-02**: **Q-Gram Jaccard** similarity (`docs/requirements.md` §7.11)
- [x] **QGRAM-03**: **Sørensen-Dice** similarity (`docs/requirements.md` §7.12)
- [x] **QGRAM-04**: **Cosine** similarity — explicit `(x*y) + z` parenthesisation, `math.Sqrt` only (no `math.Pow`), to guarantee cross-platform float determinism (`docs/requirements.md` §7.13, research/PITFALLS.md #9)
- [x] **QGRAM-05**: **Tversky** asymmetric similarity with configurable alpha/beta (`docs/requirements.md` §7.14)

### Token-based algorithms (TOKEN)

- [ ] **TOKEN-01**: **Monge-Elkan** with pluggable inner AlgoID — `MongeElkanScore(a, b, inner AlgoID)` (`docs/requirements.md` §7.15)
- [ ] **TOKEN-02**: **Token Sort Ratio** — cross-validate against RapidFuzz (Indel formula), not fuzzywuzzy (`docs/requirements.md` §7.16, research/PITFALLS.md #6)
- [ ] **TOKEN-03**: **Token Set Ratio** (`docs/requirements.md` §7.17)
- [ ] **TOKEN-04**: **Partial Ratio** (`docs/requirements.md` §7.18)
- [ ] **TOKEN-05**: **Token Jaccard** similarity over tokenised sets (`docs/requirements.md` §7.19)

### Phonetic algorithms (PHON)

- [ ] **PHON-01**: **Soundex** — Knuth/Census variant per Knuth TAOCP Vol. 3 §6.4. Mandatory reference vectors include `"Tymczak" → "T522"` (the discriminating case) (`docs/requirements.md` §7.20, research/PITFALLS.md #4)
- [x] **PHON-02**: **Double Metaphone** (Philips 2000) returning primary + alternate codes. Reference vectors per language-origin branch (Germanic, Slavic, Romance, Greek) (`docs/requirements.md` §7.21)
- [ ] **PHON-03**: **NYSIIS** (Taft 1970) with 6-char truncation (`docs/requirements.md` §7.22)
- [ ] **PHON-04**: **MRA** (NBS Tech Note 943) Match Rating Approach (`docs/requirements.md` §7.23)

### Gestalt algorithms (GESTALT)

- [ ] **GESTALT-01**: **Ratcliff-Obershelp** similarity (Dr. Dobb's Journal 1988) — explicitly the difflib-equivalent for consumers who want that semantic (`docs/requirements.md` §7.24)

### Composite Scorer (SCORER)

- [x] **SCORER-01**: `NewScorer(opts ...ScorerOption)` functional-options constructor; immutable after construction; concurrent-safe (`docs/requirements.md` §8.1)
- [x] **SCORER-02**: `DefaultScorer()` opinionated default; `DefaultScorerOptions()` returning the underlying `[]ScorerOption` for "default minus algorithm X" customisation (`docs/requirements.md` §8.2)
- [x] **SCORER-03**: Auto-normalised weights (sum-to-1 invariant) via `WithAlgorithm(AlgoID, weight)` (`docs/requirements.md` §8.3)
- [x] **SCORER-04**: `Score(a, b) float64` — composite weighted score in `[0.0, 1.0]` (`docs/requirements.md` §8.4)
- [x] **SCORER-05**: `ScoreAll(a, b) map[AlgoID]float64` — per-algorithm breakdown for tuning (`docs/requirements.md` §8.5)
- [x] **SCORER-06**: `Match(a, b) bool` + `Threshold()` accessor — threshold helper (`docs/requirements.md` §8.6)
- [x] **SCORER-07**: `Algorithms()` accessor returning the configured AlgoID set (`docs/requirements.md` §8.7)
- [x] **SCORER-08**: Normalisation control via `WithoutNormalisation()` / `WithNormalisation(opts)` (`docs/requirements.md` §8.8)

### Input validation (VALIDATE) — NEW for v1.0 (Phase 8.5)

Added during Phase 8.5 doc-alignment (2026-05-17) as the consumer-facing companion to the comparison-data leniency contract (§6.A). `Validate` surfaces problematic-but-non-fatal input shapes as warnings so consumers can audit input quality before scoring.

- [ ] **VALIDATE-01**: Public `fuzzymatch.Validate(a, b string) []Warning` function — single-pass diagnostic returning nil if no warnings apply; safe for concurrent use; never panics; never returns an error (`docs/requirements.md` §11.5)
- [ ] **VALIDATE-02**: Public `Warning` struct with `Algorithm AlgoID`, `Kind WarnKind`, `Detail string` fields (`docs/requirements.md` §11.5)
- [ ] **VALIDATE-03**: Public `WarnKind` enum with CamelCase `String()` method matching the AlgoID.String naming convention (`docs/requirements.md` §11.5 and §6 Algorithm identifiers)
- [ ] **VALIDATE-04**: Five `WarnKind` constants: `WarnEmptyInput`, `WarnUnequalLength`, `WarnNoTokensAfterNormalise`, `WarnAllNonASCIIDropped`, `WarnPathologicallyLargeInput` (`docs/requirements.md` §11.5)
- [ ] **VALIDATE-05**: Per-algorithm validation rules — one rule set per algorithm specifying which `WarnKind` constants the algorithm's degraded input triggers (e.g. Hamming → `WarnUnequalLength`; ASCII-only phonetic → `WarnAllNonASCIIDropped`; token-tier → `WarnNoTokensAfterNormalise`) (`docs/requirements.md` §11.5 "Per-WarnKind semantics")
- [ ] **VALIDATE-06**: Documentation across the six required surfaces per `.claude/skills/documentation-standards/SKILL.md` § Consumer-facing validation and diagnostics features — README Quick Start / Common Patterns, `docs/algorithms.md` (or `docs/best-practices.md`), per-algorithm godoc cross-references, `llms.txt` + `llms-full.txt`, user-guide section, and at least one runnable `examples/` program

### Collection scan (SCAN)

- [ ] **SCAN-01**: `scan.Check(items, cfg) []Warning` — within-group + cross-group passes with separate thresholds (`docs/requirements.md` §12.1)
- [ ] **SCAN-02**: Token-bucket optimisation — property-test verified equivalent to naive O(N²) via `PropCheck_BucketEquivalentToNaive` (`docs/requirements.md` §12.2)
- [ ] **SCAN-03**: Per-item `SilenceLint` flag + global `SuppressedPairs` list composing additively (`docs/requirements.md` §12.3)
- [ ] **SCAN-04**: Cross-group identical-name suppression default (configurable) (`docs/requirements.md` §12.4)
- [ ] **SCAN-05**: Deterministic output sort by `(Kind, NameA, NameB, GroupA, GroupB)` with in-line completeness assertion (`docs/requirements.md` §12.5)
- [ ] **SCAN-06**: Sentinel error hierarchy for scan-specific failures (`docs/requirements.md` §12.7)

### One-to-many search (EXTRACT) — NEW for v1.0

- [ ] **EXTRACT-01**: `Extract(query, choices, opts) []Match` — RapidFuzz `process.extract`-equivalent, returns top-N matches sorted by score (`docs/requirements.md` §4 — pulled into v1 scope 2026-05-13)
- [ ] **EXTRACT-02**: `ExtractOne(query, choices, opts) Match` — convenience wrapper returning best single match
- [ ] **EXTRACT-03**: Pluggable scorer (any `Scorer` or single `AlgoID`); cutoff threshold; result limit option
- [ ] **EXTRACT-04**: Deterministic tie-breaking (score-descending, then choice-index-ascending) — verified by property test
- [ ] **EXTRACT-05**: Documented worst-case complexity and warning about adversarial-input DoS

### Determinism (DET)

- [ ] **DET-01**: Cross-platform byte-identical output verified by golden-file test on CI matrix (linux amd64+arm64, darwin amd64+arm64, windows amd64) (`docs/requirements.md` §11.1)
- [ ] **DET-02**: Algorithm score stability across patch releases — score-changing edits require minor version bump (`docs/requirements.md` §11.2)
- [ ] **DET-03**: No map iteration on output paths (verified by property test + lint rule) (`docs/requirements.md` §11.3)
- [ ] **DET-04**: NaN, +Inf, -Inf, -0 explicit handling with `PropX_NoNaN/NoInf/NoNegativeZero` per algorithm and Scorer/scan (`docs/requirements.md` §11.4)
- [ ] **DET-05**: Golden files cover `algorithms.json`, `scorer-default.json`, `scan-default.json`, `extract-default.json`, `normalisation.json` (`docs/requirements.md` §11.5)
- [ ] **DET-06**: No transcendental float ops on output paths (`math.Sqrt`/`math.Abs`/`math.Min`/`math.Max` permitted; `Pow`/`Log`/`Exp`/`FMA` patterns forbidden) (`docs/requirements.md` §11.6, research/PITFALLS.md #9)

### Performance (PERF)

- [x] **PERF-01**: Per-algorithm allocation budgets enforced via benchmark assertions (`docs/requirements.md` §14)
- [x] **PERF-02**: ASCII fast paths for `Normalise`, Levenshtein, and other byte-level algorithms where applicable (`docs/requirements.md` §14.2)
- [x] **PERF-03**: Two-row DP (no full table) for all `O(mn)` algorithms (`docs/requirements.md` §14.3)
- [x] **PERF-04**: Benchstat regression detection in CI — `-count=10`, regression > 10% at `p < 0.05` fails (`docs/requirements.md` §14.4)
- [ ] **PERF-05**: `scan.Check` performance budget — < 2s for 10,000 items (`docs/requirements.md` §12.6)
- [ ] **PERF-06**: Stack-allocated buffers for ASCII fast paths (verified by escape-analysis check) (`docs/requirements.md` §14.5)

### Testing discipline (TEST)

- [ ] **TEST-01**: Literature reference vectors in unit tests for every algorithm, citing the source paper (`docs/requirements.md` §15.1, `.claude/skills/algorithm-correctness-standards/SKILL.md`)
- [ ] **TEST-02**: Property tests (`testing/quick`) for mathematical invariants per algorithm — identity, symmetry, range bounds, triangle inequality where applicable (`docs/requirements.md` §15.2)
- [ ] **TEST-03**: Fuzz tests with malformed UTF-8 corpus — panic-free guarantee (`docs/requirements.md` §15.3)
- [ ] **TEST-04**: Benchmark per algorithm with allocation assertions (`docs/requirements.md` §15.4)
- [ ] **TEST-05**: BDD scenarios (godog) per algorithm + Scorer composition pattern + scan suppression behaviour + Extract behaviour, isolated in `tests/bdd/` sub-module (`docs/requirements.md` §15.5)
- [ ] **TEST-06**: Meta-tests — README compiles, Makefile targets documented, llms.txt sync-check, godoc examples runnable (`docs/requirements.md` §15.6)
- [ ] **TEST-07**: Coverage ≥ 95% overall, ≥ 90% per file, 100% public API surface — enforced (not just reported) in CI (`docs/requirements.md` §15.7)
- [ ] **TEST-08**: Goleak in BDD module to catch goroutine leaks (`docs/requirements.md` §15.8)

### CI / quality gates (CI)

- [ ] **CI-01**: golangci-lint v2 with project `.golangci.yml` (`docs/requirements.md` §16.1)
- [ ] **CI-02**: `go vet` + `go test -race` (`docs/requirements.md` §16.2)
- [ ] **CI-03**: govulncheck on every PR (`docs/requirements.md` §16.3)
- [ ] **CI-04**: gosec + CodeQL security scans (`docs/requirements.md` §16.4)
- [ ] **CI-05**: `make verify-deps-allowlist` — root `go.mod` requires must be `golang.org/x/text` ONLY (`docs/requirements.md` §16.5, PROJECT.md Key Decisions)
- [ ] **CI-06**: `make verify-determinism` — golden-file diff on cross-platform matrix (`docs/requirements.md` §16.6)
- [ ] **CI-07**: Markdownlint (markdownlint-cli2) on all `*.md` (`docs/requirements.md` §16.7)
- [ ] **CI-08**: `go mod tidy` no-diff check in CI (`docs/requirements.md` §16.8)
- [ ] **CI-09**: Conventional-commit lint (`wagoid/commitlint-github-action`) on PRs (`docs/requirements.md` §17)
- [ ] **CI-10**: CLA assistant (`contributor-assistant/github-action`) on PRs (`docs/requirements.md` §17.1)
- [ ] **CI-11**: Dependabot daily for `gomod` (root + tests/bdd) and `github-actions` (`docs/requirements.md` §16.9)

### Release plumbing (REL)

- [ ] **REL-01**: GoReleaser v2 with `~> v2` version pin — generates `checksums.txt`, drives release pipeline (`docs/requirements.md` §17.1)
- [ ] **REL-02**: Cosign v3 keyless signing of `checksums.txt` (with `--bundle`) via GitHub OIDC (`docs/requirements.md` §17.2)
- [ ] **REL-03**: SBOM (Syft / SPDX) generation per release (`docs/requirements.md` §17.3)
- [ ] **REL-04**: GitHub Artifact Attestation (`actions/attest-build-provenance@v2`) (`docs/requirements.md` §17.4)
- [ ] **REL-05**: Releases happen exclusively via CI on tag push — no local `git tag`, no local `goreleaser release`, no `--no-verify` (`docs/requirements.md` §17.5, CLAUDE.md)
- [ ] **REL-06**: Apache-2.0 LICENSE, NOTICE, SECURITY.md, CODE_OF_CONDUCT.md, CONTRIBUTING.md, CHANGELOG.md (Keep-a-Changelog format) all present (`docs/requirements.md` §17.6)
- [ ] **REL-07**: Algorithm deprecation policy documented in CONTRIBUTING — within a major version, algorithms can be added but not removed; scoring changes require minor bump (research/FEATURES.md gap)

### Developer experience (DX)

- [ ] **DX-01**: README with quick-start (5-line example), feature table, algorithm catalogue summary, CI badges, mask-style polish (`docs/requirements.md` §18.1)
- [ ] **DX-02**: godoc on every public symbol with at least one `Example` per algorithm function + Scorer + scan + Extract (`docs/requirements.md` §18.2)
- [ ] **DX-03**: `llms.txt` + `llms-full.txt` AI-friendly documentation with sync-check meta-test (`docs/requirements.md` §18.3)
- [ ] **DX-04**: Algorithm-specific docs (`docs/algorithms.md`, `docs/scorer.md`, `docs/scan.md`, `docs/extending.md`, `docs/tuning.md`, `docs/performance.md`, `docs/faq.md`) (`docs/requirements.md` §18.4)
- [ ] **DX-05**: `examples/` directory with runnable mini-programs — audit-field-similarity, identifier-similarity, schema-dedup, extract demo (mirrors mask) (research/FEATURES.md gap)
- [ ] **DX-06**: FAQ entries for "Why no Needleman-Wunsch?", "Why no Metaphone 3?", "Why no embeddings?", "Why phonetic-as-binary in the Scorer?", "Why aren't algorithm functions generic?", "Why x/text but no other deps?" (research/FEATURES.md gap)
- [ ] **DX-07**: CODEOWNERS, issue templates (bug, feature, algorithm-proposal), PR template (`docs/requirements.md` §18.5)

---

## v2 Requirements

Deferred to post-v1.0. Tracked but not in current roadmap.

### API extensions (V2-API)

- **V2-API-01**: `Scorer.Explain(a, b) Explanation` — structured explanation of why a pair matched or didn't
- **V2-API-02**: `Scorer.Fingerprint() string` — stable hash of Scorer configuration for cache-key generation
- **V2-API-03**: `iter.Seq[Warning]` variant of `scan.Check` for streaming consumers
- **V2-API-04**: `ScanParallel` — opt-in goroutine-parallelised scan for very large inputs
- **V2-API-05**: `Scorer.ScoreNormalised(a, b) float64` — hot-path callers wanting to normalise-once-score-many

### Algorithms

- **V2-ALGO-01**: Q-Gram Jaccard with padding (Ukkonen) — distinct from current
- **V2-ALGO-02**: Affine variants of additional alignment algorithms beyond SWG, if demand surfaces

---

## Out of Scope

| Feature | Reason |
|---------|--------|
| **Needleman-Wunsch** | Superseded by Smith-Waterman-Gotoh for our use cases (`docs/requirements.md` §4) |
| **Soft-TFIDF** | Requires a corpus model; out of scope for a pure-function library (`docs/requirements.md` §4) |
| **Metaphone 3** | U.S. Patent 7440941; AxonOps declines patent-encumbered algorithms even where unenforced (`docs/requirements.md` §4) |
| **cgo / native bindings** | Zero-cgo is a hard portability constraint |
| **Embedding / semantic similarity** | Pure-function, stdlib-mostly library; ML lives in downstream consumers |
| **Persistent cache** | Stateless pure functions — caching is consumer's responsibility |
| **Goroutines / parallelism in algorithm or Scorer** | Determinism guarantee precludes non-deterministic reduction order |
| **I/O of any kind** | Pure functions only |
| **CLI tool** | Library only; CLI lives in downstream consumers |
| **Config files / YAML / TOML parsing** | Functional options only |
| **Web UI / HTTP handlers** | Out of scope |
| **Runtime dependencies beyond `golang.org/x/text`** | Narrow allowlist locked 2026-05-13; future additions require explicit user approval + algorithm-licensing-reviewer sign-off |
| **testify in root tests** | Stdlib `testing` only at root; testify confined to `tests/bdd/` (stricter than mask) |
| **Local releases / `git tag` from a workstation** | All releases via CI on tag push (`docs/requirements.md` §17.5) |
| **Windows-specific tooling beyond determinism gate** | windows/amd64 must pass the matrix, but no Windows-specific investment beyond that |
| **GPL/LGPL-derived code anywhere** | Apache-2.0 hygiene; primary-source fresh implementation enforced by algorithm-licensing-reviewer |

---

## Traceability

Populated by `gsd-roadmapper` on 2026-05-13. Each v1 requirement maps to exactly one phase. See `.planning/ROADMAP.md` for phase details and success criteria.

| Requirement | Phase | Status |
|-------------|-------|--------|
| FOUND-01 | Phase 1 | Pending |
| FOUND-02 | Phase 1 | Pending |
| FOUND-03 | Phase 1 | Pending |
| FOUND-04 | Phase 1 | Pending |
| FOUND-05 | Phase 1 | Complete |
| CHAR-01 | Phase 2 | Pending |
| CHAR-02 | Phase 2 | Pending |
| CHAR-03 | Phase 2 | Pending |
| CHAR-04 | Phase 2 | Pending |
| CHAR-05 | Phase 2 | Pending |
| CHAR-06 | Phase 2 | Complete |
| CHAR-07 | Phase 4 | Pending |
| CHAR-08 | Phase 3 | Pending |
| CHAR-09 | Phase 4 | Pending |
| QGRAM-01 | Phase 5 | Complete |
| QGRAM-02 | Phase 5 | Complete |
| QGRAM-03 | Phase 5 | Complete |
| QGRAM-04 | Phase 5 | Complete |
| QGRAM-05 | Phase 5 | Complete |
| TOKEN-01 | Phase 6 | Pending |
| TOKEN-02 | Phase 6 | Pending |
| TOKEN-03 | Phase 6 | Pending |
| TOKEN-04 | Phase 6 | Pending |
| TOKEN-05 | Phase 6 | Pending |
| PHON-01 | Phase 7 | Pending |
| PHON-02 | Phase 7 | Complete |
| PHON-03 | Phase 7 | Pending |
| PHON-04 | Phase 7 | Pending |
| GESTALT-01 | Phase 4 | Pending |
| SCORER-01 | Phase 8 | Complete |
| SCORER-02 | Phase 8 | Complete |
| SCORER-03 | Phase 8 | Complete |
| SCORER-04 | Phase 8 | Complete |
| SCORER-05 | Phase 8 | Complete |
| SCORER-06 | Phase 8 | Complete |
| SCORER-07 | Phase 8 | Complete |
| SCORER-08 | Phase 8 | Complete |
| VALIDATE-01 | Phase 8.5 | Pending |
| VALIDATE-02 | Phase 8.5 | Pending |
| VALIDATE-03 | Phase 8.5 | Pending |
| VALIDATE-04 | Phase 8.5 | Pending |
| VALIDATE-05 | Phase 8.5 | Pending |
| VALIDATE-06 | Phase 8.5 | Pending |
| SCAN-01 | Phase 9 | Pending |
| SCAN-02 | Phase 9 | Pending |
| SCAN-03 | Phase 9 | Pending |
| SCAN-04 | Phase 9 | Pending |
| SCAN-05 | Phase 9 | Pending |
| SCAN-06 | Phase 9 | Pending |
| EXTRACT-01 | Phase 10 | Pending |
| EXTRACT-02 | Phase 10 | Pending |
| EXTRACT-03 | Phase 10 | Pending |
| EXTRACT-04 | Phase 10 | Pending |
| EXTRACT-05 | Phase 10 | Pending |
| DET-01 | Phase 1 | Pending |
| DET-02 | Phase 2 | Pending |
| DET-03 | Phase 1 | Pending |
| DET-04 | Phase 2 | Pending |
| DET-05 | Phase 1 | Pending |
| DET-06 | Phase 1 | Pending |
| PERF-01 | Phase 2 | Complete |
| PERF-02 | Phase 2 | Complete |
| PERF-03 | Phase 2 | Complete |
| PERF-04 | Phase 1 | Complete |
| PERF-05 | Phase 9 | Pending |
| PERF-06 | Phase 1 | Pending |
| TEST-01 | Phase 2 | Pending |
| TEST-02 | Phase 2 | Pending |
| TEST-03 | Phase 1 | Pending |
| TEST-04 | Phase 2 | Pending |
| TEST-05 | Phase 2 | Pending |
| TEST-06 | Phase 1 | Pending |
| TEST-07 | Phase 1 | Pending |
| TEST-08 | Phase 1 | Pending |
| CI-01 | Phase 1 | Pending |
| CI-02 | Phase 1 | Pending |
| CI-03 | Phase 1 | Pending |
| CI-04 | Phase 1 | Pending |
| CI-05 | Phase 1 | Pending |
| CI-06 | Phase 1 | Pending |
| CI-07 | Phase 1 | Pending |
| CI-08 | Phase 1 | Pending |
| CI-09 | Phase 1 | Pending |
| CI-10 | Phase 1 | Pending |
| CI-11 | Phase 1 | Pending |
| REL-01 | Phase 1 | Pending |
| REL-02 | Phase 1 | Pending |
| REL-03 | Phase 1 | Pending |
| REL-04 | Phase 1 | Pending |
| REL-05 | Phase 1 | Pending |
| REL-06 | Phase 1 | Pending |
| REL-07 | Phase 1 | Pending |
| DX-01 | Phase 1 | Pending |
| DX-02 | Phase 2 | Pending |
| DX-03 | Phase 1 | Pending |
| DX-04 | Phase 1 | Pending |
| DX-05 | Phase 2 | Pending |
| DX-06 | Phase 1 | Pending |
| DX-07 | Phase 1 | Pending |

**Per-phase requirement counts:**

| Phase | Name | Requirements | Count |
|-------|------|--------------|-------|
| 1 | Foundation & Infrastructure | FOUND-01..05, DET-01, DET-03, DET-05, DET-06, PERF-04, PERF-06, TEST-03, TEST-06, TEST-07, TEST-08, CI-01..11, REL-01..07, DX-01, DX-03, DX-04, DX-06, DX-07 | 38 |
| 2 | Core Character Algorithms (six) | CHAR-01..06, DET-02, DET-04, PERF-01, PERF-02, PERF-03, TEST-01, TEST-02, TEST-04, TEST-05, DX-02, DX-05 | 17 |
| 3 | Smith-Waterman-Gotoh | CHAR-08 | 1 |
| 4 | Remaining Character & Gestalt | CHAR-07, CHAR-09, GESTALT-01 | 3 |
| 5 | Q-gram Algorithms | QGRAM-01..05 | 5 |
| 6 | Token-based Algorithms | TOKEN-01..05 | 5 |
| 7 | Phonetic Algorithms | PHON-01..04 | 4 |
| 8 | Composite Scorer | SCORER-01..08 | 8 |
| 8.5 | Review Remediation Gate | VALIDATE-01..06 (new); resolutions of existing requirements per the 9 scope clusters in ROADMAP | 6 (new) |
| 9 | Collection Scan Sub-package | SCAN-01..06, PERF-05 | 7 |
| 10 | Extract API | EXTRACT-01..05 | 5 |
| 11 | Integration Shakedown & v1.0.0 | (validation phase — no unique requirements; ships the work of Phases 1-10) | 0 |

**Coverage:**
- v1 requirements: 99 total (93 from the original 2026-05-13 inventory + 6 VALIDATE-* added 2026-05-17 during Phase 8.5 doc-alignment)
- Mapped to phases: 99 (100% coverage)
- Unmapped: 0
- Orphans: none
- Duplicates: none (each requirement maps to exactly one phase)

---

*Requirements defined: 2026-05-13*
*Last updated: 2026-05-17 during Phase 8.5 doc-alignment — 99 v1 requirements mapped across 12 phases (added Phase 8.5 + 6 VALIDATE-* requirements)*
