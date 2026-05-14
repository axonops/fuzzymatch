---
phase: 04-remaining-character-gestalt
plan: 02
type: execute
wave: 2
depends_on:
  - 04-01-strcmp95
files_modified:
  - lcsstr.go
  - dispatch_lcsstr.go
  - lcsstr_test.go
  - lcsstr_bench_test.go
  - lcsstr_fuzz_test.go
  - props_test.go
  - example_test.go
  - algoid_test.go
  - algorithms_golden_test.go
  - testdata/golden/_staging/lcsstr.json
  - testdata/fuzz/FuzzLCSStrScore/seed-001
  - tests/bdd/features/lcsstr.feature
  - tests/bdd/steps/algorithms_steps.go
autonomous: true
requirements:
  - CHAR-09
tags: [lcsstr, longest-common-substring, wagner-fischer-1974, two-row-dp, ascii-fast-path, leftmost-tie-break, sorensen-dice-normalisation, four-public-functions, dispatch-registration, property-tests, fuzz, benchmark, bdd, staging-golden]

must_haves:
  truths:
    # Goal-backward truths (CHAR-09; ROADMAP §"Remaining Character & Gestalt" success criteria #2)
    - "A caller can `import fuzzymatch` and call LCSStrScore(\"http_request\", \"http_request_header_fields\") and receive a deterministic float64 in [0.0, 1.0] computed as 2·len(lcs)/(len(a)+len(b)) (SPEC-PINNED at docs/requirements.md §7.1.9)"
    - "LongestCommonSubstring(\"abcXYZabc\", \"abc\") == \"abc\" — the LEFTMOST occurrence in `a` wins (CONTEXT.md §3 LOCKED tie-break; load-bearing for RESEARCH.md Pitfall 4)"
    - "LongestCommonSubstring(\"\", \"\") == \"\" AND LCSStrScore(\"\", \"\") == 1.0 (both-empty convention)"
    - "LongestCommonSubstring(\"\", \"abc\") == \"\" AND LCSStrScore(\"\", \"abc\") == 0.0 (one-empty)"
    - "LongestCommonSubstring(\"abc\", \"xyz\") == \"\" AND LCSStrScore(\"abc\", \"xyz\") == 0.0 (no overlap — the empty-string return is documented behaviour, NOT a bug; RESEARCH.md Pitfall 6)"
    - "LongestCommonSubstring(\"abc\", \"abc\") == \"abc\" AND LCSStrScore(\"abc\", \"abc\") == 1.0 (identity short-circuit + the 2·n/(n+n)=1 formula)"
    - "LCSStrScore(a, b) == LCSStrScore(b, a) for any (a, b) — symmetric on byte and rune paths"
    - "LongestCommonSubstringRunes(\"café\", \"cafe\") == \"caf\" AND LCSStrScoreRunes(\"café\", \"cafe\") ≈ 0.75 (rune path handles multi-byte UTF-8)"
    - "dispatch[AlgoLCSStr] (slot 8) is non-nil after package load and equals LCSStrScore (registered via `var _ = func() bool {...}()` in dispatch_lcsstr.go; NO init())"
    # Source-origin + DP discipline (CONTEXT.md §3; algorithm-correctness-standards; performance-standards)
    - "lcsstr.go's file-level godoc cites Wagner-Fischer 1974 as the PRIMARY source, includes the recurrence `D[i,j] = D[i-1,j-1] + 1 if a[i-1] == b[j-1] else 0` inline, states the Sørensen-Dice score normalisation, and documents the LEFTMOST-in-`a` tie-break with reference to the strict-`>` max-update"
    - "The DP kernel uses STRICT-GREATER-THAN (`>`, not `>=`) in the max-update — first-found-leftmost wins (RESEARCH.md Pitfall 4 load-bearing)"
    - "Two-row rolling buffer DP confirmed by code review: inner kernel maintains exactly two []int slices of length n+1; no [m+1][n+1]int table allocated anywhere in lcsstr.go"
    - "ASCII fast path on (n <= maxStackInputLen && isASCII(a) && isASCII(b)) allocates the two rolling rows on the stack via `var buf [(maxStackInputLen+1)*2]int`; BenchmarkLCSStrScore_ASCII_Short reports 0 B/op, 0 allocs/op"
    - "lcsstr.go contains NO init() function and NO map iteration on output paths"
    - "maxStackInputLen is REFERENCED (declared in levenshtein.go) — NEVER redeclared in lcsstr.go; isASCII (normalise.go:159-168) is referenced by name"
    # Public surface (CONTEXT.md §3 SPEC-PINNED at docs/requirements.md §7.1.9)
    - "Public surface: exactly four new exported functions — LongestCommonSubstring(a, b string) string, LongestCommonSubstringRunes(a, b string) string, LCSStrScore(a, b string) float64, LCSStrScoreRunes(a, b string) float64. Pre-existing AlgoLCSStr constant already in algoid.go slot 8"
    - "Only LCSStrScore is dispatched (dispatch table maps AlgoID → (a, b string) float64 only); LongestCommonSubstring*, LCSStrScoreRunes are public but not dispatched (PATTERNS.md §dispatch_lcsstr.go gotcha)"
    - "Identity short-circuit `if a == b { return 1.0 }` present on LCSStrScoreRunes BEFORE `[]rune(a)` allocation (IN-04 closure); LongestCommonSubstringRunes short-circuits to `a` (or empty for both-empty) similarly"
    # Determinism + performance (DET-04, DET-06, PERF-01, PERF-02, PERF-03)
    - "LCSStrScore / LCSStrScoreRunes never return NaN, +Inf, -Inf, or -0 — verified by TestProp_LCSStr*_{NoNaN,NoInf,NoNegativeZero}"
    - "No math.Pow / math.Log / math.Exp / math.FMA used anywhere in lcsstr.go — only `+`, `-`, `*`, `/`, comparisons; score normalisation uses explicit left-to-right parenthesisation per DET-06"
    - "Allocation budget: ASCII Short → 0 allocs (stack buffer); ASCII Long → ≤ 2 allocs (the two rolling rows via `make([]int, n+1)`); Rune path → ≤ 4 allocs (`[]rune(a)`, `[]rune(b)`, two rolling rows)"
    - "TestProp_LongestCommonSubstring_IsSubstringOfBoth: testing/quick → strings.Contains(a, got) && strings.Contains(b, got) for the returned substring"
    - "TestProp_LongestCommonSubstring_LengthMatchesScore: testing/quick → `2·len(LongestCommonSubstring(a,b))/(len(a)+len(b)) ≈ LCSStrScore(a, b)` within 1e-9"
    - "TestProp_LongestCommonSubstring_LeftmostTieBreak: hand-curated tied-candidate inputs → leftmost-in-`a` wins"
    # Meta-test + fuzz + BDD discipline (Phase 2/3 inherited)
    - "FuzzLCSStrScore exercises ALL FOUR public surfaces (Phase 3 WR-02 closure): LCSStrScore + LCSStrScoreRunes + LongestCommonSubstring + LongestCommonSubstringRunes — body asserts no panic on any surface, scores in [0,1], no NaN/Inf"
    - "testdata/fuzz/FuzzLCSStrScore/seed-001 exists in `go test fuzz v1` literal corpus format with seed (a, b) covering identity, both-empty, one-empty, leftmost-tie-break (abcXYZabc/abc), no-overlap (abc/xyz), multi-byte UTF-8 (café/cafe), invalid UTF-8 (\\xff\\xfe/abc)"
    - "tests/bdd/features/lcsstr.feature exists with at minimum: canonical reference-vector Scenario Outline, identity scenario, both-empty scenario, one-empty scenario, symmetry scenario, leftmost-tie-break scenario, AND a Unicode scenario (via LCSStrScore — rune-surface coverage in unit tests)"
    - "tests/bdd/steps/algorithms_steps.go appends LCSStr step bindings (iComputeTheLCSStrScoreBetween / iComputeTheSecondLCSStrScoreBetween / bothLCSStrScoresShouldBeEqual) and their ctx.Step regex registrations inside InitializeScenario"
    - "testdata/golden/_staging/lcsstr.json exists, produced by TestGolden_LCSStr_Staging via assertGoldenStaging; entries sorted alphabetically by Name; includes at minimum LCSStr_both_empty, LCSStr_identical, LCSStr_one_empty, LCSStr_no_overlap, LCSStr_substring_containment, LCSStr_leftmost_tie_break, LCSStr_unicode_cafe"
    - "algoid_test.go contains a new TestDispatch_LCSStrRegistered asserting dispatch[AlgoLCSStr] non-nil; the registered map in TestDispatch_UnregisteredSlotsAreNil adds int(AlgoLCSStr): true"
    - "FOUR new ExampleXxx funcs appended to example_test.go: ExampleLongestCommonSubstring, ExampleLongestCommonSubstringRunes, ExampleLCSStrScore, ExampleLCSStrScoreRunes — `// Output:` blocks match byte-for-byte"
    - "Coverage on lcsstr.go ≥ 90%; 100% on the four public functions; Apache-2.0 header on every new .go file (scripts/verify-license-headers.sh exits 0)"
  artifacts:
    - path: "lcsstr.go"
      provides: "LongestCommonSubstring, LongestCommonSubstringRunes, LCSStrScore, LCSStrScoreRunes (four public functions); unexported lcsstrDP two-row DP kernel with leftmost-tie-break"
      min_lines: 220
      contains: "Source: Wagner, R. A., Fischer, M. J. (1974)"
    - path: "dispatch_lcsstr.go"
      provides: "Package-load-time registration of LCSStrScore into dispatch[AlgoLCSStr] (slot 8)"
      contains: "dispatch[AlgoLCSStr] = LCSStrScore"
    - path: "lcsstr_test.go"
      provides: "Unit tests for identity, both-empty, one-empty, no-overlap disambiguation pin (Pitfall 6), Wagner-Fischer 1974 reference vectors, LEFTMOST-tie-break pin (Pitfall 4), byte-vs-rune equivalence on ASCII, multi-byte rune (café/cafe), runtime allocation gates (ASCII Short + ASCII Medium)"
    - path: "lcsstr_bench_test.go"
      provides: "Ten benchmarks: BenchmarkLCSStrScore_{ASCII_Short, ASCII_Medium, ASCII_Long, Unicode_Short}, BenchmarkLongestCommonSubstring_{ASCII_Short, ASCII_Medium, ASCII_Long, Unicode_Short}, BenchmarkLCSStrScoreRunes_Unicode_Short, BenchmarkLongestCommonSubstringRunes_Unicode_Short — alloc-asserted"
    - path: "lcsstr_fuzz_test.go"
      provides: "FuzzLCSStrScore exercising ALL FOUR surfaces (Phase 3 WR-02 closure) — panic-free, NaN/Inf-free, score-in-[0,1] invariant"
    - path: "props_test.go"
      provides: "Appended LCSStr property-test block: TestProp_LCSStr{Score,ScoreRunes}_{RangeBounds, Identity, Symmetric, NoNaN, NoInf, NoNegativeZero} PLUS LCSStr-specific TestProp_LongestCommonSubstring_{IsSubstringOfBoth, LengthMatchesScore, LeftmostTieBreak}"
    - path: "example_test.go"
      provides: "Appended ExampleLongestCommonSubstring, ExampleLongestCommonSubstringRunes, ExampleLCSStrScore, ExampleLCSStrScoreRunes"
    - path: "algorithms_golden_test.go"
      provides: "Appended buildLCSStrStagingEntries + TestGolden_LCSStr_Staging — produces _staging/lcsstr.json via assertGoldenStaging; algorithms-merge list NOT updated here (plan 04-05 owns the merge)"
      contains: "TestGolden_LCSStr_Staging"
    - path: "algoid_test.go"
      provides: "Appended TestDispatch_LCSStrRegistered; updated `registered` map flipping slot 8 to true"
      contains: "TestDispatch_LCSStrRegistered"
    - path: "testdata/golden/_staging/lcsstr.json"
      provides: "Per-algorithm staging file (Phase-2-locked pattern); merged into algorithms.json by plan 04-05"
      contains: "LCSStr_leftmost_tie_break"
    - path: "testdata/fuzz/FuzzLCSStrScore/seed-001"
      provides: "Fuzz seed corpus file in `go test fuzz v1` literal format"
    - path: "tests/bdd/features/lcsstr.feature"
      provides: "Gherkin feature with scenarios: canonical reference-vector outline, identity, both-empty, one-empty, symmetry, leftmost-tie-break, Unicode"
    - path: "tests/bdd/steps/algorithms_steps.go"
      provides: "Appended LCSStr step methods on AlgorithmContext + LCSStr ctx.Step registrations inside InitializeScenario"
  key_links:
    - from: "dispatch_lcsstr.go"
      to: "algoid.go (line 107 AlgoLCSStr declared at slot 8)"
      via: "`var _ = func() bool { dispatch[AlgoLCSStr] = LCSStrScore; return true }()`"
      pattern: "dispatch\\[AlgoLCSStr\\]\\s*=\\s*LCSStrScore"
    - from: "lcsstr.go (lcsstrDP kernel)"
      to: "levenshtein.go (maxStackInputLen + isASCII gate)"
      via: "Reference (NOT redeclaration) of maxStackInputLen (declared in levenshtein.go) and isASCII (declared in normalise.go); stack buffer `var buf [(maxStackInputLen+1)*2]int`"
      pattern: "maxStackInputLen|isASCII"
    - from: "lcsstr.go (LongestCommonSubstring score normalisation)"
      to: "docs/requirements.md §7.1.9 SPEC-PINNED formula 2·len(lcs)/(len(a)+len(b))"
      via: "Explicit left-to-right parenthesisation `(2.0 * float64(n)) / float64(la+lb)` per DET-06"
      pattern: "2\\.0 \\* float64"
