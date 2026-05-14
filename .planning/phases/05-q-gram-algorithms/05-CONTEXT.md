# Phase 5: Q-gram Algorithms - Context

**Gathered:** 2026-05-14
**Status:** Ready for planning

<domain>
## Phase Boundary

Ship a shared q-gram extraction infrastructure (`q_gram.go`) plus the four q-gram-based similarity algorithms that consume it:

- **Q-Gram Jaccard** (`AlgoQGramJaccard`, reserved) — Ukkonen 1992 / Jaccard 1912
- **Sørensen-Dice** (`AlgoSorensenDice`, reserved) — Dice 1945 / Sørensen 1948
- **Cosine** n-gram (`AlgoCosine`, reserved) — Salton & McGill 1983
- **Tversky** index (`AlgoTversky`, reserved) — Tversky 1977

The load-bearing acceptance test is **byte-identical Cosine output across the CI matrix** (linux/amd64, linux/arm64, darwin/arm64, windows/amd64). Phase 5 closes requirement IDs QGRAM-01 through QGRAM-05.

Out of scope: Monge-Elkan / Token Ratios (Phase 6 consumes the q-gram infrastructure as one permitted inner metric); phonetic algorithms (Phase 7); the Scorer surface (`WithQGramJaccardAlgorithm`, `WithCosineAlgorithm`, `WithTverskyAlgorithm`) lands in Phase 8.

</domain>

<decisions>
## Implementation Decisions

### §1. Cosine cross-platform determinism gate — LOCKED

**Mechanism: existing `testdata/golden/algorithms.json` + cross-platform CI matrix. No new keystone fixture.**

The Phase 1 `algorithms.json` + `make verify-determinism` pipeline already runs on the five-platform CI matrix and byte-diffs the JSON output. Cosine joins that golden file via the staging-golden pattern established in Phase 2 — staged in `testdata/golden/_staging/cosine.json` during plan 05-03, merged into `algorithms.json` during plan 05-05 finalisation, then byte-identical across `linux/amd64`, `linux/arm64`, `darwin/arm64`, `windows/amd64` is the gate.

**Rationale:** the existing infrastructure already detects single-byte divergence on every PR. Adding a separate `cosine_determinism.json` fixture would duplicate the gate without adding signal; `math.Float64bits` hex pins are no stronger than JSON `%.17g` float printing when the golden file is byte-compared. The reviewer suggested a property test pinning random `(a, b, n)` triples via a big.Float reference — rejected: high harness complexity for a redundant safety net.

**Implication for the planner:** the algorithms.json golden entries for Cosine MUST include both ASCII and Unicode inputs at each of `n = 2, 3, 4` to exercise the cross-platform float-reduction path on multiple intersection sizes. RESEARCH.md should enumerate the specific inputs.

### §2. Q-gram extraction public API surface — LOCKED

**Internal only. No public q-gram helper exported in v1.**

Public surface for the q-gram tier is exactly the four algorithm `Score`/`ScoreRunes` functions:

```text
QGramJaccardScore(a, b string, n int) float64
QGramJaccardScoreRunes(a, b string, n int) float64
SorensenDiceScore(a, b string, n int) float64
SorensenDiceScoreRunes(a, b string, n int) float64
CosineScore(a, b string, n int) float64
CosineScoreRunes(a, b string, n int) float64
TverskyScore(a, b string, n int, alpha, beta float64) float64
TverskyScoreRunes(a, b string, n int, alpha, beta float64) float64
```

The shared extraction helpers in `q_gram.go` are unexported — naming convention `extractQGrams(s string, n int) map[string]int` and `extractQGramsRunes(s string, n int) map[string]int` (planner picks the exact signatures, but must NOT export them). Matches the spec phrasing "consumed by the four downstream algorithms via internal API" and follows the Phase 1 / Phase 2 surface-minimisation discipline.

**Rationale:** exposing `QGrams(s, n) []string` would invite consumers to build custom q-gram-based algorithms outside the dispatch table and would cement the q-gram representation as a public contract — too early. The internal helper is the single source of truth; if real demand emerges in v1.x for a custom q-gram metric, the public helper can be promoted (additive change, no breaking guarantee).

