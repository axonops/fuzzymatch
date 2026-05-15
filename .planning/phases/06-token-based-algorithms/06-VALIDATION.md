---
phase: 6
slug: token-based-algorithms
status: draft
nyquist_compliant: false
wave_0_complete: false
created: 2026-05-15
---

# Phase 6 — Validation Strategy

> Per-phase validation contract for feedback sampling during execution.
> Synthesised from `06-RESEARCH.md` §"Validation Architecture" and the project-wide standards inherited from Phases 2–5.

---

## Test Infrastructure

| Property | Value |
|----------|-------|
| **Framework** | Go stdlib `testing` + `testing/quick` (root); godog + goleak + testify (`tests/bdd/` only) |
| **Config file** | none — Go convention; `tests/bdd/go.mod` is the structural isolation boundary |
| **Quick run command** | `go test -race -shuffle=on -count=1 ./...` |
| **Full suite command** | `make check` (fmt-check + vet + lint + verify-license-headers + verify-deps-allowlist + tidy-check + security + test + coverage + coverage-check) |
| **Cross-validation command** | `go test -run TestTokenRatios_CrossValidation -race ./...` |
| **Estimated runtime** | ~25 s (quick) / ~3 min (full `make check`) on developer hardware |

---

## Sampling Rate

- **After every task commit:** Run `go test -race -shuffle=on -count=1 ./<files-touched>` then `make fmt-check && make lint`
- **After every plan wave:** Run `make test` (full root + bdd) + `go test -run TestTokenRatios_CrossValidation -race ./...`
- **Before `/gsd-verify-work`:** `make check` green AND cross-validation green AND `bench.txt` regenerated and committed
- **Max feedback latency:** ~30 s for per-task quick runs; ~3 min for full `make check`

---

## Per-Task Verification Map

