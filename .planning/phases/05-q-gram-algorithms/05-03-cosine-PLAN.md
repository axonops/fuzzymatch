---
phase: 05-q-gram-algorithms
plan: 03
type: execute
wave: 2
depends_on:
  - 05-01-qgram-foundation-jaccard
files_modified:
  - cosine.go
  - dispatch_cosine.go
  - cosine_test.go
  - cosine_bench_test.go
  - cosine_fuzz_test.go
  - props_test.go
  - example_test.go
  - algoid_test.go
  - algorithms_golden_test.go
  - testdata/golden/_staging/cosine.json
  - testdata/fuzz/FuzzCosineScore/seed-001
  - testdata/fuzz/FuzzCosineScoreRunes/seed-001
  - tests/bdd/features/cosine.feature
  - tests/bdd/steps/algorithms_steps.go
  - llms.txt
  - llms-full.txt
autonomous: true
requirements:
  - QGRAM-04
tags: [cosine, salton-mcgill-1983, vector-space, float-determinism, sorted-keys, math-sqrt, fma-risk, load-bearing-determinism, byte-and-rune-paths, dispatch-registration, property-tests, fuzz, benchmark, bdd, staging-golden, llms-sync]

must_haves:
  truths:
    # Cosine algorithm (QGRAM-04 — LOAD-BEARING cross-platform float-determinism)
    - "A caller can `import fuzzymatch` and call CosineScore(\"abc\", \"abcd\", 2) and receive `0.8164965809277261` (RV-C1 — 17-significant-digit float64; sqrt(6) precision exercise)"
    - "CosineScore(\"abcdefgh\", \"abcdefgi\", 3) returns `0.8333333333333334` (RV-C2 — 5 intersection keys, exercises sorted-key accumulation order from CONTEXT.md §3)"
    - "CosineScore(\"abcde\", \"abcdf\", 4) returns `0.5` (RV-C4 — n=4 exercise; exactly representable rational)"
    - "CosineScoreRunes(\"café\", \"cafe\", 2) returns `0.6666666666666666` (RV-C3 — Unicode rune path)"
    - "CosineScore(\"hello\", \"hello\", 2) == 1.0 (identity)"
    - "CosineScore(\"\", \"\", 2) == 1.0 (both-empty)"
    - "CosineScore(\"\", \"abc\", 2) == 0.0 (one-empty)"
    - "CosineScore(\"abc\", \"xyz\", 2) == 0.0 (orthogonal — no shared q-grams)"
    - "CosineScore(x, x, n) == 1.0 for every non-empty x and every n ≥ 1"
    - "CosineScore(a, b, n) == CosineScore(b, a, n) for every (a, b, n) — Cosine is symmetric"
    - "CosineScore(\"hello\", \"hello\", 0) panics with message containing `invalid q-gram size`; same for n=-1"
    - "dispatch[AlgoCosine] is non-nil after package load and dispatches to a wrapper that calls CosineScore(a, b, 3) (default n=3 per CONTEXT.md Deferred §4)"
    # LOAD-BEARING DETERMINISM (CONTEXT.md §1 + §3)
    - "cosine.go iterates intersection keys in SORTED ORDER: build the intersection key set, copy to `[]string`, `sort.Strings(keys)`, then `for _, k := range keys { dot += float64(qa[k]) * float64(qb[k]) }` — CONTEXT.md §3 LOCKED"
    - "The dot-product reduction uses explicit `(x*y) + z` parenthesisation: `dot = (float64(qa[k]) * float64(qb[k])) + dot` — per CONTEXT.md §3 and DET-06"
    - "Both norms use math.Sqrt only: `normA := math.Sqrt(float64(sumSquaresA))`, `normB := math.Sqrt(float64(sumSquaresB))` — NO math.Pow, NO `**0.5`, NO custom Newton iteration (math.Sqrt is IEEE-754 correctly rounded on all four CI platforms per RESEARCH.md §3.5)"
    - "Sum-of-squares reductions iterate map values in any order (commutative integer addition; no float reduction here) — `for _, v := range qa { sumSquaresA += v * v }` is acceptable because integer addition is exactly associative; the float-determinism risk is on the dot-product reduction only"
    - "Final score computation: `cos := dot / (normA * normB)` with explicit parenthesisation `(normA * normB)` — single division on IEEE-754 correctly rounded floats"
    - "FMA risk footnote present in cosine.go: a code comment near the dot-product loop documents the OQ-1 finding (FMA fusion on arm64 does NOT prevent the recipe from passing the CI matrix; remediation = `float64(x*y) + z` cast if cross-platform divergence ever appears — see RESEARCH.md §3.4)"
    # Algorithmic correctness gates
    - "cosine.go's file-level godoc cites Salton, G., McGill, M. J. (1983). \"Introduction to Modern Information Retrieval\" McGraw-Hill — specifically §4.1 equation 4.4 (Cosine measure) on p.121 — as the PRIMARY source"
    - "The source-origin statement block is present verbatim (Primary / Cross-validation: hand-derived RV-C1..RV-C5 in cosine_test.go with 17-digit float64 derivations in test comments per CONTEXT.md §4 / Tie-break: none / GPL-LGPL: none / Code copied: none)"
    - "cosine.go contains NO init() function — determinism-reviewer flags any init() as BLOCKING"
    - "CosineScore never returns NaN, +Inf, -Inf, or -0 for any input — verified by TestProp_CosineScore_NoNaN / _NoInf / _NoNegativeZero (byte + rune)"
    - "CosineScore is deterministic: 1000 sequential calls on the same input produce byte-identical output (TestProp_CosineScore_DeterministicAcrossRuns)"
    - "No `math.Pow`, `math.Log`, `math.Exp`, `math.FMA` anywhere in cosine.go (only `+`, `-`, `*`, `/`, comparisons, `float64()` casts, and `math.Sqrt`) — DET-06 gate"
    # Public-surface + meta-test discipline
    - "Public surface added by this plan: exactly two new exported symbols (CosineScore, CosineScoreRunes) — pre-existing AlgoCosine constant already in algoid.go slot"
    - "FuzzCosineScore + FuzzCosineScoreRunes panic-free, score-in-[0,1], NaN/Inf-free for any (a, b) including invalid UTF-8; n constrained to [1, 8] via the fuzz body"
    - "testdata/fuzz/FuzzCosineScore/seed-001 and testdata/fuzz/FuzzCosineScoreRunes/seed-001 exist in byte-stable `go test fuzz v1` literal format"
    - "tests/bdd/features/cosine.feature exists with: hand-derived reference-vector Scenario Outline (RV-C1, RV-C2, RV-C3 via Runes scenario, RV-C4), identity, both-empty, one-empty, orthogonal, symmetry scenarios"
    - "tests/bdd/steps/algorithms_steps.go appends Cosine step bindings and their ctx.Step regex registrations"
    - "testdata/golden/_staging/cosine.json exists, produced by TestGolden_Cosine_Staging via assertGoldenStaging; contains 9 alphabetically-sorted entries per RESEARCH.md §2.3 Cosine slate (Cosine_both_empty, Cosine_one_empty, Cosine_identical, Cosine_orthogonal, Cosine_ascii_n2_irrational [RV-C1], Cosine_ascii_n3_large_intersection [RV-C2], Cosine_ascii_n4_exact [RV-C4], Cosine_unicode_n2_runes [RV-C3], Cosine_unicode_n3_runes); spans ASCII + Unicode at n ∈ {2, 3, 4} per CONTEXT.md §1 LOCKED"
    - "algoid_test.go contains a new TestDispatch_CosineRegistered; registered map updated to flip AlgoCosine slot to true"
    - "ExampleCosineScore and ExampleCosineScoreRunes appended to example_test.go; `// Output:` blocks match byte-for-byte"
    - "llms.txt lists `CosineScore` and `CosineScoreRunes` (the two new exported symbols). The AlgoID constant AlgoCosine is ALREADY listed (declared in Phase 1)"
    - "llms-full.txt has parallel entries with one-line rationales"
    - "Coverage on cosine.go ≥ 90%; 100% on the public CosineScore + CosineScoreRunes surface"
    - "Apache-2.0 header present on every new .go file"
  artifacts:
    - path: "cosine.go"
      provides: "CosineScore + CosineScoreRunes (two new public functions); cites Salton & McGill 1983 §4.1 eq. 4.4 p.121 as PRIMARY"
      min_lines: 130
      contains: "Source: Salton, G."
    - path: "dispatch_cosine.go"
      provides: "Package-load-time registration of a default-n=3 CosineScore wrapper into dispatch[AlgoCosine]"
      contains: "dispatch[AlgoCosine]"
    - path: "cosine_test.go"
      provides: "Unit tests with 4-5 hand-derivation comment blocks (RV-C1..RV-C4 plus optional RV-C5) reproducing the 17-significant-digit float64 calculations from Salton & McGill 1983 §4.1; reviewer-verifiable in <30s per derivation"
      min_lines: 200
      contains: "RV-C1"
    - path: "cosine_bench_test.go"
      provides: "Benchmarks: BenchmarkCosineScore_{ASCII_Short, ASCII_Medium, ASCII_Long} + BenchmarkCosineScoreRunes_Unicode_Short — alloc-asserted; reserves one []string alloc per call for the sorted-key slice (CONTEXT.md §3 implication)"
    - path: "cosine_fuzz_test.go"
      provides: "FuzzCosineScore and FuzzCosineScoreRunes — panic-free, NaN/Inf-free, score-in-[0,1]"
    - path: "props_test.go"
      provides: "Appended Cosine property-test block: TestProp_CosineScore_{RangeBounds, Identity, Symmetric, NoNaN, NoInf, NoNegativeZero} for BOTH byte and rune surfaces (12 property tests) plus TestProp_CosineScore_DeterministicAcrossRuns"
    - path: "example_test.go"
      provides: "Appended ExampleCosineScore + ExampleCosineScoreRunes"
    - path: "algoid_test.go"
      provides: "Appended TestDispatch_CosineRegistered"
    - path: "algorithms_golden_test.go"
      provides: "Appended buildCosineStagingEntries + TestGolden_Cosine_Staging — produces _staging/cosine.json with 9 entries (CONTEXT.md §1 LOCKED ASCII + Unicode at n ∈ {2,3,4})"
    - path: "testdata/golden/_staging/cosine.json"
      provides: "Per-algorithm staging file; sorted by Name; 9 entries spanning ASCII + Unicode at n ∈ {2, 3, 4}; merged into algorithms.json by plan 05-05; cross-platform byte-identical via CI matrix"
      contains: "Cosine_ascii_n2_irrational"
    - path: "testdata/fuzz/FuzzCosineScore/seed-001"
      provides: "Fuzz seed corpus in `go test fuzz v1` literal format"
    - path: "testdata/fuzz/FuzzCosineScoreRunes/seed-001"
      provides: "Fuzz seed corpus for the rune-path harness"
    - path: "tests/bdd/features/cosine.feature"
      provides: "Gherkin feature with hand-derived reference vectors at 17-digit precision, identity, both-empty, one-empty, orthogonal, symmetry, rune-path Unicode"
    - path: "tests/bdd/steps/algorithms_steps.go"
      provides: "Appended Cosine step methods + ctx.Step registrations"
    - path: "llms.txt"
      provides: "Appended 2 function entries (CosineScore, CosineScoreRunes)"
    - path: "llms-full.txt"
      provides: "Parallel entries with one-line rationales"
  key_links:
    - from: "cosine.go (CosineScore + CosineScoreRunes)"
      to: "q_gram.go (extractQGrams + extractQGramsRunes — created in plan 05-01)"
      via: "Direct call into the unexported extractor; CosineScore builds intersection key slice, sort.Strings(keys), iterates SORTED keys for dot-product reduction; norms computed via math.Sqrt on integer-derived sum-of-squares"
      pattern: "extractQGrams(Runes)?\\(|sort\\.Strings"
    - from: "cosine.go (dot-product loop)"
      to: "CONTEXT.md §3 LOCKED iteration order"
      via: "Sorted-key iteration with explicit `(x*y) + z` parenthesisation; FMA-risk footnote comment near the loop documents the OQ-1 finding"
      pattern: "sort\\.Strings|\\(.*\\*.*\\)\\s*\\+"
    - from: "testdata/golden/_staging/cosine.json"
      to: "Cross-platform CI matrix (linux/amd64, linux/arm64, darwin/arm64, windows/amd64)"
      via: "9 entries spanning ASCII + Unicode at n ∈ {2, 3, 4}; merged into algorithms.json by plan 05-05; `make verify-determinism` asserts byte-identical output on every platform; ANY single-byte drift fails the gate hard"
      pattern: "Cosine_(ascii|unicode|both|one|identical|orthogonal)"
    - from: "dispatch_cosine.go"
      to: "algoid.go (AlgoCosine declared at line 122)"
      via: "package-level closure wrapper `dispatch[AlgoCosine] = func(a, b string) float64 { return CosineScore(a, b, 3) }`"
      pattern: "dispatch\\[AlgoCosine\\]"
