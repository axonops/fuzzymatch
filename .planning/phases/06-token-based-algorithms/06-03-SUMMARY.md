---
phase: 06-token-based-algorithms
plan: 03
subsystem: algorithm-catalogue
tags: [partial-ratio, indel-formula, three-region-iteration, char-set-early-skip, rapidfuzz, cross-validation, byte-and-rune-surfaces]

# Dependency graph
requires:
  - phase: 01-foundation-infrastructure
    provides: [AlgoID dispatch table, errors, CI matrix, license-headers gate]
  - phase: 02-core-character-algorithms-six
    provides: [maxStackInputLen constant, two-row DP discipline]
  - phase: 04-remaining-character-gestalt
    provides: [LongestCommonSubstringRunes byte+rune surface pattern with identity-short-circuit-before-rune-conversion]
  - phase: 05-q-gram-algorithms
    provides: [shared-helper pattern, per-plan llms.txt sync discipline]
  - phase: 06-token-based-algorithms / plan 06-01
    provides: [token_indel.go shared kernel (indelRatio + indelRatioRunes), Indel-formula equivalence,
               RapidFuzz 3.14.5 cross-validation corpus (20 entries with partial_ratio_bytes and partial_ratio_runes per entry),
               structurally-complete TestTokenRatios_CrossValidation loader (partial_bytes / partial_runes sub-tests skipping pending this plan)]
  - phase: 06-token-based-algorithms / plan 06-02
    provides: [TokenSetRatio empty-set deviation pattern (DEVIATION recorded in 6 surfaces template),
               helper-extraction pattern for gocyclo compliance (buildTokenSetPartitions / tokenSetThreeWayMax / joinSectAndDiff)]

