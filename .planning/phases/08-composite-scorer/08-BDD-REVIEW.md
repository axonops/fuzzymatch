---
phase: 08-composite-scorer
reviewed: 2026-05-17T00:00:00Z
reviewer: bdd-scenario-reviewer
feature_file: tests/bdd/features/scorer.feature
step_file: tests/bdd/steps/scorer_steps.go
scenarios_reviewed: 12
gaps_found:
  BLOCKING: 3
  IMPORTANT: 2
  NIT: 2
status: NO-GO
---

# Phase 8 BDD Coverage Review

**Reviewed:** 2026-05-17
**Feature file:** `tests/bdd/features/scorer.feature`
**Step definitions:** `tests/bdd/steps/scorer_steps.go`
**Scenarios reviewed:** 12
**Cross-referenced:** `08-CONTEXT.md §7`, `docs/requirements.md §15.6`, `.claude/skills/go-testing-standards/SKILL.md` (BDD section)

---

## Summary

The 12 Phase 8 scenarios cover all mandatory CONTEXT.md §7 scenario classes by count.
The goleak gate is wired correctly in `tests/bdd/bdd_test.go`, the BDD sub-module
correctly confines godog + goleak + testify to `tests/bdd/go.mod`, and the
Given/When/Then structure is generally clear. Three BLOCKING gaps prevent this
from being GO.

**BLOCKING** (3): one confirmed step bug already flagged in 08-REVIEW.md
(WR-02) that causes a scenario to silently test the wrong pair; one missing
sentinel-error scenario (`ErrInvalidThreshold`) whose step infrastructure is
already in place; one scenario whose assertion does not exercise the behaviour
it claims to document (weight auto-normalisation is never directly asserted).

**IMPORTANT** (2): `WithNormalisation(custom)` composed path from CONTEXT.md §3
has no BDD coverage; the `ScoreAll` scenario verifies key presence but does not
verify that the per-algorithm values are in `[0.0, 1.0]`.

**NIT** (2): the feature-level `@scorer` tag and scenario-level `@scorer` tags
are redundant (fine), but the scenarios have no layer-independent alias tag
(the skill calls for both a category tag and a layer tag); and the
`WithoutNormalisation` scenario uses a relative assertion (less-than) where a
concrete value assertion would anchor the contract.

---

## BLOCKING Findings

### BDD-01 Missing scenario: `ErrInvalidThreshold`

**Severity:** BLOCKING
**Feature file:** `tests/bdd/features/scorer.feature`
**Missing scenario:** No scenario verifies that passing an out-of-range threshold
(`t > 1.0`, `t < 0.0`) to `WithThreshold` returns `ErrInvalidThreshold`.

The gap was already identified as IN-04 in `08-REVIEW.md`, but it is BLOCKING
here — not merely informational — because `ErrInvalidThreshold` is a documented
sentinel in `errors.go` (line 146), is listed in the CONTEXT.md §7 mandatory
error class set alongside `ErrMissingThreshold` / `ErrEmptyScorer` /
`ErrInvalidWeight`, and the `constructingTheScorerShouldReturn` step in
`scorer_steps.go:339` already has the case branch for it. The BDD suite is the
consumer contract; a documented sentinel with no scenario is an undocumented
error from the consumer's point of view.

Additionally, the step regex for the threshold in the error-path scenarios is
`(\d+\.?\d*)` — it only matches non-negative numbers. A negative-threshold
sub-case (`t = -0.1`) cannot be exercised with the current regex. The regex
must be widened to `(-?\d+\.?\d*)` to accept both the out-of-range-high and
out-of-range-low sub-cases.

**Suggested Gherkin:**

```gherkin
@scorer @errors
Scenario: WithThreshold out-of-range returns ErrInvalidThreshold
  # threshold 1.5 is outside [0.0, 1.0]; ErrInvalidThreshold fired at
  # option-application time, before NewScorer reaches the validation pipeline.
  When I attempt to construct a Scorer with Levenshtein weight 1.0 and threshold 1.5
  Then constructing the Scorer should return ErrInvalidThreshold

@scorer @errors
Scenario: WithThreshold NaN returns ErrInvalidThreshold
  # NaN passes t < 0.0 || t > 1.0 as currently coded (CR-01 in 08-REVIEW.md);
  # this scenario is intentionally BLOCKED on CR-01 being fixed first.
  # Add once scorer_options.go:258 adds math.IsNaN(t) to the guard.
  # (Included here as the intended BDD contract so test-writer can land it
  # alongside the CR-01 fix.)
  When I attempt to construct a Scorer with Levenshtein weight 1.0 and threshold NaN
  Then constructing the Scorer should return ErrInvalidThreshold
```

