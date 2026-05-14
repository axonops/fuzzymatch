---
phase: 03-smith-waterman-gotoh
phase_number: 3
date: 2026-05-14
spec_loaded: false
prior_decisions_consulted: [02-CONTEXT.md, 02-VERIFICATION.md, all 02-*-SUMMARY.md, PITFALLS.md]
---

# Phase 3: Smith-Waterman-Gotoh — Context

<domain>
**What this phase delivers:** the **Smith-Waterman-Gotoh local alignment**
algorithm with configurable **affine gap penalty**, isolated into its own
phase because the published Gotoh 1982 recurrence contains a known
erratum (initialisation step and an indexing flip) that requires
cross-validation against an independent reference implementation. One
algorithm, one requirement (CHAR-08); the discipline gate is the
biopython cross-validation evidence committed alongside the implementation.

This phase also expands the per-algorithm public surface for SWG to
include unnormalised raw alignment scores (advanced-consumer use case
beyond the spec; see <decisions> §4).
</domain>

<canonical_refs>
Downstream agents (researcher, planner, executor) MUST read these before
writing research, plan, or implementation:

| Reference | Why |
|-----------|-----|
| `docs/requirements.md` §7.1.8 — Smith-Waterman-Gotoh | Public API names locked here, default params, score normalisation, complexity, edge cases |
| `docs/requirements.md` §14 — Performance budgets | Inherits Phase 2's ASCII-fast-path + two-row DP discipline |
| `docs/requirements.md` §15 — Test strategy | Unit + property + fuzz + benchmark + BDD per algorithm |
| `.planning/research/PITFALLS.md` §3 — Gotoh erratum | The whole reason this phase is isolated. Lists the four warning signs (identity fails, gap-split asymmetry, score asymmetry, monotonicity failure) the implementation must clear. |
| `.claude/skills/algorithm-correctness-standards/SKILL.md` | Primary source citation, formula docs, mathematical-invariant property tests |
| `.claude/skills/algorithm-licensing-standards/SKILL.md` | Fresh-impl rule, no GPL/LGPL consultation. biopython is BSD-3-Clause (compatible) — used for cross-validation reference vectors only, NOT for code copying. |
| `.claude/skills/performance-standards/SKILL.md` | Two-row DP requirement, ASCII fast paths, benchstat regression |
| `.claude/skills/determinism-standards/SKILL.md` | Float-stability rules (no `math.Pow`, no `math.FMA`), no map iteration on output paths |
| `.claude/skills/go-testing-standards/SKILL.md` | Coverage floors (>= 90% per-file, >= 95% overall), property test conventions |
| `.claude/skills/fuzzymatch-review-protocol/SKILL.md` | Phase-end review gates |
| `.planning/phases/02-core-character-algorithms-six/02-CONTEXT.md` | Inherited file pattern (`<algo>.go` + `dispatch_<algo>.go` + `<algo>_test.go` + `<algo>_bench_test.go` + `<algo>_fuzz_test.go` + staging golden + BDD feature) |
| `.planning/phases/02-core-character-algorithms-six/02-PATTERNS.md` | Pattern 3 (ASCII fast-path gate LOCKED): `if shortDim <= 64 && isASCII(a) && isASCII(b)` — applies to SWG identically |
| `.planning/phases/02-core-character-algorithms-six/02-01-levenshtein-SUMMARY.md` | Canonical Wave 1 pattern Phase 3 replicates one-for-one |
| `.planning/phases/02-core-character-algorithms-six/02-VERIFICATION.md` §Re-verification (cleanup pass) | Closes WR-01..WR-04, IN-01..IN-07 — Phase 3 inherits the cleaned-up patterns directly |
| `.planning/phases/02-core-character-algorithms-six/02-07-finalisation-SUMMARY.md` | Cross-algorithm consistency test + bench.txt baseline pattern — Phase 3 extends both |
| `algoid.go` line 97-102 (existing) | `AlgoSmithWatermanGotoh` already declared at slot 6 of 23 |
| `dispatch_levenshtein.go` (existing) | Dispatch registration template — copy for `dispatch_swg.go` |
| `levenshtein.go`, `damerau_full.go` (existing) | DP kernel patterns — SWG kernel differs but the file structure, godoc shape, header, and ASCII fast-path gate match |
| `props_test.go` (existing, extend-only) | Property test conventions — SWG appends its own `TestProp_SWG*` block |
| `tests/bdd/steps/algorithms_steps.go` (existing, extend-only) | BDD step bindings — SWG appends |
| `testdata/golden/algorithms.json` (existing) | Phase 3 adds SWG entries via the staging-merge pattern locked in Phase 2 |
</canonical_refs>

