# Phase 9: Collection Scan Sub-package - Pattern Map

**Mapped:** 2026-05-19
**Files analysed:** 22 new (scan sub-package), 2 modified (root: `scorer.go`, `docs/requirements.md`), plus 4 shared touch-ups (`Makefile`, `CHANGELOG.md`, `llms.txt`, `llms-full.txt`)
**Analogs found:** 22 / 22 (every new file has at least one closest analog inside the repo)

---

## File Classification

| New/Modified File | Role | Data Flow | Closest Analog | Match Quality |
|---|---|---|---|---|
| `scan/scan.go` | public types + Check entry-point | request-response (slice in, slice + err out) | `scorer.go` (NewScorer/Score/Match/DefaultScorer surface) | exact (Layer-2 → Layer-3 transposition) |
| `scan/kind.go` | typed enum + String() | n/a (pure declaration) | `warn_kind.go` (`WarnKind`) | exact (verbatim pattern) |
| `scan/errors.go` | sentinel error catalogue | n/a | `errors.go` (root sentinel set) | exact (verbatim 4-section godoc) |
| `scan/validate.go` | internal pre-flight validation (P1..P4) | transform (in: items+cfg → out: error or nil) | `validate.go` (`Validate` public, but pattern of "accumulate then sort then emit") + Phase 8 `NewScorer` validation pipeline | role-match (validate.go is public surface; new file is internal pipeline) |
| `scan/bucket.go` | token-bucket optimisation + private const | transform (in: items+tokens → out: candidate sets) | `tokenise.go` consumer pattern + `scorer.go` `algorithmsAlgoIDSorted` slice-as-canonical-order pattern | role-match (no exact precedent — first internal-only optimisation file) |
| `scan/suppress.go` | suppression composition predicate | predicate (in: pair+kind+ctx → out: bool) | `validate.go` predicate helpers (`hasOnlyNonASCII`) | partial (predicate shape) |
| `scan/doc.go` | package-level godoc | n/a | `tokenise.go` top-of-file godoc; `scorer.go` design-notes block | role-match |
| `scan/scan_test.go` | black-box unit tests | request-response (table-driven) | `scorer_test.go` (external `_test` package + table-driven Test* functions) | exact |
| `scan/scan_internal_test.go` | internal/unexported-helper tests | n/a | `scorer_internal_test.go` (same `package fuzzymatch` test for unexported helpers) | exact |
| `scan/scan_bench_test.go` | benchmarks + D-08 sweep | n/a | `scorer_bench_test.go` (`b.ReportAllocs`, sized inputs, sink-gate) | exact |
| `scan/example_test.go` | godoc runnable examples | n/a | `examples` blocks elsewhere; closest in-tree analog is patterns inside `scorer_test.go` and the standalone `examples/scorer-composition/main.go` | role-match (one-per-public-symbol Example fn pattern) |
| `scan/props_test.go` | property tests via `testing/quick` | n/a | `scorer_test.go` lines 832–982 (PropScorer_DeterministicAcrossRuns, PropScorer_WeightSumOne) | exact |
| `scan/fuzz_test.go` | native Go fuzz (`go test -fuzz`) | n/a | `damerau_full_fuzz_test.go`, `cosine_fuzz_test.go` (typical FuzzX with seed corpus) | role-match |
| `scan/scan_golden_test.go` | golden file diff test | n/a | `scorer_golden_test.go` (verbatim envelope + entry struct + `writeGoldenFile` pattern) | exact |
| `testdata/golden/scan-default.json` | golden corpus | n/a | `testdata/golden/scorer-default.json` (`_metadata` envelope + sorted entries) | exact |
| `testdata/golden/_staging/scan.json` | staging file (pre-merge) | n/a | `testdata/golden/_staging/*.json` (existing convention) | exact |
| `tests/bdd/features/scan.feature` | Gherkin scenarios | n/a | `tests/bdd/features/scorer.feature` (Phase 8 BDD) | exact |
| `tests/bdd/features/suppression.feature` | Gherkin scenarios | n/a | `tests/bdd/features/scorer.feature` + `validate.feature` (per-Kind scenario discipline) | exact |
| `tests/bdd/steps/scan_steps.go` | godog step definitions | event-driven (per-step closure on ScanContext) | `tests/bdd/steps/scorer_steps.go` (ScorerContext + InitScorerSteps + ctx.Step regex pattern) | exact |
| `examples/scan-demo/main.go` | runnable demo | request-response | `examples/scorer-composition/main.go` (var-init helpers + pairs + tabular print) | exact |
| `examples/scan-demo/main_test.go` | stdout meta-test | n/a | `examples/scorer-composition/main_test.go` (os.Pipe stdout redirect + byte-for-byte `want`) | exact |
| `examples/scan-demo/go.mod` | per-example module | n/a | `examples/scorer-composition/go.mod` (4-line module with `replace ../..` directive) | exact |
| `docs/scan.md` | user-facing docs | n/a | `docs/scorer.md` (Quickstart → Custom → Concepts → Thread Safety) | exact |
| `scorer.go` (MODIFIED) | add `NormalisationOptions()` accessor (Open Question 1) | n/a | existing `Threshold()` accessor (scorer.go:447–449) and `Algorithms()` (scorer.go:466–472) | exact |
| `docs/requirements.md` (MODIFIED) | spec amendments (§12.1 + §12.4 + §15 + line 2010 + line 2061; SPEC OVERRIDE notes for D-01 / D-02 / D-03 / D-04) | n/a | The Phase 8 amendment trail (08-CONTEXT.md §1) + the in-source `SPEC OVERRIDE` note already at scorer.go:476 | role-match |

---

## Pattern Assignments

### `scan/scan.go` — public types + Check entry-point

**Closest analog:** `scorer.go` (Phase 8 Layer-2 surface; same shape: opinionated `Default*` helper + explicit constructor/operator pair + immutable post-construction value + four-section godoc per method).

**Pattern 1 — File-header design-notes block** (analog: `scorer.go:15-74`):

```go
// Copyright 2026 AxonOps Limited
// [Apache 2.0 boilerplate — copy verbatim]

// scan.go declares the Phase 9 Layer-3 collection-scan surface of the
// three-layer fuzzymatch architecture: Item, Config, Warning, Kind,
// Check, DefaultConfig. It consumes the Phase 8 *fuzzymatch.Scorer
// for per-pair similarity and orchestrates within/cross-group passes
// with token-bucket optimisation, deterministic sort, and suppression
// composition.
//
// SPEC OVERRIDE (Phase 9):
//   - Warning.Scores is map[fuzzymatch.AlgoID]float64 (NOT
//     map[string]float64 as in docs/requirements.md §12.1 line 1337).
//     Extends the Phase 8 ScoreAll override at §8.3 + §8.6 (08-CONTEXT.md
//     §1) for the same typed-enum-keys rationale. See 09-CONTEXT.md
//     §1 D-01.
//   - Kind type renamed from WarningKind (NOT scan.WarnKind to avoid
//     accidental symmetry with root's WarnKind). See 09-CONTEXT.md §1
//     D-02. Spec §12.1 renamed in lockstep.
//
// Design notes (per 09-CONTEXT.md + 09-RESEARCH.md):
//   - Validation pipeline order (P1..P4) is LOCKED: nil-Scorer fail-fast
//     → Config field validation → Items validation (D-03+D-06 collect-all
//     via errors.Join) → SuppressedPairs validation (D-05 collect-all).
//   - Naive ↔ bucket dispatch by group size (bucketThreshold=50, private
//     const in bucket.go — D-08).
//   - Cross-group threshold = min(1.0, scorer.Threshold() +
//     cfg.CrossGroupThresholdBoost); cross-group pass uses
//     scorer.Score (NOT scorer.Match — Match uses the within-group
//     threshold).
//   - In-line completeness assertion panics with
//     fuzzymatch.ErrInternalInvariantViolated (Phase 8.5 Gap 5).
//   - No goroutines, no channels, no mutexes. Pure-function.

package scan

import (
    "errors"
    "fmt"
    "math"
    "sort"

    "github.com/axonops/fuzzymatch"
)
```

