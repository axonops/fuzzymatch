---
phase: 03-smith-waterman-gotoh
plan: 02
type: execute
wave: 2
depends_on:
  - 03-01-swg-implementation
files_modified:
  - scripts/gen-swg-cross-validation.py
  - testdata/cross-validation/swg/vectors.json
  - swg_test.go
  - Makefile
autonomous: true
requirements:
  - CHAR-08
tags: [smith-waterman-gotoh, swg, biopython, cross-validation, gotoh-erratum-gate, reference-vectors, makefile, python-script]

must_haves:
  truths:
    # Goal-backward truths (cross-validation gate the algorithm-correctness-reviewer signs off on)
    - "A developer running `make regen-swg-cross-validation` (after `pip install --user biopython`) regenerates testdata/cross-validation/swg/vectors.json from biopython's Bio.Align.PairwiseAligner — committed JSON, no CI Python dependency"
    - "testdata/cross-validation/swg/vectors.json exists, parses as valid JSON, and is canonical-byte-stable (2-space indent, trailing LF, no BOM, deterministic CASES ordering preserved by `json.dump(sort_keys=False)` on Python 3.7+)"
    - "The vectors.json corpus contains at minimum the 8 mandatory CONTEXT.md §1 + 03-PATTERNS.md §unique-files-1 entries: identity_short, both_empty, one_empty_a, one_empty_b, two_substring, no_overlap, one_long_gap_canary, non_default_params; plus ~5-10 additional cases spanning unicode / single-char / all-mismatch / partial-middle-match to reach ~10-20 entries total"
    - "The one_long_gap_canary entry — input `abc________def` vs `abcdef` — is the Gotoh-erratum gate per PITFALLS.md §3; any failure of this entry alone forces algorithm-correctness-reviewer to block the PR"
    - "Each entry includes BOTH biopython_score (raw alignment score) AND biopython_normalised (clamp(raw/min_len, 0, 1)) — the Python script owns the normalisation reference; the Go test compares against biopython_normalised with zero in-Go normalisation logic (CONTEXT.md §1)"
    - "TestSWG_CrossValidation in swg_test.go reads the JSON corpus, parses it, and asserts |our_normalised - biopython_normalised| <= 1e-9 for every entry; the test runs as part of the default `go test ./...` cycle with NO Python runtime requirement at test time (CONTEXT.md §1)"
    - "TestSWG_CrossValidation includes the biopython version from the corpus header in any failure message so developers can correlate failures with biopython version bumps"
    - "Makefile has a new target `regen-swg-cross-validation` that invokes `python3 scripts/gen-swg-cross-validation.py`; the target is gracefully diagnostic if `python3` is not on PATH (developer-friendly error message pointing at `pip install --user biopython`)"
    - "Makefile `.PHONY` list includes `regen-swg-cross-validation`"
    - "`make check` exits 0 — including TestSWG_CrossValidation under the default `go test ./...` run"
    # Cross-cutting truths
    - "scripts/gen-swg-cross-validation.py is the first Python file in scripts/ (existing scripts are bash); it declares its dependency on biopython 1.85+ in a docstring header"
    - "The Python script's license treatment: a `#`-comment Apache-2.0 header at the top of the file if scripts/verify-license-headers.sh is configured to check .py files; otherwise a docstring attribution to biopython's BSD-3-Clause license at the file top. Confirm at execution time."
    - "The script uses biopython's Bio.Align.PairwiseAligner (NOT the deprecated pairwise2 — CONTEXT.md §1) with mode=\"local\" and explicit match_score / mismatch_score / open_gap_score / extend_gap_score assignments matching the SWGParams field semantics"
    - "Both-empty and one-empty cases are handled in the Python script BEFORE invoking the aligner (biopython's aligner.score(\"\", \"\") behaviour is implementation-defined); the script returns (0.0, 1.0) for both-empty and (0.0, 0.0) for one-empty per CONTEXT.md §1 normalisation alignment"
    - "vectors.json schema is locked: top-level `{version: 1, biopython_version: \"<x.y.z>\", entries: [...]}`; each entry has `{name, a, b, params: {match, mismatch, gap_open, gap_extend}, biopython_score, biopython_normalised}`"
    - "TestSWG_CrossValidation uses sub-tests via t.Run(e.Name, ...) so individual entry failures are visible in the test output without truncation"
    - "The cross-validation test produces a clear failure message including the entry Name, both inputs, the SWGParams used, our_score, biopython_normalised, the delta, the tolerance, and the biopython version from the corpus header"
  artifacts:
    - path: "scripts/gen-swg-cross-validation.py"
      provides: "Developer-only Python script regenerating vectors.json from biopython's PairwiseAligner; CASES list owns the deterministic case ordering"
      min_lines: 80
      contains: "Bio.Align.PairwiseAligner"
    - path: "testdata/cross-validation/swg/vectors.json"
      provides: "Committed JSON corpus of ~10-20 entries with biopython reference scores; the cross-validation fixture for TestSWG_CrossValidation"
      contains: "one_long_gap_canary"
    - path: "swg_test.go"
      provides: "Appended TestSWG_CrossValidation that reads testdata/cross-validation/swg/vectors.json and asserts agreement within 1e-9 tolerance"
      contains: "TestSWG_CrossValidation"
    - path: "Makefile"
      provides: "New regen-swg-cross-validation target (developer-only) for regenerating the corpus; added to .PHONY list"
      contains: "regen-swg-cross-validation"
  key_links:
    - from: "scripts/gen-swg-cross-validation.py"
      to: "testdata/cross-validation/swg/vectors.json"
      via: "python3 invocation via `make regen-swg-cross-validation`; writes JSON via `json.dump(obj, indent=2, sort_keys=False)` + trailing \\n"
      pattern: "testdata/cross-validation/swg/vectors\\.json"
    - from: "swg_test.go TestSWG_CrossValidation"
      to: "testdata/cross-validation/swg/vectors.json"
      via: "os.ReadFile + json.Unmarshal + per-entry sub-test asserting |our - biopython_normalised| <= 1e-9"
      pattern: "TestSWG_CrossValidation"
    - from: "swg_test.go TestSWG_CrossValidation"
      to: "swg.go SmithWatermanGotohScoreWithParams"
      via: "constructs SWGParams from the entry's params block and calls ScoreWithParams"
      pattern: "fuzzymatch\\.SmithWatermanGotohScoreWithParams"
    - from: "Makefile regen-swg-cross-validation"
      to: "scripts/gen-swg-cross-validation.py"
      via: "python3 scripts/gen-swg-cross-validation.py"
      pattern: "python3 scripts/gen-swg-cross-validation\\.py"

