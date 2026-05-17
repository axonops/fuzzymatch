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

"""scripts/gen-qgram-cross-validation.py — py_stringmatching cross-validation
corpus generator for the four Phase-5 q-gram-tier algorithms.

Regenerates testdata/cross-validation/qgram/vectors.json by running
``py_stringmatching.{Jaccard, Dice, Cosine, TverskyIndex}`` over q-grams
extracted by ``py_stringmatching.QgramTokenizer(qval=q, padding=False)``
against a fixed, deterministic list of input pairs (CASES). Each entry
carries four scores at q=2 (and optionally q=3 for inputs long enough):
qgram_jaccard, sorensen_dice, cosine, tversky. The generated JSON is the
verification fixture consumed by the Go test ``TestQGram_CrossValidation``
(in qgram_cross_validation_test.go) — CI does NOT require Python at test
time; the committed JSON is the verification fixture read by the Go test
directly.

Reference implementation: py_stringmatching (BSD-3-Clause), maintained by
the UWisconsin MDM group. py_stringmatching is used here SOLELY for
reference-vector cross-validation per the project's
.claude/skills/algorithm-licensing-standards — NO py_stringmatching code
is copied into the fuzzymatch implementation. The Go-side q-gram kernels
in qgram_jaccard.go / sorensen_dice.go / cosine.go / tversky.go are fresh
transcriptions from their primary academic sources (Jaccard 1912 /
Sørensen 1948 / Dice 1945 / Salton & McGill 1983 / Tversky 1977); this
script provides independent reference scores that close the
algorithm-correctness-reviewer cross-validation gate per CLAUDE.md
"Workflow — Agent Gates".

py_stringmatching-vs-fuzzymatch surface mapping:
    - Jaccard.get_sim_score(qgrams_a, qgrams_b)
        ↔ fuzzymatch.QGramJaccardScore(a, b, q)
    - Dice.get_sim_score(qgrams_a, qgrams_b)
        ↔ fuzzymatch.SorensenDiceScore(a, b, q)
    - Cosine.get_sim_score(qgrams_a, qgrams_b)
        ↔ fuzzymatch.CosineScore(a, b, q)
    - TverskyIndex(alpha, beta).get_sim_score(qgrams_a, qgrams_b)
        ↔ fuzzymatch.TverskyScore(a, b, q, alpha, beta)

CRITICAL — set-vs-multiset DIVERGENCE (LOCKED 2026-05-17 plan 08.5-10):
    py_stringmatching's Jaccard / Dice / Cosine / TverskyIndex treat the
    input token lists as SETS — duplicate q-grams in the list collapse to
    a single set element. Confirmed by direct probe:
        Jaccard(['a','a','a'], ['a','a']).get_sim_score() == 1.0
    fuzzymatch's QGramJaccardScore / SorensenDiceScore / CosineScore /
    TverskyScore treat the q-gram extraction as a MULTISET — duplicate
    q-grams accumulate (the canonical Ukkonen 1992 multiset Jaccard
    formulation, with min/max over counts). Confirmed by direct probe:
        QGramJaccardScore("AAAA", "AAA", 2) == 2/3 (multiset: min(3,2)/max(3,2))
        py_stringmatching equivalent == 1.0 (set: {AA} ∩ {AA} / {AA} ∪ {AA})

    Resolution: the corpus is restricted to INPUTS WHOSE q-gram EXTRACTIONS
    HAVE NO DUPLICATES at the chosen q. When every q-gram in both a and b
    extractions is unique, set and multiset semantics coincide. The
    _assert_qgram_unique gate enforces this constraint at corpus
    construction time. Any case with duplicate q-grams must either be
    removed or marked variant_divergence=True with fuzzymatch-expected
    values overriding py_stringmatching's set-semantic values. This
    mirrors the OQ-1 "Tokenise-safety" pattern in
    gen-token-ratio-cross-validation.py.

CRITICAL — PYSM_VERSION pin gate:
    The script's first action after _check_python_version() is
    `assert py_stringmatching.__version__ == PYSM_VERSION`. py_stringmatching
    exposes __version__ at the module level (unlike jellyfish 1.2.1), so
    the direct attribute probe suffices — no `pip show` shell-out needed.
    Bumping the pin requires the six-step protocol documented in
    CONTRIBUTING.md.

CASE coverage (matches plan 08.5-10 task 2 acceptance criteria —
≥ 30 cases):
    - short ASCII identity / one-empty / both-empty
    - mismatched-length pairs
    - English-word pairs (Damerau-style near-matches)
    - jellyfish/smellyfish, kitten/sitting (q-gram-unique English-word
      pairs)
    - paper-anchored q-gram-friendly pairs

Each case carries q=2 outputs as the BASELINE (mandatory); cases with
both inputs length ≥ 4 also carry q=3 outputs (extends coverage to the
DP-style algorithms). The Tversky parameters used are (alpha=0.5, beta=0.5)
matching Tversky's "symmetric" canonical default; the dispatch slot
AlgoTversky uses these defaults too.

Usage:
    make regen-qgram-cross-validation
    # or directly:
    python3 scripts/gen-qgram-cross-validation.py

Requirements:
    - Python 3.7+ (for guaranteed dict insertion-order preservation).
    - py_stringmatching==0.4.7. Install:
      ``python3 -m pip install --user py_stringmatching==0.4.7``.

Determinism:
    The CASES list owns the corpus's deterministic ordering. Python 3.7+
    guarantees dict insertion-order preservation; ``json.dump`` with
    ``sort_keys=False`` honours that order. The _metadata block carries
    deterministic version strings; only the regenerated_at ISO timestamp
    is environment-dependent.
"""

