---
status: issues_found
agent: bdd-scenario-reviewer
scope: entire BDD suite (phases 1-8)
reviewed: 2026-05-17T00:00:00Z
finding_counts:
  critical: 12
  important: 11
  improvement: 14
  total: 37
---

# BDD Scenario Coverage Review — Phases 1–8

**Reviewed:** 2026-05-17
**Feature files:** `tests/bdd/features/` (24 files)
**Step definitions:** `tests/bdd/steps/algorithms_steps.go`, `tests/bdd/steps/scorer_steps.go`
**Test runner:** `tests/bdd/bdd_test.go`
**Requirements cross-reference:** `docs/requirements.md §15.6`, `.claude/skills/go-testing-standards/SKILL.md` (BDD section)
**Prior review incorporated:** `.planning/phases/08-composite-scorer/08-BDD-REVIEW.md`

---

## Infrastructure Assessment (pre-finding)

**goleak gate:** WIRED. `tests/bdd/bdd_test.go:37` calls `goleak.VerifyTestMain(m)` before the suite runs.

**go.mod isolation:** CORRECT. `tests/bdd/go.mod` lists godog v0.15.0, goleak v1.3.0, testify v1.10.0. Root `go.mod` has zero non-stdlib require lines.

**testify confinement:** CORRECT. Only `tests/bdd/` imports testify; root tests use stdlib `testing`.

**Missing feature files (spec-required, non-existent):** `normalisation.feature`, `determinism.feature`, `scan.feature`, `suppression.feature` — all four are mandated by `docs/requirements.md §15.6` and are absent. This is the most severe structural gap in the suite.

---

## Critical Findings

### [Critical] Four mandatory feature files are entirely absent

- **File:** `tests/bdd/features/` (missing: `normalisation.feature`, `determinism.feature`, `scan.feature`, `suppression.feature`)
- **Phase introduced:** Phases 5 (normalisation), 5 (determinism), 6/9 (scan), 6/9 (suppression)
- **Issue:** `docs/requirements.md §15.6` explicitly names six feature files. Only two are present — `algorithms.feature` (spread across per-algorithm files) and `scorer.feature`. The following four are entirely absent: `normalisation.feature`, `determinism.feature`, `scan.feature`, `suppression.feature`. The Normalisation Pipeline (§9), the Determinism Guarantees (§13), the Scan sub-package (§12), and Suppression composition (§12.3) have zero BDD coverage. These represent entire documented capability layers with no consumer-facing contract documentation or regression detection at the BDD layer.
- **Standard:** `go-testing-standards/SKILL.md` §BDD — "Every algorithm, every Scorer composition pattern, every scan suppression behaviour gets a BDD scenario." `docs/requirements.md §15.6` lists all six feature files as requirements.
- **Action:** Code fix — create all four missing feature files with the scenario sets described in the BDD checklist sections of the agent system prompt.
- **Rationale:** A consumer reading the BDD suite as documentation cannot discover the library's normalisation semantics, determinism guarantees, scan behaviour, or suppression rules from any Gherkin. These are entire undocumented contracts.

---

### [Critical] `ErrInvalidThreshold` has no BDD scenario (carried from 08-BDD-REVIEW.md BDD-01)

- **File:** `tests/bdd/features/scorer.feature`
- **Phase introduced:** Phase 8
- **Issue:** `ErrInvalidThreshold` is a documented sentinel in `errors.go`. The step infrastructure in `scorer_steps.go:338-339` already has the case branch for it. Three other sentinel-error scenarios exist (`ErrMissingThreshold`, `ErrEmptyScorer`, `ErrInvalidWeight`), but `ErrInvalidThreshold` is missing. A consumer passing threshold=1.5 or threshold=-0.1 sees no documented contract at the BDD layer. Additionally, the step regex for threshold in error-path scenarios is `(\d+\.?\d*)` which only matches non-negative numbers — a negative-threshold sub-case cannot be expressed in Gherkin without a regex fix.
- **Standard:** `go-testing-standards/SKILL.md` — "Error scenarios test exact sentinel error matching via `errors.Is`." `docs/requirements.md §8.1` lists `ErrInvalidThreshold` as a construction validation sentinel.
- **Action:** Code fix — add scenario in `scorer.feature`; widen the threshold regex to `(-?\d+\.?\d*)`.
- **Rationale:** An undocumented sentinel is an invisible API contract. The step infrastructure is already present; this is a one-scenario gap.
- **Suggested fix:**
```gherkin
@scorer @errors
Scenario: WithThreshold out-of-range returns ErrInvalidThreshold
  When I attempt to construct a Scorer with Levenshtein weight 1.0 and threshold 1.5
  Then constructing the Scorer should return ErrInvalidThreshold
```

