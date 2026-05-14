# Roadmap: fuzzymatch

## Overview

`fuzzymatch` is a correctness-first pure-Go string-similarity library. The roadmap derives from a strict-downward architectural dependency — foundation primitives → algorithms (character → q-gram → token → phonetic) → composite Scorer → collection Scan → one-to-many Extract → integration shakedown → v1.0.0 stable. Phase 1 lands the entire safety scaffolding (cross-platform CI matrix, golden-file infrastructure, property-test harness, benchstat regression detection, coverage enforcement, cosign+SBOM+OIDC release pipeline, AlgoID dispatch, Normalise, Tokenise, sentinel errors) before a single algorithm is implemented, because 9 of the 20 inventoried correctness pitfalls are infrastructure-gated and every subsequent algorithm phase inherits this scaffolding. Algorithm phases follow the spec's tiering with two deliberate refinements: Smith-Waterman-Gotoh is isolated into its own phase due to the documented Gotoh 1982 erratum and the EMBOSS/biopython cross-validation overhead it requires, and phonetic algorithms are isolated into their own phase due to their unique licence-discipline and primary-source-sourcing characteristics. The Scorer lands after all 23 algorithms exist as standalone public functions (the spec's default choice — bringing it forward would force the default-Scorer composition to churn with every algorithm addition), Scan layers atop the Scorer, and Extract — added to v1.0 scope on 2026-05-13 — layers atop both single algorithms and the Scorer. The final phase consumes fuzzymatch from `axonops/audit` for real-world API ergonomic shakedown before the v1.0.0 freeze.

## Phases

**Phase Numbering:**
- Integer phases (1-11): Planned milestone work
- Decimal phases (e.g. 2.1): Urgent insertions added later

- [ ] **Phase 1: Foundation & Infrastructure** - Module, CI matrix, determinism plumbing, release pipeline, AlgoID dispatch, Normalise, Tokenise, errors
- [ ] **Phase 2: Core Character Algorithms (six)** - Levenshtein, Damerau-Levenshtein OSA + Full, Hamming, Jaro, Jaro-Winkler
- [ ] **Phase 3: Smith-Waterman-Gotoh** - Isolated for Gotoh 1982 erratum cross-validation against EMBOSS/biopython
- [ ] **Phase 4: Remaining Character & Gestalt** - Strcmp95, LCSStr, Ratcliff-Obershelp
- [ ] **Phase 5: Q-gram Algorithms** - Shared q-gram infrastructure + Jaccard, Sørensen-Dice, Cosine, Tversky
- [ ] **Phase 6: Token-based Algorithms** - Monge-Elkan, Token Sort Ratio, Token Set Ratio, Partial Ratio, Token Jaccard
- [ ] **Phase 7: Phonetic Algorithms** - Soundex, Double Metaphone, NYSIIS, MRA
- [ ] **Phase 8: Composite Scorer** - Functional-options weighted Scorer, ScoreAll, Match, normalisation control
- [ ] **Phase 9: Collection Scan Sub-package** - Within-group + cross-group passes, token-bucket optimisation, suppression composition
- [ ] **Phase 10: Extract API** - One-to-many `Extract` / `ExtractOne` search, RapidFuzz `process.extract`-equivalent
- [ ] **Phase 11: Integration Shakedown & v1.0.0** - `axonops/audit` consumption, API freeze, v1.0.0 signed release

## Phase Details

