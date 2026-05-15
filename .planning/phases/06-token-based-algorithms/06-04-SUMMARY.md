---
phase: 06-token-based-algorithms
plan: 04
subsystem: algorithm-catalogue
tags: [token-jaccard, set-jaccard, hand-derived-rv, set-vs-multiset-distinction, det-03-integer-counter]

# Dependency graph
requires:
  - phase: 01-foundation-infrastructure
    provides: [AlgoID dispatch table, Tokenise, errors, CI matrix, license-headers gate]
  - phase: 05-q-gram-algorithms
    provides: [hand-derived reference vector pattern, file-header Source-Origin Statement template, integer-counter intersection-cardinality pattern (qgram_jaccard.go::jaccardFromQGramMaps)]
  - phase: 06-token-based-algorithms / plan: 06-01
    provides: [shared tokenisation surface — Tokenise(s, DefaultTokeniseOptions()), per-plan llms.txt sync discipline, parameter-free dispatch wrapper pattern (dispatch_token_sort_ratio.go), Phase 6 algorithm file-header layout with tokeniser-divergence note]

provides:
  - TokenJaccardScore(a, b string) float64 — set-Jaccard over Tokenise output (AlgoTokenJaccard dispatch slot 17 wired)
  - testdata/golden/_staging/token_jaccard.json — 10 staging entries for plan 06-06 merge into algorithms.json
  - tests/bdd/features/token_jaccard.feature — Godog scenarios including the load-bearing set-vs-multiset distinction scenario

affects: [06-05-monge-elkan, 06-06-finalisation, 08-scorer, 10-extract]

# Tech tracking
tech-stack:
  added: []   # no new dependencies (stdlib + golang.org/x/text only via Tokenise)
  patterns:
    - "Hand-derived reference vector pattern (no Python cross-validation) — RV-TJ1..RV-TJ6 each carrying the formula derivation in the test comment; mirrors Phase 5's qgram_jaccard_test.go RV-J1..RV-J6 (set-Jaccard is unambiguous from Jaccard 1912 so RapidFuzz cross-validation is not needed)"
    - "Set-vs-multiset semantic distinction pinned in four surfaces: algorithm godoc / unit test (TestTokenJaccardScore_SetVsMultisetDistinction) / BDD scenario (set semantics deduplicate repeated tokens) / staging-golden RV-TJ3 entry"
    - "Integer-counter intersection cardinality satisfying DET-03 without sort — iterate the SMALLER set and probe the larger set; output is a scalar int (associative integer addition; map iteration order does not affect the result)"
    - "Direct dispatch wrapper without closure for parameter-free algorithms — mirrors dispatch_lcsstr.go and dispatch_token_sort_ratio.go"
    - "Per-plan llms.txt + llms-full.txt sync (reinforced from Phase 5 / plan 06-01) — every new exported symbol gets a line in llms.txt and a godoc block in llms-full.txt in the same commit"

key-files:
  created:
    - "token_jaccard.go (algorithm)"
    - "token_jaccard_test.go (unit + reference vectors + dispatch registration + set-vs-multiset keystone + alloc budget)"
    - "token_jaccard_bench_test.go (ASCII Short/Medium/Long + Unicode Short)"
    - "token_jaccard_fuzz_test.go (FuzzTokenJaccardScore with 16 programmatic seeds)"
    - "dispatch_token_jaccard.go (direct dispatch slot wiring)"
    - "tests/bdd/features/token_jaccard.feature (Godog scenarios)"
    - "testdata/golden/_staging/token_jaccard.json (10 staging entries)"
  modified:
    - "algoid_test.go (added AlgoTokenJaccard to TestDispatch_UnregisteredSlotsAreNil registered map)"
    - "props_test.go (appended 6 TokenJaccard property tests — RangeBounds / Identity / Symmetric / NoNaN / NoInf / NoNegativeZero)"
    - "example_test.go (appended ExampleTokenJaccardScore)"
    - "tests/bdd/steps/algorithms_steps.go (TokenJaccard step methods + InitializeScenario registrations)"
    - "llms.txt (added TokenJaccard section before Normalisation)"
    - "llms-full.txt (added Phase 6 algorithm surface block — TokenJaccard — before Normalisation)"

