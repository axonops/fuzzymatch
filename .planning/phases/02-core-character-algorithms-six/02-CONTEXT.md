---
phase: 02-core-character-algorithms-six
phase_number: 2
date: 2026-05-14
spec_loaded: false
prior_decisions_consulted: [01-CONTEXT.md (none), 01-VERIFICATION.md, all 01-*-SUMMARY.md]
---

# Phase 2: Core Character Algorithms (six) — Context

<domain>
**What this phase delivers:** the canonical six character-based similarity
algorithms — **Levenshtein**, **Damerau-Levenshtein OSA**, **Damerau-Levenshtein
Full**, **Hamming**, **Jaro**, **Jaro-Winkler** — each fresh-implemented from
its primary academic source, with literature reference vectors, mathematical
invariants, fuzz tests, allocation-budgeted benchmarks, BDD scenarios, and
entries in the cross-platform `algorithms.json` golden file.

This phase proves the end-to-end correctness-and-determinism pipeline on the
simplest non-trivial algorithm (Levenshtein) and locks in the two-row DP +
ASCII fast-path + stack-allocated-buffer pattern that every subsequent
DP-based algorithm reuses.
</domain>

<canonical_refs>
Downstream agents MUST read these before research or planning:

| Reference | Why |
|-----------|-----|
| `docs/requirements.md` §7.1.1 — Levenshtein | Public API, recurrence, score normalisation, invariants, reference vectors |
| `docs/requirements.md` §7.1.2 — Damerau-Levenshtein OSA | Recurrence, discriminating vector `ca`/`abc` (distance 3) |
| `docs/requirements.md` §7.1.3 — Damerau-Levenshtein Full | Lowrance-Wagner unrestricted; same vector returns distance 2 |
| `docs/requirements.md` §7.1.4 — Hamming | Equal-length-only contract per Hamming 1950; "defined behaviour for unequal length" — locked below |
| `docs/requirements.md` §7.1.5 — Jaro | Matching-window m, transpositions t, formula |
| `docs/requirements.md` §7.1.6 — Jaro-Winkler | Boost threshold 0.7, prefix cap 4, scale 0.1 — per Winkler 1990 NOT Wikipedia |
| `docs/requirements.md` §14 — Performance budgets | Allocation budget, ASCII fast path, two-row DP rules |
| `docs/requirements.md` §15 — Test strategy | Unit + property + fuzz + benchmark + BDD per algorithm |
| `.claude/skills/algorithm-correctness-standards/SKILL.md` | Primary source citation, formula docs, constants, invariants |
| `.claude/skills/algorithm-licensing-standards/SKILL.md` | Fresh-impl rule, no GPL/LGPL consultation, attribution format |
| `.claude/skills/performance-standards/SKILL.md` | Per-algorithm budgets, ASCII fast paths, benchstat regression |
| `.claude/skills/determinism-standards/SKILL.md` | Float-stability rules, map-iteration rule, golden file format |
| `.claude/skills/go-testing-standards/SKILL.md` | Coverage floors, property test conventions, fuzz harness pattern |
| `.planning/phases/01-foundation-infrastructure/01-04-determinism-infra-SUMMARY.md` | Golden canonical format LOCKED (sorted, LF-terminated, no BOM); `golden_canonical.go` is the canonical marshaller |
| `.planning/phases/01-foundation-infrastructure/01-05-primitives-algoid-errors-SUMMARY.md` | AlgoLevenshtein..AlgoJaroWinkler already declared in `algoid.go`; sentinel errors (`ErrInvalidInput`, etc.) available |
| `.planning/phases/01-foundation-infrastructure/01-06-primitives-normalise-SUMMARY.md` | Normalise + NormalisationOptions in place; algorithms accept normalised input from callers, do NOT call Normalise internally |
| `algoid.go` (existing) | AlgoID constants and dispatch skeleton ([23]func(a, b string) float64) — algorithms register themselves into the skeleton |
| `testdata/golden/normalisation.json` (existing) | Reference golden file structure to mirror for `algorithms.json` |
</canonical_refs>

<code_context>
**Phase 1 outputs that Phase 2 must compose with:**

