---
phase: 02-core-character-algorithms-six
date: 2026-05-14
mode: discuss (default)
areas_discussed: [plan_decomposition, hamming_unequal_length, identifier_similarity_example]
areas_skipped: [skip_discussion_option_not_selected]
---

# Phase 2: Discussion Log

> Human-readable record of the discussion that produced `02-CONTEXT.md`.
> Not consumed by downstream agents.

## Pre-Discussion State

- **SPEC.md:** none (requirements come from `docs/requirements.md` §7.1.1–7.1.6 directly)
- **CONTEXT.md:** none (first discussion for Phase 2)
- **Plans:** none
- **Prior carry-forward:** Phase 1 complete; 5 deferred UAT items captured (don't block Phase 2)

## Gray Areas Presented

```
☐ Plan decomposition / sequencing
☐ Hamming unequal-length behavior
☐ identifier-similarity example scope
☐ Skip discussion — spec is enough
```

## User Selections

3 of 4 areas selected for discussion. Skip-option not selected.

---

## Area 1: Plan decomposition / sequencing

**Question:** How should Phase 2 plans be structured?

**Options presented:**

1. Levenshtein-first, then 5-way parallel (Recommended) — Wave 1: Levenshtein
   alone; Wave 2: Hamming + Jaro + JW + DL-OSA + DL-Full in parallel
2. 6 sequential plans (1 wave per algorithm, 6 waves)
3. 3 family-grouped plans (edit-distance family / Hamming / Jaro family)
4. All 6 parallel in wave 1

**User selected:** Option 1 — Levenshtein-first then 5-way parallel.

**Captured in CONTEXT.md:** `<decisions>` → Plan decomposition section with
expected shared-artefact merge designs called out for the planner.

---

## Area 2: Hamming unequal-length behaviour

**Question:** When inputs are unequal length, what should Hamming do?

**Options presented:**

1. Return distance = max-length, score = 0.0 (similarity metric across catalogue)
2. Return error / ErrInvalidInput (strict — honour Hamming 1950 definition)
3. Panic on unequal length (strictest)

**User selected:** Option 2 — strict; ErrInvalidInput on unequal length.

**Captured in CONTEXT.md:** `<decisions>` → Hamming unequal-length section.
Locked: `HammingDistance` returns `(int, error)` with `ErrInvalidInput` on
length mismatch; `HammingScore` silently returns 0.0. Signature divergence
from family pattern noted; api-ergonomics-reviewer retains veto.

---

## Area 3: identifier-similarity example scope

**Question:** What should `examples/identifier-similarity/` demonstrate?

**Options presented:**

1. All 6 algorithms side-by-side on db column pairs (Recommended) — runnable
   table comparing 6–10 hardcoded identifier pairs across all 6 algorithms
2. Single algorithm (Levenshtein), simple program — minimal hello-world
3. Real-world identifier-pair dedup workflow — stdin → group-by-threshold

**User selected:** Option 1 — all 6 algorithms side-by-side, plaintext table
output, meta-tested for stable byte-identical stdout.

**Captured in CONTEXT.md:** `<decisions>` → identifier-similarity scope
section with 7 representative identifier pairs spanning known similarity
shapes (case+separator drift, abbreviation, semantic synonyms, length-equal
content-different, opposite-meaning).

---

## Deferred Ideas

(Captured in CONTEXT.md `<deferred>` section)

- Damerau-Levenshtein Full ASCII fast path (v1.x perf polish)
- Hamming-in-Scorer construction-time warning (Phase 8)
- Other example programs from DX-05 (extract-demo, audit-field-similarity,
  schema-dedup) — later phases
- Levenshtein Ukkonen banding (v1.x perf polish)
- 0-alloc fast-path target on Normalise (Phase 1 follow-up, separate)

## Claude's Discretion

(Captured in CONTEXT.md `<open_questions>` section)

- Algorithm file naming convention
- `init()` vs `var` for dispatch registration (lean toward `var` per
  determinism-standards §13.5)
- ASCII fast path buffer size (default `[64]int`, tune via benchstat)
- Rune variant strategy (eager vs lazy decoding)
- BDD feature file structure (one per algorithm vs one shared)

---

## Process Notes

- Selected previews surfaced in user answers (option 1 + 1 + 1).
- No scope creep flagged.
- Discussion completed in single pass per memory note
  `feedback_decision_fatigue` (no nested follow-ups).
- Total user questions answered: 4 (1 multi-select gray-area picker +
  3 single-select area-specific questions).
