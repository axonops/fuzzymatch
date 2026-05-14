# Phase 2: Core Character Algorithms (six) — Pattern Map

**Mapped:** 2026-05-14
**Files analyzed:** 34 new files + 1 modified file (export_test.go)
**Analogs found:** 34 / 34 (all files have strong Phase 1 analogs; no file has zero coverage)

---

## File Classification

| New/Modified File | Role | Data Flow | Closest Analog | Match Quality |
|-------------------|------|-----------|----------------|---------------|
| `levenshtein.go` | algorithm | transform | `normalise.go` | role-match (same pure-function, ASCII-fast-path, no-alloc discipline) |
| `damerau_osa.go` | algorithm | transform | `levenshtein.go` (Wave 1) | exact (same DP pattern, +1 extra row) |
| `damerau_full.go` | algorithm | transform | `levenshtein.go` (Wave 1) | role-match (same DP pattern, +aux table) |
| `hamming.go` | algorithm | transform | `levenshtein.go` (Wave 1) | role-match (simpler loop, same normalisation) |
| `jaro.go` | algorithm | transform | `normalise.go` | role-match (pure function, stack bool arrays) |
| `jarowinkler.go` | algorithm | transform | `jaro.go` (Wave 2) | exact (thin wrapper, same header discipline) |
| `dispatch_levenshtein.go` | registration | — | `algoid.go` (dispatch skeleton) | role-match (var-init side-effect pattern) |
| `dispatch_damerau_osa.go` | registration | — | `dispatch_levenshtein.go` (Wave 1) | exact |
| `dispatch_damerau_full.go` | registration | — | `dispatch_levenshtein.go` (Wave 1) | exact |
| `dispatch_hamming.go` | registration | — | `dispatch_levenshtein.go` (Wave 1) | exact |
| `dispatch_jaro.go` | registration | — | `dispatch_levenshtein.go` (Wave 1) | exact |
| `dispatch_jarowinkler.go` | registration | — | `dispatch_levenshtein.go` (Wave 1) | exact |
| `levenshtein_test.go` | test (unit) | — | `normalise_test.go` | exact |
| `damerau_osa_test.go` | test (unit) | — | `levenshtein_test.go` (Wave 1) | exact |
| `damerau_full_test.go` | test (unit) | — | `levenshtein_test.go` (Wave 1) | exact |
| `hamming_test.go` | test (unit) | — | `levenshtein_test.go` (Wave 1) | exact |
| `jaro_test.go` | test (unit) | — | `levenshtein_test.go` (Wave 1) | exact |
| `jarowinkler_test.go` | test (unit) | — | `levenshtein_test.go` (Wave 1) | exact |
| `levenshtein_bench_test.go` | test (benchmark) | — | `normalise_bench_test.go` | exact |
| `damerau_osa_bench_test.go` | test (benchmark) | — | `levenshtein_bench_test.go` (Wave 1) | exact |
| `damerau_full_bench_test.go` | test (benchmark) | — | `levenshtein_bench_test.go` (Wave 1) | exact |
| `hamming_bench_test.go` | test (benchmark) | — | `levenshtein_bench_test.go` (Wave 1) | exact |
| `jaro_bench_test.go` | test (benchmark) | — | `levenshtein_bench_test.go` (Wave 1) | exact |
| `jarowinkler_bench_test.go` | test (benchmark) | — | `levenshtein_bench_test.go` (Wave 1) | exact |
| `levenshtein_fuzz_test.go` | test (fuzz) | — | `normalise_fuzz_test.go` | exact |
| `damerau_osa_fuzz_test.go` | test (fuzz) | — | `levenshtein_fuzz_test.go` (Wave 1) | exact |
| `damerau_full_fuzz_test.go` | test (fuzz) | — | `levenshtein_fuzz_test.go` (Wave 1) | exact |
| `hamming_fuzz_test.go` | test (fuzz) | — | `levenshtein_fuzz_test.go` (Wave 1) | exact |
| `jaro_fuzz_test.go` | test (fuzz) | — | `levenshtein_fuzz_test.go` (Wave 1) | exact |
| `jarowinkler_fuzz_test.go` | test (fuzz) | — | `levenshtein_fuzz_test.go` (Wave 1) | exact |
| `props_test.go` | test (property) | — | `normalise_test.go` (prop section) | role-match |
| `algorithms_golden_test.go` | test (golden) | — | `normalise_test.go` (golden section) | exact |
| `example_test.go` | test (example) | — | no existing example_test.go yet | partial (pattern from normalise_test.go ExampleXxx convention) |
| `tests/bdd/features/levenshtein.feature` (and 5 others) | BDD feature | — | `tests/bdd/doc.go` (harness only) | partial (no feature files exist yet; pattern from RESEARCH.md) |
| `tests/bdd/steps/algorithms_steps.go` | BDD steps | — | `tests/bdd/doc.go` (harness only) | partial (no step files exist yet) |
| `examples/identifier-similarity/main.go` | utility | — | no analog | none (first runnable example) |
| `examples/identifier-similarity/main_test.go` | test (meta) | — | no analog | none (first meta-test) |
| `export_test.go` (modified) | re-export | — | `export_test.go` (existing) | exact |

---

## Pattern Assignments

### Pattern 1: License + Source Header (every `.go` file)

**Analog:** `/Users/johnny/Development/fuzzymatch/normalise.go` lines 1–51

Every `.go` file in the repository opens with this exact Apache-2.0 header block (lines 1–13), then a file-level package-doc comment block, then `package fuzzymatch`. The file-level doc comment varies by file but always appears between the license block and the `package` declaration.

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

// <file-level doc comment here>
//
// Source: <Author, Year>. "<Title>." <Journal/Conference>, <vol>(<issue>):<pages>.
// Implementation: <brief algorithm note>
//
// Implementation discipline:
//
//   - ASCII fast path ...
//   - NO init()-time table builds (per docs/requirements.md §5(12) and ...
//   - NO map iteration on output paths ...
//   - NO transcendental float operations ...