**Pattern 2 — SPEC OVERRIDE inline godoc on the Warning struct** (analog: `scorer.go:474-502` ScoreAll godoc — the canonical Phase 8 SPEC OVERRIDE template):

Reference excerpt (scorer.go:475-478):

```go
// ScoreAll returns per-algorithm raw scores for the configured algorithm set as a map[AlgoID]float64.
//
// SPEC OVERRIDE: docs/requirements.md §8.3 specifies map[string]float64; this implementation returns map[AlgoID]float64 because AlgoID is a typed enum that the rest of the library exposes, giving consumers compile-time key safety. Use AlgoID.String() for snake_case display. The spec deviation is documented in CONTEXT.md §1 (Phase 8) and api-ergonomics-reviewer signed off on this override in plan 08-03's PR.
```

Apply verbatim to `scan.Warning.Scores` and `scan.Kind` godoc — cite 09-CONTEXT.md §1 D-01 / D-02 and the lockstep spec amendment.

**Pattern 3 — DefaultConfig opinionated-helper** (analog: `scorer.go:549-559` `DefaultScorerOptions` + `scorer.go:592-594` `DefaultScorer`):

```go
// scorer.go:549-559 — pattern to mirror for scan.DefaultConfig.
func DefaultScorerOptions() []ScorerOption {
    return []ScorerOption{
        WithAlgorithm(AlgoDamerauLevenshteinOSA, 1.0),
        WithAlgorithm(AlgoJaroWinkler, 1.0),
        WithAlgorithm(AlgoTokenJaccard, 1.0),
        WithAlgorithm(AlgoQGramJaccard, 1.0),
        WithAlgorithm(AlgoSorensenDice, 1.0),
        WithAlgorithm(AlgoDoubleMetaphone, 1.0),
        WithThreshold(0.85),
    }
}
```

The shape for `scan.DefaultConfig` is given verbatim in 09-RESEARCH.md lines 731–739; mirror the godoc style (one line per default, "Typical usage" code block, SPEC OVERRIDE note for the 0.05 boost).

**Pattern 4 — Two-step normalise + parallel arrays** (analog: `scorer.go:352-360` Score pre-normalisation boundary):

```go
// scorer.go:353-360 — boundary pattern. scan.Check builds parallel
// arrays of raw and normalised names so the Scorer boundary stays clean
// (Pitfall 5 in 09-RESEARCH.md).
na, nb := a, b
if s.applyNormalisation {
    na = Normalise(a, s.normaliseOpts)
    nb = Normalise(b, s.normaliseOpts)
}
```

In `scan.Check` the equivalent is:

```go
rawNames        := make([]string, len(items))   // for scorer.Score (re-normalises)
normalisedNames := make([]string, len(items))   // for bucket keys + suppression lookup
for i, item := range items {
    rawNames[i] = item.Name
    normalisedNames[i] = fuzzymatch.Normalise(item.Name, normOpts)
}
```

**Pattern 5 — In-line completeness assertion panic** (analog: `scorer.go:592-594` `DefaultScorer` → `mustDefaultScorer` typed-panic pattern; see also `errors.go:213-240` `ErrInternalInvariantViolated`):

```go
// errors.go:213-240 — ErrInternalInvariantViolated is the typed-panic
// sentinel for library bugs (Phase 8.5 Gap 5). scan.Check post-sort
// completeness assertion panics with a wrapped error.
```

The exact pattern is given in 09-RESEARCH.md lines 428–438; copy verbatim and update the message to cite the duplicate index + 5-tuple key.

---

### `scan/kind.go` — typed enum + CamelCase String

**Closest analog:** `warn_kind.go` (root pkg) — 165 lines, verbatim shape.

**Imports + type + constants** (warn_kind.go:38-97):

```go
package fuzzymatch

import "fmt"

// WarnKind classifies the kind of input-quality concern Validate
// reports. WarnKind is a plain int (not int32 / int64 / struct) so it
// is trivially comparable and serialisable. The zero value is reserved
// as "unspecified" — every documented WarnKind starts at 1 (iota + 1).
type WarnKind int

const (
    WarnEmptyInput WarnKind = iota + 1
    WarnUnequalLength
    WarnNoTokensAfterNormalise
    WarnAllNonASCIIDropped
    WarnPathologicallyLargeInput
)
```

**String() switch with allocating fallback** (warn_kind.go:129-144):

```go
func (k WarnKind) String() string {
    switch k {
    case WarnEmptyInput:
        return "EmptyInput"
    case WarnUnequalLength:
        return "UnequalLength"
    case WarnNoTokensAfterNormalise:
        return "NoTokensAfterNormalise"
    case WarnAllNonASCIIDropped:
        return "AllNonASCIIDropped"
    case WarnPathologicallyLargeInput:
        return "PathologicallyLargeInput"
    default:
        return fmt.Sprintf("WarnKind(%d)", int(k))
    }
}
```

**Pattern notes — apply verbatim with rename Kind → scan.Kind, two cases, SPEC OVERRIDE note in the type-level godoc.** The exact 5-line String() body for `scan.Kind` is in 09-RESEARCH.md lines 693–703.

---

### `scan/errors.go` — sentinel error catalogue

**Closest analog:** `errors.go` (root pkg) — 277 lines, four-section godoc per sentinel.

**File-header godoc** (errors.go:15-35):

```go
// Package-level sentinel errors for fuzzymatch. All errors are wrappable
// via fmt.Errorf("...: %w", ErrX) and discoverable by errors.Is /
// errors.As — never via string matching. See docs/requirements.md §6
// (canonical sentinel set) and §6.A (data-vs-parameter validation
// framework + panic policy).
//
// Error message convention (per .claude/skills/go-coding-standards):
//
//   - Every message starts with the "fuzzymatch: " package prefix so
//     wrappers like fmt.Errorf("scorer: %w", err) produce readable
//     compositions ("scorer: fuzzymatch: invalid algorithm identifier").
//   - The text after the prefix is lowercase and carries no trailing
//     punctuation ('.', '!', or '?') so concatenation flows cleanly.
//   - Each sentinel is a flat errors.New value, not a typed struct;
//     richer per-item context can be added in a later phase if scan
//     or extract needs it.

package fuzzymatch

import "errors"
```

