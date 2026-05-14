---
phase: 05-q-gram-algorithms
plan: 02
type: execute
wave: 2
depends_on:
  - 05-01-qgram-foundation-jaccard
files_modified:
  - sorensen_dice.go
  - dispatch_sorensen_dice.go
  - sorensen_dice_test.go
  - sorensen_dice_bench_test.go
  - sorensen_dice_fuzz_test.go
  - props_test.go
  - example_test.go
  - algoid_test.go
  - algorithms_golden_test.go
  - testdata/golden/_staging/sorensen_dice.json
  - testdata/fuzz/FuzzSorensenDiceScore/seed-001
  - testdata/fuzz/FuzzSorensenDiceScoreRunes/seed-001
  - tests/bdd/features/sorensen_dice.feature
  - tests/bdd/steps/algorithms_steps.go
  - llms.txt
  - llms-full.txt
autonomous: true
requirements:
  - QGRAM-03
tags: [sorensen-dice, dice-1945, sorensen-1948, byte-and-rune-paths, dispatch-registration, property-tests, fuzz, benchmark, bdd, staging-golden, llms-sync]

must_haves:
  truths:
    # Sørensen-Dice algorithm (QGRAM-03)
    - "A caller can `import fuzzymatch` and call SorensenDiceScore(\"night\", \"nacht\", 2) and receive `2·1/(4+4) = 0.25` (RV-D1 — canonical NLP-textbook bigram pair)"
    - "SorensenDiceScore(\"abcdef\", \"bcdefg\", 2) == 0.8 (RV-D2 — Dice 1945 §3 high-overlap analogue: `2·4/(5+5)`)"
    - "SorensenDiceScore(\"abcdef\", \"abcXef\", 3) == 0.25 (RV-D3 — trigram variant: `2·1/(4+4)`)"
    - "SorensenDiceScore(\"hello\", \"hello\", 2) == 1.0 (RV-D4 — identity)"
    - "SorensenDiceScore(\"\", \"\", 2) == 1.0 (both-empty identity convention)"
    - "SorensenDiceScore(\"\", \"abc\", 2) == 0.0 (one-empty convention)"
    - "SorensenDiceScore(x, x, n) == 1.0 for every non-empty x and every n ≥ 1"
    - "SorensenDiceScore(a, b, n) == SorensenDiceScore(b, a, n) for every (a, b, n) — Sørensen-Dice is symmetric"
    - "SorensenDiceScoreRunes operates on the rune q-gram path consuming `extractQGramsRunes` from plan 05-01"
    - "SorensenDiceScore(\"hello\", \"hello\", 0) panics with message containing `invalid q-gram size`; same for n=-1"
    - "dispatch[AlgoSorensenDice] is non-nil after package load and dispatches to a wrapper that calls SorensenDiceScore(a, b, 3) (default n=3 per CONTEXT.md Deferred §4)"
    # Determinism + correctness gates (DET-03, DET-06)
    - "SorensenDiceScore never returns NaN, +Inf, -Inf, or -0 for any input — verified by TestProp_SorensenDiceScore_NoNaN / _NoInf / _NoNegativeZero (byte + rune)"
    - "SorensenDiceScore is deterministic: 1000 sequential calls on the same input produce byte-identical output (TestProp_SorensenDiceScore_DeterministicAcrossRuns)"
    - "No map iteration on the output path in sorensen_dice.go — intersection cardinality is an integer count, not an ordered slice; output ORDER does not depend on iteration order — DET-03 satisfied"
    - "No `math.Pow`, `math.Log`, `math.Exp`, `math.FMA` anywhere in sorensen_dice.go (only `+`, `-`, `*`, `/`, comparisons, `float64()` casts) — DET-06 gate"
    - "DSC formula uses explicit left-to-right reduction with parenthesisation `(2.0 * float64(intersection)) / (float64(lenA) + float64(lenB))` — DET-06"
    # Public-surface + meta-test discipline
    - "Public surface added by this plan: exactly two new exported symbols (SorensenDiceScore, SorensenDiceScoreRunes) — pre-existing AlgoSorensenDice constant already in algoid.go slot"
    - "FuzzSorensenDiceScore + FuzzSorensenDiceScoreRunes panic-free, score-in-[0,1], NaN/Inf-free for any (a, b) including invalid UTF-8 (\\xff\\xfe); n constrained to [1, 8] via the fuzz body"
    - "testdata/fuzz/FuzzSorensenDiceScore/seed-001 and testdata/fuzz/FuzzSorensenDiceScoreRunes/seed-001 exist in byte-stable `go test fuzz v1` literal format"
    - "tests/bdd/features/sorensen_dice.feature exists with canonical reference-vector Scenario Outline (RV-D1..RV-D4), identity, both-empty, one-empty, symmetry, AND a rune-path Unicode scenario"
    - "tests/bdd/steps/algorithms_steps.go appends SorensenDice step bindings and their ctx.Step regex registrations inside InitializeScenario"
    - "testdata/golden/_staging/sorensen_dice.json exists, produced by TestGolden_SorensenDice_Staging via assertGoldenStaging; entries sorted alphabetically by Name; includes 8-10 entries covering identity, both-empty, one-empty, RV-D1/D2/D3 (byte), and one rune-path entry"
    - "algoid_test.go contains a new TestDispatch_SorensenDiceRegistered; registered map updated to flip AlgoSorensenDice slot to true"
    - "ExampleSorensenDiceScore and ExampleSorensenDiceScoreRunes appended to example_test.go; `// Output:` blocks match byte-for-byte"
    - "llms.txt lists `SorensenDiceScore` and `SorensenDiceScoreRunes` (the two new exported symbols). The AlgoID constant AlgoSorensenDice is ALREADY listed (declared in Phase 1) — no new AlgoID entry needed"
    - "llms-full.txt has parallel entries with one-line rationales"
    - "Coverage on sorensen_dice.go ≥ 90%; 100% on the public SorensenDiceScore + SorensenDiceScoreRunes surface"
    - "Apache-2.0 header present on every new .go file (scripts/verify-license-headers.sh exits 0)"
  artifacts:
    - path: "sorensen_dice.go"
      provides: "SorensenDiceScore + SorensenDiceScoreRunes (two new public functions); cites Dice 1945 + Sørensen 1948 as primary sources"
      min_lines: 100
      contains: "Source: Dice, L. R. (1945)"
    - path: "dispatch_sorensen_dice.go"
      provides: "Package-load-time registration of a default-n=3 SorensenDiceScore wrapper into dispatch[AlgoSorensenDice]"
      contains: "dispatch[AlgoSorensenDice]"
    - path: "sorensen_dice_test.go"
      provides: "Unit tests for identity, both-empty, one-empty, RV-D1/D2/D3/D4 reference vectors (byte + rune), direct-call panic on n < 1, and a runtime allocation gate"
    - path: "sorensen_dice_bench_test.go"
      provides: "Benchmarks: BenchmarkSorensenDiceScore_{ASCII_Short, ASCII_Medium, ASCII_Long} + BenchmarkSorensenDiceScoreRunes_Unicode_Short — alloc-asserted"
    - path: "sorensen_dice_fuzz_test.go"
      provides: "FuzzSorensenDiceScore and FuzzSorensenDiceScoreRunes — panic-free, NaN/Inf-free, score-in-[0,1]"
    - path: "props_test.go"
      provides: "Appended SorensenDice property-test block: TestProp_SorensenDiceScore_{RangeBounds, Identity, Symmetric, NoNaN, NoInf, NoNegativeZero} for BOTH byte and rune surfaces (12 property tests) plus TestProp_SorensenDiceScore_DeterministicAcrossRuns"
    - path: "example_test.go"
      provides: "Appended ExampleSorensenDiceScore + ExampleSorensenDiceScoreRunes runnable godoc examples"
    - path: "algoid_test.go"
      provides: "Appended TestDispatch_SorensenDiceRegistered; updated registered map"
    - path: "algorithms_golden_test.go"
      provides: "Appended buildSorensenDiceStagingEntries + TestGolden_SorensenDice_Staging"
    - path: "testdata/golden/_staging/sorensen_dice.json"
      provides: "Per-algorithm staging file; sorted by Name; merged into algorithms.json by plan 05-05"
      contains: "SorensenDice_night_nacht"
    - path: "testdata/fuzz/FuzzSorensenDiceScore/seed-001"
      provides: "Fuzz seed corpus file in `go test fuzz v1` literal format"
    - path: "testdata/fuzz/FuzzSorensenDiceScoreRunes/seed-001"
      provides: "Fuzz seed corpus file for the rune-path harness"
    - path: "tests/bdd/features/sorensen_dice.feature"
      provides: "Gherkin feature with canonical reference vectors, identity, both-empty, one-empty, symmetry, rune-path Unicode"
    - path: "tests/bdd/steps/algorithms_steps.go"
      provides: "Appended SorensenDice step methods + ctx.Step registrations"
    - path: "llms.txt"
      provides: "Appended 2 function entries (SorensenDiceScore, SorensenDiceScoreRunes)"
    - path: "llms-full.txt"
      provides: "Parallel entries with one-line rationales"
  key_links:
    - from: "sorensen_dice.go (SorensenDiceScore + SorensenDiceScoreRunes)"
      to: "q_gram.go (extractQGrams + extractQGramsRunes — created in plan 05-01)"
      via: "Direct call into the unexported extractor; SorensenDiceScore computes |A∩B| via map-length arithmetic and DSC = 2·|A∩B|/(|QA|+|QB|) — no iteration on output path"
      pattern: "extractQGrams(Runes)?\\("
    - from: "dispatch_sorensen_dice.go"
      to: "algoid.go (AlgoSorensenDice declared at line 117)"
      via: "package-level closure wrapper `dispatch[AlgoSorensenDice] = func(a, b string) float64 { return SorensenDiceScore(a, b, 3) }`"
      pattern: "dispatch\\[AlgoSorensenDice\\]"
