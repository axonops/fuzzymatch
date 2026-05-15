# Phase 7: Phonetic Algorithms - Context

**Gathered:** 2026-05-15
**Status:** Ready for planning

<domain>
## Phase Boundary

Ship the four phonetic algorithms in the catalogue:

- **Soundex** (`AlgoSoundex`, reserved at `algoid.go:156-159`) — Knuth/Census 1880 variant per Knuth TAOCP Vol. 3 §6.4 (NOT the SQL/MySQL variant); `SoundexCode(s) string` + `SoundexScore(a,b) float64`.
- **Double Metaphone** (`AlgoDoubleMetaphone`, reserved at `algoid.go:161-165`) — Philips 2000 + the public-domain C reference; `DoubleMetaphoneKeys(s) (primary, secondary string)` + `DoubleMetaphoneScore(a,b) float64`.
- **NYSIIS** (`AlgoNYSIIS`, reserved at `algoid.go:167-170`) — original Taft 1970 algorithm with 6-character truncation, each rule applied once; primary citation **Knuth TAOCP Vol. 3 §6.4** (Taft 1970 is the algorithmic origin, but the actual canonical algorithm description used for the fresh transcription is Knuth's secondary description — Taft 1970 is famously unavailable); `NYSIISCode(s) string` + `NYSIISScore(a,b) float64`.
- **MRA** (`AlgoMRA`, reserved at `algoid.go:172-175`) — Moore-Kuhns-Trefftzs-Montgomery 1977, NBS Tech Note 943; `MRACode(s) string` + **`MRACompare(a,b) (bool, int)`** + `MRAScore(a,b) float64` — the `int` is the raw 0-6 NBS similarity counter, exposed per spec line 691.

All four algorithms are inherently boolean (1.0 if codes match, 0.0 otherwise) — they do NOT exercise float-determinism beyond the trivial division-free Score path. The cross-platform determinism gate value is the **encoded code string** itself, not a float.

Phase 7 also lands the additive extensions of Phase 6's surface:

- `permittedMongeElkanInner` map adds `AlgoSoundex`, `AlgoDoubleMetaphone`, `AlgoNYSIIS`, `AlgoMRA` (13 base + AlgoRatcliffObershelp from Phase 6 + 4 phonetic = 18 permitted inners total). Each entry is added in the **same plan** that wires the underlying algorithm dispatch.
- `monge_elkan_test.go` permitted-AlgoID panic-test fixture updates from 14 to 18 in lockstep with each plan.

Phase 7 closes requirement IDs PHON-01 through PHON-04.

Out of scope: the Scorer surface (`WithSoundexAlgorithm(weight)` etc. land in Phase 8); scan / extract sub-packages (Phases 9 / 10); any non-boolean variant of the phonetic Score functions (a future Phase 9+ `XxxCodeNormalised` wrapper accepting `NormaliseOptions` is captured under Deferred Ideas).

</domain>

<decisions>
## Implementation Decisions

### §1. Cross-validation source — LOCKED

**Literature primary + pinned jellyfish supplementary.**

Two layers of validation:

1. **Literature primary (unit tests):** hand-derived reference vectors in each `<algo>_test.go` from the algorithm's primary source — Knuth TAOCP Vol. 3 §6.4 for Soundex and NYSIIS; Philips 2000 + the public-domain C reference for Double Metaphone; NBS Tech Note 943 for MRA. 4–8 canonical reference vectors per algorithm, each with inline derivation comment.

2. **Pinned jellyfish supplementary (cross-validation corpus):**
   - Dev dep: `jellyfish==<exact-pinned-version>` (NOT a runtime dependency — root `go.mod` allowlist remains stdlib + `golang.org/x/text`).
   - `scripts/gen-phonetic-cross-validation.py` declares `JELLYFISH_VERSION = "<exact-pinned-version>"` at top, asserts `jellyfish.__version__ == JELLYFISH_VERSION`, refuses to run on any other version (Phase 6 RapidFuzz pin mechanism — identical pattern).
   - `make regen-phonetic-cross-validation` invokes the script.
   - Output: `testdata/cross-validation/phonetic/vectors.json` — committed to repo; cross-validation tests load and assert byte-stable matches.
   - `vectors.json._metadata` block carries `jellyfish_version`, `python_version`, `regenerated_at` ISO timestamp, and a script checksum.

**Corpus size — tiered by complexity:**

| Algorithm | Entries | Coverage |
|-----------|---------|----------|
| Soundex | 15 | ASCII names + Tymczak (Knuth/Census variant gate) + Ashcraft/Ashcroft (H/W handling) + at-least-one-Knuth-vs-SQL-divergence-marked entry |
| NYSIIS | 20 | ASCII names + Brown/Browne (BRAN reference) + Robert (RABAD ref) + truncation gate (>6 chars input) |
| MRA | 20 | ASCII names + length-difference >3 mismatch gate + threshold-edge pairs (similarity = threshold, ±1) |
| **Double Metaphone** | **40** | All 5 language-origin branches: Germanic (Schmidt/Smith), Slavic (Pacheco), Romance (Catherine/Katherine), Greek (Catherine), Chinese-origin sample — minimum 7 entries per major branch |

**Variant divergence handling — per-entry variant tag:**

Jellyfish's Soundex returns the SQL/MySQL variant (`Tymczak → T520`), not the Knuth/Census variant we're locked to (`T522`). Jellyfish's NYSIIS may also diverge from Knuth's algorithm description. Mechanism:

- Each corpus entry carries an optional `variant_divergence: true` flag plus a `divergent_jellyfish_value: <value>` field.
- For variant-divergent entries, the Go loader test asserts our implementation matches the **Knuth-expected value**, NOT the jellyfish value. The jellyfish value is recorded purely for transparency.
- For non-divergent entries (the majority), our impl is asserted equal to the jellyfish value (byte-stable string match).
- DM and MRA have no known variant divergence; default to direct equality.

**Rationale:** literature vectors guarantee primary-source fidelity; the jellyfish corpus broadens coverage (especially Double Metaphone's 5 language branches, which Pitfall 5 flags as the highest-risk for branch-coverage misses); the variant-tag mechanism eliminates the false-positive failure mode when jellyfish disagrees with our locked variant choice.

**Implication for the planner:**
- Add `scripts/gen-phonetic-cross-validation.py` + Makefile target `regen-phonetic-cross-validation` + `docs/cross-validation.md` extension covering the phonetic corpus.
- Write `phonetic_cross_validation_test.go` (single file with four `t.Run` sub-tests, one per algorithm) that loads the JSON and asserts.
- Document in CONTRIBUTING.md that contributors regenerating phonetic vectors must use the pinned jellyfish version.

### §2. NYSIIS primary-source citation — LOCKED

**Knuth TAOCP Vol. 3 §6.4 is the canonical algorithm-description citation; Taft 1970 cited as algorithmic origin.**

Block comment format (top of `nysiis.go`):

```text
// Source-Origin Statement:
//   Algorithmic origin:  Taft, R. L. (1970). Name search techniques.
//                        New York State Identification and Intelligence System,
//                        Special Report No. 1. Albany, NY.
//   Canonical algorithm description (primary source for fresh transcription):
//                        Knuth, D. E. (1973). The Art of Computer Programming,
//                        Vol. 3, §6.4. Addison-Wesley.
//   Note: Taft 1970 is a NY State Special Report not available through
//   academic publishers; Knuth's secondary description in TAOCP Vol. 3
//   §6.4 is the authoritative algorithm description used for this
//   implementation.
//   Code consulted for reference vectors only:  jellyfish==<pin> (MIT)
```

**Variant choice — LOCKED:** original NYSIIS-1970, **6-character truncation**, **each rule applied once** (no iterate-to-fixed-point). Modified-NYSIIS variants (NYSIIS-1991, "wonderland") are explicitly rejected.

**Cross-validation divergence:** when jellyfish's NYSIIS implementation diverges from Knuth's description, the entry carries `variant_divergence: true` (same pattern as Soundex). Our implementation is asserted against the Knuth-expected value. Surfaces the divergence transparently without compromising primary-source fidelity.

### §3. Double Metaphone implementation strategy — LOCKED

**One plan, validated by language-branch checklist; rule-table provenance audit trail; `--depth=deep` code review for plan 07-02 only.**

Pitfall 5 explicitly flags Double Metaphone as the highest-risk plan in the milestone: ~400 lines, ~200 conditional branches across 5 language-origin branches, and there are several MIT-licensed Go ports a careless implementer could shortcut to.

**Plan decomposition:** single plan **07-02** lands `double_metaphone.go` end-to-end. The implementation transcribes the rule table fresh from Philips 2000 + the public-domain C reference. Validation is a **mandatory language-branch checklist** that MUST pass in the same commit:

- Germanic: `DoubleMetaphoneKeys("Schmidt") = ("XMT", "SMT")` AND `DoubleMetaphoneKeys("Smith") = ("SM0", "XMT")` AND `DoubleMetaphoneScore("Schmidt", "Smith") = 1.0` (the XMT shared-key match).
- Slavic: at least one Slavic-origin name reference vector from the C source.
- Romance: `DoubleMetaphoneKeys("Pacheco")` contains the Spanish `PXK` variant.
- Greek: `DoubleMetaphoneKeys("Catherine") = DoubleMetaphoneKeys("Katherine") = ("K0RN", "KTRN")` — exact match.
- Chinese-origin: at least one reference vector from the C source.

This mirrors Phase 4 RatcliffObershelp's single-plan approach (which was also rule-heavy and shipped without splitting).

**Audit trail — Source-Origin Statement + rule-table provenance line:**

Block comment includes three lines:
1. **Source-Origin Statement** (Phase 4 WR-01 format): primary = Philips 2000; cross-validation = jellyfish + literature reference vectors; code consulted for reference vectors only = jellyfish/literature.
2. **Rule-table provenance line:** `// Rule table derived fresh from: <Philips C reference URL or archive citation>` — the C reference is the canonical implementation; this line documents the artifact the rule structure was transcribed from.
3. **Negative-attribution line:** `// MIT-licensed Go ports NOT consulted: CalypsoSys/godoublemetaphone, deezer/double-metaphone-go, any other Go port` — explicit so reviewers can verify by diff that the rule-table organisation is fresh.

The `algorithm-licensing-reviewer` agent signs off on plan 07-02 and the PR records the sign-off in its description.

**Review depth — `--depth=deep` for plan 07-02 only:**

Plan 07-02's PR triggers `--depth=deep` code review (cross-file analysis, rule-table provenance check, MIT-Go-port comparison). Other Phase 7 plans run standard depth. Phase-level review remains standard. Targets the cost where the risk lives.

### §4. permittedMongeElkanInner extension timing — LOCKED

**Each algorithm's own plan adds its own entry; panic-test fixture updates in lockstep.**

Phase 6's `monge_elkan.go` already documents (in the map's leading comment) that Phase 7 will additively add 4 phonetic AlgoIDs (13 base + AlgoRatcliffObershelp from Phase 6 + AlgoSoundex + AlgoDoubleMetaphone + AlgoNYSIIS + AlgoMRA = 18). The "when" question: each Phase 7 plan that wires the underlying algorithm dispatch ALSO adds its own AlgoID to the map AND updates `monge_elkan_test.go`'s panic-test fixture in the same commit.

Concretely:

| Plan | Algorithm shipped | permittedMongeElkanInner entry added | Panic-test fixture |
|------|-------------------|--------------------------------------|---------------------|
| 07-01 | Soundex | `AlgoSoundex: true` (15 permitted total) | rejected count: 14 → 13 |
| 07-02 | Double Metaphone | `AlgoDoubleMetaphone: true` (16 permitted) | rejected count: 13 → 12 |
| 07-03 | NYSIIS | `AlgoNYSIIS: true` (17 permitted) | rejected count: 12 → 11 |
| 07-04 | MRA | `AlgoMRA: true` (18 permitted) | rejected count: 11 → 10 |

**Rationale:** standard Phase-2/3/4/5/6 pattern (each PR is self-contained). The exhaustive panic test in `monge_elkan_test.go` will fail intermediately ONLY if the map and the test fixture are out of sync, which the PR boundary catches. Co-locating the allow-list mutation with the dispatch wiring eliminates the "I forgot to add it" failure mode.

**Binary-inner ME composition property test — one per phonetic inner:**

Add four new property tests in `monge_elkan_test.go`:

- `TestMongeElkanScore_BinaryInner_Soundex` — asserts `MongeElkanScore("alpha beta", "alpha gamma", AlgoSoundex, opts) == 0.5` (one of two tokens matches; the other doesn't); also asserts `MongeElkanScore("alpha beta", "alpha beta", AlgoSoundex, opts) == 1.0` (full match) and `MongeElkanScore("alpha", "gamma", AlgoSoundex, opts) == 0.0` when neither token matches phonetically.
- Same shape for `AlgoDoubleMetaphone`, `AlgoNYSIIS`, `AlgoMRA`.

Locks the binary-inner-composition behaviour against silent regression (e.g. someone changes the per-token-max accumulation logic in a way that breaks ME over discrete-valued inners).

### §5. Non-ASCII input handling — LOCKED

**Skip silently, document limitation prominently in godoc.** No `NormaliseOptions` parameter on Phase 7 phonetic functions in v1.0.

Iterate input rune-by-rune; only ASCII `[A-Za-z]` (after case-fold) participates in encoding. Non-ASCII runes (`ü`, `é`, `中`, emoji, combining marks) are skipped before encoding. Examples:

- `SoundexCode("Müller")` → encoding sees `"Mller"` → `"M460"`.
- `DoubleMetaphoneKeys("Café")` → encoding sees `"Cf"` → `("KF", "")` or similar (per actual Philips rules).
- `NYSIISCode("Brønning")` → encoding sees `"Brnning"` → result per the rules.

**Godoc warning paragraph (canonical text — copy verbatim into each phonetic algorithm's package-level godoc):**

```text
// Non-ASCII input handling: this algorithm operates on ASCII letters
// [A-Za-z] only. Non-ASCII runes (accented characters, emoji,
// combining marks) are dropped silently before encoding. For
// Unicode-aware similarity on non-ASCII input, compose with
// Normalise + diacritic stripping before calling this function, or
// use a character-based algorithm (e.g. Levenshtein, Jaro-Winkler).
```

**Fuzz seeds — both regimes:**

Each phonetic `_fuzz_test.go` seeds the corpus with both regimes:
- ASCII-only inputs (the encoded regime): every literature reference vector + typical English-language names.
- Mixed ASCII + non-ASCII inputs (the skip regime): `"Müller"`, `"Café"`, `"中文"`, `"🎉hello"`, `""`, pure-non-ASCII inputs.

Fuzz invariants verify (a) no panic, (b) range bounds 0.0/1.0 for Score, (c) the encoded code consists only of `[A-Z0-9]` characters (or `[A-Z0-9 ]` for MRA's space-separated form).

### §6. MRACompare API surface — LOCKED

**Three public functions per spec line 691: `MRACode`, `MRACompare`, `MRAScore`. Keep the `(bool, int)` return shape.**

```go
func MRACode(s string) string                       // canonical encoded form
func MRACompare(a, b string) (matched bool, simScore int)  // raw 0-6 NBS similarity counter + threshold-rule match decision
func MRAScore(a, b string) float64                  // binary 0.0 / 1.0; dispatch-table entry
```

The raw 0-6 integer counter is the granular NBS-Tech-Note-943 similarity score; exposing it directly is faithful to the primary source. `MRAScore` is the dispatch-table wrapper that converts the bool to 0.0/1.0.

**Rationale:** consumers wanting the integer counter for downstream composition (e.g. for sorting candidates by raw MRA similarity) get it without parsing internal state. Spec line 691 already promises this surface; the catalogue's float64-only convention is documented as a default, not an absolute (e.g. Smith-Waterman-Gotoh already returns a raw integer similarity via `SmithWatermanGotohRawScore`).

### §7. Phonetic golden file schema — LOCKED

**Separate `testdata/golden/phonetic-codes.json` for byte-stable code vectors; `algorithms.json` schema unchanged for Score entries.**

Two golden surfaces:

1. **`testdata/golden/algorithms.json`** carries the binary Score entries for `SoundexScore`, `DoubleMetaphoneScore`, `NYSIISScore`, `MRAScore`. Schema unchanged from Phase 6 — `{algorithm, a, b, score float64}` rows. Phase 7 finalisation extends the file to 23 algorithms with the typical per-algorithm staging-golden → merge workflow.

2. **`testdata/golden/phonetic-codes.json`** (new file in plan 07-01) carries the byte-stable code vectors:

   ```json
   {
     "_metadata": {
       "purpose": "Cross-platform byte-stable phonetic code determinism gate",
       "regenerated_at": "<ISO>"
     },
     "entries": [
       {"algorithm": "Soundex", "input": "Robert", "code": "R163"},
       {"algorithm": "Soundex", "input": "Tymczak", "code": "T522"},
       {"algorithm": "DoubleMetaphone", "input": "Schmidt", "primary": "XMT", "secondary": "SMT"},
       {"algorithm": "NYSIIS", "input": "Brown", "code": "BRAN"},
       {"algorithm": "MRA", "input": "Robert", "code": "RBRT"}
     ]
   }
   ```

   Per algorithm: 8-12 entries from the literature reference vectors + Phase 7 distinctive cases (Tymczak for Soundex, Schmidt-Smith pair for DM, Brown-Browne for NYSIIS, threshold-edge cases for MRA).

   `algorithms_golden_test.go` (or a new `phonetic_codes_golden_test.go`) loads this file and asserts byte-stable code matches on every CI platform. This is the cross-platform determinism gate for phonetic algorithms — strings instead of floats, but the same purpose.

**Rationale:** the code is the cross-platform determinism value for phonetic algorithms; algorithms.json is structured around float scores; separating concerns keeps each golden file internally consistent. The existing `algorithms.json` schema and loader stay simple.

### §8. Identifier-similarity example extension — LOCKED

**4 new score columns in `examples/identifier-similarity/main.go`; new standalone `examples/phonetic-keys/main.go` for the encoded-key surface.**

- `examples/identifier-similarity/main.go` (Phase 6 left at 19 columns) gains 4 new columns: `Soundex`, `DblMetaph`, `NYSIIS`, `MRA` — each showing the 0.0/1.0 binary score for the existing identifier-pair rows. Total: 23 columns. `main_test.go` golden stdout fixture regenerated in the same commit.
- New program `examples/phonetic-keys/main.go` with companion `main_test.go`. Demonstrates the `SoundexCode` / `DoubleMetaphoneKeys` / `NYSIISCode` / `MRACode` / `MRACompare` functions standalone. Side-by-side table of input names and their encoded forms across the four algorithms. Educational; called out in README's "Examples" section.

**Plan placement:** the identifier-similarity column extension lands in the **finalisation plan** (07-05 or 07-06 depending on plan count after planning). The new `examples/phonetic-keys` program lands ALSO in the finalisation plan to keep the new example program scoped with its golden fixture.

### §6-prior. Established patterns (LOCKED — inherited from Phase 2/3/4/5/6)

The following patterns are CARRIED FORWARD without re-discussion:

- **File-by-file structure:** `<algo>.go` (implementation) + `dispatch_<algo>.go` (AlgoID slot wiring) + `<algo>_test.go` (unit + reference vectors) + `<algo>_bench_test.go` + `<algo>_fuzz_test.go`. Phonetic algorithms have less DP machinery than Phase 6's `token_indel.go`, so no shared kernel file; each algorithm is self-contained.
- **AlgoID slots already reserved in algoid.go:** `AlgoSoundex` (156-159), `AlgoDoubleMetaphone` (161-165), `AlgoNYSIIS` (167-170), `AlgoMRA` (172-175). Planner wires `dispatch_<algo>.go` files but does NOT modify algoid.go enum positions.
- **Direct call panics; Scorer returns errors:** phonetic algorithms have no parameters (no `Options` struct), so there are no panic-on-bad-parameter paths. The Scorer (Phase 8) will return `ErrInvalidAlgoID` for unknown AlgoIDs (already declared in Phase 1).
- **No map iteration on output path (DET-03):** phonetic algorithms don't iterate maps on output paths.
- **No transcendental floats (DET-06):** phonetic algorithms compute scores via integer compare; no `math.X` at all.
- **BDD scenarios:** one feature file per algorithm in `tests/bdd/features/`. Plus a `monge_elkan_phonetic_inner.feature` (in plan 07-04 or finalisation) covering the binary-inner-composition cases.
- **Staging golden → finalisation merge:** `testdata/golden/_staging/<algo>.json` during each algorithm's plan; merged into `testdata/golden/algorithms.json` in finalisation. PLUS `testdata/golden/phonetic-codes.json` (new — per §7).
- **Per-plan llms.txt + llms-full.txt sync:** every new exported symbol gets a line in `llms.txt` and a full entry in `llms-full.txt` IN THE SAME PLAN that adds the symbol — NOT deferred to finalisation.
- **Phase 6 §5 DoS-vector godoc format:** does NOT apply to phonetic algorithms — they are O(n) or O(n²) but on bounded inputs (typical name length < 50 chars), the worst-case cost is < 1µs. No DoS-vector godoc block needed; no `Pathological_*` bench fixtures.

### Claude's Discretion

The planner (gsd-planner) chooses, without re-asking the user:

- Wave decomposition: 4 algorithms + finalisation = 5 plans. Likely shape: 07-01 Soundex (simplest, foundation for cross-validation infra), 07-02 Double Metaphone (heaviest, code-review-deep), 07-03 NYSIIS, 07-04 MRA, 07-05 finalisation. Adjust wave grouping if dependency analysis surfaces a different shape (e.g. if MRA's threshold rules suggest pairing with NYSIIS).
- Exact jellyfish version pin (planner picks current stable at planning time; records version + sha256 in script header — same pattern as RapidFuzz pin in Phase 6).
- Exact number of staging-golden entries per algorithm in `testdata/golden/_staging/<algo>.json` (8-12 per Phase 2-6 norm).
- Whether to ship `phonetic_codes_golden_test.go` as a separate test file or embed the loader into `algorithms_golden_test.go` (recommend separate file for clean separation).
- Fuzz seed counts per algorithm (8-16 seeds covering ASCII + non-ASCII regimes per §5).
- BDD scenario count per algorithm (~4-6 scenarios mirroring Phase 6 algorithm BDDs).
- Exact wording of the godoc warning paragraph beyond the canonical text in §5 (each algorithm may add specifics — e.g. NYSIIS notes its English-name tuning).
- Whether plan 07-02's `--depth=deep` review is invoked automatically by the executor or manually requested in the PR description (executor decision — same outcome).
- The granular structure of `examples/phonetic-keys/main.go` (column layout, name set, header rows) — finalisation-plan-time decision constrained by the golden stdout fixture.

</decisions>

<canonical_refs>
## Canonical References

**Downstream agents MUST read these before planning or implementing.**

### Spec & requirements (project-internal)

- `docs/requirements.md` §7.4 — full phonetic algorithm specifications (§7.4.1 Soundex, §7.4.2 Double Metaphone, §7.4.3 NYSIIS, §7.4.4 MRA)
- `docs/requirements.md` §11 — Phonetic Algorithm Integration (Scorer binary-0/1 behaviour; Phase 8 consumption)
- `docs/requirements.md` §13.3, §13.4 — determinism rules (no map iteration on output, cross-platform byte stability) — apply to encoded codes here
- `docs/requirements.md` §15.3, §15.4 — property test conventions, fuzz target conventions
- `.planning/REQUIREMENTS.md` lines 50-53 — PHON-01..PHON-04 traceability table
- `.planning/REQUIREMENTS.md` lines 224-227 — PHON requirement status (currently Pending; flips to Met after Phase 7)
- `.planning/ROADMAP.md` Phase 7 section — goal + success criteria (Tymczak/T522 + Schmidt-Smith XMT + Brown/BRAN + NBS Tech Note 943 references)
- `.planning/PROJECT.md` — Metaphone 3 patent exclusion (out-of-scope reasoning); zero-cgo and pure-Go-stdlib constraints carry forward
- `.planning/research/PITFALLS.md` Pitfall 4 — Soundex variant divergence (Tymczak T522 vs T520) — direct evidence base for §1 LOCKED
- `.planning/research/PITFALLS.md` Pitfall 5 — NYSIIS iteration / Double Metaphone language-branch coverage — direct evidence base for §2 and §3 LOCKED

### Primary academic / engineering sources (research targets)

- **Russell, R. C., Odell, M. K. (1918, 1922).** U.S. Patents 1261167 and 1435663 — Soundex original. (Now public domain; cited for completeness; algorithm description is in Knuth.)
- **Knuth, D. E. (1973).** *The Art of Computer Programming, Vol. 3: Sorting and Searching*, Section 6.4. Addison-Wesley — **canonical algorithm description for both Soundex (Knuth/Census variant) and NYSIIS**. The primary citation in every Phase 7 phonetic algorithm file (alongside the algorithmic origin).
- **Taft, R. L. (1970).** *Name search techniques*. New York State Identification and Intelligence System, Special Report No. 1. Albany, NY — NYSIIS algorithmic origin. Famously unavailable; Knuth used as canonical description per §2 LOCKED.
- **Philips, L. (2000).** "The double-metaphone search algorithm." *C/C++ Users Journal*, 18(6):38–43 — Double Metaphone primary source. Article available via archived C/C++ Users Journal scans; canonical C reference implementation in the public domain.
- **Moore, G. B., Kuhns, J. L., Trefftzs, J. L., Montgomery, C. A. (1977).** *Accessing individual records from personal data files using non-unique identifiers*. National Bureau of Standards (NIST), Technical Note 943 — MRA Match Rating Approach primary source.

### External cross-validation dependency (developer toolchain only — NOT runtime)

- **`jellyfish` Python package** (pip install) — required for `make regen-phonetic-cross-validation`. Pinned to a specific version in `scripts/gen-phonetic-cross-validation.py`. NEVER added to root `go.mod`. Jellyfish's Soundex returns the SQL/MySQL variant; jellyfish's NYSIIS may diverge from Knuth's description. Both divergences handled via the per-entry `variant_divergence` tag per §1.

### Negative-attribution (MIT-licensed Go ports NOT consulted)

These exist in the Go ecosystem but MUST NOT be copied from for any Phase 7 algorithm. The `algorithm-licensing-reviewer` agent verifies by diff.

- `github.com/xrash/smetrics` (MIT) — Soundex (clean reference, but fresh transcription mandated)
- `github.com/CalypsoSys/godoublemetaphone` (MIT) — Double Metaphone (single-file port; Pitfall 5's primary attractive-nuisance)
- `github.com/tilotech/go-phonetics` (MIT) — Soundex + Metaphone
- `github.com/UjjwalAyyangar/go-jellyfish` (MIT) — phonetic algorithms inc. NYSIIS
- `github.com/jamesturk/jellyfish` (BSD-2) — the Python lib's earlier Go reimplementations; jellyfish Python is OK as a cross-validation reference (vectors only).

### Project skills (correctness & licensing gates)

- `.claude/skills/algorithm-correctness-standards/SKILL.md` — primary-source citation format, reference vectors, formula docs (phonetic algorithms don't have formulas but DO have rule-table descriptions)
- `.claude/skills/algorithm-licensing-standards/SKILL.md` — **LOAD-BEARING for Phase 7.** Fresh-implementation discipline, attribution format, MIT Go ports as negative-attribution. plan 07-02 (Double Metaphone) gates on this.
- `.claude/skills/determinism-standards/SKILL.md` — no map iteration, golden files (phonetic codes are the determinism gate value)
- `.claude/skills/performance-standards/SKILL.md` — allocation budgets (phonetic algorithms are O(n) with small constants; budgets are easy to meet)
- `.claude/skills/go-coding-standards/SKILL.md` — Go style, no testify in root tests
- `.claude/skills/go-testing-standards/SKILL.md` — unit + property + fuzz + bench + BDD coverage targets (≥ 95% overall, ≥ 90% per file, 100% on public API)
- `.claude/skills/fuzzymatch-review-protocol/SKILL.md` — agent gate sequence; algorithm-licensing-reviewer is the gating reviewer for Phase 7

### Prior phase context (carry-forward)

- `.planning/phases/03-smith-waterman-gotoh/03-CONTEXT.md` — Python-generator-plus-committed-corpus cross-validation pattern (Phase 6 inherited from this; Phase 7 inherits the pattern again with `jellyfish` instead of `rapidfuzz` or `biopython`)
- `.planning/phases/04-remaining-character-gestalt/04-CONTEXT.md` — staging-golden → finalisation merge, identifier-similarity example extension pattern (Phase 7 extends identifier-similarity to 23 columns AND adds a new phonetic-keys example program)
- `.planning/phases/05-q-gram-algorithms/05-CONTEXT.md` — direct-call-panic + Scorer-returns-error split; per-plan llms.txt sync discipline (carries forward to Phase 7's per-algorithm llms.txt entries)
- `.planning/phases/06-token-based-algorithms/06-CONTEXT.md` — RapidFuzz pin mechanism (Phase 7's jellyfish pin mirrors exactly); `permittedMongeElkanInner` map design and Phase 7's additive extension pattern (now LOCKED in §4 above); Phase 6's DoS-vector godoc format (does NOT apply to Phase 7 — phonetic algorithms are not DoS-vectors)

</canonical_refs>

<code_context>
## Existing Code Insights

### Reusable Assets

- **`algoid.go` lines 156-175** — `AlgoSoundex`, `AlgoDoubleMetaphone`, `AlgoNYSIIS`, `AlgoMRA` enum slots already reserved with godoc citing primary sources. Slots are stable; planner wires dispatch files but does NOT renumber.
- **`algoid.go` String() / dispatch tables (lines 251-257, 302-305)** — string conversions return canonical names (`"Soundex"`, `"DoubleMetaphone"`, `"NYSIIS"`, `"MRA"`). Dispatch table at lines 280+ already includes these AlgoIDs as legal entries; `dispatch_<algo>.go` files in this phase wire the actual algorithm functions.
- **`monge_elkan.go` `permittedMongeElkanInner` map** — already documents (in leading comment) that Phase 7 adds 4 phonetic AlgoIDs. §4 LOCKED specifies the lockstep mutation pattern; each plan adds its own entry + updates panic-test fixture.
- **`errors.go`** — `ErrInvalidAlgoID` is declared at the package level; consumed by Phase 8's `WithSoundexAlgorithm(weight)` etc.
- **`testdata/golden/algorithms.json`** — established golden file (144 entries × 19 algorithms after Phase 6 finalisation). Phase 7 finalisation extends to ≈ 180 entries × 23 algorithms with the typical staging-golden → merge workflow.
- **`testdata/golden/_staging/` workflow** — pattern proven over 16 algorithm staging files. Phase 7: 4 new staging files (one per phonetic algorithm).
- **`scripts/gen-token-ratio-cross-validation.py`** — direct template for `scripts/gen-phonetic-cross-validation.py`. RapidFuzz pin mechanism translates 1:1 to jellyfish pin mechanism.
- **`docs/cross-validation.md`** — established in Phase 6; Phase 7 extends with a "Phonetic cross-validation" section describing the jellyfish corpus + variant-divergence tagging.
- **`tests/bdd/steps/algorithms_steps.go`** — accumulator file. Each new phonetic algorithm appends a `step.Step` registration following the Phase 2-6 pattern.
- **`example_test.go`** — runnable godoc examples; one new `ExampleXxx` per public function. Phase 7: ≈ 9 new examples minimum (SoundexCode, SoundexScore, DoubleMetaphoneKeys, DoubleMetaphoneScore, NYSIISCode, NYSIISScore, MRACode, MRACompare, MRAScore).
- **`props_test.go`** — appendable property-test accumulator. Phase 7 appends Five-invariant blocks per algorithm + Symmetric for all four + the 4 new `TestMongeElkanScore_BinaryInner_<Algo>` tests (per §4).
- **`bench.txt`** — Phase 6 regenerated baseline. Phase 7 finalisation full-replaces it to include the four new benchmarks. No pathological fixtures needed (phonetic algorithms are O(n) on bounded inputs).
- **`llms.txt` / `llms-full.txt`** — every new exported symbol needs a line in `llms.txt` and a full entry in `llms-full.txt` (per-plan discipline).
- **`examples/identifier-similarity/main.go` + `main_test.go`** — Phase 6 left at 19 columns. Phase 7 finalisation extends to 23 columns with `Soundex / DblMetaph / NYSIIS / MRA` score columns.

### Established Patterns (Phase 7 must follow)

- **Source-Origin Statement format** — Phase 4 WR-01 established the precedent. Phase 7's Double Metaphone extends this format with a rule-table provenance line and a MIT-Go-port negative-attribution line per §3.
- **Map-iteration discipline (DET-03 from Phase 1)** — does NOT apply to phonetic algorithms (no maps on output paths).
- **OQ-X RESOLUTION LOCKED format** — non-obvious local decisions recorded in plan SUMMARY.md as `OQ-N RESOLUTION LOCKED <date>` with rationale. Phase 7 candidate OQs: jellyfish version pin choice, exact reference-vector counts in literature corpora, exact threshold-edge MRA pairs.
- **gsd-executor.md scope-boundary rule** — out-of-scope discoveries during execution are logged to `.planning/phases/07-phonetic-algorithms/deferred-items.md` (created on first need), NOT rolled into the current commit.
- **Direct-call panic + Scorer-returned-error split** — does NOT apply to phonetic algorithms (no parameters to validate); Phase 8's Scorer wrappers will use `ErrInvalidAlgoID` for unknown AlgoIDs.

### Integration Points

- **Phase 6 (Monge-Elkan):** Phase 7's plans 07-01..07-04 each add their AlgoID to `permittedMongeElkanInner` and update `monge_elkan_test.go`'s panic-test fixture in lockstep. Phase 6's binary-inner ME composition (per §4) gets 4 new tests in `monge_elkan_test.go` (one per phonetic inner).
- **Phase 8 (Composite Scorer):** `WithSoundexAlgorithm(weight)`, `WithDoubleMetaphoneAlgorithm(weight)`, `WithNYSIISAlgorithm(weight)`, `WithMRAAlgorithm(weight)` options consume the four algorithms. The Scorer parameter-validation path uses `ErrInvalidAlgoID` for unknown AlgoIDs (already declared).
- **Phase 9 (Scan) / Phase 10 (Extract):** consume the dispatch table; phonetic algorithms participate via their AlgoID slots without scan/extract needing to know they are phonetic-flavoured.

</code_context>

<specifics>
## Specific Ideas

- **Soundex Tymczak gate vector:** `SoundexCode("Tymczak") == "T522"` MUST be in the literature unit tests as the discriminating gate vector. Asserts the Knuth/Census variant choice over the SQL/MySQL variant (Pitfall 4). Same fixture is the `variant_divergence: true` entry in the jellyfish corpus.
- **Double Metaphone language-branch fixture pairs (mandatory):**
  - Germanic: `DoubleMetaphoneKeys("Schmidt") = ("XMT", "SMT")`; `DoubleMetaphoneKeys("Smith") = ("SM0", "XMT")`; `DoubleMetaphoneScore("Schmidt", "Smith") = 1.0`.
  - Slavic: at least one name from the C source's Slavic branch.
  - Romance: `DoubleMetaphoneKeys("Pacheco")` contains `PXK` (Spanish branch).
  - Greek: `DoubleMetaphoneKeys("Catherine") = DoubleMetaphoneKeys("Katherine") = ("K0RN", "KTRN")`.
  - Chinese-origin: at least one Chinese-origin name from the C source.
- **NYSIIS Brown/Browne pair:** `NYSIISCode("Brown") = NYSIISCode("Browne") = "BRAN"`. Plus `NYSIISCode("Robert") = "RABAD"` (RABAD vs RABERT distinguishes Knuth's description from common Wikipedia variants).
- **MRA length-difference >3 mismatch gate:** `MRAScore("Smith", "JohnathanLongName") = 0.0` because `|len(MRACode(a)) - len(MRACode(b))| > 3` is a documented automatic mismatch.
- **MRA threshold-edge pairs:** at least one pair where the raw similarity counter equals the threshold (the boundary case) AND one pair where it's threshold-minus-1 (just below). Exercises both branches of the bool decision.
- **Soundex Ashcraft/Ashcroft pair:** `SoundexCode("Ashcraft") = SoundexCode("Ashcroft") = "A261"`. H/W-handling discriminating vector per Pitfall 4 (separator-as-H/W variant returns different codes).
- **Binary-inner ME composition fixtures (per §4):** `MongeElkanScore("alpha beta", "alpha gamma", AlgoSoundex, opts) == 0.5` (one token matches, one doesn't); the analogous fixture for each phonetic AlgoID.
- **Negative-attribution audit:** plan 07-02 PR description includes the line `algorithm-licensing-reviewer sign-off: <sha or hash> — rule table transcribed fresh from Philips 2000 + C reference; MIT Go ports NOT consulted (verified by diff against CalypsoSys/godoublemetaphone source layout)`.

</specifics>

<deferred>
## Deferred Ideas

- **`XxxCodeNormalised(s string, opts NormaliseOptions) string` wrappers:** considered during §5 non-ASCII discussion; deferred to Phase 9+ as an additive expansion. Each phonetic algorithm could ship a `<algo>_normalised.go` wrapper that calls `Normalise(s, opts)` to strip diacritics + NFD-fold then forwards to `XxxCode`. Future-Johnny does not need to re-litigate the non-ASCII handling decision — Phase 7 commits to the silent-skip discipline; v1.x expansion provides the diacritic-aware path for callers who want it. Document this in CONTEXT.md so the deferred option survives across context windows.
- **`NYSIISOptions{IterateToFixedPoint: bool}` for the modified-NYSIIS variant:** considered during §2 variant choice; deferred to v1.x as additive change if real demand emerges. Original Taft-1970 algorithm is sufficient for v1.0; modified-NYSIIS is a minor accuracy improvement on rare cases.
- **Soundex SQL/MySQL variant as a second public function:** considered during Pitfall 4 cross-reference; rejected for v1.0 because the spec commits to Knuth/Census exclusively. v1.x could add `SoundexSQLCode` as an additive function for consumers needing exact SQL-Soundex parity.
- **Acquiring Taft 1970 PDF via interlibrary loan:** considered during §2 NYSIIS sourcing; rejected for v1.0 because Knuth's secondary description is authoritative enough AND the ILL timeline could block phase planning. v1.x could revisit if a researcher contributes the PDF or NY State Archives digitises it.
- **Per-rule-branch attribution comments in `double_metaphone.go`:** considered during §3 audit-trail discussion; rejected as too granular for the file. Branch-level provenance is captured at the file-level via the Source-Origin Statement + rule-table provenance line + negative-attribution line.
- **Cross-validation against Python `abydos` library (a more comprehensive phonetic-algorithms package than jellyfish):** considered during §1 corpus discussion; rejected because adding a second pip dep doubles the developer-toolchain footprint without clear value. Jellyfish alone is sufficient for variant-divergence-tagged cross-validation.
- **Phonetic-aware fuzz invariant: `XxxCode` output character set:** captured as an invariant in §5 fuzz seeds (output must be `[A-Z0-9]` for Soundex/DoubleMetaphone/NYSIIS, `[A-Z0-9 ]` for MRA). Already in scope of the fuzz tests; not a deferred idea.
- **`MRACompare` integer score normalisation to float64:** considered during §6 API surface; rejected because exposing the raw 0-6 counter is faithful to the NBS-943 primary source. v1.x could add a `MRACompareNormalised(a, b) (bool, float64)` wrapper if consumers want catalogue-wide float64 uniformity.
- **Combining `phonetic-codes.json` golden file with `algorithms.json`:** considered during §7 schema discussion; rejected because the schemas are structurally different (string codes vs float scores). Separate files are clearer. v1.x could revisit if a unified loader becomes architecturally preferable.

</deferred>

---

*Phase: 7-phonetic-algorithms*
*Context gathered: 2026-05-15*
