# Phase 8: Composite Scorer - Discussion Log

> **Audit trail only.** Do not use as input to planning, research, or execution agents.
> Decisions are captured in CONTEXT.md — this log preserves the alternatives considered.

**Date:** 2026-05-16
**Phase:** 8-composite-scorer
**Areas discussed:** ScoreAll key + WithoutAlgorithm composition, Threshold default, Normalisation flow with token algorithms, Plan decomposition + scorer-default.json corpus

---

## Top-level: Which areas to discuss

| Option | Selected |
|--------|----------|
| ScoreAll key + WithoutAlgorithm composition | ✓ |
| Threshold default (no WithThreshold) | ✓ |
| Normalisation flow with token algorithms | ✓ |
| Plan decomposition + scorer-default.json corpus | ✓ |

**User's choice:** All four selected; others routed to Claude's discretion via planner.

---

## Area 1: ScoreAll key + WithoutAlgorithm composition

### Q1.1: ScoreAll(a, b) return type

| Option | Description | Selected |
|--------|-------------|----------|
| `map[string]float64` (snake_case) | Spec §8.3 wording; keys `"levenshtein"`, `"jaro_winkler"`, etc. | |
| `map[AlgoID]float64` (typed enum) | REQUIREMENTS.md SCORER-05 wording; typed keys; AlgoID.String() for display | ✓ |

**User's choice:** `map[AlgoID]float64`, typed enum keys. **OVERRIDES `docs/requirements.md` §8.3.**

**Notes:** User initially asked for clarification because they read a preview that appeared to pass `int` somewhere — confirmed that `AlgoID` IS the typed enum (`type AlgoID int` with exported typed constants), and that the only `int` access is the package-internal dispatch array index, invisible to consumers. User confirmed the typed-everywhere project discipline applies; api-ergonomics-reviewer will sign off the spec override at PR time.

### Q1.2: DefaultScorerOptions() shape — how to "default minus algorithm X"

| Option | Description | Selected |
|--------|-------------|----------|
| `DefaultScorerOptions() []ScorerOption` + new `WithoutAlgorithm(AlgoID)` | Composable via append; standard functional-options shape | ✓ |
| Just expose `DefaultScorerOptions()` — consumer slices manually | Brittle; smaller surface but exposes option-type internals | |
| Skip — only `DefaultScorer()`; consumer rebuilds from scratch | Smallest surface; rejects SCORER-02 | |

**User's choice:** Add `WithoutAlgorithm(AlgoID)` option.

**Notes:** User initially questioned the use case for "without algorithm". Explained: consumer wants defaults but one algorithm doesn't suit their data (numeric IDs → Double Metaphone is noise; Unicode-heavy domain → ASCII-only phonetics don't apply). Without the option, consumer manually re-specifies the other 5 algorithms with exact weights, and is forced to re-version when defaults change in v1.x. One-line option is cheap insurance against the most common "I want defaults but..." complaint. User agreed the use case is real.

---

## Area 2: Threshold default

| Option | Description | Selected |
|--------|-------------|----------|
| Mandatory WithThreshold — NewScorer errors without it | Returns new ErrMissingThreshold; DefaultScorer bakes 0.85 in | ✓ |
| Default 1.0 (only exact match passes) | Conservative fail mode; no construction-time error | |
| Default 0.0 (Match always true) | Loud immediate signal something is wrong | |
| Inherit DefaultScorer's 0.85 | Arbitrary for non-default compositions | |

**User's choice:** Mandatory `WithThreshold`; NewScorer errors without it via new `ErrMissingThreshold` sentinel.

**Notes:** Rationale: the default-threshold question has no safe answer. A library-wide default of 1.0 silently produces "no matches found"; 0.0 silently makes every comparison match; 0.85 is calibrated for the SPECIFIC 6-algorithm default mix, arbitrary elsewhere. Requiring `WithThreshold` forces an explicit calibration step at the construction site — the place where the consumer most has the context to choose.

---

## Area 3: Normalisation flow with token algorithms

| Option | Description | Selected |
|--------|-------------|----------|
| Pre-normalise once + zero-value opts to token algos | Single source of truth; predictable golden output | ✓ |
| Skip pre-normalise for token algos; forward Scorer opts | Two-pipeline; harder to reason about | |
| Pre-normalise once; ignore token algos' opts (idempotent re-normalise) | Simpler but wastes CPU on token paths | |

**User's choice:** Pre-normalise once at the Scorer boundary; pass `na, nb` to ALL algorithms.

