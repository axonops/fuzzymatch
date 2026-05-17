---
phase: 08-composite-scorer
reviewer: api-ergonomics-reviewer
reviewed: 2026-05-17T00:00:00Z
status: approved_with_changes
authority: final (Design Principle 13 — CLAUDE.md)
files_reviewed:
  - scorer.go
  - scorer_options.go
  - errors.go
  - scorer_test.go
  - scorer_options_test.go
  - scorer_internal_test.go
  - scorer_options_internal_test.go
  - examples/scorer-composition/main.go
  - examples/identifier-similarity/main.go
  - docs/scorer.md
  - docs/tuning.md
  - docs/requirements.md (§8)
  - .planning/phases/08-composite-scorer/08-CONTEXT.md
  - .planning/phases/08-composite-scorer/08-REVIEW.md
findings:
  blocking: 2
  strong_prefer: 6
  nit: 8
  total: 16
---

# Phase 8 — API Ergonomics Review

## Verdict

**`approved_with_changes`.**

The Phase 8 Scorer surface is, in the round, idiomatic, discoverable, and faithful to the project's three-layer progressive-disclosure model. The locked decisions in `08-CONTEXT.md` (mandatory threshold, `map[AlgoID]float64` SPEC OVERRIDE, AlgoID-sorted internal reduction, `WithoutAlgorithm` for default-minus-X composition, `DefaultScorerOptions()` returning a mutable slice) are all the right calls. The Layer 2a hello-world is 5 lines including imports — meeting the target. The Layer 2b custom-Scorer hello-world is one option longer than the spec illustration because of the mandatory `WithThreshold`, and that one extra line is defensible.

Two **BLOCKING** items must change before v1.0:

- **API-01** — `WithThreshold` accepts `NaN` (CR-01 from the code review). The Scorer then silently never matches anything. The documented contract in `docs/scorer.md` already says NaN is rejected; the code must catch up. This is correctness-via-API contract, not an opinion. Veto.
- **API-02** — `WithTverskyAlgorithm` accepts `(α=0, β=0)` and panics at `Score` time (CR-02). Every other `With*Algorithm` validates fully at option-application time. The deferral is inconsistent with the rest of the surface, the godoc admission that "typical use is satisfied" is not a defence, and the panic at `Score` time silently breaks the documented "fail loudly at construction" contract. Veto.

Six **STRONG-PREFER** items would improve the surface meaningfully but are not v1.0 blockers; recommended for resolution before the v1.0 freeze.

Eight **NIT** items are minor polish.

The SPEC OVERRIDE on `ScoreAll`'s return type (`map[AlgoID]float64` over `map[string]float64`) is **explicitly approved** and signed off here (API-09). The override is the right call and the documentation already records it correctly.

## Layer-by-Layer TTHW (Time To Hello World)

| Layer | Target (CLAUDE.md skill) | Actual | Status |
|-------|-------------------------|--------|--------|
| Layer 1 — single algorithm: `fuzzymatch.LevenshteinScore(a, b)` | 4 lines including imports | 4 lines | ✓ Met |
| Layer 2a — default Scorer | 5 lines | 5 lines | ✓ Met |
| Layer 2b — custom Scorer | "5 lines + threshold line" — bespoke target | 6 lines | ✓ Met (+1 line for mandatory `WithThreshold`, defensible) |

Layer 2b's mandatory `WithThreshold` is the right ergonomic call — see API-03 below for the rationale endorsement.

## BLOCKING findings (must fix before v1.0)

### API-01 — `WithThreshold` does not reject NaN [BLOCKING]

**Current** (`scorer_options.go:257-266`):

```go
func WithThreshold(t float64) ScorerOption {
    return func(cfg *scorerConfig) error {
        if t < 0.0 || t > 1.0 {
            return ErrInvalidThreshold
        }
        cfg.threshold = t
        cfg.thresholdSet = true
        return nil
    }
}
```

