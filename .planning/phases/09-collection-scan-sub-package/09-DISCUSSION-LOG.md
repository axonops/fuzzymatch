# Phase 9: Collection Scan Sub-package - Discussion Log

> **Audit trail only.** Do not use as input to planning, research, or execution agents.
> Decisions are captured in CONTEXT.md — this log preserves the alternatives considered.

**Date:** 2026-05-19
**Phase:** 9-Collection-Scan-Sub-package
**Areas discussed:** Warning type alignment with Phase 8, Config validation strictness (Q2 framework), Sort-key completeness mechanism, Validate / scan integration

---

## Warning type alignment with Phase 8

### Sub-question 1a — Warning.Scores key type

| Option | Description | Selected |
|--------|-------------|----------|
| `map[AlgoID]float64` (align) | Match Phase 8 ScoreAll. Consumers get compile-time type safety. Spec §12.1 line 1337 amended in lockstep. Recommended. | ✓ |
| `map[string]float64` (per spec) | Keep the original spec shape. Forces consumers to re-encode AlgoID→string namespace. Drifts from the rest of the v1.0 API. | |
| Omit Scores entirely from Warning | Only expose Score (the composite). Consumers wanting per-algorithm breakdown call Scorer.ScoreAll themselves. Slimmer Warning, loses the 'free' breakdown. | |

**User's choice:** `map[AlgoID]float64` (align). Recorded as D-01 in CONTEXT.md §1.
**Notes:** Direct application of the Phase 8 SPEC OVERRIDE pattern. The same rationale that drove `ScoreAll` to typed enum keys applies here verbatim.

### Sub-question 1b — Public-type name for scan's within/cross discriminator

| Option | Description | Selected |
|--------|-------------|----------|
| `scan.Kind` (shorten) | Package-scoped, unambiguous, no clash with root's `fuzzymatch.WarnKind`. Recommended. | ✓ |
| `scan.WarningKind` (per spec) | Keep the spec name. Verbose; the `WarnKind` / `WarningKind` asymmetry persists. | |
| `scan.WarnKind` (symmetric with root) | Match root's `fuzzymatch.WarnKind` naming. The two enums are unrelated; symmetry would mislead. | |

**User's choice:** `scan.Kind`. Recorded as D-02 in CONTEXT.md §1.
**Notes:** Preview-block AskUserQuestion misfired the first time — see user-memory `feedback-askuserquestion-previews`. The decision itself was unambiguous; the UI was not.

---

## Config validation strictness (Q2 framework)

### Sub-question 2a — Empty Item.Name handling

| Option | Description | Selected |
|--------|-------------|----------|
| Fail-fast on first invalid item | Return ErrInvalidItem with the first offending index. Cheap precheck. Caller fixes one at a time. Matches spec phrasing. | |
| Collect all invalid indices, return joined error | Walk full slice; return all empty-Name indices in one wrapped error (`errors.Join`). Caller fixes them all at once. | ✓ |
| Lenient — skip empty-Name items silently | No error returned; drops items with empty Name. Diverges from spec. | |

**User's choice:** Collect all invalid indices via `errors.Join`. Recorded as D-03 in CONTEXT.md §2.
**Notes:** Spec §12.1 line 1389 phrasing ("the offending item's index") gets amended to "every offending index, joined via errors.Join" in the planner's first commit. This `errors.Join` pattern recurs in D-05 and D-06 — uniform across all three input-validation paths.

### Sub-question 2b — CrossGroupThresholdBoost default + validation

| Option | Description | Selected |
|--------|-------------|----------|
| Zero=0.0 means 'no boost'; spec default 0.05 lives in DefaultConfig() helper | Mirrors Scorer.DefaultScorer pattern. Strict validation: NaN/±Inf/<0/>1 → ErrInvalidConfig. Recommended. | ✓ |
| Zero=0.0 means 'use spec default 0.05' (magic zero) | Silently substitutes 0.05. Looks ergonomic but hides behaviour. Cannot explicitly set boost=0.0. | |
| Lenient — clamp out-of-range silently | NaN/Inf → 0.0; <0 clamped to 0; >1 clamped to 1. Diverges from Phase 8.5 Q2. | |

**User's choice:** Zero=0.0 + `scan.DefaultConfig` helper holds 0.05. Recorded as D-04 in CONTEXT.md §2.
**Notes:** Spec §12.1 line 1359 default-0.05 statement migrates to `DefaultConfig` godoc.

### Sub-question 2c — SuppressedPairs validation

| Option | Description | Selected |
|--------|-------------|----------|
| Strict: empty strings → ErrInvalidConfig; self-pairs silently kept | Either side empty → ErrInvalidConfig wrapped with index. Self-pairs harmless (Check skips identical items). Recommended. | ✓ (with errors.Join modifier) |
| Strict on both — empty AND self-pairs error | Self-pair is a consumer typo. Strongest hygiene. | |
| Lenient — silently drop empty / self-pairs | Filter invalid entries silently. Easier on consumers, masks typos. | |

**User's choice:** Strict on empty strings (silently kept on self-pairs) WITH the `errors.Join` modifier from 2a applied — collect all invalid pair indices, return joined error. Recorded as D-05 in CONTEXT.md §2.
**Notes:** User initially attempted the AskUserQuestion selection without notes; second attempt added the errors.Join modifier. Same UI friction as 1b — see user-memory `feedback-askuserquestion-previews`. The errors.Join pattern is now consistent across D-03 (Items), D-05 (SuppressedPairs), and D-06 (duplicate Items).

---

## Sort-key completeness mechanism

