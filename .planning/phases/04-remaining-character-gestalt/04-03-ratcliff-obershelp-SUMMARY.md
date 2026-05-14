---
phase: 04-remaining-character-gestalt
plan: 03
subsystem: similarity-algorithms
tags: [ratcliff-obershelp, dr-dobbs-1988, difflib-equivalent, autojunk-false, recursive-lcsubstring, asymmetric-by-design, gestalt, oq-1-resolution, dispatch-registration, property-tests, fuzz, benchmark, bdd, staging-golden, no-init]

# Dependency graph
requires:
  - phase: 02-core-character-algorithms-six
    provides: maxStackInputLen=64 + isASCII gate (referenced by name, not used by RO — RO has no stack-fast-path because the recursion depth is data-dependent); assertGoldenStaging helper + goldenAlgorithmEntry / goldenAlgorithmsFile schema for the per-algorithm staging-golden pattern; AlgoRatcliffObershelp slot 22 (numAlgorithms-1) already declared in algoid.go
  - phase: 03-smith-waterman-gotoh
    provides: per-algorithm BDD feature + step-bindings append pattern; props_test.go append-block pattern; ExampleXxxScore append-only example_test.go discipline; testdata/fuzz/Fuzz<Algo>Score/seed-001 byte-stable format (IN-06 closure); identity-short-circuit on *Runes BEFORE []rune allocation (IN-04 closure); BDD score regex (\d+\.?\d*) integer-form acceptance (IN-03 closure); fuzz harness exercises full public surface (WR-02 closure); numerical-regression pin alongside cross-validation corpus (WR-03 closure)
  - phase: 04-remaining-character-gestalt
    provides: 04-01 strcmp95 pattern for the 3-task algorithm shape (impl+test+staging / property+bench+fuzz / BDD); 04-02 lcsstr provides the LCS-substring DP recurrence shape (two-row rolling buffer with strict-`>` max-update for leftmost tie-break) — RO inlines a substring-position-returning variant rather than reusing lcsstrDP per CONTEXT.md D-3
provides:
  - RatcliffObershelpScore(a, b string) float64 — difflib-equivalent gestalt-pattern-matching similarity (byte path; dispatched)
  - RatcliffObershelpScoreRunes(a, b string) float64 — rune-path variant for multi-byte UTF-8
  - dispatch[AlgoRatcliffObershelp] slot 22 (the LAST slot, numAlgorithms-1) populated via var-init (no init())
  - testdata/golden/_staging/ratcliff_obershelp.json — 7 entries including the asymmetric tide/diet pair (one direction)
  - testdata/fuzz/FuzzRatcliffObershelpScore/seed-001 — canonical Dr. Dobb's 1988 WIKIMEDIA/WIKIMANIA pair
  - TestRatcliffObershelp_AsymmetryPin — load-bearing OQ-1-resolution regression guard
  - TestProp_RatcliffObershelpScore_AtLeastLevenshtein_HandCurated — algorithm-specific hand-curated property
  - roMatchedLength + roFindLongestMatch (byte) and Runes variants — recursive longest-common-substring decomposition kernels
affects: [04-04-ratcliff-obershelp-cross-validation, 04-05-finalisation]

# Tech tracking
tech-stack:
  added: []
  patterns:
    - "Asymmetric-by-design algorithm in a symmetric-by-default catalogue. The standard Phase 2 Symmetric property test is DROPPED for Ratcliff-Obershelp (only TWO standard invariants are dropped, not the entire property-test block). The five remaining standard invariants (RangeBounds, Identity, NoNaN, NoInf, NoNegativeZero) still apply on both byte and rune paths. The asymmetry contract is pinned by TestRatcliffObershelp_AsymmetryPin (tide/diet=0.25, diet/tide=0.5) with an inline reference to the OQ-1 resolution locked 2026-05-14."
    - "Inlined substring-position-returning DP variant (roFindLongestMatch) rather than reusing lcsstr.go's lcsstrDP. The CONTEXT.md D-3 decision allowed either path; the inline variant was chosen because Ratcliff-Obershelp needs aLo + bLo + bHi (lcsstrDP returns only the length + end-index-in-`a`). The DP recurrence itself is identical to lcsstrDP — same two-row rolling buffer with strict-`>` max-update for leftmost-in-`a`-then-leftmost-in-`b` tie-break that mirrors Python difflib.SequenceMatcher.find_longest_match."
    - "Language-native call stack for the recursion (per CONTEXT.md D-2). Recursion depth is bounded by O(min(len(a), len(b))); no stack-overflow risk for reasonable inputs. The matchedLength function recurses into a[:aLo]/b[:bLo] and a[aHi:]/b[bHi:] after each find_longest_match — byte-stable on the CI matrix because Go strings are immutable byte slices and the recursion order is deterministic left-then-right."
    - "BDD step for long inputs via Go constants (var roAutojunkA / roAutojunkB computed by strings.Repeat). The 205-char autojunk-sensitive scenario uses a dedicated step (iComputeTheRatcliffObershelpScoreForTheAutojunkSensitivePair) rather than embedding the 205-char literals in the Gherkin Examples table. The Go constants are computed via strings.Repeat at package-init time — no hand-counting required — and the BDD step pins the expected difflib(autojunk=False).ratio() value 0.7317."