---

<objective>
Implement LCSStr (CHAR-09) — Wagner-Fischer 1974 longest common substring with FOUR public functions (LongestCommonSubstring + Runes + LCSStrScore + Runes). Two-row rolling-buffer DP with stack-buffer ASCII fast path inherited from Phase 2. Leftmost-in-`a` tie-break enforced by strict-`>` max-update (RESEARCH.md Pitfall 4). Score normalisation `2·len(lcs)/(len(a)+len(b))` SPEC-PINNED per docs/requirements.md §7.1.9. Cross-validation against Wagner-Fischer 1974 reference vectors. Full Phase 2 quality bar: unit + property + fuzz + bench + BDD + staging golden + dispatch + example.

Purpose: ship the canonical longest-common-substring similarity with a non-standard but consumer-justified substring-returning surface (schema-similarity use case — "which substring is driving the match?"). LongestCommonSubstring returning a string IS a v1.0 API commitment; future versions cannot remove the function without a major-version bump.

Output: 13 new/modified files (5 new source/test, 8 extensions to existing append-only files); single new dispatch slot wired; per-algorithm staging golden committed.
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
@.planning/phases/02-core-character-algorithms-six/02-PATTERNS.md
@.planning/phases/04-remaining-character-gestalt/04-01-strcmp95-PLAN.md
@.claude/skills/algorithm-correctness-standards/SKILL.md
@.claude/skills/algorithm-licensing-standards/SKILL.md
@.claude/skills/determinism-standards/SKILL.md
@.claude/skills/performance-standards/SKILL.md
@.claude/skills/go-coding-standards/SKILL.md
@.claude/skills/go-testing-standards/SKILL.md
@levenshtein.go
@swg.go
@dispatch_swg.go
@algoid.go
</context>

