# Phase 4: Remaining Character & Gestalt - Discussion Log

> **Audit trail only.** Do not use as input to planning, research, or execution agents.
> Decisions are captured in CONTEXT.md — this log preserves the alternatives considered.

**Date:** 2026-05-14
**Phase:** 04-remaining-character-gestalt
**Areas discussed:** Ratcliff-Obershelp cross-validation strategy, Strcmp95 similar-character table sourcing, LCSStr normalisation formula choice, Plan decomposition

---

## Ratcliff-Obershelp cross-validation strategy

### Q1: How should Ratcliff-Obershelp cross-validate against difflib?

| Option | Description | Selected |
|--------|-------------|----------|
| Python script + committed JSON corpus, mirror Phase 3 | scripts/gen-ratcliff-obershelp-cross-validation.py invokes difflib.SequenceMatcher(autojunk=False).ratio(); commits testdata/cross-validation/ratcliff-obershelp/vectors.json. TestRatcliffObershelp_CrossValidation runs from default go test ./... with 1e-9 tolerance. | ✓ |
| Hand-curated reference vectors only | Pin scores in unit tests against Dr. Dobb's 1988 paper examples + a handful of hand-computed values. No external script. | |
| Both: Dr. Dobb's vectors in unit tests + difflib JSON corpus | Dr. Dobb's pinning in unit tests + difflib JSON corpus. ~2x artefacts but airtight. | |
| stdlib difflib via os/exec | No committed corpus. Requires Python at CI test time. | |

**User's choice:** Python script + committed JSON corpus, mirror Phase 3.
**Notes:** Locks the algorithm-correctness-reviewer gate at the same shape as SWG. autojunk=False is the critical qualifier — it must appear in the script and in the godoc on `RatcliffObershelpScore`.

### Q2: What corpus categories should the difflib reference vectors cover?