**Implication for the planner:** the dispatch table (`dispatch_qgram_jaccard.go`, etc.) calls the internal extractor; no exported helper appears in `llms.txt` / `llms-full.txt` for q-gram extraction itself.

### §3. Cosine dot-product iteration order — LOCKED

**Sort intersection keys alphabetically before the dot-product loop.**

Concretely: build the intersection key set, copy into a `[]string`, `sort.Strings(keys)`, then iterate `for _, k := range keys { dot += float64(countA[k]) * float64(countB[k]) }` with explicit `(x*y) + z` parenthesisation per CHAR-Cosine determinism rules.

**Rationale:** Go map iteration is randomised; iterating the multiset map directly is the textbook determinism-burnpit. Sorting keys before iteration is the same pattern that secured Phase 4's golden-determinism gate and is the simplest argument for the determinism-reviewer (one line of code = one paragraph of justification). The insertion-order key-slice alternative is marginally cheaper but adds API discipline burden across the q-gram tier (every helper must populate two parallel structures in lock-step) — not worth the cycle savings for inputs in the 10–500 char regime that dominates the benchmark suite.

**Implication for the planner:** Cosine's hot path includes `sort.Strings` on the intersection key slice. Allocation budget for Cosine should reserve one `[]string` allocation per call for the sorted key slice; the existing `make([]string, 0, len(intersection))` + `sort.Strings` pattern is idiomatic and benchmark-friendly. The same iteration-order rule applies to any future Cosine variant.

### §4. Cross-validation reference vectors — LOCKED

**Hand-computed from primary academic sources. No external Python toolchain.**

Each algorithm's unit tests include canonical reference vectors derived by hand from the primary source. Cosine — the load-bearing cross-platform determinism algorithm — gets a higher density of hand-verified pairs to compensate for not having an independent external reference:

- **Q-Gram Jaccard** — Ukkonen 1992 §3 worked example (`"AGCT"`, `"AGCTAGCT"`, q = 2 → |A| = 3, |B| = 7, |A∩B| = 3, |A∪B| = 7, J = 3/7 ≈ 0.4286). Plus 2-3 additional pairs covering: identical strings, no-overlap strings, single-shared-q-gram. Minimum 4 reference vectors.
- **Sørensen-Dice** — Dice 1945 §3 worked example + the canonical `"night"` / `"nacht"` bigram pair widely cited in NLP textbooks. Plus 2 additional pairs covering different `n` (trigrams as well as bigrams). Minimum 4 reference vectors.
- **Cosine** — Salton & McGill 1983 §4.1 + **3-5 hand-verified pairs** where the dot product and both norms are documented in the test comment with full float64 precision (17 significant digits). Each pair shows the derivation: "`cos = (dot) / (sqrt(|A|²) × sqrt(|B|²)) = (X) / (sqrt(Y) × sqrt(Z)) = R`". Inputs must span: (a) at least one short ASCII trigram pair where the value is irrational (e.g. `0.7071067811865476` to exercise `math.Sqrt` precision), (b) at least one pair where the q-gram intersection has > 5 elements (exercises the sorted-key accumulation order from §3), (c) one Unicode/runes pair, (d) at least one pair at each of `n = 2, 3, 4`. The hand-verified density on Cosine is the load-bearing gate — without an external library, the unit-test references ARE the third-party-reviewable proof of correctness.
- **Tversky** — Tversky 1977 §2 worked examples + a discriminating asymmetric pair where `TverskyScore(a, b, n, 0.8, 0.2) ≠ TverskyScore(b, a, n, 0.8, 0.2)` to gate the asymmetry implementation. Plus the canonical `α = β = 1.0` (reduces to Jaccard) and `α = β = 0.5` (reduces to Dice) cross-check pairs that confirm Tversky degenerates correctly. Minimum 4 reference vectors.

