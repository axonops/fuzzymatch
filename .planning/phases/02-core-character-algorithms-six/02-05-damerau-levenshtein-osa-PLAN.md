---
phase: 02-core-character-algorithms-six
plan: 05
type: execute
wave: 2
depends_on: [02-01-levenshtein]
files_modified:
  - damerau_osa.go
  - dispatch_damerau_osa.go
  - damerau_osa_test.go
  - damerau_osa_bench_test.go
  - damerau_osa_fuzz_test.go
  - props_test.go
  - example_test.go
  - algoid_test.go
  - testdata/golden/_staging/damerau_osa.json
  - testdata/fuzz/FuzzDamerauLevenshteinOSAScore/seed-001
  - tests/bdd/features/damerau_osa.feature
  - tests/bdd/steps/algorithms_steps.go
autonomous: true
requirements:
  - CHAR-02
  - PERF-01
  - PERF-02
  - PERF-03
  - TEST-01
  - TEST-02
  - TEST-04
  - TEST-05
  - DET-04
  - DX-02
tags: [damerau-levenshtein, osa, three-row-dp, optimal-string-alignment, boytsov-2011]

must_haves:
  truths:
    - "DamerauLevenshteinOSADistance(\"ab\", \"ba\") returns 1 (single transposition); Score returns 0.5 exactly"
    - "DamerauLevenshteinOSADistance(\"ca\", \"abc\") returns 3 — DISCRIMINATING vector vs Full DL (which returns 2). Score returns 0.0 exactly"
    - "DamerauLevenshteinOSADistance(\"abc\", \"abc\") returns 0; Score returns 1.0 exactly"
    - "DamerauLevenshteinOSADistance(\"\", \"\") returns 0; Score returns 1.0 exactly"
    - "DamerauLevenshteinOSADistance(\"\", \"abc\") returns 3; Score returns 0.0 exactly"
    - "DamerauLevenshteinOSAScoreRunes correctly handles multi-byte UTF-8"
    - "dispatch[AlgoDamerauLevenshteinOSA] is non-nil after package init"
    - "Apache-2.0 header on every new .go file"
    - "damerau_osa.go contains `// Source: Boytsov, L. (2011)` and a reference to Damerau, F. J. (1964) in file-level godoc"
    - "Three-row DP confirmed by code review: inner loop maintains exactly three []int slices of length n+1 (prevprev, prev, curr); NO full [m+1][n+1]int table"
    - "ASCII fast path uses stack-allocated `var buf [(maxStackInputLen+1)*3]int` (1560 bytes) when n <= maxStackInputLen and inputs are ASCII"
    - "Score normalisation guard `if maxLen == 0 { return 1.0 }` present (NaN/Inf/-0 prevention)"
    - "No math.Pow / math.Log / math.Exp / math.Sqrt / math.FMA in damerau_osa.go"
    - "No init() function; dispatch_damerau_osa.go uses var _ = func() bool { ... }() registration idiom"
    - "Property tests pass for RangeBounds, Identity, Symmetric, NoNaN, NoInf, NoNegativeZero AND TriangleInequality (DL-OSA distance satisfies triangle inequality per docs/requirements.md §15.2)"
    - "Benchmark BenchmarkDamerauLevenshteinOSAScore_ASCII_Short reports 0 B/op, 0 allocs/op"
    - "Fuzz test FuzzDamerauLevenshteinOSAScore exists with at least one seed plus invalid-UTF-8; panic-free + score in [0,1]"
    - "BDD scenarios in tests/bdd/features/damerau_osa.feature exercise reference vectors AND the discriminating-vector contract (\"ca\"/\"abc\" → 3)"
    - "testdata/golden/_staging/damerau_osa.json contains DL-OSA entries (DamerauLevenshteinOSA_ab_ba, _ca_abc, _identical, _empty_empty) sorted by Name"
    - "algoid_test.go updated: AlgoDamerauLevenshteinOSA removed from unregistered-slots list"
    - "ExampleDamerauLevenshteinOSAScore runs with `// Output:` block matching byte-for-byte"
  artifacts:
    - path: "damerau_osa.go"
      provides: "DamerauLevenshteinOSADistance, DamerauLevenshteinOSADistanceRunes, DamerauLevenshteinOSAScore, DamerauLevenshteinOSAScoreRunes"
      min_lines: 130
      contains: "// Source: Boytsov"
    - path: "dispatch_damerau_osa.go"
      provides: "Package-load-time registration of DamerauLevenshteinOSAScore into dispatch[AlgoDamerauLevenshteinOSA]"
      contains: "dispatch[AlgoDamerauLevenshteinOSA] = DamerauLevenshteinOSAScore"
    - path: "testdata/golden/_staging/damerau_osa.json"
      provides: "Per-algorithm staging golden file"
      contains: "DamerauLevenshteinOSA_ca_abc"
  key_links:
    - from: "dispatch_damerau_osa.go"
      to: "algoid.go (dispatch array)"
      via: "package-level var _ = func()bool{...}()"
      pattern: "dispatch\\[AlgoDamerauLevenshteinOSA\\]"

