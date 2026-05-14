# Phase 3: Smith-Waterman-Gotoh — Pattern Map

**Mapped:** 2026-05-14
**Files analysed:** 22 (10 new + 8 extend-only + 4 unique-to-Phase-3)
**Analogs found:** 18 / 22 (the four unique-to-Phase-3 files have no Phase 2 analog and are specified from scratch)

## File Classification

| New/Modified File | Role | Closest Analog | Match Quality |
|---|---|---|---|
| `swg.go` | new-implementation | `levenshtein.go` | exact (DP byte+rune, ASCII fast path, identity short-circuit) — adapt for affine-gap M/Ix/Iy three-matrix two-row form + SWGParams |
| `dispatch_swg.go` | new-implementation | `dispatch_levenshtein.go` | exact (character-for-character copy, only AlgoXxx + XxxScore identifiers change) |
| `swg_test.go` | new-test | `levenshtein_test.go` | exact (table-driven unit tests + literature reference vectors); ALSO hosts `TestSWG_CrossValidation` (unique-to-Phase-3 — see §unique-files-1) |
| `swg_bench_test.go` | new-benchmark | `levenshtein_bench_test.go` | role-match; SWG adds two extra benches (`_WithParams_ASCII_Short`, `_RawScore_ASCII_Short`) for 6 total vs Levenshtein's 4 |
| `swg_fuzz_test.go` | new-fuzz | `levenshtein_fuzz_test.go` | exact (single `FuzzSmithWatermanGotohScore`, panic-free + score-in-[0,1] invariants, programmatic seeds + on-disk corpus) |
| `tests/bdd/features/swg.feature` | new-feature-file | `tests/bdd/features/levenshtein.feature` | exact (per-algorithm Gherkin file; reference-vector Scenario Outline + identity + both-empty + one-empty + symmetry + adds gap-split canary + params scenarios) |
| `testdata/golden/_staging/swg.json` | new-golden-staging | `testdata/golden/_staging/levenshtein.json` | exact (committed via `assertGoldenStaging` with `-update`; merged into `algorithms.json` by `TestGolden_Algorithms_Merge`) |
| `testdata/fuzz/FuzzSmithWatermanGotohScore/seed-001` | new-fuzz-seed | `testdata/fuzz/FuzzLevenshteinScore/seed-001` | exact (`go test fuzz v1` literal corpus format) |
| `testdata/cross-validation/swg/vectors.json` | new-cross-validation-corpus | (none — unique to Phase 3) | see §unique-files-1 |
| `scripts/gen-swg-cross-validation.py` | new-script | (none — first Python in repo) | see §unique-files-2 |
| `props_test.go` | extend-existing | own analog (lines 65–158 Levenshtein block + lines 727–793 rune-symmetry trailer) | exact (append SWG block; 6 standard + 3 SWG-specific property tests; rune-symmetry function appended to trailer) |
| `example_test.go` | extend-existing | own analog (`ExampleLevenshteinScore` lines 34–38) | exact (append `ExampleSmithWatermanGotohScore` + `ExampleSmithWatermanGotohRawScore`) |
| `algoid_test.go` | extend-existing | own analog (`TestDispatch_LevenshteinRegistered` lines 230–235 + `TestDispatch_UnregisteredSlotsAreNil` lines 289–312) | exact (add `TestDispatch_SmithWatermanGotohRegistered`; flip slot 6 in `registered` map; the SWG SUMMARY notes this is a two-edit append) |
| `algorithms_golden_test.go` | extend-existing | own analog (`buildLevenshteinStagingEntries` + `TestGolden_Levenshtein_Staging` lines 196–210; `TestGolden_Algorithms_Merge` lines 162–194) | exact (add `buildSWGStagingEntries` + `TestGolden_SmithWatermanGotoh_Staging`; append `_staging/swg.json` to merge list) |
| `cross_algorithm_consistency_test.go` | extend-existing | own analog (`TestCrossAlgorithm_OSA_Full_Divergence` lines 57–75) | exact (one new divergence test pinning SWG-vs-Levenshtein on substring containment) |
| `bench.txt` | extend-existing | own analog | exact (append 6 SWG benchmark series — 10 lines each — in alphabetical slot near top with the other algorithms) |
| `llms.txt` | extend-existing | own analog (lines 47–53 Levenshtein block) | exact (add `### Smith-Waterman-Gotoh local alignment similarity` block with 6 functions + SWGParams + NewSWGParams) |
| `examples/identifier-similarity/main.go` | extend-existing | own analog (lines 71–81 algorithms slice) | exact (append `{"SWG", fuzzymatch.SmithWatermanGotohScore}` to algorithms slice; widen columns if needed) |
| `examples/identifier-similarity/main_test.go` | extend-existing | own analog (`want` constant lines 40–49) | exact (regenerate `want` with the new SWG column; column-widths test already adapts) |
| `tests/bdd/steps/algorithms_steps.go` | extend-existing | own analog (Levenshtein step block lines 51–93 + `InitializeScenario` Levenshtein registrations lines 302–322) | exact (append SWG step methods; append SWG `ctx.Step(...)` registrations; new `iComputeTheSmithWatermanGotohRawScoreBetween` + `_WithParams` step) |
| `Makefile` | extend-existing | own analog (`verify-determinism` target lines 188–191) | role-match (new target `regen-swg-cross-validation` follows the `verify-*` style — invokes the Python script) |
| `docs/requirements.md` | roadmap-update | own §7.1.8 | role-match (add 3 Raw* functions to the public surface table per CONTEXT.md §4) |
| `TestSWG_CrossValidation` (inside `swg_test.go`) | new-test (special) | (none — first cross-validation gate) | see §unique-files-3 |

---

## Pattern Assignments — New Files

### `swg.go` (new-implementation)

**Analog:** `levenshtein.go` (272 lines)

**Apache-2.0 header + package + primary-source-citation godoc block** (analog lines 1–55, copy verbatim adjusting only the algorithm name and source citations):

