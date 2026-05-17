---
phase: 08-composite-scorer
analysed: 2026-05-17T00:00:00Z
analyst: test-analyst
files_analysed: 9
files_analysed_list:
  - scorer.go
  - scorer_options.go
  - scorer_test.go
  - scorer_internal_test.go
  - scorer_options_test.go
  - scorer_options_internal_test.go
  - scorer_bench_test.go
  - scorer_golden_test.go
  - testdata/golden/scorer-default.json
  - tests/bdd/features/scorer.feature
  - tests/bdd/steps/scorer_steps.go
coverage_overall: 90.2%
coverage_scorer_go: 92.6% (NewScorer) / 100% all other functions / 75.0% DefaultScorer
coverage_scorer_options_go: 100% (every option)
coverage_errors_go: implicit 100% (sentinels exercised)
findings:
  blocking: 3
  important: 6
  nit: 5
  total: 14
status: gaps_to_address
---

# Phase 8 — Composite Scorer Test-Suite Health Analysis

## Headline

The Phase 8 test surface is **substantial and well-structured** — 28
top-level test functions in `scorer_test.go` alone, exhaustive
parameterised-option coverage in `scorer_options_test.go`, an
internal-state probe layer, a cross-platform golden file with 22
entries × 5 Scorer compositions, six allocation-aware benchmarks, and 12
BDD scenarios with a `goleak` gate. **Coverage on the new files is
excellent** (every `With*` option at 100%; every public `Scorer` method
at 100%). The two areas that fall short of the project's go-testing-
standards bar are:

1. **No `Fuzz*` harness exists for `Scorer.Score` / `ScoreAll` / `Match`**
   — the skill explicitly says "Every public function has a fuzz harness".
2. **Total package coverage is 90.2%, below the 95% overall target.** The
   shortfall is NOT in the Phase 8 surface; it traces to pre-existing
   gaps in `double_metaphone.go` (`dmPrep` 58.8%, `DoubleMetaphoneKeys`
   62.0%) and `nysiis.go` (`NYSIISCode` 72.6%) carried over from earlier
   phases. Phase 8 didn't introduce the gap but inherits the
   ≥ 95% guarantee on the meta-test.

The CR-01 (NaN-in-`WithThreshold`) and CR-02 (Tversky α=β=0) defects
flagged in 08-REVIEW.md surface here as **MISSING tests** — the test
suite did not catch either, because the obvious negative-case unit
tests aren't present.

The WR-01 flaky `uint16` overflow in `TestProp_Scorer_WeightSumOne`
remains in the codebase (see scorer_test.go:824); flagged as **FLAKY**.

## Coverage Numbers (actual)

```
go test -race -coverprofile=coverage.out ./... && go tool cover -func=coverage.out
```

| File / Symbol                    | Coverage | Notes |
|----------------------------------|----------|-------|
| **Total package**                | **90.2%** | Below the 95% overall floor — inherited from earlier-phase gaps in `double_metaphone.go` / `nysiis.go`. |
| `scorer.go:NewScorer`            | 92.6%    | Defensive Step 5 (`ErrInvalidAlgorithm` re-check) and Step 8 defensive `sum == 0 → ErrInvalidWeight` are not test-driven (correctly identified as defence-in-depth; provably unreachable from public API). |
| `scorer.go:Score`                | 100.0%   | |
| `scorer.go:Match`                | 100.0%   | |
| `scorer.go:ScoreAll`             | 100.0%   | |
| `scorer.go:Threshold`            | 100.0%   | |
| `scorer.go:Algorithms`           | 100.0%   | |
| `scorer.go:DefaultScorerOptions` | 100.0%   | |
| `scorer.go:DefaultScorer`        | 75.0%    | The `panic(...)` line in the catch-all is unreachable from the public API. |
| `scorer_options.go` (all 12)     | **100.0%** each | Every `With*` option fully exercised. |
| `errors.go` sentinels            | implicit 100% | `errors_test.go` enumerates every sentinel and exercises `errors.Is` identity + wrap. |

