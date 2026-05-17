---
status: approved_with_changes
agent: api-ergonomics-reviewer
scope: entire public API (phases 1-8)
reviewed: 2026-05-17T00:00:00Z
finding_counts:
  critical: 6
  important: 18
  improvement: 21
  total: 45
---

# fuzzymatch — full-surface API ergonomics review (phases 1–8)

## Verdict

`approved_with_changes`. The 23 algorithm functions, the four foundation primitives (AlgoID, Normalise, Tokenise, errors), and the Phase 8 Scorer make a coherent, idiomatic-Go surface. The progressive-disclosure model holds: Layer 1 hello-world is 4 lines, Layer 2a is 5 lines, Layer 2b is 6 lines. The naming convention `XxxScore` / `XxxDistance` / `XxxCode` is applied consistently across the 23 algorithms. The `AlgoID` enum gives IDE-discoverability and compile-time safety.

Six items are Critical (must change before v1.0). Eighteen are Important (resolve before v1.0 freeze). Twenty-one are Improvement nits.

The **SPEC OVERRIDE** of `ScoreAll` returning `map[AlgoID]float64` (vs `map[string]float64` in `docs/requirements.md` §8.3) is **affirmed**: it is the right call. AlgoID is the typed key the rest of the library exposes; consumers wanting string display call `AlgoID.String()`. This override is the documented decision in the existing 08-API-ERGONOMICS-REVIEW.md and remains the recommendation here.

Phase 8 BLOCKING findings from 08-API-ERGONOMICS-REVIEW.md (`WithThreshold` accepting NaN; `WithTverskyAlgorithm` accepting α=β=0) are inherited into this review as Critical API-T-01 and API-T-02 — both are still open in the codebase as of `scorer_options.go:257-266` and `scorer_options.go:381-399`.

---

## Critical

### [Critical] WithThreshold accepts NaN (silent never-match Scorer)
- **File:** `/Users/johnny/Development/fuzzymatch/scorer_options.go:257-266`
- **Phase introduced:** Phase 8
- **Issue:** `if t < 0.0 || t > 1.0` does not reject `math.NaN()` — NaN passes both inequalities. A Scorer constructed with `WithThreshold(math.NaN())` never matches anything because `... >= NaN` is always false. The Scorer's documented contract (`docs/scorer.md`) and the godoc on `ErrInvalidThreshold` both state NaN is rejected; the implementation does not match. This is the original BLOCKING API-01 from 08-API-ERGONOMICS-REVIEW.md and remains open.
- **Standard:** `go-coding-standards/SKILL.md` "fail loudly at construction"; Principle 1 (Pit of Success).
- **Action:** Code fix — add `math.IsNaN(t)` guard before the range check; return `ErrInvalidThreshold` with the same sentinel.
- **Rationale:** A silent-no-match Scorer is the worst-class API failure mode — caller code does the wrong thing forever with no diagnostic. Must reject at the option-application boundary.

### [Critical] WithTverskyAlgorithm accepts α=β=0 (panic at Score time, not construction time)
- **File:** `/Users/johnny/Development/fuzzymatch/scorer_options.go:381-399`
- **Phase introduced:** Phase 8
- **Issue:** `if alpha < 0 || beta < 0` rejects negatives but accepts `(0, 0)`. `TverskyScore` then panics at `Score` time (the algorithm guards α+β > 0 by panic). Every other `With*Algorithm` validates fully at option-application time. The deferral is inconsistent with the surface; the godoc admission that "typical use is satisfied" is not a defence. Original BLOCKING API-02 from 08-API-ERGONOMICS-REVIEW.md; still open.
- **Standard:** `go-coding-standards/SKILL.md` "validation at the option boundary"; Principle 4 (Idiomatic Go: explicit error at config boundary).
- **Action:** Code fix — add `alpha+beta == 0` check returning `ErrInvalidTverskyParam`; also handle NaN α/β here for consistency with API-T-01.
- **Rationale:** Construction-time errors are typed and discoverable; Score-time panics are not. The Scorer contract says "validated at construction"; this option breaks it.

### [Critical] Direct algorithm calls panic on bad parameters instead of returning errors
- **File:** `/Users/johnny/Development/fuzzymatch/tversky.go:230` (`TverskyScore`), `/Users/johnny/Development/fuzzymatch/qgram_jaccard.go:144` (`QGramJaccardScore`), `/Users/johnny/Development/fuzzymatch/cosine.go:195` (`CosineScore`), `/Users/johnny/Development/fuzzymatch/sorensen_dice.go:158` (`SorensenDiceScore`), and the four corresponding `*Runes` variants.
- **Phase introduced:** Phase 5
- **Issue:** Direct calls to parameterised q-gram algorithms with `n < 1` (or α+β=0 for Tversky) **panic** rather than return an error or score. CONTEXT.md §5 LOCKED this as "fail loudly on programmer error" but the project's other algorithms with parameter validation (SWG, Monge-Elkan) handle bad input differently: SWG produces "deterministic-but-meaningless" output, Monge-Elkan panics only for the inner-AlgoID allow-list. The panic-for-`n<1` policy creates an asymmetry across the algorithm surface: programmer error in `LevenshteinScore("foo", "bar")` never panics, but `QGramJaccardScore("foo", "bar", 0)` does.
- **Standard:** `go-coding-standards/SKILL.md` — algorithm score functions should never panic on caller-supplied data; only on programmer-clearly-wrong invariant breaks.
- **Action:** Discuss-phase needed — either (a) clamp `n` to `max(1, n)` and document, or (b) return 0.0 (the catalogue convention for "no comparison possible"), or (c) document the panic as a hard contract and add a Vet-friendly note in the godoc. The current panic-without-test-fixture (no consumer can write `defer recover()` cleanly without an exported sentinel for `errors.Is(panicValue, ErrInvalidQGramSize)`) is the worst combination.
- **Rationale:** Library-wide consistency. The Scorer layer correctly returns the typed error; the direct-call layer should match or, at minimum, expose a way for `recover()` to discriminate the panic value.