---

### [Critical] `ErrInvalidAlgoID` has no BDD scenario

- **File:** `tests/bdd/features/scorer.feature`
- **Phase introduced:** Phase 8
- **Issue:** `docs/requirements.md §8.1` lists `ErrInvalidAlgoID` as a construction sentinel for: invalid algorithm ID passed to `WithAlgorithm`, and Monge-Elkan inner = `AlgoMongeElkan` (self-reference). Neither sub-case has a BDD scenario. The `constructingTheScorerShouldReturn` step in `scorer_steps.go:327` does not have a case for `ErrInvalidAlgoID` — the switch falls through to the `default` error path, so even if a scenario were written it would fail at the step level.
- **Standard:** `docs/requirements.md §6` sentinel errors documentation; `go-testing-standards/SKILL.md` error-path coverage requirement.
- **Action:** Code fix — add `ErrInvalidAlgoID` case to `scorer_steps.go:constructingTheScorerShouldReturn`, then add scenario in `scorer.feature`.
- **Rationale:** A documented sentinel with no BDD coverage and no step support is invisible to consumers and undetectable by the BDD regression suite.

---

### [Critical] `ErrInvalidQGramSize` has no BDD scenario

- **File:** `tests/bdd/features/scorer.feature`
- **Phase introduced:** Phase 8
- **Issue:** `docs/requirements.md §8.1` lists `ErrInvalidQGramSize` as a construction sentinel (q-gram size < 1 via `WithQGramJaccardAlgorithm`). No scenario exercises this. No step exists for it. The `constructingTheScorerShouldReturn` switch has no case for `ErrInvalidQGramSize`.
- **Standard:** `docs/requirements.md §6` sentinel errors documentation.
- **Action:** Code fix — add step support and scenario.
- **Rationale:** Same pattern as `ErrInvalidAlgoID` — documented sentinel, no coverage.

---

### [Critical] `ErrInvalidTverskyParam` has no BDD scenario

- **File:** `tests/bdd/features/scorer.feature`
- **Phase introduced:** Phase 8
- **Issue:** `docs/requirements.md §8.1` lists `ErrInvalidTverskyParam` as a construction sentinel (Tversky α or β < 0 via `WithTverskyAlgorithm`). No scenario exercises this. No step exists for it. The `constructingTheScorerShouldReturn` switch has no case for `ErrInvalidTverskyParam`.
- **Standard:** `docs/requirements.md §6` sentinel errors documentation.
- **Action:** Code fix — add step support and scenario.
- **Rationale:** Same pattern as above.

---

### [Critical] `iScoreTheSamePairWithTheDefaultScorer` hardcodes the pair (carried from 08-BDD-REVIEW.md BDD-02)

- **File:** `tests/bdd/steps/scorer_steps.go` lines 189–195; `tests/bdd/features/scorer.feature` lines 98–102
- **Phase introduced:** Phase 8
- **Issue:** The step bound to `^I score the same pair with the default Scorer$` ignores all state from the preceding `When I score "XMLParser" and "xml_parser" with the Scorer` step and instead hardcodes `sc.defaultScorer.Score("XMLParser", "xml_parser")` directly. This is a confirmed step bug: the natural-language contract of "the same pair" is violated. The test passes today only because the hardcoded values coincidentally match the scenario's When arguments. A future scenario reusing this step phrase with different inputs will silently score the wrong pair and produce misleading results. `ScorerContext` does not store `lastA`/`lastB` fields.
- **Standard:** `go-testing-standards/SKILL.md` — "Step functions return error rather than calling t.Errorf." The step must honour its Gherkin contract.
- **Action:** Code fix — add `lastA`, `lastB string` fields to `ScorerContext`; populate in `iScoreAndWithTheScorer`; rewrite `iScoreTheSamePairWithTheDefaultScorer` to use `sc.defaultScorer.Score(sc.lastA, sc.lastB)`.
- **Rationale:** This is a confirmed bug causing the scenario to silently test the wrong inputs. The scenario currently passes only by accidental coincidence.

---

### [Critical] Weight auto-normalisation is never directly asserted (carried from 08-BDD-REVIEW.md BDD-03)