**Rationale:** the q-gram formulas are arithmetic-short (each fits in 3-5 lines of formula). Hand computation is verifiable by inspection in code review; phase 4 already established that Python-stdlib-only is the right structural target for cross-validation (downgrading from Phase 3's biopython dependency). Adding `textdistance` or `scipy` would re-introduce the developer-toolchain burden Phase 4 just eliminated. The 3-5 hand-verified Cosine pairs with full float64 precision derivation in test comments serve the same purpose as a generated corpus (reviewer-verifiable expected values) without the regeneration toolchain. Reviewers can re-derive any reference value from the formula and the input pair in under 30 seconds.

**Implication for the planner:** NO `scripts/gen-qgram-cross-validation.py`. NO `testdata/cross-validation/q-gram/vectors.json`. The reference vectors live inline in `<algo>_test.go` as table-driven test cases with the formula derivation in the test's godoc comment. The plan should not allocate a wave to "build a cross-validation generator" — it doesn't exist in this phase. Cosine's test file MUST include the 3-5 hand-derivation comment blocks; an algorithm-correctness reviewer should be able to verify each one against Salton & McGill in under a minute.

**Fallback (not active):** if an algorithm-correctness reviewer surfaces a specific concern about hand-computation discipline during review, the fallback is to add `textdistance` cross-validation (single pip dep, covers all four algorithms) following the Phase 3 SWG / Phase 4 RO generator pattern. This is NOT in scope for this phase and does NOT need a placeholder plan — only document the fallback path if the situation arises.

### §5. Established patterns (LOCKED — inherited from Phase 2/3/4)

The following patterns are CARRIED FORWARD from prior phases without re-discussion:

- **File-by-file structure:** `<algo>.go` (implementation) + `dispatch_<algo>.go` (AlgoID slot wiring) + `<algo>_test.go` (unit + reference vectors) + `<algo>_bench_test.go` (ASCII Short/Medium/Long + Unicode Short) + `<algo>_fuzz_test.go` (one `Fuzz*` per public function). Plus shared `q_gram.go` for the extraction infrastructure.
- **Byte path + Rune path:** every algorithm ships `*Score` (byte path) and `*ScoreRunes` (rune path) public functions. Tversky also ships both, with `alpha`/`beta` as additional `float64` parameters.
- **Property tests via `testing/quick`:** RangeBounds, Identity, NoNaN, NoInf, NoNegativeZero on both byte and rune surfaces (FIVE invariants × 2 surfaces × 4 algorithms = 40 property tests minimum). Symmetric for Jaccard / Dice / Cosine; Tversky is symmetric only when `α = β` (asymmetric property test for `α ≠ β`).
- **Direct call panics; Scorer returns errors:** `n < 1` → `panic("fuzzymatch: invalid q-gram size")` on direct call; `WithQGramJaccardAlgorithm(weight, n)` returns `ErrInvalidQGramSize` at construction time (already declared in errors.go per Phase 1). Tversky parameter validation: `alpha < 0 || beta < 0 || alpha+beta == 0` → `panic("fuzzymatch: invalid tversky parameter")` on direct call.
- **No map iteration on output path (DET-03):** the multiset map is internal scratch space only; every output path iterates a sorted-key slice or a derived structure.
- **No transcendental floats (DET-06):** `math.Sqrt` only (Cosine norms). NO `math.Pow`, NO `math.Log`, NO `math.Exp`, NO `math.FMA`. Left-to-right reduction with explicit parenthesisation throughout.
- **AlgoID slots already reserved in algoid.go:** `AlgoQGramJaccard`, `AlgoSorensenDice`, `AlgoCosine`, `AlgoTversky` — see lines 109-128. The planner wires `dispatch_<algo>.go` files but does NOT modify algoid.go enum positions.
- **Tversky in dispatch table:** the dispatch table signature is `func(a, b string) float64`. Tversky needs `alpha`/`beta`, so the dispatch wrapper uses the default `α = β = 1.0` (reduces to Jaccard) — documented as a deliberate compromise; the real Tversky use case is via Scorer's `WithTverskyAlgorithm(weight, alpha, beta)` in Phase 8.
- **BDD scenarios:** one feature file per algorithm in `tests/bdd/features/`. Asymmetric Tversky scenario explicitly covers `α ≠ β` direction-sensitivity.
- **Staging golden → finalisation merge:** `testdata/golden/_staging/<algo>.json` during each algorithm's plan; merged into `testdata/golden/algorithms.json` in the phase-5 finalisation plan.
- **identifier-similarity example extension:** the 10-column example from Phase 4 gets four new columns (one per q-gram algorithm) during finalisation. NO new example program.

### Claude's Discretion

The planner (gsd-planner) chooses, without re-asking the user:

- Exact internal extractor signature (`extractQGrams` vs `qgramBag` vs other unexported name)
- Wave decomposition / parallelisability of plans 05-01..05-04 (the four algorithms are independent of each other; q_gram.go ships in plan 05-01 alongside Jaccard since Jaccard is the simplest consumer)
- Whether the Tversky dispatch wrapper uses `α = β = 1.0` (Jaccard fallback) or `α = β = 0.5` (Dice fallback) — recommended Jaccard fallback because Tversky+Jaccard equivalence is well-known
- Stack-buffer fast path for q-gram extraction on short ASCII inputs (likely not worth it — the map allocation dominates anyway; ASCII fast-path discipline is unchanged but the alloc budget is different from edit-distance algorithms)
- Exact number of staging-golden entries per algorithm (8-12 per Phase 2/3/4 norm)

</decisions>

<canonical_refs>
## Canonical References

**Downstream agents MUST read these before planning or implementing.**

### Spec & requirements (project-internal)

- `docs/requirements.md` §7.2 — full q-gram algorithm specifications (§7.2.1 Q-Gram Jaccard, §7.2.2 Sørensen-Dice, §7.2.3 Cosine, §7.2.4 Tversky)
- `docs/requirements.md` §13.3, §13.4, §13.6 — determinism rules (no map iteration on output, float stability, left-to-right reduction)
- `docs/requirements.md` §14.1 — per-algorithm allocation budgets
- `docs/requirements.md` §15.3, §15.4 — property test conventions, fuzz target conventions
- `.planning/REQUIREMENTS.md` — QGRAM-01..QGRAM-05 traceability table
- `.planning/ROADMAP.md` — Phase 5 goal + success criteria (lines 70-77 of phase listing)

### Primary academic sources (research targets)

- **Ukkonen, E. (1992).** "Approximate string-matching with q-grams and maximal matches." *Theoretical Computer Science*, 92(1):191–211 — Q-Gram Jaccard primary source.
- **Jaccard, P. (1912).** "The distribution of the flora in the alpine zone." *New Phytologist*, 11(2):37–50 — Jaccard coefficient origin paper.
- **Dice, L. R. (1945).** "Measures of the amount of ecologic association between species." *Ecology*, 26(3):297–302 — Sørensen-Dice coefficient primary source.
- **Sørensen, T. (1948).** "A method of establishing groups of equal amplitude in plant sociology…" *Kongelige Danske Videnskabernes Selskab*, 5(4):1–34 — independent rediscovery, same coefficient.
- **Salton, G., McGill, M. J. (1983).** *Introduction to Modern Information Retrieval*. McGraw-Hill — Cosine similarity / vector-space model textbook reference.
- **Tversky, A. (1977).** "Features of similarity." *Psychological Review*, 84(4):327–352 — Tversky asymmetric similarity primary source.

### Project skills (correctness & licensing gates)

- `.claude/skills/algorithm-correctness-standards/SKILL.md` — primary-source citation, formula docs, reference vectors
- `.claude/skills/algorithm-licensing-standards/SKILL.md` — fresh implementation discipline, attribution format
- `.claude/skills/determinism-standards/SKILL.md` — no map iteration, float stability, golden files
- `.claude/skills/performance-standards/SKILL.md` — allocation budgets, ASCII fast paths
- `.claude/skills/go-coding-standards/SKILL.md` — Go style, no testify in root tests
- `.claude/skills/go-testing-standards/SKILL.md` — unit + property + fuzz + bench + BDD coverage targets
- `.claude/skills/fuzzymatch-review-protocol/SKILL.md` — agent gate sequence

### Prior phase context (carry-forward)

- `.planning/phases/02-core-character-algorithms-six/02-CONTEXT.md` — file-by-file structure, byte+rune pattern, AlgoID dispatch (LOCKED)
- `.planning/phases/03-smith-waterman-gotoh/03-CONTEXT.md` — cross-validation discipline (Phase 3 used biopython; phase 4 + 5 downgrade to Python stdlib / hand-computed)
- `.planning/phases/04-remaining-character-gestalt/04-CONTEXT.md` — staging-golden → finalisation merge pattern, identifier-similarity example extension pattern

</canonical_refs>

<code_context>
## Existing Code Insights

### Reusable Assets

- **`algoid.go` lines 109-128** — `AlgoQGramJaccard`, `AlgoSorensenDice`, `AlgoCosine`, `AlgoTversky` enum slots are already reserved with godoc citing primary sources. Slots are stable; planner wires dispatch files but does NOT renumber.
- **`algoid.go` String() / dispatch tables** — `AlgoQGramJaccard.String()` returns `"QGramJaccard"` etc. (lines 233-240 verified). The dispatch table at lines 280+ already includes these AlgoIDs as legal entries; `dispatch_<algo>.go` files in this phase wire the actual algorithm functions.
- **`errors.go`** — `ErrInvalidQGramSize` and `ErrInvalidTverskyParam` are declared at the package level per Phase 1 (FOUND-02 wired the full error sentinel set up front).
- **`testdata/golden/algorithms.json`** — established golden file from Phase 2/3/4 (59 entries after Phase 4). Phase 5 finalisation adds entries for Cosine + the other three.
- **`testdata/golden/_staging/` workflow** — pattern proven over six algorithm staging files (jaro, jaro_winkler, smith_waterman_gotoh, strcmp95, lcsstr, ratcliff_obershelp).
- **`tests/bdd/steps/algorithms_steps.go`** — accumulator file. Each new algorithm appends a `step.Step` registration following the Phase 2/3/4 pattern.
- **`example_test.go`** — runnable godoc examples; one new `ExampleXxx` per public function (8 new examples in Phase 5 minimum: Score + ScoreRunes × 4 algorithms).
- **`props_test.go`** — appendable property-test accumulator. Phase 5 appends Five-invariant × 2-surface blocks per algorithm (RangeBounds, Identity, NoNaN, NoInf, NoNegativeZero) plus Symmetric for Jaccard/Dice/Cosine and an asymmetry property for Tversky `α ≠ β`.
- **`bench.txt`** — Phase 4 regenerated baseline (10 iterations on darwin/arm64 Apple M2, captured 2026-05-14). Phase 5 finalisation full-replaces it again to include the new benchmarks.
- **`llms.txt` / `llms-full.txt`** — every new exported symbol needs a line in `llms.txt` and a full entry in `llms-full.txt` (per-plan discipline, not deferred to finalisation — this caught Phase 4 plan 04-01 mid-flight).

### Established Patterns (Phase 5 must follow)

- **Two-row DP buffer pattern (from Phase 2 onwards)** — does NOT directly apply to q-gram algorithms (no DP table), but the underlying "stack-buffer ASCII fast path" discipline applies: the planner should evaluate whether a stack-allocated `[64]string` for short-input q-gram lists is feasible, OR document why heap allocation dominates regardless.
- **Map-iteration discipline (DET-03 from Phase 1)** — the q-gram multiset MUST NOT be iterated on any output path; the dot-product/intersection loops iterate sorted-key slices (per §3 LOCKED above).
- **Source-Origin Statement format** — every algorithm file's header has a `Sources:` block listing primary, cross-validation, and code-consulted references. WR-01 from Phase 4 (commit 65302b8) established the precedent that the block must enumerate every reference cited inline.
- **OQ-X RESOLUTION LOCKED format** — when the planner / executor make a non-obvious local decision (e.g. Tversky dispatch fallback parameters), it's recorded in the plan's SUMMARY.md as `OQ-N RESOLUTION LOCKED <date>` with rationale.
- **gsd-executor.md scope-boundary rule** — out-of-scope discoveries during execution are logged to `.planning/phases/05-q-gram-algorithms/deferred-items.md` (created on first need), NOT rolled into the current commit. Phase 4 plan 04-04 → plan 04-05 demonstrated the clean roll-in path.

### Integration Points

- **Phase 6 consumes the q-gram tier:** Monge-Elkan's `permitted inner metrics` list grows to include the four q-gram AlgoIDs; the Phase 6 plan references this CONTEXT.md when scoping the inner-metric set.
- **Phase 8 (Composite Scorer):** `WithQGramJaccardAlgorithm(weight, n int)`, `WithCosineAlgorithm(weight, n int)`, `WithTverskyAlgorithm(weight, alpha, beta float64)` options consume the four algorithms with their full parameter surface. The Scorer parameter-validation path uses `ErrInvalidQGramSize` / `ErrInvalidTverskyParam` which are already declared.
- **Phase 9 (Scan) / Phase 10 (Extract):** consume the dispatch table; q-gram algorithms participate via their AlgoID slots without scan/extract needing to know they are q-gram-flavoured.

</code_context>

<specifics>
## Specific Ideas

- **Cosine reference vectors with documented float64 precision:** 3-5 Cosine unit tests pin `cos(A, B)` to 17 significant digits with the dot product and both norms documented inline in the test comment (e.g. "`cos = 6 / (sqrt(8) * sqrt(8)) = 0.75`" so reviewers can recompute by hand). Spans `n = 2, 3, 4`, at least one Unicode pair, at least one pair with > 5 intersection elements. Per §4 LOCKED — Cosine carries the cross-validation density that would otherwise come from an external library.
- **Tversky discriminating asymmetric pair:** at least one unit test pair `(a, b)` with `α = 0.8, β = 0.2` such that `TverskyScore(a, b, n, 0.8, 0.2) ≠ TverskyScore(b, a, n, 0.8, 0.2)` — the asymmetry gate. Without this, the `α/β` parameters could be silently swapped or both halved and tests would pass.
- **Sørensen-Dice `n = 2` default:** the `SorensenDiceScore` public function signature takes `n` explicitly (no default); the `n = 2` default lives in the Phase 8 Scorer option `WithSorensenDiceAlgorithm`. The phase 5 unit tests cover `n = 2, 3` to exercise both common values.
- **Cosine `n = 3` golden entries:** the algorithms.json staging file for Cosine includes entries at `n = 2, 3, 4` to exercise the float-reduction path on multiple intersection sizes (per §1 LOCKED).

</specifics>

<deferred>
## Deferred Ideas

- **Public q-gram extraction helper (`QGrams` / `QGramsRunes`):** raised during the API-surface discussion; deferred to v1.x as additive change if real demand emerges. Promoting an unexported helper to public is a non-breaking change; the inverse is not.
- **Cosine bit-exact property test via big.Float reference:** raised during the determinism-keystone discussion; rejected as redundant — the existing CI matrix gate is the load-bearing signal. May revisit if a specific platform-specific bug surfaces in production.
- **`scripts/gen-qgram-cross-validation.py` against Python `textdistance`:** explicitly rejected per §4 LOCKED — hand-computed reference vectors from primary sources are sufficient and align with the structural simplification trend (Phase 4 already dropped biopython for stdlib `difflib`). User reconsidered mid-discussion and confirmed hand-computed is still the first choice; textdistance + 3-5 Cosine hand-derivations was logged as the fallback if reviewers surface concerns (see §4 LOCKED, Fallback paragraph).
- **n-gram size validation at the AlgoID-table level:** considered for the dispatch wrappers, but the `(a, b string) float64` dispatch signature has no place for `n` — Tversky / QGramJaccard / Cosine / Dice all use the default `n = 3` when invoked via the dispatch table. Specific `n` overrides happen via the Phase 8 Scorer option layer.

</deferred>

---

*Phase: 5-q-gram-algorithms*
*Context gathered: 2026-05-14*
