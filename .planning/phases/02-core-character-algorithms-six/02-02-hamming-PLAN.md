---
phase: 02-core-character-algorithms-six
plan: 02
type: execute
wave: 2
depends_on: [02-01-levenshtein]
files_modified:
  - hamming.go
  - dispatch_hamming.go
  - hamming_test.go
  - hamming_bench_test.go
  - hamming_fuzz_test.go
  - props_test.go
  - example_test.go
  - algoid_test.go
  - testdata/golden/_staging/hamming.json
  - testdata/fuzz/FuzzHammingScore/seed-001
  - tests/bdd/features/hamming.feature
  - tests/bdd/steps/algorithms_steps.go
autonomous: true
requirements:
  - CHAR-04
  - PERF-01
  - PERF-02
  - TEST-01
  - TEST-02
  - TEST-04
  - TEST-05
  - DET-04
  - DX-02
tags: [hamming, equal-length, silent-zero-unequal, no-dp]

must_haves:
  truths:
    - "HammingDistance(\"karolin\", \"kathrin\") returns 3 (Hamming 1950 reference vector)"
    - "HammingScore(\"karolin\", \"kathrin\") ≈ 0.5714 (within 1e-9 of 1 - 3/7)"
    - "HammingDistance(\"abc\", \"ab\") returns max(3, 2) = 3 (silent-zero policy per CONTEXT.md decision)"
    - "HammingScore(\"abc\", \"ab\") returns exactly 0.0 silently — NO error, NO panic"
    - "HammingDistance(\"\", \"\") returns 0; HammingScore(\"\", \"\") returns exactly 1.0"
    - "HammingDistanceRunes counts rune mismatches (not byte mismatches) — \"café\"/\"cafè\" returns 1, not 2"
    - "dispatch[AlgoHamming] is non-nil and equals HammingScore after package init"
    - "Apache-2.0 header on every new .go file (verified by scripts/verify-license-headers.sh)"
    - "hamming.go contains `// Source: Hamming, R. W. (1950).` block at top of file-level godoc"
    - "hamming.go file-level godoc EXPLICITLY states the unequal-length policy: 'Inputs of unequal length are not an error: HammingDistance returns max(len(a), len(b)) and HammingScore returns 0.0. Callers wanting strict Hamming-1950 equal-length semantics should length-check before calling.'"
    - "No math.Pow / math.Log / math.Exp / math.Sqrt / math.FMA in hamming.go"
    - "No init() function; dispatch_hamming.go uses var _ = func() bool { ... }() registration idiom"
    - "Property tests pass for RangeBounds, Identity, Symmetric, NoNaN, NoInf, NoNegativeZero (NO triangle inequality property — see Hamming triangle-inequality caveat in RESEARCH.md §Mathematical Invariants)"
    - "Benchmark BenchmarkHammingScore_ASCII_Short reports 0 B/op, 0 allocs/op (Hamming is a single loop — trivially zero-alloc)"
    - "Fuzz test FuzzHammingScore exists with at least one programmatic seed plus invalid-UTF-8 seed; panic-free + score in [0,1]"
    - "BDD scenarios in tests/bdd/features/hamming.feature exercise reference vectors AND the unequal-length silent-zero behaviour explicitly"
    - "testdata/golden/_staging/hamming.json contains Hamming entries (Hamming_karolin_kathrin, Hamming_identical, Hamming_unequal_length, Hamming_empty_empty) sorted by Name; final merge into algorithms.json happens in plan 02-07"
    - "algoid_test.go updated to assert dispatch[AlgoHamming] is non-nil and slot is removed from the unregistered-list"
    - "ExampleHammingScore runs with `// Output:` block matching byte-for-byte (must demonstrate BOTH the equal-length case AND the unequal-length silent-zero case)"
  artifacts:
    - path: "hamming.go"
      provides: "HammingDistance, HammingDistanceRunes, HammingScore, HammingScoreRunes"
      min_lines: 80
      contains: "// Source: Hamming"
    - path: "dispatch_hamming.go"
      provides: "Package-load-time registration of HammingScore into dispatch[AlgoHamming]"
      contains: "dispatch[AlgoHamming] = HammingScore"
    - path: "testdata/golden/_staging/hamming.json"
      provides: "Per-algorithm staging golden file for the Wave 3 merge into algorithms.json"
      contains: "Hamming_karolin_kathrin"
  key_links:
    - from: "dispatch_hamming.go"
      to: "algoid.go (dispatch array)"
      via: "package-level var _ = func()bool{ dispatch[AlgoHamming] = HammingScore; return true }()"
      pattern: "dispatch\\[AlgoHamming\\]"
    - from: "hamming.go HammingScore"
      to: "hamming.go HammingDistance"
      via: "function call inside score normaliser"
      pattern: "HammingDistance\\("