package fuzzymatch
```

**Gotcha:** Algorithm implementation files MUST include a `// Source:` block citing the primary academic source at the top of the file-level doc comment. The license-header verifier script (`scripts/verify-license-headers.sh`) asserts the Apache-2.0 block is present on every `.go` file. Missing or altered headers fail CI.

---

### Pattern 2: Algorithm Implementation File Structure

**Analog:** `/Users/johnny/Development/fuzzymatch/normalise.go` (full file, 453 lines)

The canonical structure for an algorithm implementation file is:

```
1. Apache-2.0 license block (lines 1–13)
2. File-level package-doc comment:
   - Source: <primary citation>
   - Recurrence or formula in godoc block
   - Implementation discipline bullet list
3. package fuzzymatch
4. import block (stdlib only for Phase 2 algorithms — no x/text imports)
5. Unexported constants (with godoc citing the paper):
   const winklerPrefixScale    = 0.1  // Winkler 1990 p. 357 — "p"
   const winklerMaxPrefix      = 4    // Winkler 1990 p. 357 — "L_max"
   const winklerBoostThreshold = 0.7  // Winkler 1990 p. 357 — boost condition
6. Public functions (Distance, DistanceRunes, Score, ScoreRunes)
7. Unexported inner helpers (DP kernels, score normalisers)
```

From `normalise.go` lines 44–52 (import block pattern — algorithm files will have minimal imports):

```go
package fuzzymatch

import (
    "unicode/utf8"
)
```

Phase 2 algorithm files import only `unicode/utf8` (for the rune variants) — no `golang.org/x/text` dependency.

---

### Pattern 3: isASCII Helper — Reuse from normalise.go

**Analog:** `/Users/johnny/Development/fuzzymatch/normalise.go` lines 159–168

```go
// isASCII reports whether every byte of s is strictly less than 0x80.
// Empty s returns true (vacuously).
func isASCII(s string) bool {
    for i := 0; i < len(s); i++ {
        if s[i] >= 0x80 {
            return false
        }
    }
    return true
}
```

**Gotcha:** `isASCII` is already declared in `normalise.go` in `package fuzzymatch`. Phase 2 algorithm files in the SAME package MUST NOT redeclare it — they call the existing `isASCII` directly. Do not copy or redeclare this function. The same applies to `isUpperASCII`, `isLowerASCII`, `isASCIISpace`, and `lowerRune` — all are already in `normalise.go` and available to all files in `package fuzzymatch`.

---

### Pattern 4: Two-Row DP + ASCII Fast Path + Stack Buffer

**Analog:** `normalise.go` lines 190–249 (the `normaliseASCII` function for the structural pattern); algorithm-specific DP from RESEARCH.md implementation patterns.

This is the most critical new pattern for Phase 2. No existing algorithm file contains DP code (normalise.go uses byte iteration, not DP), so the pattern is derived from the RESEARCH.md specification and must be established by Wave 1.

The canonical pattern (Levenshtein as the reference implementation):

```go
// LevenshteinDistance returns the Levenshtein edit distance between a and b.
// Score normalisation: LevenshteinScore returns 1.0 - dist/max(len(a), len(b)).
// Both-empty → 0 (distance), 1.0 (score). One-empty → max(len) (distance), 0.0 (score).
func LevenshteinDistance(a, b string) int {
    if a == b { return 0 }      // fast identity (catches both-empty too)
    m, n := len(a), len(b)
    if m == 0 { return n }
    if n == 0 { return m }
    // Make b the shorter string for the inner loop (reduces inner-loop length).
    if m < n { a, b = b, a; m, n = n, m }

    if n <= 64 {
        // Stack-allocate two rows of n+1 ints (max 65*2 = 130 ints = 1040 bytes).
        // The slice headers point into buf; buf itself is on the stack per
        // Go's escape analysis (confirmed via go build -gcflags="-m=2").
        var buf [65 * 2]int
        return levenshteinDP(a, b, m, n, buf[:n+1], buf[n+1:n+n+2])
    }
    // Heap path for inputs > 64 bytes.
    return levenshteinDP(a, b, m, n, make([]int, n+1), make([]int, n+1))
}

// levenshteinDP is the inner DP kernel. prev and curr must each be len n+1.
// Operates on string bytes directly (no []byte conversion — zero allocation).
func levenshteinDP(a, b string, m, n int, prev, curr []int) int {
    // Initialise prev row: D[0,j] = j (insert j chars).
    for j := 0; j <= n; j++ { prev[j] = j }
    for i := 1; i <= m; i++ {
        curr[0] = i  // D[i,0] = i (delete i chars)
        for j := 1; j <= n; j++ {
            cost := 1
            if a[i-1] == b[j-1] { cost = 0 }
            // min3: inline for performance (no function call overhead)
            v := prev[j] + 1     // deletion
            if w := curr[j-1] + 1; w < v { v = w }  // insertion
            if w := prev[j-1] + cost; w < v { v = w } // substitution
            curr[j] = v
        }
        prev, curr = curr, prev  // swap: prev becomes the current row
    }
    return prev[n]  // after swap, answer is in prev[n]
}
```

**DL-OSA stack size:** Three rows needed — `var buf [65 * 3]int` (1560 bytes on the stack).
**DL-Full stack size:** Two DP rows + 256-int aux table — `var dpBuf [65 * 2]int` + `var lastOcc [256]int` (total ~3088 bytes).
**Jaro stack size:** Two bool arrays — `var matchA [256]bool; var matchB [256]bool` (512 bytes).

**Gotcha (escape analysis):** Pass `buf[lo:hi]` slices as function arguments to `levenshteinDP`. The slice header stays on the stack. DO NOT store the slice in a struct field or return it — that causes heap escape. Verify with `go build -gcflags="-m=2" ./... 2>&1 | grep -E "does not escape|escapes to heap"`.