user_setup: []
---

<objective>
Implement Damerau-Levenshtein OSA (Optimal String Alignment) with three-row DP. DL-OSA extends Levenshtein's recurrence with a transposition rule restricted by the OSA constraint: each substring may participate in at most one transposition. The OSA variant is NOT a true metric — it can violate the triangle inequality on contrived inputs — but for the canonical inputs property tests use, the property still holds. (Property test failures, if any, would surface as test failures rather than as production bugs.)

Purpose: ship DL-OSA with the LOCKED discriminating-vector contract (`"ca"` / `"abc"` → 3) that distinguishes OSA from Full Damerau-Levenshtein (plan 02-06 returns 2 for the same vector).

Output: a working DamerauLevenshteinOSAScore zero-allocation on ASCII ≤ 64 chars (using a `[(maxStackInputLen+1)*3]int` stack buffer = 1560 bytes), with the discriminating-vector contract pinned in unit tests, golden, BDD, and the example.
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

<interfaces>
Wave 1 append-points (DO NOT recreate):

From algoid.go: AlgoDamerauLevenshteinOSA AlgoID = 1
From normalise.go: func isASCII(s string) bool
From levenshtein.go: const maxStackInputLen = 64 (REUSE — DL-OSA uses 3 rows so allocates [(maxStackInputLen+1)*3]int = 1560 bytes)
From dispatch_levenshtein.go: registration idiom (copy character-for-character)
From props_test.go: APPEND TestProp_DamerauLevenshteinOSA*; INCLUDE TriangleInequality (distance is a valid candidate; OSA is not a strict metric but the property usually holds — if testing/quick reports a counter-example, document it as a known-OSA-quirk and either add a constrained generator or skip with a citation)
From example_test.go: APPEND ExampleDamerauLevenshteinOSAScore
From algoid_test.go: UPDATE the dispatch test
From tests/bdd/steps/algorithms_steps.go: APPEND iComputeTheDamerauLevenshteinOSAScoreBetween + iComputeTheDamerauLevenshteinOSADistanceBetween + register the regexes
From testdata/golden/algorithms.json: DO NOT EDIT — write to _staging/damerau_osa.json
</interfaces>

<algorithm_specifics>
**DL-OSA recurrence (extends Levenshtein):**

  D[i,j] = min(
      D[i-1,j] + 1,                    // deletion
      D[i,j-1] + 1,                    // insertion
      D[i-1,j-1] + cost,               // substitution (cost=0 if a[i-1]==b[j-1], else 1)
      D[i-2,j-2] + 1                   // transposition (when i>=2 && j>=2 && a[i-1]==b[j-2] && a[i-2]==b[j-1])
  )

The OSA restriction (compared to Full DL): after a transposition, the affected characters cannot participate in further edits. This is enforced naturally by the three-row recurrence — D[i-2,j-2] is the score before the transposition; the transposition costs 1; no character is "re-edited" after.

**Reference vectors:**
  - "ab" / "ba" → distance 1 (single transposition)
  - "ca" / "abc" → distance 3 (OSA cannot freely re-edit after transposition; verified by Boytsov 2011 §3.1)
  - "abc" / "abc" → distance 0
  - "" / "" → distance 0
  - "" / "abc" → distance 3
  - "abc" / "" → distance 3

