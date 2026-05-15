# Phase 6: Token-based Algorithms - Discussion Log

**Date:** 2026-05-15
**Mode:** default (4 single-question turns per area)
**Areas selected:** all four (RapidFuzz cross-validation source, Shared `token_indel.go` foundation, Monge-Elkan inner-metric validation, DoS-vector godoc format)

This log captures the conversational record. Canonical decisions are in `06-CONTEXT.md`. This file is for human reference only — downstream agents (researcher, planner, executor) consume CONTEXT.md exclusively.

---

## Area Selection

**Question:** Phase 6 has four implementation gray areas. Which do you want to discuss before research/planning kicks off?

**Options presented (multi-select):**
- ☐ RapidFuzz cross-validation source — spec mandates RapidFuzz; not stdlib like Phase 4, not hand-computed like Phase 5
- ☐ Shared token_indel.go foundation — mirror Phase 5's q_gram.go shared-helper pattern, or inline per algorithm
- ☐ Monge-Elkan inner-metric validation — strict allow-list vs accept-any vs self-recursion-only guard
- ☐ DoS-vector godoc format — complexity-only vs concrete warning vs benchmark fixture

**User selection:** All four.

---

## §1. RapidFuzz cross-validation source

### Q1.1 — Where do reference vectors come from?

**Options:**
1. **pip install rapidfuzz + generator script (Recommended)** — adds rapidfuzz pip dep to dev toolchain (NOT runtime). Generator produces `testdata/cross-validation/token-ratios/vectors.json`. Mirrors Phase 3 (biopython for SWG) and Phase 4 (difflib for Ratcliff-Obershelp).
2. Hand-computed Indel formula vectors — Phase 5 q-gram pattern. No pip dep but harder to scale beyond 4-6 vectors per algorithm.
3. Vendor RapidFuzz test vectors verbatim — frozen corpus, no pip dep, but bespoke maintenance.

**User selection:** Option 1 (pip install rapidfuzz + generator script).

**Notes:** Mirrors the Phase 3/4 pattern. RapidFuzz is MIT-licensed, single pip package, well-maintained. Cross-validation discipline matches the established norm.

---

### Q1.2 — Vector volume and version pinning

**Options:**
1. **20-30 per algo + version pinned in script + Makefile target (Recommended)** — `RAPIDFUZZ_VERSION` constant at top of generator, asserts `rapidfuzz.__version__` matches. `make regen-token-ratio-cross-validation` invokes it.
2. 10-15 per algo, version pinned in requirements file — lighter corpus, less coverage of edge cases.
3. 20-30 per algo, version pinned via uv / pip-tools lockfile — overkill for one package.

**User selection:** Option 1 (20-30 per algo + version pinned in script + Makefile target).

**Notes:** Matches Phase 3/4 corpus density. Vector count enables coverage of ASCII short/medium, Unicode, identical, empty, one-empty, token-set-asymmetric, pathological-length cases. Generator script asserts version match before running — refuses with stale rapidfuzz.

---

### Q1.3 — Cross-validation scope (which algorithms)

**Options:**
1. **No — hand-derived from Monge-Elkan 1996 (Recommended)** — RapidFuzz corpus scoped to TokenSort + TokenSet + PartialRatio (bytes + runes); ME and TokenJaccard use hand-derived vectors.
2. Yes — add Monge-Elkan vectors with explicit inner=Jaro-Winkler match — adds ME inner-metric correctness vector.
3. Yes for ME and TokenJaccard too — unify all 5 algos under one corpus.

**User selection:** Option 1 (RapidFuzz corpus scoped to the three Indel-based ratios only).

**Notes:** Spec mandates RapidFuzz only for the Indel-based ratios. Monge-Elkan's primary source is Monge-Elkan 1996 (academic paper); RapidFuzz's ME default inner may not match ours, conflating "ME formula correctness" with "inner-metric fidelity". TokenJaccard formula is trivial (set Jaccard).

**§1 LOCKED.**

---

## §2. Shared token_indel.go foundation

### Q2.1 — Where does the LCS / Indel kernel live?

**Options:**
1. **Shared token_indel.go (Recommended)** — mirror Phase 5 q_gram.go: unexported `lcsLen` + `indelRatio` consumed by token_sort_ratio.go, token_set_ratio.go, partial_ratio.go. Single optimisation surface.
2. Inline kernel per algorithm file — three copies; risk of divergence.
3. Hybrid: shared `lcs.go` for `lcsLen` only, inline `indelRatio` per algo — dubious win.

**User selection:** Option 1 (shared token_indel.go mirroring Phase 5 q_gram.go).

**Notes:** DRY at the kernel level; algorithms remain independent at the public-API layer. The export_test.go re-export pattern from q_gram.go applies directly.

---

### Q2.2 — DP table optimisation strategy

**Options:**
1. **Two-row DP with stack-buffer fast path for short ASCII (Recommended)** — `[64]int` stack buffer when `min(m,n) ≤ 50`. PartialRatio's sliding window calls lcsLen repeatedly — stack buffer is critical for that hot path.
2. Two-row DP without stack buffer — simpler; worse worst-case alloc behaviour.
3. Full m×n DP table — easier to debug; memory O(mn) instead of O(min(m,n)). Rejected by performance-standards.

**User selection:** Option 1 (two-row DP + stack buffer).

