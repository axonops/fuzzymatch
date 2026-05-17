---
status: issues_found
agent: test-writer
scope: test architecture gaps (phases 1-8)
reviewed: 2026-05-17T06:39:10Z
finding_counts:
  critical: 18
  important: 21
  improvement: 8
  total: 47
---

# Test Architecture Gaps — Phases 1–8

**Reviewed:** 2026-05-17
**Method:** static analysis of all `*_test.go` files, `testdata/`, `tests/bdd/features/`, `export_test.go`, and `example_test.go` against the requirements in `.claude/skills/go-testing-standards/SKILL.md` and `.claude/skills/algorithm-correctness-standards/SKILL.md`.
**Coordination note:** BDD-layer findings already captured in `bdd-scenario-FINDINGS.md` are not repeated here. This review focuses on unit, property, fuzz, benchmark, example, meta-test, and cross-validation gaps.

---

## Counts at a Glance

| Category | Public functions | Fuzz harnesses | On-disk corpus dirs | Example funcs |
|----------|-----------------|----------------|---------------------|---------------|
| Total public API | 79 | 31 (harnesses) | 20 (dirs) | 39 |

Fuzz coverage: 31 harnesses but only 20 on-disk `testdata/fuzz/` directories and numerous public functions with no fuzz exposure at all.

---

## Critical Findings

### [Critical] Fuzz harnesses for Distance and ScoreRunes variants of character-tier algorithms have NO fuzz coverage

- **File:** `levenshtein_fuzz_test.go`, `hamming_fuzz_test.go`, `jaro_fuzz_test.go`, `jarowinkler_fuzz_test.go`, `damerau_osa_fuzz_test.go`, `damerau_full_fuzz_test.go`
- **Phase introduced:** Phase 2
- **Issue:** Six character-tier fuzz harnesses each call only one public function — the byte-path Score variant. The following public functions receive zero fuzz coverage:
  - `LevenshteinDistanceRunes`, `LevenshteinScoreRunes`, `LevenshteinDistance` (fuzz only covers `LevenshteinScore`)
  - `HammingDistanceRunes`, `HammingScoreRunes`, `HammingDistance` (fuzz only covers `HammingScore`)
  - `JaroScoreRunes` (fuzz only covers `JaroScore`)
  - `JaroWinklerScoreRunes` (fuzz only covers `JaroWinklerScore`)
  - `DamerauLevenshteinOSAScoreRunes`, `DamerauLevenshteinOSADistance`, `DamerauLevenshteinOSADistanceRunes` (fuzz only covers `DamerauLevenshteinOSAScore`)
  - `DamerauLevenshteinFullScoreRunes`, `DamerauLevenshteinFullDistance`, `DamerauLevenshteinFullDistanceRunes` (fuzz only covers `DamerauLevenshteinFullScore`)
  Contrast with the correct multi-function pattern: `swg_fuzz_test.go` (FuzzSmithWatermanGotohScore) covers all six SWG public functions, and `lcsstr_fuzz_test.go` covers all four LCSStr public functions.
- **Standard:** `go-testing-standards/SKILL.md` §Fuzz Tests — "Every public function has a fuzz harness."
- **Action:** Code fix — expand each existing character-tier fuzz harness to call all public variants, following the SWG multi-function pattern.
- **Rationale:** A panic-inducing bug in a Rune or Distance variant will not be caught by the fuzzer. These paths handle malformed UTF-8 (which the go fuzzer routinely generates) and are the most crash-prone.

---

### [Critical] 11 fuzz harnesses have no on-disk seed corpus directory in `testdata/fuzz/`

- **File:** `testdata/fuzz/` (missing directories for listed harnesses)
- **Phase introduced:** Phases 6 and 7
- **Issue:** The following fuzz harnesses exist in source but have no corresponding `testdata/fuzz/<FuzzName>/` directory with a seed file:
  - `FuzzDoubleMetaphone`
  - `FuzzMongeElkanScore`
  - `FuzzMongeElkanScoreSymmetric`
  - `FuzzMRA`
  - `FuzzNYSIIS`
  - `FuzzPartialRatioScore`
  - `FuzzPartialRatioScoreRunes`
  - `FuzzSoundex`
  - `FuzzTokenJaccardScore`
  - `FuzzTokenSetRatioScore`
  - `FuzzTokenSortRatioScore`
  Without an on-disk seed directory, the fuzzer starts from an empty corpus and the nightly CI runs produce no reproducible crash evidence. The 20 existing Phase 2–5 harnesses each have a `seed-001` file.
- **Standard:** `go-testing-standards/SKILL.md` §Fuzz Tests — "Corpus checked into `testdata/fuzz/`. CI runs 60 seconds per fuzzer per build."
- **Action:** Code fix — create `testdata/fuzz/<FuzzName>/seed-001` files for each listed harness, using the programmatic seed pairs already present in the `f.Add(...)` calls within the harness body.
- **Rationale:** Without persisted seeds the fuzzer restarts cold on every CI run and cannot accumulate crash-inducing corpus entries between runs. This also means `go test -run=FuzzXxx ./...` (seed-only mode without `-fuzz` flag) exercises zero deterministic cases.

---

### [Critical] No Scorer-level fuzz harness exists

- **File:** missing: `scorer_fuzz_test.go`
- **Phase introduced:** Phase 8 (Scorer)
- **Issue:** `Scorer.Score`, `Scorer.Match`, and `Scorer.ScoreAll` are public methods on a public type. No fuzz harness exercises them. The `DefaultScorer()` and `NewScorer()` functions are also public. The testing standard requires one `Fuzz*` per public function. The Scorer is the highest-value composition surface and the one most likely to surface panics via unexpected normalisation interactions or floating-point edge cases on arbitrary input pairs.
- **Standard:** `go-testing-standards/SKILL.md` §Fuzz Tests — "Every public function has a fuzz harness."
- **Action:** Code fix — create `scorer_fuzz_test.go` with a `FuzzScorer_Score` harness that constructs `DefaultScorer()` once and fuzzes `Score(a, b)` asserting range bounds, no NaN, no Inf.
- **Rationale:** The Scorer is the integration point for all 23 algorithm dispatch paths. Arbitrary inputs reaching the Scorer exercise normalisation, all dispatched algorithms, and the weighted-sum reduction simultaneously.

---

### [Critical] Missing property tests for Rune-variant full invariant suite on character-tier algorithms

- **File:** `props_test.go`
- **Phase introduced:** Phase 2
- **Issue:** For the following algorithms, the `*Runes` score variant has ONLY a `Symmetric` property test. The full invariant set (RangeBounds, Identity, NoNaN, NoInf, NoNegativeZero) is absent:
  - `LevenshteinScoreRunes` — only `TestProp_LevenshteinScoreRunes_Symmetric` exists
  - `HammingScoreRunes` — only `TestProp_HammingScoreRunes_Symmetric`
  - `JaroScoreRunes` — only `TestProp_JaroScoreRunes_Symmetric`
  - `JaroWinklerScoreRunes` — only `TestProp_JaroWinklerScoreRunes_Symmetric`
  - `DamerauLevenshteinOSAScoreRunes` — only `TestProp_DamerauLevenshteinOSAScoreRunes_Symmetric`
  - `DamerauLevenshteinFullScoreRunes` — only `TestProp_DamerauLevenshteinFullScoreRunes_Symmetric`
  - `SmithWatermanGotohScoreRunes` — only `TestProp_SmithWatermanGotohScoreRunes_Symmetric`
  Compare: `LCSStrScoreRunes` has the full set (RangeBounds, Identity, Symmetric, NoNaN, NoInf, NoNegativeZero).