> One row per public deliverable. Plan and wave columns reflect the recommended decomposition surfaced by RESEARCH.md and the planner's `Claude's Discretion` window in CONTEXT.md §6. Final wave assignment is set in PLAN.md frontmatter; this table tracks REQ→test traceability, not commit-by-commit progress.

| Task ID | Plan | Wave | Requirement | Threat Ref | Secure Behavior | Test Type | Automated Command | File Exists | Status |
|---------|------|------|-------------|------------|-----------------|-----------|-------------------|-------------|--------|
| 06-01-01 | 01 | 1 | TOKEN-02 | — | LCS DP correctness; two-row DP; stack buffer when `min(m,n) ≤ 50`; both-empty → 1.0; one-empty → 0.0 | unit + property | `go test -run TestLCSLen -race ./...` | ❌ W0 | ⬜ pending |
| 06-01-02 | 01 | 1 | TOKEN-02 | — | `indelRatio` returns `2·LCS/(|a|+|b|)` byte-identical to RapidFuzz Indel | unit + cross-validation | `go test -run TestIndelRatio -race ./...` | ❌ W0 | ⬜ pending |
| 06-01-03 | 01 | 1 | TOKEN-02 | — | `TokenSortRatioScore` matches RapidFuzz `token_sort_ratio` byte-identically on whitespace-lowercase-ASCII corpus | unit + property + fuzz + bench + BDD + cross-validation | `go test -run TestTokenSortRatio -race ./... && go test -run TestTokenRatios_CrossValidation -race ./...` | ❌ W0 | ⬜ pending |
| 06-02-01 | 02 | 2 | TOKEN-03 | DoS-T1 (asymmetric set cardinalities) | `TokenSetRatioScore` three-way max correctness; bug-for-bug empty-set → 0.0; deterministic sorted-join | unit + property + fuzz + bench + BDD + cross-validation | `go test -run TestTokenSetRatio -race ./...` | ❌ W0 | ⬜ pending |
| 06-02-02 | 02 | 2 | TOKEN-03 | DoS-T1 | Pathological asymmetric-cardinality bench fixture meets per-algorithm budget | bench | `go test -run none -bench BenchmarkTokenSetRatio_Pathological_AsymmetricSetCardinalities ./...` | ❌ W0 | ⬜ pending |
| 06-03-01 | 03 | 2 | TOKEN-04 | DoS-T2 (long–short mismatch) | `PartialRatioScore` three-region iteration matches RapidFuzz `_partial_ratio_impl`; char-set early-skip preserved | unit + property + fuzz + bench + BDD + cross-validation | `go test -run TestPartialRatio -race ./...` | ❌ W0 | ⬜ pending |
| 06-03-02 | 03 | 2 | TOKEN-04 | DoS-T2 | `PartialRatioScoreRunes` rune-aware variant produces correct scores on Unicode inputs | unit + property + fuzz + bench + BDD + cross-validation | `go test -run TestPartialRatioRunes -race ./...` | ❌ W0 | ⬜ pending |
| 06-03-03 | 03 | 2 | TOKEN-04 | DoS-T2 | `BenchmarkPartialRatio_Pathological_LongShortMismatch` meets < 10 µs per-call budget | bench | `go test -run none -bench BenchmarkPartialRatio_Pathological_LongShortMismatch ./...` | ❌ W0 | ⬜ pending |
| 06-04-01 | 04 | 2 | TOKEN-05 | — | `TokenJaccardScore` set-Jaccard over `Tokenise(s)` output; map iteration discipline (DET-03) preserved | unit + property + fuzz + bench + BDD | `go test -run TestTokenJaccard -race ./...` | ❌ W0 | ⬜ pending |
| 06-05-01 | 05 | 3 | TOKEN-01 | DoS-T3 (1000-token vs 1000-token) | `MongeElkanScore(a, b, inner, opts)` panics on non-permitted `inner` AlgoID with documented message; per-token max correctness | unit + property + fuzz + bench + BDD | `go test -run TestMongeElkan -race ./...` | ❌ W0 | ⬜ pending |
| 06-05-02 | 05 | 3 | TOKEN-01 | DoS-T3 | `MongeElkanScoreSymmetric` average-of-directions; `Symmetric` participates in `PropAlgorithmScore_Symmetric` set without exemption | property | `go test -run TestPropMongeElkanSymmetric -race ./...` | ❌ W0 | ⬜ pending |
| 06-05-03 | 05 | 3 | TOKEN-01 | DoS-T3 | `permittedMongeElkanInner` allow-list enumerated at package scope (NOT in `init()`); exhaustive panic test for every non-permitted AlgoID in dispatch table | unit | `go test -run TestMongeElkanInnerPermissions -race ./...` | ❌ W0 | ⬜ pending |
| 06-05-04 | 05 | 3 | TOKEN-01 | DoS-T3 | `BenchmarkMongeElkan_Pathological_1000Tokens` baseline recorded; meets allocation budget for 0-alloc per inner-comparison hot path | bench | `go test -run none -bench BenchmarkMongeElkan_Pathological_1000Tokens ./...` | ❌ W0 | ⬜ pending |
| 06-06-01 | 06 | 4 | TOKEN-01..05 | — | `testdata/golden/algorithms.json` extended with 5 algorithm entries; staging files merged | golden | `make verify-determinism` | ❌ W0 | ⬜ pending |
| 06-06-02 | 06 | 4 | TOKEN-01..05 | — | `bench.txt` regenerated full-replace including 3 pathological fixtures; benchstat regression detection green | bench | `make bench` then commit `bench.txt` | ❌ W0 | ⬜ pending |
| 06-06-03 | 06 | 4 | TOKEN-01..05 | — | `examples/identifier-similarity/` extended with 5 new columns | example + manual | `go run ./examples/identifier-similarity/...` | ❌ W0 | ⬜ pending |
| 06-06-04 | 06 | 4 | TOKEN-01..05 | — | `llms.txt` + `llms-full.txt` reflect every new exported symbol (per-plan sync verified) | meta | `go test -run TestLLMsTxt ./...` | ❌ W0 | ⬜ pending |

*Status: ⬜ pending · ✅ green · ❌ red · ⚠️ flaky*

> **Threat refs.** This phase has no security-relevant threats in the OWASP/STRIDE sense — token algorithms are pure-function and accept opaque strings. The `DoS-Tn` references denote algorithmic-complexity DoS classes (per `security-reviewer` agent's standing concerns and CONTEXT §5 LOCKED): T1 = TokenSet asymmetric cardinalities, T2 = PartialRatio long-short mismatch, T3 = Monge-Elkan token-count blowup. Mitigations live in the godoc `// DoS notice:` block + bench fixture + bench budget enforcement.

---

## Wave 0 Requirements

> "Wave 0" in fuzzymatch = files that must exist before per-task tests can run. Phase 6 has no shared test fixture file (each algorithm ships its own `_test.go`); the foundation kernel `token_indel.go` + `token_indel_test.go` is the only true Wave 0 artifact and is delivered by plan 06-01.