<interfaces>
<!-- Key types/functions executor MUST use without rediscovering. -->

From algoid.go (slot 8 already declared, do NOT modify):
```go
const AlgoLCSStr AlgoID = ...  // slot 8; existing declaration at algoid.go:107
```

From levenshtein.go (shared constants and fast-path discipline — DO NOT redeclare):
```go
const maxStackInputLen = 64
// ASCII fast-path gate idiom: if n <= maxStackInputLen && isASCII(a) && isASCII(b) { /* stack buffer */ }
```

From normalise.go (shared helper — DO NOT redeclare):
```go
func isASCII(s string) bool  // declared at normalise.go:159-168
```

Public surface to be created by this plan (FOUR symbols):
```go
// LongestCommonSubstring returns the leftmost-in-`a` longest common substring of a and b.
// Returns "" if either input is empty OR the inputs share no characters.
// Use LCSStrScore to disambiguate (LCSStrScore("","")=1.0 vs LCSStrScore("abc","xyz")=0.0).
func LongestCommonSubstring(a, b string) string

// LongestCommonSubstringRunes is the rune-path variant.
func LongestCommonSubstringRunes(a, b string) string

// LCSStrScore returns 2·len(lcs)/(len(a)+len(b)) in [0.0, 1.0].
func LCSStrScore(a, b string) float64

// LCSStrScoreRunes is the rune-path variant.
func LCSStrScoreRunes(a, b string) float64
```

