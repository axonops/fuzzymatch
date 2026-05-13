# Architecture Research — fuzzymatch

**Domain:** Pure-function Go string-similarity / fuzzy-matching library
**Researched:** 2026-05-13
**Confidence:** HIGH (most decisions are spec-locked in `docs/requirements.md`; open decisions are explicitly flagged below and gated by the `api-ergonomics-reviewer`)

---

## 1. Scope and Authority

This document covers system structure for `github.com/axonops/fuzzymatch`. It is the **architecture dimension** of project research. Sibling research files cover stack (`STACK.md`), features (`FEATURES.md`), and pitfalls (`PITFALLS.md`).

**Spec-locked vs. open:** the bulk of architecture is locked by `docs/requirements.md` (the authoritative spec). What this document adds:

1. A consolidated picture of the three layers and their boundaries
2. The build-order rationale for the 8-phase release plan in `docs/requirements.md` §19
3. A starting position on the open API-surface decisions, **flagged for `api-ergonomics-reviewer` final ruling**
4. Comparable Go libraries' patterns, where they confirm or diverge from our choices

The illustrative code in `docs/requirements.md` is **not** prescriptive on names, signatures, option shapes, or error names. `api-ergonomics-reviewer` and `user-guide-reviewer` have veto authority (Design Principle 13). This research document inherits that constraint: every code block below is shape-illustrative.

---

## 2. Standard Architecture

### 2.1 System Overview — three layers, strict downward dependency

```
┌─────────────────────────────────────────────────────────────────────┐
│  Consumer code  (axonops/audit issue #853, third-party Go services) │
└────────────────────┬────────────────────────────────────────────────┘
                     │ import
       ┌─────────────┴──────────────┬────────────────────────┐
       ↓                            ↓                        ↓
┌────────────────┐         ┌────────────────────┐   ┌──────────────────┐
│ Layer 1        │         │ Layer 2            │   │ Layer 3          │
│ Algorithm fns  │ ←──────│ Scorer             │←─│ scan sub-package │
│ 23 algorithms  │  use    │ Weighted composite │   │ Collection scan  │
│ Normalise      │         │ over Layer 1       │   │ over Layer 2     │
│ Tokenise       │         │ Immutable          │   │ Buckets +        │
│ AlgoID enum    │         │ Concurrent-safe    │   │ suppression      │
└────────────────┘         └────────────────────┘   └──────────────────┘
       ▲                            ▲                        ▲
       │                            │                        │
       └─ package fuzzymatch ───────┘                        │
                                                              │
                       package fuzzymatch/scan ───────────────┘

Dependency direction:  scan → Scorer → Algorithm fns
                       (downward only; never reversed)

Module boundaries:
  github.com/axonops/fuzzymatch         (root go.mod — zero non-stdlib require lines)
  └── /scan                              (sub-package, same module, no separate go.mod)
  github.com/axonops/fuzzymatch/tests/bdd (separate go.mod — godog + goleak + testify)
```

### 2.2 Component Responsibilities

| Component | Responsibility | Imports From | Imported By |
|-----------|---------------|--------------|-------------|
| **AlgoID enum** (`algoid.go`) | Stable typed identifier per algorithm. `String()` returns snake_case. `AlgoIDs()` returns full ordered list. | stdlib | Layer 1 dispatch, Scorer, scan (Warning keys, golden files) |
| **Normalise** (`normalise.go`) | Lowercase + separator strip; ASCII fast path. Pure function. | stdlib (`unicode`) | Scorer, Tokenise, consumer code |
| **Tokenise** (`tokenise.go`) | Camel/Pascal/snake/kebab split + lowercase. Returns `[]string`. | `normalise.go` | Scorer (token-based algorithms), scan (bucket keys) |
| **Algorithm fns** (`levenshtein.go`, `jaro.go`, … × 23) | One algorithm per file, byte + rune variants where meaningful. Each cites primary source in file header. | stdlib | Scorer, consumer code, scan (transitively via Scorer) |
| **Scorer** (`scorer.go`, `scorer_options.go`) | Composes any AlgoID subset with weights, threshold, normalisation. Immutable after `NewScorer`. Concurrent-safe. | Layer 1 functions, AlgoID | scan, consumer code |
| **Scorer errors** (`errors.go`) | Sentinel errors (`ErrEmptyScorer`, `ErrInvalidWeight`, …) | stdlib (`errors`) | Scorer, consumer code |
| **scan.Item / Config / Warning** (`scan/scan.go`) | Public types for scan API | `fuzzymatch.Scorer`, stdlib | Consumer code |
| **scan.Check** (`scan/scan.go`) | Compares pairs; emits `Warning` slice in deterministic sort order. | `scan/bucket.go`, `fuzzymatch.Scorer` | Consumer code |
| **scan token bucket** (`scan/bucket.go`) | Token → item-index map; reduces N² candidate space. | stdlib | `scan.Check` only |
| **scan errors** (`scan/errors.go`) | Sentinel errors (`ErrNilScorer`, `ErrInvalidItem`, `ErrInvalidConfig`) | stdlib | scan consumers |

**Strict layering rule:** the root package MUST NOT import `scan`. A consumer of Layer 1 or Layer 2 pays no compile-time or runtime cost for Layer 3. This is verified by `scripts/verify-no-runtime-deps.sh` and by an import-cycle / no-back-edge meta-test.

### 2.3 Module layout

```
github.com/axonops/fuzzymatch/
├── go.mod                     # MUST have zero non-stdlib require lines
├── go.sum                     # essentially empty (stdlib only)
├── *.go                       # root package (Layer 1 + Layer 2)
├── scan/                      # Layer 3 sub-package (same module)
│   ├── *.go
│   └── *_test.go              # uses stdlib testing only
├── tests/bdd/
│   ├── go.mod                 # SEPARATE module: godog, goleak, testify here
│   ├── go.sum
│   ├── features/*.feature
│   └── steps/*_test.go
├── testdata/
│   ├── fuzz/                  # native Go fuzz corpora
│   └── golden/                # cross-platform determinism fixtures (JSON)
└── docs/
    ├── requirements.md        # authoritative spec
    └── *.md
```

**Why the BDD module is separate:** if godog/goleak/testify lived in the root `go.mod`, consumers of `github.com/axonops/fuzzymatch` would transitively pull them on every `go mod tidy`, even though they are test-only. A separate `tests/bdd/go.mod` keeps the consumer-facing dependency graph stdlib-only. `mask` uses this exact pattern. Search confirms it's the standard Go technique for isolating test-only deps. See [Go modules reference](https://go.dev/ref/mod) and [the standard Go multi-module guidance](https://go.dev/doc/modules/managing-dependencies).

