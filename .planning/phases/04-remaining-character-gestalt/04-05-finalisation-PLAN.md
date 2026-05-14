---
phase: 04-remaining-character-gestalt
plan: 05
type: execute
wave: 5
depends_on:
  - 04-04-ratcliff-obershelp-cross-validation
files_modified:
  - testdata/golden/algorithms.json
  - algorithms_golden_test.go
  - cross_algorithm_consistency_test.go
  - examples/identifier-similarity/main.go
  - examples/identifier-similarity/main_test.go
  - llms.txt
  - llms-full.txt
  - bench.txt
autonomous: true
requirements:
  - CHAR-07
  - CHAR-09
  - GESTALT-01
tags: [finalisation, golden-merge, cross-algorithm-consistency, identifier-similarity-example, llms-sync, bench-baseline, ro-asymmetry-pin, strcmp95-jaro-winkler-hierarchy, lcsstr-levenshtein-substring-containment, ci-matrix-determinism]

must_haves:
  truths:
    # Goal-backward truths (CHAR-07 + CHAR-09 + GESTALT-01 cross-cutting; ROADMAP success criteria #4)
    - "testdata/golden/algorithms.json contains entries for Strcmp95, LCSStr, and RatcliffObershelp — merged from the three staging files committed in plans 04-01, 04-02, 04-03 via the existing TestGolden_Algorithms_Merge -update flow"
    - "`make verify-determinism` exits 0 across the cross-platform CI matrix (linux/amd64, linux/arm64, darwin/arm64, windows/amd64) — testdata/golden/algorithms.json is byte-identical on every platform"
    - "TestCrossAlgorithm_Strcmp95_AtLeastJaroWinkler asserts the algorithm hierarchy invariant on hand-pinned pairs (MARTHA/MARHTA, DWAYNE/DUANE, DIXON/DICKSONX): Strcmp95Score(a, b) >= JaroWinklerScore(a, b) — adjustments only ADD"
    - "TestCrossAlgorithm_LCSStr_AtLeastLevenshtein_SubstringContainment asserts LCSStrScore(\"http_request\", \"http_request_header_fields\") >= LevenshteinScore(\"http_request\", \"http_request_header_fields\") — LCSStr finds the contained substring; Levenshtein pays the deletion cost"
    - "TestCrossAlgorithm_RatcliffObershelp_PinnedDrDobbs asserts RatcliffObershelpScore(\"WIKIMEDIA\", \"WIKIMANIA\") within 1e-9 of the pinned difflib(autojunk=False) value (≈ 0.7777777777777778). Numerical regression pin OUTSIDE the cross-validation corpus per Phase 3 WR-03 closure"
    - "TestCrossAlgorithm_RatcliffObershelp_AsymmetryPin asserts RatcliffObershelpScore(\"tide\", \"diet\") != RatcliffObershelpScore(\"diet\", \"tide\") — the OQ-1 resolution asymmetry pin, INVERSE-form regression guard (asserts INEQUALITY rather than ordering)"
    - "TestCrossAlgorithm_RatcliffObershelp_PinnedAgainstDifflib asserts byte-for-byte difflib agreement on a pair where RO visibly differs from both Levenshtein and Jaro-Winkler — divergence-pin proving RO is the difflib-equivalent and not a near-clone of another algorithm"
    # Example program extension (DX-05)
    - "examples/identifier-similarity/main.go's `algorithms` slice grows from 7 entries to 10 entries; new rows for Strcmp95 (column label \"Strcmp95\"), LCSStr (column label \"LCSStr\"), Ratcliff-Obershelp (column label \"RO\" — truncated per PATTERNS.md to fit the existing algoWidth)"
    - "examples/identifier-similarity/main_test.go's `want` constant is regenerated via `go run .` and committed; `(cd examples/identifier-similarity && go test ./...)` exits 0; the TestExample_ColumnWidths test still passes"
    - "examples/identifier-similarity/main_test.go uses the Phase 3 WR-04 defer-restore os.Stdout pattern (no changes to the test logic — only the `want` constant changes)"
    # llms.txt + llms-full.txt (DX-03 + meta-test gate)
    - "llms.txt lists all 7 new exported symbols: Strcmp95Score, LongestCommonSubstring, LongestCommonSubstringRunes, LCSStrScore, LCSStrScoreRunes, RatcliffObershelpScore, RatcliffObershelpScoreRunes. The 3 AlgoID constants (AlgoStrcmp95, AlgoLCSStr, AlgoRatcliffObershelp) are ALREADY listed (declared in Phase 1) — no new AlgoID entries needed"
    - "llms-full.txt has parallel entries with one-line rationales for each of the 7 new symbols"
    - "ai_friendly_test.go::TestLLMs_PublicSymbolsListed passes — the meta-test parses the package AST and asserts every exported symbol is listed; missing any of the 7 new symbols fails the test"
    # bench.txt baseline (PERF-04)
    - "bench.txt is FULL-REPLACED via `make bench` from the reference benchmark hardware; the new file contains Phase 4's benchmark rows alongside Phase 2 + 3 rows; `make bench-compare` accepts the new baseline"
    - "All Phase 4 bench-rows are present: BenchmarkStrcmp95Score_{ASCII_Short,ASCII_Medium,ASCII_Long}; BenchmarkLCSStrScore_{ASCII_Short,ASCII_Medium,ASCII_Long,Unicode_Short}; BenchmarkLongestCommonSubstring_{ASCII_Short,ASCII_Medium,ASCII_Long,Unicode_Short}; BenchmarkLCSStrScoreRunes_Unicode_Short; BenchmarkLongestCommonSubstringRunes_Unicode_Short; BenchmarkRatcliffObershelpScore_{ASCII_Short,ASCII_Medium,ASCII_Long,Unicode_Short}; BenchmarkRatcliffObershelpScoreRunes_Unicode_Short"
    # Phase gate (DET-01 + CI-06)
    - "`make check` exits 0 at end of plan (golangci-lint v2 + go vet + go test -race -shuffle=on + coverage + license + deps + tidy + security)"
    - "`make test-bdd` exits 0"
    - "`make verify-determinism` exits 0 — testdata/golden/algorithms.json byte-stable on CI matrix"
    - "All Phase 4 requirement IDs (CHAR-07, CHAR-09, GESTALT-01) marked complete in ROADMAP traceability"
  artifacts:
    - path: "testdata/golden/algorithms.json"
      provides: "Canonical multi-algorithm golden file with Strcmp95 + LCSStr + RatcliffObershelp entries merged from the three Phase 4 staging files; sorted alphabetically; canonical-marshalled via CanonicalMarshalForTest"
      contains: "Strcmp95"
    - path: "algorithms_golden_test.go"
      provides: "Extended TestGolden_Algorithms_Merge stagingFiles slice now includes _staging/strcmp95.json, _staging/lcsstr.json, _staging/ratcliff_obershelp.json alongside the existing 7 SWG/Phase-2 staging files"
    - path: "cross_algorithm_consistency_test.go"
      provides: "Appended 4 new cross-algorithm tests: TestCrossAlgorithm_Strcmp95_AtLeastJaroWinkler, TestCrossAlgorithm_LCSStr_AtLeastLevenshtein_SubstringContainment, TestCrossAlgorithm_RatcliffObershelp_PinnedDrDobbs / TestCrossAlgorithm_RatcliffObershelp_PinnedAgainstDifflib, TestCrossAlgorithm_RatcliffObershelp_AsymmetryPin"
    - path: "examples/identifier-similarity/main.go"
      provides: "Extended algorithms slice 7 to 10 entries (Strcmp95, LCSStr, RO); short labels chosen to keep algoWidth compatible per PATTERNS.md gotcha note"
    - path: "examples/identifier-similarity/main_test.go"
      provides: "Regenerated `want` constant (one `go run .` capture); TestExample_ColumnWidths still passes; defer-restore os.Stdout pattern unchanged from Phase 3 WR-04"
    - path: "llms.txt"
      provides: "Appended 7 new exported-symbol entries (one per Phase 4 public function) under the catalogue section"
    - path: "llms-full.txt"
      provides: "Appended 7 parallel entries with one-line rationales"
    - path: "bench.txt"
      provides: "Full-replaced via `make bench` on the reference hardware; includes Phase 4 bench rows alongside Phase 2 + 3 rows; benchstat baseline accepts"
  key_links:
    - from: "algorithms_golden_test.go (TestGolden_Algorithms_Merge stagingFiles slice)"
      to: "testdata/golden/_staging/{strcmp95,lcsstr,ratcliff_obershelp}.json (committed in plans 04-01, 04-02, 04-03)"
      via: "Slice extension at the lines 163–171 area; running `go test -run TestGolden_Algorithms_Merge -update ./...` materialises the merged algorithms.json"
      pattern: "_staging/(strcmp95|lcsstr|ratcliff_obershelp)\\.json"
    - from: "cross_algorithm_consistency_test.go (TestCrossAlgorithm_RatcliffObershelp_AsymmetryPin)"
      to: "ratcliff_obershelp.go (asymmetric-by-design godoc paragraph per OQ-1 resolution)"
      via: "Asserts INEQUALITY RatcliffObershelpScore(tide, diet) != RatcliffObershelpScore(diet, tide); inverse-form regression guard for the OQ-1 resolution locked 2026-05-14"
      pattern: "tide.*diet|AsymmetryPin"
    - from: "examples/identifier-similarity/main.go (algorithms slice)"
      to: "examples/identifier-similarity/main_test.go (want constant)"
      via: "After editing main.go's slice, capture stdout from `go run .` and paste into main_test.go's want constant — byte-identical match required by TestExample_ColumnWidths"
      pattern: "var algorithms = \\[\\]struct"
    - from: "llms.txt (catalogue section appended)"
      to: "ai_friendly_test.go::TestLLMs_PublicSymbolsListed (meta-test gate)"
      via: "AST-based meta-test asserts every exported symbol in fuzzymatch package is listed in llms.txt"
      pattern: "Strcmp95Score|LongestCommonSubstring|LCSStrScore|RatcliffObershelp"
