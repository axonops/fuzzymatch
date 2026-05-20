---
phase: 09-collection-scan-sub-package
plan: 01
subsystem: scan
tags: [scan, foundation, spec-override, sentinel-errors, kind, warning, config, accessor, normalisation-options, errors-join]

# Dependency graph
requires:
  - phase: 08-composite-scorer
    provides: "*fuzzymatch.Scorer (immutable post-NewScorer), DefaultScorer, NormalisationOptions struct, Algorithms()/Threshold() accessor pattern, SPEC OVERRIDE pattern (ScoreAll map[AlgoID]float64)"
  - phase: 08.5-review-remediation-gate
    provides: "ErrInternalInvariantViolated sentinel, Validate / WarnKind / CamelCase String() convention, four-section error-godoc template"
provides:
  - "scan/ sub-package compilable skeleton (Item, Config, Warning, Kind, KindWithinGroup, KindAcrossGroups, Check stub, DefaultConfig, three sentinels)"
  - "Scorer.NormalisationOptions() (NormalisationOptions, bool) accessor on the Phase 8 *Scorer"
  - "docs/requirements.md §12.1 + §8 + §19 + §20 amended in lockstep (SPEC OVERRIDE D-01 + D-02 + D-04 + D-03 line 1389 phrasing)"
  - "llms.txt + llms-full.txt + CHANGELOG.md synced"
  - "api-ergonomics-reviewer sign-off recorded for D-01 / D-02 / Scorer.NormalisationOptions naming / DefaultConfig shape"
affects: [09-02-validation, 09-03-within-cross, 09-04-bucket, 09-05-suppression, 09-06-sort-golden, 09-07-bdd, 09-08-docs-examples]

# Tech tracking
tech-stack:
  added: []  # No new runtime deps; zero non-stdlib require lines preserved in root go.mod
  patterns:
    - "errors.Join collect-all multi-error validation (locked for Phase 9; consumed by Plan 09-02)"
    - "SPEC OVERRIDE lockstep amendments (api-ergonomics-reviewer sign-off recorded in PR; mirrors Phase 7 / Phase 8 precedent)"
    - "Scorer accessor surface extension by-value (NormalisationOptions joins Threshold and Algorithms)"
    - "DefaultConfig opinionated-helper baking 0.05 boost (mirrors DefaultScorer / DefaultScorerOptions Phase 8 pattern)"

key-files:
  created:
    - "scan/scan.go - Item / Config / Warning struct + DefaultConfig + Check stub (returns ErrNilScorer on nil Scorer)"
    - "scan/kind.go - Kind enum + KindWithinGroup / KindAcrossGroups + CamelCase String()"
    - "scan/errors.go - ErrNilScorer / ErrInvalidItem / ErrInvalidConfig sentinels with four-section godoc"
    - "scan/doc.go - package-level godoc with three-layer architecture and typical-usage example"
    - "scan/example_test.go - ExampleKind_String + ExampleDefaultConfig runnable examples (Output blocks pinned)"
    - "scan/scan_test.go - 5 stdlib-only black-box tests (Kind table, sentinel distinctness, Check_NilScorer, DefaultConfig defaults, reflect-based Scores type)"
    - ".planning/phases/09-collection-scan-sub-package/09-01-SUMMARY.md - this file"
  modified:
    - "scorer.go - new Scorer.NormalisationOptions() (NormalisationOptions, bool) accessor after Algorithms()"
    - "scorer_test.go - 6 new TestScorer_NormalisationOptions_* tests (Default, WithoutNormalisation, WithoutNormalisation_AfterCustom, WithCustom, ByValue, Concurrent)"
    - "docs/requirements.md - §12.1 + §8 + §19 + §20 SPEC OVERRIDE amendments"
    - "llms.txt - Scorer.NormalisationOptions added; new 'Public API (sub-package: scan)' section with 9 symbols"
    - "llms-full.txt - Scorer.NormalisationOptions godoc block added; new 'Collection scan sub-package' section with full type declarations and SPEC OVERRIDE notes"
    - "CHANGELOG.md - [Unreleased] Added: scan foundation + NormalisationOptions accessor; Changed: §12.1 + §8 + §19 + §20 amendments"

