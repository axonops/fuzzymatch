---
phase: 05-q-gram-algorithms
plan: 03
subsystem: similarity-algorithms
tags: [cosine, salton-mcgill-1983, vector-space, q-gram-frequency-vectors, float-determinism, sorted-keys, math-sqrt, fma-risk-documented, load-bearing-determinism, byte-and-rune-paths, dispatch-registration, default-n-3, hand-derived-rv-c1-to-c5, property-tests, fuzz, benchmark, bdd, staging-golden, llms-sync, no-init, no-transcendentals-except-sqrt, ieee754-clamp, rule-1-deviation-clamp]

# Dependency graph
requires:
  - phase: 02-core-character-algorithms-six
    provides: assertGoldenStaging helper + goldenAlgorithmEntry / goldenAlgorithmsFile schema; AlgoCosine slot 11 already declared in algoid.go (lines 119-122) with String() case at line 237-238; export_test.go re-export pattern; testdata/golden/_staging/ workflow
  - phase: 03-smith-waterman-gotoh
    provides: per-algorithm BDD feature + step-bindings append pattern; props_test.go append-block pattern; ExampleXxxScore append-only example_test.go discipline; testdata/fuzz/Fuzz<Algo>Score/seed-001 byte-stable format (IN-06 closure); BDD score regex (\d+\.?\d*) integer-form acceptance (IN-03 closure); fuzz harness exercises full public surface (WR-02 closure)
  - phase: 04-remaining-character-gestalt
    provides: 04-01 strcmp95 pattern for the 3-task algorithm shape (impl+test+staging / property+bench+fuzz / BDD); per-plan llms.txt + llms-full.txt sync discipline (NOT deferred to finalisation — caught mid-flight in 04-01); per-algorithm staging-golden norm (8-12 entries); OQ-X RESOLUTION LOCKED format
  - plan: 05-q-gram-algorithms/05-01-qgram-foundation-jaccard
    provides: q_gram.go with extractQGrams + extractQGramsRunes (the unexported shared helpers consumed here); ErrInvalidQGramSize error sentinel; the 3-task shape (impl/test/staging — property/bench/fuzz — BDD); fuzzCoerceN helper (in qgram_jaccard_fuzz_test.go) reused by FuzzCosineScore + FuzzCosineScoreRunes; AlgorithmContext.lastScore / lastScore2 fields (carried forward; same plan 05-01 introduction) reused by Cosine BDD steps
provides:
  - CosineScore(a, b string, n int) float64 — Cosine similarity over q-gram frequency vectors (Salton & McGill 1983 §4.1 eq. 4.4 p.121; byte path; dispatched with default n=3)
  - CosineScoreRunes(a, b string, n int) float64 — rune-path variant for multi-byte UTF-8
  - dispatch[AlgoCosine] slot 11 populated via var-init (no init()); default n=3 trigram closure
  - testdata/golden/_staging/cosine.json — 9 alphabetically-sorted entries spanning ASCII + Unicode at n ∈ {2, 3, 4} per CONTEXT.md §1 LOCKED slate (RV-C1 ascii_n2_irrational, RV-C2 ascii_n3_large_intersection, RV-C4 ascii_n4_exact, both_empty, identical, one_empty, orthogonal, RV-C3 unicode_n2_runes, unicode_n3_runes)
  - testdata/fuzz/FuzzCosineScore/seed-001 + testdata/fuzz/FuzzCosineScoreRunes/seed-001 — byte-stable seed corpus in `go test fuzz v1` literal format (RV-C1 / RV-C3 canonical)
  - tests/bdd/features/cosine.feature — 8 scenarios (1 outline with 6 examples + 7 standalone scenarios)
  - 4 cosine test/example/llms append-only contributions to existing accumulator files (props_test.go +13 property tests, example_test.go +2 examples, algoid_test.go +1 dispatch test +1 slot in registered map, algorithms_golden_test.go +buildCosineStagingEntries +TestGolden_Cosine_Staging, llms.txt +1 section, llms-full.txt +1 section)
affects: [05-05-finalisation]