- **File:** `tests/bdd/features/scorer.feature` lines 58–67
- **Phase introduced:** Phase 8
- **Issue:** The "Two-algorithm weighted Scorer composes correctly" scenario uses equal weights (0.5 + 0.5) on an identical pair (`"hello"` vs `"hello"`). With equal weights, auto-normalisation is a no-op — the weights already sum to 1.0. On an identical pair, both algorithms return 1.0 regardless of weights. This scenario does not exercise weight normalisation. The mandatory class 4 ("Custom 2-algorithm Scorer with explicit weights") is documented as requiring verification that the composite is the weighted sum, but the current scenario would pass even if normalisation were completely broken.
- **Standard:** `docs/requirements.md §8.4` — weight normalisation is a core semantic; `go-testing-standards/SKILL.md` BDD contract requirement.
- **Action:** Code fix — add a scenario with unequal raw weights on a non-identical pair with a pinned expected composite.
- **Suggested fix:**
```gherkin
@scorer @custom
Scenario: Weight auto-normalisation produces sum-to-1 composite
  # Levenshtein("kitten","sitting") ≈ 0.5714; JaroWinkler("kitten","sitting") ≈ 0.7468.
  # Composite with normalised weights (0.3 + 0.7 → already 1.0): 0.3×0.5714 + 0.7×0.7468 ≈ 0.6942.
  Given I construct a Scorer with Levenshtein weight 0.3 and JaroWinkler weight 0.7 and threshold 0.5
  When I score "kitten" and "sitting" with the Scorer
  Then the Scorer composite score should be approximately 0.6942 within 0.001
  And the Scorer match result should be true
```

---

### [Critical] No Unicode reference vector for Levenshtein

- **File:** `tests/bdd/features/levenshtein.feature`
- **Phase introduced:** Phase 2
- **Issue:** `docs/requirements.md §7.1.1` and the skill's Per-Algorithm Coverage Checklist both require at least one multi-byte UTF-8 example for Unicode-aware algorithms. Levenshtein exposes a `LevenshteinScoreRunes` variant; the feature file has zero Unicode scenarios. The reference example from the checklist template (`"東京"` vs `"東都"` → 0.5) is absent.
- **Standard:** `go-testing-standards/SKILL.md` BDD Per-Algorithm Coverage Checklist — "For Unicode-aware algorithms: at least one multi-byte UTF-8 example."
- **Action:** Code fix — add a Unicode rune-path scenario.
- **Suggested fix:**
```gherkin
Scenario: Unicode rune-level variant (CJK)
  When I compute the LevenshteinRunes score between "東京" and "東都"
  Then the score should be approximately 0.5000 within 0.0001
```

---

### [Critical] No Unicode reference vector for DL-OSA or DL-Full

- **File:** `tests/bdd/features/damerau_osa.feature`, `tests/bdd/features/damerau_full.feature`
- **Phase introduced:** Phase 2
- **Issue:** Both algorithms expose `*ScoreRunes` variants. Neither feature file has a Unicode scenario. The rune-path variant is entirely uncovered at the BDD layer.
- **Standard:** Same as Levenshtein above.
- **Action:** Code fix — add one multi-byte UTF-8 scenario per feature file.

---

### [Critical] DL-OSA has no one-empty scenario

- **File:** `tests/bdd/features/damerau_osa.feature`
- **Phase introduced:** Phase 2
- **Issue:** `docs/requirements.md §7.1.2` states edge cases are "identical to Levenshtein" — which includes one-empty → 0.0. The `damerau_osa.feature` has a both-empty scenario but NO one-empty scenario. The `damerau_full.feature` also lacks a one-empty scenario. The skill checklist requires both-empty and one-empty cases for every algorithm.
- **Standard:** `go-testing-standards/SKILL.md` BDD Per-Algorithm Coverage Checklist.
- **Action:** Code fix — add `Scenario: one-empty string scores 0.0` to both files.

---

### [Critical] SWG has no Unicode reference vector and no raw-score surface coverage

- **File:** `tests/bdd/features/swg.feature`
- **Phase introduced:** Phase 3
- **Issue:** SWG exposes `SmithWatermanGotohScoreRunes` and `SmithWatermanGotohRawScore`/`SmithWatermanGotohRawScoreWithParams`. The feature file has zero Unicode scenarios and zero `*RawScore*` scenarios. `docs/requirements.md §7.1.8` specifically documents the `*RawScore*` variant surface and its rationale ("advanced consumers who want absolute alignment quality"). The BDD suite is the consumer-facing contract; if `*RawScore*` has no scenario, consumers have no documented example of its semantics.
- **Standard:** `docs/requirements.md §7.1.8` requires documentation of both normalised and raw surfaces; skill checklist requires Unicode coverage.
- **Action:** Code fix — add a Unicode rune-path scenario and at least one `*RawScore*` scenario demonstrating that raw scores can differ from clamped scores.

---

### [Critical] Strcmp95 has no Unicode scenario (and is ASCII-only — document it explicitly)