<code_context>
**Phase 1 + Phase 2 outputs that Phase 3 must compose with:**

- `algoid.go` — `AlgoSmithWatermanGotoh` already declared at index 6 (line 102).
  Phase 3 populates `dispatch[AlgoSmithWatermanGotoh]` via the var-init pattern
  in `dispatch_swg.go` (NO `init()` functions per §5.12).
- `errors.go` — `ErrInvalidInput`, `ErrInvalidConfiguration` available, but
  SWG's `*Score` functions do not return errors (per §5.11 — algorithm score
  functions handle every edge case deterministically). No validation in
  SWG either (see <decisions> §3).
- `normalise.go` / `tokenise.go` — not used directly; SWG takes pre-normalised
  input from callers.
- `golden_canonical.go` — `CanonicalMarshalForTest` is the only legal
  marshaller for golden files; SWG entries go through it via
  `testdata/golden/_staging/swg.json` and then plan 03-finalisation merges
  into the canonical `testdata/golden/algorithms.json`.
- `cross_algorithm_consistency_test.go` (existing) — Phase 3 extends it with
  one new SWG-vs-Levenshtein test pinning the local-alignment-vs-edit-distance
  divergence on a substring-containment input (e.g. SWG on
  `http_request` vs `http_request_header_fields` should score higher than
  Levenshtein on the same pair — SWG finds the substring; Levenshtein counts
  the full edit distance).
- `bench.txt` (existing) — Phase 3 appends six benchmark rows
  (`BenchmarkSmithWatermanGotohScore_{ASCII_Short,Medium,Long,Unicode_Short}`
  + `_WithParams` + `_RawScore`).
- `examples/identifier-similarity/main.go` (existing) — Phase 3 adds an SWG
  column to the 7-row × 6-column table → 7-row × 7-column. Bench-relevant
  `TestExample_Output` and `TestExample_ColumnWidths` adjust accordingly.
- `tests/bdd/steps/algorithms_steps.go` — extend-only append pattern;
  add `iComputeTheSmithWatermanGotohScoreBetween` + raw / params variants.
