# fuzzymatch — Requirements

**Module:** `github.com/axonops/fuzzymatch`
**License:** Apache-2.0
**Go version:** 1.26+ minimum
**Status target:** Pre-release `v0.x` until first downstream consumer integration is exercised end-to-end; then `v1.0.0`.
**Document status:** Authoritative. This document is the single source of truth for the library's scope, public API, algorithm catalogue, testing requirements, performance budgets, and release acceptance criteria. Where any prior specification document, research note, or chat history disagrees with this document, this document wins.

---

> **Note on API examples.** Code blocks in this document — function signatures, struct definitions, option function names, sentinel error names, method names — are **illustrative**. They express intended shape, semantics, and feature set, not the final API. The `api-ergonomics-reviewer` and `user-guide-reviewer` agents (defined in `.claude/agents/`) hold final authority over actual API naming, signatures, option shapes, error handling patterns, and developer-experience details. Where this document and an agent's review disagree on API ergonomics, the agents win. This document is about **what** the library does and which algorithms it implements; the agents are about **how** the API surfaces that functionality to consumers. See Design Principle 13 in section 5.

---

## Table of Contents

1. Project Identity
2. Motivation
3. Goals
4. Out of Scope
5. Design Principles
6. Public API Overview
   - 6.A Error handling policy (data-vs-parameter framework)
7. Layer 1 — Algorithm Catalogue
   - 7.1 Character-based
   - 7.2 Q-gram / N-gram
   - 7.3 Token-based
   - 7.4 Phonetic
   - 7.5 Gestalt
8. Layer 2 — Scorer
9. Normalisation Pipeline
10. Tokenisation
11. Phonetic Algorithm Integration
    - 11.5 Input Validation and Diagnostics
12. Layer 3 — Scan sub-package
13. Determinism Guarantees
14. Performance Budgets
15. Testing Strategy
16. Documentation Requirements
17. CI/CD Requirements
18. Repository Layout
19. Release Phasing
20. Acceptance Criteria
21. References

---

## 1. Project Identity

`fuzzymatch` is a pure-Go, zero-dependency library for string similarity scoring. It exposes a comprehensive catalogue of similarity algorithms (Levenshtein, Damerau-Levenshtein, Jaro, Jaro-Winkler, Strcmp95, Smith-Waterman-Gotoh, LCSStr, Q-gram Jaccard, Sørensen-Dice, Cosine, Tversky, Monge-Elkan, Token Sort/Set/Partial Ratio, Token Jaccard, Soundex, Double Metaphone, NYSIIS, MRA, Ratcliff-Obershelp, Hamming) as independently-usable public functions, and provides a higher-level `Scorer` type that composes any subset of those algorithms into a weighted similarity score.

The library deliberately does NOT implement a collection-scan layer (e.g. "find all pairs in this slice that are similar"). That responsibility belongs to consumers, who can build it on top of the public algorithm functions and `Scorer` in a few lines.

- Module path: `github.com/axonops/fuzzymatch`
- Package name: `fuzzymatch` (matches the last path segment; no import alias required)
- License: Apache-2.0 with the standard AxonOps Apache-2.0 file header on every `.go` file
- Go version: 1.26+ (matches `axonops/audit` and `axonops/mask` toolchains)
- Runtime dependencies: stdlib only. Zero external `require` lines in `go.mod`
- No cgo. Anywhere. Ever.
- Test-only dependencies (`cucumber/godog`, `go.uber.org/goleak`, `testify`) live in `tests/bdd/go.mod` so consumers do not transitively depend on them
- Apache-2.0 NOTICE file attributes algorithm sources academically (no copied code)

