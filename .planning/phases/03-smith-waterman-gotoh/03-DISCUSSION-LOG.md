---
phase: 03-smith-waterman-gotoh
phase_number: 3
date: 2026-05-14
discussion_mode: discuss
areas_presented: 4
areas_discussed: 4
turns_total: 11
---

# Phase 3: Smith-Waterman-Gotoh — Discussion Log

This is the raw record of the discuss-phase session. CONTEXT.md is the
canonical artifact for downstream agents — this log exists for human
reference (audits, retrospectives).

## Gray Area Selection

**Domain stated:** Smith-Waterman-Gotoh local alignment with affine gap
penalty; one algorithm, one requirement (CHAR-08), isolated phase due to
the documented Gotoh 1982 erratum.

**Spec-locked items declared up-front (no discussion needed):**
- Public API names: `SmithWatermanGotohScore`, `SmithWatermanGotohScoreRunes`, `SmithWatermanGotohScoreWithParams`
- `SWGParams` struct shape: `{Match, Mismatch, GapOpen, GapExtend float64}`
- Default params: 1.0 / -1.0 / -1.5 / -0.5
- Score normalisation: `clamp(best / min(len(a), len(b)), 0, 1)`
- Cross-validation against EMBOSS or biopython REQUIRED (roadmap)

**Carry-forward from Phase 2:** AlgoID slot already declared, ASCII gate idiom locked, file-by-file pattern locked, golden merge pattern, BDD per-algorithm feature pattern.

**Gray areas presented (all 4 selected by user for discussion):**
- A. Cross-validation evidence path
- B. DP space-optimisation policy
- C. SWGParams API ergonomics
- D. Behaviour on negative raw alignment scores

---

## Area A: Cross-validation evidence path

### Q1: Which independent SWG implementation should we cross-validate against?
**Options presented:** biopython PairwiseAligner / EMBOSS water / both / biopython now + EMBOSS deferred.
**User chose:** biopython `Bio.Align.PairwiseAligner` (the Recommended option).
**Notes:** Pip-install footprint, easier scripting. Spec mentioned the deprecated `pairwise2.align.localxs`; we use the actively-supported `PairwiseAligner` instead.

### Q2: Where should the cross-validation evidence live?
**Options presented:** committed JSON + generator script / vectors only / PR description only / inline Go test constants.
**User chose:** committed JSON vectors + generator script (Recommended).
**Notes:** `scripts/gen-swg-cross-validation.py` → `testdata/cross-validation/swg/vectors.json`. `TestSWG_CrossValidation` reads the JSON at test time (no Python required at test time).

### Q3: Score-comparison tolerance — exact bytes or epsilon?
**Options presented:** epsilon 1e-9 / exact bit equality / epsilon 1e-12.
**User chose:** epsilon 1e-9 (Recommended).
**Notes:** Matches `cross_algorithm_consistency_test.go` convention.

### Q4: Should the cross-validation gate be part of default test suite or separate target?
**Options presented:** default `go test` + `make check` / opt-in target only / default + separate regen target.
**User chose:** default `go test` + `make check` (Recommended).
**Notes:** Separate `make regen-swg-cross-validation` developer-discretion target captured in CONTEXT.md §1.

---

## Area B: DP space-optimisation policy

### Q5: How should SWG-Gotoh be implemented — paper-faithful or space-optimised?
**Options presented:** two-row from day 1 / full O(mn) tables (PERF-03 deferral precedent) / hybrid.
**User chose:** two-row optimised from day 1 (Recommended).
**Notes:** Phase 2's DL-Full deferral was justified by DL-Full's `da` auxiliary map; SWG has no such complication. Biopython cross-validation catches transcription bugs.

### Q6: Allocation budget for SWG — same as Phase 2?
**Options presented:** same as Phase 2 (0 on ASCII Short) / looser (1 alloc/op acceptable) / determine empirically.
**User chose:** same as Phase 2 (Recommended).
**Notes:** Stack buffer `[3 * 2 * (maxStackInputLen+1)]float64` = 3120 bytes. Bench labels mirror Phase 2.