- **Standard:** `go-testing-standards/SKILL.md` §Property Tests — "Required per-algorithm properties: Identity, Range bounds, Symmetry, Never panics." `algorithm-correctness-standards/SKILL.md` §Mathematical Invariants — "All algorithms: Range bounds, Identity, Never panics."
- **Action:** Code fix — add the missing property tests for each listed Runes variant.
- **Rationale:** The rune path handles multi-byte UTF-8 code points and is a distinct code path from the byte path. A NaN or out-of-range return from the rune path would not be caught by the byte-path property tests.

---

### [Critical] Missing property tests for Distance variants of Levenshtein, Hamming, and Damerau algorithms (Runes triangle inequality)

- **File:** `props_test.go`
- **Phase introduced:** Phase 2
- **Issue:** The testing standard mandates a triangle-inequality property test for distance-based algorithms. While byte-path triangle inequality tests exist for `LevenshteinDistance`, `HammingDistance`, `DamerauLevenshteinOSADistance`, and `DamerauLevenshteinFullDistance`, none of the corresponding `*DistanceRunes` functions have triangle-inequality tests:
  - `TestProp_LevenshteinDistanceRunes_TriangleInequality` — missing
  - `TestProp_HammingDistanceRunes_TriangleInequality_EqualLength` — missing
  - `TestProp_DamerauLevenshteinOSADistanceRunes_TriangleInequality` — missing
  - `TestProp_DamerauLevenshteinFullDistanceRunes_TriangleInequality` — missing
  The rune-path implementations use independent code paths (rune conversion then the same DP, but byte boundary handling differs). A regression in rune conversion that inflates distance values would violate triangle inequality.
- **Standard:** `go-testing-standards/SKILL.md` §Property Tests — "Triangle inequality (for distance-based algorithms)." `algorithm-correctness-standards/SKILL.md` §Mathematical Invariants — "Distance-based algorithms: Triangle inequality."
- **Action:** Code fix — add four `TestProp_*DistanceRunes_TriangleInequality` property tests.
- **Rationale:** The rune-path distance functions are independent code paths exercised by a different test surface. A rune-conversion bug that changes distance outcomes would not be caught by the byte-path triangle inequality test.

---

### [Critical] No cross-validation corpus exists for character-tier algorithms (Levenshtein, Jaro, JaroWinkler, Hamming, Strcmp95, DL-OSA, DL-Full)

- **File:** `testdata/cross-validation/` (missing: `levenshtein/`, `jaro/`, `jarowinkler/`, `hamming/`, `strcmp95/`, `damerau-osa/`, `damerau-full/`)
- **Phase introduced:** Phases 2 and 3
- **Issue:** Only four cross-validation corpora exist: `swg/` (biopython), `ratcliff-obershelp/` (difflib), `token-ratios/` (RapidFuzz 3.14.5), `phonetic/` (jellyfish 1.2.1 + Metaphone 0.6). The seven character-tier algorithms implemented in Phases 2–3 have no external-reference corpus. The unit tests rely entirely on manually-transcribed literature reference vectors (which can contain transcription errors) rather than cross-validated outputs from independent implementations (jellyfish, python-Levenshtein, rapidfuzz, or NLTK). Per the algorithm-correctness-standards skill, reference vector cross-validation against existing MIT/BSD implementations is expected.
- **Standard:** `algorithm-correctness-standards/SKILL.md` §Fresh-Implementation Discipline — "The implementation may study existing MIT/BSD/Apache-licensed Go implementations for reference vector cross-validation only — to confirm that a canonical example produces the expected output." `go-testing-standards/SKILL.md` §BDD Tests cites SWG/EMBOSS and Ratcliff/difflib as the pattern to follow for all algorithms.
- **Action:** Code fix — create generator scripts and corpora for at minimum: `levenshtein/` (jellyfish), `jaro/` (jellyfish), `jarowinkler/` (jellyfish). Hamming and DL-OSA/Full are covered adequately by literature vectors; Strcmp95 and LCSStr have no widely-available Python reference and can defer.
- **Rationale:** The current reference-vector-only approach cannot detect silent transcription errors in the test data itself. A wrong reference vector masks a wrong implementation. The dual-pin pattern (literature vector + independent implementation) is the only way to detect both types of error.

---

### [Critical] No cross-validation corpus exists for q-gram tier algorithms (QGramJaccard, SorensenDice, Cosine, Tversky)

