---
phase: 05-q-gram-algorithms
plan: 04
subsystem: algorithms-q-gram
tags: [tversky, tversky-1977, asymmetric-similarity, alpha-beta-parameters, jaccard-fallback-dispatch, byte-and-rune-paths, dispatch-registration, property-tests-asymmetry, fuzz, benchmark, bdd-asymmetry-gate, staging-golden, llms-sync]

# Dependency graph
requires:
  - phase: 05-q-gram-algorithms
    provides: shared q-gram extraction (extractQGrams + extractQGramsRunes from plan 05-01); pre-declared ErrInvalidTverskyParam sentinel from plan 05-01; AlgoTversky enum slot from Phase 1
provides:
  - TverskyScore + TverskyScoreRunes — two new exported byte+rune asymmetric-similarity functions
  - dispatch[AlgoTversky] wired with α=β=1.0 (Jaccard-equivalent) wrapper per CONTEXT.md "Claude's Discretion"
  - Three-layer asymmetry direction-sensitivity gate (unit + property + BDD) — load-bearing regression detector for parameter-order correctness
  - Bit-exact algebraic cross-checks against QGramJaccardScore (α=β=1.0) and SorensenDiceScore (α=β=0.5) at both unit-test and property-test levels
  - Staging golden file testdata/golden/_staging/tversky.json with both RV-T1 and RV-T2 as separate rows (asymmetry gate at golden-file level)
  - Fuzz harnesses FuzzTverskyScore + FuzzTverskyScoreRunes with float64 α/β coerced into [0,1] with α+β>0 enforced
affects: [phase-05 finalisation (05-05 owns merge of staging golden into algorithms.json), Phase 8 Scorer (WithTverskyAlgorithm consumes TverskyScore directly with user-supplied α/β), Phase 9 scan + Phase 10 Extract via dispatch table]