key-files:
  created:
    - ratcliff_obershelp.go
    - dispatch_ratcliff_obershelp.go
    - ratcliff_obershelp_test.go
    - ratcliff_obershelp_bench_test.go
    - ratcliff_obershelp_fuzz_test.go
    - testdata/golden/_staging/ratcliff_obershelp.json
    - testdata/fuzz/FuzzRatcliffObershelpScore/seed-001
    - tests/bdd/features/ratcliff_obershelp.feature
  modified:
    - algorithms_golden_test.go (append: buildRatcliffObershelpStagingEntries + TestGolden_RatcliffObershelp_Staging)
    - algoid_test.go (append: TestDispatch_RatcliffObershelpRegistered; slot 22 in TestDispatch_UnregisteredSlotsAreNil registered map — now all 10 currently-registered slots covered)
    - example_test.go (append: ExampleRatcliffObershelpScore, ExampleRatcliffObershelpScoreRunes)
    - props_test.go (append: 10 standard property tests across byte+rune surfaces — RangeBounds/Identity/NoNaN/NoInf/NoNegativeZero × 2 — PLUS TestProp_RatcliffObershelpScore_AtLeastLevenshtein_HandCurated; Symmetric INTENTIONALLY OMITTED per OQ-1)
    - tests/bdd/steps/algorithms_steps.go (append: 2 step methods + 2 ctx.Step registrations; added strings import for the autojunk Go constants)
    - llms.txt (append: Ratcliff-Obershelp section with the 2 exported symbols — meta-test TestAIFriendly_LLMSTxtReferencesEveryExportedSymbol closure)

key-decisions:
  - "Drop the Symmetric property test from the standard Phase 2 invariant suite. Ratcliff-Obershelp is intentionally asymmetric (mirrors Python difflib.SequenceMatcher.ratio() per CPython bpo-37004) — a Symmetric property test would either fail on counterexamples like tide/diet OR force a hidden input-sorting workaround that breaks byte-for-byte difflib equivalence. OQ-1 resolution LOCKED 2026-05-14 (CONTEXT.md §4). The remaining five standard invariants (RangeBounds, Identity, NoNaN, NoInf, NoNegativeZero) still apply on both byte and rune surfaces."
  - "Inline a substring-position-returning DP variant (roFindLongestMatch returning aLo/aHi/bLo/bHi/n) rather than reusing lcsstr.go's lcsstrDP. CONTEXT.md D-3 allowed either path; inline was chosen because the abstraction does not earn its keep — RO needs aLo + bLo + bHi to drive the recursion into a[:aLo]/b[:bLo] and a[aHi:]/b[bHi:], which lcsstrDP's (length, endI) return shape cannot provide without rework."
  - "Use the language-native call stack for the recursion (CONTEXT.md D-2). For reasonable inputs (< 64KB strings) the call-stack depth is bounded by O(min(la, lb)) and well below Go's default goroutine stack. An explicit iterative stack would be more complex without buying anything for the documented input-size range."
  - "Drop the AsymmetryPin into both ratcliff_obershelp_test.go AND the cross-algorithm consistency test in plan 04-05. The local pin (TestRatcliffObershelp_AsymmetryPin) catches regressions during single-file edits; the cross-algorithm test catches regressions during catalog-wide refactors. The two layers are complementary — the local pin is fast, the cross-algorithm test is comprehensive."
  - "BDD autojunk-sensitive scenario uses Go-side constants rather than 205-char Gherkin literals. The constants live in tests/bdd/steps/algorithms_steps.go as `var roAutojunkA = strings.Repeat(\"a\", 100) + strings.Repeat(\"x\", 5) + strings.Repeat(\"a\", 100)` etc., so the character counts are arithmetic — no hand-counting required. (An earlier draft used hand-counted const string literals; that was Auto-fixed under Rule 1 — see Deviations.)"

