---
phase: 05-q-gram-algorithms
plan: 04
type: execute
wave: 2
depends_on:
  - 05-01-qgram-foundation-jaccard
files_modified:
  - tversky.go
  - dispatch_tversky.go
  - tversky_test.go
  - tversky_bench_test.go
  - tversky_fuzz_test.go
  - props_test.go
  - example_test.go
  - algoid_test.go
  - algorithms_golden_test.go
  - testdata/golden/_staging/tversky.json
  - testdata/fuzz/FuzzTverskyScore/seed-001
  - testdata/fuzz/FuzzTverskyScoreRunes/seed-001
  - tests/bdd/features/tversky.feature
  - tests/bdd/steps/algorithms_steps.go
  - llms.txt
  - llms-full.txt
autonomous: true
requirements:
  - QGRAM-05
tags: [tversky, tversky-1977, asymmetric-similarity, alpha-beta-parameters, jaccard-fallback-dispatch, byte-and-rune-paths, dispatch-registration, property-tests-asymmetry, fuzz, benchmark, bdd-asymmetry-gate, staging-golden, llms-sync]

must_haves:
  truths:
    # Tversky algorithm (QGRAM-05)
    - "A caller can `import fuzzymatch` and call TverskyScore(\"abcd\", \"abcdef\", 2, 0.8, 0.2) and receive `0.8823529411764706` (RV-T1 — Tversky 1977 §2 prototype/variant; asymmetric direction-sensitive)"
    - "TverskyScore(\"abcdef\", \"abcd\", 2, 0.8, 0.2) returns `0.6521739130434783` (RV-T2 — input swap of RV-T1; load-bearing asymmetry-discriminating pair)"
    - "RV-T1 ≠ RV-T2 (0.8823... ≠ 0.6521...): swapping the inputs produces a DIFFERENT score, proving direction-sensitivity. The asymmetry gate is the load-bearing regression test for parameter-order correctness"
    - "TverskyScore(\"abcd\", \"abce\", 2, 1.0, 1.0) returns `0.5` exactly (RV-T3 — α=β=1.0 reduces to Jaccard; cross-checked against QGramJaccardScore(\"abcd\", \"abce\", 2) which also returns 0.5)"
    - "TverskyScore(\"abcd\", \"abce\", 2, 0.5, 0.5) returns `0.6666666666666666` (RV-T4 — α=β=0.5 reduces to Sørensen-Dice; cross-checked against SorensenDiceScore(\"abcd\", \"abce\", 2) which also returns 0.6666666666666666)"
    - "TverskyScore(\"hello\", \"hello\", 2, 0.8, 0.2) == 1.0 (identity)"
    - "TverskyScore(\"\", \"\", 2, 0.5, 0.5) == 1.0 (both-empty identity convention; even with α+β > 0 the divide-by-zero is avoided by the early both-empty short-circuit)"
    - "TverskyScore(\"\", \"abc\", 2, 0.5, 0.5) == 0.0 (one-empty convention)"
    - "TverskyScore(x, x, n, α, β) == 1.0 for every non-empty x, every n ≥ 1, and every α, β with α+β > 0"
    - "TverskyScore(a, b, n, α, β) == TverskyScore(b, a, n, β, α) (parameter-swap symmetry: swapping inputs is equivalent to swapping α and β — Tversky 1977 §2 property)"
    - "TverskyScore(a, b, n, α, α) == TverskyScore(b, a, n, α, α) when α=β (algorithm becomes symmetric when parameters are equal)"
    - "TverskyScoreRunes operates on the rune q-gram path consuming `extractQGramsRunes` from plan 05-01"
    - "TverskyScore(\"hello\", \"hello\", 0, 0.5, 0.5) panics with message containing `invalid q-gram size`; n=-1 same"
    - "TverskyScore(\"hello\", \"hello\", 2, -0.1, 0.5) panics with message containing `invalid tversky parameter` (negative α); same for β=-0.1; same for α=0 AND β=0 (denominator-zero case)"
    - "TverskyScore(\"hello\", \"hello\", 2, 0.0, 0.5) does NOT panic (α=0 with β>0 is valid — α+β > 0 is the constraint); returns 1.0 by identity short-circuit"
    - "dispatch[AlgoTversky] is non-nil after package load and dispatches to a wrapper that calls TverskyScore(a, b, 3, 1.0, 1.0) — equivalent to Q-Gram Jaccard at the dispatch-table level per CONTEXT.md \"Claude's Discretion\" (Jaccard-fallback chosen because Tversky+Jaccard equivalence is well-known and well-tested via plan 05-01)"
    # Determinism + correctness gates (DET-03, DET-06)
    - "TverskyScore never returns NaN, +Inf, -Inf, or -0 for any input — verified by TestProp_TverskyScore_NoNaN / _NoInf / _NoNegativeZero (byte + rune)"
    - "TverskyScore is deterministic: 1000 sequential calls on the same input produce byte-identical output (TestProp_TverskyScore_DeterministicAcrossRuns)"
    - "No map iteration on the output path in tversky.go — `|A∩B|`, `|A−B|`, `|B−A|` are all integer counts derived via map-length arithmetic and key-membership tests; output ORDER does not depend on iteration order — DET-03 satisfied"
    - "No `math.Pow`, `math.Log`, `math.Exp`, `math.FMA` anywhere in tversky.go (only `+`, `-`, `*`, `/`, comparisons, `float64()` casts) — DET-06 gate"
    - "T(A, B) formula uses explicit left-to-right parenthesisation with documented order of operations: `T = float64(intersection) / (float64(intersection) + (alpha * float64(aMinusB)) + (beta * float64(bMinusA)))` — DET-06"
    # Asymmetry property tests (CONTEXT.md §5)
    - "TestProp_TverskyScore_SymmetricWhenAlphaEqBeta: testing/quick over arbitrary (a, b, n) with α=β → TverskyScore(a, b, n, α, β) == TverskyScore(b, a, n, α, β) (equality, not tolerance)"
    - "TestProp_TverskyScore_AsymmetricWhenAlphaNeqBeta: testing/quick over arbitrary (a, b, n) with α=0.8, β=0.2 → if |A−B| ≠ |B−A| then TverskyScore(a, b, n, 0.8, 0.2) ≠ TverskyScore(b, a, n, 0.8, 0.2) (the implication captures the asymmetry gate; equal multiset-differences trivially produce equal scores)"
    - "TestProp_TverskyScore_ParameterSwapSymmetry: testing/quick over arbitrary (a, b, n, α, β) → TverskyScore(a, b, n, α, β) == TverskyScore(b, a, n, β, α) (Tversky 1977 §2 parameter-swap property; equality, not tolerance)"
    - "TestProp_TverskyScore_JaccardCrossCheck: testing/quick over arbitrary (a, b, n) → TverskyScore(a, b, n, 1.0, 1.0) == QGramJaccardScore(a, b, n) (bit-exact equality — same arithmetic on the same q-gram counts)"
    - "TestProp_TverskyScore_DiceCrossCheck: testing/quick over arbitrary (a, b, n) → TverskyScore(a, b, n, 0.5, 0.5) == SorensenDiceScore(a, b, n) (bit-exact equality)"
    # Public-surface + meta-test discipline
    - "Public surface added by this plan: exactly two new exported symbols (TverskyScore, TverskyScoreRunes) — pre-existing AlgoTversky constant and ErrInvalidTverskyParam sentinel already in place (plan 05-01)"
    - "FuzzTverskyScore + FuzzTverskyScoreRunes panic-free, score-in-[0,1], NaN/Inf-free for any (a, b) including invalid UTF-8; n constrained to [1, 8], α and β constrained to [0.0, 1.0] with α+β > 0 enforced via fuzz body so the documented panic paths are unreachable in fuzz; the panic paths are tested separately in TestTversky_PanicsOnInvalidParams"
    - "testdata/fuzz/FuzzTverskyScore/seed-001 and testdata/fuzz/FuzzTverskyScoreRunes/seed-001 exist in byte-stable `go test fuzz v1` literal format"
    - "tests/bdd/features/tversky.feature exists with: canonical reference-vector Scenario Outline (RV-T1..RV-T4), identity, both-empty, one-empty, AND a dedicated asymmetry-direction-sensitivity scenario (RV-T1 vs RV-T2 input swap)"
    - "tests/bdd/steps/algorithms_steps.go appends Tversky step bindings (with α and β parameters in the step grammar) and their ctx.Step regex registrations"
    - "testdata/golden/_staging/tversky.json exists, produced by TestGolden_Tversky_Staging via assertGoldenStaging; entries sorted alphabetically by Name; includes 8-10 entries covering identity, both-empty, one-empty, RV-T1, RV-T2 (the asymmetry gate as a golden-file row), RV-T3, RV-T4, and one rune-path entry"
    - "algoid_test.go contains a new TestDispatch_TverskyRegistered; registered map updated to flip AlgoTversky slot to true"
    - "ExampleTverskyScore and ExampleTverskyScoreRunes appended to example_test.go; ExampleTverskyScore demonstrates BOTH a symmetric case (α=β) AND an asymmetric case in the Output block per RESEARCH.md OQ-4 recommendation"
    - "llms.txt lists `TverskyScore` and `TverskyScoreRunes` (the two new exported symbols). The AlgoID constant AlgoTversky and the error sentinel ErrInvalidTverskyParam are ALREADY listed (Phase 1 / plan 05-01) — no duplicates"
    - "llms-full.txt has parallel entries with one-line rationales; TverskyScore's rationale notes the α/β asymmetry surface and the dispatch-table Jaccard-fallback compromise"
    - "Coverage on tversky.go ≥ 90%; 100% on the public TverskyScore + TverskyScoreRunes surface"
    - "Apache-2.0 header present on every new .go file"
  artifacts:
    - path: "tversky.go"
      provides: "TverskyScore + TverskyScoreRunes (two new public functions with the asymmetric α/β surface); cites Tversky 1977 §2 as PRIMARY"
      min_lines: 130
      contains: "Source: Tversky, A. (1977)"
    - path: "dispatch_tversky.go"
      provides: "Package-load-time registration of a default-n=3, α=β=1.0 (Jaccard-fallback) TverskyScore wrapper into dispatch[AlgoTversky] per CONTEXT.md \"Claude's Discretion\""
      contains: "dispatch\\[AlgoTversky\\]"
    - path: "tversky_test.go"
      provides: "Unit tests for identity, both-empty, one-empty, RV-T1/T2/T3/T4 reference vectors (byte + rune), the load-bearing RV-T1 vs RV-T2 input-swap asymmetry assertion, direct-call panic-on-invalid-n / panic-on-invalid-alpha-beta tests, and a runtime allocation gate"
      min_lines: 200
      contains: "RV-T1"
    - path: "tversky_bench_test.go"
      provides: "Benchmarks: BenchmarkTverskyScore_{ASCII_Short, ASCII_Medium, ASCII_Long} + BenchmarkTverskyScoreRunes_Unicode_Short — alloc-asserted with b.ReportAllocs() + var sink anti-DCE"
    - path: "tversky_fuzz_test.go"
      provides: "FuzzTverskyScore and FuzzTverskyScoreRunes — panic-free, NaN/Inf-free, score-in-[0,1] with α, β coerced into the valid range"
    - path: "props_test.go"
      provides: "Appended Tversky property-test block: TestProp_TverskyScore_{RangeBounds, Identity, NoNaN, NoInf, NoNegativeZero} for BOTH byte and rune surfaces + TestProp_TverskyScore_SymmetricWhenAlphaEqBeta + TestProp_TverskyScore_AsymmetricWhenAlphaNeqBeta + TestProp_TverskyScore_ParameterSwapSymmetry + TestProp_TverskyScore_JaccardCrossCheck + TestProp_TverskyScore_DiceCrossCheck + TestProp_TverskyScore_DeterministicAcrossRuns"
    - path: "example_test.go"
      provides: "Appended ExampleTverskyScore (with BOTH symmetric AND asymmetric subcases per RESEARCH.md OQ-4) + ExampleTverskyScoreRunes runnable godoc examples"
    - path: "algoid_test.go"
      provides: "Appended TestDispatch_TverskyRegistered; updated registered map"
    - path: "algorithms_golden_test.go"
      provides: "Appended buildTverskyStagingEntries + TestGolden_Tversky_Staging"
    - path: "testdata/golden/_staging/tversky.json"
      provides: "Per-algorithm staging file; sorted by Name; 8-10 entries including BOTH RV-T1 and RV-T2 as separate rows (the input-swap pair is the asymmetry-gate fixture); merged into algorithms.json by plan 05-05"
      contains: "Tversky_abcd_abcdef"
    - path: "testdata/fuzz/FuzzTverskyScore/seed-001"
      provides: "Fuzz seed corpus in `go test fuzz v1` literal format (4 string lines: a, b, n, then α and β as separate float lines if the harness signature includes them)"
    - path: "testdata/fuzz/FuzzTverskyScoreRunes/seed-001"
      provides: "Fuzz seed corpus for the rune-path harness"
    - path: "tests/bdd/features/tversky.feature"
      provides: "Gherkin feature with canonical reference vectors, identity, both-empty, one-empty, dedicated asymmetry direction-sensitivity scenario per RESEARCH.md §5.4"
    - path: "tests/bdd/steps/algorithms_steps.go"
      provides: "Appended Tversky step methods (with α/β parameters in the step grammar) + ctx.Step registrations"
    - path: "llms.txt"
      provides: "Appended 2 function entries (TverskyScore, TverskyScoreRunes)"
    - path: "llms-full.txt"
      provides: "Parallel entries with one-line rationales (including the α/β asymmetry surface notation)"
  key_links:
    - from: "tversky.go (TverskyScore + TverskyScoreRunes)"
      to: "q_gram.go (extractQGrams + extractQGramsRunes — created in plan 05-01)"
      via: "Direct call into the unexported extractor; TverskyScore computes |A∩B|, |A−B|, |B−A| via map-length arithmetic and key-membership tests; T = |A∩B| / (|A∩B| + α·|A−B| + β·|B−A|) — no iteration on output path"
      pattern: "extractQGrams(Runes)?\\("
    - from: "tversky.go (parameter validation)"
      to: "errors.go (ErrInvalidTverskyParam declared in plan 05-01)"
      via: "Direct call panics with message containing 'invalid tversky parameter' for α<0, β<0, α+β==0 per CONTEXT.md §5; the sentinel is reserved for the Phase 8 Scorer layer (WithTverskyAlgorithm errors.Is matching)"
      pattern: "invalid tversky parameter"
    - from: "dispatch_tversky.go"
      to: "algoid.go (AlgoTversky declared at line 127) + Q-Gram Jaccard equivalence (plan 05-01)"
      via: "Closure wrapper `dispatch[AlgoTversky] = func(a, b string) float64 { return TverskyScore(a, b, 3, 1.0, 1.0) }`; per CONTEXT.md \"Claude's Discretion\" the Jaccard-fallback (α=β=1.0) is chosen — Tversky+Jaccard equivalence is RV-T3-verified"
      pattern: "dispatch\\[AlgoTversky\\]"
    - from: "tversky_test.go (RV-T1 and RV-T2)"
      to: "tests/bdd/features/tversky.feature (Asymmetry direction-sensitivity scenario)"
      via: "Both the unit test and the BDD scenario assert RV-T1 ≠ RV-T2 (input swap with fixed α/β produces different scores) — the load-bearing parameter-order regression gate"
      pattern: "0\\.882352941176|0\\.652173913043|Asymmetry"