**Why no `tests/bdd/go.work`:** workspaces are great for local development but harmful in CI — CI must build each module against its declared `go.mod` to catch declaration drift. The repository ships without a `go.work`. Local contributors who want one create it ad hoc and `.gitignore` protects against accidental commit.

---

## 3. Recommended Project Structure

```
fuzzymatch/
├── algoid.go                  # AlgoID + AlgoIDs() + String()  (FOUNDATION)
├── normalise.go               # Normalise + NormalisationOptions  (FOUNDATION)
├── tokenise.go                # Tokenise                          (FOUNDATION)
├── errors.go                  # root sentinel errors
│
├── levenshtein.go             # Character-based algorithms (9 files)
├── damerau_osa.go
├── damerau_full.go
├── hamming.go
├── jaro.go
├── jaro_winkler.go
├── strcmp95.go
├── smith_waterman_gotoh.go
├── lcsstr.go
│
├── q_gram.go                  # Q-gram / N-gram (shared extraction + 4 algorithms)
│                              #   Jaccard, Sørensen-Dice, Cosine, Tversky
│
├── monge_elkan.go             # Token-based (5 algorithms, 2 files)
├── token_ratio.go             #   TokenSort, TokenSet, Partial, TokenJaccard
│
├── soundex.go                 # Phonetic (4 files)
├── double_metaphone.go
├── nysiis.go
├── mra.go
│
├── ratcliff_obershelp.go      # Gestalt (1 file)
│
├── scorer.go                  # Scorer + NewScorer + DefaultScorer
├── scorer_options.go          # WithAlgorithm, WithNormalisation, …
│
├── *_test.go                  # External / black-box tests (one per algorithm + scorer)
├── *_internal_test.go         # Internal tests (q-gram extraction, Strcmp95 table)
├── *_bench_test.go            # Benchmarks (per algorithm + scorer + normalise)
├── props_test.go              # testing/quick property tests
├── fuzz_test.go               # native Go fuzz harnesses
├── example_test.go            # godoc runnable examples
├── doc.go                     # package-level documentation
│
├── documentation_test.go      # Meta-test: docs examples compile
├── ai_friendly_test.go        # Meta-test: llms.txt sync
├── readme_shop_front_test.go  # Meta-test: README example output
├── makefile_targets_test.go   # Meta-test: Makefile targets documented
├── internal_coverage_test.go  # Meta-test: coverage floor enforcement
│
├── scan/
│   ├── doc.go
│   ├── scan.go                # Item, Config, Warning, WarningKind, Check
│   ├── bucket.go              # token-bucket optimisation
│   ├── errors.go              # ErrNilScorer, ErrInvalidItem, ErrInvalidConfig
│   ├── scan_test.go           # external/black-box
│   ├── scan_internal_test.go  # bucket-equivalent-to-naive correctness
│   ├── scan_bench_test.go     # 200/1000/10000-item benchmarks
│   ├── example_test.go        # godoc runnable examples
│   ├── props_test.go          # property tests
│   └── fuzz_test.go           # fuzz harness for Check
│
└── tests/bdd/                 # Separate Go module
    ├── go.mod
    ├── features/
    │   ├── algorithms.feature
    │   ├── scorer.feature
    │   ├── normalisation.feature
    │   ├── determinism.feature
    │   ├── scan.feature
    │   └── suppression.feature
    └── steps/
        └── steps_test.go      # godog step definitions
```

### Structure Rationale

- **Flat root package** matches `adrg/strutil`'s `metrics/` subpackage idea but **without** the subpackage — for fuzzymatch the algorithm count and tight grouping make one-file-per-algorithm at the root level read cleaner. `hbollon/go-edlib` has all algorithms at root level too and that pattern works at this scale (~25 files for the algorithm catalogue).
- **One file per algorithm** keeps the primary-source citation, formula docstring, constants, byte-variant and rune-variant all in one place. Reviewers (and the `algorithm-correctness-reviewer` agent) can verify the citation-to-code chain without cross-file hunting.
- **Shared `q_gram.go`** is a deliberate exception — Jaccard, Sørensen-Dice, Cosine, and Tversky all consume the same q-gram extraction. Splitting the extraction across 4 files would duplicate logic. The file holds the four algorithm functions plus the shared extraction. Spec §18 endorses this.
- **`scan/` as a sub-package, not a sibling module** — this is a sub-package import (`github.com/axonops/fuzzymatch/scan`) but a single `go.mod` at the root. A separate module for `scan` would force versioned coordination on every release, which is overhead for no benefit.
- **`tests/bdd/` as a separate module** — non-negotiable to keep consumer `go.sum` stdlib-only.

---

## 4. Architectural Patterns

### Pattern 1: Algorithms as loose functions, NOT an interface

**What:** Each algorithm is a top-level public function (e.g. `LevenshteinScore(a, b string) float64`). There is no `Algorithm` or `StringMetric` interface that all algorithms implement.

**When to use:** when interface dispatch cost is measurable on the hot path AND the set of algorithms is closed (no plug-in algorithms).

**Trade-offs:**
- Pros: zero interface allocations on hot paths; consumers calling one algorithm import nothing extra; matches `hbollon/go-edlib` and `jcoruiz/strsim` patterns; `go-coding-standards` mandates this explicitly ("Algorithm dispatch via typed `AlgoID` enum, not an `Algorithm` interface").
- Cons: the Scorer needs an internal dispatch function (`func(AlgoID) func(a, b string) float64`) — pays one function-pointer call per algorithm per `Score` call but no boxing.

**Comparable libraries:**
- `adrg/strutil` — has a `StringMetric` interface in the `metrics/` subpackage. This boxes; on hot paths it allocates. fuzzymatch deliberately diverges.
- `hbollon/go-edlib` — loose functions plus integer-constant algorithm IDs passed to a dispatcher (`StringsSimilarity(a, b, algorithm)`). Closer to our pattern.
- `jcoruiz/strsim` — loose functions, no interface.

**Example shape (illustrative — `api-ergonomics-reviewer` rules on names):**

```go
// Public, loose function — what most consumers call directly:
func LevenshteinScore(a, b string) float64

// Scorer uses an internal dispatch table keyed on AlgoID:
var scoreFns = [...]func(a, b string) float64{
    AlgoLevenshtein:           LevenshteinScore,
    AlgoDamerauLevenshteinOSA: DamerauLevenshteinOSAScore,
    // …
}
```

