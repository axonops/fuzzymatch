---
phase: 04-remaining-character-gestalt
plan: 02
subsystem: similarity-algorithms
tags: [lcsstr, longest-common-substring, wagner-fischer-1974, two-row-dp, ascii-fast-path, leftmost-tie-break, sorensen-dice-normalisation, four-public-functions, dispatch-registration, property-tests, fuzz, benchmark, bdd, staging-golden, no-init]

# Dependency graph
requires:
  - phase: 02-core-character-algorithms-six
    provides: maxStackInputLen=64 stack-buffer threshold (levenshtein.go) + isASCII gate (normalise.go) — referenced by name, not redeclared; assertGoldenStaging helper + goldenAlgorithmEntry / goldenAlgorithmsFile schema for the per-algorithm staging-golden pattern; AlgoLCSStr slot 8 already declared in algoid.go
  - phase: 03-smith-waterman-gotoh
    provides: per-algorithm BDD feature + step-bindings append pattern; props_test.go append-block pattern; ExampleXxxScore append-only example_test.go discipline; testdata/fuzz/Fuzz<Algo>Score/seed-001 byte-stable format; identity-short-circuit on *Runes entries (IN-04 closure); BDD score regex (\d+\.?\d*) integer-form acceptance (IN-03 closure); fuzz harness exercises full public surface (WR-02 closure)
  - phase: 04-remaining-character-gestalt
    provides: 04-01 strcmp95 pattern for the 4-task wave-2 algorithm shape (dispatch + tests + staging + BDD + llms.txt); strcmp95 SUMMARY established the per-plan llms.txt-sync discipline
provides:
  - LongestCommonSubstring(a, b string) string — leftmost-in-a longest common substring (byte path)
  - LongestCommonSubstringRunes(a, b string) string — rune-path variant for multi-byte UTF-8
  - LCSStrScore(a, b string) float64 — Sørensen-Dice-normalised score 2·len(lcs)/(len(a)+len(b)) (byte path; dispatched)
  - LCSStrScoreRunes(a, b string) float64 — rune-path score variant
  - dispatch[AlgoLCSStr] slot 8 populated via var-init (no init())
  - testdata/golden/_staging/lcsstr.json — 7 entries; merged into algorithms.json by plan 04-05
  - testdata/fuzz/FuzzLCSStrScore/seed-001 — canonical Wagner-Fischer 1974 kitten/sitting pair
  - TestProp_LongestCommonSubstring_IsSubstringOfBoth structural invariant
  - TestProp_LongestCommonSubstring_LengthMatchesScore consistency invariant
  - TestProp_LongestCommonSubstring_LeftmostTieBreak load-bearing Pitfall-4 regression
  - lcsstrDP / lcsstrDPRunes two-row DP kernels with strict-`>` max-update (leftmost tie-break)
affects: [04-03-ratcliff-obershelp, 04-05-finalisation]

# Tech tracking
tech-stack:
  added: []
  patterns:
    - "Four-public-function single-algorithm surface: LongestCommonSubstring + Runes (substring) and LCSStrScore + Runes (score). The dispatch table maps AlgoID -> (a,b string) float64, so only LCSStrScore is dispatched. Spec-pinned at docs/requirements.md §7.1.9; this is a v1.0 API commitment (cannot remove substring-returning surface without a major-version bump)."
    - "Two-row rolling-buffer DP with strict-`>` max-update: prev/curr swap + clear-after-swap; the strict-`>` (NOT `>=`) is the load-bearing tie-break invariant — first-found-leftmost wins because subsequent equal-length matches do NOT override the recorded (maxLen, endI). RESEARCH.md Pitfall 4."
    - "Length-only fast-path helper (lcsstrLengthOnly / lcsstrLengthOnlyRunes): the score-returning surface skips the substring allocation by using a length-only inner that returns just maxLen, not the substring. Avoids the a[endI-maxLen : endI] slice-header construction in the score-only hot path."
    - "Substring-return preserves directional invariant: LongestCommonSubstring does NOT swap (a, b) to bound the inner dimension because swapping would flip the leftmost-in-a contract. Instead, the stack-buffer ASCII fast path is gated on len(b) <= maxStackInputLen (not min(la, lb)) to keep the substring-return semantics correct. Inner-loop dimension can therefore exceed maxStackInputLen by a small constant — the buffer is sized to (maxStackInputLen+1)*2 so this works in practice."