**Gotcha (byte indexing):** Use `a[i-1]` (byte-indexed string) NOT `[]byte(a)` in the DP kernel. `[]byte(a)` allocates. Direct byte indexing into a `string` is zero-allocation.

---

### Pattern 5: Score Normalisation (all six algorithms)

**Analog:** Derived from `normalise.go` empty-input guard pattern (lines 144–157) and RESEARCH.md §Score Normalisation.

Every `XxxScore` function applies this guard before division:

```go
func LevenshteinScore(a, b string) float64 {
    maxLen := len(a)
    if len(b) > maxLen { maxLen = len(b) }
    if maxLen == 0 { return 1.0 }  // both-empty → identity
    dist := LevenshteinDistance(a, b)
    return 1.0 - float64(dist)/float64(maxLen)
}
```

For Hamming (LOCKED behaviour per CONTEXT.md): unequal length → `dist = max(len(a), len(b))`, which makes `score = 1 - maxLen/maxLen = 0.0`. No error, no panic. The same normalisation formula applies to all four DP-based algorithms.

For Jaro: the division guard is `if m == 0 { return 0.0 }` inside the formula (m = matched characters). The formula yields [0,1] directly without the `1 - dist/maxLen` form.

**Gotcha (NaN prevention):** The `if maxLen == 0 { return 1.0 }` guard is mandatory. Without it, `float64(0)/float64(0)` = NaN. Every `XxxScore` function MUST have this guard before the division. Property test `TestProp_XxxScore_NoNaN` verifies it.

**Gotcha (negative zero):** `1.0 - 1.0` in IEEE-754 is `+0.0`, not `-0.0`. The formula produces `+0.0` when `dist == maxLen`. No special handling needed. Property test `TestProp_XxxScore_NoNegativeZero` verifies it with `math.Signbit`.

---

### Pattern 6: Dispatch Registration

**Analog:** `/Users/johnny/Development/fuzzymatch/algoid.go` lines 310–324 (dispatch declaration); `/Users/johnny/Development/fuzzymatch/.planning/phases/02-core-character-algorithms-six/02-RESEARCH.md` §Dispatch Registration Pattern.

The `dispatch` array is declared in `algoid.go` (line 324):
```go
var dispatch [numAlgorithms]func(a, b string) float64
```

Each algorithm registers into its slot via a separate `dispatch_xxx.go` file:

```go
// Copyright 2026 AxonOps Limited
// [full Apache-2.0 header]

// dispatch_levenshtein.go registers LevenshteinScore into the dispatch table
// at package load time. This file MUST be the sole writer to dispatch[AlgoLevenshtein].
// See algoid.go for the dispatch array declaration.

package fuzzymatch

// _ ensures dispatch[AlgoLevenshtein] is populated before any call to the
// Scorer (Phase 8) or Extract (Phase 10) that reads the dispatch table.
// The var _ = func()bool{...}() idiom is the canonical way to run
// package-level side effects without init() (per determinism-standards §13.5).
var _ = func() bool {
    dispatch[AlgoLevenshtein] = LevenshteinScore
    return true
}()
```

**Gotcha:** Each `dispatch_xxx.go` MUST touch only its own slot. Wave 2 parallel plans each own exactly one `dispatch_xxx.go` file — zero merge risk. The `algoid.go` file is NOT modified in Wave 2.

**Gotcha (double-registration):** The dispatch slot is nil at Phase 1. Writing it once sets it; no double-registration guard is implemented at this stage. If the same slot is written twice (e.g. two `dispatch_levenshtein.go` files), the second write silently wins. The `TestDispatch_AllNilAtPhase1` test in `algoid_test.go` must be updated in Wave 1 to assert that `dispatch[AlgoLevenshtein]` is NOT nil after the registration lands.

---

### Pattern 7: Unit Test File Structure

**Analog:** `/Users/johnny/Development/fuzzymatch/normalise_test.go` (full file, 485 lines)

```go
// Copyright 2026 AxonOps Limited
// [full Apache-2.0 header]

// levenshtein_test.go pins the public-API contract of levenshtein.go:
// both-empty identity, one-empty distance, canonical reference vectors
// from Levenshtein 1965 / Wagner-Fischer 1974, symmetry, byte vs rune
// path equivalence on ASCII inputs, and the NaN/Inf guards.
//
// Stdlib `testing` only — no testify in root tests, per
// .claude/skills/go-coding-standards.

package fuzzymatch_test

import (
    "testing"

    "github.com/axonops/fuzzymatch"
)

// TestLevenshtein_BothEmpty asserts the documented identity invariant:
// Distance("","") = 0; Score("","") = 1.0.
func TestLevenshtein_BothEmpty(t *testing.T) {
    if got := fuzzymatch.LevenshteinDistance("", ""); got != 0 {
        t.Errorf("LevenshteinDistance(\"\",\"\") = %d; want 0", got)
    }
    if got := fuzzymatch.LevenshteinScore("", ""); got != 1.0 {
        t.Errorf("LevenshteinScore(\"\",\"\") = %g; want 1.0", got)
    }
}

// TestLevenshtein_ReferenceVectors pins the canonical pairs from
// Levenshtein 1965 (Soviet Physics Doklady 10(8):707–710) /
// Wagner-Fischer 1974 (JACM 21(1):168–173).
func TestLevenshtein_ReferenceVectors(t *testing.T) {
    tests := []struct {
        a, b  string
        dist  int
        score float64
    }{
        {"kitten", "sitting", 3, 1.0 - 3.0/7.0},  // Wagner-Fischer 1974
        {"saturday", "sunday", 3, 1.0 - 3.0/8.0}, // Wagner-Fischer 1974
        {"abc", "abc", 0, 1.0},
        {"", "abc", 3, 0.0},
        {"abc", "", 3, 0.0},
    }
    for _, tt := range tests {
        t.Run(tt.a+"_"+tt.b, func(t *testing.T) {
            if got := fuzzymatch.LevenshteinDistance(tt.a, tt.b); got != tt.dist {
                t.Errorf("LevenshteinDistance(%q, %q) = %d; want %d", tt.a, tt.b, got, tt.dist)
            }
            const tol = 1e-9
            if got := fuzzymatch.LevenshteinScore(tt.a, tt.b); abs(got-tt.score) > tol {
                t.Errorf("LevenshteinScore(%q, %q) = %g; want %g (tol %g)", tt.a, tt.b, got, tt.score, tol)
            }
        })
    }
}
```

