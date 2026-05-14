---
phase: 02-core-character-algorithms-six
plan: 06
type: execute
wave: 2
depends_on: [02-01-levenshtein]
files_modified:
  - damerau_full.go
  - dispatch_damerau_full.go
  - damerau_full_test.go
  - damerau_full_bench_test.go
  - damerau_full_fuzz_test.go
  - props_test.go
  - example_test.go
  - algoid_test.go
  - testdata/golden/_staging/damerau_full.json
  - testdata/fuzz/FuzzDamerauLevenshteinFullScore/seed-001
  - tests/bdd/features/damerau_full.feature
  - tests/bdd/steps/algorithms_steps.go
autonomous: true
requirements:
  - CHAR-03
  - PERF-01
  - PERF-02
  - PERF-03
  - TEST-01
  - TEST-02
  - TEST-04
  - TEST-05
  - DET-04
  - DX-02
tags: [damerau-levenshtein, full, lowrance-wagner, two-row-dp-plus-aux-table, true-metric]

must_haves:
  truths:
    - "DamerauLevenshteinFullDistance(\"ab\", \"ba\") returns 1 (single transposition; same as OSA)"
    - "DamerauLevenshteinFullDistance(\"ca\", \"abc\") returns 2 — DISCRIMINATING vector vs OSA (which returns 3); Score returns 1 - 2/3 ≈ 0.3333 (within 1e-9)"
    - "DamerauLevenshteinFullDistance(\"abc\", \"abc\") returns 0; Score returns 1.0 exactly"
    - "DamerauLevenshteinFullDistance(\"\", \"\") returns 0; Score returns 1.0 exactly"
    - "DamerauLevenshteinFullDistance(\"\", \"abc\") returns 3; Score returns 0.0 exactly"
    - "DamerauLevenshteinFullScoreRunes correctly handles multi-byte UTF-8 (uses map[rune]int for the last-occurrence table — heap allocation acceptable for the rune path)"
    - "dispatch[AlgoDamerauLevenshteinFull] is non-nil after package init"
    - "Apache-2.0 header on every new .go file"
    - "damerau_full.go contains `// Source: Lowrance, R., Wagner, R. A. (1975)` block at top of file-level godoc"
    - "Two-row DP + 256-int auxiliary last-occurrence array confirmed by code review (per-byte ASCII path); NO full [m+1][n+1]int table"
    - "ASCII fast path uses stack-allocated `var dpBuf [(maxStackInputLen+1)*2]int` (1040 bytes) + `var lastOcc [256]int` (2048 bytes) = ~3088 bytes total when n <= maxStackInputLen and inputs are ASCII"
    - "Score normalisation guard `if maxLen == 0 { return 1.0 }` present"
    - "No math.Pow / math.Log / math.Exp / math.Sqrt / math.FMA in damerau_full.go"
    - "No init() function; dispatch_damerau_full.go uses var _ = func() bool { ... }() registration idiom"
    - "Property tests pass for RangeBounds, Identity, Symmetric, NoNaN, NoInf, NoNegativeZero AND TriangleInequality (DL-Full IS a true metric — triangle inequality holds unconditionally)"
    - "Benchmark BenchmarkDamerauLevenshteinFullScore_ASCII_Short reports 0 B/op, 0 allocs/op (ASCII path uses stack; rune path may allocate map[rune]int)"
    - "Fuzz test FuzzDamerauLevenshteinFullScore exists with at least one seed plus invalid-UTF-8; panic-free + score in [0,1]"
    - "BDD scenarios in tests/bdd/features/damerau_full.feature exercise reference vectors AND the DL-Full discriminating contract (\"ca\"/\"abc\" → 2)"
    - "testdata/golden/_staging/damerau_full.json contains DL-Full entries (DamerauLevenshteinFull_ab_ba, _ca_abc, _identical, _empty_empty) sorted by Name"
    - "algoid_test.go updated: AlgoDamerauLevenshteinFull removed from unregistered-slots list"
    - "ExampleDamerauLevenshteinFullScore runs with `// Output:` block matching byte-for-byte; demonstrates the OSA-divergence vector"
    - "lastOcc array (or map[rune]int in the rune path) is READ-ONLY on the output path — no `for k, v := range lastOcc` iteration to produce the result (DET-03 — no map iteration)"
  artifacts:
    - path: "damerau_full.go"
      provides: "DamerauLevenshteinFullDistance, DamerauLevenshteinFullDistanceRunes, DamerauLevenshteinFullScore, DamerauLevenshteinFullScoreRunes"
      min_lines: 160
      contains: "// Source: Lowrance"
    - path: "dispatch_damerau_full.go"
      provides: "Package-load-time registration of DamerauLevenshteinFullScore into dispatch[AlgoDamerauLevenshteinFull]"
      contains: "dispatch[AlgoDamerauLevenshteinFull] = DamerauLevenshteinFullScore"
    - path: "testdata/golden/_staging/damerau_full.json"
      provides: "Per-algorithm staging golden file"
      contains: "DamerauLevenshteinFull_ca_abc"
  key_links:
    - from: "dispatch_damerau_full.go"
      to: "algoid.go (dispatch array)"
      via: "package-level var _ = func()bool{...}()"
      pattern: "dispatch\\[AlgoDamerauLevenshteinFull\\]"

