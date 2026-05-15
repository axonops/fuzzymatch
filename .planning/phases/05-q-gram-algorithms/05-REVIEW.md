---
phase: 05-q-gram-algorithms
reviewed: 2026-05-15T00:00:00Z
depth: standard
files_reviewed: 35
files_reviewed_list:
  - algoid_test.go
  - algorithms_golden_test.go
  - cosine.go
  - cosine_bench_test.go
  - cosine_fuzz_test.go
  - cosine_test.go
  - cross_algorithm_consistency_test.go
  - dispatch_cosine.go
  - dispatch_qgram_jaccard.go
  - dispatch_sorensen_dice.go
  - dispatch_tversky.go
  - errors.go
  - example_test.go
  - examples/identifier-similarity/main.go
  - examples/identifier-similarity/main_test.go
  - export_test.go
  - props_test.go
  - q_gram.go
  - q_gram_test.go
  - qgram_jaccard.go
  - qgram_jaccard_bench_test.go
  - qgram_jaccard_fuzz_test.go
  - qgram_jaccard_test.go
  - sorensen_dice.go
  - sorensen_dice_bench_test.go
  - sorensen_dice_fuzz_test.go
  - sorensen_dice_test.go
  - tversky.go
  - tversky_bench_test.go
  - tversky_fuzz_test.go
  - tversky_test.go
  - tests/bdd/features/cosine.feature
  - tests/bdd/features/qgram_jaccard.feature
  - tests/bdd/features/sorensen_dice.feature
  - tests/bdd/features/tversky.feature
  - tests/bdd/steps/algorithms_steps.go
findings:
  critical: 0
  warning: 5
  info: 7
  total: 12
status: issues_found
---

# Phase 5: Code Review Report

**Reviewed:** 2026-05-15
**Depth:** standard
**Files Reviewed:** 35
**Status:** issues_found

## Summary

Phase 5 ships four q-gram algorithms (Q-Gram Jaccard, Sørensen-Dice, Cosine, Tversky) plus the shared `q_gram.go` extraction infrastructure. The implementation is methodical, with strong adherence to the project's correctness, determinism, and licensing standards. Primary-source citations are present and accurate (Ukkonen 1992, Dice 1945 + Sørensen 1948, Salton & McGill 1983, Tversky 1977). The math implements the formulas faithfully; all twelve algorithm functions panic on `n < 1` per the locked direct-call contract; the asymmetric Tversky semantics are correctly anchored (`α` weighs `|A−B|`, `β` weighs `|B−A|`) and defended by a four-layer regression battery (unit, property, BDD, golden file).

No correctness or determinism BLOCKER findings. The Cosine `sort.Strings`-then-iterate pattern correctly implements the LOCKED determinism gate from CONTEXT.md §3, and every `math.X` usage is restricted to `math.Sqrt` / `math.Abs` / `math.IsNaN` / `math.IsInf` (no `Pow`/`Log`/`Exp`/`FMA`). Map iteration in the algorithm hot paths is confined to integer reductions where addition is associative — DET-03 is observed.

The findings below are quality issues: dead branches that should be removed, an inconsistent precondition order in Tversky, an alloc-budget ceiling that gives 75% slack over the spec, alpha/beta validation that does not reject NaN, and one minor BDD-step subtlety that risks a false-negative on the asymmetry gate.

## Warnings

### WR-01: Dead `if denom == 0.0` branch in `tverskyFromQGramMaps` — both branches return `0.0`

**File:** `tversky.go:392-398`

**Issue:** The `intersection == 0` short-circuit contains two arms that return the identical value:

```go
if intersection == 0 {
    denom := (alpha * float64(aMinusB)) + (beta * float64(bMinusA))
    if denom == 0.0 {
        return 0.0
    }
    return 0.0 // 0 / positive = 0; explicit short-circuit avoids the division.
}
```

The local variable `denom` is computed but never used after the `== 0.0` test, and both `return 0.0` arms produce the same float64 bit pattern. The `if denom == 0.0` check is logically dead — control flow is identical regardless of the test outcome.