(Multi-select question with options: standard edge cases / Dr. Dobb's examples / autojunk-sensitive case / substring+partial+unicode. Question was asked via AskUserQuestion twice and returned empty both times; fell back to plain-text numbered list per workflow rule.)

| Option | Description | Selected |
|--------|-------------|----------|
| Standard edge cases (identity, both-empty, one-empty, no-overlap) | Phase 3 minimum. | ✓ |
| Dr. Dobb's 1988 paper examples | Canonical primary-source vectors. | ✓ |
| autojunk-sensitive case (200+ char string) | Proves autojunk=False is correctly disabled. | ✓ |
| Substring + partial-match + unicode (café/cafe) | Mid-similarity + multi-byte UTF-8 rune-path coverage. | ✓ |

**User's choice:** "All four, please."
**Notes:** Corpus will be ~15–18 entries covering every required category.

---

## Strcmp95 similar-character table sourcing

### Q1: Where does the similar-character table get its values from?

| Option | Description | Selected |
|--------|-------------|----------|
| Winkler 1994 TR-2, transcribed by hand into a var-level table | 36 letter-pair values transcribed from the published paper into `var strcmp95SimilarChars`. Census Bureau strcmp95.c consulted only for canonical reference vectors. Cleanest provenance. | ✓ |
| Winkler 1994 + Census strcmp95.c cross-checked | Same plus a Census Bureau confirmation block. Adds a second source to cite. | |
| Embed full Census Bureau table via documented script regeneration | Python script reads strcmp95.c's table and emits a Go literal. Busy-work for a static 36-entry table. | |

**User's choice:** Winkler 1994 TR-2, transcribed by hand into a var-level table.
**Notes:** Primary source is the paper. Census Bureau code is reference-vector-only. OpenRefine Java is tie-breaking-only. Locks the licensing-discipline trap Phase 3 avoided.

### Q2: ASCII-only scope?

| Option | Description | Selected |
|--------|-------------|----------|
| ASCII-only byte path; no *Runes variant | Strcmp95 is record-linkage; table is letter-pair-keyed. | |
| Byte path AND *Runes variant | Mirror SWG / Levenshtein. *Runes adds public surface for negligible semantic gain. | |
| Byte path; document Unicode users should Normalise first | Single byte path; godoc directs Unicode users to Normalise (Phase 1) for diacritic folding. | ✓ |

**User's choice:** Byte path; document Unicode users should Normalise first.
**Notes:** Connects to Phase 1's foundation. The godoc directive is the key consumer hand-off.

### Q3: Adjustment-rule scope?

| Option | Description | Selected |
|--------|-------------|----------|
| Full Winkler 1994 spec — all four adjustments | Faithful to the paper; matches Census Bureau strcmp95.c byte-for-byte. | ✓ |
| Prefix boost + similar-char only | Shorter implementation but not byte-identical to the reference. Would fail canonical cross-validation. | |
| Full spec, but isolate adjustments behind a Strcmp95Params struct | Parameterise like SWGParams. Trade-off: more surface; the canonical "Strcmp95" is the full thing. | |

**User's choice:** Full Winkler 1994 spec — all four adjustments.
**Notes:** Strcmp95 is canonical; consumers expect the full algorithm. ~150-200 lines of Go. No params (Phase 8's Scorer handles algorithm weighting at the composite level).

---

## LCSStr normalisation formula

### Q1 (skipped — spec-pinned):

The original gray area was framed as "2·|lcs|/(|a|+|b|) vs |lcs|/max(|a|, |b|)". Inspection of `docs/requirements.md §7.1.9` revealed the formula is already SPEC-PINNED as `2 · len(lcs) / (len(a) + len(b))`. Not a gray area.

The spec also locks the public surface at **4 functions** (LongestCommonSubstring + Runes + LCSStrScore + Runes) — wider than other character algorithms because exposing the substring is a deliberate consumer-facing utility.

### Q2: Tie-break when multiple equal-length matches exist?

| Option | Description | Selected |
|--------|-------------|----------|
| First occurrence by left-most starting position in a | Deterministic. Standard textbook tie-break. | ✓ |
| Last occurrence by right-most starting position in a | Also deterministic but less intuitive. | |
| All ties returned as []string | Surface expansion; deferred to v1.x. | |
| Implementation-defined; document as undefined | Rejected — violates determinism guarantees. | |

**User's choice:** First occurrence by left-most starting position in a.
**Notes:** Documented in godoc; property test `TestProp_LongestCommonSubstring_LeftmostTieBreak` enforces it.

---

## Plan decomposition

### Q1: How should Phase 4 decompose into plans?

| Option | Description | Selected |
|--------|-------------|----------|
| 4 sequential plans: per-algo + finalisation | 04-01-strcmp95, 04-02-lcsstr, 04-03-ratcliff-obershelp (cross-val folded in), 04-04-finalisation. | |
| 5 sequential plans: per-algo + Ratcliff cross-val split + finalisation | 04-01-strcmp95, 04-02-lcsstr, 04-03-ratcliff-obershelp (impl only), 04-04-ratcliff-obershelp-cross-validation, 04-05-finalisation. Mirrors Phase 3's 3-plan SWG shape. | ✓ |
| 3 plans: all-three-algos-bundled + cross-val + finalisation | Bundling 3 algorithms in one plan is hard to review/revert. | |
| Parallel waves: bundle independent algorithms | Shared-file collisions (props_test.go etc.) make the merge tax dominate the parallelism win. | |

**User's choice:** 5 sequential plans: per-algo + Ratcliff cross-val split + finalisation.
**Notes:** Mirrors Phase 3's three-plan SWG shape exactly (impl / cross-val / finalisation), extended to cover the three Phase 4 algorithms. All plans run sequentially due to shared-file dependencies.

---

## Claude's Discretion

The planner has flexibility on:

- Whether Strcmp95 calls an internal Jaro helper or re-derives the match-flag arrays.
- Whether Ratcliff-Obershelp's recursion uses the language-native call stack or an explicit stack-based iterative implementation.
- Whether Ratcliff-Obershelp's "find longest common substring" inner step reuses LCSStr's internal helper or implements the substring search inline.
- The exact bench label conventions (matching Phase 2/3 prefix-numbering).
- The filename for Ratcliff-Obershelp (`ratcliff_obershelp.go` vs `ratcliffobershelp.go`).

## Deferred Ideas

- EMBOSS as a second cross-validation source for Strcmp95 — not applicable; Census Bureau strcmp95.c is canonical.
- A `Strcmp95Params` API — deferred to v1.x if consumer-driven use case surfaces.
- `LongestCommonSubstrings()` returning all tied-longest matches — spec commits to one; future scope question.
- CI installation of Python 3 for re-verification of the Ratcliff-Obershelp corpus — Python is already in CI runners; add re-verification workflow later if needed.
- Public API freeze for `LongestCommonSubstring` tie-break — leftmost is documented; changing post-v1.0 is breaking.
