---
phase: 02-core-character-algorithms-six
plan: 07
type: execute
wave: 4
depends_on:
  - 02-01-levenshtein
  - 02-02-hamming
  - 02-03-jaro
  - 02-04-jaro-winkler
  - 02-05-damerau-levenshtein-osa
  - 02-06-damerau-levenshtein-full
files_modified:
  - examples/identifier-similarity/main.go
  - examples/identifier-similarity/main_test.go
  - examples/identifier-similarity/go.mod
  - examples/identifier-similarity/go.sum
  - cross_algorithm_consistency_test.go
  - algorithms_golden_test.go
  - testdata/golden/algorithms.json
  - bench.txt
autonomous: true
requirements:
  - DET-02
  - DX-02
  - DX-05
  - TEST-01
tags: [finalisation, identifier-similarity-example, cross-algorithm-consistency, algorithms-json-merge, bench-baseline]

must_haves:
  truths:
    - "examples/identifier-similarity/main.go is a runnable `package main` program that prints a 7-row × 6-column table of similarity scores for hardcoded database column-name pairs (per CONTEXT.md identifier-similarity decision)"
    - "Running `go run ./examples/identifier-similarity/` prints the table to stdout with all six algorithms as columns: Levenshtein, DL-OSA, DL-Full, Hamming, Jaro, JaroWinkler"
    - "The example covers the 7 reference pairs from CONTEXT.md: user_id/userId, created_at/creationTimestamp, status/state, email/e_mail, org_id/organisation_id, latitude/longitude, is_deleted/is_active"
    - "Hamming cells for length-mismatched pairs show 0.0000 (per the locked silent-zero policy from plan 02-02) — never `ERR` (the example reflects the actual behaviour, not the spec's earlier illustrative `(int, error)` shape)"
    - "examples/identifier-similarity/main_test.go's TestExample_Output captures the example's stdout and asserts byte-stable output (the `want` constant is committed and re-runs produce zero diff)"
    - "testdata/golden/algorithms.json contains entries from ALL six algorithms (Levenshtein, Hamming, Jaro, JaroWinkler, DL-OSA, DL-Full) merged from the per-algorithm `_staging/<algo>.json` files, sorted by Name, in canonical byte form"
    - "The merge step is reproducible: running `TestGolden_Algorithms_Merge -update` re-builds algorithms.json from the staging files; running without -update produces zero diff"
    - "cross_algorithm_consistency_test.go pins the load-bearing cross-algorithm contracts: DL-OSA(\"ca\",\"abc\") == 3 AND DL-Full(\"ca\",\"abc\") == 2 (the divergence); LevenshteinScore(\"abc\",\"abc\") == DamerauLevenshteinOSAScore(\"abc\",\"abc\") == DamerauLevenshteinFullScore(\"abc\",\"abc\") == HammingScore(\"abc\",\"abc\") == JaroScore(\"abc\",\"abc\") == JaroWinklerScore(\"abc\",\"abc\") == 1.0 (identity convergence); LevenshteinDistance(\"a\",\"b\") == DamerauLevenshteinOSADistance(\"a\",\"b\") == DamerauLevenshteinFullDistance(\"a\",\"b\") == HammingDistance(\"a\",\"b\") == 1 (single-character substitution agreement)"
    - "bench.txt is committed to repo root with the output of `make bench` after Wave 3 lands; this is the first benchstat baseline for v1.x"
    - "make check exits 0 (full quality gate including verify-determinism on the merged algorithms.json)"
    - "make verify-determinism exits 0 — the merged algorithms.json is byte-identical across the 5-platform CI matrix"
    - "Apache-2.0 header on every new .go file"
    - "examples/identifier-similarity/go.mod uses `replace github.com/axonops/fuzzymatch => ../..` and zero non-stdlib runtime require lines beyond fuzzymatch and (transitively) golang.org/x/text"
    - "All six per-algorithm _staging/*.json files are PRESERVED in the repo (they document the per-algorithm contribution and serve as the inputs to the merge step)"
  artifacts:
    - path: "examples/identifier-similarity/main.go"
      provides: "Runnable `package main` example demonstrating all 6 Phase 2 algorithms on database column-name pairs"
      min_lines: 80
      contains: "package main"
    - path: "examples/identifier-similarity/main_test.go"
      provides: "TestExample_Output meta-test capturing stdout and asserting byte-stable output"
      contains: "TestExample_Output"
    - path: "examples/identifier-similarity/go.mod"
      provides: "Module declaration for the example with `replace ../..` directive"
      contains: "replace github.com/axonops/fuzzymatch"
    - path: "cross_algorithm_consistency_test.go"
      provides: "Pins cross-algorithm divergence and convergence contracts in the root test suite"
      contains: "TestCrossAlgorithm_OSA_Full_Divergence"
    - path: "testdata/golden/algorithms.json"
      provides: "MERGED canonical golden file containing entries from all six algorithms"
      contains: "JaroWinkler_MARTHA_MARHTA"
    - path: "bench.txt"
      provides: "First benchstat baseline for v1.x — captures benchmark output across all 6 algorithms"
  key_links:
    - from: "testdata/golden/algorithms.json"
      to: "testdata/golden/_staging/{levenshtein,hamming,jaro,jarowinkler,damerau_osa,damerau_full}.json"
      via: "TestGolden_Algorithms_Merge reads staging files, sorts entries by Name, marshals via CanonicalMarshalForTest"
      pattern: "TestGolden_Algorithms_Merge"
    - from: "examples/identifier-similarity/main.go"
      to: "github.com/axonops/fuzzymatch"
      via: "import + 6 score function calls"
      pattern: "fuzzymatch\\.(Levenshtein|Hamming|Jaro|JaroWinkler|DamerauLevenshtein)"