**Per-sentinel four-section godoc** (errors.go:157-170 — ErrEmptyScorer, the most compact):

```go
// ErrEmptyScorer indicates NewScorer was called without any algorithm
// option — the option slice contained zero WithAlgorithm /
// With*Algorithm entries by the time validation ran. A Scorer with no
// algorithms has no meaningful composite to compute.
//
// Pass at least one WithAlgorithm option (or use DefaultScorer() for
// the opinionated six-algorithm composition).
//
// Returned by NewScorer (Phase 8) after the missing-threshold check
// passes and the option-validation pipeline finds cfg.entries empty.
//
// Discriminate via errors.Is(err, fuzzymatch.ErrEmptyScorer); never
// match the error message string.
var ErrEmptyScorer = errors.New("fuzzymatch: scorer has no algorithms (pass at least one WithAlgorithm option or use DefaultScorer)")
```

**4-section template structure:**
1. **What** — one paragraph stating the condition the sentinel signals.
2. **Common causes** — bulleted typical mistakes that produce it.
3. **Resolution** — what the caller does to fix it; reference to the canonical helper.
4. **Example** — `errors.Is` snippet (NOT a runnable godoc Example — those live in `example_test.go`).

**Three sentinels for scan/errors.go** (per 09-CONTEXT.md and 09-RESEARCH.md line 669):

```go
// 09-RESEARCH.md lines 651-669 — ErrNilScorer template ready to drop in.
var ErrNilScorer = errors.New("scan: Config.Scorer is required")
var ErrInvalidItem = errors.New("scan: invalid Item")
var ErrInvalidConfig = errors.New("scan: invalid Config")
```

**Pattern notes:** The package prefix changes from `fuzzymatch:` to `scan:`. The four sections are non-negotiable per Phase 8.5 Q4 template. Include `errors.Is(joined, scan.ErrInvalidItem)` discrimination note in the ErrInvalidItem godoc explicitly (consumers calling `errors.Join`-aware code need the cross-reference).

---

### `scan/validate.go` — pre-flight validation (P1..P4)

**Closest analog:** `validate.go` (root pkg) for the structural pattern (accumulate → sort → return); Phase 8 `NewScorer` for the validation-pipeline-ordering pattern (`scorer.go:130-200` design-notes block).

**Pattern 1 — Pipeline-ordering godoc** (analog: scorer.go:136-148 validation-order block):

```go
// scorer.go:136-148 — copy this structure for scan.validateCheck:
//
//  1. Missing-threshold (ErrMissingThreshold) — fires FIRST so a user
//     who forgets WithThreshold AND has another option problem still
//     sees a clear "you forgot the threshold" message.
//  2. Empty-algorithms (ErrEmptyScorer)
//  3. Defensive per-entry AlgoID bounds + dispatch nil-check
```

The scan-specific pipeline (per 09-CONTEXT.md §2 + Pattern 1 of RESEARCH.md):

```
P1: cfg.Scorer == nil                                    → fail-fast → ErrNilScorer
P2: cfg.CrossGroupThresholdBoost validation              → fail-fast → ErrInvalidConfig
P3: Items[] validation (empty Name + duplicate (N,G))    → collect-all → errors.Join of ErrInvalidItem wraps
P4: SuppressedPairs validation (empty strings)           → collect-all → errors.Join of ErrInvalidConfig wraps
```

**Pattern 2 — `errors.Join` collect-all** (analog: pattern not yet in the repo — locked in Phase 9 per CONTEXT.md). The drop-in template is 09-RESEARCH.md lines 325–356:

```go
func validateItems(items []Item) error {
    var errs []error
    seen := make(map[itemKey]int, len(items))

    for i, item := range items {
        // D-03: empty Name
        if item.Name == "" {
            errs = append(errs, fmt.Errorf(
                "scan: invalid item at index %d: empty name: %w",
                i, ErrInvalidItem,
            ))
            continue // don't double-flag a same-index duplicate
        }
        // D-06: duplicate (Name, Group)
        k := itemKey{Name: item.Name, Group: item.Group}
        if first, ok := seen[k]; ok {
            errs = append(errs, fmt.Errorf(
                "scan: invalid item at index %d: duplicate (Name, Group) of index %d: %w",
                i, first, ErrInvalidItem,
            ))
            continue
        }
        seen[k] = i
    }
    return errors.Join(errs...)
}
```

**Pattern 3 — `sort.SliceStable` final ordering before emission** (analog: validate.go:252-257):

```go
// validate.go:252-257 — pattern for sorting validate output. scan's
// errors.Join keeps errors in index-ascending order by walking items[]
// in slice order; no post-sort needed because the natural-order walk
// already produces ascending indices. Document this explicitly in
// validateCheck's godoc.
sort.SliceStable(warnings, func(i, j int) bool {
    if warnings[i].Algorithm != warnings[j].Algorithm {
        return warnings[i].Algorithm < warnings[j].Algorithm
    }
    return warnings[i].Kind < warnings[j].Kind
})
```

**Pattern notes:** validate.go's pattern is for output-warning ordering. scan/validate.go's `errors.Join` order is implicitly ascending because the for-loop walks `items[]` in index order; document this explicitly so the planner doesn't add a post-sort step.

---

### `scan/bucket.go` — token-bucket optimisation

**Closest analog:** No exact in-repo precedent — this is the first internal-only candidate-enumeration optimisation. Composite analog from two sources:

1. **Consumer of `Tokenise`** (`tokenise.go:179`, `tokenise.go:105` for `DefaultTokeniseOptions`):
   ```go
   tokens := fuzzymatch.Tokenise(name, fuzzymatch.DefaultTokeniseOptions())
   ```

2. **Sorted-slice-as-canonical-order pattern** (scorer.go:95-106):
   ```go
   // scorer.go:95-106 — algorithmsAlgoIDSorted slice + iterate the slice
   // (not the map) on the output path. scan/bucket.go follows the same
   // discipline: map[string][]int internally, but iterate the source
   // item's tokens (a slice, naturally ordered) to assemble the candidate
   // set; sort candidate-set indices before scoring.
   algorithmsAlgoIDSorted []scorerEntry
   ```

**Pattern 1 — Private package constant with empirical-validation comment** (analog: validate.go:60-73 `validatePathologicalThreshold`):

```go
// validate.go:60-73 — pattern for the documented private const.
const validatePathologicalThreshold = 65_536
```

For `scan/bucket.go`, per 09-CONTEXT.md D-08 (locked exact wording):

```go
// bucketThreshold is the group-size cutoff at which the token-bucket
// optimisation overtakes the naive nested-loop comparison. Empirically
// validated on darwin/arm64 by BenchmarkScanCheck_BucketVsNaive_GroupSize
// at $WALL_TIME / Plan 09-04.
const bucketThreshold = 50
```