---

<objective>
Finalise Phase 4 — merge the three per-algorithm staging goldens into testdata/golden/algorithms.json; append four cross-algorithm consistency tests (Strcmp95 ≥ JaroWinkler hierarchy; LCSStr ≥ Levenshtein on substring containment; RatcliffObershelp pinned vs difflib on a divergence-pair; RatcliffObershelp asymmetry-pin per OQ-1 resolution); extend the identifier-similarity example program from 7 to 10 algorithm columns; sync llms.txt + llms-full.txt with the 7 new exported symbols; full-replace bench.txt via `make bench` to baseline Phase 4. Run `make check`, `make test-bdd`, and `make verify-determinism` to confirm the phase ships green.

Purpose: Phase 4 closes; the Strcmp95 / LCSStr / Ratcliff-Obershelp public surface is exposed end-to-end (golden, example, llms.txt, bench baseline), the algorithm-hierarchy invariants are pinned, and the OQ-1 resolution is locked by an inverse-form regression test. After this plan lands, Phase 5 (Q-gram algorithms) can begin.

Output: 8 modified files (1 merged canonical golden, 7 extensions to existing append-only files). NO new source files — this is the integration / finalisation plan.
</objective>

<execution_context>
@$HOME/.claude/get-shit-done/workflows/execute-plan.md
@$HOME/.claude/get-shit-done/templates/summary.md
</execution_context>

