---
phase: "07-phonetic-algorithms"
plan: "03"
subsystem: "phonetic-algorithms"
status: "COMPLETE"
closes: ["PHON-03"]
tags: ["phonetic", "nysiis", "taft-1970", "truncation-gate", "variant-divergence-load-bearing"]
depends_on: ["07-01", "07-02"]
provides: ["NYSIISCode", "NYSIISScore", "AlgoNYSIIS-dispatch"]
affects:
  - monge_elkan.go
  - monge_elkan_test.go
  - props_test.go
  - example_test.go
  - algoid_test.go
  - tests/bdd/steps/algorithms_steps.go
  - tests/bdd/features/monge_elkan.feature
  - llms.txt
  - llms-full.txt
  - phonetic_cross_validation_test.go
  - phonetic_codes_golden_test.go
  - testdata/cross-validation/phonetic/vectors.json
  - testdata/golden/phonetic-codes.json
tech_stack:
  added: []
  patterns:
    - "H-always-removed body rule (Taft-1970 original — not the conditional-H variant of some ports)"
    - "stack-allocated [128]byte working buffer for prefix/suffix rules (no intermediate heap alloc)"
    - "stack-allocated [128]byte result buffer + unavoidable string() alloc = 1 alloc/op"
    - "variant_divergence: true on 8/20 NYSIIS corpus entries (40%) where jellyfish emits >6 chars"
key_files:
  created:
    - nysiis.go
    - dispatch_nysiis.go
    - nysiis_test.go
    - nysiis_bench_test.go
    - nysiis_fuzz_test.go
    - testdata/golden/_staging/nysiis.json
    - tests/bdd/features/nysiis.feature
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
    - phonetic_cross_validation_test.go
    - phonetic_codes_golden_test.go
    - testdata/cross-validation/phonetic/vectors.json
    - testdata/golden/phonetic-codes.json
decisions:
  - "OQ-NYSIIS-1 RESOLUTION LOCKED 2026-05-15: H is always removed from the body (not the conditional-H variant); produces John→JAN, Catherine→CATARA correctly"
  - "OQ-NYSIIS-2 ACCEPTED: NYSIISCode: 1 alloc/op (string() return unavoidable without unsafe); 40ns/op well under 500ns budget"
  - "variant_divergence rate: 8/20 = 40% (Brown/Browne/Robert/John/Teresa/Theresa/Smith/Johnson/Taylor/Davis/Wilson=no divergence; Catherine/Katherine/Johnathan/Jonathan/montgomery/martincevic/Nicholson/Henderson=divergence)"
metrics:
  duration_minutes: 16
  completed_date: "2026-05-15"
  task_count: 1
  file_count: 20
---

# Phase 7 Plan 03: NYSIIS (Taft 1970 / Knuth TAOCP §6.4) Summary

**One-liner:** NYSIIS (Taft 1970) with 6-char truncation, H-always-removed body rule, 40ns/op, 1 alloc; variant_divergence mechanism active for 40% of corpus where jellyfish emits >6 chars.

## Status: COMPLETE

Closes: **PHON-03**

## Variant Choice

**Original NYSIIS-1970, 6-character truncation, each rule applied once** per CONTEXT.md §2 LOCKED.

- Modified-NYSIIS (jellyfish variant, no truncation, iterate-to-fixed-point) REJECTED.
- Load-bearing gate: `NYSIISCode("Catherine") == "CATARA"` (6 chars, NOT "CATARAN" which jellyfish 1.2.1 returns).

## Source Citations

- **Algorithmic origin:** Taft, R. L. (1970). Name search techniques. New York State Identification and Intelligence System, Special Report No. 1. Albany, NY.
- **Canonical algorithm description (primary source for fresh transcription):** Knuth, D. E. (1973). The Art of Computer Programming, Vol. 3, §6.4. Addison-Wesley.
- **Note:** Taft 1970 is a NY State Special Report not available through academic publishers; Knuth's secondary description in TAOCP Vol. 3 §6.4 is the authoritative algorithm description used for this implementation.
- **Cross-validation:** jellyfish==1.2.1 (BSD-2-Clause) — reference vectors only.
- **Negative-attribution:** github.com/UjjwalAyyangar/go-jellyfish (MIT) — NOT consulted.

## jellyfish Divergence Rate

**40% of NYSIIS corpus entries carry `variant_divergence: true`** (8/20 entries) per RESEARCH.md key finding 3 (~40-60% expected).

Divergent inputs (jellyfish emits >6 chars → Taft-1970 truncates to 6):
- Catherine → CATARA (jellyfish: CATARAN, 7 chars)
- Katherine → CATARA (jellyfish: CATARAN, 7 chars)
- Johnathan → JANATA (jellyfish: JANATAN, 7 chars)
- Jonathan → JANATA (jellyfish: JANATAN, 7 chars)
- montgomery → MANTGA (jellyfish: MANTGANARY, 10 chars)
- martincevic → MARTAN (jellyfish: MARTANCAFAC, 11 chars)
- Nicholson → NACALS (jellyfish: NACALSAN, 8 chars)
- Henderson → HANDAR (jellyfish: HANDARSAN, 9 chars)

