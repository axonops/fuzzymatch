# Phase 3: Smith-Waterman-Gotoh — Research

**Researched:** 2026-05-14
**Domain:** Local sequence alignment with affine gap penalty (Smith-Waterman with Gotoh's 1982 improvement); Go implementation, biopython cross-validation harness
**Confidence:** HIGH on stack/patterns/pitfalls (all inherited from Phase 2 + locked in CONTEXT.md); HIGH on Gotoh erratum (peer-reviewed source identified); HIGH on biopython API (current docs verified)

## Summary

Phase 3 implements **one** algorithm — Smith-Waterman-Gotoh local alignment with configurable affine gap penalty — but with an unusual gating discipline: the published Gotoh 1982 recurrence contains two documented mathematical errors (an indexing mistake and a more impactful initialisation issue) that have propagated through textbooks and reference implementations. Per `.planning/research/PITFALLS.md` §3, a biorxiv survey by Flouri et al. (2015) found **5 of 10 inspected source codes incorrect, 8 of 31 lecture slides incorrect, and 16 of 31 incomplete on initialisation**. The Phase 3 discipline therefore is: primary-source-only is not enough — the implementation MUST cross-validate against an independent permissively-licensed reference (decision in `03-CONTEXT.md` §1: biopython's `Bio.Align.PairwiseAligner`, BSD-3-Clause).

Beyond that gating, Phase 3 inherits Phase 2's locked patterns one-for-one: the canonical algorithm-file layout (`swg.go` + `dispatch_swg.go` + per-algorithm unit / bench / fuzz / golden-staging / BDD files), the ASCII-fast-path gate `n <= maxStackInputLen && isASCII(a) && isASCII(b)`, the `var _ = func() bool { ... }()` dispatch-registration idiom, stdlib `testing` only in root, append-only updates to shared files (`props_test.go`, `example_test.go`, `algorithms_golden_test.go`, `tests/bdd/steps/algorithms_steps.go`, `cross_algorithm_consistency_test.go`, `bench.txt`, `llms.txt`). The CONTEXT.md decision to ship two-row DP from day 1 (full O(min(m,n)) space with six rolling rows) and to expose three additional public functions (`SmithWatermanGotohRawScore` / `*Runes` / `*WithParams`) is sound — the three-matrix form (M, Ix, Iy) transcribes cleanly into pairs of rolling rows because every cell depends only on `[i-1][j-1]`, `[i-1][j]`, and `[i][j-1]`, with no diagonal-back-two or aux-table dependencies that complicate space optimisation.

**Primary recommendation:** Implement the three-matrix M/Ix/Iy formulation with `prevM`/`currM`/`prevIx`/`currIx`/`prevIy`/`currIy` rolling rows (six `float64` rows of length `n+1`, stack-allocated via a single `[(maxStackInputLen+1) * 6]float64` buffer when `n <= maxStackInputLen && isASCII(a) && isASCII(b)`), using only `+ - * / max` and `float64()` conversion (no `math.*` calls). Cite Gotoh 1982 as primary AND Flouri et al. 2015 biorxiv as the corrected-formulation source in the file's block comment, with the erratum and corrected initialisation explicitly named. Generate the cross-validation JSON corpus via `scripts/gen-swg-cross-validation.py` (biopython `PairwiseAligner` with `mode="local"`, `match_score=Match`, `mismatch_score=Mismatch`, `open_gap_score=GapOpen`, `extend_gap_score=GapExtend`, calling `aligner.score(a, b)`); the Python script ALSO computes the normalised reference `clamp(raw / min(len(a), len(b)), 0, 1)` so the Go test has zero implementation logic to drift from the reference.

<user_constraints>
## User Constraints (from CONTEXT.md)

### Locked Decisions

**§1. Cross-validation evidence path — LOCKED.**
- Tool: biopython `Bio.Align.PairwiseAligner` (NOT the deprecated `pairwise2.align.localxs`; PairwiseAligner is biopython's actively-supported replacement since v1.79). BSD-3-Clause licensed → permissive, compatible with Apache-2.0 for reference-vector cross-validation per `algorithm-licensing-standards`. Pip-installable; implementer runs locally to regenerate vectors; CI does not require Python.
- Evidence location: committed JSON + generator script.
  - `scripts/gen-swg-cross-validation.py` — Python script with inline fixed list of test cases; invokes biopython's PairwiseAligner; writes `testdata/cross-validation/swg/vectors.json`.
  - JSON schema: `{"version": 1, "biopython_version": "<x.y.z>", "entries": [{"name": "<id>", "a": "<str>", "b": "<str>", "params": {"match": 1.0, ...}, "biopython_score": 0.123456789, "biopython_normalised": 0.456789}]}`.
  - ~10-20 entries minimum. MUST include: identity, both-empty, one-empty, two-substring (e.g. `http_request` / `http_request_header_fields`), no-overlap, **one-long-gap-canary** (specifically exercises Gotoh's initialisation per PITFALLS.md §3), and one non-default-params case.
- Gate: default `go test ./...` includes `TestSWG_CrossValidation` reading the committed JSON — no Python at test time. Tolerance: `|our_score - biopython_score| <= 1e-9` (matches `cross_algorithm_consistency_test.go` epsilon).
- Regeneration: separate Makefile target `make regen-swg-cross-validation` (developer-only); verification runs in `make check`.
- Normalisation alignment: script emits BOTH raw biopython score AND script-computed normalised reference; Go test compares against the normalised value with zero in-Go normalisation logic.

**§2. DP space-optimisation policy — LOCKED.**
- Two-row optimised from day 1 — full O(min(m,n)) space. Three two-row buffer pairs: `prevM, currM, prevIx, currIx, prevIy, currIy`, each of length `n+1`.
- Stack-allocated when `n <= maxStackInputLen && isASCII(a) && isASCII(b)`: single `var buf [(maxStackInputLen+1) * 6]float64` = 3120 bytes (65 × 6 × 8). Heap-allocated otherwise: six `make([]float64, n+1)` calls (documented expected alloc count 6 on heap path).
- Rationale: performance-standards skill explicitly mandates two-row DP. SWG has no auxiliary state of comparable complexity to DL-Full's `da` map (which is why DL-Full was deferred to v1.x — issue #1). The three-matrix two-row form is a clean stack-allocation candidate; insisting on day 1 avoids accruing more deferred technical debt.
- Allocation budget: matches Phase 2 PERF-01 contract. 0 B/op, 0 allocs/op on ASCII Short (≤ 64 bytes shorter dim). Heap path ASCII Long: proportional, 6 allocs. Rune path: 2 + 6 = 8 allocs (two `[]rune` + six row slices).
- Bench labels mirror Phase 2: `_ASCII_Short`, `_ASCII_Medium`, `_ASCII_Long`, `_Unicode_Short`. ALSO add `_WithParams_ASCII_Short` and `_RawScore_ASCII_Short`.
- ASCII gate: identical idiom to Phase 2 — `if n <= maxStackInputLen && isASCII(a) && isASCII(b) { /* stack path */ }`.

**§3. SWGParams API ergonomics — LOCKED.**
- `SWGParams` is a value struct with exported fields `{Match, Mismatch, GapOpen, GapExtend float64}` (Go idiom; allows direct field access).
- `NewSWGParams()` constructor returns a new value with documented defaults (Match=1.0, Mismatch=-1.0, GapOpen=-1.5, GapExtend=-0.5). No package-level `SWGDefaultParams` var (avoids the "is this read-only?" footgun).
- Conventional invariants documented in godoc: `Match >= 0`, `Mismatch <= 0`, `GapOpen <= GapExtend <= 0` (extending cheaper than opening).
- NO validation in `*Score` functions; nonsense params still produce a deterministic score (per §5.11 — algorithm score functions never return errors / never panic on caller-supplied params).
- Delegation: no-params form calls `*WithParams` with `NewSWGParams()` — params kernel is the single source of truth.

**§4. Raw alignment score exposure — DECIDED (scope expansion beyond spec).**
- Three new public functions: `SmithWatermanGotohRawScore(a, b string) float64`, `SmithWatermanGotohRawScoreRunes(a, b string) float64`, `SmithWatermanGotohRawScoreWithParams(a, b string, params SWGParams) float64`.
- Returns unclamped raw alignment scores (may be negative or > 1).
- Rationale: advanced consumers (bioinformatics, schema-similarity) want absolute alignment quality unaffected by normalisation. DP kernel computes raw regardless; exposing it costs zero implementation complexity.
- SCOPE NOTE: this expands SWG's public surface from 3 to 6 functions, beyond `docs/requirements.md` §7.1.8. Phase 3 deliverables include: (a) updating `docs/requirements.md` §7.1.8 to list all 6 functions; (b) flagging the new surface in the `api-ergonomics-reviewer` review; (c) godoc documents that `Raw*` returns unclamped scores and directs to `*Score` for `[0,1]` similarity; (d) godoc on normalised functions includes the clamp warning text from CONTEXT.md §4.

**§5. Property tests — INHERIT FROM PHASE 2 + 1 SWG-SPECIFIC.**

Standard Phase 2 invariants (per `props_test.go` pattern):
- `TestProp_SmithWatermanGotohScore_RangeBounds` — `[0,1]` for all inputs
- `TestProp_SmithWatermanGotohScore_Identity` — `Score(x, x) == 1.0` for non-empty x
- `TestProp_SmithWatermanGotohScore_Symmetric` — `Score(a, b) == Score(b, a)` for byte path
- `TestProp_SmithWatermanGotohScoreRunes_Symmetric` — rune path (per WR-03 cleanup pattern from Phase 2 verification)
- `TestProp_SmithWatermanGotohScore_NoNaN`, `_NoInf`, `_NoNegativeZero` — DET-04
- NO triangle inequality test (SWG is not a metric — local alignment doesn't define a distance over the full string space).

SWG-specific invariants (per PITFALLS.md §3 "warning signs"):
- `TestProp_SmithWatermanGotoh_GapSplitInvariance` — gap-split canary, hand-curated triples that exercise the canonical Gotoh-erratum case.
- `TestProp_SmithWatermanGotoh_RawNeverExceedsMatchTimesMinLen` — `RawScore <= Match * min(len(a), len(b))`.
- `TestProp_SmithWatermanGotoh_MonotonicWithMatchReward` — increasing the `Match` parameter (others fixed) cannot decrease `RawScore` for any input.

**§6. Plan decomposition — single Wave.**

Tentative plan boundaries (planner refines):
1. **`03-01-swg-implementation`** — `swg.go`, `dispatch_swg.go`, unit tests with primary-source-derived vectors, property tests, fuzz, benchmark, BDD feature, staging golden.
2. **`03-02-swg-cross-validation`** — `scripts/gen-swg-cross-validation.py`, `testdata/cross-validation/swg/vectors.json`, `TestSWG_CrossValidation`, Makefile target. May fold into 03-01 at planner's discretion.
3. **`03-03-swg-finalisation`** — merge SWG staging golden into `testdata/golden/algorithms.json`, extend `cross_algorithm_consistency_test.go` with one SWG-vs-Levenshtein divergence test (e.g. substring containment), extend `examples/identifier-similarity/main.go` to add an SWG column (7→8 columns; update `want` constant and `TestExample_ColumnWidths`), update `bench.txt`, `llms.txt`, and `docs/requirements.md` §7.1.8, file SUMMARY.

Planner can collapse 2+3 into 03-01-and-finalisation if total task count stays manageable.

**§7. Inherited Phase 2 patterns** — see `<carry_forward>` of CONTEXT.md for the full list (AlgoID dispatch, maxStackInputLen=64, per-algorithm staging golden, BDD per-algorithm feature, fuzz per-algorithm, append-only shared files, no testify in root, identity short-circuit on `*Runes`, BDD score regex `(\d+\.?\d*)`, etc.).

### Claude's Discretion

(none flagged in CONTEXT.md beyond `<open_questions>` from Phase 2; the planner picks plan-decomposition granularity per §6 and the implementation file name — `swg.go` vs `smith_waterman_gotoh.go` — at its discretion).

### Deferred Ideas (OUT OF SCOPE)

- **EMBOSS `water` second cross-validation source.** biopython alone is sufficient for v1.0. Re-evaluate if a future regression is suspected.
- **`SmithWatermanGotohAlignment(a, b) Alignment` returning the actual alignment trace.** Out of scope per spec; possible v1.x.
- **`SmithWatermanGotohDistance` function.** SWG doesn't define a metric distance; the `Raw*` functions cover the unnormalised-quality need.
- **CI installation of biopython for re-verification.** Committed JSON is the verification fixture; regeneration is developer-discretion.
- **Public API freeze for `SWGParams`.** Fields are part of v1.0 surface; adding fields in a minor release is a breaking change. Captured in this CONTEXT.md.
</user_constraints>

<phase_requirements>
## Phase Requirements

| ID | Description | Research Support |
|----|-------------|------------------|
| CHAR-08 | Smith-Waterman-Gotoh with configurable affine gap penalty — implementation MUST cross-validate against EMBOSS or biopython reference vectors due to documented Gotoh 1982 erratum (`docs/requirements.md` §7.1.8, research/PITFALLS.md #3) | Erratum source identified (Flouri et al. 2015 biorxiv `[CITED: biorxiv.org/content/10.1101/031500v1]`); biopython `PairwiseAligner` API verified `[CITED: biopython.org/docs/latest/Tutorial/chapter_pairwise.html]`; three-matrix → two-row DP transcription analysis below; six-function public surface from CONTEXT.md §3+§4; six property tests (incl. three SWG-specific canaries) from CONTEXT.md §5 + PITFALLS.md §3 warning signs. |
</phase_requirements>

## Architectural Responsibility Map

| Capability | Primary Tier | Secondary Tier | Rationale |
|------------|-------------|----------------|-----------|
| Affine-gap DP kernel (M/Ix/Iy three-matrix recurrence, two-row form) | `swg.go` algorithm file (root package `fuzzymatch`) | — | Algorithm functions are Layer 1 in fuzzymatch's three-layer architecture. Pure function, no Scorer / scan dependency. |
| ASCII vs rune dispatch | `swg.go` (byte path on string indexing; rune path on `[]rune` slices) | — | Mirrors Phase 2 — every byte-level algorithm exposes a `*Runes` rune-aware twin. |
| Dispatch-table registration | `dispatch_swg.go` | — | Per-algorithm dispatch file is the Phase 2 locked pattern; isolates merge risk. |
| Cross-validation corpus generation | `scripts/gen-swg-cross-validation.py` (developer tool) | — | Python script generates the static JSON. NOT runtime / CI code. |
| Cross-validation gate at test time | `swg_test.go` (`TestSWG_CrossValidation`) reading committed JSON | — | Zero Python runtime at test time; CI reads the static fixture. |
| Golden-file integration | `testdata/golden/_staging/swg.json` → merge into `testdata/golden/algorithms.json` via `TestGolden_Algorithms_Merge` | — | Inherited staging-merge pattern from Phase 2 finalisation. |
| BDD acceptance | `tests/bdd/features/swg.feature` + extension to `tests/bdd/steps/algorithms_steps.go` | — | Inherited per-algorithm feature file pattern. |
| Property invariants | Append to `props_test.go` | — | Locked shared-file extension point from Phase 2. |
| Docs sync (llms.txt + docs/requirements.md §7.1.8 + examples/identifier-similarity) | Multiple touch-points in Plan 3 finalisation | — | Required by `ai_friendly_test.go` gate, the `api-ergonomics-reviewer` review for the Raw* expansion, and the locked staging-merge pattern. |

## Standard Stack

> All entries below are inherited from Phase 1 + Phase 2. Phase 3 introduces **no new runtime dependencies** (stdlib only) and **one new dev-only dependency** (biopython, used by `scripts/gen-swg-cross-validation.py`, never imported by Go code).

### Core

| Library | Version | Purpose | Why Standard |
|---------|---------|---------|--------------|
| Go stdlib | 1.26.3 | `testing`, `testing/quick`, `math` (Abs/Min/Max permitted; Sqrt/Pow forbidden — see Determinism), `unicode/utf8` for rune variant | `[VERIFIED: docs/requirements.md §1, CLAUDE.md "Constraints"]` Spec-locked. |
| `golang.org/x/text` | v0.37.0+ (root `go.mod`) | NOT used by Phase 3 directly — listed for completeness | `[VERIFIED: root go.mod]` Sole runtime dep allowlist; SWG does not call Normalise itself (callers pre-normalise). |

### Supporting

| Library | Version | Purpose | When to Use |
|---------|---------|---------|-------------|
| `cucumber/godog` | v0.15.0 (tests/bdd sub-module) | Phase 3 adds `tests/bdd/features/swg.feature` and step bindings | BDD acceptance, isolated in sub-module per spec |
| `go.uber.org/goleak` | v1.3.0 (tests/bdd sub-module) | Goroutine leak detection — Phase 3 introduces no goroutines but inherits the gate | Pure-function discipline check |
| `stretchr/testify` | v1.10.0 (tests/bdd sub-module ONLY) | BDD step assertions — `testify` is FORBIDDEN in root tests per CLAUDE.md | `tests/bdd/steps/algorithms_steps.go` only |
| `biopython` | 1.85+ (current stable as of 2026-01) | `scripts/gen-swg-cross-validation.py` invokes `Bio.Align.PairwiseAligner.score` | DEV-ONLY; never imported by Go; never in CI |

### Alternatives Considered

| Instead of | Could Use | Tradeoff |
|------------|-----------|----------|
| biopython `PairwiseAligner` | EMBOSS `water` CLI | EMBOSS is the textbook canonical implementation but requires C compilation + a CLI dependency; biopython is BSD-3-Clause Python with a clean API and is named in PITFALLS.md §3 alongside EMBOSS. CONTEXT.md §1 LOCKED biopython; EMBOSS deferred ("if a future regression is suspected, also verify against EMBOSS"). |
| three-matrix M/Ix/Iy form | Single-matrix Gotoh variant | Single-matrix saves one rolling row but is harder to reason about — the textbook M/Ix/Iy form maps directly to the corrected formulation in Flouri et al. 2015. With six rolling rows on a 3120-byte stack buffer, space cost is negligible. **Recommendation: stay with three-matrix form.** |
| Stack-buffer threshold 64 | Smaller (32) or larger (128) | `maxStackInputLen=64` is the LOCKED Phase 2 shared constant. SWG uses six rows × 8 bytes/float = 48 bytes/cell vs Phase 2's 8 bytes/cell for int rows, so a 64-cell row at six rows = 3120 bytes stack — still well within typical 8 KB Go stack frames. **Do not change `maxStackInputLen` for SWG.** |
| `float64` for DP cells | `int` with scaled integer params | `float64` is required because user-supplied params are real-valued. Determinism is preserved because the kernel uses only `+ - * / max` and `float64()` conversion (all IEEE-754-deterministic across linux/arm64, darwin/arm64, windows/amd64). |

**Installation:**
- Go: already in place (Go 1.26.3 toolchain).
- biopython (dev-only): `python3 -m pip install --user biopython` on the implementer's machine. Plan 03-02's Makefile target `regen-swg-cross-validation` invokes the script; no other developer needs biopython.

**Version verification:**
- biopython current stable: **1.85** (verified 2026-05-14 via `[CITED: biopython.org/docs/latest/Tutorial/chapter_pairwise.html]`). The Python script writes the version into the JSON header as a debugging aid; vector regeneration after a biopython point release is a deliberate developer action.
- Go: 1.26.3 already in `go.mod`; no change.
- godog/goleak/testify: unchanged from Phase 2 versions; no Phase 3 modifications to `tests/bdd/go.mod`.

## Architecture Patterns

### System Architecture Diagram

```
                                                      ┌──────────────────────────────┐
  caller (e.g. axonops/audit)                         │ scripts/gen-swg-cross-       │
        │                                              │ validation.py                │
        │  SmithWatermanGotohScore("foo", "foo_bar")  │  (dev-only)                  │
        ▼                                              │                              │
   ┌─────────────────────────────────────────┐        │  for case in CASES:          │
   │ swg.go                                  │        │    aligner.score(a, b)       │
   │                                         │        │    write {raw, normalised}   │
   │  SmithWatermanGotohScore (no params)    │        └──────────────┬───────────────┘
   │   → SmithWatermanGotohScoreWithParams   │                       │ writes
   │       (a, b, NewSWGParams())            │                       ▼
   │                                         │        ┌──────────────────────────────┐
   │  SmithWatermanGotohScoreWithParams      │        │ testdata/cross-validation/   │
   │   if a == b → 1.0                       │        │ swg/vectors.json             │
   │   if len(a)==0 && len(b)==0 → 1.0       │        │  (committed; ~10-20 entries) │
   │   if either len==0 → 0.0                │        └──────────────┬───────────────┘
   │                                         │                       │ read
   │   raw = swgDP(a, b, params)             │                       ▼
   │   norm = clamp(raw / min(la,lb), 0, 1)  │        ┌──────────────────────────────┐
   │   return norm                           │ ◄─────►│ swg_test.go                  │
   │                                         │  ε≤1e-9│   TestSWG_CrossValidation    │
   │  SmithWatermanGotohRawScoreWithParams   │        │     read vectors.json        │
   │   → returns raw (unclamped)             │        │     compare for each entry   │
   │                                         │        └──────────────────────────────┘
   │  swgDP (inner kernel, three-matrix      │
   │   M/Ix/Iy → 6 rolling rows; max-based;  │        ┌──────────────────────────────┐
   │   no math.*; ASCII fast-path gate       │ ◄─────►│ algorithms_golden_test.go    │
   │   isASCII(a)&&isASCII(b)&&n≤64 → stack) │        │   TestGolden_Algorithms_Merge│
   │                                         │        │     read _staging/swg.json   │
   │  *Runes variants (eager []rune)         │        │     merge into algorithms.   │
   └──────────┬──────────────────────────────┘        │     json (byte-stable)       │
              │  registered into dispatch[6]          └──────────────────────────────┘
              ▼
   ┌────────────────────────────────────────┐         ┌──────────────────────────────┐
   │ dispatch_swg.go                        │ ◄──────►│ tests/bdd/features/          │
   │  var _ = func()bool{                   │         │   swg.feature                │
   │    dispatch[AlgoSmithWatermanGotoh]    │         │     reference-vector scenarios│
   │     = SmithWatermanGotohScore          │         │     gap-split canary scenario│
   │    return true                         │         │     params scenario          │
   │  }()                                   │         └──────────────────────────────┘
   └────────────────────────────────────────┘
              │
              ▼ Phase 8 Scorer will read dispatch[AlgoSmithWatermanGotoh] in v0.5.0
   (out of Phase 3 scope; Phase 3 only populates the slot)
```

Data flow: caller invokes one of six SWG public functions → ASCII fast-path gate → either stack-buffer DP (zero allocs on short ASCII) or heap DP → raw alignment score → optional clamping to `[0,1]` → return. Independent dev path generates the cross-validation JSON; CI test path reads it without touching Python.

### Recommended Project Structure

```
fuzzymatch/                                  (root package)
├── swg.go                       NEW       # SWG public functions + DP kernel
├── dispatch_swg.go              NEW       # var-init registration into dispatch[6]
├── swg_test.go                  NEW       # unit tests + TestSWG_CrossValidation
├── swg_bench_test.go            NEW       # 6 benchmarks: ASCII Short/Medium/Long,
│                                          #   Unicode Short, WithParams ASCII Short,
│                                          #   RawScore ASCII Short
├── swg_fuzz_test.go             NEW       # FuzzSmithWatermanGotohScore + 7 seeds
├── props_test.go                EXTEND    # +6 standard + 3 SWG-specific property tests
├── example_test.go              EXTEND    # +ExampleSmithWatermanGotohScore (+ Raw)
├── algoid_test.go               EXTEND    # +TestDispatch_SmithWatermanGotohRegistered
├── algorithms_golden_test.go    EXTEND    # +TestGolden_SmithWatermanGotoh_Staging
├── cross_algorithm_consistency_test.go  EXTEND  # +SWG-vs-Levenshtein divergence test
├── llms.txt                     EXTEND    # +6 funcs + SWGParams + NewSWGParams + AlgoID
├── docs/requirements.md         EDIT      # §7.1.8 add Raw* surface (per CONTEXT.md §4)
├── examples/identifier-similarity/
│   ├── main.go                  EXTEND    # +1 column (7→8); update `want` constant
│   └── main_test.go             EXTEND    # update TestExample_Output + ColumnWidths
├── bench.txt                    APPEND    # +6 benchmark series rows
├── scripts/
│   └── gen-swg-cross-validation.py  NEW   # Python: biopython PairwiseAligner generator
├── testdata/
│   ├── golden/
│   │   ├── _staging/swg.json    NEW       # per-algorithm staging entries
│   │   └── algorithms.json      EXTEND    # merged via TestGolden_Algorithms_Merge
│   ├── cross-validation/swg/
│   │   └── vectors.json         NEW       # committed JSON corpus
│   └── fuzz/FuzzSmithWatermanGotohScore/
│       └── seed-001             NEW       # one literature reference vector
├── tests/bdd/
│   ├── features/swg.feature     NEW       # Gherkin scenarios per CONTEXT.md
│   └── steps/algorithms_steps.go  EXTEND  # +iComputeTheSmithWatermanGotohScoreBetween
│                                          #   + raw / params step variants
└── Makefile                     EXTEND    # +regen-swg-cross-validation target
```

### Pattern 1: Three-Matrix Two-Row DP (the load-bearing SWG kernel)

**What:** Smith-Waterman with affine gap requires three DP matrices conventionally named M (match/mismatch ends), Ix (gap in a / insertion in b), and Iy (gap in b / insertion in a). With local alignment, all three cells additionally take the max with 0 to allow the alignment to "reset" at any position.

**When to use:** This is the SWG kernel. Every other algorithm pattern in Phase 3 wraps around this kernel.

**Recurrence (corrected per Flouri et al. 2015 — verify against this, NOT the original 1982 paper alone):**

For `i = 1..m`, `j = 1..n`, with `s(a[i-1], b[j-1])` = `Match` if equal else `Mismatch`:

```
M[i][j]  = max( 0,
                M[i-1][j-1]  + s(a[i-1], b[j-1]),
                Ix[i-1][j-1] + s(a[i-1], b[j-1]),
                Iy[i-1][j-1] + s(a[i-1], b[j-1]) )

Ix[i][j] = max( 0,
                M[i-1][j]    + GapOpen,
                Ix[i-1][j]   + GapExtend )

Iy[i][j] = max( 0,
                M[i][j-1]    + GapOpen,
                Iy[i][j-1]   + GapExtend )
```

**Initialisation (the load-bearing correctness point — this is where the Gotoh 1982 erratum bites):**

```
M[0][*]  = 0;  M[*][0]  = 0
Ix[0][*] = 0;  Ix[*][0] = 0
Iy[0][*] = 0;  Iy[*][0] = 0
```

For local alignment, every cell along both borders initialises to 0 (not `-Inf` and not the gap-open-plus-extend ladder used in **global** alignment). This is the key point: for LOCAL alignment, the simple zero-init is correct; the erratum in Gotoh 1982 primarily affects the global-alignment border setup, but textbook treatments often blur the local/global distinction and propagate the wrong border into local implementations. **The unit tests' identity case `Score(x, x) = 1.0` fails when this is wrong — that's warning sign #1 in PITFALLS.md §3.**

**Best score:** `bestRaw = max over all (i, j) of M[i][j]` (Ix and Iy contribute only via their feed into M). Track running max during the DP fill, not as a post-pass scan.

**Two-row transcription:** every cell at `[i][j]` depends only on `[i-1][j-1]` (diagonal), `[i-1][j]` (above), and `[i][j-1]` (left). No `[i-2]` or `[j-2]` dependencies. Therefore each of the three matrices reduces to **two rolling rows** with the standard `prev, curr = curr, prev` swap. Six rolling rows total.

**Example:** (verified via algorithm-correctness derivation; no external code consulted)

```go
// swgDPRaw computes the raw Smith-Waterman-Gotoh local alignment score.
// Caller has ensured len(a) >= len(b) > 0 (m, n) and supplied six row buffers
// each of length n+1.
func swgDPRaw(a, b string, m, n int, params SWGParams,
              prevM, currM, prevIx, currIx, prevIy, currIy []float64) float64 {
    // Local-alignment zero-init: every border cell is 0.
    for j := 0; j <= n; j++ {
        prevM[j], prevIx[j], prevIy[j] = 0, 0, 0
    }
    bestRaw := 0.0
    for i := 1; i <= m; i++ {
        currM[0], currIx[0], currIy[0] = 0, 0, 0
        ai := a[i-1]
        for j := 1; j <= n; j++ {
            var sij float64
            if ai == b[j-1] {
                sij = params.Match
            } else {
                sij = params.Mismatch
            }

            // M[i][j]: best of (start fresh = 0) / (extend match-state)
            //          / (close gap on either side then match)
            m1 := prevM[j-1] + sij
            m2 := prevIx[j-1] + sij
            m3 := prevIy[j-1] + sij
            mij := 0.0
            if m1 > mij { mij = m1 }
            if m2 > mij { mij = m2 }
            if m3 > mij { mij = m3 }
            currM[j] = mij

            // Ix[i][j]: open new gap in a from M, or extend existing Ix
            xij := 0.0
            x1 := prevM[j] + params.GapOpen
            x2 := prevIx[j] + params.GapExtend
            if x1 > xij { xij = x1 }
            if x2 > xij { xij = x2 }
            currIx[j] = xij

            // Iy[i][j]: open new gap in b from M, or extend existing Iy
            yij := 0.0
            y1 := currM[j-1] + params.GapOpen
            y2 := currIy[j-1] + params.GapExtend
            if y1 > yij { yij = y1 }
            if y2 > yij { yij = y2 }
            currIy[j] = yij

            if mij > bestRaw { bestRaw = mij }
        }
        prevM, currM = currM, prevM
        prevIx, currIx = currIx, prevIx
        prevIy, currIy = currIy, prevIy
    }
    return bestRaw
}
```

(Above is illustrative — final form per `api-ergonomics-reviewer` and per Pattern 2/3 below; written fresh from the corrected recurrence, no code copied from any source.)

### Pattern 2: ASCII Fast-Path Gate (inherited LOCKED from Phase 2 Pattern 3)

**What:** Identical idiom to Levenshtein / DL-OSA / Jaro — gates the stack-buffer path on `n <= maxStackInputLen && isASCII(a) && isASCII(b)`.

**When to use:** Every entry point to the byte-level DP — `SmithWatermanGotohScore`, `SmithWatermanGotohScoreWithParams`, `SmithWatermanGotohRawScore`, `SmithWatermanGotohRawScoreWithParams`.

**Example:**

```go
// SmithWatermanGotohScoreWithParams returns the normalised local-alignment
// similarity between a and b in [0.0, 1.0] using the supplied affine-gap params.
//
// [godoc with clamp warning per CONTEXT.md §4 omitted for brevity]
func SmithWatermanGotohScoreWithParams(a, b string, params SWGParams) float64 {
    if a == b {
        return 1.0 // identity short-circuit (covers both-empty too)
    }
    la, lb := len(a), len(b)
    if la == 0 || lb == 0 {
        return 0.0
    }
    raw := smithWatermanGotohRawByte(a, b, la, lb, params)
    minLen := lb
    if la < lb { minLen = la }
    // Clamp(raw / float64(minLen), 0, 1)
    n := raw / float64(minLen)
    if n < 0 { return 0 }
    if n > 1 { return 1 }
    return n
}

// smithWatermanGotohRawByte performs the ASCII fast-path gate then dispatches
// to swgDPRaw with either a stack-allocated or heap-allocated row set.
func smithWatermanGotohRawByte(a, b string, la, lb int, params SWGParams) float64 {
    // Ensure b is the shorter dimension so the inner loop is minimal.
    if la < lb {
        a, b = b, a
        la, lb = lb, la
    }
    if lb <= maxStackInputLen && isASCII(a) && isASCII(b) {
        var buf [(maxStackInputLen + 1) * 6]float64
        n1 := lb + 1
        return swgDPRaw(a, b, la, lb, params,
            buf[0*n1:1*n1], buf[1*n1:2*n1],
            buf[2*n1:3*n1], buf[3*n1:4*n1],
            buf[4*n1:5*n1], buf[5*n1:6*n1])
    }
    // Heap path: 6 allocations of float64 slices.
    return swgDPRaw(a, b, la, lb, params,
        make([]float64, lb+1), make([]float64, lb+1),
        make([]float64, lb+1), make([]float64, lb+1),
        make([]float64, lb+1), make([]float64, lb+1))
}
```

**Sources:** `[VERIFIED: levenshtein.go lines 88–113]`, `[VERIFIED: damerau_osa.go]` for the locked idiom; `[VERIFIED: 02-PATTERNS.md Pattern 3]` for the rationale.

### Pattern 3: Dispatch Registration (inherited LOCKED from Phase 2 Pattern 6)

**What:** `dispatch_swg.go` registers `dispatch[AlgoSmithWatermanGotoh] = SmithWatermanGotohScore` via the `var _ = func() bool { ... }()` idiom — NO `init()` functions per `determinism-standards` §13.5.

**When to use:** Single instance, in `dispatch_swg.go`.

**Example:**

```go
// dispatch_swg.go (header omitted for brevity)
package fuzzymatch

var _ = func() bool {
    dispatch[AlgoSmithWatermanGotoh] = SmithWatermanGotohScore
    return true
}()
```

**Source:** `[VERIFIED: dispatch_levenshtein.go]`, `[VERIFIED: 02-PATTERNS.md Pattern 6]`.

### Pattern 4: Rune Variant (eager `[]rune` conversion)

**What:** `*Runes` functions convert both inputs eagerly to `[]rune`, then dispatch to a rune-aware kernel that operates on `[]rune` slices instead of byte indexing.

**When to use:** `SmithWatermanGotohScoreRunes`, `SmithWatermanGotohRawScoreRunes`. Eager conversion is the LOCKED Phase 2 strategy.

**Example:**

```go
func SmithWatermanGotohScoreRunes(a, b string) float64 {
    if a == b {
        return 1.0 // identity short-circuit — saves 2 []rune allocs (per IN-02 cleanup)
    }
    ra, rb := []rune(a), []rune(b)
    la, lb := len(ra), len(rb)
    if la == 0 || lb == 0 { return 0.0 }
    raw := smithWatermanGotohRawRunes(ra, rb, la, lb, NewSWGParams())
    minLen := lb
    if la < lb { minLen = la }
    n := raw / float64(minLen)
    if n < 0 { return 0 }
    if n > 1 { return 1 }
    return n
}
```

**Source:** `[VERIFIED: 02-PATTERNS.md Pattern 8]`; identity short-circuit pattern `[VERIFIED: commit c235e0e per 02-VERIFICATION.md]`.

### Pattern 5: Cross-Validation Generator Script

**What:** A Python script `scripts/gen-swg-cross-validation.py` that uses biopython's `Bio.Align.PairwiseAligner` in `local` mode to produce reference vectors. The script is invoked manually via `make regen-swg-cross-validation`; it writes `testdata/cross-validation/swg/vectors.json` (committed) which the Go test reads at `go test` time without requiring Python.

**When to use:** Single instance, written once in Plan 03-02 (or folded into 03-01).

**Example (schematic; final form per planner / executor):**

```python
#!/usr/bin/env python3
# scripts/gen-swg-cross-validation.py
#
# Regenerate testdata/cross-validation/swg/vectors.json from biopython's
# Bio.Align.PairwiseAligner (BSD-3-Clause licensed; permissive, compatible with
# Apache-2.0 for reference-vector cross-validation per
# .claude/skills/algorithm-licensing-standards).
#
# Run via:  make regen-swg-cross-validation
# Requires: python3 -m pip install biopython

import json
import os
import Bio
from Bio.Align import PairwiseAligner

DEFAULT_PARAMS = {
    "match": 1.0, "mismatch": -1.0,
    "gap_open": -1.5, "gap_extend": -0.5,
}

CASES = [
    # name, a, b, params_override (None → defaults)
    ("identity_short",       "hello",          "hello",                          None),
    ("both_empty",           "",               "",                               None),
    ("one_empty_a",          "",               "abcdef",                         None),
    ("one_empty_b",          "abcdef",         "",                               None),
    ("two_substring",        "http_request",   "http_request_header_fields",     None),
    ("substring_reversed",   "header_fields",  "http_request_header_fields",     None),
    ("no_overlap",           "qqqq",           "zzzz",                           None),
    ("partial_middle_match", "the_quick_fox",  "the_brown_quick_fox_jumps",      None),
    ("one_long_gap_canary",  "abc________def", "abcdef",                         None),
    ("gap_split_canary",     "abc_def_ghi",    "abcdefghi",                      None),
    ("non_default_params",   "hello",          "hallo",                          {"match": 2.0, "mismatch": -2.0,
                                                                                   "gap_open": -3.0, "gap_extend": -1.0}),
    # ... add 5–10 more spanning unicode, single-char, all-mismatch, etc.
]

def score_case(a, b, params):
    aligner = PairwiseAligner()
    aligner.mode = "local"
    aligner.match_score = params["match"]
    aligner.mismatch_score = params["mismatch"]
    aligner.open_gap_score = params["gap_open"]
    aligner.extend_gap_score = params["gap_extend"]
    raw = aligner.score(a, b)
    # Normalise the same way fuzzymatch does:
    if a == "" and b == "":
        return raw, 1.0
    if a == "" or b == "":
        return raw, 0.0
    min_len = min(len(a), len(b))
    n = raw / min_len
    norm = max(0.0, min(1.0, n))
    return raw, norm

def main():
    entries = []
    for name, a, b, overrides in CASES:
        params = dict(DEFAULT_PARAMS)
        if overrides:
            params.update(overrides)
        raw, norm = score_case(a, b, params)
        entries.append({
            "name": name, "a": a, "b": b, "params": params,
            "biopython_score": raw, "biopython_normalised": norm,
        })
    out = {"version": 1, "biopython_version": Bio.__version__, "entries": entries}
    path = "testdata/cross-validation/swg/vectors.json"
    os.makedirs(os.path.dirname(path), exist_ok=True)
    with open(path, "w") as f:
        json.dump(out, f, indent=2, sort_keys=False)
        f.write("\n")

if __name__ == "__main__":
    main()
```

**Source:** biopython PairwiseAligner API verified at `[CITED: biopython.org/docs/latest/Tutorial/chapter_pairwise.html]`. Sign convention (negative gap penalties) confirmed.

### Anti-Patterns to Avoid

- **Copying code from biopython, EMBOSS, or any existing Go SWG implementation.** The implementation must be written fresh from Smith-Waterman 1981 + Gotoh 1982 + the corrected formulation per Flouri et al. 2015. Biopython is consulted ONLY to verify reference vector outputs in JSON. `algorithm-licensing-reviewer` will reject the PR if any code derivation is detectable.
- **Cribbing the recurrence from the original Gotoh 1982 paper alone.** This is exactly the failure mode the phase is gated against — the paper has documented errata. ALWAYS verify the recurrence against Flouri et al. 2015's corrected formulation.
- **Initialising borders with `-Inf` (global-alignment convention) for local alignment.** Identity test will fail. Use `0` for every border cell.
- **Hand-rolling a normalisation different from `clamp(raw / min(la, lb), 0, 1)`.** This is the spec-locked formula in `docs/requirements.md` §7.1.8.
- **Using `math.Pow`/`math.Exp`/`math.Log`/`math.FMA` anywhere.** Forbidden by determinism-standards. SWG needs only `+`, `-`, `*`, `/`, and a max-style if-comparison.
- **Iterating any map on the output path.** SWG doesn't use maps; the kernel uses plain slices. Keep it that way.
- **Adding goroutines / channels / mutexes.** Library-wide ban; SWG is pure-function.
- **`SWGDefaultParams` as a package-level exported var.** CONTEXT.md §3 explicitly LOCKED-OUT this alternative in favour of `NewSWGParams()` (the "is this read-only?" footgun). Spec mentions `SWGDefaultParams` but it's superseded by the discussion decision.
- **Validating params in `*Score` functions.** Per CONTEXT.md §3 and §5.11, nonsense params still produce a deterministic score. Documentation, not validation, is the contract.

## Don't Hand-Roll

| Problem | Don't Build | Use Instead | Why |
|---------|-------------|-------------|-----|
| Reference SWG implementation for cross-validation oracle | Custom Python re-implementation; "trusted" pure-Go reference; cribbed C `water` port | biopython `Bio.Align.PairwiseAligner` with `mode="local"` | BSD-3-Clause; actively maintained; named in PITFALLS.md §3; CONTEXT.md §1 LOCKED. Any custom oracle is just another implementation to audit. |
| Gotoh recurrence | Crib from a textbook summary | Cite Flouri et al. 2015 alongside Gotoh 1982; transcribe the corrected formulation directly | Five of ten audited implementations and 8 of 31 lecture slides reproduce the erratum. Defensive citation IS the discipline. |
| Two-row buffer plumbing | Bespoke pool / arena | Stack-allocated `[(maxStackInputLen+1)*6]float64` for short ASCII; six `make([]float64, n+1)` for the heap path | Inherited Phase 2 pattern; escape analysis verified; 0-alloc fast path mandatory per PERF-01. |
| ASCII detection | New helper | Use existing `isASCII(s string) bool` from `normalise.go` (lines 159–168) | LOCKED in 02-PATTERNS.md Pattern 3. Redeclaring in `swg.go` fails the build (package collision). |
| Test framework | New harness | Stdlib `testing` + `testing/quick` in root; `cucumber/godog` in `tests/bdd` sub-module; `stretchr/testify` ONLY in `tests/bdd/steps/` | Project-wide locked. No testify in root EVER. |
| Cross-validation JSON serialisation | Custom marshaller | Python `json.dump(obj, indent=2, sort_keys=False)` + trailing newline; Go `encoding/json` for read-side; field-order discipline by mirroring Phase 2's `canonicalMarshalForTest` semantics for the staging golden file (not the cross-validation file — that one is human-written-by-script and read once) | Determinism of the cross-validation fixture is owned by the Python script's deterministic CASES list ordering. |
| Floating-point comparison in test | `==` on `float64` | `math.Abs(got - want) <= 1e-9` (absolute epsilon) | Matches `cross_algorithm_consistency_test.go` epsilon convention. |
| Custom alignment trace return type | An `Alignment` struct returning matched positions / gaps | Defer to v1.x per CONTEXT.md `<deferred>` | Out of v1.0 scope; advanced consumers use `Raw*` for absolute quality. |
| Distance variant `SmithWatermanGotohDistance` | Hand-rolled "inverse" of score | DEFER — SWG is not a metric; `Raw*` is the unnormalised quality signal | Per CONTEXT.md `<deferred>`. |

**Key insight:** SWG is "one algorithm, six functions" in fuzzymatch's surface. The cross-validation gate is what makes Phase 3 distinct — everything else is mechanical replication of Phase 2's pattern. Don't reinvent the gate (biopython is the LOCKED oracle); don't reinvent the pattern (Phase 2 is the LOCKED template).

## Runtime State Inventory

> Phase 3 is a greenfield algorithm phase — no renames, refactors, or migrations of existing state.
> The only "state" introduced is committed test fixtures (golden + cross-validation JSON) which are
> deterministic outputs of pure computations.

Not applicable to this phase. Confirming explicitly:
- Stored data: None. SWG is pure-function; no database, no cache, no persistent store.
- Live service config: None. fuzzymatch is a library, not a service.
- OS-registered state: None.
- Secrets / env vars: None.
- Build artifacts: None — Phase 3 adds new files only; no existing build artifacts go stale.

## Common Pitfalls

### Pitfall 1: Reproducing the Gotoh 1982 initialisation erratum

**What goes wrong:** The unit test `TestSWG_Identity_OneLongGap` fails. Or `TestSWG_Symmetric` fails for some inputs. Or `TestSWG_CrossValidation` fails on the `one_long_gap_canary` entry. The implementation works on the simple reference vectors but breaks on inputs that exercise the gap-initialisation cells.

**Why it happens:** The implementer worked from Gotoh 1982 alone, or a textbook summary derived from it, and reproduced the documented initialisation error. Per Flouri et al. 2015, 5 of 10 audited implementations have this bug. The error is "easy to miss" because simple test cases pass.

**How to avoid:**
- Cite Flouri et al. 2015 in the file's block comment with a one-paragraph summary of the corrected initialisation.
- For LOCAL alignment specifically: every border cell of M, Ix, Iy is `0` (not `-Inf`, not a ladder).
- Include the `one_long_gap_canary` entry in `vectors.json` and `TestSWG_CrossValidation`.
- Include the gap-split-invariance property test (CONTEXT.md §5).
- Run `algorithm-correctness-reviewer` on the PR — that gate is mandatory and specifically flags SWG per `.claude/skills/algorithm-correctness-standards/SKILL.md` and PITFALLS.md §3.

**Warning signs:** identity test fails for some inputs; gap-split asymmetry; score asymmetry; biopython cross-validation epsilon failure.

### Pitfall 2: Indexing flip in the recurrence

**What goes wrong:** `SmithWatermanGotohScore(a, b) != SmithWatermanGotohScore(b, a)` for some inputs (asymmetry).

**Why it happens:** Gotoh 1982 has an indexing mistake separate from the initialisation issue. The implementer writes `Ix[i][j-1]` where the correct cell is `Ix[i-1][j]`, or similar.

**How to avoid:**
- Transcribe the recurrence from Flouri et al. 2015's corrected form (above in Pattern 1), NOT from Gotoh 1982 directly.
- The convention used: `Ix` accumulates gap-in-a (column advancing without row advancing — `[i-1][j]` extends `Ix`, `[i-1][j]` from `M` opens it). `Iy` is the symmetric counterpart.
- Property test `TestProp_SmithWatermanGotohScore_Symmetric` catches this on random inputs.
- Hand-curated unit test: `SmithWatermanGotohScore("ab__cd", "abcd") == SmithWatermanGotohScore("abcd", "ab__cd")`.

### Pitfall 3: Float-determinism breakage across linux/arm64 vs windows/amd64

**What goes wrong:** Golden file test fails on the cross-platform CI matrix even though the implementation passes locally.

**Why it happens:** A `math.Pow`/`math.FMA`/`math.Log` slipped into the kernel; or a sum-reduction is run in parallel (non-deterministic associativity); or an `init()`-time table introduces platform-dependent ordering.

**How to avoid:**
- Kernel uses only `+`, `-`, `*`, `/`, `float64(int)`, and the `if x > y { x = y }` max idiom. NO `math.*` calls at all.
- No `init()` functions; `NewSWGParams()` is a constructor function called by callers, not a package-init side-effect.
- No goroutines; serial DP only.
- Property test `TestProp_SmithWatermanGotohScore_NoNaN/NoInf/NoNegativeZero` catches the NaN-on-NaN-params case.
- The `_staging/swg.json` golden file in the cross-platform CI matrix gate (DET-01 — Phase 1 UAT carry-forward) is the final byte-identical check.

### Pitfall 4: 0-alloc budget regression from accidental escape

**What goes wrong:** Bench shows `1 alloc/op` on `BenchmarkSmithWatermanGotohScore_ASCII_Short` (target: 0).

**Why it happens:** The six row slices passed to `swgDPRaw` escape because the function returns one of them, or stores them in a struct field, or passes them to another function that escapes them. Or the closure form `var _ = func() bool { ... }()` introduces escape on a path. Or the slice headers are stored in a heap-allocated `params SWGParams` (shouldn't — SWGParams is a value type).

**How to avoid:**
- Pass `buf[lo:hi]` slices ONLY as function arguments. Never store in a struct, never return.
- Verify with `go build -gcflags="-m=2" ./... 2>&1 | grep -E "does not escape|escapes to heap"`.
- `b.ReportAllocs()` + benchmark's `_ASCII_Short` row must show `0 B/op, 0 allocs/op`.
- Use `testing.AllocsPerRun` for runtime alloc gate in `swg_test.go::TestSWG_ZeroAllocs_ASCII_Short` (mirror Phase 2's Levenshtein test of the same shape).

### Pitfall 5: Forgetting to update `docs/requirements.md` §7.1.8 for the Raw* surface

**What goes wrong:** API drift between the source code (6 functions) and the spec (3 functions). `api-ergonomics-reviewer` flags during PR review.

**Why it happens:** CONTEXT.md §4 expanded the public surface beyond the spec; the implementer focuses on code and forgets the docs edit.

**How to avoid:**
- Plan 03-03 finalisation has the docs/requirements.md §7.1.8 edit as an explicit deliverable.
- `llms.txt` must list all 6 functions + `SWGParams` + `NewSWGParams` + `AlgoSmithWatermanGotoh` = 9 entries. The `ai_friendly_test.go` gate catches missing entries via `TestAIFriendly_LLMSTxtReferencesEveryExportedSymbol`.

### Pitfall 6: Identity-short-circuit interaction with Raw scores

**What goes wrong:** `SmithWatermanGotohRawScore(x, x)` returns some non-obvious value because the implementer added `if a == b { return ??? }` short-circuit on the Raw path without thinking through what raw means for identical strings.

**Why it happens:** Pattern 4 (identity short-circuit on `*Runes`) makes sense for the NORMALISED `Score` family (`return 1.0`). For the `Raw` family, identical strings should return `Match * len(a)` — that's the raw alignment score for a full match of every position.

**How to avoid:**
- Identity short-circuit on normalised functions returns `1.0`.
- Identity short-circuit on raw functions either (a) returns `float64(len(a)) * params.Match` directly, or (b) falls through to the DP (safer; the DP correctly computes this value).
- Unit test: `TestSWG_RawScore_IdentityEqualsMatchTimesLen` asserts the relationship explicitly.

## Code Examples

### Public API surface (six functions + helpers)

```go
// SmithWatermanGotohScore returns the Smith-Waterman-Gotoh local-alignment
// similarity between a and b as a value in [0.0, 1.0] using the default
// affine-gap parameters (NewSWGParams). The returned score is CLAMPED: if the
// underlying alignment score is negative (e.g. two unrelated strings dominated
// by mismatch/gap penalties) the clamp returns 0.0; if it exceeds
// min(len(a), len(b)) (custom params with Match > 1.0) the clamp returns 1.0.
// Use SmithWatermanGotohRawScore for the unclamped raw alignment score.
//
// score = clamp(best_local_score / min(len(a), len(b)), 0.0, 1.0)
//
// Edge cases:
//   - SmithWatermanGotohScore("", "")        == 1.0  (both-empty identity)
//   - SmithWatermanGotohScore("", "abc")     == 0.0  (one-empty)
//   - SmithWatermanGotohScore(x, x)          == 1.0  for any non-empty x
//   - SmithWatermanGotohScore(a, b) == SmithWatermanGotohScore(b, a)  (symmetric)
func SmithWatermanGotohScore(a, b string) float64 {
    return SmithWatermanGotohScoreWithParams(a, b, NewSWGParams())
}

func SmithWatermanGotohScoreRunes(a, b string) float64       { /* ... */ }
func SmithWatermanGotohScoreWithParams(a, b string, params SWGParams) float64 { /* ... */ }

// SmithWatermanGotohRawScore returns the unclamped best local-alignment score.
// May be negative (two unrelated strings) or exceed min(len(a), len(b)) (high
// Match reward). Use SmithWatermanGotohScore for the normalised [0,1] form.
func SmithWatermanGotohRawScore(a, b string) float64 {
    return SmithWatermanGotohRawScoreWithParams(a, b, NewSWGParams())
}

func SmithWatermanGotohRawScoreRunes(a, b string) float64       { /* ... */ }
func SmithWatermanGotohRawScoreWithParams(a, b string, params SWGParams) float64 { /* ... */ }
```

### Cross-validation test

```go
// TestSWG_CrossValidation asserts every entry in testdata/cross-validation/swg/
// vectors.json matches our implementation within 1e-9. The JSON is regenerated
// from biopython via `make regen-swg-cross-validation`; the test reads the
// committed file without invoking Python.
func TestSWG_CrossValidation(t *testing.T) {
    data, err := os.ReadFile("testdata/cross-validation/swg/vectors.json")
    if err != nil { t.Fatalf("read corpus: %v", err) }
    var corpus struct {
        Version          int    `json:"version"`
        BiopythonVersion string `json:"biopython_version"`
        Entries []struct {
            Name string `json:"name"`
            A, B string
            Params struct {
                Match, Mismatch, GapOpen, GapExtend float64
            } `json:"params"`
            BiopythonScore      float64 `json:"biopython_score"`
            BiopythonNormalised float64 `json:"biopython_normalised"`
        } `json:"entries"`
    }
    if err := json.Unmarshal(data, &corpus); err != nil { t.Fatalf("unmarshal: %v", err) }
    if corpus.Version != 1 { t.Fatalf("unsupported corpus version %d", corpus.Version) }
    for _, e := range corpus.Entries {
        e := e
        t.Run(e.Name, func(t *testing.T) {
            params := fuzzymatch.SWGParams{
                Match: e.Params.Match, Mismatch: e.Params.Mismatch,
                GapOpen: e.Params.GapOpen, GapExtend: e.Params.GapExtend,
            }
            gotRaw := fuzzymatch.SmithWatermanGotohRawScoreWithParams(e.A, e.B, params)
            gotNorm := fuzzymatch.SmithWatermanGotohScoreWithParams(e.A, e.B, params)
            const eps = 1e-9
            if absFloat64(gotRaw-e.BiopythonScore) > eps {
                t.Errorf("raw score drift: a=%q b=%q params=%+v got=%g biopython=%g delta=%g (eps %g)",
                    e.A, e.B, params, gotRaw, e.BiopythonScore, gotRaw-e.BiopythonScore, eps)
            }
            if absFloat64(gotNorm-e.BiopythonNormalised) > eps {
                t.Errorf("normalised score drift: a=%q b=%q got=%g biopython_norm=%g (eps %g)",
                    e.A, e.B, gotNorm, e.BiopythonNormalised, eps)
            }
        })
    }
}
```

### Gap-split invariance property (the SWG-specific canary)

```go
// TestProp_SmithWatermanGotoh_GapSplitInvariance is the canonical Gotoh-erratum
// canary. With the corrected formulation, splitting a single long gap into two
// halves separated by zero match characters yields the SAME raw score (the
// gap-open penalty is incurred exactly once in either form because there are no
// match characters to "close" the gap between the two halves).
//
// Hand-curated triples that exercise the case.
func TestProp_SmithWatermanGotoh_GapSplitInvariance(t *testing.T) {
    // For each triple, build "a" = prefix + one_long_gap + suffix and
    // "a_split" = prefix + half + half + suffix where the original "long gap"
    // is broken into two zero-length-separation halves. Without the corrected
    // initialisation, these would score differently.
    tests := []struct {
        b           string
        prefix      string
        gapLen      int
        suffix      string
    }{
        // [...hand-curated cases — see PITFALLS.md §3 warning sign #2...]
    }
    // ... assert that raw scores are equal for the two forms within 1e-12.
}
```

## State of the Art

| Old Approach | Current Approach | When Changed | Impact |
|--------------|------------------|--------------|--------|
| Cite Gotoh 1982 only | Cite Gotoh 1982 + Flouri et al. 2015 corrected formulation | Established 2015 (Flouri et al. biorxiv); incorporated into fuzzymatch as a PHASE 3 GATE | Implementations that cite Gotoh alone risk reproducing the erratum; cross-validation is mandatory. |
| Full m×n DP table | Two-row form (six rolling rows for affine-gap M/Ix/Iy) | Standard since Gotoh 1982 abstract; mandated by `.claude/skills/performance-standards/SKILL.md` for fuzzymatch | 0-alloc fast path on short ASCII inputs; matches Phase 2 budget. |
| `pairwise2.align.localxs` for cross-validation | `Bio.Align.PairwiseAligner` (mode="local", explicit gap params) | biopython v1.79 (2021) deprecated `pairwise2` in favour of `PairwiseAligner`; `pairwise2` retained for backward compatibility but discouraged | CONTEXT.md §1 LOCKED PairwiseAligner. |
| Single matrix Gotoh variant | Three-matrix M/Ix/Iy form | Both have been used historically; M/Ix/Iy is the textbook teaching form and aligns more transparently with Flouri et al.'s corrected formulation | Recommendation: stay with M/Ix/Iy for legibility and erratum-correlation. |

**Deprecated / outdated:**
- `pairwise2.align.localxs` — still works but marked deprecated since biopython v1.79; CONTEXT.md §1 names PairwiseAligner explicitly.

## Assumptions Log

| # | Claim | Section | Risk if Wrong |
|---|-------|---------|---------------|
| A1 | The Flouri et al. 2015 biorxiv paper specifically corrects both the indexing and initialisation errata that appear in Gotoh 1982. | Pattern 1, State of the Art | LOW — `[CITED: biorxiv.org/content/10.1101/031500v1]` abstract explicitly names both errors; PITFALLS.md §3 transitively confirms via "indexing flip and initialisation issue". If the paper's correction differs from what's transcribed in Pattern 1's recurrence, the implementation may transcribe a different (also wrong) form. Mitigation: `algorithm-correctness-reviewer` and the biopython cross-validation gate catch this regardless. |
| A2 | biopython 1.85's `PairwiseAligner(mode="local")` with the four standard affine-gap params (match_score, mismatch_score, open_gap_score, extend_gap_score) implements the corrected Gotoh recurrence (not the original erratum). | Pattern 5, Cross-Validation test | LOW — biopython is widely used in bioinformatics and would have known regressions if it reproduced the erratum. Flouri et al. surveyed multiple implementations, of which biopython was found to be among the correct ones (5 of 10 audited were incorrect, leaving 5 correct; biopython is one of the most widely-used and is the reference oracle in PITFALLS.md §3). If biopython is wrong, every other audit is also wrong; cross-validation against EMBOSS (deferred per CONTEXT.md) would catch it. |
| A3 | A `float64` two-row DP with only `+ - * / max` operations produces byte-identical output across linux/amd64, linux/arm64, darwin/amd64, darwin/arm64, windows/amd64. | Determinism Risk; performance-standards skill | LOW — `.claude/skills/determinism-standards/SKILL.md` §"Float Stability" verifies this property generally. Cross-platform CI matrix (DET-01) is the final gate, currently a Phase 1 UAT carry-forward item. If a SWG-specific float operation breaks determinism, golden file fails on the matrix run. |
| A4 | The two-row form admits stack allocation of all six rolling rows from a single `[(maxStackInputLen+1)*6]float64` buffer (3120 bytes) without escape, on Go 1.26.3. | Pattern 2 | LOW — Phase 2 verified this for the two-row int buffer; float64 doubles the per-cell size but stays on the same escape-analysis pattern. Verify with `go build -gcflags="-m=2"` during implementation. |
| A5 | Smith-Waterman algorithm in software is patent-free; only specific hardware accelerations are patented. | Patent Screen | LOW — `[CITED: en.wikipedia.org/wiki/Smith-Waterman_algorithm]`, `[CITED: algorithm-licensing-standards/SKILL.md "Smith-Waterman" appears in "Clear" list]`. The 1981 publication has no patent claim; Gotoh 1982 likewise. AxonOps's algorithm-licensing-standards explicitly lists Smith-Waterman as clear. |
| A6 | Default params `{Match: 1.0, Mismatch: -1.0, GapOpen: -1.5, GapExtend: -0.5}` match `docs/requirements.md` §7.1.8 stated defaults and produce reasonable scores for typical identifier-similarity inputs. | CONTEXT.md §3 | LOW — defaults are inherited verbatim from the spec; CONTEXT.md §3 reaffirms. Phase 3's identifier-similarity example output will validate they produce intuitive rankings on the existing 7-row case set. |

**Note:** All claims tagged `[VERIFIED: ...]` in the body are not in this table — they have been confirmed against the cited source. Claims tagged `[CITED: ...]` are referenced from official sources but their precise interaction with the implementation is an assumption pending implementation. The six entries above represent the residual assumptions where implementation could surprise.

## Open Questions (RESOLVED)

1. **Should `RawScore` apply the identity short-circuit `if a == b { return float64(len(a)) * params.Match }`, or fall through to the DP?**
   - What we know: Phase 2 Pitfall 6 (in §Common Pitfalls) raises this. Pattern 4 (identity short-circuit on `*Runes`) saves the 2 `[]rune` allocs and is unambiguous for normalised functions (returns 1.0).
   - What's unclear: for the `Raw*` family, the short-circuit would have to compute `len(a) * params.Match` (or `len(ra) * params.Match` on the rune path). This is correct (the raw alignment score for identical strings is `Match` × length) but obscures the DP path's correctness on this case.
   - Earlier recommendation: fall through to the DP for `Raw*` family on identity input (add unit test asserting the relationship).
   - **RESOLVED (2026-05-14, planner): Identity short-circuit on every `*RawScore` entry point**, returning `params.Match * float64(len(a))` for the byte path and `params.Match * float64(len(ra))` for the rune path (where `ra` is the rune-converted input). Chosen over the fall-through alternative because: (a) consistency with the `*Score` family short-circuit (Pattern 4 is the project-wide locked convention for ASCII fast-paths); (b) avoids the `[]rune(a)`/`[]rune(b)` allocations on the rune-variant identity path; (c) is mathematically correct (raw alignment of identical strings IS `Match × len`). The "DP correctness exercised on identity-shape input" coverage that the fall-through alternative offered is preserved by `TestSmithWatermanGotoh_SubstringContainment` in plan 03-01 Task 2, which asserts `RawScore("http_request", "http_request_header_fields") == 12.0` — strings differ (so short-circuit is bypassed), the local-alignment optimum is a perfect 12-position substring match, and the DP must produce exactly `Match × min(len)` = 12.0. The short-circuit's correctness is covered separately by `TestSmithWatermanGotoh_Identical` which asserts `RawScore("abc", "abc") == 3.0`.

2. **File name: `swg.go` or `smith_waterman_gotoh.go`?**
   - What we know: Phase 2 uses unprefixed short names (`levenshtein.go`, `hamming.go`, `jaro.go`). The dispatch file in Phase 2 is `dispatch_levenshtein.go` (long name).
   - What's unclear: SWG is a longer name; `smith_waterman_gotoh.go` is verbose; `swg.go` is concise but less searchable.
   - **RESOLVED (2026-05-14, planner): `swg.go`** for the algorithm file, **`dispatch_swg.go`** for dispatch, matching Phase 2's concise convention. The block-comment header expands "SWG" to the full "Smith-Waterman-Gotoh" name on the first reference per `algorithm-correctness-standards`. Public function/type names use the full `SmithWatermanGotoh` prefix.

3. **Should the cross-validation JSON `params` field be a flat `{match, mismatch, gap_open, gap_extend}` object, or use the Go field names `{Match, Mismatch, GapOpen, GapExtend}`?**
   - What we know: Python uses snake_case naturally; Go uses CamelCase.
   - What's unclear: which convention to commit to in the JSON.
   - **RESOLVED (2026-05-14, planner): snake_case in JSON** (`match`, `mismatch`, `gap_open`, `gap_extend`). Python writes the file; Go's `encoding/json` tag maps the field at read time (`json:"match"`, `json:"mismatch"`, `json:"gap_open"`, `json:"gap_extend"` on the internal test-only `crossValidationEntry` struct). This is also kinder to other-language readers of the corpus.

4. **Should `TestSWG_CrossValidation` skip itself if `testdata/cross-validation/swg/vectors.json` is absent (e.g. a fresh clone before regeneration), or fail loud?**
   - What we know: Per CONTEXT.md §1, the JSON is COMMITTED. Absence means the developer deleted it.
   - **RESOLVED (2026-05-14, planner): Fail loud** via `t.Fatalf("cross-validation corpus missing: %v — re-run make regen-swg-cross-validation", err)` so an accidental delete doesn't silently weaken the verification surface.

## Environment Availability

| Dependency | Required By | Available | Version | Fallback |
|------------|------------|-----------|---------|----------|
| Go toolchain | Root package compilation, tests, benches | ✓ | 1.26.3 (per `go.mod`) | — |
| `golang.org/x/text` | Root runtime dep (allowlisted) — not used directly by SWG | ✓ | 0.37.0+ (per existing `go.mod`) | — |
| `cucumber/godog` | `tests/bdd/` sub-module | ✓ | 0.15.0 (per existing `tests/bdd/go.mod`) | — |
| `stretchr/testify` | `tests/bdd/steps/` only | ✓ | 1.10.0 (per existing `tests/bdd/go.mod`) | — |
| `go.uber.org/goleak` | `tests/bdd/` sub-module | ✓ | 1.3.0 (per existing `tests/bdd/go.mod`) | — |
| Python 3.10+ | `scripts/gen-swg-cross-validation.py` (DEV-ONLY) | ✓ assumed (standard developer environment) | — | If absent on a developer's machine, the JSON is regenerated by someone with biopython. CI does not need Python. |
| `biopython` | `scripts/gen-swg-cross-validation.py` (DEV-ONLY) | ✗ unverified on local environment | — | `python3 -m pip install --user biopython` once on the regenerating developer's machine. CI does NOT install biopython. |

**Missing dependencies with no fallback:** None blocking Phase 3 — Phase 3 work proceeds without biopython on most developer machines; the script-runner installs it once.

**Missing dependencies with fallback:** biopython on non-regenerating machines — fallback is "let one developer regenerate the JSON and commit; everyone else reads it".

## Validation Architecture

### Test Framework

| Property | Value |
|----------|-------|
| Framework | Go 1.26.3 stdlib `testing` + `testing/quick` (root); `cucumber/godog` v0.15.0 + `stretchr/testify` v1.10.0 + `go.uber.org/goleak` v1.3.0 (tests/bdd sub-module) |
| Config file | `go.mod` (Go 1.26.3); `tests/bdd/go.mod` (godog/goleak/testify); `Makefile` defines `make check`, `make test`, `make test-bdd`, plus new `make regen-swg-cross-validation` |
| Quick run command | `go test ./...` (root only) |
| Full suite command | `make check` (lint + vet + tests + coverage + license-headers + verify-deps-allowlist + tidy-check + security + tests/bdd suite) |

### Phase Requirements → Test Map

| Req ID | Behavior | Test Type | Automated Command | File Exists? |
|--------|----------|-----------|-------------------|-------------|
| CHAR-08 | `SmithWatermanGotohScore("kitten","sitting")` returns deterministic value in `[0,1]` (existence of public function) | unit | `go test -run TestSmithWatermanGotoh_BothEmpty -v ./...` | ❌ Wave 0 |
| CHAR-08 | `SmithWatermanGotohScore` reference vectors from primary source (literature-derived pairs documented inline; matches biopython within `1e-9`) | unit | `go test -run TestSmithWatermanGotoh_ReferenceVectors -v ./...` | ❌ Wave 0 |
| CHAR-08 | Identity `Score(x, x) == 1.0` for non-empty `x` | property | `go test -run TestProp_SmithWatermanGotohScore_Identity -count=1 ./...` | ❌ Wave 0 |
| CHAR-08 | Range bounds `[0,1]` for arbitrary inputs | property | `go test -run TestProp_SmithWatermanGotohScore_RangeBounds -count=1 ./...` | ❌ Wave 0 |
| CHAR-08 | Symmetry `Score(a, b) == Score(b, a)` (byte path) | property | `go test -run TestProp_SmithWatermanGotohScore_Symmetric -count=1 ./...` | ❌ Wave 0 |
| CHAR-08 | Symmetry (rune path — per WR-03 cleanup pattern) | property | `go test -run TestProp_SmithWatermanGotohScoreRunes_Symmetric -count=1 ./...` | ❌ Wave 0 |
| CHAR-08 | NaN / Inf / negative-zero forbidden | property | `go test -run 'TestProp_SmithWatermanGotohScore_No(NaN|Inf|NegativeZero)' -count=1 ./...` | ❌ Wave 0 |
| CHAR-08 | Gap-split invariance (Gotoh erratum canary — PITFALLS.md §3 warning sign #2) | property (hand-curated triples) | `go test -run TestProp_SmithWatermanGotoh_GapSplitInvariance -v ./...` | ❌ Wave 0 |
| CHAR-08 | `RawScore <= Match * min(len(a), len(b))` upper bound | property | `go test -run TestProp_SmithWatermanGotoh_RawNeverExceedsMatchTimesMinLen -count=1 ./...` | ❌ Wave 0 |
| CHAR-08 | `RawScore` monotonic with increasing Match parameter | property | `go test -run TestProp_SmithWatermanGotoh_MonotonicWithMatchReward -count=1 ./...` | ❌ Wave 0 |
| CHAR-08 | Implementation MUST cross-validate against biopython reference vectors (the phase-defining gate per PITFALLS.md §3) | integration | `go test -run TestSWG_CrossValidation -v ./...` | ❌ Wave 0 |
| CHAR-08 | Configurable affine gap penalty (`*WithParams` variant; default `NewSWGParams()` produces documented defaults) | unit | `go test -run TestSmithWatermanGotoh_WithParams -v ./...` | ❌ Wave 0 |
| CHAR-08 | Raw scores exposed (`*Raw*` family of three functions; identity → `Match × len`; symmetric) | unit | `go test -run TestSmithWatermanGotoh_RawScore -v ./...` | ❌ Wave 0 |
| CHAR-08 | 0-alloc on ASCII ≤ 64 bytes (stack-buffer fast path); allocation budget enforced | benchmark + runtime check | `go test -run TestSWG_ZeroAllocs_ASCII_Short -v ./...` + `go test -bench BenchmarkSmithWatermanGotohScore_ASCII_Short -benchmem -count=10 ./...` | ❌ Wave 0 |
| CHAR-08 | Two-row DP variant (verified by stack-buffer escape analysis test + benchmark) | structural / benchmark | `go build -gcflags='-m=2' ./... 2>&1 \| grep swg` + bench above | ❌ Wave 0 |
| CHAR-08 | Cross-platform golden file entry (`testdata/golden/_staging/swg.json` → merged into `algorithms.json`) | golden | `go test -run 'TestGolden_(SmithWatermanGotoh_Staging\|Algorithms_Merge)' -v ./...` | ❌ Wave 0 |
| CHAR-08 | BDD scenario covering the canonical long-gap reference case | BDD | `cd tests/bdd && go test ./...` | ❌ Wave 0 |
| CHAR-08 | Fuzz coverage panic-free | fuzz | `go test -run FuzzSmithWatermanGotohScore -fuzz FuzzSmithWatermanGotohScore -fuzztime 60s ./...` (CI) + seed-only run via `go test -run FuzzSmithWatermanGotohScore -v ./...` | ❌ Wave 0 |
| CHAR-08 | `examples/identifier-similarity/main.go` adds SWG column (7→8 columns) — example output byte-stable | meta-test | `cd examples/identifier-similarity && go test ./...` | ⚠️ Wave 0 (file exists; needs extension) |
| CHAR-08 | `cross_algorithm_consistency_test.go` extends with SWG-vs-Levenshtein divergence (substring containment input) | meta-test | `go test -run TestCrossAlgorithm -v ./...` | ⚠️ Wave 0 (file exists; needs extension) |
| CHAR-08 | `llms.txt` lists all 6 new functions + `SWGParams` + `NewSWGParams` + `AlgoSmithWatermanGotoh` (the `ai_friendly_test.go` gate) | meta-test | `go test -run TestAIFriendly_LLMSTxtReferencesEveryExportedSymbol -v ./...` | ⚠️ Wave 0 (file exists; needs extension) |
| CHAR-08 | `docs/requirements.md` §7.1.8 updated to list all 6 functions | manual review | `git diff main -- docs/requirements.md` shows the §7.1.8 edit | ⚠️ Wave 0 (file exists; needs edit) |

### Sampling Rate

- **Per task commit:** `go test ./...` (full root unit + property + meta test suite, includes `TestSWG_CrossValidation`)
- **Per wave merge:** `make check` (full quality gate: lint + vet + tests + coverage ≥ 95% overall ≥ 90% per file + license-headers + verify-deps-allowlist + tidy-check + security + tests/bdd suite)
- **Phase gate:** `make check` green; `make test-bdd` green; bench.txt updated and benchstat regression check ≤ 10%; cross-platform golden run (DET-01 — Phase 1 UAT) byte-identical; full suite green before `/gsd-verify-work`.

### Wave 0 Gaps

The following test files / fixtures need to be created during Phase 3 implementation:

- [ ] `swg_test.go` — unit tests: `TestSmithWatermanGotoh_{BothEmpty, OneEmpty, Identical, ReferenceVectors, Symmetry, ScoreRunes_MultiByte, ASCII_vs_Rune_Equivalence, ZeroAllocs_ASCII_Short, WithParams, RawScore, RawScore_IdentityEqualsMatchTimesLen, CrossValidation}`
- [ ] `swg_bench_test.go` — `BenchmarkSmithWatermanGotohScore_{ASCII_Short, ASCII_Medium, ASCII_Long, Unicode_Short, WithParams_ASCII_Short, RawScore_ASCII_Short}` + `b.ReportAllocs()`
- [ ] `swg_fuzz_test.go` — `FuzzSmithWatermanGotohScore` with 7+ seeds (kitten/sitting, identity, both-empty, one-empty, invalid UTF-8, Cyrillic, one long gap)
- [ ] `testdata/fuzz/FuzzSmithWatermanGotohScore/seed-001` — one literature reference vector
- [ ] `testdata/golden/_staging/swg.json` — ~5-8 entries spanning the standard golden cases
- [ ] `testdata/cross-validation/swg/vectors.json` — ~10-20 entries spanning identity / both-empty / one-empty / two-substring / no-overlap / long-gap-canary / non-default-params
- [ ] `scripts/gen-swg-cross-validation.py` — biopython generator script (BSD-3-Clause sourced; output committed)
- [ ] `tests/bdd/features/swg.feature` — Gherkin: reference-vector outline + identity + both-empty + one-empty + symmetry + gap-split-canary + with-params
- [ ] Append to `props_test.go`: 6 standard property tests + 3 SWG-specific (gap-split-invariance, raw-bound, monotonic-with-match)
- [ ] Append to `example_test.go`: `ExampleSmithWatermanGotohScore` + `ExampleSmithWatermanGotohRawScore`
- [ ] Append to `algorithms_golden_test.go`: `TestGolden_SmithWatermanGotoh_Staging`
- [ ] Append to `algoid_test.go`: `TestDispatch_SmithWatermanGotohRegistered`
- [ ] Append to `cross_algorithm_consistency_test.go`: SWG-vs-Levenshtein divergence on substring containment
- [ ] Append to `tests/bdd/steps/algorithms_steps.go`: `iComputeTheSmithWatermanGotohScoreBetween` + Raw + WithParams variants
- [ ] Makefile: `regen-swg-cross-validation` target (developer-only)

Framework install: `python3 -m pip install --user biopython` on the regenerating developer's machine. No new Go dependencies.

## Sources

### Primary (HIGH confidence)

- **biopython official documentation (latest)** — `[CITED: https://biopython.org/docs/latest/Tutorial/chapter_pairwise.html]` — PairwiseAligner API for SWG local alignment with affine-gap params; sign convention (negative penalties); `aligner.score(a, b)` returns raw float.
- **biopython source repository** — `[CITED: https://github.com/biopython/biopython]` and `[CITED: https://github.com/biopython/biopython/blob/master/LICENSE.rst]` — BSD-3-Clause confirmed (compatible with Apache-2.0 for reference-vector cross-validation per `algorithm-licensing-standards`).
- **Flouri et al. 2015 — "Are all global alignment algorithms and implementations correct?"** — `[CITED: https://www.biorxiv.org/content/10.1101/031500v1]` — the corrected Gotoh recurrence; explicit naming of the indexing error and the initialisation error; survey numbers (5/10 implementations incorrect; 8/31 lecture slides incorrect; 16/31 incomplete).
- **`docs/requirements.md` §7.1.8** (project-internal) — locked public API signatures, default params, score normalisation formula `clamp(raw / min(la,lb), 0, 1)`, complexity bounds.
- **`.planning/research/PITFALLS.md` §3** (project-internal) — the discipline gate; explicit listing of the 4 warning signs (identity fails, gap-split asymmetry, score asymmetry, monotonicity failure).
- **`.planning/phases/03-smith-waterman-gotoh/03-CONTEXT.md`** (project-internal) — `<decisions>` §1–§7 LOCKED.
- **`.planning/phases/02-core-character-algorithms-six/02-CONTEXT.md`, `02-PATTERNS.md`, `02-01-levenshtein-SUMMARY.md`, `02-07-finalisation-SUMMARY.md`, `02-VERIFICATION.md`** (project-internal) — inherited canonical patterns.
- **`.claude/skills/algorithm-correctness-standards/SKILL.md`** (project-internal) — primary-source citation requirements, formula documentation, mandatory invariants, fresh-implementation discipline.
- **`.claude/skills/algorithm-licensing-standards/SKILL.md`** (project-internal) — Smith-Waterman / Gotoh both listed in "Clear" category; biopython BSD-3-Clause compatibility for cross-validation; fresh-implementation discipline; PR "Source origin" statement requirement.
- **`.claude/skills/performance-standards/SKILL.md`** (project-internal) — explicit SWG mention "Smith-Waterman-Gotoh: < 5 µs, 0 allocations (stack buffer fits this size)"; "Three-Matrix Gotoh Pattern" section confirms three-matrix DP is the standard.
- **`.claude/skills/determinism-standards/SKILL.md`** (project-internal) — float-stability rules; no `math.Pow`/`math.FMA`; serial left-to-right reductions only.

### Secondary (MEDIUM confidence)

- **Wikipedia — Smith-Waterman algorithm** — `[CITED: https://en.wikipedia.org/wiki/Smith%E2%80%93Waterman_algorithm]` — confirms 1981 publication, public-domain status in software; only specific hardware accelerations are patent-encumbered.
- **biopython 1.85 release notes / version page** — `[CITED: anaconda.org/anaconda/biopython]` — 1.85 listed with BSD-3-Clause as of 2026-02-26.
- **JAligner project** — `[CITED: https://jaligner.sourceforge.net/]` — confirms Smith-Waterman + Gotoh's improvement is implemented in open-source Java; provides corroborating evidence on the algorithm structure (not consulted for code).
- **bioseq-js (CDC)** — `[CITED: https://cdcgov.github.io/bioseq-js/]` — US-government public-domain Smith-Waterman+affine-gap JavaScript implementation; named for licence-compatibility context.

### Tertiary (LOW confidence — flagged for ongoing validation)

- **Gotoh, O. (1982). "An improved algorithm for matching biological sequences." Journal of Molecular Biology, 162(3):705–708** — the primary source. NOT directly retrievable in this research session (paywalled / archived); cited as the named reference in `docs/requirements.md` §7.1.8 and PITFALLS.md §3. **Risk: any direct quote of the recurrence in the implementation file's block comment should be cross-checked against Flouri et al. 2015 (which reproduces the corrected form) rather than the original.**
- **Smith, T. F., Waterman, M. S. (1981). "Identification of common molecular subsequences." Journal of Molecular Biology, 147(1):195–197** — the deeper primary source. Same provenance comment as Gotoh 1982.

## Metadata

**Confidence breakdown:**
- Standard stack: HIGH — every dependency is already in `go.mod` / `tests/bdd/go.mod`; biopython is dev-only and CONTEXT-locked.
- Architecture: HIGH — Phase 2's locked patterns transfer one-for-one; the three-matrix → two-row transcription is a derivation, not a search.
- Pitfalls: HIGH — the Gotoh erratum is well-documented; Flouri et al. 2015 is peer-reviewed; cross-validation gate is locked.
- Patent screen: HIGH — Smith-Waterman explicitly in `algorithm-licensing-standards` "Clear" list.
- Determinism: HIGH — kernel uses only deterministic float ops; matches the pattern Phase 2's `Jaro` already passes on the cross-platform matrix.
- Validation Architecture: HIGH — every test maps to an existing Phase 2 test type with one analogue.

**Research date:** 2026-05-14
**Valid until:** 2026-06-14 (30 days for stable algorithmic ecosystem; biopython API has been stable since v1.79 in 2021, no upcoming changes signalled)

## Project Constraints (from CLAUDE.md)

The following constraints from `./CLAUDE.md` and the project-skill files apply to Phase 3 implementation. The planner MUST verify each is honoured:

- **Stdlib `testing` only in root tests; no testify in root** — `swg_test.go`, `props_test.go`, `swg_bench_test.go`, `swg_fuzz_test.go` use stdlib `testing` only. testify is allowed in `tests/bdd/steps/algorithms_steps.go`.
- **Zero non-stdlib runtime require lines in root `go.mod`** beyond `golang.org/x/text` — Phase 3 introduces no Go runtime dependencies. `make verify-deps-allowlist` must remain green.
- **No cgo anywhere** — Phase 3 is pure Go. `CGO_ENABLED=0 go build ./...` must succeed.
- **Apache-2.0 file header on every new `.go` file** — `swg.go`, `dispatch_swg.go`, `swg_test.go`, `swg_bench_test.go`, `swg_fuzz_test.go` all carry the standard 13-line header. `scripts/verify-license-headers.sh` gate.
- **No GPL/LGPL-derived code anywhere** — implementation is fresh from primary sources; biopython is BSD-3-Clause; no other reference implementations are consulted.
- **Releases via CI only — no local `git tag`, no `goreleaser release`, no `--no-verify`** — Phase 3 does not release; v0.x continues. This constraint is enforced by `devops` agent.
- **GitHub issues are the source of truth — TODOs reference issues** — any `// TODO(#N):` in SWG code references a GitHub issue; planner files issues for any deferred work (e.g. EMBOSS cross-validation if it surfaces).
- **Conventional commits with issue references** — every commit must conform to `feat:` / `fix:` / `test:` / `docs:` / `chore:` / `refactor:` / `perf:` per `.claude/skills/commit-standards/SKILL.md`.
- **No `init()` functions** — `dispatch_swg.go` uses `var _ = func() bool { ... }()` per `determinism-standards` §13.5.
- **No map iteration on output paths** — SWG kernel uses slices only; no maps anywhere.
- **No transcendental float ops** — kernel uses `+ - * /` and `float64()` only; no `math.Pow`/`Exp`/`Log`/`FMA`/`Sqrt` (Sqrt is permitted by skill but not needed here).
- **No goroutines / channels / mutexes** — pure-function library.
- **Coverage floors ≥ 95% overall, ≥ 90% per file, 100% public API surface** — `swg.go` coverage must be ≥ 90%; the 6 public functions + SWGParams + NewSWGParams 100% exercised.
- **Cross-platform CI matrix determinism** — DET-01 (Phase 1 UAT carry-forward) verifies byte-identical golden output across 5 platforms. SWG golden entries must contribute correctly.
- **All agent gates apply** — algorithm-licensing-reviewer, algorithm-correctness-reviewer, algorithm-performance-reviewer, determinism-reviewer, api-ergonomics-reviewer (mandatory for the Raw* surface expansion), code-reviewer, security-reviewer, go-quality, test-writer, bdd-scenario-reviewer, docs-writer, user-guide-reviewer (mandatory for the godoc clamp warning), commit-message-reviewer (every commit), issue-writer (for any new issues), issue-closer (none expected in Phase 3 since CHAR-08 is the only issue). The `algorithm-correctness-reviewer` gate is SPECIFICALLY blocking on the biopython cross-validation evidence per PITFALLS.md §3.