### Pattern 2: Scorer via functional options, immutable after construction

**What:** `NewScorer(opts ...ScorerOption) (*Scorer, error)`. Each `WithXxx` option is a function that mutates an internal `*scorerConfig` during construction; after `NewScorer` returns, the `*Scorer` is read-only.

**When to use:** when the constructor has 3+ optional parameters and consumers want progressive disclosure. This is Rob Pike's idiomatic Go option pattern.

**Trade-offs:**
- Pros: extensible (new options are additive, not breaking); discoverable (IDE autocomplete on `fuzzymatch.With…`); spec §8.2 documents the full option list.
- Cons: errors deferred to `NewScorer` (not to each option call) — the option signature `ScorerOption func(*scorerConfig) error` carries the error up to construction.

**Comparable libraries:** `adrg/strutil` uses option structs per algorithm (`metrics.Levenshtein{CaseSensitive: false, ...}`). `hbollon/go-edlib` uses direct function parameters. Neither composes algorithms; fuzzymatch's Scorer is unique in this catalogue.

**Example shape (illustrative):**

```go
scorer, err := fuzzymatch.NewScorer(
    fuzzymatch.WithAlgorithm(fuzzymatch.AlgoLevenshtein, 0.30),
    fuzzymatch.WithAlgorithm(fuzzymatch.AlgoJaroWinkler, 0.20),
    fuzzymatch.WithThreshold(0.85),
    fuzzymatch.WithNormalisation(fuzzymatch.DefaultNormalisationOptions()),
)
```

**Open API decision flagged for `api-ergonomics-reviewer`:** should parameterised algorithms (Q-Gram, Tversky, Monge-Elkan, Cosine) each get a dedicated option, or should they share a generic shape? Spec §8.2 illustrates per-algorithm options. The dedicated-option pattern is more discoverable; the generic-option pattern is more uniform. **My recommendation:** stick with the per-algorithm options the spec illustrates — they're more discoverable, and the names like `WithQGramJaccardAlgorithm(w, n)` self-document the `n` parameter.

### Pattern 3: AlgoID as a typed int enum with stable String()

**What:** `type AlgoID int` with `const AlgoLevenshtein AlgoID = iota + 1`. `String()` returns the snake_case name. `AlgoIDs()` returns every defined ID in stable order.

**When to use:** when you need a stable, sortable, hashable identifier that survives serialisation and is cheap to compare.

**Trade-offs:**
- Pros: zero allocation in dispatch; AlgoID's natural integer ordering gives deterministic Scorer-internal iteration (no map iteration in output paths); `String()` gives consumer-friendly keys for `ScoreAll` map and JSON golden files.
- Cons: adding an algorithm without breaking ordinal compatibility requires appending (never reordering) — spec §13.4 makes ordinal stability explicit.

**Note:** `iota + 1` starts at 1 so the zero value of `AlgoID` is invalid — `WithAlgorithm(AlgoID(0), 0.5)` returns `ErrInvalidAlgoID` rather than silently using the first algorithm.

### Pattern 4: Two-row DP allocation strategy (per-algorithm)

**What:** Edit-distance DP algorithms (Levenshtein, Damerau-Levenshtein) keep only the previous row + current row, swapping pointers. ASCII fast path with stack-allocated `[64]byte` and `[65]int` buffers for inputs ≤ 64 bytes; heap fallback for longer inputs.

**When to use:** ASCII strings ≤ 50 chars (audit field names, identifiers, schema columns). Stack-allocation depends on Go's escape analysis keeping the array on the stack — verified by `b.ReportAllocs() == 0` in benchmarks.

**Trade-offs:** the spec's < 1 µs / 0-allocs budget for Levenshtein is ONLY achievable with this pattern. Naive DP table allocation fails the budget.

### Pattern 5: Token-bucket optimisation in scan (Layer 3 only)

**What:** Inside `scan.Check`, build `map[token][]itemIndex` once at the top, then for each item the candidate partner set is the union of bucket lookups by that item's tokens. Items sharing zero tokens after normalisation skip the expensive Scorer entirely.

**When to use:** collection scans where naive O(N²) would burn the budget. For 10000 items, naive is 50M pair comparisons; the bucket reduces this by 10–100× depending on token distribution.

**Trade-offs:**
- Pros: spec §12.6 budget (< 2 s for 10000 items) requires this.
- Cons: only correct because of the empirical observation "pairs with similarity > ~0.5 share at least one token after normalisation". Verified by property test `PropCheck_BucketEquivalentToNaive`.

**Critical determinism rule:** the bucket is `map[string][]int` — its iteration order is non-deterministic in Go. The `Check` function must NEVER iterate the bucket in a way that affects output order. The map is used only for membership lookups; the iteration that builds the output is over the sorted item list. The output `[]Warning` is sorted explicitly by `(Kind, NameA, NameB, GroupA, GroupB)` before return.

### Pattern 6: Pure-function purity, top to bottom

**What:** No goroutines. No channels. No I/O. No global mutable state. No `init()` doing non-trivial work. No mutexes. The Scorer is immutable after `NewScorer`. `scan.Check` is a pure function.

**When to use:** when the library is library-shaped (called from consumer code that owns concurrency, lifecycle, and I/O).

**Trade-offs:**
- Pros: trivial thread-safety story — "Scorer is safe for concurrent use by any number of goroutines"; trivial reasoning about determinism; no goroutine leaks possible (verified by `goleak.VerifyTestMain` in `tests/bdd/`).
- Cons: callers handle parallelism themselves. For embarrassingly-parallel collection scans, a consumer can shard the input and merge `[]Warning`. This is documented in `docs/scan.md`.

### Anti-Pattern A: An `Algorithm` interface that every algorithm implements

**What people do:** Define `type Metric interface { Similarity(a, b string) float64; Name() string }` and have every algorithm implement it. Then `Scorer` holds `[]Metric`.

**Why it's wrong for fuzzymatch:**
- Boxing on hot paths — calling through an interface allocates and indirects.
- The function-pointer dispatch via `AlgoID` is faster and doesn't box.
- Spec design principle 11 ("Errors only where genuinely necessary") and `go-coding-standards` ("Algorithm dispatch via typed `AlgoID` enum, not an `Algorithm` interface") both rule this out.