---

<objective>
Implement Cosine n-gram similarity (QGRAM-04) — `cos(A, B) = (A · B) / (‖A‖ × ‖B‖)` over q-gram frequency vectors (Salton & McGill 1983 §4.1 equation 4.4 p.121) — atop the shared q-gram infrastructure from plan 05-01. This is the **load-bearing cross-platform float-determinism algorithm** for the phase: the algorithms.json golden file Cosine entries are the gate that detects any drift in float reduction order across the four CI platforms (linux/amd64, linux/arm64, darwin/arm64, windows/amd64). The hand-derived reference vectors RV-C1..RV-C5 with 17-significant-digit float64 derivations in test comments are the load-bearing cross-correctness proof in lieu of an external library reference (CONTEXT.md §4 LOCKED). CONTEXT.md §3 LOCKS sorted-key iteration for the dot-product loop; CONTEXT.md §4 LOCKS hand-derivation density on Cosine; CONTEXT.md §1 LOCKS the algorithms.json gate.

Purpose: ship the textbook vector-space cosine measure that consumers expect for IR / NLP-flavoured workloads, with reviewer-verifiable correctness (each RV-CN derivation is verifiable in <30 seconds against Salton & McGill 1983), bit-stable cross-platform output (CI matrix verifies byte-identical golden), and explicit documentation of the FMA-risk surface (RESEARCH.md §3 OQ-1 finding) without changing the LOCKED recipe.

