---
phase: 4
slug: remaining-character-gestalt
status: draft
nyquist_compliant: false
wave_0_complete: false
created: 2026-05-14
---

# Phase 4 — Validation Strategy

> Per-phase validation contract for feedback sampling during execution. Derived from RESEARCH.md §Validation Architecture.

---

## Test Infrastructure

| Property | Value |
|----------|-------|
| **Framework** | Go stdlib `testing` + `testing/quick` (root); `godog v0.15.0` + `goleak v1.3.0` + `testify v1.10.0` (`tests/bdd/`) |
| **Config file** | Root: none; BDD: `tests/bdd/go.mod` |
| **Quick run command** | `go test -run 'TestStrcmp95\|TestLCSStr\|TestRatcliffObershelp\|TestLongestCommonSubstring\|TestProp_Strcmp95\|TestProp_LCSStr\|TestProp_LongestCommonSubstring\|TestProp_RatcliffObershelp' ./...` |
| **Full suite command** | `make check && make test-bdd` |
| **Estimated runtime** | ~10s quick · ~60s full (Go test) + ~30s BDD |

---

## Sampling Rate

- **After every task commit:** Run quick command for the in-progress plan's symbol prefix (`TestStrcmp95*` / `TestLCSStr*` / `TestRatcliffObershelp*`)
- **After every plan wave:** Run `make check` (golangci-lint v2 + go vet + go test -race -shuffle=on + coverage + license + deps + tidy + security) AND `make test-bdd`
- **Before `/gsd-verify-work`:** Full suite green + `make verify-determinism` byte-identical across CI matrix + cross-validation corpus matches
- **Max feedback latency:** ~10s per task; ~90s per wave gate

---

## Per-Task Verification Map

