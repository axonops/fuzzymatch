# Phase 7: Phonetic Algorithms - Pattern Map

**Mapped:** 2026-05-15
**Files analyzed:** 36 new/modified files (plus llms.txt/llms-full.txt incremental)
**Analogs found:** 35 / 36 (one new schema — `testdata/golden/phonetic-codes.json` — has no prior analog and is described from CONTEXT.md §7)

## File Classification

| New/Modified File | Role | Data Flow | Closest Analog | Match Quality |
|-------------------|------|-----------|----------------|---------------|
| `soundex.go` | algorithm (small/rule-table) | request-response (pure fn) | `token_jaccard.go` | role-match (small, single-call, no DP) |
| `double_metaphone.go` | algorithm (rule-heavy) | request-response (pure fn) | `ratcliff_obershelp.go` | exact (heavy rule logic, language-branch fan-out, fresh-transcription discipline) |
| `nysiis.go` | algorithm (small/rule-list) | request-response (pure fn) | `token_jaccard.go` / `qgram_jaccard.go` | role-match |
| `mra.go` | algorithm (medium/threshold-table) | request-response (pure fn) | `ratcliff_obershelp.go` | role-match (threshold-table at `var`; multi-public-surface) |
| `dispatch_soundex.go` | dispatch wiring | startup | `dispatch_ratcliff_obershelp.go` | exact |
| `dispatch_double_metaphone.go` | dispatch wiring | startup | `dispatch_ratcliff_obershelp.go` | exact |
| `dispatch_nysiis.go` | dispatch wiring | startup | `dispatch_ratcliff_obershelp.go` | exact |
| `dispatch_mra.go` | dispatch wiring | startup | `dispatch_ratcliff_obershelp.go` | exact |
| `soundex_test.go` | unit test (literature ref vectors) | request-response | `ratcliff_obershelp_test.go` | exact |
| `double_metaphone_test.go` | unit test (language-branch checklist) | request-response | `ratcliff_obershelp_test.go` | exact |
| `nysiis_test.go` | unit test (truncation gate) | request-response | `ratcliff_obershelp_test.go` | exact |
| `mra_test.go` | unit test (threshold-edge pairs, multi-surface) | request-response | `ratcliff_obershelp_test.go` | role-match (need extra coverage for the `(bool, int)` return) |
| `soundex_bench_test.go` | benchmark | request-response | `token_jaccard_bench_test.go` | exact (no pathological fixture) |
| `double_metaphone_bench_test.go` | benchmark | request-response | `token_jaccard_bench_test.go` | exact |
| `nysiis_bench_test.go` | benchmark | request-response | `token_jaccard_bench_test.go` | exact |
| `mra_bench_test.go` | benchmark | request-response | `token_jaccard_bench_test.go` | exact |
| `soundex_fuzz_test.go` | fuzz test | request-response | `token_jaccard_fuzz_test.go` | exact (ASCII + non-ASCII seed regimes) |
| `double_metaphone_fuzz_test.go` | fuzz test | request-response | `token_jaccard_fuzz_test.go` | exact |
| `nysiis_fuzz_test.go` | fuzz test | request-response | `token_jaccard_fuzz_test.go` | exact |
| `mra_fuzz_test.go` | fuzz test | request-response | `token_jaccard_fuzz_test.go` | exact |
| `scripts/gen-phonetic-cross-validation.py` | python tool (corpus generator) | batch | `scripts/gen-token-ratio-cross-validation.py` | exact (RapidFuzz pin → jellyfish dual-pin) |
| `phonetic_cross_validation_test.go` | Go corpus loader | batch | `token_ratio_cross_validation_test.go` | exact |
| `testdata/cross-validation/phonetic/vectors.json` | committed corpus JSON | data | `testdata/cross-validation/token-ratios/vectors.json` | role-match (new schema with `variant_divergence` flag) |
| `testdata/golden/phonetic-codes.json` | byte-stable code-vector golden | data | NEW SCHEMA | none (described fresh from CONTEXT.md §7) |
| `phonetic_codes_golden_test.go` | golden loader (string equality) | request-response | `algorithms_golden_test.go` | role-match (string equality instead of float-bit equality) |
| `testdata/golden/_staging/soundex.json` | staging-golden score entries | data | `testdata/golden/_staging/ratcliff_obershelp.json` | exact |
| `testdata/golden/_staging/double_metaphone.json` | staging-golden | data | `testdata/golden/_staging/ratcliff_obershelp.json` | exact |
| `testdata/golden/_staging/nysiis.json` | staging-golden | data | `testdata/golden/_staging/ratcliff_obershelp.json` | exact |
| `testdata/golden/_staging/mra.json` | staging-golden | data | `testdata/golden/_staging/ratcliff_obershelp.json` | exact |
| `Makefile` target `regen-phonetic-cross-validation` | build target | startup/batch | Makefile `regen-token-ratio-cross-validation` target | exact |
| `docs/cross-validation.md` extension | docs (additive section) | n/a | existing `docs/cross-validation.md` (token-ratios section) | exact |
| `tests/bdd/features/soundex.feature` | BDD feature | event-driven | `tests/bdd/features/ratcliff_obershelp.feature` | exact |
| `tests/bdd/features/double_metaphone.feature` | BDD feature | event-driven | `tests/bdd/features/ratcliff_obershelp.feature` | exact |
| `tests/bdd/features/nysiis.feature` | BDD feature | event-driven | `tests/bdd/features/ratcliff_obershelp.feature` | exact |
| `tests/bdd/features/mra.feature` | BDD feature | event-driven | `tests/bdd/features/ratcliff_obershelp.feature` | exact |
| `tests/bdd/features/monge_elkan_phonetic_inner.feature` | BDD feature (binary-inner composition) | event-driven | `tests/bdd/features/ratcliff_obershelp.feature` (structurally) + `tests/bdd/features/monge_elkan.feature` (subject) | role-match |
| `tests/bdd/steps/algorithms_steps.go` (append) | step registration | event-driven | existing `algorithms_steps.go` (per-Phase-2-6 append pattern) | exact |
| `monge_elkan.go` (`permittedMongeElkanInner` map mutation) | dispatch wiring (additive) | startup | existing map at `monge_elkan.go:294-314` | exact (additive — one entry per plan) |
| `monge_elkan_test.go` (`rejected` slice fixture shrink + 4 new binary-inner tests) | unit test (mutation) | request-response | existing `rejected` slice at `monge_elkan_test.go:313-323` | exact |
| `props_test.go` (Five-invariant blocks × 4) | property test (append) | request-response | `props_test.go:1391-1451` (RatcliffObershelp block) | exact |
| `example_test.go` (≥ 9 new `ExampleXxx`) | godoc example (append) | request-response | `example_test.go:179-194` (RatcliffObershelp examples) | exact |
| `llms.txt` / `llms-full.txt` (append per-plan) | docs (additive) | n/a | existing per-Phase-5/6 sync pattern | exact |
| `examples/identifier-similarity/main.go` (19→23 columns) | example (extension) | request-response | existing `main.go` Phase 6 extension to 19 columns | exact (additive — append 4 entries to `algorithms` slice) |
| `examples/identifier-similarity/main_test.go` (golden-stdout regen) | example test | request-response | existing `main_test.go` `want` constant | exact |
| `examples/phonetic-keys/main.go` | example (new program) | request-response | `examples/identifier-similarity/main.go` | role-match (different surface — encoded keys not scores) |
| `examples/phonetic-keys/main_test.go` | example test (new) | request-response | `examples/identifier-similarity/main_test.go` | role-match (string-table golden stdout) |