user_setup: []
---

<objective>
Implement Hamming distance/score with the LOCKED silent-zero unequal-length policy from CONTEXT.md (recorded in commit `1e25e31`). Hamming is the simplest of the six algorithms — a single counting loop, no DP — but it is the load-bearing test of the canonical pattern: every Wave 2 plan extends the same files (`props_test.go`, `example_test.go`, `algoid_test.go`, `tests/bdd/steps/algorithms_steps.go`) and creates a per-algorithm staging golden file (`testdata/golden/_staging/hamming.json`) instead of editing the shared `algorithms.json` directly.

Purpose: ship Hamming with zero merge collisions against the four other Wave 2 plans (02-03 Jaro, 02-04 JW, 02-05 DL-OSA, 02-06 DL-Full), and pin the unequal-length silent-zero behaviour as the canonical, project-wide convention via godoc, unit tests, BDD, and the example.

Output: a working `HammingScore` that returns 0 allocations on any-length ASCII inputs and silently returns 0.0 (not an error) on length-mismatched inputs, plus the staging golden file `testdata/golden/_staging/hamming.json` that plan 02-07 merges into the canonical `testdata/golden/algorithms.json`.
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
Wave 1 (plan 02-01) locked these append-points the executor MUST extend (NOT recreate):

From levenshtein.go (already in package fuzzymatch — DO NOT redeclare):
  const maxStackInputLen = 64   // shared stack-buffer threshold

From normalise.go:
  func isASCII(s string) bool   // package-level helper, reusable

From algoid.go:
  AlgoHamming AlgoID = 3
  var dispatch [numAlgorithms]func(a, b string) float64

From algorithms_golden_test.go (Wave 1 plan 02-01 Task 3 — defines the canonical staging-write helper for the entire phase):
  func assertGoldenStaging(t *testing.T, relPath string, v any)
  // Signature LOCKED. Call directly — no "if helper exists / else create" branch.
  // Writes to testdata/golden/<relPath> (e.g. "_staging/hamming.json") via WriteGoldenFile.

From dispatch_levenshtein.go (template to copy character-for-character with identifiers swapped):
  var _ = func() bool {
      dispatch[AlgoLevenshtein] = LevenshteinScore
      return true
  }()

From props_test.go (Wave 1 file — APPEND your test functions; DO NOT recreate the file or its imports):
  package fuzzymatch_test
  import (..., "math", "testing", "testing/quick", "github.com/axonops/fuzzymatch")
  // Add: TestProp_HammingScore_RangeBounds, _Identity, _Symmetric, _NoNaN, _NoInf, _NoNegativeZero
  // SKIP: TestProp_HammingDistance_TriangleInequality — see RESEARCH.md (Hamming triangle inequality requires equal-length constraint; Wave 2 owner may add an equal-length-constrained variant if simple, otherwise omit and document why)

From example_test.go (Wave 1 file — APPEND ExampleHammingScore):
  // ExampleHammingScore must demonstrate BOTH equal-length and unequal-length cases per the locked silent-zero policy

From algoid_test.go (Wave 1 file — UPDATE the dispatch test):
  // Wave 1 renamed/split TestDispatch_AllNilAtPhase1; this plan removes AlgoHamming from the unregistered-slots list

From tests/bdd/steps/algorithms_steps.go (Wave 1 file — APPEND step bindings to AlgorithmContext + InitializeScenario):
  // Add: iComputeTheHammingScoreBetween(a, b string) error
  //      iComputeTheHammingDistanceBetween(a, b string) error
  //      theDistanceShouldBe(expected int) error  (if not already added by Wave 1)

From testdata/golden/algorithms.json (Wave 1 file — DO NOT EDIT):
  // Plan 02-07 (Wave 3) merges _staging/<algo>.json files into algorithms.json
  // This Wave 2 plan creates testdata/golden/_staging/hamming.json ONLY
</interfaces>
</context>

<tasks>