user_setup: []
---

<objective>
Implement Damerau-Levenshtein Full (Lowrance-Wagner 1975) — the unrestricted-transposition variant that IS a true metric. DL-Full uses two DP rows plus a 256-int auxiliary last-occurrence array (for ASCII) or a `map[rune]int` (for the rune path). The discriminating vector `"ca"` / `"abc"` returns distance 2 here (vs 3 for DL-OSA in plan 02-05).

Purpose: ship DL-Full with the LOCKED discriminating contract that it returns 2 for `"ca"`/`"abc"` (Lowrance-Wagner 1975). The OSA vs Full divergence is a load-bearing claim of the phase — this plan owns the Full half of that contract; plan 02-05 owns the OSA half; plan 02-07 (Wave 3) owns the cross-algorithm consistency gate.

Output: a working DamerauLevenshteinFullScore that returns deterministic 0-allocation scores on ASCII ≤ 64 chars (using ~3 KB of stack: 1040 bytes for two DP rows + 2048 bytes for the 256-int auxiliary array). The rune path uses a `map[rune]int` and allocates — this is documented and acceptable per RESEARCH.md's deferred ASCII-rune-fast-path optimisation.
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

From algoid.go: AlgoDamerauLevenshteinFull AlgoID = 2
From normalise.go: func isASCII(s string) bool
From levenshtein.go: const maxStackInputLen = 64 (REUSE)
From dispatch_levenshtein.go: registration idiom (copy character-for-character)
From props_test.go: APPEND TestProp_DamerauLevenshteinFull*; INCLUDE TriangleInequality (DL-Full IS a true metric — the property holds)
From example_test.go: APPEND ExampleDamerauLevenshteinFullScore
From algoid_test.go: UPDATE the dispatch test
From tests/bdd/steps/algorithms_steps.go: APPEND iComputeTheDamerauLevenshteinFullScoreBetween + DistanceBetween + register the regexes
From testdata/golden/algorithms.json: DO NOT EDIT — write to _staging/damerau_full.json
</interfaces>

<algorithm_specifics>
**DL-Full algorithm (Lowrance-Wagner 1975):**

Maintains a `lastOccurrence` table (`da` in the original paper) mapping each character to the row index where it was last seen. Plus a per-row `last_j` tracking the column position of the last matching character. Recurrence:

  let l = da[b[j-1]]    // last row where b[j-1] appeared
  let k = last_j        // last column where a[i-1] appeared in this row's prefix scan
  D[i,j] = min(
      D[i-1,j-1] + cost,
      D[i,j-1] + 1,
      D[i-1,j] + 1,
      D[l-1,k-1] + (i - l - 1) + 1 + (j - k - 1)
  )

The "transposition cost" `D[l-1,k-1] + (i-l-1) + 1 + (j-k-1)` represents: edit-distance-up-to-the-last-occurrence plus deletions of intervening characters in `a` plus the transposition itself plus deletions of intervening characters in `b`.

For two DP rows + a `[256]int` lastOccurrence array, the full algorithm needs care: D[l-1,k-1] is OFTEN OUT OF the two-row window. The algorithm requires either three rows OR a separate auxiliary array of `D[i-1,*]` values keyed by character. The Lowrance-Wagner paper's formulation traditionally uses a full `[m+2][n+2]` table.

**For Phase 2's PERF-03 two-row constraint:** the practical implementation uses a two-row DP main table PLUS a separate `[256]int` auxiliary `lastOccPerRow` mapping character → "row of last occurrence in the second-to-last completed pass". This requires careful state management. Reference implementations (`hbollon/go-edlib` MIT, consulted for algorithm structure only — NO code copying) use this two-row + auxiliary approach.

