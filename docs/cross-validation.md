# Cross-validation corpora

The fuzzymatch project cross-validates select algorithms against external
reference implementations to close the
`algorithm-correctness-reviewer` gate per the Workflow — Agent Gates
section of `CLAUDE.md`. Each cross-validation corpus lives under
`testdata/cross-validation/<algorithm-or-tier>/vectors.json` and is
loaded directly by the corresponding Go test — **CI does not require
Python at test time**. The committed JSON is the verification fixture.

Three corpora exist today:

| Algorithm tier        | Reference implementation                            | Generator script                                          | Go test                                  |
|-----------------------|-----------------------------------------------------|------------------------------------------------------------|------------------------------------------|
| Smith-Waterman-Gotoh  | biopython `Bio.Align.PairwiseAligner`               | `scripts/gen-swg-cross-validation.py`                       | `TestSWG_CrossValidation`                |
| Ratcliff-Obershelp    | Python stdlib `difflib.SequenceMatcher(autojunk=False)` | `scripts/gen-ratcliff-obershelp-cross-validation.py`        | `TestRatcliffObershelp_CrossValidation`  |
| Token tier (this doc) | `rapidfuzz==3.14.5`                                 | `scripts/gen-token-ratio-cross-validation.py`               | `TestTokenRatios_CrossValidation`        |

This document focuses on the **token tier** corpus introduced by Phase
6 plan 06-01. It is the canonical reference for the four
Indel-based ratios in the catalogue: TokenSortRatio, TokenSetRatio,
PartialRatio (bytes), PartialRatio (runes).

## Why RapidFuzz, not fuzzywuzzy

`fuzzywuzzy` (SeatGeek, 2014) is the historical Python library that
popularised the four ratios. RapidFuzz (Max Bachmann, 2020-present) is
its drop-in successor — it fixes several scoring inconsistencies in the
pure-Python vs C-extension paths of fuzzywuzzy, documents the **Indel
formula** as the canonical normalisation
(<https://rapidfuzz.github.io/RapidFuzz/Usage/distance/Indel.html>), and
is actively maintained. We cross-validate against RapidFuzz exclusively;
fuzzywuzzy is referenced only as historical context.

## Pinned version (LOCKED)

`scripts/gen-token-ratio-cross-validation.py` pins
`RAPIDFUZZ_VERSION = "3.14.5"`. The script's first action after importing
rapidfuzz is

```python
assert rapidfuzz.__version__ == RAPIDFUZZ_VERSION, …
```

which **refuses to run** if the installed rapidfuzz version is different.
The Go-side loader test asserts `_metadata.rapidfuzz_version` matches the
committed string. The two checks together prevent silent corpus drift in
both directions: a developer cannot regenerate with an unpinned
rapidfuzz, and a tampered or stale corpus surfaces in CI immediately.

### Bumping the pinned version

The five-step protocol:

1. Update `RAPIDFUZZ_VERSION` in `scripts/gen-token-ratio-cross-validation.py`.
2. Update the install hint in the Makefile target
   `regen-token-ratio-cross-validation`.
3. Run `make regen-token-ratio-cross-validation` (after installing the
   new version: `python3 -m pip install --user rapidfuzz==<new>`).
4. Commit the regenerated
   `testdata/cross-validation/token-ratios/vectors.json`.
5. Run `go test -run TestTokenRatios_CrossValidation -race ./...`. Any
   score drift between the new RapidFuzz release and the project
   implementation surfaces as a per-entry test failure and **requires
   `algorithm-correctness-reviewer` review** before merge.

## OQ-1 RESOLUTION — tokeniser-divergence

**LOCKED in plan 06-01** — recorded inline in both the generator script
header and the algorithm godoc.

RapidFuzz tokenises via Python `str.split()` — whitespace-only,
case-preserving. fuzzymatch's `Tokenise(s, DefaultTokeniseOptions())` is
**identifier-aware** (camelCase / snake_case / kebab-case / dot-case)
AND **lowercasing**. To make cross-validation tractable, the corpus is
restricted to inputs where the two tokenisations agree — pure
whitespace-separated lowercase ASCII text. The generator script's
`_assert_corpus_is_tokenise_safe` gate enforces this constraint for
TokenSortRatio / TokenSetRatio entries. It rejects any case whose `a`
or `b` field contains a character outside `[a-z ]`, exiting the script
with a clear error message.

The generator also calls `.lower()` on every input before passing to
RapidFuzz — reconciling RapidFuzz's case-preserving default with
fuzzymatch's lowercasing Tokenise. For Tokenise-safe inputs (already
lowercase) the fold is a no-op; the explicit call documents the
reconciliation and protects against accidental uppercase characters
slipping in via future contributions.

