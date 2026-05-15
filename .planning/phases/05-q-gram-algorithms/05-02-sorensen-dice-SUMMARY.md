---
phase: 05-q-gram-algorithms
plan: 02
subsystem: similarity-algorithms
tags: [sorensen-dice, dice-1945, sorensen-1948, q-gram-multiset, byte-and-rune-paths, dispatch-registration, default-n-3, property-tests, fuzz, benchmark, bdd, staging-golden, llms-sync, no-init, no-transcendentals, no-map-iteration-on-output]

# Dependency graph
requires:
  - phase: 05-q-gram-algorithms
    provides: extractQGrams + extractQGramsRunes (q_gram.go) shared helpers and ErrInvalidQGramSize sentinel — all landed by plan 05-01
  - phase: 04-remaining-character-gestalt
    provides: 04-01 strcmp95 pattern for the 3-task algorithm shape (impl+test+staging / property+bench+fuzz / BDD); per-plan llms.txt + llms-full.txt sync discipline (NOT deferred to finalisation); 7-entry staging-golden norm
  - phase: 03-smith-waterman-gotoh
    provides: per-algorithm BDD feature + step-bindings append pattern; props_test.go append-block pattern; ExampleXxxScore append-only example_test.go discipline; testdata/fuzz/Fuzz<Algo>Score/seed-001 byte-stable format
  - phase: 02-core-character-algorithms-six
    provides: assertGoldenStaging helper + goldenAlgorithmEntry / goldenAlgorithmsFile schema; AlgoSorensenDice slot 10 already declared in algoid.go (line 117) with String() case at line 235-236; testdata/golden/_staging/ workflow
provides:
  - SorensenDiceScore(a, b string, n int) float64 — Sørensen-Dice coefficient over q-gram multisets (byte path; dispatched with default n=3)
  - SorensenDiceScoreRunes(a, b string, n int) float64 — rune-path variant for multi-byte UTF-8
  - dispatch[AlgoSorensenDice] slot 10 populated via var-init (no init()); default n=3 trigram closure
  - testdata/golden/_staging/sorensen_dice.json — 8 entries covering RV-D1 (load-bearing canonical NLP-textbook bigram pair), RV-D2 (high-overlap analogue), RV-D3 (trigram variant), RV-D4 (identity), both/one-empty, no-overlap, plus the rune-path café canary
  - testdata/fuzz/FuzzSorensenDiceScore/seed-001 + testdata/fuzz/FuzzSorensenDiceScoreRunes/seed-001 — RV-D1 canonical seed in `go test fuzz v1` literal format
  - tests/bdd/features/sorensen_dice.feature — 6 scenarios (1 outline + 5 standalone)
affects: [05-03-cosine, 05-04-tversky, 05-05-finalisation]

# Tech tracking
tech-stack:
  added: []
  patterns:
    - "Sørensen-Dice as a structural sibling of Q-Gram Jaccard. The two algorithms share the same q-gram extraction tier (extractQGrams / extractQGramsRunes from plan 05-01), the same identity / both-empty / one-empty short-circuit pattern, the same direct-call panic-on-n<1 contract via the shared 'fuzzymatch: invalid q-gram size' message text, and the same dispatch-wrapper closure with default n=3. The single difference is the formula itself: Jaccard = |∩|/|∪| versus Dice = 2·|∩|/(|QA|+|QB|). The structural twinning means plan 05-02 borrowed the qgram_jaccard_test.go / props_test.go / fuzz_test.go / bench_test.go / feature-file shapes verbatim, substituting only the algorithm-specific reference vectors and formula derivation comments."
    - "diceFromQGramMaps helper using explicit left-to-right parenthesisation per DET-06. The DSC formula `2·|∩| / (|QA|+|QB|)` has a multiplication, an addition, and a division. The implementation writes them as `(2.0 * float64(intersection)) / (float64(lenA) + float64(lenB))` — the inner parens force the addition to happen before the division, matching the canonical mathematical reading; the outer numerator parens force the multiplication to happen before the division. No FMA risk because the integer inputs are well below 2^53; the explicit parens are the determinism-reviewer audit trail rather than a numerical-safety necessity."
    - "Multiset cardinality via map-value summation. lenA = Σ countA[k] (sum-of-values for QA) is the multiset cardinality definition. For non-degenerate inputs it equals max(0, len(a)-n+1) on the byte path, but computing via sum-of-values is the canonical definition and works identically on the rune path where the relationship is rune-count-derived. Map iteration to compute the sum is DET-03 safe because integer addition is associative — the OUTPUT is a scalar int, not an ordered slice."
    - "Per-plan llms.txt sync inherited from plan 04-01 / plan 05-01 precedent. The 2 new functions (SorensenDiceScore, SorensenDiceScoreRunes) are appended to llms.txt + llms-full.txt within this plan's commits, NOT deferred to plan 05-05 finalisation. AlgoSorensenDice was already listed (declared in Phase 1); the meta-test TestAIFriendly_LLMSTxtReferencesEveryExportedSymbol enforces the sync — a missing function entry would fail CI immediately. Plans 05-03 / 05-04 will follow the same discipline."