**Do this instead:** loose functions + `AlgoID` integer enum + internal dispatch table. See Pattern 1.

### Anti-Pattern B: Map iteration shaping output order

**What people do:** Build a `map[string]Warning` inside `Check`, then `for _, w := range m { out = append(out, w) }`. Go randomises map iteration; the output order changes per run.

**Why it's wrong:** breaks the determinism guarantee. Consumers building suppression lists keyed on observed scores or pair order silently break across runs. Spec §13.4 and `determinism-standards/SKILL.md` forbid this.

**Do this instead:** use the map only for membership/lookup. Build the output by iterating the sorted item list, append `Warning`s, then `sort.SliceStable` on the complete output key `(Kind, NameA, NameB, GroupA, GroupB)` before return.

### Anti-Pattern C: Floating-point operations that vary across architectures

**What people do:** Use `math.Pow(x, 2)` instead of `x * x`. Use `math.FMA` for fused multiply-add. Use parallel reductions in cosine similarity.

**Why it's wrong:** spec §13.3 promises byte-identical Scorer output on linux/darwin/windows × amd64/arm64. Transcendentals and platform-specific intrinsics break this. Parallel reductions sum in non-deterministic order.

**Do this instead:** simple `+`, `-`, `*`, `/` only; `math.Sqrt` is IEEE-754 correctly-rounded everywhere so it's safe for cosine similarity (`dot / (math.Sqrt(normA) * math.Sqrt(normB))`); guard division by zero explicitly (no NaN, no Inf). See `determinism-standards/SKILL.md`.

### Anti-Pattern D: A monolithic `Compare(a, b, algoName string) float64`

**What people do:** Single entry point keyed on string algorithm name.

**Why it's wrong:** string-typed dispatch is brittle (typos hit at runtime), allocates per call (string interning isn't free in Go), and obscures what's available at compile time.

**Do this instead:** loose typed functions + typed `AlgoID` for dispatch.

### Anti-Pattern E: An `init()` building algorithm tables

**What people do:** Use `init()` to populate Strcmp95's similar-character table or Double Metaphone rule tries.

**Why it's wrong:** init ordering across files is platform-stable in Go, but the temptation to do I/O or environment reads in `init()` grows. Spec §13.5 and `determinism-standards/SKILL.md` forbid non-trivial `init()`.

**Do this instead:** declare tables as `var x = ...` literal expressions or as functions invoked from `var` initialisation. Strcmp95's table is a `var` declaration.

---

## 5. Data Flow

### 5.1 Single algorithm call (Layer 1)

```
consumer code
    │
    │ fuzzymatch.LevenshteinScore("user_id", "userid")
    ↓
levenshtein.go: LevenshteinScore
    │ (no normalisation — raw input)
    │ ASCII fast path? yes → stack-allocated [65]int
    │ two-row DP loop
    ↓
return float64 in [0, 1]
```

No allocations, no I/O, no goroutines. < 1 µs budget per spec §14.1.

### 5.2 Scorer.Score call (Layer 2)

```
consumer code
    │
    │ scorer.Score("user_id", "userId")
    ↓
scorer.go: (*Scorer).Score
    │
    │ 1. Normalise(a, scorer.normOpts)  → "userid"
    │ 2. Normalise(b, scorer.normOpts)  → "userid"
    │ 3. For each (AlgoID, weight) in scorer.algos (sorted by AlgoID):
    │       compute score via dispatch table[AlgoID](a', b')
    │       accumulate weight * score
    │ 4. Return accumulated composite
    ↓
return float64 in [0, 1]
```

All steps deterministic. Iteration over `scorer.algos` is over a sorted slice, never a map. Spec §14.2 budgets the default Scorer at < 30 µs / ≤ 8 allocations.

### 5.3 Scorer.Match call

```
score := scorer.Score(a, b)
return score >= scorer.threshold
```

Trivial wrapper. Same cost as `Score`.

### 5.4 Scorer.ScoreAll call

```
result := make(map[string]float64, len(scorer.algos))
for each (algo, weight) in scorer.algos:
    result[algo.String()] = dispatch[algo](a', b')  // raw, unweighted
return result
```

`+1 allocation` for the result map (spec §14.2). Map iteration order is non-deterministic — but the map *values* are deterministic. Consumers wanting stable ordering call `sort.Strings(keys)` themselves.

### 5.5 scan.Check call (Layer 3)

```
consumer code
    │
    │ scan.Check(items, Config{Scorer: …, CompareAcrossGroups: true})
    ↓
scan/scan.go: Check
    │
    │ 1. Validate cfg → ErrNilScorer / ErrInvalidConfig if bad
    │ 2. Validate items → ErrInvalidItem (wrap index) if any Name == ""
    │ 3. Normalise SuppressedPairs entries once (using scorer.normOpts)
    │
    │ 4. Tokenise every item.Name once → tokens[i]
    │ 5. Build bucket: map[token][]int from tokens[i]
    │
    │ 6. Within-group pass:
    │      for each group g:
    │        candidates = bucket-lookup within g
    │        for each (i, j) in candidates with i < j:
    │          if suppressed → skip
    │          score, scores := scorer.Score(items[i].Name, items[j].Name),
    │                           scorer.ScoreAll(...)
    │          if score >= scorer.Threshold():
    │             append Warning{Kind: KindWithinGroup, …}
    │
    │ 7. Cross-group pass (if cfg.CompareAcrossGroups):
    │      effThreshold = min(scorer.Threshold() + cfg.CrossGroupThresholdBoost, 1.0)
    │      for each pair (i, j) across different groups:
    │        if suppressed → skip
    │        if !cfg.CompareIdenticalAcrossGroups && normalised(i) == normalised(j) → skip
    │        if score >= effThreshold:
    │           append Warning{Kind: KindAcrossGroups, …}
    │
    │ 8. sort.SliceStable(warnings, key=(Kind, NameA, NameB, GroupA, GroupB))
    ↓
return []Warning, nil
```

`Check` is a pure function — no goroutines, no I/O, no globals. Repeated calls return byte-identical output (verified by `PropCheck_DeterministicAcrossRuns`).

### 5.6 Lifecycle of a similarity warning

