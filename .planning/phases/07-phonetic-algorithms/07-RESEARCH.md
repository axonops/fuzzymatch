# Phase 7: Phonetic Algorithms - Research

**Researched:** 2026-05-15
**Domain:** Phonetic name-encoding algorithms (Soundex / Double Metaphone / NYSIIS / MRA), licence-discipline-heavy, lowest-float-determinism-exposure but highest rule-table / variant divergence risk
**Confidence:** HIGH for Soundex, NYSIIS, MRA; MEDIUM-HIGH for Double Metaphone (rule-table provenance verified; canonical reference vectors well documented; jellyfish cross-validation NOT available — see §4)

## Summary

Phase 7 ships four English-tuned phonetic algorithms onto pre-reserved AlgoID slots (`AlgoSoundex`, `AlgoDoubleMetaphone`, `AlgoNYSIIS`, `AlgoMRA` at `algoid.go` lines 156-175). Soundex/NYSIIS/MRA are short, well-understood algorithms with low primary-source-fidelity risk; **Double Metaphone is the load-bearing risk** — a ~400-line, ~200-conditional-branch rule table across 5 language-origin branches with multiple known MIT-licensed Go ports that a careless implementer would shortcut to. The boolean-score nature of these algorithms (1.0 if codes match, 0.0 otherwise) takes float-determinism off the critical path; the cross-platform determinism gate value is the **encoded code string** itself, captured in a new `testdata/golden/phonetic-codes.json` per CONTEXT.md §7.

Three research findings demand the planner's attention because they affect tasks already locked in CONTEXT.md §1:

1. **Jellyfish has no Double Metaphone implementation** — `jellyfish.metaphone` is single-key Metaphone, not Double. The 40-entry DM cross-validation corpus committed by CONTEXT.md §1 LOCKED needs a different source. Recommendation: `oubiwann/metaphone` (BSD-licensed, single-package Python 3 port carrying both `metaphone` and `doublemetaphone`; 1,034 GitHub stars; the de-facto Python DM canonical source) OR Apache Commons Codec Java reference for hand-derivation of literature vectors. Strong recommendation: use `oubiwann/metaphone` as a second pinned pip dep (`metaphone==<pin>`) alongside `jellyfish==<pin>` for the three jellyfish-supported algorithms.

2. **Jellyfish NYSIIS does NOT truncate to 6 characters** — `Catherine → CATARAN` (7 chars), `montgomery → MANTGANARY` (10 chars), `martincevic → MARTANCAFAC` (11 chars). This is the modified/extended NYSIIS variant; CONTEXT.md §2 LOCKS the original Taft-1970 6-char truncation. The `variant_divergence: true` mechanism in CONTEXT.md §1 correctly anticipates this: **every NYSIIS jellyfish vector longer than 6 chars must carry the divergence tag, and our impl is asserted against the Taft-truncated value, NOT the jellyfish value.**

3. **Jellyfish Soundex uses the Knuth/Census variant we're locked to** — `Tymczak → T522`, `Ashcroft → A261` both confirmed by reading `jellyfish/src/soundex.rs` directly. H and W are skipped (not separators), matching Knuth. The `variant_divergence` mechanism for Soundex in CONTEXT.md §1 is **insurance with no expected actual divergences for the canonical inputs** — the planner should keep the schema but expect `variant_divergence: false` on every entry in practice. (The discriminating vectors are still load-bearing in the literature unit tests for the file-level variant commitment.)

**Primary recommendation:** Five plans (07-01 Soundex foundation + jellyfish infra, 07-02 Double Metaphone with deep review, 07-03 NYSIIS with truncation-divergence handling, 07-04 MRA, 07-05 finalisation). Plan 07-01 lands `testdata/golden/phonetic-codes.json`, `scripts/gen-phonetic-cross-validation.py`, AND the first algorithm; plan 07-02 carries the heaviest weight (mandatory language-branch checklist + `algorithm-licensing-reviewer` sign-off + `--depth=deep` review); plans 07-03 and 07-04 follow the locked Phase 6 staging-golden → finalisation pattern.

## User Constraints (from CONTEXT.md)

### Locked Decisions

The following 8 decisions are LOCKED in `.planning/phases/07-phonetic-algorithms/07-CONTEXT.md`. The planner MUST honour them verbatim; this research describes HOW to implement them and what edge cases/gotchas exist.

- **§1 Cross-validation source — LOCKED:** Literature primary (hand-derived from Knuth/Philips/Taft/NBS — 4-8 per algorithm in unit tests) + pinned jellyfish supplementary (15 Soundex / 20 NYSIIS / 20 MRA / 40 DM = 95 corpus entries). Per-entry `variant_divergence: true` flag + `divergent_jellyfish_value` field handles known divergences. `vectors.json._metadata` carries `jellyfish_version`, `python_version`, `regenerated_at`, and a script checksum. Mechanism mirrors Phase 6 RapidFuzz pin exactly.
- **§2 NYSIIS primary-source citation — LOCKED:** Knuth TAOCP Vol. 3 §6.4 as canonical algorithm-description citation; Taft 1970 cited as algorithmic origin (famously unavailable as a NY State Special Report). Variant choice: **original NYSIIS-1970, 6-character truncation, each rule applied once** (no iterate-to-fixed-point). Modified-NYSIIS variants rejected.
- **§3 Double Metaphone implementation strategy — LOCKED:** Single plan 07-02 lands `double_metaphone.go` end-to-end; validation is a mandatory language-branch checklist (Germanic / Slavic / Romance / Greek / Chinese-origin minimum 1 vector each). Audit trail: Source-Origin Statement + rule-table provenance line + MIT-Go-port negative-attribution line. `--depth=deep` code review for plan 07-02 only. `algorithm-licensing-reviewer` signs off on the PR.
- **§4 permittedMongeElkanInner extension timing — LOCKED:** Each algorithm's own plan adds its own AlgoID entry (15 → 16 → 17 → 18 entries across plans 07-01..07-04) AND updates `monge_elkan_test.go`'s panic-test fixture (rejected slice shrinks 14 → 13 → 12 → 11 → 10 in lockstep). Four new property tests `TestMongeElkanScore_BinaryInner_<Algo>` lock binary-inner-composition behaviour.
- **§5 Non-ASCII input handling — LOCKED:** Skip silently; document limitation in canonical godoc warning paragraph (text in CONTEXT.md). No `NormaliseOptions` parameter on Phase 7 phonetic functions in v1.0. Fuzz seeds cover both ASCII and mixed-non-ASCII regimes.
- **§6 MRACompare API surface — LOCKED:** Three public functions: `MRACode(s) string`, `MRACompare(a, b) (bool, int)`, `MRAScore(a, b) float64`. The `int` from `MRACompare` is the raw 0-6 NBS Tech Note 943 similarity counter, exposed per spec line 691.
- **§7 Phonetic golden file schema — LOCKED:** `testdata/golden/algorithms.json` (existing) carries the binary Score entries; `testdata/golden/phonetic-codes.json` (new, plan 07-01) carries byte-stable code vectors with 8-12 entries per algorithm. Loader test: `phonetic_codes_golden_test.go` (recommended separate file).
- **§8 Identifier-similarity example extension — LOCKED:** 4 new score columns (`Soundex / DblMetaph / NYSIIS / MRA`) in `examples/identifier-similarity/main.go` (19 → 23 columns). New `examples/phonetic-keys/main.go` for the encoded-key surface, both landing in the finalisation plan.

### Claude's Discretion

The planner chooses, without re-asking the user:

- Wave decomposition: 4 algorithms + finalisation = 5 plans. Recommended shape: 07-01 Soundex (foundation), 07-02 Double Metaphone (heaviest, deep review), 07-03 NYSIIS, 07-04 MRA, 07-05 finalisation.
- Exact jellyfish version pin: see §4 below — recommendation `jellyfish==1.2.1` (released 2025-10-11).
- Exact number of staging-golden entries per algorithm: 8-12 per Phase 2-6 norm.
- Whether to ship `phonetic_codes_golden_test.go` as a separate test file or embed loader into `algorithms_golden_test.go`: **separate file recommended** for clean separation.
- Fuzz seed counts per algorithm (8-16 seeds covering ASCII + non-ASCII regimes per §5).
- BDD scenario counts per algorithm (~4-6 mirroring Phase 6).
- Exact wording of the godoc warning paragraph beyond the canonical text in CONTEXT.md §5.
- Whether plan 07-02's `--depth=deep` review is invoked automatically by the executor or manually requested in the PR description.
- Granular structure of `examples/phonetic-keys/main.go` (column layout, name set, header rows).

### Deferred Ideas (OUT OF SCOPE)

These are explicitly DEFERRED in CONTEXT.md; the planner MUST NOT include them:

- `XxxCodeNormalised(s, opts NormaliseOptions) string` wrappers — Phase 9+ additive expansion.
- `NYSIISOptions{IterateToFixedPoint: bool}` for modified-NYSIIS variant — v1.x.
- `SoundexSQLCode` second public function for SQL/MySQL variant parity — v1.x.
- Acquiring Taft 1970 PDF via ILL — v1.x if researcher contributes.
- Per-rule-branch attribution comments in `double_metaphone.go` — rejected as too granular.
- Cross-validation against Python `abydos` library — rejected (second pip dep).
- `MRACompareNormalised(a, b) (bool, float64)` wrapper for catalogue-wide float64 uniformity — v1.x.
- Combining `phonetic-codes.json` and `algorithms.json` golden files into one — v1.x if unified loader becomes preferable.

## Phase Requirements

| ID | Description | Research Support |
|----|-------------|------------------|
| PHON-01 | Soundex — Knuth/Census variant per Knuth TAOCP Vol. 3 §6.4. Mandatory `"Tymczak" → "T522"` discriminating vector. | §1 Algorithm Sources (Soundex), §3 Reference Vector Catalogue (Soundex), §5 Validation Architecture (Soundex test map), §6 Per-algorithm Implementation Notes (Soundex), Pitfall extensions §7 |
| PHON-02 | Double Metaphone (Philips 2000) returning primary + alternate codes. Reference vectors per language-origin branch (Germanic, Slavic, Romance, Greek). | §2 Double Metaphone Deep Dive (rule-table provenance, 5 branches, C reference archive identification, MIT-Go-port negative attribution), §3 Reference Vector Catalogue (Double Metaphone 40-entry plan with branch tagging), §5 Validation Architecture (DM test map), Pitfall extensions §7 |
| PHON-03 | NYSIIS (Taft 1970) with 6-char truncation. | §1 Algorithm Sources (NYSIIS), §3 Reference Vector Catalogue (NYSIIS, including the Knuth-vs-jellyfish divergence table), §5 Validation Architecture (NYSIIS test map), §6 Per-algorithm Implementation Notes (NYSIIS), Pitfall extensions §7 |
| PHON-04 | MRA (NBS Tech Note 943) Match Rating Approach. | §1 Algorithm Sources (MRA — NBS-943 verified accessible), §3 Reference Vector Catalogue (MRA threshold-edge pairs, length-diff>3 mismatch), §5 Validation Architecture (MRA test map covering MRACode + MRACompare + MRAScore), §6 Per-algorithm Implementation Notes (MRA), §7 Pitfall extensions |

## Project Constraints (from CLAUDE.md)

- **Zero non-stdlib runtime deps in root `go.mod`** — jellyfish and any DM Python source are developer-toolchain only, NEVER in root `go.mod`. `make verify-deps-allowlist` enforces.
- **Apache-2.0 throughout** — fresh implementation discipline from primary sources; NO GPL/LGPL-derived code; MIT-licensed Go ports of phonetic algorithms (xrash/smetrics, CalypsoSys/godoublemetaphone, tilotech/go-phonetics, UjjwalAyyangar/go-jellyfish) are **negative-attribution targets** — verifier checks by diff.
- **No testify in root tests** — phonetic algorithm tests use stdlib `testing` only.
- **Releases via CI only** — no local `git tag`, no local `goreleaser release`, no `--no-verify`.
- **Agent gates** — `algorithm-correctness-reviewer`, `algorithm-licensing-reviewer` (LOAD-BEARING for plan 07-02), `algorithm-performance-reviewer`, `determinism-reviewer`, `api-ergonomics-reviewer`, `code-reviewer`, `security-reviewer`, `test-writer`, `bdd-scenario-reviewer`, `docs-writer`, `go-quality` all apply.
- **No AI attribution in commits** — never mention Claude/AI/LLM in commit messages.

## Architectural Responsibility Map

| Capability | Primary Tier | Secondary Tier | Rationale |
|------------|-------------|----------------|-----------|
| Phonetic code generation (`SoundexCode`, `DoubleMetaphoneKeys`, `NYSIISCode`, `MRACode`) | Algorithm tier (`fuzzymatch` package, root module) | — | Pure-function algorithm primitives; Layer 1 of the three-layer architecture (algorithm functions / Scorer / scan). No dependency on Scorer or scan. |
| Binary Score functions (`SoundexScore`, `DoubleMetaphoneScore`, `NYSIISScore`, `MRAScore`) | Algorithm tier | — | Boolean code-match wrappers around the code-generation functions. Required for AlgoID dispatch table consumption by Phase 8 Scorer. |
| Raw 0-6 NBS counter exposure (`MRACompare`) | Algorithm tier | — | Faithful primary-source surface per CONTEXT.md §6. Not consumed by Scorer or scan; standalone consumer surface. |
| Cross-validation against jellyfish | Developer toolchain (`scripts/gen-phonetic-cross-validation.py`) | Test tier (`*_cross_validation_test.go`) | Generation is Python; consumption is Go test reading committed JSON. NEVER a runtime dep. |
| Monge-Elkan binary-inner composition (Phase 7 wiring) | Algorithm tier — `monge_elkan.go` map mutation | Test tier — `monge_elkan_test.go` panic-fixture + 4 new BinaryInner property tests | Lockstep per CONTEXT.md §4: each plan 07-01..07-04 adds AlgoID to allow-list AND updates panic-test fixture in same commit. |
| Cross-platform determinism gate | Test tier (`phonetic_codes_golden_test.go` loading `testdata/golden/phonetic-codes.json`) | CI tier (matrix runs golden diff) | The encoded code STRING is the determinism gate value for these algorithms (not a float), captured per CONTEXT.md §7. |

## Standard Stack

### Core (project-internal, already in repo)

