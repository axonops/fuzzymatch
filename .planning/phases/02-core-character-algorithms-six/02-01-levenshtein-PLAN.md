---
phase: 02-core-character-algorithms-six
plan: 01
type: execute
wave: 1
depends_on: []
files_modified:
  - levenshtein.go
  - dispatch_levenshtein.go
  - levenshtein_test.go
  - levenshtein_bench_test.go
  - levenshtein_fuzz_test.go
  - props_test.go
  - example_test.go
  - algorithms_golden_test.go
  - algoid_test.go
  - testdata/golden/algorithms.json
  - testdata/fuzz/FuzzLevenshteinScore/seed-001
  - tests/bdd/features/levenshtein.feature
  - tests/bdd/steps/algorithms_steps.go
autonomous: true
requirements:
  - CHAR-01
  - PERF-01
  - PERF-02
  - PERF-03
  - TEST-01
  - TEST-02
  - TEST-04
  - TEST-05
  - DET-02
  - DET-04
  - DX-02
tags: [levenshtein, two-row-dp, ascii-fast-path, dispatch-registration, canonical-pattern]

must_haves:
  truths:
    # Goal-backward truths (user-observable)
    - "A caller can `import fuzzymatch` and call LevenshteinScore(\"kitten\", \"sitting\") and receive 1 - 3/7 ≈ 0.5714 (± 1e-9)"
    - "LevenshteinDistance(\"kitten\", \"sitting\") returns the int 3"
    - "LevenshteinDistance(\"\", \"\") returns 0; LevenshteinScore(\"\", \"\") returns exactly 1.0"
    - "LevenshteinDistance(\"abc\", \"\") returns 3; LevenshteinScore(\"abc\", \"\") returns exactly 0.0"
    - "LevenshteinDistance(\"saturday\", \"sunday\") returns 3 (Wagner-Fischer 1974 reference vector)"
    - "LevenshteinScoreRunes correctly handles multi-byte UTF-8 input (e.g. \"café\" vs \"cafe\" gives the rune-aware distance, not byte-aware)"
    - "dispatch[AlgoLevenshtein] is non-nil after package init and equals LevenshteinScore"
    # Cross-cutting truths (every plan in this phase)
    - "Every new .go file starts with the exact Apache-2.0 header block from normalise.go lines 1-13 (verified by scripts/verify-license-headers.sh exit 0)"
    - "levenshtein.go contains a `// Source: Levenshtein, V. I. (1965). \"Binary codes capable of correcting deletions, insertions, and reversals.\" Soviet Physics Doklady, 10(8):707-710.` block at the top of the file-level godoc"
    - "Two-row DP confirmed by code review: inner loop maintains exactly two []int slices of length n+1, no [m+1][n+1]int table"
    - "ASCII fast path detected via the existing isASCII helper from normalise.go (NOT redeclared); rune path uses []rune(s) only when input bypasses the ASCII threshold"
    - "Score normalisation guard `if maxLen == 0 { return 1.0 }` present before the division (NaN/Inf/-0 prevention)"
    - "No math.Pow / math.Log / math.Exp / math.Sqrt / math.FMA used anywhere in levenshtein.go (grep verified)"
    - "No init() function in any new file in this plan; dispatch registration uses the `var _ = func() bool { ... }()` package-load-time idiom"
    - "Sum reductions (where any) are left-to-right; no parallel sums, no math.FMA"
    - "Property tests (testing/quick) pass for RangeBounds, Identity, Symmetric, TriangleInequality (distance), NoNaN, NoInf, NoNegativeZero — all in props_test.go"
    - "Fuzz test FuzzLevenshteinScore exists with at least one programmatic seed entry covering the canonical Wagner-Fischer pair plus an invalid-UTF-8 seed"
    - "Benchmark file uses b.ReportAllocs() before b.ResetTimer(); BenchmarkLevenshteinScore_ASCII_Short and _Medium report 0 B/op, 0 allocs/op"
    - "BDD scenario in tests/bdd/features/levenshtein.feature exercises canonical reference vectors via a Scenario Outline + Examples table; godog suite green via `cd tests/bdd && go test ./...`"
    - "testdata/golden/algorithms.json exists, is marshalled via CanonicalMarshalForTest, contains entries sorted by Name, and includes Levenshtein_kitten_sitting, Levenshtein_saturday_sunday, Levenshtein_identical, Levenshtein_empty_empty"
    - "algoid_test.go's TestDispatch_AllNilAtPhase1 is updated/renamed (e.g. TestDispatch_LevenshteinRegistered + TestDispatch_UnregisteredSlotsAreNil) so it no longer asserts slot AlgoLevenshtein is nil"
    - "ExampleLevenshteinScore runs and the `// Output:` block matches byte-for-byte"
    - "make check green (lint, vet, race, coverage-check, verify-determinism, verify-license-headers, verify-no-runtime-deps, tidy-check)"
  artifacts:
    - path: "levenshtein.go"
      provides: "LevenshteinDistance, LevenshteinDistanceRunes, LevenshteinScore, LevenshteinScoreRunes + unexported DP kernel"
      min_lines: 120
      contains: "// Source: Levenshtein"
    - path: "dispatch_levenshtein.go"
      provides: "Package-load-time registration of LevenshteinScore into dispatch[AlgoLevenshtein]"
      contains: "dispatch[AlgoLevenshtein] = LevenshteinScore"
    - path: "levenshtein_test.go"
      provides: "Unit tests for identity, both-empty, one-empty, reference vectors, symmetry, byte vs rune equivalence"
    - path: "levenshtein_bench_test.go"
      provides: "BenchmarkLevenshteinScore_ASCII_Short/Medium/Long, _Unicode_Short — alloc-asserted"
    - path: "levenshtein_fuzz_test.go"
      provides: "FuzzLevenshteinScore — panic-free + score in [0,1]"
    - path: "props_test.go"
      provides: "TestProp_LevenshteinScore_RangeBounds, _Identity, _Symmetric; TestProp_LevenshteinDistance_TriangleInequality; TestProp_LevenshteinScore_NoNaN, _NoInf, _NoNegativeZero"
    - path: "algorithms_golden_test.go"
      provides: "TestGolden_Algorithms — assembles entries, sorts by Name, calls assertGolden"
    - path: "testdata/golden/algorithms.json"
      provides: "Canonical golden file with Levenshtein entries (sorted, LF-terminated, no BOM)"
      contains: "Levenshtein_kitten_sitting"
    - path: "example_test.go"
      provides: "ExampleLevenshteinScore runnable godoc example"
    - path: "tests/bdd/features/levenshtein.feature"
      provides: "Gherkin scenarios for Levenshtein reference vectors, identity, symmetry"
    - path: "tests/bdd/steps/algorithms_steps.go"
      provides: "Shared AlgorithmContext + Levenshtein step bindings (consumed by Wave 2 plans for new algorithms)"
  key_links:
    - from: "dispatch_levenshtein.go"
      to: "algoid.go (dispatch array)"
      via: "package-level var _ = func() bool { dispatch[AlgoLevenshtein] = LevenshteinScore; return true }()"
      pattern: "dispatch\\[AlgoLevenshtein\\]"
    - from: "levenshtein.go LevenshteinScore"
      to: "levenshtein.go LevenshteinDistance"
      via: "function call inside score normaliser"
      pattern: "LevenshteinDistance\\("
    - from: "levenshtein.go fast path"
      to: "normalise.go isASCII"
      via: "package-private call (same package fuzzymatch)"
      pattern: "isASCII\\("
    - from: "algorithms_golden_test.go TestGolden_Algorithms"
      to: "testdata/golden/algorithms.json"
      via: "assertGolden via CanonicalMarshalForTest"
      pattern: "assertGolden\\(t, \"algorithms\\.json\""
    - from: "tests/bdd/features/levenshtein.feature"
      to: "tests/bdd/steps/algorithms_steps.go"
      via: "godog step regex matching"
      pattern: "I compute the Levenshtein score between"

