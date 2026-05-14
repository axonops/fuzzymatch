---
phase: 02-core-character-algorithms-six
reviewed: 2026-05-14T00:00:00Z
depth: standard
files_reviewed: 42
files_reviewed_list:
  - algoid_test.go
  - algorithms_golden_test.go
  - cross_algorithm_consistency_test.go
  - damerau_full.go
  - damerau_full_bench_test.go
  - damerau_full_discriminator_test.go
  - damerau_full_fuzz_test.go
  - damerau_full_test.go
  - damerau_osa.go
  - damerau_osa_bench_test.go
  - damerau_osa_discriminator_test.go
  - damerau_osa_fuzz_test.go
  - damerau_osa_test.go
  - dispatch_damerau_full.go
  - dispatch_damerau_osa.go
  - dispatch_hamming.go
  - dispatch_jaro.go
  - dispatch_jarowinkler.go
  - dispatch_levenshtein.go
  - example_test.go
  - examples/identifier-similarity/main.go
  - examples/identifier-similarity/main_test.go
  - export_test.go
  - hamming.go
  - hamming_bench_test.go
  - hamming_fuzz_test.go
  - hamming_test.go
  - jaro.go
  - jaro_bench_test.go
  - jaro_fuzz_test.go
  - jaro_test.go
  - jarowinkler.go
  - jarowinkler_bench_test.go
  - jarowinkler_fuzz_test.go
  - jarowinkler_test.go
  - levenshtein.go
  - levenshtein_bench_test.go
  - levenshtein_fuzz_test.go
  - levenshtein_test.go
  - props_test.go
  - tests/bdd/bdd_test.go
  - tests/bdd/steps/algorithms_steps.go
findings:
  critical: 0
  warning: 4
  info: 7
  total: 11
status: issues_found
---

# Phase 02: Code Review Report

**Reviewed:** 2026-05-14
**Depth:** standard
**Files Reviewed:** 42
**Status:** issues_found

## Summary

Phase 02 ships six character-based similarity algorithms (Levenshtein,
Hamming, Jaro, Jaro-Winkler, Damerau-Levenshtein OSA, Damerau-Levenshtein
Full) with a uniform structure: per-algorithm file with primary-source
citation in the file godoc, dispatch registration via the
`var _ = func() bool { ... }()` idiom, unit tests with literature reference
vectors, property tests for range bounds / identity / symmetry / NaN-Inf-(-0)
guards, fuzz harnesses with invalid-UTF-8 seeds, allocation-aware benchmarks,
golden-file determinism gates, BDD step definitions, and a runnable
side-by-side example program.

**Correctness.** The four DP-based algorithms (Levenshtein, DL-OSA, DL-Full,
Hamming) and the two Jaro-family algorithms produce the documented reference
values on the canonical Winkler-1990 / Wagner-Fischer-1974 / Boytsov-2011 /
Lowrance-Wagner-1975 / Hamming-1950 / Jaro-1989 / Damerau-1964 vectors. The
OSA-vs-Full discriminating vector ("ca"/"abc" → 3 vs 2) is asserted in three
separate places (per-algorithm stub, per-algorithm full test,
cross_algorithm_consistency_test.go) and is the load-bearing proof that the
two recurrences have not collapsed into each other. The MARTHA/MARHTA Jaro
walk-through reproduces 0.9444... as expected. The DL-Full Lowrance-Wagner
table indexing (paper `D[i,j]` → table `d[(i+1)*stride+(j+1)]`,
paper `D[l-1,k-1]` → `d[l*stride+k]`) is consistent. The DL-OSA three-row
rolling window correctly preserves `D[i-2, j-2]` in the `prevprev` slot
across rotations.

**Determinism.** No `math.Pow`/`Log`/`Exp`/`FMA` calls. Only `+`, `-`, `*`,
`/`, and `float64()` conversions on the hot paths. Maps appear only as
internal point-lookup tables in the DL-Full rune path (`map[rune]int da`) —
never iterated for output. Float formulas are explicitly parenthesised and
evaluate left-to-right per DET-06.

**Performance.** ASCII fast paths gate stack-buffer allocation with
`isASCII(a) && isASCII(b)` (DL-OSA, Jaro) or simply on length (Levenshtein —
see WR-01). Zero-alloc behaviour is exercised at test time via
`testing.AllocsPerRun(100, ...)` for five of the six algorithms; DL-Full's
0-alloc test is `t.Skipf`'d with a documented v1.x follow-up note (the full
DP table is heap-allocated for all inputs).