### [Critical] Scorer.Score does not validate input UTF-8 — SWG with NaN params and pathological inputs can return NaN
- **File:** `/Users/johnny/Development/fuzzymatch/scorer.go:349` (`Score`), `/Users/johnny/Development/fuzzymatch/swg.go:174` (`SmithWatermanGotohScoreWithParams`)
- **Phase introduced:** Phases 3 + 8
- **Issue:** A Scorer composed with `WithSmithWatermanGotohAlgorithm(weight, params)` where `params.Match = math.NaN()` (or Inf) accepts the option (no validation per `scorer_options.go:465-479`) and propagates NaN into the composite Score. The composite then violates the documented `[0.0, 1.0]` range guarantee silently. Documented as "nonsense values produce a deterministic-but-meaningless score" but NaN is not deterministic and not meaningless — it's a poisoned float that defeats `Match`'s `>=` comparison.
- **Standard:** `determinism-standards/SKILL.md` NaN/Inf rule (`§5 LOCKED`).
- **Action:** Code fix — `WithSmithWatermanGotohAlgorithm` should reject `math.IsNaN(...)` or `math.IsInf(...)` on any of `params.Match/Mismatch/GapOpen/GapExtend` and return `ErrInvalidConfiguration` (already exists). Alternatively, the SWG kernel should detect NaN inputs and short-circuit to `0.0`, but option-time rejection is the cleaner contract.
- **Rationale:** The Scorer's `[0.0, 1.0]` range guarantee is load-bearing — `Match` semantics depend on it. NaN admission breaks the contract silently.

### [Critical] WriteGoldenFile is publicly exported but is a test-only API
- **File:** `/Users/johnny/Development/fuzzymatch/golden_canonical.go:88`
- **Phase introduced:** Phase 1 (foundation)
- **Issue:** `func WriteGoldenFile(path string, v any) error` is exported in the production-binary code path, not in `_test.go` files or under a build tag. Consumers see it on pkg.go.dev, but its godoc says "intended for test maintenance only — production code never invokes it." This is the classic "exported because internal tests need it" trap.
- **Standard:** `go-coding-standards/SKILL.md` — public API surface is minimal; test-only helpers live in `_test.go` files or behind build tags.
- **Action:** Code fix — move `WriteGoldenFile` into `golden_canonical_test.go` (or add a `//go:build testfixtures` tag), or rename to `_writeGoldenFile` and provide an `export_test.go` re-export `var WriteGoldenFileForTest = writeGoldenFile`.
- **Rationale:** Every public symbol is a v1.x contract. A test-maintenance helper in the public surface is a permanent compatibility liability — any future Go version that breaks `encoding/json.MarshalIndent` stability or any future canonical-form change becomes a v2.0 breaking change.

### [Critical] DefaultScorer panics on internal inconsistency rather than returning a typed value
- **File:** `/Users/johnny/Development/fuzzymatch/scorer.go:586-592`
- **Phase introduced:** Phase 8
- **Issue:** `DefaultScorer` calls `panic("fuzzymatch: DefaultScorer construction failed (this is a bug): " + err.Error())`. The intent is to make the unreachable branch loud, but this means a future refactor that accidentally introduces a dispatch-table gap, a missing AlgoID, or a `WithThreshold` regression panics at consumer import time (the example uses `var defaultScorer = fuzzymatch.DefaultScorer()` in `examples/identifier-similarity/main.go:60` and `examples/scorer-composition/main.go:68`). A consumer who does the same `var x = DefaultScorer()` gets a startup panic with no `recover()` opportunity.
- **Standard:** `go-coding-standards/SKILL.md` — library code panics only on programmer errors with no possible runtime cause.
- **Action:** Code fix — either (a) replace panic with a build-time assertion (a `var _ = func() bool { if _, err := NewScorer(DefaultScorerOptions()...); err != nil { panic(...) }; return true }()` — this fires at package load, not at `DefaultScorer()` call time, where a static-analysis tool could catch the regression), or (b) accept the panic but ensure a regression test asserts `DefaultScorer()` never panics in CI under all build configurations. The current state is a latent failure: a code review that removes one dispatch entry breaks `DefaultScorer` without any test catching it before release.
- **Rationale:** A no-fail constructor that panics is a contradiction. Either move the panic to package-init (so the regression cannot ship) or remove it (with full test coverage that the static composition is valid).

---

## Important

### [Important] Asymmetric Monge-Elkan exposed publicly produces direction-sensitive scores by default
- **File:** `/Users/johnny/Development/fuzzymatch/monge_elkan.go:377`
- **Phase introduced:** Phase 6
- **Issue:** `MongeElkanScore(a, b, inner, opts)` is the **asymmetric** variant — `MongeElkanScore(a, b) != MongeElkanScore(b, a)` in general. The library's standard convention is symmetric similarity (every other algorithm in the catalogue satisfies `Score(a,b) == Score(b,a)`). Consumers used to symmetric algorithms will write `MongeElkanScore(query, candidate, ...)` and get a different score than `MongeElkanScore(candidate, query, ...)` — a subtle and dangerous footgun. The Scorer dispatch wraps `MongeElkanScoreSymmetric`, so the Scorer path is safe, but the direct surface is not.
- **Standard:** `algorithm-correctness-standards/SKILL.md` — symmetric property is the catalogue's default invariant.
- **Action:** Code fix — rename the current asymmetric surface to `MongeElkanScoreAsymmetric` and expose the symmetric form as `MongeElkanScore`. Alternatively, deprecate `MongeElkanScore` and document `MongeElkanScoreSymmetric` as the canonical entry point. The current naming inverts the principle of least surprise.
- **Suggested fix:** `MongeElkanScore` → symmetric; `MongeElkanScoreAsymmetric` (or `MongeElkanScoreDirectional`) → current asymmetric.