**Practical implementation outline:**

  func damerauFullDP(a, b []byte, m, n int) int {
      var prevprev, prev, curr [maxStackInputLen+1]int
      var da [256]int  // last occurrence row index for each ASCII byte (1-indexed; 0 means never seen)

      for j := 0; j <= n; j++ { prev[j] = j }
      for i := 1; i <= m; i++ {
          curr[0] = i
          db := 0  // last column in this row where a[i-1] matched b[*]
          for j := 1; j <= n; j++ {
              k := da[b[j-1]]    // last row b[j-1] seen
              l := db            // last column a[i-1] matched (within this row)

              cost := 1
              if a[i-1] == b[j-1] {
                  cost = 0
                  db = j   // record this column as last match for a[i-1]
              }

              // Three Levenshtein-style options:
              v := prev[j] + 1                  // deletion
              if w := curr[j-1] + 1; w < v { v = w }   // insertion
              if w := prev[j-1] + cost; w < v { v = w } // sub/match

              // Transposition option (only when k>0 and l>0):
              if k > 0 && l > 0 {
                  // Need D[k-1, l-1]. Two-row DP can ONLY access D[i-1,*] (prev) and D[i,*] (curr).
                  // For Full DL with two-row, we need to ADDITIONALLY track the value at D[k-1, l-1]
                  // for arbitrary k. This typically requires either:
                  //   (a) Promoting to a full DP table (defeats PERF-03), or
                  //   (b) Maintaining an auxiliary array dPrevForChar[256] storing D[k-1, l-1] for each character
                  //       — this is the practical approach for ASCII; one int per character byte tracking the
                  //       current "anchor" cost for that character's last occurrence.

                  // Implementation note: the cleanest two-row-equivalent form for DL-Full requires
                  // a dedicated auxiliary `H` table (per Lowrance-Wagner 1975) of size O(|Σ|) where
                  // H[c] = D[lastOccRow(c), lastOccCol(c)]. This is updated on the fly as characters
                  // are observed. For the ASCII fast path, H is a [256]int stack array.

                  trans := H[a[i-1]] + (i - k - 1) + 1 + (j - l - 1)
                  if trans < v { v = trans }
              }
              curr[j] = v
          }
          // Update H[c] for the character we just processed: H[a[i-1]] = curr[db-1] if db > 0
          // ... (exact bookkeeping per Lowrance-Wagner formulation)
          da[a[i-1]] = i  // record this row as last occurrence of a[i-1]
          // Two-way swap: prev, curr = curr, prev
          prev, curr = curr, prev
      }
      return prev[n]
  }

The above sketch is INDICATIVE. The executor MUST cross-validate the implementation against published Lowrance-Wagner 1975 reference vectors AND the discriminating vector ca/abc → 2. If the practical two-row + aux-table approach proves too complex to derive from first principles within the plan's context budget, the executor MAY fall back to a full `[m+2][n+2]int` DP table that is heap-allocated for ALL inputs (sacrificing PERF-03 for DL-Full only — note this in the SUMMARY as a deviation requiring follow-up). The discriminating vector contract takes precedence over the allocation budget.

**Reference vectors (LOCKED):**
  - "ab" / "ba" → distance 1 (same as OSA)
  - "ca" / "abc" → distance 2 (DISCRIMINATING vs OSA which returns 3)
  - "abc" / "abc" → distance 0
  - "" / "" → distance 0
  - "" / "abc" → distance 3

**Stack budget (ASCII path, n ≤ 64):**
  - Two DP rows: `var prev, curr [maxStackInputLen+1]int` = 2 × 65 × 8 = 1040 bytes
  - Auxiliary last-occurrence: `var da [256]int` = 2048 bytes
  - Auxiliary anchor: `var H [256]int` = 2048 bytes (if the implementation uses the H-table approach)
  - Total: ~5 KB on the stack — within typical goroutine stack budget
  Heap path engages for n > 64 OR non-ASCII.

**Rune path:** uses `map[rune]int` for da and H (heap allocation; documented). The map is QUERIED via `da[r]` — never iterated to produce output (DET-03).
</algorithm_specifics>
</context>

<tasks>

<task type="auto" tdd="true">
  <name>Task 1: Implement damerau_full.go (two-row DP + auxiliary tables) and dispatch_damerau_full.go</name>
  <files>damerau_full.go, dispatch_damerau_full.go</files>
  <read_first>
    - levenshtein.go (Wave 1 canonical pattern — two-row DP)
    - dispatch_levenshtein.go (registration idiom)
    - normalise.go (isASCII helper, header pattern)
    - .planning/phases/02-core-character-algorithms-six/02-RESEARCH.md §Primary Sources — DL-Full; §Implementation Patterns — DL-Full Auxiliary State; §DL-OSA vs DL-Full Divergence; §Allocation Budgets — DL-Full row
    - .planning/phases/02-core-character-algorithms-six/02-PATTERNS.md (Pattern 4 — extend with auxiliary tables; Pattern 5 score normalisation; Pattern 6 dispatch)
    - docs/requirements.md §7.1.3 (DL-Full spec)
    - .claude/skills/algorithm-correctness-standards/SKILL.md (primary-source citation: Lowrance-Wagner 1975)
    - .claude/skills/algorithm-licensing-standards/SKILL.md (cross-validation against MIT-licensed reference implementations is permitted; CODE COPYING IS NOT — implement fresh from the paper)
  </read_first>
  <action>
Create `damerau_full.go` in package fuzzymatch:

1. Apache-2.0 header copied from normalise.go lines 1-13.
2. File-level godoc opening with:

       // Source: Lowrance, R., Wagner, R. A. (1975). "An extension of the
       // string-to-string correction problem." Journal of the ACM, 22(2):177-183.
       //
       // DL-Full (the Lowrance-Wagner formulation) is the TRUE metric variant
       // of Damerau-Levenshtein distance. It permits unrestricted transpositions:
       // any pair of adjacent characters may be transposed, and the characters
       // may subsequently be edited. Compare with DamerauLevenshteinOSA (the
       // Optimal String Alignment variant), which restricts each substring to
       // at most one transposition and is NOT a metric.
       //
       // Triangle inequality holds for DL-Full unconditionally. Use this
       // variant when correctness > speed and metric properties matter.

   Include the recurrence in godoc. Note the algorithm uses two DP rows + auxiliary last-occurrence and anchor tables (or describe whichever variant is implemented).

3. Public API:
     - DamerauLevenshteinFullDistance(a, b string) int
     - DamerauLevenshteinFullDistanceRunes(a, b string) int
     - DamerauLevenshteinFullScore(a, b string) float64
     - DamerauLevenshteinFullScoreRunes(a, b string) float64

4. Implementation:
     - Identity, both-empty, one-empty, swap-for-perf — same edge-case treatment as Levenshtein.
     - ASCII fast path (n ≤ maxStackInputLen AND both inputs ASCII): stack-allocate two DP rows AS WELL AS the auxiliary `[256]int` tables required for DL-Full. Pass slices into an inner kernel.
     - Heap path: `make([]int, n+1)` × 2 + the auxiliary tables (sized appropriately for the input character set).
     - Inner kernel: implement the Lowrance-Wagner recurrence using two-row DP + auxiliary anchor table per the algorithm sketch in this plan's §algorithm_specifics. Cross-validate against the discriminating vector "ca"/"abc" → 2 BEFORE proceeding to other tests.
     - If the two-row + auxiliary-anchor formulation proves too complex to derive correctly within the plan's context budget, the executor MAY fall back to a full `[m+2][n+2]int` DP table that is heap-allocated for ALL inputs (sacrificing PERF-03 for DL-Full only). Document this fallback in the SUMMARY as a deviation requiring v1.x follow-up. The discriminating-vector contract is non-negotiable; the two-row optimisation can be deferred. Do NOT silently ship a wrong recurrence to satisfy PERF-03.
     - Score normalisation: `maxLen := max(len(a), len(b)); if maxLen == 0 { return 1.0 }; return 1.0 - float64(dist)/float64(maxLen)`.
     - Rune variants: eager `[]rune(a)`; auxiliary tables become `map[rune]int` or `map[rune]int{}` (heap allocation acceptable; document in godoc).