If the author's intent was to defend against a future refactor that might compute `0.0 / denom` (yielding `NaN` when `denom == 0`), the dead branch is a poor defence: there's no test that asserts the inner `if` is reached on the `denom == 0.0` path (no input pair currently triggers it because the early `len(qa) == 0 && len(qb) == 0` guard already catches the only configuration that would zero both `aMinusB` and `bMinusA`).

This is a WARNING (not BLOCKER) because the observable behaviour is correct. But the dead branch will trip linters (`unparam`, `staticcheck` SA9003 / SA4006) and confuses readers about the algorithm's denominator-protection invariants.

**Fix:** Either remove the inner `if` entirely (preferred — the outer `intersection == 0` short-circuit is sufficient), or keep `denom` and assert its non-zeroness in a test:

```go
// Option A: collapse to a single return
if intersection == 0 {
    return 0.0 // numerator is 0; the lower-bound clamp regardless of denom
}
```

If the author wants to retain a structural guard against a future refactor that introduces a `0.0 / denom` path, replace the dead branch with a documented invariant assertion instead, and add a test that exercises the impossible path.

---

### WR-02: Tversky precondition ordering — `n < 1` panic gated AFTER `a == b` short-circuit

**File:** `tversky.go:230-243` (and identical pattern at `tversky.go:276-285` for `TverskyScoreRunes`)

**Issue:** The validation order in `TverskyScore` is:

```go
if a == b {
    return 1.0 // identity short-circuit
}
if n < 1 {
    panic("fuzzymatch: invalid q-gram size")
}
if alpha < 0 || beta < 0 || (alpha == 0 && beta == 0) {
    panic("fuzzymatch: invalid tversky parameter")
}
```

This means `TverskyScore("hello", "hello", -5, 0.5, 0.5)` returns `1.0` instead of panicking, while `TverskyScore("hello", "world", -5, 0.5, 0.5)` panics. The two q-gram-tier siblings (`QGramJaccardScore`, `SorensenDiceScore`, `CosineScore`) all validate `n < 1` BEFORE the identity short-circuit:

```go
// qgram_jaccard.go:144-150 (and same in sorensen_dice.go, cosine.go):
func QGramJaccardScore(a, b string, n int) float64 {
    if n < 1 {
        panic("fuzzymatch: invalid q-gram size")
    }
    if a == b {
        return 1.0
    }
    ...
}
```

The Tversky tests acknowledge this divergence — `TestTversky_PanicsOnInvalidN` deliberately uses `"abc"`/`"abd"` (distinct inputs) to dodge the short-circuit ("Use distinct inputs so a == b short-circuit / does not fire"), and the `TestTversky_Identical` test asserts that identical inputs return `1.0` regardless of `n`/`α`/`β`. This is by design, but it means:

1. `TverskyScore` has subtly different precondition semantics from its three siblings — programmer error on `n` is silently absorbed when inputs happen to be identical.
2. Equally, the `α`/`β` validation is bypassed on identical inputs: `TverskyScore("hello", "hello", 2, -1.0, -1.0)` returns `1.0` rather than panicking, even though `(-1.0, -1.0)` is documented as invalid.

This is a contract divergence WARNING. Either the order should match the other three q-gram algorithms (validate `n`/`α`/`β` first, then the identity short-circuit), or the tests should assert the divergence is intentional and document it in the godoc.

**Fix:** Move all validation before the short-circuit so the four q-gram algorithms have uniform precondition semantics:

```go
func TverskyScore(a, b string, n int, alpha, beta float64) float64 {
    if n < 1 {
        panic("fuzzymatch: invalid q-gram size")
    }
    if alpha < 0 || beta < 0 || (alpha == 0 && beta == 0) {
        panic("fuzzymatch: invalid tversky parameter")
    }
    if a == b {
        return 1.0
    }
    if a == "" || b == "" {
        return 0.0
    }
    ...
}
```

Then update `TestTversky_Identical` to drop the invalid `(-1, -1)` configurations and add explicit tests for `TverskyScore("hello", "hello", 0, ...)` panicking and `TverskyScore("hello", "hello", 2, -1.0, 0.5)` panicking. Mirror the same change in `TverskyScoreRunes`.

---

### WR-03: Tversky `α`/`β` validation does not reject NaN — flagged in plan but not unit-tested