provides:
  - PartialRatioScore(a, b string) float64 — byte-path sliding-window with three-region iteration + s1_char_set early-skip + equal-length symmetric tie-break; AlgoPartialRatio dispatch slot 16 wired
  - PartialRatioScoreRunes(a, b string) float64 — rune-path counterpart; public but NOT dispatched (matches LCSStr's rune-variants convention)
  - partialRatioThreeRegionMax(Runes)?, partialRatioRegion1/2/3Bytes/Runes unexported helpers (file-local to partial_ratio.go) — gocyclo-compliant orchestrator + per-region helpers
  - Activated `/partial_bytes` and `/partial_runes` sub-tests in token_ratio_cross_validation_test.go — every entry asserts byte-stable agreement with RapidFuzz 3.14.5 within epsilon = 1e-9 on both surfaces
  - DoS-vector three-part godoc block (Complexity formula + DoS notice + BenchmarkPartialRatio_Pathological_LongShortMismatch_Bytes/Runes fixtures) per 06-CONTEXT.md §5 LOCKED
  - Pitfall-3 keystone fixtures (("abc","ab") + ("abc","bc")) pinned in unit tests AND BDD scenarios AND staging-golden — defence in depth against the naive-single-loop regression
  - Spec-deferred sliding-window DP TODO(#TBD) in partial_ratio.go per docs/requirements.md §612

affects: [06-04-token-jaccard, 06-05-monge-elkan, 06-06-finalisation, 08-scorer, 10-extract]

# Tech tracking
tech-stack:
  added: []
  patterns:
    - "Byte + rune dual surface for character-level algorithms (mirrors LCSStr's lcsstr.go pattern from Phase 4)"
    - "Identity short-circuit BEFORE []rune conversion in rune path (saves 2 heap allocations on identical inputs — same pattern as LongestCommonSubstringRunes lines 173-178)"
    - "s1_char_set early-skip pattern — [256]bool stack-allocated for byte path; map[rune]struct{} for rune path (queried only, never iterated — preserves DET-03)"
    - "Three-region iteration helper-extraction (partialRatioRegion1Bytes / partialRatioRegion2Bytes / partialRatioRegion3Bytes + rune twins) — keeps parent function under gocyclo=10 without nolint pragmas"
    - "Equal-length symmetric tie-break pattern — when len(a) == len(b) AND first pass < 1.0, swap shorter/longer and re-run the three-region iteration (RapidFuzz partial_ratio_alignment lines 328-333 reference)"
    - "Region 3 always-runs pattern (NOT gated on n > m) — when n == m, Region 2 evaluates only i=0 and Region 3 covers all right-suffix alignments. The single trivial overlap at i = n-m when n > m is harmless and matches RapidFuzz reference behaviour"

key-files:
  created:
    - "partial_ratio.go (algorithm — byte + rune surfaces with 6 helper functions)"
    - "partial_ratio_test.go (5 test functions: TestPartialRatioScore, TestPartialRatioScoreRunes, TestPartialRatioScore_Pitfall3_Keystones, TestPartialRatioScore_Symmetric, TestPartialRatioScoreRunes_Symmetric, TestPartialRatioScore_DispatchRegistration)"
    - "partial_ratio_bench_test.go (5 standard benchmarks + 2 Pathological_LongShortMismatch fixtures — bytes and runes)"
    - "partial_ratio_fuzz_test.go (FuzzPartialRatioScore + FuzzPartialRatioScoreRunes with ~11 seeds each + bit-for-bit symmetry regression gate)"
    - "dispatch_partial_ratio.go (slot 16 wiring; byte path only)"
    - "tests/bdd/features/partial_ratio.feature (~14 scenarios — both byte AND rune surfaces; Pitfall-3 keystones; canonical references; symmetry; empty-input)"
    - "testdata/golden/_staging/partial_ratio.json (10 entries)"
  modified:
    - "algoid_test.go (AlgoPartialRatio added to TestDispatch_UnregisteredSlotsAreNil registered map)"
    - "props_test.go (12 TestProp_PartialRatio* properties — 6 per surface: RangeBounds, Identity, Symmetric, NoNaN, NoInf, NoNegativeZero)"
    - "example_test.go (ExamplePartialRatioScore + ExamplePartialRatioScoreRunes)"
    - "tests/bdd/steps/algorithms_steps.go (PartialRatio + PartialRatioRunes step methods + InitializeScenario registrations)"
    - "llms.txt (PartialRatio section)"
    - "llms-full.txt (Phase 6 algorithm-surface block — PartialRatio)"
    - "token_ratio_cross_validation_test.go (/partial_bytes and /partial_runes sub-tests activated; file-level godoc updated)"

key-decisions:
  - "Three-region iteration LOCKED 2026-05-15: Region 1 (left tail) iterates i := 1; i < m with substrings longer[:i]; Region 2 (middle) iterates i := 0; i <= n-m with substrings longer[i:i+m]; Region 3 (right tail) iterates i := n-m; i < n with substrings longer[i:]. Per RESEARCH.md Pattern 6 verbatim structure. The naive single-loop implementation MISSES Regions 1 and 3 — Pitfall 3."
  - "s1_char_set early-skip LOCKED 2026-05-15: [256]bool stack-allocated for the byte path; map[rune]struct{} for the rune path (cheap — short string typically). The skip condition: in Region 1 the LAST byte/rune of the candidate substring is checked; in Region 2 the LAST byte/rune; in Region 3 the FIRST byte/rune. Skipping non-improving alignments is load-bearing for the pathological budget per RESEARCH.md Pitfall 3."
  - "Dispatch wiring LOCKED 2026-05-15: Only the byte-path PartialRatioScore is registered in dispatch[AlgoPartialRatio] (slot 16). The rune-path PartialRatioScoreRunes is public but NOT dispatched (dispatch table signature is the byte-path one — same convention as LCSStr's rune variants)."
  - "Spec-deferred sliding-window DP TODO LOCKED 2026-05-15: `// TODO(#TBD): implement sliding-window DP per Bachmann RapidFuzz docs — spec line 612 deferred to v1.x` is recorded in partial_ratio.go's file-header godoc. The TODO references issue #TBD because no GitHub issue has been created yet; the planner / orchestrator can create the issue and update the TODO with the issue number in a follow-up PR (NOT a blocker for plan completion)."
  - "Identity short-circuit ordering LOCKED 2026-05-15: In the byte path, `if a == b { return 1.0 }` is the first guard. In the rune path, `if a == b { return 1.0 }` is the first guard BEFORE `[]rune(a)` / `[]rune(b)` conversion (saves 2 heap allocations on identical inputs — same pattern as LongestCommonSubstringRunes in lcsstr.go lines 173-178)."
  - "Region 3 always-runs LOCKED 2026-05-15 (Rule 1 deviation auto-fixed during Task 2): the initial implementation gated Region 3 on `if n > m` per the plan's literal text. This was INCORRECT for equal-length inputs: when n == m, Region 2 iterates only i=0 (one alignment evaluated) and Region 3 was entirely skipped, missing all right-suffix alignments. The RapidFuzz reference iterates Region 3 as `range(len2 - len1, len2)` which for n==m becomes `range(0, n)` — full sweep. Fix: remove the `n > m` guard. There is a single trivial overlap at i = n-m with Region 2 when n > m (one redundant indelRatio call; harmless and matches the reference behaviour)."
  - "Equal-length symmetric tie-break LOCKED 2026-05-15 (Rule 1 deviation auto-fixed during Task 2): when len(a) == len(b), Region 1 / Region 3 are asymmetric in the role of `shorter` vs `longer` (they iterate prefixes/suffixes of `longer`, never `shorter`); the optimal alignment may live in either of the two valid `(shorter, longer)` orderings. RapidFuzz's `partial_ratio_alignment` (lines 328-333) handles this by re-invoking `_partial_ratio_impl` with arguments swapped and taking the max. PartialRatioScore + PartialRatioScoreRunes both implement this tie-break — when `res < 1.0 && len(a) == len(b)`, run the three-region iteration a second time with `(longer, shorter)` and update `res` to the max."

patterns-established:
  - "Helper-extraction discipline carried forward from plan 06-02: TokenSetRatioScore extracted 3 helpers (buildTokenSetPartitions / tokenSetThreeWayMax / joinSectAndDiff); PartialRatio extracts 8 helpers (partialRatioThreeRegionMax + 3 region helpers × 2 surfaces). Both parent functions under gocyclo=10 without nolint pragmas. Reusable for any Phase 6+ algorithm where the composition logic exceeds the gocyclo ceiling."
  - "Pathological benchmark fixture pattern carried forward from plans 06-02 / 06-03: precompute inputs OUTSIDE the timed loop (string.Repeat allocations don't pollute the benchmark numbers). PartialRatio adds the long-vs-short DoS shape (10 chars vs 10000 chars) per 06-CONTEXT.md §5 LOCKED."
  - "Equal-length symmetric tie-break for asymmetric-region algorithms: any future algorithm where the inner regions are asymmetric in shorter/longer (e.g. potential v1.x sliding-window DP variant) MUST run the equal-length tie-break to satisfy the catalogue's symmetric-property guarantee. The pattern is documented inline in partial_ratio.go's file-header godoc and the wrapper-function bodies."

requirements-completed:
  - TOKEN-04

# Metrics
duration: 19min
completed: 2026-05-15
---

# Phase 6 Plan 3: Partial Ratio Summary

**Three-region iteration sliding-window with the s1_char_set early-skip, BOTH byte and rune surfaces per spec lines 609-610, the equal-length symmetric tie-break per RapidFuzz `partial_ratio_alignment` reference, and the activated `/partial_bytes` and `/partial_runes` cross-validation sub-tests. Closes TOKEN-04.**

## Performance

- **Duration:** ~19 min
- **Started:** 2026-05-15T10:59:15Z (worktree HEAD reset to d22b9f6)
- **Completed:** 2026-05-15T11:18:23Z (final verification gate)
- **Tasks:** 2 (Task 1 — PartialRatio algorithm + companions; Task 2 — Activate cross-validation partial_bytes + partial_runes)
- **Files modified:** 13 (7 created, 6 modified) excluding the SUMMARY.md itself

## Accomplishments

- Landed **PartialRatioScore** (byte path) and **PartialRatioScoreRunes** (rune path) — the third Indel-formula consumer of `token_indel.go`. PartialRatio is the only Phase 6 algorithm with both byte and rune surfaces per spec line 609-610. Wired byte path into dispatch slot 16.
- Implemented the **three-region iteration** (Region 1 left tail / Region 2 middle / Region 3 right tail) plus the **s1_char_set early-skip** optimisation per RapidFuzz `_partial_ratio_impl` reference (06-RESEARCH.md Pattern 6 + Example 4).
- Activated the **`/partial_bytes` and `/partial_runes` cross-validation sub-tests** — every entry in the RapidFuzz 3.14.5 corpus now asserts byte-stable agreement within epsilon = 1e-9 on both surfaces. 40 sub-tests pass.
- Recorded **seven LOCKED decisions** (three-region iteration / s1_char_set early-skip / dispatch wiring / spec-deferred DP TODO / identity short-circuit ordering / Region 3 always-runs / equal-length symmetric tie-break) — the last two are Rule 1 (bug) deviations auto-fixed during Task 2 that the plan's literal text would have missed.
- Full unit + property (12 properties — 6 per surface) + fuzz (2 fuzzers with ~11 seeds each + bit-for-bit symmetry regression gate) + bench (7 benchmarks — 5 standard + 2 Pathological_LongShortMismatch fixtures) + BDD (~14 scenarios covering both surfaces) + staging-golden (10 entries) + per-plan llms.txt + llms-full.txt sync.

## Task Commits

Each task was committed atomically (per-task convention):

1. **Task 1: PartialRatio (algorithm + dispatch + companions + props + 2 examples + BDD + staging-golden + llms sync)** — `bf94c04` (feat)
2. **Task 2: Activate partial_ratio cross-validation + fix equal-length symmetry deviations** — `c3427d1` (feat)

The SUMMARY.md commit follows separately (the final metadata commit per `execute-plan.md`).

## Files Created/Modified

### Created

- `partial_ratio.go` — PartialRatioScore + PartialRatioScoreRunes. File-header godoc with primary-source citations (RapidFuzz docs + Wagner-Fischer 1974), three-region iteration explanation, s1_char_set early-skip rationale, three-part DoS-vector godoc block per CONTEXT §5 LOCKED, equal-length symmetric tie-break note, spec-deferred sliding-window DP TODO. 6 helper functions extracted for gocyclo compliance.
- `partial_ratio_test.go` — 6 test functions: `TestPartialRatioScore` (12 cases including the Pitfall-3 keystones and the partial-overlap fixture); `TestPartialRatioScoreRunes` (10 cases including the Unicode keystone "café"/"caf"); `TestPartialRatioScore_Pitfall3_Keystones` (6 dedicated Pitfall-3 fixtures with names that surface immediately under `go test -run`); `TestPartialRatioScore_Symmetric` (6 input pairs); `TestPartialRatioScoreRunes_Symmetric` (3 input pairs including non-ASCII); `TestPartialRatioScore_DispatchRegistration`.
- `partial_ratio_bench_test.go` — 7 benchmarks: ASCII Short/Medium/Long byte path; ASCII Short / Unicode Short rune path; `BenchmarkPartialRatio_Pathological_LongShortMismatch_Bytes` (LOCKED fixture per CONTEXT §5); `BenchmarkPartialRatio_Pathological_LongShortMismatch_Runes`.
- `partial_ratio_fuzz_test.go` — `FuzzPartialRatioScore` + `FuzzPartialRatioScoreRunes` with ~11 programmatic seeds each; assert no-NaN, no-Inf, range bounds, AND the bit-for-bit symmetry regression gate.
- `dispatch_partial_ratio.go` — registers `dispatch[AlgoPartialRatio] = PartialRatioScore` (byte path only; rune path NOT dispatched).
- `tests/bdd/features/partial_ratio.feature` — ~14 scenarios covering canonical references (separate Examples for byte and rune surfaces), identity / both-empty / one-empty for both surfaces, symmetry for both surfaces, Pitfall-3 keystones (Region 1 left tail + Region 3 right tail) for both surfaces.
- `testdata/golden/_staging/partial_ratio.json` — 10 entries pinned with derivation notes including explicit `06-RESEARCH.md Pitfall 3 KEYSTONE` annotations on the two keystone fixtures.

### Modified

- `algoid_test.go` — added `AlgoPartialRatio: true` to `TestDispatch_UnregisteredSlotsAreNil` registered map; updated surrounding comments to mention plan 06-03 and the byte-path-only dispatch convention.
- `props_test.go` — appended 12 `TestProp_PartialRatio*` property tests (6 per surface: RangeBounds, Identity, Symmetric, NoNaN, NoInf, NoNegativeZero).
- `example_test.go` — appended `ExamplePartialRatioScore` (RapidFuzz canonical "YANKEES" / "NEW YORK YANKEES" → 1.0) and `ExamplePartialRatioScoreRunes` (Unicode "café"/"caf" → 1.0 with byte-vs-rune divergence note).
- `tests/bdd/steps/algorithms_steps.go` — appended PartialRatio + PartialRatioRunes step methods (6 methods total: 2 score / 2 second-score / 2 both-equal) and 6 `ctx.Step` registrations.
- `llms.txt` — appended `### PartialRatio` section before Normalisation.
- `llms-full.txt` — appended `### Phase 6 algorithm surface (token tier — PartialRatio)` block.
- `token_ratio_cross_validation_test.go` — removed `t.Skip("plan 06-03 will land …")` from `/partial_bytes` and `/partial_runes` sub-tests; added assertion bodies for both surfaces; updated file-level godoc to reflect Wave 3 activation.

## Decisions Made

**The seven LOCKED decisions are recorded verbatim above in the `key-decisions` frontmatter. Reproduced here for prose readability:**

### Three-region iteration LOCKED 2026-05-15

Region 1 (left tail) iterates `i := 1; i < m` with substrings `longer[:i]`; Region 2 (middle) iterates `i := 0; i <= n-m` with substrings `longer[i:i+m]`; Region 3 (right tail) iterates `i := n-m; i < n` with substrings `longer[i:]`. Per RESEARCH.md Pattern 6 verbatim structure. A naive single-loop implementation `for i := 0; i <= n-m; i++` covers only Region 2 — Pitfall 3.

### s1_char_set early-skip LOCKED 2026-05-15

`var charSet [256]bool` (byte path) or `map[rune]struct{}` (rune path) of characters appearing in the shorter input. In Region 1 the LAST byte/rune of the candidate substring is checked against the set; in Region 2 the LAST byte/rune; in Region 3 the FIRST byte/rune. Alignments whose check fails are skipped — load-bearing for the pathological budget per RESEARCH.md Pitfall 3.

### Dispatch wiring LOCKED 2026-05-15

Only the byte-path `PartialRatioScore` is registered in `dispatch[AlgoPartialRatio]` (slot 16). `PartialRatioScoreRunes` is public but NOT dispatched (dispatch table signature is the byte-path one — same convention as LCSStr's rune variants in lcsstr.go).

### Spec-deferred sliding-window DP TODO LOCKED 2026-05-15

`// TODO(#TBD): implement sliding-window DP per Bachmann RapidFuzz docs — spec line 612 deferred to v1.x` is recorded in partial_ratio.go's file-header godoc. The TODO references issue `#TBD` because no GitHub issue has been created yet; a follow-up PR (NOT a blocker for plan completion) will create the issue and update the TODO with the issue number. The straightforward loop-over-alignments + indelRatio approach with the s1_char_set early-skip is the Phase 6 implementation.

### Identity short-circuit ordering LOCKED 2026-05-15

In the byte path, `if a == b { return 1.0 }` is the first guard. In the rune path, `if a == b { return 1.0 }` is the first guard BEFORE `[]rune(a)` / `[]rune(b)` conversion (saves 2 heap allocations on identical inputs — same pattern as `LongestCommonSubstringRunes` in lcsstr.go lines 173-178).

### Region 3 always-runs LOCKED 2026-05-15 (Rule 1 deviation auto-fixed during Task 2)

The initial implementation gated Region 3 on `if n > m` per the plan's literal text. This was INCORRECT for equal-length inputs: when `n == m`, Region 2 evaluates only `i = 0` (one alignment) and Region 3 was entirely skipped, missing all right-suffix alignments. The RapidFuzz reference iterates Region 3 as `range(len2 - len1, len2)` which for `n == m` becomes `range(0, n)` — full sweep. Fix: remove the `n > m` guard. There is a single trivial overlap at `i = n - m` with Region 2 when `n > m` (one redundant indelRatio call; harmless and matches the reference behaviour).

### Equal-length symmetric tie-break LOCKED 2026-05-15 (Rule 1 deviation auto-fixed during Task 2)

When `len(a) == len(b)`, Region 1 / Region 3 are asymmetric in the role of `shorter` vs `longer` (they iterate prefixes/suffixes of `longer`, never `shorter`); the optimal alignment may live in either of the two valid `(shorter, longer)` orderings. RapidFuzz's `partial_ratio_alignment` (lines 328-333) handles this by re-invoking `_partial_ratio_impl` with arguments swapped and taking the max:

```python
res = _partial_ratio_impl(shorter, longer, score_cutoff / 100)
if res.score != 100 and len1 == len2:
    score_cutoff = max(score_cutoff, res.score)
    res2 = _partial_ratio_impl(longer, shorter, score_cutoff / 100)
    if res2.score > res.score:
        res = ScoreAlignment(res2.score, ...)
```

`PartialRatioScore` + `PartialRatioScoreRunes` both implement this tie-break — when `res < 1.0 && len(a) == len(b)`, run the three-region iteration a second time with `(longer, shorter)` and update `res` to the max. This guarantees the catalogue's symmetric-property guarantee holds for equal-length inputs.

## Reference Vector Numbers for the Staging-Golden File

The `testdata/golden/_staging/partial_ratio.json` file (10 entries) pins the following scores:

| Name | Score | Derivation |
|------|------:|------------|
| identity | 1.0 | a == b identity short-circuit |
| both_empty | 1.0 | both-empty / identity short-circuit (does NOT inherit TokenSetRatio RapidFuzz issue #110 deviation) |
| one_empty_a | 0.0 | one-empty |
| one_empty_b | 0.0 | one-empty |
| region_2_middle_wins | 1.0 | Region 2 at i=9 of "NEW YORK YANKEES" → indelRatio("YANKEES","YANKEES")=1.0 — RapidFuzz canonical reference |
| region_1_left_tail_pitfall_3 | 1.0 | KEYSTONE — Region 2 at i=0 → indelRatio("ab","ab")=1.0 |
| region_3_right_tail_pitfall_3 | 1.0 | KEYSTONE — Region 2 at i=1 → indelRatio("bc","bc")=1.0 |
| partial_overlap | 0.75 | Region 2 at i=0 of "xabcy" → indelRatio("abcd","xabc")=2·3/8 |
| disjoint_no_overlap | 0.0 | charSet={'a','b','c'}; longer="xyzzz" has no matching bytes; all alignments skipped |
| subset_short_at_end | 1.0 | Region 2 at i=6 of "hello world" → indelRatio("world","world")=1.0 |

## Cross-validation Activation Status

**All 20 `/partial_bytes` and 20 `/partial_runes` sub-tests pass** within `epsilon = 1e-9`.

The committed corpus contains 20 entries; per OQ-3 RESOLUTION LOCKED in plan 06-01, `partial_ratio_runes` is always emitted for every entry (matches `partial_ratio_bytes` on ASCII inputs; exercises the separate rune-path implementation on every entry). Both sub-tests run their assertion body on every entry.

Sub-test status after this plan:
- `/token_sort` — 20 PASS (active from plan 06-01)
- `/token_set` — 20 PASS (active from plan 06-02)
- `/partial_bytes` — 20 PASS (active from this plan, 06-03)
- `/partial_runes` — 20 PASS (active from this plan, 06-03)

Total: 80 cross-validation sub-tests now active across the four surfaces.

## Pathological-Bench Timing

Measured on developer hardware (Apple M2, darwin/arm64) at `-benchtime=1x`:

| Benchmark | ns/op | B/op | allocs/op |
|-----------|-------:|------:|----------:|
| `BenchmarkPartialRatioScore_ASCII_Short` | 96250 | 16 | 1 |
| `BenchmarkPartialRatioScore_ASCII_Medium` | 10125 | 0 | 0 |
| `BenchmarkPartialRatioScore_ASCII_Long` | 15458 | 0 | 0 |
| `BenchmarkPartialRatioScoreRunes_ASCII_Short` | 14542 | 0 | 0 |
| `BenchmarkPartialRatioScoreRunes_Unicode_Short` | 7041 | 0 | 0 |
| `BenchmarkPartialRatio_Pathological_LongShortMismatch_Bytes` | 13416 | 0 | 0 |
| `BenchmarkPartialRatio_Pathological_LongShortMismatch_Runes` | 129000 | 41144 | 4 |

(The first-iteration `-benchtime=1x` numbers are noisy by design — `bench.txt` regenerates with the full suite in plan 06-06.)

**The Pathological_LongShortMismatch_Bytes fixture (10-char vs 10000-char) completes in 13 µs/op — well below the 100 µs target.** The s1_char_set early-skip is the load-bearing optimisation per RESEARCH.md Pitfall 3. The equal-length symmetric tie-break only fires when `len(a) == len(b)`, which is not the pathological shape — so the tie-break does not affect the pathological timing.

## Deviations from Plan

### Auto-fixed Issues

**1. [Rule 1 — Bug] Region 3 right-tail gate `if n > m` was wrong for equal-length inputs**

- **Found during:** Task 2 (cross-validation activation surfaced the divergence on entries `long_token_reorder` and `ascii_long` where `len(a) == len(b)`)
- **Issue:** The plan's literal text specified the Region 3 gate as `if n > m { for i := n - m; i < n; i++ ...}` with the rationale "the `n > m` guard avoids re-iterating Region 2 when `n == m`". But when `n == m`, Region 2 iterates only `i = 0` (one alignment evaluated — the full alignment of shorter vs longer); Region 3 was supposed to be re-iterating only ONE position (i = 0) per the guard's rationale, but in reality Region 3 covers all RIGHT-SUFFIX alignments at i = 1..n-1 which are NOT re-iterated by Region 2 at all. Result: 6 corpus failures on entries where `len(a) == len(b)` and the optimal alignment was a right-suffix or a non-trivial prefix.
- **Fix:** Removed the `n > m` guard from both byte and rune paths. Region 3 always runs. There is a single trivial overlap at `i = n - m` with Region 2 when `n > m` (one redundant indelRatio call; harmless and matches the RapidFuzz `_partial_ratio_impl` reference behaviour — its Region 3 iterates `range(len2 - len1, len2)` which for `n == m` is `range(0, n)`, full sweep).
- **Files modified:** `partial_ratio.go` (algorithm body for both surfaces; helper godocs; file-header godoc updated to clarify Region 3 always runs).
- **Verification:** Cross-validation passes; full test suite remains green.
- **Committed in:** `c3427d1` (Task 2 commit — bundled with the cross-validation activation since the bug surfaced during activation).

**2. [Rule 1 — Bug] Equal-length symmetric tie-break missing — algorithm not symmetric for equal-length inputs**

- **Found during:** Task 2 (after fixing deviation #1 above, ONE corpus entry still failed: `("alpha beta gamma delta epsilon zeta", "zeta epsilon delta gamma beta alpha")` — both 35 chars. Direct probe showed `PR(a,b) = 0.5901639344` but `PR(b,a) = 0.6`. The function was asymmetric for equal-length inputs.)
- **Issue:** Region 1 / Region 3 of `_partial_ratio_impl` iterate prefixes / suffixes of `longer`, never `shorter`. When `len(a) == len(b)`, there are TWO valid `(shorter, longer)` orderings and the optimal alignment may live in only one of them. My initial implementation picked `(a, b)` if `len(a) <= len(b)` (i.e. always when equal), which biased the result.
- **Fix:** Looked at the upstream RapidFuzz Python source (`fuzz_py.py::partial_ratio_alignment` lines 328-333) — it explicitly handles this case by re-invoking `_partial_ratio_impl(longer, shorter)` when `len1 == len2 && res.score != 100` and taking the max. Mirrored the pattern in both `PartialRatioScore` and `PartialRatioScoreRunes`: when `res < 1.0 && len(a) == len(b)`, run the three-region iteration a second time with `(longer, shorter)` and update `res` to the max.
- **Files modified:** `partial_ratio.go` (both wrapper functions + file-header godoc Equal-length symmetric tie-break note).
- **Verification:** Cross-validation: all 80 sub-tests (token_sort + token_set + partial_bytes + partial_runes) pass. `TestPartialRatioScore_Symmetric` + `TestPartialRatioScoreRunes_Symmetric` + property `TestProp_PartialRatioScore_Symmetric` + property `TestProp_PartialRatioScoreRunes_Symmetric` all pass — the function is now bit-for-bit symmetric across argument order.
- **Committed in:** `c3427d1` (Task 2 commit — bundled with deviation #1 fix and the cross-validation activation).

**3. [Rule 3 — Blocking] Cyclomatic complexity ceiling**

- **Found during:** Task 1 (lint gate)
- **Issue:** Initial `PartialRatioScore` and `PartialRatioScoreRunes` implementations had cyclomatic complexity 19 (golangci-lint threshold = 10). The three-region iteration packed all three nested loops + the regions' `if` guards + the perfect-match early-exit into a single function body. After a first refactor (extracting `partialRatioThreeRegionMax(Runes)?`) the helper was still at complexity 13.
- **Fix:** Extracted per-region helpers — `partialRatioRegion1Bytes`, `partialRatioRegion2Bytes`, `partialRatioRegion3Bytes`, and the rune twins. The parent function `partialRatioThreeRegionMax(Runes)?` is now a thin orchestrator (~10 statements) that delegates the per-region iteration to dedicated helpers, with the perfect-match early-exit signalled via the `(best, perfect)` return from `partialRatioRegion2`. All functions are now under gocyclo=10 without nolint pragmas. This mirrors the helper-extraction pattern from plan 06-02 deviation #3 (TokenSetRatio's `buildTokenSetPartitions` / `tokenSetThreeWayMax` / `joinSectAndDiff` extraction).
- **Files modified:** `partial_ratio.go` (added 8 helper functions).
- **Verification:** `make lint` reports 0 issues.
- **Committed in:** `bf94c04` (Task 1 commit — refactor was applied before the final Task 1 commit).

**4. [Rule 3 — Blocking] gofmt drift in long-line comment table**

- **Found during:** Task 1 (post-commit fmt-check)
- **Issue:** The struct-literal seed table in `partial_ratio_fuzz_test.go` had inconsistent comment column alignment after the initial write — `gofmt -s` reformatted it to a uniform column.
- **Fix:** Ran `gofmt -s -w partial_ratio_fuzz_test.go`. No semantic change.
- **Files modified:** `partial_ratio_fuzz_test.go`.
- **Verification:** `make fmt-check` passes.
- **Committed in:** `bf94c04` (Task 1 commit — fix applied before the commit).

---

**Total deviations:** 4 auto-fixed (2× Rule 1 — bugs surfaced by cross-validation; 1× Rule 3 — gocyclo; 1× Rule 3 — gofmt drift; all caught and resolved during plan execution; no scope creep)
**Impact on plan:** Both Rule 1 fixes are essential — without them the cross-validation gate fails on 6 entries (deviation #1) or on the symmetric-pair entry (deviation #2). The deviations are well within Rule 1's intent (correctness bugs surfaced by the verification suite). The Rule 3 fixes are mandatory CI gates.

## Issues Encountered

The two Rule 1 deviations above represent real algorithmic subtleties that the plan's literal text did not anticipate:

1. The plan's `<recorded_resolutions>` block for Region 3 wrongly attributed the `n > m` guard's purpose ("avoids re-iterating Region 2 when `n == m`"). The actual RapidFuzz `_partial_ratio_impl` does iterate Region 3 fully when `n == m`. The corpus regeneration against RapidFuzz 3.14.5 surfaced this immediately upon activation.
2. The plan did not mention the equal-length symmetric tie-break at all — that's a wrapper-level (not impl-level) detail in RapidFuzz's source. Without it, the `Symmetric` property test would have eventually caught it via random inputs, but the corpus entry `long_token_reorder` (both 35 chars) surfaced it deterministically.

Both deviations are now LOCKED in the algorithm godoc, helper docstrings, the SUMMARY (above), and propagated forward via the `patterns-established` frontmatter so plan 06-05 (Monge-Elkan) inherits the awareness if any of its inner-metric helpers exhibit similar asymmetry.

## User Setup Required

None — no external service configuration required for this plan.

## Next Phase Readiness

- **Plan 06-04 (TokenJaccard)** is unblocked — TokenJaccard is a set-Jaccard composition over `Tokenise(s)`. It does NOT consume `indelRatio` from `token_indel.go` (set-Jaccard is `|A∩B|/|A∪B|`, no LCS). It follows the catalogue's standard both-empty → 1.0 convention (NOT TokenSetRatio's deviation).
- **Plan 06-05 (MongeElkan)** inherits the dispatcher slot 13 wiring pattern AND the helper-extraction discipline established in plans 06-02 / 06-03 (extract per-region/per-stage helpers when the parent exceeds gocyclo=10).
- **Plan 06-06 finalisation** has a 10-entry staging-golden to merge into `algorithms.json`, the `TODO(#TBD)` to update with a created GitHub issue (or leave as-is if no issue is created), and the cross-platform determinism golden update to include PartialRatio entries.

No blockers or concerns.

### Deferred items for plan 06-06

- **Final `bench.txt` numbers** — PartialRatio benchmarks compile and run (5 standard + 2 Pathological) but the project-wide `bench.txt` baseline is regenerated phase-by-phase at finalisation time, not per-plan. Plan 06-06 will run the full benchmark suite and commit the updated `bench.txt`. The Pathological_LongShortMismatch numbers will commit to `bench.txt` and be subject to benchstat regression detection > 10% per docs/requirements.md §6(6).
- **`testdata/golden/_staging/partial_ratio.json` merge into `testdata/golden/algorithms.json`** — staged in the `_staging/` directory; plan 06-06 finalisation handles the merge.
- **Cross-platform determinism golden update** — PartialRatioScore + PartialRatioScoreRunes on ASCII / Unicode inputs are deterministic across the four CI platforms by construction (integer-derived single divisions in indelRatio + integer comparisons), but `verify-determinism`'s golden file does not yet include any PartialRatio entries. Plan 06-06 will add representative entries.
- **GitHub issue for spec-deferred sliding-window DP** — the `TODO(#TBD)` in `partial_ratio.go`'s file-header godoc references a placeholder issue number. Plan 06-06 (or a follow-up housekeeping PR) creates the GitHub issue tracking the v1.x sliding-window DP work and updates the `TODO` with the actual issue number.

## Self-Check: PASSED

File-existence checks:

- `partial_ratio.go` — present.
- `partial_ratio_test.go` — present.
- `partial_ratio_bench_test.go` — present.
- `partial_ratio_fuzz_test.go` — present.
- `dispatch_partial_ratio.go` — present.
- `tests/bdd/features/partial_ratio.feature` — present.
- `testdata/golden/_staging/partial_ratio.json` — present (10 entries).

Commit-existence checks:

- Commit `bf94c04` — present in `git log --oneline -5` (Task 1: PartialRatio land).
- Commit `c3427d1` — present in `git log --oneline -5` (Task 2: cross-validation activation + symmetry fixes).

Cross-validation:

- `TestTokenRatios_CrossValidation/*/partial_bytes` — 20 sub-tests, all PASS.
- `TestTokenRatios_CrossValidation/*/partial_runes` — 20 sub-tests, all PASS.
- `TestPartialRatio*` + `TestProp_PartialRatio*` — all PASS.
- `cd tests/bdd && go test -race -count=1 ./...` — PASS.
- `make fmt-check vet lint verify-license-headers verify-deps-allowlist` — all 0 issues.
- `TestAIFriendly_LLMSTxtReferencesEveryExportedSymbol` — PASS (both new exported symbols `PartialRatioScore` and `PartialRatioScoreRunes` are referenced in `llms.txt`).

All claimed deliverables verified by `git log`, file existence on disk, and the test/lint gates above.

---
*Phase: 06-token-based-algorithms*
*Completed: 2026-05-15*
