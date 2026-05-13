# Feature Research

**Domain:** Go string similarity / fuzzy-matching library
**Researched:** 2026-05-13
**Confidence:** HIGH (spec is comprehensive; ecosystem well-surveyed in `docs/prior-art-research.md`; cross-checked against RapidFuzz, FuzzyWuzzy, go-edlib, adrg/strutil, smetrics)

> **Reading note.** This document catalogues *features* — capabilities consumers of a string-similarity library expect to find or value as a differentiator. It is **not** an API design document; the `api-ergonomics-reviewer` has final authority on the surface shape. Where a feature is described with a signature, that signature is illustrative.

---

## Executive Summary

The Go string-similarity ecosystem is fragmented across single-purpose libraries (`agnivade/levenshtein`, `xrash/smetrics`), small curated collections (`adrg/strutil` — 7 algorithms, inactive in 2024+), and comprehensive but inconsistent libraries (`hbollon/go-edlib` — 10+ algorithms, mixed maintenance). No Go library in 2026 offers (a) the full 23-algorithm catalogue fuzzymatch targets, (b) a typed-algorithm composable Scorer, (c) explicit cross-platform determinism guarantees, (d) production-grade DX (godoc examples + llms.txt + BDD), and (e) zero runtime dependencies + Apache-2.0 + no cgo, simultaneously. **The combination is the differentiator**; any single dimension already exists elsewhere.

The Python ecosystem (RapidFuzz, FuzzyWuzzy) is the cross-language reference for ergonomic shape. RapidFuzz dominates because it ships Levenshtein, Damerau-Levenshtein, Jaro/Jaro-Winkler, Hamming, OSA, Indel, Postfix-LCS, partial/token/sort/set ratios, plus a `process.extract` / `cdist` collection-scan API — all behind a coherent `fuzz`/`distance`/`process` namespacing with `score_cutoff` and `processor` parameters. Spec aligns well: §7 covers the algorithms; §8 covers the Scorer; §12 covers a collection scan. **One ergonomic gap** in the spec versus RapidFuzz: there is no equivalent of `process.extract(query, choices)` / `process.cdist(queries, choices)` for one-to-many or many-to-many search — the spec scopes that explicitly out and defers to a future `fuzzymatch/search` sub-package. This is the right call for v1, but it's the single most-requested feature in comparable libraries and should be tracked.

**Spec gaps worth flagging** (most actionable first): (1) no Unicode normalisation (NFC/NFD/NFKC/NFKD) primitives — case folding only; diacritic stripping is not in the pipeline; (2) no `process.extract`-style one-to-many search API in v1 (explicit, defensible, but flag for v1.x demand); (3) no streaming / iterator API for very large collections (intentional, but worth a one-line FAQ entry); (4) `DefaultScorer()` composition is opinionated and may surprise consumers — needs a tuning guide pointer in godoc; (5) phonetic-as-binary integration into the weighted Scorer is unusual — most libraries treat phonetics as a pre-filter, not a 1.0/0.0 contributor — needs explicit documentation.

---

## Feature Landscape

### Table Stakes (Users Expect These)

Features users assume a comprehensive Go fuzzy-matching library will have. Missing any of these would feel incomplete and would invite "why not just use go-edlib?" pushback.

#### Algorithm coverage — table stakes

| Feature | Why Expected | Complexity | Notes / Spec Status |
|---------|--------------|------------|---------------------|
| **Levenshtein distance + normalised score** | Universally cited as the baseline edit-distance metric. Every Go fuzzy library has it. | LOW | **spec-confirmed** §7.1.1. Two-row DP optimisation expected. |
| **Damerau-Levenshtein (OSA variant)** | Industry baseline for "handles transpositions like keyboard typos". 21-algorithm Joensuu study ranks it joint-best for short identifiers. | LOW-MEDIUM | **spec-confirmed** §7.1.2. |
| **Jaro & Jaro-Winkler** | Canonical for short-name and identifier matching, especially prefix-aligned families. `xrash/smetrics` is the de-facto reference. | LOW | **spec-confirmed** §7.1.5, §7.1.6. |
| **Hamming distance** | Trivial to implement, frequently requested for fixed-width codes. Missing it would look incomplete. | LOW | **spec-confirmed** §7.1.4. |
| **Q-gram / Trigram Jaccard** | PostgreSQL `pg_trgm` and most fuzzy-search systems use trigrams. Consumers reach for it instinctively for partial matches. | LOW | **spec-confirmed** §7.2.1. |
| **Sørensen-Dice (bigram)** | Widely paired with Jaccard; standard for fuzzy-search engines. | LOW | **spec-confirmed** §7.2.2. |
| **Cosine similarity (n-gram)** | Standard IR metric; consumers expect it alongside Jaccard. | LOW | **spec-confirmed** §7.2.3. |
| **Token-based ratios (Sort / Set / Partial)** | RapidFuzz / FuzzyWuzzy lineage. The single most-cited reason to pick a fuzzy lib in Python-influenced codebases. | MEDIUM | **spec-confirmed** §7.3.2–§7.3.4. |
| **Soundex phonetic encoding** | The "starter phonetic algorithm" every library has. Even one-line implementations exist in stdlibs of other languages. | LOW | **spec-confirmed** §7.4.1. |
| **Composable scoring API** | `adrg/strutil` set the bar with the `StringMetric` interface; any library without composability looks dated. | MEDIUM | **spec-confirmed** §8 (Scorer + functional options). Spec uses an enum (`AlgoID`) rather than an interface; ergonomically equivalent for the user, more compile-time-safe internally. |
| **Score in `[0.0, 1.0]`** | Distances are not comparable across algorithms. Every modern library normalises. | LOW | **spec-confirmed** §6, every algorithm in §7. |