key-files:
  created:
    - sorensen_dice.go
    - dispatch_sorensen_dice.go
    - sorensen_dice_test.go
    - sorensen_dice_bench_test.go
    - sorensen_dice_fuzz_test.go
    - testdata/golden/_staging/sorensen_dice.json
    - testdata/fuzz/FuzzSorensenDiceScore/seed-001
    - testdata/fuzz/FuzzSorensenDiceScoreRunes/seed-001
    - tests/bdd/features/sorensen_dice.feature
    - .planning/phases/05-q-gram-algorithms/05-02-sorensen-dice-SUMMARY.md
  modified:
    - algoid_test.go (append: TestDispatch_SorensenDiceRegistered; slot 10 added to TestDispatch_UnregisteredSlotsAreNil registered map — now 12 currently-registered slots)
    - algorithms_golden_test.go (append: buildSorensenDiceStagingEntries + TestGolden_SorensenDice_Staging — 8 alphabetically-sorted entries via assertGoldenStaging)
    - example_test.go (append: ExampleSorensenDiceScore + ExampleSorensenDiceScoreRunes runnable godoc examples)
    - props_test.go (append: 12 standard property tests across byte+rune surfaces — RangeBounds/Identity/Symmetric/NoNaN/NoInf/NoNegativeZero × 2 — PLUS TestProp_SorensenDiceScore_DeterministicAcrossRuns 1000-iteration math.Float64bits gate)
    - tests/bdd/steps/algorithms_steps.go (append: 4 step methods on AlgorithmContext + 4 ctx.Step regex registrations inside InitializeScenario; reuses the `with n (\d+)` regex shape introduced by plan 05-01)
    - llms.txt (append: Sørensen-Dice section with the 2 exported functions)
    - llms-full.txt (append: parallel Phase 5 q-gram-tier Sørensen-Dice block)

