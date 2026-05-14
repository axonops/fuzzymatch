---
phase: 04-remaining-character-gestalt
plan: 01
type: execute
wave: 1
depends_on: []
files_modified:
  - strcmp95.go
  - dispatch_strcmp95.go
  - strcmp95_test.go
  - strcmp95_bench_test.go
  - strcmp95_fuzz_test.go
  - props_test.go
  - example_test.go
  - algoid_test.go
  - algorithms_golden_test.go
  - testdata/golden/_staging/strcmp95.json
  - testdata/fuzz/FuzzStrcmp95Score/seed-001
  - tests/bdd/features/strcmp95.feature
  - tests/bdd/steps/algorithms_steps.go
autonomous: true
requirements:
  - CHAR-07
tags: [strcmp95, winkler-1994, similar-character-table, no-init, jaro-winkler-hierarchy, ascii-only, no-runes-variant, census-bureau-cross-validation, dispatch-registration, property-tests, fuzz, benchmark, bdd, staging-golden]

must_haves:
  truths:
    # Goal-backward truths (CHAR-07; ROADMAP §"Remaining Character & Gestalt" success criteria #1)
    - "A caller can `import fuzzymatch` and call Strcmp95Score(\"MARTHA\", \"MARHTA\") and receive a deterministic float64 in [0.0, 1.0] (Winkler 1990 canonical pair)"
    - "Strcmp95Score(\"\", \"\") == 1.0 (both-empty identity, per CONTEXT.md §2 edge cases)"
    - "Strcmp95Score(\"\", \"abc\") == 0.0 (one-empty)"
    - "Strcmp95Score(x, x) == 1.0 for every non-empty x (identity short-circuit at function entry)"
    - "Strcmp95Score(a, b) == Strcmp95Score(b, a) for every (a, b) — byte-path symmetry holds (Winkler 1994 algorithm is symmetric)"
    - "Strcmp95Score(a, b) >= JaroWinklerScore(a, b) for every (a, b) — the four Winkler 1994 adjustments only ADD (RESEARCH.md Pitfall 1 warning sign #3)"
    - "On the Census Bureau canonical pairs DWAYNE/DUANE and DIXON/DICKSONX, Strcmp95Score differs from JaroWinklerScore (proves the similar-character table is firing — RESEARCH.md Pitfall 1 warning sign #2)"
    - "Strcmp95Score on long-prefix pair HAMINGTON/HAMMINGTON exceeds JaroWinklerScore (long-string adjustment fires when min(la,lb) > 4 and prefix conditions met — RESEARCH.md Pitfall 5)"
    - "dispatch[AlgoStrcmp95] (slot 5) is non-nil after package load and equals Strcmp95Score (registered via `var _ = func() bool {...}()` in dispatch_strcmp95.go; NO init())"
    # Source-origin discipline (CONTEXT.md §2; algorithm-licensing-standards)
    - "strcmp95.go's file-level godoc cites Winkler 1994 TR-2 §3 as the PRIMARY source and Census Bureau strcmp95.c as the cross-validation reference (consulted only for reference vectors; not for code/table content) — the source-origin statement block per RESEARCH.md Pattern 1 is present verbatim"
    - "strcmp95SimilarChars is declared as a package-level `var [...]struct{a, b byte; sim float64}{}` with exactly 36 entries; every entry has sim == 0.3; the 36 pairs are the canonical Winkler 1994 list (AE, AI, AO, AU, BV, EI, EO, EU, IO, IU, OU, IY, EY, CG, EF, WU, WV, XK, SZ, XS, QC, UV, MN, LI, QO, PR, IJ, 2Z, 5S, 8B, 1I, 1L, 0O, 0Q, CK, GJ)"
    - "strcmp95.go contains NO init() function — determinism-reviewer flags any init() in this file as BLOCKING per PITFALLS §14"
    - "No `*Runes` variant exists — public surface is exactly one function (Strcmp95Score). Godoc directs Unicode users at `fuzzymatch.Normalise` upstream"
    # Determinism + performance (DET-04, DET-06, PERF-01)
    - "Strcmp95Score never returns NaN, +Inf, -Inf, or -0 for any input — verified by TestProp_Strcmp95Score_NoNaN / _NoInf / _NoNegativeZero in props_test.go"
    - "Strcmp95Score is deterministic across 1000 sequential calls on the same input (TestProp_Strcmp95Score_DeterministicAcrossRuns — PITFALLS §14 closure)"
    - "No math.Pow / math.Log / math.Exp / math.FMA used anywhere in strcmp95.go (only `+`, `-`, `*`, `/`, comparisons, and `float64()` casts) — determinism-standards §13.3 gate"
    - "Score normalisation uses explicit left-to-right parenthesisation per DET-06; sum reductions left-to-right"
    - "BenchmarkStrcmp95Score_ASCII_Short reports 0 B/op, 0 allocs/op for short input (MARTHA/MARHTA fits within match-flag stack budget)"
    # Public-surface + meta-test discipline
    - "Public surface: exactly one new exported symbol (Strcmp95Score) — pre-existing AlgoStrcmp95 constant already in algoid.go slot 5"
    - "FuzzStrcmp95Score is panic-free, score-in-[0,1] for all inputs including invalid UTF-8 (\\xff\\xfe), both-empty, identity, Census-Bureau pairs, and long-prefix pair (HAMINGTON/HAMMINGTON)"
    - "testdata/fuzz/FuzzStrcmp95Score/seed-001 exists in `go test fuzz v1` literal corpus format (one `string(...)` line per fuzz parameter; format matches testdata/fuzz/FuzzSmithWatermanGotohScore/seed-001 byte-for-byte structure)"
    - "tests/bdd/features/strcmp95.feature exists with at minimum: canonical reference-vector Scenario Outline, identity scenario, both-empty scenario, one-empty scenario, symmetry scenario, AND a similar-character-table-fires scenario (Strcmp95 > JaroWinkler on a documented pair)"
    - "tests/bdd/steps/algorithms_steps.go appends Strcmp95 step bindings (iComputeTheStrcmp95ScoreBetween / iComputeTheSecondStrcmp95ScoreBetween / bothStrcmp95ScoresShouldBeEqual) and their ctx.Step regex registrations inside InitializeScenario; existing score-regex `(\\d+\\.?\\d*)` is reused (IN-03 closure — no new regex)"
    - "testdata/golden/_staging/strcmp95.json exists, produced by TestGolden_Strcmp95_Staging via assertGoldenStaging; entries sorted alphabetically by Name; includes at minimum Strcmp95_both_empty, Strcmp95_identical, Strcmp95_one_empty, Strcmp95_MARTHA_MARHTA, Strcmp95_DWAYNE_DUANE, Strcmp95_DIXON_DICKSONX, Strcmp95_HAMINGTON_HAMMINGTON"
    - "algoid_test.go contains a new TestDispatch_Strcmp95Registered asserting dispatch[AlgoStrcmp95] is non-nil; the registered map in TestDispatch_UnregisteredSlotsAreNil adds int(AlgoStrcmp95): true"
    - "ExampleStrcmp95Score appended to example_test.go; `// Output:` block matches byte-for-byte via `fmt.Printf(\"%.4f\\n\", ...)`"
    - "Coverage on strcmp95.go ≥ 90%; 100% on the public Strcmp95Score surface"
    - "Apache-2.0 header present on every new .go file (scripts/verify-license-headers.sh exits 0)"
  artifacts:
    - path: "strcmp95.go"
      provides: "Strcmp95Score (one public function); unexported strcmp95SimilarChars table (36 entries as package-level var) and strcmp95SimilarLookup helper"
      min_lines: 180
      contains: "Source: Winkler, W. E. (1994)"
    - path: "dispatch_strcmp95.go"
      provides: "Package-load-time registration of Strcmp95Score into dispatch[AlgoStrcmp95] (slot 5)"
      contains: "dispatch[AlgoStrcmp95] = Strcmp95Score"
    - path: "strcmp95_test.go"
      provides: "Unit tests for identity, both-empty, one-empty, Census Bureau reference vectors (MARTHA/MARHTA, DWAYNE/DUANE, DIXON/DICKSONX), table-invariants (36 entries, all 0.3, no duplicate pairs), Strcmp95 ≥ JaroWinkler hand-pin, and the runtime allocation gate"
    - path: "strcmp95_bench_test.go"
      provides: "Three benchmarks: BenchmarkStrcmp95Score_{ASCII_Short, ASCII_Medium, ASCII_Long} — alloc-asserted with b.ReportAllocs() + var sink anti-DCE"
    - path: "strcmp95_fuzz_test.go"
      provides: "FuzzStrcmp95Score — panic-free, NaN/Inf-free, score-in-[0,1] invariant; programmatic seeds covering Census Bureau pairs, identity, both-empty, one-empty, invalid UTF-8, long-prefix pair"
    - path: "props_test.go"
      provides: "Appended Strcmp95 property-test block: TestProp_Strcmp95Score_{RangeBounds, Identity, Symmetric, NoNaN, NoInf, NoNegativeZero} PLUS Strcmp95-specific TestProp_Strcmp95Score_AtLeastJaroWinkler and TestProp_Strcmp95Score_DeterministicAcrossRuns"
    - path: "example_test.go"
      provides: "Appended ExampleStrcmp95Score runnable godoc example"
    - path: "algorithms_golden_test.go"
      provides: "Appended buildStrcmp95StagingEntries + TestGolden_Strcmp95_Staging — produces _staging/strcmp95.json via assertGoldenStaging; the algorithms-merge slice is NOT updated here (plan 04-05 owns the merge)"
      contains: "TestGolden_Strcmp95_Staging"
    - path: "algoid_test.go"
      provides: "Appended TestDispatch_Strcmp95Registered; updated `registered` map in TestDispatch_UnregisteredSlotsAreNil flipping slot 5 to true"
      contains: "TestDispatch_Strcmp95Registered"
    - path: "testdata/golden/_staging/strcmp95.json"
      provides: "Per-algorithm staging file (Phase-2-locked pattern); sorted by Name; merged into algorithms.json by plan 04-05"
      contains: "Strcmp95_MARTHA_MARHTA"
    - path: "testdata/fuzz/FuzzStrcmp95Score/seed-001"
      provides: "Fuzz seed corpus file in `go test fuzz v1` literal format"
    - path: "tests/bdd/features/strcmp95.feature"
      provides: "Gherkin feature with scenarios: canonical reference-vector outline, identity, both-empty, one-empty, symmetry, similar-character-table-fires"
    - path: "tests/bdd/steps/algorithms_steps.go"
      provides: "Appended Strcmp95 step methods on AlgorithmContext + Strcmp95 ctx.Step registrations inside InitializeScenario"
  key_links:
    - from: "dispatch_strcmp95.go"
      to: "algoid.go (line 95 AlgoStrcmp95 declared at slot 5)"
      via: "package-level `var _ = func() bool { dispatch[AlgoStrcmp95] = Strcmp95Score; return true }()`"
      pattern: "dispatch\\[AlgoStrcmp95\\]\\s*=\\s*Strcmp95Score"
    - from: "strcmp95.go (Strcmp95Score)"
      to: "jaro.go (Jaro core)"
      via: "Strcmp95 layers four adjustments atop a Jaro pass; per CONTEXT.md D-1 the planner picks whether to call an internal jaroBytes helper or re-derive match-flag arrays — recommendation: re-derive (RESEARCH.md OQ-3) to keep Strcmp95 independent of jaro.go's internal layout"
      pattern: "(Jaro|match.flags|matchFlags)"
    - from: "strcmp95.go (strcmp95SimilarChars)"
      to: "Winkler 1994 TR-2 §3 published 36-pair similar-character list"
      via: "Hand-transcribed `var strcmp95SimilarChars = [...]struct{a, b byte; sim float64}{...}` literal — NO init(), values sourced from the paper (per CONTEXT.md §2)"
      pattern: "var strcmp95SimilarChars = \\[\\.\\.\\.\\]struct"
    - from: "tests/bdd/features/strcmp95.feature"
      to: "tests/bdd/steps/algorithms_steps.go (Strcmp95 step bindings)"
      via: "godog regex registration `^I compute the Strcmp95 score between \"([^\"]*)\" and \"([^\"]*)\"$`"
      pattern: "ctx\\.Step\\(.+Strcmp95"