---

<objective>
Implement Tversky asymmetric similarity (QGRAM-05) — `T(A, B) = |A∩B| / (|A∩B| + α·|A−B| + β·|B−A|)` (Tversky 1977 §2) — atop the shared q-gram infrastructure from plan 05-01. This is the **asymmetric** member of the q-gram tier: swapping (a, b) with fixed (α, β) generally produces a DIFFERENT score. The load-bearing test is the RV-T1 vs RV-T2 input-swap pair (`0.8823 ≠ 0.6521`) — without this pair the implementation could silently swap α and β and still pass RangeBounds + Identity tests. Plan 05-01 already declared the `ErrInvalidTverskyParam` sentinel for the Phase 8 Scorer; direct-call discipline panics per CONTEXT.md §5.

Purpose: ship the parameterised asymmetric similarity that record-linkage and prototype-matching workloads expect, with mathematical cross-validation against Jaccard (α=β=1.0) and Sørensen-Dice (α=β=0.5) — bit-exact equality both in unit tests AND in property tests. The dispatch-table wrapper uses α=β=1.0 (Jaccard-fallback) per CONTEXT.md "Claude's Discretion" — the real Tversky use case lands in Phase 8 via `WithTverskyAlgorithm(weight, alpha, beta)`.

Output: 16 new/modified files (4 new source/test files, plus extensions to 7 existing append-only files, plus 4 new test/fixture files in testdata + tests/bdd). Single new dispatch slot wired (Wave 2 — depends on plan 05-01 ONLY; parallelisable with plans 05-02 and 05-03).
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
@.planning/phases/04-remaining-character-gestalt/04-PATTERNS.md
@.claude/skills/algorithm-correctness-standards/SKILL.md
@.claude/skills/algorithm-licensing-standards/SKILL.md
@.claude/skills/determinism-standards/SKILL.md
@.claude/skills/performance-standards/SKILL.md
@.claude/skills/go-testing-standards/SKILL.md
@algoid.go
@errors.go
@q_gram.go
@qgram_jaccard.go
@sorensen_dice.go
@dispatch_qgram_jaccard.go
</context>

