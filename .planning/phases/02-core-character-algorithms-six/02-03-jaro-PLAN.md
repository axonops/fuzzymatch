---
phase: 02-core-character-algorithms-six
plan: 03
type: execute
wave: 2
depends_on: [02-01-levenshtein]
files_modified:
  - jaro.go
  - dispatch_jaro.go
  - jaro_test.go
  - jaro_bench_test.go
  - jaro_fuzz_test.go
  - props_test.go
  - example_test.go
  - algoid_test.go
  - testdata/golden/_staging/jaro.json
  - testdata/fuzz/FuzzJaroScore/seed-001
  - tests/bdd/features/jaro.feature
  - tests/bdd/steps/algorithms_steps.go
autonomous: true
requirements:
  - CHAR-05
  - PERF-01
  - PERF-02
  - TEST-01
  - TEST-02
  - TEST-04
  - TEST-05
  - DET-04
  - DX-02
tags: [jaro, match-flag-arrays, no-dp, not-a-metric]

must_haves:
  truths:
    - "JaroScore(\"MARTHA\", \"MARHTA\") returns 0.9444444444 (within 1e-6) — Jaro 1989 / Winkler 1990 canonical pair"
    - "JaroScore(\"DIXON\", \"DICKSONX\") returns 0.7666666666 (within 1e-6) — Winkler 1990 canonical pair"
    - "JaroScore(\"JELLYFISH\", \"SMELLYFISH\") returns 0.8962962962 (within 1e-6) — Jaro 1989 reference vector"
    - "JaroScore(\"\", \"\") returns 1.0 exactly (both-empty convention per RESEARCH.md §Score Normalisation)"
    - "JaroScore(\"\", \"ABC\") returns 0.0 exactly (one-empty)"
    - "JaroScore(\"ABC\", \"ABC\") returns 1.0 exactly (identity)"
    - "JaroScoreRunes correctly handles multi-byte UTF-8 (rune-aware matching window)"
    - "dispatch[AlgoJaro] is non-nil and equals JaroScore after package init"
    - "Apache-2.0 header on every new .go file (verified by scripts/verify-license-headers.sh)"
    - "jaro.go contains `// Source: Jaro, M. A. (1989).` block at top of file-level godoc"
    - "jaro.go file-level godoc EXPLICITLY states: 'Jaro is not a metric; the triangle inequality does not hold.' (per RESEARCH.md §Mathematical Invariants)"
    - "No math.Pow / math.Log / math.Exp / math.Sqrt / math.FMA in jaro.go (verified by grep)"
    - "Jaro formula uses left-to-right float reduction: `(m/lenA + m/lenB + (m-t)/m) / 3.0` — explicit parenthesisation per docs/requirements.md §13"
    - "No init() function; dispatch_jaro.go uses var _ = func() bool { ... }() registration idiom"
    - "Jaro division guard `if m == 0 { return 0.0 }` present (NaN/Inf prevention)"
    - "Property tests pass for RangeBounds, Identity, Symmetric, NoNaN, NoInf, NoNegativeZero in props_test.go (NO triangle inequality — Jaro is not a metric)"
    - "Benchmark BenchmarkJaroScore_ASCII_Short reports 0 B/op, 0 allocs/op (uses [256]bool stack arrays for inputs ≤ 256)"
    - "Fuzz test FuzzJaroScore exists with at least one programmatic seed plus invalid-UTF-8; panic-free + score in [0,1]"
    - "BDD scenarios in tests/bdd/features/jaro.feature exercise canonical reference vectors (MARTHA/MARHTA, DIXON/DICKSONX, JELLYFISH/SMELLYFISH)"
    - "testdata/golden/_staging/jaro.json contains Jaro entries (Jaro_MARTHA_MARHTA, Jaro_DIXON_DICKSONX, Jaro_identical, Jaro_empty_empty, Jaro_one_empty) sorted by Name"
    - "algoid_test.go updated: AlgoJaro removed from unregistered-slots list"
    - "ExampleJaroScore runs with `// Output:` block matching byte-for-byte"
  artifacts:
    - path: "jaro.go"
      provides: "JaroScore, JaroScoreRunes + unexported jaroDP helper using match-flag arrays"
      min_lines: 100
      contains: "// Source: Jaro"
    - path: "dispatch_jaro.go"
      provides: "Package-load-time registration of JaroScore into dispatch[AlgoJaro]"
      contains: "dispatch[AlgoJaro] = JaroScore"
    - path: "testdata/golden/_staging/jaro.json"
      provides: "Per-algorithm staging golden file for Wave 3 merge"
      contains: "Jaro_MARTHA_MARHTA"
  key_links:
    - from: "dispatch_jaro.go"
      to: "algoid.go (dispatch array)"
      via: "package-level var _ = func()bool{...}()"
      pattern: "dispatch\\[AlgoJaro\\]"

