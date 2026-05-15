---
phase: 07-phonetic-algorithms
plan: "05"
subsystem: testing
tags: [go, phonetic, golden-merge, benchmarks, identifier-similarity, bdd, llms-sync]

# Dependency graph
requires:
  - phase: 07-phonetic-algorithms
    provides: "Four phonetic algorithms (Soundex, DoubleMetaphone, NYSIIS, MRA) shipped in plans 07-01..07-04 with staging golden files, BDD features, permittedMongeElkanInner at 18 entries"

provides:
  - "testdata/golden/algorithms.json merged from 19 to 23 algorithms × 184 entries (byte-stable)"
  - "bench.txt regenerated full-replace including 4 new phonetic benchmark families"
  - "examples/identifier-similarity extended from 19 to 23 columns (Soundex, DblMetaph, NYSIIS, MRA)"
  - "examples/phonetic-keys NEW educational program demonstrating 5 encoded-key surfaces + MRACompare"
  - "tests/bdd/features/monge_elkan_phonetic_inner.feature NEW BDD coverage for ME-over-phonetic composition"
  - "ai_friendly_test.go verified covering all 9 Phase 7 exported symbols"

affects: [phase-08-scorer, phase-09-scan, phase-10-extract]

# Tech tracking
tech-stack:
  added: []
  patterns:
    - "Staging golden merge pattern: TestGolden_Algorithms_Merge now includes all 23 staging files"
    - "Example module isolation: phonetic-keys uses its own go.mod with replace directive (mirroring identifier-similarity)"
    - "BDD ME-phonetic inner pattern: Scenario Outline with {a, b, score} examples columns"

key-files:
  created:
    - examples/phonetic-keys/main.go
    - examples/phonetic-keys/main_test.go
    - examples/phonetic-keys/go.mod
    - examples/phonetic-keys/go.sum
    - tests/bdd/features/monge_elkan_phonetic_inner.feature
  modified:
    - testdata/golden/algorithms.json
    - testdata/golden/_staging/nysiis.json
    - testdata/golden/_staging/mra.json
    - algorithms_golden_test.go
    - bench.txt
    - examples/identifier-similarity/main.go
    - examples/identifier-similarity/main_test.go

key-decisions:
  - "NYSIIS and MRA staging files required reformatting from plain JSON arrays to proper {version, entries} schema with name fields before merge"
  - "phonetic-keys example uses %s (not %-8s) for last column to avoid trailing spaces in golden stdout fixture"
  - "phonetic-keys table separator uses 52 dashes to match actual header width (not the original 60)"
  - "BDD monge_elkan_phonetic_inner uses 'approximately X within 0.0001' step for 0.5 score assertions (not exact) to match existing ME step convention"

patterns-established:
  - "Algorithm golden merge: add staging file path to TestGolden_Algorithms_Merge stagingFiles slice"
  - "Example column trailing-space avoidance: use %s for last column instead of %-8s"

requirements-completed: [PHON-01, PHON-02, PHON-03, PHON-04]

# Metrics
duration: 120min
completed: "2026-05-15"
---

# Phase 7 Plan 05: Phonetic Algorithms Finalisation Summary

**algorithms.json extended to 23 algorithms × 184 entries; bench.txt regenerated with 250 phonetic benchmark lines; identifier-similarity extended to 23 columns; new phonetic-keys educational example; ME-phonetic-inner BDD coverage added**

## Performance

- **Duration:** ~120 min (dominated by bench.txt regeneration: ~30 min full suite)
- **Started:** 2026-05-15T18:00:00Z (approximate)
- **Completed:** 2026-05-15T19:56:28Z
- **Tasks:** 2 of 2
- **Files modified/created:** 12

## Accomplishments