user_setup:
  - service: biopython
    why: "Developer-only — regenerate the cross-validation corpus on a developer machine. CI does NOT install biopython; the committed JSON is the verification fixture. Required only when the corpus must be updated (e.g. biopython version bump, new test cases added)."
    env_vars: []
    dashboard_config:
      - task: "Install biopython 1.85 or later"
        location: "Developer machine: `python3 -m pip install --user biopython` (or use a virtualenv/pyenv). Python 3.7+ required (for dict insertion-order preservation in json.dump)."
---

<objective>
Provide the biopython cross-validation evidence path for the Smith-Waterman-Gotoh implementation per CONTEXT.md §1 and the algorithm-correctness-reviewer gate per the Phase 3 ROADMAP success criterion. This plan ships three artefacts and one Makefile target:

(1) A Python script `scripts/gen-swg-cross-validation.py` that uses biopython's `Bio.Align.PairwiseAligner` to compute raw + normalised reference scores for a fixed, deterministic, committed-in-source set of ~10-20 test cases spanning all CONTEXT.md §1 / 03-PATTERNS.md §unique-files-1 categories (identity, both-empty, one-empty, two-substring, no-overlap, one-long-gap-canary, non-default-params, plus ~5-10 additional cases).

(2) A committed JSON corpus `testdata/cross-validation/swg/vectors.json` that is the verification fixture — `go test ./...` reads it directly with NO Python runtime requirement at test time.

(3) A Go test `TestSWG_CrossValidation` (appended to swg_test.go from plan 03-01) that parses the corpus and asserts agreement between our `SmithWatermanGotohScoreWithParams` and the biopython reference within 1e-9 tolerance — per-entry sub-tests via `t.Run` so individual failures are visible.

(4) A Makefile target `regen-swg-cross-validation` (developer-only) for re-generating the corpus when needed (biopython version bump, new cases added). CI never runs this target.

Purpose: the corrected Gotoh-erratum implementation (plan 03-01) requires evidence stronger than "primary-source citation + property tests" per PITFALLS.md §3. This plan provides that evidence in a CI-friendly form: the corpus is regenerable on a developer machine but verifies on every CI run via the existing `go test ./...` pipeline.

Output: an independent-implementation cross-validation gate the algorithm-correctness-reviewer signs off on. Any DP transcription drift triggers an obvious, attributable failure in `TestSWG_CrossValidation` with the biopython version + entry name in the error message.
</objective>

<execution_context>
@$HOME/.claude/get-shit-done/workflows/execute-plan.md
@$HOME/.claude/get-shit-done/templates/summary.md
</execution_context>

<context>
@.planning/PROJECT.md
@.planning/ROADMAP.md
@.planning/REQUIREMENTS.md
@.planning/STATE.md
@CLAUDE.md
@.planning/phases/03-smith-waterman-gotoh/03-CONTEXT.md
@.planning/phases/03-smith-waterman-gotoh/03-RESEARCH.md
@.planning/phases/03-smith-waterman-gotoh/03-PATTERNS.md
@.planning/phases/03-smith-waterman-gotoh/03-VALIDATION.md
@.planning/phases/03-smith-waterman-gotoh/03-01-swg-implementation-PLAN.md
@.planning/research/PITFALLS.md
@docs/requirements.md
@.claude/skills/algorithm-correctness-standards/SKILL.md
@.claude/skills/algorithm-licensing-standards/SKILL.md
@.claude/skills/determinism-standards/SKILL.md
@.claude/skills/go-coding-standards/SKILL.md
@.claude/skills/go-testing-standards/SKILL.md

<interfaces>
This plan depends on plan 03-01's public surface:

From swg.go (plan 03-01 Task 1):
  type SWGParams struct {
      Match     float64
      Mismatch  float64
      GapOpen   float64
      GapExtend float64
  }
  func NewSWGParams() SWGParams   // returns {1.0, -1.0, -1.5, -0.5}
  func SmithWatermanGotohScore(a, b string) float64
  func SmithWatermanGotohScoreWithParams(a, b string, params SWGParams) float64
  // (5 other public functions also available — this plan only uses ScoreWithParams for parameterised cross-validation, plus the default-params Score path implicitly for the default-params entries.)

From swg_test.go (plan 03-01 Task 2):
  package fuzzymatch_test
  imports: math, testing, github.com/axonops/fuzzymatch
  // This plan APPENDS TestSWG_CrossValidation; it MUST add the following
  // imports to the existing import block: encoding/json, os, path/filepath.

From export_test.go (no changes required):
  // The cross-validation test runs in package fuzzymatch_test using only
  // exported symbols (SmithWatermanGotohScoreWithParams + SWGParams + NewSWGParams).

From testdata/cross-validation/ (new directory):
  // First cross-validation corpus in the repo. The Phase 2 staging-golden
  // pattern (testdata/golden/_staging/) pins our own output stability; this
  // fixture pins agreement with an external reference implementation. They
  // are orthogonal: a passing _staging file proves the Go output is
  // byte-stable; a passing vectors.json proves the Go output matches biopython.

From Makefile (extend-only):
  // Existing targets include `verify-determinism` (line ~188), `verify-license-headers`,
  // `verify-no-runtime-deps`, `security` (the closest stylistic analog for the
  // "tolerant if tool not installed" pattern). The new `regen-swg-cross-validation`
  // target follows the `verify-*` style and inserts AFTER `verify-license-headers`,
  // BEFORE `release-check`. .PHONY list at lines ~26-28 needs the new target name.

JSON corpus schema (LOCKED per CONTEXT.md §1 and 03-PATTERNS.md §unique-files-1):

  {
    "version": 1,
    "biopython_version": "1.85",
    "entries": [
      {
        "name": "identity_short",
        "a": "hello",
        "b": "hello",
        "params": {
          "match": 1.0,
          "mismatch": -1.0,
          "gap_open": -1.5,
          "gap_extend": -0.5
        },
        "biopython_score": 5.0,
        "biopython_normalised": 1.0
      },
      ...
    ]
  }

  Field order LOCKED for canonical-byte stability: name, a, b, params, biopython_score, biopython_normalised. Top-level: version, biopython_version, entries. Python json.dump preserves dict insertion order on 3.7+.