The Phase 8 surface is at 100% on every public symbol except the two
provably-unreachable defensive branches inside `NewScorer` and the
panic line in `DefaultScorer`. **Public API coverage on Phase 8 = 100%.**

The headline "90.2% overall" failure is a **pre-existing gap inherited
from phases 4 (phonetic) and 6 (tokenisation)**, not a regression
introduced by Phase 8. It is flagged here for transparency; the
remediation belongs to those phases.

## Per-Concern Findings

### Unit-test coverage of the public API

#### TEST-01 (MISSING / BLOCKING) — No `WithThreshold(NaN)` rejection test

**Files:** `scorer_options_test.go:170-194` (`TestWithThreshold_OutOfRange`).

The existing out-of-range test covers `-0.1`, `1.5`, `-100`, `100` but
NOT `math.NaN()`. The CR-01 review finding shows that
`t < 0.0 || t > 1.0` is both `false` for NaN, so `WithThreshold(NaN)`
silently succeeds, the Scorer freezes with `threshold = NaN`, and every
subsequent `Match` returns `false` (silent malfunction).

A unit test that imports `math` and asserts
`errors.Is(err, ErrInvalidThreshold)` for `math.NaN()` would have caught
this. Add:

```go
func TestWithThreshold_RejectsNaN(t *testing.T) {
    cfg, err := applyOptionForProbe(WithThreshold(math.NaN()))
    if !errors.Is(err, ErrInvalidThreshold) {
        t.Errorf("WithThreshold(NaN) err = %v; want ErrInvalidThreshold", err)
    }
    if _, set := probeThreshold(cfg); set {
        t.Errorf("thresholdSet = true after rejected WithThreshold(NaN); want false")
    }
}
```

Classification: **MISSING / BLOCKING**. Property-test gate didn't fire
either (the property test for `Score in [0,1]` doesn't generate a
NaN-threshold Scorer).

---

#### TEST-02 (MISSING / BLOCKING) — No `WithTverskyAlgorithm(α=0, β=0)` rejection test

**Files:** `scorer_options_test.go:419-445`.

`TestWithTverskyAlgorithm_InvalidAlpha` covers `alpha = -0.1` and
`TestWithTverskyAlgorithm_InvalidBeta` covers `beta = -0.1`, but the
combination `alpha = 0, beta = 0` is never exercised. Per CR-02, this
combination is documented as invalid (`tversky.go:241` panics on it)
but the option layer accepts it. The consumer's first `Score` call
panics at runtime — a behaviour-vs-contract mismatch that the test
suite does not catch.

Add:

```go
func TestWithTverskyAlgorithm_RejectsBothZero(t *testing.T) {
    _, err := applyOptionForProbe(WithTverskyAlgorithm(1.0, 0, 0, 3))
    if !errors.Is(err, ErrInvalidTverskyParam) {
        t.Errorf("WithTverskyAlgorithm(_, 0, 0, _) err = %v; want ErrInvalidTverskyParam", err)
    }
}
```

Classification: **MISSING / BLOCKING**. The Scorer's documented "fail
at construction time, never at Score time" contract is violated, and
the test suite is silent.

---

#### TEST-03 (MISSING / IMPORTANT) — No NaN/Inf score-input tests for `Scorer.Score`

The `TestProp_Scorer_NoNaN_NoInf` property test asserts `Score` never
returns NaN/Inf for arbitrary string input, but it cannot test what
happens when an **algorithm** (typically a parameterised one like
Tversky with both-zero α/β) returns NaN — because the option layer is
supposed to reject those configurations. Once TEST-02 lands, this
becomes moot.

However, the `WithNormaliseWeights(false)` path with explicitly hostile
raw weights (e.g. `+Inf`, `-Inf`, `NaN`) is not exercised. The option
layer accepts `weight = math.Inf(1)` today (only `weight <= 0` is
gated), and a Scorer with `weight = Inf` produces `Inf` composite —
silent malfunction. A unit test in `scorer_options_test.go` should
gate this:

```go
func TestWithAlgorithm_RejectsNaNInfWeight(t *testing.T) {
    for _, w := range []float64{math.NaN(), math.Inf(1), math.Inf(-1)} {
        _, err := applyOptionForProbe(WithAlgorithm(AlgoLevenshtein, w))
        if !errors.Is(err, ErrInvalidWeight) {
            t.Errorf("WithAlgorithm(_, %g) err = %v; want ErrInvalidWeight", w, err)
        }
    }
}
```

Classification: **MISSING / IMPORTANT**. Float pathology that the
option layer's `weight <= 0` gate doesn't catch.

---

#### TEST-04 (WEAK / IMPORTANT) — `Score(x, x) = 1.0` identity tested for ONE algorithm only

**Files:** `scorer_test.go:118-135` (`TestScorer_Score_Identity`).

The test uses a single-Levenshtein Scorer. The identity property
(`Score(x, x) = 1.0`) is one of the load-bearing mathematical
invariants (per `go-testing-standards/SKILL.md` line 26-27 and the
Scorer-level invariants on line 33-37 ("Composite ≥ min, ≤ max")).

The current test is sufficient to prove the single-algorithm case but
does NOT prove that `DefaultScorer().Score(x, x) = 1.0` for the
6-algorithm composition. The golden file's "hello / hello" row pins it
at `0.9999999999999999` (off by one ULP — see `scorer-default.json:115`),
which is empirically reasonable but the property is implicitly weaker
than "exactly 1.0".

Add a property test:

```go
func TestProp_Scorer_Identity(t *testing.T) {
    s := fuzzymatch.DefaultScorer()
    f := func(x string) bool {
        if x == "" {
            return true // empty-empty is a separate property
        }
        score := s.Score(x, x)
        return math.Abs(score - 1.0) < 1e-12
    }
    if err := quick.Check(f, nil); err != nil {
        t.Errorf("PropScorer_Identity: %v", err)
    }
}
```

Classification: **WEAK / IMPORTANT**. Identity is the most fundamental
invariant and the current single-algorithm test is too narrow.

---

#### TEST-05 (MISSING / IMPORTANT) — No symmetry property test for `Scorer.Score`

`Scorer.Score(a, b)` composed of symmetric algorithms (Levenshtein,
Jaro, Jaccard, Cosine, Sørensen-Dice — but NOT Tversky with α ≠ β,
NOT MongeElkan in its asymmetric form) should be symmetric.
`DefaultScorer` uses only symmetric algorithms, so the composite is
symmetric.

The Scorer-level property tests check determinism, range, weight-sum,
and finiteness — but NOT symmetry. A property test would catch any
future regression that introduces asymmetric ordering at the Scorer
boundary (e.g. a careless `applyNormalisation` change that processes
`a` differently from `b`).

```go
func TestProp_Scorer_Symmetric_DefaultScorer(t *testing.T) {
    s := fuzzymatch.DefaultScorer()
    f := func(a, b string) bool {
        return s.Score(a, b) == s.Score(b, a)
    }
    if err := quick.Check(f, nil); err != nil {
        t.Errorf("PropScorer_Symmetric: %v", err)
    }
}
```

Classification: **MISSING / IMPORTANT**. A documented invariant that
isn't gated by property test.

---

#### TEST-06 (MISSING / IMPORTANT) — No property test linking `Score` and `ScoreAll`

The class-12 BDD scenario (CONTEXT.md §7) and the
`TestScorer_ScoreAll_ValuesMatchPerAlgoCalls` unit test prove that
`ScoreAll`'s per-key values match a direct algorithm call. **But no
test proves the relationship**:

```
Score(a, b) == Σ ( algorithms[i].Weight * ScoreAll(a, b)[algorithms[i].ID] )
```

This is the algebraic identity that ties `Score` to `ScoreAll`. A
property test would catch any future change to `Score`'s reduction
loop that drifts from `ScoreAll`'s per-algorithm dispatch — for
example, if a future optimisation switches one of them to a different
normalisation gate.