Output: 16 new/modified files (4 new source/test files, plus extensions to 7 existing append-only files, plus 4 new test/fixture files in testdata + tests/bdd). Single new dispatch slot wired (Wave 2 — depends on plan 05-01 ONLY; parallelisable with plans 05-02 and 05-04).
</objective>

<execution_context>
@$HOME/.claude/get-shit-done/workflows/execute-plan.md
@$HOME/.claude/get-shit-done/templates/summary.md
</execution_context>

<context>
@.planning/PROJECT.md
@.planning/ROADMAP.md
@.planning/STATE.md
@.planning/REQUIREMENTS.md
@.planning/phases/05-q-gram-algorithms/05-CONTEXT.md
@.planning/phases/05-q-gram-algorithms/05-RESEARCH.md
@.planning/phases/05-q-gram-algorithms/05-01-qgram-foundation-jaccard-PLAN.md
@.planning/phases/04-remaining-character-gestalt/04-PATTERNS.md
@.claude/skills/algorithm-correctness-standards/SKILL.md
@.claude/skills/algorithm-licensing-standards/SKILL.md
@.claude/skills/determinism-standards/SKILL.md
@.claude/skills/performance-standards/SKILL.md
@.claude/skills/go-testing-standards/SKILL.md
@algoid.go
@q_gram.go
@qgram_jaccard.go
@dispatch_qgram_jaccard.go
</context>

<interfaces>
<!-- Key types/functions executor MUST use without rediscovering. -->

From q_gram.go (created in plan 05-01 — read-only here):
```go
func extractQGrams(s string, n int) map[string]int
func extractQGramsRunes(s string, n int) map[string]int
```

From algoid.go (slot already declared at line 122; do NOT modify):
```go
const AlgoCosine AlgoID = ... // line 122; String() case at line 237-238
```

