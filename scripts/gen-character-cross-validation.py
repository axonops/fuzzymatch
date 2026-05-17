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

"""scripts/gen-character-cross-validation.py — jellyfish==1.2.1 cross-validation
corpus generator for the four Phase-2 character-tier algorithms.

Regenerates testdata/cross-validation/character/vectors.json by running
``jellyfish.{levenshtein_distance, jaro_similarity, jaro_winkler_similarity,
damerau_levenshtein_distance}`` against a fixed, deterministic list of input
pairs (CASES). Each entry carries all four scores: levenshtein_distance,
damerau_levenshtein_osa_distance, jaro_similarity, jaro_winkler_similarity.
The generated JSON is the verification fixture consumed by the Go test
``TestCharacter_CrossValidation`` (in character_cross_validation_test.go) —
CI does NOT require Python at test time; the committed JSON is the
verification fixture read by the Go test directly.

Reference implementation: jellyfish (BSD-2-Clause), maintained by James
Turk. Jellyfish is used here SOLELY for reference-vector cross-validation
per the project's .claude/skills/algorithm-licensing-standards — NO
jellyfish code is copied into the fuzzymatch implementation. The Go-side
algorithm kernels in levenshtein.go / damerau_osa.go / jaro.go /
jarowinkler.go are fresh transcriptions from their primary academic
sources (Levenshtein 1965 / Damerau 1964 / Boytsov 2011 / Jaro 1989 /
Winkler 1990); this script provides independent reference scores that
close the algorithm-correctness-reviewer cross-validation gate per
CLAUDE.md "Workflow — Agent Gates".

jellyfish-vs-fuzzymatch surface mapping:
    - jellyfish.levenshtein_distance(a, b)
        ↔ fuzzymatch.LevenshteinDistance(a, b) — integer (exact equality).
    - jellyfish.damerau_levenshtein_distance(a, b)
        ↔ fuzzymatch.DamerauLevenshteinFullDistance(a, b)
        IMPORTANT: jellyfish 1.2.1's damerau_levenshtein_distance ships
        the UNRESTRICTED (Lowrance-Wagner 1975, "true") Damerau-Levenshtein
        variant, NOT the OSA (Optimal String Alignment / restricted)
        variant. Direct probe on the pinned release confirms:
        jellyfish.damerau_levenshtein_distance("ca", "abc") == 2 (true
        Damerau answer with a transposition + insert); whereas OSA
        returns 3 because the transposition+touch is forbidden. Per the
        Boytsov 2011 §3.1 discriminating vector. The fuzzymatch corpus
        therefore cross-validates jellyfish's output against
        DamerauLevenshteinFullDistance (the unrestricted surface).
        The OSA surface (DamerauLevenshteinOSADistance) is covered by
        primary-source reference vectors in damerau_osa_test.go and is
        NOT part of this corpus.
        Integer (exact equality).
    - jellyfish.jaro_similarity(a, b)
        ↔ fuzzymatch.JaroScore(a, b) — float within 1e-9 tolerance.
        NOTE: jellyfish.jaro_similarity("", "") returns 0.0 (both-empty
        treated as 0-similarity); fuzzymatch.JaroScore("", "") returns
        1.0 by documented catalogue convention (vacuous-identity — both
        strings empty, nothing disagrees). This is a known variant
        divergence; the corpus marks the affected entry with
        variant_divergence=True and emits divergent_jellyfish_value
        alongside the fuzzymatch-expected value. Same pattern as
        gen-phonetic-cross-validation.py for Soundex/NYSIIS variants.
    - jellyfish.jaro_winkler_similarity(a, b)
        ↔ fuzzymatch.JaroWinklerScore(a, b) — float within 1e-9
        tolerance. Same both-empty divergence as Jaro.

CRITICAL — JELLYFISH_VERSION pin gate:
    The Metaphone-style "pip show jellyfish" version probe is needed
    because jellyfish 1.2.1 dropped the __version__ module attribute
    (confirmed by `python3 -c "import jellyfish; jellyfish.__version__"`
    raising AttributeError on the pinned release). The script's first
    action after _check_python_version() is _assert_jellyfish_version()
    which shells out to `pip show jellyfish` and parses the Version: line.
    A version mismatch is REFUSED with a clear error message; the corpus
    is locked to a single jellyfish release for byte-stable cross-
    validation. Bumping requires (a) updating JELLYFISH_VERSION here,
    (b) updating the install hint in Makefile's
    regen-character-cross-validation target, (c) updating CONTRIBUTING.md
    if the pin is shared with gen-phonetic-cross-validation.py,
    (d) re-running this script, (e) committing the regenerated
    vectors.json, and (f) running
    ``go test -run TestCharacter_CrossValidation -race ./...``.

Tokeniser-divergence handling:
    None — the four algorithms are character-level (no tokenise), so the
    Tokenise-safety constraint that applies to TokenSort / TokenSet
    cross-validation does NOT apply here. Inputs may be mixed-case,
    contain separators, contain Unicode, etc. The Go-side surface
    operates on the raw input string; jellyfish operates on the raw
    Python string (codepoint-indexed). For ASCII inputs the
    fuzzymatch byte path and jellyfish output coincide; for non-ASCII
    inputs the Go test calls the rune-aware variant where the
    fuzzymatch surface differs.

CASE coverage (matches plan 08.5-10 task 1 acceptance criteria —
≥ 30 cases):
    - short ASCII identity / mismatched / one-char edits
    - identity / both-empty / one-empty pairs
    - Jaro / JaroWinkler canonical reference vectors (MARTHA/MARHTA,
      DIXON/DICKSONX, DWAYNE/DUANE) — Winkler 1990 / Jaro 1989 paper
      examples
    - Damerau-OSA transposition fixtures (ca/abc — Boytsov 2011 §2
      worked example)
    - mid-length and long ASCII pairs to exercise DP table growth
    - mismatched-length pairs to exercise asymmetric DP

Usage:
    make regen-character-cross-validation
    # or directly:
    python3 scripts/gen-character-cross-validation.py

Requirements:
    - Python 3.7+ (for guaranteed dict insertion-order preservation in
      json.dump). Enforced at script entry — see _check_python_version().
    - jellyfish==1.2.1. Install:
      ``python3 -m pip install --user jellyfish==1.2.1``.

Determinism:
    The CASES list owns the corpus's deterministic ordering. Python 3.7+
    guarantees dict insertion-order preservation; ``json.dump`` with
    ``sort_keys=False`` honours that order. The _metadata block uses
    deterministic Python / jellyfish version strings — the regenerated_at
    ISO timestamp is the only environment-dependent field, and the
    Go-side loader does not assert byte-stability on the corpus, only on
    the per-entry scores; the timestamp is informational for triage.
"""

