---
phase: 05-q-gram-algorithms
plan: 01
type: execute
wave: 1
depends_on: []
files_modified:
  - errors.go
  - q_gram.go
  - qgram_jaccard.go
  - dispatch_qgram_jaccard.go
  - qgram_jaccard_test.go
  - qgram_jaccard_bench_test.go
  - qgram_jaccard_fuzz_test.go
  - q_gram_test.go
  - props_test.go
  - example_test.go
  - algoid_test.go
  - algorithms_golden_test.go
  - testdata/golden/_staging/qgram_jaccard.json
  - testdata/fuzz/FuzzQGramJaccardScore/seed-001
  - testdata/fuzz/FuzzQGramJaccardScoreRunes/seed-001
  - tests/bdd/features/qgram_jaccard.feature
  - tests/bdd/steps/algorithms_steps.go
  - llms.txt
  - llms-full.txt
autonomous: true
requirements:
  - QGRAM-01
  - QGRAM-02
tags: [q-gram-foundation, ukkonen-1992, jaccard-1912, error-sentinels, extract-qgrams, byte-and-rune-paths, dispatch-registration, property-tests, fuzz, benchmark, bdd, staging-golden, llms-sync]

must_haves:
  truths:
    # Foundation (QGRAM-01)
    - "q_gram.go declares unexported `extractQGrams(s string, n int) map[string]int` and `extractQGramsRunes(s string, n int) map[string]int` — public surface remains the four algorithm functions only (CONTEXT.md §2 LOCKED)"
    - "Both extractors return a non-nil empty `map[string]int{}` for `n < 1`, for `len(s) < n` (byte path) or `runeCount < n` (rune path), and for the empty string"
    - "Both extractors use `make(map[string]int, expectedCap)` with capacity hint sized to `max(0, len(s)-n+1)` for the byte path and `max(0, runeCount-n+1)` for the rune path (RESEARCH.md §4.2)"
    - "Multiset semantics: repeated q-grams increment the count; e.g. `extractQGrams(\"AAAA\", 2) == {AA:3}` (overlapping windows)"
    - "q_gram.go file-level godoc cites Ukkonen 1992 §2-3 as the PRIMARY source for q-gram extraction and contains the source-origin statement block (Primary / Cross-validation / Tie-break / GPL-LGPL: none / Code copied: none)"
    - "q_gram.go contains NO init() function — determinism-reviewer flags any init() as BLOCKING (DET-04 / PITFALLS §14)"
    # Error sentinels (OQ-2 BLOCKER per RESEARCH.md §6)
    - "errors.go declares `var ErrInvalidQGramSize = errors.New(\"fuzzymatch: invalid q-gram size\")` with a godoc block explaining it is returned by Phase 8 Scorer options (e.g. WithQGramJaccardAlgorithm) when n < 1; direct algorithm calls panic instead (CONTEXT.md §5)"
    - "errors.go declares `var ErrInvalidTverskyParam = errors.New(\"fuzzymatch: invalid tversky parameter\")` with a godoc block covering α < 0, β < 0, α+β == 0 cases; direct TverskyScore calls panic instead (CONTEXT.md §5)"
    - "Both sentinels are placed alphabetically with the existing ErrInvalidAlgorithm / ErrInvalidConfiguration / ErrInvalidInput cluster and follow the exact godoc style of those existing sentinels"
    # Q-Gram Jaccard algorithm (QGRAM-02)
    - "A caller can `import fuzzymatch` and call QGramJaccardScore(\"AGCT\", \"AGCTAGCT\", 2) and receive `3.0/7.0 = 0.42857142857142855` (RV-J1 — Ukkonen 1992 §3 worked example)"
    - "QGramJaccardScore(\"hello\", \"hello\", 2) == 1.0 (identity)"
    - "QGramJaccardScore(\"abc\", \"xyz\", 2) == 0.0 (no overlap; both q-gram sets non-empty)"
    - "QGramJaccardScore(\"abcd\", \"abxy\", 2) == 0.2 (RV-J4 — single-shared-q-gram = 1/5)"
    - "QGramJaccardScore(\"\", \"\", 2) == 1.0 (both-empty identity convention)"
    - "QGramJaccardScore(\"\", \"abc\", 2) == 0.0 (one-empty convention)"
    - "QGramJaccardScore(x, x, n) == 1.0 for every non-empty x and every n ≥ 1"
    - "QGramJaccardScore(a, b, n) == QGramJaccardScore(b, a, n) for every (a, b, n) — set-Jaccard is symmetric"
    - "QGramJaccardScoreRunes(\"café\", \"cafe\", 2) returns 0.5 (rune-bigrams `[ca, af, fé]` vs `[ca, af, fe]`; |∩|=2, |∪|=4 → 2/4)"
    - "QGramJaccardScore(\"hello\", \"hello\", 1000) panics with message containing `invalid q-gram size` (direct-call discipline per CONTEXT.md §5) — wait, n=1000 is valid but the result is 1.0 (both-empty after `len(s) < n` short-circuit); the panic is reserved for n < 1. Re-state: QGramJaccardScore(\"hello\", \"hello\", 0) panics with message containing `invalid q-gram size`; same for QGramJaccardScore(\"hello\", \"hello\", -1)"
    - "dispatch[AlgoQGramJaccard] is non-nil after package load and dispatches to a wrapper that calls QGramJaccardScore(a, b, 3) (default n=3 per CONTEXT.md Deferred §4 — n is not part of the dispatch signature)"
    # Determinism + correctness gates (DET-03, DET-06)
    - "QGramJaccardScore never returns NaN, +Inf, -Inf, or -0 for any input — verified by TestProp_QGramJaccardScore_NoNaN / _NoInf / _NoNegativeZero (byte + rune)"
    - "QGramJaccardScore is deterministic: 1000 sequential calls on the same input produce byte-identical output (TestProp_QGramJaccardScore_DeterministicAcrossRuns — PITFALLS §14 closure carried forward from Phase 4)"
    - "No map iteration on the output path in q_gram.go or qgram_jaccard.go — verified by inspection plus the determinism property test; intersection and union sizes are computed by `len(...)` and arithmetic on map lengths, not by iteration over keys"
    - "No `math.Pow`, `math.Log`, `math.Exp`, `math.FMA` anywhere in q_gram.go or qgram_jaccard.go (only `+`, `-`, `*`, `/`, comparisons, `float64()` casts, and `len(...)`) — DET-06 gate"
    # Public-surface + meta-test discipline
    - "Public surface added by this plan: exactly two new exported symbols (QGramJaccardScore, QGramJaccardScoreRunes) plus two new error sentinels (ErrInvalidQGramSize, ErrInvalidTverskyParam) — pre-existing AlgoQGramJaccard constant already in algoid.go slot"
    - "FuzzQGramJaccardScore is panic-free, score-in-[0,1], NaN/Inf-free for any (a, b) including invalid UTF-8 (\\xff\\xfe), with `n` constrained to [1, 8] via the fuzz body (or coerced to 3 if out of range); same for FuzzQGramJaccardScoreRunes"
    - "testdata/fuzz/FuzzQGramJaccardScore/seed-001 and testdata/fuzz/FuzzQGramJaccardScoreRunes/seed-001 exist in `go test fuzz v1` literal format with byte-stable formatting (Phase 3 IN-06 closure)"
    - "tests/bdd/features/qgram_jaccard.feature exists with: canonical reference-vector Scenario Outline (RV-J1, RV-J2, RV-J3, RV-J4), identity scenario, both-empty scenario, one-empty scenario, symmetry scenario, AND a rune-path Unicode scenario"
    - "tests/bdd/steps/algorithms_steps.go appends QGramJaccard step bindings (iComputeTheQGramJaccardScoreBetweenWithN / Runes variant / second-score variants / equality assertion) and their ctx.Step regex registrations inside InitializeScenario"
    - "testdata/golden/_staging/qgram_jaccard.json exists, produced by TestGolden_QGramJaccard_Staging via assertGoldenStaging; entries sorted alphabetically by Name; includes at minimum 8-10 entries covering identity, both-empty, one-empty, no-overlap, RV-J1/J3/J4 (byte), single-shared-q-gram, and one rune-path entry (RV-J5)"
    - "algoid_test.go contains a new TestDispatch_QGramJaccardRegistered asserting dispatch[AlgoQGramJaccard] is non-nil; the registered map in TestDispatch_UnregisteredSlotsAreNil adds `int(AlgoQGramJaccard): true`"
    - "ExampleQGramJaccardScore and ExampleQGramJaccardScoreRunes appended to example_test.go; `// Output:` blocks match byte-for-byte"
    - "llms.txt lists `QGramJaccardScore` and `QGramJaccardScoreRunes` (the two new exported symbols) plus `ErrInvalidQGramSize` and `ErrInvalidTverskyParam` (the two new sentinels). The AlgoID constant AlgoQGramJaccard is ALREADY listed (declared in Phase 1)"
    - "llms-full.txt has parallel entries with one-line rationales for each new symbol"
    - "Coverage on q_gram.go and qgram_jaccard.go ≥ 90%; 100% on the public QGramJaccardScore + QGramJaccardScoreRunes surface"
    - "Apache-2.0 header present on every new .go file (scripts/verify-license-headers.sh exits 0)"
    - "`make check` and `make test-bdd` exit 0 at end of plan"
  artifacts:
    - path: "errors.go"
      provides: "Two new error sentinels (ErrInvalidQGramSize, ErrInvalidTverskyParam) with godoc blocks following the existing Err* style"
      contains: "ErrInvalidQGramSize"
    - path: "q_gram.go"
      provides: "Unexported extractQGrams + extractQGramsRunes helpers — single source of truth for the q-gram tier; no public surface"
      min_lines: 60
      contains: "Source: Ukkonen, E. (1992)"
    - path: "qgram_jaccard.go"
      provides: "QGramJaccardScore + QGramJaccardScoreRunes (two new public functions); cites Ukkonen 1992 + Jaccard 1912 as primary sources"
      min_lines: 100
      contains: "Source: Ukkonen, E. (1992)"
    - path: "dispatch_qgram_jaccard.go"
      provides: "Package-load-time registration of a default-n=3 QGramJaccardScore wrapper into dispatch[AlgoQGramJaccard]"
      contains: "dispatch[AlgoQGramJaccard]"
    - path: "qgram_jaccard_test.go"
      provides: "Unit tests for identity, both-empty, one-empty, no-overlap, RV-J1/J2/J3/J4 reference vectors (byte + rune), and direct-call panic on n < 1"
    - path: "q_gram_test.go"
      provides: "Unit tests for the shared extractor: empty-string returns empty map, n < 1 returns empty map, multiset increment for overlapping windows, n > len returns empty map, byte path vs rune path divergence on multi-byte input"
    - path: "qgram_jaccard_bench_test.go"
      provides: "Benchmarks: BenchmarkQGramJaccardScore_{ASCII_Short, ASCII_Medium, ASCII_Long} + BenchmarkQGramJaccardScoreRunes_Unicode_Short — alloc-asserted with b.ReportAllocs() + var sink anti-DCE"
    - path: "qgram_jaccard_fuzz_test.go"
      provides: "FuzzQGramJaccardScore and FuzzQGramJaccardScoreRunes — panic-free, NaN/Inf-free, score-in-[0,1] for any input including invalid UTF-8; n constrained to [1, 8]"
    - path: "props_test.go"
      provides: "Appended QGramJaccard property-test block: TestProp_QGramJaccardScore_{RangeBounds, Identity, Symmetric, NoNaN, NoInf, NoNegativeZero} for BOTH byte and rune surfaces (12 property tests) plus TestProp_QGramJaccardScore_DeterministicAcrossRuns"
    - path: "example_test.go"
      provides: "Appended ExampleQGramJaccardScore + ExampleQGramJaccardScoreRunes runnable godoc examples"
    - path: "algoid_test.go"
      provides: "Appended TestDispatch_QGramJaccardRegistered; updated `registered` map in TestDispatch_UnregisteredSlotsAreNil flipping AlgoQGramJaccard slot to true"
    - path: "algorithms_golden_test.go"
      provides: "Appended buildQGramJaccardStagingEntries + TestGolden_QGramJaccard_Staging — produces _staging/qgram_jaccard.json via assertGoldenStaging; the algorithms-merge slice is NOT updated here (plan 05-05 owns the merge)"
    - path: "testdata/golden/_staging/qgram_jaccard.json"
      provides: "Per-algorithm staging file; sorted by Name; merged into algorithms.json by plan 05-05"
      contains: "QGramJaccard_AGCT_AGCTAGCT"
    - path: "testdata/fuzz/FuzzQGramJaccardScore/seed-001"
      provides: "Fuzz seed corpus file in `go test fuzz v1` literal format"
    - path: "testdata/fuzz/FuzzQGramJaccardScoreRunes/seed-001"
      provides: "Fuzz seed corpus file for the rune-path harness"
    - path: "tests/bdd/features/qgram_jaccard.feature"
      provides: "Gherkin feature with scenarios: canonical reference-vector outline, identity, both-empty, one-empty, symmetry, rune-path Unicode"
    - path: "tests/bdd/steps/algorithms_steps.go"
      provides: "Appended QGramJaccard step methods on AlgorithmContext + ctx.Step registrations"
    - path: "llms.txt"
      provides: "Appended 2 function entries (QGramJaccardScore, QGramJaccardScoreRunes) + 2 error sentinel entries (ErrInvalidQGramSize, ErrInvalidTverskyParam) under the catalogue section"
    - path: "llms-full.txt"
      provides: "Parallel entries with one-line rationales"
  key_links:
    - from: "qgram_jaccard.go (QGramJaccardScore + QGramJaccardScoreRunes)"
      to: "q_gram.go (extractQGrams + extractQGramsRunes)"
      via: "Direct call into the unexported extractor; QGramJaccardScore computes |A∩B| and |A∪B| via map-length arithmetic on the two returned multisets — no iteration on output path"
      pattern: "extractQGrams(Runes)?\\("
    - from: "dispatch_qgram_jaccard.go"
      to: "algoid.go (AlgoQGramJaccard declared at line 112)"
      via: "package-level `var _ = func() bool { dispatch[AlgoQGramJaccard] = func(a, b string) float64 { return QGramJaccardScore(a, b, 3) }; return true }()` — uses default n=3 per CONTEXT.md Deferred §4"
      pattern: "dispatch\\[AlgoQGramJaccard\\]"
    - from: "errors.go (ErrInvalidQGramSize, ErrInvalidTverskyParam)"
      to: "Phase 8 Scorer (future) — WithQGramJaccardAlgorithm + WithTverskyAlgorithm options will errors.Is on these sentinels"
      via: "Sentinel declarations made available v1 per OQ-2 RESOLUTION; direct-call algorithm functions panic with a message containing the sentinel text per CONTEXT.md §5"
      pattern: "ErrInvalid(QGramSize|TverskyParam)"
    - from: "tests/bdd/features/qgram_jaccard.feature"
      to: "tests/bdd/steps/algorithms_steps.go (QGramJaccard step bindings)"
      via: "godog regex registration `^I compute the QGramJaccard(Runes)? score between \"([^\"]*)\" and \"([^\"]*)\" with n (\\d+)$`"
      pattern: "ctx\\.Step\\(.+QGramJaccard"