## Pattern Assignments

### `<algo>.go` family — header / Source-Origin Statement / file layout

**Analog:** `ratcliff_obershelp.go` (header block at lines 1-115 is the WR-01 LOCKED format)

**Apache header pattern** (lines 1-13):
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
```

**Source-Origin Statement pattern** (lines 67-77):
```go
// Source-origin discipline (per algorithm-licensing-standards):
//
//   - Primary source:        Ratcliff & Metzener 1988 (Dr. Dobb's Journal)
//   - Cross-validation:      Python difflib.SequenceMatcher (PSF licence,
//                            stdlib) — consulted ONLY for the
//                            find_longest_match tie-break contract
//                            (leftmost-in-`a` first, then leftmost-in-`b`
//                            among ties) and for cross-validation reference
//                            vectors. NOT for code copying.
//   - GPL/LGPL provenance:   none.
//   - Code copied verbatim:  none.
```

**Implementation discipline checklist** (lines 79-98) — must enumerate: D-2 (recursion choice), D-3 (inline vs reuse), NO `init()` table builds, NO map iteration on output (DET-03), NO transcendental floats (DET-06), NO goroutines, identity short-circuit BEFORE rune allocation.

**Phase 7 deltas — all four algorithms:**
- `soundex.go` Source-Origin Statement uses Knuth TAOCP Vol. 3 §6.4 as primary; `jellyfish==<pin>` (BSD-2) as cross-validation. Negative attribution list: `xrash/smetrics`, `tilotech/go-phonetics`, `UjjwalAyyangar/go-jellyfish`.
- `nysiis.go` Source-Origin Statement uses the LOCKED two-line form from CONTEXT.md §2 (Taft 1970 = algorithmic origin; Knuth TAOCP Vol. 3 §6.4 = canonical algorithm description); negative attribution `UjjwalAyyangar/go-jellyfish`.
- `mra.go` Source-Origin Statement uses NBS Tech Note 943 as primary; mentions `MRACompare (bool, int)` return shape is faithful to NBS-943 step 5/6 (spec line 691).
- **`double_metaphone.go`** Source-Origin Statement is EXTENDED with TWO additional lines per CONTEXT.md §3:
  - Rule-table provenance line: `// Rule table derived fresh from: Philips 2000 (C/C++ Users Journal, 18(6)) + the public-domain C reference (SWI-Prolog archive)`
  - Negative-attribution line: `// MIT-licensed Go ports NOT consulted: CalypsoSys/godoublemetaphone, deezer/double-metaphone-go, any other Go port`

**ASCII/rune separation:** phonetic algorithms operate on ASCII only (per CONTEXT.md §5 silent-skip discipline) — NO `XxxScoreRunes` companion functions. The package-level godoc warning paragraph from CONTEXT.md §5 is copied verbatim into each algorithm's package-level block.

---

### `dispatch_<algo>.go` files (4 new)

**Analog:** `dispatch_ratcliff_obershelp.go` (full file, 39 lines)

**Complete file pattern**:
```go
// (Apache header — 13 lines)

// dispatch_<algo>.go registers <Algo>Score into the dispatch table at
// package load time. This file MUST be the sole writer to
// dispatch[Algo<Algo>] (slot N — see algoid.go for the slot map).
//
// Only <Algo>Score is dispatched — the dispatch table maps AlgoID to
// (a, b string) float64. Companion surfaces (<Algo>Code, <Algo>Keys,
// MRACompare) are public but not dispatched.
//
// See algoid.go for the dispatch array declaration and its design
// rationale. The var _ = func() bool { ... }() idiom is the
// Phase-2-canonical form for package-level side effects without init()
// (per determinism-standards §13.5 and docs/requirements.md §5(12)).

package fuzzymatch

var _ = func() bool {
    dispatch[Algo<Algo>] = <Algo>Score
    return true
}()
```

