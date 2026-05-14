---
phase: 02-core-character-algorithms-six
plan: 07
subsystem: testing
tags: [golden-file, bench-baseline, cross-algorithm-consistency, identifier-similarity-example, algorithms-json-merge]

# Dependency graph
requires:
  - phase: 02-core-character-algorithms-six
    provides: "Levenshtein, Hamming, Jaro, JaroWinkler, DL-OSA, DL-Full algorithms; per-algorithm _staging/*.json files"
provides:
  - "testdata/golden/algorithms.json merged from all 6 staging files (32 entries, 6 algorithms, sorted by Name)"
  - "cross_algorithm_consistency_test.go with 5 cross-algorithm contracts pinned"
  - "examples/identifier-similarity/ runnable program + byte-stable meta-test"
  - "bench.txt first benchstat baseline (25 benchmark series, 386 lines)"
affects: [03-q-gram-token, phase-3+, bdd, determinism-ci]

# Tech tracking
tech-stack:
  added: []
  patterns:
    - "Staging-merge pattern: Wave N plans write _staging/<algo>.json; Wave N+1 plan merges into canonical algorithms.json via TestGolden_Algorithms_Merge"
    - "Example-in-module pattern: examples/<name>/go.mod with replace ../../ directive; package main; TestExample_Output captures stdout via os.Pipe"
    - "Cross-algorithm consistency test: five tests in cross_algorithm_consistency_test.go pin divergence/convergence/agreement contracts across the full algorithm catalogue"

key-files:
  created:
    - testdata/golden/algorithms.json (merged from 6 staging files — 32 entries)
    - cross_algorithm_consistency_test.go (5 cross-algorithm consistency tests)
    - examples/identifier-similarity/main.go
    - examples/identifier-similarity/main_test.go
    - examples/identifier-similarity/go.mod
    - examples/identifier-similarity/go.sum
    - bench.txt (first benchstat baseline — 386 lines, 25 benchmark series)
  modified:
    - algorithms_golden_test.go (TestGolden_Algorithms → TestGolden_Algorithms_Merge)
    - damerau_full.go (gofmt -s alignment fix)
    - damerau_full_test.go (gofmt -s alignment fix)
    - jarowinkler_bench_test.go (gofmt -s alignment fix)
    - jarowinkler_test.go (gofmt -s alignment fix; added coverage tests for JaroWinklerScoreRunes)

key-decisions:
  - "PERF-03 DL-Full deviation tracked as GitHub issue #1 (two-row + H-table deferred to v1.x)"
  - "TestGolden_Algorithms_Merge supersedes Wave 1 TestGolden_Algorithms; Wave 1 function removed"
  - "Example meta-test uses os.Pipe stdout capture (package main scope) rather than os/exec subprocess"
  - "bench.txt generated from go test -bench=. -benchmem -count=10 ./... (not make bench which produces bench.txt.new)"
  - "Hamming shows 0.0000 in example for all length-mismatched pairs — ERR is superseded by locked silent-zero policy"

patterns-established:
  - "Golden merge: all future phases that add algorithms should follow the _staging/ → merge pattern"
  - "Example modules: replace ../../ with package main + TestExample_Output via os.Pipe"

requirements-completed: [DET-02, DX-02, DX-05, TEST-01]

# Metrics
duration: 11min
completed: 2026-05-14
---

# Phase 2 Plan 07: Finalisation Summary

**Merged six per-algorithm staging golden files into canonical algorithms.json (32 entries), shipped runnable identifier-similarity example with byte-stable meta-test, pinned cross-algorithm OSA-vs-Full divergence and identity convergence contracts, committed first bench.txt baseline (25 benchmark series)**

## Performance

- **Duration:** ~11 min
- **Started:** 2026-05-14T07:37:30Z
- **Completed:** 2026-05-14T07:48:47Z
- **Tasks:** 3 + 1 fixup commit
- **Files modified:** 11 files changed

## Accomplishments

- `testdata/golden/algorithms.json` now contains 32 entries from all six algorithms (Levenshtein×4, Hamming×4, Jaro×6, JaroWinkler×8, DL-OSA×5, DL-Full×5), sorted alphabetically by Name, byte-stable via `CanonicalMarshalForTest`
- `cross_algorithm_consistency_test.go` pins five cross-algorithm contracts: DL-OSA vs DL-Full divergence on "ca"/"abc" (distance 3 vs 2), identity convergence (all six → 1.0), both-empty convergence, single-substitution distance agreement, one-empty score agreement
- `examples/identifier-similarity/` ships a runnable demonstration of all six algorithms on 7 database column-name pairs; `TestExample_Output` asserts byte-stable stdout via `os.Pipe`
- First `bench.txt` benchstat baseline committed (386 lines, 25 benchmark series across 6 algorithms + AlgoID_String + Normalise); mean ASCII-Short ns/op: Levenshtein≈95, DL-OSA≈40, DL-Full≈68, Hamming≈6, Jaro≈48, JaroWinkler≈49
- `make check` passes with overall coverage 96.7% and per-file coverage ≥90% for all files
- PERF-03 DL-Full heap-allocation deviation tracked as GitHub issue #1