- `algoid.go` — AlgoID enum with `AlgoLevenshtein`, `AlgoDamerauLevenshteinOSA`,
  `AlgoDamerauLevenshteinFull`, `AlgoHamming`, `AlgoJaro`, `AlgoJaroWinkler`
  already declared at correct positions (constants 0–5 of 23). Unexported
  `dispatch [23]func(a, b string) float64` skeleton sized and empty — algorithms
  in this phase populate the first 6 entries via `init()` blocks (or `var`
  initialisation pattern; the latter is preferred per determinism-standards
  §13.5 to avoid platform-sensitive init order).
- `errors.go` — `ErrInvalidInput`, `ErrInvalidConfiguration`, `ErrInvalidAlgorithm`,
  `ErrEmptyInput` available via `errors.Is`.
- `normalise.go` — `Normalise(s, opts)` + `NormalisationOptions` struct in place.
  Algorithms in Phase 2 take pre-normalised input — they do NOT call Normalise
  themselves; the Scorer (Phase 8) is responsible for invoking Normalise
  exactly once and passing the result to each algorithm in the composite.
- `tokenise.go` — `Tokenise(s, opts)` available. Not used by Phase 2 algorithms
  (those are token-based, Phase 6).
- `golden_canonical.go` — `CanonicalMarshalForTest` exports the locked
  canonical-JSON marshaller. `testdata/golden/algorithms.json` MUST be
  marshalled through this; never directly via `encoding/json`.
- `golden_test.go` — pattern for cross-platform `TestGolden_Xxx` tests; mirror
  the Normalisation test shape for `TestGolden_Algorithms`.
- `Makefile` — `make check` runs the full quality gate. `make bench-compare`
  diffs current run against the committed `bench.txt`; algorithms in this phase
  add 6 lines (one per algorithm benchmark) to `bench.txt`.
- `tests/bdd/` — isolated sub-module; `go.mod` already has godog v0.15.0 +
  goleak v1.3.0 + testify v1.10.0 with `replace ../..` pointing at root.
  BDD feature files for these 6 algorithms land under
  `tests/bdd/features/algorithms.feature` (one feature file with 6 scenarios,
  or 6 separate files — planner picks; per test-writer skill conventions,
  one-file-per-algorithm is the existing pattern signal).

**API ergonomics authority:** function names, signatures, option shapes,
error names are subject to `api-ergonomics-reviewer` and `user-guide-reviewer`
final review. The signatures in `docs/requirements.md` §7.1 are illustrative
per CLAUDE.md Design Principle 13; agents have veto authority. The names below
are the spec's recommendations and the most likely outcome.
</code_context>

<decisions>

### Plan decomposition / sequencing — LOCKED

**Structure:** Wave 1 (Levenshtein-first) → Wave 2 (5-way parallel).

```
Wave 1:
  plan 02-01-levenshtein

Wave 2 (parallel — 5 plans, intra-wave files_modified check must pass):
  plan 02-02-hamming
  plan 02-03-jaro
  plan 02-04-jaro-winkler
  plan 02-05-damerau-levenshtein-osa
  plan 02-06-damerau-levenshtein-full
```

**Rationale:** Levenshtein-first derisks the entire pipeline — 2-row DP,
ASCII fast path with stack-allocated `[64]int` buffers, dispatch-table
registration, golden-file entry, bench.txt entry, property tests for all
4 invariants (identity, symmetry, range, triangle inequality on distance),
BDD scenario template. Once the pattern is proven and merged, the 5
parallel plans replicate it with minimal pattern risk.

**Risk on the parallel wave:** worktrees share `algoid.go` (each algorithm
registers itself in the dispatch table), `testdata/golden/algorithms.json`
(each adds entries), `bench.txt` (each adds lines), and `tests/bdd/features/algorithms.feature` (or the per-algorithm equivalent).
The intra-wave `files_modified` overlap check in execute-phase.md will
detect any direct collisions and serialize affected plans — but the
*expected* collisions on these shared artefacts must be designed away:

- **`algoid.go` dispatch registration** — each algorithm's `init()` MUST
  register only its own AlgoID slot. The Wave 1 plan establishes the pattern
  (`var _ = registerLevenshteinDispatch()` or equivalent in a separate
  `dispatch_levenshtein.go` file per-algorithm). Each parallel plan touches
  only its own `dispatch_xxx.go`. The shared `algoid.go` stays untouched
  in Wave 2.
- **`testdata/golden/algorithms.json`** — Wave 1 establishes the file with
  Levenshtein entries. Wave 2 plans each write entries to their own staging
  file (`testdata/golden/_staging/<algo>.json`), and the post-merge gate
  re-marshals the combined file through `CanonicalMarshalForTest`. Planner
  must design this; surface it as a constraint.