patterns-established:
  - "Pattern: per-algorithm OQ resolution embedded in code comments. RO's godoc on RatcliffObershelpScore and the props_test.go section header BOTH cite \"OQ-1 RESOLUTION LOCKED 2026-05-14\" inline. This is more durable than a planning-doc-only reference because future maintainers grep the code, not the .planning/ directory."
  - "Pattern: numerical-regression pin OUTSIDE the cross-validation corpus (Phase 3 WR-03 closure). TestRatcliffObershelp_PinnedDrDobbsValue pins WIKIMEDIA/WIKIMANIA = 0.7777777777777778 directly in ratcliff_obershelp_test.go, not just via the corpus. Plan 04-04 will add a separate TestRatcliffObershelp_CrossValidation that reads the corpus; the local pin catches regressions even if the corpus is accidentally accepted unchanged."
  - "Pattern: per-plan llms.txt sync. Inherited from plan 04-01's discipline. Each Phase 4 plan appends its exported symbols to llms.txt within its own commits, not deferred to finalisation. The TestAIFriendly meta-test green-bar would block otherwise. RO appends two symbols (RatcliffObershelpScore + Runes)."

requirements-completed: [GESTALT-01]

# Metrics
duration: ~30min
completed: 2026-05-14
---

# Phase 4 Plan 03: Ratcliff-Obershelp Summary