Public surface to be created by this plan:
```go
// CosineScore returns the Cosine similarity of the q-gram frequency
// vectors of a and b (Salton & McGill 1983 §4.1 eq. 4.4 p.121).
// cos(A, B) = (A · B) / (‖A‖ × ‖B‖), in [0.0, 1.0].
//
// Iteration order over the intersection keys is SORTED (sort.Strings)
// per CONTEXT.md §3 LOCKED; the dot-product reduction uses explicit
// (x*y) + z parenthesisation. See cosine.go inline footnote for the
// FMA-risk surface (RESEARCH.md §3 OQ-1).
//
// Panics on n < 1 with a message containing "invalid q-gram size".
func CosineScore(a, b string, n int) float64

func CosineScoreRunes(a, b string, n int) float64
```

Dispatch wiring (matches dispatch_qgram_jaccard.go — default n=3):
```go
var _ = func() bool {
    dispatch[AlgoCosine] = func(a, b string) float64 {
        return CosineScore(a, b, 3)
    }
    return true
}()
```

Sorted-intersection dot-product loop SHAPE (from RESEARCH.md "Code Examples → Sorted-intersection dot-product loop"):
```go
// Build the intersection key set
intersectionKeys := make([]string, 0, len(qa))
for k := range qa {
    if _, ok := qb[k]; ok {
        intersectionKeys = append(intersectionKeys, k)
    }
}
sort.Strings(intersectionKeys)

// Dot product: iterate SORTED keys; explicit (x*y)+z parenthesisation
var dot float64
for _, k := range intersectionKeys {
    dot = (float64(qa[k]) * float64(qb[k])) + dot
}

// Norms: math.Sqrt only; sum-of-squares accumulation can iterate in any order (integer addition is exactly associative)
var sumSquaresA, sumSquaresB int
for _, v := range qa { sumSquaresA += v * v }
for _, v := range qb { sumSquaresB += v * v }
normA := math.Sqrt(float64(sumSquaresA))
normB := math.Sqrt(float64(sumSquaresB))

// Final cosine: single division on IEEE-754 correctly rounded floats
return dot / (normA * normB)
```
</interfaces>

<tasks>

