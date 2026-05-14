---
phase: 05-q-gram-algorithms
plan: 05
type: execute
wave: 3
depends_on:
  - 05-01-qgram-foundation-jaccard
  - 05-02-sorensen-dice
  - 05-03-cosine
  - 05-04-tversky
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
  - QGRAM-01
  - QGRAM-02
  - QGRAM-03
  - QGRAM-04
  - QGRAM-05
tags: [finalisation, golden-merge, cross-platform-determinism-gate, cross-algorithm-consistency, identifier-similarity-example, llms-sync, bench-baseline, ci-matrix-determinism, cosine-load-bearing]

must_haves:
  truths:
    # Golden file merge (QGRAM-01 + 04 + LOAD-BEARING Cosine determinism gate per CONTEXT.md §1)
    - "testdata/golden/algorithms.json contains entries for QGramJaccard, SorensenDice, Cosine, and Tversky — merged from the four staging files committed in plans 05-01, 05-02, 05-03, 05-04 via the existing TestGolden_Algorithms_Merge -update flow"
    - "`make verify-determinism` exits 0 across the cross-platform CI matrix (linux/amd64, linux/arm64, darwin/arm64, windows/amd64) — testdata/golden/algorithms.json is byte-identical on every platform"
    - "**LOAD-BEARING:** the Cosine entries in algorithms.json (9 entries spanning ASCII + Unicode at n ∈ {2, 3, 4} per CONTEXT.md §1) produce byte-identical output on every CI platform. ANY single-byte drift fails the gate HARD and surfaces as a DET-01 regression. This is the closure for the QGRAM-04 determinism requirement"
    # Cross-algorithm consistency tests (algorithm-correctness; defence-in-depth)
    - "TestCrossAlgorithm_Tversky_JaccardEquivalence asserts TverskyScore(a, b, n, 1.0, 1.0) == QGramJaccardScore(a, b, n) bit-for-bit on hand-pinned pairs — re-exercises RV-T3 cross-check from plan 05-04 at the cross-algorithm consistency layer"
    - "TestCrossAlgorithm_Tversky_DiceEquivalence asserts TverskyScore(a, b, n, 0.5, 0.5) == SorensenDiceScore(a, b, n) bit-for-bit on hand-pinned pairs — re-exercises RV-T4 cross-check"
    - "TestCrossAlgorithm_QGramJaccard_AtMostSorensenDice asserts QGramJaccardScore(a, b, n) <= SorensenDiceScore(a, b, n) on hand-pinned pairs (the algorithm-hierarchy invariant: DSC = 2·J/(1+J) ≥ J for J in [0, 1]; documented in test comment with the algebraic derivation)"
    - "TestCrossAlgorithm_Cosine_GeometricMeanBound asserts CosineScore(a, b, n) ≤ 1.0 AND CosineScore(a, b, n) ≥ 0.0 AND on identity pairs CosineScore == 1.0 bit-exact — sanity defence; the cross-platform CI matrix is the primary load-bearing gate for Cosine"
    - "TestCrossAlgorithm_Tversky_AsymmetryPin asserts TverskyScore(\"abcd\", \"abcdef\", 2, 0.8, 0.2) != TverskyScore(\"abcdef\", \"abcd\", 2, 0.8, 0.2) — INEQUALITY-form regression guard for the RV-T1 vs RV-T2 asymmetry pair (additional layer of defence beyond the plan 05-04 unit test + BDD scenario)"
    # Example program extension (CONTEXT.md §5 LOCKED — extend the existing program, NOT create a new one)
    - "examples/identifier-similarity/main.go's `algorithms` slice grows from 10 entries (after Phase 4) to 14 entries; new rows: {label \"QGramJ\", func wrapping QGramJaccardScore with default n=3}, {label \"Dice\", SorensenDiceScore with default n=3}, {label \"Cos\", CosineScore with default n=3}, {label \"Tversky\", TverskyScore with default n=3, α=β=1.0 Jaccard-fallback}"
    - "examples/identifier-similarity/main_test.go's `want` constant is regenerated via `go run .` and committed; `(cd examples/identifier-similarity && go test ./...)` exits 0; the TestExample_ColumnWidths test still passes (algoWidth raised if needed to accommodate the new 14-column layout; recommendation per RESEARCH.md OQ-3 — keep algoWidth small and use SHORT labels per Phase 4 PATTERNS.md gotcha)"
    - "examples/identifier-similarity/main_test.go uses the Phase 3 WR-04 defer-restore os.Stdout pattern (no changes to the test logic — only the `want` constant changes)"
    # llms.txt + llms-full.txt sanity check (sync verification at phase end)
    - "llms.txt lists ALL 8 new exported symbols from Phase 5: QGramJaccardScore, QGramJaccardScoreRunes, SorensenDiceScore, SorensenDiceScoreRunes, CosineScore, CosineScoreRunes, TverskyScore, TverskyScoreRunes. The 2 new error sentinels ErrInvalidQGramSize, ErrInvalidTverskyParam are also listed. The 4 AlgoID constants (AlgoQGramJaccard, AlgoSorensenDice, AlgoCosine, AlgoTversky) are ALREADY listed (declared in Phase 1) — no new AlgoID entries needed"
    - "llms-full.txt has parallel entries with one-line rationales for each new symbol"
    - "ai_friendly_test.go::TestLLMs_PublicSymbolsListed passes — the AST-based meta-test asserts every exported symbol is listed; missing any of the 8 new functions or 2 new sentinels fails the test"
    # bench.txt baseline (PERF-04)
    - "bench.txt is FULL-REPLACED via `make bench` on the reference benchmark hardware; the new file contains Phase 5's benchmark rows (32 new bench labels per RESEARCH.md §6 Dependencies) alongside Phase 2 + 3 + 4 rows; `make bench-compare` accepts the new baseline"
    - "All 32 Phase 5 bench-rows are present in bench.txt: BenchmarkQGramJaccardScore_{ASCII_Short, ASCII_Medium, ASCII_Long}, BenchmarkQGramJaccardScoreRunes_Unicode_Short (4 rows from plan 05-01); BenchmarkSorensenDiceScore_{ASCII_Short, ASCII_Medium, ASCII_Long}, BenchmarkSorensenDiceScoreRunes_Unicode_Short (4 rows from plan 05-02); BenchmarkCosineScore_{ASCII_Short, ASCII_Medium, ASCII_Long}, BenchmarkCosineScoreRunes_Unicode_Short (4 rows from plan 05-03); BenchmarkTverskyScore_{ASCII_Short, ASCII_Medium, ASCII_Long}, BenchmarkTverskyScoreRunes_Unicode_Short (4 rows from plan 05-04). Total: 16 unique Benchmark functions, each producing one row, = 16 rows minimum. If a plan adds duplicate benchmark labels (e.g. medium and large variants), the count grows proportionally"
    # Phase gate (DET-01 + CI-06 + roadmap traceability)
    - "`make check` exits 0 at end of plan (golangci-lint v2 + go vet + go test -race -shuffle=on + coverage + license + deps + tidy + security)"
    - "`make test-bdd` exits 0 — all BDD scenarios across Phase 2 + 3 + 4 + 5 algorithms pass"
    - "`make verify-determinism` exits 0 — testdata/golden/algorithms.json byte-stable on CI matrix; the LOAD-BEARING Cosine entries pass byte-diff on every platform"
    - "All Phase 5 requirement IDs (QGRAM-01, QGRAM-02, QGRAM-03, QGRAM-04, QGRAM-05) marked complete in ROADMAP traceability"
    - "ROADMAP.md Phase 5 entry has all four success criteria checked off; STATE.md `Current focus` updated to point at Phase 6 (Token-based algorithms)"
  artifacts:
    - path: "testdata/golden/algorithms.json"
      provides: "Canonical multi-algorithm golden file with QGramJaccard + SorensenDice + Cosine + Tversky entries merged from the four Phase 5 staging files; sorted alphabetically; canonical-marshalled via CanonicalMarshalForTest. Phase 4 left the file with 59 entries; Phase 5 adds ~33 new entries (RESEARCH.md §6 Dependencies estimate); final count ≈ 92"
      contains: "Cosine_ascii_n2_irrational"
    - path: "algorithms_golden_test.go"
      provides: "Extended TestGolden_Algorithms_Merge stagingFiles slice now includes _staging/qgram_jaccard.json, _staging/sorensen_dice.json, _staging/cosine.json, _staging/tversky.json alongside the existing 10 Phase 2/3/4 staging files"
    - path: "cross_algorithm_consistency_test.go"
      provides: "Appended 5 new cross-algorithm tests: TestCrossAlgorithm_Tversky_JaccardEquivalence, TestCrossAlgorithm_Tversky_DiceEquivalence, TestCrossAlgorithm_QGramJaccard_AtMostSorensenDice, TestCrossAlgorithm_Cosine_GeometricMeanBound, TestCrossAlgorithm_Tversky_AsymmetryPin"
    - path: "examples/identifier-similarity/main.go"
      provides: "Extended algorithms slice 10 → 14 entries (QGramJ, Dice, Cos, Tversky); SHORT labels chosen to fit existing algoWidth per Phase 4 PATTERNS.md gotcha. If algoWidth raised, document in commit message"
    - path: "examples/identifier-similarity/main_test.go"
      provides: "Regenerated `want` constant (one `go run .` capture); TestExample_ColumnWidths still passes; defer-restore os.Stdout pattern unchanged"
    - path: "llms.txt"
      provides: "Confirmed all 8 Phase 5 exported symbols + 2 new error sentinels listed (added incrementally per plan during 05-01..05-04; this plan verifies completeness via the meta-test)"
    - path: "llms-full.txt"
      provides: "Parallel entries with one-line rationales (added incrementally; verified here)"
    - path: "bench.txt"
      provides: "Full-replaced via `make bench` on the reference hardware; includes Phase 5 bench rows alongside Phase 2 + 3 + 4 rows; benchstat baseline accepts"
  key_links:
    - from: "algorithms_golden_test.go (TestGolden_Algorithms_Merge stagingFiles slice)"
      to: "testdata/golden/_staging/{qgram_jaccard,sorensen_dice,cosine,tversky}.json (committed in plans 05-01, 05-02, 05-03, 05-04)"
      via: "Slice extension; running `go test -run TestGolden_Algorithms_Merge -update ./...` materialises the merged algorithms.json"
      pattern: "_staging/(qgram_jaccard|sorensen_dice|cosine|tversky)\\.json"
    - from: "testdata/golden/algorithms.json (Cosine entries)"
      to: "Cross-platform CI matrix (`make verify-determinism`)"
      via: "9 Cosine entries spanning ASCII + Unicode at n ∈ {2, 3, 4} are byte-compared on linux/amd64, linux/arm64, darwin/arm64, windows/amd64; any single-byte drift fails the gate. This is the LOAD-BEARING closure for the QGRAM-04 cross-platform float-determinism requirement"
      pattern: "Cosine_(ascii|unicode)_n(2|3|4)"
    - from: "cross_algorithm_consistency_test.go (Tversky_JaccardEquivalence, Tversky_DiceEquivalence)"
      to: "qgram_jaccard.go (QGramJaccardScore) + sorensen_dice.go (SorensenDiceScore) + tversky.go (TverskyScore)"
      via: "Re-exercises RV-T3 (α=β=1.0 → Jaccard) and RV-T4 (α=β=0.5 → Dice) at the cross-algorithm consistency layer — defence-in-depth beyond the plan 05-04 unit tests + property tests"
      pattern: "TestCrossAlgorithm_Tversky_(Jaccard|Dice)Equivalence"
    - from: "examples/identifier-similarity/main.go (algorithms slice)"
      to: "examples/identifier-similarity/main_test.go (want constant)"
      via: "After editing main.go's slice, capture stdout from `go run .` and paste into main_test.go's want constant — byte-identical match required by TestExample_ColumnWidths"
      pattern: "var algorithms = \\[\\]struct"
    - from: "llms.txt (all 8 Phase 5 function entries)"
      to: "ai_friendly_test.go::TestLLMs_PublicSymbolsListed (meta-test gate)"
      via: "AST-based meta-test asserts every exported symbol in fuzzymatch package is listed in llms.txt"
      pattern: "QGramJaccardScore|SorensenDiceScore|CosineScore|TverskyScore"