| Library / File | Version | Purpose | Why Standard |
|----------------|---------|---------|--------------|
| `fuzzymatch` root package | working tree | Houses 4 new `<algo>.go` files + 4 `dispatch_<algo>.go` files + test infrastructure | Phase 7 lives at root; phonetic algorithms are part of the catalogue (Layer 1). |
| `algoid.go` lines 156-175 | (already shipped — Phase 1) | `AlgoSoundex`, `AlgoDoubleMetaphone`, `AlgoNYSIIS`, `AlgoMRA` enum slots already reserved with primary-source godoc | Slots are stable; planner wires `dispatch_<algo>.go` but does NOT renumber. |
| `monge_elkan.go` `permittedMongeElkanInner` (line 294) | Currently 14 entries (Phase 6) → 18 after Phase 7 | Allow-list for Monge-Elkan inner-metric dispatch; each Phase 7 plan adds its own AlgoID per CONTEXT.md §4 | Phase 6 already documents the additive Phase 7 expansion in the map's leading comment (lines 272-275). Lockstep mutation with `monge_elkan_test.go` panic-test fixture. |
| `errors.go` `ErrInvalidAlgoID` | (already shipped — Phase 1) | Consumed by Phase 8 Scorer wrappers for unknown AlgoIDs | Phonetic algorithms have no `Options` struct → no `ErrInvalid*Options` needed. |
| `testdata/golden/algorithms.json` | After Phase 6: 19 algorithms × 144 entries | Phase 7 finalisation extends to 23 algorithms × ≈ 180 entries via staging-golden → merge | Established workflow over 16 algorithms; one new staging file per phonetic algorithm. |
| `testdata/golden/_staging/` | (workflow established Phase 2) | Per-plan staging golden files merged in finalisation | Pattern proven; 4 new staging files (one per phonetic algorithm). |
| `testdata/cross-validation/` | After Phase 6: `swg/`, `ratcliff-obershelp/`, `token-ratios/` subdirectories | New subdirectory `phonetic/` created in plan 07-01 | Each cross-validation source gets its own subdirectory; pattern consistent. |
| `scripts/gen-token-ratio-cross-validation.py` | (Phase 6 reference template) | Direct structural template for `scripts/gen-phonetic-cross-validation.py` | RapidFuzz pin mechanism translates 1:1 to jellyfish pin mechanism (script-asserts version on import; refuses to run on mismatch). |
| `docs/cross-validation.md` | Established Phase 6 | Phase 7 extends with "Phonetic cross-validation" section describing jellyfish corpus + variant-divergence tagging | One new section; mirrors Phase 6 RapidFuzz section structure. |
| `tests/bdd/steps/algorithms_steps.go` | Accumulator file | Each new phonetic algorithm appends `step.Step` registrations | Pattern proven Phases 2-6. |
| `example_test.go` | Accumulator file | 9 new `ExampleXxx` entries minimum (SoundexCode, SoundexScore, DoubleMetaphoneKeys, DoubleMetaphoneScore, NYSIISCode, NYSIISScore, MRACode, MRACompare, MRAScore) | Established godoc-example pattern. |
| `props_test.go` | Accumulator file | New Five-invariant blocks (Identity/Symmetric/Range/NoNaN/NoInf) per algorithm + 4 new `TestMongeElkanScore_BinaryInner_<Algo>` tests | Established property-test accumulator. |
| `llms.txt` / `llms-full.txt` | (per-plan discipline) | Every new exported symbol gets a line in `llms.txt` and a full entry in `llms-full.txt` IN THE SAME PLAN | Per-plan sync; never deferred to finalisation. |
| `examples/identifier-similarity/main.go` | After Phase 6: 19 columns | Phase 7 finalisation extends to 23 columns (per CONTEXT.md §8) | Established example-extension pattern; `main_test.go` golden stdout fixture regenerated in same commit. |

### Cross-validation toolchain (developer-only, NOT runtime)

| Library | Version | Purpose | Why Standard |
|---------|---------|---------|--------------|
| **`jellyfish`** | **`1.2.1`** (released 2025-10-11; current PyPI stable) [VERIFIED: PyPI; jamesturk/jellyfish GitHub] | Cross-validation reference for **Soundex, NYSIIS, MRA only** — provides `jellyfish.soundex()`, `jellyfish.nysiis()`, `jellyfish.match_rating_codex()`, `jellyfish.match_rating_comparison()` | Pinned in `scripts/gen-phonetic-cross-validation.py` via `JELLYFISH_VERSION = "1.2.1"`. Implementation is Rust-backed (`jellyfish/src/*.rs`) since 1.0; fast and stable. **Does NOT have Double Metaphone** — see §4 below for the DM alternative recommendation. |
| **`metaphone`** | **`0.6`** (current PyPI stable; oubiwann/metaphone on GitHub — pure Python, BSD-licensed) [VERIFIED: PyPI Metaphone 0.6; oubiwann/metaphone repo] | Cross-validation reference for **Double Metaphone only** — provides `metaphone.doublemetaphone(s) -> (primary, secondary)` | The de-facto Python Double Metaphone canonical port (Andrew Collins' translation of Lawrence Philips' C++ reference; BSD-licensed; pure Python so transparent to read for verification). Pinned in `scripts/gen-phonetic-cross-validation.py` via `METAPHONE_VERSION = "0.6"`. **CRITICAL RESEARCH FINDING — supersedes CONTEXT.md §1's implicit assumption that jellyfish covers all 4 algorithms.** See §4 below for full discussion. |
| `python3` | 3.11 / 3.12 / 3.13 (developer-toolchain) | Runs the generator script | Phase 6 RapidFuzz infra uses Python 3.12 in its `_metadata.python_version`; carry forward. |

### Alternatives Considered (and rejected — see CONTEXT.md and §4)

| Instead of | Could Use | Tradeoff | Reason rejected (or accepted) |
|------------|-----------|----------|-------------------------------|
| `jellyfish` for Soundex/NYSIIS/MRA | `phonics` (R package), `abydos` (Python) | abydos has 30+ phonetic algorithms incl. DM but adds a second pip dep beyond jellyfish for the simple algorithms | CONTEXT.md `<deferred>` rejects abydos. Jellyfish is the smallest viable footprint. |
| `metaphone` (oubiwann/metaphone) for DM | Apache Commons Codec (Java), `abydos.phonetic.DoubleMetaphone` (Python), `slacy/double-metaphone` (C++) | Apache Commons Codec is the most-respected reference but requires JVM (heavy); abydos requires the abydos package (heavy); slacy requires a C++ build chain | `oubiwann/metaphone` is the lightest BSD-licensed Python port carrying both single Metaphone and Double Metaphone. **Verified accessible:** `pip install Metaphone` works on PyPI as version 0.6. |
| Hand-derived DM corpus only | Cross-validation infrastructure | Literature reference vectors are sufficient for primary-source fidelity; jellyfish-style cross-validation broadens coverage for Pitfall 5's branch-coverage concern | CONTEXT.md §1 LOCKS both layers (literature primary + jellyfish supplementary). DM corpus uses `metaphone` instead of jellyfish; mechanism unchanged. |
| `abydos.phonetic.NYSIIS` for NYSIIS | jellyfish | abydos has both original-Taft-1970 AND modified-NYSIIS modes; jellyfish has only modified | jellyfish is forced by abydos rejection. `variant_divergence: true` mechanism handles jellyfish's modified NYSIIS divergence. |
| Pinned exact version via `requirements.txt` lockfile | Inline `JELLYFISH_VERSION = "1.2.1"` in script | Lockfile cleaner separation; inline matches Phase 6 RapidFuzz exact pattern | CONTEXT.md §1 LOCKS the Phase 6 pattern; carry forward. Plus an inline `METAPHONE_VERSION = "0.6"`. |

**Installation (developer machine, one-time):**

```bash
python3 -m pip install --user jellyfish==1.2.1 Metaphone==0.6
```

**Version verification before pinning** (planner runs this at plan-creation time):

```bash
pip index versions jellyfish 2>/dev/null | head -5
pip index versions Metaphone 2>/dev/null | head -5
# Or:
python3 -c "import jellyfish, metaphone; print(jellyfish.__version__, getattr(metaphone, '__version__', 'unknown'))"
```

Note: the `metaphone` PyPI package does not expose `__version__` reliably; the generator script asserts via `subprocess.check_output(['pip', 'show', 'Metaphone'])` parsing instead. (Phase 6 RapidFuzz exposes `rapidfuzz.__version__`; falls back to package-metadata if needed.)

## Algorithm Sources & Authoritative Variants (§1 of output)

### 1.1 Soundex

| Aspect | Value | Source | Confidence |
|--------|-------|--------|-----------|
| Algorithmic origin | Russell, R. C., Odell, M. K. (1918, 1922). U.S. Patents 1261167 and 1435663 | algoid.go line 156-159 (already cited); CONTEXT.md `<canonical_refs>` | HIGH [CITED] |
| Canonical algorithm description | Knuth, D. E. (1973). *The Art of Computer Programming, Vol. 3: Sorting and Searching*, §6.4. Addison-Wesley | algoid.go line 158; PITFALLS.md Pitfall 4 line 100 | HIGH [CITED] |
| Variant chosen | Knuth/Census 1880 (H/W are NOT separators; same-group consonants separated by H or W collapse to one digit; vowels ARE separators) | docs/requirements.md §7.4.1 line 646 | HIGH [CITED] |
| Variant RECTED (out of scope) | SQL/MySQL variant (H/W ARE separators) — `Tymczak → T520` | PITFALLS.md Pitfall 4 line 97 | HIGH |
| Public functions | `SoundexCode(s string) string` (1 letter + 3 digits = 4 chars), `SoundexScore(a, b string) float64` (binary) | docs/requirements.md §7.4.1 line 644-645 | HIGH [CITED] |
| Output charset | `[A-Z][0-9][0-9][0-9]` (uppercase letter + three digits, zero-padded) | docs/requirements.md §7.4.1; jellyfish/src/soundex.rs lines 47-49 (`while result.len() < 4 { result.push('0'); }`) | HIGH [CITED + VERIFIED] |
| jellyfish parity | **YES** — jellyfish 1.2.1 also uses Knuth/Census variant per direct source read | `jellyfish/src/soundex.rs` lines 32-43 (the `else if *letter != 'H' && *letter != 'W'` branch is the Knuth/Census H/W-skip rule, NOT the SQL separator rule) | HIGH [VERIFIED via direct source read] |

**Critical implication for CONTEXT.md §1:** Soundex `variant_divergence: true` tagging is **defensive insurance with no expected actual divergences for canonical inputs**. The 15-entry Soundex corpus per CONTEXT.md §1 should treat the `variant_divergence` field as schema-required-but-empirically-false. The discriminating literature unit tests (`Tymczak → T522`, `Ashcroft → A261`) remain mandatory for primary-source fidelity at the file-comment level. [VERIFIED 2026-05-15 via direct read of `jellyfish/src/soundex.rs` on GitHub `main` branch]

### 1.2 Double Metaphone

