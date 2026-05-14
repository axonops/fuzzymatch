#!/usr/bin/env python3
# Copyright 2026 AxonOps Limited
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
#     http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.

"""scripts/gen-ratcliff-obershelp-cross-validation.py — Ratcliff-Obershelp
cross-validation corpus generator.

Regenerates testdata/cross-validation/ratcliff-obershelp/vectors.json by
running Python's stdlib ``difflib.SequenceMatcher(autojunk=False).ratio()``
on a fixed, deterministic list of input pairs (CASES). The generated JSON
is the verification fixture consumed by the Go test
``TestRatcliffObershelp_CrossValidation`` (in ratcliff_obershelp_test.go)
— CI does NOT require Python at test time; the committed JSON is read by
the Go test directly.

Reference implementation: Python stdlib ``difflib`` (PSF licence). difflib
is used here SOLELY for reference-vector cross-validation per the project's
``.claude/skills/algorithm-licensing-standards`` — no difflib code is copied
into the fuzzymatch implementation. Our Ratcliff-Obershelp kernel in
ratcliff_obershelp.go is a fresh transcription from Ratcliff & Metzener 1988
(Dr. Dobb's Journal 13(7):46-51); this script provides independent reference
scores that close the algorithm-correctness-reviewer cross-validation gate
per CLAUDE.md §"Workflow — Agent Gates".

CRITICAL — autojunk=False is REQUIRED:
    difflib's DEFAULT ``autojunk=True`` is a performance heuristic — when
    ``len(b) >= 200`` it marks characters appearing in ≥ 1% of positions as
    "junk" and excludes them from matching. This is NOT the
    Ratcliff-Obershelp algorithm — it's a difflib speed optimisation that
    distorts scores. The TRUE Ratcliff-Obershelp algorithm has
    autojunk=False. The corpus MUST be generated with autojunk=False;
    accidentally regenerating with autojunk=True silently destroys every
    score in the corpus (the 200+-char ``autojunk_sensitive`` entry exists
    specifically to surface this regression).

Byte-path semantics:
    Our Go ``RatcliffObershelpScore`` operates on byte slices (UTF-8 bytes),
    not codepoints. To produce a corpus that matches the Go byte path
    byte-for-byte for both ASCII and multi-byte UTF-8 entries, we encode
    each input string to UTF-8 bytes BEFORE calling difflib. For ASCII
    entries this is a no-op; for the ``unicode_ascii_only`` (café/cafe)
    entry this matters — bytes give 6/9 = 0.6666..., codepoints give 6/8
    = 0.75. The Go test calls the byte-path function, so the corpus must
    match the byte-path semantic.

Usage:
    make regen-ratcliff-obershelp-cross-validation
    # or directly:
    python3 scripts/gen-ratcliff-obershelp-cross-validation.py

Requirements:
    - Python 3.7+ (for guaranteed dict insertion-order preservation in
      json.dump, which is load-bearing for the corpus's byte-stable
      output). Enforced at script entry — see _check_python_version().
    - difflib (Python stdlib — NO pip install needed; this is the
      structural simplification over Phase 3's biopython dependency).

Determinism:
    The CASES list owns the corpus's deterministic ordering. Python 3.7+
    guarantees dict insertion-order preservation, and ``json.dump`` with
    ``sort_keys=False`` honours that order. No randomness, no environment
    inputs, no system clocks — the script produces byte-identical output
    on repeated runs for a given Python version.
"""

import difflib
import json
import os
import sys

# Minimum supported Python version. Python 3.7 was the first release to
# guarantee dict insertion-order preservation as a language feature (it was
# a CPython implementation detail in 3.6). The corpus's byte-stable output
# relies on that guarantee — older Pythons could produce reordered JSON.
# Enforced at script entry — see _check_python_version().
_MIN_PYTHON_VERSION = (3, 7)

# CASES — the deterministic ordered list of cross-validation entries.
#
# Format: (name, a, b). Ratcliff-Obershelp has no per-case parameters
# (unlike Phase 3's SWG with its match/mismatch/gap_open/gap_extend block).
#
# Required entries spanning ALL FOUR mandatory CONTEXT.md §1 categories:
#
# Category 1 — Standard edge cases (5 entries):
#   identity_short, both_empty, one_empty_a, one_empty_b, no_overlap
#
# Category 2 — Dr. Dobb's 1988 paper examples (2 entries):
#   wikimedia_wikimania, gestalt_paper
#
# Category 3 — autojunk-sensitive (1 entry — load-bearing per RESEARCH.md
# Pitfall 2; the corpus's keystone proof that autojunk=False is correctly
# disabled). Total length ≥ 200 chars on at least one side; character 'a'
# appears in well over 1% of positions on both sides.
#
# Category 4 — Substring / partial / unicode (8 entries):
#   substring_middle, partial_overlap, unicode_ascii_only, longer_identity,
#   substring_at_start, substring_at_end, single_char_match, near_identical
#
# Total: 16 entries.
_AUTOJUNK_A = "a" * 100 + "x" * 5 + "a" * 100
_AUTOJUNK_B = "a" * 50 + "y" * 5 + "a" * 50 + "a" * 100

