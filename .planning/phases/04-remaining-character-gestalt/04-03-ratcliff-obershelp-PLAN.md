---
phase: 04-remaining-character-gestalt
plan: 03
type: execute
wave: 3
depends_on:
  - 04-02-lcsstr
files_modified:
  - ratcliff_obershelp.go
  - dispatch_ratcliff_obershelp.go
  - ratcliff_obershelp_test.go
  - ratcliff_obershelp_bench_test.go
  - ratcliff_obershelp_fuzz_test.go
  - props_test.go
  - example_test.go
  - algoid_test.go
  - algorithms_golden_test.go
  - testdata/golden/_staging/ratcliff_obershelp.json
  - testdata/fuzz/FuzzRatcliffObershelpScore/seed-001
  - tests/bdd/features/ratcliff_obershelp.feature
  - tests/bdd/steps/algorithms_steps.go
autonomous: true
requirements:
  - GESTALT-01
tags: [ratcliff-obershelp, dr-dobbs-1988, difflib-equivalent, autojunk-false, recursive-lcsubstring, asymmetric-by-design, gestalt, oq-1-resolution, dispatch-registration, property-tests, fuzz, benchmark, bdd, staging-golden]

must_haves:
  truths:
    # Goal-backward truths (GESTALT-01; ROADMAP §"Remaining Character & Gestalt" success criteria #3)
    - "A caller can `import fuzzymatch` and call RatcliffObershelpScore(\"WIKIMEDIA\", \"WIKIMANIA\") and receive a deterministic float64 in [0.0, 1.0] matching difflib.SequenceMatcher(autojunk=False, a=\"WIKIMEDIA\", b=\"WIKIMANIA\").ratio() within 1e-9"
    - "RatcliffObershelpScore(\"\", \"\") == 1.0 (both-empty identity)"
    - "RatcliffObershelpScore(\"\", \"abc\") == 0.0 (one-empty)"
    - "RatcliffObershelpScore(x, x) == 1.0 for every non-empty x (identity short-circuit)"
    - "RatcliffObershelpScore is INTENTIONALLY ASYMMETRIC: RatcliffObershelpScore(\"tide\", \"diet\") != RatcliffObershelpScore(\"diet\", \"tide\") — mirrors Python difflib.SequenceMatcher.ratio() per CPython bpo-37004. This is documented in godoc and pinned by TestCrossAlgorithm_RatcliffObershelp_AsymmetryPin (added in plan 04-05). [OQ-1 RESOLUTION LOCKED 2026-05-14 — see CONTEXT.md §4 + planner planning_context]"
    - "RatcliffObershelpScoreRunes(\"café\", \"cafe\") returns a deterministic float64 in [0.0, 1.0] handling multi-byte UTF-8; identity short-circuit fires on identical inputs BEFORE `[]rune` allocation"
    - "On the Dr. Dobb's 1988 paper-cited pairs (WIKIMEDIA/WIKIMANIA and GESTALT/GESTALT_PATTERN_MATCHING) RatcliffObershelpScore matches difflib(autojunk=False).ratio() within 1e-9 — verified by TestRatcliffObershelp_DrDobbs1988_ReferenceVectors"
    - "TestRatcliffObershelp_PinnedDrDobbsValue pins at least one Dr. Dobb's 1988 exact value (e.g. WIKIMEDIA/WIKIMANIA ≈ 0.7777777777777778) OUTSIDE the cross-validation corpus (Phase 3 WR-03 closure — numerical regression pin alongside cross-validation, not solely via corpus)"
    - "dispatch[AlgoRatcliffObershelp] (slot 22 — the LAST slot, numAlgorithms-1) is non-nil after package load and equals RatcliffObershelpScore (registered via `var _ = func() bool {...}()`; NO init())"
    # Source-origin discipline (CONTEXT.md §4; algorithm-licensing-standards)
    - "ratcliff_obershelp.go's file-level godoc cites Ratcliff & Metzener 1988 (Dr. Dobb's Journal 13(7):46-51) as the PRIMARY source; cites Python difflib.SequenceMatcher (PSF licence on stdlib) as cross-validation source ONLY (not for code copying); includes the explicit source-origin statement block (Primary / Cross-validation / GPL-LGPL: none / Code copied: none)"
    - "Godoc on RatcliffObershelpScore OPENS with the difflib-equivalence directive verbatim from CONTEXT.md §4: \"RatcliffObershelpScore is the difflib-equivalent. If you want fuzzy string matching that behaves like Python's difflib.ratio(), use this. If you want the RapidFuzz \\\"ratio()\\\" semantics — the Indel formula 2·LCS/(|a|+|b|) used by Token Sort Ratio / Token Set Ratio / Partial Ratio — use those functions in Phase 6 instead.\""
    - "Godoc INCLUDES the exact string `difflib.SequenceMatcher(autojunk=False, a=a, b=b).ratio()` verbatim (PITFALLS §6 closure + RESEARCH.md Pitfall 2 — the autojunk=False qualifier is load-bearing)"
    - "Godoc includes a paragraph documenting the asymmetric-by-design semantics: \"RatcliffObershelpScore is NOT symmetric in argument order. This mirrors Python's difflib.SequenceMatcher(autojunk=False).ratio() behaviour (see CPython bpo-37004). For symmetric similarity, callers should sort inputs by length first or use a different algorithm (e.g. LCSStrScore).\" [OQ-1 resolution]"
    - "ratcliff_obershelp.go contains NO init() function"
    # Public surface (CONTEXT.md §4)
    - "Public surface: exactly two new exported functions — RatcliffObershelpScore(a, b string) float64, RatcliffObershelpScoreRunes(a, b string) float64. Pre-existing AlgoRatcliffObershelp constant already in algoid.go slot 22 (the LAST slot, numAlgorithms-1). No params, no Raw* variant (RO score is always in [0,1] by construction — no clamp needed)"
    - "Identity short-circuit `if a == b { return 1.0 }` present on BOTH public entry points; on RatcliffObershelpScoreRunes the short-circuit fires BEFORE `[]rune(a)` allocation (IN-04 closure)"
    # Determinism + performance (DET-04, DET-06)
    - "RatcliffObershelpScore / RatcliffObershelpScoreRunes never return NaN, +Inf, -Inf, or -0 — verified by TestProp_RatcliffObershelp*_{NoNaN,NoInf,NoNegativeZero}"
    - "TestProp_RatcliffObershelpScore_Symmetric IS NOT INCLUDED — per OQ-1 resolution the Symmetric property test is DROPPED for Ratcliff-Obershelp (asymmetric by design). The other 5 standard property tests apply (RangeBounds, Identity, NoNaN, NoInf, NoNegativeZero) on both byte and rune paths"
    - "No math.Pow / math.Log / math.Exp / math.FMA used anywhere in ratcliff_obershelp.go; score normalisation uses explicit left-to-right parenthesisation per DET-06"
    - "No map iteration on output paths; no goroutines; no I/O"
    # Tie-break + recursion correctness (CONTEXT.md §4; RESEARCH.md Pitfall 3)
    - "roFindLongestMatch implements the difflib.SequenceMatcher.find_longest_match contract: leftmost-in-`a` first, then leftmost-in-`b` among ties — same LCS-substring DP as plan 04-02's lcsstrDP. Per CONTEXT.md D-3 the planner may either reuse lcsstr.go's internal helper OR inline the substring search; either choice is acceptable provided correctness on the cross-validation corpus"
    - "roMatchedLength recursion is bounded by O(min(len(a), len(b))) depth. Per CONTEXT.md D-2 the planner may use the language-native call stack OR an explicit iterative stack; either choice is acceptable provided byte-stable output across the CI matrix"
    # Meta-test + fuzz + BDD discipline
    - "FuzzRatcliffObershelpScore exercises BOTH public surfaces (Score + ScoreRunes) per Phase 3 WR-02 closure — body asserts no panic, scores in [0,1], no NaN/Inf"
    - "testdata/fuzz/FuzzRatcliffObershelpScore/seed-001 exists in `go test fuzz v1` literal corpus format with seed covering identity, both-empty, one-empty, Dr. Dobb's pairs (WIKIMEDIA/WIKIMANIA), autojunk-sensitive 200+char pair, multi-byte UTF-8 (café/cafe), invalid UTF-8"
    - "tests/bdd/features/ratcliff_obershelp.feature exists with at minimum: canonical reference-vector Scenario Outline (Dr. Dobb's 1988 pairs), identity scenario, both-empty scenario, one-empty scenario, 200+-char autojunk-sensitive scenario, AND OMITS the symmetry scenario (OQ-1 resolution — RO is asymmetric by design)"
    - "tests/bdd/steps/algorithms_steps.go appends RatcliffObershelp step bindings (iComputeTheRatcliffObershelpScoreBetween; NO 'second' / 'equal' steps because the symmetry scenario is omitted per OQ-1) and its ctx.Step regex registration inside InitializeScenario"
    - "testdata/golden/_staging/ratcliff_obershelp.json exists, produced by TestGolden_RatcliffObershelp_Staging via assertGoldenStaging; entries sorted alphabetically by Name; includes at minimum RatcliffObershelp_both_empty, RatcliffObershelp_gestalt_paper, RatcliffObershelp_identical, RatcliffObershelp_one_empty, RatcliffObershelp_substring_middle, RatcliffObershelp_tide_diet_asymmetric, RatcliffObershelp_wikimedia_wikimania"
    - "algoid_test.go contains a new TestDispatch_RatcliffObershelpRegistered asserting dispatch[AlgoRatcliffObershelp] non-nil; the registered map in TestDispatch_UnregisteredSlotsAreNil adds int(AlgoRatcliffObershelp): true"
    - "TWO new ExampleXxx funcs appended to example_test.go: ExampleRatcliffObershelpScore (on a Dr. Dobb's pair), ExampleRatcliffObershelpScoreRunes (on café/cafe) — `// Output:` blocks match byte-for-byte"
    - "Coverage on ratcliff_obershelp.go ≥ 90%; 100% on the two public functions; Apache-2.0 header on every new .go file (scripts/verify-license-headers.sh exits 0)"
  artifacts:
    - path: "ratcliff_obershelp.go"
      provides: "RatcliffObershelpScore + RatcliffObershelpScoreRunes (two public functions); unexported roMatchedLength + roFindLongestMatch — recursive longest-common-substring decomposition per Ratcliff & Metzener 1988"
      min_lines: 200
      contains: "Source: Ratcliff, J. W., Metzener, D. E. (1988)"
    - path: "dispatch_ratcliff_obershelp.go"
      provides: "Package-load-time registration of RatcliffObershelpScore into dispatch[AlgoRatcliffObershelp] (slot 22 — numAlgorithms-1)"
      contains: "dispatch[AlgoRatcliffObershelp] = RatcliffObershelpScore"
    - path: "ratcliff_obershelp_test.go"
      provides: "Unit tests for identity, both-empty, one-empty, Dr. Dobb's 1988 paper-cited reference vectors (WIKIMEDIA/WIKIMANIA, GESTALT/GESTALT_PATTERN_MATCHING), one numerical pin OUTSIDE the corpus (WR-03 closure), byte-vs-rune equivalence on ASCII, multi-byte rune (café/cafe), asymmetric-by-design pin (tide/diet — DOCUMENTING asymmetry, not enforcing symmetry)"
    - path: "ratcliff_obershelp_bench_test.go"
      provides: "Five benchmarks: BenchmarkRatcliffObershelpScore_{ASCII_Short, ASCII_Medium, ASCII_Long, Unicode_Short} + BenchmarkRatcliffObershelpScoreRunes_Unicode_Short — alloc-asserted"
    - path: "ratcliff_obershelp_fuzz_test.go"
      provides: "FuzzRatcliffObershelpScore exercising BOTH surfaces (Phase 3 WR-02 closure) — panic-free, NaN/Inf-free, score-in-[0,1]"
    - path: "props_test.go"
      provides: "Appended RatcliffObershelp property-test block: TestProp_RatcliffObershelp{Score,ScoreRunes}_{RangeBounds, Identity, NoNaN, NoInf, NoNegativeZero} (FIVE invariants per surface — NO Symmetric per OQ-1) PLUS RO-specific TestProp_RatcliffObershelpScore_AtLeastLevenshtein_HandCurated (hand-curated substring-containment inputs only — RESEARCH.md notes the property is 'generally' true, not universal)"
    - path: "example_test.go"
      provides: "Appended ExampleRatcliffObershelpScore + ExampleRatcliffObershelpScoreRunes"
    - path: "algorithms_golden_test.go"
      provides: "Appended buildRatcliffObershelpStagingEntries + TestGolden_RatcliffObershelp_Staging — produces _staging/ratcliff_obershelp.json via assertGoldenStaging; algorithms-merge list NOT updated here (plan 04-05 owns the merge)"
      contains: "TestGolden_RatcliffObershelp_Staging"
    - path: "algoid_test.go"
      provides: "Appended TestDispatch_RatcliffObershelpRegistered; updated `registered` map flipping slot 22 to true"
      contains: "TestDispatch_RatcliffObershelpRegistered"
    - path: "testdata/golden/_staging/ratcliff_obershelp.json"
      provides: "Per-algorithm staging file; merged into algorithms.json by plan 04-05; sorted alphabetically by Name; includes the tide_diet asymmetric pair (recording ONE direction; the asymmetry-pin test in plan 04-05 verifies fwd != rev)"
      contains: "RatcliffObershelp_wikimedia_wikimania"
    - path: "testdata/fuzz/FuzzRatcliffObershelpScore/seed-001"
      provides: "Fuzz seed corpus file"
    - path: "tests/bdd/features/ratcliff_obershelp.feature"
      provides: "Gherkin feature with scenarios: canonical Dr. Dobb's 1988 reference-vector outline, identity, both-empty, one-empty, 200+-char autojunk-sensitive. NO symmetry scenario (OQ-1)"
    - path: "tests/bdd/steps/algorithms_steps.go"
      provides: "Appended RatcliffObershelp step methods + ctx.Step registrations (NO 'second / equal' methods — only the single Score step is needed)"
  key_links:
    - from: "dispatch_ratcliff_obershelp.go"
      to: "algoid.go (line 180 AlgoRatcliffObershelp declared at slot 22 — numAlgorithms-1)"
      via: "`var _ = func() bool { dispatch[AlgoRatcliffObershelp] = RatcliffObershelpScore; return true }()`"
      pattern: "dispatch\\[AlgoRatcliffObershelp\\]\\s*=\\s*RatcliffObershelpScore"
    - from: "ratcliff_obershelp.go (roFindLongestMatch)"
      to: "Python difflib find_longest_match contract (leftmost-in-`a` first, then leftmost-in-`b`)"
      via: "Same LCS-substring DP recurrence as plan 04-02's lcsstrDP; per CONTEXT.md D-3 the planner picks reuse-or-inline. Recommendation: inline a substring-position-returning variant (roFindLongestMatch returns aLo/aHi/bLo/bHi/n) since RO needs more than length"
      pattern: "roFindLongestMatch|leftmost"
    - from: "ratcliff_obershelp.go (godoc difflib-equivalence directive)"
      to: "Phase 6 TokenSortRatio / TokenSetRatio / PartialRatio cross-reference"
      via: "Godoc explicitly directs RapidFuzz-Indel-formula users at Phase 6 functions; consumers wanting difflib semantics stay here (PITFALLS §6 closure)"
      pattern: "difflib\\.SequenceMatcher\\(autojunk=False"
