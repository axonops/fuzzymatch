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

"""scripts/gen-monge-elkan-cross-validation.py — py_stringmatching
cross-validation corpus generator for the Monge-Elkan similarity with
JaroWinkler-inner and Levenshtein-inner.

Regenerates testdata/cross-validation/monge-elkan/vectors.json by running
``py_stringmatching.MongeElkan`` with two inner-similarity choices:
  - py_stringmatching.JaroWinkler().get_sim_score
  - py_stringmatching.Levenshtein().get_sim_score
against a fixed, deterministic list of input pairs (CASES). Each entry
carries FOUR scores:
  - monge_elkan_asymmetric_jw  — ME(A → B) with JaroWinkler inner
  - monge_elkan_asymmetric_lev — ME(A → B) with Levenshtein inner
  - monge_elkan_symmetric_jw   — (ME(A→B,JW) + ME(B→A,JW)) / 2
  - monge_elkan_symmetric_lev  — (ME(A→B,Lev) + ME(B→A,Lev)) / 2

The generated JSON is the verification fixture consumed by the Go test
``TestMongeElkan_CrossValidation`` (in monge_elkan_cross_validation_test.go).
CI does NOT require Python at test time; the committed JSON is the
verification fixture read by the Go test directly.

Reference implementation: py_stringmatching (BSD-3-Clause). The
MongeElkan class is the modern Python canonical reference per
Cohen-Ravikumar-Fienberg 2003 (SecondString). py_stringmatching is used
here SOLELY for reference-vector cross-validation per the project's
.claude/skills/algorithm-licensing-standards — NO py_stringmatching code
is copied into the fuzzymatch implementation. The Go-side Monge-Elkan
kernel in monge_elkan.go is a fresh transcription from Monge & Elkan
1996 §3 (KDD'96 proceedings); this script provides independent reference
scores that close the algorithm-correctness-reviewer cross-validation
gate per CLAUDE.md "Workflow — Agent Gates".

py_stringmatching-vs-fuzzymatch surface mapping:
    - py_stringmatching.MongeElkan(JaroWinkler().get_sim_score)
        .get_raw_score(tokens(a), tokens(b))
        ↔ fuzzymatch.MongeElkanScore(a, b, AlgoJaroWinkler, opts)
    - py_stringmatching.MongeElkan(Levenshtein().get_sim_score)
        .get_raw_score(tokens(a), tokens(b))
        ↔ fuzzymatch.MongeElkanScore(a, b, AlgoLevenshtein, opts)
    - The SYMMETRIC variants (mean of forward + reverse) are NOT a
      py_stringmatching primitive — the corpus computes them as
      (ME(A→B) + ME(B→A)) / 2 in Python, matching the Q3-LOCKED semantics
      in 06-CONTEXT.md §4 that fuzzymatch's MongeElkanScoreSymmetric
      implements.

Tokeniser-divergence handling (mirrors OQ-1 in gen-token-ratio-cross-
validation.py):
    fuzzymatch's Tokenise(s, DefaultTokeniseOptions()) is identifier-aware
    (camelCase / snake_case / kebab-case / dot-case) AND lowercasing. The
    Python script's str.split() is whitespace-only and case-preserving.
    To make the tokenisations agree, the corpus is restricted to inputs
    that are pure WHITESPACE-SEPARATED LOWERCASE ASCII LETTERS — the
    _assert_corpus_is_tokenise_safe gate refuses any other input.

    JaroWinkler inner-precision note:
        py_stringmatching's JaroWinkler internally uses float32 for some
        intermediate computations (confirmed by direct probe:
        JaroWinkler().get_sim_score("john", "johnathan") ==
        0.8888888955116272 vs fuzzymatch's float64 0.8888888888888889 —
        an fp32-vs-fp64 difference of ~6.6e-9 per token comparison).
        Monge-Elkan with JaroWinkler-inner can accumulate these
        differences across the per-token max-of-sim reduction. The
        Go-side test therefore uses a relaxed tolerance of 1e-6 for the
        JW-inner surfaces and 1e-9 for the Levenshtein-inner surfaces
        (Levenshtein is integer + single division, fp64-clean across both
        libraries). The relaxed JW-inner tolerance is documented in
        monge_elkan_cross_validation_test.go's file header.

CRITICAL — PYSM_VERSION pin gate:
    Same pattern as gen-qgram-cross-validation.py: assert
    py_stringmatching.__version__ == PYSM_VERSION at the top of the
    script. Bumping requires the six-step protocol documented in
    CONTRIBUTING.md.

CASE coverage (matches plan 08.5-10 task 3 acceptance criteria —
≥ 30 cases):
    - Identity / empty / one-empty
    - Single-token pairs (asymmetric reduces to symmetric)
    - Equal-token-count pairs (still symmetric — fuzzymatch property:
      MongeElkanScore(a, b) == MongeElkanScoreSymmetric(a, b) when
      |tokens(a)| == |tokens(b)| AND the inner-metric matrix is
      symmetric; verified for our 14 permitted inners)
    - Unequal-token-count pairs (asymmetric ≠ symmetric — the main
      cross-validation interest)
    - Multi-word identifier-style pairs (kept Tokenise-safe by
      lowercasing and whitespace-only separation)

Usage:
    make regen-monge-elkan-cross-validation
    # or directly:
    python3 scripts/gen-monge-elkan-cross-validation.py

Requirements:
    - Python 3.7+ (for guaranteed dict insertion-order preservation).
    - py_stringmatching==0.4.7. Install:
      ``python3 -m pip install --user py_stringmatching==0.4.7``.

Determinism:
    The CASES list owns the corpus's deterministic ordering. Python 3.7+
    guarantees dict insertion-order preservation; json.dump with
    sort_keys=False honours that order.
"""

