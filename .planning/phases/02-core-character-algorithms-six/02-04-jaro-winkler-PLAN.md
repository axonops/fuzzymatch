---
phase: 02-core-character-algorithms-six
plan: 04
type: execute
wave: 3
depends_on: [02-01-levenshtein, 02-03-jaro]
files_modified:
  - jarowinkler.go
  - dispatch_jarowinkler.go
  - jarowinkler_test.go
  - jarowinkler_bench_test.go
  - jarowinkler_fuzz_test.go
  - props_test.go
  - example_test.go
  - algoid_test.go
  - testdata/golden/_staging/jarowinkler.json
  - testdata/fuzz/FuzzJaroWinklerScore/seed-001
  - tests/bdd/features/jarowinkler.feature
  - tests/bdd/steps/algorithms_steps.go
autonomous: true
requirements:
  - CHAR-06
  - PERF-01
  - PERF-02
  - TEST-01
  - TEST-02
  - TEST-04
  - TEST-05
  - DET-04
  - DX-02
tags: [jaro-winkler, prefix-boost, winkler-1990, not-a-metric, builds-on-jaro]

must_haves:
  truths:
    - "JaroWinklerScore(\"MARTHA\", \"MARHTA\") returns 0.9611111111 (within 1e-6) — Winkler 1990 canonical pair"
    - "JaroWinklerScore(\"DIXON\", \"DICKSONX\") returns 0.8133333333 (within 1e-6) — Winkler 1990"
    - "JaroWinklerScore(\"DWAYNE\", \"DUANE\") returns ≈ 0.8400 (within 1e-3) — Winkler 1990"
    - "JaroWinklerScore(\"ABC\", \"ABC\") returns 1.0 exactly"
    - "JaroWinklerScore(\"\", \"\") returns 1.0 exactly; JaroWinklerScore(\"\", \"ABC\") returns 0.0 exactly"
    - "Jaro-Winkler constants traced to Winkler 1990 p. 357 with paper citation in godoc on each constant: winklerPrefixScale = 0.1, winklerMaxPrefix = 4, winklerBoostThreshold = 0.7"
    - "Boost is applied ONLY when underlying Jaro >= winklerBoostThreshold (0.7); otherwise JW = J"
    - "Prefix length L is capped at winklerMaxPrefix (4) — verified by reference vector and a unit test using a >4-char common prefix"
    - "JaroWinklerScoreRunes works on multi-byte UTF-8 input"
    - "dispatch[AlgoJaroWinkler] is non-nil and equals JaroWinklerScore after package init"
    - "Apache-2.0 header on every new .go file (verified by scripts/verify-license-headers.sh)"
    - "jarowinkler.go contains `// Source: Winkler, W. E. (1990).` block at top of file-level godoc"
    - "jarowinkler.go file-level godoc EXPLICITLY states: 'Jaro-Winkler is NOT a metric (inherits Jaro's non-metric property).'"
    - "No math.Pow / math.Log / math.Exp / math.Sqrt / math.FMA in jarowinkler.go (verified by grep)"
    - "JW formula uses left-to-right float reduction: J + float64(L) * winklerPrefixScale * (1.0 - J) — explicit parenthesisation"
    - "No init() function; dispatch_jarowinkler.go uses var _ = func() bool { ... }() registration idiom"
    - "Property tests pass for RangeBounds, Identity, Symmetric, NoNaN, NoInf, NoNegativeZero in props_test.go (NO triangle inequality — JW inherits Jaro's non-metric property)"
    - "Benchmark BenchmarkJaroWinklerScore_ASCII_Short reports 0 B/op, 0 allocs/op (uses Jaro's [256]bool stack arrays — JW adds only a prefix loop)"
    - "Fuzz test FuzzJaroWinklerScore exists with at least one programmatic seed plus invalid-UTF-8; panic-free + score in [0,1]"
    - "BDD scenarios exercise canonical Winkler-1990 reference vectors AND the boost-threshold gate (a pair with Jaro < 0.7 returns Jaro unchanged)"
    - "testdata/golden/_staging/jarowinkler.json contains JW entries (JaroWinkler_MARTHA_MARHTA, JaroWinkler_DIXON_DICKSONX, JaroWinkler_DWAYNE_DUANE, JaroWinkler_identical, JaroWinkler_below_threshold) sorted by Name"
    - "algoid_test.go updated: AlgoJaroWinkler removed from unregistered-slots list"
    - "ExampleJaroWinklerScore runs with `// Output:` block matching byte-for-byte"
  artifacts:
    - path: "jarowinkler.go"
      provides: "JaroWinklerScore, JaroWinklerScoreRunes + Winkler 1990 constants"
      min_lines: 70
      contains: "// Source: Winkler"
    - path: "dispatch_jarowinkler.go"
      provides: "Package-load-time registration of JaroWinklerScore into dispatch[AlgoJaroWinkler]"
      contains: "dispatch[AlgoJaroWinkler] = JaroWinklerScore"
    - path: "testdata/golden/_staging/jarowinkler.json"
      provides: "Per-algorithm staging golden file for Wave 3 merge"
      contains: "JaroWinkler_MARTHA_MARHTA"
  key_links:
    - from: "jarowinkler.go JaroWinklerScore"
      to: "jaro.go JaroScore"
      via: "function call (JW computes J first, then applies boost)"
      pattern: "JaroScore\\("
    - from: "dispatch_jarowinkler.go"
      to: "algoid.go (dispatch array)"
      via: "package-level var _ = func()bool{...}()"
      pattern: "dispatch\\[AlgoJaroWinkler\\]"

