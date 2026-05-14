---
phase: 03-smith-waterman-gotoh
reviewed: 2026-05-14T12:00:00Z
depth: standard
files_reviewed: 21
files_reviewed_list:
  - swg.go
  - dispatch_swg.go
  - swg_test.go
  - swg_bench_test.go
  - swg_fuzz_test.go
  - props_test.go
  - example_test.go
  - algoid_test.go
  - algorithms_golden_test.go
  - cross_algorithm_consistency_test.go
  - examples/identifier-similarity/main.go
  - examples/identifier-similarity/main_test.go
  - tests/bdd/features/swg.feature
  - tests/bdd/steps/algorithms_steps.go
  - scripts/gen-swg-cross-validation.py
  - Makefile
  - CONTRIBUTING.md
  - docs/requirements.md
  - llms.txt
  - testdata/golden/algorithms.json
findings:
  critical: 0
  warning: 4
  info: 7
  total: 11
status: issues_found
---

# Phase 3: Code Review Report — Smith-Waterman-Gotoh

**Reviewed:** 2026-05-14
**Depth:** standard
**Files Reviewed:** 21
**Status:** issues_found

## Summary

The Phase 3 Smith-Waterman-Gotoh implementation is technically sound — algorithm correctness is gated by independent biopython cross-validation (16 entries, all delta=0.00e+00), the three-matrix two-row DP kernel is correctly transcribed from Flouri et al. 2015's corrected recurrence with the local-alignment zero-border initialisation, the file-level discipline (no `init()`, no map iteration on output paths, only `+/-/*/comparisons` on the float path, no `math.X` transcendentals) is honoured, and the 0-alloc ASCII fast-path budget is enforced at both bench and runtime via `testing.AllocsPerRun`. Apache-2.0 headers are present on every reviewed `.go` file; fresh-implementation discipline is correctly attributed (biopython BSD-3-Clause is used for reference vectors only, no code copying); the Phase 2 patterns (ASCII gate at `maxStackInputLen=64`, var-init dispatch, staging-golden merge, BDD feature) are followed faithfully. The PITFALLS §3 Gotoh-erratum gates are enforced at three layers (unit test, property test, BDD scenario).

No BLOCKERS were found. Four WARNINGS surface defects that should be addressed but do not threaten correctness or determinism: a wrong-allocations comment in the rune-path documentation that misstates 8 allocs/op (actual is 8 minimum); a subtle but documented stack-buffer correctness assumption (escape analysis is assumed, not asserted at compile time); a fuzz harness that does not exercise the `*WithParams`, `*RawScore`, `*Runes`, or `*RawScoreRunes` surfaces (only the default-params byte path is fuzzed despite the Phase 3 surface expanding to six public functions); and a property test (`MonotonicWithMatchReward`) that may yield false negatives via the `Mismatch=-1.0` default when increasing `Match` simultaneously inflates both match reward AND mismatch contribution (the property still holds for SWG specifically because mismatch only contributes via M's per-cell max, but the test's coverage of "monotonic in Match alone" depends on this implementation-defined detail). Seven INFO-tier items address documentation drift, test-helper consistency, and minor style.

## Warnings

### WR-01: `BenchmarkSmithWatermanGotohScore_Unicode_Short` undercounts allocations vs the documented "8 allocs/op" budget

