---
phase: 07-phonetic-algorithms
reviewed: 2026-05-15T20:07:42Z
depth: standard
files_reviewed: 19
files_reviewed_list:
  - soundex.go
  - soundex_test.go
  - dispatch_soundex.go
  - double_metaphone.go
  - double_metaphone_test.go
  - dispatch_double_metaphone.go
  - nysiis.go
  - nysiis_test.go
  - dispatch_nysiis.go
  - mra.go
  - mra_test.go
  - dispatch_mra.go
  - monge_elkan.go
  - monge_elkan_test.go
  - phonetic_cross_validation_test.go
  - phonetic_codes_golden_test.go
  - examples/phonetic-keys/main.go
  - examples/identifier-similarity/main.go
  - scripts/gen-phonetic-cross-validation.py
findings:
  critical: 3
  warning: 4
  info: 2
  total: 9
status: issues_found
---

# Phase 7: Code Review Report

**Reviewed:** 2026-05-15T20:07:42Z
**Depth:** standard
**Files Reviewed:** 19
**Status:** issues_found

## Summary

Phase 7 delivers four phonetic algorithms (Soundex, Double Metaphone, NYSIIS, MRA) plus MongeElkan allow-list updates. The licensing discipline, primary-source citations, and overall architecture are sound. The Soundex and MRA implementations are correct. However, two algorithmic correctness bugs were found — one in NYSIIS (two suffix rules produce the wrong trailing character) and one in Double Metaphone (the SC-default branch consumes one extra character). Additionally, a semantic bug in MRA allows `MRAScore("", s)` to return `1.0` for very short single-character inputs, violating the one-empty → 0.0 invariant. Three quality warnings cover dead code in Double Metaphone and the cross-validation script.

---

## Critical Issues

### CR-01: NYSIIS suffix rules RD→D and ND→D leave wrong trailing character

**File:** `nysiis.go:207-222`

**Issue:** The suffix transliteration cases for `RD` and `ND` both omit the assignment `work[n-2] = 'D'` before the slice truncation. This leaves `R` (for `RD`) or `N` (for `ND`) as the trailing character instead of `D`.

Concrete trace for `"BYRD"`:
- `work[n-2]='R'`, `work[n-1]='D'`
- Code executes `work = work[:n-1]` → drops `'D'`, result ends in `'R'`
- Expected: `"BYD"` — the suffix `RD` becomes a single `D`
- Actual: `"BYR"` — the `R` survives, `D` is discarded

Contrast with the **correct** `RT→D` and `NT→D` cases (lines 202–206, 213–216), which perform `work[n-2] = 'D'` first:
```go
// RT → D (correct):
work[n-2] = 'D'
work = work[:n-1]
```

The same fix is required for both `RD` and `ND`.

**Fix:**
```go
// Line 207-212 — RD → D (BUGGY):
case n >= 2 && work[n-2] == 'R' && work[n-1] == 'D':
    work = work[:n-1]   // BUG: drops 'D', leaves 'R'
    n--

// CORRECT:
case n >= 2 && work[n-2] == 'R' && work[n-1] == 'D':
    work[n-2] = 'D'    // replace R with D
    work = work[:n-1]  // drop the now-redundant last slot
    n--

// Line 218-222 — ND → D (BUGGY):
case n >= 2 && work[n-2] == 'N' && work[n-1] == 'D':
    work = work[:n-1]   // BUG: drops 'D', leaves 'N'
    n--

// CORRECT:
case n >= 2 && work[n-2] == 'N' && work[n-1] == 'D':
    work[n-2] = 'D'    // replace N with D
    work = work[:n-1]  // drop the now-redundant last slot
    n--
```

**Coverage gap:** Neither the unit test corpus (`nysiis_test.go`) nor the cross-validation corpus (`NYSIIS_INPUTS` in the Python script) includes any `RD`-ending or `ND`-ending surname. Representative additions: `"Byrd"` (expected `"BYD"`), `"Ward"` (expected `"WAD"`), `"Bond"` (expected `"BAD"`), `"Brand"` (expected `"BRAND"`). Without these test cases the bug is currently invisible.