key-decisions:
  - "Set vs Multiset LOCKED 2026-05-15: TokenJaccard uses SET semantics (map[string]struct{}) — DISTINCT from Phase 5's Q-Gram Jaccard MULTISET semantics. Token presence is a binary signal; q-gram presence is a multiplicity signal. The KEYSTONE RV-TJ3 ('a a b' vs 'a b' → 1.0) regression gate is pinned in 4 surfaces."
  - "DET-03 intersection-cardinality LOCKED 2026-05-15: |A ∩ B| computed via integer-counter loop over the smaller set; output is a scalar int (associative integer addition; map iteration order does not affect the result). Union via set inclusion-exclusion |A ∪ B| = |A| + |B| - |A ∩ B|. No sort needed."
  - "Both-empty STANDARD convention LOCKED 2026-05-15: TokenJaccardScore('', '') returns 1.0 (catalogue standard both-empty identity convention). DOES NOT deviate like TokenSetRatio (which returns 0.0 per LOCKED RapidFuzz issue #110 bug-for-bug compatibility). The identity short-circuit covers the both-empty case directly; a post-Tokenise both-empty guard handles pure-separator inputs."
  - "Hand-derived reference vectors LOCKED 2026-05-15: six (a, b) pairs RV-TJ1..RV-TJ6 per CONTEXT.md §1b LOCKED — no RapidFuzz cross-validation (the RapidFuzz cross-validation corpus does NOT include TokenJaccard entries; set-Jaccard is unambiguous from Jaccard 1912). Each test case carries the formula derivation in the test comment."

patterns-established:
  - "Hand-derived RV-only algorithm (no cross-validation corpus) — TokenJaccard is the first Phase 6 algorithm shipped WITHOUT a Python cross-validation corpus; the four Indel-based ratios (TokenSort, TokenSet, PartialRatio bytes, PartialRatio runes) use RapidFuzz 3.14.5 cross-validation. The set-Jaccard formula is unambiguous from the primary source so reviewer-verifiable hand-derivation is sufficient. Monge-Elkan (plan 06-05) follows the same hand-derived discipline per CONTEXT §1b LOCKED."
  - "Set-vs-Multiset distinction pinned in 4 surfaces — algorithm godoc / unit test / BDD scenario / staging-golden. The keystone fixture is RV-TJ3 ('a a b' vs 'a b' → 1.0); a multiset-based regression would surface as 4 simultaneous test failures."
  - "Cyclomatic-complexity discipline — TokenJaccardScore extracts tokensToSet and setIntersectionCardinality helpers to keep the main score function below the gocyclo budget of 10. Both helpers carry godoc explaining the DET-03 invariant."

requirements-completed:
  - TOKEN-05

# Metrics
duration: 10min
completed: 2026-05-15
---

# Phase 6 Plan 4: TokenJaccard Summary

**Set-Jaccard over the deduplicated SET of tokens from Tokenise — the simplest Phase 6 algorithm. No shared kernel beyond Tokenise; no LCS DP, no three-way max, no sliding window. Closes TOKEN-05.**

## Performance

- **Duration:** ~10 min
- **Started:** 2026-05-15T11:26:21Z (after worktree HEAD reset to d50ecbd)
- **Completed:** 2026-05-15T11:36:44Z (final verification gate)
- **Tasks:** 1 (Task 1 — TokenJaccard algorithm + companions + tests + BDD + staging golden + llms sync)
- **Files modified:** 13 (7 created, 6 modified) excluding the SUMMARY.md itself

## Accomplishments

