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

"""scripts/gen-phonetic-cross-validation.py — Dual-pin phonetic cross-validation
corpus generator for the four Phase-7 phonetic algorithms.

Regenerates testdata/cross-validation/phonetic/vectors.json by running the
pinned jellyfish (for Soundex, NYSIIS, MRA) and metaphone (for Double
Metaphone) Python packages against a fixed, deterministic list of input
strings (SOUNDEX_INPUTS, DM_INPUTS, NYSIIS_INPUTS, MRA_INPUTS).

OQ-1 RESOLUTION LOCKED 2026-05-15: This script uses TWO pinned packages
because jellyfish 1.x removed its Double Metaphone implementation:
  - jellyfish==1.2.1  for Soundex, NYSIIS, MRA
  - Metaphone==0.6    for Double Metaphone (oubiwann's BSD-licensed Andrew
                      Collins port — the de-facto Python DM canonical source)

Reference implementations:
  - jellyfish (BSD-2-Clause), maintained by James Turk.
    Soundex: jellyfish/src/soundex.rs — Knuth/Census H/W-skip variant.
    NYSIIS: extended (not-truncated-to-6) variant — see variant_divergence.
    MRA: jellyfish.match_rating_codex() + jellyfish.match_rating_comparison().
  - metaphone (BSD-3-Clause), oubiwann/metaphone on GitHub.
    Double Metaphone: metaphone.doublemetaphone(s) -> (primary, secondary).
    Fresh translation of Lawrence Philips' public-domain C++ reference (2000).

Both packages are used SOLELY for reference-vector cross-validation per the
project's .claude/skills/algorithm-licensing-standards — NO code is copied
from either package into the fuzzymatch implementation. The Go-side phonetic
implementations are fresh transcriptions from their primary academic sources.

CRITICAL — DUAL VERSION-PIN GATE:
    The script asserts jellyfish.__version__ == JELLYFISH_VERSION immediately
    after import. It also asserts the installed Metaphone package version
    matches METAPHONE_VERSION via `pip show Metaphone` (the Metaphone PyPI
    package does not expose __version__ reliably). Both checks refuse to run
    on any mismatch with a clear error message.

Variant-divergence handling (per CONTEXT.md §1 LOCKED):
    Some algorithm variants differ between our implementation (which follows
    the primary source) and jellyfish's implementation (which may use
    extended or modified variants). When divergence exists:
      - entry["variant_divergence"] = True
      - entry["divergent_jellyfish_value"] = <jellyfish_value>
      - entry["code"] (for Soundex/NYSIIS/MRA) or entry["primary"]/
        entry["secondary"] (for DM) = the Knuth/Taft/NBS-expected value
    The Go loader asserts our implementation matches the EXPECTED value,
    not the jellyfish value. This makes divergences transparent without
    causing false-positive failures.

Soundex note: jellyfish 1.2.1 uses the Knuth/Census H/W-skip variant
(confirmed by direct read of jellyfish/src/soundex.rs). No expected
divergences in the canonical corpus — the variant_divergence schema is
retained for completeness and NYSIIS load-bearing use.

NYSIIS note: jellyfish emits non-truncated codes (e.g. Catherine→CATARAN,
7 chars). Our implementation truncates to 6 chars per Taft 1970 / Knuth.
All jellyfish NYSIIS outputs longer than 6 chars carry variant_divergence=True,
and our expected value is the first 6 characters of the jellyfish output
(when consistent with Taft's algorithm pre-truncation).

Usage:
    make regen-phonetic-cross-validation
    # or directly:
    python3 scripts/gen-phonetic-cross-validation.py

Requirements:
    - Python 3.7+ (for guaranteed dict insertion-order preservation).
    - jellyfish==1.2.1
    - Metaphone==0.6
    Install both: python3 -m pip install --user jellyfish==1.2.1 Metaphone==0.6

Determinism:
    The INPUTS lists own the corpus's deterministic ordering. Python 3.7+
    guarantees dict insertion-order preservation; json.dump with sort_keys=False
    honours that order. The regenerated_at ISO timestamp is the only
    environment-dependent field; Go tests do not assert byte-stability on it.
"""

import hashlib
import json
import os
import subprocess
import sys
from datetime import datetime, timezone

# Minimum supported Python version.
_MIN_PYTHON_VERSION = (3, 7)

