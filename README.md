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
- **Weighted composite `Scorer`** for mixing algorithms with caller-controlled weights and threshold (Phase 8).
- **Collection-scan sub-package** for one-shot deduplication passes (Phase 9).
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

> **Note:** Phase 1 (foundation) ships `Normalise` and `Tokenise` primitives plus the `AlgoID` enum and sentinel errors. Algorithm functions (e.g. `LevenshteinScore`) land in Phase 2. The example below uses the Phase-1 primitives; the full algorithm-driven quick start is added with Phase 2.

```go
package main

import (
    "fmt"

    "github.com/axonops/fuzzymatch"
)

func main() {
    opts := fuzzymatch.DefaultNormalisationOptions()
    opts.StripDiacritics = true

    fmt.Println(fuzzymatch.Normalise("UserCreate-Event", opts))
    // Output: user create event

    fmt.Println(fuzzymatch.Tokenise("XMLHttpRequest", fuzzymatch.DefaultTokeniseOptions()))
    // Output: [xmlhttp request]
}
```

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

The `Scorer` (Phase 8) accepts a `NormalisationOptions` value at construction time and applies it before every algorithm invocation.

See [`docs/tuning.md`](docs/tuning.md) for guidance on calibrating algorithm weights and thresholds against a domain corpus, and [`docs/scorer.md`](docs/scorer.md) for the `Scorer` API once Phase 8 lands.

---

## Thread safety

Every public function in the root package is **pure**: no shared mutable state, no goroutines, no channels, no mutexes. Concurrent callers may invoke `Normalise`, `Tokenise`, and (from Phase 2) every algorithm score function from any number of goroutines without coordination.

The `Scorer` (Phase 8) is **immutable after construction**. A constructed `Scorer` is safe for concurrent use; callers wanting a different configuration construct a fresh `Scorer`.

The `scan` sub-package (Phase 9) follows the same discipline: a constructed `scan.Config` is immutable; `scan.Check` is safe for concurrent invocation on disjoint inputs.

---

## API reference

The canonical API reference lives on pkg.go.dev: [pkg.go.dev/github.com/axonops/fuzzymatch](https://pkg.go.dev/github.com/axonops/fuzzymatch).

Every exported type, function, method, and constant carries a godoc comment that begins with the symbol name. Algorithm implementation files cite their primary academic source inline at the top of the file.

---

## Documentation

- [`docs/requirements.md`](docs/requirements.md) — the authoritative spec for what this library does.
- [`docs/algorithms.md`](docs/algorithms.md) — algorithm-by-algorithm reference (per-algorithm detail fills in as each phase lands).
- [`docs/scorer.md`](docs/scorer.md) — `Scorer` configuration and tuning (Phase 8).
- [`docs/scan.md`](docs/scan.md) — `scan` sub-package consumer guide (Phase 9).
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