`math.NaN()` evaluates `t < 0.0` and `t > 1.0` both to `false`, so NaN passes the gate. Construction then succeeds with a NaN threshold, and `Match` (which returns `s.Score(a, b) >= s.threshold`) becomes `... >= NaN` which is always false. The Scorer silently fails to match anything — the worst-class API failure mode.

**Proposed:**

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

**Comparable:** Python `rapidfuzz` validates score cutoffs eagerly (`fuzz.process.extract(score_cutoff=...)` rejects NaN at the C-extension boundary). Java `fuzzywuzzy` does not expose a threshold parameter at construction; `Jamboree` likewise. There is no precedent for accepting NaN as a similarity threshold in any comparable library.

**Rationale:** Principle 1 (Pit of Success) — the "do nothing" path with `WithThreshold(math.NaN())` should error immediately, not silently break every `Match` call. Principle 4 (Idiomatic Go) — the documented contract in `docs/scorer.md:283` and `errors.go` says NaN is rejected; the implementation must match.

**Action:** apply CR-01's fix from `08-REVIEW.md`. Add `TestWithThreshold_RejectsNaN` to `scorer_options_test.go` asserting `errors.Is(err, ErrInvalidThreshold)`. No documentation change needed (docs are already correct).

---

### API-02 — `WithTverskyAlgorithm` defers α + β > 0 check to Score-time panic [BLOCKING]

**Current** (`scorer_options.go:381-399`):

```go
func WithTverskyAlgorithm(weight, alpha, beta float64, n int) ScorerOption {
    return func(cfg *scorerConfig) error {
        if weight <= 0 { return ErrInvalidWeight }
        if n < 1 { return ErrInvalidQGramSize }
        if alpha < 0 || beta < 0 { return ErrInvalidTverskyParam }
        // ... α + β == 0 NOT checked; passes through; panic at Score time.
        cfg.entries = append(cfg.entries, scorerEntry{...})
        return nil
    }
}
```

A consumer who writes `WithTverskyAlgorithm(0.5, 0, 0, 3)` (innocent typo: meant 1, 1) gets a successful `NewScorer`, then a panic on the first `Score` call — from inside `TverskyScore`'s own guard at `tversky.go:241`. The construction-time contract is broken, the godoc waves it off ("typical use is satisfied"), and the diagnostic the user sees is a panic stack rooted in algorithm code, not a typed error rooted in their option call.

**Proposed:**

```go
func WithTverskyAlgorithm(weight, alpha, beta float64, n int) ScorerOption {
    return func(cfg *scorerConfig) error {
        if weight <= 0 { return ErrInvalidWeight }
        if n < 1 { return ErrInvalidQGramSize }
        if alpha < 0 || beta < 0 || (alpha == 0 && beta == 0) {
            return ErrInvalidTverskyParam
        }
        cfg.entries = append(cfg.entries, scorerEntry{...})
        return nil
    }
}
```

(Same fix as CR-02 from `08-REVIEW.md`.)

**Comparable:** Python `rapidfuzz` enforces `tversky(s1, s2, alpha=..., beta=...)` parameter validation at the Python-C boundary; the standalone Go `xrash/smetrics` does not expose Tversky at all. The "all options validate at option-application time" pattern is the Go-functional-options norm (see `oklog/run`, `nats.io`'s `ConnOption`, `grpc-go`'s `DialOption`). Deferring to runtime panic for documented-rejection inputs is unusual.

**Rationale:** Principle 4 (Idiomatic Go — option layer is the validation boundary). Principle 5 (Naming consistency: every other `With*Algorithm` returns a typed sentinel for malformed params; `WithTverskyAlgorithm` should match). Also: the godoc admission ("this option does not re-check it because either α or β being > 0 is satisfied by the typical use cases") is the kind of documentation that ages badly — the typical-use defence collapses the first time a junior engineer types two zeroes by accident.

**Action:** apply CR-02's fix. Add `TestWithTverskyAlgorithm_RejectsBothZero` asserting `errors.Is(err, ErrInvalidTverskyParam)`. Update the godoc to remove the "typical use cases" defence and document the gate explicitly.