```go
// Copyright 2026 AxonOps Limited
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

// swg.go implements the Smith-Waterman-Gotoh local-alignment similarity with
// affine gap penalty for the fuzzymatch catalogue.
//
// Sources:
//   - Smith, T. F. & Waterman, M. S. (1981). "Identification of common
//     molecular subsequences." J. Mol. Biol. 147:195-197 (local-alignment
//     formulation).
//   - Gotoh, O. (1982). "An improved algorithm for matching biological
//     sequences." J. Mol. Biol. 162:705-708 (affine-gap O(mn) reduction).
//   - Flouri, T. et al. (2015). "Are all global alignment algorithms and
//     implementations correct?" biorxiv 031500 — documents the Gotoh 1982
//     initialisation erratum and the corrected formulation transcribed here.
//
// Recurrence (corrected per Flouri et al. 2015; for LOCAL alignment every
// border cell of M, Ix, Iy initialises to 0, NOT -Inf):
// [insert the M / Ix / Iy max-with-0 equations as in RESEARCH.md §Pattern 1]
//
// Implementation discipline (inherits Phase 2):
//   - ASCII fast path operates on bytes directly when n <= maxStackInputLen
//     && isASCII(a) && isASCII(b); a stack-allocated
//     [(maxStackInputLen+1)*6]float64 buffer holds the six rolling rows
//     (prevM, currM, prevIx, currIx, prevIy, currIy).
//   - NO init()-time table builds (per §5(12) and determinism-standards).
//   - NO map iteration on output paths.
//   - NO transcendental float operations: only +, -, *, /, max-style if-
//     comparison, and float64() conversion.
//   - NO goroutines, channels, or mutexes.

package fuzzymatch
```

**Identity short-circuit + ASCII fast-path gate** — copy idiom from `levenshtein.go` lines 84–119:

```go
func SmithWatermanGotohScore(a, b string) float64 {
    return SmithWatermanGotohScoreWithParams(a, b, NewSWGParams())
}

func SmithWatermanGotohScoreWithParams(a, b string, params SWGParams) float64 {
    if a == b {
        return 1.0 // identity short-circuit (covers both-empty + identical inputs)
    }
    la, lb := len(a), len(b)
    if la == 0 || lb == 0 {
        return 0.0
    }
    // Ensure b is the shorter dimension so the inner-loop dimension is minimal.
    if la < lb {
        a, b = b, a
        la, lb = lb, la
    }
    if lb <= maxStackInputLen && isASCII(a) && isASCII(b) {
        var buf [(maxStackInputLen + 1) * 6]float64
        n1 := lb + 1
        raw := swgDPRaw(a, b, la, lb, params,
            buf[0*n1:1*n1], buf[1*n1:2*n1],
            buf[2*n1:3*n1], buf[3*n1:4*n1],
            buf[4*n1:5*n1], buf[5*n1:6*n1])
        return clampNormalise(raw, lb)
    }
    // Heap path: 6 allocations of float64 slices.
    raw := swgDPRaw(a, b, la, lb, params,
        make([]float64, lb+1), make([]float64, lb+1),
        make([]float64, lb+1), make([]float64, lb+1),
        make([]float64, lb+1), make([]float64, lb+1))
    return clampNormalise(raw, lb)
}
```

Note: `maxStackInputLen` is **inherited** from `levenshtein.go` line 68 — do NOT redeclare. The `isASCII` helper is **inherited** from `normalise.go` lines 159–168 — do NOT redeclare.

**Rune variant pattern** — copy from `levenshtein.go` lines 167–192 with rune-aware kernel; identity short-circuit per IN-02 (commit c235e0e):

```go
func SmithWatermanGotohScoreRunes(a, b string) float64 {
    if a == b {
        return 1.0 // fast identity — saves two []rune allocations (IN-02 pattern)
    }
    ra := []rune(a) // 1 alloc
    rb := []rune(b) // 1 alloc
    // ... rune-aware DP, same shape as smithWatermanGotohRawByte/Rune split
}
```

**SWGParams + NewSWGParams** — new shape decided in CONTEXT.md §3 (no Phase 2 analog — SWG is the first algorithm with a params struct). Documented field-by-field in the godoc block per CONTEXT.md §3 sample.

**DP kernel `swgDPRaw`** — three-matrix two-row form per RESEARCH.md §Pattern 1 lines 283–339; written fresh from the corrected recurrence, no code copied from biopython / EMBOSS / any Go implementation. `algorithm-licensing-reviewer` gate per CLAUDE.md.

---

### `dispatch_swg.go` (new-implementation)

**Analog:** `dispatch_levenshtein.go` (35 lines — copy character-for-character)

**Full file pattern** (copy and change only the AlgoXxx and XxxScore identifiers):

```go
// Copyright 2026 AxonOps Limited
// [Apache-2.0 header — copy verbatim from dispatch_levenshtein.go lines 1–13]

// dispatch_swg.go registers SmithWatermanGotohScore into the dispatch table
// at package load time. This file MUST be the sole writer to
// dispatch[AlgoSmithWatermanGotoh].

package fuzzymatch

var _ = func() bool {
    dispatch[AlgoSmithWatermanGotoh] = SmithWatermanGotohScore
    return true
}()
```

**Source pattern** (analog lines 31–34, locked Phase 2 idiom — no `init()` per determinism-standards §13.5):

```go
var _ = func() bool {
    dispatch[AlgoLevenshtein] = LevenshteinScore
    return true
}()
```

`AlgoSmithWatermanGotoh` is already declared at `algoid.go` line 102 (slot 6 of 23); Phase 3 only populates the slot.

---

### `swg_test.go` (new-test)

**Analog:** `levenshtein_test.go`

**Header + package + import pattern** (analog lines 1–30):

```go
// Copyright 2026 AxonOps Limited
// [Apache-2.0 header verbatim]

// swg_test.go pins the public-API contract of swg.go: identity, both-empty,
// one-empty, canonical reference vectors from Smith-Waterman 1981 / Gotoh
// 1982 (corrected per Flouri et al. 2015), symmetry, byte vs rune path
// equivalence on ASCII inputs, multi-byte rune handling, NaN/Inf guards,
// the SWGParams construction and default semantics, the Raw* unclamped
// surface, AND cross-validation against the biopython reference corpus.
//
// Stdlib `testing` only — no testify in root tests, per
// .claude/skills/go-coding-standards.

package fuzzymatch_test

import (
    "encoding/json"
    "math"
    "os"
    "path/filepath"
    "testing"

    "github.com/axonops/fuzzymatch"
)
```

**Both-empty / identity / one-empty / reference-vector test shape** — copy lines 46–121 of analog, substituting SmithWatermanGotohScore + SmithWatermanGotohScoreWithParams. Reference vectors come from Smith-Waterman 1981 + Gotoh 1982 + the biopython oracle (verify each one independently in the corpus).

**`TestSWG_CrossValidation` is unique to Phase 3 — see §unique-files-3 below.**

---

### `swg_bench_test.go` (new-benchmark)

**Analog:** `levenshtein_bench_test.go` (103 lines)

**Header + per-benchmark structure** (analog lines 14–55, copy idiom):