user_setup: []
---

<objective>
Implement Jaro-Winkler (Winkler 1990) as a thin wrapper over JaroScore (plan 02-03) that adds a common-prefix boost capped at 4 characters and gated by a 0.7 underlying-Jaro threshold. The constants (boost threshold 0.7, prefix cap 4, prefix scale 0.1) are LOCKED per CONTEXT.md and CHAR-06's success criterion to Winkler 1990 (NOT Wikipedia).

Purpose: ship JW with zero merge collisions against the four other Wave 2 plans. Pin the canonical Winkler-1990 reference vectors (MARTHA/MARHTA → 0.9611, DIXON/DICKSONX → 0.8133) byte-stable in unit tests, golden, BDD, and the example.

Output: a working `JaroWinklerScore` that returns deterministic, 0-allocation scores on ASCII ≤ 256 chars, with all three Winkler constants traceable to the original 1990 paper via per-constant godoc citations.
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
@.planning/phases/02-core-character-algorithms-six/02-03-jaro-SUMMARY.md
@docs/requirements.md
@CLAUDE.md
@.claude/skills/algorithm-correctness-standards/SKILL.md
@.claude/skills/performance-standards/SKILL.md
@.claude/skills/determinism-standards/SKILL.md

<interfaces>
This plan depends on plan 02-03 (Jaro):

From jaro.go (plan 02-03 output):
  func JaroScore(a, b string) float64
  func JaroScoreRunes(a, b string) float64
  const maxJaroStackLen = 256   // shared if needed; JW does not redeclare

From algoid.go: AlgoJaroWinkler AlgoID = 5
From normalise.go: func isASCII(s string) bool
From dispatch_levenshtein.go: registration idiom (copy character-for-character)

From props_test.go: APPEND TestProp_JaroWinklerScore_*; SKIP triangle inequality
From example_test.go: APPEND ExampleJaroWinklerScore
From algoid_test.go: UPDATE the dispatch test (remove AlgoJaroWinkler from unregistered-slots)
From tests/bdd/steps/algorithms_steps.go: APPEND iComputeTheJaroWinklerScoreBetween + register the regex
From testdata/golden/algorithms.json: DO NOT EDIT — write to _staging/jarowinkler.json
</interfaces>

<algorithm_specifics>
**Jaro-Winkler formula (Winkler 1990 p. 357):**

  J = JaroScore(a, b)
  if J < winklerBoostThreshold { return J }    // 0.7 gate
  L = length of common prefix, capped at winklerMaxPrefix (4)
  JW = J + float64(L) * winklerPrefixScale * (1.0 - J)    // 0.1 scale

**Constants (LOCKED, traced to Winkler 1990):**
  const (
      winklerPrefixScale    = 0.1   // Winkler 1990 p. 357 — "p"
      winklerMaxPrefix      = 4     // Winkler 1990 p. 357 — "L_max"
      winklerBoostThreshold = 0.7   // Winkler 1990 p. 357 — boost gate
  )

Each constant's godoc cites Winkler 1990 page 357.

**Worked examples (RESEARCH.md §Primary Sources — Jaro-Winkler):**

MARTHA / MARHTA:
  J = 0.9444 (from plan 02-03 Jaro)
  Prefix: M=M, A=A, R=R, T!=H → L = 3 (cap not hit)
  JW = 0.9444 + 3 * 0.1 * (1 - 0.9444) = 0.9444 + 0.3 * 0.0556 = 0.9444 + 0.01667 = 0.9611