<task type="auto" tdd="true">
  <name>Task 1: Implement hamming.go (single-loop algorithm) and dispatch_hamming.go (registration)</name>
  <files>hamming.go, dispatch_hamming.go</files>
  <read_first>
    - levenshtein.go (the canonical pattern from Wave 1 — Apache-2.0 header, file-level godoc with // Source block, public-API godoc convention, score-normalisation guard, file structure)
    - dispatch_levenshtein.go (the EXACT registration idiom — copy character-for-character, swap identifiers)
    - normalise.go (isASCII helper at lines 159-168 — call directly, do NOT redeclare)
    - .planning/phases/02-core-character-algorithms-six/02-CONTEXT.md (Hamming unequal-length LOCKED decision: silent zero, with the EXACT godoc text the file must include)
    - .planning/phases/02-core-character-algorithms-six/02-RESEARCH.md §Primary Sources — Hamming; §Score Normalisation — Hamming row
    - .planning/phases/02-core-character-algorithms-six/02-PATTERNS.md (Pattern 5 score normalisation — Hamming silent-zero on unequal length)
    - docs/requirements.md §7.1.4 (Hamming spec — note CONTEXT.md overrides the spec's illustrative `(int, error)` signature)
    - .claude/skills/algorithm-correctness-standards/SKILL.md (primary-source citation discipline)
  </read_first>
  <action>
Create `hamming.go` in package fuzzymatch:

1. Apache-2.0 header copied from normalise.go lines 1-13.
2. File-level godoc opening with `// Source: Hamming, R. W. (1950). "Error detecting and error correcting codes." Bell System Technical Journal, 29(2):147-160.` Include the algorithm description (count positions where a[i] != b[i]) and the LOCKED unequal-length godoc paragraph copied verbatim from CONTEXT.md "Hamming unequal-length behaviour":

       // Inputs of unequal length are not an error: HammingDistance returns
       // max(len(a), len(b)) and HammingScore returns 0.0. Callers wanting
       // strict Hamming-1950 equal-length semantics should length-check
       // before calling.

   This text is load-bearing — algorithm-correctness-reviewer will reject any deviation.
3. Public API:
     - HammingDistance(a, b string) int
     - HammingDistanceRunes(a, b string) int
     - HammingScore(a, b string) float64
     - HammingScoreRunes(a, b string) float64
4. Implementation:
     - HammingDistance: identity fast path `if a == b { return 0 }`; both-empty handled by the identity path; if `len(a) != len(b)` return `max(len(a), len(b))` per the locked policy; otherwise loop `i := 0; i < len(a); i++` counting `a[i] != b[i]`. Direct byte indexing — NO `[]byte(a)`.
     - HammingScore: `maxLen := max(len(a), len(b))`; guard `if maxLen == 0 { return 1.0 }`; return `1.0 - float64(HammingDistance(a, b))/float64(maxLen)`. The unequal-length case naturally yields `1 - maxLen/maxLen = 0.0` per the locked normalisation.
     - HammingDistanceRunes: eager `ra := []rune(a); rb := []rune(b)`; `if len(ra) != len(rb) { return max(len(ra), len(rb)) }`; loop counting rune mismatches.
     - HammingScoreRunes: same normalisation pattern using rune lengths.
5. Public function godoc: each starts with the function name; document the silent-zero unequal-length policy on every public function (link via godoc cross-reference to the file-level discussion).

Create `dispatch_hamming.go` mirroring dispatch_levenshtein.go exactly:

       // [Apache-2.0 header]
       // dispatch_hamming.go registers HammingScore into the dispatch table at
       // package load time. Sole writer to dispatch[AlgoHamming].
       package fuzzymatch
       var _ = func() bool {
           dispatch[AlgoHamming] = HammingScore
           return true
       }()

After writing, run `go build ./...`, `go vet ./...`, `bash scripts/verify-license-headers.sh`.
  </action>
  <verify>
    <automated>go build ./... && go vet ./... && bash scripts/verify-license-headers.sh</automated>
  </verify>
  <acceptance_criteria>
    - hamming.go starts with the Apache-2.0 header (`diff <(head -13 normalise.go) <(head -13 hamming.go)` exit 0).
    - hamming.go contains `// Source: Hamming, R. W. (1950).` literal in file-level godoc.
    - hamming.go contains the EXACT unequal-length policy paragraph from CONTEXT.md (verifiable by `grep -F 'Inputs of unequal length are not an error' hamming.go`).
    - `grep -E 'math\.(Pow|Log|Exp|Sqrt|FMA)' hamming.go` returns no matches.
    - `grep -E '^func init\(' hamming.go dispatch_hamming.go` returns no matches.
    - `grep -c '\[\]byte\(' hamming.go` returns 0.
    - dispatch_hamming.go contains `dispatch[AlgoHamming] = HammingScore` exactly once and uses the `var _ = func() bool {` idiom.
    - hamming.go declares NO `const maxStackInputLen` (it is owned by levenshtein.go; redeclaration would break the build).
    - `go build ./...` exits 0; `go vet ./...` exits 0.
  </acceptance_criteria>
  <behavior>
    - HammingDistance("karolin", "kathrin") == 3
    - HammingDistance("abc", "abc") == 0
    - HammingDistance("", "") == 0
    - HammingDistance("abc", "ab") == 3 (max(len), per silent-zero policy)
    - HammingDistance("ab", "abc") == 3 (silent-zero policy is symmetric)
    - HammingScore("karolin", "kathrin") within 1e-9 of 0.5714285714…
    - HammingScore("", "") == 1.0
    - HammingScore("abc", "ab") == 0.0 exactly
    - HammingDistanceRunes("café", "cafè") == 1 (rune-aware: only the last rune differs)
    - dispatch[AlgoHamming] non-nil after package load and equals HammingScore
  </behavior>
  <done>
    hamming.go and dispatch_hamming.go committed. Package builds. License-header verifier passes. Locked silent-zero policy documented verbatim in godoc.
  </done>
</task>

<task type="auto" tdd="true">
  <name>Task 2: Tests (unit + property + benchmark + fuzz) + extend props_test.go, example_test.go, algoid_test.go</name>
  <files>hamming_test.go, hamming_bench_test.go, hamming_fuzz_test.go, props_test.go, example_test.go, algoid_test.go, testdata/fuzz/FuzzHammingScore/seed-001</files>
  <read_first>
    - levenshtein_test.go (Wave 1 unit-test pattern — copy structure, swap identifiers)
    - levenshtein_bench_test.go (Wave 1 benchmark pattern — b.ReportAllocs/ResetTimer order, var sink, Short/Medium/Long/Unicode breakdown)
    - levenshtein_fuzz_test.go (Wave 1 fuzz pattern — programmatic seeds + invalid-UTF-8)
    - props_test.go (Wave 1 file — extend, do NOT recreate; understand the imports + package declaration before editing)
    - example_test.go (Wave 1 file — append ExampleHammingScore)
    - algoid_test.go (Wave 1's updated dispatch test — find the unregistered-slots list and remove AlgoHamming from it)
    - testdata/fuzz/FuzzLevenshteinScore/seed-001 (corpus file format reference)
    - .planning/phases/02-core-character-algorithms-six/02-PATTERNS.md (Pattern 7 unit; Pattern 9 bench; Pattern 10 fuzz; Pattern 11 props; Pattern 14 example; Pattern 13 dispatch test update)
    - .planning/phases/02-core-character-algorithms-six/02-RESEARCH.md §Mathematical Invariants (Hamming triangle inequality caveat); §Allocation Budgets — Hamming row
  </read_first>
  <action>
Create `hamming_test.go` (package fuzzymatch_test, stdlib testing only):

1. Apache-2.0 header + file-level godoc.
2. Tests (table-driven where applicable):
     - TestHamming_BothEmpty — Distance and Score on ("", "") → 0 and 1.0
     - TestHamming_Identical — Distance 0, Score 1.0 on ("abc", "abc")
     - TestHamming_ReferenceVectors — table over the four canonical vectors (karolin/kathrin → 3 / 0.5714…; 1011101/1001001 → 2 / 0.7142…; abc/abc → 0 / 1.0; ""/"" → 0 / 1.0). Float tolerance 1e-9 via math.Abs.
     - TestHamming_UnequalLength_SilentZero — explicitly tests the locked policy: HammingDistance("abc","ab")==3, HammingScore("abc","ab")==0.0, HammingDistance("ab","abc")==3 (symmetry of the policy), HammingScore("ab","abc")==0.0.
     - TestHamming_Symmetry — Score(a,b) == Score(b,a) for the reference vectors AND the unequal-length pair.
     - TestHamming_DistanceRunes_MultiByte — café/cafè → 1 (rune-aware).
     - TestHammingScore_ZeroAllocs — `testing.AllocsPerRun(100, func() { _ = fuzzymatch.HammingScore("karolin", "kathrin") })` must be 0.

Create `hamming_bench_test.go`:

1. Apache-2.0 header + file-level godoc citing the PERF-01 budget (0 allocs on ASCII, any length).
2. Benchmarks: BenchmarkHammingScore_ASCII_Short (karolin/kathrin), _ASCII_Medium (50-char pair), _ASCII_Long (500-char pair), _Unicode_Short (multi-byte rune pair). Every benchmark uses b.ReportAllocs() before b.ResetTimer() and the var-sink-prevent-DCE pattern.

Create `hamming_fuzz_test.go`:

1. Apache-2.0 header + file-level godoc.
2. FuzzHammingScore — seed with reference vectors plus invalid-UTF-8 plus a length-mismatched pair to exercise the silent-zero path. Body asserts no panic, !math.IsNaN, !math.IsInf, score in [0,1].
3. Create `testdata/fuzz/FuzzHammingScore/seed-001` in the `go test fuzz v1` format with the karolin/kathrin pair.

Extend `props_test.go` (DO NOT recreate; read the existing file first to confirm the import block and package declaration; APPEND new functions after the last existing `TestProp_*` declaration found via grep):

1. Append (do not duplicate imports):
     - TestProp_HammingScore_RangeBounds
     - TestProp_HammingScore_Identity (skip empty inputs in the predicate, like the Levenshtein equivalent)
     - TestProp_HammingScore_Symmetric
     - TestProp_HammingScore_NoNaN
     - TestProp_HammingScore_NoInf
     - TestProp_HammingScore_NoNegativeZero
2. Triangle inequality: SKIP. The general triangle inequality fails for Hamming under the silent-zero policy on unequal-length inputs (RESEARCH.md §Mathematical Invariants — Hamming triangle inequality must be constrained to equal-length inputs). If implementing an equal-length-constrained variant is simple, add `TestProp_HammingDistance_TriangleInequality_EqualLength` that generates a random base string and two same-length perturbations (e.g. via XOR-flip of bytes); otherwise add a comment in props_test.go explaining the omission and citing RESEARCH.md.

Extend `example_test.go` (append, do not recreate; append after the last `Example*` function in the file):

1. ExampleHammingScore — must demonstrate BOTH the equal-length case AND the unequal-length silent-zero case per the locked policy:
       fmt.Printf("%.4f\n", fuzzymatch.HammingScore("karolin", "kathrin"))
       fmt.Printf("%.4f\n", fuzzymatch.HammingScore("abc", "ab"))
       // Output:
       // 0.5714
       // 0.0000

Extend `algoid_test.go`:

1. Open the file and find the Wave-1-updated dispatch test (named `TestDispatch_LevenshteinRegistered_OthersNil` or split form).
2. Remove `AlgoHamming` from the unregistered-slots list. The simplest pattern: maintain an `unregisteredSlots := []fuzzymatch.AlgoID{...}` slice and assert all listed slots are nil; this plan removes AlgoHamming from that slice. The corresponding "registered" assertion (`DispatchEntryNilForTest(int(fuzzymatch.AlgoHamming)) == false`) can be added as a sibling check OR rolled into a generic `for _, registered := range []AlgoID{AlgoLevenshtein, AlgoHamming} { ... }` loop — pick whichever the Wave 1 SUMMARY documented as the canonical structure.

Run:
  go test -race -shuffle=on -count=1 -run 'TestHamming|TestProp_Hamming|TestDispatch|ExampleHammingScore' ./...
  go test -bench=BenchmarkHammingScore_ASCII -benchmem -run=^$ -count=3 ./...
  go test -fuzz=FuzzHammingScore -fuzztime=30s -run=^$ ./...
  </action>
  <verify>
    <automated>go test -race -shuffle=on -count=1 -run 'TestHamming|TestProp_Hamming|TestDispatch|ExampleHammingScore' ./... && go test -bench=BenchmarkHammingScore_ASCII_Short -benchmem -run=^$ -count=3 ./... 2>&1 | grep -E '0 B/op[[:space:]]+0 allocs/op'</automated>
  </verify>
  <acceptance_criteria>
    - All TestHamming_* tests pass (including TestHamming_UnequalLength_SilentZero).
    - All TestProp_Hamming* tests pass (excluding the omitted general triangle inequality).
    - TestHammingScore_ZeroAllocs reports 0 allocations.
    - BenchmarkHammingScore_ASCII_Short reports `0 B/op  0 allocs/op` in `-benchmem` output.
    - ExampleHammingScore output matches the two-line `// Output: 0.5714\n0.0000\n` block byte-for-byte.
    - testdata/fuzz/FuzzHammingScore/seed-001 exists and parses as `go test fuzz v1` corpus.
    - algoid_test.go assertions: `DispatchEntryNilForTest(int(AlgoHamming)) == false` AND nil-slot-check still passes for the four remaining unregistered slots (DL-OSA, DL-Full, Jaro, JW are nil at this point in Wave 2 — order of merges does not matter because the unregistered-list shrinks monotonically as Wave 2 plans land).
    - `grep -c '"github.com/stretchr/testify' hamming_test.go hamming_bench_test.go hamming_fuzz_test.go` returns 0.
  </acceptance_criteria>
  <behavior>
    - Unit tests cover identity, both-empty, reference vectors, unequal-length silent zero, symmetry, multi-byte runes, zero-alloc ASCII.
    - Property tests cover RangeBounds, Identity, Symmetric, NoNaN, NoInf, NoNegativeZero (and optionally TriangleInequality_EqualLength).
    - Benchmarks pin the 0-alloc target on ASCII (any length — Hamming has no DP buffer so the budget is trivially zero).
    - Fuzz harness panic-free on programmatic seeds plus 30s of random input including length-mismatched pairs.
    - Example demonstrates BOTH equal-length and silent-zero unequal-length output.
  </behavior>
  <done>
    All test files committed; props_test, example_test, algoid_test extended without breaking Wave 1 tests; full hamming-scoped test suite green.
  </done>
</task>

<task type="auto" tdd="true">
  <name>Task 3: Per-algorithm staging golden file + BDD feature + extend BDD steps</name>
  <files>testdata/golden/_staging/hamming.json, tests/bdd/features/hamming.feature, tests/bdd/steps/algorithms_steps.go</files>
  <read_first>
    - algorithms_golden_test.go (Wave 1 plan 02-01 Task 3 — note the goldenAlgorithmsFile + goldenAlgorithmEntry structs AND the LOCKED-signature helper `assertGoldenStaging(t *testing.T, relPath string, v any)`. Call the helper directly — there is NO "if helper exists / else create" branch in this plan. If the helper is missing, plan 02-01 has not landed yet and Wave 2 must wait per the wave dependency.)
    - testdata/golden/algorithms.json (Wave 1 — reference for the canonical byte form: 2-space indent, trailing LF, sorted entries)
    - testdata/golden/_staging/levenshtein.json (created by plan 02-01 Task 3 — confirms the staging schema and gives a structural reference)
    - golden_canonical.go (CanonicalMarshal contract)
    - tests/bdd/features/levenshtein.feature (Wave 1 BDD pattern — copy structure, swap identifiers and reference vectors)
    - tests/bdd/steps/algorithms_steps.go (Wave 1 — find AlgorithmContext + InitializeScenario; understand the existing step regex set; APPEND new step methods after the last `iComputeThe*` method and append regex registrations after the last `ctx.Step(...)` call in InitializeScenario)
    - .planning/phases/02-core-character-algorithms-six/02-PATTERNS.md (Pattern 12 golden file; Pattern 15 BDD feature + steps)
    - .planning/phases/02-core-character-algorithms-six/02-RESEARCH.md §Golden File Integration — Wave 2 staging strategy; §BDD Scenario Coverage
  </read_first>
  <action>
Create `testdata/golden/_staging/hamming.json` (the staging file Wave 3 plan 02-07 merges into algorithms.json):

1. The file has the SAME schema as algorithms.json — `{"version": 1, "entries": [...]}` with `goldenAlgorithmsFile`-shaped contents.
2. Entries (sorted by Name alphabetically):
     - Hamming_empty_empty (a "", b "", expected_score 1.0)
     - Hamming_identical (a "abc", b "abc", expected_score 1.0)
     - Hamming_karolin_kathrin (a "karolin", b "kathrin", expected_score from a live HammingScore call — 0.5714285714285714)
     - Hamming_unequal_length (a "abc", b "ab", expected_score 0.0)
3. Generate the file by adding `TestGolden_Hamming_Staging` to algorithms_golden_test.go (gated on the `-update` flag). The test builds the entries from live HammingScore calls, sorts them by Name, wraps in `goldenAlgorithmsFile{Version: 1, Entries: entries}`, and calls the LOCKED-signature helper `assertGoldenStaging(t, "_staging/hamming.json", file)` — UNCONDITIONALLY. Do NOT add a fallback "if helper exists / else create" branch; plan 02-01 Task 3 owns the helper definition. Concrete shape:

       func TestGolden_Hamming_Staging(t *testing.T) {
           entries := buildHammingStagingEntries(t)  // builds 4 entries from live calls
           sort.Slice(entries, func(i, j int) bool { return entries[i].Name < entries[j].Name })
           file := goldenAlgorithmsFile{Version: 1, Entries: entries}
           assertGoldenStaging(t, "_staging/hamming.json", file)
       }

4. Run with `-update` once to generate `testdata/golden/_staging/hamming.json`; commit the file. Re-running the test (without `-update`) must produce no diff. The canonical byte form (2-space indent, trailing LF, no BOM) is guaranteed by the helper's use of CanonicalMarshalForTest + WriteGoldenFile.

Create `tests/bdd/features/hamming.feature`:

1. File comment with primary source (Hamming 1950).
2. Feature: Hamming similarity algorithm
3. Scenario Outline "canonical reference vectors" with rows for karolin/kathrin → 0.5714, identical → 1.0, both-empty → 1.0 (tolerance 0.0001).
4. Scenario "unequal length returns silent zero" — explicitly: When I compute the Hamming score between "abc" and "ab" / Then the score should be exactly 0.0
5. Scenario "unequal length distance equals max length" — When I compute the Hamming distance between "abc" and "ab" / Then the distance should be 3
6. Scenario "score is symmetric" — verify Score(a,b) == Score(b,a) for both equal-length and unequal-length pairs.

Extend `tests/bdd/steps/algorithms_steps.go` (do NOT recreate the file or AlgorithmContext type; APPEND to it):

1. Add step methods to AlgorithmContext (or a Hamming-specific extension if AlgorithmContext is generic enough):
     - iComputeTheHammingScoreBetween(a, b string) error → calls fuzzymatch.HammingScore
     - iComputeTheHammingDistanceBetween(a, b string) error → stores distance in a new int field (e.g. `lastDistance int`); add the field to AlgorithmContext if Wave 1 did not already.
     - theDistanceShouldBe(expected int) error — compares lastDistance == expected (add this if Wave 1 did not — check via grep first).
2. Register the new step regexes inside the existing InitializeScenario (do NOT define a second InitializeScenario):
     - `^I compute the Hamming score between "([^"]*)" and "([^"]*)"$`
     - `^I compute the Hamming distance between "([^"]*)" and "([^"]*)"$`
     - `^the distance should be (\d+)$` (only if not already added by Wave 1)

Run:
  go test -race -shuffle=on -count=1 -run 'TestGolden' ./...
  (cd tests/bdd && go test -race -shuffle=on -count=1 ./...)
  make verify-determinism (informational: full algorithms.json determinism gate will be enforced after plan 02-07 merges; this Wave 2 plan only ensures _staging/hamming.json is byte-stable)
  </action>
  <verify>
    <automated>go test -race -shuffle=on -count=1 -run 'TestGolden_Hamming_Staging|TestGolden_Algorithms' ./... && (cd tests/bdd && go test -race -shuffle=on -count=1 ./...)</automated>
  </verify>
  <acceptance_criteria>
    - testdata/golden/_staging/hamming.json exists, contains exactly the four Hamming entries sorted alphabetically by Name.
    - The file is in canonical form: 2-space indent, trailing LF, no BOM (verifiable by `xxd testdata/golden/_staging/hamming.json | tail -1` showing `0a` as the last byte).
    - Re-running the staging test exits 0 with no diff (the staging file is byte-stable across runs).
    - tests/bdd/features/hamming.feature parses as valid Gherkin with at least four scenarios, including the explicit unequal-length-silent-zero scenario.
    - `cd tests/bdd && go test -count=1 ./...` exits 0; godog suite includes the Hamming feature; goleak detects no leaks.
    - tests/bdd/steps/algorithms_steps.go has exactly one AlgorithmContext type and one InitializeScenario function (Wave 2 plans extend, never duplicate).
    - testdata/golden/algorithms.json is UNCHANGED (Wave 2 must not touch it; plan 02-07 owns the merge).
  </acceptance_criteria>
  <behavior>
    - Staging file ready for plan 02-07's merge step.
    - BDD harness exercises canonical Hamming reference vectors AND the unequal-length silent-zero policy AND the distance-equals-max-length contract.
    - AlgorithmContext incrementally accumulates Wave 2 step bindings without conflict.
  </behavior>
  <done>
    Staging golden file, Hamming BDD feature, and Hamming BDD step bindings committed. Hamming-scoped quality gate green. Algorithms.json untouched (plan 02-07 will merge).
  </done>
</task>

</tasks>

<threat_model>
## Trust Boundaries

| Boundary | Description |
|----------|-------------|
| Caller → fuzzymatch.HammingScore | Untrusted strings (any length, any byte sequence). Pure-function library; no other boundary. |

## STRIDE Threat Register

| Threat ID | Category | Component | Disposition | Mitigation Plan |
|-----------|----------|-----------|-------------|-----------------|
| T-02-02-01 | Denial of Service | HammingDistance/Score on adversarial inputs | accept | Hamming is a single linear loop O(min(m,n)). No quadratic blowup, no stack growth. The only "DoS" surface is calling with very large strings, which is bounded by the caller's memory budget. PERF-01 (0 allocs on ASCII) ensures no GC pressure even at scale. |
| T-02-02-02 | Information Disclosure | Malformed UTF-8 input causing panic / leaking internal state | mitigate | Byte-level Hamming operates on bytes (no UTF-8 decoding) — invalid UTF-8 cannot panic. Rune variant uses `[]rune(s)` which Go normalises to U+FFFD on malformed bytes. FuzzHammingScore in Task 2 includes invalid-UTF-8 seeds and asserts no panic, no NaN, no Inf. |
| T-02-02-03 | Tampering | dispatch[AlgoHamming] overwritten | mitigate | dispatch is unexported; the registration runs once at package load via `var _ = func()bool{ ... }()`. T-01-05-05 mitigation from plan 01-05 still applies; same as T-02-01-04 in Wave 1. |
| T-02-02-04 | Repudiation | Unequal-length silent-zero policy interpreted as a bug by downstream callers | mitigate | The locked godoc paragraph (CONTEXT.md, copied verbatim into hamming.go) is explicit and discoverable. Tests pin the behaviour. ExampleHammingScore demonstrates it on pkg.go.dev. The only way to mistake the behaviour is to never read the documentation — at which point the test suite's TestHamming_UnequalLength_SilentZero acts as a contract for any future maintainer. |
| T-02-02-05 | Tampering | Staging golden file edited to weaken the score gate | mitigate | The staging file is committed to git; PR review enforces correctness; `make verify-determinism` will gate the final algorithms.json after plan 02-07 merges. The Wave 2 staging strategy isolates per-algorithm risk — a corrupted hamming.json affects only Hamming, not the other five algorithms. |

No high-severity items. ASVS L1 V5 (Input Validation) addressed via fuzz tests covering invalid-UTF-8 and length-mismatched inputs. Plan passes the security gate.
</threat_model>

<verification>
1. `go build ./...` exits 0.
2. `go vet ./...` exits 0.
3. `bash scripts/verify-license-headers.sh` exits 0.
4. `bash scripts/verify-no-runtime-deps.sh` exits 0.
5. `go test -race -shuffle=on -count=1 -run 'TestHamming|TestProp_Hamming|TestDispatch|ExampleHammingScore|TestGolden_Hamming_Staging' ./...` exits 0.
6. `go test -bench=BenchmarkHammingScore_ASCII_Short -benchmem -run=^$ -count=3 ./...` reports `0 B/op  0 allocs/op`.
7. `go test -fuzz=FuzzHammingScore -fuzztime=30s -run=^$ ./...` completes without crash or invariant violation.
8. `(cd tests/bdd && go test -race -shuffle=on -count=1 ./...)` exits 0.
9. testdata/golden/_staging/hamming.json is byte-stable: re-running the staging test produces no diff.
10. testdata/golden/algorithms.json is UNCHANGED relative to Wave 1's commit (Wave 2 plans must not touch it; verifiable via `git diff Wave1Commit..HEAD -- testdata/golden/algorithms.json` returning empty).
11. `make check` exits 0.
</verification>

<success_criteria>
- Hamming ships with the LOCKED unequal-length silent-zero policy (CONTEXT.md decision honoured in code, godoc, tests, BDD, example).
- Zero allocations on ASCII inputs at any length (PERF-01, PERF-02 satisfied for Hamming).
- No NaN / Inf / -0 on any path (DET-04 satisfied for Hamming).
- Per-algorithm staging golden file `testdata/golden/_staging/hamming.json` ready for plan 02-07's merge.
- BDD scenarios exercise reference vectors, unequal-length silent zero, and the distance-equals-max-length contract.
- Wave 2 collision avoidance verified: this plan touches neither algoid.go nor testdata/golden/algorithms.json; only the per-algorithm files plus the four shared append-points (props_test.go, example_test.go, algoid_test.go, tests/bdd/steps/algorithms_steps.go).
- All required gates (license headers, no-runtime-deps, lint, vet, race, tidy, coverage, BDD) pass via `make check`.
</success_criteria>

<output>
After completion, create `.planning/phases/02-core-character-algorithms-six/02-02-hamming-SUMMARY.md` recording:
- Final identifier names confirmed (HammingDistance, HammingDistanceRunes, HammingScore, HammingScoreRunes — zero drift).
- Benchmark numbers observed (0 allocs/op on Short/Medium/Long ASCII; 2 allocs/op on Unicode rune path).
- The exact text of the locked unequal-length godoc paragraph as written into hamming.go (verbatim quote for traceability).
- Whether the equal-length-constrained triangle-inequality property test was added or omitted, with rationale.
- Coverage percentages.
- Any deviations from the plan and their rationale.
</output>