user_setup: []
---

<objective>
Implement Jaro similarity (Jaro 1989) using match-flag arrays (NOT DP) with a stack-allocated `[256]bool` × 2 buffer for inputs ≤ 256 bytes. Jaro is the second of the two non-DP algorithms in Phase 2 (the first being Hamming). The plan follows the canonical Wave 1 pattern but introduces a different stack-buffer shape (`[256]bool` instead of `[(maxStackInputLen+1)*2]int`) and a different normalisation formula (formula yields `[0,1]` directly; no `1 - dist/maxLen` form).

Purpose: ship Jaro with zero merge collisions against the four other Wave 2 plans. Lock the matching-window calculation, the transposition counting, and the three-term left-to-right float reduction — all of which are reused by plan 02-04 (Jaro-Winkler) which builds on JaroScore directly.

Output: a working `JaroScore` zero-allocation on ASCII ≤ 256 bytes, with Winkler-1990-traceable reference vectors pinned in tests, golden, BDD, and the example. Plan 02-04 (JW) consumes JaroScore as its base; this plan owns the Jaro reference-vector accuracy contract.
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
@docs/requirements.md
@CLAUDE.md
@.claude/skills/algorithm-correctness-standards/SKILL.md
@.claude/skills/performance-standards/SKILL.md
@.claude/skills/determinism-standards/SKILL.md
@.claude/skills/go-testing-standards/SKILL.md

<interfaces>
Wave 1 append-points (DO NOT recreate):