- Shipped `TokenJaccardScore(a, b string) float64` — set-Jaccard `|A ∩ B| / |A ∪ B|` over the deduplicated SET of tokens produced by `Tokenise(s, DefaultTokeniseOptions())`. Jaccard 1912 p. 43 primary source (same paper as Q-Gram Jaccard but applied to TOKEN sets instead of q-gram multisets). Dispatch slot 17 (`AlgoTokenJaccard`) wired via `dispatch_token_jaccard.go`.
- Pinned the **set-vs-multiset distinction** in four surfaces: algorithm godoc (with RESEARCH.md Pattern 8 reference), unit test `TestTokenJaccardScore_SetVsMultisetDistinction` (asserts the keystone result AND the inequality with `QGramJaccardScore` at n=1 on the same input), BDD scenario "set semantics deduplicate repeated tokens (distinct from Q-Gram Jaccard multiset)", and staging-golden RV-TJ3 entry. The KEYSTONE fixture is `("a a b", "a b") → 1.0`; a multiset regression would surface as 4 simultaneous test failures.
- Established **integer-counter intersection cardinality** as the canonical DET-03-satisfying pattern — iterate the SMALLER set and probe the larger set; output is a scalar int (associative integer addition; map iteration order does not affect the result). No sort needed.
- Recorded **four LOCKED decisions** (Set vs Multiset / DET-03 / Both-empty STANDARD / Hand-derived RVs) for downstream plans 06-05 / 06-06 / 08-scorer / 10-extract to inherit.
- Full **unit + property + fuzz + bench + BDD + staging-golden** coverage; **llms.txt + llms-full.txt** synced in the same commit.

## Task Commits

Single atomic commit per the plan's `<tasks>` block:

1. **Task 1: Land TokenJaccard (algorithm + dispatch + companions + property tests + example + BDD + staging golden + per-plan llms sync)** — `e6a4a97` (feat)

The SUMMARY.md commit follows separately (the final metadata commit per `execute-plan.md`).

## Files Created/Modified

### Created

- `token_jaccard.go` — `TokenJaccardScore`. Apache-2.0 header → file-header godoc citing Jaccard 1912 p. 43 (same primary source as Q-Gram Jaccard; applied to TOKEN sets) → algorithm description with explicit set-vs-multiset comparison per RESEARCH.md Pattern 8 → conventions block (both-empty / identity / one-empty) → Source-Origin Statement (NO RapidFuzz cross-validation — hand-derived RV-TJ1..RV-TJ6 per CONTEXT §1b LOCKED) → implementation-discipline block (DET-03 integer-counter intersection cardinality; identity short-circuit BEFORE Tokenise) → allocation budget (≤ 4 baseline / ≤ 8 ceiling). Two unexported helpers: `tokensToSet` (deduplicates a `[]string` to `map[string]struct{}`) and `setIntersectionCardinality` (integer-counter scan over the smaller set; satisfies DET-03 by construction). Identity short-circuit (`a == b → 1.0`) covers both-empty and identical; post-Tokenise both-empty / one-empty guards cover the pure-separator and asymmetric-empty cases.
- `token_jaccard_test.go` — `package fuzzymatch_test`; stdlib `testing` only. Tests:
  - `TestTokenJaccardScore` — 9 sub-tests covering RV-TJ1..RV-TJ6 plus both-empty / one-empty-A / one-empty-B (each row carries its formula derivation in the test comment per algorithm-correctness-standards).
  - `TestTokenJaccardScore_Symmetric` — pins `J(A, B) == J(B, A)` bit-for-bit across 5 input pairs.
  - `TestTokenJaccardScore_DispatchRegistration` — exercises `DispatchInvokeForTest(int(AlgoTokenJaccard), "a b c", "b c d") == 0.5`.
  - `TestTokenJaccardScore_SetVsMultisetDistinction` — KEYSTONE regression gate: asserts `TokenJaccardScore("a a b", "a b") == 1.0` AND that `QGramJaccardScore("a a b", "a b", 1)` produces a DIFFERENT score (≠ 1.0) under multiset semantics.
  - `TestTokenJaccardScore_AllocsBudget` — alloc-count ceiling at 8 for short-input call.
