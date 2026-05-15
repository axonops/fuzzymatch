---
phase: 7
slug: phonetic-algorithms
status: draft
nyquist_compliant: false
wave_0_complete: false
created: 2026-05-15
---

# Phase 7 — Validation Strategy

> Per-phase validation contract for feedback sampling during execution.
> Synthesised from `07-RESEARCH.md` §"Validation Architecture" and the project-wide standards inherited from Phases 2–6.

---

## Test Infrastructure

| Property | Value |
|----------|-------|
| **Framework** | Go stdlib `testing` + `testing/quick` (root); godog v0.15.0 + goleak v1.3.0 + testify v1.10.0 (`tests/bdd/` only) |
| **Config file** | none — Go convention; `tests/bdd/go.mod` is the structural isolation boundary; `go.mod` (root) MUST stay zero-non-stdlib-`require` |
| **Quick run command** | `go test -race -shuffle=on -count=1 ./...` |
| **Full suite command** | `make check && make test-bdd && make verify-determinism` |
| **Cross-validation command** | `go test -run TestPhonetic_CrossValidation -race ./...` |
| **Cross-validation regen command** | `make regen-phonetic-cross-validation` (dev-only; requires `jellyfish==1.2.1` + `Metaphone==0.6` pip pins) |
| **Estimated runtime** | ~25 s (quick) / ~3 min (full `make check`) on developer hardware |

---

## Sampling Rate

- **After every task commit:** Run `go test -race -shuffle=on -count=1 ./<files-touched>` then `make fmt-check && make lint`
- **After every plan wave:** Run `make test` (full root + BDD) + `go test -run TestPhonetic_CrossValidation -race ./...`
- **Before `/gsd-verify-work`:** `make check` green AND cross-validation green AND `algorithm-licensing-reviewer` sign-off recorded in plan 07-02 PR description AND `--depth=deep` review completed on plan 07-02 AND `bench.txt` regenerated and committed
- **Max feedback latency:** ~30 s for per-task quick runs; ~3 min for full `make check`

---

## Per-Task Verification Map