---

<objective>
Implement Sørensen-Dice (QGRAM-03) — DSC = 2·|QA∩QB| / (|QA|+|QB|) — atop the shared q-gram infrastructure from plan 05-01. Primary sources: Dice 1945 §3 and Sørensen 1948 §3 (independent rediscoveries of the same coefficient). Both byte and rune surfaces ship together. Full Phase 2/3/4 quality bar (unit + property + fuzz + bench + BDD + staging golden + dispatch + example + llms.txt). RV-D1 ("night"/"nacht"/n=2 → 0.25) is the load-bearing canonical NLP-textbook vector. RV-D3 covers the trigram path; RV-D2 covers a high-overlap pair.

Purpose: ship the textbook q-gram coefficient that consumers building name/word-similarity workflows expect as a direct alternative to Jaccard. DSC weights the intersection higher than Jaccard (2·|∩|/(|QA|+|QB|) vs |∩|/|∪|), which is the right default for many fuzzy-match scenarios.

Output: 16 new/modified files (4 new source/test files, plus extensions to 7 existing append-only files, plus 4 new test/fixture files in testdata + tests/bdd). Single new dispatch slot wired (Wave 2 — depends on plan 05-01 ONLY; parallelisable with plans 05-03 and 05-04).
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
@.planning/phases/04-remaining-character-gestalt/04-PATTERNS.md
@.claude/skills/algorithm-correctness-standards/SKILL.md
@.claude/skills/algorithm-licensing-standards/SKILL.md
@.claude/skills/determinism-standards/SKILL.md
@.claude/skills/performance-standards/SKILL.md
@.claude/skills/go-testing-standards/SKILL.md
@algoid.go
@q_gram.go
@qgram_jaccard.go
@dispatch_qgram_jaccard.go
</context>

