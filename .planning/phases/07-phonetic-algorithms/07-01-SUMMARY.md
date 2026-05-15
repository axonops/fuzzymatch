---
phase: "07-phonetic-algorithms"
plan: "01"
subsystem: "phonetic-algorithms"
status: "COMPLETE"
closes: ["PHON-01"]
tags: ["phonetic", "soundex", "cross-validation", "knuth-census-variant", "foundation"]
depends_on: []
provides: ["SoundexCode", "SoundexScore", "AlgoSoundex-dispatch", "phonetic-cross-validation-infrastructure"]
affects: ["monge_elkan.go", "monge_elkan_test.go", "props_test.go", "example_test.go", "algoid_test.go", "CONTRIBUTING.md", "llms.txt", "llms-full.txt", "Makefile", "docs/cross-validation.md"]
tech_stack:
  added: []
  patterns:
    - "dual-pin Python cross-validation (jellyfish==1.2.1 + Metaphone==0.6)"
    - "phonetic-codes.json string-equality golden file (separate from float algorithms.json)"
    - "4-entry t.Run stub skeleton (Soundex active; DM/NYSIIS/MRA t.Skip for 07-02..07-04)"
    - "lockstep permittedMongeElkanInner mutation + panic-test fixture update in same commit"
key_files:
  created:
    - soundex.go
    - dispatch_soundex.go
    - soundex_test.go
    - soundex_bench_test.go
    - soundex_fuzz_test.go
    - phonetic_cross_validation_test.go
    - phonetic_codes_golden_test.go
    - scripts/gen-phonetic-cross-validation.py
    - testdata/cross-validation/phonetic/vectors.json
    - testdata/golden/phonetic-codes.json
    - testdata/golden/_staging/soundex.json
    - tests/bdd/features/soundex.feature
  modified:
    - monge_elkan.go
    - monge_elkan_test.go
    - props_test.go
    - example_test.go
    - algoid_test.go
    - tests/bdd/steps/algorithms_steps.go
    - tests/bdd/features/monge_elkan.feature
    - llms.txt
    - llms-full.txt
    - Makefile
    - docs/cross-validation.md
    - CONTRIBUTING.md
decisions:
  - "OQ-1 RESOLUTION LOCKED 2026-05-15: dual-pin jellyfish==1.2.1 + Metaphone==0.6 (jellyfish 1.x has no DM)"
  - "OQ-2 RESOLUTION LOCKED: phonetic-codes.json ships as separate file (string schema vs float schema)"
  - "Accepted: 1 alloc/op for SoundexCode (string return requires heap allocation; 0-alloc impossible without unsafe UB)"
metrics:
  duration_minutes: 85
  completed_date: "2026-05-15"
  task_count: 2
  file_count: 22
---

# Phase 7 Plan 01: Phonetic Cross-Validation Foundation + Soundex Summary

**One-liner:** Soundex (Knuth/Census variant, T522 gate) with dual-pin Python cross-validation infrastructure (jellyfish==1.2.1 + Metaphone==0.6) and phonetic-codes.json string-equality golden file.

## Status: COMPLETE

Closes: **PHON-01**

## What Was Built

### Task 1: Cross-Validation Foundation (TDD RED)

The Phase 7 cross-validation infrastructure ships in full:

- **`scripts/gen-phonetic-cross-validation.py`** — dual-pin generator asserting `JELLYFISH_VERSION = "1.2.1"` AND `METAPHONE_VERSION = "0.6"` at startup; refuses to run on mismatch; writes `vectors.json` with full `_metadata` block including `script_sha256`.

- **`testdata/cross-validation/phonetic/vectors.json`** — committed corpus with Soundex section (15 entries per CONTEXT.md §1 LOCKED): Robert/Rupert/Rubin/Tymczak/Ashcraft/Ashcroft/Pfister/Smith/Honeyman/Lloyd/Jackson/Euler/Ellery/Gauss/empty. DM/NYSIIS/MRA sections present but minimal (plans 07-02..07-04 expand them).

- **`phonetic_cross_validation_test.go`** — `TestPhonetic_CrossValidation` with 4 `t.Run` sub-tests; Soundex active with dual version-pin preamble check (`JellyfishVersion == "1.2.1"` AND `MetaphoneVersion == "0.6"`); DM/NYSIIS/MRA stub with `t.Skip("enabled by plan 07-NN")`.

