---
phase: 08-composite-scorer
reviewer: determinism-reviewer
reviewed: 2026-05-17
depth: standard
files_reviewed:
  - scorer.go
  - scorer_options.go
  - scorer_golden_test.go
  - scorer_test.go
  - scorer_internal_test.go
  - testdata/golden/scorer-default.json
  - algoid.go
  - errors.go
  - normalise.go
  - golden_canonical.go
findings:
  blocking: 0
  medium: 3
  low: 4
  total: 7
status: GO_WITH_FOLLOWUPS
cross_platform_ci_status: prerequisite_not_yet_executed
contract_verified:
  parens_present_textual_matches: 7
  reduction_iterates_sorted_slice: true
  weight_normalisation_left_to_right: true
  init_functions_present: false
  math_transcendentals_present: false
  map_iteration_on_output_paths: false
  scorer_default_golden_generated_at_omitted: true
---

# Phase 8 — Composite Scorer — Determinism Review

## Summary

Phase 8 carries the LOCKED determinism contract forward correctly. The
reduction loop in `Scorer.Score` is byte-for-byte identical in pattern
to the canonical Phase 5 Cosine reduction at `cosine.go:343`: explicit
`(entry.weight * score)` parenthesisation, left-to-right
`acc = acc + …` accumulation, and iteration over a slice that was
sorted at `NewScorer` time using the canonical `AlgoIDs()` order. The
weight auto-normalisation step uses the same explicit-arithmetic
left-to-right pattern (`sum = sum + sorted[i].weight`,
`sorted[i].weight = sorted[i].weight / sum`). The full `(entry.weight *
score)` literal appears **7 times** across the source (godoc + comment
+ code), matching what 08-REVIEW.md recorded.

The Phase 8 surface introduces **no new transcendental float
operations** (no `math.Pow`, `math.Log`, `math.Exp`, `math.FMA`, no
parallel reduction); the package's only `math.X` call is the Phase 5
Cosine `math.Sqrt`, which is unchanged. No `init()` function appears
in any Phase 8 file. There is no map iteration on any output path:
`ScoreAll` populates a fresh map by iterating
`algorithmsAlgoIDSorted` (a slice); `Algorithms()` builds a fresh slice
by iterating the same slice; the dedup map at construction time
(`seen`) is read in `AlgoIDs()` iota order, never via `range seen`.

The `scorer-default.json` golden file is well-formed: 22 entries
spanning 5 Scorer compositions, `_metadata` carries only `phase` and
`scorer_signature` (no `generated_at`, no machine identifier, no
timestamps anywhere). Map values inside `scoreAll` serialise with
alphabetical key order via `encoding/json`'s built-in marshal sort, so
the on-disk bytes are stable regardless of the insertion order chosen
in the Go code.

**No BLOCKING findings.** Three MEDIUM findings flag fragile patterns
that should be hardened before v1.0 (chief among them DET-01: NaN
poisoning of the threshold is a determinism-adjacent correctness bug
already tracked as CR-01). Four LOW findings are defence-in-depth /
style. The Phase 8 cross-platform byte-identity claim is not
empirically verified yet — the golden was generated on darwin/arm64
and has not been diffed against the other four CI matrix platforms
(see "Cross-Platform CI prerequisite" at the end).

---

## Verification Checklist

### Map Iteration

- [x] No public function returns output whose order depends on Go map iteration order
- [x] No internal map iteration leaks into observable output ordering
- [x] Maps used internally (`seen` dedup, `ScoreAll` return) build deterministic outputs:
  - `seen` is populated from `range cfg.entries` (slice) and read back via `range AlgoIDs()` (slice) — not `range seen`
  - `ScoreAll`'s returned map is populated by `range s.algorithmsAlgoIDSorted` (slice); JSON marshal then sorts keys alphabetically
  - `Algorithms()` builds a fresh slice in AlgoID order from `s.algorithmsAlgoIDSorted`

The grep heuristic returns six `for ... range` sites in `scorer*.go`,
**none of which iterate a map**:

```
scorer_options.go:191 — range cfg.entries (slice)
scorer.go:229         — range cfg.entries (slice, validation)
scorer.go:242         — range cfg.entries (slice, dedup write)
scorer.go:369         — range s.algorithmsAlgoIDSorted (slice, Score reduction)
scorer.go:462         — range s.algorithmsAlgoIDSorted (slice, Algorithms())
scorer.go:513         — range s.algorithmsAlgoIDSorted (slice, ScoreAll)
```