---

<objective>
Finalise Phase 5 — merge the four per-algorithm staging goldens into testdata/golden/algorithms.json; run `make verify-determinism` to confirm the cross-platform CI matrix gate passes (the LOAD-BEARING closure for QGRAM-04 Cosine determinism per CONTEXT.md §1); append five cross-algorithm consistency tests (Tversky/Jaccard equivalence at α=β=1.0; Tversky/Dice equivalence at α=β=0.5; QGramJaccard ≤ SorensenDice algebraic hierarchy; Cosine range + identity sanity; Tversky asymmetry pin via RV-T1/RV-T2 INEQUALITY assertion); extend the identifier-similarity example program from 10 to 14 algorithm columns (per CONTEXT.md §5 — extend, do NOT create new); verify llms.txt + llms-full.txt completeness across all 8 new exported symbols + 2 new sentinels (added incrementally during plans 05-01..05-04 per CONTEXT.md §5 LOCKED); full-replace bench.txt via `make bench` to baseline Phase 5. Run `make check`, `make test-bdd`, and `make verify-determinism` to confirm the phase ships green.

Purpose: Phase 5 closes; the q-gram tier public surface is exposed end-to-end (golden, example, llms.txt, bench baseline), the Cosine cross-platform float-determinism gate is verified, and the Tversky asymmetry invariant is pinned at three layers (unit / property / cross-algorithm / BDD — Phase 5 plan 05-04 + 05-05 combined). After this plan lands, Phase 6 (Token-based algorithms — TOKEN-01..TOKEN-05) can begin; Phase 6 will consume the q-gram tier as one permitted inner metric for Monge-Elkan.

