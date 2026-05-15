---
phase: 07-phonetic-algorithms
plan: "04"
subsystem: phonetic-algorithms
status: COMPLETE
tags: [phonetic, mra, nbs-tech-note-943, threshold-table, mracompare-bool-int, phase-7]
closes: [PHON-04]
dependency_graph:
  requires: [07-01-SUMMARY, 07-02-SUMMARY, 07-03-SUMMARY]
  provides: [MRACode, MRACompare, MRAScore, dispatch-AlgoMRA, permittedMongeElkanInner-18-FINAL]
  affects: [monge_elkan.go, algoid_test.go, phonetic_cross_validation_test.go, phonetic_codes_golden_test.go, props_test.go, example_test.go, llms.txt, llms-full.txt]
tech_stack:
  added: [mra.go, dispatch_mra.go, mra_test.go, mra_bench_test.go, mra_fuzz_test.go, tests/bdd/features/mra.feature]
  patterns: [package-level-var-threshold-table, mracompare-bool-int-non-float64, identity-short-circuit, length-diff-gate]
key_files:
  created:
    - mra.go
    - dispatch_mra.go
    - mra_test.go
    - mra_bench_test.go
    - mra_fuzz_test.go
    - testdata/golden/_staging/mra.json
    - tests/bdd/features/mra.feature
  modified:
    - monge_elkan.go
    - monge_elkan_test.go
    - algoid_test.go
    - phonetic_cross_validation_test.go
    - phonetic_codes_golden_test.go
    - props_test.go
    - example_test.go
    - export_test.go
    - testdata/cross-validation/phonetic/vectors.json
    - testdata/golden/phonetic-codes.json
    - tests/bdd/steps/algorithms_steps.go
    - llms.txt
    - llms-full.txt
    - double_metaphone.go (gofmt cleanup only)
    - double_metaphone_fuzz_test.go (gofmt cleanup only)
    - double_metaphone_test.go (gofmt cleanup only)
    - soundex_fuzz_test.go (gofmt cleanup only)
    - soundex_test.go (gofmt cleanup only)
decisions:
  - "MRACompare returns (false, 0) on length-diff >= 3 (NOT an error) per CONTEXT.md §6 LOCKED Open Question 2 resolution"
  - "var mraThresholdTable at package level with explicit sum>12 clamp comment per determinism-standards + RESEARCH.md Pitfall 7.C"
  - "TestMRAThresholdTable_Clamp and PropMRA_ThresholdMonotonic are load-bearing property tests for the sum>12 clamp"
  - "FINAL Phase 7 Monge-Elkan state: permittedMongeElkanInner = 18 entries, rejected slice = 5 entries"
metrics:
  duration: "15m21s"
  completed_date: "2026-05-15"
  tasks_completed: 1
  files_changed: 25
---

# Phase 7 Plan 04: MRA (Match Rating Approach) Summary

MRA algorithm fresh-implemented from NBS Tech Note 943 (Moore, Kuhns, Trefftzs, Montgomery 1977) with three public functions and the catalogue's only non-float64 return shape — MRACompare(a, b) (matched bool, simScore int).

## One-liner

MRA fresh-implemented from NBS Tech Note 943 with threshold table at package-level var, sum>12 clamp locked by property tests, and the catalogue's only (bool, int) public return shape via MRACompare.

## What Was Built

### Core Algorithm (`mra.go` — 230 LOC)

Three public surfaces per CONTEXT.md §6 LOCKED:

1. **`MRACode(s string) string`** — NBS-943 encoding: vowel-deletion (except leading), adjacent-consonant deduplication, first-3-last-3 truncation if pre-truncation length > 6. Output ≤ 6 uppercase ASCII letters.

2. **`MRACompare(a, b string) (matched bool, simScore int)`** — The ONLY public function in the fuzzymatch catalogue with a non-float64 return shape. `simScore` ∈ [0, 6] is the raw NBS Tech Note 943 similarity counter. Features: identity short-circuit, length-difference >= 3 auto-mismatch gate (returns (false, 0) per CONTEXT.md §6 LOCKED Open Question 2 resolution), Table A threshold lookup, L→R + R→L common-character elimination, similarity = 6 - max(unmatched_a, unmatched_b).

3. **`MRAScore(a, b string) float64`** — Binary 0.0/1.0 dispatch-table wrapper around MRACompare. Consistency invariant: `MRAScore(a, b) == 1.0 iff MRACompare(a, b).matched` (property-tested).

### Threshold Table (LOAD-BEARING)

`var mraThresholdTable [13]int` declared at package level (NOT inside init()) per determinism-standards SKILL.md. Explicit inline comment documents the `sum > 12 → 2` clamp per RESEARCH.md Pitfall 7.C:
- sum ≤ 4 → threshold 5
- 4 < sum ≤ 7 → threshold 4
- 7 < sum ≤ 11 → threshold 3
- sum >= 12 → threshold 2 (clamped)

