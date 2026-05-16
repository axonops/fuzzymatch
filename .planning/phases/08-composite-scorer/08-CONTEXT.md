# Phase 8: Composite Scorer - Context

**Gathered:** 2026-05-16
**Status:** Ready for planning

<domain>
## Phase Boundary

Ship Layer 2 of the three-layer fuzzymatch architecture — the **composite weighted `Scorer`** that composes any subset of the 23 dispatch-registered algorithms into a single similarity score.

Surface delivered in Phase 8:

- **Construction:** `NewScorer(opts ...ScorerOption) (*Scorer, error)` — functional-options constructor; immutable after successful construction; safe for concurrent use without locks.
- **Generic option:** `WithAlgorithm(algo AlgoID, weight float64) ScorerOption` — typed AlgoID, uses the algorithm's dispatch-registered defaults for parameterised algorithms (n=3 for Q-Gram Jaccard, α=β=1 for Tversky, Jaro-Winkler inner for Monge-Elkan, default SWGParams for Smith-Waterman-Gotoh).
- **Parameterised options:** `WithQGramJaccardAlgorithm(weight float64, n int)`, `WithSorensenDiceAlgorithm(weight float64, n int)`, `WithCosineAlgorithm(weight float64, n int)`, `WithTverskyAlgorithm(weight, alpha, beta float64, n int)`, `WithMongeElkanAlgorithm(weight float64, inner AlgoID)`, `WithSmithWatermanGotohAlgorithm(weight float64, params SWGParams)` — for non-default parameters.
- **Negation option:** `WithoutAlgorithm(id AlgoID) ScorerOption` — silently no-ops if `id` is not currently in the option slice; otherwise removes the entry. Composes with `append(DefaultScorerOptions(), WithoutAlgorithm(AlgoX))...`.
- **Normalisation control:** `WithNormalisation(opts NormalisationOptions) ScorerOption` + `WithoutNormalisation() ScorerOption` (spec §8.2 terminology — REQUIREMENTS.md SCORER-08's `WithCustomNormalisation` is a typo and is overridden here).
- **Threshold:** `WithThreshold(t float64) ScorerOption` — **mandatory for `NewScorer`** (see §2 below); validates `t ∈ [0.0, 1.0]`.
- **Weight normalisation control:** `WithNormaliseWeights(normalise bool) ScorerOption` — default `true` (sum-to-1 auto-normalisation); `false` leaves raw weights and waives the [0,1] composite guarantee.
- **Methods:** `Score(a, b string) float64`, `ScoreAll(a, b string) map[AlgoID]float64` (typed AlgoID keys — see §1 SPEC OVERRIDE), `Match(a, b string) bool`, `Threshold() float64`, `Algorithms() []ScorerAlgorithm` (fresh slice per call, sorted by AlgoID ascending).
- **Defaults:** `DefaultScorer() *Scorer` (cannot fail) with documented composition per spec §8.5; `DefaultScorerOptions() []ScorerOption` returning the same composition as a mutable option slice.
- **Errors:** `ErrEmptyScorer`, `ErrInvalidWeight`, `ErrInvalidThreshold`, `ErrMissingThreshold` (new in Phase 8) + existing `ErrInvalidQGramSize`, `ErrInvalidTverskyParam`, `ErrInvalidAlgorithm` (the spec's `ErrInvalidAlgoID`).
- **Determinism artefacts:** `testdata/golden/scorer-default.json` (cross-platform byte-identical), `PropScorer_DeterministicAcrossRuns` property test, `scorer.feature` BDD file with goleak verification.
- **Docs:** `docs/scorer.md` (currently scaffold) populated with the full Scorer API guide; `docs/tuning.md` (currently scaffold) populated with weight/threshold calibration guidance.
- **Example surfaces:** `examples/identifier-similarity/main.go` gains a final Score+Match column from `DefaultScorer()` AND a new `examples/scorer-composition/main.go` demonstrating `DefaultScorerOptions() + WithoutAlgorithm(...) + WithThreshold(...)` composition.
- **llms.txt / llms-full.txt:** synced in lockstep with each public-symbol addition per the Phase 5+ per-plan llms-sync discipline.

Closes requirement IDs `SCORER-01` through `SCORER-08`.

**Out of scope:** the `scan/` sub-package (Phase 9); the `Extract`/`ExtractOne` one-to-many API (Phase 10); any new algorithms (catalogue is complete at 23 after Phase 7); any non-Scorer evolution of `Normalise` / `Tokenise` / `dispatch[]` (all three primitives are LOCKED by Phase 1/Phase 6).

</domain>

<decisions>
## Implementation Decisions

### §1. `ScoreAll` key type — LOCKED (SPEC OVERRIDE)

**`map[AlgoID]float64` with typed enum keys. OVERRIDES `docs/requirements.md` §8.3 which says `map[string]float64`.**

```go
func (s *Scorer) ScoreAll(a, b string) map[AlgoID]float64
```

**Rationale:** the project's typed-everywhere discipline already exposes `AlgoID` as a public typed enum (`type AlgoID int` with exported constants `AlgoLevenshtein`, `AlgoMongeElkan`, etc. at `algoid.go:47-176`). Consumers writing `result[fuzzymatch.AlgoLevenshtein]` get compile-time type safety; consumers needing snake_case display use `AlgoID.String()` (already implemented). Returning `map[string]float64` would lose the type discrimination and force every consumer to re-encode the snake_case key namespace — work the type system can do for free.

**REQUIREMENTS.md SCORER-05 already uses this shape (`map[AlgoID]float64`).** Spec §8.3's `map[string]float64` is the deviation; spec §8 also embeds the api-ergonomics-reviewer veto clause ("the Scorer construction, options, and method shapes below are illustrative") so this is a deliberate ergonomics call.

**Map iteration order is still non-deterministic** (Go map semantics); the documented note from spec §8.6 carries forward — consumers requiring stable order sort the keys. Internal computation iterates AlgoID-sorted order regardless (for float-determinism, per §5 below).

**Action for the planner:**
- Add the SPEC-OVERRIDE note prominently to the `Scorer.ScoreAll` godoc.
- Cite api-ergonomics-reviewer sign-off on plan 08-03 (which lands ScoreAll) — same pattern as Phase 7 plan 07-02's algorithm-licensing-reviewer sign-off recording.
- Open a follow-up to amend `docs/requirements.md` §8.3 to say `map[AlgoID]float64` (deferred to plan 08-04's docs/scorer.md commit).

### §2. Threshold — MANDATORY for `NewScorer` — LOCKED

**`NewScorer(opts ...ScorerOption) (*Scorer, error)` returns `ErrMissingThreshold` if no `WithThreshold(t)` option was passed. `DefaultScorer()` bakes 0.85 in (per spec §8.5) so casual consumers using the default are unaffected.**

```go
// Custom Scorer construction is one line longer:
s, err := fuzzymatch.NewScorer(
    fuzzymatch.WithAlgorithm(fuzzymatch.AlgoLevenshtein, 1.0),
    fuzzymatch.WithThreshold(0.7),  // REQUIRED — no default
)

// DefaultScorer bakes the threshold:
s := fuzzymatch.DefaultScorer()
_ = s.Threshold()  // 0.85
```

**Rationale:** the default-threshold question has no safe answer. A library-wide default of `1.0` (only exact match) would silently produce "no matches found" for users who forget `WithThreshold`. A default of `0.0` would silently make every comparison a match. Inheriting `DefaultScorer`'s `0.85` is arbitrary for non-default compositions (0.85 is calibrated for the SPECIFIC 6-algorithm mix; a Levenshtein-only Scorer should not pretend 0.85 is meaningful). Requiring `WithThreshold` forces an explicit calibration step at construction time — the place where the consumer most has the context to choose.

**New sentinel error:**

```go
// ErrMissingThreshold indicates NewScorer was called without WithThreshold.
// The threshold is a calibration parameter with no universally-safe default,
// so the library refuses to guess. Pass WithThreshold(t) with t ∈ [0.0, 1.0]
// during construction, or use DefaultScorer() for the opinionated default
// composition that bakes 0.85 in.
//
// Returned by NewScorer when no WithThreshold option is present in the
// variadic opts slice.
//
// Discriminate via errors.Is(err, fuzzymatch.ErrMissingThreshold); never
// match the error message string.
var ErrMissingThreshold = errors.New("fuzzymatch: scorer threshold required (pass WithThreshold or use DefaultScorer)")
```

**Action for the planner:**
- Plan 08-01 declares `ErrMissingThreshold` alongside `ErrEmptyScorer`, `ErrInvalidWeight`, `ErrInvalidThreshold`.
- Plan 08-02 (`NewScorer` validation) gates on the missing-threshold check FIRST (before any other validation) so the error message is unambiguous when the user forgets the threshold AND has another option problem.
- `docs/scorer.md` quickstart MUST show `WithThreshold` in the first code snippet — this is a major DX shape and the user has to internalise it early.
- `docs/tuning.md` MUST have a "How to pick a threshold" section.

### §3. Normalisation flow — LOCKED

**Pre-normalise once at the Scorer boundary; pass `na, nb` to ALL algorithms (char-based AND token-based); token-based algorithms keep their hardcoded internal `Tokenise(s, DefaultTokeniseOptions())` call. Zero changes to existing token-based algorithm internals.**

```go
// Inside Scorer.Score(a, b):
na := a; nb := b
if s.applyNormalisation {
    na = Normalise(a, s.normaliseOpts)
    nb = Normalise(b, s.normaliseOpts)
}

// Single iteration in AlgoID-sorted order (DET-04):
acc := 0.0
for _, entry := range s.algorithmsAlgoIDSorted {
    score := entry.scoreFn(na, nb)  // dispatch[id] or a parameter-capturing closure
    acc = acc + (entry.weight * score)
}
return acc
```

**Why this works correctly for token-based algorithms:**

1. `Normalise(s, opts)` with default `NormalisationOptions{Lowercase: true, StripSeparators: true, SeparatorChars: "_-.:/", SplitCamelCase: true}` produces a string with **separator characters AND camelCase boundaries already replaced with spaces, lowercased** (per `normalise.go` field godoc: `SplitCamelCase: when true, insert a single ASCII space at every lowercase → uppercase rune transition before any other folding`).
2. Token-based algorithms (`Monge-Elkan`, `TokenSortRatio`, `TokenSetRatio`, `PartialRatio`, `TokenJaccard`) call `Tokenise(s, DefaultTokeniseOptions())` internally (e.g. `monge_elkan.go:394-395`, `token_jaccard.go:204-205`).
3. `DefaultTokeniseOptions()` separator-chars `"_-.:/ \t\n\r"` will split the pre-normalised "foo bar baz" on whitespace into `["foo", "bar", "baz"]` — the correct tokenisation.
4. `MongeElkanScore`'s `opts NormalisationOptions` parameter (already present at `monge_elkan.go:377`) is **vestigial — accepted-but-ignored** (`_ = opts` on line 393). The Scorer passes its own opts for forward-compat; the function body remains a no-op on opts. The Phase 7 inheritance documented this as "accepted for Phase 8 Scorer compatibility" — Phase 8 confirms the contract: Scorer pre-normalises; ME (and other token algos) do not re-normalise inside; opts is plumbed through but unused.

**`WithoutNormalisation()` semantics:** sets the Scorer's internal `applyNormalisation` flag to `false`. Score still proceeds; `na = a, nb = b` (raw input passed to every algorithm). Token algos still call `Tokenise(s, DefaultTokeniseOptions())` on raw input — this is the documented "operate on raw input" behaviour per spec §9.4 ("Algorithms then operate on raw input").

**`WithNormalisation(NormalisationOptions{})` vs `WithoutNormalisation()`:** the zero-value `NormalisationOptions{}` (`Lowercase: false, StripSeparators: false, SplitCamelCase: false`) makes `Normalise` a no-op pass-through — semantically equivalent to `WithoutNormalisation()` but takes one extra function-call cycle. Document both as valid; `WithoutNormalisation()` is the sugared form.

**Action for the planner:**
- Plan 08-02's `NewScorer` constructor stores both `applyNormalisation bool` and `normaliseOpts NormalisationOptions` in `scorerConfig`.
- Plan 08-02's `Score` method uses an `if applyNormalisation` gate around the two `Normalise` calls — avoid the function-call overhead when normalisation is disabled.
- Plan 08-04's BDD scenarios MUST cover:
  - default scorer + identifier-style input (XMLParser vs xml_parser) → match
  - `WithoutNormalisation()` + identifier-style input → no match (raw bytes differ)
  - `WithNormalisation(custom)` + Unicode input (`café` vs `cafe`) → behaviour per opts
- NO changes to monge_elkan.go, token_jaccard.go, token_sort_ratio.go, token_set_ratio.go, partial_ratio.go in Phase 8. The token-based algorithm files are NOT touched.

### §4. Plan decomposition — LOCKED at 4 plans

**Four plans, finalisation in plan 08-04. Each plan is self-contained per the Phase 2-7 standard.**

| Plan | Scope | Closes |
|------|-------|--------|
| **08-01** | Sentinel errors (`ErrEmptyScorer`, `ErrInvalidWeight`, `ErrInvalidThreshold`, `ErrMissingThreshold` — new; existing `Err{InvalidQGramSize, InvalidTverskyParam, InvalidAlgorithm}` referenced) + `ScorerOption func(*scorerConfig) error` type + `scorerConfig` struct + ALL option functions: `WithAlgorithm`, `WithoutAlgorithm`, `WithQGramJaccardAlgorithm`, `WithSorensenDiceAlgorithm`, `WithCosineAlgorithm`, `WithTverskyAlgorithm`, `WithMongeElkanAlgorithm`, `WithSmithWatermanGotohAlgorithm`, `WithNormalisation`, `WithoutNormalisation`, `WithThreshold`, `WithNormaliseWeights`. Unit tests for every option's happy path + every error path. No `Scorer` type or methods yet — options accumulate state only. | Foundation for SCORER-01..08 |
| **08-02** | `Scorer` struct + `NewScorer` validation pipeline (gates on missing-threshold first → empty algorithms → invalid weights → invalid threshold → invalid AlgoIDs → q-gram-size/Tversky-params validation per option) + weight auto-normalisation (sum-to-1) + last-write-wins behaviour for duplicate AlgoIDs (later option wins; `WithoutAlgorithm` removes) + `Score(a, b) float64` method (pre-normalise once + sorted iteration + explicit `(w*s)+acc` reduction) + `Match(a, b) bool` method. Unit tests including last-write-wins regression, weight normalisation, and the canonical 5-line quickstart. | SCORER-01, SCORER-03, SCORER-04, SCORER-06 (partial — Threshold accessor lands in 08-03) |
| **08-03** | `ScoreAll(a, b) map[AlgoID]float64` (SPEC OVERRIDE — typed AlgoID keys; api-ergonomics-reviewer sign-off recorded in PR) + `Threshold() float64` accessor + `Algorithms() []ScorerAlgorithm` (fresh slice per call, sorted by AlgoID ascending) + `ScorerAlgorithm{ID AlgoID, Weight float64}` struct + `DefaultScorer() *Scorer` (cannot fail) per spec §8.5 + `DefaultScorerOptions() []ScorerOption` returning the same composition as a mutable option slice + property tests: `PropScorer_DeterministicAcrossRuns`, `PropScorer_WeightSumOne` (when `WithNormaliseWeights(true)`), `PropScorer_ScoreInRange` (when weights normalised) + concurrent test invoking Score/ScoreAll/Match from N goroutines on the same `*Scorer` with `-race`. | SCORER-02, SCORER-05, SCORER-07 |
| **08-04** | **Finalisation.** `testdata/golden/scorer-default.json` (corpus per §6 below; uses identifier-similarity 14 rows + 8-12 Scorer-specific threshold-edge entries; loaded by `algorithms_golden_test.go` or a new `scorer_golden_test.go`) + `tests/bdd/features/scorer.feature` (8-12 scenarios per §7 below) + `goleak.VerifyTestMain(m)` hook in `tests/bdd/scorer_main_test.go` (or extension of an existing TestMain) + `examples/identifier-similarity/main.go` gains a final Score+Match column from `DefaultScorer()` + new `examples/scorer-composition/main.go` (and companion `main_test.go` with golden stdout) demonstrating `DefaultScorerOptions() + WithoutAlgorithm + WithThreshold` composition + `docs/scorer.md` populated (replacing scaffold) + `docs/tuning.md` populated with the "How to pick a threshold" + "How to pick weights" sections + `llms.txt` + `llms-full.txt` extended with all Phase 8 exported symbols. | SCORER-08 (final wiring); all v1 SCORER requirements flip to Met in REQUIREMENTS.md |

**Plan dependency:** strict linear sequence 08-01 → 08-02 → 08-03 → 08-04. No parallel waves — each plan's surface depends on the previous plan's types.

### §5. Float-determinism reduction — LOCKED (carry-forward from Phase 5 Cosine)

**Explicit `(weight * score) + acc` parenthesisation, left-to-right reduction, AlgoID-sorted iteration.**

```go
// Inside Scorer.Score(a, b):
const acc = 0.0  // mutable accumulator declared float64
for _, entry := range s.algorithmsAlgoIDSorted {  // sorted at NewScorer time
    score := entry.scoreFn(na, nb)
    acc = acc + (entry.weight * score)  // EXPLICIT parens
}
return acc
```

**Verification:**
- algorithm-correctness-reviewer + determinism-reviewer gate on plan 08-02 PR for the reduction loop.
- No `math.Pow`, no `math.FMA`, no `math.Log`, no `math.Exp`, no parallel/atomic float ops (carry-forward from PROJECT.md "What NOT to Use").
- Cross-platform golden file `scorer-default.json` (plan 08-04) is the load-bearing acceptance test.

### §6. `scorer-default.json` golden corpus — LOCKED

**Reuse the existing `examples/identifier-similarity/main.go` 14-row corpus + 8-12 Scorer-specific threshold-edge / option-combination entries. Target 22-26 entries total.**

Schema (per Phase 1 D-12 canonical-form locked across all 5 golden files):

```json
{
  "_metadata": {
    "phase": 8,
    "generated_at": "<ISO timestamp at -update time>",
    "scorer_signature": "DefaultScorer-2026-05-16"
  },
  "entries": [
    {
      "a": "user_id",
      "b": "userId",
      "score": 0.987,
      "match": true,
      "scoreAll": {
        "AlgoDamerauLevenshteinOSA": 1.0,
        "AlgoJaroWinkler": 1.0,
        "AlgoTokenJaccard": 1.0,
        "AlgoQGramJaccard": 1.0,
        "AlgoSorensenDice": 1.0,
        "AlgoDoubleMetaphone": 1.0
      },
      "scorer_config": "DefaultScorer"
    },
    ...
  ]
}
```

**Mandatory coverage rows beyond the identifier-similarity reuse:**

1. **Identity:** `("hello", "hello", DefaultScorer)` → `score: 1.0, match: true`.
2. **Both-empty:** `("", "", DefaultScorer)` → behaviour documented; pin the actual score.
3. **One-empty:** `("", "hello", DefaultScorer)` → behaviour documented; pin.
4. **Just-above threshold:** a pair whose composite score is in `[0.85, 0.86]` (`match: true`, very close to boundary).
5. **Just-below threshold:** a pair in `[0.84, 0.85)` (`match: false`).
6. **Unicode pre-/post-NFC:** `("café", "cafe")` exercising default normalisation's diacritic-stripping behaviour (pin the score regardless of which side wins).
7. **Phonetic-match-only-not-edit-similar:** `("Smith", "Schmidt")` — composite reflects DM full match but DL-OSA low.
8. **`WithoutNormalisation` variant:** `("XMLParser", "xml_parser", DefaultScorer.applied-without-normalisation)` — pin the no-match score.
9. **`WithoutAlgorithm(AlgoDoubleMetaphone)` variant:** `("Smith", "Schmidt", DefaultScorer minus DM)` — pin the recomputed score and confirm `AlgoDoubleMetaphone` absent from `scoreAll`.
10. **Custom single-algorithm Scorer:** `WithAlgorithm(AlgoLevenshtein, 1.0) + WithThreshold(0.5)` over `("kitten", "sitting")` — pin.
11. **`WithNormaliseWeights(false)` variant:** raw-weight composite that may exceed 1.0 (or under-sum); pin the actual value to confirm the no-clamp behaviour.

**Golden file regenerated via `go test ./... -update` per Phase 1 D-13.**

**Action for the planner:**
- Plan 08-04 produces the golden file via the standard `-update` flag mechanism.
- `scorer-default.json` MUST diff byte-identically across the CI matrix (linux/amd64, linux/arm64, darwin/amd64, darwin/arm64, windows/amd64) — load-bearing acceptance test for §13.3 spec.
- `docs/scorer.md` should reference the golden file as a runnable spec.

### §7. BDD coverage in `scorer.feature` — LOCKED at 8-12 scenarios

**Mandatory scenario classes:**

1. **Default scorer happy path:** identifier-style match (`user_id` vs `userId`) → `Match` returns true.
2. **Default scorer below threshold:** dissimilar inputs → `Match` returns false.
3. **Custom 1-algorithm Scorer:** `WithAlgorithm(AlgoLevenshtein, 1.0) + WithThreshold(0.7)` → behaviour matches direct `LevenshteinScore` call.
4. **Custom 2-algorithm Scorer with explicit weights:** composite score is the weighted sum.
5. **`WithoutAlgorithm` composition:** `DefaultScorerOptions() + WithoutAlgorithm(AlgoDoubleMetaphone)` produces a Scorer whose `Algorithms()` excludes DM and whose `Score` differs from `DefaultScorer().Score` on phonetic-divergent input.
6. **Last-write-wins for duplicate AlgoID:** two `WithAlgorithm(AlgoLevenshtein, w)` calls — only the latter weight survives.
7. **`WithoutNormalisation`:** `XMLParser` vs `xml_parser` → no match (raw bytes differ); same pair on default Scorer matches.
8. **Threshold mandatory:** `NewScorer(WithAlgorithm(AlgoLevenshtein, 1.0))` (no `WithThreshold`) → returns `ErrMissingThreshold`.
9. **`ErrEmptyScorer`:** `NewScorer(WithThreshold(0.5))` (no algorithms) → returns `ErrEmptyScorer`.
10. **`ErrInvalidWeight`:** `WithAlgorithm(AlgoLevenshtein, -0.5)` → `ErrInvalidWeight`.
11. **Concurrent safety:** N goroutines call `Score` / `ScoreAll` / `Match` on the same `*Scorer` — all return identical results; `goleak.VerifyTestMain` confirms no goroutine leaks.
12. **`ScoreAll` AlgoID keys:** `DefaultScorer().ScoreAll("a", "b")` map keys are typed `AlgoID` values (compile-time gate, not runtime — but the scenario asserts the documented snake_case `AlgoID.String()` display when iterating).

**Phase 7 carries forward goleak BDD discipline** — `goleak.VerifyTestMain(m)` in `tests/bdd/*_main_test.go` is the standard hook.

### §8. ME's vestigial `opts NormalisationOptions` parameter — kept as no-op

**`MongeElkanScore(a, b string, inner AlgoID, opts NormalisationOptions)` keeps its `opts` parameter accepted-but-ignored (`_ = opts` on `monge_elkan.go:393`). Phase 8 does NOT modify the ME function signature or body.**

Rationale: changing the ME signature breaks Phase 6's public API and forces a v0.x → v0.y minor bump for token-based callers who pass the param. The clean fix is to either (a) drop the param in v1.0 final freeze (Phase 11 API freeze opportunity), or (b) wire it up to control the in-ME `Tokenise` call (orthogonal feature, future v1.x). Phase 8 picks the no-touch option and documents the vestigial nature.

**Action for the planner:**
- Phase 8 plans do NOT modify `monge_elkan.go` or any other token-based algorithm file.
- Add a note to `docs/scorer.md` documenting that ME's `opts` parameter is currently a no-op when invoked through the Scorer.
- Open a Phase 11 (or v1.x) tracking item: "Decide ME `opts` parameter — drop or wire" (captured in Deferred Ideas below).

### Claude's Discretion

The planner (gsd-planner) chooses, without re-asking the user:

- **Exact `scorerConfig` internal layout** — slice-of-entries vs map-keyed-by-AlgoID — both achieve last-write-wins; planner picks based on which simplifies `WithoutAlgorithm`'s no-op-on-absent semantic.
- **Closure-capture mechanism for parameterised algorithms** — whether `WithQGramJaccardAlgorithm(w, n)` stores `n` separately and reconstructs a `func(a, b string) float64` at `NewScorer` time, or stores the closure directly in the option struct. Both produce the same observable behaviour; allocation budget governs.
- **Exact wording of every godoc line** beyond the SPEC-OVERRIDE notice on ScoreAll and the §2 ErrMissingThreshold paragraph.
- **`scorer-default.json` exact entry count between 22 and 26** — driven by which threshold-edge pairs surface during golden generation.
- **BDD scenario count between 8 and 12** — driven by feature-file readability.
- **Whether plan 08-03's concurrent test uses `t.Parallel` + `sync.WaitGroup`** or `errgroup` (root module is stdlib-only — must be `sync.WaitGroup`).
- **`docs/scorer.md` and `docs/tuning.md` exact prose structure** — docs-writer + user-guide-reviewer agents have authority.
- **Whether `examples/scorer-composition/main.go` is a single-file program** or split into multiple `main_*.go` files demonstrating different composition patterns.
- **Whether to introduce a `scorer_internal_test.go` file** for testing unexported `scorerConfig` invariants (recommend yes for last-write-wins regression).
- **Allocation budget for `DefaultScorer().Score()`** — spec §14.2 caps it at ≤ 8 allocations on ASCII ≤ 50 chars; planner records the actual count in plan 08-04's bench fixture.

</decisions>

<canonical_refs>
## Canonical References

**Downstream agents MUST read these before planning or implementing.**

### Spec & requirements (project-internal)

- `docs/requirements.md` §8 — Layer 2 Scorer authoritative spec (8.1 Construction, 8.2 Options, 8.3 Methods, 8.4 Weight semantics, 8.5 Defaults, 8.6 ScoreAll behaviour). **NOTE: §8.3's `map[string]float64` is SPEC-OVERRIDDEN by §1 above — Scorer ships `map[AlgoID]float64` per api-ergonomics-reviewer veto authority documented in spec §8 itself.**
- `docs/requirements.md` §9 — Normalisation Pipeline (used by Scorer at pre-comparison boundary)
- `docs/requirements.md` §10 — Tokenisation (called internally by token-based algorithms; Scorer does NOT call directly)
- `docs/requirements.md` §11 — Phonetic Algorithm Integration (Scorer binary 0/1 behaviour for Soundex/DM/NYSIIS/MRA — handled automatically by `dispatch[]`)
- `docs/requirements.md` §13.2 — Scorer composite stability (byte-identical for same config + same input)
- `docs/requirements.md` §13.3 — Cross-platform determinism (scorer-default.json is the gate)
- `docs/requirements.md` §13.4 — Map iteration discipline (applies to `ScoreAll`)
- `docs/requirements.md` §13.6 — Property tests including `PropScorer_DeterministicAcrossRuns`
- `docs/requirements.md` §14.2 — Scorer performance budgets (< 30µs for default; ≤ 8 allocations)
- `docs/requirements.md` §15.1 — Scorer unit-test coverage requirements
- `docs/requirements.md` §15.6 — BDD scenarios in `tests/bdd/features/scorer.feature`
- `docs/requirements.md` §15.9 — goleak verification
- `.planning/REQUIREMENTS.md` lines 59-68 — SCORER-01..SCORER-08 traceability. **NOTE: SCORER-05 says `map[AlgoID]float64` (matches §1 above); SCORER-08 says `WithCustomNormalisation()` which is a typo — the option is `WithNormalisation(NormalisationOptions)` per spec §8.2.**
- `.planning/REQUIREMENTS.md` lines 229-236 — SCORER requirement status (currently Pending; flips to Met after Phase 8)
- `.planning/ROADMAP.md` Phase 8 section — goal + 4 success criteria (concurrent-safe; DefaultScorer + DefaultScorerOptions; Score/ScoreAll/Match/Threshold/Algorithms shapes; scorer-default.json + BDD + goleak)
- `.planning/PROJECT.md` — zero-runtime-deps + no-cgo constraints carry forward; api-ergonomics-reviewer + user-guide-reviewer agents have final API authority

### Carry-forward from prior phases

- `.planning/phases/01-foundation-infrastructure/01-CONTEXT.md` D-01 (AlgoID dispatch array — Phase 8 consumes via `dispatch[id]`), D-12 (golden file canonical form sorted struct + `json.MarshalIndent(v,"","  ")`), D-13 (`-update` regen flag — Phase 8's `scorer-default.json` follows this)
- `.planning/phases/05-q-gram-algorithms/05-CONTEXT.md` — Cosine float-determinism pattern (explicit `(x*y)+z` parenthesisation, sorted-key iteration, `math.Sqrt`-only). Phase 8's Scorer reduction loop follows the exact same pattern.
- `.planning/phases/06-token-based-algorithms/06-CONTEXT.md` — Monge-Elkan's vestigial `opts NormalisationOptions` param documented as "accepted for Phase 8 Scorer compatibility" — Phase 8 §8 above confirms it stays no-op.
- `.planning/phases/07-phonetic-algorithms/07-CONTEXT.md` §1 (cross-validation pin mechanism), §3 (algorithm-licensing-reviewer sign-off recorded in PR description — same pattern for Phase 8's api-ergonomics-reviewer sign-off on the ScoreAll SPEC OVERRIDE)

### Primary code-context references

- `algoid.go` — AlgoID enum (23 algorithms, slots 0-22), `dispatch [numAlgorithms]func(a, b string) float64` array (lines 310-324), package-internal dispatch helpers (Phase 8 adds the bounds + nil check helper for ErrInvalidAlgorithm/ErrInvalidAlgoID surfacing)
- `dispatch_*.go` (23 files) — each registers `dispatch[AlgoXxx] = XxxScore` via the `var _ = func() bool {...}()` idiom. Phase 8 does NOT modify these files but READS the populated dispatch table.
- `normalise.go` — `Normalise(s string, opts NormalisationOptions) string`, `NormalisationOptions` struct, `DefaultNormalisationOptions()`. Phase 8's Scorer.Score pre-normalises via these.
- `tokenise.go` — `Tokenise(s string, opts TokeniseOptions) []string`, `TokeniseOptions` struct, `DefaultTokeniseOptions()`. Phase 8 does NOT call Tokenise directly — token-based algorithms continue to call it internally.
- `errors.go` — existing sentinels (`ErrInvalidInput`, `ErrInvalidConfiguration`, `ErrInvalidQGramSize`, `ErrInvalidTverskyParam`, `ErrInvalidAlgorithm`, `ErrEmptyInput`). Phase 8 adds 4 new sentinels: `ErrEmptyScorer`, `ErrInvalidWeight`, `ErrInvalidThreshold`, `ErrMissingThreshold`.
- `monge_elkan.go` lines 377-393 — `MongeElkanScore(a, b string, inner AlgoID, opts NormalisationOptions)` with `_ = opts`. Phase 8 confirms vestigial opts (§8); does NOT modify.
- `examples/identifier-similarity/main.go` — 14 rows × 23 columns. Phase 8 plan 08-04 extends to 14 rows × 25 columns (adds Score + Match from DefaultScorer).
- `testdata/golden/algorithms.json` (184 entries × 23 algorithms) — pinned. Phase 8 ADDS `testdata/golden/scorer-default.json` (separate file; not merged into algorithms.json).
- `testdata/golden/phonetic-codes.json` — Phase 7's string-equality golden file. Phase 8 does NOT modify; orthogonal surface.
- `tests/bdd/features/` (23 algorithm feature files + `monge_elkan_phonetic_inner.feature`). Phase 8 ADDS `scorer.feature`.

### Project skills (correctness, ergonomics, performance gates)

- `.claude/skills/algorithm-correctness-standards/SKILL.md` — primary-source citation discipline (Scorer composes algorithms; no new algorithm citations needed; §8's spec reference is the primary source for the Scorer surface itself)
- `.claude/skills/determinism-standards/SKILL.md` — **LOAD-BEARING for Phase 8**: no map iteration on output paths (ScoreAll's map iteration is documented non-deterministic, but internal Scorer.Score iterates AlgoID-sorted); no transcendentals; explicit `(w*s)+acc` parenthesisation; sorted-key iteration for the composite reduction
- `.claude/skills/performance-standards/SKILL.md` — Scorer budget < 30µs / ≤ 8 allocs (spec §14.2); benchmark fixture required in plan 08-04
- `.claude/skills/go-coding-standards/SKILL.md` — no testify in root tests, sentinel-error pattern (Phase 1 inheritance), no cgo, no goroutines in root
- `.claude/skills/go-testing-standards/SKILL.md` — coverage targets ≥ 95% overall, ≥ 90% per file, 100% on public API surface; property tests via `testing/quick`
- `.claude/skills/documentation-standards/SKILL.md` — Phase 8 lands the first heavy use: docs/scorer.md + docs/tuning.md populated from scaffold
- `.claude/skills/fuzzymatch-review-protocol/SKILL.md` — agent gate sequence; **api-ergonomics-reviewer is the gating reviewer for Phase 8** (signs off on every plan's API surface; the SPEC OVERRIDE on ScoreAll keys is the critical sign-off)

### Phase 8 review agent gates (per CLAUDE.md "Agent Gates")

- **api-ergonomics-reviewer** (LOAD-BEARING) — every plan's PR; final authority on function names, option shapes, error names. **Signs off the §1 SPEC OVERRIDE on `ScoreAll` keys explicitly in plan 08-03's PR description.**
- **algorithm-correctness-reviewer** — plan 08-02 PR (the composite reduction is the only "algorithm" surface in Phase 8); verifies the `(w*s)+acc` pattern matches spec §13.3
- **algorithm-performance-reviewer** — plan 08-04 PR; verifies < 30µs / ≤ 8 allocs budget against the new bench fixture
- **determinism-reviewer** — plan 08-04 PR; verifies cross-platform `scorer-default.json` diffs byte-identically
- **user-guide-reviewer** — plan 08-04 PR; signs off on `docs/scorer.md` and `docs/tuning.md`
- **test-writer** + **bdd-scenario-reviewer** — plan 08-04 PR; signs off on `scorer.feature` scenario coverage
- **commit-message-reviewer** — every commit
- **issue-writer** — before opening any related GitHub issues (Phase 8 may surface the "drop ME opts param" follow-up; that issue needs the writer's review)

</canonical_refs>

<code_context>
## Existing Code Insights

### Reusable Assets

- **`dispatch[]` array (algoid.go:324)** — fully populated with all 23 algorithm score functions after Phase 7. Phase 8's Scorer reads `dispatch[id]` directly for non-parameterised algorithms; parameterised options (Q-Gram, Cosine, Sørensen-Dice, Tversky, Monge-Elkan, SWG) construct closures that ignore dispatch and call the corresponding `XxxScore(...)` function with consumer-supplied params.
- **`Normalise(s string, opts NormalisationOptions) string` (normalise.go)** — Scorer.Score's pre-comparison step. Already exercises the ASCII fast path and Unicode NFC + diacritic-stripping behaviour; Phase 8 does not modify Normalise.
- **`AlgoIDs() []AlgoID` (algoid.go:282-308)** — returns the 23 AlgoIDs in canonical order. Useful for plan 08-03's `Algorithms()` accessor implementation as the canonical sort baseline.
- **`AlgoID.String()` (algoid.go ~line 200)** — returns the snake_case name. Used in `docs/scorer.md` rendering and for the `ScorerAlgorithm` debug `%s` Stringer if added.
- **Existing sentinel error files (errors.go)** — Phase 8 appends 4 new sentinels following the established godoc + Discriminate-via-errors.Is pattern.
- **goleak via `tests/bdd/go.mod`** — already a test-only dependency; Phase 8's scorer BDD feature uses the existing `goleak.VerifyTestMain` hook pattern.

### Established Patterns

- **Functional-options pattern with last-write-wins on duplicate AlgoIDs** — `Phase 5/6/7` algorithm-option patterns informed this; Phase 8 is the first multi-option composite consumer.
- **Float-determinism reduction loop** — Cosine's `(x*y) + acc` pattern (Phase 5 §) carries forward verbatim to the Scorer composite. Sorted-key iteration; explicit parens; left-to-right reduction; no transcendentals.
- **`var _ = func() bool {...}()` idiom** — used by every `dispatch_*.go` file. Phase 8 does NOT add new dispatch entries (catalogue is complete at 23). The Scorer is package-state-free at load time.
- **Staging-golden → finalisation merge** — does NOT apply to Phase 8. `scorer-default.json` is a single new golden file produced in plan 08-04; not staged across plans.
- **`-update` flag regen** — plan 08-04's `scorer-default.json` is generated via `go test -run TestScorerGolden -update ./...` per Phase 1 D-13.
- **BDD goleak verification** — `tests/bdd/scorer_main_test.go` (or extension of existing) wraps `goleak.VerifyTestMain(m)`.
- **api-ergonomics-reviewer sign-off recorded in PR description** — Phase 7 plan 07-02's algorithm-licensing-reviewer pattern. Phase 8 plan 08-03's PR description records the api-ergonomics-reviewer sign-off on the §1 ScoreAll SPEC OVERRIDE.

### Integration Points

- **Scorer.Score → dispatch[]** — the hot path. Zero allocations on the dispatch indexing; allocations come from Normalise (ASCII fast path: zero alloc; Unicode: 1-2 allocs) + ScoreAll map.
- **Scorer.ScoreAll → fresh map** — `map[AlgoID]float64` allocated per call; documented per spec §8.6.
- **Scorer.Algorithms → fresh slice** — `[]ScorerAlgorithm` allocated per call; sorted by AlgoID; documented per spec §8.3.
- **docs/scorer.md ↔ examples/scorer-composition/main.go** — quickstart example in docs MUST compile and MUST be referenced in the example program; meta-test verifies they stay in sync (Phase 1 pattern: ai_friendly_test.go).
- **llms.txt sync** — every new exported symbol gets a line in llms.txt + a full entry in llms-full.txt IN THE SAME PLAN that adds the symbol (Phase 5+ per-plan llms-sync discipline).
- **examples/identifier-similarity ↔ scorer-default.json** — both consume DefaultScorer; the corpus row in identifier-similarity should match the entry in scorer-default.json byte-identically for the same input pair.

</code_context>

<specifics>
## Specific Ideas

- **Quickstart code snippet for `docs/scorer.md`:** the FIRST code block on the page MUST be a 5-line Scorer call exactly as a v1.0 consumer will write it. This is the load-bearing acceptance test for the API surface ergonomics. Suggested form (planner adjusts wording per api-ergonomics-reviewer):

  ```go
  package main

  import "github.com/axonops/fuzzymatch"

  func main() {
      s := fuzzymatch.DefaultScorer()
      if s.Match("user_id", "userId") {
          // similar
      }
  }
  ```

  Plus a second snippet showing the custom-Scorer + WithThreshold pattern:

  ```go
  s, err := fuzzymatch.NewScorer(
      fuzzymatch.WithAlgorithm(fuzzymatch.AlgoLevenshtein, 0.6),
      fuzzymatch.WithAlgorithm(fuzzymatch.AlgoJaroWinkler, 0.4),
      fuzzymatch.WithThreshold(0.75),
  )
  ```

- **Quickstart pattern for `docs/tuning.md`:** the "How to pick a threshold" section MUST recommend running `scan.Check` on the consumer's own data corpus, recording the false-positive / false-negative trade-off, and adjusting in 0.05 increments. Reference forward to Phase 9's scan layer.

- **`examples/scorer-composition/main.go`** demonstrates the "default minus DoubleMetaphone" use case explicitly (the rationale for §1's `WithoutAlgorithm` decision):

  ```go
  // Demonstrating: default scorer minus phonetic, for numeric-identifier data
  opts := append(fuzzymatch.DefaultScorerOptions(),
      fuzzymatch.WithoutAlgorithm(fuzzymatch.AlgoDoubleMetaphone),
      fuzzymatch.WithThreshold(0.80),  // lower because we lost one signal
  )
  s, _ := fuzzymatch.NewScorer(opts...)
  ```

</specifics>

<deferred>
## Deferred Ideas

- **Drop ME's vestigial `opts NormalisationOptions` parameter** — `monge_elkan.go:377` accepts but ignores opts (`_ = opts` on line 393). Phase 8 keeps it as-is to avoid a v0.x → v0.y minor API bump. Two cleanup paths for v1.0 freeze (Phase 11) or v1.x: (a) remove the parameter from the public signature; (b) wire it through to control Tokenise's behaviour inside ME. Track as a Phase 11 API-freeze decision.

- **Amend `docs/requirements.md` §8.3** to say `map[AlgoID]float64` instead of `map[string]float64` (matching Phase 8 §1 SPEC OVERRIDE + REQUIREMENTS.md SCORER-05 wording). Land in plan 08-04's docs commit alongside `docs/scorer.md`.

- **Amend REQUIREMENTS.md SCORER-08** to replace `WithCustomNormalisation()` with `WithNormalisation(opts)` matching spec §8.2. Land in plan 08-04 docs commit.

- **Pooled `transform.Transformer` for Normalise** — `normalise.go` constructs per-call; deferred to v1.x perf revisit (already in Phase 1 carry-forward). Phase 8 does NOT touch.

- **DefaultScorer composition v1.x revisit** — if downstream consumers (e.g. axonops/audit in Phase 11) find the 0.85 threshold or the 6-algorithm mix doesn't fit their data, calibrate via the Phase 11 integration shakedown.

- **Scorer-level allocation reuse via sync.Pool** — for high-throughput callers, ScoreAll's per-call map allocation could be pooled. Phase 8 ships fresh maps; v1.x perf opportunity if benchmarks surface it.

- **`Scorer.MarshalJSON` / serialisable Scorer config** — out of v1.0 scope; consumers who persist Scorer configurations encode the options themselves.

- **Threshold-edge BDD scenarios on every option combination** — Phase 8 ships 8-12 scenarios; exhaustive coverage of 2^N option combinations is a v1.x test-writer opportunity.

</deferred>

---

*Phase: 8-composite-scorer*
*Context gathered: 2026-05-16*