# Tech tracking
tech-stack:
  added: []
  patterns:
    - "Sorted intersection-key dot-product loop (CONTEXT.md §3 LOCKED). The intersection key slice is built by walking the smaller multiset and checking membership in the larger; the slice is then sort.Strings'd in place BEFORE the dot-product reduction iterates it. The sorted slice makes the float reduction order canonical (byte-lexicographic, total, deterministic across platforms — Go's sort.Strings uses < on string per the language spec). This is the load-bearing cross-platform determinism gate for Cosine, the textbook map-iteration burnpit for vector-space algorithms — securing it with one explicit sort line is simpler than parallel-structure tracking."
    - "Factorised Salton & McGill 1983 cosine form: dot / (sqrt(normASq) * sqrt(normBSq)). Explicit (normA * normB) parenthesisation matches the 1983 textbook factorisation and avoids the int64-overflow trap of computing math.Sqrt(float64(sumSquaresA * sumSquaresB)) on long inputs (RESEARCH.md \"Pitfall 6\"). The factorised form's IEEE-754 actual differs by 1 ULP from the rational ideal (e.g. 2/sqrt(6) algebraically = 0.8164965809277261; the implementation produces 0.81649658092772592 because sqrt(2)*sqrt(3) = 2.4494897427831783, then 2.0/2.4494897427831783 = 0.81649658092772592). Test reference vectors and the staging golden pin the actual factorised-form output per RESEARCH.md \"Pitfall 2\" — the implementation is correct; the rational ideal is the textbook target but not the float64 reality."
    - "Inline FMA-risk footnote per CONTEXT.md §3 LOCKED + RESEARCH.md §3 OQ-1. The cosine.go dot-product loop has a code comment documenting that Go 1.26 may emit FMA on arm64 for the (x*y)+z pattern; parentheses do NOT defeat FMA fusion per golang/go#17895. The cross-platform CI matrix gate (testdata/golden/algorithms.json after plan 05-05 merge) is the load-bearing detector. Remediation if matrix divergence ever appears: insert an explicit float64 cast `dot = float64(float64(qa[k]) * float64(qb[k])) + dot` per RESEARCH.md §3.4. Documentation only — no code change required while CI passes."
    - "math.Sqrt is the only math.* call (DET-06). NO math.Pow, NO math.Log, NO math.Exp, NO math.FMA. math.Sqrt is intrinsified to SQRTSD on amd64 and FSQRTD on arm64 (both IEEE-754 §5.4.1 correctly rounded per RESEARCH.md §3.5) — same int input → byte-identical float64 output across all four CI platforms (linux/amd64, linux/arm64, darwin/arm64, windows/amd64)."
    - "Sum-of-squares norm reductions iterate map values in any order (DET-03 safe). The norm computation is `for _, c := range qa { sumSquaresA += c * c }` — integer addition is exactly associative regardless of map iteration order; the OUTPUT of THIS loop is a scalar int, not an ordered slice. The float-determinism risk (and therefore the sort.Strings discipline) lives only on the dot-product reduction. RESEARCH.md §3.7 confirms left-to-right reduction is the project-canonical pattern and integer associativity covers the norm path."
    - "IEEE-754 clamp on the final cosine result (Rule 1 deviation — see Deviations section below). The mathematical invariant cos(A, B) ∈ [0, 1] for non-negative count vectors holds (Cauchy-Schwarz: dot(A, B) ≤ ‖A‖·‖B‖); IEEE-754 rounding can push the result a few ULP above 1.0 in degenerate cases where QA == QB as multisets (e.g. CosineScore(\"cba\", \"abc\", 1) — algebraic limit is exactly 1.0 but the factorised form yields 1.0000000000000002 because sqrt(3)*sqrt(3) = 2.9999999999999996). The clamp `if cos > 1.0 { return 1.0 }; if cos < 0.0 { return 0.0 }` preserves the public-API [0.0, 1.0] contract. Surfaced by FuzzCosineScore on the 10s smoke run during plan 05-03 task 2."

key-files:
  created:
    - cosine.go
    - dispatch_cosine.go
    - cosine_test.go
    - cosine_bench_test.go
    - cosine_fuzz_test.go
    - testdata/golden/_staging/cosine.json
    - testdata/fuzz/FuzzCosineScore/seed-001
    - testdata/fuzz/FuzzCosineScoreRunes/seed-001
    - tests/bdd/features/cosine.feature
    - .planning/phases/05-q-gram-algorithms/05-03-cosine-SUMMARY.md
  modified:
    - algorithms_golden_test.go (append: buildCosineStagingEntries + TestGolden_Cosine_Staging — 9 alphabetically-sorted entries via assertGoldenStaging)
    - algoid_test.go (append: TestDispatch_CosineRegistered; slot 11 added to TestDispatch_UnregisteredSlotsAreNil registered map — now 13 currently-registered slots)
    - example_test.go (append: ExampleCosineScore + ExampleCosineScoreRunes runnable godoc examples; output blocks pin the actual factorised-form 16-digit output)
    - props_test.go (append: 12 standard property tests across byte+rune surfaces — RangeBounds/Identity/Symmetric/NoNaN/NoInf/NoNegativeZero × 2 — PLUS TestProp_CosineScore_DeterministicAcrossRuns 1000-iteration math.Float64bits gate on RV-C2 + RV-C3)
    - tests/bdd/steps/algorithms_steps.go (append: 4 step methods on AlgorithmContext + 4 ctx.Step regex registrations inside InitializeScenario)
    - llms.txt (append: Cosine section with the 2 exported functions; meta-test TestAIFriendly green-bar gate)
    - llms-full.txt (append: parallel Phase 5 algorithm-surface block with the IEEE-754 precision discussion footnoted)
    - go.sum (incidental: Go toolchain refreshed dependency hashes during build)