<context>
@.planning/PROJECT.md
@.planning/ROADMAP.md
@.planning/STATE.md
@.planning/REQUIREMENTS.md
@.planning/phases/04-remaining-character-gestalt/04-CONTEXT.md
@.planning/phases/04-remaining-character-gestalt/04-RESEARCH.md
@.planning/phases/04-remaining-character-gestalt/04-PATTERNS.md
@.planning/phases/04-remaining-character-gestalt/04-VALIDATION.md
@.planning/phases/04-remaining-character-gestalt/04-01-strcmp95-PLAN.md
@.planning/phases/04-remaining-character-gestalt/04-02-lcsstr-PLAN.md
@.planning/phases/04-remaining-character-gestalt/04-03-ratcliff-obershelp-PLAN.md
@.planning/phases/04-remaining-character-gestalt/04-04-ratcliff-obershelp-cross-validation-PLAN.md
@.planning/phases/02-core-character-algorithms-six/02-07-finalisation-SUMMARY.md
@.planning/phases/03-smith-waterman-gotoh/03-03-swg-finalisation-SUMMARY.md
@.claude/skills/determinism-standards/SKILL.md
@.claude/skills/performance-standards/SKILL.md
@.claude/skills/go-testing-standards/SKILL.md
@cross_algorithm_consistency_test.go
@examples/identifier-similarity/main.go
@examples/identifier-similarity/main_test.go
@llms.txt
@bench.txt
</context>

<interfaces>
<!-- Reference shape; executor extends existing append-only files. -->

From algorithms_golden_test.go (extension point — lines 162–195 area; the TestGolden_Algorithms_Merge stagingFiles slice):
A `[]string` literal listing the existing 7 staging-file paths (damerau_full, damerau_osa, hamming, jaro, jarowinkler, levenshtein, swg). Phase 4 appends three new entries: `_staging/strcmp95.json`, `_staging/lcsstr.json`, `_staging/ratcliff_obershelp.json`.

From cross_algorithm_consistency_test.go (extension point — append to bottom of file):
Existing template: `TestCrossAlgorithm_SWG_Levenshtein_SubstringDivergence` (lines 205–226 area) — assert ordering on a hand-curated pair.
Inverse-form template for the asymmetry pin: assert INEQUALITY rather than ordering.

From examples/identifier-similarity/main.go (extension point — algorithms slice around lines 73–84):
Existing slice has 7 entries: {Levenshtein, DL-OSA, DL-Full, Hamming, Jaro, Jaro-Winkler, SWG}. Phase 4 appends 3 entries: {Strcmp95, LCSStr, RO}. The `algoWidth` constant (around line 89) governs column width — Phase 4's labels MUST fit; RO is the short label per PATTERNS.md.

From llms.txt (catalogue section — Phase 4 adds 7 function entries):
Existing format: one bullet per public function, with the signature inline. The 3 AlgoID constants are ALREADY listed (declared in Phase 1).

Bench.txt: NO per-row analog — the file IS the artefact. Regeneration is `make bench` then commit the resulting `bench.txt`.
</interfaces>

<tasks>