- **File:** `tests/bdd/features/strcmp95.feature`
- **Phase introduced:** Phase 4
- **Issue:** `docs/requirements.md §7.1.7` states Strcmp95 is "ASCII-only; for Unicode input, normalise via `fuzzymatch.Normalise` first." The feature file does not document this ASCII-only restriction with a scenario that demonstrates the required pre-normalisation step. A consumer passing Unicode input directly to `Strcmp95Score` without normalisation will get silent degraded behaviour; there is no BDD documentation of this contract. The SKILL.md checklist requires documenting algorithm-specific behaviour in a separate Scenario.
- **Standard:** Skill checklist — "Scenario for any algorithm-specific behaviour."
- **Action:** Code fix — add a scenario explicitly documenting the ASCII-only restriction and the recommended normalisation pre-step.

---

## Important Findings

### [Important] `WithNormalisation(custom)` composed path has no scenario (carried from 08-BDD-REVIEW.md BDD-04)

- **File:** `tests/bdd/features/scorer.feature`
- **Phase introduced:** Phase 8
- **Issue:** `docs/requirements.md §8.2` lists `WithNormalisation(opts NormalisationOptions)` as a core Scorer option. The `WithoutNormalisation()` path is covered, the default-normalisation path is covered, but `WithNormalisation(custom opts)` — the consumer hook for diacritic-stripping, custom separator chars, or CamelCase-split control — has no scenario.
- **Standard:** Scorer Coverage Checklist — "Scenario: custom Scorer with normalisation options."
- **Action:** Code fix — add scenario and supporting step.

---

### [Important] `ScoreAll` map values are never range-checked (carried from 08-BDD-REVIEW.md BDD-05)

- **File:** `tests/bdd/features/scorer.feature` lines 140–151
- **Phase introduced:** Phase 8
- **Issue:** The "ScoreAll returns map keyed by AlgoID" scenario verifies only key presence/absence, never the values. No assertion that values are in `[0.0, 1.0]`, and no concrete value assertion for any specific algorithm on the tested pair. The `ScoreAll` contract includes that values are per-algorithm scores (not weighted contributions), which is never verified at the BDD layer.
- **Standard:** Scorer Coverage Checklist — ScoreAll returns per-algorithm breakdown with correct values.
- **Action:** Code fix — add value range assertion step and at least one concrete value assertion.

---

### [Important] Hamming has no one-empty scenario

- **File:** `tests/bdd/features/hamming.feature`
- **Phase introduced:** Phase 2
- **Issue:** The feature has both-empty and unequal-length scenarios but no scenario for `HammingScore("abc", "")` returning 0.0. The one-empty case for Hamming is distinct from the unequal-length case (one-empty IS also unequal-length, but the documentation separately defines "one empty" as an edge case in §7.1.4).
- **Standard:** SKILL.md BDD checklist — one-empty case required.
- **Action:** Code fix — add scenario.

---

### [Important] Hamming canonical reference vector table has only 2 rows; the Hamming 1950 paper reference "1011101"/"1001001" is missing

- **File:** `tests/bdd/features/hamming.feature`
- **Phase introduced:** Phase 2
- **Issue:** `docs/requirements.md §7.1.4` lists reference vectors including `"1011101"/"1001001" → distance 2, score ≈ 0.7143`. The Examples table has only `karolin/kathrin` and `abc/abc`. The binary string vector from Hamming 1950 is absent. The skill checklist requires "at least 3 canonical reference vectors from the primary source."
- **Standard:** SKILL.md — "At least 3 canonical reference vectors from the primary source."
- **Action:** Code fix — add the binary-string reference vector.

---

### [Important] Jaro reference vector table is thin (3 from primary source; checklist requires at least 3 but identity is in the table, not from Jaro 1989)

- **File:** `tests/bdd/features/jaro.feature`
- **Phase introduced:** Phase 2
- **Issue:** The Examples table has MARTHA/MARHTA, DIXON/DICKSONX, JELLYFISH/SMELLYFISH, ABC/ABC. The ABC/ABC row is an identity case, not a Jaro 1989 reference vector. The three Jaro 1989 vectors are present, but the checklist note about identity being a separate row rather than a reference-vector row means the table has exactly 3 literature vectors. This is borderline; a stronger table would include an additional non-trivial vector from Winkler 1990 p.357 to cross-validate.
- **Standard:** SKILL.md — "At least 3 canonical reference vectors from the primary source."
- **Action:** Improvement — the minimum is technically met; adding one more literature vector would strengthen coverage.

---

### [Important] SWG cross-validation corpus not exercised via BDD

