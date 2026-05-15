# Phase 7: Phonetic Algorithms - Discussion Log

> **Audit trail only.** Do not use as input to planning, research, or execution agents.
> Decisions are captured in CONTEXT.md — this log preserves the alternatives considered.

**Date:** 2026-05-15
**Phase:** 07-phonetic-algorithms
**Areas discussed:** Cross-validation source, NYSIIS primary-source path, Double Metaphone implementation strategy, permittedMongeElkanInner extension timing, Non-ASCII input handling, MRACompare API surface, Phonetic golden file schema, Identifier-similarity example columns

---

## Cross-validation source

### Approach

| Option | Description | Selected |
|--------|-------------|----------|
| Literature only (Phase 5 pattern) | Hand-derived from Knuth/Philips/Taft/NBS — 8–12 vectors per algorithm, all in unit tests. No Python pip dep added to dev toolchain. Lowest scaffolding cost; relies on literature vectors being sufficient. | |
| Pinned jellyfish + corpus (Phase 6 pattern) | Add `jellyfish==<pinned>` dev dep, write generator, commit `vectors.json`, Go loader test. Wider corpus than literature alone; defends against branch-coverage misses in Double Metaphone. | |
| Both — literature primary + jellyfish supplementary | Hand-derived literature vectors as canonical unit tests; jellyfish corpus as separate cross-validation. Highest scaffolding cost but strongest defence-in-depth. | ✓ |

**User's choice:** Both — literature primary + jellyfish supplementary.

### Variant divergence handling

| Option | Description | Selected |
|--------|-------------|----------|
| Per-entry variant tag | Generator inspects each input; variant-divergent entries carry `variant_divergence: true` + both Knuth-expected and jellyfish values. Loader test asserts our impl matches Knuth-expected. | ✓ |
| Soundex excluded from jellyfish corpus | Hand-derived Knuth vectors for Soundex only. Simpler generator; loses Soundex's wider corpus coverage. | |
| Switch to a Python lib that uses Knuth variant | Survey alternatives (abydos, phonetics, fuzzy). Adds research cost; outcome uncertain. | |

**User's choice:** Per-entry variant tag.

### Version pin policy

| Option | Description | Selected |
|--------|-------------|----------|
| Same as Phase 6 RapidFuzz pattern | Inline `JELLYFISH_VERSION = "<exact-pinned>"` at top of generator; assert at runtime; vectors.json metadata. Identical to Phase 6 mechanism. | ✓ |
| Loose pin (major.minor only) | Allow patch versions to vary. Reduces churn but loses byte-stable corpus guarantee. | |
| Pin via lockfile (requirements.txt) | Pin in separate file. Cleaner separation but breaks Phase 6 pattern. | |

**User's choice:** Same as Phase 6 RapidFuzz pattern.

### Corpus size

| Option | Description | Selected |
|--------|-------------|----------|
| 20 per algorithm — mirror Phase 6 | 80 total. Covers ASCII names + edge cases + (for DM) one input per language-origin branch. | |
| 30 per algorithm — stretch | 120 total. More edge-case coverage. | |
| Tiered by algorithm complexity | Soundex 15, NYSIIS 20, MRA 20, Double Metaphone 40. Targets the complexity hotspot. | ✓ |

**User's choice:** Tiered by algorithm complexity (Soundex 15 / NYSIIS 20 / MRA 20 / Double Metaphone 40).

**Notes:** Total 95 entries. Double Metaphone's 40 entries cover all 5 language-origin branches (Germanic / Slavic / Romance / Greek / Chinese), minimum 7 per major branch.

---

## NYSIIS primary-source path

### Citation strategy

| Option | Description | Selected |
|--------|-------------|----------|
| Knuth secondary citation | Cite Knuth TAOCP Vol. 3 §6.4 as canonical algorithm description (alongside Taft 1970 as origin). Aligns with Soundex sourcing pattern. | ✓ |
| Wikipedia + jellyfish vectors as reference | Block comment cites Taft 1970 + Wikipedia + jellyfish. Most transparent; weakest authority. | |
| Acquire Taft 1970 — spike effort | Pre-implementation spike: ILL, scan PDF, host in `docs/primary-sources/`. Highest fidelity; uncertain timeline. | |

**User's choice:** Knuth secondary citation.

### Variant choice

