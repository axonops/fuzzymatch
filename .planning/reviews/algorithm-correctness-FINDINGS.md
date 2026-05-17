---
status: issues_found
agent: algorithm-correctness-reviewer
scope: entire codebase (phases 1-8)
reviewed: 2026-05-17T08:30:00Z
finding_counts:
  critical: 1
  important: 12
  improvement: 22
  total: 35
---

## Summary

The algorithm catalogue is, on the whole, in excellent shape: every algorithm cites a primary source, formulae are documented in detail in each file's block comment, constants are unexported and traceable to the paper, mathematical invariants are covered by property tests using `testing/quick`, and four cross-validation corpora (SWG, Ratcliff-Obershelp, phonetic, token-ratios) gate regression. The fresh-implementation discipline, including the explicit listing of MIT-licensed Go ports NOT consulted, is exemplary.

The single Critical finding is a public-API contract divergence on `HammingDistance`: the spec (§7.1.4) says `(int, error)` with `ErrHammingLengthMismatch`; the implementation returns `int` and silently returns `max(len(a), len(b))` on unequal-length inputs. Either the spec or the code needs to move. Until that's resolved, the v1.0 freeze on this surface is at risk.

The Important findings are mostly documentation drift between `algoid.go`'s per-AlgoID godoc citations and the implementation files' citations (multiple algorithms cite different authors in the two locations), a couple of missing property invariants the standards document calls "All algorithms" (`NeverPanics` is not exercised as a generic property test for any algorithm — fuzz tests cover the spirit, but the standards skill names this an "All algorithms" invariant), and a few spec-vs-code divergences worth resolving before v1.0.

The Improvements are stylistic and refactoring-level observations — duplicated helpers (`runeAt` in soundex.go reinvents `utf8.DecodeRuneInString`), the dispatch registration pattern that technically performs package-init work via `var _ = func()bool{}()`, missing inner-loop swap optimisations on a couple of paths, and so on.

Phase 9 (scan sub-package) can proceed in parallel with these findings — none of them block the scan layer's design. The Critical Hamming-API finding should be triaged before v1.0 ships.

---

## Critical

### [Critical] Hamming public API diverges from spec §7.1.4
- **File:** /Users/johnny/Development/fuzzymatch/hamming.go:69-90, /Users/johnny/Development/fuzzymatch/errors.go:30-32, /Users/johnny/Development/fuzzymatch/docs/requirements.md:362-364
- **Phase introduced:** Phase 2
- **Issue:** `docs/requirements.md` §7.1.4 specifies the Hamming public surface as:
  - `HammingDistance(a, b string) (int, error)` returning `ErrHammingLengthMismatch` when `len(a) != len(b)`
  - `HammingDistanceRunes(a, b string) (int, error)`
  The actual implementation in `hamming.go` declares `HammingDistance(a, b string) int` (no error return) and silently returns `max(len(a), len(b))` on unequal-length inputs. `errors.go:31` even calls out that `ErrHammingLengthMismatch` would "land alongside the features that introduce them in later phases" — i.e. it was deferred, but no follow-up has reconciled the spec. This is a public-API contract divergence; either the spec must be amended (and the `errors.go` comment removed) or the implementation must add the error return before v1.0 freezes the surface.
- **Standard:** `docs/requirements.md` §7.1.4 (authoritative); `.claude/skills/algorithm-correctness-standards/SKILL.md` §"Edge cases" — "Length mismatch for Hamming: return 0.0 from Score; return `ErrHammingLengthMismatch` from Distance".
- **Action:** Discuss-phase needed (api-ergonomics-reviewer should weigh in — the silent-zero convention has a usability argument; the explicit-error convention has a correctness argument). Whatever the resolution, both the spec and the code must say the same thing.
- **Rationale:** Once v1.0 ships, breaking this surface needs a v2.0 bump. Resolving before v1.0 costs nothing; resolving after is expensive.
- **Suggested fix:** Either (a) amend `docs/requirements.md` §7.1.4 to match the silent-zero implementation and delete the `errors.go:31` comment about `ErrHammingLengthMismatch`, or (b) re-implement `HammingDistance` and `HammingDistanceRunes` to return `(int, error)`, add `ErrHammingLengthMismatch` to `errors.go`, and update all callers (scorer, tests, BDD scenarios).

---

## Important

### [Important] algoid.go godoc citations contradict implementation-file citations
- **File:** /Users/johnny/Development/fuzzymatch/algoid.go:67-71, 90-95, 104-107
- **Phase introduced:** Phase 1
- **Issue:** Three AlgoID godoc comments cite different primary sources than the corresponding implementation file:
  - `AlgoDamerauLevenshteinFull` (algoid.go:67-71) cites "Damerau 1964 — A technique for computer detection and correction of spelling errors." `damerau_full.go:18-19` cites "Lowrance, R., Wagner, R. A. (1975). An extension of the string-to-string correction problem." The actual algorithm IS Lowrance-Wagner 1975 (the implementation file is correct); the algoid.go godoc is wrong.
  - `AlgoStrcmp95` (algoid.go:90-95) cites "Winkler & Thibaudeau 1991 — An application of the Fellegi-Sunter model of record linkage to the 1990 U.S. decennial census." `strcmp95.go:18-19` cites "Winkler, W. E. (1994). Advanced methods for record linkage." Both are valid Winkler papers but they aren't the same paper — strcmp95.go is correct (the algorithm IS in Winkler 1994 §3).
  - `AlgoLCSStr` (algoid.go:104-107) cites "Hunt & Szymanski 1977 — A fast algorithm for computing longest common subsequences (substring variant)." `lcsstr.go:18-19` cites "Wagner, R. A., & Fischer, M. J. (1974). The string-to-string correction problem." Hunt-Szymanski is the LCS-subsequence algorithm; LCSStr is the longest common SUBSTRING (different problem). The implementation file is correct.