**Notes:** Matches Phase 2/3/4 levenshtein/jaro/swg discipline. Allocation budget: 0 alloc per call for short ASCII (≤50), 1 alloc per call for longer inputs.

**§2 LOCKED.**

---

## §3. Monge-Elkan inner-metric validation

### Q3.1 — How does Monge-Elkan validate the `inner AlgoID` parameter?

**Options:**
1. **Strict allow-list (Recommended)** — `permittedMongeElkanInner = map[AlgoID]bool{...}` with 13 currently-supported AlgoIDs (9 char + 4 q-gram). Direct call panics on disallowed; Scorer returns ErrInvalidAlgoID. Phase 7 ADDS phonetic AlgoIDs.
2. Accept any AlgoID + document consumer responsibility — risk of AlgoMongeElkan-as-inner self-reference (infinite recursion) and silent breakage of future asymmetric AlgoIDs.
3. Accept any AlgoID with self-recursion guard only — blocks AlgoMongeElkan as inner; doesn't enforce [0,1] range.

**User selection:** Option 1 (strict allow-list).

**Notes:** Only option that catches AlgoMongeElkan-as-inner self-reference at construction time per spec line 744 AND prevents future asymmetric/out-of-range AlgoIDs from silently breaking ME semantics. Map declared at package scope (not in init() per DET-13).

---

### Q3.2 — What does AlgoMongeElkan return when invoked via the dispatch table?

**Options:**
1. **Symmetric ME, default inner = Jaro-Winkler (Recommended)** — `MongeElkanScoreSymmetric(a, b, AlgoJaroWinkler, DefaultNormalisationOptions)`. Matches spec line 567 + line 571.
2. Asymmetric ME, default inner = Jaro-Winkler — cheaper but breaks AlgoMongeElkan in PropAlgorithmScore_Symmetric set; would need a Tversky-style exemption.

**User selection:** Option 1 (symmetric ME with Jaro-Winkler default).

**Notes:** Keeps AlgoMongeElkan in the standard symmetric property test set. Direct-call asymmetric ME (`MongeElkanScore`) gets its own asymmetry property test mirroring Tversky α≠β pattern from Phase 5.

**§3 LOCKED.**

---

## §4. DoS-vector godoc format

### Q4.1 — What does the DoS / complexity godoc block look like?

**Options:**
1. **Complexity + concrete pathological-input warning + benchmark fixture (Recommended)** — three-part block: `// Complexity:` formula, `// DoS notice:` paragraph naming concrete failure mode, dedicated `Benchmark<Algo>_Pathological_*` fixture in bench file.
2. Complexity statement only — satisfies criterion's letter, not spirit.
3. Complexity + DoS notice (no benchmark fixture) — cheaper than option 1; loses executable proof.

**User selection:** Option 1 (three-part block with benchmark fixture).

**Notes:** Spec criterion's intent is DoS-awareness, not formula recitation. Benchmark fixtures included in `bench.txt` baseline for regression detection.

**§4 LOCKED.**

---

## Continuation Check

**Question:** Anything else to discuss before I write CONTEXT.md and DISCUSSION-LOG.md?

**Options:**
1. **Seal it — write CONTEXT.md (Recommended)** — four LOCKED sections + inherited Phase 2-5 patterns are sufficient for research and planning to proceed.
2. One more area — surface another gray area (wave decomposition, Tokenise default options, golden-file vector counts, BDD scenario shape, llms.txt sync timing).

**User selection:** Option 1 (seal it).

---

## Deferred Ideas Captured

- Public `LCSLen` / `IndelRatio` helpers — non-breaking promotion path available in v1.x.
- Cross-validation corpus extension to Monge-Elkan via `textdistance.MongeElkan` — second pip dep; rejected.
- Vendored RapidFuzz test vectors verbatim — frozen corpus; rejected.
- PartialRatio sliding-window DP optimisation (O(nm) instead of O(nm·(n-m))) — explicitly deferred by spec line 612 to v1.x; planner adds TODO referencing future GitHub issue.
- `SoftTFIDF` token algorithm — explicitly OUT OF SCOPE in PROJECT.md.
- Token Indel ratio variant using Levenshtein distance — rejected; spec mandates LCS-based Indel.
- NFC/NFD normalisation for token comparison — `Normalise` handles via NormalisationOptions; no phase-6 decision needed.
- Custom `TokeniseOptions` for token algorithms — `DefaultTokeniseOptions{}` is spec-default.

---

## Claude's Discretion (planner picks without asking)

- Wave decomposition: 5 algorithms + foundation + finalisation. Foundation (token_indel.go) ships in plan 06-01 alongside simplest consumer (likely TokenSortRatio). Plans 06-02..06-04 ship TokenSetRatio, PartialRatio, TokenJaccard in parallel. Plan 06-05 Monge-Elkan. Plan 06-06 finalisation.
- Exact internal helper signatures.
- Stack-buffer size threshold tuning.
- Exact RapidFuzz version pin (current stable at planning time).
- Exact number of staging-golden entries per algorithm (8-12 per Phase 2-5 norm).
- Tokenise option struct (most likely DefaultTokeniseOptions).
- Inner-metric self-test coverage scope (recommended: representative subset of 4-5 with comprehensive permission-check fixture covering all 13).

---

*Phase: 6-token-based-algorithms*
*Discussion captured: 2026-05-15*