## STRONG-PREFER findings (recommend before v1.0)

### API-03 — Endorsement: mandatory `WithThreshold` for `NewScorer` is the right call [STRONG-PREFER, endorsement]

`NewScorer` returning `ErrMissingThreshold` when `WithThreshold` is absent is a meaningful one-line ergonomic cost (the custom-Scorer hello-world is 6 lines instead of 5) but the alternative defaults all fail Principle 1:

- Default `1.0` → silent "no matches found"
- Default `0.0` → silent "everything matches"
- Default `0.85` (inherited from `DefaultScorer`) → arbitrary and misleading for non-default compositions

The decision is correct and the diagnostic is excellent: the threshold check fires FIRST in the validation pipeline so a user who forgets `WithThreshold` AND has another option problem sees the unambiguous "you forgot the threshold" error first. **Approved as-is.**

The only ergonomic improvement candidate here would be a fluent `NewScorer(t, opts...)` signature where `t float64` is the threshold and `opts ...ScorerOption` are everything else — explicit positional argument for the most-load-bearing parameter, no possibility of forgetting it. **This is rejected** as a v1.0 change because (a) it diverges from the canonical Go functional-options pattern (which positions all configuration through options), (b) it asymmetrically privileges one option, and (c) the current `ErrMissingThreshold` diagnostic is already clear.

**Action:** none. Documented here so future readers see the decision rationale.

---

### API-04 — Parameterised algorithm options: parameter order is inconsistent across the family [STRONG-PREFER]

**Current** signatures (`scorer_options.go`):

```go
func WithQGramJaccardAlgorithm(weight float64, n int) ScorerOption
func WithSorensenDiceAlgorithm(weight float64, n int) ScorerOption
func WithCosineAlgorithm(weight float64, n int) ScorerOption
func WithTverskyAlgorithm(weight, alpha, beta float64, n int) ScorerOption
func WithMongeElkanAlgorithm(weight float64, inner AlgoID) ScorerOption
func WithSmithWatermanGotohAlgorithm(weight float64, params SWGParams) ScorerOption
```

`weight` is consistently first (good). The trailing parameters vary by algorithm (necessarily — n is a q-gram window, inner is an AlgoID, params is a struct). **The inconsistency** is `WithTverskyAlgorithm`'s parameter order: `(weight, alpha, beta, n)`. The natural mental model is "(weight, then everything else in the same shape as the underlying function)", which for `TverskyScore(a, b, n, alpha, beta)` would suggest `(weight, n, alpha, beta)`. The current `(weight, alpha, beta, n)` order requires the consumer to remember a parameter-shuffle.

**Proposed: align with the underlying score function's parameter order** — change to `WithTverskyAlgorithm(weight float64, n int, alpha, beta float64)`. The signature then reads "weight, q-gram size, asymmetry params" — same shape as `TverskyScore`'s `(a, b, n, alpha, beta)` minus the input strings.

**Or, simpler: extract a `TverskyParams` struct** and follow the SWG pattern:

```go
type TverskyParams struct {
    N     int
    Alpha float64
    Beta  float64
}

func WithTverskyAlgorithm(weight float64, params TverskyParams) ScorerOption
```

This generalises to a forward-compatible pattern: if any q-gram algorithm later gains additional parameters (smoothing, frequency-weighting), they go into the params struct without breaking the option signature.

**Comparable:** Python `rapidfuzz.distance.Tversky` exposes `(s1, s2, *, alpha, beta)` — keyword-only after the strings, no positional shuffle. Java `fuzzywuzzy` doesn't expose Tversky. The SWG precedent (`params SWGParams`) is already in place in this codebase; extending the pattern to Tversky is the consistent move.

**Rationale:** Principle 5 (Naming Consistency) — the With\*Algorithm family should be predictable. Principle 4 (Idiomatic Go) — params-struct-for-many-knobs is the canonical pattern (see `crypto/tls.Config`, `database/sql.TxOptions`, the SWG precedent already in the codebase).

