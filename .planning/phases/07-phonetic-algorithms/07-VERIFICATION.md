---
phase: 07-phonetic-algorithms
verified: 2026-05-16T04:22:35Z
status: passed
score: 4/4 must-haves verified
overrides_applied: 0
re_verification: false
---

# Phase 7: Phonetic Algorithms Verification Report

**Phase Goal:** Ship the four phonetic algorithms — Soundex (Knuth/Census variant per Knuth TAOCP Vol. 3 §6.4), Double Metaphone (Philips 2000, with primary + alternate codes), NYSIIS (Taft 1970, 6-char truncation), MRA (NBS Tech Note 943). Phonetic algorithms have degenerate score normalisation (typically 0.0/1.0 binary or small discrete sets). Once phonetic algorithms exist, Monge-Elkan's permitted-inner-algorithm set is expanded to include them.

**Verified:** 2026-05-16T04:22:35Z
**Status:** passed
**Re-verification:** No — initial verification

---

## Goal Achievement

### Observable Truths

| # | Truth | Status | Evidence |
|---|-------|--------|----------|
| 1 | Soundex produces "Tymczak" → "T522" (Knuth/Census variant) and "Ashcraft" → "A261"; variant name and source appear in file's block comment | VERIFIED | `TestSoundexCode_TymczakVariantGate` and `TestSoundexCode_KnuthReferenceVectors` PASS. File comment cites Russell 1918/Knuth TAOCP Vol. 3 §6.4; the H/W-skip Knuth/Census rule is documented. |
| 2 | DoubleMetaphone returns primary + alternate codes; reference vectors per language-origin branch pass (Germanic: "Schmidt" → ("XMT","SMT"); "Pacheco" contains PXK; Greek: "Catherine" and "Katherine" match); cross-validation source is Philips 2000 C reference; algorithm-licensing-reviewer sign-off recorded | VERIFIED | `TestDoubleMetaphoneKeys_LanguageBranches` passes all 5 branches. SWI-Prolog URL and CalypsoSys negative-attribution present in `double_metaphone.go`. Sign-off line present in file header. |
| 3 | NYSIIS truncates at 6 characters per Taft 1970; reference vectors match; variant_divergence mechanism operational | VERIFIED | `TestNYSIISCode_TruncationGate` PASS (`NYSIISCode("Catherine") == "CATARA"` — 6 chars, NOT 7). RV-N13 "Byrd" and RV-N14 "Bond" (CR-01 regression tests) also PASS. 40% corpus variant_divergence rate confirmed. |
| 4 | MRA matches NBS Tech Note 943 reference cases; rule tables declared as package-level `var` (NOT in init()); Monge-Elkan's permitted-inner-list updated to include all four phonetic AlgoIDs; all four algorithms have full unit + property + fuzz + benchmark + BDD coverage and algorithms.json entries | VERIFIED | `var mraThresholdTable` at package level with `sum > 12 → 2` clamp comment. `TestMRAThresholdTable_Clamp` and `PropMRA_ThresholdMonotonic` PASS. `permittedMongeElkanInner` has 18 entries (4 phonetic added). `algorithms.json` has 184 entries (23 algorithms). |

**Score:** 4/4 truths verified

### Code-Review Critical Findings (all resolved before verification)

| Finding | Status | Evidence |
|---------|--------|----------|
| CR-01: NYSIIS RD/ND suffix rules left wrong trailing character | FIXED | `work[n-2] = 'D'` assignment present at nysiis.go lines 204, 209, 214, 219. RV-N13/RV-N14 regression tests PASS (commit 81632f0). |
| CR-02: DM SC stride i+=3 vs i+=2 | REJECTED (kept i+=3) | `TestDoubleMetaphoneKeys_LiteratureReferenceVectors/RV-DM10/Sczepanski` PASS. Commit bdcf662 added inline comment documenting why original stride matches oubiwann/metaphone==0.6 corpus ("Sczepanski → SKPN"). |
| CR-03: MRAScore("", s) returns 1.0 for single-char inputs | FIXED | `if lenA == 0 || lenB == 0 { return false, 0 }` at mra.go line 257. `TestMRACompare_OneEmpty` PASS with 5 sub-cases including single-char sides. |

---

### Required Artifacts

