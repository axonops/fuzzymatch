---
phase: 05-q-gram-algorithms
plan: 01
subsystem: similarity-algorithms
tags: [q-gram-foundation, ukkonen-1992, jaccard-1912, multiset-jaccard, error-sentinels, extract-qgrams, byte-and-rune-paths, dispatch-registration, default-n-3, property-tests, fuzz, benchmark, bdd, staging-golden, llms-sync, oq-2-resolution, no-init, no-transcendentals, no-map-iteration-on-output]

# Dependency graph
requires:
  - phase: 02-core-character-algorithms-six
    provides: assertGoldenStaging helper + goldenAlgorithmEntry / goldenAlgorithmsFile schema; AlgoQGramJaccard slot 9 already declared in algoid.go (lines 109-112) with String() case at line 233-234; export_test.go re-export pattern; testdata/golden/_staging/ workflow
  - phase: 03-smith-waterman-gotoh
    provides: per-algorithm BDD feature + step-bindings append pattern; props_test.go append-block pattern; ExampleXxxScore append-only example_test.go discipline; testdata/fuzz/Fuzz<Algo>Score/seed-001 byte-stable format (IN-06 closure); BDD score regex (\d+\.?\d*) integer-form acceptance (IN-03 closure); fuzz harness exercises full public surface (WR-02 closure)
  - phase: 04-remaining-character-gestalt
    provides: 04-01 strcmp95 pattern for the 3-task algorithm shape (impl+test+staging / property+bench+fuzz / BDD); per-plan llms.txt + llms-full.txt sync discipline (NOT deferred to finalisation — caught mid-flight in 04-01); 7-entry staging-golden norm; OQ-X RESOLUTION LOCKED format
provides:
  - QGramJaccardScore(a, b string, n int) float64 — multiset Jaccard over q-grams (byte path; dispatched with default n=3)
  - QGramJaccardScoreRunes(a, b string, n int) float64 — rune-path variant for multi-byte UTF-8
  - extractQGrams(s string, n int) map[string]int — UNEXPORTED shared helper (q_gram.go); single source of truth for plans 05-02 (Sørensen-Dice), 05-03 (Cosine), 05-04 (Tversky)
  - extractQGramsRunes(s string, n int) map[string]int — UNEXPORTED rune-path helper
  - dispatch[AlgoQGramJaccard] slot 9 populated via var-init (no init()); default n=3 trigram closure
  - var ErrInvalidQGramSize = errors.New("fuzzymatch: invalid q-gram size") — declared in errors.go (was MISSING; OQ-2 BLOCKER closed)
  - var ErrInvalidTverskyParam = errors.New("fuzzymatch: invalid tversky parameter") — declared in errors.go (was MISSING; OQ-2 BLOCKER closed)
  - testdata/golden/_staging/qgram_jaccard.json — 8 entries covering RV-J1 (Ukkonen 1992 §3 worked example), RV-J3, RV-J4, identical, both-empty, one-empty, n-too-large, plus the rune-path café canary
  - testdata/fuzz/FuzzQGramJaccardScore/seed-001 + testdata/fuzz/FuzzQGramJaccardScoreRunes/seed-001 — RV-J1 canonical seed in `go test fuzz v1` literal format
  - tests/bdd/features/qgram_jaccard.feature — 6 scenarios (1 outline + 5 standalone)
  - ExtractQGramsForTest + ExtractQGramsRunesForTest — re-exports in export_test.go for q_gram_test.go's black-box access to the unexported helpers
affects: [05-02-sorensen-dice, 05-03-cosine, 05-04-tversky, 05-05-finalisation]

