---
phase: 04-remaining-character-gestalt
plan: 04
subsystem: similarity-algorithms
tags: [ratcliff-obershelp, cross-validation, difflib, autojunk-false, python-stdlib, committed-corpus, python-3-7-version-check, makefile-regen-target, contributing-doc, byte-path-semantics, utf-8-bytes-encoding, gate-test]

# Dependency graph
requires:
  - phase: 03-smith-waterman-gotoh
    provides: scripts/gen-swg-cross-validation.py (1-to-1 structural analog — 221 lines copied + biopython→difflib swap + Python version guard in place of biopython version guard); testdata/cross-validation/swg/vectors.json (schema shape, indent=2, trailing LF byte-stability convention); swg_test.go lines 411–479 (TestSWG_CrossValidation pattern); Makefile lines 196–211 (regen-swg-cross-validation target pattern); CONTRIBUTING.md line 92 (developer-only doc-entry pattern); Phase 3 IN-07 closure (Python version check at script entry)
  - phase: 04-remaining-character-gestalt
    provides: 04-03 RatcliffObershelpScore + RatcliffObershelpScoreRunes public surface (the function called by TestRatcliffObershelp_CrossValidation); 04-03 numerical-regression pin (TestRatcliffObershelp_PinnedDrDobbsValue — outside the corpus per WR-03 closure)
provides:
  - scripts/gen-ratcliff-obershelp-cross-validation.py — Python 3.7+ stdlib-only corpus generator; 16 cases; difflib.SequenceMatcher(autojunk=False) on UTF-8 byte-encoded inputs
  - testdata/cross-validation/ratcliff-obershelp/vectors.json — committed 16-entry corpus across all four mandatory CONTEXT.md §1 categories
  - TestRatcliffObershelp_CrossValidation — per-entry sub-test gate within 1e-9 tolerance; reads JSON, zero Python at test time
  - TestRatcliffObershelp_CrossValidation_CorpusShape — 15..18 entry-count gate + four-category presence assertion
  - Makefile regen-ratcliff-obershelp-cross-validation target — developer-only, shell-gated on python3, no pip install needed
  - CONTRIBUTING.md doc entry for the regen target
  - Closure of the algorithm-correctness-reviewer cross-validation gate for Ratcliff-Obershelp — plan 04-05 finalisation may now merge goldens
affects: [04-05-finalisation]

# Tech tracking
tech-stack:
  added: []
  patterns:
    - "Byte-path semantics in the cross-validation corpus. The Go RatcliffObershelpScore operates on UTF-8 BYTE slices, not codepoints, so the Python generator MUST call difflib.SequenceMatcher(autojunk=False, a=a.encode('utf-8'), b=b.encode('utf-8')).ratio() — encoding to bytes BEFORE invoking difflib. For ASCII entries the encoding is identity (UTF-8 of ASCII is ASCII), so all 14 ASCII entries are unaffected. For the unicode_ascii_only entry (café/cafe), encoding to bytes yields 6/9 = 0.6667 (matching our Go byte path); calling difflib on Python str instead would yield 6/8 = 0.75 (codepoint-path), breaking byte-for-byte equivalence. This is the structural deviation from gen-swg-cross-validation.py (where the SWG aligner operates on either ASCII or already-encoded inputs consistently); the difflib version requires explicit byte-encoding for the unicode entry. Documented inline in score_case's docstring."
    - "autojunk=False is the load-bearing keyword (the SINGLE most important detail in the entire phase per CONTEXT.md §1 + RESEARCH.md Pitfall 2). The keyword is the FIRST argument to difflib.SequenceMatcher in score_case so grep gates on 'autojunk=False' anchor to that line. The autojunk_sensitive corpus entry — 205-char inputs with character 'a' appearing in well over 1% of positions — has difflib_ratio 0.7317 with autojunk=False vs 0.2439 with autojunk=True. The TestRatcliffObershelp_CrossValidation/autojunk_sensitive sub-test is the keystone: if autojunk is ever silently re-enabled in the generator OR our impl ever grows an autojunk-like heuristic, that sub-test fails byte-for-byte and surfaces the regression."
    - "Python version assertion at script entry per Phase 3 IN-07 closure. _check_python_version() asserts sys.version_info >= (3, 7); the cutoff is the first Python release where dict insertion-order preservation is a language guarantee (3.6 had it as a CPython implementation detail). Without the guard, regenerating on 3.6 or earlier could produce reordered JSON breaking corpus byte-stability across runs. Same pattern as gen-swg-cross-validation.py's _check_biopython_version helper."
    - "Two-test split: TestRatcliffObershelp_CrossValidation (numerical equivalence per entry) + TestRatcliffObershelp_CrossValidation_CorpusShape (entry-count + four-category presence). The shape test prevents a future regression that silently trims the autojunk_sensitive keystone entry — without the shape test, removing that entry from the corpus would still leave TestRatcliffObershelp_CrossValidation green on the remaining entries, silently disarming the keystone gate. The shape test asserts entry-name substring matching for each of the four CONTEXT.md §1 categories."