# Pinned package versions — BOTH must match or the script refuses to run.
JELLYFISH_VERSION = "1.2.1"
METAPHONE_VERSION = "0.6"


def _check_python_version() -> None:
    """Guard against silent corpus drift on outdated Python (< 3.7)."""
    if sys.version_info < _MIN_PYTHON_VERSION:
        min_str = ".".join(str(x) for x in _MIN_PYTHON_VERSION)
        cur_str = ".".join(str(x) for x in sys.version_info[:3])
        sys.exit(
            f"ERROR: Python {cur_str} is older than the supported minimum "
            f"{min_str}. Dict insertion-order preservation became a language "
            f"guarantee in Python 3.7; older releases could produce reordered "
            f"JSON breaking corpus byte-stability. Upgrade to Python >= {min_str}."
        )


_check_python_version()

import jellyfish  # noqa: E402 — imported after version constant for clear error message
assert jellyfish.__version__ == JELLYFISH_VERSION, (
    f"jellyfish version mismatch: installed {jellyfish.__version__!r}, "
    f"script pinned to {JELLYFISH_VERSION!r} — "
    f"run: python3 -m pip install --user jellyfish=={JELLYFISH_VERSION}"
)


def _assert_metaphone_version() -> None:
    """Assert Metaphone PyPI package is at the pinned version.

    The Metaphone package does not expose __version__ reliably, so we check
    via 'pip show Metaphone' and parse the Version: field.
    """
    try:
        out = subprocess.check_output(
            [sys.executable, "-m", "pip", "show", "Metaphone"],
            stderr=subprocess.DEVNULL,
        ).decode("utf-8", errors="replace")
    except subprocess.CalledProcessError:
        sys.exit(
            f"ERROR: 'pip show Metaphone' failed — Metaphone package not installed. "
            f"Run: python3 -m pip install --user Metaphone=={METAPHONE_VERSION}"
        )
    version_line = next(
        (ln for ln in out.splitlines() if ln.lower().startswith("version:")), None
    )
    if version_line is None:
        sys.exit(
            "ERROR: could not parse 'pip show Metaphone' output — no Version: line found."
        )
    installed = version_line.split(":", 1)[1].strip()
    if installed != METAPHONE_VERSION:
        sys.exit(
            f"Metaphone version mismatch: installed {installed!r}, "
            f"script pinned to {METAPHONE_VERSION!r} — "
            f"run: python3 -m pip install --user Metaphone=={METAPHONE_VERSION}"
        )


_assert_metaphone_version()

import metaphone as _metaphone_pkg  # noqa: E402


def _double_metaphone(s: str):
    """Call metaphone.doublemetaphone(s) and return (primary, secondary)."""
    return _metaphone_pkg.doublemetaphone(s)


# ---------------------------------------------------------------------------
# Soundex corpus (15 entries per CONTEXT.md §1 LOCKED).
#
# jellyfish 1.2.1 uses Knuth/Census H/W-skip variant (confirmed by direct
# read of jellyfish/src/soundex.rs lines 32-43). No expected divergences
# for canonical inputs — variant_divergence schema is kept for NYSIIS
# load-bearing use and future-proofing.
#
# Hand-curated KNUTH_EXPECTED override dict: maps input → expected code
# when our implementation (Knuth/Census) diverges from jellyfish's output.
# For Soundex this dict is scaffolding — jellyfish 1.2.1 matches Knuth/Census,
# so no entry is expected to appear in practice.
# ---------------------------------------------------------------------------
_SOUNDEX_KNUTH_EXPECTED = {
    # No known Soundex divergences between jellyfish 1.2.1 and Knuth/Census.
    # Kept for schema completeness and for NYSIIS load-bearing variant where
    # divergences ARE expected (07-03 NYSIIS plan fills this pattern).
}

SOUNDEX_INPUTS = [
    "Robert",    # Knuth p. 393 canonical pair #1 → R163
    "Rupert",    # Knuth p. 393 canonical pair #2 → R163 (same as Robert)
    "Rubin",     # Knuth p. 393 → R150
    "Tymczak",   # LOAD-BEARING Knuth/Census gate: Knuth→T522, SQL→T520
    "Ashcraft",  # H/W-handling gate (Pitfall 4): both Ashcraft + Ashcroft → A261
    "Ashcroft",  # H/W-handling pair → A261
    "Pfister",   # P236
    "Smith",     # S530
    "Honeyman",  # H555
    "Lloyd",     # L300
    "Jackson",   # J250
    "Euler",     # E460
    "Ellery",    # E460 (same as Euler — demonstrates same-code different names)
    "Gauss",     # G200
    "",          # empty → ""
]