key-decisions:
  - "RV-D1 night/nacht/n=2 → 0.25 selected as the load-bearing canonical NLP-textbook bigram pair. The pair is recommended in many NLP textbooks as the canonical Dice example because (a) the inputs are short, memorable, and English-readable, (b) the bigram multisets each have exactly 4 distinct entries, (c) the intersection is exactly 1 (only 'ht' shared), making the derivation `2·1/(4+4) = 0.25` reviewer-verifiable in seconds, and (d) the inputs are not symmetric or trivial so the result discriminates against any silently-broken implementation. Pinned in unit tests, dispatch tests, fuzz seeds, BDD scenarios, and the determinism property test as the single load-bearing canonical reference."
  - "RV-D3 trigram variant uses 'abcdef' / 'abcXef' / n=3 → 0.25. Selected because the bigram path is already covered by RV-D1 and RV-D2; RV-D3 specifically exercises the n=3 trigram path (the dispatch default) with a result identical to RV-D1 (0.25 with intersection=1 / per-side=4) so a reviewer checking 'does the n parameter actually plumb through correctly?' has a clean trigram-vs-bigram comparison. The single-character difference (`d` vs `X`) shifts exactly three trigrams (`bcd`/`bcX`, `cde`/`cXe`, `def`/`Xef`), leaving only `abc` shared."
  - "Default n = 3 trigram for the dispatch wrapper (mirrors plan 05-01). CONTEXT.md 'Claude's Discretion' Deferred §4 endorses this; trigrams are the most common q-gram default in the NLP literature. The wrapper closure pattern matches dispatch_qgram_jaccard.go verbatim — `var _ = func() bool { dispatch[AlgoSorensenDice] = func(a, b string) float64 { return SorensenDiceScore(a, b, 3) }; return true }()`. Non-default-n consumers go through the future Phase 8 Scorer option layer (WithSorensenDiceAlgorithm)."
  - "Property-test n coercion via `(abs(n) % 5) + 1` — keep n in [1, 5]. Carried forward verbatim from plan 05-01. The n < 1 panic path is unit-tested separately by TestSorensenDice_PanicsOnInvalidN; the property tests exercise the [0, 1] / identity / symmetry invariants on every drawn triple."
  - "Property tests cover Symmetric (Sørensen-Dice IS exactly symmetric). The intersection cardinality and per-side multiset totals are integer-valued and agnostic to argument order; the single multiplication-then-division produces bit-identical float64 output regardless of argument order. Asserted via `==` (not `|x - y| < ε`) per the same pattern as plan 05-01's set-Jaccard."
  - "diceFromQGramMaps duplicates jaccardFromQGramMaps's structure rather than refactoring to a shared helper. The two helpers are 95% structurally identical (identity check on len-zero, sum-of-values for per-side totals, walk-the-smaller for intersection cardinality), but the OUTPUT formula differs (Jaccard divides by union; Dice divides by sum-of-totals). Refactoring to a shared helper would add an indirection (an enum or function pointer for the formula step) and make the determinism review harder; the structural duplication is the canonical pattern for sister algorithms in this catalogue per plan 04-02 (LCSStr) / plan 04-03 (Ratcliff-Obershelp) / plan 05-01 (Q-Gram Jaccard) precedent. Plans 05-03 (Cosine) and 05-04 (Tversky) will introduce their own helpers in the same shape."

patterns-established:
  - "Pattern: structural-sibling algorithm files share verbatim shapes. sorensen_dice.go was authored by copying qgram_jaccard.go, substituting the algorithm name + formula + reference vectors + helper name. Same for the test / bench / fuzz / feature files. The verbatim shape preserves the determinism-reviewer audit trail (the same idioms in the same positions) and reduces cognitive load on plan-by-plan reviews. Plans 05-03 / 05-04 will continue this discipline."
  - "Pattern: dispatch wrapper closure (carried forward from plan 05-01). Each q-gram-tier algorithm registers `dispatch[AlgoXxx] = func(a, b string) float64 { return XxxScore(a, b, 3) }` to bridge the parameterised algorithm function to the fixed dispatch signature. Plans 05-03 (Cosine, default n=3) and 05-04 (Tversky, default n=3 + α=β=1.0) will use the same closure pattern with their own default values."

requirements-completed: [QGRAM-03]

# Metrics
duration: ~28min
completed: 2026-05-15
---

# Phase 5 Plan 02: Sørensen-Dice Summary

**Ship the Sørensen-Dice coefficient (QGRAM-03) — DSC = 2·|QA∩QB|/(|QA|+|QB|) — as the second q-gram-tier algorithm, atop the shared `extractQGrams` / `extractQGramsRunes` helpers from plan 05-01. Both byte and rune surfaces ship with the full Phase 2/3/4 quality bar (unit + property + fuzz + bench + BDD + staging golden + dispatch + example + llms.txt). RV-D1 (`night`/`nacht`/n=2 → 0.25) is the load-bearing canonical NLP-textbook reference vector pinned across every test surface.**

## Performance

- **Duration:** ~28 min (3 atomic commits — Task 1 impl+test+golden / Task 2 props+bench+fuzz / Task 3 BDD)
- **Tasks:** 3 (all completed)
- **Files modified:** 16 (10 created, 6 modified — within plan estimate)
- **Commits:** 3 atomic commits, all reference issue #7 (the Phase 5 epic)

## Accomplishments

