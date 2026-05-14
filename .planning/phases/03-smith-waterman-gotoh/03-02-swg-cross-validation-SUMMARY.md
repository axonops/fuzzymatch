---
phase: 03-smith-waterman-gotoh
plan: 02
subsystem: algorithm-catalogue
tags: [smith-waterman-gotoh, swg, biopython, cross-validation, gotoh-erratum-gate, reference-vectors, python-script, makefile, contributing-doc]

# Dependency graph
requires:
  - phase: 03-smith-waterman-gotoh
    plan: 01
    provides: "Public surface SmithWatermanGotohScoreWithParams + SWGParams + NewSWGParams (consumed by TestSWG_CrossValidation); existing swg_test.go that this plan APPENDS to; staging-golden corpus that already pins the one_long_gap_canary at norm=0.5 (matches the biopython reference exactly)."
provides:
  - "scripts/gen-swg-cross-validation.py — developer-only Python generator using biopython 1.87's Bio.Align.PairwiseAligner (mode='local'); deterministic byte-stable JSON output (zero-diff on re-runs); BSD-3-Clause attribution in docstring; 16 cases in CASES list covering all 8 mandatory CONTEXT.md §1 categories plus 8 additional spanning unicode / single-char / partial-middle-match / all-mismatch / near-identical / substring-at-end."
  - "testdata/cross-validation/swg/vectors.json — committed corpus (4577 bytes, 16 entries, biopython_version=1.87); canonical-byte-form (2-space indent, trailing LF, no BOM); field order LOCKED at name, a, b, params, biopython_score, biopython_normalised."
  - "TestSWG_CrossValidation in swg_test.go — appended (76 lines); reads the JSON, asserts |our_normalised - biopython_normalised| <= 1e-9 with per-entry t.Run sub-tests; failure messages include the biopython version from the corpus header for one-step triage; ZERO in-Go normalisation logic (CONTEXT.md §1 decision (a))."
  - "Makefile regen-swg-cross-validation target — developer-only, NOT in `make check`; tolerant `command -v python3` gate; documented in CONTRIBUTING.md to satisfy TestMakefile_TargetsDocumentedInContributing meta-gate."
  - "Gotoh-erratum cross-validation gate CLOSED: the one_long_gap_canary entry (abc________def vs abcdef) records biopython_score=3.0 / biopython_normalised=0.5; our implementation matches with delta=0.00e+00 (exact float equality). algorithm-correctness-reviewer sign-off path is now wired."
affects: [03-03-finalisation]

# Tech tracking
tech-stack:
  added: ["biopython 1.87 (BSD-3-Clause, developer-only via `python3 -m pip install --user biopython`; NOT a runtime or CI dependency — script is run on-demand and the resulting JSON is committed)"]
  patterns:
    - "First cross-validation corpus in the repo — pins agreement with an external reference implementation (orthogonal to testdata/golden/_staging/ which pins our own byte-stability)."
    - "First Python file in scripts/ — existing scripts are bash; the Python script's docstring carries an Apache-2.0 `#`-comment header AND the biopython BSD-3-Clause attribution (verify-license-headers.sh targets .go only, so the docstring is the in-file licence record)."
    - "Schema-locked JSON corpus — `{version, biopython_version, entries: [{name, a, b, params, biopython_score, biopython_normalised}]}` with deterministic field order (Python 3.7+ json.dump sort_keys=False)."
    - "Both-empty / one-empty Python-side short-circuits — match Go-side identity (return 1.0 for both-empty) and one-empty (return 0.0) conventions BEFORE invoking the aligner, sidestepping biopython's implementation-defined PairwiseAligner.score(\"\", \"\") behaviour."
    - "Normalisation owned by the Python script — biopython_normalised = clamp(raw / min(len(a), len(b)), 0, 1); the Go test compares against the normalised value with zero in-Go normalisation logic (CONTEXT.md §1)."
    - "Tolerant-if-tool-not-installed Makefile idiom — mirrors the `security` target's `command -v` gate for the python3 prerequisite."