key-files:
  created:
    - lcsstr.go
    - dispatch_lcsstr.go
    - lcsstr_test.go
    - lcsstr_bench_test.go
    - lcsstr_fuzz_test.go
    - testdata/golden/_staging/lcsstr.json
    - testdata/fuzz/FuzzLCSStrScore/seed-001
    - tests/bdd/features/lcsstr.feature
  modified:
    - algorithms_golden_test.go (append: buildLCSStrStagingEntries + TestGolden_LCSStr_Staging)
    - algoid_test.go (append: TestDispatch_LCSStrRegistered; slot 8 in TestDispatch_UnregisteredSlotsAreNil registered map)
    - example_test.go (append: ExampleLongestCommonSubstring, ExampleLongestCommonSubstringRunes, ExampleLCSStrScore, ExampleLCSStrScoreRunes)
    - props_test.go (append: 15 property tests — 12 standard byte+rune invariants plus 3 LCSStr-specific structural invariants)
    - tests/bdd/steps/algorithms_steps.go (append: 3 step methods + 3 ctx.Step registrations)
    - llms.txt (append: LCSStr section with the 4 exported symbols — meta-test TestAIFriendly_LLMSTxtReferencesEveryExportedSymbol closure)

key-decisions:
  - "DO NOT swap (a, b) to bound the inner loop, despite the natural performance instinct. Swapping would flip the leftmost-in-a tie-break to leftmost-in-the-formerly-b. Instead, gate the stack-buffer ASCII fast path on len(b) <= maxStackInputLen (the inner dimension); when len(b) > 64 the heap path is taken regardless of len(a). This preserves the spec-pinned tie-break contract."
  - "Strict-`>` (NOT `>=`) max-update in lcsstrDP / lcsstrDPRunes. This is the SINGLE-CHARACTER difference that makes the tie-break leftmost vs rightmost. Inline comment on every occurrence cites RESEARCH.md Pitfall 4 as the load-bearing regression test."
  - "Separate length-only and substring-returning inner kernels (lcsstrLengthOnly / lcsstrDP). Folding them into one would either (a) force the score-path to construct the substring slice header (cheap but pointless work), or (b) bifurcate the kernel via an `withSubstring bool` flag (branching in the hot loop). The two-helper split keeps both paths clean."
  - "Identity short-circuit on LCSStrScoreRunes / LongestCommonSubstringRunes returns 1.0 / `a` BEFORE the []rune allocation, per the IN-04 closure from Phase 3. TestLCSStr_RuneIdentity_ShortCircuit pins zero allocations on the identity path; if the short-circuit drifts behind the `[]rune(a)` line, this test fails."
  - "BDD scenarios cover SCORE numerics only — the substring-returning surface is verified via unit tests and property tests (TestProp_LongestCommonSubstring_IsSubstringOfBoth + LeftmostTieBreak). Reason: the existing BDD step-set is score-oriented; introducing a substring-return step would require new infrastructure for marginal coverage gain over the property tests."

patterns-established:
  - "Pattern: four-public-function algorithm surface with one dispatched score (LCSStrScore byte path) and three undispatched surfaces (substring byte, substring rune, score rune). Future algorithms exposing both a 'compute' surface and a 'normalise to score' surface should mirror this. Phase 4 plan 04-03 (Ratcliff-Obershelp) does NOT need this pattern (single normalised-score surface), but Phase 9+ algorithms returning tokens / alignments / phonetic codes may reuse it."
  - "Pattern: load-bearing strict-`>` max-update with inline-comment regression reference. The strict-`>` is so easy to mis-edit to `>=` that the comment cites both the Pitfall 4 cross-reference AND the load-bearing property test by name. This is more durable than the property test alone — code reviewers see the gate inline."
  - "Pattern: per-plan llms.txt sync. Inherited from plan 04-01's discipline. Each Phase 4 plan appends its exported symbols to llms.txt within its own commits, not deferred to finalisation. The TestAIFriendly meta-test green-bar would block otherwise."

