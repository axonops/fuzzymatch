# Scan sub-package

The `scan` sub-package at `github.com/axonops/fuzzymatch/scan` is the
turnkey collection-scan layer over the `Scorer`. It answers the
question "which pairs in this collection are similar?" with a
deterministic, suppression-aware, group-aware API. Available since
v1.0.

A `scan.Config` is a plain data struct passed by value; `scan.Check`
is a pure function with no goroutines, channels, or mutexes — safe
for concurrent invocation on disjoint inputs given a concurrent-safe
`*fuzzymatch.Scorer` (Phase 8 guarantee). There are no global flags,
no mutable internal state, and no init-time work.

The authoritative formal specification lives in
[`docs/requirements.md`](requirements.md) §12. This document is the
consumer-facing guide.

## Quickstart

The one-line happy path is:

```go
package main

import (
    "github.com/axonops/fuzzymatch"
    "github.com/axonops/fuzzymatch/scan"
)

func main() {
    s := fuzzymatch.DefaultScorer()
    items := []scan.Item{
        {Name: "user_id", Group: "login"},
        {Name: "userId", Group: "login"},
    }
    warnings, err := scan.Check(items, scan.DefaultConfig(s))
    if err != nil {
        // handle ErrInvalidItem / ErrInvalidConfig / ErrNilScorer
    }
    _ = warnings // process the deterministically-sorted slice
}
```

