---
phase: 03-smith-waterman-gotoh
plan: 03
type: execute
wave: 3
depends_on:
  - 03-01-swg-implementation
  - 03-02-swg-cross-validation
files_modified:
  - algorithms_golden_test.go
  - testdata/golden/algorithms.json
  - cross_algorithm_consistency_test.go
  - examples/identifier-similarity/main.go
  - examples/identifier-similarity/main_test.go
  - bench.txt
  - llms.txt
  - docs/requirements.md
autonomous: true
requirements:
  - CHAR-08
tags: [smith-waterman-gotoh, swg, finalisation, algorithms-json-merge, cross-algorithm-consistency, identifier-similarity-example, bench-update, llms-sync, requirements-doc-raw-surface-update]

must_haves:
  truths:
    # Goal-backward truths
    - "testdata/golden/algorithms.json contains entries from ALL seven algorithms (Levenshtein, Hamming, Jaro, JaroWinkler, DamerauLevenshteinOSA, DamerauLevenshteinFull, SmithWatermanGotoh) merged from the seven per-algorithm _staging/*.json files via TestGolden_Algorithms_Merge"
    - "TestGolden_Algorithms_Merge's stagingFiles slice contains `_staging/swg.json` alongside the six Phase 2 staging files; the merge produces a byte-stable algorithms.json (re-running without `-update` shows zero diff)"
    - "cross_algorithm_consistency_test.go contains a new TestCrossAlgorithm_SWG_Levenshtein_SubstringDivergence test that asserts SmithWatermanGotohScore(\"http_request\", \"http_request_header_fields\") STRICTLY > LevenshteinScore(\"http_request\", \"http_request_header_fields\") — the load-bearing local-vs-global-alignment claim pinning that SWG finds the substring while Levenshtein counts every uncovered position as an edit"
    - "The existing TestCrossAlgorithm_IdentityConvergence, TestCrossAlgorithm_BothEmptyConvergence, and TestCrossAlgorithm_OneEmpty_ScoreAgreement tests EITHER include SWG in their funcs slice (so SWG identity / both-empty / one-empty convergence is pinned alongside the Phase 2 six) OR a new TestCrossAlgorithm_SWG_<topic> test is added for each of those three — choose the funcs-slice-extension form (one-line edit) per the cross_algorithm_consistency_test.go convention"
    - "examples/identifier-similarity/main.go's algorithms slice has a new {\"SWG\", fuzzymatch.SmithWatermanGotohScore} entry; the table grows from 7-row × 6-algorithm-column to 7-row × 7-algorithm-column"
    - "examples/identifier-similarity/main_test.go's `want` constant is regenerated to include the SWG column; TestExample_Output (the byte-for-byte line-by-line diff per IN-04) passes; TestExample_ColumnWidths passes (self-adapts to the new column count)"
    - "bench.txt is regenerated via `make bench` after Wave 2 lands; the new file contains the existing Phase 2 benchmark rows PLUS six new BenchmarkSmithWatermanGotoh* series (ASCII_Short, ASCII_Medium, ASCII_Long, Unicode_Short, WithParams_ASCII_Short, RawScore_ASCII_Short)"
    - "llms.txt has a new `### Smith-Waterman-Gotoh local-alignment similarity` section listing all 8 new exported symbols: type SWGParams, NewSWGParams, SmithWatermanGotohScore, *Runes, *WithParams, RawScore, RawScoreRunes, RawScoreWithParams"
    - "ai_friendly_test.go's TestAIFriendly_AllExportedSymbolsListed (or equivalent meta-test) passes — every new exported SWG symbol appears in llms.txt"
    - "docs/requirements.md §7.1.8 is updated to list ALL 6 SWG public functions (3 normalised + 3 Raw) plus the SWGParams type and NewSWGParams constructor — recording the Raw* surface expansion that the project owner approved on 2026-05-14 per CONTEXT.md §4"
    - "`make verify-determinism` exits 0 — the merged algorithms.json is byte-identical across the linux/amd64, linux/arm64, darwin/arm64, windows/amd64 CI matrix (DET-02 satisfied for SWG)"
    - "`make check` exits 0 (full quality gate green for the entire Phase 3 surface)"
    - "Phase 3 ROADMAP success criteria #1-#4 are demonstrably satisfied: (#1) SWG matches biopython byte-identically on the documented reference set including the one-long-gap canary [plan 03-02 closes this]; (#2) swg.go's file-level godoc cites Gotoh 1982 AND Flouri 2015 with the erratum + correction documented inline [plan 03-01 closes this]; (#3) Configurable affine gap penalty via SWGParams; property tests for identity, range, non-negativity invariants [plan 03-01 closes this]; (#4) Allocation budget enforced via benchmark; two-row DP variant; cross-platform golden file entry added; BDD scenario covers the canonical long-gap reference case [this plan + plan 03-01 close]"
    # Cross-cutting truths
    - "Apache-2.0 header preserved on every modified .go file; scripts/verify-license-headers.sh exits 0"
    - "Zero non-stdlib runtime require lines in root go.mod (verify-no-runtime-deps.sh exits 0)"
    - "All seven _staging/*.json files (six Phase 2 + the new swg.json) are PRESERVED in the repo (they document the per-algorithm contribution audit trail; future phases follow the same pattern)"
    - "No regression on the six existing Phase 2 benchmarks per `make bench-compare` (or, if benchstat reports a > 10% regression, the cause is investigated and either resolved or documented in 03-03-SUMMARY.md with rationale)"
    - "examples/identifier-similarity/main.go retains the IN-04 / W-2 documentation supersession comment (Hamming 0.0000 vs ERR — already present from Phase 2 plan 02-07; verify it still appears after the column addition)"
    - "The package-level llms.txt `## Algorithms` table (if such a table exists) is also updated to flip SWG's status from \"pending\" to \"shipped\" — confirm by reading the existing llms.txt structure during execution"
  artifacts:
    - path: "algorithms_golden_test.go"
      provides: "Modified TestGolden_Algorithms_Merge with _staging/swg.json added to stagingFiles slice (one-line edit); no other Phase 2 logic changes"
      contains: "_staging/swg.json"
    - path: "testdata/golden/algorithms.json"
      provides: "REGENERATED canonical golden file containing entries from all seven algorithms (Phase 2 six + SWG); byte-stable across re-runs and across the 5-platform CI matrix"
      contains: "SmithWatermanGotoh_one_long_gap_canary"
    - path: "cross_algorithm_consistency_test.go"
      provides: "Appended TestCrossAlgorithm_SWG_Levenshtein_SubstringDivergence + SWG-extended funcs slices in existing convergence tests"
      contains: "TestCrossAlgorithm_SWG_Levenshtein_SubstringDivergence"
    - path: "examples/identifier-similarity/main.go"
      provides: "Modified algorithms slice with new {\"SWG\", fuzzymatch.SmithWatermanGotohScore} entry"
      contains: "fuzzymatch.SmithWatermanGotohScore"
    - path: "examples/identifier-similarity/main_test.go"
      provides: "Regenerated want constant including the SWG column; column-widths and line-by-line output assertions pass"
      contains: "SWG"
    - path: "bench.txt"
      provides: "Regenerated benchstat baseline including six new SmithWatermanGotoh* benchmark series alongside the six Phase 2 series"
      contains: "BenchmarkSmithWatermanGotohScore"
    - path: "llms.txt"
      provides: "Appended `### Smith-Waterman-Gotoh local-alignment similarity` section listing all 8 new exported symbols"
      contains: "SmithWatermanGotohScore"
    - path: "docs/requirements.md"
      provides: "Updated §7.1.8 listing all 6 SWG public functions (3 normalised + 3 Raw) + SWGParams + NewSWGParams; records the Raw* surface expansion approved per CONTEXT.md §4"
      contains: "SmithWatermanGotohRawScore"
  key_links:
    - from: "testdata/golden/algorithms.json"
      to: "testdata/golden/_staging/{damerau_full,damerau_osa,hamming,jaro,jarowinkler,levenshtein,swg}.json"
      via: "TestGolden_Algorithms_Merge reads all seven staging files, sorts entries by Name, marshals via CanonicalMarshalForTest"
      pattern: "TestGolden_Algorithms_Merge"
    - from: "examples/identifier-similarity/main.go"
      to: "github.com/axonops/fuzzymatch"
      via: "import + 7 score function calls (now including fuzzymatch.SmithWatermanGotohScore)"
      pattern: "fuzzymatch\\.SmithWatermanGotohScore"
    - from: "examples/identifier-similarity/main_test.go TestExample_Output"
      to: "examples/identifier-similarity/main.go stdout"
      via: "os/exec captures stdout from `go run .`; line-by-line diff against the `want` constant (IN-04 pattern)"
      pattern: "TestExample_Output"
    - from: "cross_algorithm_consistency_test.go TestCrossAlgorithm_SWG_Levenshtein_SubstringDivergence"
      to: "swg.go + levenshtein.go (cross-references both)"
      via: "fuzzymatch.SmithWatermanGotohScore vs fuzzymatch.LevenshteinScore on a substring-containment input"
      pattern: "fuzzymatch\\.SmithWatermanGotohScore.*fuzzymatch\\.LevenshteinScore"
    - from: "llms.txt"
      to: "swg.go exported symbols"
      via: "ai_friendly_test.go meta-test parses go/ast and asserts every exported symbol is listed"
      pattern: "SmithWatermanGotohScore"