Output: 8 modified files (1 merged canonical golden, 7 extensions to existing append-only files). NO new source files — this is the integration / finalisation plan, following the established Phase 2 plan 02-07 / Phase 3 plan 03-03 / Phase 4 plan 04-05 precedent.
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
@.planning/phases/05-q-gram-algorithms/05-CONTEXT.md
@.planning/phases/05-q-gram-algorithms/05-RESEARCH.md
@.planning/phases/05-q-gram-algorithms/05-01-qgram-foundation-jaccard-PLAN.md
@.planning/phases/05-q-gram-algorithms/05-02-sorensen-dice-PLAN.md
@.planning/phases/05-q-gram-algorithms/05-03-cosine-PLAN.md
@.planning/phases/05-q-gram-algorithms/05-04-tversky-PLAN.md
@.planning/phases/04-remaining-character-gestalt/04-05-finalisation-PLAN.md
@.planning/phases/04-remaining-character-gestalt/04-05-finalisation-SUMMARY.md
@.planning/phases/03-smith-waterman-gotoh/03-03-swg-finalisation-SUMMARY.md
@.planning/phases/02-core-character-algorithms-six/02-07-finalisation-SUMMARY.md
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

From algorithms_golden_test.go (extension point — TestGolden_Algorithms_Merge stagingFiles slice):
A `[]string` literal listing the existing 10 staging-file paths (damerau_full, damerau_osa, hamming, jaro, jarowinkler, levenshtein, swg, strcmp95, lcsstr, ratcliff_obershelp). Phase 5 appends four new entries: `_staging/qgram_jaccard.json`, `_staging/sorensen_dice.json`, `_staging/cosine.json`, `_staging/tversky.json`. Sorted alphabetically per Phase 2-4 convention.

From cross_algorithm_consistency_test.go (extension point — append to bottom of file):
Existing templates: TestCrossAlgorithm_SWG_Levenshtein_SubstringDivergence (Phase 3), TestCrossAlgorithm_Strcmp95_AtLeastJaroWinkler / _LCSStr_AtLeastLevenshtein_SubstringContainment / _RatcliffObershelp_AsymmetryPin (Phase 4). Phase 5 appends 5 new tests with the same shape.

From examples/identifier-similarity/main.go (extension point — algorithms slice):
Existing 10 entries after Phase 4 (Levenshtein, DL-OSA, DL-Full, Hamming, Jaro, Jaro-Winkler, SWG, Strcmp95, LCSStr, RO). Phase 5 appends 4 entries.

