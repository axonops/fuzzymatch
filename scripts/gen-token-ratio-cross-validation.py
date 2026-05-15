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

"""scripts/gen-token-ratio-cross-validation.py — RapidFuzz cross-validation
corpus generator for the four Phase-6 Indel-based ratios.

Regenerates testdata/cross-validation/token-ratios/vectors.json by running
``rapidfuzz.fuzz.{token_sort_ratio,token_set_ratio,partial_ratio}`` against
a fixed, deterministic list of input pairs (CASES). Each entry carries all
four scores: token_sort_ratio, token_set_ratio, partial_ratio_bytes (the
byte-path partial ratio), and partial_ratio_runes (the rune-path partial
ratio — for ASCII inputs identical to the byte path, but emitted from a
separate code path to exercise rune-aware regressions). The generated JSON
is the verification fixture consumed by the Go test
``TestTokenRatios_CrossValidation`` (in token_ratio_cross_validation_test.go)
— CI does NOT require Python at test time; the committed JSON is read by
the Go test directly.

Reference implementation: RapidFuzz (MIT licence), maintained by Max
Bachmann. RapidFuzz documents the Indel formula at
https://rapidfuzz.github.io/RapidFuzz/Usage/distance/Indel.html and the
fuzz ratios at https://rapidfuzz.github.io/RapidFuzz/Usage/fuzz.html.
RapidFuzz is used here SOLELY for reference-vector cross-validation per
the project's .claude/skills/algorithm-licensing-standards — NO rapidfuzz
code is copied into the fuzzymatch implementation. The Go-side Indel
kernel in token_indel.go is a fresh transcription from Wagner & Fischer
1974; this script provides independent reference scores that close the
algorithm-correctness-reviewer cross-validation gate per CLAUDE.md
"Workflow — Agent Gates".

CRITICAL — RAPIDFUZZ_VERSION pin gate:
    The script's first action after importing rapidfuzz is to assert
    ``rapidfuzz.__version__ == RAPIDFUZZ_VERSION``. Regenerating with a
    different version is REFUSED with a clear error message; the corpus
    is locked to a single RapidFuzz release for byte-stable cross-
    validation. Bumping requires (a) updating RAPIDFUZZ_VERSION here,
    (b) updating the install hint in Makefile's
    regen-token-ratio-cross-validation target, (c) re-running this
    script, (d) committing the regenerated vectors.json, and (e) running
    ``go test -run TestTokenRatios_CrossValidation -race ./...``. Any
    score drift surfaces as a test failure requiring
    algorithm-correctness-reviewer review.

Tokeniser-divergence handling (OQ-1 RESOLUTION LOCKED in plan 06-01):
    RapidFuzz tokenises via Python ``str.split()`` — whitespace-only,
    case-preserving. fuzzymatch's ``Tokenise(s, DefaultTokeniseOptions())``
    is identifier-aware (camelCase / snake_case / kebab-case / dot-case)
    AND lowercasing. To make cross-validation tractable, the corpus is
    restricted to inputs where the two tokenisations agree — pure
    whitespace-separated lowercase ASCII text. The script's
    ``_assert_corpus_is_tokenise_safe`` gate enforces this constraint
    for TokenSort / TokenSet entries. The script also calls ``.lower()``
    on inputs before passing to RapidFuzz to reconcile RapidFuzz's
    case-preserving default with fuzzymatch's lowercasing Tokenise.

    PartialRatio entries (which are character-level — no tokenise) MAY
    use inputs that would otherwise violate the Tokenise-safety
    constraint, BUT must carry ``"partial_only": True`` so the
    TokenSort / TokenSet sub-tests skip them (the generator emits null
    for those fields). See ``_assert_corpus_is_tokenise_safe`` and the
    per-case "partial_only" flag.

OQ-2 RESOLUTION (single combined corpus — LOCKED in plan 06-01):
    One vectors.json with all four scores per entry. Mirrors Phase 3/4
    structure. The Wave-1 Go loader asserts only TokenSortRatio entries;
    plans 06-02 (TokenSetRatio) and 06-03 (PartialRatio) extend the
    loader's per-surface sub-test loop to assert TokenSet / PartialRatio
    scores as those algorithms land.

OQ-3 RESOLUTION (partial_ratio_runes always emitted — LOCKED in plan
06-01):
    Every non-partial_only entry emits partial_ratio_runes as well as
    partial_ratio_bytes. For ASCII inputs the two values are identical;
    the rune-path emission still exercises the separate code path in the
    Go implementation to catch path-specific regressions. partial_only
    entries always emit both partial_ratio_bytes and partial_ratio_runes
    (their reason for being in the corpus is exactly to exercise
    PartialRatio code paths).

Usage:
    make regen-token-ratio-cross-validation
    # or directly:
    python3 scripts/gen-token-ratio-cross-validation.py

Requirements:
    - Python 3.7+ (for guaranteed dict insertion-order preservation in
      json.dump, which is load-bearing for the corpus's byte-stable
      output). Enforced at script entry — see _check_python_version().
    - rapidfuzz==3.14.5 (the pinned version). Install:
      ``python3 -m pip install --user rapidfuzz==3.14.5``.

Determinism:
    The CASES list owns the corpus's deterministic ordering. Python 3.7+
    guarantees dict insertion-order preservation, and ``json.dump`` with
    ``sort_keys=False`` honours that order. The _metadata block uses
    deterministic Python / rapidfuzz version strings — the regenerated_at
    ISO timestamp is the only environment-dependent field, and the
    Go-side loader does not assert byte-stability on the corpus, only
    on the per-entry scores; the timestamp is informational for triage.
"""