#### Normalisation — table stakes

| Feature | Why Expected | Complexity | Notes / Spec Status |
|---------|--------------|------------|---------------------|
| **Case folding (lowercase)** | Every comparable library does it. Hard-coded into FuzzyWuzzy's default processor. | LOW | **spec-confirmed** §9. `Lowercase: true` default. |
| **Separator stripping (`_`, `-`, `.`, `:`, `/`)** | Critical for identifier matching (`user_id` vs `user-id`). Audit-event taxonomy use case demands it. | LOW | **spec-confirmed** §9. `StripSeparators: true`, `SeparatorChars: "_-.:/"`. |
| **CamelCase / snake_case / PascalCase tokenisation** | Identifier-style names are the dominant use case for fuzzy matching of code/schema vocabulary. | MEDIUM | **spec-confirmed** §10. Splits on uppercase boundaries, acronym handling (`XMLParser` → `[xml, parser]`). |
| **Whitespace handling** | Tokenisation must collapse multiple consecutive separators into one boundary. | LOW | **spec-confirmed** §10, step 4 "Filter empty tokens". |
| **Idempotency** | `Normalise(Normalise(x)) == Normalise(x)` is a basic correctness invariant consumers assume. | LOW | **spec-confirmed** §13.6 property test. |

#### API ergonomics — table stakes

| Feature | Why Expected | Complexity | Notes / Spec Status |
|---------|--------------|------------|---------------------|
| **Single-algorithm functions** | `LevenshteinScore(a, b)` — consumers wanting one algorithm should not need to construct a Scorer. `agnivade/levenshtein` set this bar; every Go library follows. | LOW | **spec-confirmed** §6, §7 (every algorithm has a `*Score` public function). |
| **Pure functions** | No globals, no I/O, no `init()` side effects. Idiomatic Go. | LOW | **spec-confirmed** §5 (Design Principle 3, 12). |
| **Concurrent-safe Scorer** | Server contexts call from many goroutines. Immutable-after-construction is the obvious shape. | LOW | **spec-confirmed** §5, §8. |
| **Functional options for Scorer** | Rob Pike pattern; the idiomatic Go answer to "configurable constructor with N optional params". | LOW | **spec-confirmed** §8.2. |
| **Sentinel errors with `errors.Is` discrimination** | Standard since Go 1.13. Wrapping via `%w` is required. | LOW | **spec-confirmed** §6 (sentinel errors), §8.1 (construction validation). |
| **Threshold + Match helper** | "Score > X → match" is so common consumers expect a `Match` shortcut rather than rolling their own. | LOW | **spec-confirmed** §8.3 (`Match`). |
| **Per-algorithm breakdown for debugging** | "Why did this pair score 0.7?" is the single most-asked tuning question. `ScoreAll` answers it. | LOW | **spec-confirmed** §8.3 (`ScoreAll`). |
| **Identity, symmetry, range invariants** | Mathematical correctness is the floor. Consumers will write property tests if the library doesn't claim them. | LOW | **spec-confirmed** §13.6 property tests. |

#### Edge-case handling — table stakes

| Feature | Why Expected | Complexity | Notes / Spec Status |
|---------|--------------|------------|---------------------|
| **Empty input handling** | Documented score for `("", "")`, `("", x)`, `(x, "")`. Every algorithm differs. | LOW | **spec-confirmed** §7 (every algorithm specifies edge cases). |
| **Identical-input handling** | `Score(x, x) == 1.0` for non-empty `x`. Invariant. | LOW | **spec-confirmed** §13.6 `PropAlgorithmScore_Identity`. |
| **Unicode / non-ASCII inputs** | `adrg/strutil` and `jcoruiz/strsim` ship rune-aware code. Byte-only implementations are seen as dated. | MEDIUM | **spec-confirmed** §5 (Design Principle 6), every algorithm exposes `*ScoreRunes` variant. |
| **No panic on malformed UTF-8** | Hardening floor. RapidFuzz / go-edlib both gracefully handle it. | LOW | **spec-confirmed** §13.6 `PropAllPublic_NeverPanic`. |
| **Length-mismatch handling for Hamming** | Either return error or document a degraded score. Both choices are seen in the wild. | LOW | **spec-confirmed** §7.1.4 (both variants — `*Distance` errors, `*Score` returns 0.0). |

#### Testing & quality — table stakes

| Feature | Why Expected | Complexity | Notes / Spec Status |
|---------|--------------|------------|---------------------|
| **Literature reference vectors in tests** | "kitten/sitting → 3" is the universal Levenshtein smoke test. Every library has it. | LOW | **spec-confirmed** §15.1, with reference vectors in §7 per algorithm. |
| **Race-clean (`go test -race`)** | Every modern Go library demonstrates race-safety. | LOW | **spec-confirmed** §15.10. |
| **godoc with runnable examples** | pkg.go.dev quality bar. | LOW | **spec-confirmed** §16.5. |
| **README quick-start that copies/pastes-and-runs** | First touch for evaluating a library. | LOW | **spec-confirmed** §16.1. |
| **Apache-2.0 or MIT licence** | Both are accepted; GPL/LGPL would be a non-starter for many enterprise consumers. | LOW | **spec-confirmed** §1, §5 (Design Principle 9). |