---

<objective>
Lay the q-gram tier foundation: ship `q_gram.go` (the unexported `extractQGrams` + `extractQGramsRunes` helpers consumed by all four q-gram algorithms), pre-declare the two missing error sentinels (`ErrInvalidQGramSize`, `ErrInvalidTverskyParam` — OQ-2 BLOCKER per RESEARCH.md §6), and implement the first q-gram algorithm — Q-Gram Jaccard (Ukkonen 1992 §3 / Jaccard 1912) — with the full Phase 2/3/4 quality bar (unit + property + fuzz + bench + BDD + staging golden + dispatch + example + llms.txt). Both byte and rune surfaces ship together (no Runes-deferred shortcut). Closes QGRAM-01 (foundation) and QGRAM-02 (Q-Gram Jaccard).

Purpose: provide the shared infrastructure that plans 05-02 (Sørensen-Dice), 05-03 (Cosine), and 05-04 (Tversky) depend on, plus the simplest algorithmic consumer of that infrastructure as the proof-of-life. Q-Gram Jaccard's set-Jaccard formula (`|A∩B|/|A∪B|`) is the textbook q-gram algorithm and the lowest-risk pilot for the new shared helper. The error-sentinel gap (CONTEXT.md `code_context` claimed they were pre-declared; they are NOT — verified by `grep -n "ErrInvalid" errors.go`) is closed here so plans 05-02..05-04 and the future Phase 8 Scorer can rely on them without re-asking.

