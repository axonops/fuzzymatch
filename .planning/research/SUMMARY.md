# Project Research Summary

**Project:** fuzzymatch — standalone Go string-similarity library
**Domain:** Pure-function library: 23 academic-source algorithms + weighted composite Scorer + optional collection-scan sub-package
**Researched:** 2026-05-13
**Confidence:** HIGH

Synthesis of four parallel research streams. Detailed findings live in `STACK.md`, `FEATURES.md`, `ARCHITECTURE.md`, `PITFALLS.md` — this summary is the load-bearing input for the roadmapper.

---

## Executive Summary

`fuzzymatch` is a correctness-first pure-Go library shipping a comprehensive algorithm catalogue (23 academic-source algorithms) under a coherent three-layer API (algorithm functions → Scorer → scan), with zero runtime dependencies, no cgo, Apache-2.0, and cross-platform byte-identical determinism as a hard guarantee. The combination of (a) full catalogue, (b) typed-AlgoID composable Scorer, (c) explicit cross-platform determinism, (d) production-grade DX (godoc examples, llms.txt, BDD), and (e) zero-deps / no-cgo simultaneously is unique in the Go ecosystem as of 2026 — every dimension exists elsewhere but no library combines them. The recommended approach mirrors `axonops/mask` DX patterns broadly (GoReleaser v2 + Cosign v3 keyless + SBOM + OIDC attestations, CLA Assistant, conventional-commit linting, `llms.txt`) while applying strictly tighter discipline: no testify in root tests, separate `tests/bdd/go.mod`, structural verification of zero runtime deps in CI.

The single largest correctness risk is **cross-platform float determinism** — nine of the twenty pitfalls inventoried are gated by foundation-phase test infrastructure (cross-platform CI matrix, golden files, property tests, benchstat, coverage enforcement, release workflow). Without that scaffolding, every subsequent algorithm phase carries hidden risk: a Cosine implementation passing on amd64 quietly diverges on arm64; a token-bucket scan output is deterministic locally but reorders on Windows; a score-changing edit ships in a patch release and silently breaks consumer suppression lists. The roadmap **must** land foundation infrastructure before the first algorithm. Two distinct algorithm-specific risks deserve roadmap-level visibility: (1) Gotoh 1982's published affine-gap recurrence contains a known erratum — Smith-Waterman-Gotoh requires explicit EMBOSS/biopython cross-validation, not just primary-source citation; (2) Soundex has four canonical variants with the discriminating reference vector `"Tymczak" → "T522"` (Knuth/Census) vs `"T520"` (SQL).

The spec (`docs/requirements.md`) is **authoritative and largely spec-locked**: the 8-phase §19 release plan, the three-layer architecture, the constraints (Go 1.26+, zero deps, no cgo, Apache-2.0, no testify in root, separate BDD module, determinism matrix, CI-only releases), and the 23-algorithm catalogue (with Metaphone 3 / Needleman-Wunsch / Soft-TFIDF excluded for documented reasons) are constraints, not options. The roadmapper should treat these as load-bearing. What is **open** and needs roadmap- or PROJECT.md-level decision: (a) one Unicode-normalisation feature gap in the `Normalise` pipeline (NFC/NFD + diacritic stripping); (b) whether to bring the Scorer forward from Phase 5 to Phase 3/4 for earlier `axonops/audit` integration testing; (c) the Monge-Elkan phasing wrinkle (it composes an inner algorithm by AlgoID — Phase 3 ships it with non-phonetic inners, Phase 4 expands to phonetic inners); (d) CLA-vs-DCO confirmation (recommendation: CLA, mirroring mask); (e) commitlint tooling (recommendation: `wagoid/commitlint-github-action`); (f) golangci-lint v2 + GoReleaser v2 (recommendation: yes — both are current major; v1 lines are end-of-life).

---

## Key Findings

### Recommended Stack

Pure-Go on Go 1.26.3 with zero non-stdlib `require` lines in the root module. Test-only deps (`godog`, `goleak`, `testify`) isolated in `tests/bdd/go.mod` so consumers never transitively pull them. Quality and release tooling is the v2 generation throughout: golangci-lint v2.12.2, GoReleaser v2.15.4, Cosign v3.0.1 (keyless default, `--bundle` now required), `actions/checkout@v6` + `actions/setup-go@v6` (Node 24 runtime). Determinism plumbing — cross-platform CI matrix + golden files + benchstat with `-count=10` + cosign keyless + OIDC attestations + Syft SBOMs — is mandatory infrastructure, not optional polish.

**Core technologies (all spec-locked except where noted):**
- **Go 1.26.3** — minimum 1.26+ per `docs/requirements.md` §1; 1.26.3 is the current security-patch (2026-05-07); Green Tea GC default benefits short-string allocation patterns. `CGO_ENABLED=0` enforced in CI.
- **Stdlib `testing` + `testing/quick` + native `go test -fuzz`** — root module zero non-stdlib deps; `testing/quick` for mathematical-invariant property tests; native fuzz since Go 1.18.
- **`tests/bdd/go.mod` sub-module** — `cucumber/godog v0.15.0` + `go.uber.org/goleak v1.3.0` + `stretchr/testify v1.10.0`; testify permitted ONLY in BDD module.
- **golangci-lint v2.12.2** — v2 config layout (`linters.default` + dedicated `formatters:` block); run `golangci-lint migrate` to port any legacy templates.
- **GoReleaser v2.15.4 + `goreleaser/goreleaser-action@v7`** — even though library ships no binaries, GoReleaser generates `checksums.txt`, drives cosign signing, publishes the GitHub Release; `version: "~> v2"` pin.
- **Sigstore Cosign v3.0.1 keyless** — sign `checksums.txt` via `cosign sign-blob --yes --bundle` with `--oidc-issuer=https://token.actions.githubusercontent.com`.
- **GitHub Artifact Attestation (`actions/attest-build-provenance@v2`) + Syft SBOMs** — both spec-locked in §17.1.
- **govulncheck + gosec v2.25.0 + CodeQL v4 + markdownlint-cli2 v0.22.1 + benchstat (`-count=10`, > 10% regression at p<0.05 fails)** — all spec-locked CI gates.
- **Dependabot** — daily on `gomod` (root + `tests/bdd`) and `github-actions`.