5. Use direct byte indexing (NEVER `[]byte(a)`).
6. NO init(), NO `math.X`, NO map ITERATION on output paths (map LOOKUP is fine for the rune path's auxiliary tables).
7. Add a comment in the implementation explicitly noting where map LOOKUPS happen and why they are OK (DET-03 forbids iteration, not lookup).

Create `dispatch_damerau_full.go`:

       // [Apache-2.0 header]
       // dispatch_damerau_full.go registers DamerauLevenshteinFullScore into
       // the dispatch table at package load time.
       package fuzzymatch
       var _ = func() bool {
           dispatch[AlgoDamerauLevenshteinFull] = DamerauLevenshteinFullScore
           return true
       }()

After writing: `go build ./... && go vet ./... && bash scripts/verify-license-headers.sh`. AFTER the build is green, run a quick sanity check via `go test -run TestDamerauLevenshteinFull_DiscriminatingVector -count=1 ./...` (test created in task 2) — if the discriminating vector returns 3 (matching OSA) instead of 2, the recurrence is WRONG and must be fixed before proceeding.
  </action>
  <verify>
    <automated>go build ./... && go vet ./... && bash scripts/verify-license-headers.sh</automated>
  </verify>
  <acceptance_criteria>
    - damerau_full.go starts with Apache-2.0 header.
    - damerau_full.go contains `// Source: Lowrance, R., Wagner, R. A. (1975)` literal in file-level godoc.
    - damerau_full.go contains the godoc paragraph stating DL-Full IS the true metric variant and explicitly contrasts with DL-OSA.
    - `grep -E 'math\.(Pow|Log|Exp|Sqrt|FMA)' damerau_full.go` returns no matches.
    - `grep -E '^func init\(' damerau_full.go dispatch_damerau_full.go` returns no matches.
    - `grep -c '\[\]byte\(' damerau_full.go` returns 0 (no allocating byte conversions on the byte hot path).
    - `grep -E 'for\b.*range\b.*lastOcc|for\b.*range\b.*da\b|for\b.*range\b.*H\b' damerau_full.go` returns no matches (no map iteration on auxiliary tables).
    - damerau_full.go does NOT redeclare maxStackInputLen.
    - dispatch_damerau_full.go contains `dispatch[AlgoDamerauLevenshteinFull] = DamerauLevenshteinFullScore` exactly once.
    - `go build ./...` exits 0; `go vet ./...` exits 0.
  </acceptance_criteria>
  <behavior>
    - DamerauLevenshteinFullDistance("ab", "ba") == 1
    - DamerauLevenshteinFullDistance("ca", "abc") == 2 (LOCKED discriminating vector — proves Full != OSA)
    - DamerauLevenshteinFullDistance("abc", "abc") == 0
    - DamerauLevenshteinFullDistance("", "") == 0
    - DamerauLevenshteinFullDistance("", "abc") == 3
    - DamerauLevenshteinFullScore("ca", "abc") within 1e-9 of 0.3333… (= 1 - 2/3)
    - DamerauLevenshteinFullScore("abc", "abc") == 1.0 exactly
    - DamerauLevenshteinFullScore symmetric
    - dispatch[AlgoDamerauLevenshteinFull] non-nil after package load
  </behavior>
  <done>
    damerau_full.go and dispatch_damerau_full.go committed. Package builds. Discriminating-vector behaviour verified (Distance("ca","abc") == 2 — NOT 3).
  </done>
</task>

<task type="auto" tdd="true">
  <name>Task 2: Tests (unit + property + benchmark + fuzz) + extend props_test.go, example_test.go, algoid_test.go</name>
  <files>damerau_full_test.go, damerau_full_bench_test.go, damerau_full_fuzz_test.go, props_test.go, example_test.go, algoid_test.go, testdata/fuzz/FuzzDamerauLevenshteinFullScore/seed-001</files>
  <read_first>
    - levenshtein_test.go, levenshtein_bench_test.go, levenshtein_fuzz_test.go (Wave 1 templates)
    - damerau_osa_test.go (sibling Wave 2 — same test shape; copy structure, swap identifiers; DL-Full has DIFFERENT expected value for the discriminating vector)
    - props_test.go (existing — APPEND, do not recreate; this plan's TriangleInequality property MUST pass — DL-Full is a metric)
    - example_test.go, algoid_test.go (Wave 1 — extend)
    - .planning/phases/02-core-character-algorithms-six/02-RESEARCH.md §Mathematical Invariants (DL-Full triangle inequality holds unconditionally); §Allocation Budgets — DL-Full row
    - .planning/phases/02-core-character-algorithms-six/02-PATTERNS.md (Patterns 7, 9, 10, 11, 14)
  </read_first>
  <action>
Create `damerau_full_test.go`:

1. Apache-2.0 header + file-level godoc.
2. Tests:
     - TestDamerauLevenshteinFull_BothEmpty — 0 / 1.0
     - TestDamerauLevenshteinFull_OneEmpty — ("","abc") → 3 / 0.0
     - TestDamerauLevenshteinFull_Identical — ("abc","abc") → 0 / 1.0
     - TestDamerauLevenshteinFull_ReferenceVectors — table-driven over ab/ba → 1 / 0.5; ca/abc → 2 / 0.3333…; abc/abc → 0 / 1.0; ""/"abc" → 3 / 0.0. (Float tolerance 1e-9.)
     - TestDamerauLevenshteinFull_DiscriminatingVector — explicit assertion that DamerauLevenshteinFullDistance("ca","abc") == 2 — REFERENCE the OSA-vs-Full divergence in the comment with a citation to RESEARCH.md §DL-OSA vs DL-Full Divergence and ROADMAP success criterion #2.
     - TestDamerauLevenshteinFull_Symmetry — Score(a,b) == Score(b,a) on the reference vectors.
     - TestDamerauLevenshteinFull_DistanceRunes_MultiByte — multi-byte input.
     - TestDamerauLevenshteinFullScore_ZeroAllocs_ASCII_Short — `testing.AllocsPerRun(100, func() { _ = fuzzymatch.DamerauLevenshteinFullScore("ab", "ba") })` must be 0 IF the implementation uses the two-row + aux-table approach. If the executor fell back to the full-table heap approach (per Task 1's deviation clause), this test should be REPLACED with a documented `t.Skipf("DL-Full uses heap path for v1.0; see SUMMARY for v1.x optimisation plan")` and the SUMMARY must note the fallback.

Create `damerau_full_bench_test.go`:

1. Apache-2.0 header + file-level godoc citing PERF-01 (target 0 allocs on ASCII; falls back per Task 1's deviation clause if needed).
2. Benchmarks: BenchmarkDamerauLevenshteinFullScore_ASCII_Short/Medium/Long, _Unicode_Short. Use b.ReportAllocs/ResetTimer + var-sink pattern.

Create `damerau_full_fuzz_test.go`:

1. Apache-2.0 header + file-level godoc.
2. FuzzDamerauLevenshteinFullScore — programmatic seeds including the discriminating vector + invalid-UTF-8. Body: no panic, !math.IsNaN, !math.IsInf, score in [0,1].
3. testdata/fuzz/FuzzDamerauLevenshteinFullScore/seed-001 with the canonical pair.

Extend `props_test.go` (APPEND):

1. TestProp_DamerauLevenshteinFullScore_RangeBounds
2. TestProp_DamerauLevenshteinFullScore_Identity (skip empty)
3. TestProp_DamerauLevenshteinFullScore_Symmetric
4. TestProp_DamerauLevenshteinFullScore_NoNaN
5. TestProp_DamerauLevenshteinFullScore_NoInf
6. TestProp_DamerauLevenshteinFullScore_NoNegativeZero
7. TestProp_DamerauLevenshteinFullDistance_TriangleInequality — MUST PASS (DL-Full is a true metric per Lowrance-Wagner 1975; if testing/quick reports a counter-example, the implementation is incorrect — fix the recurrence rather than disabling the test).

Extend `example_test.go` (APPEND):

       func ExampleDamerauLevenshteinFullScore() {
           // The "ca"/"abc" pair demonstrates DL-Full's divergence from DL-OSA:
           // DL-OSA returns 0.0 (distance 3); DL-Full returns 0.3333 (distance 2).
           fmt.Printf("%.4f\n", fuzzymatch.DamerauLevenshteinFullScore("ca", "abc"))
           fmt.Printf("%.4f\n", fuzzymatch.DamerauLevenshteinFullScore("ab", "ba"))
           // Output:
           // 0.3333
           // 0.5000
       }

Extend `algoid_test.go`: remove AlgoDamerauLevenshteinFull from unregistered-slots; add registered-slot assertion.

Run:
  go test -race -shuffle=on -count=1 -run 'TestDamerauLevenshteinFull|TestProp_DamerauLevenshteinFull|TestDispatch|ExampleDamerauLevenshteinFullScore' ./...
  go test -bench=BenchmarkDamerauLevenshteinFullScore_ASCII -benchmem -run=^$ -count=3 ./...
  go test -fuzz=FuzzDamerauLevenshteinFullScore -fuzztime=30s -run=^$ ./...
  </action>
  <verify>
    <automated>go test -race -shuffle=on -count=1 -run 'TestDamerauLevenshteinFull|TestProp_DamerauLevenshteinFull|TestDispatch|ExampleDamerauLevenshteinFullScore' ./...</automated>
  </verify>
  <acceptance_criteria>
    - All TestDamerauLevenshteinFull_* tests pass — including the explicit DiscriminatingVector test confirming distance == 2 (NOT 3).
    - All TestProp_DamerauLevenshteinFull* tests pass, INCLUDING TriangleInequality (DL-Full is a metric).
    - TestDamerauLevenshteinFullScore_ZeroAllocs_ASCII_Short passes OR is skipped with a documented fallback (see Task 1).
    - ExampleDamerauLevenshteinFullScore output matches the two-line `0.3333\n0.5000\n` block byte-for-byte.
    - testdata/fuzz/FuzzDamerauLevenshteinFullScore/seed-001 exists.
    - algoid_test.go: AlgoDamerauLevenshteinFull slot non-nil.
    - `grep -c '"github.com/stretchr/testify' damerau_full_test.go damerau_full_bench_test.go damerau_full_fuzz_test.go` returns 0.
  </acceptance_criteria>
  <behavior>
    - Discriminating vector ca/abc → 2 explicitly tested, distinct from DL-OSA's value of 3.
    - All seven property tests pass — DL-Full is a true metric.
    - Benchmark runs without panic; allocations either 0 (two-row path) or documented (fallback path).
    - Fuzz harness panic-free.
    - Example demonstrates DL-Full's signature divergence from DL-OSA.
  </behavior>
  <done>
    All test files committed; props_test, example_test, algoid_test extended; full DL-Full-scoped test suite green.
  </done>
</task>

<task type="auto" tdd="true">
  <name>Task 3: Per-algorithm staging golden file + BDD feature + extend BDD steps</name>
  <files>testdata/golden/_staging/damerau_full.json, tests/bdd/features/damerau_full.feature, tests/bdd/steps/algorithms_steps.go, algorithms_golden_test.go</files>
  <read_first>
    - algorithms_golden_test.go (Wave 1 — staging-write helper)
    - testdata/golden/_staging/damerau_osa.json (sibling Wave 2 plan output, if landed first; use as reference for staging form AND for cross-checking the OSA vs Full divergence)
    - tests/bdd/features/damerau_osa.feature (sibling pattern; this plan adds a parallel feature for Full)
    - tests/bdd/steps/algorithms_steps.go (current state)
    - .planning/phases/02-core-character-algorithms-six/02-PATTERNS.md (Pattern 12, Pattern 15)
    - .planning/phases/02-core-character-algorithms-six/02-RESEARCH.md §BDD Scenario Coverage — DL-Full discriminating-vector scenario
  </read_first>
  <action>
Create `testdata/golden/_staging/damerau_full.json`:

1. Schema matches algorithms.json.
2. Entries (sorted by Name):
     - DamerauLevenshteinFull_ab_ba (a "ab", b "ba", expected_score from live call: 0.5)
     - DamerauLevenshteinFull_ca_abc (a "ca", b "abc", expected_score from live call: 0.3333…)  — discriminating vector
     - DamerauLevenshteinFull_empty_empty (a "", b "", 1.0)
     - DamerauLevenshteinFull_identical (a "abc", b "abc", 1.0)
     - DamerauLevenshteinFull_one_empty (a "", b "abc", 0.0)
3. Generate via algorithms_golden_test.go's staging-write helper. Add `TestGolden_DamerauLevenshteinFull_Staging` (gated on -update). Run with `-update` once; commit. Re-run without `-update` and confirm no diff.

Create `tests/bdd/features/damerau_full.feature`:

1. File-level comment with primary source (Lowrance-Wagner 1975).
2. Feature: Damerau-Levenshtein Full (Lowrance-Wagner) similarity algorithm
3. Scenario "Full DL discriminating reference vector — diverges from OSA":
     # This vector proves Full DL != OSA. OSA returns 3 for the same pair.
     When I compute the DamerauLevenshteinFull distance between "ca" and "abc"
     Then the distance should be 2
4. Scenario Outline "canonical reference vectors" with rows for ab/ba → 0.5000, ca/abc → 0.3333, abc/abc → 1.0000 (tolerance 0.0001).
5. Scenario "Full DL is a true metric (triangle inequality holds)" — comment scenario noting the contrast with OSA. The actual triangle-inequality property is enforced via testing/quick in props_test.go; this BDD scenario serves as documentation.
6. Scenario "both-empty strings score 1.0".
7. Scenario "score is symmetric".

Extend `tests/bdd/steps/algorithms_steps.go` (APPEND):

1. iComputeTheDamerauLevenshteinFullScoreBetween(a, b string) error
2. iComputeTheDamerauLevenshteinFullDistanceBetween(a, b string) error
3. (Reuse theDistanceShouldBe step from sibling plans if present.)
4. Register the regexes in InitializeScenario.

Run:
  go test -race -shuffle=on -count=1 -run 'TestGolden_DamerauLevenshteinFull_Staging|TestGolden_Algorithms' ./...
  (cd tests/bdd && go test -race -shuffle=on -count=1 ./...)
  </action>
  <verify>
    <automated>go test -race -shuffle=on -count=1 -run 'TestGolden_DamerauLevenshteinFull_Staging|TestGolden_Algorithms' ./... && (cd tests/bdd && go test -race -shuffle=on -count=1 ./...)</automated>
  </verify>
  <acceptance_criteria>
    - testdata/golden/_staging/damerau_full.json exists with five entries sorted alphabetically by Name; the ca/abc entry has expected_score ≈ 0.3333… (not 0.0 — that would match the OSA file).
    - Canonical form: 2-space indent, trailing LF, no BOM.
    - Re-running staging test without -update produces no diff.
    - tests/bdd/features/damerau_full.feature includes the "Full DL discriminating reference vector — diverges from OSA" scenario explicitly.
    - `cd tests/bdd && go test -count=1 ./...` exits 0.
    - tests/bdd/steps/algorithms_steps.go still has exactly one AlgorithmContext type.
    - testdata/golden/algorithms.json UNCHANGED.
    - Sibling staging file _staging/damerau_osa.json is UNCHANGED (independent plan should not modify it).
  </acceptance_criteria>
  <behavior>
    - Discriminating-vector contract pinned in golden file AND BDD scenario AND unit test (three independent gates), with the OPPOSITE value to plan 02-05's DL-OSA staging file.
    - Wave 3 plan 02-07 will diff _staging/damerau_osa.json vs _staging/damerau_full.json on the ca/abc entry to verify the divergence at the cross-algorithm consistency level.
  </behavior>
  <done>
    Staging golden file, DL-Full BDD feature, and BDD step bindings committed. DL-Full-scoped quality gate green.
  </done>
</task>

</tasks>

<threat_model>
## Trust Boundaries

| Boundary | Description |
|----------|-------------|
| Caller → fuzzymatch.DamerauLevenshteinFullScore | Untrusted strings. Pure function. |

## STRIDE Threat Register

| Threat ID | Category | Component | Disposition | Mitigation Plan |
|-----------|----------|-----------|-------------|-----------------|
| T-02-06-01 | Denial of Service | DL-Full O(m·n) on adversarial inputs | accept | Same as Levenshtein. PERF-03 ensures no full-table allocation IF the two-row + aux-table implementation is shipped; if the fallback heap-table approach is used, allocation grows O(m·n) which is documented in the SUMMARY as a v1.x follow-up. |
| T-02-06-02 | Denial of Service | Stack overflow from ~5 KB stack allocation in ASCII fast path | mitigate | 5 KB is well under typical goroutine stack budget (8 KB initial, growing). Heap path engages for n > 64. |
| T-02-06-03 | Information Disclosure | Malformed UTF-8 input causing panic | mitigate | Byte-level DL-Full operates on bytes. Rune variant uses `[]rune(s)` + `map[rune]int` (Go's stdlib safely handles invalid UTF-8 via U+FFFD). FuzzDamerauLevenshteinFullScore covers invalid-UTF-8 seeds. |
| T-02-06-04 | Tampering | dispatch[AlgoDamerauLevenshteinFull] overwritten | mitigate | dispatch is unexported; registration runs once. Same as T-02-01-04. |
| T-02-06-05 | Tampering | Discriminating-vector contract weakened (e.g. recurrence drift makes ca/abc return 3, collapsing into OSA behaviour) | mitigate | TestDamerauLevenshteinFull_DiscriminatingVector + the BDD scenario + the staging golden file gate on the value 2 (not 3). Three independent enforcement points; algorithm-correctness-reviewer reviews any recurrence change. |
| T-02-06-06 | Repudiation | "DL-Full is just like OSA" misconception | mitigate | godoc explicitly contrasts with DL-OSA; ExampleDamerauLevenshteinFullScore demonstrates the divergence on pkg.go.dev; props_test.go's TriangleInequality property tests pass for Full but may fail for OSA — the documented mathematical distinction. |
| T-02-06-07 | Information Disclosure | Map iteration on rune-path auxiliary table leaking non-deterministic order | mitigate | Map is QUERIED via `da[r]` only, never iterated to produce output. Code review verifies no `for k, v := range da` or similar. DET-03 satisfied. The grep gate `grep -E 'for\b.*range\b.*lastOcc|for\b.*range\b.*da\b|for\b.*range\b.*H\b'` returns no matches per Task 1 acceptance criteria. |

No high-severity items. Plan passes the security gate.
</threat_model>

<verification>
1. `go build ./...` exits 0.
2. `go vet ./...` exits 0.
3. `bash scripts/verify-license-headers.sh` exits 0.
4. `bash scripts/verify-no-runtime-deps.sh` exits 0.
5. `go test -race -shuffle=on -count=1 -run 'TestDamerauLevenshteinFull|TestProp_DamerauLevenshteinFull|TestDispatch|ExampleDamerauLevenshteinFullScore|TestGolden_DamerauLevenshteinFull_Staging' ./...` exits 0.
6. `go test -bench=BenchmarkDamerauLevenshteinFullScore_ASCII_Short -benchmem -run=^$ -count=3 ./...` runs without panic; allocations either 0 (preferred) or documented in SUMMARY.
7. `go test -fuzz=FuzzDamerauLevenshteinFullScore -fuzztime=30s -run=^$ ./...` completes without crash.
8. `(cd tests/bdd && go test -race -shuffle=on -count=1 ./...)` exits 0.
9. testdata/golden/_staging/damerau_full.json byte-stable.
10. testdata/golden/algorithms.json UNCHANGED; _staging/damerau_osa.json UNCHANGED relative to plan 02-05's commit.
11. `make check` exits 0.
</verification>

<success_criteria>
- DL-Full discriminating vector "ca"/"abc" → 2 pinned in unit test, BDD, and staging golden file (the OPPOSITE value to DL-OSA's plan 02-05 staging file).
- Triangle inequality property holds for DL-Full (true metric per Lowrance-Wagner 1975).
- Two-row DP + auxiliary tables ideally satisfy PERF-03 + PERF-01; if a fallback heap-table approach was needed, deviation documented in SUMMARY for v1.x optimisation.
- No NaN / Inf / -0 (DET-04 satisfied for DL-Full).
- No map iteration on output paths (DET-03 satisfied; map LOOKUP only on the rune-path auxiliary tables).
- Per-algorithm staging golden file ready for Wave 3 merge.
- All required gates green via `make check`.
</success_criteria>

<output>
After completion, create `.planning/phases/02-core-character-algorithms-six/02-06-damerau-full-SUMMARY.md` recording:
- Final identifier names confirmed.
- Whether the implementation uses two-row + auxiliary tables (preferred) OR a heap-allocated full DP table (fallback) — this is the load-bearing deviation from PERF-03 if any.
- Allocation profile observed (Short/Medium/Long ASCII; rune path).
- Confirmation that TestDamerauLevenshteinFull_DiscriminatingVector and the property tests (especially TriangleInequality) pass.
- Cross-check that _staging/damerau_full.json's ca/abc entry is 0.3333… while _staging/damerau_osa.json's ca/abc entry is 0.0 — the divergence is observable in the staging golden files.
- Coverage percentages.
- Any deviations from the plan and their rationale.
</output>