<task type="auto" tdd="true">
  <name>Task 1: Merge staging goldens + extend cross-algorithm consistency tests</name>
  <files>testdata/golden/algorithms.json, algorithms_golden_test.go, cross_algorithm_consistency_test.go</files>
  <read_first>
    - algorithms_golden_test.go lines 162–195 (TestGolden_Algorithms_Merge — current stagingFiles slice and merge logic; identify the exact slice to extend)
    - cross_algorithm_consistency_test.go (full file, especially the existing TestCrossAlgorithm_SWG_Levenshtein_SubstringDivergence template around lines 205–226)
    - testdata/golden/_staging/strcmp95.json (committed in plan 04-01)
    - testdata/golden/_staging/lcsstr.json (committed in plan 04-02)
    - testdata/golden/_staging/ratcliff_obershelp.json (committed in plan 04-03)
    - testdata/golden/algorithms.json (current state — read to understand the canonical schema and the existing entries)
    - .planning/phases/04-remaining-character-gestalt/04-CONTEXT.md (specifically the cross-algorithm consistency requirements and the OQ-1 resolution)
    - .planning/phases/04-remaining-character-gestalt/04-PATTERNS.md §"algorithms_golden_test.go", §"cross_algorithm_consistency_test.go"
    - .planning/phases/04-remaining-character-gestalt/04-VALIDATION.md (rows 04-05-01, 04-05-02, 04-05-03, 04-05-04)
    - planner planning_context block on OQ-1 — TestCrossAlgorithm_RO_AsymmetryPin asserts INEQUALITY on tide/diet; TestCrossAlgorithm_RO_PinnedAgainstDifflib asserts byte-for-byte difflib agreement on a divergence pair
    - swg.go (for confirming JaroWinklerScore + LevenshteinScore signatures used in the comparisons)
  </read_first>
  <behavior>
    - `go test -run TestGolden_Algorithms_Merge -update ./...` produces a testdata/golden/algorithms.json containing entries for Strcmp95, LCSStr, and RatcliffObershelp alphabetically interleaved with existing Phase 2 + 3 entries
    - `go test -run TestGolden_Algorithms_Merge ./...` (without -update) passes — the committed golden matches the merged staging files exactly
    - TestCrossAlgorithm_Strcmp95_AtLeastJaroWinkler asserts Strcmp95Score(a, b) >= JaroWinklerScore(a, b) on hand-pinned pairs {MARTHA/MARHTA, DWAYNE/DUANE, DIXON/DICKSONX}
    - TestCrossAlgorithm_LCSStr_AtLeastLevenshtein_SubstringContainment asserts LCSStrScore("http_request", "http_request_header_fields") >= LevenshteinScore("http_request", "http_request_header_fields")
    - TestCrossAlgorithm_RatcliffObershelp_PinnedDrDobbs asserts RatcliffObershelpScore("WIKIMEDIA", "WIKIMANIA") within 1e-9 of 0.7777777777777778 (pinned difflib autojunk=False value)
    - TestCrossAlgorithm_RatcliffObershelp_PinnedAgainstDifflib asserts a documented pair where RO visibly differs from BOTH LevenshteinScore AND JaroWinklerScore but MATCHES difflib(autojunk=False).ratio() within 1e-9 — proving RO is the divergent, difflib-equivalent algorithm
    - TestCrossAlgorithm_RatcliffObershelp_AsymmetryPin asserts RatcliffObershelpScore("tide", "diet") != RatcliffObershelpScore("diet", "tide") — INEQUALITY form (inverse of the existing divergence template; OQ-1 resolution regression guard)
  </behavior>
  <action>
    Extend the TestGolden_Algorithms_Merge stagingFiles slice in algorithms_golden_test.go to append three entries: "_staging/strcmp95.json", "_staging/lcsstr.json", "_staging/ratcliff_obershelp.json". Preserve the existing 7 entries (damerau_full, damerau_osa, hamming, jaro, jarowinkler, levenshtein, swg). The slice should be sorted alphabetically by filename for readability (Phase 2 + 3 convention).

    Run `go test -run TestGolden_Algorithms_Merge -update ./...` to regenerate testdata/golden/algorithms.json. The merge logic in TestGolden_Algorithms_Merge reads each staging file, concatenates the entries, sorts alphabetically by Name, marshals via CanonicalMarshalForTest, and writes to algorithms.json. Verify by inspection: the file now contains Strcmp95_*, LCSStr_*, RatcliffObershelp_* entries alphabetically interleaved with Phase 2 + 3 entries.

    Run `go test -run TestGolden_Algorithms_Merge ./...` (without -update) to confirm the committed golden matches.

    Append FOUR new test functions to cross_algorithm_consistency_test.go (after the existing TestCrossAlgorithm_SWG_Levenshtein_SubstringDivergence). Pattern from PATTERNS.md §"cross_algorithm_consistency_test.go" (lines 1064–1126 area):

    TestCrossAlgorithm_Strcmp95_AtLeastJaroWinkler: hand-pinned table of pairs {MARTHA/MARHTA, DWAYNE/DUANE, DIXON/DICKSONX}. For each (a, b) compute gotStrcmp := fuzzymatch.Strcmp95Score(a, b) and gotJW := fuzzymatch.JaroWinklerScore(a, b); assert gotStrcmp >= gotJW; on failure t.Errorf with both values.

    TestCrossAlgorithm_LCSStr_AtLeastLevenshtein_SubstringContainment: hand-pinned pair "http_request" / "http_request_header_fields". Compute gotLCS := fuzzymatch.LCSStrScore(a, b) and gotLev := fuzzymatch.LevenshteinScore(a, b); assert gotLCS >= gotLev.

    TestCrossAlgorithm_RatcliffObershelp_PinnedDrDobbs: pin the Dr. Dobb's 1988 vector. Compute got := fuzzymatch.RatcliffObershelpScore("WIKIMEDIA", "WIKIMANIA"); assert `math.Abs(got - 0.7777777777777778) <= 1e-9` (or use the actual difflib-pinned value from Task 2 of plan 04-04 — confirm the value matches the committed corpus's wikimedia_wikimania entry). This is the Phase 3 WR-03 closure: numerical regression pin OUTSIDE the corpus.

    TestCrossAlgorithm_RatcliffObershelp_PinnedAgainstDifflib: pick a pair from the cross-validation corpus where RO visibly differs from both Levenshtein and Jaro-Winkler. Candidate: gestalt_paper ("GESTALT" / "GESTALT_PATTERN_MATCHING"). Compute gotRO := fuzzymatch.RatcliffObershelpScore("GESTALT", "GESTALT_PATTERN_MATCHING"), gotLev := fuzzymatch.LevenshteinScore(same), gotJW := fuzzymatch.JaroWinklerScore(same). Assert gotRO is NOT within 1e-6 of either gotLev or gotJW (divergence proof). Also assert gotRO matches the pinned difflib(autojunk=False) value within 1e-9.

    TestCrossAlgorithm_RatcliffObershelp_AsymmetryPin: INVERSE-form test asserting INEQUALITY. Compute fwd := fuzzymatch.RatcliffObershelpScore("tide", "diet") and rev := fuzzymatch.RatcliffObershelpScore("diet", "tide"); assert fwd != rev. On failure t.Errorf "RatcliffObershelpScore is INTENTIONALLY asymmetric for tide/diet — got fwd=%g==rev=%g (regression to symmetric behaviour)". This is the OQ-1 resolution regression guard.

    Stdlib testing only; use math.Abs from the math package; use `if !(condition)` t.Errorf form for ordering assertions to match the existing TestCrossAlgorithm_SWG_Levenshtein_SubstringDivergence shape.
  </action>
  <verify>
    <automated>cd /Users/johnny/Development/fuzzymatch && go test -run 'TestGolden_Algorithms_Merge|TestCrossAlgorithm_Strcmp95_AtLeastJaroWinkler|TestCrossAlgorithm_LCSStr_AtLeastLevenshtein_SubstringContainment|TestCrossAlgorithm_RatcliffObershelp_PinnedDrDobbs|TestCrossAlgorithm_RatcliffObershelp_PinnedAgainstDifflib|TestCrossAlgorithm_RatcliffObershelp_AsymmetryPin' -v ./... && grep -q "Strcmp95_" testdata/golden/algorithms.json && grep -q "LCSStr_" testdata/golden/algorithms.json && grep -q "RatcliffObershelp_" testdata/golden/algorithms.json</automated>
  </verify>
  <done>
    testdata/golden/algorithms.json contains the merged Phase 4 entries; TestGolden_Algorithms_Merge passes without -update. All four new cross-algorithm consistency tests green. AsymmetryPin asserts INEQUALITY (inverse form) and surfaces a regression to symmetric behaviour as a clear t.Errorf message.
  </done>