key-files:
  created:
    - "scripts/gen-swg-cross-validation.py — 145 lines; Apache-2.0 + biopython BSD-3-Clause attribution; CASES list (16 entries); DEFAULT_PARAMS (Match=1.0, Mismatch=-1.0, GapOpen=-1.5, GapExtend=-0.5); score_case + main functions; trailing newline + 2-space indent + sort_keys=False."
    - "testdata/cross-validation/swg/vectors.json — 4577 bytes; 16 entries; biopython_version=1.87; canonical-byte-form."
  modified:
    - "swg_test.go — appended TestSWG_CrossValidation (76 lines including godoc); added three stdlib imports (encoding/json, os, path/filepath) to the existing import block; no testify; per-entry t.Run sub-tests; failure messages include biopython version for triage."
    - "Makefile — added regen-swg-cross-validation target (between verify-license-headers and release-check); appended target name to .PHONY list."
    - "CONTRIBUTING.md — added a `make regen-swg-cross-validation` bullet to the \"Make Targets\" section (required by TestMakefile_TargetsDocumentedInContributing — auto-fixed during Task 3 as a Rule 3 blocking issue)."

key-decisions:
  - "biopython version pinned in corpus header: 1.87 (current PyPI release as of 2026-05-14; script records `Bio.__version__` automatically — future bumps surface in PR diff as a version-string change)."
  - "Total entry count: 16 (target range 15-20 satisfied). 8 mandatory CONTEXT.md §1 categories + 8 additional (single_char_match, single_char_diff, partial_middle_match, all_mismatch, unicode_ascii_only, identity_long, near_identical, substring_at_end)."
  - "Python script's license treatment: BOTH the Apache-2.0 `#`-comment block (lines 2-14) AND the docstring biopython BSD-3-Clause attribution (lines 19-26). Even though scripts/verify-license-headers.sh only checks .go files, the Apache header is a defensive habit + the BSD-3-Clause attribution is the load-bearing licensing record per .claude/skills/algorithm-licensing-standards."
  - "TestSWG_CrossValidation runtime: ~0.26s wall (full test invocation including go test infrastructure overhead); the 16 sub-tests themselves are sub-microsecond."
  - "TDD RED-phase verification: temporarily poisoned the one_long_gap_canary biopython_normalised value (0.5 → 0.99) and confirmed the test correctly reported delta=4.90e-01 with full context (entry name, both inputs, SWGParams, our_score, biopython_normalised, delta, tolerance, biopython version). Restored corpus byte-identically before committing."

patterns-established:
  - "Cross-validation corpus pattern — committed JSON fixture + on-demand Python regenerator + zero-Python-at-CI verification path. Applicable to any future algorithm that needs independent-implementation cross-validation."
  - "Script-owned normalisation reference — when CI verifies against an external reference, the script owns the normalisation so the Go test has zero implementation logic and captures both correctness gates simultaneously (kernel arithmetic + normalisation formula)."
  - "Sub-test naming via t.Run(e.Name) — individual entry failures are visible in test output without truncation; matches the existing staging-golden per-entry naming."

requirements-completed: [CHAR-08]

# Metrics
duration: 8min
completed: 2026-05-14
---

# Phase 3 Plan 02: Smith-Waterman-Gotoh Cross-Validation Summary