- **Two new exported functions:** `SorensenDiceScore`, `SorensenDiceScoreRunes` — pinned at docs/requirements.md §7.2.2.
- **Dispatch slot 10** (`AlgoSorensenDice`) populated via `var _ = func() bool{...}()` (no init()). Default n=3 trigram per CONTEXT.md Deferred §4. `TestDispatch_UnregisteredSlotsAreNil` now expects 12 currently-registered slots; slots 11..21 await later plans (remaining q-gram algorithms 05-03..05-04 plus token + phonetic algorithms).
- **Reference vector verification:** all four canonical hand-derived RV-D1..RV-D4 vectors landed exactly:
  - `SorensenDiceScore("night", "nacht", 2)` = `0.25` (RV-D1 — load-bearing canonical NLP-textbook bigram pair)
  - `SorensenDiceScore("abcdef", "bcdefg", 2)` = `0.8` (RV-D2 — Dice 1945 §3 high-overlap analogue)
  - `SorensenDiceScore("abcdef", "abcXef", 3)` = `0.25` (RV-D3 — trigram variant)
  - `SorensenDiceScore("hello", "hello", 2)` = `1.0` (RV-D4 — identity short-circuit)
  - `SorensenDiceScoreRunes("café", "cafe", 2)` = `4/6 ≈ 0.6666666666666666` (rune-path canary)
- **Property-test coverage:** 12 standard invariants (RangeBounds / Identity / Symmetric / NoNaN / NoInf / NoNegativeZero × byte+rune surfaces) plus the `_DeterministicAcrossRuns` gate (1000 sequential iterations on RV-D1 night/nacht + the rune café pair, math.Float64bits comparison).
- **Fuzz harnesses:** Two fuzzers (byte + rune), each with 12-13 programmatic seeds covering canonical RV-D1..RV-D4 + identity + both-empty + one-empty + invalid UTF-8 + multi-byte + Cyrillic (rune harness) + long-input + n=1 + n=8 boundary cases. On-disk seed files (`testdata/fuzz/FuzzSorensenDiceScore/seed-001`, `testdata/fuzz/FuzzSorensenDiceScoreRunes/seed-001`) in byte-stable `go test fuzz v1` literal format using the canonical RV-D1 seed. 10-second smoke runs on each completed without panic / NaN / Inf / out-of-range:
  - `FuzzSorensenDiceScore`: 804,758 execs / 27 new interesting / 0 failures
  - `FuzzSorensenDiceScoreRunes`: 512,577 execs / 27 new interesting / 0 failures