- **File:** `tests/bdd/features/swg.feature`
- **Phase introduced:** Phase 3
- **Issue:** `docs/requirements.md §7.1.8` documents a committed JSON corpus at `testdata/cross-validation/swg/vectors.json` with 16 entries, cross-validated against biopython's `Bio.Align.PairwiseAligner`. The BDD feature file exercises only 2 reference vectors (http_request/http_request_header_fields and abc/abc), neither of which is from the cross-validation corpus. The Gotoh-erratum canary (`one_long_gap_canary` at biopython_normalised=0.5) is mentioned in the requirements but has no BDD scenario. The gap-split canary in the BDD file uses a different pair than the committed corpus.
- **Standard:** SKILL.md — "Cross-validation scenarios (SWG)."
- **Action:** Code fix — add at least the Gotoh-erratum canary pair as a load-bearing scenario.

---

### [Important] Ratcliff-Obershelp has no symmetry scenario AND no documentation that it is intentionally asymmetric

- **File:** `tests/bdd/features/ratcliff_obershelp.feature`
- **Phase introduced:** Phase 4
- **Issue:** The feature file omits the symmetry scenario "per OQ-1 resolution (LOCKED 2026-05-14)." The comment in the feature file explains this in a code comment, but there is no Gherkin scenario documenting the asymmetry contract for consumers. A consumer reading only the BDD scenarios does not learn that `RatcliffObershelpScore(a, b) != RatcliffObershelpScore(b, a)` is expected behaviour, not a bug. The skill checklist notes that algorithm-specific behaviour should have its own Scenario. The asymmetry is load-bearing and should be documented in a BDD scenario (not just a code comment), even if the scenario is a "then these two scores should differ" assertion rather than a symmetry gate.
- **Standard:** SKILL.md — "Scenario for any algorithm-specific behaviour."
- **Action:** Code fix — add a scenario explicitly demonstrating and documenting the asymmetric-by-design behaviour.
- **Suggested fix:**
```gherkin
Scenario: Score is intentionally asymmetric (difflib CPython bpo-37004 parity)
  # RatcliffObershelp is asymmetric by design — mirrors Python difflib behaviour.
  # This scenario documents the asymmetry is expected, not a bug.
  When I compute the Ratcliff-Obershelp score between "WIKIMEDIA" and "WIKIMANIA"
  And I compute the second Ratcliff-Obershelp score between "WIKIMANIA" and "WIKIMEDIA"
  Then the two Ratcliff-Obershelp scores should differ
```

---

### [Important] Monge-Elkan non-permitted inner panic list is incomplete

- **File:** `tests/bdd/features/monge_elkan.feature` lines 152–160
- **Phase introduced:** Phase 6
- **Issue:** The Examples table for "non-permitted inner AlgoIDs panic" shows only 3 of the 5 rejected entries: `MongeElkan`, `TokenSortRatio`, `TokenJaccard`. Missing: `TokenSetRatio` and `PartialRatio`. `docs/requirements.md §8.1` and the feature file's own comment list all 5. A regression where `TokenSetRatio` or `PartialRatio` is accidentally added to the permitted set would not be caught by this scenario.
- **Standard:** Completeness requirement — all documented rejection cases should be covered.
- **Action:** Code fix — add the two missing rows to the Examples table.

---

### [Important] Ratcliff-Obershelp step `iComputeTheSecondRatcliffObershelpScore` does not exist

- **File:** `tests/bdd/steps/algorithms_steps.go`
- **Phase introduced:** Phase 4
- **Issue:** The algorithms_steps.go explicitly documents (lines 357–365) that "the symmetry scenario is OMITTED per OQ-1 resolution (LOCKED 2026-05-14), so no 'second' / 'equal' steps exist." However, the finding above (asymmetry documentation) recommends adding a "two scores should differ" scenario. That scenario would require a `iComputeTheSecondRatcliffObershelpScore` step. No such step exists. Adding the asymmetry-documentation scenario requires both the feature file addition and a new step.
- **Standard:** Step-definition completeness.
- **Action:** Code fix — add `iComputeTheSecondRatcliffObershelpScoreBetween` step alongside the asymmetry scenario.

---

### [Important] LCSStr has no Unicode scenario despite exposing a `*Runes` variant

- **File:** `tests/bdd/features/lcsstr.feature`
- **Phase introduced:** Phase 4
- **Issue:** `LCSStrScoreRunes` is a public function. The feature file covers only the byte path. The SKILL.md checklist requires a Unicode scenario for Unicode-aware algorithms.
- **Standard:** SKILL.md BDD Per-Algorithm Coverage Checklist.
- **Action:** Code fix — add a multi-byte UTF-8 scenario for the rune-path variant.

---

### [Important] `algoIDFromName` in scorer_steps.go uses `AlgoID.String()` reverse-lookup which requires the string form to match exactly