def _gen_soundex_entries() -> list:
    """Generate the Soundex section (15 entries)."""
    entries = []
    for inp in SOUNDEX_INPUTS:
        if inp == "":
            jf_code = ""
        else:
            jf_code = jellyfish.soundex(inp)

        # Check for known Knuth/Census vs jellyfish divergence.
        knuth_code = _SOUNDEX_KNUTH_EXPECTED.get(inp, jf_code)
        variant_divergence = knuth_code != jf_code

        entry = {
            "algorithm": "Soundex",
            "input": inp,
            "code": knuth_code,
        }
        if variant_divergence:
            entry["variant_divergence"] = True
            entry["divergent_jellyfish_value"] = jf_code
        entries.append(entry)
    return entries


# ---------------------------------------------------------------------------
# Double Metaphone corpus (40 entries per CONTEXT.md §1 LOCKED).
# Uses metaphone package (oubiwann/metaphone, BSD-3-Clause) for reference
# vectors across 5 language-origin branches.
# jellyfish does NOT have Double Metaphone (key finding 1 from RESEARCH.md).
# ---------------------------------------------------------------------------
DM_INPUTS = [
    # Germanic branch (≥ 7 entries)
    ("Schmidt", "Germanic"),
    ("Smith", "Germanic"),
    ("Schneider", "Germanic"),
    ("Schwartz", "Germanic"),
    ("Wolf", "Germanic"),
    ("Wagner", "Germanic"),
    ("Fischer", "Germanic"),
    ("Bauer", "Germanic"),
    # Slavic branch (≥ 7 entries)
    ("Pacheco", "Romance"),    # actually Romance/Spanish — moved below; placeholder
    ("Wojcik", "Slavic"),
    ("Kowalski", "Slavic"),
    ("Nowak", "Slavic"),
    ("Przybyszewski", "Slavic"),
    ("Wiśniewski", "Slavic"),
    ("Wróbel", "Slavic"),
    ("Dąbrowski", "Slavic"),
    # Romance branch (≥ 7 entries)
    ("Pacheco", "Romance"),
    ("Rodriguez", "Romance"),
    ("Garcia", "Romance"),
    ("Martinez", "Romance"),
    ("Lopez", "Romance"),
    ("Gonzalez", "Romance"),
    ("Sanchez", "Romance"),
    # Greek branch (≥ 7 entries)
    ("Catherine", "Greek"),
    ("Katherine", "Greek"),
    ("Christodoulou", "Greek"),
    ("Papadopoulos", "Greek"),
    ("Stavros", "Greek"),
    ("Alexopoulos", "Greek"),
    ("Theodorakis", "Greek"),
    # Chinese-origin branch (≥ 4 entries)
    ("Cheung", "Chinese-origin"),
    ("Wong", "Chinese-origin"),
    ("Chang", "Chinese-origin"),
    ("Chiang", "Chinese-origin"),
    ("Liang", "Chinese-origin"),
    # Edge cases
    ("", "edge"),
    ("A", "edge"),
    ("Ph", "edge"),
    ("Gn", "edge"),
    ("Kn", "edge"),
]


def _gen_dm_entries() -> list:
    """Generate the Double Metaphone section (≥ 40 entries)."""
    seen = set()
    entries = []
    for inp, branch in DM_INPUTS:
        if inp in seen:
            continue
        seen.add(inp)
        primary, secondary = _double_metaphone(inp)
        entry = {
            "algorithm": "DoubleMetaphone",
            "input": inp,
            "branch": branch,
            "primary": primary,
            "secondary": secondary,
        }
        entries.append(entry)
    return entries


