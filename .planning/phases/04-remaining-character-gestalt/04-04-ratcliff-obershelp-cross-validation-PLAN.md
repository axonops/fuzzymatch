---
phase: 04-remaining-character-gestalt
plan: 04
type: execute
wave: 4
depends_on:
  - 04-03-ratcliff-obershelp
files_modified:
  - scripts/gen-ratcliff-obershelp-cross-validation.py
  - testdata/cross-validation/ratcliff-obershelp/vectors.json
  - ratcliff_obershelp_test.go
  - Makefile
  - CONTRIBUTING.md
autonomous: true
requirements:
  - GESTALT-01
tags: [ratcliff-obershelp, cross-validation, difflib, autojunk-false, python-stdlib-difflib, committed-corpus, python-3-7-version-check, makefile-regen-target, contributing-doc, gate-test]

must_haves:
  truths:
    - "A developer can run `make regen-ratcliff-obershelp-cross-validation` on any machine with Python 3.7+ and regenerate testdata/cross-validation/ratcliff-obershelp/vectors.json byte-stably; difflib is Python stdlib so NO pip install is required (the structural simplification over Phase 3's biopython dependency)"
    - "testdata/cross-validation/ratcliff-obershelp/vectors.json exists in the repo, committed as the canonical reference fixture; CI runs `TestRatcliffObershelp_CrossValidation` against this committed file without invoking Python"
    - "The committed corpus contains 15–18 entries covering ALL FOUR mandatory CONTEXT.md §1 categories: (1) Standard edge cases (identity, both-empty, one-empty_a, one-empty_b, no-overlap); (2) Dr. Dobb's 1988 paper examples (WIKIMEDIA/WIKIMANIA, GESTALT/GESTALT_PATTERN_MATCHING); (3) autojunk-sensitive 200+ char case (load-bearing — proves autojunk=False is correctly disabled); (4) Substring + partial-match + unicode (substring_middle, partial_overlap, unicode_ascii_only e.g. café/cafe, longer_identity)"
    - "TestRatcliffObershelp_CrossValidation reads the committed JSON and asserts `|RatcliffObershelpScore(e.A, e.B) - e.DifflibRatio| <= 1e-9` for EVERY entry — including the autojunk-sensitive 200+char entry. If autojunk is ever accidentally enabled in either the impl or the generator, that entry fails byte-for-byte and surfaces the regression"
    - "scripts/gen-ratcliff-obershelp-cross-validation.py asserts `sys.version_info >= (3, 7)` at startup and emits a clear error if the running Python is too old (Phase 3 IN-07 closure — required for dict insertion-order preservation in `json.dump`)"
    - "The generator script calls `difflib.SequenceMatcher(autojunk=False, a=a, b=b).ratio()` with `autojunk=False` as the FIRST keyword argument in the `score_case` body — RESEARCH.md Pitfall 2 closure (the most important detail in the entire phase)"
    - "The corpus JSON header records `python_version` as the actual `sys.version_info` string of the generator's Python (e.g. \"3.12.4\") — same metadata-pinning pattern as Phase 3 SWG biopython_version"
    - "The corpus JSON header records `version: 1` for forward-compatibility"
    - "scripts/gen-ratcliff-obershelp-cross-validation.py uses ONLY Python stdlib (`json`, `os`, `sys`, `difflib`) — NO third-party Python deps; verifiable by `! grep -E 'import (Bio|requests|numpy|pandas)' scripts/gen-ratcliff-obershelp-cross-validation.py`"
    - "Apache-2.0 header on scripts/gen-ratcliff-obershelp-cross-validation.py (Phase 1 + Phase 3 IN-05 lock)"
    - "TestRatcliffObershelp_CrossValidation is APPENDED to ratcliff_obershelp_test.go (the file already exists from plan 04-03) — copies the structural pattern from swg_test.go lines 411–479 (`TestSWG_CrossValidation`) with substitutions: `Params paramsBlock` REMOVED, `BiopythonScore` → `DifflibRatio`, `BiopythonVersion` → `PythonVersion`, function called → `fuzzymatch.RatcliffObershelpScore(e.A, e.B)`, tolerance `1e-9`"
    - "TestRatcliffObershelp_CrossValidation_CorpusShape asserts the corpus has 15–18 entries AND covers all four mandatory categories (presence checked by entry name substring matching: at least one entry name contains \"empty\", at least one matches the Dr. Dobb's pairs, at least one contains \"autojunk\", at least one contains \"unicode\" or \"cafe\")"
    - "TestRatcliffObershelp_CrossValidation includes an embedded subtest `t.Run(\"autojunk_sensitive\", ...)` that specifically asserts the 200+char entry passes — surfacing autojunk regressions directly (VALIDATION.md row 04-04-04)"
    - "Makefile has a new `regen-ratcliff-obershelp-cross-validation` target with the EXACT shape from /Users/johnny/Development/fuzzymatch/Makefile lines 196–211 (the `regen-swg-cross-validation` analog): shell-gated on `command -v python3`; informative error if missing; runs `python3 scripts/gen-ratcliff-obershelp-cross-validation.py` and exits"
    - "The new target appears in Makefile's `.PHONY` declaration (Makefile line 28 area — Phase 3 added `regen-swg-cross-validation` to that list; append alongside)"
    - "CONTRIBUTING.md documents the new target with the exact shape from /Users/johnny/Development/fuzzymatch/CONTRIBUTING.md line 92 (the `regen-swg-cross-validation` analog): `make regen-ratcliff-obershelp-cross-validation` — developer-only; regenerates testdata/cross-validation/ratcliff-obershelp/vectors.json; requires Python 3.7+ (difflib is stdlib — no pip install needed)"
    - "The makefile_targets_test.go::TestMakefile_TargetsDocumentedInContributing meta-test passes — both the Makefile target and the CONTRIBUTING entry land in the same commit"
    - "Plan 04-04 is the algorithm-correctness-reviewer gate per Phase 3 SWG pattern — TestRatcliffObershelp_CrossValidation MUST pass green before plan 04-05 finalisation runs (the cross-platform CI matrix's golden merge depends on this gate)"
  artifacts:
    - path: "scripts/gen-ratcliff-obershelp-cross-validation.py"
      provides: "Python 3.7+ stdlib-only corpus generator. Asserts sys.version_info >= (3, 7); calls difflib.SequenceMatcher(autojunk=False).ratio() on a fixed CASES list; writes testdata/cross-validation/ratcliff-obershelp/vectors.json with version: 1, python_version, and 15–18 entries"
      min_lines: 130
      contains: "autojunk=False"
    - path: "testdata/cross-validation/ratcliff-obershelp/vectors.json"
      provides: "Committed cross-validation corpus — 15–18 entries covering 4 mandatory categories. Schema: {version: 1, python_version: 'X.Y.Z', entries: [{name, a, b, difflib_ratio}]}. CI consumes without Python at test time"
      contains: "autojunk_sensitive"
    - path: "ratcliff_obershelp_test.go"
      provides: "APPENDED TestRatcliffObershelp_CrossValidation + TestRatcliffObershelp_CrossValidation_CorpusShape — read committed JSON, assert |score - difflib_ratio| <= 1e-9 for every entry; CorpusShape asserts 15–18 entries across 4 categories"
    - path: "Makefile"
      provides: "APPENDED regen-ratcliff-obershelp-cross-validation target shell-gated on python3; added to .PHONY"
      contains: "regen-ratcliff-obershelp-cross-validation:"
    - path: "CONTRIBUTING.md"
      provides: "APPENDED documentation line describing the regen target"
      contains: "regen-ratcliff-obershelp-cross-validation"
  key_links:
    - from: "scripts/gen-ratcliff-obershelp-cross-validation.py"
      to: "Python stdlib difflib.SequenceMatcher(autojunk=False)"
      via: "Standard `import difflib` at script header; NO `pip install` required — the structural simplification over Phase 3's biopython"
      pattern: "autojunk=False"
    - from: "ratcliff_obershelp_test.go (TestRatcliffObershelp_CrossValidation)"
      to: "testdata/cross-validation/ratcliff-obershelp/vectors.json"
      via: "Reads via os.Open + json.Decoder; iterates entries; asserts math.Abs(got - e.DifflibRatio) <= 1e-9 per entry; ZERO Python invocation at test time"
      pattern: "json\\.NewDecoder|DifflibRatio|1e-9"
    - from: "Makefile (regen-ratcliff-obershelp-cross-validation target)"
      to: "CONTRIBUTING.md (documentation line)"
      via: "Both edited in the SAME commit so TestMakefile_TargetsDocumentedInContributing meta-test passes; mirrors Phase 3 SWG pattern"
      pattern: "regen-ratcliff-obershelp-cross-validation"