---

### CR-02: Double Metaphone `SC` default branch consumes one too many characters (`i += 3` should be `i += 2`)

**File:** `double_metaphone.go:735-736`

**Issue:** When `SC` is encountered in a word but is **not** followed by `I`, `E`, or `Y`, the code emits `"SK"` and then advances by `i += 3`. However `SC` is only two characters, so `i += 3` silently consumes the character immediately after `SC` without processing it.

```go
// Current (buggy):
if dmContains(v, i, "SC") {
    if at(i+2) == 'I' || at(i+2) == 'E' || at(i+2) == 'Y' {
        dmAdd(&p, &alt, "S", "")
        i += 3   // correct: consumes S + C + vowel (3 chars)
        continue
    }
    dmAdd(&p, &alt, "SK", "")
    i += 3       // BUG: SC is 2 chars, not 3; skips char at i+2
    continue
}
```

Concrete trace for `"DISCIPLE"` (v = `"DISCIPLE"`):
- At `i=2` (`'S'`): `SC` detected, `at(4)='I'` → vowel branch fires, `i += 3` → `i=5`. The `'I'` at position 4 is correctly consumed as part of `SCI`.
- This particular input takes the **vowel path** and is correct.

For `"DISCO"` (v = `"DISCO"`):
- At `i=2` (`'S'`): `SC` detected, `at(4)='O'` → not I/E/Y → `"SK"` emitted, `i += 3` → `i=5`.
- The `'O'` at position 4 is consumed without adding to the phonetic key; only `'D','I'` remain.
- Expected: advance `i += 2` → `i=4`, then `'O'` is processed normally by the vowel default.

The Philips 2000 C reference advances by 2 for the `SC`-non-vowel case (only the `S` and `C` are consumed; the following character is left for the next iteration).

**Fix:**
```go
dmAdd(&p, &alt, "SK", "")
i += 2   // SC is 2 chars; next char is processed separately
continue
```

---

### CR-03: `MRAScore("", s)` returns `1.0` for single-character inputs — violates one-empty → 0.0 invariant

**File:** `mra.go:340-349`

**Issue:** `MRAScore` is specified to return `0.0` when one input is empty (`algorithm-correctness-standards` §"Edge cases", `docs/requirements.md` §7). However the NBS Tech Note 943 comparison algorithm contains no explicit empty-codex guard. When one side has an empty codex and the other side encodes to a length-1 codex, the length-difference gate (`diff >= 3`) does not fire (`|0-1| = 1 < 3`), and the similarity calculation yields `6 - max(0, 1) = 5`, which meets the threshold of 5 (from `mraThresholdTable[1]`), producing `matched = true` → `MRAScore = 1.0`.

Concrete example:
```go
MRAScore("", "i")   // → 1.0 (bug: should be 0.0)
MRAScore("", "a")   // → 1.0 (bug: should be 0.0)
// MRACode("i") = "I" (len 1), MRACode("") = "" (len 0)
// diff=1 < 3; threshold(1)=5; sim=6-1=5 >= 5 → match
```

The same applies to `MRACompare("", s)` for any single-char `s` that encodes to a length-1 codex. Length-2 inputs are safe (sim=4 < threshold=5), and any input encoding to length ≥ 3 triggers the `diff >= 3` auto-mismatch gate.

The test suite has no `MRAScore("", one_char_input)` case.

**Fix:** Add a guard in `MRACompare` after computing `lenA` and `lenB`:
```go
// Step 0: one-empty guard — per algorithm-correctness-standards, one-empty → 0.0.
// The NBS 943 length-difference gate only fires at diff >= 3; for (0,1) and (0,2)
// the algorithm would otherwise produce a spurious match.
if lenA == 0 || lenB == 0 {
    // empty codex means the original input had no ASCII letters;
    // empty vs non-empty is never a match per the catalogue convention.
    return false, 0
}
```