<interfaces>
<!-- Key types/functions executor MUST use without rediscovering. -->

From q_gram.go (created in plan 05-01 — read-only here):
```go
func extractQGrams(s string, n int) map[string]int
func extractQGramsRunes(s string, n int) map[string]int
```

From errors.go (sentinels declared in plan 05-01 — read-only here):
```go
var ErrInvalidQGramSize = errors.New("fuzzymatch: invalid q-gram size")
var ErrInvalidTverskyParam = errors.New("fuzzymatch: invalid tversky parameter")
```

From algoid.go (slot already declared at line 127; do NOT modify):
```go
const AlgoTversky AlgoID = ... // line 127; String() case at line 239-240
```

From qgram_jaccard.go + sorensen_dice.go (plans 05-01 + 05-02 — read-only; used for Jaccard/Dice cross-check property tests):
```go
func QGramJaccardScore(a, b string, n int) float64
func SorensenDiceScore(a, b string, n int) float64
```

Public surface to be created by this plan:
```go
// TverskyScore returns the Tversky asymmetric similarity of the
// q-gram multisets of a and b (Tversky 1977 §2).
// T(A, B) = |A∩B| / (|A∩B| + α·|A−B| + β·|B−A|), in [0.0, 1.0].
//
// T is symmetric when α = β. When α ≠ β, swapping (a, b) generally
// produces a different score: T(a, b, n, α, β) ≠ T(b, a, n, α, β).
// Parameter-swap symmetry holds: T(a, b, n, α, β) = T(b, a, n, β, α).
//
// Special cases:
//   - α = β = 1.0 reduces to Q-Gram Jaccard
//   - α = β = 0.5 reduces to Sørensen-Dice
//
// Panics on n < 1 with "invalid q-gram size".
// Panics on α < 0, β < 0, or α + β == 0 with "invalid tversky parameter".
func TverskyScore(a, b string, n int, alpha, beta float64) float64

func TverskyScoreRunes(a, b string, n int, alpha, beta float64) float64
```