</task>

<task type="auto" tdd="true">
  <name>Task 2: Extend identifier-similarity example program (7 to 10 columns)</name>
  <files>examples/identifier-similarity/main.go, examples/identifier-similarity/main_test.go</files>
  <read_first>
    - examples/identifier-similarity/main.go (full file, especially lines 73–84 — the `algorithms` slice; lines 89 area — algoWidth constant; the table-rendering code that uses algoWidth)
    - examples/identifier-similarity/main_test.go (full file — the `want` constant, the TestExample_ColumnWidths test, the defer-restore os.Stdout pattern locked by Phase 3 WR-04)
    - .planning/phases/04-remaining-character-gestalt/04-PATTERNS.md §"examples/identifier-similarity/main.go (extend in plan 04-05)", §"examples/identifier-similarity/main_test.go (extend in plan 04-05)" — especially the gotcha about algoWidth and the recommended short label "RO" for Ratcliff-Obershelp
    - .planning/phases/04-remaining-character-gestalt/04-VALIDATION.md (row 04-05-05)
  </read_first>
  <behavior>
    - `(cd examples/identifier-similarity && go run .)` prints a 10-column table with the new Strcmp95, LCSStr, and RO columns appended
    - `(cd examples/identifier-similarity && go test ./...)` exits 0
    - TestExample_ColumnWidths still passes — algoWidth and the rendered column header line up correctly
    - The `want` constant in main_test.go is byte-identical to the actual stdout from `go run .`
  </behavior>
  <action>
    Extend examples/identifier-similarity/main.go's `algorithms` slice (around lines 73–84) by appending three entries: {"Strcmp95", fuzzymatch.Strcmp95Score}, {"LCSStr", fuzzymatch.LCSStrScore}, {"RO", fuzzymatch.RatcliffObershelpScore}. Use the SHORT label "RO" for Ratcliff-Obershelp per PATTERNS.md (the function name "RatcliffObershelp" is 17 chars and overflows the existing algoWidth=13; the planner recommendation per PATTERNS.md is option (b) — short label). Add a brief code comment next to the "RO" entry explaining the abbreviation.

    If a longer column header is preferred, raise algoWidth to fit "RatcliffOber" (12 chars) — but the recommended path is to keep algoWidth unchanged and use the "RO" short label (option (b) per PATTERNS.md). The decision affects TestExample_ColumnWidths.

    Capture the new output: run `cd examples/identifier-similarity && go run .` and copy the full stdout. Replace the `want` constant in examples/identifier-similarity/main_test.go with the captured output (header line + separator line + 7 data rows, each row now 10 columns wide). The defer-restore os.Stdout pattern (Phase 3 WR-04 closure) is UNCHANGED; only the `want` constant body changes.

    Re-run `(cd examples/identifier-similarity && go test ./...)` to confirm byte-stability between the example program's stdout and the `want` constant.
  </action>
  <verify>
    <automated>cd /Users/johnny/Development/fuzzymatch && (cd examples/identifier-similarity && go test ./...) && grep -q "Strcmp95" examples/identifier-similarity/main.go && grep -q "LCSStr" examples/identifier-similarity/main.go && grep -q "RatcliffObershelpScore" examples/identifier-similarity/main.go</automated>
  </verify>
  <done>
    Example program runs and tests pass. Three new algorithm columns appear in stdout; TestExample_ColumnWidths green; `want` constant byte-identical to actual output.
  </done>
