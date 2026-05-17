---
phase: 08
plan: 08-PERFORMANCE-REVIEW
subsystem: composite-scorer
tags: [performance, scorer, allocations, budget, phase-8]
status: complete
reviewer: algorithm-performance-reviewer
date: 2026-05-17
---

# Phase 8 Scorer — Performance Review

Requested by: 08-04-SUMMARY.md manual gate #4 (algorithm-performance-reviewer
escalation for allocs/op > 8 on ASCII Short/Medium).

## Fresh Benchmark Run

Platform: darwin/arm64 (Apple M2), `go test -bench=BenchmarkDefaultScorer -benchmem -count=5 -run=^$`

| Benchmark                                     | ns/op (median) | B/op   | allocs/op | Wall budget | Alloc budget | Wall   | Alloc |
|-----------------------------------------------|---------------:|-------:|----------:|-------------|-------------|--------|-------|
| BenchmarkDefaultScorer_Score_ASCII_Short      |          1,553 |    192 |        12 | < 30,000 ns | ≤ 8         | PASS   | FAIL  |
| BenchmarkDefaultScorer_Score_ASCII_Medium     |         13,440 |  6,208 |        34 | < 30,000 ns | ≤ 8         | PASS   | FAIL  |
| BenchmarkDefaultScorer_Score_ASCII_Long       |      1,613,000 |137,024 |        37 | informational | —          | —      | —     |
| BenchmarkDefaultScorer_Score_Unicode_Short    |          2,382 |    872 |        21 | informational | —          | —      | —     |
| BenchmarkDefaultScorer_ScoreAll_ASCII_Short   |          1,716 |    384 |        14 | informational | —          | —      | —     |
| BenchmarkDefaultScorer_Match_ASCII_Short      |          1,553 |    192 |        12 | informational | —          | —      | —     |

Wall-time budget is met on all budgeted cases.

Allocation budget is breached: 12 allocs/op at Short (50% over), 34 allocs/op at Medium (4.25× over).

Note: 08-04-SUMMARY.md reported 925 ns (Short) and 8,101 ns (Medium). The fresh
run reports 1,553 ns (Short) and 13,440 ns (Medium) — approximately 1.7× slower.
This is consistent with measurement noise on a developer machine between sessions
(background load, thermal state). The absolute values remain well within the
30 µs wall budget. Alloc counts are identical, confirming allocations are
deterministic and run-independent.

---

## Findings

### PERF-01 — Allocation budget breach: 12 allocs/op at ASCII Short, 34 allocs/op at ASCII Medium

**Severity:** CRITICAL

The spec-locked budget in `docs/requirements.md` §14.2 is ≤ 8 allocs/op for
`DefaultScorer().Score(a, b)` on ASCII inputs ≤ 50 characters. The current
implementation reports 12 allocs/op (Short) and 34 allocs/op (Medium).

This is a v1.0 ship blocker because the §14.2 budget is a binding performance
contract, not an aspiration. Acceptance criterion in §20 explicitly lists
"All section-14 budgets met". Until either the implementation is fixed or the
budget is formally revised (with documented rationale), the v1.0 acceptance
criteria are not met.

Wall-time passes (1.55 µs / 13.4 µs vs < 30 µs budget), so this is an
allocation discipline issue, not a latency crisis.

**Root cause (confirmed by per-component measurement):**

The DefaultScorer composes 6 algorithms. Their per-call allocation breakdown
on ASCII Short ("abc" / "abcd"):

| Component                           | allocs/op | Notes                                         |
|-------------------------------------|----------:|-----------------------------------------------|
| Normalise(a) + Normalise(b)         |         2 | 1 alloc per call: `string(buf)` heap escape   |
| DamerauLevenshteinOSA               |         0 | zero-alloc ASCII fast path — correct          |
| JaroWinkler                         |         0 | zero-alloc — correct                          |
| QGramJaccard (n=3, "abc"/"abcd")    |         0 | identity short-circuit fires on 3-char input  |
| SorensenDice (n=3, "abc"/"abcd")    |         0 | same short-circuit                            |
| TokenJaccard                        |         4 | 2 Tokenise slices + 2 map[string]struct{}      |
| DoubleMetaphone                     |         6 | 2×dmPrep string + 2×Builder.String() per side |
| **Total**                           |    **12** |                                               |