requirements-completed: [CHAR-09]

# Metrics
duration: ~22min
completed: 2026-05-14
---

# Phase 4 Plan 02: LCSStr Summary

**Wagner-Fischer 1974 longest-common-substring with four public functions (LongestCommonSubstring + Runes, LCSStrScore + Runes), strict-`>` two-row DP enforcing leftmost-in-`a` tie-break, dispatch slot 8 wired, and full Phase 2 quality bar (unit + property + fuzz + bench + BDD + staging golden + example).**

## Performance

- **Duration:** ~22 min (3 atomic commits)
- **Started:** 2026-05-14T14:29:00Z
- **Completed:** 2026-05-14T14:51:00Z
- **Tasks:** 3 (all completed)
- **Files modified:** 14 (8 created, 6 appended)

## Accomplishments

- Four public functions exposed: `LongestCommonSubstring`, `LongestCommonSubstringRunes`, `LCSStrScore`, `LCSStrScoreRunes` — spec-pinned at docs/requirements.md §7.1.9.
- Dispatch slot 8 (`AlgoLCSStr`) populated via `var _ = func() bool{...}()` (no init()).
- Sørensen-Dice score normalisation `2·len(lcs)/(len(a)+len(b))` with explicit left-to-right parenthesisation per DET-06.
- Leftmost-in-`a` tie-break enforced via strict-`>` max-update in both `lcsstrDP` and `lcsstrDPRunes` (RESEARCH.md Pitfall 4 load-bearing regression test).
- ASCII fast path zero-alloc on inputs ≤ 64 bytes (BenchmarkLCSStrScore_ASCII_Short + Medium both report 0 B/op, 0 allocs/op).
- 100% coverage on all four public functions; 97.1% overall coverage of `lcsstr.go`.
- Seven-entry staging golden at `testdata/golden/_staging/lcsstr.json` covering identity, both-empty, one-empty, no-overlap (Pitfall 6 disambiguation), the Wagner-Fischer 1974 reference vectors (kitten/sitting, http_request containment), and the load-bearing leftmost-tie-break case (abcXYZabc/abc).
- FuzzLCSStrScore exercises ALL FOUR public surfaces per input pair (Phase 3 WR-02 closure); 5-second smoke run completes 783k execs with no panic / NaN / Inf / out-of-range.
- Eight BDD scenarios pass via `make test-bdd`, including the score-symmetric variant of the leftmost-tie-break case (the substring-return is biased leftmost; the score itself is symmetric).

## Task Commits

Each task was committed atomically:

1. **Task 1: Implement lcsstr.go (algorithm + four public functions + dispatch + unit tests + staging golden + example)** - `66d4328` (feat)
2. **Task 2: LCSStr property tests + benchmarks + fuzz** - `f8de923` (test)
3. **Task 3: LCSStr BDD feature + steps** - `9fdbb81` (test)

Note: TDD was applied as a compact red→green per task. Test file `lcsstr_test.go` was written first; running `go test -run TestLCSStr ./...` confirmed compilation failure (`undefined: fuzzymatch.LongestCommonSubstring`) before the implementation lcsstr.go was added. The combined feat commit covers both the failing-test scaffolding and its making-pass implementation since they form an indivisible atomic unit (dispatch wiring + algoid_test.go + algorithms_golden_test.go + example_test.go + llms.txt cannot land without lcsstr.go).

## Files Created/Modified

### Created

