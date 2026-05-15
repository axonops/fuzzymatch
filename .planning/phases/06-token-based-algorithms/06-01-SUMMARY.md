---
phase: 06-token-based-algorithms
plan: 01
subsystem: algorithm-catalogue
tags: [token-sort-ratio, indel-formula, lcs-subsequence, wagner-fischer, rapidfuzz, cross-validation]

# Dependency graph
requires:
  - phase: 01-foundation-infrastructure
    provides: [AlgoID dispatch table, Tokenise, errors, CI matrix, license-headers gate]
  - phase: 02-core-character-algorithms-six
    provides: [maxStackInputLen constant, two-row DP discipline]
  - phase: 04-remaining-character-gestalt
    provides: [LongestCommonSubstring (used for PITFALL-6 divergence regression test), cross-validation script template]
  - phase: 05-q-gram-algorithms
    provides: [shared-helper + export_test.go re-export pattern, per-plan llms.txt sync discipline]

provides:
  - Shared token_indel.go kernel (unexported lcsLen / indelRatio / lcsLenRunes / indelRatioRunes — package-internal LCS-subsequence + Indel-formula helpers)
  - TokenSortRatioScore(a, b string) float64 — first Indel-consumer in the catalogue (AlgoTokenSortRatio dispatch slot 14 wired)
  - testdata/cross-validation/token-ratios/vectors.json (20 entries × 4 surfaces) — RapidFuzz 3.14.5 reference corpus for Wave 2/3 consumption
  - scripts/gen-token-ratio-cross-validation.py with rapidfuzz version-pin gate
  - Makefile target regen-token-ratio-cross-validation
  - docs/cross-validation.md (contributor-facing reference)
  - Structurally-complete TestTokenRatios_CrossValidation loader (token_sort asserts, token_set/partial_bytes/partial_runes skip-pending plans 06-02 / 06-03)

affects: [06-02-token-set-ratio, 06-03-partial-ratio, 06-04-token-jaccard, 06-05-monge-elkan, 06-06-finalisation, 08-scorer, 10-extract]

# Tech tracking
tech-stack:
  added:
    - rapidfuzz==3.14.5 (developer-only Python dep — NOT in root go.mod allowlist)
  patterns:
    - "Shared unexported kernel + export_test.go re-export (token_indel.go + 4 *ForTest helpers) — mirrors Phase 5 q_gram.go pattern"
    - "Cross-validation corpus with _metadata block (rapidfuzz_version + python_version + regenerated_at) — pin gate at both Python (assert) and Go (loader test) ends"
    - "Combined four-score-per-entry corpus with structurally-complete loader gated per surface — Wave-1 asserts only TokenSort, plans 06-02 / 06-03 remove t.Skip lines and add assertions"
    - "Tokeniser-divergence reconciliation (.lower() + Tokenise-safety gate) — corpus restricted to lowercase whitespace-only ASCII for tokenised algorithms; partial_only flag for character-level fixtures"
    - "Per-plan llms.txt + llms-full.txt sync (reinforced from Phase 5) — every new exported symbol gets a line in llms.txt and a godoc block in llms-full.txt in the same commit"
    - "Direct dispatch wrapper without closure for parameter-free algorithms (mirrors dispatch_lcsstr.go) — distinct from q-gram tier's default-n closure"

key-files:
  created:
    - "token_indel.go (shared LCS-subseq + Indel kernel)"
    - "token_indel_test.go (kernel + Pitfall-6 regression gate)"
    - "token_sort_ratio.go (algorithm)"
    - "token_sort_ratio_test.go / _bench_test.go / _fuzz_test.go"
    - "dispatch_token_sort_ratio.go"
    - "token_ratio_cross_validation_test.go (four-surface loader)"
    - "scripts/gen-token-ratio-cross-validation.py"
    - "testdata/cross-validation/token-ratios/vectors.json (20 entries)"
    - "testdata/golden/_staging/token_sort_ratio.json (10 entries)"
    - "tests/bdd/features/token_sort_ratio.feature"
    - "docs/cross-validation.md"
  modified:
    - "export_test.go (added 4 *ForTest re-exports)"
    - "algoid_test.go (added AlgoTokenSortRatio to TestDispatch_UnregisteredSlotsAreNil registered map)"
    - "props_test.go (appended 6 TokenSortRatio property tests)"
    - "example_test.go (appended ExampleTokenSortRatioScore)"
    - "tests/bdd/steps/algorithms_steps.go (TokenSortRatio step methods + registrations)"
    - "llms.txt (added TokenSortRatio section)"
    - "llms-full.txt (added Phase 6 algorithm surface block)"
    - "Makefile (added regen-token-ratio-cross-validation target + .PHONY)"
    - "CONTRIBUTING.md (documented regen-token-ratio-cross-validation target)"