- Merged all 4 phonetic staging goldens into `algorithms.json`, extending from 19 algorithms × 144 entries to 23 algorithms × 184 entries; cross-platform byte-stable via canonical marshal
- Regenerated `bench.txt` full-replace with 1296 lines including 250 lines of Soundex/DoubleMetaphone/NYSIIS/MRA benchmarks; PASS, no >10% regression on carry-forward
- Extended `examples/identifier-similarity/main.go` from 19 to 23 columns (4 phonetic score columns appended); golden stdout fixture regenerated in `main_test.go`
- Shipped NEW `examples/phonetic-keys/main.go` + `main_test.go` educational program demonstrating SoundexCode/DoubleMetaphoneKeys/NYSIISCode/MRACode/MRACompare on 12 curated English surnames
- Shipped NEW `tests/bdd/features/monge_elkan_phonetic_inner.feature` with 6 scenarios covering ME-over-{Soundex, DoubleMetaphone, NYSIIS, MRA} binary-inner composition per CONTEXT.md §4 LOCKED
- Verified `ai_friendly_test.go` (TestAIFriendly_LLMSTxtReferencesEveryExportedSymbol) passes covering all 9 Phase 7 exported symbols already synced per-plan in 07-01..07-04

## Task Commits

Each task was committed atomically:

1. **Task 2: identifier-similarity extension + phonetic-keys example + ME-phonetic BDD** - `2be26fa` (feat)
2. **Task 1: merge staging goldens + regenerate bench.txt** - `c18d69b` (feat)

Note: Task 2 was committed first since bench.txt regeneration (Task 1 dependency) takes ~30 minutes; Task 1 committed immediately after bench completed.

## Files Created/Modified

- `testdata/golden/algorithms.json` - Merged from 19 to 23 algorithms × 184 entries
- `testdata/golden/_staging/nysiis.json` - Reformatted from plain array to proper {version, entries} schema
- `testdata/golden/_staging/mra.json` - Reformatted from plain array to proper {version, entries} schema
- `algorithms_golden_test.go` - Added double_metaphone, mra, nysiis, soundex to TestGolden_Algorithms_Merge
- `bench.txt` - Full-replace with 1296 lines including 4 new phonetic benchmark families
- `examples/identifier-similarity/main.go` - Extended from 19 to 23 columns; doc text updated
- `examples/identifier-similarity/main_test.go` - Golden stdout want const regenerated for 23-column table
- `examples/phonetic-keys/main.go` - NEW: educational program demonstrating encoded-key surfaces
- `examples/phonetic-keys/main_test.go` - NEW: golden stdout fixture + TestExample_Output + TestExample_ColumnWidths
- `examples/phonetic-keys/go.mod` - NEW: module definition with replace directive
- `examples/phonetic-keys/go.sum` - NEW: dependency checksums
- `tests/bdd/features/monge_elkan_phonetic_inner.feature` - NEW: 6 BDD scenarios for ME-phonetic composition

## Decisions Made

- **NYSIIS/MRA staging reformatting:** Plans 07-03 and 07-04 wrote staging files as plain JSON arrays (missing `version` and `name` fields). This task reformatted them to match the proper `goldenAlgorithmsFile{Version, Entries[]goldenAlgorithmEntry}` schema used by `TestGolden_Algorithms_Merge`. The schema requires `name`, `algorithm`, `a`, `b`, `expected_score` fields per entry.
- **phonetic-keys trailing spaces:** The `%-8s` format string for the last table column produces trailing spaces that would make the golden stdout fixture fragile. Used `%s` for the last column to avoid trailing whitespace.
- **phonetic-keys separator width:** Separator initially set to 60 dashes; adjusted to 52 to match actual header width (12+1+8+1+8+1+8+1+8+1+3 = 52 chars). `TestExample_ColumnWidths` enforces header-separator width parity.
- **BDD 0.5 score assertion:** Used `approximately 0.5 within 0.0001` step rather than `exactly 0.5` to match the existing MongeElkan BDD step convention. The `approximately` step was already registered from Phase 6.

## Deviations from Plan

### Auto-fixed Issues

