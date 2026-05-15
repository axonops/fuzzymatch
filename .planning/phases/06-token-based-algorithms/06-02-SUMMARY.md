---
phase: 06-token-based-algorithms
plan: 02
subsystem: algorithm-catalogue
tags: [token-set-ratio, indel-formula, three-way-max, rapidfuzz-issue-110, deviation, cross-validation]

# Dependency graph
requires:
  - phase: 01-foundation-infrastructure
    provides: [AlgoID dispatch table, Tokenise, errors, CI matrix, license-headers gate]
  - phase: 02-core-character-algorithms-six
    provides: [maxStackInputLen constant, two-row DP discipline]
  - phase: 04-remaining-character-gestalt
    provides: [cross-validation script template]
  - phase: 05-q-gram-algorithms
    provides: [shared-helper pattern, per-plan llms.txt sync discipline]
  - phase: 06-token-based-algorithms / plan 06-01
    provides: [token_indel.go shared kernel (lcsLen / indelRatio), Indel-formula equivalence,
               RapidFuzz 3.14.5 cross-validation corpus (20 entries),
               TokenSortRatio precedent (file-header layout, OQ-1 RESOLUTION),
               structurally-complete TestTokenRatios_CrossValidation loader (token_set sub-test skipping pending this plan)]

provides:
  - TokenSetRatioScore(a, b string) float64 — three-way Indel max with bug-for-bug RapidFuzz issue #110 deviation; AlgoTokenSetRatio dispatch slot 15 wired
  - buildTokenSetPartitions / tokenSetThreeWayMax / joinSectAndDiff unexported helpers (file-local to token_set_ratio.go)
  - Activated `/token_set` sub-test in token_ratio_cross_validation_test.go — every non-partial_only entry asserts byte-stable agreement with RapidFuzz 3.14.5 within epsilon = 1e-9
  - DoS-vector three-part godoc block (Complexity formula + DoS notice + BenchmarkTokenSetRatio_Pathological_AsymmetricSetCardinalities fixture) per 06-CONTEXT.md §5 LOCKED
  - LOCKED DEVIATION (RapidFuzz issue #110) — empty-input gate fires BEFORE identity short-circuit; documented in 6 surfaces (godoc / BDD / staging-golden / cross-validation corpus / property-test docstring / SUMMARY)

affects: [06-03-partial-ratio, 06-04-token-jaccard, 06-05-monge-elkan, 06-06-finalisation, 08-scorer, 10-extract]

# Tech tracking
tech-stack:
  added: []
  patterns:
    - "Empty-input gate fires BEFORE identity short-circuit (gate ordering matters for bug-for-bug parity with upstream)"
    - "Three-way max via helper function (tokenSetThreeWayMax) — keeps parent function under gocyclo ceiling and isolates the algorithmic kernel for reviewer auditing"
    - "Set-partition helper (buildTokenSetPartitions) that iterates INPUT slices (not membership-test maps) on output paths — satisfies DET-03 even though the membership maps exist internally"

key-files:
  created:
    - "token_set_ratio.go (algorithm)"
    - "token_set_ratio_test.go (4 unit-test functions: TestTokenSetRatioScore, TestTokenSetRatioScore_EmptyDeviationDocumented, TestTokenSetRatioScore_SubsetShortCircuit, TestTokenSetRatioScore_Symmetric, TestTokenSetRatioScore_DispatchRegistration)"
    - "token_set_ratio_bench_test.go (4 standard benchmarks + BenchmarkTokenSetRatio_Pathological_AsymmetricSetCardinalities)"
    - "token_set_ratio_fuzz_test.go (FuzzTokenSetRatioScore with 12 seeds)"
    - "dispatch_token_set_ratio.go (slot 15 wiring)"
    - "tests/bdd/features/token_set_ratio.feature (7 scenarios incl. both-empty-strings-deviation + pure-separator-deviation)"
    - "testdata/golden/_staging/token_set_ratio.json (12 entries)"
  modified:
    - "algoid_test.go (AlgoTokenSetRatio added to TestDispatch_UnregisteredSlotsAreNil registered map)"
    - "props_test.go (6 TestProp_TokenSetRatioScore_* properties — identity guards on Tokenise-empty input)"
    - "example_test.go (ExampleTokenSetRatioScore — three-way max with combined-vs-combined branch winning)"
    - "tests/bdd/steps/algorithms_steps.go (TokenSetRatio step methods + InitializeScenario registrations)"
    - "llms.txt (TokenSetRatio section)"
    - "llms-full.txt (Phase 6 algorithm-surface block — TokenSetRatio)"
    - "token_ratio_cross_validation_test.go (/token_set sub-test activated; godoc updated)"

key-decisions:
  - "DEVIATION LOCKED 2026-05-15: TokenSetRatioScore returns 0.0 (not 1.0) when EITHER input is empty OR either Tokenise output is empty — bug-for-bug compat with RapidFuzz issue #110 / fuzzywuzzy. Empty-input gate fires BEFORE identity short-circuit so ('','') returns 0.0 matching RapidFuzz exactly. Documented in algorithm godoc, BDD scenarios, staging-golden, cross-validation corpus, property-test docstring, and this SUMMARY."
  - "Three-way max LOCKED 2026-05-15: the three branches are (a) indelRatio(sortedSect, combined1to2), (b) indelRatio(sortedSect, combined2to1), (c) indelRatio(combined1to2, combined2to1) — where combinedXtoY = sortedSect + ' ' + sortedDiffXY. CRITICAL CORRECTION: RESEARCH.md Pattern 5 pseudocode showed branch (a) as indelRatio(diff_ab_joined, diff_ba_joined) but the actual RapidFuzz / fuzzywuzzy reference uses indelRatio(combined1to2, combined2to1) — i.e. the third branch includes the intersection prefix on BOTH sides. Verified bit-for-bit against the committed corpus (tokens_low_overlap = 7/11 confirms the combined-vs-combined branch wins)."
  - "DET-03 sect_len LOCKED 2026-05-15: intersection / diff key slices are built via buildTokenSetPartitions, which iterates the INPUT slices (deterministic order) and uses membership-test maps for O(1) lookups. The OUTPUT slices are sort.Strings'd before any consumption, so the eventual strings.Join is deterministic byte-for-byte. No string is built from set iteration on any output path."
  - "Subset short-circuit LOCKED 2026-05-15: when intersection is non-empty AND (diffAB is empty OR diffBA is empty), the function returns 1.0 directly (RESEARCH.md Pattern 5 critical landmine 2). Per the LOCKED gate ordering, the pre-Tokenise empty-input gate AND identity short-circuit fire FIRST, so the subset short-circuit only fires for non-trivial inputs where one token set is a true proper subset of the other (with at least one shared token)."
  - "Gate ordering LOCKED 2026-05-15: (1) empty-input gate `if a == \"\" || b == \"\" return 0.0`, (2) identity short-circuit `if a == b return 1.0` (non-empty by step 1), (3) Tokenise, (4) post-Tokenise empty-set gate, (5) buildTokenSetPartitions, (6) subset short-circuit, (7) three-way max via tokenSetThreeWayMax. The ordering is load-bearing — moving the identity check before the empty-input gate breaks ('','') agreement with RapidFuzz."

patterns-established:
  - "Cyclomatic-complexity discipline: TokenSetRatioScore is structured as a thin orchestrator that delegates to two helper functions (buildTokenSetPartitions, tokenSetThreeWayMax) so the parent function lives under the gocyclo=10 threshold without nolint pragmas. This pattern is reusable for future algorithms whose composition logic exceeds the threshold."
  - "Pathological benchmark fixture: BenchmarkTokenSetRatio_Pathological_AsymmetricSetCardinalities uses precomputed inputs outside the timed loop (5-token vs 100-token, 2-token shared core). Pattern reusable for PartialRatio (plan 06-03) and MongeElkan (plan 06-05) pathological fixtures."

requirements-completed:
  - TOKEN-03

# Metrics
duration: 22min
completed: 2026-05-15
---

# Phase 6 Plan 2: Token Set Ratio Summary

**Three-way Indel max with bug-for-bug RapidFuzz issue #110 empty-set deviation. TokenSetRatio is the second Indel-formula consumer in Phase 6 (after TokenSortRatio in plan 06-01), and the most algorithmically complex of the three Indel-based ratios.**

## Performance

- **Duration:** ~22 min
- **Started:** 2026-05-15T10:29:52Z (worktree HEAD reset to b2fcf2f)
- **Completed:** 2026-05-15T10:51:58Z (final commit + verification gate)
- **Tasks:** 2 (Task 1 — TokenSetRatio algorithm + companions; Task 2 — Activate cross-validation token_set sub-test)
- **Files modified:** 13 (7 created, 6 modified) excluding the SUMMARY.md itself

## Accomplishments

- Landed **TokenSetRatioScore** — the three-way max construction over (sortedSect, combined1to2), (sortedSect, combined2to1), (combined1to2, combined2to1). Wired into dispatch slot 15. Full unit + property + fuzz + bench + BDD + staging-golden coverage.
- Activated the **`/token_set` cross-validation sub-test** — every non-partial_only entry in the RapidFuzz 3.14.5 corpus now asserts byte-stable agreement within epsilon = 1e-9. 20 sub-tests pass.
- Discovered and corrected **two critical implementation issues during execution** (documented in Deviations below):
  1. **RESEARCH.md Pattern 5 had the wrong branch formula** — the pseudocode showed branch 1 as `indelRatio(diff_ab, diff_ba)` (just the differences), but the actual RapidFuzz / fuzzywuzzy reference uses `indelRatio(combined1to2, combined2to1)` (with intersection prefix). Verified bit-for-bit against the committed corpus.
  2. **Gate ordering deviation** — the initial Task 1 implementation placed the identity short-circuit before the empty-input gate, causing `TokenSetRatioScore("", "")` to return 1.0 instead of the RapidFuzz-required 0.0. Fixed in Task 2's commit by reordering the gates.
- Locked five LOCKED decision records covering the algorithm semantics, deviation, gate ordering, DET-03 satisfaction strategy, and subset short-circuit.

## Task Commits

Each task was committed atomically (per-task convention):

1. **Task 1: TokenSetRatio (algorithm + dispatch + companions + props + example + BDD + staging-golden + llms sync)** — `880c037` (feat)
2. **Task 2: Activate token_set cross-validation + ("","") deviation fix** — `22e25f4` (feat)

The SUMMARY.md commit follows separately (the final metadata commit per `execute-plan.md`).

## Files Created/Modified

### Created

- `token_set_ratio.go` — TokenSetRatioScore. Empty-input gate before identity short-circuit; Tokenise; post-Tokenise empty-set gate; buildTokenSetPartitions; subset short-circuit; tokenSetThreeWayMax. Three-part DoS-vector godoc block per CONTEXT §5 LOCKED.
- `token_set_ratio_test.go` — 5 test functions: `TestTokenSetRatioScore` (13 cases including diff-dominant case and both-empty-strings deviation); `TestTokenSetRatioScore_EmptyDeviationDocumented` (6 pure-separator cases); `TestTokenSetRatioScore_SubsetShortCircuit` (5 subset cases); `TestTokenSetRatioScore_Symmetric` (6 input pairs); `TestTokenSetRatioScore_DispatchRegistration`.
- `token_set_ratio_bench_test.go` — 4 standard benchmarks (ASCII Short/Medium/Long, Unicode Short) + `BenchmarkTokenSetRatio_Pathological_AsymmetricSetCardinalities` (LOCKED fixture per CONTEXT §5).
- `token_set_ratio_fuzz_test.go` — `FuzzTokenSetRatioScore` with 12 programmatic seeds; asserts no-NaN, no-Inf, range bounds, AND the symmetric-pair regression check.
- `dispatch_token_set_ratio.go` — registers `dispatch[AlgoTokenSetRatio] = TokenSetRatioScore` (no closure; signature matches dispatch table).
- `tests/bdd/features/token_set_ratio.feature` — 7 Gherkin scenarios including `both-empty strings return 0.0 (RapidFuzz issue #110 deviation)` and `pure-separator inputs return 0.0`.
- `testdata/golden/_staging/token_set_ratio.json` — 12 entries with explicit `deviation_note` fields on the two deviation-related cases.

### Modified

- `algoid_test.go` — added `AlgoTokenSetRatio: true` to `TestDispatch_UnregisteredSlotsAreNil` registered map.
- `props_test.go` — appended 6 `TestProp_TokenSetRatioScore_*` property tests (RangeBounds, Identity, Symmetric, NoNaN, NoInf, NoNegativeZero). Identity guards on `len(Tokenise(x)) == 0` to skip both the literal-empty and pure-separator deviation paths.
- `example_test.go` — appended `ExampleTokenSetRatioScore` (three-way max case where combined-vs-combined branch wins).
- `tests/bdd/steps/algorithms_steps.go` — appended TokenSetRatio step methods + 3 InitializeScenario registrations.
- `llms.txt` — appended `### TokenSetRatio` section before Normalisation.
- `llms-full.txt` — appended `### Phase 6 algorithm surface (token tier — TokenSetRatio)` block.
- `token_ratio_cross_validation_test.go` — removed `t.Skip("plan 06-02 will land …")` from `/token_set` sub-test; added assertion body; updated file-level godoc to reflect Wave 2 activation.

## Decisions Made

**The five LOCKED decisions are recorded verbatim above in the `key-decisions` frontmatter. Reproduced here for prose readability:**

### DEVIATION LOCKED 2026-05-15

TokenSetRatioScore returns 0.0 (not 1.0) when EITHER input is empty OR either Tokenise output is empty — bug-for-bug compatibility with RapidFuzz issue #110 / fuzzywuzzy. The empty-input gate fires BEFORE the identity short-circuit so `TokenSetRatioScore("", "")` returns 0.0 matching RapidFuzz exactly. This is the **only** algorithm in the catalogue with this deviation; TokenJaccard / MongeElkan follow the standard both-empty → 1.0 convention. The deviation is recorded in **six** surfaces:

1. `token_set_ratio.go` file-header godoc (Algorithm step 1)
2. `token_set_ratio.go` function-level godoc (Conventions block + DEVIATION note)
3. BDD scenarios in `tests/bdd/features/token_set_ratio.feature` (both `both-empty strings return 0.0` AND `pure-separator inputs return 0.0`)
4. `testdata/golden/_staging/token_set_ratio.json` entries with explicit `deviation_note` fields
5. `testdata/cross-validation/token-ratios/vectors.json` (corpus pins `token_set_ratio: 0.0` on the `both_empty` entry)
6. `props_test.go` TestProp_TokenSetRatioScore_Identity docstring (cites RESEARCH.md Pitfall 2)

Reviewers cannot remove the deviation without breaking six independent test surfaces.

### Three-way max LOCKED 2026-05-15

The three branches are:

- **r1** = `indelRatio(sortedSect, combined1to2)` — intersection vs intersection+diff_ab
- **r2** = `indelRatio(sortedSect, combined2to1)` — intersection vs intersection+diff_ba
- **r3** = `indelRatio(combined1to2, combined2to1)` — intersection+diff_ab vs intersection+diff_ba

where:

- `combined1to2 = sortedSect + " " + sortedDiffAB` (or `sortedSect` if `sortedDiffAB == ""`, or `sortedDiffAB` if `sortedSect == ""`)
- `combined2to1 = sortedSect + " " + sortedDiffBA` (same conditional pattern)

**CRITICAL CORRECTION (Rule 1 deviation auto-discovered during Task 1):** the plan referenced RESEARCH.md Pattern 5 pseudocode that showed branch (a) as `indelRatio(diff_ab_joined, diff_ba_joined)` — i.e. just the differences, no intersection prefix. But the actual RapidFuzz / fuzzywuzzy reference uses `indelRatio(combined1to2, combined2to1)` — with the intersection prefix on BOTH sides. Verified bit-for-bit against the committed corpus:

- `tokens_low_overlap` ("hello world" / "world peace") → corpus says `token_set_ratio = 0.6363636363636364 = 7/11`. The "just-the-differences" formula would give max(0.625, 0.625, 0.2) = 0.625, which DIVERGES from the corpus. The "combined-vs-combined" formula gives max(0.625, 0.625, 7/11) = 7/11, which AGREES with the corpus.

The plan's prose (`recorded_resolutions` section) and the implementation file-header godoc both clarify this correctly; only the inline RESEARCH.md pseudocode reference (Pattern 5 lines 314-318) had the wrong branch (a). No further action needed — the implementation is correct; RESEARCH.md is reference documentation that the plan supersedes.

### DET-03 sect_len LOCKED 2026-05-15

Intersection / diff key slices are built via `buildTokenSetPartitions`, which:

1. Builds membership-test maps `setA`, `setB` for O(1) lookups
2. Iterates the **input slices** (`tokensA`, `tokensB`) — deterministic order — classifying each unique token via membership tests + dedup-tracking maps
3. Sorts each result slice via `sort.Strings` BEFORE return

The output `intersectKeys / diffABKeys / diffBAKeys` slices are byte-deterministic across all four CI platforms because:

- The input slices have deterministic order (Tokenise is deterministic)
- The membership maps are queried (not iterated) on output paths
- `sort.Strings` is byte-lex stable
- `strings.Join` (consumed downstream) is order-preserving

No string is ever built from set iteration. DET-03 is satisfied.

### Subset short-circuit LOCKED 2026-05-15

When `len(intersectKeys) > 0` AND (`len(diffABKeys) == 0` OR `len(diffBAKeys) == 0`), the function returns 1.0 directly. This matches RapidFuzz's `if intersect and (not diff_ab or not diff_ba): return 100`. The short-circuit fires for cases like `("alpha beta", "alpha beta gamma")` where B is a strict superset of A.

### Gate ordering LOCKED 2026-05-15

The seven-step pipeline (per algorithm godoc):

1. **Empty-input gate** — `if a == "" || b == "" { return 0.0 }` (LOCKED before identity per the deviation)
2. **Identity short-circuit** — `if a == b { return 1.0 }` (only fires for non-empty by step 1)
3. **Tokenise** — `Tokenise(a, DefaultTokeniseOptions())` × 2
4. **Post-Tokenise empty-set gate** — `if len(tokensA) == 0 || len(tokensB) == 0 { return 0.0 }`
5. **buildTokenSetPartitions** — produces sorted intersection / diffAB / diffBA key slices
6. **Subset short-circuit** — `if len(intersectKeys) > 0 && (len(diffABKeys) == 0 || len(diffBAKeys) == 0) { return 1.0 }`
7. **tokenSetThreeWayMax** — computes r1, r2, r3 and returns the max

Moving any gate before its predecessor breaks at least one corpus entry. The ordering is load-bearing.

## Reference Vector Numbers for the Staging-Golden File

The `testdata/golden/_staging/token_set_ratio.json` file (12 entries) pins the following scores:

| Name | Score | Derivation |
|------|------:|------------|
| identity | 1.0 | a == b non-empty identity short-circuit |
| both_empty_strings_deviation | 0.0 | LOCKED RapidFuzz issue #110 — empty-input gate fires before identity |
| both_pure_separator_deviation | 0.0 | LOCKED — both Tokenise to `[]`; post-Tokenise empty-set gate |
| one_empty_a | 0.0 | empty-input gate |
| one_empty_b | 0.0 | empty-input gate |
| subset_a_in_b | 1.0 | intersection non-empty, diffAB empty; subset short-circuit |
| subset_b_in_a | 1.0 | intersection non-empty, diffBA empty; subset short-circuit |
| three_way_max_combined_wins | 0.6363636363636364 = 7/11 | r1=r2=0.625; r3=indelRatio("world hello", "world peace")=14/22=7/11; r3 wins |
| disjoint | 0.1428571428571429 = 1/7 | intersection=∅; combined1to2="abc def" (7B); combined2to1="qrs xyz" (7B); LCS=1 (space); r3=2/14=1/7 |
| low_overlap_singletons | 0.2 | intersection=∅; r3=indelRatio("hello","world")=2/10=0.2 |
| dedup_set_equal | 1.0 | A_set={alpha,beta}; B_set={alpha,beta}; both diffs empty; subset short-circuit |
| unicode_reorder | 1.0 | café/société sets equal after Tokenise; subset short-circuit |

## Cross-validation Activation Status

**All 20 `/token_set` sub-tests pass** within epsilon = 1e-9. The committed corpus
contains 20 entries; none are `partial_only: true`, so all 20 sub-tests run their
assertion body. The full `TestTokenRatios_CrossValidation` test runs in ~10ms.

The 4 `/token_sort` sub-tests still PASS (active from plan 06-01). The 40 `/partial_*`
sub-tests still SKIP with the explicit "plan 06-03 will land …" message.

## Deviations from Plan

### Auto-fixed Issues

**1. [Rule 1 — Bug] Three-way max branch (a) — combined-vs-combined, not diff-vs-diff**

- **Found during:** Task 1 design (before the cross-validation activation in Task 2)
- **Issue:** The plan's `<recorded_resolutions>` (correctly) said branch (a) is the "Indel ratio of sorted-joined diff_ab vs sorted-joined diff_ba". RESEARCH.md Pattern 5 pseudocode (line 314-318) also showed branch (a) as `indelRatio(diff_ab_joined, diff_ba_joined)`. But the corpus (which is the ground truth) and the actual RapidFuzz fuzz_py.py source use `indelRatio(combined1to2, combined2to1)` — with the intersection prefix on BOTH sides. The discrepancy was caught by deriving the corpus entries (`tokens_low_overlap` = 7/11 requires the combined-vs-combined formula).
- **Fix:** Implemented the correct formula using `combined1to2 = sortedSect + " " + sortedDiffAB` and `combined2to1 = sortedSect + " " + sortedDiffBA`, with all three indelRatio calls operating over these strings. The plan's prose and the implementation are correct; only RESEARCH.md's pseudocode reference had the wrong shape (which is reference documentation, not load-bearing).
- **Files modified:** `token_set_ratio.go` (algorithm body and godoc); validated by cross-validation pass in Task 2.
- **Verification:** Every corpus entry asserts within epsilon = 1e-9.
- **Committed in:** `880c037` (Task 1 commit — the algorithm landed with the correct formula from inception)

**2. [Rule 1 — Bug] Gate ordering — empty-input gate must fire BEFORE identity short-circuit**

- **Found during:** Task 2 (cross-validation activation surfaced the divergence)
- **Issue:** The Task 1 implementation followed the plan's literal algorithm specification (identity short-circuit at step 1, post-Tokenise empty-set gate at step 3). But this caused `TokenSetRatioScore("", "")` to return 1.0 (via the identity short-circuit) instead of the RapidFuzz-required 0.0. The plan's `<must_haves.truths>` explicitly required `TokenSetRatioScore("", "") == 0.0`, contradicting the algorithm steps.
- **Fix:** Reordered the gates — empty-input gate (`if a == "" || b == "" { return 0.0 }`) now fires BEFORE the identity short-circuit. The identity short-circuit is guarded to non-empty identical strings only. Updated all 6 documentation surfaces (algorithm godoc, function godoc Conventions block, unit test `both_empty_strings_deviation`, property test Identity guard, staging-golden new entry, BDD new scenario) to reflect the corrected gate ordering. Cross-validation now passes on the `both_empty` entry.
- **Files modified:** `token_set_ratio.go`, `token_set_ratio_test.go`, `props_test.go`, `tests/bdd/features/token_set_ratio.feature`, `testdata/golden/_staging/token_set_ratio.json`, `llms-full.txt`.
- **Verification:** `TestTokenRatios_CrossValidation` exits 0 on all 20 token_set sub-tests.
- **Committed in:** `22e25f4` (Task 2 commit — bundled with the cross-validation activation since both are required for cross-validation to pass)

**3. [Rule 1 — Bug] Cyclomatic complexity ceiling**

- **Found during:** Task 1 (lint gate)
- **Issue:** Initial `TokenSetRatioScore` implementation had cyclomatic complexity 18 (golangci-lint threshold = 10). Extracted helper functions reduced parent complexity but didn't quite reach the ceiling on the first refactor (still at 11).
- **Fix:** Extracted three internal helpers — `buildTokenSetPartitions` (set construction + sort), `tokenSetThreeWayMax` (the three indelRatio calls + max), and `joinSectAndDiff` (combined-string conditional separator). Parent function now has complexity 9, under the ceiling without nolint pragmas.
- **Files modified:** `token_set_ratio.go`.
- **Verification:** `make lint` reports 0 issues.
- **Committed in:** `880c037` (Task 1 commit — refactor was applied before the final Task 1 commit)

---

**Total deviations:** 3 auto-fixed (3× Rule 1 — bugs; all caught and resolved during plan execution; no scope creep)
**Impact on plan:** All three fixes are essential — without them the cross-validation gate fails or the lint gate blocks the commit. The deviations are well within Rule 1's intent (correctness bugs surfaced by the verification suite).

## Issues Encountered

None requiring problem-solving beyond the three auto-fixed deviations above. The plan was well-structured and pointed at the right canonical templates (`<read_first>` blocks); implementation tracked the plan within minutes per task. The cross-validation corpus from plan 06-01 was the load-bearing verification artifact — without it, the gate-ordering bug (deviation #2) would have shipped silently and broken the eventual integration tests in later phases.

## User Setup Required

None — no external service configuration required for this plan. The Python / RapidFuzz toolchain remains developer-only (only needed when regenerating the corpus via `make regen-token-ratio-cross-validation`); CI consumes the committed `vectors.json` directly.

## Next Phase Readiness

- **Plan 06-03 (PartialRatio)** is unblocked — can compose against `indelRatio` / `indelRatioRunes` from `token_indel.go` (no further kernel changes needed). Extends `token_ratio_cross_validation_test.go` by removing the `/partial_bytes` and `/partial_runes` t.Skip lines and adding the assertion bodies. The corpus already has `partial_ratio_bytes` and `partial_ratio_runes` fields per entry (set up in plan 06-01).
- **Plan 06-04 (TokenJaccard)** is unblocked — TokenJaccard is a set-Jaccard composition (no Indel kernel needed); it shares the `Tokenise` consumer pattern from this plan but follows the catalogue's standard both-empty → 1.0 convention (NOT the TokenSetRatio deviation). Plan 06-04's plan should explicitly call out that TokenJaccard does NOT inherit TokenSetRatio's deviation.
- **Plan 06-05 (MongeElkan)** inherits the dispatcher slot 13 wiring pattern.
- **Plan 06-06 finalisation** has a 12-entry staging-golden to merge into `algorithms.json` and a new docs page (`docs/algorithms.md` extension) to reference the TokenSetRatio deviation.

No blockers or concerns.

### Deferred items for plan 06-06

- **Final `bench.txt` numbers** — TokenSetRatio benchmarks compile and run (ASCII Short/Medium/Long + Unicode Short + Pathological_AsymmetricSetCardinalities) but the project-wide `bench.txt` baseline is regenerated phase-by-phase at finalisation time, not per-plan. Plan 06-06 will run the full benchmark suite and commit the updated `bench.txt`.
- **`testdata/golden/_staging/token_set_ratio.json` merge into `testdata/golden/algorithms.json`** — staged in the `_staging/` directory; plan 06-06 finalisation handles the merge.
- **Cross-platform determinism golden update** — TokenSetRatioScore on ASCII inputs is deterministic across the four CI platforms by construction (integer-derived divisions + sort.Strings byte-lex), but `verify-determinism`'s golden file does not yet include any TokenSetRatio entries. Plan 06-06 will add representative entries.

## Self-Check: PASSED

File-existence checks:

- `token_set_ratio.go` — present.
- `token_set_ratio_test.go` — present.
- `token_set_ratio_bench_test.go` — present.
- `token_set_ratio_fuzz_test.go` — present.
- `dispatch_token_set_ratio.go` — present.
- `tests/bdd/features/token_set_ratio.feature` — present.
- `testdata/golden/_staging/token_set_ratio.json` — present (12 entries).

Commit-existence checks:

- Commit `880c037` — present in `git log --oneline -5`.
- Commit `22e25f4` — present in `git log --oneline -5`.

Cross-validation:

- `TestTokenRatios_CrossValidation/*/token_set` — 20 sub-tests, all PASS.
- `TestTokenSetRatio*` + `TestProp_TokenSetRatio*` — all PASS.
- `cd tests/bdd && go test ./...` — PASS.
- `make fmt-check lint verify-license-headers verify-deps-allowlist` — all 0 issues.

All claimed deliverables verified by `git log`, file existence on disk, and the test/lint gates above.

---
*Phase: 06-token-based-algorithms*
*Completed: 2026-05-15*