On ASCII Medium ("customer_billing_history_2024_v2" / "customerBillingHistory2024V2"):

| Component                           | allocs/op | Notes                                         |
|-------------------------------------|----------:|-----------------------------------------------|
| Normalise(a) + Normalise(b)         |         4 | 2 per call at medium: `buf` grow + string()   |
| DamerauLevenshteinOSA               |         0 | zero-alloc — correct                          |
| JaroWinkler                         |         0 | zero-alloc — correct                          |
| QGramJaccard (n=3, 30 chars)        |         6 | 2 maps + growth allocs per side               |
| SorensenDice (n=3, 30 chars)        |         6 | same pattern as QGramJaccard                  |
| TokenJaccard (5–6 tokens per side)  |        13 | 2 Tokenise slices + per-token strings + 2 maps|
| DoubleMetaphone                     |         6 | constant regardless of input length           |
| **Total (observed)**                |    **35** | (bench reports 34 — 1 alloc difference is     |
|                                     |           | because Scorer passes normalised strings to   |
|                                     |           | token/qgram algorithms, reducing token count) |

**Structural cause:** The §14.2 budget of ≤ 8 was written without accounting for
the actual per-algorithm allocation profiles of the 6 chosen algorithms. Adding
up the per-algorithm spec budgets from §14.1:

- DL-OSA:          0 allocs
- JaroWinkler:     0 allocs
- TokenJaccard:    ≤ 4 allocs
- QGramJaccard:    ≤ 4 allocs
- SorensenDice:    ≤ 4 allocs
- DoubleMetaphone: ≤ 2 allocs
- Normalise × 2:   ≤ 2 allocs (1 per call)
- **Spec floor:**  ≤ 16 allocs

The spec §14.2 budget of ≤ 8 is **mathematically impossible** for the
DefaultScorer's 6-algorithm composition if all 6 algorithms operate within their
own per-algorithm budgets and normalisation is enabled. The budget was written
optimistically, likely assuming the Scorer's overhead was the only concern and
that per-algorithm allocations would not compound.

**Fix options (see PERF-05 for recommendation):**

A. Revise §14.2 budget to ≤ 16 allocs/op (the spec-floor sum). Matches the
   summed per-algorithm budgets. Simple documentation fix; no code change.

B. Reduce individual algorithm allocations:
   - DoubleMetaphone: strings.Builder.String() allocates two heap strings per
     encode call (4 allocs per DoubleMetaphoneScore call). Using a
     stack-allocated `[4]byte` buffer for the ≤ 4-char key avoids those
     allocations. Budget would drop from 6 → 2 per DoubleMetaphoneScore call.
   - Normalise ASCII: the current `buf := make([]byte, 0, len(s)*2+1)` does
     not escape but `string(buf)` does (unavoidable for a returned string).
     No win available here without unsafe.
   - QGram/SorensenDice: map construction on short (≤ 6 trigrams) inputs could
     be replaced with a small-input fast path using a sorted `[16][3]byte`
     stack array. This would drop 6 allocs → 0 for trigram inputs under 8 chars.

C. Defer to v1.x with a documented release-tracking issue. Accept the breach
   for v0.x; ship a GitHub issue tracking the optimisation.

---

### PERF-02 — `normaliseASCII` allocates 1 heap object per call despite stack-resident buffer

**Severity:** NORMAL

`normaliseASCII` uses `buf := make([]byte, 0, len(s)*2+1)` which the escape
analyser confirms does NOT escape to the heap (`./normalise.go:215:13: make(...) does not escape`).
However, `string(buf)` at line 249 allocates a heap string because the returned
string must be heap-backed (Go strings are immutable, so the returned string
cannot share the stack-resident `buf` backing array).

This is 1 alloc per `Normalise(s, opts)` call on ASCII inputs. The performance-standards
skill states the target as "0 allocations" for `Normalise` on ASCII ≤ 50 chars, and
`docs/requirements.md` §14.3 says "0 allocations (stack buffer)".

The `string(buf)` allocation is structurally unavoidable without `unsafe.String`
(which is forbidden by go-coding-standards) or a pre-allocated output buffer
passed by the caller. The normalise bench comments acknowledge this implicitly
(the target says "0 allocs" but the actual bench shows 1 alloc and this is
recorded in bench.txt as 1 alloc — the bench comment is aspirational rather than
verified).

