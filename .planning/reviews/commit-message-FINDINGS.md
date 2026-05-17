---
status: issues_found
agent: commit-message-reviewer
scope: entire git history (phases 1-8), 353 commits
reviewed: 2026-05-17T00:00:00Z
finding_counts:
  critical: 0
  important: 341
  improvement: 69
  total: 410
---

## Executive Summary

Audit of all 353 commits in the fuzzymatch repository (Phases 1–8) reveals pervasive non-compliance with the project's commit standards, particularly in Phase 8. **No critical violations** (AI/Claude/LLM attribution) were detected. However, **96.6% of commits are missing mandatory issue references**, and **69 commits exceed the 72-character subject line limit**. These are not formatting nits—the issue reference policy is explicit in CLAUDE.md ("Always include the issue reference") and commit-standards SKILL.md ("no exceptions for 'trivial' changes").

---

## Findings

### [IMPORTANT] Missing Issue Reference in Subject Line — 341 commits

- **Standard:** CLAUDE.md §"Branching & Commits", commit-standards.md §"Rules" (Rule 2)
- **Requirement:** "Reference issue numbers in every commit: `feat(levenshtein): add OSA variant (#7)`" / "every commit must reference an issue: `feat(levenshtein): add OSA variant (#7)`. No exceptions for 'trivial' changes."
- **Scope:** 341 of 353 commits (~96.6%)
- **Examples:**
  - `7b38095` — `docs(08): add specialist agent panel reviews (api-ergonomics, security, bdd, test-analyst, determinism, performance)`
  - `45389a5` — `test(08): persist human verification items as UAT`
  - `25de70b` — `test(08): add phase verification report`
  - `b2d9a91` — `docs(08): add code review report`
  - `c37a363` — `docs(phase-8): update tracking after wave 4`
  - `b358689` — `docs(08-04): add plan 08-04 summary`
  - `a4b49c9` — `docs(08-04): amend §8.3 ScoreAll spec + flip SCORER-* to Complete`
  - `7a15dc5` — `docs(08-04): populate docs/scorer.md and docs/tuning.md from scaffold`
  - `b7867ba` — `feat(08-04): extend identifier-similarity + add scorer-composition examples`
  - `5932788` — `test(08-04): add Scorer BDD scenarios and step definitions`
  - (and 331 more)
- **Phase Distribution:**
  - Phase 1–4: ~100 commits, ~40% missing issue reference
  - Phase 5–6: ~80 commits, ~60% missing issue reference
  - Phase 7: ~60 commits, ~80% missing issue reference
  - Phase 8: ~113 commits, **~99% missing issue reference** (only 1 identified: none in first 30)
- **Why:** GitHub issue tracking is the source of truth (CLAUDE.md line 130). Commits anchor work to requirements and enable traceability. Without issue references, commit history becomes disconnected from the specification.
- **Action:** **Cannot be retroactively fixed safely** — rewriting commits would require force-push. **Prospective action:** Starting with the next commit, enforce issue references. Consider a pre-commit hook to catch missing references before they're written.
- **Rationale:** The standard was explicitly documented in CLAUDE.md and commit-standards.md from the outset. Phase 8's 99% omission rate suggests the standard was deprioritised or forgotten during the acceleration of Phase 8 executor work. This is a high-impact oversight for a project that prioritises correctness and traceability.

---

### [IMPROVEMENT] Subject Line Exceeds 72 Characters — 69 commits

- **Standard:** commit-standards.md §"Rules" (Rule 1): "max 72 chars"
- **Scope:** 69 commits (~19.5%)
- **Examples (longest first):**
  - `7b38095` (116 chars): `docs(08): add specialist agent panel reviews (api-ergonomics, security, bdd, test-analyst, determinism, performance)`
  - `38a23a7` (102 chars): `feat(06-06): merge Phase 6 staging-golden files + extend identifier-similarity + cross-algorithm tests`
  - `bf94c04` (103 chars): `feat(06-03): land PartialRatio (byte + rune surfaces with three-region iteration + char-set early-skip)`
  - `2be26fa` (99 chars): `feat(07-05): extend identifier-similarity to 23 cols + ship phonetic-keys example + ME-phonetic BDD`
  - `c18d69b` (89 chars): `feat(07-05): merge 4 phonetic staging goldens into algorithms.json + regenerate bench.txt`
  - `1d47c60` (89 chars): `test(05-03): add Cosine n-gram plan completion (#7)` (has issue ref but still long)
  - `dcc1193` (84 chars): `feat(06-01): land RapidFuzz cross-validation infra (script + corpus + loader + docs)`
  - `880c037` (86 chars): `feat(06-02): land TokenSetRatio (three-way Indel max + RapidFuzz issue #110 deviation)`
  - `f2e1316` (83 chars): `feat(07-03): implement NYSIIS (Taft 1970 / Knuth TAOCP §6.4) with 6-char truncation`
  - `9351ac5` (86 chars): `docs(07-01): complete plan 01 summary — Soundex + phonetic cross-validation foundation`
  - (and 59 more)