**Ratcliff & Metzener 1988 (Dr. Dobb's Journal 13(7):46-51) gestalt-pattern-matching similarity — the load-bearing Python-difflib-equivalent for the fuzzymatch catalogue. Two public functions (RatcliffObershelpScore + Runes), recursive longest-common-substring decomposition byte-for-byte equivalent to `difflib.SequenceMatcher(autojunk=False).ratio()` within 1e-9, dispatch slot 22 wired (the LAST slot — numAlgorithms-1), and asymmetric-by-design per OQ-1 resolution LOCKED 2026-05-14.**

## Performance

- **Duration:** ~30 min (3 atomic commits)
- **Started:** 2026-05-14T14:55Z (approx.)
- **Completed:** 2026-05-14T15:19Z
- **Tasks:** 3 (all completed)
- **Files modified:** 14 (8 created, 6 appended)

## Accomplishments

- Two public functions exposed: `RatcliffObershelpScore`, `RatcliffObershelpScoreRunes` — spec-pinned at docs/requirements.md §7.1.24.
- Dispatch slot 22 (`AlgoRatcliffObershelp`, the LAST slot — numAlgorithms-1) populated via `var _ = func() bool{...}()` (no init()). `TestDispatch_UnregisteredSlotsAreNil` now expects ALL ten currently-implemented slots registered; the remaining 13 slots (9..21) await later phases for q-gram / token / phonetic algorithms.
- Byte-for-byte equivalent to Python `difflib.SequenceMatcher(autojunk=False).ratio()` within 1e-9 on the Dr. Dobb's 1988 canonical pairs:
  - `RatcliffObershelpScore("WIKIMEDIA", "WIKIMANIA")` = 0.7777777777777778
  - `RatcliffObershelpScore("GESTALT", "GESTALT_PATTERN_MATCHING")` = 0.45161290322580644
- INTENTIONALLY asymmetric in argument order per OQ-1 resolution LOCKED 2026-05-14. `TestRatcliffObershelp_AsymmetryPin` pins both `RatcliffObershelpScore("tide", "diet") = 0.25` AND `RatcliffObershelpScore("diet", "tide") = 0.5` and asserts `fwd != rev`. The asymmetry is documented in godoc (with a reference to CPython bpo-37004) and in the props_test.go section header that explicitly omits the Symmetric property test.
- 100% coverage on both public functions; 96%–100% on the four unexported helpers (`roMatchedLength`, `roFindLongestMatch`, `roMatchedLengthRunes`, `roFindLongestMatchRunes`).
- Identity short-circuit on `*Runes` returns 1.0 BEFORE `[]rune` allocation (IN-04 closure); `TestRatcliffObershelp_RuneIdentity_ShortCircuit` pins 0 allocs/op via `testing.AllocsPerRun`.
- Seven-entry staging golden at `testdata/golden/_staging/ratcliff_obershelp.json` covering: both-empty, GESTALT (Dr. Dobb's paper-cited), identical, one-empty, substring middle (abcdef/xyzabcdefuvw), the asymmetric tide/diet pair (one direction), and WIKIMEDIA/WIKIMANIA. Alphabetical by Name. Merge into algorithms.json owned by plan 04-05.
- FuzzRatcliffObershelpScore exercises BOTH public surfaces (Phase 3 WR-02 closure); 10-second smoke run = 1.6M execs, 0 panics, 0 NaN/Inf, 0 out-of-range. Seed corpus includes the autojunk-sensitive 205-char pair (RESEARCH.md Pitfall 2 closure), Dr. Dobb's pairs, multi-byte UTF-8, Cyrillic, invalid UTF-8, and both directions of the asymmetric tide/diet pair.
- Five Gherkin scenarios in `tests/bdd/features/ratcliff_obershelp.feature`: canonical Dr. Dobb's reference vectors (Scenario Outline with 2 examples), identical, both-empty, one-empty, and the 205-char autojunk-sensitive scenario. NO symmetry scenario per OQ-1.

## Task Commits

Each task was committed atomically:

1. **Task 1: Implement ratcliff_obershelp.go (algorithm + recursive decomposition + dispatch + unit tests + staging golden + example + llms.txt sync)** — `1c0c5be` (feat)
2. **Task 2: Ratcliff-Obershelp property tests + benchmarks + fuzz** — `a019493` (test)
3. **Task 3: Ratcliff-Obershelp BDD feature + steps** — `bc7fb2a` (test)

Note: TDD was applied as a compact red→green per task. The implementation, dispatch wiring, unit tests, golden, llms.txt, example, and algoid_test.go updates form an indivisible atomic unit in Task 1 — the unit tests cannot pass without the implementation, and the example_test.go / TestAIFriendly meta-test would fail without llms.txt being in sync. Compaction into a single feat commit per task matches the inherited Phase 4 pattern from plans 04-01 and 04-02.

## Files Created/Modified

### Created

- `ratcliff_obershelp.go` — Ratcliff & Metzener 1988 recursive longest-common-substring decomposition. Two public functions (RatcliffObershelpScore + Runes) atop unexported helpers (roMatchedLength + Runes for the recursion, roFindLongestMatch + Runes for the per-level LCS-substring DP). The longest-common-substring kernel is inlined (not reused from lcsstr.go) because RO needs aLo + bLo + bHi from the DP — lcsstrDP returns only (length, endI). Language-native call stack for the recursion. No init(), no map iteration, no transcendentals. File-level godoc cites Ratcliff & Metzener 1988 as PRIMARY source and Python difflib as cross-validation only; includes the explicit source-origin statement block (Primary / Cross-validation / GPL-LGPL: none / Code copied: none); documents the autojunk=False qualifier (RESEARCH.md Pitfall 2) and the asymmetric-by-design semantics (OQ-1 resolution).
- `dispatch_ratcliff_obershelp.go` — registers `RatcliffObershelpScore` into `dispatch[AlgoRatcliffObershelp]` (slot 22 — the LAST slot, numAlgorithms-1) via the canonical var-init idiom (no init()). Only the byte-path score is dispatched; the rune-path variant is public but not dispatched.
- `ratcliff_obershelp_test.go` — eight unit tests covering both-empty, one-empty, identical, Dr. Dobb's 1988 reference vectors (WIKIMEDIA/WIKIMANIA, GESTALT/GESTALT_PATTERN_MATCHING — within 1e-9 of difflib(autojunk=False).ratio()), the pinned-value test for WR-03 closure (WIKIMEDIA/WIKIMANIA = 0.7777777777777778 OUTSIDE the cross-validation corpus), the asymmetry pin (tide/diet=0.25 vs diet/tide=0.5; fwd != rev), byte-vs-rune equivalence on ASCII, rune multi-byte handling (café/cafe rune path = 0.75; byte path = 6/9; verified they differ), and the rune-identity zero-alloc gate.
- `ratcliff_obershelp_bench_test.go` — five benchmarks: `BenchmarkRatcliffObershelpScore_{ASCII_Short, ASCII_Medium, ASCII_Long, Unicode_Short}` + `BenchmarkRatcliffObershelpScoreRunes_Unicode_Short`. `b.ReportAllocs()` on every benchmark; `var sink float64` + post-loop guard prevents DCE. Allocation reports informational — RO has no stack fast path because recursion depth is data-dependent.
- `ratcliff_obershelp_fuzz_test.go` — `FuzzRatcliffObershelpScore` exercising BOTH public surfaces (Phase 3 WR-02 closure). Body asserts no NaN, no Inf, and score in [0,1] on each surface; no-panic is implicit. Seed corpus includes standard edges (identity, both-empty, one-empty, no-overlap), Dr. Dobb's 1988 paper vectors, an autojunk-sensitive 205-char pair (RESEARCH.md Pitfall 2 closure — proves no autojunk-like heuristic), substring containment, multi-byte UTF-8 (café/cafe, Привет/привет), invalid UTF-8 (\xff\xfe, \xc0\x80), and the asymmetric tide/diet pair in BOTH directions.
- `testdata/fuzz/FuzzRatcliffObershelpScore/seed-001` — canonical Dr. Dobb's 1988 WIKIMEDIA/WIKIMANIA pair in `go test fuzz v1` literal format (byte-stable per IN-06 closure).
- `testdata/golden/_staging/ratcliff_obershelp.json` — seven entries (sorted alphabetically by Name) for plan 04-05's algorithms.json merge: RatcliffObershelp_both_empty (1.0), RatcliffObershelp_gestalt_paper (0.4516), RatcliffObershelp_identical (1.0), RatcliffObershelp_one_empty (0.0), RatcliffObershelp_substring_middle (0.6667), RatcliffObershelp_tide_diet_asymmetric (0.25 — one direction; the asymmetry-pin test verifies fwd != rev), RatcliffObershelp_wikimedia_wikimania (0.7778).
- `tests/bdd/features/ratcliff_obershelp.feature` — five scenarios with Ratcliff & Metzener 1988 attribution in the header comment: canonical Dr. Dobb's Scenario Outline (2 rows), identical, both-empty, one-empty, and the 205-char autojunk-sensitive scenario (uses a dedicated step backed by Go constants — 205-char literals in Gherkin Examples tables were rejected for readability). The header comment explicitly notes the OQ-1 resolution: NO symmetry scenario.

### Modified

- `algoid_test.go` — appended `TestDispatch_RatcliffObershelpRegistered` asserting `dispatch[AlgoRatcliffObershelp]` non-nil; extended the `registered` map in `TestDispatch_UnregisteredSlotsAreNil` to flip slot 22 (the LAST slot) to true. The test's godoc comment now reflects that slot 22 is registered by plan 04-03 — and slots 9..21 remain nil pending q-gram / token / phonetic algorithms in later phases.
- `algorithms_golden_test.go` — appended `buildRatcliffObershelpStagingEntries` and `TestGolden_RatcliffObershelp_Staging`. ExpectedScore is computed from the current implementation so the staging file stays in sync; entries sorted alphabetically by Name via `sort.Slice`. The merge into `testdata/golden/algorithms.json` is owned by plan 04-05.
- `example_test.go` — appended two runnable godoc Examples: `ExampleRatcliffObershelpScore` (WIKIMEDIA/WIKIMANIA = 0.7778) and `ExampleRatcliffObershelpScoreRunes` (café/cafe = 0.7500). Output blocks pin the formatted values; `go test -run ExampleRatcliffObershelp ./...` verifies byte-for-byte.
- `props_test.go` — appended 11 Ratcliff-Obershelp property tests: 10 standard Phase 2 invariants (RangeBounds, Identity, NoNaN, NoInf, NoNegativeZero — 5 each for byte and rune surfaces; Symmetric INTENTIONALLY OMITTED per OQ-1) plus `TestProp_RatcliffObershelpScore_AtLeastLevenshtein_HandCurated` (hand-curated substring-containment property — RESEARCH.md notes it is "generally" true, not universal). The section header explicitly cites the OQ-1 resolution.
- `tests/bdd/steps/algorithms_steps.go` — appended two step methods (`iComputeTheRatcliffObershelpScoreBetween`, `iComputeTheRatcliffObershelpScoreForTheAutojunkSensitivePair`) and their `ctx.Step` regex registrations inside `InitializeScenario`. Added `strings` to imports for the autojunk-sensitive Go constants `roAutojunkA` / `roAutojunkB` (each 205 chars, computed via `strings.Repeat`).
- `llms.txt` — appended `### Ratcliff-Obershelp (Dr. Dobb's 1988) gestalt-pattern-matching similarity` section with the two exported symbols. Required by the `TestAIFriendly_LLMSTxtReferencesEveryExportedSymbol` meta-test (per-plan llms.txt-sync discipline inherited from plan 04-01).

## Decisions Made

The five "key-decisions" entries in the frontmatter capture the substantive choices:

1. **Drop the Symmetric property test for Ratcliff-Obershelp.** OQ-1 resolution LOCKED 2026-05-14. RO is intentionally asymmetric to preserve byte-for-byte difflib equivalence; a Symmetric property test would either fail on counterexamples like tide/diet OR force a hidden workaround. The standard Identity / RangeBounds / NoNaN / NoInf / NoNegativeZero invariants still apply on both byte and rune surfaces.
2. **Inline a substring-position-returning DP variant rather than reusing lcsstr.go's helper.** CONTEXT.md D-3 allowed either path; inline was chosen because lcsstrDP returns (length, endI) but RO needs aLo + bLo + bHi to drive the recursion.
3. **Language-native call stack for the recursion** (CONTEXT.md D-2). For reasonable inputs the call-stack depth is bounded by O(min(la, lb)).
4. **AsymmetryPin in BOTH the local test file AND the plan-04-05 cross-algorithm consistency test.** Two layers of defence: local pin is fast; cross-algorithm test catches catalog-wide regressions.
5. **BDD autojunk-sensitive scenario uses Go-side constants** computed via `strings.Repeat` rather than 205-char Gherkin literals (auto-fixed from a hand-counted const literal — see Deviations).

## Deviations from Plan

### Auto-fixed Issues

**1. [Rule 1 - Bug] Hand-counted const string literals were off-by-3 / off-by-1**
- **Found during:** Task 3 (BDD step implementation)
- **Issue:** Initial draft of `roAutojunkA` and `roAutojunkB` used hand-typed string literals — 'a'×100 + 'x'×5 + 'a'×100 should be 205 chars, but the literal came in at 208 chars for A and 204 chars for B (hand-counting errors). The result would have caused a numerical mismatch with the pinned difflib value 0.7317 in the BDD scenario.
- **Fix:** Switched the declarations from `const (...)` to `var (...)` using `strings.Repeat("a", 100) + strings.Repeat("x", 5) + strings.Repeat("a", 100)` so the character counts are arithmetic rather than hand-counted. Added `"strings"` to the imports of `tests/bdd/steps/algorithms_steps.go`. This makes the lengths self-documenting and uncountable-incorrectly.
- **Files modified:** tests/bdd/steps/algorithms_steps.go
- **Verification:** `make test-bdd` exits 0 with the autojunk-sensitive scenario green; the pinned 0.7317 matches `difflib.SequenceMatcher(autojunk=False, a=a, b=b).ratio()` byte-for-byte at the 4-dp tolerance.
- **Committed in:** bc7fb2a (Task 3 commit)

**2. [Rule 2 - Missing Critical] llms.txt sync for the two new exported symbols**
- **Found during:** Task 1 (per-plan discipline inherited from plan 04-01 / 04-02 SUMMARY)
- **Issue:** Plan instructions did not explicitly call out llms.txt as a touched file, but the `TestAIFriendly_LLMSTxtReferencesEveryExportedSymbol` meta-test in `ai_friendly_test.go` would have failed once Task 1 landed without the llms.txt update. Plan 04-01 and 04-02 SUMMARYs both documented this same pattern as a Rule 2 deviation.
- **Fix:** Appended a `### Ratcliff-Obershelp (Dr. Dobb's 1988) gestalt-pattern-matching similarity` section to llms.txt listing the two new exported symbols (RatcliffObershelpScore, RatcliffObershelpScoreRunes).
- **Files modified:** llms.txt
- **Verification:** `go test ./...` passes (meta-test green); the section appears in the canonical position immediately after the LCSStr section, before Normalisation.
- **Committed in:** 1c0c5be (Task 1 commit — bundled with the rest of the algorithm landing)

---

**Total deviations:** 2 auto-fixed (1 hand-counting bug in BDD constants, 1 missing critical llms.txt update — same pattern documented in plans 04-01 and 04-02 SUMMARYs)
**Impact on plan:** Both auto-fixes essential for correctness (#1 would have failed the BDD scenario at the 0.0001 tolerance; #2 would have failed the AI-friendly meta-test). No scope creep — the RO public surface remains exactly the two spec-pinned functions.

## Issues Encountered

- **The verify-command grep gate for "no Symmetric prop test" was too broad.** The plan's verify command `! grep -E "TestProp_RatcliffObershelp.*_Symmetric" props_test.go` matches the comment `// NB: TestProp_RatcliffObershelpScore_Symmetric is INTENTIONALLY OMITTED ...` that explains why the omission is deliberate. The narrower pattern `! grep -E '^func TestProp_RatcliffObershelp.*_Symmetric' props_test.go` correctly verifies that no test function with that name is DEFINED. The intent is satisfied (no Symmetric prop function exists; the comment explicitly documents the OMISSION). Documenting here for future verifier review.
- **The BDD `symmetric` word also appears in the feature header comment and the Feature description.** The comment says "NOT symmetric in argument order" and the feature description says "NOT symmetric in argument order". The actual Scenario lines do NOT mention symmetry. Narrower grep gate `! grep -E "^\s+Scenario.*symmetr"` correctly verifies the intent.
- **Allocation count on ASCII_Short is 4 not 2.** The byte path's `roFindLongestMatch` allocates 2 rolling-row slices per recursion level. WIKIMEDIA/WIKIMANIA recursion depth happens to be 2 levels for this input, producing 4 allocations total. This is within budget for an algorithm with data-dependent recursion depth — documented in the bench file's godoc as informational.

## Threat Surface Scan

The `<threat_model>` block in the plan enumerates three threats. All three mitigations land in this plan:

- **T-fuzz-panic (mitigate):** `FuzzRatcliffObershelpScore` exercises BOTH public surfaces (Phase 3 WR-02 closure); 10-second smoke = 1.6M execs, no crashes; seed corpus includes invalid UTF-8 (\xff\xfe, \xc0\x80), Cyrillic (Привет/привет), Latin supplement (café/cafe), and the autojunk-sensitive 205-char pair.
- **T-complexity-attack (accept):** Recursive longest-common-substring decomposition is O(n²·m) for repeated-character pathological inputs; recursion depth is O(min(la, lb)) so no stack-overflow risk. Long-input bench `BenchmarkRatcliffObershelpScore_ASCII_Long` (500 chars × 500 chars at 12.7ms) establishes the regression baseline. Pure-function library — caller controls input size.
- **T-float-determinism (mitigate):** Explicit left-to-right `numer := 2.0 * float64(m); denom := float64(la + lb); return numer / denom` per DET-06; no `math.Pow`/`Log`/`Exp`/`FMA` (verified by grep gate). Cross-platform CI matrix will verify byte-identical output via `testdata/golden/_staging/ratcliff_obershelp.json` once merged in plan 04-05.

No new threat surface introduced. Omitting Threat Flags section.

## User Setup Required

None — no external service configuration required.

## Next Phase Readiness

- **Plan 04-04 (cross-validation)** will add `scripts/gen-ratcliff-obershelp-cross-validation.py` (Python stdlib only, autojunk=False, Python version assertion >= 3.7) and `testdata/cross-validation/ratcliff-obershelp/vectors.json` (~15-18 entries covering all four mandatory categories from CONTEXT.md §1). A new `TestRatcliffObershelp_CrossValidation` will be appended to `ratcliff_obershelp_test.go` — the test reads the JSON, asserts `|our_score - difflib_ratio| <= 1e-9` for every entry, no Python at test time. The Makefile gets `regen-ratcliff-obershelp-cross-validation` (developer-only). The numerical-regression pin in this plan (TestRatcliffObershelp_PinnedDrDobbsValue) catches regressions even before the corpus runs.
- **Plan 04-05 (finalisation)** will merge `testdata/golden/_staging/ratcliff_obershelp.json` (and the strcmp95 + lcsstr stagings) into `testdata/golden/algorithms.json`. The cross-algorithm consistency test will gain three new tests: `Strcmp95 >= JaroWinkler`, `LCSStr >= Levenshtein` on substring-containment input, and `RatcliffObershelp pinned against difflib on at least one pair where it visibly differs from both Levenshtein and Jaro-Winkler`. The cross-algorithm asymmetry-pin (fwd != rev for tide/diet) is the inverse-form regression guard alongside this plan's local pin.
- **Cross-platform determinism:** the staging file's `expected_score` values (including the irrational 0.45161290322580644 and 0.7777777777777778) are byte-stable on darwin/arm64; the Phase 1 CI matrix gate will re-verify across linux/amd64, linux/arm64, darwin/amd64, and windows/amd64 when plan 04-05 merges.
- **Phase 6's TokenSortRatio / TokenSetRatio / PartialRatio godoc must point users wanting `difflib.ratio()` semantics at `RatcliffObershelpScore`.** This plan ships the algorithm with the difflib-equivalence directive in godoc; Phase 6 ships the cross-reference. Mark this as a CONTEXT.md `<follow_ups>` item already tracked.

## Self-Check: PASSED

- **Files created:** `ratcliff_obershelp.go`, `dispatch_ratcliff_obershelp.go`, `ratcliff_obershelp_test.go`, `ratcliff_obershelp_bench_test.go`, `ratcliff_obershelp_fuzz_test.go`, `testdata/golden/_staging/ratcliff_obershelp.json`, `testdata/fuzz/FuzzRatcliffObershelpScore/seed-001`, `tests/bdd/features/ratcliff_obershelp.feature` — all FOUND on disk.
- **Files modified:** `algoid_test.go`, `algorithms_golden_test.go`, `example_test.go`, `props_test.go`, `tests/bdd/steps/algorithms_steps.go`, `llms.txt` — all show as touched in `git diff --name-only HEAD~3 HEAD`.
- **Commits exist:** `1c0c5be` (feat), `a019493` (test), `bc7fb2a` (test) — all confirmed via `git log --oneline -3`.
- **Verification commands green:**
  - `go build ./...` → ok
  - `go test -run 'TestRatcliffObershelp|TestProp_RatcliffObershelp|TestDispatch_RatcliffObershelpRegistered|TestGolden_RatcliffObershelp_Staging|ExampleRatcliffObershelp' ./...` → ok
  - `go test -bench=BenchmarkRatcliffObershelp -benchmem -benchtime=1x ./...` → 5 benches; allocations informational
  - `go test -fuzz=FuzzRatcliffObershelpScore -fuzztime=10s ./...` → 1.6M execs, no crashes
  - `make test-bdd` → ok with 5 RO scenarios visible and green
  - `bash scripts/verify-license-headers.sh` → 80 .go files OK
  - `! grep -q "^func init" ratcliff_obershelp.go` → OK (no init())
  - `! grep -E "math\\.(Pow|Log|Exp|FMA)" ratcliff_obershelp.go` → OK (no transcendentals)
  - `grep -q "// Source: Ratcliff" ratcliff_obershelp.go` → OK (Ratcliff & Metzener 1988 cited)
  - `grep -q "difflib.SequenceMatcher(autojunk=False" ratcliff_obershelp.go` → OK (autojunk=False directive in godoc)
  - `grep -q "NOT symmetric in argument order" ratcliff_obershelp.go` → OK (OQ-1 resolution documented in godoc)
  - `! grep -E '^func TestProp_RatcliffObershelp.*_Symmetric' props_test.go` → OK (no Symmetric prop function defined)
  - `go vet ./...` → ok
- **Coverage:** `ratcliff_obershelp.go` reports 100% on the two public functions (RatcliffObershelpScore, RatcliffObershelpScoreRunes); 100% on roMatchedLength / roMatchedLengthRunes; 96.0% on roFindLongestMatch and roFindLongestMatchRunes (above the ≥ 90% per-file floor).

---

*Phase: 04-remaining-character-gestalt*
*Completed: 2026-05-14*