# Tech tracking
tech-stack:
  added: []
  patterns:
    - "Shared internal extractor with re-export-via-export_test.go for tests. The q-gram tier ships a single unexported extractQGrams + extractQGramsRunes pair in q_gram.go consumed by all four downstream algorithms (CONTEXT.md §2 LOCKED — public surface is exactly the four algorithms' Score/ScoreRunes functions). Black-box tests access the helpers via ExtractQGramsForTest / ExtractQGramsRunesForTest re-exports in export_test.go — the canonical Go pattern for testing package internals from package fuzzymatch_test."
    - "Multiset semantics with capacity-hinted maps. extractQGrams uses make(map[string]int, len(s)-n+1) sized to the worst-case distinct q-gram count; extractQGramsRunes uses runeCount-n+1. The substring s[i:i+n] is a slice header into the input — no per-key string copy on the heap; only map hash bookkeeping allocates. Repeated q-grams accumulate (\"AAAA\"/n=2 → {AA:3}); the multiset cardinality is len(s)-n+1 for non-degenerate inputs."
    - "Dispatch wrapper closure for parameterised algorithms. The dispatch table maps AlgoID to (a, b string) float64 — no place for the n parameter. dispatch_qgram_jaccard.go binds default n=3 via `dispatch[AlgoQGramJaccard] = func(a, b string) float64 { return QGramJaccardScore(a, b, 3) }`. Specific n overrides happen via the Phase 8 Scorer option WithQGramJaccardAlgorithm(weight, n) per CONTEXT.md Deferred §4. Same pattern will apply to plans 05-02 / 05-03 / 05-04."
    - "Direct-call panic + Scorer-returned-error split (CONTEXT.md §5 LOCKED). Direct QGramJaccardScore calls panic on n < 1 with the message \"fuzzymatch: invalid q-gram size\" — programmer error fails loudly. The Phase 8 Scorer option WithQGramJaccardAlgorithm will return ErrInvalidQGramSize via errors.Is for the same input. Both paths share the same sentinel text in their messages so consumers grep for one phrase."
    - "Multiset map iteration on integer accumulators is DET-03 safe. The intersection cardinality is computed by iterating the smaller multiset and accumulating min(countA[k], countB[k]) into an INTEGER counter — the OUTPUT is a scalar int, not an ordered slice. Map iteration order does not affect the integer sum (integer addition is associative). The same iteration discipline applies to the per-side total (totalA, totalB) computations."
    - "No transcendentals in the algorithm path (DET-06). Only +, -, *, /, comparisons, and float64() casts. The single division `float64(intersection) / float64(union)` produces byte-identical output across all four CI platforms — both numerator and denominator are integer-derived float64 values that fit exactly in float64 for any input where len(a)+len(b) < 2^53. No math.Sqrt / math.Pow / math.Log / math.Exp / math.FMA."

key-files:
  created:
    - q_gram.go
    - qgram_jaccard.go
    - dispatch_qgram_jaccard.go
    - q_gram_test.go
    - qgram_jaccard_test.go
    - qgram_jaccard_bench_test.go
    - qgram_jaccard_fuzz_test.go
    - testdata/golden/_staging/qgram_jaccard.json
    - testdata/fuzz/FuzzQGramJaccardScore/seed-001
    - testdata/fuzz/FuzzQGramJaccardScoreRunes/seed-001
    - tests/bdd/features/qgram_jaccard.feature
    - .planning/phases/05-q-gram-algorithms/05-01-qgram-foundation-jaccard-SUMMARY.md
  modified:
    - errors.go (append: ErrInvalidQGramSize + ErrInvalidTverskyParam alphabetically with the existing Err* cluster — OQ-2 BLOCKER RESOLUTION)
    - export_test.go (append: ExtractQGramsForTest + ExtractQGramsRunesForTest re-exports for black-box tests)
    - algorithms_golden_test.go (append: buildQGramJaccardStagingEntries + TestGolden_QGramJaccard_Staging — 8 alphabetically-sorted entries via assertGoldenStaging)
    - algoid_test.go (append: TestDispatch_QGramJaccardRegistered; slot 9 added to TestDispatch_UnregisteredSlotsAreNil registered map — now 11 currently-registered slots)
    - example_test.go (append: ExampleQGramJaccardScore + ExampleQGramJaccardScoreRunes runnable godoc examples)
    - props_test.go (append: 12 standard property tests across byte+rune surfaces — RangeBounds/Identity/Symmetric/NoNaN/NoInf/NoNegativeZero × 2 — PLUS TestProp_QGramJaccardScore_DeterministicAcrossRuns 1000-iteration math.Float64bits gate)
    - tests/bdd/steps/algorithms_steps.go (append: 4 step methods on AlgorithmContext + 4 ctx.Step regex registrations inside InitializeScenario; new `with n (\d+)` regex shape extends the Phase 2/3/4 grammar to carry the q-gram size)
    - llms.txt (append: Q-Gram Jaccard section with the 2 exported functions; ErrInvalidQGramSize + ErrInvalidTverskyParam in the sentinel-errors section — meta-test TestAIFriendly green-bar gate)
    - llms-full.txt (append: parallel Phase 5 algorithm-surface block + parallel sentinel-error blocks with one-line rationales)