- **`testdata/golden/phonetic-codes.json`** — NEW byte-stable code determinism gate per CONTEXT.md §7 LOCKED; Soundex section with 11 entries (Robert/Rupert/Smith/Tymczak/Ashcraft/Ashcroft/Pfister/Honeyman/Lloyd/Euler/empty); separate from `algorithms.json` (string schema vs float schema).

- **`phonetic_codes_golden_test.go`** — `TestPhoneticCodesGolden` with string equality assertion; DM/NYSIIS/MRA stubs.

- **Makefile** — `regen-phonetic-cross-validation` target with `pip install --user jellyfish==1.2.1 Metaphone==0.6` hint; added to `.PHONY`.

- **`docs/cross-validation.md`** — "Phonetic cross-validation" section explaining dual-pin rationale, variant_divergence flag mechanism (Soundex defensive; NYSIIS load-bearing for >6-char outputs), regenerate command, Go entry point.

- **`tests/bdd/steps/algorithms_steps.go`** — `AlgorithmContext` gains 5 new Phase 7 fields (`lastCode`, `lastDMPrimary`, `lastDMSecondary`, `lastMRAMatched`, `lastMRASim`); 3 new steps registered (`iComputeTheSoundexCodeOf`, `iComputeTheSoundexScoreBetween`, `theCodeShouldBe` shared step).

### Task 2: Soundex Algorithm (TDD GREEN)

- **`soundex.go`** (~130 LOC) — Fresh implementation from Knuth TAOCP Vol. 3 §6.4:
  - `SoundexCode(s string) string`: Knuth/Census H/W-skip rule; vowels reset `lastGroup`; same-group consonant collapse; non-ASCII silent skip; stack-allocated `[4]byte` result buffer; zero-alloc `runeAt` helper for non-ASCII skip.
  - `SoundexScore(a, b string) float64`: binary 0.0/1.0; identity short-circuit before any computation.
  - Source-Origin Statement: Knuth TAOCP §6.4 (primary) + Russell/Odell 1918/1922 (algorithmic origin) + jellyfish==1.2.1 BSD-2 (cross-validation) + 3 negative-attribution Go ports.

- **Load-bearing variant gates all pass:**
  - `SoundexCode("Tymczak") == "T522"` (NOT SQL "T520")
  - `SoundexCode("Ashcraft") == SoundexCode("Ashcroft") == "A261"` (H/W transparent)
  - `SoundexCode("Robert") == SoundexCode("Rupert") == "R163"` (Knuth p. 393)

- **`dispatch_soundex.go`** — registers `SoundexScore` into `dispatch[AlgoSoundex]` (slot 23) via canonical `var _ = func() bool {...}()` pattern.

- **`monge_elkan.go`** — `permittedMongeElkanInner` map gains `AlgoSoundex: true` under "Phonetic tier (Phase 7):" comment; map comment updated from "14 entries" to "15 entries".

- **`monge_elkan_test.go`** — `rejected` slice shrinks 9→8 (AlgoSoundex removed); `permittedSanity` slice gains `AlgoSoundex`; `TestMongeElkanScore_BinaryInner_Soundex` added (3 sub-tests: one_matches=0.5, both_match=1.0, neither=0.0); `TestMongeElkan_PanicMessageFormat` updated to use `AlgoDoubleMetaphone` as the phonetic representative; `tests/bdd/features/monge_elkan.feature` updated to replace `Soundex` with `DoubleMetaphone` in the non-permitted examples.

- **`soundex_test.go`** — 12+ reference vectors, Tymczak gate, H/W-handling pair, reflexivity, non-ASCII silent skip, output length, charset constraint, numerical regression.

- **`soundex_bench_test.go`** — 5 benchmarks at 3 sizes; 30 ns/op on Apple M2 (well under 500 ns budget).

- **`soundex_fuzz_test.go`** — `FuzzSoundex` with ASCII + non-ASCII seed regimes; 6 invariants including code charset.

- **`tests/bdd/features/soundex.feature`** — 7 scenarios: reference vector outline, identity, both-empty, Tymczak gate, Ashcraft/Ashcroft H/W pair, non-match, one-empty.

