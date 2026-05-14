# Phase 4: Remaining Character & Gestalt — Pattern Map

**Mapped:** 2026-05-14
**Files analyzed:** 25 new files + 11 extension points (across 5 plans)
**Analogs found:** 25 / 25 new (all have strong Phase 2 or Phase 3 analogs); 11 / 11 extensions (all extend existing append-only files with established Phase 2/3 shapes)

> Phase 4 introduces **zero new architectural patterns.** Every file in this phase
> has a load-bearing analog in Phases 2 or 3. This file maps each new/extended
> file to its closest analog, names the concrete code excerpt to copy, and lists
> deviations forced by Phase 4's algorithm specifics.

---

## File Classification

### Plan 04-01 — Strcmp95

| New / Modified File | Role | Data Flow | Closest Analog | Match Quality |
|---------------------|------|-----------|----------------|---------------|
| `strcmp95.go` | algorithm | transform (pure-fn) | `/Users/johnny/Development/fuzzymatch/jarowinkler.go` | exact (thin wrapper atop Jaro, four extra adjustments stacked) |
| `dispatch_strcmp95.go` | dispatch registration | — | `/Users/johnny/Development/fuzzymatch/dispatch_swg.go` | exact |
| `strcmp95_test.go` | test (unit) | — | `/Users/johnny/Development/fuzzymatch/swg_test.go` (header + reference-vector + symmetry sections only; no cross-validation block) | exact |
| `strcmp95_bench_test.go` | test (benchmark) | — | `/Users/johnny/Development/fuzzymatch/swg_bench_test.go` (Short/Medium/Long; **drop** WithParams and RawScore variants; **drop** Unicode_Short — Strcmp95 is byte-only by §2) | role-match |
| `strcmp95_fuzz_test.go` | test (fuzz) | — | `/Users/johnny/Development/fuzzymatch/swg_fuzz_test.go` (collapse 6 surfaces to 1 surface; same seed-table shape) | role-match |
| `tests/bdd/features/strcmp95.feature` | BDD feature | — | `/Users/johnny/Development/fuzzymatch/tests/bdd/features/swg.feature` | exact |
| `testdata/golden/_staging/strcmp95.json` | golden staging | — | `/Users/johnny/Development/fuzzymatch/testdata/golden/_staging/swg.json` | exact (same schema; entries computed from `Strcmp95Score`) |
| `testdata/fuzz/FuzzStrcmp95Score/seed-001` | fuzz seed | — | `/Users/johnny/Development/fuzzymatch/testdata/fuzz/FuzzSmithWatermanGotohScore/seed-001` | exact |

### Plan 04-02 — LCSStr

| New / Modified File | Role | Data Flow | Closest Analog | Match Quality |
|---------------------|------|-----------|----------------|---------------|
| `lcsstr.go` | algorithm | transform (pure-fn, 2-row DP) | `/Users/johnny/Development/fuzzymatch/levenshtein.go` | exact (same two-row DP + ASCII fast-path gate; 4 public funcs vs 4 in levenshtein.go) |
| `dispatch_lcsstr.go` | dispatch registration | — | `/Users/johnny/Development/fuzzymatch/dispatch_swg.go` | exact |
| `lcsstr_test.go` | test (unit + property) | — | `/Users/johnny/Development/fuzzymatch/swg_test.go` (drop cross-validation block; ADD tie-break property test + LongestCommonSubstring-returns-empty-string unit tests) | role-match |
| `lcsstr_bench_test.go` | test (benchmark) | — | `/Users/johnny/Development/fuzzymatch/swg_bench_test.go` (Short/Medium/Long + Unicode_Short; ADD benches for the 3 sibling surfaces: LongestCommonSubstring + ScoreRunes + LongestCommonSubstringRunes) | role-match |
| `lcsstr_fuzz_test.go` | test (fuzz) | — | `/Users/johnny/Development/fuzzymatch/swg_fuzz_test.go` (use Phase 3 WR-02 multi-surface pattern; loop 4 surfaces over each input) | exact |
| `tests/bdd/features/lcsstr.feature` | BDD feature | — | `/Users/johnny/Development/fuzzymatch/tests/bdd/features/swg.feature` (include tie-break + Unicode scenarios) | exact |
| `testdata/golden/_staging/lcsstr.json` | golden staging | — | `/Users/johnny/Development/fuzzymatch/testdata/golden/_staging/swg.json` | exact |
| `testdata/fuzz/FuzzLCSStrScore/seed-001` | fuzz seed | — | `/Users/johnny/Development/fuzzymatch/testdata/fuzz/FuzzSmithWatermanGotohScore/seed-001` | exact |

### Plan 04-03 — Ratcliff-Obershelp (algorithm only; cross-validation in 04-04)