**Pattern 2 — Bucket build using map[string][]int, never iterated on output** (analog: pattern derived from the "no map iteration on output" rule and from scorer.go's `algorithmsAlgoIDSorted` discipline):

```go
// Per 09-RESEARCH.md anti-patterns: map iteration is contained inside
// bucket build and candidate-set construction; output paths iterate
// SLICES, not MAPS. The bucket is map[token][]itemIdx; candidate sets
// are built per source-item by walking that source item's tokens (a
// slice, naturally ordered) and accumulating into a deduped slice
// that is sorted before scoring.
type tokenBucket map[string][]int
```

**Pattern notes:**
- `bucketThreshold` comment MUST cite the wall-clock crossover and the benchmark name per D-08.
- Tokenisation happens at most once per Item — pre-compute `tokenisedNames [][]string` at Check entry (Pitfall 7 in 09-RESEARCH.md).
- The map is consumed in source-item order to build candidate slices; the slices are then sorted before scoring (no map iteration on output).
- Per Open Question 1 + 2 (09-RESEARCH.md): Normalise first via the new `Scorer.NormalisationOptions()` accessor, then `Tokenise(normalised, DefaultTokeniseOptions())`.

---

### `scan/suppress.go` — suppression composition predicate

**Closest analog:** Predicate helpers in `validate.go` (e.g., `hasOnlyNonASCII` at validate.go:311-328). Partial match — same predicate shape, different domain.

**Pattern 1 — Predicate signature + short-circuit on cheapest check** (analog: validate.go:311-328):

```go
// validate.go:311-328 — short-circuit predicate pattern (cheapest check
// first; return on first true).
func hasOnlyNonASCII(s string) bool {
    if s == "" {
        return false
    }
    for i := 0; i < len(s); {
        r, size := utf8.DecodeRuneInString(s[i:])
        if r < 0x80 && r != utf8.RuneError {
            return false
        }
        i += size
    }
    return true
}
```

**Pattern 2 — Three suppression rules composed via OR** (template in 09-RESEARCH.md lines 369-393):

```go
// 09-RESEARCH.md lines 370-393 — drop-in template; mirror the
// short-circuit ordering (cheapest check first):
//   Rule 1 (cheapest): per-item SilenceLint
//   Rule 2: SuppressedPairs map lookup (O(1))
//   Rule 3: cross-group identical-name (kind + string-equality check)
func isSuppressed(a, b Item, kind Kind, na, nb string, sc suppressionCtx) bool {
    if a.SilenceLint || b.SilenceLint {
        return true
    }
    if _, ok := sc.suppressedPairs[canonicalPair(na, nb)]; ok {
        return true
    }
    if kind == KindAcrossGroups && na == nb && !sc.compareIdenticalAcrossGroups {
        return true
    }
    return false
}
```

**Pattern notes:** No analog for `canonicalPair` ordering; use the lexicographic min/max pattern verbatim from 09-RESEARCH.md lines 397–402. SuppressedPairs entries are normalised once at Check entry (Pitfall 4 in 09-RESEARCH.md).

---

### `scan/doc.go` — package godoc landing page

**Closest analog:** Top-of-file godoc in `tokenise.go` (lines 1-100) and `scorer.go` (lines 15-74).

**Pattern — Three sections: (1) one-line summary; (2) usage example; (3) design notes** (analog: any of scorer.go / validate.go / tokenise.go top-of-file blocks).

```go
// Copyright 2026 AxonOps Limited
// [Apache 2.0 boilerplate]

// Package scan is the Layer-3 collection-scan sub-package of the
// fuzzymatch library. It composes a Phase 8 *fuzzymatch.Scorer with a
// pre-flight validation pass, within/cross-group similarity passes, a
// token-bucket optimisation for large groups, suppression composition,
// and a deterministic output sort.
//
// Typical usage:
//
//	items := []scan.Item{
//	    {Name: "user_id",  Group: "login"},
//	    {Name: "userId",   Group: "login"},
//	    {Name: "user_name", Group: "profile"},
//	}
//	cfg := scan.DefaultConfig(fuzzymatch.DefaultScorer())
//	cfg.CompareAcrossGroups = true
//	warnings, err := scan.Check(items, cfg)
//
// See docs/requirements.md §12 for the authoritative specification and
// docs/scan.md for the consumer-facing guide.
package scan
```

---

### `scan/scan_test.go` — black-box unit tests

**Closest analog:** `scorer_test.go` (`package fuzzymatch_test`, table-driven `Test*`, sub-tests, `t.Parallel()`).

**Pattern — Black-box test package + Test* function naming** (analog: scorer_test.go top of file):

```go
package scan_test

import (
    "errors"
    "testing"

    "github.com/axonops/fuzzymatch"
    "github.com/axonops/fuzzymatch/scan"
)

func TestCheck_WithinGroup_BasicMatch(t *testing.T) {
    t.Parallel()
    // table-driven
}

func TestCheck_ErrNilScorer(t *testing.T) {
    t.Parallel()
    _, err := scan.Check(nil, scan.Config{Scorer: nil})
    if !errors.Is(err, scan.ErrNilScorer) {
        t.Fatalf("want errors.Is ErrNilScorer, got %v", err)
    }
}
```

**Pattern notes:** Stdlib `testing` only (no testify in root tests per CLAUDE.md). Coverage targets ≥ 95% overall, ≥ 90% per file, 100% public API. Validation Architecture's req-ID → test map in 09-RESEARCH.md lines 569-580 lists every Test* expected.

---

### `scan/scan_internal_test.go` — internal-helper tests

**Closest analog:** `scorer_internal_test.go` (`package fuzzymatch`, tests unexported helpers).

**Pattern — Internal test package** (analog: scorer_internal_test.go header):

```go
package scan

import "testing"

// Tests for unexported helpers: canonicalPair, validateItems internals,
// bucket helpers, tokenBucket dedup, etc.
func Test_canonicalPair_LexOrder(t *testing.T) {
    // ...
}

func Test_bucketThreshold_Value(t *testing.T) {
    // pins the constant value; updates land alongside D-08 empirical
    // validation in plan 09-04.
}
```

---

### `scan/scan_bench_test.go` — benchmarks + D-08 sweep

**Closest analog:** `scorer_bench_test.go` (b.ReportAllocs, sized inputs, sink-gate pattern).

**Pattern 1 — Header docs + budget table** (analog: scorer_bench_test.go:14-58):

```go
// scorer_bench_test.go:14-58 — copy this header structure for
// scan_bench_test.go. Document:
//   - per-benchmark budget (e.g. PERF-05: < 2s for 10k items)
//   - b.ReportAllocs() on every benchmark
//   - sink-gate (var sink + sink-after-loop) to defeat compiler DCE
//   - construction OUTSIDE b.ResetTimer()
```

**Pattern 2 — Sized-input bench sweep** (analog: scorer_bench_test.go medium/short/long pattern):

```go
// For D-08's BucketVsNaive sweep, mirror the scorer_bench_test.go
// pattern of multiple BenchmarkX_Size variants:
func BenchmarkScanCheck_BucketVsNaive_GroupSize10(b *testing.B)  { benchScanCrossover(b, 10) }
func BenchmarkScanCheck_BucketVsNaive_GroupSize25(b *testing.B)  { benchScanCrossover(b, 25) }
func BenchmarkScanCheck_BucketVsNaive_GroupSize50(b *testing.B)  { benchScanCrossover(b, 50) }
func BenchmarkScanCheck_BucketVsNaive_GroupSize75(b *testing.B)  { benchScanCrossover(b, 75) }
func BenchmarkScanCheck_BucketVsNaive_GroupSize100(b *testing.B) { benchScanCrossover(b, 100) }
func BenchmarkScanCheck_BucketVsNaive_GroupSize200(b *testing.B) { benchScanCrossover(b, 200) }
func BenchmarkScanCheck_DefaultScorer_10k(b *testing.B)          { /* PERF-05 */ }
```

**Pattern notes:** Per 09-CONTEXT.md D-08, results commit to `bench.txt` at Plan 09-04 close; the chosen `bucketThreshold` value is committed alongside a one-line comment citing the wall-clock crossover.

---

### `scan/example_test.go` — godoc runnable examples

**Closest analog:** Runnable examples patterns across the root pkg (typically embedded in `*_test.go` files near the algorithm tests; the canonical pattern is a `func ExampleX()` with `// Output:` comments).

**Pattern — One `Example*` per public symbol** (analog discipline from DX-02 in 09-RESEARCH.md lines 575-580):

```go
package scan_test

import (
    "fmt"

    "github.com/axonops/fuzzymatch"
    "github.com/axonops/fuzzymatch/scan"
)

func ExampleCheck_WithinGroup() {
    items := []scan.Item{
        {Name: "user_id", Group: "login"},
        {Name: "userId",  Group: "login"},
    }
    cfg := scan.DefaultConfig(fuzzymatch.DefaultScorer())
    warnings, _ := scan.Check(items, cfg)
    fmt.Println(len(warnings))
    // Output: 1
}

func ExampleKind_String() {
    fmt.Println(scan.KindWithinGroup.String())
    fmt.Println(scan.KindAcrossGroups.String())
    // Output:
    // WithinGroup
    // AcrossGroups
}
```

**Minimum required examples (per 09-RESEARCH.md line 602):**
1. `ExampleCheck_WithinGroup`
2. `ExampleCheck_CrossGroup`
3. `ExampleCheck_WithSuppression`
4. `ExampleCheck_WithDefaultConfig`
5. `ExampleDefaultConfig`
6. `ExampleKind_String`

---

### `scan/props_test.go` — property tests via `testing/quick`

**Closest analog:** `scorer_test.go:832-982` (PropScorer_DeterministicAcrossRuns, PropScorer_WeightSumOne).

**Pattern — `quick.Check` with `quick.Config{MaxCount: N}`** (analog: scorer_test.go:854-866):

```go
// scorer_test.go:854-866 — drop-in pattern for scan property tests.
func TestProp_Scorer_DeterministicAcrossRuns(t *testing.T) {
    t.Parallel()
    f := func(a, b string) bool {
        s1 := fuzzymatch.DefaultScorer()
        s2 := fuzzymatch.DefaultScorer()
        return s1.Score(a, b) == s2.Score(a, b)
    }
    if err := quick.Check(f, nil); err != nil {
        t.Errorf("PropScorer_DeterministicAcrossRuns: %v", err)
    }
}
```

**Required properties for scan (09-RESEARCH.md line 603):**
- `PropCheck_BucketEquivalentToNaive` (SCAN-02 load-bearing)
- `PropCheck_DeterministicAcrossRuns`
- `PropCheck_NoSelfWarnings`
- `PropCheck_NoNaN`
- `PropCheck_NoInf`

`MaxCount: 100` (the testing/quick default) is sufficient for SCAN-02; for bucket-equivalence, generate `[]Item` via a custom `quick.Generator` so the property doesn't blow up at 0-item inputs.

---

### `scan/fuzz_test.go` — native Go fuzz

**Closest analog:** `damerau_full_fuzz_test.go`, `cosine_fuzz_test.go` (typical FuzzX with seed corpus).

**Pattern — FuzzX function with seed corpus** (analog: damerau_full_fuzz_test.go top):

```go
func FuzzCheck(f *testing.F) {
    // Seed corpus
    f.Add("user_id", "userId")
    f.Add("",        "x")           // ErrInvalidItem boundary
    f.Add("a", "a")                  // duplicate when same group

    f.Fuzz(func(t *testing.T, a, b string) {
        items := []scan.Item{
            {Name: a, Group: "g"},
            {Name: b, Group: "g"},
        }
        _, _ = scan.Check(items, scan.DefaultConfig(fuzzymatch.DefaultScorer()))
        // assertion: no panic on consumer-supplied input
    })
}
```

**Pattern notes:** Per Phase 9 §Security Domain (09-RESEARCH.md lines 632-642), `scan.Check` never panics on consumer-supplied input. The fuzzer's invariant is "no panic, error is acceptable."

---

### `scan/scan_golden_test.go` — golden file diff test

**Closest analog:** `scorer_golden_test.go` (verbatim envelope + entry struct + `writeGoldenFile` helper).

**Pattern 1 — File-header design block** (analog: scorer_golden_test.go:15-49):

```go
// scorer_golden_test.go:15-49 — copy this header structure. Key points:
//   - schema follows the LOCKED canonical-form contract from
//     golden_canonical.go
//   - map keys use AlgoID.String() (NOT integer-form) — preserves
//     grep-ability
//   - _metadata.generated_at INTENTIONALLY OMITTED (byte-identical CI
//     matrix gate; same as algorithms.json / scorer-default.json)
```

**Pattern 2 — Entry struct + envelope** (analog: scorer_golden_test.go:60-94):

```go
type scanGoldenEntry struct {
    Items    []scan.Item     `json:"items"`
    Warnings []scanWarningJS `json:"warnings"`
    Config   string          `json:"config"`
}

type scanGoldenMetadata struct {
    Phase           int    `json:"phase"`
    ScannerSignature string `json:"scanner_signature"`
}

type scanGoldenFile struct {
    Metadata scanGoldenMetadata `json:"_metadata"`
    Entries  []scanGoldenEntry  `json:"entries"`
}
```

**Pattern 3 — `scoreAllAsStringKeys` helper for JSON-friendly keys** (analog: scorer_golden_test.go:105-110):

```go
func scoreAllAsStringKeys(m map[fuzzymatch.AlgoID]float64) map[string]float64 {
    out := make(map[string]float64, len(m))
    for id, score := range m {
        out[id.String()] = score
    }
    return out
}
```

**Pattern notes:** Use `writeGoldenFile` from `golden_canonical.go:95` (already in-tree). Same `-update` flag pattern as `TestGolden_ScorerDefault`. Confirm `Makefile`'s `verify-determinism` enumerates `testdata/golden/*.json` rather than hard-coding filenames (09-RESEARCH.md Assumption A7 — flag for Plan 09-06 / 09-08 inspection).

---

### `testdata/golden/scan-default.json` — golden corpus

**Closest analog:** `testdata/golden/scorer-default.json` (Phase 8 envelope shape).

**Pattern — Envelope shape** (scorer-default.json:1-21):

```json
{
  "_metadata": {
    "phase": 9,
    "scanner_signature": "DefaultConfig-2026-MM-DD"
  },
  "entries": [
    {
      "config": "DefaultConfig",
      "items": [
        {"name": "user_id",  "group": "login"},
        {"name": "userId",   "group": "login"}
      ],
      "warnings": [
        {
          "kind": "WithinGroup",
          "nameA": "user_id",
          "nameB": "userId",
          "groupA": "login",
          "groupB": "login",
          "score": 0.9999999999999999,
          "scores": {
            "DamerauLevenshteinOSA": 1,
            "DoubleMetaphone": 1,
            "JaroWinkler": 1,
            "QGramJaccard": 1,
            "SorensenDice": 1,
            "TokenJaccard": 1
          }
        }
      ]
    }
  ]
}
```

**Pattern notes:** `scores` keys use `AlgoID.String()` (CamelCase), matching scorer-default.json. `kind` is `Kind.String()` (CamelCase: "WithinGroup"/"AcrossGroups"). No `generated_at` field (byte-identical CI matrix gate).

---

### `testdata/golden/_staging/scan.json` — staging file

**Closest analog:** Any file in `testdata/golden/_staging/` (existing staging convention).

**Pattern — Same shape as the final golden; staging is the pre-merge form regenerated by `-update` and diff-reviewed before promotion.** No additional pattern beyond the convention.

---

### `tests/bdd/features/scan.feature` — Gherkin scenarios

**Closest analog:** `tests/bdd/features/scorer.feature` (Phase 8 — 12 scenarios per CONTEXT.md class taxonomy).

**Pattern — File-header docstring + `@tag` discipline + per-scenario CONTEXT.md class citation** (scorer.feature:1-28):

```gherkin
# Phase 9 BDD coverage: collection-scan sub-package.
#
# Scenarios per D-03..D-07 decision (CONTEXT.md):
#   D-03 empty-Name validation       -> "Check rejects items with empty Name"
#   D-04 boost validation             -> "Check rejects Config with NaN boost"
#   D-05 SuppressedPairs validation  -> "Check rejects SuppressedPairs with empty string"
#   D-06 duplicate (Name,Group)       -> "Check rejects duplicate (Name, Group) items"
#   D-07 Validate boundary            -> "Check does not call Validate internally"
# Plus the within-group, cross-group, threshold-boost, sort-key, golden,
# and Phase 8.5 R2 deferred suppression scenarios.

@scan
Feature: Collection scan sub-package (Phase 9)
  The scan sub-package composes a Phase 8 Scorer with a pre-flight
  validation pass, within/cross-group similarity passes, a token-bucket
  optimisation, suppression composition, and deterministic sort.

  @scan @default
  Scenario: Check produces one within-group warning for similar names
    Given I have a scan items list
      | name    | group |
      | user_id | login |
      | userId  | login |
    When I call scan.Check with the default Config
    Then the scan should produce 1 warning
    And the warning should be classified as WithinGroup
```

**Pattern notes:** Minimum coverage per 09-CONTEXT.md "Claude's Discretion" is one scenario per D-03..D-07. bdd-scenario-reviewer adjusts upward as needed.

---

### `tests/bdd/features/suppression.feature` — Gherkin scenarios

**Closest analog:** `tests/bdd/features/scorer.feature` for shape; `tests/bdd/features/validate.feature` for the "one scenario per Kind" discipline.

**Pattern — Per-suppression-mode scenarios** (analog: validate.feature:24-58 per-Kind):

```gherkin
@suppression
Feature: Suppression composition in scan (Phase 9)
  Three suppression rules compose via OR:
    1. Per-item SilenceLint
    2. Pair in SuppressedPairs
    3. Cross-group identical-name (when CompareIdenticalAcrossGroups=false)

  @suppression @silencelint
  Scenario: SilenceLint suppresses within-group warnings
    # 09-CONTEXT.md §3 / SCAN-03 — per-item SilenceLint silences any
    # warning where this Item is on either side of the pair.
    Given I have a scan items list with SilenceLint=true on one item
    When I call scan.Check
    Then the scan should produce 0 warnings

  @suppression @pairs
  Scenario: SuppressedPairs silences canonical pair regardless of casing
    # 09-CONTEXT.md §3 + Pitfall 4 — SuppressedPairs entries are
    # canonicalised at Check entry using the Scorer's Normalise.
    ...
```

**Pattern notes:** Phase 8.5 R2 deferred suppression scenarios (per CONTEXT.md SCAN-03/04 + 09-RESEARCH.md line 47) land here. bdd-scenario-reviewer determines scenario count.

---

### `tests/bdd/steps/scan_steps.go` — godog step definitions

**Closest analog:** `tests/bdd/steps/scorer_steps.go` (verbatim shape: Context struct + Init function + ctx.Step regex pattern).

**Pattern 1 — Context struct + Init function** (scorer_steps.go:61-83 + 738-740):

```go
// scorer_steps.go:61-83 — pattern for ScanContext.
type ScanContext struct {
    items    []scan.Item
    cfg      scan.Config
    warnings []scan.Warning
    err      error
}

// scorer_steps.go:738-740 — pattern for InitScanSteps.
func InitScanSteps(ctx *godog.ScenarioContext) {
    sc := &ScanContext{}
    ctx.Step(
        `^I have a scan items list$`,
        sc.iHaveAScanItemsList,
    )
    // ... etc
}
```

**Pattern 2 — Wire into algorithms_steps.go's InitializeScenario** (algorithms_steps.go:1586 / 1593 / 1600 / 1607 — the InitX(ctx) calls at the bottom):

```go
// algorithms_steps.go:1581-1607 — add InitScanSteps(ctx) at the end:
InitScorerSteps(ctx)
InitValidateSteps(ctx)
InitNormalisationSteps(ctx)
InitDeterminismSteps(ctx)
InitScanSteps(ctx)  // ← Phase 9 addition
```

**Pattern notes:** testify (`assert` / `require`) is permitted in BDD step files per CLAUDE.md (`tests/bdd/` ONLY). goleak.VerifyTestMain in tests/bdd/bdd_test.go covers scan inherently — scan is pure-function with no goroutines.

---

### `examples/scan-demo/main.go` — runnable demo

**Closest analog:** `examples/scorer-composition/main.go` (verbatim shape: package main + var-pairs + ctor helper + main loop).

**Pattern 1 — File header + run command + pairs slice** (scorer-composition/main.go:15-62):

```go
// examples/scorer-composition/main.go:15-62 — pattern for scan-demo:
//   - file-header godoc documenting the demo's narrative
//   - "Run with: go run ./examples/scan-demo/" at the bottom of the doc
//   - var pairs (or var items) declared at package scope so the
//     stdout-meta-test can reference them by index
//   - var defaultS = fuzzymatch.DefaultScorer() built once

var items = []scan.Item{
    {Name: "user_id",   Group: "login"},
    {Name: "userId",    Group: "login"},
    {Name: "user_name", Group: "profile"},
    // ...
}
```

**Pattern 2 — newScorer helper with panic on invariant violation** (scorer-composition/main.go:89-105):

```go
// scorer-composition/main.go:89-105 — pattern for newScanConfig helper
// if scan-demo wants a non-default Config. Panic on programmer-error
// because the composition is statically valid.
func newMinusDMScorer() *fuzzymatch.Scorer {
    opts := append(
        fuzzymatch.DefaultScorerOptions(),
        fuzzymatch.WithoutAlgorithm(fuzzymatch.AlgoDoubleMetaphone),
        fuzzymatch.WithThreshold(0.80),
    )
    s, err := fuzzymatch.NewScorer(opts...)
    if err != nil {
        panic(fmt.Sprintf("scorer-composition example: NewScorer failed unexpectedly: %v", err))
    }
    return s
}
```

**Pattern 3 — Three demo modes required by 09-CONTEXT.md Claude's Discretion** (must demonstrate):
1. Within-group only with the default Scorer.
2. Within + cross-group with `DefaultConfig`.
3. At least one of each suppression mode (per-item `SilenceLint`, `SuppressedPairs`, identical-cross suppression).

---

### `examples/scan-demo/main_test.go` — stdout meta-test

**Closest analog:** `examples/scorer-composition/main_test.go` (os.Pipe redirect + byte-for-byte `want` constant).

**Pattern — Committed `want` constant + os.Pipe stdout redirect** (scorer-composition/main_test.go:15-50):

```go
// examples/scorer-composition/main_test.go:15-50 — pattern for
// scan-demo/main_test.go:
//   - "want" constant captures exact bytes
//   - os.Pipe() redirect of stdout
//   - line-by-line diff in t.Errorf

const want = `Pair (a / b)                         Default     MinusDM        MDef        MMDM
--------------------------------------------------------------------------------
user_id / userId                      1.0000      1.0000        true        true
...`
```

**Pattern notes:** Regenerate `want` by running `go run .` and pasting; any diff requires reviewed update (golden-style discipline).

---

### `examples/scan-demo/go.mod` — per-example module

**Closest analog:** `examples/scorer-composition/go.mod` (verbatim 4-line shape).

```text
module github.com/axonops/fuzzymatch/examples/scan-demo

go 1.26.3

replace github.com/axonops/fuzzymatch => ../..

require github.com/axonops/fuzzymatch v0.0.0-00010101000000-000000000000

require golang.org/x/text v0.37.0 // indirect
```

**Pattern notes:** Each example is its own module per CLAUDE.md "Test deps isolated"; `replace` directive uses `../..` because the example lives at depth 2 from the root.

---

### `docs/scan.md` — user-facing documentation

**Closest analog:** `docs/scorer.md` (Quickstart → Custom composition → Concepts → Thread safety).

**Pattern — Section structure** (scorer.md:1-78):

```markdown
# Scan sub-package

[One-paragraph intro positioning scan in the three-layer architecture.]

## Quickstart

[Minimal example with scan.DefaultConfig and scan.Check.]

## Items and groups

[Item struct fields + group semantics.]

## Within-group vs cross-group passes

[Two-pass model; threshold boost; identical-name suppression default.]

## Suppression composition

[Three rules; canonicalisation of SuppressedPairs at Check entry.]

## Validation

[D-03/D-05/D-06 collect-all behaviour; errors.Is / errors.Join discrimination.]

## Determinism

[Sort key; in-line completeness assertion; golden file location.]

## Performance

[< 2s/10k budget; token-bucket optimisation; bucketThreshold private const rationale.]

## Validate / scan boundary (D-07)

[Pattern: call fuzzymatch.Validate(item.Name) upstream of scan.Check.]

## Out of scope

[V2-API-03 streaming; V2-API-04 parallel; Config.BucketThreshold deferred.]

## Thread Safety

[scan.Config is plain data; scan.Check is pure; safe on disjoint inputs.]
```

**Pattern notes:** Current `docs/scan.md` is a TBD scaffold (58 lines). Plan 09-08 populates fully.

---

### `scorer.go` (MODIFIED) — add `NormalisationOptions()` accessor

**Closest analog:** Existing `Threshold()` (scorer.go:447-449) and `Algorithms()` (scorer.go:466-472) accessors.

**Pattern — Read-only accessor returning the immutable internal value** (scorer.go:447-449):

```go
// scorer.go:447-449 — pattern for the new NormalisationOptions accessor.
func (s *Scorer) Threshold() float64 {
    return s.threshold
}
```

For Phase 9's resolution of Open Question 1 (api-ergonomics-reviewer picks the exact name; recommendation per 09-RESEARCH.md line 872 is `NormalisationOptions`):

```go
// NormalisationOptions returns the NormalisationOptions stored at
// construction time, along with a boolean indicating whether the
// Scorer applies normalisation (false when WithoutNormalisation was
// applied).
//
// The returned options are by-value, mirroring the immutability of the
// Scorer: callers cannot mutate the Scorer's internal options through
// this surface.
//
// Used by the scan sub-package (github.com/axonops/fuzzymatch/scan) to
// build token buckets and canonicalise SuppressedPairs entries using
// the same normalisation pipeline the Scorer uses for scoring (per spec
// §12.3 + §12.5 + 09-CONTEXT.md §4 / 09-RESEARCH.md Open Question 1).
//
// Safe for concurrent use; the Scorer is immutable after NewScorer.
func (s *Scorer) NormalisationOptions() (opts NormalisationOptions, applied bool) {
    return s.normaliseOpts, s.applyNormalisation
}
```

**Pattern notes:** This is a new public method on Phase 8 Scorer; api-ergonomics-reviewer sign-off recorded in Plan 09-01 PR. Spec amendment to docs/requirements.md §8 in lockstep.

---

### `docs/requirements.md` (MODIFIED) — spec amendments

**Closest analog:** The Phase 8 amendment trail (Plan 08-03 amended §8.3 + §8.6 in lockstep with `Scorer.ScoreAll` returning `map[AlgoID]float64`).

**Pattern — Inline `SPEC OVERRIDE (Phase N)` block with back-reference** (analog: in-source pattern at scorer.go:476):

```text
# docs/requirements.md §12.1 (after amendment, per 09-RESEARCH.md
# Pattern 1 + State of the Art table):
type Warning struct {
    Kind           Kind
    NameA, NameB   string
    GroupA, GroupB string
    TagA, TagB     any
    Score          float64

    // SPEC OVERRIDE (Phase 9): Scores is map[AlgoID]float64; see
    // 09-CONTEXT.md §1 D-01. The Phase 8 ScoreAll override at §8.3
    // + §8.6 applies verbatim here for the same typed-enum-keys
    // rationale.
    Scores map[AlgoID]float64
}
```

**Required edits (per 09-CONTEXT.md §"Specific Ideas" + 09-RESEARCH.md lines 837-841):**
- §12.1 lines 1306-1322: `WarningKind` → `Kind` rename
- §12.1 line 1337: `Scores map[string]float64` → `Scores map[AlgoID]float64`
- §12.1 line 1359: default-0.05 statement migrates to `DefaultConfig` godoc
- §12.1 line 1389: "the offending item's index" → "every offending index, joined via errors.Join"
- §12.4 + §12.6 + §15: every `WarningKind` reference renamed to `Kind`
- §19 line 2010 (REQUIREMENTS row in phase plan): `WarningKind` reference renamed
- §19 line 2061 (Scan acceptance criteria): `WarningKind` reference renamed

---

## Shared Patterns

### Authentication / Authorization

**N/A.** scan is a pure-function library with no identity context, no session state, no access boundaries. Consumer enforces all auth.

---

### Error Handling

**Source:** `errors.go` (root pkg, lines 15-35 file header + per-sentinel four-section godoc template at lines 157-170 ErrEmptyScorer).

**Apply to:** `scan/errors.go` (verbatim template), `scan/scan.go` (in-line completeness assertion panic), `scan/validate.go` (errors.Join collect-all per RESEARCH.md Pattern 2).

**Concrete excerpt** (errors.go:157-170):

```go
// ErrEmptyScorer indicates NewScorer was called without any algorithm
// option — the option slice contained zero WithAlgorithm /
// With*Algorithm entries by the time validation ran. A Scorer with no
// algorithms has no meaningful composite to compute.
//
// Pass at least one WithAlgorithm option (or use DefaultScorer() for
// the opinionated six-algorithm composition).
//
// Returned by NewScorer (Phase 8) after the missing-threshold check
// passes and the option-validation pipeline finds cfg.entries empty.
//
// Discriminate via errors.Is(err, fuzzymatch.ErrEmptyScorer); never
// match the error message string.
var ErrEmptyScorer = errors.New("fuzzymatch: scorer has no algorithms (pass at least one WithAlgorithm option or use DefaultScorer)")
```

**`errors.Join` collect-all** (NEW pattern locked in Phase 9; reference template at 09-RESEARCH.md lines 325-356):

```go
var errs []error
for i, item := range items {
    if item.Name == "" {
        errs = append(errs, fmt.Errorf(
            "scan: invalid item at index %d: empty name: %w",
            i, ErrInvalidItem,
        ))
    }
}
return errors.Join(errs...)
```

---

### Validation

**Source:** `validate.go` (root pkg; sort.SliceStable + complete-sort-key + nil-vs-empty discipline) + Phase 8 `NewScorer` validation pipeline at `scorer.go:130-200`.

**Apply to:** `scan/validate.go` (all four validation phases P1..P4).

**Concrete excerpt — validation-pipeline ordering** (scorer.go:136-148):

```go
// NewScorer runs the validation pipeline in the order LOCKED by
// CONTEXT.md §2:
//
//  1. Missing-threshold (ErrMissingThreshold) — fires FIRST so a user
//     who forgets WithThreshold AND has another option problem still
//     sees a clear "you forgot the threshold" message.
//  2. Empty-algorithms (ErrEmptyScorer)
//  3. Defensive per-entry AlgoID bounds + dispatch nil-check
//     (ErrInvalidAlgoID).
```

**Concrete excerpt — sort.SliceStable with complete key** (validate.go:252-257):

```go
sort.SliceStable(warnings, func(i, j int) bool {
    if warnings[i].Algorithm != warnings[j].Algorithm {
        return warnings[i].Algorithm < warnings[j].Algorithm
    }
    return warnings[i].Kind < warnings[j].Kind
})
```

---

### Determinism

**Source:** `validate.go:246-260` (sort.SliceStable + sorted-slice iteration), `scorer.go:514-523` (iterate sorted slice, write into map for output).

**Apply to:** `scan/scan.go` (sort.SliceStable final output sort + in-line completeness assertion), `scan/bucket.go` (map for storage, slice for iteration), `scan/scan_golden_test.go` (`AlgoID.String()` map keys for human-readable JSON), `testdata/golden/scan-default.json` (no `generated_at` field).

**Concrete excerpt — write-into-map but iterate sorted slice** (scorer.go:514-523):

```go
// scorer.go:514-523 — pattern for any map-output path:
// iterate AlgoID-sorted SLICE; write into freshly allocated MAP.
// Consumers see deterministic CONTENTS even though Go's range over
// the returned map is randomised.
out := make(map[AlgoID]float64, len(s.algorithmsAlgoIDSorted))
for _, entry := range s.algorithmsAlgoIDSorted {
    out[entry.id] = entry.scoreFn(na, nb)
}
return out
```

---

### Thread Safety / Concurrency

**Source:** `scorer.go:60-66` (immutability godoc) + `scorer.go:348-351` (per-method "safe for concurrent use" guarantee).

**Apply to:** `scan/scan.go` Check godoc, `scan/scan.go` Config godoc, `scan/doc.go` package godoc.

**Concrete excerpt** (scorer.go:348-351):

```go
// Concurrency: Score is safe for concurrent use from any number of
// goroutines on the same *Scorer without external synchronisation. The
// Scorer is immutable after NewScorer returns; this method does no
// writes to the receiver's state.
```

For scan.Check, the equivalent is:

```go
// Concurrency: Check is safe for concurrent invocation on DISJOINT
// inputs. It is a pure function: no goroutines, no channels, no
// mutexes. Each invocation builds its own internal state; no shared
// mutable state exists. Concurrent invocations on the SAME items slice
// require the caller to ensure no mutation of the items between calls
// (the slice header is read-only inside Check, but the underlying
// strings are not copied — consumer's responsibility).
```

---

### Logging / Observability

**N/A.** scan is a pure-function library with no logging. Consumer observes via the returned `[]Warning` and `error`.

---

## No Analog Found

All 22 new files have at least a partial in-tree analog. The closest-to-greenfield areas are:

| File | Role | Data Flow | Reason |
|------|------|-----------|--------|
| `scan/bucket.go` | optimisation (private const + bucket structure) | candidate-set transform | First internal-only candidate-enumeration optimisation in the repo. Pattern is composite of `validate.go:60-73` (documented private const) + `scorer.go:95-106` (sorted-slice canonical-order discipline) + `tokenise.go` consumer call site. No single existing file matches all three. |
| `scan/suppress.go` | predicate (3 rules OR-ed) | predicate | Closest in-repo is `validate.go:311-328` `hasOnlyNonASCII` — same predicate shape, different domain. The three-rule OR composition is new. Drop-in template in 09-RESEARCH.md lines 369-393. |
| `scan/validate.go` (internal) | pre-flight pipeline | transform | `validate.go` (root) is the closest by domain but is the PUBLIC `Validate(a,b)` surface; the scan equivalent is INTERNAL. The shape of the `errors.Join` collect-all is NEW in Phase 9 — no existing precedent. |

For all three, the planner uses the **composite excerpts above** rather than a single-file copy.

---

## Metadata

**Analog search scope:**
- Root pkg `.go` files: `scorer.go`, `validate.go`, `warn_kind.go`, `errors.go`, `algoid.go`, `tokenise.go`, `normalise.go`, `scorer_test.go`, `scorer_bench_test.go`, `scorer_golden_test.go`, `golden_canonical.go`
- `tests/bdd/features/`: `scorer.feature`, `validate.feature`
- `tests/bdd/steps/`: `scorer_steps.go`, `algorithms_steps.go`
- `examples/`: `scorer-composition/main.go`, `scorer-composition/main_test.go`, `scorer-composition/go.mod`
- `testdata/golden/`: `scorer-default.json` (envelope shape)
- `docs/`: `scorer.md` (consumer doc shape), `scan.md` (current TBD scaffold)

**Files scanned:** 17 source files read in full or in targeted ranges; 8 directories listed.

**Pattern extraction date:** 2026-05-19

**Confidence:** HIGH — every analog cited by file+line number; the composite analogs for bucket.go / suppress.go / validate.go (internal) are explicitly flagged as composites with the three component sources named.

---

*Phase: 9 — Collection Scan Sub-package*
*Pattern mapping: 2026-05-19*