- **File:** `tests/bdd/steps/scorer_steps.go` lines 471–486
- **Phase introduced:** Phase 8
- **Issue:** `algoIDFromName` iterates `fuzzymatch.AlgoIDs()` and compares via `id.String() == stripped`. If `AlgoID.String()` returns `"damerau_levenshtein_osa"` (snake_case) but the Gherkin scenario uses `"AlgoDamerauLevenshteinOSA"` (PascalCase), the lookup fails. The scorer.feature scenario uses `AlgoDamerauLevenshteinOSA` and `AlgoDoubleMetaphone` — with the `Algo` prefix stripped this becomes `DamerauLevenshteinOSA` and `DoubleMetaphone`. These must match `AlgoID.String()` exactly. If the `String()` method returns `"damerau_levenshtein_osa"` (snake_case), the lookup would return an error at runtime, not at compile time. This is a latent brittle coupling — any rename of an AlgoID's string form silently breaks step lookups.
- **Standard:** Test reliability — step definitions should not have hidden string-matching brittleness.
- **Action:** Discuss-phase — verify that `AlgoID.String()` returns PascalCase (not snake_case) for the step lookup to work, or add a test that explicitly exercises `algoIDFromName` for every `AlgoID`.

---

## Improvement Findings

### [Improvement] No `@character`, `@qgram`, `@token`, `@phonetic`, `@gestalt` category tags on most algorithm feature files

- **File:** All `tests/bdd/features/*.feature` files except `monge_elkan.feature`, `monge_elkan_phonetic_inner.feature`, `token_sort_ratio.feature`, `token_set_ratio.feature`, `token_jaccard.feature`, `partial_ratio.feature`, `nysiis.feature`, `mra.feature`
- **Phase introduced:** Phases 2–7
- **Issue:** The SKILL.md BDD section requires every scenario to be tagged with at least one tag. Most algorithm feature files (`levenshtein.feature`, `hamming.feature`, `jaro.feature`, `jarowinkler.feature`, `strcmp95.feature`, `swg.feature`, `lcsstr.feature`, `damerau_osa.feature`, `damerau_full.feature`, `qgram_jaccard.feature`, `sorensen_dice.feature`, `cosine.feature`, `tversky.feature`, `soundex.feature`, `double_metaphone.feature`, `ratcliff_obershelp.feature`) have zero tags on their scenarios. Only the Phase 6 token algorithms and Phase 7 phonetic algorithms have category tags.
- **Standard:** SKILL.md — "Every scenario MUST have at least one tag."
- **Action:** Code fix — add category tags (`@character`, `@qgram`, `@gestalt`, `@phonetic`) and layer tags (`@algorithm`) to all scenarios in the untagged feature files.

---

### [Improvement] `WithNormaliseWeights(false)` opt-out has no BDD scenario

- **File:** `tests/bdd/features/scorer.feature`
- **Phase introduced:** Phase 8
- **Issue:** `docs/requirements.md §8.4` documents `WithNormaliseWeights(false)` as an opt-in escape hatch that produces a raw weighted sum. No scenario exercises this path. The CONTEXT.md golden corpus has a mandatory entry for this case.
- **Standard:** Scorer Coverage Checklist — "Scenario: weight normalisation behaviour (`WithNormaliseWeights(true)` default, `(false)` opt-out)."
- **Action:** Code fix — add scenario.

---

### [Improvement] `DefaultScorer()` composition is never explicitly verified at BDD layer

- **File:** `tests/bdd/features/scorer.feature`
- **Phase introduced:** Phase 8
- **Issue:** The Scorer Coverage Checklist requires "Scenario: `DefaultScorer` matches a known similar pair above its threshold" (covered — scenario 1) and also that the default composition is documented. The actual algorithm list (DL-OSA 0.30, JaroWinkler 0.20, TokenJaccard 0.20, QGramJaccard 0.15, SorensenDice 0.10, DoubleMetaphone 0.05) is never verified at the BDD layer. A scenario asserting `Algorithms()` returns exactly these 6 algorithms with approximately these weights would pin the default composition contract.
- **Standard:** Scorer Coverage Checklist.
- **Action:** Code fix — add scenario asserting DefaultScorer algorithm count and algorithm membership.

---

### [Improvement] Concurrent scenario tests `Score` only; `ScoreAll` and `Match` concurrent safety undocumented

- **File:** `tests/bdd/features/scorer.feature` lines 129–138
- **Phase introduced:** Phase 8
- **Issue:** `docs/requirements.md §8.1` states "`Score`, `ScoreAll`, and `Match` are safe for concurrent use." The concurrent scenario tests only `Score`. `ScoreAll` and `Match` concurrent safety have no BDD documentation.
- **Standard:** Scorer Coverage Checklist — "Scenario: concurrent Score / ScoreAll / Match is safe."
- **Action:** Code fix — add concurrent `ScoreAll` and `Match` scenarios (or expand the existing scenario to include all three methods).