</task>

<task type="auto">
  <name>Task 3: llms.txt + llms-full.txt sync</name>
  <files>llms.txt, llms-full.txt</files>
  <read_first>
    - llms.txt (full file, especially lines 87–94 — the existing SWG block as the formatting template)
    - llms-full.txt (full file, especially the SWG section as the formatting template)
    - ai_friendly_test.go (read TestLLMs_PublicSymbolsListed to understand the AST-based meta-test gate)
    - .planning/phases/04-remaining-character-gestalt/04-PATTERNS.md §"llms.txt + llms-full.txt (extend in plan 04-05)" — the 7 new symbol entries and the "AlgoID constants already listed" gotcha
    - .planning/phases/04-remaining-character-gestalt/04-VALIDATION.md (row 04-05-06)
  </read_first>
  <acceptance_criteria>
    - `grep -c "Strcmp95Score" llms.txt` returns at least 1
    - `grep -c "LongestCommonSubstring" llms.txt` returns at least 2 (one for Runes, one for byte path)
    - `grep -c "LCSStrScore" llms.txt` returns at least 2 (one for Runes, one for byte path)
    - `grep -c "RatcliffObershelpScore" llms.txt` returns at least 2 (one for Runes, one for byte path)
    - `go test -run TestLLMs ./...` exits 0
    - `go test -run TestLLMs_PublicSymbolsListed ./...` (if present) exits 0
    - `! grep -E "AlgoStrcmp95|AlgoLCSStr|AlgoRatcliffObershelp" <new_lines_only>` — the 3 AlgoID constants are ALREADY listed (Phase 1) and no duplicate entries are added
  </acceptance_criteria>
  <action>
    Append 7 new function entries to llms.txt under the catalogue section (locate the existing SWG block around lines 87–94 and follow the same formatting). Entries to add, one per line, in the same shape as the existing SWG entries:
    - func Strcmp95Score(a, b string) float64
    - func LongestCommonSubstring(a, b string) string
    - func LongestCommonSubstringRunes(a, b string) string
    - func LCSStrScore(a, b string) float64
    - func LCSStrScoreRunes(a, b string) float64
    - func RatcliffObershelpScore(a, b string) float64
    - func RatcliffObershelpScoreRunes(a, b string) float64

    Do NOT add new AlgoID entries — `AlgoStrcmp95`, `AlgoLCSStr`, `AlgoRatcliffObershelp` are already listed (declared in Phase 1's algoid.go and catalogued in llms.txt's AlgoID section).

    Append parallel entries to llms-full.txt with one-line rationales for each symbol. The shape matches the existing SWG block in llms-full.txt — function signature followed by a brief description (1–2 lines). For the difflib-equivalence directive on RatcliffObershelpScore include the autojunk=False qualifier; for LongestCommonSubstring* note the empty-return ambiguity (Pitfall 6); for Strcmp95Score note the JaroScore → JaroWinklerScore → Strcmp95Score hierarchy (RESEARCH.md "Specifics").

    Run `go test -run TestLLMs ./...` to confirm the meta-test passes. If TestLLMs_PublicSymbolsListed is the exact name in ai_friendly_test.go, target that test directly.
  </action>
  <verify>
    <automated>cd /Users/johnny/Development/fuzzymatch && go test -run TestLLMs ./... && grep -q "Strcmp95Score" llms.txt && grep -q "LongestCommonSubstring" llms.txt && grep -q "LCSStrScore" llms.txt && grep -q "RatcliffObershelpScore" llms.txt && grep -q "Strcmp95Score" llms-full.txt && grep -q "RatcliffObershelpScore" llms-full.txt</automated>
  </verify>
  <done>
    Both llms.txt and llms-full.txt contain entries for all 7 new exported symbols. The AST-based meta-test passes. No duplicate AlgoID entries (those were Phase 1 work).
  </done>
