# Phase 6: Token-based Algorithms - Context

**Gathered:** 2026-05-15
**Status:** Ready for planning

<domain>
## Phase Boundary

Ship the five token-based similarity algorithms that consume `Tokenise` (Phase 1) and the existing dispatch table:

- **Monge-Elkan** (`AlgoMongeElkan`, reserved at algoid.go:129-133) — Monge-Elkan 1996, hybrid token-outer + AlgoID-inner
- **Token Sort Ratio** (`AlgoTokenSortRatio`, reserved at algoid.go:135-139) — Indel formula `2·LCS/(|a|+|b|)`, RapidFuzz-aligned
- **Token Set Ratio** (`AlgoTokenSetRatio`, reserved at algoid.go:141-145) — three-way Indel max
- **Partial Ratio** (`AlgoPartialRatio`, reserved at algoid.go:147-150) — sliding-window Indel
- **Token Jaccard** (`AlgoTokenJaccard`, reserved at algoid.go:152-154) — set-Jaccard over `Tokenise(s)`

Plus a shared `token_indel.go` infrastructure file (unexported `lcsLen` + `indelRatio` kernel) consumed by TokenSort, TokenSet, PartialRatio.

Phase 6 closes requirement IDs TOKEN-01 through TOKEN-05.

Out of scope: phonetic algorithms (Phase 7 — but Monge-Elkan's permitted-inner-metric set is designed to extend additively to phonetics in Phase 7 without breaking changes); the Scorer surface (`WithMongeElkanAlgorithm(weight, inner)`, `WithTokenSortRatioAlgorithm(weight)` etc. land in Phase 8); scan / extract sub-packages (Phases 9 / 10).

</domain>

<decisions>
## Implementation Decisions

### §1. Cross-validation source for Indel-based ratios — LOCKED

**`pip install rapidfuzz` + Python generator script + committed JSON corpus. Mirrors Phase 3 SWG / Phase 4 Ratcliff-Obershelp pattern.**

The spec explicitly mandates RapidFuzz cross-validation for the three Indel-based ratios (Token Sort, Token Set, Partial Ratio). Phase 6 introduces `rapidfuzz` as a developer-toolchain dependency (NOT a runtime dependency — root `go.mod` allowlist remains stdlib + `golang.org/x/text`).

**Mechanism:**
- `scripts/gen-token-ratio-cross-validation.py` calls `rapidfuzz.fuzz.token_sort_ratio` / `token_set_ratio` / `partial_ratio` (and the rune-aware variants for PartialRatio runes path) to produce reference vectors.
- Top of script declares `RAPIDFUZZ_VERSION = "<exact-pinned-version>"` and asserts `rapidfuzz.__version__ == RAPIDFUZZ_VERSION`. Refuses to run if the installed version doesn't match.
- `make regen-token-ratio-cross-validation` invokes the script.
- Output: `testdata/cross-validation/token-ratios/vectors.json` — committed to repo; cross-validation tests load and assert the byte-stable corpus matches our implementation.

**Vector volume:** 20-30 entries per algorithm (20-30 × 4 surfaces — TokenSort, TokenSet, PartialRatio bytes, PartialRatio runes — ≈ 80-120 entries total). Vectors must cover: ASCII short, ASCII medium, Unicode, identical, both-empty, one-empty, token-set-asymmetric (different token counts but shared core), and at least one pathological-length pair near the per-algorithm complexity edge.

**Rationale:** the spec's RapidFuzz mandate is the canonical Indel formula attestation — `2·LCS/(|a|+|b|)` is unambiguous across implementations only when cross-validated against a reference that documents the exact formula choice (RapidFuzz publishes the formula in its docs; fuzzywuzzy does not). Hand-computed Indel vectors (Phase 5 q-gram pattern) were rejected because the corpus needs to scale beyond 4-6 vectors to exercise the three-way Indel max in TokenSet and the sliding-window in PartialRatio, and reviewers should not be expected to hand-derive 20+ LCS computations per algorithm. Vendoring RapidFuzz's published test vectors verbatim was rejected because it freezes our corpus to a snapshot that can't be extended with project-specific edge cases.

**Implication for the planner:**
- Add `scripts/gen-token-ratio-cross-validation.py` (developer-only, not committed to runtime path).
- Add `Makefile` target `regen-token-ratio-cross-validation`.
- Pin a specific RapidFuzz version in the script; record it in the cross-validation `vectors.json` header / metadata.
- Write `token_ratio_cross_validation_test.go` (or per-algorithm cross-validation tests) that load the JSON and assert byte-stable score matches.
- Document in CONTRIBUTING.md (or a new `docs/cross-validation.md`) that contributors regenerating vectors must use the pinned RapidFuzz version.

**Scope clarification:** the cross-validation corpus covers ONLY the three Indel-based ratios (TokenSort, TokenSet, PartialRatio bytes, PartialRatio runes). Monge-Elkan and TokenJaccard use hand-derived vectors per §1b below. RapidFuzz's Monge-Elkan implementation is an inner-metric variant whose default may not match ours; cross-validating against it would conflate "ME formula correctness" with "inner-metric fidelity".

### §1b. Monge-Elkan and TokenJaccard reference vectors — LOCKED

**Hand-derived from Monge-Elkan 1996 (ME) and Jaccard 1912 (TokenJaccard). Mirrors Phase 5 q-gram pattern.**

- **Monge-Elkan**: 4-6 hand-derived `(a, b, inner)` test cases per inner-metric pairing. The test comment must show the per-token max derivation (e.g. `tokens(A) = ["user","create"]; tokens(B) = ["usr","creating"]; max_inner(user, *) = JW("user","usr") = 0.875; max_inner(create, *) = JW("create","creating") = 0.95; ME = (0.875+0.95)/2 = 0.9125`). Cover at least: identical token sets → 1.0; disjoint token sets → low score; partially-overlapping with token-count asymmetry; Unicode token; explicit asymmetric pair `ME(A,B) ≠ ME(B,A)` to gate the asymmetric variant.
- **TokenJaccard**: 4-6 hand-derived `(a, b)` pairs per the q-gram Jaccard discipline (Phase 5 §4) — `J = |intersection|/|union|`. Cover identical, disjoint, partial overlap, Unicode tokens.

### §2. Shared token_indel.go foundation — LOCKED

**Mirror Phase 5's `q_gram.go` pattern. One file, one kernel, one optimisation surface.**

Create `token_indel.go` with two unexported helpers consumed by the three Indel-based ratios:

```text
lcsLen(a, b []byte) int                  // longest-common-subsequence length, two-row DP
indelRatio(a, b []byte) float64          // 2·lcsLen(a,b) / (len(a)+len(b)); both-empty → 1.0; one-empty → 0.0
```

Plus rune-aware variants for the PartialRatio runes surface:

```text
lcsLenRunes(a, b []rune) int
indelRatioRunes(a, b []rune) float64
```

**DP optimisation:** two-row `[m+1]int` DP (`prev` + `curr` rows) per the Phase 2/3/4 levenshtein/jaro/swg discipline. Stack-allocated `[64]int` buffer when `min(m, n) ≤ 50`; heap fallback for longer inputs. PartialRatio's sliding-window calls `lcsLen` repeatedly — the stack buffer is the load-bearing optimisation for that hot path.

**Test discipline:** `token_indel_test.go` covers the kernel directly via the `export_test.go` re-export pattern (e.g. `LCSLenForTest`, `IndelRatioForTest`). Each consumer (TokenSort, TokenSet, PartialRatio) tests the COMPOSITION (tokenise → kernel → result) without re-testing the kernel itself.

**Rationale:** three copies of the same DP would diverge under independent maintenance. The shared file mirrors Phase 5's `q_gram.go` (which already proved out the unexported-helper-with-export_test.go-re-export pattern). DRY at the implementation layer; algorithms remain independent at the public-API layer (each ships its own dispatch wrapper, golden file, BDD feature, fuzz harness, bench file).

**Implication for the planner:**
- `token_indel.go` ships in plan 06-01 (foundation plan).
- `token_indel_test.go` covers the kernel; consumers' test files don't duplicate kernel-level cases.
- Allocation budget for `lcsLen` on short ASCII (`min(m,n) ≤ 50`): 0 alloc per call (stack buffer). For longer inputs: 1 alloc per call (one `[]int` heap allocation for the DP rows).
- The kernel is internal; no `LCSLen` / `IndelRatio` public function in v1 (additive change available later if needed).

### §3. Monge-Elkan inner-metric validation — LOCKED

**Strict allow-list. Direct call panics, Scorer returns ErrInvalidAlgoID.**

`monge_elkan.go` declares an unexported `permittedMongeElkanInner = map[AlgoID]bool{...}` enumerating the 13 currently-supported inner AlgoIDs:

- **Character tier (9):** `AlgoLevenshtein`, `AlgoDamerauLevenshteinOSA`, `AlgoDamerauLevenshteinFull`, `AlgoHamming`, `AlgoJaro`, `AlgoJaroWinkler`, `AlgoStrcmp95`, `AlgoSmithWatermanGotoh`, `AlgoLCSStr`
- **Q-gram tier (4):** `AlgoQGramJaccard`, `AlgoSorensenDice`, `AlgoCosine`, `AlgoTversky`

**Behaviour:**
- `MongeElkanScore(a, b, inner, opts)` panics with `"fuzzymatch: AlgoID <X> not permitted as Monge-Elkan inner metric"` when `inner` is not in the map.
- Phase 8's `WithMongeElkanAlgorithm(weight, inner)` returns `ErrInvalidAlgoID` (sentinel already declared in Phase 1) when `inner` is not in the map.
- `AlgoMongeElkan` itself is explicitly NOT in the map (self-reference would infinite-recurse — spec line 744 makes this explicit).
- Token-tier AlgoIDs (`AlgoTokenSortRatio`, `AlgoTokenSetRatio`, `AlgoPartialRatio`, `AlgoTokenJaccard`) are NOT in the map (token-on-token would re-tokenise individual tokens; meaningless composition).

**Phase 7 forward-compatibility:** Phase 7's CONTEXT.md will ADD `AlgoSoundex`, `AlgoDoubleMetaphone`, `AlgoNYSIIS`, `AlgoMRA` to the map (4 new entries). This is an additive change to an internal table — no breaking surface impact.

**Rationale:** runtime panic on malformed `inner` AlgoID surfaces consumer mistakes loudly (Phase 5 §5 LOCKED direct-call panic discipline). The allow-list is the only option that catches `AlgoMongeElkan`-as-inner self-reference at construction time per spec line 744 AND prevents a future asymmetric-or-out-of-range AlgoID from silently breaking ME semantics. Accept-any was rejected because phantom-AlgoID failures would surface at runtime, not at construction.

**Implication for the planner:**
- `monge_elkan.go` declares the map at package scope (not in `init()` per DET-13 / Phase 5 §5 LOCKED).
- `monge_elkan_test.go` includes an exhaustive panic test: every NON-permitted AlgoID in the dispatch table panics with the documented message; every permitted AlgoID returns a score in [0,1].
- The map is the canonical source — when Phase 7 lands, planners ADD entries to this map and update the panic test fixture.

### §4. Monge-Elkan dispatch wrapper — LOCKED

**`dispatch_monge_elkan.go` calls `MongeElkanScoreSymmetric(a, b, AlgoJaroWinkler, DefaultNormalisationOptions)`.**

The dispatch table signature is `func(a, b string) float64`. AlgoMongeElkan needs both an inner metric AND a symmetric-vs-asymmetric choice. The dispatch wrapper picks defaults per the spec:

- Default inner = `AlgoJaroWinkler` (per Monge-Elkan 1996 §3 and spec line 567).
- Symmetric variant (per spec line 571: "Scorer documents that AlgoMongeElkan uses the symmetric variant").

**Rationale:** symmetric ME (average of `ME(A,B)` and `ME(B,A)`) keeps `AlgoMongeElkan` in the standard `PropAlgorithmScore_Symmetric` property test set without needing an exemption. The asymmetric variant is still available as `MongeElkanScore` for direct callers; the Scorer (Phase 8) will expose `WithMongeElkanAlgorithm(weight, inner)` for explicit inner-metric choice using the symmetric variant by default.

**Implication for the planner:** AlgoMongeElkan is added to the symmetric set in `props_test.go` (no exemption). Direct-call asymmetric ME (`MongeElkanScore`) gets its own asymmetry property test (mirrors Tversky α≠β pattern from Phase 5 §5 LOCKED).

### §5. DoS-vector godoc format — LOCKED

**Three-part block: `// Complexity:` formula + `// DoS notice:` paragraph + dedicated `Benchmark<Algo>_Pathological_*` fixture.**

For every algorithm with non-linear worst case (Monge-Elkan, PartialRatio, TokenSetRatio per success criterion 4):

1. **Complexity block** — formal big-O statement with variable definitions:
   ```text
   // Complexity:
   //
   //   O(|A|·|B|·cost(inner))
   //
   //   where |A|, |B| are the post-tokenisation token counts and
   //   cost(inner) is the inner metric's per-comparison complexity.
   ```

2. **DoS notice paragraph** — concrete failure-mode warning naming the input shape that triggers the worst case:
   ```text
   // DoS notice:
   //
   //   On inputs with > 1,000 tokens per side this performs ~10^6 inner-metric
   //   comparisons. In untrusted-input contexts (HTTP request body, file uploads,
   //   user-submitted identifiers), pre-validate token-count ceilings before calling.
   //   See BenchmarkMongeElkan_Pathological_1000Tokens for measured timings.
   ```

3. **Pathological-input benchmark fixture** — `Benchmark<Algo>_Pathological_<shape>` in the algorithm's `_bench_test.go`:
   - Monge-Elkan: `BenchmarkMongeElkan_Pathological_1000Tokens` (1000-token inputs both sides)
   - PartialRatio: `BenchmarkPartialRatio_Pathological_LongShortMismatch` (one short + one very long)
   - TokenSetRatio: `BenchmarkTokenSetRatio_Pathological_AsymmetricSetCardinalities` (small intersection + large difference set)

**Rationale:** the spec criterion's intent is DoS-awareness, not formula recitation. Just stating `O(|A|·|B|·cost)` in godoc satisfies the letter but not the spirit — a reader judging "is this safe to expose to user input" needs concrete numbers. The benchmark fixture is the executable proof: consumers can run it on their hardware and decide whether the worst-case cost is acceptable for their use case.

**Implication for the planner:**
- Each of `monge_elkan.go`, `partial_ratio.go`, `token_set_ratio.go` ships with the three-part block in its package-level godoc.
- The bench files include the pathological fixtures in addition to the standard ASCII Short/Medium/Long + Unicode Short bench cases.
- The `bench.txt` baseline (regenerated in plan 06-N finalisation) includes the pathological numbers — they're part of the regression-detection surface for future PRs.

### §6. Established patterns (LOCKED — inherited from Phase 2/3/4/5)

The following patterns are CARRIED FORWARD without re-discussion:

- **File-by-file structure:** `<algo>.go` (implementation) + `dispatch_<algo>.go` (AlgoID slot wiring) + `<algo>_test.go` (unit + reference vectors) + `<algo>_bench_test.go` (ASCII Short/Medium/Long + Unicode Short, plus pathological fixture per §5 where applicable) + `<algo>_fuzz_test.go` (one `Fuzz*` per public function). Plus shared `token_indel.go` for the LCS/Indel kernel (per §2).
- **Byte path + Rune path applicability:**
  - **Tokenised algorithms** (TokenSortRatio, TokenSetRatio, TokenJaccard, MongeElkan): single surface — `Tokenise` returns `[]string` and operates on UTF-8-aware tokens; rune semantics are inside `Tokenise`. NO `*ScoreRunes` variants.
  - **PartialRatio**: ships BOTH `PartialRatioScore(a, b string) float64` (byte path; sliding window in bytes) and `PartialRatioScoreRunes(a, b string) float64` (rune path) per spec line 609-610. PartialRatio doesn't tokenise.
- **Property tests via `testing/quick`:** RangeBounds, Identity, NoNaN, NoInf, NoNegativeZero on every public function. Symmetric for TokenSortRatio / TokenSetRatio / TokenJaccard / PartialRatio / MongeElkanScoreSymmetric. Asymmetric property test for `MongeElkanScore` (mirrors Tversky α≠β pattern from Phase 5 §5).
- **Direct call panics; Scorer returns errors:** Monge-Elkan inner-metric validation per §3. Other algorithms have no parameters with invalid values (NormalisationOptions defaults are always-valid).
- **No map iteration on output path (DET-03):** TokenJaccard's intersection map iteration uses sorted-key slices on the output path. TokenSet's three-way max also iterates sorted token slices. Monge-Elkan's per-token-max iteration iterates the input slice (already ordered).
- **No transcendental floats (DET-06):** all five algorithms compute scores via integer arithmetic + one final division. NO `math.Sqrt`, NO `math.Pow`, NO transcendentals. Left-to-right reduction with explicit parenthesisation.
- **AlgoID slots already reserved in algoid.go:** `AlgoMongeElkan` (lines 129-133), `AlgoTokenSortRatio` (135-139), `AlgoTokenSetRatio` (141-145), `AlgoPartialRatio` (147-150), `AlgoTokenJaccard` (152-154). Planner wires `dispatch_<algo>.go` files but does NOT modify algoid.go enum positions.
- **BDD scenarios:** one feature file per algorithm in `tests/bdd/features/`. Asymmetric Monge-Elkan scenario explicitly covers `MongeElkanScore(A,B) ≠ MongeElkanScore(B,A)` direction-sensitivity. PartialRatio scenarios cover both byte and rune surfaces.
- **Staging golden → finalisation merge:** `testdata/golden/_staging/<algo>.json` during each algorithm's plan; merged into `testdata/golden/algorithms.json` in the phase-6 finalisation plan.
- **identifier-similarity example extension:** the 14-column example from Phase 5 gets five new columns (one per Phase 6 algorithm) during finalisation. NO new example program.
- **Per-plan llms.txt + llms-full.txt sync:** every new exported symbol gets a line in `llms.txt` and a full entry in `llms-full.txt` IN THE SAME PLAN that adds the symbol — NOT deferred to finalisation (caught Phase 4 plan 04-01 mid-flight; reinforced in Phase 5 plans).

### Claude's Discretion

The planner (gsd-planner) chooses, without re-asking the user:

- Wave decomposition: 5 algorithms + foundation file + finalisation. Foundation (token_indel.go) ships in plan 06-01 alongside its simplest consumer (likely TokenSortRatio — one Indel call). Plans 06-02..06-04 ship TokenSetRatio (three-way Indel max), PartialRatio (sliding window), TokenJaccard (set-Jaccard via Tokenise) in parallel. Plan 06-05 ships Monge-Elkan (most complex — inner-metric dispatch). Plan 06-06 finalisation. Adjust wave grouping if dependency analysis surfaces a different shape.
- Exact internal helper signatures (`lcsLen` byte vs slice signature, whether `indelRatio` takes pre-computed lengths, etc.).
- Stack-buffer size threshold tuning (we recommend `[64]int` at `min(m,n) ≤ 50`; planner may adjust based on benchstat data).
- Exact RapidFuzz version pin (planner picks current stable at planning time; records version + sha256 in script header).
- Exact number of staging-golden entries per algorithm (8-12 per Phase 2-5 norm).
- Whether `Tokenise` is called with `DefaultTokeniseOptions{}` or a phase-6-specific defaults struct (most likely default; only deviation flagged in plan if needed).
- Inner-metric self-test coverage scope for ME (whether to test all 13 permitted AlgoIDs in `monge_elkan_test.go` or a representative subset — recommended: representative subset of 4-5 with comprehensive permission-check fixture covering all 13).

</decisions>

<canonical_refs>
## Canonical References

**Downstream agents MUST read these before planning or implementing.**

### Spec & requirements (project-internal)

- `docs/requirements.md` §7.3.1–§7.3.5 — full token-based algorithm specifications (Monge-Elkan, TokenSortRatio, TokenSetRatio, PartialRatio, TokenJaccard)
- `docs/requirements.md` §10 — Tokenise specification (already implemented; consumed by Phase 6)
- `docs/requirements.md` §13.3, §13.4, §13.6 — determinism rules (no map iteration on output, float stability, left-to-right reduction)
- `docs/requirements.md` §14.1 — per-algorithm allocation budgets
- `docs/requirements.md` §15.3, §15.4 — property test conventions, fuzz target conventions
- `docs/requirements.md` line 744 — `ErrInvalidAlgoID` for Monge-Elkan self-reference
- `docs/requirements.md` line 829 — Token Jaccard default Scorer weight (0.20) — Phase 8 reference
- `docs/requirements.md` lines 1220-1222 — performance budgets for Monge-Elkan / Token ratios / PartialRatio
- `.planning/REQUIREMENTS.md` — TOKEN-01..TOKEN-05 traceability table
- `.planning/ROADMAP.md` — Phase 6 goal + success criteria

### Primary academic / engineering sources (research targets)

- **Monge, A. E., Elkan, C. P. (1996).** "The field matching problem: algorithms and applications." *Proceedings of the Second International Conference on Knowledge Discovery and Data Mining*: 267–270 — Monge-Elkan primary source.
- **SeatGeek (2014).** *fuzzywuzzy* Python library — historical reference (NOT cross-validation source — spec explicitly requires RapidFuzz instead).
- **Bachmann, M. (2020–present).** *RapidFuzz* documentation, https://rapidfuzz.github.io/RapidFuzz/ — canonical Indel-formula reference and cross-validation source for TokenSort / TokenSet / PartialRatio.
- **Jaccard, P. (1912).** "The distribution of the flora in the alpine zone." *New Phytologist*, 11(2):37–50 — Jaccard coefficient origin paper (TokenJaccard primary source — same paper as Q-Gram Jaccard, applied to token sets).
- **Wagner, R. A., Fischer, M. J. (1974).** "The string-to-string correction problem." *Journal of the ACM*, 21(1):168–173 — LCS DP foundation (used by Phase 4 LCSStr; reused for token_indel.go's `lcsLen`).

### External dependency (developer toolchain only — NOT runtime)

- **`rapidfuzz` Python package** (pip install) — required for `make regen-token-ratio-cross-validation`. Pinned to a specific version in `scripts/gen-token-ratio-cross-validation.py`. NEVER added to root `go.mod`.

### Project skills (correctness & licensing gates)

- `.claude/skills/algorithm-correctness-standards/SKILL.md` — primary-source citation, formula docs, reference vectors
- `.claude/skills/algorithm-licensing-standards/SKILL.md` — fresh implementation discipline, attribution format. RapidFuzz cross-validation vectors are loaded from RapidFuzz's MIT-licensed implementation; vectors get attribution headers in the generator script.
- `.claude/skills/determinism-standards/SKILL.md` — no map iteration, float stability, golden files
- `.claude/skills/performance-standards/SKILL.md` — allocation budgets, ASCII fast paths, two-row DP discipline
- `.claude/skills/go-coding-standards/SKILL.md` — Go style, no testify in root tests
- `.claude/skills/go-testing-standards/SKILL.md` — unit + property + fuzz + bench + BDD coverage targets (≥ 95% overall, ≥ 90% per file, 100% on public API)
- `.claude/skills/fuzzymatch-review-protocol/SKILL.md` — agent gate sequence

### Prior phase context (carry-forward)

- `.planning/phases/02-core-character-algorithms-six/02-CONTEXT.md` — file-by-file structure, byte+rune pattern, AlgoID dispatch, Jaro-Winkler implementation (default Monge-Elkan inner metric)
- `.planning/phases/03-smith-waterman-gotoh/03-CONTEXT.md` — Python-generator-plus-committed-corpus cross-validation pattern (Phase 6 §1 inherits this discipline with rapidfuzz instead of biopython)
- `.planning/phases/04-remaining-character-gestalt/04-CONTEXT.md` — staging-golden → finalisation merge, identifier-similarity example extension pattern, RatcliffObershelp (the difflib-equivalent that consumers should compare token-ratios against per spec)
- `.planning/phases/05-q-gram-algorithms/05-CONTEXT.md` — q_gram.go shared-foundation pattern (token_indel.go mirrors this exactly), direct-call-panic + Scorer-returns-error split (carries to ME inner-metric validation), per-plan llms.txt sync discipline

</canonical_refs>

<code_context>
## Existing Code Insights

### Reusable Assets

- **`tokenise.go`** — Phase 1 deliverable. `Tokenise(s string, opts TokeniseOptions) []string` with camelCase / snake_case / PascalCase / kebab-case / dot-case splitters. `DefaultTokeniseOptions` has Lowercase: true and the standard separator set. `TokeniseRunes` exists for rune-aware variant. Phase 6 token-based algorithms call `Tokenise(s, DefaultTokeniseOptions())`.
- **`algoid.go` lines 129-154** — `AlgoMongeElkan`, `AlgoTokenSortRatio`, `AlgoTokenSetRatio`, `AlgoPartialRatio`, `AlgoTokenJaccard` enum slots already reserved with godoc citing primary sources. Slots are stable; planner wires dispatch files but does NOT renumber.
- **`algoid.go` String() / dispatch tables** — string conversions return canonical names (`"MongeElkan"`, `"TokenSortRatio"`, etc.). Dispatch table at lines 280+ already includes these AlgoIDs as legal entries; `dispatch_<algo>.go` files in this phase wire the actual algorithm functions.
- **`errors.go`** — `ErrInvalidAlgoID` is declared at the package level per Phase 1 (FOUND-02 wired the full error sentinel set up front); used by Phase 8's `WithMongeElkanAlgorithm` for inner-metric validation.
- **`testdata/golden/algorithms.json`** — established golden file (now ~280 entries after Phase 5 finalisation). Phase 6 finalisation adds entries for the five token algorithms.
- **`testdata/golden/_staging/` workflow** — pattern proven over 11 algorithm staging files (jaro, jaro_winkler, smith_waterman_gotoh, strcmp95, lcsstr, ratcliff_obershelp, qgram_jaccard, sorensen_dice, cosine, tversky, plus inherited Phase 1 baselines).
- **`tests/bdd/steps/algorithms_steps.go`** — accumulator file. Each new algorithm appends a `step.Step` registration following the Phase 2/3/4/5 pattern.
- **`example_test.go`** — runnable godoc examples; one new `ExampleXxx` per public function. Phase 6: 6 new examples minimum (MongeElkanScore, MongeElkanScoreSymmetric, TokenSortRatioScore, TokenSetRatioScore, PartialRatioScore, PartialRatioScoreRunes, TokenJaccardScore = 7).
- **`props_test.go`** — appendable property-test accumulator. Phase 6 appends Five-invariant blocks per algorithm + Symmetric for TokenSort/Set/Jaccard/Partial/MEsymmetric + Asymmetry for direct ME.
- **`bench.txt`** — Phase 5 regenerated baseline. Phase 6 finalisation full-replaces it again to include the five new benchmarks (plus the three pathological fixtures per §5).
- **`llms.txt` / `llms-full.txt`** — every new exported symbol needs a line in `llms.txt` and a full entry in `llms-full.txt` (per-plan discipline, NOT deferred to finalisation — Phase 5 reinforced this rule).
- **Phase 4's `RatcliffObershelp` godoc** — already cites the difflib-vs-Indel distinction; Phase 6 algorithm godocs reference RatcliffObershelp as the difflib-semantic alternative for consumers wanting the difflib formula instead of Indel.
- **Phase 5's `q_gram.go`** — direct template for `token_indel.go` (unexported helpers + export_test.go re-export + comprehensive kernel tests + consumers test composition only).

### Established Patterns (Phase 6 must follow)

- **Two-row DP buffer pattern (from Phase 2 onwards)** — directly applies to `token_indel.go`'s `lcsLen` (per §2 LOCKED).
- **Map-iteration discipline (DET-03 from Phase 1)** — TokenJaccard's intersection map MUST NOT be iterated on output paths; sort keys before any score-affecting iteration.
- **Source-Origin Statement format** — every algorithm file's header has a `Sources:` block listing primary, cross-validation, and code-consulted references. Phase 4 WR-01 (commit 65302b8) established the precedent.
- **OQ-X RESOLUTION LOCKED format** — when the planner / executor make a non-obvious local decision (e.g. RapidFuzz version pin, ME inner-metric map ordering, stack-buffer size), it's recorded in the plan's SUMMARY.md as `OQ-N RESOLUTION LOCKED <date>` with rationale.
- **gsd-executor.md scope-boundary rule** — out-of-scope discoveries during execution are logged to `.planning/phases/06-token-based-algorithms/deferred-items.md` (created on first need), NOT rolled into the current commit. Phase 4 → Phase 5 demonstrated the clean roll-in path.
- **Direct-call panic + Scorer-returned-error split** — direct ME calls panic on invalid inner AlgoID; Phase 8's `WithMongeElkanAlgorithm` returns `ErrInvalidAlgoID` for the same input.

### Integration Points

- **Phase 7 (Phonetic) extends ME's permitted-inner-metric set:** Phase 7's CONTEXT.md will add `AlgoSoundex`, `AlgoDoubleMetaphone`, `AlgoNYSIIS`, `AlgoMRA` to the `permittedMongeElkanInner` map declared in `monge_elkan.go`. Phase 7's `monge_elkan_test.go` UPDATES the panic test fixture (currently 13 permitted, becomes 17). Additive; no breaking change.
- **Phase 8 (Composite Scorer):** `WithMongeElkanAlgorithm(weight, inner AlgoID)`, `WithTokenSortRatioAlgorithm(weight)`, `WithTokenSetRatioAlgorithm(weight)`, `WithPartialRatioAlgorithm(weight)`, `WithTokenJaccardAlgorithm(weight)` options consume the five algorithms. The Scorer parameter-validation path uses `ErrInvalidAlgoID` for ME (already declared).
- **Phase 9 (Scan) / Phase 10 (Extract):** consume the dispatch table; token-based algorithms participate via their AlgoID slots without scan/extract needing to know they are token-flavoured.

</code_context>

<specifics>
## Specific Ideas

- **RapidFuzz version pin record:** the generator script header includes `RAPIDFUZZ_VERSION = "<exact-version>"` and `RAPIDFUZZ_INSTALLED_AT = "<ISO timestamp>"`. The committed `vectors.json` includes a `_metadata` block with the version + timestamp + a checksum of the script that produced it.
- **Monge-Elkan asymmetric test pair:** at least one ME unit test pair `(a, b, inner)` where `MongeElkanScore(a, b, inner, opts) ≠ MongeElkanScore(b, a, inner, opts)` — the asymmetry gate. Without this, the inner-metric directional max could be silently averaged in the wrong direction and tests would pass.
- **PartialRatio rune surface:** spec explicitly requires both `PartialRatioScore` (byte path) AND `PartialRatioScoreRunes` (rune path). The rune path uses `lcsLenRunes` from `token_indel.go`. Bench files cover both surfaces.
- **TokenSet three-way max:** TokenSet's three substrings (intersection sorted-and-joined; intersection + diff-from-A; intersection + diff-from-B) MUST be deterministic — sort tokens before joining (DET-03). Test the three-way max with at least one pair where the max is NOT the intersection-only ratio (forces all three sub-ratios to compute correctly).
- **Pathological-input bench fixtures:**
  - `BenchmarkMongeElkan_Pathological_1000Tokens` — 1000-token both-sides input
  - `BenchmarkPartialRatio_Pathological_LongShortMismatch` — `len(short)=10`, `len(long)=10000`
  - `BenchmarkTokenSetRatio_Pathological_AsymmetricSetCardinalities` — 5-token vs 100-token, 2-token shared core

</specifics>

<deferred>
## Deferred Ideas

- **Public token_indel helpers (`LCSLen`, `IndelRatio`):** considered during the §2 discussion; deferred to v1.x as additive change if real demand emerges. Promoting an unexported helper to public is non-breaking; the inverse is not.
- **Cross-validation corpus extension to Monge-Elkan via `textdistance.MongeElkan`:** considered during §1 discussion; rejected because RapidFuzz's ME default inner may not match ours and `textdistance` would add a second pip dep. Hand-derived from Monge-Elkan 1996 is sufficient per §1b.
- **Vendored RapidFuzz test vectors verbatim:** considered during §1 discussion; rejected because it freezes the corpus to a snapshot that can't be extended with project-specific edge cases.
- **PartialRatio sliding-window DP optimisation (O(nm) instead of O(nm·(n-m))):** explicitly deferred by spec line 612 to v1.x. Phase 6 ships the straightforward implementation per spec; planner adds a `// TODO(#<issue>): implement sliding-window DP per Bachmann RapidFuzz docs` referencing a future GitHub issue (planner creates the issue if it doesn't exist).
- **`SoftTFIDF` token algorithm:** explicitly OUT OF SCOPE in PROJECT.md ("requires a corpus model; out of scope for a pure-function library"). Not deferred — explicitly excluded from v1.0.
- **Token Indel ratio variant using Levenshtein distance (instead of LCS-based Indel):** considered during §2 implicit discussion; rejected because the spec explicitly mandates the Indel formula `2·LCS/(|a|+|b|)`. Levenshtein-based variant is documented as available via `1 - LevenshteinScore` composition for consumers who want it.
- **Reqcontextual UTF-8 normalisation choice between NFC and NFD for token comparison:** considered briefly during ME inner-metric discussion; deferred — `Normalise` (Phase 1) handles NFC/NFD per `NormalisationOptions`; token algorithms accept the same options struct and pass it through. No phase-6-specific decision needed.
- **Custom `TokeniseOptions` for token algorithms:** considered during scout phase; deferred — `DefaultTokeniseOptions{}` (Lowercase: true, standard separator set) is the spec-default for all token algorithms. Consumers wanting different tokenisation call `Tokenise` themselves and use a different algorithm pathway.

</deferred>

---

*Phase: 6-token-based-algorithms*
*Context gathered: 2026-05-15*