key-decisions:
  - "OQ-2 RESOLUTION LOCKED 2026-05-15: ErrInvalidQGramSize + ErrInvalidTverskyParam added to errors.go in this plan. CONTEXT.md `code_context` claimed they were pre-declared per Phase 1 — that claim was inaccurate (verified by `grep -n 'ErrInvalidQGramSize|ErrInvalidTverskyParam' errors.go` before implementation; the file declared only ErrInvalidInput / ErrInvalidConfiguration / ErrInvalidAlgorithm / ErrEmptyInput). Resolution path (A) from RESEARCH.md §6: add the two sentinels alphabetically with the existing cluster, with godoc blocks following the existing Err* style. Plans 05-02 / 05-03 / 05-04 and the future Phase 8 Scorer can now rely on these without re-asking. The direct-call panic path uses the literal message text (\"fuzzymatch: invalid q-gram size\" / \"fuzzymatch: invalid tversky parameter\") rather than the sentinel value because panic recovery is text-based — the Scorer error path uses errors.Is on the sentinel."
  - "Default n = 3 trigram for the dispatch wrapper. CONTEXT.md \"Claude's Discretion\" Deferred §4 endorses this; trigrams are the most common q-gram default in the NLP literature (Ukkonen 1992 §5 also recommends n ≈ 3 for typical natural-language input). The wrapper closure pattern (`var _ = func() bool { dispatch[AlgoQGramJaccard] = func(a, b string) float64 { return QGramJaccardScore(a, b, 3) }; return true }()`) keeps the dispatch table's (a, b string) float64 signature uniform across algorithms; non-default-n consumers go through the Phase 8 Scorer option layer."
  - "Walk the SMALLER multiset for the intersection-cardinality computation. `if len(qb) < len(qa) { small, large = qb, qa }` followed by `for k, cs := range small { if cl, ok := large[k]; ok { intersection += min(cs, cl) } }`. Walking the smaller side keeps the lookup count to len(min(qa, qb)) rather than len(max(qa, qb)) — material on heavily-asymmetric input pairs. The intersection result is INVARIANT under this swap (set intersection is symmetric); the determinism property test verifies bit-for-bit stability across runs."
  - "Both-empty fallback inside jaccardFromQGramMaps. The algorithm-level a == b short-circuit handles len(a) == 0 && len(b) == 0 (both-empty == identical). But `n > min(len(a), len(b))` produces empty multisets on both sides without the inputs being identical (e.g. \"ab\" vs \"abc\" at n=5). The helper has a defensive `if len(qa) == 0 && len(qb) == 0 { return 1.0 }` guard so that path returns 1.0 by the both-empty convention rather than 0/0 NaN. The downstream `if union == 0 { return 1.0 }` provides a redundant secondary guard."
  - "Property-test n coercion via `(abs(n) % 5) + 1` — keep n in [1, 5]. The alternative is to filter via `if n < 1 { return true }` inside the property — but quick.Check would generate many out-of-domain inputs and the effective coverage would drop. The coerce-into-range pattern guarantees every drawn triple actually exercises the [0, 1] / identity / symmetry invariant. The n < 1 panic path is unit-tested separately by TestQGramJaccard_PanicsOnInvalidN."
  - "Property tests cover Symmetric (Q-Gram Jaccard IS symmetric, unlike Ratcliff-Obershelp which had it dropped per OQ-1 in plan 04-03). Set-Jaccard's symmetry is exact (bit-for-bit equality), so the test asserts `==` not `|x - y| < ε`. This pattern carries forward to Sørensen-Dice and Cosine in plans 05-02 / 05-03; Tversky's Symmetric property fires only when α == β so plan 05-04 adds an asymmetry property test for the α ≠ β case."