</task>

<task type="auto">
  <name>Task 4: Regenerate bench.txt baseline + full quality gate</name>
  <files>bench.txt</files>
  <read_first>
    - bench.txt (current state — Phase 3 baseline; read to understand the file format and the existing benchstat rows)
    - Makefile (`make bench` and `make bench-compare` targets — understand the workflow)
    - .planning/phases/04-remaining-character-gestalt/04-PATTERNS.md §"bench.txt (full-replace in plan 04-05)" — the process is full-replace (no per-row analog)
    - .planning/phases/04-remaining-character-gestalt/04-VALIDATION.md (rows 04-05-07, 04-05-08)
  </read_first>
  <acceptance_criteria>
    - `make bench` exits 0 and produces a fresh bench.txt covering all Phase 1 + 2 + 3 + 4 benchmarks
    - `make bench-compare` exits 0 — benchstat accepts the new baseline
    - bench.txt contains all expected Phase 4 bench rows (Strcmp95Score_*, LCSStrScore_*, LongestCommonSubstring_*, LCSStrScoreRunes_*, LongestCommonSubstringRunes_*, RatcliffObershelpScore_*, RatcliffObershelpScoreRunes_*)
    - `make verify-determinism` exits 0 — testdata/golden/algorithms.json byte-identical across CI matrix
    - `make check` exits 0 — full quality gate green (golangci-lint v2 + go vet + go test -race -shuffle=on + coverage + license + deps + tidy + security)
    - `make test-bdd` exits 0 — all BDD scenarios across SWG + Phase 4 algorithms pass
  </acceptance_criteria>
  <action>
    Run `make bench` on the reference benchmark hardware (per Phase 2 + 3 baseline-pattern). This regenerates bench.txt by re-running every Benchmark* function across the repo and capturing the output in the bench.txt format. The regenerated file replaces the previous Phase 3 bench.txt entirely.

    Verify the regenerated bench.txt contains the expected Phase 4 benchmark rows by grep:
    - BenchmarkStrcmp95Score_ASCII_Short, _ASCII_Medium, _ASCII_Long (3 rows from plan 04-01)
    - BenchmarkLCSStrScore_ASCII_Short, _ASCII_Medium, _ASCII_Long, _Unicode_Short (4 rows from plan 04-02)
    - BenchmarkLongestCommonSubstring_ASCII_Short, _ASCII_Medium, _ASCII_Long, _Unicode_Short (4 rows from plan 04-02)
    - BenchmarkLCSStrScoreRunes_Unicode_Short, BenchmarkLongestCommonSubstringRunes_Unicode_Short (2 rows from plan 04-02)
    - BenchmarkRatcliffObershelpScore_ASCII_Short, _ASCII_Medium, _ASCII_Long, _Unicode_Short (4 rows from plan 04-03)
    - BenchmarkRatcliffObershelpScoreRunes_Unicode_Short (1 row from plan 04-03)
    Total: 18 new bench rows.

    Run `make bench-compare` to verify benchstat accepts the new baseline.

    Run the full quality gate: `make check && make test-bdd && make verify-determinism`. All three must exit 0.

    If `make verify-determinism` fails due to a cross-platform float drift, that is a HARD blocker — surface the failing entry and report back. The cross-platform CI matrix runs the same `go test -run TestGolden_*` against testdata/golden/algorithms.json on linux/amd64, linux/arm64, darwin/arm64, windows/amd64. Any byte-level diff is a DET-01 regression.

    Commit bench.txt as a standalone commit (matches Phase 2 02-07 finalisation pattern — bench.txt updates land as a single commit, not interleaved with algorithm changes).
  </action>
  <verify>
    <automated>cd /Users/johnny/Development/fuzzymatch && make bench && grep -q "BenchmarkStrcmp95Score" bench.txt && grep -q "BenchmarkLCSStrScore" bench.txt && grep -q "BenchmarkLongestCommonSubstring" bench.txt && grep -q "BenchmarkRatcliffObershelpScore" bench.txt && make verify-determinism && make check && make test-bdd</automated>
  </verify>
  <done>
    bench.txt regenerated with all Phase 4 rows; benchstat accepts. make verify-determinism + make check + make test-bdd all exit 0. Phase 4 ships green. ROADMAP.md can be updated to mark Phase 4 complete; STATE.md can be updated to point at Phase 5.
  </done>