<interfaces>
<!-- Key types/functions executor MUST use without rediscovering. -->

From q_gram.go (created in plan 05-01 — read-only here):
```go
func extractQGrams(s string, n int) map[string]int
func extractQGramsRunes(s string, n int) map[string]int
```

From algoid.go (slot already declared at line 117; do NOT modify):
```go
const AlgoSorensenDice AlgoID = ... // line 117; String() case at line 235-236
```

Public surface to be created by this plan:
```go
// SorensenDiceScore returns the Sørensen-Dice coefficient of the
// q-gram multiset of a and b (Dice 1945; Sørensen 1948).
// DSC = 2·|QA∩QB| / (|QA| + |QB|), in [0.0, 1.0].
// Panics on n < 1 with a message containing "invalid q-gram size".
func SorensenDiceScore(a, b string, n int) float64

func SorensenDiceScoreRunes(a, b string, n int) float64
```

Dispatch wiring (matches dispatch_qgram_jaccard.go verbatim — default n=3):
```go
var _ = func() bool {
    dispatch[AlgoSorensenDice] = func(a, b string) float64 {
        return SorensenDiceScore(a, b, 3)
    }
    return true
}()
```
</interfaces>

<tasks>

<task type="auto" tdd="true">
  <name>Task 1: Implement sorensen_dice.go + dispatch + unit tests + staging golden</name>
  <files>sorensen_dice.go, dispatch_sorensen_dice.go, sorensen_dice_test.go, testdata/golden/_staging/sorensen_dice.json, algorithms_golden_test.go, algoid_test.go, example_test.go, llms.txt, llms-full.txt</files>
  <read_first>
    - sorensen_dice.go (current state — confirm it does NOT exist; creating new)
    - q_gram.go (created in plan 05-01 — read to understand extractQGrams + extractQGramsRunes signatures)
    - qgram_jaccard.go (plan 05-01 — exact structural analog; copy the file shape verbatim, substituting Sørensen-Dice formula)
    - dispatch_qgram_jaccard.go (plan 05-01 — exact template for dispatch_sorensen_dice.go)
    - qgram_jaccard_test.go (plan 05-01 — exact structural analog for sorensen_dice_test.go)
    - .planning/phases/05-q-gram-algorithms/05-CONTEXT.md §5 (LOCKED patterns)
    - .planning/phases/05-q-gram-algorithms/05-RESEARCH.md §1.2 (Sørensen-Dice formula), §2.2 (RV-D1..RV-D5 derivations)
    - algoid.go line 117 (AlgoSorensenDice slot)
    - algorithms_golden_test.go (find the buildQGramJaccardStagingEntries function added in plan 05-01 — exact template for buildSorensenDiceStagingEntries)
    - algoid_test.go (last 50 lines — TestDispatch_QGramJaccardRegistered added in plan 05-01; analog)
    - example_test.go (last 30 lines — ExampleQGramJaccardScore from plan 05-01; analog)
    - llms.txt (find the QGramJaccard entries from plan 05-01; analog for the SorensenDice entries)
  </read_first>
  <behavior>
    - SorensenDiceScore("night", "nacht", 2) returns 0.25 exactly (RV-D1)
    - SorensenDiceScore("abcdef", "bcdefg", 2) returns 0.8 exactly (RV-D2)
    - SorensenDiceScore("abcdef", "abcXef", 3) returns 0.25 exactly (RV-D3)
    - SorensenDiceScore("hello", "hello", 2) == 1.0 (RV-D4 identity)
    - SorensenDiceScore("", "", 2) == 1.0 (both-empty)
    - SorensenDiceScore("", "abc", 2) == 0.0 (one-empty)
    - SorensenDiceScore("abc", "", 2) == 0.0 (one-empty symmetric)
    - SorensenDiceScore(a, b, n) == SorensenDiceScore(b, a, n) for arbitrary input (exact equality)
    - SorensenDiceScoreRunes("café", "cafe", 2) == 0.6666666666666666 (rune-bigrams: |QA|=|QB|=3, |∩|=2 (ca, af); DSC = 2·2/(3+3) = 4/6)
    - SorensenDiceScore("hello", "hello", 0) panics with message containing "invalid q-gram size"
    - SorensenDiceScore("hello", "hello", -1) panics with message containing "invalid q-gram size"
    - dispatch[AlgoSorensenDice]("night", "nacht") returns SorensenDiceScore("night", "nacht", 3) (dispatch default n=3 — for this pair: |QA|=3, |QB|=3, |∩|=0 → 0.0)
    - dispatch[AlgoSorensenDice] is non-nil after package load (TestDispatch_SorensenDiceRegistered)
    - TestGolden_SorensenDice_Staging produces testdata/golden/_staging/sorensen_dice.json with 8-10 alphabetically-sorted entries
  </behavior>
  <action>
    Step A — Create sorensen_dice.go per qgram_jaccard.go structural analog. File order:
    (a) Apache-2.0 header.
    (b) File-level doc: cite Dice, L. R. (1945) "Measures of the amount of ecologic association between species" Ecology 26(3):297-302 §3 as PRIMARY; cite Sørensen, T. (1948) "A method of establishing groups of equal amplitude in plant sociology" Kongelige Danske Videnskabernes Selskab 5(4):1-34 as the independent rediscovery (cite §3 specifically). Source-origin statement block (Primary / Cross-validation: hand-derived RV-D1..RV-D4 in sorensen_dice_test.go / Tie-break: none / GPL-LGPL: none / Code copied: none).
    (c) `package fuzzymatch`.
    (d) `func SorensenDiceScore(a, b string, n int) float64` — godoc with the DSC = 2·|QA∩QB| / (|QA|+|QB|) formula, the [0, 1] range, identity / both-empty / one-empty conventions, and the direct-call panic-on-n<1 contract. Body: identity short-circuit `if a == b { return 1.0 }` (covers both-empty + identical); one-empty short-circuit → 0.0; `if n < 1 { panic("fuzzymatch: invalid q-gram size") }`; call extractQGrams(a, n) and extractQGrams(b, n); compute lenA = sum-of-values(qa), lenB = sum-of-values(qb) — multiset sizes (overlapping windows, so |QA| equals max(0, len(a)-n+1), but computing via sum-of-values is correct for both byte and rune paths and lets the helper define semantics); compute intersection cardinality by iterating the SMALLER map and accumulating `min(qa[k], qb[k])` — the OUTPUT is the integer count, not an ordered slice, so DET-03 is satisfied. Note: for the standard Sørensen-Dice formulation over q-gram MULTISETS the intersection cardinality is `sum_k min(qa[k], qb[k])` (not `|qa keys ∩ qb keys|`); document this inline. Compute the result: `dsc := (2.0 * float64(intersection)) / (float64(lenA) + float64(lenB))` — explicit parenthesisation per DET-06. Return dsc. Edge case: if `lenA + lenB == 0` after both extractions (possible when `len(a) < n` AND `len(b) < n` but a != b — though `a == b` short-circuits handles `len(a)=len(b)=0`; the residual case is `len(a) < n` AND `len(b) < n` with a != b), return 1.0 per both-extractions-empty convention. ASSERT: this matches the Q-Gram Jaccard convention from plan 05-01 (both-empty post-extraction → 1.0).
    (e) `func SorensenDiceScoreRunes(a, b string, n int) float64` — analogous, calling extractQGramsRunes.
    (f) Use only `+`, `-`, `*`, `/`, comparisons, `float64()` casts — NO math.Pow/Log/Exp/FMA.

    Step B — Create dispatch_sorensen_dice.go per dispatch_qgram_jaccard.go template. Closure binds n=3. NO init(). Apache-2.0 header.

    Step C — Create sorensen_dice_test.go covering:
    - TestSorensenDice_BothEmpty
    - TestSorensenDice_OneEmpty (both directions)
    - TestSorensenDice_Identical
    - TestSorensenDice_ReferenceVectors — table-driven `t.Run(tt.name, ...)` over RV-D1 ("night"/"nacht"/n=2/0.25), RV-D2 ("abcdef"/"bcdefg"/n=2/0.8), RV-D3 ("abcdef"/"abcXef"/n=3/0.25), RV-D4 ("hello"/"hello"/n=2/1.0). Each row's test comment reproduces the formula derivation from RESEARCH.md §2.2 with the RV-D{N} identifier in the test name. Tolerance: exact equality for the rational values (0.25, 0.8, 1.0 are all bit-exact in float64).
    - TestSorensenDice_Symmetric — assert Score(a, b, n) == Score(b, a, n) on 3 hand-picked pairs (Sørensen-Dice is exactly symmetric — equality)
    - TestSorensenDice_PanicsOnInvalidN — table-driven defer-recover over n=0, n=-1, n=-100. Assert the panic message contains "invalid q-gram size".
    - TestSorensenDiceRunes_CafeReference — "café"/"cafe"/n=2 → 0.6666666666666666 within 1e-15 (derivation in test comment)
    - TestSorensenDice_AllocBound via testing.AllocsPerRun(100, ...) — document the actual alloc bound per RESEARCH.md §4.1 (≤ 4 allocs/op)
    - Stdlib testing only.

    Step D — Append buildSorensenDiceStagingEntries + TestGolden_SorensenDice_Staging to algorithms_golden_test.go. Entries (Name sorted alphabetically): `SorensenDice_abcdef_abcXef_n3` (RV-D3), `SorensenDice_abcdef_bcdefg` (RV-D2), `SorensenDice_both_empty`, `SorensenDice_cafe_runes` (RV-D5-style rune analog), `SorensenDice_identical` (RV-D4), `SorensenDice_night_nacht` (RV-D1), `SorensenDice_no_overlap` ("abc"/"xyz"/n=2 → 0.0), `SorensenDice_one_empty`. Run `go test -run TestGolden_SorensenDice_Staging -update ./...` to materialise the file.

    Step E — Append TestDispatch_SorensenDiceRegistered to algoid_test.go; flip AlgoSorensenDice slot in the registered map of TestDispatch_UnregisteredSlotsAreNil.

    Step F — Append ExampleSorensenDiceScore + ExampleSorensenDiceScoreRunes to example_test.go. Byte path uses "night"/"nacht"/n=2 → 0.2500; rune path uses "café"/"cafe"/n=2 → 0.6667. Capture exact stdout and paste.

    Step G — Update llms.txt + llms-full.txt with the two new function entries.
  </action>
  <verify>
    <automated>cd /Users/johnny/Development/fuzzymatch && go build ./... && go test -run 'TestSorensenDice|TestDispatch_SorensenDiceRegistered|TestDispatch_UnregisteredSlotsAreNil|TestGolden_SorensenDice_Staging|ExampleSorensenDiceScore' ./... && bash scripts/verify-license-headers.sh && ! grep -q "^func init" sorensen_dice.go && grep -q "Source: Dice, L. R. (1945)" sorensen_dice.go && grep -q "dispatch\[AlgoSorensenDice\]" dispatch_sorensen_dice.go && grep -q "SorensenDiceScore" llms.txt && ! grep -v '^#' sorensen_dice.go | grep -E "math\.(Pow|Log|Exp|FMA)"</automated>
  </verify>
  <done>
    All SorensenDice* unit tests, TestDispatch_SorensenDiceRegistered, TestGolden_SorensenDice_Staging, and ExampleSorensenDiceScore pass. License headers green. NO init(). Source citation present. Staging golden file exists, alphabetically sorted, canonical-marshalled. llms.txt + llms-full.txt updated.
  </done>