- **`bench.txt`** — append-only per algorithm. Each plan adds its own
  benchmark output section. The merge resolution is "concatenate-sorted-by-
  algorithm-name"; planner specifies via a small `scripts/merge-bench.sh`.
- **BDD feature files** — one feature file per algorithm
  (`tests/bdd/features/levenshtein.feature`, `hamming.feature`, etc.).
  Zero overlap, zero merge risk.

**If the planner determines the shared-artefact merge designs are too
complex** (likely with `bench.txt` and `algorithms.json`), fall back to
sequential Wave 2 (6 sequential plans across 6 waves). The execute-phase
intra-wave overlap check will force this automatically if `files_modified`
overlap is declared.

### Hamming unequal-length behaviour — LOCKED

**Decision:** Silent zero. Hamming behaves like every other algorithm in
the catalogue — no error path, always returns a number.

- `HammingDistance(a, b string) int` — signature **matches the family
  pattern** `XxxDistance(a, b string) int`. On unequal length, returns
  `max(len(a), len(b))`. This value makes the score normalisation invariant
  `score = 1 - distance / max(len)` resolve cleanly to 0.0.
- `HammingScore(a, b string) float64` — returns `0.0` silently on unequal
  length. No error, no panic, no surprise to the Scorer.
- Both-empty (`len(a)==0 && len(b)==0`): distance 0, score 1.0.
- Equal-length: counts mismatching positions per Hamming 1950.
- Rune variants (`HammingDistanceRunes`, `HammingScoreRunes`) follow the
  same pattern: rune-count mismatch → distance `max(runeCount(a),
  runeCount(b))`, score `0.0`.

**Godoc requirement:** the file-level block comment MUST explicitly state
that unequal-length inputs return distance = `max(len)` / score = 0.0
rather than an error, and direct callers wanting strict Hamming-1950
semantics to length-check upstream:

```
// Inputs of unequal length are not an error: HammingDistance returns
// max(len(a), len(b)) and HammingScore returns 0.0. Callers wanting
// strict Hamming-1950 equal-length semantics should length-check
// before calling.
```

**Rationale:**

1. Catalogue consistency — every other algorithm in the 23-algorithm
   catalogue returns `(int, float64)` without an error path. A single
   algorithm with `(int, error)` would be an API surface wart.
2. Scorer composability — the Phase 8 Scorer can include Hamming in any
   weighted composite without special length-guard logic. Mismatched
   length reads cleanly as "this metric contributed 0".
3. Precedent — RapidFuzz and FuzzyWuzzy both return silent 0 on unequal
   length. Aligning with the dominant ecosystem behaviour lowers
   migration friction.
4. Mathematical defensibility — `score = 1 - max(len)/max(len) = 0` is
   the principled normalisation when "all positions disagree" is the
   worst case. The trailing characters of the longer string are treated
   as positions that don't match.

**Impact on Phase 8 (Scorer):** no special Hamming handling required.
The Scorer composes Hamming exactly like any other algorithm.

### identifier-similarity example scope — LOCKED

**Decision:** All 6 algorithms side-by-side on database column-name pairs.

**File:** `examples/identifier-similarity/main.go`

**Behaviour:** runnable `main()` that compares 6–10 hardcoded pairs of
database column names spanning known similarity cases:

- `user_id` / `userId` — case + separator drift
- `created_at` / `creationTimestamp` — semantic equivalence, different forms
- `status` / `state` — short tokens, semantic synonyms
- `email` / `e_mail` — separator-only difference
- `org_id` / `organisation_id` — abbreviation expansion
- `latitude` / `longitude` — length-equal, content-different (Hamming
  demonstrates value here)
- `is_deleted` / `is_active` — same shape, opposite meaning (a known
  failure case for all character-similarity algorithms; teaches the
  limitation)

**Output format:** plaintext table, columns = algorithms, rows = pairs.
Each cell is the score truncated to 4 decimals, or `ERR` if the algorithm
errored (Hamming on unequal-length pairs).

**Why this matters:**

1. Demonstrates the full Phase 2 surface in one runnable command.
2. Teaches "why use which algorithm" — Jaro-Winkler ranks `user_id`/`userId`
   highest (prefix bonus); Hamming says nothing useful for length-mismatched
   pairs; Levenshtein gives a baseline.