```go
// swg_bench_test.go runs allocation-aware benchmarks for SmithWatermanGotohScore
// at six input sizes:
//   - ASCII_Short (≤ 64 bytes, stack path):   target < 2 µs/op, 0 allocs/op
//   - ASCII_Medium (50 bytes, stack path):    target 0 allocs/op
//   - ASCII_Long (> 64 bytes, heap path):     6 allocs/op (six float64 row slices)
//   - Unicode_Short (rune path):              8 allocs/op (2 []rune + 6 row slices)
//   - WithParams_ASCII_Short:                 same as ASCII_Short with custom params
//   - RawScore_ASCII_Short:                   exercises the unclamped path

func BenchmarkSmithWatermanGotohScore_ASCII_Short(b *testing.B) {
    b.ReportAllocs()
    b.ResetTimer()
    var sink float64
    for i := 0; i < b.N; i++ {
        sink = fuzzymatch.SmithWatermanGotohScore("kitten", "sitting")
    }
    if sink < 0 {
        b.Fatal("sink unexpectedly negative — compiler folded the benchmark away")
    }
}
```

**Two extra benches beyond Phase 2** — per CONTEXT.md §2 ("ALSO add `_WithParams_ASCII_Short` and `_RawScore_ASCII_Short`"):

```go
func BenchmarkSmithWatermanGotohScore_WithParams_ASCII_Short(b *testing.B) {
    params := fuzzymatch.NewSWGParams()
    params.Match = 2.0
    // ... ReportAllocs/ResetTimer/loop calling SmithWatermanGotohScoreWithParams
}

func BenchmarkSmithWatermanGotohRawScore_ASCII_Short(b *testing.B) {
    // ... loop calling SmithWatermanGotohRawScore
}
```

---

### `swg_fuzz_test.go` (new-fuzz)

**Analog:** `levenshtein_fuzz_test.go` (74 lines — exact copy with substituted identifier)

**Full pattern** (analog lines 40–73):

```go
func FuzzSmithWatermanGotohScore(f *testing.F) {
    for _, pair := range []struct{ a, b string }{
        {"kitten", "sitting"},      // canonical reference vector
        {"http_request", "http_request_header_fields"}, // substring containment
        {"abc", "abc"},             // identical
        {"", "abc"},                // one-empty
        {"", ""},                   // both-empty
        {"abc________def", "abcdef"}, // one-long-gap canary (Gotoh erratum)
        {"\xff\xfe", "abc"},        // invalid UTF-8 (byte path resilience)
        {"Привет", "привет"},       // Cyrillic (multi-byte UTF-8)
    } {
        f.Add(pair.a, pair.b)
    }
    f.Fuzz(func(t *testing.T, a, b string) {
        got := fuzzymatch.SmithWatermanGotohScore(a, b)
        if math.IsNaN(got) { t.Errorf("...") }
        if math.IsInf(got, 0) { t.Errorf("...") }
        if got < 0.0 || got > 1.0 { t.Errorf("...") }
    })
}
```

---

### `tests/bdd/features/swg.feature` (new-feature-file)

**Analog:** `tests/bdd/features/levenshtein.feature` (40 lines)

**Header + Feature block + scenarios** (analog lines 1–40):

```gherkin
# Primary source: Smith, T. F. & Waterman, M. S. (1981). "Identification of
# common molecular subsequences." J. Mol. Biol. 147:195-197.
# Affine-gap reduction: Gotoh, O. (1982). J. Mol. Biol. 162:705-708.
# Corrected initialisation: Flouri, T. et al. (2015). biorxiv 031500.
#
# Score normalisation: clamp(best_local_score / min(len(a), len(b)), 0, 1).

Feature: Smith-Waterman-Gotoh local-alignment similarity
  Local alignment with affine gap penalty. Default params (Match=1.0,
  Mismatch=-1.0, GapOpen=-1.5, GapExtend=-0.5) match the documented defaults.

  Scenario Outline: canonical reference vectors
    When I compute the SmithWatermanGotoh score between "<a>" and "<b>"
    Then the score should be approximately <score> within 0.0001

    Examples:
      | a            | b                            | score  |
      | http_request | http_request_header_fields   | 1.0000 |
      | kitten       | sitting                      | 0.5714 |
      | abc          | abc                          | 1.0000 |

  Scenario: identical strings score 1.0
    When I compute the SmithWatermanGotoh score between "user_id" and "user_id"
    Then the score should be exactly 1

  Scenario: both-empty strings score 1.0
    When I compute the SmithWatermanGotoh score between "" and ""
    Then the score should be exactly 1

  Scenario: one-empty string scores 0.0
    When I compute the SmithWatermanGotoh score between "abc" and ""
    Then the score should be exactly 0

  Scenario: score is symmetric
    When I compute the SmithWatermanGotoh score between "kitten" and "sitting"
    And I compute the second SmithWatermanGotoh score between "sitting" and "kitten"
    Then both SmithWatermanGotoh scores should be equal
```

**Gap-split canary scenario** (NEW for SWG, the load-bearing PITFALLS.md §3 gate):

```gherkin
  Scenario: gap-split canary — splitting a long gap does not improve the score
    When I compute the SmithWatermanGotoh score between "abc________def" and "abcdef"
    And I compute the second SmithWatermanGotoh score between "abc____def" and "abcdef"
    Then both SmithWatermanGotoh scores should be equal
```

**Integer-form score regex** — per IN-03 (commit 8802d0b), the BDD regex `(\d+\.?\d*)` accepts both `0`/`1` and `0.0`/`1.0`. Feature scenarios above use the integer form for clarity.

---

### `testdata/golden/_staging/swg.json` (new-golden-staging)

**Analog:** `testdata/golden/_staging/levenshtein.json`

**Schema** (Phase-2-locked):

```json
{
  "version": 1,
  "entries": [
    {
      "name": "SmithWatermanGotoh_both_empty",
      "algorithm": "SmithWatermanGotoh",
      "a": "",
      "b": "",
      "expected_score": 1
    },
    ...
  ]
}
```

**Production mechanism** — written by `TestGolden_SmithWatermanGotoh_Staging` (added to `algorithms_golden_test.go`) calling `assertGoldenStaging(t, "_staging/swg.json", file)` with the `-update` flag, mirroring `TestGolden_Levenshtein_Staging` at `algorithms_golden_test.go` lines 203–210. Entries SORTED alphabetically by Name. Algorithm string is `"SmithWatermanGotoh"` (no spaces/dashes, matches the `AlgoSmithWatermanGotoh.String()` form).

**Required entry seeds** (per CONTEXT.md §3 / RESEARCH.md identity / both-empty / one-empty / substring / no-overlap / one-long-gap-canary / non-default-params):

