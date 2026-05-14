---
phase: 02-core-character-algorithms-six
plan: 01
subsystem: algorithm
tags: [levenshtein, two-row-dp, ascii-fast-path, dispatch-registration, golden-file, bdd, property-tests, fuzz, canonical-pattern]

# Dependency graph
requires:
  - phase: 01-foundation-infrastructure
    provides: AlgoID enum (AlgoLevenshtein slot 0), dispatch array, isASCII helper, normalise.go, golden_canonical.go, export_test.go re-exports, tests/bdd go.mod

provides:
  - LevenshteinDistance(a, b string) int — byte-level two-row DP
  - LevenshteinDistanceRunes(a, b string) int — rune-level two-row DP
  - LevenshteinScore(a, b string) float64 — normalised byte-level similarity
  - LevenshteinScoreRunes(a, b string) float64 — normalised rune-level similarity
  - dispatch[AlgoLevenshtein] populated at package load (no init())
  - maxStackInputLen=64 constant shared by Wave 2 DP algorithms
  - assertGoldenStaging(t, relPath, v) helper — LOCKED signature for Wave 2
  - testdata/golden/algorithms.json — 4 Levenshtein entries, canonical byte form
  - testdata/golden/_staging/levenshtein.json — per-algorithm staging for plan 02-07 merge
  - props_test.go — shared Phase 2 property test file (Wave 2 plans append)
  - example_test.go — shared Phase 2 godoc examples file (Wave 2 plans append)
  - tests/bdd/bdd_test.go — godog runner with goleak.VerifyTestMain
  - tests/bdd/steps/algorithms_steps.go — AlgorithmContext + InitializeScenario (Wave 2 extends)
  - tests/bdd/features/levenshtein.feature — BDD scenarios for Levenshtein

affects:
  - 02-02-hamming, 02-03-jaro, 02-04-jaro-winkler, 02-05-damerau-levenshtein-osa,
    02-06-damerau-levenshtein-full (all Wave 2 plans consume the canonical pattern)
  - 02-07-algorithms-merge (reads _staging/levenshtein.json for uniform merge)

# Tech tracking
tech-stack:
  added: []
  patterns:
    - "Two-row DP with stack-allocated [(maxStackInputLen+1)*2]int buffer for ASCII ≤ 64 bytes"
    - "Dispatch registration via var _ = func() bool { ... }() idiom (no init())"
    - "Per-algorithm dispatch_<algo>.go files to avoid Wave 2 merge conflicts"
    - "assertGoldenStaging helper for per-algorithm staging golden files"
    - "testing.AllocsPerRun as runtime zero-alloc gate (not just benchmark-time)"
    - "One godog feature file per algorithm in tests/bdd/features/<algo>.feature"

key-files:
  created:
    - levenshtein.go
    - dispatch_levenshtein.go
    - levenshtein_test.go
    - levenshtein_bench_test.go
    - levenshtein_fuzz_test.go
    - props_test.go
    - example_test.go
    - algorithms_golden_test.go
    - testdata/golden/algorithms.json
    - testdata/golden/_staging/levenshtein.json
    - testdata/fuzz/FuzzLevenshteinScore/seed-001
    - tests/bdd/bdd_test.go
    - tests/bdd/features/levenshtein.feature
    - tests/bdd/steps/algorithms_steps.go
  modified:
    - algoid_test.go (renamed TestDispatch_AllNilAtPhase1 to two tests)
    - llms.txt (added LevenshteinDistance/Score/Runes variants)

key-decisions:
  - "maxStackInputLen=64: stack-buffer threshold shared across all Phase 2 DP algorithms (declared once in levenshtein.go)"
  - "Dispatch registration: var _ = func() bool { ... }() per algorithm in separate dispatch_<algo>.go (no init())"
  - "Rune variant strategy: eager []rune(s) conversion (Pattern A); 0-alloc budget applies only to ASCII byte path"
  - "assertGoldenStaging signature LOCKED: (t *testing.T, relPath string, v any) — Wave 2 plans call directly"
  - "BDD: one feature file per algorithm; shared AlgorithmContext + InitializeScenario in steps/algorithms_steps.go"
  - "Wave 1 produces both algorithms.json AND _staging/levenshtein.json for uniform plan 02-07 merge inputs"

patterns-established:
  - "Pattern: <algo>.go + dispatch_<algo>.go + <algo>_test.go + <algo>_bench_test.go + <algo>_fuzz_test.go per algorithm"
  - "Pattern: props_test.go and example_test.go are shared append-point files for Wave 2 plans"
  - "Pattern: algorithms_golden_test.go owns assertGoldenStaging helper; Wave 2 plans call it directly"
  - "Pattern: tests/bdd/steps/algorithms_steps.go is the Wave 2 extension point"

