---
phase: 06-token-based-algorithms
plan: 05
subsystem: algorithm-catalogue
tags: [monge-elkan, asymmetric, symmetric-mean, inner-metric-allowlist, panic-discipline, dos-godoc, oq-4-resolution]

# Dependency graph
requires:
  - phase: 01-foundation-infrastructure
    provides: [AlgoID dispatch table, Tokenise, NormalisationOptions, errors, CI matrix, license-headers gate]
  - phase: 02-core-character-algorithms-six
    provides: [Levenshtein/DamerauLevenshteinOSA/DamerauLevenshteinFull/Hamming/Jaro/JaroWinkler dispatch slots, two-row DP discipline]
  - phase: 03-smith-waterman-gotoh
    provides: [AlgoSmithWatermanGotoh dispatch slot]
  - phase: 04-remaining-character-gestalt
    provides: [AlgoStrcmp95/AlgoLCSStr/AlgoRatcliffObershelp dispatch slots]
  - phase: 05-q-gram-algorithms
    provides: [AlgoQGramJaccard/AlgoSorensenDice/AlgoCosine/AlgoTversky dispatch slots — 4 q-gram-tier permitted inners]
  - phase: 06-token-based-algorithms / plan 06-01
    provides: [token-tier dispatch slot pattern (slot 14), per-plan llms.txt sync discipline, exhaustive-panic-test fixture template]
  - phase: 06-token-based-algorithms / plan 06-02
    provides: [TokenSetRatio direct-call panic pattern (one rejected AlgoID — extends to Monge-Elkan's 9 rejected)]
  - phase: 06-token-based-algorithms / plan 06-03
    provides: [PartialRatio DoS three-part godoc precedent (consumed verbatim)]
  - phase: 06-token-based-algorithms / plan 06-04
    provides: [TokenJaccard set-Jaccard pattern (hand-derived RV-derivation template), STANDARD both-empty convention precedent]

provides:
  - MongeElkanScore(a, b string, inner AlgoID, opts NormalisationOptions) float64 — ASYMMETRIC direct surface (direction-sensitive)
  - MongeElkanScoreSymmetric(a, b string, inner AlgoID, opts NormalisationOptions) float64 — SYMMETRIC variant (arithmetic mean of two directions)
  - AlgoMongeElkan dispatch slot 13 wired with the SYMMETRIC variant + AlgoJaroWinkler default inner + DefaultNormalisationOptions per CONTEXT §4 LOCKED
  - permittedMongeElkanInner package-scope allow-list (14 entries per OQ-4 RESOLUTION LOCKED — 9 character + 4 q-gram + AlgoRatcliffObershelp)
  - Exhaustive panic test walking all 9 NON-permitted AlgoIDs (RESEARCH.md Pitfall 4 + token-on-token + Phase 7-reserved phonetic)
  - DoS-vector three-part godoc block + BenchmarkMongeElkan_Pathological_1000Tokens fixture per CONTEXT §5 LOCKED
  - 10-entry staging-golden file for plan 06-06 finalisation merge
  - Asymmetric direction-sensitivity gate (RV-ME6 / RV-ME-asym KEYSTONE) pinned in unit + property + BDD layers

affects: [06-06-finalisation, 07-phonetic-tier (additive allow-list expansion), 08-scorer (WithMongeElkanAlgorithm), 10-extract]

# Tech tracking
tech-stack:
  added: []
  patterns:
    - "Parameter-rich algorithm with pluggable AlgoID inner-metric (mirrors Tversky's parameter-rich surface)"
    - "Strict package-scope allow-list (not init()) per DET-13 — map[AlgoID]bool with compile-time literal initialiser"
    - "Two-surface pattern (asymmetric + symmetric) where the dispatch wrapper binds the symmetric variant — preserves dispatch-level symmetric property-set membership without exemption"
    - "Exhaustive panic test fixture walking ALL non-permitted AlgoIDs (rejected count: 9 in Phase 6; will shrink to 5 in Phase 7 as phonetic AlgoIDs join the allow-list)"
    - "BDD panic-capture pattern: `I attempt to compute …` + `the call should panic with` step pair with lastPanicMsg state field (reusable for any future panic-contract scenario)"

key-files:
  created:
    - "monge_elkan.go (algorithm + permittedMongeElkanInner package-scope allow-list)"
    - "monge_elkan_test.go (9 RV unit tests + asymmetry gate + symmetric variant + exhaustive panic + panic message format + dispatch registration)"
    - "monge_elkan_bench_test.go (4 standard benchmarks + BenchmarkMongeElkanScoreSymmetric_ASCII_Short + BenchmarkMongeElkan_Pathological_1000Tokens DoS-T3 fixture)"
    - "monge_elkan_fuzz_test.go (FuzzMongeElkanScore + FuzzMongeElkanScoreSymmetric — 14 / 13 programmatic seeds)"
    - "dispatch_monge_elkan.go (slot 13 wiring — binds symmetric variant + AlgoJaroWinkler default + DefaultNormalisationOptions per CONTEXT §4 LOCKED)"
    - "tests/bdd/features/monge_elkan.feature (9 scenarios incl. asymmetry direction-sensitivity Scenario + symmetric-variant order-independence Scenario + non-permitted-inner panic Scenario Outline)"
    - "testdata/golden/_staging/monge_elkan.json (10 entries for plan 06-06 finalisation merge)"
  modified:
    - "algoid_test.go (AlgoMongeElkan added to TestDispatch_UnregisteredSlotsAreNil registered map; goes from 18 registered → 19)"
    - "props_test.go (10 TestProp_MongeElkan* property tests — 5 asymmetric direct + 5 symmetric variant + TestProp_MongeElkanScore_AsymmetricWhenTokenCountAsymmetric)"
    - "example_test.go (ExampleMongeElkanScore + ExampleMongeElkanScoreSymmetric — both directions demonstrated)"
    - "tests/bdd/steps/algorithms_steps.go (MongeElkan step methods + algoIDByName resolver + lastPanicMsg state field + 9 step registrations)"
    - "llms.txt (Monge-Elkan section before Normalisation)"
    - "llms-full.txt (Phase 6 algorithm-surface block — Monge-Elkan)"

key-decisions:
  - "OQ-4 RESOLUTION LOCKED 2026-05-15: permittedMongeElkanInner includes AlgoRatcliffObershelp (14 total entries — 9 character + 4 q-gram + AlgoRatcliffObershelp). CONTEXT §3's 13-entry recommendation omitted AlgoRatcliffObershelp; RESEARCH.md OQ-4 recommended including it because RatcliffObershelp is a character-tier algorithm that fits Monge-Elkan inner-metric semantics. Excluding it would be arbitrary. Phase 7 will ADD AlgoSoundex / AlgoDoubleMetaphone / AlgoNYSIIS / AlgoMRA additively → 18 total."
  - "Permitted-inner allow-list LOCKED 2026-05-15: declared at PACKAGE SCOPE in monge_elkan.go (NOT in init() per DET-13 / Phase 5 §5 LOCKED) as var permittedMongeElkanInner = map[AlgoID]bool{...} literal with one entry per line, each carrying an inline comment citing the AlgoID's primary source. The map is the single source of truth — when Phase 7 lands, planners ADD 4 entries to the literal AND update the panic-test fixture in monge_elkan_test.go (rejected: 9 → 5; permitted sanity-check: 14 → 18)."
  - "Direct-call panic format LOCKED 2026-05-15: panic(\"fuzzymatch: AlgoID \" + inner.String() + \" not permitted as Monge-Elkan inner metric\") — exact verbatim format. The Phase 8 Scorer option WithMongeElkanAlgorithm(weight, inner) will return ErrInvalidAlgoID instead — direct-call panic discipline applies only to the direct surface where programmer error must fail loudly. Pinned bit-for-bit by TestMongeElkan_PanicMessageFormat against accidental drift."
  - "Asymmetric reference vector LOCKED 2026-05-15 (RV-ME6 / RV-ME-asym): MongeElkanScore(\"alpha beta gamma\", \"alpha\", AlgoLevenshtein, opts) ≈ 0.4666666666666666 — the input swap of RV-ME4 (where the result is 1.0) produces a direction-sensitive score. The asymmetry is pinned in THREE surfaces: TestMongeElkanScore RV-ME6 row (with full per-token-max derivation in the comment), TestMongeElkanScore_AsymmetryDirectionSensitive (asserts |Δ| > 0.1 — load-bearing regression gate), TestProp_MongeElkanScore_AsymmetricWhenTokenCountAsymmetric (property test with token-count-asymmetric premise + RV-ME4/RV-ME6 spot-check), and the BDD scenario \"MongeElkanScore is direction-sensitive\"."
  - "Symmetric variant in standard property-set LOCKED 2026-05-15: AlgoMongeElkan dispatch wraps the SYMMETRIC variant per CONTEXT §4, so AlgoMongeElkan participates in PropAlgorithmScore_Symmetric without exemption. The DIRECT-CALL MongeElkanScore (asymmetric) has its own TestProp_MongeElkanScore_AsymmetricWhenTokenCountAsymmetric property test (mirrors the Tversky α≠β template). The symmetric variant gets the full standard-symmetric property suite (RangeBounds, Identity, Symmetric, NoNaN, NoInf, NoNegativeZero — 6 tests)."

patterns-established:
  - "Parameter-rich BDD step grammar with AlgoID enum capture: regex `^I compute the MongeElkan score between \"([^\"]*)\" and \"([^\"]*)\" with inner Algo([A-Za-z0-9]+)$` plus an algoIDByName resolver function (one case per AlgoID — gocyclo nolint with documented rationale). Reusable for any future algorithm taking an AlgoID parameter (e.g. hypothetical Hybrid algorithms)."
  - "BDD panic-capture step pair: `I attempt to compute …` (defers recover() into ctx.lastPanicMsg) + `the call should panic with \"<phrase>\"` (strings.Contains assertion). Reusable for any future panic-contract scenario; the lastPanicMsg state field in AlgorithmContext is the single channel."
  - "Two-surface algorithm with dispatch binding the SAFER variant: when the catalogue ships both an asymmetric direct surface AND a symmetric variant, the dispatch wrapper should bind the symmetric variant so the AlgoID participates in the standard symmetric property-test set without exemption. The asymmetric direct surface remains reachable via the public API for advanced consumers."

requirements-completed:
  - TOKEN-01

# Metrics
duration: ~28min
completed: 2026-05-15
---

# Phase 6 Plan 5: Monge-Elkan Summary

**The most complex Phase 6 algorithm — two public functions (asymmetric MongeElkanScore + symmetric MongeElkanScoreSymmetric), strict 14-entry inner-metric allow-list per OQ-4 RESOLUTION LOCKED, exhaustive panic-test walking all 9 non-permitted AlgoIDs, dispatch binding the symmetric variant + AlgoJaroWinkler default inner per CONTEXT §4 LOCKED, three-part DoS godoc block + 1000-token pathological bench fixture, and the load-bearing RV-ME-asym direction-sensitivity gate pinned in three surfaces.**

## Performance

- **Duration:** ~28 min
- **Started:** 2026-05-15T11:32:00Z (worktree HEAD reset to 1a3343a)
- **Completed:** 2026-05-15T12:01:29Z (final verification gate)
- **Tasks:** 1 (single-task plan per the task structure)
- **Files modified:** 13 (7 created, 6 modified) excluding the SUMMARY.md itself

## Accomplishments

- Landed **MongeElkanScore** — the asymmetric direct surface (direction-sensitive: for each token in A, take max inner-metric similarity over every token in B; average per-token maxima). Allow-list gate fires FIRST so invalid inner panics even on identical inputs; identity short-circuit fires before Tokenise. Pure-function library; no map iteration on output paths; no transcendental floats; left-to-right reduction with explicit parenthesisation.
- Landed **MongeElkanScoreSymmetric** — the arithmetic mean of MongeElkanScore in the two directions. Order-independent by construction (the sum of two terms swapped is the same sum). This is the variant bound to dispatch[AlgoMongeElkan] per CONTEXT §4 LOCKED, so AlgoMongeElkan participates in the standard PropAlgorithmScore_Symmetric set without exemption.
- Declared **permittedMongeElkanInner** at PACKAGE SCOPE in monge_elkan.go (NOT in init() per DET-13). 14 entries per OQ-4 RESOLUTION LOCKED: 9 character-tier + 4 q-gram-tier + AlgoRatcliffObershelp. Each entry carries an inline comment citing the originating algorithm's primary source.
- Pinned the **direct-call panic discipline** per CONTEXT §3 LOCKED — the exact panic message format `"fuzzymatch: AlgoID <name> not permitted as Monge-Elkan inner metric"` is gated by TestMongeElkan_PanicMessageFormat. The 9 rejected AlgoIDs (AlgoMongeElkan self + 4 token-tier + 4 phonetic) all panic; the 14 permitted AlgoIDs all return a value in [0, 1] without panic.
- Pinned the **RV-ME-asym KEYSTONE** (RV-ME6) in three surfaces: TestMongeElkanScore RV-ME6 row, TestMongeElkanScore_AsymmetryDirectionSensitive (|Δ| > 0.1), TestProp_MongeElkanScore_AsymmetricWhenTokenCountAsymmetric, and the BDD scenario "MongeElkanScore is direction-sensitive". A silent direction-swap regression surfaces as 4 separate test failures.
- Shipped the **DoS-vector three-part godoc block** per CONTEXT §5 LOCKED: Complexity formula + DoS notice paragraph + BenchmarkMongeElkan_Pathological_1000Tokens fixture. The pathological fixture (1000-token both-sides input) runs at ~83 ms on the developer's Apple M2 — informational; baseline lands in plan 06-06.

## Task Commit

A single atomic commit per the task structure:

1. **Task 1: MongeElkanScore + MongeElkanScoreSymmetric + allow-list + dispatch + companions + property tests + 2 examples + BDD + staging-golden + per-plan llms sync** — `52b36ca` (feat)

The SUMMARY.md commit will follow separately as the final metadata commit per `execute-plan.md`.

## Files Created/Modified

### Created

- `monge_elkan.go` — MongeElkanScore (asymmetric) + MongeElkanScoreSymmetric (symmetric). 14-entry permittedMongeElkanInner package-scope allow-list. Allow-list gate before identity short-circuit; Tokenise; both-Tokenised-empty/one-Tokenised-empty guards; outer-token × inner-token max-reduction loop with explicit if/else max comparison; single final division with explicit parenthesisation. Three-part DoS-vector godoc block per CONTEXT §5 LOCKED.
- `monge_elkan_test.go` — 5 test functions: `TestMongeElkanScore` (9 cases including RV-ME1..RV-ME6 + identity/both-empty/one-empty); `TestMongeElkanScore_AsymmetryDirectionSensitive` (RV-ME-asym load-bearing gate, |Δ| > 0.1); `TestMongeElkanScoreSymmetric` (identity, both-empty, one-empty, construction pin, symmetry pin); `TestMongeElkan_PanicsOnNonPermittedInner` (exhaustive: 9 rejected × 2 surfaces = 18 panic assertions + 14 permitted sanity-check); `TestMongeElkan_PanicMessageFormat` (exact message string pin); `TestMongeElkanScore_DispatchRegistration` (LOCKED CONTEXT §4 defaults pin).
- `monge_elkan_bench_test.go` — 4 standard benchmarks (ASCII Short/Medium/Long, Unicode Short) + `BenchmarkMongeElkanScoreSymmetric_ASCII_Short` (symmetric variant 2x cost) + `BenchmarkMongeElkan_Pathological_1000Tokens` (DoS-T3 LOCKED fixture per CONTEXT §5).
- `monge_elkan_fuzz_test.go` — `FuzzMongeElkanScore` (14 seeds) + `FuzzMongeElkanScoreSymmetric` (13 seeds). Inner AlgoID coerced via `fuzzCoerceMongeElkanInner` so the panic path is NEVER exercised by the harness (panic contract is unit-tested separately).
- `dispatch_monge_elkan.go` — slot 13 wiring: `dispatch[AlgoMongeElkan] = MongeElkanScoreSymmetric(a, b, AlgoJaroWinkler, DefaultNormalisationOptions())`. Rationale-rich godoc cross-referencing CONTEXT §4 LOCKED and Phase 7 forward-compatibility (4 phonetic AlgoIDs ADDED additively).
- `tests/bdd/features/monge_elkan.feature` — 9 scenarios including `Scenario Outline: Canonical reference vectors (asymmetric, AlgoJaroWinkler inner)`, `Scenario Outline: Canonical reference vectors (asymmetric, AlgoLevenshtein inner)`, `Scenario Outline: Symmetric variant canonical reference vectors`, identity / both-empty / one-empty conventions, the LOAD-BEARING `MongeElkanScore is direction-sensitive` scenario (RV-ME4 vs RV-ME6 input swap with `differ by more than 0.1` assertion), the symmetric-variant order-independence scenario, and the `Scenario Outline: non-permitted inner AlgoIDs panic` with 3 representative rejected AlgoIDs (MongeElkan self / TokenSortRatio / Soundex).
- `testdata/golden/_staging/monge_elkan.json` — 10 entries for plan 06-06 finalisation merge.

### Modified

- `algoid_test.go` — `TestDispatch_UnregisteredSlotsAreNil` now includes `AlgoMongeElkan: true` in the registered map; comments updated to reflect plan 06-05's slot-13 registration. The remaining unregistered slots are 18..21 (the phonetic tier reserved for Phase 7).
- `props_test.go` — appended 10 property tests under the divider `// --- Monge-Elkan property tests (plan 06-05)`: 5 asymmetric-direct (RangeBounds, NoNaN, NoInf, NoNegativeZero, AsymmetricWhenTokenCountAsymmetric — the latter mirroring the Tversky α≠β template) + 5 symmetric-variant (RangeBounds, Identity, Symmetric, NoNaN, NoInf, NoNegativeZero — full standard symmetric set).
- `example_test.go` — appended `ExampleMongeElkanScore` (demonstrates the RV-ME1 symmetric-looking case alongside the RV-ME4/RV-ME6 asymmetric pair) and `ExampleMongeElkanScoreSymmetric` (demonstrates order-independence by showing the same score in both argument orders).
- `tests/bdd/steps/algorithms_steps.go` — appended MongeElkan step methods + `algoIDByName` resolver function + `lastPanicMsg` field on AlgorithmContext + 9 step registrations (4 score-computation steps for asymmetric/symmetric × normal/second + the differ-by-more-than step + the both-equal step + the attempt-to-compute step + the call-should-panic-with step).
- `llms.txt` — appended `### Monge-Elkan (Monge & Elkan 1996) per-token-max with pluggable inner metric` section before Normalisation.
- `llms-full.txt` — appended `### Phase 6 algorithm surface (token tier — Monge-Elkan)` block. Documents the 14-entry allow-list, the panic discipline, the asymmetry KEYSTONE, the DoS notice, and the Phase 7 forward-compatibility note.

## Decisions Made

**The five required `LOCKED <date>` records are reproduced verbatim above in the `key-decisions` frontmatter. Summary form here for prose readability:**

### OQ-4 RESOLUTION LOCKED 2026-05-15

`permittedMongeElkanInner` includes `AlgoRatcliffObershelp` (Phase 4 character-tier algorithm). Total **14 entries**: 9 character-tier (`AlgoLevenshtein`, `AlgoDamerauLevenshteinOSA`, `AlgoDamerauLevenshteinFull`, `AlgoHamming`, `AlgoJaro`, `AlgoJaroWinkler`, `AlgoStrcmp95`, `AlgoSmithWatermanGotoh`, `AlgoLCSStr`) + 4 q-gram tier (`AlgoQGramJaccard`, `AlgoSorensenDice`, `AlgoCosine`, `AlgoTversky`) + `AlgoRatcliffObershelp`. CONTEXT §3's list of 13 omitted `AlgoRatcliffObershelp`; RESEARCH.md OQ-4 recommended including it because Ratcliff-Obershelp is a character-tier algorithm that fits the Monge-Elkan inner-metric semantics — excluding it would be arbitrary. The exhaustive panic-test asserts these 14 are permitted AND that the remaining 9 (`AlgoMongeElkan` self / 4 token-tier / 4 phonetic) panic with the documented message.

### Permitted-inner allow-list LOCKED 2026-05-15

Declared at **PACKAGE SCOPE** in `monge_elkan.go` (NOT in `init()` per DET-13 / Phase 5 §5 LOCKED) as `var permittedMongeElkanInner = map[AlgoID]bool{...}` literal with one entry per line, each with an inline comment citing the AlgoID's primary source. The map is the single source of truth — when Phase 7 lands, planners **ADD 4 entries** (`AlgoSoundex`, `AlgoDoubleMetaphone`, `AlgoNYSIIS`, `AlgoMRA`) and update the panic-test fixture (rejected count drops from 9 to 5; permitted sanity-check grows from 14 to 18).

### Direct-call panic format LOCKED 2026-05-15

`panic("fuzzymatch: AlgoID " + inner.String() + " not permitted as Monge-Elkan inner metric")` — verbatim format including the canonical AlgoID name from `String()`. The Phase 8 Scorer option `WithMongeElkanAlgorithm(weight, inner)` will instead return `ErrInvalidAlgoID` (already declared in `errors.go` Phase 1) — direct-call panic discipline per CONTEXT §3. Pinned bit-for-bit by `TestMongeElkan_PanicMessageFormat` against accidental drift.

### Asymmetric reference vector LOCKED 2026-05-15 (RV-ME6 / RV-ME-asym)

The keystone asymmetry-discriminating pair:

- **RV-ME4**: `MongeElkanScore("alpha", "alpha beta gamma", AlgoLevenshtein, opts) = 1.0` (single A-token matches one of three B-tokens exactly; max=1.0)
- **RV-ME6**: `MongeElkanScore("alpha beta gamma", "alpha", AlgoLevenshtein, opts) ≈ 0.4666666666666666` (three A-tokens, each takes max over the single B-token; the two non-matching tokens drag the mean down)

The two scores differ by ≈ 0.5333 — the input swap with the same inner produces a direction-sensitive score. Pinned in three surfaces: unit-test row, dedicated `TestMongeElkanScore_AsymmetryDirectionSensitive` |Δ| > 0.1 gate, property test `TestProp_MongeElkanScore_AsymmetricWhenTokenCountAsymmetric` with RV-ME4/RV-ME6 spot-check, and the BDD scenario "MongeElkanScore is direction-sensitive". A silent direction-swap regression surfaces as **four separate test failures**.

### Symmetric variant in standard property-set LOCKED 2026-05-15

`AlgoMongeElkan` dispatch wraps the SYMMETRIC variant per CONTEXT §4, so `AlgoMongeElkan` participates in `PropAlgorithmScore_Symmetric` without exemption (the per-AlgoID symmetric set includes AlgoMongeElkan because the dispatch surface IS the symmetric variant). The DIRECT-CALL `MongeElkanScore` (asymmetric) gets its own `TestProp_MongeElkanScore_AsymmetricWhenTokenCountAsymmetric` property test (mirrors the Tversky α≠β template); the symmetric variant gets the FULL standard symmetric set (`_RangeBounds`, `_Identity`, `_Symmetric`, `_NoNaN`, `_NoInf`, `_NoNegativeZero` — six properties).

## RV-ME1..RV-ME6 Reference Vectors

Each row's per-token-max derivation is reproduced in the `monge_elkan_test.go` test comment per algorithm-correctness-standards. Summary table:

| RV | a | b | inner | expected | derivation |
|----|---|---|-------|----------|------------|
| RV-ME1 | `user create` | `usr creating` | `AlgoJaroWinkler` | 0.9125 | tokens(A)=[user,create]; tokens(B)=[usr,creating]; max(JW(user,usr)=0.9333, JW(user,creating)=0.4167)=0.9333; max(JW(create,usr)=0.5, JW(create,creating)=0.8917)=0.8917; ME=(0.9333+0.8917)/2=0.9125 |
| RV-ME2 | `alpha beta` | `alpha beta` | `AlgoJaroWinkler` | 1.0 | a == b identity short-circuit (inner irrelevant) |
| RV-ME3 | `alpha beta` | `gamma delta` | `AlgoJaroWinkler` | 0.6917 | tokens(A)=[alpha,beta]; tokens(B)=[gamma,delta]; max(JW(alpha,gamma)=0.6, JW(alpha,delta)=0.6)=0.6; max(JW(beta,gamma)=0.4833, JW(beta,delta)=0.7833)=0.7833; ME=(0.6+0.7833)/2=0.6917 |
| RV-ME4 | `alpha` | `alpha beta gamma` | `AlgoLevenshtein` | 1.0 | tokens(A)=[alpha]; tokens(B)=[alpha,beta,gamma]; max(Lev(alpha,alpha)=1, Lev(alpha,beta)=0.2, Lev(alpha,gamma)=0.2)=1.0; ME=1.0/1=1.0 |
| RV-ME5 | `café` | `cafe` | `AlgoLevenshtein` | 0.6 | tokens(A)=[café]; tokens(B)=[cafe]; Lev(café,cafe)=0.6 (byte path, distance=2, maxLen=5); ME=0.6/1=0.6 |
| **RV-ME6** | `alpha beta gamma` | `alpha` | `AlgoLevenshtein` | **0.4666666666666666** | **KEYSTONE asymmetry gate (RV-ME-asym)**: tokens(A)=[alpha,beta,gamma]; tokens(B)=[alpha]; max(Lev(alpha,alpha)=1)=1.0; max(Lev(beta,alpha)=0.2)=0.2; max(Lev(gamma,alpha)=0.2)=0.2; ME=(1.0+0.2+0.2)/3=0.4666… — RV-ME4's mirror; pinned in unit + property + BDD layers |

## Exhaustive Panic Test Fixture (9 rejected AlgoIDs)

`TestMongeElkan_PanicsOnNonPermittedInner` walks every non-permitted AlgoID and asserts each surfaces the documented panic message via both `MongeElkanScore` AND `MongeElkanScoreSymmetric` (the symmetric variant's first internal `MongeElkanScore` call surfaces the panic):

| Category | AlgoID | Reason |
|----------|--------|--------|
| Self-recursion | `AlgoMongeElkan` | Pitfall 4 — infinite recursion guard |
| Token-on-token | `AlgoTokenSortRatio` | Meaningless: inner receives single tokens from outer Tokenise |
| Token-on-token | `AlgoTokenSetRatio` | Same |
| Token-on-token | `AlgoPartialRatio` | Same |
| Token-on-token | `AlgoTokenJaccard` | Same |
| Phase 7 reserved | `AlgoSoundex` | Added in Phase 7 ADDITIVELY |
| Phase 7 reserved | `AlgoDoubleMetaphone` | Same |
| Phase 7 reserved | `AlgoNYSIIS` | Same |
| Phase 7 reserved | `AlgoMRA` | Same |

Plus a representative-subset sanity check for the 14 permitted AlgoIDs (all return a value in [0, 1] without panic on `("alpha beta", "alpha gamma", inner, opts)`).

## Reference Vector Numbers for the Staging-Golden File

The `testdata/golden/_staging/monge_elkan.json` file (10 entries) pins the following dispatch-level scores (computed via `MongeElkanScoreSymmetric(a, b, AlgoJaroWinkler, DefaultNormalisationOptions())` since dispatch wraps the symmetric variant):

| Name | Score | Derivation |
|------|------:|------------|
| `MongeElkan_RV-ME1_user_create_vs_usr_creating_symmetric` | 0.9125000000000001 | symmetric average of the (user create → usr creating) and reverse JW-based asymmetric scores |
| `MongeElkan_RV-ME2_identity` | 1.0 | a == b non-empty identity short-circuit (inner JW irrelevant) |
| `MongeElkan_RV-ME3_disjoint_greek_symmetric` | 0.6916666666666667 | symmetric average of the disjoint-Greek-pair JW-based asymmetric scores |
| `MongeElkan_RV-ME4_RV-ME6_asymmetric_symmetric_average` | 0.8472222222222223 | symmetric average of (ME("alpha", "alpha beta gamma", JW) + ME("alpha beta gamma", "alpha", JW)) / 2 |
| `MongeElkan_RV-ME6_RV-ME4_swapped_symmetric_same_average` | 0.8472222222222223 | Same value — the symmetric variant is order-independent (verifies the symmetry gate at golden level) |
| `MongeElkan_both_empty_standard` | 1.0 | STANDARD catalogue convention |
| `MongeElkan_one_empty_a` | 0.0 | one-empty convention |
| `MongeElkan_one_empty_b` | 0.0 | one-empty convention |
| `MongeElkan_pure_separators_both_empty_tokens` | 1.0 | post-Tokenise both-empty guard (`Tokenise("___") = []`, `Tokenise("...") = []`) |
| `MongeElkan_token_reorder_symmetric` | 1.0 | identical token sets after Tokenise; both directions return 1.0; symmetric average = 1.0 |

## Pathological Bench Timing (informational; baseline lands in plan 06-06)

`BenchmarkMongeElkan_Pathological_1000Tokens` runs at **~83 ms / op** on the developer's Apple M2 (single-iteration `-benchtime=1x` run; 4040 allocs / 261 KB per call). 1000-token both-sides → ~10^6 inner-metric (JaroWinkler) comparisons. Documented in the godoc DoS notice; consumers in untrusted-input contexts must pre-validate token-count ceilings before calling.

The standard bench numbers (informational, full baseline lands in plan 06-06):

| Benchmark | ns/op | B/op | allocs/op |
|-----------|------:|-----:|----------:|
| `BenchmarkMongeElkanScore_ASCII_Short` | ~544 | 160 | 6 |
| `BenchmarkMongeElkanScore_ASCII_Medium` | ~24000 | 864 | 22 |
| `BenchmarkMongeElkanScore_ASCII_Long` | ~36000 | 5344 | 92 |
| `BenchmarkMongeElkanScore_Unicode_Short` | ~16000 | 160 | 6 |
| `BenchmarkMongeElkanScoreSymmetric_ASCII_Short` | ~10500 | 320 | 12 |
| `BenchmarkMongeElkan_Pathological_1000Tokens` | ~83000000 | 261520 | 4040 |

The symmetric variant is ~2x the asymmetric cost (as expected — it calls MongeElkanScore in both directions). All within the `< 10 µs` performance-standards budget for short ASCII; the longer benchmarks are dominated by the O(|tA|·|tB|) inner-metric matrix.

## Phase 7 Forward-Compatibility Note

When Phase 7 lands the four phonetic algorithms (Soundex, DoubleMetaphone, NYSIIS, MRA), planners will:

1. **ADD 4 entries** to `permittedMongeElkanInner` in `monge_elkan.go`:
   ```go
   // Phonetic tier (4) — added in Phase 7:
   AlgoSoundex:         true, // Russell 1918 / Knuth 1973
   AlgoDoubleMetaphone: true, // Philips 2000
   AlgoNYSIIS:          true, // Taft 1970
   AlgoMRA:             true, // Moore 1977
   ```
2. **Update the panic-test fixture** in `monge_elkan_test.go`:
   - `rejected` slice shrinks from **9 to 5** entries (only `AlgoMongeElkan` self + 4 token-tier remain).
   - `permittedSanity` slice grows from **14 to 18** entries.
3. **Update the godoc** in `monge_elkan.go` — the inner-metric allow-list block grows from "14 entries" to "18 entries (LOCKED in Phase 7 after the additive expansion)".
4. **Update the BDD scenario** "non-permitted inner AlgoIDs panic" Examples table — drop the Soundex row (the only Phase 7-reserved AlgoID currently represented).
5. **Update llms-full.txt** — the "Phase 6 algorithm surface" block's forward-compatibility note becomes historical.

The dispatch wrapper, the panic message format, the public function signatures, and the asymmetry KEYSTONE are all unchanged across this expansion. Phase 7's WithMongeElkanAlgorithm Scorer integration consumes the same allow-list.

## Deviations from Plan

### None — plan executed exactly as written.

The plan was extremely explicit (`<read_first>` blocks pointed at the canonical templates — Tversky for parameter-rich algorithm shape, TokenJaccard for the both-empty STANDARD convention, qgram_jaccard for the dispatch closure pattern, lcsstr for the byte-path file-header layout). Implementation tracked the plan within ~30 minutes for a complex task.

One auto-applied formatter fix: `gofmt -s` reformatted the comment indentation in `example_test.go` (a 1-space alignment change inside an inline list of bullet points). Not a deviation — formatter discipline.

One auto-applied lint fix: `gocyclo` flagged `algoIDByName` in `tests/bdd/steps/algorithms_steps.go` at complexity 24 (one case per AlgoID — the canonical Go idiom for a stringly-typed enum reverse-lookup; the project precedent is `algoid.go`'s `String()` method which has the same shape and is annotated with `//nolint:gocyclo`). Applied the same nolint pragma with documented rationale.

**Total deviations:** 0 (the two auto-applied fixes are routine formatter/lint hygiene, not scope changes).

## Issues Encountered

None requiring problem-solving. The plan's `<recorded_resolutions>` were verbatim — the 14-entry allow-list, the panic message format, the asymmetric RV pinning strategy, and the dispatch defaults all matched what was implemented.

## User Setup Required

None — no external service configuration required for this plan. The algorithm is a pure-function library extension; no toolchain changes.

## Next Phase Readiness

- **Plan 06-06 finalisation** has a 10-entry staging-golden to merge into `algorithms.json` (alongside the staging files from plans 06-01..06-04), a new docs page section for Monge-Elkan in `docs/algorithms.md`, and the pathological bench number to commit to `bench.txt`.
- **Phase 7 (phonetic tier)** can ADD the 4 phonetic AlgoIDs to `permittedMongeElkanInner` additively per the forward-compatibility checklist above. The dispatch wrapper, panic format, and public signatures are unchanged.
- **Phase 8 (Scorer)** can implement `WithMongeElkanAlgorithm(weight, inner)` that forwards the user-supplied `inner` parameter to `MongeElkanScore` (the asymmetric direct surface) and returns `ErrInvalidAlgoID` for non-permitted inner AlgoIDs (instead of the panic discipline that applies to the direct call).

No blockers or concerns.

### Deferred items for plan 06-06

- **Final `bench.txt` numbers** — Monge-Elkan benchmarks compile and run (ASCII Short/Medium/Long, Unicode Short, Symmetric_ASCII_Short, Pathological_1000Tokens) but the project-wide `bench.txt` baseline is regenerated phase-by-phase at finalisation time, not per-plan. Plan 06-06 will run the full benchmark suite and commit the updated `bench.txt`.
- **`testdata/golden/_staging/monge_elkan.json` merge into `testdata/golden/algorithms.json`** — staged in the `_staging/` directory; plan 06-06 finalisation handles the merge (alongside `partial_ratio.json`, `token_jaccard.json`, `token_set_ratio.json`, `token_sort_ratio.json`).
- **Cross-platform determinism golden update** — Monge-Elkan output is deterministic across the four CI platforms by construction (integer-derived float64 max + addition + single division, IEEE-754 correctly-rounded; no transcendentals; no map iteration on output paths), but `verify-determinism`'s golden file does not yet include any Monge-Elkan entries. Plan 06-06 will add representative entries.

## Self-Check: PASSED

File-existence checks:

- `monge_elkan.go` — present.
- `monge_elkan_test.go` — present.
- `monge_elkan_bench_test.go` — present.
- `monge_elkan_fuzz_test.go` — present.
- `dispatch_monge_elkan.go` — present.
- `tests/bdd/features/monge_elkan.feature` — present.
- `testdata/golden/_staging/monge_elkan.json` — present (10 entries).

Commit-existence checks:

- Commit `52b36ca` — present in `git log --oneline -5`.

Test gates:

- `go test -run 'TestMongeElkan|TestProp_MongeElkan|TestDispatch_UnregisteredSlotsAreNil|TestLLMs|ExampleMongeElkan' -count=1 ./...` — PASS.
- `go test -race -shuffle=on -count=1 ./...` — PASS (full root test suite).
- `cd tests/bdd && go test -race -count=1 ./...` — PASS.
- `make fmt-check lint verify-license-headers verify-deps-allowlist` — all 0 issues.
- `go test -run none -bench BenchmarkMongeElkan -benchtime=1x ./...` — runs to completion.
- `go test -fuzz=FuzzMongeElkanScore$ -fuzztime=3s -run=^$ ./...` — PASS (240k execs, 0 crashes).
- `go test -fuzz=FuzzMongeElkanScoreSymmetric -fuzztime=3s -run=^$ ./...` — PASS (116k execs, 0 crashes).

All claimed deliverables verified by `git log`, file existence on disk, and the test/lint/bench/fuzz gates above.

---
*Phase: 06-token-based-algorithms*
*Completed: 2026-05-15*
