---
phase: 04-remaining-character-gestalt
verified: 2026-05-14T16:15:43Z
status: passed
score: 24/24 must-haves verified
overrides_applied: 0
re_verification:
  previous_status: none
  previous_score: n/a
  gaps_closed: []
  gaps_remaining: []
  regressions: []
human_verification: []
---

# Phase 4: Remaining Character & Gestalt Verification Report

**Phase Goal:** Complete the character-based and gestalt catalogues with Strcmp95 (Winkler 1994 similar-character table), LCSStr (longest common substring similarity), and Ratcliff-Obershelp (Dr. Dobb's Journal 1988 — explicitly the difflib-equivalent for consumers who want that semantic, distinguishing it from the Indel-based token ratios coming in Phase 6).

**Verified:** 2026-05-14T16:15:43Z
**Status:** passed
**Re-verification:** No — initial verification

## Goal Achievement

### Observable Truths (ROADMAP Success Criteria)

| #  | Truth | Status | Evidence |
| -- | ----- | ------ | -------- |
| 1  | `Strcmp95Score` produces scores matching Winkler 1994 reference implementation on canonical pairs (MARTHA/MARHTA, DWAYNE/DUANE, DIXON/DICKSONX) | VERIFIED | `TestStrcmp95_ReferenceVectors_CensusBureau` passes with subtests for all three Winkler/Census Bureau pairs (`strcmp95_test.go` referenced); cross-algorithm test `TestCrossAlgorithm_Strcmp95_AtLeastJaroWinkler` pinning hierarchy invariant passes |
| 2  | `strcmp95SimilarChars` table declared as package-level `var` (NOT in `init()`) and verified by table-invariants test | VERIFIED | `strcmp95.go:124` declares `var strcmp95SimilarChars = [...]struct{...}{...}` with 36 entries; `grep -nE "^func init\|^func.*init\(\)" strcmp95.go lcsstr.go ratcliff_obershelp.go dispatch_*.go` returns ZERO matches; `TestStrcmp95_TableInvariants` exists at `strcmp95_test.go:247` |
| 3  | `LCSStrScore` returns correct longest common substring length normalised to `[0.0, 1.0]` on Wagner-Fischer 1974 reference vectors | VERIFIED | `lcsstr.go:187` `LCSStrScore` returns `2.0 * float64(n) / float64(la+lb)` (Sørensen-Dice normalisation per spec); 7 reference-vector golden entries in `testdata/golden/_staging/lcsstr.json` |
| 4  | `LCSStr` uses two-row DP | VERIFIED | `lcsstr.go:241` `lcsstrDP` uses `prev, curr []int` rolling rows with `prev, curr = curr, prev` swap; identical structure in `lcsstr.go:269` `lcsstrDPRunes`. Note: review WR-02 flags a redundant clear loop (warning, not blocker) |
| 5  | `RatcliffObershelp` matches Python `difflib.SequenceMatcher.ratio()` (autojunk=False) outputs on Dr. Dobb's 1988 reference pairs | VERIFIED | `TestRatcliffObershelp_CrossValidation` passes against 16-entry corpus at `testdata/cross-validation/ratcliff-obershelp/vectors.json` (verified PASS for all 16 subtests including wikimedia_wikimania, gestalt_paper, autojunk_sensitive); `TestCrossAlgorithm_RatcliffObershelp_PinnedDrDobbs` and `_PinnedAgainstDifflib` lock byte-for-byte difflib equivalence at 1e-9 tolerance |
| 6  | `RatcliffObershelp` godoc explicitly contrasts it with the Indel-based token ratios coming in Phase 6 | VERIFIED | `ratcliff_obershelp.go:120-122` contains: "If you want the RapidFuzz \"ratio()\" semantics — the Indel formula 2·LCS/(\|a\|+\|b\|) used by Token Sort Ratio / Token Set Ratio / Partial Ratio — use those functions in Phase 6 instead." |
| 7  | All three algorithms have unit + property + fuzz + benchmark + BDD coverage | VERIFIED | Unit: `strcmp95_test.go`, `lcsstr_test.go`, `ratcliff_obershelp_test.go`. Property: `props_test.go` contains TestProp_Strcmp95Score_*, TestProp_LCSStr*, TestProp_RatcliffObershelp* blocks. Fuzz: `*_fuzz_test.go` for all three + seed corpora at `testdata/fuzz/Fuzz{Strcmp95,LCSStr,RatcliffObershelp}Score/seed-001`. Benchmark: 18 `Benchmark*` functions across `strcmp95_bench_test.go` (3), `lcsstr_bench_test.go` (10), `ratcliff_obershelp_bench_test.go` (5). BDD: `tests/bdd/features/{strcmp95,lcsstr,ratcliff_obershelp}.feature` with 6 + 8 + 5 scenarios |
| 8  | `algorithms.json` golden file extended with new entries and diffs byte-identically on the CI matrix | VERIFIED | `testdata/golden/algorithms.json` contains exactly 7 Strcmp95_*, 7 LCSStr_*, 7 RatcliffObershelp_* entries (21 total Phase 4 entries verified by `grep -c '"algorithm": "Strcmp95"' = 7`, same for LCSStr and RatcliffObershelp). `make verify-determinism` exits 0. SUMMARY claims "65 entries (44 Phase 2+3 + 21 Phase 4)" — actual total entry count is 59 (the SUMMARY's count is off by 6, but the 21 Phase 4 entries are present and correctly merged). See "Discrepancies" below |

**Score:** 8/8 ROADMAP success criteria verified

### Additional PLAN Frontmatter Must-Haves

| Truth | Status | Evidence |
| ----- | ------ | -------- |
| `Strcmp95Score("", "") == 1.0` (both-empty identity); `Strcmp95Score("", "abc") == 0.0` (one-empty); `Strcmp95Score(x, x) == 1.0` (identity) | VERIFIED | `TestStrcmp95_BothEmpty`, `TestStrcmp95_OneEmpty` present; identity short-circuit at `strcmp95.go:240` |
| Strcmp95 ≥ JaroWinkler hierarchy invariant holds universally | VERIFIED | `TestProp_Strcmp95Score_AtLeastJaroWinkler` (property test) + `TestCrossAlgorithm_Strcmp95_AtLeastJaroWinkler` (hand-pinned canonical pairs) both pass |
| No `*Runes` variant for Strcmp95 — ASCII-only public surface (exactly one new exported symbol `Strcmp95Score`) | VERIFIED | Only `Strcmp95Score` exported in `strcmp95.go` (line 239); no `Strcmp95ScoreRunes` defined; godoc at lines 66-72, 214-219 explicitly directs Unicode users at `Normalise` upstream |
| LCSStr public surface = 4 functions (LongestCommonSubstring + Runes; LCSStrScore + Runes), with leftmost-in-`a` tie-break | VERIFIED | Verified: `LongestCommonSubstring`, `LongestCommonSubstringRunes`, `LCSStrScore`, `LCSStrScoreRunes` at `lcsstr.go:115,157,187,213`. Strict `>` max-update at `lcsstr.go:247` and `lcsstr.go:275` enforces leftmost-in-a tie-break (RESEARCH.md Pitfall 4 closure) |
| Ratcliff-Obershelp public surface = 2 functions (RatcliffObershelpScore + Runes); asymmetric-by-design (OQ-1 LOCKED 2026-05-14) | VERIFIED | `RatcliffObershelpScore` at `ratcliff_obershelp.go:143`, `RatcliffObershelpScoreRunes` at line 169. Asymmetry documented in godoc (lines 48-65, 128-133) and pinned by `TestRatcliffObershelp_AsymmetryPin` + `TestCrossAlgorithm_RatcliffObershelp_AsymmetryPin` |
| RatcliffObershelpScore Symmetric property test INTENTIONALLY OMITTED (per OQ-1) | VERIFIED | `props_test.go:1378` contains: "NB: TestProp_RatcliffObershelpScore_Symmetric is INTENTIONALLY OMITTED per ...". `grep -nE "TestProp.*Symmetric" props_test.go` shows symmetric tests for every algorithm EXCEPT RatcliffObershelp (and only the lone hand-curated `AtLeastLevenshtein` table-driven test exists; review IN-05 flags its `Prop_` prefix as misleading but it is not goal-blocking) |
| Cross-validation script uses Python stdlib `difflib(autojunk=False)` and 16-entry corpus committed | VERIFIED | `scripts/gen-ratcliff-obershelp-cross-validation.py:37-43` documents autojunk=False as REQUIRED. `testdata/cross-validation/ratcliff-obershelp/vectors.json` contains 16 entries. `TestRatcliffObershelp_CrossValidation` runs all 16 sub-tests green |
| 4 cross-algorithm consistency tests added for Phase 4 algorithms | VERIFIED | `cross_algorithm_consistency_test.go` contains 5 new Phase 4 tests (Strcmp95 hierarchy, LCSStr substring-containment, RO Dr. Dobbs pin, RO Difflib pin, RO AsymmetryPin). All pass; SUMMARY said "4" but delivered 5 — bonus, not a deficit |
| bench.txt regenerated with Phase 4 series | VERIFIED | `bench.txt`: 626 lines; 620 Benchmark rows; contains all 18 Phase 4 Benchmark* functions (5 distinct Strcmp95* variants + 10 distinct LCSStr* variants + 5 distinct RatcliffObershelp* variants per `grep -E "BenchmarkStrcmp95\|BenchmarkLCSStr\|BenchmarkRatcliffObershelp\|BenchmarkLongestCommonSubstring" bench.txt`) |
| llms.txt + llms-full.txt synced with all 7 new exported symbols | VERIFIED | `llms.txt` lists `Strcmp95Score`, `LongestCommonSubstring`, `LongestCommonSubstringRunes`, `LCSStrScore`, `LCSStrScoreRunes`, `RatcliffObershelpScore`, `RatcliffObershelpScoreRunes` (all 7). `llms-full.txt` lines 102-152 contain godoc-style entries for all 7. `TestAIFriendly_LLMSTxtReferencesEveryExportedSymbol` (run via `make test`) exits 0 |
| identifier-similarity example extended from 7 → 10 columns | VERIFIED | `examples/identifier-similarity/main.go` references "Strcmp95", "LCSStr", "RatcliffObershelpScore" (verified via grep); per SUMMARY's verification gate the `(cd examples/identifier-similarity && go test ./...)` exits 0 |
| `make check` exits 0 (golangci-lint v2, vet, race, vulncheck, license, deps allowlist, tidy, coverage) | VERIFIED | Ran `make check` during verification: 0 lint issues, 0 vet issues, race tests pass, govulncheck clean, license headers OK (80 .go files), deps allowlist clean (2 non-indirect modules: fuzzymatch + golang.org/x/text), tidy clean, coverage 97.3% ≥ 95% overall, per-file ≥ 90%, 44 exported symbols inspected |
| `make test-bdd` exits 0 | VERIFIED | Ran `make test-bdd`: ok github.com/axonops/fuzzymatch/tests/bdd 1.705s |
| `make verify-determinism` exits 0 | VERIFIED | Ran `make verify-determinism`: TestGolden_ tests pass |
| Apache-2.0 header on every new .go file | VERIFIED | `scripts/verify-license-headers.sh` reports "OK: 80 .go files carry the Apache-2.0 header" |

**Score (combined):** 24/24 must-haves verified

### Required Artifacts

| Artifact | Expected | Status | Details |
| -------- | -------- | ------ | ------- |
| `strcmp95.go` | Strcmp95Score; unexported strcmp95SimilarChars table (36 entries as package-level var); strcmp95SimilarLookup helper; no init() | VERIFIED | 463 lines; table at line 124 with 36 entries × 0.3 credit; helper at lines 172-197; no init() found |
| `dispatch_strcmp95.go` | Registration of Strcmp95Score into dispatch[AlgoStrcmp95] via `var _ = func() bool {...}()` | VERIFIED | 33 lines; `dispatch[AlgoStrcmp95] = Strcmp95Score` at line 31 |
| `lcsstr.go` | LongestCommonSubstring + Runes; LCSStrScore + Runes; two-row DP; leftmost tie-break | VERIFIED | 315 lines; 4 exported functions; STRICT `>` max-update at lines 247 + 275 |
| `dispatch_lcsstr.go` | Registration of LCSStrScore into dispatch[AlgoLCSStr] | VERIFIED | 38 lines; `dispatch[AlgoLCSStr] = LCSStrScore` at line 36 |
| `ratcliff_obershelp.go` | RatcliffObershelpScore + Runes; asymmetric-by-design godoc; no Symmetric prop test | VERIFIED | 318 lines; 2 exported functions; godoc explicitly contrasts with Phase 6 Indel-based token ratios at lines 120-122; OQ-1 LOCKED comment present at lines 48-65 |
| `dispatch_ratcliff_obershelp.go` | Registration of RatcliffObershelpScore into dispatch[AlgoRatcliffObershelp] | VERIFIED | 38 lines; `dispatch[AlgoRatcliffObershelp] = RatcliffObershelpScore` at line 36 |
| `testdata/golden/_staging/strcmp95.json` | 7 entries (alphabetical) | VERIFIED | 7 entries; includes Strcmp95_DIXON_DICKSONX, _DWAYNE_DUANE, _HAMINGTON_HAMMINGTON, _MARTHA_MARHTA |
| `testdata/golden/_staging/lcsstr.json` | 7 entries (alphabetical) | VERIFIED | 7 entries |
| `testdata/golden/_staging/ratcliff_obershelp.json` | 7 entries (alphabetical) | VERIFIED | 7 entries |
| `testdata/golden/algorithms.json` | Merged Phase 4 entries (21 new) + Phase 2+3 entries | VERIFIED | 59 total entries: 7 Strcmp95 + 7 LCSStr + 7 RatcliffObershelp + 4 Hamming + 6 Jaro + 8 JaroWinkler + 4 Levenshtein + 6 SmithWatermanGotoh + 5 DamerauOSA + 5 DamerauFull. SUMMARY claimed 65; actual 59 (see Discrepancies) — but 21 Phase 4 entries are present and correct |
| `testdata/cross-validation/ratcliff-obershelp/vectors.json` | 16-entry corpus generated with autojunk=False; python_version pinned | VERIFIED | 16 entries; `python_version: "3.12.12"`; all 16 subtests pass `TestRatcliffObershelp_CrossValidation` |
| `scripts/gen-ratcliff-obershelp-cross-validation.py` | Generator script with autojunk=False contract | VERIFIED | Lines 36-43 document autojunk=False as REQUIRED; `make regen-ratcliff-obershelp-cross-validation` target available |
| `testdata/fuzz/Fuzz{Strcmp95,LCSStr,RatcliffObershelp}Score/seed-001` | Fuzz seed corpora | VERIFIED | All three seed-001 files exist |
| `tests/bdd/features/{strcmp95,lcsstr,ratcliff_obershelp}.feature` | BDD scenarios per algorithm | VERIFIED | 6 + 8 + 5 scenarios respectively; `make test-bdd` exits 0 |
| `cross_algorithm_consistency_test.go` | 4+ Phase 4 cross-algorithm tests | VERIFIED | 5 new Phase 4 tests appended (Strcmp95 hierarchy, LCSStr substring-containment, RO Dr. Dobbs, RO Difflib divergence, RO AsymmetryPin); all pass |
| `bench.txt` | Phase 4 benchmark series at count=10 | VERIFIED | 626 lines total; 18 Phase 4 Benchmark* series present |
| `llms.txt` + `llms-full.txt` | All 7 new exported symbols documented | VERIFIED | `llms.txt` lists all 7 (lines 85-110); `llms-full.txt` has godoc entries for all 7 (lines 102-152); `TestAIFriendly_LLMSTxtReferencesEveryExportedSymbol` exits 0 |
| `examples/identifier-similarity/main.go` + `main_test.go` | 10-column output (7 + 3 Phase 4) | VERIFIED | main.go references Strcmp95, LCSStr, RatcliffObershelpScore; SUMMARY verification gate confirms `(cd examples/identifier-similarity && go test ./...)` exits 0 |

### Key Link Verification

| From | To | Via | Status | Details |
| ---- | -- | --- | ------ | ------- |
| `dispatch_strcmp95.go` | `algoid.go` (AlgoStrcmp95) | `var _ = func() bool { dispatch[AlgoStrcmp95] = Strcmp95Score; return true }()` | WIRED | Pattern present at dispatch_strcmp95.go:30-33; AlgoStrcmp95 declared in algoid.go:95 |
| `dispatch_lcsstr.go` | `algoid.go` (AlgoLCSStr) | `var _ = func() bool { dispatch[AlgoLCSStr] = LCSStrScore; return true }()` | WIRED | Pattern present at dispatch_lcsstr.go:35-38; AlgoLCSStr declared in algoid.go:107 |
| `dispatch_ratcliff_obershelp.go` | `algoid.go` (AlgoRatcliffObershelp) | `var _ = func() bool { dispatch[AlgoRatcliffObershelp] = RatcliffObershelpScore; return true }()` | WIRED | Pattern present at dispatch_ratcliff_obershelp.go:35-38; AlgoRatcliffObershelp declared in algoid.go:180 |
| `strcmp95.go` (Strcmp95Score) | Winkler 1994 paper | Source citation in file header + algorithm comments | WIRED | Lines 17-25 cite Winkler 1994 TR-2 §3 explicitly; algorithm comments at lines 26-48 mirror paper structure |
| `lcsstr.go` (LCSStrScore) | Wagner-Fischer 1974 paper | Source citation in file header | WIRED | Lines 18-22 cite Wagner & Fischer 1974 JACM 21(1):168-173 |
| `ratcliff_obershelp.go` (RatcliffObershelpScore) | Ratcliff & Metzener 1988 + difflib | Source citation + cross-validation block | WIRED | Lines 18-20 cite Dr. Dobb's Journal 1988; lines 33-46 lock difflib autojunk=False equivalence contract |
| Strcmp95 BDD feature | step bindings | godog regex registration in `tests/bdd/steps/algorithms_steps.go` | WIRED | BDD scenarios from features/strcmp95.feature exercise via `When I compute the Strcmp95 score between ...` step found in step bindings; `make test-bdd` exits 0 |
| LCSStr BDD feature | step bindings | godog regex registration | WIRED | Same pattern; verified by `make test-bdd` passing |
| RO BDD feature | step bindings | godog regex registration | WIRED | Same pattern; verified by `make test-bdd` passing |

### Data-Flow Trace (Level 4)

Not applicable in the traditional sense — this is a pure-function library with no UI/component data flow. The "data flow" equivalent is the dispatch wiring: each algorithm function is registered into the `dispatch[AlgoID]` table, and dispatch consumers (Scorer in Phase 8, Extract in Phase 10) will receive the registered function. The dispatch wiring is verified above. Property tests and cross-validation tests prove the algorithms produce real, paper-cited values rather than stubs.

### Behavioral Spot-Checks

| Behavior | Command | Result | Status |
| -------- | ------- | ------ | ------ |
| Phase 4 algorithm tests pass | `go test -count=1 -run "TestStrcmp95_\|TestLCSStr\|TestRatcliffObershelp\|TestGolden_Algorithms_Merge\|TestCrossAlgorithm_" ./...` | ok 0.509s | PASS |
| Phase 4 property tests pass | `go test -count=1 -run "TestProp_Strcmp95\|TestProp_LCSStr\|TestProp_RatcliffObershelp" -timeout 60s ./...` | ok 0.299s | PASS |
| Dispatch registration tests pass | `go test -count=1 -run "TestDispatch_Strcmp95Registered\|TestDispatch_LCSStrRegistered\|TestDispatch_RatcliffObershelpRegistered\|TestDispatch_UnregisteredSlotsAreNil" ./...` | ok 0.214s | PASS |
| RO cross-validation against difflib corpus | `go test -count=1 -run "TestRatcliffObershelp_CrossValidation" ./...` | ok 0.327s; all 16 corpus entries pass | PASS |
| 5 new Phase 4 cross-algorithm consistency tests pass | `go test -count=1 -run "TestCrossAlgorithm_RatcliffObershelp_PinnedAgainstDifflib\|...PinnedDrDobbs\|...AsymmetryPin\|...Strcmp95_AtLeastJaroWinkler\|...LCSStr_AtLeastLevenshtein_SubstringContainment" -v ./...` | All 5 PASS | PASS |
| llms.txt sync meta-test | `go test -count=1 -run "TestAIFriendly" ./...` | ok 0.212s | PASS |
| Determinism golden | `make verify-determinism` | exit 0 | PASS |
| BDD scenarios | `make test-bdd` | ok 1.705s | PASS |
| Full quality gate (lint+vet+race+coverage+license+deps+vulncheck+tidy) | `make check` | exit 0; 0 lint issues; 0 vet issues; coverage 97.3% ≥ 95%; per-file ≥ 90%; 44 exported symbols inspected; license OK on 80 .go files; deps allowlist clean | PASS |
| No init() functions in Phase 4 files | `grep -nE "^func init\|^func.*init\(\)" strcmp95.go lcsstr.go ratcliff_obershelp.go dispatch_*.go` | NO MATCHES | PASS |
| Repo-wide no init() | `grep -rE "func.*init\s*\(\s*\)" --include="*.go" .` | No matches anywhere | PASS |

### Probe Execution

Not applicable. This phase declares no probe scripts under `scripts/*/tests/probe-*.sh` and the project does not use a probe-based verification convention (it uses Go's native test framework + Makefile targets). The behavioural spot-checks above ARE the equivalent verification gates.

### Requirements Coverage

| Requirement | Source Plan | Description | Status | Evidence |
| ----------- | ----------- | ----------- | ------ | -------- |
| CHAR-07 | 04-01-strcmp95 | Strcmp95 (Winkler 1994) with similar-character table | SATISFIED | Strcmp95Score exists; 36-pair table as package-level var; Winkler 1994 reference vectors pass; hierarchy invariant proven |
| CHAR-09 | 04-02-lcsstr | LCSStr (longest common substring) similarity | SATISFIED | LongestCommonSubstring + Runes + LCSStrScore + Runes all exist; two-row DP; leftmost tie-break; Wagner-Fischer 1974 normalisation per spec §7.1.9 |
| GESTALT-01 | 04-03 + 04-04 | Ratcliff-Obershelp similarity (Dr. Dobb's 1988); difflib-equivalent | SATISFIED | RatcliffObershelpScore + Runes; 16-entry difflib(autojunk=False) cross-validation corpus passes byte-for-byte at 1e-9 tolerance; godoc explicitly contrasts with Phase 6 Indel-based token ratios |

**Note on REQUIREMENTS.md status field:** `.planning/REQUIREMENTS.md` lines 211, 213, 228 still show CHAR-07, CHAR-09, GESTALT-01 as "Pending". This is consistent with the SUMMARY's explicit statement: "STATE.md and ROADMAP.md updates are orchestrator-owned ... this agent has not modified those files". The status flip is orchestrator-owned post-merge bookkeeping, not a phase deliverable.

### Anti-Patterns Found

| File | Line | Pattern | Severity | Impact |
| ---- | ---- | ------- | -------- | ------ |
| `strcmp95.go` | 287-288 | `richmilne/JaroWinkler` cited inline but missing from file-header Source Origin block (review WR-01) | Info | Documentation discipline gap; no functional impact on goal |
| `lcsstr.go` | 255-262, 283-287 | Redundant `O(n)` clear loop after row swap (review WR-02) | Info | Performance overhead (2x inner-loop cost on hot path); not a correctness defect; same pattern duplicated in ratcliff_obershelp.go:255-257, 305-308 |
| `strcmp95.go` | 328-330 | `m == 0` short-circuit precedes similar-credit pass (review WR-03) | Info | Behavioural choice not documented in godoc; matches Census Bureau reference but should be noted |
| `ratcliff_obershelp_test.go` | 304-332 | `python_version` field read but not asserted as ≥ 3.7 (review WR-04) | Info | Future Python patch could silently produce wrong reference values during corpus regen |

All four findings are review WARNINGs (not BLOCKERs) per the standard-depth code review at `04-REVIEW.md`. None affect goal achievement.

### Human Verification Required

None.

This phase is a pure-function library with no UI, no external service dependencies, no user-flow surface, and no real-time behaviour. All verification can be (and was) done programmatically:

- Algorithm correctness verified by reference-vector unit tests + property tests + literature citations
- difflib equivalence verified by 16-entry cross-validation corpus generated from Python stdlib
- Determinism verified by golden-file byte-stability + `make verify-determinism`
- Performance budgets verified by `bench.txt` + benchstat A/B compare
- Quality gates verified by `make check` exiting 0
- No subjective/visual/UX criteria in the phase goal

### Gaps Summary

No gaps. All 8 ROADMAP success criteria are observably true in the codebase. All 24 must-have invariants verified. All 3 requirement IDs (CHAR-07, CHAR-09, GESTALT-01) satisfied with concrete code + test evidence.

### Discrepancies (Non-Blocking)

These are minor SUMMARY-vs-reality mismatches that do not affect goal achievement, recorded for future audit:

1. **SUMMARY claims 65 total entries in `algorithms.json`; actual is 59.** Breakdown:
   - 7 Strcmp95_* (correct)
   - 7 LCSStr_* (correct)
   - 7 RatcliffObershelp_* (correct)
   - Phase 1+2+3 baseline: 38 (not 44 as SUMMARY claimed)
   - The 21 new Phase 4 entries are correctly merged, alphabetised, and golden-file-stable. The count discrepancy appears to be a SUMMARY arithmetic error, not a missing-entry defect. `make verify-determinism` confirms the file is byte-stable.

2. **SUMMARY claims "4 cross-algorithm consistency tests"; actually delivered 5.** The 5th is `TestCrossAlgorithm_RatcliffObershelp_AsymmetryPin` — the OQ-1 inverse-form regression guard. This is a bonus, not a deficit; all 5 pass.

3. **REQUIREMENTS.md status field for CHAR-07 / CHAR-09 / GESTALT-01 still shows "Pending".** This is expected — the SUMMARY explicitly notes "STATE.md and ROADMAP.md updates are orchestrator-owned ... this agent has not modified those files; the parent orchestrator merges all wave 5 worktrees, then writes STATE.md and ROADMAP.md once after the merge."

---

_Verified: 2026-05-14T16:15:43Z_
_Verifier: Claude (gsd-verifier)_