| Artifact | Expected | Status | Details |
|----------|----------|--------|---------|
| `soundex.go` | SoundexCode + SoundexScore; Knuth/Census H/W-skip | VERIFIED | Exists, 301 LOC, implements SoundexCode and SoundexScore, Tymczak→T522 gate passes |
| `dispatch_soundex.go` | Registers SoundexScore at dispatch[AlgoSoundex] | VERIFIED | `dispatch[AlgoSoundex] = SoundexScore` in package-level var init |
| `soundex_test.go` | 12+ literature reference vectors, Tymczak gate | VERIFIED | 14 reference vectors including Tymczak (T522), Ashcraft/Ashcroft (A261), Robert/Rupert (R163) |
| `soundex_bench_test.go` | Benchmarks with allocation reporting | VERIFIED | 5 benchmarks at 3 sizes; 30ns/op, 1 alloc/op |
| `soundex_fuzz_test.go` | FuzzSoundex with charset invariant | VERIFIED | ASCII + non-ASCII + invalid UTF-8 seeds; 6 invariants |
| `double_metaphone.go` | DoubleMetaphoneKeys + DoubleMetaphoneScore; 5 language branches; SWI-Prolog provenance | VERIFIED | 430+ LOC; SWI-Prolog URL present; CalypsoSys negative-attribution present; all 5 branch gates pass |
| `dispatch_double_metaphone.go` | Registers DoubleMetaphoneScore at dispatch[AlgoDoubleMetaphone] | VERIFIED | `dispatch[AlgoDoubleMetaphone] = DoubleMetaphoneScore` |
| `nysiis.go` | NYSIISCode ≤6 chars + NYSIISScore; Taft 1970 + Knuth TAOCP §6.4 citations; "Taft 1970 not available" note | VERIFIED | 180+ LOC; both citations and the "not available" note present; truncation gate Green |
| `dispatch_nysiis.go` | Registers NYSIISScore at dispatch[AlgoNYSIIS] | VERIFIED | `dispatch[AlgoNYSIIS] = NYSIISScore` |
| `mra.go` | MRACode + MRACompare(bool,int) + MRAScore; var mraThresholdTable at package level; NBS-943 URL | VERIFIED | 230 LOC; `var mraThresholdTable` at package level; NIST URL present; sum>12 clamp documented |
| `dispatch_mra.go` | Registers MRAScore at dispatch[AlgoMRA] | VERIFIED | `dispatch[AlgoMRA] = MRAScore` |
| `monge_elkan.go` | permittedMongeElkanInner has 18 entries (4 phonetic added) | VERIFIED | Lines 313-316 show AlgoSoundex, AlgoDoubleMetaphone, AlgoNYSIIS, AlgoMRA all `true` |
| `testdata/golden/algorithms.json` | 23 algorithms × 184 entries | VERIFIED | 184 `"algorithm":` entries; Soundex/DoubleMetaphone/NYSIIS/MRA all present |
| `testdata/golden/phonetic-codes.json` | String-equality golden file; all 4 phonetic sections | VERIFIED | 228 lines; all 4 algorithm sections populated |
| `scripts/gen-phonetic-cross-validation.py` | Dual-pin jellyfish==1.2.1 + Metaphone==0.6; refuses to run on mismatch | VERIFIED | `JELLYFISH_VERSION = "1.2.1"` and `METAPHONE_VERSION = "0.6"` constants present |
| `Makefile` | regen-phonetic-cross-validation target | VERIFIED | Target present and in .PHONY |
| `examples/identifier-similarity/main.go` | 23 columns (was 19) | VERIFIED | SoundexScore, DoubleMetaphoneScore, NYSIISScore, MRAScore all in algorithms slice |
| `examples/phonetic-keys/main.go` | Demonstrates MRACompare (bool, int) shape | VERIFIED | MRACompare call present; educationally demonstrates all encoded-key surfaces |
| `tests/bdd/features/soundex.feature` | BDD coverage | VERIFIED | File exists; 7 scenarios |
| `tests/bdd/features/double_metaphone.feature` | BDD coverage | VERIFIED | File exists; 9 scenarios |
| `tests/bdd/features/nysiis.feature` | BDD coverage | VERIFIED | File exists; 10 scenarios |
| `tests/bdd/features/mra.feature` | BDD coverage | VERIFIED | File exists; 9 scenarios |
| `tests/bdd/features/monge_elkan_phonetic_inner.feature` | ME-phonetic composition BDD | VERIFIED | File exists; 6 scenarios |
| `bench.txt` | 4 new phonetic benchmark families | VERIFIED | 1296 lines; BenchmarkSoundex, BenchmarkDoubleMetaphone, BenchmarkNYSIIS, BenchmarkMRA all present |
| `llms.txt` / `llms-full.txt` | All 9 Phase 7 exported symbols | VERIFIED | SoundexCode, SoundexScore, DoubleMetaphoneKeys, DoubleMetaphoneScore, NYSIISCode, NYSIISScore, MRACode, MRACompare, MRAScore all present |