The project look-and-feel mirrors [`axonops/mask`](https://github.com/axonops/mask): the same `.github/`, `docs/`, `scripts/`, `tests/bdd/` directory style, the same Makefile orchestration, the same README structure with logo / badges / emoji section headers / table of contents / status framing, the same `llms.txt` / `llms-full.txt` LLM-friendly documentation files, the same meta-test pattern (`documentation_test.go`, `ai_friendly_test.go`, `readme_shop_front_test.go`, `makefile_targets_test.go`, `internal_coverage_test.go`), the same CI/CD workflows, the same release engineering with goreleaser + sigstore keyless signing + OIDC build provenance. Mask is referenced as the structural and process template only. None of mask's API surface, masking concepts, or rule catalogue carry over.

---

## 2. Motivation

String similarity matching is a recurring need across AxonOps and Digitalis projects:

- Audit field/event taxonomy similarity warnings ("`user_id` vs `userid` vs `UserID` — pick one")
- Database schema column similarity
- API field-name consistency checks
- Kafka topic naming consistency
- Configuration vocabulary (YAML keys, environment variables, CLI flags)
- Code symbol naming (for linters)
- Glossaries and controlled vocabularies

Each project that needs this currently either rolls its own implementation or pulls a stale, unmaintained external dependency. The existing Go ecosystem for string similarity is fragmented: `agnivade/levenshtein` covers one algorithm well; `xrash/smetrics` covers a few; `adrg/strutil` exposes a clean interface but only seven algorithms; `hbollon/go-edlib` is the most comprehensive but its dependency surface, license model, and maintenance velocity vary.

The intent of `fuzzymatch` is to deliver one well-maintained, comprehensive, zero-dependency, Apache-2.0 Go library implementing the full set of practically-useful string similarity algorithms, with a clean Scorer composition layer so consumers can weight algorithms appropriately for their domain without re-implementing the underlying maths.

Comprehensiveness is the deliberate design choice. Different domains favour different algorithms (Damerau-Levenshtein for identifier typos, Jaro-Winkler for prefix-aligned name records, token-based metrics for snake_case-vs-camelCase, phonetic encoding for internationalised identifiers, n-gram for partial matches). Rather than picking three algorithms and forcing consumers to live with that choice, `fuzzymatch` ships all of them and lets the consumer compose.

---

## 3. Goals

1. **Implement the complete practical catalogue.** Twenty-three algorithms covering edit-distance, q-gram, token-based, phonetic, and gestalt pattern matching families. Every algorithm derived from a primary academic source (cited inline at the top of its implementation file). No GPL/LGPL-derived code. No copied implementations beyond what attribution explicitly permits.

2. **Expose each algorithm as an independently-callable public function.** A consumer that wants just Levenshtein calls `fuzzymatch.LevenshteinScore(a, b)` and never touches the Scorer machinery.

3. **Provide a composable Scorer layer.** Consumers select any subset of algorithms, weight them, set a match threshold, and receive a weighted composite score plus an optional per-algorithm breakdown for tuning and debugging.

4. **Normalisation as a first-class concern.** Pre-comparison normalisation (lowercase, separator stripping, camelCase splitting) is exposed both as standalone functions and as configuration on the Scorer.

5. **Determinism is a documented guarantee.** Algorithm scores are stable across patch versions. Scorer output for the same input and configuration is byte-identical across runs and platforms (linux/darwin/windows × amd64/arm64). No map iteration in output paths. Verified by property tests and a dedicated cross-platform CI matrix.

6. **Performance budgets locked in benchmarks.** Per-algorithm and per-Scorer-call budgets specified in section 14. Regressions over 10% versus the last tagged release fail CI via `benchstat`.

7. **Zero runtime dependencies, stdlib only, no cgo.** Verified structurally in CI via `scripts/verify-no-runtime-deps.sh`.

8. **Apache-2.0 throughout.** Every `.go` file carries the standard AxonOps Apache-2.0 file header. The NOTICE file attributes algorithm sources academically.

9. **Production-grade testing.** Unit tests per algorithm with literature reference vectors. Internal tests for unexported correctness invariants. Property-based tests via `testing/quick` for symmetry, range bounds, and triangle inequality (where applicable). Native Go fuzz tests for every public function. Allocation-budget benchmarks. BDD scenarios via `cucumber/godog` for the consumer-facing API. Meta-tests verifying README examples compile, `llms.txt` is in sync with the public API, and every documented `make` target exists.

10. **Pre-release `v0.x` framing.** Mirroring `axonops/mask`. The public API may break between minor versions until `v1.0.0`. `v1.0.0` is tagged only after the first downstream consumer (`axonops/audit` issue #853, scoped to build a collection-scan layer on top of `fuzzymatch`) has integrated end-to-end and exercised the public API in production-equivalent conditions.

---

## 4. Out of Scope

The following are explicit non-goals for v1. Any may be reconsidered for v1.x or v2 only after concrete demand and use cases surface.

- **Fuzzy-search-style "find similar to one query against a corpus".** Different problem class from collection-pairwise scanning. The provided `scan` sub-package (section 12) scans all pairs in a collection; a query-vs-corpus search optimised for one-to-many lookups (with indexing, prefix trees, BK-trees, or similar) is out of scope for v1. May become a future `fuzzymatch/search` sub-package if demand surfaces.
- **Needleman-Wunsch (global alignment).** Functionally redundant for short identifier-style strings given that Smith-Waterman-Gotoh (local alignment) handles the same use cases better with appropriate gap penalty settings.
- **Soft-TFIDF.** Requires a consumer-supplied corpus frequency table, which conflicts with the library's stateless pure-function design. Out of scope for v1.
- **Metaphone 3.** Covered by U.S. Patent 7440941. Even though the patent does not appear to be actively enforced and several MIT-licensed implementations exist, AxonOps declines to ship patent-encumbered algorithms regardless of enforcement posture. Consumers needing Metaphone 3 specifically should evaluate the existing third-party implementations on their own terms; Double Metaphone and NYSIIS in this library cover the majority of practical phonetic-encoding use cases without the patent overhang.
- **Embedding-based / semantic similarity.** Requires an external ML model, contradicting zero runtime dependencies.
- **Persistent state or caching across calls.** Every public function and Scorer method is pure: same input plus same configuration yields the same output, no globals, no I/O, no init-time state.
- **Configuration file parsing.** Consumers parse their own config and translate it into `Scorer` options.
- **I/O on the hot path.** No file reads, no network calls, no environment variable reads inside any algorithm or Scorer method.
- **A CLI tool.** The library can be wrapped in a CLI by consumers; not shipped here.
- **A web UI, an API server, or any non-library deliverable.**
- **Algorithm output stability across major versions.** Within a major version, algorithm scores are stable to the bit. A major version bump (e.g. v1 → v2) may change scoring, with the change documented in CHANGELOG and migration guide.

---

## 5. Design Principles

1. **Stdlib only on the runtime path.** No third-party runtime dependencies. The runtime `go.mod` has zero `require` lines beyond what the Go toolchain auto-inserts. Test-only dependencies live in `tests/bdd/go.mod`, structurally isolated.

2. **No cgo.** Anywhere.

3. **Pure-function core.** Every public function is pure: same input yields the same output, no globals, no I/O, no init-time state. The `Scorer` type is an immutable value: `NewScorer(opts...)` constructs a Scorer; once constructed it is read-only and safe for concurrent use by any number of goroutines.

4. **Deterministic output.** Algorithm scores are stable across patch versions and platforms. Scorer composite scores are stable across runs for the same input and configuration. Map iteration is never exposed to output paths. Property tests verify byte-identical output across repeated runs. A cross-platform CI matrix verifies the same byte-identical output on linux/darwin/windows × amd64/arm64.

5. **Composable.** Every algorithm is independently usable as a standalone public function. The Scorer composes algorithms into a weighted composite but does not gate access to the underlying algorithms. Consumers who want just one algorithm never instantiate a Scorer.

6. **Rune-aware UTF-8 where it matters.** Lowercasing and case-boundary detection use Unicode tables. Byte-level fast paths kick in for ASCII inputs. Public functions document whether they operate on bytes or runes; where both are meaningful (e.g. Levenshtein), both variants are exposed (e.g. `LevenshteinScore` byte-level, `LevenshteinScoreRunes` rune-level).

7. **Performance budgets locked in benchmarks.** Specified in section 14. `benchstat` regression detection over 10% versus the last tagged release fails CI.

8. **Algorithm output is stable across patch versions.** A documented promise. Score from any public `*Score` function in v1.0.0 must equal the same call in v1.5.0 to the last bit. Stability is verified by a golden-file test pinned in `testdata/golden/`. Any score change requires a minor version bump and a CHANGELOG entry.

9. **Apache-2.0 throughout.** Every `.go` file has the standard AxonOps Apache-2.0 file header. NOTICE attributes algorithm sources academically. No GPL/LGPL-derived code. No source-level copies from existing implementations (algorithms are reimplemented from primary academic sources, with the source cited in the implementation file).

10. **Minimal allocations on the hot path.** Per-algorithm allocation budgets in section 14. Stack-allocated buffers for short ASCII inputs. Heap-allocated slices only when input length exceeds the stack buffer size.

11. **Errors only where genuinely necessary.** Algorithm score functions never return errors — they handle every edge case (empty inputs, identical inputs, length mismatches for fixed-length-requiring algorithms like Hamming) by returning a defined score per the algorithm's documented edge-case behaviour. `Scorer` construction returns an error when the configuration is invalid (no algorithms, weights summing to zero, conflicting options); `Scorer` score methods are error-free.

12. **No init() functions with non-trivial work.** No package-level mutable state. No global registries.

13. **API ergonomics are determined by agent review, not by this document.** Code blocks throughout this document — type signatures, function names, option function shapes, sentinel error names, struct field layouts — illustrate intent. The `api-ergonomics-reviewer`, `user-guide-reviewer`, and `code-reviewer` agents in `.claude/agents/` hold final authority over the actual API surface. This document specifies *what* the library does (which algorithms, what semantics, what guarantees); the agents specify *how* it exposes that to consumers (naming conventions, signature shape, error handling style, idiomatic Go patterns, progressive disclosure). When implementation diverges from the illustrative code in this document but matches the agents' guidance, the agents win.

---

## 6. Public API Overview

> *Reminder: the code in this section is illustrative. The `api-ergonomics-reviewer` and `user-guide-reviewer` agents define the final naming, signature, and ergonomic shape. See Design Principle 13.*

The library exposes two layers, each independently usable.

### Layer 1 — Algorithm Functions

Each of the 23 algorithms is exposed as one or more public functions. Functions follow these naming conventions:

- `XxxScore(a, b string) float64` returns the algorithm's score in [0.0, 1.0]. Identical inputs return 1.0; both-empty returns 1.0 (documented by-convention except where mathematically undefined, in which case 0.0 is returned and documented).
- `XxxDistance(a, b string) int` returns the raw distance metric (for algorithms where a distance is meaningful, e.g. Levenshtein, Hamming).
- `XxxScoreRunes(a, b string) float64` is the rune-level variant where a byte-level fast path exists for ASCII.
- Phonetic algorithms expose `XxxCode(s string) string` (or `XxxKeys(s string) [2]string` for Double Metaphone), plus `XxxScore(a, b string) float64` returning 1.0 if codes match exactly and 0.0 otherwise.

All algorithm functions accept raw input. Applying the normalisation pipeline is the consumer's responsibility (or use the Scorer, which can apply normalisation automatically). The public function `Normalise(s string, opts NormalisationOptions) string` and `Tokenise(s string, opts NormalisationOptions) []string` are exposed for consumers building custom preprocessing pipelines.

### Layer 2 — Scorer

`Scorer` composes any subset of the 23 algorithms into a weighted similarity score.

```go
scorer, err := fuzzymatch.NewScorer(
    fuzzymatch.WithAlgorithm(fuzzymatch.AlgoLevenshtein, 0.30),
    fuzzymatch.WithAlgorithm(fuzzymatch.AlgoJaroWinkler, 0.25),
    fuzzymatch.WithAlgorithm(fuzzymatch.AlgoTokenJaccard, 0.20),
    fuzzymatch.WithAlgorithm(fuzzymatch.AlgoDamerauLevenshteinOSA, 0.15),
    fuzzymatch.WithAlgorithm(fuzzymatch.AlgoDoubleMetaphone, 0.10),
    fuzzymatch.WithNormalisation(fuzzymatch.DefaultNormalisationOptions()),
    fuzzymatch.WithThreshold(0.85),
)
if err != nil {
    return err
}

score := scorer.Score("user_id", "userId")            // 0.96 (example)
scores := scorer.ScoreAll("user_id", "userId")        // map per-algorithm
match := scorer.Match("user_id", "userId")            // true if score >= threshold
```

The Scorer is constructed via functional options. Options:

- `WithAlgorithm(algo AlgoID, weight float64)` adds an algorithm to the composite at the given raw weight. Algorithms with parameters (Q-Gram Jaccard, Tversky, Monge-Elkan, Cosine n-gram) have dedicated options (see section 8).
- `WithNormalisation(opts NormalisationOptions)` configures pre-comparison normalisation applied to both inputs before any algorithm is invoked.
- `WithThreshold(t float64)` sets the threshold used by `Match`. Default 0.85.
- `WithNormaliseWeights(false)` disables automatic weight normalisation. Default is `true` (weights are normalised to sum to 1.0 internally at Scorer construction).
- `WithQGramJaccardAlgorithm(weight float64, n int)` adds Q-Gram Jaccard with the given n. Default n = 3.
- `WithCosineAlgorithm(weight float64, n int)` adds Cosine n-gram with the given n.
- `WithTverskyAlgorithm(weight, alpha, beta float64)` adds Tversky with the given α, β. Symmetric variant (α = β = 0.5) reduces to Sørensen-Dice; (α = β = 1.0) reduces to Jaccard.
- `WithMongeElkanAlgorithm(weight float64, inner AlgoID)` adds Monge-Elkan using the given inner metric. Default inner is Jaro-Winkler per the original paper.

After construction, the Scorer is immutable. `Score`, `ScoreAll`, and `Match` are safe for concurrent use by any number of goroutines.

The Scorer applies normalisation (if configured) to both inputs once per `Score` call, then invokes each enabled algorithm on the normalised inputs, multiplies each algorithm's score by its (normalised) weight, sums, and returns the composite. `ScoreAll` returns the per-algorithm breakdown using the typed `AlgoID` enum as the map key (the original `map[string]float64` spec is SPEC-OVERRIDDEN to `map[AlgoID]float64`; see §8.3 + 08-CONTEXT.md §1). Consumers needing the CamelCase display form call `AlgoID.String()` (see Algorithm identifiers below for the locked naming convention).

### Sentinel errors

```go
var (
    ErrEmptyScorer             = errors.New("fuzzymatch: scorer has no algorithms configured")
    ErrInvalidWeight           = errors.New("fuzzymatch: invalid weight")
    ErrInvalidThreshold        = errors.New("fuzzymatch: invalid threshold")
    ErrInvalidAlgoID           = errors.New("fuzzymatch: invalid algorithm identifier")
    ErrInvalidQGramSize        = errors.New("fuzzymatch: invalid q-gram size")
    ErrInvalidTverskyParam     = errors.New("fuzzymatch: invalid tversky parameter")
    ErrInvalidInnerAlgo        = errors.New("fuzzymatch: invalid Monge-Elkan inner algorithm")
    ErrInternalInvariantViolated = errors.New("fuzzymatch: internal invariant violated (library bug — please file an issue)")
)
```

**Naming history (Phase 8.5 Gap 4 resolution).** The pre-8.5 code declared this sentinel as `ErrInvalidAlgorithm`; the spec always named it `ErrInvalidAlgoID`. Phase 8.5 renames the code to match the spec — a breaking change recorded in `CHANGELOG.md` "Breaking (pre-v1.0)". After the rename, every reference (`monge_elkan.go`, `llms-full.txt`, `scorer_options.go`, `scorer.go`, `scorer_options_test.go`, `errors_test.go`) uses `ErrInvalidAlgoID`. The non-existent-sentinel Critical finding from the Phase 8 review closes as a side-effect of the rename.

**Internal-invariant sentinel (Phase 8.5 Gap 5 resolution).** `ErrInternalInvariantViolated` is the typed panic value for "this should be impossible in correct usage; if you see it, it is a library bug, please file a GitHub issue." It replaces the bare `panic("fuzzymatch: DefaultScorer construction failed (this is a bug): " + err.Error())` at the previous `scorer.go:586` with `panic(fmt.Errorf("%w: DefaultScorer construction failed: %w", ErrInternalInvariantViolated, err))`. Consumers can `recover()` and call `errors.Is(panicValue.(error), ErrInternalInvariantViolated)` to discriminate library-internal panics from data / parameter panics. The sentinel is reserved for genuine internal invariants — it MUST NOT be used to wrap caller-supplied parameter errors, which use the typed parameter sentinels (`ErrInvalidAlgoID`, `ErrInvalidTverskyParam`, etc.).

All sentinel errors are exported from the root package, defined in `errors.go`. Error wrapping uses `fmt.Errorf("...: %w", err)` with `%w` as the final verb. Error discrimination uses `errors.Is` / `errors.As`.

### Algorithm identifiers

```go
type AlgoID int

const (
    AlgoLevenshtein AlgoID = iota + 1
    AlgoDamerauLevenshteinOSA
    AlgoDamerauLevenshteinFull
    AlgoHamming
    AlgoJaro
    AlgoJaroWinkler
    AlgoStrcmp95
    AlgoSmithWatermanGotoh
    AlgoLCSStr
    AlgoQGramJaccard
    AlgoSorensenDice
    AlgoCosine
    AlgoTversky
    AlgoMongeElkan
    AlgoTokenSortRatio
    AlgoTokenSetRatio
    AlgoPartialRatio
    AlgoTokenJaccard
    AlgoSoundex
    AlgoDoubleMetaphone
    AlgoNYSIIS
    AlgoMRA
    AlgoRatcliffObershelp
)

// String returns the CamelCase identifier matching the constant suffix
// ("Levenshtein", "JaroWinkler", "DamerauLevenshteinOSA", "NYSIIS", etc.).
func (a AlgoID) String() string

// AlgoIDs returns every defined algorithm identifier in stable order.
func AlgoIDs() []AlgoID
```

**Naming convention (LOCKED, v1.0):** `AlgoID.String()` returns the CamelCase form of the constant suffix (drop the `Algo` prefix). This matches Go's idiomatic enum-name → string convention (compare `time.Sunday.String()` returning `"Sunday"`, not `"sunday"`). The canonical strings are:

| AlgoID constant | `String()` return |
|---|---|
| `AlgoLevenshtein` | `"Levenshtein"` |
| `AlgoDamerauLevenshteinOSA` | `"DamerauLevenshteinOSA"` |
| `AlgoDamerauLevenshteinFull` | `"DamerauLevenshteinFull"` |
| `AlgoHamming` | `"Hamming"` |
| `AlgoJaro` | `"Jaro"` |
| `AlgoJaroWinkler` | `"JaroWinkler"` |
| `AlgoStrcmp95` | `"Strcmp95"` |
| `AlgoSmithWatermanGotoh` | `"SmithWatermanGotoh"` |
| `AlgoLCSStr` | `"LCSStr"` |
| `AlgoQGramJaccard` | `"QGramJaccard"` |
| `AlgoSorensenDice` | `"SorensenDice"` |
| `AlgoCosine` | `"Cosine"` |
| `AlgoTversky` | `"Tversky"` |
| `AlgoMongeElkan` | `"MongeElkan"` |
| `AlgoTokenSortRatio` | `"TokenSortRatio"` |
| `AlgoTokenSetRatio` | `"TokenSetRatio"` |
| `AlgoPartialRatio` | `"PartialRatio"` |
| `AlgoTokenJaccard` | `"TokenJaccard"` |
| `AlgoSoundex` | `"Soundex"` |
| `AlgoDoubleMetaphone` | `"DoubleMetaphone"` |
| `AlgoNYSIIS` | `"NYSIIS"` |
| `AlgoMRA` | `"MRA"` |
| `AlgoRatcliffObershelp` | `"RatcliffObershelp"` |

The same CamelCase convention applies to `WarnKind.String()` (see §11.5 Input Validation and Diagnostics) and any future enum-to-string conversion added to the public API. The `AlgoID.String()` mapping is stable across patch versions. The `AlgoIDs()` function returns the full list in stable order for use in tests, documentation generation, and consumer discovery.

### Version

```go
// Version returns the library's semantic version string.
func Version() string
```

### Layer 3 — Scan sub-package (optional)

`github.com/axonops/fuzzymatch/scan` provides a turnkey collection-scan layer for the common "iterate over a slice of names and emit a list of similar-name pairs" use case. It is a separate package and a separate import; consumers of just the algorithms or just the Scorer never depend on it. See section 12 for the full specification.

### 6.A Error handling policy

> The canonical Go declarations of these sentinels appear in the §6 code block above. The enumeration in this subsection documents semantic intent and error-handling guidance; if the two ever diverge, the §6 code block (the actual exported surface) wins.

The library distinguishes two classes of "bad input" with different policies. Reviewers and contributors apply the correct policy based on which class an input belongs to. This framework was locked during the Phase 8.5 review-findings triage (Q2).

#### Comparison-data inputs (the strings being compared)

Policy: **lenient.** Every algorithm score / distance function accepts any string input — empty, unequal lengths, weird Unicode, garbage bytes — and returns a sensible value (`0.0`, `max(len)`, etc.). It never panics; it never returns an error.

Rationale: the purpose of a similarity algorithm is to handle arbitrary strings. Callers should not have to pre-validate string shape before calling. The contract is "give me two strings; I return a number."

Examples of the lenient contract:

- `LevenshteinScore("", "hello")` returns a valid score
- `HammingDistance("ab", "abc")` returns `max(len)` — does not panic, does not error (see §7.1.4)
- `JaroWinklerScore(garbageBytes, garbageBytes)` returns a value, not a panic

Comparison-data inputs that are problematic-but-non-fatal (empty after normalisation, all non-ASCII dropped, pathologically large, etc.) are surfaced via the `Validate(a, b string) []Warning` function (see §11.5 Input Validation and Diagnostics) — warnings, not errors.

#### Parameter inputs (algorithm configuration)

Policy: **strict.** Parameters like `n` for q-gram size, `α`/`β` for Tversky, `inner AlgoID` for Monge-Elkan, and `threshold` for the Scorer have only finite valid values. Invalid values are programming errors and must fail loudly.

Validation surface depends on the entry point:

**Construction-time validation** — Scorer option functions return typed sentinel errors via the `ScorerOption func(*scorerConfig) error` signature:

| Option | Sentinel returned on invalid input |
|---|---|
| `WithQGramJaccardAlgorithm(weight, n)` (`n < 1`) | `ErrInvalidQGramSize` |
| `WithSorensenDiceAlgorithm(weight, n)` (`n < 1`) | `ErrInvalidQGramSize` |
| `WithCosineAlgorithm(weight, n)` (`n < 1`) | `ErrInvalidQGramSize` |
| `WithTverskyAlgorithm(weight, α, β, n)` (`n < 1`) | `ErrInvalidQGramSize` |
| `WithTverskyAlgorithm(weight, α, β, n)` (`α < 0`, `β < 0`, or `α + β = 0`) | `ErrInvalidTverskyParam` |
| `WithMongeElkanAlgorithm(weight, inner)` (inner is unknown / out-of-range AlgoID, or is a token-tier AlgoID, or is `AlgoMongeElkan` self-reference) | `ErrInvalidInnerAlgo` |
| `WithThreshold(t)` — see explicit check below | `ErrInvalidThreshold` |
| `WithAlgorithm(algo, weight)` (`weight ≤ 0`) | `ErrInvalidWeight` |
| `NewScorer(...)` with no algorithms | `ErrEmptyScorer` |

After successful construction the Scorer is immutable. `Score`, `ScoreAll`, and `Match` are lenient at runtime — all parameter validation has already happened.

**`WithThreshold` explicit check.** The threshold check is written as a single guard with `math.IsNaN` first so NaN does not slip past the range comparison (a NaN comparison against any finite value returns false, so a pre-Q2 `if t < 0.0 || t > 1.0` guard accepted NaN):

```go
if math.IsNaN(t) || t < 0.0 || t > 1.0 {
    return ErrInvalidThreshold
}
```

`±Inf` is rejected by the range check itself (`math.Inf(1) > 1.0` is true; `math.Inf(-1) < 0.0` is true) — no separate `±Inf` clause is needed. Phase 8.5 Q2 added only the `math.IsNaN(t)` clause; `±Inf` was already rejected before.

**Direct algorithm function calls** — direct calls to parameterised algorithm functions panic with the sentinel error as the panic value, so consumers can recover via:

```go
defer func() {
    if r := recover(); r != nil {
        if err, ok := r.(error); ok && errors.Is(err, fuzzymatch.ErrInvalidQGramSize) {
            // handle programming error
        }
    }
}()
score := fuzzymatch.QGramJaccardScore("a", "b", 0)  // panics with ErrInvalidQGramSize
```

The following direct calls panic on invalid parameter inputs:

- `QGramJaccardScore`, `QGramJaccardScoreRunes` — panic with `ErrInvalidQGramSize` if `n < 1`
- `SorensenDiceScore`, `SorensenDiceScoreRunes` — panic with `ErrInvalidQGramSize` if `n < 1`
- `CosineScore`, `CosineScoreRunes` — panic with `ErrInvalidQGramSize` if `n < 1`
- `TverskyScore`, `TverskyScoreRunes` — panic with `ErrInvalidQGramSize` if `n < 1`, or `ErrInvalidTverskyParam` if `α < 0`, `β < 0`, or `α + β ≤ 0`
- `MongeElkanScore`, `MongeElkanScoreAsymmetric` — panic with `ErrInvalidInnerAlgo` if `inner` is the zero value, an out-of-range `AlgoID`, `AlgoMongeElkan` itself (self-reference), or any token-tier `AlgoID` (`AlgoTokenSortRatio`, `AlgoTokenSetRatio`, `AlgoPartialRatio`, `AlgoTokenJaccard`). Permitted inner metrics are the character + q-gram + phonetic + gestalt tiers per the documented `permittedMongeElkanInner` allow-list.

Rationale: direct algorithm functions return just `float64` (no error tuple). Forcing `(float64, error)` returns on every direct call would be massive API churn for an error case that is "programming bug, not runtime data." Parameter errors are programmer errors (`n=0` for q-gram is never legitimate code). Silent-return-0 would hide the bug forever — wrong results would propagate downstream with no diagnostic. Panic-with-recoverable-sentinel catches the bug at first call AND gives advanced consumers a recovery path.

The Scorer construction path is the recommended entry point for any code that may receive parameters from untrusted configuration (e.g. CLI flags, environment variables, YAML). Direct algorithm calls are appropriate when parameters are hard-coded at the call site.

#### Currently-defined sentinel errors

The complete v1.0 sentinel set (declared in `errors.go`, all exported, all `fuzzymatch:`-prefixed):

- `ErrEmptyScorer` — `NewScorer` called with no algorithms configured
- `ErrInvalidWeight` — option function passed `weight ≤ 0` (or sum of weights `≤ 0`)
- `ErrInvalidThreshold` — `WithThreshold` passed value outside `[0.0, 1.0]` or `NaN`
- `ErrInvalidAlgoID` — option function passed an out-of-range `AlgoID` (renamed from `ErrInvalidAlgorithm` in Phase 8.5 — Gap 4 resolution)
- `ErrInvalidQGramSize` — q-gram option or direct call passed `n < 1`
- `ErrInvalidTverskyParam` — Tversky option or direct call passed `α < 0`, `β < 0`, or `α + β ≤ 0`
- `ErrInvalidInnerAlgo` — Monge-Elkan option or direct call passed an inner `AlgoID` that is unknown / out-of-range, the `AlgoMongeElkan` self-reference, or a token-tier `AlgoID`
- `ErrInternalInvariantViolated` — typed panic value raised when a library-internal invariant fails (e.g. `DefaultScorer()` construction failure on a dispatch-table gap that should be impossible). Never fires in correct usage; consumers seeing it should file a bug. Added in Phase 8.5 Gap 5 resolution.

Every sentinel carries the four-section godoc block (What / Common causes / Resolution / Example) per `.claude/skills/documentation-standards/SKILL.md` § Error sentinel documentation. The exemplar block below shows the locked template form for `ErrInvalidInnerAlgo`; every other sentinel follows the same structure:

```go
// ErrInvalidInnerAlgo is returned by WithMongeElkanAlgorithm and raised as
// a panic value by MongeElkanScore / MongeElkanScoreAsymmetric when the
// inner-metric AlgoID is invalid.
//
// What it means:
//   An AlgoID was passed as the inner metric for a Monge-Elkan call, but
//   the AlgoID is either (a) unknown / unregistered (zero value or out
//   of range), or (b) refers to a token-tier algorithm that cannot be
//   nested as an inner metric, or (c) is AlgoMongeElkan itself (self-
//   reference would cause infinite recursion).
//
// Common causes:
//   - Passing the zero value of AlgoID (uninitialised variable)
//   - Passing a token-tier AlgoID: AlgoTokenJaccard, AlgoTokenSortRatio,
//     AlgoTokenSetRatio, AlgoPartialRatio
//   - Passing AlgoMongeElkan itself (self-reference)
//   - Typo in AlgoID constant name producing an out-of-range value
//
// Resolution:
//   - Pass a character-tier AlgoID (Levenshtein, DamerauOSA, Hamming,
//     Jaro, JaroWinkler, etc.) as the inner metric. Q-gram and phonetic
//     AlgoIDs are also permitted; consult the algorithm catalogue for
//     the full set of valid inner AlgoIDs.
//
// Example:
//   // panics with ErrInvalidInnerAlgo on direct call:
//   _ = fuzzymatch.MongeElkanScore("a", "b", fuzzymatch.AlgoTokenJaccard)
//
//   // valid:
//   _ = fuzzymatch.MongeElkanScore("a", "b", fuzzymatch.AlgoJaroWinkler)
var ErrInvalidInnerAlgo = errors.New("fuzzymatch: invalid Monge-Elkan inner algorithm")
```

The locked four-section template for `ErrInternalInvariantViolated` (Gap 5 resolution) — same shape, different semantic class:

```go
// ErrInternalInvariantViolated is raised as the typed panic value when a
// library-internal invariant fails. It NEVER fires in correct usage.
// Consumers who observe a panic carrying this sentinel are observing a
// library bug, not a usage error — please file an issue at
// https://github.com/axonops/fuzzymatch/issues.
//
// What it means:
//   A library-internal assertion has failed. The library reached a state
//   that should be impossible given the type system, the option-time
//   validation, and the immutable-after-construction Scorer contract.
//   Examples of "should be impossible": DefaultScorer() construction
//   failure on a hard-coded option set; dispatch-table lookup miss on
//   an AlgoID that survived option-time validation; a Score path that
//   produces a NaN despite all parameter sentinels being clean.
//
//   This sentinel is the structured replacement for bare
//   `panic("library bug: ...")` strings. The error wrapping carries the
//   internal-cause error so debugging tools see the chain via
//   errors.Unwrap / errors.Is.
//
// Common causes:
//   - A library code change introduced a dispatch-table gap that the
//     compile-time tests failed to catch
//   - A platform-specific floating-point divergence escaped the
//     cross-platform CI matrix gate
//   - A future Go runtime change altered the semantics of a stdlib
//     primitive the library relies on
//
//   None of these are caller-fault. All are library-fault.
//
// Resolution:
//   - File a GitHub issue with the panic message, the calling code, the
//     Go version, and the platform. The library maintainers will fix
//     the root cause in the next patch release.
//   - As a temporary workaround, defer-recover around the affected call
//     and discriminate via errors.Is(panicValue.(error),
//     ErrInternalInvariantViolated) — but please file the issue so the
//     workaround can be removed.
//
//   ErrInternalInvariantViolated MUST NOT be used to wrap caller-supplied
//   parameter errors. Parameter errors use the dedicated parameter
//   sentinels (ErrInvalidAlgoID, ErrInvalidTverskyParam, etc.).
//
// Example:
//   // Library-internal panic (DefaultScorer() construction):
//   //   panics with: ErrInternalInvariantViolated: DefaultScorer
//   //   construction failed: <wrapped cause>
//
//   defer func() {
//       if r := recover(); r != nil {
//           if err, ok := r.(error); ok && errors.Is(err, fuzzymatch.ErrInternalInvariantViolated) {
//               // library bug — log and recover
//               log.Printf("fuzzymatch library bug: %v", err)
//           }
//       }
//   }()
var ErrInternalInvariantViolated = errors.New("fuzzymatch: internal invariant violated (library bug — please file an issue)")
```

---

## 7. Layer 1 — Algorithm Catalogue

> *Reminder: per-algorithm function signatures below are illustrative. The `api-ergonomics-reviewer` agent has final say on names, parameter ordering, byte-vs-rune variant exposure, and overall surface shape. Algorithm semantics, score normalisation rules, edge-case behaviour, and primary-source citations are NOT illustrative — those are requirements.*

This section specifies each algorithm in detail: primary source citation, formal description, complexity, score normalisation, mathematical invariants, edge cases, public function signatures, and intended use.

Every implementation must derive from the primary source cited. No copying from existing Go ports. No GPL/LGPL-derived code. Implementation may study existing MIT-licensed ports (e.g. `adrg/strutil`, `hbollon/go-edlib`, `xrash/smetrics`) for cross-validation of reference vectors, but the code itself must be written fresh.

Every algorithm implementation file (e.g. `levenshtein.go`) MUST begin (after the Apache-2.0 file header) with a block comment citing the primary source, naming the formula, and noting any deliberate deviation from the canonical formulation. Constants used in the algorithm (e.g. `winklerPrefixScale = 0.1`) are declared as unexported package consts and documented with a reference back to the originating paper.

### 7.1 Character-based algorithms

#### 7.1.1 Levenshtein

- **Category:** character-based, edit distance
- **Primary source:** Levenshtein, V. I. (1965). "Binary codes capable of correcting deletions, insertions, and reversals." *Soviet Physics Doklady*, 10(8):707–710.
- **AlgoID:** `AlgoLevenshtein`
- **Public functions:**
  - `LevenshteinDistance(a, b string) int` — raw edit distance (byte-level)
  - `LevenshteinDistanceRunes(a, b string) int` — raw edit distance (rune-level)
  - `LevenshteinScore(a, b string) float64` — normalised score in [0.0, 1.0] (byte-level)
  - `LevenshteinScoreRunes(a, b string) float64` — normalised score (rune-level)
- **Description:** the minimum number of single-character insertions, deletions, or substitutions required to transform `a` into `b`. Each edit costs 1.
- **Recurrence:** `D[i,j] = min(D[i-1,j]+1, D[i,j-1]+1, D[i-1,j-1] + (a[i-1] ≠ b[j-1]))` with `D[i,0] = i`, `D[0,j] = j`.
- **Complexity:** O(m·n) time, O(min(m,n)) space using the two-row optimisation.
- **Score normalisation:** `score = 1.0 - distance / max(len(a), len(b))`. Identical inputs (including both empty) return 1.0; one-empty returns 0.0.
- **Mathematical invariants:**
  - Identity: `Score(x, x) = 1.0`
  - Symmetry: `Score(a, b) = Score(b, a)`
  - Range: `Score(a, b) ∈ [0.0, 1.0]`
  - Triangle inequality (on the underlying distance): `Distance(a, c) ≤ Distance(a, b) + Distance(b, c)`
- **Edge cases:**
  - Both empty: distance 0, score 1.0
  - One empty: distance = `len(other)`, score 0.0
  - Identical: distance 0, score 1.0
- **Implementation notes:** ASCII fast path with stack-allocated `[64]byte` arrays for inputs under 64 bytes (zero allocations). Heap-allocated slices for longer inputs. The rune variant operates on Unicode code points and is slower; the byte variant is the default.
- **Reference vectors:** "kitten"/"sitting" → distance 3, score `1 - 3/7 ≈ 0.5714`. "saturday"/"sunday" → distance 3, score `1 - 3/8 = 0.625`. "" / "abc" → distance 3, score 0.0. "abc" / "abc" → distance 0, score 1.0.
- **Intended use:** primary edit-distance metric. Best general-purpose single algorithm for short identifier-style strings with single-character typos.

#### 7.1.2 Damerau-Levenshtein (OSA — Optimal String Alignment)

- **Category:** character-based, edit distance with transposition
- **Primary source:** Damerau, F. J. (1964). "A technique for computer detection and correction of spelling errors." *Communications of the ACM*, 7(3):171–176. The OSA variant (restricted edit distance with transpositions) is the formulation in common use; canonical reference: Boytsov, L. (2011). "Indexing methods for approximate dictionary searching: comparative analysis." *ACM Journal of Experimental Algorithmics*, 16, Article 1.
- **AlgoID:** `AlgoDamerauLevenshteinOSA`
- **Public functions:**
  - `DamerauLevenshteinOSADistance(a, b string) int`
  - `DamerauLevenshteinOSADistanceRunes(a, b string) int`
  - `DamerauLevenshteinOSAScore(a, b string) float64`
  - `DamerauLevenshteinOSAScoreRunes(a, b string) float64`
- **Description:** Levenshtein extended with adjacent character transposition as a single edit. The "Optimal String Alignment" restriction means each substring may participate in at most one transposition; substrings cannot be re-edited after a transposition. Faster than the full Damerau-Levenshtein; produces slightly different (sometimes larger) distances on inputs where a substring's already-transposed characters would otherwise be edited again.
- **Recurrence:** Levenshtein recurrence plus, when `i ≥ 2`, `j ≥ 2`, `a[i-1] = b[j-2]`, and `a[i-2] = b[j-1]`: `D[i,j] = min(D[i,j], D[i-2,j-2] + 1)`.
- **Complexity:** O(m·n) time, O(min(m,n) · 2) space.
- **Score normalisation:** identical to Levenshtein.
- **Mathematical invariants:** identity, symmetry, range bounds. Triangle inequality holds for the underlying distance.
- **Edge cases:** identical to Levenshtein.
- **Intended use:** primary choice for identifier typo detection — handles keyboard-adjacent character swaps (`creatd_at` vs `created_at`) at one edit instead of two.
- **Reference vectors:** "ab"/"ba" → distance 1 (one transposition), score `1 - 1/2 = 0.5`. "ca"/"abc" → OSA distance 3 (full DL would give 2; the OSA restriction shows here).

#### 7.1.3 Damerau-Levenshtein (Full — Adjacent Transpositions, unrestricted)

- **Category:** character-based, edit distance with transposition
- **Primary source:** Lowrance, R., Wagner, R. A. (1975). "An extension of the string-to-string correction problem." *Journal of the ACM*, 22(2):177–183. The full Damerau-Levenshtein with unrestricted transpositions (no OSA restriction).
- **AlgoID:** `AlgoDamerauLevenshteinFull`
- **Public functions:**
  - `DamerauLevenshteinFullDistance(a, b string) int`
  - `DamerauLevenshteinFullDistanceRunes(a, b string) int`
  - `DamerauLevenshteinFullScore(a, b string) float64`
  - `DamerauLevenshteinFullScoreRunes(a, b string) float64`
- **Description:** Damerau-Levenshtein without the OSA restriction. Substrings may be re-edited after a transposition. Mathematically the "correct" Damerau-Levenshtein. Slightly higher constant-factor cost than OSA due to the position-lookup tables required.
- **Algorithm:** the Lowrance-Wagner formulation maintains a `last seen` table per alphabet character; transposition cost is computed using these positions.
- **Complexity:** O(m·n · |Σ|) time worst case where Σ is the alphabet (effectively O(m·n) for fixed alphabet), O(m·n) space.
- **Score normalisation:** identical to Levenshtein.
- **Edge cases:** identical to Levenshtein.
- **Intended use:** when correctness across pathological transposition cases matters more than constant-factor speed.
- **Reference vectors:** "ca"/"abc" → full DL distance 2 (one transposition + one insertion).

#### 7.1.4 Hamming

- **Category:** character-based, equal-length
- **Primary source:** Hamming, R. W. (1950). "Error detecting and error correcting codes." *Bell System Technical Journal*, 29(2):147–160.
- **AlgoID:** `AlgoHamming`
- **Public functions:**
  - `HammingDistance(a, b string) int` — byte-level. Length-mismatch policy: returns `max(len(a), len(b))` (silent max) per the locked comparison-data leniency contract (§6.A).
  - `HammingDistanceRunes(a, b string) int` — same policy on rune-counted lengths.
  - `HammingScore(a, b string) float64` — returns `0.0` when lengths differ.
  - `HammingScoreRunes(a, b string) float64`
- **Description:** number of positions at which corresponding characters differ. Defined for equal-length strings; on unequal-length inputs the implementation returns the silent-max distance (and `0.0` for the Score variants) as comparison-data leniency.
- **Complexity:** O(n) time, O(1) space.
- **Score normalisation:** `score = 1.0 - distance / len(a)` when lengths match; `0.0` when lengths differ.
- **Mathematical invariants:** identity, symmetry, range bounds. Triangle inequality holds on equal-length inputs.
- **Edge cases:** unequal lengths → silent-max distance / `0.0` score (no error, no panic, per §6.A comparison-data leniency); both empty (distance 0, score 1.0); consumers wanting to detect unequal-length inputs explicitly can call `fuzzymatch.Validate(a, b)` and check for `WarnUnequalLength` (see §11.5).
- **Intended use:** fixed-width codes (8-character audit IDs, hex hashes, equal-length fingerprints). Not useful for general identifier comparison.
- **Reference vectors:** "karolin"/"kathrin" → distance 3, score `1 - 3/7 ≈ 0.5714`. "1011101"/"1001001" → distance 2, score `1 - 2/7 ≈ 0.7143`.
- **History:** an earlier draft of this spec specified `HammingDistance(a, b string) (int, error)` with an `ErrHammingLengthMismatch` sentinel. Phase 8.5 Q1 locked the silent-max policy to match the shipped code and the comparison-data leniency framework — code wins, spec catches up.

#### 7.1.5 Jaro

- **Category:** character-based, name-matching
- **Primary source:** Jaro, M. A. (1989). "Advances in record-linkage methodology as applied to matching the 1985 census of Tampa, Florida." *Journal of the American Statistical Association*, 84(406):414–420.
- **AlgoID:** `AlgoJaro`
- **Public functions:**
  - `JaroScore(a, b string) float64`
  - `JaroScoreRunes(a, b string) float64`
- **Description:** counts matching characters within a positional window of `max(len(a), len(b))/2 - 1`, then penalises for transpositions among matched characters. Designed for short string matching with positional tolerance.
- **Formula:** if `m = 0`, return 0.0. Otherwise `J = (m/|a| + m/|b| + (m - t/2)/m) / 3`, where `m` is the count of matching characters and `t` is the count of transpositions among matched pairs.
- **Complexity:** O(m·n) time, O(m + n) space.
- **Score normalisation:** the formula itself produces a value in [0.0, 1.0]. Identical strings return 1.0; both empty return 1.0 (by convention, documented).
- **Mathematical invariants:** identity, symmetry, range bounds. Triangle inequality does NOT hold (Jaro is not a metric).
- **Edge cases:** both empty → 1.0; one empty → 0.0; identical → 1.0.
- **Implementation notes:** stack-allocated `[256]bool` match-flag arrays for inputs under 256 characters (zero heap allocation). The rune variant uses heap-allocated slices.
- **Intended use:** record-linkage and name matching where positional tolerance and transposition matter more than substitution.
- **Reference vectors:** "MARTHA"/"MARHTA" → Jaro = 0.9444. "DIXON"/"DICKSONX" → Jaro = 0.7667. "JELLYFISH"/"SMELLYFISH" → Jaro = 0.8963.

#### 7.1.6 Jaro-Winkler

- **Category:** character-based, name-matching with prefix bonus
- **Primary source:** Winkler, W. E. (1990). "String comparator metrics and enhanced decision rules in the Fellegi-Sunter model of record linkage." *Proceedings of the Section on Survey Research Methods*, American Statistical Association: 354–359.
- **AlgoID:** `AlgoJaroWinkler`
- **Public functions:**
  - `JaroWinklerScore(a, b string) float64`
  - `JaroWinklerScoreRunes(a, b string) float64`
- **Description:** Jaro with a prefix bonus: strings sharing a common prefix score higher. Designed specifically for cases where prefix agreement is more meaningful than later-position agreement (typical for personal names and identifier families like `request_id`/`request_uuid`).
- **Formula:** `JW = J + L · p · (1 - J)` where `J` is the Jaro score, `L` is the length of the common prefix capped at 4 characters, and `p` is the prefix scale (canonical value 0.1, with `p · L_max ≤ 0.25` to keep `JW` bounded). The bonus is applied only when `J ≥ 0.7` (Winkler's canonical boost threshold).
- **Constants:** `winklerPrefixScale = 0.1`, `winklerMaxPrefix = 4`, `winklerBoostThreshold = 0.7`. Exposed as unexported package constants with godoc comments citing the originating Winkler 1990 paper.
- **Complexity:** identical to Jaro.
- **Score normalisation:** the formula produces a value in [0.0, 1.0].
- **Mathematical invariants:** identity, symmetry, range bounds. Triangle inequality does NOT hold.
- **Edge cases:** identical to Jaro.
- **Implementation notes:** computes the Jaro score first, then applies the prefix bonus.
- **Intended use:** prefix-aligned identifier families, personal names.
- **Reference vectors:** "MARTHA"/"MARHTA" → Jaro 0.9444, JW = 0.9611. "DWAYNE"/"DUANE" → JW ≈ 0.8400. "DIXON"/"DICKSONX" → JW ≈ 0.8133.

#### 7.1.7 Strcmp95

- **Category:** character-based, refined Jaro-Winkler
- **Primary source:** Winkler, W. E. (1994). "Advanced methods for record linkage." *Proceedings of the Section on Survey Research Methods*, American Statistical Association: 467–472. Reference SAS/C implementation: U.S. Census Bureau (1995), `strcmp95.c`.
- **AlgoID:** `AlgoStrcmp95`
- **Public functions:**
  - `Strcmp95Score(a, b string) float64`
- **No Runes variant.** Strcmp95 is ASCII-only by design (the similar-character table is built from ASCII letter confusions per Winkler 1994 / Census Bureau `strcmp95.c`). For Unicode inputs, normalise via `fuzzymatch.Normalise(s, opts)` first; the diacritic-stripping and case-folding steps fold most Unicode input down to comparable ASCII. Phase 8.5 Q5 explicitly locks this absence of a `*Runes` variant.
- **Description:** Jaro-Winkler with two additional refinements: (a) similar-character matching (letters considered partially-matching when they are commonly confused, e.g. `A`/`E`, `O`/`0`), and (b) a long-string bonus that further boosts scores for long strings that share substantial common characters. Implemented per the U.S. Census Bureau's `strcmp95.c` reference (which is in the public domain) and the Winkler 1994 paper.
- **Score normalisation:** [0.0, 1.0]; identical to Jaro-Winkler shape.
- **Mathematical invariants:** identity (Score(x, x) = 1.0), symmetry, range bounds. No triangle inequality.
- **Edge cases:** identical to Jaro-Winkler.
- **Implementation notes:** the similar-character table is an unexported package-level `[][]byte` documented as derived from the Winkler 1994 paper's similarity matrix. The table is initialised once at package load via a `var` declaration (no `init()` function), and is read-only after declaration.
- **Intended use:** record linkage where Jaro-Winkler's basic prefix bonus is insufficient — typically census-style name matching and survey data deduplication.
- **Reference vectors:** to be cross-validated against `xrash/smetrics` `Strcmp95` implementation (MIT-licensed, useful for reference vector cross-validation only — code is reimplemented from the Winkler paper and Census Bureau reference C code).

#### 7.1.8 Smith-Waterman-Gotoh

- **Category:** character-based, local sequence alignment
- **Primary sources:**
  - Smith, T. F., Waterman, M. S. (1981). "Identification of common molecular subsequences." *Journal of Molecular Biology*, 147(1):195–197.
  - Gotoh, O. (1982). "An improved algorithm for matching biological sequences." *Journal of Molecular Biology*, 162(3):705–708.
  - Flouri, T. et al. (2015). "Are all global alignment algorithms and implementations correct?" *bioRxiv* 031500 — documents the Gotoh 1982 initialisation erratum and the corrected formulation transcribed in the implementation.
- **AlgoID:** `AlgoSmithWatermanGotoh`
- **Public types:**
  - `type SWGParams struct { Match, Mismatch, GapOpen, GapExtend float64 }`
- **Public constructors:**
  - `NewSWGParams() SWGParams` — returns a value populated with the documented defaults (Match=1.0, Mismatch=-1.0, GapOpen=-1.5, GapExtend=-0.5). Callers may override individual fields after construction.
- **Public functions (normalised, clamped to [0.0, 1.0]):**
  - `SmithWatermanGotohScore(a, b string) float64`
  - `SmithWatermanGotohScoreRunes(a, b string) float64`
  - `SmithWatermanGotohScoreWithParams(a, b string, params SWGParams) float64`
- **Public functions (raw, unclamped):**
  - `SmithWatermanGotohRawScore(a, b string) float64`
  - `SmithWatermanGotohRawScoreRunes(a, b string) float64`
  - `SmithWatermanGotohRawScoreWithParams(a, b string, params SWGParams) float64`
- **Raw\* surface rationale:** the `*RawScore*` variants return the UNCLAMPED raw alignment score, which may be negative (two unrelated strings dominated by mismatch/gap penalties) or exceed `min(len(a), len(b))` when custom params produce `Match > 1.0`. Advanced consumers (bioinformatics, schema-similarity research) who want absolute alignment quality unaffected by the normalisation choice should use the `*RawScore*` variants; consumers who want a comparable [0.0, 1.0] similarity should use the `*Score*` variants (which apply `clamp(raw / min(len(a), len(b)), 0, 1)`). Surface expansion approved by the project owner on 2026-05-14 per `.planning/phases/03-smith-waterman-gotoh/03-CONTEXT.md` §4.
- **Description:** local sequence alignment with affine gap penalty. Unlike global-alignment algorithms (Needleman-Wunsch), Smith-Waterman finds the best-matching subsequence anywhere in the two strings without insisting on end-to-end alignment. Gotoh's improvement makes the affine gap penalty model (gap-open cost separate from gap-extend cost) efficient.
- **Default parameters:** match reward = 1.0, mismatch penalty = -1.0, gap-open penalty = -1.5, gap-extend penalty = -0.5. Obtained via `NewSWGParams()`. Customisable via `SmithWatermanGotohScoreWithParams` / `SmithWatermanGotohRawScoreWithParams`. There is intentionally no exported `SWGDefaultParams` package-level variable — callers construct fresh values via `NewSWGParams()` to avoid a "is this read-only?" footgun.
- **Parameter validation:** none in `*Score` / `*RawScore` functions; nonsense params (e.g. `GapOpen > 0`, NaN, +Inf) produce deterministic-but-meaningless results with no errors and no panics. The `SWGParams` godoc documents expected ranges: `Match >= 0`, `Mismatch <= 0`, `GapOpen <= GapExtend <= 0`. The Scorer layer (Phase 8) may add validation at composition time.
- **Complexity:** O(m·n) time. Space: O(min(m,n)) via the two-row three-matrix DP form (six rolling rows of length `min(m,n)+1` — `prevM`/`currM`/`prevIx`/`currIx`/`prevIy`/`currIy`). Stack-allocated on a single `[(maxStackInputLen+1)*6]float64` buffer (3120 bytes) when both inputs are ASCII and the shorter dimension is ≤ 64; heap-allocated otherwise.
- **Score normalisation:** `score = clamp(best_local_score / min(len(a), len(b)), 0, 1)`. Identical inputs return 1.0; both empty return 1.0; one empty returns 0.0.
- **Mathematical invariants:** identity, symmetry, range bounds. No triangle inequality (SWG is not a metric over the full string space).
- **SWG-specific invariants (property-tested):**
  - Gap-split invariance — splitting a single long gap into two halves with intervening match characters that don't affect the local alignment must NOT improve the score (canonical Gotoh-erratum canary per `.planning/research/PITFALLS.md` §3).
  - Raw upper bound — `RawScore(a, b) <= Match × min(len(a), len(b))` always (best local alignment has at most `min(len)` match positions).
  - Monotonic with Match reward — increasing the `Match` parameter (keeping others fixed) cannot decrease `RawScore` for any input pair.
- **Edge cases:** as above.
- **Implementation notes:** the affine gap penalty requires three DP matrices (M, Ix, Iy) in the Gotoh formulation, implemented here as a two-row rolling form. The Gotoh 1982 paper contains a known initialisation erratum (the global-alignment border setup that textbook treatments often blur into local alignment); this implementation uses the corrected Flouri et al. 2015 formulation where every border cell of M, Ix, Iy initialises to 0 for local alignment (NOT −∞, NOT the global-alignment gap-open ladder).
- **Cross-validation:** the implementation is cross-validated against biopython's `Bio.Align.PairwiseAligner` (mode=`"local"`) via a committed JSON corpus at `testdata/cross-validation/swg/vectors.json` (16 entries including the load-bearing `one_long_gap_canary` Gotoh-erratum gate at biopython_normalised=0.5). Tolerance: `|our_score − biopython_normalised| <= 1e-9`. Regeneration is developer-only (`make regen-swg-cross-validation`); CI consumes the committed JSON via `TestSWG_CrossValidation` without requiring Python.
- **Intended use:** detecting that one name is a substring or near-substring of another (`http_request` vs `http_request_header_fields`), or that two names share a long common middle section despite different prefixes/suffixes.

#### 7.1.9 LCSStr (Longest Common Substring)

- **Category:** character-based, common-substring length
- **Primary source:** Wagner, R. A., Fischer, M. J. (1974). "The string-to-string correction problem." *Journal of the ACM*, 21(1):168–173. Standard dynamic-programming formulation.
- **AlgoID:** `AlgoLCSStr`
- **Public functions:**
  - `LongestCommonSubstring(a, b string) string`
  - `LongestCommonSubstringRunes(a, b string) string`
  - `LCSStrScore(a, b string) float64`
  - `LCSStrScoreRunes(a, b string) float64`
- **Description:** the longest contiguous substring shared by two strings.
- **Recurrence:** `D[i,j] = D[i-1,j-1] + 1 if a[i-1] = b[j-1], else 0`. Track max value and ending position.
- **Complexity:** O(m·n) time, O(min(m,n)) space using a rolling 1-D buffer plus max tracking.
- **Score normalisation:** `score = 2 · len(lcs) / (len(a) + len(b))`. Identical → 1.0; both empty → 1.0 (by convention); one empty → 0.0; no shared characters → 0.0.
- **Mathematical invariants:** identity, symmetry, range bounds. No triangle inequality.
- **Edge cases:** as above.
- **Intended use:** detecting names with a long common middle (`my_request_id` vs `your_request_handle` shares `_request_`).

### 7.2 Q-gram / N-gram algorithms

For all q-gram / n-gram algorithms, q-grams are extracted as overlapping substrings of length `n`. Default `n = 3` (trigrams) per the recommendation in the cited primary sources. Q-grams are extracted byte-wise by default (with a rune-wise variant exposed for Unicode-heavy domains). Padding is NOT applied — q-grams are extracted from the raw string without start/end markers.

#### 7.2.1 Q-Gram Jaccard

- **Category:** q-gram, set similarity
- **Primary sources:**
  - Ukkonen, E. (1992). "Approximate string-matching with q-grams and maximal matches." *Theoretical Computer Science*, 92(1):191–211.
  - Jaccard, P. (1912). "The distribution of the flora in the alpine zone." *New Phytologist*, 11(2):37–50.
- **AlgoID:** `AlgoQGramJaccard`
- **Public functions:**
  - `QGramJaccardScore(a, b string, n int) float64`
  - `QGramJaccardScoreRunes(a, b string, n int) float64`
- **Description:** extract overlapping character n-grams from both strings into sets, compute the Jaccard index `|A ∩ B| / |A ∪ B|`.
- **Complexity:** O(|a| + |b|) time (after q-gram extraction), O(|a| + |b|) space.
- **Score normalisation:** the Jaccard index is naturally in [0.0, 1.0]. Both-empty sets return 1.0 by convention; one-empty returns 0.0; identical strings return 1.0.
- **Mathematical invariants:** identity, symmetry, range bounds. Triangle inequality does NOT hold.
- **Edge cases:**
  - `n > min(len(a), len(b))`: at least one of the q-gram sets is empty; returns 0.0 (or 1.0 if both empty)
  - `n < 1`: documented panic or returned as `ErrInvalidQGramSize` depending on whether called via direct function (panics on invalid `n`) or via Scorer (returns error at construction time)
- **Implementation notes:** q-gram extraction uses a `map[string]int` for multiset counts (Jaccard on multisets gives the standard set-theoretic Jaccard when treating duplicates as distinct elements). Implementation must NOT expose map iteration order to the output path. For deterministic output, extract q-grams in input order and accumulate into sorted slices before any output.
- **Intended use:** trigram-level partial-match detection, abbreviation handling.

#### 7.2.2 Sørensen-Dice

- **Category:** q-gram, set similarity
- **Primary sources:**
  - Dice, L. R. (1945). "Measures of the amount of ecologic association between species." *Ecology*, 26(3):297–302.
  - Sørensen, T. (1948). "A method of establishing groups of equal amplitude in plant sociology based on similarity of species and its application to analyses of the vegetation on Danish commons." *Kongelige Danske Videnskabernes Selskab*, 5(4):1–34.
- **AlgoID:** `AlgoSorensenDice`
- **Public functions:**
  - `SorensenDiceScore(a, b string, n int) float64` — defaults `n = 2` (bigrams) when called via Scorer with no explicit n
  - `SorensenDiceScoreRunes(a, b string, n int) float64`
- **Description:** Dice coefficient on character n-grams. `DSC = 2|A ∩ B| / (|A| + |B|)`.
- **Complexity:** identical to Q-Gram Jaccard.
- **Score normalisation:** naturally in [0.0, 1.0].
- **Mathematical invariants:** identity, symmetry, range bounds. No triangle inequality.
- **Edge cases:** identical to Q-Gram Jaccard.
- **Intended use:** slightly more permissive than Jaccard for partially-overlapping sets. The bigram form is widely used in fuzzy-search applications and DNA sequence similarity.

#### 7.2.3 Cosine (n-gram)

- **Category:** q-gram, vector similarity
- **Primary source:** Salton, G., McGill, M. J. (1983). *Introduction to Modern Information Retrieval*. McGraw-Hill. (The vector-space model and cosine similarity for IR are textbook standard.)
- **AlgoID:** `AlgoCosine`
- **Public functions:**
  - `CosineScore(a, b string, n int) float64`
  - `CosineScoreRunes(a, b string, n int) float64`
- **Description:** treat each string's n-gram frequencies as a vector in n-gram space; compute the cosine of the angle between the two vectors.
- **Formula:** `cos(A, B) = (A · B) / (‖A‖ · ‖B‖)`.
- **Complexity:** O(|a| + |b|) time for extraction; O(|A ∪ B|) for the dot product.
- **Score normalisation:** the cosine of an angle between non-negative vectors is naturally in [0.0, 1.0].
- **Mathematical invariants:** identity, symmetry, range bounds.
- **Edge cases:** both empty → 1.0; one empty → 0.0.
- **Intended use:** complements Jaccard and Dice by being length-asymmetry tolerant (a short string matching a long string with proportional q-gram frequency scores well).

#### 7.2.4 Tversky Index

- **Category:** q-gram, asymmetric set similarity
- **Primary source:** Tversky, A. (1977). "Features of similarity." *Psychological Review*, 84(4):327–352.
- **AlgoID:** `AlgoTversky`
- **Public functions:**
  - `TverskyScore(a, b string, n int, alpha, beta float64) float64`
  - `TverskyScoreRunes(a, b string, n int, alpha, beta float64) float64`
- **Description:** asymmetric generalisation of Jaccard and Dice. `T(A, B) = |A ∩ B| / (|A ∩ B| + α·|A − B| + β·|B − A|)`. With `α = β = 1` reduces to Jaccard; with `α = β = 0.5` reduces to Sørensen-Dice. Asymmetric when `α ≠ β`.
- **Complexity:** identical to Q-Gram Jaccard.
- **Score normalisation:** [0.0, 1.0].
- **Mathematical invariants:** identity. Symmetry holds when `α = β`. Range bounds always.
- **Edge cases:** both-empty handled as 1.0; one-empty as 0.0; identical as 1.0.
- **Intended use:** when one string is intentionally treated as a "prototype" and the other as a "variant" (e.g. a canonical schema name vs. a candidate match). For symmetric use, `α = β` is required to keep the Scorer composite well-defined.

### 7.3 Token-based algorithms

Token-based algorithms operate on the result of `Tokenise(s, opts)` rather than on raw strings or character n-grams. Tokenisation rules are specified in section 10. Defaults: lowercase, split on the SeparatorChars set, split camelCase / PascalCase / acronym boundaries.

**No Runes variants in the token tier.** Token-tier algorithms (Monge-Elkan, Token Sort Ratio, Token Set Ratio, Partial Ratio, Token Jaccard) operate on byte slices post-`Tokenise`. `Tokenise` is itself rune-aware (it splits on Unicode code-point boundaries for camelCase / PascalCase / acronym detection), so the byte-level Indel kernel produces correct results on Unicode inputs because each emitted token is already a complete code-point sequence. Consequently the token tier exposes only the byte-string entry points and does not ship `*ScoreRunes` variants. Locked Phase 8.5 Q5.

#### 7.3.1 Monge-Elkan

- **Category:** token-based, hybrid (uses an inner character-based metric)
- **Primary source:** Monge, A. E., Elkan, C. P. (1996). "The field matching problem: algorithms and applications." *Proceedings of the Second International Conference on Knowledge Discovery and Data Mining*: 267–270.
- **AlgoID:** `AlgoMongeElkan`
- **Public functions:**
  - `MongeElkanScore(a, b string, inner AlgoID) float64` — **symmetric variant** (default). Returns the average of the two directional Monge-Elkan scores: `(ME(A,B) + ME(B,A)) / 2`. This is the Scorer-facing surface.
  - `MongeElkanScoreAsymmetric(a, b string, inner AlgoID) float64` — directional variant. For each token in `A`, find the maximum-similarity token in `B` using the inner metric, then average. Inherently asymmetric.
- **Signature change (Phase 8.5 Q3):** the v0.x API had `MongeElkanScore` as the asymmetric variant and `MongeElkanScoreSymmetric` as the wrapper. In v1.0 the naming inverts: `MongeElkanScore` is the safe-default symmetric variant; consumers who explicitly want directional behaviour call `MongeElkanScoreAsymmetric`. The `NormalisationOptions` parameter has been removed from both functions — it had no effect inside Monge-Elkan (which composes tokenisation from the call site, not normalisation choices).
- **Description:** symmetric Monge-Elkan: for each token in each string, find the maximum-similarity token in the other string using the inner metric, average within each direction, then average the two directions.
- **Formula:** asymmetric `ME(A, B) = (1/|A|) · Σ_{a ∈ A} max_{b ∈ B} sim_inner(a, b)`; symmetric `(ME(A,B) + ME(B,A)) / 2`.
- **Default inner metric:** Jaro-Winkler (per the original paper).
- **Permitted inner metrics:** any AlgoID in the documented `permittedMongeElkanInner` allow-list (character + q-gram + phonetic + gestalt tiers). Token-tier AlgoIDs (`AlgoTokenSortRatio`, `AlgoTokenSetRatio`, `AlgoPartialRatio`, `AlgoTokenJaccard`) and the `AlgoMongeElkan` self-reference are rejected. Direct callers passing an invalid inner AlgoID receive a panic with `ErrInvalidInnerAlgo` per §6.A. Scorer-construction callers receive the same sentinel as a typed error from `WithMongeElkanAlgorithm(weight, inner)`.
- **Complexity:** O(|A| · |B| · cost(inner)) where cost(inner) is the inner metric's per-comparison cost; the symmetric variant doubles the constant factor.
- **Score normalisation:** the formula naturally produces a value in [0.0, 1.0] assuming the inner metric is bounded in [0.0, 1.0].
- **Symmetry:** `MongeElkanScore` (symmetric) is symmetric by construction; `MongeElkanScoreAsymmetric` violates symmetry (this is property-tested).
- **Edge cases:** empty token set on one or both sides handled per the inner metric's edge cases.
- **Intended use:** identifier families where token-level matching matters more than character-level matching, with the inner metric handling intra-token similarity (e.g. `user_create_event` vs `usr_creating_evt` — tokens align but each pair has its own similarity).

#### 7.3.2 Token Sort Ratio

- **Category:** token-based, sort-and-compare
- **Primary source:** SeatGeek (2014). *fuzzywuzzy* Python library, `fuzz.token_sort_ratio` implementation. Canonical modern reference: RapidFuzz documentation (Bachmann, M., 2020–present), <https://rapidfuzz.github.io/RapidFuzz/>. (No formal academic source exists; this is a practical engineering pattern.)
- **AlgoID:** `AlgoTokenSortRatio`
- **Public functions:**
  - `TokenSortRatioScore(a, b string, opts NormalisationOptions) float64`
- **Description:** tokenise both strings, sort the tokens, rejoin with a single space, then compute an Indel-based ratio (essentially `1 - levenshtein_distance / total_length` but using a Longest Common Subsequence formulation rather than Levenshtein — the Indel ratio is `2·LCS / (|a| + |b|)`).
- **Complexity:** O((|a| + |b|) · log(|a| + |b|)) for the sort, plus O(|a| · |b|) for the LCS.
- **Score normalisation:** [0.0, 1.0].
- **Mathematical invariants:** identity, symmetry, range bounds.
- **Edge cases:** empty tokens lists on both sides → 1.0; one empty → 0.0; identical post-sort strings → 1.0.
- **Intended use:** comparing strings that should be equal up to token reordering (`UserCreateEvent` vs `CreateUserEvent`).

#### 7.3.3 Token Set Ratio

- **Category:** token-based, set-and-compare
- **Primary source:** as Token Sort Ratio (SeatGeek `fuzzywuzzy`; modern reference RapidFuzz).
- **AlgoID:** `AlgoTokenSetRatio`
- **Public functions:**
  - `TokenSetRatioScore(a, b string, opts NormalisationOptions) float64`
- **Description:** tokenise both strings, compute three sub-strings — the intersection sorted-and-joined; intersection + difference-from-a sorted-and-joined; intersection + difference-from-b sorted-and-joined — then compute the maximum Indel ratio among the three pairwise comparisons.
- **Complexity:** O((|a| + |b|) · log(|a| + |b|)) for sorting, O(|a| · |b|) for the three LCS comparisons.
- **Score normalisation:** [0.0, 1.0].
- **Mathematical invariants:** identity, symmetry, range bounds.
- **Edge cases:** as Token Sort Ratio.
- **Intended use:** comparing strings with substantially different token counts but a meaningful shared core (`http_request` vs `http_request_body_payload`).

#### 7.3.4 Partial Ratio

- **Category:** token-based, sliding-window
- **Primary source:** as Token Sort Ratio (SeatGeek `fuzzywuzzy`; modern reference RapidFuzz `fuzz.partial_ratio`).
- **AlgoID:** `AlgoPartialRatio`
- **Public functions:**
  - `PartialRatioScore(a, b string) float64`
- **No Runes variant.** Token-tier algorithms operate on the output of `Tokenise(s, opts)`, which is itself rune-aware. Post-Tokenise the byte-level Indel kernel produces correct results on Unicode inputs because each token is a complete code-point sequence. Phase 8.5 Q5 explicitly removes `PartialRatioScoreRunes` for token-tier symmetry; the same rationale applies to Token Sort Ratio, Token Set Ratio, and Token Jaccard.
- **Description:** slide the shorter string across the longer string, compute the Indel ratio at each window position, return the maximum.
- **Complexity:** O(m·n) per window position, O(n·m·(n-m)) overall. Optimisation: use a sliding-window technique with the Indel DP to amortise. v1 implementation may use the straightforward O(n·m·(n-m)); v1.x optimisation as a separate issue.
- **Score normalisation:** [0.0, 1.0].
- **Mathematical invariants:** identity (Score(x, x) = 1.0), symmetry (by construction — based on shorter-of-the-two), range bounds.
- **Edge cases:** both empty → 1.0; one empty → 0.0.
- **Intended use:** detecting that one string contains a near-perfect match of the other as a substring (e.g. `request_id` matching anywhere inside `http_request_id_v2`).

#### 7.3.5 Token Jaccard

- **Category:** token-based, set similarity
- **Primary source:** Jaccard, P. (1912). (As Q-Gram Jaccard, applied to word tokens rather than character n-grams.)
- **AlgoID:** `AlgoTokenJaccard`
- **Public functions:**
  - `TokenJaccardScore(a, b string, opts NormalisationOptions) float64`
- **Description:** tokenise both strings, compute Jaccard index on the token sets.
- **Complexity:** O(|a| + |b|) after tokenisation.
- **Score normalisation:** [0.0, 1.0]. Both-empty → 1.0; one-empty → 0.0; identical → 1.0.
- **Mathematical invariants:** identity, symmetry, range bounds.
- **Edge cases:** as above.
- **Intended use:** camelCase / snake_case equivalence at the word level. Particularly useful for identifier-style names where word order varies but the word set is the same or near-same.

### 7.4 Phonetic algorithms

Phonetic algorithms encode input strings into pronunciation-equivalent keys. They are inherently boolean — two strings either share an encoded key or they don't. Within the Scorer, phonetic algorithms contribute 1.0 to their weighted slot if the keys match exactly and 0.0 otherwise. The underlying encoded keys are exposed via separate public functions for consumers that want richer behaviour (e.g. computing a Levenshtein score on the codes themselves).

The phonetic encoding rules are language-specific. All implementations in this library are tuned for English-language pronunciation; cross-language usage may produce poor results. This is documented in each algorithm's godoc.

**No Runes variants in the phonetic tier.** Soundex, Double Metaphone, NYSIIS, and MRA are ASCII-only by design — every published rule set operates on Latin-alphabet letters. For Unicode inputs, normalise via `fuzzymatch.Normalise(s, opts)` first; diacritic stripping plus case folding produces comparable ASCII for most input. Phase 8.5 Q5 locks the absence of `*Runes` variants on all four phonetic algorithms.

#### 7.4.1 Soundex

- **Category:** phonetic, English
- **Primary source:** Russell, R. C., Odell, M. K. (1918, 1922). U.S. Patents 1261167 and 1435663. Canonical algorithm description: Knuth, D. E. (1973). *The Art of Computer Programming, Volume 3: Sorting and Searching*, Section 6.4.
- **AlgoID:** `AlgoSoundex`
- **Public functions:**
  - `SoundexCode(s string) string` — returns the 4-character code (one uppercase letter + three digits)
  - `SoundexScore(a, b string) float64` — returns 1.0 if `SoundexCode(a) == SoundexCode(b)`, else 0.0
- **Description:** retain first letter; map remaining letters to digit groups (B/F/P/V→1, C/G/J/K/Q/S/X/Z→2, D/T→3, L→4, M/N→5, R→6, vowels and H/W dropped after the first letter, consecutive same-group letters collapsed); truncate or zero-pad to 4 characters.
- **Score normalisation:** binary 0.0 / 1.0.
- **Mathematical invariants:** identity, symmetry, range bounds.
- **Edge cases:** empty input → empty code → `SoundexScore("", "") = 1.0`; non-ASCII input handled by encoding only the ASCII letters (documented limitation).
- **Intended use:** rough pronunciation-equivalent matching for English names. Pre-filter, not primary metric.
- **Reference vectors:** "Robert" → "R163"; "Rupert" → "R163"; "Rubin" → "R150"; "Ashcraft" → "A261"; "Ashcroft" → "A261".

#### 7.4.2 Double Metaphone

- **Category:** phonetic, multi-language tolerant
- **Primary source:** Philips, L. (2000). "The double-metaphone search algorithm." *C/C++ Users Journal*, 18(6):38–43. Reference implementation: original C code by Lawrence Philips, public-domain.
- **AlgoID:** `AlgoDoubleMetaphone`
- **Public functions:**
  - `DoubleMetaphoneKeys(s string) (primary, secondary string)` — returns the primary and secondary keys (4 characters each in the canonical formulation, though longer in some extended variants; this library uses the canonical 4-character truncation)
  - `DoubleMetaphoneScore(a, b string) float64` — returns 1.0 if either of `a`'s keys matches either of `b`'s keys (i.e. primary-primary, primary-secondary, secondary-primary, or secondary-secondary), else 0.0
- **Description:** improvement over Soundex that handles non-English-origin English-language names (Germanic, Slavic, Romance, Greek-origin). Encodes each input into two possible keys reflecting alternate pronunciations.
- **Score normalisation:** binary 0.0 / 1.0.
- **Mathematical invariants:** identity, symmetry, range bounds.
- **Edge cases:** empty input → empty keys → `DoubleMetaphoneScore("", "") = 1.0`.
- **Implementation notes:** the encoding rule table is large (200+ rules in the canonical algorithm). Implementation derived from the Philips 2000 paper and the public-domain C reference. Implementation must NOT copy from MIT-licensed Go ports (e.g. `CalypsoSys/godoublemetaphone`) — algorithm rules are encoded fresh from the primary source. Cross-validation against existing implementations is permitted for reference vectors only.
- **Intended use:** robust pronunciation-equivalent matching for English-language names of various ethnic origins.
- **Reference vectors:** "Smith" → ("SM0", "XMT"); "Schmidt" → ("XMT", "SMT"); these share `XMT`, so they match. "Catherine" → ("K0RN", "KTRN"); "Katherine" → ("K0RN", "KTRN") — exact match.

#### 7.4.3 NYSIIS (New York State Identification and Intelligence System)

- **Category:** phonetic, English
- **Primary source:** Taft, R. L. (1970). *Name search techniques*. New York State Identification and Intelligence System, Special Report No. 1. Albany, NY.
- **AlgoID:** `AlgoNYSIIS`
- **Public functions:**
  - `NYSIISCode(s string) string` — returns the 6-character code (truncated)
  - `NYSIISScore(a, b string) float64` — returns 1.0 if codes match exactly, else 0.0
- **Description:** an improvement over Soundex for English-language names, developed by the New York State criminal-justice authorities in the late 1960s. Produces a 6-character code via a specific letter-substitution and reduction ruleset.
- **Score normalisation:** binary 0.0 / 1.0.
- **Mathematical invariants:** identity, symmetry, range bounds.
- **Edge cases:** empty input → empty code → score 1.0.
- **Intended use:** higher-accuracy alternative to Soundex for English names, particularly useful for surname matching.
- **Reference vectors:** "Robert" → "RABAD" (variants exist; the library uses the canonical 1970 algorithm); "Brown" → "BRAN"; "Browne" → "BRAN".

#### 7.4.4 MRA (Match Rating Approach)

- **Category:** phonetic, English
- **Primary source:** Moore, G. B., Kuhns, J. L., Trefftzs, J. L., Montgomery, C. A. (1977). *Accessing individual records from personal data files using non-unique identifiers*. National Bureau of Standards (later NIST), Technical Note 943.
- **AlgoID:** `AlgoMRA`
- **Public functions:**
  - `MRACode(s string) string` — returns the encoded form (canonical form: consonant-only after the first letter)
  - `MRACompare(a, b string) (bool, int)` — returns whether the strings match per the MRA threshold rule, plus the raw similarity score
  - `MRAScore(a, b string) float64` — returns 1.0 if MRA-comparison passes, else 0.0
- **Description:** encode each name by removing vowels (except leading), removing duplicate consonants, and truncating; then compare via a position-aware similarity counting that requires the similarity score to exceed a length-dependent threshold.
- **Score normalisation:** binary 0.0 / 1.0 (the underlying score is 0–6 integer; binary at the threshold).
- **Mathematical invariants:** identity, symmetry (the comparison is symmetric by the MRA rules), range bounds.
- **Edge cases:** strings whose encoded lengths differ by more than 3 are documented as automatic mismatch (score 0.0) per the canonical MRA algorithm.
- **Intended use:** highly-tuned phonetic match for English-language surnames in record-linkage contexts.

### 7.5 Gestalt pattern matching

#### 7.5.1 Ratcliff-Obershelp

- **Category:** gestalt, recursive longest-common-substring
- **Primary source:** Ratcliff, J. W., Metzener, D. E. (1988). "Pattern matching: the gestalt approach." *Dr. Dobb's Journal*, 13(7):46–51.
- **AlgoID:** `AlgoRatcliffObershelp`
- **Public functions:**
  - `RatcliffObershelpScore(a, b string) float64`
  - `RatcliffObershelpScoreRunes(a, b string) float64`
- **Description:** find the longest common substring; recursively apply the same to the unmatched prefix and suffix sub-pairs; the score is twice the total matched length divided by the sum of input lengths. This is the algorithm Python's `difflib.SequenceMatcher.ratio()` implements.
- **Formula:** `score = 2·M / (|a| + |b|)` where `M` is the sum of lengths of all matched substrings found by recursive longest-common-substring decomposition.
- **Complexity:** average O(n·m); worst case O(n²·m) or O(n·m²). Documented as such; not used in tight loops without consideration.
- **Score normalisation:** [0.0, 1.0]. Both-empty → 1.0; one-empty → 0.0.
- **Asymmetric by design — difflib parity (LOCKED OQ-1, 2026-05-14; re-confirmed Phase 8.5 Q6a):** `RatcliffObershelpScore(a, b)` is **not** guaranteed to equal `RatcliffObershelpScore(b, a)`. The leftmost-tie-break rule on the longest-common-substring split produces a directional decomposition that matches Python `difflib.SequenceMatcher.ratio()` byte-for-byte. This is intentional — the library ships true difflib equivalence, not a symmetrised approximation. The general property-test set excludes symmetry for `AlgoRatcliffObershelp`; cross-validation against Python `difflib(autojunk=False)` is the load-bearing acceptance test. The parallel exception note lives in `.claude/skills/algorithm-correctness-standards/SKILL.md` (referenced from `PropAlgorithmScore_Symmetric`'s allow-list).
- **Mathematical invariants:** identity, range bounds. **NOT symmetric** (see above). No triangle inequality.
- **Edge cases:** as above.
- **Intended use:** general-purpose human-perceived-similarity scoring. Often considered the closest mechanical metric to human similarity judgement for arbitrary strings. Consumers who need a symmetric variant can call `(RatcliffObershelpScore(a, b) + RatcliffObershelpScore(b, a)) / 2` at the call site.

---

## 8. Layer 2 — Scorer

> *Reminder: the Scorer construction, options, and method shapes below are illustrative. The `api-ergonomics-reviewer` and `user-guide-reviewer` agents determine the final API. Scorer behaviour and semantics (weighted composite, configurable threshold, normalisation, default composition, concurrent-use safety, deterministic output) are requirements.*

The `Scorer` composes any subset of the 23 algorithms into a weighted similarity score. Scorer is the recommended API for most consumers; the standalone algorithm functions are exposed for consumers who need just one algorithm with no overhead.

### 8.1 Construction

```go
type Scorer struct {
    // unexported fields
}

func NewScorer(opts ...ScorerOption) (*Scorer, error)
```

Construction validates the configuration and returns an error for invalid configurations (per the parameter-input strict policy in §6.A):

- No algorithms configured → `ErrEmptyScorer`
- Any weight `≤ 0` or `NaN` or `±Inf` → `ErrInvalidWeight`
- Sum of weights `≤ 0` → `ErrInvalidWeight`
- Threshold fails the guard `math.IsNaN(t) || t < 0.0 || t > 1.0` → `ErrInvalidThreshold` (Phase 8.5 Q2 added the `math.IsNaN(t)` clause; `±Inf` was already rejected by the range check, so no separate `±Inf` clause is required. See §6.A "`WithThreshold` explicit check" for the full guard.)
- Invalid algorithm ID → `ErrInvalidAlgoID`
- Q-gram size `< 1` → `ErrInvalidQGramSize`
- Tversky `α < 0`, `β < 0`, or `α + β ≤ 0` → `ErrInvalidTverskyParam` (the `α + β ≤ 0` guard was added in Phase 8.5 Q2 — `WithTverskyAlgorithm(_, 0, 0, _)` previously constructed successfully then panicked at first `Score` call)
- Monge-Elkan inner = `AlgoMongeElkan` (self-reference) or not in the permitted set → `ErrInvalidAlgoID`

After successful construction, the Scorer is immutable. All Scorer methods (`Score`, `ScoreAll`, `Match`) are safe for concurrent use and follow the comparison-data lenient policy (§6.A) at runtime — they never panic on consumer-supplied strings and never return an error.

### 8.2 Options

```go
type ScorerOption func(*scorerConfig) error

// Simple algorithms with one float weight.
func WithAlgorithm(algo AlgoID, weight float64) ScorerOption

// Algorithms with additional parameters.
func WithQGramJaccardAlgorithm(weight float64, n int) ScorerOption
func WithCosineAlgorithm(weight float64, n int) ScorerOption
func WithSorensenDiceAlgorithm(weight float64, n int) ScorerOption
func WithTverskyAlgorithm(weight, alpha, beta float64, n int) ScorerOption
func WithMongeElkanAlgorithm(weight float64, inner AlgoID) ScorerOption
func WithSmithWatermanGotohAlgorithm(weight float64, params SWGParams) ScorerOption

// Normalisation configuration.
func WithNormalisation(opts NormalisationOptions) ScorerOption
func WithoutNormalisation() ScorerOption

// Threshold for Match.
func WithThreshold(t float64) ScorerOption

// Weight normalisation control.
func WithNormaliseWeights(normalise bool) ScorerOption  // default true
```

`WithAlgorithm(algo, weight)` accepts any AlgoID. For algorithms with parameters (Q-Gram Jaccard, Sørensen-Dice, Cosine, Tversky, Monge-Elkan, Smith-Waterman-Gotoh), calling `WithAlgorithm` uses the algorithm's default parameters (e.g. `n = 3` for Q-Gram Jaccard, `α = β = 1` for Tversky-reduces-to-Jaccard, Jaro-Winkler inner for Monge-Elkan, default params for SWG). For non-default parameters, use the dedicated `With...Algorithm` options.

If the same algorithm is added multiple times via different option calls, the last one wins. This is documented behaviour.

### 8.3 Methods

```go
// Score returns the weighted composite similarity score in [0.0, 1.0].
func (s *Scorer) Score(a, b string) float64

// ScoreAll returns the per-algorithm scores keyed by AlgoID (typed
// enum, NOT string). Map iteration order is NOT stable; consumers
// requiring stable ordering should sort the keys. The returned map
// is a fresh allocation per call (safe to modify).
//
// SPEC OVERRIDE (Phase 8): originally specified as map[string]float64
// in this section; the implementation uses map[AlgoID]float64 for
// compile-time type safety. See .planning/phases/08-composite-scorer/
// 08-CONTEXT.md §1 for the rationale and the api-ergonomics-reviewer
// sign-off (recorded in plan 08-03's PR). REQUIREMENTS.md SCORER-05
// already specifies map[AlgoID]float64 — this is the canonical form.
func (s *Scorer) ScoreAll(a, b string) map[AlgoID]float64

// Match returns true if Score(a, b) >= threshold.
func (s *Scorer) Match(a, b string) bool

// Threshold returns the configured threshold.
func (s *Scorer) Threshold() float64

// Algorithms returns the list of configured algorithms with their
// normalised weights, sorted by AlgoID ascending. Returned slice is
// a fresh allocation per call.
func (s *Scorer) Algorithms() []ScorerAlgorithm

// NormalisationOptions returns the NormalisationOptions stored at
// construction time, along with a boolean indicating whether the
// Scorer applies normalisation (false when the Scorer was constructed
// with WithoutNormalisation). The returned options are by-value —
// mutating the returned struct does not affect the Scorer. Used by
// the scan sub-package (github.com/axonops/fuzzymatch/scan) to
// canonicalise SuppressedPairs entries and build token buckets using
// the same normalisation pipeline the Scorer uses for scoring, per
// §12.3 and §12.5. Safe for concurrent use; the Scorer is immutable
// after NewScorer. Added in Phase 9 plan 09-01 to resolve
// 09-RESEARCH.md Open Question 1 (how scan accesses the Scorer's
// normalisation state without coupling Config to a duplicate field).
func (s *Scorer) NormalisationOptions() (opts NormalisationOptions, applied bool)

// ScorerAlgorithm describes one algorithm configured on a Scorer.
type ScorerAlgorithm struct {
    ID     AlgoID
    Weight float64  // post-normalisation weight in [0.0, 1.0]
}
```

### 8.4 Weight semantics

If `WithNormaliseWeights(true)` (the default), the configured raw weights are normalised at construction time to sum to 1.0. The `Score` method then computes `Σ wᵢ · simᵢ` over the configured algorithms. The composite score is guaranteed to be in [0.0, 1.0] when all individual scores are in [0.0, 1.0].

If `WithNormaliseWeights(false)`, weights are used as-is. The `Score` method computes `Σ wᵢ · simᵢ`. The composite is NOT guaranteed to be in [0.0, 1.0] — it is the consumer's responsibility to choose weights that keep the composite within bounds. The library does not clamp.

The default is `true` because in practice consumers think of weights as proportions and expect a normalised composite. The opt-out is provided for advanced consumers who want raw weighted sums (e.g. for monotonic threshold comparison without normalisation overhead).

### 8.5 Defaults

`DefaultScorer()` returns a Scorer pre-configured with a sensible mixed-algorithm composite suitable for identifier-style string comparison:

```go
func DefaultScorer() *Scorer  // returns a pre-configured Scorer; cannot fail
```

Default composition:

- Damerau-Levenshtein (OSA): weight 0.30
- Jaro-Winkler: weight 0.20
- Token Jaccard: weight 0.20
- Q-Gram Jaccard (n=3): weight 0.15
- Sørensen-Dice (n=2): weight 0.10
- Double Metaphone: weight 0.05
- Threshold: 0.85
- Normalisation: `DefaultNormalisationOptions()`

These defaults are calibrated for the originating audit-field-similarity use case. They are documented as a reasonable starting point, not a one-size-fits-all configuration. Consumers tuning for their domain are expected to construct their own Scorer.

### 8.6 ScoreAll behaviour

`ScoreAll` returns a freshly-allocated `map[AlgoID]float64` per call (SPEC OVERRIDE — originally `map[string]float64`, see §8.3 above and 08-CONTEXT.md §1 for the typed-enum-keys rationale). Keys are the typed `AlgoID` enum values; consumers needing the CamelCase display form call `AlgoID.String()` (see §6 Algorithm identifiers for the locked naming convention). Map iteration order in Go is non-deterministic — this is INTENTIONAL: consumers reading the map for human display can sort keys themselves; consumers using the map programmatically don't care about order. The Scorer's internal computation does NOT depend on map iteration order (algorithms are iterated in AlgoID-sorted order internally).

---

## 9. Normalisation Pipeline

Pre-comparison normalisation is exposed both as a standalone function (`Normalise`) and as Scorer configuration (`WithNormalisation`).

```go
type NormalisationOptions struct {
    Lowercase       bool
    StripSeparators bool
    SeparatorChars  string  // default "_-.:/"
    SplitCamelCase  bool    // affects Tokenise only
}

func DefaultNormalisationOptions() NormalisationOptions {
    return NormalisationOptions{
        Lowercase:       true,
        StripSeparators: true,
        SeparatorChars:  "_-.:/",
        SplitCamelCase:  true,
    }
}

func Normalise(s string, opts NormalisationOptions) string
```

### 9.1 Operations

Applied in order:

1. **Lowercase** (if `Lowercase`): `unicode.ToLower` over runes. ASCII fast path: bitwise OR with `0x20` when `'A' ≤ c ≤ 'Z'`.
2. **Strip separators** (if `StripSeparators`): bytes (or runes) in `SeparatorChars` removed from the string.
3. **Tokenisation** (CamelCase split, see section 10): applied only when the result is used by a token-based algorithm or when `Tokenise` is called directly.

`Normalise` returns the post-step-2 string. Token-based algorithms call `Tokenise` (which applies all three steps within each token) instead of `Normalise` directly.

### 9.2 ASCII fast path

For ASCII-only inputs (verified by a single pass scanning for any byte ≥ 0x80), the implementation uses a stack-allocated `[64]byte` buffer for inputs ≤ 64 bytes (zero heap allocation) and a single-pass byte-level transformation. Non-ASCII inputs fall through to the rune-aware path with heap allocation.

### 9.3 Default behaviour

The `DefaultNormalisationOptions` configuration treats the following as equivalent for similarity purposes:

- `user_id`, `userId`, `user-id`, `User.Id`, `USER_ID` → all normalise (after lowercase + separator strip) to `userid`
- `XMLParser`, `xml_parser`, `xml-parser` → all tokenise (with CamelCase split) to `["xml", "parser"]`

### 9.4 Disabling normalisation

Pass `NormalisationOptions{}` (zero value) to disable all normalisation, or call `Scorer.WithoutNormalisation()`. Algorithms then operate on raw input.

---

## 10. Tokenisation

```go
func Tokenise(s string, opts NormalisationOptions) []string
```

Tokenisation rules (applied in order):

1. **Lowercase** (if `opts.Lowercase`) — applied to each character before tokenisation
2. **Split on separator characters** (if `opts.StripSeparators`) — characters in `SeparatorChars` are token boundaries
3. **Split on camelCase / PascalCase / acronym boundaries** (if `opts.SplitCamelCase`):
   - Insert a boundary at every uppercase letter that follows a lowercase letter (`userID` → `user`, `ID`)
   - Insert a boundary at every lowercase letter that follows two or more uppercase letters (`XMLParser` → `XML`, `Parser`)
4. **Filter empty tokens** — consecutive separators do not produce empty tokens
5. **Lowercase post-split** (if `opts.Lowercase`) — applied after tokenisation so that uppercase-letter-driven boundaries are preserved

Tokenisation examples:

- `UserCreateEvent` → `["user", "create", "event"]`
- `user_create_event` → `["user", "create", "event"]`
- `HTTP_REQUEST_V2` → `["http", "request", "v2"]`
- `httpRequestBody` → `["http", "request", "body"]`
- `XMLParser` → `["xml", "parser"]`
- `User_ID` → `["user", "id"]`

The returned slice is a fresh allocation per call.

**ASCII fast path (Phase 8.5 Q8b).** When the input is verified ASCII-only (no byte ≥ 0x80) and `opts.Lowercase == false`, `Tokenise` returns substrings of the input via `s[lo:hi]` zero-copy slicing. The returned token strings share the input's backing storage; the consumer must not assume independent lifetimes. When `opts.Lowercase == true`, an ASCII-letter sub-pass writes lowercased bytes into a single contiguous scratch buffer, and tokens are returned as substrings of that buffer — one allocation total for the buffer, zero allocations per token. The rune-aware path (non-ASCII input) continues to allocate per-token strings as before. The token-tier algorithms (Token Sort Ratio, Token Set Ratio, Partial Ratio, Token Jaccard, Monge-Elkan) consume `Tokenise` output and inherit the fast-path savings on ASCII identifier input. The fast path is BDD-tested and benchmark-asserted; the contract guarantees output equivalence with the rune-aware path on ASCII inputs.

---

## 11. Phonetic Algorithm Integration

Phonetic algorithms (Soundex, Double Metaphone, NYSIIS, MRA) are inherently boolean: two strings either share an encoded key or they don't. Their Scorer behaviour is:

- Apply normalisation to both inputs (the same way other algorithms do)
- Compute each input's phonetic code(s)
- Return 1.0 if the codes match per the algorithm's matching rule (exact match for Soundex and NYSIIS; either-of-two-keys match for Double Metaphone; MRA's length-dependent threshold rule for MRA), else 0.0

For consumers wanting richer behaviour, the public functions expose the keys directly:

```go
code := fuzzymatch.SoundexCode("Smith")
primary, secondary := fuzzymatch.DoubleMetaphoneKeys("Schmidt")
nysiis := fuzzymatch.NYSIISCode("Robert")
mra := fuzzymatch.MRACode("Brown")
```

Consumers wanting partial-match scoring on phonetic codes (e.g. Levenshtein distance on Soundex codes for near-match) compose this themselves:

```go
distance := fuzzymatch.LevenshteinDistance(
    fuzzymatch.SoundexCode(a),
    fuzzymatch.SoundexCode(b),
)
```

This is documented in `docs/algorithms.md` under "Composing phonetic algorithms with edit distance."

---

## 11.5 Input Validation and Diagnostics

The library ships a consumer-facing diagnostic function `Validate(a, b string) []Warning` that reports problematic-but-non-fatal input shapes **before** scoring runs. This is the recommended companion to any Scorer or direct algorithm call in code paths where the inputs originate from untrusted sources (user submissions, scraped data, parsed configuration). Added in Phase 8.5 Q4.

`Validate` returns warnings, not errors. The comparison-data leniency contract (§6.A) means algorithms always produce a value on any input; `Validate` is the diagnostic that tells consumers whether the value they got is meaningful. The function is fast (single pass over each input plus a `Normalise` + `Tokenise` dry run), allocation-light, and pure.

### Public API

```go
// Validate inspects the two inputs and returns warnings describing
// problematic-but-non-fatal input shapes. Returns nil if no warnings
// apply. Safe for concurrent use; never panics; never returns an error.
func Validate(a, b string) []Warning

// Warning describes one input-quality concern.
type Warning struct {
    Algorithm AlgoID  // which algorithm the warning is relevant to (0 if not algorithm-specific)
    Kind      WarnKind
    Detail    string  // human-readable description with input-shape detail
}

// WarnKind classifies the kind of input-quality concern.
type WarnKind int

const (
    WarnEmptyInput              WarnKind = iota + 1
    WarnUnequalLength           // Hamming family — silent-max policy applies
    WarnNoTokensAfterNormalise  // token-tier algorithms — would compare empty token sets
    WarnAllNonASCIIDropped      // ASCII-only algorithms — input collapsed to empty after stripping
    WarnPathologicallyLargeInput // O(m·n) DP risk; recommended to gate at the call site
)

// String returns the CamelCase form ("WarnEmptyInput", "WarnUnequalLength",
// etc.), matching the AlgoID.String() naming convention (§6).
func (k WarnKind) String() string
```

### Recommended usage pattern

```go
warnings := fuzzymatch.Validate(a, b)
if len(warnings) > 0 {
    // surface to logs, telemetry, audit trail, or reject input upstream
    for _, w := range warnings {
        log.Printf("input quality warning: %s (%s): %s",
            w.Kind, w.Algorithm, w.Detail)
    }
}
score := fuzzymatch.DefaultScorer().Score(a, b)
```

The validate-then-score pattern is the recommended idiom for code paths that audit input quality. Consumers who do not run `Validate` get the lenient algorithm contract (no panic, no error) but lose visibility into degraded-input scores (e.g. two empty strings scoring 1.0 because both encode to the empty Soundex code).

### Per-WarnKind semantics

- **`WarnEmptyInput`** — `a == ""` or `b == ""` (or both). Affects every algorithm: identity short-circuits, Hamming returns trivial values, phonetic algorithms produce empty keys that match each other.
- **`WarnUnequalLength`** — `len(a) != len(b)` and at least one of the Hamming family is relevant (the warning carries `Algorithm = AlgoHamming`). Documents the silent-max policy explicitly to the caller.
- **`WarnNoTokensAfterNormalise`** — after applying `DefaultNormalisationOptions` and `Tokenise`, one or both inputs produce an empty token list. Affects token-tier algorithms.
- **`WarnAllNonASCIIDropped`** — input contains characters but is entirely non-ASCII (or becomes empty after ASCII-only normalisation steps). Affects ASCII-only algorithms (Strcmp95, Soundex, Double Metaphone, NYSIIS, MRA — see §7.4 and §7.1.7).
- **`WarnPathologicallyLargeInput`** — input length exceeds an algorithm-specific threshold defined as a documented per-algorithm constant in the implementation. Typical thresholds range from 4 KB to 64 KB per side depending on algorithmic complexity (O(m·n) algorithms like Damerau-Levenshtein Full, Smith-Waterman-Gotoh, Ratcliff-Obershelp, Monge-Elkan, and Partial Ratio trip the warning at the lower end of the range; O(n+m) phonetic and Hamming-family algorithms at the higher end). The specific thresholds are tuned during Phase 8.5 implementation and pinned in `docs/algorithms.md` per-algorithm.

### Cross-references

`Validate` is documented prominently across all six required surfaces per `.claude/skills/documentation-standards/SKILL.md` § Consumer-facing validation and diagnostics features — README Quick Start / Common Patterns, `docs/algorithms.md` (and/or `docs/best-practices.md`), per-algorithm godoc cross-references, `llms.txt` + `llms-full.txt`, the user-guide section, and at least one runnable `examples/` program. Missing any surface is treated as a Critical documentation gap by `user-guide-reviewer` and `docs-writer`.

### Forward compatibility

The `WarnKind` enum may grow in v1.x with additional constants. Consumers must treat unrecognised values as ignorable (the `Detail` field carries human-readable context). The existing constants are stable across patch versions.

---

## 12. Layer 3 — Scan sub-package

> *Reminder: the `Item`, `Config`, `Warning`, and `Check` shapes below are illustrative. The `api-ergonomics-reviewer` and `user-guide-reviewer` agents determine the final API. Scan semantics (within-vs-cross-group passes, suppression composition, deterministic output, token-bucket optimisation equivalent to naive) are requirements.*

The `github.com/axonops/fuzzymatch/scan` sub-package provides a turnkey collection-scan layer on top of the root algorithms and Scorer. It detects pairs of similar names in a collection, with optional grouping semantics, suppression mechanisms, and a token-bucket optimisation for large collections.

The scan sub-package is **optional**. Consumers wanting only algorithm functions or only the Scorer never import `scan` and incur no cost from its existence — the root package has no dependency on the scan sub-package. The scan sub-package depends on the root package (it consumes `*fuzzymatch.Scorer`).

### 12.1 Public API

```go
package scan

import "github.com/axonops/fuzzymatch"

// Item is one named thing the scanner compares. Consumers construct
// items from whatever schema or vocabulary they care about.
type Item struct {
    // Name is the value being compared. Required, non-empty.
    Name string

    // Group scopes the comparison. Items with the same Group are
    // compared against each other in the within-group pass. Items
    // with different Group values are compared in the cross-group
    // pass when Config.CompareAcrossGroups is true. The empty
    // string is a valid Group (one global group).
    Group string

    // SilenceLint, when true, suppresses any warning involving this
    // Item (one-side suppression: flag on either side of a pair
    // silences the pair).
    SilenceLint bool

    // Tag is opaque consumer data. The library does not interpret
    // it; it appears unchanged in the Warning.TagA / Warning.TagB
    // fields. Use it to carry source-file line numbers, schema
    // paths, or any other context useful in error messages.
    Tag any
}

// SPEC OVERRIDE (Phase 9): type renamed to Kind per 09-CONTEXT.md §1
// D-02 (the previous draft used a verbose two-word name that doubled
// the noise of WarnKind in the root package). The package-scoped form
// (scan.KindWithinGroup, scan.KindAcrossGroups) is unambiguous at the
// call site and avoids accidental symmetry with the root package's
// WarnKind type (which classifies Validate diagnostics, a different
// domain). api-ergonomics-reviewer signed off on this override in
// plan 09-01's PR.

// Kind classifies a similarity warning.
type Kind int

const (
    // KindWithinGroup signals a similar-name pair where both items
    // have the same Group.
    KindWithinGroup Kind = iota + 1

    // KindAcrossGroups signals a similar-name pair where the items
    // have different Group values.
    KindAcrossGroups
)

// String returns the CamelCase form ("WithinGroup", "AcrossGroups"),
// matching the AlgoID.String() and WarnKind.String() naming convention
// locked in §6 Algorithm identifiers and Phase 8.5 Q6b.
func (k Kind) String() string

// Warning is one detected similar-name pair. NameA is the
// lexicographically smaller of the two raw item Names under
// Go-native string-byte comparison (raw-byte lex). SPEC OVERRIDE
// (Phase 9): the originally-specified "(after normalisation)"
// tiebreaker is unnecessary because Phase 9 D-06 validates that
// (Name, Group) pairs are unique at Check entry, guaranteeing the
// (Kind, NameA, NameB, GroupA, GroupB) sort key is a strict total
// order on every valid input. Plan 09-06's implementation uses
// raw-byte lex on the RAW (non-normalised) Names so consumers see
// their input Names in the output verbatim. See 09-CONTEXT.md §3
// D-06 and the api-ergonomics-reviewer sign-off recorded in plan
// 09-06's PR.
type Warning struct {
    // SPEC OVERRIDE (Phase 9): Kind field type renamed per 09-CONTEXT.md
    // §1 D-02; see the Kind declaration above.
    Kind           Kind
    NameA, NameB   string
    GroupA, GroupB string
    TagA, TagB     any

    // Score is the composite score returned by the configured Scorer.
    Score float64

    // SPEC OVERRIDE (Phase 9): Scores is map[AlgoID]float64 (typed
    // enum keys) per 09-CONTEXT.md §1 D-01. Extends the Phase 8
    // ScoreAll override at §8.3 + §8.6 for the same typed-enum-keys
    // rationale: the rest of the library exposes AlgoID and gives
    // consumers compile-time key safety. Use AlgoID.String() for
    // CamelCase display. api-ergonomics-reviewer signed off on this
    // override in plan 09-01's PR.
    Scores map[AlgoID]float64
}

// Config controls Check behaviour. Construct directly; the zero
// value is invalid (Scorer is required) and Check returns
// ErrNilScorer.
type Config struct {
    // Scorer is required. It governs both the similarity computation
    // and the emission threshold (warnings are emitted when
    // Scorer.Match returns true, with the cross-group boost applied
    // where relevant).
    Scorer *fuzzymatch.Scorer

    // CompareAcrossGroups enables the cross-group pass: items with
    // different Group values are compared against each other. When
    // false (default), only same-group pairs are compared.
    CompareAcrossGroups bool

    // CrossGroupThresholdBoost is added to the Scorer's threshold
    // when evaluating cross-group pairs. The cross-group pass is
    // inherently noisier than the within-group pass; a small
    // positive value (0.02–0.05) reduces false positives without
    // disabling the pass. Range [0.0, 1.0].
    //
    // Zero-value is 0.0 (no boost). The opinionated default 0.05 is
    // supplied by `scan.DefaultConfig(s *fuzzymatch.Scorer) Config`.
    // SPEC OVERRIDE (Phase 9): default-0.05 location migrated from
    // this field godoc to `DefaultConfig` per 09-CONTEXT.md §2 D-04,
    // mirroring Phase 8's DefaultScorer / DefaultScorerOptions pattern
    // (the zero-value of the struct is a valid minimal Config; the
    // opinionated helper bakes in the experience-tuned values).
    CrossGroupThresholdBoost float64

    // CompareIdenticalAcrossGroups, when false (default), suppresses
    // cross-group warnings for pairs whose names are byte-identical
    // after normalisation. Operators legitimately reuse the same
    // name (e.g. user_id) across groups; surfacing every such pair
    // would drown real similar-but-not-equal signals.
    CompareIdenticalAcrossGroups bool

    // SuppressedPairs is a list of name pairs that should not
    // produce a warning, in addition to per-Item SilenceLint flags.
    // Pairs are order-independent and normalised (using the
    // Scorer's normalisation options) before matching.
    SuppressedPairs [][2]string
}

// Check compares every pair of items per the Config and returns
// warnings for pairs where the Scorer's Match returns true (with
// the cross-group threshold boost applied where relevant).
//
// Output is sorted deterministically by (Kind, NameA, NameB,
// GroupA, GroupB) and is byte-identical across runs for the same
// input and Scorer configuration.
//
// Check is a pure function: it never reads files, environment
// variables, or any package-global state. The function is safe for
// concurrent use with itself and with any Scorer method.
//
// Returns ErrNilScorer if cfg.Scorer is nil. Returns ErrInvalidItem
// wrapped with every offending index, joined via errors.Join — the
// caller's `errors.Is(err, scan.ErrInvalidItem)` discriminates because
// errors.Is walks Unwrap() []error (Go 1.20+). Validation rules:
// Item.Name == "" (D-03) and duplicate (Name, Group) (D-06) — see
// 09-CONTEXT.md §2 D-03 and §3 D-06 for the SPEC OVERRIDE rationale
// (the spec originally specified single-offending-index fail-fast
// wording; collect-all via errors.Join lets the caller fix the whole
// batch in one round-trip). Returns ErrInvalidConfig wrapped with the
// specific problem if cfg is otherwise malformed (D-04: NaN / ±Inf /
// out-of-range CrossGroupThresholdBoost; D-05: empty SuppressedPairs
// entries, also joined via errors.Join). Empty items slice returns an
// empty Warning slice and no error.
func Check(items []Item, cfg Config) ([]Warning, error)
```

Sentinel errors:

```go
var (
    ErrNilScorer     = errors.New("scan: Config.Scorer is required")
    ErrInvalidItem   = errors.New("scan: invalid item")
    ErrInvalidConfig = errors.New("scan: invalid config")
)
```

### 12.2 Within-group vs cross-group passes

The default pass is **within-group only**: items are compared only against other items sharing the same Group value. The within-group pass uses the Scorer's configured threshold directly.

When `CompareAcrossGroups` is true, an additional **cross-group pass** runs: items in different Groups are compared. The effective threshold for the cross-group pass is `scorer.Threshold() + cfg.CrossGroupThresholdBoost`. This reflects that cross-group similarities are inherently noisier — operators often legitimately reuse names like `user_id` across event groups — and need a tighter threshold to be informative.

If `cfg.CrossGroupThresholdBoost` would push the effective threshold above 1.0, the threshold is clamped to 1.0 (meaning only byte-identical matches pass; combined with `CompareIdenticalAcrossGroups: false` this disables cross-group emission, which is documented).

### 12.3 Suppression composition

Two mechanisms compose. A warning is suppressed if ANY of the following applies:

- Either item in the pair has `SilenceLint = true` (one-side suppression: setting the flag on either side silences the pair)
- The pair, normalised and sorted, appears in `cfg.SuppressedPairs`
- The pair is `KindAcrossGroups` AND both names are byte-identical after normalisation AND `cfg.CompareIdenticalAcrossGroups` is false

`SuppressedPairs` entries are normalised at the start of `Check` using the Scorer's normalisation options. Lookup is O(1) per candidate pair via an internal sorted-pair set built once at the top of `Check`.

The library treats `SilenceLint` as opaque. Consumers attach their own meaning (for example, the `axonops/audit` consumer populates it from a `silence_lint: true` YAML key on field declarations).

### 12.4 Determinism

Scan output is deterministic. Two `Check` calls on the same input and same Scorer configuration produce byte-identical output. Output is sorted by `(Kind, NameA, NameB, GroupA, GroupB)`. Map iteration is never exposed on the output path. The internal token-bucket map is iterated only to construct candidate sets, which are sorted before the scoring loop.

Property test `PropCheck_DeterministicAcrossRuns` (in `scan/props_test.go`) verifies byte-identical output across runs on randomised input. The cross-platform determinism CI matrix (section 13.3) exercises the scan sub-package on every supported platform.

### 12.5 Token-bucket optimisation

Naive cross-group comparison is O(N²). For 10,000 items this is 100 million pair comparisons. The scan sub-package uses a token-bucket index to reduce this:

1. Tokenise every name once at the start of `Check`, using the Scorer's normalisation options
2. Build `map[string][]int` from token → item indices
3. For each item, candidate partners are the union of `bucket[t]` for each token `t` in that item. Items sharing no token are skipped entirely (they cannot exceed any reasonable similarity threshold)
4. Within the candidate set, invoke the Scorer's `Match` (or `Score` followed by threshold comparison)

Correctness: property test `PropCheck_BucketEquivalentToNaive` verifies the bucket-optimised result equals the naive all-pairs result on randomised input. The optimisation never causes a real warning to be dropped because any pair with similarity above ~0.5 shares at least one token after normalisation.

Within-group pass uses the same bucket when group size exceeds 50 items (default); below this threshold, naive nested loops are faster due to constant factors.

### 12.6 Performance

Targets (with the default Scorer of section 8.5):

- 200 items / 10 groups: < 10 ms
- 1000 items / 50 groups: < 100 ms
- 10000 items / 500 groups: < 2 s
- Cross-group pass enabled: at most 2× the within-group-only cost on the same input
  - **v0.x shortfall:** the per-pair token-bucket rebuild strategy currently measures ~525× within-only at 10k items / 500 groups (~189 s vs ~362 ms; verified by `BenchmarkScanCheck_DefaultScorer_10k_CrossGroup`). The ≤ 2× target requires the global-bucket strategy tracked as a v1.x optimisation in `.planning/phases/09-collection-scan-sub-package/09-CONTEXT.md` Deferred Ideas. SCAN-02 bucket-vs-naive correctness equivalence is preserved at v0.x; the gap is purely performance.

Benchmarks in `scan/scan_bench_test.go` cover within-group only, within-plus-cross-group, and the bucket-vs-naive comparison. CI regression detection via `benchstat` against the last tagged release.

### 12.7 Repository layout

```text
github.com/axonops/fuzzymatch/scan/
├── scan.go                       # Item, Config, Warning, Check
├── scan_test.go                  # external/black-box unit tests
├── scan_internal_test.go         # bucket optimisation correctness
├── scan_bench_test.go            # benchmarks
├── bucket.go                     # token-bucket implementation
├── errors.go                     # sentinel errors
├── doc.go                        # package documentation
├── example_test.go               # godoc runnable examples
├── props_test.go                 # property tests
└── fuzz_test.go                  # native fuzz harness for Check
```

The scan sub-package has its own `doc.go` and `example_test.go`. It does NOT have its own `go.mod` — it is a sub-package of `github.com/axonops/fuzzymatch` under the same module.

---

## 13. Determinism Guarantees

The library guarantees the following determinism properties. Verified by tests and CI.

### 13.1 Algorithm score stability

Algorithm scores are stable to the last bit across patch versions within a major version. Score from any public `*Score` function in v1.0.0 must equal the same call in v1.5.0 to the last bit. A golden-file test in `testdata/golden/` pins reference vectors per algorithm; any change to the algorithm output requires updating the golden file with an accompanying CHANGELOG entry and a minor version bump.

### 13.2 Scorer composite stability

`Scorer.Score(a, b)` is deterministic for the same Scorer configuration and the same input. Two calls return byte-identical (to the last bit) results.

### 13.3 Cross-platform determinism

Verified by CI matrix: linux/amd64, linux/arm64, darwin/amd64, darwin/arm64, windows/amd64. The cross-platform test runs every algorithm and every default-configured Scorer on a fixed input corpus and asserts byte-identical output across platforms.

This is achievable because: (a) all float operations are simple enough (addition, multiplication, division) that IEEE-754 reproducibility holds across the supported platforms; (b) no platform-specific intrinsics are used; (c) no goroutine scheduling affects output; (d) no map iteration affects output; (e) no time-dependent or environment-dependent values participate.

**FMA-defeating double-cast (LOCKED Phase 8.5 Q11b).** Two reduction sites in the library — the Cosine numerator accumulator at `cosine.go:343` and the Scorer composite accumulator at `scorer.go:380` — apply a preemptive `float64(float64(x*y)) + acc` double-cast to defeat compiler-emitted fused multiply-add (FMA) instructions. FMA produces a single rounding step `fma(x, y, acc) = round(x·y + acc)`, whereas the explicit form produces two roundings `round(round(x·y) + acc)`. Go's compiler may emit FMA on arm64 (where the `fmadd` instruction is cheap) but typically does not on amd64 (where FMA was historically optional). The two forms diverge by 1 ULP on inputs that cross the rounding boundary. Pre-emptive remediation costs roughly 1 ns per reduction and guarantees byte-identical cross-platform output for all inputs, including future inputs that happen to cross the ULP threshold. The pattern is documented inline at both sites and is verified by the cross-platform CI matrix golden file.

### 13.4 Map iteration

Map iteration order in Go is non-deterministic by design. The library uses `map` internally (e.g. for q-gram counting, token-set intersection) but NEVER exposes map iteration order to public output. Internal algorithms that need to iterate a map in a stable order extract keys, sort them, and iterate the sorted slice.

`Scorer.ScoreAll` returns a `map[AlgoID]float64` (typed enum keys per §8.3 + §8.6 SPEC OVERRIDE) — the map itself is returned, and iteration order is non-deterministic. This is documented; consumers requiring stable order should sort the keys.

### 13.5 No init() side effects

The library has no `init()` functions doing non-trivial work. The Strcmp95 similarity-character table is initialised via a `var` declaration, not in `init()`.

### 13.6 Property tests

The following properties are verified by property-based tests via `testing/quick`:

- `PropAlgorithmScore_RangeBounds` — every algorithm's score is in [0.0, 1.0] for arbitrary inputs
- `PropAlgorithmScore_Identity` — `Score(x, x) = 1.0` for non-empty `x`
- `PropAlgorithmScore_Symmetric` — `Score(a, b) = Score(b, a)` for all symmetric algorithms (Monge-Elkan and asymmetric Tversky excluded)
- `PropEditDistance_TriangleInequality` — for distance-based algorithms: `D(a, c) ≤ D(a, b) + D(b, c)`
- `PropScorer_DeterministicAcrossRuns` — Scorer.Score is byte-identical across runs for the same input
- `PropNormalise_Idempotent` — `Normalise(Normalise(x)) == Normalise(x)`
- `PropTokenise_NoEmptyTokens` — `Tokenise` never returns empty tokens
- `PropAllPublic_NeverPanic` — arbitrary inputs (including invalid UTF-8) never panic any public function

---

## 14. Performance Budgets

Targets (single-core, modern x86_64, Go 1.26+):

### 14.1 Per-algorithm budgets

For ASCII inputs ≤ 50 characters:

- Levenshtein, Damerau-Levenshtein OSA, Hamming: < 1 µs per call, 0 allocations on Short; on Long inputs Levenshtein and DL-OSA transition from stack to heap above the documented stack-buffer threshold (see Q7c scope note below). Damerau-Levenshtein OSA Unicode Short adds one allocation versus the byte path due to rune-decode buffering.
- Damerau-Levenshtein Full: < 10 µs per call at Medium (50×50), ≤ 1 allocation on Short (Phase 8.5 Q8a — the stack-buffer optimisation that would achieve 0 allocs requires a 34 KB stack frame, judged too fragile). Medium latency was widened from < 3 µs to < 10 µs (Q7b) because the Lowrance-Wagner formulation needs the full O(m·n) DP table, not the two-row rolling form used by DL-OSA.
- Jaro, Jaro-Winkler, Strcmp95: < 1 µs per call, 0 allocations on Short; transition to heap on Long inputs above the documented stack-buffer threshold (Q7c).
- Smith-Waterman-Gotoh: < 10 µs per call at Medium (Phase 8.5 Q7b — widened from < 5 µs because the six-row affine-gap rolling DP at 50×50 has ~3000 float64 comparisons), 0 allocations on Short (stack buffer for ≤ 50-char inputs).
- LCSStr: < 2 µs per call, 0 allocations on Short; heap fallback on Long inputs (Q7c).
- Q-Gram Jaccard, Sørensen-Dice, Cosine, Tversky: < 5 µs per call, ≤ 4 allocations on Short (q-gram map + count map). The internal `extractQGrams` capacity hint is `(len(s)-n+1)*5/4` to reduce hash-map growth allocations (Q7d).
- Monge-Elkan: < 10 µs per call (cost dominated by inner metric × token-count squared)
- Token Sort Ratio, Token Set Ratio, Token Jaccard: < 5 µs per call, ≤ 4 allocations
- Partial Ratio: < 10 µs per call for short inputs (cost grows with the length difference)
- Soundex Code: < 500 ns per call, ≤ 1 allocation on Short (Phase 8.5 Q7b — the `string(stackBuf[:n])` return is structurally unavoidable without `unsafe.String`, which is forbidden by go-coding-standards).
- Soundex Score: < 1 µs per call, ≤ 2 allocations on Short (provisional pending Phase 8.5 benchmarks — two `SoundexCode` calls each contributing ≤ 1 alloc plus the boolean equality check).
- NYSIIS Code: < 500 ns per call, ≤ 1 allocation on Short (Q7b — same rationale as Soundex).
- NYSIIS Score: < 1 µs per call, ≤ 2 allocations on Short (provisional pending Phase 8.5 benchmarks — two `NYSIISCode` calls each contributing ≤ 1 alloc).
- MRA Code: < 500 ns per call, ≤ 1 allocation on Short (Q7b — same rationale).
- MRACompare: < 1 µs per call, ≤ 2 allocations on Short (Q7b — two MRACode calls produce two heap strings).
- MRA Score: < 1 µs per call, ≤ 2 allocations on Short (provisional pending Phase 8.5 benchmarks — wraps `MRACompare`, inherits the same allocation profile).
- Double Metaphone: < 2 µs per call, ≤ 2 allocations (post-Phase-8.5 Q7a `[4]byte` optimisation replacing `strings.Builder`; the residual 2 allocations are the primary + secondary return-string heap escapes).
- Ratcliff-Obershelp: < 5 µs per call for short inputs, ≤ 4 allocations on Short (Phase 8.5 Q8d — the `roFindLongestMatch` recursive decomposition allocates per recursion level; the stack-buffer pool that would achieve a tighter budget was judged too complex for the marginal gain).

For longer inputs (50–500 characters): each algorithm's cost scales per its complexity. Benchmarks at 50/200/500 chars. The Long-input column documents the heap-allocation path for Levenshtein, Jaro, JaroWinkler, Strcmp95, LCSStr (and DL-OSA Unicode-Short) where the stack-buffer optimisation does not apply (Q7c scope note).

### 14.2 Scorer budgets

- Default-configured Scorer (6 algorithms): `Score(a, b)` < 30 µs for ASCII inputs ≤ 50 chars
  - Short (ASCII ≤ 8 chars): ≤ 8 allocations (Phase 8.5 Q8c — the post-Q7a DoubleMetaphone optimisation brings the composite to exactly 8 allocs on Short)
  - Medium (ASCII ≤ 50 chars): ≤ 20 allocations (Phase 8.5 Q8c — composite of the per-algorithm Medium budgets including q-gram map rehash and tokeniser scratch)
- Default-configured Scorer: `ScoreAll(a, b)` adds one allocation (the result map)
- `Match` matches `Score` cost

### 14.3 Normalisation budgets

- `Normalise` ASCII input ≤ 50 chars: < 200 ns, ≤ 1 allocation (Phase 8.5 Q7b — the `string(buf)` return is structurally unavoidable; the scratch buffer remains stack-allocated and the prior "0 allocations" claim was incorrect)
- `Tokenise` ASCII input ≤ 50 chars: < 500 ns, ≤ 2 allocations (token slice + storage; the ASCII fast path returns substrings of the input or of a single scratch buffer — see §10 Tokenisation, Phase 8.5 Q8b)

### 14.4 Benchmark suite

Every algorithm has a benchmark file (`xxx_bench_test.go`) covering short / medium / long inputs. Benchmark output is committed to `bench.txt` per release. CI runs `benchstat` against the last tagged release; > 10% regression on any benchmark fails the build.

The benchmark CI job uses a labelled self-hosted runner shared with `axonops/mask` and `axonops/audit` for hardware consistency. If the runner is unavailable, benchmarks run informationally on `ubuntu-latest` and regression detection is skipped with a CI annotation.

---

## 15. Testing Strategy

### 15.1 Unit tests

External / black-box: `package fuzzymatch_test` in `xxx_test.go` files.

Every algorithm has a `xxx_test.go` file covering:

- Identity: `Score(x, x) = 1.0` for non-empty `x`
- Both-empty edge case
- One-empty edge case
- Literature reference vectors (canonical examples from the cited primary source)
- Algorithm-specific invariants (symmetry where applicable, range bounds, triangle inequality for distance-based)
- Unicode behaviour documented and tested where the algorithm has rune/byte variants

The Scorer has `scorer_test.go` covering:

- Construction success and every documented failure mode
- Score is in [0.0, 1.0] (when weights normalised)
- ScoreAll returns per-algorithm scores keyed by algorithm name
- Match returns true / false consistently with Score and Threshold
- Concurrent Score / ScoreAll / Match is safe
- DefaultScorer matches documented composition
- Re-adding the same algorithm overrides the prior weight (last-wins)

### 15.2 Internal tests

`package fuzzymatch` (not `fuzzymatch_test`) in `xxx_internal_test.go` files. Used for testing unexported correctness invariants — primarily the q-gram extraction, token-set construction, and Strcmp95 similar-character table.

### 15.3 Property tests

In `props_test.go`, using `testing/quick` (stdlib only):

- See section 13.6 for the property list.

**Generator distribution (LOCKED Phase 8.5 Q12b).** The `testing/quick` Generator implementations for input strings in `props_test.go` follow a mixed-shape distribution to exercise edge cases the default `rand.String` cannot reach:

- 30% ASCII-short (`len ∈ [0, 8]`) — exercises identity, near-empty, and tight stack-buffer paths
- 30% ASCII-medium (`len ∈ [9, 50]`) — exercises the canonical operating range
- 20% Unicode-short (`len ∈ [0, 8]` measured in runes; non-ASCII code points mixed with ASCII) — exercises rune-decode paths
- 10% Unicode-medium (`len ∈ [9, 50]` runes; non-ASCII mix) — exercises heap fallback for stack-buffer algorithms
- 10% adversarial — invalid UTF-8 byte sequences, very long runs of a single character, mixed surrogate pairs, byte-order-mark prefixes

Token-count generators (used by `Tokenise`-aware properties) compute counts via `fuzzymatch.Tokenise` rather than `strings.Fields` — the two diverge on camelCase / acronym boundaries and the test must match the library's actual tokenisation. `quick.MaxCount` defaults are raised to 1000 for cheap properties (allocations, range bounds) and held at the stdlib default (100) for expensive properties (triangle inequality, FMA-sensitive reductions).

### 15.4 Fuzz tests

Native Go fuzzing in `fuzz_test.go`:

- `FuzzLevenshtein`, `FuzzDamerauLevenshteinOSA`, `FuzzDamerauLevenshteinFull`, `FuzzHamming`, `FuzzJaro`, `FuzzJaroWinkler`, `FuzzStrcmp95`, `FuzzSmithWatermanGotoh`, `FuzzLCSStr`, `FuzzQGramJaccard`, `FuzzSorensenDice`, `FuzzCosine`, `FuzzTversky`, `FuzzMongeElkan`, `FuzzTokenSortRatio`, `FuzzTokenSetRatio`, `FuzzPartialRatio`, `FuzzTokenJaccard`, `FuzzSoundex`, `FuzzDoubleMetaphone`, `FuzzNYSIIS`, `FuzzMRA`, `FuzzRatcliffObershelp`
- `FuzzNormalise`, `FuzzTokenise`
- `FuzzScorer` — constructs Scorer with arbitrary algorithm subsets, asserts score in [0,1] for arbitrary inputs

Corpus checked into `testdata/fuzz/`. CI runs 60s per fuzzer per build.

### 15.5 Benchmark tests

See section 14. Every algorithm and the Scorer have benchmark files. CI runs benchmarks against the last tagged release via `benchstat`.

### 15.6 BDD scenarios

`tests/bdd/` is a separate Go module with its own `go.mod` so consumers do not transitively depend on `cucumber/godog`. Feature files in `tests/bdd/features/`:

- `features/algorithms.feature` — per-algorithm scenarios with canonical input / score pairs, one Scenario Outline per algorithm
- `features/scorer.feature` — Scorer composition, weight normalisation, threshold matching, ScoreAll output structure
- `features/normalisation.feature` — lowercase / separator-strip / camelCase-split combinations
- `features/determinism.feature` — Scorer.Score byte-identical across repeated calls; cross-platform determinism via golden file
- `features/scan.feature` — scan sub-package: Item / Warning / Check, within-vs-cross-group, deterministic output ordering
- `features/suppression.feature` — scan sub-package suppression: per-item SilenceLint flag, SuppressedPairs list, composition between the two

Sample (algorithms.feature, abridged):

```gherkin
Feature: Similarity algorithm scores

  Scenario Outline: Levenshtein canonical reference vectors
    When I compute the Levenshtein score between "<a>" and "<b>"
    Then the score should be approximately <score>

    Examples:
      | a       | b       | score   |
      | kitten  | sitting | 0.5714  |
      | abc     | abc     | 1.0000  |
      |         |         | 1.0000  |
      | abc     |         | 0.0000  |

  Scenario Outline: Jaro-Winkler canonical reference vectors
    When I compute the Jaro-Winkler score between "<a>" and "<b>"
    Then the score should be approximately <score>

    Examples:
      | a      | b      | score   |
      | MARTHA | MARHTA | 0.9611  |
      | DWAYNE | DUANE  | 0.8400  |
```

Sample (scorer.feature, abridged):

```gherkin
Feature: Scorer composition and matching

  Scenario: Default Scorer matches similar identifiers
    Given the default Scorer
    When I compute the score between "user_id" and "userId"
    Then the score should be at least 0.85
    And Match should return true

  Scenario: Custom Scorer with single algorithm
    Given a Scorer configured with Levenshtein at weight 1.0 and threshold 0.5
    When I compute the score between "kitten" and "sitting"
    Then the score should be approximately 0.5714
    And Match should return true
```

### 15.7 Meta-tests

Root-level test files verifying project-level invariants:

- `documentation_test.go` — verifies every code block in `docs/*.md` parses as valid Go; verifies README quick-start compiles; verifies `docs/algorithms.md` reference vectors match the implementation's documented constants and unit-test reference vectors.
- `ai_friendly_test.go` — verifies `llms.txt` and `llms-full.txt` exist at repo root; verifies their "Public API" sections list every exported symbol (parsed via `go/ast`); verifies documented Scorer defaults match `DefaultScorer()` actuals.
- `readme_shop_front_test.go` — compiles and runs the README's headline example; verifies output exactly matches the documented expected output.
- `makefile_targets_test.go` — parses Makefile and verifies every target documented in `CONTRIBUTING.md` exists; conversely every target in Makefile is mentioned in `CONTRIBUTING.md`, `README.md`, or has a `## suppress: <reason>` comment.
- `internal_coverage_test.go` — mirroring `axonops/mask`. Enforces a coverage floor: any drop below the configured threshold (95% overall, 90% per file, 100% on public API) fails the test. Run as part of normal `go test`, not behind a build tag.

### 15.8 Coverage targets

- Overall (Floor 1): ≥ 95% line coverage
- Per file (Floor 2): ≥ 90% line coverage
- Per exported function (Floor 3): ≥ 90% statement coverage (Phase 8.5 Q12a — tightened from "exists-at-least-one-test" to a true statement-coverage floor, matching Floor 2's semantics. `scripts/verify-coverage-floors.sh` runs all three floors; Floor 3 uses AST-based detection of exported symbols rather than `go doc -short` parsing).
- CI fails if any threshold is missed
- Codecov upload for trend visibility

### 15.9 Goroutine leak detection

`go.uber.org/goleak` in `TestMain`:

```go
func TestMain(m *testing.M) {
    goleak.VerifyTestMain(m)
}
```

The library is pure-function and uses no goroutines. The test catches regressions if anyone introduces background work.

### 15.10 Race detector

CI runs all tests with `go test -race -count=1 ./...`. Required to pass.

### 15.11 Test dependencies

`testify`, `cucumber/godog`, `go.uber.org/goleak`. All in `tests/bdd/go.mod` (BDD) or as `_test.go`-only imports via Go's `testdata`-style isolation. The root `go.mod` retains zero non-stdlib `require` lines.

Specifically: property tests use `testing/quick` (stdlib). Fuzz tests use native `go test -fuzz`. Unit tests use `testify` — and this introduces a non-stdlib test-only dependency in the root module. **Resolution:** unit tests use stdlib `testing` package without `testify`. `testify` is permitted ONLY in `tests/bdd/`. This is stricter than the original mask reference (which uses `testify` in root tests) and is the deliberate choice for `fuzzymatch` to keep the root `go.mod` clean even for `_test.go` files.

Trade-off: tests are more verbose without `testify/assert`. Acceptable for the cleanliness benefit. To be revisited at v1.x if the verbosity becomes a maintenance problem.

### 15.12 Cross-validation corpora

Cross-validation against external reference implementations is the load-bearing acceptance test for algorithms where re-deriving correctness from the primary source alone is insufficient or fragile. Corpora are generated by a single-purpose Python script (one per tier), pinned by exact dependency version, committed to `testdata/cross-validation/<tier>/vectors.json`, and consumed by a Go loader test that diffs the implementation output against the corpus under a tolerance of `1e-9`.

Phase 8.5 Q10 locked the corpus inventory:

| Tier | Reference | Pin | Corpus file | Loader test |
|---|---|---|---|---|
| Smith-Waterman-Gotoh (§7.1.8) | `biopython` `Bio.Align.PairwiseAligner` mode=`"local"` | `biopython==1.84` | `testdata/cross-validation/swg/vectors.json` | `TestSWG_CrossValidation` |
| Ratcliff-Obershelp (§7.5.1) | Python stdlib `difflib.SequenceMatcher(autojunk=False).ratio()` | stdlib (no pin) | `testdata/cross-validation/ratcliff-obershelp/vectors.json` | `TestRatcliffObershelp_CrossValidation` |
| Token-tier ratios (§7.3.2–7.3.4) | `rapidfuzz` `fuzz.token_sort_ratio` / `fuzz.token_set_ratio` / `fuzz.partial_ratio` | `rapidfuzz==3.x` (exact pin in script) | `testdata/cross-validation/token-ratios/vectors.json` | `TestTokenRatios_CrossValidation` |
| Phonetic tier (§7.4) | `jellyfish` and `Metaphone` dual pin | `jellyfish==1.2.1` + `Metaphone==0.6` | `testdata/cross-validation/phonetic/vectors.json` | `TestPhonetic_CrossValidation` |
| Character tier (Levenshtein, DL-OSA, Hamming, Jaro, JaroWinkler) | `jellyfish` (reuse of phonetic pin) | `jellyfish==1.2.1` | `testdata/cross-validation/character/vectors.json` | `TestCharacter_CrossValidation` |
| Q-gram tier (Q-Gram Jaccard, Sørensen-Dice, Cosine, Tversky) and Monge-Elkan | `py_stringmatching` | `py_stringmatching==0.4.x` (exact pin in script) | `testdata/cross-validation/qgram/vectors.json`; `testdata/cross-validation/monge-elkan/vectors.json` | `TestQGram_CrossValidation`; `TestMongeElkan_CrossValidation` |

**Literature-only algorithms** — Strcmp95, Damerau-Levenshtein Full, and LCSStr have no documented external reference library that matches the library's exact formulation. Their test files carry an explicit block comment "No external cross-validation library; literature-anchored only" with a citation to the primary academic source. Reference vectors in their unit tests are derived directly from the published examples in the source paper.

**Regeneration is developer-only.** Each tier has a `make regen-<tier>-cross-validation` Makefile target that invokes the script with the exact dependency pin. CI consumes the committed `vectors.json` without requiring Python. Adding a new vector requires the script + regeneration + a code-review checklist item confirming the new vector was inspected before commit.

**Generators:**

- `scripts/gen-swg-cross-validation.py` (Phase 3)
- `scripts/gen-ratcliff-obershelp-cross-validation.py` (Phase 4)
- `scripts/gen-token-ratio-cross-validation.py` (Phase 6)
- `scripts/gen-phonetic-cross-validation.py` (Phase 7)
- `scripts/gen-character-cross-validation.py` (Phase 8.5 — new)
- `scripts/gen-qgram-cross-validation.py` (Phase 8.5 — new)
- `scripts/gen-monge-elkan-cross-validation.py` (Phase 8.5 — new)

---

## 16. Documentation Requirements

### 16.1 README

Structure mirroring `axonops/mask`:

1. Logo image (`.github/images/logo-readme.png`)
2. Title and tagline ("Fuzzy name matching for Go services — string similarity, weighted composite scoring, zero runtime dependencies")
3. Badges: CI, Go Reference, Go Report Card, License, Status (pre-release orange)
4. Quick links (Quick Start, Features, Algorithms, Docs, API Reference)
5. Table of contents
6. ⚠ Status section (pre-release framing)
7. 🔍 Overview — one paragraph plus quote-block tagline
8. ✨ Key Features bullet list
9. ❓ Why fuzzymatch?
10. 🚀 Quick Start with runnable example (DefaultScorer + custom Scorer)
11. 📚 Algorithm catalogue (table linking to `docs/algorithms.md`)
12. 🛠 Scorer composition
13. 🧵 Thread Safety (yes, after construction)
14. 🔧 Configuration with `DefaultScorer()` and custom examples
15. 🎯 Tuning Guidance linking to `docs/tuning.md`
16. 📖 API Reference link to pkg.go.dev
17. 🤖 For AI Assistants — pointer to `llms.txt` / `llms-full.txt`
18. 🤝 Contributing
19. 🔐 Security
20. 📄 Licence

### 16.2 docs/

- `docs/algorithms.md` — per-algorithm detail: formula, complexity, complexity citations, intended use, comparable references, reference vectors
- `docs/scorer.md` — Scorer composition, weight semantics, threshold tuning, parametric algorithms
- `docs/scan.md` — collection-scan layer: Item / Warning / Config, within-vs-cross-group passes, suppression composition, token-bucket optimisation, integration recipes
- `docs/tuning.md` — calibrating thresholds against a domain corpus; choosing the right algorithm subset for the domain
- `docs/extending.md` — composing algorithms manually; building a domain-specific Scorer; phonetic-keys with custom inner metrics
- `docs/performance.md` — benchmark numbers, optimisation notes, profiling tips
- `docs/faq.md` — common questions, "why not Needleman-Wunsch", "why not Soft-TFIDF", "why not ML embeddings", "why was Metaphone 3 excluded", etc.

### 16.3 llms.txt / llms-full.txt

Following the emerging convention. `llms.txt` is a concise reference; `llms-full.txt` includes full API signatures and examples. Verified in sync with code via `ai_friendly_test.go`.

### 16.4 Inline godoc

Every exported symbol has a godoc comment starting with the symbol name. Every algorithm file has the academic source citation at the top of the file. Constants used in algorithms (e.g. `winklerPrefixScale`) are unexported but documented with reference to the originating paper.

### 16.5 example_test.go

Runnable godoc examples (visible on pkg.go.dev). One example per major use case:

- `ExampleScorer` — DefaultScorer with a typical pair
- `ExampleNewScorer` — custom Scorer composition
- `ExampleScorer_ScoreAll` — per-algorithm breakdown
- `ExampleScorer_Match` — threshold matching
- `ExampleLevenshteinScore`, `ExampleJaroWinklerScore`, `ExampleDamerauLevenshteinOSAScore`, `ExampleTokenJaccardScore`, `ExampleDoubleMetaphoneScore`, etc. — one per algorithm
- `ExampleNormalise` — building consistent representations
- `ExampleTokenise` — extracting tokens for token-based downstream processing

---

## 17. CI/CD Requirements

### 17.1 Workflows

`.github/workflows/`:

- `ci.yml` — runs on every PR and push to `main`. Stages:
  1. Lint: `golangci-lint run` with `.golangci.yml` (matching `axonops/mask` configuration)
  2. Vet: `go vet ./...`
  3. Markdown lint: `markdownlint-cli2`
  4. Build: cross-compile for linux/amd64, linux/arm64, darwin/amd64, darwin/arm64, windows/amd64
  5. Test: `go test -race -shuffle=on -count=1 ./...`
  6. Coverage: `go test -cover -coverprofile=coverage.out ./...`. Fail if thresholds missed. Codecov upload.
  7. Fuzz (short): 60s per fuzzer
  8. Determinism: dedicated test runs Scorer.Score 10,000 times asserting byte-identical output across runs
  9. Cross-platform determinism: per-platform job runs the determinism golden-file test
  10. BDD: `cd tests/bdd && go test ./...`
  11. Vulncheck: `govulncheck ./...`
  12. License header check: `scripts/verify-license-headers.sh`
  13. No-runtime-deps check: `scripts/verify-no-runtime-deps.sh`
  14. LLMs sync check: `scripts/verify-llms-sync.sh`

  Target total CI runtime: < 5 minutes for a clean PR.

- `nightly.yml` — daily at 03:00 UTC. Long-form fuzz (5 minutes per fuzzer). Benchmark regression detection vs. last tagged release via `benchstat`. > 10% regression auto-opens a tracking issue.

- `release.yml` — triggered by `vX.Y.Z` tag push. Full CI pipeline. `goreleaser release --clean`. Sigstore keyless signing on `checksums.txt` via `cosign sign-blob --yes --oidc-issuer=https://token.actions.githubusercontent.com`. GitHub Actions OIDC build provenance attestation. CHANGELOG entry extraction → GitHub Release description. `pkg.go.dev` proxy ping for deterministic indexing.

- `security.yml` — weekly + on PR. `gosec ./...`. `govulncheck ./...`. CodeQL semantic analysis. Sigstore signature verification on the most recent release tag.

- `codeql.yml` — GitHub-native CodeQL on every push to `main` and weekly schedule.

### 17.2 Branch protection on main

- Required status checks: lint, vet, markdown-lint, build (all five platforms), test, coverage, fuzz-short, determinism, cross-platform-determinism, BDD, vulncheck, license-check, no-runtime-deps, llms-sync
- Required approving reviews: 1 from CODEOWNER
- No force-pushes
- Linear history required (squash merges only)
- Conversation resolution required before merge
- Releases happen exclusively through CI workflows — never tag locally. The devops agent enforces this.

### 17.3 Issue and PR templates

`.github/ISSUE_TEMPLATE/`:

- `bug_report.md`
- `feature_request.md`
- `algorithm_proposal.md` — for community contributions of new algorithms post-v1: academic source, license check, performance characteristics, expected use case

`.github/pull_request_template.md` — what changes, why, testing, performance impact, compatibility impact, CHANGELOG entry.

### 17.4 Dependabot

- Go modules (PRs grouped: indirect / direct / test-only)
- GitHub Actions versions
- Daily check; auto-merge enabled for patch updates that pass CI

### 17.5 CODEOWNERS

Catch-all maintainer list, with potential per-area refinement (e.g. phonetic algorithms under a domain-expert reviewer once one is identified).

---

## 18. Repository Layout

```text
github.com/axonops/fuzzymatch/
├── .github/
│   ├── workflows/{ci,nightly,release,security,codeql}.yml
│   ├── ISSUE_TEMPLATE/{bug_report,feature_request,algorithm_proposal}.md
│   ├── images/logo-readme.png
│   ├── CODEOWNERS
│   ├── dependabot.yml
│   └── pull_request_template.md
├── .claude/
│   ├── agents/                  # generated in Response C/D
│   └── skills/                  # generated in Response B
├── docs/
│   ├── algorithms.md
│   ├── scorer.md
│   ├── tuning.md
│   ├── extending.md
│   ├── performance.md
│   └── faq.md
├── scripts/
│   ├── verify-license-headers.sh
│   ├── verify-no-runtime-deps.sh
│   ├── verify-llms-sync.sh
│   └── update-bench-txt.sh
├── tests/
│   └── bdd/
│       ├── go.mod                # godog + goleak + testify here
│       ├── features/{algorithms,scorer,normalisation,determinism,scan,suppression}.feature
│       └── steps/steps_test.go
├── testdata/
│   ├── fuzz/                     # fuzz corpora
│   └── golden/                   # cross-version score stability fixtures
├── scan/                         # collection-scan sub-package (Layer 3)
│   ├── doc.go
│   ├── scan.go                   # Item, Config, Warning, Check
│   ├── bucket.go                 # token-bucket optimisation
│   ├── errors.go                 # sentinel errors
│   ├── scan_test.go              # external/black-box unit tests
│   ├── scan_internal_test.go     # bucket-equivalent-to-naive correctness
│   ├── scan_bench_test.go        # benchmarks per §12.6
│   ├── example_test.go           # godoc runnable examples
│   ├── props_test.go             # property tests
│   └── fuzz_test.go              # native fuzz harness
├── .gitignore
├── .golangci.yml
├── .goreleaser.yml
├── .markdownlint-cli2.yaml
├── CHANGELOG.md
├── CLAUDE.md                     # generated in Response B (gitignored)
├── CONTRIBUTING.md
├── LICENSE
├── Makefile
├── NOTICE
├── README.md
├── SECURITY.md
├── bench.txt
├── doc.go                        # package documentation
├── errors.go                     # sentinel errors
├── go.mod                        # zero non-stdlib require lines
├── go.sum
├── llms.txt
├── llms-full.txt
├── algoid.go                     # AlgoID + AlgoIDs() + String()
├── normalise.go                  # Normalise + NormalisationOptions
├── tokenise.go                   # Tokenise
├── scorer.go                     # Scorer + NewScorer + DefaultScorer + options
├── scorer_options.go             # functional options
├── levenshtein.go
├── damerau_osa.go
├── damerau_full.go
├── hamming.go
├── jaro.go
├── jaro_winkler.go
├── strcmp95.go
├── smith_waterman_gotoh.go
├── lcsstr.go
├── q_gram.go                     # Q-Gram Jaccard, Sorensen-Dice, Cosine, Tversky (shared q-gram extraction)
├── monge_elkan.go
├── token_ratio.go                # TokenSort, TokenSet, Partial, TokenJaccard
├── soundex.go
├── double_metaphone.go
├── nysiis.go
├── mra.go
├── ratcliff_obershelp.go
├── *_test.go                     # one external test file per algorithm + scorer + normalise
├── *_internal_test.go            # internal correctness tests where needed
├── *_bench_test.go               # benchmark per algorithm + scorer + normalise
├── example_test.go               # godoc runnable examples
├── fuzz_test.go                  # native Go fuzz harnesses
├── props_test.go                 # property tests via testing/quick
├── documentation_test.go         # meta-test: docs examples compile
├── ai_friendly_test.go           # meta-test: llms.txt sync
├── readme_shop_front_test.go     # meta-test: README example produces documented output
├── makefile_targets_test.go      # meta-test: documented make targets exist
└── internal_coverage_test.go     # meta-test: coverage floor enforcement
```

---

## 19. Release Phasing

Phased delivery. Each phase ends with a tagged release. Phase scope is approximate; the actual scope of each tag is governed by what is implemented and tested.

### Phase 1 — Bootstrap (no release)

Repository scaffolding: `LICENSE`, `NOTICE`, `README.md` skeleton, `.gitignore`, `.golangci.yml`, `.goreleaser.yml`, `.markdownlint-cli2.yaml`, `Makefile`, `CONTRIBUTING.md`, `CODE_OF_CONDUCT.md`, `SECURITY.md`, `CHANGELOG.md`, CI workflows, branch protection, Codecov, pkg.go.dev configured, CODEOWNERS, issue templates, PR template, placeholder logo, `.claude/skills/`, `.claude/agents/`, `CLAUDE.md`.

### Phase 2 — Core algorithms (`v0.1.0` and `v0.2.0`)

`v0.1.0`: Levenshtein, Damerau-Levenshtein OSA, Damerau-Levenshtein Full, Hamming, Jaro, Jaro-Winkler, plus normalisation and tokenisation. Unit tests + property tests + fuzz + benchmarks. Algorithm functions only; no Scorer yet.

`v0.2.0`: Strcmp95, Smith-Waterman-Gotoh, LCSStr, Ratcliff-Obershelp. Same test coverage.

### Phase 3 — Q-gram and token-based (`v0.3.0`)

Q-Gram Jaccard, Sørensen-Dice, Cosine, Tversky, Monge-Elkan, Token Sort Ratio, Token Set Ratio, Partial Ratio, Token Jaccard.

### Phase 4 — Phonetic (`v0.4.0`)

Soundex, Double Metaphone, NYSIIS, MRA.

### Phase 5 — Scorer (`v0.5.0`)

Scorer + NewScorer + DefaultScorer + all option functions + per-algorithm dispatch. Scorer tests + property tests. BDD scenarios for Scorer composition, weight normalisation, threshold matching.

### Phase 6 — Scan sub-package (`v0.6.0`)

`github.com/axonops/fuzzymatch/scan` package: `Item`, `Config`, `Warning`, `Kind` (SPEC OVERRIDE (Phase 9): renamed per §12.1 and 09-CONTEXT.md §1 D-02), `Check`, sentinel errors. Token-bucket optimisation with property test verifying equivalence to naive O(N²). Within-group and cross-group passes. Both suppression mechanisms (`SilenceLint`, `SuppressedPairs`) with composition test. Cross-group identical-name suppression. Deterministic output ordering verified by property test. BDD scenarios for scan and suppression. `scan/example_test.go` runnable godoc examples. Performance budgets per section 12.6 met and committed to `bench.txt`.

### Phase 7 — Integration shakedown via consumer (`v0.6.x` patches)

Re-scope `axonops/audit` issue #853 to consume `github.com/axonops/fuzzymatch` and `github.com/axonops/fuzzymatch/scan`. Surface any API ergonomic issues. `v0.6.x` patch releases as needed.

### Phase 8 — `v1.0.0` stable release

API frozen. Final CHANGELOG. Final benchmark numbers in `bench.txt`. Tag `v1.0.0`. Publish via goreleaser. Sign via cosign. Build provenance attested. Update `axonops/audit` to depend on `fuzzymatch v1.0.0`. Announcement post.

---

## 20. Acceptance Criteria

### Library structure

- [ ] Module path `github.com/axonops/fuzzymatch`
- [ ] Apache-2.0 `LICENSE`, `NOTICE`, `SECURITY.md`, `CODE_OF_CONDUCT.md`, `CONTRIBUTING.md`, `README.md`, `CHANGELOG.md` present
- [ ] Every `.go` file has the standard AxonOps Apache-2.0 header
- [ ] Root `go.mod` has zero non-stdlib `require` lines
- [ ] `tests/bdd/go.mod` is separate so consumers do not transitively depend on godog / goleak / testify
- [ ] No cgo anywhere
- [ ] Repo layout mirrors `axonops/mask` (`.github/`, `docs/`, `scripts/`, `tests/bdd/`, root .go files)
- [ ] `llms.txt` and `llms-full.txt` at repo root
- [ ] `bench.txt` committed and updated per release
- [ ] `.claude/skills/` and `.claude/agents/` populated per the Phase-1 bootstrap

### Algorithms (all 23)

- [ ] Each algorithm implemented from primary source with inline citation
- [ ] Each algorithm's score function is a public function in package fuzzymatch
- [ ] Each algorithm passes literature reference vectors in unit tests
- [ ] Byte and rune variants where meaningful (`*ScoreRunes`)
- [ ] Public `*Distance` function where a distance is meaningful (Levenshtein, Damerau-Levenshtein, Hamming)
- [ ] Phonetic algorithms expose both `*Code` (or `*Keys`) and `*Score` functions
- [ ] Every algorithm has an entry in `AlgoID` and `AlgoIDs()` returns the full ordered list
- [ ] `AlgoID.String()` returns the CamelCase name matching the constant suffix (`"Levenshtein"`, `"JaroWinkler"`, `"DamerauLevenshteinOSA"`, `"NYSIIS"`, etc., per §6 and Phase 8.5 Q6b); mapping stable across patch versions

### Scorer

- [ ] `Scorer` type + `NewScorer(opts...)` + sentinel errors
- [ ] Every `WithXxx` option per section 8.2
- [ ] `DefaultScorer()` returns the documented composition
- [ ] `Score`, `ScoreAll`, `Match`, `Threshold`, `Algorithms` methods
- [ ] Construction validates configuration with appropriate sentinel errors
- [ ] Last-wins behaviour when the same algorithm is added twice
- [ ] Concurrent use safe after construction

### Scan sub-package

- [ ] Package path `github.com/axonops/fuzzymatch/scan` under the same module
- [ ] `Item`, `Config`, `Warning`, `Kind` types per section 12.1 (SPEC OVERRIDE (Phase 9): `Kind` renamed per §12.1 and 09-CONTEXT.md §1 D-02)
- [ ] `Check(items, cfg) ([]Warning, error)` function
- [ ] `ErrNilScorer`, `ErrInvalidItem`, `ErrInvalidConfig` sentinel errors
- [ ] Within-group pass produces expected warnings
- [ ] Cross-group pass disabled by default
- [ ] Cross-group pass enabled produces warnings of correct `Kind`
- [ ] Cross-group identical names suppressed by default
- [ ] Cross-group identical names surfaced when `CompareIdenticalAcrossGroups: true`
- [ ] `Item.SilenceLint` on either side silences pair (one-side suppression)
- [ ] `Config.SuppressedPairs` entries silence pairs in both directions
- [ ] Both suppression mechanisms compose
- [ ] Output sorted deterministically by `(Kind, NameA, NameB, GroupA, GroupB)`
- [ ] Empty input → empty output, no error
- [ ] Nil Scorer → `ErrNilScorer`
- [ ] Empty item Name → `ErrInvalidItem` wrapped with offending index
- [ ] Token-bucket optimisation produces same warnings as naive O(N²) (property test)
- [ ] Performance budgets per section 12.6 met
- [ ] `scan/example_test.go` runnable godoc examples
- [ ] BDD scenarios in `features/scan.feature` and `features/suppression.feature`
- [ ] Root package has no import dependency on the scan sub-package

### Normalisation and tokenisation

- [ ] `Normalise` + `Tokenise` + `NormalisationOptions` + `DefaultNormalisationOptions` per sections 9 and 10
- [ ] Defaults: Lowercase=true, StripSeparators=true, SeparatorChars="_-.:/", SplitCamelCase=true
- [ ] ASCII fast path for short inputs

### Determinism

- [ ] Algorithm scores stable across patch versions (golden-file test)
- [ ] Scorer.Score byte-identical across runs (property test)
- [ ] Cross-platform determinism verified in CI (linux/darwin/windows × amd64/arm64)
- [ ] No map iteration in output paths

### Performance

- [ ] All section-14 budgets met
- [ ] `bench.txt` updated per release
- [ ] `benchstat` regression > 10% fails CI
- [ ] Self-hosted benchmark runner configured (shared with mask/audit)

### Testing

- [ ] Unit coverage ≥ 95% overall, ≥ 90% per file, 100% on public API
- [ ] Property-based tests via `testing/quick`
- [ ] Fuzz tests for every public function
- [ ] Goroutine leak detection via goleak in `TestMain` (in `tests/bdd/`)
- [ ] BDD scenarios for algorithms / scorer / normalisation / determinism
- [ ] Meta-tests: documentation_test.go, ai_friendly_test.go, readme_shop_front_test.go, makefile_targets_test.go, internal_coverage_test.go
- [ ] Race detector clean (`-race`)
- [ ] Root tests use stdlib `testing` only (no testify in root module)

### CI/CD

- [ ] All workflows per section 16.1
- [ ] Branch protection per section 16.2
- [ ] Issue templates + PR template per section 16.3
- [ ] Dependabot configured
- [ ] CODEOWNERS in place
- [ ] CI-only releases enforced (devops agent gate)

### Documentation

- [ ] README mirroring axonops/mask structure with logo, badges, TOC, all sections
- [ ] All `docs/*.md` per section 15.2
- [ ] godoc for every exported symbol with academic citations where relevant
- [ ] CHANGELOG following Keep-a-Changelog
- [ ] `llms.txt` and `llms-full.txt` synchronised with code (verified by `ai_friendly_test.go`)
- [ ] Runnable godoc examples in `example_test.go`

### Stability

- [ ] `v1.0.0` API guarantee documented in README
- [ ] Algorithm output stability documented (scores stable across patch versions)
- [ ] Determinism documented as a guarantee, verified by property test
- [ ] Major version bump policy documented in CONTRIBUTING.md

---

## 21. References

### Primary algorithmic sources

- Levenshtein, V. I. (1965). "Binary codes capable of correcting deletions, insertions, and reversals." *Soviet Physics Doklady*, 10(8):707–710.
- Damerau, F. J. (1964). "A technique for computer detection and correction of spelling errors." *Communications of the ACM*, 7(3):171–176.
- Lowrance, R., Wagner, R. A. (1975). "An extension of the string-to-string correction problem." *Journal of the ACM*, 22(2):177–183.
- Boytsov, L. (2011). "Indexing methods for approximate dictionary searching: comparative analysis." *ACM Journal of Experimental Algorithmics*, 16, Article 1.
- Hamming, R. W. (1950). "Error detecting and error correcting codes." *Bell System Technical Journal*, 29(2):147–160.
- Jaro, M. A. (1989). "Advances in record-linkage methodology as applied to matching the 1985 census of Tampa, Florida." *Journal of the American Statistical Association*, 84(406):414–420.
- Winkler, W. E. (1990). "String comparator metrics and enhanced decision rules in the Fellegi-Sunter model of record linkage." *Proceedings of the Section on Survey Research Methods*, American Statistical Association: 354–359.
- Winkler, W. E. (1994). "Advanced methods for record linkage." *Proceedings of the Section on Survey Research Methods*: 467–472.
- Smith, T. F., Waterman, M. S. (1981). "Identification of common molecular subsequences." *Journal of Molecular Biology*, 147(1):195–197.
- Gotoh, O. (1982). "An improved algorithm for matching biological sequences." *Journal of Molecular Biology*, 162(3):705–708.
- Wagner, R. A., Fischer, M. J. (1974). "The string-to-string correction problem." *Journal of the ACM*, 21(1):168–173.
- Ukkonen, E. (1992). "Approximate string-matching with q-grams and maximal matches." *Theoretical Computer Science*, 92(1):191–211.
- Jaccard, P. (1912). "The distribution of the flora in the alpine zone." *New Phytologist*, 11(2):37–50.
- Dice, L. R. (1945). "Measures of the amount of ecologic association between species." *Ecology*, 26(3):297–302.
- Sørensen, T. (1948). "A method of establishing groups of equal amplitude in plant sociology." *Kongelige Danske Videnskabernes Selskab*, 5(4):1–34.
- Salton, G., McGill, M. J. (1983). *Introduction to Modern Information Retrieval*. McGraw-Hill.
- Tversky, A. (1977). "Features of similarity." *Psychological Review*, 84(4):327–352.
- Monge, A. E., Elkan, C. P. (1996). "The field matching problem: algorithms and applications." *Proceedings of the Second International Conference on Knowledge Discovery and Data Mining*: 267–270.
- Russell, R. C. (1918). U.S. Patent 1261167 (Soundex).
- Knuth, D. E. (1973). *The Art of Computer Programming, Volume 3*. Section 6.4. (Canonical Soundex description.)
- Philips, L. (2000). "The double-metaphone search algorithm." *C/C++ Users Journal*, 18(6):38–43.
- Taft, R. L. (1970). *Name search techniques*. New York State Identification and Intelligence System, Special Report No. 1.
- Moore, G. B., Kuhns, J. L., Trefftzs, J. L., Montgomery, C. A. (1977). *Accessing individual records from personal data files using non-unique identifiers*. National Bureau of Standards Technical Note 943.
- Ratcliff, J. W., Metzener, D. E. (1988). "Pattern matching: the gestalt approach." *Dr. Dobb's Journal*, 13(7):46–51.
- Bachmann, M. (2020–present). RapidFuzz documentation. <https://rapidfuzz.github.io/RapidFuzz/>. (Canonical modern reference for Token Sort/Set/Partial Ratio shapes.)

### Empirical study referenced for algorithm coverage decisions

- Fränti, P., Mariescu-Istodor, R., Sengupta, L. (2016). "Similarity measures for title matching." University of Eastern Finland. (The "21 measures across 4,968 title strings" study cited for algorithm-set rationalisation.)

### Reference Go implementations (studied for reference vectors only — code reimplemented from primary sources)

- `github.com/agnivade/levenshtein` (MIT) — Levenshtein
- `github.com/xrash/smetrics` (MIT) — Jaro, Jaro-Winkler, Strcmp95, Soundex, Hamming
- `github.com/adrg/strutil` (MIT) — Smith-Waterman-Gotoh, Jaccard, Sørensen-Dice, architecture pattern reference
- `github.com/hbollon/go-edlib` (MIT) — Damerau-Levenshtein OSA + Full, Q-gram, Cosine
- `github.com/tilotech/go-phonetics` (MIT) — Metaphone reference
- `github.com/CalypsoSys/godoublemetaphone` (MIT) — Double Metaphone reference (not copied)
- `github.com/UjjwalAyyangar/go-jellyfish` (MIT) — NYSIIS reference

### Project reference

- `github.com/axonops/mask` — structural and process template (repo layout, BDD pattern, meta-tests, CI workflows, release engineering, README structure, `.claude/` skills and agents, llms.txt). None of mask's API surface or domain semantics applies to fuzzymatch.

---

*End of requirements document. This document is the authoritative specification for `github.com/axonops/fuzzymatch` v1.0.0.*
