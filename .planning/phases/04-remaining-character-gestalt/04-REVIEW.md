---
phase: 04-remaining-character-gestalt
reviewed: 2026-05-14T16:08:26Z
depth: standard
files_reviewed: 30
files_reviewed_list:
  - CONTRIBUTING.md
  - Makefile
  - algoid_test.go
  - algorithms_golden_test.go
  - cross_algorithm_consistency_test.go
  - dispatch_lcsstr.go
  - dispatch_ratcliff_obershelp.go
  - dispatch_strcmp95.go
  - example_test.go
  - examples/identifier-similarity/main.go
  - examples/identifier-similarity/main_test.go
  - export_test.go
  - lcsstr.go
  - lcsstr_bench_test.go
  - lcsstr_fuzz_test.go
  - lcsstr_test.go
  - llms-full.txt
  - llms.txt
  - props_test.go
  - ratcliff_obershelp.go
  - ratcliff_obershelp_bench_test.go
  - ratcliff_obershelp_fuzz_test.go
  - ratcliff_obershelp_test.go
  - scripts/gen-ratcliff-obershelp-cross-validation.py
  - strcmp95.go
  - strcmp95_bench_test.go
  - strcmp95_fuzz_test.go
  - strcmp95_test.go
  - testdata/cross-validation/ratcliff-obershelp/vectors.json
  - testdata/golden/_staging/lcsstr.json
  - testdata/golden/_staging/ratcliff_obershelp.json
  - testdata/golden/_staging/strcmp95.json
  - testdata/golden/algorithms.json
  - tests/bdd/features/lcsstr.feature
  - tests/bdd/features/ratcliff_obershelp.feature
  - tests/bdd/features/strcmp95.feature
  - tests/bdd/steps/algorithms_steps.go
findings:
  critical: 0
  warning: 4
  info: 6
  total: 10
status: issues_found
---

# Phase 04: Code Review Report

**Reviewed:** 2026-05-14T16:08:26Z
**Depth:** standard
**Files Reviewed:** 30
**Status:** issues_found

## Summary

Phase 04 delivers three algorithms (Strcmp95 / Winkler 1994, LCSStr /
Wagner-Fischer 1974, Ratcliff-Obershelp / 1988) plus a Python difflib
cross-validation harness and phase-4 finalisation artefacts (merged
golden, cross-algorithm consistency tests, examples, llms.txt sync,
bench.txt regen).

Overall the implementations are correct, well-documented, and densely
property-tested. Primary-source citations are present, golden files
merge cleanly, BDD coverage is complete, and the asymmetric-by-design
contract for Ratcliff-Obershelp (OQ-1 LOCKED 2026-05-14) is explicitly
preserved.

No BLOCKER-level defects were found. The findings below cluster
around (a) discipline gaps in source-origin attribution, (b) a small
amount of provably-dead-but-quadratic-cost code in the LCSStr two-row
DP kernel, (c) one minor inconsistency between fuzz-corpus seeds and
the autojunk-sensitive cross-validation fixture, and (d) several
documentation / naming polish items.

The performance concerns flagged in the Ratcliff-Obershelp benchmark
comments (heap allocation per recursion level) are out of v1 scope
per the review brief but are noted for v1.x follow-up.

## Warnings

### WR-01: Source Origin Statement omits `richmilne/JaroWinkler`