user_setup: []
---

<objective>
Finalise Phase 3 by integrating the SWG implementation into the project-wide infrastructure that was established in Phase 2 — the merged algorithms.json golden file, the cross-algorithm consistency test suite, the identifier-similarity example, the bench.txt benchstat baseline, and the project-wide documentation surfaces (llms.txt + docs/requirements.md §7.1.8). This plan owns four concerns that REQUIRE plans 03-01 and 03-02 to have merged:

(1) **Golden file merge**: extend the `stagingFiles` slice in TestGolden_Algorithms_Merge from six entries to seven by adding `_staging/swg.json`; regenerate testdata/golden/algorithms.json; verify byte-stable cross-platform.

(2) **Cross-algorithm consistency**: pin the load-bearing local-vs-global-alignment claim by adding TestCrossAlgorithm_SWG_Levenshtein_SubstringDivergence — SWG on `http_request`/`http_request_header_fields` scores STRICTLY higher than Levenshtein on the same pair (SWG finds the substring; Levenshtein counts every uncovered position as an edit). Extend the existing identity / both-empty / one-empty convergence tests to include SWG in their funcs slices.

(3) **Identifier-similarity example column**: add an `SWG` algorithm column to the existing 7-row × 6-column table; regenerate the `want` constant; the existing TestExample_Output and TestExample_ColumnWidths self-adapt or are updated; the IN-04 line-by-line diff pattern catches drift.

(4) **bench.txt + llms.txt + docs/requirements.md §7.1.8**: regenerate bench.txt with the new six SmithWatermanGotoh* benchmark series; append the SWG section to llms.txt listing all 8 new exported symbols; update docs/requirements.md §7.1.8 to record the Raw* surface expansion per CONTEXT.md §4 (the three Raw* functions beyond the original 3-function spec — this is a deliberate scope expansion approved by the project owner on 2026-05-14).

Purpose: this plan closes Phase 3. Plans 03-01 (algorithm + tests + staging golden) and 03-02 (biopython cross-validation gate) provide the algorithm; this plan integrates it into the project-wide surfaces so downstream phases (Phase 4 onwards) inherit SWG via the same staging-merge / cross-algorithm / example / bench / llms.txt patterns used for the Phase 2 six.

Output: Phase 3 is shippable. The Phase 3 ROADMAP success criteria #1-#4 are demonstrably satisfied across the four artefact families (golden file, cross-algorithm test, example program + meta-test, bench.txt + llms.txt + requirements doc).
</objective>

<execution_context>
@$HOME/.claude/get-shit-done/workflows/execute-plan.md
@$HOME/.claude/get-shit-done/templates/summary.md
</execution_context>

<context>
@.planning/PROJECT.md
@.planning/ROADMAP.md
@.planning/REQUIREMENTS.md
@.planning/STATE.md
@CLAUDE.md
@.planning/phases/03-smith-waterman-gotoh/03-CONTEXT.md
@.planning/phases/03-smith-waterman-gotoh/03-RESEARCH.md
@.planning/phases/03-smith-waterman-gotoh/03-PATTERNS.md
@.planning/phases/03-smith-waterman-gotoh/03-VALIDATION.md
@.planning/phases/03-smith-waterman-gotoh/03-01-swg-implementation-PLAN.md
@.planning/phases/03-smith-waterman-gotoh/03-02-swg-cross-validation-PLAN.md
@.planning/phases/02-core-character-algorithms-six/02-07-finalisation-SUMMARY.md
@docs/requirements.md
@.claude/skills/algorithm-correctness-standards/SKILL.md
@.claude/skills/determinism-standards/SKILL.md
@.claude/skills/go-testing-standards/SKILL.md
@.claude/skills/documentation-standards/SKILL.md

<interfaces>
This plan depends on plans 03-01 and 03-02:

From plan 03-01 (algorithm + staging golden + props/example/algoid extensions + BDD):
  - swg.go: public surface (SWGParams, NewSWGParams, 6 functions); kernel; godoc citing all three primary references
  - dispatch_swg.go: dispatch[AlgoSmithWatermanGotoh] registered
  - testdata/golden/_staging/swg.json: 6 entries sorted alphabetically; canonical byte form
  - algorithms_golden_test.go: buildSWGStagingEntries + TestGolden_SmithWatermanGotoh_Staging exist (THIS plan only extends the stagingFiles slice — does NOT modify the build helper)
  - tests/bdd/features/swg.feature + tests/bdd/steps/algorithms_steps.go extended
  - props_test.go / example_test.go / algoid_test.go extended

From plan 03-02 (cross-validation):
  - scripts/gen-swg-cross-validation.py + testdata/cross-validation/swg/vectors.json
  - swg_test.go: TestSWG_CrossValidation appended (passing)
  - Makefile: regen-swg-cross-validation target

From the existing Phase 2 finalisation infrastructure (02-07):
  - TestGolden_Algorithms_Merge (algorithms_golden_test.go): reads stagingFiles slice; merges via CanonicalMarshalForTest; THIS plan adds `_staging/swg.json` to the list (one-line edit)
  - cross_algorithm_consistency_test.go: TestCrossAlgorithm_OSA_Full_Divergence (the divergence pattern); TestCrossAlgorithm_IdentityConvergence / BothEmptyConvergence / OneEmpty_ScoreAgreement / SingleSubstitution_DistanceAgreement (the convergence patterns with funcs slices)
  - examples/identifier-similarity/main.go: algorithms slice with 6 entries; pairs slice with 7 entries; plaintext table output with fmt.Sprintf("%.4f", ...)
  - examples/identifier-similarity/main_test.go: TestExample_Output (IN-04 line-by-line diff per 02-VERIFICATION.md cleanup); TestExample_ColumnWidths (self-adapts to column count)
  - bench.txt: existing 6-algorithm × 4-benchmark dump from `make bench` (committed in 02-07)
  - llms.txt: existing per-algorithm sections; meta-test ai_friendly_test.go asserts every exported symbol is listed
  - docs/requirements.md §7.1.8: current spec lists 3 SWG functions (Score / ScoreRunes / ScoreWithParams); THIS plan expands to 6 + SWGParams + NewSWGParams per CONTEXT.md §4

