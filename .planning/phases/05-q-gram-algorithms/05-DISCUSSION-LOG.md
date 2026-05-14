# Phase 5: Q-gram Algorithms - Discussion Log

**Date:** 2026-05-14
**Phase:** 5-q-gram-algorithms
**Mode:** default (4 single-question turns)

> Human-reference audit log of the discuss-phase conversation. NOT consumed
> by downstream agents — see `05-CONTEXT.md` for decisions.

---

## Gray Areas Presented

Four phase-specific gray areas were identified after loading prior CONTEXT.md from phases 02 / 03 / 04, REQUIREMENTS.md, ROADMAP.md, and `docs/requirements.md §7.2`. All four were marked LOCKED by REQUIREMENTS.md / prior phases at the surface level, but each had genuinely open implementation choices the user could weigh in on.

The user selected all four for discussion.

---

## §1. Cosine determinism keystone

**Question:** What level of gating do you want on cross-platform byte-identical Cosine output?

**Options presented:**
- Golden + matrix only (Recommended)
- Golden + dedicated keystone fixture
- Golden + keystone + bit-exact property test

**Selected:** Golden + matrix only

**Rationale captured:** the Phase 1 `algorithms.json` + `make verify-determinism` pipeline already runs on the five-platform CI matrix and byte-diffs the JSON output. A separate `cosine_determinism.json` fixture or `math.Float64bits` hex pins would duplicate the existing gate without adding signal.

---

## §2. Q-gram extraction public API surface

**Question:** Internal-only q-gram extraction or expose a public `QGrams(s, n)` helper?

**Options presented:**
- Internal only (Recommended)
- Expose public `QGrams` + `QGramsRunes`
- Expose public `QGramMultiset`

**Selected:** Internal only

**Rationale captured:** matches `docs/requirements.md §7.2` phrasing "consumed by the four downstream algorithms via internal API". Promoting an unexported helper to public is a non-breaking change; the inverse is not. Defer to v1.x if real demand emerges from Scorer or Scan consumers writing custom q-gram-based algorithms.

---

## §3. Cosine intersection iteration order

**Question:** How should the dot-product loop iterate intersection keys?

**Options presented:**
- Sort intersection keys alphabetically (Recommended)
- Insertion-order key slice
- Sort by frequency, ties broken alphabetically

**Selected:** Sort intersection keys alphabetically

**Rationale captured:** Go map iteration is randomised; sorting keys before iteration is the clearest determinism argument and matches the same pattern that secured Phase 4's golden-determinism gate. Insertion-order key slice is marginally cheaper but adds API discipline burden across the q-gram tier.

---

## §4. Cross-validation reference vectors

**Question:** What's the cross-validation strategy for Jaccard / Dice / Cosine / Tversky?

**Options presented (first turn):**
- Hand-computed from academic formulas (Recommended)
- Python `textdistance` library
- Python `scipy.spatial.distance.cosine` for Cosine only

**Initially selected:** Hand-computed from academic formulas

**User reconsidered mid-discussion:** "If there's any comparable libraries we should run some comparison tests with, I don't think that's a bad idea. I'm quite happy to take that in and have tests that compare what we do versus the various Python libraries."

**Re-opened with refined options:**
- textdistance + scipy.cosine
- textdistance alone for all four
- textdistance + nltk + scipy
- Let me pick libraries per algorithm

**User clarification:** "First-choice: revert to hand-computed (my previous recommendation, option 1 in the previous question). Verify CONTEXT.md captures it correctly. Fallback if libraries are locked in: option 2 (textdistance alone). Plus 3–5 hand-verified edge cases for Cosine recorded in test comments."

**Final selection:** Hand-computed from primary sources (no external Python toolchain). Strengthened from "at least one Cosine reference vector" to "3-5 hand-verified Cosine pairs with full float64 precision derivation in test comments" — Cosine carries the cross-validation density that would otherwise come from an external library.

**Fallback path** (NOT active): if an algorithm-correctness reviewer surfaces a specific concern, fallback is `textdistance` (single pip dep, covers all four algorithms) following the Phase 3 SWG / Phase 4 RO generator pattern.

---

## Carry-Forward Decisions (LOCKED — not re-discussed)

The following patterns were carried forward from Phase 2 / 3 / 4 CONTEXT.md files and explicitly NOT re-litigated:

- File-by-file structure: `<algo>.go` + `dispatch_<algo>.go` + `<algo>_test.go` + `<algo>_bench_test.go` + `<algo>_fuzz_test.go`, plus shared `q_gram.go`
- Byte path + Rune path for every algorithm
- Property tests via `testing/quick` (five invariants × two surfaces)
- Direct-call invalid args → panic; Scorer invalid args → error at construction
- No map iteration on output path (DET-03)
- No transcendental floats (DET-06) — only `math.Sqrt`, no `math.Pow`/`Log`/`Exp`/`FMA`
- AlgoID slots already reserved in `algoid.go` lines 109-128 — planner wires dispatch files but does NOT renumber
- Staging golden → finalisation merge pattern (Phase 2 onward)
- identifier-similarity example gets 4 new columns (7 → 14) during finalisation
- BDD scenario per algorithm in `tests/bdd/features/`

---

## Scope Creep — Redirected

No scope creep raised during this discussion. The user kept the discussion within the four locked-by-requirements gray areas.

---

## Deferred Ideas

- Public q-gram extraction helper (`QGrams` / `QGramsRunes`) — defer to v1.x as additive change if real demand emerges
- Cosine bit-exact property test via big.Float reference — redundant with existing CI matrix gate
- `scripts/gen-qgram-cross-validation.py` against `textdistance` — fallback path only; not active in this phase
- n-gram size validation at AlgoID-table level — no place in the `(a, b string) float64` dispatch signature; happens at Scorer-option layer in Phase 8

---

## Claude's Discretion

Items the planner (`gsd-planner`) chooses without re-asking the user:

- Exact internal extractor signature (`extractQGrams` vs `qgramBag` vs other)
- Wave decomposition of plans 05-01 through 05-04 (the four algorithms are independent; `q_gram.go` ships in plan 05-01 alongside Jaccard since Jaccard is the simplest consumer)
- Tversky dispatch fallback parameters (recommendation: `α = β = 1.0` Jaccard fallback)
- Stack-buffer fast path for q-gram extraction on short ASCII inputs (likely not worth it — map allocation dominates)
- Exact number of staging-golden entries per algorithm (8-12 per Phase 2/3/4 norm)

---

*End of discussion log. See `05-CONTEXT.md` for the canonical decision record consumed by downstream agents.*