### [Important] MRACompare returns `(matched bool, simScore int)` — non-idiomatic Go return tuple
- **File:** `/Users/johnny/Development/fuzzymatch/mra.go:241`
- **Phase introduced:** Phase 7
- **Issue:** `MRACompare(a, b) (matched bool, simScore int)` — a `(bool, int)` tuple with named results is awkwardly close to the `(value, ok)` idiom but inverts the order (`matched` first, not last). Most Go consumers expect `comma-ok` form with the data value first and the boolean status last. Worse, the function's primary purpose (compute MRA score) is buried as the second return; the boolean is a derived predicate that consumers can compute themselves via `simScore >= mraThreshold(len(a)+len(b))`.
- **Standard:** `go-coding-standards/SKILL.md` — return tuples follow `(value, ok)` or `(value, error)` ordering.
- **Action:** Code fix — either (a) change to `MRACompare(a, b) (simScore int, matched bool)` to match `(value, status)` convention; or (b) keep the current signature and accept the non-idiomatic order; or (c) provide a separate `MRASimilarity(a, b) int` returning just the integer counter and remove the bool from `MRACompare`.
- **Suggested fix:** Renaming to `MRASimilarity(a, b) int` for the integer counter; keep `MRACompare(a, b) bool` for the boolean. Two functions, single-purpose each, idiomatic.

### [Important] Strcmp95Score has no rune variant — silently bytewise on Unicode input
- **File:** `/Users/johnny/Development/fuzzymatch/strcmp95.go:260`
- **Phase introduced:** Phase 4
- **Issue:** The catalogue convention is `XxxScore` (byte) + `XxxScoreRunes` (rune-aware) for every character-based algorithm. `Strcmp95Score` ships only the byte variant; the file header documents "ASCII-only — no `*Runes` variant" with the rationale that the similar-character table is upper-case ASCII letters only. This is defensible but inconsistent with the catalogue's surface promise — a consumer who has memorised `LevenshteinScore` / `LevenshteinScoreRunes` writes `Strcmp95ScoreRunes` and hits a compile error with no guidance toward the byte variant.
- **Standard:** `go-coding-standards/SKILL.md` — surface uniformity across algorithm categories.
- **Action:** Discuss-phase needed — either (a) add `Strcmp95ScoreRunes` that delegates to `Strcmp95Score` after `[]rune` conversion (with a doc note that the similar-character credit pass operates byte-wise so the rune surface is "rune-counted similarity, byte-compared content"), or (b) keep the gap and add a `Strcmp95ScoreRunes` placeholder that fails-compile with a useful error message via build-tag tricks (too clever — reject), or (c) update llms.txt and the README with an explicit note that Strcmp95 is byte-only by design.
- **Rationale:** Discoverability. A consumer iterating through the catalogue should not hit a single-algorithm exception without finding the rationale on pkg.go.dev within 2 clicks.