> One row per public deliverable. Plan and wave columns reflect the recommended decomposition surfaced by RESEARCH.md §6 and the planner's `Claude's Discretion` window in CONTEXT.md. Final wave assignment is set in PLAN.md frontmatter; this table tracks REQ→test traceability, not commit-by-commit progress.

| Task ID | Plan | Wave | Requirement | Threat Ref | Secure Behavior | Test Type | Automated Command | File Exists | Status |
|---------|------|------|-------------|------------|-----------------|-----------|-------------------|-------------|--------|
| 07-01-01 | 01 | 1 | PHON-01 | — | `SoundexCode("Tymczak") = "T522"` — Knuth/Census variant gate (NOT SQL `"T520"`) | unit (literature reference vector) | `go test -run TestSoundexCode/Tymczak -race ./...` | ❌ W0 | ⬜ pending |
| 07-01-02 | 01 | 1 | PHON-01 | — | `SoundexCode("Ashcraft") = SoundexCode("Ashcroft") = "A261"` — H/W-handling gate | unit | `go test -run 'TestSoundexCode/(Ashcraft\|Ashcroft)' -race ./...` | ❌ W0 | ⬜ pending |
| 07-01-03 | 01 | 1 | PHON-01 | — | `SoundexScore` returns binary 0.0/1.0; `("", "")` → 1.0 per algorithm-correctness-standards | unit + property | `go test -run 'TestSoundexScore\|TestProp_Soundex' -race ./...` | ❌ W0 | ⬜ pending |
| 07-01-04 | 01 | 1 | PHON-01 | — | Soundex 5-invariant property tests (Identity / Symmetric / Range / NoNaN / NoInf); CodeCharset invariant (`^[A-Z][0-9]{3}$`) | property (`testing/quick`) | `go test -run TestProp_Soundex -race ./...` | ❌ W0 | ⬜ pending |
| 07-01-05 | 01 | 1 | PHON-01 | — | Soundex fuzz no-panic on ASCII + non-ASCII regimes per CONTEXT.md §5 | fuzz | `go test -fuzz=FuzzSoundex -fuzztime=60s ./...` | ❌ W0 | ⬜ pending |
| 07-01-06 | 01 | 1 | PHON-01 | — | Soundex bench `< 500 ns, 0 allocations` per performance-standards | bench | `go test -run none -bench BenchmarkSoundex -benchmem ./...` | ❌ W0 | ⬜ pending |
| 07-01-07 | 01 | 1 | PHON-01 | — | Soundex BDD scenarios (≥ 4) cover happy path + Tymczak + Ashcraft + empty input | bdd | `make test-bdd` | ❌ W0 | ⬜ pending |
| 07-01-08 | 01 | 1 | PHON-01..04 | — | `scripts/gen-phonetic-cross-validation.py` dual-pins `JELLYFISH_VERSION = "1.2.1"` AND `METAPHONE_VERSION = "0.6"`; refuses run on mismatch; writes `vectors.json._metadata.{jellyfish_version, metaphone_version, regenerated_at}` | meta + manual | `make regen-phonetic-cross-validation` then `git status` clean | ❌ W0 | ⬜ pending |
| 07-01-09 | 01 | 1 | PHON-01..04 | — | `testdata/cross-validation/phonetic/vectors.json` committed with Soundex section (15 entries per CONTEXT.md §1) | meta | `test -f testdata/cross-validation/phonetic/vectors.json` | ❌ W0 | ⬜ pending |
| 07-01-10 | 01 | 1 | PHON-01..04 | — | `testdata/golden/phonetic-codes.json` cross-platform byte-stable code golden file created with Soundex section | golden | `go test -run TestPhoneticCodesGolden/Soundex -race ./...` | ❌ W0 | ⬜ pending |
| 07-01-11 | 01 | 1 | PHON-01 | — | `dispatch_soundex.go` wires `AlgoSoundex`; `permittedMongeElkanInner` adds `AlgoSoundex: true` (14→15 permitted); `monge_elkan_test.go` panic fixture `rejected` slice shrinks (9→8) | unit | `go test -run TestMongeElkan_PanicsOnNonPermittedInner -race ./...` | ❌ W0 | ⬜ pending |
| 07-01-12 | 01 | 1 | PHON-01 | — | `TestMongeElkanScore_BinaryInner_Soundex` asserts ME-over-Soundex binary composition (full + partial + empty matches) | unit | `go test -run TestMongeElkanScore_BinaryInner_Soundex -race ./...` | ❌ W0 | ⬜ pending |
| 07-01-13 | 01 | 1 | PHON-01 | — | `Makefile` target `regen-phonetic-cross-validation` + `docs/cross-validation.md` "Phonetic cross-validation" section added | meta | `grep -q regen-phonetic-cross-validation Makefile && grep -q 'Phonetic cross-validation' docs/cross-validation.md` | ❌ W0 | ⬜ pending |
| 07-01-14 | 01 | 1 | PHON-01 | — | `llms.txt` + `llms-full.txt` synced in-plan with `SoundexCode`, `SoundexScore` entries | meta | `go test -run TestLLMsTxt -race ./...` | ❌ W0 | ⬜ pending |
| 07-02-01 | 02 | 2 | PHON-02 | — | `DoubleMetaphoneKeys("Schmidt") = ("XMT", "SMT")` Germanic-branch gate | unit (literature reference vector) | `go test -run TestDoubleMetaphoneKeys/Schmidt -race ./...` | ❌ W0 | ⬜ pending |
| 07-02-02 | 02 | 2 | PHON-02 | — | `DoubleMetaphoneScore("Schmidt", "Smith") = 1.0` Germanic XMT cross-match | unit | `go test -run TestDoubleMetaphoneScore/SchmidtSmith -race ./...` | ❌ W0 | ⬜ pending |
| 07-02-03 | 02 | 2 | PHON-02 | — | `DoubleMetaphoneKeys("Catherine") = DoubleMetaphoneKeys("Katherine") = ("K0RN", "KTRN")` Greek-branch gate | unit | `go test -run 'TestDoubleMetaphoneKeys/(Catherine\|Katherine)' -race ./...` | ❌ W0 | ⬜ pending |
| 07-02-04 | 02 | 2 | PHON-02 | — | `DoubleMetaphoneKeys("Pacheco")` contains `"PXK"` Romance/Spanish-branch gate | unit | `go test -run TestDoubleMetaphoneKeys/Pacheco -race ./...` | ❌ W0 | ⬜ pending |
| 07-02-05 | 02 | 2 | PHON-02 | — | DM language-branch mandatory checklist per CONTEXT.md §3: ≥ 1 Slavic vector AND ≥ 1 Chinese-origin vector | unit | `go test -run 'TestDoubleMetaphoneKeys/(Wojcik\|Cheung)' -race ./...` | ❌ W0 | ⬜ pending |
| 07-02-06 | 02 | 2 | PHON-02 | — | DM 5-invariant property tests | property | `go test -run TestProp_DoubleMetaphone -race ./...` | ❌ W0 | ⬜ pending |
| 07-02-07 | 02 | 2 | PHON-02 | — | DM fuzz no-panic + key charset (`^[A-Z0]{0,4}$`) | fuzz | `go test -fuzz=FuzzDoubleMetaphone -fuzztime=60s ./...` | ❌ W0 | ⬜ pending |
| 07-02-08 | 02 | 2 | PHON-02 | — | DM cross-platform byte-stable keys via `phonetic-codes.json` | golden | `go test -run TestPhoneticCodesGolden/DoubleMetaphone -race ./...` | ❌ W0 | ⬜ pending |
| 07-02-09 | 02 | 2 | PHON-02 | — | DM cross-validation against `oubiwann/metaphone==0.6` (40-entry × 5-branch corpus); loader asserts `count(branch) >= 7` for Germanic/Slavic/Romance/Greek, ≥ 4 for Chinese-origin | cross-validation | `go test -run TestPhonetic_CrossValidation/DoubleMetaphone -race ./...` | ❌ W0 | ⬜ pending |
| 07-02-10 | 02 | 2 | PHON-02 | — | DM BDD scenarios (≥ 6 — minimum 1 per language branch + Schmidt-Smith XMT cross-match) | bdd | `make test-bdd` | ❌ W0 | ⬜ pending |
| 07-02-11 | 02 | 2 | PHON-02 | — | DM bench `< 2 µs, ≤ 2 allocations` per performance-standards | bench | `go test -run none -bench BenchmarkDoubleMetaphone -benchmem ./...` | ❌ W0 | ⬜ pending |
| 07-02-12 | 02 | 2 | PHON-02 | — | `dispatch_double_metaphone.go` wires `AlgoDoubleMetaphone`; `permittedMongeElkanInner` adds entry (15→16); panic fixture shrinks (8→7); `TestMongeElkanScore_BinaryInner_DoubleMetaphone` added | unit | `go test -run 'TestMongeElkan_PanicsOnNonPermittedInner\|TestMongeElkanScore_BinaryInner_DoubleMetaphone' -race ./...` | ❌ W0 | ⬜ pending |
| 07-02-13 | 02 | 2 | PHON-02 | — | Source-Origin Statement + rule-table provenance line (canonical SWI-Prolog `double_metaphone.c` URL) + negative-attribution line listing 5 MIT-Go ports per CONTEXT.md §3 | manual + meta | grep `double_metaphone.go` header for all three line types | ❌ W0 | ⬜ pending |
| 07-02-14 | 02 | 2 | PHON-02 | — | `algorithm-licensing-reviewer` sign-off recorded in PR description; `--depth=deep` code review completed on plan 07-02 PR per CONTEXT.md §3 | manual | PR description includes `algorithm-licensing-reviewer sign-off: <sha>` + `--depth=deep code review: <reviewer>` lines | ❌ W0 | ⬜ pending |
| 07-02-15 | 02 | 2 | PHON-02 | — | `llms.txt` + `llms-full.txt` synced in-plan with `DoubleMetaphoneKeys`, `DoubleMetaphoneScore` entries | meta | `go test -run TestLLMsTxt -race ./...` | ❌ W0 | ⬜ pending |
| 07-03-01 | 03 | 2 | PHON-03 | — | `NYSIISCode("Brown") = NYSIISCode("Browne") = "BRAN"` canonical Knuth/Taft-1970 pair | unit (literature reference vector) | `go test -run 'TestNYSIISCode/(Brown\|Browne)' -race ./...` | ❌ W0 | ⬜ pending |
| 07-03-02 | 03 | 2 | PHON-03 | — | `NYSIISCode("Robert") = "RABAD"` per Knuth TAOCP Vol. 3 §6.4 | unit | `go test -run TestNYSIISCode/Robert -race ./...` | ❌ W0 | ⬜ pending |
| 07-03-03 | 03 | 2 | PHON-03 | — | NYSIIS 6-char truncation gate: `len(NYSIISCode("Catherine")) == 6` (NOT modified-NYSIIS `len 7`) | unit + property | `go test -run 'TestNYSIISCode_Truncation\|PropNYSIIS_CodeLength' -race ./...` | ❌ W0 | ⬜ pending |
| 07-03-04 | 03 | 2 | PHON-03 | — | NYSIIS 5-invariant property tests + CodeCharset (`^[A-Z]{0,6}$`) | property | `go test -run TestProp_NYSIIS -race ./...` | ❌ W0 | ⬜ pending |
| 07-03-05 | 03 | 2 | PHON-03 | — | NYSIIS fuzz no-panic on ASCII + non-ASCII regimes | fuzz | `go test -fuzz=FuzzNYSIIS -fuzztime=60s ./...` | ❌ W0 | ⬜ pending |
| 07-03-06 | 03 | 2 | PHON-03 | — | NYSIIS cross-validation against jellyfish with `variant_divergence: true` for all >6-char jellyfish outputs (~40-60% of corpus per RESEARCH.md key finding 3); loader asserts against TRUNCATED (Taft-1970) value | cross-validation | `go test -run TestPhonetic_CrossValidation/NYSIIS -race ./...` | ❌ W0 | ⬜ pending |
| 07-03-07 | 03 | 2 | PHON-03 | — | NYSIIS cross-platform byte-stable code via `phonetic-codes.json` | golden | `go test -run TestPhoneticCodesGolden/NYSIIS -race ./...` | ❌ W0 | ⬜ pending |
| 07-03-08 | 03 | 2 | PHON-03 | — | NYSIIS BDD scenarios (≥ 4) cover Brown/Browne canonical pair + truncation gate + empty input | bdd | `make test-bdd` | ❌ W0 | ⬜ pending |
| 07-03-09 | 03 | 2 | PHON-03 | — | NYSIIS bench `< 500 ns, 0 allocations` | bench | `go test -run none -bench BenchmarkNYSIIS -benchmem ./...` | ❌ W0 | ⬜ pending |
| 07-03-10 | 03 | 2 | PHON-03 | — | `dispatch_nysiis.go` wires `AlgoNYSIIS`; `permittedMongeElkanInner` adds entry (16→17); panic fixture shrinks (7→6); `TestMongeElkanScore_BinaryInner_NYSIIS` added | unit | `go test -run 'TestMongeElkan_PanicsOnNonPermittedInner\|TestMongeElkanScore_BinaryInner_NYSIIS' -race ./...` | ❌ W0 | ⬜ pending |
| 07-03-11 | 03 | 2 | PHON-03 | — | Source-Origin Statement: Taft 1970 (algorithmic origin) + Knuth TAOCP Vol. 3 §6.4 (canonical algorithm description) per CONTEXT.md §2 | manual + meta | grep `nysiis.go` header for both citations + the "Taft 1970 not available" note | ❌ W0 | ⬜ pending |
| 07-03-12 | 03 | 2 | PHON-03 | — | `llms.txt` + `llms-full.txt` synced in-plan with `NYSIISCode`, `NYSIISScore` entries | meta | `go test -run TestLLMsTxt -race ./...` | ❌ W0 | ⬜ pending |
| 07-04-01 | 04 | 2 | PHON-04 | — | `MRACode("Byrne") = "BYRN"` literature reference vector | unit | `go test -run TestMRACode/Byrne -race ./...` | ❌ W0 | ⬜ pending |
| 07-04-02 | 04 | 2 | PHON-04 | — | `MRACode` first-3-last-3 truncation gate on long input (≥ 7 chars) per NBS-943 | unit | `go test -run TestMRACode_Truncation -race ./...` | ❌ W0 | ⬜ pending |
| 07-04-03 | 04 | 2 | PHON-04 | — | `MRACompare("Smith", "Smyth")` → `(true, planner-derived int)` threshold-pass; `MRACompare("Smith", "ZachariahMontgomery")` → `(false, 0)` length-diff>=3 auto-mismatch gate per docs/requirements.md §7.4.4 line 696 | unit | `go test -run 'TestMRACompare/(SmithSmyth\|LengthDiff)' -race ./...` | ❌ W0 | ⬜ pending |
| 07-04-04 | 04 | 2 | PHON-04 | — | `MRAScore(a, b) == 1.0 iff MRACompare(a, b).matched` strict consistency property test | unit + property | `go test -run 'TestMRAScore\|TestProp_MRAScore_Consistency' -race ./...` | ❌ W0 | ⬜ pending |
| 07-04-05 | 04 | 2 | PHON-04 | — | MRA 5-invariant property tests + `PropMRA_ThresholdMonotonic` (threshold(sumLen+1) ≤ threshold(sumLen) for sumLen ∈ [0,20]) | property | `go test -run TestProp_MRA -race ./...` | ❌ W0 | ⬜ pending |
| 07-04-06 | 04 | 2 | PHON-04 | — | MRA fuzz no-panic + code charset (`^[A-Z]{0,6}$`) | fuzz | `go test -fuzz=FuzzMRA -fuzztime=60s ./...` | ❌ W0 | ⬜ pending |
| 07-04-07 | 04 | 2 | PHON-04 | — | MRA cross-validation against jellyfish (20 entries per CONTEXT.md §1 including length-diff>3 mismatch gate + threshold-edge pairs at threshold, threshold-1) | cross-validation | `go test -run TestPhonetic_CrossValidation/MRA -race ./...` | ❌ W0 | ⬜ pending |
| 07-04-08 | 04 | 2 | PHON-04 | — | MRA cross-platform byte-stable code via `phonetic-codes.json` | golden | `go test -run TestPhoneticCodesGolden/MRA -race ./...` | ❌ W0 | ⬜ pending |
| 07-04-09 | 04 | 2 | PHON-04 | — | MRA BDD scenarios (≥ 4) cover happy path + length-diff>=3 mismatch + threshold-edge + empty input | bdd | `make test-bdd` | ❌ W0 | ⬜ pending |
| 07-04-10 | 04 | 2 | PHON-04 | — | MRA bench `< 500 ns, 0 allocations` for `MRACode`; ≤ 2 allocations for `MRACompare` | bench | `go test -run none -bench BenchmarkMRA -benchmem ./...` | ❌ W0 | ⬜ pending |
| 07-04-11 | 04 | 2 | PHON-04 | — | `var mraThresholdTable` declared at package level per determinism-standards (NO `init()`); explicit comment documents `sum > 12` clamp per NBS-943 (RESEARCH.md Pitfall 7.C) | manual + unit | grep `mra.go` for `var mraThresholdTable` and clamp comment; `go test -run TestMRAThresholdTable_Clamp -race ./...` | ❌ W0 | ⬜ pending |
| 07-04-12 | 04 | 2 | PHON-04 | — | `dispatch_mra.go` wires `AlgoMRA`; `permittedMongeElkanInner` adds final entry (17→18); panic fixture shrinks (6→5); `TestMongeElkanScore_BinaryInner_MRA` added | unit | `go test -run 'TestMongeElkan_PanicsOnNonPermittedInner\|TestMongeElkanScore_BinaryInner_MRA' -race ./...` | ❌ W0 | ⬜ pending |
| 07-04-13 | 04 | 2 | PHON-04 | — | `llms.txt` + `llms-full.txt` synced in-plan with `MRACode`, `MRACompare`, `MRAScore` entries (3 new exports — note `MRACompare` has `(bool, int)` shape per CONTEXT.md §6) | meta | `go test -run TestLLMsTxt -race ./...` | ❌ W0 | ⬜ pending |
| 07-05-01 | 05 | 3 | PHON-01..04 | — | `testdata/golden/algorithms.json` extended with 4 phonetic algorithm entries; `_staging/{soundex,double_metaphone,nysiis,mra}.json` merged | golden | `make verify-determinism` | ❌ W0 | ⬜ pending |
| 07-05-02 | 05 | 3 | PHON-01..04 | — | `bench.txt` regenerated full-replace including 4 new phonetic benchmarks; benchstat regression detection green vs Phase 6 baseline on carry-forward benchmarks | bench | `make bench` then commit `bench.txt` | ❌ W0 | ⬜ pending |
| 07-05-03 | 05 | 3 | PHON-01..04 | — | `examples/identifier-similarity/main.go` extended from 19 → 23 columns (Soundex, DblMetaph, NYSIIS, MRA); golden stdout fixture regenerated | example + meta | `go test ./examples/identifier-similarity/...` | ❌ W0 | ⬜ pending |
| 07-05-04 | 05 | 3 | PHON-01..04 | — | New `examples/phonetic-keys/main.go` + `main_test.go` demonstrating `SoundexCode` / `DoubleMetaphoneKeys` / `NYSIISCode` / `MRACode` / `MRACompare` standalone | example + meta | `go test ./examples/phonetic-keys/...` | ❌ W0 | ⬜ pending |
| 07-05-05 | 05 | 3 | PHON-01..04 | — | `tests/bdd/features/monge_elkan_phonetic_inner.feature` covers ME-over-{Soundex,DM,NYSIIS,MRA} binary-inner composition | bdd | `make test-bdd` | ❌ W0 | ⬜ pending |
| 07-05-06 | 05 | 3 | PHON-01..04 | — | `llms.txt` + `llms-full.txt` `ai_friendly_test.go` walks `go/ast` and confirms every Phase 7 exported symbol has an entry | meta | `go test -run TestLLMsTxt -race ./...` | ❌ W0 | ⬜ pending |

*Status: ⬜ pending · ✅ green · ❌ red · ⚠️ flaky*

> **Threat refs.** Phase 7 phonetic algorithms have no OWASP/STRIDE-style security threats — they are pure-function O(n) algorithms on bounded inputs (typical name length < 50 chars, worst-case ≪ 1 µs per call). No DoS-vector godoc block needed (CONTEXT.md §6-prior); no `Pathological_*` bench fixtures. The phase's discipline risk is licensing (Double Metaphone fresh-transcription) not security.

---

## Wave 0 Requirements

> "Wave 0" in fuzzymatch = files that must exist before per-task tests can run. Phase 7's cross-validation script + corpus + golden code file are the foundation artifacts; plan 07-01 delivers all three.

- [ ] `soundex.go` + `soundex_test.go` + `soundex_bench_test.go` + `soundex_fuzz_test.go` + `dispatch_soundex.go` — covers PHON-01 (Wave 1, plan 07-01)
- [ ] `double_metaphone.go` + companion files + `dispatch_double_metaphone.go` — covers PHON-02 (Wave 2, plan 07-02)
- [ ] `nysiis.go` + companion files + `dispatch_nysiis.go` — covers PHON-03 (Wave 2, plan 07-03)
- [ ] `mra.go` + companion files + `dispatch_mra.go` — covers PHON-04 (Wave 2, plan 07-04)
- [ ] `scripts/gen-phonetic-cross-validation.py` (dev-only; dual-pins `jellyfish==1.2.1` + `Metaphone==0.6` per RESEARCH.md key finding 1) — plan 07-01
- [ ] `testdata/cross-validation/phonetic/vectors.json` — committed corpus, Soundex section in plan 07-01; DM/NYSIIS/MRA sections appended by plans 07-02..07-04
- [ ] `phonetic_cross_validation_test.go` — single loader with 4 `t.Run` sub-tests; created by plan 07-01 with Soundex sub-test; plans 07-02..07-04 append theirs
- [ ] `testdata/golden/phonetic-codes.json` — cross-platform byte-stable code determinism gate (NEW file per CONTEXT.md §7) — plan 07-01 creates with Soundex section
- [ ] `phonetic_codes_golden_test.go` — separate loader for `phonetic-codes.json` per CONTEXT.md §7 (recommended) — plan 07-01 creates
- [ ] `testdata/golden/_staging/{soundex,double_metaphone,nysiis,mra}.json` — per-plan staging files; merged into `algorithms.json` in 07-05
- [ ] `Makefile` target `regen-phonetic-cross-validation` — plan 07-01
- [ ] `docs/cross-validation.md` "Phonetic cross-validation" section — plan 07-01
- [ ] `tests/bdd/features/{soundex,double_metaphone,nysiis,mra}.feature` — one per algorithm plan
- [ ] `tests/bdd/features/monge_elkan_phonetic_inner.feature` — plan 07-04 OR 07-05 finalisation
- [ ] `tests/bdd/steps/algorithms_steps.go` — append step.Step registrations (per-plan, accumulator file)
- [ ] `example_test.go` — 9 new `ExampleXxx` runnable godoc examples appended across plans 07-01..07-04
- [ ] `props_test.go` — Five-invariant property-test blocks + 4 new `TestMongeElkanScore_BinaryInner_<Algo>` tests appended across plans 07-01..07-04
- [ ] `monge_elkan.go` `permittedMongeElkanInner` map — incremental 14→15→16→17→18 across plans 07-01..07-04 (lockstep per CONTEXT.md §4)
- [ ] `monge_elkan_test.go` panic-fixture `rejected` slice — incremental shrink 9→8→7→6→5 across plans 07-01..07-04 (lockstep with map mutation per RESEARCH.md §6)
- [ ] `bench.txt` finalisation regeneration — full-replace including 4 new phonetic benchmarks (plan 07-05)
- [ ] `testdata/golden/algorithms.json` finalisation merge — adds 4 algorithm entries from `_staging/` (plan 07-05)
- [ ] `examples/identifier-similarity/main.go` extension — 19 → 23 columns (plan 07-05)
- [ ] `examples/phonetic-keys/main.go` + `main_test.go` — new program (plan 07-05)
- [ ] `llms.txt` + `llms-full.txt` per-plan sync — every new exported symbol gets an entry IN THE SAME PLAN that adds it (per CONTEXT.md §6-prior; carries forward from Phase 5)

---

## Manual-Only Verifications

| Behavior | Requirement | Why Manual | Test Instructions |
|----------|-------------|------------|-------------------|
| `jellyfish` + `Metaphone` pip pins match script header | PHON-01..04 | `pip install` is operator-side; script asserts `jellyfish.__version__ == JELLYFISH_VERSION` (and `pip show Metaphone` parsing for the second pin since `Metaphone` doesn't expose `__version__`) and refuses to run on mismatch | Operator runs `make regen-phonetic-cross-validation`; script self-verifies; if it errors with version mismatch, operator runs `pip install jellyfish==1.2.1 Metaphone==0.6` then retries |
| `algorithm-licensing-reviewer` sign-off on plan 07-02 Double Metaphone | PHON-02 | Licensing review is a manual diff-vs-MIT-Go-ports inspection that no test can automate | Reviewer (the agent or a human) diffs `double_metaphone.go` against the 5 named MIT-Go ports per CONTEXT.md §3 negative-attribution; signs off in PR description with `algorithm-licensing-reviewer sign-off: <sha or hash>` line |
| `--depth=deep` code review on plan 07-02 Double Metaphone PR | PHON-02 | Per CONTEXT.md §3, plan 07-02 specifically triggers deep cross-file analysis (rule-table provenance, branch coverage, organisation vs MIT Go ports) | Reviewer runs deep review; records `--depth=deep code review: <reviewer-id>` in PR description |
| Cross-platform determinism (linux/amd64, linux/arm64, darwin/amd64, darwin/arm64, windows/amd64) | PHON-01..04 | CI matrix runs the diff; locally we cannot diff across 5 platforms | CI workflow `cross-platform-determinism.yml` runs `make verify-determinism` on each platform on every PR; merging blocked on any platform failure |
| `examples/identifier-similarity/` rendered output is human-readable with 23 columns | PHON-01..04 | Visual / aesthetic check on column alignment, header order, score-bucket colouring | `go run ./examples/identifier-similarity` and visually confirm 4 new columns (Soundex / DblMetaph / NYSIIS / MRA) are present, correctly labelled, and visually consistent with Phase 6 columns |
| `examples/phonetic-keys/` rendered output is educational and demonstrates all 4 code surfaces | PHON-01..04 | New educational example program; visual / aesthetic check on column layout and content selection | `go run ./examples/phonetic-keys` and visually confirm the side-by-side table demonstrates `SoundexCode`, `DoubleMetaphoneKeys`, `NYSIISCode`, `MRACode`, and ideally `MRACompare` |

---

## Validation Sign-Off

- [ ] All tasks have `<automated>` verify or Wave 0 dependencies
- [ ] Sampling continuity: no 3 consecutive tasks without automated verify
- [ ] Wave 0 covers all MISSING references
- [ ] No watch-mode flags
- [ ] Feedback latency < 30 s for per-task quick runs
- [ ] `nyquist_compliant: true` set in frontmatter once all per-task automated commands resolve to green
- [ ] All four phonetic algorithms ship with unit + property + fuzz + bench + BDD per ROADMAP.md success criterion 4
- [ ] Cross-validation corpus loaded by `phonetic_cross_validation_test.go` and asserts byte-stable code matches against committed `vectors.json` (Soundex/MRA direct; NYSIIS via `variant_divergence` Taft-truncated assertion; DM via `Metaphone==0.6` not jellyfish per RESEARCH.md key finding 1)
- [ ] `phonetic-codes.json` golden file loaded by `phonetic_codes_golden_test.go`; cross-platform diff green
- [ ] `algorithms.json` golden file extended with 4 new entries; cross-platform diff green
- [ ] `bench.txt` regenerated full-replace; benchstat detects no > 10% regression vs Phase 6 baseline on carry-forward benchmarks
- [ ] `monge_elkan.go` `permittedMongeElkanInner` map at 18 entries (13 base + AlgoRatcliffObershelp + 4 phonetic) per CONTEXT.md §4
- [ ] `llms.txt` + `llms-full.txt` `ai_friendly_test.go` walks `go/ast` and confirms every Phase 7 exported symbol has an entry
- [ ] `algorithm-licensing-reviewer` sign-off recorded in plan 07-02 PR description
- [ ] `--depth=deep` code review completed on plan 07-02 PR

**Approval:** pending (awaiting Phase 7 execution → /gsd-verify-work green)