```go
func TestProp_Scorer_Score_EqualsScoreAll_WeightedSum(t *testing.T) {
    s := fuzzymatch.DefaultScorer()
    algos := s.Algorithms()
    f := func(a, b string) bool {
        score := s.Score(a, b)
        all := s.ScoreAll(a, b)
        var acc float64
        for _, A := range algos {
            acc = acc + (A.Weight * all[A.ID])  // DET-06 explicit parens
        }
        return score == acc  // byte-exact: same algorithms, same order, same arithmetic
    }
    if err := quick.Check(f, nil); err != nil {
        t.Errorf("Score must equal weighted sum of ScoreAll: %v", err)
    }
}
```

Classification: **MISSING / IMPORTANT**. A load-bearing invariant
between two public surfaces that share an implementation path.

---

### Property-test depth

#### TEST-07 (FLAKY / BLOCKING) — `TestProp_Scorer_WeightSumOne` `uint16(65535)+1 = 0` overflow

**Files:** `scorer_test.go:820-853`.

This is the WR-01 finding from 08-REVIEW.md. The `u+1` expression is
typed `uint16` (because `u` is `uint16`), so `u = 65535` overflows to
`0`. The resulting `toPositive` returns 0.0, the option layer rejects
the zero weight via `ErrInvalidWeight`, and the property function
returns `false` — failing the property test.

Per-run probability for 3 inputs over 100 quick.Check iterations is
~0.46%. Sufficient to surface intermittently in CI nightly fuzz runs.

Fix:

```go
toPositive := func(u uint16) float64 {
    return (float64(u) + 1.0) / 65536.0 * 100.0  // never zero by construction
}
```

Classification: **FLAKY / BLOCKING** (a flake-once-per-200-runs
property test undermines the determinism guarantee the suite is
trying to assert).

---

#### TEST-08 (NIT) — `quick.Check` count never raised above default 100

The Phase 8 property tests use `quick.Check(f, nil)` which defaults to
100 iterations. Scorer property tests are cheap (no DP, no per-call
allocation that grows with input length) — the constants `MaxCount`
could be raised to 1000 via `&quick.Config{MaxCount: 1000}` without
material wall-time cost.

For algorithms with `quick.Check` cost roughly 50 µs/iter (Levenshtein
on random short strings), 1000 iters is 50 ms. For Scorer-level
properties at ~5-10 µs/iter, 1000 iters is 5-10 ms — well within the
"5 second per-task verify" budget from VALIDATION.md.

Recommend raising MaxCount to 1000 on every Scorer property test
that's not constrained by external generator cost.

Classification: **NIT**. Wouldn't catch a bug the existing tests
don't, but tightens the confidence interval on the existing
invariants.

---

### Concurrent test rigour

#### TEST-09 (NIT) — `TestScorer_ConcurrentSafety` runs once, not under `-count=N`

**Files:** `scorer_test.go:903-980`.

The test launches 100 goroutines × 3 methods (Score / ScoreAll /
Match). With `-race` enabled, this is sufficient to catch any naive
data race on the immutable `*Scorer`. The Scorer is immutable after
`NewScorer` returns, so there's no shared write at all — every field
is set once. Race detector should never fire.

The single-run pattern means a hypothetical TOCTOU bug (Time Of Check
vs Time Of Use) that depends on goroutine scheduling order would only
surface intermittently. The standard remediation is `-count=10` or
`-count=100` to multiply the goroutine permutations across runs.

For Phase 8 specifically, the *Scorer immutability is provable from
source-reading (no writes to receiver fields after construction), so
the empirical risk is low. The recommendation is **NOT blocking** —
flag for v1.0 release-checklist scrutiny.

Classification: **NIT**. Current coverage is sufficient given the
immutability invariant.

---

### Benchmark coverage

#### TEST-10 (MISSING / IMPORTANT) — Scorer benchmarks not in `bench.txt`

`bench.txt` (the committed benchmark baseline used for benchstat
regression detection) does NOT contain any `BenchmarkDefaultScorer_*`
entries. The 6 scorer benchmarks in `scorer_bench_test.go` ran during
benchmark-time validation but are absent from the regression baseline.