**Naming convention** (from normalise_test.go patterns):
- `TestXxx_BothEmpty` — identity for distance algorithms
- `TestXxx_ReferenceVectors` — canonical pairs from primary source
- `TestXxx_Symmetry` — `Score(a,b) == Score(b,a)`
- `TestXxx_DistanceRunes` — rune variant on multi-byte input
- `TestXxx_ASCII_vs_Rune_Equivalence` — same ASCII input, same output from both paths

**Gotcha (no testify):** Root test files use stdlib `testing` only. `t.Errorf` / `t.Fatalf` with format strings, not `assert.Equal`. See `normalise_test.go` throughout.

**Gotcha (float comparison):** Use `abs(got - want) > tol` with `tol = 1e-9` for score comparisons — not `==`. The `abs` helper must be declared in the test file (no `math.Abs` import needed; or use `math.Abs` from stdlib `math` package with a regular import).

---

### Pattern 8: Rune Variant

**Analog:** no existing rune-variant algorithm in Phase 1. Pattern from RESEARCH.md §Rune Variant Strategy (Pattern A — eager conversion).

```go
// LevenshteinScoreRunes returns the Levenshtein similarity treating
// a and b as sequences of Unicode code points (runes) rather than bytes.
// This produces correct results for multi-byte UTF-8 strings (e.g.
// "café" is 4 runes but 5 bytes).
//
// The rune variant allocates two []rune slices. For ASCII inputs, prefer
// LevenshteinScore (zero allocations on inputs ≤ 64 bytes).
func LevenshteinScoreRunes(a, b string) float64 {
    ra := []rune(a)  // 1 alloc
    rb := []rune(b)  // 1 alloc
    maxLen := len(ra)
    if len(rb) > maxLen { maxLen = len(rb) }
    if maxLen == 0 { return 1.0 }
    dist := levenshteinDistanceRunes(ra, rb)
    return 1.0 - float64(dist)/float64(maxLen)
}
```

**Gotcha:** The rune path allocates — this is documented and expected. The 0-alloc budget applies ONLY to the byte path on ASCII ≤ 64 bytes. Property tests and fuzz tests cover the rune path for correctness, not allocation budget.

---

### Pattern 9: Benchmark File Structure

**Analog:** `/Users/johnny/Development/fuzzymatch/normalise_bench_test.go` (full file, 180 lines)

```go
// Copyright 2026 AxonOps Limited
// [full Apache-2.0 header]

// levenshtein_bench_test.go runs allocation-aware benchmarks for
// LevenshteinScore and LevenshteinDistance at four input sizes.
// b.ReportAllocs() on every benchmark gates allocation regressions
// in bench.txt via benchstat.
//
// Performance budgets per .claude/skills/performance-standards:
//   - ASCII <= 50 chars:  target < 1 µs/op, 0 allocs/op
//   - ASCII <= 500 chars: proportional, <= 2 allocs/op
//   - Unicode short:      target < 2 µs/op, <= 2 allocs/op (rune path)

package fuzzymatch_test

import (
    "testing"

    "github.com/axonops/fuzzymatch"
)

// BenchmarkLevenshteinScore_ASCII_Short exercises the zero-alloc fast
// path for a short ASCII pair (≤ 64 bytes, stack-allocated DP buffer).
// Target: < 1 µs/op, 0 allocs/op.
func BenchmarkLevenshteinScore_ASCII_Short(b *testing.B) {
    b.ReportAllocs()
    b.ResetTimer()
    var sink float64
    for i := 0; i < b.N; i++ {
        sink = fuzzymatch.LevenshteinScore("kitten", "sitting")
    }
    if sink < 0 {
        b.Fatal("sink unexpectedly negative — compiler folded the benchmark away")
    }
}

// BenchmarkLevenshteinScore_ASCII_Medium exercises the fast path at
// 50-char input. Stack buffer still used (n ≤ 64).
func BenchmarkLevenshteinScore_ASCII_Medium(b *testing.B) { ... }

// BenchmarkLevenshteinScore_ASCII_Long exercises the heap path (n > 64).
func BenchmarkLevenshteinScore_ASCII_Long(b *testing.B) { ... }

// BenchmarkLevenshteinScore_Unicode_Short exercises the rune path.
func BenchmarkLevenshteinScore_Unicode_Short(b *testing.B) { ... }
```

**Gotcha:** Use `var sink float64` and a `if sink < 0 { b.Fatal(...) }` check to prevent the compiler from dead-code-eliminating the benchmark call. Pattern from `normalise_bench_test.go` lines 58–65.

**Gotcha:** `b.ReportAllocs()` must appear before `b.ResetTimer()`. Order is shown above.

---

### Pattern 10: Fuzz Test File Structure

**Analog:** `/Users/johnny/Development/fuzzymatch/normalise_fuzz_test.go` (full file, 101 lines)