CASES = [
    # Category 1 — Standard edge cases.
    ("identity_short",       "hello",          "hello"),
    ("both_empty",           "",               ""),
    ("one_empty_a",          "",               "abcdef"),
    ("one_empty_b",          "abcdef",         ""),
    ("no_overlap",           "qqqq",           "zzzz"),

    # Category 2 — Dr. Dobb's 1988 paper examples.
    ("wikimedia_wikimania",  "WIKIMEDIA",      "WIKIMANIA"),
    ("gestalt_paper",        "GESTALT",        "GESTALT_PATTERN_MATCHING"),

    # Category 3 — autojunk-sensitive (200+ chars; load-bearing).
    ("autojunk_sensitive",   _AUTOJUNK_A,      _AUTOJUNK_B),

    # Category 4 — Substring / partial / unicode.
    ("substring_middle",     "abcdef",         "xyzabcdefuvw"),
    ("partial_overlap",      "kitten",         "sitting"),
    ("unicode_ascii_only",   "café",           "cafe"),
    ("longer_identity",      "the quick brown fox", "the quick brown fox"),
    ("substring_at_start",   "hello",          "hello_world"),
    ("substring_at_end",     "world",          "hello_world"),
    ("single_char_match",    "a",              "a"),
    ("near_identical",       "kitten",         "kittin"),
]


def score_case(a: str, b: str) -> float:
    """Compute the difflib reference ratio for a single case.

    Both-empty and one-empty cases are handled BEFORE invoking difflib so
    the conventions match our Go-side short-circuits in
    ratcliff_obershelp.go (RatcliffObershelpScore("","") = 1.0,
    Score("", non-empty) = 0.0). difflib.SequenceMatcher's behaviour on
    empty inputs is consistent with this convention, but we hard-wire the
    short-circuit here so the corpus values cannot drift if difflib's
    edge-case behaviour ever changes.

    For non-empty cases, the inputs are encoded to UTF-8 BYTES before
    being passed to difflib. This is load-bearing for the unicode_ascii_only
    entry — our Go ``RatcliffObershelpScore`` is byte-path
    (`café`/`cafe` → 6/9 = 0.6666...); calling difflib on Python strings
    would yield 6/8 = 0.75 (codepoint-path), breaking byte-for-byte
    equivalence. ASCII entries are unaffected: UTF-8 encoding of ASCII is
    identity. autojunk=False is REQUIRED — see module docstring.
    """
    if a == "" and b == "":
        # Both-empty identity: matches our Go-side `a == b` short-circuit in
        # RatcliffObershelpScore (returns 1.0). difflib's ratio is also 1.0
        # for two empty sequences, but we hard-wire to keep the corpus
        # contract explicit.
        return 1.0
    if a == "" or b == "":
        # One-empty: matches our Go-side `la == 0 || lb == 0` short-circuit.
        return 0.0

    a_bytes = a.encode("utf-8")
    b_bytes = b.encode("utf-8")
    # autojunk=False is REQUIRED — load-bearing per RESEARCH.md Pitfall 2.
    # The kwarg is intentionally the FIRST argument in the call so grep
    # gates on "autojunk=False" stay anchored to this line.
    return difflib.SequenceMatcher(autojunk=False, a=a_bytes, b=b_bytes).ratio()


def _check_python_version() -> None:
    """Guard the script against silent corpus drift on outdated Python.

    Python 3.6 had dict insertion-order preservation as a CPython
    implementation detail; 3.7 promoted it to a language guarantee that
    other implementations (PyPy, etc.) must honour. The corpus relies on
    this guarantee for byte-stable JSON output across Python versions and
    implementations — without it, a developer running on 3.6 or earlier
    could regenerate a corpus that no longer matches the committed
    reference, and the Go-side TestRatcliffObershelp_CrossValidation would
    fail with a misleading delta rather than a version-mismatch error.

    Phase 3 IN-07 closure — same pattern as gen-swg-cross-validation.py's
    _check_biopython_version helper.
    """
    if sys.version_info < _MIN_PYTHON_VERSION:
        min_str = ".".join(str(x) for x in _MIN_PYTHON_VERSION)
        cur_str = ".".join(str(x) for x in sys.version_info[:3])
        sys.exit(
            f"ERROR: Python {cur_str} is older than the supported minimum "
            f"{min_str}. Dict insertion-order preservation became a language "
            f"guarantee in Python 3.7; older releases could produce reordered "
            f"JSON breaking corpus byte-stability. Upgrade to Python "
            f">= {min_str}."
        )


def main() -> None:
    _check_python_version()
    entries = []
    for name, a, b in CASES:
        entries.append({
            "name": name,
            "a": a,
            "b": b,
            "difflib_ratio": score_case(a, b),
        })

    out = {
        "version": 1,
        "python_version": f"{sys.version_info.major}.{sys.version_info.minor}.{sys.version_info.micro}",
        "entries": entries,
    }
    path = "testdata/cross-validation/ratcliff-obershelp/vectors.json"
    os.makedirs(os.path.dirname(path), exist_ok=True)
    with open(path, "w") as f:
        json.dump(out, f, indent=2, sort_keys=False)
        f.write("\n")  # trailing LF matches Phase 2 golden-file canonical convention


if __name__ == "__main__":
    main()