<task type="auto" tdd="true">
  <name>Task 1: Implement cosine.go + dispatch + unit tests with 4-5 hand-derivation blocks + staging golden</name>
  <files>cosine.go, dispatch_cosine.go, cosine_test.go, testdata/golden/_staging/cosine.json, algorithms_golden_test.go, algoid_test.go, example_test.go, llms.txt, llms-full.txt</files>
  <read_first>
    - cosine.go (current state — confirm it does NOT exist; creating new)
    - q_gram.go (plan 05-01 output — read to understand extractor signatures)
    - qgram_jaccard.go (plan 05-01 — closest structural analog, but the Cosine reduction path differs significantly)
    - dispatch_qgram_jaccard.go (plan 05-01 — exact template for dispatch_cosine.go)
    - .planning/phases/05-q-gram-algorithms/05-CONTEXT.md §1 (algorithms.json determinism gate), §3 (sorted-keys iteration LOCKED), §4 (hand-derived RV density LOCKED), §5 (general patterns)
    - .planning/phases/05-q-gram-algorithms/05-RESEARCH.md §1.3 (Cosine formula), §2.3 (RV-C1..RV-C5 full 17-digit derivations + the 9-entry algorithms.json slate), §3 (Go 1.26 FMA / float-determinism specifics), §4.4 (sort.Strings cost), §6 OQ-1 (FMA footnote requirement), "Code Examples → Sorted-intersection dot-product loop"
    - algoid.go line 122 (AlgoCosine slot)
    - algorithms_golden_test.go (find buildQGramJaccardStagingEntries from plan 05-01 — exact template; the Cosine slate is LARGER — 9 entries per RESEARCH.md §2.3)
    - algoid_test.go (last 50 lines — TestDispatch_*Registered template)
    - example_test.go (last 30 lines)
    - llms.txt (find the QGramJaccard entries from plan 05-01)
  </read_first>
  <behavior>
    - CosineScore("abc", "abcd", 2) returns 0.8164965809277261 (RV-C1) within 1e-15 — `2/sqrt(6)`
    - CosineScore("abcdefgh", "abcdefgi", 3) returns 0.8333333333333334 (RV-C2) within 1e-15 — `5/6`
    - CosineScore("abcde", "abcdf", 4) returns 0.5 exactly (RV-C4) — `1/2` (4-grams: {abcd, bcde} vs {abcd, bcdf}; |∩|=1, ‖A‖²=2, ‖B‖²=2; cos = 1/(sqrt(2)·sqrt(2)) = 1/2 exactly)
    - CosineScoreRunes("café", "cafe", 2) returns 0.6666666666666666 (RV-C3) within 1e-15 — `2/3`
    - CosineScore("hello", "hello", 2) == 1.0 (identity)
    - CosineScore("", "", 2) == 1.0 (both-empty)
    - CosineScore("", "abc", 2) == 0.0 (one-empty)
    - CosineScore("abc", "xyz", 2) == 0.0 (orthogonal — empty intersection → dot=0 → cos=0)
    - CosineScore(a, b, n) == CosineScore(b, a, n) for arbitrary input (exact equality — sorted-key iteration is canonical regardless of input order)
    - CosineScore("hello", "hello", 0) panics with "invalid q-gram size"
    - dispatch[AlgoCosine]("hello", "hello") returns 1.0 (default n=3; identity short-circuit fires)
    - TestGolden_Cosine_Staging produces testdata/golden/_staging/cosine.json with 9 alphabetically-sorted entries (CONTEXT.md §1 LOCKED slate)
  </behavior>
  <action>
    Step A — Create cosine.go. File order:
    (a) Apache-2.0 header.
    (b) File-level doc: cite Salton, G., McGill, M. J. (1983). "Introduction to Modern Information Retrieval" McGraw-Hill — specifically §4.1 equation 4.4 (Cosine measure) on p.121 — as PRIMARY. Include the explicit formula in the godoc block (`cos(A, B) = (A · B) / (‖A‖ × ‖B‖)` with the q-gram frequency-vector interpretation). Source-origin statement block (Primary / Cross-validation: hand-derived RV-C1..RV-C5 with 17-digit float64 derivations in cosine_test.go inline comments per CONTEXT.md §4 LOCKED / Tie-break: none / GPL-LGPL: none / Code copied: none).
    (c) `import ("math"; "sort")` — `math.Sqrt` only; `sort.Strings` for the intersection key slice.
    (d) `package fuzzymatch`.
    (e) `func CosineScore(a, b string, n int) float64` — godoc reproducing the formula, the CONTEXT.md §3 sorted-keys iteration order directive, the FMA-risk footnote pointer, the [0, 1] range, and the direct-call panic-on-n<1 contract. Body:
        - Identity short-circuit `if a == b { return 1.0 }`
        - One-empty short-circuit → 0.0
        - `if n < 1 { panic("fuzzymatch: invalid q-gram size") }`
        - `qa := extractQGrams(a, n); qb := extractQGrams(b, n)`
        - If `len(qa) == 0 && len(qb) == 0 { return 1.0 }` (post-extraction both-empty); if exactly one is empty after extraction → 0.0
        - Build the intersection key slice (iterate the SMALLER map to minimise scan; the slice content is identical regardless of which side is iterated): `intersectionKeys := make([]string, 0, len(smaller))` then loop appending keys present in the other map
        - `sort.Strings(intersectionKeys)` — CONTEXT.md §3 LOCKED
        - Dot-product reduction: `var dot float64; for _, k := range intersectionKeys { dot = (float64(qa[k]) * float64(qb[k])) + dot }` — explicit `(x*y) + z` parenthesisation per CONTEXT.md §3 + DET-06
        - Norms: `var sumSquaresA, sumSquaresB int; for _, v := range qa { sumSquaresA += v * v }; for _, v := range qb { sumSquaresB += v * v }` — integer addition, exactly associative regardless of map iteration order
        - `normA := math.Sqrt(float64(sumSquaresA)); normB := math.Sqrt(float64(sumSquaresB))` — math.Sqrt only (no math.Pow, no `**0.5`)
        - Return `dot / (normA * normB)` with explicit `(normA * normB)` parenthesisation
        - Add an inline footnote comment near the dot-product loop documenting the OQ-1 finding from RESEARCH.md §3.4: `// FMA risk surface (RESEARCH.md §3, OQ-1): Go 1.26 may emit FMA on arm64 for the (x*y)+z pattern; parentheses do NOT defeat FMA fusion. The cross-platform CI matrix gate (testdata/golden/algorithms.json) is the load-bearing detector. If matrix divergence ever appears on Cosine entries, remediate by inserting an explicit float64 cast: dot = float64(float64(qa[k]) * float64(qb[k])) + dot.`
    (f) `func CosineScoreRunes(a, b string, n int) float64` — analogous, calling extractQGramsRunes.
    (g) Use only `+`, `-`, `*`, `/`, comparisons, `float64()` casts, and `math.Sqrt`. NO math.Pow/Log/Exp/FMA.

    Step B — Create dispatch_cosine.go per dispatch_qgram_jaccard.go template. Closure binds n=3. NO init(). Apache-2.0 header.

    Step C — Create cosine_test.go. This file is LOAD-BEARING per CONTEXT.md §4. Required structure:
    - Apache-2.0 header.
    - Header comment block reproducing the Salton & McGill 1983 §4.1 eq. 4.4 (p.121) formula and pointing at RESEARCH.md §2.3 for the RV-C1..RV-C5 catalogue.
    - TestCosine_BothEmpty, TestCosine_OneEmpty, TestCosine_Identical, TestCosine_Orthogonal ("abc"/"xyz"/n=2 → 0.0)
    - TestCosine_RV_C1_AsciiN2Irrational — pinned to 0.8164965809277261 within 1e-15. **Test comment reproduces the full RESEARCH.md §2.3 RV-C1 derivation verbatim** (5+ lines: QA={ab:1,bc:1} ‖A‖²=2; QB={ab:1,bc:1,cd:1} ‖B‖²=3; intersection [ab,bc]; dot=2; cos=2/sqrt(6)=0.8164965809277261). The reviewer must be able to re-derive in <30s.
    - TestCosine_RV_C2_LargeIntersectionN3 — pinned to 0.8333333333333334 within 1e-15. Test comment reproduces RV-C2 derivation (6 trigrams each; 5 intersection keys; cos=5/6).
    - TestCosine_RV_C4_N4Exact — pinned to 0.5 exactly. Test comment reproduces RV-C4 derivation (4-grams; intersection={abcd}; ‖A‖²=‖B‖²=2; cos=1/2 exactly).
    - TestCosineRunes_RV_C3_UnicodeN2 — pinned to 0.6666666666666666 within 1e-15. Test comment reproduces RV-C3 derivation (rune-bigrams of "café"=[ca,af,fé]; rune-bigrams of "cafe"=[ca,af,fe]; intersection sorted=[af,ca]; dot=2; ‖A‖²=‖B‖²=3; cos=2/3).
    - TestCosine_RV_C5_OptionalIrrational (optional per RESEARCH.md §2.3 — single-key intersection where sqrt(3) is the irrational source). Include this as a fifth derivation block for extra reviewer density.
    - TestCosine_Symmetric — assert Score(a, b, n) == Score(b, a, n) on 3 hand-picked pairs (Cosine is exactly symmetric — equality)
    - TestCosine_PanicsOnInvalidN — table-driven defer-recover over n=0, n=-1, n=-100
    - TestCosine_AllocBound via testing.AllocsPerRun(100, ...) — document the alloc count per RESEARCH.md §4.1 (≤ 5 allocs/op — two maps from extractQGrams + one intersection-keys slice from sort.Strings cap-hint = ≤ 5)
    - TestCosine_SortedKeyIteration — INVARIANT REGRESSION TEST. Compute CosineScore on a pair where the map iteration order could meaningfully differ from sorted order (any pair with |intersection| ≥ 2 works); call CosineScore 1000 times; assert all 1000 results are byte-for-byte identical via `math.Float64bits` comparison. This catches a regression where someone removes the sort.Strings call and the dot-product reduction order becomes non-deterministic. Use RV-C2 ("abcdefgh"/"abcdefgi"/n=3) as the input (5-key intersection).
    - Stdlib testing only.

    Step D — Append buildCosineStagingEntries + TestGolden_Cosine_Staging to algorithms_golden_test.go. The 9-entry slate per CONTEXT.md §1 + RESEARCH.md §2.3 (sorted alphabetically):
      1. Cosine_ascii_n2_irrational ("abc"/"abcd"/n=2/byte/0.8164965809277261 = RV-C1)
      2. Cosine_ascii_n3_large_intersection ("abcdefgh"/"abcdefgi"/n=3/byte/0.8333333333333334 = RV-C2)
      3. Cosine_ascii_n4_exact ("abcde"/"abcdf"/n=4/byte/0.5 = RV-C4)
      4. Cosine_both_empty (""/""/n=2/byte/1.0)
      5. Cosine_identical ("hello"/"hello"/n=2/byte/1.0)
      6. Cosine_one_empty (""/"abc"/n=2/byte/0.0)
      7. Cosine_orthogonal ("abc"/"xyz"/n=2/byte/0.0)
      8. Cosine_unicode_n2_runes ("café"/"cafe"/n=2/rune/0.6666666666666666 = RV-C3)
      9. Cosine_unicode_n3_runes ("héllo"/"hello"/n=3/rune/0.3333333333333333 — derivation per RESEARCH.md §2.3 entry 9: rune-trigrams of "héllo" = [hél, éll, llo]; of "hello" = [hel, ell, llo]; intersection sorted = [llo]; dot=1; ‖A‖²=‖B‖²=3; cos=1/3=0.3333333333333333)
    Run `go test -run TestGolden_Cosine_Staging -update ./...` to materialise the file.

    Step E — Append TestDispatch_CosineRegistered to algoid_test.go; flip AlgoCosine slot in the registered map.

    Step F — Append ExampleCosineScore + ExampleCosineScoreRunes to example_test.go. Byte path uses RV-C1 ("abc"/"abcd"/n=2 → 0.8164965809277261 — print with `%.16f` to show the full precision); rune path uses RV-C3 ("café"/"cafe"/n=2 → 0.6666666666666666). Capture exact stdout and paste.

    Step G — Update llms.txt + llms-full.txt with the two new function entries.
  </action>
  <verify>
    <automated>cd /Users/johnny/Development/fuzzymatch && go build ./... && go test -run 'TestCosine|TestDispatch_CosineRegistered|TestDispatch_UnregisteredSlotsAreNil|TestGolden_Cosine_Staging|ExampleCosineScore' ./... && bash scripts/verify-license-headers.sh && ! grep -q "^func init" cosine.go && grep -q "Source: Salton, G." cosine.go && grep -q "sort.Strings" cosine.go && grep -q "math.Sqrt" cosine.go && grep -q "dispatch\[AlgoCosine\]" dispatch_cosine.go && grep -q "CosineScore" llms.txt && ! grep -v '^#' cosine.go | grep -E "math\.(Pow|Log|Exp|FMA)" && grep -q "RV-C1" cosine_test.go && grep -q "RV-C2" cosine_test.go && grep -q "RV-C4" cosine_test.go && grep -q "RV-C3" cosine_test.go</automated>
  </verify>
  <done>
    All Cosine* unit tests, TestDispatch_CosineRegistered, TestGolden_Cosine_Staging, and ExampleCosineScore pass. License headers green. NO init(). Source citation present. `sort.Strings` invoked in CosineScore on the intersection key slice. `math.Sqrt` is the only math.* call used. Test file contains 4-5 RV-CN derivation blocks with 17-digit float64 expected values and full inline derivations. 9-entry staging golden file alphabetically sorted. llms.txt + llms-full.txt updated.
  </done>