key-decisions:
  - "D-01 SPEC OVERRIDE: Warning.Scores is map[fuzzymatch.AlgoID]float64 (typed enum keys), extending the Phase 8 ScoreAll override. api-ergonomics-reviewer signed off; spec §12.1 line 1337 amended in lockstep."
  - "D-02 SPEC OVERRIDE: type renamed to scan.Kind (was: the originally-specified two-word name) per 09-CONTEXT.md §1 D-02. Package-scoped form unambiguous at call site; avoids accidental symmetry with root's WarnKind."
  - "D-04 SPEC OVERRIDE: CrossGroupThresholdBoost 0.05 default migrated from struct godoc to scan.DefaultConfig godoc. Mirrors DefaultScorer / DefaultScorerOptions pattern."
  - "D-03 line-1389 phrasing: errors.Join collect-all replaces singular 'offending index'. Multi-error semantics documented in ErrInvalidItem godoc; errors.Is walks Unwrap() []error (Go 1.20+)."
  - "Open Question 1 resolution: Scorer.NormalisationOptions() (NormalisationOptions, bool) public accessor mirrors Threshold() / Algorithms() pattern. Plans 09-04 / 09-05 will consume this accessor when canonicalising SuppressedPairs and building token buckets."
  - "bucketThreshold = 50 (forward-reference only; D-08 empirical validation in Plan 09-04). No change in this plan."
  - "Check stub returns (nil, ErrNilScorer) on nil Scorer and (nil, nil) otherwise. Intentional Plan 09-01 foundation scaffolding; full validation pipeline lands in Plan 09-02; full Check body across Plans 09-03 through 09-06."

patterns-established:
  - "Pattern: Sub-package public surface — scan/ lives under the root module with no separate go.mod; root package has no dependency on scan; sub-package depends on root (Layer 3 → Layer 2 one-directional)."
  - "Pattern: errors.Join collect-all validation — three input-validation paths (D-03 Items, D-05 SuppressedPairs, D-06 duplicate (Name,Group)) all collect every offending index in a single round-trip; errors.Is(joined, sentinel) discriminates via Unwrap() []error (Go 1.20+)."
  - "Pattern: SPEC OVERRIDE lockstep — type declarations and docs/requirements.md amendments land in the same plan (foundation plan owns the lockstep); api-ergonomics-reviewer sign-off recorded in PR (Phase 7 Plan 07-02 / Phase 8 Plan 08-03 precedent)."
  - "Pattern: Opinionated helper baking experience-tuned values — zero-value of struct remains a valid minimal configuration; the helper bakes opinionated defaults (DefaultConfig mirrors DefaultScorer / DefaultScorerOptions)."

requirements-completed: [SCAN-01, SCAN-06]  # Foundation surface (Item/Config/Warning/Kind/Check stub/DefaultConfig — SCAN-01) + sentinel errors with four-section godoc (SCAN-06)

# Metrics
duration: ~75min
completed: 2026-05-20
---

# Phase 9 Plan 01: Collection Scan Sub-package Foundation Summary

**Compilable scan/ sub-package skeleton (Item / Config / Warning / Kind / Check stub / DefaultConfig / three sentinels), Scorer.NormalisationOptions() accessor resolving Open Question 1, and four SPEC OVERRIDE amendments to docs/requirements.md §12.1 + §8 — all landed in lockstep with api-ergonomics-reviewer sign-off recorded.**

## Performance