Output: 18 new/modified files (5 new source/test files, plus extensions to 7 existing append-only files, plus 4 new test/fixture files in testdata + tests/bdd, plus errors.go modification). Single new dispatch slot wired; first q-gram staging golden committed (merged into algorithms.json by plan 05-05).
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
@.planning/phases/04-remaining-character-gestalt/04-CONTEXT.md
@.planning/phases/04-remaining-character-gestalt/04-PATTERNS.md
@.planning/phases/04-remaining-character-gestalt/04-01-strcmp95-PLAN.md
@.planning/phases/04-remaining-character-gestalt/04-03-ratcliff-obershelp-PLAN.md
@.claude/skills/algorithm-correctness-standards/SKILL.md
@.claude/skills/algorithm-licensing-standards/SKILL.md
@.claude/skills/determinism-standards/SKILL.md
@.claude/skills/performance-standards/SKILL.md
@.claude/skills/go-coding-standards/SKILL.md
@.claude/skills/go-testing-standards/SKILL.md
@algoid.go
@errors.go
@jaro.go
@ratcliff_obershelp.go
@dispatch_ratcliff_obershelp.go
</context>

<interfaces>
<!-- Key types/functions executor MUST use without rediscovering. Extracted from existing source. -->

From algoid.go (slot already declared at line 112; do NOT modify):
```go
const AlgoQGramJaccard AlgoID = ... // line 112; String() case at line 233-234
```

From errors.go (current state — three sentinels exist; add two more alphabetically nearby):
```go
var ErrInvalidAlgorithm = errors.New("fuzzymatch: invalid algorithm")
var ErrInvalidConfiguration = errors.New("fuzzymatch: invalid configuration")
var ErrInvalidInput = errors.New("fuzzymatch: invalid input")
// ADD:
var ErrInvalidQGramSize = errors.New("fuzzymatch: invalid q-gram size")
var ErrInvalidTverskyParam = errors.New("fuzzymatch: invalid tversky parameter")
```

Public surface added by this plan (two new functions + two new sentinels):
```go
// QGramJaccardScore returns the Jaccard coefficient of the q-gram
// multiset of a and b (Ukkonen 1992 §3; Jaccard 1912).
// Score = |QA ∩ QB| / |QA ∪ QB|, in [0.0, 1.0].
// Panics on n < 1 with a message containing "invalid q-gram size".
func QGramJaccardScore(a, b string, n int) float64

func QGramJaccardScoreRunes(a, b string, n int) float64
```

Unexported helpers internal to q_gram.go:
```go
// extractQGrams returns the multiset of overlapping length-n BYTE q-grams.
// Returns a non-nil empty map for n < 1, empty input, or len(s) < n.
// Multiset semantics: repeated q-grams increment the count.
// MUST NOT be iterated by callers on any output path (DET-03).
func extractQGrams(s string, n int) map[string]int

// extractQGramsRunes returns the multiset of overlapping length-n RUNE q-grams.
// Keys are UTF-8 encoded strings of length-n rune slices.
func extractQGramsRunes(s string, n int) map[string]int
```

Dispatch wiring shape (matches dispatch_ratcliff_obershelp.go verbatim — default n=3):
```go
var _ = func() bool {
    dispatch[AlgoQGramJaccard] = func(a, b string) float64 {
        return QGramJaccardScore(a, b, 3)
    }
    return true
}()
```
</interfaces>

<tasks>