From algoid.go: AlgoJaro AlgoID = 4
From normalise.go: func isASCII(s string) bool
From levenshtein.go: const maxStackInputLen = 64 (Jaro does NOT use this — Jaro uses [256]bool, a different stack threshold; document why in jaro.go's godoc)
From dispatch_levenshtein.go: copy the registration idiom character-for-character
From algorithms_golden_test.go (plan 02-01 Task 3 — defines the canonical staging-write helper for the entire phase):
  func assertGoldenStaging(t *testing.T, relPath string, v any)
  // Signature LOCKED. Call directly — no "if helper exists / else create" branch.
  // Writes to testdata/golden/<relPath> (e.g. "_staging/jaro.json") via WriteGoldenFile.
From props_test.go: APPEND TestProp_JaroScore_*; SKIP triangle inequality (Jaro is not a metric)
From example_test.go: APPEND ExampleJaroScore
From algoid_test.go: UPDATE the dispatch test (remove AlgoJaro from unregistered-slots)
From tests/bdd/steps/algorithms_steps.go: APPEND iComputeTheJaroScoreBetween + register the regex inside the existing InitializeScenario
From testdata/golden/algorithms.json: DO NOT EDIT — write to _staging/jaro.json instead
</interfaces>

<algorithm_specifics>
**Jaro formula (canonical):**

  matching window w = floor(max(|s1|, |s2|) / 2) - 1   (clamp to >= 0)
  m = matched-character count (each position in s1 matches the nearest s2 position within w)
  t = number of transpositions among matched pairs / 2

  if m == 0: return 0.0
  J = (m/|s1| + m/|s2| + (m-t)/m) / 3.0

The `t` count is computed AFTER matched-character positions are identified: walk matched positions in s1 in order, walk matched positions in s2 in order, count mismatches between the two sequences, then halve.

**Worked example — MARTHA vs MARHTA (RESEARCH.md §Primary Sources — Jaro):**
  - w = floor(max(6,6)/2) - 1 = 2
  - Matches: M-M, A-A, R-R, H-H, T-T, A-A → m = 6
  - Matched s1 sequence: M,A,R,H,T,A (positions 0,1,2,3,4,5 — but H is at s1[4] and T at s1[3], so the in-order match positions need careful walking)
  - Matched s2 sequence: M,A,R,H,T,A — but in s2 the order at the H/T positions is reversed
  - Transpositions: T and H swap → 2 mismatches → t = 1
  - J = (6/6 + 6/6 + (6-1)/6) / 3 = (1 + 1 + 0.8333…) / 3 = 0.9444…

Reference vectors expected to pass exact-tolerance (1e-6) tests:
  - MARTHA / MARHTA → 0.9444…  (Winkler 1990 / Jaro 1989)
  - DIXON  / DICKSONX → 0.7666… (Winkler 1990)
  - JELLYFISH / SMELLYFISH → 0.8962… (Jaro 1989 table)
  - ABC / ABC → 1.0 (identity)
  - "" / "" → 1.0 (both-empty convention)
  - "" / "ABC" → 0.0 (one-empty)
</algorithm_specifics>
</context>

<tasks>

<task type="auto" tdd="true">
  <name>Task 1: Implement jaro.go (match-flag arrays + transposition count) and dispatch_jaro.go</name>
  <files>jaro.go, dispatch_jaro.go</files>
  <read_first>
    - levenshtein.go (Wave 1 canonical file structure — Apache-2.0 header, file-level godoc, public-API godoc style)
    - dispatch_levenshtein.go (registration idiom — copy character-for-character)
    - normalise.go (isASCII helper at lines 159-168 — call directly)
    - .planning/phases/02-core-character-algorithms-six/02-RESEARCH.md §Primary Sources — Jaro; §Implementation Patterns — Jaro Match-Flag Arrays; §Score Normalisation — Jaro row; §Determinism Constraints — left-to-right float reduction
    - .planning/phases/02-core-character-algorithms-six/02-PATTERNS.md (Pattern 2 file structure; Pattern 5 score normalisation; Pattern 6 dispatch; PATTERN_DOC_ON_NOT_A_METRIC: jaro.go godoc must explicitly say "Jaro is not a metric; triangle inequality does not hold")
    - docs/requirements.md §7.1.5 (Jaro spec)
    - .claude/skills/determinism-standards/SKILL.md (no transcendental float ops; explicit parenthesisation; left-to-right reduction)
  </read_first>
  <action>
Create `jaro.go` in package fuzzymatch:

1. Apache-2.0 header copied from normalise.go lines 1-13.
2. File-level godoc opening with `// Source: Jaro, M. A. (1989). "Advances in record-linkage methodology as applied to matching the 1985 census of Tampa, Florida." Journal of the American Statistical Association, 84(406):414-420.` Include the formula (matching window, transposition count, three-term mean) in the godoc block.
3. Include the LOAD-BEARING godoc paragraph explicitly stating Jaro is not a metric:

       // Jaro is NOT a metric: the triangle inequality does not hold for
       // Jaro similarity. Callers reasoning about distances should use
       // Levenshtein, Damerau-Levenshtein OSA, or Damerau-Levenshtein Full
       // instead.

4. Implementation discipline bullet list mirrored from normalise.go (no init, no map iteration, no transcendental floats).
5. Public API:
     - JaroScore(a, b string) float64
     - JaroScoreRunes(a, b string) float64
   (Jaro does NOT have a Distance variant — the formula yields a similarity in [0,1] directly. Document this design choice in the file-level godoc.)
6. Implementation:
     - Identity fast path: `if a == b` → return 1.0 (covers both-empty and identical inputs).
     - One-empty: `if len(a) == 0 || len(b) == 0` → return 0.0.
     - Compute matching window: `w := max(la, lb)/2 - 1; if w < 0 { w = 0 }`.
     - ASCII-bounded fast path: when both inputs are ≤ 256 bytes (use a named constant `maxJaroStackLen = 256` declared at the top of jaro.go) AND `isASCII(a) && isASCII(b)`, use stack-allocated `var matchA [256]bool; var matchB [256]bool` and pass slices `matchA[:la], matchB[:lb]` to the inner kernel.
     - Heap path: `make([]bool, la), make([]bool, lb)` for inputs > 256 bytes or non-ASCII.
     - Inner kernel `jaroBytes(a, b string, matchA, matchB []bool, w int) float64`:
         - First pass: for each i in s1, scan window [max(0, i-w), min(lb, i+w+1)] in s2 looking for an unmatched character matching a[i]. Mark both matchA[i] and matchB[j] true. Count m.
         - If m == 0, return 0.0.
         - Second pass: walk matched positions in s1 in order, walk matched positions in s2 in order, count position pairs that mismatch in character → halve to get t.
         - Return `(float64(m)/float64(la) + float64(m)/float64(lb) + float64(m-t)/float64(m)) / 3.0` — explicit parenthesisation, left-to-right.
     - JaroScoreRunes: eager `ra := []rune(a); rb := []rune(b)`; separate inner kernel operating on `[]rune`. Document the 2-alloc cost.
7. NO init(), NO `math.X` other than potentially `math.Abs` (none required for Jaro), NO map iteration.

Create `dispatch_jaro.go` mirroring dispatch_levenshtein.go exactly:

       // [Apache-2.0 header]
       // dispatch_jaro.go registers JaroScore into the dispatch table at
       // package load time. Sole writer to dispatch[AlgoJaro].
       package fuzzymatch
       var _ = func() bool {
           dispatch[AlgoJaro] = JaroScore
           return true
       }()

After writing: `go build ./... && go vet ./... && bash scripts/verify-license-headers.sh`.

Cite Winkler 1990 + Jaro 1989 in the constants section if any constants are declared (e.g. matching-window divisor 2 — declare as `const jaroMatchWindowDivisor = 2` if helpful for documentation; otherwise inline).
  </action>
  <verify>
    <automated>go build ./... && go vet ./... && bash scripts/verify-license-headers.sh</automated>
  </verify>
  <acceptance_criteria>
    - jaro.go starts with Apache-2.0 header (`diff <(head -13 normalise.go) <(head -13 jaro.go)` exit 0).
    - jaro.go contains `// Source: Jaro, M. A. (1989).` literal in file-level godoc.
    - jaro.go contains the EXACT "Jaro is NOT a metric" paragraph.
    - `grep -E 'math\.(Pow|Log|Exp|Sqrt|FMA)' jaro.go` returns no matches.
    - `grep -E '^func init\(' jaro.go dispatch_jaro.go` returns no matches.
    - `grep -c '\[\]byte\(' jaro.go` returns 0 (no allocating byte conversions on hot path).
    - jaro.go declares `const maxJaroStackLen = 256` exactly once.
    - jaro.go does NOT redeclare `maxStackInputLen` (owned by levenshtein.go).
    - dispatch_jaro.go contains `dispatch[AlgoJaro] = JaroScore` exactly once and uses the var_ idiom.
    - Float reduction in JaroScore uses explicit parens: `(float64(m)/float64(la) + float64(m)/float64(lb) + float64(m-t)/float64(m)) / 3.0` — verifiable by `grep -F '/ 3.0' jaro.go` returning at least one match.
    - `go build ./...` exits 0; `go vet ./...` exits 0.
  </acceptance_criteria>
  <behavior>
    - JaroScore("MARTHA", "MARHTA") within 1e-6 of 0.9444444444…
    - JaroScore("DIXON", "DICKSONX") within 1e-6 of 0.7666666666…
    - JaroScore("JELLYFISH", "SMELLYFISH") within 1e-6 of 0.8962962962…
    - JaroScore("ABC", "ABC") == 1.0 exactly
    - JaroScore("", "") == 1.0 exactly
    - JaroScore("", "ABC") == 0.0 exactly
    - JaroScore symmetric: Score(a,b) == Score(b,a)
    - JaroScoreRunes("café", "cafe") returns the rune-aware Jaro score (3 matches in shared positions)
    - dispatch[AlgoJaro] non-nil after package load and equals JaroScore
  </behavior>
  <done>
    jaro.go and dispatch_jaro.go committed. Package builds. License-header verifier passes.
  </done>
</task>

<task type="auto" tdd="true">
  <name>Task 2: Tests (unit + property + benchmark + fuzz) + extend props_test.go, example_test.go, algoid_test.go</name>
  <files>jaro_test.go, jaro_bench_test.go, jaro_fuzz_test.go, props_test.go, example_test.go, algoid_test.go, testdata/fuzz/FuzzJaroScore/seed-001</files>
  <read_first>
    - levenshtein_test.go, levenshtein_bench_test.go, levenshtein_fuzz_test.go (Wave 1 templates — copy structure)
    - hamming_test.go (Wave 2 sibling pattern, if landed first; not required for execution order — Wave 2 plans are independent)
    - props_test.go (the existing Wave 1 file — APPEND, do not recreate)
    - example_test.go (Wave 1 file — APPEND ExampleJaroScore)
    - algoid_test.go (Wave 1's updated dispatch test — remove AlgoJaro from unregistered-slots list)
    - .planning/phases/02-core-character-algorithms-six/02-RESEARCH.md §Mathematical Invariants (Jaro is NOT a metric — DO NOT add triangle-inequality property test); §Allocation Budgets — Jaro row
    - .planning/phases/02-core-character-algorithms-six/02-PATTERNS.md (Patterns 7, 9, 10, 11, 14)
  </read_first>
  <action>
Create `jaro_test.go` (package fuzzymatch_test, stdlib testing only):

1. Apache-2.0 header + file-level godoc.
2. Tests:
     - TestJaro_BothEmpty — Score("", "") == 1.0
     - TestJaro_OneEmpty — Score("", "ABC") and Score("ABC", "") both exactly 0.0
     - TestJaro_Identical — Score("ABC", "ABC") == 1.0
     - TestJaro_ReferenceVectors — table-driven over MARTHA/MARHTA, DIXON/DICKSONX, JELLYFISH/SMELLYFISH (tolerance 1e-6 — Jaro reference values are conventionally pinned to 4-decimal precision; using 1e-6 is conservative).
     - TestJaro_Symmetry — Score(a,b) == Score(b,a) for the reference vectors.
     - TestJaro_ScoreRunes_MultiByte — café / cafe returns the rune-aware Jaro score (do NOT pin a specific value — just assert it is in [0,1] and != the byte-level result for the same input, to verify the rune path is engaged).
     - TestJaro_ZeroAllocs_ASCII_Short — `testing.AllocsPerRun(100, func() { _ = fuzzymatch.JaroScore("MARTHA", "MARHTA") })` must be 0.

Create `jaro_bench_test.go`:

1. Apache-2.0 header + file-level godoc citing PERF-01 budget (0 allocs on ASCII ≤ 256 chars).
2. Benchmarks:
     - BenchmarkJaroScore_ASCII_Short (MARTHA / MARHTA)
     - BenchmarkJaroScore_ASCII_Medium (50-char identifier pair)
     - BenchmarkJaroScore_ASCII_Long (300-char pair — exceeds maxJaroStackLen=256, exercises heap path)
     - BenchmarkJaroScore_Unicode_Short (multi-byte rune pair)
   Each uses b.ReportAllocs() before b.ResetTimer() and var-sink pattern.

Create `jaro_fuzz_test.go`:

1. Apache-2.0 header + file-level godoc.
2. FuzzJaroScore — programmatic seeds: MARTHA/MARHTA, DIXON/DICKSONX, ""/"ABC", invalid-UTF-8 (`"\xff\xfe"`), and a length-mismatched pair. Body: no panic, !math.IsNaN, !math.IsInf, score in [0,1].
3. Create `testdata/fuzz/FuzzJaroScore/seed-001` in `go test fuzz v1` format with MARTHA/MARHTA.

Extend `props_test.go` (APPEND — read existing file's package + import block first; append new functions after the last existing `TestProp_*` declaration found via grep):

1. TestProp_JaroScore_RangeBounds
2. TestProp_JaroScore_Identity (skip empty in predicate)
3. TestProp_JaroScore_Symmetric
4. TestProp_JaroScore_NoNaN
5. TestProp_JaroScore_NoInf
6. TestProp_JaroScore_NoNegativeZero
7. NO triangle inequality property — add a comment explaining the omission per RESEARCH.md (Jaro is not a metric).

Extend `example_test.go` (APPEND — append after the last `Example*` function in the file):

       func ExampleJaroScore() {
           fmt.Printf("%.4f\n", fuzzymatch.JaroScore("MARTHA", "MARHTA"))
           // Output:
           // 0.9444
       }

Extend `algoid_test.go`:

1. Remove AlgoJaro from the unregistered-slots list. Add a registered-slot assertion `DispatchEntryNilForTest(int(fuzzymatch.AlgoJaro)) == false`.

Run:
  go test -race -shuffle=on -count=1 -run 'TestJaro|TestProp_Jaro|TestDispatch|ExampleJaroScore' ./...
  go test -bench=BenchmarkJaroScore_ASCII -benchmem -run=^$ -count=3 ./...
  go test -fuzz=FuzzJaroScore -fuzztime=30s -run=^$ ./...
  </action>
  <verify>
    <automated>go test -race -shuffle=on -count=1 -run 'TestJaro|TestProp_Jaro|TestDispatch|ExampleJaroScore' ./... && go test -bench=BenchmarkJaroScore_ASCII_Short -benchmem -run=^$ -count=3 ./... 2>&1 | grep -E '0 B/op[[:space:]]+0 allocs/op'</automated>
  </verify>
  <acceptance_criteria>
    - All TestJaro_* tests pass; reference vectors match within 1e-6.
    - All TestProp_Jaro* tests pass.
    - TestJaro_ZeroAllocs_ASCII_Short reports 0 allocations.
    - BenchmarkJaroScore_ASCII_Short reports `0 B/op  0 allocs/op` in `-benchmem` output.
    - ExampleJaroScore output matches `0.9444\n` byte-for-byte.
    - testdata/fuzz/FuzzJaroScore/seed-001 exists and parses as `go test fuzz v1` corpus.
    - algoid_test.go: AlgoJaro slot non-nil; remaining unregistered slots still pass nil assertion.
    - `grep -c '"github.com/stretchr/testify' jaro_test.go jaro_bench_test.go jaro_fuzz_test.go` returns 0.
    - props_test.go contains no TestProp_JaroDistance_TriangleInequality (Jaro has no Distance variant; triangle inequality must not appear).
  </acceptance_criteria>
  <behavior>
    - Reference vectors pin Jaro accuracy at 1e-6 tolerance.
    - Property tests cover RangeBounds, Identity, Symmetric, NoNaN, NoInf, NoNegativeZero.
    - Benchmark pins the 0-alloc target on ASCII Short (within maxJaroStackLen=256).
    - Fuzz harness panic-free + invariant-preserving on programmatic seeds plus 30s of random input.
  </behavior>
  <done>
    All test files committed; props_test, example_test, algoid_test extended; full Jaro-scoped test suite green.
  </done>
</task>

<task type="auto" tdd="true">
  <name>Task 3: Per-algorithm staging golden file + BDD feature + extend BDD steps</name>
  <files>testdata/golden/_staging/jaro.json, tests/bdd/features/jaro.feature, tests/bdd/steps/algorithms_steps.go, algorithms_golden_test.go</files>
  <read_first>
    - algorithms_golden_test.go (Wave 1 plan 02-01 Task 3 — goldenAlgorithmsFile + goldenAlgorithmEntry struct shape AND the LOCKED-signature helper `assertGoldenStaging(t *testing.T, relPath string, v any)`. Call the helper directly — there is NO "if helper exists / else create" branch in this plan. If the helper is missing, plan 02-01 has not landed yet and Wave 2 must wait per the wave dependency.)
    - testdata/golden/_staging/levenshtein.json (created by plan 02-01 — confirms the staging schema and gives a structural reference)
    - testdata/golden/_staging/hamming.json (if landed earlier — exact byte-form reference for staging files)
    - tests/bdd/features/levenshtein.feature (Wave 1 BDD pattern)
    - tests/bdd/features/hamming.feature (if landed earlier — sibling pattern)
    - tests/bdd/steps/algorithms_steps.go (current state — find AlgorithmContext + InitializeScenario; APPEND new step methods after the last `iComputeThe*` method and append regex registrations after the last `ctx.Step(...)` call in InitializeScenario)
    - .planning/phases/02-core-character-algorithms-six/02-PATTERNS.md (Pattern 12, Pattern 15)
    - .planning/phases/02-core-character-algorithms-six/02-RESEARCH.md §Golden File Integration; §BDD Scenario Coverage
  </read_first>
  <action>
Create `testdata/golden/_staging/jaro.json`:

1. Schema matches algorithms.json (`{"version": 1, "entries": [...]}`).
2. Entries (sorted by Name):
     - Jaro_DIXON_DICKSONX (a "DIXON", b "DICKSONX", expected_score from a live JaroScore call)
     - Jaro_empty_empty (a "", b "", 1.0)
     - Jaro_identical (a "ABC", b "ABC", 1.0)
     - Jaro_JELLYFISH_SMELLYFISH (a "JELLYFISH", b "SMELLYFISH", expected_score from live call)
     - Jaro_MARTHA_MARHTA (a "MARTHA", b "MARHTA", expected_score from live call)
     - Jaro_one_empty (a "", b "ABC", 0.0)
3. Generate via the LOCKED-signature helper `assertGoldenStaging` defined by plan 02-01 Task 3. Add `TestGolden_Jaro_Staging` to algorithms_golden_test.go (gated on the `-update` flag) that builds the entries list, sorts by Name, wraps in `goldenAlgorithmsFile{Version: 1, Entries: entries}`, and calls `assertGoldenStaging(t, "_staging/jaro.json", file)` — UNCONDITIONALLY. Do NOT add a fallback "if helper exists / else create" branch; plan 02-01 owns the helper definition. Run with `-update` once; commit. Re-run without `-update` and confirm zero diff.

Create `tests/bdd/features/jaro.feature`:

1. File-level comment with primary source (Jaro 1989).
2. Feature: Jaro similarity algorithm
3. Scenario Outline "canonical reference vectors" with rows for MARTHA/MARHTA → 0.9444, DIXON/DICKSONX → 0.7667, JELLYFISH/SMELLYFISH → 0.8963, ABC/ABC → 1.0000 (tolerance 0.0001).
4. Scenario "both-empty strings score 1.0" — Score("","") == 1.0 exactly.
5. Scenario "one-empty strings score 0.0" — Score("","ABC") == 0.0 exactly.
6. Scenario "Jaro is symmetric" — Score(MARTHA, MARHTA) == Score(MARHTA, MARTHA).
7. Scenario "Jaro is not a metric" — comment in feature file (Gherkin `#`) noting the omission of triangle-inequality scenarios. (BDD scenarios cannot easily test "non-property" claims; the comment serves as documentation alongside the godoc on jaro.go.)

Extend `tests/bdd/steps/algorithms_steps.go` (APPEND):

1. Add to AlgorithmContext (or as a sibling method): `iComputeTheJaroScoreBetween(a, b string) error` → calls fuzzymatch.JaroScore.
2. Register the regex in the existing InitializeScenario:
     - `^I compute the Jaro score between "([^"]*)" and "([^"]*)"$`

Run:
  go test -race -shuffle=on -count=1 -run 'TestGolden_Jaro_Staging|TestGolden_Algorithms' ./...
  (cd tests/bdd && go test -race -shuffle=on -count=1 ./...)
  </action>
  <verify>
    <automated>go test -race -shuffle=on -count=1 -run 'TestGolden_Jaro_Staging|TestGolden_Algorithms' ./... && (cd tests/bdd && go test -race -shuffle=on -count=1 ./...)</automated>
  </verify>
  <acceptance_criteria>
    - testdata/golden/_staging/jaro.json exists, contains exactly six Jaro entries sorted alphabetically by Name.
    - Canonical form: 2-space indent, trailing LF, no BOM (`xxd testdata/golden/_staging/jaro.json | tail -1` shows `0a` as last byte).
    - Re-running TestGolden_Jaro_Staging without -update produces no diff.
    - tests/bdd/features/jaro.feature parses as valid Gherkin with at least four scenarios (Scenario Outline + 3 standalone scenarios).
    - `cd tests/bdd && go test -count=1 ./...` exits 0.
    - tests/bdd/steps/algorithms_steps.go still has exactly one AlgorithmContext type and one InitializeScenario function.
    - testdata/golden/algorithms.json UNCHANGED relative to Wave 1.
  </acceptance_criteria>
  <behavior>
    - Staging file ready for Wave 3 plan 02-07 merge.
    - BDD harness exercises canonical Jaro reference vectors and edge cases.
    - AlgorithmContext extended without conflict with sibling Wave 2 plans.
  </behavior>
  <done>
    Staging golden file, Jaro BDD feature, and Jaro BDD step bindings committed. Jaro-scoped quality gate green.
  </done>
</task>

</tasks>

<threat_model>
## Trust Boundaries

| Boundary | Description |
|----------|-------------|
| Caller → fuzzymatch.JaroScore | Untrusted strings (any length, any byte sequence). Pure function; no other boundary. |

## STRIDE Threat Register

| Threat ID | Category | Component | Disposition | Mitigation Plan |
|-----------|----------|-----------|-------------|-----------------|
| T-02-03-01 | Denial of Service | JaroScore on adversarial inputs | accept | Jaro is O(la · w) where w is the matching window ≈ max(la,lb)/2. Effectively O(la · lb). Document worst-case complexity in godoc. The 0-alloc fast path is bounded at maxJaroStackLen=256; longer inputs use heap-allocated `[]bool` slices (linear memory, not quadratic). |
| T-02-03-02 | Information Disclosure | Malformed UTF-8 input causing panic | mitigate | Byte-level Jaro operates on bytes — invalid UTF-8 cannot panic. Rune variant uses `[]rune(s)` which Go normalises to U+FFFD. FuzzJaroScore in Task 2 includes invalid-UTF-8 seeds and asserts no panic, no NaN, no Inf. |
| T-02-03-03 | Tampering | dispatch[AlgoJaro] overwritten | mitigate | dispatch is unexported; registration runs once at package load. Same mitigation as T-02-01-04 (Wave 1) and T-02-02-03 (Hamming). |
| T-02-03-04 | Information Disclosure | Float-determinism violation across CI matrix platforms (linux/arm64 vs windows/amd64) | mitigate | Explicit parenthesisation of the three-term Jaro formula; left-to-right evaluation order; no `math.Pow`/`math.Log`/`math.Exp`/`math.Sqrt`/`math.FMA`. Compile to byte-identical IEEE-754 output across platforms — verified by the `_staging/jaro.json` file diff in Wave 3 + the cross-platform CI matrix `make verify-determinism` gate. |
| T-02-03-05 | Repudiation | "Jaro is a metric" misuse leading to invalid distance reasoning | mitigate | Locked godoc paragraph "Jaro is NOT a metric" appears in jaro.go file-level godoc and in jaro.feature Gherkin comments. props_test.go documents the omitted triangle-inequality property with a citation. |

No high-severity items. ASVS L1 V5 (Input Validation) addressed via fuzz tests. Plan passes the security gate.
</threat_model>

<verification>
1. `go build ./...` exits 0.
2. `go vet ./...` exits 0.
3. `bash scripts/verify-license-headers.sh` exits 0.
4. `bash scripts/verify-no-runtime-deps.sh` exits 0.
5. `go test -race -shuffle=on -count=1 -run 'TestJaro|TestProp_Jaro|TestDispatch|ExampleJaroScore|TestGolden_Jaro_Staging' ./...` exits 0.
6. `go test -bench=BenchmarkJaroScore_ASCII_Short -benchmem -run=^$ -count=3 ./...` reports `0 B/op  0 allocs/op`.
7. `go test -fuzz=FuzzJaroScore -fuzztime=30s -run=^$ ./...` completes without crash.
8. `(cd tests/bdd && go test -race -shuffle=on -count=1 ./...)` exits 0.
9. testdata/golden/_staging/jaro.json byte-stable.
10. testdata/golden/algorithms.json UNCHANGED relative to Wave 1's commit.
11. `make check` exits 0.
</verification>

<success_criteria>
- Jaro reference vectors (MARTHA/MARHTA, DIXON/DICKSONX, JELLYFISH/SMELLYFISH) pinned within 1e-6 of Jaro-1989 / Winkler-1990 expected values.
- Zero allocations on ASCII ≤ 256 chars (PERF-01, PERF-02 satisfied for Jaro).
- No NaN / Inf / -0 on any path (DET-04 satisfied for Jaro).
- "Jaro is not a metric" lock documented in godoc, props_test.go (omission rationale), and jaro.feature comments.
- Per-algorithm staging golden file ready for Wave 3 merge.
- All required gates green via `make check`.
- Plan 02-04 (Jaro-Winkler) can build on JaroScore directly without re-deriving Jaro.
</success_criteria>

<output>
After completion, create `.planning/phases/02-core-character-algorithms-six/02-03-jaro-SUMMARY.md` recording:
- Final identifier names confirmed (JaroScore, JaroScoreRunes — zero drift; no Distance variant per the formula's [0,1] direct output).
- Benchmark numbers observed (0 allocs/op on ASCII Short/Medium; heap path on Long; rune path 2 allocs/op).
- The exact `maxJaroStackLen` value shipped (locked at 256 by this plan).
- Reference vector exact float64 values committed to _staging/jaro.json (for traceability when JW plan 02-04 builds on these).
- Coverage percentages.
- Any deviations from the plan and their rationale.
</output>