### [Important] SoundexScore / DoubleMetaphoneScore / NYSIISScore / MRAScore have no Runes variant
- **File:** `/Users/johnny/Development/fuzzymatch/soundex.go:257`, `/Users/johnny/Development/fuzzymatch/double_metaphone.go`, `/Users/johnny/Development/fuzzymatch/nysiis.go:350`, `/Users/johnny/Development/fuzzymatch/mra.go:348`
- **Phase introduced:** Phase 7
- **Issue:** Same as Strcmp95 — phonetic algorithms are ASCII-only by their primary-source definition (Russell 1918, Knuth 1973 for Soundex; Philips 2000 for Double Metaphone; Taft 1970 for NYSIIS; NBS 1977 for MRA). The non-ASCII input handling is documented per-algorithm ("non-ASCII runes are dropped silently"), but the API surface lacks a `*Runes` variant — same uniformity gap as Strcmp95.
- **Standard:** `go-coding-standards/SKILL.md` surface uniformity.
- **Action:** Discuss-phase needed — these algorithms are genuinely ASCII-only at the algorithm level (Soundex's "BFPV → 1" table is letter-pattern-specific). Recommend: keep no Runes variant for phonetic and document the absence in a single shared paragraph at the top of `llms.txt` and `docs/algorithms.md#phonetic-tier`.

### [Important] TokenSortRatioScore / TokenSetRatioScore / TokenJaccardScore have no Runes variant
- **File:** `/Users/johnny/Development/fuzzymatch/token_sort_ratio.go:183`, `/Users/johnny/Development/fuzzymatch/token_set_ratio.go:281`, `/Users/johnny/Development/fuzzymatch/token_jaccard.go:200`
- **Phase introduced:** Phase 6
- **Issue:** These algorithms tokenise via `Tokenise` (which is rune-aware) and then operate on the tokenised strings byte-wise. The rationale documented in `token_sort_ratio.go:181-182` is "Tokenise is UTF-8-aware so the rune semantic is already preserved." Defensible — but `PartialRatioScore` provides a `*Runes` variant despite operating on the same tokenised representation, which creates an internal inconsistency. Consumer reading the catalogue sees `PartialRatioScoreRunes` exists and expects `TokenSortRatioScoreRunes` to exist too.
- **Standard:** `go-coding-standards/SKILL.md` surface uniformity.
- **Action:** Discuss-phase needed — either add `*Runes` variants to all three (delegating after `[]rune` conversion), or remove `PartialRatioScoreRunes` for symmetry, or document the asymmetry prominently. Recommended: remove `PartialRatioScoreRunes`'s asymmetry and instead document why all token-tier algorithms are byte-clean once tokenised.

### [Important] MongeElkanScore takes `opts NormalisationOptions` parameter that is intentionally unused
- **File:** `/Users/johnny/Development/fuzzymatch/monge_elkan.go:377` and `:467`
- **Phase introduced:** Phase 6
- **Issue:** `MongeElkanScore(a, b string, inner AlgoID, opts NormalisationOptions) float64` accepts `opts` and immediately discards it via `_ = opts`. The godoc says "accepted for forward-compatibility with the Phase 8 Scorer option WithMongeElkanAlgorithm." This is a four-parameter signature where one parameter is documented as inert — a guaranteed footgun for any consumer who passes `DefaultNormalisationOptions()` expecting it to actually normalise.
- **Standard:** `go-coding-standards/SKILL.md` "no inert parameters in public API"; Principle 2 (Make Simple Things Simple).
- **Action:** Code fix — remove `opts` from the signature. The Phase 8 Scorer's `WithMongeElkanAlgorithm` can pass `DefaultNormalisationOptions()` internally if forward-compat is genuinely needed, but the public direct-call surface should not carry an inert parameter. If consumers genuinely need to control Tokenise options, expose a `MongeElkanScoreWithOptions(a, b string, inner AlgoID, tok TokeniseOptions) float64` variant instead.
- **Rationale:** Principle 1 (Pit of Success) — the obvious way to use the function must be the right way. Passing inert `opts` is a footgun.

### [Important] AlgoID.String returns CamelCase rather than the requirements-doc-specified snake_case
- **File:** `/Users/johnny/Development/fuzzymatch/algoid.go:213-267`
- **Phase introduced:** Phase 1
- **Issue:** The current `String()` returns `"Levenshtein"`, `"DamerauLevenshteinOSA"`, `"LCSStr"`, `"NYSIIS"`, etc. The reviewer brief (instruction) says: `AlgoID.String() returns snake_case: "levenshtein", "jaro_winkler", "damerau_levenshtein_osa"`. The current implementation diverges from this convention.
- **Standard:** `api-ergonomics-reviewer` Principle 5 (Naming Consistency).
- **Action:** Discuss-phase needed — the project locked CamelCase in CONTEXT.md and updated llms.txt, llms-full.txt, docs/scorer.md to match. Two valid paths: (a) accept the deviation from the brief and document it (the project lead chose CamelCase for v1.x display labels); (b) revert to snake_case to match the brief. Recommended: accept current CamelCase; it matches Go's idiomatic enum-name → string convention (e.g., `time.Sunday.String()` returns `"Sunday"`, not `"sunday"`). Update the brief to reflect the locked decision.
- **Suggested fix:** No code change. Update the agent brief (`.claude/agents/api-ergonomics-reviewer.md`) to record CamelCase as the locked convention.

### [Important] AlgoIDs() allocates a fresh slice on every call but is not documented as a hot-path-unsafe helper
- **File:** `/Users/johnny/Development/fuzzymatch/algoid.go:282-308`
- **Phase introduced:** Phase 1
- **Issue:** `AlgoIDs()` allocates a 23-element slice on every call. The godoc says "freshly allocated on every call so the caller may freely mutate" — correct but a consumer iterating algorithms in a loop (`for _, id := range fuzzymatch.AlgoIDs() { ... }`) allocates per-call. The pattern in `scorer.go:253` (NewScorer internal use) and the example program both call it once and cache; consumers who write the obvious loop pay a hidden allocation tax.
- **Standard:** `performance-standards/SKILL.md` — allocation surprises in obvious-looking code.
- **Action:** Code fix or godoc fix — either (a) add a sibling `AlgoIDsInto(dst []AlgoID) []AlgoID` zero-alloc variant for hot paths; or (b) expose the underlying ordering via `[NumAlgorithms]AlgoID{...}` as an exported value (mutable-by-design — but indexed access is zero-alloc); or (c) update the godoc to recommend single-call caching in hot loops.
- **Suggested fix:** Add to godoc: "For hot-path iteration, cache the result once: `var allAlgos = fuzzymatch.AlgoIDs()`."

### [Important] DispatchEntryNilForTest / DispatchInvokeForTest / DispatchLenForTest exported in production binary
- **File:** `/Users/johnny/Development/fuzzymatch/export_test.go:38-73`
- **Phase introduced:** Phase 1
- **Issue:** `export_test.go` is, despite its name, a build-tag-free `_test.go` file whose symbols **are not** in the production binary — Go's `_test.go` convention makes these test-only. This is correct usage. However, the same file exports `CanonicalMarshalForTest`, `NumAlgorithmsForTest`, `WinklerPrefixScaleForTest`, `Strcmp95SimilarCharsLenForTest`, etc. — fifteen test-only re-exports. These cannot be reached by production consumers, so this is not a public-API leak. False alarm — flagging only for documentation clarity.
- **Standard:** `go-coding-standards/SKILL.md` — `_test.go` files compile only under `go test`.
- **Action:** No action — confirmed correct. Mentioning for review-record completeness.
- **Rationale:** Verified `grep -L '// +build\|//go:build' export_test.go` plus the `_test.go` suffix gates this file from production builds. The 15 re-exports remain test-only.

### [Important] Scorer struct has no public fields but doc/scorer.md shows `ScorerAlgorithm{ID, Weight}` literal in examples
- **File:** `/Users/johnny/Development/fuzzymatch/scorer.go:414` (`ScorerAlgorithm`)
- **Phase introduced:** Phase 8
- **Issue:** `ScorerAlgorithm` is a pure data struct with two exported fields (`ID`, `Weight`). Documented as "consumers must not rely on ScorerAlgorithm having a stable memory address" — but the struct **is** the API contract. A future minor release that adds a `Threshold` field, a `Disabled` bool, or a `Source string` to `ScorerAlgorithm` is a breaking change for any consumer using struct-literal construction (`ScorerAlgorithm{ID: x, Weight: 0.5}`). Recommendation: hide the constructor — provide only the `Algorithms()` reader, never let consumers build a `ScorerAlgorithm` directly.
- **Standard:** `go-coding-standards/SKILL.md` — exported structs with all-exported fields are evolution-unfriendly.
- **Action:** Discuss-phase — either (a) accept the field-evolution liability (the struct is small and stable; unlikely to grow), or (b) keyed struct-literal initialisation is enforced via `//nolint:exhaustruct` configuration, or (c) add an unexported sentinel field so struct-literal init forces consumers through a constructor `NewScorerAlgorithm(id AlgoID, weight float64) ScorerAlgorithm`. Recommendation: (a) accept — the struct is descriptive metadata, evolution is unlikely.

### [Important] ScorerOption is a function type referencing the unexported `*scorerConfig`
- **File:** `/Users/johnny/Development/fuzzymatch/scorer_options.go:65`
- **Phase introduced:** Phase 8
- **Issue:** `type ScorerOption func(*scorerConfig) error` — `ScorerOption` is exported, but its sole parameter `*scorerConfig` is unexported. Consumers cannot write their own `ScorerOption` because the function body cannot reference the unexported type. This is intentional (it locks the option set to the library's With* helpers) but should be documented explicitly: "ScorerOption values are only constructible via fuzzymatch's With* helpers; custom options are not supported."
- **Standard:** `go-coding-standards/SKILL.md` — opaque types should document opacity.
- **Action:** Godoc fix — add a sentence to the `ScorerOption` godoc: "ScorerOption is opaque: consumers compose Scorers via the library's With*Algorithm / WithThreshold / WithNormalisation options. Building a ScorerOption with a custom function body is not supported (the parameter type is unexported)."
- **Rationale:** Discoverability — a consumer reading the godoc should not have to infer opacity from compile errors.

### [Important] Scorer.Threshold returns float64 with no NaN/Inf guarantee post-Critical fix
- **File:** `/Users/johnny/Development/fuzzymatch/scorer.go:441-443`
- **Phase introduced:** Phase 8
- **Issue:** Once API-T-01 is fixed (NaN rejected at `WithThreshold`), `Threshold()` is guaranteed in `[0.0, 1.0]`. The godoc currently does not state this guarantee; consumers cannot rely on it. Documentation gap, not a code bug.
- **Standard:** `documentation-standards/SKILL.md` — invariants on returns are godoc-explicit.
- **Action:** Godoc fix — append "The returned value is in [0.0, 1.0]: NewScorer's validation pipeline rejected any out-of-range or NaN threshold at construction." (the current godoc says the first part; add the NaN note after API-T-01 fix.)
- **Rationale:** Once the contract is bulletproof, the godoc should advertise it.

### [Important] NewSWGParams returns a value type with mutable fields and no validation
- **File:** `/Users/johnny/Development/fuzzymatch/swg.go:127`
- **Phase introduced:** Phase 3
- **Issue:** `NewSWGParams()` returns `SWGParams{Match: 1.0, Mismatch: -1.0, GapOpen: -1.5, GapExtend: -0.5}`. Consumer mutates freely. The godoc warns "callers may pass nonsense values" but the lack of validation in `SmithWatermanGotohScoreWithParams` means NaN/Inf flow through and break determinism (see Critical SWG-NaN finding). Validation could happen in `NewSWGParams` (eager check) or at score time (lazy check) — currently neither.
- **Standard:** `algorithm-correctness-standards/SKILL.md`; `determinism-standards/SKILL.md` NaN/Inf rule.
- **Action:** Code fix (combine with Critical SWG-NaN) — add either an `NewSWGParamsValidated(match, mismatch, gapOpen, gapExtend float64) (SWGParams, error)` constructor that rejects NaN/Inf and sign-mismatched params (Mismatch > 0, GapOpen > 0, GapExtend > 0), or document `NewSWGParams()` as the only valid construction path and reject mutated NaN/Inf at `SmithWatermanGotohScoreWithParams` boundary.
- **Suggested fix:** Reject NaN/Inf at `WithSmithWatermanGotohAlgorithm` (covers the Scorer surface) and in `SmithWatermanGotohScoreWithParams` direct-call (covers the algorithm surface) — both return 0.0 silently with a future-considered upgrade to return an explicit error if the algorithm surface migrates to `(float64, error)`.

### [Important] WithoutAlgorithm silently no-ops on absent AlgoID — no way to verify removal succeeded
- **File:** `/Users/johnny/Development/fuzzymatch/scorer_options.go:184-199`
- **Phase introduced:** Phase 8
- **Issue:** `WithoutAlgorithm(AlgoX)` filters entries; if no match, the slice is unchanged. The godoc states this is by design ("the no-op-on-absent semantic enables composition patterns"). However, a consumer who **expects** removal (e.g., they think `DefaultScorerOptions()` includes `AlgoLevenshtein` but it doesn't) gets silent failure. There is no `MustWithoutAlgorithm` or option-return-with-removed-count helper.
- **Standard:** `go-coding-standards/SKILL.md` — explicit failure is preferred unless silence is documented as desirable.
- **Action:** Godoc fix — strengthen the godoc with an explicit "If you need to verify the algorithm was actually present before removal, call `Scorer.Algorithms()` after construction and compare lengths." Code fix optional: provide `Scorer.HasAlgorithm(AlgoID) bool` for pre-removal verification on a constructed Scorer.
- **Rationale:** Silent no-op is a footgun for typo-ridden custom compositions ("WithoutAlgorithm(AlgoLeveshtein)" — misspelled — silently does nothing).

### [Important] ErrInvalidConfiguration is declared but never returned by any current code path
- **File:** `/Users/johnny/Development/fuzzymatch/errors.go:57`
- **Phase introduced:** Phase 1 (declared); never wired
- **Issue:** Searching the codebase: `ErrInvalidConfiguration` is referenced only in errors.go itself, llms-full.txt, and a couple of tests. No production code returns it. The Scorer surface returns `ErrInvalidWeight`, `ErrInvalidThreshold`, `ErrMissingThreshold`, `ErrEmptyScorer`, `ErrInvalidAlgorithm`, `ErrInvalidQGramSize`, `ErrInvalidTverskyParam` — all specific. `ErrInvalidConfiguration` is the only "umbrella" sentinel and is dead code.
- **Standard:** `go-coding-standards/SKILL.md` — unused public symbols increase API surface without value.
- **Action:** Discuss-phase needed — either (a) remove `ErrInvalidConfiguration` before v1.0 (it's not used; CHANGELOG note "Removed unused sentinel"), or (b) wire it as a parent error (`errors.Join(ErrInvalidConfiguration, ErrInvalidWeight)`) so `errors.Is(err, ErrInvalidConfiguration)` catches every Scorer-config error in one check. Recommended: (b) — the umbrella error is genuinely useful for consumers who want a single catch-all, and the join cost is one allocation at error-construction time.

### [Important] ErrEmptyInput is declared but never returned
- **File:** `/Users/johnny/Development/fuzzymatch/errors.go:106`
- **Phase introduced:** Phase 1 (declared); never wired
- **Issue:** Same shape as `ErrInvalidConfiguration` — the godoc says "Individual algorithm score functions handle empty inputs per their per-algorithm specification… and do NOT return this error — only higher-level APIs (Scorer, Extract) that require non-degenerate input may surface it." Phase 8 Scorer does NOT return it (both-empty returns 1.0 via identity short-circuit). Phase 10 Extract is not yet implemented.
- **Standard:** `go-coding-standards/SKILL.md`.
- **Action:** Discuss-phase needed — remove before v1.0 unless Phase 10 will use it. If Phase 10 plans use it, leave the declaration and note "reserved for Phase 10 Extract surface" in the godoc.

### [Important] ErrInvalidInput is declared but never returned (also documented as rare)
- **File:** `/Users/johnny/Development/fuzzymatch/errors.go:48`
- **Phase introduced:** Phase 1 (declared); never wired
- **Issue:** Same as ErrEmptyInput. Godoc says "Most algorithms accept arbitrary bytes and do NOT return this error; the exceptions document their constraints in their own godoc." No algorithm in the codebase actually returns it.
- **Standard:** `go-coding-standards/SKILL.md`.
- **Action:** Discuss-phase needed — remove before v1.0, or document its forward-compat reservation. Recommendation: remove three unused sentinels (`ErrInvalidConfiguration`, `ErrInvalidInput`, `ErrEmptyInput`) before v1.0; re-add later if needed (additive change is non-breaking).

---

## Improvement

### [Improvement] AlgoIDs() return slice could be a sortable type for self-documenting use
- **File:** `/Users/johnny/Development/fuzzymatch/algoid.go:282`
- **Issue:** `[]AlgoID` is fine but consumers who want to compare two algorithm sets (e.g., for set-equality) write helpers. A `type AlgoIDSet []AlgoID` with `Contains(AlgoID) bool` would self-document.
- **Action:** No action — over-engineering for v1.0; revisit in v1.x.

### [Improvement] LevenshteinScore / LevenshteinDistance return-type inconsistency from text
- **File:** `/Users/johnny/Development/fuzzymatch/levenshtein.go:84` (`LevenshteinDistance int`) and `:155` (`LevenshteinScore float64`)
- **Issue:** `Distance` is `int`, `Score` is `float64`. Both correct; just noting the deliberate split — the brief expects `XxxScore(a, b) float64` and `XxxDistance(a, b) int`. Current matches.
- **Action:** None.

### [Improvement] HammingDistance unequal-length returns max(len) — non-standard Hamming definition
- **File:** `/Users/johnny/Development/fuzzymatch/hamming.go:69`
- **Issue:** Classical Hamming is defined only for equal-length strings (Hamming 1950 §1). Returning `max(len(a), len(b))` is a fuzzymatch convention to make `HammingScore` normalise to 0.0; the godoc explains. Documented; nothing to change but worth noting that academic readers will expect `HammingDistance("abc", "ab")` to return an error or panic, not silently return 3.
- **Action:** Godoc reinforcement — add a citation/aside: "This deviates from Hamming 1950's strict equal-length requirement; the project's locked policy (CONTEXT.md) is silent-max for ergonomic Score normalisation. Consumers needing the strict definition should length-check before calling."

### [Improvement] DoubleMetaphoneKeys returns `(primary, secondary string)` — named returns inconsistent with rest of catalogue
- **File:** `/Users/johnny/Development/fuzzymatch/double_metaphone.go`
- **Issue:** Named returns are uncommon in the rest of the catalogue. Consumers writing `p, s := DoubleMetaphoneKeys("Smith")` get IDE inlay hints showing "primary, secondary" — useful. Not a problem; just inconsistent with `MRACompare(a, b) (matched bool, simScore int)` which also uses named returns. Mixed style.
- **Action:** No action — named returns where the names add information are fine.

### [Improvement] SmithWatermanGotohRawScore returns unbounded float64 — no documented range
- **File:** `/Users/johnny/Development/fuzzymatch/swg.go:244`
- **Issue:** `*RawScore` variants return the unclamped raw alignment value; godoc explains the unbounded nature. Consumers may store this and compare to a threshold, then break when params change. Recommend: a worked example in godoc showing typical raw-score values for English-name pairs.
- **Action:** Godoc improvement.

### [Improvement] Six SWG public functions on the surface is a lot
- **File:** `/Users/johnny/Development/fuzzymatch/swg.go`
- **Issue:** `SmithWatermanGotohScore` / `SmithWatermanGotohScoreRunes` / `SmithWatermanGotohScoreWithParams` / `SmithWatermanGotohRawScore` / `SmithWatermanGotohRawScoreRunes` / `SmithWatermanGotohRawScoreWithParams` — six entry points for one algorithm. The Score/RawScore axis × byte/runes axis × default-params/custom-params axis = 6. The combinatorial growth is fine but the naming uses no shared suffix structure (Score-vs-RawScore is a totally different semantics swap from Runes vs no-Runes).
- **Action:** Documentation polish — add a clear table in `docs/algorithms.md#smith-waterman-gotoh` showing the 2×2×2 → 6 mapping (the 8th combinator `RawScoreWithParamsRunes` is omitted — that's a gap, see next).

### [Improvement] SWG lacks `SmithWatermanGotohRawScoreWithParamsRunes` (the 8th combinator)
- **File:** `/Users/johnny/Development/fuzzymatch/swg.go`
- **Issue:** Following the 2×2×2 logic: byte/rune × clamp/raw × default-params/custom-params = 8 functions. Current surface has 6. The two missing combinators are byte-WithParams-Raw (`SmithWatermanGotohRawScoreWithParams` — present) and rune-WithParams-Raw (`SmithWatermanGotohRawScoreWithParamsRunes` — MISSING).
- **Action:** Code fix — add `SmithWatermanGotohRawScoreWithParamsRunes(a, b string, params SWGParams) float64` for completeness.

### [Improvement] LongestCommonSubstring returns the substring; LCSStrScore returns the score — naming dissonance
- **File:** `/Users/johnny/Development/fuzzymatch/lcsstr.go:127` and `:203`
- **Issue:** `LongestCommonSubstring` is descriptive; `LCSStrScore` is an abbreviation matching the `AlgoLCSStr` enum. The byte/runes axis is split across two name conventions. Consider renaming `LCSStrScore` → `LongestCommonSubstringScore` for consistency.
- **Action:** Discuss-phase — current naming follows the AlgoID, which is right for dispatch but reads odd for direct consumers. Recommendation: keep `LCSStrScore` as the AlgoID-matched name and add a comment in godoc cross-referencing `LongestCommonSubstring` (the substring-extracting helper).

### [Improvement] DefaultScorer composition is opaque — consumer cannot discover it without reading docs/scorer.md
- **File:** `/Users/johnny/Development/fuzzymatch/scorer.go:586`
- **Issue:** `DefaultScorer()` returns six-algorithm composite; the godoc says so but the consumer can't introspect the choice. `Scorer.Algorithms()` returns the post-construction list, so a consumer can `for _, a := range DefaultScorer().Algorithms() { fmt.Println(a.ID) }` — but this is discovery-after-the-fact.
- **Action:** Godoc improvement — add the six-algorithm list as a literal table in the `DefaultScorer` godoc with the AlgoID values inline.

### [Improvement] DefaultScorerOptions includes seven options (six algos + threshold), Scorer godoc says "six algorithms"
- **File:** `/Users/johnny/Development/fuzzymatch/scorer.go:543-553`
- **Issue:** Pedantic clarity — `DefaultScorerOptions()` returns 7 options (6 algorithms + 1 threshold). The Scorer godoc reads "six algorithms" which is correct for algorithms but the count of options is seven. A consumer doing `len(DefaultScorerOptions())` gets 7.
- **Action:** Godoc nit — clarify "six algorithm options plus a threshold option" in the relevant comment block.

### [Improvement] Tversky parameter ordering `(weight, alpha, beta float64, n int)` vs `TverskyScore(a, b string, n int, alpha, beta float64)` is inconsistent
- **File:** `/Users/johnny/Development/fuzzymatch/scorer_options.go:381` and `/Users/johnny/Development/fuzzymatch/tversky.go:230`
- **Issue:** `WithTverskyAlgorithm(weight, alpha, beta float64, n int)` — n is last. `TverskyScore(a, b string, n int, alpha, beta float64)` — n is in the middle. The two call sites have mirror-image parameter orderings; a consumer who learned `(n, alpha, beta)` from `TverskyScore` writes `WithTverskyAlgorithm(weight, n, alpha, beta)` and gets a compile success (all float64) with a swapped semantic — silent miscompose.
- **Action:** Code fix — unify ordering. Recommend the WithTverskyAlgorithm signature: `WithTverskyAlgorithm(weight float64, n int, alpha, beta float64) ScorerOption` (matches TverskyScore's `n int, alpha, beta float64` suffix).
- **Suggested fix:** `WithTverskyAlgorithm(weight float64, n int, alpha, beta float64)`.

### [Improvement] WithCosineAlgorithm / WithSorensenDiceAlgorithm / WithQGramJaccardAlgorithm signature `(weight float64, n int)` is consistent and good
- **File:** `/Users/johnny/Development/fuzzymatch/scorer_options.go:300/325/350`
- **Issue:** Three q-gram options all use `(weight, n)`. Good. Note for completeness.
- **Action:** None.

### [Improvement] WithMongeElkanAlgorithm has `(weight float64, inner AlgoID)` — no opts parameter
- **File:** `/Users/johnny/Development/fuzzymatch/scorer_options.go:425`
- **Issue:** `WithMongeElkanAlgorithm` does NOT take a `NormalisationOptions` parameter, but `MongeElkanScore` direct-call does (the unused `opts` flagged earlier). When the unused-opts issue is fixed, this option will be naturally consistent. Note this here for traceability.
- **Action:** None — resolves with Important MongeElkanScore-opts fix.

### [Improvement] Scorer.Score godoc does not show example input/output values
- **File:** `/Users/johnny/Development/fuzzymatch/scorer.go:349`
- **Issue:** Long godoc, no quick "Score("foo", "foobar") returns 0.X" reference vector. Consumers reading pkg.go.dev want a number to anchor their understanding.
- **Action:** Godoc improvement — add a worked example with concrete float values.

### [Improvement] Scorer.ScoreAll godoc says "fresh map allocated on every call" but does not warn against hot-path use
- **File:** `/Users/johnny/Development/fuzzymatch/scorer.go:497`
- **Issue:** Godoc covers the allocation, but a consumer who writes `for ... { s.ScoreAll(a, b) }` may not realise the per-call map overhead matters until they benchmark. Recommend adding a quick "in hot loops, prefer Score()" note.
- **Action:** Godoc improvement — add "Hot-path callers should use Score; ScoreAll is for introspection and debugging."

### [Improvement] Scorer ScoreAll returns map[AlgoID]float64 — consumer can iterate to get keys but Algorithms() also returns ordered set
- **File:** `/Users/johnny/Development/fuzzymatch/scorer.go:497` and `:460`
- **Issue:** Two paths to the same information: `s.Algorithms()` returns sorted `[]ScorerAlgorithm`; `s.ScoreAll(a, b)` returns `map[AlgoID]float64`. Consumers iterating ScoreAll in deterministic order do `for _, a := range s.Algorithms() { score := result[a.ID] }`. This is the documented pattern but the godoc on `ScoreAll` could link directly to `Algorithms()` as the recommended sorting path.
- **Action:** Godoc improvement — cross-link the two methods.

### [Improvement] Scorer.Match is a pure wrapper; could be a function-typed property
- **File:** `/Users/johnny/Development/fuzzymatch/scorer.go:392`
- **Issue:** `Match` does `Score(a,b) >= threshold`. One-line wrapper. Acceptable as ergonomic sugar but adds a method to the surface for nothing the consumer couldn't write themselves.
- **Action:** Keep — the ergonomics gain (no exposed threshold field) is worth the line.

### [Improvement] examples/identifier-similarity/main.go uses `defer recover()` pattern via panic-recover
- **File:** `/Users/johnny/Development/fuzzymatch/examples/identifier-similarity/main.go`
- **Issue:** No `defer recover()` actually in the example, but a consumer running the example with edited pairs that include `nil` rune patterns might trigger an algorithm panic (e.g., `MongeElkanScore` with `AlgoMongeElkan` inner — see allow-list panic). The example should explicitly note panic-recovery in production code.
- **Action:** Documentation note in the example's godoc preamble: "Production code should `defer recover()` when calling algorithm functions with consumer-supplied AlgoIDs (Monge-Elkan inner) — see allow-list godoc on MongeElkanScore."

### [Improvement] examples/scorer-composition uses `_, err := NewScorer(...); if err != nil { panic(...) }` pattern
- **File:** `/Users/johnny/Development/fuzzymatch/examples/scorer-composition/main.go:89-104`
- **Issue:** The example correctly handles the error then panics as "unreachable." Good. Could be replaced with `errcheck.Must(NewScorer(...))` style if the library exposed a `MustNewScorer(opts ...ScorerOption) *Scorer` helper. Adding `MustNewScorer` would be idiomatic Go (compare `template.Must` in stdlib).
- **Action:** Code fix or skip — adding `MustNewScorer(opts ...) *Scorer` is small, low-risk, idiomatic, and removes ~5 lines of boilerplate from example code. Recommend: add it.

### [Improvement] No package-level constants for "common threshold values" (0.85, 0.80, 0.95)
- **File:** `/Users/johnny/Development/fuzzymatch/scorer.go`
- **Issue:** Consumers writing `WithThreshold(0.85)` repeat the magic number; `DefaultScorer` bakes in 0.85 but does not export it. A `const DefaultThreshold = 0.85` would let consumers reference the constant.
- **Action:** Code fix optional — export `const DefaultScorerThreshold = 0.85` so `WithThreshold(fuzzymatch.DefaultScorerThreshold)` is self-documenting.

### [Improvement] README's Quick start section still shows Phase 1 primitives, not the Phase 8 Scorer flow
- **File:** `/Users/johnny/Development/fuzzymatch/README.md:96-117`
- **Issue:** Phase 8 has landed (Scorer is in production code) but the README still shows `Normalise` + `Tokenise` as the quick-start example. A first-time visitor sees the wrong entry point.
- **Action:** Documentation fix — replace the quick start with a Phase 8-tier example (DefaultScorer or LevenshteinScore one-liner).

### [Improvement] llms.txt entry for Scorer section header is too long and references "plan 08-01/02/03"
- **File:** `/Users/johnny/Development/fuzzymatch/llms.txt:187`
- **Issue:** `### Scorer construction options (Phase 8 — plan 08-01 lays the option layer; plan 08-02 lands NewScorer + Score + Match; plan 08-03 lands ScoreAll + Threshold + Algorithms + ScorerAlgorithm + DefaultScorer + DefaultScorerOptions)` — the heading exposes internal planning phases to AI-assistant consumers. Should be clean: `### Scorer (Phase 8)`.
- **Action:** Documentation fix — trim the heading.