</task>

<task type="auto" tdd="true">
  <name>Task 2: Property tests + benchmarks + fuzz harnesses</name>
  <files>props_test.go, cosine_bench_test.go, cosine_fuzz_test.go, testdata/fuzz/FuzzCosineScore/seed-001, testdata/fuzz/FuzzCosineScoreRunes/seed-001</files>
  <read_first>
    - cosine.go (Task 1 output)
    - props_test.go (find the QGramJaccard property-test block from plan 05-01 — exact template for the Cosine append)
    - qgram_jaccard_bench_test.go (plan 05-01 — exact template, plus add one extra alloc for the sort.Strings slice)
    - qgram_jaccard_fuzz_test.go (plan 05-01 — exact template)
    - .planning/phases/05-q-gram-algorithms/05-RESEARCH.md §4.4 (sort.Strings cost on the intersection slice — adds one []string allocation per call)
  </read_first>
  <behavior>
    - 12 standard property tests (6 byte + 6 rune) + 1 determinism test
    - 4 benchmarks (3 byte + 1 rune) — alloc-asserted, expected ≤ 5 allocs/op
    - 2 fuzz harnesses: panic-free, NaN/Inf-free, score-in-[0,1]
  </behavior>
  <action>
    Step A — Extend props_test.go with the Cosine property-test block per the plan 05-01 / 05-02 template. 12 property tests + 1 determinism test. n coerced to [1, 5].

    Step B — Create cosine_bench_test.go. Three byte-path benches + one rune-path bench. Alloc expectation documented inline: ≤ 5 allocs/op (two map allocations from extractQGrams + one []string slice for the intersection keys + cap-hint backing arrays). The sort.Strings call itself does not allocate beyond the existing slice (in-place sort).

    Step C — Create cosine_fuzz_test.go. Two harnesses. Programmatic f.Add(...) seeds covering RV-C1..RV-C5, identity, both-empty, one-empty, orthogonal, invalid UTF-8, long input, n=2, n=8.

    Step D — Create testdata/fuzz/FuzzCosineScore/seed-001 and FuzzCosineScoreRunes/seed-001 in `go test fuzz v1` literal format using RV-C1 ("abc"/"abcd"/n=2) as the canonical seed for the byte harness and "café"/"cafe"/n=2 for the rune harness.
  </action>
  <verify>
    <automated>cd /Users/johnny/Development/fuzzymatch && go test -run 'TestProp_Cosine' ./... && go test -bench=BenchmarkCosine -benchmem -benchtime=1x ./... && go test -fuzz=FuzzCosineScore -fuzztime=10s ./... && go test -fuzz=FuzzCosineScoreRunes -fuzztime=10s ./... && head -1 testdata/fuzz/FuzzCosineScore/seed-001 | grep -q "^go test fuzz v1$"</automated>
  </verify>
  <done>
    All TestProp_Cosine* (byte + rune, 12+ property tests) + determinism test pass. Bench file produces 4 benches with alloc count within budget (≤ 5 allocs/op). Both fuzz harnesses pass a 10s smoke run. Byte-stable seed-001 files on disk.
  </done>