---

### Differentiators (Competitive Advantage)

Features that no single competitor offers, or features that set fuzzymatch apart from the pack.

#### Algorithm catalogue — differentiators

| Feature | Value Proposition | Complexity | Notes / Spec Status |
|---------|-------------------|------------|---------------------|
| **23-algorithm catalogue under one Apache-2.0 module** | No Go library ships this many algorithms today. `go-edlib` has ~10, `adrg/strutil` has 7, `jcoruiz/strsim` claims 15+ but is newer. Comprehensive coverage is the headline differentiator. | HIGH (cumulative) | **spec-differentiator** §1. |
| **Damerau-Levenshtein FULL (Lowrance-Wagner)** | OSA is widespread; full DL is rare in Go. Slight constant-factor cost; mathematically correct in all transposition cases. | MEDIUM | **spec-differentiator** §7.1.3. |
| **Strcmp95 (Winkler 1994 refinement)** | Refined Jaro-Winkler with similar-character table and long-string bonus. Present in `xrash/smetrics` but seldom advertised. | MEDIUM | **spec-differentiator** §7.1.7. |
| **Smith-Waterman-Gotoh with configurable affine gap penalty** | Most Go libraries either don't ship it or ship a hard-coded-penalty variant. Configurable `SWGParams` is rare. | MEDIUM | **spec-differentiator** §7.1.8. |
| **Tversky (asymmetric)** | Rarely exposed in Go libraries. Reduces to Jaccard / Dice as special cases; valuable when "this is the prototype" semantics matter. | LOW | **spec-differentiator** §7.2.4. |
| **Monge-Elkan with pluggable inner metric** | Hybrid algorithms with arbitrary inner metric (Jaro-Winkler default, Q-Gram-Jaccard, even phonetic) — none of the Go libraries currently expose this. | MEDIUM | **spec-differentiator** §7.3.1. |
| **Token Set Ratio + Token Sort Ratio + Partial Ratio (RapidFuzz parity)** | Python-derived consumers expect these by name. `go-fuzzywuzzy` is the only Go lib with them, and it's ageing. | MEDIUM | **spec-differentiator** §7.3.2–§7.3.4. Modern Indel-ratio formulation per RapidFuzz, not classic Levenshtein-ratio. |
| **Four phonetic algorithms (Soundex, Double Metaphone, NYSIIS, MRA)** | Soundex is everywhere; Double Metaphone is occasional; NYSIIS and MRA together are vanishingly rare in Go. | MEDIUM | **spec-differentiator** §7.4. |
| **Ratcliff-Obershelp (Gestalt) matching** | Python's `difflib.SequenceMatcher` is famous; the Go ecosystem mostly lacks it. | MEDIUM | **spec-differentiator** §7.5.1. |
| **LCSStr (Longest Common Substring) as a named algorithm** | Useful for "names share a long middle" use case (`my_request_id` vs `your_request_handle`). Not commonly named in Go libs even when implementable. | LOW | **spec-differentiator** §7.1.9. |

#### Composability — differentiators

| Feature | Value Proposition | Complexity | Notes / Spec Status |
|---------|-------------------|------------|---------------------|
| **Weighted composite Scorer with normalisation toggle** | go-edlib has `StringsSimilarity(a, b, algoEnum)` but no first-class composition; adrg/strutil has the interface but no weighted-composite primitive. Spec's `Scorer` is unique. | MEDIUM | **spec-differentiator** §8. |
| **Auto-normalised weights (sum-to-1 invariant)** | Consumers commonly mis-weight (e.g. 0.3 + 0.3 + 0.3 = 0.9). Auto-normalisation hides this footgun by default. | LOW | **spec-differentiator** §8.4. |
| **Last-wins behaviour for duplicate algorithm options** | Idiomatic Go (functional options frequently have this semantic); explicitly documented avoids surprise. | LOW | **spec-confirmed** §8.2. Slight differentiator: making it a documented promise. |
| **`DefaultScorer()` calibrated for identifier matching** | New users with no calibration time get a sensible starting point. None of the Go libraries ship an opinionated default. | LOW | **spec-differentiator** §8.5. **FLAG:** the default composition is opinionated; needs a docs/tuning.md pointer in the godoc to avoid surprise. |
| **Parametric algorithm options (`WithQGramJaccardAlgorithm(weight, n int)`)** | Rather than burying `n=3` as a constant, expose it. Tuning q-gram size matters more than most consumers initially realise. | LOW | **spec-differentiator** §8.2. |

#### Determinism — differentiators