- `tests/bdd/features/` — add `swg.feature` Gherkin file (one feature with
  multiple scenarios; pattern matches Phase 2's per-algorithm feature files).

**New cross-validation tooling (NOT runtime code):**

- `scripts/gen-swg-cross-validation.py` — Python script using
  `Bio.Align.PairwiseAligner` (biopython, BSD-3-Clause) to produce reference
  vectors. Runs at developer-discretion via `make regen-swg-cross-validation`.
  Output: `testdata/cross-validation/swg/vectors.json`.
- `testdata/cross-validation/swg/vectors.json` — committed JSON corpus,
  ~10-20 entries spanning identity / both-empty / one-empty / two-substring /
  no-overlap / one-long-gap (the gap case is the Gotoh-erratum canary per
  PITFALLS.md §3) / non-default params. Re-running `make verify-determinism`
  or `make check` validates against this corpus with 1e-9 epsilon tolerance.
- `TestSWG_CrossValidation` in `swg_test.go` — reads the JSON, asserts
  `|our_score - biopython_score| <= 1e-9` for every entry. No Python runtime
  required at test time.
- Makefile: new `regen-swg-cross-validation` target (developer-only) +
  `TestSWG_CrossValidation` runs via default `go test ./...` (no opt-in
  required for verification).
</code_context>

<decisions>

## §1. Cross-validation evidence path — LOCKED

**Tool: biopython `Bio.Align.PairwiseAligner`** (NOT the deprecated `pairwise2.align.localxs` mentioned in the spec; PairwiseAligner is biopython's actively-supported replacement since v1.79).

- BSD-3-Clause licensed → permissive, compatible with Apache-2.0 for reference-vector cross-validation per `algorithm-licensing-standards`.
- Pip-installable: `pip install biopython`. Implementer runs locally to regenerate vectors; CI does not require Python.
- Newer than EMBOSS for the SWG implementation; EMBOSS is deferred ("if a future regression is suspected, also verify against EMBOSS" — record as a deferred follow-up if needed).

**Evidence location: committed JSON + generator script.**

- `scripts/gen-swg-cross-validation.py` — Python script, reads a fixed list of test cases (inline in the script), invokes biopython's PairwiseAligner with the params set, writes
  `testdata/cross-validation/swg/vectors.json`.
- Schema: `{"version": 1, "biopython_version": "<x.y.z>", "entries": [{"name": "<id>", "a": "<str>", "b": "<str>", "params": {"match": 1.0, ...}, "biopython_score": 0.123456789}]}`.
- ~10-20 entries minimum. MUST include: identity, both-empty, one-empty, two-substring (e.g. `http_request` / `http_request_header_fields`), no-overlap, **one-long-gap-canary** (specifically exercises Gotoh's initialisation; PITFALLS.md §3), and one non-default-params case.

**Gate: default `go test ./...` includes `TestSWG_CrossValidation`.**

- Test reads the committed JSON, no Python at test time.
- Comparison tolerance: `|our_score - biopython_score| <= 1e-9` (matches `cross_algorithm_consistency_test.go` epsilon convention).
- Regeneration is a separate Makefile target `make regen-swg-cross-validation` (developer-only); the verification target runs in `make check`.

**Score normalisation alignment:** the biopython "score" is the raw alignment score. To compare, EITHER (a) the script emits both raw and normalised biopython scores and the test compares against the normalised one, OR (b) the test normalises biopython's raw score with the same `clamp(best / min(len), 0, 1)` formula our implementation uses. **Choose (a) — biopython's raw + script-computed normalised** so the test has zero implementation logic and the script owns the normalisation reference. Captures both gates simultaneously.

## §2. DP space-optimisation policy — LOCKED

**Two-row optimised from day 1 — full O(min(m,n)) space.**

- Three two-row buffer pairs: `prevM, currM, prevIx, currIx, prevIy, currIy`, each of length `n+1`.
- Stack-allocated when `n <= maxStackInputLen && isASCII(a) && isASCII(b)`:
  - Buffer: `var buf [(maxStackInputLen+1) * 6]float64` = 3120 bytes (65 × 6 × 8). Sized for `n+1=65` per row × 6 rows.
  - Heap-allocated otherwise: six `make([]float64, n+1)` calls. Documented expected alloc count on heap path (6 on ASCII Long).
- Rationale: performance-standards skill explicitly mandates two-row DP. Phase 2's DL-Full deferral to v1.x (PERF-03 / issue #1) was justified by DL-Full's auxiliary `da` last-occurrence map (a separate complication). SWG has no auxiliary state of comparable complexity; the three-matrix two-row form is a clean stack-allocation candidate. Insisting on day 1 also avoids accruing more deferred technical debt at the Phase 2 boundary.
- Risk-mitigation: biopython cross-validation (§1) catches any DP transcription bug regardless of whether the implementation is paper-faithful three-matrix or space-optimised two-row.

**Allocation budget: matches Phase 2 PERF-01 contract.**

- 0 B/op, 0 allocs/op on ASCII Short (≤ 64 bytes shorter dim).
- Heap path on ASCII Long (proportional, 6 allocs for the six rows).
- Rune path: 2 allocs for `[]rune(a)` and `[]rune(b)`; six more for the buffer rows; rune-aware DP kernel mirrors the byte kernel.
- Bench labels mirror Phase 2: `_ASCII_Short`, `_ASCII_Medium`, `_ASCII_Long`, `_Unicode_Short`. ALSO add `_WithParams_ASCII_Short` and `_RawScore_ASCII_Short` to exercise the params and raw paths separately.

**ASCII gate: identical idiom to Phase 2.**

- `if n <= maxStackInputLen && isASCII(a) && isASCII(b) { /* stack path */ }`
- Documented in 02-PATTERNS.md as the locked v1.0 idiom for DP-based algorithms.

## §3. SWGParams API ergonomics — LOCKED

**Construction: `NewSWGParams()` constructor returning defaults.**

```go
// SWGParams holds the affine-gap parameters for Smith-Waterman-Gotoh.
// All fields are exported and read after construction; SWGParams is a
// value type (no pointer receivers required).
type SWGParams struct {
    Match     float64 // reward for a position match, >= 0; default 1.0
    Mismatch  float64 // penalty for a position mismatch, <= 0; default -1.0
    GapOpen   float64 // penalty for opening a new gap, <= 0; default -1.5
    GapExtend float64 // penalty for extending an existing gap, <= 0; default -0.5
                     // Conventionally: GapOpen <= GapExtend <= 0 (extending is cheaper than opening).
}

// NewSWGParams returns SWGParams populated with the documented defaults
// (Match=1.0, Mismatch=-1.0, GapOpen=-1.5, GapExtend=-0.5). Callers can
// override individual fields after construction:
//
//   params := NewSWGParams()
//   params.Match = 2.0
//   score := SmithWatermanGotohScoreWithParams(a, b, params)
//
// The returned value is a fresh copy; callers can mutate freely.
func NewSWGParams() SWGParams { ... }
```

- Naked struct WITH exported fields (Go idiom; allows direct field access).
- `NewSWGParams()` constructor returns a new value with defaults — no shared mutable state, no package-level vars.
- NO `SWGDefaultParams` exported var (avoids the "is this read-only?" footgun; callers use `NewSWGParams()` instead).

**Validation: none in *Score functions; document in godoc.**

- Per §5.11 ("Errors only where genuinely necessary"), algorithm score functions never return errors and never panic on caller-supplied params.
- Caller passes nonsense (e.g. `GapOpen > 0`, NaN, +Inf)? Algorithm still produces a deterministic score; it might be meaningless but is reproducible.
- Godoc on `SWGParams` documents expected ranges: `Match >= 0`, `Mismatch <= 0`, `GapOpen <= GapExtend <= 0`.
- The Scorer layer (Phase 8) MAY add validation at composition time; that's a Phase 8 decision, not Phase 3's.

**Delegation: no-params form calls `*WithParams` with `NewSWGParams()`.**

```go
func SmithWatermanGotohScore(a, b string) float64 {
    return SmithWatermanGotohScoreWithParams(a, b, NewSWGParams())
}
```

- Cleanest factoring; the params kernel is the single source of truth.
- Default-construction cost is negligible (one stack-allocated struct).
- Defaults live in exactly one place (NewSWGParams body); no risk of drift between paths.

## §4. Raw alignment score exposure — DECIDED (scope expansion beyond spec)

**Add three new public functions** for unnormalised, unclamped raw alignment scores:

```go
func SmithWatermanGotohRawScore(a, b string) float64
func SmithWatermanGotohRawScoreRunes(a, b string) float64
func SmithWatermanGotohRawScoreWithParams(a, b string, params SWGParams) float64
```

**Rationale:**
- Advanced consumers (bioinformatics, schema-similarity research) want absolute alignment quality unaffected by the normalisation choice.
- The clamped normalised score `clamp(best / min(len), 0, 1)` discards information when the raw is negative (two unrelated strings) or exceeds `min(len)` (high-reward custom params).
- The DP kernel computes the raw score regardless; exposing it costs zero implementation complexity (one extra return path per public function).

**Scope-creep note:** this expands SWG's public surface from 3 to 6 functions, beyond what `docs/requirements.md` §7.1.8 specifies. This is a deliberate decision by the project owner (user, 2026-05-14). Phase 3's deliverables include:

1. Updating `docs/requirements.md` §7.1.8 to list all 6 functions (the Phase 3 SUMMARY records this requirements-doc edit).
2. Flagging the new surface in the `api-ergonomics-reviewer` review during code review.
3. Documenting in godoc that Raw* functions return unclamped raw alignment scores: "may be negative or > 1; use SmithWatermanGotohScore* for the normalised [0,1] similarity."

**Godoc clamp warning on the normalised functions:**

```go
// SmithWatermanGotohScore returns the Smith-Waterman-Gotoh local-alignment
// similarity between a and b as a value in [0.0, 1.0]. The returned score is
// CLAMPED: if the underlying alignment score is negative (e.g. two unrelated
// strings dominated by mismatch/gap penalties) the clamp returns 0.0; if it
// exceeds min(len(a), len(b)) (custom params with Match > 1.0) the clamp
// returns 1.0. Use SmithWatermanGotohRawScore for the unclamped raw
// alignment score.
//
// score = clamp(best_local_score / min(len(a), len(b)), 0.0, 1.0)
//
// Edge cases:
//   - SmithWatermanGotohScore("", "") == 1.0 (both-empty identity)
//   - SmithWatermanGotohScore("", "abc") == 0.0 (one-empty)
//   - SmithWatermanGotohScore(x, x) == 1.0 for any non-empty x
//   - SmithWatermanGotohScore(a, b) == SmithWatermanGotohScore(b, a) (symmetric)
```

## §5. Property tests — INHERIT FROM PHASE 2 + 1 SWG-SPECIFIC

**Standard Phase 2 invariants** (per props_test.go pattern):
- `TestProp_SmithWatermanGotohScore_RangeBounds` — `[0,1]` for all inputs
- `TestProp_SmithWatermanGotohScore_Identity` — `Score(x, x) == 1.0` for non-empty x
- `TestProp_SmithWatermanGotohScore_Symmetric` — `Score(a, b) == Score(b, a)` for byte path
- `TestProp_SmithWatermanGotohScoreRunes_Symmetric` — rune path (per WR-03 cleanup pattern)
- `TestProp_SmithWatermanGotohScore_NoNaN`, `_NoInf`, `_NoNegativeZero` — DET-04
- NO triangle inequality test (SWG is not a metric — local alignment doesn't define a distance over the full string space)

**SWG-specific invariants** (per PITFALLS.md §3 "warning signs"):
- `TestProp_SmithWatermanGotoh_GapSplitInvariance` — for an input pair containing one long gap, splitting the gap into two halves (with intervening match characters that don't affect the local alignment) should NOT improve the score. This is the canonical Gotoh-erratum canary. Implementation: hand-curated triples that exercise the case.
- `TestProp_SmithWatermanGotoh_RawNeverExceedsMatchTimesMinLen` — `RawScore <= Match * min(len(a), len(b))` always (upper bound from "best local alignment has at most min(len) match positions").
- `TestProp_SmithWatermanGotoh_MonotonicWithMatchReward` — increasing the Match parameter (keeping others fixed) cannot decrease `RawScore` for any input.

## §6. Plan decomposition — single Wave

**Structure:** 1-3 plans in a single wave, no parallelism (one algorithm, one set of shared-file extensions to props_test.go / example_test.go / algoid_test.go / tests/bdd/steps/algorithms_steps.go). Files are append-only as in Phase 2.

**Tentative plan boundaries (planner refines):**

1. **`03-01-swg-implementation`** — `swg.go` (or `smith_waterman_gotoh.go` — planner picks the cleaner name), `dispatch_swg.go`, unit tests with primary-source-derived vectors, property tests, fuzz, benchmark, BDD feature, staging golden.

2. **`03-02-swg-cross-validation`** — `scripts/gen-swg-cross-validation.py`, `testdata/cross-validation/swg/vectors.json`, `TestSWG_CrossValidation`, Makefile target. Could be folded into 03-01 if the planner prefers; the user is fine with either.

3. **`03-03-swg-finalisation`** — merge SWG staging golden into `testdata/golden/algorithms.json`, extend `cross_algorithm_consistency_test.go` with one SWG-vs-Levenshtein divergence test (e.g. substring containment input where SWG scores higher), extend `examples/identifier-similarity/main.go` to add an SWG column (7 → 8 columns; update `want` constant and `TestExample_ColumnWidths`), update `bench.txt`, update `llms.txt`, update `docs/requirements.md` §7.1.8 to record the Raw* surface expansion, file SUMMARY.

Planner can collapse 2+3 into 03-01-and-finalisation if the total task count stays manageable.

## §7. Inherited Phase 2 patterns (NOT decided here — read 02-CONTEXT.md and 02-PATTERNS.md)

- File-by-file structure: `swg.go`, `dispatch_swg.go`, `swg_test.go`, `swg_bench_test.go`, `swg_fuzz_test.go`, `tests/bdd/features/swg.feature`, `testdata/golden/_staging/swg.json`, `testdata/fuzz/FuzzSmithWatermanGotohScore/seed-001`
- Apache-2.0 header on every `.go` file
- Stdlib testing only in root tests (no testify; testify allowed in tests/bdd)
- AlgoID dispatch via `var _ = func() bool { dispatch[AlgoSmithWatermanGotoh] = SmithWatermanGotohScore; return true }()`
- Coverage floors: ≥ 95% overall, ≥ 90% per file, 100% on public API surface
- llms.txt sync — every exported symbol (the 6 SWG functions + SWGParams + NewSWGParams + AlgoSmithWatermanGotoh) listed
- Identity short-circuit on `*Runes` entries (per IN-02 cleanup): `if a == b { return 1.0 }` saves the `[]rune` conversions on the identity path
- BDD score regex relaxed `(\d+\.?\d*)` accepts integer-form (per IN-03 cleanup)

</decisions>

<deferred>
**Out-of-scope for Phase 3 — captured for later:**

- **EMBOSS `water` second cross-validation source.** Decided to defer; biopython alone is sufficient for v1.0. Re-evaluate if a future regression is suspected. Track as a roadmap-backlog item if it becomes load-bearing.
- **`SmithWatermanGotohAlignment(a, b) Alignment` returning the actual alignment trace (matched positions, gaps).** Out of scope per spec; could be a future v1.x addition if consumer demand surfaces.
- **A `SmithWatermanGotohDistance` function** (analogous to Levenshtein/Hamming "Distance" variants). SWG doesn't define a metric distance; the raw alignment score is what the bioinformatics community refers to as "the SWG score". The Raw* functions cover this need.
- **CI installation of biopython for re-verification.** The committed JSON is the verification fixture; regeneration is a developer-discretion operation. If we later want CI to re-verify the JSON-vs-biopython agreement (e.g. after a biopython version bump), add a GitHub Actions workflow then.
- **Public API freeze for `SWGParams`** — fields are part of the v1.0 public surface; adding fields in a minor release is a breaking change. Document this in 03-CONTEXT.md? Already captured here; the planner ensures docs/requirements.md reflects it.
</deferred>

<follow_ups>
**Phase 4+ tracking items surfaced during Phase 3 discussion:**

- **Update `docs/requirements.md` §7.1.8** to list all 6 SWG public functions (3 normalised + 3 Raw) once Phase 3 ships. This is a Phase 3 deliverable per §4.
- **api-ergonomics-reviewer** should flag the Raw* surface during Phase 3 PR review — the expansion is intentional but the reviewer should confirm naming + symmetry with the normalised set.
- **algorithm-correctness-reviewer** gate: cross-validation JSON must be present + green before the SWG implementation can land. Plan-phase locks this into the plan check.
</follow_ups>

<carry_forward>
**From Phase 2 (inherited automatically — no need to re-decide):**

| Decision | Source | Applies to Phase 3 |
|----------|--------|-------------------|
| AlgoID dispatch via `var _ = func() bool {...}()` (no init() functions) | 02-CONTEXT.md §dispatch | `dispatch_swg.go` |
| `maxStackInputLen = 64` + `isASCII(a) && isASCII(b)` gate | 02-CONTEXT.md §performance, 02-PATTERNS.md Pattern 3 (locked) | `swg.go` stack-buffer path |
| Per-algorithm staging golden → canonical merge in finalisation plan | 02-CONTEXT.md §golden | `testdata/golden/_staging/swg.json` → merged into `algorithms.json` |
| Per-algorithm BDD feature file | 02-CONTEXT.md §BDD | `tests/bdd/features/swg.feature` |
| Per-algorithm fuzz harness with seed-001 entry | 02-CONTEXT.md §fuzz | `swg_fuzz_test.go` + `testdata/fuzz/FuzzSmithWatermanGotohScore/seed-001` |
| props_test.go / example_test.go / algoid_test.go / tests/bdd/steps/algorithms_steps.go are extend-only append points | 02-CONTEXT.md §shared-files | SWG appends its block to each |
| Stdlib testing only in root (no testify; testify allowed in tests/bdd) | 02-CONTEXT.md §test-stack | `swg_test.go` uses stdlib `testing` only |
| Identity short-circuit `if a == b { return 1.0 }` on `*Runes` entries | IN-02 cleanup commit c235e0e | `SmithWatermanGotohScoreRunes`, `SmithWatermanGotohRawScoreRunes` |
| BDD score regex `(\d+\.?\d*)` accepts integer-form | IN-03 cleanup commit 8802d0b | `swg.feature` scenarios can use `0` and `1` |
| `theDistanceShouldBe` BDD step is algorithm-agnostic by design (no SWG distance variant needed) | IN-06 cleanup commit 8802d0b | No new "distance" step for SWG |
| Cross-platform CI matrix verifies byte-identical output (DET-02 lock) | Phase 1 verification | `testdata/golden/algorithms.json` SWG entries must be byte-stable |
| Coverage floors ≥ 95% overall, ≥ 90% per file | go-testing-standards skill | `swg.go` and `swg_*_test.go` |
| llms.txt is the AI-friendly catalogue; meta-test asserts every exported symbol is listed | Phase 1 + Phase 2 | Add SWG entries (6 funcs + SWGParams + NewSWGParams + AlgoSmithWatermanGotoh = 9 lines) |
</carry_forward>
</content>
</invoke>