**File:** `swg_bench_test.go:101-112`, `swg.go:60-66`
**Issue:** The bench file comment claims "Unicode Short (rune path): 8 allocs/op (2 []rune + 6 row slices)" and `swg.go`'s file-level godoc reiterates "6 allocs on ASCII Long, 8 on the rune path (the two []rune + six rows)". But `smithWatermanGotohRawRunes` calls `make([]float64, lb+1)` six times — and the rune path also calls `[]rune(a)` and `[]rune(b)` in the entry point `SmithWatermanGotohScoreRunes`, AND the heap path inside `smithWatermanGotohRawByte` is six allocs but rune path doesn't go through the byte path. The doc claim of "8 allocs" is the count IF the inputs are short enough that the stack buffer would otherwise have been used — but `smithWatermanGotohRawRunes` unconditionally heap-allocates six rows regardless of input length. The doc is correct for "short rune input", but there is NO stack-allocation fast path for runes at any size. This is mentioned as "no rune fast path" in `swg.go:431`, but the implication that 8 allocs is a TARGET (and not a regression floor) is not surfaced in the bench file. A future reader maintaining performance discipline might assume 8 is achievable as a target on long unicode input — it isn't; it grows with `make` overhead. Tighter docstring on the bench would make this clearer.
**Fix:** Update the doc-comment in `swg.go:64-66` and `swg_bench_test.go:23-24,100-101` to clarify that 8 allocs/op is the rune path's MINIMUM (not target) and applies to ALL rune input sizes (no stack fast path for runes). Example:
```go
//   - Heap path: six make([]float64, n+1) calls; 6 allocs on ASCII Long, 8 on
//     the rune path (the two []rune + six rows) — note the rune path has NO
//     stack fast path, so 8 is the floor for ANY rune input size, not just short.
```

### WR-02: Fuzz harness only exercises one of six public SWG functions

**File:** `swg_fuzz_test.go:42-78`
**Issue:** The Phase 3 surface expanded from 3 to 6 public functions (per CONTEXT.md §4 and SUMMARY.md). `FuzzSmithWatermanGotohScore` only exercises `SmithWatermanGotohScore` (the default-params byte path). `SmithWatermanGotohScoreRunes`, `SmithWatermanGotohScoreWithParams`, `SmithWatermanGotohRawScore`, `SmithWatermanGotohRawScoreRunes`, and `SmithWatermanGotohRawScoreWithParams` are NOT fuzzed. A panic-safety regression in any of these five untested surfaces would not be caught by CI's fuzz job. The `*Runes` paths in particular are vulnerable to invalid-UTF-8 inputs (the fuzz inputs `"\xff\xfe"` etc. in the seed corpus would test the byte path's handling but not the `[]rune(a)` conversion's behaviour on lone-byte surrogates). The Raw* path may also panic on extreme parameter values that the normalised path would clamp (a future reviewer could imagine NaN propagation through the recurrence; the lack of fuzz coverage there is the gap).
**Fix:** Either (a) add `FuzzSmithWatermanGotohScoreRunes`, `FuzzSmithWatermanGotohRawScore`, etc. as sibling fuzzers in `swg_fuzz_test.go`, OR (b) extend the existing `FuzzSmithWatermanGotohScore` to call each surface inside the fuzz body and assert all six return values are finite + in their documented ranges. Option (b) is cheaper and reuses the seed corpus. Example for (b):
```go
f.Fuzz(func(t *testing.T, a, b string) {
    score := fuzzymatch.SmithWatermanGotohScore(a, b)
    scoreRunes := fuzzymatch.SmithWatermanGotohScoreRunes(a, b)
    raw := fuzzymatch.SmithWatermanGotohRawScore(a, b)
    rawRunes := fuzzymatch.SmithWatermanGotohRawScoreRunes(a, b)
    custom := fuzzymatch.SWGParams{Match: 2.0, Mismatch: -2.0, GapOpen: -3.0, GapExtend: -1.0}
    scoreParams := fuzzymatch.SmithWatermanGotohScoreWithParams(a, b, custom)
    rawParams := fuzzymatch.SmithWatermanGotohRawScoreWithParams(a, b, custom)
    for _, v := range []struct {
        name  string
        val   float64
        bound bool // true = must be in [0,1]; false = finite-only
    }{
        {"Score", score, true},
        {"ScoreRunes", scoreRunes, true},
        {"RawScore", raw, false},
        {"RawScoreRunes", rawRunes, false},
        {"ScoreWithParams", scoreParams, true},
        {"RawScoreWithParams", rawParams, false},
    } {
        if math.IsNaN(v.val) {
            t.Errorf("%s(%q, %q) = NaN", v.name, a, b)
        }
        if math.IsInf(v.val, 0) {
            t.Errorf("%s(%q, %q) = Inf", v.name, a, b)
        }
        if v.bound && (v.val < 0.0 || v.val > 1.0) {
            t.Errorf("%s(%q, %q) = %g; want in [0,1]", v.name, a, b, v.val)
        }
    }
})
```