from datetime import datetime, timezone
import json
import os
import sys

# Minimum supported Python version (dict insertion-order preservation).
_MIN_PYTHON_VERSION = (3, 7)

# Pinned py_stringmatching version — must match gen-qgram-cross-validation.py.
PYSM_VERSION = "0.4.7"


def _check_python_version() -> None:
    """Guard against silent corpus drift on outdated Python (< 3.7)."""
    if sys.version_info < _MIN_PYTHON_VERSION:
        min_str = ".".join(str(x) for x in _MIN_PYTHON_VERSION)
        cur_str = ".".join(str(x) for x in sys.version_info[:3])
        sys.exit(
            f"ERROR: Python {cur_str} is older than the supported minimum "
            f"{min_str}. Upgrade to Python >= {min_str}."
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
#   - a, b: input strings (must be Tokenise-safe — pure whitespace-
#     separated lowercase ASCII letters; the _assert_corpus_is_tokenise_safe
#     gate enforces this).
CASES = [
    # Identity / empty / one-empty.
    {"name": "identity_short", "a": "hello", "b": "hello"},
    {"name": "identity_two_token", "a": "hello world", "b": "hello world"},
    {"name": "both_empty", "a": "", "b": ""},
    {"name": "one_empty_a", "a": "", "b": "hello"},
    {"name": "one_empty_b", "a": "hello", "b": ""},

    # Single-token pairs (asymmetric reduces to symmetric since |tA| == |tB| == 1).
    {"name": "single_token_match", "a": "martha", "b": "marhta"},
    {"name": "single_token_match_jw", "a": "dwayne", "b": "duane"},
    {"name": "single_token_match_lev", "a": "kitten", "b": "sitting"},
    {"name": "single_token_mismatch", "a": "alpha", "b": "zulu"},

    # Equal-token-count pairs (asymmetric == symmetric when the
    # inner-metric matrix is symmetric).
    {"name": "two_token_equal_count_match", "a": "hello world", "b": "hello world"},
    {"name": "two_token_equal_count_near", "a": "hello world", "b": "hallo world"},
    {"name": "two_token_equal_count_swap", "a": "alpha beta", "b": "beta alpha"},
    {"name": "three_token_equal_count_near", "a": "the quick fox", "b": "the quik fox"},

    # Unequal-token-count pairs (asymmetric ≠ symmetric — the main
    # cross-validation interest). The Q3 LOCKED semantics of
    # MongeElkanScoreSymmetric ensure the symmetric variant is
    # order-invariant.
    {"name": "asym_one_extra_token_b", "a": "alpha beta", "b": "alpha beta gamma"},
    {"name": "asym_one_extra_token_a", "a": "alpha beta gamma", "b": "alpha beta"},
    {"name": "asym_two_extra_token_b", "a": "alpha", "b": "alpha beta gamma"},
    {"name": "asym_two_extra_token_a", "a": "alpha beta gamma", "b": "alpha"},

    # Multi-word identifier-style pairs (Tokenise-safe — lowercase ASCII).
    {"name": "name_match_extra_middle", "a": "john smith", "b": "john paul smith"},
    {"name": "name_match_swap_order", "a": "smith john", "b": "john smith"},
    {"name": "name_match_typo", "a": "paul johnson", "b": "paule johnson"},
    {"name": "name_match_typo_two", "a": "paul johnson", "b": "paul jonhson"},

    # English-sentence pairs (longer, multi-token).
    {"name": "sentence_close_match", "a": "the quick brown fox", "b": "the quick brown dog"},
    {"name": "sentence_typo", "a": "the quick brown fox", "b": "the quik brwon fox"},
    {"name": "sentence_reorder", "a": "the quick brown fox", "b": "fox brown quick the"},

    # Subset / superset (asymmetric ≠ symmetric).
    {"name": "subset_short_in_long", "a": "alpha beta", "b": "alpha beta gamma delta"},
    {"name": "subset_long_in_short", "a": "alpha beta gamma delta", "b": "alpha beta"},

    # Disjoint pairs.
    {"name": "disjoint_short", "a": "alpha beta", "b": "uvw xyz"},
    {"name": "disjoint_long", "a": "the quick brown fox", "b": "all work and play"},

    # Mixed-near-and-far token pairs (one near match, one far match).
    {"name": "mixed_near_far_a", "a": "alpha beta", "b": "alpha xyz"},
    {"name": "mixed_near_far_b", "a": "alpha beta gamma", "b": "alpha xyz qrs"},

    # Long multi-token pairs.
    {"name": "long_six_token_close", "a": "the quick brown fox jumps over", "b": "the quik brown foxx jumps over"},
    {"name": "long_six_token_diff", "a": "alpha beta gamma delta epsilon zeta", "b": "alpha beta gamma delta epsilon eta"},
]


def _assert_corpus_is_tokenise_safe(case: dict) -> None:
    """Refuse to proceed if a case violates the Tokenise-safety constraint.

    Constraint: every byte of case["a"] and case["b"] must be a lowercase
    ASCII letter (a-z) OR an ASCII space (0x20). Mirrors the OQ-1
    resolution in gen-token-ratio-cross-validation.py.
    """
    for field in ("a", "b"):
        s = case[field]
        for ch in s:
            if ch == " ":
                continue
            if "a" <= ch <= "z":
                continue
            sys.exit(
                f"ERROR: case {case['name']!r} field {field!r} contains "
                f"non-Tokenise-safe character {ch!r} (only lowercase ASCII "
                f"a-z and space are allowed). Fix the input."
            )


def _tokenise(s: str) -> list:
    """Mirror fuzzymatch's Tokenise(s, DefaultTokeniseOptions()) on
    Tokenise-safe inputs.

    Because the corpus is restricted to whitespace-separated lowercase
    ASCII letters, str.split() suffices: separator chars are spaces,
    lowercase-fold is a no-op, and there is no camelCase / snake_case /
    kebab-case ambiguity to resolve. Empty / whitespace-only inputs
    return an empty list.
    """
    return s.split()


def _monge_elkan_raw(bag_a: list, bag_b: list, sim_func) -> float:
    """Compute py_stringmatching's MongeElkan.get_raw_score with explicit
    handling for the empty-input cases that fuzzymatch's surface returns.

    py_stringmatching's MongeElkan handles empty-bag and identity-bag
    edge cases internally (returns 1.0 for exact match, 0 for one-empty).
    Both library and fuzzymatch agree on these conventions.
    """
    me = sm.MongeElkan(sim_func=sim_func)
    return me.get_raw_score(bag_a, bag_b)


def score_case(case: dict) -> dict:
    """Compute the four Monge-Elkan reference scores for a single case.

    Identity / empty short-circuits are applied BEFORE invoking
    py_stringmatching so the conventions match fuzzymatch's surface (a == b
    → 1.0; one-empty → 0.0; both-empty → 1.0). py_stringmatching's
    MongeElkan applies the same conventions but going through the explicit
    short-circuit guarantees byte-identical output regardless of
    py_stringmatching's internal implementation choices.
    """
    a, b = case["a"], case["b"]
    bag_a = _tokenise(a)
    bag_b = _tokenise(b)

    jw_sim = sm.JaroWinkler().get_sim_score
    lev_sim = sm.Levenshtein().get_sim_score

    # Asymmetric and symmetric for both inners.
    if a == b:
        # Identity short-circuit — both fuzzymatch and py_stringmatching
        # agree (1.0). Using the explicit value avoids a divide-by-zero
        # surface for empty-empty (both bags []).
        me_asym_jw = 1.0
        me_asym_lev = 1.0
        me_sym_jw = 1.0
        me_sym_lev = 1.0
    elif len(bag_a) == 0 or len(bag_b) == 0:
        # One-empty (or one all-whitespace) — both libraries return 0.0
        # (or 1.0 if both empty; covered by the a == b branch above).
        me_asym_jw = 0.0
        me_asym_lev = 0.0
        me_sym_jw = 0.0
        me_sym_lev = 0.0
    else:
        # Standard case: compute forward and reverse, then derive the
        # symmetric mean per Q3-LOCKED semantics.
        fwd_jw = _monge_elkan_raw(bag_a, bag_b, jw_sim)
        rev_jw = _monge_elkan_raw(bag_b, bag_a, jw_sim)
        fwd_lev = _monge_elkan_raw(bag_a, bag_b, lev_sim)
        rev_lev = _monge_elkan_raw(bag_b, bag_a, lev_sim)
        me_asym_jw = fwd_jw
        me_asym_lev = fwd_lev
        me_sym_jw = (fwd_jw + rev_jw) / 2.0
        me_sym_lev = (fwd_lev + rev_lev) / 2.0

    return {
        "name": case["name"],
        "a": a,
        "b": b,
        "monge_elkan_asymmetric_jw": me_asym_jw,
        "monge_elkan_asymmetric_lev": me_asym_lev,
        "monge_elkan_symmetric_jw": me_sym_jw,
        "monge_elkan_symmetric_lev": me_sym_lev,
    }


def main() -> None:
    for case in CASES:
        _assert_corpus_is_tokenise_safe(case)

    entries = [score_case(c) for c in CASES]

    out = {
        "version": 1,
        "_metadata": {
            "py_stringmatching_version": PYSM_VERSION,
            "python_version": f"{sys.version_info.major}.{sys.version_info.minor}.{sys.version_info.micro}",
            "regenerated_at": datetime.now(timezone.utc).isoformat(),
        },
        "entries": entries,
    }
    path = "testdata/cross-validation/monge-elkan/vectors.json"
    os.makedirs(os.path.dirname(path), exist_ok=True)
    with open(path, "w") as f:
        json.dump(out, f, indent=2, sort_keys=False)
        f.write("\n")  # trailing LF matches Phase 2 golden-file canonical convention


if __name__ == "__main__":
    main()