## Task Commits

Each task was committed atomically:

1. **Task 1: Merge per-algorithm staging files into canonical algorithms.json** - `fbcfd43` (feat)
2. **Task 2: Cross-algorithm consistency tests** - `c8e3f23` (test)
3. **Task 3: identifier-similarity example + bench.txt baseline + quality fixes** - `f8eadef` (feat)
4. **bench.txt update: regenerate with all 6 algorithms** - `9fcc1aa` (chore)

## Files Created/Modified

- `testdata/golden/algorithms.json` — merged canonical file with 32 entries from all 6 algorithms
- `algorithms_golden_test.go` — replaced TestGolden_Algorithms with TestGolden_Algorithms_Merge; added encoding/json import
- `cross_algorithm_consistency_test.go` — 5 cross-algorithm consistency tests (package fuzzymatch_test, stdlib only)
- `examples/identifier-similarity/main.go` — package main, 7 pairs × 6 algorithms table
- `examples/identifier-similarity/main_test.go` — TestExample_Output via os.Pipe capture
- `examples/identifier-similarity/go.mod` — go 1.26.3, replace ../../, zero non-stdlib runtime deps beyond fuzzymatch
- `examples/identifier-similarity/go.sum` — golang.org/x/text transitive only
- `bench.txt` — 386 lines, first benchstat baseline
- `damerau_full.go` — gofmt -s alignment fix (pre-existing Wave 2 issue)
- `damerau_full_test.go` — gofmt -s alignment fix
- `jarowinkler_bench_test.go` — gofmt -s alignment fix
- `jarowinkler_test.go` — gofmt -s fix + added JaroWinklerScoreRunes coverage tests

## Example Program stdout (committed in main_test.go `want` constant)

```
Pair (a / b)                      Levenshtein       DL-OSA      DL-Full      Hamming         Jaro Jaro-Winkler
--------------------------------------------------------------------------------------------------------------
user_id / userId                       0.7143       0.7143       0.7143       0.0000       0.8492       0.9095
created_at / creationTimestamp         0.4118       0.4118       0.4118       0.0000       0.7152       0.8291
status / state                         0.6667       0.6667       0.6667       0.0000       0.8222       0.8933
email / e_mail                         0.8333       0.8333       0.8333       0.0000       0.9444       0.9500
org_id / organisation_id               0.4000       0.4000       0.4000       0.0000       0.6444       0.6444
latitude / longitude                   0.6667       0.6667       0.6667       0.0000       0.7500       0.7750
is_deleted / is_active                 0.4000       0.4000       0.4000       0.0000       0.6185       0.6185
```

Note: all Hamming cells show 0.0000 because all pairs are length-mismatched (`latitude`=8 chars, `longitude`=9 chars). This is correct per the locked silent-zero policy from plan 02-02.

## Benchmark Baseline (bench.txt ASCII-Short mean ns/op, Apple M2)

| Algorithm | ASCII-Short (mean) | Allocs |
|-----------|-------------------|--------|
| Levenshtein | ~95 ns/op | 0 |
| DL-OSA | ~40 ns/op | 0 |
| DL-Full | ~68 ns/op | 1 (128 B) — PERF-03 open |
| Hamming | ~6 ns/op | 0 |
| Jaro | ~48 ns/op | 0 |
| JaroWinkler | ~49 ns/op | 0 |