**Closes the algorithm-correctness-reviewer cross-validation gate for Phase 3 by shipping the biopython 1.87 reference corpus (16 entries spanning all 8 mandatory CONTEXT.md §1 categories plus 8 additional cases), the developer-only Python generator script (BSD-3-Clause attribution; deterministic byte-stable output), `TestSWG_CrossValidation` in swg_test.go (16 per-entry t.Run sub-tests asserting |ours - biopython_normalised| <= 1e-9; zero in-Go normalisation logic per CONTEXT.md §1 decision (a); biopython version embedded in failure messages for triage), and the `make regen-swg-cross-validation` developer-only Makefile target. The one_long_gap_canary entry (PITFALLS.md §3 #2 Gotoh-erratum gate) records biopython_score=3.0 / biopython_normalised=0.5 — our implementation matches with delta=0.00e+00 (exact float equality), confirming the corrected Flouri 2015 affine-gap formulation in swg.go is wired correctly.**

## Performance

- **Duration:** ~8 min
- **Tasks:** 3 (Task 1: Python script + corpus generation; Task 2: TestSWG_CrossValidation; Task 3: Makefile target + CONTRIBUTING.md doc)
- **Files created:** 2 (scripts/gen-swg-cross-validation.py, testdata/cross-validation/swg/vectors.json)
- **Files modified:** 3 (swg_test.go, Makefile, CONTRIBUTING.md)

### `TestSWG_CrossValidation` runtime

- Wall: ~0.26s for the full test invocation (go test infrastructure + 16 sub-tests).
- Per-entry overhead: dominated by JSON parse (one-time) + 16 SWG kernel invocations on short inputs (each sub-microsecond on the ASCII fast path; the rune path is taken only for `unicode_ascii_only` and `identity_long`).

### Corpus byte size

- `testdata/cross-validation/swg/vectors.json` is 4577 bytes; 16 entries × ~280 bytes/entry average (inflated by the params block per entry and the 2-space indent).

### Re-run determinism

- `python3 scripts/gen-swg-cross-validation.py` produces a byte-identical file across repeated runs (verified by `diff` against a snapshot — exit 0). Determinism rests on (a) the locked CASES list ordering, (b) Python 3.7+ dict insertion-order preservation, (c) biopython's deterministic PairwiseAligner scoring on a given version, and (d) `json.dump(sort_keys=False, indent=2)` + trailing `"\n"`.

## Accomplishments

- All 16 corpus entries match our `SmithWatermanGotohScoreWithParams` output with delta=0.00e+00 (exact float equality, well within the 1e-9 tolerance gate).
- The load-bearing one_long_gap_canary entry (abc________def vs abcdef) records biopython_score=3.0 / biopython_normalised=0.5; our implementation matches exactly — the Gotoh-erratum gate is closed at the cross-validation layer (in addition to the unit/property/BDD layers from plan 03-01).
- The TDD RED-phase verification confirmed the test correctly detects deviations: temporarily poisoning the canary value to 0.99 produced a clean failure message `delta=4.90e-01 tol=1e-09 biopython=1.87` with full context (entry name, both inputs, SWGParams, our_score, biopython_normalised) — sufficient for one-step triage if a future regression slips through.
- `make check` exits 0 at task close: fmt-check, vet, lint, license-headers (65 files), deps-allowlist, tidy-check, security (govulncheck), test (-race -shuffle=on), coverage 97.1% ≥ 95% floor, public-API funcs all exercised (37 exported symbols).
- The Python script is the first Python file in `scripts/` and the first cross-validation tooling in the repo — establishes the corpus + generator + Makefile target pattern for any future algorithm needing independent-implementation cross-validation.
- biopython's BSD-3-Clause licence is correctly attributed in the script docstring per `.claude/skills/algorithm-licensing-standards` — no biopython code is copied; only reference scores are consumed.

## Task Commits

1. **Task 1: Python script + biopython corpus** — `c6b0b17` (feat)
2. **Task 2: TestSWG_CrossValidation appended to swg_test.go** — `c3f331c` (test)
3. **Task 3: Makefile regen-swg-cross-validation target + CONTRIBUTING.md doc** — `de64f3d` (chore)

## Files Created/Modified

- `scripts/gen-swg-cross-validation.py` — created (145 lines; Apache-2.0 + BSD-3-Clause attribution; `Bio.Align.PairwiseAligner` mode=`"local"`; deterministic CASES list with 16 entries; `score_case` handles both-empty / one-empty short-circuits BEFORE invoking the aligner; `main` writes canonical-byte-form JSON via `json.dump(indent=2, sort_keys=False)` + trailing LF).
- `testdata/cross-validation/swg/vectors.json` — created (4577 bytes; 16 entries; biopython_version=1.87; field order locked: name → a → b → params → biopython_score → biopython_normalised).
- `swg_test.go` — modified (added `encoding/json`, `os`, `path/filepath` imports; appended `TestSWG_CrossValidation` with 16 per-entry t.Run sub-tests; failure message format includes the biopython version from the corpus header).
- `Makefile` — modified (added `regen-swg-cross-validation` target between `verify-license-headers` and `release-check`; tolerant `command -v python3` gate; added target name to `.PHONY` list).
- `CONTRIBUTING.md` — modified (added a `make regen-swg-cross-validation` bullet to the "Make Targets" section — auto-fixed during Task 3 as a Rule 3 blocking issue when `TestMakefile_TargetsDocumentedInContributing` fired).

## Decisions Made

- **biopython 1.87** is the version pinned in the corpus header (the script reads `Bio.__version__` at run time, so the version string is automatically captured; bumping biopython locally and re-running `make regen-swg-cross-validation` surfaces in the JSON diff).
- **16 entries** in vectors.json (target range 15-20). The 8 mandatory categories + 8 additional spanning single_char_match, single_char_diff, partial_middle_match, all_mismatch, unicode_ascii_only, identity_long, near_identical, substring_at_end.
- **Python script license header treatment:** BOTH the Apache-2.0 `#`-comment block AND the docstring biopython BSD-3-Clause attribution. `scripts/verify-license-headers.sh` only checks `.go` files (confirmed by reading the script), so the Apache header is defensive habit and the BSD-3-Clause attribution is the load-bearing licensing record per `.claude/skills/algorithm-licensing-standards`.
- **Documentation route over suppress comment** for the Makefile target meta-gate: added a bullet to CONTRIBUTING.md "Make Targets" rather than adding `## suppress: developer-only` — the target is visible to developers and warrants a one-bullet explanation.

## Deviations from Plan

### Auto-fixed Issues

**1. [Rule 3 — Blocking] Added `make regen-swg-cross-validation` entry to CONTRIBUTING.md "Make Targets"**

- **Found during:** Task 3 `make check` run
- **Issue:** `TestMakefile_TargetsDocumentedInContributing` (makefile_targets_test.go:148) requires every Makefile target to either appear in CONTRIBUTING.md's "Make Targets" section OR carry a `## suppress: <reason>` comment immediately before its rule. After adding the new target, the meta-test failed with `regen-swg-cross-validation` listed as undocumented. The plan acknowledged the documentation requirement implicitly (Acceptance Criteria mentions the target style mirroring `security` / `verify-*`) but did not enumerate the CONTRIBUTING.md update.
- **Fix:** Added a bullet to the "Make Targets" section in CONTRIBUTING.md:
  > `make regen-swg-cross-validation` — developer-only; regenerate `testdata/cross-validation/swg/vectors.json` from biopython's `Bio.Align.PairwiseAligner` via `scripts/gen-swg-cross-validation.py`. Requires `python3 -m pip install --user biopython` (1.85+). NOT part of `make check`; CI consumes the committed JSON via `TestSWG_CrossValidation` and does not require Python.
- **Files modified:** CONTRIBUTING.md
- **Verification:** `go test -run TestMakefile ./...` passes; subsequent `make check` exits 0.
- **Committed in:** `de64f3d` (Task 3 commit — discovered + fixed during Task 3 `make check` run).

---

**Total deviations:** 1 auto-fixed (Rule 3 blocking).
**Impact on plan:** Non-scope-changing — the documentation entry is the natural place for a user-facing target; the alternative `## suppress:` comment would have been a less-discoverable workaround.

## Issues Encountered

- The unicode entry (`café` / `cafe`) was emitted by `json.dump` with the `é` escaped as `é` because the script does NOT pass `ensure_ascii=False`. This is harmless (still valid JSON, still parsable) and was not specified in the plan; left as-is so the byte form remains pure ASCII (simpler for grep / cross-platform diff). Our SmithWatermanGotohScoreWithParams handles the unescaped UTF-8 input correctly (Go's `encoding/json` decoder unescapes `é` to `é` on read) — verified by the test passing with delta=0.

## User Setup Required

- For developers regenerating the corpus: `python3 -m pip install --user biopython` (Python 3.7+; biopython 1.85+). NOT required at CI time — the committed JSON is the verification fixture.

## Cross-Validation Reference Values

### Load-bearing canary entry (PITFALLS.md §3 #2 — Gotoh-erratum gate)

```
name:                 one_long_gap_canary
a:                    "abc________def"   (14 bytes; 8-position gap between abc and def)
b:                    "abcdef"           (6 bytes)
params:               {match: 1.0, mismatch: -1.0, gap_open: -1.5, gap_extend: -0.5}  (defaults)
biopython_score:      3.0
biopython_normalised: 0.5                (3.0 / min(14, 6) = 3.0 / 6 = 0.5)
our_score:            0.5                (delta 0.00e+00 — exact float equality)
```

Rationale: the best local alignment finds either the 3-position prefix "abc" or the 3-position suffix "def" fresh (raw 3 each); the full-string alignment would pay `Match × 6 + GapOpen × 1 + GapExtend × 7 = 6 - 1.5 - 3.5 = 1.0`, strictly worse. Result: bestRaw = 3.0, normalised = 0.5. This value matches plan 03-01's staging-golden expectation byte-for-byte; the Phase 3 implementation is independently verified against biopython at this corpus entry.

### Full entry list (in CASES order from the script)

1. `identity_short` — biopython_score=5.0, normalised=1.0
2. `both_empty` — biopython_score=0.0, normalised=1.0 (Python-side short-circuit)
3. `one_empty_a` — biopython_score=0.0, normalised=0.0
4. `one_empty_b` — biopython_score=0.0, normalised=0.0
5. `two_substring` — biopython_score=12.0, normalised=1.0 (substring containment; clamp engages)
6. `no_overlap` — biopython_score=0.0, normalised=0.0
7. `one_long_gap_canary` — biopython_score=3.0, normalised=0.5 (Gotoh-erratum canary)
8. `non_default_params` — biopython_score=6.0, normalised=1.0 (Match=2, hello/hallo, 3 matches × 2 = 6 > min_len → clamps)
9. `single_char_match` — biopython_score=1.0, normalised=1.0
10. `single_char_diff` — biopython_score=0.0, normalised=0.0
11. `partial_middle_match` — biopython_score=3.0, normalised=0.42857142857142855 (`xxabcyy`/`zzabczz` → 3-match "abc", 3/7)
12. `all_mismatch` — biopython_score=0.0, normalised=0.0
13. `unicode_ascii_only` — biopython_score=3.0, normalised=0.75 (`café`/`cafe` → 3-match "caf", 3/4)
14. `identity_long` — biopython_score=19.0, normalised=1.0 (19-char identity match)
15. `near_identical` — biopython_score=3.0, normalised=0.5 (`kitten`/`sitting` → 3-match "itt", 3/6)
16. `substring_at_end` — biopython_score=6.0, normalised=1.0 (substring containment; clamp engages)

All 16 entries match our implementation with delta=0.00e+00 (exact float equality, well within 1e-9 tolerance).

## Hand-off Contract

**To plan 03-03 (finalisation):**
- `testdata/cross-validation/swg/vectors.json` is **read-only** at this point — plan 03-03 does NOT modify it. The cross-validation gate is closed.
- `scripts/gen-swg-cross-validation.py` is **read-only** at this point — plan 03-03 does NOT modify it. If a future biopython bump invalidates the corpus, that's a separate developer action.
- The `make regen-swg-cross-validation` target is documented in CONTRIBUTING.md "Make Targets". Plan 03-03 does NOT touch this target or its CONTRIBUTING.md entry.
- The `TestSWG_CrossValidation` test is part of the default `go test ./...` cycle. Plan 03-03's `make check` run will exercise it automatically — any regression to swg.go's DP kernel surfaces here with delta > 1e-9 and the biopython version in the failure message.
- The algorithm-correctness-reviewer evidence chain is complete: primary-source citations (Smith-Waterman 1981, Gotoh 1982, Flouri et al. 2015) → fresh transcription of the corrected Flouri 2015 recurrence → unit tests + property tests + BDD scenarios (plan 03-01) → biopython cross-validation with 16 entries (this plan). PR review can sign off on Phase 3's Gotoh-erratum gate.

**To future biopython version bumps:**
- A developer runs `python3 -m pip install --user --upgrade biopython` locally.
- Then `make regen-swg-cross-validation` to regenerate vectors.json (the new biopython_version is automatically captured in the corpus header).
- `git diff testdata/cross-validation/swg/vectors.json` shows the version-string change and any score drift.
- If `go test -run TestSWG_CrossValidation ./...` passes, commit the updated JSON.
- If it fails, EITHER biopython changed scoring semantics (investigate biopython release notes) OR our implementation drifted (investigate swg.go). The failure message includes the biopython version for one-step triage.

## Next Phase Readiness

- Plan 03-03 (finalisation) can attach immediately — the cross-validation gate is closed; the public surface (SmithWatermanGotohScoreWithParams + SWGParams) and the corpus are stable.
- Plan 03-03's scope (per plan 03-01 SUMMARY hand-off): merge `_staging/swg.json` into `testdata/golden/algorithms.json`; extend `cross_algorithm_consistency_test.go` with the SWG-vs-Levenshtein divergence test; update `docs/requirements.md` §7.1.8 to list all 6 SWG public functions (Raw* surface expansion); regenerate `bench.txt`; extend `examples/identifier-similarity/main.go` and `llms-full.txt`. None of these touch the artefacts from this plan.

## Self-Check

Verified files-exist:
- `scripts/gen-swg-cross-validation.py`: FOUND
- `testdata/cross-validation/swg/vectors.json`: FOUND
- `swg_test.go` (with TestSWG_CrossValidation): FOUND
- `Makefile` (with regen-swg-cross-validation target): FOUND
- `CONTRIBUTING.md` (with regen-swg-cross-validation bullet): FOUND

Verified commits-exist:
- `c6b0b17`: FOUND (Task 1)
- `c3f331c`: FOUND (Task 2)
- `de64f3d`: FOUND (Task 3)

## Self-Check: PASSED

---
*Phase: 03-smith-waterman-gotoh*
*Plan: 02*
*Completed: 2026-05-14*