**1. [Rule 1 - Bug] NYSIIS and MRA staging files had wrong JSON schema**
- **Found during:** Task 1 (merge staging goldens)
- **Issue:** `testdata/golden/_staging/nysiis.json` and `mra.json` were plain JSON arrays (no `version` field, no `name` fields per entry). `TestGolden_Algorithms_Merge` tries to unmarshal into `goldenAlgorithmsFile{Version int, Entries []goldenAlgorithmEntry}` and requires `name` field on each entry for deduplication check.
- **Fix:** Reformatted both files to proper `{version: 1, entries: [{name, algorithm, a, b, expected_score}, ...]}` schema; generated canonical `name` values matching the naming pattern from other staging files (e.g., `NYSIIS_Brown_Browne_match`, `MRA_Byrne_Boern_match`).
- **Files modified:** `testdata/golden/_staging/nysiis.json`, `testdata/golden/_staging/mra.json`
- **Verification:** `go test -run TestGolden_Algorithms_Merge -update ./...` PASS; re-run without `-update` PASS.
- **Committed in:** c18d69b (Task 1 commit)

**2. [Rule 3 - Blocking] phonetic-keys example needed go.mod + go.sum**
- **Found during:** Task 2 (phonetic-keys example)
- **Issue:** `examples/phonetic-keys/` is a separate module (analog of `examples/identifier-similarity/`) and requires its own `go.mod` with `replace` directive pointing to the root package. Without it, `go test ./...` from the root doesn't find the module and `go test ./examples/phonetic-keys/...` fails.
- **Fix:** Created `go.mod` mirroring `examples/identifier-similarity/go.mod` structure; created `go.sum` with the same `golang.org/x/text` checksum.
- **Files modified:** `examples/phonetic-keys/go.mod` (created), `examples/phonetic-keys/go.sum` (created)
- **Verification:** `go test ./...` in phonetic-keys directory PASS.
- **Committed in:** 2be26fa (Task 2 commit)

---

**Total deviations:** 2 auto-fixed (1 Rule 1 bug fix, 1 Rule 3 blocking fix)
**Impact on plan:** Both auto-fixes necessary for correctness. No scope creep.

## Issues Encountered

- bench.txt regeneration took ~30 minutes (full suite with -count=10). Task 2 was committed first to avoid blocking on bench completion while other work was ready.
- The absolute-path safety requirement (#3099) for worktree operations was essential: all file edits had to go to the worktree path (`/Users/johnny/Development/fuzzymatch/.claude/worktrees/agent-a279c34f80cb0edb2/`) not the main repo path (`/Users/johnny/Development/fuzzymatch/`). Initial edits were accidentally made to the main repo and had to be redone in the worktree.

## User Setup Required

None - no external service configuration required.

## Next Phase Readiness

- Phase 7 complete: all 4 phonetic algorithms (PHON-01..PHON-04) shipped end-to-end
- `algorithms.json` at 23 algorithms × 184 entries; cross-platform byte-stable
- `bench.txt` regenerated with 4 new phonetic benchmark families
- `examples/identifier-similarity` extended to 23 columns; ready to reference in README
- `examples/phonetic-keys` new educational program; ready to reference in README
- 5 Phase 7 BDD feature files: soundex, double_metaphone, nysiis, mra, monge_elkan_phonetic_inner
- permittedMongeElkanInner at FINAL 18 entries; rejected slice at FINAL 5 entries
- Phase 8 (Composite Scorer) can proceed; no blockers from Phase 7

## Self-Check: PASSED

All claimed artifacts verified:

- FOUND: testdata/golden/algorithms.json (23 algorithms × 184 entries)
- FOUND: bench.txt (1296 lines, 250 phonetic benchmark lines, PASS)
- FOUND: examples/identifier-similarity/main.go (DblMetaph column present)
- FOUND: examples/identifier-similarity/main_test.go
- FOUND: examples/phonetic-keys/main.go (MRACompare present)
- FOUND: examples/phonetic-keys/main_test.go
- FOUND: tests/bdd/features/monge_elkan_phonetic_inner.feature
- FOUND: .planning/phases/07-phonetic-algorithms/07-05-SUMMARY.md
- FOUND commit: 2be26fa (Task 2)
- FOUND commit: c18d69b (Task 1)

All tests pass: `go test -run 'TestGolden|TestAIFriendly|TestPhonetic|TestMongeElkan|TestExample' ./...` → ok
BDD tests pass: `make test-bdd` → ok
verify-determinism: `make verify-determinism` → ok