### Phase 1: Foundation & Infrastructure
**Goal**: Land the entire safety scaffolding before a single algorithm is implemented — module bootstrap, cross-platform CI determinism matrix, golden-file infrastructure, property-test harness conventions, benchstat regression detection, coverage enforcement, the full release pipeline (GoReleaser v2 + cosign keyless + SBOM + OIDC attestations), and the foundation primitives (AlgoID dispatch table, Normalise with Unicode NFC/NFD + diacritic stripping, Tokenise, sentinel errors) that every subsequent phase composes against.
**Depends on**: Nothing (first phase)
**Requirements**: FOUND-01, FOUND-02, FOUND-03, FOUND-04, FOUND-05, DET-01, DET-03, DET-05, DET-06, PERF-04, PERF-06, TEST-03, TEST-06, TEST-07, TEST-08, CI-01, CI-02, CI-03, CI-04, CI-05, CI-06, CI-07, CI-08, CI-09, CI-10, CI-11, REL-01, REL-02, REL-03, REL-04, REL-05, REL-06, REL-07, DX-01, DX-03, DX-04, DX-06, DX-07
**Success Criteria** (what must be TRUE):
  1. `go build ./...` and `go test ./...` succeed on a fresh clone with zero non-stdlib `require` lines in root `go.mod` (only `golang.org/x/text`), verified by `make verify-deps-allowlist` in CI
  2. CI determinism matrix (linux/amd64, linux/arm64, darwin/amd64, darwin/arm64, windows/amd64) runs a placeholder golden-file test that diffs byte-identically across all five platforms
  3. `make check` runs the full quality gate green (golangci-lint v2, vet, race, vulncheck, gosec, markdownlint, tidy-check, coverage-check enforcing ≥95%/≥90%/100%)
  4. `Normalise(s)` correctly handles ASCII fast path and Unicode NFC/NFD + diacritic stripping (`café` ↔ `cafe`, `Müller` ↔ `Mueller`); `Tokenise(s)` correctly splits camelCase / snake_case / PascalCase / kebab-case with stable ordering; both have property-test-verified determinism and are pinned in `normalisation.json` golden file
  5. A simulated tag-push to a fork produces a signed `checksums.txt`, Syft SBOM, OIDC artifact attestation, and GitHub Release via CI — and no local `git tag` path can produce a release
**Plans**: TBD

### Phase 2: Core Character Algorithms (six)
**Goal**: Ship the canonical six character-based similarity algorithms — Levenshtein, Damerau-Levenshtein OSA, Damerau-Levenshtein Full, Hamming, Jaro, Jaro-Winkler — each fresh-implemented from its primary academic source with literature reference vectors, mathematical-invariant property tests, fuzz tests, allocation-budgeted benchmarks, BDD scenarios, and entries in the cross-platform `algorithms.json` golden file. This phase proves the entire correctness-and-determinism pipeline end-to-end on the simplest non-trivial algorithm (Levenshtein) and locks in the two-row DP + ASCII fast-path + stack-allocated-buffer pattern that every subsequent DP-based algorithm reuses.
**Depends on**: Phase 1
**Requirements**: CHAR-01, CHAR-02, CHAR-03, CHAR-04, CHAR-05, CHAR-06, PERF-01, PERF-02, PERF-03, TEST-01, TEST-02, TEST-04, TEST-05, DET-02, DET-04, DX-02, DX-05
**Success Criteria** (what must be TRUE):
  1. A user can write a 5-line Go program that imports `github.com/axonops/fuzzymatch`, calls `Levenshtein("kitten", "sitting")` (or the agent-determined name), and gets a deterministic similarity score in `[0.0, 1.0]`
  2. Damerau-Levenshtein OSA and Damerau-Levenshtein Full ship as **distinct** AlgoIDs with the discriminating reference vector `"ca"` / `"abc"` proving the divergence (OSA distance = 3, Full distance = 2)
  3. Jaro-Winkler reference vectors `MARTHA`/`MARHTA` → 0.9611 and `DIXON`/`DICKSONX` → 0.8133 pass; Jaro constants (boost threshold `0.7`, prefix cap `4`, scale `0.1`) verified against Winkler 1990, not Wikipedia
  4. Allocation budgets enforced per algorithm via benchmark assertions; ASCII fast paths verified by escape-analysis check; two-row DP confirmed by code review (no full DP table allocation)
  5. Cross-platform golden file `algorithms.json` contains pinned scores for all six algorithms and diffs byte-identically across the CI matrix; first `bench.txt` committed with benchstat baseline; example program in `examples/identifier-similarity/` runs and is meta-tested