user_setup: []
---

<objective>
Implement Levenshtein edit distance as the canonical pattern for Phase 2's six character-based algorithms. This plan ships the algorithm end-to-end — algorithm + dispatch registration + unit tests + property tests + benchmarks + fuzz tests + golden file + BDD scenario + godoc example — and locks the file-by-file pattern (`<algo>.go`, `dispatch_<algo>.go`, `<algo>_test.go`, `<algo>_bench_test.go`, `<algo>_fuzz_test.go`, plus the shared `props_test.go`, `algorithms_golden_test.go`, `example_test.go`, `tests/bdd/features/<algo>.feature`, `tests/bdd/steps/algorithms_steps.go`) that the five Wave 2 plans replicate without re-deciding any pattern detail.

Purpose: derisk the entire pipeline (two-row DP with stack-allocated `[(maxStackInputLen+1)*2]int` buffer, ASCII fast path via the existing `isASCII` helper, dispatch registration via the `var _ = func() bool { ... }()` idiom, golden file extension via `CanonicalMarshalForTest`, BDD step infrastructure, fuzz seed corpus layout) on the simplest non-trivial algorithm so the parallel Wave 2 plans can copy-rename-adjust without exploring the codebase.

Output: a working `LevenshteinScore` that returns deterministic, 0-allocation scores on ASCII ≤ 50 bytes, plus all the cross-cutting infrastructure (golden file, BDD harness, example, property suite, ExampleXxx) that Wave 2 plans extend.
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
@.planning/phases/02-core-character-algorithms-six/02-CONTEXT.md
@.planning/phases/02-core-character-algorithms-six/02-RESEARCH.md
@.planning/phases/02-core-character-algorithms-six/02-PATTERNS.md
@.planning/phases/02-core-character-algorithms-six/02-VALIDATION.md
@.planning/phases/01-foundation-infrastructure/01-04-determinism-infra-SUMMARY.md
@.planning/phases/01-foundation-infrastructure/01-05-primitives-algoid-errors-SUMMARY.md
@.planning/phases/01-foundation-infrastructure/01-06-primitives-normalise-SUMMARY.md
@docs/requirements.md
@.claude/skills/algorithm-correctness-standards/SKILL.md
@.claude/skills/algorithm-licensing-standards/SKILL.md
@.claude/skills/performance-standards/SKILL.md
@.claude/skills/determinism-standards/SKILL.md
@.claude/skills/go-coding-standards/SKILL.md
@.claude/skills/go-testing-standards/SKILL.md
@.claude/skills/fuzzymatch-review-protocol/SKILL.md

<interfaces>
Phase 1 contracts the Wave 1 executor consumes (extracted from existing source):

From algoid.go (lines 53-181, 310-324 — DO NOT MODIFY this file):
  type AlgoID int
  const (
      AlgoLevenshtein AlgoID = iota    // 0  — this plan registers slot 0
      AlgoDamerauLevenshteinOSA        // 1  — Wave 2 (plan 02-05)
      AlgoDamerauLevenshteinFull       // 2  — Wave 2 (plan 02-06)
      AlgoHamming                      // 3  — Wave 2 (plan 02-02)
      AlgoJaro                         // 4  — Wave 2 (plan 02-03)
      AlgoJaroWinkler                  // 5  — Wave 2 (plan 02-04)
      // ... 17 more constants
  )
  // unexported (visible inside package fuzzymatch only):
  numAlgorithms = int(AlgoRatcliffObershelp) + 1   // == 23
  var dispatch [numAlgorithms]func(a, b string) float64   // all-nil at Phase 1

From normalise.go (lines 159-168 — already in package fuzzymatch; do NOT redeclare):
  func isASCII(s string) bool   // returns true if every byte is < 0x80; "" returns true

From export_test.go (already exists, may need extension if test needs another re-export):
  const NumAlgorithmsForTest = numAlgorithms
  func DispatchLenForTest() int
  func DispatchEntryNilForTest(i int) bool
  var CanonicalMarshalForTest = canonicalMarshal

From golden_test.go:
  func assertGolden(t *testing.T, filename string, v any)
  // -update flag rewrites testdata/golden/<filename> through CanonicalMarshalForTest

From golden_canonical.go:
  func WriteGoldenFile(path string, v any) error
  // canonical form: indent two spaces, trailing LF, no BOM, UTF-8, deterministic key order