from datetime import datetime, timezone
import json
import os
import subprocess
import sys

# Minimum supported Python version. Python 3.7 was the first release to
# guarantee dict insertion-order preservation as a language feature.
_MIN_PYTHON_VERSION = (3, 7)

# Pinned jellyfish version. Bumping requires the six-step protocol
# documented in the module docstring above (and in CONTRIBUTING.md).
JELLYFISH_VERSION = "1.2.1"


def _check_python_version() -> None:
    """Guard against silent corpus drift on outdated Python (< 3.7)."""
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


_check_python_version()


def _assert_jellyfish_version() -> None:
    """Assert the installed jellyfish package matches JELLYFISH_VERSION.

    jellyfish 1.2.1 does NOT expose __version__ as a module attribute
    (confirmed by direct probe on the pinned release), so we check via
    `pip show jellyfish` and parse the Version: field. This mirrors the
    Metaphone-version probe in gen-phonetic-cross-validation.py.
    """
    try:
        out = subprocess.check_output(
            [sys.executable, "-m", "pip", "show", "jellyfish"],
            stderr=subprocess.DEVNULL,
        ).decode("utf-8", errors="replace")
    except subprocess.CalledProcessError:
        sys.exit(
            f"ERROR: 'pip show jellyfish' failed — jellyfish package not "
            f"installed. Run: python3 -m pip install --user "
            f"jellyfish=={JELLYFISH_VERSION}"
        )
    version_line = next(
        (ln for ln in out.splitlines() if ln.lower().startswith("version:")),
        None,
    )
    if version_line is None:
        sys.exit(
            "ERROR: could not parse 'pip show jellyfish' output — no "
            "Version: line found."
        )
    installed = version_line.split(":", 1)[1].strip()
    if installed != JELLYFISH_VERSION:
        sys.exit(
            f"jellyfish version mismatch: installed {installed!r}, "
            f"script pinned to {JELLYFISH_VERSION!r} — "
            f"run: python3 -m pip install --user "
            f"jellyfish=={JELLYFISH_VERSION}"
        )