- `SmithWatermanGotoh_both_empty` (1.0)
- `SmithWatermanGotoh_identical` (1.0)
- `SmithWatermanGotoh_one_empty` (0.0)
- `SmithWatermanGotoh_two_substring` (`http_request` / `http_request_header_fields`)
- `SmithWatermanGotoh_no_overlap` (`qqqq` / `zzzz`)
- `SmithWatermanGotoh_one_long_gap_canary` (`abc________def` / `abcdef`)

---

### `testdata/fuzz/FuzzSmithWatermanGotohScore/seed-001` (new-fuzz-seed)

**Analog:** `testdata/fuzz/FuzzLevenshteinScore/seed-001`

**Exact format** (Go fuzz corpus v1 literal):

```
go test fuzz v1
string("http_request")
string("http_request_header_fields")
```

The substring-containment pair is a sensible default seed because it exercises the SWG-vs-Levenshtein divergence the cross-algorithm consistency test pins.

---

## Pattern Assignments — Extended Files

### `props_test.go` (extend-existing)

**Analog:** own analog — Levenshtein block (lines 65–158) for the byte-path standard invariants, and the rune-symmetry trailer (lines 727–793) for the `*Runes` symmetry function.

**Locator pattern for new block** — append SWG block immediately BEFORE the `Rune-path symmetry property tests` separator at line 727. The boundary comment to look for is exactly:

```go
// ---------------------------------------------------------------------------
// Rune-path symmetry property tests (WR-03)
// ---------------------------------------------------------------------------
```

**SWG block template** (copy structure from the Levenshtein block at lines 65–158, substitute identifiers, OMIT the triangle-inequality test per CONTEXT.md §5):

```go
// ---------------------------------------------------------------------------
// Smith-Waterman-Gotoh property tests
// ---------------------------------------------------------------------------
//
// Standard Phase 2 invariants (range bounds, identity, symmetric, NoNaN,
// NoInf, NoNegativeZero) plus three SWG-specific canaries per PITFALLS.md
// §3 (GapSplitInvariance, RawNeverExceedsMatchTimesMinLen, MonotonicWithMatchReward).
// Triangle inequality is OMITTED — SWG is not a metric over the full string
// space.

func TestProp_SmithWatermanGotohScore_RangeBounds(t *testing.T) {
    f := func(a, b string) bool {
        s := fuzzymatch.SmithWatermanGotohScore(a, b)
        return s >= 0.0 && s <= 1.0
    }
    if err := quick.Check(f, nil); err != nil {
        t.Errorf("SmithWatermanGotohScore out of [0,1]: %v", err)
    }
}
// ... Identity / Symmetric / NoNaN / NoInf / NoNegativeZero per analog template
```

**SWG-specific property tests** (new in Phase 3 — no Phase 2 template, document per CONTEXT.md §5):

```go
// TestProp_SmithWatermanGotoh_GapSplitInvariance — hand-curated triples
// exercising the canonical Gotoh-erratum case (PITFALLS.md §3 warning sign).
// Splitting a single long gap into two halves with intervening match characters
// that don't affect the local alignment must NOT improve the score.

// TestProp_SmithWatermanGotoh_RawNeverExceedsMatchTimesMinLen — invariant:
//   RawScore(a, b) <= Match * min(len(a), len(b))
// Upper bound from "best local alignment has at most min(len) match positions".

// TestProp_SmithWatermanGotoh_MonotonicWithMatchReward — increasing the Match
// parameter (keeping others fixed) cannot decrease RawScore for any input pair.
```

**Rune-symmetry function** — append `TestProp_SmithWatermanGotohScoreRunes_Symmetric` at the end of the file (after line 793, matching the existing pattern for the other six algorithms — analog lines 786–793):

```go
func TestProp_SmithWatermanGotohScoreRunes_Symmetric(t *testing.T) {
    f := func(a, b string) bool {
        return fuzzymatch.SmithWatermanGotohScoreRunes(a, b) == fuzzymatch.SmithWatermanGotohScoreRunes(b, a)
    }
    if err := quick.Check(f, nil); err != nil {
        t.Errorf("SmithWatermanGotohScoreRunes not symmetric: %v", err)
    }
}
```

---

### `example_test.go` (extend-existing)

**Analog:** own analog — `ExampleLevenshteinScore` (lines 34–38).

**Locator pattern** — append after the last Example function (`ExampleJaroWinklerScore` at lines 93–101). No section boundary comment; append at EOF.

**SWG example template** (one Example per public-surface variant — at minimum the normalised default and the Raw variant per CONTEXT.md §4):

```go
// ExampleSmithWatermanGotohScore demonstrates the SWG local-alignment
// similarity on a substring-containment pair. The shorter input is fully
// contained in the longer; the local alignment finds the full match, so
// the normalised score (clamp(raw / min(len), 0, 1)) is 1.0000.
func ExampleSmithWatermanGotohScore() {
    fmt.Printf("%.4f\n", fuzzymatch.SmithWatermanGotohScore("http_request", "http_request_header_fields"))
    // Output:
    // 1.0000
}

// ExampleSmithWatermanGotohRawScore demonstrates the unclamped raw alignment
// score. For the same substring-containment pair, the raw score equals
// Match × min(len) = 1.0 × 12 = 12.0 (twelve match positions, no gap penalty).
func ExampleSmithWatermanGotohRawScore() {
    fmt.Printf("%.1f\n", fuzzymatch.SmithWatermanGotohRawScore("http_request", "http_request_header_fields"))
    // Output:
    // 12.0
}
```

---

### `algoid_test.go` (extend-existing)

**Analog:** own analog — `TestDispatch_LevenshteinRegistered` (lines 230–235) and the `registered` map in `TestDispatch_UnregisteredSlotsAreNil` (lines 289–312).

**Locator pattern 1 — add new dispatch-registered test** — append before `TestDispatch_UnregisteredSlotsAreNil` at line 289. Boundary comment to look for:

```go
// TestDispatch_UnregisteredSlotsAreNil asserts that all dispatch slots except
```

Template (copy analog at lines 230–235):

```go
// TestDispatch_SmithWatermanGotohRegistered asserts that
// dispatch[AlgoSmithWatermanGotoh] (slot 6) is non-nil after Phase 3 plan
// 03-01 registers SmithWatermanGotohScore.
func TestDispatch_SmithWatermanGotohRegistered(t *testing.T) {
    if fuzzymatch.DispatchEntryNilForTest(int(fuzzymatch.AlgoSmithWatermanGotoh)) {
        t.Errorf("dispatch[AlgoSmithWatermanGotoh] (%d) is nil — dispatch_swg.go must register SmithWatermanGotohScore at package load time",
            int(fuzzymatch.AlgoSmithWatermanGotoh))
    }
}
```