---

## Area C: SWGParams API ergonomics

### Q7: SWGParams construction — naked struct vs constructor vs both?
**Options presented:** naked struct + `SWGDefaultParams` exported var / naked struct only / `NewSWGParams()` constructor.
**User chose:** `NewSWGParams()` constructor returning defaults.
**Notes:** Diverged from the Recommended option (naked struct + exported var). Constructor is more explicit and matches `mask`'s pattern. Avoids the "is SWGDefaultParams read-only?" ambiguity.

### Q8: Should SWGParams validate inputs — and where?
**Options presented:** no validation + document ranges / validate in NewSWGParams only / validate everywhere (panic on invalid).
**User chose:** no validation + document expected ranges in godoc (Recommended).
**Notes:** Matches §5.11 pure-function discipline. Scorer (Phase 8) may validate at composition time.

### Q9: How should `SmithWatermanGotohScore(a, b)` (no params) relate to `NewSWGParams()`?
**Options presented:** delegate inline / constant-fold separate path / package-level frozen default.
**User chose:** ScoreWithParams kernel + no-params form calls it with `NewSWGParams()` inline (Recommended).
**Notes:** Cleanest factoring, single source of truth for defaults.

---

## Area D: Behaviour on negative raw alignment scores

### Q10: Should we expose a raw (unclamped, unnormalised) alignment score?
**Options presented:** no (spec only) / yes — add `SmithWatermanGotohRawScore` / defer to future phase.
**User chose:** yes — add `SmithWatermanGotohRawScore` (advanced consumers).
**Notes:** Diverged from the Recommended option (which was to stay strictly within spec). This is an intentional scope expansion: 3 new public functions beyond what `docs/requirements.md` §7.1.8 specifies. Phase 3 deliverables include updating the requirements doc accordingly. Logged in CONTEXT.md §4 as a deliberate decision; api-ergonomics-reviewer should review the surface.

### Q11: Signature for `SmithWatermanGotohRawScore` — rune variant, params variant, or both?
**Options presented:** all three matching the normalised set / only params variant / no-params byte + params variant.
**User chose:** all three matching the normalised set (Recommended).
**Notes:** Symmetric API surface. 3 new public functions total.

### Q12: How should the normalisation clamping be described in godoc — explicit or implicit?
**Options presented:** explicit warning + cross-ref to Raw / implicit "[0, 1]" only / show formula + let readers infer.
**User chose:** explicit warning + cross-reference to Raw (Recommended).
**Notes:** Godoc on the normalised functions explicitly states the clamp behaviour and points to `SmithWatermanGotohRawScore*` for advanced consumers.

---

## Deferred Ideas (captured during discussion)

- **EMBOSS `water` second cross-validation source** — deferred; biopython alone for v1.0.
- **`SmithWatermanGotohAlignment(a, b) Alignment`** returning actual alignment trace — out of scope per spec; future v1.x if demand surfaces.
- **`SmithWatermanGotohDistance` function** — SWG is not a metric distance; not needed.
- **CI installation of biopython** for re-verification — JSON corpus is the verification fixture; defer CI re-verification to a later workflow if needed.

## Claude's Discretion (decided without explicit user input)

- `Bio.Align.PairwiseAligner` API choice over the deprecated `pairwise2.align.localxs` from spec.
- Stack buffer size formula `[(maxStackInputLen+1) * 6]float64` = 3120 bytes.
- Property test naming (`TestProp_SwG*`, `TestProp_SwGScoreRunes_Symmetric`) — mechanical extension of Phase 2 patterns.
- The cross-algorithm-consistency test extension (SWG vs Levenshtein on substring-containment) — a natural addition given Phase 2's `cross_algorithm_consistency_test.go` precedent.
- Plan decomposition (3-plan single-wave structure with optional 2+3 collapse) — surfaced as guidance, not locked; planner refines.

---

_Discussion duration: ~10 min._
_Session: 2026-05-14, gsd-discuss-phase mode="discuss"._