key-decisions:
  - "RV-CN reference vector pinning: pin actual IEEE-754 factorised-form output, not rational limits (RESEARCH.md \"Pitfall 2\"). The hand-derivation comment blocks reproduce the Salton & McGill 1983 §4.1 rational form (e.g. 2/sqrt(6) for RV-C1) AND document the IEEE-754 actual (e.g. 0.81649658092772592 — 1 ULP shortfall from 0.8164965809277261). Reviewers can re-derive the rational form against Salton & McGill 1983 in <30s; the IEEE-754 actual is what the implementation produces and what the test pins. Without this pinning approach, every Cosine reference vector would be 1 ULP off from the test expectation. RV-C5 happens to coincide with the rational ideal (1/sqrt(3) = sqrt(1)*sqrt(3) factorised form has no last-bit drift); the others are all 1 ULP from the rational limit due to sqrt(x)*sqrt(y) rounding."
  - "RV-C4 (the 'exact' n=4 vector from CONTEXT.md / RESEARCH.md): pin to 0.4999999999999999 (= 0.49999999999999989 — same float64 bits), not 0.5 exactly. The plan / CONTEXT.md / RESEARCH.md all describe RV-C4 as 'exactly representable in float64 (0.5 = 2^-1)' — that's the rational ideal. The IEEE-754 actual from the factorised form `1.0 / (math.Sqrt(2) * math.Sqrt(2))` is 0.49999999999999989 because math.Sqrt(2) * math.Sqrt(2) = 2.0000000000000004 (not exact 2). The implementation correctly follows the Salton & McGill 1983 §4.1 factorised form per the determinism discipline; the 1-ULP shortfall is a property of the formula, not a bug. Test cosine_test.go RV-C4 derivation block documents this thoroughly with the IEEE-754 reasoning."
  - "Rule 1 deviation — IEEE-754 clamp on final cosine. FuzzCosineScore surfaced CosineScore(\"cba\", \"abc\", 1) = 1.0000000000000002 (1 ULP above the public-API [0.0, 1.0] contract). Root cause: when QA == QB as multisets (e.g. permutations of the same character set), the algebraic limit is exactly 1.0 but `dot / (sqrt(normASq) * sqrt(normBSq))` = `3.0 / 2.9999999999999996` = `1.0000000000000002`. Fix: clamp the final result to [0.0, 1.0] (`if cos > 1.0 { return 1.0 }; if cos < 0.0 { return 0.0 }`). The clamp is widely accepted for cosine implementations — the algebraic invariant cos(A, B) ∈ [0, 1] holds (Cauchy-Schwarz); IEEE-754 rounding is the source of drift, not the algorithm. Without the clamp, the [0,1] property tests would fail on any same-multiset input. Documented inline in cosineFromQGramMaps with a reference to FuzzCosineScore as the surface."
  - "Default n = 3 trigram for the dispatch wrapper (carried forward from plan 05-01 / 05-02). CONTEXT.md \"Claude's Discretion\" Deferred §4 endorses; trigrams are the most common q-gram default in the IR / NLP literature (Salton & McGill 1983 §4 also recommends n ≈ 3 for typical natural-language input). Same wrapper closure pattern as plans 05-01 / 05-02. Phase 8 Scorer option WithCosineAlgorithm(weight, n) will expose non-default n."
  - "Cosine staging-golden has 9 entries (vs Q-Gram Jaccard's 8 and Sørensen-Dice's 8). The +1 captures CONTEXT.md §1 LOCKED requirement that algorithms.json entries span ASCII + Unicode at n ∈ {2, 3, 4} for Cosine specifically (the cross-platform float-determinism gate). Entries: Cosine_ascii_n2_irrational (RV-C1), Cosine_ascii_n3_large_intersection (RV-C2), Cosine_ascii_n4_exact (RV-C4), Cosine_both_empty, Cosine_identical, Cosine_one_empty, Cosine_orthogonal, Cosine_unicode_n2_runes (RV-C3), Cosine_unicode_n3_runes. Each ExpectedScore is computed from the live implementation (`fuzzymatch.CosineScore(a, b, n)`) so the staging file stays in sync with actual output."
  - "Property-test n coercion via `(abs(n) % 5) + 1` — keep n in [1, 5] (carried forward from plan 05-01 / 05-02). The alternative of filtering via `if n < 1 { return true }` would generate many out-of-domain inputs and drop effective coverage. The coerce-into-range pattern guarantees every drawn triple actually exercises the [0, 1] / identity / symmetry invariant. The n < 1 panic path is unit-tested separately by TestCosine_PanicsOnInvalidN."
  - "Property tests cover Symmetric (Cosine IS symmetric — sorted-key iteration is canonical regardless of input order, so cos(A,B) == cos(B,A) bit-for-bit). The test asserts `==` not `|x - y| < ε`. This pattern is consistent with plans 05-01 (Q-Gram Jaccard) and 05-02 (Sørensen-Dice); plan 05-04 (Tversky) will diverge — Tversky's Symmetric property fires only when α == β, so plan 05-04 adds an asymmetry property test for the α ≠ β case."