From Makefile:
  - `make bench`: regenerates bench.txt
  - `make bench-compare`: runs benchstat against committed bench.txt to detect regressions (> 10% = fail)
  - `make verify-determinism`: runs the cross-platform golden-file gate
  - `make check`: full quality gate
</interfaces>

<canonical_decisions_locked_for_this_plan>
The decisions this plan's executor must honour without re-deriving:

1. **stagingFiles list extension is a one-line edit** — add `"_staging/swg.json"` in alphabetical order (it sorts AFTER `_staging/levenshtein.json`). Do NOT modify TestGolden_Algorithms_Merge's logic; the merge is already generic over the staging files. (02-07 plan + 03-PATTERNS.md algorithms_golden_test.go section)
2. **algorithms.json regeneration uses `-update` ONCE**, then verifies byte-stability on re-run without `-update`. (Phase 2 locked workflow)
3. **The cross-algorithm divergence test pins SWG STRICTLY GREATER THAN Levenshtein** on a substring-containment input — `assert !(gotSWG > gotLev)` is the failure condition; the inequality is strict, not `>=`. (03-PATTERNS.md cross_algorithm_consistency_test.go section)
4. **The funcs-slice extension form is preferred** for convergence tests (one-line edit per test) over creating new SWG-specific TestCrossAlgorithm_SWG_<topic> tests. The existing slice-of-{name, fn} pattern naturally extends. (03-PATTERNS.md cross_algorithm_consistency_test.go section)
5. **examples/identifier-similarity gets ONE new algorithm column "SWG"** — not "Smith-Waterman-Gotoh" (would break column width); not "S-W-G" (less clear). Display name = "SWG". (03-PATTERNS.md examples/identifier-similarity/main.go section)
6. **The IN-04 line-by-line diff in TestExample_Output is preserved** — do not regress to a byte-for-byte assertion on the whole stdout. Re-generating `want` after adding the SWG column means: run `go run .` after main.go is updated, capture stdout, paste verbatim into `want`. The line-by-line diff pattern surfaces any drift on individual rows. (02-VERIFICATION.md IN-04 cleanup)
7. **bench.txt is FULL-REPLACE, not append** — run `make bench` and replace the file wholesale (the locked Phase 2 workflow per 02-07). Do not hand-edit individual rows.
8. **llms.txt section ordering matches the algorithm catalogue order**: Smith-Waterman-Gotoh appears AFTER the Phase 2 six (Levenshtein, Hamming, Jaro, JaroWinkler, Damerau-Levenshtein OSA, Damerau-Levenshtein Full) and BEFORE the Normalisation / Tokenise / errors sections — i.e. immediately before line ~85 (the existing Normalisation section start). (03-PATTERNS.md llms.txt section)
9. **docs/requirements.md §7.1.8 update is mandatory**, not optional. The Raw* surface expansion was approved on 2026-05-14 per CONTEXT.md §4; the spec doc must record the new public surface so the api-ergonomics-reviewer can confirm the final list during PR review.
10. **The Hamming-silent-zero IN-04 / W-2 documentation supersession comment must SURVIVE** the example's column addition — confirm by reading examples/identifier-similarity/main.go after the algorithms-slice edit. (02-VERIFICATION.md / 02-07 plan W-2)
</canonical_decisions_locked_for_this_plan>
</context>

<tasks>

<task type="auto">
  <name>Task 1: Merge SWG into testdata/golden/algorithms.json + cross-algorithm consistency tests</name>
  <files>algorithms_golden_test.go, testdata/golden/algorithms.json, cross_algorithm_consistency_test.go</files>
  <read_first>
    - algorithms_golden_test.go (current state — find TestGolden_Algorithms_Merge from 02-07 and its stagingFiles slice; locate at file scan — likely lines ~160-195; understand the merge logic; confirm buildSWGStagingEntries + TestGolden_SmithWatermanGotoh_Staging from plan 03-01 Task 3 are present and unmodified)
    - testdata/golden/_staging/swg.json (output of plan 03-01 Task 3 — 6 entries, sorted alphabetically; this plan does NOT modify it)
    - testdata/golden/_staging/{damerau_full,damerau_osa,hamming,jaro,jarowinkler,levenshtein}.json (the Phase 2 staging files — read at least one to confirm the schema match)
    - testdata/golden/algorithms.json (current state from Phase 2 finalisation — contains ~32 entries from the six Phase 2 algorithms; this plan REGENERATES it via `-update`)
    - cross_algorithm_consistency_test.go (the canonical convergence/divergence test file from Phase 2 — find TestCrossAlgorithm_OSA_Full_Divergence at lines ~57-75 [the divergence pattern], TestCrossAlgorithm_IdentityConvergence at lines ~81-105 [the funcs-slice convergence pattern], TestCrossAlgorithm_BothEmptyConvergence at lines ~117-124, TestCrossAlgorithm_SingleSubstitution_DistanceAgreement, TestCrossAlgorithm_OneEmpty_ScoreAgreement at lines ~175-182)
    - .planning/phases/03-smith-waterman-gotoh/03-PATTERNS.md (`algorithms_golden_test.go` and `cross_algorithm_consistency_test.go` sections — concrete templates)
    - .planning/phases/03-smith-waterman-gotoh/03-CONTEXT.md `<code_context>` paragraph "substring-containment input" (rationale for the SWG-vs-Levenshtein divergence test)
    - .planning/phases/02-core-character-algorithms-six/02-07-finalisation-PLAN.md Task 2 (the cross-algorithm consistency template)
    - export_test.go (CanonicalMarshalForTest re-export)
    - golden_canonical.go (canonicalMarshal contract)
  </read_first>
  <action>
**Step A — Extend the stagingFiles slice in TestGolden_Algorithms_Merge:**

1. Open algorithms_golden_test.go and locate TestGolden_Algorithms_Merge (created by plan 02-07 Task 1; should contain a `stagingFiles := []string{...}` declaration with six entries).

2. Add the line `"_staging/swg.json",` to the slice in alphabetical order (sorts AFTER `_staging/levenshtein.json`):

       stagingFiles := []string{
           "_staging/damerau_full.json",
           "_staging/damerau_osa.json",
           "_staging/hamming.json",
           "_staging/jaro.json",
           "_staging/jarowinkler.json",
           "_staging/levenshtein.json",
           "_staging/swg.json",            // ADD THIS LINE
       }

3. Do NOT modify any other logic in TestGolden_Algorithms_Merge — the merge body is generic over the slice; the duplicate-Name sanity check still applies (all SWG entry names start with `SmithWatermanGotoh_` so they cannot collide with the Phase 2 entries).

**Step B — Regenerate testdata/golden/algorithms.json:**

1. Run: `go test -run TestGolden_Algorithms_Merge -update -count=1 ./...`. This re-reads all seven staging files, sorts entries by Name, and writes the canonical algorithms.json.