**Notes:** Investigation surfaced an implementation nuance: `Tokenise` takes `TokeniseOptions` (NOT `NormalisationOptions` — separate types) and current token-based algorithms hardcode `Tokenise(s, DefaultTokeniseOptions())`. `MongeElkanScore`'s `opts NormalisationOptions` parameter exists but is unused (`_ = opts` on `monge_elkan.go:393`, comment says "accepted for Phase 8 Scorer compatibility"). Mapping confirmed with user (follow-up Q): Scorer pre-normalises once with its NormalisationOptions; pre-normalised na/nb passed to all algorithms; token-based algos keep their hardcoded `Tokenise(s, DefaultTokeniseOptions())` internals; ME's opts param stays vestigial (no changes to existing token-based files in Phase 8). User confirmed.

---

## Area 4: Plan decomposition + scorer-default.json corpus

### Q4.1: Plan decomposition shape

| Option | Description | Selected |
|--------|-------------|----------|
| 3 plans (errors+options+Score/Match; ScoreAll+accessors+DefaultScorer+property; finalisation) | Foundation in one plan; ergonomics review in one PR | |
| 4 plans (errors+options; NewScorer+Score+Match; ScoreAll+accessors+DefaultScorer+property+concurrent; finalisation) | Smaller PRs; cleaner separation between option-side and method-side | ✓ |
| 2 plans (everything-except-finalisation; finalisation) | Tight; Phase 7 plan 07-02 was a similar one-PR surface | |

**User's choice:** 4 plans.

**Notes:** Layout:
- 08-01: errors + ScorerOption type + all option fns + scorerConfig type
- 08-02: NewScorer validation + Score + Match + last-write-wins
- 08-03: ScoreAll + accessors + DefaultScorer + property + concurrent
- 08-04: golden + BDD + examples + docs + llms.txt

### Q4.2: scorer-default.json golden corpus

| Option | Description | Selected |
|--------|-------------|----------|
| Reuse identifier-similarity 14 rows + 8-12 Scorer-specific threshold-edge additions | Total ~22-26 entries | ✓ |
| Fresh 30-entry curated corpus | No identifier-similarity reuse; deliberate coverage | |

**User's choice:** Reuse identifier-similarity corpus + threshold-edge additions.

**Notes:** Each entry pinned with Score (composite), Match, and ScoreAll (per-algorithm breakdown using typed AlgoID string-keys per the JSON serialisation format). Mandatory coverage rows enumerated in CONTEXT.md §6: identity, both-empty, one-empty, just-above-threshold, just-below-threshold, Unicode pre-/post-NFC, phonetic-match-only-not-edit-similar, WithoutNormalisation, WithoutAlgorithm(AlgoDoubleMetaphone), custom single-algo Scorer, WithNormaliseWeights(false).

---

## Claude's Discretion

Routed to gsd-planner / gsd-phase-researcher / specialist agents (no user input required):

- Exact `scorerConfig` internal layout (slice-of-entries vs map-keyed-by-AlgoID)
- Closure-capture mechanism for parameterised algorithms (separate field vs stored closure)
- Exact godoc wording beyond the SPEC OVERRIDE notice and ErrMissingThreshold paragraph
- Exact entry count of scorer-default.json (22-26 range)
- Exact BDD scenario count (8-12 range)
- Whether plan 08-03's concurrent test uses `sync.WaitGroup` (must be stdlib-only) vs alternatives
- Exact prose structure of `docs/scorer.md` and `docs/tuning.md` (docs-writer + user-guide-reviewer authority)
- Whether `examples/scorer-composition/main.go` is single-file or split
- Whether to introduce `scorer_internal_test.go` for unexported invariants
- Allocation-budget recording mechanism in plan 08-04's bench fixture
- Sign-off recording wording for api-ergonomics-reviewer on plan 08-03's SPEC OVERRIDE

## Deferred Ideas

- Drop or wire ME's vestigial `opts NormalisationOptions` parameter — Phase 11 API freeze or v1.x decision
- Amend `docs/requirements.md` §8.3 to say `map[AlgoID]float64` — land in plan 08-04 docs commit
- Amend REQUIREMENTS.md SCORER-08 to replace `WithCustomNormalisation()` with `WithNormalisation(opts)` — land in plan 08-04 docs commit
- Pooled `transform.Transformer` for Normalise — v1.x perf revisit (already Phase 1 carry-forward)
- DefaultScorer composition v1.x revisit — calibrate via Phase 11 integration shakedown if needed
- Scorer-level allocation reuse via sync.Pool — v1.x perf opportunity
- `Scorer.MarshalJSON` / serialisable Scorer config — out of v1.0 scope
- Threshold-edge BDD scenarios on every option combination — v1.x test-writer opportunity
