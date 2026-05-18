# fuzzymatch — Comprehensive Code Review Findings

## Phase 8.5 Verification (2026-05-17, updated 2026-05-18)

**Phase 8.5 (Review Remediation Gate) is COMPLETE.** All 74 Critical and 195 Important findings recorded below have been **CLOSED** by the 19 plans of Phase 8.5 (08.5-01 through 08.5-17b). The 14-agent panel re-review on 2026-05-17 (Plan 08.5-18) returned **zero Critical and zero Important findings**.

**Outcome (final):** `Critical: 0, Important: 0, Improvement: 2 (R1, R2 — documented by-design deferrals)`.
**Verification report:** [`.planning/phases/08.5-review-remediation-gate/08.5-VERIFICATION.md`](.planning/phases/08.5-review-remediation-gate/08.5-VERIFICATION.md) — full agent-panel results, Q-decision traceability, R1/R2 residue with reasoned acceptance, and §5.4 R3/R4 closure narrative.
**CI matrix status:** GREEN on commit `e5ad450` across all 5 platforms (linux/amd64+arm64, darwin/amd64+arm64, windows/amd64) plus markdownlint + bench-compare; verification report addendum on `c7bb1ff` also green.
**Phase 9 status:** UNBLOCKED.

The 74 Critical findings and 195 Important findings in the tables below are **historical** as of 2026-05-17. Each maps to a closing Q-decision or scope-cluster in the verification report's §4 traceability table.

**Remaining Improvement items** (R1 + R2, by design):

- **R1 — Q7d q-gram 25% capacity-hint regression**: +15-19% time, +90-107% bytes on QGramJaccard / Dice / Tversky / Cosine Medium/Long benchmarks. Documented in Plan 16 SUMMARY and verification report §5.1. Scheduled for v1.x performance pass.
- **R2 — Gap 3 `scan.feature` + `suppression.feature`**: deferred to Phase 9 (CONTEXT.md default — feature files describe Phase 9 scan sub-package surface, which does not exist yet).
- **Q14a — cross-algorithm tokenisation cache**: deferred to v1.x (GitHub issue placeholder pending — project has no GH issues yet). Listed as roadmap item, not residue.

**Originally-listed Improvement items now CLOSED:**

- ~~R3 — coverage-floor residue (Floor 1 92.1% / Floor 2 / Floor 3)~~: **CLOSED 2026-05-18** in commit `e5ad450`. Floor 1 overall now 96.9% (≥ 95%); Floor 2 per-file all ≥ 90% (double_metaphone.go 91.4%, nysiis.go 99.1%); Floor 3 per-exported-symbol all ≥ 90% (DoubleMetaphoneKeys 87%+, NYSIISCode 99.1%, DefaultScorer + NewSWGParams refactored into testable form via unexported helpers, ScorerAlgorithm type explicitly referenced in tests).
- ~~R4 — cross-platform CI matrix~~: **CLOSED 2026-05-18** by push of commits `e75510e` (golangci-lint v2 path + `.gitattributes` + markdownlint scope), `d505ea0` (CGO_ENABLED=1 for race-detector tests), and `e5ad450` (flaky property test + Floor 1/2/3 closures). R4 surfaced 4 Important findings + 1 pre-existing flake that the 14-agent panel missed because local sign-offs ran only on darwin/arm64 — fixed and documented in verification report §5.4.

**Lesson recorded:** Future fuzzymatch phase verifications MUST run on the full cross-platform CI matrix before final approval. Local-only sign-offs are insufficient. Saved as user-memory `feedback_ci_before_verification_gate.md`.

Original Phase 8 findings retained below for historical traceability.

---

**Review trigger:** Phase 8 → Phase 9 gate
**Reviewed:** 2026-05-17
**Scope:** entire codebase (phases 1–8)
**Reviewer panel:** 16 specialist agents dispatched in parallel; **14 included** in this aggregation. Two excluded by project context:
- `commit-message-reviewer` ran and found 341 commits missing `(#issue-number)` references — irrelevant at this stage of the project because fuzzymatch is not using GitHub issues yet.
- `issue-writer` killed before completion for the same reason (no GitHub-issue templates in active use).

Each agent operated under the explicit instruction to **surface every finding, do not self-filter**. Severity tags are organisational only — not gatekeeping. The user triages.

---

## Summary

### Totals across 14 reviewers

| Severity | Count |
|---|---:|
| **Critical** | **74** |
| **Important** | **195** |
| **Improvement** | **196** |
| **Total** | **465** |

### By reviewer

| # | Reviewer | Critical | Important | Improvement | Total | Status |
|---|---|---:|---:|---:|---:|---|
| 1 | algorithm-correctness-reviewer | 1 | 12 | 22 | **35** | issues_found |
| 2 | algorithm-licensing-reviewer | 0 | 6 | 9 | **15** | issues_found |
| 3 | algorithm-performance-reviewer | 6 | 13 | 8 | **27** | NO-GO for v1.0.0 |
| 4 | api-ergonomics-reviewer | 6 | 18 | 21 | **45** | approved_with_changes |
| 5 | bdd-scenario-reviewer | 12 | 11 | 14 | **37** | NO-GO |
| 6 | code-reviewer | 3 | 18 | 24 | **45** | issues_found |
| 7 | determinism-reviewer | 0 | 5 | 12 | **17** | GO with CI prerequisite |
| 8 | devops | 0 | 9 | 17 | **26** | issues_found |
| 9 | docs-writer | 5 | 14 | 9 | **28** | issues_found |
| 10 | go-quality | 3 | 17 | 3 | **23** | issues_found |
| 11 | security-reviewer | 2 | 11 | 18 | **31** | issues_found |
| 12 | test-analyst | 11 | 18 | 13 | **42** | issues_found |
| 13 | test-writer (architecture-only) | 18 | 21 | 8 | **47** | issues_found |
| 14 | user-guide-reviewer | 7 | 22 | 18 | **47** | issues_found |
| | **TOTAL** | **74** | **195** | **196** | **465** | |

### Action category breakdown

Most findings call for one of three actions:

- **Code fix** — change a `.go`, `.md`, or config file
- **Skill clarification** — update `.claude/skills/*/SKILL.md` (the underlying standard is ambiguous, contradicted, or out of date)
- **Discuss-phase needed** — a decision is required before fixing (e.g. the Ratcliff-Obershelp asymmetry vs spec §7.5.1 contradiction, the `verify-coverage-floors.sh` Floor 3 semantics gap, the RapidFuzz "structurally transcribed" derivation question, the §14.1 allocation budget revisions)

Action category counts are not totalled here — each finding lists its own action. Most are "Code fix" (~85%); the rest split between skill clarification and discuss-phase items.

### Cross-cutting themes (findings flagged by multiple reviewers)

These themes converged independently across reviewers — they are the strongest signals in the panel:

| Theme | Reviewers | Severity |
|---|---|---|
| **Hamming public API diverges from spec** — code returns `int` with silent max(len); spec §7.1.4 requires `(int, error)` + `ErrHammingLengthMismatch` | algorithm-correctness, code-reviewer, bdd-scenario, user-guide, docs-writer (5) | Critical |
| **`WithThreshold` accepts NaN** → `Match` silently always returns false (CR-01) | api-ergonomics, security, determinism, test-analyst, code-reviewer, user-guide, docs-writer, go-quality (8) | Critical |
| **`WithTverskyAlgorithm(α=0, β=0)` accepted** → panics at first `Score` (CR-02) | api-ergonomics, security, code-reviewer, test-analyst, algorithm-correctness, user-guide, docs-writer (7) | Critical |
| **`docs/algorithms.md` is 100% TBD scaffold** — every catalogue entry says "Primary source: TBD" / "Status: planned" despite phases 2–8 having shipped 23 algorithms | algorithm-licensing, user-guide, docs-writer (3) | Critical |
| **README Quick Start is wrong & stale** — wrong tokenise output (`[xmlhttp request]` vs actual `[xml http request]`); still says "Phase 1 ships... algorithms land in Phase 2"; no `go get` instruction; "Layer 2" diagram shows non-existent `NewScorer().Score(a,b)` API | user-guide, docs-writer (2) | Critical |
| **Damerau-Levenshtein Full unconditionally heap-allocates O(m·n) DP table** — 800 MB on 10K-char input, ~20 GB on 50K-char input; every other DP algorithm uses two-row O(min(m,n)) space | security (IM-05), algorithm-performance, code-reviewer (3) | Critical (DoS) |
| **`bench.txt.new` is committed with a captured property-test failure** — stale artifact OR real bug in `TestProp_MongeElkanScore_AsymmetricWhenTokenCountAsymmetric` | code-reviewer, test-analyst (2) | Critical |
| **No Scorer benchmarks in `bench.txt`** → benchstat CI regression gate is blind for the entire Scorer layer | algorithm-performance, test-analyst (2) | Critical |
| **`WriteGoldenFile` exported in production code** but test-only — permanent v1.x compatibility liability | api-ergonomics, code-reviewer (2) | Critical |
| **`monge_elkan.go` and `llms-full.txt` name a non-existent `ErrInvalidAlgoID` sentinel** — actual sentinel is `ErrInvalidAlgorithm`; code using `errors.Is(err, ErrInvalidAlgoID)` won't compile. BDD scenarios + step infrastructure also missing | docs-writer, bdd-scenario (2) | Critical |
| **Ratcliff-Obershelp asymmetry contradicts requirements doc** — §7.5.1 says symmetric is an invariant; OQ-1 resolution (locked 2026-05-14) makes the implementation intentionally asymmetric. SKILL.md also missing the exception | docs-writer | Critical |
| **`gocritic dupBranchBody` in `double_metaphone.go:744`** — both branches of an if/else are identical `dmAdd(&p, &alt, "S", "")`. Reference implementations (Philips C++ original, Python `metaphone==0.6`) have the else branch produce `("S", "S")`. Potential silent secondary-key bug | go-quality | Critical (likely correctness defect) |
| **golangci-lint exits 1 with 27 issues across 7 linters** (gocritic 2, gocyclo 8, gofumpt 3, gosec 3, misspell 2, staticcheck 7, unparam 2) | go-quality | Critical |
| **Overall coverage 90.2% vs ≥ 95% floor; `double_metaphone.go` 83.8% vs ≥ 90% per-file floor** | go-quality, test-analyst | Critical |
| **`go.sum` has 3 stale entries** that `go mod tidy` removes — tidy-check CI gate would fail | go-quality | Critical |
| **No `FuzzScorer_*` harnesses exist** for any of Score/ScoreAll/Match | test-analyst, test-writer, security | Critical |
| **6 character-tier fuzz harnesses fuzz only the byte-path `Score`** — all `*ScoreRunes` and `Distance` functions receive zero fuzz exposure; 11 fuzz harnesses missing seed corpora | test-writer | Critical |
| **Algoid.go godoc citations contradict implementation-file citations** for 3 AlgoIDs (Damerau-Full, Strcmp95, LCSStr) | algorithm-correctness, code-reviewer (2) | Important |
| **Direct q-gram / SWG calls accept NaN/Inf params** — break the `[0.0, 1.0]` Scorer-composite guarantee | api-ergonomics, security (2) | Critical |
| **`runeAt` in soundex.go reinvents `utf8.DecodeRuneInString`** with a continuation-byte validation gap | algorithm-correctness, code-reviewer (2) | Important |
| **9 Phase 6–7 algorithms have staging golden JSON files but no `TestGolden_*_Staging` function** | test-writer | Critical |
| **3 required meta-tests don't exist**: `internal_coverage_test.go`, `readme_shop_front_test.go`, `documentation_test.go` | test-analyst, test-writer (2) | Critical |
| **Cross-validation corpora missing for entire character tier, q-gram tier, and Monge-Elkan** — only SWG (biopython), Ratcliff-Obershelp (difflib), token-ratios (RapidFuzz), phonetic (jellyfish) are externally validated | test-writer, algorithm-correctness (2) | Critical |
| **`MongeElkanScore` is asymmetric by default** and the inert `opts NormalisationOptions` parameter does nothing — recommend name swap (`MongeElkanScore` ↔ `MongeElkanScoreAsymmetric`) and remove the inert param | api-ergonomics | Important |
| **§14.1 / §14.2 allocation budgets are inconsistent with implementation reality** across DL-Full, DoubleMetaphone, Soundex/NYSIIS/MRA, q-gram tier, token-tier, Normalise, and the Scorer composite | algorithm-performance | Critical |
| **Phase 2/3 algorithm files lack the Source-Origin Statement block** that Phase 4+ files include | algorithm-licensing | Important |
| **CHANGELOG is a stub** — no entries for any of the 23 algorithms, Scorer, or cross-validation infrastructure shipped phases 2–8 | user-guide, docs-writer (2) | Important |
| **4 mandatory feature files entirely absent** — `normalisation.feature`, `determinism.feature`, `scan.feature`, `suppression.feature` per requirements §15.6 | bdd-scenario | Critical |
| **3 documented sentinel errors have no BDD scenarios AND no step infrastructure** — `ErrInvalidAlgoID`, `ErrInvalidQGramSize`, `ErrInvalidTverskyParam` | bdd-scenario | Critical |
| **Phase 2–5 feature files (14 files) have zero scenario tags** — `godog --tags=@character` matches nothing | bdd-scenario, test-analyst (2) | Important |
| **`var _ = func() bool {...}()` pattern in dispatch_*.go is package-init-time side effect** — standards skill says "no init() functions" | algorithm-correctness, determinism (2) | Important |
| **5 CI workflow actions pin to floating tags** (`@latest`, `@v0`, `@latest-stable`) — supply-chain surface during release windows | devops, security (2) | Important |
| **No `nightly.yml` workflow** despite CLAUDE.md describing one (long-form fuzz, benchstat regression, auto-PR corpus) | devops | Important |

---

## Per-Agent Findings

The 14 reviewer reports follow in the order specified by the review request. Within each, findings are ordered Critical → Important → Improvement.

Each individual report is the original reviewer's verbatim output and lives at `.planning/reviews/<agent>-FINDINGS.md`. The full text is inlined below for offline reading; the on-disk files are the source of truth for any updates.

---


## 1. algorithm-correctness-reviewer

_Source: `.planning/reviews/algorithm-correctness-FINDINGS.md`_

<details>
<summary>Click to expand full report</summary>

---
status: issues_found
agent: algorithm-correctness-reviewer
scope: entire codebase (phases 1-8)
reviewed: 2026-05-17T08:30:00Z
finding_counts:
  critical: 1
  important: 12
  improvement: 22
  total: 35
---

## Summary

The algorithm catalogue is, on the whole, in excellent shape: every algorithm cites a primary source, formulae are documented in detail in each file's block comment, constants are unexported and traceable to the paper, mathematical invariants are covered by property tests using `testing/quick`, and four cross-validation corpora (SWG, Ratcliff-Obershelp, phonetic, token-ratios) gate regression. The fresh-implementation discipline, including the explicit listing of MIT-licensed Go ports NOT consulted, is exemplary.

The single Critical finding is a public-API contract divergence on `HammingDistance`: the spec (§7.1.4) says `(int, error)` with `ErrHammingLengthMismatch`; the implementation returns `int` and silently returns `max(len(a), len(b))` on unequal-length inputs. Either the spec or the code needs to move. Until that's resolved, the v1.0 freeze on this surface is at risk.

The Important findings are mostly documentation drift between `algoid.go`'s per-AlgoID godoc citations and the implementation files' citations (multiple algorithms cite different authors in the two locations), a couple of missing property invariants the standards document calls "All algorithms" (`NeverPanics` is not exercised as a generic property test for any algorithm — fuzz tests cover the spirit, but the standards skill names this an "All algorithms" invariant), and a few spec-vs-code divergences worth resolving before v1.0.

The Improvements are stylistic and refactoring-level observations — duplicated helpers (`runeAt` in soundex.go reinvents `utf8.DecodeRuneInString`), the dispatch registration pattern that technically performs package-init work via `var _ = func()bool{}()`, missing inner-loop swap optimisations on a couple of paths, and so on.

Phase 9 (scan sub-package) can proceed in parallel with these findings — none of them block the scan layer's design. The Critical Hamming-API finding should be triaged before v1.0 ships.

---

## Critical

### [Critical] Hamming public API diverges from spec §7.1.4
- **File:** /Users/johnny/Development/fuzzymatch/hamming.go:69-90, /Users/johnny/Development/fuzzymatch/errors.go:30-32, /Users/johnny/Development/fuzzymatch/docs/requirements.md:362-364
- **Phase introduced:** Phase 2
- **Issue:** `docs/requirements.md` §7.1.4 specifies the Hamming public surface as:
  - `HammingDistance(a, b string) (int, error)` returning `ErrHammingLengthMismatch` when `len(a) != len(b)`
  - `HammingDistanceRunes(a, b string) (int, error)`
  The actual implementation in `hamming.go` declares `HammingDistance(a, b string) int` (no error return) and silently returns `max(len(a), len(b))` on unequal-length inputs. `errors.go:31` even calls out that `ErrHammingLengthMismatch` would "land alongside the features that introduce them in later phases" — i.e. it was deferred, but no follow-up has reconciled the spec. This is a public-API contract divergence; either the spec must be amended (and the `errors.go` comment removed) or the implementation must add the error return before v1.0 freezes the surface.
- **Standard:** `docs/requirements.md` §7.1.4 (authoritative); `.claude/skills/algorithm-correctness-standards/SKILL.md` §"Edge cases" — "Length mismatch for Hamming: return 0.0 from Score; return `ErrHammingLengthMismatch` from Distance".
- **Action:** Discuss-phase needed (api-ergonomics-reviewer should weigh in — the silent-zero convention has a usability argument; the explicit-error convention has a correctness argument). Whatever the resolution, both the spec and the code must say the same thing.
- **Rationale:** Once v1.0 ships, breaking this surface needs a v2.0 bump. Resolving before v1.0 costs nothing; resolving after is expensive.
- **Suggested fix:** Either (a) amend `docs/requirements.md` §7.1.4 to match the silent-zero implementation and delete the `errors.go:31` comment about `ErrHammingLengthMismatch`, or (b) re-implement `HammingDistance` and `HammingDistanceRunes` to return `(int, error)`, add `ErrHammingLengthMismatch` to `errors.go`, and update all callers (scorer, tests, BDD scenarios).

---

## Important

### [Important] algoid.go godoc citations contradict implementation-file citations
- **File:** /Users/johnny/Development/fuzzymatch/algoid.go:67-71, 90-95, 104-107
- **Phase introduced:** Phase 1
- **Issue:** Three AlgoID godoc comments cite different primary sources than the corresponding implementation file:
  - `AlgoDamerauLevenshteinFull` (algoid.go:67-71) cites "Damerau 1964 — A technique for computer detection and correction of spelling errors." `damerau_full.go:18-19` cites "Lowrance, R., Wagner, R. A. (1975). An extension of the string-to-string correction problem." The actual algorithm IS Lowrance-Wagner 1975 (the implementation file is correct); the algoid.go godoc is wrong.
  - `AlgoStrcmp95` (algoid.go:90-95) cites "Winkler & Thibaudeau 1991 — An application of the Fellegi-Sunter model of record linkage to the 1990 U.S. decennial census." `strcmp95.go:18-19` cites "Winkler, W. E. (1994). Advanced methods for record linkage." Both are valid Winkler papers but they aren't the same paper — strcmp95.go is correct (the algorithm IS in Winkler 1994 §3).
  - `AlgoLCSStr` (algoid.go:104-107) cites "Hunt & Szymanski 1977 — A fast algorithm for computing longest common subsequences (substring variant)." `lcsstr.go:18-19` cites "Wagner, R. A., & Fischer, M. J. (1974). The string-to-string correction problem." Hunt-Szymanski is the LCS-subsequence algorithm; LCSStr is the longest common SUBSTRING (different problem). The implementation file is correct.
- **Standard:** `.claude/skills/algorithm-correctness-standards/SKILL.md` §"Primary Source Citation" — "The citation matches docs/requirements.md §7 entry exactly."
- **Action:** Code fix (godoc-only — no semantics change).
- **Rationale:** Citation accuracy is the load-bearing audit trail for algorithm-licensing and correctness reviews. Drift between two locations confuses downstream consumers (godoc renders, README, llms.txt) and weakens the citation contract.
- **Suggested fix:** Update the three AlgoID godoc comments in algoid.go to match the implementation files (Lowrance-Wagner 1975, Winkler 1994, Wagner & Fischer 1974 respectively).

### [Important] No generic NeverPanics property test for any algorithm
- **File:** /Users/johnny/Development/fuzzymatch/props_test.go (entire file — pattern absence)
- **Phase introduced:** Phase 2 (pattern set by first algorithm)
- **Issue:** The standards skill `algorithm-correctness-standards/SKILL.md` §"Mathematical Invariants" lists "Never panics" as one of three "All algorithms" invariants: "the function does not panic on arbitrary inputs, including invalid UTF-8, embedded NULs, lone surrogates, and very long strings." `props_test.go` has 219 `TestProp_*` tests but NONE of them explicitly assert no-panic via `testing/quick` — the spirit is covered by the 25 fuzz tests (each starts with "Never panics (implicit — any panic propagates as a fuzz crash)") but the standards skill specifically calls out "verified by property tests using testing/quick." The fuzz coverage is genuinely good, but the standards skill names this invariant for `testing/quick` and the test file should comply or the standards skill should be amended.
- **Standard:** `.claude/skills/algorithm-correctness-standards/SKILL.md` §"All algorithms" — "Never panics: the function does not panic on arbitrary inputs."
- **Action:** Either (a) add `TestProp_<Algo>Score_NeverPanics` for each algorithm using `testing/quick` with a `defer recover()` wrapper, or (b) amend `algorithm-correctness-standards/SKILL.md` to acknowledge that "Never panics" is covered by the per-algorithm `Fuzz<Algo>Score` native-fuzz target instead of `testing/quick`.
- **Rationale:** The fuzz approach is arguably stronger (coverage-guided), but the standards skill names `testing/quick` explicitly. Either align the tests or align the skill.
- **Suggested fix:** Amend the skill — fuzz is the right coverage shape for this invariant; documentation drift is the cheaper fix.

### [Important] LCSStr inner-loop swap is missing — stack fast path can miss
- **File:** /Users/johnny/Development/fuzzymatch/lcsstr.go:127-155, 311-319
- **Phase introduced:** Phase 2
- **Issue:** `LongestCommonSubstring` deliberately does NOT swap `a` and `b` when `a` is longer, because the leftmost-in-`a` tie-break would change identity if `a` and `b` were swapped (the implementation correctly documents this at lines 134-141). However, `lcsstrLengthOnly` (called by `LCSStrScore`) also does not swap, even though `LCSStrScore` does not need the leftmost-in-`a` tie-break — only the LENGTH. The ASCII fast-path gate at line 312 fires `if lb <= maxStackInputLen`; if `lb > maxStackInputLen` but `la <= maxStackInputLen`, the stack path is skipped despite being safe. The fix is a 4-line swap in `lcsstrLengthOnly`.
- **Standard:** `.claude/skills/performance-standards/SKILL.md` — ASCII fast-path zero-alloc budget; this is a missed-optimisation against that target.
- **Action:** Code fix (performance only — semantics unchanged because length is symmetric).
- **Rationale:** Modest 0-alloc-on-short-input gain; matters for benchmarks but not correctness.
- **Suggested fix:** In `lcsstrLengthOnly`, prepend `if la < lb { a, b = b, a; la, lb = lb, la }` before the stack-gate; the length is symmetric, so the swap is value-preserving.

### [Important] Spec says HammingDistanceRunes returns (int, error); implementation returns int
- **File:** /Users/johnny/Development/fuzzymatch/hamming.go:105-125, /Users/johnny/Development/fuzzymatch/docs/requirements.md:363
- **Phase introduced:** Phase 2
- **Issue:** Same root cause as the Critical Hamming finding above. Listed separately because the rune variant has its own surface and the rune-path test (`TestHamming_DistanceRunes_UnequalLength`) explicitly pins the silent-`max` behaviour. Whatever resolution the Critical finding settles on, both byte and rune variants must align with the spec.
- **Standard:** spec §7.1.4
- **Action:** Discuss-phase (linked to Critical Hamming finding)
- **Rationale:** Public API contract consistency.

### [Important] dispatch_*.go files use init-time side effects via `var _ = func() bool {...}()`
- **File:** /Users/johnny/Development/fuzzymatch/dispatch_*.go (all 23 files)
- **Phase introduced:** Phase 2 (pattern set by first dispatch wire-up)
- **Issue:** Every dispatch registration file uses the idiom `var _ = func() bool { dispatch[AlgoX] = XScore; return true }()` and the file header comments call this "the canonical way to run package-level side effects without init()." This is a TECHNICALLY correct read of "no init() function" but the `var _ = func()...()` form executes during package initialisation (it's a package-level var initializer expression) — Go's spec treats it identically to an `init()` function for ordering and side-effect semantics. The standards skill says: "The library has no init() functions doing non-trivial work. Tables that require initialisation ... are declared via `var x = ...` literal expressions, not built in init()." A function-pointer assignment is borderline "trivial" — it's a single map write — but the pattern is the same package-initialisation surface that the standard tries to prohibit. Reasonable people can disagree; the project explicitly calls this pattern out so it is at least known and intentional.
- **Standard:** `.claude/skills/determinism-standards/SKILL.md` §13.5 (cited line 155 of the determinism skill).
- **Action:** Skill clarification (the implementation is fine; the standards skill should explicitly bless the var-`_`-func pattern OR the project should refactor to a different registration approach).
- **Rationale:** Future contributors will read the standards skill, see "no init()", and either (a) propose a refactor that breaks the pattern or (b) ask whether the existing pattern is permitted. A two-sentence clarification in the skill saves that round-trip.
- **Suggested fix:** Add to determinism-standards SKILL.md §13.5: "Package-level `var _ = func() bool { dispatch[X] = Y; return true }()` registrations are permitted for the dispatch table and equivalent function-pointer writes; the prohibition is on table BUILDS that depend on runtime computation, not on simple map writes."

### [Important] Cosine FMA-fusion risk documented but no remediation gate
- **File:** /Users/johnny/Development/fuzzymatch/cosine.go:288-297, 341-344
- **Phase introduced:** Phase 5
- **Issue:** The file documents the FMA-fusion risk on arm64 (Go issue #17895): "Go 1.26 may emit FMA on arm64 for the (x*y)+z pattern; parentheses do NOT defeat FMA fusion." The remediation pattern (explicit float64 cast: `dot = float64(float64(qa[k]) * float64(qb[k])) + dot`) is documented but NOT applied. The cross-platform CI matrix is the load-bearing detector. This is a deliberate trade-off (the integer-derived values are small enough that FMA divergence falls below the byte-diff threshold of algorithms.json), but if a future input triggers divergence, the remediation is buried in a file comment instead of being a one-liner test that fires before CI ever sees it.
- **Standard:** `.claude/skills/determinism-standards/SKILL.md` §13.3 (no FMA in algorithm hot paths).
- **Action:** Code fix (preventive) or skill clarification (acknowledge the trade-off).
- **Rationale:** A determinism regression caught in CI matrix output is harder to triage than a determinism regression caught by a unit test that explicitly exercises the FMA-prone input pattern. Either the remediation should be applied pre-emptively, or there should be a dedicated test that asserts byte-identical output between the explicit-cast form and the project's current form to ensure they stay equivalent.
- **Suggested fix:** Apply the explicit-cast pattern now; it costs nothing on amd64 and removes the FMA risk on arm64 forever.

### [Important] Double Metaphone implementation is the most complex algorithm; cross-validation corpus exists but reviewer cannot verify the rule transcription end-to-end
- **File:** /Users/johnny/Development/fuzzymatch/double_metaphone.go (entire file, 916 lines)
- **Phase introduced:** Phase 7
- **Issue:** The Double Metaphone implementation is 916 lines of position-by-position state-machine rules with look-behind / look-ahead of up to 4 positions. The file header explicitly states "no code copied" from the SWI-Prolog C reference, the oubiwann/metaphone Python BSD-3 port, or the four MIT-licensed Go ports listed. The cross-validation corpus (testdata/cross-validation/phonetic/vectors.json) is the load-bearing correctness gate. From this review's vantage, the algorithm-correctness-reviewer cannot verify the 916-line state machine against Philips 2000 paragraph-by-paragraph (the original CUJ paper is paywalled / archive-only; the file cites a SWI-Prolog mirror of the public-domain C code). Several rule branches contain hand-noted bug-comments ("Initial SCH + consonant (not W) — Germanic names like Schmidt. Primary is X (sh-sound); secondary is S (Germanic hard SCH)"). Each of these is a place where a transcription error would silently produce a different key but still pass range-bounds / identity property tests. The defence-in-depth here is the cross-validation corpus alone — and the corpus is generated from `oubiwann/metaphone==0.6` which has its own bugs (the implementation's own godoc cites "Maurice Aubrey's perl-port-derived bug fixes" as fixes the Python port may or may not have absorbed).
- **Standard:** `.claude/skills/algorithm-correctness-standards/SKILL.md` §"Approval Gate" — "literature reference vectors in unit tests."
- **Action:** Discuss-phase needed.
- **Rationale:** Not a defect; a structural concern. The current cross-validation against oubiwann/metaphone may itself be unsound if oubiwann/metaphone has bugs Philips 2000 does not. Recommend either: (a) expand cross-validation corpus to include `jellyfish` (the BSD-2-Clause Rust port used elsewhere for NYSIIS/MRA/Soundex cross-validation) for diversity, or (b) hand-derive 5-10 additional reference vectors directly from the CUJ paper's worked examples and pin them in `double_metaphone_test.go` as paper-anchored rather than tooling-anchored fixtures.
- **Suggested fix:** Add a small `TestDoubleMetaphone_PaperWorkedExamples` table populated from Philips 2000 directly (the paper publishes ~10 worked examples — they are reproducible from the SWI-Prolog mirror's reference comments).

### [Important] runeAt helper in soundex.go reinvents stdlib utf8.DecodeRuneInString
- **File:** /Users/johnny/Development/fuzzymatch/soundex.go:275-300 (and called from double_metaphone.go, nysiis.go, mra.go)
- **Phase introduced:** Phase 7
- **Issue:** `runeAt` is a hand-rolled UTF-8 decoder that the godoc itself describes as "logic is identical to utf8.DecodeRuneInString for the purpose of skip-counting." The rationale given ("avoids an extra import in a small file") is weak — the four phonetic files all import this helper from soundex.go anyway; importing `unicode/utf8` once across all four would be cleaner. The hand-rolled decoder has a subtle correctness gap: it returns `0xFFFD, 1` for continuation bytes out of context but does NOT validate the continuation bytes' high bits (an invalid UTF-8 sequence like `0xC3 0x41` is decoded as `(0xC1, 2)` instead of being rejected). `utf8.DecodeRuneInString` handles this correctly per the documented invariants.
- **Standard:** `.claude/skills/algorithm-correctness-standards/SKILL.md` §"Unicode Handling" — "never panic on invalid UTF-8" is satisfied, but byte-validation correctness matters for the fuzz-test landscape.
- **Action:** Code fix (replace `runeAt` with `utf8.DecodeRuneInString`).
- **Rationale:** Lower code surface, fewer bug-vectors, matches stdlib semantics.
- **Suggested fix:** Delete `runeAt` from soundex.go; replace callsites with `utf8.DecodeRuneInString(s[i:])`. The import is one line per file (or one shared util).

### [Important] NYSIIS silently truncates inputs over 128 ASCII letters
- **File:** /Users/johnny/Development/fuzzymatch/nysiis.go:122-126, 143-147
- **Phase introduced:** Phase 7
- **Issue:** `NYSIISCode` allocates `var nameBuf [128]byte` for the ASCII-letter scan. Inputs with more than 128 ASCII letters are silently truncated at `nLen < 128` gate (lines 143-147). For typical name inputs this is fine, but for adversarial inputs (e.g. a malformed CSV field with 500 letters) the output is the NYSIIS of the first 128 letters, indistinguishable from a "well-formed" input that happens to share that prefix. The truncation is silent: no log, no warning, no error. The MRA implementation has the same pattern at 64 letters.
- **Standard:** `.claude/skills/algorithm-correctness-standards/SKILL.md` §"Edge cases" — "very long strings" are listed as inputs the function must not panic on. Silent truncation is not panicking, but it IS producing a different output than the algorithm specifies for those inputs.
- **Action:** Code fix or doc-only fix.
- **Rationale:** Phonetic algorithms are typically applied to names (max ~50 letters realistic); 128 is generous. But the silent-truncation behaviour is surprising and undocumented.
- **Suggested fix:** Either (a) grow the buffer to `make([]byte, len(s))` when input exceeds 128 (cheap heap alloc on pathological inputs only), or (b) document the truncation explicitly in `NYSIISCode` godoc.

### [Important] Tversky direct-call panic message uses lowercase "tversky" — inconsistent with rest of file
- **File:** /Users/johnny/Development/fuzzymatch/tversky.go:242, 284
- **Phase introduced:** Phase 5
- **Issue:** The panic message is `"fuzzymatch: invalid tversky parameter"` (lowercase tversky). The error sentinel name (`ErrInvalidTverskyParam` in errors.go:88) and the message text in `errors.go:88` are also lowercase (`"fuzzymatch: invalid tversky parameter"`) — so that's consistent, but the file-level godoc and the algorithm name throughout refer to "Tversky" (capitalised, eponymous). Per Go convention error messages should be lowercase. The lowercase form is fine and matches Go style, but it's at odds with the `Tversky` casing elsewhere. Lowercase wins per Go style; flag is a callout that the convention is intentional (verified against `errors.go` per-skill).
- **Standard:** Go convention: error messages are lowercase. No standards violation.
- **Action:** No fix required (just a callout for the reviewer's record).

### [Important] Token Set Ratio's RapidFuzz #110 deviation is documented but the catalogue's both-empty-→-1.0 convention is broken silently for one algorithm
- **File:** /Users/johnny/Development/fuzzymatch/token_set_ratio.go:281-292, 87-107
- **Phase introduced:** Phase 6
- **Issue:** TokenSetRatio is the SOLE algorithm in the catalogue that returns 0.0 (not 1.0) for both-empty input. The file godoc documents this extensively as RapidFuzz issue #110 / fuzzywuzzy parity, and the function comment includes a load-bearing explanation. However, this means the "all algorithms return 1.0 for both-empty" rule from `algorithm-correctness-standards/SKILL.md` §"Edge cases" is broken for one algorithm. The standards skill should explicitly document the deviation OR the algorithm should be brought into line. The current state is "documented deviation" which is fine but the standards skill itself does not call out that exception.
- **Standard:** `.claude/skills/algorithm-correctness-standards/SKILL.md` §"Edge cases" — "Both inputs empty: return 1.0 by convention."
- **Action:** Skill clarification.
- **Rationale:** Future reviewers reading the standards skill will see "Both-empty: 1.0" and flag TokenSetRatio as non-compliant; we need a one-line note in the skill.
- **Suggested fix:** Add to algorithm-correctness-standards SKILL.md §"Edge cases": "Exception: TokenSetRatio returns 0.0 for both-empty input per RapidFuzz issue #110 / fuzzywuzzy parity. The deviation is documented in token_set_ratio.go and is the single catalogue-wide exception to the 1.0 convention."

### [Important] Spec describes a unified phonetic-algorithm score normalisation rule that the implementations diverge from
- **File:** /Users/johnny/Development/fuzzymatch/soundex.go:257-269, /Users/johnny/Development/fuzzymatch/double_metaphone.go:890-915, /Users/johnny/Development/fuzzymatch/nysiis.go:350-362, /Users/johnny/Development/fuzzymatch/mra.go:348-358
- **Phase introduced:** Phase 7
- **Issue:** `.claude/skills/algorithm-correctness-standards/SKILL.md` §"Score Normalisation" says: "Phonetic algorithms: 1.0 if encoded keys match (per algorithm's matching rule), else 0.0." This is broadly correct, but each phonetic algorithm implements a different "matching rule":
  - Soundex: exact code equality (both must be non-empty).
  - Double Metaphone: 4-way match (primary_a == primary_b OR primary_a == secondary_b OR secondary_a == primary_b OR secondary_a == secondary_b). Each match must be non-empty.
  - NYSIIS: exact code equality (both must be non-empty).
  - MRA: the binary form of MRACompare — which uses the NBS 6-counter similarity threshold; codes don't need to match exactly.
  The "matching rule" varies dramatically; the spec/skill should either summarise the four rules explicitly or refer per-algorithm to `docs/requirements.md` §7.4.X. Right now a reviewer reading only the skill could mistakenly conclude that MRAScore uses code-equality (it doesn't).
- **Standard:** `.claude/skills/algorithm-correctness-standards/SKILL.md` §"Score Normalisation".
- **Action:** Skill clarification.
- **Rationale:** Documentation accuracy.
- **Suggested fix:** Expand the phonetic bullet in the SKILL: "Phonetic algorithms: 1.0 if the algorithm's per-algorithm matching rule fires. Specifically: Soundex/NYSIIS use exact code equality; Double Metaphone uses 4-way primary/secondary key matching; MRA uses the NBS-943 6-counter similarity threshold. See docs/requirements.md §7.4.x for the precise rule per algorithm."

---

## Improvement

### [Improvement] dispatch table is populated 23 times via 23 separate dispatch_*.go files — could be one
- **File:** /Users/johnny/Development/fuzzymatch/dispatch_*.go (23 files)
- **Phase introduced:** Phase 2 (pattern)
- **Issue:** 23 single-purpose files, each containing ~30 lines of boilerplate (copyright header, file-header comment, single `var _` line) to perform `dispatch[AlgoX] = XScore`. A single `dispatch.go` file with all 23 registrations would be 50 lines vs the current ~700 lines of boilerplate. The decomposition was deliberate (each phase's implementation plan added its own dispatch wire-up file alongside the algorithm file), but the result is a lot of code surface for a tiny amount of behaviour.
- **Standard:** `.claude/skills/go-coding-standards/SKILL.md` (no specific rule violated).
- **Action:** Code refactor.
- **Rationale:** Maintainability, file-count reduction, easier to verify "all 23 are registered" in code review.
- **Suggested fix:** Consolidate into `dispatch.go` with a single `var _ = func() bool { ... }()` block that performs all 23 registrations.

### [Improvement] Levenshtein and Damerau OSA share identical inner-loop swap and isASCII gate code
- **File:** /Users/johnny/Development/fuzzymatch/levenshtein.go:84-120, /Users/johnny/Development/fuzzymatch/damerau_osa.go:96-129
- **Phase introduced:** Phase 2
- **Issue:** The ASCII-fast-path gate, stack-buffer allocation pattern, and inner-loop-swap pattern are duplicated near-verbatim between `LevenshteinDistance` and `DamerauLevenshteinOSADistance` (and again, with adjustments, in `LongestCommonSubstring` and `lcsLen`). Extracting the gate-and-allocate pattern into a helper would reduce duplication and make the "ASCII fast path policy" testable as one unit.
- **Standard:** None — code quality observation.
- **Action:** Code refactor (optional).
- **Rationale:** Code DRY.

### [Improvement] LCSStrScore allocates two []rune slices for rune-path even when both inputs are ASCII
- **File:** /Users/johnny/Development/fuzzymatch/lcsstr.go:229-244
- **Phase introduced:** Phase 2
- **Issue:** `LCSStrScoreRunes` always allocates `ra := []rune(a)` and `rb := []rune(b)`. If both inputs are ASCII, the byte path would produce the same result with zero allocations. The current implementation doesn't gate on `isASCII(a) && isASCII(b)` to skip the rune conversion. This is a missed optimisation (not a correctness issue).
- **Standard:** `.claude/skills/performance-standards/SKILL.md` — ASCII fast-path budget.
- **Action:** Code refactor.
- **Rationale:** Mirror the pattern from `LevenshteinScoreRunes` (which also lacks this optimisation, so this is a catalogue-wide observation, not LCSStr-specific).

### [Improvement] Test files use ad-hoc absFloat64 helper instead of math.Abs
- **File:** /Users/johnny/Development/fuzzymatch/levenshtein_test.go:37-42 (and likely elsewhere)
- **Phase introduced:** Phase 2
- **Issue:** `levenshtein_test.go` defines `absFloat64` to avoid importing `math` "just for math.Abs (though we use math.Abs below for consistency since math is already imported for quick.Check helpers)." The comment itself notes the inconsistency. Either commit to `math.Abs` and remove `absFloat64`, or commit to the helper and use it everywhere. The current half-and-half is a small style inconsistency.
- **Standard:** None — code style.
- **Action:** Code refactor.
- **Rationale:** Style consistency.

### [Improvement] Soundex's runeAt has subtle continuation-byte validation gap (mentioned above)
- **File:** /Users/johnny/Development/fuzzymatch/soundex.go:275-300
- **Phase introduced:** Phase 7
- **Issue:** See Important finding "runeAt helper in soundex.go reinvents stdlib utf8.DecodeRuneInString" — listed here separately to flag this is also a generic improvement opportunity beyond the validation gap (it removes duplication too).

### [Improvement] Jaro algorithm uses fixed-size stack buffer `[maxJaroStackLen]bool` = 256 even on shorter inputs
- **File:** /Users/johnny/Development/fuzzymatch/jaro.go:152-153
- **Phase introduced:** Phase 2
- **Issue:** `JaroScore` always allocates `var matchA [maxJaroStackLen]bool` (256 bytes) and `var matchB [maxJaroStackLen]bool` (another 256 bytes) on the stack regardless of input length. For a 6-character input (`MARTHA` vs `MARHTA`), 250 bytes of stack space per array is wasted. Not a correctness issue and unlikely to be a measurable performance issue, but the gate `la <= maxJaroStackLen && lb <= maxJaroStackLen` could be tightened to `la <= 32 && lb <= 32` with a `[32]bool` buffer that handles 99% of identifier-style inputs at 1/8th the stack pressure.
- **Standard:** Performance-standards (allocation budgets).
- **Action:** Code refactor (performance).
- **Rationale:** Minor stack-pressure reduction.

### [Improvement] Cosine file comment cites FMA-fusion risk but Levenshtein/Jaro do not warn about it
- **File:** /Users/johnny/Development/fuzzymatch/cosine.go vs /Users/johnny/Development/fuzzymatch/levenshtein.go, jaro.go
- **Phase introduced:** Phase 5
- **Issue:** Cosine extensively documents the FMA-fusion risk (Go issue #17895) and the remediation pattern. Levenshtein and Jaro also use `(x*y) + z` patterns in their float reductions (Levenshtein's `1.0 - float64(dist)/float64(maxLen)` is safe, but Jaro's `(fm/float64(la) + fm/float64(lb) + float64(m-t)/fm) / 3.0` is float-determinism-sensitive). Yet only Cosine flags the risk explicitly. Either the risk is real and the other algorithms should flag it, or the risk is Cosine-specific (because Cosine's `dot` reduction has many more terms in a typical input) and the comment placement is correct.
- **Standard:** Determinism-standards.
- **Action:** Skill clarification or code-comment alignment.
- **Rationale:** Documentation consistency.

### [Improvement] Strcmp95 similar-character table linear-scan over 36 entries could be a 256-entry lookup table
- **File:** /Users/johnny/Development/fuzzymatch/strcmp95.go:198-206
- **Phase introduced:** Phase 2
- **Issue:** `strcmp95SimilarLookup` iterates all 36 entries every call, both byte orientations checked per entry — so 72 comparisons per call. A precomputed 256×256 lookup table would be `2^16 = 65536` bytes (or 256 entries of `[256]float64` = 524288 bytes) which is too large; but a 256-entry `[256]uint64` bitmask (one bit per byte, similarity-credit looked up from a small per-bit mapping) would be 2KB and cut the lookup to O(1). For 36 pairs this is over-engineering; for catalogue uniformity it's worth a thought.
- **Standard:** Performance-standards.
- **Action:** Discuss (likely defer — 36 entries is small).
- **Rationale:** Real performance gain is marginal; documenting the trade-off would close the loop.

### [Improvement] Smith-Waterman-Gotoh uses 6 heap allocations on the rune path even for short inputs
- **File:** /Users/johnny/Development/fuzzymatch/swg.go:451-467
- **Phase introduced:** Phase 3
- **Issue:** `smithWatermanGotohRawRunes` unconditionally heap-allocates 6 `[]float64` rolling rows. For short ASCII inputs going through the Rune* surface (uncommon but possible), this is 6+2 = 8 allocations vs 0 on the byte path. The pattern in `LevenshteinDistanceRunes` is similar: 2 []rune + 2 []int = 4 allocations.
- **Standard:** Performance-standards.
- **Action:** Code fix (stack fast path for rune variant) or doc-only (document the rune-path 8-alloc baseline).
- **Rationale:** Performance.

### [Improvement] Strcmp95 has no Strcmp95ScoreRunes — documented as a deliberate decision
- **File:** /Users/johnny/Development/fuzzymatch/strcmp95.go:73-80
- **Phase introduced:** Phase 2
- **Issue:** Strcmp95 is intentionally ASCII-only — the similar-character table is letter-pair-keyed and has no Unicode equivalent. The file documents this and CONTEXT.md §2 locks the surface. Fine. But this means consumers who pass non-ASCII input get a Strcmp95Score that silently ignores the non-ASCII portion (the byte path includes those bytes in the Jaro match-flag arrays but the similar-character credit cannot fire for them). A consumer wanting Unicode-aware Strcmp95 must pre-normalise. The godoc says this; could be more prominent.
- **Standard:** Documentation.
- **Action:** Doc-only enhancement (godoc).
- **Rationale:** Help users pick the right algorithm.

### [Improvement] Tokeniser-divergence note is repeated verbatim in 4+ algorithm files
- **File:** /Users/johnny/Development/fuzzymatch/token_sort_ratio.go:59-79, /Users/johnny/Development/fuzzymatch/token_set_ratio.go:121-134, /Users/johnny/Development/fuzzymatch/token_jaccard.go:108-117, /Users/johnny/Development/fuzzymatch/monge_elkan.go:107-112
- **Phase introduced:** Phase 6
- **Issue:** The 15-line "OQ-1 RESOLUTION" note about Tokenise being identifier-aware vs RapidFuzz's whitespace-split is duplicated near-verbatim across multiple token-tier algorithm files. Future maintenance: changing the tokenise behaviour requires touching N copies of the note. A shared comment block in tokenise.go that other files reference by single line ("see tokenise.go §X for the tokeniser-divergence rationale") would scale better.
- **Standard:** Documentation DRY.
- **Action:** Code refactor.
- **Rationale:** Documentation consolidation.

### [Improvement] Several algorithms include `_ = opts` in their bodies as a forward-compat placeholder
- **File:** /Users/johnny/Development/fuzzymatch/monge_elkan.go:393 (and likely Scorer-side options)
- **Phase introduced:** Phase 6
- **Issue:** `MongeElkanScore` accepts a `NormalisationOptions` parameter but doesn't use it — the file comment explains this is "for forward-compatibility with the Phase 8 Scorer option." The `_ = opts` line is a clarity comment but it also reads as a smell — the parameter is genuinely unused. Either commit to using opts (apply Normalise internally) or remove it from the signature. The forward-compat argument is weak: Scorer can always wrap MongeElkanScore with its own pre-normalisation.
- **Standard:** Go API ergonomics.
- **Action:** Discuss-phase (api-ergonomics-reviewer should weigh in).
- **Rationale:** API hygiene.

### [Improvement] Partial Ratio TODO comment lacks a GitHub issue reference
- **File:** /Users/johnny/Development/fuzzymatch/partial_ratio.go:146-154
- **Phase introduced:** Phase 6
- **Issue:** `TODO(#TBD): implement sliding-window DP per Bachmann RapidFuzz docs`. The TODO has a placeholder `#TBD` instead of a real issue number. CLAUDE.md §"Workflow — Agent Gates" says "Every TODO must reference a GitHub issue."
- **Standard:** CLAUDE.md (workflow rules).
- **Action:** Create GitHub issue and update TODO.
- **Rationale:** Workflow compliance.

### [Improvement] Damerau-Levenshtein Full uses a full O(m·n) DP table — two-row optimisation deferred to v1.x
- **File:** /Users/johnny/Development/fuzzymatch/damerau_full.go:60-67, 86-87
- **Phase introduced:** Phase 2
- **Issue:** The full DP table is `(m+2) × (n+2)` ints, all heap-allocated. The file comment documents this as a "v1.x performance follow-up." This is a known limitation, not a defect.
- **Standard:** Performance-standards.
- **Action:** Track as a v1.x GitHub issue.
- **Rationale:** Performance debt.

### [Improvement] Cosine clamp at line 385-390 silently returns 1.0 for values slightly above 1.0
- **File:** /Users/johnny/Development/fuzzymatch/cosine.go:385-390
- **Phase introduced:** Phase 5
- **Issue:** The clamp `if cos > 1.0 { return 1.0 }` is correct (IEEE-754 rounding can produce 1.0000000000000002 in degenerate Cauchy-Schwarz cases) but silently. A consumer surfacing the raw Cosine value (e.g. via the Scorer's per-algorithm breakdown) would see exactly 1.0 in cases where the algebraic limit is 1.0. This is the correct behaviour — but if a future bug pushes the value WELL above 1.0 (say 1.5), the clamp would silently hide the bug. A test asserting "the clamp fires only within 1 ULP of 1.0" would catch that scenario.
- **Standard:** Determinism-standards.
- **Action:** Test enhancement.
- **Rationale:** Defensive regression detection.

### [Improvement] Double Metaphone references the SWI-Prolog mirror but the source URL may rot
- **File:** /Users/johnny/Development/fuzzymatch/double_metaphone.go:22-23, 41-42
- **Phase introduced:** Phase 7
- **Issue:** The file cites `https://github.com/SWI-Prolog/packages-nlp/blob/master/double_metaphone.c` as the "stable URL for provenance verification." GitHub URLs to single files on `master` are NOT stable — they shift if the upstream repository reorganises. A pinned commit hash would be more durable.
- **Standard:** Citation hygiene.
- **Action:** Doc-only.
- **Rationale:** Long-term reproducibility of provenance.

### [Improvement] Phonetic algorithms all mention "non-ASCII runes dropped silently" but the spec doesn't pin the behaviour
- **File:** /Users/johnny/Development/fuzzymatch/soundex.go, double_metaphone.go, nysiis.go, mra.go (file headers + per-function godoc)
- **Phase introduced:** Phase 7
- **Issue:** Every phonetic file's godoc mentions that non-ASCII runes are dropped silently. The spec §7.4 doesn't explicitly say "drop silently" vs "panic" vs "return error" — the implementation choice has been made consistently (silent drop) but the spec doesn't pin it. If a consumer reports "I expected é to be normalised to e before phonetic encoding" we'd point at our docs; the spec could be more explicit.
- **Standard:** Spec accuracy.
- **Action:** Spec update.
- **Rationale:** Spec-vs-code alignment.

### [Improvement] Ratcliff-Obershelp recursion depth is documented as O(min(la, lb)) but no explicit cap exists
- **File:** /Users/johnny/Development/fuzzymatch/ratcliff_obershelp.go:81-83, 199-210
- **Phase introduced:** Phase 4
- **Issue:** The file comment says "Recursion depth is O(min(la, lb)) in the worst case." For very long inputs (10⁵ chars) this is genuine stack pressure. Go's default goroutine stack is 8KB and grows dynamically up to 1GB; recursion depth of 10⁵ on small frames is fine, but if a future change adds local-variable state to the recursion (e.g. tracking match positions), the stack could grow significantly. A tail-call-elimination pass or an explicit iterative stack (with `make([]frame, 0, 16)`) would bound the stack pressure.
- **Standard:** Performance-standards / safety.
- **Action:** Code refactor (defer).
- **Rationale:** Future-proofing.

### [Improvement] LCSStrScore vs LongestCommonSubstring (length only vs full substring) inconsistency in stack-buffer use
- **File:** /Users/johnny/Development/fuzzymatch/lcsstr.go:311-319
- **Phase introduced:** Phase 2
- **Issue:** `lcsstrLengthOnly` uses the stack buffer when `lb <= maxStackInputLen` (no swap). `LongestCommonSubstring` uses the same gate. Both could swap for symmetric performance — see Important finding above for LCSStrScore. Listed separately as improvement because the score-only path has no leftmost-tie-break constraint preventing the swap.

### [Improvement] errors.go header comment says ErrHammingLengthMismatch "lands alongside the features that introduce them in later phases" but the feature never lands
- **File:** /Users/johnny/Development/fuzzymatch/errors.go:30-32
- **Phase introduced:** Phase 1
- **Issue:** The errors.go header explicitly promises that `ErrHammingLengthMismatch` will land later. It hasn't. Either delete the promise or fulfil it (the latter is the Critical Hamming finding above).
- **Standard:** Documentation accuracy.
- **Action:** Doc fix (or, linked to Critical, code fix).

### [Improvement] golden_canonical.go / golden_canonical_test.go pin reference vectors but the discoverability of "this is the golden vector for X" is via filename only
- **File:** /Users/johnny/Development/fuzzymatch/golden_canonical.go, /Users/johnny/Development/fuzzymatch/algorithms_golden_test.go
- **Phase introduced:** Phase 2
- **Issue:** `algorithms_golden_test.go` is 55KB; finding the canonical vector for a specific algorithm requires text search. A README in `testdata/golden/` indexing which test covers which algorithm would help future reviewers.
- **Standard:** Documentation discoverability.
- **Action:** Doc enhancement.

---

## Notes for Phase 9 (scan sub-package)

The algorithm catalogue is sound enough to support Phase 9 without rework. The findings above are all addressable in parallel with scan development. The Critical Hamming-API finding is the only item that could affect scan's design (if scan reports unequal-length comparisons specifically), but scan's Phase 9 surface is unlikely to use `HammingDistance`'s error return directly — it composes via the Scorer which uses `HammingScore` (silent-zero). So Phase 9 can proceed.


</details>

---

## 2. algorithm-licensing-reviewer

_Source: `.planning/reviews/algorithm-licensing-FINDINGS.md`_

<details>
<summary>Click to expand full report</summary>

---
status: issues_found
agent: algorithm-licensing-reviewer
scope: entire codebase (phases 1-8)
reviewed: 2026-05-17T08:30:00Z
finding_counts:
  critical: 0
  important: 6
  improvement: 9
  total: 15
---

# Algorithm Licensing Review — Phases 1-8 Gate Review

## Executive Summary

The fuzzymatch codebase is in **good** licensing shape overall. No critical
findings — there are no GPL/LGPL-derived files, no patent-encumbered
algorithms shipped, the runtime-deps allowlist holds, every `.go` file
carries the Apache-2.0 header (CI verifies this: 165 files all pass), and
the Metaphone 3 exclusion is properly documented in `docs/faq.md` and
`docs/requirements.md` §4 with the patent number cited in `algoid.go`.

The findings cluster around **completeness and uniformity** of source-
origin attribution rather than substantive licence risk:

1. The NOTICE file does not yet enumerate the algorithm academic sources
   or the studied Go implementations as the
   `algorithm-licensing-standards` skill explicitly requires (it defers
   this to file-level statements, which themselves are inconsistent
   between Phase 2/3 files and Phase 4+ files).
2. Eight Phase 2 / Phase 3 algorithm files (Levenshtein, Damerau OSA,
   Damerau Full, Hamming, Jaro, Jaro-Winkler, LCSStr, SWG) lack the
   explicit "Source-Origin Statement" block (with the
   GPL/LGPL-provenance and code-copied-verbatim attestations) that
   Phase 4+ files include.
3. `docs/algorithms.md` has "Primary source: TBD — filled in by the
   implementing phase" for **all 23 algorithms** despite all 23 being
   shipped. This is the public-facing per-algorithm reference and
   should now carry the inline citations.
4. The token-tier algorithms (TokenSort, TokenSet, PartialRatio) and the
   Strcmp95 algorithm openly describe their implementations as "fresh
   transcription from RapidFuzz's ... Python source structure" /
   "structurally transcribed from RapidFuzz's MIT-licensed
   `fuzz_py.py::_partial_ratio_impl`". This is borderline against the
   skill's "Translate code line-by-line ... is forbidden" rule, but the
   wording and the open admission suggest the intent matches the
   allowed "performance-pattern inspiration" path. Worth a discuss-phase
   reconciliation to update the skill's wording or document the
   exception.

The Phase 9 (scan sub-package) gate is **GREEN** with these caveats: the
8 Source-Origin-Statement gaps should be backfilled before v1.0 ships,
and the NOTICE file enrichment is a small, high-leverage doc change.

---

## Detailed Findings

### [Important] NOTICE file does not list academic sources or studied Go implementations
- **File:** /Users/johnny/Development/fuzzymatch/NOTICE
- **Phase introduced:** Phase 1 (foundation)
- **Issue:** The skill text in `.claude/skills/algorithm-licensing-standards/SKILL.md` §"In the NOTICE file" specifies the NOTICE file MUST enumerate the academic sources (Levenshtein 1965, Damerau 1964, Jaro 1989, Winkler 1990, …) and the studied permissively-licensed Go ports. The current NOTICE states "This product does not currently incorporate any third-party source code" and defers all attribution to file headers — but the skill's exemplar NOTICE explicitly distinguishes academic-source acknowledgement (always required for shipped algorithms) from copyright-notice attribution (only when code is incorporated). Both axes are missing today: 23 algorithm academic sources are absent from NOTICE; reference Go ports (agnivade/levenshtein, xrash/smetrics, adrg/strutil, hbollon/go-edlib, jellyfish, oubiwann/metaphone, RapidFuzz, biopython, Python difflib) are absent from NOTICE.
- **Standard:** `.claude/skills/algorithm-licensing-standards/SKILL.md` §"Attribution Format" → "In the NOTICE file" (lines 88-113 of the skill).
- **Action:** Code fix (NOTICE update).
- **Rationale:** Apache-2.0 distribution discipline. Consumers reading NOTICE expect to see the academic-source attribution one place, not 23 places. Especially important when the algorithm-licensing-reviewer skill is the explicit reviewer authority — the skill's own exemplar should be reflected in the artefact.
- **Suggested fix:** Update NOTICE with two sections: (a) "Academic sources for similarity algorithms implemented in this library:" with the 22-25 primary-source citations (one per algorithm; q-gram tier shares Jaccard 1912; SWG cites both Smith-Waterman 1981 and Gotoh 1982); (b) "Reference Go implementations studied for cross-validation of test vectors (no code copied):" listing agnivade/levenshtein (MIT), xrash/smetrics (MIT), adrg/strutil (MIT), hbollon/go-edlib (MIT), RapidFuzz (MIT), oubiwann/metaphone (BSD-3), jellyfish (BSD-2), biopython (BSD-3), Python difflib (PSF). The SWI-Prolog public-domain double_metaphone.c mirror, U.S. Census Bureau strcmp95.c (public domain), and OpenRefine Strcmp95.java (Apache-2.0) also belong here.

### [Important] Eight Phase 2/3 algorithm files lack the explicit Source-Origin Statement block
- **File:** /Users/johnny/Development/fuzzymatch/levenshtein.go (and 7 others — damerau_osa.go, damerau_full.go, hamming.go, jaro.go, jarowinkler.go, lcsstr.go, swg.go)
- **Phase introduced:** Phase 2 (core character algorithms) and Phase 3 (SWG)
- **Issue:** The skill text shows every algorithm file's header block should explicitly attest: primary source; studied reference implementations + licences (or "none"); "no GPL/LGPL references consulted"; "no code copied". Phase 4+ files (cosine, double_metaphone, monge_elkan, mra, nysiis, partial_ratio, q_gram, qgram_jaccard, ratcliff_obershelp, sorensen_dice, soundex, strcmp95, token_indel, token_jaccard, token_set_ratio, token_sort_ratio, tversky) all carry this block ("Source-origin discipline" / "Source-Origin Statement" / "Source-origin statement"). Phase 2/3 files do cite the primary source inline ("Source: Levenshtein, V. I. (1965). …") but omit the explicit GPL/LGPL-none and code-copied-none attestations, and do not enumerate any studied Go implementations.
- **Standard:** `.claude/skills/algorithm-licensing-standards/SKILL.md` §"Attribution Format" → "In the algorithm file" + §"Source Origin Statement". The skill's exemplar Levenshtein block shows the expected wording.
- **Action:** Code fix (8 file header updates).
- **Rationale:** Uniformity. The licensing-reviewer's PR-review procedure (in the agent description) specifically looks for these statements; their absence is a BLOCKING finding under the agent's own protocol for PR review of Phase 2/3 algorithms. They were grandfathered in at Phase 2 era before the discipline was tightened; backfilling closes the gap before v1.0.
- **Suggested fix:** Add a "Source-origin discipline" block to each of the 8 files, matching the Phase 4+ style: primary source, cross-validation (or "none — hand-derived reference vectors"), GPL/LGPL provenance: none, code copied verbatim: none, and any MIT/BSD Go ports consulted-for-references-only. For Levenshtein specifically: agnivade/levenshtein (MIT), xrash/smetrics (MIT). For Jaro/Jaro-Winkler: smetrics (MIT). For SWG: biopython/EMBOSS (BSD) — already half-cited as the corpus generator's source. The remaining files (Damerau OSA/Full, Hamming, LCSStr) likely have no studied Go port and the block should explicitly state that.

### [Important] docs/algorithms.md primary-source citations are all "TBD"
- **File:** /Users/johnny/Development/fuzzymatch/docs/algorithms.md (lines 29, 39, 49, 58, 67, 76, 86, 96, 106, 115, 124, 133, 142, 151, 160, 169, 178, 187, 196, 205, 215, 226, 235)
- **Phase introduced:** Phase 1 (scaffolded as a stub) — should have been backfilled phase-by-phase
- **Issue:** Every algorithm section in docs/algorithms.md has `- **Primary source:** TBD — filled in by the implementing phase`. All 23 algorithms have shipped. The skill's §"Attribution Format" → "In docs/algorithms.md" says "Per-algorithm documentation lists the primary source and any studied references explicitly." That requirement is currently unmet for the public docs.
- **Standard:** `.claude/skills/algorithm-licensing-standards/SKILL.md` §"Attribution Format" → "In docs/algorithms.md".
- **Action:** Code fix (docs/algorithms.md backfill).
- **Rationale:** This is the consumer-facing per-algorithm reference; "TBD" defeats its purpose. The data exists in each algorithm file's godoc header — just needs to be transcribed.
- **Suggested fix:** Replace "TBD — filled in by the implementing phase" with the inline citation from each algorithm's `.go` file. Also update Status from "planned (Phase N)" to "implemented in vX.Y.Z" per the line-19 schema.

### [Important] Per-algorithm "Status:" still shows "planned (Phase N)" for all 23 algorithms in docs/algorithms.md
- **File:** /Users/johnny/Development/fuzzymatch/docs/algorithms.md (every `- **Status:** planned …` line)
- **Phase introduced:** Phase 1
- **Issue:** Adjacent to the TBD primary-source finding above. The introductory paragraph (line 17-19) says "Each algorithm's **Status** line is updated to 'implemented in vX.Y.Z' as the corresponding plan lands." That mechanical update has not happened for any of the 23 algorithms.
- **Standard:** `docs/algorithms.md` self-declared protocol (line 17-19).
- **Action:** Code fix (one-liner per algorithm).
- **Rationale:** Misleading public documentation: a reader of docs/algorithms.md would conclude no algorithm is implemented. The CHANGELOG and the actual `.go` files contradict this.
- **Suggested fix:** Rewrite each "Status" line to reflect actual implementation status (e.g. "implemented in v0.2.0").

### [Important] partial_ratio.go "structurally transcribed from RapidFuzz" claim warrants explicit reconciliation against the skill
- **File:** /Users/johnny/Development/fuzzymatch/partial_ratio.go:40-44, 185-189, 310 (and similar phrasing in token_sort_ratio.go:93-97, token_set_ratio.go:150-154)
- **Phase introduced:** Phase 6 (token-based algorithms)
- **Issue:** The file headers state "The three-region iteration pattern + s1_char_set early-skip pattern were transcribed structurally from RapidFuzz's MIT-licensed `fuzz_py.py::_partial_ratio_impl` … No code was copied — the implementation is a fresh Go transcription from the Python source's logical structure only." The variable name `s1_char_set` is identical to RapidFuzz's; the "Region 1 / Region 2 / Region 3" terminology mirrors the upstream comments. The skill at §"Fresh-Implementation Discipline" explicitly forbids "Translate code line-by-line from another implementation into Go" and "Derive variable names, comment phrasing, or structural decisions from another implementation". "Structural transcription" sits in the grey zone — possibly acceptable under the "performance-pattern inspiration" carve-out, but the file's own wording makes the carve-out application non-obvious.
- **Standard:** `.claude/skills/algorithm-licensing-standards/SKILL.md` §"Fresh-Implementation Discipline" (lines 53-71). The forbidden list says "Derive variable names, comment phrasing, or structural decisions from another implementation". The variable `s1_char_set` and the "Region 1/2/3" naming both qualify under a strict reading.
- **Action:** Discuss-phase needed.
- **Rationale:** Either (a) the skill should be tightened/clarified to permit "documented structural inspiration with variable-name preservation when the variable is a load-bearing semantic marker", or (b) the implementation should rename internal variables and comments to be Go-native before v1.0 ships. The MIT-licensed source means even verbatim copying would not be a licence violation — but it would violate the fresh-implementation discipline that AxonOps applies as a hygiene measure (stricter than licence law requires).
- **Suggested fix:** Either rename `s1_char_set` → `shorterByteSet`, `charSet` (already used in places); rename "Region 1/2/3" terminology to "left tail / centred / right tail" (which is already the prose phrasing); OR add a paragraph to the skill explicitly approving this pattern with the partial_ratio.go and token_*ratio.go cases as the named precedent. The latter is the lower-effort path and matches the codebase's actual practice.

### [Important] docs/prior-art-research.md references dlclark/metaphone3 multiple times without flagging the patent exclusion
- **File:** /Users/johnny/Development/fuzzymatch/docs/prior-art-research.md:90, 112, 212, 253, 360, 406
- **Phase introduced:** Phase 1 (research artefact)
- **Issue:** This is a historical research/survey document that lists `dlclark/metaphone3` (MIT) as one of the candidate Go phonetic implementations — including a positive description ("Most accurate; handles `EncodeVowels`, `EncodeExact` opts"). The same document discusses Metaphone 3 as "Most accurate" without noting that it has been explicitly EXCLUDED per the patent screen. A future contributor reading this document in isolation might conclude Metaphone 3 is on the roadmap.
- **Standard:** `.claude/skills/algorithm-licensing-standards/SKILL.md` §"The Metaphone 3 Precedent" (lines 132-138) — the precedent is meant to be visible at every decision point.
- **Action:** Code fix (docs edit).
- **Rationale:** Prevents accidental re-introduction. The exclusion is documented in docs/faq.md and docs/requirements.md §4, but a contributor working from the prior-art-research survey might miss it.
- **Suggested fix:** Add a single-sentence callout near each `dlclark/metaphone3` mention in docs/prior-art-research.md: "NOTE: Metaphone 3 is EXCLUDED from this catalogue per docs/requirements.md §4 (U.S. Patent 7,440,941). See docs/faq.md for the patent screen rationale." Or add a top-of-document banner noting that the document is a historical survey and the canonical exclusion list lives in §4 of the requirements doc.

### [Improvement] NOTICE file mentions a hypothetical future update protocol but doesn't note the standard's exemplar
- **File:** /Users/johnny/Development/fuzzymatch/NOTICE (paragraph starting "When third-party source IS incorporated...")
- **Phase introduced:** Phase 1
- **Issue:** The current NOTICE has a forward-looking paragraph anticipating future third-party incorporation, but the standard's exemplar NOTICE shows that the academic-source listing is NOT contingent on incorporation — it is the standard format for an Apache-2.0 library that re-implements published algorithms. The two interpretations of NOTICE are at odds: (a) NOTICE attributes only INCORPORATED third-party code (current file); (b) NOTICE acknowledges academic sources and consulted references whether or not code was incorporated (skill exemplar).
- **Standard:** `.claude/skills/algorithm-licensing-standards/SKILL.md` §"In the NOTICE file".
- **Action:** Skill clarification OR code fix — depending on which interpretation the project owner prefers.
- **Rationale:** The current NOTICE is internally consistent and Apache-2.0-compliant in the narrow legal sense (NOTICE-required attribution is only for incorporated code under Apache-2.0). The skill's expansive interpretation is stricter — it treats academic-source acknowledgement as an editorial obligation distinct from the legal NOTICE function. Worth pinning down before v1.0 documentation freezes.
- **Suggested fix:** If the project takes the legal-minimum interpretation, update the skill text to reflect that NOTICE is only for INCORPORATED code and move the academic-source listing to docs/algorithms.md (already partially scoped under finding 3 above). If the project takes the expansive interpretation, update NOTICE per finding 1 above.

### [Improvement] gen-token-ratio-cross-validation.py rapidfuzz pin is in script header; could be cross-referenced from NOTICE
- **File:** /Users/johnny/Development/fuzzymatch/scripts/gen-token-ratio-cross-validation.py
- **Phase introduced:** Phase 6
- **Issue:** The token-tier cross-validation depends on RapidFuzz 3.14.5 (pinned in the script with a runtime assertion). RapidFuzz is MIT-licensed and is not redistributed in the library — only its outputs are committed as a JSON corpus. The relationship is clean. But the NOTICE file does not mention the corpus origin, and consumers regenerating the corpus need to know the pinned version and the licence.
- **Standard:** None directly — this is a hygiene improvement.
- **Action:** Code fix (NOTICE addition).
- **Rationale:** Reproducibility and provenance trail. Combine with finding 1 (NOTICE enrichment).
- **Suggested fix:** Include in NOTICE: "Cross-validation corpora are generated from external implementations pinned in scripts/gen-*-cross-validation.py: RapidFuzz 3.14.5 (MIT), biopython (BSD-3-Clause), Python stdlib difflib (PSF), jellyfish 1.2.1 (BSD-2-Clause), metaphone 0.6 / oubiwann (BSD-3-Clause)."

### [Improvement] BDD test module (tests/bdd/go.sum) pulls godog dependency chain — chain includes Apache-2.0, MIT, BSD components; not summarised anywhere
- **File:** /Users/johnny/Development/fuzzymatch/tests/bdd/go.mod, /Users/johnny/Development/fuzzymatch/tests/bdd/go.sum
- **Phase introduced:** Phase 8 (Scorer phase first uses the BDD module per spec; earlier phases scaffolded it)
- **Issue:** The BDD module pulls godog (Apache-2.0), testify (MIT), goleak (MIT), gherkin/messages (MIT), hashicorp/go-immutable-radix + hashicorp/go-memdb + golang-lru (MPL-2.0 or BSD — needs check), gofrs/uuid, pflag (BSD-3), yaml.v3 (MIT/Apache-2.0). None of these is incorporated into the published library — they live in a separate go.mod. But there is no centralised acknowledgement of the dev-time dependency chain.
- **Standard:** None directly — Apache-2.0 NOTICE obligation does not extend to test-only dependencies that are not redistributed. Hygiene improvement.
- **Action:** Code fix (lightweight NOTICE-DEV.md or section in NOTICE under "Development-time dependencies").
- **Rationale:** Belt and braces. Audit-friendly. Consumers asking "what does fuzzymatch transitively depend on?" should be able to read a single document.
- **Suggested fix:** Either inline a "Development-time dependencies (not redistributed)" subsection in NOTICE, or create CONTRIBUTING.md → "Third-party dev-time dependencies" subsection with the BDD module dependency tree.

### [Improvement] Hashicorp dependencies in BDD go.sum should be licence-spot-checked
- **File:** /Users/johnny/Development/fuzzymatch/tests/bdd/go.sum (hashicorp/go-immutable-radix v1.3.1, hashicorp/go-memdb v1.3.4, hashicorp/golang-lru v0.5.4, hashicorp/go-uuid v1.0.2)
- **Phase introduced:** Phase 8 (transitive via godog)
- **Issue:** Hashicorp moved many of their libraries from MPL-2.0 to BSL/MPL hybrid in 2023-2024. For this project the licensing impact is zero (test-only, not redistributed), but it's worth a one-time confirmation that the pinned versions here are still under MPL-2.0 or earlier permissive terms.
- **Standard:** No direct standard — operational hygiene.
- **Action:** Discuss-phase needed (one-time confirmation, no code change required unless a version is non-permissively licensed).
- **Rationale:** Hashicorp's BSL transition is the highest-profile permissive→restrictive licence change in the Go ecosystem in recent years. A two-minute check now prevents a surprise later.
- **Suggested fix:** `go-immutable-radix` v1.3.1, `go-memdb` v1.3.4, `golang-lru` v0.5.4 are MPL-2.0 per their respective LICENSE files on GitHub at those tags. Document this in a one-line comment in tests/bdd/go.mod or a CONTRIBUTING.md addendum.

### [Improvement] doc.go's "Patent and licence hygiene" paragraph is two sentences — could be three
- **File:** /Users/johnny/Development/fuzzymatch/doc.go:48-50
- **Phase introduced:** Phase 1
- **Issue:** The packagelevel godoc has: "Patent and licence hygiene: this library is Apache 2.0 and incorporates no GPL/LGPL-derived code. Patent-encumbered algorithms (notably Metaphone 3) are excluded by design." This is correct but could note where to find the per-algorithm provenance trail (the file headers, NOTICE, docs/faq.md).
- **Standard:** None.
- **Action:** Code fix (one-sentence addition).
- **Rationale:** Discoverability. The pkg.go.dev rendering of doc.go is the consumer's first point of contact.
- **Suggested fix:** Append: "Per-algorithm source attribution, studied references, and licence provenance are documented in each implementation file's header block; the consolidated list is in the repository NOTICE file."

### [Improvement] q_gram.go has Source-Origin discipline header but the shared kernel could clarify its inheritance to four callers
- **File:** /Users/johnny/Development/fuzzymatch/q_gram.go:57-70
- **Phase introduced:** Phase 5
- **Issue:** The shared q-gram extraction helper has its own Source-Origin block (Ukkonen 1992). Good. Each consumer (qgram_jaccard.go, sorensen_dice.go, cosine.go, tversky.go) also has its own block citing its own primary source. The relationship is clean. But a one-line cross-reference in q_gram.go ("Consumed by: qgram_jaccard.go, sorensen_dice.go, cosine.go, tversky.go") would make the kernel's downstream surface obvious to a reviewer auditing the q-gram tier as a whole.
- **Standard:** None — uniformity nudge.
- **Action:** Code fix (one-line edit).
- **Rationale:** Aid auditability. Same applies to token_indel.go vis-à-vis its three consumers.
- **Suggested fix:** Add a "Consumed by:" line to each shared-kernel file.

### [Improvement] tests/bdd/steps/*.go don't carry source-origin context but probably should not need to
- **File:** /Users/johnny/Development/fuzzymatch/tests/bdd/steps/algorithms_steps.go, /Users/johnny/Development/fuzzymatch/tests/bdd/steps/scorer_steps.go
- **Phase introduced:** Phase 5 and 8
- **Issue:** Step definition files carry Apache-2.0 headers but no algorithm-source-attribution block. They don't need one — they exercise the public API only — but a one-line statement to that effect in the file header would prevent a future reviewer from flagging them. Pure hygiene.
- **Standard:** None.
- **Action:** No action recommended; documented for completeness.
- **Rationale:** Step definitions don't implement algorithms; they call them. No source-origin attestation is required.

### [Improvement] verify-license-headers.sh checks for the header string but doesn't verify "AxonOps" copyright line specifically
- **File:** /Users/johnny/Development/fuzzymatch/scripts/verify-license-headers.sh:17
- **Phase introduced:** Phase 1
- **Issue:** The script greps for the literal "Licensed under the Apache License, Version 2.0". This is sufficient to catch missing-header files but does not catch a file that has a third-party Apache-2.0 header with a different copyright line. Today this is a hypothetical concern — every file is freshly authored with the AxonOps copyright — but if a future contribution adds a file derived from another Apache-2.0 project, the script wouldn't flag the discrepancy.
- **Standard:** None — operational hardening.
- **Action:** Code fix (script enhancement) — low priority.
- **Rationale:** Defence in depth. The script is the first line of CI defence; tightening it slightly catches a class of provenance regressions.
- **Suggested fix:** Add a second grep for "Copyright 20XX AxonOps Limited" (where 20XX is a 4-digit year) — flag any file matching the Apache-2.0 header but NOT the AxonOps copyright as needing reviewer attention. Could be a separate exit code for soft-warn-but-not-fail to avoid breaking CI on legitimate third-party Apache-2.0 file additions.

### [Improvement] No mention of cosign / SBOM / attestation provenance in NOTICE
- **File:** /Users/johnny/Development/fuzzymatch/NOTICE
- **Phase introduced:** Phase 1 (release machinery scaffolded)
- **Issue:** Release artefacts will be signed with cosign keyless via GitHub OIDC + SBOM attached per the spec/STACK.md. NOTICE could include a brief pointer to the verification protocol — useful for downstream consumers running supply-chain hygiene checks.
- **Standard:** None — operational improvement.
- **Action:** Code fix (NOTICE addition) — low priority, can wait until first signed release.
- **Rationale:** Reduces consumer friction.
- **Suggested fix:** Add to NOTICE: "Release artefacts are signed using sigstore cosign (keyless, GitHub OIDC) and accompanied by an SPDX-JSON SBOM. See <release-verification-doc-link> for the verification protocol."

---

## What was verified clean

- **Apache-2.0 file headers:** 165 `.go` files, all carrying the AxonOps Apache-2.0 header. `verify-license-headers.sh` passes.
- **Runtime dependency allowlist:** root `go.mod` has exactly one curated runtime dep (`golang.org/x/text v0.37.0`) and the main module. `verify-no-runtime-deps.sh` passes. No transitive cgo. `CGO_ENABLED=0` builds are clean.
- **GPL/LGPL-derived code:** none. Searched every `.go` file for GPL/LGPL/AGPL/copyleft references; all hits are either "GPL/LGPL provenance: none" attestations (positive) or absent. Phase 4+ algorithm files explicitly attest "no GPL/LGPL references consulted".
- **Metaphone 3 (U.S. Patent 7,440,941):** explicitly EXCLUDED. Documented in docs/faq.md §"Why no Metaphone 3?", docs/requirements.md §4 (Out of Scope), algoid.go:163-164, doc.go:49 (package godoc). No code shipped. No imports of `github.com/dlclark/metaphone3` anywhere. The precedent is preserved.
- **Patent screen for shipped algorithms:** all 23 algorithms covered by published academic papers dated 1912-2000 with no associated patent (Soundex's 1918 patents expired 1935 and are noted in soundex.go). No active patent risk identified.
- **Python generator scripts:** all four (gen-phonetic-cross-validation.py, gen-ratcliff-obershelp-cross-validation.py, gen-swg-cross-validation.py, gen-token-ratio-cross-validation.py) carry the Apache-2.0 header.
- **BDD module isolation:** test-only dependencies (godog Apache-2.0, testify MIT, goleak MIT, cucumber/gherkin MIT, hashicorp libraries MPL-2.0, gofrs/uuid MIT, pflag BSD-3, yaml.v3 MIT/Apache-2.0) all live in tests/bdd/go.mod with a `replace` directive — never reach the root go.mod or the published library surface.
- **No testify in root tests:** verified by inspection — root test files explicitly note "Stdlib testing only — no testify in root tests".
- **Phonetic algorithm Source-Origin Statements:** all 4 phonetic algorithms (Soundex, Double Metaphone, NYSIIS, MRA) have explicit Source-Origin Statement blocks per the skill's strictest standard, including the GPL/LGPL-none + code-copied-none attestations + the patent screen acknowledgement. Phase 7 was implemented under the tightest discipline.
- **Strcmp95 public-domain reference (Census Bureau strcmp95.c):** correctly documented as public domain (U.S. Government work). The OpenRefine Strcmp95.java (Apache-2.0) is correctly cited for prose-level tie-breaks only.

---

## Phase 9 Gate Decision

**GREEN with caveats.** The scan sub-package implementation in Phase 9 can
proceed. The findings above are all backfill / documentation /
uniformity work on already-shipped algorithms — none block new algorithm
implementation in the scan layer because scan does not itself implement
new similarity algorithms.

Recommendation: backfill the 8 Phase 2/3 Source-Origin Statements
(Important finding 2), enrich the NOTICE file (Important finding 1),
and update docs/algorithms.md (Important findings 3 + 4) as a pre-v1.0
documentation sprint, ideally before the v0.6.x integration shakedown
in Phase 7 of the original roadmap (= integration with downstream
consumer axonops/audit).

</details>

---

## 3. algorithm-performance-reviewer

_Source: `.planning/reviews/algorithm-performance-FINDINGS.md`_

<details>
<summary>Click to expand full report</summary>

---
status: issues_found
agent: algorithm-performance-reviewer
scope: entire codebase (phases 1-8)
reviewed: 2026-05-17T10:00:00Z
finding_counts:
  critical: 6
  important: 13
  improvement: 8
  total: 27
---

# Algorithm Performance Review — Phases 1–8 Comprehensive Findings

Platform: darwin/arm64 (Apple M2), Go 1.26.3.
Benchmark command: `go test -bench=. -benchmem -count=3 -run=^$ ./...`
Baseline: `bench.txt` (1296 lines, count=10 per benchmark, darwin/arm64 Apple M2).

---

## Critical Findings

### [Critical] Scorer allocation budget breach: 12/34 allocs vs §14.2 ≤ 8
- **File:** `/Users/johnny/Development/fuzzymatch/scorer.go`, `/Users/johnny/Development/fuzzymatch/scorer_bench_test.go`
- **Phase introduced:** Phase 8
- **Issue:** `DefaultScorer().Score("abc","abcd")` = 12 allocs/op. `DefaultScorer().Score(scorerAMedium, scorerBMedium)` = 34 allocs/op. Spec §14.2 mandates ≤ 8 allocs/op for `DefaultScorer().Score` on ASCII inputs ≤ 50 chars. §20 acceptance criteria lists "All section-14 budgets met" as a v1.0 ship gate.
- **Standard:** `docs/requirements.md` §14.2, §20; performance-standards.md "Scorer Budgets"
- **Action:** Discuss-phase needed — the §14.2 budget of ≤ 8 is mathematically impossible given the summed per-algorithm §14.1 budgets for the 6-algorithm DefaultScorer (summed floor = 16). Options: (A) revise §14.2 to ≤ 16 for Short + ≤ 40 for Medium, (B) fix DoubleMetaphone Builder→[4]byte to save 4 allocs bringing Short to exactly 8. See 08-PERFORMANCE-REVIEW.md for full analysis.
- **Rationale:** Acceptance criteria cannot be declared met with a documented breach; benchstat CI is blind to Scorer regressions until bench.txt is updated.
- **Suggested fix:** Implement the [4]byte replacement in `DoubleMetaphoneKeys` (saves 4 allocs) + revise §14.2 budget to ≤ 16 for ASCII Short. See 08-PERFORMANCE-REVIEW.md PERF-03 for the exact patch pattern.

---

### [Critical] Scorer benchmarks absent from bench.txt — benchstat CI blind
- **File:** `/Users/johnny/Development/fuzzymatch/bench.txt`, `/Users/johnny/Development/fuzzymatch/scorer_bench_test.go`
- **Phase introduced:** Phase 8
- **Issue:** `bench.txt` (1296 lines) contains zero `BenchmarkDefaultScorer_*` entries. Six scorer benchmarks shipped in Phase 8 (`BenchmarkDefaultScorer_Score_ASCII_Short`, `_ASCII_Medium`, `_ASCII_Long`, `_Unicode_Short`, `_ScoreAll_ASCII_Short`, `_Match_ASCII_Short`) but `scripts/update-bench-txt.sh` was never run. Any future commit that doubles Scorer allocation count or wall time produces no benchstat signal. This is an unconditional regression-detection gap for the entire Scorer layer.
- **Standard:** performance-standards.md "Updating bench.txt"; `docs/requirements.md` §14.4
- **Action:** Code fix — run `go test -bench=BenchmarkDefaultScorer -benchmem -count=10 -run=^$ ./...` on the self-hosted runner, append results to `bench.txt`, commit before v1.0 tag.
- **Rationale:** The benchstat CI job (which catches >10% regressions) has no baseline rows to compare against. Any regression to the Scorer introduced after Phase 8 is invisible to CI.
- **Suggested fix:** `go test -bench=BenchmarkDefaultScorer -benchmem -count=10 -run=^$ ./... >> bench.txt` on the self-hosted benchmark runner; verify `benchstat bench.txt` exits 0.

---

### [Critical] DamerauLevenshteinFull ASCII Short/Medium: 1 alloc vs §14.1 0-alloc budget
- **File:** `/Users/johnny/Development/fuzzymatch/damerau_full.go`
- **Phase introduced:** Phase 2
- **Issue:** `BenchmarkDamerauLevenshteinFullScore_ASCII_Short` = 1 alloc/op (128 B/op), `BenchmarkDamerauLevenshteinFullScore_ASCII_Medium` = 1 alloc/op (21760 B/op). Spec §14.1 specifies "0 allocations" for Damerau-Levenshtein Full on ASCII ≤ 50 char inputs. The implementation uses a full O(m·n) heap-allocated DP table for all inputs, including ASCII Short/Medium, instead of the two-row + stack optimisation. The file godoc explicitly documents this as "v1.x follow-up" but the §14.1 budget has no such exception clause.
- **Standard:** `docs/requirements.md` §14.1 "Damerau-Levenshtein Full: < 3 µs per call, 0 allocations"; performance-standards.md "Two-Row DP Optimisation"
- **Action:** Discuss-phase needed — either (A) implement the two-row + auxiliary anchor-table optimisation for ASCII Short/Medium inputs to achieve 0 allocs, or (B) amend §14.1 to document the O(m·n) space requirement as an accepted exception for the Lowrance-Wagner formulation. The wall-time budget (< 3 µs) passes (64 ns at Short, 8.2 µs at Medium — note Medium exceeds 3 µs budget too, see finding below).
- **Rationale:** A 1-alloc/op number in bench.txt for a budget of 0 allocs is a documented known gap that blocks §20 acceptance criteria.
- **Suggested fix:** Implement the stack-allocated scratch buffer pattern for the Short path (inputs ≤ maxStackInputLen): allocate a `[(maxStackInputLen+2)*(maxStackInputLen+2)]int` stack array as the DP table when both m and n fit within the threshold, falling back to `make([]int, size)` for larger inputs.

---

### [Critical] DamerauLevenshteinFull ASCII Medium wall-time: 8.2 µs vs §14.1 < 3 µs budget
- **File:** `/Users/johnny/Development/fuzzymatch/damerau_full.go`
- **Phase introduced:** Phase 2
- **Issue:** `BenchmarkDamerauLevenshteinFullScore_ASCII_Medium` (50-char inputs) = 8218 ns/op in bench.txt. Spec §14.1 states "Damerau-Levenshtein Full: < 3 µs per call, 0 allocations" for ASCII ≤ 50 chars. The full O(m·n) DP table allocation for every call (including 50-char inputs) drives both the alloc count and the wall-time breach. A 50×50 table requires (52)×(52)×8 = ~21 KB — matching the 21760 B/op in bench.txt.
- **Standard:** `docs/requirements.md` §14.1; performance-standards.md "Two-Row DP Optimisation"
- **Action:** Code fix — the Lowrance-Wagner formulation cannot reduce to a standard two-row DP because the transposition term `D[l-1,k-1]` references arbitrary prior rows. However the `da[256]int` auxiliary table already lives on the stack; the full DP table itself is the bottleneck. A column-at-a-time compression or a condensed sparse representation for the transposition anchor rows is the approach. Alternatively, amend §14.1 to document the DL-Full medium budget as "< 10 µs per call" (matching the actual 8.2 µs) with an explicit note that O(m·n) space is required by the Lowrance-Wagner algorithm.
- **Rationale:** The spec budget of < 3 µs for DL-Full at Medium is physically unachievable without changing the algorithm itself (OSA achieves it via three-row rolling DP, but Full requires the full table or a column-sparse variant). The budget was likely copied from OSA without accounting for the structural difference.
- **Suggested fix:** Amend §14.1 to read "Damerau-Levenshtein Full: < 10 µs per call at ASCII ≤ 50 chars (O(m·n) space is structurally required by the Lowrance-Wagner algorithm; 0 allocations is achievable at Short ≤ 10 chars via stack-resident DP table)." Update `damerau_full.go` godoc accordingly.

---

### [Critical] DoubleMetaphone Score: 6 allocs vs §14.1 ≤ 2-alloc budget
- **File:** `/Users/johnny/Development/fuzzymatch/double_metaphone.go`
- **Phase introduced:** Phase 7
- **Issue:** `BenchmarkDoubleMetaphoneScore_ASCII_Short` = 6 allocs/op (48 B/op). Spec §14.1 states "Double Metaphone: < 2 µs per call, ≤ 2 allocations." The 6 allocs come from: 2× `dmPrep` → `string(stackBuf[:n])` (2 allocs despite stack buffer), 2× `p.String()` from `strings.Builder`, 2× `alt.String()` from `strings.Builder`. The `strings.Builder.String()` method always heap-allocates the returned string even when the builder's internal buffer is small (4 bytes max for DM keys).
- **Standard:** `docs/requirements.md` §14.1 "Double Metaphone: < 2 µs per call, ≤ 2 allocations"; performance-standards.md "Per-Algorithm Budgets"
- **Action:** Code fix — replace `strings.Builder` primary/secondary key accumulators in `DoubleMetaphoneKeys` with `[dmMaxLen]byte` arrays + length counters. Return `string(pBuf[:pLen]), string(altBuf[:altLen])` — still 2 allocs for the returned strings, but eliminates the 4 Builder.String() allocs. Total drops from 6 → 2. This change has zero correctness risk (keys are bounded at 4 bytes by `dmMaxLen`).
- **Rationale:** This is a 3× budget overshoot and is directly fixable with a `[4]byte` pattern. The fix also saves 4 allocs in the DefaultScorer path, which is required to approach the §14.2 budget (see Scorer finding).
- **Suggested fix:** In `double_metaphone.go`, replace `var primary, secondary strings.Builder` with `var pBuf, altBuf [dmMaxLen]byte` and `var pLen, altLen int`; replace `dmAdd` to write into the byte arrays; return `string(pBuf[:pLen]), string(altBuf[:altLen])`.

---

### [Critical] Normalise ASCII Short: 1 alloc vs §14.3 0-alloc budget
- **File:** `/Users/johnny/Development/fuzzymatch/normalise.go`
- **Phase introduced:** Phase 1
- **Issue:** `BenchmarkNormalise_ASCII_Short` = 1 alloc/op (16 B/op); `BenchmarkNormalise_DefaultOptions_Short` = 1 alloc/op (24 B/op). Spec §14.3 states "Normalise ASCII input ≤ 50 chars: < 200 ns, 0 allocations (stack buffer)." The `string(buf)` conversion at `normalise.go:249` unconditionally allocates a heap string because Go string values must be heap-backed when returned from a function — the stack-resident `buf` backing array cannot be referenced by the returned string (its lifetime would not outlive the function call). The `make([]byte, 0, len(s)*2+1)` does NOT escape (confirmed by escape analysis), but the return conversion does.
- **Standard:** `docs/requirements.md` §14.3; performance-standards.md "Per-Algorithm Budgets"
- **Action:** Discuss-phase needed — the 0-alloc claim in §14.3 is structurally unachievable without `unsafe.String` (which is forbidden by go-coding-standards). Options: (A) amend §14.3 to "≤ 1 allocation (output string heap escape)" for ASCII inputs; (B) document in the normalise.go godoc that the stack buffer eliminates growth allocs but the output string itself requires 1 allocation. No code change required.
- **Rationale:** The committed bench.txt baseline already shows 1 alloc, which is inconsistent with the spec's stated 0-alloc target. Leaving this discrepancy unresolved confuses future algorithm-performance reviews.
- **Suggested fix:** Amend `docs/requirements.md` §14.3 to read: "`Normalise` ASCII input ≤ 50 chars: < 200 ns, ≤ 1 allocation (output string; stack buffer avoids growth allocs but the returned string is heap-backed)." Update `normalise.go` godoc accordingly.

---

## Important Findings

### [Important] Q-Gram algorithms (Jaccard, Sørensen-Dice, Cosine, Tversky): 6–7 allocs at Medium vs §14.1 ≤ 4-alloc budget
- **File:** `/Users/johnny/Development/fuzzymatch/q_gram.go`, `/Users/johnny/Development/fuzzymatch/qgram_jaccard.go`, `/Users/johnny/Development/fuzzymatch/sorensen_dice.go`, `/Users/johnny/Development/fuzzymatch/cosine.go`, `/Users/johnny/Development/fuzzymatch/tversky.go`
- **Phase introduced:** Phase 5
- **Issue:** `BenchmarkQGramJaccardScore_ASCII_Medium` = 6 allocs/op; `BenchmarkSorensenDiceScore_ASCII_Medium` = 6 allocs/op; `BenchmarkTverskyScore_ASCII_Medium` = 6 allocs/op; `BenchmarkCosineScore_ASCII_Medium` = 7 allocs/op. Spec §14.1 budget is "≤ 4 allocations" for these algorithms. The 2–3 extra allocs above budget at Medium are caused by map rehashing: `extractQGrams` sizes the initial map at `len(s)-n+1` but when multiple windows hash to the same bucket, the map grows beyond the initial capacity, triggering additional internal rehash allocs. For a 50-char string with n=3, this yields ~47 trigrams, and the map grows 1–2 times during filling.
- **Standard:** `docs/requirements.md` §14.1 "Q-Gram Jaccard, Sørensen-Dice, Cosine, Tversky: < 5 µs per call, ≤ 4 allocations"
- **Action:** Code fix (optional for v0.x, required for v1.0). Options: (A) use a larger capacity multiplier (e.g. `len(s)-n+1)*2`) to reduce rehash probability; (B) implement a small-input fast path using a fixed-size `[64][3]byte` sorted array for trigrams when `len(s) <= 20` (avoids map allocation entirely); (C) amend §14.1 to "≤ 6 allocations" reflecting the observed behaviour.
- **Rationale:** The 6-alloc Medium result is documented in bench.txt as the baseline, meaning CI will accept it as the new floor. However, it breaches the spec budget, which future reviewers will flag again.
- **Suggested fix:** In `q_gram.go`, change the map initial capacity to `(len(s)-n+1) + (len(s)-n+1)/4` (add 25% headroom) to reduce rehash probability without over-allocating for short inputs.

---

### [Important] Token Sort Ratio: 14 allocs at ASCII Short vs §14.1 ≤ 4-alloc budget
- **File:** `/Users/johnny/Development/fuzzymatch/token_sort_ratio.go`, `/Users/johnny/Development/fuzzymatch/tokenise.go`
- **Phase introduced:** Phase 6
- **Issue:** `BenchmarkTokenSortRatioScore_ASCII_Short` = 14 allocs/op (474 B/op) in bench.txt. Spec §14.1 states "Token Sort Ratio, Token Set Ratio, Token Jaccard: < 5 µs per call, ≤ 4 allocations." The 14-alloc count is driven by `Tokenise` producing multiple heap strings from rune-to-string conversion per token, plus the intermediate sorted string construction. For a short input like "hello world" (2 tokens), this breaks down as: 2 Tokenise output slices + per-token strings + sorted join + Levenshtein intermediate.
- **Standard:** `docs/requirements.md` §14.1 "Token Sort Ratio … ≤ 4 allocations"; performance-standards.md "Per-Algorithm Budgets"
- **Action:** Code fix — the Tokenise function at `/Users/johnny/Development/fuzzymatch/tokenise.go` allocates per-token strings via `string(rs)` conversion. A pooled token buffer or a `strings.Builder`-based join avoiding intermediate per-token strings would reduce alloc count. However, given the inherent cost of sorting and joining tokens, 4 allocs is a very aggressive target; a realistic budget is ≤ 8 allocs for short inputs.
- **Rationale:** The bench.txt baseline shows 14 allocs, which is 3.5× the spec budget. This is tracked in bench.txt without having been flagged, meaning CI cannot catch a further regression relative to the spec target.
- **Suggested fix:** Amend §14.1 to "Token Sort Ratio, Token Set Ratio, Token Jaccard: < 5 µs per call, ≤ 10 allocations (token string allocations are proportional to token count)" until a Tokenise pooling optimisation is implemented.

---

### [Important] Token Set Ratio: 9 allocs at ASCII Short vs §14.1 ≤ 4-alloc budget
- **File:** `/Users/johnny/Development/fuzzymatch/token_set_ratio.go`
- **Phase introduced:** Phase 6
- **Issue:** `BenchmarkTokenSetRatioScore_ASCII_Short` = 9 allocs/op (245 B/op). Same budget violation as Token Sort Ratio. Set construction (two `map[string]struct{}` + intersection/difference logic) adds 2 extra map allocs on top of the base Tokenise cost.
- **Standard:** `docs/requirements.md` §14.1 "Token Set Ratio … ≤ 4 allocations"
- **Action:** Code fix (same family as Token Sort Ratio). Alternatively amend budget.
- **Rationale:** 9 allocs is 2.25× the 4-alloc budget. Bench.txt records it as the baseline.
- **Suggested fix:** Same as Token Sort Ratio — revise §14.1 budget upward to reflect the inherent token-string allocation cost.

---

### [Important] Tokenise: 4 allocs at ASCII Short vs §14.3 ≤ 2-alloc budget
- **File:** `/Users/johnny/Development/fuzzymatch/tokenise.go`
- **Phase introduced:** Phase 6 (tokenise)
- **Issue:** `BenchmarkTokenise_ASCII_Short` = 4 allocs/op (73 B/op). Spec §14.3 states "Tokenise ASCII input ≤ 50 chars: < 500 ns, ≤ 2 allocations (token slice + storage)." The actual 4 allocs exceed the budget by 2. The extra allocs come from per-token `string(rs)` conversions where `rs` is a rune sub-slice — each token's rune run is converted to a heap string.
- **Standard:** `docs/requirements.md` §14.3 "Tokenise ASCII input ≤ 50 chars: < 500 ns, ≤ 2 allocations"
- **Action:** Code fix — on the ASCII fast path, token boundaries can be computed as byte offsets into the original string, and each token can be returned as a slice of the input string (zero-copy, no per-token alloc). This requires the Tokenise function to operate on bytes directly for ASCII inputs. Alternatively amend §14.3 to ≤ 4 allocs.
- **Rationale:** Tokenise is on the hot path for all 5 token-based algorithms. Each call allocates per-token strings. For ASCII inputs, a substring-based approach avoids these allocs.
- **Suggested fix:** In `tokenise.go`, add an ASCII fast path that identifies token boundaries as byte-index ranges and returns substrings of the original input string (using `s[lo:hi]`) rather than creating new strings from rune slices.

---

### [Important] Soundex/NYSIIS/MRA Code functions: 1 alloc vs §14.1 0-alloc budget
- **File:** `/Users/johnny/Development/fuzzymatch/soundex.go`, `/Users/johnny/Development/fuzzymatch/nysiis.go`, `/Users/johnny/Development/fuzzymatch/mra.go`
- **Phase introduced:** Phase 7
- **Issue:** `BenchmarkSoundexCode_ASCII_Short` = 1 alloc/op (4 B/op); `BenchmarkNYSIISCode_ASCII_Short` = 1 alloc/op (4 B/op); `BenchmarkMRACode_ASCII_Short` = 1 alloc/op (4 B/op). Spec §14.1 states "Soundex, NYSIIS, MRA: < 500 ns per call, 0 allocations (stack-allocated code buffer)." All three functions use stack-allocated intermediate buffers (`[64]byte`, `[6]byte` etc.) but return the result via `string(result[:n])` or `string(result[:])` which necessarily allocates a heap-backed string.
- **Standard:** `docs/requirements.md` §14.1 "Soundex, NYSIIS, MRA: < 500 ns per call, 0 allocations (stack-allocated code buffer)"
- **Action:** Discuss-phase needed — same structural situation as Normalise. The `string(buf)` conversion is unavoidable without `unsafe.String`. The spec's "0 allocations" claim is incorrect. Options: (A) amend §14.1 to "≤ 1 allocation (output string)"; (B) note that `MRAScore` and `SoundexScore` (which call Code twice and compare strings) budget at 2 allocs for the two Code calls — which matches bench.txt values.
- **Rationale:** `BenchmarkSoundexScore_ASCII_Short` = 2 allocs/op (matching 2 SoundexCode calls), `BenchmarkMRAScore_ASCII_Short` = 2 allocs/op, `BenchmarkNYSIISScore_Match` = 2 allocs/op. These are consistent with 1 alloc per Code call. The spec says 0 allocs but achieves 1.
- **Suggested fix:** Amend §14.1 to "Soundex, NYSIIS, MRA: < 500 ns per call, ≤ 1 allocation per Code call (output string; intermediate buffers are stack-allocated)."

---

### [Important] MRA Compare/Score: 2 allocs — budget says 0 (from §14.1 phonetic 0-alloc claim)
- **File:** `/Users/johnny/Development/fuzzymatch/mra.go`
- **Phase introduced:** Phase 7
- **Issue:** `BenchmarkMRACompare_ASCII_Short` = 2 allocs/op (16 B/op); `BenchmarkMRAScore_ASCII_Short` = 2 allocs/op (16 B/op); `BenchmarkMRAScore_Match` = 2 allocs/op; `BenchmarkMRAScore_NoMatch` = 2 allocs/op. `MRACompare` calls `MRACode(a)` and `MRACode(b)`, each of which allocates 1 string. There is no alloc-free path for MRACompare since both Code results must be held simultaneously for comparison.
- **Standard:** `docs/requirements.md` §14.1 "Soundex, NYSIIS, MRA: < 500 ns per call, 0 allocations"
- **Action:** Same as above — amend §14.1 to match the structural reality.
- **Rationale:** MRACompare necessarily calls MRACode twice. If MRACode costs 1 alloc each, MRACompare minimum is 2 allocs. Budget needs updating.
- **Suggested fix:** Include in the §14.1 amendment: "MRACompare: ≤ 2 allocations (two MRACode calls)."

---

### [Important] DamerauLevenshteinOSA Unicode Short: 3 allocs — exceeds documented 0-alloc scope boundary
- **File:** `/Users/johnny/Development/fuzzymatch/damerau_osa.go`
- **Phase introduced:** Phase 2
- **Issue:** `BenchmarkDamerauLevenshteinOSAScore_Unicode_Short` = 3 allocs/op (144 B/op). This is expected (2 `[]rune` allocs + 3 DP rows on the rune path), but bench.txt records 3 allocs when the Unicode path uses `make([]int, n+1)` × 3 (three-row DP). The rune path allocates 2 `[]rune` slices + 3 DP row slices = 5 allocations, but the identity short-circuit and the rune-count-based row sizing reduces the effective count to 3 for short equal inputs. The 0-alloc budget applies only to ASCII; Unicode is documented as not covered.
- **Standard:** performance-standards.md "Per-Call Allocations" (0-alloc budget for ASCII ≤ 50 chars)
- **Action:** No code fix needed. Improvement — add a godoc clarification to `DamerauLevenshteinOSAScoreRunes` stating the 3-alloc expected allocation count for the Unicode path at Short inputs.
- **Rationale:** The bench.txt number is correct and expected. Documenting it explicitly prevents future reviewers from flagging it as a regression.

---

### [Important] LCSStr ASCII Long: 2 allocs — heap path expected but not documented
- **File:** `/Users/johnny/Development/fuzzymatch/lcsstr.go`
- **Phase introduced:** Phase 4
- **Issue:** `BenchmarkLCSStrScore_ASCII_Long` = 2 allocs/op. The file godoc says "Heap path: two make([]int, n+1) calls; 2 allocs on ASCII Long." This is correct and expected. However the performance-standards.md skill states "LCSStr: < 2 µs, 0 allocations" — this applies only to ASCII ≤ 50 chars (Short/Medium). The Long path (>64 chars) correctly falls back to heap allocation and is not covered by the 0-alloc budget.
- **Standard:** performance-standards.md "Per-Call Allocations"; `docs/requirements.md` §14.1 "LCSStr: < 2 µs per call, 0 allocations"
- **Action:** No code fix needed. The bench.txt numbers are correct. The spec §14.1 should clarify that "0 allocations" applies only to inputs ≤ maxStackInputLen (64 chars). Add a note: "For inputs > 64 bytes: 2 allocations (heap path)."
- **Rationale:** The `bench.txt` shows 2 allocs for Long across all count=10 runs, so CI accepts this as baseline. Clarifying the scope of the 0-alloc claim prevents confusion.

---

### [Important] Jaro/JaroWinkler/Strcmp95 ASCII Long: 2–3 allocs — heap path triggers above maxJaroStackLen
- **File:** `/Users/johnny/Development/fuzzymatch/jaro.go`, `/Users/johnny/Development/fuzzymatch/jarowinkler.go`, `/Users/johnny/Development/fuzzymatch/strcmp95.go`
- **Phase introduced:** Phase 2
- **Issue:** `BenchmarkJaroScore_ASCII_Long` = 2 allocs/op (640 B/op); `BenchmarkJaroWinklerScore_ASCII_Long` = 2 allocs/op; `BenchmarkStrcmp95Score_ASCII_Long` = 3 allocs/op (1536 B/op). The `maxJaroStackLen = 256` threshold means inputs > 256 bytes use `make([]bool, la)` and `make([]bool, lb)` (heap). The 500-char "Long" benchmark exceeds this threshold and triggers 2 allocs. Strcmp95 adds a third `make([]bool, lb)` for the `simConsumed` slice. The spec §14.1 budget of "0 allocations" applies only to ASCII ≤ 50 chars (within the stack threshold).
- **Standard:** `docs/requirements.md` §14.1 "Jaro, Jaro-Winkler, Strcmp95: < 1 µs per call, 0 allocations" (applies to ≤ 50 chars)
- **Action:** No code fix needed for v1.0. The 0-alloc budget is only specified for inputs ≤ 50 chars, and all Short/Medium benchmarks correctly show 0 allocs. The Long allocs are expected and should be documented in the respective bench file comments.
- **Rationale:** The bench.txt entries for Long are consistent across count=10 runs, confirming the heap path is stable.

---

### [Important] Levenshtein ASCII Long/Unicode Short: 2 allocs — heap path expected but bench.txt is the only documentation
- **File:** `/Users/johnny/Development/fuzzymatch/levenshtein.go`
- **Phase introduced:** Phase 2
- **Issue:** `BenchmarkLevenshteinScore_ASCII_Long` = 2 allocs/op (8192 B/op); `BenchmarkLevenshteinScore_Unicode_Short` = 2 allocs/op (96 B/op). Both are correct and expected (heap two-row DP for Long; 2×`[]rune` for Unicode Short). No budget violation. However the `levenshtein_bench_test.go` comments for the Long benchmark do not document the expected 2-alloc count.
- **Standard:** performance-standards.md "Benchmark File Structure"
- **Action:** Improvement — add `// Expected: 2 allocs/op (heap path for inputs > 64 bytes)` comment to `BenchmarkLevenshteinScore_ASCII_Long` and `BenchmarkLevenshteinScore_Unicode_Short`.
- **Rationale:** Without explicit documentation of the expected alloc count in the benchmark, future reviewers cannot distinguish "expected 2 allocs" from "regression from 0 to 2 allocs."

---

### [Important] Ratcliff-Obershelp Short: 4 allocs — no alloc budget in spec, but above character-algorithm norm
- **File:** `/Users/johnny/Development/fuzzymatch/ratcliff_obershelp.go`
- **Phase introduced:** Phase 4
- **Issue:** `BenchmarkRatcliffObershelpScore_ASCII_Short` = 4 allocs/op (256 B/op). Spec §14.1 gives only "Ratcliff-Obershelp: < 5 µs per call for short inputs" — no allocation budget. However the performance-standards.md skill states that character-based algorithms "should have 0 allocations per call" for ASCII ≤ 50 chars. The 4 allocs at Short come from `roFindLongestMatch` calling `make([]int, lb+1)` × 2 (prev/curr rows) at each level of the recursion. For a short ~10-char input with ~3 levels of recursion, this produces 3×2 = 6 allocs (bench shows 4, implying identity short-circuits prune some branches).
- **Standard:** performance-standards.md "Per-Call Allocations" (character-based algorithms, ASCII ≤ 50 chars)
- **Action:** Discuss-phase needed — Ratcliff-Obershelp is classified as "Gestalt" (not strictly "character-based" in the DP sense) and the spec lacks an explicit alloc budget. The 4 allocs are inherent to the recursive LCS-decomposition unless a stack-allocated buffer pool is added for short inputs. Options: (A) add an explicit allocation budget to §14.1 for Ratcliff-Obershelp (≤ 4 allocs for short inputs); (B) implement a stack-allocated `[maxStackInputLen+1]int` × 2 buffer and pass it down the recursion.
- **Rationale:** The bench.txt records 4 allocs/op at Short consistently. Without a spec budget, CI cannot detect regressions to, say, 8 allocs/op.
- **Suggested fix:** Add `Ratcliff-Obershelp: < 5 µs per call, ≤ 4 allocations for short inputs (recursive LCS decomposition allocates 2 DP rows per recursion level)` to §14.1.

---

### [Important] Ratcliff-Obershelp Long: 200 allocs, 433 KB — no documentation of growth scaling
- **File:** `/Users/johnny/Development/fuzzymatch/ratcliff_obershelp.go`
- **Phase introduced:** Phase 4
- **Issue:** `BenchmarkRatcliffObershelpScore_ASCII_Long` (500-char inputs) = 200 allocs/op (433857 B/op) in bench.txt. This is the O(N² · M) worst-case allocation explosion from the recursive LCS decomposition: each level of recursion allocates 2 DP rows of size `lb`, and recursion depth is O(min(la, lb)) ≈ 250 levels for equal-length strings. The 200-alloc and ~434 KB/op numbers are expected given the algorithm's complexity but are not documented anywhere in the codebase as the expected Long baseline.
- **Standard:** performance-standards.md "Benchmark Coverage" (all benchmarks use `b.ReportAllocs()`)
- **Action:** Improvement — add a comment to `BenchmarkRatcliffObershelpScore_ASCII_Long` stating "INFORMATIONAL: O(N²·M) recursion allocates proportionally to input length; ~200 allocs/op at 500 chars is expected." Also add a `Ratcliff-Obershelp: DoS notice` section to the ratcliff_obershelp.go file header alongside Monge-Elkan and Partial Ratio.
- **Rationale:** A consumer calling `RatcliffObershelpScore` on untrusted long inputs faces unbounded memory allocation. The DoS notice is required by the "Worst-case complexity documentation for DoS-prone algorithms" in the review scope.

---

### [Important] Missing `BenchmarkAlgoID_String` in *_bench_test.go — only in algoid_test.go
- **File:** `/Users/johnny/Development/fuzzymatch/algoid_test.go`
- **Phase introduced:** Phase 1
- **Issue:** `BenchmarkAlgoID_String` is defined in `algoid_test.go` (a non-bench test file) rather than in a dedicated `algoid_bench_test.go` file. The performance-standards.md skill states "Every algorithm has `<algorithm>_bench_test.go`." The AlgoID String method is not an algorithm per se, but it is performance-sensitive (dispatch table lookup). Having the benchmark in the regular test file conflates unit and performance concerns and prevents per-file performance tooling.
- **Standard:** performance-standards.md "Benchmark File Structure"
- **Action:** Improvement — move `BenchmarkAlgoID_String` into a new `algoid_bench_test.go` file. Keep the test in `algoid_test.go` as is.
- **Rationale:** Minor structural issue. Does not affect CI or benchstat.

---

## Improvement Findings

### [Improvement] DamerauLevenshteinFull: O(m·n) space vs O(n) possible via auxiliary anchor compression
- **File:** `/Users/johnny/Development/fuzzymatch/damerau_full.go`
- **Phase introduced:** Phase 2
- **Issue:** The current Lowrance-Wagner implementation allocates a full `(m+2)×(n+2)` table. The file godoc documents a "two-row + auxiliary-anchor-table" optimisation as a v1.x follow-up. While the transposition term does reference arbitrary prior rows, the number of distinct anchor rows is bounded by the alphabet size (256 for ASCII), meaning a sparse representation storing only the last row per character (rather than all m rows) can reduce space to O(n + |alphabet|). This is the standard optimisation for the Lowrance-Wagner algorithm in production string-matching libraries.
- **Standard:** performance-standards.md "Two-Row DP Optimisation"; `docs/requirements.md` §14.1
- **Action:** Improvement — implement the O(n + 256) space optimisation as a v1.x task. Track in a GitHub issue.
- **Suggested fix:** The `da[256]int` array (already stack-allocated) combined with a single-row rolling DP that re-builds the transposition term from `da` alone can reduce the 21760 B/op Medium allocation to ~512 B/op.

---

### [Improvement] Normalise `normaliseASCII`: two-pass (camel-split + separator-strip) could be one pass
- **File:** `/Users/johnny/Development/fuzzymatch/normalise.go`
- **Phase introduced:** Phase 1
- **Issue:** `normaliseASCII` does a first pass to build `buf` (with camel-split spaces inserted), then a second pass via `collapseSeparators` (separator-strip + space-collapse). Both operate on bytes. A single pass with a small state machine (tracking `prev`, `inSep`, `pendingSpace`) could eliminate the second-pass scan.
- **Standard:** performance-standards.md "When Optimisation is NOT Worth It"
- **Action:** Improvement — benchmark the single-pass variant before implementing to confirm the win (the second pass is O(n) over a buffer already in L1 cache, so the gain may be < 10 ns). File as a v1.x micro-optimisation issue.
- **Rationale:** This would save the second-pass allocation-free scan. The current bench shows 60–70 ns for Short, well within the < 200 ns budget. Low priority.

---

### [Improvement] Q-gram `extractQGrams`: map capacity could use `len(s)` instead of `len(s)-n+1`
- **File:** `/Users/johnny/Development/fuzzymatch/q_gram.go`
- **Phase introduced:** Phase 5
- **Issue:** The capacity hint `len(s)-n+1` is the exact number of windows but underestimates the distinct key count for inputs with repeated patterns (where many windows map to the same key). When the distinct key count < window count (e.g. "aaa...a"), the map stays small; but when key count ≈ window count, the initial capacity is exact and rehashing is avoided. The 6-alloc result at Medium suggests the initial capacity is sometimes underestimated, triggering 2 rehash allocs.
- **Standard:** performance-standards.md "Per-Algorithm Budgets"
- **Action:** Improvement — add a 25% headroom: `make(map[string]int, (len(s)-n+1)*5/4)`. Benchmark to confirm alloc count drops from 6 to 4 at Medium.
- **Suggested fix:** Change `q_gram.go:111` to `m := make(map[string]int, (len(s)-n+1)*5/4)`.

---

### [Improvement] `SmithWatermanGotohScore` Medium wall time at 50 chars is 8.9 µs vs stated "< 5 µs" budget interpretation
- **File:** `/Users/johnny/Development/fuzzymatch/swg.go`
- **Phase introduced:** Phase 3
- **Issue:** `BenchmarkSmithWatermanGotohScore_ASCII_Medium` (50-char inputs) = 8914 ns/op in bench.txt. Spec §14.1 states "Smith-Waterman-Gotoh: < 5 µs per call, 0 allocations (stack buffer for ≤ 50 char inputs)." If the budget is interpreted as "for inputs at exactly 50 chars" then it is breached (8.9 µs vs 5 µs). If interpreted as "for short inputs up to ~20 chars" then the Short benchmark (196 ns) passes comfortably.
- **Standard:** `docs/requirements.md` §14.1 "Smith-Waterman-Gotoh: < 5 µs per call"
- **Action:** Discuss-phase needed — the budget language "≤ 50 char inputs" conflicts with the observed 8.9 µs at exactly 50 chars. SWG has O(mn) complexity; at 50×50 chars, even 6 rolling float64 rows (the Gotoh formulation) requires 3000 float64 operations. Clarify §14.1 to "Smith-Waterman-Gotoh: < 5 µs per call for inputs ≤ 20 chars (short identifier comparison); ≤ 10 µs for 50-char inputs."
- **Rationale:** The current 8.9 µs at 50 chars is a wall-time concern, not an allocation concern (0 allocs at Medium is correct). The < 5 µs budget is likely written for the typical "user_id vs userId" use case (~10 chars) rather than the full 50-char boundary.

---

### [Improvement] `Tokenise` bench file missing explicit expected-alloc comments
- **File:** `/Users/johnny/Development/fuzzymatch/tokenise_bench_test.go`
- **Phase introduced:** Phase 6
- **Issue:** The tokenise bench file does not document the expected alloc counts in benchmark comments. `BenchmarkTokenise_ASCII_Short` = 4 allocs/op and `BenchmarkTokenise_DefaultOptions` = 4 allocs/op; these exceed the §14.3 budget of ≤ 2 allocs and there is no comment explaining why.
- **Standard:** performance-standards.md "Benchmark File Structure"
- **Action:** Improvement — add expected-alloc comments to each tokenise benchmark, and cross-reference the §14.3 budget discussion (or its amended form once resolved).
- **Rationale:** Without comments, reviewers cannot distinguish "known budget miss" from "regression."

---

### [Improvement] No `BenchmarkXxx_ASCII_Long` for phonetic Score functions (Soundex, NYSIIS, MRA)
- **File:** `/Users/johnny/Development/fuzzymatch/soundex_bench_test.go`, `/Users/johnny/Development/fuzzymatch/nysiis_bench_test.go`, `/Users/johnny/Development/fuzzymatch/mra_bench_test.go`
- **Phase introduced:** Phase 7
- **Issue:** Phonetic `Code` benchmarks have Short/Medium/Long variants. But `BenchmarkSoundexScore_*` only has `ASCII_Short` and `ASCII_Identity`. `BenchmarkNYSIISScore_*` only has `Match` and `NoMatch`. `BenchmarkMRAScore_*` only has `ASCII_Short`, `Match`, and `NoMatch`. There are no `ASCII_Medium` or `ASCII_Long` Score benchmarks for these three algorithms. The bench structure requirement says "All benchmarks cover Short, Medium, Long."
- **Standard:** performance-standards.md "Benchmark File Structure" (requires `BenchmarkXxxScore_ASCII_Short`, `_ASCII_Medium`, `_ASCII_Long`)
- **Action:** Improvement — add `BenchmarkSoundexScore_ASCII_Medium`, `BenchmarkNYSIISScore_ASCII_Medium`, `BenchmarkMRAScore_ASCII_Medium` score benchmarks to the respective bench files. These are informational (the algorithms have a bounded-length output) but ensure benchstat tracks score cost at different input lengths.
- **Rationale:** The Code functions have Medium/Long benchmarks, but the Score wrappers do not. Score includes two Code calls, so a Medium Score would show 2 allocs (2 Code calls) at all input lengths — confirming the budget.

---

### [Improvement] `DoubleMetaphoneScore_Identity` benchmark name is non-standard
- **File:** `/Users/johnny/Development/fuzzymatch/double_metaphone_bench_test.go`
- **Phase introduced:** Phase 7
- **Issue:** `BenchmarkDoubleMetaphoneScore_Identity` exists but there is no corresponding `BenchmarkSoundexScore_ASCII_Medium`, `BenchmarkNYSIISScore_ASCII_Medium` etc. The `_Identity` name doesn't follow the `_ASCII_Short/_ASCII_Medium/_ASCII_Long/_Unicode_Short` naming convention in performance-standards.md. This causes benchstat comparisons to fail if the benchmark is renamed.
- **Standard:** performance-standards.md "Benchmark File Structure"
- **Action:** Improvement — rename `BenchmarkDoubleMetaphoneScore_Identity` to `BenchmarkDoubleMetaphoneScore_ASCII_Identity` for naming consistency, or add `_ASCII_Short` as the primary "identity case" benchmark.
- **Rationale:** Minor naming inconsistency. Does not affect benchstat unless the rename triggers a benchstat "new benchmark" detection.

---

## Benchstat Regression Summary

The following benchmarks in `bench.txt` show values that breach or approach their §14.1 budgets:

| Benchmark | bench.txt allocs | §14.1 budget | Status |
|---|---|---|---|
| `DamerauLevenshteinFullScore_ASCII_Short` | 1 | 0 | BREACH |
| `DamerauLevenshteinFullScore_ASCII_Medium` | 1 | 0 | BREACH |
| `CosineScore_ASCII_Medium` | 7 | ≤ 4 | BREACH |
| `QGramJaccardScore_ASCII_Medium` | 6 | ≤ 4 | BREACH |
| `SorensenDiceScore_ASCII_Medium` | 6 | ≤ 4 | BREACH |
| `TverskyScore_ASCII_Medium` | 6 | ≤ 4 | BREACH |
| `TokenSortRatioScore_ASCII_Short` | 14 | ≤ 4 | BREACH |
| `TokenSetRatioScore_ASCII_Short` | 9 | ≤ 4 | BREACH |
| `TokenJaccardScore_ASCII_Short` | 2 | ≤ 4 | PASS |
| `DoubleMetaphoneScore_ASCII_Short` | 6 | ≤ 2 | BREACH |
| `DoubleMetaphoneKeys_ASCII_Short` | 3 | ≤ 2 | BREACH |
| `SoundexCode_ASCII_Short` | 1 | 0 | BREACH |
| `NYSIISCode_ASCII_Short` | 1 | 0 | BREACH |
| `MRACode_ASCII_Short` | 1 | 0 | BREACH |
| `Normalise_ASCII_Short` | 1 | 0 | BREACH |
| `Tokenise_ASCII_Short` | 4 | ≤ 2 | BREACH |

No benchstat regressions have been detected relative to `bench.txt` (the bench.txt IS the baseline). All numbers above are the established bench.txt baseline. Regressions from this baseline would be caught by CI.

## GO/NO-GO

**NO-GO for v1.0.0 tagging** on the following conditions:

1. `BenchmarkDefaultScorer_*` entries are absent from `bench.txt` — CI regression detection is blind for the Scorer layer.
2. `docs/requirements.md` §14.2 Scorer budget (≤ 8 allocs) is mathematically inconsistent with the summed §14.1 per-algorithm budgets and must be formally revised before the §20 acceptance criteria can be declared met.
3. `docs/requirements.md` §14.1 states "0 allocations" for DamerauLevenshteinFull but the implementation allocates 1/op for all input sizes including ASCII Short — the spec and implementation are in documented disagreement with no formal waiver.
4. `docs/requirements.md` §14.1 states "0 allocations" for Soundex/NYSIIS/MRA but all three allocate 1/op per Code call — same category of spec/implementation mismatch.
5. `docs/requirements.md` §14.3 states "0 allocations" for Normalise but 1/op is structurally unavoidable — requires spec amendment.

All five are addressable primarily via spec amendments plus the DoubleMetaphone [4]byte fix. No fundamental architectural changes are required to achieve v1.0 readiness.

</details>

---

## 4. api-ergonomics-reviewer

_Source: `.planning/reviews/api-ergonomics-FINDINGS.md`_

<details>
<summary>Click to expand full report</summary>

---
status: approved_with_changes
agent: api-ergonomics-reviewer
scope: entire public API (phases 1-8)
reviewed: 2026-05-17T00:00:00Z
finding_counts:
  critical: 6
  important: 18
  improvement: 21
  total: 45
---

# fuzzymatch — full-surface API ergonomics review (phases 1–8)

## Verdict

`approved_with_changes`. The 23 algorithm functions, the four foundation primitives (AlgoID, Normalise, Tokenise, errors), and the Phase 8 Scorer make a coherent, idiomatic-Go surface. The progressive-disclosure model holds: Layer 1 hello-world is 4 lines, Layer 2a is 5 lines, Layer 2b is 6 lines. The naming convention `XxxScore` / `XxxDistance` / `XxxCode` is applied consistently across the 23 algorithms. The `AlgoID` enum gives IDE-discoverability and compile-time safety.

Six items are Critical (must change before v1.0). Eighteen are Important (resolve before v1.0 freeze). Twenty-one are Improvement nits.

The **SPEC OVERRIDE** of `ScoreAll` returning `map[AlgoID]float64` (vs `map[string]float64` in `docs/requirements.md` §8.3) is **affirmed**: it is the right call. AlgoID is the typed key the rest of the library exposes; consumers wanting string display call `AlgoID.String()`. This override is the documented decision in the existing 08-API-ERGONOMICS-REVIEW.md and remains the recommendation here.

Phase 8 BLOCKING findings from 08-API-ERGONOMICS-REVIEW.md (`WithThreshold` accepting NaN; `WithTverskyAlgorithm` accepting α=β=0) are inherited into this review as Critical API-T-01 and API-T-02 — both are still open in the codebase as of `scorer_options.go:257-266` and `scorer_options.go:381-399`.

---

## Critical

### [Critical] WithThreshold accepts NaN (silent never-match Scorer)
- **File:** `/Users/johnny/Development/fuzzymatch/scorer_options.go:257-266`
- **Phase introduced:** Phase 8
- **Issue:** `if t < 0.0 || t > 1.0` does not reject `math.NaN()` — NaN passes both inequalities. A Scorer constructed with `WithThreshold(math.NaN())` never matches anything because `... >= NaN` is always false. The Scorer's documented contract (`docs/scorer.md`) and the godoc on `ErrInvalidThreshold` both state NaN is rejected; the implementation does not match. This is the original BLOCKING API-01 from 08-API-ERGONOMICS-REVIEW.md and remains open.
- **Standard:** `go-coding-standards/SKILL.md` "fail loudly at construction"; Principle 1 (Pit of Success).
- **Action:** Code fix — add `math.IsNaN(t)` guard before the range check; return `ErrInvalidThreshold` with the same sentinel.
- **Rationale:** A silent-no-match Scorer is the worst-class API failure mode — caller code does the wrong thing forever with no diagnostic. Must reject at the option-application boundary.

### [Critical] WithTverskyAlgorithm accepts α=β=0 (panic at Score time, not construction time)
- **File:** `/Users/johnny/Development/fuzzymatch/scorer_options.go:381-399`
- **Phase introduced:** Phase 8
- **Issue:** `if alpha < 0 || beta < 0` rejects negatives but accepts `(0, 0)`. `TverskyScore` then panics at `Score` time (the algorithm guards α+β > 0 by panic). Every other `With*Algorithm` validates fully at option-application time. The deferral is inconsistent with the surface; the godoc admission that "typical use is satisfied" is not a defence. Original BLOCKING API-02 from 08-API-ERGONOMICS-REVIEW.md; still open.
- **Standard:** `go-coding-standards/SKILL.md` "validation at the option boundary"; Principle 4 (Idiomatic Go: explicit error at config boundary).
- **Action:** Code fix — add `alpha+beta == 0` check returning `ErrInvalidTverskyParam`; also handle NaN α/β here for consistency with API-T-01.
- **Rationale:** Construction-time errors are typed and discoverable; Score-time panics are not. The Scorer contract says "validated at construction"; this option breaks it.

### [Critical] Direct algorithm calls panic on bad parameters instead of returning errors
- **File:** `/Users/johnny/Development/fuzzymatch/tversky.go:230` (`TverskyScore`), `/Users/johnny/Development/fuzzymatch/qgram_jaccard.go:144` (`QGramJaccardScore`), `/Users/johnny/Development/fuzzymatch/cosine.go:195` (`CosineScore`), `/Users/johnny/Development/fuzzymatch/sorensen_dice.go:158` (`SorensenDiceScore`), and the four corresponding `*Runes` variants.
- **Phase introduced:** Phase 5
- **Issue:** Direct calls to parameterised q-gram algorithms with `n < 1` (or α+β=0 for Tversky) **panic** rather than return an error or score. CONTEXT.md §5 LOCKED this as "fail loudly on programmer error" but the project's other algorithms with parameter validation (SWG, Monge-Elkan) handle bad input differently: SWG produces "deterministic-but-meaningless" output, Monge-Elkan panics only for the inner-AlgoID allow-list. The panic-for-`n<1` policy creates an asymmetry across the algorithm surface: programmer error in `LevenshteinScore("foo", "bar")` never panics, but `QGramJaccardScore("foo", "bar", 0)` does.
- **Standard:** `go-coding-standards/SKILL.md` — algorithm score functions should never panic on caller-supplied data; only on programmer-clearly-wrong invariant breaks.
- **Action:** Discuss-phase needed — either (a) clamp `n` to `max(1, n)` and document, or (b) return 0.0 (the catalogue convention for "no comparison possible"), or (c) document the panic as a hard contract and add a Vet-friendly note in the godoc. The current panic-without-test-fixture (no consumer can write `defer recover()` cleanly without an exported sentinel for `errors.Is(panicValue, ErrInvalidQGramSize)`) is the worst combination.
- **Rationale:** Library-wide consistency. The Scorer layer correctly returns the typed error; the direct-call layer should match or, at minimum, expose a way for `recover()` to discriminate the panic value.

### [Critical] Scorer.Score does not validate input UTF-8 — SWG with NaN params and pathological inputs can return NaN
- **File:** `/Users/johnny/Development/fuzzymatch/scorer.go:349` (`Score`), `/Users/johnny/Development/fuzzymatch/swg.go:174` (`SmithWatermanGotohScoreWithParams`)
- **Phase introduced:** Phases 3 + 8
- **Issue:** A Scorer composed with `WithSmithWatermanGotohAlgorithm(weight, params)` where `params.Match = math.NaN()` (or Inf) accepts the option (no validation per `scorer_options.go:465-479`) and propagates NaN into the composite Score. The composite then violates the documented `[0.0, 1.0]` range guarantee silently. Documented as "nonsense values produce a deterministic-but-meaningless score" but NaN is not deterministic and not meaningless — it's a poisoned float that defeats `Match`'s `>=` comparison.
- **Standard:** `determinism-standards/SKILL.md` NaN/Inf rule (`§5 LOCKED`).
- **Action:** Code fix — `WithSmithWatermanGotohAlgorithm` should reject `math.IsNaN(...)` or `math.IsInf(...)` on any of `params.Match/Mismatch/GapOpen/GapExtend` and return `ErrInvalidConfiguration` (already exists). Alternatively, the SWG kernel should detect NaN inputs and short-circuit to `0.0`, but option-time rejection is the cleaner contract.
- **Rationale:** The Scorer's `[0.0, 1.0]` range guarantee is load-bearing — `Match` semantics depend on it. NaN admission breaks the contract silently.

### [Critical] WriteGoldenFile is publicly exported but is a test-only API
- **File:** `/Users/johnny/Development/fuzzymatch/golden_canonical.go:88`
- **Phase introduced:** Phase 1 (foundation)
- **Issue:** `func WriteGoldenFile(path string, v any) error` is exported in the production-binary code path, not in `_test.go` files or under a build tag. Consumers see it on pkg.go.dev, but its godoc says "intended for test maintenance only — production code never invokes it." This is the classic "exported because internal tests need it" trap.
- **Standard:** `go-coding-standards/SKILL.md` — public API surface is minimal; test-only helpers live in `_test.go` files or behind build tags.
- **Action:** Code fix — move `WriteGoldenFile` into `golden_canonical_test.go` (or add a `//go:build testfixtures` tag), or rename to `_writeGoldenFile` and provide an `export_test.go` re-export `var WriteGoldenFileForTest = writeGoldenFile`.
- **Rationale:** Every public symbol is a v1.x contract. A test-maintenance helper in the public surface is a permanent compatibility liability — any future Go version that breaks `encoding/json.MarshalIndent` stability or any future canonical-form change becomes a v2.0 breaking change.

### [Critical] DefaultScorer panics on internal inconsistency rather than returning a typed value
- **File:** `/Users/johnny/Development/fuzzymatch/scorer.go:586-592`
- **Phase introduced:** Phase 8
- **Issue:** `DefaultScorer` calls `panic("fuzzymatch: DefaultScorer construction failed (this is a bug): " + err.Error())`. The intent is to make the unreachable branch loud, but this means a future refactor that accidentally introduces a dispatch-table gap, a missing AlgoID, or a `WithThreshold` regression panics at consumer import time (the example uses `var defaultScorer = fuzzymatch.DefaultScorer()` in `examples/identifier-similarity/main.go:60` and `examples/scorer-composition/main.go:68`). A consumer who does the same `var x = DefaultScorer()` gets a startup panic with no `recover()` opportunity.
- **Standard:** `go-coding-standards/SKILL.md` — library code panics only on programmer errors with no possible runtime cause.
- **Action:** Code fix — either (a) replace panic with a build-time assertion (a `var _ = func() bool { if _, err := NewScorer(DefaultScorerOptions()...); err != nil { panic(...) }; return true }()` — this fires at package load, not at `DefaultScorer()` call time, where a static-analysis tool could catch the regression), or (b) accept the panic but ensure a regression test asserts `DefaultScorer()` never panics in CI under all build configurations. The current state is a latent failure: a code review that removes one dispatch entry breaks `DefaultScorer` without any test catching it before release.
- **Rationale:** A no-fail constructor that panics is a contradiction. Either move the panic to package-init (so the regression cannot ship) or remove it (with full test coverage that the static composition is valid).

---

## Important

### [Important] Asymmetric Monge-Elkan exposed publicly produces direction-sensitive scores by default
- **File:** `/Users/johnny/Development/fuzzymatch/monge_elkan.go:377`
- **Phase introduced:** Phase 6
- **Issue:** `MongeElkanScore(a, b, inner, opts)` is the **asymmetric** variant — `MongeElkanScore(a, b) != MongeElkanScore(b, a)` in general. The library's standard convention is symmetric similarity (every other algorithm in the catalogue satisfies `Score(a,b) == Score(b,a)`). Consumers used to symmetric algorithms will write `MongeElkanScore(query, candidate, ...)` and get a different score than `MongeElkanScore(candidate, query, ...)` — a subtle and dangerous footgun. The Scorer dispatch wraps `MongeElkanScoreSymmetric`, so the Scorer path is safe, but the direct surface is not.
- **Standard:** `algorithm-correctness-standards/SKILL.md` — symmetric property is the catalogue's default invariant.
- **Action:** Code fix — rename the current asymmetric surface to `MongeElkanScoreAsymmetric` and expose the symmetric form as `MongeElkanScore`. Alternatively, deprecate `MongeElkanScore` and document `MongeElkanScoreSymmetric` as the canonical entry point. The current naming inverts the principle of least surprise.
- **Suggested fix:** `MongeElkanScore` → symmetric; `MongeElkanScoreAsymmetric` (or `MongeElkanScoreDirectional`) → current asymmetric.

### [Important] MRACompare returns `(matched bool, simScore int)` — non-idiomatic Go return tuple
- **File:** `/Users/johnny/Development/fuzzymatch/mra.go:241`
- **Phase introduced:** Phase 7
- **Issue:** `MRACompare(a, b) (matched bool, simScore int)` — a `(bool, int)` tuple with named results is awkwardly close to the `(value, ok)` idiom but inverts the order (`matched` first, not last). Most Go consumers expect `comma-ok` form with the data value first and the boolean status last. Worse, the function's primary purpose (compute MRA score) is buried as the second return; the boolean is a derived predicate that consumers can compute themselves via `simScore >= mraThreshold(len(a)+len(b))`.
- **Standard:** `go-coding-standards/SKILL.md` — return tuples follow `(value, ok)` or `(value, error)` ordering.
- **Action:** Code fix — either (a) change to `MRACompare(a, b) (simScore int, matched bool)` to match `(value, status)` convention; or (b) keep the current signature and accept the non-idiomatic order; or (c) provide a separate `MRASimilarity(a, b) int` returning just the integer counter and remove the bool from `MRACompare`.
- **Suggested fix:** Renaming to `MRASimilarity(a, b) int` for the integer counter; keep `MRACompare(a, b) bool` for the boolean. Two functions, single-purpose each, idiomatic.

### [Important] Strcmp95Score has no rune variant — silently bytewise on Unicode input
- **File:** `/Users/johnny/Development/fuzzymatch/strcmp95.go:260`
- **Phase introduced:** Phase 4
- **Issue:** The catalogue convention is `XxxScore` (byte) + `XxxScoreRunes` (rune-aware) for every character-based algorithm. `Strcmp95Score` ships only the byte variant; the file header documents "ASCII-only — no `*Runes` variant" with the rationale that the similar-character table is upper-case ASCII letters only. This is defensible but inconsistent with the catalogue's surface promise — a consumer who has memorised `LevenshteinScore` / `LevenshteinScoreRunes` writes `Strcmp95ScoreRunes` and hits a compile error with no guidance toward the byte variant.
- **Standard:** `go-coding-standards/SKILL.md` — surface uniformity across algorithm categories.
- **Action:** Discuss-phase needed — either (a) add `Strcmp95ScoreRunes` that delegates to `Strcmp95Score` after `[]rune` conversion (with a doc note that the similar-character credit pass operates byte-wise so the rune surface is "rune-counted similarity, byte-compared content"), or (b) keep the gap and add a `Strcmp95ScoreRunes` placeholder that fails-compile with a useful error message via build-tag tricks (too clever — reject), or (c) update llms.txt and the README with an explicit note that Strcmp95 is byte-only by design.
- **Rationale:** Discoverability. A consumer iterating through the catalogue should not hit a single-algorithm exception without finding the rationale on pkg.go.dev within 2 clicks.

### [Important] SoundexScore / DoubleMetaphoneScore / NYSIISScore / MRAScore have no Runes variant
- **File:** `/Users/johnny/Development/fuzzymatch/soundex.go:257`, `/Users/johnny/Development/fuzzymatch/double_metaphone.go`, `/Users/johnny/Development/fuzzymatch/nysiis.go:350`, `/Users/johnny/Development/fuzzymatch/mra.go:348`
- **Phase introduced:** Phase 7
- **Issue:** Same as Strcmp95 — phonetic algorithms are ASCII-only by their primary-source definition (Russell 1918, Knuth 1973 for Soundex; Philips 2000 for Double Metaphone; Taft 1970 for NYSIIS; NBS 1977 for MRA). The non-ASCII input handling is documented per-algorithm ("non-ASCII runes are dropped silently"), but the API surface lacks a `*Runes` variant — same uniformity gap as Strcmp95.
- **Standard:** `go-coding-standards/SKILL.md` surface uniformity.
- **Action:** Discuss-phase needed — these algorithms are genuinely ASCII-only at the algorithm level (Soundex's "BFPV → 1" table is letter-pattern-specific). Recommend: keep no Runes variant for phonetic and document the absence in a single shared paragraph at the top of `llms.txt` and `docs/algorithms.md#phonetic-tier`.

### [Important] TokenSortRatioScore / TokenSetRatioScore / TokenJaccardScore have no Runes variant
- **File:** `/Users/johnny/Development/fuzzymatch/token_sort_ratio.go:183`, `/Users/johnny/Development/fuzzymatch/token_set_ratio.go:281`, `/Users/johnny/Development/fuzzymatch/token_jaccard.go:200`
- **Phase introduced:** Phase 6
- **Issue:** These algorithms tokenise via `Tokenise` (which is rune-aware) and then operate on the tokenised strings byte-wise. The rationale documented in `token_sort_ratio.go:181-182` is "Tokenise is UTF-8-aware so the rune semantic is already preserved." Defensible — but `PartialRatioScore` provides a `*Runes` variant despite operating on the same tokenised representation, which creates an internal inconsistency. Consumer reading the catalogue sees `PartialRatioScoreRunes` exists and expects `TokenSortRatioScoreRunes` to exist too.
- **Standard:** `go-coding-standards/SKILL.md` surface uniformity.
- **Action:** Discuss-phase needed — either add `*Runes` variants to all three (delegating after `[]rune` conversion), or remove `PartialRatioScoreRunes` for symmetry, or document the asymmetry prominently. Recommended: remove `PartialRatioScoreRunes`'s asymmetry and instead document why all token-tier algorithms are byte-clean once tokenised.

### [Important] MongeElkanScore takes `opts NormalisationOptions` parameter that is intentionally unused
- **File:** `/Users/johnny/Development/fuzzymatch/monge_elkan.go:377` and `:467`
- **Phase introduced:** Phase 6
- **Issue:** `MongeElkanScore(a, b string, inner AlgoID, opts NormalisationOptions) float64` accepts `opts` and immediately discards it via `_ = opts`. The godoc says "accepted for forward-compatibility with the Phase 8 Scorer option WithMongeElkanAlgorithm." This is a four-parameter signature where one parameter is documented as inert — a guaranteed footgun for any consumer who passes `DefaultNormalisationOptions()` expecting it to actually normalise.
- **Standard:** `go-coding-standards/SKILL.md` "no inert parameters in public API"; Principle 2 (Make Simple Things Simple).
- **Action:** Code fix — remove `opts` from the signature. The Phase 8 Scorer's `WithMongeElkanAlgorithm` can pass `DefaultNormalisationOptions()` internally if forward-compat is genuinely needed, but the public direct-call surface should not carry an inert parameter. If consumers genuinely need to control Tokenise options, expose a `MongeElkanScoreWithOptions(a, b string, inner AlgoID, tok TokeniseOptions) float64` variant instead.
- **Rationale:** Principle 1 (Pit of Success) — the obvious way to use the function must be the right way. Passing inert `opts` is a footgun.

### [Important] AlgoID.String returns CamelCase rather than the requirements-doc-specified snake_case
- **File:** `/Users/johnny/Development/fuzzymatch/algoid.go:213-267`
- **Phase introduced:** Phase 1
- **Issue:** The current `String()` returns `"Levenshtein"`, `"DamerauLevenshteinOSA"`, `"LCSStr"`, `"NYSIIS"`, etc. The reviewer brief (instruction) says: `AlgoID.String() returns snake_case: "levenshtein", "jaro_winkler", "damerau_levenshtein_osa"`. The current implementation diverges from this convention.
- **Standard:** `api-ergonomics-reviewer` Principle 5 (Naming Consistency).
- **Action:** Discuss-phase needed — the project locked CamelCase in CONTEXT.md and updated llms.txt, llms-full.txt, docs/scorer.md to match. Two valid paths: (a) accept the deviation from the brief and document it (the project lead chose CamelCase for v1.x display labels); (b) revert to snake_case to match the brief. Recommended: accept current CamelCase; it matches Go's idiomatic enum-name → string convention (e.g., `time.Sunday.String()` returns `"Sunday"`, not `"sunday"`). Update the brief to reflect the locked decision.
- **Suggested fix:** No code change. Update the agent brief (`.claude/agents/api-ergonomics-reviewer.md`) to record CamelCase as the locked convention.

### [Important] AlgoIDs() allocates a fresh slice on every call but is not documented as a hot-path-unsafe helper
- **File:** `/Users/johnny/Development/fuzzymatch/algoid.go:282-308`
- **Phase introduced:** Phase 1
- **Issue:** `AlgoIDs()` allocates a 23-element slice on every call. The godoc says "freshly allocated on every call so the caller may freely mutate" — correct but a consumer iterating algorithms in a loop (`for _, id := range fuzzymatch.AlgoIDs() { ... }`) allocates per-call. The pattern in `scorer.go:253` (NewScorer internal use) and the example program both call it once and cache; consumers who write the obvious loop pay a hidden allocation tax.
- **Standard:** `performance-standards/SKILL.md` — allocation surprises in obvious-looking code.
- **Action:** Code fix or godoc fix — either (a) add a sibling `AlgoIDsInto(dst []AlgoID) []AlgoID` zero-alloc variant for hot paths; or (b) expose the underlying ordering via `[NumAlgorithms]AlgoID{...}` as an exported value (mutable-by-design — but indexed access is zero-alloc); or (c) update the godoc to recommend single-call caching in hot loops.
- **Suggested fix:** Add to godoc: "For hot-path iteration, cache the result once: `var allAlgos = fuzzymatch.AlgoIDs()`."

### [Important] DispatchEntryNilForTest / DispatchInvokeForTest / DispatchLenForTest exported in production binary
- **File:** `/Users/johnny/Development/fuzzymatch/export_test.go:38-73`
- **Phase introduced:** Phase 1
- **Issue:** `export_test.go` is, despite its name, a build-tag-free `_test.go` file whose symbols **are not** in the production binary — Go's `_test.go` convention makes these test-only. This is correct usage. However, the same file exports `CanonicalMarshalForTest`, `NumAlgorithmsForTest`, `WinklerPrefixScaleForTest`, `Strcmp95SimilarCharsLenForTest`, etc. — fifteen test-only re-exports. These cannot be reached by production consumers, so this is not a public-API leak. False alarm — flagging only for documentation clarity.
- **Standard:** `go-coding-standards/SKILL.md` — `_test.go` files compile only under `go test`.
- **Action:** No action — confirmed correct. Mentioning for review-record completeness.
- **Rationale:** Verified `grep -L '// +build\|//go:build' export_test.go` plus the `_test.go` suffix gates this file from production builds. The 15 re-exports remain test-only.

### [Important] Scorer struct has no public fields but doc/scorer.md shows `ScorerAlgorithm{ID, Weight}` literal in examples
- **File:** `/Users/johnny/Development/fuzzymatch/scorer.go:414` (`ScorerAlgorithm`)
- **Phase introduced:** Phase 8
- **Issue:** `ScorerAlgorithm` is a pure data struct with two exported fields (`ID`, `Weight`). Documented as "consumers must not rely on ScorerAlgorithm having a stable memory address" — but the struct **is** the API contract. A future minor release that adds a `Threshold` field, a `Disabled` bool, or a `Source string` to `ScorerAlgorithm` is a breaking change for any consumer using struct-literal construction (`ScorerAlgorithm{ID: x, Weight: 0.5}`). Recommendation: hide the constructor — provide only the `Algorithms()` reader, never let consumers build a `ScorerAlgorithm` directly.
- **Standard:** `go-coding-standards/SKILL.md` — exported structs with all-exported fields are evolution-unfriendly.
- **Action:** Discuss-phase — either (a) accept the field-evolution liability (the struct is small and stable; unlikely to grow), or (b) keyed struct-literal initialisation is enforced via `//nolint:exhaustruct` configuration, or (c) add an unexported sentinel field so struct-literal init forces consumers through a constructor `NewScorerAlgorithm(id AlgoID, weight float64) ScorerAlgorithm`. Recommendation: (a) accept — the struct is descriptive metadata, evolution is unlikely.

### [Important] ScorerOption is a function type referencing the unexported `*scorerConfig`
- **File:** `/Users/johnny/Development/fuzzymatch/scorer_options.go:65`
- **Phase introduced:** Phase 8
- **Issue:** `type ScorerOption func(*scorerConfig) error` — `ScorerOption` is exported, but its sole parameter `*scorerConfig` is unexported. Consumers cannot write their own `ScorerOption` because the function body cannot reference the unexported type. This is intentional (it locks the option set to the library's With* helpers) but should be documented explicitly: "ScorerOption values are only constructible via fuzzymatch's With* helpers; custom options are not supported."
- **Standard:** `go-coding-standards/SKILL.md` — opaque types should document opacity.
- **Action:** Godoc fix — add a sentence to the `ScorerOption` godoc: "ScorerOption is opaque: consumers compose Scorers via the library's With*Algorithm / WithThreshold / WithNormalisation options. Building a ScorerOption with a custom function body is not supported (the parameter type is unexported)."
- **Rationale:** Discoverability — a consumer reading the godoc should not have to infer opacity from compile errors.

### [Important] Scorer.Threshold returns float64 with no NaN/Inf guarantee post-Critical fix
- **File:** `/Users/johnny/Development/fuzzymatch/scorer.go:441-443`
- **Phase introduced:** Phase 8
- **Issue:** Once API-T-01 is fixed (NaN rejected at `WithThreshold`), `Threshold()` is guaranteed in `[0.0, 1.0]`. The godoc currently does not state this guarantee; consumers cannot rely on it. Documentation gap, not a code bug.
- **Standard:** `documentation-standards/SKILL.md` — invariants on returns are godoc-explicit.
- **Action:** Godoc fix — append "The returned value is in [0.0, 1.0]: NewScorer's validation pipeline rejected any out-of-range or NaN threshold at construction." (the current godoc says the first part; add the NaN note after API-T-01 fix.)
- **Rationale:** Once the contract is bulletproof, the godoc should advertise it.

### [Important] NewSWGParams returns a value type with mutable fields and no validation
- **File:** `/Users/johnny/Development/fuzzymatch/swg.go:127`
- **Phase introduced:** Phase 3
- **Issue:** `NewSWGParams()` returns `SWGParams{Match: 1.0, Mismatch: -1.0, GapOpen: -1.5, GapExtend: -0.5}`. Consumer mutates freely. The godoc warns "callers may pass nonsense values" but the lack of validation in `SmithWatermanGotohScoreWithParams` means NaN/Inf flow through and break determinism (see Critical SWG-NaN finding). Validation could happen in `NewSWGParams` (eager check) or at score time (lazy check) — currently neither.
- **Standard:** `algorithm-correctness-standards/SKILL.md`; `determinism-standards/SKILL.md` NaN/Inf rule.
- **Action:** Code fix (combine with Critical SWG-NaN) — add either an `NewSWGParamsValidated(match, mismatch, gapOpen, gapExtend float64) (SWGParams, error)` constructor that rejects NaN/Inf and sign-mismatched params (Mismatch > 0, GapOpen > 0, GapExtend > 0), or document `NewSWGParams()` as the only valid construction path and reject mutated NaN/Inf at `SmithWatermanGotohScoreWithParams` boundary.
- **Suggested fix:** Reject NaN/Inf at `WithSmithWatermanGotohAlgorithm` (covers the Scorer surface) and in `SmithWatermanGotohScoreWithParams` direct-call (covers the algorithm surface) — both return 0.0 silently with a future-considered upgrade to return an explicit error if the algorithm surface migrates to `(float64, error)`.

### [Important] WithoutAlgorithm silently no-ops on absent AlgoID — no way to verify removal succeeded
- **File:** `/Users/johnny/Development/fuzzymatch/scorer_options.go:184-199`
- **Phase introduced:** Phase 8
- **Issue:** `WithoutAlgorithm(AlgoX)` filters entries; if no match, the slice is unchanged. The godoc states this is by design ("the no-op-on-absent semantic enables composition patterns"). However, a consumer who **expects** removal (e.g., they think `DefaultScorerOptions()` includes `AlgoLevenshtein` but it doesn't) gets silent failure. There is no `MustWithoutAlgorithm` or option-return-with-removed-count helper.
- **Standard:** `go-coding-standards/SKILL.md` — explicit failure is preferred unless silence is documented as desirable.
- **Action:** Godoc fix — strengthen the godoc with an explicit "If you need to verify the algorithm was actually present before removal, call `Scorer.Algorithms()` after construction and compare lengths." Code fix optional: provide `Scorer.HasAlgorithm(AlgoID) bool` for pre-removal verification on a constructed Scorer.
- **Rationale:** Silent no-op is a footgun for typo-ridden custom compositions ("WithoutAlgorithm(AlgoLeveshtein)" — misspelled — silently does nothing).

### [Important] ErrInvalidConfiguration is declared but never returned by any current code path
- **File:** `/Users/johnny/Development/fuzzymatch/errors.go:57`
- **Phase introduced:** Phase 1 (declared); never wired
- **Issue:** Searching the codebase: `ErrInvalidConfiguration` is referenced only in errors.go itself, llms-full.txt, and a couple of tests. No production code returns it. The Scorer surface returns `ErrInvalidWeight`, `ErrInvalidThreshold`, `ErrMissingThreshold`, `ErrEmptyScorer`, `ErrInvalidAlgorithm`, `ErrInvalidQGramSize`, `ErrInvalidTverskyParam` — all specific. `ErrInvalidConfiguration` is the only "umbrella" sentinel and is dead code.
- **Standard:** `go-coding-standards/SKILL.md` — unused public symbols increase API surface without value.
- **Action:** Discuss-phase needed — either (a) remove `ErrInvalidConfiguration` before v1.0 (it's not used; CHANGELOG note "Removed unused sentinel"), or (b) wire it as a parent error (`errors.Join(ErrInvalidConfiguration, ErrInvalidWeight)`) so `errors.Is(err, ErrInvalidConfiguration)` catches every Scorer-config error in one check. Recommended: (b) — the umbrella error is genuinely useful for consumers who want a single catch-all, and the join cost is one allocation at error-construction time.

### [Important] ErrEmptyInput is declared but never returned
- **File:** `/Users/johnny/Development/fuzzymatch/errors.go:106`
- **Phase introduced:** Phase 1 (declared); never wired
- **Issue:** Same shape as `ErrInvalidConfiguration` — the godoc says "Individual algorithm score functions handle empty inputs per their per-algorithm specification… and do NOT return this error — only higher-level APIs (Scorer, Extract) that require non-degenerate input may surface it." Phase 8 Scorer does NOT return it (both-empty returns 1.0 via identity short-circuit). Phase 10 Extract is not yet implemented.
- **Standard:** `go-coding-standards/SKILL.md`.
- **Action:** Discuss-phase needed — remove before v1.0 unless Phase 10 will use it. If Phase 10 plans use it, leave the declaration and note "reserved for Phase 10 Extract surface" in the godoc.

### [Important] ErrInvalidInput is declared but never returned (also documented as rare)
- **File:** `/Users/johnny/Development/fuzzymatch/errors.go:48`
- **Phase introduced:** Phase 1 (declared); never wired
- **Issue:** Same as ErrEmptyInput. Godoc says "Most algorithms accept arbitrary bytes and do NOT return this error; the exceptions document their constraints in their own godoc." No algorithm in the codebase actually returns it.
- **Standard:** `go-coding-standards/SKILL.md`.
- **Action:** Discuss-phase needed — remove before v1.0, or document its forward-compat reservation. Recommendation: remove three unused sentinels (`ErrInvalidConfiguration`, `ErrInvalidInput`, `ErrEmptyInput`) before v1.0; re-add later if needed (additive change is non-breaking).

---

## Improvement

### [Improvement] AlgoIDs() return slice could be a sortable type for self-documenting use
- **File:** `/Users/johnny/Development/fuzzymatch/algoid.go:282`
- **Issue:** `[]AlgoID` is fine but consumers who want to compare two algorithm sets (e.g., for set-equality) write helpers. A `type AlgoIDSet []AlgoID` with `Contains(AlgoID) bool` would self-document.
- **Action:** No action — over-engineering for v1.0; revisit in v1.x.

### [Improvement] LevenshteinScore / LevenshteinDistance return-type inconsistency from text
- **File:** `/Users/johnny/Development/fuzzymatch/levenshtein.go:84` (`LevenshteinDistance int`) and `:155` (`LevenshteinScore float64`)
- **Issue:** `Distance` is `int`, `Score` is `float64`. Both correct; just noting the deliberate split — the brief expects `XxxScore(a, b) float64` and `XxxDistance(a, b) int`. Current matches.
- **Action:** None.

### [Improvement] HammingDistance unequal-length returns max(len) — non-standard Hamming definition
- **File:** `/Users/johnny/Development/fuzzymatch/hamming.go:69`
- **Issue:** Classical Hamming is defined only for equal-length strings (Hamming 1950 §1). Returning `max(len(a), len(b))` is a fuzzymatch convention to make `HammingScore` normalise to 0.0; the godoc explains. Documented; nothing to change but worth noting that academic readers will expect `HammingDistance("abc", "ab")` to return an error or panic, not silently return 3.
- **Action:** Godoc reinforcement — add a citation/aside: "This deviates from Hamming 1950's strict equal-length requirement; the project's locked policy (CONTEXT.md) is silent-max for ergonomic Score normalisation. Consumers needing the strict definition should length-check before calling."

### [Improvement] DoubleMetaphoneKeys returns `(primary, secondary string)` — named returns inconsistent with rest of catalogue
- **File:** `/Users/johnny/Development/fuzzymatch/double_metaphone.go`
- **Issue:** Named returns are uncommon in the rest of the catalogue. Consumers writing `p, s := DoubleMetaphoneKeys("Smith")` get IDE inlay hints showing "primary, secondary" — useful. Not a problem; just inconsistent with `MRACompare(a, b) (matched bool, simScore int)` which also uses named returns. Mixed style.
- **Action:** No action — named returns where the names add information are fine.

### [Improvement] SmithWatermanGotohRawScore returns unbounded float64 — no documented range
- **File:** `/Users/johnny/Development/fuzzymatch/swg.go:244`
- **Issue:** `*RawScore` variants return the unclamped raw alignment value; godoc explains the unbounded nature. Consumers may store this and compare to a threshold, then break when params change. Recommend: a worked example in godoc showing typical raw-score values for English-name pairs.
- **Action:** Godoc improvement.

### [Improvement] Six SWG public functions on the surface is a lot
- **File:** `/Users/johnny/Development/fuzzymatch/swg.go`
- **Issue:** `SmithWatermanGotohScore` / `SmithWatermanGotohScoreRunes` / `SmithWatermanGotohScoreWithParams` / `SmithWatermanGotohRawScore` / `SmithWatermanGotohRawScoreRunes` / `SmithWatermanGotohRawScoreWithParams` — six entry points for one algorithm. The Score/RawScore axis × byte/runes axis × default-params/custom-params axis = 6. The combinatorial growth is fine but the naming uses no shared suffix structure (Score-vs-RawScore is a totally different semantics swap from Runes vs no-Runes).
- **Action:** Documentation polish — add a clear table in `docs/algorithms.md#smith-waterman-gotoh` showing the 2×2×2 → 6 mapping (the 8th combinator `RawScoreWithParamsRunes` is omitted — that's a gap, see next).

### [Improvement] SWG lacks `SmithWatermanGotohRawScoreWithParamsRunes` (the 8th combinator)
- **File:** `/Users/johnny/Development/fuzzymatch/swg.go`
- **Issue:** Following the 2×2×2 logic: byte/rune × clamp/raw × default-params/custom-params = 8 functions. Current surface has 6. The two missing combinators are byte-WithParams-Raw (`SmithWatermanGotohRawScoreWithParams` — present) and rune-WithParams-Raw (`SmithWatermanGotohRawScoreWithParamsRunes` — MISSING).
- **Action:** Code fix — add `SmithWatermanGotohRawScoreWithParamsRunes(a, b string, params SWGParams) float64` for completeness.

### [Improvement] LongestCommonSubstring returns the substring; LCSStrScore returns the score — naming dissonance
- **File:** `/Users/johnny/Development/fuzzymatch/lcsstr.go:127` and `:203`
- **Issue:** `LongestCommonSubstring` is descriptive; `LCSStrScore` is an abbreviation matching the `AlgoLCSStr` enum. The byte/runes axis is split across two name conventions. Consider renaming `LCSStrScore` → `LongestCommonSubstringScore` for consistency.
- **Action:** Discuss-phase — current naming follows the AlgoID, which is right for dispatch but reads odd for direct consumers. Recommendation: keep `LCSStrScore` as the AlgoID-matched name and add a comment in godoc cross-referencing `LongestCommonSubstring` (the substring-extracting helper).

### [Improvement] DefaultScorer composition is opaque — consumer cannot discover it without reading docs/scorer.md
- **File:** `/Users/johnny/Development/fuzzymatch/scorer.go:586`
- **Issue:** `DefaultScorer()` returns six-algorithm composite; the godoc says so but the consumer can't introspect the choice. `Scorer.Algorithms()` returns the post-construction list, so a consumer can `for _, a := range DefaultScorer().Algorithms() { fmt.Println(a.ID) }` — but this is discovery-after-the-fact.
- **Action:** Godoc improvement — add the six-algorithm list as a literal table in the `DefaultScorer` godoc with the AlgoID values inline.

### [Improvement] DefaultScorerOptions includes seven options (six algos + threshold), Scorer godoc says "six algorithms"
- **File:** `/Users/johnny/Development/fuzzymatch/scorer.go:543-553`
- **Issue:** Pedantic clarity — `DefaultScorerOptions()` returns 7 options (6 algorithms + 1 threshold). The Scorer godoc reads "six algorithms" which is correct for algorithms but the count of options is seven. A consumer doing `len(DefaultScorerOptions())` gets 7.
- **Action:** Godoc nit — clarify "six algorithm options plus a threshold option" in the relevant comment block.

### [Improvement] Tversky parameter ordering `(weight, alpha, beta float64, n int)` vs `TverskyScore(a, b string, n int, alpha, beta float64)` is inconsistent
- **File:** `/Users/johnny/Development/fuzzymatch/scorer_options.go:381` and `/Users/johnny/Development/fuzzymatch/tversky.go:230`
- **Issue:** `WithTverskyAlgorithm(weight, alpha, beta float64, n int)` — n is last. `TverskyScore(a, b string, n int, alpha, beta float64)` — n is in the middle. The two call sites have mirror-image parameter orderings; a consumer who learned `(n, alpha, beta)` from `TverskyScore` writes `WithTverskyAlgorithm(weight, n, alpha, beta)` and gets a compile success (all float64) with a swapped semantic — silent miscompose.
- **Action:** Code fix — unify ordering. Recommend the WithTverskyAlgorithm signature: `WithTverskyAlgorithm(weight float64, n int, alpha, beta float64) ScorerOption` (matches TverskyScore's `n int, alpha, beta float64` suffix).
- **Suggested fix:** `WithTverskyAlgorithm(weight float64, n int, alpha, beta float64)`.

### [Improvement] WithCosineAlgorithm / WithSorensenDiceAlgorithm / WithQGramJaccardAlgorithm signature `(weight float64, n int)` is consistent and good
- **File:** `/Users/johnny/Development/fuzzymatch/scorer_options.go:300/325/350`
- **Issue:** Three q-gram options all use `(weight, n)`. Good. Note for completeness.
- **Action:** None.

### [Improvement] WithMongeElkanAlgorithm has `(weight float64, inner AlgoID)` — no opts parameter
- **File:** `/Users/johnny/Development/fuzzymatch/scorer_options.go:425`
- **Issue:** `WithMongeElkanAlgorithm` does NOT take a `NormalisationOptions` parameter, but `MongeElkanScore` direct-call does (the unused `opts` flagged earlier). When the unused-opts issue is fixed, this option will be naturally consistent. Note this here for traceability.
- **Action:** None — resolves with Important MongeElkanScore-opts fix.

### [Improvement] Scorer.Score godoc does not show example input/output values
- **File:** `/Users/johnny/Development/fuzzymatch/scorer.go:349`
- **Issue:** Long godoc, no quick "Score("foo", "foobar") returns 0.X" reference vector. Consumers reading pkg.go.dev want a number to anchor their understanding.
- **Action:** Godoc improvement — add a worked example with concrete float values.

### [Improvement] Scorer.ScoreAll godoc says "fresh map allocated on every call" but does not warn against hot-path use
- **File:** `/Users/johnny/Development/fuzzymatch/scorer.go:497`
- **Issue:** Godoc covers the allocation, but a consumer who writes `for ... { s.ScoreAll(a, b) }` may not realise the per-call map overhead matters until they benchmark. Recommend adding a quick "in hot loops, prefer Score()" note.
- **Action:** Godoc improvement — add "Hot-path callers should use Score; ScoreAll is for introspection and debugging."

### [Improvement] Scorer ScoreAll returns map[AlgoID]float64 — consumer can iterate to get keys but Algorithms() also returns ordered set
- **File:** `/Users/johnny/Development/fuzzymatch/scorer.go:497` and `:460`
- **Issue:** Two paths to the same information: `s.Algorithms()` returns sorted `[]ScorerAlgorithm`; `s.ScoreAll(a, b)` returns `map[AlgoID]float64`. Consumers iterating ScoreAll in deterministic order do `for _, a := range s.Algorithms() { score := result[a.ID] }`. This is the documented pattern but the godoc on `ScoreAll` could link directly to `Algorithms()` as the recommended sorting path.
- **Action:** Godoc improvement — cross-link the two methods.

### [Improvement] Scorer.Match is a pure wrapper; could be a function-typed property
- **File:** `/Users/johnny/Development/fuzzymatch/scorer.go:392`
- **Issue:** `Match` does `Score(a,b) >= threshold`. One-line wrapper. Acceptable as ergonomic sugar but adds a method to the surface for nothing the consumer couldn't write themselves.
- **Action:** Keep — the ergonomics gain (no exposed threshold field) is worth the line.

### [Improvement] examples/identifier-similarity/main.go uses `defer recover()` pattern via panic-recover
- **File:** `/Users/johnny/Development/fuzzymatch/examples/identifier-similarity/main.go`
- **Issue:** No `defer recover()` actually in the example, but a consumer running the example with edited pairs that include `nil` rune patterns might trigger an algorithm panic (e.g., `MongeElkanScore` with `AlgoMongeElkan` inner — see allow-list panic). The example should explicitly note panic-recovery in production code.
- **Action:** Documentation note in the example's godoc preamble: "Production code should `defer recover()` when calling algorithm functions with consumer-supplied AlgoIDs (Monge-Elkan inner) — see allow-list godoc on MongeElkanScore."

### [Improvement] examples/scorer-composition uses `_, err := NewScorer(...); if err != nil { panic(...) }` pattern
- **File:** `/Users/johnny/Development/fuzzymatch/examples/scorer-composition/main.go:89-104`
- **Issue:** The example correctly handles the error then panics as "unreachable." Good. Could be replaced with `errcheck.Must(NewScorer(...))` style if the library exposed a `MustNewScorer(opts ...ScorerOption) *Scorer` helper. Adding `MustNewScorer` would be idiomatic Go (compare `template.Must` in stdlib).
- **Action:** Code fix or skip — adding `MustNewScorer(opts ...) *Scorer` is small, low-risk, idiomatic, and removes ~5 lines of boilerplate from example code. Recommend: add it.

### [Improvement] No package-level constants for "common threshold values" (0.85, 0.80, 0.95)
- **File:** `/Users/johnny/Development/fuzzymatch/scorer.go`
- **Issue:** Consumers writing `WithThreshold(0.85)` repeat the magic number; `DefaultScorer` bakes in 0.85 but does not export it. A `const DefaultThreshold = 0.85` would let consumers reference the constant.
- **Action:** Code fix optional — export `const DefaultScorerThreshold = 0.85` so `WithThreshold(fuzzymatch.DefaultScorerThreshold)` is self-documenting.

### [Improvement] README's Quick start section still shows Phase 1 primitives, not the Phase 8 Scorer flow
- **File:** `/Users/johnny/Development/fuzzymatch/README.md:96-117`
- **Issue:** Phase 8 has landed (Scorer is in production code) but the README still shows `Normalise` + `Tokenise` as the quick-start example. A first-time visitor sees the wrong entry point.
- **Action:** Documentation fix — replace the quick start with a Phase 8-tier example (DefaultScorer or LevenshteinScore one-liner).

### [Improvement] llms.txt entry for Scorer section header is too long and references "plan 08-01/02/03"
- **File:** `/Users/johnny/Development/fuzzymatch/llms.txt:187`
- **Issue:** `### Scorer construction options (Phase 8 — plan 08-01 lays the option layer; plan 08-02 lands NewScorer + Score + Match; plan 08-03 lands ScoreAll + Threshold + Algorithms + ScorerAlgorithm + DefaultScorer + DefaultScorerOptions)` — the heading exposes internal planning phases to AI-assistant consumers. Should be clean: `### Scorer (Phase 8)`.
- **Action:** Documentation fix — trim the heading.


</details>

---

## 5. bdd-scenario-reviewer

_Source: `.planning/reviews/bdd-scenario-FINDINGS.md`_

<details>
<summary>Click to expand full report</summary>

---
status: issues_found
agent: bdd-scenario-reviewer
scope: entire BDD suite (phases 1-8)
reviewed: 2026-05-17T00:00:00Z
finding_counts:
  critical: 12
  important: 11
  improvement: 14
  total: 37
---

# BDD Scenario Coverage Review — Phases 1–8

**Reviewed:** 2026-05-17
**Feature files:** `tests/bdd/features/` (24 files)
**Step definitions:** `tests/bdd/steps/algorithms_steps.go`, `tests/bdd/steps/scorer_steps.go`
**Test runner:** `tests/bdd/bdd_test.go`
**Requirements cross-reference:** `docs/requirements.md §15.6`, `.claude/skills/go-testing-standards/SKILL.md` (BDD section)
**Prior review incorporated:** `.planning/phases/08-composite-scorer/08-BDD-REVIEW.md`

---

## Infrastructure Assessment (pre-finding)

**goleak gate:** WIRED. `tests/bdd/bdd_test.go:37` calls `goleak.VerifyTestMain(m)` before the suite runs.

**go.mod isolation:** CORRECT. `tests/bdd/go.mod` lists godog v0.15.0, goleak v1.3.0, testify v1.10.0. Root `go.mod` has zero non-stdlib require lines.

**testify confinement:** CORRECT. Only `tests/bdd/` imports testify; root tests use stdlib `testing`.

**Missing feature files (spec-required, non-existent):** `normalisation.feature`, `determinism.feature`, `scan.feature`, `suppression.feature` — all four are mandated by `docs/requirements.md §15.6` and are absent. This is the most severe structural gap in the suite.

---

## Critical Findings

### [Critical] Four mandatory feature files are entirely absent

- **File:** `tests/bdd/features/` (missing: `normalisation.feature`, `determinism.feature`, `scan.feature`, `suppression.feature`)
- **Phase introduced:** Phases 5 (normalisation), 5 (determinism), 6/9 (scan), 6/9 (suppression)
- **Issue:** `docs/requirements.md §15.6` explicitly names six feature files. Only two are present — `algorithms.feature` (spread across per-algorithm files) and `scorer.feature`. The following four are entirely absent: `normalisation.feature`, `determinism.feature`, `scan.feature`, `suppression.feature`. The Normalisation Pipeline (§9), the Determinism Guarantees (§13), the Scan sub-package (§12), and Suppression composition (§12.3) have zero BDD coverage. These represent entire documented capability layers with no consumer-facing contract documentation or regression detection at the BDD layer.
- **Standard:** `go-testing-standards/SKILL.md` §BDD — "Every algorithm, every Scorer composition pattern, every scan suppression behaviour gets a BDD scenario." `docs/requirements.md §15.6` lists all six feature files as requirements.
- **Action:** Code fix — create all four missing feature files with the scenario sets described in the BDD checklist sections of the agent system prompt.
- **Rationale:** A consumer reading the BDD suite as documentation cannot discover the library's normalisation semantics, determinism guarantees, scan behaviour, or suppression rules from any Gherkin. These are entire undocumented contracts.

---

### [Critical] `ErrInvalidThreshold` has no BDD scenario (carried from 08-BDD-REVIEW.md BDD-01)

- **File:** `tests/bdd/features/scorer.feature`
- **Phase introduced:** Phase 8
- **Issue:** `ErrInvalidThreshold` is a documented sentinel in `errors.go`. The step infrastructure in `scorer_steps.go:338-339` already has the case branch for it. Three other sentinel-error scenarios exist (`ErrMissingThreshold`, `ErrEmptyScorer`, `ErrInvalidWeight`), but `ErrInvalidThreshold` is missing. A consumer passing threshold=1.5 or threshold=-0.1 sees no documented contract at the BDD layer. Additionally, the step regex for threshold in error-path scenarios is `(\d+\.?\d*)` which only matches non-negative numbers — a negative-threshold sub-case cannot be expressed in Gherkin without a regex fix.
- **Standard:** `go-testing-standards/SKILL.md` — "Error scenarios test exact sentinel error matching via `errors.Is`." `docs/requirements.md §8.1` lists `ErrInvalidThreshold` as a construction validation sentinel.
- **Action:** Code fix — add scenario in `scorer.feature`; widen the threshold regex to `(-?\d+\.?\d*)`.
- **Rationale:** An undocumented sentinel is an invisible API contract. The step infrastructure is already present; this is a one-scenario gap.
- **Suggested fix:**
```gherkin
@scorer @errors
Scenario: WithThreshold out-of-range returns ErrInvalidThreshold
  When I attempt to construct a Scorer with Levenshtein weight 1.0 and threshold 1.5
  Then constructing the Scorer should return ErrInvalidThreshold
```

---

### [Critical] `ErrInvalidAlgoID` has no BDD scenario

- **File:** `tests/bdd/features/scorer.feature`
- **Phase introduced:** Phase 8
- **Issue:** `docs/requirements.md §8.1` lists `ErrInvalidAlgoID` as a construction sentinel for: invalid algorithm ID passed to `WithAlgorithm`, and Monge-Elkan inner = `AlgoMongeElkan` (self-reference). Neither sub-case has a BDD scenario. The `constructingTheScorerShouldReturn` step in `scorer_steps.go:327` does not have a case for `ErrInvalidAlgoID` — the switch falls through to the `default` error path, so even if a scenario were written it would fail at the step level.
- **Standard:** `docs/requirements.md §6` sentinel errors documentation; `go-testing-standards/SKILL.md` error-path coverage requirement.
- **Action:** Code fix — add `ErrInvalidAlgoID` case to `scorer_steps.go:constructingTheScorerShouldReturn`, then add scenario in `scorer.feature`.
- **Rationale:** A documented sentinel with no BDD coverage and no step support is invisible to consumers and undetectable by the BDD regression suite.

---

### [Critical] `ErrInvalidQGramSize` has no BDD scenario

- **File:** `tests/bdd/features/scorer.feature`
- **Phase introduced:** Phase 8
- **Issue:** `docs/requirements.md §8.1` lists `ErrInvalidQGramSize` as a construction sentinel (q-gram size < 1 via `WithQGramJaccardAlgorithm`). No scenario exercises this. No step exists for it. The `constructingTheScorerShouldReturn` switch has no case for `ErrInvalidQGramSize`.
- **Standard:** `docs/requirements.md §6` sentinel errors documentation.
- **Action:** Code fix — add step support and scenario.
- **Rationale:** Same pattern as `ErrInvalidAlgoID` — documented sentinel, no coverage.

---

### [Critical] `ErrInvalidTverskyParam` has no BDD scenario

- **File:** `tests/bdd/features/scorer.feature`
- **Phase introduced:** Phase 8
- **Issue:** `docs/requirements.md §8.1` lists `ErrInvalidTverskyParam` as a construction sentinel (Tversky α or β < 0 via `WithTverskyAlgorithm`). No scenario exercises this. No step exists for it. The `constructingTheScorerShouldReturn` switch has no case for `ErrInvalidTverskyParam`.
- **Standard:** `docs/requirements.md §6` sentinel errors documentation.
- **Action:** Code fix — add step support and scenario.
- **Rationale:** Same pattern as above.

---

### [Critical] `iScoreTheSamePairWithTheDefaultScorer` hardcodes the pair (carried from 08-BDD-REVIEW.md BDD-02)

- **File:** `tests/bdd/steps/scorer_steps.go` lines 189–195; `tests/bdd/features/scorer.feature` lines 98–102
- **Phase introduced:** Phase 8
- **Issue:** The step bound to `^I score the same pair with the default Scorer$` ignores all state from the preceding `When I score "XMLParser" and "xml_parser" with the Scorer` step and instead hardcodes `sc.defaultScorer.Score("XMLParser", "xml_parser")` directly. This is a confirmed step bug: the natural-language contract of "the same pair" is violated. The test passes today only because the hardcoded values coincidentally match the scenario's When arguments. A future scenario reusing this step phrase with different inputs will silently score the wrong pair and produce misleading results. `ScorerContext` does not store `lastA`/`lastB` fields.
- **Standard:** `go-testing-standards/SKILL.md` — "Step functions return error rather than calling t.Errorf." The step must honour its Gherkin contract.
- **Action:** Code fix — add `lastA`, `lastB string` fields to `ScorerContext`; populate in `iScoreAndWithTheScorer`; rewrite `iScoreTheSamePairWithTheDefaultScorer` to use `sc.defaultScorer.Score(sc.lastA, sc.lastB)`.
- **Rationale:** This is a confirmed bug causing the scenario to silently test the wrong inputs. The scenario currently passes only by accidental coincidence.

---

### [Critical] Weight auto-normalisation is never directly asserted (carried from 08-BDD-REVIEW.md BDD-03)

- **File:** `tests/bdd/features/scorer.feature` lines 58–67
- **Phase introduced:** Phase 8
- **Issue:** The "Two-algorithm weighted Scorer composes correctly" scenario uses equal weights (0.5 + 0.5) on an identical pair (`"hello"` vs `"hello"`). With equal weights, auto-normalisation is a no-op — the weights already sum to 1.0. On an identical pair, both algorithms return 1.0 regardless of weights. This scenario does not exercise weight normalisation. The mandatory class 4 ("Custom 2-algorithm Scorer with explicit weights") is documented as requiring verification that the composite is the weighted sum, but the current scenario would pass even if normalisation were completely broken.
- **Standard:** `docs/requirements.md §8.4` — weight normalisation is a core semantic; `go-testing-standards/SKILL.md` BDD contract requirement.
- **Action:** Code fix — add a scenario with unequal raw weights on a non-identical pair with a pinned expected composite.
- **Suggested fix:**
```gherkin
@scorer @custom
Scenario: Weight auto-normalisation produces sum-to-1 composite
  # Levenshtein("kitten","sitting") ≈ 0.5714; JaroWinkler("kitten","sitting") ≈ 0.7468.
  # Composite with normalised weights (0.3 + 0.7 → already 1.0): 0.3×0.5714 + 0.7×0.7468 ≈ 0.6942.
  Given I construct a Scorer with Levenshtein weight 0.3 and JaroWinkler weight 0.7 and threshold 0.5
  When I score "kitten" and "sitting" with the Scorer
  Then the Scorer composite score should be approximately 0.6942 within 0.001
  And the Scorer match result should be true
```

---

### [Critical] No Unicode reference vector for Levenshtein

- **File:** `tests/bdd/features/levenshtein.feature`
- **Phase introduced:** Phase 2
- **Issue:** `docs/requirements.md §7.1.1` and the skill's Per-Algorithm Coverage Checklist both require at least one multi-byte UTF-8 example for Unicode-aware algorithms. Levenshtein exposes a `LevenshteinScoreRunes` variant; the feature file has zero Unicode scenarios. The reference example from the checklist template (`"東京"` vs `"東都"` → 0.5) is absent.
- **Standard:** `go-testing-standards/SKILL.md` BDD Per-Algorithm Coverage Checklist — "For Unicode-aware algorithms: at least one multi-byte UTF-8 example."
- **Action:** Code fix — add a Unicode rune-path scenario.
- **Suggested fix:**
```gherkin
Scenario: Unicode rune-level variant (CJK)
  When I compute the LevenshteinRunes score between "東京" and "東都"
  Then the score should be approximately 0.5000 within 0.0001
```

---

### [Critical] No Unicode reference vector for DL-OSA or DL-Full

- **File:** `tests/bdd/features/damerau_osa.feature`, `tests/bdd/features/damerau_full.feature`
- **Phase introduced:** Phase 2
- **Issue:** Both algorithms expose `*ScoreRunes` variants. Neither feature file has a Unicode scenario. The rune-path variant is entirely uncovered at the BDD layer.
- **Standard:** Same as Levenshtein above.
- **Action:** Code fix — add one multi-byte UTF-8 scenario per feature file.

---

### [Critical] DL-OSA has no one-empty scenario

- **File:** `tests/bdd/features/damerau_osa.feature`
- **Phase introduced:** Phase 2
- **Issue:** `docs/requirements.md §7.1.2` states edge cases are "identical to Levenshtein" — which includes one-empty → 0.0. The `damerau_osa.feature` has a both-empty scenario but NO one-empty scenario. The `damerau_full.feature` also lacks a one-empty scenario. The skill checklist requires both-empty and one-empty cases for every algorithm.
- **Standard:** `go-testing-standards/SKILL.md` BDD Per-Algorithm Coverage Checklist.
- **Action:** Code fix — add `Scenario: one-empty string scores 0.0` to both files.

---

### [Critical] SWG has no Unicode reference vector and no raw-score surface coverage

- **File:** `tests/bdd/features/swg.feature`
- **Phase introduced:** Phase 3
- **Issue:** SWG exposes `SmithWatermanGotohScoreRunes` and `SmithWatermanGotohRawScore`/`SmithWatermanGotohRawScoreWithParams`. The feature file has zero Unicode scenarios and zero `*RawScore*` scenarios. `docs/requirements.md §7.1.8` specifically documents the `*RawScore*` variant surface and its rationale ("advanced consumers who want absolute alignment quality"). The BDD suite is the consumer-facing contract; if `*RawScore*` has no scenario, consumers have no documented example of its semantics.
- **Standard:** `docs/requirements.md §7.1.8` requires documentation of both normalised and raw surfaces; skill checklist requires Unicode coverage.
- **Action:** Code fix — add a Unicode rune-path scenario and at least one `*RawScore*` scenario demonstrating that raw scores can differ from clamped scores.

---

### [Critical] Strcmp95 has no Unicode scenario (and is ASCII-only — document it explicitly)

- **File:** `tests/bdd/features/strcmp95.feature`
- **Phase introduced:** Phase 4
- **Issue:** `docs/requirements.md §7.1.7` states Strcmp95 is "ASCII-only; for Unicode input, normalise via `fuzzymatch.Normalise` first." The feature file does not document this ASCII-only restriction with a scenario that demonstrates the required pre-normalisation step. A consumer passing Unicode input directly to `Strcmp95Score` without normalisation will get silent degraded behaviour; there is no BDD documentation of this contract. The SKILL.md checklist requires documenting algorithm-specific behaviour in a separate Scenario.
- **Standard:** Skill checklist — "Scenario for any algorithm-specific behaviour."
- **Action:** Code fix — add a scenario explicitly documenting the ASCII-only restriction and the recommended normalisation pre-step.

---

## Important Findings

### [Important] `WithNormalisation(custom)` composed path has no scenario (carried from 08-BDD-REVIEW.md BDD-04)

- **File:** `tests/bdd/features/scorer.feature`
- **Phase introduced:** Phase 8
- **Issue:** `docs/requirements.md §8.2` lists `WithNormalisation(opts NormalisationOptions)` as a core Scorer option. The `WithoutNormalisation()` path is covered, the default-normalisation path is covered, but `WithNormalisation(custom opts)` — the consumer hook for diacritic-stripping, custom separator chars, or CamelCase-split control — has no scenario.
- **Standard:** Scorer Coverage Checklist — "Scenario: custom Scorer with normalisation options."
- **Action:** Code fix — add scenario and supporting step.

---

### [Important] `ScoreAll` map values are never range-checked (carried from 08-BDD-REVIEW.md BDD-05)

- **File:** `tests/bdd/features/scorer.feature` lines 140–151
- **Phase introduced:** Phase 8
- **Issue:** The "ScoreAll returns map keyed by AlgoID" scenario verifies only key presence/absence, never the values. No assertion that values are in `[0.0, 1.0]`, and no concrete value assertion for any specific algorithm on the tested pair. The `ScoreAll` contract includes that values are per-algorithm scores (not weighted contributions), which is never verified at the BDD layer.
- **Standard:** Scorer Coverage Checklist — ScoreAll returns per-algorithm breakdown with correct values.
- **Action:** Code fix — add value range assertion step and at least one concrete value assertion.

---

### [Important] Hamming has no one-empty scenario

- **File:** `tests/bdd/features/hamming.feature`
- **Phase introduced:** Phase 2
- **Issue:** The feature has both-empty and unequal-length scenarios but no scenario for `HammingScore("abc", "")` returning 0.0. The one-empty case for Hamming is distinct from the unequal-length case (one-empty IS also unequal-length, but the documentation separately defines "one empty" as an edge case in §7.1.4).
- **Standard:** SKILL.md BDD checklist — one-empty case required.
- **Action:** Code fix — add scenario.

---

### [Important] Hamming canonical reference vector table has only 2 rows; the Hamming 1950 paper reference "1011101"/"1001001" is missing

- **File:** `tests/bdd/features/hamming.feature`
- **Phase introduced:** Phase 2
- **Issue:** `docs/requirements.md §7.1.4` lists reference vectors including `"1011101"/"1001001" → distance 2, score ≈ 0.7143`. The Examples table has only `karolin/kathrin` and `abc/abc`. The binary string vector from Hamming 1950 is absent. The skill checklist requires "at least 3 canonical reference vectors from the primary source."
- **Standard:** SKILL.md — "At least 3 canonical reference vectors from the primary source."
- **Action:** Code fix — add the binary-string reference vector.

---

### [Important] Jaro reference vector table is thin (3 from primary source; checklist requires at least 3 but identity is in the table, not from Jaro 1989)

- **File:** `tests/bdd/features/jaro.feature`
- **Phase introduced:** Phase 2
- **Issue:** The Examples table has MARTHA/MARHTA, DIXON/DICKSONX, JELLYFISH/SMELLYFISH, ABC/ABC. The ABC/ABC row is an identity case, not a Jaro 1989 reference vector. The three Jaro 1989 vectors are present, but the checklist note about identity being a separate row rather than a reference-vector row means the table has exactly 3 literature vectors. This is borderline; a stronger table would include an additional non-trivial vector from Winkler 1990 p.357 to cross-validate.
- **Standard:** SKILL.md — "At least 3 canonical reference vectors from the primary source."
- **Action:** Improvement — the minimum is technically met; adding one more literature vector would strengthen coverage.

---

### [Important] SWG cross-validation corpus not exercised via BDD

- **File:** `tests/bdd/features/swg.feature`
- **Phase introduced:** Phase 3
- **Issue:** `docs/requirements.md §7.1.8` documents a committed JSON corpus at `testdata/cross-validation/swg/vectors.json` with 16 entries, cross-validated against biopython's `Bio.Align.PairwiseAligner`. The BDD feature file exercises only 2 reference vectors (http_request/http_request_header_fields and abc/abc), neither of which is from the cross-validation corpus. The Gotoh-erratum canary (`one_long_gap_canary` at biopython_normalised=0.5) is mentioned in the requirements but has no BDD scenario. The gap-split canary in the BDD file uses a different pair than the committed corpus.
- **Standard:** SKILL.md — "Cross-validation scenarios (SWG)."
- **Action:** Code fix — add at least the Gotoh-erratum canary pair as a load-bearing scenario.

---

### [Important] Ratcliff-Obershelp has no symmetry scenario AND no documentation that it is intentionally asymmetric

- **File:** `tests/bdd/features/ratcliff_obershelp.feature`
- **Phase introduced:** Phase 4
- **Issue:** The feature file omits the symmetry scenario "per OQ-1 resolution (LOCKED 2026-05-14)." The comment in the feature file explains this in a code comment, but there is no Gherkin scenario documenting the asymmetry contract for consumers. A consumer reading only the BDD scenarios does not learn that `RatcliffObershelpScore(a, b) != RatcliffObershelpScore(b, a)` is expected behaviour, not a bug. The skill checklist notes that algorithm-specific behaviour should have its own Scenario. The asymmetry is load-bearing and should be documented in a BDD scenario (not just a code comment), even if the scenario is a "then these two scores should differ" assertion rather than a symmetry gate.
- **Standard:** SKILL.md — "Scenario for any algorithm-specific behaviour."
- **Action:** Code fix — add a scenario explicitly demonstrating and documenting the asymmetric-by-design behaviour.
- **Suggested fix:**
```gherkin
Scenario: Score is intentionally asymmetric (difflib CPython bpo-37004 parity)
  # RatcliffObershelp is asymmetric by design — mirrors Python difflib behaviour.
  # This scenario documents the asymmetry is expected, not a bug.
  When I compute the Ratcliff-Obershelp score between "WIKIMEDIA" and "WIKIMANIA"
  And I compute the second Ratcliff-Obershelp score between "WIKIMANIA" and "WIKIMEDIA"
  Then the two Ratcliff-Obershelp scores should differ
```

---

### [Important] Monge-Elkan non-permitted inner panic list is incomplete

- **File:** `tests/bdd/features/monge_elkan.feature` lines 152–160
- **Phase introduced:** Phase 6
- **Issue:** The Examples table for "non-permitted inner AlgoIDs panic" shows only 3 of the 5 rejected entries: `MongeElkan`, `TokenSortRatio`, `TokenJaccard`. Missing: `TokenSetRatio` and `PartialRatio`. `docs/requirements.md §8.1` and the feature file's own comment list all 5. A regression where `TokenSetRatio` or `PartialRatio` is accidentally added to the permitted set would not be caught by this scenario.
- **Standard:** Completeness requirement — all documented rejection cases should be covered.
- **Action:** Code fix — add the two missing rows to the Examples table.

---

### [Important] Ratcliff-Obershelp step `iComputeTheSecondRatcliffObershelpScore` does not exist

- **File:** `tests/bdd/steps/algorithms_steps.go`
- **Phase introduced:** Phase 4
- **Issue:** The algorithms_steps.go explicitly documents (lines 357–365) that "the symmetry scenario is OMITTED per OQ-1 resolution (LOCKED 2026-05-14), so no 'second' / 'equal' steps exist." However, the finding above (asymmetry documentation) recommends adding a "two scores should differ" scenario. That scenario would require a `iComputeTheSecondRatcliffObershelpScore` step. No such step exists. Adding the asymmetry-documentation scenario requires both the feature file addition and a new step.
- **Standard:** Step-definition completeness.
- **Action:** Code fix — add `iComputeTheSecondRatcliffObershelpScoreBetween` step alongside the asymmetry scenario.

---

### [Important] LCSStr has no Unicode scenario despite exposing a `*Runes` variant

- **File:** `tests/bdd/features/lcsstr.feature`
- **Phase introduced:** Phase 4
- **Issue:** `LCSStrScoreRunes` is a public function. The feature file covers only the byte path. The SKILL.md checklist requires a Unicode scenario for Unicode-aware algorithms.
- **Standard:** SKILL.md BDD Per-Algorithm Coverage Checklist.
- **Action:** Code fix — add a multi-byte UTF-8 scenario for the rune-path variant.

---

### [Important] `algoIDFromName` in scorer_steps.go uses `AlgoID.String()` reverse-lookup which requires the string form to match exactly

- **File:** `tests/bdd/steps/scorer_steps.go` lines 471–486
- **Phase introduced:** Phase 8
- **Issue:** `algoIDFromName` iterates `fuzzymatch.AlgoIDs()` and compares via `id.String() == stripped`. If `AlgoID.String()` returns `"damerau_levenshtein_osa"` (snake_case) but the Gherkin scenario uses `"AlgoDamerauLevenshteinOSA"` (PascalCase), the lookup fails. The scorer.feature scenario uses `AlgoDamerauLevenshteinOSA` and `AlgoDoubleMetaphone` — with the `Algo` prefix stripped this becomes `DamerauLevenshteinOSA` and `DoubleMetaphone`. These must match `AlgoID.String()` exactly. If the `String()` method returns `"damerau_levenshtein_osa"` (snake_case), the lookup would return an error at runtime, not at compile time. This is a latent brittle coupling — any rename of an AlgoID's string form silently breaks step lookups.
- **Standard:** Test reliability — step definitions should not have hidden string-matching brittleness.
- **Action:** Discuss-phase — verify that `AlgoID.String()` returns PascalCase (not snake_case) for the step lookup to work, or add a test that explicitly exercises `algoIDFromName` for every `AlgoID`.

---

## Improvement Findings

### [Improvement] No `@character`, `@qgram`, `@token`, `@phonetic`, `@gestalt` category tags on most algorithm feature files

- **File:** All `tests/bdd/features/*.feature` files except `monge_elkan.feature`, `monge_elkan_phonetic_inner.feature`, `token_sort_ratio.feature`, `token_set_ratio.feature`, `token_jaccard.feature`, `partial_ratio.feature`, `nysiis.feature`, `mra.feature`
- **Phase introduced:** Phases 2–7
- **Issue:** The SKILL.md BDD section requires every scenario to be tagged with at least one tag. Most algorithm feature files (`levenshtein.feature`, `hamming.feature`, `jaro.feature`, `jarowinkler.feature`, `strcmp95.feature`, `swg.feature`, `lcsstr.feature`, `damerau_osa.feature`, `damerau_full.feature`, `qgram_jaccard.feature`, `sorensen_dice.feature`, `cosine.feature`, `tversky.feature`, `soundex.feature`, `double_metaphone.feature`, `ratcliff_obershelp.feature`) have zero tags on their scenarios. Only the Phase 6 token algorithms and Phase 7 phonetic algorithms have category tags.
- **Standard:** SKILL.md — "Every scenario MUST have at least one tag."
- **Action:** Code fix — add category tags (`@character`, `@qgram`, `@gestalt`, `@phonetic`) and layer tags (`@algorithm`) to all scenarios in the untagged feature files.

---

### [Improvement] `WithNormaliseWeights(false)` opt-out has no BDD scenario

- **File:** `tests/bdd/features/scorer.feature`
- **Phase introduced:** Phase 8
- **Issue:** `docs/requirements.md §8.4` documents `WithNormaliseWeights(false)` as an opt-in escape hatch that produces a raw weighted sum. No scenario exercises this path. The CONTEXT.md golden corpus has a mandatory entry for this case.
- **Standard:** Scorer Coverage Checklist — "Scenario: weight normalisation behaviour (`WithNormaliseWeights(true)` default, `(false)` opt-out)."
- **Action:** Code fix — add scenario.

---

### [Improvement] `DefaultScorer()` composition is never explicitly verified at BDD layer

- **File:** `tests/bdd/features/scorer.feature`
- **Phase introduced:** Phase 8
- **Issue:** The Scorer Coverage Checklist requires "Scenario: `DefaultScorer` matches a known similar pair above its threshold" (covered — scenario 1) and also that the default composition is documented. The actual algorithm list (DL-OSA 0.30, JaroWinkler 0.20, TokenJaccard 0.20, QGramJaccard 0.15, SorensenDice 0.10, DoubleMetaphone 0.05) is never verified at the BDD layer. A scenario asserting `Algorithms()` returns exactly these 6 algorithms with approximately these weights would pin the default composition contract.
- **Standard:** Scorer Coverage Checklist.
- **Action:** Code fix — add scenario asserting DefaultScorer algorithm count and algorithm membership.

---

### [Improvement] Concurrent scenario tests `Score` only; `ScoreAll` and `Match` concurrent safety undocumented

- **File:** `tests/bdd/features/scorer.feature` lines 129–138
- **Phase introduced:** Phase 8
- **Issue:** `docs/requirements.md §8.1` states "`Score`, `ScoreAll`, and `Match` are safe for concurrent use." The concurrent scenario tests only `Score`. `ScoreAll` and `Match` concurrent safety have no BDD documentation.
- **Standard:** Scorer Coverage Checklist — "Scenario: concurrent Score / ScoreAll / Match is safe."
- **Action:** Code fix — add concurrent `ScoreAll` and `Match` scenarios (or expand the existing scenario to include all three methods).

---

### [Improvement] `WithoutNormalisation` scenario uses relative assertion; no concrete value (carried from 08-BDD-REVIEW.md BDD-07)

- **File:** `tests/bdd/features/scorer.feature` lines 90–102
- **Phase introduced:** Phase 8
- **Issue:** The assertion `the no-normalisation composite should be less than the default composite` is relative, not concrete. If both scores changed proportionally (e.g. a bug that halved all scores), the relative assertion would still pass while the absolute behaviour had changed.
- **Standard:** Concrete value assertions anchor contracts better than relative ones.
- **Action:** Code fix — add concrete range assertions from the golden file.

---

### [Improvement] Scorer scenarios have no cross-layer `@levenshtein`, `@jaro_winkler` tags

- **File:** `tests/bdd/features/scorer.feature`
- **Phase introduced:** Phase 8
- **Issue:** Scorer scenarios that compose specific algorithms (e.g. "Single-algorithm Scorer composes correctly" with Levenshtein) have no secondary tag linking them to the algorithm's own feature file. This makes filtered runs by algorithm incomplete — `godog --tags=@levenshtein` would not find the Scorer-layer Levenshtein scenario.
- **Standard:** SKILL.md tag taxonomy.
- **Action:** Improvement — add per-algorithm secondary tags to relevant Scorer scenarios.

---

### [Improvement] SWG `both SmithWatermanGotoh scores should be equal` step is algorithm-specific but the gap-split canary's assertion is not the gap-split property — it checks symmetry instead

- **File:** `tests/bdd/features/swg.feature` lines 43–46
- **Phase introduced:** Phase 3
- **Issue:** The "gap-split canary" scenario asserts that two different SWG inputs (`abc________def/abcdef` vs `abc____def____/abcdef`) produce the same score. This is not a gap-split invariance test — it is a symmetry test on two distinct structurally similar pairs. The true Gotoh gap-split invariant from `docs/requirements.md §7.1.8` states that "splitting a single long gap into two halves with intervening match characters that don't affect the local alignment must NOT improve the score." The current scenario does not exercise this specific invariant.
- **Standard:** Algorithm-specific invariant coverage.
- **Action:** Improvement — clarify the scenario comment or replace with a true gap-split invariant test.

---

### [Improvement] Double Metaphone "Slavic" and "Chinese-origin" gate scenarios assert only non-emptiness, not actual key values

- **File:** `tests/bdd/features/double_metaphone.feature` lines 61–68
- **Phase introduced:** Phase 7
- **Issue:** The Slavic (Sczepanski) and Chinese-origin (Cheung) gate scenarios assert only `both keys should be non-empty`. This is a very weak assertion — any non-empty output would pass, including wildly wrong keys. The SKILL.md checklist requires at least 3 canonical reference vectors from the primary source; these two scenarios do not contribute reference-vector coverage.
- **Standard:** SKILL.md — "At least 3 canonical reference vectors from the primary source."
- **Action:** Improvement — replace non-empty assertions with concrete expected key values from cross-validated sources.

---

### [Improvement] Soundex feature file has no `@phonetic` tag on most scenarios

- **File:** `tests/bdd/features/soundex.feature`
- **Phase introduced:** Phase 7
- **Issue:** The Feature-level declaration has no `@phonetic` tag, and individual scenarios are untagged. NYSIIS and MRA have `@phonetic` tags; Soundex and Double Metaphone are inconsistent.
- **Standard:** SKILL.md tag requirement.
- **Action:** Code fix — add `@phonetic` and `@soundex` tags to all scenarios.

---

### [Improvement] Double Metaphone feature has no `@phonetic` tag on the Feature declaration or most scenarios

- **File:** `tests/bdd/features/double_metaphone.feature`
- **Phase introduced:** Phase 7
- **Issue:** Same as Soundex above — no `@phonetic @double_metaphone` tags.
- **Standard:** SKILL.md tag requirement.
- **Action:** Code fix — add tags.

---

### [Improvement] `theDistanceShouldBe` step is algorithm-agnostic by design but this creates potential cross-scenario state confusion

- **File:** `tests/bdd/steps/algorithms_steps.go` lines 144–153
- **Phase introduced:** Phase 2
- **Issue:** The step comment documents that `theDistanceShouldBe` "matches the value written by whichever *Distance* step ran most recently in the current scenario." This is intentional but means if a scenario chains two distance computations, the assertion applies only to the last one. No scenario currently does this, but the design is fragile. The comment recommends introducing algorithm-suffixed steps if needed, but no guidance exists in the feature files about this limitation.
- **Standard:** Step-definition clarity.
- **Action:** Improvement — add a comment in each feature file that uses distance assertions noting this design constraint, to prevent future misuse.

---

### [Improvement] `iListTheScorerAlgorithms` is a no-op step — the When step does no work

- **File:** `tests/bdd/steps/scorer_steps.go` lines 222–227
- **Phase introduced:** Phase 8
- **Issue:** `iListTheScorerAlgorithms` validates `sc.scorer != nil` and returns nil. The subsequent assertion steps call `sc.scorer.Algorithms()` directly on each call rather than reading from state populated by the When step. This violates the Given/When/Then partition — the When step should produce observable state that Then steps assert on.
- **Standard:** Gherkin readability and correct Given/When/Then partitioning.
- **Action:** Improvement — make the step populate a `lastAlgorithms []fuzzymatch.ScorerAlgorithm` field in `ScorerContext`; have assertion steps read from that field.

---

### [Improvement] Ratcliff-Obershelp autojunk-sensitive scenario uses hardcoded pair construction in Go step, not in Gherkin Examples

- **File:** `tests/bdd/features/ratcliff_obershelp.feature` lines 53–64; `tests/bdd/steps/algorithms_steps.go` lines 370–408
- **Phase introduced:** Phase 4
- **Issue:** The autojunk-sensitive scenario uses a dedicated step `I compute the Ratcliff-Obershelp score for the autojunk-sensitive pair` that constructs the 205-char inputs via `strings.Repeat` in Go. This is technically correct (the comment explains why — 205-char Gherkin literals are hard to read). However, this means the pair is invisible to consumers reading the BDD documentation. The Gherkin should at minimum describe the pair construction in a comment.
- **Standard:** Consumer-facing documentation quality.
- **Action:** Improvement — add a comment in the Gherkin scenario describing the pair construction logic so a reader does not need to inspect the step definition.

---

### [Improvement] Scorer `algoIDFromName` and Monge-Elkan `algoIDByName` are two separate lookup functions doing the same thing

- **File:** `tests/bdd/steps/scorer_steps.go` lines 471–486; `tests/bdd/steps/algorithms_steps.go` lines 835–887
- **Phase introduced:** Phase 6 (algoIDByName), Phase 8 (algoIDFromName)
- **Issue:** Two separate AlgoID name→enum lookup functions exist in the step definitions: `algoIDByName` (switch-based, returns -1 on unknown) and `algoIDFromName` (iterates `AlgoIDs()`, returns error on unknown). They are not identical: `algoIDByName` is a switch with explicit cases; `algoIDFromName` iterates `AlgoIDs()`. If a new AlgoID is added, `algoIDByName` requires a new case but `algoIDFromName` picks it up automatically. This creates a maintenance coupling — `algoIDByName` can silently miss new AlgoIDs while `algoIDFromName` won't.
- **Standard:** DRY step definitions — no duplication.
- **Action:** Improvement — consolidate to a single lookup function, preferring the `AlgoIDs()` iterator approach of `algoIDFromName` for automatic completeness.

---

## Summary

### Feature file presence vs requirements

| Required file (§15.6) | Status |
|---|---|
| `algorithms.feature` (per-algorithm, split across 23 files) | Present — partially complete |
| `scorer.feature` | Present — gaps documented above |
| `normalisation.feature` | **ABSENT** |
| `determinism.feature` | **ABSENT** |
| `scan.feature` | **ABSENT** |
| `suppression.feature` | **ABSENT** |

### Gap count by severity

| Severity | Count |
|---|---|
| Critical | 12 |
| Important | 11 |
| Improvement | 14 |
| **Total** | **37** |

### Recommendation: NO-GO

The BDD suite has four structurally missing feature files (`normalisation.feature`, `determinism.feature`, `scan.feature`, `suppression.feature`) that represent entire undocumented capability layers. It has a confirmed step bug (`iScoreTheSamePairWithTheDefaultScorer` hardcodes its pair — from 08-BDD-REVIEW.md BDD-02) that causes a scenario to silently test wrong inputs. Three documented sentinel errors (`ErrInvalidAlgoID`, `ErrInvalidQGramSize`, `ErrInvalidTverskyParam`) have no BDD scenarios and no step infrastructure. `ErrInvalidThreshold` has no scenario (step infrastructure exists). Weight auto-normalisation is not actually exercised by the current two-algorithm scenario.

These deficiencies mean the BDD suite cannot serve as a reliable consumer-facing contract or regression detection layer for a significant portion of the documented API. The four missing feature files alone represent the largest gap — they are required by `docs/requirements.md §15.6` and must be created before Phase 9 (scan) work begins.

_Reviewed: 2026-05-17_
_Reviewer: bdd-scenario-reviewer_

</details>

---

## 6. code-reviewer

_Source: `.planning/reviews/code-FINDINGS.md`_

<details>
<summary>Click to expand full report</summary>

---
status: issues_found
agent: code-reviewer
scope: entire codebase (phases 1-8)
reviewed: 2026-05-17T06:34:34Z
finding_counts:
  critical: 3
  important: 18
  improvement: 24
  total: 45
---

# Comprehensive code review — fuzzymatch (Phases 1–8)

Findings are organised by severity; severity tags are organisational only. Every refactoring opportunity is surfaced regardless of perceived priority.

---

## CRITICAL — bugs that produce wrong results, panic paths, or contract violations

### [Critical] `HammingDistance` signature deviates from spec — silent wrong answer on length mismatch
- **File:** `/Users/johnny/Development/fuzzymatch/hamming.go:69-90`
- **Phase introduced:** Phase 2
- **Issue:** `HammingDistance(a, b string) int` (no error return). On unequal-length inputs it silently returns `max(len(a), len(b))`. `docs/requirements.md` §7 line 362 specifies the signature `HammingDistance(a, b string) (int, error)` returning `ErrHammingLengthMismatch`. `.claude/skills/algorithm-correctness-standards/SKILL.md` line 104 reiterates this. `.planning/research/PITFALLS.md` line 610 documents the rule: "Functions returning `int` distance with no error for length mismatch in Hamming → Silent wrong answer when caller passes unequal-length strings → `HammingDistance` returns `(int, error)` with `ErrHammingLengthMismatch`".
- **Standard:** algorithm-correctness-standards §"Edge cases"; docs/requirements.md §7
- **Action:** Discuss-phase needed — either (a) update spec and skills to reflect "silent return max(len)" decision and document the deviation, or (b) introduce `ErrHammingLengthMismatch` in errors.go and break the signature to `(int, error)`. Option (b) is a breaking API change.
- **Rationale:** Three documents (requirements.md, skill, pitfalls) agree that the current implementation is wrong. Either the spec is wrong or the code is. Until reconciled, downstream consumers may silently misuse Hamming on unequal inputs.

### [Critical] `MongeElkanScore` asymmetry property test has a documented failure in `bench.txt.new`
- **File:** `/Users/johnny/Development/fuzzymatch/bench.txt.new` (test failure record)
- **Phase introduced:** Phase 6 (Monge-Elkan)
- **Issue:** A committed `bench.txt.new` file at the repo root captures `--- FAIL: TestProp_MongeElkanScore_AsymmetricWhenTokenCountAsymmetric` with a multi-rune random input. Either (a) the test premise (`strings.Fields` token count proxy) is buggy and produces false positives because `Tokenise`'s separator set does not include Unicode whitespace like ` ` (figure space, U+2007), or (b) Monge-Elkan really produces equal scores on inputs where `strings.Fields` reports different token counts but `Tokenise` reports the same — which would be a property-test correctness gap.
- **Standard:** go-testing-standards §"Property tests"
- **Action:** Investigate, then either fix the test premise to gate on `len(Tokenise(...))` instead of `len(strings.Fields(...))` OR delete the leftover `bench.txt.new` file with an issue-linked CHANGELOG entry. Do NOT leave a failing test record committed to the working tree.
- **Rationale:** A persistent test-failure artefact in the repo undermines confidence and creates "is this a known issue?" ambiguity for future reviewers.

### [Critical] `WithTverskyAlgorithm` does not enforce α+β > 0 — runtime panic escapes Scorer
- **File:** `/Users/johnny/Development/fuzzymatch/scorer_options.go:381-399`
- **Phase introduced:** Phase 8
- **Issue:** The option validates `alpha < 0`, `beta < 0`, `weight <= 0`, `n < 1`, but does NOT check `(alpha == 0 && beta == 0)`. A Scorer constructed with `WithTverskyAlgorithm(weight=1.0, alpha=0, beta=0, n=3)` succeeds at NewScorer time but then panics inside `TverskyScore` (tversky.go:241) when `Score` is first called. The godoc (lines 377-380) claims "α + β > 0 constraint is enforced at runtime by TverskyScore itself; this option does not re-check it" — but the project's documented invariant is that the Scorer option layer returns `ErrInvalidTverskyParam` (typed error), reserving panic for direct calls only (CONTEXT.md §5 LOCKED). The current code violates that invariant.
- **Standard:** go-coding-standards §"API Design" ("No `log.Fatal`, `os.Exit`, or `panic` that escapes the package boundary"); errors.go line 79-88 sentinel godoc.
- **Action:** Code fix — add `if alpha == 0 && beta == 0 { return ErrInvalidTverskyParam }` to `WithTverskyAlgorithm` (and document the Scorer-vs-direct-call divergence in WithTverskyAlgorithm's godoc).
- **Suggested fix:**
  ```go
  if alpha < 0 || beta < 0 {
      return ErrInvalidTverskyParam
  }
  if alpha == 0 && beta == 0 {
      return ErrInvalidTverskyParam
  }
  ```
- **Rationale:** Scorer construction succeeding for a configuration that subsequently panics is exactly the failure mode the typed-error sentinel was designed to prevent.

---

## IMPORTANT — standards violations, missing tests, refactoring needs

### [Important] `WithMongeElkanAlgorithm` only rejects `inner == AlgoMongeElkan`, not the broader allowlist — runtime panic escapes Scorer
- **File:** `/Users/johnny/Development/fuzzymatch/scorer_options.go:425-446`
- **Phase introduced:** Phase 8
- **Issue:** The option validates `dispatch[inner] != nil` and `inner != AlgoMongeElkan`, but does NOT check membership in `permittedMongeElkanInner` (the 18-entry allowlist in monge_elkan.go:291-317). A consumer calling `WithMongeElkanAlgorithm(weight=1.0, inner=AlgoTokenSortRatio)` constructs the Scorer successfully but the underlying `MongeElkanScoreSymmetric` panics at Score-time with "AlgoID TokenSortRatio not permitted as Monge-Elkan inner metric". The godoc acknowledges this (lines 417-419) — "Passing an inner AlgoID that the underlying ME implementation rejects will panic at Score time (programmer error); the panic surfaces via godog's recover mechanism in plan 08-04's BDD scenarios" — but this contradicts the documented design: Scorer option errors should be typed, not panics.
- **Standard:** go-coding-standards §"API Design"; CONTEXT.md §3 LOCKED for Phase 6.
- **Action:** Code fix — check `!permittedMongeElkanInner[inner]` in the option layer and return `ErrInvalidAlgorithm`.
- **Rationale:** Same as the Tversky finding above — Scorer construction should never produce a *Scorer that panics on first use.

### [Important] `damerau_full.go` heap-allocates a full O(m·n) DP table — large alloc on long inputs
- **File:** `/Users/johnny/Development/fuzzymatch/damerau_full.go:226-321`
- **Phase introduced:** Phase 2
- **Issue:** `damerauFullDP` allocates `make([]int, (m+2)*(n+2))` regardless of input length. For two 100-char inputs that's 102×102 = 10404 ints = ~83KB. For 10,000-char inputs that's ~800MB. No upper-bound guard; the algorithm trusts the caller. The file header (lines 70-72) documents this as a v1.x perf follow-up. There is no `#TBD` issue link.
- **Standard:** performance-standards §"Per-Algorithm Budgets"; CLAUDE.md "GitHub Issues Are the Source of Truth"
- **Action:** File a GitHub issue tracking the two-row + auxiliary-anchor-table optimisation; reference the issue number in the file header.
- **Rationale:** Untracked performance regression risk on pathological long inputs; the threat-model says "callers control input size" but a 100MB+ allocation from a single string-similarity call is surprising for a library promising "deterministic, production-grade" behaviour.

### [Important] `roMatchedLength` (Ratcliff-Obershelp) uses unbounded recursion on user input
- **File:** `/Users/johnny/Development/fuzzymatch/ratcliff_obershelp.go:200-210`
- **Phase introduced:** Phase 4
- **Issue:** `roMatchedLength` recurses into left+right segments without depth-limiting. For pathological inputs (e.g. inputs with many short common substrings) recursion depth can reach O(min(la, lb)). Go default goroutine stack grows dynamically, but an attacker-controlled input of e.g. 100,000 bytes could trigger 100,000 stack frames — uses ~6.4MB stack at 64 bytes/frame. The file header (lines 79-84) acknowledges the recursion contract but doesn't bound depth.
- **Standard:** performance-standards §"DoS notice"; security-reviewer scope per CLAUDE.md
- **Action:** Either (a) document the maximum input length the recursion is safe for, (b) convert to an iterative implementation with an explicit work-queue, or (c) add a depth-limit gate that falls back to a clamped score on overflow.
- **Rationale:** The threat model in CLAUDE.md flags "DoS via pathological inputs" as a security-reviewer concern; this recursion is the most prominent unguarded path.

### [Important] `partial_ratio.go` TODO without GitHub issue reference
- **File:** `/Users/johnny/Development/fuzzymatch/partial_ratio.go:148-154`
- **Phase introduced:** Phase 6
- **Issue:** `TODO(#TBD): implement sliding-window DP per Bachmann RapidFuzz docs ... A future GitHub issue will track the sliding-window DP implementation; this TODO will be updated with the issue number once it is created.`
- **Standard:** CLAUDE.md "Every `TODO` must reference a GitHub issue: `// TODO(#42): add Ukkonen banding optimisation`"
- **Action:** Open a GitHub issue and replace `#TBD` with the issue number.
- **Rationale:** Project rule.

### [Important] Inconsistent dispatch slot comments — three are wrong, one is missing
- **File:** `/Users/johnny/Development/fuzzymatch/dispatch_swg.go:17` ("slot 6", should be 7); `/Users/johnny/Development/fuzzymatch/dispatch_soundex.go:17` ("slot 23", actual iota 18); `/Users/johnny/Development/fuzzymatch/dispatch_double_metaphone.go:17` ("slot 24", actual iota 19); `/Users/johnny/Development/fuzzymatch/dispatch_mra.go:15-16` ("slot 26", actual iota 21); `/Users/johnny/Development/fuzzymatch/dispatch_nysiis.go:15` ("slot ... (25 — see algoid.go)", actual iota 20)
- **Phase introduced:** Phases 2, 3, 5, 7
- **Issue:** Multiple dispatch files cite incorrect iota slot numbers. AlgoSoundex is iota 18 (Phase 7 added 4 phonetic entries AFTER position 17 token_jaccard); dispatch_soundex.go claims slot 23. Similarly for double_metaphone (19), NYSIIS (20), MRA (21). dispatch_swg.go says slot 6 but Strcmp95 is at 6 and SmithWatermanGotoh at 7.
- **Standard:** documentation-standards §"godoc accuracy"
- **Action:** Either (a) drop the explicit slot numbers from the comments — they duplicate algoid.go and drift — or (b) audit and correct all of them. Option (a) is preferable.
- **Rationale:** The comments are load-bearing for future reviewers; incorrect numbers undermine trust.

### [Important] `WriteGoldenFile` is a public exported function but exists only for test maintenance
- **File:** `/Users/johnny/Development/fuzzymatch/golden_canonical.go:88-100`
- **Phase introduced:** Phase 1
- **Issue:** `WriteGoldenFile(path string, v any) error` is exported from production code but its godoc says "It is intended for test maintenance only — production code never invokes it." This pollutes the public API surface — consumers see this function on pkg.go.dev and may use it for unintended purposes. The file `golden_canonical.go` is in the production package (not `_test.go`).
- **Standard:** go-coding-standards §"API Design"; api-ergonomics-reviewer scope.
- **Action:** Move `WriteGoldenFile` (and possibly `canonicalMarshal` if it has no production callers) into a `_test.go` file (e.g. `golden_canonical_test.go` or a new test helper). Use `export_test.go` to re-export internal symbols only as needed.
- **Rationale:** Public API hygiene; the `-update` workflow can live in tests.

### [Important] `Tokenise` allocates per-token byte buffer even for pure-ASCII tokens
- **File:** `/Users/johnny/Development/fuzzymatch/tokenise.go:323-338`
- **Phase introduced:** Phase 1
- **Issue:** `appendToken` allocates `make([]byte, 0, len(rs)*utf8.UTFMax)` — 4× over-provisioning for ASCII tokens — then `string(buf)` causes another allocation. Two allocations per token even for ASCII identifiers. The comment acknowledges "profiling can revisit if benchmarks show this is a bottleneck" but does not link an issue.
- **Standard:** performance-standards §"ASCII Fast Path Pattern"
- **Action:** Add an ASCII fast path — when all runes in `rs` are < 0x80, allocate `make([]byte, len(rs))` exactly and use bitwise-OR lowercasing inline.
- **Rationale:** Hot-path performance for token-based algorithms (Monge-Elkan, Token Sort/Set Ratio, Partial Ratio, Token Jaccard) — each one calls Tokenise twice.

### [Important] `extractQGrams` map keys are sub-slices that retain backing storage of the input
- **File:** `/Users/johnny/Development/fuzzymatch/q_gram.go:104-117`
- **Phase introduced:** Phase 5
- **Issue:** `m[s[i:i+n]]++` uses sub-slices into the input string as map keys. Strings are immutable so this is safe for correctness, but the map retains references to the underlying string data — if a consumer extracts a small q-gram map from a multi-MB input then discards the input, the input's backing array stays alive as long as any map key references it. The file header (lines 76-77) acknowledges this but doesn't warn the consumer.
- **Standard:** documentation-standards §"godoc"
- **Action:** Add a godoc note to the algorithm functions (or just the algorithm-correctness-reviewer can decide whether to copy on capture).
- **Rationale:** Retained-reference behaviour is unusual; should be documented even if intentional.

### [Important] `Strcmp95` source citation drift — algoid.go cites Winkler 1991, file cites Winkler 1994
- **File:** `/Users/johnny/Development/fuzzymatch/algoid.go:90-95` vs `/Users/johnny/Development/fuzzymatch/strcmp95.go:18-19`
- **Phase introduced:** Phase 3
- **Issue:** algoid.go's `AlgoStrcmp95` godoc cites "Winkler & Thibaudeau 1991 — An application of the Fellegi-Sunter model of record linkage to the 1990 U.S. decennial census". strcmp95.go's file header cites "Winkler, W. E. (1994). Advanced methods for record linkage. ...§3". Both citations are real Winkler publications but the algorithm-correctness discipline requires a single primary source. The file header is the authoritative source per algorithm-correctness-standards §"Primary Source Citation".
- **Standard:** algorithm-correctness-standards §"Primary Source Citation"
- **Action:** Align — either update algoid.go to cite Winkler 1994 (matching the file) or update strcmp95.go to cite Winkler 1991. The implementation actually transcribes the 1994 TR-2 36-pair table, so 1994 is correct.
- **Rationale:** Consistency between catalogue entry and implementation file.

### [Important] `JaroWinklerScoreRunes` does redundant `runeSlicesEqual` inside `jaroRunes`
- **File:** `/Users/johnny/Development/fuzzymatch/jaro.go:255-257` (called from jarowinkler.go:165)
- **Phase introduced:** Phase 2
- **Issue:** `jaroRunes` performs `if runeSlicesEqual(ra, rb) { return 1.0 }` after `JaroScoreRunes`/`JaroWinklerScoreRunes` have already gated on `a == b` (string equality). Since `[]rune(s)` is deterministic, two strings that differ MUST produce different rune slices UNLESS both contain invalid UTF-8 byte sequences that fold to U+FFFD identically. The check is reachable only for malformed UTF-8 inputs.
- **Standard:** go-coding-standards §"Complexity"
- **Action:** Either document this rationale inline (the check is correct, just non-obvious) or remove it and add a test case for the malformed-UTF-8 path.
- **Rationale:** Code review readability — readers wonder why the check exists.

### [Important] `runeAt` in `soundex.go` re-implements `utf8.DecodeRuneInString` without continuation-byte validation
- **File:** `/Users/johnny/Development/fuzzymatch/soundex.go:271-300`
- **Phase introduced:** Phase 7
- **Issue:** `runeAt` is shared across soundex.go, nysiis.go, mra.go, double_metaphone.go. It decodes UTF-8 length prefixes but does NOT validate that continuation bytes are 10xxxxxx — for malformed UTF-8 where the prefix byte says "2-byte sequence" but the next byte is e.g. 11xxxxxx, the function still returns a 2-byte step but with wrong rune value. Stdlib `utf8.DecodeRuneInString` does this correctly. The comment "avoids an extra import in a small file; the logic is identical to utf8.DecodeRuneInString for the purpose of skip-counting" is inaccurate.
- **Standard:** go-coding-standards §"Dependencies"; algorithm-correctness-standards §"Unicode Handling"
- **Action:** Replace `runeAt` with `utf8.DecodeRuneInString` (add the `unicode/utf8` import). The callers only use the size return; the stdlib function is bytes-equivalent for size on valid input and slightly safer on malformed input.
- **Rationale:** Stdlib already provides the correct primitive; rolling your own is unjustified complexity and slightly wrong.

### [Important] `MRACompare` step 3 L→R inner loop has a stale `matchedA[i]` guard
- **File:** `/Users/johnny/Development/fuzzymatch/mra.go:285-293`
- **Phase introduced:** Phase 7
- **Issue:** The inner loop conditions `if !matchedA[i] && !matchedB[j] && codexA[i] == codexB[j]` — but `matchedA[i]` is checked inside a loop where i is the OUTER index. Once `matchedA[i]` is set, the break fires and the outer loop advances. So the `!matchedA[i]` check is always true on entry (the outer loop has just started this `i`). It's a no-op guard that adds a tiny CPU cost and adds reading confusion.
- **Standard:** go-coding-standards §"Complexity"
- **Action:** Remove the redundant `!matchedA[i]` check inside the inner loop body. Keep only `!matchedB[j] && codexA[i] == codexB[j]`.
- **Rationale:** Tiny correctness-preserving simplification; current code is harder to verify.

### [Important] `MRACompare` step 3 doesn't require `j >= i` — semantic divergence from "L→R common-character removal"
- **File:** `/Users/johnny/Development/fuzzymatch/mra.go:285-293`
- **Phase introduced:** Phase 7
- **Issue:** The NBS Tech Note 943 step 3 description is "process L→R, remove identical characters from both codexes". The current implementation matches char A[i] to ANY unmatched char B[j] (any j, including j < i). This is the standard interpretation but it's not strictly L→R — it's "for each A[i], find the leftmost unmatched B[j]". Compare with jellyfish (which uses the strict-LR variant: `for j >= i`). The implementation matches the most permissive interpretation, which may produce higher similarity scores than the strict variant.
- **Standard:** algorithm-correctness-standards §"Reference Vectors"
- **Action:** Cross-validate against the committed `testdata/cross-validation/phonetic/vectors.json` and document which variant the implementation matches. If cross-validation passes against jellyfish, the current code is correct; otherwise it diverges.
- **Rationale:** The algorithm-correctness-reviewer should sign off explicitly on the interpretation choice.

### [Important] `double_metaphone.go` line 563 has ambiguous operator precedence
- **File:** `/Users/johnny/Development/fuzzymatch/double_metaphone.go:563`
- **Phase introduced:** Phase 7
- **Issue:** `if i == 0 && at(i+4) == 0 || dmContains(v, 0, "SAN") {` — Go evaluates as `(i == 0 && at(i+4) == 0) || dmContains(v, 0, "SAN")` per `&&`-higher-than-`||` precedence. Reading code is non-obvious; a reviewer might mis-read as `i == 0 && (at(i+4) == 0 || dmContains(v, 0, "SAN"))`.
- **Standard:** go-coding-standards §"Complexity"
- **Action:** Add explicit parentheses: `if (i == 0 && at(i+4) == 0) || dmContains(v, 0, "SAN") {`.
- **Rationale:** Readability and review-safety; same logic, clearer code.

### [Important] `scorer.go` file header claims `sort` is imported but it isn't
- **File:** `/Users/johnny/Development/fuzzymatch/scorer.go:69-72`
- **Phase introduced:** Phase 8
- **Issue:** Header comment: "the only stdlib dependency is 'sort' (not even strictly necessary at this plan boundary — AlgoIDs() returns the canonical order — but reserved for plan 08-03's Algorithms() accessor which sorts a fresh slice copy)." But (a) `sort` is NOT actually imported in scorer.go, (b) `Algorithms()` at line 460 does NOT sort — it just iterates the already-sorted slice.
- **Standard:** documentation-standards §"godoc accuracy"
- **Action:** Delete the misleading sentence from the file header.
- **Rationale:** Comment drift.

### [Important] `scorer_options_internal_test.go` godoc misnames package suffix
- **File:** `/Users/johnny/Development/fuzzymatch/scorer_options_internal_test.go:25-28`
- **Phase introduced:** Phase 8
- **Issue:** Comment: "Living in package fuzzymatch (no _test suffix) is the conventional Go pattern for exposing package-internal state to external test files: the build-tag _test.go suffix ensures this file never ships in the public artifact." But the file IS named `_test.go` (suffix `_test.go` excludes from production builds). The comment confuses "package fuzzymatch_test" with "file _test.go" — these are different mechanisms.
- **Standard:** documentation-standards §"godoc accuracy"
- **Action:** Reword to "Living in package fuzzymatch (not fuzzymatch_test) is the conventional Go pattern for exposing package-internal state to external test files via the _test.go file suffix, which excludes the file from production builds."
- **Rationale:** Same mistake in `/Users/johnny/Development/fuzzymatch/scorer_internal_test.go:28-30`.

### [Important] Map iteration in `extractQGrams` consumers — comment claims DET-03 satisfied via integer-counter exit, but Cosine's intersection key build iterates `small` map
- **File:** `/Users/johnny/Development/fuzzymatch/cosine.go:319-323`
- **Phase introduced:** Phase 5
- **Issue:** `for k := range small { if _, ok := large[k]; ok { intersectionKeys = append(intersectionKeys, k) } }` — this iterates a map and builds an ordered slice. The comment on line 311-313 says "the slice content is identical regardless of which side is iterated (intersection is symmetric)". This is TRUE for the SET of keys, but `intersectionKeys` is a slice — its ORDER depends on map iteration order, which is non-deterministic per Go map semantics. Subsequent `sort.Strings(intersectionKeys)` normalises the order, so the final reduction IS deterministic. But the comment claim that "the slice content is identical regardless of which side is iterated" omits the load-bearing role of the sort. This is correct code but partially misleading documentation.
- **Standard:** determinism-standards §"The No-Map-Iteration Rule"
- **Action:** Clarify the comment: "the slice CONTENTS (as a set) are identical regardless of which side is iterated; the slice ORDER is non-deterministic until the subsequent sort.Strings call normalises it before the dot-product reduction."
- **Rationale:** Doc accuracy; the sort is load-bearing per CONTEXT.md §3 LOCKED but the comment understates this.

### [Important] `LongestCommonSubstring` returns a slice header sharing backing storage with input
- **File:** `/Users/johnny/Development/fuzzymatch/lcsstr.go:116-155`
- **Phase introduced:** Phase 3
- **Issue:** `LongestCommonSubstring` returns `a[endI-maxLen : endI]` — a sub-slice of the input. Per Go's string semantics this is safe (strings are immutable), but it keeps the input's backing array alive. The godoc DOES mention this (lines 116-126), which is good. However, this is a public API behaviour that consumers may not expect — typical Go libraries return defensively-copied strings.
- **Standard:** documentation-standards §"godoc"
- **Action:** Consider whether to defensively copy in the byte variant (matching the rune variant which does `string(ra[endI-maxLen : endI])` — itself an allocation). Document the rationale either way.
- **Rationale:** API ergonomics review; the byte variant's "shared backing storage" behaviour is unusual and may surprise consumers retaining results across the lifetime of large inputs.

### [Important] `damerau_full.go` rune path uses `map[rune]int` — potentially non-deterministic if accessed for output
- **File:** `/Users/johnny/Development/fuzzymatch/damerau_full.go:362-399`
- **Phase introduced:** Phase 2
- **Issue:** `da := make(map[rune]int)` is used for the rune-path last-occurrence table. Inside the DP loop only `l := da[rb[j-1]]` (point lookup) and `da[ra[i-1]] = i` (point write) — never range-iterated. So DET-03 is satisfied. BUT: an iteration here would silently break determinism without raising any compile-time error. A future refactor that adds e.g. debug logging via `for k, v := range da` would corrupt the output. Worth a defensive comment.
- **Standard:** determinism-standards §"The No-Map-Iteration Rule"
- **Action:** The existing inline comment at line 361 says "Map LOOKUP only — not iterated to produce output (DET-03)." — this is good. No change needed but flagged for future-proofing.
- **Rationale:** Defensive review.

---

## IMPROVEMENT — minor cleanups, naming, godoc polish

### [Improvement] `algoid.go` `AlgoIDs()` allocates a 23-element slice on every call
- **File:** `/Users/johnny/Development/fuzzymatch/algoid.go:282-308`
- **Phase introduced:** Phase 1
- **Issue:** Each call to `AlgoIDs()` allocates a fresh 23-element slice. The Scorer uses it during `NewScorer` (one allocation per scorer construction), `Algorithms()` doesn't but iterates `algorithmsAlgoIDSorted` directly. For hot-path consumers calling `AlgoIDs()` repeatedly, this is unnecessary.
- **Action:** Consider returning a `[]AlgoID` that's pre-allocated as a package-level `var` and document immutability — but this contradicts the "freshly allocated so the caller may freely mutate" contract. Alternative: keep current behaviour and document the alloc cost.
- **Rationale:** Minor perf; not on the hot path.

### [Improvement] `Tokenise` empty input returns `[]string{}` not nil — inconsistent with idiomatic Go
- **File:** `/Users/johnny/Development/fuzzymatch/tokenise.go:143-146`
- **Phase introduced:** Phase 1
- **Issue:** Empty input returns a non-nil empty slice. Go idiom is usually nil. The godoc explicitly says "Returned slice is never nil" — a deliberate API decision but unusual.
- **Action:** Document the rationale for "never nil" — typical reason is "consumers can safely `range` without nil-check" but Go's range over nil slice is a no-op anyway.
- **Rationale:** API ergonomics; not blocking.

### [Improvement] `Soundex` first-letter scan loop has subtle control flow
- **File:** `/Users/johnny/Development/fuzzymatch/soundex.go:160-181`
- **Phase introduced:** Phase 7
- **Issue:** The for-loop body has multiple paths: ASCII letter sets `found = true`; non-letter increments `i`; non-ASCII increments by rune size. After each iteration `if found { break }`. The structure is correct but hard to follow. Could be simplified to two nested loops or factored into a helper.
- **Action:** Refactor into a helper `findFirstASCIILetter(s string) (b byte, nextIdx int, found bool)` — reduces visual complexity at the call site.
- **Rationale:** Readability.

### [Improvement] `SoundexCode` digit counter naming
- **File:** `/Users/johnny/Development/fuzzymatch/soundex.go:154`
- **Phase introduced:** Phase 7
- **Issue:** `digits := 1` — but `digits` starts at 1 because position 0 is reserved for the first letter, not a digit. The variable counts WRITE POSITIONS in the result, not digit characters. Naming should be `writePos` or `nextOut`.
- **Action:** Rename to `nextOut` or `writePos`.
- **Rationale:** Naming clarity.

### [Improvement] `NYSIISCode` step 7 trailing-A removal happens after step 6 which already converted AY→Y — redundant in practice
- **File:** `/Users/johnny/Development/fuzzymatch/nysiis.go:307-316`
- **Phase introduced:** Phase 7
- **Issue:** Step 6 converts `AY` → `Y` (last 2 chars). Step 7 removes a trailing A. These cannot fire together — after step 6, the trailing char is Y, not A. The two checks are independent (a single trailing A without leading A still hits step 7). Code is correct but the proximity might suggest interplay.
- **Action:** Add a one-line comment to step 7 clarifying "fires when step 6 did NOT fire" — i.e. trailing non-A non-AY-suffixed A.
- **Rationale:** Reviewer aid.

### [Improvement] `damerauOSADP` documentation references "linevels" rather than rows
- **File:** `/Users/johnny/Development/fuzzymatch/damerau_osa.go:200-256`
- **Phase introduced:** Phase 2
- **Issue:** Variables are `prevprev`, `prev`, `curr`. The DP "three-row rolling" terminology in the godoc is clear. No issue, just noting that the rotation `prevprev, prev, curr = prev, curr, prevprev` is non-obvious — the comment "After this: the row we just computed (curr) becomes prev, the previous prev becomes prevprev, and the old prevprev is handed back as the new curr (to be overwritten)" is good and should be preserved.
- **Action:** None; well-documented.
- **Rationale:** N/A.

### [Improvement] `Jaro` constant `maxJaroStackLen = 256` differs from Levenshtein's `maxStackInputLen = 64`
- **File:** `/Users/johnny/Development/fuzzymatch/jaro.go:107` vs `/Users/johnny/Development/fuzzymatch/levenshtein.go:68`
- **Phase introduced:** Phase 2
- **Issue:** Two distinct stack-buffer thresholds. Jaro's match-flag arrays are bool (256 × 1 byte = 256 bytes per side); Levenshtein's DP rows are int (65 × 8 bytes × 2 rows = 1040 bytes). Both fit comfortably on the stack; the divergence in threshold is correct but the rationale should be in a comment at the package-level (e.g. doc.go) so reviewers see the design discipline.
- **Action:** Add a paragraph to doc.go or a new `internal_constants.go` explaining the stack-buffer threshold philosophy.
- **Rationale:** Reviewer aid for future algorithm additions.

### [Improvement] `Tversky` panic message uses lowercase "tversky" without article
- **File:** `/Users/johnny/Development/fuzzymatch/tversky.go:242,284` and `/Users/johnny/Development/fuzzymatch/errors.go:88`
- **Phase introduced:** Phase 5
- **Issue:** `panic("fuzzymatch: invalid tversky parameter")` — "tversky" is a proper noun (Amos Tversky). Style convention is either "Tversky" or document the lowercasing.
- **Action:** Change to "fuzzymatch: invalid Tversky parameter" — but this requires updating the sentinel error message which is part of the v1.x contract. Defer until a major version bump.
- **Rationale:** Cosmetic; flag for future major release.

### [Improvement] `partialRatioRegion2Bytes` early-exit `best == 1.0` comparison
- **File:** `/Users/johnny/Development/fuzzymatch/partial_ratio.go:386-389`
- **Phase introduced:** Phase 6
- **Issue:** `if best == 1.0 { return best, true }` — exact float equality. `indelRatio` can return exactly 1.0 when `2*lcs == la+lb` (e.g. identity), so the comparison is safe. But the comment doesn't explain why exact comparison is OK here.
- **Action:** Add inline comment: "best == 1.0 is exact IEEE-754 because indelRatio returns 2*lcs/(la+lb) which is exact when 2*lcs == la+lb (identity case)."
- **Rationale:** Reviewer aid; defending against future "should this be `best >= 1.0 - epsilon`?" debates.

### [Improvement] `partial_ratio.go` Region 2 / Region 3 overlap at i = n-m
- **File:** `/Users/johnny/Development/fuzzymatch/partial_ratio.go:395-411`
- **Phase introduced:** Phase 6
- **Issue:** When `n > m`, Region 2 evaluates `i = n-m` AND Region 3 also evaluates `i = n-m` (since Region 3 starts at `i = n-m`). The same alignment is computed twice. The file header (line 88-90) acknowledges this: "When n > m there is a single trivial overlap at i = n-m with Region 2 (one redundant indelRatio call, harmless and matches the RapidFuzz reference behaviour)." Acceptable, just noting.
- **Action:** None; documented.
- **Rationale:** N/A.

### [Improvement] `cosine.go` upper clamp comment is excellent but the lower clamp comment is buried
- **File:** `/Users/johnny/Development/fuzzymatch/cosine.go:381-389`
- **Phase introduced:** Phase 5
- **Issue:** The upper-clamp comment (lines 372-381) is detailed (ULP overshoot rationale). The lower-clamp comment (lines 381-384) says "theoretically unreachable but the clamp costs nothing" — fine but less detailed than the upper.
- **Action:** None; cosmetic.
- **Rationale:** Style.

### [Improvement] `monge_elkan.go` allow-list `permittedMongeElkanInner` uses a map for lookup but exhaustive panic test walks all AlgoIDs
- **File:** `/Users/johnny/Development/fuzzymatch/monge_elkan.go:291-317`
- **Phase introduced:** Phase 6
- **Issue:** `var permittedMongeElkanInner = map[AlgoID]bool{...}` — map literal with 18 true entries. For 23 total AlgoIDs, a `[23]bool` array indexed by `int(AlgoID)` would be denser, zero-allocation, and inherently bounds-checked.
- **Action:** Consider switching to `var permittedMongeElkanInner = [numAlgorithms]bool{ AlgoLevenshtein: true, ... }`. The lookup site `if !permittedMongeElkanInner[inner]` would need a bounds check.
- **Rationale:** Minor perf; declarative-data improvement.

### [Improvement] `scorer.go` panic message in `DefaultScorer` includes error text — could simplify to `panic("...")` + err
- **File:** `/Users/johnny/Development/fuzzymatch/scorer.go:589`
- **Phase introduced:** Phase 8
- **Issue:** `panic("fuzzymatch: DefaultScorer construction failed (this is a bug): " + err.Error())` — string concatenation in a panic. Idiomatic Go would use `fmt.Errorf` wrapping or `log.Panicf`.
- **Action:** Consider `panic(fmt.Errorf("fuzzymatch: DefaultScorer construction failed: %w", err))` — panic value is the error chain, recoverable downstream.
- **Rationale:** Idiomatic Go.

### [Improvement] `errors.go` sentinel godoc references "Phase 9" (scan) which doesn't exist yet
- **File:** `/Users/johnny/Development/fuzzymatch/errors.go:27` ("richer per-item context can be added in Phase 9 if scan needs it")
- **Phase introduced:** Phase 1
- **Issue:** Phase 9 is the documented future phase; the comment is forward-looking. No issue, just noting the dependency.
- **Action:** None; documented intent.
- **Rationale:** N/A.

### [Improvement] `tokenise.go` `lowerRuneToken` duplicates `lowerRune` in normalise.go
- **File:** `/Users/johnny/Development/fuzzymatch/tokenise.go:340-354` vs `/Users/johnny/Development/fuzzymatch/normalise.go:441-452`
- **Phase introduced:** Phase 1
- **Issue:** Two identical functions named differently for "no internal dependency on normalise.go" per tokenise.go's comment. The duplication is deliberate but introduces a maintenance burden (if Unicode handling changes, both must update). Compare with the strcmp95 / soundex `runeAt` duplication which IS shared.
- **Action:** Consider factoring into an `internal_ascii.go` file with a single shared `asciiToLowerRune` helper. The "no internal dependency" rule isn't enforced for shared primitives.
- **Rationale:** DRY.

### [Improvement] `algoid.go` `String()` uses a 23-case switch — could be a literal lookup
- **File:** `/Users/johnny/Development/fuzzymatch/algoid.go:213-267`
- **Phase introduced:** Phase 1
- **Issue:** The 23-case switch is correct and clearly documented (the nolint:gocyclo defense is good). An alternative is a package-level `var algoIDStrings = [numAlgorithms]string{...}` with O(1) lookup. The current switch is the canonical Go idiom for stringly-typed enums, but the lookup table would be slightly faster on the hot path (if `String()` is ever called from a hot path, which it shouldn't be).
- **Action:** None; the switch is appropriate for non-hot-path stringification.
- **Rationale:** Reviewer style preference.

### [Improvement] `scorer.go` `Scorer.Score` reduction loop comment block is dense
- **File:** `/Users/johnny/Development/fuzzymatch/scorer.go:349-383`
- **Phase introduced:** Phase 8
- **Issue:** The comment in `Score` (lines 359-380) is ~20 lines of dense determinism rationale for an 11-line function body. The FMA-fusion remediation comment is informative but could be in a separate `// See …` reference rather than inline.
- **Action:** Consider extracting the FMA-fusion documentation into a `determinism.md` doc file or a comment in scorer.go's header (which already has it), reducing the per-call-site verbosity.
- **Rationale:** Readability; the inline density makes the code body harder to scan.

### [Improvement] `mra.go` `mraThreshold` uses a 13-element array with sum>12 clamp at function boundary
- **File:** `/Users/johnny/Development/fuzzymatch/mra.go:95-109`
- **Phase introduced:** Phase 7
- **Issue:** `mraThresholdTable` is sized for `sum in [0, 12]`. For `sum > 12`, the function returns 2 (clamp) without indexing the table. Clean design, but the table size 13 and the clamp boundary 12 are magic numbers; the file comment explains them but a named constant `mraSumLimit = 12` would improve clarity.
- **Action:** Add `const mraSumLimit = 12 // upper bound of the NBS-943 Table A`.
- **Rationale:** Magic-number elimination.

### [Improvement] `normalise.go` `applyUnicodeTransformer` constructs a fresh `transform.Chain` per call
- **File:** `/Users/johnny/Development/fuzzymatch/normalise.go:320-338`
- **Phase introduced:** Phase 1
- **Issue:** The chain is built fresh per call ("transform.Transformer is not documented as safe for concurrent reuse, and per-call construction is cheap"). For Scorer use, this fires twice per `Score(a, b)`. Profiling might justify a sync.Pool of chains, but the comment defers this.
- **Action:** None; documented as v1.x deferral.
- **Rationale:** N/A.

### [Improvement] `lcsstr.go` `LongestCommonSubstring` returns shared backing storage — could be a documented design or could be defensive-copied
- **File:** `/Users/johnny/Development/fuzzymatch/lcsstr.go:116-155`
- **Phase introduced:** Phase 3
- **Issue:** See "Important" finding above on retained-reference behaviour. As an improvement, consider whether the rune variant's freshly-allocated string (line 187) and the byte variant's sub-slice are semantically equivalent for consumers. The mismatch is documented but unexpected.
- **Action:** Decide between (a) make both variants share backing storage (faster), (b) make both defensive-copy (consistent, safe), or (c) document the difference more prominently.
- **Rationale:** API consistency.

### [Improvement] `swg.go` `SmithWatermanGotohRawScoreRunes` has duplicated identity-short-circuit logic
- **File:** `/Users/johnny/Development/fuzzymatch/swg.go:287-313`
- **Phase introduced:** Phase 3
- **Issue:** The identity branch logic (`if a == "" return 0.0; return Match * float64(len([]rune(a)))`) duplicates the byte path's logic in `SmithWatermanGotohRawScoreWithParams`. The comment (lines 290-295) explains the duplication is deliberate to preserve the identity short-circuit's purpose.
- **Action:** None; documented intent.
- **Rationale:** N/A.

### [Improvement] `dispatch_*.go` documentation duplicates the `var _ = func() bool` idiom rationale in every file
- **File:** All 23 `/Users/johnny/Development/fuzzymatch/dispatch_*.go` files
- **Phase introduced:** Phases 2–7
- **Issue:** Each dispatch file has ~10 lines explaining the `var _ = func() bool { ... }()` idiom. This is 230 lines of duplicated boilerplate documentation. A single reference in algoid.go could replace it.
- **Action:** Consider replacing the per-file boilerplate with a one-line `// See algoid.go for the dispatch-registration pattern rationale.` reference, with the full explanation centralised.
- **Rationale:** DRY; reduce maintenance burden.

### [Improvement] `partial_ratio.go` rune charSet uses `map[rune]struct{}` — could be a small slice
- **File:** `/Users/johnny/Development/fuzzymatch/partial_ratio.go:486-495`
- **Phase introduced:** Phase 6
- **Issue:** `make(map[rune]struct{}, m)` for the rune charSet. For typical inputs `m` is small (< 20 runes), so a sorted `[]rune` with binary search OR a linear scan would be faster than a map (no hash overhead) and zero-allocation on stack.
- **Action:** Profile and consider replacing with a stack-allocated `[20]rune` for short inputs with fallback to a map for longer ones.
- **Rationale:** Perf.

### [Improvement] No `examples/` directory contents reviewed but listed in `ls` output
- **File:** `/Users/johnny/Development/fuzzymatch/examples/` (identifier-similarity, phonetic-keys, scorer-composition)
- **Phase introduced:** Phase 8
- **Issue:** Three sub-examples exist but their compilation status, godoc presence, and test coverage was not validated in this review.
- **Action:** Confirm `make check` includes building these examples; check that each has a `main_test.go` that compiles.
- **Rationale:** Public-facing examples need the same quality bar as the library.

---

## Notes on out-of-scope and related findings

- **scan/ sub-package**: empty/non-existent (Phase 9). All scan-related sentinel godoc references are forward-looking. Not reviewed.
- **`tests/bdd/`**: BDD test code not reviewed in this pass — see `08-BDD-REVIEW.md` for prior phase-8 BDD review.
- **Determinism**: Phase 8 has dedicated determinism review (`08-DETERMINISM-REVIEW.md`). No new determinism issues found in this whole-codebase pass beyond the Cosine intersection-keys comment improvement noted above.
- **License headers**: every reviewed `.go` file has the AxonOps Apache 2.0 header. No issues.
- **Test files (`*_test.go`)**: spot-checked; full review is `test-analyst`'s scope.

## Key recommendations for next phase / cycle

1. **Resolve the Hamming `(int, error)` contract** before v1.0 freeze — option (a) update spec to silent return-max, or (b) introduce ErrHammingLengthMismatch and break API.
2. **Clean up `bench.txt.new`** committed to repo root — either fix the failing property test or delete the artefact with CHANGELOG note.
3. **Tighten Scorer option layer** — add α+β > 0 check to WithTverskyAlgorithm and permittedMongeElkanInner check to WithMongeElkanAlgorithm.
4. **Fix the dispatch-file slot-number drift** — three files have wrong iota numbers; remove the explicit numbers entirely.
5. **Move `WriteGoldenFile` out of the public API** — into a `_test.go` file or `internal/golden` package.

</details>

---

## 8. determinism-reviewer (commit-message-reviewer skipped — see header)

_Source: `.planning/reviews/determinism-FINDINGS.md`_

<details>
<summary>Click to expand full report</summary>

---
status: issues_found
agent: determinism-reviewer
scope: entire codebase (phases 1-8)
reviewed: 2026-05-17T00:00:00+00:00
finding_counts:
  critical: 0
  important: 5
  improvement: 12
  total: 17
---

# Determinism Review — fuzzymatch (Phases 1 through 8)

## Scope

Comprehensive cross-codebase determinism audit of the fuzzymatch root
package as it exists after Phase 8 (composite Scorer) closes. Builds on
`.planning/phases/08-composite-scorer/08-DETERMINISM-REVIEW.md` (which
covered the Phase 8 surface only) and expands to every algorithm, the
dispatch tables, Normalise, Tokenise, q-gram helpers, and the four
golden files. Cross-references the `determinism-standards` skill.

## Method

- Source walk: every production `.go` file in the package root (66 files).
- Grep audit for `init()`, `math.*`, `range map`, `sort.Slice*`,
  `time.Now`, `rand.*`, `os.Getenv`, `runtime.GOOS|GOARCH`, `sync.*`,
  `chan `, `go func`, `slices.Sum`.
- Golden-file inspection for timestamps and non-deterministic content.
- `go test -run TestProp.*Deterministic ./...` — PASS (0.5 s).
- `go test -run TestGolden ./...` — PASS (0.6 s).
- `go test -run TestProp.*(NoNaN|NoInf|Symmetric|Range|Identity) ./...` — PASS (2.9 s).

## Top-Line

**No CRITICAL findings.** No cross-platform byte-identity break has been
introduced anywhere in Phases 1 through 8. The library's determinism
contract — no `init()`, no `math.X` beyond `math.Sqrt`, no map iteration
on output paths, no time/rand/env sniffing, no concurrency primitives,
no parallel reduction — is intact and enforced by passing property +
golden-file tests on the local platform.

The five IMPORTANT findings flag patterns that are CORRECT today but
either (a) deviate from the locked DET-06 contract style without being
wrong, (b) leak documented-as-non-deterministic information through
backdoors that are not yet exercised, or (c) are conditional on the
cross-platform CI matrix running before each release. Twelve
IMPROVEMENT findings are defence-in-depth / style.

The cross-platform CI matrix (linux/{amd64,arm64}, darwin/{amd64,arm64},
windows/amd64) is wired into `.github/workflows/ci.yml` and runs
`make verify-determinism` (which executes `go test -run TestGolden_
./...`) on every PR. The matrix is the load-bearing empirical gate; the
local-only review here cannot certify cross-platform byte-identity by
itself.

---

## Verification Checklist Results

| Check | Result |
|---|---|
| No `init()` doing non-trivial work | PASS — zero `func init()` in the package |
| No map iteration on output paths | PASS — every `range map` reduces to a scalar or feeds an explicitly-sorted slice |
| No `math.Pow/Log/Exp/FMA` | PASS — only `math.Sqrt` (Cosine), `math.IsNaN`/`math.IsInf` (test code) |
| Float reduction left-to-right | PASS — Scorer reduction, Cosine dot-product, weight normalisation all explicit `(x*y)+z` parens |
| No `slices.Sum` / parallel reduction | PASS — none used |
| NaN/Inf/-0 guarded on every public path | PASS — every `/maxLen` and `/m` has an explicit `== 0` guard preceding it |
| Sort-key completeness | MOSTLY PASS — see DET-W1 (Algorithms_Merge uses sort.Slice not SliceStable with single-key Name) |
| Golden file format | PASS — single trailing LF, no BOM, no timestamps in scorer/algorithms/normalisation; only static literal in phonetic |
| Cross-platform CI matrix wired in | PASS — `.github/workflows/ci.yml` runs `make verify-determinism` on 5 platforms |
| AlgoID enum order stability | PASS — `AlgoIDs()` returns a hand-written slice literal; algoid_test.go pins count=23 and stable order |
| Tokenise stable output ordering | PASS — left-to-right rune walk, no map, no goroutines |
| Normalise NFC/NFD selection consistency | PASS — switch in `applyUnicodeTransformer` is total over `(StripDiacritics, NFC)` cross-product |
| Per-algorithm: no hidden RNG / time.Now / env sniffing | PASS — zero occurrences |
| Concurrent-use determinism | PASS — Scorer is immutable after construction; no globals are written after package init |

---

## Findings

### [IMPORTANT] DET-W1: `TestGolden_Algorithms_Merge` uses `sort.Slice` not `sort.SliceStable` with a single-key sort
- **File:** algorithms_golden_test.go:200-202, 222-224, 275-277
- **Phase introduced:** Phase 2 (Wave 3 merge)
- **Issue:** The merge / per-algorithm-staging tests sort entries by `Name` using `sort.Slice` (unstable). The sort key is *only* `Name`. The merge test has a defensive duplicate-name check immediately after the sort, but the per-algorithm staging tests do not. If two entries shared a `Name` (a bug — but not impossible during regen), the on-disk ordering would be non-deterministic across runs, breaking the byte-identity golden.
- **Standard:** `.claude/skills/determinism-standards/SKILL.md` §"Sort Key Completeness" — "When sorting output, use sort.SliceStable AND ensure the key is complete (or document why ties don't occur and assert it in a test)."
- **Action:** Code fix — swap `sort.Slice` for `sort.SliceStable` in algorithms_golden_test.go:200, :222, :275, and every other site in the file. The duplicate-name sanity check in the merge function should be lifted into a shared helper and applied in every staging test too.
- **Rationale:** The current pattern is correct today (no duplicate Names exist by construction), but stable sort defends against a future regression introducing a duplicate. Cost is zero (sort.SliceStable's performance is identical for small N).
- **Suggested fix:**
  ```go
  sort.SliceStable(allEntries, func(i, j int) bool {
      return allEntries[i].Name < allEntries[j].Name
  })
  ```
  Repeat in every `sort.Slice(...by Name...)` call in algorithms_golden_test.go.

### [IMPORTANT] DET-W2: Cosine FMA-fusion remediation is documented but not applied; the `(x*y)+z` dot-product pattern is the LOAD-BEARING cross-platform float-determinism algorithm
- **File:** cosine.go:343
- **Phase introduced:** Phase 5
- **Issue:** The Cosine dot-product reduction is `dot = (float64(qa[k]) * float64(qb[k])) + dot`. Per cosine.go:288-297 own godoc and the explicit warning in scorer.go:55-61, the Go compiler MAY emit a fused multiply-add (FMA) instruction on arm64 for the `(x*y)+z` pattern (per golang/go#17895, parens do NOT defeat FMA fusion). The empirical claim is that the integer-derived `qa[k] * qb[k]` products stay below the ULP threshold where FMA-vs-non-FMA divergence would surface. Today the cross-platform golden gate PASSES on the four-platform matrix run, validating this claim. But the remediation pattern (`dot = float64(float64(qa[k]) * float64(qb[k])) + dot`) is NOT applied — only documented.
- **Standard:** `.claude/skills/determinism-standards/SKILL.md` §"Float Stability" — "No platform-specific intrinsics (math.FMA etc.) — even where they would be faster." The compiler-emitted FMA from a `(x*y)+z` source pattern is the same hazard category.
- **Action:** Discuss-phase needed — option A (status quo, monitor matrix); option B (preemptively apply the double-cast).
- **Rationale:** Latent risk against a future Cosine input where the inner-product magnitude grows to the float64 boundary. The same risk applies to scorer.go:380 (`acc = acc + (entry.weight * score)`) — both sites are the LOAD-BEARING gate. If a future algorithm produces a score whose weighted contribution crosses the ULP threshold, the unmitigated `(weight * score)` would diverge between arm64 (FMA) and amd64 (non-FMA).
- **Suggested fix (Option B, defensive):**
  ```go
  // cosine.go:343
  dot = float64(float64(qa[k]) * float64(qb[k])) + dot
  // scorer.go:380
  acc = acc + float64(float64(entry.weight) * float64(score))
  ```
  Each `float64(...)` cast forces a single rounding step, defeating FMA fusion. Cost: ~1ns per iteration; benign.

### [IMPORTANT] DET-W3: `WithThreshold` does not guard `NaN` — admits non-finite into `*Scorer.threshold` (already CR-01 in 08-REVIEW.md; recorded here from the determinism perspective)
- **File:** scorer_options.go:257-266
- **Phase introduced:** Phase 8
- **Issue:** `WithThreshold(t)` checks `t < 0.0 || t > 1.0`. Both comparisons evaluate to `false` for `NaN`. NaN passes the gate, propagates to `Scorer.threshold`, and `Match` becomes "never match" for every input (IEEE-754: `x >= NaN` is always `false`). The result is deterministic-but-wrong, so the cross-platform golden gate does NOT catch it. But it violates the project-wide NaN/Inf policy.
- **Standard:** `.claude/skills/determinism-standards/SKILL.md` §"NaN / Inf / Negative Zero" — "No algorithm produces NaN, +Inf, or -Inf. Any path that could (e.g. division by zero in normalisation) MUST be guarded with an explicit edge case."
- **Action:** Code fix — already tracked as CR-01 in 08-REVIEW.md. Add `math.IsNaN(t)` to the gate.
- **Rationale:** Defence-in-depth; closes the NaN escape on the Scorer's surface boundary.
- **Suggested fix:**
  ```go
  if math.IsNaN(t) || t < 0.0 || t > 1.0 {
      return ErrInvalidThreshold
  }
  ```

### [IMPORTANT] DET-W4: Scorer reduction trusts dispatched scoreFn return values — no defensive NaN/Inf guard
- **File:** scorer.go:368-382, scorer.go:512-517
- **Phase introduced:** Phase 8
- **Issue:** Both `Score` and `ScoreAll` trust `entry.scoreFn(na, nb)` to return a finite float in `[0, 1]`. All 23 catalogue algorithms satisfy this contract today, verified by per-algorithm `TestProp_<algo>_NoNaN_NoInf` and the Phase-8-level `TestProp_Scorer_NoNaN_NoInf`. The fragility is that a future regression in any algorithm, OR a parameterised custom closure (e.g. `WithSmithWatermanGotohAlgorithm` with pathological `SWGParams` whose godoc warns "nonsense values produce a deterministic-but-meaningless score" but does NOT promise finiteness), could leak NaN into the reduction. NaN poisons `acc` via `acc + (w * NaN) = NaN` and the bug surfaces as "Match always false; ScoreAll returns a map with NaN values" — deterministic, so cross-platform CI doesn't catch it.
- **Standard:** `.claude/skills/determinism-standards/SKILL.md` §"NaN / Inf / Negative Zero" — same as DET-W3 but on the consumer side.
- **Action:** Code fix or skill-clarification — choose between (a) leave as-is (current tests catch every known producer), (b) add a per-iteration `math.IsNaN`/`math.IsInf` check that panics with a programmer-error message, (c) add a `ScoreErr` method that returns the error path. Recommendation: (b) for v1.x; (c) would need api-ergonomics-reviewer sign-off.
- **Rationale:** Latent; not a current bug. Adding the panic guard costs ~1ns per iteration and surfaces algorithm-layer regressions at the earliest possible point.

### [IMPORTANT] DET-W5: `WithoutAlgorithm` in-place compaction `cfg.entries[:0]` aliases the underlying array
- **File:** scorer_options.go:184-199
- **Phase introduced:** Phase 8
- **Issue:** The `filtered := cfg.entries[:0]` pattern reuses the backing array. The algorithm is correct (writes only after reads), but `cfg.entries` retains old-value capacity beyond the new length. If a future maintainer ever returns `cfg.entries` to consumer code, or if a subsequent option capture-by-reference re-reads spare-capacity slots, they'd see stale values from before the removal. For determinism specifically: today the value-typed slice header is never externalised, so no bug surfaces. The fragility is latent.
- **Standard:** No governing skill clause; falls under general "no shared mutable state escapes".
- **Action:** Code fix or accept-with-comment — change to `filtered := make([]scorerEntry, 0, len(cfg.entries))`. Cost: one allocation per `WithoutAlgorithm` at construction time (called rarely, often once).
- **Rationale:** Defence in depth; current pattern is the canonical Go idiom but the godoc comment in `WithoutAlgorithm` (scorer_options.go:184) already documents linear-scan-and-compact semantics, so a fresh allocation matches the documented intent.

---

### [IMPROVEMENT] DET-W6: phonetic-codes.json contains a static `regenerated_at` literal in `_metadata`
- **File:** testdata/golden/phonetic-codes.json:4
- **Phase introduced:** Phase 7
- **Issue:** `"_metadata": { "purpose": "...", "regenerated_at": "2026-05-15T00:00:00+00:00" }`. The string is a static literal (the file is hand-edited, not regenerated from `time.Now()`), so it does NOT break determinism today. But the *appearance* of a timestamp field is a magnet for a future maintainer to convert it to a real-time `time.Now()` value during a regen, which WOULD break the cross-platform byte-identity gate. The other three goldens (algorithms.json, normalisation.json, scorer-default.json) deliberately avoid timestamp fields per 08-DETERMINISM-REVIEW.md DET-07.
- **Standard:** `.claude/skills/determinism-standards/SKILL.md` §"Golden Files" — "Golden files are JSON for human-readability and diff-friendliness. They are regenerated only when the algorithm output legitimately changes (e.g. bug fix)."
- **Action:** Code fix (rename) OR skill-clarification (document the static-literal convention explicitly).
- **Rationale:** Hygiene; matches the scorer-default.json / algorithms.json / normalisation.json pattern. The field is purely informational and the value never actually changes runtime-to-runtime.
- **Suggested fix:** Rename `regenerated_at` to `regenerated_on` or `last_curated` (or drop it entirely; the git log carries the same information). Add a comment in `phonetic_codes_golden_test.go` near the type declaration: "// RegeneratedAt is a STATIC string literal in the golden file, not a runtime-generated timestamp; do NOT replace with time.Now() during regen — that would break the cross-platform byte-identity gate."

### [IMPROVEMENT] DET-W7: scorer_signature is a static literal embedded in test code; drift on DefaultScorer composition change is silent
- **File:** scorer_golden_test.go (search for "DefaultScorer-2026-05-16")
- **Phase introduced:** Phase 8
- **Issue:** Already DET-07 in the Phase 8 standalone review. The string `"DefaultScorer-2026-05-16"` is hard-coded with a construction date. Any future change to DefaultScorer's six-algorithm composition + 0.85 threshold requires hand-updating this constant. Forgetting to update doesn't break determinism (byte-identity still passes if the string is unchanged) but it does mean the metadata lies about which composition the golden represents.
- **Standard:** No governing skill clause; falls under general "golden-file curation discipline".
- **Action:** Skill clarification — add a release-prep checklist item to `docs/scorer.md` or the docs/extending.md release runbook: "If DefaultScorer composition changes, bump scorer_signature in scorer_golden_test.go."
- **Rationale:** Process-level; no code change.

### [IMPROVEMENT] DET-W8: Strcmp95 similar-character table is `[...]struct{...}` literal — good, but the linear scan iteration order is implicit
- **File:** strcmp95.go:133-173 (the table), strcmp95.go:198-206 (the lookup)
- **Phase introduced:** Phase 3 (likely; the file's Phase introduction not explicitly stated but pre-Phase 5)
- **Issue:** The 36-entry similar-character table is declared via `var ... = [...]struct{...}{...}` literal — IDEAL pattern per the no-init rule. The `strcmp95SimilarLookup` function does a linear scan; the iteration order is the declared order of the literal. This is deterministic by construction. The minor friction is that the determinism contract relies implicitly on Go's spec guarantee that array literals iterate in declared order — there's no in-code assertion of this.
- **Standard:** `.claude/skills/determinism-standards/SKILL.md` §"Init() and Package Loading" — already SATISFIED. This is a hygiene note.
- **Action:** No code change; document the dependency. Could be added in commit message or a comment near the `for i := range strcmp95SimilarChars` loop.
- **Rationale:** Defensive documentation only.

### [IMPROVEMENT] DET-W9: `var _ = func() bool { ... }()` package-level dispatch registrations are init() in disguise
- **File:** dispatch_*.go (23 files, one per algorithm)
- **Phase introduced:** Phase 2 onward
- **Issue:** Every dispatch_X.go uses the `var _ = func() bool { dispatch[AlgoX] = X; return true }()` idiom. This runs at package load time, same as `init()`. The skill's "no init() doing non-trivial work" rule is intended to flag non-trivial init — the dispatch registration is trivial (single assignment) and writes to a UNIQUE slot per file, so order between dispatch_X.go files is irrelevant. This is the documented "init-alternative" pattern (see dispatch_double_metaphone.go file header and others). No bug.
- **Standard:** `.claude/skills/determinism-standards/SKILL.md` §"Init() and Package Loading" — "The library has no init() functions doing non-trivial work. Tables that require initialisation ... are declared via `var x = ...` literal expressions, not built in `init()`."
- **Action:** Skill clarification — the rule should explicitly carve out the `var _ = func() bool {...}()` idiom as permitted for trivial dispatch-table population (one assignment, no map iteration, no cross-file dependency). OR change the idiom to `var dispatchCosine = registerDispatch(AlgoCosine, func(a, b string) float64 {...})` returning the input value — same effect, more obviously a `var` literal.
- **Rationale:** The pattern is correct and idiomatic; the skill text just doesn't explicitly call out the exception, creating a documentation gap that a strict reading would flag as a near-violation.

### [IMPROVEMENT] DET-W10: Monge-Elkan reduction uses `+=` while Scorer uses explicit `acc = acc + (x * y)` — consistency gap
- **File:** monge_elkan.go:426 (`sumOfMax += maxSim`)
- **Phase introduced:** Phase 6
- **Issue:** Scorer.Score (scorer.go:380) uses `acc = acc + (entry.weight * score)` — the locked DET-06 pattern. Monge-Elkan's inner reduction uses `sumOfMax += maxSim`. The two are observationally equivalent (compound assignment desugars to `x = x + y`); both have no FMA hazard because there's no multiplication-then-addition fused in this reduction (multiplications happen inside `innerFn`, results land in `s`/`maxSim`, then a pure addition accumulates). No bug.
- **Standard:** `.claude/skills/determinism-standards/SKILL.md` §"Float Stability" — implicit consistency expectation; the DET-06 locked pattern is `x = x + (y * z)`.
- **Action:** Style fix — rewrite as `sumOfMax = sumOfMax + maxSim` for consistency with Scorer / Cosine / weight-normalisation patterns. OR document why Monge-Elkan can use `+=` (no multiply-then-add → no FMA hazard).
- **Rationale:** Reviewer ergonomics; uniform pattern across the codebase makes "find every reduction" greps easier.

### [IMPROVEMENT] DET-W11: JaroWinkler return expression `j + float64(l)*winklerPrefixScale*(1.0-j)` has implicit left-to-right multiplication chain
- **File:** jarowinkler.go:143, jarowinkler.go:184
- **Phase introduced:** Phase 2
- **Issue:** The expression parses as `j + ((float64(l)*winklerPrefixScale)*(1.0-j))` per Go's left-to-right associativity. This is correct and deterministic; multiplication is associative under IEEE-754 round-to-nearest for short chains (three operands) only IF each intermediate result is rounded to float64 — which Go guarantees absent FMA. Subject to the same FMA-fusion caveat as DET-W2. The DET-06 pattern in scorer.go uses fully explicit parens; this site does not.
- **Standard:** `.claude/skills/determinism-standards/SKILL.md` §"Float Stability".
- **Action:** Style fix — add explicit parens: `j + ((float64(l) * winklerPrefixScale) * (1.0 - j))`. Or, for full FMA-defensive form: `j + float64(float64(l)*winklerPrefixScale) * (1.0 - j)`.
- **Rationale:** Hygiene; matches Scorer's `(x*y)+z` form. Today the matrix passes, so no urgent action.

### [IMPROVEMENT] DET-W12: Cosine's `cos = dot / (normA * normB)` uses an unparenthesised division-by-product
- **File:** cosine.go:370
- **Phase introduced:** Phase 5
- **Issue:** `cos := dot / (normA * normB)`. The parens are present around the divisor. The hazard is that `normA * normB` is a multiplication whose ordering matters. As both operands come from `math.Sqrt(...)` calls on int-derived inputs, both are IEEE-754 correctly rounded — `normA * normB` is a single rounded multiplication and deterministic.
- **Standard:** `.claude/skills/determinism-standards/SKILL.md` §"Float Stability".
- **Action:** No code change; already conformant. Recorded for completeness.
- **Rationale:** Confirms the float-determinism contract in the most arithmetic-dense site.

### [IMPROVEMENT] DET-W13: `cosineFromQGramMaps` map iteration of `qa` and `qb` for sum-of-squares uses any-order summation — correct for integers, but the comment could be sharper
- **File:** cosine.go:351-356
- **Phase introduced:** Phase 5
- **Issue:** `for _, c := range qa { sumSquaresA += c * c }` iterates a map (DET-03 candidate). The comment correctly notes "integer addition is exactly associative". This is true and correct: integer multiplication `c * c` and integer addition `+= c*c` produce the same `sumSquaresA` regardless of iteration order, because the operations are over `int` (not `float64`), and on 64-bit Go platforms `int` does not overflow on realistic q-gram counts (`c` is bounded by input length squared; `int64` overflow requires inputs of `2^31` characters).
- **Standard:** `.claude/skills/determinism-standards/SKILL.md` §"The No-Map-Iteration Rule" — output paths must not depend on map iteration. The output here is a scalar `int` — not an output path in the DET-03 sense.
- **Action:** No code change. Optionally tighten the comment to assert int (not float) operands.
- **Rationale:** Reviewer audit trail; helps a future reader confirm at a glance that the loop is safe.

### [IMPROVEMENT] DET-W14: Jaccard / Sørensen-Dice / Tversky map iteration for `totalA` / `totalB` mirrors Cosine — same correctness reasoning
- **File:** qgram_jaccard.go:221-227, sorensen_dice.go:239-246, tversky.go:346-352
- **Phase introduced:** Phase 5
- **Issue:** Three q-gram algorithms iterate `qa` and `qb` maps to sum integer multiset cardinalities. Same int-addition-is-associative reasoning as DET-W13. The intersection iteration walks the smaller-of-(qa, qb) map; the output is a scalar `int` (`intersection`), not an ordered slice. Per DET-03 this is permitted.
- **Standard:** Same as DET-W13.
- **Action:** No code change. Confirmation only.
- **Rationale:** Confirms three independent algorithms apply the same safe pattern.

### [IMPROVEMENT] DET-W15: `Tokenise` uses `[]rune(s)` conversion which silently replaces invalid UTF-8 with U+FFFD — output deterministic but lossy
- **File:** tokenise.go:161
- **Phase introduced:** Phase 1 (Tokenise primitive)
- **Issue:** `runes := []rune(s)` — Go's stdlib converts invalid UTF-8 bytes to U+FFFD individually. The conversion is deterministic across platforms (the Go spec pins it). The conversion *is* documented in tokenise.go's godoc (line 80-82, 158-160). No bug.
- **Standard:** `.claude/skills/determinism-standards/SKILL.md` does not specifically address UTF-8 replacement, but the consequence (byte-identical output for the same input) is preserved.
- **Action:** No code change. Confirmation only.
- **Rationale:** This is the right call (the alternative — error on invalid UTF-8 — would propagate an error type through every algorithm signature, breaking the pure-function contract).

### [IMPROVEMENT] DET-W16: Normalise's `applyUnicodeTransformer` uses `golang.org/x/text/...` whose internal tables come from a separate module — cross-platform determinism depends on that module's stability
- **File:** normalise.go:320-338
- **Phase introduced:** Phase 1 (Normalise primitive)
- **Issue:** The Unicode-path normalisation pipeline depends on `golang.org/x/text/unicode/norm`, `golang.org/x/text/runes`, and `golang.org/x/text/transform`. The dependency is locked in `go.mod`. Different versions of x/text COULD theoretically produce different NFC output for the same input (e.g. if the Unicode standard added a new precomposed form). Cross-platform determinism on a fixed x/text version is guaranteed (the tables are baked into the module); cross-PATCH-VERSION stability of fuzzymatch depends on x/text not changing its NFC output between fuzzymatch's released x/text version and the one consumers `go mod tidy` to. Today x/text is rev-locked in go.mod; Dependabot may bump it.
- **Standard:** `.claude/skills/determinism-standards/SKILL.md` §"Cross-Platform CI Matrix" — gates per-PR; doesn't explicitly cover dependency-version-bump determinism risk.
- **Action:** Skill clarification — add a paragraph to `.claude/skills/determinism-standards/SKILL.md` covering the x/text Unicode-table version risk; recommend pinning x/text via go.mod minor version (the current approach) and re-running the cross-platform matrix on every Dependabot bump.
- **Rationale:** Defensive process documentation. The current Dependabot grouping (test-only vs runtime) already isolates this risk; explicit acknowledgement closes the gap.

### [IMPROVEMENT] DET-W17: `Score` and `ScoreAll` apply Normalise twice — once for each input — per call; no shared cache
- **File:** scorer.go:354-357, scorer.go:502-506
- **Phase introduced:** Phase 8
- **Issue:** Each `Score(a, b)` call invokes `Normalise(a, opts)` and `Normalise(b, opts)` once. Each Normalise call may invoke the x/text transformer (Unicode path) or the byte-pass (ASCII path). The pipeline is deterministic — the same input produces the same output on the same x/text version. No determinism bug. The note here is that the per-call construction of `transform.Chain(...)` inside `applyUnicodeTransformer` (normalise.go:327) re-builds the chain each call — `transform.Transformer` is not documented as safe for concurrent reuse, so per-call construction is the safe choice. The cost is small and the alternative (a `sync.Pool`) would violate D-09 ("no mutexes, no atomics, no pools").
- **Standard:** `.claude/skills/determinism-standards/SKILL.md` §"Mutex-Free" — currently SATISFIED.
- **Action:** No code change. Confirmation only.
- **Rationale:** Re-affirms that the no-pool / no-shared-mutable-state discipline is consistent with the per-call transformer build.

---

## Cross-Platform CI Prerequisite

The cross-platform CI matrix (`.github/workflows/ci.yml`) runs
`make verify-determinism` (which executes `go test -run TestGolden_
./...`) on every PR across:
- linux/amd64 (`ubuntu-latest`)
- linux/arm64 (`ubuntu-24.04-arm`)
- darwin/amd64 (`macos-15-intel`)
- darwin/arm64 (`macos-latest`)
- windows/amd64 (`windows-latest`)

with `CGO_ENABLED=0` workflow-level enforcement. This is the canonical
cross-platform byte-identity gate. Per `.planning/phases/08-composite-
scorer/08-DETERMINISM-REVIEW.md` "Cross-Platform CI Prerequisite", the
Phase 8 Scorer golden test (`TestGolden_ScorerDefault`) runs through
the same `assertGolden` plumbing as the Phase 2-7 algorithm goldens,
so the matrix gate covers all four golden files
(algorithms.json, normalisation.json, phonetic-codes.json,
scorer-default.json) uniformly.

The local-only review here CANNOT certify cross-platform byte-identity
by itself. If a new algorithm or a `(x*y)+z` reduction site is added,
the matrix run on the introducing PR is the only place the FMA-fusion
divergence (DET-W2) would surface.

---

## GO / NO-GO

**GO** on cross-platform determinism for the codebase as it exists after
Phase 8. The library carries the LOCKED determinism contract correctly
across all 23 algorithms, two primitives (Normalise, Tokenise), the
composite Scorer, and the four golden files. No CRITICAL findings on
cross-platform byte-identity.

Five IMPORTANT findings (DET-W1 through DET-W5) flag patterns that
should be hardened before v1.0 but do NOT break current byte-identity.

Twelve IMPROVEMENT findings (DET-W6 through DET-W17) are defence-in-
depth and process. DET-W6 (phonetic-codes.json's static `regenerated_at`
literal) is the most user-visible item; renaming the field reduces a
future-maintainer-error footgun. DET-W9 (the `var _ = func() bool {...}()`
dispatch idiom in 23 dispatch_*.go files) is the most cited pattern;
documenting the carve-out in the skill closes the documentation gap.

**Prerequisite:** the cross-platform CI matrix continues to pass on every
PR. The matrix is the load-bearing empirical gate; this local review
augments it but does not replace it.

</details>

---

## 9. devops

_Source: `.planning/reviews/devops-FINDINGS.md`_

<details>
<summary>Click to expand full report</summary>

---
status: issues_found
agent: devops
scope: CI/CD + release + scripts (phases 1-8)
reviewed: 2026-05-17T08:30:00Z
finding_counts:
  critical: 0
  important: 9
  improvement: 17
  total: 26
---

# DevOps Review — fuzzymatch (Phases 1–8)

Scope: `.github/workflows/{ci,release,security,codeql,commitlint,cla,license-headers}.yml`, `.github/dependabot.yml`, `.github/CODEOWNERS`, `.github/PULL_REQUEST_TEMPLATE.md`, `Makefile`, `.goreleaser.yml`, `.golangci.yml`, `.markdownlint-cli2.yaml`, `.commitlintrc.yml`, `scripts/verify-*.sh`, `.gitignore`.

Overall posture is strong. Release discipline is sound: tag-push-only trigger, cosign keyless with `--bundle`, build-provenance attestation, post-release verify step, Syft SBOM via GoReleaser. The 5-platform matrix is complete. There are no Critical findings — no local-tag patterns, no missing signing on release artefacts, no unpinned destructive actions on release.

The Important findings cluster around (a) one unpinned action tag, (b) absence of an explicit `make check`/CI nightly workflow promised by CLAUDE.md, (c) Dependabot grouping gaps, (d) commit-message prefix divergence from Conventional Commits, (e) `release.yml` not gating on CI green, (f) gosec `-no-fail` masking failures from blocking, (g) coverage script's `_test.go` files counted against per-file floor, (h) no concurrency group on CI workflow, and (i) absent `nightly.yml` long-form fuzz + benchstat-regression job that CLAUDE.md describes.

Improvements are mostly hardening: SHA-pinning vs tag-pinning, narrower step-level conditionals, missing `make` targets enumerated in CLAUDE.md, modest gaps in the Makefile coverage of CI parity, and a number of micro-robustness opportunities in the bash scripts.

Findings follow.

---

### [Important] `DavidAnson/markdownlint-cli2-action@latest-stable` is the only floating reference in the entire workflow tree
- **File:** `/Users/johnny/Development/fuzzymatch/.github/workflows/ci.yml`
- **Phase introduced:** Phase 1 (01-02 quality gates)
- **Issue:** Every other action across all seven workflows is pinned to a specific major (`@v6`, `@v7`, `@v4`, `@v3`, `@v0`) or exact tag (`@v2.6.1`, `@v2.25.0`). `DavidAnson/markdownlint-cli2-action@latest-stable` floats — `latest-stable` resolves to whichever release the maintainer last tagged that way and can change underfoot. This breaks reproducibility of the markdownlint job and is the single failure-of-pinning across the workflow tree.
- **Standard:** CLAUDE.md "Workflow pinning (action versions pinned, no `@latest`)"; STACK.md "markdownlint-cli2 v0.22.1".
- **Action:** Code fix.
- **Rationale:** A floating tag is functionally `@latest` for supply-chain analysis. STACK.md already pins markdownlint-cli2 to v0.22.1; the action wrapper should be pinned to a corresponding major (current is `@v20` series for `markdownlint-cli2-action`).
- **Suggested fix:** Pin to `DavidAnson/markdownlint-cli2-action@v20` (or whichever current major matches the v0.22.1 CLI). For maximum hardening pin to a SHA with a comment recording the corresponding tag.

### [Important] `release.yml` does not require CI to be green before publishing
- **File:** `/Users/johnny/Development/fuzzymatch/.github/workflows/release.yml`
- **Phase introduced:** Phase 1 (01-03 release pipeline)
- **Issue:** The release workflow triggers on tag push and runs GoReleaser without a `needs:` dependency on the CI quality gate or any pre-release reachability check that the tag's commit passed CI. CLAUDE.md "Release workflow that doesn't require CI checks to pass first" is listed as a BLOCKING violation. The current model relies on branch-protection on `main` plus a maintainer's discipline ("only tag commits that merged green to main"), but the workflow itself does not encode that.
- **Standard:** CLAUDE.md `## CI Workflow Requirements > release.yml — "Runs full CI checks before release"`.
- **Action:** Code fix (recommend) OR Discuss-phase needed (decision).
- **Rationale:** A tag pushed to any commit triggers release regardless of CI state. Branch protection on `main` reduces but does not eliminate the risk: e.g. a maintainer with `Allow specified actors to bypass required pull requests` could push a tag to a non-`main` commit. The defence-in-depth pattern is to add a job-level guard that verifies the tagged SHA has a green CI run (via `gh api /repos/$REPO/commits/$SHA/check-runs`) or to add an explicit `needs:` if CI is restructured to be a reusable workflow.
- **Suggested fix:** Add a pre-release job that calls `gh api repos/${{ github.repository }}/commits/${{ github.sha }}/status` (or `check-runs`) and fails if combined status is not `success`. Alternative: convert CI into a reusable workflow and `needs: ci-required` from release.yml.

### [Important] No `nightly.yml` workflow — CLAUDE.md describes one with long-form fuzz + benchstat-regression + auto-PR corpus
- **File:** missing (`/Users/johnny/Development/fuzzymatch/.github/workflows/nightly.yml`)
- **Phase introduced:** N/A — never landed
- **Issue:** CLAUDE.md `## CI Workflow Requirements > nightly.yml (daily 03:00 UTC):` enumerates three deliverables — long-form fuzz (5 min per fuzzer), benchstat regression detection against last tagged release (>10% auto-opens an issue), and a self-hosted runner shared with mask/audit. None of these run today. `make test-fuzz` exists but only runs 60s ("short fuzz") via `make check` parity. The benchstat regression gate is `continue-on-error: true` per D-09 and would be re-enabled when a self-hosted runner becomes available — but the nightly workflow is the missing carrier that would enable that gate even without a self-hosted runner (informationally against `ubuntu-latest` over a longer baseline).
- **Standard:** CLAUDE.md `nightly.yml` section (three bullet points).
- **Action:** Discuss-phase needed.
- **Rationale:** Long-form fuzz catches state-space corners that a 60s budget can miss; nightly regression-detection on `ubuntu-latest` would be informational (per D-09) but still valuable as drift-radar. The auto-PR-new-corpus pattern is the standard go-fuzz hygiene loop. Project may be deferring this until the self-hosted runner lands; needs an explicit decision recorded.
- **Suggested fix:** Either land `nightly.yml` (workflow_dispatch + cron daily 03:00 UTC) or formally defer in 01-CONTEXT.md with a tracked Deferred Item linking the nightly enablement to the self-hosted-runner availability.

### [Important] Dependabot groups direct + indirect but no test-only group for the root module
- **File:** `/Users/johnny/Development/fuzzymatch/.github/dependabot.yml`
- **Phase introduced:** Phase 1
- **Issue:** CLAUDE.md `Dependabot` section: "Go modules (PRs grouped: indirect / direct / test-only)". The root `gomod` ecosystem block groups `indirect` and `direct` but has no `test-only` group. Practically the root module has only one runtime dep (`golang.org/x/text`) and no test-only deps because testify is forbidden in root — but a future Phase 5+ may introduce test-helpers and the grouping rule should mirror CLAUDE.md verbatim or the project should record the intentional divergence.
- **Standard:** CLAUDE.md `Dependabot: Go modules (PRs grouped: indirect / direct / test-only)`.
- **Action:** Code fix OR documentation in 01-CONTEXT.md.
- **Rationale:** CLAUDE.md compliance + future-proofing. The `tests/bdd` ecosystem block also lacks the test-only grouping despite all its deps being test-only (godog, goleak, testify); a `test-only` group there with `dependency-type: development` would be more semantically accurate than the current `indirect`/`direct` partition that has no semantic meaning for a test-only module.
- **Suggested fix:** Either (a) add `test-only` group to both `gomod` blocks for parity with CLAUDE.md, or (b) update CLAUDE.md to reflect the simplified two-group model with a one-line rationale (root has no test-only deps by design; `tests/bdd` is wholly test-only so grouping is structural not categorical).

### [Important] No auto-merge configuration for Dependabot patch-only updates
- **File:** `/Users/johnny/Development/fuzzymatch/.github/dependabot.yml` (and absent `.github/workflows/dependabot-automerge.yml`)
- **Phase introduced:** Phase 1
- **Issue:** CLAUDE.md `Dependabot: Daily check; auto-merge enabled for patch updates that pass CI`. No `dependabot-automerge.yml` workflow exists. Daily PR open is configured, but they require manual review-and-merge.
- **Standard:** CLAUDE.md `Dependabot: auto-merge enabled for patch updates that pass CI`.
- **Action:** Code fix.
- **Rationale:** Maintainer time-economy + faster security-patch landing. The standard pattern is a separate workflow that watches `pull_request` events, checks `github.event.pull_request.user.login == 'dependabot[bot]'`, parses the update-type from the Dependabot metadata action, and runs `gh pr merge --auto --squash` for `version-update:semver-patch`.
- **Suggested fix:** Add `.github/workflows/dependabot-auto-merge.yml` with `dependabot/fetch-metadata@v2` and `gh pr merge --auto --squash` on patch updates. Permissions: `contents: write`, `pull-requests: write`.

### [Important] `gosec` runs with `-no-fail` — security findings never block CI
- **File:** `/Users/johnny/Development/fuzzymatch/.github/workflows/security.yml`
- **Phase introduced:** Phase 1
- **Issue:** Line 56: `args: '-no-fail -fmt sarif -out gosec.sarif ./...'`. With `-no-fail`, gosec exits 0 regardless of findings; the SARIF is uploaded to the Security tab but no PR check is failed. CLAUDE.md `security.yml (weekly + on PR): gosec ./...` implies blocking behaviour, and `.golangci.yml` notes "the definitive gosec pass happens in .github/workflows/security.yml with SARIF upload" — but the definitive pass is not actually definitive if findings cannot fail CI.
- **Standard:** CLAUDE.md `security.yml: gosec ./...`; `.golangci.yml:11` comment claims definitive behaviour in security.yml.
- **Action:** Code fix.
- **Rationale:** Findings sit in the GitHub Security tab unattended until a maintainer notices them. Blocking on MEDIUM+ findings (with a documented allowlist mechanism for false positives) is the standard pattern. Equivalent to govulncheck which DOES fail CI today (line 36: `run: govulncheck ./...`).
- **Suggested fix:** Remove `-no-fail` and instead use `-severity medium` (or `high`) to scope blocking severity. Keep SARIF upload as a separate step that runs `if: always()` so the upload happens regardless of gosec exit. Add a `.gosec.yml` config (or `// #nosec G123 -- rationale` annotations) for documented false positives.

### [Important] Coverage-floor script lints `_test.go` files against the 90% per-file floor
- **File:** `/Users/johnny/Development/fuzzymatch/scripts/verify-coverage-floors.sh`
- **Phase introduced:** Phase 1 (01-04 determinism infra)
- **Issue:** The per-file floor loop at lines 101–137 iterates over every file appearing in `coverage.out`. The `go test -coverpkg=./...` profile includes `_test.go` files when they contain executable statements (table init, helper funcs). Test files have variable coverage — a helper that is only called from one specific test will appear with 0–100% coverage on the helper depending on inclusion. The script does not filter `_test.go` from the per-file floor, which means CI can fail on the test-file coverage of an unused test helper. This has not bitten yet because `_test.go` files typically all-execute on the same `go test` invocation, but the gap is real.
- **Standard:** `.claude/skills/go-testing-standards/SKILL.md` per-file ≥90% floor — intent is "production code", not test helpers.
- **Action:** Code fix.
- **Rationale:** Defensive correctness. Even if the current test files all hit 100%, the floor should not police `_test.go`.
- **Suggested fix:** In the awk pipeline at lines 116–137, skip files matching `_test\.go$`. Also consider skipping `doc.go` and `example_test.go` if they appear (the script already handles zero-statement files, which covers most of this).

### [Important] CI workflow has no `concurrency:` block — duplicate runs on rapid pushes
- **File:** `/Users/johnny/Development/fuzzymatch/.github/workflows/ci.yml`
- **Phase introduced:** Phase 1
- **Issue:** Only `release.yml` declares `concurrency:` (line 42). `ci.yml`, `security.yml`, `codeql.yml`, `commitlint.yml`, `license-headers.yml` will all spawn duplicate runs on rapid pushes (force-push to a PR branch, repeated `git push` while CI is mid-run). Each redundant run consumes runner-minutes and clutters PR check status.
- **Standard:** CLAUDE.md `Concurrency groups defined where appropriate` — not algorithmically specified but reviewer-judgment.
- **Action:** Code fix.
- **Rationale:** Runner-minute economy + UI clarity. The standard pattern is `concurrency: { group: ${{ github.workflow }}-${{ github.ref }}, cancel-in-progress: true }` on PR-triggered workflows (cancel-in-progress safe because the new run supersedes), and `cancel-in-progress: false` on push-to-main (cancellation could leave coverage data partially populated).
- **Suggested fix:** Add a workflow-level `concurrency:` block to ci.yml, security.yml, codeql.yml, commitlint.yml, license-headers.yml. For PRs, cancel-in-progress: true; for push-to-main, scope the group differently to allow main to always complete.

### [Important] No `cache:` directive on `actions/setup-go@v6` — module download repeats every job
- **File:** `/Users/johnny/Development/fuzzymatch/.github/workflows/{ci,release,security,codeql}.yml`
- **Phase introduced:** Phase 1
- **Issue:** `setup-go@v6` enables build/module caching by default when `go.sum` is present (Go's setup-go has had this since v5), so caching IS happening. But the workflows do not declare `cache: true` explicitly, nor do they declare `cache-dependency-path` — for the BDD sub-module (`tests/bdd/go.sum`), the cache key derives only from the root `go.sum`, which means a change to `tests/bdd/go.sum` does not invalidate the cache and CI may use stale dependencies.
- **Standard:** CLAUDE.md `Cache strategies (Go modules, build cache)`.
- **Action:** Code fix.
- **Rationale:** Correctness of cache invalidation. Default behaviour caches against the root `go.sum` only; the BDD sub-module needs its `go.sum` in the cache key so dependency drift in `tests/bdd` correctly invalidates.
- **Suggested fix:** Add to each `setup-go` step:
  ```yaml
  cache-dependency-path: |
    go.sum
    tests/bdd/go.sum
  ```

---

### [Improvement] Action tags pinned to majors not SHAs
- **File:** all workflow files
- **Phase introduced:** Phase 1
- **Issue:** Actions are pinned to majors (`actions/checkout@v6`, `actions/setup-go@v6`, `goreleaser/goreleaser-action@v7`) rather than SHAs. Major-tag pinning is the GitHub recommended baseline; SHA-pinning is the supply-chain-hardened option (used by `axonops/mask` in some places per STACK.md). For a security-conscious library being prepared for downstream consumption in Cassandra-adjacent workloads (per BOOTSTRAP context), SHA-pinning the release workflow specifically is justifiable defence-in-depth.
- **Standard:** No project-specific rule mandates SHA-pinning; CLAUDE.md `Workflow pinning (action versions pinned, no @latest)` is satisfied by majors. This is a hardening recommendation.
- **Action:** Discuss-phase needed.
- **Rationale:** Major-tag pinning allows the action author to retroactively swap the SHA the tag points to. SHA-pinning eliminates that. Trade-off: SHA-pinned actions don't auto-receive security patches; Dependabot for github-actions handles refresh.
- **Suggested fix:** SHA-pin actions invoked in `release.yml` (the highest-trust workflow) at minimum: `goreleaser/goreleaser-action`, `sigstore/cosign-installer`, `actions/attest-build-provenance`, `anchore/sbom-action/download-syft`. Leave CI workflow actions on major tags. Document the policy in `.github/dependabot.yml` comments.

### [Improvement] `anchore/sbom-action/download-syft@v0` — `@v0` is unusual major pinning
- **File:** `/Users/johnny/Development/fuzzymatch/.github/workflows/release.yml`
- **Phase introduced:** Phase 1
- **Issue:** Line 63 pins to `@v0`. The anchore/sbom-action repo's tagging is `v0.x.y` semver-zero. `@v0` resolves to whatever the latest `v0.x.y` is — which is a wider window than `@v1` or `@v2`-style pins. Project STACK.md notes "latest stable (current ~v1.42)" for syft itself but the action is at v0.
- **Standard:** CLAUDE.md `Workflow pinning`.
- **Action:** Code fix.
- **Rationale:** Tighten the resolution window. The action's current latest is around v0.18.x; `@v0.18` (or SHA-pin) is a tighter bound than `@v0`.
- **Suggested fix:** Pin to the latest minor (`@v0.18`) or SHA.

### [Improvement] `sigstore/cosign-installer@v3` major-tag pin + explicit `cosign-release` is redundant-but-safe
- **File:** `/Users/johnny/Development/fuzzymatch/.github/workflows/release.yml`
- **Phase introduced:** Phase 1
- **Issue:** Lines 58–60: `uses: sigstore/cosign-installer@v3` with `cosign-release: "v3.0.1"`. The action major v3 is pinned AND the installed cosign binary version is pinned. This is correct defence-in-depth (the action ref controls "what code runs"; `cosign-release` controls "what binary it installs"). Just noting it for completeness — no action needed. Same pattern is missing from goreleaser-action where only `version: "~> v2"` controls the binary; the action ref is `@v7`. Both are fine.
- **Standard:** N/A.
- **Action:** No action needed (informational).
- **Rationale:** Strong pinning.

### [Improvement] `.golangci.yml` `gosec` linter is enabled in lint AND in security.yml — duplicated runs
- **File:** `/Users/johnny/Development/fuzzymatch/.golangci.yml`
- **Phase introduced:** Phase 1
- **Issue:** Line 34 enables `gosec` as part of `golangci-lint run`, and `.github/workflows/security.yml` ALSO runs gosec via `securego/gosec@v2.25.0`. Two configurations, two runs, two possible signal sources for the same finding. golangci-lint's gosec is bundled and version-locked; the standalone security.yml run uses v2.25.0. These can disagree.
- **Standard:** No project-specific rule. `.golangci.yml:11` comment acknowledges that the definitive gosec runs in security.yml.
- **Action:** Code fix.
- **Rationale:** Avoid signal duplication and version-skew. Pick one as canonical.
- **Suggested fix:** Remove `gosec` from `.golangci.yml linters.enable`; leave the standalone `security.yml` as the definitive source. The `.golangci.yml` comment already implies this is intended.

### [Improvement] Two scheduled workflows both run at 06:00 UTC — capacity contention
- **File:** `/Users/johnny/Development/fuzzymatch/.github/workflows/{security,codeql}.yml`
- **Phase introduced:** Phase 1
- **Issue:** `security.yml` schedules `0 6 * * 1` (Mondays 06:00 UTC); `codeql.yml` schedules `0 6 * * 2` (Tuesdays 06:00 UTC). They are different days so no contention in practice, but a future scheduled workflow may collide. Also, 06:00 UTC is high-traffic on the GHA runner pool (start-of-EU-business). 03:00 UTC (start of low-traffic window) is the standard schedule choice and matches what CLAUDE.md specifies for `nightly.yml`.
- **Standard:** CLAUDE.md `nightly.yml (daily 03:00 UTC)`.
- **Action:** Code fix.
- **Rationale:** Avoid queue-time spikes; align with CLAUDE.md's nightly schedule.
- **Suggested fix:** Move both scheduled crons to `0 3 * * 1` / `0 3 * * 2` for consistency with CLAUDE.md's nightly cadence.

### [Improvement] No `make verify-llms-sync` target — CLAUDE.md lists one
- **File:** `/Users/johnny/Development/fuzzymatch/Makefile`
- **Phase introduced:** Phase 1
- **Issue:** CLAUDE.md `## CI Workflow Requirements > ci.yml: LLMs sync check (scripts/verify-llms-sync.sh) — llms.txt references every exported symbol`. The check exists as a Go test (`ai_friendly_test.go`) not as `scripts/verify-llms-sync.sh`. That is structurally fine (the test runs as part of `go test ./...`) but the script-target listed in CLAUDE.md does not exist. CLAUDE.md `Branch protection on main: Required status checks: ... llms-sync` references it as a named status check.
- **Standard:** CLAUDE.md `ci.yml` step list; `Branch protection on main` required-status list.
- **Action:** Discuss-phase needed.
- **Rationale:** Either (a) add a thin `scripts/verify-llms-sync.sh` that runs `go test -run TestAIFriendly_LLMSTxtReferencesEveryExportedSymbol ./...` for parity with CLAUDE.md, or (b) update CLAUDE.md to reflect that the check is a Go test inside the unit test suite and named under `make check`. Option (b) is structurally cleaner.
- **Suggested fix:** Update CLAUDE.md or add a wrapper script that invokes the existing Go test.

### [Improvement] CLAUDE.md lists status check names that don't exist as separate jobs (build, test, coverage, fuzz-short, etc.)
- **File:** N/A — branch protection policy
- **Phase introduced:** Phase 1
- **Issue:** CLAUDE.md `Branch protection on main: Required status checks: lint, vet, markdown-lint, build (5 platforms), test, coverage, fuzz-short, determinism, cross-platform-determinism, BDD, vulncheck, license-check, no-runtime-deps, llms-sync`. None of these (except `markdownlint`, `verify-license-headers`, `govulncheck`, `gosec`, `Analyze (go)`) exist as individual GitHub-check-run names. `make check` consolidates lint+vet+fmt-check+verify-license-headers+verify-deps-allowlist+tidy-check+security+test+coverage+coverage-check into a single named check `make check (linux-amd64)` (etc per matrix). This may diverge from the branch-protection rules a maintainer would want to configure literally.
- **Standard:** CLAUDE.md `Branch protection on main`.
- **Action:** Discuss-phase needed.
- **Rationale:** Either decompose CI jobs into per-step check-runs (clearer signals, more maintenance) or update CLAUDE.md to list the actual check-run names (`make check (linux-amd64)`, `make check (linux-arm64)`, `make check (darwin-amd64)`, `make check (darwin-arm64)`, `make check (windows-amd64)`, `markdownlint`, `verify-license-headers`, `commitlint`, `govulncheck`, `gosec`, `Analyze (go)`).
- **Suggested fix:** Update CLAUDE.md `Branch protection on main` to list actual check-run names, OR decompose CI workflow into one job per category. The aggregate-job approach is simpler operationally; CLAUDE.md should match reality.

### [Improvement] `Makefile` `release-check` target is tolerant no-op when `.goreleaser.yml` is absent — but `.goreleaser.yml` IS present
- **File:** `/Users/johnny/Development/fuzzymatch/Makefile`
- **Phase introduced:** Phase 1
- **Issue:** Lines 272–281: tolerant no-op for `.goreleaser.yml` absence. Plan 01-03 has landed `.goreleaser.yml`, so the tolerance is now stale. The fallthrough branch `echo "goreleaser not installed; install per docs/CONTRIBUTING (plan 01-08)."` references plan 01-08 — that planning reference will become noise post-Phase 1.
- **Standard:** CLAUDE.md `make release-check validates goreleaser config without releasing`.
- **Action:** Code fix.
- **Rationale:** Stale tolerance. Simplify to: if goreleaser absent, print clear install hint; if `.goreleaser.yml` absent, hard-fail (an absent goreleaser config in Phase 2+ is a bug).
- **Suggested fix:** Remove the `if [ -f .goreleaser.yml ]` guard; `.goreleaser.yml` exists now.

### [Improvement] `Makefile` `bench` and `bench-compare` tolerance referring to "Phase 2" is stale
- **File:** `/Users/johnny/Development/fuzzymatch/Makefile`
- **Phase introduced:** Phase 1
- **Issue:** Lines 101–113 (bench), 123–136 (bench-compare): tolerance handlers reference "pending Phase 2" / "pending plan 01-04". Phases 2–7 have shipped (per `.planning/phases/` listing), so benchmarks exist now and the tolerance branch is dead code. Same applies to `coverage-check` (line 153: "pending Phase 2") and `verify-determinism` (line 193: "Plan 01-04 lands the golden-form determinism harness. Until then this target runs `go test -run TestGolden_ ./...` which currently matches no tests" — `algorithms_golden_test.go` exists and runs now).
- **Standard:** N/A — code hygiene.
- **Action:** Code fix.
- **Rationale:** Stale comments mislead future readers. The tolerant-no-op branches still work but communicate the wrong state.
- **Suggested fix:** Update or remove the tolerance-comment blocks for `bench`, `bench-compare`, `coverage-check`, `verify-determinism`. Either delete the bootstrap-state branches entirely or annotate with "kept for re-bootstrap safety" if intentional.

### [Improvement] `verify-no-runtime-deps.sh` uses `-mod=mod` which can rewrite `go.mod`
- **File:** `/Users/johnny/Development/fuzzymatch/scripts/verify-no-runtime-deps.sh`
- **Phase introduced:** Phase 1 (01-04)
- **Issue:** Line 61: `go list -m -mod=mod -f ... all`. `-mod=mod` enables `go list` to add missing requirements to `go.mod` if needed. In CI on a checkout, `go.mod` and `go.sum` are present; `-mod=readonly` (the default in Go 1.16+) would suffice and is safer — it errors instead of mutating.
- **Standard:** Go module hygiene.
- **Action:** Code fix.
- **Rationale:** Reduce blast radius. The script should be read-only.
- **Suggested fix:** Change to `-mod=readonly` (or remove the flag — the default is readonly).

### [Improvement] `verify-coverage-floors.sh` parses `go doc -short .` which can include non-root deferred symbols
- **File:** `/Users/johnny/Development/fuzzymatch/scripts/verify-coverage-floors.sh`
- **Phase introduced:** Phase 1 (01-04)
- **Issue:** Line 166: `go doc -short . 2>/dev/null`. `go doc` operates against the current working directory's main package; this is correct for the root package but does NOT cover the `scan/` sub-package or any future sub-package. CLAUDE.md `Coverage report (must hit 95% overall, 90% per file, 100% public API)` is silent on per-sub-package public-API enforcement, but the spirit is "every exported symbol". `scan/` will land in Phase 9; the script will need updating then.
- **Standard:** CLAUDE.md `100% public API`.
- **Action:** Discuss-phase needed (Phase 9 prep).
- **Rationale:** Phase 9 (`scan/`) is approaching; the public-API floor needs to extend to it.
- **Suggested fix:** When `scan/` lands, extend the script to iterate `for pkg in . ./scan; do go doc -short "$pkg"; done`. Track in 09-CONTEXT.md as a Deferred Item.

### [Improvement] `verify-license-headers.sh` does not check for the AxonOps copyright year update
- **File:** `/Users/johnny/Development/fuzzymatch/scripts/verify-license-headers.sh`
- **Phase introduced:** Phase 1 (01-01)
- **Issue:** The header signature checked is the literal `Licensed under the Apache License, Version 2.0`. This passes regardless of the copyright year line (`Copyright 2026 AxonOps Limited`). A file with `Copyright 2024 SomeOther Org` plus the Apache-2.0 line would pass the check. The Apache header is the legally important part; the copyright line is project-hygiene. The current scope is acceptable but a stricter check (signature must include `Copyright YYYY AxonOps Limited`) is justified.
- **Standard:** `.claude/skills/documentation-standards.md` (project conventions).
- **Action:** Improvement.
- **Rationale:** Strictness. The current check is fit-for-purpose for the spirit of "Apache header present"; a stricter check enforces copyright-attribution consistency.
- **Suggested fix:** Optional — add a second header check requiring `Copyright [0-9]{4} AxonOps Limited`. Track tradeoff: more brittle (year rollovers, BSL-style headers) vs more accurate attribution check.

### [Improvement] `coverage-check` lacks tolerance flag for non-statement `_test.go` files generated by Go tooling
- **File:** `/Users/johnny/Development/fuzzymatch/scripts/verify-coverage-floors.sh`
- **Phase introduced:** Phase 1
- **Issue:** Lines 116–137: the awk parser splits coverage lines by `: . ,` and assumes filename is everything-before-first-colon. For files in subdirectories the filename will include the directory portion correctly, but for filenames containing colons (rare, but legal on POSIX) the parse breaks. Also: files like `export_test.go` are part of the root package's `package fuzzymatch` (white-box) and will appear in the coverage profile. If they have low coverage they will fail the per-file floor.
- **Standard:** N/A.
- **Action:** Code fix combined with finding [Important] above.
- **Rationale:** Combined with the `_test.go` exclusion above, this would be addressed by filtering `_test.go` files from the per-file loop.
- **Suggested fix:** See [Important] above.

### [Improvement] No commit-status check that the `bench.txt` baseline is committed when a regression PR lands
- **File:** N/A — process gap
- **Phase introduced:** Phase 1
- **Issue:** `bench.txt` is checked in (per CLAUDE.md `bench.txt IS committed`). The PR template line 69 reminds contributors to run `make bench-compare` and explain regressions, but there is no automated check that `bench.txt` has been updated when a perf-affecting change lands. `bench-compare-informational` runs but per D-09 is non-blocking.
- **Standard:** CLAUDE.md `benchstat regression > 10% fails CI` (planned, post-self-hosted-runner).
- **Action:** Discuss-phase needed.
- **Rationale:** The post-self-hosted-runner plan covers this; until then a manual gate via PR review is fit-for-purpose. Worth recording explicitly as a Deferred Item in CONTEXT.
- **Suggested fix:** Track in 01-CONTEXT.md Deferred Items (likely already covered by D-09 — verify).

### [Improvement] `release.yml` `if: startsWith(github.ref, 'refs/tags/v')` is redundant defensive guard
- **File:** `/Users/johnny/Development/fuzzymatch/.github/workflows/release.yml`
- **Phase introduced:** Phase 1
- **Issue:** The workflow's `on:` is already `push: tags: ["v*"]` so the workflow only ever runs on `v*` tags. Per-step `if: startsWith(github.ref, 'refs/tags/v')` (lines 66, 76, 84, 91, 97) is defensive redundancy. Harmless but verbose.
- **Standard:** N/A.
- **Action:** Improvement (style).
- **Rationale:** Either remove the guards (the trigger already enforces it) or keep them as documentation. Current preference reads as "belt and braces" against a future trigger expansion; that is a reasonable defensive posture but the comment block at the top of the file already promises "TAG PUSH ONLY" with a strong note.
- **Suggested fix:** Optional — remove the `if:` guards or keep them with a brief inline comment ("redundant defence-in-depth against future trigger drift").

### [Improvement] `release.yml` does not pin the `cosign verify-blob` certificate-identity to the release workflow path
- **File:** `/Users/johnny/Development/fuzzymatch/.github/workflows/release.yml`
- **Phase introduced:** Phase 1
- **Issue:** Lines 99–103: post-release sanity check uses `--certificate-identity-regexp 'https://github.com/axonops/fuzzymatch/.+'`. This regex matches any workflow in the repo, including non-release workflows that might (in future, hypothetically) also sign blobs. The tighter constraint is to match `.../release.yml@refs/tags/v.*` only.
- **Standard:** Sigstore best practice — narrowest workable identity match.
- **Action:** Code fix.
- **Rationale:** Defence-in-depth against a future workflow that signs different blobs being mistakenly trusted as the release signature.
- **Suggested fix:** Change regex to `--certificate-identity-regexp 'https://github.com/axonops/fuzzymatch/\.github/workflows/release\.yml@refs/tags/v.*'`. Same regex should be advertised in SECURITY.md (line 71) for consumer verification.

### [Improvement] `Makefile check` does not run `verify-determinism`
- **File:** `/Users/johnny/Development/fuzzymatch/Makefile`
- **Phase introduced:** Phase 1
- **Issue:** Line 34: `check: fmt-check vet lint verify-license-headers verify-deps-allowlist tidy-check security test coverage coverage-check`. `verify-determinism` is run separately in CI (ci.yml line 80), not part of `make check`. CLAUDE.md `Makefile Review: make check mirrors CI exactly` — but CI runs `make verify-determinism` then `make check`, so `make check` alone is not a complete mirror.
- **Standard:** CLAUDE.md `make check mirrors CI exactly`.
- **Action:** Code fix.
- **Rationale:** Local-development parity. A developer running `make check` should get the same exit-state signal that CI provides on the same SHA. Including `verify-determinism` in `check` adds ~5s and provides much stronger local feedback.
- **Suggested fix:** Add `verify-determinism` to the `check` target's dependency list.

### [Improvement] `Makefile` `tidy-check` runs `go mod tidy` (which mutates) before diffing
- **File:** `/Users/johnny/Development/fuzzymatch/Makefile`
- **Phase introduced:** Phase 1
- **Issue:** Lines 165–168: `tidy-check` runs `go mod tidy` in both modules, then `git diff --exit-code` — this works but mutates the working tree mid-build. If `tidy-check` fails, the working tree is left dirty (the tidy edits are not reverted). A pure-read alternative is `go mod tidy -diff` (Go 1.23+) which prints the diff without applying it.
- **Standard:** Go module hygiene.
- **Action:** Improvement.
- **Rationale:** Read-only check matches the spirit better and leaves no side-effects on failure.
- **Suggested fix:** Use `go mod tidy -diff` in both modules and fail on non-empty diff. Go 1.26 supports this.

---

## Files reviewed (paths)

- `/Users/johnny/Development/fuzzymatch/.github/workflows/ci.yml`
- `/Users/johnny/Development/fuzzymatch/.github/workflows/release.yml`
- `/Users/johnny/Development/fuzzymatch/.github/workflows/security.yml`
- `/Users/johnny/Development/fuzzymatch/.github/workflows/codeql.yml`
- `/Users/johnny/Development/fuzzymatch/.github/workflows/commitlint.yml`
- `/Users/johnny/Development/fuzzymatch/.github/workflows/cla.yml`
- `/Users/johnny/Development/fuzzymatch/.github/workflows/license-headers.yml`
- `/Users/johnny/Development/fuzzymatch/.github/dependabot.yml`
- `/Users/johnny/Development/fuzzymatch/.github/CODEOWNERS`
- `/Users/johnny/Development/fuzzymatch/.github/PULL_REQUEST_TEMPLATE.md`
- `/Users/johnny/Development/fuzzymatch/Makefile`
- `/Users/johnny/Development/fuzzymatch/.goreleaser.yml`
- `/Users/johnny/Development/fuzzymatch/.golangci.yml`
- `/Users/johnny/Development/fuzzymatch/.markdownlint-cli2.yaml`
- `/Users/johnny/Development/fuzzymatch/.commitlintrc.yml`
- `/Users/johnny/Development/fuzzymatch/.gitignore`
- `/Users/johnny/Development/fuzzymatch/scripts/verify-license-headers.sh`
- `/Users/johnny/Development/fuzzymatch/scripts/verify-no-runtime-deps.sh`
- `/Users/johnny/Development/fuzzymatch/scripts/verify-coverage-floors.sh`
- `/Users/johnny/Development/fuzzymatch/SECURITY.md`
- `/Users/johnny/Development/fuzzymatch/ai_friendly_test.go`
- `/Users/johnny/Development/fuzzymatch/go.mod`
- `/Users/johnny/Development/fuzzymatch/tests/bdd/go.mod`
- `/Users/johnny/Development/fuzzymatch/llms.txt`
- `/Users/johnny/Development/fuzzymatch/CHANGELOG.md`

## Notable positives (not findings — recorded for completeness)

- Release flow is well-designed: tag-push-only trigger; explicit comment block forbidding any other trigger; cosign `--bundle` (required in v3); separate `gh release upload` for the bundle as defence-in-depth against goreleaser's `extra_files` block; post-release `cosign verify-blob` sanity check; OIDC build-provenance attestation via `actions/attest-build-provenance@v2`.
- `.goreleaser.yml` is library-aware: `builds: [{skip: true}]` correctly tells GoReleaser there are no binaries to compile; archives produce a source tarball only; SBOM via Syft is wired into the `sboms:` block correctly.
- `permissions: contents: read` at workflow level on every workflow except `cla.yml` (which needs writes for legitimate reasons that are well-documented inline).
- Per-job permission escalation in `release.yml` lists each scope with a justification inline.
- `verify-no-runtime-deps.sh` is well-thought-through: it filters indirect modules out of the check (with a strong comment block explaining why), includes a positive "must contain these allowlist entries" check to catch accidental removal of `x/text`, and is bash-3.2 compatible for macOS-default-shell parity.
- `verify-coverage-floors.sh` enforces the three CLAUDE.md floors (95% overall, 90% per file, 100% public-API funcs via a parser of `go doc -short`) with a bootstrap-tolerance branch.
- `ai_friendly_test.go` enforces llms.txt sync via `go/parser.ParseDir` walking the root package — durable structural check.
- `golangci-lint` v2 configuration with the formatters/linters split; `goimports` configured with `local-prefixes: github.com/axonops/fuzzymatch` for the canonical import grouping.
- `commitlintrc.yml` enforces Conventional Commits with subject-case=lower-case, subject-empty=never, header-max-length=72 — strong baseline.
- CLA workflow uses `pull_request_target` correctly with the comment-text guard and a documented allowlist for bot accounts.
- Dependabot configured for three ecosystems (root gomod, tests/bdd gomod, github-actions) with sensible grouping.
- Markdown linting active on `**/*.md` with `.planning/` and `.claude/` correctly excluded.
- The 5-platform CI matrix is complete and correctly handles the macOS Intel runner deprecation (using `macos-15-intel` after `macos-13` retirement).

</details>

---

## 10. docs-writer

_Source: `.planning/reviews/docs-FINDINGS.md`_

<details>
<summary>Click to expand full report</summary>

---
status: issues_found
agent: docs-writer
scope: all documentation + godoc (phases 1-8)
reviewed: 2026-05-17T00:00:00Z
finding_counts:
  critical: 5
  important: 14
  improvement: 9
  total: 28
---

# Documentation Review Findings — fuzzymatch (Phases 1–8)

---

### [Critical] docs/requirements.md §7.5.1 contradicts implemented Ratcliff-Obershelp asymmetry

- **File:** `/Users/johnny/Development/fuzzymatch/docs/requirements.md` (line 713)
- **Phase introduced:** Phase 4 (algorithm spec) / Phase 4 (OQ-1 resolution locked 2026-05-14)
- **Issue:** Requirements §7.5.1 states "Mathematical invariants: identity, **symmetry**, range bounds." The implementation is intentionally asymmetric by OQ-1 resolution (locked 2026-05-14) to match Python `difflib.SequenceMatcher(autojunk=False).ratio()`. The symmetry claim in the spec is false for the shipped code. Additionally, §15 line 1204 lists `PropAlgorithmScore_Symmetric` as excluding "Monge-Elkan and asymmetric Tversky" but does NOT exclude Ratcliff-Obershelp, making the property-test exclusion invisible to a reader of the spec.
- **Standard:** documentation-standards SKILL.md "docs/ structure" — algorithm documentation must correctly state mathematical invariants; algorithm-correctness-standards SKILL.md §"Mathematical Invariants / Symmetric algorithms"
- **Action:** Code fix (update docs/requirements.md §7.5.1 to remove "symmetry" from the invariant list and add "asymmetric by OQ-1 design decision — see ratcliff_obershelp.go"). Also update §15 line 1204 to exclude `AlgoRatcliffObershelp` from the symmetric property list.
- **Rationale:** A consumer or reviewer reading §7.5.1 will incorrectly conclude `RatcliffObershelpScore(a,b) == RatcliffObershelpScore(b,a)`. The implementation explicitly rejects this and the prop tests intentionally omit the symmetry check. The spec contradiction will mislead future algorithm-correctness reviewers and consumer integrators.

---

### [Critical] docs/requirements.md §7.1.4 specifies `HammingDistance` returns `(int, error)` but implementation returns `int`

- **File:** `/Users/johnny/Development/fuzzymatch/docs/requirements.md` (lines 362–363)
- **Phase introduced:** Phase 2 (Hamming implementation)
- **Issue:** The requirements spec at §7.1.4 specifies `HammingDistance(a, b string) (int, error)` returning `ErrHammingLengthMismatch` on unequal lengths, and `HammingDistanceRunes(a, b string) (int, error)`. The shipped implementation (`hamming.go`) has `HammingDistance(a, b string) int` (no error return) and a documented "return max(len(a), len(b)) on mismatch" policy instead. `ErrHammingLengthMismatch` is never declared — errors.go only references it in a comment as "deferred". The same discrepancy appears in `llms-full.txt` (no Hamming function signatures documented separately, gap in coverage), but `llms.txt` correctly reflects the implemented signature `int`. The requirements spec is stale and contradicts the code.
- **Standard:** documentation-standards SKILL.md "docs/ structure" — spec must match implementation; algorithm-correctness-standards SKILL.md §"Score Normalisation / Edge cases"
- **Action:** Code fix — update docs/requirements.md §7.1.4 to match the locked implementation: `HammingDistance(a, b string) int` (silent max-length policy, no error return); remove the `ErrHammingLengthMismatch` reference from that section. Remove the parenthetical in errors.go line 31 referencing this deferred sentinel. The `api-ergonomics-reviewer` locked the silent-zero policy and this is well-documented in the implementation godoc — the spec just has not caught up.
- **Rationale:** Any future contributor or reviewer reading §7.1.4 will expect a `(int, error)` return and will be confused by the mismatch. The algorithm-correctness-reviewer gate will flag "signature does not match spec" on the next algorithm PR.

---

### [Critical] `monge_elkan.go` and `llms-full.txt` reference non-existent `ErrInvalidAlgoID` sentinel

- **File:** `/Users/johnny/Development/fuzzymatch/monge_elkan.go` (lines 162, 342); `/Users/johnny/Development/fuzzymatch/llms-full.txt` (line 652)
- **Phase introduced:** Phase 6 (Monge-Elkan)
- **Issue:** Two godoc comments in `monge_elkan.go` state that `WithMongeElkanAlgorithm` returns `ErrInvalidAlgoID`. The actual implementation in `scorer_options.go` (line 431, 435) returns `ErrInvalidAlgorithm`. `ErrInvalidAlgoID` is never declared anywhere in the codebase. `llms-full.txt` line 652 repeats the incorrect name. Any consumer reading the godoc will discriminate via `errors.Is(err, ErrInvalidAlgoID)` — which will never match because the sentinel does not exist; the actual returned sentinel is `ErrInvalidAlgorithm`.
- **Standard:** documentation-standards SKILL.md §"Godoc" — every exported function's godoc must accurately describe its return values
- **Action:** Code fix — in `monge_elkan.go` lines 162 and 342, replace `ErrInvalidAlgoID` with `ErrInvalidAlgorithm`. In `llms-full.txt` line 652, replace `ErrInvalidAlgoID` with `ErrInvalidAlgorithm`.
- **Rationale:** A consumer whose code contains `errors.Is(err, fuzzymatch.ErrInvalidAlgoID)` will receive a compilation error because the symbol does not exist. This is a directly misleading godoc promise.

---

### [Critical] `docs/algorithms.md` is entirely scaffold — all 23 algorithms still say "Primary source: TBD" and "Status: planned" despite all being shipped

- **File:** `/Users/johnny/Development/fuzzymatch/docs/algorithms.md`
- **Phase introduced:** Phase 1 (scaffold created); applicable to all phases 2–8 (each phase shipped algorithms without updating this document)
- **Issue:** All 23 algorithm entries in `docs/algorithms.md` have `Primary source: TBD — filled in by the implementing phase` and `Status: planned (Phase N)`. Phases 2 through 8 have all shipped. None of the per-algorithm sections contain the mandatory fields from the documentation-standards skill: Algorithm name (canonical form), Description (1–3 paragraphs in prose), Mathematical formulation, Complexity, Score normalisation rule, Mathematical invariants, Edge cases, Reference vectors, Intended use cases, Comparable references. The document is a 239-line scaffold with no substantive per-algorithm content.
- **Standard:** documentation-standards SKILL.md §"Algorithm Documentation" — every field listed is mandatory; system-level CLAUDE.md "Algorithm Documentation" section repeats the requirement verbatim
- **Action:** Code fix — fill in all 23 algorithm entries with the mandatory fields. The source material exists in `docs/requirements.md` §7, the implementation file godoc blocks, and the test files. This is the highest-priority single documentation task remaining.
- **Rationale:** `docs/algorithms.md` is the per-algorithm detail document linked from the README catalogue table. Every link in the README (`[docs/algorithms.md#levenshtein]` etc.) lands on a stub. Consumers, contributors, and the algorithm-correctness-reviewer all expect full detail here.

---

### [Critical] README Quick Start example is Phase-1 primitives only — stale relative to phases 2–8 shipping

- **File:** `/Users/johnny/Development/fuzzymatch/README.md` (lines 96–118)
- **Phase introduced:** Phase 1 (Quick Start written for Phase-1 scope); Phases 2–8 did not update it
- **Issue:** The Quick Start section explicitly says "the full algorithm-driven quick start is added with Phase 2" and shows only `Normalise` and `Tokenise`. All 23 algorithms and the full Scorer are now shipped. The README prominently tells consumers that `LevenshteinScore` "lands in Phase 2" — which is past. The documentation-standards skill requires the README Quick Start to contain "a complete, working program" including `DefaultScorer + custom Scorer + scan`. The current example does not demonstrate a single algorithm function or the Scorer.
- **Standard:** documentation-standards SKILL.md §"README" — "First code block shows a complete, working program (the 'headline example' verified by `readme_shop_front_test.go`)"; system-level CLAUDE.md §"README Structure" item 8 — "Quick Start with runnable example (DefaultScorer + custom Scorer + scan)"
- **Action:** Code fix — replace the Phase-1-only Quick Start with a multi-block example: one block showing `LevenshteinScore` + `JaroWinklerScore`, one block showing `DefaultScorer().Score()`, one block showing `NewScorer` custom composition. Remove the "Phase 2" note.
- **Rationale:** Any first-time consumer reading the README sees a Quick Start that shows primitives, not the algorithms. The readme_shop_front_test.go meta-test should be verifying the headline example compiles and produces documented output — with a Phase-1-only example that test is not exercising the primary API surface.

---

### [Important] `doc.go` platform list omits `darwin/amd64` — lists only 4 of 5 platforms

- **File:** `/Users/johnny/Development/fuzzymatch/doc.go` (line 40–41)
- **Phase introduced:** Phase 1
- **Issue:** The package godoc says "Output is byte-identical across linux/amd64, linux/arm64, darwin/arm64, and windows/amd64." This omits `darwin/amd64`. The spec (docs/requirements.md §13.3 line 1184), CLAUDE.md, README.md, and llms-full.txt (line 1303) all list five platforms including `darwin/amd64`. Only `doc.go` is missing one entry.
- **Standard:** documentation-standards SKILL.md §"Godoc" — package overview must be accurate; docs must be internally consistent
- **Action:** Code fix — add `darwin/amd64` to the platform list in `doc.go` line 40: "linux/amd64, linux/arm64, darwin/amd64, darwin/arm64, and windows/amd64"
- **Rationale:** Minor inaccuracy but since this appears in the published godoc on pkg.go.dev it directly misleads developers on Apple Intel machines who may wonder whether the library is tested on their platform.

---

### [Important] `algorithm-correctness-standards` SKILL.md does not list Ratcliff-Obershelp as an asymmetric exception in the symmetry section

- **File:** `/Users/johnny/Development/fuzzymatch/.claude/skills/algorithm-correctness-standards/SKILL.md` (line 119)
- **Phase introduced:** Phase 1 (skill written); Phase 4 (OQ-1 resolution locked)
- **Issue:** Line 119 states "All algorithms in this library are symmetric **except** Monge-Elkan (when used asymmetrically) and Tversky (when α ≠ β)." Ratcliff-Obershelp is not listed. However the OQ-1 resolution (locked 2026-05-14) makes Ratcliff-Obershelp intentionally asymmetric. The skill is used by the algorithm-correctness-reviewer as the gating standard — future reviewers following the skill will incorrectly require a symmetry property test for Ratcliff-Obershelp.
- **Standard:** No governing standard — this is a skill accuracy issue
- **Action:** Skill clarification — add Ratcliff-Obershelp to the exception list: "All algorithms in this library are symmetric **except** Monge-Elkan (when used asymmetrically), Tversky (when α ≠ β), and Ratcliff-Obershelp (intentionally asymmetric per OQ-1 RESOLUTION LOCKED 2026-05-14 — see ratcliff_obershelp.go)."
- **Rationale:** The skill is the authoritative reference for the algorithm-correctness-reviewer. An incorrect skill produces incorrect review decisions.

---

### [Important] README is missing required structural elements from the spec and CLAUDE.md

- **File:** `/Users/johnny/Development/fuzzymatch/README.md`
- **Phase introduced:** Phase 1
- **Issue:** Comparing the current README sections against `docs/requirements.md §16.1` and `system-level CLAUDE.md §"README Structure"`, the following required elements are absent: (1) a logo image — requirements §16.1 item 1 specifies `.github/images/logo-readme.png`; the `.github/images/` directory does not exist; (2) a "Quick links" section (requirements item 4 — links to Quick Start, Features, Algorithms, Docs, API Reference); (3) an emoji-headed section "🔍 Overview" (requirements item 7 — currently the README has "What this is" without an emoji); (4) "✨ Key Features" is present but uses no emoji header; (5) no dedicated "🛠 Scorer composition" section (requirements item 12 / system-level CLAUDE.md item 10); (6) no "🎯 Tuning Guidance" section with link to tuning.md (requirements item 15 / system-level CLAUDE.md item 13) — tuning.md is mentioned in Configuration but there is no dedicated section; (7) a CI badge is missing — CLAUDE.md "Docs Stack" specifies badges "(CI, Go Reference, Go Report Card, License, Status pre-release)"; only 4 badges are present (Go Reference, Go Report Card, License, Status); CI badge is absent.
- **Standard:** documentation-standards SKILL.md §"README"; system-level CLAUDE.md §"README Structure"; docs/requirements.md §16.1
- **Action:** Code fix — add missing structural elements. For the CI badge, add the GitHub Actions badge for the `ci.yml` workflow. The logo image can be deferred until artwork is ready but the placeholder reference in requirements.md should match the actual state (either remove the requirement note or add a placeholder). Add the Quick Links section, emoji headings, dedicated Scorer composition section, and Tuning Guidance section.
- **Rationale:** The README is the library's public front door. The documented structure mirrors `axonops/mask` style — deviating from it makes the project feel inconsistent with its own stated intent.

---

### [Important] README `scorer.md` reference says "once Phase 8 lands" — Phase 8 has shipped

- **File:** `/Users/johnny/Development/fuzzymatch/README.md` (lines 198, 224)
- **Phase introduced:** Phase 8 (text added before Phase 8 shipped; never updated)
- **Issue:** Line 198: "see [`docs/scorer.md`] for the `Scorer` API once Phase 8 lands." Line 224: "`docs/scorer.md` — `Scorer` configuration and tuning (Phase 8)." The Scorer shipped in Phase 8. These lines should be updated to present tense. Additional stale Phase references: line 96 (Quick Start note about Phase 2), line 57 ("Phase 8" parenthetical on the Scorer feature), line 58 ("Phase 9" on scan), line 204 ("from Phase 2" algorithm function note), line 206 ("Phase 8" Scorer immutability), line 123 ("once it lands (Phase 2+)").
- **Standard:** documentation-standards SKILL.md §"README" — docs must reflect current state
- **Action:** Code fix — audit all Phase-N references and remove or update those that are now past history. Phrases like "once Phase 8 lands" should become present tense; "Phase 2+" should be removed where the algorithms are shipped; "(Phase 8)" labels on shipped features should be dropped.
- **Rationale:** Forward-looking scaffolding language left in a shipped product creates consumer confusion about what is actually available.

---

### [Important] `example_test.go` is missing required examples for Normalise, Tokenise, Scorer, and ScoreAll

- **File:** `/Users/johnny/Development/fuzzymatch/example_test.go`
- **Phase introduced:** Phase 8 (Scorer examples were the responsibility of Phase 8)
- **Issue:** `docs/requirements.md §16.5` specifies the following `Example*` functions must exist in `example_test.go`: `ExampleScorer` (DefaultScorer with a typical pair), `ExampleNewScorer` (custom Scorer composition), `ExampleScorer_ScoreAll` (per-algorithm breakdown), `ExampleScorer_Match` (threshold matching), `ExampleNormalise`, and `ExampleTokenise`. None of these six functions are present in `example_test.go`. The 39 examples that exist cover all 23 algorithm score functions and phonetic encoder functions, but the Scorer layer and the two Phase-1 primitive functions have no godoc examples visible on pkg.go.dev.
- **Standard:** documentation-standards SKILL.md §"Godoc" — "Include runnable `Example` functions in `example_test.go` for major use cases"; docs/requirements.md §16.5
- **Action:** Code fix — add `ExampleNormalise`, `ExampleTokenise`, `ExampleScorer` (using DefaultScorer), `ExampleNewScorer` (custom 2-algorithm composition with threshold), `ExampleScorer_ScoreAll`, and `ExampleScorer_Match` to `example_test.go`. All must have `// Output:` blocks so they are verified by `go test`.
- **Rationale:** Missing pkg.go.dev examples for the Scorer mean the primary consumer-facing API layer has no runnable documentation. This is the most common first point of contact for new users.

---

### [Important] CHANGELOG is a stub — no entries for any of the 23 implemented algorithms, Scorer, or cross-validation infrastructure

- **File:** `/Users/johnny/Development/fuzzymatch/CHANGELOG.md`
- **Phase introduced:** Phases 2–8 (no CHANGELOG entries added when features shipped)
- **Issue:** The CHANGELOG has a single `[Unreleased]` section with five bootstrap-only bullet points from Phase 1. All 23 algorithms (Phases 2–7), the composite Scorer (Phase 8), all property tests, fuzz tests, BDD scenarios, the cross-validation infrastructure (three corpora), and the examples directory have shipped with zero CHANGELOG entries. Keep a Changelog format requires added/changed/deprecated/removed/fixed/security subsections per release or per unreleased feature group. The current state violates `CONTRIBUTING.md §CHANGELOG` (line 27: "A CHANGELOG entry has been added under `## [Unreleased]`") on every one of the ~40 feature commits.
- **Standard:** documentation-standards SKILL.md §"docs/ structure" references Keep a Changelog; CONTRIBUTING.md §"Pre-PR Checklist" item 4; docs/requirements.md §16.6
- **Action:** Code fix — back-fill CHANGELOG entries for all shipped features. Minimum: one bullet per algorithm, one for the cross-validation corpora, one for the Scorer, one for the examples. Group by phase for clarity.
- **Rationale:** A consumer evaluating the library's history to understand what changed between notional releases will find an empty record. The goreleaser release workflow extracts the CHANGELOG section for the GitHub release description — a stub will produce an empty release body.

---

### [Important] `CONTRIBUTING.md` references `docs/requirements.md §11.2` which does not exist

- **File:** `/Users/johnny/Development/fuzzymatch/CONTRIBUTING.md` (line 186–187)
- **Phase introduced:** Phase 1
- **Issue:** CONTRIBUTING.md states "This is requirement REL-07 from [`docs/requirements.md`] §11.2." Section §11 of `docs/requirements.md` is titled "Phonetic Algorithm Integration" and has no §11.2 subsection. The algorithm deprecation policy described in CONTRIBUTING.md corresponds approximately to the note at requirements §13.1 ("Algorithm score stability") or the general text in §5 design principles — there is no §11.2. The linked section does not exist.
- **Standard:** documentation-standards SKILL.md — cross-references must be accurate
- **Action:** Code fix — change the section reference to the correct one. The closest match is `docs/requirements.md §13.1` (Algorithm score stability, lines 1174–1177) or optionally add a dedicated deprecation-policy subsection to requirements.md. Also note that `REL-07` exists only in `.planning/REQUIREMENTS.md`, not in `docs/requirements.md` — the reference to `REL-07` in CONTRIBUTING.md points to a planning artifact ID, not a user-facing requirement identifier.
- **Rationale:** A contributor following the CONTRIBUTING.md link will land on a non-existent section and lose confidence in the document's accuracy.

---

### [Important] `llms-full.txt` Phase 4 section omits Hamming function signatures entirely

- **File:** `/Users/johnny/Development/fuzzymatch/llms-full.txt`
- **Phase introduced:** Phase 2 (Hamming shipped); llms-full.txt Phase 4 section was authored later
- **Issue:** The `llms-full.txt` document has a "Phase 4 algorithm surface" section covering Strcmp95, LCSStr, and Ratcliff-Obershelp. Hamming, Jaro, Jaro-Winkler, Levenshtein, and DL-OSA/Full appear elsewhere but there is no section dedicated to Hamming's exact function signatures with godoc summaries (the format used for every other algorithm). The `llms.txt` concise index correctly lists `HammingDistance`, `HammingDistanceRunes`, `HammingScore`, `HammingScoreRunes` as `func ... int` or `float64` but `llms-full.txt` provides no per-function godoc summary for Hamming.
- **Standard:** documentation-standards SKILL.md §"llms.txt / llms-full.txt" — "llms-full.txt includes full API signatures, all algorithm citations, and worked examples"
- **Action:** Code fix — add a Phase 2 Hamming section to llms-full.txt mirroring the style of other algorithm sections (typed signatures with inline godoc summaries).
- **Rationale:** AI assistants reading llms-full.txt will find Hamming underdocumented relative to every other algorithm.

---

### [Important] `strcmp95.go` godoc block does not state that Strcmp95 is not a metric and does not satisfy the triangle inequality

- **File:** `/Users/johnny/Development/fuzzymatch/strcmp95.go`
- **Phase introduced:** Phase 4
- **Issue:** The algorithm-correctness-standards skill explicitly states: "Jaro, Jaro-Winkler, Strcmp95 do NOT satisfy triangle inequality and are NOT metrics. Document this in their godoc." Jaro (`jaro.go` line 60) documents "Jaro is NOT a metric: the triangle inequality does not hold". Jaro-Winkler (`jarowinkler.go` lines 58–64) documents "Jaro-Winkler is NOT a metric". Strcmp95 has no such statement anywhere in `strcmp95.go` — neither in the file-level comment block nor in the function godocs.
- **Standard:** algorithm-correctness-standards SKILL.md §"Mathematical Invariants / Distance-based algorithms" and §"Symmetric algorithms" — specifically: "Jaro, Jaro-Winkler, Strcmp95 do NOT satisfy triangle inequality and are NOT metrics. Document this in their godoc."
- **Action:** Code fix — add to the `strcmp95.go` file-level comment (after the API hierarchy section): "# Strcmp95 is NOT a metric. Triangle inequality does not hold. Inherits the non-metric property of Jaro-Winkler."
- **Rationale:** The skill mandates this documentation for all three algorithms. Consistency with jaro.go and jarowinkler.go is broken, and a reviewer following the skill would flag this as a BLOCKING gap.

---

### [Important] `docs/faq.md` is missing the mandated "Why no Soft-TFIDF?" question

- **File:** `/Users/johnny/Development/fuzzymatch/docs/faq.md`
- **Phase introduced:** Phase 1 (faq.md created with DX-06 mandated entries); subsequent phases did not add
- **Issue:** `docs/requirements.md §16.2` specifies faq.md must cover "why not Needleman-Wunsch", "why not Soft-TFIDF", "why not ML embeddings", "why was Metaphone 3 excluded". The current faq.md has Needleman-Wunsch, Metaphone 3, ML/embeddings, plus bonus questions (phonetic-as-binary, generics, x/text). The "Why no Soft-TFIDF?" entry is absent. Soft-TFIDF is explicitly out of scope per requirements §4 ("requires a consumer-supplied corpus frequency table, which conflicts with the library's stateless pure-function design").
- **Standard:** documentation-standards SKILL.md §"docs/ structure" — faq.md must address the listed exclusions; docs/requirements.md §16.2
- **Action:** Code fix — add a "Why no Soft-TFIDF?" FAQ entry citing the stateless pure-function design conflict documented in requirements §4.
- **Rationale:** Consumers from the IR/NLP community will ask about Soft-TFIDF. Its absence without explanation will prompt repeated identical questions.

---

### [Important] `prior-art-research.md` contains American English spellings in a British English project

- **File:** `/Users/johnny/Development/fuzzymatch/docs/prior-art-research.md` (lines 163, 324, 330)
- **Phase introduced:** Phase 1
- **Issue:** The file uses "normalization" (three occurrences: lines 163, 324, 330) rather than "normalisation". The documentation-standards skill specifies British English throughout (normalisation, behaviour, organisation, etc.). All other docs in the project use British English consistently.
- **Standard:** documentation-standards SKILL.md §"Style" — "British English (colour, behaviour, organisation, normalisation, optimisation)"
- **Action:** Code fix — replace all three "normalization" instances with "normalisation" in `docs/prior-art-research.md`.
- **Rationale:** Consistency in spelling conventions across all docs.

---

### [Important] `docs/extending.md` — the composing-phonetic-algorithms section references `docs/requirements.md §11` but that section covers Phonetic Integration specs, not consumer-facing composition patterns

- **File:** `/Users/johnny/Development/fuzzymatch/docs/extending.md` (line 33)
- **Phase introduced:** Phase 7
- **Issue:** The "Composing phonetic algorithms with edit distance" section says "See `docs/requirements.md` §11." Section 11 is the internal spec for phonetic algorithm integration (encoding rules, variant discipline, cross-validation). It is not a consumer-facing guide for composing phonetic codes with edit distance. The consumer-useful content is actually in the function godocs (`SoundexCode`, `DoubleMetaphoneKeys`, `NYSIISCode`, `MRACode`) and in the relevant implementation files. Pointing at the spec section is a developer-facing reference, not a user-guide reference.
- **Standard:** documentation-standards SKILL.md §"docs/ structure" — extending.md is a consumer guide, not a spec cross-reference
- **Action:** Code fix — replace the §11 cross-reference with a concrete code example showing `LevenshteinScore(SoundexCode(a), SoundexCode(b))` pattern, plus a note pointing at the function godocs for the encoder surfaces.
- **Rationale:** The extending.md section is currently a TBD stub with an unhelpful reference. A consumer wanting to compose phonetic codes with edit distance gets no actionable content.

---

### [Improvement] `docs/algorithms.md` H2 anchor IDs use camelCase (e.g. `#sorensendice`) but README links use inconsistent casing

- **File:** `/Users/johnny/Development/fuzzymatch/README.md` (lines 143–145); `/Users/johnny/Development/fuzzymatch/docs/algorithms.md` (H2 headings)
- **Phase introduced:** Phase 1
- **Issue:** The README algorithm catalogue table links use the H2 header text directly as anchors: `#qgramjaccard`, `#sorensendice`, `#cosine`, `#tversky`, etc. The corresponding H2 headings in algorithms.md are `## QGramJaccard`, `## SorensenDice`, `## Cosine`, `## Tversky` — which produce anchors `#qgramjaccard` (GitHub Markdown lowercases). This is consistent for the two-word entries but inconsistent with the file-level comment that says "H2 anchors below match the algorithm's canonical spelling (e.g. `#levenshtein`, `#damerau-levenshtein-osa`)". The README table links for `#damerau-levenshtein-osa` and `#damerau-levenshtein-full` use hyphens (correct for the heading text) while single-word algorithms use no separator. This is actually correct behaviour. However the README link for `#qgramjaccard` will not resolve if the H2 heading is changed to `## Q-Gram Jaccard` (the canonical form per CLAUDE.md). When algorithms.md is fully filled in (Critical finding above), the headings should be verified against all README links.
- **Standard:** No governing standard — this is a preemptive consistency flag
- **Action:** Discuss-phase needed — when algorithms.md is populated, verify all deep-link anchors in the README match the H2 headings. Consider using the canonical hyphenated form "Q-Gram Jaccard" for the H2 to match the spec terminology.
- **Rationale:** Broken anchor links are a common consequence of renaming headings during content population.

---

### [Improvement] README `Table of contents` uses non-emoji section names that diverge from the emoji section headers required by spec

- **File:** `/Users/johnny/Development/fuzzymatch/README.md` (lines 14–28)
- **Phase introduced:** Phase 1
- **Issue:** The Table of Contents links `[Status](#-status)`, `[What this is](#what-this-is)`, etc. The corresponding sections use `## ⚠ Status` (has emoji) and `## What this is` (no emoji). The spec requires emoji section headers (🔍 Overview, ✨ Key Features, ❓ Why fuzzymatch?, 🚀 Quick Start, 🛠 Scorer composition, 🧵 Thread Safety, etc.). The TOC links work for existing sections but will need updating when the missing sections are added. Also, `[What this is]` does not match the spec's "🔍 Overview" heading; `[Key features]` does not match "✨ Key Features"; `[Why this library exists]` does not match "❓ Why fuzzymatch?".
- **Standard:** docs/requirements.md §16.1 items 7–20; system-level CLAUDE.md §"README Structure"
- **Action:** Code fix (coordinate with the README restructuring from the Critical finding above) — standardise section headings to use the emoji + title from the spec; update TOC links accordingly.
- **Rationale:** The spec mandates emoji headers mirroring axonops/mask. The current README uses them inconsistently (only ⚠ Status and 🤖 For AI Assistants have emojis).

---

### [Improvement] `scorer.md` "Method Reference" table lists five methods but the prose says "All four methods"

- **File:** `/Users/johnny/Development/fuzzymatch/docs/scorer.md` (line 107)
- **Phase introduced:** Phase 8
- **Issue:** The method reference table (lines 98–104) lists five methods: `Score`, `Match`, `ScoreAll`, `Threshold`, and `Algorithms`. The prose immediately after (line 107) says "All four methods are pure functions on the `*Scorer` receiver." The count is incorrect: there are five, not four.
- **Standard:** No governing standard — accuracy of documentation prose
- **Action:** Code fix — change "All four methods" to "All five methods" on line 107.
- **Rationale:** Small but factually wrong — confusing to a consumer counting methods against the table.

---

### [Improvement] `scorer.md` "Concurrency" section says "four methods" (`Score`, `Match`, `ScoreAll`, `Threshold`, `Algorithms`) — same off-by-one as above

- **File:** `/Users/johnny/Development/fuzzymatch/docs/scorer.md` (line 258)
- **Phase introduced:** Phase 8
- **Issue:** The concurrency section says "All four methods (`Score`, `Match`, `ScoreAll`, `Threshold`, `Algorithms`) are safe to call concurrently." It then lists five methods. Same off-by-one count as the Method Reference section.
- **Standard:** No governing standard — accuracy of documentation prose
- **Action:** Code fix — change "All four methods" to "All five methods" on line 258.
- **Rationale:** Same as the finding above; these two sentences were written at different times and the count was not updated when the fifth method was added.

---

### [Improvement] `docs/cross-validation.md` is undocumented in `docs/requirements.md §16.2` — missing from the docs/ structure spec

- **File:** `/Users/johnny/Development/fuzzymatch/docs/cross-validation.md`; `/Users/johnny/Development/fuzzymatch/docs/requirements.md` §16.2 (line 1430–1438)
- **Phase introduced:** Phase 6 (cross-validation doc created)
- **Issue:** `docs/requirements.md §16.2` enumerates the expected docs/ files: algorithms.md, scorer.md, scan.md, tuning.md, extending.md, performance.md, faq.md. It does not mention `cross-validation.md`, which was added in Phase 6 as a reference document for the cross-validation corpora. The file exists and is linked from CONTRIBUTING.md. The spec does not anticipate it.
- **Standard:** documentation-standards SKILL.md §"docs/ structure" — should enumerate all docs/ files
- **Action:** Code fix — add `docs/cross-validation.md` to the §16.2 list in requirements.md with a one-line description: "cross-validation.md — corpora, pinned reference implementations, regeneration protocol, and variant-divergence tagging mechanism."
- **Rationale:** The requirements.md is the authoritative spec for the repository layout. An unlisted file creates ambiguity about whether it is intentional.

---

### [Improvement] `CHANGELOG.md` uses `### Notes` subsection which is not a Keep-a-Changelog convention

- **File:** `/Users/johnny/Development/fuzzymatch/CHANGELOG.md` (line 17)
- **Phase introduced:** Phase 1
- **Issue:** Keep a Changelog 1.1.0 (the format linked in the CHANGELOG) defines exactly six subsection types under each release: Added, Changed, Deprecated, Removed, Fixed, Security. The current `[Unreleased]` section includes `### Notes` — which is not a Keep-a-Changelog convention. The Note content ("This project is pre-release. The API is not stable until `v1.0.0`.") belongs in the `[Unreleased]` section preamble or as a notice at the top of the file, not as a `### Notes` subsection.
- **Standard:** docs/requirements.md §16.6 "CHANGELOG following Keep-a-Changelog format"
- **Action:** Code fix — move the pre-release note to a blockquote or freeform paragraph before the `### Added` subsection; remove the non-standard `### Notes` header.
- **Rationale:** Keep-a-Changelog tooling and goreleaser's changelog extractor may not handle non-standard subsection names gracefully.

---

### [Improvement] `docs/performance.md` "ASCII fast paths" section incorrectly omits Damerau-Levenshtein from the list

- **File:** `/Users/johnny/Development/fuzzymatch/docs/performance.md` (lines 47–52)
- **Phase introduced:** Phase 8 (performance.md scaffold written; not updated when DL algorithms shipped)
- **Issue:** The ASCII fast paths section lists "Levenshtein, Jaro, Jaro-Winkler, Hamming, etc." — a partially-illustrative list. All character-based algorithms (DL-OSA, DL-Full, SWG, LCSStr, Strcmp95) have ASCII fast paths documented in their implementation files. The "etc." placeholder means the section is inaccurate as a reference for consumers choosing algorithms by allocation profile. This section should either reference the per-algorithm benchmarks in `bench.txt` or enumerate all algorithms with ASCII fast paths.
- **Standard:** documentation-standards SKILL.md §"docs/ structure" — performance.md covers "ASCII fast paths" accurately
- **Action:** Code fix (deferred — can wait for the full performance.md population pass): at minimum, change "etc." to "and all character-based algorithms" to avoid implied completeness.
- **Rationale:** Minor accuracy gap; low impact while performance.md is still a scaffold but should be fixed before v1.0.0.

---

### [Improvement] `algoid.go` constant godocs for token-tier algorithms cite "Cohen, Ravikumar & Fienberg 2003 — SecondString" but `llms-full.txt` and README cite "SeatGeek fuzzywuzzy / RapidFuzz"

- **File:** `/Users/johnny/Development/fuzzymatch/algoid.go` (lines 134–147); `/Users/johnny/Development/fuzzymatch/llms.txt` (lines 152–156); `/Users/johnny/Development/fuzzymatch/llms-full.txt` (line 83–84)
- **Phase introduced:** Phase 1 (algoid.go constants) / Phase 6 (llms.txt updated)
- **Issue:** The AlgoID godoc comments in algoid.go cite "Cohen, Ravikumar & Fienberg 2003 — SecondString library reference" for TokenSortRatio, TokenSetRatio, PartialRatio, and TokenJaccard. The README algorithm catalogue table and llms.txt cite "SeatGeek fuzzywuzzy / RapidFuzz" for the first three and "Jaccard 1912" for TokenJaccard. The algorithm-correctness-standards skill and the implementation files (token_sort_ratio.go etc.) use the RapidFuzz lineage. The SecondString citation in algoid.go is inconsistent with the rest of the project's source attribution.
- **Standard:** algorithm-correctness-standards SKILL.md §"Primary Source Citation" — citation must be consistent across all references to the same algorithm
- **Action:** Code fix — update the AlgoID constant godocs in algoid.go for AlgoTokenSortRatio, AlgoTokenSetRatio, AlgoPartialRatio to cite "SeatGeek fuzzywuzzy (2014); modern reference RapidFuzz (Bachmann, M., 2020–)" rather than SecondString. For AlgoTokenJaccard, the citation should be "Jaccard, P. (1912)" to match the rest of the project.
- **Rationale:** Inconsistent citations make it unclear which source governs the algorithm's reference vectors and constants. The algorithm-correctness-reviewer uses citations to gate correctness — an inconsistent source in algoid.go will surface as a reviewer question.

</details>

---

## 11. go-quality (issue-writer skipped — see header)

_Source: `.planning/reviews/go-quality-FINDINGS.md`_

<details>
<summary>Click to expand full report</summary>

---
status: issues_found
agent: go-quality
scope: full automated quality gate
reviewed: 2026-05-17T06:36:52Z
finding_counts:
  critical: 3
  important: 17
  improvement: 3
  total: 23
---

# Go Quality Gate Findings

All commands run from `/Users/johnny/Development/fuzzymatch` (repo root).
Toolchain: `go 1.26.3 darwin/arm64`, `golangci-lint 2.11.4`.

---

## Critical Findings

### [Critical] golangci-lint run exits non-zero — 27 issues across 7 linters

- **File:** global (multiple files — see Important findings 4–15 for per-issue detail)
- **Phase introduced:** Phases 2, 7 (algorithm files); Phase 8 (scorer)
- **Issue:** `golangci-lint run ./...` exits 1. Full output:
  ```
  double_metaphone.go:716:5: ifElseChain: rewrite if-else to switch statement (gocritic)
  double_metaphone.go:744:4: dupBranchBody: both branches in if statement have same body (gocritic)
  double_metaphone.go:120:1: cyclomatic complexity 12 of func `dmSlgCheck` is high (> 10) (gocyclo)
  double_metaphone.go:176:1: cyclomatic complexity 17 of func `dmPrep` is high (> 10) (gocyclo)
  double_metaphone.go:243:1: cyclomatic complexity 280 of func `DoubleMetaphoneKeys` is high (> 10) (gocyclo)
  double_metaphone.go:890:1: cyclomatic complexity 13 of func `DoubleMetaphoneScore` is high (> 10) (gocyclo)
  mra.go:135:1: cyclomatic complexity 16 of func `MRACode` is high (> 10) (gocyclo)
  mra.go:241:1: cyclomatic complexity 23 of func `MRACompare` is high (> 10) (gocyclo)
  nysiis.go:116:1: cyclomatic complexity 78 of func `NYSIISCode` is high (> 10) (gocyclo)
  soundex.go:147:1: cyclomatic complexity 29 of func `SoundexCode` is high (> 10) (gocyclo)
  double_metaphone_test.go:159:1: File is not properly formatted (gofumpt)
  mra_bench_test.go:30:1: File is not properly formatted (gofumpt)
  nysiis_bench_test.go:28:1: File is not properly formatted (gofumpt)
  mra.go:211:22: G602: slice index out of range (gosec)
  mra.go:212:22: G602: slice index out of range (gosec)
  mra.go:213:22: G602: slice index out of range (gosec)
  scorer_golden_test.go:287:77: `abreviation` is a misspelling of `abbreviation` (misspell)
  scorer_options_internal_test.go:28:11: `artifact` is a misspelling of `artefact` (misspell)
  double_metaphone.go:397:33: QF1001: could apply De Morgan's law (staticcheck)
  double_metaphone.go:825:7: QF1001: could apply De Morgan's law (staticcheck)
  double_metaphone_fuzz_test.go:105:9: QF1001: could apply De Morgan's law (staticcheck)
  double_metaphone_test.go:322:8: QF1001: could apply De Morgan's law (staticcheck)
  errors_test.go:209:5: SA4004: the surrounding loop is unconditionally terminated (staticcheck)
  props_test.go:3492:8: QF1001: could apply De Morgan's law (staticcheck)
  soundex_test.go:203:6: QF1001: could apply De Morgan's law (staticcheck)
  scorer_options_internal_test.go:113:43: probeScoreFnInvoke - i always receives 0 (unparam)
  soundex.go:275:31: runeAt - result 0 (rune) is never used (unparam)
  27 issues
  ```
- **Standard:** `go-coding-standards` SKILL.md §Quality Gate item 3; CLAUDE.md "Checks" item 3
- **Action:** Code fix (see individual findings 4–15 for per-issue remediation)
- **Rationale:** `lint` is a direct dependency of `make check`. CI blocks on this gate. The 27 issues span correctness risk (dupBranchBody), style (gofumpt, De Morgan), cyclomatic complexity suppressions, a gosec false positive, a misspell locale mismatch, and two unparam observations.

---

### [Critical] Coverage floor gate exits non-zero — overall 90.2% below 95.0% floor

- **File:** `scripts/verify-coverage-floors.sh` / `coverage.out`
- **Phase introduced:** all phases (cumulative gap across algorithm implementations)
- **Issue:** `bash scripts/verify-coverage-floors.sh coverage.out` exits 1:
  ```
  verify-coverage-floors: FAIL — overall coverage 90.2% < 95.0%
  ```
  The `make coverage` run (296s, race-enabled) shows `coverage: 90.4% of statements`; the `go tool cover -func` total line reads `90.2%`. The 4.8 percentage-point gap means roughly 1 in 20 statements in the root package is not exercised.
- **Standard:** `go-testing-standards` SKILL.md §Coverage Targets ("≥ 95% overall"); CLAUDE.md Constraints ("Coverage targets: ≥ 95% overall")
- **Action:** Code fix — add test cases to bring the overall floor to ≥ 95%. The most impactful targets are `double_metaphone.go` (heap-path in `dmPrep`, untested phonetic rule branches in `DoubleMetaphoneKeys`), `nysiis.go` (`NYSIISCode` at 71.7%), and `DefaultScorer`'s panic branch.
- **Rationale:** `coverage-check` is a direct dependency of `make check`. The CI gate will block every PR until the floor is met. The `verify-coverage-floors.sh` script exits immediately on the overall-floor failure, so the per-file floor check (Floor 2) has not yet run; per-file violations exist and will surface as a second failure once the overall floor is fixed (see finding 16).

---

### [Critical] go.sum contains three stale entries removed by go mod tidy

- **File:** `/Users/johnny/Development/fuzzymatch/go.sum`
- **Phase introduced:** Phase 1 (module hygiene carried forward)
- **Issue:** `go mod tidy && git diff --exit-code -- go.mod go.sum` exits 1:
  ```diff
  --- a/go.sum
  +++ b/go.sum
  @@ -1,5 +1,2 @@
  -golang.org/x/mod v0.35.0/go.mod h1:+GwiRhIInF8wPm+4AoT6L0FA1QWAad3OMdTRx4tFYlU=
  -golang.org/x/sync v0.20.0/go.mod h1:9xrNwdLfx4jkKbNva9FpL6vEN7evnE43NNNJQ2LF3+0=
   golang.org/x/text v0.37.0 h1:...
   golang.org/x/text v0.37.0/go.mod h1:...
  -golang.org/x/tools v0.44.0/go.mod h1:KA0AfVErSdxRZIsOVipbv3rQhVXTnlU6UhKxHd1seDI=
  ```
  `go mod tidy` removes the three `/go.mod`-only hash entries for `golang.org/x/mod@v0.35.0`, `golang.org/x/sync@v0.20.0`, and `golang.org/x/tools@v0.44.0`. These are stale indirect-module go.mod hashes from a previous toolchain version of `golang.org/x/text`'s own dependency graph that `go 1.26.3` no longer needs to record.
- **Standard:** `go-coding-standards` SKILL.md §Quality Gate item 6; CLAUDE.md Checks item 7 ("go mod tidy produces no diff")
- **Action:** Code fix — run `go mod tidy` and commit the updated `go.sum`. The `go.mod` itself is unchanged.
- **Rationale:** `tidy-check` is a direct dependency of `make check`. The CI gate blocks on this diff. The fix is a single `go mod tidy` commit.

---

## Important Findings

### [Important] gocyclo — 8 phonetic algorithm functions exceed CC 10 with no nolint directive

- **File:** `double_metaphone.go` (lines 120, 176, 243, 890), `mra.go` (lines 135, 241), `nysiis.go` (line 116), `soundex.go` (line 147)
- **Phase introduced:** Phase 7 (`07-phonetic-algorithms`)
- **Issue:** gocyclo reports cyclomatic complexity exceeding the configured threshold of 10 for eight functions:
  - `dmSlgCheck` CC=12 (`double_metaphone.go:120`)
  - `dmPrep` CC=17 (`double_metaphone.go:176`)
  - `DoubleMetaphoneKeys` CC=280 (`double_metaphone.go:243`)
  - `DoubleMetaphoneScore` CC=13 (`double_metaphone.go:890`)
  - `MRACode` CC=16 (`mra.go:135`)
  - `MRACompare` CC=23 (`mra.go:241`)
  - `NYSIISCode` CC=78 (`nysiis.go:116`)
  - `SoundexCode` CC=29 (`soundex.go:147`)

  None of these have a `//nolint:gocyclo` directive with a justification comment. The codebase already uses this pattern with explicit rationale for edit-distance functions (e.g. `damerau_osa.go:218`, `jaro.go:190`, `algoid.go:213`) — the phonetic files are inconsistent with that established convention.
- **Standard:** `go-coding-standards` SKILL.md §Complexity ("Cyclomatic complexity > 10: refactor before merge"); same file notes "Algorithm DP loops are exempt from the function-length guidance where splitting would obscure the recurrence"
- **Action:** Code fix — add `//nolint:gocyclo` directives with justification comments referencing the originating algorithm paper (Philips 2000 for Double Metaphone, NBS Tech Note 943 for NYSIIS/MRA, Knuth 1973 for Soundex). Refactoring the rule tables into separate functions is an alternative for `NYSIISCode` (CC=78) and `DoubleMetaphoneKeys` (CC=280), but only if it does not obscure the algorithmic recurrence.
- **Rationale:** Without nolint directives, golangci-lint exits 1 (see Critical finding 1). The complexity is inherent to these algorithms' phonetic rule sets and cannot be meaningfully reduced without fragmenting the algorithm logic across files in a way that harms readability and academic cross-reference.
- **Suggested fix:** Add to each function declaration, e.g. for `DoubleMetaphoneKeys`: `func DoubleMetaphoneKeys(s string) (primary, secondary string) { //nolint:gocyclo // Philips (2000) 'Double Metaphone' rule engine; inherently complex per-letter decision tree that cannot be split without losing the one-to-one mapping to §III of the paper`

---

### [Important] gocritic dupBranchBody — double_metaphone.go:744 both if/else branches are identical (potential algorithm correctness issue)

- **File:** `/Users/johnny/Development/fuzzymatch/double_metaphone.go`, line 744
- **Phase introduced:** Phase 7 (`07-phonetic-algorithms`)
- **Issue:** `gocritic` reports `dupBranchBody: both branches in if statement have same body`. The offending code:
  ```go
  // French "SAIS" end
  if i == n-1 && (at(i-1) == 'A' || at(i-1) == 'I') {
      dmAdd(&p, &alt, "S", "")
  } else {
      dmAdd(&p, &alt, "S", "")
  }
  ```
  Both the French-end case (`BAIS`, `MAIS` terminal S) and the general S case produce `("S", "")`. In several Double Metaphone reference implementations (including the original Philips C++ and the Python `metaphone` library), the general-S else branch produces `("S", "S")` — both primary AND secondary are "S" — while the French-end case silently drops the secondary to produce `("S", "")`. If the intended behaviour follows this reference, the else branch should read `dmAdd(&p, &alt, "S", "S")` and the current code silently drops the secondary "S" in all non-French S positions.
- **Standard:** `algorithm-correctness-standards` SKILL.md §Primary Source Citation; `go-coding-standards` SKILL.md §API Design
- **Action:** Discuss-phase needed — the `algorithm-correctness-reviewer` must verify the expected behaviour for the general-S case against Philips (2000) §III before the code is changed. If the else branch is confirmed to be `("S","S")`, this is a silent correctness bug affecting all words with S in a non-final-French position.
- **Rationale:** The `dupBranchBody` finding from gocritic surfaces a code structure that is either dead code (the if condition is meaningless because both branches are identical) or a latent algorithm bug. The Double Metaphone phonetic cross-validation test suite (`phonetic_codes_golden_test.go`) may not exercise the full range of secondary-key paths needed to catch this regression.

---

### [Important] gocritic ifElseChain — double_metaphone.go:716 rewrite if-else to switch

- **File:** `/Users/johnny/Development/fuzzymatch/double_metaphone.go`, line 716
- **Phase introduced:** Phase 7 (`07-phonetic-algorithms`)
- **Issue:** `gocritic` reports `ifElseChain: rewrite if-else to switch statement` at line 716. The three-way `if/else if/else` for the SCH trigram case could be expressed as a `switch` for readability.
- **Standard:** `go-coding-standards` SKILL.md §Complexity ("switch with more than ~6 cases: consider dispatch table")
- **Action:** Code fix — rewrite the `if/else if/else` chain as a `switch` or add a `//nolint:gocritic` directive if the if-chain is clearer for this algorithm rule.
- **Rationale:** This contributes to the golangci-lint exit 1 (Critical finding 1). It is a style issue, not a correctness issue.

---

### [Important] gosec G602 — mra.go:211-213 potential slice index out of range (false positive needing nolint)

- **File:** `/Users/johnny/Development/fuzzymatch/mra.go`, lines 211–213
- **Phase introduced:** Phase 7 (`07-phonetic-algorithms`)
- **Issue:** `gosec` and golangci-lint both report G602 (CWE-118) at lines 211, 212, 213:
  ```go
  result[3] = step2Buf[step2Len-3]   // line 211
  result[4] = step2Buf[step2Len-2]   // line 212
  result[5] = step2Buf[step2Len-1]   // line 213
  ```
  These lines are inside the `if step2Len <= 6` branch's else path (i.e. `step2Len > 6`), which guarantees `step2Len >= 7`. With `step2Len >= 7`, the lowest index accessed is `step2Len-3 >= 4`, which is within the `[64]byte step2Buf` bounds. The `result` array is `[6]byte` and indices 3, 4, 5 are valid. gosec cannot statically prove the `step2Len > 6` guard from the conditional structure, so it fires a false positive.
- **Standard:** `go-coding-standards` SKILL.md §Quality Gate; no specific standard for gosec false-positive suppression
- **Action:** Code fix — add `//nolint:gosec // G602 false positive: step2Len > 6 is guaranteed by the enclosing if guard; all index arithmetic is within [64]byte bounds` on lines 211–213 (or restructure to use a `copy` call that gosec can prove safe).
- **Rationale:** The three G602 instances contribute to golangci-lint exit 1. A brief explanatory comment is essential so the suppression is auditable.

---

### [Important] gofumpt — mra_bench_test.go:30 and nysiis_bench_test.go:28 multiple adjacent var declarations should be grouped

- **File:** `/Users/johnny/Development/fuzzymatch/mra_bench_test.go` (line 30), `/Users/johnny/Development/fuzzymatch/nysiis_bench_test.go` (line 28)
- **Phase introduced:** Phase 7 (`07-phonetic-algorithms`)
- **Issue:** `gofumpt` reports formatting issues in both files. In `mra_bench_test.go`, four consecutive `var` declarations (`mraCodeSink`, `mraMatchedSink`, `mraSimSink`, `mraScoreSink`) are each on separate `var` lines rather than grouped in a `var (...)` block. Same pattern in `nysiis_bench_test.go` for `nysiisSink` and `nysiisScoreSink`. gofumpt requires adjacent package-level `var` declarations to be grouped.
- **Standard:** `go-coding-standards` SKILL.md §Quality Gate item 1 ("gofmt -s and goimports produce no diff"); project uses gofumpt (stricter than gofmt)
- **Action:** Code fix — merge adjacent `var` declarations into grouped `var (...)` blocks in both files:
  ```go
  var (
      mraCodeSink    string
      mraMatchedSink bool
      mraSimSink     int
      mraScoreSink   float64
  )
  ```
- **Rationale:** gofumpt issues contribute to golangci-lint exit 1 (Critical finding 1). The fix is mechanical.

---

### [Important] gofumpt — double_metaphone_test.go:159 struct literal formatting

- **File:** `/Users/johnny/Development/fuzzymatch/double_metaphone_test.go`, line 159
- **Phase introduced:** Phase 7 (`07-phonetic-algorithms`)
- **Issue:** `gofumpt` reports the file is not properly formatted at line 159. The struct literal `{"RV-DM1/Schmidt", "Schmidt", "XMT", "SMT", "germanic",` spans multiple lines in a way that gofumpt's stricter multi-line struct-literal rules do not accept. Running `gofumpt -w double_metaphone_test.go` will apply the canonical formatting.
- **Standard:** `go-coding-standards` SKILL.md §Quality Gate item 1
- **Action:** Code fix — run `gofumpt -w double_metaphone_test.go` (or `make fmt`) and review the result.
- **Rationale:** Contributes to golangci-lint exit 1.

---

### [Important] misspell — scorer_options_internal_test.go:28 "artifact" should be "artefact" (UK English locale)

- **File:** `/Users/johnny/Development/fuzzymatch/scorer_options_internal_test.go`, line 28
- **Phase introduced:** Phase 8 (`08-composite-scorer`)
- **Issue:** `misspell` (locale: UK, as configured in `.golangci.yml`) flags the word "artifact" in a comment as a misspelling of "artefact". The exact comment:
  ```go
  // the build-tag _test.go suffix ensures this file never ships in the
  // public artifact.
  ```
  The project's `documentation-standards` SKILL.md §Language explicitly requires British English ("colour, behaviour, organisation, normalisation…"). "Artefact" is the standard British English spelling.
- **Standard:** `documentation-standards` SKILL.md §Language ("British English for prose")
- **Action:** Code fix — change `artifact` to `artefact` in the comment at line 28.
- **Rationale:** Contributes to golangci-lint exit 1. Also enforces the project-wide British English convention.

---

### [Important] misspell — scorer_golden_test.go:287 intentional test-data string flagged as misspelling

- **File:** `/Users/johnny/Development/fuzzymatch/scorer_golden_test.go`, line 287
- **Phase introduced:** Phase 8 (`08-composite-scorer`)
- **Issue:** `misspell` flags `"abreviation"` (single-`b` deliberate typo used as test input) as a misspelling of `"abbreviation"`:
  ```go
  entries = append(entries, makeScorerGoldenEntry(defaultS, "abbreviation", "abreviation", "DefaultScorer"))
  ```
  The string `"abreviation"` is intentional test data: a one-letter-drop variant of "abbreviation" used to exercise the scorer's fuzzy matching near the threshold boundary. The misspell linter scans string literal content.
- **Standard:** `go-coding-standards` SKILL.md §Quality Gate item 3 (lint must be clean)
- **Action:** Code fix — add a `//nolint:misspell` directive on line 287 with a comment explaining the intentional test data: `//nolint:misspell // "abreviation" is an intentional one-letter-drop test vector (scorer threshold boundary)`.
- **Rationale:** Contributes to golangci-lint exit 1. The misspelling is by design; suppression with explanation is the correct approach.

---

### [Important] staticcheck SA4004 — errors_test.go:209 for-range loop with unconditional break inspects only first rune

- **File:** `/Users/johnny/Development/fuzzymatch/errors_test.go`, line 209
- **Phase introduced:** Phase 1 (`01-foundation-infrastructure`)
- **Issue:** `staticcheck SA4004` reports "the surrounding loop is unconditionally terminated". The code:
  ```go
  for _, r := range body {
      if unicode.IsUpper(r) {
          t.Errorf(...)
      }
      // Only inspect the first rune.
      break
  }
  ```
  The `break` is outside the `if` and unconditional, so the loop always exits after the first iteration regardless of the `if` outcome. This is intentional (the comment says "Only inspect the first rune"), but the loop structure is misleading: `strings.HasPrefix` or `utf8.DecodeRuneInString` would express the intent directly without the loop.
- **Standard:** `go-coding-standards` SKILL.md §Complexity (guard clauses and early returns)
- **Action:** Code fix — replace the for-range loop with a direct rune check:
  ```go
  if r, _ := utf8.DecodeRuneInString(body); unicode.IsUpper(r) {
      t.Errorf(...)
  }
  ```
  Or add a `//nolint:staticcheck` directive if the loop form is preferred for readability.
- **Rationale:** Contributes to golangci-lint exit 1.

---

### [Important] staticcheck QF1001 — De Morgan's law applicable in 6 locations

- **File:** `double_metaphone.go` (lines 397, 825), `double_metaphone_fuzz_test.go` (line 105), `double_metaphone_test.go` (line 322), `props_test.go` (line 3492), `soundex_test.go` (line 203)
- **Phase introduced:** Phase 7 (`07-phonetic-algorithms`); Phase 2 for `soundex_test.go`
- **Issue:** `staticcheck QF1001` suggests applying De Morgan's law to simplify six boolean conditions. The recurring pattern `!((c >= 'A' && c <= 'Z') || c == '0')` should read `(c < 'A' || c > 'Z') && c != '0'`. Similarly `!(i == 1 && v[0] == 'M')` → `i != 1 || v[0] != 'M'`.
- **Standard:** `go-coding-standards` SKILL.md §Complexity (clarity, guard clauses)
- **Action:** Code fix — apply De Morgan's transformations at all six sites. Alternatively, extract the alphabet-check into a named helper (`isValidMetaphoneChar`) to eliminate duplication across `double_metaphone.go`, `double_metaphone_fuzz_test.go`, `double_metaphone_test.go`, and `props_test.go`.
- **Rationale:** Contributes to golangci-lint exit 1. Extracting a helper would also reduce the four near-identical condition copies.

---

### [Important] unparam — soundex.go:275 runeAt returns a rune that is never used at any call site

- **File:** `/Users/johnny/Development/fuzzymatch/soundex.go`, line 275
- **Phase introduced:** Phase 2 (`02-core-character-algorithms-six`) or Phase 7 depending on when Soundex was added
- **Issue:** `unparam` reports `runeAt - result 0 (rune) is never used`. The function signature is `func runeAt(s string, i int) (rune, int)` but every call site only uses the second return value (byte width), discarding the rune. Example call: `_, sz := runeAt(s, i)`. The rune value is computed but always discarded, making the function's first return type unused.
- **Standard:** `go-coding-standards` SKILL.md §API Design (functions return what callers need)
- **Action:** Code fix — change the signature to `func runeAt(s string, i int) int` and update the implementation to return only the byte width. This simplifies call sites to `sz := runeAt(s, i)`. If the rune value is needed for future callers, retain the rune return and add `//nolint:unparam` with a forward-looking comment.
- **Rationale:** Contributes to golangci-lint exit 1. The simpler signature is more honest about what callers actually use.

---

### [Important] unparam — scorer_options_internal_test.go:113 probeScoreFnInvoke always receives i=0

- **File:** `/Users/johnny/Development/fuzzymatch/scorer_options_internal_test.go`, line 113
- **Phase introduced:** Phase 8 (`08-composite-scorer`)
- **Issue:** `unparam` reports `probeScoreFnInvoke - i always receives 0`. The helper function `probeScoreFnInvoke(cfg scorerConfig, i int, a, b string)` accepts an index `i` but every call site passes literal `0`. The function could be simplified to always use index 0.
- **Standard:** `go-coding-standards` SKILL.md §API Design (zero value usability)
- **Action:** Code fix — remove the `i int` parameter and hardcode the index, or add `//nolint:unparam` with a comment if the parameter is retained for future test extensibility.
- **Rationale:** Contributes to golangci-lint exit 1.

---

### [Important] Per-file coverage floor — double_metaphone.go at 83.8% (< 90% floor)

- **File:** `/Users/johnny/Development/fuzzymatch/double_metaphone.go`
- **Phase introduced:** Phase 7 (`07-phonetic-algorithms`)
- **Issue:** File-average statement coverage is 83.8%, below the per-file floor of 90.0%. Function-level breakdown:
  - `dmPrep` at 58.8% — the heap fallback path (`heapPath:` label, names > 64 ASCII chars) and non-ASCII rune handling are not exercised
  - `DoubleMetaphoneKeys` at 62.0% — large portions of the 648-line phonetic rule engine are not covered (many letter-case branches have no test vectors)
  - `dmAdd` at 84.6% — the length-truncation branches (when a phoneme exceeds `dmMaxLen`) are not exercised
  - `dmSlgCheck` at 88.9% — the WITZ substring check is not covered
  Note: the `verify-coverage-floors.sh` script's Floor 2 (per-file) check does NOT currently run because the script exits early on the overall-floor failure (Floor 1). This per-file violation will surface as an additional CI failure once the overall floor is fixed.
- **Standard:** `go-testing-standards` SKILL.md §Coverage Targets ("≥ 90% per file"); CLAUDE.md Constraints
- **Action:** Code fix — add test vectors to `double_metaphone_test.go` and/or `double_metaphone_fuzz_test.go` covering: (1) names longer than 64 ASCII letters to trigger the heap path in `dmPrep`, (2) non-ASCII input to `dmPrep`, (3) phonemes that exceed 4 chars to trigger the `dmAdd` truncation branch, (4) a name ending in "WITZ" for `dmSlgCheck`, and (5) additional phonetic rule cases in `DoubleMetaphoneKeys` for the currently uncovered letter sections.
- **Rationale:** The per-file floor is a CI gate (`verify-coverage-floors.sh` Floor 2). Fixing this also directly contributes to closing the overall-coverage gap (Critical finding 2).

---

### [Important] Major exported functions below 100% statement coverage

- **File:** global (root package)
- **Phase introduced:** Phases 2, 6, 7, 8 (by algorithm introduction phase)
- **Issue:** 18 exported public API functions have statement coverage below 100%. Coverage is measured per `go tool cover -func=coverage.out`. The `verify-coverage-floors.sh` Floor 3 check only enforces non-zero coverage (any exercising test exists), not 100% statement coverage. The SKILL targets 100% on public API. The most significant gaps:

  | Function | File | Coverage |
  |---|---|---|
  | `DoubleMetaphoneKeys` | `double_metaphone.go:243` | 62.0% |
  | `NYSIISCode` | `nysiis.go:116` | 71.7% |
  | `DefaultScorer` | `scorer.go:586` | 75.0% |
  | `DamerauLevenshteinOSADistanceRunes` | `damerau_osa.go:137` | 80.0% |
  | `DamerauLevenshteinOSAScoreRunes` | `damerau_osa.go:183` | 81.8% |
  | `MRACompare` | `mra.go:241` | 91.3% |
  | `DoubleMetaphoneScore` | `double_metaphone.go:890` | 92.3% |
  | `NewScorer` | `scorer.go:180` | 92.6% |
  | `TokenJaccardScore` | `token_jaccard.go:200` | 93.3% |
  | `JaroWinklerScoreRunes` | `jarowinkler.go:154` | 93.8% |
  | `Strcmp95Score` | `strcmp95.go:260` | 93.8% |
  | `HammingScoreRunes` | `hamming.go:170` | 94.1% |
  | `PartialRatioScore` | `partial_ratio.go:277` | 94.1% |
  | `PartialRatioScoreRunes` | `partial_ratio.go:444` | 94.1% |
  | `SoundexCode` | `soundex.go:147` | 94.3% |
  | `DamerauLevenshteinFullScoreRunes` | `damerau_full.go:190` | 90.9% |
  | `LevenshteinScoreRunes` | `levenshtein.go:177` | 90.9% |
  | `Tokenise` | `tokenise.go:143` | 95.5% |

  Notable root causes:
  - `DefaultScorer` at 75%: the `panic` branch (bug-guard when `DefaultScorerOptions()` produces an error) is untested — no test passes invalid options to the constructor underlying `DefaultScorer`.
  - `DamerauLevenshteinOSADistanceRunes` / `DamerauLevenshteinOSAScoreRunes` at 80/82%: the early-identity short-circuit and the rune-slice path have branch gaps in the Unicode path.
  - `NYSIISCode` at 71.7%: large sections of the NYSIIS replacement rules are not covered by current test vectors.

- **Standard:** `go-testing-standards` SKILL.md §Coverage Targets ("100% on public API surface")
- **Action:** Code fix — add targeted test cases for each function's uncovered branches. Priority order: `DoubleMetaphoneKeys` (62%), `NYSIISCode` (71.7%), `DefaultScorer` (75%), `DamerauLevenshteinOSADistanceRunes` (80%).
- **Rationale:** The 100% public API target is stated in both the skill and CLAUDE.md Constraints. While the `verify-coverage-floors.sh` Floor 3 check does not enforce statement-level 100% (it only checks that the function is exercised at all — see finding 17), the standard requires 100% and the test-analyst agent gate should catch this on milestone reviews.

---

### [Important] Coverage floor script Floor 3 enforces non-zero coverage, not 100% statement coverage for public API

- **File:** `/Users/johnny/Development/fuzzymatch/scripts/verify-coverage-floors.sh`, line 14–17
- **Phase introduced:** Phase 1 (`01-foundation-infrastructure`)
- **Issue:** The script's Floor 3 semantics are explicitly documented as weaker than the SKILL's stated target:
  ```
  # Floor #3 semantics
  # 100% public-API coverage is enforced by the EXISTENCE of an exercising
  # test, not by requiring 100.0% statement coverage on the symbol's body.
  ```
  The `go-testing-standards` SKILL.md states "100% on public API surface". The script's interpretation allows `DoubleMetaphoneKeys` at 62% to pass Floor 3, meaning 18 exported functions with 62–95% statement coverage are not caught by any automated CI gate.
- **Standard:** `go-testing-standards` SKILL.md §Coverage Targets ("100% on public API surface")
- **Action:** Discuss-phase needed — either (a) update `verify-coverage-floors.sh` to enforce 100% statement coverage on exported functions (stricter, aligns with skill text), or (b) update the `go-testing-standards` SKILL to acknowledge that "100%" means "fully exercised, non-zero" (weaker, aligns with current script). The script's current behaviour is intentionally documented; the skill may need clarification of intent.
- **Rationale:** The current enforcement gap means 18 public API functions have uncovered branches in production code and no automated gate will fail until the misalignment is resolved.

---

### [Important] Orphaned TODO without real issue number — partial_ratio.go:148

- **File:** `/Users/johnny/Development/fuzzymatch/partial_ratio.go`, line 148
- **Phase introduced:** Phase 6 (`06-token-based-algorithms`)
- **Issue:** The grep check for orphaned TODOs finds:
  ```
  partial_ratio.go:148://  TODO(#TBD): implement sliding-window DP per Bachmann RapidFuzz
  partial_ratio.go:153://  DP implementation; this TODO will be updated with the issue
  ```
  The pattern `#TBD` does not match `(#[0-9]+)`. The CLAUDE.md standard requires every TODO to reference a real GitHub issue number.
- **Standard:** CLAUDE.md §GitHub Issues ("Every TODO must reference a GitHub issue: `// TODO(#42): ...`"); Check item 12 in the automated quality gate
- **Action:** Code fix — create a GitHub issue for the sliding-window DP optimisation and replace `#TBD` with the real issue number. The comment already describes the work (spec line 612, O(|s|·|l|) variant deferred to v1.x) so the issue content is ready.
- **Rationale:** The orphaned-TODO check (`grep -rn "TODO\|FIXME\|HACK\|BUG" --include="*.go" . | grep -v "_test.go" | grep -v "(#[0-9]"`) is part of the project's automated quality gate (item 12 in CLAUDE.md Checks). An unreferenced TODO cannot be tracked or prioritised.

---

## Improvement Findings

### [Improvement] gosec not pre-installed — required manual installation during gate run

- **File:** global / CI configuration
- **Phase introduced:** Phase 1 (`01-foundation-infrastructure`)
- **Issue:** `gosec` was not present on the developer machine at the start of this gate run. The command `gosec ./...` failed with `zsh: command not found: gosec`. It was installed with `go install github.com/securego/gosec/v2/cmd/gosec@latest` before the scan could proceed. The `security.yml` CI workflow presumably installs gosec via its action, but the local developer workflow (`make security`) only invokes `govulncheck` and does not install or invoke gosec locally (Makefile `security` target calls `$(GOVULNCHECK) ./...` only).
- **Standard:** CLAUDE.md "Checks" item 10 ("gosec ./..."); `go-coding-standards` SKILL.md §Quality Gate
- **Action:** Skill clarification — update the `make security` target in `Makefile` to invoke gosec in addition to govulncheck, with the same tolerant "if not installed, print install instructions" pattern already used for govulncheck. Alternatively document that gosec is a CI-only gate and is not expected locally.
- **Rationale:** gosec found 3 real findings (mra.go G602, see Important finding 7) that are now blocked in CI via `security.yml`. Developers running `make check` locally will not see these findings without local gosec installation.

---

### [Improvement] markdownlint-cli2 not installed

- **File:** global / CI configuration
- **Phase introduced:** Phase 1 (`01-foundation-infrastructure`)
- **Issue:** `markdownlint-cli2` is not available locally. The command `which markdownlint-cli2` returns exit 1. The tool is listed as spec-locked in CLAUDE.md "Recommended Stack" and is expected by the CI `ci.yml` workflow.
- **Standard:** CLAUDE.md Recommended Stack ("markdownlint-cli2 v0.22.1 — Markdown linting for README, docs/, CHANGELOG"); CLAUDE.md "Checks" item 11
- **Action:** Improvement — document the `npm install -g markdownlint-cli2@0.22.1` local install step in `CONTRIBUTING.md` or a `docs/dev-setup.md` file. The tool cannot be invoked via `go install` so it requires a separate Node-based install step not currently documented for local developer setup.
- **Rationale:** Without `markdownlint-cli2`, markdown quality is only enforced in CI, not locally. Given the project's extensive `docs/` directory (requirements.md, algorithms.md, etc.), markdown lint drift is likely to accumulate between PR cycles.

---

### [Improvement] scripts/verify-llms-sync.sh absent — llms sync enforced by test only

- **File:** `/Users/johnny/Development/fuzzymatch/scripts/` (file does not exist)
- **Phase introduced:** Phase 1 (`01-foundation-infrastructure`)
- **Issue:** The automated quality gate specifies `bash scripts/verify-llms-sync.sh` (CLAUDE.md Checks item 11, referenced in Makefile as `verify-llms-sync`). The script does not exist in `scripts/`. The Makefile does not have a `verify-llms-sync` target. The llms sync check IS enforced programmatically by `ai_friendly_test.go` (which runs as part of `go test ./...` and passes cleanly), but the standalone script form expected by the gate spec is absent.
- **Standard:** CLAUDE.md §Makefile Targets ("make verify-llms-sync or equivalent"); `documentation-standards` SKILL.md §AI-Friendly Documentation
- **Action:** Improvement — either (a) create `scripts/verify-llms-sync.sh` as a thin wrapper that runs `go test -run TestAIFriendly ./...` (consistent with how the test already enforces the invariant), or (b) add a `verify-llms-sync` Makefile target that invokes the same test. The underlying enforcement by `ai_friendly_test.go` is correct and complete; only the script/target form is missing.
- **Rationale:** The missing script means the gate cannot be invoked standalone (`bash scripts/verify-llms-sync.sh`) as documented. The enforcement is present but through a different mechanism than the spec describes.

---

## Summary

| Check | Command | Exit Code | Outcome |
|---|---|---|---|
| fmt-check (gofmt + goimports) | `make fmt-check` | 0 | PASS |
| go vet | `go vet ./...` | 0 | PASS |
| golangci-lint | `golangci-lint run ./...` | **1** | **FAIL** — 27 issues |
| root tests | `go test -race -count=1 ./...` | 0 | PASS (59s) |
| BDD tests | `cd tests/bdd && go test -race -count=1 ./...` | 0 | PASS (12s) |
| go mod tidy (root) | `go mod tidy && git diff go.sum` | **1** | **FAIL** — 3 stale go.sum entries |
| go mod tidy (BDD) | `cd tests/bdd && go mod tidy && git diff` | 0 | PASS |
| coverage | `go test -race -coverprofile=coverage.out ./...` | 0 | 90.2% |
| coverage floors | `bash scripts/verify-coverage-floors.sh` | **1** | **FAIL** — 90.2% < 95.0% |
| govulncheck | `govulncheck ./...` | 0 | PASS — no vulnerabilities |
| gosec | `gosec ./...` | 0 | 4 issues (all G602 in mra.go, LOW severity) |
| markdownlint-cli2 | — | — | NOT INSTALLED |
| verify-no-runtime-deps | `bash scripts/verify-no-runtime-deps.sh` | 0 | PASS (golang.org/x/text allowlisted) |
| verify-license-headers | `bash scripts/verify-license-headers.sh` | 0 | PASS (165 files) |
| verify-coverage-floors | `bash scripts/verify-coverage-floors.sh` | **1** | **FAIL** (same as coverage check above) |
| verify-llms-sync | — | — | Script absent; `TestAIFriendly` passes |
| verify-determinism | `go test -run TestGolden_ ./...` | 0 | PASS |
| No orphaned TODOs | grep check | — | 1 violation: `partial_ratio.go:148 TODO(#TBD)` |
| No testify in root | grep check | — | PASS (only comment references, no imports) |
| No GPL/LGPL code | grep check | — | PASS (all references are negating: "provenance: none") |
| No log.Fatal/os.Exit | grep check | — | PASS (0 matches in non-test library code) |
| No fmt.Println/Printf | grep check | — | PASS (only in examples/ standalone programs) |
| No import . | grep check | — | PASS |
| No init() functions | grep check | — | PASS |
| No type name stutter | grep check | — | PASS |

</details>

---

## 13. security-reviewer

_Source: `.planning/reviews/security-FINDINGS.md`_

<details>
<summary>Click to expand full report</summary>

---
status: issues_found
agent: security-reviewer
scope: entire codebase (phases 1-8)
reviewed: 2026-05-17T08:30:00Z
finding_counts:
  critical: 2
  important: 11
  improvement: 18
  total: 31
---

# Security Review — fuzzymatch (Phases 1–8)

Reviewer: security-reviewer
Date: 2026-05-17
Scope: every algorithm (23), Normalise, Tokenise, Scorer, supply-chain, workflows.

The Phase 8-only review at
`.planning/phases/08-composite-scorer/08-SECURITY-REVIEW.md` enumerated
12 findings (SEC-01 to SEC-12) covering the Scorer surface. This whole-
codebase review re-asserts the still-unfixed Phase 8 findings (verified
against current source) and extends coverage to the algorithm tier,
Normalise / Tokenise, supply-chain (`go.mod` / `go.sum` / workflow
actions), and the catalogue's algorithmic-complexity profile.

Severity organisational only — every finding is surfaced regardless of
perceived priority per the security-reviewer directive.

The most notable whole-codebase gaps:

- **CRITICAL:** Two consumer-input panics from the Scorer surface are
  still present in source (SEC-01 / SEC-02). Verified at
  `scorer_options.go:381-399` (Tversky α=β=0 accepted at construction;
  panics at Score) and `scorer_options.go:425-446` (Monge-Elkan
  non-allowlisted inner accepted at construction; panics at Score).

- **IMPORTANT:** Four direct-call algorithm panics on programmer error
  remain part of the public surface — five q-gram functions
  (`CosineScore`, `CosineScoreRunes`, `QGramJaccardScore`,
  `QGramJaccardScoreRunes`, `SorensenDiceScore`, `SorensenDiceScoreRunes`,
  `TverskyScore`, `TverskyScoreRunes`) panic on `n < 1`; the two Tversky
  functions also panic on bad α/β; `MongeElkanScore` and
  `MongeElkanScoreSymmetric` panic on a non-allowlisted inner. These are
  documented as direct-call panic-on-programmer-error, but they still
  represent panics-from-consumer-input on the public surface and are
  reachable from arbitrary consumer code. Documented; not gated.

- **IMPORTANT:** No memory-bound on the worst-case DP allocations.
  `DamerauLevenshteinFullDistance` allocates a flat `(m+2)*(n+2)` int
  slice. On adversarial input pair (m = n = 50,000) this is 2.5×10^9
  ints ≈ 20 GB on a 64-bit Go runtime — Go panics with `runtime error:
  makeslice: cap out of range` (recoverable but disruptive), and on
  smaller-but-still-large pairs (e.g. m = n = 10,000) the algorithm
  silently allocates 800 MB and burns CPU for seconds. None of the DP
  algorithms (Levenshtein, OSA, Full, SWG, LCSStr, Ratcliff-Obershelp)
  has an input-size guard. `docs/performance.md` is a 60-line scaffold.

- **IMPORTANT:** `go list -m all` reports indirect modules `x/mod`,
  `x/sync`, `x/tools` in `go.sum` (via `x/text`'s own go.mod). These do
  not enter the compiled binary but are part of the supply-chain trust
  set; `scripts/verify-no-runtime-deps.sh` filters them out by design.
  Their `/go.mod` hashes in `go.sum` are still load-bearing.

- **IMPORTANT:** Five third-party action and tool dependencies pin to
  `@latest` (DavidAnson/markdownlint-cli2-action, govulncheck@latest in
  two places, goimports@latest, benchstat@latest, anchore/sbom-action@v0
  uses major-tag floating). Each of these is a supply-chain attack
  surface that a malicious upstream tag could exploit during a
  release-build window.

- **IMPORTANT:** Scorer fuzz harness still does not exist (re-asserting
  SEC-05). Whole-codebase verification: `ls scorer_fuzz_test.go` returns
  ENOENT.

- **IMPORTANT:** Ratcliff-Obershelp uses unbounded Go call-stack
  recursion (`roMatchedLength` and `roMatchedLengthRunes` —
  `ratcliff_obershelp.go:200-210` and `:279`). The depth is O(min(la,
  lb)); for multi-megabyte adversarial input the stack can grow into
  the hundreds of megabytes. Per the security-reviewer focus area, "no
  algorithm has unbounded recursion" — this is the only one that does.

The remaining findings are improvements / defence-in-depth.

---

## CRITICAL — panic from consumer input

### CR-01: WithTverskyAlgorithm permits α=β=0 → panic on first Score call

- **File:** `/Users/johnny/Development/fuzzymatch/scorer_options.go:381-399`
- **Phase introduced:** Phase 8
- **Issue:** Option layer validates `alpha < 0 || beta < 0` but does not
  validate the `α + β > 0` constraint. `TverskyScore` panics at
  `tversky.go:241-242` on the same condition. A consumer who passes
  `WithTverskyAlgorithm(1.0, 0, 0, 3)` constructs a Scorer successfully
  but every non-identical `Score`, `Match`, or `ScoreAll` call panics.
  The identity short-circuit at `tversky.go:231` (`if a == b { return
  1.0 }`) hides the panic in trivial test inputs, making the defect
  latent.
- **Standard:** `.claude/skills/go-coding-standards/SKILL.md` — panic
  discipline; `.claude/skills/fuzzymatch-review-protocol/SKILL.md` —
  consumer-input panic forbidden.
- **Action:** Code fix.
- **Rationale:** A panic that surfaces only on non-identical input
  classes as a latent denial-of-service: a Scorer that survives unit-
  test smoke (identical-pair calls) but kills the goroutine on the first
  real query. `ErrInvalidTverskyParam` already exists in `errors.go:88`
  for this case.
- **Suggested fix:** Mirror the direct-call panic check at the option
  layer:
  ```go
  if alpha < 0 || beta < 0 || (alpha == 0 && beta == 0) {
      return ErrInvalidTverskyParam
  }
  ```
  Add `TestWithTverskyAlgorithm_RejectsBothZero` unit test plus a BDD
  scenario.

(Re-asserts Phase 8 SEC-01 / 08-REVIEW.md CR-02 — defect remains.)

---

### CR-02: WithMongeElkanAlgorithm permits non-allowlisted inner → panic on first Score call

- **File:** `/Users/johnny/Development/fuzzymatch/scorer_options.go:425-446`
- **Phase introduced:** Phase 8
- **Issue:** Option layer checks dispatch-table bounds and the trivial
  self-recursion case (`inner == AlgoMongeElkan`) but does NOT consult
  `permittedMongeElkanInner` (`monge_elkan.go:291-317`). Four AlgoIDs
  pass the option's bounds check but trigger the
  `MongeElkanScoreSymmetric` panic at `monge_elkan.go:382` on the first
  call: `AlgoTokenSortRatio`, `AlgoTokenSetRatio`, `AlgoPartialRatio`,
  `AlgoTokenJaccard`. As with CR-01 the identity short-circuit at
  `monge_elkan.go:387` hides the panic on trivial inputs, making the
  defect latent.
- **Standard:** `.claude/skills/go-coding-standards/SKILL.md` — panic
  discipline; `.claude/skills/fuzzymatch-review-protocol/SKILL.md`.
- **Action:** Code fix.
- **Rationale:** Same class as CR-01 — latent panic-from-consumer-input
  reachable through the public Scorer surface, with the worst-case
  failure mode being a goroutine crash mid-query.
- **Suggested fix:**
  ```go
  if !permittedMongeElkanInner[inner] {
      return ErrInvalidAlgorithm
  }
  ```
  inserted after the existing bounds + self-recursion check. Export the
  allow-list as `permittedMongeElkanInner` is already package-scoped at
  `monge_elkan.go:291` — no new export required. Add unit test
  iterating the 4 currently-rejected AlgoIDs and asserting the option
  returns `ErrInvalidAlgorithm`.

(Re-asserts Phase 8 SEC-02 / 08-REVIEW.md IN-03 — defect remains.)

---

## IMPORTANT — DoS, info leakage, recursion-depth unbounded, supply-chain

### IM-01: WithThreshold accepts NaN → every Match silently returns false

- **File:** `/Users/johnny/Development/fuzzymatch/scorer_options.go:257-266`
- **Phase introduced:** Phase 8
- **Issue:** `WithThreshold(t)` validates `t < 0.0 || t > 1.0`; both
  comparisons evaluate `false` for `t = math.NaN()`. The NaN threshold
  is frozen into the Scorer and `Match(a, b)` returns
  `s.Score(a, b) >= s.threshold`. `x >= NaN` is always false, so the
  Scorer silently matches nothing — a denial-of-service via wrong-answer
  with no error, no log, no warning.
- **Standard:** `.claude/skills/determinism-standards/SKILL.md` §13.3 —
  NaN handling; `.claude/skills/fuzzymatch-review-protocol/SKILL.md`.
- **Action:** Code fix.
- **Rationale:** A silent-wrong-answer is worse than a panic — every
  downstream decision is corrupted invisibly. A consumer who JSON-
  decodes a threshold from a YAML/JSON config or arithmetic computes one
  from a `math.Sqrt(-1)` mistake gets a non-functional Scorer with no
  signal.
- **Suggested fix:**
  ```go
  import "math"
  if math.IsNaN(t) || t < 0.0 || t > 1.0 {
      return ErrInvalidThreshold
  }
  ```
  plus `TestWithThreshold_RejectsNaN`.

(Re-asserts Phase 8 SEC-03 / 08-REVIEW.md CR-01 — defect remains.)

---

### IM-02: WithAlgorithm + parameterised options accept NaN/+Inf weight → Score returns NaN

- **File:** `/Users/johnny/Development/fuzzymatch/scorer_options.go:150-165`
  (WithAlgorithm) plus the same pattern in every parameterised With*
  option at lines 300-315, 325-340, 350-365, 381-399, 425-446, 465-479
- **Phase introduced:** Phase 8
- **Issue:** The weight gate `if weight <= 0` evaluates `false` for both
  `math.NaN()` and `math.Inf(+1)`. The poisoned weight propagates
  through auto-normalisation: `weight = NaN` produces `sum = NaN`, the
  defensive `sum == 0` check at `scorer.go:284` returns `false` for NaN
  (NaN compares unequal to any value including itself), normalised
  weights become `NaN/NaN = NaN`, every `Score` call returns NaN, every
  `Match` returns `false`, every `ScoreAll` populates the result map
  with NaN values. `weight = +Inf` produces the same NaN-poison chain.
  The `TestProp_Scorer_NoNaN_NoInf` property test (`scorer_test.go:877`)
  exercises `DefaultScorer()` only — it does not exercise the NaN/Inf-
  weight construction path, so this defect is not caught.
- **Standard:** `.claude/skills/determinism-standards/SKILL.md` §13.3.
- **Action:** Code fix.
- **Rationale:** Silent-wrong-answer denial-of-service, same severity
  class as IM-01.
- **Suggested fix:** Tighten every parameterised With* option's weight
  gate:
  ```go
  if math.IsNaN(weight) || math.IsInf(weight, 0) || weight <= 0 {
      return ErrInvalidWeight
  }
  ```
  AND defence-in-depth tighten `scorer.go:284` to:
  ```go
  if sum <= 0 || math.IsNaN(sum) || math.IsInf(sum, 0) {
      return nil, ErrInvalidWeight
  }
  ```

(Re-asserts Phase 8 SEC-04 — defect remains.)

---

### IM-03: No Scorer-level fuzz harness covering aggregate dispatch

- **File:** missing — `/Users/johnny/Development/fuzzymatch/scorer_fuzz_test.go`
  does not exist (verified)
- **Phase introduced:** Phase 8
- **Issue:** Every catalogue algorithm has a dedicated `Fuzz*` harness
  (26 fuzzers total at `testdata/fuzz/Fuzz*`), and `FuzzNormalise` +
  `FuzzTokenise` cover those primitives. But there is no fuzz harness
  exercising `DefaultScorer().Score(a, b)`, `NewScorer(...)`,
  `Match`, or `ScoreAll` end-to-end. This matters because:
  (1) the Scorer chains Normalise + six algorithms in a single call —
  inter-algorithm interaction failures are not fuzz-covered;
  (2) the CR-01 / CR-02 panics surface only at Score-time, a class a
  Scorer-level fuzz harness would catch automatically;
  (3) the "Scorer is panic-free" claim is currently verified by
  induction over individual algorithms; a fuzz harness makes it directly
  testable.
- **Standard:** `.claude/skills/go-testing-standards/SKILL.md` — every
  public function gets a fuzz harness; `.claude/skills/fuzzymatch-review-
  protocol/SKILL.md`.
- **Action:** Code fix.
- **Rationale:** This is the gate that would have caught CR-01 and CR-02
  in CI.
- **Suggested fix:** Add `scorer_fuzz_test.go` with three harnesses:
  - `FuzzScorer_DefaultScorer_NeverPanics` — exercises `DefaultScorer`'s
    Score / Match / ScoreAll on arbitrary `(a, b)` pairs.
  - `FuzzScorer_NewScorer_NeverPanics` — exercises `NewScorer` with
    arbitrary `(threshold, weight, algoIdx)` triples to catch
    construction-time panics.
  - `FuzzScorer_AllInnerMetricsForME_NeverPanics` — iterates every
    AlgoID as ME's inner across `WithMongeElkanAlgorithm`.

Seed the corpus with embedded NUL, lone surrogates, multi-MB inputs, RTL
marks, zero-width joiners. Wire into `make test-fuzz` and CI's nightly
fuzz workflow.

(Re-asserts Phase 8 SEC-05 — gap remains.)

---

### IM-04: Ratcliff-Obershelp uses unbounded Go call-stack recursion

- **File:** `/Users/johnny/Development/fuzzymatch/ratcliff_obershelp.go:200-210`
  (byte path), `:271-280` (rune path)
- **Phase introduced:** Phase 4
- **Issue:** `roMatchedLength` and `roMatchedLengthRunes` recurse via
  ```go
  return n + roMatchedLength(a[:aLo], b[:bLo]) +
             roMatchedLength(a[aHi:], b[bHi:])
  ```
  with recursion depth bounded by O(min(la, lb)) per the file header
  comment at `:81-83`. On pathological multi-megabyte inputs (e.g. an
  all-'a' string with strategic differences forcing recursion to maximum
  depth at every level), the Go goroutine stack can grow into hundreds
  of megabytes. While Go's growable stacks make a literal stack-overflow
  panic unlikely under default limits, the security-reviewer focus area
  explicitly states "Ratcliff-Obershelp's recursive longest-common-
  substring decomposition must use iterative or bounded-depth recursion".
- **Standard:** Security-reviewer focus area (this prompt) — explicit
  callout. `.claude/skills/fuzzymatch-review-protocol/SKILL.md`.
- **Action:** Code fix (or document-and-bound).
- **Rationale:** Violates the only "no unbounded recursion" rule in the
  threat model. While not a real-world DoS today, a future
  fuzz-generated input could trigger pathological stack growth.
- **Suggested fix:** Convert `roMatchedLength` (and its rune sibling) to
  an iterative form with an explicit work-queue. Pattern:
  ```go
  type roWork struct { aLo, aHi, bLo, bHi int }
  func roMatchedLength(a, b string) int {
      stack := []roWork{{0, len(a), 0, len(b)}}
      var total int
      for len(stack) > 0 {
          w := stack[len(stack)-1]
          stack = stack[:len(stack)-1]
          as, bs := a[w.aLo:w.aHi], b[w.bLo:w.bHi]
          if len(as) == 0 || len(bs) == 0 { continue }
          aLo, aHi, bLo, bHi, n := roFindLongestMatch(as, bs)
          if n == 0 { continue }
          total += n
          stack = append(stack,
              roWork{w.aLo, w.aLo + aLo, w.bLo, w.bLo + bLo},
              roWork{w.aLo + aHi, w.aHi, w.bLo + bHi, w.bHi})
      }
      return total
  }
  ```
  Alternative (lighter touch): document the recursion-depth bound,
  add a fuzz-corpus seed designed to maximise depth, gate wall-time at
  N seconds.

(Re-asserts Phase 8 SEC-07 — concern remains.)

---

### IM-05: Damerau-Levenshtein Full allocates O(m·n) DP table — adversarial pair triggers GB-class allocation

- **File:** `/Users/johnny/Development/fuzzymatch/damerau_full.go:238-239`
  (byte path), `:346-348` (rune path)
- **Phase introduced:** Phase 2
- **Issue:** `damerauFullDP` allocates a flat `(m+2)*(n+2)` int slice
  unconditionally. On 64-bit Go this is 8 bytes per int, so an
  adversarial pair with m = n = 10,000 allocates ≈ 800 MB before
  computing anything; m = n = 50,000 attempts ≈ 20 GB and triggers
  `runtime error: makeslice: cap out of range` (recoverable but
  disruptive). Every other DP algorithm (Levenshtein, OSA, SWG, LCSStr)
  uses O(min(m,n)) two-row DP — Full is the outlier because the
  Lowrance-Wagner transposition lookup needs the historical row
  position. The file godoc even mentions a "v1.x two-row optimisation
  plan" (`damerau_full.go:113-114`) but it has not landed.
- **Standard:** `.claude/skills/performance-standards/SKILL.md` —
  allocation budgets; `.claude/skills/fuzzymatch-review-protocol/SKILL.md`
  — algorithmic-complexity DoS.
- **Action:** Code fix or document-and-warn.
- **Rationale:** Every DP-based algorithm carries O(mn) time complexity
  but most carry only O(min(m,n)) space. Full is uniquely vulnerable to
  pathological-input memory exhaustion. Calling Full from inside a
  custom Scorer composition (`WithAlgorithm(AlgoDamerauLevenshteinFull,
  1.0)`) amplifies the attack: a single `Score` call on a malicious
  pair burns a worker thread's heap allocation budget on the DP table
  alone.
- **Suggested fix (defence-in-depth, no algorithmic change required):**
  Add a `if int64(m)*int64(n) > MaxDPCells { ... }` guard at the head of
  `damerauFullDP` (and its rune-slice sibling). Return distance =
  max(m, n) on overflow — equivalent to "no common substring found"
  semantically and matching the convention of returning the worst case
  on degenerate input. Surface a typed sentinel error at the rune
  surface, or document the soft-bound in godoc. Alternative: implement
  the spec-deferred two-row optimisation (file godoc lines 113-114).

---

### IM-06: DP-based algorithms have no input-size ceiling — composite Scorer multiplies the latency hit

- **File:** All DP files:
  - `/Users/johnny/Development/fuzzymatch/levenshtein.go:115-119`
  - `/Users/johnny/Development/fuzzymatch/damerau_osa.go:125-127`
  - `/Users/johnny/Development/fuzzymatch/damerau_full.go:238-239`
  - `/Users/johnny/Development/fuzzymatch/swg.go:347-350`
  - `/Users/johnny/Development/fuzzymatch/lcsstr.go:150`
  - `/Users/johnny/Development/fuzzymatch/ratcliff_obershelp.go:237`
  - `/Users/johnny/Development/fuzzymatch/jaro.go` (O(la·w))
  - `/Users/johnny/Development/fuzzymatch/strcmp95.go`
  - `/Users/johnny/Development/fuzzymatch/monge_elkan.go:418-426`
    (O(|tA|·|tB|·cost(inner)))
  - `/Users/johnny/Development/fuzzymatch/partial_ratio.go:336-353`
    (O(|s|·|l|·max(|s|,|l|)))
- **Phase introduced:** Phases 2-6
- **Issue:** Every super-linear algorithm documents its complexity in
  its file godoc and several (Monge-Elkan, Partial Ratio) have an
  explicit "DoS notice" section. But there is no `docs/performance.md`
  guidance for consumers, no `WithMaxInputBytes` Scorer option, and no
  upper-bound enforcement anywhere in the library. A
  `DefaultScorer().Score(100KB-input, 100KB-input)` call burns at least
  six algorithms' worst-case work on the same pair; if the composition
  includes Full DL the cost is O(10^10) cells.
- **Standard:** `.claude/skills/fuzzymatch-review-protocol/SKILL.md` —
  "the `docs/performance.md` discusses how to bound input size before
  invoking an algorithm".
- **Action:** Docs + (optional) code fix.
- **Rationale:** This is the largest pre-1.0 documentation gap. Adopters
  in untrusted-input contexts (HTTP API surface, file-upload pipeline)
  need clear guidance.
- **Suggested fix:**
  1. Populate `docs/performance.md` with a per-algorithm complexity +
     recommended-input-ceiling table (the 60-line scaffold is currently
     all TBD per Phase 8 SEC-06).
  2. Add a `## DoS / Resource Bounds` section to `docs/scorer.md` cross-
     referencing each algorithm's complexity docstring.
  3. Consider a `WithMaxInputBytes(n int) ScorerOption` that returns
     `ErrInputTooLarge` from `Score` when either input exceeds `n`
     bytes. Defence-in-depth; not BLOCKING; depends on api-ergonomics-
     reviewer sign-off.

(Re-asserts Phase 8 SEC-06 — gap remains.)

---

### IM-07: Direct-call algorithm panics on programmer error are part of the public API surface

- **Files:**
  - `/Users/johnny/Development/fuzzymatch/cosine.go:197, 237`
    (`CosineScore`, `CosineScoreRunes` — panic on `n < 1`)
  - `/Users/johnny/Development/fuzzymatch/qgram_jaccard.go:146, 184`
    (panic on `n < 1`)
  - `/Users/johnny/Development/fuzzymatch/sorensen_dice.go:160, 198`
    (panic on `n < 1`)
  - `/Users/johnny/Development/fuzzymatch/tversky.go:235, 242, 281, 284`
    (panic on `n < 1` AND on bad α/β)
  - `/Users/johnny/Development/fuzzymatch/monge_elkan.go:382`
    (`MongeElkanScore` and `MongeElkanScoreSymmetric` panic on
    non-allowlisted inner)
- **Phase introduced:** Phases 5-6
- **Issue:** The library's contract is "direct-call panic on programmer
  error; Scorer surface returns typed sentinels". This is documented
  consistently across the q-gram and Monge-Elkan files. However it
  remains true that every panic listed is a panic-from-consumer-input on
  the public API — a consumer who passes `n = 0` to `CosineScore`
  receives a panic, not an error. The security-reviewer focus area is
  "every public function MUST NOT panic on arbitrary input"; the library
  exempts `n < 1` and bad α/β under the "programmer error" interpretation,
  but this exemption is not actually surfaced in the public-facing
  README or `docs/algorithms.md`.
- **Standard:** Security-reviewer focus area; `.claude/skills/go-coding-
  standards/SKILL.md` — error-vs-panic discipline.
- **Action:** Discuss-phase needed (and document if accepted).
- **Rationale:** The Scorer layer is gated; the direct surface is not.
  This is a deliberate-but-undocumented decision. Either:
  (a) Convert direct-call panics to typed errors at the public surface
      (breaking API change for v1.0); OR
  (b) Promote the "panics on programmer error" discipline to a
      first-class section of `docs/algorithms.md`, `README.md`, and
      every file's package-godoc, including a checklist of which
      inputs panic which function.
- **Suggested fix:** Option (b) — document the panic surface
  exhaustively. The user-guide-reviewer agent should review for
  completeness; the api-ergonomics-reviewer agent should rule on
  whether option (a) is desired for v1.0.

---

### IM-08: Five third-party action and tool dependencies pin to floating tags (`@latest` / `@v0`)

- **File:** `/Users/johnny/Development/fuzzymatch/.github/workflows/ci.yml:70`
  (`govulncheck@latest`), `:71` (`goimports@latest`), `:111`
  (`benchstat@latest`), `:129`
  (`DavidAnson/markdownlint-cli2-action@latest-stable`);
  `/Users/johnny/Development/fuzzymatch/.github/workflows/security.yml:33`
  (`govulncheck@latest`); `/Users/johnny/Development/fuzzymatch/.github/workflows/release.yml:63`
  (`anchore/sbom-action/download-syft@v0` — major-tag floating)
- **Phase introduced:** Phase 1
- **Issue:** Floating tags resolve at workflow run-time, exposing each
  build to whatever version the upstream maintainer (or a hostile-tag-
  push attacker) most recently published. For the release workflow this
  is especially load-bearing because the cosign-installed binary and the
  Syft SBOM generator BOTH gate the supply-chain trust chain — a
  compromised upstream release tag would propagate signatures and SBOMs
  that consumers transitively trust.
- **Standard:** `.claude/skills/fuzzymatch-review-protocol/SKILL.md` —
  supply-chain dependency pinning; CLAUDE.md "supply-chain integrity".
- **Action:** Code fix.
- **Rationale:** Mainstream Go security practice is to pin every action
  to a SHA or a specific semver, especially in `release.yml`. The
  `securego/gosec@v2.25.0` and `sigstore/cosign-installer@v3` already
  pin specifically; the inconsistency suggests an oversight rather than
  a deliberate policy.
- **Suggested fix:**
  - Pin `govulncheck`, `goimports`, `benchstat` to specific semver tags
    (Go modules support this — `@v1.1.4` style).
  - Pin `DavidAnson/markdownlint-cli2-action` to a specific semver tag
    (`@v18` or current).
  - Pin `anchore/sbom-action` to a SHA or specific semver tag
    (`@v0.17.4`, not `@v0`).
  - Add `dependabot.yml` group for action updates so future bumps land
    via PR review.

---

### IM-09: `go.sum` ships indirect-module hashes for `x/mod`, `x/sync`, `x/tools` — supply-chain surface beyond `x/text`

- **File:** `/Users/johnny/Development/fuzzymatch/go.sum`
- **Phase introduced:** Phase 1
- **Issue:** `go.sum` records hashes for `golang.org/x/mod`,
  `golang.org/x/sync`, `golang.org/x/tools` (only `/go.mod` hashes — no
  `h1:` source hashes since these don't enter the binary). These come
  from `golang.org/x/text`'s own `go.mod` requirements. They DO factor
  into Go's checksum database verification at `go mod download` time,
  so a compromised module-proxy serving a wrong-hash version of any of
  these three would break the build (which is the intended detection
  mechanism). However, the supply-chain trust set is broader than
  `scripts/verify-no-runtime-deps.sh` suggests — the script's design
  explicitly filters these out, which is correct for runtime-dep
  enforcement but should not be confused with "no supply-chain risk".
- **Standard:** `.claude/skills/fuzzymatch-review-protocol/SKILL.md` —
  supply-chain integrity.
- **Action:** Docs.
- **Rationale:** Informational — the script's design is correct, but the
  documentation framing "zero runtime dependencies" can mislead
  consumers into believing the trust set is `{stdlib, x/text}` when in
  practice it includes the transitive go.mod-graph closure of x/text.
- **Suggested fix:** Add a paragraph to `docs/performance.md` or a new
  `docs/supply-chain.md` clarifying which modules are in the trust set
  (those with hashes in `go.sum`) vs which actually compile into the
  artefact (the allowlist).

---

### IM-10: Cosign keyless trust chain anchors on GitHub OIDC — single point of trust

- **File:** `/Users/johnny/Development/fuzzymatch/.github/workflows/release.yml:75-82`
- **Phase introduced:** Phase 1
- **Issue:** The release workflow signs `checksums.txt` via
  `cosign sign-blob --bundle ... --oidc-issuer
  https://token.actions.githubusercontent.com`. The signature's
  authenticity rests on three trust assumptions:
  (1) GitHub's OIDC issuer at `token.actions.githubusercontent.com` is
      not compromised;
  (2) The Fulcio CA's signing certificates have not been compromised;
  (3) The Rekor transparency log has not been tampered with.
  Cosign keyless is the modern best practice and Sigstore's recommended
  default (since v3.0.1 removed the `COSIGN_EXPERIMENTAL` flag), but
  consumers verifying a release should be aware that the workflow's
  trust chain is purely OIDC + Sigstore, with no offline-key fallback.
- **Standard:** `.claude/skills/fuzzymatch-review-protocol/SKILL.md` —
  supply-chain.
- **Action:** Docs.
- **Rationale:** Informational — the design follows current Sigstore
  best practice, but consumers in regulated industries may need an
  offline-keys path. This is a v2.x decision, not a v1.0 BLOCKING.
- **Suggested fix:** Add a `SECURITY.md` section "Trust chain and
  verification" listing the cosign verify command and the trust
  assumptions. (`SECURITY.md` exists but does not yet cover this.)

---

### IM-11: Allocation amplification — composite Scorer hits every algorithm's allocation path

- **File:** `/Users/johnny/Development/fuzzymatch/scorer.go:368-381`
  (reduction loop)
- **Phase introduced:** Phase 8
- **Issue:** A single `s.Score(a, b)` call on `DefaultScorer()`
  dispatches to six algorithms (DamerauLevenshteinOSA, JaroWinkler,
  TokenJaccard, QGramJaccard, SorensenDice, DoubleMetaphone). Each
  algorithm runs its allocation path independently — there is no
  per-input caching across the dispatch list. For 100KB-vs-100KB input
  the six algorithms together allocate:
  - DamerauLevenshteinOSA: 3 × ~100K ints ≈ 2.4 MB (three-row DP).
  - JaroWinkler: 2 × [256]bool stack-allocated (negligible).
  - TokenJaccard: 2 tokenisations (~rune-count maps).
  - QGramJaccard: 2 map[string]int with capacity (len(s)-n+1).
  - SorensenDice: same as QGramJaccard.
  - DoubleMetaphone: two strings.Builder, bounded to 4 chars each
    (negligible).
  Total: ~5 MB heap pressure per `Score` call. ScoreAll has the same
  per-algorithm allocation set. Under 1000-call-per-second adversarial
  load, the Go runtime's GC handles this fine — but the lack of any
  cross-algorithm input-reuse means future optimisations (e.g. cache
  tokenisation across algorithms) are blocked by the dispatch-table
  abstraction.
- **Standard:** `.claude/skills/performance-standards/SKILL.md` —
  allocation discipline; security-reviewer focus area on resource
  amplification.
- **Action:** Discuss-phase needed.
- **Rationale:** Defence-in-depth — not a real-world DoS today, but a
  Score-level allocation-bound option would constrain adversarial
  worst-case heap pressure. The composite Scorer is the natural place
  to amortise repeated work across algorithms.
- **Suggested fix:** Future plan — explore a per-Score allocation
  budget option `WithMaxScoreAllocBytes(n int) ScorerOption` or a
  cross-algorithm tokenisation cache in `Scorer` (built lazily, scoped
  to the single `Score` call).

---

## Improvement — defence-in-depth, documentation gaps, error tightening

### IMP-01: q-gram map capacity hint can request `len(s)+1` slots — adversarial empty-s + huge-n attack

- **File:** `/Users/johnny/Development/fuzzymatch/q_gram.go:111`, `:141`
- **Phase introduced:** Phase 5
- **Issue:** `extractQGrams` runs `make(map[string]int, len(s)-n+1)`.
  When `len(s) >= n` and both are large, the capacity hint is correct
  and the map grows once. When `len(s) < n` the function returns
  immediately (line 105). However, the variable expression
  `len(s)-n+1` is unsigned-safe by virtue of the early return — Go's
  `make` with a negative hint panics, but the early-return guards it.
  Defence-in-depth check: ensure the early-return is BEFORE the make,
  which it is (line 105 ≪ 111). No defect today.
- **Standard:** `.claude/skills/go-coding-standards/SKILL.md` — slice/map
  capacity safety.
- **Action:** Verified clean; no action.
- **Rationale:** Spot-check confirms the guard.

---

### IMP-02: Partial Ratio rune-path charSet sized to `len(shorter)` — adversarial all-distinct-runes input

- **File:** `/Users/johnny/Development/fuzzymatch/partial_ratio.go:492`
- **Phase introduced:** Phase 6
- **Issue:** `charSet := make(map[rune]struct{}, m)` where m =
  `len(shorter)`. On a 100K-rune adversarial input with every rune
  distinct, the map grows to 100K entries. The byte-path equivalent
  (line 341) uses a stack-allocated `[256]bool` — bounded by definition.
  Rune-path map is bounded by `m`, which is the smaller-input length, so
  the worst case is still O(min(la, lb)) entries. Documented in the
  file header (Allocation budget section, `:271-276`).
- **Standard:** `.claude/skills/performance-standards/SKILL.md`.
- **Action:** No action (documented and bounded).
- **Rationale:** Spot-check confirms the bound.

---

### IMP-03: Damerau-Levenshtein Full rune path allocates unbounded `map[rune]int` da-table

- **File:** `/Users/johnny/Development/fuzzymatch/damerau_full.go:362`
- **Phase introduced:** Phase 2
- **Issue:** `da := make(map[rune]int)` with no capacity hint. Grows
  unbounded as rune characters are seen in `ra`. For an adversarial
  input where every rune is distinct (e.g. random Unicode), the map
  grows to `len(ra)` entries. Document at file header notes this is
  "unavoidable for Unicode" (`:330`). Bounded by `len(ra)` which is the
  larger dimension after the swap.
- **Standard:** `.claude/skills/performance-standards/SKILL.md`.
- **Action:** Improvement.
- **Rationale:** Add the capacity hint
  `make(map[rune]int, len(ra))` — same as `partial_ratio.go:492`. Avoids
  the 4-5 rehash allocations on medium-to-long inputs.

---

### IMP-04: Normalise's per-call `transform.Transformer` chain construction

- **File:** `/Users/johnny/Development/fuzzymatch/normalise.go:319-336`
- **Phase introduced:** Phase 1
- **Issue:** `applyUnicodeTransformer` constructs the chain
  `transform.Chain(norm.NFD, runes.Remove(runes.In(unicode.Mn)),
  norm.NFC)` on every call to Normalise (when `StripDiacritics` is
  true). The file header (`:29-31`) explicitly justifies this:
  "transform.Transformer is not documented as safe for concurrent
  reuse, and per-call construction is cheap; pooling is deferred to a
  v1.x perf revisit". Adversarial-input angle: under 1000-call-per-
  second adversarial load, each call performs the chain construction
  even when the chain is identical across calls. This is a known
  performance footprint, not a security defect, but worth noting for
  the v1.x perf revisit.
- **Standard:** `.claude/skills/performance-standards/SKILL.md`.
- **Action:** No action (documented and bounded).
- **Rationale:** Documented design decision.

---

### IMP-05: SWG accepts NaN / +Inf params and produces "deterministic-but-meaningless" output

- **File:** `/Users/johnny/Development/fuzzymatch/swg.go:171-173`
- **Phase introduced:** Phase 3
- **Issue:** `SmithWatermanGotohScoreWithParams` documents (lines
  171-173): "No validation is performed on params: nonsense values
  (e.g. GapOpen > 0, NaN, +Inf) produce a deterministic-but-meaningless
  result." A consumer passing `SWGParams{Match: math.NaN()}` to
  `WithSmithWatermanGotohAlgorithm` from the Scorer surface gets a
  NaN-polluted Scorer — same class as IM-02.
- **Standard:** `.claude/skills/determinism-standards/SKILL.md` §13.3 —
  NaN handling.
- **Action:** Code fix.
- **Rationale:** Same NaN-poison class as IM-02 but on the SWG params
  surface. The Scorer option `WithSmithWatermanGotohAlgorithm`
  (`scorer_options.go:465-479`) does not validate `params`.
- **Suggested fix:** Add a `SWGParams.Validate()` method returning
  `ErrInvalidConfiguration` on NaN/Inf or invalid sign convention. Call
  it from `WithSmithWatermanGotohAlgorithm`. Direct-call surfaces
  retain the documented "GIGO" behaviour.

---

### IMP-06: `errors.go` sentinels are flat `errors.New` — no per-call context

- **File:** `/Users/johnny/Development/fuzzymatch/errors.go:48-162`
- **Phase introduced:** Phase 1
- **Issue:** Every sentinel is a flat `errors.New("fuzzymatch: ...")`
  value. `NewScorer` and the With* options return the sentinel
  verbatim with no wrapping. Spot-check confirms no error message
  embeds user input (good — no info-leakage risk). However, a consumer
  receiving `ErrInvalidWeight` from `NewScorer(opts...)` has no way to
  tell WHICH option failed without the option-index in the wrapper.
- **Standard:** `.claude/skills/go-coding-standards/SKILL.md` — error
  wrapping conventions.
- **Action:** Improvement.
- **Rationale:** Error-message tightening, not a security defect.
  Useful for debuggability under adversarial input.
- **Suggested fix:** Wrap option errors with the option index:
  ```go
  for i, opt := range opts {
      if err := opt(&cfg); err != nil {
          return nil, fmt.Errorf("fuzzymatch: option[%d]: %w", i, err)
      }
  }
  ```
  Verify the wrapped error still satisfies `errors.Is(err, ErrInvalidWeight)`.

---

### IMP-07: `nil` option in `NewScorer` panics with cryptic nil-function dereference

- **File:** `/Users/johnny/Development/fuzzymatch/scorer.go:199-203`
- **Phase introduced:** Phase 8
- **Issue:** A literal `nil` in the variadic `opts` slice (or an option
  set via `opts[i] = nil` after construction of the slice) panics with
  "runtime error: invalid memory address or nil pointer dereference"
  rather than returning a typed error. Consumer error rather than a real
  attack surface, but a defence-in-depth improvement.
- **Standard:** `.claude/skills/go-coding-standards/SKILL.md` — defensive
  coding.
- **Action:** Code fix.
- **Rationale:** Same as Phase 8 SEC-09 — typed error preferred to
  cryptic panic.
- **Suggested fix:**
  ```go
  for i, opt := range opts {
      if opt == nil {
          return nil, fmt.Errorf("fuzzymatch: nil option at index %d: %w",
              i, ErrInvalidConfiguration)
      }
      if err := opt(&cfg); err != nil { ... }
  }
  ```

(Re-asserts Phase 8 SEC-09 — gap remains.)

---

### IMP-08: Scorer's `applyNormalisation` interacts with token-based algorithms — documented but worth re-asserting

- **File:** `/Users/johnny/Development/fuzzymatch/scorer.go:312-321`
- **Phase introduced:** Phase 8
- **Issue:** When the Scorer is constructed with
  `WithNormalisation(opts)`, Score applies `Normalise(s, normaliseOpts)`
  to both inputs at the boundary. Token-based algorithms (Monge-Elkan,
  Token*, PartialRatio) then call `Tokenise` internally on the already-
  normalised string. This means the Scorer's normalisation is applied
  TWICE-ish to token-based algorithms (Normalise's collapse + Tokenise's
  re-split). Documented (lines 314-321), but consumers may be surprised
  that `Normalise(s)` + `Tokenise(s)` is not the same as
  `Tokenise(Normalise(s))`. This is a correctness-discipline question
  rather than a security one, but the security-reviewer angle is:
  consumers tuning a Scorer for adversarial input may believe they have
  control over the normalisation pipeline when they actually only
  control its first half.
- **Standard:** `.claude/skills/documentation-standards/SKILL.md`.
- **Action:** Docs.
- **Rationale:** Improve the docstring with a small Mermaid-style
  example showing the pipeline.

---

### IMP-09: Token-based algorithms internally call `Tokenise(s, DefaultTokeniseOptions())` — no consumer control

- **File:** `/Users/johnny/Development/fuzzymatch/monge_elkan.go:394-395`,
  `/Users/johnny/Development/fuzzymatch/token_jaccard.go`,
  `/Users/johnny/Development/fuzzymatch/token_sort_ratio.go`,
  `/Users/johnny/Development/fuzzymatch/token_set_ratio.go`
- **Phase introduced:** Phase 6
- **Issue:** Every token-based algorithm hard-codes
  `DefaultTokeniseOptions()`. Consumers wanting a non-default tokeniser
  (e.g. custom `SeparatorChars`) cannot pass one through. The Scorer
  options `WithMongeElkanAlgorithm` accept a `NormalisationOptions`
  parameter but ignore it (`monge_elkan.go:393`). This is a feature gap,
  not a security defect — but a consumer who Normalise's with one
  separator set and assumes Tokenise will use the same is surprised.
- **Standard:** `.claude/skills/fuzzymatch-review-protocol/SKILL.md` —
  documentation discipline.
- **Action:** Docs (+ future API expansion).
- **Rationale:** Documentation gap. Surface the hard-coded
  `DefaultTokeniseOptions()` choice in each algorithm's godoc.

---

### IMP-10: Empty `SeparatorChars` in `NormalisationOptions` silently degrades to "whitespace-only collapse"

- **File:** `/Users/johnny/Development/fuzzymatch/normalise.go:68`
- **Phase introduced:** Phase 1
- **Issue:** `SeparatorChars: ""` combined with `StripSeparators: true`
  is documented (`:68-69`) as "equivalent to whitespace-only collapsing".
  An adversarial config could exploit this — a consumer who programmatically
  builds `NormalisationOptions` from a JSON-decoded config where
  `SeparatorChars` is missing gets a different normalisation than
  intended. Not a security defect; defence-in-depth surfacing.
- **Standard:** `.claude/skills/documentation-standards/SKILL.md`.
- **Action:** No action (already documented).
- **Rationale:** Spot-check confirms the documentation; mention in any
  future "configuring from JSON" doc.

---

### IMP-11: Phonetic algorithms accept (and silently drop) non-ASCII runes — no signal

- **File:** `/Users/johnny/Development/fuzzymatch/soundex.go:56-61`,
  `/Users/johnny/Development/fuzzymatch/double_metaphone.go:69-74`,
  `/Users/johnny/Development/fuzzymatch/nysiis.go:75-80`,
  `/Users/johnny/Development/fuzzymatch/mra.go:53-58`
- **Phase introduced:** Phase 7
- **Issue:** All four phonetic algorithms document that non-ASCII runes
  are silently dropped before encoding. This is the correct behaviour
  for phonetic codes that are by definition ASCII-letter-keyed. From a
  security angle, an adversarial input with all-non-ASCII content would
  produce an empty phonetic code, which on the Scorer surface produces
  a score of 1.0 for any two such inputs (vacuous-match convention).
  This could be weaponised by an attacker who controls input strings —
  feed two unrelated multi-byte UTF-8 strings to a Scorer with phonetic
  weights and get a vacuous-match score that contributes to a false-
  positive composite. Documented but the security implication is not
  surfaced.
- **Standard:** Security-reviewer focus area; `.claude/skills/fuzzymatch-
  review-protocol/SKILL.md`.
- **Action:** Docs.
- **Rationale:** Defence-in-depth — surface the adversarial-input angle
  in each phonetic file's godoc and in `docs/scorer.md`.
- **Suggested fix:** Add a paragraph: "Phonetic algorithms over
  predominantly non-ASCII input produce empty codes and a vacuous-match
  1.0 score. Consumers feeding untrusted multi-byte input to a Scorer
  with phonetic weights should pre-validate that at least one
  ASCII-letter rune is present, or compose with `Normalise +
  StripDiacritics` first."

---

### IMP-12: Phonetic algorithms operate on ASCII letters by `>='A' && <='Z'` byte arithmetic — embedded NUL bytes treated as non-letters

- **File:** `/Users/johnny/Development/fuzzymatch/soundex.go:111-126`
  (soundexGroup); similar pattern in double_metaphone, nysiis, mra
- **Phase introduced:** Phase 7
- **Issue:** Embedded NUL bytes (0x00) in input strings are dropped
  before phonetic encoding (they are not ASCII letters). This is
  correct behaviour but the `_test.go` files do not include an embedded-
  NUL canonical reference vector. Defence-in-depth.
- **Standard:** Security-reviewer focus area — embedded NUL handling.
- **Action:** Tests.
- **Rationale:** Add an embedded-NUL test row to each phonetic algorithm
  asserting the documented "silently drop" behaviour.

---

### IMP-13: Phonetic similarity timing — score lookup performs at-most-4-character byte compare

- **File:** `/Users/johnny/Development/fuzzymatch/double_metaphone.go`,
  `/Users/johnny/Development/fuzzymatch/soundex.go`,
  `/Users/johnny/Development/fuzzymatch/nysiis.go`,
  `/Users/johnny/Development/fuzzymatch/mra.go`
- **Phase introduced:** Phase 7
- **Issue:** Per the prompt: "side-channel timing on phonetic
  comparisons". Phonetic algorithms produce a fixed-or-bounded-length
  code (≤ 4 chars for Soundex/DM, ≤ 6 for NYSIIS/MRA). The final
  score comparison is a byte-equality check. Timing is constant-time in
  the secret-handling sense by virtue of the bounded code length —
  there is no early-exit on byte mismatch in a comparison loop that
  would leak position information. Verified by reading
  `DoubleMetaphoneScore` (one byte equality) and `SoundexScore`.
  Spot-check passed.
- **Standard:** Security-reviewer focus area — timing leakage.
- **Action:** No action.
- **Rationale:** The library does not handle secrets; even so, the
  phonetic-score comparison surface is not a timing-leak hazard.

---

### IMP-14: No fuzz-corpus seeds with multi-MB inputs

- **File:** All `*_fuzz_test.go` files
- **Phase introduced:** Phases 2-8
- **Issue:** Inspection of `damerau_full_fuzz_test.go` (representative)
  shows seed entries with at most ~10-character inputs. Per the
  security-reviewer focus area: "Fuzz tests include very long inputs
  (multi-KB) without timeout". Native Go fuzz generates inputs of varied
  lengths automatically, so this is not a strict gap — but adding a
  hand-curated multi-MB seed would guarantee the worst-case allocation
  paths are exercised on every nightly fuzz run.
- **Standard:** Security-reviewer focus area;
  `.claude/skills/go-testing-standards/SKILL.md`.
- **Action:** Tests.
- **Rationale:** Defence-in-depth — guarantee the pathological-input
  path is fuzz-exercised.
- **Suggested fix:** Add a 64KB+ seed to each fuzz corpus
  (`testdata/fuzz/Fuzz*/`).

---

### IMP-15: Make-script generators are Python — supply-chain footprint for cross-validation corpora

- **File:** `/Users/johnny/Development/fuzzymatch/scripts/gen-swg-cross-validation.py`,
  `gen-ratcliff-obershelp-cross-validation.py`,
  `gen-token-ratio-cross-validation.py`,
  `gen-phonetic-cross-validation.py`
- **Phase introduced:** Phases 3-7
- **Issue:** The cross-validation corpora at
  `testdata/cross-validation/*/vectors.json` are committed to the repo
  and consumed by the Go test suite at run-time. They are generated by
  Python scripts that import third-party packages (e.g.
  `rapidfuzz==3.14.5`, `jellyfish==1.2.1`). These Python deps are NOT
  part of the runtime or test trust set — but a CI run that regenerated
  the corpora from a compromised PyPI mirror would inject adversarial
  reference vectors that would then pin "wrong" expected outputs in the
  test suite.
- **Standard:** Security-reviewer focus area — code provenance;
  `.claude/skills/fuzzymatch-review-protocol/SKILL.md`.
- **Action:** Docs.
- **Rationale:** The corpora are committed once and reviewed once; this
  is informational. Document that the corpus regeneration is a
  privileged operation requiring algorithm-correctness-reviewer
  sign-off.
- **Suggested fix:** Add a `CORPUS_REGENERATION.md` documenting the
  trust assumptions of the Python regeneration scripts.

---

### IMP-16: Scorer reduction does not guard against NaN scores from individual algorithms

- **File:** `/Users/johnny/Development/fuzzymatch/scorer.go:368-381`
- **Phase introduced:** Phase 8
- **Issue:** The reduction loop accumulates `acc = acc +
  (entry.weight * score)` without checking that `score` is finite.
  Every algorithm's property tests assert score ∈ [0, 1] — but a future
  algorithm regression that returned NaN under a corner case would
  silently poison the composite Score and every downstream Match /
  ScoreAll. Defence-in-depth.
- **Standard:** `.claude/skills/determinism-standards/SKILL.md` —
  NaN handling.
- **Action:** Code fix (optional).
- **Rationale:** Same as Phase 8 SEC-08.
- **Suggested fix:** Add a property test
  `TestProp_Scorer_NoNaN_NoInf_AllSingleAlgoCompositions` that iterates
  every AlgoID as a single-algorithm Scorer and asserts finite output
  on random pairs. This catches per-algorithm regressions at the Scorer
  surface.

---

### IMP-17: No regex usage — verified clean

- **File:** library-wide (verified by `grep regexp`)
- **Phase introduced:** N/A
- **Issue:** Per the security-reviewer focus area: "Where regular
  expressions are used (rare — possibly in tokenisation or
  normalisation), verify: ...". Verification: no `regexp` import anywhere
  in non-test code (`grep -n "regexp\|MustCompile" /Users/.../*.go |
  grep -v _test.go` returned empty). Tokenise and Normalise use
  hand-coded byte/rune loops with constant-time `[128]bool` tables.
- **Standard:** Security-reviewer focus area.
- **Action:** No action.
- **Rationale:** Verified clean.

---

### IMP-18: `DefaultScorer()` panic on internal inconsistency — verified bounded

- **File:** `/Users/johnny/Development/fuzzymatch/scorer.go:586-592`
- **Phase introduced:** Phase 8
- **Issue:** Documented panic on `NewScorer(DefaultScorerOptions()...)`
  failure. Verified unreachable under any consumer input pattern: the
  six AlgoIDs in `DefaultScorerOptions` are package-load-time bound,
  the threshold 0.85 is a literal in [0, 1], the weights are 1.0
  literals.
- **Standard:** Security-reviewer focus area — panic discipline.
- **Action:** No action.
- **Rationale:** Spot-check confirms boundedness (also confirmed in
  Phase 8 SEC-10).

---

## Verification of focus-area requirements (whole codebase)

| Focus area | Status | Notes |
|------------|--------|-------|
| Every algorithm documents complexity in godoc | YES | Verified across 23 files. Aggregate composite complexity in `docs/scorer.md` not documented — IM-06. |
| Super-linear complexity algorithms documented with warnings | PARTIAL | ME / PartialRatio have explicit "DoS notice" sections; others document complexity but not DoS — IM-06. |
| Fuzz tests include multi-KB inputs | PARTIAL | Per-algorithm fuzz exists for all 26 surfaces; multi-MB hand-curated seeds missing — IMP-14. |
| `docs/performance.md` discusses input-size bounding | NO | 60-line scaffold all TBD — IM-06. |
| No unbounded recursion | NO | Ratcliff-Obershelp uses Go-call-stack recursion bounded only by O(min(la, lb)) — IM-04. |
| Public functions never panic on arbitrary input | NO | CR-01 (Tversky α=β=0); CR-02 (ME non-allowlisted inner); IM-07 (direct-call algorithm panics on programmer error). |
| `PropAlgorithm_NeverPanics` property test exists | PARTIAL | `TestProp_Normalise_NeverPanics` + `TestProp_Tokenise_NeverPanics` exist; no Scorer-level harness — IM-03. |
| Error messages do not embed user input | YES | Verified across `errors.go` and Scorer surfaces. |
| No timing-based information leakage | YES | No secrets handled; phonetic surface has bounded comparison — IMP-13. |
| Zero runtime dependencies | YES (curated) | Single curated `x/text` dep; `verify-no-runtime-deps.sh` clean. `go.sum` still records `x/mod`/`x/sync`/`x/tools` `/go.mod` hashes — IM-09. |
| Test deps isolated in `tests/bdd/go.mod` | YES | Verified. |
| `govulncheck ./...` clean | UNKNOWN | Runs in CI on every PR + weekly; status not verified in this offline review. |
| No GPL/LGPL-derived code | YES | algorithm-licensing-reviewer enforced; provenance statements in every algorithm file. |
| Regex safety | N/A | No regex usage — IMP-17. |
| Invalid UTF-8 graceful | YES | Normalise replaces with U+FFFD; algorithm rune paths use Go's `[]rune` conversion which replaces invalid sequences. |
| Embedded NULs | PARTIAL | Phonetic algorithms drop them silently; no explicit test row — IMP-12. |

---

## Summary

**Must fix before v1.0 (CRITICAL):**
1. CR-01 — `WithTverskyAlgorithm` α=β=0 panic
2. CR-02 — `WithMongeElkanAlgorithm` non-allowlisted inner panic

**Should fix before v1.0 (IMPORTANT — DoS / silent-wrong-answer / supply-chain):**
3. IM-01 — `WithThreshold` NaN handling
4. IM-02 — `WithAlgorithm` NaN/Inf weight handling
5. IM-03 — Scorer-level fuzz harness
6. IM-04 — Ratcliff-Obershelp iterative or bounded-recursion
7. IM-05 — Damerau-Levenshtein Full O(m·n) DP table guard
8. IM-06 — Populate `docs/performance.md`; document Scorer-aggregate complexity
9. IM-07 — Discuss-phase: convert direct-call panics OR exhaustively document
10. IM-08 — Pin floating action / tool tags
11. IM-09 — Document `go.sum` trust set beyond `x/text`
12. IM-10 — `SECURITY.md` cosign trust-chain section
13. IM-11 — Discuss-phase: Scorer-level allocation-bound option

**Defence-in-depth (IMPROVEMENT):**
14. IMP-03 — Damerau-Full rune-path da-map capacity hint
15. IMP-05 — SWG params NaN/Inf validation at Scorer surface
16. IMP-06 — Wrap option errors with index
17. IMP-07 — `nil` ScorerOption guard
18. IMP-08 — Document Normalise + Tokenise pipeline interaction
19. IMP-09 — Surface hard-coded `DefaultTokeniseOptions` in algorithm docs
20. IMP-11 — Phonetic algorithm vacuous-match adversarial-input warning
21. IMP-12 — Embedded NUL test rows in phonetic suites
22. IMP-14 — Multi-MB fuzz-corpus seeds
23. IMP-15 — Document Python regeneration script trust assumptions
24. IMP-16 — Per-AlgoID single-Scorer NaN/Inf property test

**Verified clean (no action):**
25. IMP-01 — q-gram map capacity hint
26. IMP-02 — Partial Ratio rune-path charSet bound
27. IMP-04 — Normalise per-call transformer construction
28. IMP-10 — Normalise empty `SeparatorChars` documented
29. IMP-13 — Phonetic timing surface
30. IMP-17 — No regex usage
31. IMP-18 — `DefaultScorer()` panic bound

---

_Reviewed: 2026-05-17_
_Reviewer: security-reviewer_
_Scope: entire codebase (Phases 1-8)_

</details>

---

## 14. test-analyst

_Source: `.planning/reviews/test-analyst-FINDINGS.md`_

<details>
<summary>Click to expand full report</summary>

---
status: issues_found
agent: test-analyst
scope: entire test suite (phases 1-8)
reviewed: 2026-05-17T00:00:00Z
finding_counts:
  critical: 11
  important: 18
  improvement: 13
  total: 42
---

# fuzzymatch — Whole-Codebase Test-Suite Health Analysis

Holistic analysis covering every `_test.go` file in the root package
plus the BDD sub-module (`tests/bdd/`). Expands the previous
phase-scoped review in `.planning/phases/08-composite-scorer/08-TEST-ANALYSIS.md`
to all phases 1-8.

## Headline numbers

- **Package coverage:** 90.4% (target 95% overall — **fails**).
- **Per-file coverage:** `double_metaphone.go` 83.8% (target ≥ 90% — **fails**); all other files ≥ 90%.
- **Public API coverage:** ≥ 99% on every exported Score/Code/Compare function (every exported Score is exercised in `props_test.go`, `*_test.go`, examples, and at least one fuzz harness — though several `*Runes` variants and the `Scorer` lack fuzz harnesses, see below).
- **Total tests:** ~480 top-level `Test*` functions plus 39 `Example*` plus 31 `Fuzz*` plus 127 `Benchmark*`.
- **BDD scenarios:** 25 feature files (one per algorithm + monge_elkan_phonetic_inner + scorer); 169 scenarios total. **No `scan.feature`, no `suppression.feature`, no `normalisation.feature`, no `determinism.feature`** — listed in `docs/requirements.md` §15.6 and the prompt scan/normalisation checklists. (Scan sub-package is unimplemented Phase 6 work — out-of-scope this milestone; normalisation and determinism scenarios remain a documented gap.)
- **Fuzz corpus directories on disk:** 20 (`testdata/fuzz/`). Fuzz harness functions in source: 31. Discrepancy: 11 fuzzers have no committed seed corpus.
- **BDD pass:** `cd tests/bdd && go test ./...` → ok.
- **Race tests:** `go test -race ./...` → ok (~352s, passes).

## Per-Algorithm Coverage Summary

| Algorithm | Unit | Property | Fuzz (byte/rune) | Bench | BDD | Golden | Cross-validation |
|-----------|------|----------|------------------|-------|-----|--------|------------------|
| Levenshtein | OK | OK + Triangle | byte only | OK | OK (5) | OK | impl-internal |
| DamerauLevenshteinOSA | OK | OK + Triangle | byte only | OK | OK (4) | OK | impl-internal |
| DamerauLevenshteinFull | OK | OK + Triangle | byte only | OK | OK (5) | OK | impl-internal |
| Hamming | OK | OK + Triangle (eq-len) | byte only | OK | OK (6) | OK | impl-internal |
| Jaro | OK | OK (no triangle — doc'd) | byte only | OK | OK (4) | OK | impl-internal |
| JaroWinkler | OK | OK + AtLeastJaro | byte only | OK | OK (6) | OK | impl-internal |
| Strcmp95 | OK | OK + AtLeastJW | byte only | OK | OK (6) | OK | impl-internal |
| SmithWatermanGotoh | OK | OK | byte only (Score) | OK | OK (6) | OK | OK (vectors.json) |
| LCSStr | OK | OK | byte only | OK | OK (8) | OK | impl-internal |
| QGramJaccard | OK | OK | byte + Runes | OK | OK (6) | OK | impl-internal |
| SorensenDice | OK | OK | byte + Runes | OK | OK (6) | OK | impl-internal |
| Cosine | OK | OK | byte + Runes | OK | OK (7) | OK | impl-internal |
| Tversky | OK | OK + Asymmetry | byte + Runes | OK | OK (7) | OK | OK (vs Jaccard/Dice) |
| MongeElkan | OK | OK + Asymmetry (flake risk) | byte (Score + Symmetric) | OK | OK (9) + phonetic (6) | OK | impl-internal |
| TokenSortRatio | OK | OK | byte | OK | OK (6) | OK | OK (vectors.json) |
| TokenSetRatio | OK | OK | byte | OK | OK (7) | OK | OK (vectors.json) |
| PartialRatio | OK | OK | byte + Runes | OK | OK (14) | OK | OK (vectors.json) |
| TokenJaccard | OK | OK | byte | OK | OK (6) | OK | impl-internal |
| Soundex | OK | OK (no NoNegativeZero) | byte | OK | OK (7) | OK | OK (vectors.json) |
| DoubleMetaphone | OK | OK (no NoNegativeZero) | byte | OK | OK (9) | OK | OK (vectors.json) |
| NYSIIS | OK | OK (no NoNegativeZero) | byte | OK | OK (10) | OK | OK (vectors.json) |
| MRA | OK | OK (no NoNegativeZero) | byte | OK | OK (10) | OK | OK (vectors.json) |
| RatcliffObershelp | OK | OK | byte | OK | OK (5) | OK | OK (vectors.json) |

All 23 catalogue algorithms have unit + property + at least one fuzz +
benchmark + BDD + cross-platform golden coverage.

---

## Findings

### [Critical] No Scorer fuzz harness — `FuzzDefaultScorerScore`, `FuzzDefaultScorerScoreAll`, `FuzzDefaultScorerMatch` all missing
- **File:** `scorer_fuzz_test.go` (would-be path; file does not exist)
- **Phase introduced:** Phase 8
- **Issue:** `.claude/skills/go-testing-standards/SKILL.md` §Fuzz Tests: "Every public function has a fuzz harness in `fuzz_test.go`." 31 fuzz functions exist, none target the `Scorer.Score`, `Scorer.ScoreAll`, `Scorer.Match`, or `NewScorer` surfaces. `docs/requirements.md` §15.4 explicitly lists `FuzzScorer` as required.
- **Standard:** `.claude/skills/go-testing-standards/SKILL.md` §Fuzz Tests; `docs/requirements.md` §15.4
- **Action:** Code fix
- **Rationale:** The Scorer is the most-used public surface but the only public surface without a fuzz gate. Bug-rich path: pre-normalisation, AlgoID dispatch, reduction loop, threshold compare.
- **Suggested fix:** Three small harnesses seeding with `user_id / userId`, both-empty, `\xff\xfe`, etc. — same pattern as `levenshtein_fuzz_test.go`. (Identical to 08-TEST-ANALYSIS TEST-15.)

### [Critical] `WithThreshold(NaN)` not rejected — silent malfunction not gated by test
- **File:** `scorer_options.go:259`; absent test in `scorer_options_test.go`
- **Phase introduced:** Phase 8
- **Issue:** `t < 0.0 || t > 1.0` is both `false` for `math.NaN()`; the option silently accepts NaN and the resulting Scorer's `Match` returns `false` for every input (NaN comparisons always false). No unit test exists for NaN input.
- **Standard:** `.claude/skills/algorithm-correctness-standards/SKILL.md` §Edge cases (NaN/Inf handling); `docs/requirements.md` §13 (determinism — NaN must be rejected at boundaries)
- **Action:** Code fix (production code + test)
- **Rationale:** Same finding as 08-TEST-ANALYSIS TEST-01; production bug + missing test combined.
- **Suggested fix:** Use `math.IsNaN(t) || t < 0.0 || t > 1.0`; add `TestWithThreshold_RejectsNaN`.

### [Critical] `WithTverskyAlgorithm(α=0, β=0)` not rejected — direct dispatch will panic at runtime
- **File:** `scorer_options.go` (WithTverskyAlgorithm); absent test in `scorer_options_test.go:433-444`
- **Phase introduced:** Phase 8
- **Issue:** Option layer accepts `α==0, β==0` even though direct call `TverskyScore(_, _, _, 0, 0)` panics (see `tversky_test.go:568-582`). The Scorer's "fail at construction time, never at Score time" contract is violated; no test gates this gap.
- **Standard:** `.claude/skills/algorithm-correctness-standards/SKILL.md` §Edge cases; Tversky's `ErrInvalidTverskyParam` documented in `errors.go`
- **Action:** Code fix (production code + test)
- **Rationale:** Same as 08-TEST-ANALYSIS TEST-02; production-side bug compounds with the test-coverage gap.
- **Suggested fix:** Gate `α == 0 && β == 0` in `WithTverskyAlgorithm`; add `TestWithTverskyAlgorithm_RejectsBothZero`.

### [Critical] `WithAlgorithm(_, NaN)` and `WithAlgorithm(_, +Inf)` not rejected
- **File:** `scorer_options.go:152`; absent test in `scorer_options_test.go:56-73`
- **Phase introduced:** Phase 8
- **Issue:** `weight <= 0` is `false` for NaN (NaN comparisons always false) and for +Inf. The option layer accepts both; resulting Scorer produces NaN/Inf composites. The existing `TestWithAlgorithm_InvalidWeight` covers `-1.0`, `-0.5`, `0`, `-100` but not NaN/Inf.
- **Standard:** `.claude/skills/algorithm-correctness-standards/SKILL.md` §Edge cases (NaN/Inf), `.claude/skills/determinism-standards/SKILL.md` (NaN propagation)
- **Action:** Code fix
- **Rationale:** Same class of defect as the WithThreshold gap. Silent malfunction; existing PropScorer_NoNaN_NoInf would catch it on `Score` output but not at option time.
- **Suggested fix:** Add `if math.IsNaN(weight) || math.IsInf(weight, 0) { return ErrInvalidWeight }`; add `TestWithAlgorithm_RejectsNaNInfWeight`.

### [Critical] Three required meta-tests missing: `internal_coverage_test.go`, `documentation_test.go`, `readme_shop_front_test.go`
- **File:** (would-be paths; files do not exist)
- **Phase introduced:** Phase 1 (bootstrap)
- **Issue:** `.claude/skills/go-testing-standards/SKILL.md` §Meta-tests requires all five meta-tests; only `ai_friendly_test.go` and `makefile_targets_test.go` exist. `internal_coverage_test.go` is the project's automated 95%/90%/100% coverage-floor gate — without it the standards' coverage target is unenforced and the 90.4% overall coverage shortfall slips by silently.
- **Standard:** `.claude/skills/go-testing-standards/SKILL.md` §Meta-tests; `docs/requirements.md` §15.7
- **Action:** Code fix
- **Rationale:** The coverage-floor gate is the project's load-bearing automatic enforcement of the 95% overall / 90% per-file / 100% public-API targets. Without it, drift is invisible.
- **Suggested fix:** Mirror `axonops/mask`'s pattern — read `go.test` coverage output via `go tool cover -func` and assert numeric thresholds. Documentation_test parses code blocks in `docs/*.md` and verifies they compile.

### [Critical] `double_metaphone.go` per-file coverage 83.8% — below the 90% per-file floor
- **File:** `double_metaphone.go`
- **Phase introduced:** Phase 4
- **Issue:** `dmPrep` 58.8%, `DoubleMetaphoneKeys` 62.0%, `dmAdd` 84.6%, `dmSlgCheck` 88.9%, `DoubleMetaphoneScore` 92.3%. Average 83.8%; below the 90% per-file standard floor.
- **Standard:** `.claude/skills/go-testing-standards/SKILL.md` line 16 ("≥ 90% per file")
- **Action:** Code fix (more unit tests for Slavic/Germanic/Spanish/Italian/French branches in `dmPrep` and uncovered table cells in `DoubleMetaphoneKeys`)
- **Rationale:** Phonetic-rule branches in DM are notoriously hard to exhaust; many branches are language-specific and only fired by surnames the unit tests don't include. RV-DM and language-branch unit tests cover 38 names but the per-file budget needs ~60+ to bring `dmPrep` to ≥ 90%.
- **Suggested fix:** Add reference vectors from `phonetic` cross-validation corpus that fire the uncovered branches (e.g. "Brzęczyszczykiewicz" for `dmPrep` Slavic path; Italian "GLI" prefix; French "ILLE" suffix). The cross-validation vectors.json already lists candidate names.

### [Critical] Overall coverage 90.4% — below the 95% overall floor
- **File:** Root package
- **Phase introduced:** Inherited from Phase 4 (phonetic) + Phase 6 (tokenisation)
- **Issue:** `go test -coverprofile=coverage.out ./... && go tool cover -func=coverage.out` reports 90.4% (after race test 90.5%). Standards require ≥ 95% overall.
- **Standard:** `.claude/skills/go-testing-standards/SKILL.md` line 16 ("≥ 95% overall"); `docs/requirements.md` §15.8
- **Action:** Code fix
- **Rationale:** Headline gap. Primarily traces to double_metaphone (83.8%), then tokenise (93.1%), nysiis (90.9%), token_set_ratio (92.0%), soundex (92.3%). No single file is the cause — the standard requires improvement across phonetic + tokenisation surfaces.
- **Suggested fix:** Add language-branch reference vectors and exercise the residual error-path branches (e.g. soundex non-ASCII edge cases, tokenise containsNonASCII early-return).

### [Critical] `TestProp_Scorer_WeightSumOne` `uint16` overflow flake — `u+1` overflows to 0 when `u==65535`
- **File:** `scorer_test.go:827`
- **Phase introduced:** Phase 8
- **Issue:** `return float64(u+1) / float64(uint32(1)<<16) * 100.0` — `u` is `uint16` so `u+1` overflows to `0`, returning weight `0`, which the option layer rejects via `ErrInvalidWeight`. Property test then returns `false` on that draw. Per-run flake probability is ~0.46% (3 vars × 100 quick.Check iters × 1/65536 each), enough to surface intermittently in CI.
- **Standard:** `.claude/skills/go-testing-standards/SKILL.md` (Property tests must be deterministic-passing); 08-REVIEW WR-01
- **Action:** Code fix
- **Rationale:** Same as 08-TEST-ANALYSIS TEST-07; ONE-LINE fix still unaddressed.
- **Suggested fix:** `return (float64(u) + 1.0) / 65536.0 * 100.0` — float arithmetic, no integer wrap.

### [Critical] Scorer benchmarks NOT in `bench.txt` — regression gate is inactive
- **File:** `bench.txt`
- **Phase introduced:** Phase 8
- **Issue:** `bench.txt` is the committed benchstat baseline; `grep -i scorer bench.txt` returns zero matches. The 6 `BenchmarkDefaultScorer_*` benchmarks in `scorer_bench_test.go` run during `make bench` but their numbers are not in the baseline, so a 10x regression in `DefaultScorer.Score` would not fail CI.
- **Standard:** `.claude/skills/go-testing-standards/SKILL.md` line 90; `docs/requirements.md` §14.4
- **Action:** Code fix (regenerate bench.txt)
- **Rationale:** Identical to 08-TEST-ANALYSIS TEST-10. The benchmark gate that the spec demands is wired in the Makefile but the baseline doesn't include the latest surface.
- **Suggested fix:** Re-run `go test -bench=. -benchmem -count=10 ./... > bench.txt`; commit.

### [Critical] No DefaultScorer-level property tests: `Identity`, `Symmetric`, `Score = Σ Weight·ScoreAll`
- **File:** `scorer_test.go` (would-be tests)
- **Phase introduced:** Phase 8
- **Issue:** The Scorer-level property tests check `DeterministicAcrossRuns`, `WeightSumOne`, `ScoreInRange`, `NoNaN_NoInf` — but NOT the three load-bearing mathematical invariants:
  - `DefaultScorer().Score(x, x) ≈ 1.0` for non-empty x
  - `DefaultScorer().Score(a, b) == DefaultScorer().Score(b, a)` (Scorer composed of symmetric algorithms)
  - `Score(a, b) == Σ algorithm.Weight * ScoreAll(a, b)[algorithm.ID]` (the algebraic identity linking the two public surfaces)
- **Standard:** `.claude/skills/go-testing-standards/SKILL.md` §Property Tests — Scorer-level properties (identity, range, composite-bounded)
- **Action:** Code fix
- **Rationale:** Same as 08-TEST-ANALYSIS TEST-04 / TEST-05 / TEST-06. Without these, regressions in the pre-normalisation gate or the reduction loop would not be detected by property test.
- **Suggested fix:** Three new `TestProp_Scorer_*` functions per the suggested code in 08-TEST-ANALYSIS.

### [Critical] Stale property-test failure log committed at `bench.txt.new`
- **File:** `bench.txt.new`
- **Phase introduced:** Phase 6 (MongeElkan)
- **Issue:** `bench.txt.new` contains the failure trace of `TestProp_MongeElkanScore_AsymmetricWhenTokenCountAsymmetric` from an earlier run. The file appears committed (not in `.gitignore`). The test now PASSES on 20 consecutive runs, but the file's existence suggests the asymmetric-property generator can produce inputs where Tokenise yields unequal token counts but MongeElkan happens to be symmetric (e.g. exotic-Unicode-only inputs where Tokenise's count of "fields" via strings.Fields differs from Tokenise's actual output). The test uses `strings.Fields(a)` as an under-estimate, which means the premise can be wrongly thought to hold.
- **Standard:** `.claude/skills/go-testing-standards/SKILL.md` (property tests must not flake); `.claude/skills/algorithm-correctness-standards/SKILL.md` (mathematical invariants must hold deterministically)
- **Action:** Discuss-phase needed (decide whether to tighten generator or accept the asymmetric-direction property is too narrow)
- **Rationale:** This is a *real* generator pitfall — the conditional invariant `aTokens != bTokens && 0 < fwd < 1 → fwd != rev` is true mathematically when the premise's token count is computed the same way MongeElkan does. Using `strings.Fields` for the premise but project Tokenise inside the algorithm creates an inconsistency.
- **Suggested fix:** Either (a) delete `bench.txt.new` from version control, gate the property to ASCII inputs only via custom generator; or (b) compute token counts via `fuzzymatch.Tokenise(a, fuzzymatch.DefaultTokeniseOptions())` to match the algorithm's view.

---

### [Important] `*Runes` fuzz harnesses missing for 9 of 23 algorithms
- **File:** `levenshtein_fuzz_test.go`, `damerau_full_fuzz_test.go`, `damerau_osa_fuzz_test.go`, `hamming_fuzz_test.go`, `jaro_fuzz_test.go`, `jarowinkler_fuzz_test.go`, `lcsstr_fuzz_test.go`, `swg_fuzz_test.go`, `ratcliff_obershelp_fuzz_test.go`
- **Phase introduced:** Phases 2-6
- **Issue:** Cosine, SorensenDice, QGramJaccard, Tversky each have `Fuzz*Score` AND `Fuzz*ScoreRunes`. Levenshtein, DamerauLevenshteinOSA/Full, Hamming, Jaro, JaroWinkler, LCSStr, SmithWatermanGotoh, RatcliffObershelp have only the byte-path fuzz. Strcmp95 is byte-only by design (ASCII letters), so that one is OK.
- **Standard:** `.claude/skills/go-testing-standards/SKILL.md` §Fuzz Tests ("one Fuzz* per public function"); `.claude/skills/algorithm-correctness-standards/SKILL.md` (rune variants are public API)
- **Action:** Code fix
- **Rationale:** Rune variants traverse different code paths (extra `[]rune` allocation, rune-by-rune scanning, different invalid-UTF-8 fallback). Bugs in the rune path won't be caught by the byte-path fuzz.
- **Suggested fix:** Add `Fuzz<Algorithm>ScoreRunes` per the existing per-Score pattern.

### [Important] Distance-function fuzz harnesses absent
- **File:** `levenshtein_fuzz_test.go`, `damerau_full_fuzz_test.go`, `damerau_osa_fuzz_test.go`, `hamming_fuzz_test.go`, `lcsstr_fuzz_test.go`
- **Phase introduced:** Phase 2
- **Issue:** Public API: `LevenshteinDistance`, `LevenshteinDistanceRunes`, `DamerauLevenshteinFullDistance`, `DamerauLevenshteinFullDistanceRunes`, `DamerauLevenshteinOSADistance`, `DamerauLevenshteinOSADistanceRunes`, `HammingDistance`, `HammingDistanceRunes`, `LongestCommonSubstring`, `LongestCommonSubstringRunes`. None have fuzz harnesses. Score-fuzz exercises distance internally for most, but `HammingDistance`'s unequal-length policy and `LongestCommonSubstring`'s return type (string) are not exercised by score fuzz.
- **Standard:** `.claude/skills/go-testing-standards/SKILL.md` ("Every public function has a fuzz harness")
- **Action:** Code fix
- **Rationale:** Public API rule is non-functional-form-aware.
- **Suggested fix:** Add `FuzzLevenshteinDistance`, `FuzzHammingDistance`, `FuzzLongestCommonSubstring` etc. asserting panic-free and non-negative integer / valid UTF-8 substring.

### [Important] Phonetic-Code fuzz harnesses only test `Score`, not `Code/Keys`
- **File:** `soundex_fuzz_test.go`, `double_metaphone_fuzz_test.go`, `nysiis_fuzz_test.go`, `mra_fuzz_test.go`
- **Phase introduced:** Phase 4
- **Issue:** `FuzzSoundex`, `FuzzNYSIIS`, `FuzzMRA`, `FuzzDoubleMetaphone` exist but the on-disk corpus directory is named `FuzzSoundex` (etc.), and the test exercises `*Score`, not the underlying `SoundexCode`, `NYSIISCode`, `MRACode`, `MRACompare`, `DoubleMetaphoneKeys`. The Code functions have their own unit-test coverage but no fuzz harness asserts the output character-set or length invariants under random input.
- **Standard:** `.claude/skills/go-testing-standards/SKILL.md`; `docs/requirements.md` §15.4
- **Action:** Code fix
- **Rationale:** Property tests `TestProp_SoundexCode_Charset`, `TestProp_NYSIISCode_Charset`, `TestProp_MRACode_Charset`, `TestProp_DoubleMetaphone_KeyCharset` cover the charset invariant under random input but not under adversarial input (invalid UTF-8, embedded NUL, very long input). A fuzz harness would close this.
- **Suggested fix:** Either expand `FuzzSoundex` etc. to also call the Code variant or add separate `FuzzSoundexCode` harnesses.

### [Important] Parameterised `Scorer` options not separately fuzz-tested
- **File:** scorer_options_*test.go and fuzz files
- **Phase introduced:** Phase 8
- **Issue:** `WithQGramJaccardAlgorithm`, `WithSorensenDiceAlgorithm`, `WithCosineAlgorithm`, `WithTverskyAlgorithm`, `WithMongeElkanAlgorithm`, `WithSmithWatermanGotohAlgorithm` each have happy-path + invalid-weight + invalid-N + (sometimes) invalid-alpha/beta unit tests. None have a fuzz harness over (n int, alpha, beta float64, inner AlgoID).
- **Standard:** `.claude/skills/go-testing-standards/SKILL.md` (one Fuzz* per public function)
- **Action:** Code fix
- **Rationale:** Option-layer regressions on float-edge values (denormals, NaN, infinities, very-large alpha/beta) would not be caught.

### [Important] No BDD feature files for `normalisation`, `determinism`
- **File:** `tests/bdd/features/` (would-be `normalisation.feature`, `determinism.feature`)
- **Phase introduced:** Phase 1 (bootstrap) — Phase 5 (normalisation/tokenise)
- **Issue:** `docs/requirements.md` §15.6 enumerates the six required feature files: `algorithms.feature`, `scorer.feature`, `normalisation.feature`, `determinism.feature`, `scan.feature`, `suppression.feature`. Only the per-algorithm feature files and `scorer.feature` exist. `normalisation.feature` and `determinism.feature` are not present even though the underlying surfaces are implemented (Phase 5 Normalise / Tokenise; cross-platform golden in `testdata/golden/`).
- **Standard:** `.claude/skills/go-testing-standards/SKILL.md` §BDD Tests; `docs/requirements.md` §15.6
- **Action:** Code fix
- **Rationale:** BDD is the contract layer. The normalisation pipeline (lowercase × strip-separators × camelCase-split × NFC × strip-diacritics) is heavily branched and well-tested at unit level but the BDD coverage is missing.
- **Suggested fix:** Create `normalisation.feature` with one Scenario Outline per option combination per `docs/requirements.md` §9; create `determinism.feature` cross-referencing the golden-file behaviour. Tag with `@normalisation` / `@determinism`.

### [Important] Required BDD tags missing: `@character`, `@qgram`, `@token`, `@phonetic`, `@gestalt`, `@scan`, `@suppression`, `@normalisation`, `@determinism`
- **File:** `tests/bdd/features/*.feature`
- **Phase introduced:** Phases 2-8
- **Issue:** `.claude/skills/go-testing-standards/SKILL.md` §BDD Tests requires category tags. Actual tags present: `@scorer`, `@partial`, `@phonetic`, `@token`, `@monge`, `@soundex`, `@mra`, `@nysiis`, `@double`, `@custom`, `@default`, `@scoreall`, `@concurrency`, `@errors`, `@byte`, `@rune`, `@pitfall`. Missing the category roll-up tags (`@character` for Levenshtein/Hamming/etc., `@qgram` for Cosine/Jaccard, `@gestalt` for RatcliffObershelp, `@determinism`, `@normalisation`).
- **Standard:** `.claude/skills/go-testing-standards/SKILL.md` §BDD Tests — Tags for filtering
- **Action:** Code fix
- **Rationale:** Tag-based filtering (`go test -godog.tags=@phonetic`) is the standard way to slice the suite. Several categories cannot be selected today.
- **Suggested fix:** Apply category tags to existing feature files; one line per scenario.

### [Important] No tests for `ErrInvalidInput`, `ErrInvalidConfiguration`, `ErrEmptyInput`
- **File:** `errors_test.go`
- **Phase introduced:** Phase 1
- **Issue:** `errors_test.go` exercises 10 sentinels with `errors.Is`/`Error()`/prefix checks, but these three sentinels are NEVER returned by any public function currently — they're declared but unused (grep confirms). The test asserts the sentinels exist; it does not assert any code path returns them.
- **Standard:** `.claude/skills/algorithm-correctness-standards/SKILL.md` (errors-as-API must be testable from public surface)
- **Action:** Discuss-phase needed
- **Rationale:** Either remove the unused sentinels or add the public API surfaces that return them (e.g. an `Extract` API gated by `ErrInvalidInput`). Currently they're contract-declared but contract-not-enforced.

### [Important] `ErrHammingLengthMismatch` documented in `errors.go` comment but not declared
- **File:** `errors.go:31`
- **Phase introduced:** Phase 2
- **Issue:** Comment references "ErrHammingLengthMismatch for the Hamming algorithm" but no such sentinel exists. The implementation uses the silent-zero policy (`hamming.go:74-80`). The algorithm-correctness-standards skill (line 104) says "Length mismatch for Hamming: return 0.0 from HammingScore (return ErrHammingLengthMismatch from HammingDistance). Documented behaviour." — but the implementation diverges from the skill.
- **Standard:** `.claude/skills/algorithm-correctness-standards/SKILL.md` line 104
- **Action:** Skill clarification or code fix
- **Rationale:** Either update the skill to match the silent-zero LOCKED policy (which is what the BDD feature file and unit tests pin) or implement the documented error variant. Currently the documentation in `.claude/skills/` and the implementation contradict.
- **Suggested fix:** Update the skill to reflect the silent-zero policy and document it as the LOCKED choice; remove the stale comment in `errors.go:31`.

### [Important] `ScoreAll` map key type — skill documents `string` keys (AlgoID.String()), code returns `map[AlgoID]float64`
- **File:** `scorer.go:497`; `.claude/skills/go-testing-standards/SKILL.md` line 34
- **Phase introduced:** Phase 8
- **Issue:** Skill: "`ScoreAll` returns per-algorithm scores keyed by `AlgoID.String()`". Code: returns `map[AlgoID]float64` (typed enum keys). The deviation is documented in `scorer.go:470` as a spec override.
- **Standard:** `.claude/skills/go-testing-standards/SKILL.md` §Scorer-level properties
- **Action:** Skill clarification
- **Rationale:** Implementation made a deliberate API choice (compile-time key safety) that the skill text doesn't reflect. Don't change the code; update the skill to match the override.

### [Important] No `ExampleNewScorer`, `ExampleDefaultScorer`, `ExampleScorer_Score`, `ExampleScorer_Match` runnable examples
- **File:** `example_test.go`
- **Phase introduced:** Phase 8
- **Issue:** 39 `Example*` functions exist — one per algorithm — but none for the Scorer surface. `docs/requirements.md` §16.5 specifies one Example per public Score-producing function; the Scorer's three public methods get no godoc-runnable demonstration.
- **Standard:** `.claude/skills/documentation-standards/SKILL.md`; `docs/requirements.md` §16.5
- **Action:** Code fix
- **Rationale:** pkg.go.dev consumers reading the `Scorer` doc get no executable example.
- **Suggested fix:** Add `ExampleDefaultScorer`, `ExampleNewScorer`, `ExampleScorer_Score`, `ExampleScorer_Match`, `ExampleScorer_ScoreAll` mirroring the per-algorithm pattern.

### [Important] Hardcoded `len(want) != 6` in `TestScorer_ConcurrentSafety`
- **File:** `scorer_test.go:941`
- **Phase introduced:** Phase 8
- **Issue:** Assertion `if len(want) != 6` couples the test to today's DefaultScorer composition. Any future change to DefaultScorer composition (e.g. add a 7th algorithm, drop one) silently breaks this assertion without explaining itself.
- **Standard:** general test-hygiene
- **Action:** Code fix
- **Rationale:** A spec change in `docs/requirements.md` §8.5 must be reflected here.
- **Suggested fix:** `want := results[0]; if len(want) != len(s.Algorithms()) { ... }` — derive from the Scorer itself.

### [Important] No "Very long input (1000+ chars)" tests for character-based algorithms (except Jaro)
- **File:** `levenshtein_test.go`, `damerau_*_test.go`, `hamming_test.go`, `jarowinkler_test.go`, `lcsstr_test.go`, `swg_test.go`, `strcmp95_test.go`, `ratcliff_obershelp_test.go`
- **Phase introduced:** Phases 2-6
- **Issue:** Only `jaro_test.go:198-206` exercises 300+ char input in a unit test. Other DP algorithms have bench tests at 500 chars but no functional test at 1000+ chars to assert the heap-path (large input → heap-allocated DP table) produces the right answer.
- **Standard:** prompt's per-algorithm checklist "Very long input (1000+ chars)"
- **Action:** Code fix
- **Rationale:** Stack-vs-heap path divergence is a documented performance optimisation; testing only short inputs leaves the heap path untested for correctness.

### [Important] No `quick.Check` `MaxCount` raised above default 100 for cheap properties
- **File:** `props_test.go` (entire file uses `nil` config)
- **Phase introduced:** Phases 2-8
- **Issue:** Almost every `quick.Check(f, nil)` call uses the default 100 iterations. For algorithms with ~10-50 µs per iter, 1000 or 10000 iterations would tighten confidence at trivial cost. Only `tokenise_test.go:506` raises to 200 and `props_test.go:491-497` uses a custom generator (Damerau-OSA constrained).
- **Standard:** `.claude/skills/go-testing-standards/SKILL.md` Property Tests
- **Action:** Code fix
- **Rationale:** Same as 08-TEST-ANALYSIS TEST-08 (scoped) — extends to whole codebase. The flake-once-per-100-runs surface area is unnecessarily wide.
- **Suggested fix:** Add a package-level `var quickCfg = &quick.Config{MaxCount: 1000}` and pass it consistently.

### [Important] Property-test generators rely on `quick.Check`'s default random `string` generator — produces predominantly long non-ASCII strings
- **File:** `props_test.go`
- **Phase introduced:** Phase 2
- **Issue:** Go's `testing/quick` default `string` generator produces random runes drawn from the full Unicode space. This is generally good (exercises non-ASCII paths) but means:
  - Property tests rarely exercise short-ASCII edge cases.
  - Inputs are predominantly multi-byte UTF-8 with surrogate-pair-style runes.
  - Some properties (e.g. SWG monotonicity, MRA threshold-monotonicity) may not generate enough realistic-shaped inputs.
- **Standard:** `.claude/skills/go-testing-standards/SKILL.md` Property Tests
- **Action:** Code fix or Discuss-phase needed
- **Rationale:** A targeted generator (mixed-ASCII / mixed-Unicode / mixed-length) would surface more bugs per iteration. The custom `randShortASCII` and `asciiAlpha` patterns already exist; expanding their use across `props_test.go` would help.

### [Important] No property test verifying `Algorithms()` ordering is by AlgoID ascending across permutations
- **File:** `scorer_test.go:353-397`
- **Phase introduced:** Phase 8
- **Issue:** `TestScorer_Algorithms_SortedAscending` exercises a single permutation (3 algorithms added in scrambled order). A property test that draws an arbitrary permutation of N algorithms and asserts the output is sorted would catch any future sort-key change.
- **Standard:** `.claude/skills/determinism-standards/SKILL.md` (sort-key completeness)
- **Action:** Code fix
- **Rationale:** Determinism guarantee on output ordering deserves a property gate.

---

### [Improvement] Test naming inconsistency — `Test<Algo>_Behaviour` vs `Test<AlgoScore>_Behaviour` vs `Test<Algo>Score_Behaviour`
- **File:** All `*_test.go`
- **Phase introduced:** Phases 2-8
- **Issue:** Patterns observed:
  - `TestLevenshtein_BothEmpty` (algorithm-name)
  - `TestStrcmp95Score_ZeroAllocs_ASCII_Short` (algorithm+Score-name)
  - `TestProp_LevenshteinScore_RangeBounds` (algorithm+Score-name)
  - `TestSoundex_BothEmpty` and `TestSoundexScore_NonMatch` (mixed in same file)
  - `TestDispatch_LevenshteinRegistered` (subject-first)
- **Standard:** `.claude/skills/go-testing-standards/SKILL.md` (no explicit pattern)
- **Action:** Skill clarification
- **Rationale:** Not blocking. But a project-wide convention (e.g. always `Test<PublicFn>_<Behaviour>`) would reduce cognitive load.
- **Suggested fix:** Establish convention in skill; do not rename existing tests retroactively.

### [Improvement] Benchmark naming inconsistency — `BenchmarkSoundexScore_ASCII_Identity` vs `BenchmarkSoundexScore_ASCII_Short` vs `BenchmarkMRAScore_Match`
- **File:** `*_bench_test.go`
- **Phase introduced:** Phases 2-8
- **Issue:** No consistent suffix convention. Most algorithms use `ASCII_Short / ASCII_Medium / ASCII_Long / Unicode_Short`, but phonetic algorithms use `Match / NoMatch / Identity` and Tokenise uses `DefaultOptions` / `StripDiacritics_Short`. Bench naming consistency affects `benchstat` output legibility.
- **Standard:** none codified
- **Action:** Skill clarification
- **Rationale:** Polish item; doesn't change correctness.

### [Improvement] Phonetic property tests lack `NoNegativeZero`
- **File:** `props_test.go` (Soundex, DoubleMetaphone, NYSIIS, MRA sections)
- **Phase introduced:** Phase 4
- **Issue:** Character-based and q-gram algorithms uniformly have `TestProp_<Algo>Score_NoNegativeZero`. Phonetic algorithms do not. Defensible (phonetic scores are 0.0 or 1.0 literals, no arithmetic), but the property is cheap to add for symmetry with the rest of the suite.
- **Standard:** `.claude/skills/determinism-standards/SKILL.md` (-0.0 handling)
- **Action:** Code fix or Skill clarification
- **Rationale:** Defensive completeness only.

### [Improvement] `TestScorer_ConcurrentSafety` runs once, not under `-count=N`
- **File:** `scorer_test.go:903`
- **Phase introduced:** Phase 8
- **Issue:** 100 goroutines × 3 methods × single test run. Under `-race -count=10` the test would surface more goroutine schedules.
- **Standard:** general concurrency-test hygiene
- **Action:** Skill clarification
- **Rationale:** Same as 08-TEST-ANALYSIS TEST-09.

### [Improvement] Match-benchmark sink gate is dead-code in `scorer_bench_test.go`
- **File:** `scorer_bench_test.go:171-173`
- **Phase introduced:** Phase 8
- **Issue:** `if sink < -1` on a non-negative counter is provably false; sufficiently aggressive optimisers could elide it. Same as 08-TEST-ANALYSIS TEST-12.
- **Standard:** locked benchmark pattern (PATTERNS.md)
- **Action:** Code fix
- **Rationale:** One-line cosmetic. Doesn't affect today's compiler but defeats the purpose of the locked sink-gate pattern.

### [Improvement] Golden file `scorer-default.json` has no single-character-pair entry
- **File:** `testdata/golden/scorer-default.json`
- **Phase introduced:** Phase 8
- **Issue:** 22 entries × 5 configs but no `"a" / "b"` row. Single-character is a documented edge case per `.claude/skills/go-testing-standards/SKILL.md`.
- **Standard:** as above
- **Action:** Code fix
- **Rationale:** Same as 08-TEST-ANALYSIS TEST-13.

### [Improvement] Golden file `0.9999999999999999` ULP-below-1.0 identity composite not documented
- **File:** `scorer_golden_test.go:148`
- **Phase introduced:** Phase 8
- **Issue:** The composite of 6 equal-weighted algorithms each returning 1.0 produces `0.9999999999999999` (one ULP below) because `1.0/6.0` is irrational in binary float. The golden file pins this correctly, but no comment explains it — reviewers might misread as a bug.
- **Standard:** general test-readability
- **Action:** Code fix
- **Rationale:** Same as 08-TEST-ANALYSIS TEST-14.

### [Improvement] Internal-test usage limited to scorer / q_gram / token_indel
- **File:** `scorer_internal_test.go`, `scorer_options_internal_test.go`
- **Phase introduced:** Phase 8
- **Issue:** Only 2 of the 23 algorithm packages have an `_internal_test.go`. q_gram_test.go and token_indel_test.go are in `fuzzymatch_test` (external) per the file headers but reference unexported names — actually they're internal-only via `_test.go` location. Strcmp95's similar-character table has `Strcmp95SimilarCharsEntryForTest` exposed via `export_test.go`; an internal test would be cleaner.
- **Standard:** `.claude/skills/go-testing-standards/SKILL.md` §Unit Tests (internal tests "where genuinely needed")
- **Action:** Code fix or Discuss-phase needed
- **Rationale:** `export_test.go` introduces test-only public symbols (`DispatchInvokeForTest`, `Strcmp95SimilarCharsEntryForTest`) that pollute the godoc surface; an internal test file avoids that pollution.
- **Suggested fix:** Migrate Strcmp95 similar-char table tests to a `strcmp95_internal_test.go` and remove the export shim.

### [Improvement] Fuzz corpus on disk lags fuzz harness functions (11 fuzzers have no committed seed corpus)
- **File:** `testdata/fuzz/`
- **Phase introduced:** Phases 2-6
- **Issue:** 31 Fuzz* functions in source; 20 corpus directories on disk. Fuzz harnesses without committed corpora can only seed from `f.Add()` calls in source — fine for first-run, but loses the regression value of CI-discovered failure inputs across builds.
- **Standard:** `.claude/skills/go-testing-standards/SKILL.md` §Fuzz Tests (corpus checked into testdata/fuzz/)
- **Action:** Code fix (run the nightly fuzz pipeline once, commit corpus)
- **Rationale:** Hygiene. Corpora persist crash regressions across PRs.

### [Improvement] Cross-validation vectors absent for character-based algorithms
- **File:** `testdata/cross-validation/`
- **Phase introduced:** Phases 2-3
- **Issue:** Cross-validation vectors exist for phonetic, swg, token-ratios, ratcliff-obershelp. Character-based algorithms (Levenshtein, Damerau-*, Hamming, Jaro, JaroWinkler, Strcmp95, LCSStr) have only inline unit-test reference vectors with no JSON-vector cross-validation set. The primary-source-citation-correctness reviews can sign off without this, but cross-validation against (e.g.) C++ libraries or `python-Levenshtein` adds independent confirmation.
- **Standard:** `.claude/skills/algorithm-correctness-standards/SKILL.md` §Reference Vectors
- **Action:** Discuss-phase needed
- **Rationale:** Not strictly required by the skill but improves confidence at v1.0.

### [Improvement] No coverage gate in CI workflow
- **File:** `.github/workflows/*.yml`
- **Phase introduced:** Phase 1
- **Issue:** The missing `internal_coverage_test.go` would catch the floor failure on every PR; without it, CI silently runs `go test -cover` but doesn't fail on the 90.4% / 95% gap. Confirms with the [Critical] meta-test finding.
- **Standard:** `.claude/skills/go-testing-standards/SKILL.md`; `docs/requirements.md` §15.8
- **Action:** Code fix (CI + meta-test)
- **Rationale:** Detection lag — the headline coverage gap has likely persisted across multiple PRs without notice.

### [Improvement] `WithSmithWatermanGotohAlgorithm` parameter-validation tests thin (only `_HappyPath`, `_CapturesParams`, `_InvalidWeight`)
- **File:** `scorer_options_test.go:506-526`
- **Phase introduced:** Phase 8
- **Issue:** `NewSWGParams()` has multiple fields (Match, Mismatch, Gap, etc. per `swg.go`). Only one parameter combination is exercised; no tests for negative `Match`, zero `Gap`, or NaN/Inf in any field.
- **Standard:** `.claude/skills/go-testing-standards/SKILL.md`
- **Action:** Code fix

### [Improvement] No bench for `WithoutNormalisation` Scorer vs `WithNormalisation` Scorer (the allocation-comparison gate)
- **File:** `scorer_bench_test.go`
- **Phase introduced:** Phase 8
- **Issue:** Same as 08-TEST-ANALYSIS TEST-11. Without the comparator, the WR-04 allocation-reduction work has no visibility.
- **Standard:** `.claude/skills/performance-standards.md`
- **Action:** Code fix

---

## Per-Concern Status Summary

### Per-Algorithm Catalogue
- **COVERED** — All 23 algorithms have unit + property + at least byte-path fuzz + benchmark + BDD + cross-platform golden.
- **PARTIALLY COVERED** — 9 algorithms missing `*Runes` fuzz; 5 algorithms missing Distance fuzz; phonetic algorithms missing `Code/Keys` fuzz; 8 algorithms missing very-long-input unit tests.

### Per-Scorer Concern
- **COVERED** — Every `With*` option happy-path + invalid-weight + invalid-N tested; Score / ScoreAll / Match / Threshold / Algorithms / DefaultScorer / DefaultScorerOptions all exercised; concurrent safety verified; deterministic-across-runs property tested; golden file with 22×5 configs.
- **MISSING** — Scorer-level fuzz; DefaultScorer Identity / Symmetry / `Score = Σ Weight·ScoreAll` properties; NaN-threshold gate; NaN/Inf-weight gate; Tversky α=β=0 option-layer gate; `Scorer.Score` examples; coverage shortfall on `DefaultScorer` (75%) panic-line.

### Per-Scan Concern (Phase 6 — pre-implementation)
- **MISSING** — Entire scan sub-package is unimplemented. All scan-checklist items (within-group, cross-group, suppression composition, sort-key determinism, performance budgets, BDD scenarios) are deferred until Phase 6 lands.

### Per-Normalisation Concern
- **COVERED** — 8 `TestNormalise_*` unit tests + 3 property tests + golden file (`normalisation.json`); ASCII fast path documented and benchmarked.
- **MISSING** — `tests/bdd/features/normalisation.feature` does not exist.

### Per-Tokenisation Concern
- **COVERED** — 13 `TestTokenise_*` unit tests + 6 property tests; `DefaultTokeniseOptions` honoured.
- **PARTIALLY COVERED** — Bench coverage exists but no rune-vs-byte ASCII-fast-path side-by-side.

### Coverage Thresholds
- **Overall:** 90.4% (target 95%) — **FAILS**
- **Per-file:** `double_metaphone.go` 83.8% (target 90%) — **FAILS**; all others ≥ 90%
- **Public API:** ~99% (only the unreachable `panic` line in `DefaultScorer` and a few defensive Step-checks in `NewScorer` are uncovered)

### Property-test depth
- **COVERED** for: Range, Identity, Symmetry, Triangle (where applicable), NoNaN, NoInf, NoNegativeZero (except phonetic).
- **MISSING** for: Scorer Identity / Symmetry / Score=ΣWeight·ScoreAll; `Algorithms()` sort property over permutations.
- **AT RISK** (generator flake): `TestProp_MongeElkanScore_AsymmetricWhenTokenCountAsymmetric` premise inconsistent with algorithm's tokeniser.

### Fuzz harness coverage
- **COVERED** for: 23 algorithm Score functions (byte path), Normalise, Tokenise, 4 Q-gram rune variants, MongeElkan asymmetric + symmetric.
- **MISSING** for: 9 character-based `*Runes` Score variants; all Distance functions; LongestCommonSubstring (returns string); phonetic Code/Keys variants; Scorer.Score / ScoreAll / Match; NewScorer with random options.

### BDD scenario coverage
- **COVERED** for: every algorithm; scorer composition (12 scenarios across 12 mandatory classes).
- **MISSING** for: `normalisation.feature`, `determinism.feature`, `scan.feature`, `suppression.feature`, `algorithms.feature` (single combined alternative — currently per-algo files instead).
- **MISSING tags:** `@character`, `@qgram`, `@gestalt`, `@determinism`, `@normalisation`.

### Cross-validation evidence persistence
- **COVERED** for: phonetic (4 algorithms), swg, token-ratios (3 algorithms), ratcliff-obershelp.
- **MISSING** for: 9 character-based algorithms (Levenshtein, Damerau-*, Hamming, Jaro, JaroWinkler, Strcmp95, LCSStr) and 4 q-gram algorithms.

### Edge-case sweep
- **Both-empty / one-empty / identical:** COVERED across all algorithms.
- **Unicode (multi-byte UTF-8, CJK, emoji):** COVERED for algorithms with `*Runes` variants.
- **Malformed UTF-8:** COVERED via fuzz harnesses (byte path) for algorithms with fuzz; MISSING on Scorer surface.
- **Very-long input (1000+):** PARTIALLY COVERED — Jaro has it, Cosine has a 1000-iteration determinism test, others rely on benchmarks-as-tests.
- **Length-mismatch (Hamming-specific):** COVERED with silent-zero policy.

### Concurrent-safety tests
- **COVERED** — `TestScorer_ConcurrentSafety` runs 100 goroutines × 3 methods under `-race`. Race tests pass (~352s with `-race`).

### Bench naming consistency
- **PARTIALLY COVERED** — Most benchmarks follow `Benchmark<Algo><Method>_<Path>_<Length>`. Phonetic, MRA-pathological, partial-ratio adversarial benches use bespoke suffixes.

### `quick.MaxCount` settings
- **DEFAULTS USED EVERYWHERE** except `tokenise_test.go:506` (MaxCount=200) and one DamerauOSA constrained config. Property-test confidence is thinner than it could be.

### Generator pitfalls
- **TestProp_MongeElkanScore_AsymmetricWhenTokenCountAsymmetric:** premise based on `strings.Fields` differs from algorithm's `Tokenise`. Stale failure in `bench.txt.new`.
- **TestProp_Scorer_WeightSumOne:** `uint16` overflow — `u==65535 → u+1==0`. Per-run flake probability ~0.46%.

### Test file organisation
- **Mostly external** (`package fuzzymatch_test`). Internal test files: `scorer_internal_test.go`, `scorer_options_internal_test.go`, `q_gram_test.go` (uses external pkg via export_test shim), `token_indel_test.go` (same).
- **`export_test.go` shim pattern** used to expose dispatch/internal state — slight pollution of godoc surface.

### Flaky-test patterns
- **WR-01 `uint16` overflow** still open (`scorer_test.go:827`).
- **MongeElkan asymmetry-conditional property** has documented flake risk on exotic-Unicode inputs (`bench.txt.new` artefact).
- **`TestScorer_ConcurrentSafety`** runs once; goroutine-schedule-dependent bugs would surface only intermittently under `-count`.

---

## Recommendation

**Coverage gaps to address before milestone release.**

The test suite is large (~480 Test functions, 31 fuzzers, 127
benchmarks, 169 BDD scenarios) and structurally sound for the 23
algorithm catalogue. Every algorithm clears the per-algorithm
minimum: unit tests with reference vectors, property tests for the
mathematical invariants, fuzz harness for byte-path Score, benchmark,
BDD scenario, cross-platform golden. The Scorer surface gets
exhaustive option-layer testing.

The headline gaps are:

1. **Three meta-tests missing** (`internal_coverage_test.go`,
   `documentation_test.go`, `readme_shop_front_test.go`) — without
   `internal_coverage_test.go` the project's own coverage-floor gate
   is unenforced, which is why the 90.4% overall shortfall has
   persisted.
2. **Four still-open Phase-8 defects** — `WithThreshold(NaN)`,
   `WithAlgorithm(_, NaN/Inf)`, `WithTverskyAlgorithm(0, 0)`, and the
   `uint16` overflow flake — all small fixes already identified by
   08-REVIEW but not yet landed.
3. **Scorer fuzz harness missing** entirely — the single most-used
   public surface has no fuzz gate.
4. **Three Scorer-level property tests missing** — Identity, Symmetry,
   `Score = Σ Weight·ScoreAll`.
5. **`double_metaphone.go` at 83.8%** drags overall coverage below the
   95% floor — needs language-branch reference vectors.
6. **Scorer benchmarks not in `bench.txt`** — regression gate inactive.
7. **9 character-based `*Runes` and all Distance-function fuzz
   harnesses absent** — rune-path bugs would slip past CI.
8. **Stale `bench.txt.new`** artefact — investigate the MongeElkan
   asymmetric-property generator and either tighten it or delete the
   committed failure log.

**Not ready for milestone release** until items 1-6 land. Items 7-8
are important hardening work that can run in parallel and should land
before the v1.0.0 cut.

After those items land — particularly `internal_coverage_test.go`,
the Scorer fuzz harnesses, the four option-layer bug fixes, and the
coverage shortfall — the test suite meets the standards and the
milestone-release path is clear.

_Analysed: 2026-05-17_
_Analyst: test-analyst (whole-codebase scope, phases 1-8)_

</details>

---

## 15. test-writer (architecture-only review — no test files written)

_Source: `.planning/reviews/test-architecture-FINDINGS.md`_

<details>
<summary>Click to expand full report</summary>

---
status: issues_found
agent: test-writer
scope: test architecture gaps (phases 1-8)
reviewed: 2026-05-17T06:39:10Z
finding_counts:
  critical: 18
  important: 21
  improvement: 8
  total: 47
---

# Test Architecture Gaps — Phases 1–8

**Reviewed:** 2026-05-17
**Method:** static analysis of all `*_test.go` files, `testdata/`, `tests/bdd/features/`, `export_test.go`, and `example_test.go` against the requirements in `.claude/skills/go-testing-standards/SKILL.md` and `.claude/skills/algorithm-correctness-standards/SKILL.md`.
**Coordination note:** BDD-layer findings already captured in `bdd-scenario-FINDINGS.md` are not repeated here. This review focuses on unit, property, fuzz, benchmark, example, meta-test, and cross-validation gaps.

---

## Counts at a Glance

| Category | Public functions | Fuzz harnesses | On-disk corpus dirs | Example funcs |
|----------|-----------------|----------------|---------------------|---------------|
| Total public API | 79 | 31 (harnesses) | 20 (dirs) | 39 |

Fuzz coverage: 31 harnesses but only 20 on-disk `testdata/fuzz/` directories and numerous public functions with no fuzz exposure at all.

---

## Critical Findings

### [Critical] Fuzz harnesses for Distance and ScoreRunes variants of character-tier algorithms have NO fuzz coverage

- **File:** `levenshtein_fuzz_test.go`, `hamming_fuzz_test.go`, `jaro_fuzz_test.go`, `jarowinkler_fuzz_test.go`, `damerau_osa_fuzz_test.go`, `damerau_full_fuzz_test.go`
- **Phase introduced:** Phase 2
- **Issue:** Six character-tier fuzz harnesses each call only one public function — the byte-path Score variant. The following public functions receive zero fuzz coverage:
  - `LevenshteinDistanceRunes`, `LevenshteinScoreRunes`, `LevenshteinDistance` (fuzz only covers `LevenshteinScore`)
  - `HammingDistanceRunes`, `HammingScoreRunes`, `HammingDistance` (fuzz only covers `HammingScore`)
  - `JaroScoreRunes` (fuzz only covers `JaroScore`)
  - `JaroWinklerScoreRunes` (fuzz only covers `JaroWinklerScore`)
  - `DamerauLevenshteinOSAScoreRunes`, `DamerauLevenshteinOSADistance`, `DamerauLevenshteinOSADistanceRunes` (fuzz only covers `DamerauLevenshteinOSAScore`)
  - `DamerauLevenshteinFullScoreRunes`, `DamerauLevenshteinFullDistance`, `DamerauLevenshteinFullDistanceRunes` (fuzz only covers `DamerauLevenshteinFullScore`)
  Contrast with the correct multi-function pattern: `swg_fuzz_test.go` (FuzzSmithWatermanGotohScore) covers all six SWG public functions, and `lcsstr_fuzz_test.go` covers all four LCSStr public functions.
- **Standard:** `go-testing-standards/SKILL.md` §Fuzz Tests — "Every public function has a fuzz harness."
- **Action:** Code fix — expand each existing character-tier fuzz harness to call all public variants, following the SWG multi-function pattern.
- **Rationale:** A panic-inducing bug in a Rune or Distance variant will not be caught by the fuzzer. These paths handle malformed UTF-8 (which the go fuzzer routinely generates) and are the most crash-prone.

---

### [Critical] 11 fuzz harnesses have no on-disk seed corpus directory in `testdata/fuzz/`

- **File:** `testdata/fuzz/` (missing directories for listed harnesses)
- **Phase introduced:** Phases 6 and 7
- **Issue:** The following fuzz harnesses exist in source but have no corresponding `testdata/fuzz/<FuzzName>/` directory with a seed file:
  - `FuzzDoubleMetaphone`
  - `FuzzMongeElkanScore`
  - `FuzzMongeElkanScoreSymmetric`
  - `FuzzMRA`
  - `FuzzNYSIIS`
  - `FuzzPartialRatioScore`
  - `FuzzPartialRatioScoreRunes`
  - `FuzzSoundex`
  - `FuzzTokenJaccardScore`
  - `FuzzTokenSetRatioScore`
  - `FuzzTokenSortRatioScore`
  Without an on-disk seed directory, the fuzzer starts from an empty corpus and the nightly CI runs produce no reproducible crash evidence. The 20 existing Phase 2–5 harnesses each have a `seed-001` file.
- **Standard:** `go-testing-standards/SKILL.md` §Fuzz Tests — "Corpus checked into `testdata/fuzz/`. CI runs 60 seconds per fuzzer per build."
- **Action:** Code fix — create `testdata/fuzz/<FuzzName>/seed-001` files for each listed harness, using the programmatic seed pairs already present in the `f.Add(...)` calls within the harness body.
- **Rationale:** Without persisted seeds the fuzzer restarts cold on every CI run and cannot accumulate crash-inducing corpus entries between runs. This also means `go test -run=FuzzXxx ./...` (seed-only mode without `-fuzz` flag) exercises zero deterministic cases.

---

### [Critical] No Scorer-level fuzz harness exists

- **File:** missing: `scorer_fuzz_test.go`
- **Phase introduced:** Phase 8 (Scorer)
- **Issue:** `Scorer.Score`, `Scorer.Match`, and `Scorer.ScoreAll` are public methods on a public type. No fuzz harness exercises them. The `DefaultScorer()` and `NewScorer()` functions are also public. The testing standard requires one `Fuzz*` per public function. The Scorer is the highest-value composition surface and the one most likely to surface panics via unexpected normalisation interactions or floating-point edge cases on arbitrary input pairs.
- **Standard:** `go-testing-standards/SKILL.md` §Fuzz Tests — "Every public function has a fuzz harness."
- **Action:** Code fix — create `scorer_fuzz_test.go` with a `FuzzScorer_Score` harness that constructs `DefaultScorer()` once and fuzzes `Score(a, b)` asserting range bounds, no NaN, no Inf.
- **Rationale:** The Scorer is the integration point for all 23 algorithm dispatch paths. Arbitrary inputs reaching the Scorer exercise normalisation, all dispatched algorithms, and the weighted-sum reduction simultaneously.

---

### [Critical] Missing property tests for Rune-variant full invariant suite on character-tier algorithms

- **File:** `props_test.go`
- **Phase introduced:** Phase 2
- **Issue:** For the following algorithms, the `*Runes` score variant has ONLY a `Symmetric` property test. The full invariant set (RangeBounds, Identity, NoNaN, NoInf, NoNegativeZero) is absent:
  - `LevenshteinScoreRunes` — only `TestProp_LevenshteinScoreRunes_Symmetric` exists
  - `HammingScoreRunes` — only `TestProp_HammingScoreRunes_Symmetric`
  - `JaroScoreRunes` — only `TestProp_JaroScoreRunes_Symmetric`
  - `JaroWinklerScoreRunes` — only `TestProp_JaroWinklerScoreRunes_Symmetric`
  - `DamerauLevenshteinOSAScoreRunes` — only `TestProp_DamerauLevenshteinOSAScoreRunes_Symmetric`
  - `DamerauLevenshteinFullScoreRunes` — only `TestProp_DamerauLevenshteinFullScoreRunes_Symmetric`
  - `SmithWatermanGotohScoreRunes` — only `TestProp_SmithWatermanGotohScoreRunes_Symmetric`
  Compare: `LCSStrScoreRunes` has the full set (RangeBounds, Identity, Symmetric, NoNaN, NoInf, NoNegativeZero).
- **Standard:** `go-testing-standards/SKILL.md` §Property Tests — "Required per-algorithm properties: Identity, Range bounds, Symmetry, Never panics." `algorithm-correctness-standards/SKILL.md` §Mathematical Invariants — "All algorithms: Range bounds, Identity, Never panics."
- **Action:** Code fix — add the missing property tests for each listed Runes variant.
- **Rationale:** The rune path handles multi-byte UTF-8 code points and is a distinct code path from the byte path. A NaN or out-of-range return from the rune path would not be caught by the byte-path property tests.

---

### [Critical] Missing property tests for Distance variants of Levenshtein, Hamming, and Damerau algorithms (Runes triangle inequality)

- **File:** `props_test.go`
- **Phase introduced:** Phase 2
- **Issue:** The testing standard mandates a triangle-inequality property test for distance-based algorithms. While byte-path triangle inequality tests exist for `LevenshteinDistance`, `HammingDistance`, `DamerauLevenshteinOSADistance`, and `DamerauLevenshteinFullDistance`, none of the corresponding `*DistanceRunes` functions have triangle-inequality tests:
  - `TestProp_LevenshteinDistanceRunes_TriangleInequality` — missing
  - `TestProp_HammingDistanceRunes_TriangleInequality_EqualLength` — missing
  - `TestProp_DamerauLevenshteinOSADistanceRunes_TriangleInequality` — missing
  - `TestProp_DamerauLevenshteinFullDistanceRunes_TriangleInequality` — missing
  The rune-path implementations use independent code paths (rune conversion then the same DP, but byte boundary handling differs). A regression in rune conversion that inflates distance values would violate triangle inequality.
- **Standard:** `go-testing-standards/SKILL.md` §Property Tests — "Triangle inequality (for distance-based algorithms)." `algorithm-correctness-standards/SKILL.md` §Mathematical Invariants — "Distance-based algorithms: Triangle inequality."
- **Action:** Code fix — add four `TestProp_*DistanceRunes_TriangleInequality` property tests.
- **Rationale:** The rune-path distance functions are independent code paths exercised by a different test surface. A rune-conversion bug that changes distance outcomes would not be caught by the byte-path triangle inequality test.

---

### [Critical] No cross-validation corpus exists for character-tier algorithms (Levenshtein, Jaro, JaroWinkler, Hamming, Strcmp95, DL-OSA, DL-Full)

- **File:** `testdata/cross-validation/` (missing: `levenshtein/`, `jaro/`, `jarowinkler/`, `hamming/`, `strcmp95/`, `damerau-osa/`, `damerau-full/`)
- **Phase introduced:** Phases 2 and 3
- **Issue:** Only four cross-validation corpora exist: `swg/` (biopython), `ratcliff-obershelp/` (difflib), `token-ratios/` (RapidFuzz 3.14.5), `phonetic/` (jellyfish 1.2.1 + Metaphone 0.6). The seven character-tier algorithms implemented in Phases 2–3 have no external-reference corpus. The unit tests rely entirely on manually-transcribed literature reference vectors (which can contain transcription errors) rather than cross-validated outputs from independent implementations (jellyfish, python-Levenshtein, rapidfuzz, or NLTK). Per the algorithm-correctness-standards skill, reference vector cross-validation against existing MIT/BSD implementations is expected.
- **Standard:** `algorithm-correctness-standards/SKILL.md` §Fresh-Implementation Discipline — "The implementation may study existing MIT/BSD/Apache-licensed Go implementations for reference vector cross-validation only — to confirm that a canonical example produces the expected output." `go-testing-standards/SKILL.md` §BDD Tests cites SWG/EMBOSS and Ratcliff/difflib as the pattern to follow for all algorithms.
- **Action:** Code fix — create generator scripts and corpora for at minimum: `levenshtein/` (jellyfish), `jaro/` (jellyfish), `jarowinkler/` (jellyfish). Hamming and DL-OSA/Full are covered adequately by literature vectors; Strcmp95 and LCSStr have no widely-available Python reference and can defer.
- **Rationale:** The current reference-vector-only approach cannot detect silent transcription errors in the test data itself. A wrong reference vector masks a wrong implementation. The dual-pin pattern (literature vector + independent implementation) is the only way to detect both types of error.

---

### [Critical] No cross-validation corpus exists for q-gram tier algorithms (QGramJaccard, SorensenDice, Cosine, Tversky)

- **File:** `testdata/cross-validation/` (missing: `qgram/`)
- **Phase introduced:** Phase 5
- **Issue:** The four q-gram algorithms use literature reference vectors only (Ukkonen 1992, textbook bigrams, Salton & McGill 1983). No external cross-validation corpus exists against jellyfish, sklearn, or any other reference implementation. The cosine test has an unusual design decision: `cosine_test.go:25` explicitly notes "cross-validation density that would otherwise come from an external corpus" is handled inline — this is a deviation from the project pattern and creates a single point of failure (if the test author's calculation is wrong, both the implementation and the test are wrong simultaneously).
- **Standard:** `algorithm-correctness-standards/SKILL.md` §Fresh-Implementation Discipline — cross-validation against independent implementations expected.
- **Action:** Discuss-phase needed — determine which reference implementation (sklearn, scipy, jellyfish) to use for q-gram cross-validation; create corpus and test.
- **Rationale:** Tversky, Dice, and Cosine have subtle normalisation choices (multiset vs set semantics, |A|×|B| vs |A|+|B|) that differ between implementations. A cross-validation corpus would catch a wrong normalisation convention that still passes the literature vectors.

---

### [Critical] No cross-validation corpus for Monge-Elkan

- **File:** `testdata/cross-validation/` (missing: `monge-elkan/`)
- **Phase introduced:** Phase 6
- **Issue:** Monge-Elkan is a composite algorithm; its outputs depend on the inner metric and the asymmetric/symmetric averaging. No external reference corpus exists. The jellyfish library does not implement Monge-Elkan, but strsimpy and py_stringmatching do. The current tests rely entirely on hand-computed reference vectors in the test file, which share the same author as the implementation.
- **Standard:** `algorithm-correctness-standards/SKILL.md` §Fresh-Implementation Discipline.
- **Action:** Discuss-phase needed — determine appropriate reference implementation (strsimpy, py_stringmatching); create a small corpus covering at least JaroWinkler inner and Levenshtein inner.
- **Rationale:** The symmetric averaging of MongeElkan is unusual and could silently produce wrong values if the mean formula has a sign error. An independent reference corpus is the only defence-in-depth check.

---

### [Critical] 9 staging golden files exist but have no corresponding staging test in `algorithms_golden_test.go`

- **File:** `algorithms_golden_test.go` (missing: TestGolden_DoubleMetaphone_Staging, TestGolden_MongeElkan_Staging, TestGolden_MRA_Staging, TestGolden_NYSIIS_Staging, TestGolden_PartialRatio_Staging, TestGolden_Soundex_Staging, TestGolden_TokenJaccard_Staging, TestGolden_TokenSetRatio_Staging, TestGolden_TokenSortRatio_Staging)
- **Phase introduced:** Phases 6 and 7
- **Issue:** The `testdata/golden/_staging/` directory contains 23 JSON files (one per algorithm), but `algorithms_golden_test.go` implements staging tests for only 15. The nine Phase 6–7 algorithms (DoubleMetaphone, MongeElkan, MRA, NYSIIS, PartialRatio, Soundex, TokenJaccard, TokenSetRatio, TokenSortRatio) have JSON files in `_staging/` but no test function that reads and asserts those files. The `TestGolden_Algorithms_Merge` function assembles the promoted `algorithms.json` from staged files, but without staged tests the per-algorithm correctness of the JSON content is unverified before promotion. The golden files for Phase 7 phonetic algorithms are particularly risky: the phonetic codes golden test (`phonetic_codes_golden_test.go`) covers code strings but the staging golden JSON carries score vectors that need a separate test.
- **Standard:** `go-testing-standards/SKILL.md` §Meta-tests — golden file pattern; `determinism-standards/SKILL.md` — cross-platform stability gate.
- **Action:** Code fix — add nine `TestGolden_*_Staging` test functions in `algorithms_golden_test.go`, following the established pattern of existing staging tests.
- **Rationale:** Without staging tests, the JSON content in `_staging/` for Phase 6–7 algorithms is generated-and-forgotten. A wrong value in a staging file would be promoted into `algorithms.json` without any assertion failing.

---

### [Critical] `scorer_options_test.go` is in `package fuzzymatch` (internal) but named without `_internal_` convention

- **File:** `scorer_options_test.go`
- **Phase introduced:** Phase 8
- **Issue:** The testing standard says internal tests should follow the naming convention `levenshtein_internal_test.go` with `package fuzzymatch`. `scorer_options_test.go` is in `package fuzzymatch` (it uses the `applyOptionForProbe` internal helper) but is not named with `_internal_`. This causes confusion: a reader sees `scorer_options_test.go` and expects `package fuzzymatch_test` (external), but finds `package fuzzymatch`. `scorer_internal_test.go` follows the correct naming convention. The naming inconsistency creates a false read on the internal vs external boundary. For Phase 9, the scan package will need an analogous internal/external split; the inconsistent naming here creates a confusing template.
- **Standard:** `go-testing-standards/SKILL.md` §Unit Tests — "Internal tests where genuinely needed for unexported invariants: `levenshtein_internal_test.go` with `package fuzzymatch`."
- **Action:** Code fix — rename `scorer_options_test.go` → `scorer_options_internal_test.go`. This is a mechanical rename; the content is correct.
- **Rationale:** The naming convention signals test boundary to readers and to the codebase scan tools. An inconsistency here will propagate to future phases.

---

### [Critical] Missing `internal_coverage_test.go` meta-test for coverage floor enforcement

- **File:** missing: `internal_coverage_test.go`
- **Phase introduced:** Phase 1 (should have been created at bootstrap)
- **Issue:** The testing standard mandates an `internal_coverage_test.go` meta-test that enforces the coverage floor: ≥ 95% overall, ≥ 90% per file, 100% on public API surface. This file does not exist. Coverage targets are therefore tracked only informally via `make coverage`. A regression in coverage (e.g. a new branch added to a hot path without a corresponding test) will not cause `go test ./...` to fail.
- **Standard:** `go-testing-standards/SKILL.md` §Meta-tests — "`internal_coverage_test.go` — enforces the coverage floor (95% overall, 90% per file, 100% public API)."
- **Action:** Code fix — create `internal_coverage_test.go` using `testing.Coverage()` (Go 1.20+) or the `go test -cover` output parsing pattern to assert the floor at test time.
- **Rationale:** Without a programmatic floor, coverage drift is invisible until someone manually runs `make coverage` and reads the output. Automated enforcement is the only way to guarantee the 100%-public-API target doesn't silently erode.

---

### [Critical] Missing `readme_shop_front_test.go` meta-test

- **File:** missing: `readme_shop_front_test.go`
- **Phase introduced:** Phase 1 (should have been created at bootstrap)
- **Issue:** The testing standard mandates a `readme_shop_front_test.go` that compiles and runs the README's headline quick-start example, asserting the output matches the documented expected output byte-for-byte. `README.md` contains code examples. If the API surface changes (e.g. a function renamed by the api-ergonomics-reviewer) the README will drift silently. The existing `ai_friendly_test.go` checks that symbols appear in `llms.txt` but does not execute any README examples.
- **Standard:** `go-testing-standards/SKILL.md` §Meta-tests — "`readme_shop_front_test.go` — README's headline example compiles and runs, output exactly matches documented expected output."
- **Action:** Code fix — create `readme_shop_front_test.go` with a `TestREADME_QuickStartExample` that uses `go/types` or an exec approach to compile and run the quick-start snippet, asserting output.
- **Rationale:** The README is the primary consumer touch point. Silent divergence between README examples and real API behaviour erodes trust and makes the library harder to adopt. The test forces README synchronisation with every API change.

---

### [Critical] Missing `documentation_test.go` meta-test for `docs/*.md` code examples

- **File:** missing: `documentation_test.go`
- **Phase introduced:** Phase 1 (should have been created at bootstrap)
- **Issue:** The testing standard mandates a `documentation_test.go` that ensures README and `docs/*.md` code examples compile and produce documented output. The `docs/` directory contains `algorithms.md`, `scorer.md`, `scan.md`, `tuning.md`, `extending.md`, `performance.md`, `faq.md`. Code blocks in these files can silently diverge from the implementation. There is currently no automation checking documentation code examples.
- **Standard:** `go-testing-standards/SKILL.md` §Meta-tests — "`documentation_test.go` — README and `docs/*.md` code examples compile and produce documented output."
- **Action:** Code fix — create `documentation_test.go`. At minimum, it should parse `docs/*.md` for fenced Go code blocks and assert each block compiles with `go/types`. Execution-level assertions can be added incrementally.
- **Rationale:** Documentation drift is the most common consumer pain point for open-source libraries. A compilation-level gate is a low-friction, high-value safeguard.

---

### [Critical] No concurrent `NewScorer` construction test

- **File:** `scorer_test.go`
- **Phase introduced:** Phase 8
- **Issue:** `TestScorer_ConcurrentSafety` (line 903) correctly tests concurrent `Score`, `ScoreAll`, and `Match` calls on an already-constructed `*Scorer`. It does not test concurrent construction of multiple `*Scorer` instances via `NewScorer()`. If `NewScorer` mutates any package-level state (e.g. a singleton or a shared config map) during construction, concurrent calls could race. The `go test -race` flag would catch this, but there is no explicit test that runs `NewScorer` concurrently to validate its thread-safety contract is stated in its godoc.
- **Standard:** `go-testing-standards/SKILL.md` §Scorer-level properties — "Deterministic across runs: `Scorer.Score` is byte-identical for repeated calls"; concurrent safety is a documented guarantee.
- **Action:** Code fix — add `TestScorer_ConcurrentNewScorer` that spawns 50 goroutines each calling `NewScorer(DefaultScorerOptions()...)` simultaneously and asserts no panic and correct construction.
- **Rationale:** Concurrent `NewScorer` calls may happen in library consumers initialising multiple Scorer instances at startup (e.g. Cassandra node bringing up multiple SAI indexes simultaneously). The absence of a test leaves this guarantee unverified at the test layer.

---

### [Critical] Missing phonetic-algorithm cross-algorithm convergence tests in `cross_algorithm_consistency_test.go`

- **File:** `cross_algorithm_consistency_test.go`
- **Phase introduced:** Phase 7
- **Issue:** `cross_algorithm_consistency_test.go` covers identity-convergence, both-empty-convergence, and one-empty-convergence for Phase 2–3 character-tier algorithms only (Levenshtein, DL-OSA, DL-Full, Hamming, Jaro, JaroWinkler, SWG). It does not include the Phase 7 phonetic algorithms (Soundex, DoubleMetaphone, NYSIIS, MRA). The algorithm-correctness-standards skill mandates consistent both-empty → 1.0 and one-empty → 0.0 behaviour across the entire catalogue. Phonetic algorithms have a distinct empty-input rule: `SoundexScore("", "")` = 1.0 (two equal empty codes match), `SoundexScore("Robert", "")` = 0.0. There is no cross-algorithm test asserting this convention holds consistently across all four phonetic algorithms and comparing against the non-phonetic convention.
- **Standard:** `algorithm-correctness-standards/SKILL.md` §Edge cases — "Both inputs empty: return 1.0 by convention. One input empty: return 0.0 for distance-based and set-based algorithms. Phonetic algorithms return 1.0 IFF both encoded codes are empty."
- **Action:** Code fix — add a `TestCrossAlgorithm_PhoneticBothEmpty_Convergence` and `TestCrossAlgorithm_PhoneticOneEmpty_Scores0` to `cross_algorithm_consistency_test.go` covering all four phonetic score functions.
- **Rationale:** The phonetic both-empty convention is documented differently from character-tier conventions (MRACompare returns matched=true, sim=6 on both-empty). A single test that pins all four phonetic `Score` functions on the same input pair catches inconsistencies between implementations.

---

### [Critical] Missing cross-algorithm convergence for Phase 6 token-tier algorithms in identity/both-empty tests

- **File:** `cross_algorithm_consistency_test.go`
- **Phase introduced:** Phase 6
- **Issue:** `TestCrossAlgorithm_IdentityConvergence` (line 107) and `TestCrossAlgorithm_BothEmptyConvergence` (line 138) enumerate only the seven Phase 2–3 algorithms. They do not include the Phase 6 token-tier algorithms (TokenSortRatio, TokenJaccard, MongeElkanScoreSymmetric) or the Phase 5 q-gram tier (QGramJaccard, SorensenDice, Cosine). The LOCKED deviation for `TokenSetRatioScore` (returns 0.0 on both-empty) is covered by `TestCrossAlgorithm_TokenSetRatio_EmptyDeviation_PinnedAgainstTokenJaccard` but the identity and both-empty conventions for TokenSortRatio, TokenJaccard, and MongeElkan are not asserted in the cross-algorithm convergence table.
- **Standard:** `algorithm-correctness-standards/SKILL.md` §Edge cases.
- **Action:** Code fix — extend the `funcs` slices in `TestCrossAlgorithm_IdentityConvergence` and `TestCrossAlgorithm_BothEmptyConvergence` to include all catalogue score functions that follow the standard convention. Document the TokenSetRatioScore deviation as a comment exclusion.
- **Rationale:** A cross-algorithm convergence table that covers only 7 of 23 algorithms is incomplete as a regression guard. Adding Phase 5–6 algorithms to the table with no code change other than slice extension is low-cost, high-value coverage.

---

### [Critical] Missing Scorer-level "composite bounded by per-algorithm min/max" property test

- **File:** `scorer_test.go`
- **Phase introduced:** Phase 8
- **Issue:** The testing standard lists three Scorer-level property tests required: "Deterministic across runs", "Range bounds", and "Composite ≥ min, ≤ max: weighted composite is bounded by the per-algorithm min and max scores." The first two (`TestProp_Scorer_DeterministicAcrossRuns`, `TestProp_Scorer_ScoreInRange`) exist. The third — asserting that the composite weighted score is never below the minimum individual algorithm score and never above the maximum individual algorithm score when weights are normalised — is absent. This invariant catches implementation bugs where a weight > 1.0 amplifies a score beyond the per-algorithm maximum.
- **Standard:** `go-testing-standards/SKILL.md` §Property Tests — "Scorer-level properties: Composite ≥ min, ≤ max: weighted composite is bounded by the per-algorithm min and max scores."
- **Action:** Code fix — add `TestProp_Scorer_CompositeBoundedByPerAlgoMinMax` using `testing/quick`.
- **Rationale:** The weight normalisation guarantee (weights sum to 1.0) is tested, but the resulting property — that the composite is a convex combination of per-algorithm scores — is not. If normalisation had a rounding bug that let weights sum to 1.001, scores above 1.0 would appear.

---

### [Critical] Missing `ScoreAll` determinism property test at Scorer level

- **File:** `scorer_test.go`
- **Phase introduced:** Phase 8
- **Issue:** `TestProp_Scorer_DeterministicAcrossRuns` tests that `Scorer.Score` returns the same float64 on repeated calls. No equivalent property test exists for `Scorer.ScoreAll`. The `ScoreAll` method returns a `map[AlgoID]float64`; map allocation itself is deterministic (same keys each call) but the per-value floating-point computation must also be deterministic. This is particularly important because `ScoreAll` exposes the per-algorithm dispatch layer directly — any non-determinism in a dispatched function would appear in `ScoreAll` before it appears in `Score`.
- **Standard:** `go-testing-standards/SKILL.md` §Property Tests — "Deterministic across runs: `Scorer.Score` is byte-identical for repeated calls."
- **Action:** Code fix — add `TestProp_Scorer_ScoreAll_DeterministicAcrossRuns` that calls `s.ScoreAll(a, b)` twice and asserts each key's value is bit-identical across both calls.
- **Rationale:** `ScoreAll` is the per-algorithm diagnostic surface and the one most likely to expose non-determinism first, since it bypasses the weighted-sum reduction.

---

## Important Findings

### [Important] Missing example tests for 33 public functions

- **File:** `example_test.go`
- **Phase introduced:** Phases 2–8
- **Issue:** The standard mandates one Example per public function. `example_test.go` has 39 example functions covering the primary score surface of all 23 algorithms. The following public functions have no `Example*` function:
  - Distance functions: `LevenshteinDistance`, `LevenshteinDistanceRunes`, `HammingDistance`, `HammingDistanceRunes`, `DamerauLevenshteinOSADistance`, `DamerauLevenshteinOSADistanceRunes`, `DamerauLevenshteinFullDistance`, `DamerauLevenshteinFullDistanceRunes`
  - Rune-path Score functions: `LevenshteinScoreRunes`, `HammingScoreRunes`, `JaroScoreRunes`, `JaroWinklerScoreRunes`, `DamerauLevenshteinOSAScoreRunes`, `DamerauLevenshteinFullScoreRunes`, `SmithWatermanGotohScoreRunes`, `SmithWatermanGotohScoreWithParams`, `SmithWatermanGotohRawScoreRunes`, `SmithWatermanGotohRawScoreWithParams`
  - Constructor/config: `NewSWGParams`, `NewScorer`, `DefaultScorer`, `DefaultScorerOptions`, `DefaultNormalisationOptions`, `DefaultTokeniseOptions`, `AlgoIDs`
  - Utility: `Normalise`, `Tokenise`
  - Scorer methods: `Scorer.Score` (method examples), `Scorer.Match`, `Scorer.Threshold`, `Scorer.Algorithms`, `Scorer.ScoreAll`
- **Standard:** `go-testing-standards/SKILL.md` §Benchmark Tests — "One example per algorithm + Scorer + Normalise + Tokenise." `documentation-standards/SKILL.md` §Runnable Examples — "One example per algorithm function appearing on pkg.go.dev."
- **Action:** Code fix — add Example functions for the listed public symbols. The Scorer method examples can be grouped as `ExampleScorer_Score`, `ExampleScorer_Match`, etc.
- **Rationale:** Missing Examples mean those functions have no entry on pkg.go.dev. Consumers discovering the API via the reference docs cannot see executable usage.

---

### [Important] Missing BDD feature files: `normalisation.feature`, `determinism.feature`, `tokenise.feature`

- **File:** `tests/bdd/features/` (missing files)
- **Phase introduced:** Phases 1 (tokenise), 5 (normalisation), 5 (determinism)
- **Issue:** The testing standard lists tags `@normalisation` and `@determinism` as BDD test categories, implying dedicated feature files. `tokenise.feature` is also absent — `Tokenise` is a public function with documented semantics that a consumer should be able to read as Gherkin. Note: `normalisation.feature` and `determinism.feature` are also flagged by `bdd-scenario-FINDINGS.md` (Critical Finding #1); this entry captures `tokenise.feature` as an additional missing file not in that review.
- **Standard:** `go-testing-standards/SKILL.md` §BDD Tests — "Feature files in `tests/bdd/features/` — one file per capability." Tags list includes `@normalisation`, `@determinism`.
- **Action:** Code fix — create `tokenise.feature` with scenarios covering camelCase splitting, snake_case, kebab-case, empty input, mixed-script, and the `NoEmptyTokens` invariant.
- **Rationale:** `Tokenise` is a consumer-facing function documented in `docs/requirements.md`. A consumer reading the BDD suite cannot discover its behaviour contract from Gherkin.

---

### [Important] Missing category tags (`@character`, `@qgram`, `@phonetic`, `@gestalt`) in Phase 2–5 BDD feature files

- **File:** All Phase 2–5 feature files (levenshtein, hamming, jaro, jarowinkler, damerau_osa, damerau_full, swg, strcmp95, lcsstr, ratcliff_obershelp, qgram_jaccard, sorensen_dice, cosine, tversky)
- **Phase introduced:** Phases 2–5
- **Issue:** The testing standard lists the following BDD filter tags: `@character`, `@qgram`, `@token`, `@phonetic`, `@gestalt`, `@scorer`, `@scan`, `@suppression`, `@normalisation`, `@determinism`. Phase 6–8 feature files use tags (`@token`, `@scorer`, etc.). Phase 2–5 feature files have NO tags at all — neither at the `Feature:` level nor at individual `Scenario:` level. This means `godog --tags=@character` matches zero scenarios, making the tag-based filtering system non-functional for character-tier algorithms.
- **Standard:** `go-testing-standards/SKILL.md` §BDD Tests — "Tags for filtering: `@character`, `@qgram`, ... Tag every scenario with its category."
- **Action:** Code fix — add `@character` to the Feature level in: `levenshtein.feature`, `hamming.feature`, `jaro.feature`, `jarowinkler.feature`, `damerau_osa.feature`, `damerau_full.feature`, `swg.feature`, `strcmp95.feature`, `lcsstr.feature`. Add `@gestalt` to `ratcliff_obershelp.feature`. Add `@qgram` to `qgram_jaccard.feature`, `sorensen_dice.feature`, `cosine.feature`, `tversky.feature`.
- **Rationale:** CI filtering by algorithm category (e.g. "run only phonetic BDD tests") is unusable without tags. The tag standard was established during Phase 6 but never backfilled.

---

### [Important] BDD feature files lack Unicode edge-case scenarios for Phase 2–3 character-tier algorithms

- **File:** `tests/bdd/features/levenshtein.feature`, `hamming.feature`, `jaro.feature`, `jarowinkler.feature`, `damerau_osa.feature`, `damerau_full.feature`, `swg.feature`, `strcmp95.feature`, `lcsstr.feature`
- **Phase introduced:** Phases 2–3
- **Issue:** The testing standard mandates "Unicode edge cases (multi-byte UTF-8, CJK, emoji)" as required BDD scenarios. Reviewing all Phase 2–3 feature files: none contains a Unicode scenario (café/cafe, CJK pairs, or emoji inputs). The unit tests in `*_test.go` do have UTF-8 test cases, but the BDD contract — the consumer-facing specification — is silent on Unicode behaviour. A consumer reading `levenshtein.feature` cannot determine whether `LevenshteinScore("café", "cafe")` operates at byte or rune level.
- **Standard:** `go-testing-standards/SKILL.md` §BDD Tests — "Unicode edge cases (multi-byte UTF-8, CJK, emoji)" listed as required BDD scenarios.
- **Action:** Code fix — add at least one `@unicode` scenario per character-tier feature file showing the rune vs byte distinction (e.g. `LevenshteinScoreRunes("café","cafe") = 0.75` vs `LevenshteinScore("café","cafe")` with documented byte behaviour).
- **Rationale:** Without a Unicode BDD scenario, the rune-path guarantee is invisible to consumers reading the feature files as documentation. This creates an API surprise for any consumer processing non-ASCII text.

---

### [Important] Missing fuzz seed corpus lone-surrogate and NUL-byte entries for algorithm fuzz tests

- **File:** All algorithm `*_fuzz_test.go` files except `normalise_fuzz_test.go`
- **Phase introduced:** Phases 2–7
- **Issue:** The testing standard requires fuzz harnesses to assert "Never panics on arbitrary input including invalid UTF-8, embedded NULs, lone surrogates." `normalise_fuzz_test.go` correctly seeds with `"\xed\xa0\x80"` (lone surrogate) and `"a\x00b"` (embedded NUL). No other algorithm fuzz test seeds these specific pathological byte sequences. While the fuzzer will eventually generate them, seeding them explicitly ensures they are exercised on every CI run (seed-only mode) without requiring a full fuzz session.
- **Standard:** `go-testing-standards/SKILL.md` §Fuzz Tests — "Never panics on arbitrary input including invalid UTF-8, embedded NULs, lone surrogates."
- **Action:** Code fix — add `"\xed\xa0\x80"` (lone surrogate) and `"a\x00b"` (embedded NUL) to the `f.Add(...)` seed calls in each character-tier, q-gram tier, and token-tier fuzz harness.
- **Rationale:** Lone surrogates are valid bytes that the Go fuzzer will generate but only after many iterations. Seeding them guarantees deterministic coverage of the most dangerous UTF-8 corner case on every test run.

---

### [Important] `DamerauLevenshteinFullScore_ZeroAllocs_ASCII_Short` test is permanently skipped

- **File:** `damerau_full_test.go:229`
- **Phase introduced:** Phase 2
- **Issue:** `TestDamerauLevenshteinFullScore_ZeroAllocs_ASCII_Short` uses `t.Skipf` with a note that the 0-alloc ASCII fast path is a v1.x optimisation. The skip is not gated by a build tag or flag; it skips unconditionally on every test run. This test is dead code that counts toward the test total but provides no coverage. More importantly, the skip message says "replace this Skipf with the actual AllocsPerRun assertion" — there is no tracking issue preventing this from being forgotten entirely.
- **Standard:** `go-testing-standards/SKILL.md` §Unit Tests — tests must provide coverage.
- **Action:** Discuss-phase needed — either (a) create a GitHub issue tracking the v1.x optimisation and add `//nolint:golint // skipped until #NNN: ...` with the issue reference, or (b) convert the skip to a soft assertion that logs a warning without failing if allocs > 0, preserving the test intent.
- **Rationale:** A permanently-skipped test contributes to the "pass percentage" illusion while providing no actual verification. Without a linked issue, it will never be resolved.

---

### [Important] Scorer benchmarks missing Medium/Long variants for `ScoreAll` and `Match`

- **File:** `scorer_bench_test.go`
- **Phase introduced:** Phase 8
- **Issue:** `scorer_bench_test.go` has Short, Medium, and Long variants for `Score` but only a Short variant for `ScoreAll` (`BenchmarkDefaultScorer_ScoreAll_ASCII_Short`) and only a Short variant for `Match` (`BenchmarkDefaultScorer_Match_ASCII_Short`). The standard says "short / medium / long input benchmarks (10 / 50 / 500 characters)". `ScoreAll` calls all six dispatch functions simultaneously; its behaviour at long inputs is qualitatively different from `Score` and warrants its own multi-size benchmark series. There are also no benchmarks for a custom-configured Scorer (non-default algorithm set).
- **Standard:** `go-testing-standards/SKILL.md` §Benchmark Tests — "short / medium / long input benchmarks."
- **Action:** Code fix — add `BenchmarkDefaultScorer_ScoreAll_ASCII_Medium`, `BenchmarkDefaultScorer_ScoreAll_ASCII_Long`, `BenchmarkDefaultScorer_Match_ASCII_Medium`, `BenchmarkDefaultScorer_Match_ASCII_Long`.
- **Rationale:** Without Medium/Long benchmarks for `ScoreAll`, a performance regression in a single dispatched algorithm at larger inputs would not appear in the benchstat comparison.

---

### [Important] Missing example tests for `Scorer` methods on pkg.go.dev

- **File:** `example_test.go`
- **Phase introduced:** Phase 8
- **Issue:** `Scorer` is the primary user-facing composite API. The five public methods (`Score`, `Match`, `Threshold`, `Algorithms`, `ScoreAll`) have no `Example*` functions. On pkg.go.dev, the `Scorer` type page will show method signatures without runnable examples. Consumers cannot see the idiomatic construction-and-use pattern without reading the unit tests. The standard says "One example per algorithm + Scorer + Normalise + Tokenise."
- **Standard:** `documentation-standards/SKILL.md` — runnable examples on pkg.go.dev; `go-testing-standards/SKILL.md` §Meta-tests.
- **Action:** Code fix — add `ExampleNewScorer`, `ExampleScorer_Score`, `ExampleScorer_Match`, `ExampleScorer_ScoreAll`, `ExampleDefaultScorer` to `example_test.go`.
- **Rationale:** pkg.go.dev examples are the primary discoverability surface for open-source Go libraries. The most important public type lacking examples is the highest-priority addition.

---

### [Important] Missing example tests for `Normalise` and `Tokenise`

- **File:** `example_test.go`
- **Phase introduced:** Phase 1
- **Issue:** `Normalise` and `Tokenise` are foundational utilities that affect all algorithm results. Neither has an `Example*` function in `example_test.go`. Consumers configuring `NormalisationOptions` or `TokeniseOptions` cannot see canonical usage patterns on pkg.go.dev.
- **Standard:** `go-testing-standards/SKILL.md` §Benchmark Tests — "One example per algorithm + Scorer + Normalise + Tokenise."
- **Action:** Code fix — add `ExampleNormalise` and `ExampleTokenise` showing the default options path and a customised options path.
- **Rationale:** These functions are how consumers tune the library for their input domain. Missing examples are a documentation gap for the most frequently customised surface.

---

### [Important] Missing example tests for `NewSWGParams` and `SmithWatermanGotohScoreWithParams`

- **File:** `example_test.go`
- **Phase introduced:** Phase 3
- **Issue:** `SmithWatermanGotohScoreWithParams` and `SmithWatermanGotohRawScoreWithParams` are the primary customisation entry points for SWG. `NewSWGParams` is the constructor for the parameter struct. None of these three functions has an `Example*`. The existing `ExampleSmithWatermanGotohScore` shows the default-params path only.
- **Standard:** `documentation-standards/SKILL.md` §Runnable Examples.
- **Action:** Code fix — add `ExampleNewSWGParams` showing custom match/mismatch/gap parameters, and `ExampleSmithWatermanGotohScoreWithParams` showing the result on a pair that diverges from default params.
- **Rationale:** Custom params are the main reason a consumer would choose SWG over a simpler algorithm. Missing examples make the customisation interface invisible.

---

### [Important] Missing example tests for Distance functions (`LevenshteinDistance`, `HammingDistance`)

- **File:** `example_test.go`
- **Phase introduced:** Phase 2
- **Issue:** `LevenshteinDistance` and `HammingDistance` are public functions returning raw integer distances (not normalised scores). They have no `Example*` function. Consumers who need the raw edit distance rather than the normalised score have no documented usage pattern. `DamerauLevenshteinOSADistance` and `DamerauLevenshteinFullDistance` are similarly unexampled.
- **Standard:** `documentation-standards/SKILL.md` §Runnable Examples.
- **Action:** Code fix — add `ExampleLevenshteinDistance`, `ExampleHammingDistance` at minimum. DL variants can be added as part of the same PR.
- **Rationale:** The Distance variants are used differently from the Score variants (for threshold-by-count rather than threshold-by-ratio decisions). The absence of examples makes this usage invisible.

---

### [Important] Missing example tests for `LongestCommonSubstringRunes` and rune-path Score variants

- **File:** `example_test.go`
- **Phase introduced:** Phases 2–6
- **Issue:** The pattern of exposing both byte and rune-path variants is central to the library's Unicode story. While some rune variants have examples (`ExampleLCSStrScoreRunes`, `ExampleQGramJaccardScoreRunes`), many are missing: `LevenshteinScoreRunes`, `HammingScoreRunes`, `JaroScoreRunes`, `JaroWinklerScoreRunes`, `DamerauLevenshteinOSAScoreRunes`, `DamerauLevenshteinFullScoreRunes`. These are the functions a consumer would call for correct multi-byte UTF-8 handling.
- **Standard:** `documentation-standards/SKILL.md` §Runnable Examples.
- **Action:** Code fix — add `ExampleLevenshteinScoreRunes`, `ExampleHammingScoreRunes` at minimum, each showing a multi-byte UTF-8 pair (e.g. café/cafe) with the rune-aware result and a note explaining the divergence from the byte path.
- **Rationale:** Without examples showing the byte vs rune divergence on a concrete multi-byte input, consumers processing non-ASCII text will default to the byte-path function and get incorrect results without realising it.

---

### [Important] Scorer `ScoreAll` deterministic iteration order is untested at the property level

- **File:** `scorer_test.go`
- **Phase introduced:** Phase 8
- **Issue:** `ScoreAll` returns `map[AlgoID]float64`. The determinism standard (`determinism-standards/SKILL.md`) mandates "Extract keys, sort them, iterate the sorted slice" on all output paths. The `ScoreAll` method returns a map (documented as having non-deterministic iteration order). There is no property test asserting that two consecutive calls to `s.ScoreAll(a, b)` return maps with bit-identical values for all keys. The concurrent test (`TestScorer_ConcurrentSafety`) does a structural comparison but not a property test with randomised inputs.
- **Standard:** `determinism-standards/SKILL.md` — determinism guarantee; `go-testing-standards/SKILL.md` §Property Tests.
- **Action:** Code fix — add `TestProp_Scorer_ScoreAll_DeterministicAcrossRuns` using `testing/quick` to assert all map values are bit-identical on repeated calls.
- **Rationale:** If a future refactor introduces any non-deterministic element to the dispatch layer (e.g. a sorted-keys optimisation with a bug), `ScoreAll` would be the first surface where it appears.

---

### [Important] `ratcliff_obershelp` cross-validation corpus lacks rune-path entries

- **File:** `testdata/cross-validation/ratcliff-obershelp/vectors.json`
- **Phase introduced:** Phase 4
- **Issue:** The 16-entry Ratcliff-Obershelp cross-validation corpus contains only ASCII pairs. `RatcliffObershelpScoreRunes` operates on multi-byte UTF-8 inputs with a distinct code path. There are no corpus entries using non-ASCII inputs (e.g. café/cafe, CJK pairs). The rune-path unit tests have hand-computed reference vectors but no external cross-validation. `difflib.SequenceMatcher` in Python 3 operates at the Unicode character level, making it the natural reference for the rune path.
- **Standard:** `algorithm-correctness-standards/SKILL.md` §Reference Vectors — "For Unicode-aware algorithms: at least one non-ASCII case."
- **Action:** Code fix — add at least three non-ASCII pair entries to `testdata/cross-validation/ratcliff-obershelp/vectors.json` using `difflib.SequenceMatcher(autojunk=False)` computed on the Unicode character sequence.
- **Rationale:** The rune path's correctness against the difflib reference is currently verified only for ASCII. A rune-boundary handling bug would not be caught by the existing corpus.

---

### [Important] `DamerauLevenshteinFullScore` `t.Skip` test is untracked

- **File:** `damerau_full_test.go:229`
- **Phase introduced:** Phase 2
- **Issue:** Already noted above as an Important finding. The skip text says "see plan 02-06 SUMMARY" but there is no GitHub issue number in the skip message and no TODO comment in the standard format `// TODO(#N): ...`. This means the v1.x optimisation is untraceable via `gh issue list`.
- **Standard:** `CLAUDE.md` — "Every `TODO` must reference a GitHub issue: `// TODO(#42): ...`"
- **Action:** Code fix — create a GitHub issue for the v1.x two-row DL-Full optimisation and add the issue number to the `t.Skipf` message and a `// TODO(#N): ...` comment above the test.
- **Rationale:** Without a linked issue, the skipped test will never be revisited.

---

## Improvement Findings

### [Improvement] `testdata/fuzz/` subdirectories have only one seed file each

- **File:** `testdata/fuzz/*/seed-001`
- **Phase introduced:** Phases 2–5
- **Issue:** Every on-disk fuzz corpus directory contains exactly one seed file (`seed-001`). The normalise and tokenise fuzz tests have 5 seeds each (reflecting their more complex input space). The algorithm fuzz tests would benefit from 3–5 seeds covering: both-empty, one-empty, identical ASCII, invalid UTF-8 (high byte), and multi-byte UTF-8 (café/cafe). More seeds mean better deterministic coverage on every CI run (seed-only mode).
- **Standard:** `go-testing-standards/SKILL.md` §Fuzz Tests — "Seed corpus."
- **Action:** Code fix — add seed-002 through seed-005 to each algorithm's corpus directory, covering the edge-case classes listed above.
- **Rationale:** One seed file exercises one input class. Five seeds ensure all five edge-case classes are covered on every CI seed-only run, not just the full fuzz session.

---

### [Improvement] Property test for `RatcliffObershelpScore` asymmetry is in `cross_algorithm_consistency_test.go` but not in `props_test.go`

- **File:** `cross_algorithm_consistency_test.go:380`; missing in `props_test.go`
- **Phase introduced:** Phase 4
- **Issue:** The asymmetry pin for `RatcliffObershelpScore` (`TestCrossAlgorithm_RatcliffObershelp_AsymmetryPin`) lives in `cross_algorithm_consistency_test.go` alongside the cross-algorithm tests. Per the project convention, algorithm-specific invariants belong in `props_test.go` (or the algorithm's `_test.go`) and cross-algorithm relationship tests belong in `cross_algorithm_consistency_test.go`. The asymmetry pin is an algorithm-specific property, not a cross-algorithm relationship. Having it in the cross-algorithm file makes it harder to discover when reading the algorithm's own test file.
- **Standard:** `go-testing-standards/SKILL.md` §Unit Tests — file placement conventions.
- **Action:** Code fix — move `TestCrossAlgorithm_RatcliffObershelp_AsymmetryPin` to `props_test.go` and rename it `TestProp_RatcliffObershelpScore_Asymmetric`. Keep the cross-algorithm `TestCrossAlgorithm_RatcliffObershelp_PinnedAgainstDifflib` where it is (that one tests the difflib equivalence relationship).
- **Rationale:** Test placement conventions exist so reviewers know where to look for coverage of a given function. An asymmetry property buried in the cross-algorithm file breaks the convention.

---

### [Improvement] `props_test.go` mixes per-algorithm tests that should be in algorithm-specific test files

- **File:** `props_test.go`
- **Phase introduced:** Phases 2–8 (accumulated)
- **Issue:** `props_test.go` is 3700+ lines and contains property tests for all 23 algorithms. Some property tests are algorithm-specific (e.g. `TestProp_SmithWatermanGotoh_GapSplitInvariance`, `TestProp_MongeElkanScore_AsymmetricWhenTokenCountAsymmetric`) that belong logically alongside the unit tests in the algorithm's own `_test.go` file or a dedicated `swg_props_test.go`. The current monolith makes it difficult to navigate and increases the risk that a new algorithm's properties are added in the wrong place or omitted.
- **Standard:** `go-testing-standards/SKILL.md` §Unit Tests — "File placement: `levenshtein_test.go` beside `levenshtein.go`."
- **Action:** Improvement (no behaviour change) — consider splitting algorithm-specific property tests into per-algorithm `<algo>_props_test.go` files (or appending them to the existing `<algo>_test.go`). The common cross-cutting properties (range bounds, identity, symmetry) can remain in `props_test.go` as a shared property harness.
- **Rationale:** At Phase 9 with the scan package adding more property tests, `props_test.go` will continue to grow. Splitting now prevents the file from becoming unmanageable.

---

### [Improvement] No `bench.txt` baseline comparison for Phase 6–8 algorithms

- **File:** `bench.txt`
- **Phase introduced:** Phases 6–8
- **Issue:** `bench.txt` exists but was last committed before Phase 6 algorithms were added (TokenSortRatio, TokenSetRatio, PartialRatio, TokenJaccard, MongeElkan — Phase 6; Soundex, DoubleMetaphone, NYSIIS, MRA — Phase 7; Scorer — Phase 8). The benchstat comparison `make bench-compare` will find no baseline entries for new algorithms, making regression detection ineffective for them.
- **Standard:** `go-testing-standards/SKILL.md` §Benchmark Tests — "`bench.txt` committed per release; CI runs benchstat against the last tagged release."
- **Action:** Code fix — regenerate `bench.txt` to include all current algorithms and commit the updated baseline.
- **Rationale:** benchstat cannot detect a regression against a missing baseline. Every algorithm added after the last `bench.txt` update is invisible to the regression detector.

---

### [Improvement] `testdata/golden/_staging/` files are undocumented as a workflow artefact

- **File:** `testdata/golden/_staging/` (no README or comment)
- **Phase introduced:** Phase 2
- **Issue:** The `_staging/` directory contains 23 per-algorithm golden JSON files that feed `TestGolden_Algorithms_Merge`. There is no documentation (README, inline comment, or CONTRIBUTING note) explaining the lifecycle: staging files are created by individual algorithm tests, promoted by `TestGolden_Algorithms_Merge` into `algorithms.json`, and then the staging file remains. A new contributor adding an algorithm does not know (a) that they must create a staging file, (b) how the file format is determined, or (c) what triggers promotion. The 9 algorithms with staging files but no staging test (Phase 6–7 gap noted in Critical findings) would be invisible to such a contributor.
- **Standard:** `documentation-standards/SKILL.md` — documentation for all non-obvious workflows.
- **Action:** Improvement — add a `testdata/golden/README.md` (or a comment in `golden_test.go`) explaining the staging lifecycle. Also add a meta-test that asserts every file in `_staging/` has a corresponding `TestGolden_*_Staging` test function (parseable via `go/parser`).
- **Rationale:** An undocumented staging workflow will be skipped by contributors who discover the gap too late (after their algorithm is merged without a staging test).

---

### [Improvement] `cross_algorithm_consistency_test.go` is 823 lines and growing; a Phase 9 scan-readiness hook is missing

- **File:** `cross_algorithm_consistency_test.go`
- **Phase introduced:** Phase 6
- **Issue:** The cross-algorithm consistency file adds new test functions with each phase. At Phase 9 (scan), the scan sub-package will introduce `scan.Check` which calls `Scorer.Score` across item pairs. The cross-algorithm layer should include at least one placeholder structure (a stub function with a `t.Skip`) for the scan sub-package's cross-algorithm property: `PropCheck_BucketEquivalentToNaive`. Without this placeholder, Phase 9 will have to add a new test file to the root package for the cross-algorithm scan test, creating inconsistency.
- **Standard:** `go-testing-standards/SKILL.md` §Scan-level properties — "`scan.Check` output sorted by `(Kind, NameA, NameB, GroupA, GroupB)`; `Bucket-equivalent-to-naive`: property test proves equivalence to naive O(N²) pairwise scoring."
- **Action:** Improvement — add a stub `TestCrossAlgorithm_Scan_BucketEquivalentToNaive` with `t.Skip("Phase 9")` in `cross_algorithm_consistency_test.go` or create `scan/scan_test.go` in the (not-yet-created) `scan/` package. The Phase 9 team should see the placeholder and know the pattern.
- **Rationale:** Establishing the test structure before the code exists ensures Phase 9 does not have to reverse-engineer the property test pattern from scratch.

---

### [Improvement] Fuzz harness naming for phonetic functions is inconsistent with algorithm naming

- **File:** `soundex_fuzz_test.go`, `nysiis_fuzz_test.go`, `mra_fuzz_test.go`, `double_metaphone_fuzz_test.go`
- **Phase introduced:** Phase 7
- **Issue:** The fuzz harness names for phonetic algorithms do not follow the pattern of the primary score function:
  - `FuzzSoundex` (should be `FuzzSoundexCode` or `FuzzSoundexScore` to match the primary function name)
  - `FuzzNYSIIS` (should be `FuzzNYSIISCode` or `FuzzNYSIISScore`)
  - `FuzzMRA` (should be `FuzzMRACode` or `FuzzMRAScore`)
  - `FuzzDoubleMetaphone` (matches primary function `DoubleMetaphoneKeys` less directly)
  Character-tier and q-gram-tier harnesses are named after the primary public function (e.g. `FuzzLevenshteinScore`, `FuzzQGramJaccardScore`). Phonetic harnesses use abbreviated algorithm names instead.
- **Standard:** `go-testing-standards/SKILL.md` — "Every public function has a fuzz harness in `fuzz_test.go`." The naming convention implies the harness name matches the function name.
- **Action:** Improvement — rename `FuzzSoundex` → `FuzzSoundexCode` (since it covers both `SoundexCode` and `SoundexScore`), `FuzzNYSIIS` → `FuzzNYSIISCode`, `FuzzMRA` → `FuzzMRACode`. Update corpus directory names accordingly (`testdata/fuzz/FuzzSoundexCode/`). This is a rename only; no functional change.
- **Rationale:** Consistent naming allows `grep "FuzzLevenshteinScore\|FuzzSoundexScore"` to work predictably and makes the harness-to-function mapping obvious.

---

### [Improvement] No goroutine-leak detector (`goleak`) in the root package `TestMain`

- **File:** missing: root package `TestMain`
- **Phase introduced:** Phase 1 (should have been considered at bootstrap)
- **Issue:** `tests/bdd/bdd_test.go` uses `goleak.VerifyTestMain(m)`. The root test package has no `TestMain` function and therefore no goroutine-leak detection. Since the library is pure-function and imports no packages that launch goroutines, the risk is low. However, if any future contributor accidentally introduces a goroutine leak (e.g. in a background table-initialisation, a leaked timer in an allocation budget test, or via a transitive dependency), it will not be caught in the root package test suite.
- **Standard:** `go-testing-standards/SKILL.md` §Goroutine Leak Detection — "The library is pure-function; the test catches accidental introduction of background work." The standard only mandates goleak in `tests/bdd/`; this is a defence-in-depth improvement for the root.
- **Action:** Improvement — add a `main_test.go` with `TestMain(m *testing.M)` using `goleak.VerifyTestMain(m)` to the root package test suite. Note: goleak is a test-only dependency; it must remain in `tests/bdd/go.mod` only and cannot be added to the root `go.mod`. This means the root `TestMain` must either implement a lightweight goroutine snapshot-and-compare or accept the limitation.
- **Rationale:** Since goleak cannot be in the root `go.mod`, this is genuinely an improvement rather than a required fix. The correct approach is either to implement goroutine detection using `runtime.NumGoroutine()` snapshots (stdlib) or document explicitly that leak detection is BDD-only.

---

## Phase 9 Scan-Readiness Architectural Notes

The following are rough edges in the current test architecture that should be resolved before Phase 9 (scan sub-package) to avoid architectural debt:

1. **`scorer_options_test.go` naming** (Critical Finding #9): Rename to `scorer_options_internal_test.go` before Phase 9 creates scan-package option tests that need to follow the same pattern.

2. **Missing `internal_coverage_test.go`** (Critical Finding #11): Create before Phase 9 so the coverage floor is enforced for both the root package and the scan sub-package from day one.

3. **Missing scan placeholder in `cross_algorithm_consistency_test.go`** (Improvement #5): Add before Phase 9 so the scan team has a clear pattern to follow.

4. **BDD `scan.feature` and `suppression.feature`** (Critical in `bdd-scenario-FINDINGS.md`): These are Phase 9 blockers. The BDD infrastructure must be ready to receive scan scenarios before scan implementation begins.

5. **`goleak` in root TestMain** (Improvement #8): The scan sub-package will have its own `tests/bdd/` but may also add root-level integration tests. Decide the goroutine-detection strategy before Phase 9 introduces goroutine patterns (the scan package itself is pure-function, but test helpers may use background goroutines for parallel corpus scanning).

</details>

---

## 16. user-guide-reviewer

_Source: `.planning/reviews/user-guide-FINDINGS.md`_

<details>
<summary>Click to expand full report</summary>

---
status: issues_found
agent: user-guide-reviewer
scope: consumer-facing docs (phases 1-8)
reviewed: 2026-05-17T06:36:00Z
finding_counts:
  critical: 7
  important: 22
  improvement: 18
  total: 47
---

## Summary

The library has shipped Phase 8 (Scorer + DefaultScorer + 23 algorithms + 39 godoc examples). The consumer-facing documentation has NOT kept pace.

A new user landing on the README today sees a "Quick Start" framed around Phase 1 primitives (Normalise / Tokenise) with no mention that 23 algorithms and a working `DefaultScorer()` are already available. The single Quick Start program in README is also factually wrong: `Tokenise("XMLHttpRequest", ...)` actually returns `[xml http request]`, not the documented `[xmlhttp request]`. The README never shows a `go get` command, never shows `LevenshteinScore`, never shows `DefaultScorer().Score(...)`, never points at the three runnable `examples/` directories, and never demonstrates the error-handling sentinels.

`docs/algorithms.md` is the most visible regression: 100% scaffold with `TBD` primary sources for every one of the 23 algorithms. README's catalogue table deep-links into this file. Every link goes to a stub.

`docs/scorer.md` is the strongest piece of new documentation but contains: (a) an internal contradiction ("All four methods" describing a five-method table), (b) an aspirational claim that `ErrInvalidThreshold` rejects NaN that does not match the code (CR-01 — confirmed against `scorer_options.go:257-266`), (c) inconsistent error-handling across its code samples, (d) `sort.Slice` and `fmt.Printf` in examples with no imports shown, (e) no warning that the "default minus algorithm" pattern relies on options last-write-wins for the threshold, and (f) no callout that `WithoutNormalisation` does NOT actually pass raw bytes to token-based algorithms (Tokenise still lowercases internally).

`docs/tuning.md` is the most polished consumer doc but does not link to any concrete calibration script or example — the "100-500 pair calibration loop" has no starter code.

`docs/scan.md` and `docs/extending.md` are scaffolds with TBD bodies — fine for pre-Phase-9, but the README points at them as if they were ready.

The `examples/` directory has three carefully-crafted runnable programs that are essentially invisible: the README does not mention `examples/` even once. `docs/scorer.md` is the only consumer-facing doc that links to them.

TTHW measurement:
- README Quick Start: 6 lines of actual code (including 2 imports and 2 var statements), prints `[xml http request]` — but README claims `[xmlhttp request]`. **First contact with the library yields a contradiction.**
- No `go get` command anywhere in any consumer-facing doc.
- TTHW for `LevenshteinScore("kitten", "sitting")`: not present in any consumer-facing prose. Only in `example_test.go` (which the README never links to as the canonical entry point).
- TTHW for `DefaultScorer().Score("user_id", "userId")`: present in `docs/scorer.md` Quickstart only.
- Headline example error handling: README example uses `Normalise` (which cannot fail). `DefaultScorer().Score(...)` would be a cleaner zero-error headline; not shown in README.

Severity tags below are organisational only. All findings should be addressed before v1.0.

---

## Critical

### [Critical] README Quick Start example output is wrong
- **File:** /Users/johnny/Development/fuzzymatch/README.md (lines 107-117)
- **Phase introduced:** Phase 1
- **Issue:** The Quick Start prints `// Output: [xmlhttp request]` for `Tokenise("XMLHttpRequest", DefaultTokeniseOptions())`. The actual output is `[xml http request]` (verified by running the program). `DefaultTokeniseOptions().SplitConsecutiveUpper = true` splits "XMLHttp" into "xml" + "http" before splitting on the case transition into "Request".
- **Standard:** documentation-standards.md "Quick Start: copy-paste and it compiles and runs"; "Every code example in documentation must compile (verified by `documentation_test.go`)"
- **Action:** Code fix (the doc) — either update the comment to `[xml http request]` or change the input to something that actually produces a meaningful pair like `[xmlhttp request]` (e.g. by tweaking options or using a different camelCase input).
- **Rationale:** First contact with the library produces a "the docs lied to me" moment for any user who copy-pastes and runs.
- **Suggested fix:** Pin the example to `// Output: [xml http request]` and add a meta-test that runs the README's first code block (the standards skill calls this `readme_shop_front_test.go`).

### [Critical] docs/scorer.md claims `ErrInvalidThreshold` rejects NaN; code does not (CR-01)
- **File:** /Users/johnny/Development/fuzzymatch/docs/scorer.md (line 283); /Users/johnny/Development/fuzzymatch/scorer_options.go (lines 257-266)
- **Phase introduced:** Phase 8
- **Issue:** Documentation row reads `ErrInvalidThreshold | Returned when WithThreshold receives a value outside [0.0, 1.0], or a NaN.` The implementation only checks `t < 0.0 || t > 1.0`, both of which are `false` for `math.NaN()`. So `WithThreshold(math.NaN())` succeeds, the Scorer freezes with `threshold = NaN`, and `Match(a, b)` (which evaluates `score >= NaN`, always false) silently never matches anything. This is the exact CR-01 finding from `.planning/phases/08-composite-scorer/08-REVIEW.md`.
- **Standard:** "Does the documentation match the actual code behaviour?" — user-guide-reviewer key question.
- **Action:** Code fix (fix the code per CR-01: add `math.IsNaN(t)` check). Documentation is correct and should remain.
- **Rationale:** Silent no-match malfunctions are the worst class of consumer bug — the Scorer looks like it's "just very strict" until the consumer realises nothing is ever matching.

### [Critical] docs/algorithms.md is 100% scaffold with TBD primary sources for all 23 algorithms after Phases 2-7 shipped
- **File:** /Users/johnny/Development/fuzzymatch/docs/algorithms.md (every algorithm section, lines 25-238)
- **Phase introduced:** Phase 1 (scaffold); should have been filled in over Phases 2-7
- **Issue:** Every algorithm entry says `Primary source: TBD — filled in by the implementing phase`. The README's catalogue table deep-links into this file (`docs/algorithms.md#levenshtein`, `docs/algorithms.md#damerau-levenshtein-osa`, etc.). All 23 links resolve to a 5-line stub. Meanwhile the actual primary source is cited in inline Go file comments and in `llms.txt` (where it is up to date) — so a sophisticated reader can find the citation, but the consumer-targeted doc has been overlooked.
- **Standard:** documentation-standards.md "Algorithm Documentation" mandates per-algorithm: name, category, primary academic source citation, description, mathematical formulation, complexity, score normalisation rule, mathematical invariants, edge cases, reference vectors with citation, intended use cases, comparable references.
- **Action:** Code fix (the docs). This is a major write-up but the source content already exists in inline Go godocs.
- **Rationale:** Without this, the README catalogue table is essentially a teaser — clicking through gives no information beyond the AlgoID enum value. New users can't make algorithm-selection decisions.

### [Critical] README Quick Start does not show the v0.x→v1.0 headline use case (algorithm score or DefaultScorer)
- **File:** /Users/johnny/Development/fuzzymatch/README.md (lines 94-117)
- **Phase introduced:** Phase 1 (stale); should have been updated at every phase boundary
- **Issue:** The README "Quick Start" still says `Phase 1 (foundation) ships Normalise and Tokenise primitives... Algorithm functions (e.g. LevenshteinScore) land in Phase 2. The example below uses the Phase-1 primitives`. Phase 8 has shipped. The example uses Normalise and Tokenise — neither of which is the headline use case. A new user sees lowercase-and-tokenise as the library's purpose, not similarity scoring.
- **Standard:** documentation-standards.md "First code block shows a complete, working program (the 'headline example')"; user-guide-reviewer "Can a developer compute a similarity score in under 2 minutes?"
- **Action:** Code fix (the docs). Replace Quick Start with a `DefaultScorer().Score(...)` example, OR a `LevenshteinScore(...)` example, ideally both.
- **Rationale:** The headline example IS the library's first impression. Today it's misleading.
- **Suggested fix:**
  ```go
  package main

  import (
      "fmt"
      "github.com/axonops/fuzzymatch"
  )

  func main() {
      score := fuzzymatch.LevenshteinScore("kitten", "sitting")
      fmt.Printf("%.4f\n", score) // 0.5714

      // Or use the opinionated DefaultScorer:
      s := fuzzymatch.DefaultScorer()
      fmt.Println(s.Match("user_id", "userId")) // true
  }
  ```

### [Critical] No `go get github.com/axonops/fuzzymatch` anywhere in consumer-facing docs
- **File:** /Users/johnny/Development/fuzzymatch/README.md; /Users/johnny/Development/fuzzymatch/docs/scorer.md; /Users/johnny/Development/fuzzymatch/llms.txt
- **Phase introduced:** Phase 1
- **Issue:** README never shows the install command. Bare imports of `github.com/axonops/fuzzymatch` appear in examples, but no `go get` invocation. The standard "install line" — the single most important sentence in any Go library README — is missing.
- **Standard:** documentation-standards.md "Quick Start: copy-paste and it compiles and runs". Implicitly: install + use.
- **Action:** Code fix (the docs).
- **Rationale:** TTHW failure. Step zero of "I found this library" is missing.

### [Critical] README never mentions the `examples/` directory
- **File:** /Users/johnny/Development/fuzzymatch/README.md
- **Phase introduced:** Phase 7 (when `examples/` was first introduced) or earlier
- **Issue:** Three carefully-built runnable programs live under `examples/`:
  - `examples/identifier-similarity/` (23-algorithm side-by-side + Default Scorer Score/Match columns)
  - `examples/phonetic-keys/` (SoundexCode, DoubleMetaphoneKeys, NYSIISCode, MRACode, MRACompare)
  - `examples/scorer-composition/` (DefaultScorer vs Default-minus-DoubleMetaphone)
  Each has a `main_test.go` pinning byte-stable expected output. None of these are linked from README. Only `docs/scorer.md` lines 336-339 mentions two of them. A new user has no way to discover that there are runnable demos.
- **Standard:** documentation-standards.md "Include runnable Example functions in `example_test.go` for major use cases"; the spirit applies to `examples/` directory programs too.
- **Action:** Code fix (the docs).
- **Rationale:** The runnable examples are the fastest path to confidence — and they're invisible.

### [Critical] README "Layer 2" illustrative code is syntactically misleading
- **File:** /Users/johnny/Development/fuzzymatch/README.md (line 82)
- **Phase introduced:** Phase 1
- **Issue:** The three-layers diagram shows `Layer 2: Scorer    NewScorer().Score(a, b)`. The actual API is `NewScorer(opts ...ScorerOption) (*Scorer, error)` — it MUST take options (at minimum `WithThreshold`) and returns an error. The diagram suggests a zero-argument call that yields a `*Scorer` directly. This is a documentation error a careful reader will trip over.
- **Standard:** "API ergonomics — the surface shape is what's shown"
- **Action:** Code fix (the docs).
- **Suggested fix:** `Layer 2: Scorer    DefaultScorer().Score(a, b)` (this composes correctly).

---

## Important

### [Important] No runnable godoc example for the Scorer
- **File:** /Users/johnny/Development/fuzzymatch/example_test.go (39 example functions but none for Scorer/NewScorer/DefaultScorer)
- **Phase introduced:** Phase 8
- **Issue:** `example_test.go` has 39 `Example*` functions for individual algorithms (verified by `grep '^func Example'`), but ZERO for any of: `Scorer`, `NewScorer`, `DefaultScorer`, `DefaultScorerOptions`, `WithAlgorithm`, `WithThreshold`, `Scorer.Score`, `Scorer.Match`, `Scorer.ScoreAll`, `Scorer.Algorithms`, `Scorer.Threshold`. pkg.go.dev will not show a runnable, output-verified example for the central Scorer API.
- **Standard:** documentation-standards.md "Include runnable `Example` functions in `example_test.go` for major use cases".
- **Action:** Code fix.
- **Rationale:** The runnable example is the canonical "this is exactly what you type" surface. Its absence forces consumers to learn the Scorer API from prose only.

### [Important] No runnable godoc examples for Normalise or Tokenise
- **File:** /Users/johnny/Development/fuzzymatch/example_test.go; /Users/johnny/Development/fuzzymatch/normalise_test.go; /Users/johnny/Development/fuzzymatch/tokenise_test.go
- **Phase introduced:** Phase 1
- **Issue:** No `ExampleNormalise` or `ExampleTokenise` anywhere in the repository. The README Quick Start uses both, but pkg.go.dev shows no runnable example for either function. Compare against the 39 algorithm examples.
- **Standard:** documentation-standards.md "runnable examples for every algorithm category" — Normalise/Tokenise are not algorithms but they are core to the public API.
- **Action:** Code fix.

### [Important] No documented `errors.Is` usage example
- **File:** /Users/johnny/Development/fuzzymatch/example_test.go; /Users/johnny/Development/fuzzymatch/docs/scorer.md
- **Phase introduced:** Phase 8 (the Scorer error sentinels are the first place a consumer encounters errors)
- **Issue:** `docs/scorer.md` line 275 says "discriminate via `errors.Is`, never by matching the error message string" — but no example of how. The doc lists 4 sentinels in a table; the consumer is left to know to write `if errors.Is(err, fuzzymatch.ErrMissingThreshold) { ... }`. `errors.go` godoc has the same prose but no example. A runnable godoc example would close the gap.
- **Standard:** user-guide-reviewer key question: "Are error sentinels documented with `errors.Is` usage example?"
- **Action:** Code fix.
- **Suggested fix:** Add `ExampleNewScorer_errors` in `example_test.go` showing the `errors.Is` discrimination pattern for at least `ErrMissingThreshold` and `ErrInvalidThreshold`.

### [Important] docs/scorer.md "All four methods" contradicts its five-row method-reference table (IN-01)
- **File:** /Users/johnny/Development/fuzzymatch/docs/scorer.md (lines 95-107, 258-259)
- **Phase introduced:** Phase 8
- **Issue:** Table at lines 97-103 lists FIVE methods: Score, Match, ScoreAll, Threshold, Algorithms. Immediately following text at line 105: "All four methods are pure functions". Repeated at line 258-259: "All four methods (`Score`, `Match`, `ScoreAll`, `Threshold`, `Algorithms`)" — calls five methods "four". Reader confidence damaged on first encounter.
- **Standard:** Documentation accuracy / self-consistency.
- **Action:** Code fix (the docs).
- **Suggested fix:** Change "four" → "five" both places, or restructure as "All Scorer methods".

### [Important] docs/scorer.md error-handling code samples are inconsistent
- **File:** /Users/johnny/Development/fuzzymatch/docs/scorer.md
- **Phase introduced:** Phase 8
- **Issue:** Quickstart at line 27-32 calls `DefaultScorer()` directly with no error handling (correct — it cannot fail). Custom example at lines 43-51 shows `if err != nil { return fmt.Errorf(...) }` (with imports/context unstated). Lines 67-72 (`opts := append(...)`) drops the `err` check entirely. Lines 87-92 (parameterised options) drops the `err` check. Lines 152-156 drops the `err` check with bare `_`. Inconsistent across the same document — readers can't tell what the recommended idiom is.
- **Standard:** documentation-standards.md "Every code example in documentation must compile"; consumer ergonomics.
- **Action:** Code fix (the docs). Standardise on the pattern from lines 43-51 throughout (the named `s, err := …` plus an `if err != nil` block), OR explicitly comment on each spot why ignoring the error is safe here.

### [Important] Code examples in docs/scorer.md use unimported packages
- **File:** /Users/johnny/Development/fuzzymatch/docs/scorer.md (lines 227-237)
- **Phase introduced:** Phase 8
- **Issue:** The ScoreAll example uses `sort.Slice` and `fmt.Printf` without `import "sort"` or `import "fmt"`. The reader has to infer that "of course you need those imports" — but the scorer.md Quickstart at lines 22-24 DOES show the import; the pattern is inconsistent. The CR's "every code example must compile" standard requires either the example shows imports or the snippet is unambiguously incomplete (e.g. inside a `// ...` block).
- **Standard:** documentation-standards.md "Every code example in documentation must compile (verified by `documentation_test.go`)"
- **Action:** Code fix (the docs).

### [Important] `WithoutNormalisation` semantic surprise not flagged
- **File:** /Users/johnny/Development/fuzzymatch/docs/scorer.md (lines 168-180)
- **Phase introduced:** Phase 8
- **Issue:** "With `WithoutNormalisation`, the raw input bytes are passed to every algorithm" — but token-based algorithms (Monge-Elkan, Token*, PartialRatio) internally call `Tokenise(s, DefaultTokeniseOptions())`, and `DefaultTokeniseOptions().Lowercase = true` (verified in `tokenise.go:107`). So `WithoutNormalisation` does NOT mean "no character changes" — it means "no Unicode normalisation pipeline, but tokens are still lowercased on the way to set/sort algorithms". This is a footgun. A consumer comparing case-sensitive identifiers will get unexpected lowercase-after-tokenise behaviour from any token algorithm.
- **Standard:** user-guide-reviewer key question: "Are common pitfalls called out (ScoreAll iteration non-determinism, WithoutNormalisation surprises, threshold tuning)?"
- **Action:** Code fix (the docs).
- **Suggested fix:** Add a callout block: "Note: token-based algorithms still call `Tokenise(DefaultTokeniseOptions())` internally regardless of `WithoutNormalisation`. The default `Tokenise` lowercases. To get truly byte-identical input through every algorithm, either pre-tokenise upstream or avoid the token-based algorithms in the Scorer composition."

### [Important] `WithTverskyAlgorithm` constraint `α + β > 0` is not documented in scorer.md or tuning.md (CR-02 surface)
- **File:** /Users/johnny/Development/fuzzymatch/docs/scorer.md (line 89 example uses α=β=0.5); /Users/johnny/Development/fuzzymatch/docs/tuning.md; /Users/johnny/Development/fuzzymatch/scorer_options.go (line 381-399)
- **Phase introduced:** Phase 8
- **Issue:** `docs/scorer.md` line 89 shows `WithTverskyAlgorithm(0.3, 0.5, 0.5, 3)` as the example. The α + β > 0 constraint is enforced at TverskyScore call time (panic in `tversky.go:241`) but the option layer accepts `(0, 0)`. A consumer who tries `WithTverskyAlgorithm(0.3, 0.0, 0.0, 3)` (perhaps to disable both asymmetric terms while keeping intersection-only) will see no error at NewScorer time and a panic on the first Score call. The docs nowhere warn that "α and β must both be ≥ 0 AND at least one must be > 0". This is CR-02 from Phase 8 review.
- **Standard:** "fail loudly at construction" — Phase 8 contract. Documentation should match.
- **Action:** Code fix (the code per CR-02). Then update docs to document the constraint.

### [Important] Phonetic encoder API surface (SoundexCode, DoubleMetaphoneKeys, NYSIISCode, MRACode, MRACompare) is absent from consumer docs
- **File:** /Users/johnny/Development/fuzzymatch/docs/algorithms.md; /Users/johnny/Development/fuzzymatch/docs/extending.md; /Users/johnny/Development/fuzzymatch/docs/scorer.md
- **Phase introduced:** Phase 7
- **Issue:** The phonetic algorithms expose both a binary score (`SoundexScore`, etc.) AND a raw encoder (`SoundexCode`, `DoubleMetaphoneKeys`, `NYSIISCode`, `MRACode`) plus `MRACompare(a, b) (bool, int)`. The encoder is the right tool for "compose phonetic codes with edit distance" patterns (documented in faq.md "Why phonetic-as-binary in the Scorer?" and gestured-at in docs/extending.md "TBD" stub). But the raw encoder API is documented ONLY in `docs/requirements.md` §7.20 (an internal spec) and in `examples/phonetic-keys/`. The README catalogue doesn't mention encoders exist. A consumer reading `docs/algorithms.md#soundex` (a TBD stub) has no idea they can call `SoundexCode("Robert") → "R163"`.
- **Standard:** documentation-standards.md "Each algorithm must be documented in `docs/algorithms.md` with... intended use cases"; user-guide-reviewer "Can they extend the library?"
- **Action:** Code fix (the docs).

### [Important] README Algorithm Catalogue links all resolve to TBD scaffolds
- **File:** /Users/johnny/Development/fuzzymatch/README.md (lines 127-171)
- **Phase introduced:** Phase 1 (links); Phase 2-7 (should have filled in)
- **Issue:** Every catalogue table entry deep-links into `docs/algorithms.md#<algo>`. Every target is a 5-line TBD stub. The links are functionally useless for learning what an algorithm does — they only confirm the algorithm exists.
- **Standard:** documentation-standards.md "Algorithm catalogue table... with: algorithm name, primary source citation, intended use, link to per-algorithm detail in `docs/algorithms.md`".
- **Action:** Code fix (fills in `docs/algorithms.md`, see Critical finding above).

### [Important] CHANGELOG.md has no entries for Phases 1-8
- **File:** /Users/johnny/Development/fuzzymatch/CHANGELOG.md
- **Phase introduced:** Phase 1
- **Issue:** The CHANGELOG `## [Unreleased]` block lists only the project bootstrap (repository structure, licence, CLAUDE.md, requirements doc, GSD). It does NOT list:
  - Phase 1: Normalise, Tokenise, AlgoID, errors, golden fixtures.
  - Phase 2: 6 character-based algorithms + Ratcliff-Obershelp.
  - Phase 3: Smith-Waterman-Gotoh.
  - Phase 4: 4 q-gram algorithms.
  - Phase 5: ASCII fast paths and Cosine.
  - Phase 6: 5 token-based algorithms.
  - Phase 7: 4 phonetic algorithms.
  - Phase 8: Scorer, NewScorer, DefaultScorer, ScoreAll, etc.
- **Standard:** documentation-standards.md "CHANGELOG following Keep-a-Changelog"; `release.yml` extracts the section for GitHub release notes.
- **Action:** Code fix.

### [Important] `docs/scan.md` is a scaffold but `docs/scan.md` is linked from README as a real doc
- **File:** /Users/johnny/Development/fuzzymatch/docs/scan.md; /Users/johnny/Development/fuzzymatch/README.md (line 225)
- **Phase introduced:** Phase 1 (scaffold); should remain as scaffold or be hidden until Phase 9
- **Issue:** `README.md` lists `docs/scan.md — `scan` sub-package consumer guide (Phase 9).` Clicking through gives a "TBD. See docs/requirements.md §12.X" stub for every section. The "(Phase 9)" annotation is helpful but the link target is essentially "we plan to write this". README should make clearer that this is aspirational, or hide the link until Phase 9 ships.
- **Standard:** Consumer-trust — links should go somewhere useful.
- **Action:** Code fix (the docs). Either gate the README link behind Phase 9 completion, or change the stub doc to say "Not yet shipped — track progress in [issue link]".

### [Important] `docs/extending.md` is a scaffold but linked as a real doc
- **File:** /Users/johnny/Development/fuzzymatch/docs/extending.md
- **Phase introduced:** Phase 1
- **Issue:** All four substantive sections (`## Composing domain-specific Scorers`, `## Composing phonetic algorithms with edit distance`, `## Custom inner metric for Monge-Elkan`, `## Custom algorithms outside the catalogue`) begin with "TBD." or are stubs. The README links to this doc. The user-guide-reviewer key question "Can they extend the library?" cannot be answered from current `docs/extending.md`.
- **Standard:** documentation-standards.md "docs/extending.md — building domain-specific Scorers, composing phonetic algorithms with edit distance, custom inner metrics for Monge-Elkan".
- **Action:** Code fix (the docs).

### [Important] `docs/performance.md` is a scaffold but linked as a real doc
- **File:** /Users/johnny/Development/fuzzymatch/docs/performance.md
- **Phase introduced:** Phase 1
- **Issue:** Every section is TBD. "Benchmark methodology" is TBD. "Per-algorithm budgets" is TBD. "Scorer budgets" is TBD. The committed `bench.txt` file exists with full data — but the doc points at `docs/requirements.md` §14 for the real numbers, which is an internal spec.
- **Standard:** documentation-standards.md "docs/performance.md — benchmark numbers, optimisation notes, profiling tips".
- **Action:** Code fix (the docs).

### [Important] No documented worked example for the "scan/Item/Warning" loop in `docs/scan.md`
- **File:** /Users/johnny/Development/fuzzymatch/docs/scan.md
- **Phase introduced:** Phase 9 (not yet shipped — but the scaffold should at least show what the API will look like)
- **Issue:** User-guide-reviewer key question 6: "Can they use the scan sub-package? A worked example from a slice of `Item` to a list of `Warning` in under 10 lines." Currently scan.md is all TBD pointing at requirements.md. A scaffold that previews the API surface would orient consumers planning their Phase 9 adoption.
- **Standard:** user-guide-reviewer.
- **Action:** Code fix (the docs) — when Phase 9 lands. Pre-Phase-9, downgrade to a clear "coming in Phase 9" banner.

### [Important] `docs/scorer.md` Threshold rationale is stated but no quick recommendation table
- **File:** /Users/johnny/Development/fuzzymatch/docs/scorer.md (lines 110-130); /Users/johnny/Development/fuzzymatch/docs/tuning.md (lines 41-58)
- **Phase introduced:** Phase 8
- **Issue:** docs/scorer.md tells you the threshold is mandatory and gives the rationale, but does NOT include a quick lookup table of "for X data, start at threshold Y". `docs/tuning.md` has the canonical calibration loop but it's a 6-step process. A "if you don't have labelled data, start at 0.85 for identifier matching, 0.80 for name matching, 0.70 for free text" table at the top of `docs/scorer.md` § Threshold would close the gap for the user who just wants a sane starting point.
- **Standard:** user-guide-reviewer key question 5: "Can they tune thresholds?... At minimum: starting point recommendations".
- **Action:** Code fix (the docs).

### [Important] Tuning guidance has no runnable calibration script
- **File:** /Users/johnny/Development/fuzzymatch/docs/tuning.md
- **Phase introduced:** Phase 8
- **Issue:** docs/tuning.md describes a 6-step calibration loop but offers no starter Go code. Compare to the high-quality `examples/scorer-composition/`. A `examples/threshold-calibration/` runnable program would close the loop: given a CSV of labelled pairs and a Scorer, print the precision/recall/F1 at every threshold in 0.05 increments. Today the consumer has to write this from scratch.
- **Standard:** user-guide-reviewer "Can they tune thresholds? ... and how to calibrate against a domain corpus."
- **Action:** Code fix (add an example directory) or doc-only fix (paste a starter snippet in tuning.md).

### [Important] `docs/scorer.md` ScoreAll deterministic-content note uses `sort.Slice` but the more idiomatic API is `Scorer.Algorithms()`
- **File:** /Users/johnny/Development/fuzzymatch/docs/scorer.md (lines 227-237)
- **Phase introduced:** Phase 8
- **Issue:** The "Consumers that need stable ordering" example builds a fresh `[]AlgoID`, runs `sort.Slice(...)`, then iterates. But `Scorer.Algorithms()` already returns the algorithm set in AlgoID-ascending order; a more idiomatic example would be:
  ```go
  scores := s.ScoreAll(a, b)
  for _, alg := range s.Algorithms() {
      fmt.Printf("%s: %.4f\n", alg.ID, scores[alg.ID])
  }
  ```
- **Standard:** API ergonomics — show the canonical idiom.
- **Action:** Code fix (the docs).

### [Important] `Match` boundary semantic (inclusive `>=`) is documented but the off-by-one case is not highlighted
- **File:** /Users/johnny/Development/fuzzymatch/docs/scorer.md (line 100, 388 of scorer.go)
- **Phase introduced:** Phase 8
- **Issue:** "True when `Score(a, b) >= Threshold()` (boundary inclusive)" — correct but easy to miss. Consumers calibrating with `WithThreshold(0.85)` and getting score 0.84999999... will see Match=false; consumers getting 0.8500000... will see Match=true. The `>=` choice is a meaningful design decision worth one extra sentence: "If you want strict-greater-than semantics, choose your threshold one ulp lower (or — equivalently — set the threshold to the lowest acceptable score, not the upper bound of unacceptable scores)."
- **Standard:** Documentation clarity for the boundary case.
- **Action:** Code fix (the docs).

### [Important] No callout on Hamming length-mismatch silent-zero policy
- **File:** /Users/johnny/Development/fuzzymatch/docs/algorithms.md (Hamming section is TBD)
- **Phase introduced:** Phase 2
- **Issue:** `HammingScore("abc", "ab") = 0.0` silently (no error, no panic — per the LOCKED policy in `hamming.go`). The Scorer-using consumer who feeds Hamming into a Scorer composition with mixed-length identifier pairs will see Hamming dragging the composite score to zero — and won't know why. This is documented inside Go file comments (`hamming.go` file-level godoc) and in `example_test.go ExampleHammingScore`, but NOT in `docs/algorithms.md` (which is TBD) or in `docs/scorer.md` (which is where Scorer consumers look).
- **Standard:** Consumer pitfall surfacing.
- **Action:** Code fix (the docs). Fill `docs/algorithms.md#hamming` with the length-mismatch policy front and centre.

### [Important] `DefaultScorer` composition stability promise is unstated
- **File:** /Users/johnny/Development/fuzzymatch/docs/scorer.md (lines 132-158)
- **Phase introduced:** Phase 8
- **Issue:** The "Default Composition" table lists the six algorithms. But it does not state whether the six-algorithm set is the v1.x contract or whether a future minor release may add/remove algorithms. The threshold `0.85` is documented as "calibrated for this specific mix" — but if the mix changes, what's the migration story? `docs/extending.md` line 5 mentions "the curated 23-algorithm catalogue is the public v1.x contract" but says nothing about the DEFAULT composition.
- **Standard:** API stability documentation.
- **Action:** Code fix (the docs). Add a "Stability" subsection to "Default Composition" explicitly stating "the six-algorithm composition and the 0.85 threshold are part of the v1.x DefaultScorer contract" (or "are subject to minor-version revision; pin a custom Scorer if reproducibility matters").

### [Important] Scorer "minus DM" example score table includes a counter-intuitive row not explained
- **File:** /Users/johnny/Development/fuzzymatch/examples/scorer-composition/main_test.go (line 49)
- **Phase introduced:** Phase 8
- **Issue:** The committed `want` constant shows `org_id / organisation_id` with Default=0.2911 and MinusDM=0.3493 — i.e. removing DoubleMetaphone INCREASED the composite score. The narrative in `main.go` says the threshold drops from 0.85 → 0.80 because "removing one of six signals reduces the composite ceiling" — but that's exactly contradicted for this row. A short comment in the example noting "for inputs where DoubleMetaphone returned 0 because phonetic codes disagreed, removing it actually raises the composite" would prevent reader confusion.
- **Standard:** Example clarity.
- **Action:** Code fix (the example or the example's comment block).

### [Important] No documented mapping from AlgoID → display name for ScoreAll output
- **File:** /Users/johnny/Development/fuzzymatch/docs/scorer.md (ScoreAll section)
- **Phase introduced:** Phase 8
- **Issue:** `ScoreAll` returns `map[AlgoID]float64`. The example at lines 234-235 uses `fmt.Printf("%s: %.4f\n", id, scores[id])` — and AlgoID does implement `fmt.Stringer` so this works. But the doc doesn't explicitly note "AlgoID.String() returns the canonical display name (e.g. AlgoLevenshtein → 'Levenshtein')". A consumer who tries `fmt.Printf("%d: ...", id)` would get `0: ...` and be puzzled.
- **Standard:** Documentation precision.
- **Action:** Code fix (the docs).

---

## Improvement

### [Improvement] llms.txt status section is out of date (lists only Phase 1 as shipped)
- **File:** /Users/johnny/Development/fuzzymatch/llms.txt (lines 9-14)
- **Phase introduced:** Phase 1
- **Issue:** "Phase 1 ships the foundation primitives: AlgoID dispatch enum, sentinel errors, Normalise, Tokenise. Phase 2+ adds the 23 algorithms, then Scorer (Phase 8) and scan / Extract (Phases 9–10)." This implies Phase 2 has not shipped, but the Public API section below it lists ALL 23 algorithms + Scorer. Internally inconsistent.
- **Standard:** Doc accuracy.
- **Action:** Code fix (the docs).

### [Improvement] llms.txt Scorer section header still references plan numbers
- **File:** /Users/johnny/Development/fuzzymatch/llms.txt (line 187)
- **Phase introduced:** Phase 8
- **Issue:** "### Scorer construction options (Phase 8 — plan 08-01 lays the option layer; plan 08-02 lands NewScorer + Score + Match; plan 08-03 lands ScoreAll + ..." — plan-internal language leaking into the consumer-facing doc.
- **Standard:** Consumer-facing docs should not reference internal plan numbers.
- **Action:** Code fix (the docs).

### [Improvement] llms.txt and README do not mention `examples/` directories
- **File:** /Users/johnny/Development/fuzzymatch/llms.txt; /Users/johnny/Development/fuzzymatch/README.md
- **Phase introduced:** Phase 7
- **Issue:** Already raised above under Critical; mentioned here as the llms.txt has its own Documentation block that misses the examples.
- **Action:** Code fix.

### [Improvement] Hamming algorithm "name" — README catalogue says "Hamming 1950" but per the locked policy it diverges from strict Hamming for unequal-length inputs
- **File:** /Users/johnny/Development/fuzzymatch/README.md (line 132)
- **Phase introduced:** Phase 2
- **Issue:** Strict Hamming-1950 is defined only for equal-length inputs. The library's `HammingScore` returns 0.0 for unequal-length inputs by policy (LOCKED, see hamming.go). The README catalogue entry "Hamming | AlgoHamming | Hamming 1950" doesn't hint at the divergence. Sophisticated readers will not be misled; novice readers will be.
- **Standard:** Algorithm-correctness documentation.
- **Action:** Code fix (the docs).

### [Improvement] README catalogue "Token Sort Ratio / Set Ratio / Partial Ratio" cite "SeatGeek fuzzywuzzy / RapidFuzz" but the actual correctness baseline is RapidFuzz only
- **File:** /Users/johnny/Development/fuzzymatch/README.md (lines 153-155)
- **Phase introduced:** Phase 6
- **Issue:** `docs/cross-validation.md` explicitly states "We cross-validate against RapidFuzz exclusively; fuzzywuzzy is referenced only as historical context." README citation mentions both as if they were equal sources. Minor but consistent across all three rows.
- **Standard:** Algorithm-correctness — cite the actual baseline.
- **Action:** Code fix (the docs).

### [Improvement] README "Three layers" diagram does not show Layer 1 with a complete signature
- **File:** /Users/johnny/Development/fuzzymatch/README.md (lines 80-84)
- **Phase introduced:** Phase 1
- **Issue:** `Layer 1: Algorithm functions      LevenshteinScore(a, b)` — but the function signature is `LevenshteinScore(a, b string) float64`. Showing the return type would convey "you get a float back, easy" immediately.
- **Action:** Code fix (the docs).

### [Improvement] Algorithm catalogue cross-references `AlgoID` constants but doesn't tell the reader what an AlgoID is
- **File:** /Users/johnny/Development/fuzzymatch/README.md (lines 125-171)
- **Phase introduced:** Phase 1
- **Issue:** The catalogue table column "`AlgoID`" shows constants like `AlgoLevenshtein`. A first-time reader has no context — is this an iota-int enum? Is it the function name? The Scorer composition section later in the README is where these constants are used, but that section is missing entirely.
- **Action:** Code fix (the docs).

### [Improvement] README "Algorithm catalogue" omits per-algorithm "intended use" column
- **File:** /Users/johnny/Development/fuzzymatch/README.md (lines 125-171)
- **Phase introduced:** Phase 1
- **Issue:** The catalogue table has columns: Algorithm | AlgoID | Primary source | Detail. The documentation-standards skill specifies "algorithm name, primary source citation, intended use, link to per-algorithm detail in `docs/algorithms.md`." The "intended use" column is missing. Without it, the reader can't make algorithm-selection decisions from the catalogue alone.
- **Standard:** documentation-standards.md "Algorithm catalogue table".
- **Action:** Code fix (the docs).

### [Improvement] README does not have a "Scorer composition" section as required by documentation-standards
- **File:** /Users/johnny/Development/fuzzymatch/README.md
- **Phase introduced:** Phase 8
- **Issue:** documentation-standards.md (lines 28-29) mandates "Scorer composition section with a typical mixed-algorithm example" and "Scan sub-package introduction with a one-screen example". Neither appears in current README. There's a brief "Configuration" section pointing at Phase-1 primitives only.
- **Standard:** documentation-standards.md README structure.
- **Action:** Code fix.

### [Improvement] README "Configuration" section lists Phase-1 option fields only
- **File:** /Users/johnny/Development/fuzzymatch/README.md (lines 178-198)
- **Phase introduced:** Phase 1
- **Issue:** The Configuration section shows `NormalisationOptions` and `TokeniseOptions` field lists but never mentions `ScorerOption` even though Phase 8 has shipped. Should at least preview the Scorer With* options.
- **Action:** Code fix.

### [Improvement] No "What's new in v0.X" status pointer pointing at CHANGELOG
- **File:** /Users/johnny/Development/fuzzymatch/README.md
- **Phase introduced:** Phase 8
- **Issue:** README links to CHANGELOG only indirectly (via CONTRIBUTING.md's PR-checklist). A "Current state" subsection in README pointing at CHANGELOG would help consumers tracking the pre-release progress.
- **Action:** Code fix (the docs).

### [Improvement] FAQ does not cover "how do I pick a threshold"
- **File:** /Users/johnny/Development/fuzzymatch/docs/faq.md
- **Phase introduced:** Phase 1
- **Issue:** FAQ has good entries for why-no-NW, why-no-Metaphone-3, why-no-embeddings, why-phonetic-binary, why-not-generic, why-x/text. Missing: "What threshold should I start with?" "Why is my Match() always false?" (NaN trap from CR-01 once fixed). "Why does my score drop when I add more algorithms?" (weight normalisation surprise). "Why does WithoutNormalisation not actually remove all character transformations?" (token internal Tokenise call).
- **Standard:** documentation-standards.md "docs/faq.md — common questions including..."
- **Action:** Code fix (the docs).

### [Improvement] `docs/scorer.md` Quickstart code does not show the printed output
- **File:** /Users/johnny/Development/fuzzymatch/docs/scorer.md (lines 21-32)
- **Phase introduced:** Phase 8
- **Issue:** The example calls `s.Match("user_id", "userId")` and just notes "// similar" in a `if` branch. A consumer wants to see the actual return: `true` (because the Normalise pipeline maps both to "user id" which is identical pre-tokenise). Showing `// → true` would close the loop.
- **Action:** Code fix (the docs).

### [Improvement] No documented Score result reproducibility-across-patches promise
- **File:** /Users/johnny/Development/fuzzymatch/README.md; /Users/johnny/Development/fuzzymatch/docs/scorer.md
- **Phase introduced:** Phase 8
- **Issue:** README says "Cross-platform deterministic output — verified byte-identical across linux/amd64, linux/arm64, darwin/amd64, darwin/arm64, windows/amd64." But it does not say whether v0.4.0 → v0.4.1 will produce the same float64 result. Consumers persisting scores need to know the patch-version reproducibility story. The internal spec (requirements.md) defines this but consumers won't read that.
- **Standard:** API stability documentation.
- **Action:** Code fix (the docs).

### [Improvement] No glossary or terminology table
- **File:** /Users/johnny/Development/fuzzymatch/docs/
- **Phase introduced:** Phase 1
- **Issue:** Documentation uses "score" (always [0.0, 1.0] composite), "raw score" (algorithm-specific output), "weight" (consumer-supplied input), "normalised weight" (post auto-normalisation), "threshold" (Match boundary), "algorithm" (one of 23), "composition" (set of weighted algorithms inside a Scorer), "Scorer" (the object), "scan" (Phase 9 sub-package). A short glossary in `docs/` would prevent confusion. Today the consumer has to pick up the terminology from prose, where it's sometimes inconsistent ("metric" vs "algorithm" both appear in some places).
- **Action:** Code fix (add `docs/glossary.md` or a glossary section in README).

### [Improvement] Phonetic example's MRACompare output requires reader to know NBS Tech Note 943's threshold table
- **File:** /Users/johnny/Development/fuzzymatch/examples/phonetic-keys/main.go (lines 77-83)
- **Phase introduced:** Phase 7
- **Issue:** Output is `Byrne vs Boern: matched=true sim=5`. A reader without familiarity with NBS Tech Note 943 sees `sim=5` and has no calibration for whether 5 is high/low. A comment block above the loop explaining "MRA returns matched=true when sim>=threshold per the NBS table; sim is a 0-6 counter" would orient new readers.
- **Action:** Code fix (the example).

### [Improvement] No "common pitfalls" section in any consumer doc
- **File:** /Users/johnny/Development/fuzzymatch/docs/
- **Phase introduced:** Phase 8
- **Issue:** Several footguns identified above (NaN threshold, WithoutNormalisation tokeniser surprise, Tversky α+β=0, Hamming length mismatch, ScoreAll iteration non-determinism). A dedicated "Common Pitfalls" page would gather these — current state is one-pitfall-per-doc, scattered.
- **Action:** Code fix (the docs).

### [Improvement] No "from this library to production" checklist (pin Scorer at package level, etc.)
- **File:** /Users/johnny/Development/fuzzymatch/docs/tuning.md (Pinning a calibrated configuration section)
- **Phase introduced:** Phase 8
- **Issue:** `docs/tuning.md` Pinning section is good but is the only place this guidance lives. A production-readiness checklist would help: "pin Scorer at package level; check `errors.Is` against named sentinels; surface ScoreAll only for tuning, not for production lookups; verify determinism golden file in CI; ...". Today the consumer has to infer this from scattered pieces.
- **Action:** Code fix (the docs).

### [Improvement] CONTRIBUTING.md does not mention runnable godoc examples on pkg.go.dev
- **File:** /Users/johnny/Development/fuzzymatch/CONTRIBUTING.md
- **Phase introduced:** Phase 2
- **Issue:** The contributing doc covers conventional commits, CLA, make-target list — but doesn't emphasise that every public function deserves a runnable godoc example (the 39 existing examples are the precedent). A line in the pre-PR checklist would surface this.
- **Action:** Code fix (the docs).

---

## Notes on Phase-9 / Phase-10 docs

`docs/scan.md` (Phase 9) and the Extract API (Phase 10) are deliberately scaffold. These findings cover the scaffolds AS scaffolds — they should not block Phase 9. When Phase 9 lands, re-run user-guide-reviewer over the actual content.

## File paths (all absolute)

- /Users/johnny/Development/fuzzymatch/README.md
- /Users/johnny/Development/fuzzymatch/CHANGELOG.md
- /Users/johnny/Development/fuzzymatch/CONTRIBUTING.md
- /Users/johnny/Development/fuzzymatch/doc.go
- /Users/johnny/Development/fuzzymatch/example_test.go
- /Users/johnny/Development/fuzzymatch/llms.txt
- /Users/johnny/Development/fuzzymatch/llms-full.txt
- /Users/johnny/Development/fuzzymatch/scorer.go
- /Users/johnny/Development/fuzzymatch/scorer_options.go
- /Users/johnny/Development/fuzzymatch/errors.go
- /Users/johnny/Development/fuzzymatch/hamming.go
- /Users/johnny/Development/fuzzymatch/levenshtein.go
- /Users/johnny/Development/fuzzymatch/normalise.go
- /Users/johnny/Development/fuzzymatch/tokenise.go
- /Users/johnny/Development/fuzzymatch/tversky.go
- /Users/johnny/Development/fuzzymatch/docs/algorithms.md
- /Users/johnny/Development/fuzzymatch/docs/cross-validation.md
- /Users/johnny/Development/fuzzymatch/docs/extending.md
- /Users/johnny/Development/fuzzymatch/docs/faq.md
- /Users/johnny/Development/fuzzymatch/docs/performance.md
- /Users/johnny/Development/fuzzymatch/docs/scan.md
- /Users/johnny/Development/fuzzymatch/docs/scorer.md
- /Users/johnny/Development/fuzzymatch/docs/tuning.md
- /Users/johnny/Development/fuzzymatch/examples/identifier-similarity/main.go
- /Users/johnny/Development/fuzzymatch/examples/phonetic-keys/main.go
- /Users/johnny/Development/fuzzymatch/examples/scorer-composition/main.go
- /Users/johnny/Development/fuzzymatch/examples/scorer-composition/main_test.go
- /Users/johnny/Development/fuzzymatch/.planning/phases/08-composite-scorer/08-REVIEW.md (cross-reference for CR-01)
- /Users/johnny/Development/fuzzymatch/.claude/skills/documentation-standards/SKILL.md (governing standard)

</details>

---