- [ ] `token_indel.go` + `token_indel_test.go` + `export_test.go` re-export — covers foundation kernel for TOKEN-02/03/04 (Wave 1, plan 06-01)
- [ ] `monge_elkan.go` + companion files (`dispatch_monge_elkan.go`, `monge_elkan_test.go`, `monge_elkan_bench_test.go`, `monge_elkan_fuzz_test.go`) — covers TOKEN-01 (Wave 3, plan 06-05)
- [ ] `token_sort_ratio.go` + companion files — covers TOKEN-02 (Wave 1, plan 06-01)
- [ ] `token_set_ratio.go` + companion files — covers TOKEN-03 (Wave 2, plan 06-02)
- [ ] `partial_ratio.go` + companion files — covers TOKEN-04 (Wave 2, plan 06-03)
- [ ] `token_jaccard.go` + companion files — covers TOKEN-05 (Wave 2, plan 06-04)
- [ ] `scripts/gen-token-ratio-cross-validation.py` (developer-only; pinned RapidFuzz version) — generator entry point (plan 06-01)
- [ ] `testdata/cross-validation/token-ratios/vectors.json` — committed cross-validation corpus (plan 06-01)
- [ ] `Makefile` target `regen-token-ratio-cross-validation` (plan 06-01)
- [ ] `tests/bdd/features/{monge_elkan,token_sort_ratio,token_set_ratio,partial_ratio,token_jaccard}.feature` — one BDD feature file per algorithm (per-plan)
- [ ] `tests/bdd/steps/algorithms_steps.go` — append step.Step registrations (per-plan)
- [ ] `props_test.go` appendage — ~30 new property tests (5 algorithms × ~6 invariants each; appended per-plan)
- [ ] `example_test.go` appendage — 7 new `ExampleXxx` runnable godoc examples (appended per-plan)
- [ ] `llms.txt` + `llms-full.txt` per-plan sync — every new exported symbol gets an entry IN THE SAME PLAN that adds it (REINFORCED FROM PHASE 5)
- [ ] `bench.txt` finalisation regeneration — includes 5 algorithm benchmarks + 3 pathological fixtures (plan 06-06)
- [ ] `testdata/golden/algorithms.json` finalisation merge — adds 5 algorithm entries from `_staging/` (plan 06-06)
- [ ] `examples/identifier-similarity/` extension — 5 new columns (plan 06-06)

---

## Manual-Only Verifications

| Behavior | Requirement | Why Manual | Test Instructions |
|----------|-------------|------------|-------------------|
| RapidFuzz version pin matches script header | TOKEN-02/03/04 | `pip install` is operator-side; script asserts `rapidfuzz.__version__ == RAPIDFUZZ_VERSION` and refuses to run on mismatch | Operator runs `make regen-token-ratio-cross-validation`; script self-verifies; if it errors with version mismatch, operator runs `pip install rapidfuzz==<pinned-version>` then retries |
| Cross-platform determinism (linux/amd64, linux/arm64, darwin/amd64, darwin/arm64, windows/amd64) | TOKEN-01..05 | CI matrix runs the diff; locally we cannot diff across 5 platforms | CI workflow `cross-platform-determinism.yml` runs `make verify-determinism` on each platform; merging blocked on failure |
| `examples/identifier-similarity/` rendered output is human-readable | TOKEN-01..05 | Visual / aesthetic check on column alignment, header order, score-bucket colouring | `go run ./examples/identifier-similarity` and visually confirm 5 new columns are present, correctly labelled, and visually consistent with Phase 5 columns |

---

## Validation Sign-Off

- [ ] All tasks have `<automated>` verify or Wave 0 dependencies
- [ ] Sampling continuity: no 3 consecutive tasks without automated verify
- [ ] Wave 0 covers all MISSING references
- [ ] No watch-mode flags
- [ ] Feedback latency < 30 s for per-task quick runs
- [ ] `nyquist_compliant: true` set in frontmatter once all per-task automated commands resolve to green
- [ ] All five algorithms ship with unit + property + fuzz + bench + BDD per spec line 14 success criterion 4
- [ ] Cross-validation corpus loaded by `token_ratio_cross_validation_test.go` and asserts byte-stable score matches against committed `vectors.json`
- [ ] `algorithms.json` golden file extended with 5 new entries; cross-platform diff green
- [ ] `bench.txt` regenerated full-replace; benchstat detects no > 10% regression vs Phase 5 baseline on the carry-forward benchmarks
- [ ] `llms.txt` + `llms-full.txt` `ai_friendly_test.go` walks `go/ast` and confirms every Phase 6 exported symbol has an entry

**Approval:** pending (awaiting Phase 6 execution → /gsd-verify-work green)