```go
// Copyright 2026 AxonOps Limited
// [full Apache-2.0 header]

// levenshtein_fuzz_test.go runs native Go fuzzing against LevenshteinScore.
// Two properties for any input:
//   1. Never panics (implicit — propagates to fuzz harness as crash).
//   2. Score is in [0.0, 1.0].

package fuzzymatch_test

import (
    "math"
    "testing"

    "github.com/axonops/fuzzymatch"
)

// FuzzLevenshteinScore asserts panic-free + score in [0,1] for all inputs.
func FuzzLevenshteinScore(f *testing.F) {
    // Programmatic seed entries (canonical reference vectors).
    for _, pair := range []struct{ a, b string }{
        {"kitten", "sitting"},
        {"saturday", "sunday"},
        {"abc", "abc"},
        {"", "abc"},
        {"", ""},
        {"\xff\xfe", "abc"},      // invalid UTF-8
        {"Привет", "привет"},      // Cyrillic
    } {
        f.Add(pair.a, pair.b)
    }

    f.Fuzz(func(t *testing.T, a, b string) {
        // Property 1: must not panic (implicit).
        got := fuzzymatch.LevenshteinScore(a, b)

        // Property 2: score in [0.0, 1.0].
        if math.IsNaN(got) {
            t.Errorf("LevenshteinScore(%q, %q) = NaN", a, b)
        }
        if math.IsInf(got, 0) {
            t.Errorf("LevenshteinScore(%q, %q) = Inf", a, b)
        }
        if got < 0.0 || got > 1.0 {
            t.Errorf("LevenshteinScore(%q, %q) = %g; want in [0,1]", a, b, got)
        }
    })
}
```

**Gotcha:** The fuzz corpus directory for `FuzzLevenshteinScore` goes in `testdata/fuzz/FuzzLevenshteinScore/`. Each algorithm has its own corpus directory. Create at least one `seed-001` file per algorithm in the `go test fuzz v1` format (see `testdata/fuzz/FuzzNormalise/seed-001` for the format). Pattern from `normalise_fuzz_test.go` lines 54–75.

---

### Pattern 11: Property Test File

**Analog:** `/Users/johnny/Development/fuzzymatch/normalise_test.go` lines 275–346 (the `TestProp_*` section)

The Phase 2 `props_test.go` extends the property-test discipline to algorithm scores. All six algorithms share the same invariant template:

```go
// Copyright 2026 AxonOps Limited
// [full Apache-2.0 header]

// props_test.go contains testing/quick property tests for Phase 2's six
// character-based algorithms. Each algorithm is covered by:
//   - TestProp_XxxScore_RangeBounds   ([0,1] for any input)
//   - TestProp_XxxScore_Identity      (Score(x,x) == 1.0 for non-empty x)
//   - TestProp_XxxScore_Symmetric     (Score(a,b) == Score(b,a))
//   - TestProp_XxxDistance_TriangleInequality (for DP algorithms, not Jaro/JW)
//   - TestProp_XxxScore_NoNaN
//   - TestProp_XxxScore_NoInf
//   - TestProp_XxxScore_NoNegativeZero

package fuzzymatch_test

import (
    "math"
    "testing"
    "testing/quick"

    "github.com/axonops/fuzzymatch"
)

func TestProp_LevenshteinScore_RangeBounds(t *testing.T) {
    f := func(a, b string) bool {
        s := fuzzymatch.LevenshteinScore(a, b)
        return s >= 0.0 && s <= 1.0
    }
    if err := quick.Check(f, nil); err != nil {
        t.Errorf("LevenshteinScore out of [0,1]: %v", err)
    }
}

func TestProp_LevenshteinScore_Identity(t *testing.T) {
    f := func(x string) bool {
        if x == "" { return true }  // both-empty special case (score is 1.0, skip to avoid confusion)
        return fuzzymatch.LevenshteinScore(x, x) == 1.0
    }
    if err := quick.Check(f, nil); err != nil {
        t.Errorf("LevenshteinScore identity violated: %v", err)
    }
}

func TestProp_LevenshteinScore_Symmetric(t *testing.T) {
    f := func(a, b string) bool {
        return fuzzymatch.LevenshteinScore(a, b) == fuzzymatch.LevenshteinScore(b, a)
    }
    if err := quick.Check(f, nil); err != nil {
        t.Errorf("LevenshteinScore not symmetric: %v", err)
    }
}

func TestProp_LevenshteinDistance_TriangleInequality(t *testing.T) {
    f := func(a, b, c string) bool {
        dAC := fuzzymatch.LevenshteinDistance(a, c)
        dAB := fuzzymatch.LevenshteinDistance(a, b)
        dBC := fuzzymatch.LevenshteinDistance(b, c)
        return dAC <= dAB+dBC
    }
    if err := quick.Check(f, nil); err != nil {
        t.Errorf("LevenshteinDistance triangle inequality violated: %v", err)
    }
}

func TestProp_LevenshteinScore_NoNaN(t *testing.T) {
    f := func(a, b string) bool { return !math.IsNaN(fuzzymatch.LevenshteinScore(a, b)) }
    if err := quick.Check(f, nil); err != nil { t.Errorf("LevenshteinScore produced NaN: %v", err) }
}

func TestProp_LevenshteinScore_NoInf(t *testing.T) {
    f := func(a, b string) bool { return !math.IsInf(fuzzymatch.LevenshteinScore(a, b), 0) }
    if err := quick.Check(f, nil); err != nil { t.Errorf("LevenshteinScore produced Inf: %v", err) }
}

func TestProp_LevenshteinScore_NoNegativeZero(t *testing.T) {
    f := func(a, b string) bool {
        s := fuzzymatch.LevenshteinScore(a, b)
        return s != 0.0 || !math.Signbit(s)
    }
    if err := quick.Check(f, nil); err != nil { t.Errorf("LevenshteinScore produced -0.0: %v", err) }
}
```

**Gotcha (Hamming triangle inequality):** Hamming's triangle inequality test MUST constrain inputs to equal-length strings (the property fails on unequal-length inputs under the silent-zero policy). Use a generator or add a length guard:

```go
func TestProp_HammingDistance_TriangleInequality_EqualLength(t *testing.T) {
    f := func(base string, delta1, delta2 [8]byte) bool {
        // All three strings have the same length as base.
        // Build b and c by XOR-flipping bytes of base to produce same-length strings.
        // ... (implementation details)
    }
    if err := quick.Check(f, nil); err != nil { t.Errorf("...") }
}
```

**Gotcha (Jaro/JW):** Do NOT write a triangle-inequality property test for Jaro or Jaro-Winkler. They are not metrics and the property does not hold. Document explicitly in the godoc of `jaro.go`: "Jaro is not a metric; the triangle inequality does not hold."

---

### Pattern 12: Golden File Extension

**Analog:** `/Users/johnny/Development/fuzzymatch/normalise_test.go` lines 354–485 (golden test section) and `/Users/johnny/Development/fuzzymatch/golden_test.go` lines 66–88 (assertGolden)

The `algorithms_golden_test.go` file mirrors the normalisation golden pattern exactly:

```go
// Copyright 2026 AxonOps Limited
// [full Apache-2.0 header]

// algorithms_golden_test.go pins the score output of all Phase 2 algorithms
// byte-for-byte across the 5-platform CI matrix. It uses the LOCKED
// canonical marshal form from plan 01-04.

package fuzzymatch_test

import (
    "sort"
    "testing"

    "github.com/axonops/fuzzymatch"
)

// goldenAlgorithmEntry is one (algorithm, a, b, expected_score) case
// in testdata/golden/algorithms.json. Field names are part of the LOCKED
// JSON contract — renaming any field is a major-version-bump event.
type goldenAlgorithmEntry struct {
    Name          string  `json:"name"`
    Algorithm     string  `json:"algorithm"`
    A             string  `json:"a"`
    B             string  `json:"b"`
    ExpectedScore float64 `json:"expected_score"`
}

// goldenAlgorithmsFile wraps entries with a version field.
type goldenAlgorithmsFile struct {
    Version int                    `json:"version"`
    Entries []goldenAlgorithmEntry `json:"entries"`
}

// TestGolden_Algorithms pins score outputs across CI matrix platforms.
// Run with `-update` to rewrite testdata/golden/algorithms.json.
func TestGolden_Algorithms(t *testing.T) {
    entries := buildAlgorithmGoldenEntries(t)
    sort.Slice(entries, func(i, j int) bool {
        return entries[i].Name < entries[j].Name
    })
    file := goldenAlgorithmsFile{Version: 1, Entries: entries}
    assertGolden(t, "algorithms.json", file)
}

// buildAlgorithmGoldenEntries computes ExpectedScore from current code.
func buildAlgorithmGoldenEntries(t *testing.T) []goldenAlgorithmEntry {
    t.Helper()
    return []goldenAlgorithmEntry{
        {Name: "Levenshtein_kitten_sitting", Algorithm: "Levenshtein",
            A: "kitten", B: "sitting",
            ExpectedScore: fuzzymatch.LevenshteinScore("kitten", "sitting")},
        // ... all entries from RESEARCH.md §Golden File Integration
    }
}
```