**Plans**: 7 plans
  - [x] 02-01-levenshtein-PLAN.md — Implement Levenshtein and lock the canonical Phase 2 pattern (algorithm + dispatch + tests + golden + BDD + example)
  - [x] 02-02-hamming-PLAN.md — Implement Hamming with the LOCKED silent-zero unequal-length policy
  - [x] 02-03-jaro-PLAN.md — Implement Jaro (match-flag arrays; not a metric) with Winkler-1990-traceable reference vectors
  - [x] 02-04-jaro-winkler-PLAN.md — Implement Jaro-Winkler as Jaro + prefix boost; pin three Winkler-1990 constants
  - [x] 02-05-damerau-levenshtein-osa-PLAN.md — Implement Damerau-Levenshtein OSA (three-row DP) with discriminating vector ca/abc → 3
  - [x] 02-06-damerau-levenshtein-full-PLAN.md — Implement Damerau-Levenshtein Full (Lowrance-Wagner) with discriminating vector ca/abc → 2
  - [x] 02-07-finalisation-PLAN.md — Merge per-algorithm staging files into algorithms.json; identifier-similarity example + meta-test; cross-algorithm consistency tests; first bench.txt baseline

### Phase 3: Smith-Waterman-Gotoh
**Goal**: Implement Smith-Waterman-Gotoh local alignment with configurable affine gap penalty, **isolated into its own phase** because the published Gotoh 1982 affine-gap recurrence contains a known erratum (initialisation step and an indexing flip — a biorxiv survey found 8 of 31 lecture-slide reproductions inherit the bug) and primary-source citation alone is insufficient. The implementation must cross-validate against EMBOSS or biopython reference vectors, the erratum must be called out explicitly in the file's block comment, and `algorithm-correctness-reviewer` review is gated on the cross-validation evidence appearing in the PR description.
**Depends on**: Phase 2
**Requirements**: CHAR-08
**Success Criteria** (what must be TRUE):
  1. `SmithWatermanGotoh` (agent-determined name) produces scores byte-identical to EMBOSS `water` and biopython `pairwise2.align.localxs` on a documented reference-vector set including one long-gap case that specifically exercises the Gotoh 1982 initialisation step
  2. The implementation file's block comment cites Gotoh 1982 AND explicitly names the corrected formulation source (Flouri et al. biorxiv survey or equivalent), with the erratum and its correction documented inline
  3. Configurable affine gap penalty (gap-open + gap-extend) exposed via per-algorithm option; default values documented with rationale; property tests verify identity, range, and non-negativity invariants
  4. Allocation budget enforced via benchmark; two-row DP variant; cross-platform golden file entry added; BDD scenario covers the canonical long-gap reference case
**Plans**: 3 plans
  - [ ] 03-01-swg-implementation-PLAN.md — Implement swg.go + dispatch + tests + property + fuzz + bench + BDD + staging golden (the core SWG surface)
  - [ ] 03-02-swg-cross-validation-PLAN.md — biopython corpus generator + committed vectors.json + TestSWG_CrossValidation + Makefile regen target (Gotoh-erratum gate)
  - [ ] 03-03-swg-finalisation-PLAN.md — Merge swg.json into algorithms.json + SWG-vs-Levenshtein divergence test + identifier-similarity SWG column + bench.txt + llms.txt + docs/requirements.md §7.1.8 update (Raw* surface)