</task>

<task type="auto" tdd="true">
  <name>Task 2: Property tests + benchmarks + fuzz harnesses</name>
  <files>props_test.go, sorensen_dice_bench_test.go, sorensen_dice_fuzz_test.go, testdata/fuzz/FuzzSorensenDiceScore/seed-001, testdata/fuzz/FuzzSorensenDiceScoreRunes/seed-001</files>
  <read_first>
    - sorensen_dice.go (Task 1 output)
    - props_test.go (find the QGramJaccard property-test block from plan 05-01 — exact template for the SorensenDice append; both byte and rune sections)
    - qgram_jaccard_bench_test.go (plan 05-01 — exact template for sorensen_dice_bench_test.go)
    - qgram_jaccard_fuzz_test.go (plan 05-01 — exact template for sorensen_dice_fuzz_test.go)
    - testdata/fuzz/FuzzQGramJaccardScore/seed-001 (plan 05-01 — byte-stable format reference)
    - .planning/phases/05-q-gram-algorithms/05-RESEARCH.md §4.1 (alloc budget — same as QGramJaccard: ≤ 4 allocs/op)
  </read_first>
  <behavior>
    - TestProp_SorensenDiceScore_RangeBounds, _Identity, _Symmetric, _NoNaN, _NoInf, _NoNegativeZero — byte path
    - TestProp_SorensenDiceScoreRunes_RangeBounds, _Identity, _Symmetric, _NoNaN, _NoInf, _NoNegativeZero — rune path
    - TestProp_SorensenDiceScore_DeterministicAcrossRuns — 1000 sequential calls on the RV-D1 pair produce byte-identical output
    - BenchmarkSorensenDiceScore_ASCII_Short / _ASCII_Medium / _ASCII_Long + BenchmarkSorensenDiceScoreRunes_Unicode_Short
    - FuzzSorensenDiceScore + FuzzSorensenDiceScoreRunes: panic-free, NaN/Inf-free, score-in-[0,1]
  </behavior>
  <action>
    Step A — Extend props_test.go with the SorensenDice block per the plan 05-01 template. 12 standard property tests (6 byte + 6 rune) + 1 determinism test. testing/quick with default 100 iterations. n coerced to [1, 5] via `n = (n % 5) + 1`.

    Step B — Create sorensen_dice_bench_test.go. Three byte-path benches + one rune-path bench, all alloc-asserted via b.ReportAllocs() + var sink anti-DCE. Byte path uses RV-D1, RV-D2 (medium), and a 200-char realistic name pair (long). Rune path uses "café"/"cafe"/n=2.

    Step C — Create sorensen_dice_fuzz_test.go. Two harnesses; programmatic f.Add(...) seeds cover RV-D1..RV-D4, identity, both-empty, one-empty, invalid UTF-8, long input. Fuzz body coerces n to [1, 8] and asserts no NaN/Inf and score-in-[0,1].

    Step D — Create testdata/fuzz/FuzzSorensenDiceScore/seed-001 and FuzzSorensenDiceScoreRunes/seed-001 in `go test fuzz v1` literal format using RV-D1 ("night"/"nacht"/n=2) as the canonical seed.
  </action>
  <verify>
    <automated>cd /Users/johnny/Development/fuzzymatch && go test -run 'TestProp_SorensenDice' ./... && go test -bench=BenchmarkSorensenDice -benchmem -benchtime=1x ./... && go test -fuzz=FuzzSorensenDiceScore -fuzztime=10s ./... && go test -fuzz=FuzzSorensenDiceScoreRunes -fuzztime=10s ./... && head -1 testdata/fuzz/FuzzSorensenDiceScore/seed-001 | grep -q "^go test fuzz v1$"</automated>
  </verify>
  <done>
    All TestProp_SorensenDice* (byte + rune, 12+ property tests) pass. Bench file produces 4 benches with alloc count within budget. Both fuzz harnesses pass a 10s smoke run. Byte-stable seed-001 files on disk.
  </done>