Dispatch wiring (Jaccard-fallback per CONTEXT.md "Claude's Discretion"):
```go
var _ = func() bool {
    dispatch[AlgoTversky] = func(a, b string) float64 {
        return TverskyScore(a, b, 3, 1.0, 1.0)
    }
    return true
}()
```
</interfaces>

<tasks>

<task type="auto" tdd="true">
  <name>Task 1: Implement tversky.go + dispatch + unit tests with asymmetry gate + staging golden</name>
  <files>tversky.go, dispatch_tversky.go, tversky_test.go, testdata/golden/_staging/tversky.json, algorithms_golden_test.go, algoid_test.go, example_test.go, llms.txt, llms-full.txt</files>
  <read_first>
    - tversky.go (current state — confirm it does NOT exist; creating new)
    - q_gram.go, qgram_jaccard.go, sorensen_dice.go (plans 05-01 + 05-02 — read to understand the surface; Tversky cross-checks against Jaccard + Dice in unit tests)
    - errors.go (find ErrInvalidTverskyParam declared in plan 05-01; the godoc lists the panic paths)
    - dispatch_qgram_jaccard.go (plan 05-01 — exact template for dispatch_tversky.go)
    - .planning/phases/05-q-gram-algorithms/05-CONTEXT.md §5 (LOCKED patterns — Tversky parameter validation; dispatch wrapper Jaccard-fallback)
    - .planning/phases/05-q-gram-algorithms/05-RESEARCH.md §1.4 (Tversky formula), §2.4 (RV-T1..RV-T6 with the load-bearing RV-T1 vs RV-T2 asymmetry pair)
    - algoid.go line 127 (AlgoTversky slot)
    - algorithms_golden_test.go (find buildQGramJaccardStagingEntries from plan 05-01 — exact template)
  </read_first>
  <behavior>
    - TverskyScore("abcd", "abcdef", 2, 0.8, 0.2) returns 0.8823529411764706 (RV-T1) within 1e-15 — derivation: |QA|=3 {ab,bc,cd}, |QB|=5 {ab,bc,cd,de,ef}, |A∩B|=3, |A−B|=0, |B−A|=2; T = 3/(3 + 0.8·0 + 0.2·2) = 3/3.4
    - TverskyScore("abcdef", "abcd", 2, 0.8, 0.2) returns 0.6521739130434783 (RV-T2 — INPUT SWAP) within 1e-15 — derivation: |QA|=5, |QB|=3, |A∩B|=3, |A−B|=2, |B−A|=0; T = 3/(3 + 0.8·2 + 0.2·0) = 3/4.6
    - RV-T1 ≠ RV-T2 (LOAD-BEARING asymmetry gate — explicit assertion `math.Abs(rvT1 - rvT2) > 0.1` in TestTversky_AsymmetryDirectionSensitive)
    - TverskyScore("abcd", "abce", 2, 1.0, 1.0) returns 0.5 exactly (RV-T3 — Jaccard cross-check); ALSO assert TverskyScore("abcd", "abce", 2, 1.0, 1.0) == QGramJaccardScore("abcd", "abce", 2) bit-for-bit
    - TverskyScore("abcd", "abce", 2, 0.5, 0.5) returns 0.6666666666666666 (RV-T4 — Dice cross-check); ALSO assert TverskyScore("abcd", "abce", 2, 0.5, 0.5) == SorensenDiceScore("abcd", "abce", 2) bit-for-bit
    - TverskyScore("hello", "hello", 2, 0.8, 0.2) == 1.0 (identity)
    - TverskyScore("", "", 2, 0.5, 0.5) == 1.0 (both-empty)
    - TverskyScore("", "abc", 2, 0.5, 0.5) == 0.0 (one-empty)
    - TverskyScoreRunes("café", "cafe", 2, 0.5, 0.5) — derivation: rune-bigrams [ca,af,fé] vs [ca,af,fe]; |∩|=2, |A−B|=1, |B−A|=1; T = 2/(2 + 0.5·1 + 0.5·1) = 2/3 = 0.6666666666666666
    - TverskyScore("hello", "hello", 0, 0.5, 0.5) panics with "invalid q-gram size"
    - TverskyScore("hello", "hello", 2, -0.1, 0.5) panics with "invalid tversky parameter" (α < 0)
    - TverskyScore("hello", "hello", 2, 0.5, -0.1) panics with "invalid tversky parameter" (β < 0)
    - TverskyScore("hello", "hello", 2, 0.0, 0.0) panics with "invalid tversky parameter" (α + β == 0)
    - TverskyScore("hello", "hello", 2, 0.0, 0.5) does NOT panic (α=0 is valid when β > 0); returns 1.0 by identity short-circuit
    - dispatch[AlgoTversky]("abcd", "abcdef") returns TverskyScore("abcd", "abcdef", 3, 1.0, 1.0) which equals QGramJaccardScore("abcd", "abcdef", 3) (Jaccard-fallback)
    - TestGolden_Tversky_Staging produces testdata/golden/_staging/tversky.json with 8-10 alphabetically-sorted entries — BOTH RV-T1 AND RV-T2 are separate rows (the asymmetry gate as a golden fixture)
  </behavior>
  <action>
    Step A — Create tversky.go. File order:
    (a) Apache-2.0 header.
    (b) File-level doc: cite Tversky, A. (1977). "Features of similarity." Psychological Review 84(4):327-352 — specifically §2 (the asymmetric similarity model and the worked examples on pp.331-332) — as PRIMARY. The godoc includes the T(A, B) = |A∩B| / (|A∩B| + α·|A−B| + β·|B−A|) formula, the asymmetric-when-α≠β property, the parameter-swap symmetry (T(a, b, α, β) = T(b, a, β, α)), and the special cases (α=β=1 → Jaccard; α=β=0.5 → Dice). Source-origin statement block (Primary / Cross-validation: hand-derived RV-T1..RV-T4 in tversky_test.go + bit-exact cross-check property tests against QGramJaccardScore and SorensenDiceScore / Tie-break: none / GPL-LGPL: none / Code copied: none).
    (c) `package fuzzymatch`.
    (d) `func TverskyScore(a, b string, n int, alpha, beta float64) float64` — godoc reproducing the formula, the asymmetry property, the [0, 1] range, identity / both-empty / one-empty conventions, and BOTH panic contracts (n<1 → "invalid q-gram size"; α<0 || β<0 || α+β==0 → "invalid tversky parameter"). Body:
        - Identity short-circuit `if a == b { return 1.0 }` (covers both-empty + identical)
        - `if n < 1 { panic("fuzzymatch: invalid q-gram size") }`
        - `if alpha < 0 || beta < 0 || (alpha == 0 && beta == 0) { panic("fuzzymatch: invalid tversky parameter") }` — note the precise predicate: `α + β == 0` is implemented as `α == 0 && β == 0` to dodge any float-comparison anxiety on α+β being slightly less than 0
        - One-empty short-circuit → 0.0
        - `qa := extractQGrams(a, n); qb := extractQGrams(b, n)`
        - If `len(qa) == 0 && len(qb) == 0 { return 1.0 }` (post-extraction both-empty)
        - Compute intersection cardinality: iterate the SMALLER map, accumulating `intersection += min(qa[k], qb[k])` when key present in both — integer accumulator, no float reduction
        - Compute aOnly = sum(qa) - intersection; bOnly = sum(qb) - intersection (multiset semantics — `sum(qa)` is the total weight summing over values; equivalent to `len(a)-n+1` for non-empty input but computing via sum-of-values keeps the byte/rune path uniform)
        - Compute the score: `denom := float64(intersection) + (alpha * float64(aOnly)) + (beta * float64(bOnly))`. Document the parenthesisation: explicit `(α * |A−B|)` and `(β * |B−A|)` per DET-06.
        - Return `float64(intersection) / denom`. Note: when intersection=0 AND aOnly=0 AND bOnly=0 → denom=0 → division by zero. But this case is unreachable because we already short-circuited both-empty → 1.0 and one-empty → 0.0 above; the only post-extraction empty case is len(qa)==0 && len(qb)==0 which is also handled by the early return. Document this invariant in a code comment.
    (e) `func TverskyScoreRunes(a, b string, n int, alpha, beta float64) float64` — analogous, calling extractQGramsRunes.
    (f) Use only `+`, `-`, `*`, `/`, comparisons, `float64()` casts — NO math.Pow/Log/Exp/FMA.

    Step B — Create dispatch_tversky.go per dispatch_qgram_jaccard.go template. Closure binds n=3, α=β=1.0 (Jaccard-fallback per CONTEXT.md "Claude's Discretion"). NO init(). Apache-2.0 header. Add a clear godoc note: "The dispatch wrapper binds α=β=1.0 (Jaccard-equivalent) because the `dispatch[AlgoID]` signature has no slot for α/β parameters. The real Tversky use case is via the Phase 8 Scorer option `WithTverskyAlgorithm(weight, alpha, beta)` which forwards user-supplied α/β to TverskyScore directly."

    Step C — Create tversky_test.go covering:
    - TestTversky_BothEmpty (with default α=β=0.5)
    - TestTversky_OneEmpty (both directions)
    - TestTversky_Identical (with several α/β combinations)
    - TestTversky_ReferenceVectors — table-driven covering RV-T1, RV-T2, RV-T3, RV-T4. Each row's test comment reproduces the formula derivation from RESEARCH.md §2.4 with the RV-T{N} identifier in the test name and the full intermediate values (|∩|, |A−B|, |B−A|, computed denom, final score). Tolerance: bit-exact for RV-T3 (0.5) and the Jaccard cross-check; 1e-15 for RV-T1 (0.8823...) and RV-T2 (0.6521...) and RV-T4 (0.6666...).
    - **TestTversky_AsymmetryDirectionSensitive** — THE LOAD-BEARING ASYMMETRY GATE. Compute rvT1 := TverskyScore("abcd", "abcdef", 2, 0.8, 0.2) and rvT2 := TverskyScore("abcdef", "abcd", 2, 0.8, 0.2); assert `math.Abs(rvT1 - rvT2) > 0.1` (the actual difference is 0.2302). On failure t.Errorf with a message about parameter-order regression. This test is what prevents a silent α/β swap from passing the suite.
    - TestTversky_JaccardCrossCheck — TverskyScore(a, b, n, 1.0, 1.0) == QGramJaccardScore(a, b, n) for several hand-picked pairs; bit-exact equality via `math.Float64bits` comparison
    - TestTversky_DiceCrossCheck — TverskyScore(a, b, n, 0.5, 0.5) == SorensenDiceScore(a, b, n) for several hand-picked pairs; bit-exact equality
    - TestTversky_ParameterSwapSymmetry — TverskyScore(a, b, n, α, β) == TverskyScore(b, a, n, β, α) for several hand-picked pairs; bit-exact equality. This pin asserts the algebraic identity that asymmetry is the result of α≠β, not a one-sided coding error.
    - TestTversky_SymmetricWhenAlphaEqBeta — TverskyScore(a, b, n, α, α) == TverskyScore(b, a, n, α, α); bit-exact equality
    - TestTverskyRunes_CafeReference — derivation in test comment; result 0.6666666666666666 within 1e-15
    - TestTversky_PanicsOnInvalidN — table-driven defer-recover over n=0, n=-1, n=-100. Assert panic message contains "invalid q-gram size".
    - TestTversky_PanicsOnInvalidParams — table-driven defer-recover over (α=-0.1, β=0.5), (α=0.5, β=-0.1), (α=0, β=0). Assert panic message contains "invalid tversky parameter".
    - TestTversky_AllocBound via testing.AllocsPerRun(100, ...) — document the alloc count per RESEARCH.md §4.1 (≤ 4 allocs/op)
    - Stdlib testing only.

    Step D — Append buildTverskyStagingEntries + TestGolden_Tversky_Staging to algorithms_golden_test.go. Entries (Name sorted alphabetically): `Tversky_abcd_abce_jaccard_eq` (RV-T3; α=β=1.0; score=0.5), `Tversky_abcd_abce_dice_eq` (RV-T4; α=β=0.5; score=0.6666666666666666), `Tversky_abcd_abcdef_asym` (RV-T1; α=0.8, β=0.2; score=0.8823529411764706), `Tversky_abcdef_abcd_asym_swap` (RV-T2; α=0.8, β=0.2; score=0.6521739130434783), `Tversky_both_empty` (α=β=0.5; score=1.0), `Tversky_cafe_runes` (rune path; α=β=0.5; score=0.6666666666666666), `Tversky_identical` ("hello"/"hello"/n=2/α=0.8/β=0.2; score=1.0), `Tversky_one_empty` (α=β=0.5; score=0.0). The two asymmetry-pair rows (RV-T1 + RV-T2) are intentionally adjacent in the alphabetical sort for reviewer clarity. Run `go test -run TestGolden_Tversky_Staging -update ./...` to materialise.

    Step E — Append TestDispatch_TverskyRegistered to algoid_test.go; flip AlgoTversky slot in the registered map.

    Step F — Append ExampleTverskyScore + ExampleTverskyScoreRunes to example_test.go per RESEARCH.md OQ-4 recommendation. ExampleTverskyScore demonstrates BOTH a symmetric case AND an asymmetric case in the Output block:
    ```go
    func ExampleTverskyScore() {
        // Symmetric case: α=β=1.0 reduces to Jaccard
        fmt.Printf("%.4f\n", fuzzymatch.TverskyScore("abcd", "abce", 2, 1.0, 1.0))
        // Asymmetric case: α≠β; swapping inputs produces different scores
        fmt.Printf("%.4f\n", fuzzymatch.TverskyScore("abcd", "abcdef", 2, 0.8, 0.2))
        fmt.Printf("%.4f\n", fuzzymatch.TverskyScore("abcdef", "abcd", 2, 0.8, 0.2))
        // Output:
        // 0.5000
        // 0.8824
        // 0.6522
    }
    ```
    Capture exact stdout once via `go test -run ExampleTverskyScore ./...` and paste byte-for-byte.

    Step G — Update llms.txt + llms-full.txt with the two new function entries. The llms-full.txt entry for TverskyScore notes the α/β asymmetry surface and the dispatch-table Jaccard-fallback compromise.
  </action>
  <verify>
    <automated>cd /Users/johnny/Development/fuzzymatch && go build ./... && go test -run 'TestTversky|TestDispatch_TverskyRegistered|TestDispatch_UnregisteredSlotsAreNil|TestGolden_Tversky_Staging|ExampleTverskyScore' ./... && bash scripts/verify-license-headers.sh && ! grep -q "^func init" tversky.go && grep -q "Source: Tversky, A. (1977)" tversky.go && grep -q "dispatch\[AlgoTversky\]" dispatch_tversky.go && grep -q "TverskyScore" llms.txt && grep -v '^#' tversky.go | ! grep -E "math\.(Pow|Log|Exp|FMA)" && grep -q "invalid tversky parameter" tversky.go && grep -q "RV-T1" tversky_test.go && grep -q "RV-T2" tversky_test.go && grep -q "AsymmetryDirectionSensitive" tversky_test.go</automated>
  </verify>
  <done>
    All Tversky* unit tests, TestDispatch_TverskyRegistered, TestGolden_Tversky_Staging, and ExampleTverskyScore pass. License headers green. NO init(). Source citation present. Two panic paths exercised (n<1 + invalid α/β). Asymmetry gate (RV-T1 vs RV-T2 input swap) explicitly tested. Jaccard + Dice cross-checks bit-exact. Staging golden file alphabetically sorted with BOTH RV-T1 and RV-T2 rows. llms.txt + llms-full.txt updated.
  </done>
