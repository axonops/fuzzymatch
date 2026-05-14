---
phase: 02-core-character-algorithms-six
plan: 02
subsystem: algorithm
tags: [hamming, equal-length, silent-zero-unequal, no-dp, zero-alloc, dispatch-registration, golden-file, bdd, property-tests, fuzz]

# Dependency graph
requires:
  - phase: 02-core-character-algorithms-six
    plan: 01
    provides: dispatch array, assertGoldenStaging helper, props_test.go, example_test.go, tests/bdd/steps/algorithms_steps.go append points

provides:
  - HammingDistance(a, b string) int — byte-level single-loop counting
  - HammingDistanceRunes(a, b string) int — rune-level counting (eager []rune)
  - HammingScore(a, b string) float64 — normalised byte-level similarity
  - HammingScoreRunes(a, b string) float64 — normalised rune-level similarity
  - dispatch[AlgoHamming] populated at package load (no init())
  - testdata/golden/_staging/hamming.json — 4-entry staging file for plan 02-07 merge
  - props_test.go extended — 7 Hamming property tests (including equal-length triangle inequality)
  - example_test.go extended — ExampleHammingScore demonstrating locked silent-zero policy
  - tests/bdd/features/hamming.feature — 6 BDD scenarios
  - tests/bdd/steps/algorithms_steps.go extended — 5 Hamming step bindings

affects:
  - 02-07-algorithms-merge (reads _staging/hamming.json for uniform merge)
  - Wave 2 plans consuming props_test.go, example_test.go, algorithms_steps.go as append points

# Tech tracking
tech-stack:
  added: []
  patterns:
    - "Single-loop O(min(m,n)) Hamming counting — no DP buffer — 0 allocs at any ASCII length"
    - "Silent-zero unequal-length policy: Distance returns max(len(a),len(b)); Score returns 0.0"
    - "Locked godoc paragraph verbatim from CONTEXT.md (load-bearing contract)"
    - "Equal-length-constrained triangle inequality property test (XOR-flip byte generation)"
    - "BDD distance step: lastDistance int field + theDistanceShouldBe(expected int) step"

key-files:
  created:
    - hamming.go
    - dispatch_hamming.go
    - hamming_test.go
    - hamming_bench_test.go
    - hamming_fuzz_test.go
    - testdata/fuzz/FuzzHammingScore/seed-001
    - testdata/golden/_staging/hamming.json
    - tests/bdd/features/hamming.feature
  modified:
    - props_test.go (appended 7 Hamming property tests)
    - example_test.go (appended ExampleHammingScore)
    - algoid_test.go (added TestDispatch_HammingRegistered; updated TestDispatch_UnregisteredSlotsAreNil)
    - algorithms_golden_test.go (added buildHammingStagingEntries + TestGolden_Hamming_Staging)
    - tests/bdd/steps/algorithms_steps.go (added lastDistance field + 4 step methods + 5 registrations)
    - llms.txt (added Hamming section — required by TestAIFriendly gate)

key-decisions:
  - "HammingScore(a,b string) float64 — no error return, per locked CONTEXT.md decision"
  - "Unequal-length: Distance returns max(len(a),len(b)); Score returns 0.0 silently"
  - "Triangle inequality: equal-length-constrained variant added (TestProp_HammingDistance_TriangleInequality_EqualLength); general TI omitted per silent-zero policy"
  - "Unicode rune path: 0 allocs observed on Apple M2 (Go compiler stack-allocates short rune slices) — better than the documented 2-alloc budget"

requirements-completed: [CHAR-04, PERF-01, PERF-02, TEST-01, TEST-02, TEST-04, TEST-05, DET-04, DX-02]

# Metrics
duration: 18min
completed: 2026-05-14
---

# Phase 2 Plan 02: Hamming Summary

**Hamming distance/score shipped end-to-end with the LOCKED silent-zero unequal-length policy: single counting loop, 0 allocs at any ASCII length, locked godoc paragraph, staging golden file, BDD scenarios, and dispatch registration — establishing the canonical Wave 2 pattern for Hamming as the simplest algorithm in the catalogue**

## Performance

- **Duration:** 18 min
- **Started:** 2026-05-14
- **Completed:** 2026-05-14
- **Tasks:** 3
- **Files created:** 9 new files, 6 modified

## Accomplishments