| Option | Description | Selected |
|--------|-------------|----------|
| Original NYSIIS-1970, 6-char truncation | Per spec §7.4.3: Taft 1970 algorithm with 6-character truncation, each rule applied once. | ✓ |
| Original + IterateToFixedPoint flag | Ship original but expose `NYSIISOptions{IterateToFixedPoint: bool}` for modified variant. Adds surface; deferred to v1.x. | |
| Modified-NYSIIS ("modern" variant) | Diverges from spec. Rejected. | |

**User's choice:** Original NYSIIS-1970, 6-char truncation, each rule applied once.

### Cross-validation divergence

| Option | Description | Selected |
|--------|-------------|----------|
| Tag divergent entries, our impl wins | Same pattern as Soundex. For NYSIIS entries where jellyfish diverges from Knuth, tag `variant_divergence: true`. | ✓ |
| Restrict NYSIIS corpus to non-divergent inputs | Smaller but cleaner corpus; risk of insufficient coverage. | |
| Skip NYSIIS in jellyfish corpus | Hand-derived literature vectors only for NYSIIS. | |

**User's choice:** Tag divergent entries, our impl wins.

---

## Double Metaphone implementation strategy

### Plan decomposition

| Option | Description | Selected |
|--------|-------------|----------|
| One plan, validated by language-branch | Single plan 07-02 lands full `double_metaphone.go`. Validation is a mandatory language-branch checklist. Mirrors Phase 4 RatcliffObershelp. | ✓ |
| Split by language-origin branch | 5 plans (Germanic / Slavic / Romance / Greek / Chinese). More wave overhead, more dependencies. | |
| Two plans — spine + edge cases | 07-02 Germanic+Slavic, 07-03 Romance+Greek+Chinese. Preserves parallelism. | |

**User's choice:** One plan, validated by language-branch.

### Audit trail format

| Option | Description | Selected |
|--------|-------------|----------|
| Source-Origin Statement + rule-table provenance line | Block comment includes (a) Source-Origin Statement, (b) rule-table provenance line, (c) MIT-Go-port negative-attribution line. algorithm-licensing-reviewer signs off. | ✓ |
| Source-Origin Statement only | Standard Phase 4 format. Lower friction; weaker audit trail. | |
| Per-rule-branch attribution comments | Each language-branch carries provenance comment. Highest transparency; impedes refactoring. | |

**User's choice:** Source-Origin Statement + rule-table provenance line + MIT-Go-port negative-attribution line.

### Review depth

| Option | Description | Selected |
|--------|-------------|----------|
| Deep review on plan 07-02; standard elsewhere | Plan 07-02 PR triggers `--depth=deep` code review. Other phase-7 plans standard. | ✓ |
| Standard depth everywhere | Same depth across all plans. Lower CI cost. | |
| Additional algorithm-licensing-reviewer pass on 07-02 | Two reviewer passes + dedicated licensing-reviewer with provenance manifest. Highest assurance. | |

**User's choice:** Deep review on plan 07-02; standard elsewhere.

---

## permittedMongeElkanInner extension timing

### When each entry is added

| Option | Description | Selected |
|--------|-------------|----------|
| Each algorithm's own plan adds its own entry | Plans 07-01..07-04 each add `Algo<X>: true` + update panic-test fixture in lockstep. Each PR self-contained. Standard pattern. | ✓ |
| Dedicated finalisation step adds all 4 at once | Allow-list stays at 14 entries until plan 07-05. Risk of intermediate test failure. | |
| All 4 added up-front in plan 07-01 | Add all 4 even though dispatch slots aren't wired. Risks dispatch[inner] returning nil. | |

**User's choice:** Each algorithm's own plan adds its own entry.

### Binary-inner ME composition tests

| Option | Description | Selected |
|--------|-------------|----------|
| Add one property test per phonetic inner | 4 `TestMongeElkanScore_BinaryInner_<Algo>` tests in `monge_elkan_test.go`. Locks binary-inner-composition against regression. | ✓ |
| Add one BDD scenario per phonetic inner | Same coverage as property tests; longer feedback loop. | |
| No additional tests | Existing property tests already cover composition. Adding binary-inner tests is double-coverage. | |

**User's choice:** Add one property test per phonetic inner.

---

## Non-ASCII input handling