| Option | Description | Selected |
|--------|-------------|----------|
| Reject duplicate Name+Group as ErrInvalidItem at validation | Detect during pre-flight; collect all duplicate-key indices via errors.Join. Sort key provably unique. In-line assertion becomes a hard invariant. | ✓ |
| Extend sort key with item indices (synthetic tiebreaker) | Sort by (Kind, NameA, NameB, GroupA, GroupB, IndexA, IndexB). Allows duplicates; exposes input order through the warning order. | |
| Dedupe — first-pair-wins | Keep the first, drop the rest. Loses information; masks consumer schema issue. | |
| Promote Tag into the sort key | Tag is `any` — cannot be ordered without reflection. Strongly NOT recommended. | |

**User's choice:** Reject duplicates at validation; collect all duplicate-pair indices via `errors.Join`. Recorded as D-06 in CONTEXT.md §3.
**Notes:** User rejected the AskUserQuestion tool entirely on this turn ("Your UI doesn't fucking work") — see user-memory `feedback-askuserquestion-previews`. Decision communicated in plain prose: option 1 + the errors.Join modifier. With duplicates rejected at the door, the SCAN-05 in-line completeness assertion becomes a `panic(fuzzymatch.ErrInternalInvariantViolated)` library-bug guard, not a fallback tiebreaker.

---

## Validate / scan integration

| Option | Description | Selected |
|--------|-------------|----------|
| Separate concerns | scan.Check does similarity only. Consumer calls fuzzymatch.Validate explicitly. Two clean APIs. Matches Phase 8.5 Q4. Recommended. | ✓ |
| Side-channel return | Check signature becomes ([]scan.Warning, []fuzzymatch.Warning, error). Couples two diagnostic systems. | |
| Validate-on-emit only | scan.Warning gains optional Diagnostics field. Bloats struct; runs Validate inside scoring loop. | |

**User's choice:** Separate concerns. Recorded as D-07 in CONTEXT.md §4.
**Notes:** Clean two-API boundary. Validate and Check have different cardinalities (Validate: O(N) per item; Check: O(N²) before bucketing) and folding them would couple unrelated systems. docs/scan.md quickstart will show the Validate-then-Check composition pattern explicitly.

---

## Claude's Discretion

The following items were not surfaced as gray areas — Claude / the planner makes the call within the locked invariants:

- **Item validation error format** — `fmt.Errorf("scan: invalid item at index %d: %w", idx, ErrInvalidItem)` (Go idiomatic) vs a dedicated `*InvalidItemError` struct with `.Index int`. Either acceptable as long as `errors.Is(err, scan.ErrInvalidItem)` works and the index is in the message.
- **Token-bucket internal data layout** — `map[string][]int` per spec §12.5 vs a different structure with the same complexity. `PropCheck_BucketEquivalentToNaive` is the load-bearing correctness gate.
- **`tokeniseAll` helper vs inline tokenisation** — either acceptable; each Item.Name tokenised at most once per Check invocation.
- **`examples/scan-demo/main.go` narrative shape** — must demonstrate (a) within-group default Scorer, (b) within + cross-group with DefaultConfig, (c) at least one of each suppression mode. Beyond that, docs-writer's call.
- **`scan.feature` / `suppression.feature` scenario count** — bdd-scenario-reviewer's call. Minimum: one scenario per D-03..D-07 decision, plus the Phase 8.5 R2 deferred suppression scenarios.

---

## Follow-up gray area locked without AskUserQuestion (post-area-4)

### D-08 — Bucket threshold: private constant vs Config field

After Area 4 closed, Claude raised the bucket-threshold knob as a candidate fifth area. User locked it inline rather than going through another AskUserQuestion turn:

**Decision:** `bucketThreshold = 50` is a private package constant in `scan/bucket.go`. NOT a `Config.BucketThreshold` field.

**User rationale (paraphrased; saved to user-memory `feedback-yagni-for-tuning-knobs`):**
1. YAGNI applies hard to optimisation knobs — consumers don't reason about threshold values, they reason about whether pairs are flagged.
2. API expansion is cheap, contraction is impossible — promoting to a Config field in v1.x is non-breaking; demoting would break consumers.
3. Phase 8.5 philosophy was tightening (Q3/Q4/Q5 all removed surface) — adding new knobs no consumer has demanded violates that direction.
4. Benchmarking doesn't need public exposure — maintainers verify the threshold in internal tests where the constant is accessible.

**User-added requirement:** Before v1.0 ships, Phase 9 must include `BenchmarkScanCheck_BucketVsNaive_GroupSize` sweeping group sizes 10 / 25 / 50 / 75 / 100 / 200. If the empirical crossover differs materially from 50, **update the constant** before the phase closes. The spec's stated 50 is a starting hypothesis, not a load-bearing value.

Recorded as D-08 in CONTEXT.md §5 with full action items for the planner.

---

## Deferred Ideas

(See CONTEXT.md `<deferred>` section for the full list with rationale.)

- `Config.BucketThreshold` public field — revisit post-v1.0 if tuning need surfaces; track via GitHub issue.
- Side-channel `Validate` integration (e.g. `scan.CheckWithDiagnostics`) — revisit if boilerplate signal emerges.
- Streaming `iter.Seq[Warning]` variant — V2-API-03, already tracked.
- Parallel goroutine-fanned scan — V2-API-04, already tracked.
- Dedicated `scan.Item` constructor — not raised; plain struct literal is fine for v1.0.

---

*Phase: 9-Collection-Scan-Sub-package*
*Discussion log written: 2026-05-19*