2. Inspect the result:
   - Entry count = sum of staging-file entry counts (Phase 2 ~32 + SWG 6 = ~38 entries; exact count depends on the Phase 2 plans' final entry choices and SWG's 6).
   - Sorted alphabetically by Name (the SWG entries sort with the rest by `SmithWatermanGotoh_*` prefix — e.g. SmithWatermanGotoh_both_empty appears between SmithWatermanGotoh_* entries and the Phase 2 entries that sort earlier alphabetically).
   - Canonical byte form: 2-space indent, trailing LF byte (0x0a), no BOM, no tabs.

3. Commit the regenerated algorithms.json.

4. Run `go test -run TestGolden_Algorithms_Merge -count=1 ./...` WITHOUT `-update`. Must exit 0 with zero diff (byte-stable).

5. Run `make verify-determinism`. Must exit 0.

**Step C — Add the SWG cross-algorithm tests to cross_algorithm_consistency_test.go:**

1. Append a new test at the file trailer (after the existing TestCrossAlgorithm_* tests, before EOF). The new test is the load-bearing local-vs-global-alignment divergence pin per 03-PATTERNS.md:

       // TestCrossAlgorithm_SWG_Levenshtein_SubstringDivergence asserts the
       // load-bearing local-vs-global-alignment claim: on a substring-
       // containment input, Smith-Waterman-Gotoh (LOCAL alignment) scores
       // STRICTLY HIGHER than Levenshtein (GLOBAL edit distance), because SWG
       // finds the substring while Levenshtein counts every uncovered position
       // as an edit.
       //
       // "http_request" is fully contained in "http_request_header_fields":
       //   - SmithWatermanGotohScore = 1.0   (full local match found; clamp
       //                                      returns 1 because raw == min(len))
       //   - LevenshteinScore        ≈ 0.46  (≈ 14 edits over max-length 26)
       //
       // This test will fail if either algorithm regresses on this contract.
       func TestCrossAlgorithm_SWG_Levenshtein_SubstringDivergence(t *testing.T) {
           a, b := "http_request", "http_request_header_fields"
           gotSWG := fuzzymatch.SmithWatermanGotohScore(a, b)
           gotLev := fuzzymatch.LevenshteinScore(a, b)
           if !(gotSWG > gotLev) {
               t.Errorf("SWG (%v) must score STRICTLY higher than Levenshtein (%v) on substring-containment pair %q/%q (local-vs-global divergence)",
                   gotSWG, gotLev, a, b)
           }
       }

2. Extend the funcs slice in TestCrossAlgorithm_IdentityConvergence to include SWG:
   - Locate the existing test (lines ~81-105 per 02-07 plan). The funcs slice has 6 entries (one per Phase 2 algorithm — LevenshteinScore, HammingScore, JaroScore, JaroWinklerScore, DamerauLevenshteinOSAScore, DamerauLevenshteinFullScore).
   - Add `{name: "SmithWatermanGotoh", fn: fuzzymatch.SmithWatermanGotohScore}` to the slice (preserve the existing struct shape; the slice element type was locked in 02-07).

3. Extend the funcs slice in TestCrossAlgorithm_BothEmptyConvergence (same pattern — add SWG entry).

4. Extend the funcs slice in TestCrossAlgorithm_OneEmpty_ScoreAgreement (same pattern — add SWG entry).

5. TestCrossAlgorithm_SingleSubstitution_DistanceAgreement covers Distance variants (Levenshtein, DL-OSA, DL-Full, Hamming) — SWG has NO Distance variant per CONTEXT.md §7 inherited from IN-06. DO NOT add SWG to this test.

6. TestCrossAlgorithm_OSA_Full_Divergence is the OSA-vs-Full pin from 02-07 — leave it unchanged.

**Step D — Verify:**

       go test -race -shuffle=on -count=1 -run 'TestGolden_Algorithms_Merge|TestCrossAlgorithm' ./...