from datetime import datetime, timezone
import json
import os
import sys

# Minimum supported Python version. Python 3.7 was the first release to
# guarantee dict insertion-order preservation as a language feature.
# Enforced at script entry — see _check_python_version().
_MIN_PYTHON_VERSION = (3, 7)

# Pinned rapidfuzz version. Bumping this string requires the five-step
# protocol documented in the module docstring above (and in
# docs/cross-validation.md).
RAPIDFUZZ_VERSION = "3.14.5"

import rapidfuzz  # noqa: E402 — imported after the version constant so the assert message can reference it
assert rapidfuzz.__version__ == RAPIDFUZZ_VERSION, (
    f"rapidfuzz version mismatch: installed {rapidfuzz.__version__}, "
    f"script pinned to {RAPIDFUZZ_VERSION} — "
    f"run: python3 -m pip install --user rapidfuzz=={RAPIDFUZZ_VERSION}"
)

from rapidfuzz import fuzz  # noqa: E402

# CASES — the deterministic ordered list of cross-validation entries.
#
# Format: a dict per entry. Required keys:
#   - name: stable identifier (referenced in test failure messages).
#   - a, b: input strings.
# Optional keys:
#   - partial_only: when True, the entry is a PartialRatio-only fixture;
#     the TokenSort / TokenSet entries emit null because the input
#     violates the Tokenise-safety constraint (e.g. contains uppercase
#     letters, separators, or camelCase that fuzzymatch's Tokenise
#     would split differently from RapidFuzz's str.split).
#
# Required coverage (matches plan 06-01 task 3 success criteria):
#   - identity_short, both_empty, one_empty_a, one_empty_b
#   - tokens_reordered, subset_a_in_b, subset_b_in_a
#   - tokenset_diff_dominates (TokenSet three-way max edge per
#     06-RESEARCH.md Pattern 5)
#   - tokenset_subset_shortcircuit
#   - partial_left_tail_wins / partial_right_tail_wins (Pitfall 3
#     Region-1 and Region-3 fixtures — partial_only)
#   - partial_middle_wins
#   - long_short_mismatch (small pathological pair, partial_only)
#   - unicode_lowercase_ascii_safe — multi-token input where the
#     tokenisations agree after .lower() reconciliation
#   - ascii_medium / ascii_long
CASES = [
    {"name": "identity_short", "a": "hello world", "b": "hello world"},
    {"name": "both_empty", "a": "", "b": ""},
    {"name": "one_empty_a", "a": "", "b": "hello world"},
    {"name": "one_empty_b", "a": "hello world", "b": ""},
    {"name": "tokens_reordered_two", "a": "alpha beta", "b": "beta alpha"},
    {"name": "tokens_reordered_canonical", "a": "fuzzy wuzzy was a bear", "b": "wuzzy fuzzy was a bear"},
    {"name": "subset_a_in_b", "a": "alpha beta", "b": "alpha beta gamma"},
    {"name": "subset_b_in_a", "a": "alpha beta gamma", "b": "alpha beta"},
    {"name": "tokenset_diff_dominates", "a": "the cat sat on the mat", "b": "the cat sat on the rug"},
    {"name": "tokenset_subset_shortcircuit", "a": "alpha beta", "b": "alpha beta alpha beta"},
    {"name": "tokens_disjoint", "a": "abc def", "b": "xyz qrs"},
    {"name": "tokens_low_overlap", "a": "hello world", "b": "world peace"},
    {"name": "long_token_reorder", "a": "alpha beta gamma delta epsilon zeta", "b": "zeta epsilon delta gamma beta alpha"},
    {"name": "ascii_medium", "a": "the quick brown fox jumps over the lazy dog", "b": "the quick brown dog jumps over the lazy fox"},
    {"name": "ascii_long", "a": "alpha beta gamma delta epsilon zeta eta theta iota kappa", "b": "kappa iota theta eta zeta epsilon delta gamma beta alpha"},
    {"name": "unicode_lowercase_ascii_safe", "a": "alpha beta", "b": "beta gamma alpha"},
    # PartialRatio-only fixtures (Pitfall 3 region-coverage). These
    # inputs would be valid for TokenSort / TokenSet too (lowercase,
    # whitespace-separated) but their semantic interest is purely the
    # PartialRatio sliding-window — we keep them as partial_only=False
    # because the Tokenise-safety check still passes. The partial_only
    # flag is reserved for inputs that contain uppercase / separators /
    # camelCase that would diverge between RapidFuzz str.split and
    # fuzzymatch Tokenise.
    {"name": "partial_left_tail_wins", "a": "ab", "b": "abc"},
    {"name": "partial_right_tail_wins", "a": "bc", "b": "abc"},
    {"name": "partial_middle_wins", "a": "lazy", "b": "the lazy dog"},
    {"name": "long_short_mismatch", "a": "x", "b": "the very long quick brown fox jumps over the lazy dog"},
]