**Stack buffer:** `var buf [(maxStackInputLen+1)*3]int` = `[195]int` = 1560 bytes. Three rows: prevprev = buf[:n+1], prev = buf[n+1 : 2*(n+1)], curr = buf[2*(n+1) : 3*(n+1)]. After each row, swap: prevprev = prev; prev = curr; allocate new curr from buf[2*(n+1) : 3*(n+1)] OR rotate via slice swaps. The simplest correct pattern is a three-way slice swap: `prevprev, prev, curr = prev, curr, prevprev` after computing each row.

**Triangle inequality caveat:** DL-OSA is NOT a strict metric — it can violate the triangle inequality on contrived inputs. The Wave 1 Levenshtein plan's TestProp_LevenshteinDistance_TriangleInequality passes deterministically; for DL-OSA, the analogous property test MAY find a counter-example over testing/quick's random inputs. The plan executes the property test optimistically; if testing/quick reports a counter-example, the executor MUST document the failure in a comment in props_test.go citing RESEARCH.md (Boytsov 2011 notes OSA is not a metric) and either replace the test with a constrained-input variant or omit it with a citation. Do NOT silently delete the test.
</algorithm_specifics>
</context>

<tasks>

<task type="auto" tdd="true">
  <name>Task 1: Implement damerau_osa.go (three-row DP) and dispatch_damerau_osa.go</name>
  <files>damerau_osa.go, dispatch_damerau_osa.go</files>
  <read_first>
    - levenshtein.go (Wave 1 canonical pattern — two-row DP; DL-OSA extends to three rows)
    - dispatch_levenshtein.go (registration idiom)
    - normalise.go (isASCII helper, header pattern)
    - .planning/phases/02-core-character-algorithms-six/02-RESEARCH.md §Primary Sources — DL-OSA; §Implementation Patterns — DL-OSA Additional Row; §DL-OSA vs DL-Full Divergence; §Allocation Budgets — DL-OSA row
    - .planning/phases/02-core-character-algorithms-six/02-PATTERNS.md (Pattern 4 two-row DP — extend to three rows; Pattern 5 score normalisation; Pattern 6 dispatch)
    - docs/requirements.md §7.1.2 (DL-OSA spec)
    - .claude/skills/algorithm-correctness-standards/SKILL.md (primary-source citation: Boytsov 2011, with Damerau 1964 as historical source)
  </read_first>
  <action>
Create `damerau_osa.go` in package fuzzymatch:

1. Apache-2.0 header copied from normalise.go lines 1-13.
2. File-level godoc opening with both the primary and historical source citations:

       // Source: Boytsov, L. (2011). "Indexing methods for approximate dictionary
       // searching: comparative analysis." ACM Journal of Experimental Algorithmics,
       // 16, Article 1. (OSA formulation cited from §3.1.)
       //
       // Historical source: Damerau, F. J. (1964). "A technique for computer detection
       // and correction of spelling errors." Communications of the ACM, 7(3):171-176.
       // (Original transposition paper; does not name "OSA" explicitly.)

   Include the recurrence in godoc and explicitly state the OSA restriction:

       // OSA constraint: each substring participates in at most one
       // transposition. After a transposition, the affected characters
       // cannot be edited again. This makes DL-OSA NOT a strict metric —
       // triangle inequality may fail on contrived inputs. Use
       // DamerauLevenshteinFull for the metric variant.

3. Public API:
     - DamerauLevenshteinOSADistance(a, b string) int
     - DamerauLevenshteinOSADistanceRunes(a, b string) int
     - DamerauLevenshteinOSAScore(a, b string) float64
     - DamerauLevenshteinOSAScoreRunes(a, b string) float64