---

<objective>
Ship the Ratcliff-Obershelp difflib-equivalence cross-validation gate: a Python stdlib-only generator script (Python 3.7+; `difflib.SequenceMatcher(autojunk=False).ratio()`), a committed 15–18-entry corpus JSON, an appended Go cross-validation test, a developer-only Makefile regen target, and a CONTRIBUTING.md doc line. NO PIP INSTALL required — difflib is Python stdlib (the structural simplification over Phase 3's biopython).

Purpose: lock the load-bearing difflib(autojunk=False) byte-for-byte equivalence contract for RatcliffObershelpScore. The committed corpus is the verification fixture; CI never runs Python. The 200+-char autojunk-sensitive entry is the keystone gate — its presence in the corpus, matching our impl within 1e-9, proves both that the generator runs with autojunk disabled AND that our impl has no autojunk-like heuristic.

Output: 5 new/modified files (1 new Python script, 1 new JSON corpus, 1 append to existing Go test file, 1 Makefile extension, 1 CONTRIBUTING extension). This is the algorithm-correctness-reviewer gate per Phase 3 SWG pattern — must land green before plan 04-05 finalisation.
</objective>

<execution_context>
@$HOME/.claude/get-shit-done/workflows/execute-plan.md
@$HOME/.claude/get-shit-done/templates/summary.md
</execution_context>

<context>
@.planning/PROJECT.md
@.planning/ROADMAP.md
@.planning/STATE.md
@.planning/REQUIREMENTS.md
@.planning/phases/04-remaining-character-gestalt/04-CONTEXT.md
@.planning/phases/04-remaining-character-gestalt/04-RESEARCH.md
@.planning/phases/04-remaining-character-gestalt/04-PATTERNS.md
@.planning/phases/04-remaining-character-gestalt/04-VALIDATION.md
@.planning/phases/04-remaining-character-gestalt/04-03-ratcliff-obershelp-PLAN.md
@.planning/phases/03-smith-waterman-gotoh/03-02-swg-cross-validation-PLAN.md
@.planning/phases/03-smith-waterman-gotoh/03-02-swg-cross-validation-SUMMARY.md
@.claude/skills/algorithm-correctness-standards/SKILL.md
@.claude/skills/algorithm-licensing-standards/SKILL.md
@scripts/gen-swg-cross-validation.py
</context>

<interfaces>
<!-- Reference shape; executor copies and adapts from the SWG analog. -->

Phase 3 analog (full structural copy): `scripts/gen-swg-cross-validation.py` (221 lines) — DIRECT STRUCTURAL ANALOG. Substitutions during port:
- `import Bio.Align` becomes `import difflib`
- `_check_biopython_version()` becomes `_check_python_version()` asserting `sys.version_info >= (3, 7)`
- `DEFAULT_PARAMS` + per-case `overrides` + `aligner` setup are REMOVED (Ratcliff-Obershelp has no params)
- `aligner.score(a, b)` becomes `difflib.SequenceMatcher(autojunk=False, a=a, b=b).ratio()`
- Output field rename: `biopython_score` / `biopython_normalised` to `difflib_ratio`
- Header field rename: `biopython_version` to `python_version`

Committed corpus schema (JSON object): top-level keys `version` (int = 1), `python_version` (string, e.g. "3.12.4"), `entries` (array). Each entry has `name` (string), `a` (string), `b` (string), `difflib_ratio` (float64).

Test struct (defined in Go in ratcliff_obershelp_test.go): named `roCrossValidationEntry` for one entry with json tags matching the JSON keys; `roCrossValidationCorpus` for the top-level object holding `Version int`, `PythonVersion string`, and `Entries []roCrossValidationEntry`.
</interfaces>

<tasks>

<task type="auto" tdd="true">
  <name>Task 1: Python generator script + initial corpus generation</name>
  <files>scripts/gen-ratcliff-obershelp-cross-validation.py, testdata/cross-validation/ratcliff-obershelp/vectors.json</files>
  <read_first>
    - scripts/gen-swg-cross-validation.py (221 lines — exact structural analog; the executor MUST read this top-to-bottom before writing the new script)
    - .planning/phases/04-remaining-character-gestalt/04-CONTEXT.md §1 (Ratcliff-Obershelp cross-validation locked decisions: 15–18 entries, 4 mandatory categories, autojunk=False, Python ≥ 3.7, 1e-9 tolerance)
    - .planning/phases/04-remaining-character-gestalt/04-RESEARCH.md — Pitfall 2 (autojunk=False is load-bearing), Code Examples (lines 805–895 — the Python generator template)
    - .planning/phases/04-remaining-character-gestalt/04-PATTERNS.md §"scripts/gen-ratcliff-obershelp-cross-validation.py" (1-to-1 structural copy guidance)
    - .planning/phases/04-remaining-character-gestalt/04-VALIDATION.md (rows 04-04-01, 04-04-04)
    - .planning/phases/03-smith-waterman-gotoh/03-02-swg-cross-validation-SUMMARY.md (Phase 3 pattern outcome — IN-07 closure on Python version check)
    - .claude/skills/algorithm-licensing-standards/SKILL.md (PSF licence on stdlib is consulted for reference-vector cross-validation only, NOT for code copying)
    - testdata/cross-validation/swg/vectors.json (existing committed analog — schema shape, indent, trailing LF; the file you write must match this byte-pattern apart from field renames)
  </read_first>
  <behavior>
    - `python3 scripts/gen-ratcliff-obershelp-cross-validation.py` exits 0 on Python 3.7+; exits non-zero with an informative error on Python < 3.7
    - The generated file `testdata/cross-validation/ratcliff-obershelp/vectors.json` parses as valid JSON; has `version: 1`, `python_version` populated from `sys.version_info`, and 15–18 entries
    - All 4 mandatory categories are covered (entry names include at least one match for "empty", "wikimedia"/"gestalt", "autojunk", "unicode"/"cafe")
    - Re-running the script with the same Python version produces a byte-identical output file (deterministic order via insertion-order-preserving dict + `json.dump(sort_keys=False, indent=2)` + trailing LF)
    - Every entry's `difflib_ratio` is in [0.0, 1.0]; both-empty and identity entries are 1.0; one-empty entries are 0.0; no-overlap entries are 0.0
    - The autojunk_sensitive entry's value differs from what autojunk=True would have produced (smoke-checked by running the generator once with autojunk=True locally; not part of the committed test gate — the corpus entry's presence is the gate)
  </behavior>
  <action>
    Create scripts/gen-ratcliff-obershelp-cross-validation.py per PATTERNS.md §"scripts/gen-ratcliff-obershelp-cross-validation.py" and RESEARCH.md Code Examples (lines 805–895). Use the exact structural skeleton from scripts/gen-swg-cross-validation.py.

    File order: shebang `#!/usr/bin/env python3`; Apache-2.0 header (copy gen-swg-cross-validation.py lines 1–14 verbatim, substituting the script-purpose paragraph); module docstring explaining the purpose (regenerate corpus via difflib.SequenceMatcher(autojunk=False).ratio() on a fixed CASES list); licence note (PSF stdlib, consulted for reference vectors only); the autojunk=False requirement justification (RESEARCH.md Pitfall 2); the Python 3.7+ requirement (insertion-order dicts for byte-stable JSON).

    Imports section: only stdlib modules — `json`, `os`, `sys`, `difflib`. NO `import Bio.Align` or other third-party imports.

    Define `_MIN_PYTHON_VERSION = (3, 7)` as a module-level constant.

    Define `CASES` as a list of (name, a, b) tuples. Required entries covering all 4 categories from CONTEXT.md §1:
    - Category 1 — Standard edge cases (4–5 entries): identity_short ("hello"/"hello"), both_empty (""/""), one_empty_a (""/"abcdef"), one_empty_b ("abcdef"/""), no_overlap ("qqqq"/"zzzz")
    - Category 2 — Dr. Dobb's 1988 paper examples (2 entries): wikimedia_wikimania ("WIKIMEDIA"/"WIKIMANIA"), gestalt_paper ("GESTALT"/"GESTALT_PATTERN_MATCHING")
    - Category 3 — autojunk-sensitive (1 entry; load-bearing per Pitfall 2): autojunk_sensitive constructed as `"a" * 100 + "x" * 5 + "a" * 100` vs `"a" * 50 + "y" * 5 + "a" * 50 + "a" * 100` (or similar — ≥ 200 chars on at least one side, with character "a" exceeding 1% so autojunk=True would mark it as junk). Recommendation: copy the example from RESEARCH.md Code Examples line 844–845 verbatim
    - Category 4 — Substring + partial + unicode (4–6 entries to reach 15–18 total): substring_middle ("abcdef"/"xyzabcdefuvw"), partial_overlap ("kitten"/"sitting"), unicode_ascii_only ("café"/"cafe"), longer_identity ("the quick brown fox"/"the quick brown fox"). Add 1–2 more substring/partial cases as needed to hit the 15–18 range

    Define `def score_case(a: str, b: str) -> float:` — handle both-empty (return 1.0) and one-empty (return 0.0) BEFORE invoking difflib; otherwise return `difflib.SequenceMatcher(autojunk=False, a=a, b=b).ratio()`. The autojunk=False keyword MUST be the FIRST argument (load-bearing per Pitfall 2).

    Define `def _check_python_version() -> None:` — assert `sys.version_info >= _MIN_PYTHON_VERSION`; on failure `sys.exit(f"ERROR: Python {sys.version_info[:2]} < required {_MIN_PYTHON_VERSION}")`. Phase 3 IN-07 closure pattern.

    Define `def main() -> None:` — call `_check_python_version()`; loop CASES building entries dict `{"name": name, "a": a, "b": b, "difflib_ratio": score_case(a, b)}`; build top-level dict `{"version": 1, "python_version": f"{sys.version_info.major}.{sys.version_info.minor}.{sys.version_info.micro}", "entries": entries}`; write to `testdata/cross-validation/ratcliff-obershelp/vectors.json` with `json.dump(out, f, indent=2, sort_keys=False)` followed by `f.write("\n")` (trailing LF for byte-stability — same as gen-swg-cross-validation.py around line 216).

    Standard `if __name__ == "__main__": main()` guard at file end.

    Run the script ONCE on the local Python (must be ≥ 3.7) to generate testdata/cross-validation/ratcliff-obershelp/vectors.json. Verify by hand: file exists at the expected path, JSON parses, 15–18 entries, all 4 categories covered, autojunk_sensitive entry's difflib_ratio is in (0, 1) — not 0.0 or 1.0.

    Commit both scripts/gen-ratcliff-obershelp-cross-validation.py AND the generated testdata/cross-validation/ratcliff-obershelp/vectors.json in the same commit.
  </action>
  <verify>
    <automated>cd /Users/johnny/Development/fuzzymatch && python3 scripts/gen-ratcliff-obershelp-cross-validation.py && python3 -c "import json; d=json.load(open('testdata/cross-validation/ratcliff-obershelp/vectors.json')); assert d['version']==1; assert 15 <= len(d['entries']) <= 18; names=[e['name'] for e in d['entries']]; assert any('empty' in n for n in names); assert any(n in ('wikimedia_wikimania','gestalt_paper') for n in names); assert any('autojunk' in n for n in names); assert any('unicode' in n or 'cafe' in n for n in names); print('OK', len(d['entries']))" && grep -q "autojunk=False" scripts/gen-ratcliff-obershelp-cross-validation.py && grep -q "sys.version_info" scripts/gen-ratcliff-obershelp-cross-validation.py && ! grep -E "^import (Bio|requests|numpy|pandas)" scripts/gen-ratcliff-obershelp-cross-validation.py</automated>
  </verify>
  <done>
    Script exists; runs cleanly on Python 3.7+; generates the committed JSON corpus byte-stably. Corpus has 15–18 entries covering all 4 categories. autojunk=False present in the script; Python version check present; no third-party imports.
  </done>