# Tech tracking
tech-stack:
  added: []  # No new dependencies; uses stdlib + existing q_gram.go + errors.go
  patterns:
    - Asymmetric-similarity discipline — per-side residual weights (α weighs first arg's |A−B|, β weighs second arg's |B−A|) with three-layer regression defence
    - Algebraic-equivalence cross-checks — bit-exact (math.Float64bits) equality with Jaccard at α=β=1 and Dice at α=β=0.5, pinned in BOTH unit and property tests
    - Two-panic-path direct-call discipline — n<1 panics with "invalid q-gram size"; α<0/β<0/α+β==0 panics with "invalid tversky parameter" (per CONTEXT.md §5 LOCKED)
    - Float-parameter coercion in fuzz body — squashing function `|x|/(|x|+1)` maps R → [0,1) bounded; combined-zero (α+β=0) coerced to (1.0, 0.0) to keep panic paths unreachable in the harness
    - 0/0 guard on partial-empty multisets — when intersection==0 short-circuit to 0.0 (matches QGramJaccardScore's partial-empty convention)

key-files:
  created:
    - tversky.go
    - dispatch_tversky.go
    - tversky_test.go
    - tversky_bench_test.go
    - tversky_fuzz_test.go
    - testdata/golden/_staging/tversky.json
    - testdata/fuzz/FuzzTverskyScore/seed-001
    - testdata/fuzz/FuzzTverskyScoreRunes/seed-001
    - tests/bdd/features/tversky.feature
  modified:
    - props_test.go
    - algorithms_golden_test.go
    - algoid_test.go
    - example_test.go
    - tests/bdd/steps/algorithms_steps.go
    - llms.txt
    - llms-full.txt

key-decisions:
  - "OQ-RT1 RESOLVED 2026-05-15: dispatch[AlgoTversky] wrapper binds α=β=1.0 (Jaccard-equivalent) — CONTEXT.md \"Claude's Discretion\" recommendation endorsed; the equivalence is RV-T3-verified bit-for-bit"
  - "OQ-RT2 RESOLVED 2026-05-15: α+β>0 implemented as `α==0 && β==0` to dodge any float-comparison anxiety on slightly-negative sums"
  - "OQ-RT3 RESOLVED 2026-05-15: 0/0 NaN bug (Rule 1 deviation, found by fuzz) fixed by short-circuiting to 0.0 when intersection==0 — matches QGramJaccardScore's partial-empty convention"

patterns-established:
  - "Asymmetric q-gram metric — α/β weights with per-side residual coverage; structurally distinct from the symmetric Jaccard/Dice/Cosine pattern"
  - "Three-layer asymmetry regression gate — unit (TestTversky_AsymmetryDirectionSensitive), property (AsymmetricWhenAlphaNeqBeta + ParameterSwapSymmetry), BDD (Asymmetry direction-sensitivity scenario)"
  - "Bit-exact algebraic cross-check via math.Float64bits — Tversky degenerates to Jaccard at α=β=1 and Dice at α=β=0.5 EXACTLY"
  - "float64 fuzz seed literal format — testdata/fuzz/FuzzTverskyScore/seed-001 uses `float64(0.8)\\nfloat64(0.2)` lines after the int(n) line"

requirements-completed: [QGRAM-05]

# Metrics
duration: 31min
completed: 2026-05-15
---

# Phase 5 Plan 04: Tversky asymmetric similarity (QGRAM-05) Summary

**Tversky asymmetric q-gram similarity with α/β weighted residuals (Tversky 1977 §2 eq. (1) p.332), plus three-layer asymmetry direction-sensitivity gate (unit + property + BDD) and bit-exact algebraic cross-checks against Q-Gram Jaccard (α=β=1) and Sørensen-Dice (α=β=0.5)**

## Performance

- **Duration:** 31 min
- **Started:** 2026-05-15T05:44:57Z
- **Completed:** 2026-05-15T06:16:07Z
- **Tasks:** 3
- **Files modified:** 16 (9 created, 7 modified)
- **Commits:** 3 atomic per-task commits

## Accomplishments

- **TverskyScore + TverskyScoreRunes** — two new exported asymmetric-similarity functions; T(A, B, α, β) = |A∩B| / (|A∩B| + α·|A−B| + β·|B−A|).
- **Dispatch slot 12 (AlgoTversky)** wired with α=β=1.0 (Jaccard-equivalent) wrapper; the asymmetric use case is reserved for Phase 8 Scorer's WithTverskyAlgorithm(weight, α, β).
- **Load-bearing asymmetry gate at three layers**:
  - **Unit:** TestTversky_AsymmetryDirectionSensitive asserts |RV-T1 − RV-T2| > 0.1 (actual difference ≈ 0.2302) plus inline endpoint pins on each value.
  - **Property:** TestProp_TverskyScore_AsymmetricWhenAlphaNeqBeta uses an implication structure (length-mismatched + non-zero score → asymmetric output) plus a spot-check on the canonical RV-T1/RV-T2 pair.
  - **BDD:** tests/bdd/features/tversky.feature `Asymmetry direction-sensitivity gate` scenario asserts the two scores differ by more than 0.1.
- **Bit-exact algebraic cross-checks** (via math.Float64bits) against Q-Gram Jaccard (α=β=1.0) and Sørensen-Dice (α=β=0.5) at both unit-test and property-test levels — eight pairs covering ASCII, orthogonal, and rune surfaces.
- **Parameter-swap symmetry** T(a, b, α, β) = T(b, a, β, α) pinned in unit + property + BDD — the algebraic identity that proves asymmetry is a consequence of α≠β (rather than a one-sided coding bug).
- **Two-panic-path direct-call discipline** with seven exhaustively-tested invalid configurations (n=0, n=-1, n=-100, n=MinInt32; α<0; β<0; α+β=0) plus the boundary-valid case (α=0 with β>0 must NOT panic).
- **Staging golden** testdata/golden/_staging/tversky.json with 8 alphabetically-sorted entries including both RV-T1 (Tversky_abcd_abcdef_asym) and RV-T2 (Tversky_abcdef_abcd_asym_swap) as separate rows — the asymmetry gate at the golden-file level.
- **Property-test surface** of 21 tests (5 standard invariants × 2 surfaces + 5 Tversky-specific × 2 surfaces + 1 determinism) covering the full algebraic structure.
- **Fuzz harnesses** FuzzTverskyScore + FuzzTverskyScoreRunes — passed 10s and 15s smoke runs without finding additional issues after the Rule-1 fix; α/β coerced via the squashing function `|x|/(|x|+1)`.

## Task Commits

Each task was committed atomically:

1. **Task 1: tversky.go + dispatch + unit tests + staging golden + example + llms** — `b5dadc4` (feat)
2. **Task 2: property tests + benchmarks + fuzz harnesses + Rule-1 0/0 NaN fix** — `6f8bb72` (test)
3. **Task 3: BDD feature + steps with asymmetry direction-sensitivity scenario** — `0fe70a2` (test)

_Note: This plan is `type=execute` (not full TDD-gate); the orchestrator's plan-level RED/GREEN/REFACTOR doesn't apply. Each task carries `tdd="true"` at the per-task level — the unit tests in Task 1 were authored alongside the implementation; Tasks 2 and 3 layer property/fuzz/BDD coverage on top._

## Files Created/Modified

### Created (9 files)

- `tversky.go` — TverskyScore + TverskyScoreRunes public functions + tverskyFromQGramMaps shared helper. 367 lines including the file-level godoc with full source citation, formula derivation, asymmetry-discriminating reference vector pair (RV-T1 vs RV-T2), and the load-bearing regression gate explanation.
- `dispatch_tversky.go` — Package-load-time registration of the Jaccard-fallback wrapper (α=β=1.0, n=3) into dispatch[AlgoTversky]. NO init().
- `tversky_test.go` — Unit tests: BothEmpty, OneEmpty, Identical (× multiple α/β configs), ReferenceVectors (RV-T1..RV-T4), AsymmetryDirectionSensitive (load-bearing), JaccardCrossCheck (bit-exact), DiceCrossCheck (bit-exact), ParameterSwapSymmetry, SymmetricWhenAlphaEqBeta, TverskyRunes_CafeReference, PanicsOnInvalidN, PanicsOnInvalidParams, ZeroAlphaWithPositiveBeta, DispatchEqualsJaccard, AllocsBudget. 591 lines.
- `tversky_bench_test.go` — 4 benchmarks (3 byte + 1 rune) with α=0.8/β=0.2 (asymmetric configuration); alloc-asserted via b.ReportAllocs().
- `tversky_fuzz_test.go` — 2 fuzz harnesses (byte + rune) with α/β coerced via the squashing function; 16 programmatic seeds covering the full reference-vector slate plus boundary α/β values.
- `testdata/golden/_staging/tversky.json` — 8 alphabetically-sorted entries; both asymmetry-pair rows present.
- `testdata/fuzz/FuzzTverskyScore/seed-001` — Canonical RV-T1 seed in `go test fuzz v1` literal format with float64(α) and float64(β) lines.
- `testdata/fuzz/FuzzTverskyScoreRunes/seed-001` — Same as byte path (RV-T1 inputs are ASCII, so byte and rune paths produce equivalent multisets).
- `tests/bdd/features/tversky.feature` — Gherkin feature with 8 scenarios including the LOAD-BEARING Asymmetry direction-sensitivity gate scenario and the Parameter-swap symmetry algebraic identity scenario.

### Modified (7 files)

- `props_test.go` — Appended Tversky property-test block: standard 6 invariants × 2 surfaces (RangeBounds, Identity, NoNaN, NoInf, NoNegativeZero) + 5 Tversky-specific × 2 surfaces (SymmetricWhenAlphaEqBeta, AsymmetricWhenAlphaNeqBeta, ParameterSwapSymmetry, JaccardCrossCheck, DiceCrossCheck) + 1 determinism test = 21 property tests total.
- `algorithms_golden_test.go` — Appended buildTverskyStagingEntries + TestGolden_Tversky_Staging.
- `algoid_test.go` — Appended TestDispatch_TverskyRegistered; updated TestDispatch_UnregisteredSlotsAreNil to flip slot 12 to registered.
- `example_test.go` — Appended ExampleTverskyScore (with both symmetric AND asymmetric subcases per RESEARCH.md OQ-4) + ExampleTverskyScoreRunes.
- `tests/bdd/steps/algorithms_steps.go` — Appended 5 step methods (iComputeTheTverskyScoreBetweenWithNAlphaBeta, iComputeTheSecondTverskyScoreBetweenWithNAlphaBeta, iComputeTheTverskyRunesScoreBetweenWithNAlphaBeta, theTwoTverskyScoresShouldDifferByMoreThan, bothTverskyScoresShouldBeEqual) + their ctx.Step regex registrations.
- `llms.txt` — Added the Tversky function block (TverskyScore + TverskyScoreRunes).
- `llms-full.txt` — Added the parallel Phase 5 algorithm surface block with one-line rationales noting the α/β asymmetry surface and the dispatch-table Jaccard-fallback compromise.

## Decisions Made

- **Dispatch wrapper binds α=β=1.0 (Jaccard-equivalent)** rather than α=β=0.5 (Dice-equivalent). The Jaccard equivalence is the more widely-known reference algorithm and is verified by RV-T3 in tversky_test.go — TverskyScore("abcd", "abce", 2, 1.0, 1.0) == QGramJaccardScore("abcd", "abce", 2) bit-for-bit. This makes the dispatch-table output for AlgoTversky and AlgoQGramJaccard identical at the default n=3, which is a reviewer-friendly auditability property.
- **α+β>0 implemented as `α==0 && β==0`** rather than literal `α+β==0` evaluation. Same logical effect on non-negative inputs (which are the only inputs reaching this gate, since α<0 and β<0 are caught explicitly above), but the conjunction form dodges any float-comparison anxiety about associativity of α+β being slightly less than 0.
- **Asymmetry-conditional property test uses an approximate implication via input length comparison.** The premise "|A−B| ≠ |B−A|" is impossible to compute cheaply inside the closure without re-implementing extraction; substitute "len(a) ≠ len(b) AND both ≥ n AND fwd > 0" as a tractable approximation that captures the asymmetry-exercising cases. The unit test TestTversky_AsymmetryDirectionSensitive pins the LOAD-BEARING RV-T1 vs RV-T2 inequality at known input pairs as the deterministic complement.
- **Tolerance epsilon set at 1e-15** (per the plan's locked convention) rather than the 1e-9 used by Q-Gram Jaccard / Sørensen-Dice. Tversky's actual numerical accuracy is far higher than 1e-9 for the integer-derived computations involved; the tighter tolerance documents that fact and would surface any future reduction-order regression that introduced more than 1 ULP of error.

## Deviations from Plan

### Auto-fixed Issues

**1. [Rule 1 - Bug] Fix 0/0 NaN in tverskyFromQGramMaps when one input's q-gram extraction yields an empty multiset AND the only nonzero residual is α/β-zeroed**

- **Found during:** Task 2 (fuzz harness execution — first run produced a deterministic failure within ~90 milliseconds).
- **Issue:** The input `("0000100", "0", n=4, α=0, β=0.5)` exposed a 0/0 NaN return. Walkthrough: qb is empty (`len("0") < 4`), qa has 4 trigrams, intersection=0, aMinusB=4, bMinusA=0; the original denominator collapsed to `0·4 + 0.5·0 = 0` → `0/0 = NaN`. The full-both-empty short-circuit didn't fire (qa is non-empty); the intersection-positive division branch wasn't gated for the partial-empty case where the residual product collapses.
- **Fix:** Added an `if intersection == 0 { return 0.0 }` short-circuit in tverskyFromQGramMaps before the division. The semantic justification: zero intersection means the two inputs share no q-grams, which is the "no match" case; QGramJaccardScore("0000100", "0", 4) returns 0/4 = 0.0 on the same input, and Tversky's partial-empty convention should mirror that.
- **Files modified:** `tversky.go` (one new short-circuit branch in tverskyFromQGramMaps; ~10 lines including explanatory comment).
- **Verification:** Re-ran the failing fuzz input (`go test -run='FuzzTverskyScore/a0438cc1394d07f5'` → PASS); deleted the spurious fuzz finding file from testdata/fuzz/FuzzTverskyScore/. Re-ran the full test suite (all green). Property tests _NoNaN now hold over the broader random surface; bench output and staging golden file remain byte-identical (the existing entries don't exercise the partial-empty case).
- **Committed in:** `6f8bb72` (Task 2 commit — alongside the property/bench/fuzz files; the fix and the harness that found it ship together for review clarity).

---

**Total deviations:** 1 auto-fixed (Rule 1 — bug)
**Impact on plan:** The deviation was discovered by the same fuzz harness the plan prescribed (Task 2), within ~90ms of the first run. The fix is self-contained in tverskyFromQGramMaps and does not alter the algebraic semantics for any non-degenerate input. No scope creep — the fix is essential for correctness on the partial-empty-multiset case.

## Issues Encountered

- **Float64 fuzz seed literal format** was undocumented in any existing seed file in the project (the q-gram-tier predecessors all used `(string, string, int)` triples). Spent ~5 minutes confirming via a sandbox experiment that the syntax is `float64(0.8)` on its own line — same lexical pattern as `string("...")` and `int(2)`. Once confirmed, the seed files were straightforward to author.
- **Asymmetry-conditional property test design** — the original draft of `TestProp_TverskyScore_AsymmetricWhenAlphaNeqBeta` had a degenerate `return true` body that would have slipped past `quick.Check` without exercising the asymmetry. Refactored to use an approximate implication (length-mismatched inputs with non-zero score → asymmetric output) plus a deterministic spot-check on the canonical RV-T1/RV-T2 pair. The combined structure ensures both random coverage AND deterministic regression detection.

## User Setup Required

None - no external service configuration required. The library remains pure-function with zero runtime dependencies.

## Threat Flags

| Flag | File | Description |
|------|------|-------------|
| (none) | — | No new security-relevant surface introduced. The added panic paths (n<1, α<0, β<0, α+β==0) are pre-declared in CONTEXT.md §5 LOCKED; the Phase 8 Scorer will translate them to ErrInvalidQGramSize / ErrInvalidTverskyParam. |

## Findings (per the plan's `<output>` request)

- **Asymmetry-gate test design is sound but the property test is necessarily approximate.** The unit-level RV-T1/RV-T2 pin is the deterministic regression detector; the property test is the broader generative companion. A future enhancement could compute the residuals precisely inside the closure (extracting the q-grams via the public API surface) but at non-trivial harness complexity cost. Defer unless reviewers surface a concern.
- **NaN-α/β fuzz hardening is recommended for v1.x.** The fuzz body squashes NaN α/β to 0.5 to keep the documented panic paths unreachable, but the production API does NOT explicitly handle NaN α/β — direct calls with NaN would propagate NaN through the multiplication and into the score, violating the _NoNaN property silently (since the unit-test discipline only exercises specific NaN scenarios). Recommend either: (a) extending the public-API gate to panic on NaN α/β with a documented message, OR (b) extending the fuzz body to exercise NaN inputs as a regression detector. Logged here as a future-hardening recommendation per the plan's `<output>` block.
- **The α+β==0 panic is implemented as `α==0 && β==0`** which is logically equivalent on non-negative inputs but defends against an unlikely edge case where α and β are tiny opposite-sign floats summing to zero (or near-zero) before the gate fires. Since the per-side `α<0`/`β<0` checks fire first, this defence is currently redundant — but the equivalent code is one line longer for zero benefit, so the conjunction form is the simpler shape to read.

## Next Phase Readiness

- Plan 05-04 is complete and ready for plan 05-05 (Phase 5 finalisation) to consume.
- Plan 05-05 will:
  - Merge testdata/golden/_staging/tversky.json into testdata/golden/algorithms.json (alongside the qgram_jaccard, sorensen_dice, and cosine staging files).
  - Update TestGolden_Algorithms_Merge's stagingFiles slice to include the four new q-gram staging files.
  - Regenerate bench.txt to include the four new q-gram benchmark families (Tversky's 4 benches plus Q-Gram Jaccard's, Sørensen-Dice's, and Cosine's).
  - Extend the identifier-similarity example with the four new q-gram columns.
- Phase 8 (Scorer) consumes TverskyScore directly via WithTverskyAlgorithm(weight, α, β) — the Phase 8 plan should reference the ErrInvalidTverskyParam sentinel (already declared per plan 05-01) for option-construction validation.

## Self-Check: PASSED

**Files verified to exist:**
- `tversky.go`: FOUND
- `dispatch_tversky.go`: FOUND
- `tversky_test.go`: FOUND
- `tversky_bench_test.go`: FOUND
- `tversky_fuzz_test.go`: FOUND
- `testdata/golden/_staging/tversky.json`: FOUND
- `testdata/fuzz/FuzzTverskyScore/seed-001`: FOUND
- `testdata/fuzz/FuzzTverskyScoreRunes/seed-001`: FOUND
- `tests/bdd/features/tversky.feature`: FOUND

**Commits verified to exist:**
- `b5dadc4`: FOUND (Task 1 — feat: implement Tversky asymmetric similarity)
- `6f8bb72`: FOUND (Task 2 — test: property/bench/fuzz + fix 0/0 in partial-empty case)
- `0fe70a2`: FOUND (Task 3 — test: BDD feature with asymmetry direction-sensitivity gate)

**Verification commands (re-run before SUMMARY commit):**
- `go build ./...` → exits 0
- `go test -run 'TestTversky|TestProp_Tversky|TestDispatch_TverskyRegistered|TestGolden_Tversky_Staging|ExampleTversky' ./...` → ok
- `go test -bench='BenchmarkTversky' -benchmem -benchtime=1x -run=NONE ./...` → 4 benches; ≤ 6 allocs/op on all
- `go test -fuzz=FuzzTverskyScore$ -fuzztime=15s -run=NONE ./...` → PASS
- `go test -fuzz=FuzzTverskyScoreRunes$ -fuzztime=15s -run=NONE ./...` → PASS
- `make test-bdd` → ok (120 scenarios; 8 new Tversky scenarios all green)
- `bash scripts/verify-license-headers.sh` → OK: 102 .go files carry the Apache-2.0 header
- `bash scripts/verify-no-runtime-deps.sh` → OK: root go.mod allowlist clean
- `! grep -q "^func init" tversky.go` → ok (no init)
- `grep -q "Source: Tversky, A. (1977)" tversky.go` → ok (citation present)
- `grep -q "dispatch\[AlgoTversky\]" dispatch_tversky.go` → ok (dispatch wired)
- `grep -q "invalid tversky parameter" tversky.go` → ok (panic message present)
- `grep -q "RV-T1" tversky_test.go && grep -q "RV-T2" tversky_test.go && grep -q "AsymmetryDirectionSensitive" tversky_test.go` → ok (asymmetry gate test present)
- `grep -v '^//' tversky.go | grep -E "math\\.(Pow|Log|Exp|FMA)"` → no match (DET-06 gate)

---
*Phase: 05-q-gram-algorithms*
*Completed: 2026-05-15*
