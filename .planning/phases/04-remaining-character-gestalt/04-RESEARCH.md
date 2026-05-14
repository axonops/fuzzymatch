# Phase 4: Remaining Character & Gestalt — Research

**Researched:** 2026-05-14
**Domain:** Character-based + gestalt similarity algorithms (Strcmp95, LCSStr, Ratcliff-Obershelp)
**Confidence:** HIGH

## Summary

This phase ships three fresh-implementation algorithms — **Strcmp95** (Winkler 1994
TR-2 enhancement of Jaro-Winkler), **LCSStr** (longest common substring, four
public functions), and **Ratcliff-Obershelp** (Dr. Dobb's Journal 1988, the
load-bearing Python `difflib.SequenceMatcher.ratio()` equivalent). Every
implementation discipline that Phase 4 needs is already locked by CONTEXT.md
or inherited from Phases 1–3; **the research surface for this phase is the
algorithmic content of each primary source, the cross-validation traps, and the
single high-priority finding flagged in Open Questions below**.

The phase introduces no new architectural patterns. It extends three append-
only files (`props_test.go`, `cross_algorithm_consistency_test.go`,
`examples/identifier-similarity/main.go`), adds three sets of per-algorithm
files (impl + dispatch + tests + bench + fuzz + BDD feature + staging
golden), merges three staging-golden files into `algorithms.json`, ships one
new Python cross-validation script for Ratcliff-Obershelp, and adds seven new
public function entries to `llms.txt`.

**Primary recommendation:** Plan 04-01 (Strcmp95) and plan 04-02 (LCSStr) can
proceed straight from CONTEXT.md without further discussion. Plan 04-03
(Ratcliff-Obershelp) **MUST resolve OQ-1 (the difflib-symmetry contradiction)**
before implementation lands; the question has three viable answers and the
planner should pick one explicitly. Plan 04-04 (cross-validation corpus) and
plan 04-05 (finalisation) follow the Phase 3 SWG pattern verbatim — `gen-ratcliff-obershelp-cross-validation.py` is a one-to-one structural copy of `gen-swg-cross-validation.py` with `difflib.SequenceMatcher(autojunk=False)` in place of `Bio.Align.PairwiseAligner`.

## User Constraints (from CONTEXT.md)

### Locked Decisions

