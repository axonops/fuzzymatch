---
phase: 08-composite-scorer
review_type: security
reviewed: 2026-05-17T00:00:00Z
reviewer: gsd-security-reviewer
depth: standard
files_reviewed: 14
files_reviewed_list:
  - scorer.go
  - scorer_options.go
  - scorer_test.go
  - scorer_internal_test.go
  - scorer_bench_test.go
  - errors.go
  - normalise.go
  - monge_elkan.go
  - tversky.go
  - cosine.go
  - qgram_jaccard.go
  - sorensen_dice.go
  - ratcliff_obershelp.go
  - partial_ratio.go
findings:
  critical: 2
  high: 3
  medium: 3
  low: 4
  total: 12
status: issues_found
---

# Phase 8: Security Review

**Reviewed:** 2026-05-17
**Reviewer:** gsd-security-reviewer
**Depth:** standard (focus: DoS via pathological inputs, algorithmic
complexity attacks, panic safety on malformed UTF-8)
**Status:** issues_found

## Summary

The Phase 8 Scorer composes up to 22 algorithms with worst-case
complexities ranging from O(n) (Hamming) to O(N²·M) (Ratcliff-Obershelp)
to O(|s|·|l|·max(|s|,|l|)) (Partial Ratio). A single `Score` call to a
6-algorithm `DefaultScorer` triggers six algorithm dispatches plus a pre-
normalisation pass — every input-driven slowdown is six-times-multiplied
by construction.

Two **CRITICAL** panic-on-consumer-input paths exist, both already on
the team's radar via 08-REVIEW.md but security-relevant enough to call
out as blocking from the security perspective:

- **SEC-01:** `WithTverskyAlgorithm(weight, 0, 0, n)` is accepted at
  construction time but panics inside `TverskyScore` on the first
  `Score`/`Match`/`ScoreAll` call (`tversky.go:241`). The Scorer's
  documented contract is "fail loudly at construction"; this is a
  panic-from-consumer-supplied-input vulnerability.
- **SEC-02:** `WithMongeElkanAlgorithm(weight, inner)` accepts any inner
  AlgoID in the dispatch table but `MongeElkanScoreSymmetric` panics at
  Score time if `inner` is not in the 18-entry allow-list
  (`monge_elkan.go:382`). Passing e.g. `AlgoTokenSortRatio` is accepted
  by the option layer and panics at first invocation.

Three additional **HIGH**-severity concerns relate to NaN/Inf propagation
(SEC-03, SEC-04) and the absence of a Scorer-level fuzz harness (SEC-05).
The remaining findings cover DoS-via-pathological-input documentation
gaps, an unbounded recursion concern on Ratcliff-Obershelp, and several
defence-in-depth suggestions.

The library itself is goroutine-free, init-free, and has no concurrency
primitives, so there is no race-condition surface and no time-of-check-
time-of-use ambiguity. `Scorer` is correctly immutable after `NewScorer`
returns (verified at scorer.go:91-125 — every field unexported; no method
writes the receiver). `Algorithms()` returns a fresh slice with no
internal scoreFn exposed (scorer.go:460-466 — confirms only ID + Weight
are surfaced).

Map-iteration discipline on output paths is preserved: `ScoreAll`
iterates the AlgoID-sorted slice to populate its return map
(scorer.go:512-516); `Algorithms()` walks the same slice
(scorer.go:461-465); no error message embeds user input.

---

## CRITICAL — panic from consumer-supplied input

### SEC-01: `WithTverskyAlgorithm` permits α+β==0 → panic on first Score call

**File:** `/Users/johnny/Development/fuzzymatch/scorer_options.go:381-399`
**Severity:** CRITICAL (panic vulnerability; consumer-controlled input)

The Tversky option validates `alpha < 0 || beta < 0` but does NOT
validate the `α + β > 0` constraint. `TverskyScore`
(`tversky.go:241-243`) panics on this case:

```go
if alpha < 0 || beta < 0 || (alpha == 0 && beta == 0) {
    panic("fuzzymatch: invalid tversky parameter")
}
```

Reproducer:

```go
s, err := fuzzymatch.NewScorer(
    fuzzymatch.WithTverskyAlgorithm(1.0, 0, 0, 3),
    fuzzymatch.WithThreshold(0.5),
)
// err == nil; s != nil
s.Score("abc", "abc")  // PANIC at first call
```

Note the identity short-circuit at `tversky.go:231` (`if a == b { return
1.0 }`) hides the panic until non-identical inputs arrive — the
construction-time test suite may pass while production identifier-style
mismatches trigger the panic at runtime. This makes the issue latent.

The Scorer's documented contract (08-CONTEXT.md §2, scorer_options.go
file header, plus every other With\*Algorithm option) is "fail loudly at
construction with a typed sentinel". `ErrInvalidTverskyParam` already
exists in errors.go for exactly this case; the option layer just fails
to use it.

**Fix:**

```go
func WithTverskyAlgorithm(weight, alpha, beta float64, n int) ScorerOption {
    return func(cfg *scorerConfig) error {
        if weight <= 0 {
            return ErrInvalidWeight
        }
        if n < 1 {
            return ErrInvalidQGramSize
        }
        if alpha < 0 || beta < 0 || (alpha == 0 && beta == 0) {
            return ErrInvalidTverskyParam
        }
        // ...
    }
}
```

Add `TestWithTverskyAlgorithm_RejectsBothZero` and a BDD scenario.

This is also flagged in 08-REVIEW.md CR-02; reproducing here because a
panic-from-consumer-input is a security-grade defect even if functional
review classes it as a correctness defect.

---

### SEC-02: `WithMongeElkanAlgorithm` permits non-allowlisted inner → panic on first Score call

**File:** `/Users/johnny/Development/fuzzymatch/scorer_options.go:425-446`
**Severity:** CRITICAL (panic vulnerability; consumer-controlled input)

The Monge-Elkan option checks dispatch-table bounds and the self-
recursion case but does NOT consult `permittedMongeElkanInner`
(`monge_elkan.go:291-317`). Five AlgoIDs are valid `int(inner) <
numAlgorithms` AND have `dispatch[inner] != nil` AND are not
`AlgoMongeElkan`, yet are rejected by `MongeElkanScoreSymmetric` at
runtime via panic at `monge_elkan.go:382`:

- `AlgoTokenSortRatio`
- `AlgoTokenSetRatio`
- `AlgoPartialRatio`
- `AlgoTokenJaccard`

Reproducer:

```go
s, _ := fuzzymatch.NewScorer(
    fuzzymatch.WithMongeElkanAlgorithm(1.0, fuzzymatch.AlgoTokenSortRatio),
    fuzzymatch.WithThreshold(0.5),
)
s.Score("hello world", "hello there")  // PANIC at first call
```

Like SEC-01, this is hidden by the identity short-circuit
(`monge_elkan.go:387`); the panic surfaces only on non-identical input,
making it a latent runtime failure.

The fix-deferred decision is already tracked as IN-03 in 08-REVIEW.md;
from a security perspective the option-layer-vs-Score-time-panic
discrepancy is the panic-surface concern.

**Fix (recommended):**

```go
func WithMongeElkanAlgorithm(weight float64, inner AlgoID) ScorerOption {
    return func(cfg *scorerConfig) error {
        if weight <= 0 {
            return ErrInvalidWeight
        }
        if int(inner) < 0 || int(inner) >= numAlgorithms || dispatch[inner] == nil {
            return ErrInvalidAlgorithm
        }
        if !permittedMongeElkanInner[inner] {
            return ErrInvalidAlgorithm
        }
        // ...
    }
}
```

This mirrors the panic-shouldn't-cross-the-API-boundary discipline used
in every other option, at the cost of exporting the ME allow-list across
the package-internal boundary (the map is already package-scoped in
`monge_elkan.go:291`, so no new export is needed).