</task>

<task type="auto" tdd="true">
  <name>Task 3: BDD feature + steps</name>
  <files>tests/bdd/features/sorensen_dice.feature, tests/bdd/steps/algorithms_steps.go</files>
  <read_first>
    - tests/bdd/features/qgram_jaccard.feature (plan 05-01 — exact template)
    - tests/bdd/steps/algorithms_steps.go (find QGramJaccard step methods added in plan 05-01 — exact analog for the SorensenDice append)
    - .planning/phases/05-q-gram-algorithms/05-RESEARCH.md §5.2 (sorensen_dice.feature skeleton)
  </read_first>
  <behavior>
    - godog runs and passes all SorensenDice scenarios
    - Existing approximately-step regex `(\d+\.?\d*)` reused; new SorensenDice-specific n-parameterised step regexes added
  </behavior>
  <action>
    Create tests/bdd/features/sorensen_dice.feature per RESEARCH.md §5.2 skeleton. Header comment: `# Primary sources: Dice, L. R. (1945). Ecology 26(3):297-302. Sørensen, T. (1948). Kgl. Danske Videnskab. Selskab 5(4):1-34.` Scenarios:
    - `Feature: Sørensen-Dice similarity`
    - `Scenario Outline: Canonical reference vectors` covering RV-D1 (night/nacht/n=2/0.2500), RV-D2 (abcdef/bcdefg/n=2/0.8000), RV-D3 (abcdef/abcXef/n=3/0.2500), RV-D4 (hello/hello/n=2/1.0000)
    - `Scenario: identical strings score 1.0`
    - `Scenario: both-empty strings score 1.0`
    - `Scenario: one-empty string scores 0.0`
    - `Scenario: score is symmetric`
    - `Scenario: rune-path Unicode pair` ("café"/"cafe"/n=2 → 0.6667)

    Extend tests/bdd/steps/algorithms_steps.go by appending SorensenDice step methods: iComputeTheSorensenDiceScoreBetweenWithN, iComputeTheSorensenDiceRunesScoreBetweenWithN, iComputeTheSecondSorensenDiceScoreBetweenWithN, bothSorensenDiceScoresShouldBeEqual. Register their regexes inside InitializeScenario. Reuse the existing approximately-step regex.
  </action>
  <verify>
    <automated>cd /Users/johnny/Development/fuzzymatch && make test-bdd 2>&1 | grep -i 'sorensen_dice\|SorensenDice' && (cd tests/bdd && go test -run 'TestFeatures' ./...)</automated>
  </verify>
  <done>
    `make test-bdd` exits 0 with the new SorensenDice scenarios green. Feature file covers all required scenarios. No approximately-step regex drift.
  </done>