```
Item{Name: "user_id", Group: "audit_events"}
Item{Name: "userId",  Group: "audit_events"}
    │
    ↓ tokenise → ["user", "id"], ["user", "id"]
    ↓ bucket["user"] = [0, 1]; bucket["id"] = [0, 1]
    │
    ↓ candidate pair (0, 1) — same group, same tokens
    ↓ Scorer.Score(normalised, normalised)  → 0.96
    ↓ 0.96 ≥ 0.85 (threshold) → emit
    │
Warning{
    Kind: KindWithinGroup,
    NameA: "user_id", NameB: "userId",         // lex-sorted post-normalisation
    GroupA: "audit_events", GroupB: "audit_events",
    Score: 0.96,
    Scores: map[string]float64{ "levenshtein": …, "jaro_winkler": …, … },
}
```

### 5.7 Lifecycle of a suppression

```
Item{Name: "request_id",  Group: "http",  SilenceLint: false}
Item{Name: "response_id", Group: "http",  SilenceLint: false}
Item{Name: "legacy_id",   Group: "http",  SilenceLint: true}    ← flag

Config{
    Scorer: …,
    SuppressedPairs: [["request_id", "response_id"]],
}

→ pair (request_id, response_id) → suppressed by SuppressedPairs
→ pair (request_id, legacy_id)   → suppressed by SilenceLint=true on either side
→ pair (response_id, legacy_id)  → suppressed by SilenceLint=true on either side

Output: no warnings.
```

Suppression composes: a warning emits only if it survives all three filters (per-item SilenceLint, pair list, cross-group identical-name rule).

---

## 6. Build Order (foundational → derived)

The release phasing in `docs/requirements.md` §19 is the default. The build-order rationale below explains *why* — and flags one architecturally significant deviation: the Scorer can be built much earlier than Phase 5 if we want a working integration target sooner.

### 6.1 Foundation tier (Phase 1 — pre-release scaffolding)

```
algoid.go        ← independent foundation (typed enum, no dependencies)
normalise.go     ← depends on stdlib unicode only
tokenise.go      ← depends on normalise.go
errors.go        ← independent
```

These four files have no dependencies on any algorithm and no algorithm has any dependency on a Scorer. They are the first thing to land. Without them, no algorithm can be tested with normalised input, no Scorer can dispatch.

### 6.2 First-algorithm tier (Phase 2.0 — `v0.1.0`)

```
levenshtein.go            ← canonical reference: simplest non-trivial algorithm
damerau_osa.go            ← extends Levenshtein with adjacent transposition
damerau_full.go           ← extends Damerau-OSA with the Lowrance-Wagner full variant
hamming.go                ← trivial; useful for testing the test infrastructure
jaro.go                   ← independent; sets the pattern for non-edit-distance algorithms
jaro_winkler.go           ← extends Jaro with prefix bonus
```

Six algorithms in `v0.1.0`. Why these six first:
1. **Levenshtein is the canonical DP example** — landing it correctly proves the testing infrastructure (unit + property + fuzz + bench + BDD), the citation discipline, and the determinism workflow.
2. **Damerau-Levenshtein OSA + Full pair** lets us shake out the "two variants, slightly different outputs" pattern; spec §7.1.2 / §7.1.3 illustrate the divergence on `"ca"`/`"abc"`.
3. **Jaro + Jaro-Winkler** introduces the second algorithm category (record-linkage / name-matching) and the formula-based (vs DP-based) shape.
4. **Hamming** is trivial and a useful early sanity check.

### 6.3 Second-algorithm tier (Phase 2.1 — `v0.2.0`)

```
strcmp95.go
smith_waterman_gotoh.go
lcsstr.go
ratcliff_obershelp.go
```

The remaining character-based / gestalt algorithms. Each is a self-contained file; none has a cross-algorithm dependency.

### 6.4 Q-gram and token tier (Phase 3 — `v0.3.0`)

```
q_gram.go               ← shared q-gram extraction + Jaccard, Sørensen-Dice, Cosine, Tversky
monge_elkan.go          ← composes an inner algorithm via AlgoID (depends on dispatch table OR Scorer)
token_ratio.go          ← TokenSort, TokenSet, Partial, TokenJaccard
```

**Architectural note on Monge-Elkan:** it composes an inner metric. Spec §8.2 allows the inner to be configured via `WithMongeElkanAlgorithm(weight, inner AlgoID)`. The standalone `MongeElkanScore` function MUST therefore know how to dispatch to other algorithms by AlgoID. This means the AlgoID-keyed dispatch table needs to exist before Monge-Elkan does — but that's already required for the Scorer (Phase 5). **Open architectural question for `api-ergonomics-reviewer`:** does the standalone `MongeElkanScore` take an `inner AlgoID` parameter, or is Monge-Elkan only callable via the Scorer? My recommendation: expose `MongeElkanScore(a, b string, inner AlgoID) float64` for symmetry with other public algorithm functions. Default-inner-when-zero (`AlgoJaroWinkler` per spec) lets `MongeElkanScore(a, b, 0)` work as a sensible shortcut.

### 6.5 Phonetic tier (Phase 4 — `v0.4.0`)

```
soundex.go              ← simplest; long-expired patent
double_metaphone.go     ← largest rule table; must be reimplemented from Philips 2000 fresh
nysiis.go
mra.go
```

Phonetic algorithms are intentionally last in the algorithm-only phases because:
1. Their score normalisation is degenerate (0.0 or 1.0) — they exercise less of the float-determinism machinery.
2. Double Metaphone has the longest rule table (200+ rules) and is the highest licence-discipline risk — landing it last gives the algorithm-licensing-reviewer the most context.

### 6.6 Scorer tier (Phase 5 — `v0.5.0`)

```
scorer_options.go       ← functional options
scorer.go               ← Scorer + NewScorer + DefaultScorer + Score/ScoreAll/Match
```

By Phase 5, all 23 algorithms exist as public functions. The Scorer is the composition layer over them. It depends on `AlgoID`, normalisation, and the dispatch table that maps `AlgoID → score function`.

**Architectural deviation worth considering:** the Scorer could land earlier — even after `v0.1.0` — with a smaller initial algorithm set. This would let `axonops/audit` start integration testing against fuzzymatch sooner. The trade-off is that early integration finds API-surface issues before the algorithm catalogue is locked. My recommendation: **flag this for roadmap discussion**. The spec's "all algorithms first, then Scorer" approach is the default; bringing Scorer forward to `v0.2.0` or `v0.3.0` is an arguable improvement if integration feedback is valuable. The `api-ergonomics-reviewer` agent should rule on the trade-off.

### 6.7 Scan tier (Phase 6 — `v0.6.0`)

```
scan/scan.go            ← Item, Config, Warning, Check
scan/bucket.go          ← token-bucket optimisation
scan/errors.go          ← sentinel errors
```