Per `go-testing-standards/SKILL.md` line 90: "`bench.txt` committed
per release; CI runs `benchstat` against the last tagged release with
> 10% regression failing the build."

Without scorer benchmarks in `bench.txt`, a future change that
regresses `DefaultScorer.Score` from 6 µs to 60 µs would NOT fail CI.
The benchmark exists in the source tree but is not gated.

Fix: regenerate `bench.txt` to include the scorer benchmarks before
the next milestone release — either by appending the current
benchmark output (10-run benchstat-compatible) or by re-running the
full `go test -bench=. -benchmem -count=10 ./...` suite.

Classification: **MISSING / IMPORTANT**. The benchmarks exist but
they aren't actually gating.

---

#### TEST-11 (MISSING / IMPORTANT) — Missing comparative benchmark variants

The 6 benchmarks cover the right dimensions (ASCII Short / Medium /
Long, Unicode Short, ScoreAll Short, Match Short) but do NOT include:

- **DefaultScorer vs MinimalScorer baseline** — a 1-algorithm
  Levenshtein-only Scorer would establish the per-algorithm overhead.
  `DefaultScorer.Score` is 6 algorithms; the difference between
  Levenshtein-only and 6-algorithm tells reviewers where the time
  goes.
- **WithNormalisation vs WithoutNormalisation** — the 08-REVIEW
  WR-04 finding identifies `Normalise(a, opts)` + `Normalise(b, opts)`
  as a major allocation contributor. A side-by-side benchmark would
  quantify the contribution and gate the eventual ASCII fast path.
- **Composite-with-1 vs Composite-with-6** — same as the first item
  with explicit naming.

These are NOT strictly required for Phase 8 acceptance (the wall-time
budget passes; the allocation budget is documented as a follow-up to
algorithm-performance-reviewer per WR-04), but the absence of
comparative benchmarks makes the allocation-reduction work (when it
happens) harder to validate.

Classification: **MISSING / IMPORTANT**. Track as a v1.0 release
prerequisite — without these comparators the WR-04 allocation budget
work has no visibility.

---

#### TEST-12 (NIT) — Match-benchmark sink gate is structurally dead-code

**Files:** `scorer_bench_test.go:171-173`.

This is the IN-05 finding from 08-REVIEW. The `if sink < -1` gate on
a non-negative `int` counter is provably false at compile time;
sufficiently aggressive optimisers could elide it. The current Go
compiler does not, so the benchmark works — but the locked-pattern
docstring promises this idiom defeats dead-code elimination, and this
particular instance is the weakest defence in the file.

Fix: use `sink == b.N` (compiler cannot fold `b.N`) or a bool sink
with `sink && !sink` (same idiom, harder to fold).

Classification: **NIT**. Won't fail today but documents an
imprecision in the locked benchmark pattern.

---

### Golden file rigour

#### TEST-13 (NIT) — Golden file 22 entries × 5 configs is solid; one edge case missing

**Files:** `testdata/golden/scorer-default.json`,
`scorer_golden_test.go:148-309`.

The 22-entry corpus covers the documented mandatory rows:

- 7 identifier-similarity reuse rows
- identity (`hello / hello`), both-empty (`/`), one-empty (`/hello`)
- Unicode NFC (`café / cafe`)
- phonetic divergent (`Smith / Schmidt`) on 3 different Scorers
- threshold-edge (`config / configs` and `abbreviation / abreviation`)
- all-different (`abc / xyz`)
- WithoutNormalisation variant (`XMLParser / xml_parser` and `café / cafe`)
- WithoutAlgorithm variant (`Smith / Schmidt` minus DM)
- Custom Levenshtein-only and raw-weights variants

Missing: **single-character pair** (`"a" / "b"` — exercises minimum
non-empty input). Single-character input is a documented edge case
in `go-testing-standards/SKILL.md` line 47, and several algorithms
(notably Q-Gram-Jaccard with default n=3) have a documented
"too-short-for-q-grams" path that returns 0.0.