3. Doubles as the canonical reference for downstream consumers
   (axonops/audit, the trigger consumer per PROJECT.md) when they evaluate
   which algorithm to weight in their Scorer.
4. Meta-tested via `examples/identifier-similarity/main_test.go` — a
   `TestExample_Output` test that captures stdout and verifies the table is
   stable byte-for-byte (cross-platform determinism extends to examples).

**Lands in:** Wave 2 — coupled with the parallel plans (any of them can
extend the example with their entry). Planner picks; if simplest, attach
to plan 02-06 as the "final" plan that completes the table once all 6
algorithms exist. Alternative: a dedicated plan 02-07-identifier-similarity-example.

</decisions>

<deferred>

**Noted for later (Out of scope for Phase 2):**

- **Damerau-Levenshtein Full ASCII fast path** is a possible v1.x perf
  follow-up — the algorithm needs an alphabet-sized auxiliary table (256
  entries for ASCII, 65536+ for Unicode), which inflates allocation budget
  for short inputs. The planner should NOT optimise prematurely; ship the
  correct algorithm first, optimise later if benchstat shows it matters.
- **examples/extract-demo/, examples/audit-field-similarity/, examples/schema-dedup/** — per DX-05 / Phase 5–10 backlog. Phase 2 ships only
  `examples/identifier-similarity/`.
- **The 1-alloc-on-ASCII-fast-path follow-up from Normalise** (Phase 1
  Wave 6 SUMMARY) is a separate concern — the per-algorithm budgets for
  Phase 2 algorithms are `0 allocs on ASCII ≤ 64 bytes` (target) per
  PERF-01 + PERF-02; the planner uses stack-allocated `[64]int` arrays
  with escape-analysis checks.
- **Levenshtein with Ukkonen banding optimisation** — sub-quadratic
  worst-case time for low-distance pairs. Not required for v1.0; flag for
  v1.x performance polish if benchmarks show it matters at scale.
- **Cross-algorithm consistency tests** — when two algorithms agree on a
  reference vector (e.g. `OSA(abc, abc) == Full(abc, abc) == 1.0`), pin
  that consistency in a separate meta-test. Phase 2 task; planner decides
  whether to add a `cross_algorithm_consistency_test.go` or fold into
  individual files.

</deferred>

<open_questions>

The planner and researcher have authority to decide:

1. **Algorithm file naming.** `levenshtein.go` vs `algo_levenshtein.go`?
   The existing primitive files (algoid.go, normalise.go, tokenise.go,
   errors.go) suggest unprefixed names — `levenshtein.go`, `hamming.go`,
   `jaro.go`, `jarowinkler.go`, `damerau_osa.go`, `damerau_full.go`.
2. **Dispatch registration pattern.** `init()` vs `var _ = register...()`.
   Per determinism-standards §13.5, avoid `init()` for tables; prefer
   `var` initialisation. Algorithm registration is at module-load time
   so either is acceptable; planner picks the one that mirrors mask's
   pattern if any precedent exists.
3. **ASCII fast path buffer size.** Target `[64]int` stack-allocated
   buffer; falls back to heap-allocated `[]int` for inputs > 64. The 64
   threshold is a starting point; benchstat tuning happens at PR time.
4. **Rune variant strategy.** Spec says `XxxDistanceRunes` for every
   algorithm. Reality check: Hamming runes is straightforward (compare
   rune counts then compare runes). For Levenshtein/Damerau/Jaro, the
   rune path is a parallel implementation that operates on `[]rune`
   rather than `[]byte`. Planner specifies the rune-conversion strategy
   (eager `[]rune(s)` vs lazy `utf8.DecodeRune` loop) and which is
   tested at scale.
5. **BDD scenario shape.** One feature file per algorithm
   (`tests/bdd/features/levenshtein.feature`) with multiple scenarios
   covering reference vectors, edge cases, and invariants — vs a single
   `algorithms.feature` with one scenario outline per algorithm.
   test-writer + bdd-scenario-reviewer agents decide.

</open_questions>

<verification_carry_forward>

5 items remain pending on Phase 1's `01-HUMAN-UAT.md` — none block Phase 2:

1. Tag-push signed release on fork
2. 5-platform CI matrix byte-identical golden file
3. Branch protection on main
4. CLA Assistant signing flow
5. markdownlint-cli2 local pin

These surface in `/gsd-progress` and `/gsd-audit-uat` until verified.

</verification_carry_forward>