`scan` depends on the Scorer (consumes `*fuzzymatch.Scorer`). It is the last library layer.

### 6.8 Integration shakedown (Phase 7 — `v0.6.x`)

`axonops/audit` issue #853 integrates fuzzymatch and scan end-to-end. API ergonomic issues surface. Patch releases as needed.

### 6.9 Stable (Phase 8 — `v1.0.0`)

API frozen. Algorithm score stability guaranteed across patch versions within v1. Cross-platform determinism golden files committed.

---

## 7. Cross-platform determinism architecture

This deserves its own subsection because it is a hard architectural constraint, not a quality nice-to-have.

### 7.1 What gets pinned

`testdata/golden/` contains JSON golden files:

| File | Content | What it pins |
|------|---------|--------------|
| `algorithms.json` | One entry per algorithm, fixed input corpus, expected float64 outputs at full precision | Algorithm score stability across patch versions AND across platforms |
| `scorer-default.json` | `DefaultScorer().Score` on the same corpus | Scorer composition stability + default-weight stability |
| `scan-default.json` | `scan.Check` on a fixed `Item` corpus with `DefaultScorer()` | Scan output ordering + warning composition stability |
| `normalisation.json` | `Normalise(s, DefaultNormalisationOptions())` for a corpus of inputs | Normalisation stability (this catches regressions in the ASCII fast path) |

JSON is the chosen format because (a) it's human-readable, (b) `diff -u` on JSON gives understandable conflict markers, (c) Go's encoding/json gives deterministic key ordering when the source is a struct with explicit fields. Float values are serialised with `%.17g` (or equivalent) to capture full IEEE-754 precision.

### 7.2 CI matrix

Spec §13.3 mandates the matrix:

```
                  amd64       arm64
   linux           ✓           ✓
   darwin          ✓           ✓
   windows         ✓           —     (no windows/arm64 — low ROI)
```

Every PR runs the determinism golden-file test on every platform in the matrix. Any platform producing a different file content fails CI.

### 7.3 What can break determinism

1. **Transcendental float ops** (`math.Log`, `math.Pow`, `math.Exp`) — IEEE-754 doesn't mandate exact rounding for these; platforms diverge. Avoid them. `math.Sqrt` is exact under IEEE-754 on all Go platforms, so cosine similarity is safe.
2. **Parallel reductions** — summing a slice via goroutines (or via SIMD on some platforms) produces non-deterministic order. Always sum left-to-right via a single goroutine.
3. **Map iteration leaking to output** — Go intentionally randomises map iteration. Any output path that consumes a map iteration order is broken. See Anti-Pattern B.
4. **Concurrent execution** — goroutine scheduling is non-deterministic. The library has zero goroutines, so this is constrained at the library boundary, but consumers parallelising the scan layer need to merge outputs deterministically.
5. **Sort instability** — `sort.Slice` is not stable. Use `sort.SliceStable` with a complete sort key (no ties).
6. **Time / environment / random** — none of the library reads time, env, or `crypto/rand`. Verified by `security-reviewer` and `algorithm-correctness-reviewer`.

### 7.4 Golden file regeneration policy

Golden files are regenerated only when:
- A bug is fixed that changes the algorithm's documented output (with primary-source justification)
- The score normalisation rule changes (with CHANGELOG entry + minor version bump)
- The default Scorer composition changes (with CHANGELOG entry + minor version bump)

Regeneration is a deliberate act with code review. Drive-by `go test -update-golden` runs are forbidden — `update-bench-txt.sh` and `update-golden.sh` are the only sanctioned paths and they ship CHANGELOG-entry checks.

### 7.5 Property tests as a complement

Spec §13.6 mandates `testing/quick` property tests. They complement golden files by exercising arbitrary input (not just the pinned corpus):

- `PropAlgorithm_DeterministicAcrossRuns` (per algorithm) — `score := f(a, b); for i := 0; i < 100; i++ { if f(a, b) != score { t.Fatal(...) } }` for fuzzed inputs
- `PropScorer_DeterministicAcrossRuns` — same for the Scorer
- `PropAlgorithm_NoNaN`, `PropAlgorithm_NoInf` — no float pathology on arbitrary input
- `PropAllPublic_NeverPanic` — no panic on any input including invalid UTF-8

Golden files pin specific inputs; property tests cover everything else.

---

## 8. Scaling Considerations

This library is a leaf dependency, not a system. "Scaling" means handling larger inputs and more concurrent callers, not horizontal sharding.

| Scale | Architecture Adjustments |
|-------|---------------------------|
| Pairwise comparison (Layer 1) | < 1 µs per algorithm call. ASCII fast path covers identifier-length strings (≤ 64 bytes) zero-alloc. Beyond that, two-row DP heap-allocates O(min(m,n)) ints. |
| Default Scorer composite | < 30 µs per Score call (6 algorithms × ~5 µs each). 8 allocations per call max (spec §14.2). |
| 200 items / 10 groups (scan) | < 10 ms total. Naive O(N²) is fine; bucket optimisation may be skipped at this scale (≤ 50 items per group). |
| 1000 items / 50 groups (scan) | < 100 ms. Bucket optimisation kicks in. |
| 10 000 items / 500 groups (scan) | < 2 s. Bucket optimisation is mandatory. Without it, 50M pair comparisons × 30 µs = 25 minutes. |
| Concurrent callers | Scorer is immutable after construction → any number of goroutines may call `Score`/`Match`/`ScoreAll` concurrently. No locks needed. |
| Beyond 10 000 items / multiple cores | **Consumer responsibility.** Shard `Items` into per-group slices, scan each in a goroutine, merge `[]Warning` deterministically (sort the merged output). Documented in `docs/scan.md`. The library itself never spawns a goroutine. |

### Scaling priorities

1. **First bottleneck for collection scan: pair-comparison count.** Token-bucket reduces candidate set from N² to N·k where k = items-sharing-at-least-one-token. Spec §12.5 mandates this.
2. **Second bottleneck: Scorer.Score per pair.** Default Scorer at 30 µs × 1M candidate pairs = 30 s. If the consumer needs faster, they can construct a Scorer with fewer algorithms (e.g. just Damerau-Levenshtein + Token Jaccard), trading recall for throughput.
3. **Third bottleneck: allocations under heavy parallel use.** Q-gram and token algorithms each allocate a small map per call. Under heavy parallelism this stresses the allocator. A future optimisation (post-v1) could pool these via `sync.Pool` — but this introduces complexity for marginal benefit and is explicitly out of v1 scope.