</task>

</tasks>

<threat_model>
## Trust Boundaries

| Boundary | Description |
|----------|-------------|
| caller → SorensenDiceScore / SorensenDiceScoreRunes | Untrusted (a, b string, n int) input crosses the API surface; library is pure-function with no I/O |

## STRIDE Threat Register (ASVS Level 1)

| Threat ID | Category | Component | Disposition | Mitigation Plan |
|-----------|----------|-----------|-------------|-----------------|
| T-fuzz-panic | D | SorensenDiceScore / SorensenDiceScoreRunes on malformed UTF-8, extreme inputs, n=0 / n<0 | mitigate | Task 2 ships fuzz harnesses with ≥ 60s budget and seed corpus covering invalid UTF-8, identity, both-empty, one-empty, RV-D1, long inputs. n<1 panic path tested by TestSorensenDice_PanicsOnInvalidN (Task 1) |
| T-complexity-attack | D | extractQGrams on pathological inputs | accept | Same disposition as plan 05-01 — O(la+lb) bounded by input length; PERF-01 budget enforced via bench regression detection |
| T-float-determinism | T | DSC = 2·|∩|/(|QA|+|QB|) reduction | mitigate | Single multiplication + single addition + single division on integer-derived floats; no FMA risk on q-gram counts ≤ 2^53. Cross-platform CI matrix verifies byte-identical golden output via _staging/sorensen_dice.json merged in plan 05-05. Left-to-right parenthesisation per DET-06 |
| T-map-iteration-leak | T | Sorensen-Dice intersection cardinality computation | mitigate | Intersection cardinality is an integer count `sum_k min(qa[k], qb[k])`, not an ordered slice; OUTPUT does not depend on iteration order. DET-03 satisfied. Cross-platform golden gate provides secondary verification |
</threat_model>