**Phase 7 deltas:**
- Slot numbers are already reserved in `algoid.go` lines 156-175 (`AlgoSoundex` = slot 23, `AlgoDoubleMetaphone` = 24, `AlgoNYSIIS` = 25, `AlgoMRA` = 26 — confirm exact numeric indices via `algoid.go`'s `numAlgorithms` constant at planning time).
- `dispatch_mra.go` wires `MRAScore` (NOT `MRACompare` — the dispatch table is `(a, b string) float64`-valued; `MRACompare` is public but unwired).

---

### `<algo>_test.go` family

**Analog:** `ratcliff_obershelp_test.go` (lines 1-200 cover the canonical structure)

**File doc pattern** (lines 15-26):
```go
// <algo>_test.go pins the public-API contract of <algo>.go: identity,
// both-empty, one-empty, the canonical <primary-source> reference vectors
// (<list of named fixtures>), a numerical-regression pin OUTSIDE the
// cross-validation corpus, the <variant-choice> pin (Knuth/Census variant
// over SQL/MySQL for Soundex; original Taft-1970 over modified for
// NYSIIS), and the ASCII-only-with-non-ASCII-silent-skip pin.
//
// Stdlib `testing` only — no testify in root tests, per
// .claude/skills/go-coding-standards.

package fuzzymatch_test

import (
    "math"
    "testing"

    "github.com/axonops/fuzzymatch"
)
```

**Both-empty + one-empty pattern** (lines 65-96):
```go
func TestSoundex_BothEmpty(t *testing.T) {
    if got := fuzzymatch.SoundexCode(""); got != "" {
        t.Errorf("SoundexCode(\"\") = %q; want \"\"", got)
    }
    if got := fuzzymatch.SoundexScore("", ""); got != 1.0 {
        t.Errorf("SoundexScore(\"\", \"\") = %g; want 1.0", got)
    }
}

func TestSoundex_OneEmpty(t *testing.T) {
    // ... 0.0 in both directions
}
```

**Literature reference vector pattern** (lines 128-146 — Dr. Dobb's 1988 pair):
```go
func TestSoundexCode_KnuthReferenceVectors(t *testing.T) {
    tests := []struct {
        a    string
        want string
    }{
        // Knuth TAOCP Vol. 3 §6.4 canonical examples:
        {"Robert", "R163"},   // Knuth p. 393 first example
        {"Rupert", "R163"},   // Knuth p. 393 — same code as Robert
        {"Rubin",  "R150"},   // Knuth p. 393
        {"Tymczak", "T522"},  // Knuth/Census variant gate — SQL would yield "T520"
        {"Ashcraft", "A261"}, // H/W-handling gate (Pitfall 4)
        {"Ashcroft", "A261"}, // H/W-handling pair
    }
    for _, tt := range tests {
        t.Run(tt.a, func(t *testing.T) {
            if got := fuzzymatch.SoundexCode(tt.a); got != tt.want {
                t.Errorf("SoundexCode(%q) = %q; want %q", tt.a, got, tt.want)
            }
        })
    }
}
```

**Variant-gate pin pattern** (mirrors `TestRatcliffObershelp_PinnedDrDobbsValue` at lines 161-169):
```go
// TestSoundexCode_TymczakVariantGate is the LOAD-BEARING discriminator
// against the SQL/MySQL variant. Knuth/Census returns "T522"; SQL returns
// "T520". A regression that silently flips the variant choice would
// produce a different value here. Per CONTEXT.md §1 LOCKED.
func TestSoundexCode_TymczakVariantGate(t *testing.T) {
    const want = "T522"
    got := fuzzymatch.SoundexCode("Tymczak")
    if got != want {
        t.Errorf("SoundexCode(\"Tymczak\") = %q; want %q (Knuth/Census variant per CONTEXT.md §1)", got, want)
    }
}
```

**Phase 7 deltas:**
- Soundex tests: Tymczak (T522) variant gate + Ashcraft/Ashcroft (A261) H/W pair as load-bearing gates.
- Double Metaphone tests: 5 mandatory language-branch sub-tests in a `TestDoubleMetaphoneKeys_LanguageBranches` Scenario Outline-style table (Germanic Schmidt/Smith XMT-match + Slavic ≥1 + Romance Pacheco PXK + Greek Catherine=Katherine + Chinese ≥1).
- NYSIIS tests: Brown/Browne BRAN pair + Robert RABAD + truncation gate (`len(NYSIISCode("Catherine")) == 6` NOT 7).
- MRA tests: must cover BOTH `MRACode`, `MRACompare` (returning `(bool, int)`), AND `MRAScore`. Need a `TestMRACompare_ConsistencyPin` asserting `MRAScore(a,b) == 1.0 iff MRACompare(a,b).matched`. Threshold-edge pairs (similarity = threshold, similarity = threshold-1).

---

### `<algo>_bench_test.go` family

**Analog:** `token_jaccard_bench_test.go` (full file)

**Header pattern** (lines 15-39):
```go
// <algo>_bench_test.go runs allocation-aware benchmarks for <Algo>Score
// at multiple input sizes. b.ReportAllocs() on every benchmark gates
// allocation regressions in bench.txt via benchstat.
//
// Performance budget per .claude/skills/performance-standards:
//   - <Algo>: < 500 ns, 0 allocations (Soundex/NYSIIS); < 2 µs, ≤ 2
//     allocations (DoubleMetaphone); < 500 ns, 0 (MRACode), 2 (MRACompare)
//
// No pathological fixture: phonetic algorithms are O(n) on bounded input
// (typical name length < 50 chars); worst-case < 1µs (per CONTEXT.md
// §6-prior — Phase 6 DoS-vector format does NOT apply).
//
// `var sink` outside the loop + a `sink < 0` gate after the loop
// prevents compiler dead-code elimination (locked Phase 2 pattern).
```

**Benchmark function pattern** (lines 58-71):
```go
func BenchmarkSoundexCode_ASCII_Short(b *testing.B) {
    b.ReportAllocs()
    b.ResetTimer()
    var sink string
    for i := 0; i < b.N; i++ {
        sink = fuzzymatch.SoundexCode("Robert")
    }
    if sink == "" {
        b.Fatal("sink unexpectedly empty — compiler folded the benchmark away")
    }
}
```

**Phase 7 deltas:**
- Score benchmarks gate on `var sink float64; if sink < 0` (matches existing pattern).
- Code benchmarks gate on `var sink string; if sink == ""` (Phase 7 needs this because phonetic Code surfaces return strings).
- No Pathological fixtures (per CONTEXT.md §6-prior).
- Benchmark sizes: ASCII short (~10 chars typical name), ASCII medium (~30 chars), ASCII long (~50 chars). MRACompare benchmarks add a "length-difference >3 mismatch shortcut" fixture (asserts early-exit cost).

---

### `<algo>_fuzz_test.go` family

**Analog:** `token_jaccard_fuzz_test.go` (full file, 98 lines)

**Header + invariants block pattern** (lines 15-34):
```go
// <algo>_fuzz_test.go runs native Go fuzzing against the <Algo>Score
// (and <Algo>Code) public surfaces. Properties checked per input:
//
//   1. Never panics (implicit — any panic propagates as a fuzz crash).
//   2. Score never returns NaN.
//   3. Score never returns ±Inf.
//   4. Score returns a value in [0.0, 1.0].
//   5. <Algo>Code output character set: [A-Z0-9] for Soundex/NYSIIS/DM,
//      [A-Z0-9 ] for MRA.
//   6. Identity short-circuit holds — <Algo>Score(x, x) == 1.0 for any x.
//
// On-disk seed corpus: programmatic f.Add seeds cover ASCII-only inputs
// (the encoded regime) AND mixed ASCII+non-ASCII inputs (the silent-skip
// regime per CONTEXT.md §5).
```

**Seed pattern with two regimes** (analog at lines 50-73 expanded for Phase 7):
```go
func FuzzSoundex(f *testing.F) {
    for _, seed := range []struct{ a, b string }{
        // ASCII regime — literature reference vectors:
        {"Robert", "Rupert"},
        {"Tymczak", "Tymczak"},
        {"Ashcraft", "Ashcroft"},
        // Non-ASCII silent-skip regime (CONTEXT.md §5):
        {"Müller", "Muller"},
        {"Café", "Cafe"},
        {"中文", "abc"},
        {"🎉hello", "hello"},
        {"", ""},
        {"a", ""},
        // Invalid UTF-8:
        {"\xff\xfe", "abc"},
    } {
        f.Add(seed.a, seed.b)
    }
    f.Fuzz(func(t *testing.T, a, b string) {
        got := fuzzymatch.SoundexScore(a, b)
        // ... NaN/Inf/range invariants ...
        // Charset invariant on Code surface:
        code := fuzzymatch.SoundexCode(a)
        for _, c := range code {
            if !((c >= 'A' && c <= 'Z') || (c >= '0' && c <= '9')) {
                t.Fatalf("SoundexCode(%q) = %q; non-[A-Z0-9] char %q", a, code, c)
            }
        }
    })
}
```

**Phase 7 deltas:**
- Add a CHARSET invariant per CONTEXT.md §5 fuzz-invariant list — not present in `token_jaccard_fuzz_test.go`.
- MRA charset invariant accepts space (`MRACode` returns space-separated form per spec).
- DM fuzz needs to assert on BOTH primary AND secondary keys.

---

### `scripts/gen-phonetic-cross-validation.py`

**Analog:** `scripts/gen-token-ratio-cross-validation.py` (full file)

**Header + pin gate pattern** (lines 16-54, 124-134):
```python
"""scripts/gen-phonetic-cross-validation.py — jellyfish dual-pin cross-
validation corpus generator for the four Phase-7 phonetic algorithms.

Regenerates testdata/cross-validation/phonetic/vectors.json by running
the pinned jellyfish (Soundex/NYSIIS/MRA) AND the pinned `metaphone`
PyPI package (Double Metaphone) against a fixed CASES list. Per RESEARCH.md
key finding 1, jellyfish 1.x deprecated double_metaphone; the separate
`metaphone` package (Andrew Collins port) is the canonical Python DM
reference.

CRITICAL — JELLYFISH_VERSION + METAPHONE_VERSION pin gates:
    JELLYFISH_VERSION = "<pinned>"  asserted via jellyfish.__version__
    METAPHONE_VERSION = "<pinned>"  asserted via pip show parsing
    (Metaphone package doesn't expose __version__).
"""

JELLYFISH_VERSION = "1.2.1"
METAPHONE_VERSION = "0.6"

import jellyfish  # noqa: E402
assert jellyfish.__version__ == JELLYFISH_VERSION, (
    f"jellyfish version mismatch: installed {jellyfish.__version__}, "
    f"script pinned to {JELLYFISH_VERSION} — "
    f"run: python3 -m pip install --user jellyfish=={JELLYFISH_VERSION}"
)
```

**Variant-divergence flag pattern** (NEW for Phase 7 — not in analog):
```python
def soundex_entry(a: str) -> dict:
    code = jellyfish.soundex(a)
    # Detect SQL/MySQL vs Knuth/Census divergence.
    # Hand-curated table of known-divergent inputs:
    divergent = a in {"Tymczak"}  # ... extend per CONTEXT.md §1
    entry = {"input": a, "code": code}
    if divergent:
        entry["variant_divergence"] = True
        entry["divergent_jellyfish_value"] = code
        entry["code"] = KNUTH_EXPECTED[a]  # Knuth-expected value
    return entry
```

**Phase 7 deltas:**
- Two version pins instead of one (jellyfish + metaphone).
- Per-entry `variant_divergence: true` + `divergent_jellyfish_value` fields (CONTEXT.md §1).
- `_metadata` block additionally carries `metaphone_version` field beyond the existing `rapidfuzz_version` analog.
- Output path `testdata/cross-validation/phonetic/vectors.json` (new directory `phonetic/` created in plan 07-01).
- Per CONTEXT.md §1 corpus-size table: Soundex 15 entries, NYSIIS 20, MRA 20, Double Metaphone 40 covering all 5 language branches.

---

### `phonetic_cross_validation_test.go`

**Analog:** `token_ratio_cross_validation_test.go` (full file, ~210 lines)

**Type declarations** (lines 76-104):
```go
type phoneticEntry struct {
    Algorithm                string `json:"algorithm"`
    Input                    string `json:"input"`
    Code                     string `json:"code,omitempty"`        // single-key (Soundex, NYSIIS, MRA)
    Primary                  string `json:"primary,omitempty"`     // DM
    Secondary                string `json:"secondary,omitempty"`   // DM
    VariantDivergence        bool   `json:"variant_divergence,omitempty"`
    DivergentJellyfishValue  string `json:"divergent_jellyfish_value,omitempty"`
}

type phoneticMetadata struct {
    JellyfishVersion string `json:"jellyfish_version"`
    MetaphoneVersion string `json:"metaphone_version"`
    PythonVersion    string `json:"python_version"`
    RegeneratedAt    string `json:"regenerated_at"`
}

type phoneticCorpus struct {
    Version  int               `json:"version"`
    Metadata phoneticMetadata  `json:"_metadata"`
    Entries  []phoneticEntry   `json:"entries"`
}
```

**Per-entry assertion loop pattern** (lines 157-209 — adapted for string equality):
```go
for _, e := range c.Entries {
    e := e
    t.Run(e.Algorithm+"/"+e.Input, func(t *testing.T) {
        switch e.Algorithm {
        case "Soundex":
            got := fuzzymatch.SoundexCode(e.Input)
            // When variant_divergence=true, assert against the Knuth-
            // expected value (entry.Code) NOT the jellyfish value
            // (entry.DivergentJellyfishValue). This is the LOCKED
            // mechanism per CONTEXT.md §1.
            if got != e.Code {
                t.Errorf("SoundexCode(%q) = %q; want %q (jellyfish=%q, variant_divergence=%v, jellyfish=%s)",
                    e.Input, got, e.Code, e.DivergentJellyfishValue, e.VariantDivergence,
                    c.Metadata.JellyfishVersion)
            }
        case "DoubleMetaphone":
            primary, secondary := fuzzymatch.DoubleMetaphoneKeys(e.Input)
            if primary != e.Primary || secondary != e.Secondary {
                t.Errorf("DoubleMetaphoneKeys(%q) = (%q, %q); want (%q, %q) (metaphone=%s)",
                    e.Input, primary, secondary, e.Primary, e.Secondary,
                    c.Metadata.MetaphoneVersion)
            }
        // ... NYSIIS, MRA ...
        }
    })
}
```

**Phase 7 deltas:**
- String equality (`!=`) instead of float-tolerance comparison (`math.Abs(got - want) > eps`).
- `variant_divergence` flag handled inside the assertion — Go test asserts against the recorded `entry.Code` (the Knuth-expected truncated value), with the jellyfish value preserved purely for failure messages.
- Dual version-pin checks in the preamble (jellyfish AND metaphone), not just one.
- Single file with 4 effective sub-tests (one per algorithm) addressed via the `switch e.Algorithm` dispatch.

---

### `testdata/golden/phonetic-codes.json`

**Analog:** NEW SCHEMA (no prior analog).

**Schema from CONTEXT.md §7:**
```json
{
  "_metadata": {
    "purpose": "Cross-platform byte-stable phonetic code determinism gate",
    "regenerated_at": "<ISO>"
  },
  "entries": [
    {"algorithm": "Soundex",          "input": "Robert",  "code": "R163"},
    {"algorithm": "Soundex",          "input": "Tymczak", "code": "T522"},
    {"algorithm": "DoubleMetaphone",  "input": "Schmidt", "primary": "XMT", "secondary": "SMT"},
    {"algorithm": "NYSIIS",           "input": "Brown",   "code": "BRAN"},
    {"algorithm": "MRA",              "input": "Robert",  "code": "RBRT"}
  ]
}
```

Per algorithm: 8-12 entries from literature reference vectors + Phase 7 distinctive cases (Tymczak Soundex / Schmidt-Smith DM pair / Brown-Browne NYSIIS / threshold-edge MRA).

**Phase 7 delta:** plan 07-01 creates this file with Soundex section; plans 07-02..07-04 extend (append entries) for their algorithms.

---

### `phonetic_codes_golden_test.go`

**Analog:** `algorithms_golden_test.go` (lines 47-100 for the JSON loader pattern; lines 80-140 for `assertGoldenStaging` helper)

**Type declarations + loader pattern (adapt for string codes):**
```go
type phoneticGoldenEntry struct {
    Algorithm string `json:"algorithm"`
    Input     string `json:"input"`
    Code      string `json:"code,omitempty"`
    Primary   string `json:"primary,omitempty"`
    Secondary string `json:"secondary,omitempty"`
}

type phoneticGoldenFile struct {
    Metadata map[string]any         `json:"_metadata"`
    Entries  []phoneticGoldenEntry  `json:"entries"`
}

func TestPhoneticCodesGolden(t *testing.T) {
    path := filepath.Join("testdata", "golden", "phonetic-codes.json")
    raw, err := os.ReadFile(path)
    // ... parse ...
    for _, e := range f.Entries {
        e := e
        t.Run(e.Algorithm+"/"+e.Input, func(t *testing.T) {
            switch e.Algorithm {
            case "Soundex":
                got := fuzzymatch.SoundexCode(e.Input)
                if got != e.Code {
                    t.Errorf("SoundexCode(%q) = %q; want %q (determinism gate)", e.Input, got, e.Code)
                }
            // ... DoubleMetaphone, NYSIIS, MRA ...
            }
        })
    }
}
```

**Phase 7 deltas:**
- Asserts STRING equality (not the float-bit-pattern equality used by `algorithms_golden_test.go`).
- DM entries assert on BOTH primary and secondary keys.
- Separate from `algorithms_golden_test.go` per CONTEXT.md §7 (and §6-prior recommendation in CONTEXT.md).

---

### `testdata/golden/_staging/<algo>.json` (×4)

**Analog:** `testdata/golden/_staging/ratcliff_obershelp.json`

**Schema pattern** (first 40 lines):
```json
{
  "version": 1,
  "entries": [
    {
      "name": "Soundex_both_empty",
      "algorithm": "Soundex",
      "a": "",
      "b": "",
      "expected_score": 1
    },
    {
      "name": "Soundex_tymczak_self",
      "algorithm": "Soundex",
      "a": "Tymczak",
      "b": "Tymczak",
      "expected_score": 1
    },
    {
      "name": "Soundex_robert_rupert_match",
      "algorithm": "Soundex",
      "a": "Robert",
      "b": "Rupert",
      "expected_score": 1
    },
    ...
  ]
}
```

**Phase 7 delta:** 8-12 entries per algorithm; entries sorted alphabetically by `name`; canonical-marshal (encoding/json MarshalIndent with two-space indent, single trailing LF) via the `assertGoldenStaging` helper. Merged into `testdata/golden/algorithms.json` in plan 07-05 (finalisation) via the existing `TestGolden_Algorithms_Merge` pattern.

---

### `Makefile` target `regen-phonetic-cross-validation`

**Analog:** Makefile lines 241-246 (`regen-token-ratio-cross-validation` target)

**Target pattern**:
```makefile
# Regenerate testdata/cross-validation/phonetic/vectors.json. Requires
#   python3 -m pip install --user jellyfish==<pin> metaphone==<pin>
regen-phonetic-cross-validation:
	@if ! command -v python3 >/dev/null 2>&1; then \
	  echo "python3 not found; install Python 3.7+ and run:"; \
	  echo "  python3 -m pip install --user jellyfish==<pin> metaphone==<pin>"; \
	  exit 1; \
	fi
	python3 scripts/gen-phonetic-cross-validation.py
```

**Phase 7 delta:** install hint contains TWO pip packages instead of one. Target name added to the `.PHONY` list at Makefile lines 28-29.

---

### `docs/cross-validation.md` extension

**Analog:** existing `docs/cross-validation.md` (the token-ratios section already documented per Phase 6 plan 06-01)

**Pattern:** append a new "Phonetic cross-validation" section describing:
1. Why dual-pin (jellyfish + metaphone — per RESEARCH.md key finding 1).
2. The `variant_divergence` flag mechanism (Soundex Knuth/SQL split; NYSIIS Knuth/modified split).
3. The regenerate command + pip install hint.
4. The Go-side test entry point: `TestPhonetic_CrossValidation`.

---

### `tests/bdd/features/<algo>.feature` (×4 + 1)

**Analog:** `tests/bdd/features/ratcliff_obershelp.feature` (full file)

**Feature header pattern** (lines 1-21):
```gherkin
# Primary source: <citation>
# Cross-validation: jellyfish==<pin> (BSD-2) — reference vectors only,
#                   no code copying.
# Surface: <Algo>Code returns the encoded form; <Algo>Score returns
#          binary 0.0/1.0 from code equality. <Algo>Score is the
#          dispatched surface.

Feature: <Algo> phonetic encoding (Knuth/Census variant / Taft-1970 /
         Philips 2000 / NBS-943)
  <one-paragraph description>

  Scenario Outline: literature reference vectors
    When I compute the <Algo> code of "<input>"
    Then the code should be "<expected>"

    Examples:
      | input    | expected |
      | Robert   | R163     |
      | Tymczak  | T522     |
      | Ashcraft | A261     |

  Scenario: identical strings score 1.0
    When I compute the <Algo> score between "user" and "user"
    Then the score should be exactly 1

  Scenario: both-empty strings score 1.0
    When I compute the <Algo> score between "" and ""
    Then the score should be exactly 1

  Scenario: variant gate (<variant-specific>)
    When I compute the <Algo> code of "<gate-input>"
    Then the code should be "<gate-expected>"
```

**Phase 7 deltas:**
- Each algorithm needs at least 4 scenarios (per CONTEXT.md §6-prior — BDD scenarios per algorithm).
- DoubleMetaphone feature needs ≥ 6 scenarios (one per language branch per CONTEXT.md §3 mandatory checklist).
- MRA feature includes a `When I compare with MRA "<a>" and "<b>"` step bringing the `(bool, int)` return into the scenario state.
- Soundex feature includes Tymczak (T522) variant gate scenario.
- NYSIIS feature includes a truncation gate scenario (`len(NYSIISCode("Catherine")) == 6`).
- `monge_elkan_phonetic_inner.feature` (plan 07-04 or finalisation) covers the binary-inner-composition fixture from CONTEXT.md §4: `MongeElkanScore("alpha beta", "alpha gamma", AlgoSoundex, opts) == 0.5`.

---

### `tests/bdd/steps/algorithms_steps.go` (append)

**Analog:** existing `algorithms_steps.go` lines 53-66 (per-algorithm step function pattern) + lines 991-1100 (`InitializeScenario` registration pattern)

**Step function pattern (one per algorithm step verb)**:
```go
// iComputeTheSoundexCodeOf computes SoundexCode(s) and stores the result
// in lastCode.
func (ctx *AlgorithmContext) iComputeTheSoundexCodeOf(s string) error {
    ctx.lastCode = fuzzymatch.SoundexCode(s)
    return nil
}

// iComputeTheSoundexScoreBetween computes SoundexScore(a, b) and stores
// the result in lastScore.
func (ctx *AlgorithmContext) iComputeTheSoundexScoreBetween(a, b string) error {
    ctx.lastScore = fuzzymatch.SoundexScore(a, b)
    return nil
}
```

**State extension** (lines 46-51 — add new fields):
```go
type AlgorithmContext struct {
    lastScore    float64
    lastScore2   float64
    lastDistance int
    lastPanicMsg string
    lastCode     string // populated by "I compute the <Algo> code of" steps (plan 07-01+)
    lastDMPrimary, lastDMSecondary string // populated by DM steps (plan 07-02+)
    lastMRAMatched bool // populated by MRACompare steps (plan 07-04+)
    lastMRASim     int  // populated by MRACompare steps (plan 07-04+)
}
```

**Registration pattern** (lines 994-1014 — append at end of `InitializeScenario`):
```go
// Soundex step definitions (plan 07-01).
ctx.Step(
    `^I compute the Soundex code of "([^"]*)"$`,
    a.iComputeTheSoundexCodeOf,
)
ctx.Step(
    `^I compute the Soundex score between "([^"]*)" and "([^"]*)"$`,
    a.iComputeTheSoundexScoreBetween,
)
ctx.Step(
    `^the code should be "([^"]*)"$`,
    a.theCodeShouldBe,
)
```

**Phase 7 deltas:**
- New `lastCode`, `lastDMPrimary`, `lastDMSecondary`, `lastMRAMatched`, `lastMRASim` state fields.
- New shared step `^the code should be "([^"]*)"$` (covers Soundex/NYSIIS/MRA single-key Code surfaces).
- DM-specific step `^the keys should be "([^"]*)" and "([^"]*)"$` (handles two-key output).
- MRA-specific step `^the MRA similarity should be (\d+)$` (the integer counter from `MRACompare`).
- One `step.Step` block appended per algorithm per plan.