requirements-completed: [CHAR-01, PERF-01, PERF-02, PERF-03, TEST-01, TEST-02, TEST-04, TEST-05, DET-02, DET-04, DX-02]

# Metrics
duration: 11min
completed: 2026-05-14
---

# Phase 2 Plan 01: Levenshtein Summary

**Levenshtein edit-distance similarity shipped end-to-end as the Phase 2 canonical pattern: two-row DP with stack-allocated buffer (0 allocs on ASCII ≤ 64 bytes), dispatch registration via var-init idiom, golden file, BDD harness, and locked assertGoldenStaging helper for Wave 2 plans**

## Performance

- **Duration:** 11 min
- **Started:** 2026-05-14T05:58:26Z
- **Completed:** 2026-05-14T06:10:13Z
- **Tasks:** 3
- **Files created:** 14 new files, 2 modified

## Accomplishments

- Implemented four public functions (LevenshteinDistance, LevenshteinDistanceRunes, LevenshteinScore, LevenshteinScoreRunes) from primary sources (Levenshtein 1965 / Wagner-Fischer 1974) with two-row DP achieving 0 allocs on ASCII ≤ 64 bytes (confirmed: 90 ns/op, 0 B/op, 0 allocs/op on Apple M2)
- Established the complete canonical file pattern (algorithm + dispatch + unit + property + benchmark + fuzz + golden + staging + BDD + example) that Wave 2's five parallel plans replicate without re-deciding any detail
- Locked assertGoldenStaging helper and produced both algorithms.json and _staging/levenshtein.json so plan 02-07's merge step has uniform inputs across all six algorithms

## Benchmark Numbers (Apple M2, darwin/arm64)

| Benchmark | ns/op | B/op | allocs/op |
|-----------|-------|------|-----------|
| ASCII_Short (6/7 bytes, stack path) | 90 | 0 | 0 |
| ASCII_Medium (50 bytes, stack path) | 6,318 | 0 | 0 |
| ASCII_Long (500 bytes, heap path) | 754,677 | 8,192 | 2 |
| Unicode_Short (café/cafe, rune path) | 69 | 96 | 2 |

## Coverage

- **Overall:** 97.6% (floor: 95%)
- **levenshtein.go:** 100% (all statements)
- **Public symbols:** LevenshteinDistance 100%, LevenshteinDistanceRunes 100%, LevenshteinScore 100%, LevenshteinScoreRunes 100%

## Canonical Pattern Locked for Wave 2

| Decision | Locked Value |
|----------|-------------|
| maxStackInputLen | 64 |
| Stack buffer size | [(64+1)*2]int = 130 ints = 1040 bytes |
| Dispatch registration idiom | `var _ = func() bool { dispatch[AlgoXxx] = XxxScore; return true }()` |
| Rune variant strategy | Eager `[]rune(s)` conversion (2 allocs, documented) |
| assertGoldenStaging signature | `(t *testing.T, relPath string, v any)` |
| BDD feature file location | `tests/bdd/features/<algo>.feature` |
| Staging file location | `testdata/golden/_staging/<algo>.json` |

## Wave 2 Hand-off Contract

Wave 2 plans (02-02 through 02-06) consume these append-point files:
- **props_test.go** — append TestProp_XxxScore_* functions (6 property tests per algorithm)
- **example_test.go** — append ExampleXxxScore() function
- **tests/bdd/steps/algorithms_steps.go** — append step functions + register in InitializeScenario
- **algorithms_golden_test.go** — call `assertGoldenStaging(t, "_staging/<algo>.json", file)` directly; helper already defined

## Task Commits

1. **Task 1: Algorithm + dispatch** - `5dbb295` (feat)
2. **Task 2: Unit/property/benchmark/fuzz tests + dispatch assertion update** - `484d785` (test)
3. **Task 3: Golden files + staging helper + BDD harness + llms.txt** - `94bf93f` (feat)

## Files Created/Modified