Both-empty / one-empty Python-side conventions (CONTEXT.md §1):

  if a == "" and b == "":
      return 0.0, 1.0    # raw=0, normalised=1 (matches our both-empty identity short-circuit)
  if a == "" or b == "":
      return 0.0, 0.0    # raw=0, normalised=0 (matches our one-empty case)

  For non-empty cases, the script invokes biopython's PairwiseAligner in local
  mode and computes `normalised = clamp(raw / min(len(a), len(b)), 0, 1)`.
</interfaces>

<canonical_decisions_locked_for_this_plan>
The decisions this plan's executor must honour without re-deriving:

1. **biopython is the cross-validation tool** (NOT EMBOSS — deferred per CONTEXT.md §1). EMBOSS is captured in CONTEXT.md `<deferred>` for future regression triage.
2. **Bio.Align.PairwiseAligner with mode="local"** (NOT the deprecated `pairwise2.align.localxs` mentioned in docs/requirements.md §7.1.8 — PairwiseAligner has been biopython's actively-supported replacement since v1.79).
3. **The committed JSON is the verification fixture; the Python script regenerates on developer demand.** CI never installs biopython. (CONTEXT.md §1)
4. **Normalisation is owned by the Python script.** Each entry exposes BOTH `biopython_score` (raw) and `biopython_normalised` (script-computed clamp). The Go test compares against `biopython_normalised` with zero in-Go normalisation logic — captures both gates simultaneously. (CONTEXT.md §1)
5. **Tolerance is 1e-9** — matches the existing cross_algorithm_consistency_test.go epsilon convention. (CONTEXT.md §1)
6. **Entry ordering is deterministic** — the script's `CASES` list owns it; Python json.dump(sort_keys=False) preserves dict insertion order (Python 3.7+ guarantee). (03-PATTERNS.md §unique-files-2 determinism note)
7. **All 8 mandatory categories MUST be present**: identity, both-empty, one-empty_a, one-empty_b, two-substring, no-overlap, one-long-gap-canary, non-default-params. Plus ~5-10 additional spanning unicode / single-char / all-mismatch / partial-middle-match to reach ~10-20 entries.
8. **The one-long-gap-canary is the load-bearing entry** — its biopython_normalised value pins the Flouri 2015 corrected initialisation; failure forces algorithm-correctness-reviewer to block the PR.
9. **TestSWG_CrossValidation uses sub-tests via `t.Run(e.Name, ...)`** so individual entry failures are visible without truncation.
10. **biopython is BSD-3-Clause licensed** — compatible with Apache-2.0 for reference-vector cross-validation per `algorithm-licensing-standards`. Code is NOT copied from biopython — only its computed reference scores are committed. (CONTEXT.md §1 + algorithm-licensing-standards skill)
</canonical_decisions_locked_for_this_plan>
</context>

<tasks>

<task type="auto">
  <name>Task 1: scripts/gen-swg-cross-validation.py — biopython corpus generator</name>
  <files>scripts/gen-swg-cross-validation.py, testdata/cross-validation/swg/vectors.json</files>
  <read_first>
    - .planning/phases/03-smith-waterman-gotoh/03-CONTEXT.md §1 (cross-validation evidence path — tool, schema, normalisation alignment decision (a))
    - .planning/phases/03-smith-waterman-gotoh/03-PATTERNS.md §unique-files-2 (Python skeleton template — full reference implementation including CASES list, DEFAULT_PARAMS, score_case + main functions; determinism notes)
    - .planning/phases/03-smith-waterman-gotoh/03-PATTERNS.md §unique-files-1 (JSON schema and required entries)
    - .planning/phases/03-smith-waterman-gotoh/03-RESEARCH.md §Pattern 5 lines 449-534 (the Python skeleton; cross-reference with PATTERNS.md to confirm consistency)
    - scripts/verify-license-headers.sh (check whether .py files are included in the license-header gate — if YES, add the Apache-2.0 header as a `#`-comment block; if NO, the docstring header is sufficient)
    - .claude/skills/algorithm-licensing-standards/SKILL.md (biopython BSD-3-Clause attribution — script docstring must record biopython as the reference implementation with its license)
    - .claude/skills/determinism-standards/SKILL.md (no randomness, no environment dependence, deterministic ordering)
    - docs/requirements.md §7.1.8 (the default SWGParams values to ensure script matches Go-side defaults exactly)
  </read_first>
  <action>
Create scripts/gen-swg-cross-validation.py as a developer-only Python script following the skeleton from 03-PATTERNS.md §unique-files-2.

1. Top-of-file shebang: `#!/usr/bin/env python3`.

2. Apache-2.0 license header as a `#`-comment block at the top (mirror the .sh files in scripts/ — verify-license-headers.sh handles .sh; if it also handles .py, this header satisfies the gate; if not, the comment is still good practice). Then a long docstring/comment block explaining:
   - Purpose: regenerates testdata/cross-validation/swg/vectors.json from biopython's Bio.Align.PairwiseAligner.
   - License attribution: biopython is BSD-3-Clause; this script consumes biopython solely for reference-vector cross-validation per .claude/skills/algorithm-licensing-standards (no code copied from biopython).
   - Usage: `make regen-swg-cross-validation` or `python3 scripts/gen-swg-cross-validation.py`.
   - Requirements: biopython 1.85+ via `python3 -m pip install --user biopython`. Python 3.7+ for dict-order-preservation in json.dump.
   - Normalisation note: script computes BOTH the raw biopython alignment score AND the script-side normalised reference (clamp(raw / min(len(a), len(b)), 0, 1)). The Go test compares against `biopython_normalised` with zero in-Go normalisation logic.

3. Imports: `json`, `os`, `Bio` (for `Bio.__version__`), `from Bio.Align import PairwiseAligner`.

4. Module-level constants:

       DEFAULT_PARAMS = {
           "match": 1.0,
           "mismatch": -1.0,
           "gap_open": -1.5,
           "gap_extend": -0.5,
       }

5. CASES list — fixed, deterministic, ordered tuples `(name, a, b, params_override)` where params_override is `None` for default params and a dict otherwise. Required entries (the 8 mandatory categories from CONTEXT.md §1 + 03-PATTERNS.md §unique-files-1):

       CASES = [
           ("identity_short",       "hello",          "hello",                          None),
           ("both_empty",           "",               "",                               None),
           ("one_empty_a",          "",               "abcdef",                         None),
           ("one_empty_b",          "abcdef",         "",                               None),
           ("two_substring",        "http_request",   "http_request_header_fields",     None),
           ("no_overlap",           "qqqq",           "zzzz",                           None),
           ("one_long_gap_canary",  "abc________def", "abcdef",                         None),
           ("non_default_params",   "hello",          "hallo",
               {"match": 2.0, "mismatch": -2.0, "gap_open": -3.0, "gap_extend": -1.0}),
           # Additional ~5-10 cases spanning the rest of the input space:
           ("single_char_match",    "a",              "a",                              None),
           ("single_char_diff",     "a",              "b",                              None),
           ("partial_middle_match", "xxabcyy",        "zzabczz",                        None),
           ("all_mismatch",         "abcd",           "wxyz",                           None),
           ("unicode_ascii_only",   "café",           "cafe",                           None),
           ("identity_long",        "the quick brown fox",  "the quick brown fox",      None),
           ("near_identical",       "kitten",         "sitting",                        None),
           ("substring_at_end",     "ending",         "the long ending",                None),
       ]

   (The script may include 1-2 additional cases at the executor's discretion to reach the upper end of the ~10-20 range; do not exceed 20.)

6. score_case(a, b, params) function:
   - If `a == "" and b == ""`: return `(0.0, 1.0)` (raw=0, normalised=1 — matches our both-empty identity short-circuit).
   - Elif `a == "" or b == ""`: return `(0.0, 0.0)` (raw=0, normalised=0).
   - Else: instantiate `aligner = PairwiseAligner()`; set `aligner.mode = "local"`; set `aligner.match_score = params["match"]`, `aligner.mismatch_score = params["mismatch"]`, `aligner.open_gap_score = params["gap_open"]`, `aligner.extend_gap_score = params["gap_extend"]`. Compute `raw = aligner.score(a, b)`. Compute `min_len = min(len(a), len(b))`. Compute `norm = raw / min_len`. Clamp: `norm = max(0.0, min(1.0, norm))`. Return `(raw, norm)`.

7. main() function:
   - Initialise `entries = []`.
   - For each `(name, a, b, overrides)` in CASES: compute `params = dict(DEFAULT_PARAMS)` then `if overrides: params.update(overrides)`; call `raw, norm = score_case(a, b, params)`; append a dict with field order `{"name": name, "a": a, "b": b, "params": params, "biopython_score": raw, "biopython_normalised": norm}`.
   - Build `out = {"version": 1, "biopython_version": Bio.__version__, "entries": entries}`.
   - `path = "testdata/cross-validation/swg/vectors.json"`.
   - `os.makedirs(os.path.dirname(path), exist_ok=True)`.
   - Open the file for writing; call `json.dump(out, f, indent=2, sort_keys=False)`; write a trailing `"\n"` (matches Phase 2 golden-file canonical convention — trailing LF byte).

8. `if __name__ == "__main__": main()` at the bottom.

9. Determinism: the script must produce byte-identical output on repeated runs (assuming the same biopython version). No random number generators, no environment-dependent inputs, no system clocks.

After writing the script:

   - Install biopython locally if not already installed: `python3 -m pip install --user biopython` (or in a virtualenv).
   - Run the script: `python3 scripts/gen-swg-cross-validation.py`.
   - Inspect the output: open testdata/cross-validation/swg/vectors.json; confirm it contains a `version: 1` top-level field, a `biopython_version` string matching the installed version, and at least 15 entries with the field order specified above.
   - Confirm trailing newline: `xxd testdata/cross-validation/swg/vectors.json | tail -1` should show the file ending with `0a` (LF).
   - Confirm canonical byte form: 2-space indent (`grep -c '^  ' testdata/cross-validation/swg/vectors.json | xargs -I{} test {} -gt 10`).
   - Re-run the script and `diff` against the previous output — must be zero diff (deterministic regeneration).

Commit both `scripts/gen-swg-cross-validation.py` and `testdata/cross-validation/swg/vectors.json`.
  </action>
  <verify>
    <automated>test -f scripts/gen-swg-cross-validation.py && test -f testdata/cross-validation/swg/vectors.json && python3 -c "import json; d = json.load(open('testdata/cross-validation/swg/vectors.json')); assert d['version'] == 1; assert 'biopython_version' in d; assert len(d['entries']) >= 15; names = {e['name'] for e in d['entries']}; required = {'identity_short', 'both_empty', 'one_empty_a', 'one_empty_b', 'two_substring', 'no_overlap', 'one_long_gap_canary', 'non_default_params'}; assert required.issubset(names), f'Missing required entries: {required - names}'; print('OK', len(d['entries']), 'entries')"</automated>
  </verify>
  <acceptance_criteria>
    - scripts/gen-swg-cross-validation.py exists, is executable (`test -x scripts/gen-swg-cross-validation.py` OR runs via `python3 scripts/gen-swg-cross-validation.py`), and contains the literal text `Bio.Align.PairwiseAligner` (`grep -c 'Bio.Align.PairwiseAligner' scripts/gen-swg-cross-validation.py` returns ≥ 1).
    - scripts/gen-swg-cross-validation.py contains the literal `mode = "local"` or `mode=\"local\"` (PairwiseAligner local-mode setup).
    - scripts/gen-swg-cross-validation.py docstring/comment block names biopython's BSD-3-Clause license and references .claude/skills/algorithm-licensing-standards (or equivalent attribution phrasing).
    - scripts/gen-swg-cross-validation.py uses `json.dump(...indent=2, sort_keys=False...)` (verifiable by `grep -c 'sort_keys=False' scripts/gen-swg-cross-validation.py` returning ≥ 1).
    - testdata/cross-validation/swg/vectors.json exists and parses as valid JSON (`python3 -c "import json; json.load(open('testdata/cross-validation/swg/vectors.json'))"` exits 0).
    - testdata/cross-validation/swg/vectors.json has top-level fields `version` (== 1), `biopython_version` (non-empty string), and `entries` (array).
    - testdata/cross-validation/swg/vectors.json `entries` array length is ≥ 15 AND ≤ 20.
    - All 8 mandatory entry names present (identity_short, both_empty, one_empty_a, one_empty_b, two_substring, no_overlap, one_long_gap_canary, non_default_params).
    - The `one_long_gap_canary` entry has `a == "abc________def"` and `b == "abcdef"` and a finite `biopython_normalised` value in [0, 1].
    - Each entry has the exact field order `name, a, b, params, biopython_score, biopython_normalised` (verifiable by `python3 -c "import json; d = json.load(open('testdata/cross-validation/swg/vectors.json')); e = d['entries'][0]; assert list(e.keys()) == ['name', 'a', 'b', 'params', 'biopython_score', 'biopython_normalised']"`).
    - Each `params` dict has the exact keys `match, mismatch, gap_open, gap_extend` with float values (default entries use 1.0, -1.0, -1.5, -0.5).
    - File is canonical byte form: 2-space indent (`grep -cE '^  "' testdata/cross-validation/swg/vectors.json` returns ≥ 5), trailing LF (`xxd testdata/cross-validation/swg/vectors.json | tail -1` shows `0a` at end), no BOM (`xxd testdata/cross-validation/swg/vectors.json | head -1` does NOT show `efbb bf`).
    - Re-running `python3 scripts/gen-swg-cross-validation.py` produces zero diff against the committed file (deterministic regeneration).
  </acceptance_criteria>
  <behavior>
    - The script regenerates the corpus deterministically: same biopython version → byte-identical output across runs.
    - The corpus covers all 8 mandatory CONTEXT.md §1 categories plus 7+ additional cases (target ~15-18 total).
    - The corpus's biopython_normalised values are in [0, 1] for every entry (clamp invariant).
    - The corpus exposes BOTH biopython_score (raw) and biopython_normalised (script-side clamped) — supporting CONTEXT.md §1 decision (a) where the Go test compares against the normalised value with zero in-Go normalisation logic.
    - Both-empty case returns (raw=0.0, normalised=1.0); one-empty cases return (raw=0.0, normalised=0.0); these match our Go-side identity short-circuit and one-empty short-circuit exactly.
  </behavior>
  <done>
    scripts/gen-swg-cross-validation.py committed with biopython BSD-3-Clause attribution; testdata/cross-validation/swg/vectors.json committed with ≥ 15 entries covering all 8 mandatory categories; both files byte-stable across re-runs. Ready for Task 2 to consume the corpus.
  </done>
</task>

<task type="auto" tdd="true">
  <name>Task 2: Append TestSWG_CrossValidation to swg_test.go (the cross-validation gate)</name>
  <files>swg_test.go</files>
  <read_first>
    - swg_test.go (current state from plan 03-01 Task 2 — find the file trailer; this task APPENDS TestSWG_CrossValidation after the existing tests)
    - testdata/cross-validation/swg/vectors.json (output of Task 1 — the corpus the test loads)
    - swg.go (confirm SWGParams + SmithWatermanGotohScoreWithParams signatures from plan 03-01 Task 1)
    - .planning/phases/03-smith-waterman-gotoh/03-PATTERNS.md §unique-files-3 (TestSWG_CrossValidation full test loader skeleton — including the type-definitions block, the path constant, the corpus loading logic, the per-entry t.Run sub-test, and the failure message format)
    - .planning/phases/03-smith-waterman-gotoh/03-CONTEXT.md §1 (tolerance 1e-9; biopython version in failure message)
    - .planning/phases/03-smith-waterman-gotoh/03-VALIDATION.md (task 03-02-01, 03-02-02 — the verification commands for the cross-validation entries and the green-gate criterion)
    - cross_algorithm_consistency_test.go (existing epsilon convention — confirms 1e-9 is the project-wide tolerance for cross-validation against external references)
    - .claude/skills/go-testing-standards/SKILL.md (sub-test naming, error message conventions, stdlib testing only in root tests)
  </read_first>
  <action>
Append `TestSWG_CrossValidation` to swg_test.go (do NOT create a new file; this test belongs in the same file as the rest of the SWG unit tests per 03-PATTERNS.md §unique-files-3 ("inside swg_test.go")).

1. Add the necessary imports to the existing import block at the top of swg_test.go (the block was created in plan 03-01 Task 2 with `math` + `testing` + `github.com/axonops/fuzzymatch`). This task ADDS three stdlib imports: `encoding/json`, `os`, `path/filepath`. After this task, the import block should be (alphabetised):

       import (
           "encoding/json"
           "math"
           "os"
           "path/filepath"
           "testing"

           "github.com/axonops/fuzzymatch"
       )

2. Append the test function at the file trailer (after the last existing TestSmithWatermanGotoh_* function):

       // TestSWG_CrossValidation asserts agreement between our SmithWatermanGotoh
       // implementation and the biopython reference corpus committed at
       // testdata/cross-validation/swg/vectors.json.
       //
       // Tolerance: |our_normalised - biopython_normalised| <= 1e-9 (matches the
       // cross_algorithm_consistency_test.go epsilon convention).
       //
       // The corpus is regenerated by `make regen-swg-cross-validation` (developer-
       // only); CI does NOT require Python. If this test fails after a corpus
       // regeneration, EITHER our DP kernel drifted from the corrected Gotoh
       // formulation OR the biopython version emitted different scores (the
       // biopython version from the corpus header is included in the failure
       // message for triage).
       //
       // Per-entry sub-tests via t.Run so individual entry failures are visible
       // without truncation in the test output.
       func TestSWG_CrossValidation(t *testing.T) {
           const epsilon = 1e-9
           type paramsBlock struct {
               Match     float64 `json:"match"`
               Mismatch  float64 `json:"mismatch"`
               GapOpen   float64 `json:"gap_open"`
               GapExtend float64 `json:"gap_extend"`
           }
           type entry struct {
               Name                string      `json:"name"`
               A                   string      `json:"a"`
               B                   string      `json:"b"`
               Params              paramsBlock `json:"params"`
               BiopythonScore      float64     `json:"biopython_score"`
               BiopythonNormalised float64     `json:"biopython_normalised"`
           }
           type corpus struct {
               Version          int     `json:"version"`
               BiopythonVersion string  `json:"biopython_version"`
               Entries          []entry `json:"entries"`
           }
           path := filepath.Join("testdata", "cross-validation", "swg", "vectors.json")
           raw, err := os.ReadFile(path)
           if err != nil {
               t.Fatalf("TestSWG_CrossValidation: read %s: %v (regenerate with `make regen-swg-cross-validation`)", path, err)
           }
           var c corpus
           if err := json.Unmarshal(raw, &c); err != nil {
               t.Fatalf("TestSWG_CrossValidation: parse %s: %v", path, err)
           }
           if c.Version != 1 {
               t.Fatalf("TestSWG_CrossValidation: unsupported corpus version %d (want 1)", c.Version)
           }
           if len(c.Entries) == 0 {
               t.Fatalf("TestSWG_CrossValidation: empty corpus")
           }
           for _, e := range c.Entries {
               e := e // local copy for the closure
               t.Run(e.Name, func(t *testing.T) {
                   params := fuzzymatch.SWGParams{
                       Match:     e.Params.Match,
                       Mismatch:  e.Params.Mismatch,
                       GapOpen:   e.Params.GapOpen,
                       GapExtend: e.Params.GapExtend,
                   }
                   got := fuzzymatch.SmithWatermanGotohScoreWithParams(e.A, e.B, params)
                   delta := math.Abs(got - e.BiopythonNormalised)
                   if delta > epsilon {
                       t.Errorf("SmithWatermanGotohScoreWithParams(%q, %q, %+v) = %.12f; biopython_normalised = %.12f (delta %.2e, tol %g, biopython %s)",
                           e.A, e.B, params, got, e.BiopythonNormalised, delta, epsilon, c.BiopythonVersion)
                   }
               })
           }
       }

3. NO testify in this test (stdlib testing only — per CLAUDE.md "Constraints" and .claude/skills/go-coding-standards).

4. Run the test:

       go test -race -shuffle=on -count=1 -run TestSWG_CrossValidation ./...

   Expected: exits 0 with one PASS per entry (sub-test names match the entry Names from vectors.json).

5. Also confirm the test runs as part of the default `go test ./...` cycle (no special tags or build flags required):

       go test -race -count=1 ./...

   Expected: TestSWG_CrossValidation appears in the test output alongside the rest of the suite, all green.
  </action>
  <verify>
    <automated>go test -race -shuffle=on -count=1 -run TestSWG_CrossValidation ./... -v 2>&1 | grep -cE '=== RUN[[:space:]]+TestSWG_CrossValidation/' | xargs -I{} test {} -ge 8</automated>
  </verify>
  <acceptance_criteria>
    - swg_test.go's import block contains all of `encoding/json`, `math`, `os`, `path/filepath`, `testing`, `github.com/axonops/fuzzymatch` (verifiable by `grep -E '"encoding/json"|"math"|"os"|"path/filepath"|"testing"|"github.com/axonops/fuzzymatch"' swg_test.go | wc -l` returning ≥ 6).
    - swg_test.go contains `func TestSWG_CrossValidation(t *testing.T)` exactly once (`grep -c 'func TestSWG_CrossValidation' swg_test.go` returns 1).
    - The test uses `t.Run(e.Name, ...)` for per-entry sub-tests (`grep -c 't.Run(e.Name' swg_test.go` returns 1, or `grep -cE 't\.Run\([^)]*e\.Name' swg_test.go` returns ≥ 1).
    - The test uses the literal tolerance `1e-9` and the literal `epsilon` constant (`grep -c '1e-9' swg_test.go` returns ≥ 1).
    - The test uses `math.Abs(got - e.BiopythonNormalised) > epsilon` as the comparison (`grep -c 'math.Abs(got' swg_test.go` returns ≥ 1).
    - The test loads the corpus from the exact path `testdata/cross-validation/swg/vectors.json` (`grep -cE 'testdata.*cross.validation.*swg.*vectors\.json' swg_test.go` returns ≥ 1 OR `grep -c 'filepath.Join("testdata", "cross-validation", "swg", "vectors.json")' swg_test.go` returns ≥ 1).
    - The failure message includes the biopython version (verifiable by `grep -c 'BiopythonVersion' swg_test.go` returning ≥ 1 OR `grep -c 'biopython %s' swg_test.go` returning ≥ 1).
    - The test uses fuzzymatch.SmithWatermanGotohScoreWithParams and fuzzymatch.SWGParams (`grep -c 'fuzzymatch.SmithWatermanGotohScoreWithParams' swg_test.go` returns ≥ 1; `grep -c 'fuzzymatch.SWGParams' swg_test.go` returns ≥ 1).
    - `grep -c 'github.com/stretchr/testify' swg_test.go` returns 0 (no testify in root tests).
    - `go test -race -shuffle=on -count=1 -run TestSWG_CrossValidation ./...` exits 0.
    - Sub-test count from `go test -run TestSWG_CrossValidation -v ./...` matches the corpus entry count (verifiable by counting `--- PASS: TestSWG_CrossValidation/` lines and confirming it equals `len(d['entries'])` from the corpus).
    - The one_long_gap_canary sub-test passes (this is the load-bearing Gotoh-erratum gate — if it fails, the algorithm is wrong).
    - `go test -race -count=1 ./...` (full suite) exits 0 — TestSWG_CrossValidation runs as part of the default cycle with no opt-in.
  </acceptance_criteria>
  <behavior>
    - TestSWG_CrossValidation reads the committed JSON corpus, parses ~15 entries, and runs one sub-test per entry.
    - Each sub-test compares our SmithWatermanGotohScoreWithParams output against biopython_normalised within 1e-9 tolerance.
    - Failure messages include the entry name, both input strings, the SWGParams used, our_score, biopython_normalised, the delta, the tolerance, and the biopython version from the corpus header — sufficient for one-step triage.
    - The one_long_gap_canary sub-test is the PITFALLS.md §3 #2 gate — its passing means the corrected Flouri 2015 initialisation is wired in correctly.
    - Test runs as part of `go test ./...` with no build tags, no environment variables, no Python runtime requirement at test time.
  </behavior>
  <done>
    TestSWG_CrossValidation appended to swg_test.go; runs green on the corpus from Task 1; all ~15 sub-tests pass within 1e-9 tolerance; the one_long_gap_canary entry passes (Gotoh-erratum gate closed).
  </done>
</task>

<task type="auto">
  <name>Task 3: Makefile target regen-swg-cross-validation (developer-only)</name>
  <files>Makefile</files>
  <read_first>
    - Makefile (lines 26-28 — `.PHONY` list; lines ~188-191 — `verify-determinism` target style; lines ~171-176 — `security` target as the "tolerant if tool not installed" pattern analog; line ~194 — `verify-license-headers` target as the insertion point; line ~197 — `release-check` target as the boundary marker)
    - .planning/phases/03-smith-waterman-gotoh/03-PATTERNS.md (`Makefile` section + §unique-files-4 — full target template with python3 prerequisite check)
    - .planning/phases/03-smith-waterman-gotoh/03-CONTEXT.md §1 (regen-swg-cross-validation as the developer-only Makefile target; TestSWG_CrossValidation in `make check` covers verification)
  </read_first>
  <action>
Append a new `regen-swg-cross-validation` target to the Makefile and update the `.PHONY` list.

1. Locate the `.PHONY` list (around lines 26-28). Add `regen-swg-cross-validation` to the list, keeping alphabetical ordering where applicable (or appending at the end of the existing list; match the file's local convention).

2. Locate the boundary: AFTER the `verify-license-headers` target (around line 194), BEFORE the `release-check` target (around line 197). Insert the new target with the following template (mirrors the `security` target's "tolerant if tool not installed" pattern from ~lines 171-176):

       # Regenerates testdata/cross-validation/swg/vectors.json by invoking the
       # biopython-based generator script. Developer-only — NOT included in
       # `make check`. The committed JSON is the verification fixture; CI does
       # NOT require Python at test time. Re-run this target when biopython is
       # bumped or when new test cases are added to scripts/gen-swg-cross-validation.py.
       #
       # Requires: python3 + biopython 1.85+
       #   python3 -m pip install --user biopython
       regen-swg-cross-validation:
       	@if ! command -v python3 >/dev/null 2>&1; then \
       	  echo "python3 not found; install Python 3.x and run: python3 -m pip install --user biopython"; \
       	  exit 1; \
       	fi
       	python3 scripts/gen-swg-cross-validation.py

   The leading whitespace on the recipe lines MUST be hard tabs (Makefile syntax).

3. Confirm the target is callable. From the repo root:

       make regen-swg-cross-validation

   Expected outcome (with biopython installed): re-runs scripts/gen-swg-cross-validation.py; testdata/cross-validation/swg/vectors.json is regenerated; git status shows zero diff (the script is deterministic — if the same biopython version is installed as when Task 1 wrote the file, the output is byte-identical).

   Expected outcome (with python3 missing): the target prints the diagnostic message and exits 1 (the `command -v python3` gate).

4. Confirm `make check` is unchanged — the new target is NOT in `make check`'s dependency chain (just like `verify-determinism` is referenced by `check` but `regen-swg-cross-validation` is not). TestSWG_CrossValidation (added in Task 2) IS in `make check` via the default `go test ./...` path.

5. Run `make check` end-to-end:

       make check

   Expected: exits 0 (full quality gate green, including TestSWG_CrossValidation under the default go test cycle).
  </action>
  <verify>
    <automated>grep -E '^regen-swg-cross-validation:' Makefile && grep -c 'regen-swg-cross-validation' Makefile | xargs -I{} test {} -ge 2 && make check</automated>
  </verify>
  <acceptance_criteria>
    - Makefile contains a target line `regen-swg-cross-validation:` (`grep -cE '^regen-swg-cross-validation:' Makefile` returns 1).
    - The target recipe invokes `python3 scripts/gen-swg-cross-validation.py` (`grep -c 'python3 scripts/gen-swg-cross-validation.py' Makefile` returns ≥ 1).
    - The target has a `command -v python3` prerequisite check with a diagnostic message pointing at `pip install --user biopython` (`grep -c 'command -v python3' Makefile` returns ≥ 1; `grep -c 'biopython' Makefile` returns ≥ 1).
    - The `.PHONY` declaration list includes `regen-swg-cross-validation` (verifiable by `grep -c 'regen-swg-cross-validation' Makefile` returning ≥ 2 — appears in both the target definition and the .PHONY list).
    - The new target is NOT a dependency of the `check` target (`grep -E '^check:' Makefile` does NOT contain `regen-swg-cross-validation` as a prerequisite).
    - `make regen-swg-cross-validation` succeeds when python3 + biopython are installed (deterministic regeneration produces zero diff against the committed vectors.json).
    - `make regen-swg-cross-validation` fails gracefully with the documented diagnostic message when python3 is missing (the `command -v` gate returns non-zero).
    - `make check` exits 0 — the full quality gate is green, TestSWG_CrossValidation runs under `go test ./...` and all sub-tests pass.
  </acceptance_criteria>
  <behavior>
    - Developers can re-generate the cross-validation corpus on demand via `make regen-swg-cross-validation` after installing biopython locally.
    - CI does NOT invoke the regen target; CI validates against the committed JSON via `go test ./...` → TestSWG_CrossValidation.
    - When biopython is bumped (e.g. to a future 1.86), a developer runs the regen target, commits the updated vectors.json (with the new biopython_version recorded), and the CI matrix verifies our implementation still matches within 1e-9.
    - Missing python3 produces a friendly, actionable error message rather than a cryptic Make error.
  </behavior>
  <done>
    Makefile updated with the new regen-swg-cross-validation target and .PHONY entry; `make regen-swg-cross-validation` works on a machine with biopython installed; `make check` exits 0; Phase 3's cross-validation gate is fully wired (script + corpus + Go test + Makefile target).
  </done>
</task>

</tasks>

<threat_model>
## Trust Boundaries

| Boundary | Description |
|----------|-------------|
| Developer machine → biopython 1.85+ | The Python script runs on a developer machine to regenerate the corpus. Supply chain: biopython is installed via pip from PyPI. If pip is compromised or biopython releases a buggy version, the regenerated corpus could contain wrong reference scores. Mitigated by the human review of the resulting JSON diff before commit, AND by the fact that the corpus is COMMITTED to git (CI never re-runs the script). |
| go test → testdata/cross-validation/swg/vectors.json | Read-only JSON file load. JSON parse errors fail the test loudly. The file is committed under version control; tampering surfaces in git diff. |
| go test → swg.go SmithWatermanGotohScoreWithParams | Pure function call; no I/O. Same trust boundaries as plan 03-01. |

## STRIDE Threat Register

| Threat ID | Category | Component | Disposition | Mitigation Plan |
|-----------|----------|-----------|-------------|-----------------|
| T-3-04 | Tampering / Supply Chain | biopython corpus could be tampered with (corrupted vectors.json committed to repo) | mitigate | The committed JSON is the verification fixture, version-controlled and visible in PR diffs. Regeneration is a deliberate developer action via `make regen-swg-cross-validation` — the script is deterministic for a given biopython version. PR review enforces correctness before merge. The corpus header records the biopython version used; mismatches between corpus version and current local biopython surface in TestSWG_CrossValidation failure messages. HIGH severity per planning context; mitigation is complete (commit-then-review + deterministic regeneration + version header). |
| T-3-04b | Spoofing | An attacker submits a PR that subtly tampers with vectors.json to mask a Gotoh-erratum regression in swg.go | mitigate | PR review requires algorithm-correctness-reviewer sign-off per CLAUDE.md "Workflow — Agent Gates" §2. Any vectors.json change without a corresponding script change OR a documented biopython version bump is a red flag. The one_long_gap_canary entry is the load-bearing canary — tampering with it to mask a Gotoh-erratum bug requires explicit malice and surfaces in PR diff. |
| T-3-CV-01 | Denial of Service | scripts/gen-swg-cross-validation.py never used in CI; cannot DoS the CI pipeline | accept | The script is developer-only; CI invokes only `go test`. No mitigation needed. |
| T-3-CV-02 | Information Disclosure | TestSWG_CrossValidation failure message includes input strings (which could be sensitive if the corpus contained PII) | accept | The corpus contains only hand-curated public strings (`hello`, `abc`, `http_request`, etc.) — no PII, no secrets. Future corpus additions must follow the same convention. |
| T-3-CV-03 | Tampering | A subtle bug in scripts/gen-swg-cross-validation.py produces wrong normalised values silently | mitigate | The script's logic is small (~80 lines) and reviewable. The score_case function's clamp logic mirrors the documented `clamp(raw/min_len, 0, 1)` formula explicitly. The both-empty / one-empty short-circuits exactly match our Go-side conventions. PR review enforces correctness. As a defense-in-depth check, a developer can spot-verify one or two entries by hand against biopython's documented behaviour (e.g. identity_short should have biopython_normalised == 1.0 — trivially verifiable). |

Severity assessment: T-3-04 (corpus tampering) is HIGH; the mitigation is complete via commit-then-review + deterministic regeneration. T-3-04b is MEDIUM; mitigation via algorithm-correctness-reviewer gate. Others are LOW or accepted with rationale. ASVS L1 V14 (Configuration) addressed: the cross-validation fixture is version-controlled. Plan passes the security gate.
</threat_model>

<verification>
1. Build: `go build ./...` exits 0 (no Go code changes break the build — the only Go change is appending TestSWG_CrossValidation to swg_test.go).
2. Vet: `go vet ./...` exits 0.
3. License headers: `bash scripts/verify-license-headers.sh` exits 0 (no new .go files; only the Python script is new — its license treatment per Task 1's acceptance criteria).
4. No-runtime-deps: `bash scripts/verify-no-runtime-deps.sh` exits 0 (the Python script is in scripts/, the JSON corpus is testdata; neither affects root go.mod).
5. JSON validity: `python3 -c "import json; json.load(open('testdata/cross-validation/swg/vectors.json'))"` exits 0.
6. Corpus completeness: corpus contains ≥ 15 entries spanning all 8 mandatory categories (verified by the Task 1 `<automated>` block).
7. Cross-validation test: `go test -race -shuffle=on -count=1 -run TestSWG_CrossValidation ./...` exits 0; sub-test count matches corpus entry count; all sub-tests pass including the one_long_gap_canary canary.
8. Regen idempotence: `make regen-swg-cross-validation` re-runs the script; `git diff testdata/cross-validation/swg/vectors.json` shows zero diff (deterministic regeneration with the same biopython version).
9. Full quality gate: `make check` exits 0 — including TestSWG_CrossValidation in the default `go test ./...` cycle.
</verification>

<success_criteria>
- A developer can run `make regen-swg-cross-validation` (after installing biopython locally) and regenerate testdata/cross-validation/swg/vectors.json deterministically.
- The committed vectors.json corpus contains ≥ 15 entries covering all 8 mandatory CONTEXT.md §1 categories — the one_long_gap_canary entry is present (PITFALLS.md §3 #2 gate).
- TestSWG_CrossValidation runs as part of the default `go test ./...` cycle, with no Python runtime requirement at test time, and verifies our SmithWatermanGotohScoreWithParams matches biopython_normalised within 1e-9 tolerance for every entry.
- Failure messages include the biopython version, the entry name, both input strings, the SWGParams used, our_score, biopython_normalised, the delta, and the tolerance — sufficient for one-step triage.
- The algorithm-correctness-reviewer agent has its cross-validation evidence per CLAUDE.md "Workflow — Agent Gates" §2; PR review can sign off on the Gotoh-erratum gate.
- `make check` exits 0; Phase 3's cross-validation discipline is fully wired and CI-verifiable.
</success_criteria>

<output>
After completion, create `.planning/phases/03-smith-waterman-gotoh/03-02-swg-cross-validation-SUMMARY.md` per the standard summary template, recording:
- The biopython version pinned in the corpus header (e.g. 1.85, or whatever the developer installed).
- The exact entry count in vectors.json and the list of entry Names (sorted as they appear in the file).
- The one_long_gap_canary entry's biopython_score and biopython_normalised values (these are the load-bearing Gotoh-erratum reference numbers; record them so future regressions are easy to diff).
- The TestSWG_CrossValidation runtime (typical: < 100 ms for a 15-entry corpus).
- Whether the Python script's license treatment used the Apache-2.0 `#`-comment header or the docstring attribution (per scripts/verify-license-headers.sh's coverage).
- Any deviations from the plan and their rationale.
- The hand-off contract to plan 03-03: vectors.json is read-only at this point (plan 03-03 does NOT modify it); the cross-validation gate is closed and ready for the finalisation step.
</output>