- Implemented four public functions (HammingDistance, HammingDistanceRunes, HammingScore, HammingScoreRunes) from Hamming 1950 as a single O(min(m,n)) counting loop — no DP table, no buffer — achieving 0 allocs at any ASCII length (confirmed via `testing.AllocsPerRun` and `-benchmem`)
- Honoured the LOCKED silent-zero unequal-length policy from CONTEXT.md verbatim in godoc, tests, BDD, and ExampleHammingScore
- Created testdata/golden/_staging/hamming.json (4 entries sorted alphabetically) ready for plan 02-07 Wave 3 merge
- Extended all four Wave 2 shared append-point files (props_test.go, example_test.go, algoid_test.go, tests/bdd/steps/algorithms_steps.go) without breaking any Wave 1 tests; hamming.go achieved 100% per-function coverage

## Locked Unequal-Length Godoc Paragraph

The following text was written verbatim into hamming.go (file-level godoc block, paragraph 4):

```
// Inputs of unequal length are not an error: HammingDistance returns
// max(len(a), len(b)) and HammingScore returns 0.0. Callers wanting
// strict Hamming-1950 equal-length semantics should length-check
// before calling.
```

This exact text also appears in the per-function godoc of HammingDistance, HammingDistanceRunes, HammingScore, and HammingScoreRunes (with adapted phrasing for the rune variants).

## Benchmark Numbers (Apple M2, darwin/arm64)

| Benchmark | ns/op | B/op | allocs/op |
|-----------|-------|------|-----------|
| ASCII_Short (7/7 bytes, karolin/kathrin) | 8.0 | 0 | 0 |
| ASCII_Medium (50 bytes) | 32.8 | 0 | 0 |
| ASCII_Long (500 bytes) | 302.6 | 0 | 0 |
| Unicode_Short (café/cafè, rune path) | 27.7 | 0 | 0 |

Note: the Unicode_Short rune path also shows 0 allocs/op. The Go compiler on arm64 stack-allocates the two short `[]rune` slices for "café"/"cafè" (4 runes each, 32 bytes total). The plan documented a 2-alloc budget as the expected worst case; 0 is strictly better. `testing.AllocsPerRun` confirms 0 for the Unicode path as well.

## Triangle Inequality Decision

The general Hamming triangle inequality was **omitted** for arbitrary inputs. Under the silent-zero policy, calling `HammingDistance("abc","")` returns 3 (max-len path) rather than a true edit-distance, making the standard metric property undefined for mixed-length comparisons.

An **equal-length-constrained variant** was added instead:

```go
func TestProp_HammingDistance_TriangleInequality_EqualLength(t *testing.T)
```

This generates a random base string, then b and c as same-length variants (XOR-flipped bytes at controlled positions), verifying D(base,c) ≤ D(base,b) + D(b,c). All inputs have the same byte length, so max-len is never triggered.

## Coverage

- **Overall:** 97.9% (floor: 95%)
- **hamming.go:** 100% (all four public functions)
- **Public symbols:** HammingDistance 100%, HammingDistanceRunes 100%, HammingScore 100%, HammingScoreRunes 100%

## Task Commits

1. **Task 1: Algorithm + dispatch** — `d78a07b` (feat)
2. **Task 2: Unit/property/benchmark/fuzz tests + dispatch assertion update** — `84bbe08` (test)
3. **Task 3: Golden staging + BDD + shared file extensions** — `a704300` (feat)

## Files Created/Modified

- `hamming.go` — Four public functions + locked unequal-length godoc; no math.X transcendentals; no init(); no []byte conversions; no maxStackInputLen redeclaration
- `dispatch_hamming.go` — Package-load-time registration via var _ = func() bool idiom
- `hamming_test.go` — Tests: BothEmpty, Identical, ReferenceVectors (karolin/kathrin + binary pair), UnequalLength_SilentZero, Symmetry, DistanceRunes_MultiByte (café/cafè=1 rune), DistanceRunes_UnequalLength, ScoreRunes_BothEmpty, ZeroAllocs
- `hamming_bench_test.go` — Benchmarks at 4 sizes: ASCII Short/Medium/Long (all 0 allocs) + Unicode Short (0 allocs)
- `hamming_fuzz_test.go` — FuzzHammingScore with invalid-UTF-8 and unequal-length seeds; 30s run clean
- `testdata/fuzz/FuzzHammingScore/seed-001` — karolin/kathrin pair in go test fuzz v1 format
- `testdata/golden/_staging/hamming.json` — 4 entries sorted alphabetically; canonical byte form (2-space indent, trailing LF, no BOM)
- `tests/bdd/features/hamming.feature` — 6 scenarios: canonical reference vectors, both-empty, unequal-length silent-zero, distance-equals-max-length, symmetry (equal+unequal-length)
- `props_test.go` (modified) — Appended 7 TestProp_HammingScore_* tests + equal-length TI variant with rationale comment
- `example_test.go` (modified) — Appended ExampleHammingScore demonstrating both equal-length (0.5714) and unequal-length silent-zero (0.0000)
- `algoid_test.go` (modified) — Added TestDispatch_HammingRegistered; refactored TestDispatch_UnregisteredSlotsAreNil to use map-based registered-set
- `algorithms_golden_test.go` (modified) — Added buildHammingStagingEntries + TestGolden_Hamming_Staging
- `tests/bdd/steps/algorithms_steps.go` (modified) — Added lastDistance field; 4 step methods; 5 step registrations in InitializeScenario
- `llms.txt` (modified) — Added Hamming section (4 exported symbols)