---

### Key Link Verification

| From | To | Via | Status | Details |
|------|----|-----|--------|---------|
| `soundex.go` | `dispatch[AlgoSoundex]` | `dispatch_soundex.go` var init | WIRED | `dispatch[AlgoSoundex] = SoundexScore` confirmed |
| `double_metaphone.go` | `dispatch[AlgoDoubleMetaphone]` | `dispatch_double_metaphone.go` var init | WIRED | `dispatch[AlgoDoubleMetaphone] = DoubleMetaphoneScore` confirmed |
| `nysiis.go` | `dispatch[AlgoNYSIIS]` | `dispatch_nysiis.go` var init | WIRED | `dispatch[AlgoNYSIIS] = NYSIISScore` confirmed |
| `mra.go` | `dispatch[AlgoMRA]` | `dispatch_mra.go` var init | WIRED | `dispatch[AlgoMRA] = MRAScore` confirmed |
| `monge_elkan.go permittedMongeElkanInner` | 4 phonetic AlgoIDs | lockstep mutation per CONTEXT.md §4 | WIRED | All 4 AlgoIDs in map; `TestMongeElkan_PanicsOnNonPermittedInner` PASS with 5 rejected / 18 permitted |
| `phonetic_cross_validation_test.go` | `testdata/cross-validation/phonetic/vectors.json` | os.ReadFile + json.Unmarshal + per-entry t.Run | WIRED | All 4 sub-tests active; `TestPhonetic_CrossValidation` PASS |
| `phonetic_codes_golden_test.go` | `testdata/golden/phonetic-codes.json` | string equality assertion | WIRED | All 4 branches active; `TestPhoneticCodesGolden` PASS |
| `mra.go MRAScore` | `mra.go MRACompare` | `matched, _ := MRACompare(a, b)` | WIRED | `PropMRA_ScoreCompareConsistency` PASS |
| `scripts/gen-phonetic-cross-validation.py` | jellyfish==1.2.1 + Metaphone==0.6 | pin assertion at startup | WIRED | Both `JELLYFISH_VERSION` and `METAPHONE_VERSION` constants and runtime assertions present |
| `ai_friendly_test.go` | `llms.txt` + `llms-full.txt` | go/ast walk asserting every exported symbol | WIRED | `TestAIFriendly_LLMSTxtReferencesEveryExportedSymbol` PASS |

---

### Data-Flow Trace (Level 4)

Not applicable — all four phonetic algorithm files are pure-function implementations with no dynamic data source or rendering. They compute and return values from string inputs.

---

### Behavioral Spot-Checks

| Behavior | Command | Result | Status |
|----------|---------|--------|--------|
| Soundex Tymczak→T522 gate | `TestSoundexCode_TymczakVariantGate` | PASS | PASS |
| DM Schmidt→("XMT","SMT") + Smith→("SM0","XMT") XMT cross-match | `TestDoubleMetaphoneKeys_LanguageBranches/Germanic` | PASS | PASS |
| DM Catherine==Katherine→("K0RN","KTRN") | `TestDoubleMetaphoneKeys_LanguageBranches/Greek` | PASS | PASS |
| NYSIIS 6-char truncation gate | `TestNYSIISCode_TruncationGate` | `len("CATARA")==6` | PASS |
| MRA threshold sum>12 clamp | `TestMRAThresholdTable_Clamp` | `mraThreshold(13)==2` | PASS |
| MRA one-empty guard (CR-03 fix) | `TestMRACompare_OneEmpty` | 5 sub-cases all (false,0) | PASS |
| All 23 algorithms dispatch registered | `TestDispatch_UnregisteredSlotsAreNil` | PASS | PASS |
| permittedMongeElkanInner final state (18 entries, 5 rejected) | `TestMongeElkan_PanicsOnNonPermittedInner` | 5 rejected PANIC, 18 permitted no-panic | PASS |
| Full test suite | `go test -count=1 ./...` | ok (1.291s) | PASS |