key-files:
  created:
    - scripts/gen-ratcliff-obershelp-cross-validation.py
    - testdata/cross-validation/ratcliff-obershelp/vectors.json
    - .planning/phases/04-remaining-character-gestalt/deferred-items.md
  modified:
    - ratcliff_obershelp_test.go (appended: roCrossValidationEntry / roCrossValidationCorpus struct types + TestRatcliffObershelp_CrossValidation + TestRatcliffObershelp_CrossValidation_CorpusShape; imports extended to include encoding/json, os, path/filepath, strings)
    - Makefile (appended: regen-ratcliff-obershelp-cross-validation target after regen-swg-cross-validation; .PHONY list extended)
    - CONTRIBUTING.md (appended: doc entry for regen-ratcliff-obershelp-cross-validation in the Make Targets section)

key-decisions:
  - "UTF-8 byte encoding in the generator script. The Python difflib operates on Python strings (codepoints) by default; our Go RatcliffObershelpScore is a byte-path function. To get byte-for-byte equivalence on the unicode_ascii_only entry, the generator encodes each input to UTF-8 bytes BEFORE invoking difflib. For ASCII inputs this is a no-op (UTF-8 of ASCII is identity, both forms agree). For multi-byte inputs (café/cafe) this matters — bytes give 6/9 = 0.6667, codepoints give 6/8 = 0.75. Documented inline in score_case's docstring + the module-level Byte-path semantics section."
  - "Corpus has exactly 16 entries (within the 15..18 plan range). Breakdown: 5 standard edges (identity_short, both_empty, one_empty_a, one_empty_b, no_overlap), 2 Dr. Dobb's 1988 (wikimedia_wikimania, gestalt_paper), 1 autojunk-sensitive (205-char keystone), 8 substring/partial/unicode (substring_middle, partial_overlap, unicode_ascii_only, longer_identity, substring_at_start, substring_at_end, single_char_match, near_identical). 16 was chosen rather than 15 because all eight Category 4 entries earn their keep — substring_at_start vs substring_at_end pins different DP-table traversal patterns, near_identical pins 5-of-6 match behaviour, single_char_match pins the trivial 1-char case."
  - "Hard-wired both-empty (returns 1.0) and one-empty (returns 0.0) short-circuits in score_case BEFORE the difflib call. Mirrors the Go-side short-circuits in ratcliff_obershelp.go. difflib's own behaviour agrees with this for both-empty and one-empty (1.0 and 0.0 respectively), but the explicit short-circuit removes any risk of corpus drift if difflib's empty-input behaviour ever changes in a future Python release. Same pattern as gen-swg-cross-validation.py's score_case helper."
  - "Two-test split rather than a single TestRatcliffObershelp_CrossValidation with embedded shape assertions. The shape assertions are deliberately a separate test function so failures point cleanly at the corpus structure issue (count, category coverage) rather than mixing with per-entry numerical failures. TestRatcliffObershelp_CrossValidation_CorpusShape exists to prevent silent removal of the autojunk_sensitive keystone — a future PR trimming the corpus to 14 entries would fail this shape gate even if the remaining entries still match difflib."

patterns-established:
  - "Pattern: byte-path-aware cross-validation. When generating a corpus that pins a Go byte-path function against a Python reference, encode strings to UTF-8 bytes in Python BEFORE invoking the reference algorithm. This makes the corpus byte-for-byte equivalent on both ASCII and multi-byte inputs without needing separate codepoint and byte tracks. Future phases that add cross-validation against Python references should apply this pattern when the Go target is byte-path."
  - "Pattern: corpus shape test alongside numerical equivalence test. The shape test (TestRatcliffObershelp_CrossValidation_CorpusShape) asserts entry count + category coverage independent of the numerical gates. This prevents silent corpus reductions that would still pass the numerical test. Future phases adding cross-validation corpora should ship a shape test if the corpus has load-bearing keystone entries (autojunk_sensitive is the RO keystone; the Phase 3 SWG corpus's one_long_gap_canary is analogous)."
  - "Pattern: regen Makefile target + CONTRIBUTING entry land in the same commit. The makefile_targets_test.go::TestMakefile_TargetsDocumentedInContributing meta-test fails otherwise. Phase 3 established this; plan 04-04 inherits."