Unexported kernel (internal to lcsstr.go):
```go
// lcsstrDP returns (length, endIndexInA). prev and curr must each be n+1.
// Leftmost-in-`a` tie-break via STRICT `>` (NOT `>=`) max-update.
func lcsstrDP(a, b string, m, n int, prev, curr []int) (length, endI int)
```

Dispatch wiring shape (matches dispatch_swg.go verbatim; only the score function is dispatched):
```go
var _ = func() bool { dispatch[AlgoLCSStr] = LCSStrScore; return true }()
```
</interfaces>

<tasks>

<task type="auto" tdd="true">
  <name>Task 1: Implement lcsstr.go (algorithm + four public functions + dispatch + unit tests)</name>
  <files>lcsstr.go, dispatch_lcsstr.go, lcsstr_test.go, testdata/golden/_staging/lcsstr.json, algorithms_golden_test.go, algoid_test.go, example_test.go</files>
  <read_first>
    - lcsstr.go (current state — confirm it does NOT exist; creating new)
    - .planning/phases/04-remaining-character-gestalt/04-CONTEXT.md §3 (LCSStr locked decisions: 4 public functions, SPEC-PINNED score normalisation, leftmost-in-`a` tie-break, two-row DP)
    - .planning/phases/04-remaining-character-gestalt/04-RESEARCH.md — Pattern 2 (LCSStr two-row DP + max tracking + leftmost-tie-break), Code Examples (lines 739–769 for the DP kernel), Pitfall 4 (`>=` vs `>` tie-break), Pitfall 6 (empty-return disambiguation), Pitfall 8 (left-to-right reduction)
    - .planning/phases/04-remaining-character-gestalt/04-PATTERNS.md §"lcsstr.go", §"dispatch_lcsstr.go", §"lcsstr_test.go", §"isASCII + maxStackInputLen reuse", §"Identity short-circuit on *Runes", §"License + Source Header"
    - .planning/phases/04-remaining-character-gestalt/04-VALIDATION.md (rows 04-02-01..04-02-05, 04-02-09)
    - levenshtein.go (analog for file shape, the two-row DP kernel, maxStackInputLen declaration, and the ASCII fast-path gate)
    - swg_test.go lines 1–409 (reference-vector + byte-vs-rune-equivalence + multi-byte rune + AllocsPerRun shapes)
    - jaro.go lines 172–179 (JaroScoreRunes identity short-circuit pattern — IN-04 closure)
    - dispatch_swg.go (full file — exact template)
    - algoid.go (AlgoLCSStr declared at slot 8 — line 107)
    - algorithms_golden_test.go lines 584–656 (buildSWGStagingEntries + TestGolden_SWG_Staging template)
    - algoid_test.go lines 284–323 (TestDispatch_* template + slot-map extension point)
    - example_test.go lines 108–122 (ExampleSmithWatermanGotohScore template)
    - docs/requirements.md §7.1.9 (LCSStr API SPEC-PINNED — 4 functions, 2·len/(la+lb) formula)
  </read_first>
  <behavior>
    - LCSStrScore("http_request", "http_request_header_fields") returns 2·12/(12+26) ≈ 0.6316 (substring fully contained in b — full match length 12; tolerance 1e-9)
    - LongestCommonSubstring("http_request", "http_request_header_fields") == "http_request"
    - LongestCommonSubstring("abcXYZabc", "abc") == "abc" — FIRST occurrence in `a` (leftmost-tie-break load-bearing for Pitfall 4)
    - LongestCommonSubstring("", "") == "" AND LCSStrScore("", "") == 1.0
    - LongestCommonSubstring("", "abc") == "" AND LCSStrScore("", "abc") == 0.0
    - LongestCommonSubstring("abc", "xyz") == "" AND LCSStrScore("abc", "xyz") == 0.0 (no overlap — Pitfall 6 disambiguation pin)
    - LongestCommonSubstring("abc", "abc") == "abc" AND LCSStrScore("abc", "abc") == 1.0
    - LongestCommonSubstringRunes("café", "cafe") == "caf"; LCSStrScoreRunes("café", "cafe") ≈ 0.75 (multi-byte UTF-8; len in runes for the rune path)
    - LCSStrScore(a, b) == LCSStrScore(b, a) for any (a, b) — symmetric byte AND rune
    - LCSStrScoreRunes("abc", "abc") == 1.0 WITHOUT calling `[]rune` (identity short-circuit per IN-04 closure)
    - dispatch[AlgoLCSStr] non-nil after package load (TestDispatch_LCSStrRegistered)
    - On byte path with ASCII Short input fitting in stack buffer, testing.AllocsPerRun(100, ...) reports 0 allocs
    - DP kernel uses STRICT `>` in max-update — verified by reading lcsstr.go source
  </behavior>
  <action>
    Create lcsstr.go per PATTERNS.md §"lcsstr.go" and RESEARCH.md Pattern 2 + Code Examples (lines 739–769). File order:
    (a) Apache-2.0 header (copy normalise.go lines 1–13 verbatim).
    (b) File-level doc block: cite Wagner-Fischer 1974 as the PRIMARY source with full citation; include the recurrence `D[i,j] = D[i-1,j-1] + 1 if a[i-1] == b[j-1] else 0` inline; state the SPEC-PINNED score normalisation `2·len(lcs)/(len(a)+len(b))` (Sørensen-Dice form); document the LEFTMOST-in-`a` tie-break with reference to the strict-`>` max-update; document the empty-string-return ambiguity for LongestCommonSubstring (Pitfall 6); list edge cases; list implementation discipline bullets (ASCII fast path via maxStackInputLen + isASCII, no init(), no map iteration, no transcendentals).
    (c) `package fuzzymatch`.
    (d) `func LongestCommonSubstring(a, b string) string` — identity short-circuit `if a == b { return a }`; both-empty → ""; one-empty → ""; then byte-path DP via lcsstrDP using ASCII fast-path gate when applicable. Return `a[endI-maxLen : endI]` (or "" if maxLen == 0). Choose the shorter input as the inner dim (n) to bound the stack buffer.
    (e) `func LongestCommonSubstringRunes(a, b string) string` — identity short-circuit BEFORE `[]rune` allocation per IN-04 closure; rune-path DP; return as `string(runeSlice[start:end])`.
    (f) `func LCSStrScore(a, b string) float64` — identity short-circuit `if a == b { return 1.0 }`; both-empty → 1.0; one-empty → 0.0; compute LCS length via an internal length-only helper (`lcsstrLengthOnly`) that AVOIDS allocating the substring; score via explicit left-to-right `numer := 2.0 * float64(n); denom := float64(la+lb); return numer / denom` per DET-06.
    (g) `func LCSStrScoreRunes(a, b string) float64` — identity short-circuit; rune-path length-only computation; same score formula.
    (h) Unexported `lcsstrDP(a, b string, m, n int, prev, curr []int) (length, endI int)` — STRICT `>` max-update per RESEARCH.md Code Examples lines 748–769. Recurrence resets to 0 on mismatch. After each row swap `prev, curr = curr, prev` then zero the new `curr`.
    (i) Optionally unexported `lcsstrLengthOnly` helper (or fold into LCSStrScore body).

    The DP kernel MUST use STRICT `>` (NOT `>=`) in the max-update — this is the load-bearing regression test for Pitfall 4. Document this inline.

    Create dispatch_lcsstr.go per PATTERNS.md §"dispatch_lcsstr.go" — full file pattern substituting `LCSStr`. Only `LCSStrScore` is registered; `LongestCommonSubstring*` and `LCSStrScoreRunes` are public but not dispatched (dispatch table maps AlgoID → `(a, b string) float64` only — PATTERNS.md gotcha). NO init().

    Create lcsstr_test.go with at minimum 9 tests per PATTERNS.md §"lcsstr_test.go":
    1. TestLCSStr_BothEmpty — LongestCommonSubstring("","")=="" AND LCSStrScore("","")==1.0
    2. TestLCSStr_OneEmpty — LongestCommonSubstring("","abc")=="" AND LCSStrScore("","abc")==0.0
    3. TestLCSStr_NoOverlap_DisambiguationPin — LongestCommonSubstring("abc","xyz")=="" AND LCSStrScore("abc","xyz")==0.0 (Pitfall 6)
    4. TestLCSStr_ReferenceVectors_WagnerFischer1974 — canonical pairs from Wagner-Fischer 1974 (table-driven). Include http_request/http_request_header_fields, kitten/sitting (LCS substring is "itt" length 3 → 2·3/13 ≈ 0.4615), and at least one additional pair.
    5. TestLCSStr_LeftmostTieBreak_Pinned — LongestCommonSubstring("abcXYZabc","abc")=="abc" (load-bearing for Pitfall 4 — the strict-`>` regression test). Hand-curated additional tied case.
    6. TestLCSStr_ByteVsRune_Equivalence — ASCII inputs must agree on byte and rune paths.
    7. TestLCSStr_RuneMultiByte — LongestCommonSubstringRunes("café","cafe")=="caf"; LCSStrScoreRunes("café","cafe") ≈ 0.75.
    8. TestLCSStrScore_ZeroAllocs_ASCII_Short — testing.AllocsPerRun(100, ...) on a short pair; assert 0 allocs.
    9. TestLCSStrScore_ZeroAllocs_ASCII_Medium — same shape on ~50-char pair (still within stack budget).

    Append buildLCSStrStagingEntries + TestGolden_LCSStr_Staging to algorithms_golden_test.go (copy lines 584–656 shape). Entries (alphabetical): LCSStr_both_empty, LCSStr_http_request_containment, LCSStr_identical, LCSStr_kitten_sitting, LCSStr_leftmost_tie_break, LCSStr_no_overlap, LCSStr_one_empty, LCSStr_unicode_cafe (computed via LCSStrScoreRunes — note: the golden's "algorithm" field for unicode_cafe should be "LCSStr" to remain in the dispatch family; the entry simply records LCSStrScore on the same byte input or LCSStrScoreRunes — the planner's call. Recommendation: record LCSStrScore on byte-string inputs only in the golden; rune-path scores live in unit tests, not the golden — consistent with Phase 2 staging-golden pattern). Run `go test -run TestGolden_LCSStr_Staging -update ./...` to materialise the file.

    Append TestDispatch_LCSStrRegistered to algoid_test.go; extend the `registered` map in TestDispatch_UnregisteredSlotsAreNil with `int(AlgoLCSStr): true`.

    Append FOUR ExampleXxx funcs to example_test.go (copy lines 108–122 shape for each):
    - ExampleLongestCommonSubstring (use `fmt.Println(fuzzymatch.LongestCommonSubstring("http_request","http_request_header_fields"))` → "http_request")
    - ExampleLongestCommonSubstringRunes (use `fmt.Println(fuzzymatch.LongestCommonSubstringRunes("café","cafe"))` → "caf")
    - ExampleLCSStrScore (use `fmt.Printf("%.4f\n", fuzzymatch.LCSStrScore("http_request","http_request_header_fields"))`)
    - ExampleLCSStrScoreRunes (use `fmt.Printf("%.4f\n", fuzzymatch.LCSStrScoreRunes("café","cafe"))`)
    Capture exact `// Output:` blocks by running `go test -run ExampleLCSStr|ExampleLongestCommonSubstring ./...` and pasting.
  </action>
  <verify>
    <automated>cd /Users/johnny/Development/fuzzymatch && go build ./... && go test -run 'TestLCSStr|TestDispatch_LCSStrRegistered|TestDispatch_UnregisteredSlotsAreNil|TestGolden_LCSStr_Staging|ExampleLCSStr|ExampleLongestCommonSubstring' ./... && bash scripts/verify-license-headers.sh && ! grep -q "^func init" lcsstr.go && ! grep -E "make\\(\\[\\]\\[\\]int" lcsstr.go && grep -q "// Source: Wagner" lcsstr.go && grep -q "if .* > maxLen" lcsstr.go</automated>
  </verify>
  <done>
    All TestLCSStr* unit tests, TestDispatch_LCSStrRegistered, TestDispatch_UnregisteredSlotsAreNil, TestGolden_LCSStr_Staging, and the four LCSStr/LongestCommonSubstring Example* funcs pass. License headers green. NO init() in lcsstr.go. NO full-table allocation `[][]int{...}`. Wagner-Fischer 1974 cited in godoc. DP kernel uses strict `>` (grep gate).
  </done>