**Action:** **decision deferred to v1.0 freeze.** This is API-breaking (signature change), so it must land before the v1.0 cut. Recommended path: introduce `TverskyParams` struct, change signature to `WithTverskyAlgorithm(weight float64, params TverskyParams)`. The q-gram algorithms (QGram/Cosine/Dice) can keep their `(weight, n int)` shape because n is the only parameter and the shape is consistent across the three; only Tversky deviates and only Tversky needs the struct.

---

### API-05 — `WithoutAlgorithm` / `WithoutNormalisation` naming reads idiomatically; affirmative spelling endorsed [STRONG-PREFER, endorsement]

The user's concern: does `WithoutAlgorithm` / `WithoutNormalisation` read naturally against the affirmative `With*` siblings? Does it match Go precedent?

**Yes on both counts.** The stdlib precedent is `context.WithoutCancel(parent context.Context) context.Context` (Go 1.21+), which is exactly this pattern: an affirmative `WithCancel` and a removal-shaped `WithoutCancel`. The fuzzymatch `WithAlgorithm` / `WithoutAlgorithm` pair mirrors this. The grammatical reading "construct a Scorer with these algorithms, without DoubleMetaphone, with this threshold" is natural English and produces a discoverable IDE-autocomplete cluster (typing `fuzzymatch.With` reveals `WithAlgorithm`, `WithoutAlgorithm`, `WithoutNormalisation`, etc. — adjacent in the alphabetised list).

**Alternative considered: `RemoveAlgorithm`, `DisableAlgorithm`, `ExcludeAlgorithm`.** All worse:

- `Remove*` implies a mutator — the option layer is purely additive, options are appended to a config that the constructor freezes. `Remove` is misleading.
- `Disable*` implies a toggleable runtime flag — the algorithm slot is genuinely absent from the final Scorer, not disabled-but-present. Misleading.
- `Exclude*` reads OK but is not the Go precedent.

**Action:** none. Naming approved.

---

### API-06 — Single `WithParameterised(AlgoID, ...interface{})` form considered and rejected [STRONG-PREFER, decision]

The user's concern: should the 6 parameterised `With*Algorithm` options be collapsed into a single variadic form `WithParameterised(AlgoID, ...interface{})`?

**No.** The current 6-function family is the right call:

1. **Type safety.** `WithParameterised(AlgoTversky, 1.0, 0.5, 0.5, 3)` requires every callsite to remember the positional layout of the underlying parameters and provides zero compile-time gate on argument-type errors (`WithParameterised(AlgoTversky, 1.0, 0.5, 0.5, 3.0)` — n as `float64` — compiles, then fails at runtime). The current `WithTverskyAlgorithm(1.0, 0.5, 0.5, 3)` rejects this at compile time.
2. **IDE discoverability.** Typing `fuzzymatch.With` and seeing `WithQGramJaccardAlgorithm`, `WithSorensenDiceAlgorithm`, etc. — each with their own parameter signature visible in hover-doc — is dramatically more discoverable than typing `WithParameterised` and having to consult external documentation for the per-algorithm parameter layout.
3. **`interface{}` (or `any`) is a Go-idiom anti-signal.** The stdlib uses `any` sparingly and almost never for typed-API parameters; the precedent for parameter-rich functional options is per-feature constructors (`grpc.WithKeepaliveParams(KeepaliveParameters{...})`, `nats.MaxReconnects(int)`, etc.).

The duplication cost of 6 nearly-identical option closures is real but small (six 8-line functions; the redundancy is mechanical and unlikely to grow because the catalogue is frozen at 23 algorithms).

**Action:** none. Per-algorithm `With*Algorithm` family approved.

---

### API-07 — `DefaultScorerOptions()` returning the option slice is the right pattern for "default minus X" [STRONG-PREFER, endorsement]