</task>

<task type="auto" tdd="true">
  <name>Task 2: Property tests + benchmarks + fuzz harnesses</name>
  <files>props_test.go, tversky_bench_test.go, tversky_fuzz_test.go, testdata/fuzz/FuzzTverskyScore/seed-001, testdata/fuzz/FuzzTverskyScoreRunes/seed-001</files>
  <read_first>
    - tversky.go (Task 1 output)
    - qgram_jaccard.go, sorensen_dice.go (plans 05-01 + 05-02 — needed for the JaccardCrossCheck + DiceCrossCheck property tests)
    - props_test.go (find the QGramJaccard + SorensenDice property-test blocks — exact template; the Tversky block ADDS five Tversky-specific property tests on top of the standard six)
    - qgram_jaccard_bench_test.go (plan 05-01 — exact template; Tversky benches must pass α and β parameters)
    - qgram_jaccard_fuzz_test.go (plan 05-01 — exact template; Tversky fuzz harness body must coerce α and β into the valid range)
    - .planning/phases/05-q-gram-algorithms/05-CONTEXT.md §5 ("Symmetric for Jaccard / Dice / Cosine; Tversky is symmetric only when α = β — asymmetric property test for α ≠ β")
    - .planning/phases/05-q-gram-algorithms/05-RESEARCH.md §2.4 (RV-T1..RV-T6; the asymmetry mandate; the Jaccard/Dice cross-check derivations)
  </read_first>
  <behavior>
    - TestProp_TverskyScore_RangeBounds, _Identity, _NoNaN, _NoInf, _NoNegativeZero (byte + rune) — STANDARD invariants
    - TestProp_TverskyScore_SymmetricWhenAlphaEqBeta (byte + rune) — quick over (a, b, n, α) with β=α; assert exact equality
    - TestProp_TverskyScore_AsymmetricWhenAlphaNeqBeta (byte + rune) — quick over (a, b, n) with fixed α=0.8, β=0.2; assert that if `|A−B| ≠ |B−A|` (computed via map-length comparison on the helper output) then the two scores differ. The implication structure: equal-difference inputs produce trivially-equal scores; only asymmetric-difference inputs exercise the asymmetry, so the property is `multisetDifferAcrossSides ⇒ scoresDiffer`.
    - TestProp_TverskyScore_ParameterSwapSymmetry (byte + rune) — quick over (a, b, n, α, β); assert TverskyScore(a, b, n, α, β) == TverskyScore(b, a, n, β, α) bit-exact
    - TestProp_TverskyScore_JaccardCrossCheck (byte + rune) — quick over (a, b, n); assert TverskyScore(a, b, n, 1.0, 1.0) == QGramJaccardScore(a, b, n) bit-exact
    - TestProp_TverskyScore_DiceCrossCheck (byte + rune) — quick over (a, b, n); assert TverskyScore(a, b, n, 0.5, 0.5) == SorensenDiceScore(a, b, n) bit-exact
    - TestProp_TverskyScore_DeterministicAcrossRuns — 1000 calls on RV-T1 input; assert byte-identical output
    - 4 benchmarks (3 byte + 1 rune) with α=0.8, β=0.2; alloc-asserted
    - 2 fuzz harnesses: panic-free, NaN/Inf-free, score-in-[0,1]; α and β coerced into [0.0, 1.0] with α+β > 0
  </behavior>
  <action>
    Step A — Extend props_test.go with the Tversky block. Standard property tests (5 × 2 surfaces = 10) PLUS five Tversky-specific property tests (SymmetricWhenAlphaEqBeta × 2, AsymmetricWhenAlphaNeqBeta × 2, ParameterSwapSymmetry × 2, JaccardCrossCheck × 2, DiceCrossCheck × 2) = 20 property tests + 1 determinism test = 21 property tests for Tversky. Default 100 quick.Check iterations.

    Step B — Create tversky_bench_test.go. Three byte-path benches + one rune-path bench, all with α=0.8, β=0.2 (the asymmetric configuration — exercises the full code path). Alloc-asserted; expected ≤ 4 allocs/op (q-gram extractor maps + cap-hint backing arrays — no sort.Strings here, so one less alloc than Cosine).

    Step C — Create tversky_fuzz_test.go. Two harnesses; fuzz body coerces α and β into [0.0, 1.0] via `α = math.Abs(α) / (math.Abs(α) + 1.0)` (squashes to [0, 1)) then enforces α+β > 0 via `if α+β <= 0 { α, β = 1.0, 0.0 }`. Assert no NaN/Inf, score in [0, 1].

    Step D — Create testdata/fuzz/FuzzTverskyScore/seed-001 and testdata/fuzz/FuzzTverskyScoreRunes/seed-001 in `go test fuzz v1` literal format using RV-T1 ("abcd"/"abcdef"/n=2/α=0.8/β=0.2) as the canonical seed.
  </action>
  <verify>
    <automated>cd /Users/johnny/Development/fuzzymatch && go test -run 'TestProp_Tversky' ./... && go test -bench=BenchmarkTversky -benchmem -benchtime=1x ./... && go test -fuzz=FuzzTverskyScore -fuzztime=10s ./... && go test -fuzz=FuzzTverskyScoreRunes -fuzztime=10s ./... && head -1 testdata/fuzz/FuzzTverskyScore/seed-001 | grep -q "^go test fuzz v1$"</automated>
  </verify>
  <done>
    All TestProp_Tversky* (21 property tests across byte + rune surfaces, with asymmetry-conditional + Jaccard cross-check + Dice cross-check + parameter-swap-symmetry) pass. Bench file produces 4 benches with alloc count within budget. Both fuzz harnesses pass a 10s smoke run. Byte-stable seed-001 files on disk.
  </done>