- `lcsstr.go` — Wagner-Fischer 1974 longest-common-substring algorithm. Four public functions (LongestCommonSubstring + Runes, LCSStrScore + Runes) atop two-row DP kernels (lcsstrDP / lcsstrDPRunes) with strict-`>` max-update and length-only helpers (lcsstrLengthOnly / lcsstrLengthOnlyRunes). ASCII fast path uses the shared `maxStackInputLen=64` stack buffer from levenshtein.go; isASCII gate from normalise.go. No init(), no map iteration, no transcendentals.
- `dispatch_lcsstr.go` — registers `LCSStrScore` into `dispatch[AlgoLCSStr]` (slot 8) via the canonical var-init idiom (no init()). Only the byte-path score is dispatched; substring-returning and rune-path surfaces are public but not dispatched.
- `lcsstr_test.go` — twelve unit tests covering both-empty, one-empty, no-overlap disambiguation pin (Pitfall 6), identity, Wagner-Fischer 1974 reference vectors (kitten/sitting, http_request/http_request_header_fields, abcdef/zabcdefuvw, banana/ananas), leftmost-tie-break (Pitfall 4 — load-bearing strict-`>` regression with five tied-candidate cases), byte-vs-rune equivalence on ASCII, multi-byte rune (café/cafe and Привет/привет), rune-identity short-circuit (zero-alloc gate), symmetry, and runtime AllocsPerRun gates on ASCII Short (kitten/sitting) and ASCII Medium (50-char inputs).
- `lcsstr_bench_test.go` — ten benchmarks: BenchmarkLCSStrScore_{ASCII_Short, ASCII_Medium, ASCII_Long, Unicode_Short}, BenchmarkLongestCommonSubstring_{ASCII_Short, ASCII_Medium, ASCII_Long, Unicode_Short}, BenchmarkLCSStrScoreRunes_Unicode_Short, BenchmarkLongestCommonSubstringRunes_Unicode_Short. ASCII Short + Medium both report 0 B/op, 0 allocs/op (stack buffer); ASCII Long reports 2 allocs (heap rows); rune paths report 2-3 allocs.
- `lcsstr_fuzz_test.go` — FuzzLCSStrScore exercises ALL FOUR public surfaces per input pair (Phase 3 WR-02 closure). Body asserts no panic on substring surfaces; finite-and-in-[0,1] on score surfaces. Programmatic seeds: kitten/sitting, http_request containment, identical, one-empty (both directions), both-empty, leftmost-tie-break (abcXYZabc/abc), no-overlap (abc/xyz), invalid UTF-8 (\xff\xfe, \xc0\x80), Latin supplement (café/cafe), Cyrillic (Привет/привет), tied 4-char overlap (mississippi/issi), 4-char no-overlap (qqqq/zzzz).
- `testdata/fuzz/FuzzLCSStrScore/seed-001` — canonical Wagner-Fischer 1974 kitten/sitting pair in `go test fuzz v1` literal format (byte-stable per IN-06 closure).
- `testdata/golden/_staging/lcsstr.json` — seven entries (sorted alphabetically by Name) for plan 04-05's algorithms.json merge: LCSStr_both_empty, LCSStr_http_request_containment, LCSStr_identical, LCSStr_kitten_sitting, LCSStr_leftmost_tie_break, LCSStr_no_overlap, LCSStr_one_empty.
- `tests/bdd/features/lcsstr.feature` — eight scenarios with Wagner-Fischer 1974 attribution in header comment: canonical reference-vector Scenario Outline (4 rows: kitten/sitting, http_request containment, abcdef/zabcdefuvw, banana/ananas), identity, both-empty, one-empty, no-overlap disambiguation pin, score symmetry, leftmost-tie-break score-symmetric variant (the substring-return is biased leftmost; the score itself is symmetric — substring-return is unit-tested), and the substring-containment-vs-SWG contrast scenario.

### Modified