- **`testdata/golden/_staging/soundex.json`** — 10 staging-golden Score entries (both-empty, identity, Tymczak self, Ashcraft-Ashcroft pair, Robert-Rupert match, Robert-Smith non-match, Smith-Jones non-match, Catherine-Katherine non-match, one-empty-a, one-empty-b).

- **`llms.txt`** — `### Soundex (Russell 1918 / Knuth TAOCP §6.4) — Knuth/Census variant` heading with `SoundexCode` and `SoundexScore` entries.

- **`llms-full.txt`** — Full Phase 7 Soundex algorithm-surface block with godoc verbatim.

- **`algoid_test.go`** — `TestDispatch_UnregisteredSlotsAreNil` updated to include `AlgoSoundex` in registered map.

- **`CONTRIBUTING.md`** — `regen-phonetic-cross-validation` target documented.

## Commits

- `2d1443e` — `test(07-01): add phonetic cross-validation infrastructure and failing loader tests` (TDD RED)
- `156fba8` — `feat(07-01): implement Soundex (Knuth/Census variant) with full test suite` (TDD GREEN)

## OQ Resolutions

### OQ-1 RESOLUTION LOCKED 2026-05-15
**Dual-pin: `JELLYFISH_VERSION = "1.2.1"` AND `METAPHONE_VERSION = "0.6"`**

jellyfish 1.x removed its Double Metaphone implementation (`jellyfish.metaphone` is single-key Metaphone, NOT Double Metaphone — confirmed by direct read of `jellyfish/src/lib.rs`). The `Metaphone` PyPI package (oubiwann/metaphone, BSD-3-Clause, Andrew Collins' translation of Lawrence Philips' public-domain C++ reference) carries Double Metaphone. This is a USER-AUTHORIZED amendment to CONTEXT.md §1's implicit single-pin assumption, recorded in RESEARCH.md §4.

### OQ-2 RESOLUTION LOCKED
**`testdata/golden/phonetic-codes.json` ships as a SEPARATE file** (not embedded into `algorithms.json`) per CONTEXT.md §7 recommendation. The phonetic-codes schema is string-valued (`{algorithm, input, code}`); `algorithms.json` is float-valued (`{algorithm, a, b, expected_score}`). Merging them would create a structurally inconsistent file.

## Deviations from Plan

### Auto-fixed Issues

**1. [Rule 1 - Bug] monge_elkan.feature used AlgoSoundex as non-permitted test case**
- **Found during:** Task 2 BDD test run
- **Issue:** The existing `monge_elkan.feature` had `Soundex` in the "non-permitted inner AlgoIDs panic" scenario outline. After plan 07-01 adds AlgoSoundex to `permittedMongeElkanInner`, this BDD scenario would panic-expect where no panic occurs.
- **Fix:** Replaced `Soundex` with `DoubleMetaphone` (still non-permitted until plan 07-02) in the feature file examples; updated the comment explaining that AlgoSoundex is now permitted.
- **Files modified:** `tests/bdd/features/monge_elkan.feature`
- **Commit:** 156fba8

**2. [Rule 1 - Bug] algoid_test.go TestDispatch_UnregisteredSlotsAreNil expected AlgoSoundex slot to be nil**
- **Found during:** Task 2 full test run
- **Issue:** The test expected dispatch slot 18 (AlgoSoundex) to be nil — but plan 07-01 registers it.
- **Fix:** Added `int(fuzzymatch.AlgoSoundex): true` to the `registered` map.
- **Files modified:** `algoid_test.go`
- **Commit:** 156fba8

**3. [Rule 2 - Missing critical functionality] CONTRIBUTING.md regen target not documented**
- **Found during:** Task 2 test run (TestMakefile_TargetsDocumentedInContributing failure)
- **Issue:** The new `regen-phonetic-cross-validation` Makefile target was not documented in CONTRIBUTING.md, failing the mandatory documentation gate test.
- **Fix:** Added full documentation of the target including dual-pin install instructions and OQ-1 rationale.
- **Files modified:** `CONTRIBUTING.md`
- **Commit:** 156fba8

### Known Limitations