</task>

<task type="auto" tdd="true">
  <name>Task 3: BDD feature + steps</name>
  <files>tests/bdd/features/cosine.feature, tests/bdd/steps/algorithms_steps.go</files>
  <read_first>
    - tests/bdd/features/qgram_jaccard.feature (plan 05-01 — exact template)
    - tests/bdd/steps/algorithms_steps.go (find QGramJaccard step methods added in plan 05-01 — exact analog for the Cosine append; the Cosine BDD step grammar mirrors the Jaccard one — same n-parameterised shape)
    - .planning/phases/05-q-gram-algorithms/05-RESEARCH.md §5.3 (cosine.feature skeleton — note the 17-digit precision in the scenario outline)
  </read_first>
  <behavior>
    - godog runs and passes all Cosine scenarios
    - At least RV-C1 / RV-C2 / RV-C4 covered in the Examples table at 17-digit precision; RV-C3 in the rune-path scenario
    - Existing approximately-step regex `(\d+\.?\d*)` accepts the 17-digit form (verified per Phase 4 IN-03 closure)
  </behavior>
  <action>
    Create tests/bdd/features/cosine.feature per RESEARCH.md §5.3 skeleton. Header comment: `# Primary source: Salton, G., McGill, M. J. (1983). "Introduction to Modern Information Retrieval" McGraw-Hill, §4.1 eq. 4.4, p.121.` Scenarios:
    - `Feature: Cosine n-gram similarity`
    - `Scenario Outline: Hand-derived reference vectors with float64 precision` covering RV-C1 (abc/abcd/n=2/0.8164965809277261), RV-C2 (abcdefgh/abcdefgi/n=3/0.8333333333333334), RV-C4 (abcde/abcdf/n=4/0.5000), identity (hello/hello/n=2/1.0000), orthogonal (abc/xyz/n=2/0.0000)
    - `Scenario: identical strings score 1.0`
    - `Scenario: both-empty strings score 1.0`
    - `Scenario: one-empty string scores 0.0`
    - `Scenario: score is symmetric`
    - `Scenario: rune-path Unicode pair` ("café"/"cafe"/n=2 → 0.6666666666666666 via the CosineRunes step)
    Use tolerance 0.0001 in the approximately step for the 4-decimal scenarios; pin the 17-digit values directly for the high-precision scenarios.

    Extend tests/bdd/steps/algorithms_steps.go by appending Cosine step methods: iComputeTheCosineScoreBetweenWithN, iComputeTheCosineRunesScoreBetweenWithN, iComputeTheSecondCosineScoreBetweenWithN, bothCosineScoresShouldBeEqual. Register their regexes inside InitializeScenario. Reuse the existing approximately-step regex.
  </action>
  <verify>
    <automated>make test-bdd 2>&1 | grep -i 'cosine\|Cosine' && (cd tests/bdd && go test -run 'TestFeatures' ./...)</automated>
  </verify>
  <done>
    `make test-bdd` exits 0 with the new Cosine scenarios green. Feature file covers identity, both-empty, one-empty, orthogonal, RV-C1/C2/C4 reference vectors at 17-digit precision, symmetry, rune-path Unicode (RV-C3).
  </done>
</task>

</tasks>

<threat_model>
## Trust Boundaries