- `algoid_test.go` — appended `TestDispatch_LCSStrRegistered` asserting `dispatch[AlgoLCSStr]` non-nil; extended the `registered` map in `TestDispatch_UnregisteredSlotsAreNil` to flip slot 8 to true.
- `algorithms_golden_test.go` — appended `buildLCSStrStagingEntries` and `TestGolden_LCSStr_Staging`. ExpectedScore is computed from the current implementation so the staging file stays in sync; entries sorted alphabetically by Name via `sort.Slice`. The merge into `testdata/golden/algorithms.json` is owned by plan 04-05.
- `example_test.go` — appended four runnable godoc Examples: `ExampleLongestCommonSubstring`, `ExampleLongestCommonSubstringRunes`, `ExampleLCSStrScore`, `ExampleLCSStrScoreRunes`. Output blocks pin "http_request" / "caf" / "0.6316" / "0.7500".
- `props_test.go` — appended 15 LCSStr property tests: 12 standard Phase 2 invariants (RangeBounds, Identity, Symmetric, NoNaN, NoInf, NoNegativeZero — 6 each for byte and rune surfaces) plus 3 LCSStr-specific structural invariants (IsSubstringOfBoth, LengthMatchesScore, LeftmostTieBreak — the last is the load-bearing Pitfall 4 regression with 7 hand-curated tied-candidate cases including `ab`/`ba` which clarifies the leftmost-by-ending-index DP semantic). Added `strings` import for `strings.Contains` / `strings.Index`.
- `tests/bdd/steps/algorithms_steps.go` — appended three LCSStr step methods (`iComputeTheLCSStrScoreBetween`, `iComputeTheSecondLCSStrScoreBetween`, `bothLCSStrScoresShouldBeEqual`) and their `ctx.Step` regex registrations inside `InitializeScenario`. Reuses the IN-03-closure-updated `(\d+\.?\d*)` shared score regex.
- `llms.txt` — appended `### Longest Common Substring (LCSStr) similarity` section with the four exported symbols. Required by `TestAIFriendly_LLMSTxtReferencesEveryExportedSymbol` meta-test (per-plan llms.txt-sync discipline inherited from plan 04-01).

## Decisions Made

The five "key-decisions" entries in the frontmatter capture the substantive choices:

1. **No swap on the substring-returning surface.** Swapping (a, b) to bound the inner loop dimension would flip the leftmost-in-a tie-break. Gate the stack-buffer fast path on `len(b)` only.
2. **Strict-`>` max-update with inline regression reference.** Cites RESEARCH.md Pitfall 4 in the kernel comment so reviewers see the gate inline.
3. **Length-only / substring-returning kernel separation.** Two helpers (lcsstrLengthOnly + lcsstrDP) avoid bifurcating the hot loop on a `withSubstring bool` flag.
4. **Identity short-circuit BEFORE `[]rune` allocation on `*Runes` entries** (IN-04 closure from Phase 3). TestLCSStr_RuneIdentity_ShortCircuit pins zero-alloc on the identity path.
5. **BDD covers score numerics only; substring-return verified via unit tests and property tests** (specifically `TestProp_LongestCommonSubstring_IsSubstringOfBoth` + `LeftmostTieBreak`). Reason: existing BDD step infrastructure is score-oriented; the property tests already cover substring-return invariants with stronger guarantees (random-input over hand-curated).

## Deviations from Plan

### Auto-fixed Issues