</task>

<task type="auto" tdd="true">
  <name>Task 2: LCSStr property tests + benchmarks + fuzz</name>
  <files>props_test.go, lcsstr_bench_test.go, lcsstr_fuzz_test.go, testdata/fuzz/FuzzLCSStrScore/seed-001</files>
  <read_first>
    - lcsstr.go (created in Task 1)
    - props_test.go lines 737–909 (SWG property-test block — copy template for both LCSStrScore and LCSStrScoreRunes)
    - swg_bench_test.go (analog — but extended for 4 surfaces; PATTERNS.md §lcsstr_bench_test.go)
    - swg_fuzz_test.go (analog — Phase 3 WR-02 multi-surface loop pattern — exercises all 4 LCSStr surfaces)
    - testdata/fuzz/FuzzSmithWatermanGotohScore/seed-001 (literal corpus format)
    - .planning/phases/04-remaining-character-gestalt/04-PATTERNS.md §"lcsstr_bench_test.go", §"lcsstr_fuzz_test.go", §"props_test.go (extend...)", §"var sink", §"testing.AllocsPerRun"
    - .planning/phases/04-remaining-character-gestalt/04-RESEARCH.md — Pitfall 4 (leftmost tie-break property test), Pitfall 6 (empty-return property), Pitfall 8 (left-to-right reduction; range bounds property)
    - .planning/phases/04-remaining-character-gestalt/04-VALIDATION.md (rows 04-02-03, 04-02-04, 04-02-05, 04-02-06, 04-02-07)
  </read_first>
  <behavior>
    - TestProp_LCSStrScore_RangeBounds / _Identity / _Symmetric / _NoNaN / _NoInf / _NoNegativeZero — testing/quick over arbitrary (a, b) byte path
    - TestProp_LCSStrScoreRunes_RangeBounds / _Identity / _Symmetric / _NoNaN / _NoInf / _NoNegativeZero — testing/quick rune path
    - TestProp_LongestCommonSubstring_IsSubstringOfBoth: testing/quick → `strings.Contains(a, got) && strings.Contains(b, got)` for the returned substring on every (a, b)
    - TestProp_LongestCommonSubstring_LengthMatchesScore: testing/quick → `|2·len(LongestCommonSubstring(a,b))/(la+lb) - LCSStrScore(a, b)| < 1e-9`
    - TestProp_LongestCommonSubstring_LeftmostTieBreak: hand-curated tied-candidate inputs (e.g. "abcXYZabc"/"abc") — assert the leftmost match wins; loop ≥ 5 hand-crafted cases
    - BenchmarkLCSStrScore_ASCII_Short: 0 B/op, 0 allocs/op
    - BenchmarkLCSStrScore_ASCII_Long: ≤ 2 allocs/op (two rolling rows via `make`)
    - BenchmarkLongestCommonSubstring_*: similar budgets; `var sink string` anti-DCE
    - BenchmarkLCSStrScoreRunes_Unicode_Short: ≤ 4 allocs/op
    - FuzzLCSStrScore: panic-free, NaN/Inf-free, score-in-[0,1] across ALL FOUR public surfaces (LCSStrScore, LCSStrScoreRunes, LongestCommonSubstring, LongestCommonSubstringRunes)
  </behavior>
  <action>
    Extend props_test.go by appending a new sectioned block at end-of-file:
    // ---------------------------------------------------------------------------
    // LCSStr property tests (plan 04-02)
    // ---------------------------------------------------------------------------
    Append TWELVE standard property tests (six for LCSStrScore byte path, six for LCSStrScoreRunes rune path) — copy lines 737–810 shape verbatim, substituting names. ALSO append THREE LCSStr-specific property tests:
    - TestProp_LongestCommonSubstring_IsSubstringOfBoth — testing/quick over (a, b): let `got := LongestCommonSubstring(a, b)`; if got == "" then return true; else assert `strings.Contains(a, got) && strings.Contains(b, got)`.
    - TestProp_LongestCommonSubstring_LengthMatchesScore — testing/quick: `score := LCSStrScore(a, b)`; `got := LongestCommonSubstring(a, b)`; `expected := 2.0 * float64(len(got)) / float64(len(a)+len(b))`; assert `|score - expected| < 1e-9` (handle both-empty special case where expected == 1.0).
    - TestProp_LongestCommonSubstring_LeftmostTieBreak — hand-curated table of ≥ 5 tied-candidate inputs: ("abcXYZabc", "abc", "abc"), ("xy_abc_xy_abc", "abc", "abc"), ("aaa", "aa", "aa"), etc. For each (a, b, wantSubstring) assert `LongestCommonSubstring(a, b) == wantSubstring` AND the substring's starting index in `a` is the leftmost match (verified via `strings.Index(a, got)`).

    Create lcsstr_bench_test.go per PATTERNS.md §"lcsstr_bench_test.go" with TEN benches:
    - BenchmarkLCSStrScore_{ASCII_Short, ASCII_Medium, ASCII_Long, Unicode_Short}
    - BenchmarkLongestCommonSubstring_{ASCII_Short, ASCII_Medium, ASCII_Long, Unicode_Short} — `var sink string` anti-DCE; gate `if len(sink) < 0 { b.Fatal(...) }`
    - BenchmarkLCSStrScoreRunes_Unicode_Short
    - BenchmarkLongestCommonSubstringRunes_Unicode_Short
    All use b.ReportAllocs() + b.ResetTimer(). ASCII Short inputs MUST fit in the stack buffer (≤ 64 bytes per maxStackInputLen).

    Create lcsstr_fuzz_test.go per PATTERNS.md §"lcsstr_fuzz_test.go" — single FuzzLCSStrScore harness that EXERCISES ALL FOUR PUBLIC SURFACES per Phase 3 WR-02 closure. Seed corpus from PATTERNS.md: kitten/sitting, abc/abc, ""/"", ""/"abc", abcXYZabc/abc (leftmost-tie-break), abc/xyz (no-overlap), \xff\xfe/abc (invalid UTF-8), café/cafe (multi-byte), Привет/привет (Cyrillic). Fuzz body:
    - Call LCSStrScore + LCSStrScoreRunes; assert no NaN/Inf, in [0, 1].
    - Call LongestCommonSubstring + LongestCommonSubstringRunes; assert no panic (use `_ = ...`).

    Create testdata/fuzz/FuzzLCSStrScore/seed-001 in `go test fuzz v1` literal format with kitten/sitting as the canonical seed (3 lines: header + 2 `string(...)` lines). Format byte-stable per IN-06 closure.
  </action>
  <verify>
    <automated>cd /Users/johnny/Development/fuzzymatch && go test -run 'TestProp_LCSStr|TestProp_LongestCommonSubstring' ./... && go test -bench=BenchmarkLCSStr -benchmem -benchtime=1x ./... && go test -bench=BenchmarkLongestCommonSubstring -benchmem -benchtime=1x ./... && go test -fuzz=FuzzLCSStrScore -fuzztime=10s ./... && head -1 testdata/fuzz/FuzzLCSStrScore/seed-001 | grep -q "^go test fuzz v1$"</automated>
  </verify>
  <done>
    All TestProp_LCSStr*/TestProp_LongestCommonSubstring_* tests pass. Ten benches present and produce allocation reports within the budgets. Fuzz harness 10s smoke run succeeds without panic/NaN/Inf/out-of-range across all four surfaces. On-disk seed file byte-stable.
  </done>