**Locator pattern 2 — flip slot 6 in `registered` map** — analog lines 292–299 (the map literal):

```go
registered := map[int]bool{
    int(fuzzymatch.AlgoLevenshtein):            true,
    int(fuzzymatch.AlgoDamerauLevenshteinOSA):  true,
    int(fuzzymatch.AlgoDamerauLevenshteinFull): true,
    int(fuzzymatch.AlgoHamming):                true,
    int(fuzzymatch.AlgoJaro):                   true,
    int(fuzzymatch.AlgoJaroWinkler):            true,
    int(fuzzymatch.AlgoSmithWatermanGotoh):     true,  // ADD THIS LINE
}
```

Also update the godoc above the function (lines 284–288) to mention `AlgoSmithWatermanGotoh (slot 6)` is now registered.

---

### `algorithms_golden_test.go` (extend-existing)

**Analog:** own analog — `buildLevenshteinStagingEntries` + `TestGolden_Levenshtein_Staging` (lines 196–210); the merge `stagingFiles` list (lines 163–170) in `TestGolden_Algorithms_Merge`.

**Locator pattern 1 — staging-files merge list** (lines 163–170, add one entry alphabetically — `swg.json` sorts after `levenshtein.json`):

```go
stagingFiles := []string{
    "_staging/damerau_full.json",
    "_staging/damerau_osa.json",
    "_staging/hamming.json",
    "_staging/jaro.json",
    "_staging/jarowinkler.json",
    "_staging/levenshtein.json",
    "_staging/swg.json",  // ADD THIS LINE
}
```

**Locator pattern 2 — add `buildSWGStagingEntries` + `TestGolden_SmithWatermanGotoh_Staging`** at EOF of the file (matching the Wave 2 pattern — analog `buildJaroWinklerStagingEntries` + `TestGolden_JaroWinkler_Staging` at line 574+):

```go
// buildSWGStagingEntries returns the Smith-Waterman-Gotoh entries used by
// TestGolden_SmithWatermanGotoh_Staging. ExpectedScore is computed from the
// current implementation so the staging file stays in sync with actual output.
func buildSWGStagingEntries(t *testing.T) []goldenAlgorithmEntry {
    t.Helper()
    return []goldenAlgorithmEntry{
        {Name: "SmithWatermanGotoh_both_empty", Algorithm: "SmithWatermanGotoh", A: "", B: "",
            ExpectedScore: fuzzymatch.SmithWatermanGotohScore("", "")},
        // ... identical / one_empty / two_substring / no_overlap / one_long_gap_canary
    }
}

func TestGolden_SmithWatermanGotoh_Staging(t *testing.T) {
    entries := buildSWGStagingEntries(t)
    sort.Slice(entries, func(i, j int) bool { return entries[i].Name < entries[j].Name })
    file := goldenAlgorithmsFile{Version: 1, Entries: entries}
    assertGoldenStaging(t, "_staging/swg.json", file)
}
```

---

### `cross_algorithm_consistency_test.go` (extend-existing)

**Analog:** own analog — `TestCrossAlgorithm_OSA_Full_Divergence` (lines 57–75) for the divergence pattern; `TestCrossAlgorithm_IdentityConvergence` (lines 81–105) for the funcs-slice pattern.

**Locator pattern** — append before EOF (line 192). Optionally add SWG to the `funcs` slices in the four existing convergence tests (lines 87–94, 117–124, 175–182) so SWG identity / both-empty / one-empty convergence is also pinned.

**New SWG-vs-Levenshtein divergence test** (per CONTEXT.md `<code_context>` paragraph "substring-containment input"):

```go
// TestCrossAlgorithm_SWG_Levenshtein_SubstringDivergence asserts the
// load-bearing local-vs-global-alignment claim: on a substring-containment
// input, Smith-Waterman-Gotoh (local alignment) scores STRICTLY HIGHER than
// Levenshtein (global edit distance), because SWG finds the substring while
// Levenshtein counts every uncovered position as an edit.
//
// "http_request" is fully contained in "http_request_header_fields":
//   - SmithWatermanGotohScore = 1.0 (full local match found)
//   - LevenshteinScore       ≈ 0.46 (13 edits over max-length 26)
func TestCrossAlgorithm_SWG_Levenshtein_SubstringDivergence(t *testing.T) {
    a, b := "http_request", "http_request_header_fields"
    gotSWG := fuzzymatch.SmithWatermanGotohScore(a, b)
    gotLev := fuzzymatch.LevenshteinScore(a, b)
    if !(gotSWG > gotLev) {
        t.Errorf("SWG (%v) must score strictly higher than Levenshtein (%v) on substring-containment pair %q/%q (local-vs-global divergence)",
            gotSWG, gotLev, a, b)
    }
}
```

---

### `bench.txt` (extend-existing)

**Analog:** own analog — the file is the canonical benchstat-format dump of `make bench`. Lines 5–14 (BenchmarkAlgoID_String × 10) show the per-benchmark format.

**Locator pattern** — entries are sorted alphabetically by benchmark name. `BenchmarkSmithWatermanGotohScore_*` series insert after `BenchmarkLevenshteinScore_Unicode_Short` block (or wherever the alphabetical sort places them — `S` comes after `L`). Six new series × 10 rows each = 60 new lines.

**Production mechanism** — run `make bench` after the implementation lands and replace `bench.txt` with the new dump (full file replacement is the locked workflow). Do NOT hand-edit individual rows.

---

### `llms.txt` (extend-existing)

**Analog:** own analog — Levenshtein block at lines 47–53.

**Locator pattern** — `### Smith-Waterman-Gotoh` section inserts in catalogue order (after Jaro-Winkler at lines 80–83, before Normalisation at line 85). The `ai_friendly_test.go` meta-test verifies every exported symbol is listed.

**SWG block template** (matches the analog line-shape — `- func Name(args) ReturnType`, one per line):

```
### Smith-Waterman-Gotoh local-alignment similarity

- type SWGParams struct
- func NewSWGParams() SWGParams
- func SmithWatermanGotohScore(a, b string) float64
- func SmithWatermanGotohScoreRunes(a, b string) float64
- func SmithWatermanGotohScoreWithParams(a, b string, params SWGParams) float64
- func SmithWatermanGotohRawScore(a, b string) float64
- func SmithWatermanGotohRawScoreRunes(a, b string) float64
- func SmithWatermanGotohRawScoreWithParams(a, b string, params SWGParams) float64
```

Nine new lines (1 type + 1 constructor + 6 funcs + section header). `AlgoSmithWatermanGotoh` is already listed at line 30 (added in Phase 1).

---

### `examples/identifier-similarity/main.go` (extend-existing)