</task>

<task type="auto" tdd="true">
  <name>Task 3: BDD feature + steps with asymmetry scenario</name>
  <files>tests/bdd/features/tversky.feature, tests/bdd/steps/algorithms_steps.go</files>
  <read_first>
    - tests/bdd/features/qgram_jaccard.feature (plan 05-01 — exact template for the canonical-vectors block)
    - tests/bdd/steps/algorithms_steps.go (find QGramJaccard step methods from plan 05-01 — exact analog; Tversky steps add α and β parameters to the grammar)
    - .planning/phases/05-q-gram-algorithms/05-RESEARCH.md §5.4 (tversky.feature skeleton — INCLUDES the asymmetry scenario per CONTEXT.md §5)
  </read_first>
  <behavior>
    - godog runs and passes all Tversky scenarios
    - Canonical reference vectors (RV-T1..RV-T4) covered in the Examples table
    - Dedicated asymmetry direction-sensitivity scenario asserting RV-T1 ≠ RV-T2 via the "two scores should differ by more than 0.1" step (or equivalent pin-both-values pattern)
  </behavior>
  <action>
    Create tests/bdd/features/tversky.feature per RESEARCH.md §5.4 skeleton. Header comment: `# Primary source: Tversky, A. (1977). "Features of similarity." Psychological Review 84(4):327-352, §2.` Scenarios:
    - `Feature: Tversky asymmetric similarity`
    - `Scenario Outline: Canonical reference vectors` covering RV-T1 (abcd/abcdef/n=2/α=0.8/β=0.2/0.8823529411764706), RV-T3 (abcd/abce/n=2/α=1.0/β=1.0/0.5000), RV-T4 (abcd/abce/n=2/α=0.5/β=0.5/0.6666666666666666), identity (hello/hello/n=2/α=0.8/β=0.2/1.0000)
    - `Scenario: both-empty strings score 1.0`
    - `Scenario: one-empty string scores 0.0`
    - `Scenario: identical strings score 1.0`
    - **`Scenario: Asymmetry direction-sensitivity gate`** — compute Tversky("abcd", "abcdef", 2, 0.8, 0.2) AND Tversky("abcdef", "abcd", 2, 0.8, 0.2); assert both scores differ by more than 0.1 (or pin both values explicitly: 0.8824 and 0.6522 with tolerance 0.0001). The choice between the "differ by" step and "pin both" step is the planner's discretion; RESEARCH.md §5.4 recommends the "differ by" form when the BDD grammar supports it, and "pin both" otherwise.
    - `Scenario: parameter-swap symmetry` — compute Tversky(a, b, n, α, β) AND Tversky(b, a, n, β, α); assert equality (these two should match within 1e-9)
    - `Scenario: rune-path Unicode pair` ("café"/"cafe"/n=2/α=0.5/β=0.5/0.6667)

    Extend tests/bdd/steps/algorithms_steps.go by appending Tversky step methods. Required new methods on AlgorithmContext:
    - iComputeTheTverskyScoreBetweenWithNAlphaBeta(a, b string, n int, alpha, beta float64) error
    - iComputeTheTverskyRunesScoreBetweenWithNAlphaBeta(a, b string, n int, alpha, beta float64) error
    - iComputeTheSecondTverskyScoreBetweenWithNAlphaBeta(a, b string, n int, alpha, beta float64) error
    - theTwoTverskyScoresShouldDifferByMoreThan(threshold float64) error (the asymmetry-gate step)
    - bothTverskyScoresShouldBeEqual() error (the parameter-swap-symmetry step)
    Register their regexes inside InitializeScenario. The regex pattern for the α/β floats is `([-+]?\d*\.?\d+)` (same shape as the existing approximately-step regex).
  </action>
  <verify>
    <automated>cd /Users/johnny/Development/fuzzymatch && make test-bdd 2>&1 | grep -i 'tversky\|Tversky' && (cd tests/bdd && go test -run 'TestFeatures' ./...)</automated>
  </verify>
  <done>
    `make test-bdd` exits 0 with the new Tversky scenarios green. Feature file covers identity, both-empty, one-empty, RV-T1..RV-T4 reference vectors, asymmetry direction-sensitivity gate, parameter-swap symmetry, rune-path Unicode. The dedicated asymmetry scenario is the load-bearing BDD gate.
  </done>