key-decisions:
  - "OQ-1 RESOLUTION LOCKED 2026-05-15: tokeniser-divergence handled by (a) corpus restricted to lowercase whitespace-only ASCII for tokenised entries, (b) .lower() reconciliation in generator, (c) docs/cross-validation.md documents the divergence, (d) algorithm godoc explicitly notes Tokenise vs str.split asymmetry"
  - "OQ-2 RESOLUTION LOCKED 2026-05-15: single combined vectors.json with all four scores per entry (token_sort_ratio, token_set_ratio, partial_ratio_bytes, partial_ratio_runes); per-entry per-surface t.Run sub-tests"
  - "OQ-3 RESOLUTION LOCKED 2026-05-15: partial_ratio_runes always emitted (matches partial_ratio_bytes on ASCII inputs; exercises the separate rune-path implementation in plan 06-03)"
  - "RapidFuzz version pin LOCKED 2026-05-15: RAPIDFUZZ_VERSION = \"3.14.5\" — recorded in script header, vectors.json _metadata, docs/cross-validation.md, and the Go loader test (assertion on c.Metadata.RapidFuzzVersion field)"
  - "Stack-buffer threshold LOCKED 2026-05-15: lcsLen / lcsLenRunes use [maxStackInputLen+1]int = [65]int stack-allocated DP rows when min(|a|,|b|) <= 64; heap fallback make([]int, m+1) for longer inputs (mirrors levenshtein.go's maxStackInputLen budget)"

patterns-established:
  - "Shared token-tier kernel: token_indel.go follows the q_gram.go template — primary-source citation block, Source-Origin Statement, DET-03/DET-06 reaffirmation, allocation-budget block, plan-introducer/consumer-plan godoc note in export_test.go re-exports"
  - "Pitfall-6 regression gate: TestLCSLen_DistinctFromLCSStr asserts BOTH absolute lcsLen('abc','axc')==2 AND len(LongestCommonSubstring('abc','axc'))==1 AND their inequality in a single test function so the LCS-subsequence vs LCS-substring divergence cannot regress silently"
  - "Cross-validation generator with rapidfuzz version-pin gate: script's first action after import is assert rapidfuzz.__version__ == RAPIDFUZZ_VERSION; refuses to run on any other version"
  - "Tokenise-safety gate in generator: _assert_corpus_is_tokenise_safe rejects non-partial_only entries containing anything outside [a-z ]; .lower() reconciliation applied to RapidFuzz inputs"
  - "Four-surface loader skip-pending pattern: token_set / partial_bytes / partial_runes sub-tests skip with explicit plan references; subsequent plans remove t.Skip and add assertions without restructuring the test"

requirements-completed:
  - TOKEN-02

# Metrics
duration: 33min
completed: 2026-05-15
---

# Phase 6 Plan 1: Token Sort Ratio + Token-Tier Foundation Summary

**Shared LCS-subsequence + Indel-formula kernel (token_indel.go) plus TokenSortRatio (Wagner-Fischer 1974 / RapidFuzz Indel formula) with the four-surface RapidFuzz 3.14.5 cross-validation corpus and structurally-complete loader gated for plans 06-02 / 06-03 extension.**

## Performance

- **Duration:** ~33 min
- **Started:** 2026-05-15T10:11:34Z (worktree HEAD reset)
- **Completed:** 2026-05-15T10:25:20Z (final verification gate)
- **Tasks:** 3 (Task 1 — kernel; Task 2 — TokenSortRatio; Task 3 — cross-validation infra)
- **Files modified:** 19 (10 created, 9 modified) excluding the SUMMARY.md itself

## Accomplishments