- **Standard:** `.claude/skills/algorithm-correctness-standards/SKILL.md` §"Primary Source Citation" — "The citation matches docs/requirements.md §7 entry exactly."
- **Action:** Code fix (godoc-only — no semantics change).
- **Rationale:** Citation accuracy is the load-bearing audit trail for algorithm-licensing and correctness reviews. Drift between two locations confuses downstream consumers (godoc renders, README, llms.txt) and weakens the citation contract.
- **Suggested fix:** Update the three AlgoID godoc comments in algoid.go to match the implementation files (Lowrance-Wagner 1975, Winkler 1994, Wagner & Fischer 1974 respectively).

### [Important] No generic NeverPanics property test for any algorithm
- **File:** /Users/johnny/Development/fuzzymatch/props_test.go (entire file — pattern absence)
- **Phase introduced:** Phase 2 (pattern set by first algorithm)
- **Issue:** The standards skill `algorithm-correctness-standards/SKILL.md` §"Mathematical Invariants" lists "Never panics" as one of three "All algorithms" invariants: "the function does not panic on arbitrary inputs, including invalid UTF-8, embedded NULs, lone surrogates, and very long strings." `props_test.go` has 219 `TestProp_*` tests but NONE of them explicitly assert no-panic via `testing/quick` — the spirit is covered by the 25 fuzz tests (each starts with "Never panics (implicit — any panic propagates as a fuzz crash)") but the standards skill specifically calls out "verified by property tests using testing/quick." The fuzz coverage is genuinely good, but the standards skill names this invariant for `testing/quick` and the test file should comply or the standards skill should be amended.
- **Standard:** `.claude/skills/algorithm-correctness-standards/SKILL.md` §"All algorithms" — "Never panics: the function does not panic on arbitrary inputs."
- **Action:** Either (a) add `TestProp_<Algo>Score_NeverPanics` for each algorithm using `testing/quick` with a `defer recover()` wrapper, or (b) amend `algorithm-correctness-standards/SKILL.md` to acknowledge that "Never panics" is covered by the per-algorithm `Fuzz<Algo>Score` native-fuzz target instead of `testing/quick`.
- **Rationale:** The fuzz approach is arguably stronger (coverage-guided), but the standards skill names `testing/quick` explicitly. Either align the tests or align the skill.
- **Suggested fix:** Amend the skill — fuzz is the right coverage shape for this invariant; documentation drift is the cheaper fix.

### [Important] LCSStr inner-loop swap is missing — stack fast path can miss
- **File:** /Users/johnny/Development/fuzzymatch/lcsstr.go:127-155, 311-319
- **Phase introduced:** Phase 2
- **Issue:** `LongestCommonSubstring` deliberately does NOT swap `a` and `b` when `a` is longer, because the leftmost-in-`a` tie-break would change identity if `a` and `b` were swapped (the implementation correctly documents this at lines 134-141). However, `lcsstrLengthOnly` (called by `LCSStrScore`) also does not swap, even though `LCSStrScore` does not need the leftmost-in-`a` tie-break — only the LENGTH. The ASCII fast-path gate at line 312 fires `if lb <= maxStackInputLen`; if `lb > maxStackInputLen` but `la <= maxStackInputLen`, the stack path is skipped despite being safe. The fix is a 4-line swap in `lcsstrLengthOnly`.
- **Standard:** `.claude/skills/performance-standards/SKILL.md` — ASCII fast-path zero-alloc budget; this is a missed-optimisation against that target.
- **Action:** Code fix (performance only — semantics unchanged because length is symmetric).
- **Rationale:** Modest 0-alloc-on-short-input gain; matters for benchmarks but not correctness.
- **Suggested fix:** In `lcsstrLengthOnly`, prepend `if la < lb { a, b = b, a; la, lb = lb, la }` before the stack-gate; the length is symmetric, so the swap is value-preserving.

### [Important] Spec says HammingDistanceRunes returns (int, error); implementation returns int
- **File:** /Users/johnny/Development/fuzzymatch/hamming.go:105-125, /Users/johnny/Development/fuzzymatch/docs/requirements.md:363
- **Phase introduced:** Phase 2
- **Issue:** Same root cause as the Critical Hamming finding above. Listed separately because the rune variant has its own surface and the rune-path test (`TestHamming_DistanceRunes_UnequalLength`) explicitly pins the silent-`max` behaviour. Whatever resolution the Critical finding settles on, both byte and rune variants must align with the spec.
- **Standard:** spec §7.1.4
- **Action:** Discuss-phase (linked to Critical Hamming finding)
- **Rationale:** Public API contract consistency.