**File:** `strcmp95.go:287-288`
**Issue:** The inline godoc on `strcmp95Bytes` cites
`richmilne/JaroWinkler` as a reference consulted "for the algorithm
structure", but this third-party reference does NOT appear in the
file-header Source Origin Statement (lines 89-97). The header lists
only Winkler 1994 (primary), Census Bureau strcmp95.c
(cross-validation), and OpenRefine Strcmp95.java (tie-break). Per
`.claude/skills/algorithm-licensing-standards/SKILL.md` ("Attribution
Format" §73-86), every consulted reference must appear in the file
header's source-origin block — the algorithm-licensing-reviewer relies
on the header for the consulted-source audit. Failure mode: a future
licensing-screen review reads only the header, misses the
`richmilne/JaroWinkler` reference, and cannot verify its licence.

**Fix:** Add `richmilne/JaroWinkler` to the header's source-origin
block, including its licence ("MIT" per a quick GitHub search) and
the scope of consultation:

```go
// Sources:
//   - Winkler, W. E. (1994) ...
//   - Cross-validation reference: U.S. Census Bureau (1995) strcmp95.c ...
//   - OpenRefine Strcmp95.java (Apache-2.0) ...
//   - richmilne/JaroWinkler (MIT, https://github.com/richmilne/JaroWinkler)
//     consulted ONLY for the algorithm structure (step ordering of the
//     four Winkler 1994 adjustments); no code copied, no variable names
//     or comment phrasing derived.
```

Verify the repo's actual licence before pasting the SPDX identifier.

### WR-02: `lcsstrDP` redundant row-clear is `O(n)` dead code in the hot path

**File:** `lcsstr.go:255-262` and `lcsstr.go:283-287`
**Issue:** After the `prev, curr = curr, prev` swap, the code zeros
all `n+1` entries of the new `curr`. This is unconditionally dead
work: the next outer iteration's inner loop writes `curr[j]` for
every `j` in `1..n` (either `curr[j] = prev[j-1] + 1` on a byte match
or `curr[j] = 0` on a mismatch), and `curr[0]` is never read. The
stack buffer is zero-initialised once at `var buf [...]int`
declaration; the heap rows from `make([]int, n+1)` are also
zero-init. The "stale values from two rows back" failure mode the
comment cites cannot occur because all `j >= 1` positions are
overwritten on every iteration.

This is not a correctness bug, but it doubles the inner-loop cost
(`O(2n)` per row instead of `O(n)`) and silently consumes 50% of the
LCSStr benchmark budget on the canonical ASCII paths. Because the
same pattern appears in both the byte path (`lcsstrDP`) and the rune
path (`lcsstrDPRunes`) and was also copied into
`roFindLongestMatch` / `roFindLongestMatchRunes` in
ratcliff_obershelp.go (lines 255-257, 305-308), removing it would
affect four kernels.

**Fix:** Remove the redundant clear loops; rely on the unconditional
write of `curr[j]` in the inner loop. If defensive zeroing is
desired (e.g. to silence reviewers reading the code top-to-bottom),
convert the dead clear to a single-line comment explaining why the
buffers do NOT need re-initialisation:

```go
prev, curr = curr, prev
// No need to zero `curr` — the inner loop writes curr[j] for every
// j in 1..n unconditionally (matched branch writes prev[j-1]+1;
// mismatched branch writes 0). curr[0] is never read.
```

If keeping the defensive zero, add a benchmark gate that pins the
overhead so a future reviewer doesn't try to optimise it away
without realising it's deliberate.

### WR-03: Strcmp95 silently swallows similar-character credit when `m == 0`

**File:** `strcmp95.go:328-330`
**Issue:** The early return `if m == 0 { return 0.0 }` fires AFTER
the Jaro matching pass but BEFORE the similar-character credit pass.
Consequence: a pair like "WO"/"UE" (where W~U and O~E are both in
the Winkler 1994 similar-character table) returns exactly 0.0 even
though every character has a documented similar-pair credit
available. The Census Bureau strcmp95.c reference is silent on
whether similar-credit fires when m == 0; this implementation chose
"no". That choice is defensible (and matches the canonical Census
Bureau early-exit) but it is not documented anywhere in the godoc.

This is a documentation gap that could surface as a confusing
score on niche input pairs ("Two strings with NO matching characters
but every character pair in the similar table — score = 0?").

**Fix:** Document the "m == 0 short-circuit precedes similar-credit"
behaviour in `Strcmp95Score`'s godoc. Either accept the current
behaviour with an explicit note, or change the implementation to
run the similar-credit pass when m == 0 (in which case the third
Jaro term becomes `(numCom - 0)/numCom = 1`, and the score would be
`(numCom/la + numCom/lb + 1) / 3` — non-zero. This deviates from
strcmp95.c, so document the deviation if chosen).

Add a unit test pinning the chosen behaviour on the W/U + O/E case:

```go
func TestStrcmp95_AllSimilarNoMatches_ScoresZero(t *testing.T) {
    // Document the m==0 short-circuit precedes similar-credit:
    // "WO"/"UE" — W~U and O~E are both in the Winkler 1994 table,
    // but the Jaro pass finds no exact-byte matches (m=0) and the
    // early-return fires before similar-credit accumulates.
    if got := fuzzymatch.Strcmp95Score("WO", "UE"); got != 0.0 {
        t.Errorf("Strcmp95Score(\"WO\", \"UE\") = %g; want 0.0 (m==0 precedes similar-credit)", got)
    }
}
```

### WR-04: Cross-validation corpus lacks Python version pinning at parse time

**File:** `ratcliff_obershelp_test.go:304-332`
**Issue:** The Go test reads `python_version` from the corpus header
and includes it in error messages, but does NOT assert any minimum
version or validate the field is non-empty. If a future Python
patch changes `difflib.SequenceMatcher` behaviour (e.g. the matcher
ordering for ties — a real risk per CPython bpo-37004 history),
silently regenerating the corpus on a "wrong" Python version would
produce subtly-off `difflib_ratio` values, the Go test would compare
against those wrong values, and the test would PASS while the
algorithm's difflib parity contract was actually broken.

The script-side `_check_python_version()` (lines 179-203) enforces
Python ≥ 3.7 at generation time, but that check only runs when the
corpus is regenerated. The committed corpus carries `python_version:
"3.12.12"`, which is informational only.

**Fix:** Either (a) hard-pin the corpus header's Python version to
the version that produced the committed values, and assert at parse
time, or (b) at minimum assert the `python_version` field is non-empty
and `>= 3.7` so a malformed or stub corpus surfaces immediately:

```go
if c.PythonVersion == "" {
    t.Fatalf("TestRatcliffObershelp_CrossValidation: corpus missing python_version (regenerate with `make regen-ratcliff-obershelp-cross-validation`)")
}
// Parse the major.minor portion and assert >= 3.7 — the script-side
// minimum that guarantees dict-insertion-order stability. The third
// component is informational (patch-level).
parts := strings.SplitN(c.PythonVersion, ".", 3)
if len(parts) < 2 {
    t.Fatalf("malformed python_version %q", c.PythonVersion)
}
major, _ := strconv.Atoi(parts[0])
minor, _ := strconv.Atoi(parts[1])
if major < 3 || (major == 3 && minor < 7) {
    t.Fatalf("corpus generated with Python %s; require >= 3.7 for difflib stability", c.PythonVersion)
}
```

This belongs paired with WR-03's documentation note.

## Info

### IN-01: `strcmp95.go` transposition comment misattributes the `/2` halving

**File:** `strcmp95.go:351`
**Issue:** The comment `t /= 2 // each mismatch was counted in both
directions` is copied verbatim from `jaro.go:237`, but neither
implementation actually counts the mismatch twice — the single
forward-sweep records each transposition pair exactly once, and the
`/2` is the classical Jaro formula's `T/2` halving (Wikipedia: "half
the number of matching but different sequence order characters").
For "ab"/"ba" the sweep finds t=2 (A vs B at i=0, B vs A at i=1)
and `t/2 = 1`, which is the correct transposition count per Jaro's
spec. The comment misleads a reader who tries to verify the math.

**Fix:** Replace the misleading comment in both places with the
Jaro-canonical wording:

```go
t /= 2 // Jaro transposition count = (mismatched matched-pairs) / 2
```

This applies to both `strcmp95.go:351` and the upstream
`jaro.go:237` (out of scope for this phase but worth a follow-up
issue).

### IN-02: BDD HAMINGTON expected score is not pinned in `_test.go`

**File:** `strcmp95_test.go:100-129` and `tests/bdd/features/strcmp95.feature:28`
**Issue:** The four `TestStrcmp95_ReferenceVectors_CensusBureau`
entries cover MARTHA/MARHTA, DWAYNE/DUANE, DIXON/DICKSONX but NOT
the HAMINGTON/HAMMINGTON pair. The Strcmp95 BDD feature's Examples
table includes `HAMINGTON / HAMMINGTON | 0.9820`, and the staging
golden pins the exact value `0.9819696969696969`. The unit-test pin
for the long-string adjustment is at `TestStrcmp95_LongStringAdjustment_Triggers`
(lines 169-186) but only asserts the inequality `Strcmp95 > JaroWinkler`,
not the exact value. If the long-string-adjustment denominator
arithmetic drifts (e.g. by accidentally substituting `numCom` for `m`
in the denominator), the inequality might still hold but the value
would differ — caught by the BDD scenario and the golden but not by
the unit-test pin.

**Fix:** Add HAMINGTON/HAMMINGTON to
`TestStrcmp95_ReferenceVectors_CensusBureau`'s table with the exact
value `0.9819696969696969` and an absolute tolerance of `1e-9`. The
unit test is faster than BDD to diagnose and is the canonical place
to pin paper-cited reference vectors.

### IN-03: `LongestCommonSubstring` substring escape risk is not benchmarked

**File:** `lcsstr.go:115-143`
**Issue:** The `LongestCommonSubstring` byte-path returns
`a[endI-maxLen : endI]` — a slice header into `a`'s backing storage.
If the consumer retains this returned substring while `a` is GC'd,
the underlying `[]byte` is kept alive. For typical small-string
usage this is fine, but for a consumer that constructs an ephemeral
mega-string and calls `LongestCommonSubstring` to extract a tiny
shared segment, the entire mega-string is retained. This is standard
Go behaviour but is undocumented on the public surface.

**Fix:** Add a short godoc paragraph noting "the returned string
shares its backing storage with `a`; callers retaining the result
across the lifetime of `a` will keep `a`'s backing storage alive."
The phrasing matches Go-stdlib convention for `strings.SplitN` and
similar functions that return shared substrings.

### IN-04: `strcmp95Bytes` defensive clamp could mask an algorithm bug

**File:** `strcmp95.go:406-409`
**Issue:** After the three-term Jaro formula, the code clamps `j >
1.0` to `1.0` before applying the Winkler prefix boost. The inline
comment explains "similar-character credit may push the three-term
sum just past 1.0 on degenerate inputs". This is a defensive clamp
that hides any genuine overshoot: if a future refactor of the
similar-credit pass produces `j = 1.5`, the clamp pulls it back to
1.0 and downstream tests pass silently. The hierarchy property test
`TestProp_Strcmp95Score_AtLeastJaroWinkler` validates the lower
bound but not the upper bound.

**Fix:** Either (a) add a property test asserting the pre-clamp `j`
value is in `[0, 1 + ε]` for some small `ε` (catches genuine
overshoots while tolerating ULP-sized arithmetic noise), or (b)
strengthen the clamp to log/panic on `j > 1.0 + 1e-9` in debug
builds. Option (a) is simpler and aligns with the project's
"property tests as safety net" philosophy. Export a test hook via
`export_test.go` that returns the pre-clamp value.

### IN-05: Hand-curated `AtLeastLevenshtein` property is not asserted as a UNIVERSAL property

**File:** `props_test.go:1508-1541`
**Issue:** `TestProp_RatcliffObershelpScore_AtLeastLevenshtein_HandCurated`
covers six pairs and asserts RO >= Lev on each. The function name
contains "Prop" (suggesting `testing/quick` property testing) but the
implementation is a deterministic table-driven test, not a generator-
driven property. The mismatched naming risks confusing a future
maintainer who expects to find a generator and discovers a fixed
table.

The godoc explains the rationale ("the property is 'generally' true,
not universal"), but the symbol name still asserts a property
contract.

**Fix:** Rename to `TestRatcliffObershelpScore_AtLeastLevenshtein_OnSubstringContainment`
(no `Prop_` prefix) so the symbol name matches the actual test
shape — table-driven, not property-driven. The rationale paragraph
in the godoc stays.

### IN-06: `examples/identifier-similarity/main_test.go` uses fragile stdout-redirect pattern

**File:** `examples/identifier-similarity/main_test.go:59-109`
**Issue:** The test redirects `os.Stdout` to a pipe writer, calls
`main()`, then restores `os.Stdout` via `defer`. Two robustness gaps:

1. If `os.Pipe()` fails (rare but possible on resource exhaustion),
   the `defer` to restore `os.Stdout` does NOT fire because the
   redirection block panics out before `defer` is set up. The
   `t.Fatalf` at line 67 fires first and exits the test, but in a
   parallel-test environment another test running concurrently could
   write to the now-stale `os.Stdout` pointer.
2. `w.Close()` (line 76) is unchecked. On rare I/O errors the close
   fails silently; if the writer side of the pipe has a problem,
   the reader would never see EOF and `io.Copy` could block.

Neither failure mode is likely on `go test` happy paths, but the
pattern is fragile.

**Fix:** Wrap the stdout-redirect in a helper that captures stdout
to a `bytes.Buffer` via the same pattern but with both the pipe
creation and the `defer` restore guarded under a single `t.Helper()`
wrapper. Alternatively, use `os.CreateTemp` for the capture target
— more robust than a pipe for this use case (no buffer-size limits,
no risk of pipe-blocking semantics). See `testscript` or the
`testify` capture helpers for canonical patterns.

This is low risk (the test passes on every reviewed platform), but
the pattern is worth tightening before the example expands.

---

_Reviewed: 2026-05-14T16:08:26Z_
_Reviewer: Claude (gsd-code-reviewer)_
_Depth: standard_