| Task ID | Plan | Wave | Requirement | Threat Ref | Secure Behavior | Test Type | Automated Command | File Exists | Status |
|---------|------|------|-------------|------------|-----------------|-----------|-------------------|-------------|--------|
| 04-01-01 | 04-01-strcmp95 | 1 | CHAR-07 | — | N/A | unit | `go test -run TestStrcmp95_ReferenceVectors_Winkler1994 ./...` | ❌ W0 | ⬜ pending |
| 04-01-02 | 04-01-strcmp95 | 1 | CHAR-07 | — | Similar-char table is package-level `var` (no init()) | static + unit | `make lint && go test -run TestStrcmp95_TableInvariants ./...` | ❌ W0 | ⬜ pending |
| 04-01-03 | 04-01-strcmp95 | 1 | CHAR-07 | — | Strcmp95 ≥ JaroWinkler hierarchy invariant | property | `go test -run TestProp_Strcmp95Score_AtLeastJaroWinkler ./...` | ❌ W0 (props_test.go append) | ⬜ pending |
| 04-01-04 | 04-01-strcmp95 | 1 | CHAR-07 | — | Determinism across 1000 calls (PITFALL §14 closure) | property | `go test -run TestProp_Strcmp95Score_DeterministicAcrossRuns ./...` | ❌ W0 | ⬜ pending |
| 04-01-05 | 04-01-strcmp95 | 1 | CHAR-07 | — | Census Bureau strcmp95.c cross-validation (reference vectors only — NOT code) | unit | `go test -run TestStrcmp95_CensusBureau ./...` | ❌ W0 | ⬜ pending |
| 04-01-06 | 04-01-strcmp95 | 1 | CHAR-07 | T-fuzz-panic | Panic-free + score in [0,1] for arbitrary input | fuzz | `go test -fuzz=FuzzStrcmp95Score -fuzztime=60s` | ❌ W0 | ⬜ pending |
| 04-01-07 | 04-01-strcmp95 | 1 | CHAR-07 | — | ASCII Short 0-alloc (PERF-01 budget) | benchmark | `go test -bench=BenchmarkStrcmp95Score_ASCII_Short -benchmem` | ❌ W0 | ⬜ pending |
| 04-01-08 | 04-01-strcmp95 | 1 | CHAR-07 | — | BDD identity/empty/canonical/long-string-adjustment scenarios | BDD | `make test-bdd` (`tests/bdd/features/strcmp95.feature`) | ❌ W0 | ⬜ pending |
| 04-01-09 | 04-01-strcmp95 | 1 | CHAR-07 | — | Staging golden for Strcmp95 entries | golden | `go test -run TestGolden_Algorithms_Strcmp95_Staging ./...` | ❌ W0 (`_staging/strcmp95.json`) | ⬜ pending |
| 04-02-01 | 04-02-lcsstr | 2 | CHAR-09 | — | LCS length on Wagner-Fischer 1974 reference vectors | unit | `go test -run TestLCSStr_ReferenceVectors_WagnerFischer1974 ./...` | ❌ W0 | ⬜ pending |
| 04-02-02 | 04-02-lcsstr | 2 | CHAR-09 | — | Two-row DP (no O(mn) table allocation) | benchmark + code review | `go test -bench=BenchmarkLCSStrScore_ASCII_Long -benchmem -benchtime=1x` (assert ≤ 2 allocs/op heap path) | ❌ W0 | ⬜ pending |
| 04-02-03 | 04-02-lcsstr | 2 | CHAR-09 | — | LongestCommonSubstring tie-break = leftmost in `a` | property | `go test -run TestProp_LongestCommonSubstring_LeftmostTieBreak ./...` | ❌ W0 | ⬜ pending |
| 04-02-04 | 04-02-lcsstr | 2 | CHAR-09 | — | Returned substring is substring of both | property | `go test -run TestProp_LongestCommonSubstring_IsSubstringOfBoth ./...` | ❌ W0 | ⬜ pending |
| 04-02-05 | 04-02-lcsstr | 2 | CHAR-09 | — | Length-matches-score identity `2·len(lcs)/(la+lb) == LCSStrScore` | property | `go test -run TestProp_LongestCommonSubstring_LengthMatchesScore ./...` | ❌ W0 | ⬜ pending |
| 04-02-06 | 04-02-lcsstr | 2 | CHAR-09 | T-fuzz-panic | Fuzz exercises ALL 4 public functions (Phase 3 WR-02 closure) | fuzz | `go test -fuzz=FuzzLCSStrScore -fuzztime=60s` | ❌ W0 | ⬜ pending |
| 04-02-07 | 04-02-lcsstr | 2 | CHAR-09 | — | ASCII Short 0-alloc; rune-path ≤ 4 allocs | benchmark | `go test -bench=BenchmarkLCSStr -benchmem` | ❌ W0 | ⬜ pending |
| 04-02-08 | 04-02-lcsstr | 2 | CHAR-09 | — | BDD identity/empty/canonical/tie-break/Unicode scenarios | BDD | `make test-bdd` (`tests/bdd/features/lcsstr.feature`) | ❌ W0 | ⬜ pending |
| 04-02-09 | 04-02-lcsstr | 2 | CHAR-09 | — | Staging golden for LCSStr entries | golden | `go test -run TestGolden_Algorithms_LCSStr_Staging ./...` | ❌ W0 (`_staging/lcsstr.json`) | ⬜ pending |
| 04-03-01 | 04-03-ratcliff-obershelp | 3 | GESTALT-01 | — | difflib(autojunk=False).ratio() byte-for-byte on Dr. Dobb's 1988 pairs | unit | `go test -run TestRatcliffObershelp_DrDobbs1988 ./...` | ❌ W0 | ⬜ pending |
| 04-03-02 | 04-03-ratcliff-obershelp | 3 | GESTALT-01 | — | Numerical regression pin OUTSIDE the corpus (Phase 3 WR-03 closure) | unit | `go test -run TestRatcliffObershelp_PinnedDrDobbsValue ./...` | ❌ W0 | ⬜ pending |
| 04-03-03 | 04-03-ratcliff-obershelp | 3 | GESTALT-01 | — | Godoc contrasts with Phase 6 Indel ratios (PITFALL §6 closure) | docs grep | `grep -q "difflib.SequenceMatcher(autojunk=False" ratcliff_obershelp.go` | ❌ W0 | ⬜ pending |
| 04-03-04 | 04-03-ratcliff-obershelp | 3 | GESTALT-01 | — | OQ-1 resolution: RO is **asymmetric** by design (mirrors difflib); Symmetric property test DROPPED for RO; one hand-curated asymmetric pair pinned in cross_algorithm_consistency_test.go | unit | `go test -run TestCrossAlgorithm_RO_AsymmetricByDesign ./...` (added in plan 04-05) | ❌ W0 | ⬜ pending |
| 04-03-05 | 04-03-ratcliff-obershelp | 3 | GESTALT-01 | T-fuzz-panic | Fuzz exercises both Score + ScoreRunes (Phase 3 WR-02 closure) | fuzz | `go test -fuzz=FuzzRatcliffObershelpScore -fuzztime=60s` | ❌ W0 | ⬜ pending |
| 04-03-06 | 04-03-ratcliff-obershelp | 3 | GESTALT-01 | — | BDD identity/empty/canonical/200+char autojunk-sensitive scenarios | BDD | `make test-bdd` (`tests/bdd/features/ratcliff_obershelp.feature`) | ❌ W0 | ⬜ pending |
| 04-03-07 | 04-03-ratcliff-obershelp | 3 | GESTALT-01 | — | Staging golden for RatcliffObershelp entries | golden | `go test -run TestGolden_Algorithms_RatcliffObershelp_Staging ./...` | ❌ W0 (`_staging/ratcliff_obershelp.json`) | ⬜ pending |
| 04-04-01 | 04-04-ratcliff-obershelp-cross-validation | 4 | GESTALT-01 | — | Python generator script asserts `sys.version_info >= (3, 7)` (Phase 3 IN-07 closure) and `autojunk=False` | script | manual: `python3 scripts/gen-ratcliff-obershelp-cross-validation.py --dry-run` | ❌ W0 | ⬜ pending |
| 04-04-02 | 04-04-ratcliff-obershelp-cross-validation | 4 | GESTALT-01 | — | Corpus has 15–18 entries covering 4 mandatory categories (standard edge cases / Dr. Dobb's 1988 pairs / autojunk-sensitive 200+char / substring+partial+unicode) | unit | `go test -run TestRatcliffObershelp_CrossValidation_CorpusShape ./...` | ❌ W0 | ⬜ pending |
| 04-04-03 | 04-04-ratcliff-obershelp-cross-validation | 4 | GESTALT-01 | — | `\|our_score − difflib_ratio\| ≤ 1e-9` for every corpus entry | cross-validation | `go test -run TestRatcliffObershelp_CrossValidation ./...` | ❌ W0 | ⬜ pending |
| 04-04-04 | 04-04-ratcliff-obershelp-cross-validation | 4 | GESTALT-01 | — | autojunk-sensitive case (≥ 200 char) present in corpus and proves autojunk=False is correctly disabled | cross-validation | embedded subtest `autojunk_sensitive` in TestRatcliffObershelp_CrossValidation | ❌ W0 | ⬜ pending |
| 04-04-05 | 04-04-ratcliff-obershelp-cross-validation | 4 | GESTALT-01 | — | Makefile `regen-ratcliff-obershelp-cross-validation` target documented in CONTRIBUTING.md | meta-test | `go test -run TestMakefile_TargetsDocumentedInContributing ./...` | ❌ W0 | ⬜ pending |
| 04-05-01 | 04-05-finalisation | 5 | CHAR-07,CHAR-09,GESTALT-01 | — | Merge staging goldens into canonical `testdata/golden/algorithms.json` | golden | `go test -run TestGolden_Algorithms_Merge ./...` | ❌ W0 | ⬜ pending |
| 04-05-02 | 04-05-finalisation | 5 | CHAR-07 | — | Cross-algorithm consistency: Strcmp95 ≥ JaroWinkler on at least one input | unit | `go test -run TestCrossAlgorithm_Strcmp95_AtLeastJaroWinkler ./...` | ❌ W0 (append) | ⬜ pending |
| 04-05-03 | 04-05-finalisation | 5 | CHAR-09 | — | Cross-algorithm consistency: LCSStr ≥ Levenshtein on substring-containment input | unit | `go test -run TestCrossAlgorithm_LCSStr_AtLeastLevenshtein_Substring ./...` | ❌ W0 (append) | ⬜ pending |
| 04-05-04 | 04-05-finalisation | 5 | GESTALT-01 | — | Cross-algorithm consistency: RatcliffObershelp pinned vs difflib on visible-divergence pair | unit | `go test -run TestCrossAlgorithm_RO_PinnedAgainstDifflib ./...` | ❌ W0 (append) | ⬜ pending |
| 04-05-05 | 04-05-finalisation | 5 | CHAR-07,CHAR-09,GESTALT-01 | — | identifier-similarity example: 7 → 10 columns; `want` constant updated; `TestExample_ColumnWidths` green | meta-test | `(cd examples/identifier-similarity && go test ./...)` | ❌ W0 | ⬜ pending |
| 04-05-06 | 04-05-finalisation | 5 | CHAR-07,CHAR-09,GESTALT-01 | — | llms.txt + llms-full.txt list all 7 new exported symbols | meta-test | `go test -run TestLLMs ./...` | ❌ W0 | ⬜ pending |
| 04-05-07 | 04-05-finalisation | 5 | CHAR-07,CHAR-09,GESTALT-01 | — | bench.txt regenerated; benchstat baseline accepts Phase 4 rows | benchmark | `make bench` then `make bench-compare` | ❌ W0 | ⬜ pending |
| 04-05-08 | 04-05-finalisation | 5 | CHAR-07,CHAR-09,GESTALT-01 | — | Byte-stable across CI matrix (linux/amd64, linux/arm64, darwin/arm64, windows/amd64) | golden | `make verify-determinism` | ❌ W0 | ⬜ pending |
| Cross | every plan | every | every | — | Apache-2.0 license header on every new `.go` file | static | `bash scripts/verify-license-headers.sh` | ❌ W0 | ⬜ pending |
| Cross | every plan | every | every | — | Zero non-stdlib runtime deps maintained | static | `bash scripts/verify-no-runtime-deps.sh` | ✓ existing | ⬜ pending |
| Cross | every plan | every | every | — | Coverage ≥ 95% overall, ≥ 90% per file | coverage gate | `make coverage-check` | ❌ W0 | ⬜ pending |

*Status: ⬜ pending · ✅ green · ❌ red · ⚠️ flaky*

---

## Wave 0 Requirements

All Phase 4 source files are new — no Wave 0 framework or fixture gaps (Phase 1+2+3 infrastructure complete). The plans below ARE the Wave 0 work:

- [ ] `strcmp95.go` + `dispatch_strcmp95.go` + `strcmp95_test.go` + `strcmp95_bench_test.go` + `strcmp95_fuzz_test.go` — CHAR-07 (plan 04-01)
- [ ] `lcsstr.go` + `dispatch_lcsstr.go` + `lcsstr_test.go` + `lcsstr_bench_test.go` + `lcsstr_fuzz_test.go` — CHAR-09 (plan 04-02)
- [ ] `ratcliff_obershelp.go` + `dispatch_ratcliff_obershelp.go` + `ratcliff_obershelp_test.go` + `ratcliff_obershelp_bench_test.go` + `ratcliff_obershelp_fuzz_test.go` — GESTALT-01 algorithm path (plan 04-03)
- [ ] `scripts/gen-ratcliff-obershelp-cross-validation.py` + `testdata/cross-validation/ratcliff-obershelp/vectors.json` + `TestRatcliffObershelp_CrossValidation` appended to `ratcliff_obershelp_test.go` + Makefile target + CONTRIBUTING.md doc — GESTALT-01 difflib-equivalence gate (plan 04-04)
- [ ] `props_test.go` appends · `cross_algorithm_consistency_test.go` appends · `examples/identifier-similarity/main.go` extension · `bench.txt` regenerate · `llms.txt`/`llms-full.txt` extensions · `testdata/golden/algorithms.json` merge — cross-cutting (plan 04-05)
- [ ] Per-algorithm BDD feature files (`strcmp95.feature`, `lcsstr.feature`, `ratcliff_obershelp.feature`) + `tests/bdd/steps/algorithms_steps.go` extends — BDD coverage (per-algorithm plans)
- [ ] Per-algorithm staging goldens (`_staging/strcmp95.json`, `_staging/lcsstr.json`, `_staging/ratcliff_obershelp.json`) + on-disk fuzz seeds (`testdata/fuzz/Fuzz<Algo>Score/seed-001`)

---

## Manual-Only Verifications

| Behavior | Requirement | Why Manual | Test Instructions |
|----------|-------------|------------|-------------------|
| Cross-validation corpus regeneration | GESTALT-01 | Developer-discretion; Python in CI is not required for verification (the JSON is the fixture) | Run `make regen-ratcliff-obershelp-cross-validation` after Python stdlib updates; diff committed JSON; commit if intentional |
| Strcmp95 similar-character table transcription | CHAR-07 | Source is the Winkler 1994 TR-2 paper (non-CI artifact); table values reviewed against published 36-pair list | Reviewer compares `strcmp95SimilarChars` entries to Winkler 1994 TR-2 §3 pair list during PR review |

---

## Validation Sign-Off

- [ ] All tasks have `<automated>` verify or Wave 0 dependencies
- [ ] Sampling continuity: no 3 consecutive tasks without automated verify
- [ ] Wave 0 covers all MISSING references
- [ ] No watch-mode flags
- [ ] Feedback latency < 90s (full wave gate)
- [ ] `nyquist_compliant: true` set in frontmatter

**Approval:** pending