Alternatively, add the guard in `MRAScore` directly before calling `MRACompare`, consistent with how `SoundexScore` and `NYSIISScore` handle this via `ca != "" && ca == cb`.

---

## Warnings

### WR-01: `dmSlgCheck` dead WITZ branch — the `W` check makes it unreachable

**File:** `double_metaphone.go:128-132`

**Issue:** In `dmSlgCheck`, the first condition `c == 'W' || c == 'K'` already returns `true` for any `'W'` character. The subsequent check `c == 'W' && i+3 < len(s) && s[i+1] == 'I' && ...` (line 129-132) is unreachable: any name containing the substring `WITZ` will have `W` matched by the first condition before the second condition is evaluated.

This means the `WITZ` pattern was intended as a separate SlavoGermanic indicator but accidentally became dead code. Functionally this produces the correct result (names with `WITZ` trigger SlavoGermanic because they contain `W`) but the stated intent — to specifically recognise `WITZ` as a multi-character signal — is not achieved. It also means `WITZ` could be silently removed from the check in the future without a test failure.

**Fix:**
```go
// Remove the dead WITZ branch, or restructure to make WITZ detection meaningful:
func dmSlgCheck(s string) bool {
    for i := 0; i < len(s); i++ {
        c := s[i]
        if c == 'K' {
            return true
        }
        if c == 'C' && i+1 < len(s) && s[i+1] == 'Z' {
            return true
        }
        // WITZ/WICZ detected as multi-char pattern (W alone is also sufficient,
        // but retaining explicit WITZ/WICZ for documentation clarity):
        if c == 'W' {
            return true
        }
    }
    return false
}
```

If the WITZ check was intentionally kept for documentation, add a comment explaining why it is redundant and can never fire.

---

### WR-02: `DoubleMetaphoneKeys` unused `padded` variable — dead allocation

**File:** `double_metaphone.go:259-274`

**Issue:** The `padded` variable is computed (`padded := "  " + v + "     "`) and then immediately suppressed with `_ = padded`. This is a string concatenation that allocates a heap string on every call to `DoubleMetaphoneKeys`, even though the code was refactored to work directly on `v` with the bounds-safe `dmContains` helper and `at()` closure. The allocation is wasted on every invocation.

```go
// Lines 259-274:
padded := "  " + v + "     "  // allocates every call
// ...
_ = padded  // suppresses "declared but not used" — confirmed dead code
```

**Fix:** Remove both lines entirely:
```go
// Delete:
// padded := "  " + v + "     "
// _ = padded
```

---

### WR-03: `default C` branch in Double Metaphone contains space-literal dead checks that can never fire

**File:** `double_metaphone.go:429-433`

**Issue:** The default `C` branch checks for `" C"`, `" Q"`, and `" G"` patterns after the current position:
```go
if dmContains(v, i+1, " C") || dmContains(v, i+1, " Q") || dmContains(v, i+1, " G") {
    i += 3
} else {
    i++
}
```
The string `v` comes from `dmPrep`, which filters input to ASCII letters `[A-Z]` only — spaces are never present. Therefore `dmContains(v, i+1, " C")` always returns `false`, the `i += 3` branch is unreachable, and `i++` always fires.

This is dead code inherited from Philips' C reference which operated on the raw (space-containing) input string. After `dmPrep` normalisation, these checks have no effect.

**Fix:** Remove the unreachable branch:
```go
// Replace:
if dmContains(v, i+1, " C") || dmContains(v, i+1, " Q") || dmContains(v, i+1, " G") {
    i += 3
} else {
    i++
}
// With:
i++
```

---

### WR-04: Cross-validation script has a stale placeholder entry and non-ASCII Slavic inputs

**File:** `scripts/gen-phonetic-cross-validation.py:252-257`