---

## 9. Integration Points

### 9.1 External services

None. fuzzymatch is a leaf library. No network, no filesystem, no database, no env vars, no time, no random.

### 9.2 Internal boundaries

| Boundary | Communication | Notes |
|----------|---------------|-------|
| Consumer ↔ Layer 1 algorithms | Direct function call | Zero allocation on ASCII fast path; no error returns |
| Consumer ↔ Scorer | `NewScorer` returns `(*Scorer, error)`; methods return values | Errors only on construction; Score/Match/ScoreAll error-free |
| Consumer ↔ scan.Check | `Check` returns `([]Warning, error)` | Errors on invalid config or items; output deterministic |
| Scorer → Layer 1 algorithms | Dispatch table `[AlgoID]func(a, b string) float64` | Internal; no consumer visibility |
| scan.Check → Scorer | Calls `scorer.Score()` and `scorer.Match()` | Public API only |
| scan.Check → Tokenise / Normalise | Calls public functions using scorer.normOpts | scan's internal pre-pass uses the same Tokenise as the Scorer for token-bucket consistency |
| Tests → BDD steps | `tests/bdd/` module imports root module by `replace` directive | The BDD module's `go.mod` will need a `replace github.com/axonops/fuzzymatch => ../..` for local development; CI strips the replace before publishing |

### 9.3 Public API surface (illustrative — `api-ergonomics-reviewer` rules on final shape)

**Layer 1 (algorithms):**
```
fuzzymatch.LevenshteinDistance(a, b string) int
fuzzymatch.LevenshteinDistanceRunes(a, b string) int
fuzzymatch.LevenshteinScore(a, b string) float64
fuzzymatch.LevenshteinScoreRunes(a, b string) float64
… ×23 algorithms × {Score, ScoreRunes, Distance, DistanceRunes, Code, Keys} as applicable

fuzzymatch.AlgoID                   (typed int enum, 23 values)
fuzzymatch.AlgoIDs() []AlgoID
fuzzymatch.AlgoID.String() string

fuzzymatch.Normalise(s string, opts NormalisationOptions) string
fuzzymatch.Tokenise(s string, opts NormalisationOptions) []string
fuzzymatch.NormalisationOptions     (struct with 4 fields)
fuzzymatch.DefaultNormalisationOptions() NormalisationOptions
```

**Layer 2 (Scorer):**
```
fuzzymatch.Scorer                   (opaque struct)
fuzzymatch.NewScorer(opts ...ScorerOption) (*Scorer, error)
fuzzymatch.DefaultScorer() *Scorer
fuzzymatch.ScorerOption             (type)
fuzzymatch.With…                    (8+ option functions, spec §8.2)
fuzzymatch.ScorerAlgorithm          (struct: ID, Weight)

(*Scorer).Score(a, b string) float64
(*Scorer).ScoreAll(a, b string) map[string]float64
(*Scorer).Match(a, b string) bool
(*Scorer).Threshold() float64
(*Scorer).Algorithms() []ScorerAlgorithm

fuzzymatch.ErrEmptyScorer, ErrInvalidWeight, ErrInvalidThreshold,
ErrInvalidAlgoID, ErrInvalidQGramSize, ErrInvalidTverskyParam   (sentinels)

fuzzymatch.Version() string
```

**Layer 3 (scan):**
```
scan.Item                           (struct: Name, Group, SilenceLint, Tag)
scan.Config                         (struct: Scorer, CompareAcrossGroups, …)
scan.Warning                        (struct: Kind, NameA, NameB, GroupA, GroupB, TagA, TagB, Score, Scores)
scan.WarningKind                    (typed int enum: KindWithinGroup, KindAcrossGroups)
scan.WarningKind.String() string

scan.Check(items []Item, cfg Config) ([]Warning, error)

scan.ErrNilScorer, ErrInvalidItem, ErrInvalidConfig             (sentinels)
```

### 9.4 Open architectural decisions flagged for `api-ergonomics-reviewer`

1. **`Algorithm` interface or not?** Spec and skills rule this out (no interface). My recommendation: confirmed no interface — use AlgoID dispatch table.
2. **Standalone Monge-Elkan signature.** Does `MongeElkanScore` take `inner AlgoID` directly? My recommendation: yes, with zero-value default. Open for `api-ergonomics-reviewer` ruling.
3. **Phonetic Score function semantics.** Spec §7.4 has them return 1.0/0.0. Should we also expose `*ScoreFuzzy` that takes Levenshtein-on-codes for partial-match scoring? Spec §11 explicitly says "consumers compose this themselves" — recommendation: keep it consumer composed; don't add `*ScoreFuzzy`.
4. **Scorer option for parameterised algorithms.** Spec §8.2 has dedicated options (`WithQGramJaccardAlgorithm`, etc.). My recommendation: keep dedicated, more discoverable than a generic `WithParameterisedAlgorithm`. `api-ergonomics-reviewer` may prefer a uniform shape.
5. **Bringing Scorer forward in the phase plan.** Default plan: Phase 5 (`v0.5.0`). My recommendation: flag for roadmap discussion — `axonops/audit` integration testing could start earlier with a partial-catalogue Scorer.
6. **Error returns from algorithm Score functions.** Spec design principle 11 forbids errors on Score; Hamming Distance returns an error for length mismatch, HammingScore returns 0.0. My recommendation: keep this asymmetry — it's explicit in the spec and `go-coding-standards` ratifies it.
7. **`Scorer.Algorithms()` return type.** Spec §8.3 illustrates `[]ScorerAlgorithm`. Allocation per call — acceptable for an introspection method (not on the hot path).
8. **`ScoreAll` map key type.** `map[string]float64` keyed on snake_case AlgoID name. My recommendation: confirmed — strings give consumers a stable serialisable key without re-import of AlgoID.

---

## 10. Comparable Go Library Patterns

| Library | Layout | Algorithm shape | Composition | Notes for fuzzymatch |
|---------|--------|-----------------|-------------|----------------------|
| `agnivade/levenshtein` | Single algorithm, single file | Loose function | N/A | Confirms zero-allocation Levenshtein is possible in pure Go |
| `xrash/smetrics` | Flat root package, ~7 files | Loose functions | N/A | Confirms flat root works for ~7 algorithms; we go bigger but use the same pattern |
| `adrg/strutil` | Root + `metrics/` subpackage | `StringMetric` interface | `Similarity(a, b, metric)` | We DIVERGE: no interface (avoids boxing); no subpackage (file-per-algo at root) |
| `hbollon/go-edlib` | Flat root | Loose functions + integer-constant dispatch | `StringsSimilarity(a, b, Algorithm)` | Closest to our dispatch pattern. We add a Scorer composition layer on top. |
| `jcoruiz/strsim` | Flat root | Loose functions, unified API | N/A | Confirms flat root, loose functions, zero-alloc rune handling |