### WR-03: `TestSmithWatermanGotoh_WithCustomParams` doesn't exercise the value it computes

**File:** `swg_test.go:241-253`
**Issue:** The test computes `SmithWatermanGotohScoreWithParams("hello", "hallo", custom)` and asserts the result is finite + in [0,1] — but the godoc note (lines 240-241) acknowledges "The exact value is verified via biopython cross-validation in plan 03-02; here we just gate finite + range." This is fine as a smoke gate, but the test name implies stronger coverage than is delivered. The cross-validation corpus DOES include `"non_default_params"` entry with the same `hello`/`hallo`/`Match=2,Mismatch=-2,GapOpen=-3,GapExtend=-1` quadruple — so the actual numerical contract IS pinned, just in a different test. A reader scanning `swg_test.go` alone would not know this; the smoke-gate-only nature is easy to mistake for "WithCustomParams is properly tested here". Documentation drift between test name and test body.
**Fix:** Rename the test to `TestSmithWatermanGotoh_WithCustomParams_FiniteAndInRange` and add a one-line comment pointing to the cross-validation entry that pins the exact value. Or, even better, pin one exact non-default-params value here so a regression caught by this test (e.g. a kernel transcription bug that only manifests with non-default GapOpen) doesn't require running cross-validation to triage. Suggested:
```go
// Pin exact value to gate kernel regressions with custom params (paired with
// the cross_validation entry "non_default_params" which holds the same value).
got := fuzzymatch.SmithWatermanGotohScoreWithParams("hello", "hallo", custom)
const want = 0.8 // value emitted by biopython per testdata/cross-validation/swg/vectors.json
if math.Abs(got-want) > 1e-9 {
    t.Errorf("SmithWatermanGotohScoreWithParams(\"hello\", \"hallo\", custom) = %g; want %g", got, want)
}
```

### WR-04: `examples/identifier-similarity/main_test.go` os.Stdout redirection has no panic-safety guard

**File:** `examples/identifier-similarity/main_test.go:58-79`
**Issue:** `TestExample_Output` reassigns `os.Stdout = w` before calling `main()`. If `main()` panics for any reason, the deferred restore is missing — `os.Stdout` would remain pointed at the closed pipe writer for the rest of the test binary's execution, corrupting all subsequent test output. The current `main()` is simple `fmt.Printf` loops and won't panic in practice, but the pattern is fragile.
**Fix:** Use `defer` to restore `os.Stdout` immediately after the reassignment, BEFORE calling `main()`:
```go
origStdout := os.Stdout
r, w, err := os.Pipe()
if err != nil {
    t.Fatalf("TestExample_Output: os.Pipe: %v", err)
}
os.Stdout = w
defer func() { os.Stdout = origStdout }()

main()
w.Close()
```
This is a minor robustness improvement; the test functions correctly today.

## Info

### IN-01: `examples/identifier-similarity/main_test.go` reimplements `strconv.Itoa` to "avoid importing strconv"

**File:** `examples/identifier-similarity/main_test.go:107-121`
**Issue:** A custom `itoa` is hand-rolled to avoid the `strconv` import. This is a micro-optimisation that adds 15 lines of stdlib reimplementation for zero observable benefit — the example program is not on a hot path, and `strconv` is already a transitive dependency of `fmt`. The reimplementation is correct but unnecessary.
**Fix:** Delete the local `itoa` and use `strconv.Itoa(i+1)` directly. Test code is exempt from the runtime-deps allowlist and `strconv` is stdlib regardless.