**Issue 1:** `DM_INPUTS` contains `("Pacheco", "Romance")` at line 252 with the comment `# actually Romance/Spanish — moved below; placeholder`. This entry is in the Slavic block but labelled Romance. The `seen` set in `_gen_dm_entries` deduplicates it correctly (the second `"Pacheco"` at line 261 is skipped), so only one Romance Pacheco entry appears in the corpus. However, this leaves the Slavic block with only 7 entries instead of the stated 8: `Wojcik`, `Kowalski`, `Nowak`, `Przybyszewski`, `Wiśniewski`, `Wróbel`, `Dąbrowski` = 7 real Slavic entries. The branch-count guard in `TestPhonetic_CrossValidation` requires `>= 7` for Slavic, so this currently passes — but the stale placeholder and comment cause confusion.

**Issue 2:** Three Slavic entries — `"Wiśniewski"`, `"Wróbel"`, `"Dąbrowski"` — contain non-ASCII characters. The Python `metaphone.doublemetaphone()` library processes these with their diacritics intact, but the Go implementation's `dmPrep` strips non-ASCII bytes, producing a different normalised form. This creates a latent mismatch risk: if the Python `metaphone` package processes `ś` differently from the Go implementation's silent-drop, the generated corpus entries will produce test failures when the corpus is regenerated.

**Fix:** Replace the stale placeholder with a genuine Slavic entry and use pure-ASCII Slavic surnames for the corpus:
```python
# Replace the stale line 252 with a real Slavic entry:
("Sczepanski", "Slavic"),  # replaces the misplaced Pacheco placeholder

# Replace non-ASCII Slavic entries with ASCII alternatives:
("Wojcik", "Slavic"),
("Kowalski", "Slavic"),
("Nowak", "Slavic"),
("Przybyszewski", "Slavic"),
("Wisniewski", "Slavic"),   # ASCII approximation of Wiśniewski
("Wrobel", "Slavic"),       # ASCII approximation of Wróbel
("Dabrowski", "Slavic"),    # ASCII approximation of Dąbrowski
```

---

## Info

### IN-01: Missing test coverage for NYSIIS `RD`-ending and `ND`-ending surnames

**File:** `nysiis_test.go`

**Issue:** The unit test corpus and cross-validation input list (`NYSIIS_INPUTS`) do not include any surname ending in `RD` or `ND`. This is the direct reason CR-01 was not caught by the test suite. The `TestNYSIISCode_KnuthReferenceVectors` table has 12 entries but none exercises the `RD→D` or `ND→D` suffix rules.

**Fix:** Add reference vector entries:
```go
{
    name:       "RV-N13: Byrd",
    input:      "Byrd",
    want:       "BYD",   // (or verify with jellyfish after CR-01 is fixed)
    derivation: "B→B; Y→Y; RD suffix → D; dedup: BYD",
},
{
    name:       "RV-N14: Bond",
    input:      "Bond",
    want:       "BAD",   // B→B; O→A(vowel); ND suffix → D
    derivation: "B→B; O→A(vowel); ND suffix → D; dedup: BAD",
},
```

Regenerate `NYSIIS_INPUTS` in the Python script to include `"Byrd"` and `"Bond"` after fixing CR-01.

---

### IN-02: Missing test coverage for `MRAScore` one-empty edge case with short inputs

**File:** `mra_test.go`

**Issue:** No test asserts `MRAScore("", "i")` or `MRAScore("", "a")` returns `0.0`. The existing `TestMRAScore_LiteratureReferenceVectors` tests only the `"Ad"` / `"ZachariahMontgomery"` non-empty pair (which correctly returns `0.0` because the length-difference gate fires at `diff = 4 >= 3`). The one-empty → 0.0 invariant for single-character inputs is not covered.

**Fix:** Add a test case after implementing CR-03:
```go
// In TestMRAScore_LiteratureReferenceVectors or a new TestMRAScore_OneEmpty:
{"one_empty_single_char_a", "", "i",  0.0, "one-empty: empty vs single-char → 0.0"},
{"one_empty_single_char_b", "a", "", 0.0, "one-empty: single-char vs empty → 0.0"},
```

---

_Reviewed: 2026-05-15T20:07:42Z_
_Reviewer: Claude (gsd-code-reviewer)_
_Depth: standard_