| Option | Description | Selected |
|--------|-------------|----------|
| Skip silently, document limitation | Only `[A-Za-z]` participates in encoding. Godoc warning paragraph. No options struct. | ✓ |
| Refuse non-ASCII with sentinel error | Add `ErrNonASCIIInput`. Breaks spec line 649. | |
| Accept NormaliseOptions parameter | Signature gains `opts NormaliseOptions`. Adds surface. | |

**User's choice (with annotation):** Skip silently and document limitation — **but** capture `SoundexCodeNormalised` (and analogues) as an explicit deferred Phase 9+ expansion idea in CONTEXT.md so future-Johnny doesn't have to re-litigate this.

**Notes:** Captured in CONTEXT.md `<deferred>` section as the first item.

---

## MRACompare API surface

| Option | Description | Selected |
|--------|-------------|----------|
| Keep as spec — public (bool, int) surface | `MRACode`, `MRACompare(a,b) (bool, int)`, `MRAScore`. Int is raw 0-6 NBS counter. | ✓ |
| Demote MRACompare to unexported | Only `MRACode` and `MRAScore` public. Smaller surface; rejects spec line 691. | |
| Normalise int to float64 | `MRACompare(a,b) (bool, float64)` where float is `int/6.0`. Spec deviation. | |

**User's choice:** Keep as spec — public `(bool, int)` surface.

---

## Phonetic golden file schema

| Option | Description | Selected |
|--------|-------------|----------|
| Separate `phonetic-codes.json` + score entries in algorithms.json | Two golden surfaces with internally consistent schemas. `algorithms.json` unchanged. | ✓ |
| Extend algorithms.json schema with optional codes field | Single golden file with schema change rippling to all phases. | |
| Codes in *Code-test fixtures only | No golden file for codes; loses cross-platform code determinism gate. | |

**User's choice:** Separate `testdata/golden/phonetic-codes.json` + Score entries continue in `testdata/golden/algorithms.json`.

---

## Identifier-similarity example columns

| Option | Description | Selected |
|--------|-------------|----------|
| Score columns only | 4 new columns (`Soundex / DblMetaph / NYSIIS / MRA`) showing 0.0/1.0 scores. 23 columns total. | |
| Score + code columns side-by-side | 8 new columns showing both binary scores and encoded keys. Doubles table width. | |
| Score columns + separate phonetic-keys example program | 4 score columns in identifier-similarity (23 total); new `examples/phonetic-keys/main.go` for the `XxxCode`/`XxxKeys` surface standalone. | ✓ |

**User's choice:** Score columns in identifier-similarity + separate `examples/phonetic-keys/main.go` program.

---

## Claude's Discretion

Areas where the user deferred to Claude / the planner:

- Wave decomposition shape (planner picks: likely 07-01 Soundex, 07-02 DM, 07-03 NYSIIS, 07-04 MRA, 07-05 finalisation).
- Exact jellyfish version pin (planner picks current stable at plan time).
- Exact staging-golden entry counts per algorithm (8-12 per Phase 2-6 norm).
- Whether `phonetic_codes_golden_test.go` is a separate file or embedded into `algorithms_golden_test.go` (recommended: separate file).
- Fuzz seed counts and exact BDD scenario counts per algorithm.
- Exact wording of the godoc warning paragraph beyond the canonical text in CONTEXT.md §5.
- Granular structure of `examples/phonetic-keys/main.go` (column layout, header rows, name set).

## Deferred Ideas

(Detailed in CONTEXT.md `<deferred>` section.)

- `XxxCodeNormalised(s, opts NormaliseOptions) string` wrappers as a Phase 9+ additive expansion — explicitly annotated by the user.
- `NYSIISOptions{IterateToFixedPoint: bool}` for modified-NYSIIS variant — v1.x.
- `SoundexSQLCode` as a second public function for SQL/MySQL variant parity — v1.x.
- Acquiring Taft 1970 PDF via ILL — v1.x if researcher contributes or NY State Archives digitises.
- Per-rule-branch attribution comments in `double_metaphone.go` — rejected as too granular.
- Cross-validation against Python `abydos` library — rejected (second pip dep).
- `MRACompareNormalised(a, b) (bool, float64)` wrapper for catalogue-wide float64 uniformity — v1.x.
- Combining `phonetic-codes.json` and `algorithms.json` golden files — v1.x if a unified loader is preferable.