_assert_jellyfish_version()

import jellyfish  # noqa: E402 — imported after version constant for clear error message


# CASES — the deterministic ordered list of cross-validation entries.
#
# Format: a dict per entry. Required keys:
#   - name: stable identifier (referenced in test failure messages).
#   - a, b: input strings.
#
# Coverage (≥ 30 cases — see module docstring for the full rubric):
CASES = [
    # Identity / empty / one-empty.
    {"name": "identity_short", "a": "hello", "b": "hello"},
    {"name": "identity_word", "a": "fuzzymatch", "b": "fuzzymatch"},
    {"name": "both_empty", "a": "", "b": ""},
    {"name": "one_empty_a", "a": "", "b": "hello"},
    {"name": "one_empty_b", "a": "hello", "b": ""},

    # Levenshtein canonical reference vectors.
    {"name": "lev_kitten_sitting", "a": "kitten", "b": "sitting"},
    {"name": "lev_saturday_sunday", "a": "Saturday", "b": "Sunday"},
    {"name": "lev_book_back", "a": "book", "b": "back"},
    {"name": "lev_single_insert", "a": "abc", "b": "abcd"},
    {"name": "lev_single_delete", "a": "abcd", "b": "abc"},
    {"name": "lev_single_sub", "a": "abc", "b": "abd"},

    # Damerau (unrestricted / true Damerau-Levenshtein) transposition
    # fixtures. The dl_full_ prefix reflects that jellyfish's
    # damerau_levenshtein_distance is the UNRESTRICTED variant — see
    # module docstring "jellyfish-vs-fuzzymatch surface mapping". The
    # dl_full_boytsov_example fixture is load-bearing: it surfaces the
    # OSA-vs-full discriminating vector ("ca" vs "abc" → OSA=3, full=2)
    # so a future jellyfish regression to OSA semantics fails loudly.
    {"name": "dl_full_transpose_ab", "a": "ab", "b": "ba"},
    {"name": "dl_full_transpose_word", "a": "the cat", "b": "the act"},
    {"name": "dl_full_boytsov_example", "a": "ca", "b": "abc"},
    {"name": "dl_full_two_transpositions", "a": "abcd", "b": "badc"},

    # Jaro canonical reference vectors (Jaro 1989 / Winkler 1990 paper).
    {"name": "jaro_martha_marhta", "a": "MARTHA", "b": "MARHTA"},
    {"name": "jaro_dixon_dicksonx", "a": "DIXON", "b": "DICKSONX"},
    {"name": "jaro_dwayne_duane", "a": "DWAYNE", "b": "DUANE"},
    {"name": "jaro_jellyfish_smellyfish", "a": "jellyfish", "b": "smellyfish"},

    # JaroWinkler-specific (common-prefix bonus matters).
    {"name": "jw_short_prefix_match", "a": "TRATE", "b": "TRACE"},
    {"name": "jw_long_prefix_match", "a": "TROON", "b": "TROOPER"},
    {"name": "jw_no_prefix_match", "a": "ABCDE", "b": "ZBCDE"},

    # Mismatched-length and asymmetric DP exercise.
    {"name": "len_mismatch_short_long", "a": "abc", "b": "abcdefghij"},
    {"name": "len_mismatch_long_short", "a": "abcdefghij", "b": "abc"},
    {"name": "len_off_by_two", "a": "abc", "b": "axbxc"},

    # Disjoint / fully-different.
    {"name": "fully_different_equal_len", "a": "aaaa", "b": "bbbb"},
    {"name": "fully_different_short", "a": "x", "b": "y"},

    # Mid-length and long ASCII pairs (DP-table growth).
    {"name": "mid_length_close_match", "a": "the quick brown fox", "b": "the quick brwon fox"},
    {"name": "long_ascii_close", "a": "the quick brown fox jumps over the lazy dog", "b": "the quick brown fox leaps over the lazy dog"},
    {"name": "long_ascii_far", "a": "the quick brown fox jumps over the lazy dog", "b": "all work and no play makes jack a dull boy"},

    # Single-character pairs.
    {"name": "single_char_match", "a": "a", "b": "a"},
    {"name": "single_char_mismatch", "a": "a", "b": "b"},

    # Case-sensitivity exposure (Jaro/JaroWinkler are case-sensitive in
    # jellyfish — the Go surface mirrors this for the byte path).
    {"name": "case_diff_upper_lower", "a": "HELLO", "b": "hello"},
]