PartialRatio entries are character-level — they don't tokenise — and
**may** carry the `"partial_only": true` flag if their input shape
would violate the Tokenise-safety constraint. The generator emits
`null` for `token_sort_ratio` / `token_set_ratio` on those entries; the
Go loader skips the corresponding sub-tests via the `if e.TokenSortRatio
== nil` check.

## OQ-2 RESOLUTION — single combined corpus

**LOCKED in plan 06-01.**

One `vectors.json` carries all four scores per entry —
`token_sort_ratio`, `token_set_ratio`, `partial_ratio_bytes`,
`partial_ratio_runes`. Per-entry sub-tests via
`t.Run(e.Name+"/<surface>", …)` mirror the Phase 3 / Phase 4 structure.
The Wave-1 loader (plan 06-01) asserts only TokenSortRatio entries;
plans 06-02 (TokenSetRatio) and 06-03 (PartialRatio) remove the
`t.Skip` calls in the existing per-surface sub-tests and add the
assertion bodies.

## OQ-3 RESOLUTION — `partial_ratio_runes` always emitted

**LOCKED in plan 06-01.**

The generator emits `partial_ratio_runes` for every entry — for ASCII
inputs the value coincides with `partial_ratio_bytes` (RapidFuzz
operates over codepoint-indexed Python strings; the byte/rune
distinction is a fuzzymatch implementation detail). Emitting both
fields lets the rune-path Go implementation be cross-validated against a
dedicated surface, catching any rune-path-specific regression even
when the input is ASCII.

## Empty-token-set deviation (TokenSetRatio only)

RapidFuzz's `token_set_ratio` returns `0` (not `100`) when either
token-set is empty. This is bug-for-bug compatibility with fuzzywuzzy
issue #110. fuzzymatch's TokenSetRatioScore (plan 06-02) matches this
deviation — see the plan 06-02 spec for the locked decision. Other
tokenised algorithms (TokenJaccard, MongeElkan) follow the standard
both-empty → 1.0 convention.

## Regeneration

Developer command:

```bash
python3 -m pip install --user rapidfuzz==3.14.5    # once per workstation
make regen-token-ratio-cross-validation
```

The Makefile target invokes `python3
scripts/gen-token-ratio-cross-validation.py`, which writes the corpus to
`testdata/cross-validation/token-ratios/vectors.json` with the
`_metadata.rapidfuzz_version`, `_metadata.python_version`, and
`_metadata.regenerated_at` fields populated. Repeated runs produce
byte-identical scoring fields (only the timestamp varies). The
generator is intentionally simple — no caching, no parallelism, no
randomness.

CI does NOT regenerate the corpus. The committed JSON is the fixture;
CI runs `go test ./...` against it directly. This keeps Python out of
the CI hot path and ensures cross-validation results are reproducible
across runners.

## Phonetic cross-validation

Phase 7 introduces a new cross-validation corpus for the four phonetic
algorithms: `testdata/cross-validation/phonetic/vectors.json`. The
generator script and Go loader mirror the token-ratio pattern but with
two pinned Python packages instead of one.

### Why dual-pin (jellyfish + Metaphone)