---

## HIGH — DoS / data integrity

### SEC-03: NaN slips through `WithThreshold` → Match silently returns false on every input

**File:** `/Users/johnny/Development/fuzzymatch/scorer_options.go:257-266`
**Severity:** HIGH (silent malfunction, denial-of-service via silent
no-match)

`WithThreshold(t)` checks `t < 0.0 || t > 1.0`; both comparisons evaluate
`false` for `t = math.NaN()`. The NaN threshold is then frozen into the
Scorer, and `Match(a, b)` returns `s.Score(a, b) >= s.threshold` — and
`x >= NaN` is always `false`. The Scorer never matches anything.

From a security perspective this is a **denial-of-service-via-silent-
failure** — a consumer who constructs a Scorer from configuration where
NaN can slip in (e.g. a JSON-decoded YAML value, an environment-variable
arithmetic accident, a `math.Sqrt(-1)` mistake) gets a Scorer that
silently misclassifies every input as non-matching. There is no error,
no warning, no metric, no log. The docs at `docs/scorer.md:283` claim
NaN is rejected; the code does not match.

This is 08-REVIEW.md CR-01; surfaced here at HIGH severity because of
the silent-failure mode (worse than a panic — at least a panic surfaces;
silent wrong-answer corrupts every downstream decision).

**Fix:**

```go
import "math"

func WithThreshold(t float64) ScorerOption {
    return func(cfg *scorerConfig) error {
        if math.IsNaN(t) || t < 0.0 || t > 1.0 {
            return ErrInvalidThreshold
        }
        cfg.threshold = t
        cfg.thresholdSet = true
        return nil
    }
}
```

Add `TestWithThreshold_RejectsNaN` plus a fuzz-corpus seed for
`math.NaN()` in any future Scorer fuzz harness (see SEC-05).

---

### SEC-04: NaN/+Inf slip through `WithAlgorithm` weight → Score returns NaN

**File:** `/Users/johnny/Development/fuzzymatch/scorer_options.go:150-165`
(WithAlgorithm); the same defect pattern in every parameterised With\*
option (lines 300-315, 325-340, 350-365, 381-399, 425-446, 465-479)
**Severity:** HIGH (silent malfunction; composite Scorer output becomes
NaN under default normalisation, propagates through Match and ScoreAll)

The weight gate `if weight <= 0` evaluates `false` for both
`math.NaN()` and `math.Inf(+1)`. The poisoned weight propagates through
auto-normalisation:

- `weight = math.NaN()`: `sum = sum + NaN = NaN`. The defensive `sum
  == 0` gate at `scorer.go:284` returns `false` for NaN sums (the
  comment at scorer.go:269-273 explicitly mentions "the divisor never
  produces NaN/Inf" — but the gate does not actually catch NaN).
  Every normalised weight becomes `weight/NaN = NaN`. Every `Score`
  call returns NaN. Every `Match` returns false. Every `ScoreAll`
  populates the result map with NaN values.
- `weight = math.Inf(1)`: `sum = +Inf`. Normalised weight = `Inf/Inf =
  NaN`. Same downstream poison as above.
- `weight = math.Inf(-1)`: caught by `weight <= 0` and rejected — fine.

Reproducer:

```go
s, _ := fuzzymatch.NewScorer(
    fuzzymatch.WithAlgorithm(fuzzymatch.AlgoLevenshtein, math.NaN()),
    fuzzymatch.WithThreshold(0.5),
)
score := s.Score("kitten", "kitten")  // NaN
match := s.Match("kitten", "kitten")  // false (NaN >= 0.5 is false)
```

The property test `TestProp_Scorer_NoNaN_NoInf` (scorer_test.go:877-887)
verifies Score returns finite values for *random string inputs against
`DefaultScorer()`*, but does NOT exercise the NaN/Inf-weight construction
path — so the test passes despite the defect.

**Fix:**

```go
func WithAlgorithm(algo AlgoID, weight float64) ScorerOption {
    return func(cfg *scorerConfig) error {
        if math.IsNaN(weight) || math.IsInf(weight, 0) || weight <= 0 {
            return ErrInvalidWeight
        }
        // ...
    }
}
```

Apply the same pattern to every parameterised With\* option's weight
gate. Tighten `scorer.go:284` to `if sum <= 0 || math.IsNaN(sum) ||
math.IsInf(sum, 0)` as defence-in-depth — even with the option-layer
fix, an explicit guard at the normalisation point catches any future
weight-injection path.

Add `TestWithAlgorithm_RejectsNaNWeight` and `TestWithAlgorithm_RejectsInfWeight`
plus a property test that asserts no weight in
`s.algorithmsAlgoIDSorted` is non-finite after construction.

---

### SEC-05: No Scorer-level fuzz harness covering the aggregated 6-algorithm dispatch path

**File:** missing — there is no `scorer_fuzz_test.go`
**Severity:** HIGH (test-coverage gap; every individual algorithm is
fuzz-covered but the aggregate Scorer surface is not)

Every catalogue algorithm has a dedicated `Fuzz*` harness
(`testdata/fuzz/Fuzz*` lists 26 fuzzers), and `FuzzNormalise` covers the
Normalise pipeline in isolation. But there is no fuzz harness exercising
`DefaultScorer().Score(a, b)` end-to-end. This matters because:

1. The Scorer chains Normalise + six algorithms in a single call.
   Inter-algorithm interactions are not fuzz-covered. A panic that
   only surfaces under a specific Normalise(NFC-fold) → DM-encode →
   TokenJaccard tokenise chain would not be caught by the per-algorithm
   harnesses.
2. SEC-01 and SEC-02 are panics that surface ONLY when the Scorer
   reaches Score-time. A Scorer-level fuzz harness with a corpus
   covering several Scorer compositions (`DefaultScorer`,
   `WithTverskyAlgorithm`, `WithMongeElkanAlgorithm`) would catch
   these classes of bug automatically. Today the panic is verified
   only by hand-authored unit tests.
3. The "Scorer is panic-free on consumer input" claim is currently
   verified by induction-over-individual-algorithms; a fuzz harness
   makes it directly testable.

**Fix:** Add `scorer_fuzz_test.go` with three harnesses:

```go
// FuzzScorer_DefaultScorer_NeverPanics asserts Score, Match, and
// ScoreAll never panic on arbitrary string inputs against the
// opinionated default composition.
func FuzzScorer_DefaultScorer_NeverPanics(f *testing.F) {
    f.Add("", "")
    f.Add("kitten", "sitting")
    f.Add("café", "cafe")
    f.Add(string([]byte{0xff, 0xfe, 0xfd}), "valid utf8")  // lone-surrogate path
    s := fuzzymatch.DefaultScorer()
    f.Fuzz(func(t *testing.T, a, b string) {
        _ = s.Score(a, b)
        _ = s.Match(a, b)
        _ = s.ScoreAll(a, b)
    })
}

// FuzzScorer_NewScorer_NeverPanics asserts NewScorer never panics on
// any combination of legal option values (catches future regressions
// like SEC-01 and SEC-02 when new options ship).
func FuzzScorer_NewScorer_NeverPanics(f *testing.F) {
    f.Fuzz(func(t *testing.T, threshold, weight float64, algoIdx uint8) {
        algo := fuzzymatch.AlgoID(int(algoIdx) % 23)
        _, _ = fuzzymatch.NewScorer(
            fuzzymatch.WithAlgorithm(algo, weight),
            fuzzymatch.WithThreshold(threshold),
        )
    })
}

// FuzzScorer_AggregatePathScoreInRange asserts that every successful
// construction produces Score in [0,1] regardless of input — a
// stronger property than the existing PropScorer_ScoreInRange because
// it exercises non-default compositions.
```

Seed the corpus with embedded NUL, lone surrogates, multi-MB inputs,
RTL marks, and zero-width joiners. Wire it into `make test-fuzz` and
the nightly fuzz workflow.

---

## MEDIUM — DoS via pathological input

### SEC-06: Monge-Elkan token-count DoS multiplied by Scorer dispatch — no Scorer-boundary input-size bound

**File:** `/Users/johnny/Development/fuzzymatch/monge_elkan.go:258-267`
(documented), `/Users/johnny/Development/fuzzymatch/scorer.go:349-383`
(applied)
**Severity:** MEDIUM (algorithmic-complexity DoS; the algorithm
documents the cost, but the Scorer layer does not surface it)

Monge-Elkan's complexity is O(|tA|·|tB|·cost(inner)). On a 1000-token
input pair with Jaro-Winkler inner this is ~10^7 character operations
(monge_elkan.go:259-267 documents this explicitly with a "DoS notice"
section). When Monge-Elkan participates in a Scorer composition, the
cost is multiplied by 1 (it's a single entry in the reduction) but the
*latency* of the full `Score` call rises by that amount.

The DefaultScorer composition does NOT include Monge-Elkan (the six are
DamerauLevenshteinOSA / JaroWinkler / TokenJaccard / QGramJaccard /
SorensenDice / DoubleMetaphone — see scorer.go:544-552), so the DoS is
opt-in. However:

1. `docs/scorer.md` does NOT document the per-algorithm DoS profile
   of compositions including ME, PartialRatio, RatcliffObershelp,
   SWG, LCSStr, Levenshtein, or DamerauLevenshteinFull.
2. `docs/performance.md` is a scaffold (60 lines, all "TBD") and
   does NOT document input-size-bound guidance for any algorithm —
   the security-reviewer focus area "the `docs/performance.md`
   discusses how to bound input size before invoking an algorithm"
   is explicitly missing.
3. No Scorer-boundary input-size bound exists. A consumer feeding
   100KB-vs-100KB input to a custom Scorer that includes SWG (O(mn))
   and RatcliffObershelp (O(N²·M)) burns ~10^10 cell updates per
   Score call — a single ScoreAll call could lock a goroutine for
   minutes.

**Fix:**

1. Add a `## DoS / Resource Bounds` section to `docs/scorer.md`
   listing the per-algorithm worst-case complexity (cross-reference
   each algorithm file's existing complexity docstring) with a clear
   "Pre-validate input length before calling Score on untrusted input"
   warning at the top.
2. Populate `docs/performance.md` with the input-size-bound table
   that the security-reviewer focus area cites — at minimum a
   "recommended input length ceiling per algorithm" mini-table for
   the six algorithms with super-linear complexity.
3. Consider a soft input-size guard option:
   `WithMaxInputBytes(n int) ScorerOption` that returns an error
   from Score when either input exceeds `n` bytes. Optional — only
   if api-ergonomics-reviewer agrees. The reviewer's current
   position (per Phase 8 surface) is to keep Scorer surface minimal;
   the soft guard is defence-in-depth, not BLOCKING.

---

### SEC-07: Ratcliff-Obershelp uses unbounded Go call-stack recursion

**File:** `/Users/johnny/Development/fuzzymatch/ratcliff_obershelp.go:200-210`
**Severity:** MEDIUM (theoretical stack-overflow on pathological multi-
MB inputs; in practice mitigated by Go's growable goroutine stacks but
the security-reviewer focus area explicitly calls out "no algorithm has
unbounded recursion")

`roMatchedLength` recurses via:

```go
return n + roMatchedLength(a[:aLo], b[:bLo]) + roMatchedLength(a[aHi:], b[bHi:])
```

with depth bounded by O(min(la, lb)) per the file header
(ratcliff_obershelp.go:81-83 + 109-114). On a multi-MB pathological
input (e.g. all-'a' strings with strategic differences forcing the
recursion to maximum depth at every level), the call stack can balloon
to millions of frames.

Go's growable goroutine stack means a literal stack-overflow panic is
unlikely under normal limits, but:

1. The security-reviewer focus area explicitly states "Ratcliff-
   Obershelp's recursive longest-common-substring decomposition
   must use iterative or bounded-depth recursion".
2. The 23-algorithm `algorithm-complexity attacks` threat enumeration
   in CLAUDE.md names Ratcliff-Obershelp under quadratic-or-worse
   complexity.
3. A pathological corpus row in a future `FuzzRatcliffObershelpScore`
   run (or in the missing SEC-05 Scorer-level fuzz) could trigger
   stack growth to gigabytes before resolution.

This is BELOW the threshold of a real-world panic vulnerability — the
Go runtime grows the goroutine stack up to ~1GB before failing, and
each recursion frame is small. But it does violate the security-
reviewer skill's explicit "no unbounded recursion" requirement.

**Fix (defence-in-depth):**

Convert `roMatchedLength` to an iterative form with an explicit
work-queue:

```go
type roWork struct {
    aLo, aHi, bLo, bHi int  // alternatively pass slices
}

func roMatchedLength(a, b string) int {
    var stack []roWork
    stack = append(stack, roWork{0, len(a), 0, len(b)})
    total := 0
    for len(stack) > 0 {
        w := stack[len(stack)-1]
        stack = stack[:len(stack)-1]
        as, bs := a[w.aLo:w.aHi], b[w.bLo:w.bHi]
        if len(as) == 0 || len(bs) == 0 {
            continue
        }
        aLo, aHi, bLo, bHi, n := roFindLongestMatch(as, bs)
        if n == 0 {
            continue
        }
        total += n
        // Push both sub-problems; order does not affect the sum.
        stack = append(stack,
            roWork{w.aLo, w.aLo + aLo, w.bLo, w.bLo + bLo},
            roWork{w.aLo + aHi, w.aHi, w.bLo + bHi, w.bHi},
        )
    }
    return total
}
```

Or document the recursion-depth bound and add a fuzz-corpus seed
designed to maximise depth, asserting wall-time stays below a
configured budget. The latter is the lighter touch.

---

### SEC-08: Scorer reduction does not guard against NaN scores from individual algorithms

**File:** `/Users/johnny/Development/fuzzymatch/scorer.go:368-381`
**Severity:** MEDIUM (defence-in-depth; the algorithms today are
NaN-free but a future algorithm or a future code-path could introduce
a NaN that propagates silently)

The reduction loop:

```go
var acc float64
for _, entry := range s.algorithmsAlgoIDSorted {
    score := entry.scoreFn(na, nb)
    acc = acc + (entry.weight * score)
}
return acc
```

does not check that `score` is finite. The contract (every algorithm
returns a value in [0,1]) is verified per-algorithm by property tests,
but a future algorithm regression that returned NaN under some corner
case would propagate silently — `acc + (w * NaN) = NaN`. Match becomes
silently false; ScoreAll's map gets a NaN value.

The risk today is low because every algorithm's property test asserts
the [0,1] range. But if a future algorithm regresses (e.g. a division-
by-zero corner case in a new Phase 9+ algorithm; a NFC transformer bug
that produces an empty-but-not-empty string), the Scorer carries the
NaN through with no signal.

**Fix (defence-in-depth):** Add an optional `WithStrictRange()` option
that, when enabled, asserts each `score ∈ [0, 1]` and returns 0.0 (or
panics — discuss with api-ergonomics-reviewer) on violation. Not
required for v0.x but worth tracking for v1.0.

Alternatively: a property test
`TestProp_Scorer_NoNaN_NoInf_AllCompositions` that iterates every
single-algorithm Scorer (one per AlgoID), runs 100 random pairs through
Score, and asserts finiteness — this would catch a per-algorithm
regression at the Scorer surface.

---

## LOW — defence-in-depth

### SEC-09: `nil` ScorerOption in opts slice panics with nil-function dereference

**File:** `/Users/johnny/Development/fuzzymatch/scorer.go:199-203`
**Severity:** LOW (consumer error; not a real-world attack surface, but
the panic message is unhelpful)

```go
for _, opt := range opts {
    if err := opt(&cfg); err != nil {
        return nil, err
    }
}
```

If a consumer passes a literal `nil` (e.g. `NewScorer(nil,
WithThreshold(0.5))`) or accidentally mutates a `DefaultScorerOptions()`
slice such that an element becomes nil, the loop calls `nil(&cfg)`
which panics with "runtime error: invalid memory address or nil pointer
dereference". `TestDefaultScorerOptions_FreshSlice` at scorer_test.go:593
does `opts1[0] = nil` (testing the fresh-slice invariant) but does NOT
pass the mutated slice back to `NewScorer` — so the nil-handling defect
is not exercised by the test suite.

This is consumer error rather than malicious input, but the panic is
cryptic. A typed `ErrInvalidOption` (or just a skip-nil-with-defer-style
nil-guard) would surface a useful error.

**Fix:**

```go
for i, opt := range opts {
    if opt == nil {
        return nil, fmt.Errorf("fuzzymatch: nil option at index %d: %w", i, ErrInvalidConfiguration)
    }
    if err := opt(&cfg); err != nil {
        return nil, err
    }
}
```

(`ErrInvalidConfiguration` already exists in errors.go:57.)

---

### SEC-10: `DefaultScorer()` panic on internal inconsistency — bounded scope verified

**File:** `/Users/johnny/Development/fuzzymatch/scorer.go:586-592`
**Severity:** LOW (intentional fail-loud; verified bounded)

`DefaultScorer()` panics if `NewScorer(DefaultScorerOptions()...)`
returns an error. The panic is documented and intentional — Phase 7
populates all 23 dispatch slots, the weights are 1.0 each, and the
threshold lies in [0,1], so the construction "cannot fail under normal
operation".

Verified bounded:

- `DefaultScorerOptions()` returns six fixed `WithAlgorithm` calls plus
  `WithThreshold(0.85)`. None of the dispatch entries (AlgoLevenshtein
  through AlgoDoubleMetaphone) can be unregistered at runtime (they
  are package-load-time bound via `var _ = func() bool { … }()`
  idiom).
- The threshold 0.85 is a literal, in [0,1].
- The weights 1.0 are literals, > 0.

The panic is genuinely unreachable under any consumer input pattern.
Good defensive coding. No action required.

---

### SEC-11: `WithoutAlgorithm` allocation slicing reuses backing array — no aliasing risk

**File:** `/Users/johnny/Development/fuzzymatch/scorer_options.go:190-198`
**Severity:** LOW (informational; checked for the documented in-place
compaction)

```go
filtered := cfg.entries[:0]
for _, e := range cfg.entries {
    if e.id != id {
        filtered = append(filtered, e)
    }
}
cfg.entries = filtered
```

Verified safe: the write index never outruns the read index in
`scan-and-compact`. No aliasing of an external slice into `cfg.entries`
exists — `cfg` is a stack-local in `NewScorer` and the entries slice
is fully owned. The stale comment (08-REVIEW.md WR-03 — claims reverse
iteration, code iterates forward) is a documentation issue not a
security one.

No action from security perspective.

---

### SEC-12: Error messages do not embed user input

**File:** `/Users/johnny/Development/fuzzymatch/errors.go` (all sentinels)
**Severity:** LOW (informational; spot-check passed)

All Phase 8 sentinel errors are flat `errors.New("fuzzymatch: …")`
values with no `fmt.Errorf` wrapping at the option/NewScorer layer
that would embed user-supplied strings (no `%s` / `%q` formatting of
`a`, `b`, or option values). The error-leakage focus area is therefore
satisfied.

Spot-check confirms:

- `errors.go:48-162` — every sentinel is a flat `errors.New`.
- `scorer.go:200-298` — NewScorer returns sentinels verbatim, never
  wraps with user-supplied data.
- `scorer_options.go:150-479` — every option returns a sentinel
  verbatim, never wraps with user-supplied data.

No action required.

---

## Verification of focus-area requirements

| Focus area | Status | Notes |
|------------|--------|-------|
| Every algorithm documents complexity in godoc | PARTIAL | Algorithms have complexity sections; Scorer composition does not surface aggregate complexity to consumers — SEC-06 |
| Super-linear complexity algorithms documented with warnings | PARTIAL | Per-algorithm docs OK (ME, PartialRatio explicit); docs/scorer.md missing — SEC-06 |
| Fuzz tests include multi-KB inputs | PARTIAL | Per-algorithm yes; Scorer aggregate path missing — SEC-05 |
| docs/performance.md discusses input-size bounding | NO | File is a 60-line scaffold; all sections "TBD" — SEC-06 |
| No unbounded recursion | PARTIAL | Ratcliff-Obershelp uses Go call stack with O(min(la,lb)) depth — SEC-07 |
| Public functions never panic on arbitrary input | NO | SEC-01 (Tversky α+β==0), SEC-02 (ME non-allowlisted inner) |
| Property test `PropAlgorithm_NeverPanics` exists | PARTIAL | `TestProp_Scorer_NoNaN_NoInf` covers DefaultScorer; SEC-05 fills the gap for non-default compositions |
| Error messages do not embed user input | YES | SEC-12 confirmed |
| No timing-based information leakage | N/A | No secrets handled |
| Zero runtime dependencies | YES | `go.mod` has zero non-stdlib `require` lines (verified externally) |
| Test deps isolated in tests/bdd/go.mod | YES | Phase 8 testify use is BDD-only |
| `govulncheck ./...` clean | OUT OF SCOPE | CI verifies; no Phase 8 deps changed |
| No GPL/LGPL-derived code | YES | Phase 8 has no algorithm work; carry-forward from Phase 1-7 reviews |
| Regex safety | N/A | No regex use in scorer.go / scorer_options.go |
| Invalid UTF-8 graceful | YES | Normalise replaces with U+FFFD per Go convention (normalise.go:82) |
| Embedded NULs | YES | Algorithms compare byte-by-byte; no NUL-specific gate exists |

---

## Recommendations summary

**Must fix before v1.0 (CRITICAL + HIGH):**

1. SEC-01 — `WithTverskyAlgorithm` α+β==0 panic
2. SEC-02 — `WithMongeElkanAlgorithm` non-allowlisted inner panic
3. SEC-03 — `WithThreshold` NaN handling
4. SEC-04 — `WithAlgorithm` NaN/Inf weight handling
5. SEC-05 — Scorer-level fuzz harness

**Should fix before v1.0 (MEDIUM):**

6. SEC-06 — Document DoS profile in docs/scorer.md + docs/performance.md
7. SEC-07 — Iterative Ratcliff-Obershelp (or document bound + add corpus
   seed)
8. SEC-08 — Add property test for per-AlgoID single-Scorer NaN/Inf

**Defence-in-depth (LOW):**

9. SEC-09 — nil-option guard with `ErrInvalidConfiguration` wrap
10. SEC-10 — verified bounded, no action
11. SEC-11 — verified safe, no action
12. SEC-12 — confirmed clean, no action

**Test-coverage gaps (also flagged in 08-REVIEW.md):**

- No `scorer_fuzz_test.go` — SEC-05
- No NaN-weight unit test — SEC-04
- No NaN-threshold unit test — CR-01 in 08-REVIEW.md, SEC-03 here
- No α+β==0 Tversky unit test — CR-02 in 08-REVIEW.md, SEC-01 here

---

_Reviewed: 2026-05-17_
_Reviewer: gsd-security-reviewer_
_Depth: standard_