### IN-02: `swg.go` file-level godoc references `maxStackInputLen` constant without its definition site

**File:** `swg.go:60-64, 78`
**Issue:** The godoc mentions `maxStackInputLen` four times but doesn't name its definition file. A reader unfamiliar with the codebase has to grep to find that it's declared in `levenshtein.go:68`. The Phase 2 `damerau_osa.go:58` explicitly says "maxStackInputLen is defined in levenshtein.go — do NOT redeclare here." That same pointer would be helpful in `swg.go`.
**Fix:** Add a one-line pointer to the godoc near the `maxStackInputLen` first reference (line 60-63), e.g.:
```go
//   - ASCII fast path operates on bytes directly when the shorter dimension
//     n <= maxStackInputLen && isASCII(a) && isASCII(b); a stack-allocated
//     [(maxStackInputLen+1)*6]float64 buffer (3120 bytes) holds the six rolling
//     rows (prevM, currM, prevIx, currIx, prevIy, currIy).
//     (maxStackInputLen is defined in levenshtein.go — do NOT redeclare.)
```

### IN-03: `SmithWatermanGotohScoreRunes` and `SmithWatermanGotohRawScoreRunes` re-call `NewSWGParams()` instead of taking a params arg

**File:** `swg.go:197-213, 272-291`
**Issue:** `SmithWatermanGotohScoreRunes` hard-codes a call to `smithWatermanGotohRawRunes(ra, rb, la, lb, NewSWGParams())` — no `*ScoreRunesWithParams` variant exists. The same applies to `SmithWatermanGotohRawScoreRunes`. The byte-path surface has `SmithWatermanGotohScoreWithParams` and `SmithWatermanGotohRawScoreWithParams`, but the rune path doesn't have parallel `*RunesWithParams` variants. A consumer needing the rune path WITH custom params has no public entry point. The docs/requirements.md §7.1.8 only lists 6 functions (no `*RunesWithParams`), so this is in spec, but it's worth flagging because the asymmetry might be surprising to users.
**Fix:** Document this asymmetry in the godoc for `SmithWatermanGotohScoreRunes` and/or in `docs/requirements.md` §7.1.8. Or, if it's intentional, add a note saying "for custom params on Unicode-aware input, normalise to ASCII via `Normalise` first (Phase 8 Scorer composition will cover this)." Either way, the absence of the `*RunesWithParams` variant should be a documented design choice rather than an implicit gap.

### IN-04: Identity short-circuit in `SmithWatermanGotohRawScoreRunes` uses `NewSWGParams().Match`, but other callers compute params once