---

### `monge_elkan.go` `permittedMongeElkanInner` map (additive mutation)

**Analog:** existing map at `monge_elkan.go:294-314`

**Map declaration pattern** (lines 294-314):
```go
var permittedMongeElkanInner = map[AlgoID]bool{
    // Character tier (9):
    AlgoLevenshtein:            true,
    AlgoDamerauLevenshteinOSA:  true,
    // ... 6 more ...
    AlgoLCSStr:                 true,
    // Q-gram tier (4):
    AlgoQGramJaccard: true,
    AlgoSorensenDice: true,
    AlgoCosine:       true,
    AlgoTversky:      true,
    // Gestalt tier (1) — OQ-4 RESOLUTION LOCKED 2026-05-15:
    AlgoRatcliffObershelp: true,
}
```

**Phase 7 deltas — each plan adds its own entry plus a comment tier line:**
- Plan 07-01: append `// Phonetic tier (Phase 7):` and `AlgoSoundex: true, // Russell 1918 / Knuth TAOCP §6.4` (15 total).
- Plan 07-02: append `AlgoDoubleMetaphone: true, // Philips 2000` (16 total).
- Plan 07-03: append `AlgoNYSIIS: true, // Taft 1970 / Knuth TAOCP §6.4` (17 total).
- Plan 07-04: append `AlgoMRA: true, // Moore 1977 / NBS Tech Note 943` (18 total).
- Each plan also updates the map's leading comment line count: `14 entries (9 character-tier ...)` → `15 entries (... + 1 phonetic-tier)`.