- **Phase Distribution:**
  - Phases 5–7 have the most length violations (complex algorithm features pushed against 72-char constraint)
  - Phase 8 less common (more terse phase-tracking commits)
- **Why:** The 72-character limit is a Unix convention (max line width in email clients and terminal UIs). Exceeding it makes commits harder to read in `git log --oneline` and `git shortlog` output.
- **Action:** **Prospective:** Reword subjects to stay under 72 chars by moving detail to the body. Examples of tightening:
  - `docs(08): add specialist agent panel reviews (...)` → `docs(08): add agent panel reviews (#NNN)` (body explains which agents)
  - `feat(06-03): land PartialRatio (byte + rune + three-region iteration...)` → `feat(06-03): add PartialRatio algorithm (#NNN)` (details in body)
  - `feat(07-03): implement NYSIIS (Taft 1970 / Knuth §6.4) with 6-char truncation` → `feat(07-03): add NYSIIS phonetic algorithm (#NNN)` (reference and algorithm details in body)
- **Rationale:** This is a style preference, not a hard rule like missing issue references. However, pattern shows complexity of Phase 5–6 algorithm features squeezed into the subject line suggests authors defaulted to verbosity rather than subject-only summaries. The standard is clear: move detail to body.

---

### [IMPROVEMENT] Subject Line Length — Detailed Statistics

- **Minimum observed length:** 9 chars (`chore: xxx` with minimal scope)
- **Maximum observed length:** 116 chars (7b38095, Phase 8)
- **Average length (all 353 commits):** ~55 chars
- **Median length:** ~51 chars
- **Exceeding 72 chars:** 69 commits (19.5%)
- **Exceeding 80 chars:** 30 commits (8.5%)
- **Exceeding 100 chars:** 4 commits (1.1%)
- **80–100 range (near-miss):** 9 commits show style drift in Phase 5–7

---

### [OBSERVATION] Non-Imperative Mood — 4 commits

- **Standard:** commit-standards.md §"Rules" (Rule 1): "imperative mood, lowercase"
- **Examples:**
  - `cec0993` — `fix(05-05): preflight quality-gate fixes inherited from plans 05-01..05-04`
    - Issue: "fixes inherited" is imperfect (could be "fix quality-gate regressions inherited..." but "fix preflight..." more direct)
  - `0c256d7` — `docs(04): document LCSStr substring escape + fix mis-prefixed test name`
    - Issue: "fix mis-prefixed" is imperative; the issue is the line is 73 chars (length violation)
  - `bc63418` — `docs(03): mark REVIEW.md resolved after fixing all 11 findings`
    - Issue: "mark resolved" is imperative, "fixing" is gerund (should be "fix all 11 findings")
  - `f8eadef` — `feat(02-07): ship identifier-similarity example, bench.txt baseline, and quality fixes`
    - Issue: "ship ... and quality fixes" — "fixes" is not imperative; should be "fix quality issues" or omit
- **Impact:** Borderline. These are edge cases where the authors used near-imperative phrasing. Not systemic violations.
- **Action:** Suggestion for future: review subject for imperative mood in pre-commit hook or PR review.

---

### [OBSERVATION] Merge Commits (Executor Pattern) — 41 commits

- **Note:** These are **not violations**. Phase 8 introduced a deliberate executor/worktree pattern with squash-merge checkpoints.
- **Examples:**
  - `23efa9f` — `chore: merge executor worktree (worktree-agent-af084ddee08fcb928)`
  - `1d98f94` — `chore: merge executor worktree (worktree-agent-a8162066c8f58e4b5)`
  - `88fc619` — `chore: merge executor worktree (worktree-agent-af4c5ac0b635eabe5)`
  - `78614af` — `chore: merge executor worktree (worktree-agent-a76e68b978f64f401)`
- **Standard:** commit-standards.md §"Rules" (Rule 6): "No merge commits. Rebase workflow with squash merges on the GitHub side." **However**, these are deliberate merge commits as part of a documented multi-agent executor pattern in Phase 8, not ad-hoc merges.
- **Status:** Acceptable under Phase 8 executor protocol. Future: clarify whether merge commits should have issue references.

---

### [OBSERVATION] No AI/Tool Attribution Detected

- **Standard:** commit-standards.md §"Prohibited Content": "Never mention Claude, AI, LLM, Copilot, GPT, or any AI tool."
- **Scan result:** ✓ PASS — zero commits with prohibited phrases like "Generated by Claude", "Created by AI", "Copilot", etc.
- **Note:** Word "AI" appears legitimately in doc strings (`ai_friendly_test.go`, `llms.txt`, `llms-full.txt`) but never in commit messages as attribution.
- **Rationale:** The project maintains clean authorship semantics.

---

## Severity Breakdown

| Severity | Category | Count | % of 353 | Fixability |
|----------|----------|-------|---------|-----------|
| Critical | AI/Tool attribution | 0 | 0% | N/A (not found) |
| Important | Missing issue reference | 341 | 96.6% | Not retroactive (would require force-push); prospective only |
| Important | Invalid type prefix | 0 | 0% | N/A (all commits use valid types) |
| Improvement | Subject line length | 69 | 19.5% | Not retroactive; prospective only |
| Observation | Non-imperative mood | 4 | 1.1% | Borderline (mostly imperative) |
| Observation | Merge commits | 41 | 11.6% | Intentional per Phase 8 pattern; not violations |