patterns-established:
  - "Pattern: per-plan re-exports in export_test.go for unexported helpers consumed by tests. The shared q-gram extractors are package-internal per CONTEXT.md §2 LOCKED, but q_gram_test.go needs to assert their multiset semantics directly (not just through the algorithm-level Score functions). The export_test.go re-export pattern (`var ExtractQGramsForTest = extractQGrams`) keeps the helpers internal to consumers while exposing them to black-box tests. Plans 05-02 / 05-03 / 05-04 inherit the helpers without needing to extend the re-export — q_gram_test.go in this plan is the authoritative test of the extractor."
  - "Pattern: per-plan llms.txt sync (carried forward from Phase 4 plan 04-01 precedent). The 2 new functions and 2 new sentinels are appended to llms.txt + llms-full.txt within this plan's commits, NOT deferred to plan 05-05 finalisation. The TestAIFriendly meta-test green-bar enforces this — a missing entry would fail CI immediately. Plans 05-02 / 05-03 / 05-04 will follow the same discipline."
  - "Pattern: dispatch wrapper closure for algorithms with extra parameters. dispatch[AlgoQGramJaccard] = `func(a, b string) float64 { return QGramJaccardScore(a, b, 3) }` instead of `dispatch[AlgoQGramJaccard] = QGramJaccardScore` (which would be a type error — wrong signature). The closure technique is the canonical bridge between the parameterised algorithm function and the fixed dispatch signature. Plans 05-02 / 05-03 / 05-04 will use the same shape with their own default-n / default-α-β values."
  - "Pattern: BDD step grammar extension for parameterised algorithms. The Phase 2/3/4 score-step regex `^I compute the X score between \"([^\"]*)\" and \"([^\"]*)\"$` does not carry the n parameter. Plan 05-01 introduces the `with n (\\d+)` suffix; plans 05-02 / 05-03 will use the same shape; plan 05-04 will add `alpha (\\d+\\.?\\d*) beta (\\d+\\.?\\d*)` for Tversky. The existing approximately-step regex `(\\d+\\.?\\d*)` is reused unchanged — no regex drift on the existing grammar."

requirements-completed: [QGRAM-01, QGRAM-02]

# Metrics
duration: ~45min
completed: 2026-05-15
---

# Phase 5 Plan 01: Q-Gram Foundation + Q-Gram Jaccard Summary

**Lay the q-gram tier foundation: ship the unexported `extractQGrams` / `extractQGramsRunes` helpers in `q_gram.go` (Ukkonen 1992 §2-3) and the first algorithmic consumer — Q-Gram Jaccard (Ukkonen 1992 §3 / Jaccard 1912) — with byte and rune surfaces, dispatch wrapper (default n=3), full property + bench + fuzz + BDD coverage, and the per-algorithm staging golden file. Also close OQ-2 BLOCKER by adding the two missing error sentinels (ErrInvalidQGramSize + ErrInvalidTverskyParam) that CONTEXT.md `code_context` mistakenly claimed were pre-declared.**

## Performance

- **Duration:** ~45 min (3 atomic commits)
- **Tasks:** 3 (all completed)
- **Files modified:** 22 (12 created, 10 appended/modified)
- **Commits:** 3 atomic commits, all reference issue #7 (the Phase 5 epic)

## Accomplishments