---

### `monge_elkan_test.go` panic-fixture (additive shrink)

**Analog:** existing `rejected` slice at `monge_elkan_test.go:313-323`

**Current state** (9 rejected entries):
```go
rejected := []fuzzymatch.AlgoID{
    fuzzymatch.AlgoMongeElkan,
    fuzzymatch.AlgoTokenSortRatio,
    fuzzymatch.AlgoTokenSetRatio,
    fuzzymatch.AlgoPartialRatio,
    fuzzymatch.AlgoTokenJaccard,
    fuzzymatch.AlgoSoundex,         // Phase 7 reserved
    fuzzymatch.AlgoDoubleMetaphone, // Phase 7 reserved
    fuzzymatch.AlgoNYSIIS,          // Phase 7 reserved
    fuzzymatch.AlgoMRA,             // Phase 7 reserved
}
```

**Phase 7 deltas — each plan REMOVES its own AlgoID from `rejected`:**
- Plan 07-01: remove `AlgoSoundex` (8 rejected); same plan adds `AlgoSoundex` to `permittedSanity` slice at lines 370-385.
- Plan 07-02: remove `AlgoDoubleMetaphone` (7 rejected); add to `permittedSanity`.
- Plan 07-03: remove `AlgoNYSIIS` (6 rejected); add to `permittedSanity`.
- Plan 07-04: remove `AlgoMRA` (5 rejected); add to `permittedSanity`.