---

### Probe Execution

No probes declared or applicable (pure library phase; no scripts/tests/probe-*.sh).

---

### Requirements Coverage

| Requirement | Source Plan | Description | Status | Evidence |
|-------------|------------|-------------|--------|----------|
| PHON-01 | 07-01 | Soundex (Knuth/Census variant, TAOCP Vol. 3 §6.4) | SATISFIED | SoundexCode + SoundexScore implemented; Tymczak→T522 gate passes; all reference vectors pass; dispatch wired; ME permitted |
| PHON-02 | 07-02 | Double Metaphone (Philips 2000), primary + alternate codes, 5 language branches | SATISFIED | DoubleMetaphoneKeys + DoubleMetaphoneScore implemented; all 5 branch gates pass; SWI-Prolog provenance; licensing sign-off; dispatch wired |
| PHON-03 | 07-03 | NYSIIS (Taft 1970), 6-char truncation | SATISFIED | NYSIISCode + NYSIISScore implemented; 6-char truncation gate passes; variant_divergence mechanism (40% corpus); CR-01 fixed with RV-N13/RV-N14; dispatch wired |
| PHON-04 | 07-04 | MRA (NBS Tech Note 943) | SATISFIED | MRACode + MRACompare(bool,int) + MRAScore implemented; threshold table at package var; sum>12 clamp; CR-03 one-empty guard fixed; dispatch wired; ME at FINAL 18-entry state |

---

### Anti-Patterns Found

| File | Line | Pattern | Severity | Impact |
|------|------|---------|----------|--------|
| `double_metaphone.go` | ~260 | `padded` allocated and immediately suppressed (`_ = padded`) — WR-02 from code review | Warning | Dead allocation on every call; time budget (127 ns/op) still comfortably met; reviewer marked as warning, not blocking |
| `double_metaphone.go` | ~128 | `dmSlgCheck` WITZ branch unreachable (W match fires first) — WR-01 | Warning | Logic produces correct result; WITZ names correctly flagged as SlavoGermanic; no behavioral impact; documentation concern only |
| `double_metaphone.go` | ~429 | Default C branch checks space-literal patterns after `dmPrep` normalisation — WR-03 | Warning | Dead branch (spaces never in `v`); `i++` always fires; no behavioral impact |
| `scripts/gen-phonetic-cross-validation.py` | ~252 | Stale Pacheco placeholder in Slavic block — WR-04 | Warning | Corpus deduplication handles it correctly; branch-count guard still passes (7 real Slavic entries); cosmetic/documentation issue |

No TBD, FIXME, or XXX markers found in any modified files (clean scan).

The three warnings above were all surfaced during the in-phase code review (07-REVIEW.md). WR-02/WR-03/WR-04 were not fixed (non-critical; behavioral impact is nil). They are acceptable known issues documented in the review. No blocker anti-patterns found.

---

### Human Verification Required

None — all critical truths can be verified programmatically for this pure-function algorithm library.

---

### Gaps Summary

No gaps. All four PHON requirements (PHON-01 through PHON-04) are satisfied in the codebase. All phase goal criteria are met:

1. Soundex: T522 gate and A261 H/W gate verified live.
2. Double Metaphone: All 5 mandatory language-branch vectors pass live; provenance documentation verified.
3. NYSIIS: 6-char truncation gate and variant_divergence mechanism verified live; CR-01 bug fix confirmed with regression tests.
4. MRA: Threshold table at package level; sum>12 clamp; CR-03 one-empty guard fixed and tested; all three public surfaces consistent.
5. Monge-Elkan at FINAL 18-entry permitted state with 5 rejected.
6. algorithms.json at 23 algorithms × 184 entries; phonetic-codes.json (228 lines); bench.txt (1296 lines).
7. Full test suite `go test -count=1 ./...` — ok in 1.291s. No failures.

---

_Verified: 2026-05-16T04:22:35Z_
_Verifier: Claude (gsd-verifier)_