Must exit 0. The new TestCrossAlgorithm_SWG_Levenshtein_SubstringDivergence must pass (gotSWG = 1.0, gotLev ≈ 0.46, so the strict-greater-than check passes). The four extended convergence/agreement tests must pass (SWG behaves identically to the Phase 2 algorithms on identity, both-empty, one-empty inputs per its locked-in short-circuits).
  </action>
  <verify>
    <automated>go test -race -shuffle=on -count=1 -run 'TestGolden_Algorithms_Merge|TestCrossAlgorithm' ./... && make verify-determinism && jq '[.entries[].algorithm] | unique | sort' testdata/golden/algorithms.json | grep -qE 'SmithWatermanGotoh'</automated>
  </verify>
  <acceptance_criteria>
    - algorithms_golden_test.go's TestGolden_Algorithms_Merge stagingFiles slice contains `_staging/swg.json` (`grep -c '"_staging/swg.json"' algorithms_golden_test.go` returns ≥ 1).
    - testdata/golden/algorithms.json contains entries from all seven algorithms; `jq '[.entries[].algorithm] | unique | sort' testdata/golden/algorithms.json` returns a JSON array containing the seven names: DamerauLevenshteinFull, DamerauLevenshteinOSA, Hamming, Jaro, JaroWinkler, Levenshtein, SmithWatermanGotoh.
    - algorithms.json contains the SmithWatermanGotoh_one_long_gap_canary entry (`grep -c 'SmithWatermanGotoh_one_long_gap_canary' testdata/golden/algorithms.json` returns ≥ 1).
    - algorithms.json is in canonical byte form (2-space indent, trailing LF, no BOM, sorted entries by Name).
    - Re-running `go test -run TestGolden_Algorithms_Merge -count=1 ./...` WITHOUT `-update` exits 0 with zero diff (byte-stable).
    - All seven _staging/*.json files are still present in the repo (NOT deleted by the merge).
    - cross_algorithm_consistency_test.go contains `func TestCrossAlgorithm_SWG_Levenshtein_SubstringDivergence` exactly once.
    - The new test references both `fuzzymatch.SmithWatermanGotohScore` and `fuzzymatch.LevenshteinScore` (`grep -c 'fuzzymatch.SmithWatermanGotohScore' cross_algorithm_consistency_test.go` returns ≥ 1; `grep -c 'fuzzymatch.LevenshteinScore' cross_algorithm_consistency_test.go` returns ≥ 1).
    - The new test uses the STRICT inequality form (`!(gotSWG > gotLev)` — verifiable by `grep -cE '!\(gotSWG > gotLev\)|gotSWG <= gotLev' cross_algorithm_consistency_test.go` returning ≥ 1).
    - TestCrossAlgorithm_IdentityConvergence, TestCrossAlgorithm_BothEmptyConvergence, TestCrossAlgorithm_OneEmpty_ScoreAgreement each contain a SmithWatermanGotoh entry in their funcs slice (`grep -c 'SmithWatermanGotoh' cross_algorithm_consistency_test.go` returns ≥ 4 — three convergence tests + the new divergence test).
    - `go test -race -shuffle=on -count=1 -run 'TestCrossAlgorithm_SWG_Levenshtein_SubstringDivergence' ./...` exits 0.
    - `go test -race -shuffle=on -count=1 -run 'TestCrossAlgorithm' ./...` exits 0 (all cross-algorithm tests pass, including the SWG-extended convergence tests).
    - `make verify-determinism` exits 0 (merged algorithms.json byte-stable + cross-platform-identical).
  </acceptance_criteria>
  <behavior>
    - testdata/golden/algorithms.json is the merged canonical golden file containing entries from all seven algorithms (six Phase 2 + SWG).
    - The merge is byte-stable across re-runs and (verified by CI matrix) cross-platform-identical.
    - The SWG-vs-Levenshtein divergence test pins the load-bearing local-vs-global-alignment claim — a regression in either algorithm on this contract triggers a clear, attributable failure.
    - The funcs-slice-extended convergence tests pin that SWG identity (Score(x,x)==1.0), both-empty (Score("","")==1.0), and one-empty (Score("","abc")==0.0) behave identically to the Phase 2 algorithms.
    - The new test's failure message includes both algorithm scores and the input pair — sufficient for one-step triage.
  </behavior>
  <done>
    algorithms_golden_test.go stagingFiles extended; testdata/golden/algorithms.json regenerated and byte-stable; cross_algorithm_consistency_test.go has the new divergence test + extended convergence funcs slices; `make verify-determinism` green.
  </done>
</task>

<task type="auto">
  <name>Task 2: identifier-similarity example — add SWG column (main.go + main_test.go)</name>
  <files>examples/identifier-similarity/main.go, examples/identifier-similarity/main_test.go</files>
  <read_first>
    - examples/identifier-similarity/main.go (current state — locate the algorithms slice at lines ~71-81 from 02-07's plan; understand the 6-column table-printing logic; confirm the W-2 Hamming-silent-zero documentation comment is still present)
    - examples/identifier-similarity/main_test.go (current state — `want` constant at lines ~40-49; TestExample_Output uses the IN-04 line-by-line diff per 02-VERIFICATION.md cleanup commit a1f02f6; TestExample_ColumnWidths self-adapts to column count per IN-04 cleanup)
    - examples/identifier-similarity/go.mod (replace directive — unchanged; this plan does not modify the module structure)
    - .planning/phases/02-core-character-algorithms-six/02-VERIFICATION.md IN-04 cleanup section (line-by-line diff pattern; column-widths self-adapt rationale)
    - .planning/phases/03-smith-waterman-gotoh/03-CONTEXT.md `<code_context>` paragraph "7-row × 6-column → 7-row × 7-column" (column count goes from 6 → 7 algorithm columns)
    - .planning/phases/03-smith-waterman-gotoh/03-PATTERNS.md (`examples/identifier-similarity/main.go` and `examples/identifier-similarity/main_test.go` sections — concrete templates)
    - swg.go (confirm fuzzymatch.SmithWatermanGotohScore signature: `func SmithWatermanGotohScore(a, b string) float64` — matches the slice's `func(a, b string) float64` type)
  </read_first>
  <action>
**Step A — Add SWG to the algorithms slice in main.go:**

1. Open examples/identifier-similarity/main.go and locate the algorithms slice (created by plan 02-07; should have 6 entries — Levenshtein, DL-OSA, DL-Full, Hamming, Jaro, Jaro-Winkler).

2. Append a new entry as the 7th element:

       var algorithms = []struct {
           name string
           fn   func(a, b string) float64
       }{
           {"Levenshtein",  fuzzymatch.LevenshteinScore},
           {"DL-OSA",       fuzzymatch.DamerauLevenshteinOSAScore},
           {"DL-Full",      fuzzymatch.DamerauLevenshteinFullScore},
           {"Hamming",      fuzzymatch.HammingScore},
           {"Jaro",         fuzzymatch.JaroScore},
           {"Jaro-Winkler", fuzzymatch.JaroWinklerScore},
           {"SWG",          fuzzymatch.SmithWatermanGotohScore},   // ADD THIS LINE
       }

   Display name `"SWG"` (4 characters) — keeps the existing column width allowance reasonable. Do NOT use longer forms like `"Smith-Waterman-Gotoh"` (20 chars — would force column-width changes throughout the formatting logic).

3. Confirm the W-2 / IN-04 Hamming-silent-zero documentation comment (the file-level godoc note explaining that Hamming on length-mismatched pairs shows `0.0000` not `ERR`) is still present after the edit. If the comment was accidentally removed during a refactor, restore it from 02-07's plan W-2 text.

4. Do NOT modify any other logic in main.go — the formatting loop is generic over the algorithms slice; adding one entry produces one new column automatically.

**Step B — Regenerate the want constant in main_test.go:**

1. Build and run the example to capture the new stdout:

       (cd examples/identifier-similarity && go run .) > /tmp/example-output.txt

2. Inspect /tmp/example-output.txt — should be the 1 header row + 7 pair rows × 7 algorithm columns. Confirm:
   - The header row has 7 algorithm column names (Levenshtein, DL-OSA, DL-Full, Hamming, Jaro, Jaro-Winkler, SWG) plus the pair column.
   - Each pair row has 7 decimal scores formatted as `%.4f` (e.g. `1.0000`, `0.5714`).
   - For Hamming columns where the pair has unequal length: the cell shows `0.0000` (the W-2 / IN-04 locked silent-zero policy — NOT `ERR`).
   - The SWG column shows non-trivial values for the 7 pairs (`user_id` vs `userId` should be high; `latitude` vs `longitude` should be lower; etc.).

3. Open examples/identifier-similarity/main_test.go and replace the existing `want` constant with the captured stdout, formatted as a Go raw-string literal:

       const want = `<paste captured stdout verbatim — preserve every space and newline>`

   The raw-string form (backticks) preserves whitespace exactly. Do NOT use a double-quoted string with escapes.

4. Confirm TestExample_Output's logic is unchanged — it does the IN-04 line-by-line diff against `want`. The diff loop self-adapts to whatever `want` is set to.

5. Confirm TestExample_ColumnWidths logic is unchanged — it derives column widths from `want` per IN-04 / 02-VERIFICATION.md.

**Step C — Verify:**

       (cd examples/identifier-similarity && go test -race -count=1 ./...)

Must exit 0. TestExample_Output passes (line-by-line stdout matches `want`); TestExample_ColumnWidths passes (column widths consistent across header + 7 pair rows).

       (cd examples/identifier-similarity && go run .)

Run the example manually. Confirm the output is readable, all 7 pair rows present, all 7 algorithm columns present (with SWG as the new rightmost column), all cells show 4-decimal scores with no `ERR` strings.
  </action>
  <verify>
    <automated>(cd examples/identifier-similarity && go test -race -count=1 ./...) && grep -c '"SWG"' examples/identifier-similarity/main.go | xargs -I{} test {} -ge 1 && grep -c 'SWG' examples/identifier-similarity/main_test.go | xargs -I{} test {} -ge 1</automated>
  </verify>
  <acceptance_criteria>
    - examples/identifier-similarity/main.go contains `{"SWG", fuzzymatch.SmithWatermanGotohScore}` (or equivalent — `grep -E '"SWG".*SmithWatermanGotohScore' examples/identifier-similarity/main.go` returns ≥ 1 match).
    - examples/identifier-similarity/main.go algorithms slice now has 7 entries (verifiable by counting struct literal elements; the existing 6 + the new SWG entry).
    - examples/identifier-similarity/main.go retains the W-2 / IN-04 Hamming-silent-zero documentation comment (verifiable by `grep -c 'silent.zero\|0\.0000.*ERR\|ERR.*0\.0000' examples/identifier-similarity/main.go` returns ≥ 1 — match the existing comment pattern from 02-07).
    - examples/identifier-similarity/main_test.go's `want` constant contains the literal string "SWG" in the header row (`grep -c 'SWG' examples/identifier-similarity/main_test.go` returns ≥ 1).
    - examples/identifier-similarity/main_test.go's `want` constant is regenerated with the captured stdout from the updated main.go (running `go run .` produces output byte-identical to `want`).
    - `(cd examples/identifier-similarity && go run .)` exits 0 and prints a 7-row × 8-column (1 pair-col + 7 algorithm-cols) table.
    - `(cd examples/identifier-similarity && go test -race -count=1 ./...)` exits 0 — TestExample_Output passes (line-by-line diff per IN-04); TestExample_ColumnWidths passes (column widths consistent).
    - No cell in the captured stdout contains the literal string `ERR` (the W-2 silent-zero policy is preserved).
    - The SWG column shows reasonable scores: for `user_id` vs `userId` (case difference) the SWG score is high (> 0.5); for `latitude` vs `longitude` it is lower but > 0.
    - `(cd examples/identifier-similarity && go mod tidy)` produces zero diff (go.mod still has the `replace` directive; no new require entries).
  </acceptance_criteria>
  <behavior>
    - The example demonstrates all 7 Phase 1-3 algorithms (six Phase 2 + SWG) side-by-side on real-world database identifier pairs.
    - The meta-test pins the example's stdout byte-stable across runs and platforms via the IN-04 line-by-line diff pattern.
    - The Hamming column on length-mismatched pairs still shows `0.0000` (W-2 supersession preserved).
    - The SWG column shows the local-alignment-vs-edit-distance characteristic: substring-containment pairs score higher in SWG than in Levenshtein.
  </behavior>
  <done>
    examples/identifier-similarity/main.go has the SWG column; examples/identifier-similarity/main_test.go's want constant is regenerated; both tests pass; the example runs and prints a 7-row × 7-algorithm-column table.
  </done>
</task>

<task type="auto">
  <name>Task 3: bench.txt + llms.txt + docs/requirements.md §7.1.8 (final phase-wide updates) + make check</name>
  <files>bench.txt, llms.txt, docs/requirements.md</files>
  <read_first>
    - Makefile (find the `bench` target and any `bench-baseline` / `bench-compare` targets; understand how bench.txt is produced — typically `go test -bench=. -benchmem -run=^$ -count=N ./... | tee bench.txt`)
    - bench.txt (current state — committed in plan 02-07; contains 6 Phase 2 algorithm benchmark series at the top; lines 5-14 show the per-benchmark format)
    - llms.txt (current state — `### Levenshtein` at line ~47, six per-algorithm sections, then `### Normalisation` at line ~85; the file follows the format `- type X struct`, `- func Name(args) ReturnType` one symbol per line)
    - swg.go (confirm the final exported symbol names from plan 03-01 Task 1 — type SWGParams, NewSWGParams, SmithWatermanGotohScore, *Runes, *WithParams, RawScore, RawScoreRunes, RawScoreWithParams = 8 new symbols)
    - ai_friendly_test.go (the llms.txt sync meta-test — parses go/ast and asserts every exported symbol is listed; understand what the test expects for the 8 new symbols)
    - docs/requirements.md §7.1.8 (current state — lists 3 SWG functions per the original spec; THIS plan expands to 6 functions + SWGParams + NewSWGParams per CONTEXT.md §4)
    - .planning/phases/03-smith-waterman-gotoh/03-CONTEXT.md §4 (Raw* surface expansion: 3 normalised + 3 Raw + SWGParams + NewSWGParams = 8 new exports; spec doc update is a Phase 3 deliverable)
    - .planning/phases/03-smith-waterman-gotoh/03-PATTERNS.md (`bench.txt`, `llms.txt`, `docs/requirements.md` sections — concrete templates including the 9-line llms.txt block)
  </read_first>
  <action>
**Step A — Regenerate bench.txt:**

1. Run `make bench` from the repo root. This invokes the bench target (typically `go test -bench=. -benchmem -run=^$ -count=10 ./... | tee bench.txt` — confirm the exact command in the Makefile).

2. Confirm the produced bench.txt contains:
   - The existing six Phase 2 algorithm benchmark series (BenchmarkLevenshteinScore_*, BenchmarkHammingScore_*, etc. — preserved from 02-07).
   - The six new SmithWatermanGotoh* benchmark series in alphabetical order:
     - BenchmarkSmithWatermanGotohScore_ASCII_Short
     - BenchmarkSmithWatermanGotohScore_ASCII_Medium
     - BenchmarkSmithWatermanGotohScore_ASCII_Long
     - BenchmarkSmithWatermanGotohScore_Unicode_Short
     - BenchmarkSmithWatermanGotohScore_WithParams_ASCII_Short
     - BenchmarkSmithWatermanGotohRawScore_ASCII_Short
   - Each benchmark series has ~10 rows from `count=10` (or whatever the Makefile's count value is).

3. Commit the regenerated bench.txt (full-replace, not append — locked Phase 2 workflow).

4. Run `make bench-compare` to confirm the new baseline produces no regression on a re-run (informational; a fresh baseline by definition has zero diff against itself). If `make bench-compare` shows a > 10% regression against the previous (Phase 2) bench.txt on any of the six Phase 2 series, investigate: (a) the regression may be noise — re-run; (b) the regression may be real (e.g. a Phase 3 change inadvertently affected Phase 2 algorithm performance) — investigate before committing; (c) document any unresolved regression in 03-03-SUMMARY.md with rationale.

**Step B — Append the SWG section to llms.txt:**

1. Locate the existing per-algorithm sections in llms.txt. They are typically headed by `### <algorithm name>` and contain one symbol per line in the form `- type X struct` or `- func Name(args) ReturnType`.

2. Append the new SWG section IMMEDIATELY BEFORE the `## Normalisation` (or equivalent) section break — i.e. at the end of the per-algorithm catalogue (after Jaro-Winkler if alphabetical, or per the existing structure ordering — confirm during execution):

       ### Smith-Waterman-Gotoh local-alignment similarity

       - type SWGParams struct
       - func NewSWGParams() SWGParams
       - func SmithWatermanGotohScore(a, b string) float64
       - func SmithWatermanGotohScoreRunes(a, b string) float64
       - func SmithWatermanGotohScoreWithParams(a, b string, params SWGParams) float64
       - func SmithWatermanGotohRawScore(a, b string) float64
       - func SmithWatermanGotohRawScoreRunes(a, b string) float64
       - func SmithWatermanGotohRawScoreWithParams(a, b string, params SWGParams) float64

   Eight lines (1 type + 1 constructor + 6 funcs). Plus the section header (`### Smith-Waterman-Gotoh local-alignment similarity`) and a blank line above and below per the file's existing style.

3. Note: `AlgoSmithWatermanGotoh` constant is already listed in llms.txt at the AlgoID enum section (line ~30, added in Phase 1 when the enum was declared). Do NOT duplicate it under the SWG algorithm section.

4. If llms.txt has a top-level `## Algorithms` summary table or status list with per-algorithm status indicators (e.g. "Levenshtein — shipped", "SWG — pending"), flip the SWG row from "pending" to "shipped" — confirm by reading the existing llms.txt structure during execution. If no such table exists, this step is a no-op.

5. Run `go test -race -shuffle=on -count=1 -run 'TestAIFriendly' ./...`. Must exit 0 — the meta-test confirms all 8 new exported symbols are listed in llms.txt and that no extra symbols are listed that don't exist in the code.

**Step C — Update docs/requirements.md §7.1.8:**

1. Locate the §7.1.8 section in docs/requirements.md. The current state (Phase 0 spec) lists 3 SWG functions:
   - SmithWatermanGotohScore(a, b string) float64
   - SmithWatermanGotohScoreRunes(a, b string) float64
   - SmithWatermanGotohScoreWithParams(a, b string, params SWGParams) float64
   (Plus SWGParams as the parameter type.)

2. Expand the §7.1.8 public-API table / list to include the three new Raw* functions per CONTEXT.md §4:
   - SmithWatermanGotohRawScore(a, b string) float64
   - SmithWatermanGotohRawScoreRunes(a, b string) float64
   - SmithWatermanGotohRawScoreWithParams(a, b string, params SWGParams) float64

3. Also list:
   - The SWGParams struct (Match, Mismatch, GapOpen, GapExtend float64 fields).
   - The NewSWGParams() constructor returning the documented defaults (1.0, -1.0, -1.5, -0.5).

4. Add a brief paragraph or note in §7.1.8 explaining the Raw* surface:

       The Raw* variants return the UNCLAMPED raw alignment score (which may be
       negative or exceed min(len(a), len(b)) when custom params produce
       Match > 1.0). Advanced consumers (bioinformatics, schema-similarity
       research) who want absolute alignment quality unaffected by the
       normalisation choice should use *RawScore; consumers who want a
       comparable [0, 1] similarity should use *Score (which applies
       clamp(raw / min(len(a), len(b)), 0, 1)). Decision recorded 2026-05-14
       per phase 03 CONTEXT.md §4.

5. Confirm the §7.1.8 update preserves the rest of the section's structure (parameter descriptions, default values, complexity statement, edge cases, cross-validation note about biopython); only the public-API list/table needs the Raw* additions.

6. Run `make markdownlint` (or equivalent — check Makefile for the markdownlint target) to confirm docs/requirements.md still passes markdown linting.

**Step D — Final phase-wide quality gate:**

       make check

Must exit 0. This is the LOAD-BEARING pre-shippable gate for Phase 3:
- lint (golangci-lint v2)
- vet
- race (`go test -race`)
- coverage (≥ 95% overall, ≥ 90% per file, 100% on public API)
- verify-license-headers (all new .go files Apache-2.0)
- verify-no-runtime-deps (root go.mod allowlist-clean)
- tidy-check (zero diff after `go mod tidy`)
- verify-determinism (the merged algorithms.json byte-stable cross-platform — Phase 3's load-bearing gate)
- BDD (godog suite green; goleak detects no leaks; includes the new swg.feature)
- markdownlint (docs/requirements.md + llms.txt + README parse clean)
- vulncheck (govulncheck — no HIGH/CRITICAL CVEs)
- gosec (no high-severity issues)

If `make check` fails, the failure is BLOCKING for Phase 3 shippability. Fix the failure and re-run before declaring the phase complete.
  </action>
  <verify>
    <automated>grep -c 'BenchmarkSmithWatermanGotohScore' bench.txt | xargs -I{} test {} -ge 6 && grep -c 'SmithWatermanGotohScore\|SmithWatermanGotohRawScore\|SWGParams\|NewSWGParams' llms.txt | xargs -I{} test {} -ge 8 && grep -c 'SmithWatermanGotohRawScore' docs/requirements.md | xargs -I{} test {} -ge 3 && make check</automated>
  </verify>
  <acceptance_criteria>
    - bench.txt exists at repo root and contains benchmark output for all seven algorithms (six Phase 2 + SWG); `grep -c 'BenchmarkSmithWatermanGotohScore' bench.txt` returns ≥ 6 (the six new SWG benchmark series), and `grep -c 'BenchmarkLevenshteinScore' bench.txt` returns ≥ 4 (the existing four Phase 2 Levenshtein series — preserved).
    - bench.txt has been regenerated this plan (file mtime newer than plan 02-07's commit; or the file's content reflects the new SWG rows).
    - `make bench-compare` exits 0 (or reports a regression that has been investigated and documented in 03-03-SUMMARY.md).
    - llms.txt has a new `### Smith-Waterman-Gotoh local-alignment similarity` section header (`grep -c '### Smith-Waterman-Gotoh' llms.txt` returns ≥ 1).
    - llms.txt lists all 8 new SWG exported symbols (1 type + 1 constructor + 6 functions): SWGParams, NewSWGParams, SmithWatermanGotohScore, SmithWatermanGotohScoreRunes, SmithWatermanGotohScoreWithParams, SmithWatermanGotohRawScore, SmithWatermanGotohRawScoreRunes, SmithWatermanGotohRawScoreWithParams. Verifiable by `grep -c 'SmithWatermanGotoh\|SWGParams\|NewSWGParams' llms.txt` returning ≥ 8.
    - `go test -race -shuffle=on -count=1 -run 'TestAIFriendly' ./...` exits 0 (meta-test confirms llms.txt sync).
    - docs/requirements.md §7.1.8 lists all 6 SWG public functions (`grep -c 'SmithWatermanGotohScore\|SmithWatermanGotohRawScore' docs/requirements.md` returns ≥ 6); the Raw* variants are explicitly mentioned (`grep -c 'SmithWatermanGotohRawScore' docs/requirements.md` returns ≥ 3 — once per Raw variant).
    - docs/requirements.md §7.1.8 references the 2026-05-14 CONTEXT.md §4 decision OR includes an equivalent rationale paragraph explaining the Raw* surface.
    - docs/requirements.md §7.1.8 lists the SWGParams struct fields and NewSWGParams constructor (`grep -c 'SWGParams\|NewSWGParams' docs/requirements.md` returns ≥ 2).
    - `bash scripts/verify-license-headers.sh` exits 0.
    - `bash scripts/verify-no-runtime-deps.sh` exits 0 (root go.mod still only `golang.org/x/text`).
    - `make verify-determinism` exits 0 (merged algorithms.json byte-stable cross-platform — Task 1 verified this, but `make check` re-runs it).
    - `make check` exits 0 — the load-bearing pre-shippable gate for Phase 3.
  </acceptance_criteria>
  <behavior>
    - bench.txt is the regenerated benchstat baseline including all seven algorithms; future PRs run `make bench-compare` against it to detect regressions.
    - llms.txt is the AI-friendly catalogue updated to include SWG's 8 new exported symbols; the meta-test enforces sync.
    - docs/requirements.md §7.1.8 is the authoritative public-API spec for SWG, now including the Raw* surface expansion per CONTEXT.md §4.
    - The full Phase 3 quality gate (lint, vet, race, coverage, license headers, no-runtime-deps, tidy-check, verify-determinism, BDD, markdownlint, vulncheck, gosec) passes via `make check`.
  </behavior>
  <done>
    bench.txt regenerated; llms.txt extended with the SWG section (8 new symbols); docs/requirements.md §7.1.8 updated to record the Raw* surface expansion; make check green. Phase 3 is shippable.
  </done>
</task>

</tasks>

<threat_model>
## Trust Boundaries

| Boundary | Description |
|----------|-------------|
| Caller → fuzzymatch.* algorithms (seven total — six Phase 2 + SWG) | Same trust boundaries as plans 03-01 and 03-02; this Wave 3 plan introduces no new untrusted-input paths. The example program uses hardcoded inputs. |
| go test → testdata/golden/_staging/*.json (seven files) | Read-only; the merge test parses each staging file. JSON parse errors fail the test loudly (no silent corruption). |
| go test → testdata/golden/algorithms.json | Read-only (after `-update` regeneration). Cross-platform CI matrix diff is the determinism gate. |
| go run examples/identifier-similarity → stdout → meta-test | Pure-function pipeline; deterministic output (DET-02 / DET-04 enforced per algorithm). |

## STRIDE Threat Register

| Threat ID | Category | Component | Disposition | Mitigation Plan |
|-----------|----------|-----------|-------------|-----------------|
| T-3-FN-01 | Tampering | A staging file is corrupted between Wave 2 and Wave 3 (e.g. someone hand-edits _staging/swg.json) | mitigate | TestGolden_Algorithms_Merge re-parses each staging file via encoding/json; corruption causes the test to fail fast with the file name in the error message. The duplicate-Name check (inherited from 02-07) catches accidental cross-staging-file collisions. PR review enforces correctness before merge. |
| T-3-FN-02 | Tampering | The merged algorithms.json is hand-edited to weaken the cross-platform determinism gate | mitigate | The file is regenerated from staging files on every `-update` run; any hand-edit is overwritten. CI matrix runs `make verify-determinism` on five platforms; any inconsistency surfaces. |
| T-3-FN-03 | Information Disclosure | The example's stdout meta-test fails on a CI platform due to non-deterministic float output | mitigate | The seven algorithms are deterministic (DET-02 / DET-04 enforced per algorithm — plan 03-01's property tests + plan 03-02's biopython cross-validation cover SWG specifically). Float formatting via `fmt.Sprintf("%.4f", score)` is platform-stable for the values in [0,1]. If any platform reports a diff, the underlying bug is in the algorithm — fix at the algorithm level, not by relaxing the meta-test. |
| T-3-FN-04 | Denial of Service | bench.txt grows without bound across many releases | accept | bench.txt is overwritten on each baseline update (not appended to). Long-term, benchstat consumes the file via `make bench-compare` for diff comparison; the file remains O(per-algorithm-benchmark-count) in size. |
| T-3-FN-05 | Spoofing | Wave 3's algorithms.json drift makes axonops/audit (Phase 11 consumer) get different scores than expected | mitigate | DET-02 (algorithm score stability across patch releases) is enforced via the golden file gate. Any score change requires a minor version bump per REL-07. The cross_algorithm_consistency_test.go gate makes drift in identity / divergence / convergence behaviour visible immediately. |
| T-3-FN-06 | Tampering | docs/requirements.md §7.1.8 is silently expanded beyond what CONTEXT.md §4 approves (e.g. adding a SmithWatermanGotohAlignment function not actually shipped) | mitigate | api-ergonomics-reviewer agent reviews the §7.1.8 update during PR per CLAUDE.md "Workflow — Agent Gates" §5. The plan's acceptance criteria pin the exact 8 symbols listed; any drift surfaces in PR diff. |

No high-severity items in this plan beyond what 03-01 and 03-02 already covered. T-3-FN-03 carries forward the float-determinism mitigation chain established by plan 03-01 (T-3-03). Plan passes the security gate.
</threat_model>

<verification>
1. Build: `go build ./...` exits 0 (root module).
2. Example build: `(cd examples/identifier-similarity && go build ./...)` exits 0.
3. Tidy: `(cd examples/identifier-similarity && go mod tidy)` exits 0 with zero diff.
4. Vet: `go vet ./...` exits 0.
5. License headers: `bash scripts/verify-license-headers.sh` exits 0.
6. No-runtime-deps: `bash scripts/verify-no-runtime-deps.sh` exits 0.
7. `go test -race -shuffle=on -count=1 -run 'TestGolden_Algorithms_Merge|TestCrossAlgorithm|TestAIFriendly' ./...` exits 0.
8. `(cd examples/identifier-similarity && go test -race -count=1 ./...)` exits 0 (TestExample_Output + TestExample_ColumnWidths green).
9. `(cd tests/bdd && go test -race -shuffle=on -count=1 ./...)` exits 0 (all seven algorithms' BDD scenarios pass, including swg.feature).
10. `make verify-determinism` exits 0 (merged algorithms.json byte-stable cross-platform).
11. bench.txt contains all seven algorithms' benchmarks; `make bench-compare` shows no unexpected regression.
12. llms.txt lists all 8 new SWG symbols; `TestAIFriendly` meta-test green.
13. docs/requirements.md §7.1.8 lists all 6 SWG public functions + SWGParams + NewSWGParams; markdownlint green.
14. `make check` exits 0 — the load-bearing pre-shippable gate.
</verification>

<success_criteria>
- testdata/golden/algorithms.json contains entries from all seven algorithms (six Phase 2 + SWG), byte-stable, cross-platform-identical (DET-02 fully satisfied for the entire algorithm catalogue through Phase 3).
- cross_algorithm_consistency_test.go pins the SWG-vs-Levenshtein local-vs-global-alignment divergence on substring-containment inputs AND extends the identity/both-empty/one-empty convergence tests to include SWG.
- examples/identifier-similarity/ now demonstrates all seven Phase 1-3 algorithms side-by-side, meta-tested for stdout stability (DX-05 carries forward).
- bench.txt is regenerated with the new six SmithWatermanGotoh* benchmark series; `make bench-compare` confirms no regression on the Phase 2 baselines.
- llms.txt lists all 8 new SWG exported symbols; the meta-test enforces sync going forward.
- docs/requirements.md §7.1.8 is updated to record the Raw* surface expansion per CONTEXT.md §4; the api-ergonomics-reviewer can verify the final list during PR review.
- Phase 3 ROADMAP success criteria #1-#4 are demonstrably satisfied (cross-validation against biopython including one-long-gap canary — plan 03-02; primary-source citations + erratum documented — plan 03-01; configurable affine gap + property tests — plan 03-01; allocation budget + two-row DP + cross-platform golden file + BDD long-gap scenario — plans 03-01 + this plan).
- `make check` exits 0 — Phase 3 is shippable.
- Phase 3 is ready for `/gsd-verify-work` and the algorithm-correctness-reviewer / api-ergonomics-reviewer / determinism-reviewer / security-reviewer / code-reviewer / go-quality agent gates.
</success_criteria>

<output>
After completion, create `.planning/phases/03-smith-waterman-gotoh/03-03-swg-finalisation-SUMMARY.md` per the standard summary template, recording:
- The exact entry count in the merged algorithms.json (sum of staging-file entry counts; expect Phase 2 ~32 + SWG 6 = ~38).
- The captured stdout from the regenerated `examples/identifier-similarity/` (as a verbatim block or by reference to the committed `want` constant in main_test.go) — for traceability.
- The six new BenchmarkSmithWatermanGotoh* numbers from the regenerated bench.txt (summary table: benchmark × ns/op × B/op × allocs/op).
- Any regressions on the Phase 2 benchmarks detected by `make bench-compare`, with rationale and resolution.
- Coverage percentages (overall + per-file swg.go + public-symbol).
- The exact 8 SWG public symbols listed in llms.txt (verify zero drift from the plan's locked list).
- The docs/requirements.md §7.1.8 update text (paste the new public-surface table or the diff against the previous version).
- The git-log evidence that the Hamming-silent-zero IN-04 / W-2 documentation comment in examples/identifier-similarity/main.go survived the SWG column addition.
- Any deviations from the plan and their rationale.
- A self-check confirming: (a) `make check` is green; (b) TestSWG_CrossValidation is green (plan 03-02's gate); (c) the new TestCrossAlgorithm_SWG_Levenshtein_SubstringDivergence is green; (d) the regenerated algorithms.json is byte-stable under re-run without `-update`; (e) Phase 3 is shippable.
- The hand-off contract to Phase 4: the staging-merge pattern continues (`_staging/strcmp95.json`, `_staging/lcsstr.json`, `_staging/ratcliff_obershelp.json` extend the stagingFiles slice in 04-XX); the bench.txt + llms.txt + docs/requirements.md update pattern continues for each new algorithm; the cross-algorithm consistency convergence tests extend each new algorithm's funcs-slice entry; the example column-addition pattern continues for major new algorithms (or batches; choice per Phase 4 planning).
</output>
</content>