patterns-established:
  - "Pattern: pin IEEE-754 factorised-form output for irrational reference vectors, not rational limits. The hand-derivation comment blocks document the Salton & McGill 1983 §4.1 rational form for reviewer-verification, then document the IEEE-754 actual (1 ULP from the rational limit, because sqrt(x)*sqrt(y) is not exactly correctly rounded — only sqrt itself is). The test pin is the IEEE-754 actual. This pattern carries forward to plan 05-04 (Tversky's IEEE-754 division paths) and any future Cosine variants."
  - "Pattern: clamp final score to [0.0, 1.0] for algorithms that combine math.Sqrt with division. Cosine, Tversky (when α + β > 0 produces a denominator < numerator due to rounding), and any future algorithm whose final form is `numerator / (math.Sqrt(...) * math.Sqrt(...))` should clamp to preserve the public-API range contract. The fuzz harness is the regression detector for these IEEE-754 boundary cases. Documented inline at the clamp site with a reference to the surfacing fuzzer."
  - "Pattern: load-bearing per-algorithm determinism gate via TestProp_<Algo>_DeterministicAcrossRuns + TestProp_<Algo>_SortedKeyIteration. The 1000-iteration math.Float64bits gate is the run-time regression detector for any future refactor that re-introduces map-iteration order dependence on the output path (CONTEXT.md §3 LOCKED). The gate is per-process (within a single test run); the cross-platform CI matrix gate (testdata/golden/algorithms.json after plan 05-05 merge) is the cross-platform regression detector. Both layers are needed."
  - "Pattern: per-plan llms.txt + llms-full.txt sync (carried forward from Phase 4 plan 04-01 precedent and plans 05-01 / 05-02). The 2 new functions are appended to llms.txt + llms-full.txt within this plan's commits, NOT deferred to plan 05-05 finalisation. The TestAIFriendly meta-test green-bar enforces this — a missing entry would fail CI immediately."
  - "Pattern: BDD step grammar reuse for parameterised q-gram algorithms. The plan 05-01 `with n (\\d+)` suffix grammar carries forward unchanged; the existing approximately-step regex `(\\d+\\.?\\d*)` accepts the 17-digit IEEE-754 form used by the high-precision examples in cosine.feature without modification (IN-03 closure carried forward). Plan 05-04 (Tversky) will extend the grammar with `alpha (\\d+\\.?\\d*) beta (\\d+\\.?\\d*)` for the α/β parameters."

requirements-completed: [QGRAM-04]

# Metrics
duration: ~26min
completed: 2026-05-15
---

# Phase 5 Plan 03: Cosine n-gram Similarity Summary

**Ship the textbook vector-space Cosine similarity over q-gram frequency vectors (Salton & McGill 1983 §4.1 eq. 4.4 p.121) — the LOAD-BEARING cross-platform float-determinism algorithm for Phase 5 — with byte and rune surfaces, sort.Strings dispatch on the intersection key slice (CONTEXT.md §3 LOCKED), explicit (x*y)+z parenthesisation per DET-06, math.Sqrt-only norm computation (IEEE-754 correctly rounded per RESEARCH.md §3.5), inline FMA-risk footnote (RESEARCH.md §3 OQ-1), 5 hand-derived RV-C1..RV-C5 reference vectors at IEEE-754 factorised-form precision (per RESEARCH.md "Pitfall 2"), full property + bench + fuzz + BDD coverage, and the per-algorithm staging golden file with 9 alphabetically-sorted entries spanning ASCII + Unicode at n ∈ {2, 3, 4} per CONTEXT.md §1 LOCKED.**

## Performance

- **Duration:** ~26 min (3 atomic commits)
- **Tasks:** 3 (all completed)
- **Files modified:** 17 (10 created, 7 appended/modified — counting go.sum)
- **Commits:** 3 atomic commits, all reference issue #7 (the Phase 5 epic)

## Accomplishments

- **Two new exported functions:** `CosineScore`, `CosineScoreRunes` — pinned at docs/requirements.md §7.2.3.
- **Dispatch slot 11** (`AlgoCosine`) populated via `var _ = func() bool{...}()` (no init()). Default n=3 trigram per CONTEXT.md Deferred §4. `TestDispatch_UnregisteredSlotsAreNil` now expects 13 currently-registered slots; slots 12..21 await later plans (Tversky 05-04, plus token + phonetic algorithms).
- **Five hand-derived RV-C1..RV-C5 reference vectors** in `cosine_test.go` with 17-significant-digit IEEE-754 derivations in test comments per CONTEXT.md §4 LOCKED:
  - **RV-C1** ascii n=2 irrational: `CosineScore("abc", "abcd", 2) = 0.81649658092772592` (factorised-form actual; rational limit 2/sqrt(6) = 0.8164965809277261 is 1 ULP higher)
  - **RV-C2** ascii n=3 large intersection (5 keys): `CosineScore("abcdefgh", "abcdefgi", 3) = 0.83333333333333348` (rational limit 5/6 = 0.83333333333333337 is 1 ULP lower)
  - **RV-C3** rune-path Unicode: `CosineScoreRunes("café", "cafe", 2) = 0.66666666666666674` (rational limit 2/3 = 0.66666666666666663 is 1 ULP lower)
  - **RV-C4** ascii n=4: `CosineScore("abcde", "abcdf", 4) = 0.49999999999999989` (rational limit 1/2 = 0.5 is 1 ULP higher; sqrt(2)*sqrt(2) = 2.0000000000000004)
  - **RV-C5** single-key intersection: `CosineScore("ab", "abcd", 2) = 0.5773502691896258` (matches the rational limit 1/sqrt(3) — sqrt(1)*sqrt(3) factorised form has no last-bit drift)