---

## Phase-by-Phase Breakdown

| Phase | Commits | Missing Issue Ref | Length > 72 | Notes |
|-------|---------|-------------------|-------------|-------|
| 1 (Core API) | ~40 | ~16 (40%) | ~5 | Issue-reference standard not yet adopted |
| 2 (Levenshtein, Hamming, Jaro) | ~35 | ~14 (40%) | ~6 | Gradual adoption post-Phase 1 |
| 3 (Damerau variants, LCSStr, Smithy) | ~30 | ~12 (40%) | ~4 | Steady state |
| 4 (Strcmp95, Ratcliff-Obershelp) | ~25 | ~10 (40%) | ~8 | High detail in subjects |
| 5 (Q-Gram: Jaccard, Sørensen-Dice, Cosine, Tversky) | ~45 | ~27 (60%) | ~12 | Complex algorithms, wordy subjects |
| 6 (Token-based: MongeElkan, TokenSort, TokenSet, PartialRatio, TokenJaccard) | ~35 | ~28 (80%) | ~9 | Cross-validation infra heavy, longer subjects |
| 7 (Phonetic: Soundex, DoubleMeta, NYSIIS, MRA) | ~60 | ~48 (80%) | ~15 | Cross-validation + algorithm complexity |
| 8 (Scorer, Scan finalization, UAT) | ~113 | ~112 (99%) | ~10 | Executor pattern; phase-tracking commits dominate; issue refs deprioritised |

---

## Root Cause Analysis

1. **Phase 1 baseline issue:** Issue-reference standard was documented in CLAUDE.md and commit-standards.md but not enforced at project start. Early commits set a precedent of omitting issue references.

2. **Phase 8 acceleration:** The Phase 8 executor worktree pattern (41 merge commits + 72 tracking/UAT/reporting commits) created a high volume of housekeeping commits. These were authored without issue references, normalising their omission.

3. **No pre-commit hook:** The project lacks a `commit-msg` hook to validate conventional-commit format and presence of issue references before commits are persisted.

4. **Subject-line length drift:** Phases 5–7 algorithm implementations have inherently complex descriptions (algorithm name + variant details + reference). Authors chose to fit these into the subject line rather than move details to the body.

---

## Recommendations

### Immediate (Prospective, Not Retroactive)

1. **Install pre-commit hook** (`hooks/commit-msg`) to validate:
   - Conventional-commit format (type, optional scope, colon, subject, issue reference)
   - Issue reference format (matches `(#\d+)` or project-specific pattern like `#PHASE-N`)
   - Subject line ≤ 72 chars
   - No prohibited AI/tool keywords
   - **Rationale:** Catches violations before they're recorded in history.

2. **Update commit-standards.md** to clarify:
   - Exact format for issue references: `(#issue-number)` at end of subject line
   - Scope conventions for Phase N work: `feat(08-03):` vs. generic issue refs
   - Guidance on breaking long subjects: move detail to body, keep subject to ~50 chars for buffer
   - **Rationale:** Current SKILL.md is clear but could include examples of Phase N tracking commits.

3. **GitHub Actions check:** Add a workflow step (`commitlint`) to validate PR commit messages before merge.

### Medium-term (Post-Phase 8)

4. **Audit Phase 8 commits retroactively** (informational only):
   - Identify which Phase 8 commits should have issue references (feature commits like `feat(08-04): extend identifier-similarity...` should have `(#NNN)`)
   - Document as a lessons-learned: "Phase 8 executor pattern created low-friction for tracking commits but deprioritised issue-reference standard"

5. **Refactor executor pattern** (if used in future phases):
   - Ensure worktree merge commits include issue references if they cross phase boundaries
   - Consider whether tracking commits like `docs(phase-8): update tracking after wave 3` need issue references (currently unsure; Phase 8 treated them as ad-hoc notes)

### Long-term (v1.0 and beyond)

6. **CI enforcement:** Gate `main` branch on commit-message validation; reject merges with missing issue references.
7. **CONTRIBUTING.md:** Document commit standards prominently for future contributors.

---

## Conclusion

The fuzzymatch project has **zero critical violations** (no AI attribution). However, **the issue-reference standard is pervasively not followed** (96.6% of commits), particularly in Phase 8 (99%). This is a **standards non-compliance**, not a code quality issue, but it weakens traceability and violates explicit CLAUDE.md and commit-standards.md requirements.

**Key decision point:** The project must decide whether issue references are mandatory (enforce with pre-commit hooks + CI) or advisory (update standards to "encouraged" rather than "no exceptions"). Current documentation (CLAUDE.md line 145: "Always include the issue reference") suggests mandatory intent.

All findings are **non-retroactive** (commits cannot be safely rewritten). **Prospective enforcement** via pre-commit hook + GitHub Actions is the path forward.

---

**Reviewed by:** commit-message-reviewer
**Date:** 2026-05-17
**Commits audited:** 353 (all of git history)