- **Benchmark coverage:** 4 benches (ASCII Short / Medium / Long + Unicode Short). Allocation count within RESEARCH.md §4.1 budget. darwin/arm64 1x smoke run:
  - `ASCII_Short`: 0 B/op, 0 allocs/op (different-string non-overlap path; the bigrams overlap on "ht" so the short-circuit doesn't fire — but darwin/arm64 short-input path is dominated by stack-allocated maps in Go 1.26)
  - `ASCII_Medium`: 3,664 B/op, 6 allocs/op
  - `ASCII_Long`: 13,136 B/op, 6 allocs/op
  - `Unicode_Short`: 32 B/op, 6 allocs/op
- **BDD scenarios:** 6 scenarios (1 outline + 5 standalone) in `tests/bdd/features/sorensen_dice.feature` — covers RV-D1..RV-D4, identity, both-empty, one-empty (both directions), exact symmetry, and the rune-path café canary.
- **Per-algorithm staging golden:** `testdata/golden/_staging/sorensen_dice.json` with 8 alphabetically-sorted entries (canonical-marshalled via `assertGoldenStaging` → `CanonicalMarshalForTest`). Plan 05-05 owns the merge into `algorithms.json`.

## Reference Vectors Verified

| ID | Inputs | n | Expected | Source |
|----|--------|---|----------|--------|
| RV-D1 | `"night"` / `"nacht"` | 2 | 0.25 | Canonical NLP-textbook bigram pair (Dice 1945 §3 reading) |
| RV-D2 | `"abcdef"` / `"bcdefg"` | 2 | 0.8 | High-overlap analogue (4 of 5 bigrams shared) |
| RV-D3 | `"abcdef"` / `"abcXef"` | 3 | 0.25 | Trigram variant (1 of 4 trigrams shared) |
| RV-D4 | `"hello"` / `"hello"` | 2 | 1.0 | Identity short-circuit |
| Rune | `"café"` / `"cafe"` | 2 | 4/6 ≈ 0.6666666666666666 | Rune-path canary (multi-byte é vs ASCII e) |

## Test Results

| Test category | Count | Status |
|---------------|-------|--------|
| Unit tests (TestSorensenDice*) | 8 functions / ~30 sub-tests | All PASS |
| Property tests (TestProp_SorensenDice*) | 13 functions × 100 quick.Check iterations + 1000-iter determinism gate | All PASS |
| Benchmarks | 4 (ASCII Short/Medium/Long + Unicode Short) | All within RESEARCH.md §4.1 alloc budget |
| Fuzz (10s smoke per harness) | 2 harnesses (FuzzSorensenDiceScore, FuzzSorensenDiceScoreRunes) | No panics, no NaN/Inf, all in [0, 1] |
| BDD scenarios | 6 in sorensen_dice.feature | All PASS via `make test-bdd` (96 total scenarios / 210 steps green) |
| Dispatch tests (TestDispatch_SorensenDiceRegistered + TestDispatch_UnregisteredSlotsAreNil) | 2 | Both PASS; slot 10 now registered |
| Golden file (TestGolden_SorensenDice_Staging) | 1 (8 entries) | PASS — file canonical-marshalled, byte-stable |
| Examples (ExampleSorensenDiceScore, ExampleSorensenDiceScoreRunes) | 2 | Both PASS — Output blocks match byte-for-byte |
| TestAIFriendly_LLMSTxtReferencesEveryExportedSymbol | 1 | PASS — llms.txt sync verified |

## Determinism Gates

- **No `init()` functions** in `sorensen_dice.go` or `dispatch_sorensen_dice.go` — verified via `grep -q "^func init"` per-task verification.
- **No transcendental floats** in `sorensen_dice.go` — only `+`, `-`, `*`, `/`, comparisons, and `float64()` casts. Verified via `grep -E "math\.(Pow|Log|Exp|FMA)"`.
- **Explicit left-to-right parenthesisation** in the DSC formula: `(2.0 * float64(intersection)) / (float64(lenA) + float64(lenB))` — the inner parens force the addition before the division; the outer numerator parens force the multiplication before the division. No FMA / associativity ambiguity across CI platforms (DET-06).
- **No map iteration on output paths** in `sorensen_dice.go` — the intersection cardinality is computed by walking the smaller multiset and accumulating `min(cs, cl)` into an INTEGER counter; the per-side totals (lenA, lenB) are computed by summing map values into INTEGER counters. Integer addition is associative, so the SUMS are deterministic regardless of iteration order. The OUTPUT is a single float64 derived from integer counts.
- **Cross-platform float determinism**: integer-derived `(2.0 * float64(intersection)) / (float64(lenA) + float64(lenB))` is IEEE-754 correctly rounded on all four CI platforms; both numerator and denominator fit exactly in float64 for any input where `len(a) + len(b) < 2^53`.
- **Per-process determinism**: `TestProp_SorensenDiceScore_DeterministicAcrossRuns` runs 1000 sequential iterations on the RV-D1 night/nacht byte pair AND the café/cafe rune pair, comparing via `math.Float64bits` for bit-level equality.

## Deviations from Plan

None — the plan was executed exactly as written. RV-D1..RV-D4 reference vectors landed at the documented values. The structural sibling pattern from plan 05-01 (qgram_jaccard.go / qgram_jaccard_test.go / qgram_jaccard_bench_test.go / qgram_jaccard_fuzz_test.go / props_test.go append-block / qgram_jaccard.feature) was applied verbatim with the algorithm-specific substitutions; no rediscovery work was needed.

## Forward References

- **Plan 05-03 (Cosine)** consumes `extractQGrams` / `extractQGramsRunes` and adds the `sort.Strings(keys)` + sorted-iteration dot-product loop per CONTEXT.md §3 LOCKED. Cosine is the load-bearing cross-platform determinism algorithm; the plan 05-01 q-gram foundation + this plan's structural-sibling pattern are the prerequisites.
- **Plan 05-04 (Tversky)** consumes the same helpers and uses `ErrInvalidTverskyParam` (declared in plan 05-01) for its α/β validation in the future Phase 8 Scorer option layer. Tversky's Symmetric property fires only when α == β so plan 05-04 will add an asymmetry property test for the α ≠ β case (departure from this plan's exact-symmetry test).
- **Plan 05-05 (Finalisation)** merges this plan's `testdata/golden/_staging/sorensen_dice.json` into the canonical `testdata/golden/algorithms.json` via `TestGolden_Algorithms_Merge`'s stagingFiles slice.
- **Phase 6 (Monge-Elkan / Token Ratios)** consumes `AlgoSorensenDice` (now dispatched) as a permitted inner metric for `MongeElkanScore`.
- **Phase 8 (Scorer)** uses `ErrInvalidQGramSize` (declared in plan 05-01) for `WithSorensenDiceAlgorithm(weight, n)` parameter validation via `errors.Is`.