Recommended addition (one row, +1 to corpus):

```go
entries = append(entries, makeScorerGoldenEntry(defaultS, "a", "b", "DefaultScorer"))
```

Verifying:
- `_metadata.generated_at` is correctly absent (verified
  `scorer-default.json:2-4` has only `phase` and `scorer_signature`).
- Byte-stability: `encoding/json` sorts map keys alphabetically on
  marshal, so the on-disk file is stable across runs.
- Cross-platform: gate is the CI matrix run on linux/amd64,
  linux/arm64, darwin/amd64, darwin/arm64, windows/amd64. Local
  darwin/arm64 verified passing.

Classification: **NIT**. Single-character is the only documented
edge case the golden file doesn't pin.

---

#### TEST-14 (NIT) — Golden file uses `0.9999999999999999` for identity case

**Files:** `scorer-default.json:10, 115, 130`.

The identity rows (`user_id / userId` after normalisation,
`hello / hello`, both-empty) all produce composite `0.9999999999999999`
— one ULP below exact `1.0`. This is mathematically correct (six
weighted contributions of `1.0/6.0` each: `6 × (1.0/6.0) = 1.0` in
infinite precision but `≈ 0.9999999999999999` in IEEE-754 because
`1.0/6.0` is not exactly representable).

The golden file faithfully pins this, which is the correct behaviour
for cross-platform determinism. But it surfaces a subtle property:
**the composite of N equal-weighted algorithms each returning 1.0 is
NOT byte-exactly 1.0 for N > 2**. Reviewers reading the golden file
might assume this is a regression.

Recommended: add a one-line comment in `scorer_golden_test.go`
explaining the `0.9999999999999999` is exact-arithmetic-correct for
the 6-algorithm equal-weight identity case.

Classification: **NIT**. Documentation hygiene, not a behavioural
defect.

---

### Fuzz testing

#### TEST-15 (MISSING / IMPORTANT) — No `Fuzz*` harness for `Scorer.Score` / `Match` / `ScoreAll`

`go-testing-standards/SKILL.md` line 45: "Every public function has a
fuzz harness in `fuzz_test.go`. Fuzz tests assert range bounds and
absence of panics".

Existing fuzz harnesses (per `testdata/fuzz/` corpus directories): 21
exist, one per algorithm + Normalise + Tokenise. **None for the
Scorer.** This is the largest gap in Phase 8's test surface.

The minimum Phase 8 fuzz set would be:

```go
// scorer_fuzz_test.go

func FuzzDefaultScorerScore(f *testing.F) {
    seeds := []struct{ a, b string }{
        {"user_id", "userId"}, {"hello", "hello"}, {"", ""}, {"", "x"},
        {"\xff\xfe", "abc"}, {"café", "cafe"}, {"Smith", "Schmidt"},
    }
    for _, s := range seeds { f.Add(s.a, s.b) }

    s := fuzzymatch.DefaultScorer()
    f.Fuzz(func(t *testing.T, a, b string) {
        score := s.Score(a, b)
        if math.IsNaN(score) { t.Errorf("Score = NaN for (%q, %q)", a, b) }
        if math.IsInf(score, 0) { t.Errorf("Score = Inf for (%q, %q)", a, b) }
        if score < 0.0 || score > 1.0 { t.Errorf("Score = %g out of [0,1] for (%q, %q)", score, a, b) }
    })
}

func FuzzDefaultScorerMatch(f *testing.F) {
    // ... mirrors above, asserts panic-free
}

func FuzzDefaultScorerScoreAll(f *testing.F) {
    // ... asserts every value in returned map satisfies range bound
}
```

Plus a `FuzzNewScorer` that randomises option construction (weight
bytes, algorithm IDs, threshold bytes) and asserts every error path
is hit consistently — though this is more complex because the option
type needs careful seed construction.

Classification: **MISSING / IMPORTANT**. Standard project pattern is
violated; the Phase 8 surface should have at least 3 `Fuzz*` harnesses
to match the per-algorithm convention.

---

### Edge cases