The NaN scenario also requires a new step for `threshold NaN` — the floating-
point step regex cannot parse `NaN` as a `float64` from a Gherkin string; a
dedicated `I attempt to construct a Scorer with NaN threshold` step is cleaner.
Add that step alongside the CR-01 fix.

---

### BDD-02 Bug: `iScoreTheSamePairWithTheDefaultScorer` hardcodes the pair (WR-02)

**Severity:** BLOCKING
**Feature file:** `tests/bdd/features/scorer.feature` lines 98-102
**Step file:** `tests/bdd/steps/scorer_steps.go` lines 189-195

The Gherkin reads:

```gherkin
When I score "XMLParser" and "xml_parser" with the Scorer
And I record the no-normalisation Scorer composite score
And I score the same pair with the default Scorer
```

The step bound to `^I score the same pair with the default Scorer$` ignores
the pair captured in the preceding `I score X and Y` step and instead
hardcodes `sc.defaultScorer.Score("XMLParser", "xml_parser")` directly. The
pair happens to match what the scenario uses, so the test passes today. But:

1. The natural-language contract of "the same pair" is violated — the step does
   not actually use the same pair. A future scenario reusing this step phrase
   with different inputs will silently score `XMLParser`/`xml_parser` and
   produce meaningless results.
2. The `ScorerContext` does not store `lastA` / `lastB`, so there is no path
   for the step to actually honour its Gherkin contract.

This is a confirmed bug, not a style issue. The scenario currently passes only
by accidental coincidence of the hardcoded values matching the scenario's
`When I score` step arguments.

**Fix (from 08-REVIEW.md WR-02, reproduced here for completeness):**
Add `lastA`, `lastB string` fields to `ScorerContext`. Populate them inside
`iScoreAndWithTheScorer`. Then rewrite `iScoreTheSamePairWithTheDefaultScorer`
to use `sc.defaultScorer.Score(sc.lastA, sc.lastB)`.

---

### BDD-03 Weak scenario: weight auto-normalisation never directly asserted

**Severity:** BLOCKING
**Feature file:** `tests/bdd/features/scorer.feature` lines 58-67

The CONTEXT.md §7 mandatory class 4 is "Custom 2-algorithm Scorer with explicit
weights: composite score is the weighted sum." The scenario "Two-algorithm
weighted Scorer composes correctly" uses equal weights (0.5 + 0.5) on an
identical pair (`"hello"` vs `"hello"`):

```gherkin
Given I construct a Scorer with Levenshtein weight 0.5 and JaroWinkler weight 0.5 and threshold 0.7
When I score "hello" and "hello" with the Scorer
Then the Scorer composite score should be exactly 1.0
```

The assertion `score == 1.0` is true because both algorithms return 1.0 on an
identical pair, and `0.5×1.0 + 0.5×1.0 = 1.0`. This does not exercise weight
auto-normalisation at all:

- With equal weights (0.5, 0.5) the auto-normalisation step is a no-op — the
  weights already sum to 1.0.
- A weighted composite of 1.0 + 1.0 is 1.0 regardless of whether weights are
  0.5/0.5 or 0.3/0.7 or 1.0/1.0.

The CONTEXT.md §3 action item for plan 08-04 explicitly requires a BDD scenario
for "default scorer + identifier-style input → match" (covered) and for the
`WithoutNormalisation` path (covered), but also for the auto-normalisation
behaviour itself: when raw weights sum to something other than 1.0, the
composite is still in [0,1]. There is no scenario that uses unequal weights on
a pair with a known non-trivial composite to verify the normalised sum.

The mandatory class 6 ("last-write-wins") scenario uses
`WithAlgorithm(AlgoLevenshtein, 0.3)` then `WithAlgorithm(AlgoLevenshtein, 0.7)`,
but asserts only `Algorithms() length == 1` and `weight == 1.0` after
normalisation. It does not verify that a Score call produces the expected value.