From llms.txt:
All 8 Phase 5 function entries + 2 sentinel entries already added incrementally in plans 05-01..05-04 per CONTEXT.md §5 ("llms.txt entry per plan, not deferred to finalisation"). This plan VERIFIES completeness via TestLLMs_PublicSymbolsListed — it does NOT re-add the entries.

Bench.txt: full-replace; regenerate via `make bench`.
</interfaces>

<tasks>

<task type="auto" tdd="true">
  <name>Task 1: Merge staging goldens + extend cross-algorithm consistency tests</name>
  <files>testdata/golden/algorithms.json, algorithms_golden_test.go, cross_algorithm_consistency_test.go</files>
  <read_first>
    - algorithms_golden_test.go (find the TestGolden_Algorithms_Merge stagingFiles slice; identify the existing 10 Phase 2/3/4 entries; the Phase 5 append point)
    - cross_algorithm_consistency_test.go (full file — examine the Phase 3 TestCrossAlgorithm_SWG_Levenshtein_SubstringDivergence and Phase 4 TestCrossAlgorithm_RatcliffObershelp_AsymmetryPin templates; new Phase 5 tests follow the same shape)
    - testdata/golden/_staging/qgram_jaccard.json (committed in plan 05-01)
    - testdata/golden/_staging/sorensen_dice.json (committed in plan 05-02)
    - testdata/golden/_staging/cosine.json (committed in plan 05-03 — 9 entries spanning ASCII + Unicode at n ∈ {2, 3, 4})
    - testdata/golden/_staging/tversky.json (committed in plan 05-04 — includes BOTH RV-T1 and RV-T2 rows)
    - testdata/golden/algorithms.json (current state — Phase 4 left 59 entries; Phase 5 adds ~33)
    - .planning/phases/05-q-gram-algorithms/05-CONTEXT.md §1 (Cosine cross-platform determinism gate — LOAD-BEARING)
    - .planning/phases/04-remaining-character-gestalt/04-05-finalisation-PLAN.md Task 1 (exact structural analog from Phase 4)
  </read_first>
  <behavior>
    - `go test -run TestGolden_Algorithms_Merge -update ./...` produces a testdata/golden/algorithms.json containing entries for QGramJaccard, SorensenDice, Cosine, Tversky alphabetically interleaved with existing Phase 2/3/4 entries
    - `go test -run TestGolden_Algorithms_Merge ./...` (without -update) passes — the committed golden matches the merged staging files exactly
    - TestCrossAlgorithm_Tversky_JaccardEquivalence asserts TverskyScore(a, b, n, 1.0, 1.0) == QGramJaccardScore(a, b, n) bit-for-bit on hand-pinned pairs (e.g. "abcd"/"abce"/n=2, RV-T3; "AGCT"/"AGCTAGCT"/n=2 from RV-J1)
    - TestCrossAlgorithm_Tversky_DiceEquivalence asserts TverskyScore(a, b, n, 0.5, 0.5) == SorensenDiceScore(a, b, n) bit-for-bit
    - TestCrossAlgorithm_QGramJaccard_AtMostSorensenDice asserts QGramJaccardScore(a, b, n) ≤ SorensenDiceScore(a, b, n) on hand-pinned pairs from RV-J1..J4 and RV-D1..D4
    - TestCrossAlgorithm_Cosine_GeometricMeanBound — sanity test: identity pairs return CosineScore == 1.0 bit-exact; all hand-pinned pairs return values in [0, 1]
    - TestCrossAlgorithm_Tversky_AsymmetryPin — INEQUALITY-form regression guard: asserts TverskyScore("abcd", "abcdef", 2, 0.8, 0.2) != TverskyScore("abcdef", "abcd", 2, 0.8, 0.2). On failure t.Errorf with a message about parameter-order regression
  </behavior>
  <action>
    Step A — Extend the TestGolden_Algorithms_Merge stagingFiles slice in algorithms_golden_test.go to append FOUR entries (alphabetically sorted): `"_staging/cosine.json"`, `"_staging/qgram_jaccard.json"`, `"_staging/sorensen_dice.json"`, `"_staging/tversky.json"`. Preserve the existing 10 entries. The full sorted slice should now contain 14 entries.

    Run `go test -run TestGolden_Algorithms_Merge -update ./...` to regenerate testdata/golden/algorithms.json. The merge logic concatenates entries from each staging file, sorts alphabetically by Name, marshals via CanonicalMarshalForTest, writes to algorithms.json. Verify by inspection: the file now contains Cosine_*, QGramJaccard_*, SorensenDice_*, Tversky_* entries alphabetically interleaved with Phase 2/3/4 entries.

    Run `go test -run TestGolden_Algorithms_Merge ./...` (without -update) to confirm the committed golden matches.

    Step B — Append FIVE new test functions to cross_algorithm_consistency_test.go. Pattern from the Phase 4 plan 04-05 template:

    TestCrossAlgorithm_Tversky_JaccardEquivalence: table-driven over hand-pinned pairs {("abcd", "abce", 2), ("AGCT", "AGCTAGCT", 2), ("hello", "world", 2), ("café", "cafe", 2)}. For each (a, b, n): gotTversky := fuzzymatch.TverskyScore(a, b, n, 1.0, 1.0); gotJaccard := fuzzymatch.QGramJaccardScore(a, b, n); assert `math.Float64bits(gotTversky) == math.Float64bits(gotJaccard)` (bit-exact); on failure t.Errorf with both values and the bit-patterns. Test comment cites RESEARCH.md §1.4 / §2.4 RV-T3 for the algebraic justification.

    TestCrossAlgorithm_Tversky_DiceEquivalence: same shape but α=β=0.5; assert bit-exact equality with SorensenDiceScore. Test comment cites RV-T4.

    TestCrossAlgorithm_QGramJaccard_AtMostSorensenDice: hand-pinned table covering RV-J1..J4 and RV-D1..D4 pairs. For each (a, b, n): gotJ := QGramJaccardScore(a, b, n); gotD := SorensenDiceScore(a, b, n); assert gotJ ≤ gotD; on failure t.Errorf with both values. Test comment derives the algebraic identity DSC = 2·J/(1+J): for J ∈ [0, 1], DSC ≥ J because 2·J/(1+J) ≥ J ⟺ 2J ≥ J(1+J) ⟺ 2J ≥ J + J² ⟺ J ≥ J² ⟺ J(1-J) ≥ 0 ✓ (true for J ∈ [0, 1]).

    TestCrossAlgorithm_Cosine_GeometricMeanBound: sanity table covering RV-C1..RV-C5 inputs plus identity ("hello"/"hello"/n=2) and orthogonal ("abc"/"xyz"/n=2). For each: assert 0 ≤ got ≤ 1; on the identity pair assert got == 1.0 bit-exact via math.Float64bits comparison. This is a defence-in-depth sanity test layered atop the cross-platform CI matrix gate (which is the load-bearing detector).

    TestCrossAlgorithm_Tversky_AsymmetryPin: INVERSE-form INEQUALITY test (same shape as Phase 4's RatcliffObershelp_AsymmetryPin). Compute fwd := fuzzymatch.TverskyScore("abcd", "abcdef", 2, 0.8, 0.2) and rev := fuzzymatch.TverskyScore("abcdef", "abcd", 2, 0.8, 0.2); assert fwd != rev. On failure t.Errorf "TverskyScore is INTENTIONALLY asymmetric for abcd/abcdef with α=0.8, β=0.2 — got fwd=%g == rev=%g (regression to symmetric behaviour or silent α/β swap)". Additionally assert math.Abs(fwd - rev) > 0.1 to surface near-equality regressions early.

    Stdlib testing only; use math from the math package; use the `if !(condition)` t.Errorf form to match the existing Phase 3/4 template.
  </action>
  <verify>
    <automated>cd /Users/johnny/Development/fuzzymatch && go test -run 'TestGolden_Algorithms_Merge|TestCrossAlgorithm_Tversky_JaccardEquivalence|TestCrossAlgorithm_Tversky_DiceEquivalence|TestCrossAlgorithm_QGramJaccard_AtMostSorensenDice|TestCrossAlgorithm_Cosine_GeometricMeanBound|TestCrossAlgorithm_Tversky_AsymmetryPin' -v ./... && grep -q "QGramJaccard_" testdata/golden/algorithms.json && grep -q "SorensenDice_" testdata/golden/algorithms.json && grep -q "Cosine_" testdata/golden/algorithms.json && grep -q "Tversky_" testdata/golden/algorithms.json</automated>
  </verify>
  <done>
    testdata/golden/algorithms.json contains the merged Phase 5 entries; TestGolden_Algorithms_Merge passes without -update. All five new cross-algorithm consistency tests green. AsymmetryPin asserts INEQUALITY (inverse form) and surfaces parameter-order regressions as clear t.Errorf messages.
  </done>
</task>

<task type="auto" tdd="true">
  <name>Task 2: Extend identifier-similarity example program (10 → 14 columns)</name>
  <files>examples/identifier-similarity/main.go, examples/identifier-similarity/main_test.go</files>
  <read_first>
    - examples/identifier-similarity/main.go (full file — the `algorithms` slice, algoWidth constant, table-rendering code; Phase 4 left this with 10 entries)
    - examples/identifier-similarity/main_test.go (full file — the `want` constant, TestExample_ColumnWidths, defer-restore os.Stdout pattern)
    - .planning/phases/04-remaining-character-gestalt/04-PATTERNS.md §"examples/identifier-similarity/main.go" — algoWidth gotcha
    - .planning/phases/05-q-gram-algorithms/05-RESEARCH.md OQ-3 (example column count + label width recommendations)
  </read_first>
  <behavior>
    - `(cd examples/identifier-similarity && go run .)` prints a 14-column table with the new QGramJ, Dice, Cos, Tversky columns appended
    - `(cd examples/identifier-similarity && go test ./...)` exits 0
    - TestExample_ColumnWidths still passes — algoWidth and the rendered column header line up correctly
    - The `want` constant in main_test.go is byte-identical to the actual stdout from `go run .`
  </behavior>
  <action>
    Extend examples/identifier-similarity/main.go's `algorithms` slice by appending four entries. The dispatch signature for the slice is `(a, b string) float64` (no n parameter), so the new entries wrap the q-gram algorithms with default n=3:
    ```go
    {"QGramJ", func(a, b string) float64 { return fuzzymatch.QGramJaccardScore(a, b, 3) }},
    {"Dice",   func(a, b string) float64 { return fuzzymatch.SorensenDiceScore(a, b, 3) }},
    {"Cos",    func(a, b string) float64 { return fuzzymatch.CosineScore(a, b, 3) }},
    {"Tversky", func(a, b string) float64 { return fuzzymatch.TverskyScore(a, b, 3, 1.0, 1.0) }}, // α=β=1.0 → Jaccard-fallback (dispatch convention)
    ```
    Use SHORT labels (QGramJ, Dice, Cos, Tversky) to fit the existing algoWidth — measure first: read algoWidth's current value; "Tversky" is 7 chars; "QGramJ" is 6 chars. If algoWidth=13 (Phase 4 baseline), all four labels fit without raising it. If algoWidth needs to be raised, document in the commit message and update TestExample_ColumnWidths accordingly per Phase 4 PATTERNS.md.

    Capture the new output: run `(cd examples/identifier-similarity && go run .)` and copy the full stdout. Replace the `want` constant in main_test.go with the captured output (header line + separator line + data rows, each row now 14 columns wide). The defer-restore os.Stdout pattern is UNCHANGED.

    Re-run `(cd examples/identifier-similarity && go test ./...)` to confirm byte-stability.
  </action>
  <verify>
    <automated>cd /Users/johnny/Development/fuzzymatch && (cd examples/identifier-similarity && go test ./...) && grep -q "QGramJ" examples/identifier-similarity/main.go && grep -q "Dice" examples/identifier-similarity/main.go && grep -q "Cos" examples/identifier-similarity/main.go && grep -q "Tversky" examples/identifier-similarity/main.go && grep -q "QGramJaccardScore" examples/identifier-similarity/main.go && grep -q "SorensenDiceScore" examples/identifier-similarity/main.go && grep -q "CosineScore" examples/identifier-similarity/main.go && grep -q "TverskyScore" examples/identifier-similarity/main.go</automated>
  </verify>
  <done>
    Example program runs and tests pass. Four new algorithm columns appear in stdout; TestExample_ColumnWidths green; `want` constant byte-identical to actual output.
  </done>
</task>

<task type="auto">
  <name>Task 3: llms.txt + llms-full.txt sync verification</name>
  <files>llms.txt, llms-full.txt</files>
  <read_first>
    - llms.txt (full file — should already contain all 8 Phase 5 function entries + 2 sentinel entries added incrementally in plans 05-01..05-04)
    - llms-full.txt (full file — should already contain parallel entries)
    - ai_friendly_test.go (TestLLMs_PublicSymbolsListed — the AST-based meta-test that asserts every exported symbol is listed)
    - .planning/phases/04-remaining-character-gestalt/04-05-finalisation-PLAN.md Task 3 (Phase 4 sync verification analog)
  </read_first>
  <acceptance_criteria>
    - `grep -c "QGramJaccardScore" llms.txt` returns at least 2 (byte-path + Runes variant)
    - `grep -c "SorensenDiceScore" llms.txt` returns at least 2
    - `grep -c "CosineScore" llms.txt` returns at least 2
    - `grep -c "TverskyScore" llms.txt` returns at least 2
    - `grep -q "ErrInvalidQGramSize" llms.txt && grep -q "ErrInvalidTverskyParam" llms.txt`
    - `go test -run TestLLMs ./...` exits 0
    - `! grep -E "AlgoQGramJaccard|AlgoSorensenDice|AlgoCosine|AlgoTversky" <new_lines_only>` — the 4 AlgoID constants are ALREADY listed (Phase 1) and no duplicate entries are added by Phase 5
  </acceptance_criteria>
  <action>
    Step A — Verify llms.txt completeness. Each of plans 05-01..05-04 was expected to add 2 function entries (Score + ScoreRunes) per CONTEXT.md §5 LOCKED. Plan 05-01 additionally added the 2 error sentinel entries. Total Phase 5 additions: 8 function lines + 2 sentinel lines.

    If any entries are MISSING (this should NOT happen if plans 05-01..05-04 followed the LOCKED per-plan discipline, but defensive verification is the value-add of this task), APPEND the missing entries here. Format matches the existing Phase 4 entries (one bullet per public function, signature inline).

    Step B — Verify llms-full.txt completeness. Each public function should have a parallel entry with a one-line rationale. Tversky's rationale should note the α/β asymmetry surface and the dispatch-table Jaccard-fallback compromise. Cosine's rationale should note the LOAD-BEARING cross-platform float-determinism gate and the sorted-key iteration discipline.

    Step C — Run `go test -run TestLLMs ./...` to confirm the AST-based meta-test passes. If TestLLMs_PublicSymbolsListed is the exact name in ai_friendly_test.go, target that test directly. The meta-test parses the package AST and asserts every exported symbol is listed; missing any of the 8 new functions or 2 new sentinels fails the test.

    Step D — Verify NO duplicate AlgoID entries. AlgoQGramJaccard, AlgoSorensenDice, AlgoCosine, AlgoTversky were declared in Phase 1 and listed in llms.txt at that time. If a plan accidentally re-added them, remove the duplicates here (single source of truth: Phase 1's entries).
  </action>
  <verify>
    <automated>cd /Users/johnny/Development/fuzzymatch && go test -run TestLLMs ./... && grep -c "QGramJaccardScore" llms.txt | awk '$1 >= 2 { exit 0 } { exit 1 }' && grep -c "SorensenDiceScore" llms.txt | awk '$1 >= 2 { exit 0 } { exit 1 }' && grep -c "CosineScore" llms.txt | awk '$1 >= 2 { exit 0 } { exit 1 }' && grep -c "TverskyScore" llms.txt | awk '$1 >= 2 { exit 0 } { exit 1 }' && grep -q "ErrInvalidQGramSize" llms.txt && grep -q "ErrInvalidTverskyParam" llms.txt && [ "$(grep -c 'AlgoQGramJaccard\|AlgoSorensenDice\|AlgoCosine\|AlgoTversky' llms.txt)" -le 4 ]</automated>
  </verify>
  <done>
    llms.txt and llms-full.txt contain entries for all 8 Phase 5 exported functions + 2 new error sentinels. The AST-based meta-test passes. No duplicate AlgoID entries (those were Phase 1 work).
  </done>
</task>

<task type="auto">
  <name>Task 4: Regenerate bench.txt baseline + full quality gate</name>
  <files>bench.txt</files>
  <read_first>
    - bench.txt (current state — Phase 4 baseline; understand the file format and the existing benchstat rows)
    - Makefile (`make bench` and `make bench-compare` targets — understand the workflow)
    - .planning/phases/04-remaining-character-gestalt/04-05-finalisation-PLAN.md Task 4 (exact structural analog — full-replace pattern)
    - .planning/phases/05-q-gram-algorithms/05-RESEARCH.md §6 Dependencies (32 new bench labels — 8 per algorithm × 4 algorithms)
  </read_first>
  <acceptance_criteria>
    - `make bench` exits 0 and produces a fresh bench.txt covering all Phase 1 + 2 + 3 + 4 + 5 benchmarks
    - `make bench-compare` exits 0 — benchstat accepts the new baseline
    - bench.txt contains all 16+ expected Phase 5 bench rows (BenchmarkQGramJaccardScore_*, BenchmarkSorensenDiceScore_*, BenchmarkCosineScore_*, BenchmarkTverskyScore_*, plus their *Runes variants)
    - `make verify-determinism` exits 0 — testdata/golden/algorithms.json byte-identical across CI matrix (THE LOAD-BEARING CLOSURE FOR QGRAM-04)
    - `make check` exits 0 — full quality gate green (golangci-lint v2 + go vet + go test -race -shuffle=on + coverage + license + deps + tidy + security)
    - `make test-bdd` exits 0 — all BDD scenarios across Phase 2 + 3 + 4 + 5 algorithms pass
  </acceptance_criteria>
  <action>
    Run `make bench` on the reference benchmark hardware (Phase 2-4 baseline pattern). This regenerates bench.txt by re-running every Benchmark* function across the repo. The regenerated file replaces the previous Phase 4 bench.txt entirely.

    Verify the regenerated bench.txt contains the expected Phase 5 benchmark rows by grep:
    - BenchmarkQGramJaccardScore_ASCII_Short, _ASCII_Medium, _ASCII_Long, BenchmarkQGramJaccardScoreRunes_Unicode_Short (4 rows from plan 05-01)
    - BenchmarkSorensenDiceScore_ASCII_Short, _ASCII_Medium, _ASCII_Long, BenchmarkSorensenDiceScoreRunes_Unicode_Short (4 rows from plan 05-02)
    - BenchmarkCosineScore_ASCII_Short, _ASCII_Medium, _ASCII_Long, BenchmarkCosineScoreRunes_Unicode_Short (4 rows from plan 05-03)
    - BenchmarkTverskyScore_ASCII_Short, _ASCII_Medium, _ASCII_Long, BenchmarkTverskyScoreRunes_Unicode_Short (4 rows from plan 05-04)
    Total: 16 unique Benchmark functions = 16 rows minimum.

    Run `make bench-compare` to verify benchstat accepts the new baseline.

    Run the full quality gate: `make check && make test-bdd && make verify-determinism`. All three must exit 0.

    **If `make verify-determinism` fails on Cosine entries**, that is a HARD blocker — surface the failing entry and report back. The cross-platform CI matrix runs the same `go test -run TestGolden_*` against testdata/golden/algorithms.json on linux/amd64, linux/arm64, darwin/arm64, windows/amd64. Any byte-level diff is a DET-01 regression. Remediation path per RESEARCH.md §3.4: insert `float64(x*y) + z` explicit cast in cosine.go's dot-product loop — but this should NOT be needed per RESEARCH.md §3.3 empirical finding (the recipe is empirically safe on integer-derived float counts).

    Commit bench.txt as a standalone commit (matches Phase 2 / 3 / 4 finalisation pattern — bench.txt updates land as a single commit).

    Update ROADMAP.md to mark Phase 5 complete: all four success criteria checked off; all five requirement IDs (QGRAM-01..QGRAM-05) marked complete in the traceability table. Update STATE.md `Current focus` to point at Phase 6 (Token-based algorithms).
  </action>
  <verify>
    <automated>cd /Users/johnny/Development/fuzzymatch && make bench && grep -q "BenchmarkQGramJaccardScore" bench.txt && grep -q "BenchmarkSorensenDiceScore" bench.txt && grep -q "BenchmarkCosineScore" bench.txt && grep -q "BenchmarkTverskyScore" bench.txt && make verify-determinism && make check && make test-bdd</automated>
  </verify>
  <done>
    bench.txt regenerated with all Phase 5 rows; benchstat accepts. make verify-determinism + make check + make test-bdd all exit 0. Phase 5 ships green. ROADMAP.md marks Phase 5 complete with all 5 QGRAM-NN requirement IDs checked off; STATE.md points at Phase 6.
  </done>
</task>

</tasks>

<threat_model>
## Trust Boundaries

| Boundary | Description |
|----------|-------------|
| developer → bench.txt | Local benchmark run produces a committed file; reference benchmark hardware controls determinism of bench results |
| CI matrix → testdata/golden/algorithms.json | Multi-platform CI reads the canonical golden and asserts byte-identical match; any drift fails verify-determinism. **LOAD-BEARING for Phase 5 — Cosine entries are the gate** |

## STRIDE Threat Register (ASVS Level 1)

| Threat ID | Category | Component | Disposition | Mitigation Plan |
|-----------|----------|-----------|-------------|-----------------|
| T-fuzz-panic | D | QGramJaccard / SorensenDice / Cosine / Tversky (re-exercised by cross-algorithm consistency tests) | mitigate | Already mitigated per-plan in 05-01..05-04 via FuzzXxx harnesses (≥ 60s each). This finalisation plan re-exercises hand-pinned pairs but does not introduce new public surface |
| T-complexity-attack | D | bench.txt regeneration on long-input benches | accept | Bench rows for ASCII_Long inputs document the worst-case performance; benchstat regression > 10% in subsequent phases fails CI per PERF-04. No new mitigation in this plan beyond the existing bench infrastructure |
| T-float-determinism-cosine | T | testdata/golden/algorithms.json merged from staging files — Cosine entries are LOAD-BEARING | mitigate | LOAD-BEARING. Cross-platform CI matrix verifies byte-identical golden output (make verify-determinism). The canonical marshal via CanonicalMarshalForTest enforces deterministic byte ordering. The 9 Cosine entries span ASCII + Unicode at n ∈ {2, 3, 4} per CONTEXT.md §1 LOCKED — multiple intersection sizes exercise the float-reduction path. Any drift on linux/amd64 vs linux/arm64 vs darwin/arm64 vs windows/amd64 fails the gate HARD. Five cross-algorithm consistency tests (Task 1) layer defence-in-depth atop the CI matrix |
| T-parameter-order-bug | T (Tampering — silent α/β swap in Tversky) | TverskyScore implementation | mitigate | LOAD-BEARING. TestCrossAlgorithm_Tversky_AsymmetryPin (Task 1) is the fourth-layer defence (after plan 05-04's unit test + property test + BDD scenario). The combination of unit + property + cross-algorithm + BDD coverage forms a four-layer guard against α/β regressions |
</threat_model>

<verification>
- `go test -run 'TestGolden_Algorithms_Merge|TestCrossAlgorithm_Tversky_JaccardEquivalence|TestCrossAlgorithm_Tversky_DiceEquivalence|TestCrossAlgorithm_QGramJaccard_AtMostSorensenDice|TestCrossAlgorithm_Cosine_GeometricMeanBound|TestCrossAlgorithm_Tversky_AsymmetryPin' ./...` exits 0.
- `grep -q "QGramJaccard_" testdata/golden/algorithms.json && grep -q "SorensenDice_" testdata/golden/algorithms.json && grep -q "Cosine_" testdata/golden/algorithms.json && grep -q "Tversky_" testdata/golden/algorithms.json` — all four Phase 5 algorithms present.
- `(cd examples/identifier-similarity && go test ./...)` exits 0.
- `go test -run TestLLMs ./...` exits 0; grep counts for each new symbol in llms.txt are ≥ 2.
- `make bench` exits 0; bench.txt contains all 16+ expected Phase 5 bench rows.
- `make bench-compare` exits 0.
- `make verify-determinism` exits 0 — golden byte-identical on CI matrix (THE LOAD-BEARING CLOSURE for QGRAM-04).
- `make check` exits 0 — full quality gate.
- `make test-bdd` exits 0 — all BDD scenarios green.
- ROADMAP.md Phase 5 marked complete with all five QGRAM-NN requirement IDs checked off; STATE.md updated to point at Phase 6.
</verification>

<success_criteria>
- All four tasks complete; all listed verification commands green.
- testdata/golden/algorithms.json byte-stable across the CI matrix (linux/amd64, linux/arm64, darwin/arm64, windows/amd64) — **LOAD-BEARING DET-01 + QGRAM-04 closure**.
- QGramJaccard + SorensenDice + Cosine + Tversky end-to-end coverage: dispatch + algorithm + tests + bench + fuzz + BDD + staging-golden + canonical-golden + cross-algorithm-consistency + example program + llms.txt + bench.txt baseline — full Phase 2/3/4 quality bar inherited.
- The Tversky asymmetry invariant is pinned at FOUR layers: unit test (plan 05-04 TestTversky_AsymmetryDirectionSensitive), property test (plan 05-04 TestProp_TverskyScore_AsymmetricWhenAlphaNeqBeta + ParameterSwapSymmetry), cross-algorithm consistency test (this plan TestCrossAlgorithm_Tversky_AsymmetryPin), BDD scenario (plan 05-04 tversky.feature Asymmetry scenario).
- All five Phase 5 requirement IDs (QGRAM-01, QGRAM-02, QGRAM-03, QGRAM-04, QGRAM-05) marked complete in ROADMAP traceability.
- STATE.md `Current focus` points at Phase 6 (Token-based algorithms — TOKEN-01..TOKEN-05).
- Phase 6 can begin — Phase 5 ships green.
</success_criteria>

<output>
After completion, create `.planning/phases/05-q-gram-algorithms/05-05-finalisation-SUMMARY.md` per the GSD summary template. Also update `.planning/STATE.md` to mark Phase 5 complete and point `Current focus` at Phase 6, and update `.planning/ROADMAP.md` traceability for QGRAM-01..QGRAM-05. If any deferred items surfaced during plans 05-01..05-04 execution, roll them up into `.planning/phases/05-q-gram-algorithms/deferred-items.md` (created on first need per the gsd-executor scope-boundary rule).
</output>