- Landed the load-bearing shared kernel in `token_indel.go` — four unexported helpers (`lcsLen`, `indelRatio`, `lcsLenRunes`, `indelRatioRunes`) backed by the Wagner-Fischer 1974 two-row DP. Stack buffer when `min(|a|,|b|) ≤ 64`; heap fallback otherwise. Zero heap allocations on short ASCII inputs.
- Shipped `TokenSortRatioScore(a, b string) float64` — the simplest Indel-formula consumer in the catalogue. Identity short-circuit, tokenise-sort-join composition, and dispatch slot 14 wiring. Full unit + property + fuzz + bench + BDD + staging-golden coverage.
- Established the **RapidFuzz cross-validation infrastructure** for the entire token tier in one commit: pinned `rapidfuzz==3.14.5` Python generator, committed `vectors.json` with 20 entries × 4 scores per entry, structurally-complete Go loader that asserts TokenSortRatio agreement on every entry while skipping `token_set` / `partial_bytes` / `partial_runes` sub-tests with explicit "plan 06-NN will land" messages until plans 06-02 / 06-03 land.
- Recorded **five `OQ-N RESOLUTION LOCKED` decisions** for downstream plans to inherit: tokeniser-divergence handling, single-combined-corpus shape, partial_ratio_runes always-emitted, RapidFuzz version pin, stack-buffer threshold.

## Task Commits

Each task was committed atomically (per-task convention; no TDD RED/GREEN sub-commits because all three tasks ship pre-verified code with full test suites that exercise the new surface):

1. **Task 1: token_indel.go shared LCS-subsequence + Indel-ratio kernel** — `db0167d` (feat)
2. **Task 2: TokenSortRatio (algorithm + dispatch + companions + BDD + golden)** — `f820f92` (feat)
3. **Task 3: RapidFuzz cross-validation infra (script + corpus + loader + docs)** — `dcc1193` (feat)

The SUMMARY.md commit follows separately (the final metadata commit per `execute-plan.md`).

## Files Created/Modified

### Created

- `token_indel.go` — shared kernel: Wagner-Fischer 1974 LCS-subsequence two-row DP + Indel-formula normalisation. Four unexported helpers; stack-buffer fast path when `min(|a|,|b|) ≤ 64`.
- `token_indel_test.go` — kernel-level unit tests + the **load-bearing PITFALL 6 regression gate** (`TestLCSLen_DistinctFromLCSStr` asserts `lcsLen("abc","axc")==2` and `len(LongestCommonSubstring("abc","axc"))==1` and their inequality in a single test function).
- `token_sort_ratio.go` — `TokenSortRatioScore`. Identity short-circuit before Tokenise; both-Tokenised-empty → 1.0; one-Tokenised-empty → 0.0; sort.Strings + strings.Join + indelRatio.
- `token_sort_ratio_test.go` — table-driven unit tests covering identity / both-empty / one-empty / token-reorder / subset / disjoint / Unicode / pure-separator-both / pure-separator-one / dispatch registration / symmetry.
- `token_sort_ratio_bench_test.go` — ASCII Short / Medium / Long + Unicode Short benchmarks with `b.ReportAllocs()` + sink-gate.
- `token_sort_ratio_fuzz_test.go` — `FuzzTokenSortRatioScore` with 11 programmatic seeds; asserts no-NaN, no-Inf, range bounds, AND the identical-input regression check (load-bearing for the identity short-circuit).
- `dispatch_token_sort_ratio.go` — registers `dispatch[AlgoTokenSortRatio] = TokenSortRatioScore` (no closure; signature matches dispatch table).
- `token_ratio_cross_validation_test.go` — `TestTokenRatios_CrossValidation` four-surface loader (token_sort asserts; token_set / partial_bytes / partial_runes skip-pending plans 06-02 / 06-03) + `TestTokenRatios_CrossValidation_CorpusShape` (≥ 20 entries; category coverage gate).
- `scripts/gen-token-ratio-cross-validation.py` — RapidFuzz-pinned generator (3.14.5). Refuses to run on any other rapidfuzz version. Python ≥ 3.7 gate via `_check_python_version()`. Tokenise-safety gate via `_assert_corpus_is_tokenise_safe`. Emits all four scores per entry plus `_metadata` block.
- `testdata/cross-validation/token-ratios/vectors.json` — 20-entry RapidFuzz reference corpus. `_metadata.rapidfuzz_version == "3.14.5"`. Score fields byte-stable across repeated regenerations (only `regenerated_at` timestamp varies; timestamp is informational, not asserted).
- `testdata/golden/_staging/token_sort_ratio.json` — 10 staging-golden entries for merge into `testdata/golden/algorithms.json` by plan 06-06 finalisation.
- `tests/bdd/features/token_sort_ratio.feature` — Gherkin scenarios: canonical reference vectors, identity, both-empty, one-empty, token-reorder, symmetry. All scenarios tagged `@token @token-sort-ratio`.
- `docs/cross-validation.md` — contributor-facing reference for the corpus (RapidFuzz pin protocol, OQ-1 / OQ-2 / OQ-3 resolutions, TokenSetRatio empty-set deviation, regeneration command).