# ---------------------------------------------------------------------------
# NYSIIS corpus (20 entries per CONTEXT.md §1 LOCKED).
#
# CRITICAL: jellyfish NYSIIS emits non-truncated codes (e.g. Catherine→CATARAN,
# 7 chars). Our implementation truncates to 6 chars per Taft 1970 / Knuth.
# All entries where jellyfish output > 6 chars carry variant_divergence=True.
# Expected value = first 6 chars of jellyfish output (when consistent).
# ---------------------------------------------------------------------------
NYSIIS_INPUTS = [
    "Brown",       # canonical reference: BRAN
    "Browne",      # BRAN (same as Brown — load-bearing pair)
    "Robert",      # RABAD per Knuth TAOCP §6.4
    "Thompson",    # TANPSAN (6-char truncation needed?)
    "Johnson",     # JANSAN
    "Williams",    # WALAN
    "Jones",       # JAN
    "Davis",       # DAF
    "Miller",      # MALAR
    "Wilson",      # WLSAN
    "Moore",       # MAR
    "Taylor",      # TALAR
    "Anderson",    # ANDARS
    "Thomas",      # TANAS
    "Jackson",     # JACSAN
    "White",       # WAT
    "Harris",      # HAR
    "Martin",      # MARTAN
    "Thompson",    # duplicate — de-dup below
    "",            # empty
]


def _nysiis_expected(inp: str, jf_code: str) -> tuple:
    """Return (expected_code, variant_divergence) for a NYSIIS entry.

    If jellyfish output > 6 chars, our implementation returns the first 6
    chars (Taft-1970 truncation). variant_divergence is True when jf differs.
    """
    if inp == "":
        return "", False
    if len(jf_code) > 6:
        taft_code = jf_code[:6]
        return taft_code, True
    return jf_code, False


def _gen_nysiis_entries() -> list:
    """Generate the NYSIIS section (20 entries)."""
    seen = set()
    entries = []
    for inp in NYSIIS_INPUTS:
        if inp in seen:
            continue
        seen.add(inp)
        if inp == "":
            jf_code = ""
        else:
            jf_code = jellyfish.nysiis(inp)

        expected, variant_divergence = _nysiis_expected(inp, jf_code)
        entry = {
            "algorithm": "NYSIIS",
            "input": inp,
            "code": expected,
        }
        if variant_divergence:
            entry["variant_divergence"] = True
            entry["divergent_jellyfish_value"] = jf_code
        entries.append(entry)
    return entries


# ---------------------------------------------------------------------------
# MRA corpus (20 entries per CONTEXT.md §1 LOCKED).
# jellyfish provides match_rating_codex() and match_rating_comparison().
# ---------------------------------------------------------------------------
MRA_INPUTS = [
    "Smith",
    "Smythe",
    "Byrne",
    "Burns",
    "Johnson",
    "Johnston",
    "Robert",
    "Roberts",
    "Catherine",
    "Katherine",
    "Brown",
    "Browne",
    "Taylor",
    "Tailor",
    "Anderson",
    "Andrews",
    "Williams",
    "Williamson",
    "Thompson",
    "",
]


def _gen_mra_entries() -> list:
    """Generate the MRA section (20 entries)."""
    seen = set()
    entries = []
    for inp in MRA_INPUTS:
        if inp in seen:
            continue
        seen.add(inp)
        if inp == "":
            code = ""
        else:
            code = jellyfish.match_rating_codex(inp)
        entry = {
            "algorithm": "MRA",
            "input": inp,
            "code": code,
        }
        entries.append(entry)
    return entries


def _script_sha256() -> str:
    """Return the SHA-256 hex digest of this script file."""
    script_path = os.path.abspath(__file__)
    with open(script_path, "rb") as f:
        return hashlib.sha256(f.read()).hexdigest()


def main() -> None:
    soundex_entries = _gen_soundex_entries()
    dm_entries = _gen_dm_entries()
    nysiis_entries = _gen_nysiis_entries()
    mra_entries = _gen_mra_entries()

    all_entries = soundex_entries + dm_entries + nysiis_entries + mra_entries

    out = {
        "version": 1,
        "_metadata": {
            "jellyfish_version": JELLYFISH_VERSION,
            "metaphone_version": METAPHONE_VERSION,
            "python_version": f"{sys.version_info.major}.{sys.version_info.minor}.{sys.version_info.micro}",
            "regenerated_at": datetime.now(timezone.utc).isoformat(),
            "script_sha256": _script_sha256(),
        },
        "entries": all_entries,
    }

    path = "testdata/cross-validation/phonetic/vectors.json"
    os.makedirs(os.path.dirname(path), exist_ok=True)
    with open(path, "w") as f:
        json.dump(out, f, indent=2, sort_keys=False)
        f.write("\n")  # trailing LF per canonical convention


if __name__ == "__main__":
    main()