**Note on numbering:** the CONTEXT.md §4 table claims a 14→13→12→11→10 fixture-count shrink. RESEARCH.md Wave 0 line 608 corrects this to a 9→8→7→6→5 actual count (matching the actual current state in `monge_elkan_test.go:313-323` — 9 rejected entries). Plans should follow the RESEARCH.md numbering (the actual code state).

---

### `monge_elkan_test.go` — 4 new binary-inner ME composition tests

**Analog:** `monge_elkan_test.go` existing per-inner permitted tests (lines 386-389 for the simple sanity-loop pattern) + RESEARCH.md §6-prior fixture `MongeElkanScore("alpha beta", "alpha gamma", AlgoSoundex, opts) == 0.5`

**Test function pattern (one per phonetic AlgoID):**
```go
// TestMongeElkanScore_BinaryInner_Soundex pins the binary-inner-composition
// behaviour: one of two tokens matches phonetically, the other doesn't,
// so the asymmetric ME score is exactly 0.5 (per-token max average over
// the left side). Per CONTEXT.md §4 LOCKED.
func TestMongeElkanScore_BinaryInner_Soundex(t *testing.T) {
    opts := fuzzymatch.DefaultNormalisationOptions()
    cases := []struct {
        name    string
        a, b    string
        want    float64
    }{
        {"one_matches", "alpha beta", "alpha gamma", 0.5},
        {"both_match",  "alpha beta", "alpha beta",  1.0},
        {"neither",     "alpha",      "gamma",       0.0},
    }
    for _, tt := range cases {
        t.Run(tt.name, func(t *testing.T) {
            got := fuzzymatch.MongeElkanScore(tt.a, tt.b, fuzzymatch.AlgoSoundex, opts)
            if got != tt.want {
                t.Errorf("MongeElkanScore(%q, %q, AlgoSoundex, opts) = %g; want %g", tt.a, tt.b, got, tt.want)
            }
        })
    }
}
```