---

<objective>
Implement Strcmp95 (CHAR-07) — Winkler's 1994 enhancement of Jaro-Winkler with the similar-character table and four stacked adjustments (similar-character credit, Winkler prefix boost, long-string adjustment, AS/I-S/RS-RB letter-pair adjustments). The algorithm is ASCII-only (no `*Runes` variant per CONTEXT.md §2). The similar-character table is hand-transcribed from Winkler 1994 TR-2 §3 into a package-level `var` with NO init() (PITFALLS §14). Cross-validation against the Census Bureau strcmp95.c reference vectors. Full Phase 2 quality bar: unit + property + fuzz + bench + BDD + staging golden + dispatch + example.

Purpose: complete the surname-/record-linkage tier of the character-similarity catalogue with the canonical Winkler-1994 algorithm, layered atop Jaro-Winkler. Provides consumers with the JaroScore → JaroWinklerScore → Strcmp95Score hierarchy that record-linkage / surname-matching workloads expect.

Output: 13 new/modified files (5 new source/test, 8 extensions to existing append-only files); single new dispatch slot wired; first per-algorithm staging golden committed (merged into algorithms.json by plan 04-05).
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
@.planning/phases/02-core-character-algorithms-six/02-CONTEXT.md
@.planning/phases/02-core-character-algorithms-six/02-PATTERNS.md
@.claude/skills/algorithm-correctness-standards/SKILL.md
@.claude/skills/algorithm-licensing-standards/SKILL.md
@.claude/skills/determinism-standards/SKILL.md
@.claude/skills/performance-standards/SKILL.md
@.claude/skills/go-coding-standards/SKILL.md
@.claude/skills/go-testing-standards/SKILL.md
@jarowinkler.go
@jaro.go
@dispatch_swg.go
@algoid.go
</context>