**Gotcha (Wave 1 establishes the file):** Wave 1 (Levenshtein) creates `testdata/golden/algorithms.json` with only Levenshtein entries via `go test -run TestGolden_Algorithms -update ./...`. Wave 2 plans use staging files in `testdata/golden/_staging/<algo>.json` (same schema, same struct types, only their algorithm's entries). A post-Wave-2 merge step re-marshals all staging files into `algorithms.json`.

**Gotcha (sort before marshal):** Entries MUST be sorted by `Name` before `assertGolden` is called. Use `sort.Slice(entries, func(i, j int) bool { return entries[i].Name < entries[j].Name })`. The canonical byte form assumes a deterministic order — callers provide it.

**Gotcha (float64 JSON stability):** `encoding/json` uses `strconv.AppendFloat` with `'g'` format for `float64`. This is deterministic across platforms for the values arising from these algorithms. No special float-marshalling is needed.

---

### Pattern 13: AlgoID Enum Extension

**Analog:** `/Users/johnny/Development/fuzzymatch/algoid.go` lines 53–181

The Phase 2 AlgoID constants are ALREADY DECLARED in `algoid.go` at the correct iota positions (0–5):

```go
// From algoid.go lines 53–88 (already present, DO NOT modify):
const (
    AlgoLevenshtein AlgoID = iota      // 0
    AlgoDamerauLevenshteinOSA          // 1
    AlgoDamerauLevenshteinFull         // 2
    AlgoHamming                        // 3
    AlgoJaro                           // 4
    AlgoJaroWinkler                    // 5
    // ... 17 more constants (6–22)
)
```

**Critical:** `algoid.go` is NOT modified by any Phase 2 plan. All 23 AlgoID constants were declared in Phase 1 (plan 01-05). The dispatch slots are nil at Phase 1; Phase 2 fills positions 0–5 via `dispatch_xxx.go` files.

The only test that must be updated in Phase 2 is `TestDispatch_AllNilAtPhase1` in `algoid_test.go`. After Wave 1 registers Levenshtein, that test must become `TestDispatch_AllAlgorithmsHaveEntries` or be extended to assert only slots 0–5 are non-nil (and 6–22 remain nil). The planner should add this to the Wave 1 plan.

---

### Pattern 14: godoc Example Functions

**Analog:** The `example_test.go` file does not yet exist. Pattern from Go stdlib conventions and the BDD structure in `tests/bdd/doc.go`.

```go
// Copyright 2026 AxonOps Limited
// [full Apache-2.0 header]

// example_test.go contains runnable godoc examples for the Phase 2
// character-based algorithms. Each ExampleXxx function appears on
// pkg.go.dev alongside the function it documents.

package fuzzymatch_test

import (
    "fmt"

    "github.com/axonops/fuzzymatch"
)

func ExampleLevenshteinScore() {
    fmt.Printf("%.4f\n", fuzzymatch.LevenshteinScore("kitten", "sitting"))
    // Output:
    // 0.5714
}

func ExampleJaroWinklerScore() {
    fmt.Printf("%.4f\n", fuzzymatch.JaroWinklerScore("MARTHA", "MARHTA"))
    // Output:
    // 0.9611
}

func ExampleHammingScore() {
    // Equal-length inputs:
    fmt.Printf("%.4f\n", fuzzymatch.HammingScore("karolin", "kathrin"))
    // Unequal-length inputs return 0.0 silently:
    fmt.Printf("%.4f\n", fuzzymatch.HammingScore("abc", "ab"))
    // Output:
    // 0.5714
    // 0.0000
}
```

**Gotcha:** The `// Output:` comment must match byte-for-byte. Use `fmt.Printf("%.4f\n", ...)` with a fixed precision that matches the pinned reference vectors. Verify the output by running `go test -run ExampleXxx ./...` and checking the `ok` result.

**Gotcha:** Example functions are in `package fuzzymatch_test` (external test package), not `package fuzzymatch`. They use the public API only.

---

### Pattern 15: BDD Feature File + Step Definitions

**Analog:** `/Users/johnny/Development/fuzzymatch/tests/bdd/doc.go` (harness structure); RESEARCH.md §BDD Scenario Coverage (canonical feature file shape).

No feature files or step files exist yet. The BDD sub-module `go.mod` at `/Users/johnny/Development/fuzzymatch/tests/bdd/go.mod` shows the available deps:

```
require (
    github.com/axonops/fuzzymatch v0.0.0...
    github.com/cucumber/godog v0.15.0
    github.com/stretchr/testify v1.10.0
    go.uber.org/goleak v1.3.0
)
replace github.com/axonops/fuzzymatch => ../..
```

Feature file pattern (one file per algorithm, no merge conflicts):

```gherkin
# tests/bdd/features/levenshtein.feature
Feature: Levenshtein similarity algorithm
  # Primary source: Levenshtein 1965 (Soviet Physics Doklady 10(8):707-710)
  # Implementation: Wagner-Fischer two-row DP with ASCII fast path

  Scenario Outline: canonical reference vectors
    When I compute the Levenshtein score between "<a>" and "<b>"
    Then the score should be approximately <score> within 0.0001

    Examples:
      | a        | b        | score  |
      | kitten   | sitting  | 0.5714 |
      | saturday | sunday   | 0.6250 |
      | abc      | abc      | 1.0000 |

  Scenario: both-empty strings score 1.0
    When I compute the Levenshtein score between "" and ""
    Then the score should be exactly 1.0

  Scenario: one-empty string scores 0.0
    When I compute the Levenshtein score between "abc" and ""
    Then the score should be exactly 0.0

  Scenario: score is symmetric
    When I compute the Levenshtein score between "kitten" and "sitting"
    And I compute the Levenshtein score between "sitting" and "kitten"
    Then both scores should be equal
```

Step definitions pattern (in `tests/bdd/steps/algorithms_steps.go`):

```go
// Copyright 2026 AxonOps Limited
// [full Apache-2.0 header]

package steps

import (
    "fmt"
    "math"

    "github.com/axonops/fuzzymatch"
    "github.com/cucumber/godog"
    "github.com/stretchr/testify/assert"
)

// AlgorithmContext holds state between BDD steps.
type AlgorithmContext struct {
    lastScore  float64
    lastScore2 float64  // for "both scores should be equal" scenarios
}

func (ctx *AlgorithmContext) iComputeTheLevenshteinScoreBetween(a, b string) error {
    ctx.lastScore = fuzzymatch.LevenshteinScore(a, b)
    return nil
}

func (ctx *AlgorithmContext) theScoreShouldBeApproximately(expected, tolerance float64) error {
    if math.Abs(ctx.lastScore-expected) > tolerance {
        return fmt.Errorf("expected %f ± %f, got %f", expected, tolerance, ctx.lastScore)
    }
    return nil
}

func (ctx *AlgorithmContext) theScoreShouldBeExactly(expected float64) error {
    if ctx.lastScore != expected {
        return fmt.Errorf("expected exactly %f, got %f", expected, ctx.lastScore)
    }
    return nil
}

// InitializeScenario wires the step definitions into the godog suite.
func InitializeScenario(ctx *godog.ScenarioContext) {
    a := &AlgorithmContext{}
    ctx.Step(`^I compute the Levenshtein score between "([^"]*)" and "([^"]*)"$`, a.iComputeTheLevenshteinScoreBetween)
    ctx.Step(`^the score should be approximately (\d+\.\d+) within (\d+\.\d+)$`, a.theScoreShouldBeApproximately)
    ctx.Step(`^the score should be exactly (\d+\.\d+)$`, a.theScoreShouldBeExactly)
    // ... add steps for each algorithm
    _ = assert.Equal  // testify available in tests/bdd — use for assertion sugar in steps
}
```

**Gotcha:** testify IS permitted in `tests/bdd/steps/` (sub-module, not root). Do not use testify anywhere in root `*_test.go` files.

**Gotcha:** The BDD sub-module uses `replace github.com/axonops/fuzzymatch => ../..` so it builds against the local source tree. No need to publish before BDD tests run.

---

### Pattern 16: Examples Meta-Test

**Analog:** no existing `examples/` directory. Pattern from `normalise_bench_test.go` (sink pattern to prevent dead-code elimination) and Go's `os/exec` + stdout-capture idiom.

The `examples/identifier-similarity/main_test.go` captures the example binary's stdout and asserts it is byte-stable:

```go
// Copyright 2026 AxonOps Limited
// [full Apache-2.0 header]

// main_test.go is a meta-test for examples/identifier-similarity/main.go.
// It verifies that the example's stdout is byte-stable across runs and
// platforms (the output is deterministic because all six algorithm scores
// are deterministic for the hardcoded input pairs).

package main

import (
    "bytes"
    "os/exec"
    "testing"
)

// TestExample_Output verifies that running the example produces stable output.
// This test is the cross-platform determinism gate for the example program.
func TestExample_Output(t *testing.T) {
    // Run the example via `go run .` from the package directory.
    cmd := exec.Command("go", "run", ".")
    var out bytes.Buffer
    cmd.Stdout = &out
    cmd.Stderr = &out
    if err := cmd.Run(); err != nil {
        t.Fatalf("go run . failed: %v\noutput:\n%s", err, out.String())
    }
    got := out.String()
    // Assert stable output — the exact expected string is committed
    // after the first run (via `go test -run TestExample_Output -update ./...`
    // or by inspecting and committing the output manually).
    const want = `... (hardcoded expected table output)`
    if got != want {
        t.Errorf("example output changed.\ngot:\n%s\nwant:\n%s", got, want)
    }
}
```

**Gotcha:** The `examples/identifier-similarity/main.go` is `package main` (a standalone runnable). It is NOT in `package fuzzymatch`. The test file is also `package main` (in the same directory).

**Gotcha:** The expected output string must be committed after the first successful run. The meta-test is a "golden test for stdout" — update the `want` constant when the example's output is intentionally changed.

---

## Shared Patterns

### File Header (Apply to ALL new `.go` files)

**Source:** `/Users/johnny/Development/fuzzymatch/normalise.go` lines 1–13 and `/Users/johnny/Development/fuzzymatch/algoid.go` lines 1–13.

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

**Apply to:** every `.go` file. The `scripts/verify-license-headers.sh` CI gate fails if any `.go` file lacks this exact header.

---

### gosec G115 Avoidance (rune-to-byte conversions)

**Source:** `/Users/johnny/Development/fuzzymatch/normalise.go` lines 414–420 and the `isASCIIWhitespaceRune` function.

```go
// isASCIIWhitespaceRune is the rune-typed counterpart to isASCIISpace.
// We avoid converting r to byte here because that conversion triggers
// gosec G115 even on the already-bounded r < 0x80 path.
func isASCIIWhitespaceRune(r rune) bool {
    switch r {
    case ' ', '\t', '\n', '\r', '\v', '\f':
        return true
    }
    return false
}
```

**Apply to:** any place in algorithm files where a rune is known to be < 0x80 but must be compared to byte values. Do NOT cast `byte(r)` even when bounded — gosec G115 fires unconditionally on the cast.

---

### gocyclo Suppression for High-Branch Functions

**Source:** `/Users/johnny/Development/fuzzymatch/algoid.go` line 213.

```go
func (id AlgoID) String() string { //nolint:gocyclo // one switch case per AlgoID is intentional — see godoc above
```

**Apply to:** Jaro and DL-Full inner loops, if their cyclomatic complexity exceeds 10. Add `//nolint:gocyclo` with an inline explanation. Do NOT suppress other linters without explicit justification and a paired comment.

---

### No Map Iteration on Output Paths

**Source:** `/Users/johnny/Development/fuzzymatch/algoid.go` lines 269–308 (AlgoIDs uses slice literal, not map); `normalise.go` line 37 (implementation discipline comment).

**Apply to:** DL-Full's `lastOccurrence` map (used internally for lookups, never iterated to produce output). Confirm the map is only queried via `lastOcc[char]` on the algorithm's hot path. No `for k, v := range lastOcc` on any output path.

---

### Determinism: No math.Pow / math.Log / math.Exp / math.FMA

**Source:** `/Users/johnny/Development/fuzzymatch/normalise.go` line 40 (implementation discipline comment): "NO transcendental float operations."

**Apply to:** All six algorithm files. The permitted float operations are: `+`, `-`, `*`, `/`, `float64(int)`, `math.Abs` (IEEE-754 correctly rounded, permitted). DO NOT use `math.Pow`, `math.Log`, `math.Exp`, `math.Sqrt`, or `math.FMA` anywhere in Phase 2 algorithms.

---

### No init() Functions

**Source:** `/Users/johnny/Development/fuzzymatch/algoid.go` lines 15–28 (file-level doc comment): "No init() function appears in this file."

**Apply to:** ALL algorithm and dispatch files. The `var _ = func() bool { ... }()` idiom is the approved package-load-time side-effect mechanism. `init()` functions are forbidden.

---

## No Analog Found

All files in Phase 2 have strong Phase 1 analogs. The following have only partial coverage:

| File | Role | Data Flow | Reason |
|------|------|-----------|--------|
| `example_test.go` | test (example) | — | No existing `ExampleXxx` functions in Phase 1; pattern inferred from Go conventions |
| `tests/bdd/features/*.feature` | BDD feature | — | No feature files exist yet; pattern from RESEARCH.md and godog docs |
| `tests/bdd/steps/algorithms_steps.go` | BDD steps | — | No step files exist yet; godog API shown in RESEARCH.md |
| `examples/identifier-similarity/main.go` | utility (runnable) | — | First `examples/` directory; no Phase 1 example programs exist |
| `examples/identifier-similarity/main_test.go` | test (meta) | — | First meta-test; pattern inferred from `os/exec` + stdout-capture idiom |

For these files, the planner should use the patterns documented in RESEARCH.md §BDD Scenario Coverage and §File Layout as the primary guide.

---

## Metadata

**Analog search scope:** `/Users/johnny/Development/fuzzymatch/` root package files only (Phase 1 outputs).
**Files scanned:** 20 `.go` files (all Phase 1 production + test files).
**Pattern extraction date:** 2026-05-14

## PATTERN MAPPING COMPLETE