from datetime import datetime, timezone
import json
import os
import sys

# Minimum supported Python version (dict insertion-order preservation).
_MIN_PYTHON_VERSION = (3, 7)

# Pinned py_stringmatching version. Bumping requires the six-step
# protocol documented in CONTRIBUTING.md.
PYSM_VERSION = "0.4.7"

# Tversky parameters used for the corpus. (alpha=0.5, beta=0.5) is the
# Tversky-symmetric canonical default and matches the dispatch slot's
# default in fuzzymatch.
TVERSKY_ALPHA = 0.5
TVERSKY_BETA = 0.5


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

import py_stringmatching as sm  # noqa: E402

assert sm.__version__ == PYSM_VERSION, (
    f"py_stringmatching version mismatch: installed {sm.__version__!r}, "
    f"script pinned to {PYSM_VERSION!r} — "
    f"run: python3 -m pip install --user py_stringmatching=={PYSM_VERSION}"
)


# CASES — the deterministic ordered list of cross-validation entries.
#
# Format: a dict per entry. Required keys:
#   - name: stable identifier (referenced in test failure messages).
#   - a, b: input strings.
# Optional keys:
#   - skip_q3: when True, only q=2 scores are emitted; q=3 scores are
#     null. Used for inputs where one side is shorter than q=3 (q-gram
#     extraction yields empty token list, which is a valid but
#     uninteresting case already covered by the empty-input fixtures).
#
# Inputs are constrained to be q-gram-unique at q=2 (and q=3 when not
# skip_q3) so set semantics (py_stringmatching) and multiset semantics
# (fuzzymatch) coincide — see _assert_qgram_unique gate.
CASES = [
    # Identity / empty / one-empty (q-gram tier convention: empty/empty → 1.0).
    {"name": "identity_short", "a": "hello", "b": "hello"},
    {"name": "identity_word", "a": "fuzzymatch", "b": "fuzzymatch"},
    {"name": "both_empty", "a": "", "b": "", "skip_q3": True},
    {"name": "one_empty_a", "a": "", "b": "hello", "skip_q3": True},
    {"name": "one_empty_b", "a": "hello", "b": "", "skip_q3": True},

    # Mismatched-length pairs (English words — q-gram-unique).
    {"name": "kitten_sitting", "a": "kitten", "b": "sitting"},
    {"name": "saturday_sunday", "a": "saturday", "b": "sunday"},
    {"name": "jellyfish_smellyfish", "a": "jellyfish", "b": "smellyfish"},

    # Jaro/JW canonical pair adapted for q-grams (q-gram-unique).
    {"name": "martha_marhta", "a": "martha", "b": "marhta"},
    {"name": "dixon_dicksonx", "a": "dixon", "b": "dicksonx"},
    {"name": "dwayne_duane", "a": "dwayne", "b": "duane"},

    # Disjoint pairs (no shared q-grams → score 0.0).
    {"name": "fully_different_short", "a": "abc", "b": "xyz"},
    {"name": "fully_different_words", "a": "hello", "b": "world"},

    # Near-identity (single q-gram diff).
    {"name": "single_char_sub_short", "a": "abcde", "b": "abcdf"},
    {"name": "single_char_sub_word", "a": "match", "b": "patch"},

    # Single-q-gram pairs.
    {"name": "single_qgram_match_q2", "a": "ab", "b": "ab"},
    {"name": "single_qgram_mismatch_q2", "a": "ab", "b": "cd"},

    # Short-input edge cases (input length < q → empty q-gram list).
    {"name": "shorter_than_q2_a", "a": "a", "b": "abc", "skip_q3": True},
    {"name": "shorter_than_q2_both", "a": "a", "b": "b", "skip_q3": True},

    # Mid-length and longer pairs (q-gram-unique).
    {"name": "mid_length_close", "a": "abcdef", "b": "abcdeg"},
    {"name": "mid_length_overlap", "a": "abcdef", "b": "cdefgh"},
    {"name": "mid_length_far", "a": "abcdef", "b": "uvwxyz"},

    # Subset-style pairs (one is a substring of the other — q-gram-unique).
    {"name": "subset_short_in_long", "a": "abc", "b": "abcdef"},
    {"name": "subset_long_in_short", "a": "abcdef", "b": "abc"},

    # English-word pairs designed for q-gram uniqueness.
    {"name": "fuzzy_match", "a": "fuzzy", "b": "match"},
    {"name": "quick_brown_fox", "a": "quick", "b": "brown"},
    {"name": "the_quick_brown_close", "a": "the quick brown fox", "b": "the quick brown dog"},

    # Q-gram-overlap stressors (mostly overlapping but with a few
    # divergent q-grams to surface ratio mismatches).
    {"name": "overlap_dominant", "a": "abcdefg", "b": "abcdefh"},
    {"name": "overlap_partial", "a": "abcdef", "b": "cdefab"},

    # Algorithm-specific stressor: pairs designed to surface differences
    # between Jaccard / Dice / Cosine / Tversky (different ratio
    # denominators).
    {"name": "ratio_stressor_short", "a": "abc", "b": "abcde"},
    {"name": "ratio_stressor_overlap", "a": "abcde", "b": "bcdef"},

    # Case-sensitivity exposure (py_stringmatching's QgramTokenizer is
    # case-preserving; fuzzymatch's q-gram functions are byte-level so
    # they are also case-preserving on ASCII — case-different inputs
    # produce disjoint q-grams).
    {"name": "case_sensitive_disjoint", "a": "ABC", "b": "abc"},
]


