---
phase: 03-smith-waterman-gotoh
plan: 01
type: execute
wave: 1
depends_on: []
files_modified:
  - swg.go
  - dispatch_swg.go
  - swg_test.go
  - swg_bench_test.go
  - swg_fuzz_test.go
  - props_test.go
  - example_test.go
  - algoid_test.go
  - algorithms_golden_test.go
  - testdata/golden/_staging/swg.json
  - testdata/fuzz/FuzzSmithWatermanGotohScore/seed-001
  - tests/bdd/features/swg.feature
  - tests/bdd/steps/algorithms_steps.go
autonomous: true
requirements:
  - CHAR-08
tags: [smith-waterman-gotoh, swg, affine-gap, two-row-dp, ascii-fast-path, gotoh-erratum, swg-params, raw-score-surface, dispatch-registration, property-tests, fuzz, benchmark, bdd, staging-golden]

must_haves:
  truths:
    # Goal-backward truths (user-observable behaviour for the Phase 3 success criteria)
    - "A caller can `import fuzzymatch` and call SmithWatermanGotohScore(\"http_request\", \"http_request_header_fields\") and receive 1.0 exactly (substring fully contained)"
    - "SmithWatermanGotohScore(\"\", \"\") returns exactly 1.0 (both-empty identity)"
    - "SmithWatermanGotohScore(\"\", \"abc\") returns exactly 0.0 (one-empty)"
    - "SmithWatermanGotohScore(x, x) returns exactly 1.0 for every non-empty x (identity short-circuit on byte and rune entry points per IN-02)"
    - "SmithWatermanGotohScore(a, b) == SmithWatermanGotohScore(b, a) for any a, b (symmetric — warning sign #3 in PITFALLS.md §3 must NOT trigger)"
    - "Splitting a long gap with intervening matches does NOT increase the score: SmithWatermanGotohScore(\"abc________def\", \"abcdef\") <= SmithWatermanGotohScore(\"abc____def____\", \"abcdef\") fails when the Gotoh-erratum bug is present and passes when the corrected Flouri 2015 formulation is used"
    - "SmithWatermanGotohRawScore(a, b) returns the unclamped raw alignment score; for the substring pair (http_request, http_request_header_fields) with default params it returns 12.0 exactly (12 match positions × Match=1.0)"
    - "SmithWatermanGotohScoreWithParams(a, b, p) accepts a non-default SWGParams value and produces a deterministic, reproducible result for the given params"
    - "NewSWGParams() returns SWGParams{Match: 1.0, Mismatch: -1.0, GapOpen: -1.5, GapExtend: -0.5} — a fresh value, callers can mutate freely"
    - "dispatch[AlgoSmithWatermanGotoh] (slot 6) is non-nil after package load and equals SmithWatermanGotohScore (no init() — populated by var _ = func() bool {...}() in dispatch_swg.go)"
    # Cross-cutting truths (phase-3 quality bar)
    - "Every new .go file (swg.go, dispatch_swg.go, swg_test.go, swg_bench_test.go, swg_fuzz_test.go) starts with the exact Apache-2.0 header block from normalise.go lines 1-13 (scripts/verify-license-headers.sh exits 0)"
    - "swg.go's file-level godoc cites Smith-Waterman 1981, Gotoh 1982, AND Flouri et al. 2015 (the corrected-formulation source), AND explicitly names the Gotoh 1982 erratum and its correction inline — this is the PITFALLS.md §3 documentary gate"
    - "swg.go declares NO new top-level constants or helpers that duplicate Phase 2 inheritance — maxStackInputLen (levenshtein.go:68) and isASCII (normalise.go:159-168) are referenced by name, NEVER redeclared"
    - "Two-row (three-matrix, six rolling rows) DP confirmed by code review: inner kernel maintains exactly six []float64 slices of length n+1; no [m+1][n+1]float64 table allocated anywhere in swg.go"
    - "ASCII fast path on (lb <= maxStackInputLen && isASCII(a) && isASCII(b)) allocates the six rolling rows on the stack via `var buf [(maxStackInputLen+1)*6]float64` (3120 bytes); benchmark verifies 0 B/op, 0 allocs/op for ASCII Short and Medium"
    - "Normalisation clamp `clamp(raw / float64(min(la, lb)), 0, 1)` is implemented inline in swg.go entry points (no shared helper); the clamp is documented in godoc per CONTEXT.md §4"
    - "Score never returns NaN, +Inf, -Inf, or -0 for any input — verified by TestProp_SmithWatermanGotoh_NoNaN / _NoInf / _NoNegativeZero in props_test.go"
    - "No math.Pow / math.Log / math.Exp / math.FMA used anywhere in swg.go (only `+`, `-`, `*`, `/`, comparisons, and `float64()` casts); grep verified — determinism-standards §13.3 gate"
    - "No init() function in any new file in this plan; no map iteration on output paths (none required — the kernel uses []float64 slices only)"
    - "Sum reductions are left-to-right; no parallel sums, no `math.FMA`, no atomic float operations"
    - "Public surface: SWGParams (type), NewSWGParams (constructor), SmithWatermanGotohScore, SmithWatermanGotohScoreRunes, SmithWatermanGotohScoreWithParams, SmithWatermanGotohRawScore, SmithWatermanGotohRawScoreRunes, SmithWatermanGotohRawScoreWithParams — 6 functions + 1 type + 1 constructor = 8 new exports plus pre-existing AlgoSmithWatermanGotoh"
    - "Identity short-circuit `if a == b { return 1.0 }` present on every public *Score and *RawScore entry point (per IN-02 cleanup pattern — saves []rune allocations on rune variants)"
    - "Property tests in props_test.go cover: RangeBounds, Identity, Symmetric (byte), Symmetric (rune), NoNaN, NoInf, NoNegativeZero, GapSplitInvariance, RawNeverExceedsMatchTimesMinLen, MonotonicWithMatchReward — all pass under testing/quick (default 100 iterations)"
    - "FuzzSmithWatermanGotohScore harness in swg_fuzz_test.go has at least one programmatic seed covering substring-containment, identity, both-empty, one-empty, the gap-split canary, invalid UTF-8, and a multi-byte rune pair; fuzz body asserts no panic, no NaN, no Inf, score in [0,1]"
    - "testdata/fuzz/FuzzSmithWatermanGotohScore/seed-001 exists in the `go test fuzz v1` literal corpus format (matches testdata/fuzz/FuzzLevenshteinScore/seed-001 shape)"
    - "Benchmark file uses b.ReportAllocs() before b.ResetTimer(); six benches exist (ASCII_Short, ASCII_Medium, ASCII_Long, Unicode_Short, WithParams_ASCII_Short, RawScore_ASCII_Short); the two _Short variants report `0 B/op  0 allocs/op` in `go test -bench=... -benchmem -count=3`"
    - "Runtime alloc gate: a TestSmithWatermanGotohScore_ZeroAllocs_ASCII_Short test calls testing.AllocsPerRun(100, ...) and fails if the count exceeds 0"
    - "BDD feature tests/bdd/features/swg.feature exists with at least: canonical reference-vector Scenario Outline, identity scenario, both-empty scenario, one-empty scenario, symmetry scenario, and the gap-split canary scenario; `cd tests/bdd && go test ./...` exits 0"
    - "tests/bdd/steps/algorithms_steps.go appends SWG step bindings (iComputeTheSmithWatermanGotohScoreBetween / iComputeTheSecondSmithWatermanGotohScoreBetween / bothSmithWatermanGotohScoresShouldBeEqual) and their ctx.Step regex registrations inside InitializeScenario; the existing theScoreShouldBeApproximately/Exactly steps are algorithm-agnostic and reused (IN-06 lock — no SWG-specific Distance step needed)"
    - "testdata/golden/_staging/swg.json exists, produced by TestGolden_SmithWatermanGotoh_Staging via assertGoldenStaging; entries sorted alphabetically by Name; includes at minimum SmithWatermanGotoh_both_empty, SmithWatermanGotoh_identical, SmithWatermanGotoh_one_empty, SmithWatermanGotoh_two_substring, SmithWatermanGotoh_no_overlap, SmithWatermanGotoh_one_long_gap_canary"
    - "algoid_test.go contains a new TestDispatch_SmithWatermanGotohRegistered asserting dispatch[AlgoSmithWatermanGotoh] non-nil, AND the registered map in TestDispatch_UnregisteredSlotsAreNil adds int(AlgoSmithWatermanGotoh): true so the 'remaining nil slots' assertion still passes"
    - "ExampleSmithWatermanGotohScore and ExampleSmithWatermanGotohRawScore appended to example_test.go; their // Output: blocks match byte-for-byte (the Raw example pins the 12.0 raw value)"
    - "Coverage on swg.go ≥ 90%; coverage on the SWG public surface 100% (every exported symbol covered by a unit test or godoc example)"
    - "make check exits 0 at the end of this plan (lint, vet, race, coverage-check, verify-determinism, verify-license-headers, verify-no-runtime-deps, tidy-check, BDD)"
  artifacts:
    - path: "swg.go"
      provides: "SWGParams (type), NewSWGParams (constructor), SmithWatermanGotohScore/Runes/WithParams (normalised), SmithWatermanGotohRawScore/Runes/WithParams (unclamped) — six public functions, one public type, one public constructor; unexported swgDPRaw kernel (three-matrix two-row form)"
      min_lines: 220
      contains: "// Sources:"
    - path: "dispatch_swg.go"
      provides: "Package-load-time registration of SmithWatermanGotohScore into dispatch[AlgoSmithWatermanGotoh] (slot 6)"
      contains: "dispatch[AlgoSmithWatermanGotoh] = SmithWatermanGotohScore"
    - path: "swg_test.go"
      provides: "Unit tests for identity, both-empty, one-empty, substring-containment, canonical reference vectors (Smith-Waterman 1981 + Gotoh 1982 + Flouri 2015), the Gotoh-erratum gap-split canary, symmetry, byte vs rune equivalence on ASCII, multi-byte rune handling (Cyrillic / café), SWGParams default semantics, the Raw* unclamped path, AND the runtime allocation gate test"
    - path: "swg_bench_test.go"
      provides: "Six benchmarks: ASCII_Short, ASCII_Medium, ASCII_Long, Unicode_Short, WithParams_ASCII_Short, RawScore_ASCII_Short — alloc-asserted with b.ReportAllocs() + var sink anti-DCE"
    - path: "swg_fuzz_test.go"
      provides: "FuzzSmithWatermanGotohScore — panic-free, NaN/Inf-free, score-in-[0,1] invariant; programmatic seeds covering all CONTEXT.md §1 categories"
    - path: "props_test.go"
      provides: "Appended SWG block: TestProp_SmithWatermanGotohScore_RangeBounds, _Identity, _Symmetric, _NoNaN, _NoInf, _NoNegativeZero PLUS three SWG-specific TestProp_SmithWatermanGotoh_GapSplitInvariance / _RawNeverExceedsMatchTimesMinLen / _MonotonicWithMatchReward; rune-symmetry TestProp_SmithWatermanGotohScoreRunes_Symmetric appended at trailer (per WR-03 pattern)"
    - path: "example_test.go"
      provides: "Appended runnable godoc examples ExampleSmithWatermanGotohScore (substring pair → 1.0000) and ExampleSmithWatermanGotohRawScore (substring pair → 12.0)"
    - path: "algorithms_golden_test.go"
      provides: "Appended buildSWGStagingEntries + TestGolden_SmithWatermanGotoh_Staging — produces _staging/swg.json via assertGoldenStaging (locked signature from plan 02-01); merge list NOT updated here (plan 03-03 owns the merge)"
      contains: "TestGolden_SmithWatermanGotoh_Staging"
    - path: "algoid_test.go"
      provides: "Appended TestDispatch_SmithWatermanGotohRegistered; updated `registered` map in TestDispatch_UnregisteredSlotsAreNil flipping slot 6 to true"
      contains: "TestDispatch_SmithWatermanGotohRegistered"
    - path: "testdata/golden/_staging/swg.json"
      provides: "Per-algorithm staging file (Phase-2-locked pattern); merged into algorithms.json by plan 03-03; sorted by Name; canonical byte form via assertGoldenStaging → CanonicalMarshalForTest"
      contains: "SmithWatermanGotoh_one_long_gap_canary"
    - path: "testdata/fuzz/FuzzSmithWatermanGotohScore/seed-001"
      provides: "Fuzz seed corpus file in `go test fuzz v1` literal format (substring-containment pair as the default seed)"
    - path: "tests/bdd/features/swg.feature"
      provides: "Gherkin feature with scenarios: canonical reference-vector outline, identity, both-empty, one-empty, symmetry, gap-split canary (per PITFALLS.md §3 warning sign #2)"
    - path: "tests/bdd/steps/algorithms_steps.go"
      provides: "Appended SWG step methods on AlgorithmContext + SWG ctx.Step registrations inside InitializeScenario"
  key_links:
    - from: "dispatch_swg.go"
      to: "algoid.go (dispatch array, line 102 AlgoSmithWatermanGotoh declared)"
      via: "package-level var _ = func() bool { dispatch[AlgoSmithWatermanGotoh] = SmithWatermanGotohScore; return true }()"
      pattern: "dispatch\\[AlgoSmithWatermanGotoh\\]"
    - from: "swg.go SmithWatermanGotohScore"
      to: "swg.go SmithWatermanGotohScoreWithParams (delegation via NewSWGParams())"
      via: "no-params form calls *WithParams with NewSWGParams() (single source of truth for defaults — CONTEXT.md §3)"
      pattern: "SmithWatermanGotohScoreWithParams\\("
    - from: "swg.go swgDPRaw kernel"
      to: "stack buffer var buf [(maxStackInputLen+1)*6]float64 (3120 bytes)"
      via: "ASCII fast path gate `if lb <= maxStackInputLen && isASCII(a) && isASCII(b)`"
      pattern: "\\[\\(maxStackInputLen \\+ 1\\) \\* 6\\]float64"
    - from: "swg.go fast path"
      to: "normalise.go isASCII (lines 159-168, inherited from Phase 1)"
      via: "package-private call (same package fuzzymatch)"
      pattern: "isASCII\\("
    - from: "algorithms_golden_test.go TestGolden_SmithWatermanGotoh_Staging"
      to: "testdata/golden/_staging/swg.json"
      via: "assertGoldenStaging helper (LOCKED signature from plan 02-01 Task 3) writing via CanonicalMarshalForTest"
      pattern: "assertGoldenStaging\\(t, \"_staging/swg\\.json\""
    - from: "tests/bdd/features/swg.feature"
      to: "tests/bdd/steps/algorithms_steps.go"
      via: "godog step regex matching (`I compute the SmithWatermanGotoh score between`, `I compute the second SmithWatermanGotoh score between`, `both SmithWatermanGotoh scores should be equal` — and the algorithm-agnostic `the score should be approximately`/`exactly` steps registered once in the Levenshtein block)"
      pattern: "I compute the SmithWatermanGotoh score between"