**Analog:** own analog — `algorithms` slice at lines 71–81.

**Locator pattern** — append an SWG entry to the slice. With 7 algorithms, the `algoWidth=13` column width may need to shrink to accommodate the table width (review at execution time):

```go
var algorithms = []struct {
    name string
    fn   func(a, b string) float64
}{
    {"Levenshtein", fuzzymatch.LevenshteinScore},
    {"DL-OSA", fuzzymatch.DamerauLevenshteinOSAScore},
    {"DL-Full", fuzzymatch.DamerauLevenshteinFullScore},
    {"Hamming", fuzzymatch.HammingScore},
    {"Jaro", fuzzymatch.JaroScore},
    {"Jaro-Winkler", fuzzymatch.JaroWinklerScore},
    {"SWG", fuzzymatch.SmithWatermanGotohScore},  // ADD
}
```

The CONTEXT.md mentions "7-row × 6-column → 7-row × 7-column" (i.e. column count goes 6 → 7 for algorithms; the pair column makes it 7 → 8 total table columns). Display name `"SWG"` keeps the 13-char column width unchanged.

---

### `examples/identifier-similarity/main_test.go` (extend-existing)

**Analog:** own analog — the `want` constant at lines 40–49.

**Locator pattern** — replace the `want` constant with a regenerated version containing the SWG column. Procedure: run `go run ./examples/identifier-similarity/` AFTER `main.go` is updated, capture stdout, paste into the `want` constant. The `TestExample_ColumnWidths` test self-adapts (it derives column widths from `want`).

The `TestExample_Output` test (lines 58–105, IN-04 line-by-line diff) requires no changes — its byte-for-byte assertion catches any drift in the regenerated `want`.

---

### `tests/bdd/steps/algorithms_steps.go` (extend-existing)

**Analog:** own analog — Levenshtein step methods (lines 51–93); `InitializeScenario` Levenshtein step registrations (lines 302–322).

**Locator pattern 1 — step method block** — append after the Damerau-Levenshtein Full block (file EOF). Boundary comment style:

```go
// ---------------------------------------------------------------------------
// Smith-Waterman-Gotoh step definitions (plan 03-01)
// ---------------------------------------------------------------------------

func (ctx *AlgorithmContext) iComputeTheSmithWatermanGotohScoreBetween(a, b string) error {
    ctx.lastScore = fuzzymatch.SmithWatermanGotohScore(a, b)
    return nil
}

func (ctx *AlgorithmContext) iComputeTheSecondSmithWatermanGotohScoreBetween(a, b string) error {
    ctx.lastScore2 = fuzzymatch.SmithWatermanGotohScore(a, b)
    return nil
}

func (ctx *AlgorithmContext) bothSmithWatermanGotohScoresShouldBeEqual() error {
    if ctx.lastScore != ctx.lastScore2 {
        return fmt.Errorf("scores not equal: %f != %f", ctx.lastScore, ctx.lastScore2)
    }
    return nil
}
```

**Locator pattern 2 — `InitializeScenario` registration block** — append before the closing brace at line 413. Boundary marker is the comment at line 396 (`// Damerau-Levenshtein Full step definitions (plan 02-06)`); add SWG block immediately after the Full block:

```go
// Smith-Waterman-Gotoh step definitions (plan 03-01).
ctx.Step(
    `^I compute the SmithWatermanGotoh score between "([^"]*)" and "([^"]*)"$`,
    a.iComputeTheSmithWatermanGotohScoreBetween,
)
ctx.Step(
    `^I compute the second SmithWatermanGotoh score between "([^"]*)" and "([^"]*)"$`,
    a.iComputeTheSecondSmithWatermanGotohScoreBetween,
)
ctx.Step(
    `^both SmithWatermanGotoh scores should be equal$`,
    a.bothSmithWatermanGotohScoresShouldBeEqual,
)
```

**Note:** the `theScoreShouldBeApproximately` / `theScoreShouldBeExactly` / `theDistanceShouldBe` steps are already algorithm-agnostic (registered once in the Levenshtein block at lines 311–318 and the Hamming block at lines 341–344 respectively); SWG reuses them. Per IN-06 (commit 8802d0b), the distance step is algorithm-agnostic by design — SWG has no Distance variant so no new distance step is required.

---

### `Makefile` (extend-existing)

**Analog:** own analog — `verify-determinism` target (lines 188–191) for the bash-target style.

**Locator pattern** — add new target after `verify-license-headers` (line 194), before `release-check`. Also extend `.PHONY` list at lines 26–28.

**Target template** (developer-only — invokes the Python script; tolerant if biopython not installed):

```makefile
# Regenerates testdata/cross-validation/swg/vectors.json by invoking the
# biopython-based generator script. Developer-only (not in `make check`);
# the committed JSON is the verification fixture.
#
# Requires: python3 -m pip install --user biopython
regen-swg-cross-validation:
	@if ! command -v python3 >/dev/null 2>&1; then \
	  echo "python3 not found; install Python 3.x and run: python3 -m pip install --user biopython"; \
	  exit 1; \
	fi
	python3 scripts/gen-swg-cross-validation.py
```

Add `regen-swg-cross-validation` to the `.PHONY` list (line 26–28).

`TestSWG_CrossValidation` runs as part of the default `make test` / `make check` cycle (CONTEXT.md §1: "default `go test ./...` includes `TestSWG_CrossValidation`"); no separate Makefile target needed for verification.

---

### `docs/requirements.md` (roadmap-update)

**Analog:** §7.1.8 own analog.

**Locator pattern** — find the `### 7.1.8 Smith-Waterman-Gotoh` (or equivalent) section and update the public-surface table per CONTEXT.md §4: extend from 3 functions (Score / ScoreRunes / ScoreWithParams) to 6 (add Raw equivalents). The `api-ergonomics-reviewer` flags the expansion during PR review per CONTEXT.md §4.

---

## Shared Patterns

### Apache-2.0 file header
**Source:** `levenshtein.go` lines 1–13 (verbatim across all 6 Phase 2 `.go` files).
**Apply to:** `swg.go`, `dispatch_swg.go`, `swg_test.go`, `swg_bench_test.go`, `swg_fuzz_test.go`. Verified at CI by `scripts/verify-license-headers.sh`.

### `maxStackInputLen` + `isASCII` are SHARED constants
**Source:** `levenshtein.go` line 68 (`const maxStackInputLen = 64`) and `normalise.go` lines 159–168 (`isASCII`).
**Apply to:** `swg.go` references them by name. MUST NOT redeclare (build error).

### Dispatch registration via `var _ = func() bool { ... }()`
**Source:** `dispatch_levenshtein.go` lines 31–34 (LOCKED Phase 2 idiom; no `init()` per determinism-standards §13.5).
**Apply to:** `dispatch_swg.go` only.

### Identity short-circuit on `*Runes`
**Source:** `levenshtein.go` lines 177–180 (per IN-02 cleanup commit c235e0e — `if a == b { return 1.0 }` saves the two `[]rune` conversions on the identity path).
**Apply to:** `SmithWatermanGotohScoreRunes`, `SmithWatermanGotohRawScoreRunes` in `swg.go`.

### BDD score regex accepts integer-form
**Source:** `tests/bdd/steps/algorithms_steps.go` lines 312, 316 (per IN-03 cleanup commit 8802d0b — `(\d+\.?\d*)` matches `0`, `1`, `0.0`, `1.0`).
**Apply to:** `tests/bdd/features/swg.feature` — feature scenarios may use the integer form (`Then the score should be exactly 1`).

### Stdlib `testing` only in root
**Source:** CLAUDE.md "Constraints"; verified across `levenshtein_test.go`, `props_test.go`, etc.
**Apply to:** `swg_test.go`, `swg_bench_test.go`, `swg_fuzz_test.go`, all extensions to existing `_test.go` files. NO testify in root. testify in `tests/bdd/` is permitted but not required for the SWG step block (the existing block uses plain error returns).

### Append-only shared files
**Source:** Locked Phase 2 pattern per CONTEXT.md §carry_forward.
**Apply to:** `props_test.go`, `example_test.go`, `algoid_test.go`, `algorithms_golden_test.go`, `cross_algorithm_consistency_test.go`, `tests/bdd/steps/algorithms_steps.go`, `llms.txt`, `bench.txt`. Each is extended by appending a SWG block — no rewrites of existing test bodies.

---

## Unique-to-Phase-3 Files (No Phase 2 Analog)

### §unique-files-1 — `testdata/cross-validation/swg/vectors.json` (new-cross-validation-corpus)

**No Phase 2 analog.** This is the first cross-validation corpus in the repository (Phase 2's golden files pin Go-output stability; this fixture pins agreement with an external reference implementation).

**JSON schema** (locked by CONTEXT.md §1; RESEARCH.md §user_constraints):

```json
{
  "version": 1,
  "biopython_version": "1.85",
  "entries": [
    {
      "name": "identity_short",
      "a": "hello",
      "b": "hello",
      "params": {"match": 1.0, "mismatch": -1.0, "gap_open": -1.5, "gap_extend": -0.5},
      "biopython_score": 5.0,
      "biopython_normalised": 1.0
    },
    {
      "name": "one_long_gap_canary",
      "a": "abc________def",
      "b": "abcdef",
      "params": {"match": 1.0, "mismatch": -1.0, "gap_open": -1.5, "gap_extend": -0.5},
      "biopython_score": 6.0,
      "biopython_normalised": 1.0
    },
    ...
  ]
}
```

**Required entries** (~10–20 minimum per CONTEXT.md §1; MUST include all six categories):
1. `identity_short` (e.g. `"hello"`/`"hello"`)
2. `both_empty` (`""`/`""`)
3. `one_empty_a` (`""`/`"abcdef"`)
4. `one_empty_b` (`"abcdef"`/`""`)
5. `two_substring` (`"http_request"`/`"http_request_header_fields"`)
6. `no_overlap` (`"qqqq"`/`"zzzz"`)
7. `one_long_gap_canary` (`"abc________def"`/`"abcdef"`) — Gotoh-erratum gate
8. `non_default_params` (custom Match/Mismatch/GapOpen/GapExtend)
9. ~5–10 additional spanning unicode / single-char / all-mismatch / partial-middle-match / etc.

**Production mechanism** — written by `scripts/gen-swg-cross-validation.py` via `make regen-swg-cross-validation`. The JSON is COMMITTED to the repo; CI reads it without Python. Regeneration after a biopython version bump is a deliberate developer action.

**Determinism** — the script's deterministic `CASES` list ordering owns the corpus's byte-stability. Field ordering in each entry: `name`, `a`, `b`, `params`, `biopython_score`, `biopython_normalised` (matches the schema declaration order; Python `json.dump(obj, indent=2, sort_keys=False)` preserves insertion order on Python 3.7+).

---

### §unique-files-2 — `scripts/gen-swg-cross-validation.py` (new-script)

**No Phase 2 analog.** First Python file in the repo. `scripts/` directory previously held only bash files (`verify-coverage-floors.sh`, `verify-license-headers.sh`, `verify-no-runtime-deps.sh`).

**Python skeleton** (per RESEARCH.md §Pattern 5 lines 455–534):

```python
#!/usr/bin/env python3
# scripts/gen-swg-cross-validation.py
#
# Regenerates testdata/cross-validation/swg/vectors.json from biopython's
# Bio.Align.PairwiseAligner (BSD-3-Clause licensed; permissive, compatible
# with Apache-2.0 for reference-vector cross-validation per
# .claude/skills/algorithm-licensing-standards).
#
# Run via:  make regen-swg-cross-validation
# Requires: python3 -m pip install --user biopython (1.85+)
#
# The script computes BOTH the raw biopython alignment score AND the
# script-side normalised reference (clamp(raw / min(len(a), len(b)), 0, 1)).
# The Go test compares against `biopython_normalised` with zero in-Go
# normalisation logic — the script owns the reference normalisation.

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
    ("no_overlap",           "qqqq",           "zzzz",                           None),
    ("one_long_gap_canary",  "abc________def", "abcdef",                         None),
    ("non_default_params",   "hello",          "hallo",
        {"match": 2.0, "mismatch": -2.0, "gap_open": -3.0, "gap_extend": -1.0}),
    # ... additional 5-10 cases
]

def score_case(a, b, params):
    aligner = PairwiseAligner()
    aligner.mode = "local"
    aligner.match_score = params["match"]
    aligner.mismatch_score = params["mismatch"]
    aligner.open_gap_score = params["gap_open"]
    aligner.extend_gap_score = params["gap_extend"]
    if a == "" and b == "":
        return 0.0, 1.0   # both-empty convention: raw 0, normalised 1
    if a == "" or b == "":
        return 0.0, 0.0   # one-empty convention: raw 0, normalised 0
    raw = aligner.score(a, b)
    min_len = min(len(a), len(b))
    norm = raw / min_len
    norm = max(0.0, min(1.0, norm))   # clamp(raw/min_len, 0, 1)
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
        f.write("\n")  # trailing newline matches Phase 2 golden-file convention

if __name__ == "__main__":
    main()
```

**License-header treatment** — the verify-license-headers.sh script (per Makefile line 194) targets `.go` files; Python scripts are typically NOT covered. Confirm at execution time; if Python files are checked, add the Apache-2.0 header as a `#`-comment block at the top per the `.sh` convention. Otherwise the docstring comment serves as the in-file attribution to biopython's BSD-3-Clause licence.

**Determinism** — the `CASES` list is the single source of ordering. `json.dump(..., sort_keys=False)` preserves Python's dict insertion order (guaranteed in 3.7+). No randomness, no environment dependence beyond biopython's deterministic scoring.

---

### §unique-files-3 — `TestSWG_CrossValidation` (inside `swg_test.go`) (new-test)

**No Phase 2 analog.** First test that compares Go output against an external reference corpus.

**Test loader skeleton** (per CONTEXT.md §1 and the system architecture diagram in RESEARCH.md):

```go
// TestSWG_CrossValidation asserts agreement between our SmithWatermanGotoh
// implementation and the biopython reference corpus committed at
// testdata/cross-validation/swg/vectors.json.
//
// Tolerance: |our_score - biopython_normalised| <= 1e-9 (matches the
// cross_algorithm_consistency_test.go epsilon convention).
//
// The corpus is regenerated by `make regen-swg-cross-validation` (developer-
// only); CI does NOT require Python. If this test fails after a corpus
// regeneration, EITHER our DP kernel drifted from the corrected Gotoh
// formulation OR the biopython version emitted different scores (record the
// biopython version from the corpus header in the failure message).
func TestSWG_CrossValidation(t *testing.T) {
    const epsilon = 1e-9
    type entry struct {
        Name                 string  `json:"name"`
        A                    string  `json:"a"`
        B                    string  `json:"b"`
        Params               struct {
            Match     float64 `json:"match"`
            Mismatch  float64 `json:"mismatch"`
            GapOpen   float64 `json:"gap_open"`
            GapExtend float64 `json:"gap_extend"`
        } `json:"params"`
        BiopythonScore       float64 `json:"biopython_score"`
        BiopythonNormalised  float64 `json:"biopython_normalised"`
    }
    type corpus struct {
        Version          int     `json:"version"`
        BiopythonVersion string  `json:"biopython_version"`
        Entries          []entry `json:"entries"`
    }
    path := filepath.Join("testdata", "cross-validation", "swg", "vectors.json")
    raw, err := os.ReadFile(path)
    if err != nil {
        t.Fatalf("TestSWG_CrossValidation: read %s: %v (regenerate with `make regen-swg-cross-validation`)", path, err)
    }
    var c corpus
    if err := json.Unmarshal(raw, &c); err != nil {
        t.Fatalf("TestSWG_CrossValidation: parse %s: %v", path, err)
    }
    if c.Version != 1 {
        t.Fatalf("TestSWG_CrossValidation: unsupported corpus version %d (want 1)", c.Version)
    }
    for _, e := range c.Entries {
        t.Run(e.Name, func(t *testing.T) {
            params := fuzzymatch.SWGParams{
                Match: e.Params.Match, Mismatch: e.Params.Mismatch,
                GapOpen: e.Params.GapOpen, GapExtend: e.Params.GapExtend,
            }
            got := fuzzymatch.SmithWatermanGotohScoreWithParams(e.A, e.B, params)
            if math.Abs(got - e.BiopythonNormalised) > epsilon {
                t.Errorf("SmithWatermanGotohScoreWithParams(%q, %q, %+v) = %.12f; biopython_normalised = %.12f (delta %.2e, tol %g, biopython %s)",
                    e.A, e.B, params, got, e.BiopythonNormalised,
                    math.Abs(got - e.BiopythonNormalised), epsilon, c.BiopythonVersion)
            }
        })
    }
}
```

**Imports required in `swg_test.go`** (in addition to the analog's `testing` + `fuzzymatch`): `encoding/json`, `math`, `os`, `path/filepath`.

**Comparison anchor** — the Go test compares against `BiopythonNormalised` (NOT `BiopythonScore`) per CONTEXT.md §1: "Go test compares against the normalised value with zero in-Go normalisation logic". This means the script owns the normalisation reference; the Go test is a pure equality check within tolerance.

---

### §unique-files-4 — `regen-swg-cross-validation` Makefile target

Already covered above under "Pattern Assignments — Extended Files → Makefile". Repeating the locator for completeness:

- **Insertion point:** after `verify-license-headers` (Makefile line 194), before `release-check` (line 197).
- **`.PHONY` list update:** add `regen-swg-cross-validation` to lines 26–28.
- **Style anchor:** the `security` target (lines 171–176) is the closest analog for the "tolerant if tool not installed" pattern; reuse that idiom for the biopython prerequisite check.

---

## No Analog Found

Files genuinely without a Phase 2 analog (specified from scratch above):

| File | Reason |
|---|---|
| `testdata/cross-validation/swg/vectors.json` | First cross-validation corpus; not the same shape as Phase 2 golden files |
| `scripts/gen-swg-cross-validation.py` | First Python file in the repo |
| `TestSWG_CrossValidation` (in `swg_test.go`) | First test that consumes an external-reference fixture |
| `regen-swg-cross-validation` Makefile target | First developer-only generator target (`verify-determinism` is the closest stylistic kin) |

Phase 3 also exposes a new shape to the public API surface — the `SWGParams` value type and `NewSWGParams()` constructor (per CONTEXT.md §3). This is the first parameterised algorithm in the catalogue; no Phase 2 algorithm exposes a config struct. The shape is documented in CONTEXT.md §3 with the godoc template; the `api-ergonomics-reviewer` agent gates the final form (per CLAUDE.md "Workflow — Agent Gates" §5).

---

## Metadata

**Analog search scope:**
- Root package: 60+ `.go` files surveyed; six Phase 2 algorithm families (`levenshtein.go`, `damerau_full.go`, `damerau_osa.go`, `hamming.go`, `jaro.go`, `jarowinkler.go`) provide the canonical templates.
- `tests/bdd/features/` (6 feature files); `tests/bdd/steps/algorithms_steps.go` (413 lines).
- `testdata/golden/` (algorithms.json + 6 staging files); `testdata/fuzz/` (8 fuzzer corpora).
- `scripts/` (3 bash files); `Makefile` (212 lines, 19 targets).
- `examples/identifier-similarity/` (main.go + main_test.go).
- Project documentation: `llms.txt`, `docs/requirements.md` §7.1.8.

**Files scanned for analog selection:** ~30 files. Strong matches found for all 18 non-unique files (exact-template match for 16; role-match for `Makefile` and `docs/requirements.md`).

**Pattern extraction date:** 2026-05-14.