requirements-completed: [GESTALT-01]

# Metrics
duration: ~20min
completed: 2026-05-14
---

# Phase 4 Plan 04: Ratcliff-Obershelp Cross-Validation Summary

**Ships the byte-for-byte difflib(autojunk=False) cross-validation gate for RatcliffObershelpScore: a Python 3.7+ stdlib-only generator script (no pip install needed — difflib is stdlib), a committed 16-entry corpus JSON spanning all four mandatory CONTEXT.md §1 categories with the load-bearing 205-char autojunk_sensitive keystone, an appended Go cross-validation test asserting |our_score - difflib_ratio| <= 1e-9 per entry, a developer-only Makefile regen target, and the CONTRIBUTING doc entry. Closes the algorithm-correctness-reviewer gate; plan 04-05 finalisation may now merge goldens.**

## Performance

- **Duration:** ~20 min (3 atomic commits)
- **Started:** 2026-05-14T15:13Z (approx — first commit timestamp)
- **Completed:** 2026-05-14T15:30Z
- **Tasks:** 3 (all completed)
- **Files modified:** 5 (2 created, 3 appended)

## Accomplishments

- 16-entry committed corpus at `testdata/cross-validation/ratcliff-obershelp/vectors.json`. Schema: `{version: 1, python_version: "3.12.12", entries: [{name, a, b, difflib_ratio}]}`. All four mandatory CONTEXT.md §1 categories covered:
  - Category 1 (standard edges, 5): identity_short, both_empty, one_empty_a, one_empty_b, no_overlap
  - Category 2 (Dr. Dobb's 1988, 2): wikimedia_wikimania (0.7778), gestalt_paper (0.4516)
  - Category 3 (autojunk-sensitive keystone, 1): autojunk_sensitive (0.7317 — clearly distinct from autojunk=True's 0.2439)
  - Category 4 (substring/partial/unicode, 8): substring_middle, partial_overlap, unicode_ascii_only (0.6667 byte-path), longer_identity, substring_at_start, substring_at_end, single_char_match, near_identical
- Python 3.7+ stdlib-only generator script — zero third-party imports verified by `! grep -E "^import (Bio|requests|numpy|pandas)" scripts/gen-ratcliff-obershelp-cross-validation.py`. autojunk=False present as the FIRST kwarg in score_case (grep gate `grep -q "autojunk=False"` exits 0). sys.version_info check present.
- TestRatcliffObershelp_CrossValidation green: all 16 entries within 1e-9. Per-entry sub-tests visible (`t.Run(e.Name, ...)`); the autojunk_sensitive sub-test is the keystone gate per VALIDATION.md row 04-04-04.
- TestRatcliffObershelp_CrossValidation_CorpusShape green: 16 entries; all four categories covered.
- Makefile `regen-ratcliff-obershelp-cross-validation` target with shell-gated python3 check; added to .PHONY alongside regen-swg-cross-validation.
- CONTRIBUTING.md doc entry in the Make Targets section; TestMakefile_TargetsDocumentedInContributing green.
- Byte-stable corpus regeneration verified: `make regen-ratcliff-obershelp-cross-validation` produces byte-identical output on repeated runs.

## Task Commits

Each task was committed atomically:

1. **Task 1: Python generator script + initial corpus generation** — `96f264b` (feat)
2. **Task 2: TestRatcliffObershelp_CrossValidation + TestRatcliffObershelp_CrossValidation_CorpusShape appended** — `8642a46` (test)
3. **Task 3: Makefile regen target + CONTRIBUTING.md doc line** — `16468d8` (chore)

## Files Created/Modified

### Created

- `scripts/gen-ratcliff-obershelp-cross-validation.py` (199 lines) — Python 3.7+ stdlib-only corpus generator. Direct structural copy of gen-swg-cross-validation.py with the load-bearing substitutions: biopython→difflib (stdlib), DEFAULT_PARAMS / per-case overrides REMOVED (Ratcliff-Obershelp has no params), aligner setup REMOVED in favour of `difflib.SequenceMatcher(autojunk=False, a=a.encode('utf-8'), b=b.encode('utf-8')).ratio()`, `_check_biopython_version()` → `_check_python_version()` asserting `sys.version_info >= (3, 7)`, output field rename `biopython_score`/`biopython_normalised` → `difflib_ratio`, header field rename `biopython_version` → `python_version`. Apache-2.0 header + module docstring documenting the autojunk=False contract + byte-path semantics + Python 3.7+ requirement + PSF licence note.
- `testdata/cross-validation/ratcliff-obershelp/vectors.json` (102 lines, 16 entries) — the committed cross-validation corpus. Schema follows `gen-swg-cross-validation.py` shape minus the `params` block.
- `.planning/phases/04-remaining-character-gestalt/deferred-items.md` — out-of-scope discovery log (records that strcmp95.go fails `gofmt -s` per the pre-existing state from plan 04-01; not touched by plan 04-04).

### Modified

- `ratcliff_obershelp_test.go` — appended two struct types (`roCrossValidationEntry`, `roCrossValidationCorpus`) and two test functions (`TestRatcliffObershelp_CrossValidation`, `TestRatcliffObershelp_CrossValidation_CorpusShape`). Imports extended: added `encoding/json`, `os`, `path/filepath`, `strings`. The cross-validation test reads the committed JSON, asserts `math.Abs(got - e.DifflibRatio) <= 1e-9` per entry with per-entry sub-tests via `t.Run(e.Name, ...)`. The shape test asserts 15..18 entries and entry-name substring matching for the four CONTEXT.md §1 categories.
- `Makefile` — appended `regen-ratcliff-obershelp-cross-validation` target after `regen-swg-cross-validation` (lines 196–211 analog). Shell-gated on `command -v python3`; informative error if absent; runs `python3 scripts/gen-ratcliff-obershelp-cross-validation.py`. Target appended to `.PHONY` line.
- `CONTRIBUTING.md` — appended doc entry for the new regen target adjacent to the existing regen-swg-cross-validation entry. Cites Python 3.7+ requirement + "no pip install needed — difflib is stdlib" simplification.

## Decisions Made

The four entries in the frontmatter `key-decisions` capture the substantive choices:

1. **UTF-8 byte encoding in the generator script before calling difflib.** Required for byte-for-byte equivalence on the unicode_ascii_only entry — bytes give 6/9 = 0.6667 (matches our Go byte-path RatcliffObershelpScore); codepoints would give 6/8 = 0.75 (Python str default behaviour). For ASCII this is a no-op.
2. **16 corpus entries (within the 15..18 plan range).** All eight Category 4 entries earn their keep — substring_at_start vs substring_at_end pin different DP-table traversal patterns; near_identical pins 5-of-6 match behaviour; single_char_match pins the trivial 1-char case.
3. **Hard-wired both-empty and one-empty short-circuits in score_case BEFORE invoking difflib.** Mirrors the Go-side short-circuits and removes any risk of corpus drift if difflib's empty-input behaviour ever changes.
4. **Two-test split: numerical equivalence + corpus shape.** The shape test prevents silent removal of the autojunk_sensitive keystone — a future PR trimming to 14 entries would fail the shape gate even if the remaining entries still match difflib.

## Deviations from Plan

None — plan 04-04 executed exactly as written. No auto-fix rules triggered, no architectural decisions needed.

## Issues Encountered

- **Pre-existing `make fmt-check` failure on `strcmp95.go`** — `gofmt -s` reports `strcmp95.go` (committed 7fb6319 by plan 04-01) needs reformatting. NOT introduced by plan 04-04; not in scope per `gsd-executor.md` scope-boundary rule. Logged to `deferred-items.md`. Recommended resolution: a small `style(04): gofmt -s strcmp95.go` follow-up, or roll the fix into plan 04-05 finalisation when strcmp95.go is touched.

## Threat Surface Scan

The plan's `<threat_model>` block enumerates three threats. All three mitigations land in this plan:

- **T-fixture-tampering (mitigate):** Corpus is committed; PR review surfaces any modification. The regen target is developer-only and reproducible (byte-stable output verified). The autojunk_sensitive keystone (TestRatcliffObershelp_CrossValidation/autojunk_sensitive sub-test) catches any silent autojunk=True regression in the generator. The shape test (TestRatcliffObershelp_CrossValidation_CorpusShape) catches silent trimming of the keystone entry.
- **T-fuzz-panic (mitigate):** Already mitigated in plan 04-03 via FuzzRatcliffObershelpScore. Plan 04-04 re-exercises the public byte-path surface on the Dr. Dobb's pairs + the 205-char autojunk-sensitive case via TestRatcliffObershelp_CrossValidation — additional smoke coverage.
- **T-float-determinism (mitigate):** `math.Abs(got - e.DifflibRatio) <= 1e-9` per-entry tolerance matches Phase 3 SWG convention. Cross-platform CI matrix (linux/amd64+arm64, darwin/arm64, windows/amd64) will verify the byte-stability assertion holds across architectures on the next CI run.

No new threat surface introduced. Omitting Threat Flags section.

## User Setup Required

None — no external service configuration. Python 3.7+ is required only by developers regenerating the corpus; CI consumes the committed JSON directly.

## Next Phase Readiness

- **Plan 04-05 (finalisation)** is now unblocked. The algorithm-correctness-reviewer cross-validation gate for Ratcliff-Obershelp is closed; finalisation may proceed to merge `testdata/golden/_staging/{strcmp95,lcsstr,ratcliff_obershelp}.json` into `testdata/golden/algorithms.json`, extend `cross_algorithm_consistency_test.go`, update the identifier-similarity example (7 → 10 columns), refresh `bench.txt`, and sync `llms.txt`/`llms-full.txt`.
- **Cross-platform CI gate:** the corpus contains irrational float values (e.g. `0.45161290322580644`, `0.7777777777777778`, `0.7317073170731707`); the Phase 1 cross-platform CI matrix will re-verify byte-identical output across architectures on the next CI run. Phase 3 SWG corpus established the precedent — irrational values from a Python reference round-trip cleanly through Go's `encoding/json` decoder at all five matrix platforms.
- **Style follow-up (deferred-items.md):** `strcmp95.go` fails `gofmt -s` — pre-existing from plan 04-01; not in plan 04-04 scope. Either a small `style(04)` follow-up or roll into plan 04-05 finalisation.

## Self-Check: PASSED

- **Files created:** `scripts/gen-ratcliff-obershelp-cross-validation.py`, `testdata/cross-validation/ratcliff-obershelp/vectors.json`, `.planning/phases/04-remaining-character-gestalt/deferred-items.md` — all FOUND on disk.
- **Files modified:** `ratcliff_obershelp_test.go`, `Makefile`, `CONTRIBUTING.md` — all show as touched in `git diff --name-only HEAD~3 HEAD`.
- **Commits exist:** `96f264b` (feat), `8642a46` (test), `16468d8` (chore) — all confirmed via `git log --oneline -3`.
- **Verification commands green:**
  - `python3 scripts/gen-ratcliff-obershelp-cross-validation.py` → exit 0
  - `python3 -c "import json; d=json.load(open('testdata/cross-validation/ratcliff-obershelp/vectors.json')); ..." → version=1, entries=16, all 4 categories covered
  - `go test -run 'TestRatcliffObershelp_CrossValidation$|TestRatcliffObershelp_CrossValidation_CorpusShape$' -v ./...` → all 18 sub-tests pass (16 cross-validation entries + 2 test functions)
  - `go test -run 'TestRatcliffObershelp_CrossValidation/autojunk_sensitive' -v ./...` → PASS (keystone sub-test specifically green)
  - `go test ./...` → ok (no regressions across the entire test suite)
  - `go test -run 'TestMakefile_TargetsDocumentedInContributing|TestMakefile_HasCanonicalTargets' -v ./...` → both PASS
  - `make regen-ratcliff-obershelp-cross-validation` → exit 0; corpus byte-stable on regeneration
  - `grep -q "^regen-ratcliff-obershelp-cross-validation:" Makefile` → OK
  - `grep -q "regen-ratcliff-obershelp-cross-validation" CONTRIBUTING.md` → OK
  - `! grep -E "^import (Bio|requests|numpy|pandas)" scripts/gen-ratcliff-obershelp-cross-validation.py` → OK (zero third-party imports)
  - `grep -q "autojunk=False" scripts/gen-ratcliff-obershelp-cross-validation.py` → OK
  - `grep -q "sys.version_info" scripts/gen-ratcliff-obershelp-cross-validation.py` → OK
  - `bash scripts/verify-license-headers.sh` → OK (80 .go files; .py headers not scanned by this script — the Apache-2.0 header is present manually-verified on the new .py file)
  - `go vet ./...` → ok
- **Pre-existing issue (out of scope):** `make fmt-check` reports strcmp95.go needs `gofmt -s` reformatting — pre-existing from plan 04-01 commit 7fb6319; not introduced by plan 04-04; logged to deferred-items.md.

---

*Phase: 04-remaining-character-gestalt*
*Completed: 2026-05-14*