`scan.DefaultConfig(s)` cannot fail. It bakes the opinionated
`CrossGroupThresholdBoost = 0.05` and `CompareIdenticalAcrossGroups
= false` defaults — see [Threshold boost](#threshold-boost) and
[Suppression composition](#suppression-composition) below.

For tighter precision the recommended composition is **Validate-then-Check**:

```go
for i, it := range items {
    if ws := fuzzymatch.Validate(it.Name, ""); len(ws) > 0 {
        // log + decide whether to continue (Phase 8.5 Q4 — Validate
        // is the input-quality diagnostic surface; the scan layer
        // accepts every non-empty Name and trusts the upstream caller
        // to gate on Validate output).
        _ = i
    }
}
warnings, err := scan.Check(items, scan.DefaultConfig(s))
```

See [`docs/best-practices.md`](best-practices.md) for the broader
Validate-then-Score idiom that applies to every layer.

## Public API

| Function / Type / Constant         | Signature                                                                | Description                                                                                                                                  |
| ---------------------------------- | ------------------------------------------------------------------------ | -------------------------------------------------------------------------------------------------------------------------------------------- |
| `Check`                            | `func Check(items []Item, cfg Config) ([]Warning, error)`                | Cross-comparison entry-point with within-group + opt-in cross-group passes, suppression composition, deterministic output.                   |
| `DefaultConfig`                    | `func DefaultConfig(s *fuzzymatch.Scorer) Config`                        | Opinionated `Config`: `CrossGroupThresholdBoost = 0.05`, `CompareIdenticalAcrossGroups = false`.                                             |
| `Item`                             | `struct { Name, Group string; SilenceLint bool; Tag any }`               | One named thing the scanner compares. `Name` is required (non-empty); `Group` scopes comparisons.                                            |
| `Kind`                             | `type Kind int`                                                          | Pair-scope discriminator (within-group vs cross-group). SPEC OVERRIDE (Phase 9): renamed from `WarningKind` per [09-CONTEXT D-02](#decision-references). |
| `KindWithinGroup`, `KindAcrossGroups` | `const Kind = iota+1`                                                 | The two Kind values. Iota+1 so the zero value is "unset".                                                                                    |
| `Kind.String`                      | `func (k Kind) String() string`                                          | CamelCase form: `"WithinGroup"` / `"AcrossGroups"` — matching `AlgoID.String()` and `WarnKind.String()` (Phase 8.5 Q6b).                     |
| `Warning`                          | `struct { Kind; NameA, NameB; GroupA, GroupB; TagA, TagB any; Score; Scores map[fuzzymatch.AlgoID]float64 }` | One detected similar-name pair. `Scores` is typed-key per [09-CONTEXT D-01](#decision-references) (SPEC OVERRIDE).                          |
| `Config`                           | `struct { Scorer *fuzzymatch.Scorer; CompareAcrossGroups bool; CrossGroupThresholdBoost float64; CompareIdenticalAcrossGroups bool; SuppressedPairs [][2]string }` | Controls Check behaviour. Zero value: nil Scorer → `ErrNilScorer`. Zero `CrossGroupThresholdBoost` = 0.0 (no boost). |
| `ErrNilScorer`                     | `var error`                                                              | `Config.Scorer` was nil when `Check` was invoked.                                                                                            |
| `ErrInvalidItem`                   | `var error`                                                              | Empty `Item.Name` (D-03) or duplicate `(Name, Group)` (D-06); offending indices collected via `errors.Join`.                                |
| `ErrInvalidConfig`                 | `var error`                                                              | NaN/±Inf/out-of-range `CrossGroupThresholdBoost` (D-04) or empty entry in `SuppressedPairs` (D-05).                                          |
| `(*fuzzymatch.Scorer).NormalisationOptions` | `func() (NormalisationOptions, bool)`                            | New Phase 9 accessor on `*Scorer` returning the Scorer's normalisation options. Used internally by `scan` to canonicalise `SuppressedPairs`. |

## Within-group vs cross-group passes

`scan.Check` runs two distinct passes:

- **Within-group pass.** For every group (in sorted-key order), every
  `i<j` pair is evaluated via `cfg.Scorer.Score` and compared against
  the Scorer's threshold. Always on (default). Items with no Group
  share the empty-string group; the same iteration discipline applies.

- **Cross-group pass.** Opt-in via `cfg.CompareAcrossGroups = true`.
  For every pair of distinct groups `(gi < gj)`, every pair `(i ∈ gi,
  j ∈ gj)` is evaluated against the **effective cross-group
  threshold**:

  ```text
  effectiveThreshold = min(1.0, scorer.Threshold() + cfg.CrossGroupThresholdBoost)
  ```

  The `min(1.0, ...)` clamp pins the effective threshold at 1.0 even
  when the arithmetic sum would exceed it (e.g. `Threshold(0.85) +
  Boost(0.20)` arithmetically yields 1.05; the clamp makes that 1.0,
  meaning only byte-identical-post-normalise pairs reach the
  threshold).

The cross-group pass is inherently noisier than the within-group
pass — operators legitimately have schemas where similar names appear
across distinct groups. The boost + identical-Name suppression
(Rule 3) work together to keep the signal-to-noise ratio reasonable.

## Suppression composition

Three rules compose via short-circuit OR — any rule firing suppresses
the pair pre-emission. Order of evaluation is cheapest-first; the
final emission set is identical regardless of order because OR
composition is order-independent on its output.

### Rule 1 — per-item `SilenceLint`

`Item.SilenceLint = true` suppresses every pair involving that
`Item`. One-sided semantics: setting the flag on either side of a
pair silences the pair.

```go
items := []scan.Item{
    {Name: "user_id", Group: "login", SilenceLint: true},
    {Name: "userId",  Group: "login"},
}
warnings, _ := scan.Check(items, scan.DefaultConfig(s))
// len(warnings) == 0 — items[0]'s SilenceLint suppresses the pair.
```

### Rule 2 — `Config.SuppressedPairs` canonical-pair lookup

A `[][2]string` of name pairs that should never produce a warning.
Pairs are canonicalised **once at Check entry** via the Scorer's
`NormalisationOptions()` (D-05) — consumers may pass raw forms
regardless of case or separators when the Scorer's normalisation
absorbs those.

```go
cfg := scan.DefaultConfig(s)
cfg.SuppressedPairs = [][2]string{
    {"user_id", "userId"}, // canonicalised — case + separator drift absorbed
}
warnings, _ := scan.Check(items, cfg)
```

Self-pairs (`a == b` after normalisation) are silently retained — they
are harmless because `Check` never emits a self-warning under D-06
(duplicate `(Name, Group)` is rejected at validation) and the `i<j`
pair-iteration discipline. **Caveat:** a self-pair entry whose
normalised form coincides with the canonical key of a DISTINCT-name
candidate pair will also suppress that distinct pair. This is the
inevitable consequence of canonical-pair semantics — flagged here so
consumers building suppression lists programmatically know to expect
it.

### Rule 3 — cross-group identical-Name default

When `cfg.CompareAcrossGroups == true` and
`cfg.CompareIdenticalAcrossGroups == false` (the `DefaultConfig`
default), identical Names across different Groups are silently
suppressed. Operators legitimately reuse the same name (e.g.
`user_id`) across groups; surfacing every such pair would drown real
similar-but-not-equal signals. Opt out by setting
`cfg.CompareIdenticalAcrossGroups = true`.

The check uses normalised-name equality, so case/separator-only
differences across groups are also silenced under the default
normalisation pipeline.

### Composition

The three rules compose via OR. The `isSuppressed` predicate is
applied **pre-emission** on both the naive and token-bucket emission
paths (so SCAN-02 bucket-vs-naive equivalence holds under
suppression). Suppression is never applied post-emission; the
emitted slice is precisely the unsuppressed set.

## Threshold boost

The cross-group effective threshold is:

```text
effectiveThreshold = min(1.0, scorer.Threshold() + cfg.CrossGroupThresholdBoost)
```

`CrossGroupThresholdBoost` is validated strictly at Check entry:
NaN, ±Inf, `< 0`, or `> 1` → `ErrInvalidConfig` (D-04). The
zero-value is `0.0` (no boost); `scan.DefaultConfig` bakes the
opinionated `0.05` default. The `0.05` value reflects production-
calibration evidence that cross-group matches tend to be ~5
percentage points noisier than within-group matches.

## Validation surface

Three subsections per locked decisions D-03 / D-04 / D-05 / D-06.
All validation collects **every** offending index in a single pass
and joins them via `errors.Join`, so callers can fix the whole batch
in one round-trip rather than playing whack-a-mole.

### D-03 — Empty `Item.Name`

`ErrInvalidItem` is returned with `errors.Join` across all offending
indices. Discriminate via `errors.Is`:

```go
warnings, err := scan.Check(items, cfg)
if errors.Is(err, scan.ErrInvalidItem) {
    // one or more items had empty Name or were duplicates;
    // walk the joined error to enumerate every offending index.
}
```

### D-04 — Invalid `Config.CrossGroupThresholdBoost`

`ErrInvalidConfig` is returned when the boost is NaN, ±Inf, `< 0`,
or `> 1`. Strict-range validation per the Phase 8.5 Q2 parameter
contract (parameters vs comparison-data are validated differently —
parameters fail fast, comparison data passes silently).

```go
cfg := scan.DefaultConfig(s)
cfg.CrossGroupThresholdBoost = math.NaN()
_, err := scan.Check(items, cfg)
// errors.Is(err, scan.ErrInvalidConfig) == true
```

### D-05 — Empty `SuppressedPairs` entry

`ErrInvalidConfig` is returned with `errors.Join` across all
offending entry indices when one or both strings in a
`SuppressedPairs` entry are empty. Self-pairs (`a == b`) are
permitted (see Rule 2 caveat).

### D-06 — Duplicate `(Name, Group)`

`ErrInvalidItem` is returned with `errors.Join` across all duplicate
indices. The duplicate-detection invariant guarantees the
deterministic sort key `(Kind, NameA, NameB, GroupA, GroupB)` is a
strict total order on every valid input — eliminating the need for a
post-sort tiebreaker.

## Determinism

`scan.Check` guarantees byte-identical output across runs and
platforms given identical input:

- **Lex canonicalisation.** Before sorting, every Warning is
  canonicalised so `NameA <= NameB` under raw-byte lex compare;
  `GroupA/GroupB` and `TagA/TagB` swap in lockstep so the
  `(Name, Group, Tag)` triple still describes the same source Item.

- **Sort key.** `sort.SliceStable` on the 5-tuple
  `(Kind, NameA, NameB, GroupA, GroupB)`. Every field participates
  in the comparator, so the sort is a strict total order on valid
  input. `D-06`'s duplicate-detection invariant guarantees no two
  Warnings can share the full 5-tuple key.

- **No map iteration on output paths.** The `groupIndices` map is
  iterated only when building the `sortedGroups []string` slice
  (which is then sorted). Every downstream emission iterates the
  sorted slice. Per
  [`.claude/skills/determinism-standards`](../.claude/skills/determinism-standards/SKILL.md).

- **In-line completeness assertion.** After sorting, `Check` scans
  adjacent warnings linearly; any pair sharing the full 5-tuple sort
  key triggers a `panic(fuzzymatch.ErrInternalInvariantViolated)`.
  The assertion is unreachable on valid input (D-06 rejects
  duplicate `(Name, Group)` at validation) — it exists as the
  documented invariant gate per the Phase 8.5 Gap 5 typed-panic
  convention. Consumers discriminate via `errors.Is(recovered,
  fuzzymatch.ErrInternalInvariantViolated)`.

- **NaN/Inf/-0 never appear in output** (DET-04). Property tests
  `PropCheck_NoNaN` + `PropCheck_NoInf` verify this on randomised
  input.

The cross-platform CI matrix (`linux/amd64`, `linux/arm64`,
`darwin/amd64`, `darwin/arm64`, `windows/amd64`) verifies
byte-identical output via the golden file at
`testdata/golden/scan-default.json`. The `make verify-determinism`
target runs the golden-file test on every platform.

## Performance

`scan.Check` performance is dominated by the per-pair Scorer cost.
The library applies two optimisations to amortise the dispatch
overhead:

- **Tokenise-once.** Item Names are normalised + tokenised once at
  Check entry (per Item, not per pair). The resulting token slice
  is reused by every per-pair Scorer call within the same Check
  invocation. Avoids the `O(N²)` re-tokenisation pattern that
  inflated v0.x performance roughly 3×.

- **Token-bucket pre-filter.** For groups (within-group) or
  group-pair unions (cross-group) exceeding the internal
  `bucketThreshold` (50 items, empirically validated on
  `darwin/arm64`), candidate pairs are pre-filtered by shared
  tokens before the Scorer is consulted. Pairs sharing no token
  are eliminated without paying `Scorer.Score`. The
  `PropCheck_BucketEquivalentToNaive` property test
  (`scan/props_test.go`) proves the bucket path produces the
  identical warning set as the naive O(N²) reference on randomised
  input — SCAN-02 load-bearing gate.

**PERF-05 budget:** `scan.Check` completes in **< 2 seconds for
10,000 items / 500 groups** within-group on the reference hardware,
verified by `BenchmarkScanCheck_DefaultScorer_10k` (Plan 09-04). The
< 10% benchstat regression gate runs in CI on every PR; the
committed `bench.txt` is the regression baseline.

**Cross-group performance note (v0.x baseline):** The cross-group
pass at 10,000 items / 500 groups runs substantially slower than the
within-group pass on the same input shape — the v0.x baseline is
approximately 189 seconds on `darwin/arm64` for the
cross-group-enabled run, vs ~361 milliseconds within-only. Spec
§12.6's "at most 2× within-only" claim is documented as a v0.x
shortfall; the optimisation candidate (build a single global token
bucket once at Check entry and apply per-pair group filters) is
tracked in `09-CONTEXT.md` "Deferred Ideas" for v1.x consideration.
Consumers with large cross-group workloads at v0.x should expect
multi-second wall-clock times and may need to chunk their input or
restrict `CompareAcrossGroups` to selected group subsets.

## Concurrency

`scan.Config` is a plain data struct passed by value to `Check`.
`*fuzzymatch.Scorer` is concurrent-safe by construction (Phase 8
locked guarantee — immutable after `NewScorer`, no internal mutex
or atomic state). `scan.Check` is a pure function with no
goroutines, channels, or mutexes — safe for concurrent invocation
on disjoint `[]Item` slices given a shared `*Scorer`.

Consumers passing the same `Config` to multiple concurrent `Check`
invocations must not mutate the `SuppressedPairs` slice between
calls. The slice is read once at Check entry to build the
canonical-pair map; subsequent mutations would race with that read.

The `Warning.Scores` map in each emitted Warning is freshly
allocated by `Scorer.ScoreAll` — consumers may mutate it freely
without affecting other Warning values or other Check invocations.

## Errors

Three sentinels. All are typed `error` values exported from the
package; discriminate via `errors.Is`, never by matching the error
message string.

| Sentinel             | When it fires                                                                                                                                     |
| -------------------- | ------------------------------------------------------------------------------------------------------------------------------------------------- |
| `ErrNilScorer`       | `Config.Scorer` was nil when `Check` was invoked. A nil Scorer has no algorithms, no threshold, and no normalisation options.                     |
| `ErrInvalidItem`     | Empty `Item.Name` (D-03) or duplicate `(Name, Group)` (D-06). `errors.Join` across all offending indices.                                         |
| `ErrInvalidConfig`   | NaN/±Inf/out-of-range `CrossGroupThresholdBoost` (D-04) or empty entry in `SuppressedPairs` (D-05). `errors.Join` across all offending indices.   |

`errors.Is` walks `Unwrap() []error` (Go 1.20+), so the joined
error still discriminates correctly:

```go
warnings, err := scan.Check(items, cfg)
switch {
case errors.Is(err, scan.ErrNilScorer):
    // diagnostic
case errors.Is(err, scan.ErrInvalidItem):
    // walk the joined error to enumerate every offending index
case errors.Is(err, scan.ErrInvalidConfig):
    // either CrossGroupThresholdBoost or SuppressedPairs is invalid
case err != nil:
    // unexpected — file an issue
}
```

## Decision references

The implementation honours a set of locked decisions recorded in
[`09-CONTEXT.md`](../.planning/phases/09-collection-scan-sub-package/09-CONTEXT.md)
that materially shape the public surface. The most consumer-visible
ones (also surfaced as SPEC OVERRIDE notes inline above):

- **D-01.** `Warning.Scores` is `map[fuzzymatch.AlgoID]float64`
  (typed enum keys), NOT `map[string]float64` as
  [`docs/requirements.md`](requirements.md) §12.1 originally
  specified. Extends Phase 8's `ScoreAll` override for the same
  compile-time-safety rationale.

- **D-02.** The within/cross discriminator is `Kind`, NOT
  `WarningKind` as the spec originally named it. The package-scoped
  form `scan.KindWithinGroup` is unambiguous at the call site and
  avoids accidental symmetry with the root package's `WarnKind`
  (Phase 8.5 Q4 — a different domain).

- **D-04.** The opinionated `CrossGroupThresholdBoost = 0.05`
  default lives in `scan.DefaultConfig`, NOT as the zero-value of
  `Config.CrossGroupThresholdBoost`. Mirrors the Phase 8
  `DefaultScorer` / `DefaultScorerOptions` pattern.

- **D-06.** `Check` rejects duplicate `(Name, Group)` at validation
  time. Combined with the raw-byte lex canonicalisation of NameA/
  NameB, this guarantees the 5-tuple sort key is a strict total
  order without a tiebreaker — eliminating the spec's earlier
  "lexicographically smaller (after normalisation)" tiebreaker
  wording, which is no longer needed.

- **D-07.** The recommended composition is **Validate-then-Check** —
  see [Quickstart](#quickstart) and
  [`docs/best-practices.md`](best-practices.md).

## See also

- godoc: <https://pkg.go.dev/github.com/axonops/fuzzymatch/scan>
- Runnable example: [`examples/scan-demo/`](../examples/scan-demo/)
- Authoritative spec: [`docs/requirements.md`](requirements.md) §12
- Validate-then-Score idiom: [`docs/best-practices.md`](best-practices.md)
- Scorer composition: [`docs/scorer.md`](scorer.md)
- Threshold tuning: [`docs/tuning.md`](tuning.md)
