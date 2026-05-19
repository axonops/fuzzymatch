# fuzzymatch

> Fuzzy name matching for Go services — string similarity, suppression, zero runtime dependencies.

[![License](https://img.shields.io/badge/license-Apache_2.0-blue.svg)](LICENSE)
[![Status](https://img.shields.io/badge/status-pre--release-orange.svg)](#-status)
[![Go Reference](https://pkg.go.dev/badge/github.com/axonops/fuzzymatch.svg)](https://pkg.go.dev/github.com/axonops/fuzzymatch)
[![Go Report Card](https://goreportcard.com/badge/github.com/axonops/fuzzymatch)](https://goreportcard.com/report/github.com/axonops/fuzzymatch)

---

## Table of contents

- [Status](#-status)
- [What this is](#what-this-is)
- [Key features](#key-features)
- [Why this library exists](#why-this-library-exists)
- [Three layers](#three-layers)
- [Quick start](#quick-start)
- [Algorithm catalogue](#algorithm-catalogue)
- [Configuration](#configuration)
- [Thread safety](#thread-safety)
- [API reference](#api-reference)
- [Documentation](#documentation)
- [For AI Assistants](#-for-ai-assistants)
- [Contributing](#contributing)
- [Security](#security)
- [License](#license)

---

## ⚠ Status

This library is **pre-release**. The API is not yet stable and may change without notice until the `v1.0.0` tag. Do not use in production until the first stable release ships.

See [`docs/requirements.md`](docs/requirements.md) for the authoritative specification of what this library will do.

---

## What this is

A pure-Go library detecting pairs of similar names in a collection. Fuzzy matching for "these two probably mean the same thing" cases that humans miss when authoring schemas, taxonomies, configuration vocabularies, API field sets, database column lists, environment variable names, CLI flag sets, and any other structured naming domain.

The library is domain-agnostic. It knows about strings, weights, and thresholds — not about YAML, taxonomies, or any specific format. Consumers translate their own data into the library's generic types and process the warnings in whatever way fits their domain.

**Module path:** `github.com/axonops/fuzzymatch`
**License:** Apache-2.0
**Go version:** 1.26.3 minimum
**Runtime dependencies:** stdlib + a single curated dep (`golang.org/x/text` for Unicode normalisation). No other runtime deps. No cgo.

---

## Key features

- **Twenty-three string-similarity algorithms** across five categories: character-based, q-gram, token-based, phonetic, gestalt.
- **Fresh implementations from primary academic sources.** Every algorithm cites its originating paper inline; no GPL/LGPL-derived code; no patent-encumbered algorithms (Metaphone 3 is explicitly excluded).
- **Weighted composite `Scorer`** for mixing algorithms with caller-controlled weights and threshold.
- **Collection-scan sub-package** for one-shot deduplication passes.
- **Input-quality diagnostics** via [`fuzzymatch.Validate`](docs/algorithms.md#input-validation-with-fuzzymatchvalidate) — a pure, non-panicking function that surfaces problematic-but-non-fatal input shapes (empty input, unequal length for Hamming, no tokens after normalisation, all-non-ASCII for ASCII-only algorithms, pathologically large input) as a sorted `[]Warning` slice.
- **Cross-platform deterministic output** — verified byte-identical across linux/amd64, linux/arm64, darwin/amd64, darwin/arm64, windows/amd64 via golden-file tests.
- **Pure-function library.** No goroutines, no channels, no I/O, no config files, no background work.
- **Property-tested and fuzz-tested.** Mathematical invariants (symmetry, identity, range bounds, triangle inequality where applicable) verified via `testing/quick`; every public function has a native Go fuzzer.
- **Apache-2.0 throughout.** Compatible with the BSD-3-Clause licence of RE2 and the various MIT-licensed prior-art Go implementations consulted for reference vectors only.

---

## Why this library exists

Existing Go fuzzy-matching libraries fall into three buckets:

1. **Single-algorithm packages** (`adrg/strutil`, `hbollon/go-edlib`, `xrash/smetrics`) — useful but require the consumer to assemble a multi-algorithm strategy by hand.
2. **Python ports** (`fuzzywuzzy` / `rapidfuzz` ports) — token-style only, no phonetic algorithms, often MIT-derived which the consumer must vet.
3. **Heavyweight ML matchers** — depend on embedding models, slow, hard to verify, hard to audit.

fuzzymatch is the missing fourth bucket: a curated, audit-friendly catalogue of pure-Go algorithms tied together by a weighted `Scorer` and a turnkey `scan` layer. Everything is determinism-first, allocation-budgeted, and licence-clean.

---

## Three layers

```text
Layer 1: Algorithm functions      LevenshteinScore(a, b)            ─┐
Layer 2: Scorer                   NewScorer().Score(a, b)            │  Same library,
Layer 3: Scan sub-package         fuzzymatch/scan.Check(items, cfg) ─┘  three depths.
```

Consumers pick the layer that matches their question:

- **"How similar are these two strings?"** → Layer 1 (one algorithm function).
- **"How similar are these two strings overall?"** → Layer 2 (weighted composite via `Scorer`).
- **"Which pairs in this collection are similar?"** → Layer 3 (`scan.Check`).

---

## Quick start

The opinionated default `Scorer` composes six algorithms at equal weight with a baked-in threshold of `0.85`. One line constructs it; another decides whether two strings are similar.

```go
package main

import (
    "fmt"

    "github.com/axonops/fuzzymatch"
)

func main() {
    s := fuzzymatch.DefaultScorer()

    fmt.Println(s.Match("user_id", "userId"))
    // Output: true
}
```

`DefaultScorer()` cannot fail. See [`docs/scorer.md`](docs/scorer.md) for the composition and [`docs/tuning.md`](docs/tuning.md) for the threshold-calibration loop.

For lower-level primitives — `Normalise`, `Tokenise`, and the 23 individual algorithm functions — see [Common Patterns](#common-patterns) below and the runnable programmes at [`examples/identifier-similarity/`](examples/identifier-similarity/main.go) and [`examples/scorer-composition/`](examples/scorer-composition/main.go).

## Common Patterns

### Validate-then-Score: audit input quality before scoring

`fuzzymatch.Validate(a, b)` surfaces problematic-but-non-fatal input shapes as a sorted `[]Warning`. It is pure, never panics, and returns `nil` when no warnings apply. Use it on code paths that audit input quality — typically log lines, telemetry, or pre-flight checks ahead of `Scorer.Score`.

<!-- docs:skip-compile — illustrative snippet; full programme in examples/validate-input-quality/ -->
```go
package main

import (
    "fmt"
    "log"

    "github.com/axonops/fuzzymatch"
)

func main() {
    a, b := "user_id", "" // empty input — degenerate case

    for _, w := range fuzzymatch.Validate(a, b) {
        log.Printf("input-quality warning: %s (%s): %s",
            w.Kind, w.Algorithm, w.Detail)
    }

    score := fuzzymatch.DefaultScorer().Score(a, b)
    fmt.Printf("score = %.4f\n", score)
}
```

The full runnable programme is at [`examples/validate-input-quality/`](examples/validate-input-quality/main.go). See [`docs/algorithms.md#input-validation-with-fuzzymatchvalidate`](docs/algorithms.md#input-validation-with-fuzzymatchvalidate) for the per-`WarnKind` semantics and the [Panic surface](docs/algorithms.md#panic-surface) section for the typed-panic discipline (`Validate` itself never panics).

---

## Algorithm catalogue

Twenty-three algorithms in five categories. Every entry has its primary academic source cited in [`docs/algorithms.md`](docs/algorithms.md) and in the implementation file once it lands (Phase 2+).

### Character-based (9)

| Algorithm | `AlgoID` | Primary source | Detail |
|-----------|----------|----------------|--------|
| Levenshtein | `AlgoLevenshtein` | Levenshtein 1965 | [docs/algorithms.md#levenshtein](docs/algorithms.md#levenshtein) |
| Damerau-Levenshtein (OSA) | `AlgoDamerauLevenshteinOSA` | Boytsov 2011; Damerau 1964 | [docs/algorithms.md#damerau-levenshtein-osa](docs/algorithms.md#damerau-levenshtein-osa) |
| Damerau-Levenshtein (Full) | `AlgoDamerauLevenshteinFull` | Lowrance & Wagner 1975 | [docs/algorithms.md#damerau-levenshtein-full](docs/algorithms.md#damerau-levenshtein-full) |
| Hamming | `AlgoHamming` | Hamming 1950 | [docs/algorithms.md#hamming](docs/algorithms.md#hamming) |
| Jaro | `AlgoJaro` | Jaro 1989 | [docs/algorithms.md#jaro](docs/algorithms.md#jaro) |
| Jaro-Winkler | `AlgoJaroWinkler` | Winkler 1990 | [docs/algorithms.md#jaro-winkler](docs/algorithms.md#jaro-winkler) |
| Strcmp95 | `AlgoStrcmp95` | Winkler 1994; U.S. Census 1995 | [docs/algorithms.md#strcmp95](docs/algorithms.md#strcmp95) |
| Smith-Waterman-Gotoh | `AlgoSmithWatermanGotoh` | Smith & Waterman 1981; Gotoh 1982 | [docs/algorithms.md#smith-waterman-gotoh](docs/algorithms.md#smith-waterman-gotoh) |
| LCSStr | `AlgoLCSStr` | Wagner & Fischer 1974 | [docs/algorithms.md#lcsstr](docs/algorithms.md#lcsstr) |

### Q-gram / n-gram (4)

| Algorithm | `AlgoID` | Primary source | Detail |
|-----------|----------|----------------|--------|
| Q-Gram Jaccard | `AlgoQGramJaccard` | Ukkonen 1992; Jaccard 1912 | [docs/algorithms.md#q-gram-jaccard](docs/algorithms.md#q-gram-jaccard) |
| Sørensen-Dice | `AlgoSorensenDice` | Sørensen 1948; Dice 1945 | [docs/algorithms.md#sørensen-dice](docs/algorithms.md#sørensen-dice) |
| Cosine (n-gram) | `AlgoCosine` | Salton & McGill 1983 | [docs/algorithms.md#cosine](docs/algorithms.md#cosine) |
| Tversky | `AlgoTversky` | Tversky 1977 | [docs/algorithms.md#tversky](docs/algorithms.md#tversky) |

### Token-based (5)

| Algorithm | `AlgoID` | Primary source | Detail |
|-----------|----------|----------------|--------|
| Monge-Elkan | `AlgoMongeElkan` | Monge & Elkan 1996 | [docs/algorithms.md#monge-elkan](docs/algorithms.md#monge-elkan) |
| Token Sort Ratio | `AlgoTokenSortRatio` | SeatGeek fuzzywuzzy / RapidFuzz | [docs/algorithms.md#token-sort-ratio](docs/algorithms.md#token-sort-ratio) |
| Token Set Ratio | `AlgoTokenSetRatio` | SeatGeek fuzzywuzzy / RapidFuzz | [docs/algorithms.md#token-set-ratio](docs/algorithms.md#token-set-ratio) |
| Partial Ratio | `AlgoPartialRatio` | SeatGeek fuzzywuzzy / RapidFuzz | [docs/algorithms.md#partial-ratio](docs/algorithms.md#partial-ratio) |
| Token Jaccard | `AlgoTokenJaccard` | Jaccard 1912 | [docs/algorithms.md#token-jaccard](docs/algorithms.md#token-jaccard) |

### Phonetic (4)

| Algorithm | `AlgoID` | Primary source | Detail |
|-----------|----------|----------------|--------|
| Soundex | `AlgoSoundex` | Russell 1918; Knuth 1973 | [docs/algorithms.md#soundex](docs/algorithms.md#soundex) |
| Double Metaphone | `AlgoDoubleMetaphone` | Philips 2000 | [docs/algorithms.md#double-metaphone](docs/algorithms.md#double-metaphone) |
| NYSIIS | `AlgoNYSIIS` | Taft 1970 | [docs/algorithms.md#nysiis](docs/algorithms.md#nysiis) |
| MRA | `AlgoMRA` | Moore et al. 1977 (NBS TN 943) | [docs/algorithms.md#mra](docs/algorithms.md#mra) |

### Gestalt (1)

| Algorithm | `AlgoID` | Primary source | Detail |
|-----------|----------|----------------|--------|
| Ratcliff-Obershelp | `AlgoRatcliffObershelp` | Ratcliff & Metzener 1988 | [docs/algorithms.md#ratcliff-obershelp](docs/algorithms.md#ratcliff-obershelp) |

**Metaphone 3 is explicitly NOT included** due to U.S. Patent 7,440,941. See [`docs/faq.md`](docs/faq.md#why-no-metaphone-3) for the full patent screen rationale.

---

## Configuration

The Phase-1 primitives expose two option structs. Both are passed by value, both have a `Default…Options()` constructor, and both are immutable inputs (callers building variant configurations construct fresh values).

```go
// Normalisation with strict ASCII casing + diacritic stripping for "café → cafe".
opts := fuzzymatch.DefaultNormalisationOptions()
opts.StripDiacritics = true
n := fuzzymatch.Normalise("Café Müller", opts)
// n == "cafe muller"

// Tokenisation with default split rules (camelCase, snake_case, kebab-case, dot-case).
tokens := fuzzymatch.Tokenise("User-CreateEvent.v2", fuzzymatch.DefaultTokeniseOptions())
// tokens == []string{"user", "create", "event", "v2"}
```

`NormalisationOptions` fields: `Lowercase`, `StripSeparators`, `SeparatorChars`, `SplitCamelCase`, `NFC`, `StripDiacritics`.
`TokeniseOptions` fields: `Lowercase`, `SplitCamelCase`, `SplitConsecutiveUpper`, `SeparatorChars`.

The `Scorer` accepts a `NormalisationOptions` value at construction time and applies it before every algorithm invocation.

See [`docs/tuning.md`](docs/tuning.md) for guidance on calibrating algorithm weights and thresholds against a domain corpus, and [`docs/scorer.md`](docs/scorer.md) for the `Scorer` API reference.

---

## Thread safety

Every public function in the root package is **pure**: no shared mutable state, no goroutines, no channels, no mutexes. Concurrent callers may invoke `Normalise`, `Tokenise`, `Validate`, and every algorithm score function from any number of goroutines without coordination.

The `Scorer` is **immutable after construction**. A constructed `Scorer` is safe for concurrent use; callers wanting a different configuration construct a fresh `Scorer`.

The `scan` sub-package follows the same discipline: a constructed `scan.Config` is immutable; `scan.Check` is safe for concurrent invocation on disjoint inputs.

---

## API reference

The canonical API reference lives on pkg.go.dev: [pkg.go.dev/github.com/axonops/fuzzymatch](https://pkg.go.dev/github.com/axonops/fuzzymatch).

Every exported type, function, method, and constant carries a godoc comment that begins with the symbol name. Algorithm implementation files cite their primary academic source inline at the top of the file.

---

## Documentation

- [`docs/requirements.md`](docs/requirements.md) — the authoritative spec for what this library does.
- [`docs/algorithms.md`](docs/algorithms.md) — algorithm-by-algorithm reference (per-algorithm detail fills in as each phase lands).
- [`docs/scorer.md`](docs/scorer.md) — `Scorer` configuration and tuning.
- [`docs/scan.md`](docs/scan.md) — `scan` sub-package consumer guide.
- [`docs/best-practices.md`](docs/best-practices.md) — production patterns including the Validate-then-Score idiom.
- [`docs/tuning.md`](docs/tuning.md) — threshold tuning and calibration.
- [`docs/extending.md`](docs/extending.md) — adding a custom algorithm.
- [`docs/performance.md`](docs/performance.md) — benchmark numbers and optimisation notes.
- [`docs/faq.md`](docs/faq.md) — common questions, exclusions, and rationale.

---

## 🤖 For AI Assistants

This repository ships [`llms.txt`](llms.txt) (concise index) and [`llms-full.txt`](llms-full.txt) (full API reference + algorithm citations) at the repo root. AI assistants and code generators should consult these first.

The contents are verified in sync with the public Go API by `ai_friendly_test.go`, which parses every exported root-package symbol via `go/ast` and asserts each appears in `llms.txt`. Drift fails CI.

This project is built with [GSD](https://github.com/gsd-build/get-shit-done) for spec-driven development. Domain-specific review agents in `.claude/agents/` gate every change. See `.claude/skills/fuzzymatch-review-protocol/SKILL.md` for the review protocol.

---

## Contributing

Pre-release. External contributions welcome once `v1.0.0` ships. Until then, please file issues for discussion rather than PRs. See [`CONTRIBUTING.md`](CONTRIBUTING.md) for the local development setup, conventional-commit rules, the algorithm deprecation policy, and the release-via-CI-only discipline.

The repo's review gates are documented in [`.claude/skills/fuzzymatch-review-protocol/SKILL.md`](.claude/skills/fuzzymatch-review-protocol/SKILL.md): every change passes through algorithm-licensing, algorithm-correctness, algorithm-performance, determinism, api-ergonomics, code-reviewer, and security-reviewer agents as applicable.

---

## Security

Vulnerabilities go to the private channel documented in [`SECURITY.md`](SECURITY.md). Do NOT open public issues for security reports. Release signatures are verifiable via [`cosign`](https://github.com/sigstore/cosign) — see SECURITY.md for the verification command.

---

## License

Apache-2.0. See [`LICENSE`](LICENSE) and [`NOTICE`](NOTICE).