**File:** `tversky.go:241,283` (validation expression `alpha < 0 || beta < 0 || (alpha == 0 && beta == 0)`)

**Issue:** The validation expression `alpha < 0 || beta < 0 || (alpha == 0 && beta == 0)` returns `false` when `alpha` or `beta` is `NaN`, because every comparison against `NaN` is `false` in IEEE-754. The function then proceeds to `float64(intersection) / (float64(intersection) + (NaN * float64(aMinusB)) + ...)` which produces `NaN`, in violation of the `PropAlgorithm_NoNaN` invariant the project promises (`determinism-standards`, "NaN/Inf/Negative Zero").

The fuzz harness (`tversky_fuzz_test.go:74-83`) has a `fuzzCoerceTverskyParam` helper that explicitly maps NaN to `0.5` to avoid exercising this path:

```go
func fuzzCoerceTverskyParam(p float64) float64 {
    if math.IsNaN(p) {
        return 0.5
    }
    ...
}
```

The fuzz file's own godoc acknowledges the gap as deferred hardening:

> Future hardening (deferred): NaN α / β inputs are not directly exercised by this harness because the squashing function maps NaN to a safe interior point. If a NaN-input fuzz finding emerges in practice, the fuzz body can be loosened to pass NaN through and the public-API gate updated to detect it.

The plan mentions this as a "v1.x hardening item" per the review prompt. WARNING because:

1. The public API can produce NaN today on `TverskyScore("a", "b", 2, math.NaN(), 0.5)` — direct call, not via the Scorer.
2. No unit test asserts the current behaviour, so there's no regression detector for whether NaN is rejected, accepted-and-returns-NaN, or accepted-and-returns-0.0 once hardening lands.
3. The property tests (`TestProp_TverskyScore_NoNaN`) already use the same NaN-coercing helper, so they'll never catch this either.

**Fix:** Add an explicit NaN guard to the validation, and add a unit test that pins the chosen behaviour:

```go
// In TverskyScore + TverskyScoreRunes:
if math.IsNaN(alpha) || math.IsNaN(beta) {
    panic("fuzzymatch: invalid tversky parameter")
}
if alpha < 0 || beta < 0 || (alpha == 0 && beta == 0) {
    panic("fuzzymatch: invalid tversky parameter")
}

// New test in tversky_test.go:
func TestTversky_PanicsOnNaNParams(t *testing.T) {
    tests := []struct{ alpha, beta float64 }{
        {math.NaN(), 0.5},
        {0.5, math.NaN()},
        {math.NaN(), math.NaN()},
    }
    for _, tt := range tests {
        // ... defer recover() asserting panic with "invalid tversky parameter"
    }
}
```

If the team has decided to defer this fix to v1.x, that's fine — but please record an issue with the chosen behaviour pinned (`silently produces NaN` is a worse status quo than `documented panic` for a v0.x algorithm).

---

### WR-04: Alloc-budget ceiling at 6 (Jaccard / Dice / Tversky) is 50% above the spec budget of 4

**File:** `qgram_jaccard_test.go:292-300`, `sorensen_dice_test.go:299-307`, `tversky_test.go:718-726`

**Issue:** The three q-gram algorithms test against a ceiling of `6.0` allocs/op, but `performance-standards/SKILL.md` documents the budget as `≤ 4 allocations`:

> Q-Gram Jaccard, Sørensen-Dice, Cosine, Tversky: < 5 µs, ≤ 4 allocations

The test comments acknowledge the divergence:

> CONTEXT.md §5 / RESEARCH.md §4.1 budget is "≤ 4 allocations" but the realistic table in RESEARCH.md §4.1 acknowledges 4–11 across input sizes (the 4 floor is the canonical-source ideal; the realistic ceiling is what matters for regression detection).

This is an admission that the actual implementation does not meet the spec. The test ceiling at `6` will pass even when the implementation regresses by 50% over the budget. The benchmark file's comments document the same ceiling as `≤ 4 allocs` for short inputs (`qgram_jaccard_bench_test.go:23`):

> ASCII Short  (~10 chars):  ≤ 4 allocs/op (two map allocations)