### Modified

- `export_test.go` — appended 4 `*ForTest` re-exports (`LCSLenForTest`, `IndelRatioForTest`, `LCSLenRunesForTest`, `IndelRatioRunesForTest`) with plan-introducer / consumer-plan godoc paragraphs.
- `algoid_test.go` — added `AlgoTokenSortRatio: true` to `TestDispatch_UnregisteredSlotsAreNil`'s `registered` map (plan 06-01 flips slot 14 to registered).
- `props_test.go` — appended 6 `TestProp_TokenSortRatioScore_*` property tests (RangeBounds, Identity, Symmetric, NoNaN, NoInf, NoNegativeZero) using `testing/quick`.
- `example_test.go` — appended `ExampleTokenSortRatioScore` (canonical fuzzy-wuzzy reorder pair → 1.0).
- `tests/bdd/steps/algorithms_steps.go` — appended TokenSortRatio step methods (`iComputeTheTokenSortRatioScoreBetween`, `iComputeTheSecondTokenSortRatioScoreBetween`, `bothTokenSortRatioScoresShouldBeEqual`) and their `ctx.Step` registrations.
- `llms.txt` — appended `### TokenSortRatio` section before Normalisation.
- `llms-full.txt` — appended `### Phase 6 algorithm surface (token tier — TokenSortRatio)` block before Normalisation. Documents the Indel-formula equivalence, OQ-1 tokeniser-divergence note, and per-plan llms-sync discipline reference.
- `Makefile` — added `regen-token-ratio-cross-validation` target with the rapidfuzz install hint; added the target to `.PHONY`.
- `CONTRIBUTING.md` — documented `regen-token-ratio-cross-validation` alongside the existing two regen-* targets.

## Decisions Made

**The five required `OQ-N RESOLUTION LOCKED` decisions and the RapidFuzz / stack-buffer locks are recorded verbatim:**