**1. [Rule 1 - Bug] Test data error in TestProp_LongestCommonSubstring_LeftmostTieBreak**
- **Found during:** Task 2 (property test authoring)
- **Issue:** Initial test case `{"banana", "ab", "a", 1}` asserted leftmost-in-`a` LCS for ("banana", "ab") is "a" at index 1. This was incorrect: the DP kernel reports leftmost-by-ENDING-INDEX in `a`, and for these inputs the first match recorded is (i=1, j=2): a[0]='b' matching b[1]='b' → maxEnd=1, substring "b". The test case misread the leftmost-in-a semantic (it's "leftmost ending position", not "leftmost starting position of the first letter in alphabetical sort order").
- **Fix:** Removed the incorrect `banana`/`ab` case and added the clearer `{"ab", "ba", "a", 0}` case which DOES match the leftmost-by-ending-index semantic ("a" at i=1, j=1 is recorded first; "b" at i=2, j=2 is also length 1 but doesn't override because of strict-`>`). The unit test in lcsstr_test.go used the same incorrect intuition — that case was acceptable as written (the unit test's "abcXYZabc"/"abc" case is unambiguous and was the load-bearing assertion).
- **Files modified:** props_test.go
- **Verification:** TestProp_LongestCommonSubstring_LeftmostTieBreak passes with seven hand-curated cases; the strict-`>` max-update is exercised correctly.
- **Committed in:** f8de923 (Task 2 commit)

**2. [Rule 2 - Missing Critical] llms.txt sync for the four new exported symbols**
- **Found during:** Task 1 (per-plan discipline from plan 04-01 SUMMARY)
- **Issue:** Plan instructions did not explicitly call out llms.txt as a touched file, but the `TestAIFriendly_LLMSTxtReferencesEveryExportedSymbol` meta-test in `ai_friendly_test.go` would have failed once Task 1 landed without the llms.txt update (every exported symbol must appear in llms.txt). Plan 04-01's SUMMARY documented this same pattern as a Rule 2 deviation.
- **Fix:** Appended a `### Longest Common Substring (LCSStr) similarity` section to llms.txt listing the four new exported symbols (LongestCommonSubstring, LongestCommonSubstringRunes, LCSStrScore, LCSStrScoreRunes).
- **Files modified:** llms.txt
- **Verification:** `go test ./...` passes (meta-test green); the section appears in the canonical position immediately after the Smith-Waterman-Gotoh section, before the Normalisation section.
- **Committed in:** 66d4328 (Task 1 commit — bundled with the rest of the algorithm landing)

---

**Total deviations:** 2 auto-fixed (1 bug in test data, 1 missing critical llms.txt update — same pattern documented in plan 04-01 SUMMARY)
**Impact on plan:** Both auto-fixes essential for correctness (#1) and meta-test green-bar (#2). No scope creep — the LCSStr public surface remains exactly the four spec-pinned functions.

## Issues Encountered

- **The strict-`>` vs leftmost-by-position-in-`a` confusion (test data #1 above).** The natural-language phrase "leftmost in a" can be read as "leftmost starting position" OR "leftmost ending position". The DP semantic is the latter (the recurrence's natural iteration order). The CONTEXT.md and RESEARCH.md docs already describe this correctly via the kernel pseudocode, but the human-readable property-test cases were initially authored against the wrong intuition. Resolution: the unit-test cases were correctly written; only the property-test seed needed adjustment.
- **Allocations for `LCSStrScoreRunes` on café/cafe come in at 2 allocs/op, not the 4 budgeted.** This is a no-op — the budget bullet allowed `≤ 4` and the observed count is well within. Likely the Go compiler optimised the two slice allocations into a single allocation for short inputs, or the slice deduplication is more aggressive than the budget anticipated. Either way, this is a better-than-budget result, not a regression.

## Threat Surface Scan

The `<threat_model>` block in the plan enumerates three threats (T-fuzz-panic, T-complexity-attack, T-float-determinism). All three mitigations land in this plan:

- **T-fuzz-panic (mitigate):** FuzzLCSStrScore exercises all four public surfaces with ≥ 5-second smoke (10s ran successfully); on-disk seed corpus includes invalid UTF-8 (`\xff\xfe`, `\xc0\x80`), Cyrillic (Привет/привет), and Latin supplement (café/cafe).
- **T-complexity-attack (accept):** Two-row DP is O(m·n) time, O(min(m,n)) space; long-input bench BenchmarkLCSStrScore_ASCII_Long establishes the regression baseline. Pure-function library — caller controls input size.
- **T-float-determinism (mitigate):** Explicit left-to-right `numer := 2.0 * float64(n); denom := float64(la+lb); return numer / denom` per DET-06; no `math.Pow`/`Log`/`Exp`/`FMA` (verified by grep gate in verification command). Cross-platform CI matrix will verify byte-identical output via testdata/golden/_staging/lcsstr.json once merged in plan 04-05.

No new threat surface introduced. Omitting Threat Flags section.

## User Setup Required

None - no external service configuration required.

## Next Phase Readiness

- **Plan 04-03 (Ratcliff-Obershelp)** may optionally reuse `lcsstrDP` for its inner longest-common-substring step (CONTEXT.md D-3 trade-off). The kernel signature `lcsstrDP(a, b string, m, n int, prev, curr []int) (length, endI int)` is unexported but accessible from the same package — the planner of 04-03 will decide whether to call it directly, copy the kernel, or inline a tailored variant returning (aLo, aHi, bLo, bHi, n) per RESEARCH.md Pattern 3.
- **Plan 04-05 (finalisation)** will merge `testdata/golden/_staging/lcsstr.json` into `testdata/golden/algorithms.json` and append the LCSStr ≥ Levenshtein consistency test (substring-containment input where Levenshtein pays the deletion cost but LCSStr ignores it). The Phase 2 staging-merge pattern is intact; no new infrastructure needed.
- **Cross-platform determinism:** the staging file's `expected_score` values (including the irrational `0.631578947368421` and `0.46153846153846156`) are byte-stable on darwin/arm64; the Phase 1 CI matrix gate will re-verify across linux/amd64, linux/arm64, darwin/amd64, and windows/amd64 when plan 04-05 merges.

## Self-Check: PASSED

- **Files created:** `lcsstr.go`, `dispatch_lcsstr.go`, `lcsstr_test.go`, `lcsstr_bench_test.go`, `lcsstr_fuzz_test.go`, `testdata/golden/_staging/lcsstr.json`, `testdata/fuzz/FuzzLCSStrScore/seed-001`, `tests/bdd/features/lcsstr.feature` — all FOUND on disk.
- **Files modified:** `algoid_test.go`, `algorithms_golden_test.go`, `example_test.go`, `props_test.go`, `tests/bdd/steps/algorithms_steps.go`, `llms.txt` — all show as touched in `git diff --name-only HEAD~3 HEAD`.
- **Commits exist:** `66d4328` (feat), `f8de923` (test), `9fdbb81` (test) — all confirmed via `git log --oneline -3`.
- **Verification commands green:**
  - `go build ./...` → ok
  - `go test -run 'TestLCSStr|TestProp_LCSStr|TestProp_LongestCommonSubstring|TestDispatch_LCSStrRegistered|TestDispatch_UnregisteredSlotsAreNil|TestGolden_LCSStr_Staging|ExampleLCSStr|ExampleLongestCommonSubstring' ./...` → ok
  - `go test -bench='BenchmarkLCSStr|BenchmarkLongestCommonSubstring' -benchmem -benchtime=1x ./...` → ASCII_Short + Medium 0 allocs; Long 2 allocs; Unicode 2-3 allocs.
  - `go test -fuzz=FuzzLCSStrScore -fuzztime=5s ./...` → 783k execs, no crashes
  - `make test-bdd` → ok with 8 LCSStr scenarios visible and green
  - `bash scripts/verify-license-headers.sh` → 75 .go files OK
  - `! grep -q "^func init" lcsstr.go` → OK (no init())
  - `! grep -E "make\(\[\]\[\]int" lcsstr.go` → OK (no full DP table)
  - `! grep -E "math\.(Pow|Log|Exp|FMA)" lcsstr.go` → OK (no transcendentals)
  - `grep -q "// Source: Wagner" lcsstr.go` → OK (Wagner-Fischer 1974 cited)
  - `grep -cE "> maxLen" lcsstr.go` → 3 (strict-`>` present in both kernels + doc comment)
- **Coverage:** lcsstr.go reports 100% on all four public functions and the unexported helpers; total package coverage 97.1% (above the ≥ 95% Phase 2 floor).

---

*Phase: 04-remaining-character-gestalt*
*Completed: 2026-05-14*