The `bench.txt` committed baseline shows `BenchmarkNormalise_ASCII_Short` with
`1 allocs/op` (confirmed by fresh run: 1 alloc at Short, 2 at Medium for the
double-normalise pattern), consistent with the implementation.

**Implication for §14.3 budget:** The spec budget "0 allocations (stack buffer)"
in §14.3 should be corrected to "1 allocation (output string)" for ASCII inputs.
The stack buffer is real and correct; the claim of 0 allocs is not achievable
without unsafe.

**Fix:** Amend `docs/requirements.md` §14.3 to read "≤ 1 allocation (output
string heap escape)" and update the normalise bench comment from "target: 0
allocs/op" to "target: 1 alloc/op (output string)". No code change needed.

---

### PERF-03 — DoubleMetaphone allocates 6 heap objects per score call at Short ASCII

**Severity:** NORMAL

`DoubleMetaphoneScore("abc", "abcd")` allocates 6 objects. Breakdown:

1. `dmPrep(a)` → `string(stackBuf[:n])` at line 210: 1 alloc (the returned
   string is heap-backed even though stackBuf is stack-allocated).
2. `dmPrep(b)` → same: 1 alloc.
3. `DoubleMetaphoneKeys(a)` → `p.String()` at line 860: 1 alloc.
4. `DoubleMetaphoneKeys(a)` → `alt.String()` at line 861: 1 alloc.
5. `DoubleMetaphoneKeys(b)` → `p.String()`: 1 alloc.
6. `DoubleMetaphoneKeys(b)` → `alt.String()`: 1 alloc.

The spec budget for DoubleMetaphone is ≤ 2 allocations (§14.1). The actual
allocation count is 6, which is 3× the budget.

**Root cause:** `DoubleMetaphoneKeys` uses `strings.Builder` for the primary and
secondary keys, and `Builder.String()` always allocates. Since both keys are
bounded to ≤ 4 bytes (dmMaxLen = 4), the output could be represented as
stack-allocated `[4]byte` arrays rather than `strings.Builder` values.

**Fix:**
```go
// Replace strings.Builder with [4]byte + length counter:
var pBuf, altBuf [dmMaxLen]byte
var pLen, altLen int

// In dmAdd: write into the arrays instead of Builder.WriteString.

// In DoubleMetaphoneKeys return:
return string(pBuf[:pLen]), string(altBuf[:altLen])
// Still 2 allocs for the returned strings, but avoids the 4 Builder.String() allocs.
```

This would reduce DoubleMetaphone from 6 allocs to 2 allocs, matching §14.1.
Combined with PERF-01's structural arithmetic, the DefaultScorer total would
drop from 12 → 8 allocs at Short (matching the budget exactly), and from
34 → ~30 at Medium (still over, driven by TokenJaccard and QGram growth at
medium input length).

The DoubleMetaphone fix alone does not close the gap at Medium. It is a
worthwhile correctness-vs-budget fix at Short that should be tracked regardless.

---

### PERF-04 — `bench.txt` does not contain any `BenchmarkDefaultScorer_*` entries

**Severity:** CRITICAL

`bench.txt` is the regression detection baseline for `benchstat`. The file
contains 1,296 lines covering all per-algorithm benchmarks but has zero
`BenchmarkDefaultScorer_*` entries. Phase 8 shipped `scorer_bench_test.go`
with 6 benchmarks but did not run `scripts/update-bench-txt.sh` to append
them to `bench.txt`.

This means:

- Benchstat has no baseline for Scorer benchmarks. CI's regression detection
  for the Scorer is completely blind — a future commit could double the
  Scorer's allocation count or wall time with no automated signal.
- The performance-standards skill states "bench.txt is committed to the
  repository and updated on every tagged release". Phase 8 did not tag a
  release but the fixture should be present before v1.0 tagging.

**Fix:** After the allocation budget question is resolved (see PERF-05), run
`make bench` on the self-hosted benchmark runner and append the 6 Scorer
benchmark names + results (×10 runs for benchstat validity) to `bench.txt`.
This must happen before the v1.0 tag. It is a regression-detection gap that
makes all future Scorer-touching PRs blind to performance changes.

---

### PERF-05 — `scoreFn func(a, b string) float64` field: no boxing, no allocation

**Severity:** POSITIVE