The `for _, id := range AlgoIDs()` at `scorer.go:253` (Step 7) is the
only place a map (`seen`) is consulted, and it is consulted by *key
lookup*, not by range iteration. This is the canonical "extract keys,
sort them, iterate the sorted slice" pattern from the
determinism-standards skill — implemented here at compile time
(`AlgoIDs()` is a hand-written ordered slice literal in algoid.go).

### Sort Keys

Not applicable to Phase 8 — `Scorer.Score` returns a scalar; `ScoreAll`
returns a map (documented as non-deterministic iteration order);
`Algorithms()` returns a slice whose order is the canonical
`AlgoIDs()` iota order (a complete total ordering — every AlgoID is
unique by `int` value).

### Float Stability

- [x] No `math.Pow`, `math.Log`, `math.Exp`, `math.Atan2`, `math.SinH`, `math.CosH` anywhere in Phase 8 files
- [x] No `math.FMA` (the FMA-fusion caveat for the compiler-emitted variant on arm64 is documented and mitigated empirically — see DET-05 LOW)
- [x] Sum reductions are left-to-right in a single loop (weight normalisation Step 8; Score reduction at line 380)
- [x] No reliance on float operation reordering that the compiler could exploit — every accumulation site uses explicit `x = x + y` rather than `+=` or reordered associativity

### NaN / Inf / Negative Zero

- [x] `TestProp_Scorer_NoNaN_NoInf` (scorer_test.go:877) gates `DefaultScorer().Score` over random input — passes per 08-04 SUMMARY
- [x] `TestProp_Scorer_ScoreInRange` (scorer_test.go:860) gates `[0,1]` range
- [ ] **WithThreshold does NOT guard NaN** — see DET-01 MEDIUM (this is already CR-01 in 08-REVIEW.md; recorded here from the determinism perspective)
- [x] Negative-zero is not produced anywhere in the reduction (the only operations are `+` and `*` on non-negative algorithm scores in [0,1] and positive weights — IEEE-754 `0.0 + (w * 0.0) = 0.0`, not `-0.0`)
- [x] Defensive `sum == 0` guard on weight normalisation (scorer.go:284) prevents division-by-zero NaN/Inf escape

### Init() and Package Load

- [x] No `init()` functions in `scorer.go`, `scorer_options.go`, or `algoid.go`
- [x] No package-level mutable state introduced by Phase 8 (every variable is a struct field on `*Scorer` or `scorerConfig`)
- [x] `dispatch[]` registration is via per-file `var _ = func() bool { dispatch[X] = … ; return true }()` idioms already audited in earlier phases — Phase 8 does not add any new dispatch slot

### Concurrent-Use Determinism