## Decisions Made

- Unequal-length policy: `max(len(a), len(b))` as distance, `0.0` as score — per CONTEXT.md locked decision
- No `[]byte(s)` conversion anywhere — direct string byte indexing is zero-allocation
- No `maxStackInputLen` usage — Hamming needs no DP buffer
- Triangle inequality: equal-length-constrained variant added; general TI documented as omitted with CONTEXT.md rationale
- Unicode rune variant: eager `[]rune(s)` per Pattern 8 (Wave 1 locked decision)

## Deviations from Plan

### Auto-fixed Issues

**1. [Rule 2 - Missing Critical] Updated llms.txt to include Hamming symbols**
- **Found during:** Task 3 (make check quality gate — TestAIFriendly_LLMSTxtReferencesEveryExportedSymbol)
- **Issue:** 4 new exported symbols (HammingDistance, HammingDistanceRunes, HammingScore, HammingScoreRunes) absent from llms.txt
- **Fix:** Added "Hamming distance similarity" section to llms.txt public API listing
- **Files modified:** llms.txt
- **Verification:** go test -run TestAIFriendly -count=1 ./... exits 0
- **Committed in:** a704300 (Task 3 commit)
- **Note:** Same deviation as Wave 1 plan 02-01 — the TestAIFriendly gate requires llms.txt updates for every new exported symbol

**2. [Rule 1 - Bug] Fixed staticcheck ST1005 in algorithms_steps.go**
- **Found during:** Task 3 (make check — cd tests/bdd && golangci-lint run ./...)
- **Issue:** `fmt.Errorf("Hamming scores not equal: ...")` — staticcheck ST1005 rejects capitalised error strings
- **Fix:** Changed to lowercase: `"hamming scores not equal: ..."`
- **Files modified:** tests/bdd/steps/algorithms_steps.go
- **Verification:** golangci-lint run ./... exits 0 with 0 issues in tests/bdd/
- **Committed in:** a704300 (Task 3 commit)

**3. [Rule 2 - Missing Critical] Added rune-variant coverage tests (TestHamming_DistanceRunes_UnequalLength, TestHamming_ScoreRunes_BothEmpty)**
- **Found during:** Task 3 (make check — verify-coverage-floors.sh)
- **Issue:** hamming.go per-file coverage was 86.96% (< 90% floor); HammingDistanceRunes was at 75%, HammingScoreRunes at 80% — missing the rune-count inequality and both-empty paths
- **Fix:** Added two test functions to hamming_test.go covering all uncovered paths
- **Files modified:** hamming_test.go
- **Verification:** go tool cover -func=coverage.out shows 100% per-function on all four Hamming functions; overall 97.9%
- **Committed in:** a704300 (Task 3 commit)

**4. [Rule 1 - Bug] Fixed gofmt alignment in hamming_fuzz_test.go**
- **Found during:** Task 3 (make check — fmt-check target)
- **Issue:** gofmt -s flagged trailing-space alignment in the struct literal inside FuzzHammingScore
- **Fix:** gofmt -s -w hamming_fuzz_test.go
- **Files modified:** hamming_fuzz_test.go
- **Committed in:** a704300 (Task 3 commit)

---

**Total deviations:** 4 auto-fixed (2 missing critical, 2 bugs)
**Impact on plan:** All four fixes required for quality gate compliance. No scope creep.

## Known Stubs

None. All four public functions are fully wired with live computation. ExampleHammingScore verifies byte-for-byte output. The staging golden file entries are generated from live HammingScore calls.

## Threat Flags

No new network endpoints, auth paths, file access patterns, or schema changes introduced. T-02-02-01 through T-02-02-05 as documented in the plan's threat model — all accepted or mitigated by existing mechanisms (fuzz tests cover T-02-02-02; dispatch registration is unexported covering T-02-02-03).

---
*Phase: 02-core-character-algorithms-six*
*Completed: 2026-05-14*