</task>

</tasks>

<threat_model>
## Trust Boundaries

| Boundary | Description |
|----------|-------------|
| caller → TverskyScore / TverskyScoreRunes | Untrusted (a, b string, n int, α, β float64) input crosses the API surface; library is pure-function with no I/O |

## STRIDE Threat Register (ASVS Level 1)

| Threat ID | Category | Component | Disposition | Mitigation Plan |
|-----------|----------|-----------|-------------|-----------------|
| T-fuzz-panic | D | TverskyScore / TverskyScoreRunes on malformed inputs, n=0 / n<0, α<0 / β<0 / α+β==0, NaN α or β | mitigate | Task 2 ships fuzz harnesses with ≥ 60s budget. Fuzz body coerces α and β into [0.0, 1.0] AND enforces α+β > 0 so the documented panic paths are unreachable in fuzz; the panic paths are tested separately in TestTversky_PanicsOnInvalidN and TestTversky_PanicsOnInvalidParams (Task 1). NaN α/β: not explicitly tested — document as future hardening if a NaN-input fuzz finding emerges |
| T-complexity-attack | D | extractQGrams on pathological inputs | accept | Same disposition as plans 05-01 / 05-02; PERF-01 budget enforced via bench regression detection |
| T-float-determinism-tversky | T | TverskyScore final division `intersection / (intersection + α·aOnly + β·bOnly)` | mitigate | Single multiplication + addition + division on integer-derived floats (counts ≤ 2^53). α/β are user-supplied float64 — IEEE-754 multiplication is correctly rounded. No FMA risk per RESEARCH.md §3 since the parenthesisation gives the compiler limited fusion opportunity and the q-gram count magnitudes are small. Cross-platform CI matrix verifies byte-identical golden output via _staging/tversky.json merged in plan 05-05. TestProp_TverskyScore_DeterministicAcrossRuns (Task 2) pins per-process determinism |
| T-map-iteration-leak | T | Tversky |A∩B|, |A−B|, |B−A| computation | mitigate | All three cardinalities are integer counts; OUTPUT does not depend on iteration order. DET-03 satisfied. Cross-platform golden gate provides secondary verification |
| T-parameter-order-bug | T (Tampering — silent α/β swap) | TverskyScore implementation | mitigate | LOAD-BEARING. TestTversky_AsymmetryDirectionSensitive (Task 1) asserts RV-T1 ≠ RV-T2 with `math.Abs(rvT1 - rvT2) > 0.1`. The asymmetry scenario in tests/bdd/features/tversky.feature (Task 3) re-exercises the same pair at the BDD layer. TestProp_TverskyScore_ParameterSwapSymmetry (Task 2) asserts the algebraic identity T(a,b,α,β) = T(b,a,β,α) — if the implementation silently swapped α and β, this property would fail. The combination of unit-level + property-level + BDD-level coverage forms a three-layer defence against parameter-order regressions |
</threat_model>

