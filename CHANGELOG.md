# Changelog

All notable changes to `github.com/axonops/fuzzymatch` will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/), and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

### Added

- Project bootstrap: repository structure, Apache-2.0 licensing, CLAUDE.md, domain skills, review agents, comprehensive requirements specification.
- Spec-driven development via [GSD](https://github.com/gsd-build/get-shit-done).
- Twenty-three-algorithm catalogue specified in `docs/requirements.md` §7.
- Three-layer public API design (algorithm functions / `Scorer` / `scan` sub-package).
- Patent screen for Metaphone 3 documented as the "Metaphone 3 Precedent" — patent-encumbered algorithms are excluded.
- `fuzzymatch.Validate(a, b string) []Warning` — returns warnings for problematic-but-non-fatal input shapes (empty input, unequal length where the algorithm cares, no tokens after normalise, all non-ASCII dropped, pathologically large input). Companion to the lenient comparison-data contract: algorithms always produce a value; `Validate` reports whether the value is meaningful. Per Phase 8.5 Q4.
- `Warning` and `WarnKind` types accompanying `Validate`. `WarnKind` constants: `WarnEmptyInput`, `WarnUnequalLength`, `WarnNoTokensAfterNormalise`, `WarnAllNonASCIIDropped`, `WarnPathologicallyLargeInput`. `WarnKind.String()` returns CamelCase matching the constant suffix, per the AlgoID.String naming convention. Per Phase 8.5 Q4 + Q6b.
- `ErrInternalInvariantViolated` sentinel — typed panic value for library-internal invariants that should be impossible in correct usage (e.g. `DefaultScorer()` construction failure on a dispatch-table gap that survived option-time validation). Replaces the bare `panic("...this is a bug...")` at `scorer.go:586` with a structured panic carrying the wrapped internal cause. Consumers seeing this sentinel are observing a library bug, not a usage error, and should file an issue at <https://github.com/axonops/fuzzymatch/issues>. Discriminated via `errors.Is(panicValue.(error), ErrInternalInvariantViolated)`. The sentinel MUST NOT be used to wrap caller-supplied parameter errors. Per Phase 8.5 Gap 5 resolution.
- `docs/best-practices.md` — new consumer-facing document covering the Validate-then-Score idiom, Scorer-pinning patterns, concurrency discipline, error-handling conventions, and the three-layer architecture choice. Per Phase 8.5 Plan 17b (VALIDATE-06 surface 5).
- `examples/validate-input-quality/` — new runnable example demonstrating `fuzzymatch.Validate` against 2-3 input pairs with a meta-test pinning the stdout output byte-for-byte. Per Phase 8.5 Plan 17b (VALIDATE-06 surface 6).
- `tests/bdd/features/normalisation.feature` + `tests/bdd/features/determinism.feature` — new BDD feature files covering `Normalise` (NFC / NFD, diacritic stripping, idempotence, ASCII fast path) and the cross-call determinism contract. Per Phase 8.5 Plan 17b (Gap 3 default — scan + suppression features deferred to Phase 9).

### Changed

- `WithThreshold` now rejects NaN (returns `ErrInvalidThreshold`). `±Inf` was already rejected by the existing range check; this change closes the NaN escape. Guard form: `if math.IsNaN(t) || t < 0.0 || t > 1.0 { return ErrInvalidThreshold }`. Per Phase 8.5 Q2.
- `WithAlgorithm` now rejects weight values that are NaN, `±Inf`, or ≤ 0 (returns `ErrInvalidWeight`). Per Phase 8.5 Q2 extension covering all Scorer-construction parameter validation.
- `WithTverskyAlgorithm` now rejects `α + β ≤ 0` at construction time (returns `ErrInvalidTverskyParam`). Previously `WithTverskyAlgorithm(_, 0, 0, _)` constructed successfully then panicked at first `Score` call. The same guard applies to the direct `TverskyScore` call path, which panics with the same sentinel per the data-vs-parameter framework documented in §6.A. Per Phase 8.5 Q2.
- Every exported error sentinel now carries the four-section godoc block (What / Common causes / Resolution / Example) per `.claude/skills/documentation-standards/SKILL.md` § Error sentinel documentation. Per Phase 8.5 Q4.
- `AlgoID.String()` is documented as CamelCase matching the constant suffix (`"Levenshtein"`, `"JaroWinkler"`, `"NYSIIS"`, `"DamerauLevenshteinOSA"`, etc.) — locking the convention against earlier draft language that referenced snake_case forms. Per Phase 8.5 Q6b.
- Monge-Elkan direct calls (`MongeElkanScore`, `MongeElkanScoreAsymmetric`) now panic with `ErrInvalidInnerAlgo` when passed an invalid inner AlgoID (unknown AlgoID, `AlgoMongeElkan` self-reference, or a token-tier AlgoID). This aligns Monge-Elkan with the data-vs-parameter framework locked in Phase 8.5 Q2. Scorer-construction callers receive the same sentinel as a typed error. New sentinel `ErrInvalidInnerAlgo` added to the canonical `errors.go` set. Per Phase 8.5 Q4 follow-up.
- Renamed `ErrInvalidAlgorithm` → `ErrInvalidAlgoID` to match the canonical sentinel name in `docs/requirements.md §6` and resolve the dangling `monge_elkan.go` / `llms-full.txt` non-existent-sentinel Critical finding from the Phase 8 review. Code and spec are now aligned. Per Phase 8.5 Gap 4 resolution.

### Breaking (pre-v1.0)

- `MongeElkanScore` is now **symmetric** by default. The v0.x directional behaviour is available as the new `MongeElkanScoreAsymmetric`. The `NormalisationOptions` parameter has been removed from both functions — it had no effect inside Monge-Elkan. Per Phase 8.5 Q3.
- `PartialRatioScoreRunes` has been removed. Token-tier algorithms operate on the output of `Tokenise`, which is itself rune-aware; the byte-level Indel kernel produces correct results on Unicode post-`Tokenise`. Per Phase 8.5 Q5.
- Three unused error sentinels removed: `ErrInvalidConfiguration`, `ErrInvalidInput`, `ErrEmptyInput`. None had call sites in the library. Per Phase 8.5 Q4.
- `HammingDistance` and `HammingDistanceRunes` now return `int` (no error tuple). On unequal-length input the silent-max policy applies — they return `max(len(a), len(b))`. `HammingScore` continues to return `0.0` on unequal length. This matches existing code behaviour; the spec catch-up removed the placeholder `ErrUnequalLength` sentinel. Per Phase 8.5 Q1.

### Notes

This project is pre-release. The API is not stable until `v1.0.0`. Phase 8.5 (Review Remediation Gate) is the breaking-change consolidation phase ahead of v1.0; further breaking changes are expected to be minimal.

---

[Unreleased]: https://github.com/axonops/fuzzymatch/compare/main...HEAD