If the actual measurement shows 4 allocs on short inputs as the bench file claims, the test should pin `≤ 4`. If 4 is unattainable in practice (the realistic table says so), the spec should be updated and `performance-standards/SKILL.md` brought into agreement, with a CHANGELOG note explaining the divergence.

The Cosine ceiling of `7.0` (`cosine_test.go:459`) sits closer to the documented "Jaccard + 1 sorted-key slice" rationale, but it shares the same gap-with-spec issue: if the spec says `≤ 4` then `≤ 5` would be the disciplined ceiling for Cosine.

**Fix:** Either:

1. Tighten the test ceilings to match the spec (`4` for Jaccard/Dice/Tversky, `5` for Cosine) and accept that the test will fail until the implementation hits the budget — then deliver the optimisation that gets it there (e.g. lazy / reusable map pools, single-pass per-side counting).
2. Update `performance-standards/SKILL.md` and `docs/requirements.md` §14.1 with the realistic ceiling and a rationale, then keep the test pin matching the new spec.

The current state — test passes at `6`, spec demands `4`, comments wave both flags — is the worst of both worlds: there's no regression detection on the actual budget AND the documentation lies to consumers about per-call cost.

---

### WR-05: Cosine `theTwoTverskyScoresShouldDifferByMoreThan` regex captures floats but lastScore reuse is brittle

**File:** `tests/bdd/steps/algorithms_steps.go:613-619` plus the underlying `lastScore`/`lastScore2` field design

**Issue:** The Tversky asymmetry-direction-sensitivity step relies on `ctx.lastScore` and `ctx.lastScore2` being set by the previous two `When` clauses. The step body is:

```go
func (ctx *AlgorithmContext) theTwoTverskyScoresShouldDifferByMoreThan(threshold float64) error {
    delta := math.Abs(ctx.lastScore - ctx.lastScore2)
    if delta <= threshold {
        return fmt.Errorf("tversky asymmetry gate FAILED: ...")
    }
    return nil
}
```

`lastScore` and `lastScore2` are zero-valued `float64` at the start of every scenario. If a future scenario author writes only ONE `When I compute the Tversky score ...` step (forgetting the `second`) and follows with `Then the two Tversky scores should differ by more than 0.1`, the comparison becomes `|computed_score - 0.0| > 0.1` — which silently passes for any non-tiny computed score.

The same risk exists for every `bothXxxScoresShouldBeEqual` step (Jaccard, Dice, Cosine, Tversky) — `0.0 == 0.0` would pass even if neither `When` step ran.

This is WARNING (not BLOCKER) because the existing scenarios in `tversky.feature` do call both steps in order, and the load-bearing asymmetry pin is also covered by the unit test (`TestTversky_AsymmetryDirectionSensitive`), the property test (`TestProp_TverskyScore_AsymmetricWhenAlphaNeqBeta`), and the golden file (RV-T1/RV-T2 entries). The defence-in-depth holds. But the BDD step contract is silently fragile against a future scenario author error.

**Fix:** Track an `lastScoreSet`/`lastScore2Set` bool pair on the context (or use `*float64` with `nil` as the unset marker) and have every `bothXxx` / `theTwoXxx` step assert both flags are true:

```go
func (ctx *AlgorithmContext) theTwoTverskyScoresShouldDifferByMoreThan(threshold float64) error {
    if !ctx.lastScoreSet || !ctx.lastScore2Set {
        return fmt.Errorf("internal: scores not populated — scenario must call both Tversky `When` steps before this `Then`")
    }
    ...
}
```

Reset both flags in each scenario's `Before` hook or in the `iComputeTheXxx` step bodies (set the flag, mirror in `iComputeTheSecondXxx`).

## Info

### IN-01: Inconsistent epsilon convention across the four q-gram algorithm tests

**File:** `qgram_jaccard_test.go:42`, `sorensen_dice_test.go:43`, `cosine_test.go:49`, `tversky_test.go:48`

**Issue:** The four files use three different epsilons:

- `jaccardEpsilon = 1e-9` (Jaccard) — comment says "Phase 2/3/4 convention is 1e-9"
- `sorensenDiceEpsilon = 1e-9` (Dice) — same convention
- `cosineEpsilon = 1e-15` (Cosine) — comment says "convention is locked at 1e-15"
- `tverskyEpsilon = 1e-15` (Tversky) — same as Cosine