### [Important] dispatch_*.go files use init-time side effects via `var _ = func() bool {...}()`
- **File:** /Users/johnny/Development/fuzzymatch/dispatch_*.go (all 23 files)
- **Phase introduced:** Phase 2 (pattern set by first dispatch wire-up)
- **Issue:** Every dispatch registration file uses the idiom `var _ = func() bool { dispatch[AlgoX] = XScore; return true }()` and the file header comments call this "the canonical way to run package-level side effects without init()." This is a TECHNICALLY correct read of "no init() function" but the `var _ = func()...()` form executes during package initialisation (it's a package-level var initializer expression) — Go's spec treats it identically to an `init()` function for ordering and side-effect semantics. The standards skill says: "The library has no init() functions doing non-trivial work. Tables that require initialisation ... are declared via `var x = ...` literal expressions, not built in init()." A function-pointer assignment is borderline "trivial" — it's a single map write — but the pattern is the same package-initialisation surface that the standard tries to prohibit. Reasonable people can disagree; the project explicitly calls this pattern out so it is at least known and intentional.
- **Standard:** `.claude/skills/determinism-standards/SKILL.md` §13.5 (cited line 155 of the determinism skill).
- **Action:** Skill clarification (the implementation is fine; the standards skill should explicitly bless the var-`_`-func pattern OR the project should refactor to a different registration approach).
- **Rationale:** Future contributors will read the standards skill, see "no init()", and either (a) propose a refactor that breaks the pattern or (b) ask whether the existing pattern is permitted. A two-sentence clarification in the skill saves that round-trip.
- **Suggested fix:** Add to determinism-standards SKILL.md §13.5: "Package-level `var _ = func() bool { dispatch[X] = Y; return true }()` registrations are permitted for the dispatch table and equivalent function-pointer writes; the prohibition is on table BUILDS that depend on runtime computation, not on simple map writes."

### [Important] Cosine FMA-fusion risk documented but no remediation gate
- **File:** /Users/johnny/Development/fuzzymatch/cosine.go:288-297, 341-344
- **Phase introduced:** Phase 5
- **Issue:** The file documents the FMA-fusion risk on arm64 (Go issue #17895): "Go 1.26 may emit FMA on arm64 for the (x*y)+z pattern; parentheses do NOT defeat FMA fusion." The remediation pattern (explicit float64 cast: `dot = float64(float64(qa[k]) * float64(qb[k])) + dot`) is documented but NOT applied. The cross-platform CI matrix is the load-bearing detector. This is a deliberate trade-off (the integer-derived values are small enough that FMA divergence falls below the byte-diff threshold of algorithms.json), but if a future input triggers divergence, the remediation is buried in a file comment instead of being a one-liner test that fires before CI ever sees it.
- **Standard:** `.claude/skills/determinism-standards/SKILL.md` §13.3 (no FMA in algorithm hot paths).
- **Action:** Code fix (preventive) or skill clarification (acknowledge the trade-off).
- **Rationale:** A determinism regression caught in CI matrix output is harder to triage than a determinism regression caught by a unit test that explicitly exercises the FMA-prone input pattern. Either the remediation should be applied pre-emptively, or there should be a dedicated test that asserts byte-identical output between the explicit-cast form and the project's current form to ensure they stay equivalent.
- **Suggested fix:** Apply the explicit-cast pattern now; it costs nothing on amd64 and removes the FMA risk on arm64 forever.

### [Important] Double Metaphone implementation is the most complex algorithm; cross-validation corpus exists but reviewer cannot verify the rule transcription end-to-end
- **File:** /Users/johnny/Development/fuzzymatch/double_metaphone.go (entire file, 916 lines)
- **Phase introduced:** Phase 7
- **Issue:** The Double Metaphone implementation is 916 lines of position-by-position state-machine rules with look-behind / look-ahead of up to 4 positions. The file header explicitly states "no code copied" from the SWI-Prolog C reference, the oubiwann/metaphone Python BSD-3 port, or the four MIT-licensed Go ports listed. The cross-validation corpus (testdata/cross-validation/phonetic/vectors.json) is the load-bearing correctness gate. From this review's vantage, the algorithm-correctness-reviewer cannot verify the 916-line state machine against Philips 2000 paragraph-by-paragraph (the original CUJ paper is paywalled / archive-only; the file cites a SWI-Prolog mirror of the public-domain C code). Several rule branches contain hand-noted bug-comments ("Initial SCH + consonant (not W) — Germanic names like Schmidt. Primary is X (sh-sound); secondary is S (Germanic hard SCH)"). Each of these is a place where a transcription error would silently produce a different key but still pass range-bounds / identity property tests. The defence-in-depth here is the cross-validation corpus alone — and the corpus is generated from `oubiwann/metaphone==0.6` which has its own bugs (the implementation's own godoc cites "Maurice Aubrey's perl-port-derived bug fixes" as fixes the Python port may or may not have absorbed).
- **Standard:** `.claude/skills/algorithm-correctness-standards/SKILL.md` §"Approval Gate" — "literature reference vectors in unit tests."
- **Action:** Discuss-phase needed.
- **Rationale:** Not a defect; a structural concern. The current cross-validation against oubiwann/metaphone may itself be unsound if oubiwann/metaphone has bugs Philips 2000 does not. Recommend either: (a) expand cross-validation corpus to include `jellyfish` (the BSD-2-Clause Rust port used elsewhere for NYSIIS/MRA/Soundex cross-validation) for diversity, or (b) hand-derive 5-10 additional reference vectors directly from the CUJ paper's worked examples and pin them in `double_metaphone_test.go` as paper-anchored rather than tooling-anchored fixtures.
- **Suggested fix:** Add a small `TestDoubleMetaphone_PaperWorkedExamples` table populated from Philips 2000 directly (the paper publishes ~10 worked examples — they are reproducible from the SWI-Prolog mirror's reference comments).

### [Important] runeAt helper in soundex.go reinvents stdlib utf8.DecodeRuneInString
- **File:** /Users/johnny/Development/fuzzymatch/soundex.go:275-300 (and called from double_metaphone.go, nysiis.go, mra.go)
- **Phase introduced:** Phase 7
- **Issue:** `runeAt` is a hand-rolled UTF-8 decoder that the godoc itself describes as "logic is identical to utf8.DecodeRuneInString for the purpose of skip-counting." The rationale given ("avoids an extra import in a small file") is weak — the four phonetic files all import this helper from soundex.go anyway; importing `unicode/utf8` once across all four would be cleaner. The hand-rolled decoder has a subtle correctness gap: it returns `0xFFFD, 1` for continuation bytes out of context but does NOT validate the continuation bytes' high bits (an invalid UTF-8 sequence like `0xC3 0x41` is decoded as `(0xC1, 2)` instead of being rejected). `utf8.DecodeRuneInString` handles this correctly per the documented invariants.
- **Standard:** `.claude/skills/algorithm-correctness-standards/SKILL.md` §"Unicode Handling" — "never panic on invalid UTF-8" is satisfied, but byte-validation correctness matters for the fuzz-test landscape.
- **Action:** Code fix (replace `runeAt` with `utf8.DecodeRuneInString`).
- **Rationale:** Lower code surface, fewer bug-vectors, matches stdlib semantics.
- **Suggested fix:** Delete `runeAt` from soundex.go; replace callsites with `utf8.DecodeRuneInString(s[i:])`. The import is one line per file (or one shared util).

### [Important] NYSIIS silently truncates inputs over 128 ASCII letters
- **File:** /Users/johnny/Development/fuzzymatch/nysiis.go:122-126, 143-147
- **Phase introduced:** Phase 7
- **Issue:** `NYSIISCode` allocates `var nameBuf [128]byte` for the ASCII-letter scan. Inputs with more than 128 ASCII letters are silently truncated at `nLen < 128` gate (lines 143-147). For typical name inputs this is fine, but for adversarial inputs (e.g. a malformed CSV field with 500 letters) the output is the NYSIIS of the first 128 letters, indistinguishable from a "well-formed" input that happens to share that prefix. The truncation is silent: no log, no warning, no error. The MRA implementation has the same pattern at 64 letters.
- **Standard:** `.claude/skills/algorithm-correctness-standards/SKILL.md` §"Edge cases" — "very long strings" are listed as inputs the function must not panic on. Silent truncation is not panicking, but it IS producing a different output than the algorithm specifies for those inputs.
- **Action:** Code fix or doc-only fix.
- **Rationale:** Phonetic algorithms are typically applied to names (max ~50 letters realistic); 128 is generous. But the silent-truncation behaviour is surprising and undocumented.
- **Suggested fix:** Either (a) grow the buffer to `make([]byte, len(s))` when input exceeds 128 (cheap heap alloc on pathological inputs only), or (b) document the truncation explicitly in `NYSIISCode` godoc.

### [Important] Tversky direct-call panic message uses lowercase "tversky" — inconsistent with rest of file
- **File:** /Users/johnny/Development/fuzzymatch/tversky.go:242, 284
- **Phase introduced:** Phase 5
- **Issue:** The panic message is `"fuzzymatch: invalid tversky parameter"` (lowercase tversky). The error sentinel name (`ErrInvalidTverskyParam` in errors.go:88) and the message text in `errors.go:88` are also lowercase (`"fuzzymatch: invalid tversky parameter"`) — so that's consistent, but the file-level godoc and the algorithm name throughout refer to "Tversky" (capitalised, eponymous). Per Go convention error messages should be lowercase. The lowercase form is fine and matches Go style, but it's at odds with the `Tversky` casing elsewhere. Lowercase wins per Go style; flag is a callout that the convention is intentional (verified against `errors.go` per-skill).
- **Standard:** Go convention: error messages are lowercase. No standards violation.
- **Action:** No fix required (just a callout for the reviewer's record).

### [Important] Token Set Ratio's RapidFuzz #110 deviation is documented but the catalogue's both-empty-→-1.0 convention is broken silently for one algorithm
- **File:** /Users/johnny/Development/fuzzymatch/token_set_ratio.go:281-292, 87-107
- **Phase introduced:** Phase 6
- **Issue:** TokenSetRatio is the SOLE algorithm in the catalogue that returns 0.0 (not 1.0) for both-empty input. The file godoc documents this extensively as RapidFuzz issue #110 / fuzzywuzzy parity, and the function comment includes a load-bearing explanation. However, this means the "all algorithms return 1.0 for both-empty" rule from `algorithm-correctness-standards/SKILL.md` §"Edge cases" is broken for one algorithm. The standards skill should explicitly document the deviation OR the algorithm should be brought into line. The current state is "documented deviation" which is fine but the standards skill itself does not call out that exception.
- **Standard:** `.claude/skills/algorithm-correctness-standards/SKILL.md` §"Edge cases" — "Both inputs empty: return 1.0 by convention."
- **Action:** Skill clarification.
- **Rationale:** Future reviewers reading the standards skill will see "Both-empty: 1.0" and flag TokenSetRatio as non-compliant; we need a one-line note in the skill.
- **Suggested fix:** Add to algorithm-correctness-standards SKILL.md §"Edge cases": "Exception: TokenSetRatio returns 0.0 for both-empty input per RapidFuzz issue #110 / fuzzywuzzy parity. The deviation is documented in token_set_ratio.go and is the single catalogue-wide exception to the 1.0 convention."

### [Important] Spec describes a unified phonetic-algorithm score normalisation rule that the implementations diverge from
- **File:** /Users/johnny/Development/fuzzymatch/soundex.go:257-269, /Users/johnny/Development/fuzzymatch/double_metaphone.go:890-915, /Users/johnny/Development/fuzzymatch/nysiis.go:350-362, /Users/johnny/Development/fuzzymatch/mra.go:348-358
- **Phase introduced:** Phase 7
- **Issue:** `.claude/skills/algorithm-correctness-standards/SKILL.md` §"Score Normalisation" says: "Phonetic algorithms: 1.0 if encoded keys match (per algorithm's matching rule), else 0.0." This is broadly correct, but each phonetic algorithm implements a different "matching rule":
  - Soundex: exact code equality (both must be non-empty).
  - Double Metaphone: 4-way match (primary_a == primary_b OR primary_a == secondary_b OR secondary_a == primary_b OR secondary_a == secondary_b). Each match must be non-empty.
  - NYSIIS: exact code equality (both must be non-empty).
  - MRA: the binary form of MRACompare — which uses the NBS 6-counter similarity threshold; codes don't need to match exactly.
  The "matching rule" varies dramatically; the spec/skill should either summarise the four rules explicitly or refer per-algorithm to `docs/requirements.md` §7.4.X. Right now a reviewer reading only the skill could mistakenly conclude that MRAScore uses code-equality (it doesn't).
- **Standard:** `.claude/skills/algorithm-correctness-standards/SKILL.md` §"Score Normalisation".
- **Action:** Skill clarification.
- **Rationale:** Documentation accuracy.
- **Suggested fix:** Expand the phonetic bullet in the SKILL: "Phonetic algorithms: 1.0 if the algorithm's per-algorithm matching rule fires. Specifically: Soundex/NYSIIS use exact code equality; Double Metaphone uses 4-way primary/secondary key matching; MRA uses the NBS-943 6-counter similarity threshold. See docs/requirements.md §7.4.x for the precise rule per algorithm."

---

## Improvement

### [Improvement] dispatch table is populated 23 times via 23 separate dispatch_*.go files — could be one
- **File:** /Users/johnny/Development/fuzzymatch/dispatch_*.go (23 files)
- **Phase introduced:** Phase 2 (pattern)
- **Issue:** 23 single-purpose files, each containing ~30 lines of boilerplate (copyright header, file-header comment, single `var _` line) to perform `dispatch[AlgoX] = XScore`. A single `dispatch.go` file with all 23 registrations would be 50 lines vs the current ~700 lines of boilerplate. The decomposition was deliberate (each phase's implementation plan added its own dispatch wire-up file alongside the algorithm file), but the result is a lot of code surface for a tiny amount of behaviour.
- **Standard:** `.claude/skills/go-coding-standards/SKILL.md` (no specific rule violated).
- **Action:** Code refactor.
- **Rationale:** Maintainability, file-count reduction, easier to verify "all 23 are registered" in code review.
- **Suggested fix:** Consolidate into `dispatch.go` with a single `var _ = func() bool { ... }()` block that performs all 23 registrations.

### [Improvement] Levenshtein and Damerau OSA share identical inner-loop swap and isASCII gate code
- **File:** /Users/johnny/Development/fuzzymatch/levenshtein.go:84-120, /Users/johnny/Development/fuzzymatch/damerau_osa.go:96-129
- **Phase introduced:** Phase 2
- **Issue:** The ASCII-fast-path gate, stack-buffer allocation pattern, and inner-loop-swap pattern are duplicated near-verbatim between `LevenshteinDistance` and `DamerauLevenshteinOSADistance` (and again, with adjustments, in `LongestCommonSubstring` and `lcsLen`). Extracting the gate-and-allocate pattern into a helper would reduce duplication and make the "ASCII fast path policy" testable as one unit.
- **Standard:** None — code quality observation.
- **Action:** Code refactor (optional).
- **Rationale:** Code DRY.

### [Improvement] LCSStrScore allocates two []rune slices for rune-path even when both inputs are ASCII
- **File:** /Users/johnny/Development/fuzzymatch/lcsstr.go:229-244
- **Phase introduced:** Phase 2
- **Issue:** `LCSStrScoreRunes` always allocates `ra := []rune(a)` and `rb := []rune(b)`. If both inputs are ASCII, the byte path would produce the same result with zero allocations. The current implementation doesn't gate on `isASCII(a) && isASCII(b)` to skip the rune conversion. This is a missed optimisation (not a correctness issue).
- **Standard:** `.claude/skills/performance-standards/SKILL.md` — ASCII fast-path budget.
- **Action:** Code refactor.
- **Rationale:** Mirror the pattern from `LevenshteinScoreRunes` (which also lacks this optimisation, so this is a catalogue-wide observation, not LCSStr-specific).

### [Improvement] Test files use ad-hoc absFloat64 helper instead of math.Abs
- **File:** /Users/johnny/Development/fuzzymatch/levenshtein_test.go:37-42 (and likely elsewhere)
- **Phase introduced:** Phase 2
- **Issue:** `levenshtein_test.go` defines `absFloat64` to avoid importing `math` "just for math.Abs (though we use math.Abs below for consistency since math is already imported for quick.Check helpers)." The comment itself notes the inconsistency. Either commit to `math.Abs` and remove `absFloat64`, or commit to the helper and use it everywhere. The current half-and-half is a small style inconsistency.
- **Standard:** None — code style.
- **Action:** Code refactor.
- **Rationale:** Style consistency.

### [Improvement] Soundex's runeAt has subtle continuation-byte validation gap (mentioned above)
- **File:** /Users/johnny/Development/fuzzymatch/soundex.go:275-300
- **Phase introduced:** Phase 7
- **Issue:** See Important finding "runeAt helper in soundex.go reinvents stdlib utf8.DecodeRuneInString" — listed here separately to flag this is also a generic improvement opportunity beyond the validation gap (it removes duplication too).

### [Improvement] Jaro algorithm uses fixed-size stack buffer `[maxJaroStackLen]bool` = 256 even on shorter inputs
- **File:** /Users/johnny/Development/fuzzymatch/jaro.go:152-153
- **Phase introduced:** Phase 2
- **Issue:** `JaroScore` always allocates `var matchA [maxJaroStackLen]bool` (256 bytes) and `var matchB [maxJaroStackLen]bool` (another 256 bytes) on the stack regardless of input length. For a 6-character input (`MARTHA` vs `MARHTA`), 250 bytes of stack space per array is wasted. Not a correctness issue and unlikely to be a measurable performance issue, but the gate `la <= maxJaroStackLen && lb <= maxJaroStackLen` could be tightened to `la <= 32 && lb <= 32` with a `[32]bool` buffer that handles 99% of identifier-style inputs at 1/8th the stack pressure.
- **Standard:** Performance-standards (allocation budgets).
- **Action:** Code refactor (performance).
- **Rationale:** Minor stack-pressure reduction.

### [Improvement] Cosine file comment cites FMA-fusion risk but Levenshtein/Jaro do not warn about it
- **File:** /Users/johnny/Development/fuzzymatch/cosine.go vs /Users/johnny/Development/fuzzymatch/levenshtein.go, jaro.go
- **Phase introduced:** Phase 5
- **Issue:** Cosine extensively documents the FMA-fusion risk (Go issue #17895) and the remediation pattern. Levenshtein and Jaro also use `(x*y) + z` patterns in their float reductions (Levenshtein's `1.0 - float64(dist)/float64(maxLen)` is safe, but Jaro's `(fm/float64(la) + fm/float64(lb) + float64(m-t)/fm) / 3.0` is float-determinism-sensitive). Yet only Cosine flags the risk explicitly. Either the risk is real and the other algorithms should flag it, or the risk is Cosine-specific (because Cosine's `dot` reduction has many more terms in a typical input) and the comment placement is correct.
- **Standard:** Determinism-standards.
- **Action:** Skill clarification or code-comment alignment.
- **Rationale:** Documentation consistency.

### [Improvement] Strcmp95 similar-character table linear-scan over 36 entries could be a 256-entry lookup table
- **File:** /Users/johnny/Development/fuzzymatch/strcmp95.go:198-206
- **Phase introduced:** Phase 2
- **Issue:** `strcmp95SimilarLookup` iterates all 36 entries every call, both byte orientations checked per entry — so 72 comparisons per call. A precomputed 256×256 lookup table would be `2^16 = 65536` bytes (or 256 entries of `[256]float64` = 524288 bytes) which is too large; but a 256-entry `[256]uint64` bitmask (one bit per byte, similarity-credit looked up from a small per-bit mapping) would be 2KB and cut the lookup to O(1). For 36 pairs this is over-engineering; for catalogue uniformity it's worth a thought.
- **Standard:** Performance-standards.
- **Action:** Discuss (likely defer — 36 entries is small).
- **Rationale:** Real performance gain is marginal; documenting the trade-off would close the loop.

### [Improvement] Smith-Waterman-Gotoh uses 6 heap allocations on the rune path even for short inputs
- **File:** /Users/johnny/Development/fuzzymatch/swg.go:451-467
- **Phase introduced:** Phase 3
- **Issue:** `smithWatermanGotohRawRunes` unconditionally heap-allocates 6 `[]float64` rolling rows. For short ASCII inputs going through the Rune* surface (uncommon but possible), this is 6+2 = 8 allocations vs 0 on the byte path. The pattern in `LevenshteinDistanceRunes` is similar: 2 []rune + 2 []int = 4 allocations.
- **Standard:** Performance-standards.
- **Action:** Code fix (stack fast path for rune variant) or doc-only (document the rune-path 8-alloc baseline).
- **Rationale:** Performance.

### [Improvement] Strcmp95 has no Strcmp95ScoreRunes — documented as a deliberate decision
- **File:** /Users/johnny/Development/fuzzymatch/strcmp95.go:73-80
- **Phase introduced:** Phase 2
- **Issue:** Strcmp95 is intentionally ASCII-only — the similar-character table is letter-pair-keyed and has no Unicode equivalent. The file documents this and CONTEXT.md §2 locks the surface. Fine. But this means consumers who pass non-ASCII input get a Strcmp95Score that silently ignores the non-ASCII portion (the byte path includes those bytes in the Jaro match-flag arrays but the similar-character credit cannot fire for them). A consumer wanting Unicode-aware Strcmp95 must pre-normalise. The godoc says this; could be more prominent.
- **Standard:** Documentation.
- **Action:** Doc-only enhancement (godoc).
- **Rationale:** Help users pick the right algorithm.

### [Improvement] Tokeniser-divergence note is repeated verbatim in 4+ algorithm files
- **File:** /Users/johnny/Development/fuzzymatch/token_sort_ratio.go:59-79, /Users/johnny/Development/fuzzymatch/token_set_ratio.go:121-134, /Users/johnny/Development/fuzzymatch/token_jaccard.go:108-117, /Users/johnny/Development/fuzzymatch/monge_elkan.go:107-112
- **Phase introduced:** Phase 6
- **Issue:** The 15-line "OQ-1 RESOLUTION" note about Tokenise being identifier-aware vs RapidFuzz's whitespace-split is duplicated near-verbatim across multiple token-tier algorithm files. Future maintenance: changing the tokenise behaviour requires touching N copies of the note. A shared comment block in tokenise.go that other files reference by single line ("see tokenise.go §X for the tokeniser-divergence rationale") would scale better.
- **Standard:** Documentation DRY.
- **Action:** Code refactor.
- **Rationale:** Documentation consolidation.

### [Improvement] Several algorithms include `_ = opts` in their bodies as a forward-compat placeholder
- **File:** /Users/johnny/Development/fuzzymatch/monge_elkan.go:393 (and likely Scorer-side options)
- **Phase introduced:** Phase 6
- **Issue:** `MongeElkanScore` accepts a `NormalisationOptions` parameter but doesn't use it — the file comment explains this is "for forward-compatibility with the Phase 8 Scorer option." The `_ = opts` line is a clarity comment but it also reads as a smell — the parameter is genuinely unused. Either commit to using opts (apply Normalise internally) or remove it from the signature. The forward-compat argument is weak: Scorer can always wrap MongeElkanScore with its own pre-normalisation.
- **Standard:** Go API ergonomics.
- **Action:** Discuss-phase (api-ergonomics-reviewer should weigh in).
- **Rationale:** API hygiene.

### [Improvement] Partial Ratio TODO comment lacks a GitHub issue reference
- **File:** /Users/johnny/Development/fuzzymatch/partial_ratio.go:146-154
- **Phase introduced:** Phase 6
- **Issue:** `TODO(#TBD): implement sliding-window DP per Bachmann RapidFuzz docs`. The TODO has a placeholder `#TBD` instead of a real issue number. CLAUDE.md §"Workflow — Agent Gates" says "Every TODO must reference a GitHub issue."
- **Standard:** CLAUDE.md (workflow rules).
- **Action:** Create GitHub issue and update TODO.
- **Rationale:** Workflow compliance.

### [Improvement] Damerau-Levenshtein Full uses a full O(m·n) DP table — two-row optimisation deferred to v1.x
- **File:** /Users/johnny/Development/fuzzymatch/damerau_full.go:60-67, 86-87
- **Phase introduced:** Phase 2
- **Issue:** The full DP table is `(m+2) × (n+2)` ints, all heap-allocated. The file comment documents this as a "v1.x performance follow-up." This is a known limitation, not a defect.
- **Standard:** Performance-standards.
- **Action:** Track as a v1.x GitHub issue.
- **Rationale:** Performance debt.

### [Improvement] Cosine clamp at line 385-390 silently returns 1.0 for values slightly above 1.0
- **File:** /Users/johnny/Development/fuzzymatch/cosine.go:385-390
- **Phase introduced:** Phase 5
- **Issue:** The clamp `if cos > 1.0 { return 1.0 }` is correct (IEEE-754 rounding can produce 1.0000000000000002 in degenerate Cauchy-Schwarz cases) but silently. A consumer surfacing the raw Cosine value (e.g. via the Scorer's per-algorithm breakdown) would see exactly 1.0 in cases where the algebraic limit is 1.0. This is the correct behaviour — but if a future bug pushes the value WELL above 1.0 (say 1.5), the clamp would silently hide the bug. A test asserting "the clamp fires only within 1 ULP of 1.0" would catch that scenario.
- **Standard:** Determinism-standards.
- **Action:** Test enhancement.
- **Rationale:** Defensive regression detection.

### [Improvement] Double Metaphone references the SWI-Prolog mirror but the source URL may rot
- **File:** /Users/johnny/Development/fuzzymatch/double_metaphone.go:22-23, 41-42
- **Phase introduced:** Phase 7
- **Issue:** The file cites `https://github.com/SWI-Prolog/packages-nlp/blob/master/double_metaphone.c` as the "stable URL for provenance verification." GitHub URLs to single files on `master` are NOT stable — they shift if the upstream repository reorganises. A pinned commit hash would be more durable.
- **Standard:** Citation hygiene.
- **Action:** Doc-only.
- **Rationale:** Long-term reproducibility of provenance.

### [Improvement] Phonetic algorithms all mention "non-ASCII runes dropped silently" but the spec doesn't pin the behaviour
- **File:** /Users/johnny/Development/fuzzymatch/soundex.go, double_metaphone.go, nysiis.go, mra.go (file headers + per-function godoc)
- **Phase introduced:** Phase 7
- **Issue:** Every phonetic file's godoc mentions that non-ASCII runes are dropped silently. The spec §7.4 doesn't explicitly say "drop silently" vs "panic" vs "return error" — the implementation choice has been made consistently (silent drop) but the spec doesn't pin it. If a consumer reports "I expected é to be normalised to e before phonetic encoding" we'd point at our docs; the spec could be more explicit.
- **Standard:** Spec accuracy.
- **Action:** Spec update.
- **Rationale:** Spec-vs-code alignment.

### [Improvement] Ratcliff-Obershelp recursion depth is documented as O(min(la, lb)) but no explicit cap exists
- **File:** /Users/johnny/Development/fuzzymatch/ratcliff_obershelp.go:81-83, 199-210
- **Phase introduced:** Phase 4
- **Issue:** The file comment says "Recursion depth is O(min(la, lb)) in the worst case." For very long inputs (10⁵ chars) this is genuine stack pressure. Go's default goroutine stack is 8KB and grows dynamically up to 1GB; recursion depth of 10⁵ on small frames is fine, but if a future change adds local-variable state to the recursion (e.g. tracking match positions), the stack could grow significantly. A tail-call-elimination pass or an explicit iterative stack (with `make([]frame, 0, 16)`) would bound the stack pressure.
- **Standard:** Performance-standards / safety.
- **Action:** Code refactor (defer).
- **Rationale:** Future-proofing.

### [Improvement] LCSStrScore vs LongestCommonSubstring (length only vs full substring) inconsistency in stack-buffer use
- **File:** /Users/johnny/Development/fuzzymatch/lcsstr.go:311-319
- **Phase introduced:** Phase 2
- **Issue:** `lcsstrLengthOnly` uses the stack buffer when `lb <= maxStackInputLen` (no swap). `LongestCommonSubstring` uses the same gate. Both could swap for symmetric performance — see Important finding above for LCSStrScore. Listed separately as improvement because the score-only path has no leftmost-tie-break constraint preventing the swap.

### [Improvement] errors.go header comment says ErrHammingLengthMismatch "lands alongside the features that introduce them in later phases" but the feature never lands
- **File:** /Users/johnny/Development/fuzzymatch/errors.go:30-32
- **Phase introduced:** Phase 1
- **Issue:** The errors.go header explicitly promises that `ErrHammingLengthMismatch` will land later. It hasn't. Either delete the promise or fulfil it (the latter is the Critical Hamming finding above).
- **Standard:** Documentation accuracy.
- **Action:** Doc fix (or, linked to Critical, code fix).

### [Improvement] golden_canonical.go / golden_canonical_test.go pin reference vectors but the discoverability of "this is the golden vector for X" is via filename only
- **File:** /Users/johnny/Development/fuzzymatch/golden_canonical.go, /Users/johnny/Development/fuzzymatch/algorithms_golden_test.go
- **Phase introduced:** Phase 2
- **Issue:** `algorithms_golden_test.go` is 55KB; finding the canonical vector for a specific algorithm requires text search. A README in `testdata/golden/` indexing which test covers which algorithm would help future reviewers.
- **Standard:** Documentation discoverability.
- **Action:** Doc enhancement.

---

## Notes for Phase 9 (scan sub-package)

The algorithm catalogue is sound enough to support Phase 9 without rework. The findings above are all addressable in parallel with scan development. The Critical Hamming-API finding is the only item that could affect scan's design (if scan reports unequal-length comparisons specifically), but scan's Phase 9 surface is unlikely to use `HammingDistance`'s error return directly — it composes via the Scorer which uses `HammingScore` (silent-zero). So Phase 9 can proceed.