- [x] `*Scorer` is immutable after `NewScorer` returns (every field is written once in the constructor's final `return &Scorer{…}` and never again)
- [x] No mutexes, no atomic ops, no goroutines
- [x] `TestScorer_ConcurrentSafety` (scorer_test.go:899) covers 100 goroutines × Score/ScoreAll/Match × 3 rounds — passes per 08-02 SUMMARY

### Golden Files

- [x] `testdata/golden/scorer-default.json` is byte-stable on the run that produced it (TestGolden_ScorerDefault re-run without `-update` exits 0 on darwin/arm64 per 08-04 SUMMARY)
- [x] `_metadata` has no incidental timestamps; only `phase: 8` and `scorer_signature: "DefaultScorer-2026-05-16"` (a static literal)
- [x] 22 entries × 5 Scorer compositions; insertion order is deterministic by construction (literal slice + literal pair list in `buildScorerGoldenEntries`)
- [x] `scoreAll` map values serialise with alphabetical key order via `encoding/json`'s built-in marshal sort — independent of `range m` iteration order in `scoreAllAsStringKeys`
- [ ] **Cross-platform empirical verification has not yet executed** — see "Cross-Platform CI prerequisite" at the end

### Cross-Platform CI

- [x] The test `TestGolden_ScorerDefault` is a plain Go test with no `//go:build` tag, no `runtime.GOOS` / `runtime.GOARCH` conditional, and no platform-specific assertion — it runs on all five matrix platforms uniformly
- [x] `assertGolden` reads the same `testdata/golden/scorer-default.json` byte stream on every platform; `bytes.Equal` is the comparator (not a tolerance-based one)
- [x] `canonicalMarshal` (`golden_canonical.go:61`) uses `json.MarshalIndent` with hard-coded `prefix=""`, `indent="  "`, appends a single LF — no CRLF risk on Windows
- [ ] **The matrix has not yet been run** — see prerequisite section at the end

---

## Findings

### DET-01: WithThreshold silently accepts NaN; Match becomes "never match" with no diagnostic
**Severity:** MEDIUM
**File:** scorer_options.go:257-266
**Issue:** `WithThreshold(t)` checks only `t < 0.0 || t > 1.0`. Both
comparisons evaluate to `false` for `t == math.NaN()`. The option
then stores NaN into `cfg.threshold` and sets `thresholdSet = true`.
The resulting `*Scorer.threshold` is NaN, and `(s *Scorer) Match`
returns `s.Score(a, b) >= s.threshold` — which is `false` for every
input (IEEE-754: `x >= NaN` is always `false`).

This is recorded as CR-01 in 08-REVIEW.md and is primarily a
correctness defect. I classify it MEDIUM **from the determinism
perspective** because:

1. The result is deterministic-but-wrong (`false` always), so the
   cross-platform byte-identity gate would *not* fail on this bug.
2. NaN values propagating through float arithmetic break the
   library's documented "no NaN" guarantee
   (`.claude/skills/determinism-standards/SKILL.md` §"NaN / Inf /
   Negative Zero"). Strictly, this defect leaks NaN into observable
   state (`Threshold()` would return NaN) even though `Score` itself
   never produces NaN.
3. `docs/scorer.md:283` documents the intended (NaN-rejecting)
   behaviour; the code is the regression.

**Impact:** No cross-platform byte-identity break. But a NaN value
inside `*Scorer` is a leak of "non-finite float" past the API
boundary and violates the project-wide NaN/Inf policy.

**Fix:** Add `math.IsNaN(t)` to the gate (as proposed in CR-01):

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

Already fully specified in 08-REVIEW.md CR-01 — no separate action
required from determinism-reviewer beyond confirming the
determinism-policy angle.

### DET-02: Defensive NaN/Inf propagation guard absent from the Score reduction
**Severity:** MEDIUM
**File:** scorer.go:368-382
**Issue:** The reduction loop trusts every `entry.scoreFn(na, nb)`
return value to be a finite float in `[0, 1]`. All 23 catalogue
algorithms satisfy this contract today (verified by
`TestProp_<algo>_NoNaN_NoInf` per-algorithm property tests), and the
Phase 8 `TestProp_Scorer_NoNaN_NoInf` property test exercises the
composite path. So this is **not** a current bug.

The fragility surfaces when:

- A future algorithm regression returns NaN on a corner-case input
  the per-algorithm property test missed.
- A consumer registers a custom `scoreFn` via the parameterised
  `With*Algorithm` closures (e.g. `WithSmithWatermanGotohAlgorithm`
  with pathological `SWGParams` — the godoc explicitly says
  "nonsense values produce a deterministic-but-meaningless score"
  but doesn't promise finiteness).

If any `entry.scoreFn` returns NaN, `acc + (entry.weight * NaN) ==
NaN` and NaN poisons every subsequent iteration and the return
value. Match would then always be `false`, and `ScoreAll` would
return a map containing a NaN value. Both behaviours are
deterministic (NaN propagation is IEEE-754-mandated and identical
across platforms), so this does NOT break cross-platform
byte-identity — but it does silently break the documented `[0, 1]`
contract.

**Impact:** No cross-platform divergence risk. Latent NaN/Inf leak
risk if any algorithm's per-pair contract is ever violated.

**Fix:** Two viable options, in increasing strength:

1. *(Lightweight)* Add a panic guard inside the reduction loop —
   `if math.IsNaN(score) || math.IsInf(score, 0) { panic(...) }`.
   Catches programmer bugs in the algorithm layer at the earliest
   possible point. Cost: ~1 ns per iteration.
2. *(Defence in depth)* Same as option 1 but call a small helper
   that wraps the panic in an error-returning path via a new
   `Scorer.ScoreErr` method — would require API-ergonomics review
   and is out of scope for v0.x.

Recommendation: defer to v0.y unless a NaN-producing algorithm is
ever introduced. Track via a tightly-scoped issue.

### DET-03: WithoutAlgorithm's in-place compaction aliases the underlying entries array
**Severity:** MEDIUM
**File:** scorer_options.go:184-199
**Issue:** `WithoutAlgorithm` uses the in-place compaction idiom
`filtered := cfg.entries[:0]`. This reuses the backing array of
`cfg.entries`, walking forwards. Because every iteration writes to
`filtered[k]` only after reading `cfg.entries[i]` with `i >= k`, the
write does not clobber any value that the iteration still needs to
read — so the algorithm is *correct*.

The fragility is that **the resulting `cfg.entries` slice retains the
old backing array's spare capacity**. If a subsequent option appends
new entries, those appends reuse the slots beyond the new length
(harmless), but **any code that ever exposed the underlying array
externally** (it currently doesn't) would see the *post-removal*
values in those spare slots, not the original ones. This is a
latent footgun if a future maintainer ever returns `cfg.entries` or
shares a slice header with downstream code.

For determinism specifically: the option-application order semantics
(WithAlgorithm + WithoutAlgorithm + WithAlgorithm again) remain
deterministic — `cfg.entries` is a value-typed slice header on the
config struct, never escapes, and the dedup step in `NewScorer`
collapses everything through the `seen` map.

**Impact:** No current determinism bug. Latent fragility if the
config slice is ever externalised.

**Fix:** Either (a) leave it (the in-place compaction is the
canonical Go idiom and well-understood), or (b) for extra hygiene,
allocate a fresh slice: `filtered := make([]scorerEntry, 0,
len(cfg.entries))`. The allocation cost is negligible (called once
per WithoutAlgorithm at construction time). I lean toward (a) — the
godoc comment in `WithoutAlgorithm` already documents the linear-
scan-and-compact semantics; the fragility is theoretical.

Recommendation: no action; document as accepted at the call site if
ever revisited.

### DET-04: scorer-default.json carries a "two short Smith/Schmidt entries differ only by composition" risk of golden-file drift
**Severity:** LOW
**File:** testdata/golden/scorer-default.json:172-185, 232-245, 282-292
**Issue:** Three of the 22 entries are `Smith` / `Schmidt` pairs
under three different Scorer compositions (`DefaultScorer`,
`DefaultScorer-MinusDoubleMetaphone`,
`Raw-Weights-Lev-1-JW-3-NoNorm`). Their `scorer_config` field
disambiguates them, but a developer skimming the JSON in a
mismatch-log excerpt could mistake one for another. The golden
*file* is deterministic — that's not the issue. The issue is the
**human ergonomics** of cross-platform CI failure triage: if a
divergence appears on, say, linux/arm64, the failing-bytes diff in
the CI log might land between two near-identical Smith/Schmidt rows
and the reviewer might pin the wrong scorer_config when reading.

**Impact:** Zero impact on determinism correctness. Slows
human triage on a hypothetical cross-platform mismatch.

**Fix:** No action required. If a follow-up ever reworks the golden
corpus, consider grouping by `scorer_config` (insertion-order sort)
or interleaving so adjacent rows differ on the input pair as well
as the composition. Strictly nit-level.

### DET-05: FMA-fusion remediation pattern is documented but not exercised
**Severity:** LOW
**File:** scorer.go:55-61 (godoc) and scorer.go:364-368 (in-body comment)
**Issue:** The header godoc and the inline comment both document the
remediation pattern for FMA fusion on arm64 — if matrix divergence
ever appears, insert an explicit `float64()` round-trip to defeat
the compiler's auto-FMA — but this pattern is **not actually
applied** anywhere in the Score reduction loop. The empirical
observation is that `score * weight` products in `[0, 1]` stay below
the ULP threshold of the cross-platform golden gate.

This is acceptable today (the Phase 5 Cosine path has the same
unmitigated pattern and the algorithms.json golden gate has passed
on the matrix for 23 algorithms). But it is a **conditional
guarantee**: the moment a future algorithm produces a score whose
weighted contribution lands close to a representable-float boundary,
the unmitigated `(weight * score)` pattern *could* diverge across
arm64 (FMA-emitting) and amd64 (non-FMA).

**Impact:** Zero current divergence risk. Latent risk against a
future composition where per-algorithm score × weight magnitudes
cross the ULP boundary.

**Fix:** No action while the matrix passes. If `scorer-default.json`
ever fails the matrix gate on a single platform, the documented
remediation is:

```go
// Replace: acc = acc + (entry.weight * score)
// With:    acc = acc + float64(float64(entry.weight) * float64(score))
```

The intermediate `float64()` cast forces a rounded intermediate
result, defeating FMA fusion. Track as a documented contingency,
not a present bug.

### DET-06: ScoreAll documented as "iteration order not deterministic" — verify the doc actually says this
**Severity:** LOW
**File:** scorer.go:471-472
**Issue:** Verified — the godoc at scorer.go:472 says:
"Map iteration order is non-deterministic per Go map semantics. Map
CONTENTS are deterministic byte-for-byte … Consumers requiring
stable iteration order MUST sort the keys themselves — typically via
fuzzymatch.AlgoIDs() then key-lookup." This matches
`.claude/skills/determinism-standards/SKILL.md` §"What's NOT
Deterministic (and that's fine)".

I checked the example/test code to confirm no example consumer
emits the map values in a `for k, v := range m` loop and then
serialises the output to stdout / a log / a file. Two call sites
exist:

- `scorer_golden_test.go:107` (`scoreAllAsStringKeys`) — copies into
  another map, which is then JSON-marshalled (alphabetical sort by
  `encoding/json`, so output-ordering deterministic).
- `examples/identifier-similarity/main.go` and
  `examples/scorer-composition/main.go` — not yet read; flag as a
  follow-up to confirm they sort before emitting.

**Impact:** None on the library itself; only on consumer
documentation hygiene.

**Fix:** Spot-check the two example programs and any future
`docs/scorer.md` snippets to confirm they sort keys before iterating
for output. If any of them range over the map directly and print,
patch the example.

### DET-07: scorer_signature is a static literal in scorer_golden_test.go
**Severity:** LOW
**File:** scorer_golden_test.go:330
**Issue:** `scorer_signature: "DefaultScorer-2026-05-16"` is a
hard-coded string with the construction date baked in. This is
intentional per 08-REVIEW.md / 08-CONTEXT.md (the alternative — a
runtime-computed signature — would either embed a generated_at
timestamp or hash the dispatch table; both introduce non-determinism
risks the current static-literal form avoids).

The fragility is that **any future change to DefaultScorer's
composition** (the 6 algorithms + 0.85 threshold) requires hand-
updating this string as part of the same PR. Forgetting to update
it does not break determinism — the byte-identity check still
passes if the string is unchanged — but it does mean the golden
file's metadata becomes a lie about which DefaultScorer composition
it represents.

**Impact:** None on cross-platform byte-identity. Drift risk on
human-readable metadata.

**Fix:** No code change. Recommend adding a release-prep checklist
item in `docs/scorer.md` or the release runbook: "If DefaultScorer
composition changes, bump scorer_signature in
scorer_golden_test.go." Strictly process-level.

---

## Cross-Platform CI Prerequisite

The Phase 8 byte-identity claim has been verified locally on
**darwin/arm64 only**. The full five-platform matrix (linux/amd64,
linux/arm64, darwin/amd64, darwin/arm64, windows/amd64) must run
`go test -run TestGolden_ScorerDefault ./...` and produce exit code 0
on every platform before the byte-identity guarantee for
`scorer-default.json` can be considered empirically verified.

This is not a code defect — it is an environment / CI configuration
matter outside the scope of the source-code determinism review. The
review unblocks Phase 8 from the determinism perspective on the
assumption that the matrix CI will be exercised before tagging
v0.x with Phase 8 included. If the matrix has not yet been wired in
for the Scorer (the algorithms.json gate already runs on the matrix
per Phase 1+ infrastructure), the determinism-reviewer
recommendation is to confirm the Scorer golden test runs through the
same `make verify-determinism` invocation that the algorithms golden
goes through. Per scorer_golden_test.go's `assertGolden` plumbing,
this should be automatic (the helper is the same).

---

## GO / NO-GO

**GO with follow-ups.** Phase 8's composite Scorer carries the LOCKED
determinism contract from Phase 5 Cosine forward correctly. The
reduction loop, weight normalisation, AlgoID-sorted iteration, no-init
discipline, no-map-iteration discipline, and golden-file schema are
all in order. No BLOCKING findings on cross-platform byte-identity.

The three MEDIUM findings (DET-01 NaN guard, DET-02 NaN propagation
defence, DET-03 in-place compaction) are quality / robustness issues,
not byte-identity risks. DET-01 overlaps with CR-01 in 08-REVIEW.md
and should be fixed there (the determinism angle is recorded here
for completeness). DET-02 and DET-03 can be deferred to a follow-up
issue.

The four LOW findings (DET-04 golden-file ergonomics, DET-05 FMA
remediation not exercised, DET-06 example-consumer check, DET-07
signature drift) are defence-in-depth / process improvements.

**Prerequisite for merging the v0.x Phase 8 release tag:** the
cross-platform CI matrix must exercise `TestGolden_ScorerDefault` on
all five platforms (linux/{amd64,arm64}, darwin/{amd64,arm64},
windows/amd64) and pass byte-identically. The source code is in
shape for this; only the CI run needs to occur.