4. Implementation:
     - Identity fast path: `if a == b { return 0 }`.
     - Empty guards: `if m == 0 { return n }; if n == 0 { return m }`.
     - Symmetry-for-perf: swap so b is the shorter inner-loop dimension.
     - ASCII fast path: when `n <= maxStackInputLen` AND inputs are ASCII, allocate `var buf [(maxStackInputLen+1)*3]int` on the stack and pass three slices to the inner kernel.
     - Heap path: `make([]int, n+1)` × 3 for inputs > 64 OR non-ASCII.
     - Inner DP kernel `damerauOSADP(a, b string, m, n int, prevprev, prev, curr []int) int`:
         - Initialise prev[j] = j for j in [0, n].
         - For i = 1..m:
             - curr[0] = i
             - For j = 1..n:
                 - cost = 1 if a[i-1] != b[j-1] else 0
                 - curr[j] = min(prev[j]+1, curr[j-1]+1, prev[j-1]+cost)
                 - if i >= 2 && j >= 2 && a[i-1] == b[j-2] && a[i-2] == b[j-1] {
                       curr[j] = min(curr[j], prevprev[j-2]+1)
                   }
             - Three-way swap: prevprev, prev, curr = prev, curr, prevprev. After the swap, the OLD curr (now prevprev) contains stale data — it will be overwritten by the next row's curr[0]=i and curr[j]=... assignments. Verify by code review that no read of curr happens before its overwrite.
         - After the loop, the answer is in `prev[n]` (the swap means the last computed row is in `prev`).
     - Score normalisation: `maxLen := max(len(a), len(b)); if maxLen == 0 { return 1.0 }; return 1.0 - float64(dist)/float64(maxLen)`.
     - DamerauLevenshteinOSADistanceRunes / OSAScoreRunes: eager `[]rune(a)`, separate kernel operating on `[]rune`.
5. Use direct byte indexing `a[i-1]`, `b[j-1]` (NEVER `[]byte(a)`).
6. NO init(), NO `math.X` (Sqrt/Pow/etc), NO map iteration.

Create `dispatch_damerau_osa.go`:

       // [Apache-2.0 header]
       // dispatch_damerau_osa.go registers DamerauLevenshteinOSAScore into
       // the dispatch table at package load time. Sole writer to
       // dispatch[AlgoDamerauLevenshteinOSA].
       package fuzzymatch
       var _ = func() bool {
           dispatch[AlgoDamerauLevenshteinOSA] = DamerauLevenshteinOSAScore
           return true
       }()

After writing: `go build ./... && go vet ./... && bash scripts/verify-license-headers.sh`.
  </action>
  <verify>
    <automated>go build ./... && go vet ./... && bash scripts/verify-license-headers.sh</automated>
  </verify>
  <acceptance_criteria>
    - damerau_osa.go starts with Apache-2.0 header.
    - damerau_osa.go contains `// Source: Boytsov, L. (2011)` literal AND the reference to Damerau 1964 in file-level godoc.
    - damerau_osa.go contains the OSA-restriction godoc paragraph stating it is NOT a strict metric.
    - `grep -E 'math\.(Pow|Log|Exp|Sqrt|FMA)' damerau_osa.go` returns no matches.
    - `grep -E '^func init\(' damerau_osa.go dispatch_damerau_osa.go` returns no matches.
    - `grep -c '\[\]byte\(' damerau_osa.go` returns 0.
    - damerau_osa.go does NOT redeclare maxStackInputLen (it reuses the constant from levenshtein.go).
    - dispatch_damerau_osa.go contains `dispatch[AlgoDamerauLevenshteinOSA] = DamerauLevenshteinOSAScore` exactly once.
    - `go build ./...` exits 0; `go vet ./...` exits 0.
  </acceptance_criteria>
  <behavior>
    - DamerauLevenshteinOSADistance("ab", "ba") == 1
    - DamerauLevenshteinOSADistance("ca", "abc") == 3 (LOCKED discriminating vector)
    - DamerauLevenshteinOSADistance("abc", "abc") == 0
    - DamerauLevenshteinOSADistance("", "") == 0
    - DamerauLevenshteinOSADistance("", "abc") == 3
    - DamerauLevenshteinOSAScore("ab", "ba") == 0.5
    - DamerauLevenshteinOSAScore("ca", "abc") == 0.0
    - DamerauLevenshteinOSAScore("abc", "abc") == 1.0
    - DamerauLevenshteinOSAScore symmetric
    - dispatch[AlgoDamerauLevenshteinOSA] non-nil after package load
  </behavior>
  <done>
    damerau_osa.go and dispatch_damerau_osa.go committed. Package builds. Discriminating-vector behaviour verified.
  </done>