| Boundary | Description |
|----------|-------------|
| caller → CosineScore / CosineScoreRunes | Untrusted (a, b string, n int) input crosses the API surface; library is pure-function with no I/O |
| CI matrix → testdata/golden/_staging/cosine.json | Multi-platform CI reads the canonical golden (after plan 05-05 merge) and asserts byte-identical match; any drift fails verify-determinism HARD |

## STRIDE Threat Register (ASVS Level 1)

| Threat ID | Category | Component | Disposition | Mitigation Plan |
|-----------|----------|-----------|-------------|-----------------|
| T-fuzz-panic | D | CosineScore / CosineScoreRunes on malformed UTF-8, extreme inputs, n=0 / n<0 | mitigate | Task 2 ships fuzz harnesses with ≥ 60s budget and seed corpus covering invalid UTF-8, identity, both-empty, one-empty, RV-C1..RV-C5, long inputs |
| T-complexity-attack | D | extractQGrams + sort.Strings on pathological inputs | accept | sort.Strings is O(k log k) where k = |intersection| ≤ min(la, lb) - n + 1. Combined with O(la + lb) extraction, the total complexity is bounded by input length. PERF-01 budget enforced via bench regression detection |
| T-float-determinism-cosine | T (Tampering of float reduction order across architectures) | CosineScore dot-product reduction + math.Sqrt + final division | mitigate | LOAD-BEARING. (1) Sorted-key iteration via sort.Strings makes the dot-product reduction order canonical regardless of input map iteration order — CONTEXT.md §3 LOCKED. (2) Explicit `(x*y) + z` parenthesisation per DET-06; FMA-risk footnote documented inline per RESEARCH.md §3 OQ-1. (3) math.Sqrt only — IEEE-754 correctly rounded on all four CI platforms per RESEARCH.md §3.5. (4) Final division on IEEE-754 correctly rounded floats. (5) Cross-platform CI matrix verifies byte-identical golden output via _staging/cosine.json (9 entries spanning ASCII + Unicode at n ∈ {2, 3, 4}) merged in plan 05-05. (6) TestCosine_SortedKeyIteration (Task 1) pins per-process determinism. (7) TestProp_CosineScore_DeterministicAcrossRuns (Task 2) reinforces per-process bit-stability across 1000 calls. If CI matrix divergence ever appears, remediation = `float64(x*y) + z` cast per RESEARCH.md §3.4 — documented as a footnote, no code change required |
| T-map-iteration-leak | T | Cosine intersection-key slice construction | mitigate | The intersection-key slice is built by iterating ONE map and checking membership in the other — the BUILT slice may be in any order, but it is IMMEDIATELY sorted via sort.Strings before any output-affecting use. DET-03 satisfied. The norm computations iterate map VALUES in any order, but integer addition is exactly associative — no determinism risk |
</threat_model>

<verification>
- `go build ./...` succeeds.
- `go test -run 'TestCosine|TestProp_Cosine|TestDispatch_CosineRegistered|TestGolden_Cosine_Staging|ExampleCosine' ./...` exits 0.
- `go test -bench=BenchmarkCosine -benchmem -benchtime=1x ./...` reports alloc count within the documented ≤ 5 allocs/op budget.
- `go test -fuzz=FuzzCosineScore -fuzztime=60s ./...` and `go test -fuzz=FuzzCosineScoreRunes -fuzztime=60s ./...` complete without failure (10s smoke for per-task gate).
- `make test-bdd` green; Cosine scenarios visible.
- `bash scripts/verify-license-headers.sh` exits 0.
- `bash scripts/verify-no-runtime-deps.sh` exits 0.
- `! grep -q "^func init" cosine.go`.
- `grep -q "Source: Salton, G." cosine.go`.
- `grep -q "sort.Strings" cosine.go && grep -q "math.Sqrt" cosine.go`.
- `grep -v '^#' cosine.go | ! grep -E "math\.(Pow|Log|Exp|FMA)"` (DET-06 gate).
- `grep -q "CosineScore" llms.txt && grep -q "CosineScoreRunes" llms.txt`.
- `grep -c "RV-C" cosine_test.go` returns ≥ 4 (RV-C1 + RV-C2 + RV-C3 + RV-C4 derivation block markers; RV-C5 optional).
- `make coverage-check` confirms cosine.go ≥ 90% per-file coverage.
- `make check` exits 0.
</verification>

<success_criteria>
- All three tasks complete; all listed verification commands green.
- testdata/golden/_staging/cosine.json exists, canonical-marshalled, 9 alphabetically-sorted entries spanning ASCII + Unicode at n ∈ {2, 3, 4} per CONTEXT.md §1 LOCKED.
- testdata/fuzz/FuzzCosineScore/seed-001 and FuzzCosineScoreRunes/seed-001 byte-stable.
- Public surface: exactly TWO new exported functions (CosineScore, CosineScoreRunes).
- Dispatch slot wired with default n=3 wrapper.
- LOAD-BEARING DETERMINISM: cosine.go iterates intersection keys via sort.Strings; dot-product reduction uses `(x*y)+z` parenthesisation; math.Sqrt only; FMA footnote present.
- Plan 05-05 finalisation can begin once plans 05-02 and 05-04 also ship — plan 05-05 will merge the Cosine staging golden into algorithms.json and validate cross-platform byte-identical output via `make verify-determinism`.
</success_criteria>

<output>
After completion, create `.planning/phases/05-q-gram-algorithms/05-03-cosine-SUMMARY.md` per the GSD summary template. Note any FMA-related findings or CI matrix observations in the SUMMARY for downstream phases.
</output>