### Phase 4: Remaining Character & Gestalt
**Goal**: Complete the character-based and gestalt catalogues with Strcmp95 (Winkler 1994 similar-character table), LCSStr (longest common substring similarity), and Ratcliff-Obershelp (Dr. Dobb's Journal 1988 — explicitly the difflib-equivalent for consumers who want that semantic, distinguishing it from the Indel-based token ratios coming in Phase 6).
**Depends on**: Phase 3
**Requirements**: CHAR-07, CHAR-09, GESTALT-01
**Success Criteria** (what must be TRUE):
  1. `Strcmp95` produces scores matching the Winkler 1994 reference implementation on canonical pairs; similar-character table declared as package-level `var` (NOT in `init()`) and verified by a static analysis check
  2. `LCSStr` returns the correct longest common substring length normalised to `[0.0, 1.0]` on Wagner-Fischer 1974 reference vectors; uses two-row DP
  3. `RatcliffObershelp` matches Python `difflib.SequenceMatcher.ratio()` outputs on the canonical Dr. Dobb's Journal 1988 reference pairs (proving fuzzymatch ships a true difflib-equivalent); its godoc explicitly contrasts it with the Indel-based token ratios coming in Phase 6
  4. All three algorithms have unit + property + fuzz + benchmark + BDD coverage; `algorithms.json` golden file extended with their entries and diffs byte-identically on the CI matrix

### Phase 5: Q-gram Algorithms
**Goal**: Ship the shared q-gram extraction infrastructure (`q_gram.go`) and the four q-gram-based algorithms that consume it — Q-Gram Jaccard, Sørensen-Dice, Cosine, Tversky. Cosine is the highest float-determinism risk in the catalogue (compiler-detected `x*y+z` patterns emit FMA on arm64 but typically not on amd64), so its implementation must use explicit `(x*y) + z` parenthesisation, `math.Sqrt` only (NO `math.Pow` for square roots), and left-to-right reduction; cross-platform byte-identical output is the load-bearing acceptance test.
**Depends on**: Phase 4
**Requirements**: QGRAM-01, QGRAM-02, QGRAM-03, QGRAM-04, QGRAM-05
**Success Criteria** (what must be TRUE):
  1. Shared q-gram extraction infrastructure (`q_gram.go`) is the single source of q-gram generation; the four downstream algorithms consume it via internal API; q-gram bag uses `map[string]int` internally but output paths never iterate the map (verified by lint rule)
  2. Cosine produces byte-identical scores across linux/amd64, linux/arm64, darwin/arm64, windows/amd64 on the cross-platform CI matrix; implementation uses explicit `(x*y) + z` parenthesisation and `math.Sqrt` only; `math.Pow` does not appear anywhere in the file
  3. Q-Gram Jaccard and Sørensen-Dice produce reference-vector matches against Ukkonen 1992 and Dice 1945 source examples; Tversky reference vectors with non-symmetric `alpha`/`beta` parameters verified against Tversky 1977
  4. All five files (`q_gram.go` plus the four algorithms) ship with unit + property + fuzz + benchmark + BDD; symmetry property test verifies Cosine/Jaccard/Sørensen-Dice symmetry, asymmetry property test verifies Tversky asymmetric behaviour

### Phase 6: Token-based Algorithms
**Goal**: Ship Monge-Elkan (with pluggable inner AlgoID dispatching against the already-established AlgoID table from Phase 1), Token Sort Ratio, Token Set Ratio, Partial Ratio, and Token Jaccard. The token ratios (Sort/Set/Partial) MUST cross-validate against RapidFuzz (Indel formula: `2·LCS/(|a|+|b|)`), NOT against fuzzywuzzy — fuzzywuzzy's pure-Python and C-extension paths produce different scores for the same library call, and the Indel formula is the unambiguous canonical choice. Documentation must explicitly state the Indel formula and contrast it with `RatcliffObershelp` (shipped in Phase 4) for consumers wanting the difflib semantic.
**Depends on**: Phase 5
**Requirements**: TOKEN-01, TOKEN-02, TOKEN-03, TOKEN-04, TOKEN-05
**Success Criteria** (what must be TRUE):
  1. `MongeElkan` (agent-determined signature) accepts an inner AlgoID parameter dispatching against the AlgoID table established in Phase 1; permitted inner algorithms cover the character + q-gram tiers; reference vectors match Monge-Elkan 1996
  2. Token Sort Ratio, Token Set Ratio, and Partial Ratio match RapidFuzz outputs byte-identically on a documented reference-vector set; each algorithm's godoc cites the Indel formula `2·LCS/(|a|+|b|)` explicitly and references RapidFuzz as the cross-validation source
  3. Token Jaccard produces correct set-Jaccard scores over `Tokenise(s)` output; uses `Tokenise` from Phase 1; map iteration discipline verified
  4. All five algorithms ship with unit + property + fuzz + benchmark + BDD; `algorithms.json` golden file extended; worst-case complexity for Monge-Elkan / Partial Ratio / Token Set Ratio (DoS vectors) documented explicitly in their godoc

### Phase 7: Phonetic Algorithms
**Goal**: Ship the four phonetic algorithms — Soundex (Knuth/Census variant per Knuth TAOCP Vol. 3 §6.4), Double Metaphone (Philips 2000, with primary + alternate codes), NYSIIS (Taft 1970, 6-char truncation), MRA (NBS Tech Note 943). Phonetic algorithms have degenerate score normalisation (typically 0.0/1.0 binary or small discrete sets), so they exercise less of the float-determinism machinery, but they carry the highest licence-discipline risk (Double Metaphone has the largest rule table with ~200 conditional branches across Germanic, Slavic, Romance, Greek branches — the public-domain C reference is the cross-validation source, NOT any GPL/LGPL Go port) and the hardest primary-source-sourcing (NYSIIS Taft 1970 is a NY State Special Report that's difficult to obtain — may require citing Knuth or a secondary review article). Once phonetic algorithms exist, Monge-Elkan's permitted-inner-algorithm set is expanded to include them.
**Depends on**: Phase 6
**Requirements**: PHON-01, PHON-02, PHON-03, PHON-04
**Success Criteria** (what must be TRUE):
  1. `Soundex` produces `"Tymczak" → "T522"` (the discriminating Knuth/Census variant reference vector — NOT the `"T520"` SQL variant) and `"Ashcraft" → "A261"` matching Knuth TAOCP Vol. 3 §6.4; the variant name and source appear in the file's block comment
  2. `DoubleMetaphone` returns primary + alternate codes; reference vectors per language-origin branch (Germanic: `"Schmidt" → ("XMT", "SMT")`; Slavic / Romance: `"Pacheco"` contains `PXK`; Greek: `"Catherine"` and `"Katherine"` match) pass; cross-validation source is the public-domain Philips 2000 C reference, no GPL/LGPL Go port referenced, verified by `algorithm-licensing-reviewer` sign-off recorded in the PR
  3. `NYSIIS` truncates at 6 characters per Taft 1970; reference vectors match either Knuth or a documented secondary-source citation (source choice recorded in block comment)
  4. `MRA` Match Rating Approach matches NBS Tech Note 943 reference cases; rule tables declared as package-level `var` (NOT in `init()`); Monge-Elkan's permitted-inner-list is updated to include the four phonetic AlgoIDs; all four algorithms have full unit + property + fuzz + benchmark + BDD coverage and `algorithms.json` entries

### Phase 8: Composite Scorer
**Goal**: Ship the composite weighted Scorer (Layer 2 of the three-layer architecture) — `NewScorer(opts ...ScorerOption)` functional-options constructor, immutable after construction, concurrent-safe, dispatch against the AlgoID table established in Phase 1 and used by Monge-Elkan since Phase 6. `DefaultScorer()` provides the opinionated default; `DefaultScorerOptions()` returns the underlying option slice for "default minus algorithm X" customisation. `Score`, `ScoreAll`, `Match`, `Threshold`, `Algorithms` methods. Weights auto-normalise to sum-to-1. Cross-platform `scorer-default.json` golden file pinned. BDD module sees its first heavy use (godog + goleak + testify in `tests/bdd/`).
**Depends on**: Phase 7
**Requirements**: SCORER-01, SCORER-02, SCORER-03, SCORER-04, SCORER-05, SCORER-06, SCORER-07, SCORER-08
**Success Criteria** (what must be TRUE):
  1. A user can write `s := fuzzymatch.NewScorer(fuzzymatch.WithAlgorithm(...))`, then call `s.Score(a, b)` from multiple goroutines concurrently without data races (`go test -race` clean) and get deterministic results
  2. `DefaultScorer()` produces a documented opinionated default; `DefaultScorerOptions()` returns the option slice so consumers can do `NewScorer(append(DefaultScorerOptions(), WithoutAlgorithm(...))...)` without rebuild-from-scratch; weight auto-normalisation invariant (sum-to-1) verified by property test
  3. `Score(a, b) float64` returns `[0.0, 1.0]`; `ScoreAll(a, b) map[AlgoID]float64` returns per-algorithm breakdown with deterministic key set; `Match(a, b) bool` honours `Threshold()`; `Algorithms()` returns the configured set; `WithoutNormalisation()` / `WithCustomNormalisation()` work as documented
  4. `scorer-default.json` golden file diffs byte-identically across the CI matrix; the Scorer composite explicit-parenthesisation + left-to-right reduction is verified by code review and documented; BDD scenarios in `tests/bdd/features/scorer.feature` exercise composition patterns; goleak confirms zero goroutine leaks

### Phase 9: Collection Scan Sub-package
**Goal**: Ship the `scan/` sub-package (Layer 3 of the three-layer architecture) — turnkey collection-scan layer over the Scorer with within-group + cross-group passes (separate thresholds), token-bucket optimisation property-test-verified equivalent to naive O(N²), suppression composition (per-item `SilenceLint` flag + global `SuppressedPairs` list + cross-group identical-name default), deterministic output sort by `(Kind, NameA, NameB, GroupA, GroupB)` with in-line completeness assertion that no duplicate sort keys remain. Performance budget < 2s for 10,000 items committed to `bench.txt`. Sentinel error hierarchy for scan-specific failures. Cross-platform `scan-default.json` golden file pinned.
**Depends on**: Phase 8
**Requirements**: SCAN-01, SCAN-02, SCAN-03, SCAN-04, SCAN-05, SCAN-06, PERF-05
**Success Criteria** (what must be TRUE):
  1. `scan.Check(items, cfg) []Warning` runs within-group + cross-group passes with separate thresholds; on a 10,000-item input, performance budget < 2s holds and is benchmark-asserted
  2. Property test `PropCheck_BucketEquivalentToNaive` proves the token-bucket optimisation produces a warning set identical (under sort) to the naive O(N²) implementation for randomly-generated input
  3. Suppression composition works: per-item `SilenceLint=true` suppresses; global `SuppressedPairs` list suppresses additively; cross-group identical-name default suppresses unless disabled; suppression interactions are BDD-tested in `tests/bdd/features/suppression.feature`
  4. Output is deterministically sorted by `(Kind, NameA, NameB, GroupA, GroupB)` with an in-line assertion that the sort key is complete (no equal sort-key pairs survive); `scan-default.json` golden file diffs byte-identically across the CI matrix; sentinel error hierarchy composes with `errors.Is`/`errors.As`

### Phase 10: Extract API
**Goal**: Ship the one-to-many `Extract` / `ExtractOne` search API (added to v1.0 scope on 2026-05-13 in response to RapidFuzz `process.extract` being the most-requested feature in comparable libraries). Pluggable scorer (any `*Scorer` or single `AlgoID`); cutoff threshold; result limit option; deterministic tie-breaking (score-descending, then choice-index-ascending) property-test-verified. Worst-case complexity and adversarial-input DoS warning documented prominently.
**Depends on**: Phase 9 (uses Scorer from Phase 8 and follows scan's API-shape conventions)
**Requirements**: EXTRACT-01, EXTRACT-02, EXTRACT-03, EXTRACT-04, EXTRACT-05
**Success Criteria** (what must be TRUE):
  1. A user can write `matches := fuzzymatch.Extract(query, choices, fuzzymatch.WithScorer(s), fuzzymatch.WithCutoff(0.7), fuzzymatch.WithLimit(5))` (agent-determined exact shape) and get the top-5 matches above the cutoff, sorted score-descending
  2. `ExtractOne(query, choices, opts) Match` returns the single best match; tie-breaking is deterministic (score-descending, then choice-index-ascending) and verified by property test
  3. Pluggable scorer works for both `*Scorer` and single `AlgoID`; cutoff threshold and result limit options compose as documented; `extract-default.json` golden file diffs byte-identically across the CI matrix
  4. Worst-case complexity documented in godoc; adversarial-input DoS warning appears prominently (since Extract is the most exposed one-to-many surface); fuzz tests pass; `examples/extract-demo/` mini-program runs and is meta-tested

### Phase 11: Integration Shakedown & v1.0.0
**Goal**: Final phase — re-scope `axonops/audit` (issue #853) to consume `github.com/axonops/fuzzymatch` and `github.com/axonops/fuzzymatch/scan` (and `Extract`) end-to-end. Surface any API ergonomic issues from real consumer code; apply patch fixes; finalise documentation; freeze the API. Ship v1.0.0 via signed CI release: cosign keyless `checksums.txt`, Syft SBOM, OIDC artifact attestation, GitHub Release. `axonops/audit` updated to depend on `fuzzymatch v1.0.0`. Announcement post. No new requirements introduced; this phase exists to validate that everything works in a real consumer and to formally lock the API surface.
**Depends on**: Phase 10
**Requirements**: (none unique — this phase validates and ships the work of Phases 1-10; all v1 requirements are mapped above)
**Success Criteria** (what must be TRUE):
  1. `axonops/audit` (issue #853) imports `github.com/axonops/fuzzymatch` and `github.com/axonops/fuzzymatch/scan`, replaces its previous fuzzy-matching code, and its full test suite passes with the new dependency
  2. Any API ergonomic issues surfaced during integration are either resolved before v1.0.0 (via `api-ergonomics-reviewer` review and a patch release) or explicitly deferred to v1.x / v2 with a tracking issue
  3. `v1.0.0` tag is pushed via the CI release workflow (no local `git tag`, no local `goreleaser release`); the GitHub Release contains a cosign-signed `checksums.txt`, a Syft SBOM, and an OIDC artifact attestation, all verifiable by a downstream consumer
  4. All 93 v1 requirements are marked complete in REQUIREMENTS.md traceability; all five cross-platform golden files (`algorithms.json`, `scorer-default.json`, `scan-default.json`, `extract-default.json`, `normalisation.json`) are byte-identical across the CI matrix and pinned for v1.x stability; release announcement post links the GitHub Release, the docs site, and the `axonops/audit` consumer

## Progress

**Execution Order:**
Phases execute in numeric order: 1 → 2 → 3 → 4 → 5 → 6 → 7 → 8 → 9 → 10 → 11

| Phase | Plans Complete | Status | Completed |
|-------|----------------|--------|-----------|
| 1. Foundation & Infrastructure | 0/TBD | Not started | - |
| 2. Core Character Algorithms (six) | 0/7 | Not started | - |
| 3. Smith-Waterman-Gotoh | 0/3 | Not started | - |
| 4. Remaining Character & Gestalt | 0/TBD | Not started | - |
| 5. Q-gram Algorithms | 0/TBD | Not started | - |
| 6. Token-based Algorithms | 0/TBD | Not started | - |
| 7. Phonetic Algorithms | 0/TBD | Not started | - |
| 8. Composite Scorer | 0/TBD | Not started | - |
| 9. Collection Scan Sub-package | 0/TBD | Not started | - |
| 10. Extract API | 0/TBD | Not started | - |
| 11. Integration Shakedown & v1.0.0 | 0/TBD | Not started | - |