| # | Decision | Where |
|---|----------|-------|
| §1 | Ratcliff-Obershelp cross-validation: Python `difflib.SequenceMatcher(autojunk=False)` (stdlib — no external Python deps); 15–18-entry committed JSON corpus covering all 4 mandatory categories (standard edge cases, Dr. Dobb's 1988 paper examples, autojunk-sensitive 200+ char case, substring/partial/unicode); Python version asserted `>= 3.7`; comparison tolerance `1e-9`; regeneration via `make regen-ratcliff-obershelp-cross-validation` (developer-only) | CONTEXT.md §1 |
| §2 | Strcmp95: full Winkler 1994 spec (Jaro + prefix boost + similar-character credit + long-string ≥ 5 chars adjustment + AS/I-S/RS-RB letter-pair adjustments); **ASCII-only**, NO `Strcmp95ScoreRunes` variant; NO `Strcmp95Params` (canonical algorithm has no consumer-tunable parameters); similar-character table transcribed by hand from Winkler 1994 TR-2 paper into a package-level `var`; Census Bureau `strcmp95.c` consulted ONLY for reference vectors; OpenRefine `Strcmp95.java` consulted ONLY for tie-breaks in prose ambiguities | CONTEXT.md §2 |
| §3 | LCSStr: **four** public functions (`LongestCommonSubstring` + `LongestCommonSubstringRunes` + `LCSStrScore` + `LCSStrScoreRunes`); score `2·len(lcs)/(len(a)+len(b))` SPEC-PINNED; **leftmost-in-`a`** tie-break for multiple equal-length matches; two-row rolling-buffer DP with max tracking; `[(maxStackInputLen+1)*2]int` stack buffer; allocation budget matches Phase 2 PERF-01 | CONTEXT.md §3 |
| §4 | Ratcliff-Obershelp: **two** public functions (`RatcliffObershelpScore` + `RatcliffObershelpScoreRunes`); no Raw* variant (always in [0,1] by construction); no params; algorithm = recursive longest-common-substring decomposition per Ratcliff & Metzener 1988; score `2·matched/(len(a)+len(b))`; **must match `difflib.SequenceMatcher(autojunk=False).ratio()` byte-for-byte within `1e-9`** on the committed corpus; godoc opens with "If you want fuzzy string matching that behaves like Python's difflib.ratio(), use this." | CONTEXT.md §4 |
| §5 | Property tests inherit the Phase 2 invariant template (RangeBounds, Identity, Symmetric, NoNaN, NoInf, NoNegativeZero); no triangle inequality (none is a metric); plus algorithm-specific extensions: Strcmp95 ≥ JaroWinkler; Strcmp95 deterministic across 1000 runs; LCSStr substring-is-subset-of-both; LCSStr length-matches-score; LCSStr leftmost-tie-break; Ratcliff-Obershelp ≥ Levenshtein on substring-containment inputs (hand-curated) | CONTEXT.md §5 |
| §6 | 5 sequential plans (no parallel waves): 04-01-strcmp95, 04-02-lcsstr, 04-03-ratcliff-obershelp, 04-04-ratcliff-obershelp-cross-validation, 04-05-finalisation | CONTEXT.md §6 |
| §7 | All Phase 2 + 3 inherited patterns apply intact: AlgoID dispatch via `var _ = func() bool{...}()` (no init()); `maxStackInputLen = 64` + `isASCII(a) && isASCII(b)` gate; staging-golden → canonical merge; per-algorithm BDD feature file; per-algorithm fuzz harness with seed-001; stdlib testing only in root; identity short-circuit `if a == b { return 1.0 }` on `*Runes` entries; BDD score regex `(\d+\.?\d*)` accepts integer-form; coverage floors ≥ 95% overall, ≥ 90% per file; llms.txt sync meta-test; Python version assertion in generator scripts; fuzz harness exercises full public surface (LCSStr: all 4 functions; RO: both functions); cross-validation evidence is the algorithm-correctness-reviewer gate (plan 04-04 must land green before plan 04-05 finalisation merges goldens) | CONTEXT.md §7 + carry_forward |

### Claude's Discretion

| # | Free choice | Trade-off |
|---|-------------|-----------|
| D-1 | Whether Strcmp95 calls an internal Jaro helper or re-derives the match-flag arrays | DRY (call jaroBytes inner) vs algorithm independence (replicate match-flag loop). Either passes Census Bureau reference vectors. |
| D-2 | Whether Ratcliff-Obershelp recursion uses the language-native call stack or an explicit stack-based iterative implementation | Either acceptable; byte-stable output is the only constraint. Native stack is simpler; explicit stack is defensive against future pathological inputs (but recursion is bounded by string length — no stack-overflow risk for reasonable inputs). |
| D-3 | Whether Ratcliff-Obershelp's inner "find longest common substring" step reuses LCSStr's internal helper or inlines the substring search | Reuse if the abstraction earns its keep; inline if the inner step needs different state (e.g. byte ranges rather than the substring itself). |
| D-4 | Exact bench label conventions (matching Phase 2/3 prefix-numbering) | Cosmetic; follow `Benchmark<Algo><Variant>_<Size>` precedent |
| D-5 | Filename: `ratcliff_obershelp.go` (underscore) vs `ratcliffobershelp.go` | Underscore is more readable for a wordier name (SWG used the short `swg.go`); precedent neutral. |

### Deferred Ideas (OUT OF SCOPE)

- EMBOSS `water` as a second cross-validation source for Strcmp95 (Census Bureau strcmp95.c is canonical; EMBOSS is for sequence alignment).
- A `Strcmp95Params` API (not in v1.0; canonical algorithm has no parameters).
- `LongestCommonSubstrings()` returning all tied-longest matches (not in v1.0; one substring committed).
- CI installation of Python 3 for re-verification of the Ratcliff-Obershelp corpus (committed JSON is the verification fixture; regeneration is developer-discretion).
- Public-API freeze for `LongestCommonSubstring` tie-break semantics — leftmost-in-`a` is documented and property-tested; changing post-v1.0 is a breaking change.

## Phase Requirements

| ID | Description | Research Support |
|----|-------------|------------------|
| CHAR-07 | Strcmp95 (Winkler 1994) with similar-character table | Public surface, table sourcing, four-adjustment algorithm, allocation budget all locked in CONTEXT.md §2; primary-source content and cross-validation traps detailed in Standard Stack + Common Pitfalls below. |
| CHAR-09 | LCSStr (longest common substring) similarity | Public surface (4 functions), score normalisation `2·len(lcs)/(len(a)+len(b))`, leftmost tie-break, two-row DP, allocation budget all locked in CONTEXT.md §3; DP recurrence verified against Wagner-Fischer 1974 below. |
| GESTALT-01 | Ratcliff-Obershelp (Dr. Dobb's Journal 1988) — difflib-equivalent | Public surface (2 functions), difflib-equivalence contract, autojunk=False gate, recursive LCSubstr algorithm all locked in CONTEXT.md §4; **OQ-1 (difflib asymmetry vs CONTEXT.md §5 Symmetric property test) MUST be resolved by the planner**. |

## Architectural Responsibility Map

| Capability | Primary Tier | Secondary Tier | Rationale |
|------------|-------------|----------------|-----------|
| Score similarity computation (Strcmp95) | root package fuzzymatch | — | Pure-function algorithm; matches Phase 2/3 pattern (e.g. JaroWinklerScore in jarowinkler.go). |
| Score similarity computation (LCSStr) | root package fuzzymatch | — | Pure-function algorithm; LongestCommonSubstring exposes the substring directly (consumer-facing utility) — same package. |
| Score similarity computation (Ratcliff-Obershelp) | root package fuzzymatch | — | Pure-function algorithm; matches gestalt section of `docs/requirements.md` §7.5. |
| AlgoID dispatch registration | root package (`dispatch_<algo>.go`) | — | Established pattern from Phase 2; each algorithm owns its dispatch slot. |
| Cross-validation corpus (RO only) | `testdata/cross-validation/ratcliff-obershelp/vectors.json` | Python generator (`scripts/gen-ratcliff-obershelp-cross-validation.py`) | Mirrors Phase 3 SWG pattern: corpus committed, generator developer-only. |
| Property tests | `props_test.go` (extend-only) | — | Established Phase 2 append point. |
| BDD scenarios | `tests/bdd/features/<algo>.feature` (new) + `tests/bdd/steps/algorithms_steps.go` (extend-only) | testify (steps only) | One feature file per algorithm — zero merge risk with prior phases. |
| Golden file extension | `testdata/golden/_staging/<algo>.json` (new) + merge into `testdata/golden/algorithms.json` | — | Established Phase 2 staging-merge pattern. |
| Cross-algorithm consistency | `cross_algorithm_consistency_test.go` (extend-only) | — | Three new tests appended: Strcmp95 ≥ JaroWinkler, LCSStr ≥ Levenshtein on substring containment, RO pinned-against-difflib. |
| Example program columns | `examples/identifier-similarity/main.go` (extend) | — | 7 → 10 column extension; `want` constant in `main_test.go` updated. |
| Benchmark baseline | `bench.txt` (full replace via `make bench`) | — | Phase 4 finalisation regenerates the file with three new algorithms' rows. |
| llms.txt sync | `llms.txt` + `llms-full.txt` (extend) | meta-test `TestLLMs_PublicSymbolsListed` | 7 new entries: 1 (Strcmp95) + 4 (LCSStr) + 2 (RO). |

## Standard Stack

### Core (already present; Phase 4 uses without modification)

| Library | Version | Purpose | Why Standard |
|---------|---------|---------|--------------|
| Go stdlib (1.26.3) | 1.26.3 | `testing`, `testing/quick`, `encoding/json`, `unicode/utf8`, native fuzz | [VERIFIED: tested in Phases 1-3, project locked] Spec-locked at `Go 1.26+` per `docs/requirements.md` §1; root `go.mod` has zero non-stdlib `require` lines beyond `golang.org/x/text`. |
| `golang.org/x/text` | v0.30.0 (latest at root go.mod) | Unicode NFC/NFD via `unicode/norm` in `normalise.go` | [VERIFIED: Phase 1 dependency] Not consumed by Phase 4 algorithms directly; Strcmp95's godoc directs Unicode users to `Normalise` upstream. |

### Supporting (BDD sub-module — already present)

| Library | Version | Purpose | When to Use |
|---------|---------|---------|-------------|
| `github.com/cucumber/godog` | v0.15.0 | BDD scenario harness | Three new feature files per algorithm. |
| `go.uber.org/goleak` | v1.3.0 | Goroutine leak detection | Already wired in `tests/bdd/TestMain`; no changes needed. |
| `github.com/stretchr/testify` | v1.10.0 | Step-definition assertion sugar | Permitted ONLY in `tests/bdd/steps/`; root tests stdlib only. |

### Cross-validation tooling (Python — developer-only, NOT a runtime or CI dependency)

| Library | Version | Purpose | Why Recommended |
|---------|---------|---------|-----------------|
| Python | 3.7+ | Generator runtime; `difflib` is stdlib | [VERIFIED: CONTEXT.md §1 locks 3.7+ for dict insertion-order preservation; mirrors Phase 3 IN-07 closure pattern] |
| Python `difflib.SequenceMatcher` | stdlib | Reference implementation for Ratcliff-Obershelp | [CITED: https://docs.python.org/3/library/difflib.html] **PSF licence** on stdlib — used for reference-vector cross-validation only, NOT for code copying. **`autojunk=False` is REQUIRED** (load-bearing — see Pitfall 2 below). |

### Alternatives Considered

| Instead of | Could Use | Tradeoff |
|------------|-----------|----------|
| Hand-transcribed Winkler 1994 similar-character table | Census Bureau strcmp95.c table values | CONTEXT.md §2 LOCKED: table values transcribed from Winkler 1994 TR-2 paper; Census Bureau strcmp95.c consulted ONLY for reference vectors. The character pairs are identical across both sources (AE, AI, AO, AU, BV, EI, EO, EU, IO, IU, OU, IY, EY, CG, EF, WU, WV, XK, SZ, XS, QC, UV, MN, LI, QO, PR, IJ, 2Z, 5S, 8B, 1I, 1L, 0O, 0Q, CK, GJ — 36 pairs), but the source-origin discipline is non-negotiable per `algorithm-licensing-standards`. |
| difflib for RO cross-validation | RapidFuzz `Ratcliff` | RapidFuzz also computes Ratcliff-Obershelp but the project's contract is to be Python-difflib-equivalent (the "common consumer mental model"). difflib is stdlib (no install friction); RapidFuzz adds external pip dep. **Decision: difflib (CONTEXT.md §1).** |
| Single-step recursive RO | Iterative stack-based | Either acceptable per CONTEXT.md D-2. Recursion is bounded by string length (~10⁴ stack frames worst-case is fine for the budget). Native call stack is simpler; defensive style is iterative. **Planner's choice.** |

**Installation (developer-only — NOT required at CI time):**

```bash
# Python 3.7+ — already on most developer machines and CI runners.
python3 --version  # must be >= 3.7

# difflib is stdlib — NO pip install needed (unlike Phase 3's biopython).
# This is a structural simplification over Phase 3.
```

**Version verification:** Already covered above. No `npm view`-equivalent step needed — difflib is Python stdlib; the generator script asserts `sys.version_info >= (3, 7)` and pins the running Python version into the corpus header, matching Phase 3 IN-07 closure pattern.

## Architecture Patterns

### System Architecture Diagram

```
┌──────────────────────────────────────────────────────────────────────┐
│                        Phase 4 Public Surface                        │
│  (one new function for Strcmp95; four for LCSStr; two for RO = 7)    │
└─────────────────┬────────────────────────────────────────────────────┘
                  │
   ┌──────────────┼──────────────────────────────┐
   ▼              ▼                              ▼
strcmp95.go    lcsstr.go                  ratcliff_obershelp.go
   │              │                              │
   │              │  ┌─────────────────────────┐ │
   │              │  │ LongestCommonSubstring  │ │  ┌────────────────────┐
   │              │  │ LongestCommonSubstring  │ │  │ RatcliffObershelp  │
   │              │  │   Runes                 │◀┼──│   Score            │
   │              │  │ LCSStrScore             │ │  │ RatcliffObershelp  │
   │              │  │ LCSStrScoreRunes        │ │  │   ScoreRunes       │
   │              │  └─────────────────────────┘ │  └────────────────────┘
   │              │              │               │            │
   │              │              ▼               │            │
   │              │  ┌─────────────────────────┐ │            │
   │              │  │ lcsstrDP (inner kernel) │ │            │
   │              │  │  • two-row rolling DP   │◀┼────────────┘ (D-3:
   │              │  │  • track max + pos      │ │       may reuse
   │              │  │  • leftmost tie-break   │ │       lcsstr's
   │              │  └─────────────────────────┘ │       inner helper)
   │              │                              │
   ▼              │                              │
┌─────────────────────────────┐                  │
│ jaroBytes (existing, jaro.go│                  │
│  D-1: Strcmp95 may call      │                  │
│  this inner helper OR        │                  │
│  re-derive match-flag arrays)│                  │
│                              │                  │
│  + Strcmp95 four adjustments:│                  │
│    1. Similar-char credit    │                  │
│    2. Winkler prefix boost   │                  │
│    3. Long-string (≥5) adj.  │                  │
│    4. AS/I-S/RS-RB pair adj. │                  │
└──────────────────────────────┘                  │
                                                  │
                       ┌──────────────────────────┘
                       ▼
       testdata/cross-validation/
        ratcliff-obershelp/vectors.json
       (15-18 entries, autojunk=False,
        consumed by TestRatcliffObershelp_
        CrossValidation; ZERO Python at CI)
                       ▲
                       │  developer-only:
                       │  `make regen-ratcliff-obershelp-cross-validation`
       scripts/gen-ratcliff-obershelp-cross-validation.py
       (stdlib difflib only; no pip install)

Dispatch registration (3 new files):
  dispatch_strcmp95.go            → dispatch[AlgoStrcmp95]            = Strcmp95Score
  dispatch_lcsstr.go              → dispatch[AlgoLCSStr]              = LCSStrScore
  dispatch_ratcliff_obershelp.go  → dispatch[AlgoRatcliffObershelp]   = RatcliffObershelpScore
```

The diagram shows three independent algorithm modules. The only **conditional**
cross-module call is the optional Strcmp95 → `jaroBytes` reuse (D-1) and the
optional Ratcliff-Obershelp → `lcsstr` inner-helper reuse (D-3). All other
dependencies are existing Phase 1 primitives: `isASCII`, `maxStackInputLen`,
`AlgoID`, `dispatch[]`, `golden_canonical.go`'s `CanonicalMarshalForTest`.

### Recommended Project Structure (Phase 4 additions only)

```
fuzzymatch/                              # root package
├── strcmp95.go                          # NEW — algorithm impl
├── dispatch_strcmp95.go                 # NEW — registration
├── strcmp95_test.go                     # NEW — unit + Winkler/Census ref vectors
├── strcmp95_bench_test.go               # NEW — ASCII_{Short,Medium,Long}
├── strcmp95_fuzz_test.go                # NEW — panic-free + [0,1] range
├── lcsstr.go                            # NEW — algorithm impl (4 funcs)
├── dispatch_lcsstr.go                   # NEW — registration
├── lcsstr_test.go                       # NEW — unit + Wagner-Fischer ref vectors + tie-break
├── lcsstr_bench_test.go                 # NEW — ASCII_{Short,Medium,Long} + Unicode_Short
├── lcsstr_fuzz_test.go                  # NEW — exercises ALL 4 functions (Phase 3 WR-02 closure)
├── ratcliff_obershelp.go                # NEW — algorithm impl
├── dispatch_ratcliff_obershelp.go       # NEW — registration
├── ratcliff_obershelp_test.go           # NEW — unit + Dr.Dobb's 1988 + TestRatcliffObershelp_CrossValidation
├── ratcliff_obershelp_bench_test.go     # NEW — ASCII_{Short,Medium,Long} + Unicode_Short
├── ratcliff_obershelp_fuzz_test.go      # NEW — panic-free + [0,1] + both Score/ScoreRunes
├── props_test.go                        # EXTEND — append three property-test blocks
├── cross_algorithm_consistency_test.go  # EXTEND — append 3 cross-algorithm tests
├── llms.txt                             # EXTEND — 7 new exported-symbol entries
├── llms-full.txt                        # EXTEND — full docstrings for 7 symbols
├── bench.txt                            # FULL REPLACE (via `make bench`) — adds Phase 4 rows
├── examples/identifier-similarity/
│   ├── main.go                          # EXTEND — 7-col → 10-col table; 3 new algorithm function refs
│   └── main_test.go                     # EXTEND — `want` constant updated
├── tests/bdd/features/
│   ├── strcmp95.feature                 # NEW — scenarios with score-regex (\d+\.?\d*)
│   ├── lcsstr.feature                   # NEW
│   └── ratcliff_obershelp.feature       # NEW — must include difflib-parity scenario
├── tests/bdd/steps/algorithms_steps.go  # EXTEND — three step blocks
├── testdata/golden/_staging/
│   ├── strcmp95.json                    # NEW — staging golden
│   ├── lcsstr.json                      # NEW — staging golden
│   └── ratcliff_obershelp.json          # NEW — staging golden
├── testdata/golden/algorithms.json      # MERGE — finalisation plan 04-05 incorporates 3 stagings
├── testdata/fuzz/FuzzStrcmp95Score/seed-001              # NEW
├── testdata/fuzz/FuzzLCSStrScore/seed-001                # NEW (one seed; fuzz body exercises all 4 funcs)
├── testdata/fuzz/FuzzRatcliffObershelpScore/seed-001     # NEW
├── testdata/cross-validation/ratcliff-obershelp/
│   └── vectors.json                     # NEW — committed corpus, 15-18 entries
├── scripts/gen-ratcliff-obershelp-cross-validation.py    # NEW — developer-only
├── Makefile                             # EXTEND — `regen-ratcliff-obershelp-cross-validation` target
└── CONTRIBUTING.md                      # EXTEND — document the new make target
```

### Pattern 1: Strcmp95 — building atop Jaro's match-flag arrays (D-1 split)

**What:** Strcmp95 = Jaro + similar-character credit + Winkler prefix boost +
long-string adjustment + AS/I-S/RS-RB letter-pair adjustments. Four adjustments
stacked atop Jaro.

**When to use:** Strcmp95Score only. There is NO `Strcmp95ScoreRunes` variant
(CONTEXT.md §2 locks ASCII-only — the similar-character table is letter-pair-keyed
and has no Unicode equivalent in Winkler 1994).

**Example structure (planner picks D-1 left or right):**

```go
// Source: Winkler, W. E. (1994). "Advanced methods for record linkage."
// Proceedings of the Section on Survey Research Methods, ASA: 467-472.
// Reference implementation: U.S. Census Bureau (1995) strcmp95.c
//   (public domain — U.S. Government work; consulted ONLY for reference
//    vectors per .claude/skills/algorithm-licensing-standards).
//
// Strcmp95 layers four adjustments atop Jaro:
//   1. Similar-character credit (Winkler 1994 §3): unmatched character
//      pairs in the Jaro match-flag arrays that appear in
//      strcmp95SimilarChars receive 0.3 credit toward the match count.
//   2. Winkler prefix boost (Winkler 1990 — already in JaroWinklerScore):
//      W = J + L · 0.1 · (1 - J)  where L = common-prefix length, max 4.
//      Applies only when J >= 0.7 (winklerBoostThreshold).
//   3. Long-string adjustment (Winkler 1994 §3): when min(la, lb) > 4 and
//      Num_com > i+1 and 2·Num_com >= minLen+i,
//      W = W + ((1 - W) · (Num_com - i - 1) / (la + lb - i)).
//   4. Similar-character substitution + position bonus (the "AS/I-S/RS-RB"
//      block in strcmp95.c): the similar-character pass adds fractional
//      match credit, normalised back into the Jaro denominator.
//
// Source-origin statement:
//   Primary: Winkler 1994 TR-2 paper for the similar-character table
//            content + the four-adjustment algorithm.
//   Cross-validation: Census Bureau strcmp95.c reference vectors
//            (public domain — U.S. Government work).
//   Tie-break: OpenRefine Strcmp95.java (Apache-2.0) for prose
//             ambiguities in Winkler 1994 (e.g. adjustment-application order).
//   GPL/LGPL: none consulted.
//   Code copied: none.

package fuzzymatch

// strcmp95SimilarChars is the upper-case ASCII letter-pair similarity table
// from Winkler 1994 TR-2 §3 "An improved string comparator". Each entry is
// bidirectional — the lookup is symmetric in (a, b) → (b, a). All published
// pairs carry similarity 0.3.
//
// Determinism (PITFALL §14): the table is a `var` declaration — NEVER built
// in init() — to ensure (1) deterministic value across init order, (2) zero
// first-call latency cost, (3) byte-stable test output. The
// determinism-reviewer flags any init() in this file as BLOCKING.
//
// Visibility: unexported and not modifiable from outside the package.
var strcmp95SimilarChars = [...]struct {
    a, b byte
    sim  float64
}{
    {'A', 'E', 0.3}, {'A', 'I', 0.3}, {'A', 'O', 0.3}, {'A', 'U', 0.3},
    {'B', 'V', 0.3}, {'E', 'I', 0.3}, {'E', 'O', 0.3}, {'E', 'U', 0.3},
    {'I', 'O', 0.3}, {'I', 'U', 0.3}, {'O', 'U', 0.3}, {'I', 'Y', 0.3},
    {'E', 'Y', 0.3}, {'C', 'G', 0.3}, {'E', 'F', 0.3}, {'W', 'U', 0.3},
    {'W', 'V', 0.3}, {'X', 'K', 0.3}, {'S', 'Z', 0.3}, {'X', 'S', 0.3},
    {'Q', 'C', 0.3}, {'U', 'V', 0.3}, {'M', 'N', 0.3}, {'L', 'I', 0.3},
    {'Q', 'O', 0.3}, {'P', 'R', 0.3}, {'I', 'J', 0.3}, {'2', 'Z', 0.3},
    {'5', 'S', 0.3}, {'8', 'B', 0.3}, {'1', 'I', 0.3}, {'1', 'L', 0.3},
    {'0', 'O', 0.3}, {'0', 'Q', 0.3}, {'C', 'K', 0.3}, {'G', 'J', 0.3},
}

func Strcmp95Score(a, b string) float64 {
    if a == b { return 1.0 }
    // (D-1) Either call internal jaroBytes helper with match-flag arrays
    // exposed back, OR re-derive the match-flag arrays here. Then apply
    // the four adjustments in order.
    //
    // The planner picks. Recommendation: re-derive — keeps Strcmp95
    // independent of jaro.go's internal layout (the alternative requires
    // jaroBytes to expose its match-flag arrays as outputs, which couples
    // its API to Strcmp95's needs).
    // ...
}
```

**Algorithm order matters:** Per the strcmp95.c reference (richmilne/JaroWinkler), the four adjustments apply in this order — base Jaro → similar-character credit (modifies the match count) → prefix boost (Winkler) → long-string adjustment (only when J > 0.7 AND min(la, lb) > 4). The OpenRefine Strcmp95.java is consulted only when prose-level ambiguity remains.

### Pattern 2: LCSStr — two-row DP + max tracking + leftmost-in-`a` tie-break

**What:** Longest common substring via dynamic programming with recurrence
`D[i,j] = D[i-1,j-1] + 1 if a[i-1] == b[j-1], else 0`. Track max value and ending position. Score normalisation: `2·len(lcs)/(len(a)+len(b))`.

**When to use:** LCSStr is the lightest of the three; reuses Phase 2's two-row
DP + ASCII-fast-path-gate pattern verbatim. The substring-returning surface is
non-standard — most similarity libraries return only a score — but justified by
the schema-similarity use case ("which substring is driving the match?").

**Example structure:**

```go
// Source: Wagner, R. A., Fischer, M. J. (1974). "The string-to-string
// correction problem." Journal of the ACM, 21(1):168-173 — standard
// DP formulation for longest common substring.
//
// Recurrence (0-indexed; cost = 1 if a[i-1] == b[j-1], else 0):
//   D[i, j] = D[i-1, j-1] + 1 if a[i-1] == b[j-1]
//             0                 otherwise
//   max_len, end_i = max over all (i, j) of D[i, j], tracking first-found.
//
// Score normalisation (SPEC-PINNED at docs/requirements.md §7.1.9):
//   score = 2·len(lcs) / (len(a) + len(b))   (Sørensen-Dice form)
//
// Edge cases:
//   - LCSStrScore("", "")     == 1.0  (both-empty convention)
//   - LCSStrScore("", "abc")  == 0.0  (one-empty)
//   - LCSStrScore("abc", "abc") == 1.0  (identity; 2·3/6 = 1)
//   - LCSStrScore with no shared chars: numerator 0 → 0.0
//
// Tie-break (LOCKED CONTEXT.md §3): when multiple longest common
// substrings of equal length exist, the LEFTMOST occurrence in `a` is
// returned. This is the natural left-to-right DP iteration order —
// `D[i, j] = max` is replaced only by STRICTLY-LARGER values, so the
// first-found-leftmost wins.

func LongestCommonSubstring(a, b string) string {
    // Two-row DP + max tracking. ASCII fast path uses [(N+1)*2]int stack
    // buffer when shorter dim n ≤ maxStackInputLen && isASCII(a) && isASCII(b).
    // ...
    return a[endI-maxLen : endI]
}

func LCSStrScore(a, b string) float64 {
    // Edge case: both-empty → 1.0; one-empty → 0.0.
    if len(a) == 0 && len(b) == 0 { return 1.0 }
    if len(a) == 0 || len(b) == 0 { return 0.0 }
    n := lcsLengthOnly(a, b)  // optional inner — avoids substring allocation
    return 2.0 * float64(n) / float64(len(a)+len(b))
}

// LongestCommonSubstringRunes — operates on []rune; same algorithm.
func LongestCommonSubstringRunes(a, b string) string { /* ... */ }
func LCSStrScoreRunes(a, b string) float64 { /* ... */ }
```

**Critical detail — strict-greater-than vs greater-than-or-equal:** The
leftmost tie-break is established by writing the max-update as `if D[i][j] > maxLen { ... }` (NOT `>=`). The `>` means the first-found-leftmost wins.
**Property test `TestProp_LongestCommonSubstring_LeftmostTieBreak` is the
load-bearing regression test for this.**

### Pattern 3: Ratcliff-Obershelp — recursive longest-common-substring decomposition

**What:** Find the longest common substring of `a` and `b`. Recursively apply
to the left-of-substring portions and right-of-substring portions. Sum matched-character count across all recursion levels. Score = `2·M / (|a| + |b|)`.

**When to use:** Whenever the consumer wants `difflib.SequenceMatcher.ratio()`
semantics. This is the load-bearing contract — Phase 6's TokenSortRatio /
TokenSetRatio / PartialRatio godoc points consumers wanting difflib here.

**Example structure:**

```go
// Source: Ratcliff, J. W., Metzener, D. E. (1988). "Pattern matching:
// the gestalt approach." Dr. Dobb's Journal, 13(7):46-51.
// Reference: Python difflib.SequenceMatcher (PSF licence — used for
//   cross-validation reference vectors only; no code copied).
//
// Algorithm (Ratcliff & Metzener 1988):
//   1. Find the longest common substring of a and b. Let its position
//      in a be [i, i+n), and in b be [j, j+n).
//   2. Recursively apply step 1 to (a[:i], b[:j]).
//   3. Recursively apply step 1 to (a[i+n:], b[j+n:]).
//   4. M = sum of n over all recursion levels.
//   5. Score = 2·M / (len(a) + len(b)).
//
// difflib-equivalence (load-bearing per docs/requirements.md §7.5.1 +
// CONTEXT.md §4): RatcliffObershelpScore MUST match
// difflib.SequenceMatcher(autojunk=False, a=a, b=b).ratio() byte-for-byte
// within 1e-9 tolerance on the committed corpus at
// testdata/cross-validation/ratcliff-obershelp/vectors.json.
//
// autojunk=False is REQUIRED: difflib's default autojunk=True is a
// performance heuristic that marks "popular" characters as junk when
// len(b) >= 200, distorting scores. The TRUE algorithm has autojunk=False.
// The 200+-character autojunk-sensitive test case in the corpus PROVES
// our impl doesn't accidentally implement an autojunk-like heuristic.
//
// asymmetry note (OQ-1): difflib.ratio() is documented as NOT symmetric
// in argument order because the recursive decomposition is left-anchored
// in `a`. Whether our impl matches that asymmetry or imposes symmetry
// (and absorbs cross-validation tolerance) is decided by the planner —
// see RESEARCH.md Open Questions §OQ-1.

func RatcliffObershelpScore(a, b string) float64 {
    if a == b { return 1.0 }  // identity short-circuit (also covers both-empty)
    if len(a) == 0 || len(b) == 0 { return 0.0 }
    m := roMatchedLength(a, b)  // recursive (or explicit-stack) helper
    return 2.0 * float64(m) / float64(len(a)+len(b))
}

func RatcliffObershelpScoreRunes(a, b string) float64 { /* ... rune-slice variant */ }
```

**RO godoc directive (CONTEXT.md §4 LOCKED — PITFALL §6 closure):**

```go
// RatcliffObershelpScore is the difflib-equivalent. If you want fuzzy
// string matching that behaves like Python's difflib.ratio(), use this.
//
// If you want the RapidFuzz "ratio()" semantics — the Indel formula
// 2·LCS/(|a|+|b|) used by Token Sort Ratio / Token Set Ratio /
// Partial Ratio — use those functions in Phase 6 instead.
```

### Anti-Patterns to Avoid

- **`init()` for the Strcmp95 table.** Use `var strcmp95SimilarChars = [...]struct{...}{...}` literal. `init()` is BLOCKING per the determinism-reviewer agent (PITFALL §14).
- **Calling `Normalise` from inside Strcmp95.** Strcmp95 is ASCII-only by design; callers normalise upstream. The godoc explicitly directs Unicode users to `fuzzymatch.Normalise`.
- **Returning a `[]string` from `LongestCommonSubstring` for tied matches.** Spec commits to ONE string (leftmost). Returning all ties is OUT-OF-SCOPE (deferred).
- **`math.Pow`, `math.Log`, `math.Exp`, `math.FMA`** anywhere in Phase 4 algorithms. Determinism-standards forbids transcendentals. Only `+`, `-`, `*`, `/`, `math.Sqrt` (NOT needed here), `math.Abs`, `math.Min`, `math.Max` (last three not needed — use if-comparison).
- **`map[byte]float64` for the Strcmp95 similar-character lookup.** A flat slice scan (36 entries; symmetric: `if (a==t.a && b==t.b) || (a==t.b && b==t.a)`) is faster than a map AND deterministic (no map iteration). For a 36-entry table the linear scan is sub-100ns and inlines.
- **Iterative-stack RO without byte-stability verification.** If D-2 picks iterative, the planner MUST property-test against the recursive baseline on a 100-pair corpus to confirm byte-identical output. (Recursive is simpler — recommend unless a specific reason emerges.)
- **`autojunk=True` (default) in the Python generator.** Single most important detail of the entire phase. Re-stated in Pitfall 2 below.

## Don't Hand-Roll

| Problem | Don't Build | Use Instead | Why |
|---------|-------------|-------------|-----|
| Strcmp95 similar-character table | Custom letter-pair similarity matrix | The Winkler 1994 TR-2 table (36 pairs at 0.3 each — table content reproduced in Pattern 1 above) | The table content is canonical and reference-vector-pinned. Any change is a major-version-bump event. |
| Longest-common-substring inner step (for RO) | Re-implement substring search from scratch | Optionally reuse LCSStr's internal helper (D-3) | If LCSStr's inner helper exposes a "find substring + position + length" signature, RO can call it. The planner decides whether the abstraction earns its keep. |
| Cross-validation reference scores for RO | Hand-compute expected scores | `difflib.SequenceMatcher(autojunk=False).ratio()` (Python stdlib) | Python difflib is THE reference implementation by community convention. The committed JSON is verified once; CI consumes it without Python. |
| Python generator infrastructure | Write from scratch | Copy structure from `scripts/gen-swg-cross-validation.py` verbatim, swap biopython for difflib | Phase 3 pattern is proven (16 entries, ~145 lines, BSD-3-Clause attribution swapped to PSF stdlib note, version-pinning preserved). |
| BDD step parsing for scores | Custom regex | `(\d+\.?\d*)` (Phase 3 IN-03 closure — accepts integer-form) | Existing step regex; reuse so feature files can use `0` and `1` without `.0` suffix. |
| Allocation-budget benchmark structure | New benchmark idiom | Phase 2 `var sink float64` + `if sink < 0 { b.Fatal(...) }` pattern | Prevents compiler dead-code elimination; established and audited. |
| Staging-golden assertion | New helper | `assertGoldenStaging` in `algorithms_golden_test.go` (LOCKED Phase 2 helper signature) | Phase 2 pattern — DO NOT change signature. |
| Two-row DP buffer sizing for LCSStr | New constant | `maxStackInputLen = 64` from `levenshtein.go` | Shared Phase 2 constant; same allocation-budget contract. |
| Identity short-circuit on rune variants | Manual rune-slice equality | `if a == b { return 1.0 }` at function entry (IN-04 closure) | Saves two `[]rune` allocations on identical inputs (including both-empty). |

**Key insight:** Phase 4 is structurally the lightest of the three algorithm
phases so far (Phase 2 had 6 algorithms + new infrastructure; Phase 3 had 1
algorithm + new cross-validation pattern; Phase 4 has 3 algorithms + zero new
patterns). The risk surface is concentrated in:
1. **OQ-1** (the difflib-symmetry contradiction — see Open Questions).
2. **Strcmp95 similar-character table transcription accuracy** (36 pairs; a
   single typo silently shifts every Strcmp95 score involving those letters).
3. **`autojunk=False` enforcement** in the Python generator (silently wrong
   scores if `autojunk=True` leaks in).

## Common Pitfalls

### Pitfall 1: Strcmp95 similar-character table transcription typos

**What goes wrong:** The Winkler 1994 similar-character table is 36 letter
pairs at 0.3 similarity each. Transcribing by hand introduces silent typos —
swapped letters, missed pairs, duplicate pairs, wrong similarity value (e.g.
0.3 → 0.03). The result: Strcmp95 scores quietly drift on inputs containing
the misspelt letters.

**Why it happens:** No machine-readable canonical source. The Census Bureau's
strcmp95.c contains the table but the project's licensing discipline forbids
copying — every value must come from the Winkler 1994 TR-2 paper.

**How to avoid:**
- The 36 pairs are: AE, AI, AO, AU, BV, EI, EO, EU, IO, IU, OU, IY, EY, CG, EF, WU, WV, XK, SZ, XS, QC, UV, MN, LI, QO, PR, IJ, 2Z, 5S, 8B, 1I, 1L, 0O, 0Q, CK, GJ. [CITED: https://github.com/richmilne/JaroWinkler/blob/master/jaro/strcmp95.c — Census Bureau public-domain reference; consulted for cross-validation vectors only per CONTEXT.md §2]
- Internal test in `strcmp95_internal_test.go`: assert `len(strcmp95SimilarChars) == 36` AND each pair appears exactly once (no duplicates) AND every similarity value is exactly 0.3.
- Cross-validate Strcmp95 output against the Census Bureau strcmp95.c reference vectors on canonical surnames (Winkler 1990 set: MARTHA/MARHTA, DWAYNE/DUANE, DIXON/DICKSONX — Strcmp95 scores should differ from JaroWinkler ONLY on pairs where the table fires).
- Property test: `TestProp_Strcmp95Score_DeterministicAcrossRuns` calls the function 1000 times with the same input → byte-identical output (PITFALL §14 closure).

**Warning signs:**
- `TestStrcmp95_ReferenceVectors_CensusBureau` fails on a specific surname → trace which letter pair fired; check transcription.
- Strcmp95Score(a, b) == JaroWinklerScore(a, b) on EVERY input → similar-character table not firing; either missing pairs or lookup logic wrong.
- Strcmp95Score(a, b) < JaroWinklerScore(a, b) on ANY input → similar-character credit decreasing the score (it must only ADD). This is the `TestProp_Strcmp95Score_AtLeastJaroWinkler` invariant.

### Pitfall 2: difflib `autojunk=True` (default) silently changes every score

**What goes wrong:** Python's `difflib.SequenceMatcher` defaults to
`autojunk=True`. The heuristic marks characters as "junk" when `len(b) >= 200`
AND a character appears in `> 1%` of `b`. Junk characters are skipped during
matching, distorting the score. If the Python generator script omits
`autojunk=False`, the corpus quietly records junk-distorted scores; our
implementation (which doesn't implement autojunk) then fails to match the
corpus — but the failure looks like "our recursion is wrong" rather than "the
corpus was generated incorrectly".

**Why it happens:** `autojunk` defaults to `True`. The keyword is hidden under
the constructor's optional arguments. A copy-paste error from a tutorial that
uses default args produces incorrect reference scores.

**How to avoid:**
- Python script MUST call `SequenceMatcher(autojunk=False, a=a, b=b).ratio()`. The `autojunk=False` kwarg is the FIRST line of `score_case()`.
- The committed corpus MUST include the 200+-character autojunk-sensitive test case (CONTEXT.md §1 Category 3). This case is constructed to differ between autojunk=True and autojunk=False — its presence in the corpus, matching our impl byte-for-byte, proves both that the Python generator runs with autojunk disabled AND that our impl doesn't have an autojunk-like heuristic.
- RatcliffObershelpScore's godoc MUST state: "behaves like `difflib.SequenceMatcher(autojunk=False, a=a, b=b).ratio()`". The qualifier is load-bearing.
- BDD scenario: include an autojunk-sensitive scenario in `ratcliff_obershelp.feature` with the same 200+-character input.

**Warning signs:**
- The 200+-character autojunk-sensitive corpus entry has `difflib_ratio ≈ 1.0` (way too high) → autojunk=True was used; popular characters were skipped.
- `TestRatcliffObershelp_CrossValidation` passes on short inputs but fails on the 200+-character entry → autojunk heuristic in either side (us or the generator).
- difflib output looks too close to identity score on long noisy inputs → autojunk firing.

[CITED: https://docs.python.org/3/library/difflib.html — "If an item's duplicates (after the first one) account for more than 1% of the sequence and the sequence is at least 200 items long, this item is marked as 'popular' and is treated as junk for the purpose of sequence matching. This heuristic can be turned off by setting the autojunk argument to False when creating the SequenceMatcher."]

### Pitfall 3: Ratcliff-Obershelp recursion finds a different decomposition than difflib

**What goes wrong:** Ratcliff-Obershelp's recursive longest-common-substring
decomposition can find multiple valid decompositions of the same input. The
final score (`2·M/T`) is the same across all decompositions in many cases, but
**not always** — because the choice of "which longest common substring to
peel off first" propagates into the recursive sub-problems' answers. Our impl
may find a different decomposition than difflib, producing a different `M`
sum and a different score.

**Why it happens:** difflib's `find_longest_match()` (CPython `Lib/difflib.py`)
uses a specific tie-break: among ties of equal length, it picks the leftmost
match in `a` AND the leftmost match in `b`. Subtle variations on tie-break
choice change which substring is peeled, which changes the left and right sub-
problems, which propagates differently.

**How to avoid:**
- The committed cross-validation corpus is the load-bearing regression test. Any decomposition divergence shows up as a `delta > 1e-9` failure with the full inputs in the error message (matching Phase 3 SWG cross-validation pattern).
- Document the difflib tie-break in `ratcliff_obershelp.go`'s godoc block: "Tie-break in `find_longest_match()`: leftmost match in `a` first, then leftmost match in `b`."
- The unit test `TestRatcliffObershelp_DrDobbs1988_ReferenceVectors` pins the canonical paper-cited examples (WIKIMEDIA/WIKIMANIA, GESTALT/GESTALT_PATTERN_MATCHING) byte-for-byte against difflib — even if the rest of the corpus passes, a regression in tie-break logic shows up here.

**Warning signs:**
- Cross-validation passes on identity / empty / no-overlap entries but fails on entries with multiple potential matches → tie-break mismatch.
- The corpus's "partial_middle_match" entry passes but a hand-curated tie-break case fails → tie-break differs from difflib.

[CITED: https://github.com/python/cpython/blob/main/Lib/difflib.py — `find_longest_match`: "find_longest_match returns the longest match. If there is a tie for longest match, the one that occurs earliest in a is preferred."]

### Pitfall 4: LCSStr leftmost-in-`a` tie-break breaks under `>=` vs `>`

**What goes wrong:** The LCSStr DP recurrence tracks the maximum value and
ending position. If the max-update is written as `if D[i][j] >= maxLen`, then
each EQUAL match overrides the previous — yielding the RIGHTMOST tied
substring. If written as `if D[i][j] > maxLen`, the first-found-leftmost wins
— yielding the LEFTMOST tied substring, which is what CONTEXT.md §3 locks.

**Why it happens:** A one-character bug (`>=` vs `>`). DP code is often
written with `>=` to "update on every match" — the leftmost-tie-break invariant
requires `>` specifically.

**How to avoid:**
- Use strict-greater-than in the max-update: `if D[i][j] > maxLen { maxLen = D[i][j]; endI = i }`.
- Property test `TestProp_LongestCommonSubstring_LeftmostTieBreak` constructs hand-curated inputs with deliberately-tied substring candidates and asserts the leftmost wins. Example: `a = "abcXYZabc"`, `b = "abc"` — both occurrences of "abc" in `a` are tied at length 3; leftmost is the first.
- Unit test pins a canonical tie-break example: `LongestCommonSubstring("abcXYZabc", "abc") == "abc"` (the first occurrence in `a`, starting at position 0).

**Warning signs:**
- `LongestCommonSubstring("abcXYZabc", "abc")` returns position 6's "abc" → the bug.
- BDD scenario "leftmost tie-break" fails.

### Pitfall 5: Strcmp95 long-string adjustment fires too late or too often

**What goes wrong:** The Winkler 1994 long-string adjustment has THREE
conditions: `min(la, lb) > 4` AND `Num_com > i+1` AND `2·Num_com >= minLen+i`,
where `i` is the common-prefix length AND `Num_com` is the post-similar-char-credit match count. Misreading any condition produces wrong scores on long inputs that share long prefixes.

**Why it happens:** Prose in Winkler 1994 mixes notation; Census Bureau
strcmp95.c expresses the conditions in compact C; OpenRefine Strcmp95.java
spells them out. Conditions may be misread from either source.

**How to avoid:**
- Reference-vector cross-validation against Census Bureau strcmp95.c on long-input pairs: e.g. long real-world surnames (HAMINGTON/HAMMINGTON, ABCDEFGHIJ/ABCDEFGHIJ, etc.). The unit test `TestStrcmp95_LongStringAdjustment_Triggers` pins at least one input where the adjustment fires (score > JaroWinkler) and one where it does NOT (length ≤ 4 or prefix-match insufficient).
- Property test `TestProp_Strcmp95Score_AtLeastJaroWinkler`: for ANY input, `Strcmp95Score(a, b) >= JaroWinklerScore(a, b)`. The adjustments can only ADD, never subtract.
- Document the three conditions inline in the algorithm file with a worked example.

**Warning signs:**
- Strcmp95Score on a long-prefix pair (e.g. "HAMINGTON"/"HAMMINGTON") equals JaroWinklerScore exactly → long-string adjustment not firing.
- Strcmp95Score on a short input (e.g. "AB"/"AC") returns more than JaroWinklerScore → long-string adjustment firing where it shouldn't (length ≤ 4 should disable it).

### Pitfall 6: LCSStr `LongestCommonSubstring` returning empty string is ambiguous

**What goes wrong:** When `a` and `b` share NO characters,
`LongestCommonSubstring("abc", "xyz")` returns `""` — but `""` is ALSO the
correct answer when `a == "" && b == ""`. Two distinct semantic cases collapse
to the same return value. Consumers cannot distinguish "no overlap" from "both
empty".

**Why it happens:** The substring-returning surface (CONTEXT.md §3's 4-function
API) inherently has this ambiguity. The score-returning surface doesn't:
`LCSStrScore("abc", "xyz") == 0.0` and `LCSStrScore("", "") == 1.0`.

**How to avoid:**
- Document explicitly in `LongestCommonSubstring`'s godoc: "Returns the empty string when (a) both inputs are empty, OR (b) the inputs share no characters. Use `LCSStrScore` to disambiguate — `LCSStrScore("", "") = 1.0` whereas `LCSStrScore("abc", "xyz") = 0.0`."
- Unit test pins both edge cases:
  ```go
  if got := LongestCommonSubstring("", ""); got != "" { t.Errorf(...) }
  if got := LongestCommonSubstring("abc", "xyz"); got != "" { t.Errorf(...) }
  ```
- Property test: `LongestCommonSubstring(a, b) == ""` IFF `a == "" || b == "" || no shared chars`. This is a documented invariant, not a bug.

**Warning signs:**
- Consumers report "LongestCommonSubstring gave me back empty string and I don't know why" → the godoc isn't explicit enough.
- A new property test pattern requires a non-empty LongestCommonSubstring for non-empty disjoint inputs → the invariant is being misread.

### Pitfall 7: BDD score regex misses Strcmp95's integer-form scores

**What goes wrong:** The Phase 2 score regex `(\d+\.\d+)` (decimal-only) was
updated in Phase 3 to `(\d+\.?\d*)` (accepts integer-form per IN-03 closure).
Phase 4 feature files MUST use the IN-03-updated regex — Strcmp95Score on
identical inputs returns `1` (which serialises as `1` not `1.0`), and BDD
scenarios that test `Then the score should be 1` need the updated regex.

**Why it happens:** Easy to copy a Phase 2 feature file and forget that the
regex was updated.

**How to avoid:**
- All three new feature files (`strcmp95.feature`, `lcsstr.feature`, `ratcliff_obershelp.feature`) use the integer-tolerant regex pattern from IN-03 closure.
- Step definitions in `tests/bdd/steps/algorithms_steps.go` already use `(\d+\.?\d*)` — the new step blocks reuse the same regex.

**Warning signs:**
- godog complains "step definition not found" for a scenario asserting `score should be 1` → regex doesn't accept integer-form.

### Pitfall 8: Three new algorithms each need `[0,1]` clamp verification on adversarial inputs

**What goes wrong:** Floating-point arithmetic can produce `1.0 + ε` or
`-ε` on adversarial inputs. Strcmp95's long-string adjustment is the highest
clamp risk (the long-string boost can push the score above 1.0 in degenerate
arithmetic cases). LCSStr's `2·n/(n+n) = 1.0` should be exact but
`2·max_n/(la+lb)` with `la == lb == max_n` requires careful float
parenthesisation. Ratcliff-Obershelp inherits this from the same `2·M/T` form.

**Why it happens:** The score formulas all involve division and multiplication
in different orders. Per DET-06, `2·M/T` must be parenthesised left-to-right:
`(2.0 * float64(M)) / float64(T)` is NOT the same byte-for-byte as
`2.0 * (float64(M) / float64(T))` on all platforms.

**How to avoid:**
- Pin all three score normalisations to left-to-right form. Example:
  ```go
  // LCSStrScore — explicit left-to-right per DET-06:
  numer := 2.0 * float64(n)  // multiplication first
  denom := float64(len(a) + len(b))
  return numer / denom
  ```
- Property tests `TestProp_<Algo>Score_RangeBounds`: assert `0.0 <= s <= 1.0` for any input.
- Cross-platform CI matrix (already in place) verifies byte-identical golden output.

**Warning signs:**
- `TestProp_<Algo>Score_RangeBounds` fails on a specific adversarial input.
- Cross-platform CI fails on arm64 but passes on amd64 → float-reduction order issue.

## Runtime State Inventory

> N/A — Phase 4 is a greenfield algorithm-implementation phase. No rename, refactor, or migration. No stored data, live service configs, OS-registered state, secrets, or build artifacts to track. The `[VERIFIED]` answer for every category is "None — verified by phase scope".

## Code Examples

### Strcmp95 — table declaration (LOCKED form per CONTEXT.md §2 + Pitfall §14)

```go
// strcmp95SimilarChars is the upper-case ASCII letter-pair similarity table
// from Winkler 1994 TR-2 §3 "An improved string comparator". Each entry is
// bidirectional: the lookup is symmetric in (a, b) → (b, a). All published
// pairs carry similarity 0.3.
//
// Source: Winkler, W. E. (1994). "Advanced methods for record linkage."
// Proceedings of the Section on Survey Research Methods, ASA: 467-472, §3.
//
// Determinism (PITFALL §14): the table is a `var` declaration with no init()
// side effect, guaranteeing byte-stable values across init order, zero
// first-call latency cost, and deterministic property-test output.
var strcmp95SimilarChars = [...]struct {
    a, b byte
    sim  float64
}{
    {'A', 'E', 0.3}, {'A', 'I', 0.3}, {'A', 'O', 0.3}, {'A', 'U', 0.3},
    {'B', 'V', 0.3}, {'E', 'I', 0.3}, {'E', 'O', 0.3}, {'E', 'U', 0.3},
    {'I', 'O', 0.3}, {'I', 'U', 0.3}, {'O', 'U', 0.3}, {'I', 'Y', 0.3},
    {'E', 'Y', 0.3}, {'C', 'G', 0.3}, {'E', 'F', 0.3}, {'W', 'U', 0.3},
    {'W', 'V', 0.3}, {'X', 'K', 0.3}, {'S', 'Z', 0.3}, {'X', 'S', 0.3},
    {'Q', 'C', 0.3}, {'U', 'V', 0.3}, {'M', 'N', 0.3}, {'L', 'I', 0.3},
    {'Q', 'O', 0.3}, {'P', 'R', 0.3}, {'I', 'J', 0.3}, {'2', 'Z', 0.3},
    {'5', 'S', 0.3}, {'8', 'B', 0.3}, {'1', 'I', 0.3}, {'1', 'L', 0.3},
    {'0', 'O', 0.3}, {'0', 'Q', 0.3}, {'C', 'K', 0.3}, {'G', 'J', 0.3},
}

// strcmp95SimilarLookup returns 0.3 if (a, b) or (b, a) is in the similar-
// character table, else 0.0. Linear scan over 36 entries; sub-100ns and
// inlines on the hot path.
func strcmp95SimilarLookup(a, b byte) float64 {
    for _, t := range strcmp95SimilarChars {
        if (a == t.a && b == t.b) || (a == t.b && b == t.a) {
            return t.sim
        }
    }
    return 0.0
}
```

[CITED: https://github.com/richmilne/JaroWinkler/blob/master/jaro/strcmp95.c — 36 pair values; consulted ONLY for cross-validation reference vectors per CONTEXT.md §2 source-origin discipline]

### LCSStr — two-row DP kernel with leftmost-tie-break

```go
// lcsstrDP runs the two-row DP recurrence and returns (length, endIndexInA).
// prev and curr must each have length n+1. Operates on string bytes directly.
//
// Leftmost-in-`a` tie-break is established by the STRICT-GREATER-THAN
// (`>`, not `>=`) max-update: first-found-leftmost wins because subsequent
// equal-length matches do NOT override.
func lcsstrDP(a, b string, m, n int, prev, curr []int) (length, endI int) {
    // prev[j] and curr[j] are initialised to zero by the caller (stack vars
    // are zero; make([]int, ...) is zero).
    var maxLen, maxEnd int
    for i := 1; i <= m; i++ {
        for j := 1; j <= n; j++ {
            if a[i-1] == b[j-1] {
                curr[j] = prev[j-1] + 1
                if curr[j] > maxLen {  // STRICT > — leftmost tie-break
                    maxLen = curr[j]
                    maxEnd = i  // exclusive end index in `a`
                }
            } else {
                curr[j] = 0  // recurrence resets on mismatch
            }
        }
        prev, curr = curr, prev
        // Clear the new `curr` for the next pass (rolling buffer reset).
        for j := 0; j <= n; j++ { curr[j] = 0 }
    }
    return maxLen, maxEnd
}
```

### Ratcliff-Obershelp — recursive longest-common-substring (CPython-style)

```go
// roMatchedLength returns the total matched-character count across the
// recursive longest-common-substring decomposition of a and b.
//
// Tie-break (matches difflib): when multiple longest common substrings of
// equal length exist, the leftmost-in-`a` is chosen; if still tied, the
// leftmost-in-`b` among those is chosen. This is the CPython difflib
// find_longest_match contract.
//
// Recursion is bounded by O(min(len(a), len(b))) depth in pathological
// cases; for typical inputs depth is O(log n).
func roMatchedLength(a, b string) int {
    if len(a) == 0 || len(b) == 0 {
        return 0
    }
    aLo, aHi, bLo, bHi, n := roFindLongestMatch(a, b)
    if n == 0 {
        return 0  // no shared characters at this level
    }
    return n +
        roMatchedLength(a[:aLo], b[:bLo]) +
        roMatchedLength(a[aHi:], b[bHi:])
}

// roFindLongestMatch returns the leftmost-longest match: (aLo, aHi, bLo, bHi, n).
// Equivalent to difflib.SequenceMatcher.find_longest_match with autojunk=False.
// Algorithm: standard LCS-substring DP with leftmost-tie-break (same as Pattern 2).
func roFindLongestMatch(a, b string) (aLo, aHi, bLo, bHi, n int) { /* ... */ }
```

### Python generator — Phase 3 SWG structure adapted for difflib

```python
#!/usr/bin/env python3
# Copyright 2026 AxonOps Limited
# ... [Apache-2.0 header omitted for brevity]

"""scripts/gen-ratcliff-obershelp-cross-validation.py

Regenerates testdata/cross-validation/ratcliff-obershelp/vectors.json by
running Python's stdlib difflib.SequenceMatcher(autojunk=False).ratio() on
a fixed list of test cases.

difflib is Python stdlib — PSF licence; used for reference-vector
cross-validation only, NOT for code copying per the project's
.claude/skills/algorithm-licensing-standards.

CRITICAL: autojunk=False is REQUIRED. The default autojunk=True is a
performance heuristic (marks "popular" characters as junk when len(b) >= 200)
that distorts scores. The TRUE Ratcliff-Obershelp algorithm has autojunk=False.
"""

import json
import os
import sys
import difflib

_MIN_PYTHON_VERSION = (3, 7)  # for dict insertion-order preservation

CASES = [
    # Category 1: Standard edge cases
    ("identity_short",     "hello",        "hello"),
    ("both_empty",         "",             ""),
    ("one_empty_a",        "",             "abcdef"),
    ("one_empty_b",        "abcdef",       ""),
    ("no_overlap",         "qqqq",         "zzzz"),
    # Category 2: Dr. Dobb's 1988 paper examples
    ("wikimedia_wikimania", "WIKIMEDIA",   "WIKIMANIA"),
    ("gestalt_paper",      "GESTALT",      "GESTALT_PATTERN_MATCHING"),
    # Category 3: autojunk-sensitive case (200+ char with 1%+ duplicates)
    ("autojunk_sensitive", "a" * 100 + "x" * 5 + "a" * 100,
                           "a" * 50 + "y" * 5 + "a" * 50 + "a" * 100),
    # Category 4: Substring + partial-match + unicode
    ("substring_middle",   "abcdef",       "xyzabcdefuvw"),
    ("partial_overlap",    "kitten",       "sitting"),
    ("unicode_ascii_only", "café",         "cafe"),
    ("longer_identity",    "the quick brown fox",  "the quick brown fox"),
    # ... ~3-6 more to reach 15-18 total
]


def score_case(a, b):
    """Compute difflib.SequenceMatcher.ratio() with autojunk=False."""
    if a == "" and b == "":
        return 1.0  # both-empty: matches our Go-side identity short-circuit
    if a == "" or b == "":
        return 0.0  # one-empty: matches our Go-side guard
    return difflib.SequenceMatcher(autojunk=False, a=a, b=b).ratio()


def _check_python_version() -> None:
    if sys.version_info < _MIN_PYTHON_VERSION:
        sys.exit(f"ERROR: Python {sys.version_info[:2]} < required "
                 f"{_MIN_PYTHON_VERSION} (for byte-stable dict order).")


def main():
    _check_python_version()
    entries = []
    for name, a, b in CASES:
        entries.append({
            "name": name,
            "a": a,
            "b": b,
            "difflib_ratio": score_case(a, b),
        })

    out = {
        "version": 1,
        "python_version": f"{sys.version_info.major}.{sys.version_info.minor}.{sys.version_info.micro}",
        "entries": entries,
    }
    path = "testdata/cross-validation/ratcliff-obershelp/vectors.json"
    os.makedirs(os.path.dirname(path), exist_ok=True)
    with open(path, "w") as f:
        json.dump(out, f, indent=2, sort_keys=False)
        f.write("\n")


if __name__ == "__main__":
    main()
```

## State of the Art

| Old Approach | Current Approach | When Changed | Impact |
|--------------|------------------|--------------|--------|
| difflib `autojunk=True` (default Python ≥ 3.2) for fuzzy-match cross-validation | difflib `autojunk=False` for true Ratcliff-Obershelp | Python 3.2 added the heuristic; CONTEXT.md §1 LOCKS the disable | Without `autojunk=False`, long-input scores silently drift; corpus generation is incorrect; our impl appears wrong when it's actually right. |
| Census Bureau strcmp95.c (1995) as primary table source | Winkler 1994 TR-2 paper as primary source; Census Bureau code as reference-vector source only | algorithm-licensing-standards LOCKS the source-of-record discipline (CONTEXT.md §2) | Pure-implementation legal posture: no derivation from C reference code; only hand-transcribed paper values. |
| Wagner-Fischer 1974 score normalisation `len(lcs)/max(len(a),len(b))` | Sørensen-Dice form `2·len(lcs)/(len(a)+len(b))` | LOCKED in docs/requirements.md §7.1.9 — SPEC-PINNED | The Sørensen-Dice form is what every modern implementation uses (it's `2·M/T` which mirrors the RO formula — natural cross-consistency). The Wagner-Fischer original `D/max(la,lb)` form would give a different, less-canonical normalisation. |
| Ratcliff-Obershelp asymmetric output (difflib documents this) | We may choose to match difflib's asymmetry OR enforce symmetry (with documented divergence) | **OPEN — OQ-1** | See Open Questions. |

**Deprecated/outdated:**
- **`init()` function for table initialisation** — outdated per Go's determinism-standards. Use `var` literal (CONTEXT.md §2 + PITFALL §14).
- **`map[string]float64` for the Strcmp95 similar-character lookup** — outdated. Map iteration order is non-deterministic; slice scan is faster on a 36-entry table.

## Assumptions Log

> All claims in this research are tagged `[VERIFIED]`, `[CITED]`, or come directly from CONTEXT.md (which is locked). No `[ASSUMED]` claims.

| # | Claim | Section | Risk if Wrong |
|---|-------|---------|---------------|

**Table is empty — all claims in this research were verified against source files, CONTEXT.md, or cited from primary documentation. No user confirmation needed.**

## Open Questions

### OQ-1 (HIGH PRIORITY — planner must resolve before plan 04-03 implementation lands)

**Question:** difflib's `SequenceMatcher.ratio()` is documented as NOT symmetric across argument order. CPython issue python/cpython#81185 (and Python tracker bpo-37004) explicitly notes that `ratio()` is "noncommutative". CONTEXT.md §5 lists `TestProp_RatcliffObershelpScore_Symmetric` as a standard inherited Phase 2 invariant. **Which wins?**

**What we know:**
- difflib.SequenceMatcher(a, b).ratio() != difflib.SequenceMatcher(b, a).ratio() on inputs like 'tide'/'diet' (0.25 vs 0.5). [CITED: https://docs.python.org/3/library/difflib.html — "Caution: The result of a ratio() call may depend on the order of the arguments."]
- This is NOT caused by autojunk — it persists with autojunk=False. The asymmetry is structural: difflib caches `b2j` (a position-mapping table for `b`) and the recursive longest-common-substring decomposition is anchored leftmost-in-`a`-first. Reversing arguments picks a different decomposition.
- CONTEXT.md §4 LOCKS: "RatcliffObershelpScore MUST match difflib.SequenceMatcher(autojunk=False, a=a, b=b).ratio() byte-for-byte within 1e-9 tolerance on the committed corpus."
- CONTEXT.md §5 LOCKS: standard property tests INHERIT FROM PHASE 2 — which includes `TestProp_<Algo>Score_Symmetric`.

**The contradiction:** If we match difflib byte-for-byte, our impl is also asymmetric — and `TestProp_RatcliffObershelpScore_Symmetric` will FAIL on inputs like 'tide'/'diet'. If we enforce symmetry (e.g. by sorting `a` and `b` by length before recursion), we DIVERGE from difflib on those same inputs — and the cross-validation corpus would need to be hand-curated to avoid asymmetric cases.

**What's unclear:** Which value is more important — the symmetry invariant (mathematically clean) OR the difflib-equivalence contract (the load-bearing consumer mental model)? **CONTEXT.md does not address this contradiction; the planner needs to pick.**

**Three viable resolutions:**

1. **Drop the Symmetric property test for Ratcliff-Obershelp.** Keep difflib byte-for-byte equivalence. Document the asymmetry in the godoc: "RatcliffObershelpScore is NOT symmetric in argument order (mirrors difflib). For symmetric similarity, sort inputs by length before calling, or use a different algorithm (e.g. LCSStr)." `cross_algorithm_consistency_test.go` adds a hand-curated asymmetric example to pin the behaviour. **PROS:** difflib-equivalence is preserved; consumer mental model intact. **CONS:** Diverges from the inherited Phase 2 invariant template; introduces a footgun documented in godoc only.

2. **Enforce symmetry by sorting arguments by length.** `RatcliffObershelpScore(a, b)` internally swaps so the longer string is always `a` (or some other canonical order). Keep the Symmetric property test. Document the divergence from difflib: "Symmetrically normalised — see godoc." Cross-validation corpus must avoid asymmetric input pairs OR include both `(a, b)` and `(b, a)` and assert our score matches difflib's MAX-of-both. **PROS:** Symmetry invariant preserved. **CONS:** Diverges from difflib byte-for-byte equivalence on asymmetric inputs; the godoc directive "behaves like difflib.ratio()" is partially false.

3. **Keep Symmetric as a "tolerance" invariant.** Replace exact symmetry with `|score(a,b) - score(b,a)| <= 0.5` (or some hand-tuned epsilon). Document. **PROS:** Compromise. **CONS:** Bogus invariant — neither difflib-equivalent NOR truly symmetric. Not recommended.

**Recommendation: Option 1.** difflib byte-for-byte equivalence is the load-bearing CONTEXT.md §4 contract — it's what the algorithm exists for. Mathematical symmetry is a "nice to have" for the property-test template but, for the specific algorithm whose entire purpose is "be the difflib-equivalent", inheriting difflib's documented asymmetry is the only coherent option. Document prominently. CONTEXT.md §5 should be amended to drop `Symmetric` from the RO inherited template. The other 5 standard property tests (RangeBounds, Identity, NoNaN, NoInf, NoNegativeZero) all still apply.

### OQ-2 (LOW priority — implementation detail)

**Question:** Should the Ratcliff-Obershelp `*Runes` variant convert to `[]rune` once at the entry point and pass `[]rune` slices through the recursion (cleaner, but each recursive call passes slice headers), OR convert to `[]rune` once and use integer index bounds throughout (slightly faster, but more error-prone)?

**What we know:** Both work; both are byte-stable. Phase 2's `levenshteinDistanceRuneSlices` pattern (pass `[]rune` slices) is the established precedent.

**Recommendation:** Use the Phase 2 precedent — pass `[]rune` slices. Recursive calls slice into the same backing array; no additional allocations.

### OQ-3 (LOW priority — Strcmp95 internal architecture decision deferred to planner per CONTEXT.md D-1)

**Question:** Should Strcmp95 call `jaroBytes` internally (DRY) or re-derive the match-flag arrays (independence)?

**What we know:** `jaroBytes` is in `jaro.go` and is unexported. Its current signature returns only the score, not the match-flag arrays. Strcmp95 needs the match-flag arrays to apply its similar-character credit AND the long-string adjustment. Reusing would require exposing match-flag arrays as outputs — coupling Jaro's internal API to Strcmp95's needs.

**Recommendation: Re-derive.** Keep Strcmp95 independent. The performance penalty is minor (the match-flag loop is O(la·w) and runs once in either case). The architectural decoupling outweighs the DRY benefit. This is consistent with Strcmp95Score and JaroWinklerScore — JaroWinkler also delegates to JaroScore (which doesn't expose match-flag arrays). Strcmp95 needs MORE than JaroScore returns, so the "re-derive" path is the natural fit.

## Environment Availability

| Dependency | Required By | Available | Version | Fallback |
|------------|------------|-----------|---------|----------|
| Go toolchain | All plans | ✓ | 1.26.3 (per go.mod directive) | — |
| `gofmt`, `go vet`, `go test -race` | All plans | ✓ | bundled with Go 1.26.3 | — |
| `golangci-lint` | All plans (`make lint`) | ✓ | v2.x (per CI; Phase 1 locked) | — |
| `python3` | plan 04-04 only (developer-only — regenerating the corpus) | varies by developer machine; default macOS has it | 3.7+ (asserted by generator script) | Same as Phase 3: `make regen-ratcliff-obershelp-cross-validation` shell-gates on `command -v python3`; absent → "skipping regen, commit existing JSON" message. CI never runs the regen target. |
| `difflib` Python module | plan 04-04 only (the generator) | ✓ — Python stdlib since 2.1 | bundled with Python 3.7+ | — (zero install friction; this is THE simplification over Phase 3 biopython) |

**Missing dependencies with no fallback:** None.
**Missing dependencies with fallback:** None.

**Critical simplification over Phase 3:** No `pip install biopython` step. difflib is stdlib. The developer-only `make regen-ratcliff-obershelp-cross-validation` target works on any machine with Python 3.7+.

## Validation Architecture

### Test Framework

| Property | Value |
|----------|-------|
| Framework | Go stdlib `testing` + `testing/quick` (root); `godog v0.15.0` + `goleak v1.3.0` + `testify v1.10.0` (`tests/bdd/`) |
| Config file | Root: none; BDD: `tests/bdd/go.mod` |
| Quick run command | `go test -run 'TestStrcmp95\|TestLCSStr\|TestRatcliffObershelp' ./...` |
| Full suite command | `make check` (golangci-lint v2 + go vet + go test -race -shuffle=on + coverage + license + deps + tidy + security) AND `make test-bdd` |

### Phase Requirements → Test Map

| Req ID | Behavior | Test Type | Automated Command | File Exists? |
|--------|----------|-----------|-------------------|--------------|
| **CHAR-07** | Strcmp95Score produces scores matching Winkler 1994 reference impl | unit | `go test -run TestStrcmp95_ReferenceVectors_Winkler1994 ./...` | ❌ Wave 0 (plan 04-01) |
| CHAR-07 | Similar-character table declared as package-level `var` (NOT init()) | static + lint | `make lint` (determinism-reviewer-pattern); plus internal test `go test -run TestStrcmp95_TableInvariants ./...` | ❌ Wave 0 (plan 04-01) |
| CHAR-07 | Strcmp95 ≥ JaroWinkler property holds | property | `go test -run TestProp_Strcmp95Score_AtLeastJaroWinkler ./...` | ❌ Wave 0 (plan 04-01) — appended to props_test.go |
| CHAR-07 | Strcmp95 deterministic across 1000 runs (PITFALL §14 closure) | property | `go test -run TestProp_Strcmp95Score_DeterministicAcrossRuns ./...` | ❌ Wave 0 (plan 04-01) |
| CHAR-07 | Strcmp95 cross-validates against Census Bureau strcmp95.c reference vectors | unit (hand-pinned) | `go test -run TestStrcmp95_CensusBureau ./...` | ❌ Wave 0 (plan 04-01) |
| CHAR-07 | Strcmp95 fuzz panic-free + score in [0,1] | fuzz | `go test -fuzz=FuzzStrcmp95Score -fuzztime=60s` | ❌ Wave 0 (plan 04-01) |
| CHAR-07 | Strcmp95 ASCII Short 0-alloc | benchmark | `go test -bench=BenchmarkStrcmp95Score_ASCII_Short -benchmem` | ❌ Wave 0 (plan 04-01) |
| CHAR-07 | Strcmp95 BDD scenarios cover identity / empty / canonical / long-string-adjustment | BDD | `make test-bdd` | ❌ Wave 0 (plan 04-01) — `tests/bdd/features/strcmp95.feature` |
| CHAR-07 | Strcmp95 byte-stable across CI matrix via algorithms.json | golden | `make verify-determinism` | ❌ Wave 0 (plan 04-01 stages; plan 04-05 merges) |
| **CHAR-09** | LCSStr returns correct LCS length on Wagner-Fischer 1974 reference vectors | unit | `go test -run TestLCSStr_ReferenceVectors_WagnerFischer1974 ./...` | ❌ Wave 0 (plan 04-02) |
| CHAR-09 | LCSStr uses two-row DP (no full table allocated) | benchmark + code review | `go test -bench=BenchmarkLCSStrScore_ASCII_Long -benchmem -benchtime=1x` (assert ≤ 2 allocs/op) | ❌ Wave 0 (plan 04-02) |
| CHAR-09 | LongestCommonSubstring tie-break = leftmost in `a` | property | `go test -run TestProp_LongestCommonSubstring_LeftmostTieBreak ./...` | ❌ Wave 0 (plan 04-02) |
| CHAR-09 | LongestCommonSubstring is substring of both inputs | property | `go test -run TestProp_LongestCommonSubstring_IsSubstringOfBoth ./...` | ❌ Wave 0 (plan 04-02) |
| CHAR-09 | Length matches score: 2·len(lcs)/(la+lb) == LCSStrScore | property | `go test -run TestProp_LongestCommonSubstring_LengthMatchesScore ./...` | ❌ Wave 0 (plan 04-02) |
| CHAR-09 | LCSStr fuzz exercises ALL 4 public functions (Phase 3 WR-02 closure) | fuzz | `go test -fuzz=FuzzLCSStrScore -fuzztime=60s` | ❌ Wave 0 (plan 04-02) |
| CHAR-09 | LCSStr ASCII Short 0-alloc; rune-path 2+2 = 4 allocs | benchmark | `go test -bench=BenchmarkLCSStr -benchmem` | ❌ Wave 0 (plan 04-02) |
| CHAR-09 | LCSStr BDD scenarios cover identity / empty / canonical / tie-break / Unicode | BDD | `make test-bdd` | ❌ Wave 0 (plan 04-02) — `tests/bdd/features/lcsstr.feature` |
| CHAR-09 | LCSStr byte-stable across CI matrix | golden | `make verify-determinism` | ❌ Wave 0 (plan 04-02 stages; plan 04-05 merges) |
| **GESTALT-01** | RatcliffObershelpScore matches difflib.SequenceMatcher(autojunk=False).ratio() byte-for-byte on Dr. Dobb's 1988 reference pairs | unit | `go test -run TestRatcliffObershelp_DrDobbs1988 ./...` | ❌ Wave 0 (plan 04-03) |
| GESTALT-01 | Numerical regression pin alongside cross-validation corpus (Phase 3 WR-03 closure) | unit | `go test -run TestRatcliffObershelp_PinnedDrDobbsValue ./...` | ❌ Wave 0 (plan 04-03 — at least one exact-value pin OUTSIDE the corpus) |
| GESTALT-01 | RatcliffObershelp cross-validates against 15-18-entry committed corpus | cross-validation | `go test -run TestRatcliffObershelp_CrossValidation ./...` | ❌ Wave 0 (plan 04-04) |
| GESTALT-01 | RO godoc explicitly contrasts with Phase 6 Indel-based ratios (PITFALL §6 closure) | docs review | grep `tests/bdd/features/ratcliff_obershelp.feature` for difflib-direction scenario; grep `ratcliff_obershelp.go` for godoc text | ❌ Wave 0 (plan 04-03) |
| GESTALT-01 | RO fuzz exercises both Score + ScoreRunes (Phase 3 WR-02 closure) | fuzz | `go test -fuzz=FuzzRatcliffObershelpScore -fuzztime=60s` | ❌ Wave 0 (plan 04-03) |
| GESTALT-01 | RO BDD scenarios cover identity / empty / canonical / 200+char autojunk-sensitive | BDD | `make test-bdd` | ❌ Wave 0 (plan 04-03) — `tests/bdd/features/ratcliff_obershelp.feature` |
| GESTALT-01 | RO byte-stable across CI matrix | golden | `make verify-determinism` | ❌ Wave 0 (plan 04-03 stages; plan 04-05 merges) |
| GESTALT-01 | autojunk-sensitive case in committed corpus proves autojunk=False | cross-validation | embedded in `TestRatcliffObershelp_CrossValidation` (entry name: `autojunk_sensitive`) | ❌ Wave 0 (plan 04-04) |
| **Cross-cutting** | Phase 4 algorithms appear in `bench.txt` baseline | benchmark | `make bench` (regenerates full bench.txt) | ❌ Wave 0 (plan 04-05) |
| Cross-cutting | Three new cross-algorithm consistency tests pass | unit | `go test -run TestCrossAlgorithm_Strcmp95_AtLeastJaroWinkler\|TestCrossAlgorithm_LCSStr_AtLeastLevenshtein_Substring\|TestCrossAlgorithm_RO_PinnedAgainstDifflib ./...` | ❌ Wave 0 (plan 04-05) — appended to cross_algorithm_consistency_test.go |
| Cross-cutting | identifier-similarity example shows three new columns | meta-test | `(cd examples/identifier-similarity && go test ./...)` | ❌ Wave 0 (plan 04-05) — `want` constant updated |
| Cross-cutting | llms.txt + llms-full.txt list all 7 new exported symbols | meta-test | `go test -run TestLLMs ./...` | ❌ Wave 0 (plan 04-05) |
| Cross-cutting | algorithms.json contains entries for all 3 new algorithms | golden | `go test -run TestGolden_Algorithms_Merge ./...` | ❌ Wave 0 (plan 04-05) |
| Cross-cutting | Apache-2.0 license header on every new .go file | static | `bash scripts/verify-license-headers.sh` | ❌ Wave 0 (every new file) |
| Cross-cutting | Zero non-stdlib runtime deps maintained | static | `bash scripts/verify-no-runtime-deps.sh` | ✓ existing |
| Cross-cutting | Coverage ≥ 95% overall, ≥ 90% per file | coverage gate | `make coverage-check` | ❌ Wave 0 (per-plan coverage compliance) |
| Cross-cutting | Makefile `regen-ratcliff-obershelp-cross-validation` target documented in CONTRIBUTING.md | meta-test | `go test -run TestMakefile_TargetsDocumentedInContributing ./...` | ❌ Wave 0 (plan 04-04) |

### Sampling Rate

- **Per task commit:** `go test -run '<plan-scope>' ./...` (per-file unit + property; <10s).
- **Per wave merge:** `make check` (full quality gate; ~60s + `make test-bdd`).
- **Phase gate:** Full suite green + `make verify-determinism` byte-identical across CI matrix + cross-validation corpus matches before `/gsd-verify-work`.

### Wave 0 Gaps

**All Phase 4 source files are new** — no Wave 0 framework or fixture gaps (the Phase 1+2+3 infrastructure is complete). The plans below ARE the Wave 0 work:

- [ ] `strcmp95.go` + `dispatch_strcmp95.go` + `strcmp95_test.go` + `strcmp95_bench_test.go` + `strcmp95_fuzz_test.go` — covers CHAR-07 (plan 04-01)
- [ ] `lcsstr.go` + `dispatch_lcsstr.go` + `lcsstr_test.go` + `lcsstr_bench_test.go` + `lcsstr_fuzz_test.go` — covers CHAR-09 (plan 04-02)
- [ ] `ratcliff_obershelp.go` + `dispatch_ratcliff_obershelp.go` + `ratcliff_obershelp_test.go` + `ratcliff_obershelp_bench_test.go` + `ratcliff_obershelp_fuzz_test.go` — covers GESTALT-01 algorithm path (plan 04-03)
- [ ] `scripts/gen-ratcliff-obershelp-cross-validation.py` + `testdata/cross-validation/ratcliff-obershelp/vectors.json` + `TestRatcliffObershelp_CrossValidation` appended to `ratcliff_obershelp_test.go` + Makefile target + CONTRIBUTING.md doc — covers GESTALT-01 difflib-equivalence gate (plan 04-04)
- [ ] `props_test.go` appends, `cross_algorithm_consistency_test.go` appends, `examples/identifier-similarity/main.go` extension, `bench.txt` regenerate, `llms.txt`/`llms-full.txt` extensions, `testdata/golden/algorithms.json` merge — covers cross-cutting requirements (plan 04-05)
- [ ] Per-algorithm BDD feature files (`strcmp95.feature`, `lcsstr.feature`, `ratcliff_obershelp.feature`) + `tests/bdd/steps/algorithms_steps.go` extends — covers GESTALT-01/CHAR-07/CHAR-09 BDD (in the per-algorithm plans)
- [ ] Per-algorithm staging goldens (`_staging/strcmp95.json`, `_staging/lcsstr.json`, `_staging/ratcliff_obershelp.json`) + on-disk fuzz seeds (`testdata/fuzz/Fuzz<Algo>Score/seed-001`)

## Security Domain

### Applicable ASVS Categories

| ASVS Category | Applies | Standard Control |
|---------------|---------|-----------------|
| V2 Authentication | no | Library has no auth surface. |
| V3 Session Management | no | Library is stateless / pure-function. |
| V4 Access Control | no | Library is stateless / pure-function. |
| V5 Input Validation | yes | Each algorithm MUST handle malformed UTF-8 + adversarial-length inputs without panic. Fuzz tests are the enforcement gate. |
| V6 Cryptography | no | No crypto in this phase. |
| V7 Error Handling | yes | No `error` returns from Phase 4 functions (consistent with the catalogue); inputs that "should fail" return deterministic-but-meaningless float (e.g. score 0.0 on disjoint input). No info leak via error messages. |
| V13 Resilience | yes | No goroutines, no I/O — denial-of-service vector is algorithmic complexity (worst-case O(n²) for RO, O(m·n) for LCSStr / Strcmp95). Documented in godoc. |

### Known Threat Patterns for {pure-Go stdlib + Go 1.26.3}

| Pattern | STRIDE | Standard Mitigation |
|---------|--------|---------------------|
| UTF-8 panic via malformed input | D (Denial of Service) | Fuzz tests on every public function (`FuzzStrcmp95Score`, `FuzzLCSStrScore`, `FuzzRatcliffObershelpScore`); ≥ 60s each in `make test-fuzz`; seed corpus includes `\xff\xfe` and lone-surrogate sequences. Established Phase 2 pattern. |
| Algorithmic complexity attack (long inputs) | D (Denial of Service) | Document worst-case in godoc; benchmark suite covers Long (500+ char) inputs; PITFALL §12 closure. RO worst case is O(n²·m) — flagged in godoc. |
| Score-deserialisation drift across versions | T (Tampering of test fixtures) | `testdata/golden/algorithms.json` byte-stable across CI matrix; any score change triggers a deliberate PR with version-bump justification (PITFALL §15). |
| Reference-implementation contamination | I (Information disclosure of GPL-licensed code) | algorithm-licensing-reviewer gate: only MIT/BSD/Apache + public-domain U.S. Census Bureau code consulted. No GPL/LGPL. Source-origin statement required in every file's block comment (PITFALL §16). |
| Patent ambush on a new algorithm | I (Information disclosure of patent encumbrance) | Phase 4 algorithms are pre-screened (Strcmp95 + LCSStr are unencumbered academic algorithms; Ratcliff-Obershelp is from 1988 academic publication — well past patent applicability). PITFALL §17 closure: per-algorithm screen already in the spec catalogue. |
| Float-determinism drift across architectures | T (Tampering of float reduction order) | Explicit left-to-right parenthesisation; no `math.Pow` / `math.Log` / `math.Exp` / `math.FMA`; cross-platform CI matrix verifies byte-identical golden output. PITFALL §9 closure. |

## Sources

### Primary (HIGH confidence — Verified)

- **CONTEXT.md** (`.planning/phases/04-remaining-character-gestalt/04-CONTEXT.md`) — all locked decisions for the phase
- **REQUIREMENTS.md** (`.planning/REQUIREMENTS.md`) — CHAR-07, CHAR-09, GESTALT-01 IDs and traceability
- **ROADMAP.md** (`.planning/ROADMAP.md`) — Phase 4 goal, depends-on (Phase 3), success criteria
- **PITFALLS.md** (`.planning/research/PITFALLS.md`) — §6 (token-vs-difflib), §7 (byte-vs-rune), §9 (float determinism), §10 (NaN/Inf), §14 (tables via var not init), §15 (score stability), §16 (GPL/LGPL contamination), §17 (patent screen)
- **docs/requirements.md** — §7.1.7 (Strcmp95 API), §7.1.9 (LCSStr API, normalisation SPEC-PINNED), §7.5.1 (Ratcliff-Obershelp API), §13.5 (no init() side effects), §13.6 (property tests), §14.1 (perf budgets), §15.1-15.5 (test strategy)
- **Phase 2 patterns** (`.planning/phases/02-core-character-algorithms-six/02-CONTEXT.md` + `02-PATTERNS.md` + `02-07-finalisation-SUMMARY.md`) — file structure, ASCII fast-path gate, staging-merge pattern, cross-algorithm consistency tests, bench.txt baseline pattern
- **Phase 3 patterns** (`.planning/phases/03-smith-waterman-gotoh/03-CONTEXT.md` + `03-02-swg-cross-validation-SUMMARY.md`) — cross-validation evidence path, Python generator + committed corpus pattern (Phase 4 replicates with difflib)
- **Existing source files:** `algoid.go`, `jaro.go`, `jarowinkler.go`, `levenshtein.go`, `swg.go`, `swg_test.go`, `dispatch_swg.go`, `algorithms_golden_test.go`, `props_test.go`, `cross_algorithm_consistency_test.go`, `examples/identifier-similarity/main.go`, `scripts/gen-swg-cross-validation.py`

### Secondary (MEDIUM-HIGH confidence — Cited)

- [Python docs — difflib](https://docs.python.org/3/library/difflib.html) — `SequenceMatcher`, `autojunk` parameter, `ratio()` formula `2.0*M/T`, "Caution: noncommutative" note
- [CPython `Lib/difflib.py`](https://github.com/python/cpython/blob/main/Lib/difflib.py) — `find_longest_match` tie-break (leftmost in `a` then leftmost in `b`)
- [Census Bureau strcmp95.c reference](https://github.com/richmilne/JaroWinkler/blob/master/jaro/strcmp95.c) — similar-character table (36 pairs); algorithm structure; consulted ONLY for reference vectors per CONTEXT.md §2
- [Abydos `_strcmp95.py`](https://abydos.readthedocs.io/en/latest/_modules/abydos/distance/_strcmp95.html) — secondary confirmation of the 36-pair similar-character table (Apache-2.0; not consulted for code)

### Tertiary (LOW confidence — flagged for resolution)

- **OQ-1 (difflib symmetry vs CONTEXT.md §5 invariant template):** the contradiction is documented; recommendation provided (Option 1: drop Symmetric for RO); **planner decides**.
- bpo-37004 / python/cpython#81185 — "SequenceMatcher.ratio() noncommutativity not well-documented" — secondary confirmation that asymmetry is real and documented, not just an interpretation mistake.

## Metadata

**Confidence breakdown:**

- Standard stack: HIGH — all dependencies already in place from Phases 1-3; difflib is Python stdlib (verified docs); Go 1.26.3 verified.
- Architecture: HIGH — Phase 4 introduces zero new patterns. Every architectural decision is inherited from Phase 2 + 3 and locked in CONTEXT.md.
- Pitfalls: HIGH — all 8 documented pitfalls trace to specific lines in CONTEXT.md, PITFALLS.md, or cited Python/Census Bureau sources. OQ-1 is the one HIGH-priority unresolved item — flagged explicitly.
- Cross-validation: HIGH — the Phase 3 SWG cross-validation pattern is byte-stable, audited, and proved out by 16 entries matching biopython with delta=0. Phase 4 replicates the structure for difflib (simpler — stdlib, no install friction).
- Security: HIGH — pure-function library, no auth/session/crypto; threats are limited to UTF-8 panic (fuzz-tested), complexity-attack (documented), float-determinism (cross-platform CI).

**Research date:** 2026-05-14
**Valid until:** 2026-06-14 (estimate — 30 days for a stable phase; difflib semantics are unlikely to drift; CONTEXT.md locked decisions are the canonical source-of-truth).