## Self-Check: PASSED

- `sorensen_dice.go` exists at `sorensen_dice.go` (FOUND).
- `dispatch_sorensen_dice.go` exists at `dispatch_sorensen_dice.go` (FOUND).
- `sorensen_dice_test.go` exists at `sorensen_dice_test.go` (FOUND).
- `sorensen_dice_bench_test.go` exists at `sorensen_dice_bench_test.go` (FOUND).
- `sorensen_dice_fuzz_test.go` exists at `sorensen_dice_fuzz_test.go` (FOUND).
- `testdata/golden/_staging/sorensen_dice.json` exists (FOUND, 8 entries, canonical-marshalled).
- `testdata/fuzz/FuzzSorensenDiceScore/seed-001` exists (FOUND, byte-stable `go test fuzz v1` format).
- `testdata/fuzz/FuzzSorensenDiceScoreRunes/seed-001` exists (FOUND, byte-stable format).
- `tests/bdd/features/sorensen_dice.feature` exists (FOUND, 6 scenarios).
- `algoid_test.go` updated with `TestDispatch_SorensenDiceRegistered` and slot 10 in `TestDispatch_UnregisteredSlotsAreNil` (FOUND).
- `algorithms_golden_test.go` updated with `buildSorensenDiceStagingEntries` + `TestGolden_SorensenDice_Staging` (FOUND).
- `props_test.go` updated with the 13 Sørensen-Dice property tests (FOUND).
- `example_test.go` updated with `ExampleSorensenDiceScore` + `ExampleSorensenDiceScoreRunes` (FOUND).
- `tests/bdd/steps/algorithms_steps.go` updated with 4 step methods + 4 ctx.Step registrations (FOUND).
- `llms.txt` updated with Sørensen-Dice section (FOUND).
- `llms-full.txt` updated with parallel Phase 5 Sørensen-Dice block (FOUND).
- Commit `0dbeab8` (Task 1) exists (FOUND in git log).
- Commit `3469fc1` (Task 2) exists (FOUND in git log).
- Commit `c1bc8f9` (Task 3) exists (FOUND in git log).
- `make test-bdd` exits 0 (VERIFIED — 96 scenarios / 210 steps all PASS, including 6 new SorensenDice scenarios).
- `bash scripts/verify-license-headers.sh` exits 0 (VERIFIED — 92 .go files carry the Apache-2.0 header).
- `bash scripts/verify-no-runtime-deps.sh` exits 0 (VERIFIED — root go.mod allowlist clean: 2 non-indirect modules — github.com/axonops/fuzzymatch + golang.org/x/text).
- `! grep -q "^func init" sorensen_dice.go` (VERIFIED — no init() in the algorithm file).
- `! grep -q "^func init" dispatch_sorensen_dice.go` (VERIFIED — dispatch uses var _ idiom).
- `grep -q "Source: Dice, L. R. (1945)" sorensen_dice.go` (VERIFIED — primary source citation present).
- `grep -v '^//' sorensen_dice.go | grep -E "math\.(Pow|Log|Exp|FMA)"` returns nothing (VERIFIED — no transcendentals in the algorithm path).
- `grep -q "SorensenDiceScore" llms.txt && grep -q "SorensenDiceScoreRunes" llms.txt` (VERIFIED — both new functions listed).