**Phase 7 deltas:**
- Four near-identical functions (`_Soundex`, `_DoubleMetaphone`, `_NYSIIS`, `_MRA`).
- Each lands in the same plan that wires the underlying algorithm dispatch (plans 07-01..07-04 respectively).
- Inputs chosen so the inner returns exactly 0.0 or 1.0 — keeps the test independent of the specific phonetic encoding (`alpha`/`gamma` are unrelated; `alpha`/`alpha` is identical).

---

### `props_test.go` — Five-invariant blocks per algorithm

**Analog:** `props_test.go:1389-1451` (the `RatcliffObershelpScore_*` block — 5 invariants × ~12 lines each)

**Five-invariant pattern per algorithm** (the load-bearing five for Phase 7 per RESEARCH.md):

```go
// TestProp_SoundexScore_RangeBounds asserts the score is in [0.0, 1.0]
// for any pair of strings. DET-04 range-bounds invariant.
func TestProp_SoundexScore_RangeBounds(t *testing.T) {
    f := func(a, b string) bool {
        s := fuzzymatch.SoundexScore(a, b)
        return s >= 0.0 && s <= 1.0
    }
    if err := quick.Check(f, nil); err != nil {
        t.Errorf("SoundexScore out of [0,1]: %v", err)
    }
}

// TestProp_SoundexScore_Identity asserts Score(x, x) == 1.0 for non-empty x.
// TestProp_SoundexScore_Symmetric asserts Score(a, b) == Score(b, a).
// TestProp_SoundexScore_NoNaN asserts the score is never NaN.
// TestProp_SoundexScore_NoInf asserts the score never returns ±Inf.
// (NoNegativeZero is OPTIONAL — phonetic scores are integer 0 / 1 — drop)
```

**Phase 7 deltas:**
- Phonetic algorithms ARE symmetric (unlike RO) — include the `_Symmetric` invariant (DROPPED for RO, KEPT for phonetics).
- `NoNegativeZero` is OPTIONAL — phonetic scores compute 0.0 / 1.0 directly as IEEE-754 positives, not from a division; the invariant is trivially satisfied. Include it for consistency with the Phase 2-6 pattern.
- Append to `props_test.go` at the file's tail in plan order: 07-01 Soundex block, 07-02 DM, 07-03 NYSIIS, 07-04 MRA.

---

### `example_test.go` — 9+ new `ExampleXxx` entries

**Analog:** `example_test.go:179-194` (RatcliffObershelp examples — 16 lines for two functions)

**Example function pattern**:
```go
// ExampleSoundexCode demonstrates the Soundex code encoding on Knuth's
// canonical "Robert" reference vector. The Knuth/Census variant returns
// "R163" (4-char fixed-length code: first letter + 3 digit groups, vowel-
// separated; H and W are skipped without resetting the lastGroup state).
func ExampleSoundexCode() {
    fmt.Println(fuzzymatch.SoundexCode("Robert"))
    // Output:
    // R163
}

// ExampleSoundexScore demonstrates the binary similarity: "Robert" and
// "Rupert" share the Soundex code R163, so the score is 1.0.
func ExampleSoundexScore() {
    fmt.Printf("%.1f\n", fuzzymatch.SoundexScore("Robert", "Rupert"))
    // Output:
    // 1.0
}
```

**Phase 7 deltas — 9 examples minimum:**
- `ExampleSoundexCode`, `ExampleSoundexScore`
- `ExampleDoubleMetaphoneKeys`, `ExampleDoubleMetaphoneScore`
- `ExampleNYSIISCode`, `ExampleNYSIISScore`
- `ExampleMRACode`, `ExampleMRACompare`, `ExampleMRAScore`

`ExampleMRACompare` demonstrates the `(bool, int)` return shape:
```go
func ExampleMRACompare() {
    matched, sim := fuzzymatch.MRACompare("Byrne", "Boern")
    fmt.Printf("matched=%v sim=%d\n", matched, sim)
    // Output:
    // matched=true sim=5
}
```

Each `ExampleXxx` lands in the same plan that ships the corresponding public function.

---

### `llms.txt` + `llms-full.txt` (per-plan sync)

**Analog:** existing `llms.txt` lines 107-109 (RatcliffObershelp entry) + `llms-full.txt` lines 87-90 (phonetic AlgoID table — already present)

**Pattern:** every plan that adds a public function adds:
1. A one-line entry in `llms.txt` under the appropriate section heading.
2. A full multi-line entry in `llms-full.txt` with godoc verbatim.

**Phase 7 deltas (entries to add):**
- Plan 07-01: `### Soundex (Russell 1918 / Knuth TAOCP §6.4)` heading + `- func SoundexCode(s string) string` + `- func SoundexScore(a, b string) float64`.
- Plan 07-02: `### Double Metaphone (Philips 2000)` + `- func DoubleMetaphoneKeys(s string) (primary, secondary string)` + `- func DoubleMetaphoneScore(a, b string) float64`.
- Plan 07-03: `### NYSIIS (Taft 1970 / Knuth TAOCP §6.4)` + `- func NYSIISCode(s string) string` + `- func NYSIISScore(a, b string) float64`.
- Plan 07-04: `### MRA (Moore 1977 / NBS Tech Note 943)` + `- func MRACode(s string) string` + `- func MRACompare(a, b string) (matched bool, simScore int)` + `- func MRAScore(a, b string) float64`.

**Discipline:** these sync ops happen in the SAME COMMIT as the function-add (per CONTEXT.md §6-prior). The `ai_friendly_test.go` test asserts `go/ast` symbol enumeration matches `llms.txt`, so a missed sync fails CI.

---

### `examples/identifier-similarity/main.go` (19→23 columns)

**Analog:** existing `main.go` (full file) — Phase 6 extension from 14→19 columns is the template

**Mutation pattern** (append to `algorithms` slice at lines 110-130):
```go
// Phase 7 phonetic tier (binary 0/1 scores):
{"Soundex",   fuzzymatch.SoundexScore},
{"DblMetaph", fuzzymatch.DoubleMetaphoneScore},
{"NYSIIS",    fuzzymatch.NYSIISScore},
{"MRA",       fuzzymatch.MRAScore},
```

**Phase 7 deltas:**
- Update file-doc text "all nineteen Phase 2 + 3 + 4 + 5 + 6 character-based, gestalt, q-gram, and token-based algorithms" → "all twenty-three Phase 2 + 3 + 4 + 5 + 6 + 7 character-based, gestalt, q-gram, token-based, and phonetic algorithms".
- The `algoWidth = 13` constant stays unchanged — `DblMetaph` (9 chars) is the longest new label, fits within 13.
- Lands in plan 07-05 (finalisation) per CONTEXT.md §8.

---

### `examples/identifier-similarity/main_test.go` (golden-stdout `want` regen)

**Analog:** existing `main_test.go` lines 42-51 (the `want` const)