---

<objective>
Implement Ratcliff-Obershelp (GESTALT-01) — the difflib-equivalent for consumers who want `difflib.SequenceMatcher.ratio()` semantics. Recursive longest-common-substring decomposition per Ratcliff & Metzener (Dr. Dobb's Journal 1988). Two public functions (Score + ScoreRunes; no Raw, no params). NO `*Score_Symmetric` property test — per OQ-1 resolution (LOCKED 2026-05-14) the algorithm is asymmetric by design to preserve byte-for-byte difflib equivalence. Cross-validation corpus is plan 04-04's deliverable; this plan ships the algorithm + Dr. Dobb's paper-pinned unit tests + the standard Phase 2/3 quality bar.

Purpose: ship the load-bearing Python-difflib-equivalent. Phase 6's TokenSortRatio / TokenSetRatio / PartialRatio godoc will point users wanting `difflib.ratio()` semantics here — the cross-reference is the entire reason Ratcliff-Obershelp exists as a separate algorithm in the catalogue.

Output: 13 new/modified files (5 new source/test, 8 extensions to existing append-only files); single new dispatch slot wired; per-algorithm staging golden committed; godoc directive establishing the difflib-equivalence contract.
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
@.planning/phases/04-remaining-character-gestalt/04-01-strcmp95-PLAN.md
@.planning/phases/04-remaining-character-gestalt/04-02-lcsstr-PLAN.md
@.claude/skills/algorithm-correctness-standards/SKILL.md
@.claude/skills/algorithm-licensing-standards/SKILL.md
@.claude/skills/determinism-standards/SKILL.md
@.claude/skills/performance-standards/SKILL.md
@.claude/skills/go-coding-standards/SKILL.md
@.claude/skills/go-testing-standards/SKILL.md
@swg.go
@dispatch_swg.go
@lcsstr.go
@algoid.go
</context>

<interfaces>
<!-- Key types/functions executor MUST use without rediscovering. -->

From algoid.go (slot 22 already declared — the LAST slot, numAlgorithms-1; do NOT modify):
```go
const AlgoRatcliffObershelp AlgoID = ...  // slot 22; existing declaration at algoid.go:180
const numAlgorithms = int(AlgoRatcliffObershelp) + 1  // algoid.go:187
```

From lcsstr.go (plan 04-02 — RO may optionally reuse the DP kernel for its longest-common-substring inner step per CONTEXT.md D-3):
```go
// Unexported `lcsstrDP` returns (length, endIndexInA). RO needs aLo/aHi/bLo/bHi/n,
// so the simplest path is INLINE a substring-position-returning variant
// `roFindLongestMatch(a, b string) (aLo, aHi, bLo, bHi, n int)`.
```

Public surface to be created by this plan (TWO symbols):
```go
// RatcliffObershelpScore returns the Ratcliff-Obershelp gestalt-pattern-matching
// similarity in [0.0, 1.0]. Behaves like
// difflib.SequenceMatcher(autojunk=False, a=a, b=b).ratio().
//
// NOT symmetric in argument order (mirrors difflib — see CPython bpo-37004).
func RatcliffObershelpScore(a, b string) float64

// RatcliffObershelpScoreRunes is the rune-path variant.
func RatcliffObershelpScoreRunes(a, b string) float64
```

Unexported helpers internal to ratcliff_obershelp.go:
```go
// roMatchedLength returns the total matched-character count across the
// recursive longest-common-substring decomposition of a and b.
func roMatchedLength(a, b string) int

// roFindLongestMatch returns the leftmost-longest match equivalent to
// difflib.SequenceMatcher.find_longest_match (autojunk=False).
func roFindLongestMatch(a, b string) (aLo, aHi, bLo, bHi, n int)

// roMatchedLengthRunes / roFindLongestMatchRunes — rune-slice variants.
```

Dispatch wiring shape (matches dispatch_swg.go verbatim):
```go
var _ = func() bool { dispatch[AlgoRatcliffObershelp] = RatcliffObershelpScore; return true }()
```
</interfaces>

<tasks>

<task type="auto" tdd="true">
  <name>Task 1: Implement ratcliff_obershelp.go (algorithm + recursive decomposition + dispatch + unit tests)</name>
  <files>ratcliff_obershelp.go, dispatch_ratcliff_obershelp.go, ratcliff_obershelp_test.go, testdata/golden/_staging/ratcliff_obershelp.json, algorithms_golden_test.go, algoid_test.go, example_test.go</files>
  <read_first>
    - ratcliff_obershelp.go (current state — confirm it does NOT exist; creating new)
    - .planning/phases/04-remaining-character-gestalt/04-CONTEXT.md §4 (Ratcliff-Obershelp locked decisions: 2 public functions, no Raw, no params, recursive LCSubstr, difflib(autojunk=False) equivalence within 1e-9, godoc directive)
    - .planning/phases/04-remaining-character-gestalt/04-RESEARCH.md — Pattern 3 (recursive LCSubstr decomposition), Code Examples (lines 772–801 for roMatchedLength + roFindLongestMatch), Pitfall 2 (autojunk=False is load-bearing), Pitfall 3 (decomposition tie-break must match difflib leftmost-in-`a`-then-leftmost-in-`b`), Open Questions OQ-1 (resolution: drop Symmetric for RO)
    - .planning/phases/04-remaining-character-gestalt/04-PATTERNS.md §"ratcliff_obershelp.go", §"dispatch_ratcliff_obershelp.go", §"ratcliff_obershelp_test.go", §"Identity short-circuit on *Runes", §"License + Source Header"
    - .planning/phases/04-remaining-character-gestalt/04-VALIDATION.md (rows 04-03-01..04-03-04, 04-03-07)
    - swg.go lines 1–82 (file-level multi-source attribution block — copy structure; substitute Ratcliff & Metzener 1988 + difflib as cross-validation)
    - swg_test.go lines 1–409 (reference-vector + byte-vs-rune-equivalence + multi-byte rune shapes)
    - lcsstr.go (from plan 04-02 — read to understand the DP kernel that RO's `roFindLongestMatch` mirrors; CONTEXT.md D-3 decision)
    - dispatch_swg.go (full file — exact template)
    - algoid.go (AlgoRatcliffObershelp declared at slot 22, line 180; numAlgorithms-1)
    - algorithms_golden_test.go lines 584–656 (buildSWGStagingEntries template)
    - algoid_test.go lines 284–323 (TestDispatch_* template)
    - CPython Lib/difflib.py (consulted ONLY for the find_longest_match tie-break contract — leftmost-in-`a` first, then leftmost-in-`b`; no code copied per algorithm-licensing-standards)
  </read_first>
  <behavior>
    - RatcliffObershelpScore("WIKIMEDIA", "WIKIMANIA") ≈ 0.7777777777777778 (difflib autojunk=False — Dr. Dobb's 1988 canonical pair; tolerance 1e-9)
    - RatcliffObershelpScore("GESTALT", "GESTALT_PATTERN_MATCHING") matches difflib(autojunk=False).ratio() within 1e-9
    - RatcliffObershelpScore("", "") == 1.0; RatcliffObershelpScore("", "abc") == 0.0
    - RatcliffObershelpScore("abc", "abc") == 1.0 (identity short-circuit)
    - RatcliffObershelpScoreRunes("café", "cafe") returns a deterministic float in [0, 1] handling multi-byte UTF-8; identity short-circuit fires on "café"/"café" BEFORE `[]rune` allocation
    - RatcliffObershelpScore("tide", "diet") != RatcliffObershelpScore("diet", "tide") — asymmetric by design (pinned in TestRatcliffObershelp_AsymmetryPin in this task; cross-algorithm consistency test added in plan 04-05)
    - dispatch[AlgoRatcliffObershelp] non-nil after package load
  </behavior>
  <action>
    Create ratcliff_obershelp.go per PATTERNS.md §"ratcliff_obershelp.go" and RESEARCH.md Pattern 3 + Code Examples (lines 772–801). File order:
    (a) Apache-2.0 header (copy normalise.go lines 1–13 verbatim).
    (b) File-level doc block: cite Ratcliff, J. W., Metzener, D. E. (1988). "Pattern matching: the gestalt approach." Dr. Dobb's Journal, 13(7):46-51 as the PRIMARY source. Describe the algorithm in 5 numbered steps (find longest common substring → recurse left → recurse right → sum matched-character count → score = 2·M/(la+lb)). Cite Python difflib.SequenceMatcher (PSF licence on stdlib) as the cross-validation reference; include the explicit source-origin statement block (Primary / Cross-validation / GPL-LGPL: none / Code copied: none). Document the autojunk=False qualifier and why it matters (Pitfall 2). Document the asymmetric-by-design semantics per OQ-1 resolution.
    (c) `package fuzzymatch`.
    (d) `func RatcliffObershelpScore(a, b string) float64` — OPENS with the difflib-equivalence directive verbatim per CONTEXT.md §4 (4 lines pointing at RapidFuzz-Indel users at Phase 6, the difflib(autojunk=False) qualifier, the asymmetry paragraph per OQ-1, edge cases). Body: identity short-circuit `if a == b { return 1.0 }`; one-empty → 0.0; compute `m := roMatchedLength(a, b)`; score via explicit left-to-right `numer := 2.0 * float64(m); denom := float64(len(a)+len(b)); return numer / denom` per DET-06.
    (e) `func RatcliffObershelpScoreRunes(a, b string) float64` — identity short-circuit BEFORE `[]rune` per IN-04 closure; rune-path recursion; same score formula in rune length.
    (f) Unexported `roMatchedLength(a, b string) int` — per RESEARCH.md Code Examples lines 775–796. Recursion is bounded by O(min(la, lb)) depth. Per CONTEXT.md D-2 the planner may use the language-native call stack OR an iterative explicit stack — RECOMMENDATION: native call stack (simpler; byte-stable; recursion depth bounded). Document the choice in a code comment.
    (g) Unexported `roFindLongestMatch(a, b string) (aLo, aHi, bLo, bHi, n int)` — same LCS-substring DP recurrence as lcsstrDP from plan 04-02, but returns START indices in both `a` and `b` plus the length. Tie-break: leftmost-in-`a` first (strict `>` max-update), then leftmost-in-`b` among ties. CONTEXT.md D-3 — RECOMMENDATION: INLINE the substring search (not reuse lcsstr.go's helper) — RO needs aLo/bLo, not just the length, so the abstraction does not earn its keep here.
    (h) Unexported `roMatchedLengthRunes(ra, rb []rune) int` + `roFindLongestMatchRunes` — rune-slice variants. Pass `[]rune` slices through the recursion per RESEARCH.md OQ-2 recommendation (Phase 2 precedent).

    Use ONLY `+`, `-`, `*`, `/`, comparisons in the float arithmetic — NO `math.Pow`/`math.Log`/`math.Exp`/`math.FMA`. NO init(). NO map iteration.

    Create dispatch_ratcliff_obershelp.go per PATTERNS.md §"dispatch_ratcliff_obershelp.go" — full file pattern substituting `RatcliffObershelp`. NO init().

    Create ratcliff_obershelp_test.go per PATTERNS.md §"ratcliff_obershelp_test.go" Plan 04-03 content:
    - Apache-2.0 header + package fuzzymatch_test + imports.
    - TestRatcliffObershelp_BothEmpty (copy swg_test.go lines 38–54 shape).
    - TestRatcliffObershelp_OneEmpty (copy lines 56–77 shape).
    - TestRatcliffObershelp_Identical (copy lines 79–102 shape).
    - TestRatcliffObershelp_DrDobbs1988_ReferenceVectors — canonical pairs (WIKIMEDIA/WIKIMANIA, GESTALT/GESTALT_PATTERN_MATCHING) per CONTEXT.md §1 Category 2. Table-driven `t.Run(tt.a+"_"+tt.b, ...)`. Tolerance 1e-9. The expected values come from running Python `difflib.SequenceMatcher(autojunk=False, a=a, b=b).ratio()` once and pasting the values into the test as float64 literals.
    - TestRatcliffObershelp_PinnedDrDobbsValue — PIN at least ONE exact-value Dr. Dobb's pair OUTSIDE the cross-validation corpus (Phase 3 WR-03 closure): e.g. `RatcliffObershelpScore("WIKIMEDIA", "WIKIMANIA")` within `1e-9` of the hand-computed difflib value (0.7777777777777778). This catches a regression even if the cross-validation corpus is somehow accepted unchanged.
    - TestRatcliffObershelp_AsymmetryPin — pin `RatcliffObershelpScore("tide", "diet") != RatcliffObershelpScore("diet", "tide")` (asymmetric-by-design — OQ-1 resolution). Compute both directions, assert inequality. This is the load-bearing local pin; plan 04-05's cross-algorithm consistency test adds an inverse-form regression guard.
    - TestRatcliffObershelp_ByteVsRune_Equivalence (ASCII-only inputs must agree on byte and rune paths; copy swg_test.go lines 162–186 shape).
    - TestRatcliffObershelp_RuneMultiByte — `RatcliffObershelpScoreRunes("café","cafe")` returns a deterministic value in [0, 1]; verify it differs from `RatcliffObershelpScore("café","cafe")` because the byte path treats "café" as 5 bytes vs 4 bytes (the byte-path result is a function of the byte string, not the Unicode codepoint sequence).

    Append buildRatcliffObershelpStagingEntries + TestGolden_RatcliffObershelp_Staging to algorithms_golden_test.go (copy lines 584–656 shape). Entries (alphabetical): RatcliffObershelp_both_empty, RatcliffObershelp_gestalt_paper, RatcliffObershelp_identical, RatcliffObershelp_one_empty, RatcliffObershelp_substring_middle (e.g. "abcdef"/"xyzabcdefuvw"), RatcliffObershelp_tide_diet_asymmetric (record ONE direction — the asymmetry is verified in plan 04-05's cross-algorithm test), RatcliffObershelp_wikimedia_wikimania. Run `go test -run TestGolden_RatcliffObershelp_Staging -update ./...` to materialise.

    Append TestDispatch_RatcliffObershelpRegistered to algoid_test.go; extend the `registered` map in TestDispatch_UnregisteredSlotsAreNil with `int(AlgoRatcliffObershelp): true` — this also flips the LAST slot to true, which means TestDispatch_UnregisteredSlotsAreNil now expects every slot 0..numAlgorithms-1 to be registered. Update the test's godoc comment accordingly.

    Append TWO ExampleXxx funcs to example_test.go (copy lines 108–122 shape):
    - ExampleRatcliffObershelpScore (use a Dr. Dobb's pair; pin the value)
    - ExampleRatcliffObershelpScoreRunes (use café/cafe; pin the value)
    Capture exact `// Output:` blocks by running `go test -run ExampleRatcliffObershelp ./...` once.
  </action>
  <verify>
    <automated>cd /Users/johnny/Development/fuzzymatch && go build ./... && go test -run 'TestRatcliffObershelp|TestDispatch_RatcliffObershelpRegistered|TestDispatch_UnregisteredSlotsAreNil|TestGolden_RatcliffObershelp_Staging|ExampleRatcliffObershelp' ./... && bash scripts/verify-license-headers.sh && ! grep -q "^func init" ratcliff_obershelp.go && grep -q "// Source: Ratcliff, J. W., Metzener, D. E. (1988)" ratcliff_obershelp.go && grep -q "difflib.SequenceMatcher(autojunk=False" ratcliff_obershelp.go && grep -q "NOT symmetric in argument order" ratcliff_obershelp.go</automated>
  </verify>
  <done>
    All TestRatcliffObershelp* unit tests pass (Dr. Dobb's vectors within 1e-9 of expected; asymmetry pinned). TestDispatch_RatcliffObershelpRegistered green; TestDispatch_UnregisteredSlotsAreNil updated to expect ALL slots registered. TestGolden_RatcliffObershelp_Staging materialises _staging/ratcliff_obershelp.json. Both ExampleRatcliffObershelp* funcs green. License headers green. NO init(). Ratcliff & Metzener 1988 cited. difflib(autojunk=False) directive present in godoc. Asymmetry paragraph present in godoc.
  </done>
</task>

<task type="auto" tdd="true">
  <name>Task 2: Ratcliff-Obershelp property tests + benchmarks + fuzz</name>
  <files>props_test.go, ratcliff_obershelp_bench_test.go, ratcliff_obershelp_fuzz_test.go, testdata/fuzz/FuzzRatcliffObershelpScore/seed-001</files>
  <read_first>
    - ratcliff_obershelp.go (created in Task 1)
    - props_test.go lines 737–909 (SWG property-test block — template; NOTE: OMIT Symmetric for RO per OQ-1)
    - swg_bench_test.go (analog; adapt for two surfaces only)
    - swg_fuzz_test.go (analog; multi-surface loop pattern for Score + ScoreRunes)
    - testdata/fuzz/FuzzSmithWatermanGotohScore/seed-001 (literal corpus format)
    - .planning/phases/04-remaining-character-gestalt/04-PATTERNS.md §"ratcliff_obershelp_bench_test.go", §"ratcliff_obershelp_fuzz_test.go", §"props_test.go (extend in plans 04-01, 04-02, 04-03)" — see the "Critical deviation per OQ-1" note
    - .planning/phases/04-remaining-character-gestalt/04-RESEARCH.md Pattern 3, Pitfall 2 (autojunk-sensitive seed), Open Questions OQ-1 (Symmetric DROPPED)
    - .planning/phases/04-remaining-character-gestalt/04-VALIDATION.md (rows 04-03-04, 04-03-05)
  </read_first>
  <behavior>
    - TestProp_RatcliffObershelpScore_RangeBounds / _Identity / _NoNaN / _NoInf / _NoNegativeZero — testing/quick over arbitrary (a, b) byte path
    - TestProp_RatcliffObershelpScoreRunes_RangeBounds / _Identity / _NoNaN / _NoInf / _NoNegativeZero — testing/quick rune path
    - NO TestProp_RatcliffObershelpScore_Symmetric AND NO TestProp_RatcliffObershelpScoreRunes_Symmetric (OQ-1 resolution — RO is asymmetric by design)
    - TestProp_RatcliffObershelpScore_AtLeastLevenshtein_HandCurated — hand-curated substring-containment inputs only (not testing/quick over all strings; the property is "generally" true, not universal — RESEARCH.md note)
    - BenchmarkRatcliffObershelpScore_*: alloc-asserted; document the recursion's allocation behaviour (each recursive call slices the input string, which is allocation-free in Go)
    - FuzzRatcliffObershelpScore: panic-free, NaN/Inf-free, score-in-[0,1] across BOTH surfaces (Score + ScoreRunes)
  </behavior>
  <action>
    Extend props_test.go by appending a new sectioned block at end-of-file:
    // ---------------------------------------------------------------------------
    // Ratcliff-Obershelp property tests (plan 04-03)
    //
    // NB: TestProp_RatcliffObershelpScore_Symmetric is INTENTIONALLY OMITTED per
    //     OQ-1 resolution (locked 2026-05-14). Ratcliff-Obershelp is asymmetric
    //     by design to preserve byte-for-byte difflib equivalence. See
    //     ratcliff_obershelp.go's godoc and CONTEXT.md §4 for the rationale.
    // ---------------------------------------------------------------------------
    Append TEN standard property tests (FIVE for byte path, FIVE for rune path) — RangeBounds, Identity, NoNaN, NoInf, NoNegativeZero on each. Copy lines 737–810 shape, OMITTING the Symmetric block. Default 100 testing/quick iterations.

    Append TestProp_RatcliffObershelpScore_AtLeastLevenshtein_HandCurated — a hand-curated table of substring-containment inputs (e.g. ("http_request", "http_request_header_fields"), ("abc", "xyzabcdef"), ("kitten", "the_kitten_purrs")). For each (a, b) assert `RatcliffObershelpScore(a, b) >= LevenshteinScore(a, b)`. NOT testing/quick over all strings — the property is "generally" true per RESEARCH.md.

    Create ratcliff_obershelp_bench_test.go per PATTERNS.md §"ratcliff_obershelp_bench_test.go" with FIVE benches:
    - BenchmarkRatcliffObershelpScore_{ASCII_Short, ASCII_Medium, ASCII_Long, Unicode_Short}
    - BenchmarkRatcliffObershelpScoreRunes_Unicode_Short
    `var sink float64` + `if sink < 0 { b.Fatal(...) }` anti-DCE pattern. b.ReportAllocs() + b.ResetTimer().

    Create ratcliff_obershelp_fuzz_test.go per PATTERNS.md §"ratcliff_obershelp_fuzz_test.go" — single FuzzRatcliffObershelpScore exercising BOTH surfaces per Phase 3 WR-02 closure. Seed corpus from PATTERNS.md:
    - Standard edges: identity ("abc"/"abc"), both-empty (""/""), one-empty (""/"abc"), no-overlap ("abc"/"xyz")
    - Dr. Dobb's pairs: WIKIMEDIA/WIKIMANIA, GESTALT/GESTALT_PATTERN_MATCHING
    - Autojunk-sensitive 200+char case: e.g. `"a"*100 + "x"*5 + "a"*100` / `"a"*50 + "y"*5 + "a"*150`
    - Substring containment: "abcdef"/"xyzabcdefuvw"
    - Multi-byte UTF-8: café/cafe, Привет/привет
    - Invalid UTF-8: \xff\xfe/abc
    Fuzz body:
    - Call RatcliffObershelpScore(a, b) AND RatcliffObershelpScoreRunes(a, b); assert no NaN (math.IsNaN), no Inf (math.IsInf), score in [0, 1] on each surface.

    Create testdata/fuzz/FuzzRatcliffObershelpScore/seed-001 in `go test fuzz v1` literal format with WIKIMEDIA/WIKIMANIA as the canonical seed. Format byte-stable per IN-06 closure.
  </action>
  <verify>
    <automated>cd /Users/johnny/Development/fuzzymatch && go test -run 'TestProp_RatcliffObershelp' ./... && ! grep -E "TestProp_RatcliffObershelp.*_Symmetric" props_test.go && go test -bench=BenchmarkRatcliffObershelp -benchmem -benchtime=1x ./... && go test -fuzz=FuzzRatcliffObershelpScore -fuzztime=10s ./... && head -1 testdata/fuzz/FuzzRatcliffObershelpScore/seed-001 | grep -q "^go test fuzz v1$"</automated>
  </verify>
  <done>
    All TestProp_RatcliffObershelp* property tests pass. No Symmetric test exists for RatcliffObershelp (grep gate confirms). Five benches produce allocation reports. Fuzz harness 10s smoke run succeeds without panic/NaN/Inf/out-of-range across both surfaces. On-disk seed file byte-stable.
  </done>
</task>

<task type="auto" tdd="true">
  <name>Task 3: Ratcliff-Obershelp BDD feature + steps</name>
  <files>tests/bdd/features/ratcliff_obershelp.feature, tests/bdd/steps/algorithms_steps.go</files>
  <read_first>
    - tests/bdd/features/swg.feature (analog — but OMIT the symmetry scenario per OQ-1)
    - tests/bdd/steps/algorithms_steps.go (LCSStr block from plan 04-02 — extend with RatcliffObershelp block; NO 'second' / 'equal' methods needed since symmetry scenario is omitted)
    - .planning/phases/04-remaining-character-gestalt/04-PATTERNS.md §"tests/bdd/features/...feature" (`ratcliff_obershelp.feature` deviations: ADD 200+-char autojunk-sensitive scenario; OMIT symmetric scenario per OQ-1)
    - .planning/phases/04-remaining-character-gestalt/04-VALIDATION.md (row 04-03-06)
    - ratcliff_obershelp.go (created in Task 1)
  </read_first>
  <behavior>
    - godog runs and passes all RatcliffObershelp scenarios
    - At minimum: identity, both-empty, one-empty, Dr. Dobb's 1988 reference vectors (Scenario Outline), 200+-char autojunk-sensitive case
    - NO symmetry scenario (per OQ-1)
  </behavior>
  <action>
    Create tests/bdd/features/ratcliff_obershelp.feature per PATTERNS.md §"tests/bdd/features/...feature" with the RO-specific deviations:
    - Header comment citing Ratcliff & Metzener 1988 (Dr. Dobb's Journal).
    - `Feature: Ratcliff-Obershelp similarity (difflib-equivalent)` with one-paragraph description noting "Behaves like Python's difflib.SequenceMatcher(autojunk=False).ratio(). NOT symmetric in argument order."
    - `Scenario Outline: canonical Dr. Dobb's 1988 reference vectors` with 2+ Examples rows (WIKIMEDIA/WIKIMANIA, GESTALT/GESTALT_PATTERN_MATCHING — values copied from the unit tests in Task 1). Tolerance 0.0001.
    - `Scenario: identical strings score 1.0`.
    - `Scenario: both-empty strings score 1.0`.
    - `Scenario: one-empty string scores 0.0` ("abc"/"").
    - `Scenario: 200+ character autojunk-sensitive input` — use the same constructed input as the fuzz seed (e.g. `"a"*100 + "x"*5 + "a"*100` vs `"a"*50 + "y"*5 + "a"*150`); assert the score matches a pinned value from running difflib(autojunk=False).ratio() once. The presence of this scenario proves the impl does NOT have an autojunk-like heuristic (RESEARCH.md Pitfall 2).
    - DO NOT include a symmetric scenario (OQ-1 resolution).

    Extend tests/bdd/steps/algorithms_steps.go by appending the RatcliffObershelp step-method block per PATTERNS.md template — ONLY the `iComputeTheRatcliffObershelpScoreBetween` method is needed (no 'second' / 'equal' methods since the symmetry scenario is omitted). Register its regex inside InitializeScenario: `^I compute the Ratcliff-Obershelp score between "([^"]*)" and "([^"]*)"$`. The existing approximately-step regex `(\d+\.?\d*)` is reused.

    Note: long input strings in the BDD Gherkin file (e.g. the 205-char autojunk-sensitive case) can be embedded via Gherkin doc-string syntax or as escaped strings in the Examples table — pick whichever godog handles cleanly. If doc-string syntax is awkward, an alternative: define the long inputs as Go constants in `algorithms_steps.go` and add a step like `^I compute the Ratcliff-Obershelp score for the autojunk-sensitive pair$` that uses those constants — pin the score with approximately-step.
  </action>
  <verify>
    <automated>cd /Users/johnny/Development/fuzzymatch && make test-bdd 2>&1 | grep -i 'ratcliff' && cd tests/bdd && go test -run 'RatcliffObershelp|Test' ./... && ! grep -i "symmetric" tests/bdd/features/ratcliff_obershelp.feature</automated>
  </verify>
  <done>
    `make test-bdd` exits 0 with the new RatcliffObershelp scenarios green. Feature file has NO symmetry scenario (grep gate). 200+-char autojunk-sensitive scenario passes against the pinned difflib(autojunk=False) value.
  </done>
</task>

</tasks>

<threat_model>
## Trust Boundaries

| Boundary | Description |
|----------|-------------|
| caller → RatcliffObershelpScore* | Untrusted (a, b string) input crosses two public function entry points; library is pure-function |

## STRIDE Threat Register (ASVS Level 1)

| Threat ID | Category | Component | Disposition | Mitigation Plan |
|-----------|----------|-----------|-------------|-----------------|
| T-fuzz-panic | D (Denial of Service via panic on malformed input) | RatcliffObershelpScore* on invalid UTF-8 / extreme inputs | mitigate | Task 2 ships FuzzRatcliffObershelpScore exercising BOTH public surfaces per Phase 3 WR-02 closure with ≥ 60s harness budget. Seed corpus covers identity, both-empty, one-empty, Dr. Dobb's pairs, AUTOJUNK-SENSITIVE 200+char pair (RESEARCH.md Pitfall 2 closure), substring containment, multi-byte UTF-8, Cyrillic, invalid UTF-8. VALIDATION.md row 04-03-05 |
| T-complexity-attack | D (Denial of Service via algorithmic complexity) | RatcliffObershelp recursive LCSubstr on pathological inputs | accept | Worst-case is O(n²·m) for repeated-character pathological inputs (e.g. "aaaa..." / "aaaa..." with strategic differences). RESEARCH.md notes this; godoc documents the worst case. Recursion depth is bounded by O(min(la, lb)) so no stack-overflow risk for reasonable inputs. PERF-01 budget documented; long-input benches in Task 2 establish regression baseline. Pure-function library — caller controls input size |
| T-float-determinism | T (Tampering of float reduction order across architectures) | RatcliffObershelpScore* score normalisation | mitigate | Explicit left-to-right `numer := 2.0 * float64(m); denom := float64(la+lb); return numer / denom` per DET-06; no math.Pow/Log/Exp/FMA (grep gate in Task 1 verify command); cross-platform CI matrix verifies byte-identical golden output via testdata/golden/_staging/ratcliff_obershelp.json merged in plan 04-05 |
</threat_model>

<verification>
- `go build ./...` succeeds.
- `go test -run 'TestRatcliffObershelp|TestProp_RatcliffObershelp|TestDispatch_RatcliffObershelpRegistered|TestGolden_RatcliffObershelp_Staging|ExampleRatcliffObershelp' ./...` exits 0.
- `go test -bench=BenchmarkRatcliffObershelp -benchmem -benchtime=1x ./...` succeeds; allocation budget documented.
- `go test -fuzz=FuzzRatcliffObershelpScore -fuzztime=60s ./...` no failures (10s smoke OK per-task).
- `make test-bdd` green; RatcliffObershelp scenarios visible in godog output.
- `bash scripts/verify-license-headers.sh` exits 0.
- `! grep -q "^func init" ratcliff_obershelp.go` (no init()).
- `! grep -E "math\\.(Pow|Log|Exp|FMA)" ratcliff_obershelp.go` (DET-06).
- `grep -q "// Source: Ratcliff" ratcliff_obershelp.go` (Ratcliff & Metzener 1988 cited).
- `grep -q "difflib.SequenceMatcher(autojunk=False" ratcliff_obershelp.go` (autojunk=False qualifier in godoc — Pitfall 2 closure).
- `grep -q "NOT symmetric in argument order" ratcliff_obershelp.go` (OQ-1 resolution documented).
- `! grep -E "TestProp_RatcliffObershelp.*_Symmetric" props_test.go` (Symmetric DROPPED per OQ-1).
- `make coverage-check` confirms ratcliff_obershelp.go ≥ 90% coverage and 100% on the two public functions.
</verification>

<success_criteria>
- All three tasks complete; all listed verification commands green.
- testdata/golden/_staging/ratcliff_obershelp.json exists and is canonical-marshalled; entries alphabetically sorted; includes the tide/diet asymmetric pair (one direction).
- testdata/fuzz/FuzzRatcliffObershelpScore/seed-001 exists byte-stable.
- Public surface is exactly two new exported functions (RatcliffObershelpScore, RatcliffObershelpScoreRunes); pre-existing AlgoRatcliffObershelp constant unchanged.
- Dispatch slot 22 (the LAST slot, numAlgorithms-1) wired; TestDispatch_UnregisteredSlotsAreNil now expects ALL slots registered (godoc comment updated).
- Phase 4 plan 04-04 (cross-validation corpus) can begin — `TestRatcliffObershelp_CrossValidation` will be appended to ratcliff_obershelp_test.go by that plan.
</success_criteria>

<output>
After completion, create `.planning/phases/04-remaining-character-gestalt/04-03-ratcliff-obershelp-SUMMARY.md` per the GSD summary template.
</output>