| Feature | Value Proposition | Complexity | Notes / Spec Status |
|---------|-------------------|------------|---------------------|
| **Documented cross-platform byte-identical output** | No Go fuzzy-match library makes this an explicit guarantee. Verified by CI matrix (linux/darwin/windows × amd64/arm64) + golden-file test. | MEDIUM | **spec-differentiator** §13.3. Audit-event consumer requires it for reproducible warnings. |
| **Algorithm output stability across patch versions** | `Score(a, b)` in v1.0 must equal `Score(a, b)` in v1.5 to the last bit. Pin via golden file. | LOW (implementation; HIGH for discipline) | **spec-differentiator** §5 (Design Principle 8), §13.1. |
| **No map iteration on output paths** | Subtle determinism trap (`ScoreAll`'s map is documented as non-deterministic order; everything internal is sorted). | LOW | **spec-differentiator** §13.4. Documented and property-tested. |
| **NaN/Inf/-0.0 handled explicitly** | Float-edge correctness is rare in fuzzy-match libraries. | LOW | **spec-differentiator** implied §13.3. |

#### Collection scan (`scan` sub-package) — differentiators

| Feature | Value Proposition | Complexity | Notes / Spec Status |
|---------|-------------------|------------|---------------------|
| **Within-group + cross-group passes with separate thresholds** | Real-world taxonomy: items in the same category are noisy; items across categories are noisier still. Two-pass distinction is novel. | MEDIUM | **spec-differentiator** §12.2. |
| **Per-item `SilenceLint` suppression + global `SuppressedPairs` list, composing** | Two complementary mechanisms with documented composition rules. Most fuzzy-match libraries don't ship suppression at all. | MEDIUM | **spec-differentiator** §12.3. |
| **Cross-group identical-name suppression default** | Real schemas legitimately reuse `user_id` across event groups; surfacing every such pair drowns real signal. | LOW | **spec-differentiator** §12.3. |
| **Token-bucket optimisation with property-test-verified naive equivalence** | O(N²) at 10k items is 100M comparisons; the bucket index reduces to O(N + candidate set). Verified equivalent to naive by property test. | HIGH | **spec-differentiator** §12.5. |
| **Deterministic warning ordering by `(Kind, NameA, NameB, GroupA, GroupB)`** | Stable golden output, reproducible audit reports. | LOW | **spec-differentiator** §12.4. |
| **Opaque `Tag any` field for consumer context** | Consumer attaches source-line numbers, schema paths, etc. without API friction. | LOW | **spec-differentiator** §12.1. |

#### Developer experience — differentiators

| Feature | Value Proposition | Complexity | Notes / Spec Status |
|---------|-------------------|------------|---------------------|
| **Zero runtime dependencies, structurally verified in CI** | `script/verify-no-runtime-deps.sh` makes the zero-dep claim auditable. | LOW | **spec-differentiator** §17.1, §5 (Design Principle 1). |
| **No cgo, anywhere** | Cross-compilation, supply-chain hygiene, embedded-system viability. | LOW | **spec-differentiator** §1, §5 (Design Principle 2). |
| **`llms.txt` + `llms-full.txt` for AI assistant consumption** | Emerging convention; AI-friendly docs are increasingly expected by 2026 consumers. Verified in sync via `ai_friendly_test.go`. | LOW | **spec-differentiator** §16.3. |
| **BDD scenarios (godog) per algorithm and per Scorer/scan pattern** | Specification-as-tests; rare in Go ecosystem libraries. | MEDIUM | **spec-differentiator** §15.6. |
| **Meta-tests: README example compiles, docs examples compile, Makefile-targets-match-docs** | Documentation rot is the biggest hidden cost. Meta-tests prevent it. Mirrors `axonops/mask`. | MEDIUM | **spec-differentiator** §15.7. |
| **Sigstore keyless signing + OIDC build provenance attestation on every release** | Supply-chain best practice; not common in mid-size Go libraries. | LOW (CI work) | **spec-differentiator** §17.1 `release.yml`. |
| **CI-only releases (no local tagging permitted)** | Reproducibility. Devops agent enforces it as a gating rule. | LOW (governance) | **spec-differentiator** §17.2. |
| **Coverage targets enforced in CI (95% overall, 90% per file, 100% public API)** | Hard floor with build failure rather than aspirational target. | LOW | **spec-differentiator** §15.8. |
| **Per-algorithm allocation budgets in benchmark CI** | `benchstat` regression > 10% fails the build. Most libraries don't enforce this. | MEDIUM | **spec-differentiator** §14.4. |
| **ASCII fast paths with stack-allocated buffers** | `[64]byte` stack buffer for ≤ 64-byte inputs → zero heap allocation for the common case. | MEDIUM | **spec-differentiator** §5 (Design Principle 10), §7.1.1 implementation notes. |

---

### Anti-Features (Commonly Requested, Often Problematic)

Features that surface in feature-request issues for comparable libraries but that fuzzymatch should deliberately *not* ship. Each has a documented "why not" and an alternative.

| Anti-Feature | Why Requested | Why Problematic | Alternative |
|--------------|---------------|-----------------|-------------|
| **`process.extract(query, choices)` / `process.cdist` one-to-many search API** | RapidFuzz's most-used function. Consumers ask for it instinctively. | Different problem class from collection-pairwise scanning. Optimised one-to-many search wants BK-trees, prefix indexes, or similar — a v1 implementation would either be naive (slow) or scope-creep. | **spec-out-of-scope** §4. Defer to a future `fuzzymatch/search` sub-package gated on demand. Document the workaround (loop + algorithm function) in FAQ. **FLAG:** this is the single most-requested feature in comparable libraries — track demand from v0.6.0 onward. |
| **Embedding-based / semantic similarity** | "Why not use embeddings?" appears in every fuzzy-matching library's issues. | Requires an external ML model. Contradicts zero-runtime-dep, no-I/O, no-cgo, pure-function design. ML and fuzzy-matching are different problem classes with different latency profiles. | **spec-out-of-scope** §4. Document FAQ: "If you have an embedding model, compose it externally — fuzzymatch is the deterministic fallback when you don't." |
| **Persistent state / caching across calls** | "Why does the Scorer not cache?" Performance optimisation request. | Adds eviction policy, thread-safety, observable state. Pure-function design is the contract. Consumers can wrap. | **spec-out-of-scope** §4. Consumer-owned LRU on top of Scorer is the answer; document as a recipe in docs/extending.md. |
| **Goroutines / parallel scanning in `scan.Check`** | "Why not use all my cores?" | Goroutine scheduling is non-deterministic. Determinism guarantee (§13) would break. Reproducibility wins over wall-clock for the audit-event use case. | **spec-out-of-scope** by Design Principle 4 (deterministic output). Document: callers can shard inputs themselves if they want parallelism. **Note:** if benchmarks show this matters, consider a `ScanParallel` opt-in variant in v1.x that explicitly trades determinism. |
| **I/O on the hot path** | "Why can't I pass a file?" | Library design boundary; I/O is the consumer's problem. | **spec-out-of-scope** §4. |
| **Configuration file parsing** | "Why no YAML config?" | The consumer parses their own config and translates to `Scorer` options. Keeps the library stateless. | **spec-out-of-scope** §4. |
| **CLI tool / binary** | Standalone usability. | Library, not a tool. Wrap externally. | **spec-out-of-scope** §4. |
| **Web UI / API server** | "Demo UI?" | Out of scope. Consumer responsibility. | **spec-out-of-scope** §4. |
| **Needleman-Wunsch (global alignment)** | "It's a canonical algorithm." | Functionally redundant given Smith-Waterman-Gotoh handles the same use cases better for identifier-style strings. Adds maintenance without value. | **spec-out-of-scope** §4. Document in FAQ: "Use Smith-Waterman-Gotoh with appropriate gap penalties." |
| **Soft-TFIDF** | Empirical studies rank it competitive. | Requires consumer-supplied corpus frequency table; conflicts with pure-function stateless design. | **spec-out-of-scope** §4. Document in FAQ. |
| **Metaphone 3** | "It's the most accurate phonetic algorithm." | U.S. Patent 7440941. AxonOps declines to ship patent-encumbered algorithms even when not actively enforced. | **spec-out-of-scope** §4. Document in FAQ; point consumers at Double Metaphone (covered) and external libraries if they specifically need Metaphone 3. |
| **`testify` in root module tests** | "Tests would be more readable." | Adds a test-only dependency to the root module. Stricter than `mask`. Stdlib `testing` is verbose but clean. | **spec-out-of-scope** §15.11. `testify` permitted in `tests/bdd/` only. |
| **Windows-specific tooling investment** | "Make windows a first-class platform." | Windows must pass the determinism gate (and does, per §13.3). Beyond that, no Windows-specific tooling. | **spec-out-of-scope** §4 of PROJECT.md. Determinism gate is sufficient. |
| **Streaming / iterator API for scan** | "Can I scan as a stream?" | Adds API surface and stateful semantics. Consumers can shard their input slices to bound memory if needed. | **spec-out-of-scope** by §12.1 (`Check(items []Item, cfg Config) ([]Warning, error)`). Document recipe in `docs/scan.md` for chunked input. **FLAG:** worth a one-line FAQ entry. |
| **Iterator API via Go 1.23+ `iter.Seq`** | New language feature; "modern" Go. | The scan API returns `[]Warning`; the iter-of-warning shape is plausible but adds duplicated surface. Defer until a use case demands it. | **spec-out-of-scope** by §6 / §12.1. Track Go community adoption of `iter.Seq` for library APIs through v1.x. |
| **Generics for algorithm dispatch (`Score[T constraints.Ordered]`)** | Generics are idiomatic in Go 1.18+. | String similarity operates on `string` (sometimes `[]byte` / `[]rune`). Generic dispatch would obscure the byte-vs-rune distinction the spec explicitly surfaces. Concrete types win for clarity. | **Out of scope** by inference from spec §6. **FLAG:** worth a one-line rationale in `docs/faq.md` ("Why aren't algorithm functions generic?"). |

---

### Spec Gaps — Worth Surfacing

Features that consumers may expect (or that comparable libraries provide) where the spec is currently silent or weak. Flagged for user review and roadmap discussion.

| Gap | Consumer Expectation | Severity | Recommendation |
|-----|----------------------|----------|----------------|
| **No Unicode normalisation (NFC/NFD/NFKC/NFKD) in the `Normalise` pipeline** | "`café` and `cafe` should match" is a frequent fuzzy-matching expectation. Spec covers lowercase + separator strip + camelCase split — none of those handle precomposed-vs-decomposed Unicode or diacritic stripping. | MEDIUM | **Recommend:** add `NormalisationOptions.UnicodeForm` (one of `None`, `NFC`, `NFD`) and `NormalisationOptions.StripDiacritics` (bool, applies after NFD decomposition). Both default to off in v1.0 to preserve byte-identical output for ASCII inputs, but expose them. Use `golang.org/x/text/unicode/norm` — wait, that's a non-stdlib dep. **Alternative:** implement a minimal NFC/NFD normaliser inline, or document the gap explicitly and recommend pre-normalisation by the consumer using `x/text/unicode/norm` before passing to fuzzymatch. **Decision needed.** |
| **No diacritic / accent stripping** | `naïve` vs `naive`, `Müller` vs `Mueller`, `café` vs `cafe` — all common audit-field-name edge cases. | MEDIUM | **Recommend:** see above. If `x/text/unicode/norm` is disallowed, document this as an explicit consumer responsibility with a recipe. |
| **No `process.extract(query, choices, limit=N)` one-to-many search** | Most-requested feature in RapidFuzz / FuzzyWuzzy / go-fuzzywuzzy. | LOW | **spec-out-of-scope** §4 by explicit design. Worth tracking demand from v0.6.0 onward; sub-package `fuzzymatch/search` in v1.x or v2 if demand surfaces. |
| **`DefaultScorer()` composition is opinionated; no mechanism for "default minus algorithm X"** | Consumers wanting a tweak of the default must rebuild from scratch. | LOW | **Recommend:** `DefaultScorerOptions()` returning the `[]ScorerOption` slice, which consumers can append to / filter. Minor ergonomic win. **Decision needed.** |
| **Phonetic algorithms as binary contributors (1.0/0.0) into a weighted composite is unusual** | Most libraries treat phonetics as pre-filter, not weighted contributor. Mid-weight 0.05 contribution from Double Metaphone (per `DefaultScorer`) may not align with consumer mental models. | LOW | **Recommend:** explicit godoc + docs/tuning.md section explaining the phonetic-as-binary integration. **Documentation gap, not API gap.** |
| **No "explain why this score" beyond `ScoreAll`** | "Score is 0.71; which algorithm pulled it down?" `ScoreAll` answers it, but a `Scorer.Explain(a, b) ScorerExplanation` struct with per-algorithm score + weight + contribution + threshold-delta would be richer. | LOW | **Recommend:** v1.x feature. `ScoreAll` is sufficient for v1.0; richer explanation can come later. |
| **No `Score` variant that takes pre-normalised inputs** | Hot-path callers want to normalise once and call `Score` many times. The Scorer currently re-normalises every call. | MEDIUM (performance) | **Recommend:** `Scorer.ScoreNormalised(a, b)` accepts already-normalised inputs and skips normalisation. Or document that consumers can build Scorer with `WithoutNormalisation()` and pre-normalise themselves. **Decision needed for v1.x.** |
| **No `Scorer.Equal(other) bool` or `Scorer.Fingerprint() string`** | Consumers caching by Scorer config want to compare or fingerprint configs. | LOW | **Recommend:** defer to v1.x. Workaround is to track the config externally. |
| **Bench targets specify ASCII inputs ≤ 50 chars; no targets for Unicode-heavy or long-string inputs** | Real-world strings can be longer (paths, full schema names). | LOW | **Already implicitly addressed** by "50/200/500 chars" benchmarks in §14.1. Documentation could foreground this more. |
| **No deprecation policy for algorithms** | If a primary source is retracted or an algorithm is later found incorrect, what's the deprecation path? | LOW | **Recommend:** document in CONTRIBUTING.md — algorithms can be added but not removed within a major version; scoring changes require minor bump + CHANGELOG entry per §5 (Design Principle 8). |
| **No mention of stability for `Tokenise` output across versions** | If `Tokenise("XMLParser")` changes from `["xml", "parser"]` to `["xmlparser"]` between versions, every downstream token-based score changes. | MEDIUM | **Recommend:** explicitly include `Tokenise` and `Normalise` outputs in the determinism golden-file (§13.1). Spec implies this; making it explicit closes the gap. |
| **No `examples/` directory at repo root** | Consumers like having runnable example programs (not just godoc examples). `axonops/mask` has `examples/`. | LOW | **Recommend:** `examples/` directory with `examples/audit-fields/`, `examples/identifier-similarity/`, `examples/schema-deduplication/` runnable mini-programs. Mirrors mask. |

---

## Feature Dependencies

```
Algorithm functions (Layer 1)
    │
    ├── Normalise / Tokenise (foundational primitives)
    │       │
    │       └── Used by token-based algorithms (7.3.1–7.3.5)
    │       └── Used by Scorer normalisation pipeline (§8)
    │       └── Used by scan token-bucket optimisation (§12.5)
    │
    ├── AlgoID enum + AlgoIDs() + AlgoID.String()
    │       │
    │       └── Required by Scorer for algorithm dispatch (§8)
    │       └── Required by ScoreAll map keys (§8.3)
    │
    └── Per-algorithm score functions
            │
            └── Each requires:
                  - Primary source citation
                  - Reference vector unit tests
                  - Property tests (identity, symmetry, range)
                  - Fuzz test
                  - Benchmark with allocation budget

Scorer (Layer 2)
    │
    ├── Requires: all 23 algorithm functions (Layer 1)
    ├── Requires: Normalise / Tokenise / NormalisationOptions
    ├── Requires: AlgoID enum
    ├── Requires: sentinel errors
    │
    └── Provides:
          - Composable weighted scoring
          - Match (threshold helper)
          - ScoreAll (per-algorithm breakdown)
          - DefaultScorer (opinionated default)

Scan sub-package (Layer 3)
    │
    ├── Requires: *fuzzymatch.Scorer (Layer 2)
    ├── Requires: Tokenise (for token-bucket optimisation)
    │
    └── Provides:
          - Item / Warning / WarningKind / Config / Check
          - Within-group + cross-group passes
          - Suppression composition
          - Token-bucket optimisation
          - Deterministic output ordering
```

### Dependency Notes

- **Token-based algorithms (Monge-Elkan, TokenSort, TokenSet, Partial, TokenJaccard) require `Tokenise`.** Tokenise must land before or with token-based algorithms (Phase 3 in spec §19 release phasing).
- **Scorer requires all 23 algorithms.** Spec defers Scorer to Phase 5 (`v0.5.0`) — sensible, since the Scorer is only useful once the full catalogue exists. **Alternative:** ship a partial Scorer alongside algorithm phases, but the partial-then-final shape risks API churn. Spec choice is correct.
- **Scan requires Scorer.** Phase 6 (`v0.6.0`). Cannot precede Phase 5.
- **Monge-Elkan with phonetic inner metric** depends on phonetic algorithms shipping before or with token-based. Spec phasing (Phase 3 token-based, Phase 4 phonetic) means Monge-Elkan ships in Phase 3 but only gains the phonetic-inner option in Phase 4. Either: (a) ship Monge-Elkan in Phase 4, or (b) ship in Phase 3 with non-phonetic-inner only, then expand the permitted-inner-list in Phase 4. **Decision needed.** Recommend (b): ship Monge-Elkan in Phase 3 with Jaro-Winkler default + permit any non-phonetic algorithm as inner, then expand to phonetic-as-inner in Phase 4. Phase 4 becomes "phonetic algorithms + Monge-Elkan inner expansion".
- **Token-bucket optimisation in scan** uses `Tokenise` directly — depends on tokenisation being feature-complete in Phase 2 (which spec confirms).
- **Cross-platform determinism CI matrix** can be wired in Phase 1 (bootstrap); doesn't depend on any algorithm being present. Recommend wiring early.

---

## MVP Definition

### Launch With (v0.1.0 — Phase 2 first cut)

Spec §19 Phase 2 first cut (`v0.1.0`):

- [ ] **Levenshtein distance + normalised score** (`*Distance`, `*Score`, `*Runes` variants) — baseline of any fuzzy library
- [ ] **Damerau-Levenshtein OSA** — primary identifier-typo metric
- [ ] **Damerau-Levenshtein Full** — correctness completeness
- [ ] **Hamming** — fixed-width-code matching, trivial
- [ ] **Jaro** — name-matching baseline
- [ ] **Jaro-Winkler** — prefix-aware identifier families
- [ ] **`Normalise` + `NormalisationOptions` + `DefaultNormalisationOptions`** — foundational
- [ ] **`Tokenise`** — required by Phase 3 algorithms; foundational
- [ ] **`AlgoID` enum + `AlgoIDs()` + `AlgoID.String()`** — required by Scorer in Phase 5
- [ ] **Sentinel errors (`errors.go`)** — required by all phases
- [ ] **Cross-platform determinism CI** — wire early; cheap
- [ ] **CI scaffolding** (lint, vet, test, race, fuzz-short, coverage, vulncheck, license-check) — from bootstrap
- [ ] **README quick-start, godoc examples for each algorithm, `llms.txt` skeleton**

### Add After Validation (v0.2.0 — v0.5.0)

- [ ] **v0.2.0**: Strcmp95, Smith-Waterman-Gotoh, LCSStr, Ratcliff-Obershelp — completing the character-based catalogue
- [ ] **v0.3.0**: Q-Gram Jaccard, Sørensen-Dice, Cosine, Tversky, Monge-Elkan, TokenSort, TokenSet, Partial, TokenJaccard — q-gram and token-based catalogue
- [ ] **v0.4.0**: Soundex, Double Metaphone, NYSIIS, MRA — phonetic catalogue, plus Monge-Elkan-with-phonetic-inner expansion
- [ ] **v0.5.0**: `Scorer` + `NewScorer` + `DefaultScorer` + every `With*` option + `Score` / `ScoreAll` / `Match` / `Threshold` / `Algorithms` methods + BDD scenarios
- [ ] **v0.6.0**: `scan` sub-package — `Item` / `Warning` / `Config` / `Check`, token-bucket, suppression composition, BDD scenarios
- [ ] **v0.6.x**: integration shakedown via `axonops/audit` consumer; surface and fix API ergonomic issues

### v1.0.0 Cut

- [ ] All 23 algorithms shipped, tested, benchmarked
- [ ] Scorer + scan sub-package shipped, tested, benchmarked
- [ ] Determinism CI matrix green on all 5 platforms
- [ ] Coverage targets met (95% / 90% / 100% public API)
- [ ] All `bench.txt` budgets met
- [ ] `axonops/audit` integration green
- [ ] Documentation complete (README, `docs/*`, godoc, `llms.txt`, `llms-full.txt`)

### Future Consideration (v1.x+ / v2)

- [ ] **Unicode normalisation in pipeline** (NFC/NFD/NFKC, diacritic stripping) — flagged as spec gap; resolve before or in v1.x
- [ ] **`fuzzymatch/search` sub-package** for one-to-many search (RapidFuzz `process.extract` equivalent) — gated on consumer demand
- [ ] **`Scorer.Explain(a, b)`** — richer explainability beyond `ScoreAll`
- [ ] **`Scorer.ScoreNormalised(a, b)`** — hot-path optimisation skipping per-call normalisation
- [ ] **`Scorer.Equal(other)` / `Scorer.Fingerprint()`** — config comparison / caching support
- [ ] **`iter.Seq[Warning]` variant of `scan.Check`** — Go 1.23+ iterator idiom, gated on community adoption
- [ ] **`ScanParallel` opt-in variant** — parallel scan trading determinism for wall-clock — gated on benchmark demand
- [ ] **Domain-specific Scorer presets** beyond `DefaultScorer()` — e.g. `ScorerForNames()`, `ScorerForCodeSymbols()`, `ScorerForFileNames()` — gated on consumer use cases
- [ ] **`fuzzymatch/process` sub-package** — RapidFuzz-style processor utilities (canonicalisation pipelines, ratio strategies) — gated on demand
- [ ] **Algorithm-specific optimisations** flagged as v1.x in spec: Ukkonen banding for Levenshtein-with-threshold, Hyyrö bit-parallel Damerau-Levenshtein, sliding-window Partial Ratio

---

## Sources

- **`docs/requirements.md`** — authoritative spec, all 21 sections (1812 lines). Primary input.
- **`docs/prior-art-research.md`** — Go ecosystem survey, 21-algorithm Joensuu study, MIT-license reference implementations. Primary input.
- **`PROJECT.md`** — project context, OOS reasoning, key decisions.
- **`CLAUDE.md`** — agent gate workflow, algorithm correctness discipline.
- [adrg/strutil GitHub](https://github.com/adrg/strutil) — interface-based architecture reference (7 algorithms; `StringMetric` interface; inactive in 2024+).
- [adrg/strutil package docs](https://pkg.go.dev/github.com/adrg/strutil) — `Similarity(a, b, metric)` shape; verifies the standalone-algorithm-function-plus-composer pattern is established.
- [hbollon/go-edlib GitHub](https://github.com/hbollon/go-edlib) — comprehensive single-library reference (10+ algorithms; `StringsSimilarity(a, b, algoEnum)`; 100% test coverage).
- [hbollon/go-edlib package docs](https://pkg.go.dev/github.com/hbollon/go-edlib) — exposes `FuzzySearch` and `FuzzySearchThreshold`; confirms that one-to-many search is the most-requested feature in the Go ecosystem.
- [rapidfuzz/RapidFuzz GitHub](https://github.com/rapidfuzz/RapidFuzz) — cross-language reference for API ergonomics; `fuzz`/`distance`/`process` namespacing; `score_cutoff` / `processor` parameters.
- [RapidFuzz 3.14.5 documentation](https://rapidfuzz.github.io/RapidFuzz/) — canonical modern reference for Token Sort/Set/Partial Ratio shapes (cited in spec §7.3).
- [RapidFuzz fuzz module docs](https://rapidfuzz.github.io/RapidFuzz/Usage/fuzz.html) — verifies token set ratio / partial ratio semantics.
- [closestmatch GitHub](https://github.com/schollz/closestmatch) — bag-of-words / n-gram pre-computation approach; confirms one-to-many search is a distinct problem class with optimised approaches.
- [sahilm/fuzzy GitHub](https://github.com/sahilm/fuzzy) — filename / code-symbol fuzzy matching; different niche (interactive-typing matching, not similarity scoring).
- [lithammer/fuzzysearch GitHub](https://github.com/lithammer/fuzzysearch) — Levenshtein + ranking; small surface.
- [paul-mannino/go-fuzzywuzzy package docs](https://pkg.go.dev/github.com/paul-mannino/go-fuzzywuzzy) — Go port of FuzzyWuzzy; confirms TokenSetRatio / TokenSortRatio nomenclature is established in Go.
- [Go blog — Text normalization in Go](https://go.dev/blog/normalization) — Unicode normalisation requires `golang.org/x/text/unicode/norm` (non-stdlib); relevant to the NFC/NFD spec gap.
- [Unicode UTR #30 — Character Foldings](https://www.unicode.org/reports/tr30/tr30-1.html) — case-folding spec; verifies that case-fold ≠ Unicode-normalise.
- [Unicode Normalization NFC NFD NFKC NFKD Guide — SymbolFYI](https://symbolfyi.com/guides/unicode-normalization-guide/) — verifies the four normalisation forms and their use cases.
- [Approximate string matching — Wikipedia](https://en.wikipedia.org/wiki/Approximate_string_matching) — domain taxonomy reference.
- [Similarity Measures for Title Matching — Fränti et al. 2016 (Joensuu study)](https://www.cs.joensuu.fi/sipu/pub/TitleSimilarity-ICPR.pdf) — empirical ranking of 21 measures cited in spec for algorithm-coverage decisions.
- [`xrash/smetrics` package docs](https://pkg.go.dev/github.com/xrash/smetrics) — Jaro / Jaro-Winkler / Soundex / Hamming / Strcmp95 / Ukkonen reference; Debian-packaged.