**Phase 7 delta:** regenerate the `want` constant by running `go run ./examples/identifier-similarity/` and pasting output. Test structure unchanged. Lands in plan 07-05.

---

### `examples/phonetic-keys/main.go` (NEW program)

**Analog:** `examples/identifier-similarity/main.go` (full file — layout, table-rendering, godoc shape) — RE-USE the existing pattern, NOT the analog's specific surface

**Pattern:**
```go
// (Apache header — 13 lines)

// Package main demonstrates the four Phase 7 phonetic encoding functions
// (SoundexCode, DoubleMetaphoneKeys, NYSIISCode, MRACode + MRACompare)
// on a curated set of English-language surname samples.

package main

import (
    "fmt"
    "github.com/axonops/fuzzymatch"
)

var names = []string{"Robert", "Rupert", "Tymczak", "Schmidt", "Smith",
    "Catherine", "Katherine", "Brown", "Browne", "Pacheco", "Byrne", "Boern"}

func main() {
    // Header: Name | Soundex | DM-primary | DM-secondary | NYSIIS | MRA
    fmt.Printf("%-12s %-8s %-8s %-8s %-8s %-8s\n",
        "Name", "Soundex", "DM-pri", "DM-sec", "NYSIIS", "MRA")
    fmt.Println(strings.Repeat("-", 60))
    for _, n := range names {
        primary, secondary := fuzzymatch.DoubleMetaphoneKeys(n)
        fmt.Printf("%-12s %-8s %-8s %-8s %-8s %-8s\n",
            n,
            fuzzymatch.SoundexCode(n),
            primary, secondary,
            fuzzymatch.NYSIISCode(n),
            fuzzymatch.MRACode(n))
    }
}
```

**Phase 7 deltas:**
- Output is STRING-valued (encoded keys), NOT float-valued — the `algorithms` struct slice from the analog becomes a fixed column list.
- DM gets TWO columns (primary + secondary keys) — wider table than identifier-similarity per-pair single-score cells.
- Lands in plan 07-05 (finalisation) per CONTEXT.md §8.

---

### `examples/phonetic-keys/main_test.go` (NEW)

**Analog:** `examples/identifier-similarity/main_test.go` (full file — `want` constant pattern + os.Pipe stdout capture)

**Phase 7 deltas:**
- The `want` constant carries the string-valued table — no float-formatting precision issues.
- Same `TestExample_Output` shape + same `TestExample_ColumnWidths` shape.

## Shared Patterns

### Apache 2.0 file header
**Source:** every `.go` file in the repo (e.g. `ratcliff_obershelp.go:1-13`).
**Apply to:** every new `.go` file in Phase 7.
```go
// Copyright 2026 AxonOps Limited
//
// Licensed under the Apache License, Version 2.0 (the "License");
// ...
// limitations under the License.
```

### Source-Origin Statement (algorithm-licensing-standards)
**Source:** `ratcliff_obershelp.go:67-77` (WR-01 LOCKED format from Phase 4).
**Apply to:** all four `<algo>.go` files. Phase 7 EXTENDS with rule-table provenance line + MIT-Go-port negative attribution for `double_metaphone.go` only (per CONTEXT.md §3).

### Non-ASCII silent-skip godoc warning
**Source:** CONTEXT.md §5 canonical text block.
**Apply to:** the package-level godoc block of all four `<algo>.go` files (Soundex, Double Metaphone, NYSIIS, MRA).
```text
// Non-ASCII input handling: this algorithm operates on ASCII letters
// [A-Za-z] only. Non-ASCII runes (accented characters, emoji,
// combining marks) are dropped silently before encoding. For
// Unicode-aware similarity on non-ASCII input, compose with
// Normalise + diacritic stripping before calling this function, or
// use a character-based algorithm (e.g. Levenshtein, Jaro-Winkler).
```

### Dispatch wiring via `var _ = func() bool{...}()`
**Source:** `dispatch_ratcliff_obershelp.go:35-38`.
**Apply to:** all four `dispatch_<algo>.go` files.
```go
var _ = func() bool {
    dispatch[Algo<Algo>] = <Algo>Score
    return true
}()
```

### Identity short-circuit BEFORE any allocation
**Source:** `ratcliff_obershelp.go:144-146` + `ratcliff_obershelp.go:170-172` (IN-04 closure).
**Apply to:** all four `<Algo>Score` functions.
```go
if a == b {
    return 1.0 // identity short-circuit (covers both-empty too)
}
```

### Stack-allocated result buffer for short codes
**Source:** RESEARCH.md §6.1 sketch (`var result [4]byte` for Soundex).
**Apply to:** `soundex.go` (`[4]byte`), `nysiis.go` (`[6]byte`), MRA encoded-key path (`[6]byte`). DM uses two `strings.Builder` (per RESEARCH.md §6.2 — primary+secondary key result allocations are the only allocs).

### `var sink` benchmark dead-code-elimination guard
**Source:** `token_jaccard_bench_test.go:62-71`.
**Apply to:** all four `_bench_test.go` files. Adapt the sink type per surface (`float64` for Score, `string` for Code, `bool` for `MRACompare`).

### Fuzz invariants (no panic / no NaN / no Inf / range bounds / identity)
**Source:** `token_jaccard_fuzz_test.go:75-96`.
**Apply to:** all four `_fuzz_test.go` files. Phase 7 EXTENDS with a charset invariant on the Code surface (`[A-Z0-9]` or `[A-Z0-9 ]` for MRA) — not present in the analog.

### Variant version pin + `pip install` hint in script + Makefile
**Source:** `scripts/gen-token-ratio-cross-validation.py:124-134` + `Makefile:241-246`.
**Apply to:** `scripts/gen-phonetic-cross-validation.py` + Makefile new target. Phase 7 EXTENDS with a SECOND pin (metaphone package) per RESEARCH.md key finding 1.

### Per-plan llms.txt + llms-full.txt sync
**Source:** Phase 5+ discipline; existing `llms.txt:107-109` (RatcliffObershelp) + `llms-full.txt:87-90` (phonetic AlgoID table).
**Apply to:** every Phase 7 plan that adds a public function. The `ai_friendly_test.go` test enforces this via `go/ast`.

## No Analog Found

| File | Role | Data Flow | Reason |
|------|------|-----------|--------|
| `testdata/golden/phonetic-codes.json` | byte-stable code-vector golden | data | New schema — phonetic algorithms are the first to need string-valued (not float-valued) cross-platform determinism gates. Schema described fresh from CONTEXT.md §7 (which provides a complete example block). The loader `phonetic_codes_golden_test.go` adapts the float-valued `algorithms_golden_test.go` to assert string equality instead. |

## Metadata

**Analog search scope:** repo root (algorithm `.go` files), `scripts/`, `testdata/golden/_staging/`, `testdata/cross-validation/`, `tests/bdd/features/`, `tests/bdd/steps/`, `examples/`.
**Files scanned:** ~60 files; 35 unique analogs identified.
**Pattern extraction date:** 2026-05-15.