- **OQ-1 RESOLUTION LOCKED 2026-05-15** — Tokeniser-divergence handled by (a) restricting the cross-validation corpus to whitespace-only lowercase ASCII inputs (enforced by `_assert_corpus_is_tokenise_safe` in the generator), (b) the generator script calls `.lower()` on every input before passing to RapidFuzz, (c) `docs/cross-validation.md` documents the divergence prominently, (d) `token_sort_ratio.go`'s godoc explicitly notes that the algorithm uses `Tokenise(s, DefaultTokeniseOptions())` (camelCase-aware) — for whitespace-only-lowercase-ASCII inputs the behaviour matches RapidFuzz; for mixed identifier-style inputs the project tokenisation produces semantically richer splits. PartialRatio entries (character-level, no tokenise) may carry `"partial_only": true` to skip the tokenise-safety check; their `token_sort_ratio` and `token_set_ratio` fields are emitted as `null` and the Go loader skips those per-surface sub-tests via the `*float64` pointer-nil check.
- **OQ-2 RESOLUTION LOCKED 2026-05-15** — Single combined `vectors.json` file with all four scores per entry (`token_sort_ratio`, `token_set_ratio`, `partial_ratio_bytes`, `partial_ratio_runes`). Per-entry sub-tests via `t.Run(e.Name+"/<surface>", ...)`. Mirrors Phase 3/4 structure. The Wave-1 loader asserts only TokenSortRatio entries; plans 06-02 / 06-03 extend the loop to assert TokenSetRatio / PartialRatio surfaces as those algorithms land (they remove the per-surface `t.Skip` lines and add the assertion bodies; no structural change to the loader required).
- **OQ-3 RESOLUTION LOCKED 2026-05-15** — `partial_ratio_runes` is always included in every entry — for ASCII inputs the value matches `partial_ratio_bytes` (the rune path's separate code path is exercised even on ASCII to catch path-specific regressions).
- **RapidFuzz version pin LOCKED 2026-05-15** — `RAPIDFUZZ_VERSION = "3.14.5"` (current stable per PyPI registry; verified installed in the developer workstation and the corpus regenerated against it). Recorded in script header (top-level `assert rapidfuzz.__version__ == RAPIDFUZZ_VERSION`), in `vectors.json` `_metadata.rapidfuzz_version`, in `docs/cross-validation.md`, AND in the Go loader (`TestTokenRatios_CrossValidation_CorpusShape` asserts `c.Metadata.RapidFuzzVersion == "3.14.5"`). Bumping requires the five-step protocol documented in `docs/cross-validation.md`.
- **Stack-buffer threshold LOCKED 2026-05-15** — `[maxStackInputLen+1]int` = `[65]int` stack-allocated DP rows when `min(|a|,|b|) <= maxStackInputLen` (= 64 from `levenshtein.go`); heap fallback `make([]int, m+1)` for longer inputs. CONTEXT.md §2 recommended `[64]int` at `min(m,n) ≤ 50`; the planner finalised at 64/64 to share `maxStackInputLen` with the Phase 2 budget. This matches the levenshtein.go / damerau_*.go / lcsstr.go / swg.go discipline.

**Reference vector numbers for the staging-golden file** (`testdata/golden/_staging/token_sort_ratio.json` — 10 entries, all locked):

| Name | Score | Derivation (`joinedA` / `joinedB` / lcs / formula) |
|------|------:|----------------------------------------------------|
| both_empty | 1.0 | a == b identity short-circuit |
| identity | 1.0 | a == b identity short-circuit |
| one_empty_a | 0.0 | one-empty convention |
| one_empty_b | 0.0 | one-empty convention |
| token_reorder_two | 1.0 | `"alpha beta"`/`"alpha beta"`; lcs=10; 2·10/20 = 1 |
| token_reorder_canonical | 1.0 | identical sorted-joined strings |
| subset_short | 0.6666… | `"alpha"`/`"alpha beta"`; lcs=5; 2·5/15 |
| subset_mid | 0.7692… | `"alpha beta"`/`"alpha beta gamma"`; lcs=10; 2·10/26 |
| disjoint | 0.0 | `"abc"`/`"xyz"`; lcs=0 |
| low_overlap | 0.2 | `"hello"`/`"world"`; lcs=1 (`"l"`); 2·1/10 |

## Deviations from Plan

### Auto-fixed Issues

**1. [Rule 3 - Blocking] Update `TestDispatch_UnregisteredSlotsAreNil` to include `AlgoTokenSortRatio`**
- **Found during:** Task 2 (TokenSortRatio dispatch registration)
- **Issue:** `algoid_test.go::TestDispatch_UnregisteredSlotsAreNil` enumerates all registered dispatch slots in a `map[int]bool{...}` and asserts every other slot is `nil`. Registering `dispatch[AlgoTokenSortRatio]` in `dispatch_token_sort_ratio.go` made slot 14 non-nil and broke this test.
- **Fix:** Added `int(fuzzymatch.AlgoTokenSortRatio): true` to the `registered` map, updated the surrounding comments to mention plan 06-01 and the slot-14 registration milestone.
- **Files modified:** `algoid_test.go`
- **Verification:** `go test -run TestDispatch_UnregisteredSlotsAreNil -count=1 ./...` passes after the update.
- **Committed in:** `f820f92` (Task 2 commit — bundled with the dispatch wiring it accompanies)

**2. [Rule 2 - Missing Critical] Document `regen-token-ratio-cross-validation` in CONTRIBUTING.md**
- **Found during:** Task 3 (Makefile target addition)
- **Issue:** `makefile_targets_test.go::TestMakefile_TargetsDocumentedInContributing` requires every Makefile target to appear in `CONTRIBUTING.md`'s Make Targets section OR carry a `## suppress: <reason>` comment. The plan didn't specify the documentation update.
- **Fix:** Added a bullet for `regen-token-ratio-cross-validation` in `CONTRIBUTING.md` alongside the existing two regen-* targets (mirrors the documentation pattern from plans 03-01 and 04-04).
- **Files modified:** `CONTRIBUTING.md`
- **Verification:** `go test -run TestMakefile_TargetsDocumentedInContributing -count=1 ./...` passes after the update.
- **Committed in:** `dcc1193` (Task 3 commit — bundled with the Makefile target it documents)

---

**Total deviations:** 2 auto-fixed (1 blocking — test-fixture sync; 1 missing critical — required-by-test-gate documentation)
**Impact on plan:** Both auto-fixes are mandatory for the build to pass. No scope creep — both deviations are well within Rules 2/3's intent (essential corrections to keep CI green).

## Issues Encountered

None requiring problem-solving. The plan was extremely explicit (`<read_first>` blocks pointed at the canonical templates; `<action>` blocks named every file and the structure to copy). Implementation tracked the plan within minutes per task.

A noted observation rather than an issue: the RapidFuzz-3.14.5 install was already in place from the developer workstation; in a clean CI environment the corpus is read as a static fixture and Python is not invoked, so this is not a problem.

## User Setup Required

None — no external service configuration required for this plan. The RapidFuzz Python dependency is developer-only (required only when regenerating the corpus via `make regen-token-ratio-cross-validation`); CI consumes the committed `vectors.json` directly without invoking Python. The corresponding install command (`python3 -m pip install --user rapidfuzz==3.14.5`) is documented in three places: the Makefile target, CONTRIBUTING.md, and docs/cross-validation.md.

## Next Phase Readiness

- **Plan 06-02 (TokenSetRatio)** is unblocked — can compose against `indelRatio` from `token_indel.go` (via `LCSLenForTest` / `IndelRatioForTest` in test code), extend the cross-validation loader by removing the `token_set` sub-test `t.Skip` and adding the assertion body, and inherit the corpus's `token_set_ratio` field that already exists for every Tokenise-safe entry. The TokenSetRatio empty-token-set deviation (returns 0.0 for both-empty per RapidFuzz issue #110 / fuzzywuzzy parity) is documented in `docs/cross-validation.md`.
- **Plans 06-03 (PartialRatio) and 06-05 (MongeElkan inner-metric integration)** inherit the same kernel (`lcsLen` / `indelRatio` / rune-aware twins) without re-implementing the DP. The rune-aware kernel (`lcsLenRunes` / `indelRatioRunes`) is in place for PartialRatio's runes surface.
- **Plan 06-06 finalisation** has a 10-entry staging-golden to merge into `algorithms.json` and a new docs page (`docs/cross-validation.md`) to cross-reference from the per-algorithm docs.

No blockers or concerns.

### Deferred items for plan 06-06

- **Final `bench.txt` numbers** — TokenSortRatio benchmarks compile and run (ASCII Short/Medium/Long + Unicode Short) but the project-wide `bench.txt` baseline is regenerated phase-by-phase at finalisation time, not per-plan. Plan 06-06 will run the full benchmark suite and commit the updated `bench.txt`.
- **`testdata/golden/_staging/token_sort_ratio.json` merge into `testdata/golden/algorithms.json`** — staged in the `_staging/` directory; plan 06-06 finalisation handles the merge (mirrors Phase 4 / Phase 5 finalisation flow).
- **Cross-platform determinism golden update** — TokenSortRatioScore on ASCII inputs is deterministic across the four CI platforms by construction (integer-derived single division), but `verify-determinism`'s golden file does not yet include any TokenSort entries. Plan 06-06 will add representative TokenSort entries to the golden file as part of the phase-wide determinism gate refresh.

## Self-Check: PASSED

- `token_indel.go` — present.
- `token_indel_test.go` — present.
- `token_sort_ratio.go` — present.
- `token_sort_ratio_test.go` — present.
- `token_sort_ratio_bench_test.go` — present.
- `token_sort_ratio_fuzz_test.go` — present.
- `dispatch_token_sort_ratio.go` — present.
- `token_ratio_cross_validation_test.go` — present.
- `scripts/gen-token-ratio-cross-validation.py` — present.
- `testdata/cross-validation/token-ratios/vectors.json` — present (20 entries; `_metadata.rapidfuzz_version == "3.14.5"`).
- `testdata/golden/_staging/token_sort_ratio.json` — present (10 entries).
- `tests/bdd/features/token_sort_ratio.feature` — present.
- `docs/cross-validation.md` — present.
- Commit `db0167d` — present in `git log`.
- Commit `f820f92` — present in `git log`.
- Commit `dcc1193` — present in `git log`.

All claimed deliverables verified by `git log --oneline -5` and `ls`-equivalent file existence on disk.

---
*Phase: 06-token-based-algorithms*
*Completed: 2026-05-15*