<interfaces>
<!-- Key types/functions executor MUST use without rediscovering. Extracted from existing source. -->

From algoid.go (slot 5 already declared, do NOT modify the declaration):
```go
const AlgoStrcmp95 AlgoID = ...  // slot 5; existing declaration at algoid.go:95
```

From jaro.go (existing public + unexported surface — Strcmp95 may call internally OR re-derive per D-1):
```go
func JaroScore(a, b string) float64
func JaroScoreRunes(a, b string) float64
// jaroBytes is unexported; signature returns only the score, not match-flag arrays.
```

From jarowinkler.go (existing public symbol Strcmp95 must dominate per the AtLeastJaroWinkler invariant):
```go
func JaroWinklerScore(a, b string) float64
func JaroWinklerScoreRunes(a, b string) float64
```

Public surface to be created by this plan (one symbol):
```go
// Strcmp95Score returns Winkler's 1994 Strcmp95 similarity in [0.0, 1.0].
// Strcmp95 = Jaro + similar-character credit + Winkler prefix boost
//          + long-string adjustment + AS/I-S/RS-RB letter-pair adjustments.
// ASCII-only — no Strcmp95ScoreRunes variant (CONTEXT.md §2).
func Strcmp95Score(a, b string) float64
```

Unexported helpers internal to strcmp95.go:
```go
var strcmp95SimilarChars = [...]struct{ a, b byte; sim float64 }{ /* 36 entries from Winkler 1994 TR-2 §3 */ }
func strcmp95SimilarLookup(a, b byte) float64   // returns 0.3 if (a,b) or (b,a) in table; else 0.0
```

Dispatch wiring shape (matches dispatch_swg.go verbatim):
```go
var _ = func() bool { dispatch[AlgoStrcmp95] = Strcmp95Score; return true }()
```
</interfaces>

<tasks>