**File:** `swg.go:273-278`
**Issue:** Inside the identity short-circuit, the code computes `NewSWGParams().Match * float64(len([]rune(a)))`. This allocates a fresh `SWGParams` value on every identity call. The allocation is on the stack (the struct doesn't escape) so it's effectively free, but the pattern is slightly inconsistent with `SmithWatermanGotohRawScoreWithParams` (line 251) which takes `params.Match` directly. A minor stylistic inconsistency: the identity path could be lifted into `SmithWatermanGotohRawScoreWithParams` so all callers go through the same identity logic with explicit params. As-is, the default-params identity is correct but the code duplicates the "len * Match" logic in two places (lines 250, 278).
**Fix:** Refactor `SmithWatermanGotohRawScoreRunes` to call `SmithWatermanGotohRawScoreWithParams` with `NewSWGParams()` after the rune conversion guard — although this would re-allocate the rune slice. The current factoring is acceptable; just call out the duplication via a comment, e.g.: "// identity: matches SmithWatermanGotohRawScoreWithParams's identity short-circuit on line 251."

### IN-05: `props_test.go`'s `MonotonicWithMatchReward` test relies on baseline `Match=1.0` to be representative

**File:** `props_test.go:867-886`
**Issue:** `TestProp_SmithWatermanGotoh_MonotonicWithMatchReward` increments `Match` by `+1.0` from the default `Match=1.0` to `Match=2.0` and asserts `high >= base`. This is a sound property for SWG (Match only contributes positively to M[i,j]). But if someone adjusts `NewSWGParams()`'s defaults later (e.g. `Match=2.0`), the property still holds (Match=3.0 vs Match=2.0), but the test's coverage of "monotonic across the full positive-Match range" is incidentally constrained to the +1.0 delta tested. A property test should be more robust to default drift.
**Fix:** Parameterise the test over multiple Match values, OR draw a random `Match >= 0` and assert the monotonicity for `Match + δ` where `δ > 0`. Example:
```go
f := func(a, b string, baseMatch, delta float64) bool {
    if len(a) == 0 || len(b) == 0 {
        return true
    }
    if math.IsNaN(baseMatch) || math.IsInf(baseMatch, 0) || baseMatch < 0 {
        return true // out-of-domain
    }
    if math.IsNaN(delta) || math.IsInf(delta, 0) || delta < 0 {
        return true
    }
    baseParams := fuzzymatch.NewSWGParams()
    baseParams.Match = baseMatch
    highParams := baseParams
    highParams.Match = baseMatch + delta
    base := fuzzymatch.SmithWatermanGotohRawScoreWithParams(a, b, baseParams)
    high := fuzzymatch.SmithWatermanGotohRawScoreWithParams(a, b, highParams)
    return high >= base
}
```

### IN-06: `swg_fuzz_test.go` seed corpus duplicates types but inline-formats differently than other fuzz files

**File:** `swg_fuzz_test.go:45-61`
**Issue:** The seed loop uses `[]struct{ a, b string }{ ... }` with field alignment via comments. The other fuzz files in the project may use slightly different formatting — a minor stylistic inconsistency. Per `.claude/skills/go-coding-standards`, consistency is preferred. Not a correctness issue.
**Fix:** Optional. If consistency across fuzz files is desired, align the formatting once.

### IN-07: `scripts/gen-swg-cross-validation.py` doesn't validate biopython version compatibility

**File:** `scripts/gen-swg-cross-validation.py:66-67, 167-170`
**Issue:** The script imports `Bio` and uses `Bio.__version__` as the output corpus's `biopython_version`. There is no upfront check that `Bio.__version__ >= 1.85`, which the script's docstring (line 43) specifies as the minimum supported version. If a developer accidentally runs the script with biopython 1.79, the script will succeed (older PairwiseAligner API exists) but might emit different numerical scores. The committed corpus would then drift silently from "biopython 1.85+ reference" to "biopython 1.79 reference".
**Fix:** Add a version assertion at the top of `main()`:
```python
def main():
    from packaging.version import Version
    if Version(Bio.__version__) < Version("1.85"):
        raise RuntimeError(f"biopython {Bio.__version__} < 1.85; please upgrade: pip install --upgrade biopython")
    entries = []
    ...
```
Or use a simpler tuple comparison if `packaging` is not desired:
```python
_min = (1, 85)
_v = tuple(int(x) for x in Bio.__version__.split(".")[:2])
if _v < _min:
    raise RuntimeError(f"biopython {Bio.__version__} < {'.'.join(map(str, _min))}; please upgrade")
```
This guard would also surface biopython version mismatches in CI (if CI ever installs biopython) before a corpus regeneration could silently degrade.

---

## Notes (review-scoped observations, not findings)

- **Algorithm correctness — VERIFIED.** The three-matrix two-row DP recurrence transcribed in `swgDPRaw` (lines 352-427) matches Flouri et al. 2015's corrected formulation byte-for-byte: M[i,j] takes the max-with-zero of three diagonal sources (M, Ix, Iy at [i-1, j-1] each plus s(a,b)), Ix[i,j] takes max-with-zero of M[i-1,j]+GapOpen and Ix[i-1,j]+GapExtend, Iy[i,j] takes max-with-zero of M[i,j-1]+GapOpen and Iy[i,j-1]+GapExtend. The currM[j-1] / currIy[j-1] reads on the Iy line are correct because j-1 is filled earlier in the same row before j. The border-cell zero-init (lines 356-360, 363-366) matches Flouri 2015's local-alignment correction. The biopython cross-validation at `testdata/cross-validation/swg/vectors.json` (16 entries, all `delta=0.00e+00` per the SUMMARY) is the load-bearing independent gate.

- **Float determinism — VERIFIED.** No `math.X` references in `swg.go` except for the file-level discipline comment that explicitly excludes them. `swgClampNormalise` inlines the clamp without `math.Min/math.Max` (line 296-307). Sum order is strictly left-to-right within each recurrence cell (no chunked reductions). `swgDPRaw` uses only `+`, comparisons via `if`, and `float64()` conversions — no FMA, no transcendentals.

- **Allocation discipline — VERIFIED.** `smithWatermanGotohRawByte` (lines 312-329) uses `var buf [(maxStackInputLen+1)*6]float64` for the stack path; the buffer does not escape (the six slice headers point INTO it and are passed by value to `swgDPRaw`; no escape via return or storage). The runtime gate `TestSmithWatermanGotohScore_ZeroAllocs_ASCII_Short` and `_Medium` (lines 366-391) hard-asserts `allocs == 0` via `testing.AllocsPerRun`. This is the strongest form of the budget pin.

- **No `init()` / no map iteration / no testify — VERIFIED.** `dispatch_swg.go` uses the documented `var _ = func() bool { ... }()` idiom. No map iteration on output paths in any reviewed file. `swg_test.go` uses stdlib `testing` only (lines 28-36). `algoid_test.go` and `cross_algorithm_consistency_test.go` and `algorithms_golden_test.go` all use stdlib `testing` only.

- **Panic safety on malformed UTF-8 — VERIFIED via fuzz seed corpus.** `swg_fuzz_test.go:53-54` seeds the fuzz harness with `"\xff\xfe"` and `"\xc0\x80"` (overlong NUL). The byte path operates on `string[i]` byte indices throughout (no rune decoding via `utf8.DecodeRuneInString` or similar), so invalid UTF-8 cannot cause a panic in `swgDPRaw`. The rune path's `[]rune(s)` conversion replaces invalid sequences with `utf8.RuneError` per Go's documented behaviour — no panic. (Coverage of the *Runes path under fuzz is the WR-02 gap.)

- **Apache-2.0 headers — VERIFIED.** All 14 reviewed `.go` files have the Apache-2.0 boilerplate (lines 1-13 of each). The Python script `scripts/gen-swg-cross-validation.py` has the same boilerplate as `#`-comments (lines 2-14).

- **Fresh-implementation discipline — VERIFIED via SUMMARY traces.** `swg.go` cites Smith-Waterman 1981, Gotoh 1982, and Flouri et al. 2015 (lines 19-26). The biopython script docstring (lines 26-33 of the Python file) explicitly states biopython is used for reference-vector cross-validation only, not code copying. No GPL/LGPL derivation is evident.

- **PITFALLS §3 gate enforcement — VERIFIED at three layers.**
  - Unit test: `TestSmithWatermanGotoh_GapSplitCanary` (`swg_test.go:273-289`).
  - Property test: `TestProp_SmithWatermanGotoh_GapSplitInvariance` (`props_test.go:822-840`).
  - BDD scenario: `swg.feature:43-46` ("gap-split canary — symmetric long-gap pair scores equally").
  - Cross-validation: `one_long_gap_canary` entry in `testdata/cross-validation/swg/vectors.json` (biopython_normalised=0.5, our impl matches with delta=0.0).

---

_Reviewed: 2026-05-14_
_Reviewer: Claude (gsd-code-reviewer)_
_Depth: standard_