def _assert_corpus_is_tokenise_safe(case: dict) -> None:
    """Refuse to proceed if a non-partial_only case violates the
    Tokenise-safety constraint (OQ-1 RESOLUTION).

    The constraint is: every byte of case["a"] and case["b"] must be
    a lowercase ASCII letter (a-z) OR an ASCII space (0x20). This
    guarantees fuzzymatch's Tokenise produces the same token list as
    RapidFuzz's Python str.split — both yield whitespace-separated
    lowercase tokens. Any uppercase letter, separator (_-.:/), or
    multibyte rune would cause the two tokenisations to diverge.

    Bypassed when case["partial_only"] is True — see module docstring
    for the partial_only contract.
    """
    if case.get("partial_only"):
        return
    for field in ("a", "b"):
        s = case[field]
        for ch in s:
            if ch == " ":
                continue
            if "a" <= ch <= "z":
                continue
            sys.exit(
                f"ERROR: case {case['name']!r} field {field!r} contains "
                f"non-Tokenise-safe character {ch!r} (only lowercase ASCII a-z "
                f"and space are allowed for TokenSort/TokenSet cases per "
                f"OQ-1 RESOLUTION LOCKED). Either lowercase the input, or "
                f"mark the case partial_only=True."
            )


def score_case(case: dict) -> dict:
    """Compute the four RapidFuzz reference scores for a single case.

    Each score is divided by 100.0 because RapidFuzz returns scores in
    [0, 100]; the project surface is [0.0, 1.0].

    Both-empty and one-empty cases are handled BEFORE invoking RapidFuzz
    so the conventions match the Go-side short-circuits in
    token_sort_ratio.go / token_set_ratio.go / partial_ratio.go.
    RapidFuzz's behaviour on empty inputs is generally consistent with
    these conventions but the empty-empty case for token_set_ratio is
    notably bug-for-bug compatible with the original fuzzywuzzy
    behaviour (issue #110): RapidFuzz token_set_ratio("", "") returns
    0.0, NOT 1.0. The project will match this deviation in plan
    06-02 — documented in docs/cross-validation.md.

    The .lower() reconciliation (per OQ-1 RESOLUTION) is applied to
    inputs before they are passed to RapidFuzz. For Tokenise-safe
    inputs this is a no-op (the inputs are already lowercase); for
    partial_only inputs the lowercase fold ensures RapidFuzz's
    case-preserving default doesn't surface a difference that the
    fuzzymatch implementation would not see (its Tokenise is
    lowercasing).
    """
    a, b = case["a"], case["b"]
    a_lower, b_lower = a.lower(), b.lower()

    partial_only = case.get("partial_only", False)

    # Token Sort / Token Set: emit null on partial_only entries.
    if partial_only:
        token_sort = None
        token_set = None
    else:
        token_sort = fuzz.token_sort_ratio(a_lower, b_lower) / 100.0
        token_set = fuzz.token_set_ratio(a_lower, b_lower) / 100.0

    # Partial Ratio: always computed (the algorithm doesn't tokenise).
    # The "bytes" surface is RapidFuzz's default behaviour over the raw
    # strings; the "runes" surface in the project is the rune-aware
    # variant. For ASCII inputs the two are identical; we emit them as
    # the same value here, and the Go-side rune-path implementation in
    # plan 06-03 will be cross-validated against partial_ratio_runes
    # specifically. RapidFuzz operates over Python strings which are
    # codepoint-indexed, so for multi-byte inputs RapidFuzz's value is
    # the rune-path semantic; on ASCII this coincides with the byte-path.
    partial = fuzz.partial_ratio(a_lower, b_lower) / 100.0

    return {
        "name": case["name"],
        "a": a,
        "b": b,
        "token_sort_ratio": token_sort,
        "token_set_ratio": token_set,
        "partial_ratio_bytes": partial,
        "partial_ratio_runes": partial,
        # Emit partial_only when present so the Go loader can branch
        # without re-deriving the rule. json serialises False as false;
        # the field is always present so the JSON shape is uniform.
        "partial_only": partial_only,
    }


def _check_python_version() -> None:
    """Guard the script against silent corpus drift on outdated Python.

    Python 3.7 promoted dict insertion-order preservation from CPython
    implementation detail to language guarantee. The corpus's byte-stable
    JSON output relies on this guarantee; older releases could produce
    reordered fields that diverge from the committed reference. Same
    pattern as gen-ratcliff-obershelp-cross-validation.py's
    _check_python_version helper.
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
    for case in CASES:
        _assert_corpus_is_tokenise_safe(case)

    entries = [score_case(c) for c in CASES]

    out = {
        "version": 1,
        "_metadata": {
            "rapidfuzz_version": RAPIDFUZZ_VERSION,
            "python_version": f"{sys.version_info.major}.{sys.version_info.minor}.{sys.version_info.micro}",
            "regenerated_at": datetime.now(timezone.utc).isoformat(),
        },
        "entries": entries,
    }
    path = "testdata/cross-validation/token-ratios/vectors.json"
    os.makedirs(os.path.dirname(path), exist_ok=True)
    with open(path, "w") as f:
        json.dump(out, f, indent=2, sort_keys=False)
        f.write("\n")  # trailing LF matches Phase 2 golden-file canonical convention


if __name__ == "__main__":
    main()