Both Cosine and Tversky note the actual accuracy is "far higher than 1e-9" or "last-bit level". The Jaccard and Dice tests use the looser convention even though their accuracy is also "far higher than 1e-9" (the formula is a single integer-valued division). Two epsilons across four sibling algorithms in the same phase suggests a copy-paste convention drift.

**Fix:** Pick one tolerance convention for the q-gram tier and apply it to all four files. Recommend `1e-15` (matches Cosine/Tversky and the actual algorithm accuracy) and the comment can call out that the Jaccard/Dice formulas are effectively bit-exact for any input fitting in float64. The looser `1e-9` epsilon hides real rounding regressions if the formula is ever refactored.

---

### IN-02: `TestQGramJaccardRunes_CafeReference` uses exact equality but `TestSorensenDiceRunes_CafeReference` uses tolerance

**File:** `qgram_jaccard_test.go:223-229`, `sorensen_dice_test.go:229-236`

**Issue:** The Jaccard rune test uses `got != want` with `want := 0.5` (exact equality):

```go
func TestQGramJaccardRunes_CafeReference(t *testing.T) {
    got := fuzzymatch.QGramJaccardScoreRunes("café", "cafe", 2)
    want := 0.5
    if got != want {
```

The Dice rune test uses `math.Abs(got-want) > 1e-15` with `want := 4.0/6.0`:

```go
func TestSorensenDiceRunes_CafeReference(t *testing.T) {
    got := fuzzymatch.SorensenDiceScoreRunes("café", "cafe", 2)
    want := 4.0 / 6.0
    if math.Abs(got-want) > 1e-15 {
```

Both algorithms compute their result via a single integer-valued division on small ints. The Jaccard formula is `intersection / union = 2 / 4 = 0.5` — exact. The Dice formula is `2*intersection / (lenA + lenB) = 4 / 6 ≈ 0.6666...` — irrational in binary float64. So the inconsistency is justified by the maths, but it's worth a comment explaining "why exact for Jaccard but tolerance for Dice".

**Fix:** Add a brief comment to each test explaining the choice (`// 2/4 = 0.5 is exactly representable in float64` vs `// 4/6 = 0.6666... is irrational in binary; tolerance required`).

---

### IN-03: Magic threshold `0.1` in `TestTversky_AsymmetryDirectionSensitive` and BDD step are not derived from a constant

**File:** `tversky_test.go:273`, `cross_algorithm_consistency_test.go:577`, `tests/bdd/features/tversky.feature:86`

**Issue:** The asymmetry-magnitude minimum (`0.1`) is repeated in three places without a shared constant. The actual delta is `≈ 0.2302` (documented in the test comment). If a future tightening of the rounding behaviour pulls the delta below `0.1` (unlikely but possible), the three sites must be updated independently.

**Fix:** Declare an unexported constant in the test file (e.g. `const tverskyAsymmetryMinDelta = 0.1`) and reference it from both Go test sites. The BDD scenario value cannot be made a constant directly — but a comment cross-referencing the Go constant would document the source of truth.

---

### IN-04: `q_gram.go` allocation-budget claim conflicts with realistic ceiling in tests

**File:** `q_gram.go:80-86`

**Issue:** The file-level godoc states:

> Allocation budget (RESEARCH.md §4.1, docs/requirements.md §14.1):
>   - Two map allocations per algorithm call (one per side).
>   - Up to ~2 additional map-growth allocations on medium-to-long inputs (capacity hint mitigates this).

So the documented ceiling is `~4` allocs (two extractor calls × two map allocations apiece, minus growth on short inputs). The corresponding `TestQGramJaccard_AllocsBudget` ceiling is `6.0`. Either the godoc undercounts (algorithm overhead, hashing, etc.), the tests over-budget (cf. WR-04), or the documentation lies to consumers reading the file.

**Fix:** Reconcile q_gram.go's allocation-budget block with the actual test ceilings AND `performance-standards/SKILL.md`. Once the WR-04 reconciliation lands, this comment auto-resolves.

---

### IN-05: Cosine `cos < 0.0` clamp is unreachable and uncovered by any test

**File:** `cosine.go:385-390`