**API hygiene.** Every public function has a godoc comment that opens with
the function name, lists edge cases, describes byte-vs-rune semantics, and
cites reference vectors where applicable. File-level godocs cite primary
academic sources and include the recurrence relation.

**Findings below** are 4 WARNING-level issues (inconsistent ASCII gating
between Levenshtein and DL-OSA, an unused defensive guard in DL-Full, a
subtle assertion gap in the Jaro symmetry property test for rune inputs,
and a doc/test gap in the OSA triangle-inequality property test) and 7
INFO-level items (small inefficiencies, minor documentation drift, BDD
regex limitations). No BLOCKERs were found.

## Warnings

### WR-01: Levenshtein stack-buffer path missing ASCII gate; inconsistent with DL-OSA

**File:** `levenshtein.go:104-107`
**Issue:** `LevenshteinDistance` selects the stack-allocated DP buffer
purely on `if n <= maxStackInputLen`, with no `isASCII(a) && isASCII(b)`
guard. By contrast, `damerau_osa.go:116` requires both
`n <= maxStackInputLen && isASCII(a) && isASCII(b)` before taking the same
stack path. Both algorithms are documented as byte-level
(`"This function operates on bytes"`) and the byte DP works correctly on any
byte content, so this is not a correctness defect — but the **inconsistency
hides intent**: a future reader cannot tell whether the missing ASCII guard
in Levenshtein is deliberate or a transcription oversight. If the design
intent is that DP byte-level algorithms always take the stack path for
shorter dimension ≤ 64 (regardless of content), the DL-OSA guard is
over-restrictive and should be removed. If the design intent is to gate on
ASCII (perhaps to reserve the stack buffer for ASCII-only inputs to keep
the rune path on its own deterministic alloc profile), Levenshtein should
acquire the same guard. The Phase 2 patterns document
(`02-PATTERNS.md`) does not resolve this — pick one and apply it uniformly
to all four DP-based algorithms.

**Fix:** Choose one of:
1. Add `isASCII(a) && isASCII(b)` to `levenshtein.go:104`. This matches
   `damerau_osa.go:116` and produces a uniform "ASCII fast path" idiom
   across all DP-based algorithms:
   ```go
   if n <= maxStackInputLen && isASCII(a) && isASCII(b) {
       var buf [(maxStackInputLen + 1) * 2]int
       return levenshteinDP(a, b, m, n, buf[:n+1], buf[n+1:2*(n+1)])
   }
   ```
2. Remove the `isASCII(...)` guards from `damerau_osa.go:116`. The byte DP
   works on any bytes; the ASCII gate is unnecessary for correctness and
   denies stack-buffer benefits to short non-ASCII inputs.

Document the chosen pattern in `02-PATTERNS.md` so wave-3 phases (Q-Gram,
token, phonetic) inherit a single idiom.

### WR-02: DL-Full transposition guard `l > 0 && k > 0` is redundant given the phantom-sentinel design — pick one

**File:** `damerau_full.go:302, damerau_full.go:397`
**Issue:** The DL-Full implementation deploys **two** independent
protections against the "never-seen" character case in the transposition
term:

1. **Phantom sentinel rows/columns** at table index 0, filled with
   `bigVal = m + n` (lines 238-244). The paper's `D[-1, *]` and `D[*, -1]`
   map to table row 0 and column 0, both bigVal. If `l == 0` ("never seen
   in a"), then paper `D[l-1, k-1] = D[-1, k-1]` → table `d[0*stride + k]
   = bigVal`. The transposition cost becomes
   `bigVal + (i - 0 - 1) + 1 + (j - k - 1) ≥ m + n`, which exceeds any
   real edit distance ≤ `max(m, n) ≤ m + n` and so can never be selected.

2. **Explicit guard** `if l > 0 && k > 0 { ... }` (line 302) which skips
   the transposition cost entirely.

Either mechanism alone is sufficient. Having both makes the recurrence
harder to verify against Lowrance-Wagner 1975 — a reader cross-checking
the paper sees the sentinel-only or guard-only form, never the union. The
redundancy also obscures whether the sentinel is load-bearing (and would
break if removed) or vestigial.

**Fix:** Pick one mechanism and remove the other.

Option A — keep the explicit guard (clearer intent, no reliance on
arithmetic sentinel domination), drop the phantom rows:
```go
// Remove lines 238-244 (sentinel initialisation).
// The if l > 0 && k > 0 check (line 302) already prevents the
// out-of-bounds-in-spirit access.
```
This also lets `bigVal` be deleted (it's only read by the sentinel rows).

Option B — keep the sentinel approach (closer to Lowrance-Wagner's
implicit ∞ formulation), drop the guard:
```go
// Line 302: remove the if l > 0 && k > 0 wrapper.
// The sentinel rows naturally produce a large cost for never-seen chars.
trans := d[l*stride+k] + (i - l - 1) + 1 + (j - k - 1)
if trans < v {
    v = trans
}
```

Option A is preferred for readability; Option B is closer to the paper.
Either way, leave a one-line godoc explaining the choice. Apply the same
change to both `damerauFullDP` (byte path) and
`damerauFullDistanceRuneSlices` (rune path) for symmetry.

### WR-03: Property test `TestProp_JaroScore_Symmetric` runs only on byte inputs; rune-path symmetry is untested as a property

**File:** `props_test.go:260-267`
**Issue:** `TestProp_JaroScore_Symmetric` asserts
`fuzzymatch.JaroScore(a, b) == fuzzymatch.JaroScore(b, a)` for `quick.Check`-
generated random strings. `quick.Check`'s default generator does produce
multi-byte UTF-8, so this is partial coverage of the rune-relevant cases —
but it tests `JaroScore` (byte path), not `JaroScoreRunes` (rune path).
There is **no property test for `JaroScoreRunes` symmetry**. The same gap
applies to `LevenshteinScoreRunes`, `HammingScoreRunes`,
`DamerauLevenshteinOSAScoreRunes`, `DamerauLevenshteinFullScoreRunes`, and
`JaroWinklerScoreRunes`: none have a `TestProp_*Score_Symmetric` test that
calls the `*Runes` variant.

This is a gap because the rune paths use separate code (e.g.
`jaroRunes` vs `jaroBytes`, `levenshteinDistanceRuneSlices` vs
`levenshteinDP`) — a bug in rune-path symmetry would not be caught by the
byte-path property test alone. The hardcoded rune tests
(`TestJaroScoreRunes_EdgeCases` etc.) test fixed pairs but do not
quick.Check-fuzz the symmetry property.

**Fix:** Add a `*ScoreRunes_Symmetric` quick.Check test for each algorithm
that has a rune variant. Example for Jaro:
```go
func TestProp_JaroScoreRunes_Symmetric(t *testing.T) {
    f := func(a, b string) bool {
        return fuzzymatch.JaroScoreRunes(a, b) == fuzzymatch.JaroScoreRunes(b, a)
    }
    if err := quick.Check(f, nil); err != nil {
        t.Errorf("JaroScoreRunes not symmetric: %v", err)
    }
}
```
Add equivalents for Levenshtein, Hamming, DL-OSA, DL-Full, and
Jaro-Winkler rune paths. The pattern is mechanical (~6 small functions).

### WR-04: OSA triangle-inequality property test silently passes when generated strings exceed the 6-char constraint

**File:** `props_test.go:442-466`
**Issue:** `TestProp_DamerauLevenshteinOSADistance_TriangleInequality_Constrained`
uses an "early-return-true" idiom to filter quick.Check inputs:
```go
if len(a) > maxLen || len(b) > maxLen || len(c) > maxLen {
    return true // skip inputs that exceed the constrained domain
}
for _, s := range []string{a, b, c} {
    for i := 0; i < len(s); i++ {
        if s[i] < 0x20 || s[i] >= 0x7f {
            return true // skip non-printable / non-ASCII
        }
    }
}
```
This means the property test reports PASS even when **every single
generated triple** is filtered out — i.e. when `quick.Check` never
actually exercises the property. `testing/quick`'s default config draws
strings whose length is typically much longer than 6 bytes and which often
contain non-printable bytes, so the actual rate at which all three of
`(a, b, c)` clear both gates is very low. A reader inspecting the suite
might conclude DL-OSA triangle inequality is property-tested on short
ASCII; in practice the test is mostly a no-op.

This is not a correctness defect (the conditional return is logically
fine), but it weakens the safety net the property test is supposed to
provide. If the OSA recurrence ever regresses to violate triangle
inequality even on short ASCII triples, the test could still pass simply
because quick.Check happened not to generate triples in the small valid
domain.

**Fix:** Either:
1. Replace the filtering with a custom `quick.Config` that constrains the
   generator to the desired domain (preferred — see
   `quick.Config.Values`):
   ```go
   cfg := &quick.Config{
       Values: func(args []reflect.Value, r *rand.Rand) {
           args[0] = reflect.ValueOf(randShortASCII(r, 6))
           args[1] = reflect.ValueOf(randShortASCII(r, 6))
           args[2] = reflect.ValueOf(randShortASCII(r, 6))
       },
   }
   if err := quick.Check(f, cfg); err != nil {
       t.Errorf("DL-OSA triangle inequality: %v", err)
   }
   ```
2. Add a counter inside `f` and `t.Logf` the number of triples actually
   exercised so a reader can see coverage. If the counter is zero, fail
   the test.

The same pattern would benefit `TestProp_HammingDistance_TriangleInequality_EqualLength`
(props_test.go:314), which only skips empty `base` but doesn't check
that a meaningful number of non-empty bases were drawn.

## Info

### IN-01: JaroWinklerScoreRunes does duplicate `[]rune(a)` and `[]rune(b)` allocations

**File:** `jarowinkler.go:156-179`
**Issue:** `JaroWinklerScoreRunes` calls `JaroScoreRunes(a, b)` first
(which internally does `[]rune(a)` and `[]rune(b)` via line 173-174 of
`jaro.go`), then does `ra := []rune(a)` and `rb := []rune(b)` AGAIN at
`jarowinkler.go:163-164` to compute the prefix length. The same two rune
slices are allocated twice per call on the rune path. Documented in the
godoc as "plus an additional []rune conversion" — but it is two extra
conversions, not one, and the godoc undercounts.

**Fix:** Either (a) factor `JaroScoreRunes`-on-runes-and-prefix into an
internal helper that takes pre-converted rune slices, or (b) inline the
Jaro algorithm into `JaroWinklerScoreRunes` and re-use the `[]rune`
conversions. Update the godoc to "plus two additional []rune conversions"
in the interim:
```go
// The rune variant allocates two []rune slices (from the underlying
// JaroScoreRunes call) plus an additional TWO []rune conversions to
// compute the common prefix on rune boundaries.
```

### IN-02: JaroScoreRunes does not short-circuit on `a == b` before allocation

**File:** `jaro.go:172-176`
**Issue:** `JaroScoreRunes("café", "café")` performs `[]rune(a)` and
`[]rune(b)` (2 heap allocations) before `jaroRunes`'s
`runeSlicesEqual(ra, rb)` check returns 1.0. A string-equality
short-circuit before allocation would save the two allocations on the
identity path. The byte path (`JaroScore`) does have this short-circuit
at `jaro.go:131`.

**Fix:** Add the string-equality check at the entry point:
```go
func JaroScoreRunes(a, b string) float64 {
    if a == b {
        return 1.0 // fast identity — covers both-empty and identical
    }
    ra := []rune(a) // 1 alloc
    rb := []rune(b) // 1 alloc
    return jaroRunes(ra, rb)
}
```
Apply the same pattern to `LevenshteinDistanceRunes`,
`LevenshteinScoreRunes`, `HammingDistanceRunes`, `HammingScoreRunes`,
`DamerauLevenshteinOSADistanceRunes`, `DamerauLevenshteinOSAScoreRunes`,
`DamerauLevenshteinFullDistanceRunes`, `DamerauLevenshteinFullScoreRunes`,
and `JaroWinklerScoreRunes` for consistency. None of these currently
have an entry-point `if a == b` short-circuit on the rune variant.

### IN-03: BDD step regex `(\d+\.\d+)` rejects integer-form scores

**File:** `tests/bdd/steps/algorithms_steps.go:300, 304`
**Issue:** The BDD step regexes for score assertions use
`(\d+\.\d+)` — at least one digit before AND after the decimal point. A
feature file writing `the score should be exactly 0` or
`the score should be approximately 1 within 0.001` would fail to match
the step. Authors must always write `0.0` and `1.0`. This is a minor
authoring pitfall; godog reports it as "undefined step" rather than a
helpful diagnostic.

**Fix:** Relax the regex to `(\d+\.?\d*)` (one or more digits, optional
decimal point and fractional part) or `(-?\d+(?:\.\d+)?)` if negative
values ever need to be expressed (they shouldn't — scores are in [0,1]).
Alternatively, document explicitly in the file comment that scores in
features must use decimal form.

### IN-04: Example output `want` constant is brittle to whitespace drift

**File:** `examples/identifier-similarity/main_test.go:39-48`
**Issue:** The `want` constant uses literal `Pair (a / b)` followed by a
long run of spaces aligned by `%-32s` formatting. A future edit that
changes any column width (e.g. `pairWidth` from 32 to 30, or `algoWidth`
from 13 to 14) regenerates the entire `want` block. A reader investigating
a failure cannot easily tell whether the diff is a score change (real
regression) or a column-width change (cosmetic). The fact that all 7
pairs and 6 algorithm columns embed in a single string makes the failure
message a wall of text.

**Fix:** Either (a) compute `want` programmatically from a per-cell map
in the test (so failures show which cell drifted), or (b) parse the
captured output into rows/columns and diff cell-by-cell with a clearer
error message. As a lighter alternative, add a TestExample_ColumnWidths
test that pins the layout constants separately from the cell values, so a
column-width change does not require regenerating the score data.

### IN-05: `damerau_full.go` Lowrance-Wagner constant `bigVal = m + n` is undocumented as load-bearing

**File:** `damerau_full.go:225`
**Issue:** `bigVal := m + n` is chosen as the phantom-sentinel value. The
maximum legitimate edit distance is `max(m, n) ≤ m + n`, and the
transposition cost when `D[l-1, k-1] = bigVal` is
`bigVal + (i - l - 1) + 1 + (j - k - 1) ≥ m + n + 1` — which exceeds any
real distance. So `m + n` works.

But the relationship `bigVal > max_legitimate_distance` is the load-
bearing invariant: if someone optimised by computing `bigVal = max(m, n)`
or `bigVal = m + n - 1` to save a byte of memory, the sentinel rows could
get selected as real winners, silently producing wrong distances on
contrived inputs. The current godoc says "large enough to prevent
phantom-sentinel transpositions from being selected" but doesn't show the
algebra. A future reader could change it without realising.

**Fix:** Document the invariant inline:
```go
// bigVal must satisfy bigVal > max(m, n) — any real edit distance is
// ≤ max(m, n), so a transposition cost involving a sentinel reading is
// always at least bigVal + 1 > any real D[i,j]. The minimum legal value
// is max(m, n) + 1; we use m + n for safety margin and to match the
// canonical Lowrance-Wagner sentinel choice.
bigVal := m + n
```
If WR-02 is taken and the sentinel mechanism is removed, this finding
goes away.

### IN-06: `theDistanceShouldBe` BDD step is shared across Hamming and Damerau scenarios without disambiguation

**File:** `tests/bdd/steps/algorithms_steps.go:131-136, 320, 331`
**Issue:** The step
```gherkin
Then the distance should be <N>
```
is registered once (line 329-332) for Hamming, and the
`iComputeTheDamerauLevenshteinOSADistanceBetween` /
`iComputeTheDamerauLevenshteinFullDistanceBetween` step functions write
to the same `lastDistance` field. If a future scenario mixes "compute the
Hamming distance" with "compute the DamerauLevenshteinOSA distance" in
the same scenario and asserts on `lastDistance`, the assertion uses
whichever computation came last. This is correct as written
(per-scenario contexts are isolated), but a reader can't tell that the
"distance" step is algorithm-agnostic by design.

**Fix:** Add a one-line comment on `theDistanceShouldBe` clarifying the
intent ("matches lastDistance written by any *Distance step in the
current scenario; if a scenario mixes algorithm distance steps, the
assertion applies to the last computed"), or add per-algorithm distance
assertion steps if that ambiguity ever causes a scenario authoring bug.

### IN-07: `example_test.go` ExampleHammingScore comment contradicts the file godoc

**File:** `example_test.go:43-52`
**Issue:** The godoc above `ExampleHammingScore` reads:
> "demonstrates the LOCKED unequal-length silent-zero policy: inputs of
> different lengths return 0.0 silently — no error, no panic."

This is correct. But the inline comment in the function body at line 47
says:
> `// Unequal-length: silent-zero policy (max(3,2)=3, score = 1-3/3 = 0.0).`

That math is right (`HammingDistance("abc", "ab") = max(3, 2) = 3`, so
`score = 1 - 3/3 = 0`). However, the **outer godoc says the silent-zero
policy applies to "inputs of different lengths"** — but the actual
distance formula in `hamming.go:69-90` returns `max(len(a), len(b))` for
unequal-length inputs, not 0. The score normalisation (not the distance)
is what produces 0.0. A pedantic reader could parse "silent-zero policy"
as "the distance is zero" which would be wrong.

**Fix:** Tighten the godoc wording from "silent-zero policy" to
"silent-zero-score policy" or "score-zero policy on length mismatch" so
the distinction between `HammingDistance` (returns `max(len)`) and
`HammingScore` (returns `0.0`) is preserved. Apply the same
wording-tightening to `hamming.go:29-31, 56-66, 124-135` for consistency.

---

_Reviewed: 2026-05-14_
_Reviewer: Claude (gsd-code-reviewer)_
_Depth: standard_