The user's concern: is `append(DefaultScorerOptions(), WithoutAlgorithm(...))...` the right pattern? Or should there be a `WithDefaults()` option for declarative composition?

**The current pattern is right.** Three reasons:

1. **Explicit shape.** `append(DefaultScorerOptions(), WithoutAlgorithm(AlgoX))` makes the option list visible at the call site — the consumer can see exactly which 6 algorithms are in the baseline and which one is being removed. A hypothetical `NewScorer(WithDefaults(), WithoutAlgorithm(AlgoX), WithThreshold(0.80))` hides the baseline behind a token, making the composition less inspectable.
2. **Threshold override semantics.** With the current pattern, a later `WithThreshold(0.80)` in the appended slice correctly overrides the baseline `WithThreshold(0.85)` via last-write-wins. With `WithDefaults()`, the consumer would need separate reasoning to know that the threshold from `WithDefaults()` is overrideable.
3. **Slice mutability is the Go-functional-options idiom for prebuilt configurations.** `grpc.DefaultDialOptions()` would be the equivalent; Go's prevalent pattern is "return a slice you can append to or splice into".

A `WithDefaults() ScorerOption` would also be possible — it could re-apply the same set of options inside its closure. But this introduces ambiguity about ordering (is `WithDefaults` an option-level mutation or a config-level one?), conflicts with the duplicate-AlgoID last-write-wins semantic, and adds nothing the slice form doesn't already provide.

**Action:** none. Pattern approved. **Recommendation for `docs/scorer.md`:** add a one-line explanation of *why* `DefaultScorerOptions()` returns a slice (Principle 3 progressive disclosure: the consumer is taught the composition idiom).

---

### API-08 — Sentinel error names: `ErrEmptyScorer` reads better than `ErrNoAlgorithms`; `ErrInvalidWeight` reads better than `ErrWeightOutOfRange` [STRONG-PREFER, endorsement]

The user's concern: `ErrEmptyScorer` vs `ErrNoAlgorithms`? `ErrInvalidWeight` vs `ErrWeightOutOfRange`?

**Current names are correct.** Two principles in play:

1. **The error describes the consumer-visible state, not the implementation cause.** `ErrEmptyScorer` describes what the caller built ("an empty Scorer"); `ErrNoAlgorithms` describes the implementation's missing field. The caller's mental model is "I built a Scorer; it's empty"; the implementation's mental model is "the algorithms slice has zero entries". The former is the API contract.
2. **`Invalid` is the canonical Go prefix for parameter-validation failures.** `fs.ErrInvalid`, `net.InvalidAddrError`, `crypto/rsa.ErrVerification`, `errors.New("invalid utf-8")` (stdlib `unicode/utf8`). `ErrInvalidWeight` is consistent with this. `ErrWeightOutOfRange` is more specific (it tells you *why* the weight is invalid) but the message string already does that work (`"fuzzymatch: invalid algorithm weight (must be > 0)"` — the parenthetical carries the diagnostic).

**`ErrMissingThreshold` is correctly distinct from `ErrInvalidThreshold`.** The two errors describe different scenarios ("you forgot to pass it" vs "you passed something I can't use") and the consumer might want to programmatically discriminate via `errors.Is`. **Approved as-is.**

**Action:** none. Sentinel naming approved. **Recommendation:** the message strings are all good; the prefix discipline (`fuzzymatch: …`) is consistent and matches the documented convention in `errors.go`.

## API-09 — SPEC OVERRIDE: `ScoreAll` returns `map[AlgoID]float64` [API-ERGONOMICS-REVIEWER SIGN-OFF]

The user's concern: is the SPEC OVERRIDE (`map[AlgoID]float64` over `map[string]float64`) actually better than the spec text? What's the discoverability cost vs the type-safety win?

**Approved unanimously. Signed off by the api-ergonomics-reviewer.**

Three reasons:

1. **Type safety in the consumer call site.** `result[fuzzymatch.AlgoLevenshtein]` is a typed lookup that the compiler checks. `result["levenshtein"]` (the spec form) is a string lookup with zero compile-time validation — a typo (`"levenstein"`) silently returns zero. The Go-idiomatic move when you have a typed enum is to key by the enum, not by its string representation.
2. **Discoverability is unchanged.** Consumers can iterate `AlgoIDs()` (already exported) and call `id.String()` for the snake_case display form — IDE autocomplete reveals both the keys (typing `fuzzymatch.Algo` shows all 23 enum values) and the display strings (the `.String()` method). The spec form (`map[string]float64`) would require the consumer to discover the snake_case key namespace separately.
3. **Internal computation is already AlgoID-keyed.** The dispatch array, the `Algorithms()` accessor, the `WithAlgorithm` option, and the `ScorerAlgorithm.ID` field all use `AlgoID`. Adding a single `map[string]float64` return type would force a snake_case re-encoding pass on every call and surface the string namespace as the only place strings leak through the Scorer surface — inconsistent with the rest of the layer.

**Documentation already correctly records the override** at `scorer.go:468-472` and `docs/scorer.md:216-220`. The spec text at `docs/requirements.md` §8.3 also already carries the SPEC OVERRIDE note inline. **Nothing more required here.**

**Action:** none. Sign-off recorded.

---

## API-10 — `ScorerAlgorithm` struct shape and `Algorithms()` slice return [STRONG-PREFER, endorsement]

The user's concern: is the `ScorerAlgorithm{ID, Weight}` shape right? Does `Algorithms()` returning a slice of it match how consumers will use it?

**Yes, both right.** Three confirmations:

1. **Two-field struct is the minimum useful shape.** ID and post-normalisation Weight are the only things a consumer can meaningfully introspect — every other internal field of the Scorer is implementation detail. Adding any more (e.g. `RawWeight`, `ScoreFn`) would force the API to commit to internals that should remain free to evolve.
2. **Slice over map is the right collection type.** A slice preserves the documented AlgoID-ascending order (deterministic for logging and serialisation). A `map[AlgoID]ScorerAlgorithm` would lose that order. A `map[AlgoID]float64` (just the weights) would lose the discoverable structure (the consumer would need to call both `Algorithms()` and something else for the weights). Slice-of-struct gives both.
3. **Fresh allocation per call is the right concurrency contract.** Consumers may freely mutate, sort, or filter the returned slice without affecting subsequent `Algorithms()` calls or the Scorer's internal state. The cost (one slice allocation per call, typically 6-23 entries × 16 bytes) is negligible.

**Minor recommendation:** consider adding a `String()` method to `ScorerAlgorithm` that formats as `"<AlgoID.String()>: <Weight:.4f>"` — this enables one-line debug printing (`fmt.Println(scorer.Algorithms())` produces a readable summary). Low-impact; defer to v1.x if desired.

**Action:** none for the core shape. Optional v1.x: add `ScorerAlgorithm.String()` for debug ergonomics.

## NIT findings (polish)

### API-11 — `WithThreshold` mandatory is *not* visible in the godoc summary line [NIT]

The first-line godoc on `WithThreshold` reads: *"WithThreshold returns a ScorerOption that sets the Scorer's match threshold to t."* — there's no mention that it's mandatory until the second paragraph. A consumer scanning godoc summaries in their IDE sees only the first line.

**Fix:** Promote "MANDATORY" to the first line.

```go
// WithThreshold returns a ScorerOption that sets the Scorer's match
// threshold to t. WithThreshold is MANDATORY for NewScorer — see
// ErrMissingThreshold.
```

---

### API-12 — `DefaultScorerOptions()` godoc misses an explanation of WHY a slice [NIT]

The godoc explains the composition and the safe-to-mutate pattern, but not why the function returns a slice instead of a `Scorer`. A consumer seeing two functions (`DefaultScorer()` and `DefaultScorerOptions()`) needs a one-liner answer to "which should I use?"

**Fix:** Add a leading paragraph:

```go
// DefaultScorerOptions returns the option slice that DefaultScorer is
// built from, so consumers can derive a customised Scorer from the
// default composition. Use DefaultScorer if you want the default
// directly; use DefaultScorerOptions if you want to add, remove, or
// override options before constructing.
```

---

### API-13 — `WithoutNormalisation` godoc claim about `normOpts` reuse is fiction (WR-05 in code review) [NIT]

Already documented in `08-REVIEW.md` WR-05. The godoc claims that `normOpts` is left populated to support an "inspect-and-reuse" path that no callsite actually uses. Fix the godoc to say the real reason: clearing the field would require sentinel-value checks at the read sites for no behavioural gain.

---

### API-14 — `WithoutAlgorithm` godoc claims reverse iteration (WR-03 in code review) [NIT]

Already documented in `08-REVIEW.md` WR-03. Code iterates forward; comment says reverse. Fix the comment.

---

### API-15 — `docs/scorer.md` method-count says "four" but lists five (IN-01 in code review) [NIT]

Already documented in `08-REVIEW.md` IN-01. Two places, line 105 and 259. Fix to "All five methods" or rephrase to "All methods".

---

### API-16 — `WithMongeElkanAlgorithm` defers the 18-entry allow-list check to Score-time panic (IN-03 in code review) [NIT]

Already documented in `08-REVIEW.md` IN-03. The option validates `numAlgorithms` bounds + the trivial-recursion case, but does NOT validate against `MongeElkanScoreSymmetric`'s 18-entry allow-list. An inner AlgoID outside the allow-list passes construction and panics at `Score` time.

**My view (the user explicitly asked):** the BLOCKING bar for API-02 (`WithTverskyAlgorithm` panic) applies here too in principle — every `With*Algorithm` option should validate fully at construction time. **But** the ME allow-list is enforced inside `MongeElkanScoreSymmetric` and is owned by the ME implementation; duplicating the allow-list in the option layer creates a maintenance burden (two places to update when the ME allow-list changes).

**Recommended resolution:** Export an `IsValidMongeElkanInner(AlgoID) bool` helper from the ME implementation file. The option calls the helper; ME's `Score` function calls the helper. One source of truth, validation at construction time, no panic at Score time.

This is **NIT not BLOCKING** because the ME documentation explicitly flags the deferral and the runtime panic is at least typed-and-grep-able. Tracking as a v1.0 polish item.

---

### API-17 — `examples/scorer-composition/main.go` could showcase ScoreAll [NIT]

The example demonstrates `Score` and `Match` side-by-side for default vs default-minus-DM. It does not demonstrate `ScoreAll`, which is the only way for a consumer to see the per-algorithm breakdown that explains *why* the composite differs between the two Scorers. Adding a third column or a follow-on table would teach the calibration workflow that `docs/tuning.md` describes.

**Action:** optional. The example is already 138 lines; bloat resistance is fair. Defer to v1.x or to a separate `examples/scorer-breakdown/` if desired.

---

### API-18 — `Scorer` zero-value usability is undocumented [NIT]

`scorer.go:82-84` says "The zero-value Scorer is NOT usable. Always obtain a *Scorer from NewScorer or DefaultScorer". This is correct but the failure mode of using the zero value isn't pinned: a zero-value Scorer has `algorithmsAlgoIDSorted = nil`, `threshold = 0`, `applyNormalisation = false` — `Score(a, b)` returns 0 silently, `Match(a, b)` returns true for everything (because `0 >= 0`). Silent malfunction on the worst input (forgot to construct).

**Recommendation:** there's no clean way to make `(*Scorer)(nil).Score(...)` panic informatively without adding a nil-check to every method (which costs a branch on every call). The zero-value case is harder — Go doesn't have value-level nil sentinels. Best mitigation: a `*Scorer` returned by `NewScorer` cannot be the zero value, and the godoc directs the user to the constructor. **Accept as-is, but consider:**