**Issue:** The Cauchy-Schwarz inequality plus the non-negative integer-derived dot product makes `cos < 0.0` impossible:

```go
if cos > 1.0 {
    return 1.0
}
if cos < 0.0 {
    return 0.0
}
return cos
```

The comment correctly describes this:

> The lower clamp is a defensive guard — the dot product of non-negative integer-derived values is non-negative, and floor(0.0 / positive) is +0.0 in IEEE-754; `cos < 0.0` is theoretically unreachable but the clamp costs nothing.

This is dead defensive code. It will be flagged by any coverage report (the `< 0.0` branch is uncoverable). For a project that targets ≥ 95% overall and ≥ 90% per-file coverage, an unreachable branch in a hot-path file lowers the achievable coverage ceiling. The "costs nothing" claim is true at runtime but not at coverage time.

**Fix:** Either:

1. Remove the `if cos < 0.0` branch and rely on the maths (recommended).
2. Add a `//nolint:gocyclo` or coverage exclusion comment that the linter picks up.
3. Keep the branch and document in `cosine.go` that the per-file coverage target is reduced because of this defensive guard.

---

### IN-06: `tversky.go` `lastScoreSet` discipline echoed across four BDD step methods invites copy-paste drift

**File:** `tests/bdd/steps/algorithms_steps.go:441-632`

**Issue:** The four q-gram algorithm step blocks (Jaccard, Dice, Cosine, Tversky) duplicate the "compute / store / compare" pattern with near-identical bodies. Each block has three step methods (`iComputeTheXxxScoreBetweenWithN`, `iComputeTheSecondXxxScoreBetweenWithN`, `iComputeTheXxxRunesScoreBetweenWithN`) plus an assertion (`bothXxxScoresShouldBeEqual`). All four assertion bodies use the same template:

```go
func (ctx *AlgorithmContext) bothXxxScoresShouldBeEqual() error {
    if ctx.lastScore != ctx.lastScore2 {
        return fmt.Errorf("xxx scores not equal: %f != %f", ctx.lastScore, ctx.lastScore2)
    }
    return nil
}
```

A small refactor would replace the four assertion functions with one parameterised by the algorithm name string. Same applies to the score-equality step regex registrations.

**Fix:** Introduce a generic `bothScoresShouldBeEqual(algorithmName string) error` helper and call it from each algorithm-specific step. This reduces ~30 lines of near-duplicate code to ~5. Lower priority because the duplication is mechanical, not subtly drifted.

---

### IN-07: `Cosine_unicode_n3_runes` uses the rune path but is filed under the `Cosine` algorithm — golden entry naming does not encode the surface

**File:** `algorithms_golden_test.go:1252-1259`

**Issue:** The Cosine staging file contains entries computed via both `CosineScore` (byte path) and `CosineScoreRunes` (rune path):

```go
{
    Name:          "Cosine_unicode_n2_runes",
    Algorithm:     "Cosine",  // <-- same as byte-path entries
    A:             "café",
    B:             "cafe",
    ExpectedScore: fuzzymatch.CosineScoreRunes("café", "cafe", 2),
},
```

The `Algorithm` field reads `"Cosine"` for both surfaces, even though the rune-path entries are computed by the `*Runes` variant. A consumer reading the golden file cannot determine which surface produced which score without checking the input strings for non-ASCII content (and that's not a reliable heuristic either — `"hello"` could be computed via `CosineScoreRunes` and produce the identical score). The entry name does encode `_runes` but the schema's `Algorithm` field does not.

This is INFO because the golden file's load-bearing role is byte-stability across platforms, which is unaffected. But for downstream consumers (e.g. an audit tool comparing two algorithm versions), the surface ambiguity could matter.

**Fix:** Either:

1. Add a per-entry `Surface` field (`"byte"` or `"rune"`) to the schema. Note this is a v1.x schema-bump event per the file-level godoc.
2. Change the `Algorithm` field for rune entries to `"CosineRunes"` (and similarly for the other q-gram rune entries).
3. Document that the `Algorithm` field is the algorithm family and the surface is implied by the entry name suffix.

---

_Reviewed: 2026-05-15_
_Reviewer: Claude (gsd-code-reviewer)_
_Depth: standard_