DIXON / DICKSONX:
  J = 0.7667 (above threshold)
  Prefix: D=D, I=I, X!=C → L = 2
  JW = 0.7667 + 2 * 0.1 * (1 - 0.7667) = 0.7667 + 0.2 * 0.2333 = 0.7667 + 0.0467 = 0.8133

DWAYNE / DUANE:
  J = ≈ 0.8222 (computed by Jaro)
  Prefix: D=D, W!=U → L = 1
  JW = 0.8222 + 1 * 0.1 * (1 - 0.8222) ≈ 0.8400

A pair with Jaro < 0.7 (e.g. completely-different short strings):
  J < 0.7 → return J unchanged (no boost). Test this explicitly with a chosen pair.

The prefix-cap test: pair like "TESTING" vs "TESTERS" has 4-char shared prefix (TEST); a longer-prefix pair like "TESTABCD" vs "TESTABCE" has 7-char shared prefix but L is capped at 4. Use this kind of pair as a unit test.
</algorithm_specifics>
</context>

<tasks>

<task type="auto" tdd="true">
  <name>Task 1: Implement jarowinkler.go (Jaro + prefix boost) and dispatch_jarowinkler.go</name>
  <files>jarowinkler.go, dispatch_jarowinkler.go</files>
  <read_first>
    - jaro.go (Wave 2 sibling — JW calls JaroScore directly; understand its API and behaviour)
    - levenshtein.go (canonical file structure)
    - dispatch_levenshtein.go (registration idiom)
    - normalise.go (isASCII helper, file-header pattern)
    - .planning/phases/02-core-character-algorithms-six/02-CONTEXT.md (Jaro-Winkler constants locked: 0.7 / 4 / 0.1)
    - .planning/phases/02-core-character-algorithms-six/02-RESEARCH.md §Primary Sources — Jaro-Winkler; §Determinism Constraints (left-to-right float reduction)
    - .planning/phases/02-core-character-algorithms-six/02-PATTERNS.md (Pattern 1, 2, 5, 6)
    - docs/requirements.md §7.1.6 (Jaro-Winkler spec)
    - .claude/skills/algorithm-correctness-standards/SKILL.md (per-constant paper citation)
  </read_first>
  <action>
Create `jarowinkler.go` in package fuzzymatch (filename LOCKED to no-underscore form per planning decision):

1. Apache-2.0 header copied from normalise.go lines 1-13.
2. File-level godoc opening with `// Source: Winkler, W. E. (1990). "String comparator metrics and enhanced decision rules in the Fellegi-Sunter model of record linkage." Proceedings of the Section on Survey Research Methods, American Statistical Association: 354-359.` Include the formula (J + L*p*(1-J) gated by J >= boostThreshold) in the godoc block.
3. Include the load-bearing godoc paragraph:

       // Jaro-Winkler is NOT a metric (inherits the non-metric property of
       // the underlying Jaro similarity). Triangle inequality does not hold.

4. Three unexported constants with PER-CONSTANT godoc citing Winkler 1990 page 357:

       const (
           // winklerPrefixScale is the prefix-bonus scale factor "p" from
           // Winkler 1990 p. 357. The value 0.1 is the canonical default
           // and is LOCKED for v1.x by REQUIREMENTS.md CHAR-06 + CONTEXT.md.
           winklerPrefixScale = 0.1

           // winklerMaxPrefix is the maximum effective common-prefix length
           // ("L_max" in Winkler 1990 p. 357). The value 4 is the canonical
           // cap; longer common prefixes saturate at this value.
           winklerMaxPrefix = 4

           // winklerBoostThreshold is the underlying-Jaro threshold below
           // which the prefix bonus is NOT applied (Winkler 1990 p. 357).
           // Pairs with J < 0.7 return JaroScore unchanged.
           winklerBoostThreshold = 0.7
       )

5. Public API:
     - JaroWinklerScore(a, b string) float64
     - JaroWinklerScoreRunes(a, b string) float64
6. Implementation:
     - JaroWinklerScore:
         - j := JaroScore(a, b)   // delegates to plan 02-03's Jaro
         - if j < winklerBoostThreshold { return j }
         - Compute L = common prefix length on bytes (since the underlying Jaro is byte-level), capped at winklerMaxPrefix. Loop `i := 0; i < min(len(a), len(b), winklerMaxPrefix); i++` while `a[i] == b[i]`; L = the count.
         - return `j + float64(L) * winklerPrefixScale * (1.0 - j)` — explicit parenthesisation; left-to-right.
     - JaroWinklerScoreRunes:
         - j := JaroScoreRunes(a, b)
         - if j < winklerBoostThreshold { return j }
         - Compute L on runes via `utf8.DecodeRuneInString` loop OR by `[]rune` conversion (the rune path already allocates from the underlying JaroScoreRunes call, so an additional `[]rune` is acceptable; document the cost).
         - return same formula.