**What is missing:** A scenario that uses two algorithms with unequal non-trivial
raw weights (e.g. Levenshtein 0.3 + JaroWinkler 0.7) on a non-identical pair
with a pinned expected composite, demonstrating that the Scorer correctly
normalises to sum-to-1 and produces `0.3/1.0 × lev_score + 0.7/1.0 × jw_score`.

**Suggested Gherkin:**

```gherkin
@scorer @custom
Scenario: Weight auto-normalisation produces sum-to-1 composite
  # Raw weights 0.3 + 0.7 already sum to 1.0, but the intent is distinct from
  # the 0.5+0.5 case: the test drives the Scorer with explicitly unequal weights
  # on a non-trivial pair to verify the reduction loop, not just the identity case.
  # Levenshtein("kitten","sitting") ≈ 0.5714; JaroWinkler("kitten","sitting") ≈ 0.7468.
  # Composite (pre-normalised weights 0.3 + 0.7) ≈ 0.3×0.5714 + 0.7×0.7468 ≈ 0.6942.
  Given I construct a Scorer with Levenshtein weight 0.3 and JaroWinkler weight 0.7 and threshold 0.5
  When I score "kitten" and "sitting" with the Scorer
  Then the Scorer composite score should be approximately 0.6942 within 0.001
  And the Scorer match result should be true
```