- **Two new exported functions:** `QGramJaccardScore`, `QGramJaccardScoreRunes` — pinned at docs/requirements.md §7.2.1.
- **Two new error sentinels:** `ErrInvalidQGramSize`, `ErrInvalidTverskyParam` — closes OQ-2 BLOCKER from RESEARCH.md §6 (CONTEXT.md `code_context` claim that they were pre-declared was inaccurate; verified by grep before implementation).
- **Shared internal extractor:** `extractQGrams` + `extractQGramsRunes` in `q_gram.go` — single source of truth for plans 05-02 / 05-03 / 05-04. Re-exported as `ExtractQGramsForTest` / `ExtractQGramsRunesForTest` in `export_test.go` for black-box test access.
- **Dispatch slot 9** (`AlgoQGramJaccard`) populated via `var _ = func() bool{...}()` (no init()). Default n=3 trigram per CONTEXT.md Deferred §4. `TestDispatch_UnregisteredSlotsAreNil` now expects 11 currently-registered slots; slots 10..21 await later plans (05-02..05-04 plus token + phonetic algorithms).
- **Reference vector verification:** Ukkonen 1992 §3 worked-example pin landed exactly:
  - `QGramJaccardScore("AGCT", "AGCTAGCT", 2)` = `3/7 = 0.42857142857142855` (RV-J1)
  - `QGramJaccardScore("abcd", "abxy", 2)` = `0.2` (RV-J4)
  - `QGramJaccardScore("abc", "xyz", 2)` = `0.0` (RV-J3)
  - `QGramJaccardScoreRunes("café", "cafe", 2)` = `0.5` (RV-J5-Runes)
- **Property-test coverage:** 12 standard invariants (RangeBounds / Identity / Symmetric / NoNaN / NoInf / NoNegativeZero × byte+rune surfaces) plus the `_DeterministicAcrossRuns` gate (1000 sequential iterations on RV-J1 + the rune café pair, math.Float64bits comparison).
- **Fuzz harnesses:** Two fuzzers (byte + rune), each with 11-12 programmatic seeds covering canonical reference vectors + identity + both-empty + one-empty + invalid UTF-8 + multi-byte + long-input + n=1 + n=8 boundary cases. On-disk seed files (`testdata/fuzz/FuzzQGramJaccardScore/seed-001`, `testdata/fuzz/FuzzQGramJaccardScoreRunes/seed-001`) in byte-stable `go test fuzz v1` literal format. 5-second smoke run on each completes without panic / NaN / Inf / out-of-range.
- **Benchmark coverage:** 4 benches (ASCII Short / Medium / Long + Unicode Short). Allocation count within RESEARCH.md §4.1 budget (≤ 4 on the canonical-source ideal for short inputs; ≤ 8 on the realistic ceiling for long inputs).
- **BDD scenarios:** 6 scenarios (1 outline + 5 standalone) in `tests/bdd/features/qgram_jaccard.feature` — covers RV-J1..RV-J4, identity, both-empty, one-empty (both directions), symmetry, and the rune-path café canary.
- **Per-algorithm staging golden:** `testdata/golden/_staging/qgram_jaccard.json` with 8 alphabetically-sorted entries (canonical-marshalled via `assertGoldenStaging` → `CanonicalMarshalForTest`). Plan 05-05 owns the merge into `algorithms.json`.

## Reference Vectors Verified

| ID | Inputs | n | Expected | Source |
|----|--------|---|----------|--------|
| RV-J1 | `"AGCT"` / `"AGCTAGCT"` | 2 | 3/7 ≈ 0.42857142857142855 | Ukkonen 1992 §3 worked example |
| RV-J2 | `"hello"` / `"hello"` | 2 | 1.0 | identity short-circuit |
| RV-J3 | `"abc"` / `"xyz"` | 2 | 0.0 | orthogonal q-gram sets |
| RV-J4 | `"abcd"` / `"abxy"` | 2 | 0.2 | single-shared bigram (1/5) |
| RV-J5-Runes | `"café"` / `"cafe"` | 2 | 0.5 | rune-path |∩|=2, |∪|=4 |
| RV-J6 | `"ab"` / `"abc"` | 5 | 1.0 | both q-gram views empty (n > min len) |

## Test Results