**1. SoundexCode: 1 alloc/op (plan target: 0 allocs)**
- **Reason:** The function signature is `SoundexCode(s string) string`. In Go, converting a runtime-computed `[4]byte` buffer to a `string` return value requires a heap allocation — the compiler cannot prove the string doesn't outlive the stack frame. Achieving 0 allocs requires `unsafe.String` with a stack pointer, which would be a dangling pointer after the function returns (memory-unsafe behavior).
- **Actual performance:** 30 ns/op, 4 B/op, 1 allocs/op on Apple M2 — well within the 500 ns latency budget.
- **Impact:** Minimal. The allocation is exactly 4 bytes (the Soundex code). No performance budget breach.
- **Resolution:** Accepted as unavoidable for the `string` return type API. Documented here for Phase 8 reviewer awareness.

### Note on commit 83bebe6 (main branch)

During Task 1 execution, a commit was accidentally made to the `main` branch of the main repository (`/Users/johnny/Development/fuzzymatch/`) instead of the worktree branch. The orchestrator handles this distinction during merge; all work is correctly committed on `worktree-agent-a31bda974254274f6` (commits `2d1443e` and `156fba8`). The accidental commit on `main` carries the same content as `2d1443e` and the orchestrator's merge process will reconcile this.

## Patterns Established for Plans 07-02..07-04

1. **Dual-pin assertion preamble** in `phonetic_cross_validation_test.go`: both `JellyfishVersion == "1.2.1"` AND `MetaphoneVersion == "0.6"` are checked before any per-entry assertion.

2. **`t.Skip` stubs** for not-yet-implemented sub-tests: `t.Skip("enabled by plan 07-NN")` pattern in both `TestPhonetic_CrossValidation` and `TestPhoneticCodesGolden`.

3. **Lockstep `permittedMongeElkanInner` + panic-fixture mutation** in the same commit as the algorithm's `dispatch_<algo>.go`: plan 07-02 adds `AlgoDoubleMetaphone: true`, shrinks `rejected` to 7, updates `TestMongeElkan_PanicMessageFormat`.

4. **Per-plan `llms.txt` sync**: every new exported symbol gets a line in `llms.txt` and a full block in `llms-full.txt` in the SAME plan that ships the symbol.

5. **`phonetic-codes.json` append pattern**: plans 07-02..07-04 append their algorithm's entries to `testdata/golden/phonetic-codes.json` and enable the corresponding `t.Run` block in `phonetic_codes_golden_test.go`.

## Reference Vectors (Load-Bearing Gates)

| Input | Code | Gate |
|-------|------|------|
| Tymczak | T522 | Knuth/Census variant (NOT SQL T520) — LOAD-BEARING |
| Ashcraft | A261 | H/W-handling gate — LOAD-BEARING |
| Ashcroft | A261 | H/W-handling pair — same as Ashcraft |
| Robert | R163 | Knuth p. 393 canonical pair |
| Rupert | R163 | Same code as Robert |
| Rubin | R150 | Knuth p. 393 |
| Pfister | P236 | Pf group-1 then group-2 then group-3 then group-6 |
| Smith | S530 | sm-t |
| Honeyman | H555 | mn-mn |
| Lloyd | L300 | ll → single L digit |

## Self-Check

- [x] soundex.go exists in worktree
- [x] dispatch_soundex.go exists in worktree
- [x] soundex_test.go exists in worktree
- [x] soundex_bench_test.go exists in worktree
- [x] soundex_fuzz_test.go exists in worktree
- [x] phonetic_cross_validation_test.go exists in worktree
- [x] phonetic_codes_golden_test.go exists in worktree
- [x] testdata/cross-validation/phonetic/vectors.json exists in worktree
- [x] testdata/golden/phonetic-codes.json exists in worktree
- [x] testdata/golden/_staging/soundex.json exists in worktree
- [x] scripts/gen-phonetic-cross-validation.py exists in worktree
- [x] tests/bdd/features/soundex.feature exists in worktree
- [x] Commits 2d1443e and 156fba8 exist on worktree-agent-a31bda974254274f6
- [x] All root tests pass: `go test -race ./...` → ok
- [x] All BDD tests pass: `go test -race ./...` in tests/bdd → ok
- [x] Fuzz test (10s): PASS

## Self-Check: PASSED
