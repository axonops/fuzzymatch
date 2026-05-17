---
status: issues_found
agent: test-analyst
scope: entire test suite (phases 1-8)
reviewed: 2026-05-17T00:00:00Z
finding_counts:
  critical: 11
  important: 18
  improvement: 13
  total: 42
---

# fuzzymatch — Whole-Codebase Test-Suite Health Analysis

Holistic analysis covering every `_test.go` file in the root package
plus the BDD sub-module (`tests/bdd/`). Expands the previous
phase-scoped review in `.planning/phases/08-composite-scorer/08-TEST-ANALYSIS.md`
to all phases 1-8.

## Headline numbers

- **Package coverage:** 90.4% (target 95% overall — **fails**).
- **Per-file coverage:** `double_metaphone.go` 83.8% (target ≥ 90% — **fails**); all other files ≥ 90%.
- **Public API coverage:** ≥ 99% on every exported Score/Code/Compare function (every exported Score is exercised in `props_test.go`, `*_test.go`, examples, and at least one fuzz harness — though several `*Runes` variants and the `Scorer` lack fuzz harnesses, see below).
- **Total tests:** ~480 top-level `Test*` functions plus 39 `Example*` plus 31 `Fuzz*` plus 127 `Benchmark*`.
- **BDD scenarios:** 25 feature files (one per algorithm + monge_elkan_phonetic_inner + scorer); 169 scenarios total. **No `scan.feature`, no `suppression.feature`, no `normalisation.feature`, no `determinism.feature`** — listed in `docs/requirements.md` §15.6 and the prompt scan/normalisation checklists. (Scan sub-package is unimplemented Phase 6 work — out-of-scope this milestone; normalisation and determinism scenarios remain a documented gap.)
- **Fuzz corpus directories on disk:** 20 (`testdata/fuzz/`). Fuzz harness functions in source: 31. Discrepancy: 11 fuzzers have no committed seed corpus.
- **BDD pass:** `cd tests/bdd && go test ./...` → ok.
- **Race tests:** `go test -race ./...` → ok (~352s, passes).

## Per-Algorithm Coverage Summary