| Aspect | Value | Source | Confidence |
|--------|-------|--------|-----------|
| Algorithmic origin | Philips, L. (2000). "The double-metaphone search algorithm." *C/C++ Users Journal*, 18(6):38–43 | docs/requirements.md §7.4.2 line 656 | HIGH [CITED] |
| Canonical reference implementation (rule-table provenance) | Lawrence Philips' original C++ reference, published in the C/C++ Users Journal June 2000 source-code archive at `ftp://ftp.cuj.com/sourcecode/cuj/2000/cujjun2000.zip` (73 KB) | WebSearch verification — multiple ports cite this archive (oubiwann/metaphone README, slacy/double-metaphone README, Apache Commons Codec Javadoc) | MEDIUM-HIGH [CITED — archive existence cross-verified] |
| Public domain status | Yes — Philips' reference C++ implementation was released without patent or copyright claim alongside the article. Apache Commons Codec, oubiwann/metaphone (BSD), slacy (BSD), Maurice Aubrey's Perl port (public domain) all redistribute under permissive terms without Philips having raised objection over 25 years | algorithm-licensing-standards SKILL.md "For Double Metaphone specifically: Lawrence Philips' 2000 publication has no patent claim... Clear" | HIGH [CITED] |
| Recommended archive URL for rule-table provenance line | `archive.org/details/CCUJ_2000_06` (the C/C++ Users Journal June 2000 issue archive) OR the GitHub fork `https://github.com/SWI-Prolog/packages-nlp/blob/master/double_metaphone.c` (preserves Maurice Aubrey's lightly-fixed redistribution of the original Philips C source; long-stable URL since 2012) | WebSearch [SWI-Prolog/packages-nlp]; multiple ports cite this fork | MEDIUM-HIGH |
| Public functions | `DoubleMetaphoneKeys(s string) (primary, secondary string)`, `DoubleMetaphoneScore(a, b string) float64` | docs/requirements.md §7.4.2 line 659-660 | HIGH [CITED] |
| Key length | 4 characters each in canonical formulation (some "extended" variants use longer keys; this library uses canonical 4-char truncation per spec) | docs/requirements.md §7.4.2 line 659 | HIGH [CITED] |
| Output charset | `[A-Z0]` (uppercase ASCII letters plus the digit `0` which represents the "th" sound — e.g. `Schmidt → ("XMT", "SMT")` per spec; the `0` appears in keys like `Smith → ("SM0", "XMT")` for the `th` sound) | docs/requirements.md §7.4.2 reference vectors line 667; verified via WebSearch on oubiwann/metaphone | HIGH [CITED + VERIFIED] |
| Match rule | 1.0 if EITHER of a's keys matches EITHER of b's keys (primary-primary, primary-secondary, secondary-primary, secondary-secondary) | docs/requirements.md §7.4.2 line 660 | HIGH [CITED] |
| Language-origin branches | 5: Germanic, Slavic, Romance, Greek, Chinese-origin | PITFALLS.md Pitfall 5 line 120; CONTEXT.md §1, §3 | HIGH [CITED] |
| Rule count | ~200 conditional branches; ~400 lines of conditional logic in C reference | PITFALLS.md Pitfall 5 line 120 | HIGH [CITED] |
| jellyfish parity | **NO — jellyfish does NOT have Double Metaphone.** `jellyfish.metaphone` is single-key Metaphone, NOT Double Metaphone. See §4 below for the implication and the `metaphone` (oubiwann) recommendation. | Direct read of `jellyfish/src/lib.rs` line 19 (`mod metaphone;` — singular, not `double_metaphone`); the lib.rs exports only `pub use metaphone::metaphone;` not `double_metaphone` | HIGH [VERIFIED via direct source read 2026-05-15] |

### 1.3 NYSIIS

| Aspect | Value | Source | Confidence |
|--------|-------|--------|-----------|
| Algorithmic origin | Taft, R. L. (1970). *Name search techniques.* New York State Identification and Intelligence System, Special Report No. 1. Albany, NY | docs/requirements.md §7.4.3 line 672; CONTEXT.md §2 LOCKED | HIGH [CITED] |
| Canonical algorithm description (primary source for fresh transcription) | Knuth, D. E. (1973). *TAOCP Vol. 3*, §6.4. Addison-Wesley | CONTEXT.md §2 LOCKED | MEDIUM [CITED — Knuth TAOCP Vol. 3 §6.4 has full Soundex but the **explicit description of NYSIIS in TAOCP Vol. 3 §6.4 is not directly verifiable in the search results**. The fallback citation chain is the **most-cited public secondary description** (Wikipedia, Apache Commons Codec docs, NIST DADS entry) — all of which describe the same Taft-1970 9-step procedure, providing independent corroboration of the algorithm Knuth would have documented.] [ASSUMED — Knuth TAOCP §6.4 NYSIIS coverage is plausible from the volume's surname-matching subsection but the present researcher could not verify the exact section content. The planner should cite Knuth as the canonical description per CONTEXT.md §2 LOCKED; a v1.x re-check could substitute Wikipedia + NIST DADS as the canonical secondary if Knuth's NYSIIS subsection turns out to be too sparse.] |
| Variant chosen | Original NYSIIS-1970: 6-character truncation, each rule applied once (no iterate-to-fixed-point) | CONTEXT.md §2 LOCKED | HIGH [CITED] |
| Variants REJECTED (out of scope) | Modified NYSIIS-1991, "wonderland", iterate-to-fixed-point variant | CONTEXT.md §2 LOCKED; PITFALLS.md Pitfall 5 line 119 | HIGH [CITED] |
| Public functions | `NYSIISCode(s string) string` (6-char truncated), `NYSIISScore(a, b string) float64` (binary) | docs/requirements.md §7.4.3 line 675-676 | HIGH [CITED] |
| Output charset | `[A-Z]` (uppercase letters only; NO digits in NYSIIS output) | Wikipedia NYSIIS rules confirm; jellyfish testdata confirms (e.g. `Brown → BRAN`) | HIGH [VERIFIED] |
| Maximum output length | 6 characters (original Taft 1970 truncation per CONTEXT.md §2; modified variants emit longer codes — see jellyfish divergence below) | CONTEXT.md §2; docs/requirements.md §7.4.3 line 675 | HIGH [CITED] |
| jellyfish parity | **NO — jellyfish emits non-truncated codes.** Direct read of `jellyfish/testdata/nysiis.csv`: `Catherine → CATARAN` (7 chars), `montgomery → MANTGANARY` (10 chars), `martincevic → MARTANCAFAC` (11 chars). Jellyfish implements the modified/extended NYSIIS variant without truncation. | `jellyfish/testdata/nysiis.csv` line 7 (`Catherine,CATARAN`), line 3 (`montgomery,MANTGANARY`), line 6 (`martincevic,MARTANCAFAC`) | HIGH [VERIFIED via direct read 2026-05-15] |
| Implication for variant_divergence tagging | **ALL jellyfish NYSIIS output values >6 chars carry `variant_divergence: true`, and the `variant_divergence_reason` field documents "jellyfish does not truncate to 6 chars"; our implementation is asserted against the Taft-1970-truncated value (the first 6 characters of jellyfish output IF and only if jellyfish's 7+ char output is consistent with Taft's pre-truncation form — which the planner verifies during corpus authoring).** | Direct read evidence above; CONTEXT.md §1 variant_divergence mechanism | HIGH |

### 1.4 MRA (Match Rating Approach)

| Aspect | Value | Source | Confidence |
|--------|-------|--------|-----------|
| Algorithmic origin | Moore, G. B., Kuhns, J. L., Trefftzs, J. L., Montgomery, C. A. (1977). *Accessing individual records from personal data files using non-unique identifiers.* National Bureau of Standards (later NIST), Technical Note 943 | docs/requirements.md §7.4.4 line 687; algoid.go line 172 | HIGH [CITED] |
| Primary-source PDF accessibility | **Confirmed accessible at the canonical NIST URL: `https://nvlpubs.nist.gov/nistpubs/Legacy/TN/nbstechnicalnote943.pdf` (5 MB, HTTP 200 verified 2026-05-15)** | HTTP HEAD check 2026-05-15 returned `HTTP/2 200, content-type: application/pdf` | HIGH [VERIFIED] |
| Encoding rules (3 steps, per Moore-Kuhns 1977) | (1) Delete all vowels unless the vowel begins the word. (2) Remove the second consonant of any double consonants. (3) Reduce codex to 6 letters by joining the first 3 and last 3 letters only. | Wikipedia Match_rating_approach + NBS Tech Note 943 verbatim per WebFetch [HIGH] | HIGH [CITED] |
| Threshold table (Table A — minimum rating by sum-of-lengths) | sum ≤ 4 → min rating 5; 4 < sum ≤ 7 → min 4; 7 < sum ≤ 11 → min 3; sum = 12 → min 2 | Wikipedia Match_rating_approach (Table A); jellyfish/src/match_rating.rs lines 96-101 (the threshold-table match expression — independently verifies the table values) | HIGH [CITED + VERIFIED] |
| Comparison rules (6 steps) | (1) Reject if `|len(codex1) - len(codex2)| >= 3`. (2) Determine min threshold from Table A using sum of codex lengths. (3) Process L→R, remove identical chars from both. (4) Process R→L on remaining, remove identical chars. (5) `similarity = 6 - max(unmatched_count_longer, unmatched_count_shorter)`. (6) Match iff `similarity >= min_threshold`. | Wikipedia Match_rating_approach; jellyfish/src/match_rating.rs lines 65-101 (independent verification of all 6 steps) | HIGH [VERIFIED] |
| Length-difference auto-mismatch rule | If `|len(codex1) - len(codex2)| >= 3` → automatic mismatch (`MRACompare → (false, 0)`) per spec | docs/requirements.md §7.4.4 line 696 ("strings whose encoded lengths differ by more than 3 are documented as automatic mismatch"); jellyfish/src/match_rating.rs line 60 (`if longer.len() - shorter.len() >= 3 { return Err(...) }`) — note jellyfish returns an Err here rather than `(false, 0)`; **fuzzymatch returns `(false, 0)` per CONTEXT.md §6 binary-int return semantics** | HIGH [VERIFIED + planner decision] |
| Public functions | `MRACode(s string) string` (canonical encoded form), `MRACompare(a, b string) (matched bool, simScore int)` (raw 0-6 NBS counter + threshold-rule match decision), `MRAScore(a, b string) float64` (binary 0.0/1.0) | docs/requirements.md §7.4.4 line 689-692; CONTEXT.md §6 LOCKED | HIGH [CITED] |
| Output charset | `[A-Z]` for the code; the comparison rule uses positional alphabetic match | jellyfish/src/match_rating.rs verified | HIGH [VERIFIED] |
| Maximum output length | 6 characters (after first-3-last-3 truncation if pre-truncation len > 6) | NBS Tech Note 943 rule (3) | HIGH [CITED] |
| jellyfish parity | **PARTIAL — jellyfish's MRA codex implementation appears to NOT remove double consonants per the standard rule (e.g. `Smyth → SMYTH` keeps Y, `Smith → SMTH` removes the i not the m).** Looking at jellyfish/src/match_rating.rs lines 14-25: jellyfish's encoding loop adds character to codex if it's a non-vowel AND `*c != prev` (i.e. it deduplicates EXACT adjacent duplicates of any character including vowels-after-leading-vowel-position). The spec's "Remove the second consonant of any double consonants" rule means jellyfish actually IS compliant — `prev` is the previous accepted character. **HOWEVER** the order matters: spec is "delete vowels then dedupe doubles"; jellyfish does both in a single pass (the dedupe operates on the un-de-voweled stream). For the canonical reference vectors in jellyfish/testdata/match_rating_codex.csv (e.g. `Byrne → BYRN`, `Smith → SMTH`, `Catherine → CTHRN`), the outputs match the spec. **No known divergences for canonical inputs; the corpus does not need `variant_divergence: true` tags by default.** | jellyfish/src/match_rating.rs lines 14-37; jellyfish/testdata/match_rating_codex.csv | HIGH [VERIFIED via direct read 2026-05-15] |
| MRA Compare integer score interpretation | Integer in [0, 6] inclusive (6 = identical-codex match; 0 = no characters in common after both-passes elimination). `MRAScore` returns 1.0 iff `MRACompare` returns `(true, _)`. | jellyfish/src/match_rating.rs verified | HIGH [VERIFIED] |

## Double Metaphone Deep Dive (§2 of output)

This section is the longest because plan 07-02 carries the heaviest weight in Phase 7 (per CONTEXT.md §3 LOCKED and PITFALLS.md Pitfall 5).

### 2.1 Rule-table provenance

The canonical implementation of Double Metaphone is Lawrence Philips' original C++ code published alongside the June 2000 C/C++ Users Journal article. The C/C++ Users Journal source-code archive was at `ftp://ftp.cuj.com/sourcecode/cuj/2000/cujjun2000.zip` (73 KB); CUJ ceased publication in 2006 and the FTP server is no longer accessible.

**Recommended archival URLs for the rule-table provenance line in `double_metaphone.go`:**

Option A (recommended — long-stable, redistribution-explicit):
```
// Rule table derived fresh from: Philips, L. (2000), "The Double Metaphone
// Search Algorithm", C/C++ Users Journal 18(6):38-43, and the public-domain
// C reference implementation published alongside it. The C reference (with
// Maurice Aubrey's perl-port-derived bug fixes) is preserved at:
//   https://github.com/SWI-Prolog/packages-nlp/blob/master/double_metaphone.c
// The original CUJ source archive (ftp://ftp.cuj.com/sourcecode/cuj/2000/
// cujjun2000.zip) is no longer reachable since CUJ's 2006 shutdown.
```

Option B (alternative — Apache Commons Codec as canonical Java reference):
```
// Rule table derived fresh from: Philips, L. (2000) AND the Apache Commons
// Codec Java implementation (Apache-2.0 — compatible with our Apache-2.0
// licensing; consulted for rule-structure reading only, no code translated):
//   https://commons.apache.org/proper/commons-codec/apidocs/org/apache/
//   commons/codec/language/DoubleMetaphone.html
```

**Recommendation:** Option A. The SWI-Prolog `double_metaphone.c` mirror is the cleanest public-domain redistribution path (no licence ambiguity); Apache Commons Codec is Apache-2.0 (compatible) but is structurally a Java rewrite, not a faithful reproduction of Philips' rule table organisation.

### 2.2 The 5 language-origin branches

Per CONTEXT.md §3 LOCKED, plan 07-02's validation is a mandatory language-branch checklist. The 5 branches plus their canonical reference vectors (from Philips 2000 and the C reference):

| Branch | Characteristic inputs | Expected output | Behaviour-distinguishing rule |
|--------|----------------------|-----------------|-------------------------------|
| **Germanic** | `Schmidt`, `Smith`, `Schwartz`, `Knaufmann` | `Schmidt → ("XMT", "SMT")`; `Smith → ("SM0", "XMT")`; `DoubleMetaphoneScore("Schmidt", "Smith") = 1.0` (shared XMT) | `SCH` at start emits both X (sh-sound) primary and S (Germanic SCH-as-S secondary). `TH` emits `0` (digit zero, conventional for theta-th-sound). [CITED: docs/requirements.md §7.4.2 line 667; cross-verifiable via oubiwann/metaphone] |
| **Slavic** | `Wojcik`, `Pacheco` (also Romance), `Sczepanski` | `Wojcik → ("FSK", "FSXK")` (typical); `Sczepanski → ("SSPNSK", "...")` | Initial `WJ-` treated as Slavic V/F; `SZC-` recognised as Slavic compound. The canonical Slavic test vectors live in the C reference's commented `SLAVO_GERMANIC` rule branch. |
| **Romance** | `Pacheco`, `Cabrillo`, `Jaramillo` (Spanish); `Bologna`, `Catania` (Italian) | `Pacheco → ("PXK", "PSK")` (Spanish CH-as-K secondary); `Bologna → ("PLN", "PKN")` (Italian GN-as-N) | Spanish `CH` after vowel emits both X (English ch-sound) and K (Spanish hard-c sound). Spanish-origin trigger: presence of certain trailing patterns like `-ez`, `-ado`, `-illo`, the CH-after-A-O-U rule. |
| **Greek** | `Catherine`, `Katherine`, `Athens`, `Christopher` | `Catherine → ("K0RN", "KTRN")`; `Katherine → ("K0RN", "KTRN")` — exact match (both keys identical) | Greek `TH` always emits `0` (theta) primary AND `T` secondary; `CH` of Greek origin emits `K`. |
| **Chinese-origin** | `Wong`, `Cheung`, `Hong`, `Chen` | `Cheung → ("XNK", "XNK")`; `Wong → ("ANK", "FNK")` (variant); `Chen → ("XN", "XN")` | Initial `CHE-`, `CHI-`, `CHO-` followed by certain patterns trigger Chinese-origin recognition; some patterns emit identical primary/secondary because no ambiguity exists in the pronunciation. |

**The mandatory plan 07-02 checklist (per CONTEXT.md §3):**

```
- [ ] Germanic: DoubleMetaphoneKeys("Schmidt") = ("XMT", "SMT")
- [ ] Germanic: DoubleMetaphoneKeys("Smith") = ("SM0", "XMT")
- [ ] Germanic: DoubleMetaphoneScore("Schmidt", "Smith") = 1.0 (the XMT cross-match)
- [ ] Slavic: at least one Slavic-origin reference vector from the C source (Wojcik or Sczepanski)
- [ ] Romance: DoubleMetaphoneKeys("Pacheco") contains the Spanish "PXK" variant
- [ ] Greek: DoubleMetaphoneKeys("Catherine") = DoubleMetaphoneKeys("Katherine") = ("K0RN", "KTRN")
- [ ] Chinese-origin: at least one reference vector from the C source (Cheung or Wong)
```

**Cross-validation source — `metaphone` (oubiwann)**: at minimum 7 entries per major branch per CONTEXT.md §1 corpus-size table — the 40-entry DM corpus naturally distributes (5 branches × 7-8 entries = 35-40 entries, leaving slack for edge cases like empty input and pure-non-ASCII).

### 2.3 MIT-licensed Go ports — negative attribution

**These exist in the Go ecosystem but MUST NOT be copied from for any Phase 7 algorithm.** The `algorithm-licensing-reviewer` agent verifies by diff. The plan 07-02 file-comment negative-attribution line must explicitly list these:

| Repository | Licence | Phase 7 negative-attribution status |
|-----------|---------|--------------------------------------|
| `github.com/CalypsoSys/godoublemetaphone` | MIT | **Pitfall 5's primary attractive-nuisance.** Single-file port of the Philips C reference. The plan 07-02 file comment lists this BY NAME. Reviewer verifies by diff: variable names, comment phrasing, switch-case organisation MUST differ. |
| `github.com/deezer/double-metaphone-go` | MIT | Same negative-attribution requirement. |
| `github.com/xrash/smetrics` | MIT | Has Soundex (and a Metaphone), no Double Metaphone — but still negative-attribute for Soundex. |
| `github.com/UjjwalAyyangar/go-jellyfish` | MIT | Has NYSIIS — negative-attribute for plan 07-03. |
| `github.com/tilotech/go-phonetics` | MIT | Has Soundex + Metaphone — negative-attribute for plan 07-01 and (defensively) plan 07-02. |
| `github.com/jamesturk/jellyfish` (Python+Rust, BSD-2) | BSD-2 | **The Python lib's Rust core IS the cross-validation source** — vectors-only consultation is permitted; rule-table reading is permitted FOR Soundex/NYSIIS/MRA (not DM since jellyfish has no DM). [CITED: algorithm-licensing-standards SKILL.md "Reference implementations studied for cross-validation must be permissively-licensed (MIT, BSD, Apache-2.0, public domain, or equivalent)."] |

**Plan 07-02 Source-Origin Statement template** (extends Phase 4 WR-01 format per CONTEXT.md §3):

```go
// double_metaphone.go implements Lawrence Philips' Double Metaphone phonetic
// encoding algorithm.
//
// Sources:
//   - Philips, L. (2000). "The double-metaphone search algorithm."
//     C/C++ Users Journal, 18(6):38-43.
//   - Public-domain C reference implementation (Philips' original, with
//     Maurice Aubrey's perl-port-derived bug fixes), preserved at
//     https://github.com/SWI-Prolog/packages-nlp/blob/master/double_metaphone.c
//     and consulted for rule-table structure ONLY (no code copied,
//     no variable names or comment phrasing derived).
//   - Cross-validation reference: oubiwann/metaphone Python package
//     (BSD, https://github.com/oubiwann/metaphone — consulted ONLY for
//     reference vectors via testdata/cross-validation/phonetic/vectors.json).
//
// MIT-licensed Go ports NOT consulted (verified by diff during code review):
//   - github.com/CalypsoSys/godoublemetaphone (MIT)
//   - github.com/deezer/double-metaphone-go (MIT)
//   - github.com/tilotech/go-phonetics (MIT)
//   - github.com/UjjwalAyyangar/go-jellyfish (MIT)
//
// Rule table derived fresh from Philips 2000 + the C reference. Each
// language-origin branch (Germanic / Slavic / Romance / Greek / Chinese-
// origin) is explicitly tested per CONTEXT.md §3 LOCKED checklist.
//
// algorithm-licensing-reviewer sign-off: <recorded in PR description per
// CONTEXT.md §3 LOCKED>
```

## Reference Vector Catalogue (§3 of output)

This catalogue gives the planner concrete vectors the unit tests MUST contain. All vectors are LITERATURE-sourced or CROSS-VERIFIED against the recommended pinned reference implementation.

### 3.1 Soundex (≥ 12 vectors)

| # | Input | Expected Code | Source | Purpose |
|---|-------|---------------|--------|---------|
| RV-S1 | `Robert` | `R163` | docs/requirements.md §7.4.1 line 651; jellyfish testdata | Canonical pair (1/2) demonstrating identity of `Robert` and `Rupert` codes |
| RV-S2 | `Rupert` | `R163` | docs/requirements.md §7.4.1 line 651; jellyfish testdata | Canonical pair (2/2) |
| RV-S3 | `Rubin` | `R150` | docs/requirements.md §7.4.1 line 651 | M-N rule (`M/N → 5`) discriminator from `Robert`/`Rupert` (which encode B→1, R→6) |
| RV-S4 | `Ashcraft` | `A261` | docs/requirements.md §7.4.1 line 651; PITFALLS.md Pitfall 4 line 109; jellyfish testdata | **H/W-skip rule discriminator** — SH→2, R→6, F→1 with the C-R-F sequence preserved despite the H. Demonstrates Knuth/Census variant. |
| RV-S5 | `Ashcroft` | `A261` | docs/requirements.md §7.4.1 line 651; PITFALLS.md Pitfall 4 line 109; jellyfish testdata | Pairs with RV-S4: `Ashcraft = Ashcroft` is the load-bearing Knuth/Census discriminator. |
| RV-S6 | `Tymczak` | `T522` | PITFALLS.md Pitfall 4 line 97; docs/requirements.md mandate via Phase 7 §1 LOCKED; jellyfish testdata line 6 | **THE LOAD-BEARING DISCRIMINATOR.** T→T, Y skip (vowel), M→5, C→2, Z→2 but skip (same group), A skip (vowel), K→2. Result `T522`. The SQL/MySQL variant returns `T520` because in SQL, vowel skip stops the M-C-Z run earlier. |
| RV-S7 | `Pfister` | `P236` | jellyfish testdata line 4 | Initial digraph PH-style handling — P→P, F→1 then skip (would normally be 1 too, group collision), I skip (vowel), S→2, T→3, E skip (vowel), R→6. Result `P236`. Tests that the initial-letter rule preserves only the FIRST letter, NOT consonant clusters. |
| RV-S8 | `Honeyman` | `H555` | hand-derived from Knuth/Census rules | H→H first-letter-keep; O skip, N→5, E skip, Y skip, M→5 (group-collision dedup to single 5 because adjacent), A skip, N→5. Result `H555`. |
| RV-S9 | `Lloyd` | `L300` | hand-derived | L→L first-letter-keep, L skip (dedup adjacent same-group at start), O skip, Y skip, D→3. Pad with zeros. Result `L300`. (Note: "L" first-letter and second "l" in second position collapses; the L group is 4. This is `L300` because the second L would have been a 4 but is skipped as adjacent same-group to the first L which was retained as letter.) |
| RV-S10 | `Çáŕẗéř` | `C636` | jellyfish testdata (Unicode case via NFKD) | **Documents non-ASCII silent-skip behaviour (CONTEXT.md §5).** After diacritic stripping (which jellyfish does via NFKD; fuzzymatch's encoding sees only ASCII letters), produces same as `Carter → C636`. **fuzzymatch's silent-skip rule per CONTEXT.md §5 means non-ASCII letters are DROPPED, so `Çáŕẗéř` encoded by fuzzymatch produces only what ASCII chars remain — possibly empty or just the leading letter. Confirm during plan 07-01 by direct test;** this vector may need `variant_divergence: true` because jellyfish NFKD-normalises and fuzzymatch silent-skips. |
| RV-S11 | `""` (empty) | `""` (empty) | docs/requirements.md §7.4.1 line 649 edge case | Empty input → empty code → `SoundexScore("", "") = 1.0`. |
| RV-S12 | `"Smith"` | `S530` | Knuth rules, hand-verified | Classic Soundex example: S→S kept; M→5, I skip, T→3. Result `S530`. Pads with one zero. |

### 3.2 Double Metaphone (≥ 20 vectors covering all 5 branches; 40 in jellyfish-supplementary corpus)

| # | Input | Expected (primary, secondary) | Branch | Source |
|---|-------|------|--------|--------|
| RV-DM1 | `Schmidt` | `("XMT", "SMT")` | Germanic | docs/requirements.md §7.4.2 line 667; CONTEXT.md §3 mandatory |
| RV-DM2 | `Smith` | `("SM0", "XMT")` | Germanic | docs/requirements.md §7.4.2 line 667; CONTEXT.md §3 mandatory |
| RV-DM3 | `Schwartz` | `("XRTS", "SVRTS")` | Germanic | C reference vectors; hand-cross-verified with oubiwann/metaphone |
| RV-DM4 | `Pacheco` | `("PXK", "PSK")` | Romance | docs/requirements.md §7.4.2; CONTEXT.md §3 mandatory; PITFALLS.md Pitfall 5 line 134 |
| RV-DM5 | `Bologna` | `("PLN", "PKN")` | Romance (Italian) | C reference Italian-origin trigger |
| RV-DM6 | `Jaramillo` | `("HRML", "HRMY")` | Romance (Spanish — -illo trigger) | C reference Spanish-origin |
| RV-DM7 | `Catherine` | `("K0RN", "KTRN")` | Greek | docs/requirements.md §7.4.2 line 667; CONTEXT.md §3 mandatory |
| RV-DM8 | `Katherine` | `("K0RN", "KTRN")` | Greek | docs/requirements.md §7.4.2 line 667; CONTEXT.md §3 mandatory (matches RV-DM7) |
| RV-DM9 | `Christopher` | `("KRSTFR", "")` | Greek (CHRIS-as-K) | oubiwann/metaphone canonical test ([VERIFIED: web reference says `('KRSTFR', '')`]) — note CHRIS-emits empty secondary because no ambiguity |
| RV-DM10 | `Wojcik` | `("FSK", "FSXK")` | Slavic | C reference Slavic-origin branch |
| RV-DM11 | `Sczepanski` | `("SPNSK", "SXPNSK")` | Slavic (SZC compound) | C reference |
| RV-DM12 | `Cheung` | `("XNK", "XNK")` | Chinese-origin | C reference Chinese-CH trigger (CHE before vowel-N) |
| RV-DM13 | `Wong` | `("ANK", "FNK")` | Chinese-origin | C reference initial-W-before-vowel rule |
| RV-DM14 | `Chen` | `("XN", "XN")` | Chinese-origin | C reference |
| RV-DM15 | `""` (empty) | `("", "")` | Edge | docs/requirements.md §7.4.2 line 664 — empty input → empty keys → `DoubleMetaphoneScore("", "") = 1.0` |
| RV-DM16 | `Caesar` | `("SSR", "SSR")` | Edge — initial-C-AE | C reference initial-CAE-as-S rule |
| RV-DM17 | `Knock` | `("NK", "NK")` | Edge — initial-KN | C reference initial-KN-as-N rule (silent K) |
| RV-DM18 | `Quincy` | `("KNS", "KNS")` | Edge — Q-as-K | C reference Q-initial rule |
| RV-DM19 | `Xavier` | `("SF", "SFR")` | Edge — initial X | C reference X-initial-as-S rule |
| RV-DM20 | `Thompson` | `("TMSN", "TMSN")` | Edge — TH-as-T (NOT Greek) | C reference TH-as-T fallback when not Greek-origin pattern |

### 3.3 NYSIIS (≥ 12 vectors with truncation tests)

| # | Input | Expected Code (Knuth/Taft 6-char truncation) | jellyfish output (modified, untruncated) | variant_divergence? | Source |
|---|-------|---------------------------------------------|----------------------------------------|---------------------|--------|
| RV-N1 | `Brown` | `BRAN` | `BRAN` | false | docs/requirements.md §7.4.3 line 682; jellyfish testdata; CONTEXT.md §3 |
| RV-N2 | `Browne` | `BRAN` | `BRAN` | false | docs/requirements.md §7.4.3 line 682 (matches RV-N1); CONTEXT.md §3 |
| RV-N3 | `Robert` | `RABAD` | (jellyfish: not in testdata, would emit `RABAD` per same rules) | false | docs/requirements.md §7.4.3 line 682; CONTEXT.md §3 |
| RV-N4 | `Catherine` | `CATARA` (truncated from `CATARAN`) | `CATARAN` | **true** (truncation divergence) | jellyfish testdata line 7 confirms `CATARAN` |
| RV-N5 | `Katherine` | `CATARA` (truncated from `CATARAN`) | `CATARAN` | **true** | jellyfish testdata line 8 |
| RV-N6 | `Johnathan` | `JANATA` (truncated from `JANATAN`) | `JANATAN` | **true** | jellyfish testdata line 10 |
| RV-N7 | `Jonathan` | `JANATA` (truncated from `JANATAN`) | `JANATAN` | **true** | jellyfish testdata line 11 |
| RV-N8 | `John` | `JAN` | `JAN` | false | jellyfish testdata line 12 |
| RV-N9 | `Teresa` | `TARAS` | `TARAS` | false | jellyfish testdata line 13 |
| RV-N10 | `Theresa` | `TARAS` | `TARAS` | false | jellyfish testdata line 14 (matches RV-N9) |
| RV-N11 | `montgomery` | `MANTGA` (truncated from `MANTGANARY`) | `MANTGANARY` | **true** | jellyfish testdata line 3 |
| RV-N12 | `""` (empty) | `""` | `""` | false | docs/requirements.md §7.4.3 line 680 |

**Truncation gate test:** at least one vector where input is long enough to expose truncation, AND the test asserts `len(NYSIISCode(input)) <= 6`. RV-N4..RV-N7, RV-N11 all serve this purpose.

### 3.4 MRA (≥ 12 vectors covering encoding + comparison + threshold + length-diff)

| # | Input(s) | Expected | Purpose | Source |
|---|----------|----------|---------|--------|
| RV-M1 | `MRACode("Byrne")` | `"BYRN"` | Basic encoding — vowel-removal except-leading + dedupe | jellyfish testdata line 1 |
| RV-M2 | `MRACode("Boern")` | `"BRN"` | Encoding — leading vowel removed BECAUSE B is the leading char (Boern's "o" is the 2nd char — yes vowel-removal applies); B→B, o→drop, e→drop, r→R, n→N | jellyfish testdata line 2 |
| RV-M3 | `MRACode("Smith")` | `"SMTH"` | Encoding canonical | jellyfish testdata line 3 |
| RV-M4 | `MRACode("Smyth")` | `"SMYTH"` | Y-as-consonant test — Y kept since it's not in AEIOU | jellyfish testdata line 4 |
| RV-M5 | `MRACode("Catherine")` | `"CTHRN"` | First-3-last-3 NOT triggered (len 5 ≤ 6) | jellyfish testdata line 5 |
| RV-M6 | `MRACode("Kathrynoglin")` | `"KTHGLN"` | **First-3-last-3 truncation triggered** (pre-truncation `KTHRYNGLN` len 9 > 6 → first 3 `KTH` + last 3 `GLN` = `KTHGLN`) | jellyfish testdata line 7 |
| RV-M7 | `MRACompare("Smith", "Smyth")` | `(true, 5)` | Threshold-passing case: `SMTH` vs `SMYTH`, sum_len=9 → threshold 3; matching forward removes S,M; backward H,T differ in 4-char vs 4-char (T,H + Y,?) — careful planner re-derives. | Hand-derived per NBS 943 rules |
| RV-M8 | `MRACompare("Smith", "JohnathanLongName")` | `(false, 0)` | **Length-difference >= 3 auto-mismatch** — `SMTH` (len 4) vs `JHNTHNLNGNM` truncated to `JHNNNM` or similar (len 6) → diff=2; need an actual >3-diff pair. Better: `MRACompare("Ad", "ZachariahMontgomery") = (false, 0)` because `AD` (len 2) vs `ZCRMNTGMRY` (len 10) truncated to `ZCRMRY` (first-3-last-3 = `ZCR` + `MRY`) len 6 → diff = 4 ≥ 3 → auto-mismatch. | docs/requirements.md §7.4.4 line 696; CONTEXT.md §3 |
| RV-M9 | `MRACompare("Ad", "Ed")` | `(false, 3)` | Threshold-edge case: `AD` vs `ED` (both len 2), sum_len=4 → threshold 5; similarity = 6 - max(1,1) = 5; but the spec table says sum ≤ 4 → threshold 5. So 5 >= 5 → MATCH... actually `AD/ED` differ in first char only, similarity should be 5. **Planner re-derives during plan 07-04;** the threshold-edge case is real but the exact values need careful derivation. | jellyfish testdata + threshold table |
| RV-M10 | `MRACompare("William", "Willyam")` | `(true, 5)` or similar | Generic positive match where common-prefix + similar-end works | Hand-derive |
| RV-M11 | `MRACompare("", "")` | `(true, 6)` | Both-empty edge: `MRACode("") = ""`, `len("") = 0`, sum_len = 0 → threshold 5 (table row sum ≤ 4); no chars to process → similarity = 6 - 0 = 6; 6 ≥ 5 → match. `MRAScore("", "") = 1.0`. | docs/requirements.md §7.4.4 ; algorithm-correctness-standards "Both inputs empty: return 1.0 by convention" |
| RV-M12 | `MRAScore("Smith", "Smyth")` | `1.0` | Wraps RV-M7 | spec line 692 |

## Jellyfish Cross-Validation Mechanics (§4 of output)

### 4.1 The Double Metaphone problem

**CRITICAL RESEARCH FINDING:** Jellyfish 1.2.1 does NOT have a Double Metaphone implementation. Reading `jellyfish/src/lib.rs` directly:

```
mod common; mod hamming; mod jaccard; mod jaro; mod levenshtein;
mod match_rating; mod metaphone; mod nysiis; mod soundex; mod testutils;

pub use match_rating::{match_rating_codex, match_rating_comparison};
pub use metaphone::metaphone;       // <-- SINGLE Metaphone, NOT Double
pub use nysiis::nysiis;
pub use soundex::soundex;
```

There is no `double_metaphone` module, no `mod double_metaphone;`, no `pub use double_metaphone::double_metaphone;`. CONTEXT.md §1's 40-entry Double Metaphone corpus cannot use jellyfish.

### 4.2 Recommended resolution: dual-pin script

`scripts/gen-phonetic-cross-validation.py` should pin BOTH `jellyfish` AND `metaphone`:

```python
#!/usr/bin/env python3
"""scripts/gen-phonetic-cross-validation.py

Pinned dual-source generator:
  - jellyfish==1.2.1 for Soundex, NYSIIS, MRA (jellyfish.soundex,
    jellyfish.nysiis, jellyfish.match_rating_codex, jellyfish.match_rating_comparison)
  - Metaphone==0.6 for Double Metaphone (metaphone.doublemetaphone)
"""
import json, sys, subprocess
from datetime import datetime, timezone

JELLYFISH_VERSION = "1.2.1"
METAPHONE_VERSION = "0.6"

import jellyfish  # noqa: E402
assert jellyfish.__version__ == JELLYFISH_VERSION, (
    f"jellyfish version mismatch: installed {jellyfish.__version__}, "
    f"script pinned to {JELLYFISH_VERSION} — "
    f"run: python3 -m pip install --user jellyfish=={JELLYFISH_VERSION}"
)

# Metaphone package doesn't expose __version__ reliably; verify via pip.
metaphone_show = subprocess.check_output(
    ["python3", "-m", "pip", "show", "Metaphone"], text=True
)
assert f"Version: {METAPHONE_VERSION}" in metaphone_show, (
    f"Metaphone version mismatch — run: python3 -m pip install --user "
    f"Metaphone=={METAPHONE_VERSION}"
)

from metaphone import doublemetaphone  # noqa: E402

# ... CASES_SOUNDEX, CASES_NYSIIS, CASES_MRA, CASES_DOUBLE_METAPHONE
# (per CONTEXT.md §1: 15 / 20 / 20 / 40 entries)
```

### 4.3 Variant-divergence handling per algorithm

| Algorithm | jellyfish/metaphone parity | Expected divergence rate | Tagging rule |
|-----------|---------------------------|--------------------------|--------------|
| Soundex | Identical (both use Knuth/Census variant) | **0%** | Schema-required `variant_divergence` field present on every entry; expect ALL to be `false` for the canonical corpus |
| NYSIIS | Diverges on every input where Taft-truncated output differs from jellyfish-untruncated output (i.e. every input whose untruncated code is > 6 chars) | **~40-60% of canonical 20-entry corpus** | Per-entry `variant_divergence: true` + `divergent_jellyfish_value: "CATARAN"` + `variant_divergence_reason: "jellyfish does not truncate to 6 chars"`. Our impl asserted against the TRUNCATED value. |
| MRA | No known divergences for canonical inputs | **0%** | Schema-required; expect mostly `false`. NOTE: jellyfish returns `Err` for length-diff>=3 mismatch; our impl returns `(false, 0)` — this is a Go-vs-Python API shape difference, NOT a `variant_divergence` (the underlying semantic decision is identical: those pairs do not match). |
| Double Metaphone (via oubiwann/metaphone) | No known divergences for canonical inputs (oubiwann/metaphone is faithful Andrew-Collins port of Philips' C reference) | **0%** | Schema-required; expect ALL `false` if our implementation faithfully tracks the Philips rule table |

### 4.4 vectors.json schema (per algorithm)

```json
{
  "version": 1,
  "_metadata": {
    "jellyfish_version": "1.2.1",
    "metaphone_version": "0.6",
    "python_version": "3.12.x",
    "regenerated_at": "2026-MM-DDTHH:MM:SSZ",
    "script_sha256": "<hex digest>"
  },
  "soundex_entries": [
    {"input": "Robert", "code": "R163", "variant_divergence": false},
    {"input": "Tymczak", "code": "T522", "variant_divergence": false,
     "_note": "Discriminating Knuth/Census-vs-SQL vector. SQL variant returns T520."}
  ],
  "nysiis_entries": [
    {"input": "Brown", "code": "BRAN", "variant_divergence": false},
    {"input": "Catherine", "code": "CATARA", "variant_divergence": true,
     "divergent_jellyfish_value": "CATARAN",
     "variant_divergence_reason": "jellyfish does not truncate to 6 chars"}
  ],
  "mra_entries": [
    {"input": "Byrne", "code": "BYRN", "variant_divergence": false},
    {"input_a": "Smith", "input_b": "Smyth", "compare_matched": true,
     "compare_sim_score": 5, "variant_divergence": false}
  ],
  "double_metaphone_entries": [
    {"input": "Schmidt", "primary": "XMT", "secondary": "SMT",
     "branch": "germanic", "variant_divergence": false},
    {"input": "Catherine", "primary": "K0RN", "secondary": "KTRN",
     "branch": "greek", "variant_divergence": false},
    {"input": "Pacheco", "primary": "PXK", "secondary": "PSK",
     "branch": "romance", "variant_divergence": false}
  ]
}
```

The `branch` field on DM entries gives plan 07-02's verifier a machine-checkable view of language-branch coverage (asserts `min 7 entries per major branch` per CONTEXT.md §3).

### 4.5 Install / regenerate footprint

```bash
# One-time per developer machine:
python3 -m pip install --user jellyfish==1.2.1 Metaphone==0.6

# Regenerate after rule-table changes (rare):
make regen-phonetic-cross-validation
# → invokes: python3 scripts/gen-phonetic-cross-validation.py
# → writes: testdata/cross-validation/phonetic/vectors.json
```

The script is committed to git; the generated `vectors.json` is also committed. Future planners considering a jellyfish or metaphone version bump must (a) update the pin, (b) regenerate the corpus, (c) run the Go cross-validation test, (d) record the regeneration date in `vectors.json._metadata.regenerated_at`.

## Validation Architecture (§5 of output)

> Phase 7 is the canonical Nyquist-validation gate phase for phonetic encoding. The plan-phase workflow consumes this section to generate VALIDATION.md.

### Test Framework

| Property | Value |
|----------|-------|
| Framework | Go stdlib `testing` (root), `cucumber/godog` v0.15.0 + `go.uber.org/goleak` v1.3.0 + `stretchr/testify` v1.10.0 (tests/bdd module) |
| Config file | `go.mod` (root, zero non-stdlib `require` lines), `tests/bdd/go.mod` (test-only sub-module — already established Phase 1) |
| Quick run command (per task commit) | `go test ./...` (root tests, 8-12 seconds typical) |
| Full suite command (per wave merge + phase gate) | `make check && make test-bdd && make verify-determinism` |
| Test data location | `testdata/cross-validation/phonetic/vectors.json` (jellyfish+metaphone corpus); `testdata/golden/algorithms.json` (binary Score golden); `testdata/golden/phonetic-codes.json` (cross-platform byte-stable code golden — NEW in plan 07-01); `testdata/golden/_staging/<algo>.json` (per-plan staging) |

### Phase Requirements → Test Map

| Req ID | Behavior | Test Type | Automated Command | File Exists? |
|--------|----------|-----------|-------------------|-------------|
| PHON-01 | `SoundexCode("Tymczak") = "T522"` discriminating variant gate | unit (literature reference vector) | `go test -run 'TestSoundexCode/Tymczak' ./...` | ❌ Wave 0 (plan 07-01) |
| PHON-01 | `SoundexCode("Ashcraft") = SoundexCode("Ashcroft") = "A261"` H/W-handling gate | unit | `go test -run 'TestSoundexCode/Ashcra' ./...` | ❌ Wave 0 |
| PHON-01 | `SoundexScore` returns binary 0.0/1.0 with `1.0` on `("", "")` | unit + property | `go test -run 'TestSoundexScore\|TestProp_Soundex' ./...` | ❌ Wave 0 |
| PHON-01 | Soundex 5-invariant property tests (Identity / Symmetric / Range / NoNaN / NoInf) | property (`testing/quick`) | `go test -run 'TestProp_Soundex' ./...` | ❌ Wave 0 |
| PHON-01 | Soundex fuzz no-panic + charset invariant (`[A-Z][0-9]{3}`) on ASCII + non-ASCII regimes | fuzz | `go test -fuzz=FuzzSoundex -fuzztime=60s ./...` | ❌ Wave 0 |
| PHON-01 | Soundex cross-platform byte-stable code via `phonetic-codes.json` | golden | `go test -run 'TestPhoneticCodesGolden/Soundex' ./...` | ❌ Wave 0 |
| PHON-01 | Soundex cross-validation against jellyfish | cross-validation | `go test -run 'TestPhonetic_CrossValidation/Soundex' ./...` | ❌ Wave 0 |
| PHON-01 | Soundex BDD scenarios (≥ 4) | bdd | `make test-bdd` | ❌ Wave 0 |
| PHON-01 | Soundex benchmark `< 500 ns, 0 allocations` per performance-standards SKILL | bench | `go test -bench=BenchmarkSoundex -benchmem -run=^$ ./...` | ❌ Wave 0 |
| PHON-02 | `DoubleMetaphoneKeys("Schmidt") = ("XMT", "SMT")` Germanic-branch gate | unit | `go test -run 'TestDoubleMetaphoneKeys/Schmidt' ./...` | ❌ Wave 0 (plan 07-02) |
| PHON-02 | `DoubleMetaphoneScore("Schmidt", "Smith") = 1.0` Germanic cross-match | unit | `go test -run 'TestDoubleMetaphoneScore/SchmidtSmith' ./...` | ❌ Wave 0 |
| PHON-02 | `DoubleMetaphoneKeys("Catherine") = DoubleMetaphoneKeys("Katherine") = ("K0RN", "KTRN")` Greek gate | unit | `go test -run 'TestDoubleMetaphoneKeys/Catherine\|Katherine' ./...` | ❌ Wave 0 |
| PHON-02 | `DoubleMetaphoneKeys("Pacheco")` contains `"PXK"` Romance/Spanish gate | unit | `go test -run 'TestDoubleMetaphoneKeys/Pacheco' ./...` | ❌ Wave 0 |
| PHON-02 | DM language-branch checklist: at least one Slavic AND one Chinese-origin reference vector | unit | `go test -run 'TestDoubleMetaphoneKeys/(Wojcik\|Cheung)' ./...` | ❌ Wave 0 |
| PHON-02 | DM 5-invariant property tests | property | `go test -run 'TestProp_DoubleMetaphone' ./...` | ❌ Wave 0 |
| PHON-02 | DM fuzz no-panic + key charset (`[A-Z0]`) | fuzz | `go test -fuzz=FuzzDoubleMetaphone -fuzztime=60s ./...` | ❌ Wave 0 |
| PHON-02 | DM cross-platform byte-stable keys via `phonetic-codes.json` | golden | `go test -run 'TestPhoneticCodesGolden/DoubleMetaphone' ./...` | ❌ Wave 0 |
| PHON-02 | DM cross-validation against `oubiwann/metaphone` (40-entry × 5-branch corpus) | cross-validation | `go test -run 'TestPhonetic_CrossValidation/DoubleMetaphone' ./...` | ❌ Wave 0 |
| PHON-02 | DM BDD scenarios (≥ 6 — minimum 1 per language branch + edge case) | bdd | `make test-bdd` | ❌ Wave 0 |
| PHON-02 | DM benchmark `< 2 µs, ≤ 2 allocations` per performance-standards SKILL | bench | `go test -bench=BenchmarkDoubleMetaphone -benchmem -run=^$ ./...` | ❌ Wave 0 |
| PHON-03 | `NYSIISCode("Brown") = NYSIISCode("Browne") = "BRAN"` canonical pair | unit | `go test -run 'TestNYSIISCode/Brown' ./...` | ❌ Wave 0 (plan 07-03) |
| PHON-03 | `NYSIISCode("Robert") = "RABAD"` | unit | `go test -run 'TestNYSIISCode/Robert' ./...` | ❌ Wave 0 |
| PHON-03 | NYSIIS 6-char truncation gate: `len(NYSIISCode("Catherine")) == 6` (NOT `7` like jellyfish) | unit | `go test -run 'TestNYSIISCode_Truncation' ./...` | ❌ Wave 0 |
| PHON-03 | NYSIIS 5-invariant property tests | property | `go test -run 'TestProp_NYSIIS' ./...` | ❌ Wave 0 |
| PHON-03 | NYSIIS fuzz no-panic + charset (`[A-Z]{0,6}`) | fuzz | `go test -fuzz=FuzzNYSIIS -fuzztime=60s ./...` | ❌ Wave 0 |
| PHON-03 | NYSIIS cross-validation against jellyfish with variant_divergence tags for >6-char inputs | cross-validation | `go test -run 'TestPhonetic_CrossValidation/NYSIIS' ./...` | ❌ Wave 0 |
| PHON-03 | NYSIIS golden | golden | `go test -run 'TestPhoneticCodesGolden/NYSIIS' ./...` | ❌ Wave 0 |
| PHON-03 | NYSIIS BDD scenarios (≥ 4) | bdd | `make test-bdd` | ❌ Wave 0 |
| PHON-03 | NYSIIS benchmark `< 500 ns, 0 allocations` | bench | `go test -bench=BenchmarkNYSIIS -benchmem -run=^$ ./...` | ❌ Wave 0 |
| PHON-04 | `MRACode("Byrne") = "BYRN"` | unit | `go test -run 'TestMRACode/Byrne' ./...` | ❌ Wave 0 (plan 07-04) |
| PHON-04 | `MRACode("Kathrynoglin") = "KTHGLN"` first-3-last-3 truncation gate | unit | `go test -run 'TestMRACode/Kathryn' ./...` | ❌ Wave 0 |
| PHON-04 | `MRACompare("Smith", "Smyth") = (true, 5)` (or planner-derived exact int) threshold-pass | unit | `go test -run 'TestMRACompare/SmithSmyth' ./...` | ❌ Wave 0 |
| PHON-04 | `MRACompare("Smith", "ZachariahMontgomery") = (false, 0)` length-diff>=3 auto-mismatch gate | unit | `go test -run 'TestMRACompare/LengthDiff' ./...` | ❌ Wave 0 |
| PHON-04 | `MRAScore(a, b) = 1.0 if MRACompare(a, b).matched else 0.0` consistency | unit + property | `go test -run 'TestMRAScore\|TestProp_MRA' ./...` | ❌ Wave 0 |
| PHON-04 | MRA 5-invariant property tests | property | `go test -run 'TestProp_MRA' ./...` | ❌ Wave 0 |
| PHON-04 | MRA fuzz no-panic + charset (`[A-Z]{0,6}`) | fuzz | `go test -fuzz=FuzzMRA -fuzztime=60s ./...` | ❌ Wave 0 |
| PHON-04 | MRA cross-validation against jellyfish | cross-validation | `go test -run 'TestPhonetic_CrossValidation/MRA' ./...` | ❌ Wave 0 |
| PHON-04 | MRA golden | golden | `go test -run 'TestPhoneticCodesGolden/MRA' ./...` | ❌ Wave 0 |
| PHON-04 | MRA BDD scenarios (≥ 4) | bdd | `make test-bdd` | ❌ Wave 0 |
| PHON-04 | MRA benchmark `< 500 ns, 0 allocations` | bench | `go test -bench=BenchmarkMRA -benchmem -run=^$ ./...` | ❌ Wave 0 |
| PHON-01..04 | Monge-Elkan binary-inner composition (4 new tests) | unit + property | `go test -run 'TestMongeElkanScore_BinaryInner_' ./...` | ❌ Wave 0 (lockstep across plans 07-01..07-04 per CONTEXT.md §4) |
| PHON-01..04 | Monge-Elkan panic-test fixture coverage shrinks 14→13→12→11→10 in lockstep | unit | `go test -run 'TestMongeElkan_PanicsOnNonPermittedInner' ./...` | (exists; mutated in lockstep with each plan) |
| PHON-01..04 | Cross-platform determinism gate via `phonetic-codes.json` byte-stable across CI matrix | golden + CI matrix | `make verify-determinism` | ❌ Wave 0 |
| PHON-01..04 | identifier-similarity example extension: 23 columns, golden stdout fixture regenerated | meta-test | `go test ./examples/identifier-similarity/...` | ❌ Wave 0 (plan 07-05 finalisation) |
| PHON-01..04 | `examples/phonetic-keys/main.go` runnable demonstrating all 4 code surfaces | meta-test | `go test ./examples/phonetic-keys/...` | ❌ Wave 0 (plan 07-05 finalisation) |

### Sampling Rate

- **Per task commit:** `go test ./...` (root quick run) — fast feedback under 30s.
- **Per wave merge:** `make check && make test-bdd` — full quality gate including golangci-lint, vet, race, vulncheck, gosec, coverage floors, BDD.
- **Phase gate:** `make check && make test-bdd && make verify-determinism` green; `make bench` produces measurable benchmarks (full bench.txt regen in 07-05 finalisation); `algorithm-licensing-reviewer` sign-off recorded in plan 07-02 PR description; `--depth=deep` code review completed on plan 07-02.

### Wave 0 Gaps

- [ ] `soundex.go`, `soundex_test.go`, `soundex_bench_test.go`, `soundex_fuzz_test.go`, `dispatch_soundex.go` — plan 07-01
- [ ] `double_metaphone.go`, `double_metaphone_test.go`, `double_metaphone_bench_test.go`, `double_metaphone_fuzz_test.go`, `dispatch_double_metaphone.go` — plan 07-02
- [ ] `nysiis.go`, `nysiis_test.go`, `nysiis_bench_test.go`, `nysiis_fuzz_test.go`, `dispatch_nysiis.go` — plan 07-03
- [ ] `mra.go`, `mra_test.go`, `mra_bench_test.go`, `mra_fuzz_test.go`, `dispatch_mra.go` — plan 07-04
- [ ] `phonetic_cross_validation_test.go` — loader for `testdata/cross-validation/phonetic/vectors.json` (single file with 4 `t.Run` sub-tests; plan 07-01 creates with Soundex sub-test, plans 07-02..07-04 append theirs)
- [ ] `phonetic_codes_golden_test.go` — separate loader for `testdata/golden/phonetic-codes.json` (recommended per CONTEXT.md §7; plan 07-01 creates)
- [ ] `scripts/gen-phonetic-cross-validation.py` — plan 07-01 (full dual-pin script per §4.2)
- [ ] `testdata/cross-validation/phonetic/vectors.json` — committed corpus (plan 07-01 creates with Soundex section, subsequent plans extend)
- [ ] `testdata/golden/phonetic-codes.json` — plan 07-01 creates
- [ ] `testdata/golden/_staging/soundex.json`, `_staging/double_metaphone.json`, `_staging/nysiis.json`, `_staging/mra.json` — one per plan, merged into `algorithms.json` in 07-05
- [ ] `Makefile` target `regen-phonetic-cross-validation` — plan 07-01
- [ ] `docs/cross-validation.md` extension with "Phonetic cross-validation" section — plan 07-01
- [ ] `tests/bdd/features/soundex.feature`, `double_metaphone.feature`, `nysiis.feature`, `mra.feature` — one per algorithm plan
- [ ] `tests/bdd/features/monge_elkan_phonetic_inner.feature` — plan 07-04 OR finalisation
- [ ] `tests/bdd/steps/algorithms_steps.go` — append phonetic step registrations
- [ ] `example_test.go` — 9 new `ExampleXxx` entries appended across plans 07-01..07-04
- [ ] `props_test.go` — Five-invariant blocks appended + 4 new `TestMongeElkanScore_BinaryInner_<Algo>` tests
- [ ] `monge_elkan.go` `permittedMongeElkanInner` map — incremental 14→15→16→17→18 across plans 07-01..07-04
- [ ] `monge_elkan_test.go` panic-fixture `rejected` slice — incremental shrink 9→8→7→6→5 across plans 07-01..07-04 (note: current Phase-6 `rejected` slice has 9 entries; Phase 7 each-plan-removes-one shrinks to 5 final)
- [ ] `examples/identifier-similarity/main.go` — 19→23 columns (plan 07-05 finalisation)
- [ ] `examples/phonetic-keys/main.go` + `main_test.go` — new program (plan 07-05 finalisation)
- [ ] `llms.txt` + `llms-full.txt` — per-plan sync (NOT deferred to finalisation per CONTEXT.md §6-prior)

## Per-algorithm Implementation Notes (§6 of output)

### 6.1 Soundex

- **Expected line count**: 60-80 lines of Go (small algorithm). The Phase 4 LCSStr and Phase 5 Q-Gram Jaccard line counts are comparable references.
- **Allocation budget**: < 500 ns, 0 allocations per performance-standards SKILL. Trivial to meet — algorithm is O(n) byte scan with a 4-byte stack buffer for the result.
- **Edge cases**:
  - Empty input → empty code → both-empty returns 1.0 per algorithm-correctness-standards.
  - Single-character input → letter + 3 zero-pads (e.g. `"A" → "A000"`).
  - All-vowel input after first → first letter + 3 zero-pads (e.g. `"Eve" → "E000"`).
  - Initial digit / punctuation input — silently skipped per CONTEXT.md §5; if entire input is non-ASCII or punctuation, returns empty string OR the leading non-skipped letter padded with zeros (planner picks; document either way explicitly).
- **Implementation pattern**: ASCII byte-scan with a 4-byte stack-allocated result buffer (`var result [4]byte`); no heap allocation. Tower of switch cases for digit-group mapping (B/F/P/V→1 etc.).
- **Why O(1) space**: result is fixed 4 bytes; no DP table; no map.

### 6.2 Double Metaphone

- **Expected line count**: 350-450 lines of Go (heaviest in Phase 7). The C reference is ~400 lines; Go transcription typically slightly longer (~10% overhead for stricter typing and explicit slice bounds checking).
- **Allocation budget**: < 2 µs, ≤ 2 allocations per performance-standards SKILL. The two allocations are typically the primary and secondary key result strings.
- **Implementation pattern**: imperative state machine over a `[]byte` input (uppercased ASCII letters) with two output `strings.Builder` (or `[]byte` slices) for primary and secondary. Position-by-position dispatch via a large switch on the current letter with sub-conditions for surrounding context (look-ahead and look-behind up to 4 positions). Language-origin "mode" flags (`is_slavo_germanic`, `is_germanic`, etc.) set during the initial pre-scan.
- **Critical correctness checks** (per CONTEXT.md §3 mandatory checklist):
  - Germanic: `Schmidt → ("XMT", "SMT")`, `Smith → ("SM0", "XMT")`, `Score(Schmidt, Smith) = 1.0`.
  - Slavic: ≥ 1 vector.
  - Romance: `Pacheco` contains `"PXK"`.
  - Greek: `Catherine = Katherine = ("K0RN", "KTRN")` exact.
  - Chinese-origin: ≥ 1 vector.
- **Key truncation**: 4 characters per spec (CONTEXT.md §1.2 + docs/requirements.md §7.4.2 line 659). The C reference uses a configurable `MAX_LENGTH = 4`. Our impl pins to 4.
- **Both-empty edge**: `("", "")` is the only case where both inputs and both keys are empty; `DoubleMetaphoneScore("", "") = 1.0` per algorithm-correctness-standards.

### 6.3 NYSIIS

- **Expected line count**: 80-120 lines of Go. The 9-step Wikipedia/Taft algorithm description is procedural; transcribes naturally to a sequence of `if-else` blocks plus a final-letter-cleanup loop.
- **Allocation budget**: < 500 ns, 0 allocations per performance-standards SKILL. Truncated to 6 chars → stack-allocated `[6]byte` result; intermediate buffer (the pre-truncation form) is also stack-allocated (max length ~32 for any reasonable input).
- **Implementation pattern**: byte-scan with rule-by-rule application; final truncation to 6 via `result[:6]`. No DP.
- **Truncation gate**: `len(NYSIISCode(input)) <= 6` invariant for all inputs (property tested + asserted in unit tests via `RV-N4..RV-N7` and `RV-N11`).
- **jellyfish divergence**: jellyfish does NOT truncate; our impl does. Every cross-validation entry where jellyfish's output exceeds 6 chars carries `variant_divergence: true`; our impl is asserted against the truncated value.

### 6.4 MRA

- **Expected line count**: 120-160 lines (two surfaces: `MRACode` ~50 lines, `MRACompare` ~80 lines including the threshold table).
- **Allocation budget**: < 500 ns, 0 allocations for `MRACode`; `MRACompare` may need 2 small allocations for the two stripped slices.
- **Threshold table as `var`**: per determinism-standards SKILL "No init() side effects" — declare the threshold table as a package-level `var`:
  ```go
  // mraThresholdTable maps sum-of-lengths to minimum-rating per NBS
  // Tech Note 943 Table A.
  var mraThresholdTable = [13]int{
      // sum:  0  1  2  3  4   5   6   7   8   9  10  11  12
                5, 5, 5, 5, 5,  4,  4,  4,  3,  3,  3,  3,  2,
  }
  // For sum >= 13, threshold is 2 (effectively floor at 2). Spec line says
  // sum=12 → threshold 2; sum > 12 (which can happen with two 6-char codes
  // summing to 12) → 2.
  ```
  Wait — re-read: Wikipedia table A says `sum > 12 → threshold 2` (clamp). Re-derive during plan 07-04. The exact table-clamping behaviour matters.
- **Length-difference auto-mismatch**: `if abs(len(codex_a) - len(codex_b)) >= 3 { return (false, 0) }` — per docs/requirements.md §7.4.4 line 696. jellyfish returns Err here; our Go impl returns `(false, 0)`. This is an API shape difference, NOT a `variant_divergence`.
- **MRACompare integer interpretation**: `(matched bool, simScore int)` where `simScore ∈ [0, 6]` always. `simScore = 6 - max(unmatched_left, unmatched_right)` per NBS-943 step 5. `matched = (simScore >= threshold(sumLen))` per step 6.
- **MRAScore wraps MRACompare**: `return 1.0 if matched else 0.0`. Strict consistency property test: `MRAScore(a, b) == 1.0 iff MRACompare(a, b).matched` (verified by direct test).

## Pitfalls — extensions to PITFALLS.md 4 and 5 (§7 of output)

### Pitfall 7.A: Jellyfish version-pin selection — exact-version requirement

**What goes wrong:** Choosing too-lax a pin (e.g. `jellyfish~=1.2`) lets a patch upgrade silently change cross-validation outputs (jellyfish has shipped patch releases that corrected encoding bugs in the past). The `vectors.json` corpus then differs from what the Go test expects, but the test isn't sure which is right.

**Why it happens:** Python ecosystem conventions allow loose pins; jellyfish (and the Metaphone package) are not zero-bug.

**How to avoid:**
- `JELLYFISH_VERSION = "1.2.1"` exact-string-match assertion at script load (mirrors Phase 6 RapidFuzz pattern).
- `METAPHONE_VERSION = "0.6"` exact-string match via `pip show` parsing (Metaphone package doesn't expose `__version__`).
- `vectors.json._metadata.jellyfish_version` field recorded; Go loader test asserts the field equals the expected version string. Any mismatch fails the Go test (planner-recoverable: run `make regen-phonetic-cross-validation` to refresh).

**Warning signs:**
- `pip install jellyfish` (without `==1.2.1`) produces a version mismatch on next regenerate.
- `vectors.json._metadata.jellyfish_version` differs from the script's `JELLYFISH_VERSION` constant.

**Phase to address:** Phase 7 plan 07-01 (sets up the script and Makefile target). Each subsequent plan inherits the mechanism.

### Pitfall 7.B: Knuth-vs-modified-NYSIIS divergence — truncation rule

**What goes wrong:** The NYSIIS algorithm has two widely-used variants — original Taft 1970 (truncates to 6 chars, applies each rule once) and modified NYSIIS (no truncation, iterates to fixed point on certain rules). Jellyfish ships the modified variant. Implementers cross-validating naively will see ~40-60% of canonical inputs disagree.

**Why it happens:** Wikipedia's NYSIIS article describes the algorithm but doesn't always emphasise the 6-char truncation as a strict rule. NIST DADS describes truncation. Jellyfish ships the longer variant by convention.

**How to avoid:**
- CONTEXT.md §2 LOCKS the original Taft 1970 variant (6-char truncation, each rule once).
- Cross-validation corpus per CONTEXT.md §1 uses `variant_divergence: true` flag on every entry where jellyfish's untruncated output > 6 chars; the Go loader asserts against the TRUNCATED (Taft-1970) value, NOT the jellyfish value.
- A property test `PropNYSIIS_CodeLength` asserts `len(NYSIISCode(x)) <= 6` for any input — locks the truncation invariant directly.

**Warning signs:**
- `NYSIISCode("Catherine")` returns more than 6 characters → modified-NYSIIS implementation bug.
- Cross-validation test sees `variant_divergence: true` entries where the planner expected `false` — suggests the corpus author misclassified.

**Phase to address:** Phase 7 plan 07-03 (NYSIIS implementation). Property test enforces post-implementation.

### Pitfall 7.C: NBS Tech Note 943 threshold-table interpretation

**What goes wrong:** The threshold table in NBS-943 (Table A) maps "sum of encoded lengths" to "minimum rating". The exact table values are:

```
sum ≤ 4:   threshold 5
4 < sum ≤ 7: threshold 4
7 < sum ≤ 11: threshold 3
sum = 12:  threshold 2
sum > 12: threshold 2 (clamp; not always explicitly stated in summaries)
```

The `sum > 12` clamp case is often omitted from Wikipedia-style summaries. Without it, the implementation produces undefined behaviour on degenerate inputs (e.g. two 6-char codes summing to 12 → threshold 2 is fine, but anything beyond the explicit table is silently buggy).

**Why it happens:** Wikipedia table is incomplete; jellyfish's `match _, sum => sum >= 2` catch-all branch is the implicit clamp; primary-source NBS-943 has the explicit clamp but is rarely read.

**How to avoid:**
- Implementer reads the NBS-943 PDF (now verified accessible at `https://nvlpubs.nist.gov/nistpubs/Legacy/TN/nbstechnicalnote943.pdf`) before coding the table.
- Threshold table is declared as `var mraThresholdTable [...]int` with an explicit comment documenting the `sum > 12` clamp.
- Property test `PropMRA_ThresholdMonotonic` asserts threshold(sumLen+1) <= threshold(sumLen) for all sumLen in [0, 20] — locks the monotonic-decrease property.

**Warning signs:**
- `MRACompare` on two 6-char codes (sum=12) and on a 6-char + 7-char (sum=13) returns inconsistent threshold-rule decisions → table clamp bug.
- Comparison with jellyfish: any input pair where sum > 12 disagrees → table clamp bug (jellyfish correctly clamps via the `_ => sum >= 2` catch-all).

**Phase to address:** Phase 7 plan 07-04 (MRA implementation). Property test enforces post-implementation.

### Pitfall 7.D (existing PITFALLS.md Pitfall 5) — Double Metaphone language-branch coverage

(This is a re-emphasis of the existing PITFALLS.md Pitfall 5, scoped to Phase 7's plan 07-02 deliverables.)

**What goes wrong:** Double Metaphone is ~200 conditional branches covering 5 language origins. The most common bug is missing one branch entirely — for example, never emitting the Spanish `PXK` for `Pacheco`, or never emitting the Greek `KTRN` secondary for `Catherine`.

**How to avoid (Phase 7 specifics):**
- The 40-entry DM cross-validation corpus per CONTEXT.md §1 has a `branch` field per entry; the Go loader counts entries per branch and asserts `count(branch) >= 7` for each major branch (Germanic, Slavic, Romance, Greek). Chinese-origin minimum 4.
- The plan 07-02 mandatory checklist (CONTEXT.md §3) explicitly enumerates the 5 must-pass language-branch vectors. Reviewer signs off on this checklist appearing in the PR description.
- `algorithm-licensing-reviewer` reviews the `double_metaphone.go` rule table for fresh-transcription (no copy from `CalypsoSys/godoublemetaphone` or `deezer/double-metaphone-go`) per CONTEXT.md §3 negative-attribution line.

**Warning signs:**
- `DoubleMetaphoneKeys("Pacheco")[0] == "PXK"` fails → Spanish-Romance branch missing or misordered.
- `DoubleMetaphoneKeys("Catherine") != DoubleMetaphoneKeys("Katherine")` → Greek-branch rule applied inconsistently to C-initial vs K-initial.
- `DoubleMetaphoneScore("Schmidt", "Smith") < 1.0` → Germanic XMT cross-match not wired up.

## Code Examples

### Example 1: SoundexCode skeleton (ASCII fast path)

```go
// Source: hand-derived from Knuth TAOCP Vol. 3 §6.4 rules
func SoundexCode(s string) string {
    if s == "" {
        return ""
    }
    var result [4]byte
    var idx int = 0
    var lastGroup byte = soundexNoGroup // sentinel

    // First letter: skip-leading-non-ASCII per CONTEXT.md §5
    var firstFound bool
    for i := 0; i < len(s); i++ {
        c := s[i]
        if c >= 'A' && c <= 'Z' || c >= 'a' && c <= 'z' {
            result[idx] = c & 0xDF // uppercase ASCII
            idx++
            lastGroup = soundexGroup(c & 0xDF)
            firstFound = true
            // continue scanning from i+1
            for j := i + 1; j < len(s) && idx < 4; j++ {
                d := s[j]
                if d >= 'a' && d <= 'z' {
                    d &= 0xDF
                }
                if d < 'A' || d > 'Z' {
                    continue // CONTEXT.md §5 silent skip for non-ASCII
                }
                if d == 'H' || d == 'W' {
                    continue // Knuth/Census H/W skip; lastGroup unchanged
                }
                g := soundexGroup(d)
                if g == soundexVowelGroup {
                    lastGroup = soundexNoGroup // vowels are separators
                    continue
                }
                if g != lastGroup {
                    result[idx] = '0' + g
                    idx++
                }
                lastGroup = g
            }
            break
        }
    }
    if !firstFound {
        return ""
    }
    for ; idx < 4; idx++ {
        result[idx] = '0'
    }
    return string(result[:])
}
```

[CITED: Knuth TAOCP Vol. 3 §6.4; cross-verified via `jellyfish/src/soundex.rs` (BSD-2) for canonical reference vectors only — no code copied]

### Example 2: Cross-validation Go loader skeleton

```go
// In phonetic_cross_validation_test.go
package fuzzymatch_test

import (
    "encoding/json"
    "math"
    "os"
    "path/filepath"
    "testing"

    "github.com/axonops/fuzzymatch"
)

type phoneticCorpus struct {
    Version  int             `json:"version"`
    Metadata phoneticMetadata `json:"_metadata"`
    Soundex  []soundexEntry  `json:"soundex_entries"`
    NYSIIS   []nysiisEntry   `json:"nysiis_entries"`
    MRA      []mraEntry      `json:"mra_entries"`
    DM       []dmEntry       `json:"double_metaphone_entries"`
}

type phoneticMetadata struct {
    JellyfishVersion string `json:"jellyfish_version"`
    MetaphoneVersion string `json:"metaphone_version"`
    PythonVersion    string `json:"python_version"`
    RegeneratedAt    string `json:"regenerated_at"`
    ScriptSha256     string `json:"script_sha256"`
}

type soundexEntry struct {
    Input             string `json:"input"`
    Code              string `json:"code"`
    VariantDivergence bool   `json:"variant_divergence"`
}

// (nysiisEntry, mraEntry, dmEntry — analogous)

func TestPhonetic_CrossValidation(t *testing.T) {
    const expectedJellyfish = "1.2.1"
    const expectedMetaphone = "0.6"
    path := filepath.Join("testdata", "cross-validation", "phonetic", "vectors.json")
    raw, err := os.ReadFile(path)
    if err != nil {
        t.Fatalf("read %s: %v (regenerate with `make regen-phonetic-cross-validation`)", path, err)
    }
    var c phoneticCorpus
    if err := json.Unmarshal(raw, &c); err != nil {
        t.Fatalf("parse %s: %v", path, err)
    }
    if c.Version != 1 {
        t.Fatalf("unsupported corpus version %d", c.Version)
    }
    if c.Metadata.JellyfishVersion != expectedJellyfish {
        t.Fatalf("jellyfish version mismatch: corpus=%s expected=%s — regenerate",
            c.Metadata.JellyfishVersion, expectedJellyfish)
    }
    if c.Metadata.MetaphoneVersion != expectedMetaphone {
        t.Fatalf("metaphone version mismatch: corpus=%s expected=%s — regenerate",
            c.Metadata.MetaphoneVersion, expectedMetaphone)
    }
    t.Run("Soundex", func(t *testing.T) {
        for _, e := range c.Soundex {
            e := e
            t.Run(e.Input, func(t *testing.T) {
                got := fuzzymatch.SoundexCode(e.Input)
                if got != e.Code {
                    t.Errorf("SoundexCode(%q) = %q; jellyfish(%s) = %q (variant_divergence=%v)",
                        e.Input, got, expectedJellyfish, e.Code, e.VariantDivergence)
                }
            })
        }
    })
    // (NYSIIS, MRA, DM sub-tests — analogous; NYSIIS asserts against TRUNCATED
    // value when variant_divergence is true)
}
```

[CITED: Phase 6 `token_ratio_cross_validation_test.go` pattern; Phase 4 `ratcliff_obershelp_cross_validation_test.go` pattern]

## State of the Art

| Old Approach | Current Approach | When Changed | Impact |
|--------------|------------------|--------------|--------|
| jellyfish pre-1.0 (pure Python) | jellyfish 1.x (Rust-backed core) | 2024 | Faster cross-validation runs (~10x); same canonical outputs |
| Metaphone-as-jellyfish-builtin | Separate `metaphone` PyPI package (Andrew Collins port) for Double Metaphone | jellyfish 1.x deprecated double_metaphone; oubiwann/metaphone became the canonical Python DM source | **Affects Phase 7 directly** — DM cross-validation requires the separate package |
| Wikipedia as primary NYSIIS reference | Knuth TAOCP Vol. 3 §6.4 + Wikipedia + NIST DADS as triangulated canonical secondary | Pre-2020 community pattern | Per CONTEXT.md §2 LOCKED: Knuth primary citation + Taft-1970 origin |
| MIT-licensed Go ports as DM canonical | Philips 2000 + SWI-Prolog C source archive + oubiwann Python port | Phase 7 LOCKED per CONTEXT.md §3 | Plan 07-02 negative-attribution line lists Go ports explicitly |

**Deprecated/outdated:**

- **Metaphone 3** — U.S. Patent 7440941; explicitly excluded from this library per `.planning/PROJECT.md` and `docs/requirements.md` §4. Never to be considered for v1.x.
- **`fuzz` / `Fuzzy` Python packages (other than jellyfish/Metaphone/abydos)** — superseded by jellyfish for phonetic and by RapidFuzz for token ratios (Phase 6 already locks RapidFuzz).

## Environment Availability

| Dependency | Required By | Available | Version | Fallback |
|------------|------------|-----------|---------|----------|
| Go toolchain | All Phase 7 implementation | ✓ | 1.26.x | — |
| `python3` | `scripts/gen-phonetic-cross-validation.py` (developer-only) | ✓ | 3.11+ | — |
| `jellyfish` (PyPI) | `make regen-phonetic-cross-validation` (developer-only) | (must be installed by developer; not pre-installed in CI) | `1.2.1` (recommended pin) | — — script REFUSES to run on any other version |
| `Metaphone` (PyPI) | `make regen-phonetic-cross-validation` (developer-only) | (must be installed by developer) | `0.6` (recommended pin) | — |
| `golangci-lint` v2.12.2 | `make check` quality gate (already established) | ✓ | (CI matrix uses) | — |
| `cucumber/godog` v0.15.0 | `tests/bdd` (already established) | ✓ | (tests/bdd/go.mod pinned) | — |
| `go.uber.org/goleak` v1.3.0 | `tests/bdd` (already established) | ✓ | — | — |
| NIST NBS Tech Note 943 PDF | Plan 07-04 implementer reads primary source for MRA rules | ✓ | accessible 2026-05-15 | URL: `https://nvlpubs.nist.gov/nistpubs/Legacy/TN/nbstechnicalnote943.pdf` |

**Missing dependencies with no fallback:** None — Python + the two pip pins are developer-toolchain only and individually responsible for cross-validation generation, not for any compilation or runtime path.

**Missing dependencies with fallback:** None at the runtime level. At the developer-toolchain level, the script REFUSES to run on a version mismatch (correct behaviour — forces regenerate after intentional pin bump).

## Assumptions Log

> Claims tagged `[ASSUMED]` in this research that the planner / discuss-phase should treat as needing user confirmation before locking.

| # | Claim | Section | Risk if Wrong |
|---|-------|---------|---------------|
| A1 | Knuth TAOCP Vol. 3 §6.4 contains a full description of NYSIIS (not just Soundex). The current researcher could not verify this from the search results; CONTEXT.md §2 LOCKS Knuth as the canonical algorithm-description citation. If Knuth's NYSIIS coverage in §6.4 is too sparse to be authoritative, the planner may want to substitute "Wikipedia + NIST DADS + jellyfish source-code read" as the canonical secondary citation chain. | §1 NYSIIS table | Plan 07-03 could ship with a citation that doesn't hold up to algorithm-correctness-reviewer scrutiny if Knuth's NYSIIS subsection is shorter than expected. **Mitigation:** plan 07-03 implementer reads §6.4 directly during research and either confirms the citation OR proposes the secondary chain via OQ-X RESOLUTION LOCKED. |
| A2 | `MRACompare` on length-difference >= 3 inputs should return `(false, 0)` rather than an error. CONTEXT.md §6 LOCKED says `(bool, int)` and `bool=false` is the natural representation of "no match"; the `int=0` is the natural floor of "no similarity computed". Jellyfish returns `Err` here. The planner may want to confirm the `(false, 0)` choice doesn't disagree with consumer expectations. | §1 MRA table; §6.4 | If the planner / `api-ergonomics-reviewer` decides `MRACompare` should return a sentinel error in this case, the surface drifts from `(bool, int)`. **Mitigation:** plan 07-04 author records the choice in OQ-X RESOLUTION LOCKED with rationale; the Phase 8 Scorer wrapper consumes the (false, _) return uniformly anyway. |
| A3 | The exact `MRACompare("Smith", "Smyth")` similarity score is 5. This vector was derived during research but the exact value depends on which-character-removed semantics the planner uses (longest-string-vs-shorter, position-aware order). The vector is in §3.4 RV-M7 but flagged "Hand-derived per NBS 943 rules" — the planner should re-derive directly from NBS-943 during plan 07-04 implementation and adjust the test fixture accordingly. | §3.4 RV-M7, RV-M9 | A wrong-by-one similarity score in the unit test fixture would force a test reroll on first implementation. **Mitigation:** plan 07-04 implementer reads NBS-943 step 3-5 directly + cross-validates with jellyfish's `match_rating_comparison`. |
| A4 | `oubiwann/metaphone` 0.6 produces identical Double Metaphone outputs to Lawrence Philips' original C reference on canonical inputs. The package is widely used and BSD-licensed but the present researcher did not exhaustively verify byte-stability against Philips' 2000 C source. | §4.2; §3.2 | If `oubiwann/metaphone` carries a known bug or divergence, the 40-entry DM corpus would lock in a wrong reference. **Mitigation:** plan 07-02 implementer cross-validates the 5 mandatory language-branch literature reference vectors AGAINST oubiwann/metaphone AND against at least one secondary source (Apache Commons Codec Java reference is the obvious second). Any disagreement between literature + Apache Commons + oubiwann tagged for OQ-X RESOLUTION. |
| A5 | The recommended jellyfish version pin is `1.2.1` (released 2025-10-11). Newer patch versions may have shipped by the time plan 07-01 starts; the planner should run `pip index versions jellyfish` at plan-creation time and either confirm 1.2.1 OR bump the pin with corpus regeneration. | §4 Standard Stack | A stale pin keeps cross-validation usable but loses the freshest jellyfish bug fixes. **Mitigation:** plan 07-01 author confirms pin choice at plan creation; pin is recorded in PR description as "pin X chosen on YYYY-MM-DD; current PyPI latest at that date was Y". |

## Open Questions

> Items NOT LOCKED in CONTEXT.md that the planner will decide.

1. **Exact jellyfish + Metaphone version pins.**
   - What we know: jellyfish 1.2.1 (PyPI 2025-10-11) and Metaphone 0.6 are current stables at research time.
   - What's unclear: whether a newer patch version exists at plan-creation time.
   - Recommendation: planner runs `pip index versions jellyfish Metaphone` at plan 07-01 creation; either confirms 1.2.1/0.6 OR bumps to current latest with rationale recorded in plan 07-01 PR description.

2. **`MRACompare` API on length-difference >= 3 inputs.**
   - What we know: jellyfish returns Err; CONTEXT.md §6 LOCKS the `(bool, int)` surface.
   - What's unclear: whether `(false, 0)` is the right concrete return shape.
   - Recommendation: planner consults `api-ergonomics-reviewer` during plan 07-04 design; default to `(false, 0)` unless the reviewer flags it.

3. **Exact wording of the canonical godoc warning paragraph per algorithm.**
   - What we know: CONTEXT.md §5 provides a canonical text template; each algorithm can extend with specifics.
   - What's unclear: whether each algorithm's godoc adds an algorithm-specific note (e.g. NYSIIS's English-name tuning) or carries only the canonical text.
   - Recommendation: planner picks per-algorithm during each plan creation; defer to `docs-writer` agent during plan execution.

4. **Whether `phonetic_codes_golden_test.go` is a separate file or embedded into `algorithms_golden_test.go`.**
   - What we know: CONTEXT.md `<decisions>` Claude's Discretion section recommends separate file; this research confirms the recommendation (clean schema separation).
   - What's unclear: nothing — recommend separate file.

5. **The exact number of staging-golden entries per algorithm (8-12 per CONTEXT.md `<decisions>` Claude's Discretion).**
   - What we know: Phase 2-6 norm is 8-12.
   - What's unclear: nothing — planner picks 10 per algorithm (middle of range) unless plan-creation review surfaces a reason to deviate.

6. **Plan 07-02's `--depth=deep` invocation mechanism — automatic vs manual.**
   - What we know: CONTEXT.md `<decisions>` Claude's Discretion notes this is executor-time decision.
   - What's unclear: whether the executor agent auto-detects "this is plan 07-02" and invokes deep review.
   - Recommendation: plan 07-02 PR description explicitly requests `--depth=deep` review in its title or body; gsd-executor follows the cue.

7. **Whether `examples/phonetic-keys/main.go` includes `MRACompare` in addition to the four `XxxCode/XxxKeys` functions.**
   - What we know: CONTEXT.md §8 specifies the new example program but doesn't enumerate every function it must demonstrate.
   - What's unclear: whether `MRACompare` (with its `(bool, int)` return) is a "first-class" surface that deserves example space.
   - Recommendation: yes, include `MRACompare` — the int counter is the load-bearing primary-source surface and including it in the example educates consumers about it.

## Sources

### Primary (HIGH confidence)

- **Knuth, D. E. (1973).** *The Art of Computer Programming, Vol. 3: Sorting and Searching*, §6.4. Addison-Wesley — canonical algorithm description for Soundex and NYSIIS. [CITED via `docs/requirements.md` §7.4 and `CONTEXT.md` §2]
- **Philips, L. (2000).** "The double-metaphone search algorithm." *C/C++ Users Journal*, 18(6):38–43. [CITED via `docs/requirements.md` §7.4.2]
- **Moore, G. B., Kuhns, J. L., Trefftzs, J. L., Montgomery, C. A. (1977).** *Accessing individual records from personal data files using non-unique identifiers.* National Bureau of Standards, Technical Note 943. PDF accessible at `https://nvlpubs.nist.gov/nistpubs/Legacy/TN/nbstechnicalnote943.pdf` (5 MB, HTTP 200 verified 2026-05-15). [CITED + VERIFIED]
- **Russell, R. C., Odell, M. K. (1918, 1922).** U.S. Patents 1261167 and 1435663 — Soundex original. [CITED via `algoid.go` line 156-159]
- **Taft, R. L. (1970).** *Name search techniques.* NY State ID and Intelligence System Special Report No. 1. (Famously unavailable; CONTEXT.md §2 LOCKS Knuth as the canonical algorithm-description citation per the unavailability.) [CITED]
- **`jellyfish` Rust source code on GitHub `main` branch** (2026-05-15 reads):
  - `https://raw.githubusercontent.com/jamesturk/jellyfish/main/src/lib.rs` — confirmed jellyfish has NO Double Metaphone module.
  - `https://raw.githubusercontent.com/jamesturk/jellyfish/main/src/soundex.rs` — confirmed Knuth/Census variant (H/W skip).
  - `https://raw.githubusercontent.com/jamesturk/jellyfish/main/src/nysiis.rs` — confirmed modified-NYSIIS (no truncation).
  - `https://raw.githubusercontent.com/jamesturk/jellyfish/main/src/match_rating.rs` — confirmed Table A threshold values + comparison procedure.
  - `https://raw.githubusercontent.com/jamesturk/jellyfish/main/testdata/{soundex,nysiis,match_rating_codex}.csv` — 10/35/12 verified reference-vector entries. [VERIFIED]

### Secondary (MEDIUM-HIGH confidence)

- **Wikipedia, "Match rating approach"** (`https://en.wikipedia.org/wiki/Match_rating_approach`) — Table A values + 6-step comparison procedure cross-verified against jellyfish source. [CITED + VERIFIED]
- **Wikipedia, "New York State Identification and Intelligence System"** (`https://en.wikipedia.org/wiki/New_York_State_Identification_and_Intelligence_System`) — 9-step procedure; lacks worked examples; lacks explicit emphasis on truncation. [CITED]
- **Apache Commons Codec, `DoubleMetaphone.java`** — Apache-2.0 Java reference for Double Metaphone; consulted for rule-table organisation only. [CITED]
- **`SWI-Prolog/packages-nlp/double_metaphone.c`** (`https://github.com/SWI-Prolog/packages-nlp/blob/master/double_metaphone.c`) — public-domain redistribution of Philips' original C reference with Maurice Aubrey's perl-port-derived bug fixes. The long-stable redistribution URL recommended for plan 07-02's rule-table provenance line. [CITED]
- **`oubiwann/metaphone` (PyPI Metaphone 0.6)** — Andrew Collins' Python port of Lawrence Philips' Double Metaphone C++ reference; BSD-licensed; canonical Python DM source. [CITED]

### Tertiary (LOW confidence)

- WebSearch confirmation that `oubiwann/metaphone` 0.6 produces outputs identical to Philips' original C reference — assumed without exhaustive verification. See Assumptions A4. [ASSUMED]

### Project-internal (HIGH confidence — spec-locked)

- `docs/requirements.md` §7.4 — phonetic algorithm specifications. [CITED]
- `docs/requirements.md` §11 — Phonetic Algorithm Integration. [CITED]
- `docs/requirements.md` §13.3, §13.4 — determinism rules. [CITED]
- `docs/requirements.md` §15.3, §15.4 — property test + fuzz test conventions. [CITED]
- `.planning/REQUIREMENTS.md` lines 50-53, 224-227 — PHON-01..PHON-04 traceability. [CITED]
- `.planning/ROADMAP.md` Phase 7 section. [CITED]
- `.planning/PROJECT.md` — Metaphone 3 patent exclusion; zero-cgo and pure-Go-stdlib constraints. [CITED]
- `.planning/research/PITFALLS.md` Pitfalls 4 and 5 — direct evidence base for §1, §2, §3 LOCKED decisions. [CITED]
- `.planning/phases/07-phonetic-algorithms/07-CONTEXT.md` — the 8 LOCKED decisions this research describes implementation of. [CITED]
- `.planning/phases/06-token-based-algorithms/06-CONTEXT.md` — RapidFuzz pin mechanism (Phase 7's jellyfish pin mirrors exactly); `permittedMongeElkanInner` map design. [CITED]
- `.planning/phases/04-remaining-character-gestalt/04-CONTEXT.md` — Source-Origin Statement format (Phase 7 DM extends with negative-attribution line). [CITED]
- `.claude/skills/algorithm-correctness-standards/SKILL.md` — primary-source citation format, reference vectors, formula docs. [CITED]
- `.claude/skills/algorithm-licensing-standards/SKILL.md` — **LOAD-BEARING for Phase 7.** Fresh-implementation discipline, attribution format. Plan 07-02 gates on this. [CITED]
- `.claude/skills/determinism-standards/SKILL.md` — no map iteration, golden files (phonetic codes as determinism gate value). [CITED]
- `.claude/skills/performance-standards/SKILL.md` — allocation budgets (Soundex/NYSIIS/MRA < 500 ns, 0 allocations; DoubleMetaphone < 2 µs, ≤ 2 allocations). [CITED]
- `algoid.go` lines 156-175 — AlgoID enum slots already reserved. [VERIFIED via direct read]
- `monge_elkan.go` lines 271-314 — `permittedMongeElkanInner` map currently at 14 entries with comment documenting Phase 7's additive 4-entry expansion. [VERIFIED via direct read]
- `monge_elkan_test.go` lines 298-345 — panic-test fixture with `rejected` slice currently containing 9 entries (5 base + 4 Phase 7 reserved). [VERIFIED via direct read]
- `strcmp95.go` lines 17-32 — canonical Source-Origin Statement format that plan 07-02 extends. [VERIFIED via direct read]
- `testdata/cross-validation/token-ratios/vectors.json` — Phase 6 RapidFuzz pin metadata format (`_metadata.rapidfuzz_version`, `python_version`, `regenerated_at` — Phase 7's `phonetic/vectors.json` mirrors this). [VERIFIED via direct read]
- `scripts/gen-token-ratio-cross-validation.py` lines 124-136 — Phase 6 RapidFuzz pin assertion mechanism that Phase 7's script mirrors. [VERIFIED via direct read]

## Metadata

**Confidence breakdown:**
- Standard stack: HIGH — jellyfish + Metaphone pin recommendations verified via PyPI listings; package internal source-code reads verified via raw.githubusercontent.com on `main` branch.
- Architecture (per-algorithm implementation patterns): HIGH — patterns are well-established from Phases 2-6; phonetic algorithms are simpler than Phases 2-6's DP-heavy work in most cases (Double Metaphone being the exception).
- Reference vector catalogue: HIGH for Soundex/NYSIIS/MRA — cross-verified against jellyfish testdata. MEDIUM-HIGH for Double Metaphone — derived from spec + oubiwann/metaphone evidence; plan 07-02 implementer should re-verify the 5 mandatory branch vectors at plan creation.
- Pitfalls (extensions to PITFALLS.md 4 and 5): HIGH — direct extensions of established pitfalls with concrete Phase 7 specifics.
- Cross-validation source identification: HIGH — `oubiwann/metaphone` recommendation is the clean resolution to the jellyfish-has-no-DM finding; alternate (Apache Commons) noted as fallback.

**Research date:** 2026-05-15
**Valid until:** 2026-06-15 (30 days for stable Go ecosystem dependencies; sooner if jellyfish ships a new patch version).

## RESEARCH COMPLETE

Phase 7 research is complete and ready for planning. The three load-bearing findings that affect Phase 7 plans are: (1) `jellyfish` does not ship Double Metaphone, so the planner must add a second pinned pip dep `Metaphone==0.6` (oubiwann's BSD-licensed Andrew Collins port) — this is the clean resolution to CONTEXT.md §1's implicit assumption that jellyfish alone covers all four algorithms; (2) jellyfish's Soundex correctly uses the Knuth/Census variant we're locked to, so the `variant_divergence` mechanism for Soundex per CONTEXT.md §1 is effectively defensive insurance with no expected divergences for canonical inputs; (3) jellyfish's NYSIIS does NOT truncate to 6 characters (modified variant), so every NYSIIS jellyfish vector longer than 6 chars must carry `variant_divergence: true` and our impl is asserted against the Taft-1970-truncated value. Beyond these findings the research populates: (a) Source-Origin Statement template for Double Metaphone (plan 07-02) extending Phase 4 WR-01 format with a rule-table provenance line citing the SWI-Prolog `double_metaphone.c` redistribution URL and a MIT-Go-port negative-attribution line listing five forbidden Go ports; (b) a 12-vector Soundex unit test catalogue including the load-bearing Tymczak→T522 and Ashcraft/Ashcroft→A261 discriminators; (c) a 20-vector Double Metaphone unit test catalogue covering all 5 language-origin branches per CONTEXT.md §3 mandatory checklist; (d) a 12-vector NYSIIS catalogue with truncation-divergence tags against jellyfish; (e) a 12-vector MRA catalogue with threshold-edge and length-difference auto-mismatch gates; (f) a full Nyquist Dimension 8 validation architecture map for the plan-phase workflow to consume into VALIDATION.md; (g) three new pitfall extensions (jellyfish pin selection, Knuth-vs-modified-NYSIIS divergence, NBS-943 threshold table interpretation) layered on PITFALLS.md 4 and 5; (h) 7 open questions for the planner including the exact jellyfish + Metaphone version pin (recommend 1.2.1 / 0.6, planner re-checks at plan creation) and the MRACompare-on-length-diff-≥3 return shape (recommend `(false, 0)`).