From tests/bdd/go.mod (sub-module — testify, godog, goleak available):
  github.com/cucumber/godog v0.15.0
  github.com/stretchr/testify v1.10.0    // permitted in tests/bdd/ only
  go.uber.org/goleak v1.3.0
  replace github.com/axonops/fuzzymatch => ../..
</interfaces>

<canonical_pattern_decisions>
The five decisions this plan locks for the entire phase (overrides RESEARCH.md open questions):

1. Algorithm file naming: unprefixed concatenation. Files: `levenshtein.go`, `hamming.go`, `jaro.go`, `jarowinkler.go`, `damerau_osa.go`, `damerau_full.go`. Rationale: matches `normalise.go`/`tokenise.go`/`algoid.go`/`errors.go`; underscores reserved for disambiguating the two Damerau variants.

2. Stack buffer threshold: `const maxStackInputLen = 64` declared once in `levenshtein.go` (package-level unexported); reused by DL-OSA and DL-Full plans. Two-row DP allocates `[(maxStackInputLen+1)*2]int` = `[130]int` = 1040 bytes on the stack.

3. Dispatch registration: separate `dispatch_<algo>.go` per algorithm, each using:
       var _ = func() bool { dispatch[AlgoXxx] = XxxScore; return true }()
   Wave 2 plans each touch only their own dispatch file — zero merge conflicts on `algoid.go`.

4. Rune variant strategy: eager `[]rune(s)` conversion (Pattern A from RESEARCH.md §Rune Variant Strategy). Documented in each algorithm file's godoc: "The rune variant allocates two []rune slices." The 0-alloc budget applies only to the byte ASCII path.

5. BDD: one feature file per algorithm — `tests/bdd/features/<algo>.feature`. Shared step definitions in `tests/bdd/steps/algorithms_steps.go` (this plan creates it; Wave 2 plans extend it with their step regexes).

Plus this plan establishes:

6. Cross-platform golden file strategy for Wave 2 parallelism: Wave 2 plans each write to `testdata/golden/_staging/<algo>.json` (per-algorithm entries only, same `goldenAlgorithmsFile` schema). Plan 02-07 (Wave 3 finalisation) merges all staging files into the canonical `testdata/golden/algorithms.json`. This Wave 1 plan creates `algorithms.json` with only Levenshtein entries; it does NOT create the `_staging` directory (that is Wave 2's responsibility).
</canonical_pattern_decisions>
</context>

<tasks>

<task type="auto" tdd="true">
  <name>Task 1: Implement levenshtein.go (algorithm) and dispatch_levenshtein.go (registration)</name>
  <files>levenshtein.go, dispatch_levenshtein.go</files>
  <read_first>
    - normalise.go (Apache-2.0 header block lines 1-13; isASCII helper lines 159-168; package-level doc-comment style; how a pure pure-function file is organised)
    - algoid.go (lines 53-181 — AlgoID enum; lines 310-324 — dispatch array declaration and godoc; lines 15-28 — "no init()" rationale)
    - tokenise.go (cross-check the unprefixed-filename + Apache-2.0-header pattern)
    - .planning/phases/02-core-character-algorithms-six/02-PATTERNS.md (Pattern 1 license header; Pattern 2 file structure; Pattern 3 isASCII reuse; Pattern 4 two-row DP + stack buffer; Pattern 5 score normalisation; Pattern 6 dispatch registration)
    - .planning/phases/02-core-character-algorithms-six/02-RESEARCH.md §Primary Sources — Levenshtein; §Implementation Patterns — Two-row DP; §Score Normalisation; §Allocation Budgets — Levenshtein
    - docs/requirements.md §7.1.1 (Levenshtein public API spec)
    - .claude/skills/algorithm-correctness-standards/SKILL.md (primary-source citation format)
    - .claude/skills/performance-standards/SKILL.md (two-row DP rules, escape-analysis check)
  </read_first>
  <action>
Create `levenshtein.go` in package fuzzymatch with:

1. Exact Apache-2.0 license header copied from normalise.go lines 1-13.
2. File-level godoc comment block opening with `// Source: Levenshtein, V. I. (1965). "Binary codes capable of correcting deletions, insertions, and reversals." Soviet Physics Doklady, 10(8):707-710.` plus a reference to Wagner & Fischer 1974 (JACM 21(1):168-173) for the two-row DP formulation. Include the recurrence in godoc (D[i,j] = min(D[i-1,j]+1, D[i,j-1]+1, D[i-1,j-1]+cost)). End the doc block with the "Implementation discipline" bullet list mirrored from normalise.go lines 36-43.
3. Declare the package-level unexported constant `maxStackInputLen = 64` with a godoc paragraph stating the stack-buffer threshold rationale (this constant is SHARED across Phase 2 — Wave 2 plans must NOT redeclare it).
4. Public API (exact identifiers — locked per CONTEXT.md and RESEARCH.md):
     - LevenshteinDistance(a, b string) int
     - LevenshteinDistanceRunes(a, b string) int
     - LevenshteinScore(a, b string) float64
     - LevenshteinScoreRunes(a, b string) float64
5. Each public function has a godoc starting with the function name. Document edge cases (both-empty, one-empty, identical), the score normalisation formula, and (on the rune variants) the "allocates two []rune slices" note. Document worst-case time O(m·n) and space O(min(m,n)).
6. Implementation:
     - Identity fast path: `if a == b { return 0 }` (covers both-empty and identical inputs).
     - Empty guards: `if m == 0 { return n }; if n == 0 { return m }`.
     - Symmetry-for-perf: swap so b is the shorter inner-loop dimension.
     - ASCII fast path: when `n <= maxStackInputLen` AND `isASCII(a)` AND `isASCII(b)`, allocate `var buf [(maxStackInputLen+1)*2]int` on the stack and pass `buf[:n+1]` and `buf[n+1:2*(n+1)]` to the inner DP kernel.
     - Heap path: when n exceeds threshold OR non-ASCII bytes present (the byte-level DP is semantically incorrect on multi-byte UTF-8 — callers wanting Unicode use LevenshteinScoreRunes), use `make([]int, n+1)` for both rows.
     - Inner DP kernel `levenshteinDP(a, b string, m, n int, prev, curr []int) int` — initialise `prev[j] = j`; iterate i=1..m: set `curr[0]=i`; iterate j=1..n: compute three-way min; swap prev/curr; return `prev[n]`. Use direct byte indexing `a[i-1]`, `b[j-1]` (NEVER `[]byte(a)`).
     - LevenshteinScore: compute `maxLen := max(len(a), len(b))`; guard `if maxLen == 0 { return 1.0 }`; return `1.0 - float64(dist)/float64(maxLen)`.
     - LevenshteinDistanceRunes / LevenshteinScoreRunes: eager `ra := []rune(a); rb := []rune(b)`, then a separate inner kernel operating on `[]rune` slices.
7. NO `init()` functions in this file. NO `math.Pow`, `math.Log`, `math.Exp`, `math.Sqrt`, `math.FMA`. NO map iteration. NO `[]byte(a)` allocations on the hot path (use direct byte indexing).

Create `dispatch_levenshtein.go` in package fuzzymatch with:

1. Same Apache-2.0 header.
2. File-level godoc comment: "dispatch_levenshtein.go registers LevenshteinScore into the dispatch table at package load time. Sole writer to dispatch[AlgoLevenshtein]." Reference algoid.go for the dispatch array declaration.
3. The registration idiom (NO init()):
       var _ = func() bool {
           dispatch[AlgoLevenshtein] = LevenshteinScore
           return true
       }()
4. NOTHING else — this file is registration-only. The blank-identifier variable + immediately-invoked-function pattern is the LOCKED canonical form for Phase 2; Wave 2 plans copy it character-for-character with only the AlgoXxx and XxxScore identifiers changed.

After writing both files, run `go build ./...` to confirm the package compiles; run `bash scripts/verify-license-headers.sh` to confirm the header gate passes; run `go vet ./...` to confirm no static issues.
  </action>
  <verify>
    <automated>go build ./... && go vet ./... && bash scripts/verify-license-headers.sh && go test -run TestDispatch_SizedForCatalogue -count=1 ./...</automated>
  </verify>
  <acceptance_criteria>
    - levenshtein.go starts with the Apache-2.0 header (matches normalise.go lines 1-13 byte-for-byte; verified by `diff <(head -13 normalise.go) <(head -13 levenshtein.go)` exit 0).
    - levenshtein.go contains the literal string `// Source: Levenshtein, V. I. (1965).` in its file-level godoc block.
    - `grep -E 'math\.(Pow|Log|Exp|Sqrt|FMA)' levenshtein.go` returns no matches.
    - `grep -E '^func init\(' levenshtein.go dispatch_levenshtein.go` returns no matches.
    - `grep -c '\[\]byte\(' levenshtein.go` returns 0 (no allocating string-to-bytes conversions on the hot path).
    - levenshtein.go declares `const maxStackInputLen = 64` exactly once.
    - dispatch_levenshtein.go contains the literal text `dispatch[AlgoLevenshtein] = LevenshteinScore` exactly once.
    - dispatch_levenshtein.go uses `var _ = func() bool {` registration idiom (NOT `init()`).
    - `go build ./...` exits 0.
    - `go vet ./...` exits 0.
    - `bash scripts/verify-license-headers.sh` exits 0.
  </acceptance_criteria>
  <behavior>
    - LevenshteinDistance("kitten", "sitting") == 3
    - LevenshteinDistance("saturday", "sunday") == 3
    - LevenshteinDistance("", "") == 0
    - LevenshteinDistance("", "abc") == 3
    - LevenshteinDistance("abc", "abc") == 0
    - LevenshteinScore("kitten", "sitting") in [0.571428, 0.571429] (within 1e-9 of 1 - 3/7)
    - LevenshteinScore("", "") == 1.0 (exact)
    - LevenshteinScore("abc", "") == 0.0 (exact)
    - LevenshteinScore returns symmetric values: LevenshteinScore(a, b) == LevenshteinScore(b, a) for any a, b
    - LevenshteinScoreRunes("café", "cafe") returns the rune-aware score (3 runes shared, 1 differs → 1 - 1/4 = 0.75)
    - dispatch[AlgoLevenshtein] is non-nil and equals LevenshteinScore after package load
  </behavior>
  <done>
    levenshtein.go and dispatch_levenshtein.go committed. Package compiles. License-header verifier passes. dispatch[AlgoLevenshtein] is populated (verifiable via the next task's tests).
  </done>
</task>

<task type="auto" tdd="true">
  <name>Task 2: Write tests (unit + property + benchmark + fuzz) + update algoid_test.go dispatch assertion</name>
  <files>levenshtein_test.go, levenshtein_bench_test.go, levenshtein_fuzz_test.go, props_test.go, algoid_test.go, testdata/fuzz/FuzzLevenshteinScore/seed-001, example_test.go</files>
  <read_first>
    - normalise_test.go (test file structure; t.Errorf/t.Fatalf usage; "stdlib testing only" comment block; table-driven test pattern; TestProp_* convention lines 275-346)
    - normalise_bench_test.go (b.ReportAllocs() before b.ResetTimer(); var sink pattern lines 58-65 to defeat compiler dead-code elimination; ASCII Short/Medium/Long/Unicode breakdown)
    - normalise_fuzz_test.go (f.Add seed corpus pattern lines 54-75; FuzzXxx panic-free + invariant template; testdata/fuzz/FuzzXxx/seed-NNN file format)
    - algoid_test.go (existing TestDispatch_SizedForCatalogue and TestDispatch_AllNilAtPhase1 lines 212-240 — must be updated)
    - .planning/phases/02-core-character-algorithms-six/02-PATTERNS.md (Pattern 7 unit test structure; Pattern 9 benchmark structure; Pattern 10 fuzz structure; Pattern 11 property tests; Pattern 14 godoc examples; Pattern 13 AlgoID enum + dispatch test update)
    - .planning/phases/02-core-character-algorithms-six/02-RESEARCH.md §Mathematical Invariants; §Allocation Budgets; §Determinism Constraints — DET-04 properties
    - .claude/skills/go-testing-standards/SKILL.md (coverage floors, property-test conventions, no testify in root)
  </read_first>
  <action>
Create `levenshtein_test.go` (package `fuzzymatch_test`, stdlib `testing` only — NO testify):

1. Apache-2.0 header + file-level godoc comment per Pattern 7.
2. Test functions (table-driven where applicable):
     - TestLevenshtein_BothEmpty — Distance and Score on ("", "") return 0 and 1.0
     - TestLevenshtein_OneEmpty — covers ("", "abc") and ("abc", "") → distance 3, score 0.0
     - TestLevenshtein_Identical — ("abc", "abc") → distance 0, score 1.0
     - TestLevenshtein_ReferenceVectors — table-driven over the four canonical pairs from RESEARCH.md §Primary Sources — Levenshtein (kitten/sitting → 3 / 0.5714…; saturday/sunday → 3 / 0.6250; abc/abc → 0 / 1.0; ""/abc → 3 / 0.0). Float comparison uses `math.Abs(got - want) <= 1e-9`.
     - TestLevenshtein_Symmetry — verifies Score(a,b) == Score(b,a) for the reference-vector pairs.
     - TestLevenshtein_DistanceRunes_MultiByte — "café"/"cafe" returns the rune-aware distance (1) not the byte-level distance (2). Document the cited expectation in a comment.
     - TestLevenshtein_ASCII_vs_Rune_Equivalence — for an ASCII-only pair, byte and rune variants return identical results.
3. Use stdlib `math.Abs` for float tolerance (no helper redeclaration). Import `math` + `testing` + the package under test.

Create `levenshtein_bench_test.go` (package `fuzzymatch_test`):

1. Apache-2.0 header + file-level godoc comment per Pattern 9. Document the allocation targets: 0 allocs on ASCII ≤ 50 chars (PERF-01 budget).
2. Benchmarks:
     - BenchmarkLevenshteinScore_ASCII_Short — kitten / sitting (6 and 7 bytes)
     - BenchmarkLevenshteinScore_ASCII_Medium — two 50-char ASCII identifiers (use a `const a50 = "..."` declared once at top of file)
     - BenchmarkLevenshteinScore_ASCII_Long — two 500-char ASCII strings (the heap path; alloc count > 0 expected)
     - BenchmarkLevenshteinScore_Unicode_Short — a multi-byte UTF-8 pair (rune path; expect 2 allocs from `[]rune` conversion)
3. Every benchmark calls `b.ReportAllocs()` then `b.ResetTimer()` (in that order). Use `var sink float64` and `if sink < 0 { b.Fatal(...) }` to defeat DCE.
4. Add an `AllocsPerRun` allocation-assertion guard inside the Short/Medium benchmarks as a runtime gate: a `TestLevenshteinScore_ZeroAllocs_ASCII_Short` (in `levenshtein_test.go`, not the bench file) that calls `testing.AllocsPerRun(100, func() { _ = fuzzymatch.LevenshteinScore("kitten", "sitting") })` and fails if it exceeds 0. This pins the budget at test time, not benchmark time.

Create `levenshtein_fuzz_test.go` (package `fuzzymatch_test`):

1. Apache-2.0 header + file-level godoc per Pattern 10.
2. `FuzzLevenshteinScore` — programmatically seed with the reference vectors plus invalid-UTF-8 (`"\xff\xfe"`, `"\xc0\x80"`) and Cyrillic (`"Привет"`/`"привет"`) seeds. Fuzz body asserts: no panic (implicit), `!math.IsNaN(got)`, `!math.IsInf(got, 0)`, `got >= 0.0 && got <= 1.0`.
3. Create `testdata/fuzz/FuzzLevenshteinScore/seed-001` in the `go test fuzz v1` format mirroring `testdata/fuzz/FuzzNormalise/seed-001`. Include a representative pair (e.g. `kitten` / `sitting`).

Create `props_test.go` (package `fuzzymatch_test`) — shared file for ALL Phase 2 property tests:

1. Apache-2.0 header + file-level godoc explaining this is the shared property-test file extended by Wave 2 plans per Pattern 11.
2. Property tests for Levenshtein (Wave 2 plans append their algorithm's properties to the same file):
     - TestProp_LevenshteinScore_RangeBounds — score in [0,1] for any a, b via testing/quick.Check
     - TestProp_LevenshteinScore_Identity — Score(x, x) == 1.0 for non-empty x
     - TestProp_LevenshteinScore_Symmetric — Score(a,b) == Score(b,a)
     - TestProp_LevenshteinDistance_TriangleInequality — Distance(a,c) <= Distance(a,b) + Distance(b,c)
     - TestProp_LevenshteinScore_NoNaN — !math.IsNaN(Score(a,b))
     - TestProp_LevenshteinScore_NoInf — !math.IsInf(Score(a,b), 0)
     - TestProp_LevenshteinScore_NoNegativeZero — when score == 0.0, !math.Signbit(score)

Create `example_test.go` (package `fuzzymatch_test`) — runnable godoc example file:

1. Apache-2.0 header + file-level godoc per Pattern 14.
2. `func ExampleLevenshteinScore()` printing `fmt.Printf("%.4f\n", fuzzymatch.LevenshteinScore("kitten", "sitting"))` with the `// Output: 0.5714` block. Wave 2 plans add their ExampleXxx functions to the SAME file.

Modify `algoid_test.go` — update the existing `TestDispatch_AllNilAtPhase1` (lines 226-240) to accept that AlgoLevenshtein is now populated:

1. Rename to `TestDispatch_LevenshteinRegistered_OthersNil` (or split into two tests: `TestDispatch_LevenshteinRegistered` + `TestDispatch_UnregisteredSlotsAreNil`).
2. The new test asserts `DispatchEntryNilForTest(int(fuzzymatch.AlgoLevenshtein)) == false` AND for every i in [int(AlgoDamerauLevenshteinOSA), 22], `DispatchEntryNilForTest(i) == true`.
3. Update the file's godoc to reflect the new Phase-2 reality of mixed populated/nil slots. Wave 2 plans further update this list, removing their algorithm's slot from the nil list.

Run the test suite, fix any failures, then verify:

  go test -race -shuffle=on -count=1 -run 'TestLevenshtein|TestProp_Levenshtein|TestDispatch_Levenshtein|ExampleLevenshteinScore' ./...
  go test -bench=BenchmarkLevenshteinScore_ASCII -benchmem -run=^$ -count=3 ./...
  go test -fuzz=FuzzLevenshteinScore -fuzztime=30s ./...

The first command must exit 0. The second must report `0 B/op  0 allocs/op` for `BenchmarkLevenshteinScore_ASCII_Short` and `BenchmarkLevenshteinScore_ASCII_Medium` (the M and L variants may allocate). The fuzz run is informational locally; verify it does not crash.
  </action>
  <verify>
    <automated>go test -race -shuffle=on -count=1 -run 'TestLevenshtein|TestProp_Levenshtein|TestDispatch_Levenshtein|ExampleLevenshteinScore' ./... && go test -bench=BenchmarkLevenshteinScore_ASCII_Short -benchmem -run=^$ -count=3 ./... 2>&1 | grep -E '0 B/op[[:space:]]+0 allocs/op'</automated>
  </verify>
  <acceptance_criteria>
    - All TestLevenshtein_* tests pass.
    - All TestProp_Levenshtein* tests pass under testing/quick (default 100 random invocations).
    - TestLevenshteinScore_ZeroAllocs_ASCII_Short reports 0 allocations via testing.AllocsPerRun.
    - BenchmarkLevenshteinScore_ASCII_Short and BenchmarkLevenshteinScore_ASCII_Medium each report `0 B/op  0 allocs/op` in `go test -bench=... -benchmem -count=3`.
    - ExampleLevenshteinScore output matches `0.5714\n` byte-for-byte (test exit code 0).
    - FuzzLevenshteinScore completes 30s of fuzzing without panic or invariant violation.
    - testdata/fuzz/FuzzLevenshteinScore/seed-001 exists and parses as `go test fuzz v1` corpus.
    - algoid_test.go's renamed/updated TestDispatch_* assertions pass — Levenshtein slot non-nil, slots 1-22 nil.
    - `grep -c '"github.com/stretchr/testify' levenshtein_test.go levenshtein_bench_test.go levenshtein_fuzz_test.go props_test.go example_test.go` returns 0 (no testify in root tests).
  </acceptance_criteria>
  <behavior>
    - Unit test suite covers identity, both-empty, one-empty, reference vectors, symmetry, multi-byte rune handling, ASCII vs rune equivalence.
    - Property tests cover all six DET-04 + invariant categories (RangeBounds, Identity, Symmetric, TriangleInequality, NoNaN, NoInf, NoNegativeZero).
    - Benchmarks pin the 0-alloc target for ASCII Short/Medium; heap and rune paths run without panic.
    - Fuzz harness panic-free + invariant-preserving on programmatic seeds plus 30s of random input.
    - ExampleLevenshteinScore visible on pkg.go.dev with the Output block verified.
  </behavior>
  <done>
    All test files committed. Test suite green for the Levenshtein-scoped pattern. Dispatch assertion updated. Wave 2 plans now have a copy-template for their algorithm-specific tests.
  </done>
</task>

<task type="auto" tdd="true">
  <name>Task 3: Golden file + BDD harness + final phase quality gate</name>
  <files>algorithms_golden_test.go, testdata/golden/algorithms.json, tests/bdd/features/levenshtein.feature, tests/bdd/steps/algorithms_steps.go</files>
  <read_first>
    - .planning/phases/01-foundation-infrastructure/01-04-determinism-infra-SUMMARY.md (golden-file canonical-byte-form rules: indent 2 spaces, trailing LF, no BOM, sorted entries)
    - golden_canonical.go (canonicalMarshal + WriteGoldenFile — the LOCKED marshaller)
    - golden_test.go (assertGolden + -update flag pattern; how normalisation.json's TestGolden_Normalisation works)
    - export_test.go (CanonicalMarshalForTest re-export)
    - testdata/golden/normalisation.json (reference schema — version + entries with typed structs sorted by Name)
    - tests/bdd/doc.go (the existing BDD harness skeleton in tests/bdd/)
    - tests/bdd/go.mod (godog v0.15.0, testify v1.10.0, goleak v1.3.0, replace ../..)
    - .planning/phases/02-core-character-algorithms-six/02-PATTERNS.md (Pattern 12 golden file extension; Pattern 15 BDD feature + steps)
    - .planning/phases/02-core-character-algorithms-six/02-RESEARCH.md §Golden File Integration; §BDD Scenario Coverage
  </read_first>
  <action>
Create `algorithms_golden_test.go` in the root (package `fuzzymatch_test`):

1. Apache-2.0 header + file-level godoc comment explaining the file pins the byte-stable canonical golden form for ALL Phase 2 algorithm scores. Reference plan 01-04's locked canonical byte form (indent 2 spaces, trailing LF, no BOM).
2. Declare typed structs matching the schema in RESEARCH.md §Golden File Integration:

       type goldenAlgorithmEntry struct {
           Name          string  `json:"name"`
           Algorithm     string  `json:"algorithm"`
           A             string  `json:"a"`
           B             string  `json:"b"`
           ExpectedScore float64 `json:"expected_score"`
       }

       type goldenAlgorithmsFile struct {
           Version int                    `json:"version"`
           Entries []goldenAlgorithmEntry `json:"entries"`
       }

3. `func TestGolden_Algorithms(t *testing.T)` — calls `buildAlgorithmGoldenEntries()` (helper below), sorts the result by `Name` using `sort.Slice`, wraps in `goldenAlgorithmsFile{Version: 1, Entries: entries}`, calls `assertGolden(t, "algorithms.json", file)`.
4. `func buildAlgorithmGoldenEntries(t *testing.T) []goldenAlgorithmEntry` — returns the four Levenshtein entries from RESEARCH.md §Minimum entry set:
     - `Levenshtein_kitten_sitting` — a "kitten", b "sitting", ExpectedScore from a live `fuzzymatch.LevenshteinScore` call
     - `Levenshtein_saturday_sunday` — a "saturday", b "sunday"
     - `Levenshtein_identical` — a "abc", b "abc"
     - `Levenshtein_empty_empty` — a "", b ""
   Wave 2 plans extend this function (or merge from `_staging/<algo>.json` in plan 02-07) — explicitly comment that hook.
5. Generate the initial `testdata/golden/algorithms.json`: run `go test -run TestGolden_Algorithms -update -count=1 ./...` once. Inspect the result; commit the file. The file must be byte-stable across re-runs (rerun the test without `-update` and confirm it exits 0).

Create `tests/bdd/features/levenshtein.feature`:

1. File-level comment with primary source (Levenshtein 1965).
2. Feature: Levenshtein similarity algorithm
3. Scenario Outline "canonical reference vectors" with the four reference-vector rows (a/b/score columns; tolerance 0.0001).
4. Scenario "identical strings score 1.0" — user_id / user_id → 1.0 exact.
5. Scenario "both-empty strings score 1.0" — "" / "" → 1.0 exact.
6. Scenario "score is symmetric" — compute Levenshtein between (kitten, sitting) and (sitting, kitten); both scores equal.

Create `tests/bdd/steps/algorithms_steps.go` (package `steps`):

1. Apache-2.0 header.
2. Imports: `fmt`, `math`, `github.com/axonops/fuzzymatch`, `github.com/cucumber/godog`. (testify is permitted in tests/bdd but is not required for this initial harness; document at the top that step functions return `error` rather than calling t.Errorf.)
3. `type AlgorithmContext struct { lastScore float64; lastScore2 float64 }` — Wave 2 plans extend this if they need per-algorithm state.
4. Step functions (each returns `error`):
     - `(ctx *AlgorithmContext) iComputeTheLevenshteinScoreBetween(a, b string) error` — calls fuzzymatch.LevenshteinScore.
     - `(ctx *AlgorithmContext) iComputeTheLevenshteinScoreBetween2(a, b string) error` — overwrites lastScore2 (used by the "symmetric" scenario).
     - `(ctx *AlgorithmContext) theScoreShouldBeApproximately(expected, tolerance float64) error` — math.Abs(ctx.lastScore - expected) <= tolerance.
     - `(ctx *AlgorithmContext) theScoreShouldBeExactly(expected float64) error` — ctx.lastScore == expected.
     - `(ctx *AlgorithmContext) bothScoresShouldBeEqual() error` — ctx.lastScore == ctx.lastScore2.
5. `func InitializeScenario(ctx *godog.ScenarioContext)` — instantiates AlgorithmContext, registers the regex step bindings. Wave 2 plans append their algorithm's step regexes to the SAME InitializeScenario function.
6. If tests/bdd/doc.go has a TestMain wiring godog with goleak.VerifyTestMain, leave it untouched and confirm it picks up the new feature file from `tests/bdd/features/`. If no TestMain exists, defer to the existing tests/bdd convention — do not invent a new one.

Run the full quality gate:

  go test -race -shuffle=on -count=1 ./...
  cd tests/bdd && go test ./... && cd ..
  make verify-determinism
  make check

`make check` must exit 0. The golden file must be canonical-form (re-running TestGolden_Algorithms WITHOUT `-update` must succeed and produce no diff).
  </action>
  <verify>
    <automated>go test -race -shuffle=on -count=1 -run 'TestGolden_Algorithms' ./... && (cd tests/bdd && go test -count=1 ./...) && make verify-determinism</automated>
  </verify>
  <acceptance_criteria>
    - testdata/golden/algorithms.json exists and contains exactly four Levenshtein entries sorted alphabetically by Name (Levenshtein_empty_empty, Levenshtein_identical, Levenshtein_kitten_sitting, Levenshtein_saturday_sunday).
    - The file ends with a single LF byte; no BOM; 2-space indent (verifiable via `xxd testdata/golden/algorithms.json | head` showing canonical form).
    - Re-running `go test -run TestGolden_Algorithms -count=1 ./...` without `-update` exits 0 (the file is byte-stable; no diff).
    - tests/bdd/features/levenshtein.feature parses as valid Gherkin and includes at least four scenarios (the Scenario Outline counts as one feature element with multiple Examples rows).
    - `cd tests/bdd && go test -count=1 ./...` exits 0 (godog suite green; goleak detects no leaks).
    - `make verify-determinism` exits 0.
    - `make check` exits 0 (lint, vet, race, coverage, license headers, no-runtime-deps, tidy-check, verify-determinism all green).
  </acceptance_criteria>
  <behavior>
    - algorithms.json schema mirrors normalisation.json (version + entries sorted by Name).
    - The golden file is byte-stable across re-runs and cross-platform-identical (the CI matrix will diff it; locally the canonical form is enforced).
    - BDD harness exercises canonical reference vectors via Scenario Outline and pins identity, symmetry, both-empty as separate scenarios.
    - The algorithms_steps.go AlgorithmContext + InitializeScenario shape is the template Wave 2 plans extend.
  </behavior>
  <done>
    Golden file committed and stable. BDD feature + steps committed. `make check` green. Wave 1's canonical pattern is now established for Wave 2 to copy: every Wave 2 plan touches exactly its own `<algo>.go`, `dispatch_<algo>.go`, `<algo>_{test,bench_test,fuzz_test}.go`, appends to `props_test.go` and `example_test.go`, writes `testdata/golden/_staging/<algo>.json`, creates `tests/bdd/features/<algo>.feature`, and appends to `tests/bdd/steps/algorithms_steps.go`.
  </done>
</task>

</tasks>

<threat_model>
## Trust Boundaries

| Boundary | Description |
|----------|-------------|
| Caller → fuzzymatch.LevenshteinScore | Untrusted strings (any length, any byte sequence including invalid UTF-8) crossing into the algorithm. No other boundary — this is a pure-function library; no I/O, no network, no parsing of user-controlled config. |

## STRIDE Threat Register

| Threat ID | Category | Component | Disposition | Mitigation Plan |
|-----------|----------|-----------|-------------|-----------------|
| T-02-01-01 | Denial of Service | LevenshteinDistance/Score (O(m·n) on adversarial inputs) | accept | Document worst-case O(m·n) in godoc; caller is responsible for input-length caps. The library refuses to enforce a cap because that would break the pure-function contract. PERF-01 + benchstat ensure no super-linear slowdown sneaks in. Fuzz tests cover panic-freedom on arbitrary inputs. |
| T-02-01-02 | Denial of Service | DP buffer escape-to-heap on adversarial inputs > maxStackInputLen | mitigate | Heap path uses `make([]int, n+1)` which is O(n) memory — bounded linearly, NOT quadratically. No `[m+1][n+1]int` table is ever allocated. Verified by code review and the two-row-DP property (PERF-03). |
| T-02-01-03 | Information Disclosure | Malformed UTF-8 input causing panic / leaking internal state | mitigate | Byte-level Levenshtein operates on bytes directly (no UTF-8 decoding) — invalid UTF-8 cannot panic. Rune variants use `[]rune(s)` which Go's stdlib normalises to U+FFFD on malformed bytes (no panic). FuzzLevenshteinScore harness in Task 2 includes invalid-UTF-8 seeds (`"\xff\xfe"`, `"\xc0\x80"`) and asserts no panic, no NaN, no Inf, no out-of-[0,1] return. |
| T-02-01-04 | Tampering | dispatch[AlgoLevenshtein] overwritten by a later plan or external code | mitigate | dispatch is unexported (package fuzzymatch only). The var _ = func() bool { ... }() registration runs once at package load. There is no public mutator. Any future code attempting to overwrite the slot must edit the package source — caught by code review. T-01-05-05 mitigation from plan 01-05 still applies. |
| T-02-01-05 | Tampering | Golden file (algorithms.json) tampered with to weaken cross-platform determinism gate | mitigate | The golden file is committed to git; PR review + the CI matrix diff (`make verify-determinism` runs on all 5 platforms) detect any divergence. The CanonicalMarshalForTest re-export ensures the byte form is locked. Algorithm-correctness-reviewer agent reviews every golden-file update. |
| T-02-01-06 | Repudiation | "Score changed silently" — non-deterministic output across patch releases | mitigate | DET-02 enforced by the golden file gate: any change to a Levenshtein output requires updating algorithms.json, which surfaces in the PR diff. The CHANGELOG.md (Keep-a-Changelog format established in Phase 1) records score changes per release; score-changing edits require a minor version bump per REL-07. |

Severity assessment: all `high` items are `mitigate` or have a clear mitigation path. No `accept` of a `high` threat. ASVS L1 V5.1 (Input Validation) and V8.2 (Resilience to malformed input) addressed via fuzz tests + property tests. Plan passes the security gate.
</threat_model>

<verification>
1. Build: `go build ./...` exits 0.
2. Vet: `go vet ./...` exits 0.
3. License headers: `bash scripts/verify-license-headers.sh` exits 0 (all 4 new .go files in root + algoid_test.go modification carry the Apache-2.0 header).
4. No-runtime-deps: `bash scripts/verify-no-runtime-deps.sh` exits 0 (no new require entries in root go.mod).
5. Unit + property tests: `go test -race -shuffle=on -count=1 -run 'TestLevenshtein|TestProp_Levenshtein|TestDispatch_Levenshtein' ./...` exits 0.
6. Example: `go test -run ExampleLevenshteinScore ./...` exits 0.
7. Allocation budget: `go test -bench=BenchmarkLevenshteinScore_ASCII_Short -benchmem -run=^$ -count=3 ./...` reports `0 B/op  0 allocs/op` for the Short benchmark.
8. Fuzz smoke: `go test -fuzz=FuzzLevenshteinScore -fuzztime=30s -run=^$ ./...` completes without crash or invariant violation (informational locally; CI runs longer fuzz windows per plan 01-02).
9. Golden file: `go test -run TestGolden_Algorithms -count=1 ./...` exits 0 WITHOUT `-update` (file is byte-stable).
10. BDD: `(cd tests/bdd && go test -race -shuffle=on -count=1 ./...)` exits 0.
11. Determinism: `make verify-determinism` exits 0.
12. Coverage: `make coverage && make coverage-check` — overall ≥ 95%, per-file ≥ 90%, 100% on the four new public symbols (LevenshteinDistance, LevenshteinDistanceRunes, LevenshteinScore, LevenshteinScoreRunes).
13. Full quality gate: `make check` exits 0.
</verification>

<success_criteria>
- A caller can `import "github.com/axonops/fuzzymatch"` and obtain deterministic Levenshtein distance and similarity scores via the four public functions.
- The byte-level fast path on ASCII ≤ 50 chars allocates zero bytes (PERF-01, PERF-02, PERF-06 satisfied for Levenshtein).
- The implementation uses two-row DP, no full DP table (PERF-03 satisfied for Levenshtein).
- Score is in [0.0, 1.0] for any input; never NaN, Inf, or -0 (DET-04 satisfied for Levenshtein).
- Cross-platform byte-identical golden output via testdata/golden/algorithms.json (DET-02 partial — full coverage lands in plan 02-07 after all six algorithms ship).
- The canonical pattern (file naming, dispatch idiom, test/bench/fuzz/golden/BDD/example layout) is locked and documented in code comments so Wave 2 plans can copy it deterministically.
- All required gates (license headers, no-runtime-deps, lint, vet, race, tidy, coverage, determinism, BDD) pass via `make check`.
</success_criteria>

<output>
After completion, create `.planning/phases/02-core-character-algorithms-six/02-01-levenshtein-SUMMARY.md` per the standard summary template, recording:
- Final identifier names (LevenshteinDistance, LevenshteinDistanceRunes, LevenshteinScore, LevenshteinScoreRunes — confirm zero drift from this plan).
- Benchmark numbers observed locally (B/op, allocs/op for ASCII Short/Medium/Long, Unicode Short).
- Coverage percentages (overall, per-file levenshtein.go, public-symbol).
- The exact maxStackInputLen value shipped (locked at 64 by this plan; later phases may revisit).
- The Wave 2 hand-off contract: which functions in algorithms_golden_test.go, props_test.go, example_test.go, and tests/bdd/steps/algorithms_steps.go are append-points for Wave 2 plans.
- Any deviations from the plan and their rationale.
</output>