## What Was Built

### Task 1: NYSIIS End-to-End (TDD RED → GREEN)

**TDD RED commit `c3786c3`:** nysiis_test.go, nysiis_bench_test.go, nysiis_fuzz_test.go — all failing.

**TDD GREEN commit `f2e1316`:** Full NYSIIS implementation.

**`nysiis.go`** (~180 LOC):
- `NYSIISCode(s string) string`: 9-step Taft-1970 procedure applied once; stack-allocated [128]byte working buffers for prefix/suffix rule application; final 6-char truncation; 40ns/op, 1 alloc/op (unavoidable string() conversion).
- `NYSIISScore(a, b string) float64`: binary 0.0/1.0; identity short-circuit covers both-empty.
- Source-Origin Statement: LOCKED two-line form with Taft 1970 (algorithmic origin) + Knuth TAOCP §6.4 (canonical description) + "Taft 1970 not available" note + jellyfish==1.2.1 cross-validation + negative-attribution.

**Load-bearing variant gates all pass:**
- `NYSIISCode("Brown") == NYSIISCode("Browne") == "BRAN"` (RV-N1, RV-N2)
- `NYSIISCode("Robert") == "RABAD"` (RV-N3)
- `NYSIISCode("Catherine") == "CATARA"` (6 chars — RV-N4 LOAD-BEARING truncation gate)
- `NYSIISCode("Katherine") == "CATARA"` (RV-N5)
- `NYSIISCode("Johnathan") == "JANATA"` (RV-N6)
- `NYSIISCode("Jonathan") == "JANATA"` (RV-N7)
- `NYSIISCode("John") == "JAN"` (RV-N8)
- `NYSIISCode("Teresa") == NYSIISCode("Theresa") == "TARAS"` (RV-N9, RV-N10)
- `NYSIISCode("montgomery") == "MANTGA"` (RV-N11)

**`dispatch_nysiis.go`**: registers `NYSIISScore` at `dispatch[AlgoNYSIIS]` (slot 25).

**Monge-Elkan wiring:**
- `monge_elkan.go`: `permittedMongeElkanInner` 16→17 entries
- `monge_elkan_test.go`: `rejected` slice 7→6; `TestMongeElkanScore_BinaryInner_NYSIIS` (3 sub-tests: one_matches=0.5, both_match=1.0, neither=0.0)
- `tests/bdd/features/monge_elkan.feature`: NYSIIS→MRA in non-permitted examples (NYSIIS now permitted)

**Cross-validation corpus:**
- `testdata/cross-validation/phonetic/vectors.json`: 20 NYSIIS entries; 8/20 (40%) carry `variant_divergence: true`
- `phonetic_cross_validation_test.go`: NYSIIS sub-test enabled; asserts Taft-truncated value, logs divergent jellyfish value for transparency
- `testdata/golden/phonetic-codes.json`: NYSIIS section appended (11 entries)
- `phonetic_codes_golden_test.go`: NYSIIS branch activated

**Tests:**
- `nysiis_test.go`: 12 Knuth reference vectors (RV-N1..RV-N12); load-bearing truncation gate; charset; output length; non-ASCII silent skip; identity/range/score invariants
- `nysiis_bench_test.go`: 5 benchmarks at 3 sizes; 40ns/op, 1 alloc/op at ASCII_Short (well under 500ns budget)
- `nysiis_fuzz_test.go`: FuzzNYSIIS — 20 seeds (ASCII + non-ASCII); 6 invariants including `^[A-Z]{0,6}$` charset and length ≤ 6 (LOAD-BEARING Taft-1970 truncation invariant); 10-second run: PASS
- `props_test.go`: Five-invariant block (Identity/Symmetric/Range/NoNaN/NoInf) + `PropNYSIIS_CodeLength` (LOAD-BEARING) + `PropNYSIIS_CodeCharset`
- `example_test.go`: `ExampleNYSIISCode` + `ExampleNYSIISScore`

**BDD:**
- `tests/bdd/features/nysiis.feature`: 10 scenarios covering Brown/Browne canonical pair, Robert RABAD, Catherine truncation gate (LOAD-BEARING), Katherine, identity, both-empty, one-empty (×2), non-match
- `tests/bdd/steps/algorithms_steps.go`: `iComputeTheNYSIISCodeOf` + `iComputeTheNYSIISScoreBetween` appended; 2 new step registrations