7. Public function godocs starting with the function name; document delegation to Jaro and the prefix-boost gate.

Create `dispatch_jarowinkler.go` mirroring dispatch_levenshtein.go:

       // [Apache-2.0 header]
       // dispatch_jarowinkler.go registers JaroWinklerScore into the dispatch
       // table at package load time. Sole writer to dispatch[AlgoJaroWinkler].
       package fuzzymatch
       var _ = func() bool {
           dispatch[AlgoJaroWinkler] = JaroWinklerScore
           return true
       }()

After writing: `go build ./... && go vet ./... && bash scripts/verify-license-headers.sh`.
  </action>
  <verify>
    <automated>go build ./... && go vet ./... && bash scripts/verify-license-headers.sh</automated>
  </verify>
  <acceptance_criteria>
    - jarowinkler.go starts with Apache-2.0 header.
    - jarowinkler.go contains `// Source: Winkler, W. E. (1990).` literal in file-level godoc.
    - jarowinkler.go declares the three constants (winklerPrefixScale, winklerMaxPrefix, winklerBoostThreshold) each with a godoc paragraph citing Winkler 1990 page 357.
    - Constant values are exactly 0.1, 4, 0.7.
    - jarowinkler.go contains the "Jaro-Winkler is NOT a metric" paragraph.
    - `grep -E 'math\.(Pow|Log|Exp|Sqrt|FMA)' jarowinkler.go` returns no matches.
    - `grep -E '^func init\(' jarowinkler.go dispatch_jarowinkler.go` returns no matches.
    - jarowinkler.go calls JaroScore (verifiable by `grep -F 'JaroScore(' jarowinkler.go` returning at least one match) — JW is implemented as a wrapper, NOT a re-implementation of Jaro.
    - dispatch_jarowinkler.go contains `dispatch[AlgoJaroWinkler] = JaroWinklerScore` exactly once and uses var_ idiom.
    - `go build ./...` exits 0; `go vet ./...` exits 0.
  </acceptance_criteria>
  <behavior>
    - JaroWinklerScore("MARTHA", "MARHTA") within 1e-6 of 0.9611111111
    - JaroWinklerScore("DIXON", "DICKSONX") within 1e-6 of 0.8133333333
    - JaroWinklerScore("DWAYNE", "DUANE") within 1e-3 of 0.8400
    - JaroWinklerScore("ABC", "ABC") == 1.0
    - JaroWinklerScore("", "") == 1.0
    - JaroWinklerScore("", "ABC") == 0.0
    - For a pair with underlying Jaro < 0.7, JaroWinklerScore returns the unmodified Jaro value (boost gated)
    - For a pair with 7-char common prefix (like "TESTABCD" vs "TESTABCE"), L is capped at 4 (boost = 4 * 0.1 * (1-J), not 7 * 0.1 * (1-J))
    - dispatch[AlgoJaroWinkler] non-nil after package load
  </behavior>
  <done>
    jarowinkler.go and dispatch_jarowinkler.go committed. Package builds. Three Winkler constants present and documented.
  </done>
</task>