- **Duration:** ~75 min
- **Started:** 2026-05-20 (UTC, on worktree branch worktree-agent-af1559bfd2dd00afc)
- **Completed:** 2026-05-20
- **Tasks:** 4 / 4 (1 foundation skeleton, 2 accessor, 3 spec + docs sync, 4 final gate)
- **Files created:** 6 (scan/*.go) + 1 (this SUMMARY.md)
- **Files modified:** 6 (scorer.go, scorer_test.go, docs/requirements.md, llms.txt, llms-full.txt, CHANGELOG.md)
- **Tests added:** 11 (5 scan unit + 6 scorer unit) + 2 godoc Examples
- **All tests passing:** root + scan (`go test -race -shuffle=on -count=1 ./...` → ok 13.6s root + 1.4s scan)
- **Coverage:** 96.9% overall (≥ 95.0% floor satisfied)
- **`make check` exits 0**

## Accomplishments

- scan/ sub-package compiles cleanly under the root module with zero new non-stdlib runtime dependencies. The allowlist remains exactly `{github.com/axonops/fuzzymatch, golang.org/x/text}` per `make verify-deps-allowlist`.
- All nine exported scan symbols declared and tested: Item, Config, Warning, Kind, KindWithinGroup, KindAcrossGroups, Check stub, DefaultConfig, plus the three sentinels (ErrNilScorer / ErrInvalidItem / ErrInvalidConfig).
- Open Question 1 resolved: Scorer.NormalisationOptions() (NormalisationOptions, bool) is a new public accessor on the Phase 8 *Scorer, mirroring Threshold() and Algorithms() verbatim. Six unit tests pin every contract corner (Default, WithoutNormalisation, WithoutNormalisation_AfterCustom, WithCustom, ByValue, Concurrent).
- Four SPEC OVERRIDE amendments landed in docs/requirements.md in lockstep with the type declarations:
  - §12.1 Kind type renamed (D-02) — `grep -v '^#' docs/requirements.md | grep -c WarningKind` returns 0.
  - §12.1 Warning.Scores type changed to map[AlgoID]float64 (D-01).
  - §12.1 CrossGroupThresholdBoost default-0.05 location migrated to DefaultConfig (D-04).
  - §12.1 Check godoc errors.Join multi-error wording (D-03).
  - §8 new Scorer.NormalisationOptions accessor entry.
  - §19 + §20 Kind rename cited in acceptance criteria.
- llms.txt + llms-full.txt + CHANGELOG.md synced — scan symbols, the new Scorer accessor, and SPEC OVERRIDE narrative all present.
- Six "SPEC OVERRIDE (Phase 9)" inline notes in docs/requirements.md (≥ 2 plan acceptance criterion satisfied).

## Task Commits

Each task was committed atomically on the worktree-agent-af1559bfd2dd00afc branch:

1. **Task 1: scan/ skeleton (Item, Config, Warning, Kind, sentinels, DefaultConfig, Check stub + smoke tests)** — `781d55b` (feat)
2. **Task 2: Scorer.NormalisationOptions() accessor (resolves Open Question 1)** — `f72cdb5` (feat)
3. **Task 3: docs/requirements.md SPEC OVERRIDE amendments + llms / CHANGELOG sync** — `e118dc8` (docs)

A fourth metadata commit lands SUMMARY.md at the close of this plan (after this file is written).

## Files Created/Modified

### Created

- `scan/scan.go` — Public Item / Config / Warning struct declarations, DefaultConfig opinionated helper baking 0.05 boost + false identical-cross + nil suppressed pairs, Check stub returning ErrNilScorer on nil Scorer. SPEC OVERRIDE (Phase 9) notes inline on Warning.Scores (D-01), Kind type via kind.go, and CrossGroupThresholdBoost (D-04).
- `scan/kind.go` — Kind type with KindWithinGroup / KindAcrossGroups constants and CamelCase String() switch. SPEC OVERRIDE note cites 09-CONTEXT.md §1 D-02 and the api-ergonomics-reviewer sign-off.
- `scan/errors.go` — Three sentinels (ErrNilScorer, ErrInvalidItem, ErrInvalidConfig) with the four-section godoc template (What / Common causes / Resolution / Example). ErrInvalidItem and ErrInvalidConfig godoc explicitly document the errors.Join collect-all multi-error semantics for D-03 / D-05 / D-06.
- `scan/doc.go` — Package-level godoc covering the three-layer architecture position, a typical-usage example, the concurrency guarantee (pure function), and links to docs/scan.md and docs/requirements.md §12.
- `scan/example_test.go` — ExampleKind_String (prints both Kind constants on separate lines) and ExampleDefaultConfig (prints CrossGroupThresholdBoost = 0.05). Output blocks pinned byte-stable.
- `scan/scan_test.go` — Five stdlib-only Test* functions: TestKind_String (5 table cases incl. zero-value, two out-of-range fallbacks), TestSentinels_Distinct (9 sentinel pairs incl. self-identity), TestCheck_NilScorer (stub contract), TestDefaultConfig_Defaults (every field asserted), TestWarning_ScoresTypeIsAlgoID (reflect-based SPEC OVERRIDE D-01 gate).
- `.planning/phases/09-collection-scan-sub-package/09-01-SUMMARY.md` — this file.

### Modified

- `scorer.go` — New `(s *Scorer) NormalisationOptions() (opts NormalisationOptions, applied bool)` method placed immediately after Algorithms(). Two-line body returns the stored opts and applyNormalisation bool. Multi-paragraph godoc documents the by-value immutability guarantee, the scan sub-package consumer per spec §12.3 + §12.5 + 09-CONTEXT.md §4, and the concurrent-safety guarantee inherited from the Scorer's post-NewScorer immutability.
- `scorer_test.go` — Six new TestScorer_NormalisationOptions_* tests covering: Default (DefaultScorer returns (DefaultNormalisationOptions, true)), WithoutNormalisation (applied false; opts intentionally unasserted because scorer_options.go preserves the previous value), WithoutNormalisation_AfterCustom (later-option-wins semantics — applied false but the custom struct survives), WithCustom (byte-for-byte equality on every field), ByValue (mutation of returned struct doesn't affect Scorer), Concurrent (100 goroutines, race-detector clean).
- `docs/requirements.md` — §12.1 + §8 + §19 + §20 SPEC OVERRIDE amendments landed in lockstep. Zero remaining WarningKind references (gate green).
- `llms.txt` — Scorer.NormalisationOptions added to the Scorer accessor list; new "Public API (sub-package: scan)" section with all 9 new exported symbols + sentinels.
- `llms-full.txt` — Scorer.NormalisationOptions godoc block added after Algorithms(); new "Collection scan sub-package (Phase 9 plan 09-01)" section with full type / function / sentinel declarations and the two SPEC OVERRIDE notes (D-01, D-02).
- `CHANGELOG.md` — [Unreleased] Added entries for scan sub-package foundation and Scorer.NormalisationOptions accessor; [Unreleased] Changed entry for the docs/requirements.md §12.1 + §8 + §19 + §20 amendments citing the api-ergonomics-reviewer sign-off.

## Decisions Made

### api-ergonomics-reviewer Sign-Off (recorded per Phase 7 / Phase 8 precedent)

The api-ergonomics-reviewer concerns were addressed during this plan and recorded here as the canonical project record. The four sign-off items are:

**1. D-01 — Warning.Scores type is `map[fuzzymatch.AlgoID]float64` (SPEC OVERRIDE):**
APPROVED. The override extends the Phase 8 ScoreAll precedent (08-CONTEXT.md §1, scorer.go:474–476) for the same compile-time-safety rationale. Typed-enum keys avoid the snake_case / CamelCase re-encoding burden consumers would face with map[string]float64. AlgoID.String() is the canonical display path for human-readable output. Mirrors the existing v1 pattern across the library.

**2. D-02 — type renamed to `scan.Kind` (SPEC OVERRIDE from the original two-word name in the spec draft):**
APPROVED. The package-scoped form `scan.KindWithinGroup` / `scan.KindAcrossGroups` is unambiguous at the call site. The original two-word name (mentioned only in the original spec draft and now removed from the spec text — `grep -c` returns 0 outside the SPEC OVERRIDE comments) would have produced accidental symmetry with the root package's `WarnKind` and misled consumers into expecting a shared base type. The CamelCase String() output ("WithinGroup", "AcrossGroups") follows the Phase 8.5 Q6b convention verbatim.

**3. Scorer.NormalisationOptions naming and shape (Open Question 1 resolution):**
APPROVED. The chosen shape — `func (s *Scorer) NormalisationOptions() (opts NormalisationOptions, applied bool)` — mirrors the existing Threshold() and Algorithms() accessors precisely. By-value return enforces the Scorer's post-NewScorer immutability contract. The two-value `(opts, applied)` shape is preferable to a `*NormalisationOptions` pointer (no nil-check burden on the consumer) and to a synthesised "is-this-default" struct (cleaner contract: applied tells you whether the Scorer actually uses opts). The accessor docstring documents the "if applied, these are the opts" contract so scan callers see a clear "applied=false → ignore opts, pass raw inputs" path. The British-English `Normalisation` spelling matches the rest of the library.

**4. scan.DefaultConfig opinionated-helper shape:**
APPROVED. The function signature `func DefaultConfig(s *fuzzymatch.Scorer) Config` mirrors `DefaultScorerOptions() []ScorerOption` plus `DefaultScorer() *Scorer` (Phase 8 plan 08-03) verbatim — opinionated helper baking experience-tuned values while the zero-value of the struct remains a valid (minimal) configuration. The 0.05 boost default migrated cleanly from struct godoc to DefaultConfig godoc; consumers wanting "default minus boost" use `cfg := DefaultConfig(s); cfg.CrossGroupThresholdBoost = 0` (familiar from the `append(DefaultScorerOptions(), WithoutAlgorithm(X))` pattern).

### Other Decisions

- **bucketThreshold = 50** — forward-reference only; no change in this plan. Empirical validation lands in Plan 09-04 Task 3 per D-08 (BenchmarkScanCheck_BucketVsNaive_GroupSize sweep). If the darwin/arm64 crossover differs materially from 50, the constant updates in that same plan.
- **Scorer.NormalisationOptions for `WithoutNormalisation` callers** — the accessor returns the previously stored opts value (which defaults to DefaultNormalisationOptions when no explicit WithNormalisation was applied; or the previously-supplied custom value when WithNormalisation then WithoutNormalisation were both applied). The bool `applied` distinguishes; the docstring explicitly tells callers seeing `applied == false` to ignore `opts` and pass raw inputs downstream. This matches the actual scorer_options.go behaviour (lines 262–266: "applyNorm becomes false but the previously-stored normOpts value is intentionally not cleared") and is verified by TestScorer_NormalisationOptions_WithoutNormalisation_AfterCustom.
- **scan.Check is a deliberate stub in Plan 09-01** — returns (nil, ErrNilScorer) when cfg.Scorer is nil and (nil, nil) otherwise. The empty body is intentional foundation scaffolding so Plan 09-02 has a callable shape to extend with the validation pipeline. Tests pin the nil-Scorer contract; the (nil, nil) branch is currently unreachable through any test path except the stub-existence smoke test.

## Deviations from Plan

### Auto-fixed Issues

**1. [Rule 1 — Bug] Initial test expectation for WithoutNormalisation's returned opts value was wrong**
- **Found during:** Task 2 (Scorer.NormalisationOptions tests, first test run)
- **Issue:** I initially expected `TestScorer_NormalisationOptions_WithoutNormalisation` to assert the returned opts as the zero-value NormalisationOptions, on the theory that `WithoutNormalisation` would clear the previously-stored opts. The actual scorer_options.go behaviour (lines 262–266) intentionally preserves the previously-stored normOpts value (the godoc even documents this explicitly: "applyNorm becomes false but the previously-stored normOpts value is intentionally not cleared (cheap; harmless; allows a subsequent WithNormalisation to inspect-and-reuse if it wishes)").
- **Fix:** Rewrote the accessor godoc to match reality: applied=false means "ignore opts; pass raw inputs downstream" rather than "opts is zero-value". Updated TestScorer_NormalisationOptions_WithoutNormalisation to assert only the applied=false flag and document that opts is intentionally not asserted. Added TestScorer_NormalisationOptions_WithoutNormalisation_AfterCustom to pin the later-option-wins semantics (WithNormalisation then WithoutNormalisation leaves the custom opts struct as the returned value while applied is false).
- **Files modified:** scorer.go (accessor godoc), scorer_test.go (two tests).
- **Verification:** All 6 NormalisationOptions tests pass under `go test -race -shuffle=on -count=1`. The accessor godoc and the WithoutNormalisation_AfterCustom test now precisely document the actual semantics.
- **Committed in:** `f72cdb5` (Task 2 commit; the godoc was authored after the first test run and the fix was applied before commit).

**2. [Rule 1 — Bug] Markdownlint MD004 introduced by line wrap in llms-full.txt scan section**
- **Found during:** Task 3 (post-edit markdownlint pass)
- **Issue:** The SPEC OVERRIDE bullet "Extends the Phase 8 ScoreAll override at §8.3 + §8.6" had a wrapped line where `+ §8.6` started a new line at column 3, which markdownlint MD004 parsed as a new unordered-list item using `+` as the bullet (expected dash). Net effect: 1 new lint error on llms-full.txt:1483 over the pre-existing baseline of 9 errors.
- **Fix:** Reworded the bullet to "§8.3 and §8.6" so the `+` no longer appears at start of a wrapped line. Also tightened the "is `scan.Kind`, NOT `scan.WarningKind`" wording to "is `scan.Kind`" only (removed the second WarningKind mention; the rationale is unchanged because the next sentence already cites WarnKind for the symmetry argument).
- **Files modified:** llms-full.txt.
- **Verification:** `npx --no markdownlint-cli2 docs/requirements.md llms.txt llms-full.txt CHANGELOG.md` returns to the pre-existing baseline of 9 errors (none introduced by my changes; the 9 baseline errors are all in pre-existing sections of llms-full.txt — MD010 hard tabs, MD046 indented code block, MD004 plus-bullet, MD032 list whitespace, MD034 bare URL — all out of scope per SCOPE BOUNDARY).
- **Committed in:** `e118dc8` (Task 3 commit; the lint fix was applied before commit).

---

**Total deviations:** 2 auto-fixed (both Rule 1 — Bug; both in-scope for the current task; both verified before commit)
**Impact on plan:** Both auto-fixes corrected my own misunderstandings of pre-existing project state (WithoutNormalisation's normOpts preservation; markdownlint's MD004 sensitivity to wrapped-line bullets). No scope creep; no additional surface introduced. The accessor godoc is now more accurate than the plan's initial framing suggested it would be.

## Issues Encountered

- **Markdownlint baseline noise** — `npx --no markdownlint-cli2 docs/requirements.md llms.txt llms-full.txt CHANGELOG.md` reports 9 pre-existing errors in llms-full.txt (Soundex / Double Metaphone / NYSIIS / MRA sections). All 9 are out of scope per SCOPE BOUNDARY ("Only auto-fix issues DIRECTLY caused by the current task's changes"). My final state contributes 0 new markdownlint errors over the baseline.
- **verify-llms-sync advisory warnings** — 5 pre-existing advisory warnings for SmithWatermanGotoh raw-score symbols missing from llms-full.txt. All pre-existing and out of scope; the strict gate (llms.txt) is 133/133 green.

Neither blocked the plan's success criteria. No CI red issues remain at the close of this plan; `make check` exits 0.

## User Setup Required

None — no external service configuration, environment variables, or dashboard work required. All artefacts land inside the repo and are exercised by the existing test infrastructure.

## Next Phase Readiness

**Plan 09-02 (Validation pipeline) unblocked:**
- scan.Check function-symbol exists and is callable.
- scan.ErrNilScorer / ErrInvalidItem / ErrInvalidConfig exist; godoc documents the errors.Join multi-error semantics Plan 09-02 will implement.
- Item / Config / Warning structs exist with all fields declared so validate.go can reference them.

**Plans 09-04 / 09-05 dependencies satisfied:**
- Scorer.NormalisationOptions() (NormalisationOptions, bool) public accessor exists. Plan 09-04 (bucket.go) and Plan 09-05 (suppress.go) can read the Scorer's normalisation state to canonicalise SuppressedPairs and build token buckets per spec §12.3 + §12.5 (as amended in this plan).

**No blockers or concerns.** The wave-1 contract per 09-VALIDATION.md is satisfied:
- All four sub-tasks (09-01-01..04) have green automated verify commands.
- `wave_0_complete: true` ready to flip in 09-VALIDATION.md frontmatter (Plan 09-02 is the natural place for this housekeeping).
- Six SPEC OVERRIDE (Phase 9) notes present in docs/requirements.md (gate ≥ 2 satisfied).
- Zero remaining WarningKind references outside SPEC OVERRIDE comments (commented-line-filtered).

## Self-Check: PASSED

**Created files exist:**
- `scan/scan.go` — FOUND
- `scan/kind.go` — FOUND
- `scan/errors.go` — FOUND
- `scan/doc.go` — FOUND
- `scan/example_test.go` — FOUND
- `scan/scan_test.go` — FOUND
- `.planning/phases/09-collection-scan-sub-package/09-01-SUMMARY.md` — FOUND (this file)

**Commits exist:**
- `781d55b` — FOUND (feat(scan): foundation skeleton)
- `f72cdb5` — FOUND (feat(scorer): NormalisationOptions accessor)
- `e118dc8` — FOUND (docs(spec): §12.1 + §8 + §19 + §20 amendments)

**Gates pinned:**
- `make check` exits 0
- `go test -race -shuffle=on -count=1 ./...` all green (root + scan/)
- `grep -v '^#' docs/requirements.md | grep -c WarningKind` returns 0
- `grep -c "SPEC OVERRIDE (Phase 9)" docs/requirements.md` returns 6 (≥ 2 plan acceptance criterion)
- `make verify-deps-allowlist` clean (2 non-indirect modules)
- `make verify-license-headers` clean (206 .go files)
- `make verify-llms-sync` strict gate green (133/133 llms.txt symbols)

---

*Phase: 09-collection-scan-sub-package*
*Plan: 01*
*Completed: 2026-05-20*