| New / Modified File | Role | Data Flow | Closest Analog | Match Quality |
|---------------------|------|-----------|----------------|---------------|
| `ratcliff_obershelp.go` | algorithm | transform (recursive longest-common-substring) | `/Users/johnny/Development/fuzzymatch/swg.go` (file structure, godoc shape, fast-path discipline) | role-match (same file shape; different DP — recursive LCSubstr, no two-row DP buffer) |
| `dispatch_ratcliff_obershelp.go` | dispatch registration | — | `/Users/johnny/Development/fuzzymatch/dispatch_swg.go` | exact |
| `ratcliff_obershelp_test.go` | test (unit + Dr. Dobb's pins; cross-validation block APPENDED in plan 04-04) | — | `/Users/johnny/Development/fuzzymatch/swg_test.go` (drop SWGParams sections; drop alloc-gate tests since RO recursion has no stack-buffer path; ADD numerical regression pin per Phase 3 WR-03 closure) | role-match |
| `ratcliff_obershelp_bench_test.go` | test (benchmark) | — | `/Users/johnny/Development/fuzzymatch/swg_bench_test.go` (Short/Medium/Long + Unicode_Short for both Score + ScoreRunes; **drop** WithParams and RawScore — RO has neither) | role-match |
| `ratcliff_obershelp_fuzz_test.go` | test (fuzz) | — | `/Users/johnny/Development/fuzzymatch/swg_fuzz_test.go` (Phase 3 WR-02 multi-surface pattern; loop Score + ScoreRunes) | exact |
| `tests/bdd/features/ratcliff_obershelp.feature` | BDD feature | — | `/Users/johnny/Development/fuzzymatch/tests/bdd/features/swg.feature` (ADD 200+-char autojunk-sensitive scenario; ADD asymmetry-pin scenario per OQ-1 resolution) | role-match |
| `testdata/golden/_staging/ratcliff_obershelp.json` | golden staging | — | `/Users/johnny/Development/fuzzymatch/testdata/golden/_staging/swg.json` | exact |
| `testdata/fuzz/FuzzRatcliffObershelpScore/seed-001` | fuzz seed | — | `/Users/johnny/Development/fuzzymatch/testdata/fuzz/FuzzSmithWatermanGotohScore/seed-001` | exact |

### Plan 04-04 — Ratcliff-Obershelp cross-validation

| New / Modified File | Role | Data Flow | Closest Analog | Match Quality |
|---------------------|------|-----------|----------------|---------------|
| `scripts/gen-ratcliff-obershelp-cross-validation.py` | tooling (developer-only) | — | `/Users/johnny/Development/fuzzymatch/scripts/gen-swg-cross-validation.py` | exact (1-to-1 structural copy; swap biopython → stdlib difflib; swap version guard) |
| `testdata/cross-validation/ratcliff-obershelp/vectors.json` | committed corpus | — | `/Users/johnny/Development/fuzzymatch/testdata/cross-validation/swg/vectors.json` | exact (same shape; `biopython_*` → `difflib_*` field renames) |
| `ratcliff_obershelp_test.go` (append cross-validation block) | test (cross-validation) | — | `/Users/johnny/Development/fuzzymatch/swg_test.go` lines 411–479 (`TestSWG_CrossValidation`) | exact |
| `Makefile` (append target) | build tooling | — | `/Users/johnny/Development/fuzzymatch/Makefile` lines 196–211 (`regen-swg-cross-validation`) | exact |
| `CONTRIBUTING.md` (append doc line) | docs | — | `/Users/johnny/Development/fuzzymatch/CONTRIBUTING.md` line 92 (`make regen-swg-cross-validation` entry) | exact |

### Plan 04-05 — Finalisation (existing append-only files)

| Existing File | Extension Shape | Closest Analog (within the same file) |
|---------------|-----------------|----------------------------------------|
| `props_test.go` | append 3 property-test blocks (Strcmp95, LCSStr, RatcliffObershelp); NB **omit** `Symmetric` for RO per OQ-1 | `props_test.go` lines 737–810 (SWG block — 6 standard invariants); lines 812–909 (3 SWG-specific extensions) |
| `cross_algorithm_consistency_test.go` | append 3 new tests + 1 RO-asymmetry pin | `cross_algorithm_consistency_test.go` lines 205–226 (`TestCrossAlgorithm_SWG_Levenshtein_SubstringDivergence`) |
| `examples/identifier-similarity/main.go` | extend `algorithms` slice 7 → 10 entries; update `algoWidth` if needed | `main.go` lines 73–84 (the `algorithms` slice; entries 1–7) |
| `examples/identifier-similarity/main_test.go` | regenerate `want` constant (one `go run .` regen) | `main_test.go` lines 41–50 (the `want` constant; line-by-line diff helper is reusable as-is) |
| `tests/bdd/steps/algorithms_steps.go` | append 3 step-method blocks + 3 `ctx.Step(...)` regex registrations | `algorithms_steps.go` lines 290–316 (SWG step methods); lines 443–455 (SWG step regex registrations) |
| `algorithms_golden_test.go` | append `buildStrcmp95StagingEntries`, `buildLCSStrStagingEntries`, `buildRatcliffObershelpStagingEntries`, three `TestGolden_*_Staging` tests; extend `TestGolden_Algorithms_Merge`'s `stagingFiles` slice | `algorithms_golden_test.go` lines 584–656 (`buildSWGStagingEntries` + `TestGolden_SWG_Staging`); lines 162–195 (`TestGolden_Algorithms_Merge`) |
| `algoid_test.go` | append 3 `TestDispatch_<Algo>Registered` tests; extend `TestDispatch_UnregisteredSlotsAreNil`'s registered-slots map | `algoid_test.go` lines 284–292 (`TestDispatch_SmithWatermanGotohRegistered`); lines 299–323 (`TestDispatch_UnregisteredSlotsAreNil`) |
| `example_test.go` | append 7 `ExampleXxx` functions (1 Strcmp95 + 4 LCSStr + 2 RatcliffObershelp) | `example_test.go` lines 108–122 (the SWG `ExampleSmithWatermanGotohScore` + `ExampleSmithWatermanGotohRawScore` block) |
| `bench.txt` | full-replace via `make bench` | — (no per-file analog; the file IS the artefact) |
| `llms.txt` / `llms-full.txt` | append 7 new exported-symbol entries | `llms.txt` lines 87–94 (SWG entries — `SWGParams`, `NewSWGParams`, 6 functions) |
| `testdata/golden/algorithms.json` | merge 3 new staging files into the canonical file (via `TestGolden_Algorithms_Merge` with `-update`) | — (merge is automatic once the 3 new staging files exist and the merge test's `stagingFiles` slice lists them) |

---

## Pattern Assignments

### `strcmp95.go` (Plan 04-01)

**Analog:** `/Users/johnny/Development/fuzzymatch/jarowinkler.go` (185 lines — exact analog: a thin algorithm built atop Jaro). Secondary analog for the `var`-level table: this is a NEW shape — Phase 2 has no equivalent constant-table file. Pattern given in RESEARCH.md lines 696–735 is canonical.

**Imports + license header pattern** (copy from `jarowinkler.go` lines 1–13; then file-level doc block per `swg.go` lines 15–82):

```go
// Copyright 2026 AxonOps Limited
// [...Apache-2.0 header...]

// strcmp95.go implements Winkler's Strcmp95 enhancement of Jaro-Winkler.
//
// Sources:
//   - Winkler, W. E. (1994). "Advanced methods for record linkage."
//     Proceedings of the Section on Survey Research Methods, ASA: 467-472.
//   - Cross-validation reference: U.S. Census Bureau strcmp95.c (public
//     domain — U.S. Government work; consulted ONLY for reference vectors
//     per .claude/skills/algorithm-licensing-standards).
//   - OpenRefine Strcmp95.java (Apache-2.0) consulted ONLY for prose-level
//     tie-breaks in Winkler 1994; no code copied.
//
// Strcmp95 layers four adjustments atop Jaro: similar-character credit,
// Winkler prefix boost, long-string adjustment, AS/I-S/RS-RB block.
// [...full algorithm godoc per RESEARCH.md Pattern 1...]
//
// Source-origin statement:
//   Primary: Winkler 1994 TR-2 paper.
//   Cross-validation: Census Bureau strcmp95.c (public domain).
//   Tie-break: OpenRefine Strcmp95.java (Apache-2.0).
//   GPL/LGPL: none consulted.
//   Code copied: none.

package fuzzymatch
```

**Similar-character table pattern (NEW shape — no Phase 2/3 analog):** copy verbatim from RESEARCH.md lines 706–735, including the `var` declaration discipline (PITFALLS §14 — NO `init()`):

```go
// strcmp95SimilarChars is the upper-case ASCII letter-pair similarity table
// from Winkler 1994 TR-2 §3 "An improved string comparator".
//
// Determinism (PITFALL §14): the table is a `var` declaration with no init()
// side effect — determinism-reviewer flags any init() in this file as BLOCKING.
var strcmp95SimilarChars = [...]struct {
    a, b byte
    sim  float64
}{
    {'A', 'E', 0.3}, {'A', 'I', 0.3}, /* ... 36 entries from RESEARCH.md ... */
}

// strcmp95SimilarLookup returns 0.3 if (a, b) or (b, a) is in the table.
func strcmp95SimilarLookup(a, b byte) float64 {
    for _, t := range strcmp95SimilarChars {
        if (a == t.a && b == t.b) || (a == t.b && b == t.a) {
            return t.sim
        }
    }
    return 0.0
}
```

**Public function pattern** (copy file shape from `jarowinkler.go`; godoc shape from `swg.go` lines 130–158 — opens with the layering paragraph per RESEARCH.md "Specifics"):

```go
// Strcmp95Score returns Winkler's Strcmp95 similarity between a and b in
// [0.0, 1.0]. Strcmp95 = Jaro + similar-character credit + Winkler prefix
// boost + long-string adjustment + AS/I-S/RS-RB letter-pair adjustments.
//
// API hierarchy (the load-bearing consumer mental model):
//   - JaroScore        — base similarity.
//   - JaroWinklerScore — Jaro + prefix boost (shared-prefix bias).
//   - Strcmp95Score    — Jaro-Winkler + similar-character credit + adjustments
//                        (record-linkage / surname matching).
//
// Strcmp95 operates on ASCII letters. For non-ASCII input, normalise via
// fuzzymatch.Normalise first (NFC/NFD + diacritic folding). There is NO
// Strcmp95ScoreRunes variant — the similar-character table is letter-pair-
// keyed and has no Unicode equivalent in Winkler 1994.
//
// Edge cases:
//   - Strcmp95Score("", "") == 1.0 (both-empty identity)
//   - Strcmp95Score("", "abc") == 0.0 (one-empty)
//   - Strcmp95Score(x, x) == 1.0 for any non-empty x
//   - Strcmp95Score(a, b) == Strcmp95Score(b, a) (symmetric)
//   - Strcmp95Score(a, b) >= JaroWinklerScore(a, b) (always — adjustments
//     can only add, never subtract)
func Strcmp95Score(a, b string) float64
```

**Deviations from `jarowinkler.go`:**

- NO `*Runes` variant (CONTEXT.md §2 lock).
- Adds the `strcmp95SimilarChars` table + lookup helper.
- Re-derives match-flag arrays per RESEARCH.md OQ-3 recommendation (Strcmp95 needs the match-flag arrays for the similar-character credit pass; `jaroBytes` doesn't return them). Or call `jaroBytes` and accept the duplicated work — D-1 is planner's choice.
- Apply the four adjustments in the order specified in RESEARCH.md Pattern 1: base Jaro → similar-character credit (modifies match count) → prefix boost → long-string adjustment.

**Source-origin statement:** mandatory at top of file per `algorithm-licensing-standards`. Pattern: see `swg.go` lines 18–26 (its multi-source attribution block).

---

### `dispatch_strcmp95.go` (Plan 04-01)

**Analog:** `/Users/johnny/Development/fuzzymatch/dispatch_swg.go` (33 lines — full file is the template).

**Full file pattern** (lines 1–33; substitute `SmithWatermanGotoh` → `Strcmp95`):

```go
// Copyright 2026 AxonOps Limited
// [...Apache-2.0 header...]

// dispatch_strcmp95.go registers Strcmp95Score into the dispatch table at
// package load time. This file MUST be the sole writer to
// dispatch[AlgoStrcmp95] (slot 5).
//
// See algoid.go for the dispatch array declaration. The var _ = func() bool
// { ... }() idiom is the canonical form for package-level side effects
// without init() (per determinism-standards §13.5 and
// docs/requirements.md §5(12)).

package fuzzymatch

var _ = func() bool {
    dispatch[AlgoStrcmp95] = Strcmp95Score
    return true
}()
```

**Gotcha:** `algoid.go` already declares `AlgoStrcmp95` at slot 5. This file is the SOLE writer to that slot — no `init()`, no other file touches `dispatch[AlgoStrcmp95]`.

---

### `strcmp95_test.go` (Plan 04-01)

**Analog:** `/Users/johnny/Development/fuzzymatch/swg_test.go` lines 1–325 (everything BEFORE the cross-validation block). Drop the SWG-specific sections (SWGParams construction, Raw* surface, alloc-gate); keep the canonical-reference-vector + symmetry + edge-case shape.

**Header + import block** (copy `swg_test.go` lines 1–36):

```go
// Copyright 2026 AxonOps Limited
// [...Apache-2.0 header...]

// strcmp95_test.go pins the public-API contract of strcmp95.go: identity,
// both-empty, one-empty, canonical reference vectors from Winkler 1994 +
// the Census Bureau strcmp95.c reference set (MARTHA/MARHTA, DWAYNE/DUANE,
// DIXON/DICKSONX), symmetry, the similar-character table invariants
// (PITFALLS §14), and the Strcmp95 ≥ JaroWinkler hierarchy property.
//
// Stdlib `testing` only — no testify in root tests.

package fuzzymatch_test

import (
    "testing"

    "github.com/axonops/fuzzymatch"
)
```

**Reference-vector test pattern** (copy `swg_test.go` lines 38–54 shape; reference vectors from Census Bureau strcmp95.c per RESEARCH.md Pitfall 1):

```go
// TestStrcmp95_ReferenceVectors_CensusBureau pins canonical Winkler 1994 /
// Census Bureau strcmp95.c surnames.
func TestStrcmp95_ReferenceVectors_CensusBureau(t *testing.T) {
    tests := []struct {
        a, b string
        want float64
    }{
        {"MARTHA", "MARHTA", 0.9611}, // Winkler 1990 (Strcmp95 == JaroWinkler here — no similar pair fires)
        {"DWAYNE", "DUANE",  0.840},  // Census Bureau strcmp95.c
        {"DIXON",  "DICKSONX", 0.8133}, // Census Bureau strcmp95.c
        // ...+ at least one input where the similar-character table fires,
        // per Pitfall 1 ("Strcmp95Score == JaroWinklerScore on every input
        // means the table isn't firing").
    }
    for _, tt := range tests {
        t.Run(tt.a+"_"+tt.b, func(t *testing.T) {
            const tol = 1e-3
            got := fuzzymatch.Strcmp95Score(tt.a, tt.b)
            if d := got - tt.want; d > tol || d < -tol {
                t.Errorf("Strcmp95Score(%q, %q) = %g; want %g (tol %g)", tt.a, tt.b, got, tt.want, tol)
            }
        })
    }
}
```

**Table-invariants test pattern** (NEW shape — no Phase 2/3 analog; per RESEARCH.md Pitfall 1):

```go
// TestStrcmp95_TableInvariants asserts the similar-character table is well-
// formed: exactly 36 entries, every entry has similarity 0.3, no duplicate
// pairs (counting (a,b) and (b,a) as the same pair).
//
// NOTE: the table is unexported; if this test cannot reach it directly via
// the export_test.go re-export pattern, then the assertions are smoked
// indirectly: call Strcmp95Score on 36 hand-picked input pairs (one per
// table entry) and assert each diverges from JaroWinklerScore.
func TestStrcmp95_TableInvariants(t *testing.T) { /* ... */ }
```

**Hierarchy + alloc-gate test patterns:** copy `swg_test.go` lines 384–409 (alloc-gate using `testing.AllocsPerRun`); ADD a `TestStrcmp95_AtLeastJaroWinkler_OnReferenceVectors` smoke test.

**Deviations from `swg_test.go`:**

- DROP: `TestSmithWatermanGotoh_NewSWGParams_Defaults`, `TestSmithWatermanGotoh_WithCustomParams`, `TestSmithWatermanGotoh_RawScore_UnclampedSemantics`, `TestSmithWatermanGotoh_ScoreWithHighMatch_ClampsUpper`, `TestSmithWatermanGotoh_GapSplitCanary`, `TestSWG_CrossValidation` (Strcmp95 has no params, no Raw, no gap-split, no Python cross-validation).
- DROP: `TestSmithWatermanGotoh_RuneMultiByte`, `TestSmithWatermanGotoh_ByteVsRune_Equivalence` (no `*Runes` variant per CONTEXT.md §2).
- ADD: `TestStrcmp95_TableInvariants` (PITFALLS §14 closure).
- ADD: at least one input pair from the Census Bureau set where the similar-character table provably fires (to lock Pitfall 1's warning sign #2).

---

### `strcmp95_bench_test.go` (Plan 04-01)

**Analog:** `/Users/johnny/Development/fuzzymatch/swg_bench_test.go` (148 lines — exact analog for shape).

**Header + per-bench shape pattern** (copy `swg_bench_test.go` lines 1–67; rename `SmithWatermanGotohScore` → `Strcmp95Score`):

```go
// Copyright 2026 AxonOps Limited
// [...Apache-2.0 header...]

// strcmp95_bench_test.go runs allocation-aware benchmarks for Strcmp95Score
// at three input sizes. b.ReportAllocs() on every benchmark gates allocation
// regressions in bench.txt via benchstat.
//
// Performance budgets:
//   - ASCII Short (≤ 64 bytes):  target < 2 µs/op, 0 allocs/op
//   - ASCII Medium (50 bytes):   target 0 allocs/op
//   - ASCII Long (> 64 bytes):   may allocate for match-flag arrays
//
// NB: Strcmp95 is byte-only by §2 — there is NO Unicode_Short benchmark.

package fuzzymatch_test

// BenchmarkStrcmp95Score_ASCII_Short uses the same var sink + b.Fatal
// pattern as swg_bench_test.go.
func BenchmarkStrcmp95Score_ASCII_Short(b *testing.B) {
    b.ReportAllocs()
    b.ResetTimer()
    var sink float64
    for i := 0; i < b.N; i++ {
        sink = fuzzymatch.Strcmp95Score("MARTHA", "MARHTA")
    }
    if sink < 0 {
        b.Fatal("sink unexpectedly negative — compiler folded the benchmark away")
    }
}
```

**Deviations:**

- DROP: `Unicode_Short`, `WithParams_ASCII_Short`, `RawScore_ASCII_Short` benches.
- KEEP: Short, Medium, Long.
- Total bench count: 3 (vs 6 for SWG).

---

### `strcmp95_fuzz_test.go` (Plan 04-01)

**Analog:** `/Users/johnny/Development/fuzzymatch/swg_fuzz_test.go` (122 lines).

**Shape pattern** (copy `swg_fuzz_test.go` lines 1–122; collapse the 6-surface loop to a 1-surface direct call):

```go
// Copyright 2026 AxonOps Limited
// [...Apache-2.0 header...]

// strcmp95_fuzz_test.go runs native Go fuzzing against Strcmp95Score.
// Properties: panic-free, finite (no NaN/Inf), in [0.0, 1.0].
//
// Single surface (no Runes, no Params, no Raw) — the multi-surface pattern
// from swg_fuzz_test.go collapses to a single direct call here.

package fuzzymatch_test

import (
    "math"
    "testing"

    "github.com/axonops/fuzzymatch"
)

func FuzzStrcmp95Score(f *testing.F) {
    for _, pair := range []struct{ a, b string }{
        {"MARTHA", "MARHTA"},          // Winkler 1990 canonical
        {"DWAYNE", "DUANE"},           // Census Bureau
        {"abc", "abc"},                // identical
        {"", "abc"},                   // one-empty
        {"", ""},                      // both-empty
        {"\xff\xfe", "abc"},           // invalid UTF-8 — Strcmp95 is byte-only
        {"\xc0\x80", "abc"},           // invalid UTF-8 (overlong NUL)
        {"HAMINGTON", "HAMMINGTON"},   // long-string-adjustment trigger (Pitfall 5)
        {"AB", "AC"},                  // length-≤4 (long-string adj should NOT fire)
    } {
        f.Add(pair.a, pair.b)
    }
    f.Fuzz(func(t *testing.T, a, b string) {
        got := fuzzymatch.Strcmp95Score(a, b)
        if math.IsNaN(got) { t.Errorf("Strcmp95Score(%q, %q) = NaN", a, b) }
        if math.IsInf(got, 0) { t.Errorf("Strcmp95Score(%q, %q) = Inf", a, b) }
        if got < 0.0 || got > 1.0 { t.Errorf("Strcmp95Score(%q, %q) = %g; want in [0,1]", a, b, got) }
    })
}
```

**On-disk corpus pattern** (copy `testdata/fuzz/FuzzSmithWatermanGotohScore/seed-001` shape verbatim — first line `go test fuzz v1`; one `string(...)` argument per line per fuzz parameter):

```
go test fuzz v1
string("MARTHA")
string("MARHTA")
```

---

### `lcsstr.go` (Plan 04-02)

**Analog:** `/Users/johnny/Development/fuzzymatch/levenshtein.go` (271 lines — exact analog for the two-row DP + ASCII fast-path gate; identical kernel shape).

**Header + file-level doc** (copy `levenshtein.go` lines 1–60 shape; substitute Wagner-Fischer 1974 LCS-substring recurrence per RESEARCH.md Pattern 2):

```go
// Copyright 2026 AxonOps Limited
// [...Apache-2.0 header...]

// lcsstr.go implements the Longest Common Substring similarity for the
// fuzzymatch catalogue. FOUR public functions are exported:
//   - LongestCommonSubstring(a, b string) string
//   - LongestCommonSubstringRunes(a, b string) string
//   - LCSStrScore(a, b string) float64
//   - LCSStrScoreRunes(a, b string) float64
//
// Source: Wagner, R. A., Fischer, M. J. (1974). "The string-to-string
// correction problem." Journal of the ACM, 21(1):168-173 — standard DP
// formulation for longest common substring.
//
// Recurrence (0-indexed; D[i,j] = D[i-1,j-1] + 1 if a[i-1] == b[j-1] else 0):
//   max_len, end_i = max over all (i, j) of D[i, j], FIRST-FOUND-LEFTMOST.
//
// Score normalisation (SPEC-PINNED at docs/requirements.md §7.1.9):
//   score = 2·len(lcs) / (len(a) + len(b))   (Sørensen-Dice form)
//
// Tie-break (LOCKED CONTEXT.md §3): when multiple longest common substrings
// of equal length exist, the LEFTMOST occurrence in `a` is returned. This
// is enforced by the STRICT-GREATER-THAN (`>`, not `>=`) max-update.
//
// LongestCommonSubstring returning the empty string is AMBIGUOUS: it
// returns "" both when both inputs are empty AND when the inputs share no
// characters. Use LCSStrScore to disambiguate: LCSStrScore("","")=1.0 vs
// LCSStrScore("abc","xyz")=0.0. See RESEARCH.md Pitfall 6.
//
// Implementation discipline (inherits Phase 2):
//   - ASCII fast path operates on string bytes directly when n ≤
//     maxStackInputLen && isASCII(a) && isASCII(b); stack-allocated
//     [(maxStackInputLen+1)*2]int buffer holds the two rolling rows.
//     (maxStackInputLen is defined in levenshtein.go — DO NOT redeclare.)
//   - NO init()-time table builds.
//   - NO map iteration on output paths.
//   - NO transcendental float operations (DET-06).

package fuzzymatch
```

**Score normalisation (left-to-right per DET-06 — RESEARCH.md Pitfall 8):**

```go
func LCSStrScore(a, b string) float64 {
    if a == b { return 1.0 } // identity short-circuit (covers both-empty + identical)
    la, lb := len(a), len(b)
    if la == 0 || lb == 0 { return 0.0 }
    n := lcsstrLengthOnly(a, b)
    // Explicit left-to-right parenthesisation per DET-06:
    numer := 2.0 * float64(n)
    denom := float64(la + lb)
    return numer / denom
}
```

**Two-row DP kernel + leftmost-tie-break** (copy `levenshtein.go` levenshteinDP shape; substitute LCSStr recurrence per RESEARCH.md Pattern 2 lines 748–769):

```go
// lcsstrDP returns (length, endIndexInA). prev and curr must each be n+1.
// Leftmost-in-`a` tie-break: STRICT `>` (NOT `>=`) — first-found wins.
func lcsstrDP(a, b string, m, n int, prev, curr []int) (length, endI int) {
    var maxLen, maxEnd int
    for i := 1; i <= m; i++ {
        for j := 1; j <= n; j++ {
            if a[i-1] == b[j-1] {
                curr[j] = prev[j-1] + 1
                if curr[j] > maxLen { // STRICT > — leftmost tie-break
                    maxLen = curr[j]
                    maxEnd = i // exclusive end index in a
                }
            } else {
                curr[j] = 0 // recurrence resets on mismatch
            }
        }
        prev, curr = curr, prev
        for j := 0; j <= n; j++ { curr[j] = 0 }
    }
    return maxLen, maxEnd
}
```

**ASCII fast-path gate** (LOCKED v1.0 idiom from Phase 2 PATTERNS Pattern 3 + `levenshtein.go` lines 165–182):

```go
if n <= maxStackInputLen && isASCII(a) && isASCII(b) {
    var buf [(maxStackInputLen + 1) * 2]int
    return lcsstrDP(a, b, m, n, buf[:n+1], buf[n+1:n+n+2])
}
return lcsstrDP(a, b, m, n, make([]int, n+1), make([]int, n+1))
```

**Deviations from `levenshtein.go`:**

- 4 public functions (vs 4 for Levenshtein: Distance/DistanceRunes/Score/ScoreRunes); SAME count, different names.
- Returns `string` not `int` for the substring-returning variants.
- Recurrence is "reset on mismatch" (not Levenshtein's edit-cost min).
- The Runes-variant identity short-circuit is the IN-04 closure pattern: `if a == b { return ... }` BEFORE `[]rune(a)`.

---

### `dispatch_lcsstr.go` (Plan 04-02)

**Analog:** `/Users/johnny/Development/fuzzymatch/dispatch_swg.go`. Exact same shape; substitute slot index/name.

```go
package fuzzymatch

var _ = func() bool {
    dispatch[AlgoLCSStr] = LCSStrScore
    return true
}()
```

**Gotcha:** Only `LCSStrScore` is dispatched. `LongestCommonSubstring*` and `LCSStrScoreRunes` are public but not dispatched (dispatch table maps AlgoID → (a, b string) float64 only). RESEARCH.md notes this explicitly.

---

### `lcsstr_test.go` (Plan 04-02)

**Analog:** `/Users/johnny/Development/fuzzymatch/swg_test.go` lines 1–409 (drop cross-validation block at lines 411+).

**Reference-vector pattern** (copy `swg_test.go` table-driven shape from lines 38–101):

Tests to include:
1. `TestLCSStr_BothEmpty` — `LongestCommonSubstring("","") == ""`, `LCSStrScore("","") == 1.0`.
2. `TestLCSStr_OneEmpty` — `LongestCommonSubstring("","abc") == ""`, `LCSStrScore("","abc") == 0.0`.
3. `TestLCSStr_NoOverlap_DisambiguationPin` — pins the Pitfall 6 ambiguity: `LongestCommonSubstring("abc","xyz") == ""` AND `LCSStrScore("abc","xyz") == 0.0` (the empty-string return is documented behaviour, NOT a bug).
4. `TestLCSStr_ReferenceVectors_WagnerFischer1974` — canonical pairs from Wagner-Fischer 1974.
5. `TestLCSStr_LeftmostTieBreak_Pinned` — `LongestCommonSubstring("abcXYZabc", "abc") == "abc"` (the FIRST occurrence). This is the load-bearing regression test for the strict-`>` recurrence (RESEARCH.md Pitfall 4).
6. `TestLCSStr_ByteVsRune_Equivalence` (copy `swg_test.go` lines 162–186 pattern).
7. `TestLCSStr_RuneMultiByte` (copy `swg_test.go` lines 188–218 pattern; assert `LongestCommonSubstringRunes("café", "cafe") == "caf"` and `LCSStrScoreRunes("café", "cafe") ≈ 0.75`).
8. `TestLCSStrScore_ZeroAllocs_ASCII_Short` (copy `swg_test.go` lines 384–409 pattern; same `testing.AllocsPerRun(100, ...)` form).
9. `TestLCSStrScore_ZeroAllocs_ASCII_Medium` (same shape).

**Deviations:**

- DROP: every SWGParams test, every Raw* test, every cross-validation test.
- ADD: leftmost-tie-break pin (load-bearing for RESEARCH.md Pitfall 4); empty-return disambiguation pin (Pitfall 6).

---

### `lcsstr_bench_test.go` (Plan 04-02)

**Analog:** `/Users/johnny/Development/fuzzymatch/swg_bench_test.go`. Drop the WithParams + RawScore variants; ADD benches for the 3 sibling public functions.

Benchmarks to land (`Benchmark<Function>_<Size>`):

- `BenchmarkLCSStrScore_{ASCII_Short, ASCII_Medium, ASCII_Long, Unicode_Short}` (4 — exact `swg_bench_test.go` Short/Medium/Long/Unicode_Short shape)
- `BenchmarkLongestCommonSubstring_{ASCII_Short, ASCII_Medium, ASCII_Long, Unicode_Short}` (4 — substring-returning surface)
- `BenchmarkLCSStrScoreRunes_Unicode_Short` (1)
- `BenchmarkLongestCommonSubstringRunes_Unicode_Short` (1)

Total: 10 benches.

**Gotcha:** the `LongestCommonSubstring*` benches must use `var sink string` (not `float64`) and check `if len(sink) < 0 { b.Fatal(...) }` — adapt the `swg_bench_test.go` sink pattern for the substring-returning surface.

---

### `lcsstr_fuzz_test.go` (Plan 04-02)

**Analog:** `/Users/johnny/Development/fuzzymatch/swg_fuzz_test.go` (Phase 3 WR-02 multi-surface pattern; closure for CHAR-09 in RESEARCH.md §Test Map).

**Shape pattern** (copy `swg_fuzz_test.go` lines 56–122; replace the 6 SWG surfaces with the 4 LCSStr surfaces):

```go
func FuzzLCSStrScore(f *testing.F) {
    for _, pair := range []struct{ a, b string }{
        {"kitten", "sitting"},
        {"abc", "abc"},
        {"", ""},
        {"", "abc"},
        {"abcXYZabc", "abc"},       // leftmost-tie-break (Pitfall 4)
        {"abc", "xyz"},             // no-overlap empty-return (Pitfall 6)
        {"\xff\xfe", "abc"},        // invalid UTF-8
        {"café", "cafe"},           // multi-byte UTF-8 (rune surface)
        {"Привет", "привет"},       // Cyrillic
    } {
        f.Add(pair.a, pair.b)
    }
    f.Fuzz(func(t *testing.T, a, b string) {
        // Property: every public surface is panic-free.
        // Scores must be in [0, 1]; substrings must not panic.
        type scoreSurface struct {
            name string
            val  float64
        }
        scores := []scoreSurface{
            {"LCSStrScore",      fuzzymatch.LCSStrScore(a, b)},
            {"LCSStrScoreRunes", fuzzymatch.LCSStrScoreRunes(a, b)},
        }
        // Exercise substring-returning surfaces — they MUST NOT panic.
        _ = fuzzymatch.LongestCommonSubstring(a, b)
        _ = fuzzymatch.LongestCommonSubstringRunes(a, b)
        for _, s := range scores {
            if math.IsNaN(s.val)     { t.Errorf("%s(%q,%q) = NaN", s.name, a, b) }
            if math.IsInf(s.val, 0)  { t.Errorf("%s(%q,%q) = Inf", s.name, a, b) }
            if s.val < 0 || s.val > 1 { t.Errorf("%s(%q,%q) = %g; want in [0,1]", s.name, a, b, s.val) }
        }
    })
}
```

---

### `ratcliff_obershelp.go` (Plan 04-03)

**Analog:** `/Users/johnny/Development/fuzzymatch/swg.go` (540 lines — file structure; godoc shape with multi-source attribution). Note: RO does NOT use a two-row DP buffer (recursive LCS-substring decomposition); SWG's DP discipline is the wrong template, but the source-attribution + edge-case + ASCII fast-path discipline IS the right template.

**Header + multi-source attribution** (copy `swg.go` lines 1–82 structure; substitute Ratcliff & Metzener 1988 per RESEARCH.md Pattern 3):

```go
// Copyright 2026 AxonOps Limited
// [...Apache-2.0 header...]

// ratcliff_obershelp.go implements the Ratcliff-Obershelp gestalt pattern-
// matching similarity for the fuzzymatch catalogue.
//
// Source: Ratcliff, J. W., Metzener, D. E. (1988). "Pattern matching: the
// gestalt approach." Dr. Dobb's Journal, 13(7):46-51.
//
// Cross-validation reference: Python difflib.SequenceMatcher (PSF licence
// on stdlib — used for reference-vector cross-validation only, NOT for
// code copying per .claude/skills/algorithm-licensing-standards).
//
// Algorithm:
//  1. Find the longest common substring of a and b at positions
//     a[aLo:aHi] and b[bLo:bHi], where aHi-aLo == bHi-bLo == n.
//  2. Recursively apply to (a[:aLo], b[:bLo]).
//  3. Recursively apply to (a[aHi:], b[bHi:]).
//  4. M = sum of n across all recursion levels.
//  5. Score = 2·M / (len(a) + len(b)).
//
// difflib-equivalence (load-bearing): RatcliffObershelpScore MUST match
// difflib.SequenceMatcher(autojunk=False, a=a, b=b).ratio() byte-for-byte
// within 1e-9 tolerance on the committed corpus at
// testdata/cross-validation/ratcliff-obershelp/vectors.json.
//
// autojunk=False is REQUIRED: difflib's default autojunk=True is a
// performance heuristic that distorts scores on inputs ≥ 200 chars. Our
// implementation has no autojunk-equivalent; the 200+-char corpus entry
// proves this.
//
// ASYMMETRY (CONTEXT.md §4 + OQ-1 resolution): difflib.ratio() is NOT
// symmetric in argument order — the recursive decomposition is left-
// anchored in `a`. RatcliffObershelpScore inherits this asymmetry to
// preserve byte-for-byte difflib equivalence. For symmetric similarity,
// use LCSStrScore.
//
// Source-origin statement:
//   Primary: Ratcliff & Metzener 1988 (Dr. Dobb's Journal).
//   Cross-validation: Python difflib (PSF licence on stdlib).
//   GPL/LGPL: none consulted.
//   Code copied: none.

package fuzzymatch
```

**Public function pattern** (open with the difflib-equivalence directive — PITFALLS §6 closure per RESEARCH.md Pattern 3):

```go
// RatcliffObershelpScore is the difflib-equivalent. If you want fuzzy string
// matching that behaves like Python's difflib.ratio(), use this.
//
// If you want the RapidFuzz "ratio()" semantics — the Indel formula
// 2·LCS/(|a|+|b|) used by Token Sort Ratio / Token Set Ratio / Partial
// Ratio — use those functions in Phase 6 instead.
//
// Behaves like difflib.SequenceMatcher(autojunk=False, a=a, b=b).ratio().
//
// Edge cases:
//   - RatcliffObershelpScore("", "") == 1.0 (both-empty identity)
//   - RatcliffObershelpScore("", "abc") == 0.0 (one-empty)
//   - RatcliffObershelpScore(x, x) == 1.0 for any non-empty x
//   - RatcliffObershelpScore is NOT symmetric in argument order (mirrors
//     difflib). For symmetric similarity, sort inputs by length first or
//     use LCSStrScore.
func RatcliffObershelpScore(a, b string) float64 {
    if a == b { return 1.0 } // identity short-circuit (covers both-empty + identical)
    if len(a) == 0 || len(b) == 0 { return 0.0 }
    m := roMatchedLength(a, b)
    // Left-to-right per DET-06:
    return 2.0 * float64(m) / float64(len(a)+len(b))
}

func RatcliffObershelpScoreRunes(a, b string) float64 {
    if a == b { return 1.0 } // IN-04 identity short-circuit BEFORE []rune alloc
    ra := []rune(a)
    rb := []rune(b)
    if len(ra) == 0 || len(rb) == 0 { return 0.0 }
    m := roMatchedLengthRunes(ra, rb)
    return 2.0 * float64(m) / float64(len(ra)+len(rb))
}
```

**Recursive longest-common-substring helper** (copy RESEARCH.md Pattern 3 code at lines 785–801 verbatim — the recursive variant; OQ-2 lock):

```go
func roMatchedLength(a, b string) int {
    if len(a) == 0 || len(b) == 0 { return 0 }
    aLo, aHi, bLo, bHi, n := roFindLongestMatch(a, b)
    if n == 0 { return 0 }
    return n +
        roMatchedLength(a[:aLo], b[:bLo]) +
        roMatchedLength(a[aHi:], b[bHi:])
}

// roFindLongestMatch returns the leftmost-longest match equivalent to
// difflib.SequenceMatcher.find_longest_match (autojunk=False): leftmost in
// `a` first; if still tied, leftmost in `b`.
func roFindLongestMatch(a, b string) (aLo, aHi, bLo, bHi, n int) { /* DP — same recurrence as LCSStr; track aLo/bLo as well as n */ }
```

**Deviations from `swg.go`:**

- NO two-row DP buffer (recursive LCS-substring decomposition, no rolling rows).
- NO params API, no Raw* variant, no `*WithParams` surface.
- NO stack-buffer ASCII fast path (recursion has no DP-table allocation to optimise).
- Inherits SWG's source-origin + multi-source-attribution block discipline.
- Inherits the identity-short-circuit + `*Runes` `if a == b { return 1.0 }` pattern (IN-04 closure).

**Optional D-3 lever:** the planner may either re-use `lcsstr.go`'s internal helper for finding the longest common substring, OR inline the LCS-substring search in `ratcliff_obershelp.go`. RESEARCH.md OQ-3 leaves this open; either is acceptable.

---

### `dispatch_ratcliff_obershelp.go` (Plan 04-03)

**Analog:** `/Users/johnny/Development/fuzzymatch/dispatch_swg.go`. Substitute slot index/name.

```go
package fuzzymatch

var _ = func() bool {
    dispatch[AlgoRatcliffObershelp] = RatcliffObershelpScore
    return true
}()
```

---

### `ratcliff_obershelp_test.go` (Plan 04-03 + appended in 04-04)

**Analog:** `/Users/johnny/Development/fuzzymatch/swg_test.go` (482 lines).

**Plan 04-03 content:**

- Header + import block (copy `swg_test.go` lines 1–36).
- `TestRatcliffObershelp_BothEmpty` (copy lines 38–54 shape).
- `TestRatcliffObershelp_OneEmpty` (copy lines 56–77 shape).
- `TestRatcliffObershelp_Identical` (copy lines 79–102 shape).
- `TestRatcliffObershelp_DrDobbs1988_ReferenceVectors` — canonical pairs (`WIKIMEDIA`/`WIKIMANIA`, `GESTALT`/`GESTALT_PATTERN_MATCHING`) per CONTEXT.md §1 Category 2. Same table-driven `t.Run(tt.a+"_"+tt.b, ...)` shape as `swg_test.go` reference-vector tests.
- `TestRatcliffObershelp_PinnedDrDobbsValue` — at least ONE exact-value pin from Dr. Dobb's 1988 in this file (per Phase 3 WR-03 closure: numerical regression pin alongside cross-validation corpus, not solely via corpus). Pattern from `swg_test.go` lines 250–271 (`TestSmithWatermanGotoh_WithCustomParams`'s numerical pin).
- `TestRatcliffObershelp_AsymmetryPin` — pin a known asymmetric pair from CPython issue python/cpython#81185 (e.g. `tide`/`diet` returns different scores depending on argument order). Per OQ-1 resolution Option 1: this is documented + tested, not enforced symmetric.
- `TestRatcliffObershelp_ByteVsRune_Equivalence` (copy `swg_test.go` lines 162–186 shape — ASCII-only inputs must agree).
- `TestRatcliffObershelp_RuneMultiByte` (copy `swg_test.go` lines 188–218 shape — `café`/`cafe`).

**Plan 04-04 appends:** `TestRatcliffObershelp_CrossValidation` — direct one-to-one copy of `/Users/johnny/Development/fuzzymatch/swg_test.go` lines 411–479 (`TestSWG_CrossValidation`). Substitutions:

- Path: `testdata/cross-validation/ratcliff-obershelp/vectors.json`.
- Corpus struct fields: `Params paramsBlock` REMOVED; `BiopythonScore` → `DifflibRatio`; `BiopythonVersion` → `PythonVersion`.
- Function called: `fuzzymatch.RatcliffObershelpScore(e.A, e.B)` (no params).
- Tolerance: `1e-9` (matches Phase 3 epsilon).

**Deviations from `swg_test.go`:**

- DROP: All SWGParams / Raw* / WithParams tests, the gap-split canary test, the clamp-engagement tests (RO has no clamp), the alloc-gate tests (RO has no stack-buffer path to gate).
- ADD: Dr. Dobb's 1988 paper-cited reference vectors (per Pitfall 3 — load-bearing on tie-break).
- ADD: numerical pin for at least one Dr. Dobb's 1988 value (Phase 3 WR-03 closure).
- ADD: asymmetric-pair pin (OQ-1 resolution Option 1: document asymmetry via test, not enforce symmetry).

---

### `ratcliff_obershelp_bench_test.go` (Plan 04-03)

**Analog:** `/Users/johnny/Development/fuzzymatch/swg_bench_test.go`. Drop WithParams + RawScore variants.

Benchmarks: `Benchmark<Function>_<Size>`:

- `BenchmarkRatcliffObershelpScore_{ASCII_Short, ASCII_Medium, ASCII_Long, Unicode_Short}` (4)
- `BenchmarkRatcliffObershelpScoreRunes_Unicode_Short` (1)

Total: 5 benches.

---

### `ratcliff_obershelp_fuzz_test.go` (Plan 04-03)

**Analog:** `/Users/johnny/Development/fuzzymatch/swg_fuzz_test.go`. Adapt the multi-surface pattern for 2 surfaces (Score + ScoreRunes).

Seed corpus (per RESEARCH.md autojunk-sensitive + asymmetric + multi-byte requirements):

- Standard edges: identity, both-empty, one-empty, no-overlap.
- Dr. Dobb's 1988 pairs: `WIKIMEDIA`/`WIKIMANIA`, `GESTALT`/`GESTALT_PATTERN_MATCHING`.
- Autojunk-sensitive 200+ char case.
- Substring containment.
- Multi-byte UTF-8: `café`/`cafe`, `Привет`/`привет`.
- Invalid UTF-8: `\xff\xfe`/`abc`.

---

### `tests/bdd/features/{strcmp95,lcsstr,ratcliff_obershelp}.feature` (Plans 04-01, 04-02, 04-03)

**Analog:** `/Users/johnny/Development/fuzzymatch/tests/bdd/features/swg.feature` (46 lines — exact template).

**Shape pattern** (copy `swg.feature` verbatim; substitute algorithm name + canonical pairs):

```gherkin
# Primary source: <Author> <Year>. "<Title>." <Journal>:<pages>.

Feature: <Algorithm> similarity
  <One-paragraph description.>

  Scenario Outline: canonical reference vectors
    When I compute the <Algorithm> score between "<a>" and "<b>"
    Then the score should be approximately <score> within 0.0001
    Examples:
      | a       | b        | score  |
      | ...     | ...      | ...    |

  Scenario: identical strings score 1.0
    When I compute the <Algorithm> score between "user_id" and "user_id"
    Then the score should be exactly 1

  Scenario: both-empty strings score 1.0
    When I compute the <Algorithm> score between "" and ""
    Then the score should be exactly 1

  Scenario: one-empty string scores 0.0
    When I compute the <Algorithm> score between "abc" and ""
    Then the score should be exactly 0

  Scenario: score is symmetric        # SKIP for Ratcliff-Obershelp per OQ-1
    When I compute the <Algorithm> score between "kitten" and "sitting"
    And I compute the second <Algorithm> score between "sitting" and "kitten"
    Then both <Algorithm> scores should be equal
```

**Deviations:**

- `strcmp95.feature` — ADD a scenario where the similar-character table fires (Strcmp95 > JaroWinkler).
- `lcsstr.feature` — ADD a leftmost-tie-break scenario; ADD a Unicode scenario via the `*Runes` surface (if BDD steps expose it) OR document that BDD covers only `LCSStrScore` and rune-path coverage is in unit tests.
- `ratcliff_obershelp.feature` — ADD a 200+-char autojunk-sensitive scenario; OMIT the symmetric scenario (OQ-1 resolution); ADD an asymmetric-pin scenario.
- Score-regex pattern: `(\d+\.?\d*)` accepts integer-form (IN-03 closure). Use `0` and `1` (no `.0` suffix) freely in the feature files.

---

### `testdata/golden/_staging/{strcmp95,lcsstr,ratcliff_obershelp}.json` (Plans 04-01, 04-02, 04-03)

**Analog:** `/Users/johnny/Development/fuzzymatch/testdata/golden/_staging/swg.json` (committed file; same JSON schema).

Schema (copy verbatim — `version: 1`, `entries: [{name, algorithm, a, b, expected_score}]`, sorted alphabetically by `Name`, marshalled via `CanonicalMarshalForTest`):

```json
{
  "version": 1,
  "entries": [
    { "name": "Strcmp95_both_empty", "algorithm": "Strcmp95", "a": "", "b": "", "expected_score": 1 },
    /* ... more entries ... */
  ]
}
```

**Gotcha:** Never edit these by hand. They are generated by `go test -run TestGolden_<Algo>_Staging -update ./...` (the staging test owns the file; per `algorithms_golden_test.go` `assertGoldenStaging` helper at lines 80–106).

---

### `testdata/fuzz/Fuzz<Algo>Score/seed-001` (Plans 04-01, 04-02, 04-03)

**Analog:** `/Users/johnny/Development/fuzzymatch/testdata/fuzz/FuzzSmithWatermanGotohScore/seed-001` (3 lines — exact format).

```
go test fuzz v1
string("a")
string("b")
```

**Gotcha:** First line is exactly `go test fuzz v1`. Each subsequent line is one argument to `f.Fuzz(func(t *testing.T, ...))`. For a 2-arg fuzz, exactly 2 `string(...)` lines (LCSStr, Strcmp95, RatcliffObershelp all take `(a, b string)` so the format is identical). DO NOT drift formatting (Phase 3 IN-06 closure; `gofmt` doesn't apply to these files but the format is byte-stable as written by `go test -fuzz` itself).

---

### `scripts/gen-ratcliff-obershelp-cross-validation.py` (Plan 04-04)

**Analog:** `/Users/johnny/Development/fuzzymatch/scripts/gen-swg-cross-validation.py` (221 lines — 1-to-1 structural copy).

**Header + version guard pattern** (copy lines 1–61; substitute biopython references with stdlib difflib):

```python
#!/usr/bin/env python3
# Copyright 2026 AxonOps Limited
# [...Apache-2.0 header...]

"""scripts/gen-ratcliff-obershelp-cross-validation.py — Ratcliff-Obershelp
cross-validation corpus generator.

Regenerates testdata/cross-validation/ratcliff-obershelp/vectors.json by
running Python's stdlib difflib.SequenceMatcher(autojunk=False).ratio() on
a fixed, deterministic list of input pairs (CASES).

difflib is Python stdlib — PSF licence; used for reference-vector cross-
validation only, NOT for code copying per the project's
.claude/skills/algorithm-licensing-standards.

CRITICAL: autojunk=False is REQUIRED. The default autojunk=True is a
performance heuristic (marks "popular" characters as junk when len(b) >= 200)
that distorts scores. The TRUE Ratcliff-Obershelp algorithm has
autojunk=False.

Usage:
    make regen-ratcliff-obershelp-cross-validation
    # or directly:
    python3 scripts/gen-ratcliff-obershelp-cross-validation.py

Requirements:
    - Python 3.7+ (for dict insertion-order preservation in json.dump).
    - difflib (stdlib — NO pip install needed; this is the structural
      simplification over Phase 3's biopython).
"""

import json
import os
import sys
import difflib

_MIN_PYTHON_VERSION = (3, 7)
```

**CASES + score_case + version_check pattern** (copy `gen-swg-cross-validation.py` lines 85–189 shape; per RESEARCH.md lines 833–894):

- `CASES` is a Python list of `(name, a, b)` tuples (Strcmp95 + LCSStr + RO have no per-case params, unlike SWG).
- `score_case(a, b)` handles both-empty and one-empty before calling `difflib.SequenceMatcher(autojunk=False, a=a, b=b).ratio()` (mirroring `swg.py`'s short-circuit pattern at lines 134–142).
- `_check_python_version()` — copy `_check_biopython_version()` shape; assert `sys.version_info >= _MIN_PYTHON_VERSION`. Phase 3 IN-07 closure.
- `main()` writes JSON to `testdata/cross-validation/ratcliff-obershelp/vectors.json`, `indent=2`, `sort_keys=False`, trailing LF — exact same `json.dump` arguments as `gen-swg-cross-validation.py` line 216.

**Output JSON schema:**

```json
{
  "version": 1,
  "python_version": "3.X.Y",
  "entries": [
    { "name": "identity_short", "a": "hello", "b": "hello", "difflib_ratio": 1.0 },
    /* ... 14-17 more entries ... */
  ]
}
```

**Deviations from `gen-swg-cross-validation.py`:**

- DROP all biopython imports / version-check helper / DEFAULT_PARAMS / per-case `overrides` / `score_case` `aligner` setup.
- ADD difflib stdlib import; ADD Python-version-check helper (`sys.version_info >= (3, 7)`); ADD `autojunk=False` as the FIRST line of the active-case `score_case` body (RESEARCH.md Pitfall 2 is load-bearing).
- Output field rename: `biopython_score` / `biopython_normalised` → `difflib_ratio` (one ratio, no normalisation pre-step needed — difflib already returns `[0, 1]`).
- Header field rename: `biopython_version` → `python_version`.

---

### `testdata/cross-validation/ratcliff-obershelp/vectors.json` (Plan 04-04)

**Analog:** `/Users/johnny/Development/fuzzymatch/testdata/cross-validation/swg/vectors.json` (committed corpus file).

Schema:

```json
{
  "version": 1,
  "python_version": "3.X.Y",
  "entries": [
    {"name": "identity_short",     "a": "hello",   "b": "hello",   "difflib_ratio": 1.0},
    {"name": "both_empty",         "a": "",        "b": "",        "difflib_ratio": 1.0},
    {"name": "one_empty_a",        "a": "",        "b": "abcdef",  "difflib_ratio": 0.0},
    {"name": "wikimedia_wikimania","a": "WIKIMEDIA","b": "WIKIMANIA","difflib_ratio": ...},
    {"name": "gestalt_paper",      "a": "GESTALT", "b": "GESTALT_PATTERN_MATCHING", "difflib_ratio": ...},
    {"name": "autojunk_sensitive", "a": "...", "b": "...", "difflib_ratio": ...},
    /* ... at least 15 total, covering all 4 CONTEXT.md §1 categories ... */
  ]
}
```

**Categories per CONTEXT.md §1 (load-bearing — autojunk-sensitive case is the corpus's keystone proof):**

1. Standard edge cases (4 entries): identity, both-empty, one-empty_a, one-empty_b, no-overlap.
2. Dr. Dobb's 1988 paper examples (2+ entries): WIKIMEDIA/WIKIMANIA, GESTALT/GESTALT_PATTERN_MATCHING.
3. autojunk-sensitive 200+ char case (1 entry): exactly proves `autojunk=False` is correctly disabled.
4. Substring + partial-match + unicode (4–6 entries): substring_middle, partial_overlap, unicode_ascii_only (`café`/`cafe`), longer_identity, etc.

Total: 15–18 entries.

---

### `Makefile` (append target in Plan 04-04)

**Analog:** `/Users/johnny/Development/fuzzymatch/Makefile` lines 196–211 (`regen-swg-cross-validation` target).

```makefile
# Regenerate testdata/cross-validation/ratcliff-obershelp/vectors.json by
# running scripts/gen-ratcliff-obershelp-cross-validation.py. Committed JSON
# is the verification fixture; CI does NOT require Python at test time. No
# pip install needed — difflib is stdlib.
regen-ratcliff-obershelp-cross-validation:
	@if ! command -v python3 >/dev/null 2>&1; then \
	  echo "python3 not found; install Python 3.7+"; \
	  exit 1; \
	fi
	python3 scripts/gen-ratcliff-obershelp-cross-validation.py
```

**Phony declaration:** also append `regen-ratcliff-obershelp-cross-validation` to the `.PHONY` line at top of Makefile (Makefile line 28 — Phase 3 already added `regen-swg-cross-validation` to that list; append the new target alongside).

---

### `CONTRIBUTING.md` (append doc line in Plan 04-04)

**Analog:** `/Users/johnny/Development/fuzzymatch/CONTRIBUTING.md` line 92 (the entry for `make regen-swg-cross-validation`).

```markdown
- `make regen-ratcliff-obershelp-cross-validation` — developer-only; regenerate
  the difflib cross-validation corpus (`testdata/cross-validation/ratcliff-obershelp/vectors.json`).
  Requires Python 3.7+ (difflib is stdlib — no pip install needed).
```

**Gotcha:** `makefile_targets_test.go::TestMakefile_TargetsDocumentedInContributing` will fail if the target is added to Makefile without a corresponding CONTRIBUTING line. Both changes land together in plan 04-04.

---

### `props_test.go` (extend in plans 04-01, 04-02, 04-03)

**Analog:** `/Users/johnny/Development/fuzzymatch/props_test.go` lines 737–909 (the SWG property-test block — 6 standard invariants + 3 SWG-specific).

**Extension pattern per algorithm** (append a sectioned block per algorithm; copy the exact `TestProp_SmithWatermanGotohScore_*` shape):

For each of Strcmp95 / LCSStr / RatcliffObershelp, append:

```go
// ---------------------------------------------------------------------------
// <Algorithm> property tests (plan 04-XX)
// ---------------------------------------------------------------------------

func TestProp_<Algo>Score_RangeBounds(t *testing.T) { /* copy lines 737–747 shape */ }
func TestProp_<Algo>Score_Identity(t *testing.T)    { /* copy lines 749–761 shape */ }
func TestProp_<Algo>Score_Symmetric(t *testing.T)   { /* copy lines 763–772 shape — OMIT for RatcliffObershelp per OQ-1 */ }
func TestProp_<Algo>Score_NoNaN(t *testing.T)       { /* copy lines 774–785 shape */ }
func TestProp_<Algo>Score_NoInf(t *testing.T)       { /* copy lines 787–797 shape */ }
func TestProp_<Algo>Score_NoNegativeZero(t *testing.T) { /* copy lines 799–810 shape */ }
```

**Algorithm-specific extensions per CONTEXT.md §5:**

- **Strcmp95:** add `TestProp_Strcmp95Score_AtLeastJaroWinkler` (RESEARCH.md Pitfall 1 warning sign #3; loop over hand-curated or quick.Check-generated inputs and assert `Strcmp95Score(a, b) >= JaroWinklerScore(a, b)`). Add `TestProp_Strcmp95Score_DeterministicAcrossRuns` (PITFALLS §14 closure: same input → byte-identical output across 1000 sequential calls).
- **LCSStr:** add `TestProp_LongestCommonSubstring_IsSubstringOfBoth` (`strings.Contains(a, got) && strings.Contains(b, got)`); `TestProp_LongestCommonSubstring_LengthMatchesScore` (`2·len(got)/(la+lb) ≈ LCSStrScore(a, b)` within `1e-9`); `TestProp_LongestCommonSubstring_LeftmostTieBreak` (hand-curated inputs — pattern from `cross_algorithm_consistency_test.go` lines 65–82's hand-curated divergence-pair shape).
- **RatcliffObershelp:** add `TestProp_RatcliffObershelpScore_AtLeastLevenshtein_HandCurated` (hand-curated substring-containment inputs; not testing/quick over all strings — RESEARCH.md notes the property is "generally" true, not universal).

**Critical deviation per OQ-1 (CONTEXT.md §5 NB):** **DROP `TestProp_RatcliffObershelpScore_Symmetric`** — RO is asymmetric by design (matches difflib). The other 5 standard property tests (RangeBounds, Identity, NoNaN, NoInf, NoNegativeZero) all still apply.

---

### `cross_algorithm_consistency_test.go` (extend in plan 04-05)

**Analog:** `/Users/johnny/Development/fuzzymatch/cross_algorithm_consistency_test.go` lines 205–226 (`TestCrossAlgorithm_SWG_Levenshtein_SubstringDivergence` — the strongest divergence-pin template).

**Three new tests + one RO asymmetry pin** (append to bottom of file; pattern from lines 205–226):

```go
// TestCrossAlgorithm_Strcmp95_AtLeastJaroWinkler asserts the algorithm-
// hierarchy invariant per CONTEXT.md §5: Strcmp95 only ADDS adjustments
// atop Jaro-Winkler, so it must never score lower.
func TestCrossAlgorithm_Strcmp95_AtLeastJaroWinkler(t *testing.T) {
    pairs := [][2]string{
        {"MARTHA", "MARHTA"},        // adjustment may not fire
        {"DWAYNE", "DUANE"},         // adjustment likely fires
        {"DIXON",  "DICKSONX"},      // adjustment fires
    }
    for _, p := range pairs {
        a, b := p[0], p[1]
        gotStrcmp := fuzzymatch.Strcmp95Score(a, b)
        gotJW     := fuzzymatch.JaroWinklerScore(a, b)
        if !(gotStrcmp >= gotJW) {
            t.Errorf("Strcmp95(%q, %q)=%v < JaroWinkler(%q, %q)=%v (adjustments must only ADD)", a, b, gotStrcmp, a, b, gotJW)
        }
    }
}

// TestCrossAlgorithm_LCSStr_AtLeastLevenshtein_SubstringContainment asserts
// the substring-containment divergence: LCSStr finds the full substring;
// Levenshtein pays the deletion cost.
func TestCrossAlgorithm_LCSStr_AtLeastLevenshtein_SubstringContainment(t *testing.T) {
    a, b := "http_request", "http_request_header_fields"
    gotLCS := fuzzymatch.LCSStrScore(a, b)
    gotLev := fuzzymatch.LevenshteinScore(a, b)
    if !(gotLCS >= gotLev) {
        t.Errorf("LCSStr(%q,%q)=%v < Levenshtein(%q,%q)=%v on substring-containment", a, b, gotLCS, a, b, gotLev)
    }
}

// TestCrossAlgorithm_RatcliffObershelp_PinnedDrDobbs asserts at least one
// hand-pinned Dr. Dobb's 1988 vector — RO must converge with difflib on
// the canonical paper examples.
func TestCrossAlgorithm_RatcliffObershelp_PinnedDrDobbs(t *testing.T) {
    const tol = 1e-9
    // WIKIMEDIA vs WIKIMANIA: difflib returns ~0.7778 (hand-verified).
    got := fuzzymatch.RatcliffObershelpScore("WIKIMEDIA", "WIKIMANIA")
    if math.Abs(got - 0.7777777777777778) > tol {
        t.Errorf("RatcliffObershelpScore(WIKIMEDIA, WIKIMANIA)=%g; want ≈0.7778", got)
    }
}

// TestCrossAlgorithm_RatcliffObershelp_AsymmetryPin pins the load-bearing
// OQ-1 resolution: RO is intentionally asymmetric (mirrors difflib's
// noncommutative ratio()). For documented asymmetric pairs, Score(a,b) !=
// Score(b,a). Documented in godoc; this test pins the contract.
func TestCrossAlgorithm_RatcliffObershelp_AsymmetryPin(t *testing.T) {
    fwd := fuzzymatch.RatcliffObershelpScore("tide", "diet")
    rev := fuzzymatch.RatcliffObershelpScore("diet", "tide")
    if fwd == rev {
        t.Errorf("RatcliffObershelpScore is INTENTIONALLY asymmetric for tide/diet — got fwd=%g==rev=%g (regression to symmetric behaviour)", fwd, rev)
    }
}
```

**Deviations from existing tests:** the asymmetry pin is the inverse-form of the existing `TestCrossAlgorithm_SWG_Levenshtein_SubstringDivergence` — it asserts INEQUALITY (Score(a,b) != Score(b,a)) rather than ordering.

---

### `examples/identifier-similarity/main.go` (extend in plan 04-05)

**Analog:** `/Users/johnny/Development/fuzzymatch/examples/identifier-similarity/main.go` lines 73–84 (the `algorithms` slice).

**Append three entries to the `algorithms` slice** (between line 83 and `}`):

```go
var algorithms = []struct {
    name string
    fn   func(a, b string) float64
}{
    {"Levenshtein",   fuzzymatch.LevenshteinScore},
    {"DL-OSA",        fuzzymatch.DamerauLevenshteinOSAScore},
    {"DL-Full",       fuzzymatch.DamerauLevenshteinFullScore},
    {"Hamming",       fuzzymatch.HammingScore},
    {"Jaro",          fuzzymatch.JaroScore},
    {"Jaro-Winkler",  fuzzymatch.JaroWinklerScore},
    {"SWG",           fuzzymatch.SmithWatermanGotohScore},
    {"Strcmp95",      fuzzymatch.Strcmp95Score},          // NEW
    {"LCSStr",        fuzzymatch.LCSStrScore},            // NEW
    {"RatcliffOber",  fuzzymatch.RatcliffObershelpScore}, // NEW (truncated to fit column)
}
```

**Gotcha:** `algoWidth = 13` (from line 89). The longest existing column name is "Jaro-Winkler" (12 chars). "RatcliffObershelp" is 17 chars — TOO LONG. Either:
- (a) Increase `algoWidth` to 18+ (causes a wider table — touches `TestExample_ColumnWidths`).
- (b) Truncate the display name to "RatcliffOber" or "RO" (the planner picks — column header is a label, not the function name).

Recommendation: option (b) — short label "RO" for compactness, matching SWG's already-truncated label. Document the abbreviation in a code comment near the `algorithms` slice. The `want` constant in `main_test.go` is regenerated via `go run .` and committed; no manual edit.

---

### `examples/identifier-similarity/main_test.go` (extend in plan 04-05)

**Analog:** `/Users/johnny/Development/fuzzymatch/examples/identifier-similarity/main_test.go` lines 41–50 (the `want` constant).

**Process** (no manual edit):

1. Update `main.go`'s `algorithms` slice (above).
2. `cd examples/identifier-similarity && go run .` — capture stdout.
3. Paste the captured output into the `want` constant in `main_test.go` (replacing the existing 9-line table block — header + separator + 7 data rows; the new block is header + separator + 7 data rows with 3 more columns each).
4. Re-run `go test ./...` to confirm byte-stability.

**Gotcha:** Phase 3 WR-04 closure already locks the defer-restore pattern for `os.Stdout` (lines 64–70). No changes to the test logic — only the `want` constant changes. The line-by-line diff helper (lines 89–108) is unchanged.

---

### `tests/bdd/steps/algorithms_steps.go` (extend in plans 04-01, 04-02, 04-03)

**Analog:** `/Users/johnny/Development/fuzzymatch/tests/bdd/steps/algorithms_steps.go` lines 290–316 (SWG step methods) + lines 443–455 (SWG step regex registrations).

**Per-algorithm step-method block** (append to the bottom of the file before `InitializeScenario`):

```go
// ---------------------------------------------------------------------------
// Strcmp95 step definitions (plan 04-01)
// ---------------------------------------------------------------------------

func (ctx *AlgorithmContext) iComputeTheStrcmp95ScoreBetween(a, b string) error {
    ctx.lastScore = fuzzymatch.Strcmp95Score(a, b)
    return nil
}

func (ctx *AlgorithmContext) iComputeTheSecondStrcmp95ScoreBetween(a, b string) error {
    ctx.lastScore2 = fuzzymatch.Strcmp95Score(a, b)
    return nil
}

func (ctx *AlgorithmContext) bothStrcmp95ScoresShouldBeEqual() error {
    if ctx.lastScore != ctx.lastScore2 {
        return fmt.Errorf("strcmp95 scores not equal: %f != %f", ctx.lastScore, ctx.lastScore2)
    }
    return nil
}
```

**Step regex registration** (append three blocks to `InitializeScenario` at the bottom, after the SWG block at lines 443–455):

```go
// Strcmp95 step definitions (plan 04-01).
ctx.Step(`^I compute the Strcmp95 score between "([^"]*)" and "([^"]*)"$`, a.iComputeTheStrcmp95ScoreBetween)
ctx.Step(`^I compute the second Strcmp95 score between "([^"]*)" and "([^"]*)"$`, a.iComputeTheSecondStrcmp95ScoreBetween)
ctx.Step(`^both Strcmp95 scores should be equal$`, a.bothStrcmp95ScoresShouldBeEqual)

// LCSStr step definitions (plan 04-02). [same shape]
// RatcliffObershelp step definitions (plan 04-03). [same shape — but NO "second / equal" steps if the feature file omits the symmetry scenario per OQ-1]
```

**Gotcha:** testify IS permitted in this file (sub-module, not root). The existing SWG block uses `fmt.Errorf` for assertion; Phase 4 stays consistent.

**Score regex `(\d+\.?\d*)`:** ALREADY registered at lines 341 + 345 (IN-03 closure — accepts integer-form). Phase 4 feature files can use `0` and `1` (no `.0` suffix); no new score-regex registrations needed.

---

### `algoid_test.go` (extend in plans 04-01, 04-02, 04-03)

**Analog:** `/Users/johnny/Development/fuzzymatch/algoid_test.go` lines 284–292 (`TestDispatch_SmithWatermanGotohRegistered`) + lines 294–323 (`TestDispatch_UnregisteredSlotsAreNil`).

**Per-algorithm dispatch test** (copy `TestDispatch_SmithWatermanGotohRegistered` shape; substitute names):

```go
// TestDispatch_Strcmp95Registered asserts that dispatch[AlgoStrcmp95] (slot 5)
// is non-nil after plan 04-01 registers Strcmp95Score.
func TestDispatch_Strcmp95Registered(t *testing.T) {
    if fuzzymatch.DispatchEntryNilForTest(int(fuzzymatch.AlgoStrcmp95)) {
        t.Errorf("dispatch[AlgoStrcmp95] (%d) is nil — dispatch_strcmp95.go must register Strcmp95Score at package load time",
            int(fuzzymatch.AlgoStrcmp95))
    }
}
```

Add three: `TestDispatch_Strcmp95Registered`, `TestDispatch_LCSStrRegistered`, `TestDispatch_RatcliffObershelpRegistered`.

**Extend `TestDispatch_UnregisteredSlotsAreNil`** (lines 299–323): extend the `registered` map with the three new entries. Be careful — slot numbers per `algoid.go`:
- `AlgoStrcmp95` = slot 5
- `AlgoLCSStr` = slot 8
- `AlgoRatcliffObershelp` = slot 22 (the LAST slot, `numAlgorithms - 1`)

Update the godoc comment on `TestDispatch_UnregisteredSlotsAreNil` to list "plan 04-01, 04-02, 04-03" as new registrants.

---

### `example_test.go` (extend in plans 04-01, 04-02, 04-03)

**Analog:** `/Users/johnny/Development/fuzzymatch/example_test.go` lines 103–122 (the two SWG `ExampleXxx` functions).

**Per-algorithm `ExampleXxx` block** (append after the existing SWG block):

```go
// ExampleStrcmp95Score demonstrates Winkler's Strcmp95 enhancement on the
// canonical Winkler 1990 reference pair.
func ExampleStrcmp95Score() {
    fmt.Printf("%.4f\n", fuzzymatch.Strcmp95Score("MARTHA", "MARHTA"))
    // Output:
    // 0.XXXX
}
```

Functions to add (7 total per CONTEXT.md §6 plan 04-05):

1. `ExampleStrcmp95Score`
2. `ExampleLongestCommonSubstring`
3. `ExampleLongestCommonSubstringRunes`
4. `ExampleLCSStrScore`
5. `ExampleLCSStrScoreRunes`
6. `ExampleRatcliffObershelpScore`
7. `ExampleRatcliffObershelpScoreRunes`

**Gotcha:** `// Output:` byte-stability — same as Phase 2/3. The output literal must match `fmt.Printf` to the byte. For `LongestCommonSubstring` examples, use `fmt.Println` (string output) rather than `%.4f`.

---

### `algorithms_golden_test.go` (extend in plan 04-05)

**Analog:** `/Users/johnny/Development/fuzzymatch/algorithms_golden_test.go` lines 584–656 (`buildSWGStagingEntries` + `TestGolden_SWG_Staging`) + lines 162–195 (the `TestGolden_Algorithms_Merge` `stagingFiles` slice).

**Per-algorithm builder + staging test** (copy `buildSWGStagingEntries` lines 593–641 shape; substitute):

```go
func buildStrcmp95StagingEntries(t *testing.T) []goldenAlgorithmEntry {
    t.Helper()
    return []goldenAlgorithmEntry{
        {Name: "Strcmp95_both_empty",      Algorithm: "Strcmp95", A: "",        B: "",        ExpectedScore: fuzzymatch.Strcmp95Score("", "")},
        {Name: "Strcmp95_identical",       Algorithm: "Strcmp95", A: "abc",     B: "abc",     ExpectedScore: fuzzymatch.Strcmp95Score("abc", "abc")},
        {Name: "Strcmp95_MARTHA_MARHTA",   Algorithm: "Strcmp95", A: "MARTHA",  B: "MARHTA",  ExpectedScore: fuzzymatch.Strcmp95Score("MARTHA", "MARHTA")},
        {Name: "Strcmp95_DWAYNE_DUANE",    Algorithm: "Strcmp95", A: "DWAYNE",  B: "DUANE",   ExpectedScore: fuzzymatch.Strcmp95Score("DWAYNE", "DUANE")},
        // ... more entries
    }
}

func TestGolden_Strcmp95_Staging(t *testing.T) {
    entries := buildStrcmp95StagingEntries(t)
    sort.Slice(entries, func(i, j int) bool { return entries[i].Name < entries[j].Name })
    file := goldenAlgorithmsFile{Version: 1, Entries: entries}
    assertGoldenStaging(t, "_staging/strcmp95.json", file)
}
```

Add three builders + three staging tests: `buildStrcmp95StagingEntries`, `buildLCSStrStagingEntries`, `buildRatcliffObershelpStagingEntries`.

**Extend `TestGolden_Algorithms_Merge`** (lines 163–171): append three new staging paths to the slice:

```go
stagingFiles := []string{
    "_staging/damerau_full.json",
    "_staging/damerau_osa.json",
    "_staging/hamming.json",
    "_staging/jaro.json",
    "_staging/jarowinkler.json",
    "_staging/levenshtein.json",
    "_staging/swg.json",
    "_staging/strcmp95.json",            // NEW
    "_staging/lcsstr.json",              // NEW
    "_staging/ratcliff_obershelp.json",  // NEW
}
```

---

### `llms.txt` + `llms-full.txt` (extend in plan 04-05)

**Analog:** `/Users/johnny/Development/fuzzymatch/llms.txt` lines 87–94 (the SWG block — `SWGParams`, `NewSWGParams`, 6 function entries).

**Append 7 new entries to `llms.txt`** (one line per exported symbol; format: `- type|func|const <Symbol>`):

```
- func Strcmp95Score(a, b string) float64
- func LongestCommonSubstring(a, b string) string
- func LongestCommonSubstringRunes(a, b string) string
- func LCSStrScore(a, b string) float64
- func LCSStrScoreRunes(a, b string) float64
- func RatcliffObershelpScore(a, b string) float64
- func RatcliffObershelpScoreRunes(a, b string) float64
```

**Append parallel entries to `llms-full.txt`** with one-line rationales (same shape as SWG block in `llms-full.txt`).

**Gotcha:** `ai_friendly_test.go::TestLLMs_PublicSymbolsListed` parses the package's AST and asserts every exported symbol is present in `llms.txt`. The meta-test fails if any of the 7 new symbols is missing. The 3 AlgoID constants (`AlgoStrcmp95`, `AlgoLCSStr`, `AlgoRatcliffObershelp`) are ALREADY listed in `llms.txt` (declared in Phase 1, listed at lines 30, 31, ... — they're in the catalogue block). No new AlgoID entries needed.

---

### `bench.txt` (full-replace in plan 04-05)

**Analog:** none — the file IS the artefact. Process:

1. Implement plans 04-01 through 04-04 first.
2. Run `make bench` from a clean state on the reference benchmark hardware.
3. Replace `bench.txt` with the resulting output.
4. Commit alongside the algorithm changes.

**Gotcha:** `make bench-compare` will fail on the new benches until `bench.txt` is replaced. The finalisation plan 04-05 must include the full replace as a single commit (per Phase 2 02-07 finalisation pattern).

---

## Shared Patterns

### License + Source Header (applies to ALL new `.go` files)

**Source:** `/Users/johnny/Development/fuzzymatch/swg.go` lines 1–13 (Apache-2.0 block); lines 15–82 (file-level doc with `Sources:`, recurrence/formula in godoc, `Implementation discipline` bullets); plus source-origin statement per `algorithm-licensing-standards`.

**Apply to:** every new `.go` file. `scripts/verify-license-headers.sh` is the CI gate.

```go
// Copyright 2026 AxonOps Limited
// [...Apache-2.0 header — 13 lines verbatim...]

// <file_name>.go implements <algorithm> for the fuzzymatch catalogue.
//
// Sources:
//   - <Author> <Year>. "<Title>." <Venue>:<pages>.
//   - <secondary citation if cross-validation source differs from primary>
//
// <formula / recurrence in godoc block>
//
// Source-origin statement:
//   Primary: <paper>.
//   Cross-validation: <reference impl + licence>.
//   GPL/LGPL: none consulted.
//   Code copied: none.
//
// Implementation discipline (inherits Phase 2):
//   - ASCII fast path ...
//   - NO init()-time table builds ...
//   - NO map iteration on output paths ...
//   - NO transcendental float operations ...

package fuzzymatch
```

---

### isASCII + maxStackInputLen reuse (applies to `lcsstr.go`; NOT to `strcmp95.go` or `ratcliff_obershelp.go`)

**Source:** `/Users/johnny/Development/fuzzymatch/normalise.go` lines 159–168 (`isASCII`); `/Users/johnny/Development/fuzzymatch/levenshtein.go` (declares `maxStackInputLen = 64`).

**Apply to:** `lcsstr.go` ONLY. Strcmp95 has no DP buffer; Ratcliff-Obershelp has no DP buffer (recursive LCS-substring search has no rolling rows).

```go
// In lcsstr.go (do NOT redeclare isASCII or maxStackInputLen):
if n <= maxStackInputLen && isASCII(a) && isASCII(b) {
    var buf [(maxStackInputLen + 1) * 2]int
    return lcsstrDP(a, b, m, n, buf[:n+1], buf[n+1:n+n+2])
}
return lcsstrDP(a, b, m, n, make([]int, n+1), make([]int, n+1))
```

---

### Identity short-circuit on `*Runes` (applies to `LCSStrScoreRunes`, `LongestCommonSubstringRunes`, `RatcliffObershelpScoreRunes`)

**Source:** `/Users/johnny/Development/fuzzymatch/jaro.go` lines 172–179 (`JaroScoreRunes`); IN-04 cleanup pattern.

```go
func <Algo>ScoreRunes(a, b string) float64 {
    if a == b { return 1.0 } // covers both-empty + identical; saves []rune allocs
    ra := []rune(a)
    rb := []rune(b)
    // ...
}
```

**Apply to:** every `*Runes` function in `lcsstr.go` and `ratcliff_obershelp.go`. Strcmp95 has no `*Runes` variant (CONTEXT.md §2).

---

### No init() / var-only tables (applies to `strcmp95.go`)

**Source:** PITFALLS.md §14; `algoid.go` lines 15–28 ("No init() function appears in this file"); RESEARCH.md Pattern 1 + Code Examples for Strcmp95 table.

**Apply to:** `strcmp95.go`'s `strcmp95SimilarChars` table. Determinism-reviewer flags any `init()` in this file as BLOCKING.

---

### Left-to-right float reduction (applies to all three score normalisations)

**Source:** `/Users/johnny/Development/fuzzymatch/jaro.go` lines 44–47 (Jaro formula); DET-06 in `determinism-standards`; RESEARCH.md Pitfall 8.

```go
// Always parenthesise explicitly:
numer := 2.0 * float64(m)        // multiply first
denom := float64(la + lb)
return numer / denom              // then divide
```

**Apply to:** `LCSStrScore`, `LCSStrScoreRunes`, `RatcliffObershelpScore`, `RatcliffObershelpScoreRunes`, and Strcmp95's adjustment-application loop. Cross-platform CI matrix verifies byte-identical golden output.

---

### `var sink` + `if sink < 0 { b.Fatal(...) }` (applies to all bench files)

**Source:** `/Users/johnny/Development/fuzzymatch/swg_bench_test.go` lines 60–66 (canonical pattern).

```go
b.ReportAllocs()
b.ResetTimer()
var sink float64  // or 'var sink string' for LongestCommonSubstring benches
for i := 0; i < b.N; i++ {
    sink = fuzzymatch.<Algo>Score(...)
}
if sink < 0 { b.Fatal("sink unexpectedly negative — compiler folded the benchmark away") }
```

For substring-returning benches, swap `var sink string` and `if len(sink) < 0` (impossible — but it forces the compiler to keep `sink` alive).

---

### testing.AllocsPerRun(100, ...) alloc gate at runtime (applies to LCSStr tests only)

**Source:** `/Users/johnny/Development/fuzzymatch/swg_test.go` lines 384–409 (`TestSmithWatermanGotohScore_ZeroAllocs_ASCII_Short` + `_Medium`).

**Apply to:** `lcsstr_test.go` — `TestLCSStrScore_ZeroAllocs_ASCII_Short` + `_Medium`. NOT applicable to `strcmp95_test.go` (no DP buffer; alloc behaviour depends on Jaro match-flag arrays) or `ratcliff_obershelp_test.go` (recursion has no alloc-gate target).

---

### Stdlib testing only in root (applies to all root tests)

**Source:** `/Users/johnny/Development/fuzzymatch/swg_test.go` line 23 ("Stdlib `testing` only — no testify in root tests, per .claude/skills/go-coding-standards.").

**Apply to:** every new `_test.go` file in root. Use `t.Errorf`, `t.Fatalf`, table-driven sub-tests via `t.Run`. testify is permitted ONLY in `tests/bdd/steps/algorithms_steps.go`.

---

## No Analog Found

All Phase 4 files have strong analogs from Phase 2 (the 6 character algorithms) or Phase 3 (SWG). The closest things to "no analog" are listed below — but each has a clean derivation path documented in RESEARCH.md.

| File / Pattern | Reason for partial coverage | Resolution |
|----------------|---------------------------|------------|
| `strcmp95SimilarChars` table declaration | No analog in Phase 1–3 (Strcmp95 is the first algorithm with a constant table) | Pattern documented in RESEARCH.md Code Examples (lines 696–735) verbatim; PITFALLS §14 governs the `var`-only discipline. |
| `LongestCommonSubstring` returning `string` | No analog in Phase 1–3 (every existing algorithm returns `int` or `float64`) | Function shape mirrors `LevenshteinDistance` (also returns `int`), with `string` substituted; bench file's `var sink string` replaces `var sink float64`. |
| Recursive longest-common-substring decomposition (`roMatchedLength`) | No analog (SWG uses iterative DP; LCSStr uses iterative DP) | RESEARCH.md Pattern 3 code at lines 785–801 is the canonical reference. OQ-2 locks the recursion pattern (call stack, not explicit stack). |
| Asymmetric algorithm pin (RO `TestCrossAlgorithm_RatcliffObershelp_AsymmetryPin`) | All existing algorithms are symmetric; no analog for an "asymmetry is intentional" test | OQ-1 resolution Option 1; pattern is `if fwd == rev { t.Errorf(...) }` — inverse of existing symmetry pins. |

---

## Metadata

**Analog search scope:**
- `/Users/johnny/Development/fuzzymatch/*.go` (root package — Phase 1+2+3 production + test files)
- `/Users/johnny/Development/fuzzymatch/scripts/*.py` (Python cross-validation generator)
- `/Users/johnny/Development/fuzzymatch/tests/bdd/features/*.feature`
- `/Users/johnny/Development/fuzzymatch/tests/bdd/steps/algorithms_steps.go`
- `/Users/johnny/Development/fuzzymatch/testdata/golden/_staging/` + `testdata/fuzz/`
- `/Users/johnny/Development/fuzzymatch/Makefile` + `CONTRIBUTING.md`
- `/Users/johnny/Development/fuzzymatch/examples/identifier-similarity/`

**Files scanned:** ~30 (all Phase 1+2+3 production + test + tooling files relevant to Phase 4's surface)

**Pattern extraction date:** 2026-05-14