user_setup: []
---

<objective>
Implement Smith-Waterman-Gotoh local-alignment similarity (CHAR-08) end-to-end as a single algorithm plan that replicates the canonical Phase 2 file-by-file pattern (plan 02-01-levenshtein-PLAN.md template) extended with three Phase-3-unique elements: (1) the corrected Gotoh affine-gap recurrence per Flouri et al. 2015 with all the documentary gates (file-level source citation, inline erratum note, gap-split property test, gap-split BDD canary); (2) the SWGParams value type + NewSWGParams() constructor — the first parameterised algorithm in the catalogue; (3) the expanded Raw* surface — three additional public functions (SmithWatermanGotohRawScore / *Runes / *WithParams) per CONTEXT.md §4 decision, with godoc explicitly contrasting raw-unclamped vs normalised-clamped semantics.

Purpose: deliver a deterministic, ASCII-fast-path, two-row-DP, fuzz-and-property-tested implementation that PITFALLS.md §3's four warning signs cannot trigger — identity short-circuits to 1.0, symmetry holds for byte and rune paths, the gap-split canary fires only when the implementation is buggy, and the monotonicity-with-Match-reward invariant pins the affine-gap behaviour. Cross-validation against the biopython reference corpus is a separate plan (03-02) that loads on top of this implementation.

Output: a working swg.go + dispatch_swg.go + test/bench/fuzz/BDD suite, a populated testdata/golden/_staging/swg.json staging file (Phase-2 pattern — merged into the canonical algorithms.json by plan 03-03), and the algoid_test.go / example_test.go / props_test.go / tests/bdd/steps/algorithms_steps.go append-points wired up so plan 03-02's cross-validation test attaches into a working surface.
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
@.planning/phases/03-smith-waterman-gotoh/03-CONTEXT.md
@.planning/phases/03-smith-waterman-gotoh/03-RESEARCH.md
@.planning/phases/03-smith-waterman-gotoh/03-PATTERNS.md
@.planning/phases/03-smith-waterman-gotoh/03-VALIDATION.md
@.planning/phases/02-core-character-algorithms-six/02-CONTEXT.md
@.planning/phases/02-core-character-algorithms-six/02-PATTERNS.md
@.planning/phases/02-core-character-algorithms-six/02-01-levenshtein-SUMMARY.md
@.planning/phases/02-core-character-algorithms-six/02-VERIFICATION.md
@.planning/phases/02-core-character-algorithms-six/02-07-finalisation-SUMMARY.md
@.planning/research/PITFALLS.md
@docs/requirements.md
@.claude/skills/algorithm-correctness-standards/SKILL.md
@.claude/skills/algorithm-licensing-standards/SKILL.md
@.claude/skills/performance-standards/SKILL.md
@.claude/skills/determinism-standards/SKILL.md
@.claude/skills/go-coding-standards/SKILL.md
@.claude/skills/go-testing-standards/SKILL.md
@.claude/skills/fuzzymatch-review-protocol/SKILL.md

<interfaces>
Phase 1 + Phase 2 contracts this plan consumes (extracted from existing source — DO NOT MODIFY or redeclare):

From algoid.go (line 102 — already declared, slot 6 of 23):
  const AlgoSmithWatermanGotoh AlgoID = ...  // index 6
  var dispatch [numAlgorithms]func(a, b string) float64  // unexported

From levenshtein.go (line 68 — SHARED package-level constant, inherited from Phase 2):
  const maxStackInputLen = 64
  // Stack buffer threshold for two-row DP algorithms; SWG references this by
  // name and MUST NOT redeclare.

From normalise.go (lines 159-168 — package-private helper, inherited from Phase 1):
  func isASCII(s string) bool   // true iff every byte < 0x80; "" returns true

From export_test.go (re-exports for the *_test.go files):
  func DispatchEntryNilForTest(i int) bool
  var CanonicalMarshalForTest = canonicalMarshal

From algorithms_golden_test.go (LOCKED signature from plan 02-01 Task 3):
  func assertGoldenStaging(t *testing.T, relPath string, v any)
  // relPath is relative to testdata/golden/, e.g. "_staging/swg.json"
  // Marshals v via CanonicalMarshalForTest, writes through WriteGoldenFile
  // when -update is set; otherwise asserts byte-equality against the existing file.
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

From tests/bdd/steps/algorithms_steps.go (extend-only; existing infrastructure):
  type AlgorithmContext struct { lastScore float64; lastScore2 float64; ... }
  func InitializeScenario(ctx *godog.ScenarioContext)
  // The following step regexes are already registered ONCE (Phase 2,
  // Levenshtein and Hamming blocks) and are algorithm-agnostic — SWG REUSES
  // them, do NOT re-register:
  //   ^the score should be approximately (\d+\.?\d*) within (\d+\.?\d*)$
  //   ^the score should be exactly (\d+\.?\d*)$
  //   ^the distance should be (\d+)$   (IN-06 — algorithm-agnostic; SWG has no Distance variant so no new distance step is required)

NEW public surface this plan introduces (consumed by plan 03-02 cross-validation + plan 03-03 finalisation):

  // SWGParams is a value type — not a pointer. Exported fields, no methods
  // required (CONTEXT.md §3). The Scorer layer (Phase 8) may add validation
  // later; this phase performs none.
  type SWGParams struct {
      Match     float64   // reward for a match, >= 0; default 1.0
      Mismatch  float64   // penalty for a mismatch, <= 0; default -1.0
      GapOpen   float64   // penalty for opening a gap, <= 0; default -1.5
      GapExtend float64   // penalty for extending a gap, <= 0; default -0.5
  }

  func NewSWGParams() SWGParams
  // Returns SWGParams{Match: 1.0, Mismatch: -1.0, GapOpen: -1.5, GapExtend: -0.5}.

  func SmithWatermanGotohScore(a, b string) float64
  func SmithWatermanGotohScoreRunes(a, b string) float64
  func SmithWatermanGotohScoreWithParams(a, b string, params SWGParams) float64
  func SmithWatermanGotohRawScore(a, b string) float64
  func SmithWatermanGotohRawScoreRunes(a, b string) float64
  func SmithWatermanGotohRawScoreWithParams(a, b string, params SWGParams) float64

  // Score normalisation (in *Score entry points):
  //   raw := swgDPRaw(...)
  //   norm := raw / float64(min(len(a), len(b)))
  //   return clamp(norm, 0, 1)
  // Documented in the *Score godoc per CONTEXT.md §4 (clamp warning); raw
  // value is what *RawScore exposes.