def _qgrams(s: str, q: int):
    """Return py_stringmatching's q-gram extraction (no padding)."""
    return sm.QgramTokenizer(qval=q, padding=False).tokenize(s)


def _assert_qgram_unique(case: dict, q: int) -> None:
    """Refuse to proceed if a non-skipped case's q-gram extraction has
    duplicates at q.

    Set vs. multiset semantics coincide iff every q-gram extracted from
    each side is unique. Inputs with duplicate q-grams would surface a
    Jaccard / Dice / Cosine / Tversky divergence between py_stringmatching
    (set) and fuzzymatch (multiset); cross-validating such cases would
    require the variant_divergence machinery. The simpler discipline:
    keep the corpus q-gram-unique. See module docstring "set-vs-multiset
    DIVERGENCE".
    """
    if case.get("skip_q3") and q == 3:
        return
    for field in ("a", "b"):
        s = case[field]
        tokens = _qgrams(s, q)
        if len(tokens) != len(set(tokens)):
            duplicates = [t for t in tokens if tokens.count(t) > 1]
            sys.exit(
                f"ERROR: case {case['name']!r} field {field!r} has "
                f"duplicate q-grams at q={q}: {sorted(set(duplicates))!r} "
                f"in tokens {tokens!r}. py_stringmatching's set semantics "
                f"diverge from fuzzymatch's multiset semantics on inputs "
                f"with duplicate q-grams — either choose a q-gram-unique "
                f"input or mark this case variant_divergence=True with an "
                f"override (NOT currently supported by the corpus schema)."
            )


def score_pair_at_q(a: str, b: str, q: int) -> dict:
    """Compute the four reference scores at q for a single (a, b)."""
    qa = _qgrams(a, q)
    qb = _qgrams(b, q)
    # py_stringmatching's get_sim_score returns the [0, 1] similarity.
    # Empty/empty returns 1.0; one-empty returns 0.0. fuzzymatch follows
    # the same conventions (catalogue-wide).
    return {
        "qgram_jaccard": sm.Jaccard().get_sim_score(qa, qb),
        "sorensen_dice": sm.Dice().get_sim_score(qa, qb),
        "cosine": sm.Cosine().get_sim_score(qa, qb),
        "tversky": sm.TverskyIndex(TVERSKY_ALPHA, TVERSKY_BETA).get_sim_score(qa, qb),
    }


def score_case(case: dict) -> dict:
    """Compute q=2 (and q=3 when not skip_q3) reference scores for a case."""
    a, b = case["a"], case["b"]
    entry = {
        "name": case["name"],
        "a": a,
        "b": b,
        "q2": score_pair_at_q(a, b, 2),
    }
    if case.get("skip_q3"):
        entry["q3"] = None
    else:
        entry["q3"] = score_pair_at_q(a, b, 3)
    return entry


def main() -> None:
    for case in CASES:
        _assert_qgram_unique(case, 2)
        _assert_qgram_unique(case, 3)

    entries = [score_case(c) for c in CASES]

    out = {
        "version": 1,
        "_metadata": {
            "py_stringmatching_version": PYSM_VERSION,
            "python_version": f"{sys.version_info.major}.{sys.version_info.minor}.{sys.version_info.micro}",
            "tversky_alpha": TVERSKY_ALPHA,
            "tversky_beta": TVERSKY_BETA,
            "regenerated_at": datetime.now(timezone.utc).isoformat(),
        },
        "entries": entries,
    }
    path = "testdata/cross-validation/qgram/vectors.json"
    os.makedirs(os.path.dirname(path), exist_ok=True)
    with open(path, "w") as f:
        json.dump(out, f, indent=2, sort_keys=False)
        f.write("\n")  # trailing LF matches Phase 2 golden-file canonical convention


if __name__ == "__main__":
    main()