For `WithNormaliseWeights(false)` (raw weights bypass normalisation), the
CONTEXT.md golden corpus has a mandatory entry (#11 in §6), but there is no
BDD scenario. A scenario asserting the raw composite differs from the
auto-normalised composite when weights sum ≠ 1 would close that gap. This is
classified IMPORTANT below (BDD-04) as it is an opt-in escape hatch rather
than the default path.

---

## IMPORTANT Findings

### BDD-04 Missing scenario: `WithNormalisation(custom)` composed path

**Severity:** IMPORTANT
**Feature file:** `tests/bdd/features/scorer.feature`

The CONTEXT.md §3 action list for plan 08-04's BDD scenarios explicitly
requires three normalisation scenarios:

> - default scorer + identifier-style input → match  (COVERED: scenario 1)
> - `WithoutNormalisation()` + identifier-style input → no match (COVERED: scenario 7)
> - `WithNormalisation(custom)` + Unicode input (`café` vs `cafe`) → behaviour per opts

The third path — `WithNormalisation(custom NormalisationOptions)` composed into
a Scorer, demonstrating that the Scorer applies the caller-supplied normalisation
opts rather than the default — has no scenario. This option is distinct from
both the default-normalisation path and the no-normalisation path; it is the
hook a consumer uses to configure diacritic-stripping, custom separator chars,
or CamelCase-split control without writing their own normalisation. The
`WithNormalisation` option is declared in `scorer_options.go:215-221` and the
`iConstructTheDefaultScorerWithoutNormalisation` step already demonstrates the
corresponding `WithoutNormalisation` pattern; a `WithNormalisation` companion
scenario is straightforward to add.

**Suggested Gherkin:**

```gherkin
@scorer @custom
Scenario: WithNormalisation with custom options controls pre-comparison folding
  # NormalisationOptions{Lowercase:true, StripSeparators:false, SplitCamelCase:false}
  # lowercases but does not strip separators or split camelCase.
  # "UserID" and "user_id" differ only in case and the underscore; with
  # Lowercase:true the strings become "userid" and "user_id" — still different
  # due to the retained underscore — so the score is lower than DefaultScorer.
  Given I construct a Scorer with Levenshtein weight 1.0 and threshold 0.5 and lowercase-only normalisation
  When I score "UserID" and "user_id" with the Scorer
  Then the Scorer composite score should be approximately 0.8571 within 0.001
```

(The step `I construct a Scorer with ... and lowercase-only normalisation` is
a new step that would need to be added to `scorer_steps.go`.)

---

### BDD-05 Weak scenario: `ScoreAll` map values never range-checked

**Severity:** IMPORTANT
**Feature file:** `tests/bdd/features/scorer.feature` lines 140-151

The "ScoreAll returns map keyed by AlgoID" scenario asserts only key presence
and key absence:

```gherkin
Then the ScoreAll map should contain AlgoDamerauLevenshteinOSA
And the ScoreAll map should contain AlgoDoubleMetaphone
And the ScoreAll map should not contain AlgoCosine
```

The scenario comment acknowledges this is a "runtime sanity check" for typed
key safety (a compile-time concern). But the scenario does not verify any
behavioural property of the map values:

- That the values are in `[0.0, 1.0]`
- That the value for `AlgoDamerauLevenshteinOSA` on the identity pair
  `"user_id"/"userId"` is the documented 1.0 after normalisation folds the
  case/separator difference
- That the sum of the map values equals the result of `Score(a, b)` divided by
  the count (since weights are 1/6 each, the map values are raw per-algorithm
  scores and their weighted sum is the composite — but this is never verified)

A scenario that asserts at least one concrete value pins the contract that
`ScoreAll` values are per-algorithm scores (not weighted contributions, not
composites). This is the only scenario exercising `ScoreAll` and it should
document the return semantics more completely.

**Suggested addition to existing scenario:**

```gherkin
And the ScoreAll value for AlgoDamerauLevenshteinOSA should be exactly 1.0
And all ScoreAll values should be in range [0.0, 1.0]
```

---

## NIT Findings

### BDD-06 Missing layer-independent `@algorithm` tag on scorer scenarios

**Severity:** NIT
**Feature file:** `tests/bdd/features/scorer.feature`

The go-testing-standards skill BDD section lists the tag taxonomy:
`@character`, `@qgram`, `@token`, `@phonetic`, `@gestalt` (category) and
`@algorithm`, `@scorer`, `@scan`, `@normalisation`, `@determinism`,
`@suppression` (layer). The scorer scenarios are tagged `@scorer` and `@scorer
@default` / `@scorer @custom` etc., which is correct for the layer tag. But the
system-level agent prompt specifies a separate `@scorer` layer tag distinct from
the category tags, and notes that every scenario must have at least one tag —
which they do.

The actual gap: the feature has no `@algorithm` companion for the
`@scorer @custom` scenarios that compose algorithm-specific behaviour (e.g.
"Single-algorithm Scorer composes correctly" with Levenshtein, "Two-algorithm
weighted Scorer" with Levenshtein + JaroWinkler). These scenarios exercise
algorithm identity through the Scorer layer and would benefit from a secondary
`@levenshtein` or `@jaro_winkler` tag enabling filtered runs that trace
algorithm behaviour across layers. The algorithm feature files (e.g.
`levenshtein.feature`) have no such cross-layer tags either, so this is a
project-wide gap that should be raised with the test-writer, not blocked on
here. Noted as NIT.

---

### BDD-07 `WithoutNormalisation` scenario uses relative assertion; no concrete value pinned

**Severity:** NIT
**Feature file:** `tests/bdd/features/scorer.feature` lines 90-102

The scenario for `WithoutNormalisation` asserts:

```gherkin
Then the no-normalisation composite should be less than the default composite
```

This is a relative assertion (no-norm < default). It correctly documents the
expected ordering but does not pin a concrete value. If a refactoring changed
both scores in the same direction (e.g. both decreased proportionally), the
relative assertion would still pass while the absolute behaviour had changed.

The golden file `testdata/golden/scorer-default.json` is supposed to contain
a `WithoutNormalisation` entry (CONTEXT.md §6 mandatory row #8: "XMLParser" /
"xml_parser" no-norm variant). Pinning the concrete scores in the BDD scenario
would align the BDD contract with the golden file contract. This is a NIT
because the relative assertion is not wrong — it documents the intended
semantics (normalisation improves score on identifier-style inputs) without
being fragile to floating-point variation in the algorithms. But a concrete
range would be stronger.

**Suggested addition:**

```gherkin
And the no-normalisation composite should be less than 0.80
And the default composite should be greater than 0.90
```

(Exact threshold values depend on the actual algorithm outputs for
`XMLParser`/`xml_parser` through the six DefaultScorer algorithms — pin from
the golden file.)

---

## Coverage Checklist vs CONTEXT.md §7

| Class | Required scenario | Status |
|-------|------------------|--------|
| 1 | Default scorer happy path | COVERED — scenario 1 |
| 2 | Default scorer below threshold | COVERED — scenario 2 |
| 3 | Custom 1-algorithm Scorer | COVERED — scenario 3 (asserts Match=true, not raw score) |
| 4 | Custom 2-algorithm Scorer with weights | WEAK — BDD-03: equal weights on identity pair; normalisation not actually exercised |
| 5 | `WithoutAlgorithm` composition | COVERED — scenario 5 (asserts Algorithms() excludes DM) |
| 6 | Last-write-wins duplicate AlgoID | COVERED — scenario 6 (asserts length + weight) |
| 7 | `WithoutNormalisation` | COVERED — scenario 7 (relative assertion; BDD-07 NIT) |
| 8 | `ErrMissingThreshold` | COVERED — scenario 8 |
| 9 | `ErrEmptyScorer` | COVERED — scenario 9 |
| 10 | `ErrInvalidWeight` | COVERED — scenario 10 |
| 11 | Concurrent safety + goleak gate | COVERED — scenario 11 |
| 12 | `ScoreAll` AlgoID keys | WEAK — BDD-05: values not range-checked |
| — | `ErrInvalidThreshold` | MISSING — BDD-01 BLOCKING |
| — | Weight auto-normalisation (unequal raw weights) | MISSING — BDD-03 BLOCKING |
| — | `WithNormalisation(custom opts)` | MISSING — BDD-04 IMPORTANT |

---

## Infrastructure Assessment

**goleak gate:** WIRED. `tests/bdd/bdd_test.go:37` calls
`goleak.VerifyTestMain(m)` before the suite runs. The concurrent scenario uses
`sync.WaitGroup` (no goroutine escapes the `wg.Wait()` call); goleak would
catch any regression that introduced unjoined goroutines.

**go.mod isolation:** CORRECT. `tests/bdd/go.mod` lists godog v0.15.0, goleak
v1.3.0, testify v1.10.0. The root `go.mod` has zero non-stdlib require lines
(the replace directive in `tests/bdd/go.mod` points to `../..`). This matches
the spec-locked constraint in `docs/requirements.md §15.11`.

**ScorerContext isolation:** CORRECT. `InitScorerSteps` creates a fresh
`&ScorerContext{}` per scenario (`tests/bdd/steps/scorer_steps.go:498`). State
does not leak across scenarios.

**Step structure:** mostly clean, but `iListTheScorerAlgorithms` (line 222-227)
is a no-op — it validates `sc.scorer != nil` and returns nil. The "When I list
the Scorer algorithms" step does no work; the subsequent assertion steps call
`sc.scorer.Algorithms()` directly. This is not a bug (the step is a
readability bridge), but it means the "When" step does not capture state that
the "Then" steps could read, making the Given/When/Then partition slightly
misleading. The assertion steps directly call `sc.scorer.Algorithms()` on each
call rather than storing the result once — this is safe (each call returns a
fresh slice) but is worth noting.

**Determinism of concurrent scenario:** CORRECT. 100 goroutines writing to
pre-allocated `sc.concurrentResults[i]` by index with no shared counter avoids
data races. The `wg.Wait()` join before `allGoroutineResultsShouldBeByteIdentical`
ensures all goroutines have completed before assertion.

---

## Summary Count

| Severity | Count |
|----------|-------|
| BLOCKING | 3 (BDD-01, BDD-02, BDD-03) |
| IMPORTANT | 2 (BDD-04, BDD-05) |
| NIT | 2 (BDD-06, BDD-07) |
| **Total** | **7** |

**Recommendation: NO-GO**

BDD-02 is a confirmed step bug (hardcoded pair) that causes a scenario to
silently exercise the wrong inputs. BDD-01 is a missing scenario for a
documented, exercisable sentinel error whose step infrastructure is already
present. BDD-03 is a weak scenario that does not actually exercise the
documented class (weighted composite with non-trivial normalisation). These
three must be resolved before the BDD suite can be considered a valid contract
for the Phase 8 surface.

IMPORTANT items (BDD-04, BDD-05) should be resolved in the same PR to avoid
a follow-up cycle, but they do not block a green build on their own.

_Reviewed: 2026-05-17_
_Reviewer: bdd-scenario-reviewer_