| Algorithm | Unit | Property | Fuzz (byte/rune) | Bench | BDD | Golden | Cross-validation |
|-----------|------|----------|------------------|-------|-----|--------|------------------|
| Levenshtein | OK | OK + Triangle | byte only | OK | OK (5) | OK | impl-internal |
| DamerauLevenshteinOSA | OK | OK + Triangle | byte only | OK | OK (4) | OK | impl-internal |
| DamerauLevenshteinFull | OK | OK + Triangle | byte only | OK | OK (5) | OK | impl-internal |
| Hamming | OK | OK + Triangle (eq-len) | byte only | OK | OK (6) | OK | impl-internal |
| Jaro | OK | OK (no triangle — doc'd) | byte only | OK | OK (4) | OK | impl-internal |
| JaroWinkler | OK | OK + AtLeastJaro | byte only | OK | OK (6) | OK | impl-internal |
| Strcmp95 | OK | OK + AtLeastJW | byte only | OK | OK (6) | OK | impl-internal |
| SmithWatermanGotoh | OK | OK | byte only (Score) | OK | OK (6) | OK | OK (vectors.json) |
| LCSStr | OK | OK | byte only | OK | OK (8) | OK | impl-internal |
| QGramJaccard | OK | OK | byte + Runes | OK | OK (6) | OK | impl-internal |
| SorensenDice | OK | OK | byte + Runes | OK | OK (6) | OK | impl-internal |
| Cosine | OK | OK | byte + Runes | OK | OK (7) | OK | impl-internal |
| Tversky | OK | OK + Asymmetry | byte + Runes | OK | OK (7) | OK | OK (vs Jaccard/Dice) |
| MongeElkan | OK | OK + Asymmetry (flake risk) | byte (Score + Symmetric) | OK | OK (9) + phonetic (6) | OK | impl-internal |
| TokenSortRatio | OK | OK | byte | OK | OK (6) | OK | OK (vectors.json) |
| TokenSetRatio | OK | OK | byte | OK | OK (7) | OK | OK (vectors.json) |
| PartialRatio | OK | OK | byte + Runes | OK | OK (14) | OK | OK (vectors.json) |
| TokenJaccard | OK | OK | byte | OK | OK (6) | OK | impl-internal |
| Soundex | OK | OK (no NoNegativeZero) | byte | OK | OK (7) | OK | OK (vectors.json) |
| DoubleMetaphone | OK | OK (no NoNegativeZero) | byte | OK | OK (9) | OK | OK (vectors.json) |
| NYSIIS | OK | OK (no NoNegativeZero) | byte | OK | OK (10) | OK | OK (vectors.json) |
| MRA | OK | OK (no NoNegativeZero) | byte | OK | OK (10) | OK | OK (vectors.json) |
| RatcliffObershelp | OK | OK | byte | OK | OK (5) | OK | OK (vectors.json) |

All 23 catalogue algorithms have unit + property + at least one fuzz +
benchmark + BDD + cross-platform golden coverage.

---

## Findings

### [Critical] No Scorer fuzz harness — `FuzzDefaultScorerScore`, `FuzzDefaultScorerScoreAll`, `FuzzDefaultScorerMatch` all missing
- **File:** `scorer_fuzz_test.go` (would-be path; file does not exist)
- **Phase introduced:** Phase 8
- **Issue:** `.claude/skills/go-testing-standards/SKILL.md` §Fuzz Tests: "Every public function has a fuzz harness in `fuzz_test.go`." 31 fuzz functions exist, none target the `Scorer.Score`, `Scorer.ScoreAll`, `Scorer.Match`, or `NewScorer` surfaces. `docs/requirements.md` §15.4 explicitly lists `FuzzScorer` as required.
- **Standard:** `.claude/skills/go-testing-standards/SKILL.md` §Fuzz Tests; `docs/requirements.md` §15.4
- **Action:** Code fix
- **Rationale:** The Scorer is the most-used public surface but the only public surface without a fuzz gate. Bug-rich path: pre-normalisation, AlgoID dispatch, reduction loop, threshold compare.
- **Suggested fix:** Three small harnesses seeding with `user_id / userId`, both-empty, `\xff\xfe`, etc. — same pattern as `levenshtein_fuzz_test.go`. (Identical to 08-TEST-ANALYSIS TEST-15.)

### [Critical] `WithThreshold(NaN)` not rejected — silent malfunction not gated by test
- **File:** `scorer_options.go:259`; absent test in `scorer_options_test.go`
- **Phase introduced:** Phase 8
- **Issue:** `t < 0.0 || t > 1.0` is both `false` for `math.NaN()`; the option silently accepts NaN and the resulting Scorer's `Match` returns `false` for every input (NaN comparisons always false). No unit test exists for NaN input.
- **Standard:** `.claude/skills/algorithm-correctness-standards/SKILL.md` §Edge cases (NaN/Inf handling); `docs/requirements.md` §13 (determinism — NaN must be rejected at boundaries)
- **Action:** Code fix (production code + test)
- **Rationale:** Same finding as 08-TEST-ANALYSIS TEST-01; production bug + missing test combined.
- **Suggested fix:** Use `math.IsNaN(t) || t < 0.0 || t > 1.0`; add `TestWithThreshold_RejectsNaN`.

### [Critical] `WithTverskyAlgorithm(α=0, β=0)` not rejected — direct dispatch will panic at runtime
- **File:** `scorer_options.go` (WithTverskyAlgorithm); absent test in `scorer_options_test.go:433-444`
- **Phase introduced:** Phase 8
- **Issue:** Option layer accepts `α==0, β==0` even though direct call `TverskyScore(_, _, _, 0, 0)` panics (see `tversky_test.go:568-582`). The Scorer's "fail at construction time, never at Score time" contract is violated; no test gates this gap.
- **Standard:** `.claude/skills/algorithm-correctness-standards/SKILL.md` §Edge cases; Tversky's `ErrInvalidTverskyParam` documented in `errors.go`
- **Action:** Code fix (production code + test)
- **Rationale:** Same as 08-TEST-ANALYSIS TEST-02; production-side bug compounds with the test-coverage gap.
- **Suggested fix:** Gate `α == 0 && β == 0` in `WithTverskyAlgorithm`; add `TestWithTverskyAlgorithm_RejectsBothZero`.

### [Critical] `WithAlgorithm(_, NaN)` and `WithAlgorithm(_, +Inf)` not rejected
- **File:** `scorer_options.go:152`; absent test in `scorer_options_test.go:56-73`
- **Phase introduced:** Phase 8
- **Issue:** `weight <= 0` is `false` for NaN (NaN comparisons always false) and for +Inf. The option layer accepts both; resulting Scorer produces NaN/Inf composites. The existing `TestWithAlgorithm_InvalidWeight` covers `-1.0`, `-0.5`, `0`, `-100` but not NaN/Inf.
- **Standard:** `.claude/skills/algorithm-correctness-standards/SKILL.md` §Edge cases (NaN/Inf), `.claude/skills/determinism-standards/SKILL.md` (NaN propagation)
- **Action:** Code fix
- **Rationale:** Same class of defect as the WithThreshold gap. Silent malfunction; existing PropScorer_NoNaN_NoInf would catch it on `Score` output but not at option time.
- **Suggested fix:** Add `if math.IsNaN(weight) || math.IsInf(weight, 0) { return ErrInvalidWeight }`; add `TestWithAlgorithm_RejectsNaNInfWeight`.

### [Critical] Three required meta-tests missing: `internal_coverage_test.go`, `documentation_test.go`, `readme_shop_front_test.go`
- **File:** (would-be paths; files do not exist)
- **Phase introduced:** Phase 1 (bootstrap)
- **Issue:** `.claude/skills/go-testing-standards/SKILL.md` §Meta-tests requires all five meta-tests; only `ai_friendly_test.go` and `makefile_targets_test.go` exist. `internal_coverage_test.go` is the project's automated 95%/90%/100% coverage-floor gate — without it the standards' coverage target is unenforced and the 90.4% overall coverage shortfall slips by silently.
- **Standard:** `.claude/skills/go-testing-standards/SKILL.md` §Meta-tests; `docs/requirements.md` §15.7
- **Action:** Code fix
- **Rationale:** The coverage-floor gate is the project's load-bearing automatic enforcement of the 95% overall / 90% per-file / 100% public-API targets. Without it, drift is invisible.
- **Suggested fix:** Mirror `axonops/mask`'s pattern — read `go.test` coverage output via `go tool cover -func` and assert numeric thresholds. Documentation_test parses code blocks in `docs/*.md` and verifies they compile.

### [Critical] `double_metaphone.go` per-file coverage 83.8% — below the 90% per-file floor
- **File:** `double_metaphone.go`
- **Phase introduced:** Phase 4
- **Issue:** `dmPrep` 58.8%, `DoubleMetaphoneKeys` 62.0%, `dmAdd` 84.6%, `dmSlgCheck` 88.9%, `DoubleMetaphoneScore` 92.3%. Average 83.8%; below the 90% per-file standard floor.
- **Standard:** `.claude/skills/go-testing-standards/SKILL.md` line 16 ("≥ 90% per file")
- **Action:** Code fix (more unit tests for Slavic/Germanic/Spanish/Italian/French branches in `dmPrep` and uncovered table cells in `DoubleMetaphoneKeys`)
- **Rationale:** Phonetic-rule branches in DM are notoriously hard to exhaust; many branches are language-specific and only fired by surnames the unit tests don't include. RV-DM and language-branch unit tests cover 38 names but the per-file budget needs ~60+ to bring `dmPrep` to ≥ 90%.
- **Suggested fix:** Add reference vectors from `phonetic` cross-validation corpus that fire the uncovered branches (e.g. "Brzęczyszczykiewicz" for `dmPrep` Slavic path; Italian "GLI" prefix; French "ILLE" suffix). The cross-validation vectors.json already lists candidate names.

### [Critical] Overall coverage 90.4% — below the 95% overall floor
- **File:** Root package
- **Phase introduced:** Inherited from Phase 4 (phonetic) + Phase 6 (tokenisation)
- **Issue:** `go test -coverprofile=coverage.out ./... && go tool cover -func=coverage.out` reports 90.4% (after race test 90.5%). Standards require ≥ 95% overall.
- **Standard:** `.claude/skills/go-testing-standards/SKILL.md` line 16 ("≥ 95% overall"); `docs/requirements.md` §15.8
- **Action:** Code fix
- **Rationale:** Headline gap. Primarily traces to double_metaphone (83.8%), then tokenise (93.1%), nysiis (90.9%), token_set_ratio (92.0%), soundex (92.3%). No single file is the cause — the standard requires improvement across phonetic + tokenisation surfaces.
- **Suggested fix:** Add language-branch reference vectors and exercise the residual error-path branches (e.g. soundex non-ASCII edge cases, tokenise containsNonASCII early-return).

### [Critical] `TestProp_Scorer_WeightSumOne` `uint16` overflow flake — `u+1` overflows to 0 when `u==65535`
- **File:** `scorer_test.go:827`
- **Phase introduced:** Phase 8
- **Issue:** `return float64(u+1) / float64(uint32(1)<<16) * 100.0` — `u` is `uint16` so `u+1` overflows to `0`, returning weight `0`, which the option layer rejects via `ErrInvalidWeight`. Property test then returns `false` on that draw. Per-run flake probability is ~0.46% (3 vars × 100 quick.Check iters × 1/65536 each), enough to surface intermittently in CI.
- **Standard:** `.claude/skills/go-testing-standards/SKILL.md` (Property tests must be deterministic-passing); 08-REVIEW WR-01
- **Action:** Code fix
- **Rationale:** Same as 08-TEST-ANALYSIS TEST-07; ONE-LINE fix still unaddressed.
- **Suggested fix:** `return (float64(u) + 1.0) / 65536.0 * 100.0` — float arithmetic, no integer wrap.

### [Critical] Scorer benchmarks NOT in `bench.txt` — regression gate is inactive
- **File:** `bench.txt`
- **Phase introduced:** Phase 8
- **Issue:** `bench.txt` is the committed benchstat baseline; `grep -i scorer bench.txt` returns zero matches. The 6 `BenchmarkDefaultScorer_*` benchmarks in `scorer_bench_test.go` run during `make bench` but their numbers are not in the baseline, so a 10x regression in `DefaultScorer.Score` would not fail CI.
- **Standard:** `.claude/skills/go-testing-standards/SKILL.md` line 90; `docs/requirements.md` §14.4
- **Action:** Code fix (regenerate bench.txt)
- **Rationale:** Identical to 08-TEST-ANALYSIS TEST-10. The benchmark gate that the spec demands is wired in the Makefile but the baseline doesn't include the latest surface.
- **Suggested fix:** Re-run `go test -bench=. -benchmem -count=10 ./... > bench.txt`; commit.

### [Critical] No DefaultScorer-level property tests: `Identity`, `Symmetric`, `Score = Σ Weight·ScoreAll`
- **File:** `scorer_test.go` (would-be tests)
- **Phase introduced:** Phase 8
- **Issue:** The Scorer-level property tests check `DeterministicAcrossRuns`, `WeightSumOne`, `ScoreInRange`, `NoNaN_NoInf` — but NOT the three load-bearing mathematical invariants:
  - `DefaultScorer().Score(x, x) ≈ 1.0` for non-empty x
  - `DefaultScorer().Score(a, b) == DefaultScorer().Score(b, a)` (Scorer composed of symmetric algorithms)
  - `Score(a, b) == Σ algorithm.Weight * ScoreAll(a, b)[algorithm.ID]` (the algebraic identity linking the two public surfaces)
- **Standard:** `.claude/skills/go-testing-standards/SKILL.md` §Property Tests — Scorer-level properties (identity, range, composite-bounded)
- **Action:** Code fix
- **Rationale:** Same as 08-TEST-ANALYSIS TEST-04 / TEST-05 / TEST-06. Without these, regressions in the pre-normalisation gate or the reduction loop would not be detected by property test.
- **Suggested fix:** Three new `TestProp_Scorer_*` functions per the suggested code in 08-TEST-ANALYSIS.

### [Critical] Stale property-test failure log committed at `bench.txt.new`
- **File:** `bench.txt.new`
- **Phase introduced:** Phase 6 (MongeElkan)
- **Issue:** `bench.txt.new` contains the failure trace of `TestProp_MongeElkanScore_AsymmetricWhenTokenCountAsymmetric` from an earlier run. The file appears committed (not in `.gitignore`). The test now PASSES on 20 consecutive runs, but the file's existence suggests the asymmetric-property generator can produce inputs where Tokenise yields unequal token counts but MongeElkan happens to be symmetric (e.g. exotic-Unicode-only inputs where Tokenise's count of "fields" via strings.Fields differs from Tokenise's actual output). The test uses `strings.Fields(a)` as an under-estimate, which means the premise can be wrongly thought to hold.
- **Standard:** `.claude/skills/go-testing-standards/SKILL.md` (property tests must not flake); `.claude/skills/algorithm-correctness-standards/SKILL.md` (mathematical invariants must hold deterministically)
- **Action:** Discuss-phase needed (decide whether to tighten generator or accept the asymmetric-direction property is too narrow)
- **Rationale:** This is a *real* generator pitfall — the conditional invariant `aTokens != bTokens && 0 < fwd < 1 → fwd != rev` is true mathematically when the premise's token count is computed the same way MongeElkan does. Using `strings.Fields` for the premise but project Tokenise inside the algorithm creates an inconsistency.
- **Suggested fix:** Either (a) delete `bench.txt.new` from version control, gate the property to ASCII inputs only via custom generator; or (b) compute token counts via `fuzzymatch.Tokenise(a, fuzzymatch.DefaultTokeniseOptions())` to match the algorithm's view.

---

### [Important] `*Runes` fuzz harnesses missing for 9 of 23 algorithms
- **File:** `levenshtein_fuzz_test.go`, `damerau_full_fuzz_test.go`, `damerau_osa_fuzz_test.go`, `hamming_fuzz_test.go`, `jaro_fuzz_test.go`, `jarowinkler_fuzz_test.go`, `lcsstr_fuzz_test.go`, `swg_fuzz_test.go`, `ratcliff_obershelp_fuzz_test.go`
- **Phase introduced:** Phases 2-6
- **Issue:** Cosine, SorensenDice, QGramJaccard, Tversky each have `Fuzz*Score` AND `Fuzz*ScoreRunes`. Levenshtein, DamerauLevenshteinOSA/Full, Hamming, Jaro, JaroWinkler, LCSStr, SmithWatermanGotoh, RatcliffObershelp have only the byte-path fuzz. Strcmp95 is byte-only by design (ASCII letters), so that one is OK.
- **Standard:** `.claude/skills/go-testing-standards/SKILL.md` §Fuzz Tests ("one Fuzz* per public function"); `.claude/skills/algorithm-correctness-standards/SKILL.md` (rune variants are public API)
- **Action:** Code fix
- **Rationale:** Rune variants traverse different code paths (extra `[]rune` allocation, rune-by-rune scanning, different invalid-UTF-8 fallback). Bugs in the rune path won't be caught by the byte-path fuzz.
- **Suggested fix:** Add `Fuzz<Algorithm>ScoreRunes` per the existing per-Score pattern.

### [Important] Distance-function fuzz harnesses absent
- **File:** `levenshtein_fuzz_test.go`, `damerau_full_fuzz_test.go`, `damerau_osa_fuzz_test.go`, `hamming_fuzz_test.go`, `lcsstr_fuzz_test.go`
- **Phase introduced:** Phase 2
- **Issue:** Public API: `LevenshteinDistance`, `LevenshteinDistanceRunes`, `DamerauLevenshteinFullDistance`, `DamerauLevenshteinFullDistanceRunes`, `DamerauLevenshteinOSADistance`, `DamerauLevenshteinOSADistanceRunes`, `HammingDistance`, `HammingDistanceRunes`, `LongestCommonSubstring`, `LongestCommonSubstringRunes`. None have fuzz harnesses. Score-fuzz exercises distance internally for most, but `HammingDistance`'s unequal-length policy and `LongestCommonSubstring`'s return type (string) are not exercised by score fuzz.
- **Standard:** `.claude/skills/go-testing-standards/SKILL.md` ("Every public function has a fuzz harness")
- **Action:** Code fix
- **Rationale:** Public API rule is non-functional-form-aware.
- **Suggested fix:** Add `FuzzLevenshteinDistance`, `FuzzHammingDistance`, `FuzzLongestCommonSubstring` etc. asserting panic-free and non-negative integer / valid UTF-8 substring.

### [Important] Phonetic-Code fuzz harnesses only test `Score`, not `Code/Keys`
- **File:** `soundex_fuzz_test.go`, `double_metaphone_fuzz_test.go`, `nysiis_fuzz_test.go`, `mra_fuzz_test.go`
- **Phase introduced:** Phase 4
- **Issue:** `FuzzSoundex`, `FuzzNYSIIS`, `FuzzMRA`, `FuzzDoubleMetaphone` exist but the on-disk corpus directory is named `FuzzSoundex` (etc.), and the test exercises `*Score`, not the underlying `SoundexCode`, `NYSIISCode`, `MRACode`, `MRACompare`, `DoubleMetaphoneKeys`. The Code functions have their own unit-test coverage but no fuzz harness asserts the output character-set or length invariants under random input.
- **Standard:** `.claude/skills/go-testing-standards/SKILL.md`; `docs/requirements.md` §15.4
- **Action:** Code fix
- **Rationale:** Property tests `TestProp_SoundexCode_Charset`, `TestProp_NYSIISCode_Charset`, `TestProp_MRACode_Charset`, `TestProp_DoubleMetaphone_KeyCharset` cover the charset invariant under random input but not under adversarial input (invalid UTF-8, embedded NUL, very long input). A fuzz harness would close this.
- **Suggested fix:** Either expand `FuzzSoundex` etc. to also call the Code variant or add separate `FuzzSoundexCode` harnesses.

### [Important] Parameterised `Scorer` options not separately fuzz-tested
- **File:** scorer_options_*test.go and fuzz files
- **Phase introduced:** Phase 8
- **Issue:** `WithQGramJaccardAlgorithm`, `WithSorensenDiceAlgorithm`, `WithCosineAlgorithm`, `WithTverskyAlgorithm`, `WithMongeElkanAlgorithm`, `WithSmithWatermanGotohAlgorithm` each have happy-path + invalid-weight + invalid-N + (sometimes) invalid-alpha/beta unit tests. None have a fuzz harness over (n int, alpha, beta float64, inner AlgoID).
- **Standard:** `.claude/skills/go-testing-standards/SKILL.md` (one Fuzz* per public function)
- **Action:** Code fix
- **Rationale:** Option-layer regressions on float-edge values (denormals, NaN, infinities, very-large alpha/beta) would not be caught.

### [Important] No BDD feature files for `normalisation`, `determinism`
- **File:** `tests/bdd/features/` (would-be `normalisation.feature`, `determinism.feature`)
- **Phase introduced:** Phase 1 (bootstrap) — Phase 5 (normalisation/tokenise)
- **Issue:** `docs/requirements.md` §15.6 enumerates the six required feature files: `algorithms.feature`, `scorer.feature`, `normalisation.feature`, `determinism.feature`, `scan.feature`, `suppression.feature`. Only the per-algorithm feature files and `scorer.feature` exist. `normalisation.feature` and `determinism.feature` are not present even though the underlying surfaces are implemented (Phase 5 Normalise / Tokenise; cross-platform golden in `testdata/golden/`).
- **Standard:** `.claude/skills/go-testing-standards/SKILL.md` §BDD Tests; `docs/requirements.md` §15.6
- **Action:** Code fix
- **Rationale:** BDD is the contract layer. The normalisation pipeline (lowercase × strip-separators × camelCase-split × NFC × strip-diacritics) is heavily branched and well-tested at unit level but the BDD coverage is missing.
- **Suggested fix:** Create `normalisation.feature` with one Scenario Outline per option combination per `docs/requirements.md` §9; create `determinism.feature` cross-referencing the golden-file behaviour. Tag with `@normalisation` / `@determinism`.

### [Important] Required BDD tags missing: `@character`, `@qgram`, `@token`, `@phonetic`, `@gestalt`, `@scan`, `@suppression`, `@normalisation`, `@determinism`
- **File:** `tests/bdd/features/*.feature`
- **Phase introduced:** Phases 2-8
- **Issue:** `.claude/skills/go-testing-standards/SKILL.md` §BDD Tests requires category tags. Actual tags present: `@scorer`, `@partial`, `@phonetic`, `@token`, `@monge`, `@soundex`, `@mra`, `@nysiis`, `@double`, `@custom`, `@default`, `@scoreall`, `@concurrency`, `@errors`, `@byte`, `@rune`, `@pitfall`. Missing the category roll-up tags (`@character` for Levenshtein/Hamming/etc., `@qgram` for Cosine/Jaccard, `@gestalt` for RatcliffObershelp, `@determinism`, `@normalisation`).
- **Standard:** `.claude/skills/go-testing-standards/SKILL.md` §BDD Tests — Tags for filtering
- **Action:** Code fix
- **Rationale:** Tag-based filtering (`go test -godog.tags=@phonetic`) is the standard way to slice the suite. Several categories cannot be selected today.
- **Suggested fix:** Apply category tags to existing feature files; one line per scenario.

### [Important] No tests for `ErrInvalidInput`, `ErrInvalidConfiguration`, `ErrEmptyInput`
- **File:** `errors_test.go`
- **Phase introduced:** Phase 1
- **Issue:** `errors_test.go` exercises 10 sentinels with `errors.Is`/`Error()`/prefix checks, but these three sentinels are NEVER returned by any public function currently — they're declared but unused (grep confirms). The test asserts the sentinels exist; it does not assert any code path returns them.
- **Standard:** `.claude/skills/algorithm-correctness-standards/SKILL.md` (errors-as-API must be testable from public surface)
- **Action:** Discuss-phase needed
- **Rationale:** Either remove the unused sentinels or add the public API surfaces that return them (e.g. an `Extract` API gated by `ErrInvalidInput`). Currently they're contract-declared but contract-not-enforced.

### [Important] `ErrHammingLengthMismatch` documented in `errors.go` comment but not declared
- **File:** `errors.go:31`
- **Phase introduced:** Phase 2
- **Issue:** Comment references "ErrHammingLengthMismatch for the Hamming algorithm" but no such sentinel exists. The implementation uses the silent-zero policy (`hamming.go:74-80`). The algorithm-correctness-standards skill (line 104) says "Length mismatch for Hamming: return 0.0 from HammingScore (return ErrHammingLengthMismatch from HammingDistance). Documented behaviour." — but the implementation diverges from the skill.
- **Standard:** `.claude/skills/algorithm-correctness-standards/SKILL.md` line 104
- **Action:** Skill clarification or code fix
- **Rationale:** Either update the skill to match the silent-zero LOCKED policy (which is what the BDD feature file and unit tests pin) or implement the documented error variant. Currently the documentation in `.claude/skills/` and the implementation contradict.
- **Suggested fix:** Update the skill to reflect the silent-zero policy and document it as the LOCKED choice; remove the stale comment in `errors.go:31`.

### [Important] `ScoreAll` map key type — skill documents `string` keys (AlgoID.String()), code returns `map[AlgoID]float64`
- **File:** `scorer.go:497`; `.claude/skills/go-testing-standards/SKILL.md` line 34
- **Phase introduced:** Phase 8
- **Issue:** Skill: "`ScoreAll` returns per-algorithm scores keyed by `AlgoID.String()`". Code: returns `map[AlgoID]float64` (typed enum keys). The deviation is documented in `scorer.go:470` as a spec override.
- **Standard:** `.claude/skills/go-testing-standards/SKILL.md` §Scorer-level properties
- **Action:** Skill clarification
- **Rationale:** Implementation made a deliberate API choice (compile-time key safety) that the skill text doesn't reflect. Don't change the code; update the skill to match the override.

### [Important] No `ExampleNewScorer`, `ExampleDefaultScorer`, `ExampleScorer_Score`, `ExampleScorer_Match` runnable examples
- **File:** `example_test.go`
- **Phase introduced:** Phase 8
- **Issue:** 39 `Example*` functions exist — one per algorithm — but none for the Scorer surface. `docs/requirements.md` §16.5 specifies one Example per public Score-producing function; the Scorer's three public methods get no godoc-runnable demonstration.
- **Standard:** `.claude/skills/documentation-standards/SKILL.md`; `docs/requirements.md` §16.5
- **Action:** Code fix
- **Rationale:** pkg.go.dev consumers reading the `Scorer` doc get no executable example.
- **Suggested fix:** Add `ExampleDefaultScorer`, `ExampleNewScorer`, `ExampleScorer_Score`, `ExampleScorer_Match`, `ExampleScorer_ScoreAll` mirroring the per-algorithm pattern.

### [Important] Hardcoded `len(want) != 6` in `TestScorer_ConcurrentSafety`
- **File:** `scorer_test.go:941`
- **Phase introduced:** Phase 8
- **Issue:** Assertion `if len(want) != 6` couples the test to today's DefaultScorer composition. Any future change to DefaultScorer composition (e.g. add a 7th algorithm, drop one) silently breaks this assertion without explaining itself.
- **Standard:** general test-hygiene
- **Action:** Code fix
- **Rationale:** A spec change in `docs/requirements.md` §8.5 must be reflected here.
- **Suggested fix:** `want := results[0]; if len(want) != len(s.Algorithms()) { ... }` — derive from the Scorer itself.

### [Important] No "Very long input (1000+ chars)" tests for character-based algorithms (except Jaro)
- **File:** `levenshtein_test.go`, `damerau_*_test.go`, `hamming_test.go`, `jarowinkler_test.go`, `lcsstr_test.go`, `swg_test.go`, `strcmp95_test.go`, `ratcliff_obershelp_test.go`
- **Phase introduced:** Phases 2-6
- **Issue:** Only `jaro_test.go:198-206` exercises 300+ char input in a unit test. Other DP algorithms have bench tests at 500 chars but no functional test at 1000+ chars to assert the heap-path (large input → heap-allocated DP table) produces the right answer.
- **Standard:** prompt's per-algorithm checklist "Very long input (1000+ chars)"
- **Action:** Code fix
- **Rationale:** Stack-vs-heap path divergence is a documented performance optimisation; testing only short inputs leaves the heap path untested for correctness.

### [Important] No `quick.Check` `MaxCount` raised above default 100 for cheap properties
- **File:** `props_test.go` (entire file uses `nil` config)
- **Phase introduced:** Phases 2-8
- **Issue:** Almost every `quick.Check(f, nil)` call uses the default 100 iterations. For algorithms with ~10-50 µs per iter, 1000 or 10000 iterations would tighten confidence at trivial cost. Only `tokenise_test.go:506` raises to 200 and `props_test.go:491-497` uses a custom generator (Damerau-OSA constrained).
- **Standard:** `.claude/skills/go-testing-standards/SKILL.md` Property Tests
- **Action:** Code fix
- **Rationale:** Same as 08-TEST-ANALYSIS TEST-08 (scoped) — extends to whole codebase. The flake-once-per-100-runs surface area is unnecessarily wide.
- **Suggested fix:** Add a package-level `var quickCfg = &quick.Config{MaxCount: 1000}` and pass it consistently.

### [Important] Property-test generators rely on `quick.Check`'s default random `string` generator — produces predominantly long non-ASCII strings
- **File:** `props_test.go`
- **Phase introduced:** Phase 2
- **Issue:** Go's `testing/quick` default `string` generator produces random runes drawn from the full Unicode space. This is generally good (exercises non-ASCII paths) but means:
  - Property tests rarely exercise short-ASCII edge cases.
  - Inputs are predominantly multi-byte UTF-8 with surrogate-pair-style runes.
  - Some properties (e.g. SWG monotonicity, MRA threshold-monotonicity) may not generate enough realistic-shaped inputs.
- **Standard:** `.claude/skills/go-testing-standards/SKILL.md` Property Tests
- **Action:** Code fix or Discuss-phase needed
- **Rationale:** A targeted generator (mixed-ASCII / mixed-Unicode / mixed-length) would surface more bugs per iteration. The custom `randShortASCII` and `asciiAlpha` patterns already exist; expanding their use across `props_test.go` would help.

### [Important] No property test verifying `Algorithms()` ordering is by AlgoID ascending across permutations
- **File:** `scorer_test.go:353-397`
- **Phase introduced:** Phase 8
- **Issue:** `TestScorer_Algorithms_SortedAscending` exercises a single permutation (3 algorithms added in scrambled order). A property test that draws an arbitrary permutation of N algorithms and asserts the output is sorted would catch any future sort-key change.
- **Standard:** `.claude/skills/determinism-standards/SKILL.md` (sort-key completeness)
- **Action:** Code fix
- **Rationale:** Determinism guarantee on output ordering deserves a property gate.

---

### [Improvement] Test naming inconsistency — `Test<Algo>_Behaviour` vs `Test<AlgoScore>_Behaviour` vs `Test<Algo>Score_Behaviour`
- **File:** All `*_test.go`
- **Phase introduced:** Phases 2-8
- **Issue:** Patterns observed:
  - `TestLevenshtein_BothEmpty` (algorithm-name)
  - `TestStrcmp95Score_ZeroAllocs_ASCII_Short` (algorithm+Score-name)
  - `TestProp_LevenshteinScore_RangeBounds` (algorithm+Score-name)
  - `TestSoundex_BothEmpty` and `TestSoundexScore_NonMatch` (mixed in same file)
  - `TestDispatch_LevenshteinRegistered` (subject-first)
- **Standard:** `.claude/skills/go-testing-standards/SKILL.md` (no explicit pattern)
- **Action:** Skill clarification
- **Rationale:** Not blocking. But a project-wide convention (e.g. always `Test<PublicFn>_<Behaviour>`) would reduce cognitive load.
- **Suggested fix:** Establish convention in skill; do not rename existing tests retroactively.

### [Improvement] Benchmark naming inconsistency — `BenchmarkSoundexScore_ASCII_Identity` vs `BenchmarkSoundexScore_ASCII_Short` vs `BenchmarkMRAScore_Match`
- **File:** `*_bench_test.go`
- **Phase introduced:** Phases 2-8
- **Issue:** No consistent suffix convention. Most algorithms use `ASCII_Short / ASCII_Medium / ASCII_Long / Unicode_Short`, but phonetic algorithms use `Match / NoMatch / Identity` and Tokenise uses `DefaultOptions` / `StripDiacritics_Short`. Bench naming consistency affects `benchstat` output legibility.
- **Standard:** none codified
- **Action:** Skill clarification
- **Rationale:** Polish item; doesn't change correctness.

### [Improvement] Phonetic property tests lack `NoNegativeZero`
- **File:** `props_test.go` (Soundex, DoubleMetaphone, NYSIIS, MRA sections)
- **Phase introduced:** Phase 4
- **Issue:** Character-based and q-gram algorithms uniformly have `TestProp_<Algo>Score_NoNegativeZero`. Phonetic algorithms do not. Defensible (phonetic scores are 0.0 or 1.0 literals, no arithmetic), but the property is cheap to add for symmetry with the rest of the suite.
- **Standard:** `.claude/skills/determinism-standards/SKILL.md` (-0.0 handling)
- **Action:** Code fix or Skill clarification
- **Rationale:** Defensive completeness only.

### [Improvement] `TestScorer_ConcurrentSafety` runs once, not under `-count=N`
- **File:** `scorer_test.go:903`
- **Phase introduced:** Phase 8
- **Issue:** 100 goroutines × 3 methods × single test run. Under `-race -count=10` the test would surface more goroutine schedules.
- **Standard:** general concurrency-test hygiene
- **Action:** Skill clarification
- **Rationale:** Same as 08-TEST-ANALYSIS TEST-09.

### [Improvement] Match-benchmark sink gate is dead-code in `scorer_bench_test.go`
- **File:** `scorer_bench_test.go:171-173`
- **Phase introduced:** Phase 8
- **Issue:** `if sink < -1` on a non-negative counter is provably false; sufficiently aggressive optimisers could elide it. Same as 08-TEST-ANALYSIS TEST-12.
- **Standard:** locked benchmark pattern (PATTERNS.md)
- **Action:** Code fix
- **Rationale:** One-line cosmetic. Doesn't affect today's compiler but defeats the purpose of the locked sink-gate pattern.

### [Improvement] Golden file `scorer-default.json` has no single-character-pair entry
- **File:** `testdata/golden/scorer-default.json`
- **Phase introduced:** Phase 8
- **Issue:** 22 entries × 5 configs but no `"a" / "b"` row. Single-character is a documented edge case per `.claude/skills/go-testing-standards/SKILL.md`.
- **Standard:** as above
- **Action:** Code fix
- **Rationale:** Same as 08-TEST-ANALYSIS TEST-13.

### [Improvement] Golden file `0.9999999999999999` ULP-below-1.0 identity composite not documented
- **File:** `scorer_golden_test.go:148`
- **Phase introduced:** Phase 8
- **Issue:** The composite of 6 equal-weighted algorithms each returning 1.0 produces `0.9999999999999999` (one ULP below) because `1.0/6.0` is irrational in binary float. The golden file pins this correctly, but no comment explains it — reviewers might misread as a bug.
- **Standard:** general test-readability
- **Action:** Code fix
- **Rationale:** Same as 08-TEST-ANALYSIS TEST-14.

### [Improvement] Internal-test usage limited to scorer / q_gram / token_indel
- **File:** `scorer_internal_test.go`, `scorer_options_internal_test.go`
- **Phase introduced:** Phase 8
- **Issue:** Only 2 of the 23 algorithm packages have an `_internal_test.go`. q_gram_test.go and token_indel_test.go are in `fuzzymatch_test` (external) per the file headers but reference unexported names — actually they're internal-only via `_test.go` location. Strcmp95's similar-character table has `Strcmp95SimilarCharsEntryForTest` exposed via `export_test.go`; an internal test would be cleaner.
- **Standard:** `.claude/skills/go-testing-standards/SKILL.md` §Unit Tests (internal tests "where genuinely needed")
- **Action:** Code fix or Discuss-phase needed
- **Rationale:** `export_test.go` introduces test-only public symbols (`DispatchInvokeForTest`, `Strcmp95SimilarCharsEntryForTest`) that pollute the godoc surface; an internal test file avoids that pollution.
- **Suggested fix:** Migrate Strcmp95 similar-char table tests to a `strcmp95_internal_test.go` and remove the export shim.

### [Improvement] Fuzz corpus on disk lags fuzz harness functions (11 fuzzers have no committed seed corpus)
- **File:** `testdata/fuzz/`
- **Phase introduced:** Phases 2-6
- **Issue:** 31 Fuzz* functions in source; 20 corpus directories on disk. Fuzz harnesses without committed corpora can only seed from `f.Add()` calls in source — fine for first-run, but loses the regression value of CI-discovered failure inputs across builds.
- **Standard:** `.claude/skills/go-testing-standards/SKILL.md` §Fuzz Tests (corpus checked into testdata/fuzz/)
- **Action:** Code fix (run the nightly fuzz pipeline once, commit corpus)
- **Rationale:** Hygiene. Corpora persist crash regressions across PRs.

### [Improvement] Cross-validation vectors absent for character-based algorithms
- **File:** `testdata/cross-validation/`
- **Phase introduced:** Phases 2-3
- **Issue:** Cross-validation vectors exist for phonetic, swg, token-ratios, ratcliff-obershelp. Character-based algorithms (Levenshtein, Damerau-*, Hamming, Jaro, JaroWinkler, Strcmp95, LCSStr) have only inline unit-test reference vectors with no JSON-vector cross-validation set. The primary-source-citation-correctness reviews can sign off without this, but cross-validation against (e.g.) C++ libraries or `python-Levenshtein` adds independent confirmation.
- **Standard:** `.claude/skills/algorithm-correctness-standards/SKILL.md` §Reference Vectors
- **Action:** Discuss-phase needed
- **Rationale:** Not strictly required by the skill but improves confidence at v1.0.

### [Improvement] No coverage gate in CI workflow
- **File:** `.github/workflows/*.yml`
- **Phase introduced:** Phase 1
- **Issue:** The missing `internal_coverage_test.go` would catch the floor failure on every PR; without it, CI silently runs `go test -cover` but doesn't fail on the 90.4% / 95% gap. Confirms with the [Critical] meta-test finding.
- **Standard:** `.claude/skills/go-testing-standards/SKILL.md`; `docs/requirements.md` §15.8
- **Action:** Code fix (CI + meta-test)
- **Rationale:** Detection lag — the headline coverage gap has likely persisted across multiple PRs without notice.

### [Improvement] `WithSmithWatermanGotohAlgorithm` parameter-validation tests thin (only `_HappyPath`, `_CapturesParams`, `_InvalidWeight`)
- **File:** `scorer_options_test.go:506-526`
- **Phase introduced:** Phase 8
- **Issue:** `NewSWGParams()` has multiple fields (Match, Mismatch, Gap, etc. per `swg.go`). Only one parameter combination is exercised; no tests for negative `Match`, zero `Gap`, or NaN/Inf in any field.
- **Standard:** `.claude/skills/go-testing-standards/SKILL.md`
- **Action:** Code fix

### [Improvement] No bench for `WithoutNormalisation` Scorer vs `WithNormalisation` Scorer (the allocation-comparison gate)
- **File:** `scorer_bench_test.go`
- **Phase introduced:** Phase 8
- **Issue:** Same as 08-TEST-ANALYSIS TEST-11. Without the comparator, the WR-04 allocation-reduction work has no visibility.
- **Standard:** `.claude/skills/performance-standards.md`
- **Action:** Code fix

---

## Per-Concern Status Summary

### Per-Algorithm Catalogue
- **COVERED** — All 23 algorithms have unit + property + at least byte-path fuzz + benchmark + BDD + cross-platform golden.
- **PARTIALLY COVERED** — 9 algorithms missing `*Runes` fuzz; 5 algorithms missing Distance fuzz; phonetic algorithms missing `Code/Keys` fuzz; 8 algorithms missing very-long-input unit tests.

### Per-Scorer Concern
- **COVERED** — Every `With*` option happy-path + invalid-weight + invalid-N tested; Score / ScoreAll / Match / Threshold / Algorithms / DefaultScorer / DefaultScorerOptions all exercised; concurrent safety verified; deterministic-across-runs property tested; golden file with 22×5 configs.
- **MISSING** — Scorer-level fuzz; DefaultScorer Identity / Symmetry / `Score = Σ Weight·ScoreAll` properties; NaN-threshold gate; NaN/Inf-weight gate; Tversky α=β=0 option-layer gate; `Scorer.Score` examples; coverage shortfall on `DefaultScorer` (75%) panic-line.

### Per-Scan Concern (Phase 6 — pre-implementation)
- **MISSING** — Entire scan sub-package is unimplemented. All scan-checklist items (within-group, cross-group, suppression composition, sort-key determinism, performance budgets, BDD scenarios) are deferred until Phase 6 lands.

### Per-Normalisation Concern
- **COVERED** — 8 `TestNormalise_*` unit tests + 3 property tests + golden file (`normalisation.json`); ASCII fast path documented and benchmarked.
- **MISSING** — `tests/bdd/features/normalisation.feature` does not exist.

### Per-Tokenisation Concern
- **COVERED** — 13 `TestTokenise_*` unit tests + 6 property tests; `DefaultTokeniseOptions` honoured.
- **PARTIALLY COVERED** — Bench coverage exists but no rune-vs-byte ASCII-fast-path side-by-side.

### Coverage Thresholds
- **Overall:** 90.4% (target 95%) — **FAILS**
- **Per-file:** `double_metaphone.go` 83.8% (target 90%) — **FAILS**; all others ≥ 90%
- **Public API:** ~99% (only the unreachable `panic` line in `DefaultScorer` and a few defensive Step-checks in `NewScorer` are uncovered)

### Property-test depth
- **COVERED** for: Range, Identity, Symmetry, Triangle (where applicable), NoNaN, NoInf, NoNegativeZero (except phonetic).
- **MISSING** for: Scorer Identity / Symmetry / Score=ΣWeight·ScoreAll; `Algorithms()` sort property over permutations.
- **AT RISK** (generator flake): `TestProp_MongeElkanScore_AsymmetricWhenTokenCountAsymmetric` premise inconsistent with algorithm's tokeniser.

### Fuzz harness coverage
- **COVERED** for: 23 algorithm Score functions (byte path), Normalise, Tokenise, 4 Q-gram rune variants, MongeElkan asymmetric + symmetric.
- **MISSING** for: 9 character-based `*Runes` Score variants; all Distance functions; LongestCommonSubstring (returns string); phonetic Code/Keys variants; Scorer.Score / ScoreAll / Match; NewScorer with random options.

### BDD scenario coverage
- **COVERED** for: every algorithm; scorer composition (12 scenarios across 12 mandatory classes).
- **MISSING** for: `normalisation.feature`, `determinism.feature`, `scan.feature`, `suppression.feature`, `algorithms.feature` (single combined alternative — currently per-algo files instead).
- **MISSING tags:** `@character`, `@qgram`, `@gestalt`, `@determinism`, `@normalisation`.

### Cross-validation evidence persistence
- **COVERED** for: phonetic (4 algorithms), swg, token-ratios (3 algorithms), ratcliff-obershelp.
- **MISSING** for: 9 character-based algorithms (Levenshtein, Damerau-*, Hamming, Jaro, JaroWinkler, Strcmp95, LCSStr) and 4 q-gram algorithms.

### Edge-case sweep
- **Both-empty / one-empty / identical:** COVERED across all algorithms.
- **Unicode (multi-byte UTF-8, CJK, emoji):** COVERED for algorithms with `*Runes` variants.
- **Malformed UTF-8:** COVERED via fuzz harnesses (byte path) for algorithms with fuzz; MISSING on Scorer surface.
- **Very-long input (1000+):** PARTIALLY COVERED — Jaro has it, Cosine has a 1000-iteration determinism test, others rely on benchmarks-as-tests.
- **Length-mismatch (Hamming-specific):** COVERED with silent-zero policy.

### Concurrent-safety tests
- **COVERED** — `TestScorer_ConcurrentSafety` runs 100 goroutines × 3 methods under `-race`. Race tests pass (~352s with `-race`).

### Bench naming consistency
- **PARTIALLY COVERED** — Most benchmarks follow `Benchmark<Algo><Method>_<Path>_<Length>`. Phonetic, MRA-pathological, partial-ratio adversarial benches use bespoke suffixes.

### `quick.MaxCount` settings
- **DEFAULTS USED EVERYWHERE** except `tokenise_test.go:506` (MaxCount=200) and one DamerauOSA constrained config. Property-test confidence is thinner than it could be.

### Generator pitfalls
- **TestProp_MongeElkanScore_AsymmetricWhenTokenCountAsymmetric:** premise based on `strings.Fields` differs from algorithm's `Tokenise`. Stale failure in `bench.txt.new`.
- **TestProp_Scorer_WeightSumOne:** `uint16` overflow — `u==65535 → u+1==0`. Per-run flake probability ~0.46%.

### Test file organisation
- **Mostly external** (`package fuzzymatch_test`). Internal test files: `scorer_internal_test.go`, `scorer_options_internal_test.go`, `q_gram_test.go` (uses external pkg via export_test shim), `token_indel_test.go` (same).
- **`export_test.go` shim pattern** used to expose dispatch/internal state — slight pollution of godoc surface.

### Flaky-test patterns
- **WR-01 `uint16` overflow** still open (`scorer_test.go:827`).
- **MongeElkan asymmetry-conditional property** has documented flake risk on exotic-Unicode inputs (`bench.txt.new` artefact).
- **`TestScorer_ConcurrentSafety`** runs once; goroutine-schedule-dependent bugs would surface only intermittently under `-count`.

---

## Recommendation

**Coverage gaps to address before milestone release.**

The test suite is large (~480 Test functions, 31 fuzzers, 127
benchmarks, 169 BDD scenarios) and structurally sound for the 23
algorithm catalogue. Every algorithm clears the per-algorithm
minimum: unit tests with reference vectors, property tests for the
mathematical invariants, fuzz harness for byte-path Score, benchmark,
BDD scenario, cross-platform golden. The Scorer surface gets
exhaustive option-layer testing.

The headline gaps are:

1. **Three meta-tests missing** (`internal_coverage_test.go`,
   `documentation_test.go`, `readme_shop_front_test.go`) — without
   `internal_coverage_test.go` the project's own coverage-floor gate
   is unenforced, which is why the 90.4% overall shortfall has
   persisted.
2. **Four still-open Phase-8 defects** — `WithThreshold(NaN)`,
   `WithAlgorithm(_, NaN/Inf)`, `WithTverskyAlgorithm(0, 0)`, and the
   `uint16` overflow flake — all small fixes already identified by
   08-REVIEW but not yet landed.
3. **Scorer fuzz harness missing** entirely — the single most-used
   public surface has no fuzz gate.
4. **Three Scorer-level property tests missing** — Identity, Symmetry,
   `Score = Σ Weight·ScoreAll`.
5. **`double_metaphone.go` at 83.8%** drags overall coverage below the
   95% floor — needs language-branch reference vectors.
6. **Scorer benchmarks not in `bench.txt`** — regression gate inactive.
7. **9 character-based `*Runes` and all Distance-function fuzz
   harnesses absent** — rune-path bugs would slip past CI.
8. **Stale `bench.txt.new`** artefact — investigate the MongeElkan
   asymmetric-property generator and either tighten it or delete the
   committed failure log.

**Not ready for milestone release** until items 1-6 land. Items 7-8
are important hardening work that can run in parallel and should land
before the v1.0.0 cut.

After those items land — particularly `internal_coverage_test.go`,
the Scorer fuzz harnesses, the four option-layer bug fixes, and the
coverage shortfall — the test suite meets the standards and the
milestone-release path is clear.

_Analysed: 2026-05-17_
_Analyst: test-analyst (whole-codebase scope, phases 1-8)_