#### TEST-16 (PARTIAL / NIT) — Edge case sweep coverage

| Edge case | Tested? | Where |
|-----------|---------|-------|
| Both empty (`"", ""`)         | Yes | Golden row 9; `TestProp_Scorer_*` (random strings include `""`) |
| One empty (`"", "x"`)         | Yes | Golden row 10 |
| Identical input               | Yes | `TestScorer_Score_Identity` (Levenshtein), Golden row 8 (DefaultScorer) |
| Single-character pair         | No  | Not in unit, not in golden, not in BDD |
| Very-long string (1000+ char) | No  | Bench has 500-char; no test for 1000+ |
| Malformed UTF-8               | No  | No `\xff\xfe` style test on Scorer (algorithm fuzzers cover this individually) |
| WithoutNormalisation + Unicode | Yes | Golden row 21 (`café / cafe` with WithoutNormalisation) |
| NaN / Inf string-input        | N/A | Strings are byte-strings; no NaN concept |
| Identical pre-normalisation, different post-normalisation | Implicit | `TestScorer_WithoutNormalisation` covers it |

The two missing rows (single-character, very-long) are minor — both
algorithms exercise these in their own tests, and the composite
behaviour is the weighted average of well-tested components.

Classification: **NIT**. Edge-case coverage is sound; the gaps are
defensible.

---

## Per-Concern Status Summary

| Concern | Status | Notes |
|---------|--------|-------|
| Unit tests for With* options (happy + error) | **COVERED** | 100% coverage on all 12 options. |
| Sentinel error identity / wrap-Is | **COVERED** | `errors_test.go` exhaustive. |
| All 6 Scorer methods + 2 package functions | **COVERED** | Every public symbol exercised. |
| Stdlib-`testing`-only rule | **COVERED** | Verified — no testify in root tests. |
| ≥ 95% overall coverage | **PARTIALLY COVERED** | 90.2% (inherited from earlier-phase gaps). |
| ≥ 90% per-file coverage | **PARTIALLY COVERED** | `scorer.go` 100% all funcs except 92.6% NewScorer + 75% DefaultScorer (provably-unreachable branches). |
| 100% public API coverage | **COVERED** | Every public Scorer symbol at 100%. |
| Property: WeightSumOne | **COVERED but FLAKY** | TEST-07 overflow bug. |
| Property: ScoreInRange | **COVERED** | |
| Property: DeterministicAcrossRuns | **COVERED** | |
| Property: NoNaN_NoInf | **COVERED** | |
| Property: Identity (Score(x,x)=1.0) | **MISSING** | TEST-04 weak; needs DefaultScorer-level property. |
| Property: Symmetry | **MISSING** | TEST-05. |
| Property: Score = Σ Weight·ScoreAll | **MISSING** | TEST-06. |
| Concurrent safety | **COVERED** | 100 goroutines × 3 methods under -race. |
| Benchmarks in source | **COVERED** | 6 benchmarks. |
| Benchmarks in bench.txt | **MISSING** | TEST-10. |
| Allocation budget (≤ 8 short / medium) | **NOT MET** | 12 / 34 allocs/op respectively — pre-flagged WR-04. |
| Golden file 22+ entries, 5 configs | **COVERED** | Single-character pair missing (TEST-13 NIT). |
| Cross-platform determinism CI | **COVERED** | (Assumes CI matrix is wired — Manual-only per VALIDATION.md.) |
| Fuzz harness for Scorer.Score / Match / ScoreAll | **MISSING** | TEST-15. |
| BDD scenario count = 12 mandatory classes | **COVERED** | All 12 represented in scorer.feature. |
| BDD goleak gate | **COVERED** | `tests/bdd/bdd_test.go` (existing). |
| Empty-input edge cases | **COVERED** | Golden + property tests. |
| Identical-input edge case | **COVERED** | Unit + golden. |
| Unicode edge cases | **COVERED** | Golden rows 11, 21. |
| Malformed UTF-8 on Scorer | **MISSING** | TEST-15 would cover. |
| Single-character input | **MISSING** | TEST-13 NIT; TEST-15 would cover. |