</task>

<task type="auto" tdd="true">
  <name>Task 3: LCSStr BDD feature + steps</name>
  <files>tests/bdd/features/lcsstr.feature, tests/bdd/steps/algorithms_steps.go</files>
  <read_first>
    - tests/bdd/features/swg.feature (analog)
    - tests/bdd/steps/algorithms_steps.go (Strcmp95 block from plan 04-01 — extend with LCSStr block following the same pattern)
    - .planning/phases/04-remaining-character-gestalt/04-PATTERNS.md §"tests/bdd/features/...feature", §"tests/bdd/steps/algorithms_steps.go"
    - .planning/phases/04-remaining-character-gestalt/04-VALIDATION.md (row 04-02-08)
    - lcsstr.go (created in Task 1)
  </read_first>
  <behavior>
    - godog runs and passes all LCSStr scenarios in tests/bdd/features/lcsstr.feature
    - At minimum: identity, both-empty, one-empty, canonical reference vectors (Wagner-Fischer + http_request containment), symmetry, leftmost-tie-break, Unicode (using LCSStrScore on byte input — rune-path coverage remains in unit tests)
  </behavior>
  <action>
    Create tests/bdd/features/lcsstr.feature per PATTERNS.md §"tests/bdd/features/...feature". Required scenarios:
    - Header comment citing Wagner-Fischer 1974.
    - `Feature: LCSStr similarity (longest common substring)` with one-paragraph description noting LCSStrScore is dispatched and that LongestCommonSubstring exposes the actual substring.
    - `Scenario Outline: canonical reference vectors` with at least 3 rows (kitten/sitting, http_request/http_request_header_fields, identical).
    - `Scenario: identical strings score 1.0`.
    - `Scenario: both-empty strings score 1.0`.
    - `Scenario: one-empty string scores 0.0`.
    - `Scenario: no-overlap inputs score 0.0` — "abc" / "xyz" → 0 (the empty-return ambiguity is documented in godoc; here the SCORE makes the case explicit).
    - `Scenario: score is symmetric` (kitten/sitting and sitting/kitten via the second-score step).
    - `Scenario: leftmost-tie-break` — assert LCSStrScore("abcXYZabc","abc") is the SAME value as LCSStrScore("abc","abcXYZabc") (pin the score to a documented value, e.g. 2·3/(9+3) = 0.5) — the score is symmetric; the substring-return is what's leftmost-biased. The substring-return is verified in unit tests, not via BDD (BDD steps cover the score only).
    - `Scenario: Unicode rune path via LCSStrScore` — call LCSStrScore on the byte form of "café"/"cafe" — note that the byte path treats these differently from the rune path. Recommendation: use ASCII-only inputs in BDD; rune-path is unit-tested. ALTERNATIVELY add a scenario `Scenario: ASCII-equivalent identical-bytes` using identical pairs that exercise the identity short-circuit.

    Extend tests/bdd/steps/algorithms_steps.go by appending the LCSStr step-method block per PATTERNS.md template (iComputeTheLCSStrScoreBetween, iComputeTheSecondLCSStrScoreBetween, bothLCSStrScoresShouldBeEqual). Register their regexes inside InitializeScenario alongside the existing Strcmp95 + SWG registrations. Reuse the existing `(\d+\.?\d*)` approximately-step regex (IN-03 closure — accepts integer-form).
  </action>
  <verify>
    <automated>cd /Users/johnny/Development/fuzzymatch && make test-bdd 2>&1 | grep -i 'lcsstr' && cd tests/bdd && go test -run 'LCSStr|Test' ./...</automated>
  </verify>
  <done>
    `make test-bdd` exits 0 with the new LCSStr scenarios green. Feature file covers identity, both-empty, one-empty, no-overlap (score-disambiguation), canonical reference vectors, symmetry, leftmost-tie-break (via score-symmetry equivalent).
  </done>