</interfaces>

<canonical_decisions_locked_for_this_plan>
The decisions this plan's executor must honour without re-deriving:

1. **Three-matrix two-row DP**: six rolling rows (prevM, currM, prevIx, currIx, prevIy, currIy), each length n+1. Stack-allocated via `var buf [(maxStackInputLen+1)*6]float64` (3120 bytes) on the ASCII fast path; six `make([]float64, n+1)` calls on the heap path. (CONTEXT.md §2, RESEARCH.md Pattern 1)
2. **Corrected Flouri 2015 initialisation**: all six border rows initialise to 0 for LOCAL alignment (NOT `-Inf` and NOT the global-alignment gap-open ladder — this is the Gotoh-erratum trap PITFALLS.md §3 warning sign #1 catches). (RESEARCH.md Pattern 1, init block)
3. **ASCII fast-path gate** is the Phase-2-locked idiom `if lb <= maxStackInputLen && isASCII(a) && isASCII(b)` (the LB-as-shorter-dimension swap happens BEFORE the gate). (02-PATTERNS.md Pattern 3, inherited)
4. **No params validation** in *Score / *RawScore: caller-supplied `GapOpen > 0`, NaN, +Inf, etc. produce deterministic-but-meaningless scores; algorithm functions never return errors and never panic on params. (CONTEXT.md §3, §5.11)
5. **No-params form delegates to *WithParams** via `NewSWGParams()` — single source of truth for defaults, no risk of drift. (CONTEXT.md §3)
6. **Raw* surface is intentional** — three additional functions (Raw, RawRunes, RawWithParams) beyond docs/requirements.md §7.1.8's current 3-function spec. Plan 03-03 updates the spec doc; this plan implements + tests + documents the surface. (CONTEXT.md §4)
7. **Identity short-circuit** on EVERY public entry point (`if a == b { return 1.0 }` on Score, `if a == b { return Match * float64(len([]rune(a))) }` doesn't apply — the *RawScore identity case returns 0.0 for both-empty and `Match * float64(len(a))` for non-empty `x == x`; actually for *RawScore identity-with-non-empty the value is `Match * len(x)` because every position matches with no gaps — see <action> for the exact rule). (IN-02 pattern from Phase 2 cleanup)
8. **No `math.Pow`, `math.Log`, `math.Exp`, `math.FMA`** anywhere in swg.go — only +, -, *, /, comparisons, float64() cast, optional math.Abs / math.Min / math.Max if needed (none required for this kernel). (determinism-standards §13.3)
9. **No init() functions; dispatch registration via** `var _ = func() bool { dispatch[AlgoSmithWatermanGotoh] = SmithWatermanGotohScore; return true }()`. (02-PATTERNS.md Pattern 6, inherited LOCKED)
10. **Staging-golden file uses Phase-2-locked assertGoldenStaging helper** — do not re-implement the helper. (02-01-levenshtein-PLAN.md §canonical_pattern_decisions item 7)
11. **Files are append-only** for props_test.go, example_test.go, algoid_test.go, algorithms_golden_test.go, tests/bdd/steps/algorithms_steps.go — locate the insertion point per 03-PATTERNS.md and APPEND a SWG block; never rewrite existing test bodies.
</canonical_decisions_locked_for_this_plan>
</context>

<tasks>

<task type="auto" tdd="true">
  <name>Task 1: Implement swg.go (algorithm + SWGParams) and dispatch_swg.go (registration)</name>
  <files>swg.go, dispatch_swg.go</files>
  <read_first>
    - levenshtein.go (LOCKED Phase 2 template: Apache-2.0 header lines 1-13, file-level source-citation godoc lines 15-55, package declaration, maxStackInputLen line 68 — DO NOT redeclare, public *Score/*Distance/*ScoreRunes signatures lines 84-200, two-row DP kernel structure)
    - dispatch_levenshtein.go (character-for-character template for dispatch_swg.go — only AlgoXxx + XxxScore identifiers change)
    - damerau_full.go (closest analog for a more complex DP kernel using helpers — pattern for separating Byte and Rune kernels)
    - normalise.go lines 1-13 (Apache-2.0 header — verbatim source), lines 159-168 (isASCII — referenced by name, DO NOT redeclare)
    - algoid.go (AlgoSmithWatermanGotoh declared at slot 6; dispatch[] array unexported)
    - .planning/phases/03-smith-waterman-gotoh/03-PATTERNS.md (`swg.go` and `dispatch_swg.go` blocks — concrete identifier templates)
    - .planning/phases/03-smith-waterman-gotoh/03-RESEARCH.md §Pattern 1 (the corrected Flouri 2015 recurrence and the illustrative `swgDPRaw` kernel; §Pattern 2 the ASCII fast-path gate code; §Pattern 4 the rune variant pattern)
    - .planning/phases/03-smith-waterman-gotoh/03-CONTEXT.md §1, §2, §3, §4, §7 (locked decisions for params shape, two-row DP day 1, Raw* surface, inherited patterns)
    - .planning/research/PITFALLS.md §3 (Gotoh erratum — the four warning signs the implementation must clear)
    - docs/requirements.md §7.1.8 (current public API spec — note this plan EXPANDS it; plan 03-03 updates the doc)
    - .claude/skills/algorithm-correctness-standards/SKILL.md (primary-source citation format, formula-in-godoc rule, constants-with-paper-reference rule)
    - .claude/skills/algorithm-licensing-standards/SKILL.md (fresh-implementation rule: write from the corrected primary source — Smith-Waterman 1981 + Gotoh 1982 + Flouri 2015; no code copied from biopython / EMBOSS / any Go port; cross-validation only, in plan 03-02)
    - .claude/skills/performance-standards/SKILL.md (two-row DP requirement, ASCII-fast-path rule, allocation budget)
    - .claude/skills/determinism-standards/SKILL.md (no math.Pow/Log/Exp/FMA, only +-*/sqrt; no map iteration on output paths; sum reductions left-to-right)
  </read_first>
  <action>
Create swg.go in package fuzzymatch (root module) with the following structure:

1. Copy the Apache-2.0 license header verbatim from normalise.go lines 1-13 (byte-for-byte).

2. File-level godoc block (per 03-PATTERNS.md `swg.go` section):
   - Opening sentence: `// swg.go implements the Smith-Waterman-Gotoh local-alignment similarity with affine gap penalty for the fuzzymatch catalogue.`
   - Sources block listing all three primary references explicitly:
     - Smith, T. F. & Waterman, M. S. (1981) — local-alignment formulation, J. Mol. Biol. 147:195-197
     - Gotoh, O. (1982) — affine-gap O(mn) reduction, J. Mol. Biol. 162:705-708
     - Flouri, T. et al. (2015) — "Are all global alignment algorithms and implementations correct?" biorxiv 031500 — documents the Gotoh 1982 initialisation erratum
   - Explicit erratum statement: `// Gotoh 1982 contains a known erratum in the affine-gap initialisation step (the global-alignment border setup that textbook treatments often blur into local alignment); this implementation uses the corrected formulation per Flouri et al. 2015: every border cell of M, Ix, Iy initialises to 0 for LOCAL alignment.`
   - Recurrence block in godoc using the M/Ix/Iy equations from 03-RESEARCH.md Pattern 1 lines 252-265 (verbatim text, in a comment block — not executable code).
   - Implementation discipline list mirrored from levenshtein.go style: ASCII fast path with stack buffer, no init(), no map iteration on output paths, no transcendental math, no goroutines/channels/mutexes.

3. SWGParams type — exact shape from CONTEXT.md §3:
   - struct with four exported float64 fields Match, Mismatch, GapOpen, GapExtend.
   - Each field has a godoc line stating its sign convention and default (Match=1.0, Mismatch=-1.0, GapOpen=-1.5, GapExtend=-0.5) and the conventional ordering `GapOpen <= GapExtend <= 0`.
   - No methods on SWGParams; value type, callers mutate fields after construction.

4. NewSWGParams() constructor — returns `SWGParams{Match: 1.0, Mismatch: -1.0, GapOpen: -1.5, GapExtend: -0.5}`. Godoc states: returns a fresh copy with documented defaults; callers may override individual fields after construction.

5. Public entry points (six functions — CONTEXT.md §3 + §4):

   - SmithWatermanGotohScore(a, b string) float64 — delegates to SmithWatermanGotohScoreWithParams(a, b, NewSWGParams()). Godoc per CONTEXT.md §4 sample: opening doc line, clamp warning, normalisation formula, edge-case list (both-empty=1.0, one-empty=0.0, identical=1.0, symmetric).
   - SmithWatermanGotohScoreRunes(a, b string) float64 — identity short-circuit `if a == b { return 1.0 }` (saves []rune allocations on identity per IN-02), then convert to []rune and call the rune-aware kernel with default params.
   - SmithWatermanGotohScoreWithParams(a, b string, params SWGParams) float64 — identity short-circuit, empty-guards, swap so b is shorter, ASCII-fast-path gate, call swgDPRaw, normalise via `raw / float64(min(la, lb))` and clamp to [0, 1].
   - SmithWatermanGotohRawScore(a, b string) float64 — delegates to SmithWatermanGotohRawScoreWithParams(a, b, NewSWGParams()). Godoc states the value is UNCLAMPED raw alignment score and may be negative or > 1; recommends *Score for the normalised value.
   - SmithWatermanGotohRawScoreRunes(a, b string) float64 — identity short-circuit returns `params.Match * float64(len([]rune(a)))` for non-empty x; for both-empty returns 0.0; otherwise converts to []rune and calls the rune kernel.
   - SmithWatermanGotohRawScoreWithParams(a, b string, params SWGParams) float64 — same shape as the normalised *WithParams but returns the raw kernel output WITHOUT the clamp/normalise step (an empty-input case returns 0.0; an `a == b` non-empty case returns `params.Match * float64(len(a))` — for the byte path; the rune analog uses `len([]rune(a))` — see the dedicated rune entry).

   For the byte-path *Score / *RawScore: ensure the la-vs-lb swap happens ONCE inside the *Score / *RawScore body BEFORE entering the ASCII gate, and that swgDPRaw is called with `la >= lb > 0` invariant.

   Identity short-circuit on EVERY *Score entry point: `if a == b { return 1.0 }` (this also covers the both-empty case since "" == "" is true).
   Identity short-circuit on EVERY *RawScore entry point: special-case — for *RawScore, identity-with-non-empty returns `params.Match * float64(len(a))` (every position matches, no gaps); identity-with-both-empty returns 0.0 (no positions to score). Document this in the *RawScore godoc.

6. swgDPRaw unexported kernel — implement the three-matrix two-row form from 03-RESEARCH.md Pattern 1 lines 283-339 fresh, transcribing the equations only (no code copied from biopython / EMBOSS / any Go port). Signature:

       func swgDPRaw(a, b string, m, n int, params SWGParams,
                     prevM, currM, prevIx, currIx, prevIy, currIy []float64) float64

   The kernel must:
   - Zero-initialise all six rows before the outer loop (Flouri 2015 corrected init for LOCAL alignment).
   - For each i in 1..m: zero currM[0]/currIx[0]/currIy[0]; for each j in 1..n: compute the M/Ix/Iy recurrences max-with-0; track bestRaw = max of all M[i][j] cells; swap prev/curr at end of i.
   - Return bestRaw.

   A separate swgDPRawRunes kernel mirrors the byte kernel operating on []rune slices (callers convert eagerly per Pattern 4); document the rune kernel's allocation cost (two []rune slices at the entry point per CONTEXT.md §2).

7. Stack-buffer fast path inside *ScoreWithParams / *RawScoreWithParams (byte path): `var buf [(maxStackInputLen + 1) * 6]float64` (3120 bytes); slice it into six length-(n+1) rows via `n1 := lb + 1; buf[0*n1:1*n1]` ... `buf[5*n1:6*n1]`; pass to swgDPRaw.

8. Heap fallback when `lb > maxStackInputLen || !isASCII(a) || !isASCII(b)`: six `make([]float64, lb+1)` calls passed to swgDPRaw. Expected alloc cost: 6 allocs/op on byte heap path; 8 on rune path (the two []rune + six rows).

9. NO new top-level constants. maxStackInputLen is inherited from levenshtein.go:68 — reference by name, do NOT redeclare. isASCII is inherited from normalise.go:159-168 — reference by name, do NOT redeclare. Build will fail with a "redeclared" error if either is duplicated.

10. NO init() function. NO map iteration. NO `math.Pow/Log/Exp/FMA`. NO goroutines/channels/mutexes. NO `[]byte(a)` conversions on the hot path — use direct byte indexing `a[i-1]`.

Create dispatch_swg.go in package fuzzymatch:

1. Copy Apache-2.0 header verbatim from dispatch_levenshtein.go lines 1-13.
2. File-level godoc: `// dispatch_swg.go registers SmithWatermanGotohScore into the dispatch table at package load time. Sole writer to dispatch[AlgoSmithWatermanGotoh] (slot 6).`
3. Body — exactly the LOCKED idiom from dispatch_levenshtein.go (only AlgoXxx + XxxScore identifiers change):

       var _ = func() bool {
           dispatch[AlgoSmithWatermanGotoh] = SmithWatermanGotohScore
           return true
       }()

4. Nothing else — file is registration-only; no helpers, no constants.

Run `go build ./...` to confirm the package compiles; run `bash scripts/verify-license-headers.sh`; run `go vet ./...`.
  </action>
  <verify>
    <automated>go build ./... && go vet ./... && bash scripts/verify-license-headers.sh && grep -c 'maxStackInputLen' swg.go | xargs -I{} test {} -ge 1 && ! grep -Eq 'math\.(Pow|Log|Exp|FMA)' swg.go && ! grep -Eq '^func init\(' swg.go dispatch_swg.go && grep -q 'dispatch\[AlgoSmithWatermanGotoh\] = SmithWatermanGotohScore' dispatch_swg.go</automated>
  </verify>
  <acceptance_criteria>
    - swg.go starts with the Apache-2.0 header (matches normalise.go lines 1-13 byte-for-byte; verifiable by `diff <(head -13 normalise.go) <(head -13 swg.go)` exit 0).
    - swg.go file-level godoc contains literal `// Sources:` AND `Smith, T. F. & Waterman, M. S. (1981)` AND `Gotoh, O. (1982)` AND `Flouri, T. et al. (2015)` (the three primary references — all three required for the PITFALLS.md §3 documentary gate).
    - swg.go file-level godoc contains the literal string `Gotoh 1982 contains a known erratum` (verifiable by `grep -c "Gotoh 1982 contains a known erratum" swg.go` returning ≥ 1).
    - swg.go declares the SWGParams struct with exactly the four exported float64 fields: Match, Mismatch, GapOpen, GapExtend (verifiable by `grep -E 'Match\s+float64' swg.go` and three sibling matches; or by parsing with `go doc ./... SWGParams`).
    - swg.go declares NewSWGParams() returning the documented defaults — verifiable by a Go test in Task 2 (TestSWG_NewSWGParams_Defaults).
    - swg.go declares all six public entry points: SmithWatermanGotohScore, SmithWatermanGotohScoreRunes, SmithWatermanGotohScoreWithParams, SmithWatermanGotohRawScore, SmithWatermanGotohRawScoreRunes, SmithWatermanGotohRawScoreWithParams. Verifiable by `grep -cE '^func SmithWatermanGotoh(Score|RawScore)(Runes|WithParams)?\b' swg.go` returning exactly 6.
    - swg.go does NOT redeclare maxStackInputLen or isASCII; verifiable by `grep -E '^const maxStackInputLen\b|^func isASCII\b' swg.go` returning no matches.
    - swg.go references maxStackInputLen at least once (for the stack-buffer fast path); verifiable by `grep -c 'maxStackInputLen' swg.go` returning ≥ 1.
    - swg.go contains the literal stack-buffer declaration `var buf [(maxStackInputLen + 1) * 6]float64` (verifiable by `grep -F 'var buf [(maxStackInputLen + 1) * 6]float64' swg.go` returning ≥ 1).
    - `grep -E 'math\.(Pow|Log|Exp|FMA)' swg.go` returns no matches (transcendentals + FMA prohibited per determinism-standards §13.3).
    - `grep -vE '^[[:space:]]*//' swg.go | grep -cE '^func init\(' ` returns 0 (no init() function — comments naming `init` are allowed). Also `grep -vE '^[[:space:]]*//' dispatch_swg.go | grep -cE '^func init\(' ` returns 0.
    - `grep -c '\[\]byte(' swg.go` returns 0 (no string-to-bytes conversion on the hot path).
    - dispatch_swg.go contains the literal text `dispatch[AlgoSmithWatermanGotoh] = SmithWatermanGotohScore` exactly once.
    - dispatch_swg.go uses the `var _ = func() bool {` registration idiom (NOT `init()`); verifiable by `grep -c 'var _ = func() bool' dispatch_swg.go` returning 1.
    - `go build ./...` exits 0.
    - `go vet ./...` exits 0.
    - `bash scripts/verify-license-headers.sh` exits 0.
  </acceptance_criteria>
  <behavior>
    - NewSWGParams() returns SWGParams{1.0, -1.0, -1.5, -0.5} byte-for-byte (verifiable by struct equality in Task 2).
    - SmithWatermanGotohScore("http_request", "http_request_header_fields") == 1.0 (substring fully contained → clamp returns 1).
    - SmithWatermanGotohScore("", "") == 1.0 (both-empty identity short-circuit).
    - SmithWatermanGotohScore("abc", "") == 0.0 (one-empty).
    - SmithWatermanGotohScore("abc", "abc") == 1.0 (identity short-circuit).
    - SmithWatermanGotohScore(a, b) == SmithWatermanGotohScore(b, a) for any a, b (symmetric).
    - SmithWatermanGotohRawScore("http_request", "http_request_header_fields") == 12.0 (12 match positions, no gap penalty; Match=1.0).
    - SmithWatermanGotohRawScore("abc", "abc") == 3.0 (identity short-circuit returns Match * len(a)).
    - SmithWatermanGotohRawScore("", "") == 0.0 (both-empty raw returns 0).
    - SmithWatermanGotohScoreWithParams("hello", "hello", NewSWGParams()) == 1.0.
    - dispatch[AlgoSmithWatermanGotoh] is non-nil and equals SmithWatermanGotohScore after package load (verified by Task 2's TestDispatch_SmithWatermanGotohRegistered).
  </behavior>
  <done>
    swg.go and dispatch_swg.go committed. Package compiles. License-header verifier passes. SWGParams + NewSWGParams + 6 public functions + dispatch registration in place. Inherited maxStackInputLen + isASCII referenced (not redeclared). No init() / no transcendentals / no map iteration. Ready for Task 2 to write the test suite that exercises every public symbol.
  </done>
</task>

<task type="auto" tdd="true">
  <name>Task 2: Tests (unit + property + benchmark + fuzz + dispatch + example) and the runtime alloc gate</name>
  <files>swg_test.go, swg_bench_test.go, swg_fuzz_test.go, props_test.go, example_test.go, algoid_test.go, testdata/fuzz/FuzzSmithWatermanGotohScore/seed-001</files>
  <read_first>
    - swg.go (the implementation from Task 1 — confirm exact identifier names before writing tests against them)
    - levenshtein_test.go (LOCKED template: file header, fuzzymatch_test package, table-driven unit tests, math.Abs tolerance pattern, no testify)
    - levenshtein_bench_test.go (alloc-aware benchmark template: b.ReportAllocs() BEFORE b.ResetTimer(); var sink anti-DCE pattern; ASCII Short/Medium/Long + Unicode Short labels)
    - levenshtein_fuzz_test.go (FuzzXxx skeleton: f.Add programmatic seeds + f.Fuzz body asserting no panic / no NaN / no Inf / score in [0,1])
    - props_test.go (existing property-test file — locate the insertion point at the "Rune-path symmetry property tests" boundary at line 727; APPEND the SWG block above this boundary; append the rune-symmetry property test at the trailer per WR-03)
    - example_test.go (existing godoc-example file — append two new ExampleSmithWatermanGotoh* functions at EOF after the last existing example)
    - algoid_test.go (existing — locate TestDispatch_UnregisteredSlotsAreNil at line ~289 and the `registered` map literal at lines ~292-299; flip slot AlgoSmithWatermanGotoh to true; add a new TestDispatch_SmithWatermanGotohRegistered immediately BEFORE TestDispatch_UnregisteredSlotsAreNil)
    - testdata/fuzz/FuzzLevenshteinScore/seed-001 (corpus format reference for the new SWG seed file)
    - .planning/phases/03-smith-waterman-gotoh/03-PATTERNS.md (`swg_test.go`, `swg_bench_test.go`, `swg_fuzz_test.go`, `props_test.go`, `example_test.go`, `algoid_test.go` sections — concrete templates)
    - .planning/phases/03-smith-waterman-gotoh/03-CONTEXT.md §5 (property-test inventory: 6 standard inherited from Phase 2 + 3 SWG-specific canaries; rune symmetry as separate trailer function)
    - .planning/research/PITFALLS.md §3 (the four warning signs the property tests must clear)
    - .claude/skills/go-testing-standards/SKILL.md (coverage floors, property-test conventions, no testify in root tests)
    - .claude/skills/determinism-standards/SKILL.md (DET-04 invariants: NoNaN, NoInf, NoNegativeZero)
  </read_first>
  <action>
**Create swg_test.go (package fuzzymatch_test, stdlib `testing` only — NO testify):**

1. Apache-2.0 header from normalise.go lines 1-13.
2. File-level godoc per 03-PATTERNS.md swg_test.go section, opening with: `// swg_test.go pins the public-API contract of swg.go: identity, both-empty, one-empty, canonical reference vectors from Smith-Waterman 1981 / Gotoh 1982 (corrected per Flouri et al. 2015), symmetry, byte vs rune path equivalence on ASCII, multi-byte rune handling, NaN/Inf guards, the SWGParams construction and default semantics, and the Raw* unclamped surface.`
3. Imports: `math`, `testing`, `github.com/axonops/fuzzymatch`. (TestSWG_CrossValidation in plan 03-02 adds encoding/json, os, path/filepath — not in this plan.)
4. Tests (each a separate `func TestSWG_<Topic>(t *testing.T)` or `func TestSmithWatermanGotoh_<Topic>` — pick the prefix consistent with the existing levenshtein_test.go convention; choose `TestSmithWatermanGotoh_<Topic>` to keep the testname-vs-algorithm-name relation explicit):
   - TestSmithWatermanGotoh_BothEmpty — Score("", "") == 1.0; RawScore("", "") == 0.0; ScoreWithParams("", "", custom) == 1.0.
   - TestSmithWatermanGotoh_OneEmpty — Score("", "abc") == 0.0; Score("abc", "") == 0.0; RawScore variants likewise 0.0.
   - TestSmithWatermanGotoh_Identical — Score("abc", "abc") == 1.0; RawScore("abc", "abc") == 3.0 (Match=1.0 * 3 positions); Score("user_id", "user_id") == 1.0.
   - TestSmithWatermanGotoh_SubstringContainment — Score("http_request", "http_request_header_fields") == 1.0; RawScore("http_request", "http_request_header_fields") == 12.0 (12 match positions).
   - TestSmithWatermanGotoh_NoOverlap — Score("qqqq", "zzzz") returns a score in [0, ~0.5] (no matches → raw is 0 or negative → clamp returns 0.0). Verify exact value: with default params (mismatch=-1, gap=-1.5/-0.5), the best raw alignment is 0 (zero-init local-alignment property); normalised = 0/min(4,4) = 0.0. Assert exact 0.0.
   - TestSmithWatermanGotoh_Symmetry — Score(a, b) == Score(b, a) for the reference-vector pairs.
   - TestSmithWatermanGotoh_ByteVsRune_Equivalence — for an ASCII-only pair, Score and ScoreRunes return identical results.
   - TestSmithWatermanGotoh_RuneMultiByte — Score with Cyrillic input ("Привет", "привет") — exact comparison or "differs by case but rune-aware"; assert the rune path returns a deterministic value (do NOT hardcode without computing; pick a stable expected like `ScoreRunes("Привет", "Привет") == 1.0` for the safe identity assertion and `ScoreRunes("café", "cafe") < 1.0 && > 0.0` for the case-differs-by-one-rune assertion).
   - TestSmithWatermanGotoh_NewSWGParams_Defaults — assert NewSWGParams() == SWGParams{Match: 1.0, Mismatch: -1.0, GapOpen: -1.5, GapExtend: -0.5} byte-for-byte (struct equality).
   - TestSmithWatermanGotoh_WithCustomParams — ScoreWithParams("hello", "hallo", SWGParams{Match: 2.0, Mismatch: -2.0, GapOpen: -3.0, GapExtend: -1.0}) returns a deterministic value (compute by hand or assert it falls in [0,1] and matches a pre-computed expected derived from running the kernel on paper); also assert it is finite (not NaN, not Inf).
   - TestSmithWatermanGotoh_GapSplitCanary — Score("abc________def", "abcdef") <= Score("abc____def", "abcdef") AND >= Score("abcdef", "abcdef") - epsilon; specifically the long-gap pair MUST NOT score higher than the single-substring case (this is the PITFALLS.md §3 warning sign #2 gate in unit-test form; the property test in props_test.go covers the general invariant).
   - TestSmithWatermanGotoh_RawScore_UnclampedNegative — for two no-overlap strings with custom Match=0.1 and Mismatch=-10, RawScoreWithParams returns a negative value (verifying the raw surface is NOT clamped, contrasting with *Score which would return 0).
   - TestSmithWatermanGotohScore_ZeroAllocs_ASCII_Short — uses testing.AllocsPerRun(100, ...) to assert SmithWatermanGotohScore("kitten", "sitting") allocates 0; fails if AllocsPerRun > 0. This is the runtime alloc gate that pins PERF-01 inside the test suite (not just the benchmark).

   Float comparisons use stdlib `math.Abs(got - want) <= 1e-9` (no helper redeclaration).

**Create swg_bench_test.go (package fuzzymatch_test):**

1. Apache-2.0 header.
2. File-level godoc per 03-PATTERNS.md swg_bench_test.go section listing the six benchmarks and their alloc targets (ASCII_Short / ASCII_Medium = 0 allocs; ASCII_Long = 6 allocs heap rows; Unicode_Short = 8 allocs; WithParams_ASCII_Short = 0 allocs; RawScore_ASCII_Short = 0 allocs).
3. Benchmarks (every benchmark calls b.ReportAllocs() then b.ResetTimer() in that order; uses `var sink float64` and `if sink < 0 { b.Fatal(...) }` to defeat DCE):
   - BenchmarkSmithWatermanGotohScore_ASCII_Short — Score("kitten", "sitting") in the loop.
   - BenchmarkSmithWatermanGotohScore_ASCII_Medium — Score against a `const a50 = ...; const b50 = ...` pair (50-char ASCII identifiers; choose two distinct but similar identifiers).
   - BenchmarkSmithWatermanGotohScore_ASCII_Long — Score against two 500-char ASCII strings (heap path; allocations expected).
   - BenchmarkSmithWatermanGotohScore_Unicode_Short — Score against a multi-byte UTF-8 pair (rune path through ScoreRunes is the closest analog — choose Cyrillic; allocations expected from []rune conversion).
   - BenchmarkSmithWatermanGotohScore_WithParams_ASCII_Short — declare a `params := fuzzymatch.NewSWGParams(); params.Match = 2.0` outside the loop; call ScoreWithParams in the loop.
   - BenchmarkSmithWatermanGotohRawScore_ASCII_Short — RawScore in the loop (exercises the unclamped path).

**Create swg_fuzz_test.go (package fuzzymatch_test):**

1. Apache-2.0 header + file-level godoc per 03-PATTERNS.md swg_fuzz_test.go section.
2. FuzzSmithWatermanGotohScore — programmatic seeds (at MINIMUM all of CONTEXT.md §1 + RESEARCH.md required-case categories):
   - {"kitten", "sitting"} canonical reference vector
   - {"http_request", "http_request_header_fields"} substring containment
   - {"abc", "abc"} identical
   - {"", "abc"} one-empty
   - {"", ""} both-empty
   - {"abc________def", "abcdef"} one-long-gap canary (Gotoh-erratum gate)
   - {"\xff\xfe", "abc"} invalid UTF-8 (byte-path resilience)
   - {"Привет", "привет"} Cyrillic multi-byte UTF-8
   Fuzz body asserts: implicit no-panic; `!math.IsNaN(got)`; `!math.IsInf(got, 0)`; `got >= 0.0 && got <= 1.0`; error messages include the input pair for triage.

**Create testdata/fuzz/FuzzSmithWatermanGotohScore/seed-001** in the `go test fuzz v1` literal format (mirrors testdata/fuzz/FuzzLevenshteinScore/seed-001 shape):

       go test fuzz v1
       string("http_request")
       string("http_request_header_fields")

The substring-containment pair is the default seed (exercises the SWG-vs-Levenshtein divergence plan 03-03's cross-algorithm consistency test pins).

**Modify props_test.go — append the SWG block above the "Rune-path symmetry property tests" boundary (line ~727), and append the rune-symmetry property test at the trailer:**

1. Locate the boundary comment block:

       // ---------------------------------------------------------------------------
       // Rune-path symmetry property tests (WR-03)
       // ---------------------------------------------------------------------------

   Append the SWG block IMMEDIATELY BEFORE this boundary (so the existing rune-symmetry separator and its functions remain at the file trailer).

2. SWG block contents (function bodies use testing/quick.Check; each test follows the Phase-2 template — see 03-PATTERNS.md props_test.go section):
   - TestProp_SmithWatermanGotohScore_RangeBounds — score in [0,1] for any a, b.
   - TestProp_SmithWatermanGotohScore_Identity — Score(x, x) == 1.0 for non-empty x.
   - TestProp_SmithWatermanGotohScore_Symmetric — Score(a, b) == Score(b, a).
   - TestProp_SmithWatermanGotohScore_NoNaN — !math.IsNaN(Score(a, b)).
   - TestProp_SmithWatermanGotohScore_NoInf — !math.IsInf(Score(a, b), 0).
   - TestProp_SmithWatermanGotohScore_NoNegativeZero — when Score == 0.0, !math.Signbit(Score).
   - TestProp_SmithWatermanGotoh_GapSplitInvariance — hand-curated triples exercising the Gotoh-erratum case: splitting a single long gap into two halves with intervening match characters that don't affect the local alignment must NOT improve the score. Use a small fixed table (not testing/quick — the property is hand-crafted per CONTEXT.md §5); assert Score("abc________def", "abcdef") <= Score("abcdef", "abcdef") (trivially: clamped to ≤ 1.0); add stricter triples like Score("aXXXXXXXXb", "ab") <= Score("aXXXXb", "ab") if the gap-split direction.
   - TestProp_SmithWatermanGotoh_RawNeverExceedsMatchTimesMinLen — testing/quick: for random a, b: RawScore(a, b) <= NewSWGParams().Match * float64(min(len(a), len(b))). Skip if min == 0.
   - TestProp_SmithWatermanGotoh_MonotonicWithMatchReward — testing/quick: for random a, b: increasing only the Match field (e.g. from 1.0 to 2.0) must NOT decrease RawScoreWithParams. Skip degenerate cases where min == 0.

3. Rune-symmetry trailer — append `TestProp_SmithWatermanGotohScoreRunes_Symmetric` at the very end of props_test.go (after the existing rune-symmetry block for the six Phase-2 algorithms, line ~793). Body:

       func TestProp_SmithWatermanGotohScoreRunes_Symmetric(t *testing.T) {
           f := func(a, b string) bool {
               return fuzzymatch.SmithWatermanGotohScoreRunes(a, b) == fuzzymatch.SmithWatermanGotohScoreRunes(b, a)
           }
           if err := quick.Check(f, nil); err != nil {
               t.Errorf("SmithWatermanGotohScoreRunes not symmetric: %v", err)
           }
       }

   NO triangle-inequality test (SWG is not a metric — CONTEXT.md §5 explicit exclusion).

**Modify example_test.go — append two new ExampleSmithWatermanGotoh* functions at EOF:**

1. ExampleSmithWatermanGotohScore — uses substring-containment pair to print `1.0000` (the locked Output value from CONTEXT.md §4 / 03-PATTERNS.md example_test.go section).
2. ExampleSmithWatermanGotohRawScore — same pair to print `12.0`.
3. Each example uses `fmt.Printf("%.4f\n", ...)` / `fmt.Printf("%.1f\n", ...)` matching the Output block precision.

**Modify algoid_test.go — append TestDispatch_SmithWatermanGotohRegistered + flip slot 6 in the `registered` map:**

1. Insert TestDispatch_SmithWatermanGotohRegistered IMMEDIATELY BEFORE TestDispatch_UnregisteredSlotsAreNil (around line 289):

       func TestDispatch_SmithWatermanGotohRegistered(t *testing.T) {
           if fuzzymatch.DispatchEntryNilForTest(int(fuzzymatch.AlgoSmithWatermanGotoh)) {
               t.Errorf("dispatch[AlgoSmithWatermanGotoh] (%d) is nil — dispatch_swg.go must register SmithWatermanGotohScore at package load time",
                   int(fuzzymatch.AlgoSmithWatermanGotoh))
           }
       }

2. Flip slot AlgoSmithWatermanGotoh to true in the `registered` map literal inside TestDispatch_UnregisteredSlotsAreNil (lines 292-299):

       int(fuzzymatch.AlgoSmithWatermanGotoh):     true,  // ADD THIS LINE

   Add it in alphabetical/numeric order alongside the existing 6 entries; the godoc above the function (lines 284-288) should be updated to mention AlgoSmithWatermanGotoh (slot 6) is now registered.

**Run the test suite:**

       go test -race -shuffle=on -count=1 -run 'TestSmithWatermanGotoh|TestProp_SmithWatermanGotoh|TestDispatch_SmithWatermanGotoh|ExampleSmithWatermanGotoh' ./...
       go test -bench=BenchmarkSmithWatermanGotohScore_ASCII -benchmem -run=^$ -count=3 ./...
       go test -fuzz=FuzzSmithWatermanGotohScore -fuzztime=30s -run=^$ ./...

The first must exit 0. The second must report `0 B/op  0 allocs/op` for the _Short and _Medium variants (the _Long and _Unicode variants may allocate; _WithParams_Short and _RawScore_Short must also be 0). The fuzz run is informational locally; verify it doesn't crash.
  </action>
  <verify>
    <automated>go test -race -shuffle=on -count=1 -run 'TestSmithWatermanGotoh|TestProp_SmithWatermanGotoh|TestDispatch_SmithWatermanGotoh|ExampleSmithWatermanGotoh' ./... && go test -bench=BenchmarkSmithWatermanGotohScore_ASCII_Short -benchmem -run=^$ -count=3 ./... 2>&1 | grep -E '0 B/op[[:space:]]+0 allocs/op'</automated>
  </verify>
  <acceptance_criteria>
    - All TestSmithWatermanGotoh_* unit tests pass.
    - All TestProp_SmithWatermanGotoh* property tests pass under testing/quick (default 100 random invocations; the gap-split-invariance hand-curated table also passes).
    - TestSmithWatermanGotohScore_ZeroAllocs_ASCII_Short reports 0 allocations via testing.AllocsPerRun (the runtime alloc gate).
    - BenchmarkSmithWatermanGotohScore_ASCII_Short, _ASCII_Medium, _WithParams_ASCII_Short, _RawScore_ASCII_Short EACH report `0 B/op  0 allocs/op` in `go test -bench=... -benchmem -count=3` (verifiable by grepping the benchstat-style output for each benchmark name).
    - ExampleSmithWatermanGotohScore Output block matches `1.0000\n` byte-for-byte (test exit code 0).
    - ExampleSmithWatermanGotohRawScore Output block matches `12.0\n` byte-for-byte.
    - FuzzSmithWatermanGotohScore completes 30s of fuzzing without panic or invariant violation (informational locally; CI runs longer windows per plan 01-02).
    - testdata/fuzz/FuzzSmithWatermanGotohScore/seed-001 exists and parses as `go test fuzz v1` corpus (verifiable by `head -1 testdata/fuzz/FuzzSmithWatermanGotohScore/seed-001` returning `go test fuzz v1`).
    - algoid_test.go's renamed/updated TestDispatch_* assertions pass: TestDispatch_SmithWatermanGotohRegistered passes; TestDispatch_UnregisteredSlotsAreNil still passes with slot 6 in the registered map.
    - `grep -c '"github.com/stretchr/testify' swg_test.go swg_bench_test.go swg_fuzz_test.go props_test.go example_test.go` returns 0 (no testify in root tests).
    - `grep -c 'TestProp_SmithWatermanGotoh' props_test.go` returns at least 9 (6 standard + 3 SWG-specific + 1 rune-symmetry).
    - `grep -c 'ExampleSmithWatermanGotoh' example_test.go` returns at least 2 (Score and RawScore examples).
    - `grep -c 'TestDispatch_SmithWatermanGotohRegistered' algoid_test.go` returns 1.
    - `grep -c 'AlgoSmithWatermanGotoh' algoid_test.go` returns at least 2 (the new test + the `registered` map entry).
  </acceptance_criteria>
  <behavior>
    - Unit test suite covers identity (byte + rune), both-empty, one-empty, substring containment, no-overlap, symmetry, multi-byte rune handling, ASCII vs rune equivalence, custom params, raw unclamped negative case, AND the explicit GapSplitCanary unit test.
    - Property tests cover all six DET-04 + invariant categories (RangeBounds, Identity, Symmetric byte and rune, NoNaN, NoInf, NoNegativeZero) PLUS the three SWG-specific canaries (GapSplitInvariance, RawNeverExceedsMatchTimesMinLen, MonotonicWithMatchReward). NO triangle-inequality test (SWG is not a metric).
    - Benchmarks pin 0-alloc targets for ASCII Short, ASCII Medium, WithParams Short, RawScore Short; heap and rune paths run without panic.
    - Fuzz harness panic-free + invariant-preserving on programmatic seeds plus 30s of random input including invalid UTF-8 and the gap-split canary.
    - ExampleSmithWatermanGotohScore + ExampleSmithWatermanGotohRawScore visible on pkg.go.dev with the Output blocks verified byte-for-byte.
    - Dispatch assertion confirms SmithWatermanGotohScore is registered at slot 6.
  </behavior>
  <done>
    All test/bench/fuzz files committed. props_test.go + example_test.go + algoid_test.go appended. Test suite green for the SWG-scoped pattern. The runtime alloc gate enforces PERF-01 for SWG. Ready for Task 3 to wire the BDD harness and the staging golden file.
  </done>
</task>

<task type="auto" tdd="true">
  <name>Task 3: BDD feature + steps + staging golden file (testdata/golden/_staging/swg.json)</name>
  <files>tests/bdd/features/swg.feature, tests/bdd/steps/algorithms_steps.go, algorithms_golden_test.go, testdata/golden/_staging/swg.json</files>
  <read_first>
    - tests/bdd/features/levenshtein.feature (LOCKED Gherkin template — header comment with primary source, Feature header, Scenario Outline with Examples table, identity/both-empty/symmetry scenarios)
    - tests/bdd/features/damerau_full.feature (closest analog with a discriminating-vector scenario — pattern for the SWG gap-split canary scenario)
    - tests/bdd/steps/algorithms_steps.go (existing 413-line file — locate the Damerau-Levenshtein Full step block at the EOF and the InitializeScenario registration block around line ~396-413; the algorithm-agnostic `theScoreShouldBeApproximately`/`Exactly` and `theDistanceShouldBe` steps are registered ONCE in the Levenshtein/Hamming blocks per IN-06 — REUSE them, do NOT re-register)
    - algorithms_golden_test.go (LOCKED helper assertGoldenStaging from plan 02-01 Task 3; existing buildLevenshteinStagingEntries / TestGolden_Levenshtein_Staging template at lines ~196-210; existing buildJaroWinklerStagingEntries / TestGolden_JaroWinkler_Staging template at line ~574+; goldenAlgorithmsFile + goldenAlgorithmEntry struct definitions)
    - testdata/golden/_staging/levenshtein.json (canonical-byte-form reference: 2-space indent, trailing LF, sorted entries by Name)
    - golden_canonical.go (canonicalMarshal + WriteGoldenFile — invoked via assertGoldenStaging)
    - export_test.go (CanonicalMarshalForTest re-export)
    - .planning/phases/03-smith-waterman-gotoh/03-PATTERNS.md (`tests/bdd/features/swg.feature`, `tests/bdd/steps/algorithms_steps.go`, `testdata/golden/_staging/swg.json`, `algorithms_golden_test.go` sections — concrete templates)
    - .planning/phases/02-core-character-algorithms-six/02-VERIFICATION.md §IN-03 cleanup (BDD score regex relaxed to `(\d+\.?\d*)` — accepts integer-form like `0` / `1`; SWG scenarios may use the integer form)
  </read_first>
  <action>
**Create tests/bdd/features/swg.feature (Gherkin):**

1. Comment header (top of file, `#` comments) citing the three primary sources Smith-Waterman 1981 / Gotoh 1982 / Flouri 2015 + the score normalisation formula `clamp(best_local_score / min(len(a), len(b)), 0, 1)`.
2. `Feature: Smith-Waterman-Gotoh local-alignment similarity` with a one-line description naming the default params (Match=1.0, Mismatch=-1.0, GapOpen=-1.5, GapExtend=-0.5).
3. Scenarios:
   - **Scenario Outline: canonical reference vectors** with an Examples table containing at minimum: `(http_request, http_request_header_fields, 1.0000)`, `(abc, abc, 1.0000)`. Optionally add the kitten/sitting case if its expected approx-score is computable in advance and pre-validated against the kernel. Tolerance `within 0.0001`. Use the step `When I compute the SmithWatermanGotoh score between "<a>" and "<b>"` and `Then the score should be approximately <score> within 0.0001`.
   - **Scenario: identical strings score 1.0** — `When I compute the SmithWatermanGotoh score between "user_id" and "user_id"` / `Then the score should be exactly 1` (integer-form per IN-03).
   - **Scenario: both-empty strings score 1.0** — `When I compute the SmithWatermanGotoh score between "" and ""` / `Then the score should be exactly 1`.
   - **Scenario: one-empty string scores 0.0** — `When I compute the SmithWatermanGotoh score between "abc" and ""` / `Then the score should be exactly 0`.
   - **Scenario: score is symmetric** — `When I compute the SmithWatermanGotoh score between "kitten" and "sitting"` / `And I compute the second SmithWatermanGotoh score between "sitting" and "kitten"` / `Then both SmithWatermanGotoh scores should be equal`.
   - **Scenario: gap-split canary — splitting a long gap does not improve the score** (the load-bearing PITFALLS.md §3 #2 gate) — `When I compute the SmithWatermanGotoh score between "abc________def" and "abcdef"` / `And I compute the second SmithWatermanGotoh score between "abc____def____" and "abcdef"` / `Then both SmithWatermanGotoh scores should be equal`. (Alternative: structure as an inequality if equality is too strict for the chosen mismatch/gap params; the unit test in Task 2 already covers the stricter form.)

**Modify tests/bdd/steps/algorithms_steps.go — append SWG step block + InitializeScenario registrations:**

1. Append a new SWG step method block at file EOF (after the existing Damerau-Levenshtein Full block). Boundary comment style mirrors the Phase 2 blocks:

       // ---------------------------------------------------------------------------
       // Smith-Waterman-Gotoh step definitions (plan 03-01)
       // ---------------------------------------------------------------------------

       func (ctx *AlgorithmContext) iComputeTheSmithWatermanGotohScoreBetween(a, b string) error {
           ctx.lastScore = fuzzymatch.SmithWatermanGotohScore(a, b)
           return nil
       }

       func (ctx *AlgorithmContext) iComputeTheSecondSmithWatermanGotohScoreBetween(a, b string) error {
           ctx.lastScore2 = fuzzymatch.SmithWatermanGotohScore(a, b)
           return nil
       }

       func (ctx *AlgorithmContext) bothSmithWatermanGotohScoresShouldBeEqual() error {
           if ctx.lastScore != ctx.lastScore2 {
               return fmt.Errorf("scores not equal: %f != %f", ctx.lastScore, ctx.lastScore2)
           }
           return nil
       }

2. Append the matching ctx.Step registrations inside InitializeScenario BEFORE its closing brace, immediately after the Damerau-Levenshtein Full block:

       // Smith-Waterman-Gotoh step definitions (plan 03-01).
       ctx.Step(
           `^I compute the SmithWatermanGotoh score between "([^"]*)" and "([^"]*)"$`,
           a.iComputeTheSmithWatermanGotohScoreBetween,
       )
       ctx.Step(
           `^I compute the second SmithWatermanGotoh score between "([^"]*)" and "([^"]*)"$`,
           a.iComputeTheSecondSmithWatermanGotohScoreBetween,
       )
       ctx.Step(
           `^both SmithWatermanGotoh scores should be equal$`,
           a.bothSmithWatermanGotohScoresShouldBeEqual,
       )

3. REUSE the existing algorithm-agnostic steps: `theScoreShouldBeApproximately` and `theScoreShouldBeExactly` are registered once in the Levenshtein/Hamming blocks per IN-03 / IN-06 — do NOT re-register them. SWG has NO Distance variant (per CONTEXT.md §7 inherited / IN-06), so no new `theDistanceShouldBe` step is required.

**Modify algorithms_golden_test.go — append buildSWGStagingEntries + TestGolden_SmithWatermanGotoh_Staging at EOF:**

1. Append immediately AFTER the existing buildJaroWinklerStagingEntries + TestGolden_JaroWinkler_Staging block (file EOF, ~line 574+).

2. Add the helper function:

       // buildSWGStagingEntries returns the Smith-Waterman-Gotoh entries used by
       // TestGolden_SmithWatermanGotoh_Staging. ExpectedScore is computed from the
       // current implementation so the staging file stays in sync with actual output.
       func buildSWGStagingEntries(t *testing.T) []goldenAlgorithmEntry {
           t.Helper()
           return []goldenAlgorithmEntry{
               {Name: "SmithWatermanGotoh_both_empty",          Algorithm: "SmithWatermanGotoh", A: "",                B: "",                            ExpectedScore: fuzzymatch.SmithWatermanGotohScore("", "")},
               {Name: "SmithWatermanGotoh_identical",           Algorithm: "SmithWatermanGotoh", A: "abc",             B: "abc",                         ExpectedScore: fuzzymatch.SmithWatermanGotohScore("abc", "abc")},
               {Name: "SmithWatermanGotoh_one_empty",           Algorithm: "SmithWatermanGotoh", A: "abc",             B: "",                            ExpectedScore: fuzzymatch.SmithWatermanGotohScore("abc", "")},
               {Name: "SmithWatermanGotoh_two_substring",       Algorithm: "SmithWatermanGotoh", A: "http_request",    B: "http_request_header_fields",  ExpectedScore: fuzzymatch.SmithWatermanGotohScore("http_request", "http_request_header_fields")},
               {Name: "SmithWatermanGotoh_no_overlap",          Algorithm: "SmithWatermanGotoh", A: "qqqq",            B: "zzzz",                        ExpectedScore: fuzzymatch.SmithWatermanGotohScore("qqqq", "zzzz")},
               {Name: "SmithWatermanGotoh_one_long_gap_canary", Algorithm: "SmithWatermanGotoh", A: "abc________def",  B: "abcdef",                      ExpectedScore: fuzzymatch.SmithWatermanGotohScore("abc________def", "abcdef")},
           }
       }

3. Add the staging test:

       func TestGolden_SmithWatermanGotoh_Staging(t *testing.T) {
           entries := buildSWGStagingEntries(t)
           sort.Slice(entries, func(i, j int) bool { return entries[i].Name < entries[j].Name })
           file := goldenAlgorithmsFile{Version: 1, Entries: entries}
           assertGoldenStaging(t, "_staging/swg.json", file)
       }

4. DO NOT modify the `stagingFiles` slice inside TestGolden_Algorithms_Merge — plan 03-03 owns the merge-list update (and the matching golden-file regeneration). This plan only writes the staging file.

**Generate the initial _staging/swg.json:**

Run `go test -run TestGolden_SmithWatermanGotoh_Staging -update -count=1 ./...` once. Inspect testdata/golden/_staging/swg.json: 6 entries sorted alphabetically by Name, 2-space indent, trailing LF byte, no BOM. Commit the file. Re-run without `-update` and confirm zero diff.

**Run the full BDD + golden gate:**

       go test -race -shuffle=on -count=1 -run 'TestGolden_SmithWatermanGotoh_Staging' ./...
       (cd tests/bdd && go test -race -shuffle=on -count=1 ./...)
       make check

`make check` must exit 0. `tests/bdd` tests pass (godog suite green; goleak detects no leaks). The staging file is byte-stable across re-runs (re-running TestGolden_SmithWatermanGotoh_Staging WITHOUT `-update` produces zero diff).
  </action>
  <verify>
    <automated>go test -race -shuffle=on -count=1 -run 'TestGolden_SmithWatermanGotoh_Staging' ./... && (cd tests/bdd && go test -race -shuffle=on -count=1 ./...) && make check</automated>
  </verify>
  <acceptance_criteria>
    - tests/bdd/features/swg.feature exists and parses as valid Gherkin; contains at minimum 6 scenario elements (Scenario Outline + identity + both-empty + one-empty + symmetry + gap-split-canary). Verifiable by `grep -cE '^[[:space:]]*(Scenario|Scenario Outline):' tests/bdd/features/swg.feature` returning ≥ 6.
    - tests/bdd/features/swg.feature top comment cites Smith-Waterman 1981, Gotoh 1982, and Flouri 2015 (`grep -cE 'Smith.*Waterman.*1981|Gotoh.*1982|Flouri.*2015' tests/bdd/features/swg.feature` returns ≥ 3).
    - tests/bdd/steps/algorithms_steps.go contains the three new SWG step methods iComputeTheSmithWatermanGotohScoreBetween / iComputeTheSecondSmithWatermanGotohScoreBetween / bothSmithWatermanGotohScoresShouldBeEqual. Verifiable by `grep -cE 'func \(ctx \*AlgorithmContext\) (iComputeTheSmithWatermanGotoh|bothSmithWatermanGotoh)' tests/bdd/steps/algorithms_steps.go` returning 3.
    - tests/bdd/steps/algorithms_steps.go InitializeScenario registers the three SWG step regexes. Verifiable by `grep -c 'I compute the SmithWatermanGotoh score between' tests/bdd/steps/algorithms_steps.go` returning ≥ 2 (the function-level registration + the regex string itself appear).
    - algorithms_golden_test.go contains buildSWGStagingEntries and TestGolden_SmithWatermanGotoh_Staging (`grep -c 'buildSWGStagingEntries\|TestGolden_SmithWatermanGotoh_Staging' algorithms_golden_test.go` returns ≥ 2).
    - testdata/golden/_staging/swg.json exists, contains exactly 6 entries sorted alphabetically by Name, in canonical byte form (2-space indent, trailing LF, no BOM). Verifiable by `jq '.entries | length' testdata/golden/_staging/swg.json` returning 6.
    - testdata/golden/_staging/swg.json contains literal `SmithWatermanGotoh_one_long_gap_canary` (`grep -c 'SmithWatermanGotoh_one_long_gap_canary' testdata/golden/_staging/swg.json` returns ≥ 1).
    - Re-running `go test -run TestGolden_SmithWatermanGotoh_Staging -count=1 ./...` WITHOUT `-update` exits 0 (file byte-stable across re-runs).
    - `(cd tests/bdd && go test -race -shuffle=on -count=1 ./...)` exits 0 (godog suite green; goleak detects no leaks).
    - `make check` exits 0 (full quality gate: lint, vet, race, coverage-check, verify-license-headers, verify-no-runtime-deps, tidy-check, BDD, plus the inherited determinism-gate which now includes the SWG staging file).
    - Coverage on swg.go ≥ 90% (`go test -cover ./...` or `make coverage`); coverage on the SWG public surface 100%.
  </acceptance_criteria>
  <behavior>
    - BDD harness exercises canonical SWG reference vectors via the Scenario Outline AND pins identity, both-empty, one-empty, symmetry, and the gap-split canary as separate scenarios.
    - The algorithms_steps.go append pattern keeps existing Phase 2 blocks untouched; SWG steps coexist with the algorithm-agnostic theScoreShouldBeApproximately/Exactly steps already registered in the Levenshtein/Hamming blocks.
    - testdata/golden/_staging/swg.json is the Phase-3 contribution to the Phase-2-locked staging pattern; plan 03-03's merge step reads it alongside the six Phase 2 staging files to produce the canonical algorithms.json.
    - Phase 3's algorithm-correctness + safety quality bar is met: lint, vet, race, coverage, BDD, license headers, no-runtime-deps, tidy-check, determinism golden file (Phase 2 entries) all green.
  </behavior>
  <done>
    swg.feature + SWG step block + buildSWGStagingEntries + TestGolden_SmithWatermanGotoh_Staging + _staging/swg.json committed. BDD suite green. _staging/swg.json byte-stable. make check green. Plan 03-02 can now attach the cross-validation test against this stable surface; plan 03-03 can merge _staging/swg.json into algorithms.json.
  </done>
</task>

</tasks>

<threat_model>
## Trust Boundaries

| Boundary | Description |
|----------|-------------|
| Caller → fuzzymatch.SmithWatermanGotoh* (six functions) | Untrusted strings (any length, any byte sequence including invalid UTF-8) plus untrusted SWGParams values (any float64 including NaN, +Inf, -Inf, +0, -0) crossing into the algorithm. No other boundary — pure-function library; no I/O, no network, no parsing of user-controlled config. |
| Caller → SWGParams construction | Untrusted parameter values (Match, Mismatch, GapOpen, GapExtend). The library does not validate; per CONTEXT.md §3 the function produces a deterministic-but-meaningless score on nonsense params. |
| go test → testdata/fuzz/FuzzSmithWatermanGotohScore/ | Fuzz corpus loaded from disk; managed by the test harness. |

## STRIDE Threat Register

| Threat ID | Category | Component | Disposition | Mitigation Plan |
|-----------|----------|-----------|-------------|-----------------|
| T-3-01 | Denial of Service | swgDPRaw kernel (O(m·n) on adversarial inputs) | mitigate | Document worst-case O(m·n) in swg.go godoc; caller responsible for input-length caps. Heap path uses 6 × `make([]float64, n+1)` = O(n) memory, bounded linearly NOT quadratically. PERF-01 + benchstat ensure no super-linear slowdown sneaks in (plan 03-03 commits the bench.txt baseline). FuzzSmithWatermanGotohScore harness (Task 2) covers panic-freedom on arbitrary inputs over 30s+ random input. No exponential paths exist — the three-matrix two-row recurrence is strictly O(m·n) time, O(min(m,n)) space. HIGH severity but mitigation is mandatory and complete. |
| T-3-02 | Information Disclosure | Malformed UTF-8 input causing panic / leaking internal state | mitigate | Byte-level kernel operates on bytes directly (no UTF-8 decoding) — invalid UTF-8 cannot panic the byte path. Rune variants use `[]rune(s)` which Go's stdlib normalises to U+FFFD on malformed bytes (no panic). FuzzSmithWatermanGotohScore harness includes invalid-UTF-8 seeds (`"\xff\xfe"`) and asserts no panic, no NaN, no Inf, no out-of-[0,1] return. HIGH severity but mitigation is mandatory and complete. |
| T-3-03 | Tampering | Float-determinism leak — platform-specific intrinsics (FMA on arm64 emitted automatically by the compiler for `x*y+z` patterns) produce divergent output across linux/amd64 vs linux/arm64 vs darwin/arm64 vs windows/amd64 | mitigate | Determinism-standards rules: NO math.Pow / math.Log / math.Exp / math.FMA anywhere in swg.go (grep gate in Task 1 acceptance criteria). Only +, -, *, /, comparisons, float64() casts. Recurrence equations expressed as explicit sequential statements (no `a*b + c*d` fused-multiply-add-eligible patterns — the kernel uses `var sij float64; if ai == b[j-1] { sij = params.Match } else { sij = params.Mismatch }` then `m1 := prevM[j-1] + sij` — addition only, NOT multiply-add). Sum reductions left-to-right; no parallel sums. Cross-platform CI matrix in plan 03-03 (CI gate on the merged algorithms.json) catches any divergence. HIGH severity but mitigation is mandatory and complete. |
| T-3-04 | Tampering | dispatch[AlgoSmithWatermanGotoh] overwritten by a later plan or external code | mitigate | dispatch is unexported (package fuzzymatch only). The `var _ = func() bool { ... }()` registration runs once at package load. There is no public mutator. Any future code attempting to overwrite the slot must edit the package source — caught by code review. Phase 2 T-02-01-04 mitigation pattern still applies. |
| T-3-05 | Cache / memo poisoning | N/A — algorithm has no caching, pure function | accept | No mitigation needed. SWG entry points compute fresh on every call; no package-level state mutated. |
| T-3-06 | Spoofing | Staging file (testdata/golden/_staging/swg.json) tampered with to weaken determinism gate | mitigate | The staging file is committed to git; PR review + the CI matrix diff (`make verify-determinism` runs on all 5 platforms after plan 03-03's merge) detect any divergence. CanonicalMarshalForTest re-export ensures the byte form is locked. Plan 03-02's biopython cross-validation (separate corpus) provides defense-in-depth. |

Severity assessment: T-3-01 / T-3-02 / T-3-03 are HIGH; all three have mandatory and complete mitigations in this plan (the runtime alloc gate, the fuzz harness, the float-determinism discipline rules with grep gates). T-3-04 / T-3-06 are LOW (mitigated structurally). T-3-05 N/A. No HIGH severity item accepted without mitigation. ASVS L1 V5.1 (Input Validation) and V8.2 (Resilience to malformed input) addressed via fuzz tests + property tests. Plan passes the security gate.
</threat_model>

<verification>
1. Build: `go build ./...` exits 0.
2. Vet: `go vet ./...` exits 0.
3. License headers: `bash scripts/verify-license-headers.sh` exits 0 (5 new .go files in root all carry the Apache-2.0 header).
4. No-runtime-deps: `bash scripts/verify-no-runtime-deps.sh` exits 0 (no new require entries in root go.mod).
5. Unit + property tests: `go test -race -shuffle=on -count=1 -run 'TestSmithWatermanGotoh|TestProp_SmithWatermanGotoh|TestDispatch_SmithWatermanGotoh|ExampleSmithWatermanGotoh' ./...` exits 0.
6. Examples: `go test -run 'ExampleSmithWatermanGotoh' ./...` exits 0 (both Output blocks match byte-for-byte).
7. Allocation budget (test-time gate): `go test -run TestSmithWatermanGotohScore_ZeroAllocs ./...` exits 0.
8. Allocation budget (bench-time): `go test -bench=BenchmarkSmithWatermanGotohScore_ASCII_Short -benchmem -run=^$ -count=3 ./...` reports `0 B/op  0 allocs/op` for the Short benchmark (likewise Medium, WithParams_Short, RawScore_Short).
9. Fuzz smoke: `go test -fuzz=FuzzSmithWatermanGotohScore -fuzztime=30s -run=^$ ./...` completes without crash or invariant violation.
10. Staging golden: `go test -run TestGolden_SmithWatermanGotoh_Staging -count=1 ./...` exits 0 WITHOUT `-update` (file byte-stable).
11. BDD: `(cd tests/bdd && go test -race -shuffle=on -count=1 ./...)` exits 0.
12. No-redeclare gate: `grep -E '^const maxStackInputLen\b|^func isASCII\b' swg.go` returns no matches.
13. No-init gate: `grep -vE '^[[:space:]]*//' swg.go | grep -cE '^func init\('` returns 0; likewise for dispatch_swg.go.
14. No-FMA gate: `grep -E 'math\.(Pow|Log|Exp|FMA)' swg.go` returns no matches.
15. Coverage: `make coverage` — overall ≥ 95%, swg.go ≥ 90%, 100% on all six new public symbols + SWGParams + NewSWGParams.
16. Full quality gate: `make check` exits 0.
</verification>

<success_criteria>
- A caller can `import "github.com/axonops/fuzzymatch"` and obtain deterministic SWG similarity AND raw alignment scores via the six public functions and the SWGParams + NewSWGParams pair.
- The byte-level fast path on ASCII shorter dimension ≤ 64 chars allocates zero bytes (PERF-01, PERF-02, PERF-06 satisfied for SWG via Task 2's runtime alloc gate AND the four 0-alloc benchmarks).
- The implementation uses three-matrix two-row DP (six rolling rows), no full DP table (PERF-03 satisfied for SWG).
- Score is in [0.0, 1.0] for any input; never NaN, Inf, or -0 (DET-04 satisfied for SWG via the seven property tests + the fuzz harness).
- The Gotoh-erratum gate is closed: file-level godoc cites all three primary references (Smith-Waterman 1981, Gotoh 1982, Flouri 2015) AND names the erratum inline; the GapSplitInvariance property test + the unit-test GapSplitCanary case + the BDD gap-split scenario all pass (PITFALLS.md §3 warning signs #1-4 cleared).
- testdata/golden/_staging/swg.json is byte-stable and ready for plan 03-03's merge step (plan 03-03 reads it alongside the six Phase 2 staging files).
- TestSWG_CrossValidation (plan 03-02) attaches to a working public surface (Score, ScoreWithParams, SWGParams all available).
- The canonical pattern from Phase 2 (file naming, dispatch idiom, test/bench/fuzz/golden/BDD/example layout) is replicated faithfully; Phase 3's only innovations beyond the locked Phase 2 pattern are (a) the SWGParams + NewSWGParams shape (the first parameterised algorithm) and (b) the Raw* surface expansion (per CONTEXT.md §4).
- All required gates (license headers, no-runtime-deps, lint, vet, race, tidy, coverage, determinism for Phase 2 entries, BDD) pass via `make check`.
</success_criteria>

<output>
After completion, create `.planning/phases/03-smith-waterman-gotoh/03-01-swg-implementation-SUMMARY.md` per the standard summary template, recording:
- Final identifier names (SmithWatermanGotohScore, *Runes, *WithParams, *RawScore, *RawScoreRunes, *RawScoreWithParams, SWGParams, NewSWGParams — confirm zero drift from this plan; if api-ergonomics-reviewer requests a rename during code review, record both the as-shipped name and the as-planned name).
- Benchmark numbers observed locally (B/op, allocs/op for ASCII Short / Medium / Long, Unicode Short, WithParams Short, RawScore Short).
- Coverage percentages (overall, per-file swg.go, public-symbol).
- The exact stack-buffer size shipped (3120 bytes for `[(maxStackInputLen+1)*6]float64`; confirm no drift from CONTEXT.md §2).
- The SWG_two_substring entry's exact ExpectedScore in _staging/swg.json (should be 1.0; if the kernel returns something else the implementation is wrong).
- The Gotoh-erratum gate evidence: a verbatim quote of the file-level godoc block from swg.go showing the three primary-reference citations and the erratum statement.
- Any deviations from the plan and their rationale.
- The hand-off contract to plan 03-02: which public surface (SmithWatermanGotohScoreWithParams + SWGParams) the cross-validation test consumes; which staging file (_staging/swg.json) plan 03-03 reads.
</output>
</content>