---

### [Improvement] `WithoutNormalisation` scenario uses relative assertion; no concrete value (carried from 08-BDD-REVIEW.md BDD-07)

- **File:** `tests/bdd/features/scorer.feature` lines 90–102
- **Phase introduced:** Phase 8
- **Issue:** The assertion `the no-normalisation composite should be less than the default composite` is relative, not concrete. If both scores changed proportionally (e.g. a bug that halved all scores), the relative assertion would still pass while the absolute behaviour had changed.
- **Standard:** Concrete value assertions anchor contracts better than relative ones.
- **Action:** Code fix — add concrete range assertions from the golden file.

---

### [Improvement] Scorer scenarios have no cross-layer `@levenshtein`, `@jaro_winkler` tags

- **File:** `tests/bdd/features/scorer.feature`
- **Phase introduced:** Phase 8
- **Issue:** Scorer scenarios that compose specific algorithms (e.g. "Single-algorithm Scorer composes correctly" with Levenshtein) have no secondary tag linking them to the algorithm's own feature file. This makes filtered runs by algorithm incomplete — `godog --tags=@levenshtein` would not find the Scorer-layer Levenshtein scenario.
- **Standard:** SKILL.md tag taxonomy.
- **Action:** Improvement — add per-algorithm secondary tags to relevant Scorer scenarios.

---

### [Improvement] SWG `both SmithWatermanGotoh scores should be equal` step is algorithm-specific but the gap-split canary's assertion is not the gap-split property — it checks symmetry instead

- **File:** `tests/bdd/features/swg.feature` lines 43–46
- **Phase introduced:** Phase 3
- **Issue:** The "gap-split canary" scenario asserts that two different SWG inputs (`abc________def/abcdef` vs `abc____def____/abcdef`) produce the same score. This is not a gap-split invariance test — it is a symmetry test on two distinct structurally similar pairs. The true Gotoh gap-split invariant from `docs/requirements.md §7.1.8` states that "splitting a single long gap into two halves with intervening match characters that don't affect the local alignment must NOT improve the score." The current scenario does not exercise this specific invariant.
- **Standard:** Algorithm-specific invariant coverage.
- **Action:** Improvement — clarify the scenario comment or replace with a true gap-split invariant test.

---

### [Improvement] Double Metaphone "Slavic" and "Chinese-origin" gate scenarios assert only non-emptiness, not actual key values

- **File:** `tests/bdd/features/double_metaphone.feature` lines 61–68
- **Phase introduced:** Phase 7
- **Issue:** The Slavic (Sczepanski) and Chinese-origin (Cheung) gate scenarios assert only `both keys should be non-empty`. This is a very weak assertion — any non-empty output would pass, including wildly wrong keys. The SKILL.md checklist requires at least 3 canonical reference vectors from the primary source; these two scenarios do not contribute reference-vector coverage.
- **Standard:** SKILL.md — "At least 3 canonical reference vectors from the primary source."
- **Action:** Improvement — replace non-empty assertions with concrete expected key values from cross-validated sources.

---

### [Improvement] Soundex feature file has no `@phonetic` tag on most scenarios

- **File:** `tests/bdd/features/soundex.feature`
- **Phase introduced:** Phase 7
- **Issue:** The Feature-level declaration has no `@phonetic` tag, and individual scenarios are untagged. NYSIIS and MRA have `@phonetic` tags; Soundex and Double Metaphone are inconsistent.
- **Standard:** SKILL.md tag requirement.
- **Action:** Code fix — add `@phonetic` and `@soundex` tags to all scenarios.

---

### [Improvement] Double Metaphone feature has no `@phonetic` tag on the Feature declaration or most scenarios

- **File:** `tests/bdd/features/double_metaphone.feature`
- **Phase introduced:** Phase 7
- **Issue:** Same as Soundex above — no `@phonetic @double_metaphone` tags.
- **Standard:** SKILL.md tag requirement.
- **Action:** Code fix — add tags.

---

### [Improvement] `theDistanceShouldBe` step is algorithm-agnostic by design but this creates potential cross-scenario state confusion

- **File:** `tests/bdd/steps/algorithms_steps.go` lines 144–153
- **Phase introduced:** Phase 2
- **Issue:** The step comment documents that `theDistanceShouldBe` "matches the value written by whichever *Distance* step ran most recently in the current scenario." This is intentional but means if a scenario chains two distance computations, the assertion applies only to the last one. No scenario currently does this, but the design is fragile. The comment recommends introducing algorithm-suffixed steps if needed, but no guidance exists in the feature files about this limitation.
- **Standard:** Step-definition clarity.
- **Action:** Improvement — add a comment in each feature file that uses distance assertions noting this design constraint, to prevent future misuse.