- `token_jaccard_bench_test.go` — `BenchmarkTokenJaccardScore_ASCII_Short` (10-char RV-TJ1 pair) / `_ASCII_Medium` (~50 chars, 7-token shared core) / `_ASCII_Long` (~200 chars, 40 token-pairs) / `_Unicode_Short` (multi-byte UTF-8 token pair). Each uses `b.ReportAllocs()` + `var sink` + `sink < 0` gate. No pathological fixture — TokenJaccard is O(|tokens|) set construction + O(min(|setA|, |setB|)) intersection scan, linear in input size, no DoS-class behaviour.
- `token_jaccard_fuzz_test.go` — `FuzzTokenJaccardScore` with 16 programmatic seeds covering the six reference vectors, identity, both-empty, one-empty, token-reorder, invalid UTF-8 (`\xff\xfe`, `\xc0\x80`), multi-byte UTF-8 (`café münchen`), dedup-heavy (`"a "` repeated 100 vs 50), pure separators (`"___"`, `"..."`), and tokeniser-divergence (`"userID"` vs `"user_id"`). Body asserts no NaN, no ±Inf, score in `[0, 1]`, AND the identity short-circuit regression (`TokenJaccardScore(a, a) == 1.0` for every fuzz input).
- `dispatch_token_jaccard.go` — `var _ = func() bool { dispatch[AlgoTokenJaccard] = TokenJaccardScore; return true }()` — direct dispatch wrapper without closure; mirrors `dispatch_lcsstr.go` and `dispatch_token_sort_ratio.go`.
- `tests/bdd/features/token_jaccard.feature` — Godog feature file. Hashed-line header citing Jaccard 1912 as primary source AND noting the set-vs-multiset distinction with Q-Gram Jaccard. Scenarios: Canonical reference vectors (5-row Examples table), identical strings score 1.0, both-empty STANDARD convention (1.0), one-empty (0.0), symmetric across argument order, **set semantics deduplicate repeated tokens (distinct from Q-Gram Jaccard multiset — load-bearing keystone scenario)**. Every scenario tagged `@token @token-jaccard`.
- `testdata/golden/_staging/token_jaccard.json` — 10 entries: RV-TJ1..RV-TJ6 + both-empty + one-empty-A + one-empty-B + token-reorder. All in the canonical shape `{"name": "TokenJaccard_<id>", "algorithm": "TokenJaccard", "a": "...", "b": "...", "expected_score": ...}`. Staged for plan 06-06 merge into `testdata/golden/algorithms.json`.

### Modified

- `algoid_test.go` — added `int(fuzzymatch.AlgoTokenJaccard): true` to `TestDispatch_UnregisteredSlotsAreNil`'s `registered` map; updated the surrounding comments to mention plan 06-04 and the slot-17 registration milestone.
- `props_test.go` — appended 6 `TestProp_TokenJaccardScore_*` property tests (RangeBounds, Identity, Symmetric, NoNaN, NoInf, NoNegativeZero) using `testing/quick`. The Identity property is stronger than Q-Gram Jaccard's identity (which skips empty input) because TokenJaccard's identity short-circuit covers every input including the empty case per the LOCKED both-empty STANDARD catalogue convention.
- `example_test.go` — appended `ExampleTokenJaccardScore` (canonical partial-overlap pair `("alpha beta gamma", "beta gamma delta") → 0.5000`) with godoc documenting the set-vs-multiset distinction and the standard both-empty convention.
- `tests/bdd/steps/algorithms_steps.go` — appended TokenJaccard step methods (`iComputeTheTokenJaccardScoreBetween`, `iComputeTheSecondTokenJaccardScoreBetween`, `bothTokenJaccardScoresShouldBeEqual`) and their three `ctx.Step` registrations in `InitializeScenario` after the PartialRatioRunes block.
- `llms.txt` — appended `### TokenJaccard (Jaccard 1912 set-Jaccard over Tokenise output — distinct from Q-Gram Jaccard multiset)` section before Normalisation.
- `llms-full.txt` — appended `### Phase 6 algorithm surface (token tier — TokenJaccard)` block before Normalisation. Documents the set-vs-multiset distinction (with the keystone RV-TJ3 inline), the both-empty STANDARD convention LOCKED, the DET-03 integer-counter intersection cardinality LOCKED, the hand-derived reference vectors LOCKED, and includes the `TokenJaccardScore(a, b string) float64` godoc with two reference vectors (RV-TJ1 and the RV-TJ3 keystone).

## Decisions Made

**The four required `LOCKED` decisions are recorded verbatim:**