DL-Full is the only algorithm with a non-zero allocation on ASCII-Short inputs due to the heap-allocated full DP table (PERF-03 deviation, tracked as GitHub issue #1).

## Decisions Made

- **TestGolden_Algorithms_Merge supersedes Wave 1 TestGolden_Algorithms**: The Wave 1 stub (Levenshtein-only) is deleted; the Merge form reads all six _staging/*.json files and is the canonical golden gate going forward. The six staging files are preserved as the per-algorithm audit trail.
- **Example meta-test uses os.Pipe rather than os/exec subprocess**: Calling `main()` directly in package main scope with redirected stdout is simpler, faster, and avoids spawning a subprocess that would require the Go toolchain at test time.
- **PERF-03 DL-Full deviation tracked as GitHub issue #1**: The heap-allocated full DP table for DL-Full was documented in plan 02-06's SUMMARY as a v1.x follow-up. Per Task 2's acceptance criteria, this deviation must have an issue link before Phase 3 begins. Issue #1 was created at https://github.com/axonops/fuzzymatch/issues/1.
- **Hamming silent-zero supersedes CONTEXT.md ERR format**: CONTEXT.md deferred section mentioned `ERR` for Hamming length-mismatch. The locked Hamming policy from plan 02-02 (commit 1e25e31) supersedes this — the example shows `0.0000`, not `ERR`. This is a documentation supersession noted in main.go's file-level godoc.

## Deviations from Plan

### Auto-fixed Issues

**1. [Rule 3 - Blocking] Fix pre-existing gofmt formatting issues blocking make check**
- **Found during:** Task 3 (make check gate)
- **Issue:** `damerau_full.go`, `damerau_full_test.go`, `jarowinkler_bench_test.go`, `jarowinkler_test.go` had trailing-alignment whitespace that `gofmt -s` would reformat — pre-existing from Wave 2 plans
- **Fix:** `gofmt -s -w` on all four files
- **Files modified:** damerau_full.go, damerau_full_test.go, jarowinkler_bench_test.go, jarowinkler_test.go
- **Verification:** `make fmt-check` passes
- **Committed in:** f8eadef (Task 3 commit)

**2. [Rule 3 - Blocking] Add JaroWinklerScoreRunes coverage tests to pass 90% per-file floor**
- **Found during:** Task 3 (make coverage-check gate)
- **Issue:** `jarowinkler.go` `JaroWinklerScoreRunes` function was at 78.6% coverage (below 90% per-file floor), specifically missing the `j < winklerBoostThreshold` branch in the rune path
- **Fix:** Added `TestJaroWinklerScoreRunes_BelowBoostThreshold` and `TestJaroWinklerScoreRunes_IdentityAndEmpty` to `jarowinkler_test.go`; brought `JaroWinklerScoreRunes` coverage to 92.9%
- **Files modified:** jarowinkler_test.go
- **Verification:** `make coverage-check` passes; overall 96.7%; all per-file ≥90%
- **Committed in:** f8eadef (Task 3 commit)

**3. [Rule 3 - Blocking] Regenerate bench.txt with full benchmark suite**
- **Found during:** After Task 3 commit (post-commit inspection)
- **Issue:** Initial bench.txt (from make bench before the coverage fix run) was missing Levenshtein, Jaro, JaroWinkler benchmarks because `make bench` wrote to `bench.txt.new` at a point when only DL and Hamming benchmarks existed in the test cache
- **Fix:** Re-ran `go test -bench=. -benchmem -count=10 ./...` directly to produce complete output; committed as separate chore commit
- **Files modified:** bench.txt
- **Committed in:** 9fcc1aa (fixup commit)

---

**Total deviations:** 3 auto-fixed (all Rule 3 — blocking quality gate issues)
**Impact on plan:** All auto-fixes necessary for make check to pass. The gofmt and coverage fixes were pre-existing Wave 2 quality issues. No scope creep.

## Issues Encountered

- `make bench` writes to `bench.txt.new` not `bench.txt` — ran `go test -bench=.` directly to populate `bench.txt`
- `go mod tidy` on the example module upgraded `go 1.26` to `go 1.26.3` (matching the root module's directive) — expected behaviour

## PERF-03 Deviation Gate (Wave 3 required check)

Per Task 2 acceptance criteria, plan 02-06-SUMMARY.md was reviewed for PERF-03 disposition:

- **DL-Full PERF-03 status:** Deviation documented — heap-allocated full `(m+2)×(n+2)` DP table used for all inputs; two-row + H-table optimisation deferred to v1.x
- **GitHub issue filed:** https://github.com/axonops/fuzzymatch/issues/1
- **Gate disposition:** PERF-03 is tracked. Phase 3 may proceed.

## Next Phase Readiness

- Phase 2 is shippable: `make check` passes, all 6 algorithms golden-gated, cross-algorithm contracts pinned, example runnable, bench.txt baseline committed
- Phase 3 (q-gram and token-based algorithms) can begin; the six character-based algorithms and their patterns (two-row DP, ASCII fast path, dispatch registration, _staging golden files) serve as the implementation template
- PERF-03 (DL-Full two-row + H-table) is tracked as issue #1 — resolve before v1.0 but not blocking Phase 3

## Self-Check

All required artifacts verified present:

- [x] `testdata/golden/algorithms.json` — FOUND; 32 entries; 6 unique algorithm names
- [x] `cross_algorithm_consistency_test.go` — FOUND; contains TestCrossAlgorithm_OSA_Full_Divergence
- [x] `examples/identifier-similarity/main.go` — FOUND; package main; imports fuzzymatch
- [x] `examples/identifier-similarity/main_test.go` — FOUND; TestExample_Output passes
- [x] `examples/identifier-similarity/go.mod` — FOUND; replace github.com/axonops/fuzzymatch
- [x] `examples/identifier-similarity/go.sum` — FOUND
- [x] `bench.txt` — FOUND; 386 lines; all 6 algorithms present (88+ benchmark lines)

All commits verified:
- [x] fbcfd43 — Task 1 (algorithms.json merge)
- [x] c8e3f23 — Task 2 (cross-algorithm consistency tests)
- [x] f8eadef — Task 3 (example + bench.txt + quality fixes)
- [x] 9fcc1aa — bench.txt complete baseline

Final quality gates:
- [x] `make check` exits 0 — coverage 96.7%, lint 0 issues, vet clean
- [x] `make verify-determinism` exits 0
- [x] `(cd examples/identifier-similarity && go test -race -count=1 ./...)` exits 0
- [x] `bash scripts/verify-license-headers.sh` exits 0
- [x] `bash scripts/verify-no-runtime-deps.sh` exits 0
- [x] `make test-bdd` exits 0

## Self-Check: PASSED

---
*Phase: 02-core-character-algorithms-six*
*Completed: 2026-05-14*