### Dispatch (`dispatch_mra.go`)

`var _ = func() bool { dispatch[AlgoMRA] = MRAScore; return true }()` wires `MRAScore` into slot 26. `MRACode` and `MRACompare` are public but NOT dispatched (dispatch table is `(a, b string) float64` valued).

### Monge-Elkan (FINAL Phase 7 State)

- `permittedMongeElkanInner` in `monge_elkan.go`: 17 → **18 entries** (FINAL Phase 7 count)
- `rejected` slice in `monge_elkan_test.go`: 6 → **5 entries** (FINAL Phase 7 count)
- `TestMongeElkanScore_BinaryInner_MRA`: three sub-cases lock binary-inner-composition
- `algoid_test.go`: `registered` map updated with `AlgoMRA` (slot 21)

### Tests

- `mra_test.go`: 10 test functions including all RV-M1..RV-M12 reference vectors, TestMRAThresholdTable_Clamp (LOAD-BEARING for Pitfall 7.C), consistency pin tests
- `mra_bench_test.go`: 9 benchmarks with allocation budgets (MRACode < 500ns 1 alloc; MRACompare/MRAScore < 500ns ≤ 2 allocs)
- `mra_fuzz_test.go`: FuzzMRA with 4 MRA-specific invariants (charset, integer range, binary score, score-compare consistency); 10s fuzz run passed
- `props_test.go`: 8 new property tests (5-invariant block + PropMRA_CodeCharset + PropMRA_CompareIntegerRange + PropMRA_ScoreCompareConsistency + PropMRA_ThresholdMonotonic)
- `export_test.go`: `var MRAThresholdForTest = mraThreshold` for property tests

### BDD (`tests/bdd/features/mra.feature`)

9 scenarios covering: Byrne/Boern encoding, Smith/Smyth encoding, Smith/Smyth compare (matched=true sim=5), length-diff auto-mismatch, MRAScore Smith/Smyth, identity, both-empty, truncation gate, empty-input score.

### Cross-validation & Golden Files

- `testdata/cross-validation/phonetic/vectors.json`: 20 MRA entries (stub replaced)
- `testdata/golden/phonetic-codes.json`: 10 MRA entries appended
- `testdata/golden/_staging/mra.json`: 10 staging Score entries
- `phonetic_cross_validation_test.go`: MRA sub-test ENABLED (t.Skip replaced)
- `phonetic_codes_golden_test.go`: MRA branch ENABLED (t.Skip replaced)

### Examples (`example_test.go`)

Three new ExampleXxx demonstrating all three public surfaces:
- `ExampleMRACode`: Byrne → BYRN, Kathrynoglin → KTHGLN
- `ExampleMRACompare`: Smith/Smyth → (matched=true sim=5), both-empty → (matched=true sim=6) — demonstrates the (bool, int) shape
- `ExampleMRAScore`: Smith/Smyth → 1.0, Ad/ZachariahMontgomery → 0.0

### Documentation (`llms.txt`, `llms-full.txt`)

3 new exported symbols added in-plan:
- `func MRACode(s string) string`
- `func MRACompare(a, b string) (matched bool, simScore int)` (noted as only non-float64 return)
- `func MRAScore(a, b string) float64`

## Reference Vectors Verified

| Vector | Input | Expected | Result |
|--------|-------|----------|--------|
| RV-M1 | MRACode("Byrne") | "BYRN" | PASS |
| RV-M2 | MRACode("Boern") | "BRN" | PASS |
| RV-M3 | MRACode("Smith") | "SMTH" | PASS |
| RV-M4 | MRACode("Smyth") | "SMYTH" | PASS |
| RV-M5 | MRACode("Catherine") | "CTHRN" | PASS |
| RV-M6 | MRACode("Kathrynoglin") | "KTHGLN" | PASS (truncation gate) |
| RV-M7 | MRACompare("Smith", "Smyth") | (true, 5) | PASS |
| RV-M8 | MRACompare("Ad", "ZachariahMontgomery") | (false, 0) | PASS (length-diff gate) |
| RV-M11 | MRACompare("", "") | (true, 6) | PASS (both-empty) |
| RV-M12 | MRAScore("Smith", "Smyth") | 1.0 | PASS |

## Acceptance Criteria Status