user_setup: []
---

<objective>
Finalise Phase 2 by merging the six per-algorithm staging golden files into the canonical `testdata/golden/algorithms.json`, shipping the runnable `examples/identifier-similarity/` program with its meta-test, pinning the cross-algorithm divergence/convergence contracts in `cross_algorithm_consistency_test.go`, and committing the first `bench.txt` benchstat baseline.

Purpose: this Wave 3 plan owns three concerns that REQUIRE all six Wave 2 plans to have merged: (1) the algorithms.json merge (each Wave 2 plan wrote `_staging/<algo>.json`; this plan combines them into the canonical artefact and turns on the cross-platform CI determinism gate for the full algorithm set); (2) the identifier-similarity example which calls all six algorithms side-by-side and would not compile until all six existed; (3) the cross-algorithm consistency tests that verify OSA-vs-Full divergence and identity-vs-distance convergence across the catalogue.

Output: a runnable `examples/identifier-similarity/` program meta-tested for byte-stable stdout, a merged canonical `testdata/golden/algorithms.json` byte-stable across the CI matrix, a `cross_algorithm_consistency_test.go` test file pinning the load-bearing cross-algorithm contracts, and a committed `bench.txt` baseline.
</objective>

<execution_context>
@$HOME/.claude/get-shit-done/workflows/execute-plan.md
@$HOME/.claude/get-shit-done/templates/summary.md
</execution_context>

<context>
@.planning/phases/02-core-character-algorithms-six/02-CONTEXT.md
@.planning/phases/02-core-character-algorithms-six/02-RESEARCH.md
@.planning/phases/02-core-character-algorithms-six/02-PATTERNS.md
@.planning/phases/02-core-character-algorithms-six/02-VALIDATION.md
@.planning/phases/02-core-character-algorithms-six/02-01-levenshtein-SUMMARY.md
@.planning/phases/02-core-character-algorithms-six/02-02-hamming-SUMMARY.md
@.planning/phases/02-core-character-algorithms-six/02-03-jaro-SUMMARY.md
@.planning/phases/02-core-character-algorithms-six/02-04-jarowinkler-SUMMARY.md
@.planning/phases/02-core-character-algorithms-six/02-05-damerau-osa-SUMMARY.md
@.planning/phases/02-core-character-algorithms-six/02-06-damerau-full-SUMMARY.md
@docs/requirements.md
@CLAUDE.md
@.claude/skills/algorithm-correctness-standards/SKILL.md
@.claude/skills/determinism-standards/SKILL.md
@.claude/skills/go-testing-standards/SKILL.md

<interfaces>
This plan depends on ALL six Wave 1 + Wave 2 plans:

From Wave 1 (plan 02-01): LevenshteinDistance, LevenshteinDistanceRunes, LevenshteinScore, LevenshteinScoreRunes; testdata/golden/algorithms.json (Wave 1 form with Levenshtein-only entries — this plan REPLACES the file with the merged form)

From Wave 2:
  - HammingDistance, HammingScore (silent-zero unequal length)
  - JaroScore (no Distance variant)
  - JaroWinklerScore
  - DamerauLevenshteinOSADistance, DamerauLevenshteinOSAScore (discriminating vector ca/abc → 3)
  - DamerauLevenshteinFullDistance, DamerauLevenshteinFullScore (discriminating vector ca/abc → 2)