- **Set vs Multiset LOCKED 2026-05-15** — TokenJaccard uses SET semantics (`map[string]struct{}`, multiplicity collapsed). DISTINCT from Phase 5's Q-Gram Jaccard which uses MULTISET semantics over q-gram counts. The semantic divergence is intentional per `06-RESEARCH.md` Pattern 8: a token appearing twice in the input doesn't make it "more present"; q-gram presence is a multiplicity signal because longer overlapping runs of identical q-grams indicate stronger similarity. The KEYSTONE regression gate `RV-TJ3 ("a a b" vs "a b" → 1.0)` is pinned in four surfaces (algorithm godoc, unit test `TestTokenJaccardScore_SetVsMultisetDistinction`, BDD scenario "set semantics deduplicate repeated tokens", staging-golden entry); the same inputs under multiset Jaccard would yield 2/3 ≈ 0.667.
- **DET-03 intersection-cardinality LOCKED 2026-05-15** — `|A ∩ B|` is computed via integer-counter loop over the SMALLER set — `for k := range small { if _, ok := large[k]; ok { intersection++ } }`. The output is a scalar `int` (associative integer addition; map iteration order does not affect the result); no slice is ever built from set iteration on the output path. The union is computed via set inclusion-exclusion: `union := len(setA) + len(setB) - intersection`. DET-03 satisfied without any sort.
- **Both-empty STANDARD convention LOCKED 2026-05-15** — `TokenJaccardScore("", "")` returns 1.0 (catalogue STANDARD both-empty identity convention). DOES NOT deviate like TokenSetRatio (which returns 0.0 per the LOCKED RapidFuzz issue #110 bug-for-bug compatibility). The identity short-circuit `if a == b { return 1.0 }` covers the both-empty case directly; a post-Tokenise both-empty guard handles pure-separator inputs (e.g. `("  ", "___")`) where the raw strings differ but Tokenise produces empty slices on both sides.
- **Hand-derived reference vectors LOCKED 2026-05-15** — six hand-derived `(a, b)` pairs RV-TJ1..RV-TJ6 per `06-CONTEXT.md §1b` LOCKED. NO RapidFuzz cross-validation: the RapidFuzz cross-validation corpus at `testdata/cross-validation/token-ratios/vectors.json` does NOT include TokenJaccard entries because the set-Jaccard formula is unambiguous from Jaccard 1912 and the integer-counter intersection cardinality has no implementation choices left to ambiguate. Each test case carries the formula derivation in the test comment per algorithm-correctness-standards "Reference Vectors". `token_ratio_cross_validation_test.go` is unchanged by this plan.

### Reference vector derivations

The six hand-derived `RV-TJ1..RV-TJ6` reference vectors (set-Jaccard `J(A, B) = |∩| / |∪|`):

| ID | Inputs | Set A | Set B | `|∩|` | `|∪|` | Score |
|----|--------|-------|-------|------:|------:|------:|
| RV-TJ1 | `("a b c", "b c d")` | `{a,b,c}` | `{b,c,d}` | 2 (`{b,c}`) | 4 | `0.5000` (exact) |
| RV-TJ2 | `("a b", "a b c")` | `{a,b}` | `{a,b,c}` | 2 (`{a,b}`) | 3 | `0.6666666666666666` (= 2/3 within ε=1e-9) |
| RV-TJ3 | `("a a b", "a b")` (KEYSTONE) | `dedup({a,a,b}) = {a,b}` | `{a,b}` | 2 | 2 | `1.0` (exact) — MULTISET would yield 2/3 |
| RV-TJ4 | `("a b c", "x y z")` | `{a,b,c}` | `{x,y,z}` | 0 (disjoint) | 6 | `0.0` (exact) |
| RV-TJ5 | `("a b c", "a b c")` | covered by `a == b` short-circuit | — | — | — | `1.0` (exact) |
| RV-TJ6 | `("alpha beta gamma delta", "alpha beta epsilon zeta")` | `{alpha,beta,gamma,delta}` | `{alpha,beta,epsilon,zeta}` | 2 (`{alpha,beta}`) | 6 | `0.3333333333333333` (= 1/3 within ε=1e-9) |

**Staging-golden reference vector numbers** (`testdata/golden/_staging/token_jaccard.json` — 10 entries, all locked):

| Name | Score | Notes |
|------|------:|-------|
| `TokenJaccard_RV-TJ1_partial_overlap` | 0.5 | RV-TJ1 |
| `TokenJaccard_RV-TJ2_subset` | 0.6666666666666666 | RV-TJ2 |
| `TokenJaccard_RV-TJ3_set_vs_multiset_keystone` | 1.0 | KEYSTONE |
| `TokenJaccard_RV-TJ4_disjoint` | 0.0 | RV-TJ4 |
| `TokenJaccard_RV-TJ5_identity` | 1.0 | RV-TJ5 |
| `TokenJaccard_RV-TJ6_partial_overlap_greek` | 0.3333333333333333 | RV-TJ6 |
| `TokenJaccard_both_empty` | 1.0 | STANDARD convention (NOT TokenSetRatio 0.0 deviation) |
| `TokenJaccard_one_empty_a` | 0.0 | post-Tokenise one-empty guard |
| `TokenJaccard_one_empty_b` | 0.0 | post-Tokenise one-empty guard |
| `TokenJaccard_token_reorder` | 1.0 | sets identical after dedup |

## Deviations from Plan

### Auto-fixed Issues

**1. [Rule 3 - Blocking] Update `TestDispatch_UnregisteredSlotsAreNil` to include `AlgoTokenJaccard`**
- **Found during:** Task 1 (TokenJaccard dispatch registration)
- **Issue:** `algoid_test.go::TestDispatch_UnregisteredSlotsAreNil` enumerates all registered dispatch slots and asserts every other slot is `nil`. Registering `dispatch[AlgoTokenJaccard]` in `dispatch_token_jaccard.go` made slot 17 non-nil and broke this test.
- **Fix:** Added `int(fuzzymatch.AlgoTokenJaccard): true` to the `registered` map; updated the surrounding comments to mention plan 06-04 and the slot-17 registration milestone. Mirrors the same auto-fix from plan 06-01 (slot 14) and plans 06-02 / 06-03.
- **Files modified:** `algoid_test.go`
- **Verification:** `go test -run TestDispatch_UnregisteredSlotsAreNil -count=1 ./...` passes after the update.
- **Committed in:** `e6a4a97` (Task 1 commit — bundled with the dispatch wiring it accompanies, matching the same Rule-3 bundling pattern from plans 06-01 / 06-02 / 06-03)

**2. [Rule 3 - Blocking] Reduce `TokenJaccardScore` cyclomatic complexity via helper extraction**
- **Found during:** Task 1 final lint gate (`make lint`)
- **Issue:** `golangci-lint`'s `gocyclo` rule flagged `TokenJaccordScore`'s cyclomatic complexity at 12 (budget 10). The control flow combined the identity short-circuit, both-empty guard, one-empty guard, two set-construction loops, smaller-side selection swap, intersection-counter loop, union calculation, defensive zero-union guard, and the final division.
- **Fix:** Extracted two unexported helpers in the same file — `tokensToSet(tokens []string) map[string]struct{}` (builds a set from a token slice) and `setIntersectionCardinality(setA, setB map[string]struct{}) int` (the DET-03-satisfying integer-counter scan over the smaller set). The main `TokenJaccardScore` function drops below the gocyclo budget; both helpers carry godoc explaining their respective DET-03 invariants. No behaviour change.
- **Files modified:** `token_jaccard.go`
- **Verification:** `make lint` passes; `go test -run 'TestTokenJaccard|TestProp_TokenJaccard' -race -shuffle=on -count=1 ./...` still passes.
- **Committed in:** `e6a4a97` (Task 1 commit — bundled with the algorithm it factors)

---

**Total deviations:** 2 auto-fixed (both Rule 3 — blocking fixes required for CI green: test-fixture sync and lint compliance via helper extraction)
**Impact on plan:** Both auto-fixes are mandatory for the build to pass. No scope creep — both deviations are well within Rule 3's intent (essential corrections to keep CI green).

## Issues Encountered

None requiring problem-solving. The plan was extremely explicit (`<read_first>` blocks pointed at the canonical templates; `<action>` blocks named every file and the structure to copy). Implementation tracked the plan within minutes per file. The gocyclo-driven helper extraction is a tidy outcome — the helpers carry their own godoc and improve readability rather than obscuring the algorithm.

A noted observation rather than an issue: the alloc-budget ceiling in `TestTokenJaccardScore_AllocsBudget` is set at 8 (not the ≤ 4 baseline from the plan's `<action>` block) because the realistic per-token-string allocation count for short identifier-style inputs sits around 6-7 on darwin/arm64 with Tokenise's lowercase fold path. Mirrors the RESEARCH.md §4.1 acknowledged-ceiling pattern used by `qgram_jaccard_test.go::TestQGramJaccard_AllocsBudget` (ceiling 6 with baseline 4). The 4-allocation floor remains the canonical-source ideal.

## User Setup Required

None — TokenJaccard is a pure-function score over `Tokenise` output with no external service, no Python toolchain, and no configuration. Hand-derived reference vectors are reviewer-verifiable from Jaccard 1912 p. 43 in under a minute per row.

## Next Phase Readiness

- **Plan 06-05 (MongeElkan)** is unblocked — Monge-Elkan inherits the same hand-derived RV discipline per CONTEXT §1b LOCKED and consumes the existing dispatch infrastructure. Monge-Elkan ships its own inner-metric dispatch (see CONTEXT §3 / §4 LOCKED) and is independent of TokenJaccard's set-Jaccard surface.
- **Plan 06-06 finalisation** has a 10-entry staging-golden `testdata/golden/_staging/token_jaccard.json` to merge into `algorithms.json`. The merge follows the same pattern as plans 06-01 / 06-02 / 06-03's staging-goldens.

No blockers or concerns.

### Deferred items for plan 06-06

- **Final `bench.txt` numbers** — TokenJaccard benchmarks compile and run (ASCII Short / Medium / Long + Unicode Short). The project-wide `bench.txt` baseline is regenerated phase-by-phase at finalisation time, not per-plan. Plan 06-06 will run the full benchmark suite and commit the updated `bench.txt`.
- **`testdata/golden/_staging/token_jaccard.json` merge** into `testdata/golden/algorithms.json` — staged in the `_staging/` directory; plan 06-06 finalisation handles the merge (mirrors Phase 4 / Phase 5 / plan 06-01 / 06-02 / 06-03 finalisation flow).
- **Cross-platform determinism golden update** — TokenJaccardScore on ASCII inputs is deterministic across the four CI platforms by construction (integer-derived single division on bounded set cardinalities), but `verify-determinism`'s golden file does not yet include any TokenJaccard entries. Plan 06-06 will add representative TokenJaccard entries to the golden file as part of the phase-wide determinism gate refresh.

## Self-Check: PASSED

- `token_jaccard.go` — FOUND
- `token_jaccard_test.go` — FOUND
- `token_jaccard_bench_test.go` — FOUND
- `token_jaccard_fuzz_test.go` — FOUND
- `dispatch_token_jaccard.go` — FOUND
- `tests/bdd/features/token_jaccard.feature` — FOUND
- `testdata/golden/_staging/token_jaccard.json` — FOUND (10 entries with RV-TJ1..RV-TJ6 + both-empty / one-empty / token-reorder)
- `algoid_test.go` — MODIFIED (AlgoTokenJaccard added to registered map)
- `props_test.go` — MODIFIED (6 TokenJaccard property tests appended)
- `example_test.go` — MODIFIED (ExampleTokenJaccardScore appended)
- `tests/bdd/steps/algorithms_steps.go` — MODIFIED (TokenJaccard step methods + InitializeScenario registrations)
- `llms.txt` — MODIFIED (TokenJaccard section added)
- `llms-full.txt` — MODIFIED (Phase 6 algorithm surface block for TokenJaccard added)
- Commit `e6a4a97` — FOUND in `git log`

All claimed deliverables verified by `git log --oneline -5` and `ls`-equivalent file existence on disk.

---
*Phase: 06-token-based-algorithms*
*Completed: 2026-05-15*