<task type="auto" tdd="true">
  <name>Task 2: Tests (unit + property + benchmark + fuzz) + extend props_test.go, example_test.go, algoid_test.go</name>
  <files>jarowinkler_test.go, jarowinkler_bench_test.go, jarowinkler_fuzz_test.go, props_test.go, example_test.go, algoid_test.go, testdata/fuzz/FuzzJaroWinklerScore/seed-001</files>
  <read_first>
    - jaro_test.go (sibling Wave 2 plan output — copy structure)
    - levenshtein_test.go, levenshtein_bench_test.go, levenshtein_fuzz_test.go (Wave 1 templates)
    - props_test.go (existing — APPEND, do not recreate)
    - example_test.go (existing — APPEND ExampleJaroWinklerScore)
    - algoid_test.go (Wave 1's dispatch test — remove AlgoJaroWinkler from unregistered-slots list)
    - .planning/phases/02-core-character-algorithms-six/02-RESEARCH.md §Mathematical Invariants (JW is NOT a metric — DO NOT add triangle-inequality property)
    - .planning/phases/02-core-character-algorithms-six/02-PATTERNS.md (Patterns 7, 9, 10, 11, 14)
  </read_first>
  <action>
Create `jarowinkler_test.go` (package fuzzymatch_test, stdlib testing only):

1. Apache-2.0 header + file-level godoc.
2. Tests:
     - TestJaroWinkler_BothEmpty — Score("","") == 1.0
     - TestJaroWinkler_OneEmpty — Score("","ABC") and Score("ABC","") == 0.0
     - TestJaroWinkler_Identical — Score("ABC","ABC") == 1.0
     - TestJaroWinkler_ReferenceVectors — table-driven over MARTHA/MARHTA → 0.9611, DIXON/DICKSONX → 0.8133, DWAYNE/DUANE → 0.8400 (looser tolerance ~1e-3 for DWAYNE/DUANE per RESEARCH.md notation; tighter 1e-6 for the two pinned-precision pairs).
     - TestJaroWinkler_BoostThresholdGate — pick a pair with Jaro < 0.7 (e.g. "abc" vs "xyz" — likely 0 matches → Jaro = 0; or compute one with Jaro just under 0.7 and verify JW returns Jaro unchanged). Assert `JaroWinklerScore(a, b) == JaroScore(a, b)` for that pair.
     - TestJaroWinkler_PrefixCapAt4 — input pair with 7+ character common prefix (e.g. "TESTABCD" vs "TESTABCE"); compute the expected JW with L=4 (not L=7) and assert match within 1e-9.
     - TestJaroWinkler_Symmetry — Score(a,b) == Score(b,a) for the reference vectors.
     - TestJaroWinkler_ConstantsTraceable — sanity test reading the constants via an internal export (declare an `export_test.go` re-export `WinklerPrefixScaleForTest = winklerPrefixScale`, similarly for MaxPrefix and BoostThreshold) and asserting their values are exactly 0.1, 4, 0.7. This pins the constants against accidental modification.
     - TestJaroWinklerScore_ZeroAllocs_ASCII_Short — `testing.AllocsPerRun(100, func() { _ = fuzzymatch.JaroWinklerScore("MARTHA", "MARHTA") })` must be 0.

Note on the ConstantsTraceable test: the simplest implementation extends `export_test.go` with three re-exports:

       const WinklerPrefixScaleForTest    = winklerPrefixScale
       const WinklerMaxPrefixForTest      = winklerMaxPrefix
       const WinklerBoostThresholdForTest = winklerBoostThreshold

Then `jarowinkler_test.go` asserts the values through the re-exports. This is the same export_test.go pattern Wave 1 plans 01-04 / 01-05 established.

Create `jarowinkler_bench_test.go`:

1. Apache-2.0 header + file-level godoc citing PERF-01 (0 allocs on ASCII Short).
2. Benchmarks:
     - BenchmarkJaroWinklerScore_ASCII_Short (MARTHA / MARHTA)
     - BenchmarkJaroWinklerScore_ASCII_Medium (50-char identifier pair)
     - BenchmarkJaroWinklerScore_ASCII_Long (300-char pair, exceeds maxJaroStackLen)
     - BenchmarkJaroWinklerScore_Unicode_Short
   b.ReportAllocs() before b.ResetTimer(); var sink pattern.

Create `jarowinkler_fuzz_test.go`:

1. Apache-2.0 header + file-level godoc.
2. FuzzJaroWinklerScore — programmatic seeds: MARTHA/MARHTA, DIXON/DICKSONX, DWAYNE/DUANE, ""/"ABC", invalid-UTF-8. Body: no panic, !math.IsNaN, !math.IsInf, score in [0,1].
3. Create testdata/fuzz/FuzzJaroWinklerScore/seed-001 with MARTHA/MARHTA.

Extend `props_test.go` (APPEND):

1. TestProp_JaroWinklerScore_RangeBounds
2. TestProp_JaroWinklerScore_Identity (skip empty)
3. TestProp_JaroWinklerScore_Symmetric
4. TestProp_JaroWinklerScore_NoNaN
5. TestProp_JaroWinklerScore_NoInf
6. TestProp_JaroWinklerScore_NoNegativeZero
7. NO triangle inequality (JW inherits Jaro's non-metric status).
8. (Optional but useful) TestProp_JaroWinklerScore_AtLeastJaro — for any a, b where JaroScore(a,b) >= 0.7, JaroWinklerScore(a,b) >= JaroScore(a,b). The boost is non-negative when the gate is open.

Extend `example_test.go` (APPEND):

       func ExampleJaroWinklerScore() {
           fmt.Printf("%.4f\n", fuzzymatch.JaroWinklerScore("MARTHA", "MARHTA"))
           // Output:
           // 0.9611
       }

Extend `algoid_test.go`:

1. Remove AlgoJaroWinkler from the unregistered-slots list. Add registered-slot assertion `DispatchEntryNilForTest(int(fuzzymatch.AlgoJaroWinkler)) == false`.

Run:
  go test -race -shuffle=on -count=1 -run 'TestJaroWinkler|TestProp_JaroWinkler|TestDispatch|ExampleJaroWinklerScore' ./...
  go test -bench=BenchmarkJaroWinklerScore_ASCII -benchmem -run=^$ -count=3 ./...
  go test -fuzz=FuzzJaroWinklerScore -fuzztime=30s -run=^$ ./...
  </action>
  <verify>
    <automated>go test -race -shuffle=on -count=1 -run 'TestJaroWinkler|TestProp_JaroWinkler|TestDispatch|ExampleJaroWinklerScore' ./... && go test -bench=BenchmarkJaroWinklerScore_ASCII_Short -benchmem -run=^$ -count=3 ./... 2>&1 | grep -E '0 B/op[[:space:]]+0 allocs/op'</automated>
  </verify>
  <acceptance_criteria>
    - All TestJaroWinkler_* tests pass — including BoostThresholdGate, PrefixCapAt4, ConstantsTraceable.
    - All TestProp_JaroWinkler* tests pass.
    - TestJaroWinklerScore_ZeroAllocs_ASCII_Short reports 0 allocations.
    - BenchmarkJaroWinklerScore_ASCII_Short reports `0 B/op  0 allocs/op`.
    - ExampleJaroWinklerScore output matches `0.9611\n` byte-for-byte.
    - testdata/fuzz/FuzzJaroWinklerScore/seed-001 exists.
    - algoid_test.go: AlgoJaroWinkler slot non-nil; remaining unregistered slots still pass nil assertion.
    - export_test.go has the three Winkler constant re-exports.
    - `grep -c '"github.com/stretchr/testify' jarowinkler_test.go jarowinkler_bench_test.go jarowinkler_fuzz_test.go` returns 0.
    - props_test.go contains no TestProp_JaroWinklerDistance_TriangleInequality (no Distance variant; not a metric).
  </acceptance_criteria>
  <behavior>
    - Reference vectors pin JW accuracy at 1e-6 on the canonical Winkler-1990 pairs.
    - BoostThresholdGate test verifies JW = Jaro for pairs below threshold.
    - PrefixCapAt4 test verifies L is capped at 4 for long-prefix pairs.
    - ConstantsTraceable test pins 0.1 / 4 / 0.7 against drift.
    - Property tests cover RangeBounds, Identity, Symmetric, NoNaN, NoInf, NoNegativeZero plus the optional AtLeastJaro monotonicity.
    - Benchmark pins 0-alloc target on ASCII Short.
    - Fuzz harness panic-free + invariant-preserving.
  </behavior>
  <done>
    All test files committed; props_test, example_test, algoid_test, export_test extended; full JW-scoped test suite green.
  </done>
</task>

<task type="auto" tdd="true">
  <name>Task 3: Per-algorithm staging golden file + BDD feature + extend BDD steps</name>
  <files>testdata/golden/_staging/jarowinkler.json, tests/bdd/features/jarowinkler.feature, tests/bdd/steps/algorithms_steps.go, algorithms_golden_test.go</files>
  <read_first>
    - algorithms_golden_test.go (Wave 1 — staging-write helper)
    - testdata/golden/_staging/jaro.json (sibling Wave 2 file; reference for staging file form)
    - tests/bdd/features/jaro.feature (sibling Wave 2 BDD pattern)
    - tests/bdd/steps/algorithms_steps.go (current state)
    - .planning/phases/02-core-character-algorithms-six/02-PATTERNS.md (Pattern 12, Pattern 15)
    - .planning/phases/02-core-character-algorithms-six/02-RESEARCH.md §Golden File Integration; §BDD Scenario Coverage
  </read_first>
  <action>
Create `testdata/golden/_staging/jarowinkler.json`:

1. Schema matches algorithms.json.
2. Entries (sorted by Name):
     - JaroWinkler_below_threshold (a "abc", b "xyz" or another known sub-0.7-Jaro pair, expected_score from a live JaroWinklerScore call — should equal the underlying JaroScore for the same pair)
     - JaroWinkler_DIXON_DICKSONX (a "DIXON", b "DICKSONX")
     - JaroWinkler_DWAYNE_DUANE (a "DWAYNE", b "DUANE")
     - JaroWinkler_empty_empty (a "", b "")
     - JaroWinkler_identical (a "ABC", b "ABC")
     - JaroWinkler_MARTHA_MARHTA (a "MARTHA", b "MARHTA")
     - JaroWinkler_one_empty (a "", b "ABC")
     - JaroWinkler_prefix_cap (a "TESTABCD", b "TESTABCE")
3. Generate via the staging-write helper extension to algorithms_golden_test.go:

       func TestGolden_JaroWinkler_Staging(t *testing.T) {
           if !*updateGolden { t.Skip("only runs with -update; produces _staging/jarowinkler.json") }
           ... build, sort, write via CanonicalMarshalForTest
       }

   Run with `-update` once; commit the file. Re-run without `-update` and confirm zero diff.

Create `tests/bdd/features/jarowinkler.feature`:

1. File-level comment with primary source (Winkler 1990).
2. Feature: Jaro-Winkler similarity algorithm
3. Scenario Outline "canonical reference vectors" with rows for MARTHA/MARHTA → 0.9611, DIXON/DICKSONX → 0.8133, DWAYNE/DUANE → 0.8400 (tolerance 0.001 for DWAYNE/DUANE; 0.0001 for the others), ABC/ABC → 1.0000.
4. Scenario "boost gated by threshold" — pair with Jaro < 0.7 (use the same pair from the unit test); compute JW; assert JW score equals the underlying Jaro score (use a step "the JW score should equal the Jaro score for the same pair" — extends BDD steps if needed).
5. Scenario "prefix length capped at 4" — pair with 7-char shared prefix; assert score within tolerance of the L=4-bonus expectation.
6. Scenario "score is symmetric".

Extend `tests/bdd/steps/algorithms_steps.go` (APPEND):

1. iComputeTheJaroWinklerScoreBetween(a, b string) error → calls fuzzymatch.JaroWinklerScore.
2. (If needed) compareJaroAndJaroWinklerForPair(a, b string) error → calls both, stores both in AlgorithmContext, exposes a "the scores should be equal within tolerance" step.
3. Register the regexes in the existing InitializeScenario.

Run:
  go test -race -shuffle=on -count=1 -run 'TestGolden_JaroWinkler_Staging|TestGolden_Algorithms' ./...
  (cd tests/bdd && go test -race -shuffle=on -count=1 ./...)
  </action>
  <verify>
    <automated>go test -race -shuffle=on -count=1 -run 'TestGolden_JaroWinkler_Staging|TestGolden_Algorithms' ./... && (cd tests/bdd && go test -race -shuffle=on -count=1 ./...)</automated>
  </verify>
  <acceptance_criteria>
    - testdata/golden/_staging/jarowinkler.json exists with eight entries sorted alphabetically by Name.
    - Canonical form: 2-space indent, trailing LF, no BOM.
    - Re-running TestGolden_JaroWinkler_Staging without -update produces no diff.
    - tests/bdd/features/jarowinkler.feature includes the boost-threshold-gated scenario AND the prefix-cap-at-4 scenario AND the canonical reference vectors.
    - `cd tests/bdd && go test -count=1 ./...` exits 0.
    - tests/bdd/steps/algorithms_steps.go still has exactly one AlgorithmContext type and one InitializeScenario function.
    - testdata/golden/algorithms.json UNCHANGED relative to Wave 1.
  </acceptance_criteria>
  <behavior>
    - Staging file ready for Wave 3 plan 02-07 merge with comprehensive entry coverage (canonical pairs, edge cases, threshold gate, prefix cap).
    - BDD harness exercises the boost-gate and prefix-cap behaviour explicitly — these are Winkler-1990-specific contracts not covered by Jaro alone.
  </behavior>
  <done>
    Staging golden file, JW BDD feature, and JW BDD step bindings committed. JW-scoped quality gate green.
  </done>
</task>

</tasks>

<threat_model>
## Trust Boundaries

| Boundary | Description |
|----------|-------------|
| Caller → fuzzymatch.JaroWinklerScore | Untrusted strings. Pure function delegating to JaroScore plus a prefix loop. |

## STRIDE Threat Register

| Threat ID | Category | Component | Disposition | Mitigation Plan |
|-----------|----------|-----------|-------------|-----------------|
| T-02-04-01 | Denial of Service | JW prefix loop on adversarial inputs | accept | Prefix loop is bounded at min(len(a), len(b), winklerMaxPrefix=4). O(1) — constant-bounded extra work over Jaro. JW inherits Jaro's O(la·lb) worst case from JaroScore but adds nothing. |
| T-02-04-02 | Information Disclosure | Malformed UTF-8 input causing panic | mitigate | Byte-level prefix loop operates on bytes — invalid UTF-8 cannot panic. Underlying Jaro already mitigates this (T-02-03-02). FuzzJaroWinklerScore in Task 2 includes invalid-UTF-8 seeds. |
| T-02-04-03 | Tampering | Winkler constants drift to Wikipedia values (e.g. 0.25 prefix scale instead of 0.1) | mitigate | TestJaroWinkler_ConstantsTraceable pins 0.1 / 4 / 0.7 against drift. Per-constant godoc cites Winkler 1990 page 357. algorithm-correctness-reviewer agent will reject any PR modifying these constants without paper-citation evidence. |
| T-02-04-04 | Tampering | dispatch[AlgoJaroWinkler] overwritten | mitigate | dispatch is unexported; registration runs once at package load. Same mitigation as T-02-01-04. |
| T-02-04-05 | Information Disclosure | Float-determinism violation (especially the multiplication chain `L * winklerPrefixScale * (1 - J)`) | mitigate | Explicit left-to-right evaluation order; no math.FMA; no math.Pow. The three-term multiplication is associativity-stable in IEEE-754 left-to-right; cross-platform CI matrix verifies via `_staging/jarowinkler.json` + `make verify-determinism`. |
| T-02-04-06 | Repudiation | "Jaro-Winkler is a metric" misuse | mitigate | jarowinkler.go file-level godoc states "Jaro-Winkler is NOT a metric". props_test.go documents the omitted triangle-inequality test. jarowinkler.feature inherits this lock by reference. |

No high-severity items. Plan passes the security gate.
</threat_model>

<verification>
1. `go build ./...` exits 0.
2. `go vet ./...` exits 0.
3. `bash scripts/verify-license-headers.sh` exits 0.
4. `bash scripts/verify-no-runtime-deps.sh` exits 0.
5. `go test -race -shuffle=on -count=1 -run 'TestJaroWinkler|TestProp_JaroWinkler|TestDispatch|ExampleJaroWinklerScore|TestGolden_JaroWinkler_Staging' ./...` exits 0.
6. `go test -bench=BenchmarkJaroWinklerScore_ASCII_Short -benchmem -run=^$ -count=3 ./...` reports `0 B/op  0 allocs/op`.
7. `go test -fuzz=FuzzJaroWinklerScore -fuzztime=30s -run=^$ ./...` completes without crash.
8. `(cd tests/bdd && go test -race -shuffle=on -count=1 ./...)` exits 0.
9. testdata/golden/_staging/jarowinkler.json byte-stable.
10. testdata/golden/algorithms.json UNCHANGED.
11. `make check` exits 0.
</verification>

<success_criteria>
- JW reference vectors (MARTHA/MARHTA → 0.9611, DIXON/DICKSONX → 0.8133, DWAYNE/DUANE → 0.8400) pin Winkler-1990 accuracy.
- Three Winkler constants (0.1 / 4 / 0.7) declared with per-constant Winkler-1990-page-357 godoc citation, pinned against drift via TestJaroWinkler_ConstantsTraceable.
- Boost gate (J >= 0.7) and prefix cap (L <= 4) verified by dedicated unit tests AND BDD scenarios.
- Zero allocations on ASCII Short (PERF-01 satisfied for JW; same allocation profile as Jaro since JW adds only a constant-bounded prefix loop).
- No NaN / Inf / -0 (DET-04 satisfied for JW).
- "Jaro-Winkler is not a metric" lock documented in godoc and the property test omission.
- Per-algorithm staging golden file ready for Wave 3 merge.
- All required gates green via `make check`.
</success_criteria>

<output>
After completion, create `.planning/phases/02-core-character-algorithms-six/02-04-jarowinkler-SUMMARY.md` recording:
- Final identifier names confirmed (JaroWinklerScore, JaroWinklerScoreRunes; constants winklerPrefixScale, winklerMaxPrefix, winklerBoostThreshold).
- File name shipped: `jarowinkler.go` (no underscore — locked planning decision).
- Benchmark numbers observed.
- The exact Winkler-1990 page citation text used in each constant's godoc.
- Coverage percentages.
- Any deviations from the plan and their rationale.
</output>