</task>

</tasks>

<threat_model>
## Trust Boundaries

| Boundary | Description |
|----------|-------------|
| caller → LCSStr* | Untrusted (a, b string) input crosses four public function entry points; library is pure-function |

## STRIDE Threat Register (ASVS Level 1)

| Threat ID | Category | Component | Disposition | Mitigation Plan |
|-----------|----------|-----------|-------------|-----------------|
| T-fuzz-panic | D (Denial of Service via panic on malformed input) | LongestCommonSubstring*, LCSStrScore* on invalid UTF-8 / extreme inputs | mitigate | Task 2 ships FuzzLCSStrScore exercising ALL FOUR public surfaces per Phase 3 WR-02 closure with ≥ 60s harness budget and on-disk seed corpus covering invalid UTF-8, identity, both-empty, one-empty, leftmost-tie-break (abcXYZabc/abc), no-overlap (abc/xyz), multi-byte UTF-8 (café/cafe), Cyrillic (Привет/привет). VALIDATION.md row 04-02-06 |
| T-complexity-attack | D (Denial of Service via algorithmic complexity) | LCSStr two-row DP on pathological extreme-length inputs | accept | LCSStr is O(m·n) time, O(min(m,n)) space (two rolling rows). PERF-01 budget documented; long-input bench (BenchmarkLCSStrScore_ASCII_Long) establishes regression baseline. Pure-function library — caller controls input size. The maxStackInputLen=64 ASCII fast-path threshold limits stack-buffer growth; heap path uses two `make([]int, n+1)` allocations bounded by input length |
| T-float-determinism | T (Tampering of float reduction order across architectures) | LCSStrScore / LCSStrScoreRunes score normalisation | mitigate | Explicit left-to-right `numer := 2.0 * float64(n); denom := float64(la+lb); return numer / denom` per DET-06; no math.Pow/Log/Exp/FMA (grep gate in Task 1 verify command); cross-platform CI matrix verifies byte-identical golden output via testdata/golden/_staging/lcsstr.json merged in plan 04-05 |
</threat_model>