<verification>
- `go build ./...` succeeds.
- `go test -run 'TestSorensenDice|TestProp_SorensenDice|TestDispatch_SorensenDiceRegistered|TestGolden_SorensenDice_Staging|ExampleSorensenDice' ./...` exits 0.
- `go test -bench=BenchmarkSorensenDice -benchmem -benchtime=1x ./...` reports alloc count within budget.
- `go test -fuzz=FuzzSorensenDiceScore -fuzztime=60s ./...` and `go test -fuzz=FuzzSorensenDiceScoreRunes -fuzztime=60s ./...` complete without failure (10s smoke for per-task gate).
- `make test-bdd` green; SorensenDice scenarios visible.
- `bash scripts/verify-license-headers.sh` exits 0.
- `bash scripts/verify-no-runtime-deps.sh` exits 0.
- `! grep -q "^func init" sorensen_dice.go`.
- `grep -q "Source: Dice, L. R. (1945)" sorensen_dice.go`.
- `grep -v '^#' sorensen_dice.go | ! grep -E "math\.(Pow|Log|Exp|FMA)"`.
- `grep -q "SorensenDiceScore" llms.txt && grep -q "SorensenDiceScoreRunes" llms.txt`.
- `make coverage-check` confirms sorensen_dice.go ≥ 90% per-file coverage.
- `make check` exits 0.
</verification>

<success_criteria>
- All three tasks complete; all listed verification commands green.
- testdata/golden/_staging/sorensen_dice.json exists and is canonical-marshalled.
- testdata/fuzz/FuzzSorensenDiceScore/seed-001 and FuzzSorensenDiceScoreRunes/seed-001 byte-stable.
- Public surface: exactly TWO new exported functions (SorensenDiceScore, SorensenDiceScoreRunes).
- Dispatch slot wired with default n=3 wrapper.
- Plan 05-05 finalisation can begin once plans 05-03 and 05-04 also ship.
</success_criteria>

<output>
After completion, create `.planning/phases/05-q-gram-algorithms/05-02-sorensen-dice-SUMMARY.md` per the GSD summary template.
</output>