</task>

</tasks>

<threat_model>
## Trust Boundaries

| Boundary | Description |
|----------|-------------|
| developer → bench.txt | Local benchmark run produces a committed file; reference benchmark hardware controls determinism of bench results |
| CI matrix → testdata/golden/algorithms.json | Multi-platform CI reads the canonical golden and asserts byte-identical match; any drift fails verify-determinism |

## STRIDE Threat Register (ASVS Level 1)

| Threat ID | Category | Component | Disposition | Mitigation Plan |
|-----------|----------|-----------|-------------|-----------------|
| T-fuzz-panic | D (Denial of Service via panic on malformed input) | Strcmp95Score / LCSStr* / RatcliffObershelp* (re-exercised by cross-algorithm consistency tests) | mitigate | Already mitigated per-plan in 04-01 / 04-02 / 04-03 via FuzzXxxScore harnesses (≥ 60s each). This finalisation plan re-exercises hand-pinned pairs but does not introduce new public surface |
| T-complexity-attack | D (Denial of Service via algorithmic complexity) | bench.txt regeneration on long-input benches | accept | Bench rows for ASCII_Long inputs document the worst-case performance; benchstat regression > 10% in subsequent phases fails CI per PERF-04. No new mitigation in this plan beyond the existing bench infrastructure |
| T-float-determinism | T (Tampering of float reduction order across architectures) | testdata/golden/algorithms.json merged from staging files | mitigate | Cross-platform CI matrix verifies byte-identical golden output (make verify-determinism). The canonical marshal via CanonicalMarshalForTest enforces deterministic byte ordering. Any drift on linux/amd64 vs linux/arm64 vs darwin/arm64 vs windows/amd64 fails the gate hard. The four new cross-algorithm consistency tests use the same `math.Abs(diff) <= 1e-9` tolerance convention as Phase 3 |
</threat_model>

<verification>
- `go test -run 'TestGolden_Algorithms_Merge|TestCrossAlgorithm_Strcmp95_AtLeastJaroWinkler|TestCrossAlgorithm_LCSStr_AtLeastLevenshtein_SubstringContainment|TestCrossAlgorithm_RatcliffObershelp_PinnedDrDobbs|TestCrossAlgorithm_RatcliffObershelp_PinnedAgainstDifflib|TestCrossAlgorithm_RatcliffObershelp_AsymmetryPin' ./...` exits 0.
- `grep -q "Strcmp95_" testdata/golden/algorithms.json && grep -q "LCSStr_" testdata/golden/algorithms.json && grep -q "RatcliffObershelp_" testdata/golden/algorithms.json` — all three Phase 4 algorithms present in the canonical golden.
- `(cd examples/identifier-similarity && go test ./...)` exits 0.
- `go test -run TestLLMs ./...` exits 0; `grep -q "Strcmp95Score|RatcliffObershelpScore" llms.txt` finds the new entries.
- `make bench` exits 0; bench.txt contains all 18 expected Phase 4 bench rows.
- `make bench-compare` exits 0.
- `make verify-determinism` exits 0 — golden byte-identical on CI matrix.
- `make check` exits 0 — full quality gate.
- `make test-bdd` exits 0 — all BDD scenarios green.
- ROADMAP.md Phase 4 marked complete; STATE.md updated to point at Phase 5.
</verification>

<success_criteria>
- All four tasks complete; all listed verification commands green.
- testdata/golden/algorithms.json byte-stable across the CI matrix (linux/amd64, linux/arm64, darwin/arm64, windows/amd64) — DET-01 gate closed.
- Strcmp95 + LCSStr + Ratcliff-Obershelp end-to-end coverage: dispatch + algorithm + tests + bench + fuzz + BDD + staging-golden + canonical-golden + cross-algorithm-consistency + example program + llms.txt + bench.txt baseline — full Phase 2 + 3 quality bar inherited.
- OQ-1 resolution locked by TestCrossAlgorithm_RatcliffObershelp_AsymmetryPin (inverse-form INEQUALITY test) and TestCrossAlgorithm_RatcliffObershelp_PinnedAgainstDifflib.
- All three Phase 4 requirement IDs (CHAR-07, CHAR-09, GESTALT-01) marked complete in ROADMAP traceability.
- Phase 5 (Q-gram algorithms) can begin — Phase 4 ships green.
</success_criteria>

<output>
After completion, create `.planning/phases/04-remaining-character-gestalt/04-05-finalisation-SUMMARY.md` per the GSD summary template. Also update `.planning/STATE.md` to mark Phase 4 complete and point `Current focus` at Phase 5, and update `.planning/ROADMAP.md` traceability for CHAR-07, CHAR-09, GESTALT-01.
</output>
