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

"""scripts/gen-swg-cross-validation.py — Smith-Waterman-Gotoh cross-validation
corpus generator.

Regenerates testdata/cross-validation/swg/vectors.json by running biopython's
``Bio.Align.PairwiseAligner`` (mode="local") on a fixed, deterministic list of
input pairs (CASES). The generated JSON is the verification fixture consumed
by the Go test ``TestSWG_CrossValidation`` (in swg_test.go) — CI does NOT
require Python or biopython at test time; the committed JSON is read by the
Go test directly.

Reference implementation: biopython 1.85+ (BSD-3-Clause licensed). biopython
is used here SOLELY for reference-vector cross-validation per the project's
``.claude/skills/algorithm-licensing-standards`` — no biopython code is copied
into the fuzzymatch implementation. Our Smith-Waterman-Gotoh kernel in swg.go
is a fresh transcription of the corrected Flouri et al. 2015 recurrence; this
script provides independent reference scores that close the
algorithm-correctness-reviewer cross-validation gate per CLAUDE.md §"Workflow
— Agent Gates".

Usage:
    make regen-swg-cross-validation
    # or directly:
    python3 scripts/gen-swg-cross-validation.py

Requirements:
    - Python 3.7+ (for guaranteed dict insertion-order preservation in
      json.dump, which is load-bearing for the corpus's byte-stable output).
    - biopython 1.85+ — install via:
          python3 -m pip install --user biopython

Normalisation note:
    Each entry exposes BOTH the raw biopython alignment score
    (``biopython_score``) AND the script-side normalised reference
    (``biopython_normalised`` = clamp(raw / min(len(a), len(b)), 0, 1)).
    The Go test compares against ``biopython_normalised`` with zero in-Go
    normalisation logic; this means the Python script owns the normalisation
    reference and the Go test is a pure equality check within 1e-9 tolerance.
    See ``.planning/phases/03-smith-waterman-gotoh/03-CONTEXT.md`` §1.

Determinism:
    The CASES list owns the corpus's deterministic ordering. Python 3.7+
    guarantees dict insertion-order preservation, and ``json.dump`` with
    ``sort_keys=False`` honours that order. No randomness, no environment
    inputs, no system clocks — the script produces byte-identical output on
    repeated runs for a given biopython version.
"""

import json
import os
import sys

import Bio
from Bio.Align import PairwiseAligner

# Minimum supported biopython version. PairwiseAligner's behaviour on the
# affine-gap kernel stabilised at 1.85; older releases (notably 1.79) used
# different defaults and could silently produce scores that diverge from the
# committed corpus. Enforced at script entry — see _check_biopython_version().
_MIN_BIOPYTHON_VERSION = (1, 85)

# Default SWG parameters — must match Go-side ``NewSWGParams()`` in swg.go
# byte-for-byte (Match=1.0, Mismatch=-1.0, GapOpen=-1.5, GapExtend=-0.5).
DEFAULT_PARAMS = {
    "match": 1.0,
    "mismatch": -1.0,
    "gap_open": -1.5,
    "gap_extend": -0.5,
}

# CASES — the deterministic ordered list of cross-validation entries.
#
# Format: (name, a, b, params_override). params_override is None for default
# params and a dict otherwise.
#
# Required entries (CONTEXT.md §1 + 03-PATTERNS.md §unique-files-1):
#   1. identity_short        — Score(x, x) == 1.0
#   2. both_empty            — Score("", "") == 1.0 (raw 0)
#   3. one_empty_a           — Score("", non-empty) == 0.0
#   4. one_empty_b           — Score(non-empty, "") == 0.0
#   5. two_substring         — substring containment → clamp engages at 1.0
#   6. no_overlap            — no characters in common → 0.0 (local-zero-floor)
#   7. one_long_gap_canary   — Gotoh-erratum gate (PITFALLS.md §3 #2)
#   8. non_default_params    — custom Match/Mismatch/GapOpen/GapExtend
# Plus ~5-10 additional cases spanning unicode / single-char / all-mismatch /
# partial-middle-match / etc. to reach ~15-18 entries total.
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
    # Additional cases spanning the rest of the input space:
    ("single_char_match",    "a",              "a",                              None),
    ("single_char_diff",     "a",              "b",                              None),
    ("partial_middle_match", "xxabcyy",        "zzabczz",                        None),
    ("all_mismatch",         "abcd",           "wxyz",                           None),
    ("unicode_ascii_only",   "café",           "cafe",                           None),
    ("identity_long",        "the quick brown fox", "the quick brown fox",       None),
    ("near_identical",       "kitten",         "sitting",                        None),
    ("substring_at_end",     "ending",         "the long ending",                None),
]