<verification>
- `go build ./...` succeeds.
- `go test -run 'TestTversky|TestProp_Tversky|TestDispatch_TverskyRegistered|TestGolden_Tversky_Staging|ExampleTversky' ./...` exits 0.
- `go test -bench=BenchmarkTversky -benchmem -benchtime=1x ./...` reports alloc count within budget (≤ 4 allocs/op).
- `go test -fuzz=FuzzTverskyScore -fuzztime=60s ./...` and `go test -fuzz=FuzzTverskyScoreRunes -fuzztime=60s ./...` complete without failure (10s smoke for per-task gate).
- `make test-bdd` green; Tversky scenarios visible including the asymmetry direction-sensitivity scenario.
- `bash scripts/verify-license-headers.sh` exits 0.
- `bash scripts/verify-no-runtime-deps.sh` exits 0.
- `! grep -q "^func init" tversky.go`.
- `grep -q "Source: Tversky, A. (1977)" tversky.go`.
- `grep -v '^#' tversky.go | ! grep -E "math\.(Pow|Log|Exp|FMA)"` (DET-06 gate).
- `grep -q "invalid tversky parameter" tversky.go` (panic message present).
- `grep -q "TverskyScore" llms.txt && grep -q "TverskyScoreRunes" llms.txt`.
- `grep -q "RV-T1" tversky_test.go && grep -q "RV-T2" tversky_test.go && grep -q "AsymmetryDirectionSensitive" tversky_test.go` (load-bearing asymmetry gate explicitly tested).
- `make coverage-check` confirms tversky.go ≥ 90% per-file coverage.
- `make check` exits 0.
</verification>

<success_criteria>
- All three tasks complete; all listed verification commands green.
- testdata/golden/_staging/tversky.json exists, canonical-marshalled, 8-10 alphabetically-sorted entries including BOTH RV-T1 and RV-T2 as separate rows.
- testdata/fuzz/FuzzTverskyScore/seed-001 and FuzzTverskyScoreRunes/seed-001 byte-stable.
- Public surface: exactly TWO new exported functions (TverskyScore, TverskyScoreRunes); pre-existing AlgoTversky constant and ErrInvalidTverskyParam sentinel reused.
- Dispatch slot wired with α=β=1.0 (Jaccard-fallback) wrapper.
- The asymmetry-direction-sensitivity gate is exercised at THREE layers: unit (TestTversky_AsymmetryDirectionSensitive), property (TestProp_TverskyScore_AsymmetricWhenAlphaNeqBeta + ParameterSwapSymmetry), BDD (tversky.feature Asymmetry scenario). A silent α/β swap fails all three.
- Jaccard cross-check (RV-T3, α=β=1.0) and Dice cross-check (RV-T4, α=β=0.5) are bit-exact equalities in both unit and property tests — confirming Tversky degenerates correctly to the special cases.
- Plan 05-05 finalisation can begin once plans 05-02 and 05-03 also ship.
</success_criteria>

<output>
After completion, create `.planning/phases/05-q-gram-algorithms/05-04-tversky-SUMMARY.md` per the GSD summary template. Note any findings about asymmetry-gate test design or NaN-α/β fuzz hardening recommendations.
</output>