---

### [Improvement] `iListTheScorerAlgorithms` is a no-op step — the When step does no work

- **File:** `tests/bdd/steps/scorer_steps.go` lines 222–227
- **Phase introduced:** Phase 8
- **Issue:** `iListTheScorerAlgorithms` validates `sc.scorer != nil` and returns nil. The subsequent assertion steps call `sc.scorer.Algorithms()` directly on each call rather than reading from state populated by the When step. This violates the Given/When/Then partition — the When step should produce observable state that Then steps assert on.
- **Standard:** Gherkin readability and correct Given/When/Then partitioning.
- **Action:** Improvement — make the step populate a `lastAlgorithms []fuzzymatch.ScorerAlgorithm` field in `ScorerContext`; have assertion steps read from that field.

---

### [Improvement] Ratcliff-Obershelp autojunk-sensitive scenario uses hardcoded pair construction in Go step, not in Gherkin Examples

- **File:** `tests/bdd/features/ratcliff_obershelp.feature` lines 53–64; `tests/bdd/steps/algorithms_steps.go` lines 370–408
- **Phase introduced:** Phase 4
- **Issue:** The autojunk-sensitive scenario uses a dedicated step `I compute the Ratcliff-Obershelp score for the autojunk-sensitive pair` that constructs the 205-char inputs via `strings.Repeat` in Go. This is technically correct (the comment explains why — 205-char Gherkin literals are hard to read). However, this means the pair is invisible to consumers reading the BDD documentation. The Gherkin should at minimum describe the pair construction in a comment.
- **Standard:** Consumer-facing documentation quality.
- **Action:** Improvement — add a comment in the Gherkin scenario describing the pair construction logic so a reader does not need to inspect the step definition.

---

### [Improvement] Scorer `algoIDFromName` and Monge-Elkan `algoIDByName` are two separate lookup functions doing the same thing

- **File:** `tests/bdd/steps/scorer_steps.go` lines 471–486; `tests/bdd/steps/algorithms_steps.go` lines 835–887
- **Phase introduced:** Phase 6 (algoIDByName), Phase 8 (algoIDFromName)
- **Issue:** Two separate AlgoID name→enum lookup functions exist in the step definitions: `algoIDByName` (switch-based, returns -1 on unknown) and `algoIDFromName` (iterates `AlgoIDs()`, returns error on unknown). They are not identical: `algoIDByName` is a switch with explicit cases; `algoIDFromName` iterates `AlgoIDs()`. If a new AlgoID is added, `algoIDByName` requires a new case but `algoIDFromName` picks it up automatically. This creates a maintenance coupling — `algoIDByName` can silently miss new AlgoIDs while `algoIDFromName` won't.
- **Standard:** DRY step definitions — no duplication.
- **Action:** Improvement — consolidate to a single lookup function, preferring the `AlgoIDs()` iterator approach of `algoIDFromName` for automatic completeness.

---

## Summary

### Feature file presence vs requirements

| Required file (§15.6) | Status |
|---|---|
| `algorithms.feature` (per-algorithm, split across 23 files) | Present — partially complete |
| `scorer.feature` | Present — gaps documented above |
| `normalisation.feature` | **ABSENT** |
| `determinism.feature` | **ABSENT** |
| `scan.feature` | **ABSENT** |
| `suppression.feature` | **ABSENT** |

### Gap count by severity

| Severity | Count |
|---|---|
| Critical | 12 |
| Important | 11 |
| Improvement | 14 |
| **Total** | **37** |

### Recommendation: NO-GO

The BDD suite has four structurally missing feature files (`normalisation.feature`, `determinism.feature`, `scan.feature`, `suppression.feature`) that represent entire undocumented capability layers. It has a confirmed step bug (`iScoreTheSamePairWithTheDefaultScorer` hardcodes its pair — from 08-BDD-REVIEW.md BDD-02) that causes a scenario to silently test wrong inputs. Three documented sentinel errors (`ErrInvalidAlgoID`, `ErrInvalidQGramSize`, `ErrInvalidTverskyParam`) have no BDD scenarios and no step infrastructure. `ErrInvalidThreshold` has no scenario (step infrastructure exists). Weight auto-normalisation is not actually exercised by the current two-algorithm scenario.

These deficiencies mean the BDD suite cannot serve as a reliable consumer-facing contract or regression detection layer for a significant portion of the documented API. The four missing feature files alone represent the largest gap — they are required by `docs/requirements.md §15.6` and must be created before Phase 9 (scan) work begins.

_Reviewed: 2026-05-17_
_Reviewer: bdd-scenario-reviewer_