def score_case(a, b, params):
    """Compute (raw, normalised) reference scores for a single case.

    Both-empty and one-empty cases are handled BEFORE invoking the aligner
    because ``PairwiseAligner.score("", "")`` behaviour is implementation-
    defined; we hard-wire the conventions used by our Go-side identity and
    one-empty short-circuits (swg.go).

    For non-empty cases, the aligner is configured in local mode with the
    given match/mismatch/gap_open/gap_extend scores; the raw alignment score
    is then normalised as ``clamp(raw / min(len(a), len(b)), 0, 1)``.
    """
    if a == "" and b == "":
        # Both-empty identity: matches our Go-side ``a == b`` short-circuit in
        # SmithWatermanGotohScore (returns 1.0; raw is 0 by definition — no
        # positions to score).
        return 0.0, 1.0
    if a == "" or b == "":
        # One-empty: matches our Go-side ``la == 0 || lb == 0`` short-circuit.
        return 0.0, 0.0

    aligner = PairwiseAligner()
    aligner.mode = "local"
    aligner.match_score = params["match"]
    aligner.mismatch_score = params["mismatch"]
    aligner.open_gap_score = params["gap_open"]
    aligner.extend_gap_score = params["gap_extend"]

    raw = float(aligner.score(a, b))
    min_len = min(len(a), len(b))
    norm = raw / min_len
    # Clamp to [0, 1] — matches our Go-side swgClampNormalise.
    norm = max(0.0, min(1.0, norm))
    return raw, norm


def _check_biopython_version() -> None:
    """Guard the script against silent corpus drift on outdated biopython.

    Parses ``Bio.__version__`` into a (major, minor) tuple and exits with a
    helpful message if it is older than _MIN_BIOPYTHON_VERSION. Without this
    guard, a developer running with biopython 1.79 (the version still pinned
    by some legacy distros) would regenerate a corpus that no longer matches
    the committed reference, and the Go-side TestSWG_CrossValidation would
    fail with a misleading delta rather than a version-mismatch error.
    """
    raw_parts = Bio.__version__.split(".")
    try:
        version = tuple(int(part) for part in raw_parts[:2])
    except ValueError:
        print(
            f"WARNING: could not parse biopython version "
            f"{Bio.__version__!r}; cannot enforce minimum {_MIN_BIOPYTHON_VERSION}.",
            file=sys.stderr,
        )
        return
    if version < _MIN_BIOPYTHON_VERSION:
        min_str = ".".join(str(x) for x in _MIN_BIOPYTHON_VERSION)
        sys.exit(
            f"ERROR: biopython {Bio.__version__} is older than the supported "
            f"minimum {min_str}. PairwiseAligner's affine-gap defaults shifted "
            f"between 1.79 and 1.85; regenerating with an older release would "
            f"silently produce a corpus that diverges from the committed "
            f"reference vectors. Upgrade with:\n"
            f"    python3 -m pip install --upgrade 'biopython>={min_str}'"
        )


def main():
    _check_biopython_version()
    entries = []
    for name, a, b, overrides in CASES:
        params = dict(DEFAULT_PARAMS)
        if overrides:
            params.update(overrides)
        raw, norm = score_case(a, b, params)
        entries.append({
            "name": name,
            "a": a,
            "b": b,
            "params": params,
            "biopython_score": raw,
            "biopython_normalised": norm,
        })

    out = {
        "version": 1,
        "biopython_version": Bio.__version__,
        "entries": entries,
    }
    path = "testdata/cross-validation/swg/vectors.json"
    os.makedirs(os.path.dirname(path), exist_ok=True)
    with open(path, "w") as f:
        json.dump(out, f, indent=2, sort_keys=False)
        f.write("\n")  # trailing LF matches Phase 2 golden-file canonical convention


if __name__ == "__main__":
    main()