<verification>
- `go build ./...` succeeds.
- `go test -run 'TestLCSStr|TestProp_LCSStr|TestProp_LongestCommonSubstring|TestDispatch_LCSStrRegistered|TestGolden_LCSStr_Staging|ExampleLCSStr|ExampleLongestCommonSubstring' ./...` exits 0.
- `go test -bench=BenchmarkLCSStr -benchmem -benchtime=1x ./...` reports 0 B/op, 0 allocs/op for ASCII_Short; ≤ 2 allocs/op for ASCII_Long.
- `go test -fuzz=FuzzLCSStrScore -fuzztime=60s ./...` no failures (10s smoke OK per-task).
- `make test-bdd` green; LCSStr scenarios visible in godog output.
- `bash scripts/verify-license-headers.sh` exits 0.
- `! grep -q "^func init" lcsstr.go` (no init()).
- `! grep -E "make\\(\\[\\]\\[\\]int" lcsstr.go` (no full DP-table allocation — two-row only).
- `! grep -E "math\\.(Pow|Log|Exp|FMA)" lcsstr.go` (DET-06).
- `grep -q "// Source: Wagner" lcsstr.go` (Wagner-Fischer 1974 cited).
- `make coverage-check` confirms lcsstr.go ≥ 90% coverage and 100% on the four public functions.
</verification>

<success_criteria>
- All three tasks complete; all listed verification commands green.
- testdata/golden/_staging/lcsstr.json exists and is canonical-marshalled.
- testdata/fuzz/FuzzLCSStrScore/seed-001 exists byte-stable.
- Public surface is exactly four new exported functions (LongestCommonSubstring, LongestCommonSubstringRunes, LCSStrScore, LCSStrScoreRunes); pre-existing AlgoLCSStr constant unchanged.
- Dispatch slot 8 wired; TestDispatch_UnregisteredSlotsAreNil updated to flip slot 8 to true.
- Phase 4 plan 04-03 (Ratcliff-Obershelp) can begin — RatcliffObershelp may optionally reuse lcsstr.go's DP kernel for its longest-common-substring inner step per CONTEXT.md D-3.
</success_criteria>

<output>
After completion, create `.planning/phases/04-remaining-character-gestalt/04-02-lcsstr-SUMMARY.md` per the GSD summary template.
</output>