- **File:** `testdata/cross-validation/` (missing: `qgram/`)
- **Phase introduced:** Phase 5
- **Issue:** The four q-gram algorithms use literature reference vectors only (Ukkonen 1992, textbook bigrams, Salton & McGill 1983). No external cross-validation corpus exists against jellyfish, sklearn, or any other reference implementation. The cosine test has an unusual design decision: `cosine_test.go:25` explicitly notes "cross-validation density that would otherwise come from an external corpus" is handled inline — this is a deviation from the project pattern and creates a single point of failure (if the test author's calculation is wrong, both the implementation and the test are wrong simultaneously).
- **Standard:** `algorithm-correctness-standards/SKILL.md` §Fresh-Implementation Discipline — cross-validation against independent implementations expected.
- **Action:** Discuss-phase needed — determine which reference implementation (sklearn, scipy, jellyfish) to use for q-gram cross-validation; create corpus and test.
- **Rationale:** Tversky, Dice, and Cosine have subtle normalisation choices (multiset vs set semantics, |A|×|B| vs |A|+|B|) that differ between implementations. A cross-validation corpus would catch a wrong normalisation convention that still passes the literature vectors.

---

### [Critical] No cross-validation corpus for Monge-Elkan

- **File:** `testdata/cross-validation/` (missing: `monge-elkan/`)
- **Phase introduced:** Phase 6
- **Issue:** Monge-Elkan is a composite algorithm; its outputs depend on the inner metric and the asymmetric/symmetric averaging. No external reference corpus exists. The jellyfish library does not implement Monge-Elkan, but strsimpy and py_stringmatching do. The current tests rely entirely on hand-computed reference vectors in the test file, which share the same author as the implementation.
- **Standard:** `algorithm-correctness-standards/SKILL.md` §Fresh-Implementation Discipline.
- **Action:** Discuss-phase needed — determine appropriate reference implementation (strsimpy, py_stringmatching); create a small corpus covering at least JaroWinkler inner and Levenshtein inner.
- **Rationale:** The symmetric averaging of MongeElkan is unusual and could silently produce wrong values if the mean formula has a sign error. An independent reference corpus is the only defence-in-depth check.

---

### [Critical] 9 staging golden files exist but have no corresponding staging test in `algorithms_golden_test.go`

- **File:** `algorithms_golden_test.go` (missing: TestGolden_DoubleMetaphone_Staging, TestGolden_MongeElkan_Staging, TestGolden_MRA_Staging, TestGolden_NYSIIS_Staging, TestGolden_PartialRatio_Staging, TestGolden_Soundex_Staging, TestGolden_TokenJaccard_Staging, TestGolden_TokenSetRatio_Staging, TestGolden_TokenSortRatio_Staging)
- **Phase introduced:** Phases 6 and 7
- **Issue:** The `testdata/golden/_staging/` directory contains 23 JSON files (one per algorithm), but `algorithms_golden_test.go` implements staging tests for only 15. The nine Phase 6–7 algorithms (DoubleMetaphone, MongeElkan, MRA, NYSIIS, PartialRatio, Soundex, TokenJaccard, TokenSetRatio, TokenSortRatio) have JSON files in `_staging/` but no test function that reads and asserts those files. The `TestGolden_Algorithms_Merge` function assembles the promoted `algorithms.json` from staged files, but without staged tests the per-algorithm correctness of the JSON content is unverified before promotion. The golden files for Phase 7 phonetic algorithms are particularly risky: the phonetic codes golden test (`phonetic_codes_golden_test.go`) covers code strings but the staging golden JSON carries score vectors that need a separate test.
- **Standard:** `go-testing-standards/SKILL.md` §Meta-tests — golden file pattern; `determinism-standards/SKILL.md` — cross-platform stability gate.
- **Action:** Code fix — add nine `TestGolden_*_Staging` test functions in `algorithms_golden_test.go`, following the established pattern of existing staging tests.
- **Rationale:** Without staging tests, the JSON content in `_staging/` for Phase 6–7 algorithms is generated-and-forgotten. A wrong value in a staging file would be promoted into `algorithms.json` without any assertion failing.

---

### [Critical] `scorer_options_test.go` is in `package fuzzymatch` (internal) but named without `_internal_` convention

- **File:** `scorer_options_test.go`
- **Phase introduced:** Phase 8
- **Issue:** The testing standard says internal tests should follow the naming convention `levenshtein_internal_test.go` with `package fuzzymatch`. `scorer_options_test.go` is in `package fuzzymatch` (it uses the `applyOptionForProbe` internal helper) but is not named with `_internal_`. This causes confusion: a reader sees `scorer_options_test.go` and expects `package fuzzymatch_test` (external), but finds `package fuzzymatch`. `scorer_internal_test.go` follows the correct naming convention. The naming inconsistency creates a false read on the internal vs external boundary. For Phase 9, the scan package will need an analogous internal/external split; the inconsistent naming here creates a confusing template.
- **Standard:** `go-testing-standards/SKILL.md` §Unit Tests — "Internal tests where genuinely needed for unexported invariants: `levenshtein_internal_test.go` with `package fuzzymatch`."
- **Action:** Code fix — rename `scorer_options_test.go` → `scorer_options_internal_test.go`. This is a mechanical rename; the content is correct.
- **Rationale:** The naming convention signals test boundary to readers and to the codebase scan tools. An inconsistency here will propagate to future phases.

---

### [Critical] Missing `internal_coverage_test.go` meta-test for coverage floor enforcement

- **File:** missing: `internal_coverage_test.go`
- **Phase introduced:** Phase 1 (should have been created at bootstrap)
- **Issue:** The testing standard mandates an `internal_coverage_test.go` meta-test that enforces the coverage floor: ≥ 95% overall, ≥ 90% per file, 100% on public API surface. This file does not exist. Coverage targets are therefore tracked only informally via `make coverage`. A regression in coverage (e.g. a new branch added to a hot path without a corresponding test) will not cause `go test ./...` to fail.
- **Standard:** `go-testing-standards/SKILL.md` §Meta-tests — "`internal_coverage_test.go` — enforces the coverage floor (95% overall, 90% per file, 100% public API)."
- **Action:** Code fix — create `internal_coverage_test.go` using `testing.Coverage()` (Go 1.20+) or the `go test -cover` output parsing pattern to assert the floor at test time.
- **Rationale:** Without a programmatic floor, coverage drift is invisible until someone manually runs `make coverage` and reads the output. Automated enforcement is the only way to guarantee the 100%-public-API target doesn't silently erode.

---

### [Critical] Missing `readme_shop_front_test.go` meta-test

- **File:** missing: `readme_shop_front_test.go`
- **Phase introduced:** Phase 1 (should have been created at bootstrap)
- **Issue:** The testing standard mandates a `readme_shop_front_test.go` that compiles and runs the README's headline quick-start example, asserting the output matches the documented expected output byte-for-byte. `README.md` contains code examples. If the API surface changes (e.g. a function renamed by the api-ergonomics-reviewer) the README will drift silently. The existing `ai_friendly_test.go` checks that symbols appear in `llms.txt` but does not execute any README examples.
- **Standard:** `go-testing-standards/SKILL.md` §Meta-tests — "`readme_shop_front_test.go` — README's headline example compiles and runs, output exactly matches documented expected output."
- **Action:** Code fix — create `readme_shop_front_test.go` with a `TestREADME_QuickStartExample` that uses `go/types` or an exec approach to compile and run the quick-start snippet, asserting output.
- **Rationale:** The README is the primary consumer touch point. Silent divergence between README examples and real API behaviour erodes trust and makes the library harder to adopt. The test forces README synchronisation with every API change.

---

### [Critical] Missing `documentation_test.go` meta-test for `docs/*.md` code examples

- **File:** missing: `documentation_test.go`
- **Phase introduced:** Phase 1 (should have been created at bootstrap)
- **Issue:** The testing standard mandates a `documentation_test.go` that ensures README and `docs/*.md` code examples compile and produce documented output. The `docs/` directory contains `algorithms.md`, `scorer.md`, `scan.md`, `tuning.md`, `extending.md`, `performance.md`, `faq.md`. Code blocks in these files can silently diverge from the implementation. There is currently no automation checking documentation code examples.
- **Standard:** `go-testing-standards/SKILL.md` §Meta-tests — "`documentation_test.go` — README and `docs/*.md` code examples compile and produce documented output."
- **Action:** Code fix — create `documentation_test.go`. At minimum, it should parse `docs/*.md` for fenced Go code blocks and assert each block compiles with `go/types`. Execution-level assertions can be added incrementally.
- **Rationale:** Documentation drift is the most common consumer pain point for open-source libraries. A compilation-level gate is a low-friction, high-value safeguard.

---

### [Critical] No concurrent `NewScorer` construction test

- **File:** `scorer_test.go`
- **Phase introduced:** Phase 8
- **Issue:** `TestScorer_ConcurrentSafety` (line 903) correctly tests concurrent `Score`, `ScoreAll`, and `Match` calls on an already-constructed `*Scorer`. It does not test concurrent construction of multiple `*Scorer` instances via `NewScorer()`. If `NewScorer` mutates any package-level state (e.g. a singleton or a shared config map) during construction, concurrent calls could race. The `go test -race` flag would catch this, but there is no explicit test that runs `NewScorer` concurrently to validate its thread-safety contract is stated in its godoc.
- **Standard:** `go-testing-standards/SKILL.md` §Scorer-level properties — "Deterministic across runs: `Scorer.Score` is byte-identical for repeated calls"; concurrent safety is a documented guarantee.
- **Action:** Code fix — add `TestScorer_ConcurrentNewScorer` that spawns 50 goroutines each calling `NewScorer(DefaultScorerOptions()...)` simultaneously and asserts no panic and correct construction.
- **Rationale:** Concurrent `NewScorer` calls may happen in library consumers initialising multiple Scorer instances at startup (e.g. Cassandra node bringing up multiple SAI indexes simultaneously). The absence of a test leaves this guarantee unverified at the test layer.

---

### [Critical] Missing phonetic-algorithm cross-algorithm convergence tests in `cross_algorithm_consistency_test.go`

- **File:** `cross_algorithm_consistency_test.go`
- **Phase introduced:** Phase 7
- **Issue:** `cross_algorithm_consistency_test.go` covers identity-convergence, both-empty-convergence, and one-empty-convergence for Phase 2–3 character-tier algorithms only (Levenshtein, DL-OSA, DL-Full, Hamming, Jaro, JaroWinkler, SWG). It does not include the Phase 7 phonetic algorithms (Soundex, DoubleMetaphone, NYSIIS, MRA). The algorithm-correctness-standards skill mandates consistent both-empty → 1.0 and one-empty → 0.0 behaviour across the entire catalogue. Phonetic algorithms have a distinct empty-input rule: `SoundexScore("", "")` = 1.0 (two equal empty codes match), `SoundexScore("Robert", "")` = 0.0. There is no cross-algorithm test asserting this convention holds consistently across all four phonetic algorithms and comparing against the non-phonetic convention.
- **Standard:** `algorithm-correctness-standards/SKILL.md` §Edge cases — "Both inputs empty: return 1.0 by convention. One input empty: return 0.0 for distance-based and set-based algorithms. Phonetic algorithms return 1.0 IFF both encoded codes are empty."
- **Action:** Code fix — add a `TestCrossAlgorithm_PhoneticBothEmpty_Convergence` and `TestCrossAlgorithm_PhoneticOneEmpty_Scores0` to `cross_algorithm_consistency_test.go` covering all four phonetic score functions.
- **Rationale:** The phonetic both-empty convention is documented differently from character-tier conventions (MRACompare returns matched=true, sim=6 on both-empty). A single test that pins all four phonetic `Score` functions on the same input pair catches inconsistencies between implementations.

---

### [Critical] Missing cross-algorithm convergence for Phase 6 token-tier algorithms in identity/both-empty tests

- **File:** `cross_algorithm_consistency_test.go`
- **Phase introduced:** Phase 6
- **Issue:** `TestCrossAlgorithm_IdentityConvergence` (line 107) and `TestCrossAlgorithm_BothEmptyConvergence` (line 138) enumerate only the seven Phase 2–3 algorithms. They do not include the Phase 6 token-tier algorithms (TokenSortRatio, TokenJaccard, MongeElkanScoreSymmetric) or the Phase 5 q-gram tier (QGramJaccard, SorensenDice, Cosine). The LOCKED deviation for `TokenSetRatioScore` (returns 0.0 on both-empty) is covered by `TestCrossAlgorithm_TokenSetRatio_EmptyDeviation_PinnedAgainstTokenJaccard` but the identity and both-empty conventions for TokenSortRatio, TokenJaccard, and MongeElkan are not asserted in the cross-algorithm convergence table.
- **Standard:** `algorithm-correctness-standards/SKILL.md` §Edge cases.
- **Action:** Code fix — extend the `funcs` slices in `TestCrossAlgorithm_IdentityConvergence` and `TestCrossAlgorithm_BothEmptyConvergence` to include all catalogue score functions that follow the standard convention. Document the TokenSetRatioScore deviation as a comment exclusion.
- **Rationale:** A cross-algorithm convergence table that covers only 7 of 23 algorithms is incomplete as a regression guard. Adding Phase 5–6 algorithms to the table with no code change other than slice extension is low-cost, high-value coverage.

---

### [Critical] Missing Scorer-level "composite bounded by per-algorithm min/max" property test

- **File:** `scorer_test.go`
- **Phase introduced:** Phase 8
- **Issue:** The testing standard lists three Scorer-level property tests required: "Deterministic across runs", "Range bounds", and "Composite ≥ min, ≤ max: weighted composite is bounded by the per-algorithm min and max scores." The first two (`TestProp_Scorer_DeterministicAcrossRuns`, `TestProp_Scorer_ScoreInRange`) exist. The third — asserting that the composite weighted score is never below the minimum individual algorithm score and never above the maximum individual algorithm score when weights are normalised — is absent. This invariant catches implementation bugs where a weight > 1.0 amplifies a score beyond the per-algorithm maximum.
- **Standard:** `go-testing-standards/SKILL.md` §Property Tests — "Scorer-level properties: Composite ≥ min, ≤ max: weighted composite is bounded by the per-algorithm min and max scores."
- **Action:** Code fix — add `TestProp_Scorer_CompositeBoundedByPerAlgoMinMax` using `testing/quick`.
- **Rationale:** The weight normalisation guarantee (weights sum to 1.0) is tested, but the resulting property — that the composite is a convex combination of per-algorithm scores — is not. If normalisation had a rounding bug that let weights sum to 1.001, scores above 1.0 would appear.

---

### [Critical] Missing `ScoreAll` determinism property test at Scorer level

- **File:** `scorer_test.go`
- **Phase introduced:** Phase 8
- **Issue:** `TestProp_Scorer_DeterministicAcrossRuns` tests that `Scorer.Score` returns the same float64 on repeated calls. No equivalent property test exists for `Scorer.ScoreAll`. The `ScoreAll` method returns a `map[AlgoID]float64`; map allocation itself is deterministic (same keys each call) but the per-value floating-point computation must also be deterministic. This is particularly important because `ScoreAll` exposes the per-algorithm dispatch layer directly — any non-determinism in a dispatched function would appear in `ScoreAll` before it appears in `Score`.
- **Standard:** `go-testing-standards/SKILL.md` §Property Tests — "Deterministic across runs: `Scorer.Score` is byte-identical for repeated calls."
- **Action:** Code fix — add `TestProp_Scorer_ScoreAll_DeterministicAcrossRuns` that calls `s.ScoreAll(a, b)` twice and asserts each key's value is bit-identical across both calls.
- **Rationale:** `ScoreAll` is the per-algorithm diagnostic surface and the one most likely to expose non-determinism first, since it bypasses the weighted-sum reduction.

---

## Important Findings

### [Important] Missing example tests for 33 public functions

- **File:** `example_test.go`
- **Phase introduced:** Phases 2–8
- **Issue:** The standard mandates one Example per public function. `example_test.go` has 39 example functions covering the primary score surface of all 23 algorithms. The following public functions have no `Example*` function:
  - Distance functions: `LevenshteinDistance`, `LevenshteinDistanceRunes`, `HammingDistance`, `HammingDistanceRunes`, `DamerauLevenshteinOSADistance`, `DamerauLevenshteinOSADistanceRunes`, `DamerauLevenshteinFullDistance`, `DamerauLevenshteinFullDistanceRunes`
  - Rune-path Score functions: `LevenshteinScoreRunes`, `HammingScoreRunes`, `JaroScoreRunes`, `JaroWinklerScoreRunes`, `DamerauLevenshteinOSAScoreRunes`, `DamerauLevenshteinFullScoreRunes`, `SmithWatermanGotohScoreRunes`, `SmithWatermanGotohScoreWithParams`, `SmithWatermanGotohRawScoreRunes`, `SmithWatermanGotohRawScoreWithParams`
  - Constructor/config: `NewSWGParams`, `NewScorer`, `DefaultScorer`, `DefaultScorerOptions`, `DefaultNormalisationOptions`, `DefaultTokeniseOptions`, `AlgoIDs`
  - Utility: `Normalise`, `Tokenise`
  - Scorer methods: `Scorer.Score` (method examples), `Scorer.Match`, `Scorer.Threshold`, `Scorer.Algorithms`, `Scorer.ScoreAll`
- **Standard:** `go-testing-standards/SKILL.md` §Benchmark Tests — "One example per algorithm + Scorer + Normalise + Tokenise." `documentation-standards/SKILL.md` §Runnable Examples — "One example per algorithm function appearing on pkg.go.dev."
- **Action:** Code fix — add Example functions for the listed public symbols. The Scorer method examples can be grouped as `ExampleScorer_Score`, `ExampleScorer_Match`, etc.
- **Rationale:** Missing Examples mean those functions have no entry on pkg.go.dev. Consumers discovering the API via the reference docs cannot see executable usage.

---

### [Important] Missing BDD feature files: `normalisation.feature`, `determinism.feature`, `tokenise.feature`

- **File:** `tests/bdd/features/` (missing files)
- **Phase introduced:** Phases 1 (tokenise), 5 (normalisation), 5 (determinism)
- **Issue:** The testing standard lists tags `@normalisation` and `@determinism` as BDD test categories, implying dedicated feature files. `tokenise.feature` is also absent — `Tokenise` is a public function with documented semantics that a consumer should be able to read as Gherkin. Note: `normalisation.feature` and `determinism.feature` are also flagged by `bdd-scenario-FINDINGS.md` (Critical Finding #1); this entry captures `tokenise.feature` as an additional missing file not in that review.
- **Standard:** `go-testing-standards/SKILL.md` §BDD Tests — "Feature files in `tests/bdd/features/` — one file per capability." Tags list includes `@normalisation`, `@determinism`.
- **Action:** Code fix — create `tokenise.feature` with scenarios covering camelCase splitting, snake_case, kebab-case, empty input, mixed-script, and the `NoEmptyTokens` invariant.
- **Rationale:** `Tokenise` is a consumer-facing function documented in `docs/requirements.md`. A consumer reading the BDD suite cannot discover its behaviour contract from Gherkin.

---

### [Important] Missing category tags (`@character`, `@qgram`, `@phonetic`, `@gestalt`) in Phase 2–5 BDD feature files

- **File:** All Phase 2–5 feature files (levenshtein, hamming, jaro, jarowinkler, damerau_osa, damerau_full, swg, strcmp95, lcsstr, ratcliff_obershelp, qgram_jaccard, sorensen_dice, cosine, tversky)
- **Phase introduced:** Phases 2–5
- **Issue:** The testing standard lists the following BDD filter tags: `@character`, `@qgram`, `@token`, `@phonetic`, `@gestalt`, `@scorer`, `@scan`, `@suppression`, `@normalisation`, `@determinism`. Phase 6–8 feature files use tags (`@token`, `@scorer`, etc.). Phase 2–5 feature files have NO tags at all — neither at the `Feature:` level nor at individual `Scenario:` level. This means `godog --tags=@character` matches zero scenarios, making the tag-based filtering system non-functional for character-tier algorithms.
- **Standard:** `go-testing-standards/SKILL.md` §BDD Tests — "Tags for filtering: `@character`, `@qgram`, ... Tag every scenario with its category."
- **Action:** Code fix — add `@character` to the Feature level in: `levenshtein.feature`, `hamming.feature`, `jaro.feature`, `jarowinkler.feature`, `damerau_osa.feature`, `damerau_full.feature`, `swg.feature`, `strcmp95.feature`, `lcsstr.feature`. Add `@gestalt` to `ratcliff_obershelp.feature`. Add `@qgram` to `qgram_jaccard.feature`, `sorensen_dice.feature`, `cosine.feature`, `tversky.feature`.
- **Rationale:** CI filtering by algorithm category (e.g. "run only phonetic BDD tests") is unusable without tags. The tag standard was established during Phase 6 but never backfilled.

---

### [Important] BDD feature files lack Unicode edge-case scenarios for Phase 2–3 character-tier algorithms

- **File:** `tests/bdd/features/levenshtein.feature`, `hamming.feature`, `jaro.feature`, `jarowinkler.feature`, `damerau_osa.feature`, `damerau_full.feature`, `swg.feature`, `strcmp95.feature`, `lcsstr.feature`
- **Phase introduced:** Phases 2–3
- **Issue:** The testing standard mandates "Unicode edge cases (multi-byte UTF-8, CJK, emoji)" as required BDD scenarios. Reviewing all Phase 2–3 feature files: none contains a Unicode scenario (café/cafe, CJK pairs, or emoji inputs). The unit tests in `*_test.go` do have UTF-8 test cases, but the BDD contract — the consumer-facing specification — is silent on Unicode behaviour. A consumer reading `levenshtein.feature` cannot determine whether `LevenshteinScore("café", "cafe")` operates at byte or rune level.
- **Standard:** `go-testing-standards/SKILL.md` §BDD Tests — "Unicode edge cases (multi-byte UTF-8, CJK, emoji)" listed as required BDD scenarios.
- **Action:** Code fix — add at least one `@unicode` scenario per character-tier feature file showing the rune vs byte distinction (e.g. `LevenshteinScoreRunes("café","cafe") = 0.75` vs `LevenshteinScore("café","cafe")` with documented byte behaviour).
- **Rationale:** Without a Unicode BDD scenario, the rune-path guarantee is invisible to consumers reading the feature files as documentation. This creates an API surprise for any consumer processing non-ASCII text.

---

### [Important] Missing fuzz seed corpus lone-surrogate and NUL-byte entries for algorithm fuzz tests

- **File:** All algorithm `*_fuzz_test.go` files except `normalise_fuzz_test.go`
- **Phase introduced:** Phases 2–7
- **Issue:** The testing standard requires fuzz harnesses to assert "Never panics on arbitrary input including invalid UTF-8, embedded NULs, lone surrogates." `normalise_fuzz_test.go` correctly seeds with `"\xed\xa0\x80"` (lone surrogate) and `"a\x00b"` (embedded NUL). No other algorithm fuzz test seeds these specific pathological byte sequences. While the fuzzer will eventually generate them, seeding them explicitly ensures they are exercised on every CI run (seed-only mode) without requiring a full fuzz session.
- **Standard:** `go-testing-standards/SKILL.md` §Fuzz Tests — "Never panics on arbitrary input including invalid UTF-8, embedded NULs, lone surrogates."
- **Action:** Code fix — add `"\xed\xa0\x80"` (lone surrogate) and `"a\x00b"` (embedded NUL) to the `f.Add(...)` seed calls in each character-tier, q-gram tier, and token-tier fuzz harness.
- **Rationale:** Lone surrogates are valid bytes that the Go fuzzer will generate but only after many iterations. Seeding them guarantees deterministic coverage of the most dangerous UTF-8 corner case on every test run.

---

### [Important] `DamerauLevenshteinFullScore_ZeroAllocs_ASCII_Short` test is permanently skipped

- **File:** `damerau_full_test.go:229`
- **Phase introduced:** Phase 2
- **Issue:** `TestDamerauLevenshteinFullScore_ZeroAllocs_ASCII_Short` uses `t.Skipf` with a note that the 0-alloc ASCII fast path is a v1.x optimisation. The skip is not gated by a build tag or flag; it skips unconditionally on every test run. This test is dead code that counts toward the test total but provides no coverage. More importantly, the skip message says "replace this Skipf with the actual AllocsPerRun assertion" — there is no tracking issue preventing this from being forgotten entirely.
- **Standard:** `go-testing-standards/SKILL.md` §Unit Tests — tests must provide coverage.
- **Action:** Discuss-phase needed — either (a) create a GitHub issue tracking the v1.x optimisation and add `//nolint:golint // skipped until #NNN: ...` with the issue reference, or (b) convert the skip to a soft assertion that logs a warning without failing if allocs > 0, preserving the test intent.
- **Rationale:** A permanently-skipped test contributes to the "pass percentage" illusion while providing no actual verification. Without a linked issue, it will never be resolved.

---

### [Important] Scorer benchmarks missing Medium/Long variants for `ScoreAll` and `Match`

- **File:** `scorer_bench_test.go`
- **Phase introduced:** Phase 8
- **Issue:** `scorer_bench_test.go` has Short, Medium, and Long variants for `Score` but only a Short variant for `ScoreAll` (`BenchmarkDefaultScorer_ScoreAll_ASCII_Short`) and only a Short variant for `Match` (`BenchmarkDefaultScorer_Match_ASCII_Short`). The standard says "short / medium / long input benchmarks (10 / 50 / 500 characters)". `ScoreAll` calls all six dispatch functions simultaneously; its behaviour at long inputs is qualitatively different from `Score` and warrants its own multi-size benchmark series. There are also no benchmarks for a custom-configured Scorer (non-default algorithm set).
- **Standard:** `go-testing-standards/SKILL.md` §Benchmark Tests — "short / medium / long input benchmarks."
- **Action:** Code fix — add `BenchmarkDefaultScorer_ScoreAll_ASCII_Medium`, `BenchmarkDefaultScorer_ScoreAll_ASCII_Long`, `BenchmarkDefaultScorer_Match_ASCII_Medium`, `BenchmarkDefaultScorer_Match_ASCII_Long`.
- **Rationale:** Without Medium/Long benchmarks for `ScoreAll`, a performance regression in a single dispatched algorithm at larger inputs would not appear in the benchstat comparison.

---

### [Important] Missing example tests for `Scorer` methods on pkg.go.dev

- **File:** `example_test.go`
- **Phase introduced:** Phase 8
- **Issue:** `Scorer` is the primary user-facing composite API. The five public methods (`Score`, `Match`, `Threshold`, `Algorithms`, `ScoreAll`) have no `Example*` functions. On pkg.go.dev, the `Scorer` type page will show method signatures without runnable examples. Consumers cannot see the idiomatic construction-and-use pattern without reading the unit tests. The standard says "One example per algorithm + Scorer + Normalise + Tokenise."
- **Standard:** `documentation-standards/SKILL.md` — runnable examples on pkg.go.dev; `go-testing-standards/SKILL.md` §Meta-tests.
- **Action:** Code fix — add `ExampleNewScorer`, `ExampleScorer_Score`, `ExampleScorer_Match`, `ExampleScorer_ScoreAll`, `ExampleDefaultScorer` to `example_test.go`.
- **Rationale:** pkg.go.dev examples are the primary discoverability surface for open-source Go libraries. The most important public type lacking examples is the highest-priority addition.

---

### [Important] Missing example tests for `Normalise` and `Tokenise`

- **File:** `example_test.go`
- **Phase introduced:** Phase 1
- **Issue:** `Normalise` and `Tokenise` are foundational utilities that affect all algorithm results. Neither has an `Example*` function in `example_test.go`. Consumers configuring `NormalisationOptions` or `TokeniseOptions` cannot see canonical usage patterns on pkg.go.dev.
- **Standard:** `go-testing-standards/SKILL.md` §Benchmark Tests — "One example per algorithm + Scorer + Normalise + Tokenise."
- **Action:** Code fix — add `ExampleNormalise` and `ExampleTokenise` showing the default options path and a customised options path.
- **Rationale:** These functions are how consumers tune the library for their input domain. Missing examples are a documentation gap for the most frequently customised surface.

---

### [Important] Missing example tests for `NewSWGParams` and `SmithWatermanGotohScoreWithParams`

- **File:** `example_test.go`
- **Phase introduced:** Phase 3
- **Issue:** `SmithWatermanGotohScoreWithParams` and `SmithWatermanGotohRawScoreWithParams` are the primary customisation entry points for SWG. `NewSWGParams` is the constructor for the parameter struct. None of these three functions has an `Example*`. The existing `ExampleSmithWatermanGotohScore` shows the default-params path only.
- **Standard:** `documentation-standards/SKILL.md` §Runnable Examples.
- **Action:** Code fix — add `ExampleNewSWGParams` showing custom match/mismatch/gap parameters, and `ExampleSmithWatermanGotohScoreWithParams` showing the result on a pair that diverges from default params.
- **Rationale:** Custom params are the main reason a consumer would choose SWG over a simpler algorithm. Missing examples make the customisation interface invisible.

---

### [Important] Missing example tests for Distance functions (`LevenshteinDistance`, `HammingDistance`)

- **File:** `example_test.go`
- **Phase introduced:** Phase 2
- **Issue:** `LevenshteinDistance` and `HammingDistance` are public functions returning raw integer distances (not normalised scores). They have no `Example*` function. Consumers who need the raw edit distance rather than the normalised score have no documented usage pattern. `DamerauLevenshteinOSADistance` and `DamerauLevenshteinFullDistance` are similarly unexampled.
- **Standard:** `documentation-standards/SKILL.md` §Runnable Examples.
- **Action:** Code fix — add `ExampleLevenshteinDistance`, `ExampleHammingDistance` at minimum. DL variants can be added as part of the same PR.
- **Rationale:** The Distance variants are used differently from the Score variants (for threshold-by-count rather than threshold-by-ratio decisions). The absence of examples makes this usage invisible.

---

### [Important] Missing example tests for `LongestCommonSubstringRunes` and rune-path Score variants

- **File:** `example_test.go`
- **Phase introduced:** Phases 2–6
- **Issue:** The pattern of exposing both byte and rune-path variants is central to the library's Unicode story. While some rune variants have examples (`ExampleLCSStrScoreRunes`, `ExampleQGramJaccardScoreRunes`), many are missing: `LevenshteinScoreRunes`, `HammingScoreRunes`, `JaroScoreRunes`, `JaroWinklerScoreRunes`, `DamerauLevenshteinOSAScoreRunes`, `DamerauLevenshteinFullScoreRunes`. These are the functions a consumer would call for correct multi-byte UTF-8 handling.
- **Standard:** `documentation-standards/SKILL.md` §Runnable Examples.
- **Action:** Code fix — add `ExampleLevenshteinScoreRunes`, `ExampleHammingScoreRunes` at minimum, each showing a multi-byte UTF-8 pair (e.g. café/cafe) with the rune-aware result and a note explaining the divergence from the byte path.
- **Rationale:** Without examples showing the byte vs rune divergence on a concrete multi-byte input, consumers processing non-ASCII text will default to the byte-path function and get incorrect results without realising it.

---

### [Important] Scorer `ScoreAll` deterministic iteration order is untested at the property level

- **File:** `scorer_test.go`
- **Phase introduced:** Phase 8
- **Issue:** `ScoreAll` returns `map[AlgoID]float64`. The determinism standard (`determinism-standards/SKILL.md`) mandates "Extract keys, sort them, iterate the sorted slice" on all output paths. The `ScoreAll` method returns a map (documented as having non-deterministic iteration order). There is no property test asserting that two consecutive calls to `s.ScoreAll(a, b)` return maps with bit-identical values for all keys. The concurrent test (`TestScorer_ConcurrentSafety`) does a structural comparison but not a property test with randomised inputs.
- **Standard:** `determinism-standards/SKILL.md` — determinism guarantee; `go-testing-standards/SKILL.md` §Property Tests.
- **Action:** Code fix — add `TestProp_Scorer_ScoreAll_DeterministicAcrossRuns` using `testing/quick` to assert all map values are bit-identical on repeated calls.
- **Rationale:** If a future refactor introduces any non-deterministic element to the dispatch layer (e.g. a sorted-keys optimisation with a bug), `ScoreAll` would be the first surface where it appears.

---

### [Important] `ratcliff_obershelp` cross-validation corpus lacks rune-path entries

- **File:** `testdata/cross-validation/ratcliff-obershelp/vectors.json`
- **Phase introduced:** Phase 4
- **Issue:** The 16-entry Ratcliff-Obershelp cross-validation corpus contains only ASCII pairs. `RatcliffObershelpScoreRunes` operates on multi-byte UTF-8 inputs with a distinct code path. There are no corpus entries using non-ASCII inputs (e.g. café/cafe, CJK pairs). The rune-path unit tests have hand-computed reference vectors but no external cross-validation. `difflib.SequenceMatcher` in Python 3 operates at the Unicode character level, making it the natural reference for the rune path.
- **Standard:** `algorithm-correctness-standards/SKILL.md` §Reference Vectors — "For Unicode-aware algorithms: at least one non-ASCII case."
- **Action:** Code fix — add at least three non-ASCII pair entries to `testdata/cross-validation/ratcliff-obershelp/vectors.json` using `difflib.SequenceMatcher(autojunk=False)` computed on the Unicode character sequence.
- **Rationale:** The rune path's correctness against the difflib reference is currently verified only for ASCII. A rune-boundary handling bug would not be caught by the existing corpus.

---

### [Important] `DamerauLevenshteinFullScore` `t.Skip` test is untracked

- **File:** `damerau_full_test.go:229`
- **Phase introduced:** Phase 2
- **Issue:** Already noted above as an Important finding. The skip text says "see plan 02-06 SUMMARY" but there is no GitHub issue number in the skip message and no TODO comment in the standard format `// TODO(#N): ...`. This means the v1.x optimisation is untraceable via `gh issue list`.
- **Standard:** `CLAUDE.md` — "Every `TODO` must reference a GitHub issue: `// TODO(#42): ...`"
- **Action:** Code fix — create a GitHub issue for the v1.x two-row DL-Full optimisation and add the issue number to the `t.Skipf` message and a `// TODO(#N): ...` comment above the test.
- **Rationale:** Without a linked issue, the skipped test will never be revisited.

---

## Improvement Findings

### [Improvement] `testdata/fuzz/` subdirectories have only one seed file each

- **File:** `testdata/fuzz/*/seed-001`
- **Phase introduced:** Phases 2–5
- **Issue:** Every on-disk fuzz corpus directory contains exactly one seed file (`seed-001`). The normalise and tokenise fuzz tests have 5 seeds each (reflecting their more complex input space). The algorithm fuzz tests would benefit from 3–5 seeds covering: both-empty, one-empty, identical ASCII, invalid UTF-8 (high byte), and multi-byte UTF-8 (café/cafe). More seeds mean better deterministic coverage on every CI run (seed-only mode).
- **Standard:** `go-testing-standards/SKILL.md` §Fuzz Tests — "Seed corpus."
- **Action:** Code fix — add seed-002 through seed-005 to each algorithm's corpus directory, covering the edge-case classes listed above.
- **Rationale:** One seed file exercises one input class. Five seeds ensure all five edge-case classes are covered on every CI seed-only run, not just the full fuzz session.

---

### [Improvement] Property test for `RatcliffObershelpScore` asymmetry is in `cross_algorithm_consistency_test.go` but not in `props_test.go`

- **File:** `cross_algorithm_consistency_test.go:380`; missing in `props_test.go`
- **Phase introduced:** Phase 4
- **Issue:** The asymmetry pin for `RatcliffObershelpScore` (`TestCrossAlgorithm_RatcliffObershelp_AsymmetryPin`) lives in `cross_algorithm_consistency_test.go` alongside the cross-algorithm tests. Per the project convention, algorithm-specific invariants belong in `props_test.go` (or the algorithm's `_test.go`) and cross-algorithm relationship tests belong in `cross_algorithm_consistency_test.go`. The asymmetry pin is an algorithm-specific property, not a cross-algorithm relationship. Having it in the cross-algorithm file makes it harder to discover when reading the algorithm's own test file.
- **Standard:** `go-testing-standards/SKILL.md` §Unit Tests — file placement conventions.
- **Action:** Code fix — move `TestCrossAlgorithm_RatcliffObershelp_AsymmetryPin` to `props_test.go` and rename it `TestProp_RatcliffObershelpScore_Asymmetric`. Keep the cross-algorithm `TestCrossAlgorithm_RatcliffObershelp_PinnedAgainstDifflib` where it is (that one tests the difflib equivalence relationship).
- **Rationale:** Test placement conventions exist so reviewers know where to look for coverage of a given function. An asymmetry property buried in the cross-algorithm file breaks the convention.

---

### [Improvement] `props_test.go` mixes per-algorithm tests that should be in algorithm-specific test files

- **File:** `props_test.go`
- **Phase introduced:** Phases 2–8 (accumulated)
- **Issue:** `props_test.go` is 3700+ lines and contains property tests for all 23 algorithms. Some property tests are algorithm-specific (e.g. `TestProp_SmithWatermanGotoh_GapSplitInvariance`, `TestProp_MongeElkanScore_AsymmetricWhenTokenCountAsymmetric`) that belong logically alongside the unit tests in the algorithm's own `_test.go` file or a dedicated `swg_props_test.go`. The current monolith makes it difficult to navigate and increases the risk that a new algorithm's properties are added in the wrong place or omitted.
- **Standard:** `go-testing-standards/SKILL.md` §Unit Tests — "File placement: `levenshtein_test.go` beside `levenshtein.go`."
- **Action:** Improvement (no behaviour change) — consider splitting algorithm-specific property tests into per-algorithm `<algo>_props_test.go` files (or appending them to the existing `<algo>_test.go`). The common cross-cutting properties (range bounds, identity, symmetry) can remain in `props_test.go` as a shared property harness.
- **Rationale:** At Phase 9 with the scan package adding more property tests, `props_test.go` will continue to grow. Splitting now prevents the file from becoming unmanageable.

---

### [Improvement] No `bench.txt` baseline comparison for Phase 6–8 algorithms

- **File:** `bench.txt`
- **Phase introduced:** Phases 6–8
- **Issue:** `bench.txt` exists but was last committed before Phase 6 algorithms were added (TokenSortRatio, TokenSetRatio, PartialRatio, TokenJaccard, MongeElkan — Phase 6; Soundex, DoubleMetaphone, NYSIIS, MRA — Phase 7; Scorer — Phase 8). The benchstat comparison `make bench-compare` will find no baseline entries for new algorithms, making regression detection ineffective for them.
- **Standard:** `go-testing-standards/SKILL.md` §Benchmark Tests — "`bench.txt` committed per release; CI runs benchstat against the last tagged release."
- **Action:** Code fix — regenerate `bench.txt` to include all current algorithms and commit the updated baseline.
- **Rationale:** benchstat cannot detect a regression against a missing baseline. Every algorithm added after the last `bench.txt` update is invisible to the regression detector.

---

### [Improvement] `testdata/golden/_staging/` files are undocumented as a workflow artefact

- **File:** `testdata/golden/_staging/` (no README or comment)
- **Phase introduced:** Phase 2
- **Issue:** The `_staging/` directory contains 23 per-algorithm golden JSON files that feed `TestGolden_Algorithms_Merge`. There is no documentation (README, inline comment, or CONTRIBUTING note) explaining the lifecycle: staging files are created by individual algorithm tests, promoted by `TestGolden_Algorithms_Merge` into `algorithms.json`, and then the staging file remains. A new contributor adding an algorithm does not know (a) that they must create a staging file, (b) how the file format is determined, or (c) what triggers promotion. The 9 algorithms with staging files but no staging test (Phase 6–7 gap noted in Critical findings) would be invisible to such a contributor.
- **Standard:** `documentation-standards/SKILL.md` — documentation for all non-obvious workflows.
- **Action:** Improvement — add a `testdata/golden/README.md` (or a comment in `golden_test.go`) explaining the staging lifecycle. Also add a meta-test that asserts every file in `_staging/` has a corresponding `TestGolden_*_Staging` test function (parseable via `go/parser`).
- **Rationale:** An undocumented staging workflow will be skipped by contributors who discover the gap too late (after their algorithm is merged without a staging test).

---

### [Improvement] `cross_algorithm_consistency_test.go` is 823 lines and growing; a Phase 9 scan-readiness hook is missing

- **File:** `cross_algorithm_consistency_test.go`
- **Phase introduced:** Phase 6
- **Issue:** The cross-algorithm consistency file adds new test functions with each phase. At Phase 9 (scan), the scan sub-package will introduce `scan.Check` which calls `Scorer.Score` across item pairs. The cross-algorithm layer should include at least one placeholder structure (a stub function with a `t.Skip`) for the scan sub-package's cross-algorithm property: `PropCheck_BucketEquivalentToNaive`. Without this placeholder, Phase 9 will have to add a new test file to the root package for the cross-algorithm scan test, creating inconsistency.
- **Standard:** `go-testing-standards/SKILL.md` §Scan-level properties — "`scan.Check` output sorted by `(Kind, NameA, NameB, GroupA, GroupB)`; `Bucket-equivalent-to-naive`: property test proves equivalence to naive O(N²) pairwise scoring."
- **Action:** Improvement — add a stub `TestCrossAlgorithm_Scan_BucketEquivalentToNaive` with `t.Skip("Phase 9")` in `cross_algorithm_consistency_test.go` or create `scan/scan_test.go` in the (not-yet-created) `scan/` package. The Phase 9 team should see the placeholder and know the pattern.
- **Rationale:** Establishing the test structure before the code exists ensures Phase 9 does not have to reverse-engineer the property test pattern from scratch.

---

### [Improvement] Fuzz harness naming for phonetic functions is inconsistent with algorithm naming

- **File:** `soundex_fuzz_test.go`, `nysiis_fuzz_test.go`, `mra_fuzz_test.go`, `double_metaphone_fuzz_test.go`
- **Phase introduced:** Phase 7
- **Issue:** The fuzz harness names for phonetic algorithms do not follow the pattern of the primary score function:
  - `FuzzSoundex` (should be `FuzzSoundexCode` or `FuzzSoundexScore` to match the primary function name)
  - `FuzzNYSIIS` (should be `FuzzNYSIISCode` or `FuzzNYSIISScore`)
  - `FuzzMRA` (should be `FuzzMRACode` or `FuzzMRAScore`)
  - `FuzzDoubleMetaphone` (matches primary function `DoubleMetaphoneKeys` less directly)
  Character-tier and q-gram-tier harnesses are named after the primary public function (e.g. `FuzzLevenshteinScore`, `FuzzQGramJaccardScore`). Phonetic harnesses use abbreviated algorithm names instead.
- **Standard:** `go-testing-standards/SKILL.md` — "Every public function has a fuzz harness in `fuzz_test.go`." The naming convention implies the harness name matches the function name.
- **Action:** Improvement — rename `FuzzSoundex` → `FuzzSoundexCode` (since it covers both `SoundexCode` and `SoundexScore`), `FuzzNYSIIS` → `FuzzNYSIISCode`, `FuzzMRA` → `FuzzMRACode`. Update corpus directory names accordingly (`testdata/fuzz/FuzzSoundexCode/`). This is a rename only; no functional change.
- **Rationale:** Consistent naming allows `grep "FuzzLevenshteinScore\|FuzzSoundexScore"` to work predictably and makes the harness-to-function mapping obvious.

---

### [Improvement] No goroutine-leak detector (`goleak`) in the root package `TestMain`

- **File:** missing: root package `TestMain`
- **Phase introduced:** Phase 1 (should have been considered at bootstrap)
- **Issue:** `tests/bdd/bdd_test.go` uses `goleak.VerifyTestMain(m)`. The root test package has no `TestMain` function and therefore no goroutine-leak detection. Since the library is pure-function and imports no packages that launch goroutines, the risk is low. However, if any future contributor accidentally introduces a goroutine leak (e.g. in a background table-initialisation, a leaked timer in an allocation budget test, or via a transitive dependency), it will not be caught in the root package test suite.
- **Standard:** `go-testing-standards/SKILL.md` §Goroutine Leak Detection — "The library is pure-function; the test catches accidental introduction of background work." The standard only mandates goleak in `tests/bdd/`; this is a defence-in-depth improvement for the root.
- **Action:** Improvement — add a `main_test.go` with `TestMain(m *testing.M)` using `goleak.VerifyTestMain(m)` to the root package test suite. Note: goleak is a test-only dependency; it must remain in `tests/bdd/go.mod` only and cannot be added to the root `go.mod`. This means the root `TestMain` must either implement a lightweight goroutine snapshot-and-compare or accept the limitation.
- **Rationale:** Since goleak cannot be in the root `go.mod`, this is genuinely an improvement rather than a required fix. The correct approach is either to implement goroutine detection using `runtime.NumGoroutine()` snapshots (stdlib) or document explicitly that leak detection is BDD-only.

---

## Phase 9 Scan-Readiness Architectural Notes

The following are rough edges in the current test architecture that should be resolved before Phase 9 (scan sub-package) to avoid architectural debt:

1. **`scorer_options_test.go` naming** (Critical Finding #9): Rename to `scorer_options_internal_test.go` before Phase 9 creates scan-package option tests that need to follow the same pattern.

2. **Missing `internal_coverage_test.go`** (Critical Finding #11): Create before Phase 9 so the coverage floor is enforced for both the root package and the scan sub-package from day one.

3. **Missing scan placeholder in `cross_algorithm_consistency_test.go`** (Improvement #5): Add before Phase 9 so the scan team has a clear pattern to follow.

4. **BDD `scan.feature` and `suppression.feature`** (Critical in `bdd-scenario-FINDINGS.md`): These are Phase 9 blockers. The BDD infrastructure must be ready to receive scan scenarios before scan implementation begins.

5. **`goleak` in root TestMain** (Improvement #8): The scan sub-package will have its own `tests/bdd/` but may also add root-level integration tests. Decide the goroutine-detection strategy before Phase 9 introduces goroutine patterns (the scan package itself is pure-function, but test helpers may use background goroutines for parallel corpus scanning).