</task>

<task type="auto" tdd="true">
  <name>Task 2: Append TestRatcliffObershelp_CrossValidation to ratcliff_obershelp_test.go</name>
  <files>ratcliff_obershelp_test.go</files>
  <read_first>
    - ratcliff_obershelp_test.go (current state — created in plan 04-03; read to understand the import block and Test* functions already present)
    - swg_test.go lines 411–479 (TestSWG_CrossValidation — the direct structural analog to copy)
    - testdata/cross-validation/ratcliff-obershelp/vectors.json (created in Task 1 — read to understand the JSON schema you're decoding)
    - .planning/phases/04-remaining-character-gestalt/04-CONTEXT.md §1 (tolerance 1e-9; 4 mandatory categories)
    - .planning/phases/04-remaining-character-gestalt/04-PATTERNS.md §"ratcliff_obershelp_test.go" Plan 04-04 appends section
    - .planning/phases/04-remaining-character-gestalt/04-VALIDATION.md (rows 04-04-02, 04-04-03, 04-04-04)
  </read_first>
  <behavior>
    - TestRatcliffObershelp_CrossValidation reads testdata/cross-validation/ratcliff-obershelp/vectors.json once, iterates every entry, and asserts `math.Abs(fuzzymatch.RatcliffObershelpScore(e.A, e.B) - e.DifflibRatio) <= 1e-9` for each entry. Use `t.Run(e.Name, func(t *testing.T) { ... })` for per-entry sub-test naming so failures point at the offending pair
    - Embedded subtest specifically asserts the autojunk_sensitive entry passes — labelled `t.Run("autojunk_sensitive", ...)` (VALIDATION.md row 04-04-04)
    - TestRatcliffObershelp_CrossValidation_CorpusShape asserts the corpus has 15–18 entries AND coverage of all 4 categories (presence checked by entry-name substring matching)
    - Test runs as part of default `go test ./...` — no Python at test time
  </behavior>
  <action>
    Append TWO test functions plus their supporting struct definitions to ratcliff_obershelp_test.go (after the unit tests from plan 04-03). Pattern: copy swg_test.go lines 411–479 (TestSWG_CrossValidation) structure verbatim, substituting names and field types per the <interfaces> section above.

    Struct definitions (place inside the test file near the new tests, or in a small type-block at the top of the file): `roCrossValidationEntry` with fields Name, A, B (strings), DifflibRatio (float64), each tagged with the matching json key from the corpus JSON. `roCrossValidationCorpus` with Version (int), PythonVersion (string), Entries ([]roCrossValidationEntry).

    `TestRatcliffObershelp_CrossValidation(t *testing.T)` body:
    - Open testdata/cross-validation/ratcliff-obershelp/vectors.json via `os.Open` (use `t.Fatalf` on error)
    - Decode via `json.NewDecoder(f).Decode(&corpus)`; `t.Fatalf` on error
    - Assert `corpus.Version == 1` (use `t.Fatalf` since this is a fixture-shape issue)
    - Assert `len(corpus.Entries) >= 15 && len(corpus.Entries) <= 18`
    - Loop entries: for each `e := range corpus.Entries` run a sub-test `t.Run(e.Name, func(t *testing.T) { got := fuzzymatch.RatcliffObershelpScore(e.A, e.B); if d := math.Abs(got - e.DifflibRatio); d > 1e-9 { t.Errorf("RatcliffObershelpScore(%q, %q) = %g; want %g (delta %g exceeds 1e-9)", e.A, e.B, got, e.DifflibRatio, d) } })`

    `TestRatcliffObershelp_CrossValidation_CorpusShape(t *testing.T)` body:
    - Open + decode the corpus (same pattern)
    - Assert `len(corpus.Entries) >= 15 && len(corpus.Entries) <= 18`
    - Build a slice of `names` from the entries; assert at least one name contains "empty" (Category 1); at least one name is "wikimedia_wikimania" OR "gestalt_paper" (Category 2); at least one contains "autojunk" (Category 3); at least one contains "unicode" OR "cafe" (Category 4). Use `strings.Contains`

    Stdlib testing only — NO testify. Use `t.Errorf` for non-fatal failures and `t.Fatalf` for setup failures (open / decode / corpus-shape).

    NOTE: the SWG cross-validation test name pattern in the existing file is `TestSWG_CrossValidation`. The RO equivalent is `TestRatcliffObershelp_CrossValidation` — exact name per VALIDATION.md row 04-04-03.
  </action>
  <verify>
    <automated>cd /Users/johnny/Development/fuzzymatch && go test -run 'TestRatcliffObershelp_CrossValidation$|TestRatcliffObershelp_CrossValidation_CorpusShape$' -v ./... && go test -run 'TestRatcliffObershelp_CrossValidation/autojunk_sensitive' -v ./...</automated>
  </verify>
  <done>
    TestRatcliffObershelp_CrossValidation green: every corpus entry passes within 1e-9; the autojunk_sensitive sub-test passes specifically. TestRatcliffObershelp_CrossValidation_CorpusShape green: corpus has 15–18 entries covering all 4 categories.
  </done>
</task>

<task type="auto">
  <name>Task 3: Makefile regen target + CONTRIBUTING.md doc line</name>
  <files>Makefile, CONTRIBUTING.md</files>
  <read_first>
    - Makefile (current state — read lines 25–35 area for .PHONY and lines 195–215 area for the regen-swg-cross-validation target)
    - CONTRIBUTING.md (current state — read line ~92 for the regen-swg-cross-validation entry that the new entry mirrors)
    - .planning/phases/04-remaining-character-gestalt/04-PATTERNS.md §"Makefile (append target in Plan 04-04)", §"CONTRIBUTING.md (append doc line in Plan 04-04)"
    - .planning/phases/04-remaining-character-gestalt/04-VALIDATION.md (row 04-04-05)
    - makefile_targets_test.go (if it exists in the repo — read TestMakefile_TargetsDocumentedInContributing to understand what shape the meta-test expects)
  </read_first>
  <acceptance_criteria>
    - `make regen-ratcliff-obershelp-cross-validation` exists in Makefile and exits 0 on a machine with Python 3.7+
    - The target emits a clear error and exits non-zero if python3 is not on PATH (shell-gate via `command -v python3`)
    - The new target name appears in the Makefile's .PHONY declaration alongside the existing `regen-swg-cross-validation`
    - CONTRIBUTING.md contains a line documenting the new target (matching the shape of the existing regen-swg-cross-validation entry around line 92)
    - `grep -q "^regen-ratcliff-obershelp-cross-validation:" Makefile` exits 0
    - `grep -q "regen-ratcliff-obershelp-cross-validation" CONTRIBUTING.md` exits 0
    - `go test -run TestMakefile_TargetsDocumentedInContributing ./...` exits 0 (if this test exists in the repo)
  </acceptance_criteria>
  <action>
    Append a `regen-ratcliff-obershelp-cross-validation` target to Makefile, immediately after the existing `regen-swg-cross-validation` target. Pattern per PATTERNS.md §"Makefile (append target in Plan 04-04)": shell-gate on `command -v python3` (emit "python3 not found; install Python 3.7+" and exit 1 if absent); otherwise invoke `python3 scripts/gen-ratcliff-obershelp-cross-validation.py`. The recipe body uses the SAME shell-gate idiom as the SWG target (Makefile lines 196–211 — copy verbatim and substitute the script path). Preserve TAB indentation on recipe lines (Makefile syntax requires).

    Also extend the `.PHONY` declaration at the top of the Makefile (Phase 3 added `regen-swg-cross-validation`; append `regen-ratcliff-obershelp-cross-validation` alongside on the same .PHONY line).

    Append a documentation line to CONTRIBUTING.md adjacent to the existing `make regen-swg-cross-validation` entry (line ~92 area). Use the exact phrasing from PATTERNS.md §"CONTRIBUTING.md (append doc line in Plan 04-04)": developer-only, regenerates the difflib cross-validation corpus path, Python 3.7+ required (difflib is stdlib — no pip install needed). Place the new bullet under the same heading as the SWG entry so the meta-test's grep finds it.

    Both edits land in the SAME commit per PATTERNS.md gotcha note (the makefile_targets_test.go meta-test will fail otherwise). If makefile_targets_test.go does not exist in the repo, this task is still required for human-readable contributor docs.
  </action>
  <verify>
    <automated>cd /Users/johnny/Development/fuzzymatch && grep -q "^regen-ratcliff-obershelp-cross-validation:" Makefile && grep -q "regen-ratcliff-obershelp-cross-validation" CONTRIBUTING.md && make regen-ratcliff-obershelp-cross-validation && (test ! -f makefile_targets_test.go || go test -run TestMakefile_TargetsDocumentedInContributing ./...)</automated>
  </verify>
  <done>
    Makefile has the new target with shell-gated python3 check; .PHONY updated. CONTRIBUTING.md has the documentation entry. `make regen-ratcliff-obershelp-cross-validation` re-runs the generator producing the same byte-identical corpus. Meta-test passes if present.
  </done>
</task>

</tasks>

<threat_model>
## Trust Boundaries

| Boundary | Description |
|----------|-------------|
| developer → generator script | Developer runs `make regen-ratcliff-obershelp-cross-validation` locally; produces committed JSON. NOT a CI execution path |
| CI / test → committed JSON | TestRatcliffObershelp_CrossValidation reads testdata/cross-validation/ratcliff-obershelp/vectors.json at test time; ZERO Python invocation in CI |

## STRIDE Threat Register (ASVS Level 1)

| Threat ID | Category | Component | Disposition | Mitigation Plan |
|-----------|----------|-----------|-------------|-----------------|
| T-fixture-tampering | T (Tampering of cross-validation fixture) | testdata/cross-validation/ratcliff-obershelp/vectors.json | mitigate | Corpus is committed to git; any modification surfaces in PR review; the regen target is developer-only and reproducible (byte-stable output from the same Python version). Phase 3 SWG pattern proves the approach (16 entries matching biopython with delta=0). The autojunk_sensitive entry is the keystone — any silent autojunk=True regression in the generator surfaces immediately as TestRatcliffObershelp_CrossValidation/autojunk_sensitive failure |
| T-fuzz-panic | D (Denial of Service via panic on malformed input) | RatcliffObershelpScore (called by cross-validation test) | mitigate | Already mitigated in plan 04-03 via FuzzRatcliffObershelpScore (Phase 3 WR-02 closure). Cross-validation test re-exercises the public surface on Dr. Dobb's pairs and the autojunk-sensitive 200+char case — additional smoke coverage |
| T-float-determinism | T (Tampering of float reduction order across architectures) | TestRatcliffObershelp_CrossValidation comparison | mitigate | `math.Abs(got - e.DifflibRatio) <= 1e-9` per-entry assertion is the byte-stability gate. The tolerance is the Phase 3 epsilon convention (matches SWG cross-validation). Any cross-platform float drift > 1e-9 fails the gate. CI matrix verifies on linux/amd64, linux/arm64, darwin/arm64, windows/amd64 |
</threat_model>

<verification>
- `python3 scripts/gen-ratcliff-obershelp-cross-validation.py` exits 0 on Python 3.7+ and produces testdata/cross-validation/ratcliff-obershelp/vectors.json.
- The generated JSON has `version: 1`, populated `python_version`, and 15–18 entries spanning all 4 categories.
- `go test -run 'TestRatcliffObershelp_CrossValidation' -v ./...` exits 0; every corpus entry's sub-test green within 1e-9; the autojunk_sensitive sub-test specifically green.
- `go test -run 'TestRatcliffObershelp_CrossValidation_CorpusShape' ./...` exits 0.
- `make regen-ratcliff-obershelp-cross-validation` exits 0 and reproduces the same JSON byte-stably.
- `grep -q "^regen-ratcliff-obershelp-cross-validation:" Makefile` and `grep -q "regen-ratcliff-obershelp-cross-validation" CONTRIBUTING.md` both exit 0.
- `! grep -E "^import (Bio|requests|numpy|pandas)" scripts/gen-ratcliff-obershelp-cross-validation.py` (zero third-party Python imports).
- `grep -q "autojunk=False" scripts/gen-ratcliff-obershelp-cross-validation.py` (Pitfall 2 load-bearing keyword present).
- `grep -q "sys.version_info" scripts/gen-ratcliff-obershelp-cross-validation.py` (Python 3.7+ version check present per Phase 3 IN-07 closure).
- `bash scripts/verify-license-headers.sh` exits 0 (if the script covers .py files).
- `make check` exits 0 at end of plan.
</verification>

<success_criteria>
- All three tasks complete; all listed verification commands green.
- testdata/cross-validation/ratcliff-obershelp/vectors.json committed; 15–18 entries; 4 categories.
- scripts/gen-ratcliff-obershelp-cross-validation.py committed; Python 3.7+ stdlib-only; autojunk=False.
- TestRatcliffObershelp_CrossValidation + TestRatcliffObershelp_CrossValidation_CorpusShape green; autojunk_sensitive sub-test specifically green.
- Makefile + CONTRIBUTING.md updated together; meta-test passes (if present).
- Plan 04-05 (finalisation) can begin — the algorithm-correctness-reviewer gate is closed.
</success_criteria>

<output>
After completion, create `.planning/phases/04-remaining-character-gestalt/04-04-ratcliff-obershelp-cross-validation-SUMMARY.md` per the GSD summary template.
</output>