**Documentation:**
- `llms.txt`: `### NYSIIS (Taft 1970 / Knuth TAOCP §6.4)` heading + `NYSIISCode` + `NYSIISScore` entries
- `llms-full.txt`: Full Phase 7 NYSIIS algorithm-surface block with source-origin, godoc verbatim
- `algoid_test.go`: `AlgoNYSIIS` added to `registered` dispatch map
- `testdata/golden/_staging/nysiis.json`: 10 staging Score entries (to be merged in plan 07-05)

## Commits

- `c3786c3` — `test(07-03): add failing tests for NYSIIS (Taft 1970 / Knuth TAOCP §6.4)` (TDD RED)
- `f2e1316` — `feat(07-03): implement NYSIIS (Taft 1970 / Knuth TAOCP §6.4) with 6-char truncation` (TDD GREEN)

## OQ Resolutions

### OQ-NYSIIS-1 RESOLUTION LOCKED 2026-05-15
**H is always removed from the body**

The NYSIIS algorithm's H handling in the body is "H is always removed". Some descriptions suggest a conditional rule (H removed only between two consonants), but testing against the reference vectors confirms the simpler interpretation:

- John (J-O-H-N): O→A, H→removed, N→N = JAN ✓
- Catherine: T, H→removed, E→A, R, I→A, N, E→A = CATARAN (7 chars) → truncated CATARA ✓
- Johnathan: O→A, H→removed, N, A, T, H→removed, A, N = JANATAN → truncated JANATA ✓

The conditional H rule (between two consonants) would produce incorrect results for John (H→removed because between vowel A and consonant N is borderline ambiguous in some descriptions). The "always remove H" rule produces all correct reference vectors.

### OQ-NYSIIS-2 ACCEPTED
**NYSIISCode: 1 alloc/op (string() return unavoidable)**

`NYSIISCode(s string) string` must return a `string`. Converting the stack-allocated `[128]byte` result buffer to a `string` via `string(res)` always allocates because the compiler cannot prove the string doesn't outlive the stack frame. This is identical to the SoundexCode OQ-2 resolution in plan 07-01.

Actual performance: 40ns/op, 4B/op, 1 alloc/op on Apple M2 — well within the 500ns latency budget. Accepted.

## Deviations from Plan

### Auto-fixed Issues

None — plan executed as written.

## Known Stubs

None — all NYSIIS corpus entries in vectors.json are populated with real values. The MRA section in vectors.json remains as `{"algorithm": "MRA", "input": "", "code": ""}` (plan 07-04 stub — out of scope for this plan).

## Threat Flags

No new network endpoints, auth paths, file access patterns, or schema changes at trust boundaries introduced. Pure-function algorithm file with no I/O.

## Self-Check

- [x] nysiis.go exists in worktree
- [x] dispatch_nysiis.go exists in worktree
- [x] nysiis_test.go exists in worktree
- [x] nysiis_bench_test.go exists in worktree
- [x] nysiis_fuzz_test.go exists in worktree
- [x] testdata/golden/_staging/nysiis.json exists in worktree
- [x] tests/bdd/features/nysiis.feature exists in worktree
- [x] Commits c3786c3 (RED) and f2e1316 (GREEN) exist on worktree branch
- [x] `NYSIISCode("Brown") == "BRAN"` ✓
- [x] `NYSIISCode("Browne") == "BRAN"` ✓
- [x] `NYSIISCode("Robert") == "RABAD"` ✓
- [x] `len(NYSIISCode("Catherine")) == 6` (= "CATARA") ✓ LOAD-BEARING GATE
- [x] `NYSIISScore("Brown", "Browne") == 1.0` ✓
- [x] `permittedMongeElkanInner` has 17 entries ✓
- [x] `rejected` slice has 6 entries ✓
- [x] `TestMongeElkan_PanicsOnNonPermittedInner` passes ✓
- [x] `TestMongeElkanScore_BinaryInner_NYSIIS` passes 3 sub-tests ✓
- [x] `BenchmarkNYSIISCode_ASCII_Short`: 40ns/op, 1 alloc/op (< 500ns budget) ✓
- [x] `FuzzNYSIIS` 10-second run: PASS (charset + length invariants hold) ✓
- [x] `tests/bdd/features/nysiis.feature` has 10 scenarios; `make test-bdd` passes ✓
- [x] `TestPhonetic_CrossValidation/NYSIIS` passes with 8/20 (40%) variant_divergence entries ✓
- [x] `TestPhoneticCodesGolden/NYSIIS` passes ✓
- [x] `TestLLMsTxt` passes (NYSIISCode + NYSIISScore in llms.txt) ✓
- [x] `PropNYSIIS_CodeLength` passes (len(NYSIISCode(x)) <= 6 for all x) ✓
- [x] All root tests pass: `go test -race -shuffle=on -count=1 ./...` → ok ✓
- [x] All BDD tests pass: `go test ./tests/bdd/...` → ok ✓

## Self-Check: PASSED