</task>

<task type="auto" tdd="true">
  <name>Task 2: Tests (unit + property + benchmark + fuzz) + extend props_test.go, example_test.go, algoid_test.go</name>
  <files>damerau_osa_test.go, damerau_osa_bench_test.go, damerau_osa_fuzz_test.go, props_test.go, example_test.go, algoid_test.go, testdata/fuzz/FuzzDamerauLevenshteinOSAScore/seed-001</files>
  <read_first>
    - levenshtein_test.go, levenshtein_bench_test.go, levenshtein_fuzz_test.go (Wave 1 templates)
    - props_test.go (existing — APPEND, do not recreate; review the TriangleInequality property pattern)
    - example_test.go, algoid_test.go (Wave 1 — extend)
    - .planning/phases/02-core-character-algorithms-six/02-RESEARCH.md §Mathematical Invariants (DL-OSA triangle inequality CAN fail; document accordingly); §Allocation Budgets — DL-OSA row
    - .planning/phases/02-core-character-algorithms-six/02-PATTERNS.md (Patterns 7, 9, 10, 11, 14)
  </read_first>
  <action>
Create `damerau_osa_test.go`:

1. Apache-2.0 header + file-level godoc.
2. Tests:
     - TestDamerauLevenshteinOSA_BothEmpty — Distance and Score on ("","") → 0 / 1.0
     - TestDamerauLevenshteinOSA_OneEmpty — ("","abc") → 3 / 0.0
     - TestDamerauLevenshteinOSA_Identical — ("abc","abc") → 0 / 1.0
     - TestDamerauLevenshteinOSA_ReferenceVectors — table-driven over ab/ba → 1 / 0.5; ca/abc → 3 / 0.0; abc/abc → 0 / 1.0; ""/abc → 3 / 0.0.
     - TestDamerauLevenshteinOSA_DiscriminatingVector — explicit assertion that DamerauLevenshteinOSADistance("ca","abc") == 3 (cite RESEARCH.md and ROADMAP success criterion #2 in the comment).
     - TestDamerauLevenshteinOSA_Symmetry — Score(a,b) == Score(b,a) for the reference vectors.
     - TestDamerauLevenshteinOSA_DistanceRunes_MultiByte — multi-byte UTF-8 input.
     - TestDamerauLevenshteinOSAScore_ZeroAllocs_ASCII_Short — `testing.AllocsPerRun(100, func() { _ = fuzzymatch.DamerauLevenshteinOSAScore("ab", "ba") })` must be 0.

Create `damerau_osa_bench_test.go`:

1. Apache-2.0 header + file-level godoc citing PERF-01 (0 allocs on ASCII ≤ 50 chars; stack buffer is 1560 bytes for n ≤ 64).
2. Benchmarks: BenchmarkDamerauLevenshteinOSAScore_ASCII_Short (ab/ba; expand if needed), _ASCII_Medium (50-char pair), _ASCII_Long (500-char heap path), _Unicode_Short.

Create `damerau_osa_fuzz_test.go`:

1. Apache-2.0 header + file-level godoc.
2. FuzzDamerauLevenshteinOSAScore — programmatic seeds: ca/abc, ab/ba, abc/abc, ""/"abc", invalid-UTF-8. Body: no panic, !math.IsNaN, !math.IsInf, score in [0,1].
3. testdata/fuzz/FuzzDamerauLevenshteinOSAScore/seed-001 with the canonical pair.

Extend `props_test.go` (APPEND):

1. TestProp_DamerauLevenshteinOSAScore_RangeBounds
2. TestProp_DamerauLevenshteinOSAScore_Identity (skip empty)
3. TestProp_DamerauLevenshteinOSAScore_Symmetric
4. TestProp_DamerauLevenshteinOSAScore_NoNaN
5. TestProp_DamerauLevenshteinOSAScore_NoInf
6. TestProp_DamerauLevenshteinOSAScore_NoNegativeZero
7. TestProp_DamerauLevenshteinOSADistance_TriangleInequality — execute optimistically. If it passes under testing/quick's default 100 random invocations, leave it. If it FAILS (testing/quick reports a counter-example), do NOT silently disable it — instead, replace the predicate with a constrained-input form (e.g. limit string lengths to 4-8 ASCII characters), OR comment-out the test with a citation: `// DL-OSA can violate triangle inequality on contrived inputs (Boytsov 2011 §3.1; this is by design — see godoc on damerau_osa.go). Property test SKIPPED; use DamerauLevenshteinFull for the metric variant.`. Document the decision in the SUMMARY.

Extend `example_test.go` (APPEND):

       func ExampleDamerauLevenshteinOSAScore() {
           fmt.Printf("%.4f\n", fuzzymatch.DamerauLevenshteinOSAScore("ab", "ba"))
           fmt.Printf("%.4f\n", fuzzymatch.DamerauLevenshteinOSAScore("ca", "abc"))
           // Output:
           // 0.5000
           // 0.0000
       }

Extend `algoid_test.go`: remove AlgoDamerauLevenshteinOSA from unregistered-slots; add registered-slot assertion.

Run:
  go test -race -shuffle=on -count=1 -run 'TestDamerauLevenshteinOSA|TestProp_DamerauLevenshteinOSA|TestDispatch|ExampleDamerauLevenshteinOSAScore' ./...
  go test -bench=BenchmarkDamerauLevenshteinOSAScore_ASCII -benchmem -run=^$ -count=3 ./...
  go test -fuzz=FuzzDamerauLevenshteinOSAScore -fuzztime=30s -run=^$ ./...
  </action>
  <verify>
    <automated>go test -race -shuffle=on -count=1 -run 'TestDamerauLevenshteinOSA|TestProp_DamerauLevenshteinOSA|TestDispatch|ExampleDamerauLevenshteinOSAScore' ./... && go test -bench=BenchmarkDamerauLevenshteinOSAScore_ASCII_Short -benchmem -run=^$ -count=3 ./... 2>&1 | grep -E '0 B/op[[:space:]]+0 allocs/op'</automated>
  </verify>
  <acceptance_criteria>
    - All TestDamerauLevenshteinOSA_* tests pass — including the explicit DiscriminatingVector test.
    - All TestProp_DamerauLevenshteinOSA* tests pass — OR the TriangleInequality test is constrained/skipped with a citation comment in props_test.go.
    - TestDamerauLevenshteinOSAScore_ZeroAllocs_ASCII_Short reports 0 allocations.
    - BenchmarkDamerauLevenshteinOSAScore_ASCII_Short reports `0 B/op  0 allocs/op`.
    - ExampleDamerauLevenshteinOSAScore output matches the two-line `0.5000\n0.0000\n` block byte-for-byte.
    - testdata/fuzz/FuzzDamerauLevenshteinOSAScore/seed-001 exists.
    - algoid_test.go: AlgoDamerauLevenshteinOSA slot non-nil.
    - `grep -c '"github.com/stretchr/testify' damerau_osa_test.go damerau_osa_bench_test.go damerau_osa_fuzz_test.go` returns 0.
  </acceptance_criteria>
  <behavior>
    - Discriminating vector ca/abc → 3 explicitly tested.
    - Property tests cover the standard six invariants plus an attempt at TriangleInequality with a documented disposition.
    - 0-alloc target on ASCII Short pinned via testing.AllocsPerRun and the benchmark.
    - Fuzz harness panic-free + invariant-preserving.
  </behavior>
  <done>
    All test files committed; props_test, example_test, algoid_test extended; full DL-OSA-scoped test suite green.
  </done>
</task>

<task type="auto" tdd="true">
  <name>Task 3: Per-algorithm staging golden file + BDD feature + extend BDD steps</name>
  <files>testdata/golden/_staging/damerau_osa.json, tests/bdd/features/damerau_osa.feature, tests/bdd/steps/algorithms_steps.go, algorithms_golden_test.go</files>
  <read_first>
    - algorithms_golden_test.go (Wave 1 — staging-write helper)
    - testdata/golden/_staging/hamming.json (Wave 2 sibling reference for staging form)
    - tests/bdd/features/levenshtein.feature (Wave 1 BDD pattern)
    - tests/bdd/steps/algorithms_steps.go (current state)
    - .planning/phases/02-core-character-algorithms-six/02-PATTERNS.md (Pattern 12, Pattern 15)
    - .planning/phases/02-core-character-algorithms-six/02-RESEARCH.md §BDD Scenario Coverage — DL-OSA discriminating-vector scenario
  </read_first>
  <action>
Create `testdata/golden/_staging/damerau_osa.json`:

1. Schema matches algorithms.json.
2. Entries (sorted by Name):
     - DamerauLevenshteinOSA_ab_ba (a "ab", b "ba", expected_score from live call: 0.5)
     - DamerauLevenshteinOSA_ca_abc (a "ca", b "abc", expected_score: 0.0)  — discriminating vector
     - DamerauLevenshteinOSA_empty_empty (a "", b "", 1.0)
     - DamerauLevenshteinOSA_identical (a "abc", b "abc", 1.0)
     - DamerauLevenshteinOSA_one_empty (a "", b "abc", 0.0)
3. Generate via algorithms_golden_test.go's staging-write helper. Add `TestGolden_DamerauLevenshteinOSA_Staging` (gated on -update). Run with `-update` once; commit. Re-run without `-update` and confirm no diff.

Create `tests/bdd/features/damerau_osa.feature`:

1. File-level comment with primary source (Boytsov 2011 / Damerau 1964).
2. Feature: Damerau-Levenshtein OSA similarity algorithm
3. Scenario "OSA discriminating reference vector" — explicit:
     # This vector proves OSA != Full DL (Full returns 2 for the same pair)
     When I compute the DamerauLevenshteinOSA distance between "ca" and "abc"
     Then the distance should be 3
4. Scenario Outline "canonical reference vectors" with rows for ab/ba → 0.5000, ca/abc → 0.0000, abc/abc → 1.0000 (tolerance 0.0001).
5. Scenario "both-empty strings score 1.0".
6. Scenario "score is symmetric".

Extend `tests/bdd/steps/algorithms_steps.go` (APPEND):

1. iComputeTheDamerauLevenshteinOSAScoreBetween(a, b string) error
2. iComputeTheDamerauLevenshteinOSADistanceBetween(a, b string) error
3. (Reuse the theDistanceShouldBe step from Hamming if present; otherwise add it.)
4. Register the regexes in InitializeScenario.

Run:
  go test -race -shuffle=on -count=1 -run 'TestGolden_DamerauLevenshteinOSA_Staging|TestGolden_Algorithms' ./...
  (cd tests/bdd && go test -race -shuffle=on -count=1 ./...)
  </action>
  <verify>
    <automated>go test -race -shuffle=on -count=1 -run 'TestGolden_DamerauLevenshteinOSA_Staging|TestGolden_Algorithms' ./... && (cd tests/bdd && go test -race -shuffle=on -count=1 ./...)</automated>
  </verify>
  <acceptance_criteria>
    - testdata/golden/_staging/damerau_osa.json exists with five entries sorted alphabetically by Name.
    - Canonical form: 2-space indent, trailing LF, no BOM.
    - Re-running staging test without -update produces no diff.
    - tests/bdd/features/damerau_osa.feature includes the "OSA discriminating reference vector" scenario explicitly.
    - `cd tests/bdd && go test -count=1 ./...` exits 0.
    - tests/bdd/steps/algorithms_steps.go still has exactly one AlgorithmContext type.
    - testdata/golden/algorithms.json UNCHANGED.
  </acceptance_criteria>
  <behavior>
    - Discriminating-vector contract pinned in golden file AND BDD scenario AND unit test (three independent gates).
    - Wave 3 plan 02-07 will diff this against the matching DL-Full staging file to verify the OSA vs Full divergence at the cross-algorithm consistency level.
  </behavior>
  <done>
    Staging golden file, DL-OSA BDD feature, and BDD step bindings committed. DL-OSA-scoped quality gate green.
  </done>
</task>

</tasks>

<threat_model>
## Trust Boundaries

| Boundary | Description |
|----------|-------------|
| Caller → fuzzymatch.DamerauLevenshteinOSAScore | Untrusted strings. Pure function. |

## STRIDE Threat Register

| Threat ID | Category | Component | Disposition | Mitigation Plan |
|-----------|----------|-----------|-------------|-----------------|
| T-02-05-01 | Denial of Service | DL-OSA O(m·n) on adversarial inputs | accept | Same as Levenshtein T-02-01-01. Document worst-case in godoc. PERF-03 ensures three-row DP, not full-table. |
| T-02-05-02 | Denial of Service | Stack overflow from `[(maxStackInputLen+1)*3]int` = 1560 bytes | mitigate | 1560 bytes is well under typical goroutine stack (8 KB initial). Heap path engages for n > 64 (linear allocation, not quadratic). |
| T-02-05-03 | Information Disclosure | Malformed UTF-8 input causing panic | mitigate | Byte-level DL-OSA operates on bytes — invalid UTF-8 cannot panic. Rune variant uses `[]rune(s)`. FuzzDamerauLevenshteinOSAScore in Task 2 covers invalid-UTF-8 seeds. |
| T-02-05-04 | Tampering | dispatch[AlgoDamerauLevenshteinOSA] overwritten | mitigate | dispatch is unexported; registration runs once at package load. Same as T-02-01-04. |
| T-02-05-05 | Repudiation | "DL-OSA is a metric" misuse leading to invalid distance reasoning | mitigate | Locked godoc paragraph in damerau_osa.go file-level godoc states OSA is NOT a strict metric, with a pointer to DamerauLevenshteinFull for the metric variant. Property test disposition (constrained or skipped TriangleInequality) is documented in props_test.go. |
| T-02-05-06 | Tampering | Discriminating-vector contract weakened (e.g. someone tweaks the recurrence and ca/abc starts returning 2) | mitigate | TestDamerauLevenshteinOSA_DiscriminatingVector + the BDD scenario + the staging golden file all gate on the same value. Three independent enforcement points; any drift triggers all three. algorithm-correctness-reviewer must approve any recurrence change. |

No high-severity items. Plan passes the security gate.
</threat_model>

<verification>
1. `go build ./...` exits 0.
2. `go vet ./...` exits 0.
3. `bash scripts/verify-license-headers.sh` exits 0.
4. `bash scripts/verify-no-runtime-deps.sh` exits 0.
5. `go test -race -shuffle=on -count=1 -run 'TestDamerauLevenshteinOSA|TestProp_DamerauLevenshteinOSA|TestDispatch|ExampleDamerauLevenshteinOSAScore|TestGolden_DamerauLevenshteinOSA_Staging' ./...` exits 0.
6. `go test -bench=BenchmarkDamerauLevenshteinOSAScore_ASCII_Short -benchmem -run=^$ -count=3 ./...` reports `0 B/op  0 allocs/op`.
7. `go test -fuzz=FuzzDamerauLevenshteinOSAScore -fuzztime=30s -run=^$ ./...` completes without crash.
8. `(cd tests/bdd && go test -race -shuffle=on -count=1 ./...)` exits 0.
9. testdata/golden/_staging/damerau_osa.json byte-stable.
10. testdata/golden/algorithms.json UNCHANGED.
11. `make check` exits 0.
</verification>

<success_criteria>
- DL-OSA discriminating vector "ca"/"abc" → 3 pinned in unit test, BDD, and staging golden file.
- Three-row DP confirmed by code review; stack-allocated `[(maxStackInputLen+1)*3]int` buffer for n ≤ 64 ASCII inputs (PERF-03 satisfied).
- Zero allocations on ASCII Short (PERF-01 satisfied for DL-OSA).
- No NaN / Inf / -0 (DET-04 satisfied for DL-OSA).
- "OSA is not a strict metric" lock documented in godoc and the property-test disposition.
- Per-algorithm staging golden file ready for Wave 3 merge.
- All required gates green via `make check`.
</success_criteria>

<output>
After completion, create `.planning/phases/02-core-character-algorithms-six/02-05-damerau-osa-SUMMARY.md` recording:
- Final identifier names confirmed (DamerauLevenshteinOSADistance, DamerauLevenshteinOSADistanceRunes, DamerauLevenshteinOSAScore, DamerauLevenshteinOSAScoreRunes).
- The TriangleInequality property-test disposition (passed / constrained / skipped) with rationale.
- Benchmark numbers observed.
- Coverage percentages.
- Any deviations from the plan.
</output>