- **Property-test coverage:** 12 standard invariants (RangeBounds / Identity / Symmetric / NoNaN / NoInf / NoNegativeZero × byte+rune surfaces) plus the `_DeterministicAcrossRuns` gate (1000 sequential iterations on RV-C2 + RV-C3, math.Float64bits comparison). PLUS in-file `TestCosine_SortedKeyIteration` regression test (1000 iterations on RV-C2's 5-key intersection — the load-bearing input for sort.Strings discipline).
- **Fuzz harnesses:** Two fuzzers (byte + rune), each with 13-15 programmatic seeds covering RV-C1..RV-C5 + identity + both-empty + one-empty + orthogonal + invalid UTF-8 + multi-byte + Cyrillic + long-input + n=1 + n=8 boundary cases. On-disk seed files (`testdata/fuzz/FuzzCosineScore/seed-001` with RV-C1, `testdata/fuzz/FuzzCosineScoreRunes/seed-001` with RV-C3) in byte-stable `go test fuzz v1` literal format. 10-second smoke runs on each completed without panic / NaN / Inf / out-of-range — after the Rule 1 clamp deviation landed (see below).
- **Benchmark coverage:** 4 benches (ASCII Short / Medium / Long + Unicode Short). Allocation count within RESEARCH.md §4.1 budget (≤ 7 allocs/op on the realistic ceiling — Sørensen-Dice's 4-6 + 1 sorted-key slice).
- **BDD scenarios:** 8 scenarios (1 outline with 6 examples + 7 standalone) in `tests/bdd/features/cosine.feature` — covers RV-C1, RV-C2, RV-C4, RV-C5 reference vectors at 17-digit precision, identity, both-empty, one-empty (both directions), orthogonal, symmetry, and the rune-path RV-C3 café canary.
- **Per-algorithm staging golden:** `testdata/golden/_staging/cosine.json` with 9 alphabetically-sorted entries (canonical-marshalled via `assertGoldenStaging` → `CanonicalMarshalForTest`) spanning ASCII + Unicode at n ∈ {2, 3, 4} per CONTEXT.md §1 LOCKED. Plan 05-05 owns the merge into `algorithms.json`. **LOAD-BEARING per CONTEXT.md §1** — this is the cross-platform float-determinism gate that detects any drift on linux/amd64 vs linux/arm64 vs darwin/arm64 vs windows/amd64.

## Reference Vectors Verified

| ID | Inputs | n | Expected (IEEE-754 actual) | Rational Ideal | ULP Δ | Source |
|----|--------|---|----------------------------|----------------|-------|--------|
| RV-C1 | `"abc"` / `"abcd"` | 2 | 0.81649658092772592 | 2/sqrt(6) = 0.8164965809277261 | -1 | Salton & McGill 1983 §4.1 hand-derivation |
| RV-C2 | `"abcdefgh"` / `"abcdefgi"` | 3 | 0.83333333333333348 | 5/6 = 0.83333333333333337 | +1 | hand-derivation (5-key intersection — load-bearing for sort.Strings) |
| RV-C3 | `"café"` / `"cafe"` | 2 (RUNES) | 0.66666666666666674 | 2/3 = 0.66666666666666663 | +1 | hand-derivation (rune-path canary) |
| RV-C4 | `"abcde"` / `"abcdf"` | 4 | 0.49999999999999989 | 1/2 = 0.5 | -1 | hand-derivation (n=4 exercise; 1-ULP shortfall from sqrt(2)*sqrt(2) = 2.0000000000000004) |
| RV-C5 | `"ab"` / `"abcd"` | 2 | 0.5773502691896258 | 1/sqrt(3) = 0.5773502691896258 | 0 | optional irrational; no last-bit drift |
| Orthogonal | `"abc"` / `"xyz"` | 2 | 0.0 | 0.0 (exact) | 0 | empty intersection → dot=0 |
| Identity | any x with x | any n≥1 | 1.0 | 1.0 (exact) | 0 | a == b short-circuit |
| Both-empty | `""` / `""` | any n≥1 | 1.0 | 1.0 (exact) | 0 | a == b short-circuit covers both-empty |

The IEEE-754 / rational ideal divergence on RV-C1, RV-C2, RV-C3, RV-C4 is intrinsic to the Salton & McGill 1983 §4.1 factorised form `dot / (sqrt(normASq) * sqrt(normBSq))` — `sqrt(x) * sqrt(y)` is not exactly correctly rounded (only `sqrt` itself is, per IEEE-754 §5.4.1). RV-C5 happens to be exact because `sqrt(1) = 1.0` exactly and the factorised form reduces to a single `sqrt(3)` precision exercise. Per RESEARCH.md "Pitfall 2", the test pins the IEEE-754 actual; the rational ideal is documented for reviewer cross-check but is not the test expectation.

## Test Results

| Test category | Count | Status |
|---------------|-------|--------|
| Unit tests (TestCosine* — 12 functions, including 5 RV-CN derivation blocks) | 12 functions / many sub-tests | All PASS |
| Property tests (TestProp_Cosine*) | 13 functions × 100 quick.Check iterations + 1000-iter determinism gate | All PASS |
| Benchmarks | 4 (ASCII Short/Medium/Long + Unicode Short) | All within RESEARCH.md §4.1 alloc budget (≤ 7 allocs/op realistic ceiling) |
| Fuzz (10s smoke per harness) | 2 harnesses (FuzzCosineScore: 154k+ execs; FuzzCosineScoreRunes: 276k+ execs) | No panics, no NaN/Inf, all in [0, 1] (after Rule 1 clamp landed) |
| BDD scenarios | 8 in cosine.feature | All PASS via `make test-bdd` |
| Dispatch tests (TestDispatch_CosineRegistered + TestDispatch_UnregisteredSlotsAreNil) | 2 | Both PASS; slot 11 now registered |
| Golden file (TestGolden_Cosine_Staging) | 1 (9 entries) | PASS — file canonical-marshalled, byte-stable |
| Examples (ExampleCosineScore, ExampleCosineScoreRunes) | 2 | Both PASS — Output blocks match byte-for-byte at 16-digit precision |
| TestAIFriendly_LLMSTxtReferencesEveryExportedSymbol | 1 | PASS — llms.txt sync verified |

## Determinism Gates

- **No `init()` functions** in `cosine.go` or `dispatch_cosine.go` — verified via grep in the per-task verification.
- **No transcendental floats** in `cosine.go` except `math.Sqrt` — only `+`, `-`, `*`, `/`, comparisons, `float64()` casts, and `math.Sqrt`. Verified via `grep -E "math\\.(Pow|Log|Exp|FMA)"` returning empty.
- **Sorted-key iteration on the dot-product reduction** (CONTEXT.md §3 LOCKED): `sort.Strings(intersectionKeys)` is called BEFORE the dot-product loop iterates the slice. This makes the float reduction order canonical regardless of map-iteration randomisation. `TestCosine_SortedKeyIteration` (1000 iterations, math.Float64bits comparison) is the regression detector.
- **Norm computations iterate map values in any order** — integer addition is exactly associative; the OUTPUT of the norm loops is a scalar int (then a single `math.Sqrt` per side). DET-03 satisfied.
- **Cross-platform float determinism**: `math.Sqrt` is IEEE-754 correctly rounded on all four CI platforms (RESEARCH.md §3.5). Same int input → byte-identical float64 output across linux/amd64, linux/arm64, darwin/arm64, windows/amd64. The cross-platform CI matrix gate (`testdata/golden/algorithms.json` after plan 05-05 merge) is the load-bearing detector for any drift.
- **FMA risk surface documented inline** (RESEARCH.md §3 OQ-1): the cosine.go cosineFromQGramMaps godoc and the dot-product loop comment both document the FMA risk and the remediation pattern (`float64(x*y) + z` cast). Per RESEARCH.md §3.3, the empirical observation is that the (integer-derived) values of qa[k] and qb[k] are small enough that any FMA-vs-non-FMA divergence falls below the byte-diff threshold of the algorithms.json gate. The Phase 2/3/4 codebase has been using the same `(x*y) + z` recipe on accumulator paths and passing the cross-platform gate.
- **Per-process determinism**: `TestProp_CosineScore_DeterministicAcrossRuns` runs 1000 sequential iterations on the RV-C2 5-key-intersection byte pair AND the RV-C3 café rune pair, comparing via `math.Float64bits` for bit-level equality.
- **IEEE-754 [0.0, 1.0] clamp** at the final cosine result preserves the public-API range contract against IEEE-754 rounding overshoots in degenerate cases (see Deviations).

## Deviations from Plan

### Auto-fixed Issues

**1. [Rule 1 — Bug] IEEE-754 rounding overshoot above 1.0 on permutation-of-same-multiset inputs**

- **Found during:** Task 2 — FuzzCosineScore 10s smoke run.
- **Issue:** `CosineScore("cba", "abc", 1) = 1.0000000000000002`. The fuzzer surfaced a 1-ULP overshoot above 1.0 caused by `sqrt(3)*sqrt(3) = 2.9999999999999996` (1 ULP shortfall from 3.0); `3.0 / 2.9999999999999996 = 1.0000000000000002`. The mathematical limit is exactly 1.0 because QA == QB as multisets ({a:1, b:1, c:1} on both sides). This violates the public-API contract `cos(A,B) ∈ [0.0, 1.0]` declared in the godoc and asserted by `TestProp_CosineScore_RangeBounds`.
- **Fix:** Added a final clamp `if cos > 1.0 { return 1.0 }; if cos < 0.0 { return 0.0 }` at the end of `cosineFromQGramMaps`. The lower clamp is defensive (the dot product of non-negative integer-derived values is non-negative; `0.0 / positive = +0.0` in IEEE-754; `cos < 0.0` is theoretically unreachable). The upper clamp is the load-bearing fix for the IEEE-754 overshoot. The clamp is widely accepted in cosine implementations — the algebraic invariant `cos ∈ [0, 1]` for non-negative count vectors holds (Cauchy-Schwarz: `dot(A,B) ≤ ‖A‖·‖B‖`); IEEE-754 rounding is the source of drift, not the algorithm.
- **Files modified:** `cosine.go` (added 2 conditional returns at the end of `cosineFromQGramMaps` plus a multi-line comment documenting the fix and citing FuzzCosineScore as the surface).
- **Commit:** `1d47c60` (Task 2 commit; the fix landed in the same commit as the fuzz harness that surfaced it).

**2. [Rule 1 — Bug] Hand-derived rational ideals 1 ULP off from IEEE-754 actual on RV-C1, RV-C2, RV-C3, RV-C4**

- **Found during:** Task 1 — first run of `TestCosine_RV_C4_N4Exact` (and subsequently confirmed for RV-C1, RV-C2, RV-C3 via the staging golden file output).
- **Issue:** The plan / CONTEXT.md / RESEARCH.md §2.3 derived the rational ideals (e.g. RV-C1 = 2/sqrt(6) ≈ 0.8164965809277261, RV-C4 = 1/2 = 0.5 exactly). The implementation uses the Salton & McGill 1983 §4.1 factorised form `dot / (sqrt(normASq) * sqrt(normBSq))`, which is 1 ULP below the rational limit on RV-C1 (0.81649658092772592), 1 ULP above on RV-C2 (0.83333333333333348) and RV-C3 (0.66666666666666674), and 1 ULP below on RV-C4 (0.49999999999999989). The discrepancy is intrinsic to IEEE-754 — `sqrt(x) * sqrt(y)` is not exactly correctly rounded (only `sqrt` itself is).
- **Fix:** Pinned all four affected RV-CN reference vectors to the IEEE-754 actual factorised-form output. Each test comment now documents both the rational ideal and the IEEE-754 actual with the exact arithmetic chain that produces the 1-ULP drift (per RESEARCH.md "Pitfall 2": the implementation is correct; the hand-derivation is the rational limit, which is the textbook target but not the float64 reality). The `cosineEpsilon = 1e-15` tolerance was already generous enough to absorb the drift, but pinning the actual values produces a cleaner invariant and a stronger regression test. The staging golden file (`testdata/golden/_staging/cosine.json`) similarly pins the IEEE-754 actual via the live `fuzzymatch.CosineScore` call.
- **Files modified:** `cosine_test.go` (RV-C1, RV-C2, RV-C3, RV-C4 derivation blocks updated with IEEE-754-actual `want` constants and expanded comment derivations). RV-C5 was unaffected — its `want` constant (`0.5773502691896258`) coincides with the rational limit because `sqrt(1) = 1.0` exactly and the factorised form reduces to a single `sqrt(3)` precision exercise.
- **Commit:** `29dcd0e` (Task 1 commit; the fix landed before commit creation).

No other deviations. The plan was executed exactly as written; the 9-entry staging slate, the dispatch wiring, the BDD scenarios, the property tests, and the FMA-risk footnote all landed per the plan specification.

## Auth Gates

None. The plan executed end-to-end without any external authentication.

## OQ-1 Confirmed In-Plan 2026-05-15

> **OQ-1 (FMA on arm64) resolution per RESEARCH.md §3 / §6** is documented inline in `cosine.go` `cosineFromQGramMaps` godoc and dot-product loop comment. The empirical observation from Phase 2/3/4 (RESEARCH.md §3.3 — `lcsstr.go` line 215 et al. use the same `(x*y) + z` recipe on accumulator paths and pass the cross-platform CI matrix) is the basis for shipping the LOCKED CONTEXT.md §3 recipe without a defensive `float64(x*y) + z` cast. The remediation path is documented inline as a code comment near the dot-product loop; if a future CI matrix divergence ever surfaces on Cosine entries, the one-line cast is the fix and no other code change is required. The cross-platform CI matrix gate (testdata/golden/algorithms.json after plan 05-05 merge) is the load-bearing detector. The 9-entry Cosine staging slate (5 ASCII at n ∈ {2,3,4} + 4 Unicode/edge cases) gives multiple intersection sizes and rune patterns — strong coverage for surfacing any FMA-induced divergence.

## Forward References

- **Plan 05-04 (Tversky)** consumes the same `extractQGrams` / `extractQGramsRunes` helpers from plan 05-01 and the `ErrInvalidTverskyParam` sentinel declared in plan 05-01. Tversky's `α + β` denominator + division has its own IEEE-754 precision surface — plan 05-04 may need a similar [0.0, 1.0] clamp following the Cosine pattern (Rule 1 deviation 1 above). The `TestCosine_SortedKeyIteration` pattern is reusable for any future Tversky variant that iterates an intersection slice (Tversky's straight cardinality-based form does not require sort discipline because the OUTPUT is integer-derived; only Cosine needs the sort).
- **Plan 05-05 (Finalisation)** merges this plan's `testdata/golden/_staging/cosine.json` into the canonical `testdata/golden/algorithms.json` via `TestGolden_Algorithms_Merge`'s stagingFiles slice. The merged file then runs through `make verify-determinism` on the cross-platform CI matrix to validate byte-identical output across linux/amd64, linux/arm64, darwin/arm64, windows/amd64. **This is the ultimate gate for the LOAD-BEARING Cosine determinism claim** (CONTEXT.md §1).
- **Phase 6 (Monge-Elkan / Token Ratios)** consumes `AlgoCosine` (now dispatched) as a permitted inner metric for `MongeElkanScore`. The default-n=3 dispatch wrapper is the entry point; the inner metric runs over token pairs.
- **Phase 8 (Scorer)** uses `ErrInvalidQGramSize` (declared in plan 05-01) for `WithCosineAlgorithm(weight, n)` parameter validation via `errors.Is`. The Scorer option layer is where consumers configure non-default `n` values; the dispatch wrapper's hard-coded `n = 3` is acceptable per CONTEXT.md "Claude's Discretion" Deferred §4.

## Self-Check: PASSED

- `cosine.go` exists at `cosine.go` (FOUND).
- `dispatch_cosine.go` exists at `dispatch_cosine.go` (FOUND).
- `cosine_test.go` exists at `cosine_test.go` (FOUND, RV-C1..RV-C5 derivation blocks present — 16 occurrences of "RV-C" in file).
- `cosine_bench_test.go` exists at `cosine_bench_test.go` (FOUND).
- `cosine_fuzz_test.go` exists at `cosine_fuzz_test.go` (FOUND).
- `testdata/golden/_staging/cosine.json` exists (FOUND, 9 entries, alphabetically sorted, canonical-marshalled).
- `testdata/fuzz/FuzzCosineScore/seed-001` exists (FOUND, byte-stable `go test fuzz v1` format with RV-C1 canonical seed).
- `testdata/fuzz/FuzzCosineScoreRunes/seed-001` exists (FOUND, byte-stable format with RV-C3 café canonical seed).
- `tests/bdd/features/cosine.feature` exists (FOUND, 8 scenarios).
- `algorithms_golden_test.go` updated with `buildCosineStagingEntries` + `TestGolden_Cosine_Staging` (FOUND).
- `algoid_test.go` updated with `TestDispatch_CosineRegistered` and slot 11 in `TestDispatch_UnregisteredSlotsAreNil` (FOUND).
- `props_test.go` updated with the 13 Cosine property tests (FOUND).
- `example_test.go` updated with `ExampleCosineScore` + `ExampleCosineScoreRunes` (FOUND).
- `tests/bdd/steps/algorithms_steps.go` updated with 4 step methods + 4 ctx.Step registrations (FOUND).
- `llms.txt` updated with Cosine section (FOUND).
- `llms-full.txt` updated with parallel Phase 5 algorithm-surface block (FOUND).
- Commit `29dcd0e` (Task 1 — feat: cosine.go + dispatch + tests + staging) exists (FOUND in git log).
- Commit `1d47c60` (Task 2 — test: property + bench + fuzz; Rule 1 clamp deviation) exists (FOUND in git log).
- Commit `f215d2d` (Task 3 — test: BDD feature + steps) exists (FOUND in git log).
- `make test-bdd` exits 0 (VERIFIED — Cosine scenarios green).
- `bash scripts/verify-license-headers.sh` exits 0 (VERIFIED — 97 .go files carry the Apache-2.0 header).
- `bash scripts/verify-no-runtime-deps.sh` exits 0 (VERIFIED — root go.mod allowlist clean).
- `! grep -q "^func init" cosine.go` (VERIFIED — no init()).
- `grep -q "Source: Salton, G." cosine.go` (VERIFIED — primary source citation present).
- `grep -q "sort.Strings" cosine.go` (VERIFIED — CONTEXT.md §3 LOCKED).
- `grep -q "math.Sqrt" cosine.go` (VERIFIED — only math.* call).
- `grep -v '^//' cosine.go | grep -E "math\.(Pow|Log|Exp|FMA)"` returns empty (VERIFIED — DET-06 gate).
- `grep -q "CosineScore" llms.txt && grep -q "CosineScoreRunes" llms.txt` (VERIFIED — both new functions in llms.txt).
- `grep -c "RV-C" cosine_test.go` returns 16 (RV-C1, RV-C2, RV-C3, RV-C4, RV-C5 derivation blocks present — well above the ≥ 4 plan threshold).