<task type="auto" tdd="true">
  <name>Task 1: Implement q_gram.go + errors.go sentinels + qgram_jaccard.go + dispatch + unit tests + staging golden</name>
  <files>errors.go, q_gram.go, qgram_jaccard.go, dispatch_qgram_jaccard.go, q_gram_test.go, qgram_jaccard_test.go, testdata/golden/_staging/qgram_jaccard.json, algorithms_golden_test.go, algoid_test.go, example_test.go</files>
  <read_first>
    - errors.go (full file — confirm only the three existing sentinels are declared; identify the alphabetical insertion point for the two new sentinels)
    - q_gram.go (current state — confirm it does NOT exist; creating new)
    - qgram_jaccard.go (current state — confirm it does NOT exist; creating new)
    - .planning/phases/05-q-gram-algorithms/05-CONTEXT.md §2 (API surface — INTERNAL helpers only) and §3 (sorted-keys iteration order — applies to Cosine but the discipline is shared)
    - .planning/phases/05-q-gram-algorithms/05-RESEARCH.md §1.1 (Q-Gram Jaccard formula), §2.1 (RV-J1..RV-J4 derivations), §4.1-4.3 (allocation budget + capacity hint + no stack-buffer fast path), §6 OQ-2 (sentinel BLOCKER), §7 (decomposition recommendation), "Code Examples → Q-gram extraction helper"
    - algoid.go lines 109-128 (AlgoQGramJaccard slot — line 112) and lines 233-240 (String() case — already wired)
    - jaro.go (analog for byte+rune file shape; godoc style)
    - ratcliff_obershelp.go (analog for the byte+rune file shape from Phase 4)
    - dispatch_ratcliff_obershelp.go (full file — exact template for dispatch_qgram_jaccard.go; the wrapper-closure pattern for algorithms that need extra parameters at dispatch time)
    - algorithms_golden_test.go (last ~300 lines — the buildXxxStagingEntries + TestGolden_Xxx_Staging template; identify the Phase 4 RatcliffObershelp build function as the closest analog)
    - algoid_test.go (last ~50 lines — TestDispatch_XxxRegistered template + TestDispatch_UnregisteredSlotsAreNil registered map shape)
    - example_test.go (last ~30 lines — ExampleXxxScore + ExampleXxxScoreRunes templates from Phase 4)
    - .planning/phases/04-remaining-character-gestalt/04-PATTERNS.md (the entire file — every per-file template that Phase 5 inherits without re-discussion)
  </read_first>
  <behavior>
    - errors.go declares ErrInvalidQGramSize + ErrInvalidTverskyParam with godoc blocks
    - extractQGrams("AGCT", 2) returns {AG:1, GC:1, CT:1}; len() == 3
    - extractQGrams("AGCTAGCT", 2) returns {AG:2, GC:2, CT:2, TA:1}; len() == 4 (4 distinct keys; total multiset weight 7)
    - extractQGrams("", 2) returns map[string]int{} (non-nil, empty)
    - extractQGrams("ab", 5) returns map[string]int{}
    - extractQGrams("hello", 0) returns map[string]int{} (graceful empty rather than panic — algorithm-level validation handles n<1)
    - extractQGrams("hello", -1) returns map[string]int{}
    - extractQGramsRunes("café", 2) returns {ca:1, af:1, fé:1} — three rune-bigrams, "fé" as a multi-byte string key
    - QGramJaccardScore("AGCT", "AGCTAGCT", 2) returns 0.42857142857142855 within 1e-15 (RV-J1)
    - QGramJaccardScore("hello", "hello", 2) == 1.0 (RV-J2 identity)
    - QGramJaccardScore("abc", "xyz", 2) == 0.0 (RV-J3 no-overlap)
    - QGramJaccardScore("abcd", "abxy", 2) == 0.2 (RV-J4 single-shared = 1/5)
    - QGramJaccardScore("", "", 2) == 1.0 (both-empty)
    - QGramJaccardScore("", "abc", 2) == 0.0 (one-empty)
    - QGramJaccardScore("abc", "", 2) == 0.0 (one-empty symmetric)
    - QGramJaccardScoreRunes("café", "cafe", 2) == 0.5 (rune-path: |∩|=2 {ca, af}, |∪|=4 {ca, af, fé, fe})
    - QGramJaccardScore("hello", "hello", 0) panics with message containing "invalid q-gram size"
    - QGramJaccardScore("hello", "hello", -1) panics with message containing "invalid q-gram size"
    - dispatch[AlgoQGramJaccard]("AGCT", "AGCTAGCT") returns QGramJaccardScore("AGCT", "AGCTAGCT", 3) (dispatch default n=3)
    - dispatch[AlgoQGramJaccard] is non-nil after package load (TestDispatch_QGramJaccardRegistered)
    - TestGolden_QGramJaccard_Staging produces testdata/golden/_staging/qgram_jaccard.json with 8-10 alphabetically-sorted entries
  </behavior>
  <action>
    Step A — Extend errors.go. Add two new `var Err...` declarations alphabetically with the existing three. Each declaration gets a godoc block in the same style as ErrInvalidAlgorithm / ErrInvalidConfiguration / ErrInvalidInput. ErrInvalidQGramSize's godoc explicitly notes: "Returned by Phase 8 Scorer options (e.g. `WithQGramJaccardAlgorithm`) when `n < 1`. Direct algorithm calls (`QGramJaccardScore`, `SorensenDiceScore`, `CosineScore`, `TverskyScore`) panic with a message containing this sentinel's text per CONTEXT.md §5." ErrInvalidTverskyParam's godoc covers α < 0, β < 0, α+β == 0. Run `go build ./...` to verify the package still compiles.

    Step B — Create q_gram.go. File order:
    (a) Apache-2.0 header (copy ratcliff_obershelp.go's header verbatim).
    (b) File-level doc block: `// Package fuzzymatch — q_gram.go provides the unexported q-gram extraction helpers shared by Q-Gram Jaccard, Sørensen-Dice, Cosine, and Tversky.` Add source-origin statement block: PRIMARY = Ukkonen, E. (1992). "Approximate string-matching with q-grams and maximal matches." Theoretical Computer Science 92(1):191–211 (especially §2 q-gram definition and §3 set Jaccard formulation). Cross-validation = none required (hand-derived RV-J1..RV-J4 in qgram_jaccard_test.go). Tie-break = none. GPL-LGPL: none. Code copied: none.
    (c) `package fuzzymatch`.
    (d) `func extractQGrams(s string, n int) map[string]int` — exactly the implementation pattern from RESEARCH.md "Code Examples → Q-gram extraction helper" with capacity hint `make(map[string]int, len(s)-n+1)`. Returns non-nil empty map for n < 1, empty input, or len(s) < n.
    (e) `func extractQGramsRunes(s string, n int) map[string]int` — analogous, but operates on `[]rune(s)`. Use the rune slice for window indexing; produce `string(runes[i:i+n])` as the multi-byte UTF-8 key. Capacity hint `make(map[string]int, runeCount-n+1)`.
    (f) Godoc on both functions notes DET-03 ("MUST NOT be iterated by callers on any output path") and the multiset semantics (overlapping windows; repeated q-grams accumulate).

    Step C — Create qgram_jaccard.go. File order:
    (a) Apache-2.0 header.
    (b) File-level doc: cite Ukkonen 1992 §3 as PRIMARY and Jaccard 1912 as the underlying set-coefficient origin. Source-origin statement block (Primary / Cross-validation: hand-derived RV-J1 from Ukkonen 1992 §3 + RV-J2/J3/J4 from primary-source first-principles derivation in `qgram_jaccard_test.go` / Tie-break: none / GPL-LGPL: none / Code copied: none).
    (c) `package fuzzymatch`.
    (d) `func QGramJaccardScore(a, b string, n int) float64` — godoc with the J(a, b) = |QA∩QB| / |QA∪QB| formula, the [0, 1] range, the identity / both-empty / one-empty conventions, and the direct-call panic-on-n<1 contract. Body: identity short-circuit `if a == b { return 1.0 }` (covers both-empty and identical); one-empty short-circuit (if exactly one is empty → 0.0); `if n < 1 { panic("fuzzymatch: invalid q-gram size") }`; call extractQGrams on both inputs; compute intersection cardinality via map-iteration on the SMALLER map (this is internal scratch — counts the matches but does NOT produce ordered output; the result is an int, not a slice — DET-03 satisfied because no output ORDER depends on iteration); compute union cardinality = len(qa) + len(qb) - intersection; return float64(intersection) / float64(union) using explicit `(num) / (denom)` parenthesisation per DET-06.
    (e) `func QGramJaccardScoreRunes(a, b string, n int) float64` — analogous but calls extractQGramsRunes.
    (f) Use only `+`, `-`, `*`, `/`, comparisons, `float64()` casts — NO math.Pow/Log/Exp/FMA.

    Step D — Create dispatch_qgram_jaccard.go per dispatch_ratcliff_obershelp.go shape (~22 lines). Use a closure to bind n=3:
    ```go
    var _ = func() bool {
        dispatch[AlgoQGramJaccard] = func(a, b string) float64 {
            return QGramJaccardScore(a, b, 3)
        }
        return true
    }()
    ```
    NO init(). Apache-2.0 header. Add a brief godoc explaining the dispatch wrapper uses default n=3 per CONTEXT.md Deferred §4 ("Specific n overrides happen via the Phase 8 Scorer option layer").

    Step E — Create q_gram_test.go covering the shared extractor:
    - TestExtractQGrams_Empty (empty string → empty map; n=0 → empty; n=-1 → empty; n > len → empty)
    - TestExtractQGrams_Multiset (overlapping windows accumulate counts; e.g. "AAAA"/n=2 → {AA:3})
    - TestExtractQGrams_AGCTAGCT (Ukkonen 1992 §3 worked-example fixture: {AG:2, GC:2, CT:2, TA:1} for n=2)
    - TestExtractQGramsRunes_MultiByte ("café"/n=2 → {ca:1, af:1, fé:1}; assert "fé" is present as a string key)
    - TestExtractQGramsRunes_DivergesFromBytes ("é"/n=1 — byte path returns a 2-byte key per byte-bigram; rune path returns a single rune; document the divergence)
    - Stdlib testing only.
    Since the extractors are unexported, use the existing `export_test.go` re-export pattern (look at existing file — extend with `var ExtractQGramsForTest = extractQGrams; var ExtractQGramsRunesForTest = extractQGramsRunes` if needed). If export_test.go does not exist, create one with the Apache-2.0 header + `package fuzzymatch` + the two `var ...ForTest = ...` declarations.

    Step F — Create qgram_jaccard_test.go covering the algorithm:
    - TestQGramJaccard_BothEmpty (RV-J5 — both-empty → 1.0)
    - TestQGramJaccard_OneEmpty (both directions — 0.0)
    - TestQGramJaccard_Identical (RV-J2 — identical → 1.0)
    - TestQGramJaccard_ReferenceVectors — table-driven `t.Run(tt.name, ...)`. Each row references the RV-JN identifier from RESEARCH.md §2.1 in the test name and embeds the formula derivation as a t.Logf or godoc-style block comment (e.g. RV-J1 reproduces `|QA∩QB|=3, |QA∪QB|=7, J=3/7=0.42857142857142855` in the test comment). Tolerance: `math.Abs(got - want) <= 1e-15` for irrational values; exact float comparison for rational values like 0.2 and 0.0.
    - TestQGramJaccard_Symmetric — assert Score(a, b, n) == Score(b, a, n) on 3 hand-picked pairs (set-Jaccard is exactly symmetric — equality, not "within tolerance")
    - TestQGramJaccard_PanicsOnInvalidN — table-driven `defer recover()` covering n=0, n=-1, n=-100. Assert the panic message contains "invalid q-gram size".
    - TestQGramJaccardRunes_CafeReference (RV-J5-Runes — "café"/"cafe"/n=2 → 0.5 — derivation in test comment)
    - TestQGramJaccard_ZeroAllocsOnHotPath via testing.AllocsPerRun(100, ...) on a short ASCII pair — DOCUMENT the actual alloc bound (q-gram extraction allocates maps; per RESEARCH.md §4.1, the budget is ≤ 4 allocs per call). Assert ≤ 4 allocs/op rather than 0; the test FAILS if the implementation exceeds the budget.
    - Stdlib testing only.

    Step G — Append buildQGramJaccardStagingEntries + TestGolden_QGramJaccard_Staging to algorithms_golden_test.go. Entries (Name sorted alphabetically): `QGramJaccard_AGCT_AGCTAGCT` (RV-J1; n=2; score=0.42857142857142855), `QGramJaccard_abcd_abxy` (RV-J4; n=2; score=0.2), `QGramJaccard_both_empty` (n=2; score=1.0), `QGramJaccard_cafe_runes` (RV-J5-Runes; rune path; n=2; score=0.5), `QGramJaccard_identical` (RV-J2; "hello"/"hello"/n=2; score=1.0), `QGramJaccard_n_too_large` (RV-J6; "ab"/"abc"/n=5; score=1.0 by both-empty convention), `QGramJaccard_no_overlap` (RV-J3; "abc"/"xyz"/n=2; score=0.0), `QGramJaccard_one_empty` (""/"abc"/n=2; score=0.0). Run `go test -run TestGolden_QGramJaccard_Staging -update ./...` to materialise the file via assertGoldenStaging → CanonicalMarshalForTest. Do NOT extend the algorithms-merge slice — plan 05-05 owns that.

    Step H — Append TestDispatch_QGramJaccardRegistered to algoid_test.go and extend the `registered` map in TestDispatch_UnregisteredSlotsAreNil to flip AlgoQGramJaccard slot to true.

    Step I — Append ExampleQGramJaccardScore + ExampleQGramJaccardScoreRunes to example_test.go. The byte-path example uses Ukkonen 1992 §3's worked-example pair (`fmt.Printf("%.4f\n", fuzzymatch.QGramJaccardScore("AGCT", "AGCTAGCT", 2))`); the rune-path example uses the café pair. Capture exact stdout once via `go test -run Example ./...` and paste into the `// Output:` block byte-for-byte.

    Step J — Update llms.txt and llms-full.txt. Append two new function entries to llms.txt under the catalogue section + two new error-sentinel entries. Add parallel entries to llms-full.txt with one-line rationales. The AlgoID constant AlgoQGramJaccard is ALREADY listed (declared in Phase 1) — do NOT duplicate.
  </action>
  <verify>
    <automated>cd /Users/johnny/Development/fuzzymatch && go build ./... && go test -run 'TestExtractQGrams|TestQGramJaccard|TestDispatch_QGramJaccardRegistered|TestDispatch_UnregisteredSlotsAreNil|TestGolden_QGramJaccard_Staging|ExampleQGramJaccardScore' ./... && bash scripts/verify-license-headers.sh && ! grep -q "^func init" q_gram.go && ! grep -q "^func init" qgram_jaccard.go && grep -q "Source: Ukkonen, E. (1992)" q_gram.go && grep -q "Source: Ukkonen, E. (1992)" qgram_jaccard.go && grep -q "ErrInvalidQGramSize" errors.go && grep -q "ErrInvalidTverskyParam" errors.go && grep -q "dispatch\[AlgoQGramJaccard\]" dispatch_qgram_jaccard.go && grep -q "QGramJaccardScore" llms.txt && grep -q "ErrInvalidQGramSize" llms.txt && ! grep -v '^#' qgram_jaccard.go | grep -E "math\.(Pow|Log|Exp|FMA)"</automated>
  </verify>
  <done>
    All TestExtractQGrams*, TestQGramJaccard*, TestDispatch_QGramJaccardRegistered, TestGolden_QGramJaccard_Staging, and ExampleQGramJaccardScore tests pass. License headers green. NO init() in q_gram.go or qgram_jaccard.go. Two new error sentinels declared in errors.go with godoc. Staging golden file exists, alphabetically sorted, canonical-marshalled. llms.txt + llms-full.txt updated with the new symbols.
  </done>
</task>

<task type="auto" tdd="true">
  <name>Task 2: Property tests + benchmarks + fuzz harnesses</name>
  <files>props_test.go, qgram_jaccard_bench_test.go, qgram_jaccard_fuzz_test.go, testdata/fuzz/FuzzQGramJaccardScore/seed-001, testdata/fuzz/FuzzQGramJaccardScoreRunes/seed-001</files>
  <read_first>
    - qgram_jaccard.go (created in Task 1 — read to understand the surface)
    - q_gram.go (created in Task 1)
    - props_test.go (last ~300 lines — find the RatcliffObershelp property-test block from Phase 4; exact template for the QGramJaccard append; identify the byte-path AND rune-path sections)
    - .planning/phases/05-q-gram-algorithms/05-RESEARCH.md §4.1 (allocation budget), §4.2 (capacity hint), §4.3 (no stack-buffer fast path — heap dominates regardless)
    - .planning/phases/05-q-gram-algorithms/05-CONTEXT.md §5 (property test conventions — 5 invariants × 2 surfaces; Symmetric for Jaccard/Dice/Cosine)
    - .planning/phases/04-remaining-character-gestalt/04-03-ratcliff-obershelp-PLAN.md (analog: byte + rune path bench + fuzz harnesses)
    - testdata/fuzz/FuzzRatcliffObershelpScore/seed-001 (`go test fuzz v1` literal-corpus format)
  </read_first>
  <behavior>
    - TestProp_QGramJaccardScore_RangeBounds (byte + rune surfaces) — testing/quick over arbitrary string inputs + n ∈ [1, 5] → score always in [0.0, 1.0]
    - TestProp_QGramJaccardScore_Identity (byte + rune) — Score(x, x, n) == 1.0 for non-empty x and n ≥ 1
    - TestProp_QGramJaccardScore_Symmetric (byte + rune) — Score(a, b, n) == Score(b, a, n) (exact equality — set-Jaccard is symmetric without float arithmetic depending on order)
    - TestProp_QGramJaccardScore_NoNaN / _NoInf / _NoNegativeZero (byte + rune)
    - TestProp_QGramJaccardScore_DeterministicAcrossRuns — 1000 sequential calls on the same input produce byte-identical output (PITFALLS §14 closure)
    - BenchmarkQGramJaccardScore_ASCII_Short / _ASCII_Medium / _ASCII_Long + BenchmarkQGramJaccardScoreRunes_Unicode_Short — alloc-asserted with b.ReportAllocs() + var sink anti-DCE; document expected alloc count (≤ 4 per RESEARCH.md §4.1)
    - FuzzQGramJaccardScore: panic-free, NaN/Inf-free, score-in-[0,1] for arbitrary (a, b, n) with n constrained to [1, 8] via the fuzz body; same for FuzzQGramJaccardScoreRunes
  </behavior>
  <action>
    Step A — Extend props_test.go by appending a new sectioned block at end-of-file per the Phase 4 template:
    ```
    // ---------------------------------------------------------------------------
    // Q-Gram Jaccard property tests (plan 05-01)
    // ---------------------------------------------------------------------------
    ```
    Add property tests in this order (six byte-path + six rune-path = 12 standard invariants):
    - TestProp_QGramJaccardScore_RangeBounds (testing/quick generates `(a, b string, n int)`; coerce n via `n = (n % 5) + 1` to keep in [1, 5]; assert 0 ≤ got ≤ 1 — strict ≤ on both ends; also assert !math.IsNaN(got) and !math.IsInf(got, 0) inline for parsimony)
    - TestProp_QGramJaccardScore_Identity (non-empty x and n ≥ 1; Score(x, x, n) == 1.0 exactly — set-Jaccard identity is bit-exact)
    - TestProp_QGramJaccardScore_Symmetric (exact equality)
    - TestProp_QGramJaccardScore_NoNaN, _NoInf, _NoNegativeZero (separate dedicated tests — the inline assertions in RangeBounds are redundant but harmless; explicit dedicated tests document the invariant)
    - TestProp_QGramJaccardScoreRunes_RangeBounds / _Identity / _Symmetric / _NoNaN / _NoInf / _NoNegativeZero (rune-path mirror of the above)
    - TestProp_QGramJaccardScore_DeterministicAcrossRuns (1000 sequential calls on "AGCT"/"AGCTAGCT"/n=2; assert all 1000 results == result_0 bit-for-bit via `math.Float64bits` comparison)
    Default 100 quick.Check iterations; explicit `quick.Check` calls.

    Step B — Create qgram_jaccard_bench_test.go per the Phase 4 bench-file template. Apache-2.0 header. Three byte-path benches + one rune-path bench:
    - BenchmarkQGramJaccardScore_ASCII_Short ("AGCT"/"AGCTAGCT"/n=2 — RV-J1)
    - BenchmarkQGramJaccardScore_ASCII_Medium (~50-char realistic pair, e.g. randomised hex strings of length 50/n=3)
    - BenchmarkQGramJaccardScore_ASCII_Long (~200-char pair / n=3)
    - BenchmarkQGramJaccardScoreRunes_Unicode_Short ("café"/"cafe"/n=2 — RV-J5)
    Pattern: `b.ReportAllocs(); b.ResetTimer(); var sink float64; for i := 0; i < b.N; i++ { sink = fuzzymatch.QGramJaccardScore(a, b, n) }; if sink < 0 { b.Fatal(...) }`. Add a comment above each Benchmark documenting the expected alloc count from RESEARCH.md §4.1 (≤ 4 allocs/op — two extractQGrams maps + at most two ancillary allocations for the cap-hinted backing arrays).

    Step C — Create qgram_jaccard_fuzz_test.go. Two harnesses: FuzzQGramJaccardScore (byte path) and FuzzQGramJaccardScoreRunes (rune path). Each harness:
    - Programmatic f.Add(...) seeds covering: RV-J1 pair, identity, both-empty, one-empty, no-overlap, invalid UTF-8 ("\xff\xfe"), long input (≥ 200 chars), n=1, n=8.
    - Fuzz body: coerce `n` into [1, 8] via `n = ((n%8) + 8) % 8 + 1`; call the algorithm; assert `!math.IsNaN(got)`, `!math.IsInf(got, 0)`, `got >= 0 && got <= 1`. Use t.Fatalf for any violation. No panic catching — panics are real failures (the n<1 panic path is unreachable because of the coercion).

    Step D — Create testdata/fuzz/FuzzQGramJaccardScore/seed-001 and testdata/fuzz/FuzzQGramJaccardScoreRunes/seed-001 in `go test fuzz v1` literal format. Three lines: header + 2 `string(...)` lines for (a, b) + 1 `int(...)` line for n. Use RV-J1 pair (AGCT/AGCTAGCT/n=2) as the canonical seed for both harnesses (same seed format; the rune harness will treat the bytes as UTF-8 — RV-J1 is pure ASCII so byte/rune paths align on that input). Byte-stable format per Phase 3 IN-06 closure — no extra whitespace.
  </action>
  <verify>
    <automated>cd /Users/johnny/Development/fuzzymatch && go test -run 'TestProp_QGramJaccard' ./... && go test -bench=BenchmarkQGramJaccard -benchmem -benchtime=1x ./... && go test -fuzz=FuzzQGramJaccardScore -fuzztime=10s ./... && go test -fuzz=FuzzQGramJaccardScoreRunes -fuzztime=10s ./... && head -1 testdata/fuzz/FuzzQGramJaccardScore/seed-001 | grep -q "^go test fuzz v1$" && head -1 testdata/fuzz/FuzzQGramJaccardScoreRunes/seed-001 | grep -q "^go test fuzz v1$"</automated>
  </verify>
  <done>
    All TestProp_QGramJaccardScore_* (byte + rune surfaces, 12+ property tests) pass under quick.Check default 100 iterations. Bench file produces 4 benches with alloc count within the documented budget. Both fuzz harnesses pass a 10s smoke run without panic/NaN/Inf/out-of-range. On-disk seed files present with byte-stable format.
  </done>
</task>

<task type="auto" tdd="true">
  <name>Task 3: BDD feature + steps</name>
  <files>tests/bdd/features/qgram_jaccard.feature, tests/bdd/steps/algorithms_steps.go</files>
  <read_first>
    - tests/bdd/features/ratcliff_obershelp.feature (Phase 4 analog — exact template)
    - tests/bdd/steps/algorithms_steps.go (full file — find RatcliffObershelp step methods + ctx.Step registrations inside InitializeScenario; identify the existing approximately-step regex `(\d+\.?\d*)` that already accepts integer-form per Phase 4 IN-03 closure)
    - .planning/phases/05-q-gram-algorithms/05-RESEARCH.md §5.1 (qgram_jaccard.feature skeleton)
    - .planning/phases/05-q-gram-algorithms/05-CONTEXT.md §5 (one BDD feature file per algorithm)
  </read_first>
  <behavior>
    - godog runs and passes all QGramJaccard scenarios
    - All scenarios use the existing `(\d+\.?\d*)` score regex — IF the existing step grammar does not accept a `with n <n>` clause, NEW step regex registrations are needed for the n-parameterised steps (byte + rune variants)
    - At least one scenario exercises each of RV-J1, RV-J2, RV-J3, RV-J4 and the rune-path café pair
  </behavior>
  <action>
    Create tests/bdd/features/qgram_jaccard.feature per RESEARCH.md §5.1 skeleton. Header comment: `# Primary source: Ukkonen, E. (1992). "Approximate string-matching with q-grams and maximal matches." TCS 92(1):191-211. § 3 worked example.` Scenarios:
    - `Feature: Q-Gram Jaccard similarity` with one-paragraph description
    - `Scenario Outline: Canonical reference vectors` with Examples table covering RV-J1 (AGCT/AGCTAGCT/n=2/0.4286), RV-J2 (hello/hello/n=2/1.0000), RV-J3 (abc/xyz/n=2/0.0000), RV-J4 (abcd/abxy/n=2/0.2000)
    - `Scenario: identical strings score 1.0` ("user_id"/"user_id"/n=2)
    - `Scenario: both-empty strings score 1.0` (""/""/n=2)
    - `Scenario: one-empty string scores 0.0` ("abc"/""/n=2 AND ""/"abc"/n=2)
    - `Scenario: score is symmetric` (compute "AGCT"/"AGCTAGCT"/n=2 and the swap; assert equality via the existing two-score-equality step or pin both to the same value)
    - `Scenario: rune-path Unicode pair` ("café"/"cafe"/n=2 → 0.5000) using the QGramJaccardRunes step
    Use tolerance 0.0001 in the approximately step (the existing 4-decimal convention).

    Extend tests/bdd/steps/algorithms_steps.go by appending the QGramJaccard step-method block. Required new methods on AlgorithmContext:
    - iComputeTheQGramJaccardScoreBetweenWithN(a, b string, n int) error
    - iComputeTheQGramJaccardRunesScoreBetweenWithN(a, b string, n int) error
    - iComputeTheSecondQGramJaccardScoreBetweenWithN(a, b string, n int) error
    - bothQGramJaccardScoresShouldBeEqual() error
    Register their regexes inside InitializeScenario:
    - `^I compute the QGramJaccard score between "([^"]*)" and "([^"]*)" with n (\d+)$`
    - `^I compute the QGramJaccardRunes score between "([^"]*)" and "([^"]*)" with n (\d+)$`
    - `^I compute the second QGramJaccard score between "([^"]*)" and "([^"]*)" with n (\d+)$`
    - `^both QGramJaccard scores should be equal$`
    Do NOT alter the existing `(\d+\.?\d*)` approximately-step regex. testify IS permitted in this file (sub-module) but the existing Phase 4 pattern uses `fmt.Errorf` — stay consistent.
  </action>
  <verify>
    <automated>cd /Users/johnny/Development/fuzzymatch && make test-bdd 2>&1 | grep -i 'qgram_jaccard\|QGramJaccard' && (cd tests/bdd && go test -run 'TestFeatures' ./...)</automated>
  </verify>
  <done>
    `make test-bdd` exits 0 with the new QGramJaccard scenarios green. Feature file covers identity, both-empty, one-empty, RV-J1..RV-J4 reference vectors, symmetry, rune-path Unicode. No regex drift; existing approximately-step regex reused.
  </done>
</task>

</tasks>

<threat_model>
## Trust Boundaries

| Boundary | Description |
|----------|-------------|
| caller → QGramJaccardScore / QGramJaccardScoreRunes | Untrusted (a, b string, n int) input crosses the API surface; library is pure-function with no I/O, auth, or session state |

## STRIDE Threat Register (ASVS Level 1)

| Threat ID | Category | Component | Disposition | Mitigation Plan |
|-----------|----------|-----------|-------------|-----------------|
| T-fuzz-panic | D (Denial of Service via panic on malformed input) | QGramJaccardScore / QGramJaccardScoreRunes on invalid UTF-8 / lone-surrogate / extreme-length inputs / n=0 / n<0 | mitigate | Task 2 ships FuzzQGramJaccardScore + FuzzQGramJaccardScoreRunes with ≥ 60s harness budget and on-disk seed corpus covering `\xff\xfe`, identity, both-empty, one-empty, RV-J1, long-input, n=1, n=8. Fuzz body coerces n to [1, 8] so the documented n<1 panic path is unreachable in fuzz; the panic path is exercised by TestQGramJaccard_PanicsOnInvalidN (Task 1) with explicit defer-recover assertions on the message text |
| T-complexity-attack | D (Denial of Service via algorithmic complexity) | extractQGrams / extractQGramsRunes on pathological inputs (huge input + small n produces a large map; huge input + n=1 produces a tiny map but slow scan) | accept | Q-gram extraction is O(la) time + O(la) map allocations bounded by min(la, b-cap-hint). PERF-01 documents the worst-case budget; long-input benches in Task 2 establish the regression baseline. Pure-function library — caller controls input size; godoc on QGramJaccardScore notes the O(la+lb) complexity |
| T-float-determinism | T (Tampering of float reduction order across architectures) | QGramJaccardScore final division `float64(intersection)/float64(union)` | mitigate | Single division on integer-derived floats. Both `intersection` and `union` are int (map-length differences); `float64(int)` is exact for counts up to 2^53; the single division is IEEE-754 correctly rounded on all four CI platforms. Cross-platform CI matrix verifies byte-identical golden output via testdata/golden/_staging/qgram_jaccard.json merged in plan 05-05; TestProp_QGramJaccardScore_DeterministicAcrossRuns (Task 2) pins per-process determinism. Set-Jaccard has no reduction chain — the float-determinism risk is essentially zero |
| T-map-iteration-leak | T (Tampering — non-deterministic output via map iteration) | QGramJaccardScore intersection computation must NOT produce ordered output | mitigate | The intersection cardinality computation iterates the SMALLER multiset map and accumulates an integer count via `if _, ok := qb[k]; ok { intersection += min(qa[k], qb[k]) }`. The OUTPUT is the integer count — NOT an ordered slice. Map iteration order does not affect the count. DET-03 satisfied. Cross-platform golden file verification provides the secondary guard |
</threat_model>

<verification>
- `go build ./...` succeeds.
- `go test -run 'TestExtractQGrams|TestQGramJaccard|TestProp_QGramJaccard|TestDispatch_QGramJaccardRegistered|TestGolden_QGramJaccard_Staging|ExampleQGramJaccard' ./...` exits 0.
- `go test -bench=BenchmarkQGramJaccard -benchmem -benchtime=1x ./...` runs cleanly; alloc counts within the documented RESEARCH.md §4.1 budget (≤ 4 allocs/op).
- `go test -fuzz=FuzzQGramJaccardScore -fuzztime=60s ./...` and `go test -fuzz=FuzzQGramJaccardScoreRunes -fuzztime=60s ./...` complete without failure (10s smoke OK for the per-task gate; 60s for `/gsd-verify-work`).
- `make test-bdd` green; QGramJaccard scenarios visible in godog output.
- `bash scripts/verify-license-headers.sh` exits 0.
- `bash scripts/verify-no-runtime-deps.sh` exits 0.
- `! grep -q "^func init" q_gram.go && ! grep -q "^func init" qgram_jaccard.go` (no init()).
- `grep -q "Source: Ukkonen, E. (1992)" q_gram.go && grep -q "Source: Ukkonen, E. (1992)" qgram_jaccard.go`.
- `grep -q "ErrInvalidQGramSize" errors.go && grep -q "ErrInvalidTverskyParam" errors.go`.
- `grep -v '^#' qgram_jaccard.go | ! grep -E "math\.(Pow|Log|Exp|FMA)"` (DET-06 gate; only `+`, `-`, `*`, `/` permitted).
- `grep -q "QGramJaccardScore" llms.txt && grep -q "QGramJaccardScoreRunes" llms.txt && grep -q "ErrInvalidQGramSize" llms.txt && grep -q "ErrInvalidTverskyParam" llms.txt`.
- `make coverage-check` confirms q_gram.go ≥ 90% per-file coverage and qgram_jaccard.go ≥ 90%; 100% on the public QGramJaccardScore + QGramJaccardScoreRunes surface.
- `make check` exits 0.
</verification>

<success_criteria>
- All three tasks complete; all listed verification commands green.
- testdata/golden/_staging/qgram_jaccard.json exists and is canonical-marshalled (no manual edits).
- testdata/fuzz/FuzzQGramJaccardScore/seed-001 and testdata/fuzz/FuzzQGramJaccardScoreRunes/seed-001 exist in byte-stable `go test fuzz v1` literal format.
- Public surface added by this plan is exactly TWO new exported functions (QGramJaccardScore, QGramJaccardScoreRunes) plus TWO new error sentinels (ErrInvalidQGramSize, ErrInvalidTverskyParam).
- Shared `extractQGrams` / `extractQGramsRunes` helpers in q_gram.go are unexported — single source of truth for plans 05-02 (Sørensen-Dice), 05-03 (Cosine), 05-04 (Tversky).
- Dispatch slot wired (default n=3 wrapper); TestDispatch_UnregisteredSlotsAreNil registered map flipped for AlgoQGramJaccard.
- Plans 05-02, 05-03, 05-04 can begin without further blockers.
</success_criteria>

<output>
After completion, create `.planning/phases/05-q-gram-algorithms/05-01-qgram-foundation-jaccard-SUMMARY.md` per the GSD summary template. Note the OQ-2 RESOLUTION LOCKED ({date}) in the SUMMARY: "ErrInvalidQGramSize + ErrInvalidTverskyParam added to errors.go; CONTEXT.md `code_context` claim that they were pre-declared was inaccurate (verified by grep before implementation)."
</output>