## Recommendations

### Blocking (must be addressed before v1.0)

1. **Fix CR-01 (NaN in WithThreshold) and add TEST-01.** The
   documented contract (`docs/scorer.md:283`) says NaN is rejected;
   the implementation accepts it. The test suite should have caught
   it. Add the unit test once the fix lands.

2. **Fix CR-02 (Tversky α=0,β=0) and add TEST-02.** Same pattern — a
   documented invalid configuration that the test suite is silent on.

3. **Fix TEST-07 (uint16 overflow flake in
   `TestProp_Scorer_WeightSumOne`).** Flake-once-per-200-runs is too
   noisy for CI; the fix is one line.

### Important (should land in Phase 8 follow-up or early Phase 9)

4. **Add `FuzzDefaultScorerScore`, `FuzzDefaultScorerMatch`,
   `FuzzDefaultScorerScoreAll`** (TEST-15). The project's convention
   is "one Fuzz* per public function"; the Scorer surface violates
   this. Three small harnesses + seed corpora close the gap.

5. **Add Identity / Symmetry / `Score = Σ Weight·ScoreAll` property
   tests** (TEST-04, TEST-05, TEST-06). These are load-bearing
   mathematical invariants that the existing property suite leaves
   un-gated. The Symmetry test specifically would catch any future
   regression in the pre-normalisation gate.

6. **Regenerate `bench.txt` to include the 6 Scorer benchmarks**
   (TEST-10). Without this, the WR-04 allocation work has no
   regression baseline.

7. **Add comparative benchmarks** (TEST-11) — MinimalScorer baseline,
   With/WithoutNormalisation side-by-side, Composite-with-1 vs
   Composite-with-6. These quantify what WR-04's allocation
   reduction needs to achieve.

8. **Add `WithAlgorithm` rejection test for NaN/Inf weight** (TEST-03).
   The `weight <= 0` gate doesn't catch NaN/Inf; option layer
   silently accepts a poisoned weight.

9. **Document the `0.9999999999999999` identity composite** (TEST-14)
   with an inline comment in the golden test, so reviewers don't
   misread it as a regression.

### Nit (polish, not blocking)

10. **Raise `quick.Check` MaxCount to 1000** for Scorer property
    tests (TEST-08). Cheap; tightens confidence.

11. **Repeat `TestScorer_ConcurrentSafety` under `-count=10`** in CI
    (TEST-09). Belt-and-braces.

12. **Fix the Match-benchmark sink gate** (TEST-12). One-line cosmetic.

13. **Add a single-character pair row to scorer-default.json** (TEST-13).

14. **Address overall package coverage shortfall to ≥ 95%** by
    improving `double_metaphone.go` (`dmPrep` 58.8%,
    `DoubleMetaphoneKeys` 62.0%) and `nysiis.go` (`NYSIISCode` 72.6%)
    test coverage. These are inherited gaps not Phase 8 work, but
    Phase 8 inherits the failing meta-test target.

## Recommendation

**Coverage gaps to address.** The Phase 8 test surface is high-quality
overall — 100% on every public Scorer symbol, exhaustive option-layer
testing, a cross-platform golden file, BDD scenarios for every one of
the 12 mandatory classes, and a goroutine-leak gate. The three
blocking items (TEST-01, TEST-02, TEST-07) are small fixes that close
real defects already flagged in 08-REVIEW. The fuzz-harness gap
(TEST-15) is the only non-trivial new work and is bounded — three
small harnesses, mirroring the well-established per-algorithm fuzz
pattern.

**Not ready for a milestone release** until TEST-01, TEST-02, and
TEST-07 land (these are blocking per the code-review report and
correspond to silent-malfunction defects), and the Scorer benchmarks
are written into `bench.txt` so the regression gate is real (TEST-10).

After those four items land, Phase 8's test surface meets the project's
quality bar and the milestone-release path is clear.

_Analysed: 2026-05-17_
_Analyst: Claude (test-analyst)_