From the testdata/golden/_staging/ directory (created by Wave 2 plans):
  - testdata/golden/_staging/levenshtein.json   (Wave 1 plan 02-01 Task 3 — committed by plan 02-01 alongside the assertGoldenStaging helper)
  - testdata/golden/_staging/hamming.json       (Wave 2 plan 02-02)
  - testdata/golden/_staging/jaro.json          (Wave 2 plan 02-03)
  - testdata/golden/_staging/jarowinkler.json   (Wave 2 plan 02-04)
  - testdata/golden/_staging/damerau_osa.json   (Wave 2 plan 02-05)
  - testdata/golden/_staging/damerau_full.json  (Wave 2 plan 02-06)

  NOTE: Plan 02-01 Task 3 (revised) commits BOTH testdata/golden/algorithms.json (Levenshtein entries) AND testdata/golden/_staging/levenshtein.json so this Wave 3 plan's merge step has uniform inputs across all six algorithms — no Levenshtein special-case logic. The merge simply reads all six staging files and rewrites algorithms.json.

From algorithms_golden_test.go (Wave 1 + extended by each Wave 2 plan with TestGolden_<algo>_Staging helpers):
  - goldenAlgorithmEntry struct
  - goldenAlgorithmsFile struct
  - assertGolden_writeStaging helper (added by Wave 1 / Wave 2-Hamming as the staging-file write convention)

From CLAUDE.md (project guidelines): conventional commits without AI attribution; no `--no-verify`; releases via CI only.
</interfaces>
</context>

<tasks>