<task type="auto" tdd="true">
  <name>Task 1: Implement strcmp95.go (algorithm + table + dispatch + minimal unit tests)</name>
  <files>strcmp95.go, dispatch_strcmp95.go, strcmp95_test.go, testdata/golden/_staging/strcmp95.json, algorithms_golden_test.go, algoid_test.go, example_test.go</files>
  <read_first>
    - strcmp95.go (current state — confirm it does NOT exist; creating new)
    - .planning/phases/04-remaining-character-gestalt/04-CONTEXT.md §2 (Strcmp95 locked decisions)
    - .planning/phases/04-remaining-character-gestalt/04-RESEARCH.md — Pattern 1 (Strcmp95 building atop Jaro), Code Examples (lines 696–735 for the table declaration), Pitfall 1 (transcription typos), Pitfall 5 (long-string adjustment conditions), Open Questions OQ-3 (D-1 re-derive vs reuse)
    - .planning/phases/04-remaining-character-gestalt/04-PATTERNS.md §"strcmp95.go", §"dispatch_strcmp95.go", §"strcmp95_test.go", §"License + Source Header", §"No init() / var-only tables"
    - .planning/phases/04-remaining-character-gestalt/04-VALIDATION.md (per-task verification rows 04-01-01..04-01-05, 04-01-09)
    - jarowinkler.go (analog for file-shape; godoc style)
    - jaro.go (Strcmp95 layers atop Jaro — read to understand D-1 trade-off)
    - swg.go lines 1–82 (file-level multi-source attribution block — copy structure verbatim, substitute Winkler 1994)
    - swg_test.go lines 1–101 + 384–409 (reference-vector test shape + AllocsPerRun runtime alloc-gate shape)
    - dispatch_swg.go (full file — exact template for dispatch_strcmp95.go)
    - algoid.go (AlgoStrcmp95 declared at slot 5 — line 95)
    - algorithms_golden_test.go lines 584–656 (buildSWGStagingEntries + TestGolden_SWG_Staging — exact template for buildStrcmp95StagingEntries + TestGolden_Strcmp95_Staging)
    - algoid_test.go lines 284–323 (TestDispatch_SmithWatermanGotohRegistered + TestDispatch_UnregisteredSlotsAreNil — template + slot-map extension point)
    - example_test.go lines 108–122 (ExampleSmithWatermanGotohScore — template for ExampleStrcmp95Score)
  </read_first>
  <behavior>
    - Strcmp95Score("MARTHA", "MARHTA") returns Winkler-1990-compatible score (≈ 0.9611; tolerance 1e-3) — adjustments may or may not fire on this pair, agnostic to the test value within tolerance
    - Strcmp95Score("DWAYNE", "DUANE") ≈ 0.840 (Census Bureau — adjustments fire)
    - Strcmp95Score("DIXON", "DICKSONX") ≈ 0.8133 (Census Bureau)
    - Strcmp95Score("", "") == 1.0
    - Strcmp95Score("", "abc") == 0.0
    - Strcmp95Score("abc", "abc") == 1.0 (identity short-circuit)
    - Strcmp95Score(a, b) == Strcmp95Score(b, a) for any (a, b) — symmetric byte path
    - Strcmp95Score(a, b) >= JaroWinklerScore(a, b) for any (a, b) — adjustments only ADD (asserted on the three Census Bureau pairs in this task; full property-test in Task 2)
    - Strcmp95Score("HAMINGTON", "HAMMINGTON") > JaroWinklerScore("HAMINGTON", "HAMMINGTON") — long-string adjustment fires (RESEARCH.md Pitfall 5)
    - Strcmp95Score("AB", "AC") == JaroWinklerScore("AB", "AC") — long-string adjustment does NOT fire (length ≤ 4 disables it)
    - dispatch[AlgoStrcmp95] non-nil after package load (TestDispatch_Strcmp95Registered)
    - strcmp95SimilarChars has exactly 36 entries, every sim == 0.3, no duplicate (a,b) or (b,a) pair (TestStrcmp95_TableInvariants — load-bearing for RESEARCH.md Pitfall 1)
    - TestStrcmp95_ZeroAllocs_ASCII_Short via testing.AllocsPerRun(100, ...) reports 0 allocs (or documented small bound) for MARTHA/MARHTA
  </behavior>
  <action>
    Create strcmp95.go per PATTERNS.md §"strcmp95.go" and RESEARCH.md Pattern 1 + Code Examples (lines 696–735). File order:
    (a) Apache-2.0 header (copy normalise.go lines 1–13 verbatim — verify-license-headers.sh gate).
    (b) File-level doc block: cite Winkler 1994 TR-2 §3 as the PRIMARY source; cite Census Bureau strcmp95.c (public domain) as cross-validation source ONLY; cite OpenRefine Strcmp95.java (Apache-2.0) for prose-level tie-breaks ONLY; include the explicit source-origin statement block per algorithm-licensing-standards (Primary / Cross-validation / Tie-break / GPL-LGPL: none / Code copied: none). Include the JaroScore → JaroWinklerScore → Strcmp95Score hierarchy paragraph from RESEARCH.md "Specifics".
    (c) `package fuzzymatch`.
    (d) `var strcmp95SimilarChars = [...]struct{a, b byte; sim float64}{ ...36 entries... }` — pairs verbatim from RESEARCH.md Code Examples (AE, AI, AO, AU, BV, EI, EO, EU, IO, IU, OU, IY, EY, CG, EF, WU, WV, XK, SZ, XS, QC, UV, MN, LI, QO, PR, IJ, 2Z, 5S, 8B, 1I, 1L, 0O, 0Q, CK, GJ — every sim == 0.3). Add godoc explaining the var-not-init discipline (PITFALLS §14).
    (e) `func strcmp95SimilarLookup(a, b byte) float64` — linear scan over the 36 entries; return 0.3 if (a==t.a && b==t.b) OR (a==t.b && b==t.a), else 0.0.
    (f) `func Strcmp95Score(a, b string) float64` — open with the godoc directive per RESEARCH.md Pattern 1 (the hierarchy paragraph + the ASCII-only directive + edge cases + the Strcmp95 >= JaroWinkler invariant). Body: identity short-circuit `if a == b { return 1.0 }`; both-empty → 1.0; one-empty → 0.0; then per CONTEXT.md D-1 RE-DERIVE the Jaro match-flag arrays inline (recommendation per RESEARCH.md OQ-3 — keeps Strcmp95 independent of jaro.go's internal layout). Apply the four Winkler 1994 adjustments in order: (1) base Jaro pass over match-flag arrays; (2) similar-character credit pass (modify the match count using strcmp95SimilarLookup); (3) Winkler prefix boost (when J >= 0.7, with prefix cap 4 and scale 0.1); (4) long-string adjustment per the THREE conditions documented in RESEARCH.md Pitfall 5 (min(la,lb) > 4 AND Num_com > i+1 AND 2·Num_com >= minLen+i; document inline with a worked example). Use only `+`, `-`, `*`, `/`, comparisons — NO `math.Pow`/`math.Log`/`math.Exp`/`math.FMA`. Score normalisation uses explicit left-to-right parenthesisation per DET-06.

    Create dispatch_strcmp95.go per PATTERNS.md §"dispatch_strcmp95.go" — full file pattern (~22 lines) substituting `SmithWatermanGotoh` → `Strcmp95` throughout. NO init().

    Create strcmp95_test.go with:
    - Apache-2.0 header.
    - `package fuzzymatch_test`; imports `testing` + `github.com/axonops/fuzzymatch`.
    - TestStrcmp95_BothEmpty, TestStrcmp95_OneEmpty, TestStrcmp95_Identical (copy swg_test.go lines 38–102 shape).
    - TestStrcmp95_ReferenceVectors_CensusBureau — table-driven `t.Run(tt.a+"_"+tt.b, ...)` over Census Bureau / Winkler 1990 pairs (MARTHA/MARHTA, DWAYNE/DUANE, DIXON/DICKSONX) with tolerance 1e-3 (per PATTERNS.md). Include AT LEAST one input where the similar-character table provably fires (Strcmp95Score != JaroWinklerScore by a documented amount) to lock RESEARCH.md Pitfall 1 warning sign #2.
    - TestStrcmp95_TableInvariants — assert len(strcmp95SimilarChars) == 36, every entry has sim == 0.3, no duplicate pairs. Use the export_test.go re-export pattern if the table is unexported (look at existing export_test.go — extend with `var Strcmp95SimilarCharsForTest = strcmp95SimilarChars` if needed, or run the assertions indirectly via call-the-function-with-each-pair smoke test if export_test.go is locked).
    - TestStrcmp95_AtLeastJaroWinkler_OnReferenceVectors — three Census Bureau pairs, assert Strcmp95Score >= JaroWinklerScore on each (full property version lands in Task 2).
    - TestStrcmp95_LongStringAdjustment_Triggers — pin Strcmp95(HAMINGTON, HAMMINGTON) > JaroWinklerScore (adjustment fires); Strcmp95(AB, AC) == JaroWinklerScore (adjustment does NOT fire). RESEARCH.md Pitfall 5 closure.
    - TestStrcmp95_ZeroAllocs_ASCII_Short — testing.AllocsPerRun(100, ...) on MARTHA/MARHTA, fail if allocs > 0 (or documented small bound; copy swg_test.go lines 384–409 shape).
    - Stdlib testing only — NO testify.

    Create testdata/golden/_staging/strcmp95.json by appending buildStrcmp95StagingEntries + TestGolden_Strcmp95_Staging to algorithms_golden_test.go per PATTERNS.md §"algorithms_golden_test.go" (lines 1284–1305 template). Entries (Name sorted alphabetically): Strcmp95_DIXON_DICKSONX, Strcmp95_DWAYNE_DUANE, Strcmp95_HAMINGTON_HAMMINGTON, Strcmp95_MARTHA_MARHTA, Strcmp95_both_empty, Strcmp95_identical, Strcmp95_one_empty. Run `go test -run TestGolden_Strcmp95_Staging -update ./...` to materialise the staging file via assertGoldenStaging → CanonicalMarshalForTest. Do NOT extend the algorithms-merge list — plan 04-05 owns that.

    Append TestDispatch_Strcmp95Registered to algoid_test.go (copy lines 284–292 shape). Extend the `registered` map in TestDispatch_UnregisteredSlotsAreNil (lines 299–323) — add `int(AlgoStrcmp95): true`.

    Append ExampleStrcmp95Score to example_test.go (copy lines 108–122 shape). `fmt.Printf("%.4f\n", fuzzymatch.Strcmp95Score("MARTHA", "MARHTA"))`. The `// Output:` block must match byte-for-byte — capture the actual output by running `go test -run ExampleStrcmp95Score ./...` once and paste into the example.
  </action>
  <verify>
    <automated>cd /Users/johnny/Development/fuzzymatch && go build ./... && go test -run 'TestStrcmp95|TestDispatch_Strcmp95Registered|TestDispatch_UnregisteredSlotsAreNil|TestGolden_Strcmp95_Staging|ExampleStrcmp95Score' ./... && bash scripts/verify-license-headers.sh && ! grep -q "^func init" strcmp95.go && grep -q "var strcmp95SimilarChars" strcmp95.go && grep -q "// Source: Winkler, W. E. (1994)" strcmp95.go</automated>
  </verify>
  <done>
    All Strcmp95* unit tests, TestDispatch_Strcmp95Registered, TestDispatch_UnregisteredSlotsAreNil, TestGolden_Strcmp95_Staging, and ExampleStrcmp95Score pass. License headers green. NO init() in strcmp95.go. Table declaration present with Winkler 1994 source comment. Staging golden file exists, alphabetically sorted, canonical-marshalled.
  </done>
</task>

<task type="auto" tdd="true">
  <name>Task 2: Strcmp95 property tests + benchmarks + fuzz</name>
  <files>props_test.go, strcmp95_bench_test.go, strcmp95_fuzz_test.go, testdata/fuzz/FuzzStrcmp95Score/seed-001</files>
  <read_first>
    - strcmp95.go (created in Task 1 — read to understand the Strcmp95Score surface)
    - props_test.go lines 737–909 (SWG property-test block — 6 standard invariants + 3 SWG-specific extensions; exact template for the Strcmp95 append)
    - swg_bench_test.go (148 lines — analog for the bench file)
    - swg_fuzz_test.go (122 lines — analog for the fuzz harness; collapse the 6-surface loop to single-surface)
    - testdata/fuzz/FuzzSmithWatermanGotohScore/seed-001 (`go test fuzz v1` literal-corpus format)
    - .planning/phases/04-remaining-character-gestalt/04-PATTERNS.md §"strcmp95_bench_test.go", §"strcmp95_fuzz_test.go", §"props_test.go (extend in plans 04-01, 04-02, 04-03)", §"var sink", §"License + Source Header"
    - .planning/phases/04-remaining-character-gestalt/04-RESEARCH.md — Pitfall 1 (table transcription typos + AtLeastJaroWinkler property), Pitfall 5 (long-string adjustment), Pitfall 14 (DeterministicAcrossRuns)
    - .planning/phases/04-remaining-character-gestalt/04-VALIDATION.md (rows 04-01-03, 04-01-04, 04-01-06, 04-01-07)
  </read_first>
  <behavior>
    - TestProp_Strcmp95Score_RangeBounds: testing/quick over arbitrary string inputs → score always in [0.0, 1.0]
    - TestProp_Strcmp95Score_Identity: testing/quick → Score(x, x) == 1.0 for non-empty x
    - TestProp_Strcmp95Score_Symmetric: testing/quick → Score(a, b) == Score(b, a) (byte path)
    - TestProp_Strcmp95Score_NoNaN, _NoInf, _NoNegativeZero: testing/quick → no NaN / Inf / -0 output
    - TestProp_Strcmp95Score_AtLeastJaroWinkler: testing/quick over arbitrary (a, b) → Strcmp95Score(a, b) >= JaroWinklerScore(a, b) (the algorithm hierarchy invariant — RESEARCH.md Pitfall 1 warning sign #3)
    - TestProp_Strcmp95Score_DeterministicAcrossRuns: same input → byte-identical output across 1000 sequential calls (PITFALLS §14 closure)
    - BenchmarkStrcmp95Score_ASCII_Short: MARTHA/MARHTA → 0 allocs/op (b.ReportAllocs() gate)
    - BenchmarkStrcmp95Score_ASCII_Medium / _ASCII_Long: alloc-asserted; document expected allocations
    - FuzzStrcmp95Score: panic-free, NaN/Inf-free, score-in-[0,1] for arbitrary (a, b) including invalid UTF-8 — minimum 60s harness run
  </behavior>
  <action>
    Extend props_test.go by appending a new sectioned block at end-of-file per PATTERNS.md template:
    // ---------------------------------------------------------------------------
    // Strcmp95 property tests (plan 04-01)
    // ---------------------------------------------------------------------------
    Add seven property tests (six standard + one algorithm-specific): TestProp_Strcmp95Score_RangeBounds, _Identity, _Symmetric, _NoNaN, _NoInf, _NoNegativeZero (copy lines 737–810 shape verbatim, substituting `SmithWatermanGotoh` → `Strcmp95`), AND TestProp_Strcmp95Score_AtLeastJaroWinkler (loop over testing/quick-generated (a, b) and assert `Strcmp95Score(a, b) >= JaroWinklerScore(a, b)`), AND TestProp_Strcmp95Score_DeterministicAcrossRuns (call Strcmp95Score 1000 times with the same input — e.g. MARTHA/MARHTA — and assert all 1000 results are byte-identical to the first). Default 100 testing/quick iterations; explicit `quick.Check` calls. NO Runes variant (CONTEXT.md §2).

    Create strcmp95_bench_test.go per PATTERNS.md §"strcmp95_bench_test.go" — header + three benches: BenchmarkStrcmp95Score_ASCII_Short ("MARTHA" / "MARHTA"), _ASCII_Medium (~50-char realistic surname pair, e.g. "HAMINGTONSWORTH" / "HAMMINGTONSWORTH"), _ASCII_Long (~200+-char pair). Pattern: `b.ReportAllocs(); b.ResetTimer(); var sink float64; for i := 0; i < b.N; i++ { sink = fuzzymatch.Strcmp95Score(a, b) }; if sink < 0 { b.Fatal(...) }`. DROP Unicode_Short / WithParams / RawScore variants (CONTEXT.md §2 — no Runes; no Params; no Raw).

    Create strcmp95_fuzz_test.go per PATTERNS.md §"strcmp95_fuzz_test.go" — single-surface FuzzStrcmp95Score with seed corpus from RESEARCH.md (MARTHA/MARHTA, DWAYNE/DUANE, "abc"/"abc", ""/"abc", ""/"", "\xff\xfe"/"abc", "\xc0\x80"/"abc", HAMINGTON/HAMMINGTON, "AB"/"AC"). Fuzz body asserts no panic, no NaN (math.IsNaN), no Inf (math.IsInf), score in [0, 1].

    Create testdata/fuzz/FuzzStrcmp95Score/seed-001 in `go test fuzz v1` literal format (3 lines: header + 2 `string(...)` lines for the (a, b) parameters); use MARTHA/MARHTA as the canonical seed. Format MUST be byte-stable per Phase 3 IN-06 closure — do not introduce extra whitespace.
  </action>
  <verify>
    <automated>cd /Users/johnny/Development/fuzzymatch && go test -run 'TestProp_Strcmp95' ./... && go test -bench=BenchmarkStrcmp95Score -benchmem -benchtime=1x ./... && go test -fuzz=FuzzStrcmp95Score -fuzztime=10s ./... && head -1 testdata/fuzz/FuzzStrcmp95Score/seed-001 | grep -q "^go test fuzz v1$"</automated>
  </verify>
  <done>
    All TestProp_Strcmp95Score_* property tests pass under quick.Check (100 iterations default). Bench file produces 3 benches with 0 B/op, 0 allocs/op for ASCII_Short. Fuzz harness 10s smoke run completes without panic/NaN/Inf/out-of-range. On-disk seed file present with byte-stable format.
  </done>
</task>

<task type="auto" tdd="true">
  <name>Task 3: Strcmp95 BDD feature + steps</name>
  <files>tests/bdd/features/strcmp95.feature, tests/bdd/steps/algorithms_steps.go</files>
  <read_first>
    - tests/bdd/features/swg.feature (46 lines — exact template)
    - tests/bdd/steps/algorithms_steps.go lines 290–316 (SWG step methods) + lines 341–345 (existing score-regex registrations `(\d+\.?\d*)` — IN-03 closure) + lines 443–455 (SWG step regex registrations inside InitializeScenario)
    - .planning/phases/04-remaining-character-gestalt/04-PATTERNS.md §"tests/bdd/features/{strcmp95,lcsstr,ratcliff_obershelp}.feature", §"tests/bdd/steps/algorithms_steps.go"
    - .planning/phases/04-remaining-character-gestalt/04-RESEARCH.md — Pitfall 7 (BDD score regex must accept integer-form `0` / `1` per IN-03 closure)
    - .planning/phases/04-remaining-character-gestalt/04-VALIDATION.md (row 04-01-08)
  </read_first>
  <behavior>
    - godog runs and passes all Strcmp95 scenarios in tests/bdd/features/strcmp95.feature
    - All scenarios use the existing `(\d+\.?\d*)` score regex; no new regex registrations needed
    - At least one scenario asserts Strcmp95 > JaroWinkler on a similar-character-table-fires input pair (locks Pitfall 1 warning sign #2)
  </behavior>
  <action>
    Create tests/bdd/features/strcmp95.feature per PATTERNS.md §"tests/bdd/features/...feature" — copy swg.feature 46-line structure. Required scenarios:
    - Comment header: `# Primary source: Winkler, W. E. (1994). "Advanced methods for record linkage." ASA: 467-472.`
    - `Feature: Strcmp95 similarity` with a one-paragraph description.
    - `Scenario Outline: canonical reference vectors` with Examples table (Census Bureau pairs: MARTHA/MARHTA, DWAYNE/DUANE, DIXON/DICKSONX). Tolerance 0.0001 per the existing approximately-step regex.
    - `Scenario: identical strings score 1.0` (use "user_id" / "user_id").
    - `Scenario: both-empty strings score 1.0`.
    - `Scenario: one-empty string scores 0.0` ("abc" / "").
    - `Scenario: score is symmetric` (compute kitten/sitting and sitting/kitten via the second-score step; assert equality).
    - `Scenario: similar-character table fires` — compute Strcmp95("DWAYNE", "DUANE") AND JaroWinklerScore("DWAYNE", "DUANE"); assert Strcmp95 score > JaroWinkler score via a new step OR use the approximately-step against the documented pinned value (e.g. Strcmp95 ≈ 0.840, JaroWinkler ≈ 0.812 — both pinned). If a new step is needed, register `^I compute the JaroWinkler score between "([^"]*)" and "([^"]*)" for comparison$` in algorithms_steps.go and add a new "the Strcmp95 score should exceed the JaroWinkler score" step; recommended approach is the simpler "two approximately-pinned scores" pattern (no new step).

    Extend tests/bdd/steps/algorithms_steps.go by appending the Strcmp95 step-method block per PATTERNS.md template (three methods: iComputeTheStrcmp95ScoreBetween, iComputeTheSecondStrcmp95ScoreBetween, bothStrcmp95ScoresShouldBeEqual). Register their regexes inside InitializeScenario alongside the existing SWG registrations (lines 443–455 area). DO NOT alter the existing `(\d+\.?\d*)` approximately-step regex — it already accepts integer-form per IN-03 closure. testify IS permitted in this file (sub-module) — but the existing SWG pattern uses `fmt.Errorf`; stay consistent.
  </action>
  <verify>
    <automated>cd /Users/johnny/Development/fuzzymatch && make test-bdd 2>&1 | grep -i 'strcmp95' && cd tests/bdd && go test -run 'Strcmp95|Test' ./...</automated>
  </verify>
  <done>
    `make test-bdd` exits 0 with the new Strcmp95 scenarios green. The feature file covers identity, both-empty, one-empty, canonical reference vectors, symmetry, and similar-character-table-fires. No regex drift; existing score-regex reused.
  </done>
</task>

</tasks>

<threat_model>
## Trust Boundaries

| Boundary | Description |
|----------|-------------|
| caller → Strcmp95Score | Untrusted (a, b string) input crosses the API surface; library is pure-function with no I/O, auth, or session state |

## STRIDE Threat Register (ASVS Level 1)

| Threat ID | Category | Component | Disposition | Mitigation Plan |
|-----------|----------|-----------|-------------|-----------------|
| T-fuzz-panic | D (Denial of Service via panic on malformed input) | Strcmp95Score on invalid UTF-8 / lone-surrogate / extreme-length inputs | mitigate | Task 2 ships FuzzStrcmp95Score with ≥ 60s harness budget and on-disk seed corpus covering `\xff\xfe`, `\xc0\x80`, identity, both-empty, one-empty, Census Bureau pairs, and the long-prefix HAMINGTON/HAMMINGTON pair. Fuzz body asserts no panic, no NaN, no Inf, score-in-[0,1]. Validation map row 04-01-06 |
| T-complexity-attack | D (Denial of Service via algorithmic complexity) | Strcmp95Score on pathological extreme-length inputs | accept | Strcmp95 is O(la·w) where w is the Jaro match window — bounded; the four adjustments are O(min(la,lb)). PERF-01 documents the worst-case budget; long-input benches in Task 2 establish the regression baseline. No new mitigation beyond the documented godoc complexity note. Pure-function library — caller controls input size |
| T-float-determinism | T (Tampering of float reduction order across architectures) | Strcmp95Score sum reductions (Jaro + four adjustments) | mitigate | Explicit left-to-right parenthesisation per DET-06; no math.Pow/Log/Exp/FMA (grep gate in Task 1's verify command); cross-platform CI matrix verifies byte-identical golden output via testdata/golden/_staging/strcmp95.json merged in plan 04-05; TestProp_Strcmp95Score_DeterministicAcrossRuns (Task 2) pins per-process determinism |
</threat_model>

<verification>
- `go build ./...` succeeds.
- `go test -run 'TestStrcmp95|TestProp_Strcmp95|TestDispatch_Strcmp95Registered|TestGolden_Strcmp95_Staging|ExampleStrcmp95Score' ./...` exits 0.
- `go test -bench=BenchmarkStrcmp95Score -benchmem -benchtime=1x ./...` reports 0 B/op, 0 allocs/op for ASCII_Short.
- `go test -fuzz=FuzzStrcmp95Score -fuzztime=60s ./...` completes with no failures (10s smoke OK for the per-task gate; 60s for `/gsd-verify-work`).
- `make test-bdd` green; Strcmp95 scenarios visible in godog output.
- `bash scripts/verify-license-headers.sh` exits 0.
- `bash scripts/verify-no-runtime-deps.sh` exits 0.
- `! grep -q "^func init" strcmp95.go` (no init() per CONTEXT.md §2 + PITFALLS §14).
- `grep -q "Source: Winkler, W. E. (1994)" strcmp95.go` (primary-source citation present).
- `grep -q "var strcmp95SimilarChars" strcmp95.go` (table declared as var, not built in init()).
- `! grep -E "math\\.(Pow|Log|Exp|FMA)" strcmp95.go` (DET-06 gate; only `+`, `-`, `*`, `/` permitted in algorithm float arithmetic).
- `make coverage-check` confirms strcmp95.go ≥ 90% per-file coverage and 100% on the public Strcmp95Score surface.
</verification>

<success_criteria>
- All three tasks complete; all listed verification commands green.
- testdata/golden/_staging/strcmp95.json exists and is canonical-marshalled (no manual edits).
- testdata/fuzz/FuzzStrcmp95Score/seed-001 exists in byte-stable `go test fuzz v1` literal format.
- Public surface is exactly one new exported function (Strcmp95Score); pre-existing AlgoStrcmp95 constant unchanged.
- Dispatch slot 5 wired; TestDispatch_UnregisteredSlotsAreNil updated to flip slot 5 to true.
- Phase 4 plan 04-02 (LCSStr) can begin without further blockers.
</success_criteria>

<output>
After completion, create `.planning/phases/04-remaining-character-gestalt/04-01-strcmp95-SUMMARY.md` per the GSD summary template.
</output>