| Test category | Count | Status |
|---------------|-------|--------|
| Unit tests (TestExtractQGrams* + TestQGramJaccard*) | 22 sub-tests | All PASS |
| Property tests (TestProp_QGramJaccard*) | 13 functions × 100 quick.Check iterations + 1000-iter determinism gate | All PASS |
| Benchmarks | 4 (ASCII Short/Medium/Long + Unicode Short) | All within RESEARCH.md §4.1 alloc budget |
| Fuzz (5s smoke per harness) | 2 harnesses (FuzzQGramJaccardScore, FuzzQGramJaccardScoreRunes) | No panics, no NaN/Inf, all in [0, 1] |
| BDD scenarios | 6 in qgram_jaccard.feature | All PASS via `make test-bdd` |
| Dispatch tests (TestDispatch_QGramJaccardRegistered + TestDispatch_UnregisteredSlotsAreNil) | 2 | Both PASS; slot 9 now registered |
| Golden file (TestGolden_QGramJaccard_Staging) | 1 (8 entries) | PASS — file canonical-marshalled, byte-stable |
| Examples (ExampleQGramJaccardScore, ExampleQGramJaccardScoreRunes) | 2 | Both PASS — Output blocks match byte-for-byte |
| TestAIFriendly_LLMSTxtReferencesEveryExportedSymbol | 1 | PASS — llms.txt sync verified |

## Determinism Gates

- **No `init()` functions** in `q_gram.go` or `qgram_jaccard.go` — verified via `grep -q "^func init"` in CI script and per-task verification.
- **No transcendental floats** in `qgram_jaccard.go` — only `+`, `-`, `*`, `/`, comparisons, and `float64()` casts. Verified via `grep -E "math\\.(Pow|Log|Exp|FMA)"`.
- **No map iteration on output paths** in `q_gram.go` or `qgram_jaccard.go` — the intersection cardinality is computed by walking the smaller multiset and accumulating `min(cs, cl)` into an INTEGER counter; the OUTPUT is a single float64 derived from the integer counts. Map iteration order does not affect the result.
- **Cross-platform float determinism**: integer-derived `float64(intersection) / float64(union)` is IEEE-754 correctly rounded on all four CI platforms; both numerator and denominator fit exactly in float64 for any input where `len(a) + len(b) < 2^53`.
- **Per-process determinism**: `TestProp_QGramJaccardScore_DeterministicAcrossRuns` runs 1000 sequential iterations on the RV-J1 byte pair AND the café rune pair, comparing via `math.Float64bits` for bit-level equality.

## Deviations from Plan

### Auto-fixed Issues

**1. [Rule 1 — Bug] Naming clash with existing `itoa` helper**
- **Found during:** Task 1 unit tests
- **Issue:** Initial draft of `qgram_jaccard_test.go` defined a local `itoa(int) string` helper for panic-test sub-names; `golden_test.go` already declared an `itoa` at package scope. Build failed with `itoa redeclared in this block`.
- **Fix:** Switched to `strconv.Itoa` from the stdlib (added `strconv` to the test-file imports) and removed the local helper.
- **Files modified:** `qgram_jaccard_test.go`
- **Commit:** `e70bba3` (Task 1 commit; the fix landed before the commit was created — the commit is correct as-is)

No other deviations. The plan was executed exactly as written; OQ-2 was a known BLOCKER from RESEARCH.md and is recorded under "Key Decisions" rather than as a deviation.

## OQ-2 Resolution Locked 2026-05-15

> **Verified by `grep -n "ErrInvalid" errors.go` BEFORE implementation:** at the time of plan execution, `errors.go` declared only `ErrInvalidInput`, `ErrInvalidConfiguration`, `ErrInvalidAlgorithm`, and `ErrEmptyInput`. CONTEXT.md `code_context` (lines 161-162) claimed: *"`errors.go` — `ErrInvalidQGramSize` and `ErrInvalidTverskyParam` are declared at the package level per Phase 1"* — that claim was inaccurate. RESEARCH.md §6 OQ-2 surfaced the gap as a BLOCKER for plan 05-01 and recommended Resolution (A): add the two sentinels in this plan. Resolution (A) was implemented: both sentinels are now declared alphabetically with the existing `Err*` cluster, with godoc blocks following the same style as `ErrInvalidAlgorithm` / `ErrInvalidConfiguration` / `ErrInvalidInput`. Plans 05-02 / 05-03 / 05-04 and the future Phase 8 Scorer can now rely on these without re-asking; the meta-test `TestAIFriendly_LLMSTxtReferencesEveryExportedSymbol` passed after appending the two new sentinels to `llms.txt`.

## Forward References