<task type="auto">
  <name>Task 1: Merge per-algorithm staging files into canonical testdata/golden/algorithms.json</name>
  <files>algorithms_golden_test.go, testdata/golden/algorithms.json</files>
  <read_first>
    - algorithms_golden_test.go (current state — Wave 1 + 5 Wave 2 extensions; understand the existing TestGolden_Algorithms function and the per-algorithm TestGolden_<algo>_Staging helpers)
    - testdata/golden/algorithms.json (Wave 1's Levenshtein-only contents — this file will be rewritten by this task)
    - testdata/golden/_staging/hamming.json, jaro.json, jarowinkler.json, damerau_osa.json, damerau_full.json (the five staging files Wave 2 plans created — read each to understand the exact entry shapes and counts)
    - golden_canonical.go (canonicalMarshal contract: 2-space indent, trailing LF, no BOM, deterministic key ordering)
    - golden_test.go (assertGolden + -update flag pattern)
    - export_test.go (CanonicalMarshalForTest re-export)
    - .planning/phases/02-core-character-algorithms-six/02-RESEARCH.md §Golden File Integration — Wave 2 staging strategy (the merge is THIS plan's responsibility)
  </read_first>
  <action>
1. Read each of the five Wave 2 staging files (hamming.json, jaro.json, jarowinkler.json, damerau_osa.json, damerau_full.json) and confirm they exist and are valid JSON matching the goldenAlgorithmsFile schema.

2. Confirm `_staging/levenshtein.json` exists. Plan 02-01 Task 3 (revised) committed it alongside the `assertGoldenStaging` helper — this Wave 3 plan does NOT need to create it. If for any reason it is missing, re-run plan 02-01's `TestGolden_Levenshtein_Staging -update` before proceeding; do NOT inline-build Levenshtein entries in this plan.

3. Refactor algorithms_golden_test.go's TestGolden_Algorithms (Wave 1's original) into a new `TestGolden_Algorithms_Merge`:

       // TestGolden_Algorithms_Merge reads all six per-algorithm staging
       // files from testdata/golden/_staging/ and merges them into the
       // canonical testdata/golden/algorithms.json via CanonicalMarshalForTest.
       // Wave 3 of phase 02 (plan 02-07) owns the merge step — Wave 2 plans
       // each wrote one staging file to avoid algorithms.json merge conflicts.
       func TestGolden_Algorithms_Merge(t *testing.T) {
           stagingFiles := []string{
               "_staging/damerau_full.json",
               "_staging/damerau_osa.json",
               "_staging/hamming.json",
               "_staging/jaro.json",
               "_staging/jarowinkler.json",
               "_staging/levenshtein.json",
           }
           var allEntries []goldenAlgorithmEntry
           for _, f := range stagingFiles {
               raw, err := os.ReadFile(filepath.Join("testdata/golden", f))
               if err != nil { t.Fatalf("read %s: %v", f, err) }
               var staged goldenAlgorithmsFile
               if err := json.Unmarshal(raw, &staged); err != nil { t.Fatalf("parse %s: %v", f, err) }
               allEntries = append(allEntries, staged.Entries...)
           }
           sort.Slice(allEntries, func(i, j int) bool { return allEntries[i].Name < allEntries[j].Name })
           // Verify no duplicate Names (sanity check across staging files):
           for i := 1; i < len(allEntries); i++ {
               if allEntries[i].Name == allEntries[i-1].Name {
                   t.Fatalf("duplicate entry Name across staging files: %q", allEntries[i].Name)
               }
           }
           file := goldenAlgorithmsFile{Version: 1, Entries: allEntries}
           assertGolden(t, "algorithms.json", file)
       }

4. Delete or rename the original TestGolden_Algorithms function from Wave 1 (it built only Levenshtein entries inline; the Merge form supersedes it). The function name `TestGolden_Algorithms_Merge` is the canonical name going forward.

5. Generate the merged file: `go test -run TestGolden_Algorithms_Merge -update -count=1 ./...`. Inspect the result: should contain entries from all six algorithms (4 Levenshtein + 4 Hamming + 6 Jaro + 8 JaroWinkler + 5 DamerauLevenshteinOSA + 5 DamerauLevenshteinFull = ~32 entries; exact count depends on each plan's exact entry choices). Sorted alphabetically by Name. Canonical form (2-space indent, trailing LF, no BOM).

6. Commit the regenerated `testdata/golden/algorithms.json`.

7. Run the merge test WITHOUT -update: `go test -run TestGolden_Algorithms_Merge -count=1 ./...`. Must exit 0 with zero diff. This is the deterministic-byte-form gate.

8. Run `make verify-determinism`. Must exit 0.

After this task, the Wave 1 stub `TestGolden_Algorithms` is replaced by `TestGolden_Algorithms_Merge`. The six staging files are preserved in the repo (they are the algorithm-by-algorithm contribution audit trail; future phases follow the same pattern).
  </action>
  <verify>
    <automated>go test -race -shuffle=on -count=1 -run 'TestGolden_Algorithms_Merge|TestGolden_.*_Staging' ./... && make verify-determinism</automated>
  </verify>
  <acceptance_criteria>
    - testdata/golden/_staging/levenshtein.json exists (created by plan 02-01 Task 3; this Wave 3 plan only verifies its presence).
    - testdata/golden/algorithms.json exists and contains entries from all six algorithms — verifiable by `jq '[.entries[].algorithm] | unique' testdata/golden/algorithms.json` returning a JSON array containing all six algorithm names: Levenshtein, Hamming, Jaro, JaroWinkler, DamerauLevenshteinOSA, DamerauLevenshteinFull.
    - The file is in canonical form (2-space indent, trailing LF byte 0x0a, no BOM, sorted by Name).
    - Re-running TestGolden_Algorithms_Merge WITHOUT -update produces zero diff (the merged file is byte-stable across re-runs).
    - `make verify-determinism` exits 0.
    - All six _staging/*.json files are still present (NOT deleted by the merge — they remain as the audit trail).
    - No duplicate entry Names across staging files (the merge test asserts this; if it fails, the executor must reconcile the duplicate Names by renaming entries in the offending staging files via -update on the staging tests).
  </acceptance_criteria>
  <done>
    Merged algorithms.json committed and byte-stable. _staging/levenshtein.json (or fallback) committed. TestGolden_Algorithms_Merge replaces the Wave 1 TestGolden_Algorithms function. make verify-determinism green.
  </done>
</task>

<task type="auto" tdd="true">
  <name>Task 2: Cross-algorithm consistency test (divergence + convergence + length-1-substitution agreement)</name>
  <files>cross_algorithm_consistency_test.go</files>
  <read_first>
    - levenshtein.go, hamming.go, jaro.go, jarowinkler.go, damerau_osa.go, damerau_full.go (the six algorithm files; understand the public API of each)
    - levenshtein_test.go (Wave 1 test pattern; all algorithms' tests exist as sibling references)
    - .planning/phases/02-core-character-algorithms-six/02-CONTEXT.md (Deferred Items: cross-algorithm consistency test was flagged as planner discretion; this plan includes it per the planning decision in the planning context)
    - .planning/phases/02-core-character-algorithms-six/02-RESEARCH.md §DL-OSA vs DL-Full Divergence; §Mathematical Invariants (the convergence claims for identity)
    - ROADMAP.md Phase 2 Success Criterion #2 (the OSA-vs-Full discriminating-vector contract — this test pins it cross-algorithm)
  </read_first>
  <action>
Create `cross_algorithm_consistency_test.go` (package fuzzymatch_test, stdlib testing only):

1. Apache-2.0 header copied from normalise.go lines 1-13.
2. File-level godoc explaining this file pins cross-algorithm contracts that span the entire Phase 2 catalogue:

       // cross_algorithm_consistency_test.go pins relationships across the six
       // Phase 2 character-based algorithms:
       //
       //   - DIVERGENCE: DL-OSA and DL-Full produce DIFFERENT distances for
       //     the canonical discriminating vector "ca"/"abc" (3 vs 2). This
       //     is a load-bearing claim of the phase (ROADMAP success criterion #2);
       //     the file pins it cross-algorithm rather than only inside each
       //     algorithm's own _test.go file.
       //
       //   - CONVERGENCE: every algorithm's score on identical inputs is
       //     exactly 1.0 (identity).
       //
       //   - SINGLE-SUBSTITUTION AGREEMENT: Levenshtein, DL-OSA, DL-Full,
       //     and Hamming all return distance == 1 for a single-character
       //     substitution between equal-length strings (e.g. "a"/"b").
       //
       // Stdlib `testing` only — no testify in root tests, per
       // .claude/skills/go-coding-standards.

3. Tests:

   - TestCrossAlgorithm_OSA_Full_Divergence:
       Asserts DamerauLevenshteinOSADistance("ca", "abc") == 3 AND DamerauLevenshteinFullDistance("ca", "abc") == 2 in the SAME test; document in the comment that this divergence is the canonical proof of the algorithm distinction (Boytsov 2011 §3.1 for OSA, Lowrance-Wagner 1975 for Full).

   - TestCrossAlgorithm_IdentityConvergence:
       For a non-empty input "abc", asserts ALL six algorithms return Score(input, input) == 1.0:
         - LevenshteinScore("abc", "abc") == 1.0
         - DamerauLevenshteinOSAScore("abc", "abc") == 1.0
         - DamerauLevenshteinFullScore("abc", "abc") == 1.0
         - HammingScore("abc", "abc") == 1.0
         - JaroScore("abc", "abc") == 1.0
         - JaroWinklerScore("abc", "abc") == 1.0
       Use a slice of `{name, fn}` and a single loop with sub-tests via `t.Run`.

   - TestCrossAlgorithm_BothEmptyConvergence:
       For both-empty inputs ("", ""), all six algorithms return Score(input, input) == 1.0. Same loop pattern.

   - TestCrossAlgorithm_SingleSubstitution_DistanceAgreement:
       For "a"/"b" (single-character substitution between equal-length strings), Levenshtein, DL-OSA, DL-Full, and Hamming all return Distance == 1. (Jaro and JW have no Distance variant; they return Score values that are NOT directly comparable to distances — exclude them from this test.)

   - TestCrossAlgorithm_OneEmpty_ScoreAgreement:
       For ""/"abc", all six algorithms return Score == 0.0. Same loop pattern.

4. Use stdlib `math.Abs` for any float tolerance (none of the convergence tests need tolerance — they assert exact equality with 1.0 or 0.0).

5. Run: `go test -race -shuffle=on -count=1 -run 'TestCrossAlgorithm' ./...`. Must exit 0.
  </action>
  <verify>
    <automated>go test -race -shuffle=on -count=1 -run 'TestCrossAlgorithm' ./...</automated>
  </verify>
  <acceptance_criteria>
    - cross_algorithm_consistency_test.go starts with the Apache-2.0 header.
    - All five TestCrossAlgorithm_* tests pass.
    - TestCrossAlgorithm_OSA_Full_Divergence references both expected values (3 for OSA, 2 for Full) explicitly in the test body — verifiable by `grep -E 'OSA.*== 3|Full.*== 2' cross_algorithm_consistency_test.go` returning at least one match each.
    - `grep -c '"github.com/stretchr/testify' cross_algorithm_consistency_test.go` returns 0.
    - The file is in package fuzzymatch_test (NOT fuzzymatch) — uses public API only.
    - **PERF-03 deviation gate (W-3):** Confirm plan 02-06's PERF-03 disposition by reading 02-06-damerau-full-SUMMARY.md. If the fallback path was taken (full DP table allocated for any input regime — e.g. the rune mode), open a GitHub issue, add it to KNOWN_ISSUES.md (or the project's equivalent tracking file; if neither exists, create the issue and reference it in 02-SUMMARY.md), and reference the issue number in 02-SUMMARY.md before the phase is declared shippable. PERF-03 is BLOCKING for two-row DP family algorithms — do NOT proceed to Phase 3 with an unmitigated DL-Full PERF-03 deviation unless the issue is tracked. The audit-trail SUMMARY note alone is insufficient; an issue link is required.
  </acceptance_criteria>
  <behavior>
    - DL-OSA vs DL-Full divergence pinned in a single test that cannot pass if either algorithm's recurrence is wrong.
    - Identity convergence pinned across all six algorithms.
    - Both-empty convergence pinned (the documented project-wide convention from RESEARCH.md §Score Normalisation).
    - Single-substitution distance agreement pinned for the four distance-bearing algorithms.
    - One-empty score agreement pinned across all six.
  </behavior>
  <done>
    cross_algorithm_consistency_test.go committed; all five cross-algorithm contracts pass.
  </done>
</task>

<task type="auto" tdd="true">
  <name>Task 3: identifier-similarity example (main + meta-test) + first bench.txt baseline</name>
  <files>examples/identifier-similarity/main.go, examples/identifier-similarity/main_test.go, examples/identifier-similarity/go.mod, examples/identifier-similarity/go.sum, bench.txt</files>
  <read_first>
    - .planning/phases/02-core-character-algorithms-six/02-CONTEXT.md (identifier-similarity example LOCKED specification: 6-10 hardcoded pairs, 6 algorithm columns, plaintext table, 4-decimal scores, meta-test capturing stdout)
    - .planning/phases/02-core-character-algorithms-six/02-PATTERNS.md (Pattern 14 godoc example structure; Pattern 16 examples meta-test pattern from `os/exec` + stdout-capture)
    - .planning/phases/02-core-character-algorithms-six/02-RESEARCH.md §File Layout — examples/identifier-similarity row
    - levenshtein.go, hamming.go, jaro.go, jarowinkler.go, damerau_osa.go, damerau_full.go (the six public APIs the example calls)
    - tests/bdd/go.mod (reference for `replace github.com/axonops/fuzzymatch => ../..` directive structure)
    - Makefile (find the `bench` target; understand what `make bench` produces and how it writes/reads bench.txt; if a `bench-update` or similar target exists for committing baselines, use it)
    - .planning/phases/01-foundation-infrastructure/01-02-quality-gates-SUMMARY.md (find any prior decisions about bench.txt format or location)
  </read_first>
  <action>
**Step A — Create the example program:**

Create `examples/identifier-similarity/main.go` (package main):

1. Apache-2.0 header copied from normalise.go lines 1-13.
2. File-level godoc explaining the program demonstrates all six Phase 2 algorithms side-by-side on database column-name pairs. Reference CONTEXT.md for the chosen pairs and the rationale (semantic equivalence detection in axonops/audit).
3. Imports: `fmt` and `github.com/axonops/fuzzymatch`.
4. Hardcoded pairs slice (in the EXACT order from CONTEXT.md):

       var pairs = []struct{ a, b string }{
           {"user_id",     "userId"},
           {"created_at",  "creationTimestamp"},
           {"status",      "state"},
           {"email",       "e_mail"},
           {"org_id",      "organisation_id"},
           {"latitude",    "longitude"},
           {"is_deleted",  "is_active"},
       }

5. Algorithm-column slice with Score function pointers (all six):

       var algorithms = []struct{
           name string
           fn   func(a, b string) float64
       }{
           {"Levenshtein",  fuzzymatch.LevenshteinScore},
           {"DL-OSA",       fuzzymatch.DamerauLevenshteinOSAScore},
           {"DL-Full",      fuzzymatch.DamerauLevenshteinFullScore},
           {"Hamming",      fuzzymatch.HammingScore},
           {"Jaro",         fuzzymatch.JaroScore},
           {"Jaro-Winkler", fuzzymatch.JaroWinklerScore},
       }

6. main() function:
     - Print a header row with column names, fixed-width formatting (e.g. `%-30s` for the pair column, `%12s` for each algorithm column).
     - For each pair, compute all six scores and print a row. Use `fmt.Sprintf("%.4f", score)` for cell values (4-decimal precision per CONTEXT.md). Hamming on length-mismatched pairs returns 0.0000 (per the LOCKED silent-zero policy from plan 02-02 — NOT `ERR`).

     **Documentation supersession note (W-2 — include this exact comment in main.go's file-level godoc):**

       // Note: CONTEXT.md `<deferred>` identifier-similarity format spec'd
       // `ERR` for Hamming length-mismatch BEFORE the Hamming silent-zero
       // policy was locked (commit 1e25e31). The locked Hamming policy
       // supersedes that earlier illustrative format — the example shows
       // `0.0000` and never `ERR`. This resolution is a documentation
       // supersession, not a scope reduction.
     - Output is plaintext, deterministic, byte-stable.

Create `examples/identifier-similarity/main_test.go` (package main):

1. Apache-2.0 header.
2. File-level godoc describing the meta-test purpose: pin the example's stdout byte-for-byte across runs and platforms (extension of the project-wide cross-platform determinism gate).
3. TestExample_Output:
     - Use `os/exec` to run `go run .` from the package directory (or use `runtime/debug` to capture the current binary; `os/exec` is simpler and matches Pattern 16).
     - Capture stdout into a bytes.Buffer.
     - Assert the captured output equals a `const want = ...` string committed in the test file.
     - On first run: leave `want` empty AND fail with a t.Errorf showing the captured output, so the executor inspects it and copies it into `want`. After committing the captured output, the test must pass on re-runs.
     - Alternative simpler implementation: use the `golden_test.go` `assertGolden` helper to write the expected output to a file like `testdata/example_identifier_similarity.txt` and compare the captured stdout against that file. This is cleaner than embedding a multi-line `want` constant. Choose the approach that matches the existing project conventions; if assertGolden is sufficient, prefer it (less in-source noise).

Create `examples/identifier-similarity/go.mod`:

       module github.com/axonops/fuzzymatch/examples/identifier-similarity

       go 1.26

       require github.com/axonops/fuzzymatch v0.0.0

       replace github.com/axonops/fuzzymatch => ../..

Run `cd examples/identifier-similarity && go mod tidy` to populate go.sum. Confirm zero non-stdlib runtime require lines beyond fuzzymatch (and golang.org/x/text transitively).

Run `cd examples/identifier-similarity && go test -count=1 ./...` to confirm the meta-test passes.

Run `cd examples/identifier-similarity && go run .` to inspect the table output manually. Confirm it is readable, includes all 7 pair rows and 6 algorithm columns, and shows expected behaviour (Hamming on length-mismatched rows is 0.0000).

**Step B — Commit the first bench.txt baseline:**

1. Run `make bench` from the repo root. This should produce benchmark output for all six algorithms' `BenchmarkXxxScore_ASCII_Short/Medium/Long/Unicode_Short` series.
2. Capture the output into `bench.txt` at the repo root. The exact format depends on the Makefile target — typically `go test -bench=. -benchmem -run=^$ -count=10 ./... | tee bench.txt`. If the Makefile has a dedicated `bench-baseline` or similar target, use it.
3. Commit `bench.txt` to the repo root.
4. The Phase 1 plan 01-02 Summary may already document the bench.txt convention — read it before deciding the exact contents.
5. Run `make bench-compare` to confirm the just-committed baseline produces no regression on a re-run (informational; a fresh baseline by definition has zero diff).

**Step C — Final phase-wide quality gate:**

Run `make check`. Must exit 0. This is the load-bearing pre-commit gate for Phase 2.
  </action>
  <verify>
    <automated>(cd examples/identifier-similarity && go mod tidy && go test -race -count=1 ./...) && go test -race -shuffle=on -count=1 -run 'TestGolden_Algorithms_Merge|TestCrossAlgorithm' ./... && make check</automated>
  </verify>
  <acceptance_criteria>
    - examples/identifier-similarity/main.go exists, is package main, imports fuzzymatch, and prints a 7-row × 6-column table to stdout.
    - examples/identifier-similarity/main_test.go's TestExample_Output passes: the captured stdout matches the committed expected output byte-for-byte.
    - examples/identifier-similarity/go.mod has `replace github.com/axonops/fuzzymatch => ../..` and zero non-stdlib runtime require lines beyond fuzzymatch.
    - `cd examples/identifier-similarity && go test -count=1 ./...` exits 0.
    - `cd examples/identifier-similarity && go run .` exits 0 and prints the table.
    - bench.txt exists at repo root and contains benchmark output for all six algorithms (verifiable by `grep -c 'BenchmarkLevenshteinScore\|BenchmarkHammingScore\|BenchmarkJaroScore\|BenchmarkJaroWinklerScore\|BenchmarkDamerauLevenshteinOSAScore\|BenchmarkDamerauLevenshteinFullScore' bench.txt` returning a count > 0 for each algorithm — at least 6 lines minimum).
    - `make check` exits 0 (full quality gate green for Phase 2).
    - `bash scripts/verify-license-headers.sh` exits 0 (the new example .go files carry the Apache-2.0 header).
    - `bash scripts/verify-no-runtime-deps.sh` exits 0 (root go.mod still has only fuzzymatch + golang.org/x/text).
  </acceptance_criteria>
  <behavior>
    - The example program is a runnable, deterministic demonstration of all six Phase 2 algorithms on real-world database identifier pairs.
    - The meta-test pins the example's stdout byte-stable across runs and platforms.
    - bench.txt is the first committed benchstat baseline; future PRs run `make bench-compare` against it to detect regressions.
    - The full Phase 2 quality gate (lint, vet, race, coverage, license headers, no-runtime-deps, tidy-check, verify-determinism, BDD) passes via `make check`.
  </behavior>
  <done>
    Example program + meta-test committed; example go.mod with replace directive committed; bench.txt baseline committed; make check green. Phase 2 is shippable.
  </done>
</task>

</tasks>

<threat_model>
## Trust Boundaries

| Boundary | Description |
|----------|-------------|
| Caller → fuzzymatch.* algorithms (six total) | Same trust boundaries as Wave 1 + Wave 2 plans; this Wave 3 plan introduces no new untrusted-input paths. The example program uses hardcoded inputs. |
| go test → testdata/golden/_staging/*.json | Read-only; the merge test parses the staging files. JSON parse errors fail the test loudly (no silent corruption). |

## STRIDE Threat Register

| Threat ID | Category | Component | Disposition | Mitigation Plan |
|-----------|----------|-----------|-------------|-----------------|
| T-02-07-01 | Tampering | A staging file is corrupted between Wave 2 and Wave 3 (e.g. someone hand-edits hamming.json) | mitigate | TestGolden_Algorithms_Merge re-parses each staging file via encoding/json; corruption causes the test to fail fast with the file name in the error message. The duplicate-Name check catches accidental cross-staging-file collisions. PR review enforces correctness before merge. |
| T-02-07-02 | Tampering | The merged algorithms.json is hand-edited to weaken the cross-platform determinism gate | mitigate | The file is regenerated from staging files on every -update run; any hand-edit is overwritten. CI matrix runs `make verify-determinism` on five platforms; any inconsistency surfaces. |
| T-02-07-03 | Information Disclosure | The example's stdout meta-test fails on a CI platform due to non-deterministic float output | mitigate | The six algorithms are deterministic (DET-02 / DET-04 enforced per algorithm). Float formatting via `fmt.Sprintf("%.4f", score)` is platform-stable for the values in [0,1] arising from these algorithms (Go's stdlib uses strconv.AppendFloat with deterministic rounding). If any platform reports a diff, the underlying bug is in the algorithm — fix at the algorithm level, not by relaxing the meta-test. |
| T-02-07-04 | Denial of Service | bench.txt grows without bound across many releases | accept | bench.txt is overwritten on each baseline update (not appended to). Long-term, benchstat consumes the file via `make bench-compare` for diff comparison; the file remains O(per-algorithm-benchmark-count) in size. |
| T-02-07-05 | Spoofing | Wave 3's algorithms.json drift makes axonops/audit (Phase 11 consumer) get different scores than expected | mitigate | DET-02 (algorithm score stability across patch releases) is enforced via the golden file gate. Any score change requires a minor version bump per REL-07. The cross_algorithm_consistency_test.go gate makes drift in identity / divergence / convergence behaviour visible immediately. |

No high-severity items. Plan passes the security gate.
</threat_model>

<verification>
1. `go build ./...` exits 0 (root module).
2. `(cd examples/identifier-similarity && go build ./...)` exits 0 (example module).
3. `(cd examples/identifier-similarity && go mod tidy)` exits 0 with no diff after the first run.
4. `go vet ./...` exits 0.
5. `bash scripts/verify-license-headers.sh` exits 0 (covers all new .go files).
6. `bash scripts/verify-no-runtime-deps.sh` exits 0 (root go.mod still allowlist-clean).
7. `go test -race -shuffle=on -count=1 -run 'TestCrossAlgorithm|TestGolden_Algorithms_Merge|TestGolden_.*_Staging' ./...` exits 0.
8. `(cd examples/identifier-similarity && go test -race -count=1 ./...)` exits 0.
9. `(cd tests/bdd && go test -race -shuffle=on -count=1 ./...)` exits 0 (all six algorithms' BDD scenarios pass).
10. `make verify-determinism` exits 0 (the merged algorithms.json byte-stable; Hamming, Jaro, JW, OSA, Full, Levenshtein _staging files all byte-stable).
11. bench.txt exists and contains benchmarks for all six algorithms.
12. `make check` exits 0 (full quality gate).
</verification>

<success_criteria>
- testdata/golden/algorithms.json contains entries from all six Phase 2 algorithms, byte-stable, cross-platform-identical (DET-02 fully satisfied for Phase 2).
- examples/identifier-similarity/ ships a runnable demonstration of the entire Phase 2 surface, meta-tested for stdout stability (DX-05 satisfied).
- cross_algorithm_consistency_test.go pins the OSA-vs-Full divergence and identity-vs-distance convergence contracts in a single audit-trail-friendly file.
- bench.txt is committed as the first benchstat baseline; future PRs detect regressions via `make bench-compare`.
- All required gates (license headers, no-runtime-deps, lint, vet, race, tidy, coverage, BDD, determinism) pass via `make check`.
- ROADMAP success criterion #5 satisfied: cross-platform golden file `algorithms.json` contains pinned scores for all six algorithms and diffs byte-identically across the CI matrix; first `bench.txt` committed with benchstat baseline; example program runs and is meta-tested.
- Phase 2 is shippable (ready for `/gsd-verify-work`).
</success_criteria>

<output>
After completion, create `.planning/phases/02-core-character-algorithms-six/02-07-finalisation-SUMMARY.md` recording:
- The exact entry count in the merged algorithms.json (sum of per-staging-file entries).
- Which staging-file path was used for Levenshtein (preferred = synthesised _staging/levenshtein.json; fallback = inline in the merge test).
- The captured stdout from `examples/identifier-similarity/` as a verbatim block (or a reference to the testdata file if assertGolden was used) — for traceability of the meta-test's expected value.
- Benchmark numbers from the committed bench.txt (summary table: algorithm × ASCII-Short ns/op).
- Coverage percentages.
- Any deviations from the plan and their rationale.
- A self-check confirming the cross-algorithm consistency tests AND the merged golden file AND the example meta-test all pass on the local machine before declaring the phase shippable.
</output>