**Open stack decisions** (roadmapper or user picks; recommendations italicised):
- *CLA vs DCO* — recommend **CLA Assistant Lite** (mirrors mask).
- *Conventional-commit lint tooling* — recommend **`wagoid/commitlint-github-action`** (Node runs only in CI).
- *go.mod toolchain pin* — recommend `go 1.26.3` directive, omit `toolchain` line.

### Expected Features

The Go fuzzy-matching ecosystem (`agnivade/levenshtein`, `xrash/smetrics`, `adrg/strutil`, `hbollon/go-edlib`, `jcoruiz/strsim`) is fragmented across single-purpose or small-catalogue libraries with inconsistent maintenance. RapidFuzz (Python) is the cross-language reference for ergonomic shape. No Go library in 2026 combines (a) full 23-algorithm catalogue, (b) typed composable Scorer, (c) explicit determinism, (d) production DX, (e) zero deps + Apache-2.0 + no cgo. fuzzymatch's value is the combination, not any single dimension. Spec coverage of table stakes is comprehensive; five spec gaps deserve explicit roadmap consideration.

**Must have (table stakes — all spec-confirmed):**
- Levenshtein, Damerau-Levenshtein OSA, Jaro / Jaro-Winkler, Hamming, Q-Gram Jaccard, Sørensen-Dice, Cosine, Token Sort/Set/Partial Ratio, Soundex — universally expected.
- Composable scoring API with score normalised to `[0.0, 1.0]`.
- Case-folding + separator-stripping + camelCase/snake_case/PascalCase tokenisation.
- Pure functions, concurrent-safe Scorer (immutable after construction), functional options.
- Identity / symmetry / range / triangle-inequality invariants verified by property tests.
- Empty-input, identical-input, Unicode, malformed-UTF-8 (no-panic) handling.
- Literature reference vectors in unit tests; race-clean (`go test -race`); godoc with runnable examples.

**Should have (differentiators — all spec-confirmed unless flagged):**
- The full 23-algorithm catalogue under one Apache-2.0 module (no Go library currently ships this many).
- Damerau-Levenshtein FULL (Lowrance-Wagner) variant alongside OSA — rare in Go.
- Smith-Waterman-Gotoh with configurable affine gap penalty; Tversky (asymmetric); Monge-Elkan with pluggable inner metric; Token Sort/Set/Partial via Indel-ratio (RapidFuzz semantics, NOT difflib); four phonetic algorithms (Soundex, Double Metaphone, NYSIIS, MRA); Ratcliff-Obershelp.
- Weighted composite Scorer with auto-normalised weights (sum-to-1 invariant); `DefaultScorer()` opinionated default; `ScoreAll` per-algorithm breakdown for tuning; `Match` threshold helper.
- Documented cross-platform byte-identical output; algorithm score stability across patch versions pinned by golden files.
- `scan` sub-package: within-group + cross-group passes with separate thresholds; per-item `SilenceLint` + global `SuppressedPairs` composing; cross-group identical-name suppression default; token-bucket optimisation property-test-verified equivalent to naive; deterministic output ordering by `(Kind, NameA, NameB, GroupA, GroupB)`.
- Zero-deps structurally verified in CI; no-cgo; `llms.txt` + `llms-full.txt` AI-friendly docs sync-verified; BDD scenarios per algorithm; meta-tests (README compiles, Makefile targets documented, etc.); cosign keyless + OIDC attestations + Syft SBOMs.