- **Plan 05-02 (Sørensen-Dice)** consumes `extractQGrams` / `extractQGramsRunes` from `q_gram.go` and follows the same 3-task shape: impl + test + staging / property + bench + fuzz / BDD.
- **Plan 05-03 (Cosine)** consumes the same helpers and adds the `sort.Strings(keys)` + sorted-iteration dot-product loop per CONTEXT.md §3 LOCKED. Cosine is the load-bearing cross-platform determinism algorithm; this plan's `q_gram.go` is the foundation that gate depends on.
- **Plan 05-04 (Tversky)** consumes the same helpers and uses `ErrInvalidTverskyParam` (declared in this plan) for its α/β validation in the Phase 8 Scorer option layer.
- **Plan 05-05 (Finalisation)** merges this plan's `testdata/golden/_staging/qgram_jaccard.json` into the canonical `testdata/golden/algorithms.json` via `TestGolden_Algorithms_Merge`'s stagingFiles slice.
- **Phase 6 (Monge-Elkan / Token Ratios)** consumes `AlgoQGramJaccard` (now dispatched) as a permitted inner metric for `MongeElkanScore`.
- **Phase 8 (Scorer)** uses `ErrInvalidQGramSize` (declared in this plan) for `WithQGramJaccardAlgorithm(weight, n)` parameter validation via `errors.Is`.

## Self-Check: PASSED

- `q_gram.go` exists at `q_gram.go` (FOUND).
- `qgram_jaccard.go` exists at `qgram_jaccard.go` (FOUND).
- `dispatch_qgram_jaccard.go` exists at `dispatch_qgram_jaccard.go` (FOUND).
- `q_gram_test.go` exists at `q_gram_test.go` (FOUND).
- `qgram_jaccard_test.go` exists at `qgram_jaccard_test.go` (FOUND).
- `qgram_jaccard_bench_test.go` exists at `qgram_jaccard_bench_test.go` (FOUND).
- `qgram_jaccard_fuzz_test.go` exists at `qgram_jaccard_fuzz_test.go` (FOUND).
- `testdata/golden/_staging/qgram_jaccard.json` exists (FOUND, 8 entries, canonical-marshalled).
- `testdata/fuzz/FuzzQGramJaccardScore/seed-001` exists (FOUND, byte-stable format).
- `testdata/fuzz/FuzzQGramJaccardScoreRunes/seed-001` exists (FOUND, byte-stable format).
- `tests/bdd/features/qgram_jaccard.feature` exists (FOUND, 6 scenarios).
- `errors.go` updated with `ErrInvalidQGramSize` + `ErrInvalidTverskyParam` (FOUND).
- `export_test.go` updated with `ExtractQGramsForTest` + `ExtractQGramsRunesForTest` (FOUND).
- `algoid_test.go` updated with `TestDispatch_QGramJaccardRegistered` and slot 9 in `TestDispatch_UnregisteredSlotsAreNil` (FOUND).
- `algorithms_golden_test.go` updated with `buildQGramJaccardStagingEntries` + `TestGolden_QGramJaccard_Staging` (FOUND).
- `props_test.go` updated with the 13 Q-Gram Jaccard property tests (FOUND).
- `example_test.go` updated with `ExampleQGramJaccardScore` + `ExampleQGramJaccardScoreRunes` (FOUND).
- `tests/bdd/steps/algorithms_steps.go` updated with 4 step methods + 4 ctx.Step registrations (FOUND).
- `llms.txt` updated with QGramJaccard section + new sentinel entries (FOUND).
- `llms-full.txt` updated with parallel Phase 5 algorithm-surface block + sentinel-error blocks (FOUND).
- Commit `e70bba3` (Task 1) exists (FOUND in git log).
- Commit `d0e2657` (Task 2) exists (FOUND in git log).
- Commit `55d494e` (Task 3) exists (FOUND in git log).
- `make test-bdd` exits 0 (VERIFIED).
- `bash scripts/verify-license-headers.sh` exits 0 (VERIFIED — 87 .go files carry the Apache-2.0 header).
- `bash scripts/verify-no-runtime-deps.sh` exits 0 (VERIFIED — root go.mod allowlist clean).