- Add a docs note: "Constructing a Scorer via struct literal `&Scorer{}` is unsupported and will silently produce zero scores and false-positive matches."
- Add a test that asserts the zero-value misbehaviour (so it's pinned: future refactor can't accidentally make it look-usable).

---

## Summary table — by severity

| ID | Title | Severity | Action |
|----|-------|----------|--------|
| API-01 | `WithThreshold` does not reject NaN | BLOCKING | Apply CR-01 fix |
| API-02 | `WithTverskyAlgorithm` defers α+β check to Score-time panic | BLOCKING | Apply CR-02 fix |
| API-03 | Mandatory `WithThreshold` for NewScorer | STRONG-PREFER (endorsement) | None — approved |
| API-04 | Tversky parameter order inconsistent with score-fn shape | STRONG-PREFER | Move to `TverskyParams` struct before v1.0 |
| API-05 | `WithoutAlgorithm` / `WithoutNormalisation` naming | STRONG-PREFER (endorsement) | None — approved |
| API-06 | `WithParameterised(...interface{})` rejected | STRONG-PREFER (decision) | None — current family approved |
| API-07 | `DefaultScorerOptions()` returning slice | STRONG-PREFER (endorsement) | Optional docs polish |
| API-08 | Sentinel error names | STRONG-PREFER (endorsement) | None — approved |
| API-09 | SPEC OVERRIDE on ScoreAll return type | API-ERGONOMICS SIGN-OFF | Recorded here |
| API-10 | `ScorerAlgorithm` shape + `Algorithms()` slice | STRONG-PREFER (endorsement) | Optional `String()` method for v1.x |
| API-11 | `WithThreshold` MANDATORY not in summary line | NIT | Promote to first line |
| API-12 | `DefaultScorerOptions()` godoc rationale | NIT | Add leading paragraph |
| API-13 | `WithoutNormalisation` godoc fiction (WR-05) | NIT | Fix per code review |
| API-14 | `WithoutAlgorithm` godoc reverse-iteration claim (WR-03) | NIT | Fix per code review |
| API-15 | `docs/scorer.md` "four methods" / lists five (IN-01) | NIT | Fix per code review |
| API-16 | `WithMongeElkanAlgorithm` allow-list deferred | NIT | Optional `IsValidMongeElkanInner` helper |
| API-17 | Example could showcase `ScoreAll` | NIT | Optional |
| API-18 | Zero-value Scorer misbehaviour | NIT | Optional docs + test pinning |

## Sign-off

**Decisions reconciled with 08-CONTEXT.md:**

- §1 SPEC OVERRIDE (ScoreAll → `map[AlgoID]float64`): **approved (API-09).**
- §2 Mandatory threshold for NewScorer: **approved (API-03).**
- §3 Normalisation flow (pre-normalise once at Scorer boundary): **approved** — no findings here. Consistent with the spec, the implementation, and the documented contract.
- §4 Plan decomposition: outside ergonomics scope.
- §5 Float-determinism reduction loop: outside ergonomics scope (correctness-reviewer / determinism-reviewer territory).
- §8 ME vestigial opts parameter: **noted as deferred** (will revisit at v1.0 freeze per CONTEXT deferred-ideas list).

**Pre-v1.0 blocking work:**
1. API-01 — `WithThreshold` NaN rejection (CR-01 in code review)
2. API-02 — `WithTverskyAlgorithm` α+β==0 rejection (CR-02 in code review)

**Pre-v1.0 strong-prefer work:**
3. API-04 — `TverskyParams` struct migration (signature-breaking; must land before v1.0 cut)

**v1.x polish (non-blocking):**
- All NIT items above.

**This review constitutes the api-ergonomics-reviewer's final sign-off on the Phase 8 Scorer surface, subject to the two BLOCKING items being resolved and API-04 being landed before v1.0.**

---

*Reviewed: 2026-05-17*
*Reviewer: api-ergonomics-reviewer*
*Authority: final (CLAUDE.md Design Principle 13)*