- [x] `MRACode("Byrne") == "BYRN"`
- [x] `MRACode("Kathrynoglin") == "KTHGLN"` (first-3-last-3 truncation gate)
- [x] `MRACompare("Ad", "ZachariahMontgomery") == (false, 0)` (length-diff >= 3 auto-mismatch)
- [x] `MRACompare("", "") == (true, 6)` (both-empty)
- [x] `MRAScore(a, b) == 1.0 iff MRACompare(a, b).matched` (consistency invariant — property tested)
- [x] `var mraThresholdTable` at package level (NOT in init())
- [x] `mra.go` documents `sum > 12 → 2` clamp inline
- [x] `TestMRAThresholdTable_Clamp` passes (mraThreshold(13)/(15)/(20) all return 2)
- [x] `PropMRA_ThresholdMonotonic` passes for sumLen ∈ [0, 20]
- [x] `PropMRA_ScoreCompareConsistency` passes for arbitrary inputs
- [x] `permittedMongeElkanInner` has 18 entries (FINAL Phase 7)
- [x] `rejected` slice has 5 entries (FINAL Phase 7)
- [x] `TestMongeElkan_PanicsOnNonPermittedInner` passes
- [x] `TestMongeElkanScore_BinaryInner_MRA` passes three sub-tests
- [x] BDD `mra.feature` has 9 scenarios; `make test-bdd` passes
- [x] `TestPhonetic_CrossValidation/MRA` passes (20 entries)
- [x] `TestPhoneticCodesGolden/MRA` passes (10 entries)
- [x] `TestLLMsTxt` passes
- [x] 3 new ExampleXxx entries compile and produce expected output
- [x] FuzzMRA 10s run passes with all 4 MRA-specific invariants
- [x] BenchmarkMRACode < 500ns 1 alloc; BenchmarkMRACompare < 500ns ≤ 2 allocs

## Deviations from Plan

### Auto-fixed Issues

**1. [Rule 1 - Bug] Fixed incorrect test case for length-difference gate**
- **Found during:** Task 1 (TestMRACompare_LengthDifferenceAutoMismatch)
- **Issue:** Test used "A" vs "William" expecting (false, 0), but MRACode("William")="WLM" (len 3) and MRACode("A")="A" (len 1), giving diff=2 which is NOT >= 3 — not a gate-triggering pair
- **Fix:** Replaced with "A" vs "Kathrynoglin" (MRACode→"KTHGLN", len 6; diff=5 >= 3 ✓)
- **Files modified:** mra_test.go
- **Commit:** 862055f

**2. [Rule 3 - Blocking] algoid_test.go TestDispatch_UnregisteredSlotsAreNil failed**
- **Found during:** Task 1 (go test -race ./...)
- **Issue:** The test's `registered` map did not include `AlgoMRA` (slot 21), causing it to fail when dispatch_mra.go registered the slot
- **Fix:** Added `int(fuzzymatch.AlgoMRA): true` to the registered map and updated the comment
- **Files modified:** algoid_test.go
- **Commit:** 862055f

**3. [Rule 3 - Blocking] gofmt formatting violations after make fmt**
- **Found during:** Task 1 (make check)
- **Issue:** `make fmt` reformatted several pre-existing files (double_metaphone.go, soundex_fuzz_test.go, soundex_test.go, etc.) which were detected as modified
- **Fix:** Included the formatted versions in the commit to keep the repo clean
- **Files modified:** double_metaphone.go, double_metaphone_fuzz_test.go, double_metaphone_test.go, soundex_fuzz_test.go, soundex_test.go
- **Commit:** 862055f

### Pre-existing Lint Issues (Out of Scope per Scope-Boundary Rule)

The following golangci-lint failures were present before this plan and are out of scope:
- `double_metaphone_test.go:322:8` QF1001 (De Morgan's law) — from plan 07-02
- `props_test.go:3492:8` QF1001 (De Morgan's law) — from plan 07-02's DoubleMetaphone charset test
- `soundex_test.go:203:6` QF1001 (De Morgan's law) — from plan 07-01
- `soundex.go:275:31` unparam (runeAt) — from plan 07-01
- Various gocyclo/gofumpt issues — from prior plans

These were logged and NOT fixed per the scope-boundary rule.

## Known Stubs

None — all three public surfaces are fully wired with real implementation.

## Threat Flags

None — MRA is a pure-function algorithm with no network endpoints, auth paths, file access patterns, or schema changes at trust boundaries.

## Self-Check

## Self-Check: PASSED

| Check | Status |
|-------|--------|
| mra.go exists | FOUND |
| dispatch_mra.go exists | FOUND |
| mra_test.go exists | FOUND |
| mra_bench_test.go exists | FOUND |
| mra_fuzz_test.go exists | FOUND |
| testdata/golden/_staging/mra.json exists | FOUND |
| tests/bdd/features/mra.feature exists | FOUND |
| 07-04-SUMMARY.md exists | FOUND |
| Commit 862055f exists | FOUND |
| `var mraThresholdTable` at package level | VERIFIED |
| sum>12 clamp documented in mra.go | VERIFIED |
| AlgoMRA in permittedMongeElkanInner | VERIFIED (1 entry) |
| permittedMongeElkanInner = 18 entries | VERIFIED |
| rejected slice = 5 entries | VERIFIED |