- `levenshtein.go` — Four public functions + levenshteinDP kernel + levenshteinDistanceRuneSlices; maxStackInputLen=64; primary-source citation; no math.X transcendentals; no init(); no []byte conversions
- `dispatch_levenshtein.go` — Package-load-time registration via var _ = func() bool idiom
- `levenshtein_test.go` — Unit tests: BothEmpty, OneEmpty, Identical, ReferenceVectors, Symmetry, DistanceRunes_MultiByte, ASCII_vs_Rune_Equivalence, ZeroAllocs_ASCII_Short/Medium
- `levenshtein_bench_test.go` — Benchmarks at 4 sizes: ASCII Short/Medium/Long + Unicode Short
- `levenshtein_fuzz_test.go` — FuzzLevenshteinScore with 10 seeds including invalid UTF-8 and Cyrillic
- `props_test.go` — 7 testing/quick property tests (RangeBounds, Identity, Symmetric, TriangleInequality, NoNaN, NoInf, NoNegativeZero); shared file for Wave 2
- `example_test.go` — ExampleLevenshteinScore with verified Output block (0.5714); shared for Wave 2
- `algorithms_golden_test.go` — assertGoldenStaging (LOCKED signature), TestGolden_Algorithms, TestGolden_Levenshtein_Staging, goldenAlgorithmEntry/goldenAlgorithmsFile types
- `testdata/golden/algorithms.json` — 4 entries sorted alphabetically; canonical byte form (2-space indent, trailing LF, no BOM)
- `testdata/golden/_staging/levenshtein.json` — Same 4 entries; produced by Wave 1 for uniform plan 02-07 inputs
- `testdata/fuzz/FuzzLevenshteinScore/seed-001` — kitten/sitting pair in go test fuzz v1 format
- `tests/bdd/bdd_test.go` — godog TestBDDSuite runner with goleak.VerifyTestMain
- `tests/bdd/features/levenshtein.feature` — 5 scenarios: Scenario Outline (3 reference vectors), identical=1.0, both-empty=1.0, one-empty=0.0, symmetry
- `tests/bdd/steps/algorithms_steps.go` — AlgorithmContext + InitializeScenario; Wave 2 extension point
- `algoid_test.go` (modified) — Replaced TestDispatch_AllNilAtPhase1 with TestDispatch_LevenshteinRegistered + TestDispatch_UnregisteredSlotsAreNil
- `llms.txt` (modified) — Added 4 Levenshtein symbols to public API section (required by ai_friendly_test.go gate)

## Decisions Made

- maxStackInputLen=64 declared in levenshtein.go (not algoid.go) — closest to the implementation; Wave 2 DP algorithms reference it from here
- Dispatch registration uses separate dispatch_<algo>.go files per algorithm to avoid any Wave 2 merge conflicts on shared files
- assertGoldenStaging creates parent directories (MkdirAll) so Wave 2 plans don't need to pre-create _staging/
- levenshtein_test.go uses absFloat64 helper (inline) rather than importing math solely for math.Abs, keeping import list minimal; math is already imported in props_test.go
- llms.txt updated as part of Task 3 after ai_friendly_test.go gate surfaced the gap (Rule 2: auto-add missing critical functionality — the gate is a correctness requirement)

## Deviations from Plan

### Auto-fixed Issues

**1. [Rule 2 - Missing Critical] Updated llms.txt to include Levenshtein symbols**
- **Found during:** Task 3 (make check quality gate)
- **Issue:** TestAIFriendly_LLMSTxtReferencesEveryExportedSymbol failed — 4 new exported symbols (LevenshteinDistance, LevenshteinDistanceRunes, LevenshteinScore, LevenshteinScoreRunes) were absent from llms.txt
- **Fix:** Added "Levenshtein edit-distance similarity" section to llms.txt public API listing
- **Files modified:** llms.txt
- **Verification:** go test -run TestAIFriendly -count=1 ./... exits 0
- **Committed in:** 94bf93f (Task 3 commit)

**2. [Rule 1 - Bug] Fixed misspelling "artifacts" → "artefacts" in levenshtein_test.go**
- **Found during:** Task 3 (golangci-lint misspell linter)
- **Issue:** `misspell` linter flagged "artifacts" (American spelling) vs project's British English convention ("artefacts")
- **Fix:** Changed the comment spelling to "artefacts"
- **Files modified:** levenshtein_test.go
- **Verification:** golangci-lint run ./... exits 0 with 0 issues
- **Committed in:** 94bf93f (Task 3 commit)

---

**Total deviations:** 2 auto-fixed (1 missing critical, 1 bug)
**Impact on plan:** Both auto-fixes necessary for quality gate compliance. No scope creep.

## Issues Encountered

- `gofmt -s` reformatted several files after initial creation (alignment in struct literals, buf declaration). Fixed with a single `gofmt -s -w` pass before the Task 3 commit.
- `goimports` reformatted import order in `tests/bdd/steps/algorithms_steps.go` (third-party before local module). Fixed with `goimports -local github.com/axonops/fuzzymatch -w`.

## Next Phase Readiness

Wave 2 (plans 02-02 through 02-06 in parallel) can now start. The canonical pattern is established and verified:
- All five Wave 2 plans can copy the algorithm file pattern without re-deciding any detail
- assertGoldenStaging is deployed with the locked signature
- props_test.go, example_test.go, and tests/bdd/steps/algorithms_steps.go are extend-only append points
- dispatch_levenshtein.go is the exact template for dispatch_{hamming,jaro,jarowinkler,damerau_osa,damerau_full}.go

No blockers. make check exits 0. Coverage 97.6%.

---
*Phase: 02-core-character-algorithms-six*
*Completed: 2026-05-14*
