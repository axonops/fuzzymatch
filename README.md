# fuzzymatch

> Fuzzy name matching for Go services — string similarity, suppression, zero runtime dependencies.

[![License](https://img.shields.io/badge/license-Apache_2.0-blue.svg)](LICENSE)
[![Status](https://img.shields.io/badge/status-pre--release-orange.svg)](#status)

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
**Runtime dependencies:** stdlib only. Zero external dependencies. No cgo.

---

## Three layers

```
Layer 1: Algorithm functions      LevenshteinScore(a, b)            ─┐
Layer 2: Scorer                   NewScorer().Score(a, b)            │  Same library,
Layer 3: Scan sub-package         fuzzymatch/scan.Check(items, cfg) ─┘  three depths.
```

Consumers pick the layer that matches their question:

- **"How similar are these two strings?"** → Layer 1 (one algorithm function).
- **"How similar are these two strings overall?"** → Layer 2 (weighted composite via `Scorer`).
- **"Which pairs in this collection are similar?"** → Layer 3 (`scan.Check`).

---

## Algorithms

Twenty-three string-similarity algorithms across four categories:

- **Character-based (9):** Levenshtein, Damerau-Levenshtein (OSA + Full), Hamming, Jaro, Jaro-Winkler, Strcmp95, Smith-Waterman-Gotoh, LCSStr.
- **Q-gram / n-gram (4):** Q-Gram Jaccard, Sørensen-Dice, Cosine (n-gram), Tversky.
- **Token-based (5):** Monge-Elkan, Token Sort Ratio, Token Set Ratio, Partial Ratio, Token Jaccard.
- **Phonetic (4):** Soundex, Double Metaphone, NYSIIS, MRA.
- **Gestalt (1):** Ratcliff-Obershelp.

Each implemented fresh from primary academic sources. Each cited inline. Each tested against literature reference vectors.

Metaphone 3 is explicitly **not** included due to U.S. Patent 7,440,941. See `docs/faq.md` (when published) for the full patent screen.

---

## Documentation

This is pre-release. Documentation is in active development.

- [`docs/requirements.md`](docs/requirements.md) — the authoritative spec for what this library does.
- `docs/algorithms.md` — algorithm-by-algorithm reference (forthcoming).
- `docs/scorer.md` — `Scorer` configuration and tuning (forthcoming).
- `docs/scan.md` — `scan` sub-package consumer guide (forthcoming).
- `docs/tuning.md` — threshold tuning and calibration (forthcoming).
- `docs/extending.md` — adding a custom algorithm (forthcoming).
- `docs/performance.md` — benchmark numbers and optimisation notes (forthcoming).
- `docs/faq.md` — common questions, exclusions, and rationale (forthcoming).

---

## 🤖 For AI Assistants

When `llms.txt` and `llms-full.txt` are published, they will live at the repo root and stay in sync with the public API.

Until then, AI assistants working on this repo should consult `CLAUDE.md` (project conventions) and `docs/requirements.md` (authoritative spec).

This project is built with [GSD](https://github.com/gsd-build/get-shit-done) for spec-driven development. Domain-specific review agents in `.claude/agents/` gate every change. See `.claude/skills/fuzzymatch-review-protocol/SKILL.md` for the review protocol.

---

## Contributing

Pre-release. External contributions welcome once `v1.0.0` ships. Until then, please file issues for discussion rather than PRs.

---

## License

Apache-2.0. See [`LICENSE`](LICENSE) and [`NOTICE`](NOTICE).