**Defer (v2+ / future) — explicit anti-features and out-of-scope:**
- `process.extract(query, choices)` one-to-many search API (RapidFuzz's most-used function; intentionally out of v1 scope; track demand from v0.6.0 — most-requested feature in comparable libraries).
- Embedding / semantic similarity, persistent cache, parallel goroutine scan, I/O, CLI tool, config-file parsing, web UI — out of scope per `docs/requirements.md` §4.
- Needleman-Wunsch (superseded by SWG), Soft-TFIDF (needs corpus), **Metaphone 3 (U.S. Patent 7440941 — declined regardless of enforcement)**.
- `Scorer.Explain()`, `Scorer.ScoreNormalised()`, `Scorer.Fingerprint()`, `iter.Seq[Warning]` variant, `ScanParallel` — v1.x candidates gated on demand.

**Spec gaps flagged for PROJECT.md / roadmap attention:**
- **Unicode NFC/NFD normalisation + diacritic stripping in `Normalise` pipeline** (MEDIUM severity) — `café` vs `cafe`, `Müller` vs `Mueller` are common audit-field-name cases the current pipeline doesn't handle. `golang.org/x/text/unicode/norm` is non-stdlib (forbidden); options are (a) implement a minimal NFC/NFD inline, (b) document as consumer responsibility with recipe, (c) accept the gap for v1.0. **Decision needed.**
- **`Tokenise` and `Normalise` outputs in determinism golden-file** — spec implies it; make it explicit so downstream token-based scores can't drift between versions.
- **`DefaultScorerOptions()` returning the `[]ScorerOption` slice** — enables "default minus algorithm X" without rebuild-from-scratch; minor ergonomic win.
- **`examples/` directory** with runnable mini-programs (audit-fields, identifier-similarity, schema-dedup) — mirrors mask; low effort.
- **No deprecation policy for algorithms** documented in CONTRIBUTING — within a major version, algorithms can be added but not removed; scoring changes need minor bump.

### Architecture Approach

Three layers with strict downward dependency: algorithms (Layer 1) → Scorer (Layer 2) → scan sub-package (Layer 3). Root package and `scan/` live in the same Go module (`github.com/axonops/fuzzymatch` + `github.com/axonops/fuzzymatch/scan`); `tests/bdd/` is a separate Go module with a local `replace` directive — no `go.work` is committed (workspaces help local dev but harm CI's declaration-drift detection). Algorithm dispatch is via typed `AlgoID` int enum + internal dispatch table (`[AlgoID]func(a, b string) float64`), explicitly NOT an `Algorithm` interface — interface dispatch boxes on hot paths and breaks the per-algorithm allocation budget. Scorer is immutable after `NewScorer` (functional options, errors deferred to construction). `scan.Check` is a pure function; the token-bucket map is for membership lookup only — output is iterated from the sorted item list and re-sorted by a complete `(Kind, NameA, NameB, GroupA, GroupB)` key.

**Major components:**
1. **Foundation primitives** (`algoid.go`, `normalise.go`, `tokenise.go`, `errors.go`) — typed enum, normalisation (case-fold + separator-strip + ASCII fast path), tokenisation (camel/snake/Pascal/kebab split), sentinel errors. No algorithm dependencies; everything else depends on these.
2. **Algorithm functions** — one file per algorithm at root level (~25 files for the 23 algorithms; q-gram extraction is shared in `q_gram.go`). Each file cites primary source in its block comment, declares constants with godoc, exposes byte and rune variants where applicable, uses two-row DP and ASCII fast paths to hit the allocation budget.
3. **Scorer** (`scorer.go`, `scorer_options.go`) — functional-options constructor, immutable after `NewScorer`, concurrent-safe, dispatch via AlgoID table. `Score` / `ScoreAll` / `Match` / `Threshold` / `Algorithms` methods. `DefaultScorer()` opinionated default.
4. **Scan sub-package** (`scan/`) — `Item` / `Config` / `Warning` / `Check`, token-bucket optimisation, suppression composition (per-item flag + pair list + cross-group identical-name), deterministic output ordering. Property test `PropCheck_BucketEquivalentToNaive` proves the bucket optimisation is correctness-preserving.
5. **Cross-platform determinism infrastructure** — `testdata/golden/algorithms.json` + `scorer-default.json` + `scan-default.json` + `normalisation.json` pinned across linux/amd64, linux/arm64, darwin/amd64, darwin/arm64, windows/amd64; property tests `PropX_DeterministicAcrossRuns` and `PropX_NoNaN/NoInf/NoNegativeZero` per algorithm and the Scorer/scan.

**Architectural patterns spec-locked:** AlgoID enum (not interface), two-row DP + stack-allocated buffers for ASCII fast paths, pure-function-top-to-bottom (no goroutines, no I/O, no globals, no non-trivial `init()`), maps used for lookup only (never for output iteration), `sort.SliceStable` with provably-complete sort keys, `math.Sqrt`/`math.Abs`/`math.Min`/`math.Max` only (no `Pow`/`Log`/`Exp`/`FMA`).

**Open architecture decisions flagged for `api-ergonomics-reviewer`:** exact function names (Score vs Similarity), exact option names (`WithAlgorithm` vs `Algorithm`), standalone `MongeElkanScore` signature (does it take `inner AlgoID` directly? — recommendation: yes with zero-value default), parameterised algorithm options (per-algorithm `WithQGramJaccardAlgorithm` vs uniform — recommendation: per-algorithm, more discoverable), `DefaultScorer()` caching. Detail in `ARCHITECTURE.md` §9.4.

**Open architecture decision flagged for ROADMAP:** whether to bring the Scorer forward from Phase 5 (`v0.5.0`) to Phase 3 (`v0.3.0`) with a partial-catalogue, to enable `axonops/audit` integration feedback sooner. Default plan (Scorer at Phase 5) is the spec's; bringing it forward is an arguable improvement. **Decision needed at roadmap time.**

### Critical Pitfalls

Twenty pitfalls inventoried in `PITFALLS.md`. Five most roadmap-shaping:

1. **9 of 20 pitfalls are foundation-phase test-infrastructure gated** (pitfalls 7, 8, 9, 10, 13, 15, 18, 19, 20: byte-vs-rune indexing, map iteration leaking to output, float reduction order across architectures, NaN/Inf/-0 leakage, full DP table allocation, score change in patch release breaking consumer suppression, local-tag accidental release, benchmark regressions without `benchstat -count=10`, coverage thresholds reported but not enforced). **Roadmap implication:** these are NOT distributable across algorithm phases — the property-test harness, golden-file infrastructure, cross-platform CI matrix, benchstat regression detection, coverage enforcement, and release-workflow guardrails must all land in the foundation phase BEFORE any algorithm is implemented. Without them, every subsequent algorithm phase carries hidden risk.

2. **Smith-Waterman-Gotoh needs primary-source AND independent cross-validation.** Gotoh 1982 has a known erratum in the affine-gap recurrence (initialisation step and an indexing flip). Multiple textbooks (Gusfield, Waterman) reproduce the error; a biorxiv survey of 31 lecture slides found 8 reproducing the bug. **Roadmap implication:** SWG cannot follow the standard "primary source + property tests" pattern — its implementation issue must explicitly require cross-validation against EMBOSS or biopython reference vectors and flag this in the file's block comment. Allocate extra time and `algorithm-correctness-reviewer` scrutiny for SWG (Phase 2.1 / `v0.2.0`).

3. **Soundex has four canonical variants** (American/Census 1880, SQL Server/MySQL, Russell, NARA) that disagree on common inputs. `"Tymczak"` → `T522` (Knuth/Census) vs `T520` (SQL) is the single discriminating reference vector. Spec commits to Knuth/Census/Russell-Odell 1918+1922 per Knuth TAOCP Vol. 3 §6.4. **Roadmap implication:** include `"Tymczak" = "T522"` and `"Ashcraft" = "Ashcroft"` as mandatory reference vectors in the Soundex issue; variant name must appear in the file's block comment.

4. **Indel-vs-difflib token-ratio ambiguity.** `fuzzywuzzy`'s `ratio()` defaults to Python's `difflib.SequenceMatcher` (Ratcliff-Obershelp) in pure-Python and to Indel `2·LCS/(|a|+|b|)` in its C extension — giving different results across paths of the same library. RapidFuzz fixed this by standardising on Indel. **Roadmap implication:** Token Sort / Token Set / Partial Ratio cross-validate against **RapidFuzz**, not fuzzywuzzy; the Indel formula must be cited explicitly in each algorithm's godoc; `RatcliffObershelpScore` is the difflib-equivalent for consumers who want that semantic.

5. **Float determinism across architectures is the single hardest correctness constraint.** Compiler-detected `x*y + z` patterns emit FMA on arm64 but typically not on amd64 without `GOAMD64=v3`. `math.Pow`, `math.Log`, `math.Exp` are not bit-identical across platforms. Parallel reductions sum in non-deterministic order. **Roadmap implication:** the Cosine, Monge-Elkan, SWG implementations and the Scorer composite are the most exposed; the cross-platform CI matrix is mandatory by the time the first float-summing algorithm lands (Jaro in Phase 2.0 / `v0.1.0`). Explicit parenthesisation `(x*y) + z` and left-to-right reduction are non-negotiable patterns from the first algorithm.

Other pitfalls worth tracking (briefly): OSA vs Full Damerau-Levenshtein must ship as distinct AlgoIDs with the `"ca"`/`"abc"` divergence reference vector (Pitfall 1); Jaro-Winkler boost threshold `0.7` + prefix cap `4` + scale `0.1` are all easy to mis-implement from Wikipedia (Pitfall 2); NYSIIS truncation at 6 and Double Metaphone multilingual rule branches need extra scrutiny (Pitfall 5); Monge-Elkan / Partial Ratio / Token Set Ratio are DoS vectors on adversarial input — document complexity (Pitfall 12); GPL/LGPL contamination via reference implementations and patent-gap on newly-added algorithms are governance pitfalls handled by `algorithm-licensing-reviewer` at issue triage (Pitfalls 16, 17).

---

## Implications for Roadmap

The spec's §19 release phasing is a sound default. The architecturally significant deviations the roadmapper should weigh: bring Scorer forward to Phase 3/4, split Phase 2 between "core six" (`v0.1.0`) and "remaining four" (`v0.2.0`) as the spec already does, and consider whether the Phase 2.1 grouping (Strcmp95 + SWG + LCSStr + Ratcliff-Obershelp in one tag) is too dense given SWG's cross-validation overhead.

### Phase 1: Foundation / Bootstrap (no release)

**Rationale:** 9 of 20 inventoried pitfalls are foundation-infrastructure gated. Every subsequent phase inherits this scaffolding; landing it last is a much larger cost.
**Delivers:** Module init (`go.mod` with zero non-stdlib requires, verified by `scripts/verify-no-runtime-deps.sh`). `tests/bdd/go.mod` sub-module with `replace` directive. Apache-2.0 + NOTICE + SECURITY.md + CODE_OF_CONDUCT.md + CONTRIBUTING.md + CHANGELOG.md (Keep-a-Changelog) + README skeleton + `.gitignore`. Every CI workflow: ci (lint + vet + test + race + fuzz-short + coverage with enforcement + vulncheck + license-check), nightly (extended fuzz, 5min/fuzzer), release (GoReleaser v2 + cosign keyless + OIDC attestation + Syft SBOM), security (gosec + govulncheck + CodeQL), codeql, dependabot. Cross-platform determinism matrix (linux/amd64, linux/arm64, darwin/amd64, darwin/arm64, windows/amd64) wired to a placeholder golden test. `.golangci.yml` (v2 layout), `.goreleaser.yml` (v2), `.markdownlint-cli2.yaml`, Makefile (`check`, `test`, `test-bdd`, `test-fuzz`, `lint`, `vet`, `fmt`, `bench`, `bench-compare`, `coverage`, `coverage-check`, `tidy`, `tidy-check`, `security`, `verify-determinism`, `verify-no-deps`, `release-check`, `clean`). CLA workflow (`contributor-assistant/github-action`). Commitlint workflow (`wagoid/commitlint-github-action`). CODEOWNERS + issue templates + PR template. `llms.txt` + `llms-full.txt` skeletons + `ai_friendly_test.go` sync-check. `bench.txt` placeholder. `examples/` directory scaffolding.
**Addresses (features):** all OSS-first DX requirements; zero-deps structural verification; cosign + OIDC + SBOM release plumbing.
**Avoids (pitfalls):** 7 (byte-vs-rune test harness), 8 (map-iteration property test harness), 9 (cross-platform CI matrix), 10 (NaN/Inf property test harness), 13 (benchmark allocation enforcement), 15 (golden-file infrastructure for score stability), 18 (release workflow + tag-source verification), 19 (benchstat `-count=10` + self-hosted runner setup), 20 (coverage-check Makefile target).
**Research flag:** SKIP — patterns are well-documented; `axonops/mask` is the structural template. Minor research only on (a) whether self-hosted bench runner is available, (b) CLA Assistant configuration specifics, (c) `actions/checkout@v6` + Node 24 runner version requirements.

### Phase 2: Core algorithms — first six (`v0.1.0`)

**Rationale:** Levenshtein is the canonical DP example — landing it correctly proves the testing infrastructure (unit + property + fuzz + bench + BDD), the citation discipline, the determinism workflow. Damerau-Levenshtein OSA + Full pair shakes out the two-variants-of-same-name pattern. Jaro + Jaro-Winkler introduces the formula-based (non-DP) shape. Hamming is trivial and a useful sanity check on the test harness.
**Delivers:** `algoid.go` (AlgoID enum + `AlgoIDs()` + `String()`), `normalise.go` (with ASCII fast path), `tokenise.go`, `errors.go` (sentinel errors), Levenshtein + Damerau-Levenshtein OSA + Damerau-Levenshtein Full + Hamming + Jaro + Jaro-Winkler. Each algorithm: byte + rune variants, primary-source citation in block comment, formula in godoc, unit tests with literature reference vectors, property tests (identity, symmetry, range, triangle inequality where applicable), fuzz test with malformed UTF-8 corpus, benchmark with allocation budget enforced. First `bench.txt` committed. First cross-platform determinism golden file (`algorithms.json`) populated.
**Implements (architecture):** Foundation primitives + first algorithm-implementation tier. Two-row DP + stack-allocated buffer pattern lands here.
**Avoids (pitfalls):** 1 (distinct AlgoIDs for OSA and Full from day one), 2 (Jaro-Winkler constants verified against Winkler 1990, not Wikipedia), 13 (two-row DP from first commit).
**Research flag:** **NEEDS PHASE RESEARCH** — primary sources per algorithm: Levenshtein 1965, Damerau 1964 + Boytsov 2011 (OSA vs Full), Lowrance-Wagner 1975 (Full DL position-table), Hamming 1950, Jaro 1989, Winkler 1990. Reference vectors: `kitten`/`sitting`, `"ca"`/`"abc"` (the OSA-vs-Full divergence), `MARTHA`/`MARHTA` (Jaro = 0.9444, Jaro-Winkler = 0.9611), `DIXON`/`DICKSONX` (Jaro = 0.7667, JW boost gate fires → 0.8133). The `algorithm-correctness-reviewer` gates every PR.

### Phase 3: Core algorithms — remaining four (`v0.2.0`)

**Rationale:** Strcmp95 + Smith-Waterman-Gotoh + LCSStr + Ratcliff-Obershelp completes the character-based / gestalt catalogue. SWG is the highest-correctness-risk algorithm in the catalogue (Gotoh 1982 erratum); pairing it with three lighter algorithms in one tag is the spec's grouping but the roadmapper may wish to consider splitting.
**Delivers:** Four algorithms with the same test/bench/BDD coverage standard as Phase 2.
**Implements (architecture):** Character-based completion tier.
**Avoids (pitfalls):** 3 (Gotoh 1982 erratum — SWG implementation MUST cross-validate against EMBOSS or biopython; explicit in PR description), 14 (Strcmp95 similarity-character table as `var`, not `init()`).
**Research flag:** **NEEDS PHASE RESEARCH (HIGH PRIORITY FOR SWG)** — Winkler 1994 (Strcmp95 similar-character table), Gotoh 1982 + EMBOSS/biopython for SWG corrected formulation, Wagner-Fischer 1974 (LCSStr), Ratcliff-Obershelp 1988 (Dr. Dobb's Journal). SWG specifically: cross-validation reference vectors with one long gap testing the initialisation step.

### Phase 4: Q-gram and token-based (`v0.3.0`)

**Rationale:** Q-Gram extraction is shared infrastructure (`q_gram.go`) consumed by Jaccard / Sørensen-Dice / Cosine / Tversky. Token-based algorithms (Monge-Elkan, Token Sort / Set / Partial, Token Jaccard) require `tokenise.go` (already in foundation) and may require partial Scorer dispatch table.
**Delivers:** Q-Gram Jaccard, Sørensen-Dice, Cosine, Tversky, Monge-Elkan, Token Sort Ratio, Token Set Ratio, Partial Ratio, Token Jaccard. Shared `q_gram.go` extraction. Cosine implementation must explicitly parenthesise `(x*y) + z` and use `math.Sqrt` (NOT `math.Pow`) for norms. Token-ratio algorithms cite Indel formula explicitly.
**Implements (architecture):** Q-gram and token-based tier. Map-iteration discipline is heavily exercised here (q-grams use `map[string]int` internally — output paths must sort).
**Avoids (pitfalls):** 6 (Indel-vs-difflib documented explicitly; cross-validation against RapidFuzz, not fuzzywuzzy), 8 (map iteration in q-gram output paths), 9 (Cosine float-determinism on arm64 vs amd64), 12 (Monge-Elkan / Partial Ratio / Token Set Ratio DoS — document worst-case complexity in godoc).
**Research flag:** **NEEDS PHASE RESEARCH** — Ukkonen 1992 (Q-gram), Sørensen 1948 + Dice 1945, Salton 1975 (Cosine in IR), Tversky 1977 (asymmetric features), Monge-Elkan 1996, RapidFuzz documentation for Token Ratios Indel formula. **Architecturally important question for this phase:** Monge-Elkan composes an inner algorithm by AlgoID — does Phase 4's `MongeElkanScore` accept an `inner AlgoID` parameter (recommended), or is Monge-Elkan only callable via the Scorer (deferring to Phase 6)? If parameter-on-public-function: the dispatch table must exist before Monge-Elkan does, partial Scorer infrastructure must be in place by Phase 4.

### Phase 5: Phonetic algorithms (`v0.4.0`)

**Rationale:** Phonetic algorithms have degenerate score normalisation (0.0 or 1.0) so they exercise less of the float-determinism machinery. Double Metaphone has the largest rule table (~200 conditional branches across Germanic, Slavic, Romance, Greek branches) and the highest licence-discipline risk; landing it last gives `algorithm-licensing-reviewer` maximum context. Monge-Elkan's inner-algorithm permitted set expands to phonetic algorithms in this phase.
**Delivers:** Soundex (Knuth/Census variant, `"Tymczak" = "T522"` reference vector), Double Metaphone (Philips 2000 + public-domain C reference, four-language reference vectors), NYSIIS (Taft 1970, 6-char truncation), MRA (NBS Tech Note 943). Monge-Elkan permitted-inner-list expands to include phonetic algorithms.
**Implements (architecture):** Phonetic tier; Strcmp95-style `var` rule tables (no `init()`).
**Avoids (pitfalls):** 4 (Soundex variant explicit), 5 (NYSIIS truncation, Double Metaphone multilingual branch coverage), 14 (tables as `var` not `init()`), 16 (no GPL/LGPL references), 17 (patent screen — already cleared for these four; Metaphone 3 excluded).
**Research flag:** **NEEDS PHASE RESEARCH (HIGH PRIORITY)** — Russell-Odell 1918+1922 (canonicalised in Knuth TAOCP Vol. 3 §6.4), Philips 2000 (Double Metaphone, plus the public-domain C reference for rule cross-validation), Taft 1970 (NYSIIS — NY State Special Report No. 1, hard to obtain; cite Knuth or secondary review), NBS Tech Note 943 (MRA). Patent-screen results recorded as comments on each algorithm issue. Reference vectors per language-origin branch for Double Metaphone (`"Schmidt"` → `("XMT", "SMT")`, `"Pacheco"` → contains `PXK`, `"Catherine"`/`"Katherine"` should match).

### Phase 6: Scorer (`v0.5.0`)

**Rationale:** By this phase, all 23 algorithms exist as public functions. The Scorer is the composition layer.
**Delivers:** `scorer.go` + `scorer_options.go` (functional-options pattern, immutable after `NewScorer`, concurrent-safe, dispatch via AlgoID table). `NewScorer` + `DefaultScorer` + every `WithX` option. Methods: `Score` / `ScoreAll` / `Match` / `Threshold` / `Algorithms`. Auto-normalised weights (sum-to-1 invariant). BDD scenarios in `tests/bdd/features/scorer.feature` (first heavy use of godog + goleak + testify in the BDD sub-module). Second cross-platform determinism golden file (`scorer-default.json`). `DefaultScorerOptions()` returning the option slice (decision item from FEATURES.md gap analysis).
**Implements (architecture):** Layer 2 in full.
**Avoids (pitfalls):** 9 (Scorer composite float-summing — explicit parenthesisation + left-to-right reduction).
**Research flag:** **NEEDS PHASE RESEARCH (MEDIUM)** — primarily API-ergonomic research and consultation with `api-ergonomics-reviewer`: exact option names, exact method names, `MongeElkanScore` signature finalisation, parameterised-algorithm option shapes. Default Scorer weight rationale needs documentation.

### Phase 7: Scan sub-package (`v0.6.0`)

**Rationale:** Scan depends on Scorer. Token-bucket optimisation is correctness-critical and property-test-verified equivalent to naive O(N²).
**Delivers:** `scan/scan.go` (`Item`, `Config`, `Warning`, `WarningKind`, `Check`), `scan/bucket.go` (token-bucket), `scan/errors.go` (sentinel errors). Within-group + cross-group passes with separate thresholds. Suppression composition (per-item `SilenceLint` + global `SuppressedPairs` + cross-group identical-name). Deterministic sort by `(Kind, NameA, NameB, GroupA, GroupB)` with in-line assertion that no duplicates remain. Property test `PropCheck_BucketEquivalentToNaive`. Third cross-platform determinism golden file (`scan-default.json`). BDD scenarios in `scan.feature` + `suppression.feature`. Performance budgets per spec §12.6 (< 2s for 10,000 items) committed to `bench.txt`.
**Implements (architecture):** Layer 3 in full.
**Avoids (pitfalls):** 11 (sort-key completeness with in-line assertion), 12 (DoS — scan candidate-set short-circuit via token bucket).
**Research flag:** SKIP — architecture is fully spec'd; no external research needed beyond the property-test invariant proofs.

### Phase 8: Integration shakedown (`v0.6.x` patches)

**Rationale:** Re-scope `axonops/audit` issue #853 to consume `github.com/axonops/fuzzymatch` and `github.com/axonops/fuzzymatch/scan` end-to-end. API ergonomic issues surface in real-world use. Patch releases as needed.
**Delivers:** API refinements from integration feedback; documentation updates; any necessary patch fixes. Confirms zero-runtime-deps in practice via downstream consumer.
**Implements (architecture):** Integration validation.
**Research flag:** SKIP.

### Phase 9: v1.0.0 stable

**Rationale:** API frozen. Algorithm score stability guaranteed across patch versions within v1. Cross-platform determinism golden files locked.
**Delivers:** First signed release via cosign keyless; SBOM published; OIDC attestation attached; `axonops/audit` updated to depend on `fuzzymatch v1.0.0`; announcement post.
**Implements (architecture):** API freeze.
**Research flag:** SKIP.

### Phase Ordering Rationale

- **Foundation before anything else** — 9 of 20 pitfalls require infrastructure that EVERY algorithm phase depends on. Cross-platform CI, golden files, property test harness, benchstat regression detection, coverage enforcement, release workflow guardrails, and the cosign+SBOM+OIDC plumbing all must exist before the first algorithm lands. Without them, every algorithm phase carries silent risk.
- **Levenshtein first because it's the canonical DP example** — proves the entire testing pipeline (unit + property + fuzz + bench + BDD + golden file + cross-platform matrix) end-to-end on the simplest non-trivial algorithm. The ASCII fast-path + two-row DP + stack-allocated buffer pattern lands here and is replicated thereafter.
- **Damerau-Levenshtein OSA + Full together** — shipping one without the other invites the "Damerau-Levenshtein" ambiguity (Pitfall 1); distinct AlgoIDs from day one.
- **SWG paired with Strcmp95 in Phase 3 (`v0.2.0`)** — the spec groups four algorithms in `v0.2.0`; the roadmapper may consider isolating SWG given its erratum-driven cross-validation overhead. Either grouping works; the cross-validation discipline is the load-bearing constraint, not the tag.
- **Q-gram before token-based, but in the same tag (`v0.3.0`)** — token-based algorithms depend on `Tokenise` (foundation), but Monge-Elkan composes an inner algorithm by AlgoID, requiring partial Scorer dispatch infrastructure. The Monge-Elkan-inner-AlgoID question is architecturally significant and must be answered before Phase 4 lands.
- **Phonetic last among algorithm tiers (`v0.4.0`)** — degenerate float normalisation, largest rule tables (Double Metaphone), highest licence-discipline risk; gives `algorithm-licensing-reviewer` maximum context.
- **Scorer after all algorithms** — the spec's choice. The default. Bringing Scorer forward to Phase 4 with a partial catalogue is an arguable improvement that enables earlier `axonops/audit` integration testing — **flag for roadmap discussion**.
- **Scan after Scorer** — non-negotiable dependency.
- **Integration shakedown before v1.0.0** — surface API ergonomic issues in real consumer code before the API freezes.

### Research Flags

**Phases likely needing deeper research during planning** (gsd-research-phase candidates):
- **Phase 2 (Core algorithms — first six):** primary-source citations and reference vectors for Levenshtein, Damerau-Levenshtein OSA + Full, Hamming, Jaro, Jaro-Winkler. Standard but volume-heavy. The reviewing agents (`algorithm-correctness-reviewer`, `algorithm-licensing-reviewer`) are pre-existing gates.
- **Phase 3 (Core algorithms — remaining four), SWG specifically:** Gotoh 1982 erratum + EMBOSS/biopython cross-validation procedure. HIGH PRIORITY — standard "primary source + property tests" is insufficient.
- **Phase 4 (Q-gram and token-based):** RapidFuzz Token Sort/Set/Partial Ratio Indel formula confirmation; Cosine float-determinism implementation pattern (explicit parenthesisation, `math.Sqrt`, no `math.Pow`); Monge-Elkan inner-algorithm dispatch architecture decision.
- **Phase 5 (Phonetic):** HIGH PRIORITY — Soundex variant + reference vector, Double Metaphone multilingual rule branches + reference implementation cross-validation discipline (no GPL/LGPL references), NYSIIS canonical formulation, MRA NBS Tech Note 943 acquisition.
- **Phase 6 (Scorer):** API-ergonomic research and consultation with `api-ergonomics-reviewer` — option names, method names, default-Scorer weight rationale.

**Phases with standard patterns (skip research-phase):**
- **Phase 1 (Foundation/Bootstrap):** well-documented; `axonops/mask` is the structural template; only minor research on (a) self-hosted bench runner availability, (b) CLA Assistant configuration, (c) `wagoid/commitlint-github-action` setup. Most of this is "copy-and-adjust from mask".
- **Phase 7 (Scan):** architecture fully spec'd in `docs/requirements.md` §12; property-test invariants are the only verification work; no external research needed.
- **Phase 8 (Integration shakedown):** observation-driven, not research-driven.
- **Phase 9 (v1.0.0):** release-mechanics only; the release workflow has already been exercised through v0.x.

---

## Confidence Assessment

| Area | Confidence | Notes |
|------|------------|-------|
| Stack | HIGH | Every version verified against authoritative release pages within the last 90 days (Go 1.26.3 released 2026-05-07, golangci-lint v2.12.2 released 2026-05-06, GoReleaser v2.15.4 released 2026-04-21, Cosign v3.0.1 keyless default verified, gosec v2.25.0 released 2026-03-19, etc.). Architectural constraints (zero deps, no cgo, no testify in root, separate BDD module) are spec-locked. Three minor open decisions clearly identified (CLA/DCO, commitlint, toolchain pin) with recommendations. |
| Features | HIGH | Spec is comprehensive (`docs/requirements.md`, 1812 lines, 23 algorithms specified per-algorithm); Go ecosystem surveyed (`docs/prior-art-research.md`); RapidFuzz cross-checked as cross-language reference for API ergonomics. Five spec gaps clearly flagged with severity and recommendations. |
| Architecture | HIGH for spec-locked decisions (three layers, AlgoID enum, immutable Scorer, sub-package vs sub-module choice, determinism plumbing); MEDIUM for API-ergonomic specifics (final names/signatures gated by `api-ergonomics-reviewer`, eight open decisions documented). |
| Pitfalls | HIGH for algorithmic/determinism pitfalls (verified against primary sources, Go issue tracker, RapidFuzz docs, biorxiv survey for Gotoh erratum); MEDIUM for some phonetic-variant pitfalls (multiple canonical formulations exist — Soundex variant explicitly chosen in spec, but NYSIIS canonical formulation needs phase-research validation). |

**Overall confidence:** HIGH

### Gaps to Address

These don't undermine the roadmap but flag attention during planning or PROJECT.md updates:

- **Unicode NFC/NFD normalisation + diacritic stripping in `Normalise` pipeline** — MEDIUM-severity spec gap. `golang.org/x/text/unicode/norm` is non-stdlib (forbidden by zero-deps constraint); options are (a) implement a minimal NFC/NFD inline, (b) document as consumer responsibility with recipe, (c) accept the gap for v1.0 and revisit. **Surface for PROJECT.md decision before Phase 2 lands.**
- **Scorer phase placement** — bring forward from Phase 6 to Phase 4 (`v0.3.0`) for earlier `axonops/audit` integration testing? Default plan (Phase 6 / `v0.5.0`) is the spec's; bringing it forward is an arguable improvement. **Roadmap-level decision.**
- **Monge-Elkan signature** — does `MongeElkanScore(a, b string, inner AlgoID)` take the inner algorithm as a public parameter? Recommendation: yes, with zero-value default. Decision routes through `api-ergonomics-reviewer` but the architectural implication (partial Scorer dispatch infrastructure must exist by Phase 4) is roadmap-shaping.
- **`process.extract` one-to-many search API** — most-requested feature in comparable libraries; explicitly out of v1 scope. Track demand from v0.6.0 onward; v1.x or v2 candidate. **No action needed at roadmap time, but should appear in PROJECT.md "Future Consideration" notes.**
- **CLA vs DCO confirmation** — recommendation is CLA (mirrors mask), but should be explicitly confirmed.
- **Self-hosted bench runner availability** — Phase 1 needs to know whether `axonops/mask`/`axonops/audit` shared runner is available; falls back to `ubuntu-latest` informationally with regression detection skipped if unavailable.
- **`tests/bdd/go.mod` `replace` directive handling in CI** — CI must strip the local `replace github.com/axonops/fuzzymatch => ../..` before publishing or testing against the published module; verify in Phase 1.
- **NYSIIS primary-source acquisition** — Taft 1970 (NY State Special Report No. 1) is hard to obtain; Phase 5 research may need to cite Knuth or a secondary review article that documents the algorithm.

---

## Sources

### Primary (HIGH confidence)
- `docs/requirements.md` — authoritative spec, 1812 lines, 21 sections
- `docs/prior-art-research.md` — Go ecosystem survey + algorithm taxonomy
- `.planning/PROJECT.md` — project context, constraints, key decisions
- `CLAUDE.md` — agent gates, release discipline, branching, OOS reasoning
- `.claude/skills/research-guidance/SKILL.md` — what to research, what is settled
- `.claude/skills/algorithm-correctness-standards/SKILL.md` — primary-source citation, mathematical invariants, edge cases, Unicode handling
- `.claude/skills/algorithm-licensing-standards/SKILL.md` — patent screen, GPL/LGPL avoidance, source-origin statement, Metaphone 3 precedent
- `.claude/skills/determinism-standards/SKILL.md` — no-map-iteration, float stability, NaN/Inf/-0, golden files, cross-platform matrix
- `.claude/skills/performance-standards/SKILL.md` — two-row DP, ASCII fast path, stack allocation, benchstat regression detection
- `.claude/skills/go-coding-standards/SKILL.md` — AlgoID enum, no testify in root, no `init()`, sentinel errors
- `.planning/research/STACK.md` — full stack research with version verification
- `.planning/research/FEATURES.md` — table stakes vs differentiators vs anti-features + 5 spec gaps
- `.planning/research/ARCHITECTURE.md` — three-layer architecture + 6 patterns + 5 anti-patterns + open decisions
- `.planning/research/PITFALLS.md` — 20 pitfalls + technical debt patterns + integration gotchas + recovery strategies + pitfall-to-phase mapping

### Authoritative external (HIGH confidence — release/spec pages verified within 90 days)
- [Go 1.26 Release Notes](https://go.dev/doc/go1.26), [Go release history](https://go.dev/doc/devel/release)
- [golangci-lint v2 changelog](https://golangci-lint.run/docs/product/changelog/), [migration guide](https://golangci-lint.run/docs/product/migration-guide/)
- [GoReleaser v2](https://goreleaser.com/blog/goreleaser-v2/), [supply-chain example](https://github.com/goreleaser/example-supply-chain)
- [Sigstore Cosign v3](https://github.com/sigstore/cosign), [CHANGELOG](https://github.com/sigstore/cosign/blob/main/CHANGELOG.md)
- [goreleaser-action v7](https://github.com/goreleaser/goreleaser-action), [actions/checkout v6](https://github.com/actions/checkout/releases), [actions/setup-go v6](https://github.com/actions/setup-go/releases)
- [CodeQL Go 1.26 support](https://github.blog/changelog/2026-02-24-codeql-adds-go-1-26-and-kotlin-2-3-10-support-and-improves-query-accuracy/)
- [govulncheck](https://pkg.go.dev/golang.org/x/vuln/cmd/govulncheck), [gosec releases](https://github.com/securego/gosec/releases)
- [godog](https://pkg.go.dev/github.com/cucumber/godog), [goleak](https://github.com/uber-go/goleak), [testify](https://github.com/stretchr/testify)
- [RapidFuzz documentation](https://rapidfuzz.github.io/RapidFuzz/) — Indel-vs-difflib semantics
- [adrg/strutil](https://github.com/adrg/strutil), [hbollon/go-edlib](https://github.com/hbollon/go-edlib), [xrash/smetrics](https://pkg.go.dev/github.com/xrash/smetrics) — Go ecosystem comparables
- [axonops/mask](https://github.com/axonops/mask) — structural template

### Algorithm-correctness sources (HIGH confidence)
- Damerau-Levenshtein Wikipedia + RDocumentation `OSA` reference — OSA-vs-Full divergence on `"ca"`/`"abc"`
- Apache Commons Text `JaroWinklerDistance.java` — canonical `PREFIX_LENGTH_LIMIT = 4`
- biorxiv "Are all global alignment algorithms and implementations correct?" (Flouri et al.) — Gotoh 1982 erratum prevalence
- Go issue #17895 (FMA reproducibility), #71204 (GOAMD64=v3 FMA detection) — float cross-platform determinism
- Soundex Wikipedia / Rosetta Code — Census-vs-SQL-vs-Russell variants
- [llms.txt spec](https://llmstxt.org/) — emerging convention
- [Go modules reference](https://go.dev/ref/mod), [Managing dependencies](https://go.dev/doc/modules/managing-dependencies)

---

*Synthesis completed: 2026-05-13*
*Inputs: `.planning/research/{STACK,FEATURES,ARCHITECTURE,PITFALLS}.md`*
*Ready for roadmap: yes*