**OQ-1 RESOLUTION LOCKED 2026-05-15:** `jellyfish` 1.x removed its Double
Metaphone implementation — `jellyfish.metaphone` is single-key (Metaphone),
not Double Metaphone. To cross-validate the 40-entry Double Metaphone corpus
(covering 5 language-origin branches: Germanic, Slavic, Romance, Greek,
Chinese-origin), a second Python package is required.

The `Metaphone` PyPI package (`oubiwann/metaphone`, BSD-3-Clause), version
0.6, is Andrew Collins' Python translation of Lawrence Philips' public-domain
C++ reference implementation (2000). It provides `metaphone.doublemetaphone(s)
→ (primary, secondary)` and is the de-facto Python Double Metaphone canonical
source (1,034 GitHub stars as of May 2026).

Both pins are declared at the top of `scripts/gen-phonetic-cross-validation.py`:

```python
JELLYFISH_VERSION = "1.2.1"   # Soundex, NYSIIS, MRA
METAPHONE_VERSION = "0.6"     # Double Metaphone only
```

The script asserts both versions at startup and refuses to run on any mismatch.

### Algorithm coverage

| Algorithm       | Reference package | Notes |
|-----------------|-------------------|-------|
| Soundex         | `jellyfish==1.2.1` | Knuth/Census H/W-skip variant — matches our implementation exactly (RESEARCH.md §1.1) |
| Double Metaphone | `Metaphone==0.6` | 40 entries, 5 language-origin branches (≥ 7 per branch) |
| NYSIIS          | `jellyfish==1.2.1` | Modified (non-truncated) variant — see variant_divergence below |
| MRA             | `jellyfish==1.2.1` | `match_rating_codex()` + `match_rating_comparison()` |

### Variant-divergence flag mechanism

Some algorithms have variant choices where our implementation (following
the primary academic source) differs from jellyfish's default:

- **Soundex:** jellyfish 1.2.1 uses the Knuth/Census variant (H/W are NOT
  separators). Our implementation is identical — no expected divergences for
  canonical inputs. The `variant_divergence` schema is retained for
  completeness.

- **NYSIIS (load-bearing):** jellyfish emits non-truncated codes (e.g.
  `Catherine → CATARAN`, 7 chars). Our implementation truncates to 6 characters
  per Taft 1970 / Knuth TAOCP Vol. 3 §6.4. **All jellyfish NYSIIS outputs
  longer than 6 characters carry `variant_divergence: true`.** The Go loader
  asserts our implementation matches the truncated value (first 6 chars of the
  jellyfish output), NOT the jellyfish value.

- **Double Metaphone:** No known variant divergence; direct equality assertion.

- **MRA:** No known variant divergence; direct equality assertion.

Per-entry schema for divergent entries:

```json
{
  "algorithm": "NYSIIS",
  "input": "Catherine",
  "code": "CATARA",
  "variant_divergence": true,
  "divergent_jellyfish_value": "CATARAN"
}
```

The Go loader test `TestPhonetic_CrossValidation` checks `e.VariantDivergence`
and emits a diagnostic message distinguishing Knuth-expected vs. jellyfish
divergent values in failure output.

### Go entry point

```bash
go test -run TestPhonetic_CrossValidation -race ./...
```

Sub-tests are organised by algorithm: `/Soundex`, `/DoubleMetaphone`,
`/NYSIIS`, `/MRA`. Each algorithm's sub-tests are activated by the
corresponding plan (07-01 through 07-04); earlier plans stub the not-yet-
implemented sub-tests with `t.Skip`.

### Regeneration

Developer command (requires both pins installed):

```bash
python3 -m pip install --user jellyfish==1.2.1 Metaphone==0.6
make regen-phonetic-cross-validation
```

The Makefile target invokes `python3 scripts/gen-phonetic-cross-validation.py`,
which writes the corpus to
`testdata/cross-validation/phonetic/vectors.json` with `_metadata` fields for
both pins, the Python version, and the regeneration timestamp. Only the
timestamp varies between runs; scoring fields are byte-stable.

CI does NOT regenerate the corpus. The committed JSON is the fixture.