**Key architectural lesson from this survey:** the existing Go ecosystem leans flat (loose functions at root). The `adrg/strutil` interface pattern is the outlier — it boxes on hot paths. fuzzymatch combines the best parts: loose functions like the majority pattern, **plus** a Scorer that composes them via AlgoID dispatch (no interface boxing) for the use case the existing libraries don't cover.

---

## 11. Spec-Locked vs. Open Decisions

### 11.1 Spec-locked (DO NOT relitigate without spec change)

- Three layers: algorithms / Scorer / scan
- Root module zero non-stdlib deps; BDD in separate `tests/bdd/go.mod`
- No cgo, no goroutines, no I/O, no globals in the library
- Scorer is immutable after construction
- AlgoID typed int enum (`go-coding-standards` confirms)
- `scan` is a sub-package of the same module, not a separate Go module
- Cross-platform determinism on linux/darwin/windows × amd64/arm64
- Map iteration never shapes output
- Sort key for scan output: `(Kind, NameA, NameB, GroupA, GroupB)`
- Token-bucket optimisation in scan with property test verifying equivalence to naive

### 11.2 Open (flagged for `api-ergonomics-reviewer`)

- Exact function names and signatures (Score vs Similarity, ScoreRunes vs RuneScore, etc.)
- Exact option function names (`WithAlgorithm` vs `Algorithm` vs `Algo`)
- Whether `MongeElkanScore` takes `inner AlgoID` as a public parameter
- Whether parameterised algorithm options are per-algorithm (`WithQGramJaccardAlgorithm`) or uniform (`WithParameterisedAlgorithm`)
- Whether Scorer ships in `v0.5.0` (default) or earlier (e.g. `v0.2.0` for earlier `audit` integration)
- Whether `DefaultScorer()` should be a top-level function or a `Scorer` constant — Go disallows non-trivial constants, so it has to be a function; but should it cache the result?

### 11.3 Open (flagged for roadmap discussion)

- Order within the algorithm phases. Spec §19 lumps "Strcmp95, SWG, LCSStr, Ratcliff-Obershelp" together in `v0.2.0`. Roadmapper may prefer to split (e.g. SWG and Ratcliff-Obershelp are individually substantial; pairing them with smaller algorithms may be wiser).
- Whether to ship a partial-catalogue early Scorer in `v0.2.0` / `v0.3.0` for `axonops/audit` integration testing
- Whether to add a `fuzzymatch/search` sub-package for query-vs-corpus search (out of v1 scope per spec §4; may surface as a v2 feature)

---

## 12. Sources

### Primary (HIGH confidence)
- [docs/requirements.md](file:///Users/johnny/Development/fuzzymatch/docs/requirements.md) — authoritative spec, §6 (Public API), §8 (Scorer), §9 (Normalisation), §10 (Tokenisation), §12 (Scan), §13 (Determinism), §18 (Repository Layout), §19 (Release Phasing)
- [docs/prior-art-research.md](file:///Users/johnny/Development/fuzzymatch/docs/prior-art-research.md) — Go ecosystem survey, architecture pattern analysis (`adrg/strutil` interface pattern, `hbollon/go-edlib` flat layout)
- [.claude/skills/go-coding-standards/SKILL.md](file:///Users/johnny/Development/fuzzymatch/.claude/skills/go-coding-standards/SKILL.md) — AlgoID enum vs interface decision, zero-dep root module, no testify in root, file size and grouping
- [.claude/skills/determinism-standards/SKILL.md](file:///Users/johnny/Development/fuzzymatch/.claude/skills/determinism-standards/SKILL.md) — golden file architecture, no-map-iteration rule, sort key completeness, float stability rules
- [.claude/skills/performance-standards/SKILL.md](file:///Users/johnny/Development/fuzzymatch/.claude/skills/performance-standards/SKILL.md) — two-row DP, ASCII fast path, stack-allocated buffer pattern
- [.claude/skills/algorithm-correctness-standards/SKILL.md](file:///Users/johnny/Development/fuzzymatch/.claude/skills/algorithm-correctness-standards/SKILL.md) — primary-source citation in file header, fresh-implementation discipline
- [.claude/skills/algorithm-licensing-standards/SKILL.md](file:///Users/johnny/Development/fuzzymatch/.claude/skills/algorithm-licensing-standards/SKILL.md) — Metaphone 3 precedent, fresh-implementation
- [.planning/PROJECT.md](file:///Users/johnny/Development/fuzzymatch/.planning/PROJECT.md) — project context, constraints, key decisions

### Comparable libraries surveyed (MEDIUM confidence — WebFetch on README/structure pages)
- [adrg/strutil](https://github.com/adrg/strutil) — `StringMetric` interface in `metrics/` subpackage; we deliberately diverge
- [hbollon/go-edlib](https://github.com/hbollon/go-edlib) — flat root, loose functions, integer-constant dispatch; closest comparable pattern
- [agnivade/levenshtein](https://github.com/agnivade/levenshtein) — confirms zero-allocation Levenshtein achievable
- [xrash/smetrics](https://github.com/xrash/smetrics) — flat root, ~7 algorithms, MIT
- [axonops/mask](https://github.com/axonops/mask) — structural template; single-module layout (mask uses testify in root tests — fuzzymatch is stricter)

### Go module patterns (MEDIUM confidence)
- [Go modules reference](https://go.dev/ref/mod) — multi-module repository layout
- [Managing dependencies (Go docs)](https://go.dev/doc/modules/managing-dependencies)
- [oneuptime.com — Multi-Module Go Projects with Workspaces](https://oneuptime.com/blog/post/2026-01-25-multi-module-go-projects-workspaces/view) — confirms separate `tests/bdd/go.mod` is idiomatic for isolating test-only deps; warns against committing `go.work` in CI

---

*Architecture research for: Go pure-function string-similarity library*
*Researched: 2026-05-13*
*Confidence: HIGH (spec-locked decisions); MEDIUM (API ergonomic specifics — gated by api-ergonomics-reviewer)*