def score_case(case: dict) -> dict:
    """Compute the four jellyfish reference scores for a single case.

    Empty-input handling — jellyfish vs fuzzymatch divergence:
        - jellyfish.jaro_similarity("", "") returns 0.0 (both-empty
          treated as 0-similarity); fuzzymatch.JaroScore("", "") returns
          1.0 by documented catalogue convention (vacuous-identity).
        - Same divergence applies to JaroWinkler.
        - Levenshtein and Damerau distances on empty strings return the
          length of the non-empty string (standard; both libraries agree).
        - When the divergence applies (a == "" AND b == ""), the entry's
          jaro_similarity and jaro_winkler_similarity fields carry the
          FUZZYMATCH-expected values (1.0, 1.0); the jellyfish raw values
          are emitted alongside as divergent_jellyfish_jaro and
          divergent_jellyfish_jw, and variant_divergence=True is set on
          the entry. The Go-side loader test asserts against the
          fuzzymatch-expected value (NOT the jellyfish value), so the
          divergence is transparent at the corpus boundary rather than
          spuriously failing.
    """
    a, b = case["a"], case["b"]
    jaro_jf = jellyfish.jaro_similarity(a, b)
    jw_jf = jellyfish.jaro_winkler_similarity(a, b)

    # Both-empty Jaro/JW divergence: fuzzymatch returns 1.0; jellyfish
    # returns 0.0. Use the catalogue convention for the expected value.
    if a == "" and b == "":
        jaro_expected = 1.0
        jw_expected = 1.0
        variant_divergence = True
    else:
        jaro_expected = jaro_jf
        jw_expected = jw_jf
        variant_divergence = False

    entry = {
        "name": case["name"],
        "a": a,
        "b": b,
        "levenshtein_distance": jellyfish.levenshtein_distance(a, b),
        "damerau_levenshtein_distance": jellyfish.damerau_levenshtein_distance(a, b),
        "jaro_similarity": jaro_expected,
        "jaro_winkler_similarity": jw_expected,
    }
    if variant_divergence:
        entry["variant_divergence"] = True
        entry["divergent_jellyfish_jaro"] = jaro_jf
        entry["divergent_jellyfish_jw"] = jw_jf
    return entry


def main() -> None:
    entries = [score_case(c) for c in CASES]

    out = {
        "version": 1,
        "_metadata": {
            "jellyfish_version": JELLYFISH_VERSION,
            "python_version": f"{sys.version_info.major}.{sys.version_info.minor}.{sys.version_info.micro}",
            "regenerated_at": datetime.now(timezone.utc).isoformat(),
        },
        "entries": entries,
    }
    path = "testdata/cross-validation/character/vectors.json"
    os.makedirs(os.path.dirname(path), exist_ok=True)
    with open(path, "w") as f:
        json.dump(out, f, indent=2, sort_keys=False)
        f.write("\n")  # trailing LF matches Phase 2 golden-file canonical convention


if __name__ == "__main__":
    main()