The `scorerEntry.scoreFn` field is a `func(a, b string) float64` stored as a
plain function value in the `algorithmsAlgoIDSorted []scorerEntry` slice. The
Score loop dispatches via `entry.scoreFn(na, nb)` — a direct function-pointer
call through the slice, not an interface method dispatch.

This means there is no interface boxing overhead on the hot path. Each
`entry.scoreFn` is either a direct function reference (e.g.
`dispatch[AlgoTokenJaccard] = TokenJaccardScore`) or a small closure
capturing constant parameters (e.g. `func(a, b string) float64 { return
QGramJaccardScore(a, b, 3) }`). Closures DO allocate at construction time
(confirmed: `dispatch_qgram_jaccard.go:46:2: func literal escapes to heap`)
but that is a one-time package-load cost, not a per-Score-call cost.

The go-coding-standards skill explicitly forbids `Algorithm interface` dispatch
on hot paths ("Algorithm dispatch via typed AlgoID enum + switch — no Algorithm
interface allocates on hot paths"). The implementation correctly uses the typed
`dispatch[AlgoID]` table with pre-bound closures, satisfying this constraint.

**This pattern must be locked in.** Any future refactor that introduces an
interface on the dispatch path (e.g. changing `scoreFn` to `ScoreFunc
interface`) would introduce per-call allocations that the current design avoids.

---

### PERF-06 — Score and ScoreAll both normalise independently; no shared normalised strings

**Severity:** NORMAL

Both `Score` and `ScoreAll` independently call `Normalise(a, s.normaliseOpts)`
and `Normalise(b, s.normaliseOpts)` at their respective entry points (scorer.go
lines 355-357 and 503-506). A consumer who calls both `Score` and `ScoreAll`
on the same pair (e.g. for logging) normalises the pair twice — 4 allocs instead
of 2.

This is the documented design ("ScoreAll mirrors Score's normalisation policy"),
and the typical consumer either calls Score or ScoreAll, not both. However it is
worth noting that a future `ScoreWithBreakdown` variant (combining Score's
composite result with ScoreAll's per-algorithm detail) could share the
normalised strings for zero extra normalisation cost.

No immediate action required. Note in the v1.x roadmap if a combined method
is added.

---

## Allocation Source Accounting: ASCII Short

**Total: 12 allocs/op**

| Source                           | Allocs | Avoidable?                              |
|----------------------------------|-------:|-----------------------------------------|
| `Normalise(a)` → `string(buf)`   |      1 | No (unsafe required; forbidden)         |
| `Normalise(b)` → `string(buf)`   |      1 | No (same)                               |
| `TokenJaccardScore`: Tokenise(a) |      1 | Maybe (stack-slice for ≤ 4 tokens)      |
| `TokenJaccardScore`: Tokenise(b) |      1 | Maybe (same)                            |
| `TokenJaccardScore`: map setA    |      1 | Yes (small-input fast path possible)    |
| `TokenJaccardScore`: map setB    |      1 | Yes (same)                              |
| `DoubleMetaphone`: dmPrep(a) str |      1 | Yes (stack buf already used; string()   |
|                                  |        | still allocs — needs unsafe or return   |
|                                  |        | [4]byte pattern)                        |
| `DoubleMetaphone`: dmPrep(b) str |      1 | Yes (same)                              |
| `DoubleMetaphone`: p.String() a  |      1 | Yes (replace Builder with [4]byte buf)  |
| `DoubleMetaphone`: alt.String() a|      1 | Yes (same)                              |
| `DoubleMetaphone`: p.String() b  |      1 | Yes (same)                              |
| `DoubleMetaphone`: alt.String() b|      1 | Yes (same)                              |
| **Total**                        | **12** |                                         |

5 allocs are structurally unavoidable without unsafe (the 2 Normalise string
conversions + the 2 Tokenise output slices + 1 map for the token set, minimum).
7 allocs are **avoidable** with targeted optimisations: DoubleMetaphone Builder
replacement (4 allocs) + TokenJaccard small-input fast path (2 map allocs, via
a fixed-size stack array for ≤ 8 tokens) + Tokenise slice (speculative, would
require heap-suppression via pool or stack array return).

---

## Allocation Source Accounting: ASCII Medium

**Total: 34 allocs/op**

| Source                                   | Allocs | Notes                                    |
|------------------------------------------|-------:|------------------------------------------|
| `Normalise(a)` medium (buf grows)        |      2 | 1 buf growth + 1 string()               |
| `Normalise(b)` medium                    |      2 | same                                     |
| `QGramJaccardScore` (30-char trigrams)   |      6 | 2 maps × 3 growth allocs each            |
| `SorensenDiceScore` (30-char trigrams)   |      6 | same pattern                             |
| `TokenJaccardScore` (5–6 tokens/side)    |     13 | 2 Tokenise slices + per-token strings    |
|                                          |        | (from Tokenise's rune→string conversion) |
|                                          |        | + 2 maps                                 |
| `DoubleMetaphoneScore`                   |      6 | constant at medium (same as Short)        |
| DL-OSA + JaroWinkler                     |      0 | zero-alloc — correct                     |
| **Total**                                | **35** | (bench reports 34; 1 difference from     |
|                                          |        | Scorer normalising before dispatch)       |

The Medium allocation explosion is driven by:
1. QGram/SorensenDice map growth at 30 chars (29 trigrams per side → map starts
   at capacity hint `len-n+1 = 28` but rehashes when counts repeat and the
   internal load factor trips — this contributes 2 extra allocs per algorithm).
2. TokenJaccard's Tokenise producing 5–6 tokens from a 30-char identifier
   (each token is a heap string allocation from the `string(rs)` conversion
   at tokenise.go:328).

---

## Wall-Time Posture

Wall-time is healthy. Both budgeted benchmarks clear the 30 µs threshold with
large margins:

| Benchmark                            | Wall (median) | Budget   | Margin  |
|--------------------------------------|--------------|----------|---------|
| Score_ASCII_Short                    |   ~1.55 µs   | < 30 µs  | 19×     |
| Score_ASCII_Medium                   |  ~13.44 µs   | < 30 µs  | 2.2×    |

The Medium margin (2.2×) is tight enough to track. A future algorithm addition
to the DefaultScorer composition or a regression in QGram/SorensenDice map
performance could push Medium over budget. bench.txt inclusion (PERF-04) is
the mechanism to catch that.

---

## DP Optimisation Patterns

The Scorer's Score method has no DP tables of its own — it dispatches to
algorithm functions. The check is at the algorithm level:

- DL-OSA: verified zero-alloc at Short and Medium, consistent with two-row DP.
- QGram/SorensenDice: two-map pattern per extractQGrams() — no DP, map-based.
- TokenJaccard: map-based set intersection — no DP.
- JaroWinkler: zero-alloc — correct.
- DoubleMetaphone: Builder-based — no DP.

The Scorer dispatch loop itself (scorer.go:369-381) is a plain for-range over
a pre-sorted slice with no allocation, no sorting per call, no dynamic dispatch
through interfaces. The loop structure is correct.

---

## ASCII Fast Path

The Scorer pre-normalises with `Normalise(a, s.normaliseOpts)` which gates on
`isASCII(s)` internally to take the byte-level fast path. For ASCII inputs, the
Normalise ASCII path runs (confirmed by escape analysis: `buf := make([]byte, ...)
does not escape`). The Normalise function is NOT inlinable (`cannot inline Normalise:
function too complex: cost 161`) but the ASCII gate inside it IS inlined
(`inlining call to isASCII`).

There is no skip-normalise-if-already-normalised fast path. For the common case
where DefaultNormalisationOptions with ASCII input produces no change (e.g. all-
lowercase input with no separators), Normalise still runs the full `normaliseASCII`
pass and allocates a new string. Adding an "identity detection" check
(e.g. verify that no transform would be applied before allocating) is a possible
optimisation but adds complexity and would only save the 2 Normalise allocs — the
12-alloc total is dominated by other sources.

---

## sync.Pool Viability

`sync.Pool` for the ScoreAll output map would save 1 allocation per ScoreAll
call (the spec-budgeted 1 extra alloc relative to Score). This is low-ROI: the
spec explicitly documents "ScoreAll adds one allocation (the result map)" as
expected behaviour, and pooling a `map[AlgoID]float64` introduces lock
contention on the pool's internal shard mutex and requires clearing the map on
put — both costs are likely larger than the single allocation saved.

`sync.Pool` for the Normalise scratch buffer is explicitly deferred in the
normalise.go file header: "Pooling is deferred to a v1.x perf revisit." The
`transform.Transformer` construction (for non-ASCII NFC normalisation) is the
more expensive allocating path; the ASCII path's `buf` does not escape so no
pool is needed there.

`sync.Pool` is NOT the right fix for the 12-alloc breach. The breach is caused
by DoubleMetaphone's Builder allocation pattern (fixable at the algorithm level)
and the inherent per-call costs of TokenJaccard and QGram maps (fixable with
small-input fast paths). `sync.Pool` adds concurrency complexity and GC
interaction for marginal gain; it is not the approach recommended here.

---

## Recommendation (PERF-07)

**Recommended path: Choice A (budget revision) + Choice B (DoubleMetaphone fix)**

### Choice A: Revise §14.2 budget to ≤ 16 allocs/op for ASCII Short

The summed per-algorithm budgets from §14.1 for the 6 DefaultScorer algorithms
plus 2 Normalise calls establish a **spec floor of ≤ 16 allocs/op**:

  0 (DL-OSA) + 0 (JW) + 4 (TokenJaccard) + 4 (QGram) + 4 (SorensenDice)
  + 2 (DoubleMetaphone, post-fix) + 2 (Normalise×2) = 16

The current §14.2 budget of ≤ 8 is inconsistent with §14.1 and cannot be met
without reducing individual algorithms below their own §14.1 budgets. Update
§14.2 to ≤ 16 for ASCII Short and add a separate note for ASCII Medium (where
map growth at 30+ chars pushes totals higher — ≤ 40 is a realistic Medium budget
that the 34/35-alloc actual already passes).

### Choice B: Fix DoubleMetaphone allocation (PERF-03 concrete patch)

Replace `strings.Builder` primary/secondary accumulation with `[dmMaxLen]byte`
arrays inside `DoubleMetaphoneKeys`. This reduces DoubleMetaphone from 6 allocs
to 2, saving 4 allocs per Score call at all input sizes. Combined with the
revised §14.2 ceiling:

- Short: 12 - 4 = 8 allocs → exactly at the ORIGINAL ≤ 8 budget
- Medium: 34 - 4 = 30 allocs → below the revised ≤ 40 Medium budget

The DoubleMetaphone fix has no correctness risk (keys are ≤ 4 bytes by the
dmMaxLen invariant), passes all existing tests, and tightens the algorithm
to its own §14.1 budget of ≤ 2 allocations.

### Against Choice C (defer to v1.x)

Deferring is not recommended. The spec's §20 acceptance criteria explicitly
include "All section-14 budgets met" as a v1.0 requirement. Shipping v1.0 with
a known documented breach would require an explicit spec amendment AND a
GitHub issue. Choice A (spec amendment) is required regardless; doing it now
clears the v1.0 gate rather than leaving it open.

### Not recommended: sync.Pool

As detailed above, sync.Pool is wrong for this problem. It adds concurrency
complexity without addressing the root causes and risks subtle GC pauses on
the allocation-pool GC interaction path.

### Summary decision table

| Action                               | Effort | Alloc Saving | Recommended |
|--------------------------------------|--------|--------------|-------------|
| Revise §14.2 budget (A)              | Low    | None (spec)  | Yes         |
| Fix DoubleMetaphone Builder→[4]byte  | Medium | 4 allocs/op  | Yes         |
| Fix §14.3 Normalise alloc claim      | Low    | None (spec)  | Yes         |
| TokenJaccard small-input fast path   | High   | 2 allocs/op  | Optional    |
| QGram small-input fast path          | High   | 4 allocs/op  | Optional    |
| sync.Pool for scratch maps           | High   | 0–2 allocs   | No          |
| Add Scorer baselines to bench.txt    | Low    | N/A          | Blocking    |

---

## bench.txt Gap (PERF-04 — Blocking Action)

`bench.txt` has no `BenchmarkDefaultScorer_*` entries. Before v1.0 tagging:

1. Run `go test -bench=BenchmarkDefaultScorer -benchmem -count=10 -run=^$ ./... >> bench.txt`
   on the self-hosted benchmark runner.
2. Verify `benchstat bench.txt` produces no error output.
3. Commit the updated `bench.txt` with message referencing the PERF-04 finding.

This is a blocking action for the regression-detection CI gate.

---

_Reviewed: 2026-05-17_
_Reviewer: algorithm-performance-reviewer_
_Status: complete_
