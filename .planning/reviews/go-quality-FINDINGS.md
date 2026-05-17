---
status: issues_found
agent: go-quality
scope: full automated quality gate
reviewed: 2026-05-17T06:36:52Z
finding_counts:
  critical: 3
  important: 17
  improvement: 3
  total: 23
---

# Go Quality Gate Findings

All commands run from `/Users/johnny/Development/fuzzymatch` (repo root).
Toolchain: `go 1.26.3 darwin/arm64`, `golangci-lint 2.11.4`.

---

## Critical Findings

### [Critical] golangci-lint run exits non-zero — 27 issues across 7 linters

- **File:** global (multiple files — see Important findings 4–15 for per-issue detail)
- **Phase introduced:** Phases 2, 7 (algorithm files); Phase 8 (scorer)
- **Issue:** `golangci-lint run ./...` exits 1. Full output:
  ```
  double_metaphone.go:716:5: ifElseChain: rewrite if-else to switch statement (gocritic)
  double_metaphone.go:744:4: dupBranchBody: both branches in if statement have same body (gocritic)
  double_metaphone.go:120:1: cyclomatic complexity 12 of func `dmSlgCheck` is high (> 10) (gocyclo)
  double_metaphone.go:176:1: cyclomatic complexity 17 of func `dmPrep` is high (> 10) (gocyclo)
  double_metaphone.go:243:1: cyclomatic complexity 280 of func `DoubleMetaphoneKeys` is high (> 10) (gocyclo)
  double_metaphone.go:890:1: cyclomatic complexity 13 of func `DoubleMetaphoneScore` is high (> 10) (gocyclo)
  mra.go:135:1: cyclomatic complexity 16 of func `MRACode` is high (> 10) (gocyclo)
  mra.go:241:1: cyclomatic complexity 23 of func `MRACompare` is high (> 10) (gocyclo)
  nysiis.go:116:1: cyclomatic complexity 78 of func `NYSIISCode` is high (> 10) (gocyclo)
  soundex.go:147:1: cyclomatic complexity 29 of func `SoundexCode` is high (> 10) (gocyclo)
  double_metaphone_test.go:159:1: File is not properly formatted (gofumpt)
  mra_bench_test.go:30:1: File is not properly formatted (gofumpt)
  nysiis_bench_test.go:28:1: File is not properly formatted (gofumpt)
  mra.go:211:22: G602: slice index out of range (gosec)
  mra.go:212:22: G602: slice index out of range (gosec)
  mra.go:213:22: G602: slice index out of range (gosec)
  scorer_golden_test.go:287:77: `abreviation` is a misspelling of `abbreviation` (misspell)
  scorer_options_internal_test.go:28:11: `artifact` is a misspelling of `artefact` (misspell)
  double_metaphone.go:397:33: QF1001: could apply De Morgan's law (staticcheck)
  double_metaphone.go:825:7: QF1001: could apply De Morgan's law (staticcheck)
  double_metaphone_fuzz_test.go:105:9: QF1001: could apply De Morgan's law (staticcheck)
  double_metaphone_test.go:322:8: QF1001: could apply De Morgan's law (staticcheck)
  errors_test.go:209:5: SA4004: the surrounding loop is unconditionally terminated (staticcheck)
  props_test.go:3492:8: QF1001: could apply De Morgan's law (staticcheck)
  soundex_test.go:203:6: QF1001: could apply De Morgan's law (staticcheck)
  scorer_options_internal_test.go:113:43: probeScoreFnInvoke - i always receives 0 (unparam)
  soundex.go:275:31: runeAt - result 0 (rune) is never used (unparam)
  27 issues
  ```
- **Standard:** `go-coding-standards` SKILL.md §Quality Gate item 3; CLAUDE.md "Checks" item 3
- **Action:** Code fix (see individual findings 4–15 for per-issue remediation)
- **Rationale:** `lint` is a direct dependency of `make check`. CI blocks on this gate. The 27 issues span correctness risk (dupBranchBody), style (gofumpt, De Morgan), cyclomatic complexity suppressions, a gosec false positive, a misspell locale mismatch, and two unparam observations.

---

### [Critical] Coverage floor gate exits non-zero — overall 90.2% below 95.0% floor

- **File:** `scripts/verify-coverage-floors.sh` / `coverage.out`
- **Phase introduced:** all phases (cumulative gap across algorithm implementations)
- **Issue:** `bash scripts/verify-coverage-floors.sh coverage.out` exits 1:
  ```
  verify-coverage-floors: FAIL — overall coverage 90.2% < 95.0%
  ```
  The `make coverage` run (296s, race-enabled) shows `coverage: 90.4% of statements`; the `go tool cover -func` total line reads `90.2%`. The 4.8 percentage-point gap means roughly 1 in 20 statements in the root package is not exercised.
- **Standard:** `go-testing-standards` SKILL.md §Coverage Targets ("≥ 95% overall"); CLAUDE.md Constraints ("Coverage targets: ≥ 95% overall")
- **Action:** Code fix — add test cases to bring the overall floor to ≥ 95%. The most impactful targets are `double_metaphone.go` (heap-path in `dmPrep`, untested phonetic rule branches in `DoubleMetaphoneKeys`), `nysiis.go` (`NYSIISCode` at 71.7%), and `DefaultScorer`'s panic branch.
- **Rationale:** `coverage-check` is a direct dependency of `make check`. The CI gate will block every PR until the floor is met. The `verify-coverage-floors.sh` script exits immediately on the overall-floor failure, so the per-file floor check (Floor 2) has not yet run; per-file violations exist and will surface as a second failure once the overall floor is fixed (see finding 16).

---

### [Critical] go.sum contains three stale entries removed by go mod tidy

- **File:** `/Users/johnny/Development/fuzzymatch/go.sum`
- **Phase introduced:** Phase 1 (module hygiene carried forward)
- **Issue:** `go mod tidy && git diff --exit-code -- go.mod go.sum` exits 1:
  ```diff
  --- a/go.sum
  +++ b/go.sum
  @@ -1,5 +1,2 @@
  -golang.org/x/mod v0.35.0/go.mod h1:+GwiRhIInF8wPm+4AoT6L0FA1QWAad3OMdTRx4tFYlU=
  -golang.org/x/sync v0.20.0/go.mod h1:9xrNwdLfx4jkKbNva9FpL6vEN7evnE43NNNJQ2LF3+0=
   golang.org/x/text v0.37.0 h1:...
   golang.org/x/text v0.37.0/go.mod h1:...
  -golang.org/x/tools v0.44.0/go.mod h1:KA0AfVErSdxRZIsOVipbv3rQhVXTnlU6UhKxHd1seDI=
  ```
  `go mod tidy` removes the three `/go.mod`-only hash entries for `golang.org/x/mod@v0.35.0`, `golang.org/x/sync@v0.20.0`, and `golang.org/x/tools@v0.44.0`. These are stale indirect-module go.mod hashes from a previous toolchain version of `golang.org/x/text`'s own dependency graph that `go 1.26.3` no longer needs to record.
- **Standard:** `go-coding-standards` SKILL.md §Quality Gate item 6; CLAUDE.md Checks item 7 ("go mod tidy produces no diff")
- **Action:** Code fix — run `go mod tidy` and commit the updated `go.sum`. The `go.mod` itself is unchanged.
- **Rationale:** `tidy-check` is a direct dependency of `make check`. The CI gate blocks on this diff. The fix is a single `go mod tidy` commit.

---

## Important Findings

### [Important] gocyclo — 8 phonetic algorithm functions exceed CC 10 with no nolint directive

- **File:** `double_metaphone.go` (lines 120, 176, 243, 890), `mra.go` (lines 135, 241), `nysiis.go` (line 116), `soundex.go` (line 147)
- **Phase introduced:** Phase 7 (`07-phonetic-algorithms`)
- **Issue:** gocyclo reports cyclomatic complexity exceeding the configured threshold of 10 for eight functions:
  - `dmSlgCheck` CC=12 (`double_metaphone.go:120`)
  - `dmPrep` CC=17 (`double_metaphone.go:176`)
  - `DoubleMetaphoneKeys` CC=280 (`double_metaphone.go:243`)
  - `DoubleMetaphoneScore` CC=13 (`double_metaphone.go:890`)
  - `MRACode` CC=16 (`mra.go:135`)
  - `MRACompare` CC=23 (`mra.go:241`)
  - `NYSIISCode` CC=78 (`nysiis.go:116`)
  - `SoundexCode` CC=29 (`soundex.go:147`)

  None of these have a `//nolint:gocyclo` directive with a justification comment. The codebase already uses this pattern with explicit rationale for edit-distance functions (e.g. `damerau_osa.go:218`, `jaro.go:190`, `algoid.go:213`) — the phonetic files are inconsistent with that established convention.
- **Standard:** `go-coding-standards` SKILL.md §Complexity ("Cyclomatic complexity > 10: refactor before merge"); same file notes "Algorithm DP loops are exempt from the function-length guidance where splitting would obscure the recurrence"
- **Action:** Code fix — add `//nolint:gocyclo` directives with justification comments referencing the originating algorithm paper (Philips 2000 for Double Metaphone, NBS Tech Note 943 for NYSIIS/MRA, Knuth 1973 for Soundex). Refactoring the rule tables into separate functions is an alternative for `NYSIISCode` (CC=78) and `DoubleMetaphoneKeys` (CC=280), but only if it does not obscure the algorithmic recurrence.
- **Rationale:** Without nolint directives, golangci-lint exits 1 (see Critical finding 1). The complexity is inherent to these algorithms' phonetic rule sets and cannot be meaningfully reduced without fragmenting the algorithm logic across files in a way that harms readability and academic cross-reference.
- **Suggested fix:** Add to each function declaration, e.g. for `DoubleMetaphoneKeys`: `func DoubleMetaphoneKeys(s string) (primary, secondary string) { //nolint:gocyclo // Philips (2000) 'Double Metaphone' rule engine; inherently complex per-letter decision tree that cannot be split without losing the one-to-one mapping to §III of the paper`

---

### [Important] gocritic dupBranchBody — double_metaphone.go:744 both if/else branches are identical (potential algorithm correctness issue)

- **File:** `/Users/johnny/Development/fuzzymatch/double_metaphone.go`, line 744
- **Phase introduced:** Phase 7 (`07-phonetic-algorithms`)
- **Issue:** `gocritic` reports `dupBranchBody: both branches in if statement have same body`. The offending code:
  ```go
  // French "SAIS" end
  if i == n-1 && (at(i-1) == 'A' || at(i-1) == 'I') {
      dmAdd(&p, &alt, "S", "")
  } else {
      dmAdd(&p, &alt, "S", "")
  }
  ```
  Both the French-end case (`BAIS`, `MAIS` terminal S) and the general S case produce `("S", "")`. In several Double Metaphone reference implementations (including the original Philips C++ and the Python `metaphone` library), the general-S else branch produces `("S", "S")` — both primary AND secondary are "S" — while the French-end case silently drops the secondary to produce `("S", "")`. If the intended behaviour follows this reference, the else branch should read `dmAdd(&p, &alt, "S", "S")` and the current code silently drops the secondary "S" in all non-French S positions.
- **Standard:** `algorithm-correctness-standards` SKILL.md §Primary Source Citation; `go-coding-standards` SKILL.md §API Design
- **Action:** Discuss-phase needed — the `algorithm-correctness-reviewer` must verify the expected behaviour for the general-S case against Philips (2000) §III before the code is changed. If the else branch is confirmed to be `("S","S")`, this is a silent correctness bug affecting all words with S in a non-final-French position.
- **Rationale:** The `dupBranchBody` finding from gocritic surfaces a code structure that is either dead code (the if condition is meaningless because both branches are identical) or a latent algorithm bug. The Double Metaphone phonetic cross-validation test suite (`phonetic_codes_golden_test.go`) may not exercise the full range of secondary-key paths needed to catch this regression.

---

### [Important] gocritic ifElseChain — double_metaphone.go:716 rewrite if-else to switch

- **File:** `/Users/johnny/Development/fuzzymatch/double_metaphone.go`, line 716
- **Phase introduced:** Phase 7 (`07-phonetic-algorithms`)
- **Issue:** `gocritic` reports `ifElseChain: rewrite if-else to switch statement` at line 716. The three-way `if/else if/else` for the SCH trigram case could be expressed as a `switch` for readability.
- **Standard:** `go-coding-standards` SKILL.md §Complexity ("switch with more than ~6 cases: consider dispatch table")
- **Action:** Code fix — rewrite the `if/else if/else` chain as a `switch` or add a `//nolint:gocritic` directive if the if-chain is clearer for this algorithm rule.
- **Rationale:** This contributes to the golangci-lint exit 1 (Critical finding 1). It is a style issue, not a correctness issue.

---

### [Important] gosec G602 — mra.go:211-213 potential slice index out of range (false positive needing nolint)

- **File:** `/Users/johnny/Development/fuzzymatch/mra.go`, lines 211–213
- **Phase introduced:** Phase 7 (`07-phonetic-algorithms`)
- **Issue:** `gosec` and golangci-lint both report G602 (CWE-118) at lines 211, 212, 213:
  ```go
  result[3] = step2Buf[step2Len-3]   // line 211
  result[4] = step2Buf[step2Len-2]   // line 212
  result[5] = step2Buf[step2Len-1]   // line 213
  ```
  These lines are inside the `if step2Len <= 6` branch's else path (i.e. `step2Len > 6`), which guarantees `step2Len >= 7`. With `step2Len >= 7`, the lowest index accessed is `step2Len-3 >= 4`, which is within the `[64]byte step2Buf` bounds. The `result` array is `[6]byte` and indices 3, 4, 5 are valid. gosec cannot statically prove the `step2Len > 6` guard from the conditional structure, so it fires a false positive.
- **Standard:** `go-coding-standards` SKILL.md §Quality Gate; no specific standard for gosec false-positive suppression
- **Action:** Code fix — add `//nolint:gosec // G602 false positive: step2Len > 6 is guaranteed by the enclosing if guard; all index arithmetic is within [64]byte bounds` on lines 211–213 (or restructure to use a `copy` call that gosec can prove safe).
- **Rationale:** The three G602 instances contribute to golangci-lint exit 1. A brief explanatory comment is essential so the suppression is auditable.

---

### [Important] gofumpt — mra_bench_test.go:30 and nysiis_bench_test.go:28 multiple adjacent var declarations should be grouped

- **File:** `/Users/johnny/Development/fuzzymatch/mra_bench_test.go` (line 30), `/Users/johnny/Development/fuzzymatch/nysiis_bench_test.go` (line 28)
- **Phase introduced:** Phase 7 (`07-phonetic-algorithms`)
- **Issue:** `gofumpt` reports formatting issues in both files. In `mra_bench_test.go`, four consecutive `var` declarations (`mraCodeSink`, `mraMatchedSink`, `mraSimSink`, `mraScoreSink`) are each on separate `var` lines rather than grouped in a `var (...)` block. Same pattern in `nysiis_bench_test.go` for `nysiisSink` and `nysiisScoreSink`. gofumpt requires adjacent package-level `var` declarations to be grouped.
- **Standard:** `go-coding-standards` SKILL.md §Quality Gate item 1 ("gofmt -s and goimports produce no diff"); project uses gofumpt (stricter than gofmt)
- **Action:** Code fix — merge adjacent `var` declarations into grouped `var (...)` blocks in both files:
  ```go
  var (
      mraCodeSink    string
      mraMatchedSink bool
      mraSimSink     int
      mraScoreSink   float64
  )
  ```
- **Rationale:** gofumpt issues contribute to golangci-lint exit 1 (Critical finding 1). The fix is mechanical.

---

### [Important] gofumpt — double_metaphone_test.go:159 struct literal formatting

- **File:** `/Users/johnny/Development/fuzzymatch/double_metaphone_test.go`, line 159
- **Phase introduced:** Phase 7 (`07-phonetic-algorithms`)
- **Issue:** `gofumpt` reports the file is not properly formatted at line 159. The struct literal `{"RV-DM1/Schmidt", "Schmidt", "XMT", "SMT", "germanic",` spans multiple lines in a way that gofumpt's stricter multi-line struct-literal rules do not accept. Running `gofumpt -w double_metaphone_test.go` will apply the canonical formatting.
- **Standard:** `go-coding-standards` SKILL.md §Quality Gate item 1
- **Action:** Code fix — run `gofumpt -w double_metaphone_test.go` (or `make fmt`) and review the result.
- **Rationale:** Contributes to golangci-lint exit 1.

---

### [Important] misspell — scorer_options_internal_test.go:28 "artifact" should be "artefact" (UK English locale)

- **File:** `/Users/johnny/Development/fuzzymatch/scorer_options_internal_test.go`, line 28
- **Phase introduced:** Phase 8 (`08-composite-scorer`)
- **Issue:** `misspell` (locale: UK, as configured in `.golangci.yml`) flags the word "artifact" in a comment as a misspelling of "artefact". The exact comment:
  ```go
  // the build-tag _test.go suffix ensures this file never ships in the
  // public artifact.
  ```
  The project's `documentation-standards` SKILL.md §Language explicitly requires British English ("colour, behaviour, organisation, normalisation…"). "Artefact" is the standard British English spelling.
- **Standard:** `documentation-standards` SKILL.md §Language ("British English for prose")
- **Action:** Code fix — change `artifact` to `artefact` in the comment at line 28.
- **Rationale:** Contributes to golangci-lint exit 1. Also enforces the project-wide British English convention.

---

### [Important] misspell — scorer_golden_test.go:287 intentional test-data string flagged as misspelling

- **File:** `/Users/johnny/Development/fuzzymatch/scorer_golden_test.go`, line 287
- **Phase introduced:** Phase 8 (`08-composite-scorer`)
- **Issue:** `misspell` flags `"abreviation"` (single-`b` deliberate typo used as test input) as a misspelling of `"abbreviation"`:
  ```go
  entries = append(entries, makeScorerGoldenEntry(defaultS, "abbreviation", "abreviation", "DefaultScorer"))
  ```
  The string `"abreviation"` is intentional test data: a one-letter-drop variant of "abbreviation" used to exercise the scorer's fuzzy matching near the threshold boundary. The misspell linter scans string literal content.
- **Standard:** `go-coding-standards` SKILL.md §Quality Gate item 3 (lint must be clean)
- **Action:** Code fix — add a `//nolint:misspell` directive on line 287 with a comment explaining the intentional test data: `//nolint:misspell // "abreviation" is an intentional one-letter-drop test vector (scorer threshold boundary)`.
- **Rationale:** Contributes to golangci-lint exit 1. The misspelling is by design; suppression with explanation is the correct approach.

---

### [Important] staticcheck SA4004 — errors_test.go:209 for-range loop with unconditional break inspects only first rune

- **File:** `/Users/johnny/Development/fuzzymatch/errors_test.go`, line 209
- **Phase introduced:** Phase 1 (`01-foundation-infrastructure`)
- **Issue:** `staticcheck SA4004` reports "the surrounding loop is unconditionally terminated". The code:
  ```go
  for _, r := range body {
      if unicode.IsUpper(r) {
          t.Errorf(...)
      }
      // Only inspect the first rune.
      break
  }
  ```
  The `break` is outside the `if` and unconditional, so the loop always exits after the first iteration regardless of the `if` outcome. This is intentional (the comment says "Only inspect the first rune"), but the loop structure is misleading: `strings.HasPrefix` or `utf8.DecodeRuneInString` would express the intent directly without the loop.
- **Standard:** `go-coding-standards` SKILL.md §Complexity (guard clauses and early returns)
- **Action:** Code fix — replace the for-range loop with a direct rune check:
  ```go
  if r, _ := utf8.DecodeRuneInString(body); unicode.IsUpper(r) {
      t.Errorf(...)
  }
  ```
  Or add a `//nolint:staticcheck` directive if the loop form is preferred for readability.
- **Rationale:** Contributes to golangci-lint exit 1.

---

### [Important] staticcheck QF1001 — De Morgan's law applicable in 6 locations

- **File:** `double_metaphone.go` (lines 397, 825), `double_metaphone_fuzz_test.go` (line 105), `double_metaphone_test.go` (line 322), `props_test.go` (line 3492), `soundex_test.go` (line 203)
- **Phase introduced:** Phase 7 (`07-phonetic-algorithms`); Phase 2 for `soundex_test.go`
- **Issue:** `staticcheck QF1001` suggests applying De Morgan's law to simplify six boolean conditions. The recurring pattern `!((c >= 'A' && c <= 'Z') || c == '0')` should read `(c < 'A' || c > 'Z') && c != '0'`. Similarly `!(i == 1 && v[0] == 'M')` → `i != 1 || v[0] != 'M'`.
- **Standard:** `go-coding-standards` SKILL.md §Complexity (clarity, guard clauses)
- **Action:** Code fix — apply De Morgan's transformations at all six sites. Alternatively, extract the alphabet-check into a named helper (`isValidMetaphoneChar`) to eliminate duplication across `double_metaphone.go`, `double_metaphone_fuzz_test.go`, `double_metaphone_test.go`, and `props_test.go`.
- **Rationale:** Contributes to golangci-lint exit 1. Extracting a helper would also reduce the four near-identical condition copies.

---

### [Important] unparam — soundex.go:275 runeAt returns a rune that is never used at any call site

- **File:** `/Users/johnny/Development/fuzzymatch/soundex.go`, line 275
- **Phase introduced:** Phase 2 (`02-core-character-algorithms-six`) or Phase 7 depending on when Soundex was added
- **Issue:** `unparam` reports `runeAt - result 0 (rune) is never used`. The function signature is `func runeAt(s string, i int) (rune, int)` but every call site only uses the second return value (byte width), discarding the rune. Example call: `_, sz := runeAt(s, i)`. The rune value is computed but always discarded, making the function's first return type unused.
- **Standard:** `go-coding-standards` SKILL.md §API Design (functions return what callers need)
- **Action:** Code fix — change the signature to `func runeAt(s string, i int) int` and update the implementation to return only the byte width. This simplifies call sites to `sz := runeAt(s, i)`. If the rune value is needed for future callers, retain the rune return and add `//nolint:unparam` with a forward-looking comment.
- **Rationale:** Contributes to golangci-lint exit 1. The simpler signature is more honest about what callers actually use.

---

### [Important] unparam — scorer_options_internal_test.go:113 probeScoreFnInvoke always receives i=0

- **File:** `/Users/johnny/Development/fuzzymatch/scorer_options_internal_test.go`, line 113
- **Phase introduced:** Phase 8 (`08-composite-scorer`)
- **Issue:** `unparam` reports `probeScoreFnInvoke - i always receives 0`. The helper function `probeScoreFnInvoke(cfg scorerConfig, i int, a, b string)` accepts an index `i` but every call site passes literal `0`. The function could be simplified to always use index 0.
- **Standard:** `go-coding-standards` SKILL.md §API Design (zero value usability)
- **Action:** Code fix — remove the `i int` parameter and hardcode the index, or add `//nolint:unparam` with a comment if the parameter is retained for future test extensibility.
- **Rationale:** Contributes to golangci-lint exit 1.

---

### [Important] Per-file coverage floor — double_metaphone.go at 83.8% (< 90% floor)

- **File:** `/Users/johnny/Development/fuzzymatch/double_metaphone.go`
- **Phase introduced:** Phase 7 (`07-phonetic-algorithms`)
- **Issue:** File-average statement coverage is 83.8%, below the per-file floor of 90.0%. Function-level breakdown:
  - `dmPrep` at 58.8% — the heap fallback path (`heapPath:` label, names > 64 ASCII chars) and non-ASCII rune handling are not exercised
  - `DoubleMetaphoneKeys` at 62.0% — large portions of the 648-line phonetic rule engine are not covered (many letter-case branches have no test vectors)
  - `dmAdd` at 84.6% — the length-truncation branches (when a phoneme exceeds `dmMaxLen`) are not exercised
  - `dmSlgCheck` at 88.9% — the WITZ substring check is not covered
  Note: the `verify-coverage-floors.sh` script's Floor 2 (per-file) check does NOT currently run because the script exits early on the overall-floor failure (Floor 1). This per-file violation will surface as an additional CI failure once the overall floor is fixed.
- **Standard:** `go-testing-standards` SKILL.md §Coverage Targets ("≥ 90% per file"); CLAUDE.md Constraints
- **Action:** Code fix — add test vectors to `double_metaphone_test.go` and/or `double_metaphone_fuzz_test.go` covering: (1) names longer than 64 ASCII letters to trigger the heap path in `dmPrep`, (2) non-ASCII input to `dmPrep`, (3) phonemes that exceed 4 chars to trigger the `dmAdd` truncation branch, (4) a name ending in "WITZ" for `dmSlgCheck`, and (5) additional phonetic rule cases in `DoubleMetaphoneKeys` for the currently uncovered letter sections.
- **Rationale:** The per-file floor is a CI gate (`verify-coverage-floors.sh` Floor 2). Fixing this also directly contributes to closing the overall-coverage gap (Critical finding 2).

---

### [Important] Major exported functions below 100% statement coverage

- **File:** global (root package)
- **Phase introduced:** Phases 2, 6, 7, 8 (by algorithm introduction phase)
- **Issue:** 18 exported public API functions have statement coverage below 100%. Coverage is measured per `go tool cover -func=coverage.out`. The `verify-coverage-floors.sh` Floor 3 check only enforces non-zero coverage (any exercising test exists), not 100% statement coverage. The SKILL targets 100% on public API. The most significant gaps:

  | Function | File | Coverage |
  |---|---|---|
  | `DoubleMetaphoneKeys` | `double_metaphone.go:243` | 62.0% |
  | `NYSIISCode` | `nysiis.go:116` | 71.7% |
  | `DefaultScorer` | `scorer.go:586` | 75.0% |
  | `DamerauLevenshteinOSADistanceRunes` | `damerau_osa.go:137` | 80.0% |
  | `DamerauLevenshteinOSAScoreRunes` | `damerau_osa.go:183` | 81.8% |
  | `MRACompare` | `mra.go:241` | 91.3% |
  | `DoubleMetaphoneScore` | `double_metaphone.go:890` | 92.3% |
  | `NewScorer` | `scorer.go:180` | 92.6% |
  | `TokenJaccardScore` | `token_jaccard.go:200` | 93.3% |
  | `JaroWinklerScoreRunes` | `jarowinkler.go:154` | 93.8% |
  | `Strcmp95Score` | `strcmp95.go:260` | 93.8% |
  | `HammingScoreRunes` | `hamming.go:170` | 94.1% |
  | `PartialRatioScore` | `partial_ratio.go:277` | 94.1% |
  | `PartialRatioScoreRunes` | `partial_ratio.go:444` | 94.1% |
  | `SoundexCode` | `soundex.go:147` | 94.3% |
  | `DamerauLevenshteinFullScoreRunes` | `damerau_full.go:190` | 90.9% |
  | `LevenshteinScoreRunes` | `levenshtein.go:177` | 90.9% |
  | `Tokenise` | `tokenise.go:143` | 95.5% |

  Notable root causes:
  - `DefaultScorer` at 75%: the `panic` branch (bug-guard when `DefaultScorerOptions()` produces an error) is untested — no test passes invalid options to the constructor underlying `DefaultScorer`.
  - `DamerauLevenshteinOSADistanceRunes` / `DamerauLevenshteinOSAScoreRunes` at 80/82%: the early-identity short-circuit and the rune-slice path have branch gaps in the Unicode path.
  - `NYSIISCode` at 71.7%: large sections of the NYSIIS replacement rules are not covered by current test vectors.

- **Standard:** `go-testing-standards` SKILL.md §Coverage Targets ("100% on public API surface")
- **Action:** Code fix — add targeted test cases for each function's uncovered branches. Priority order: `DoubleMetaphoneKeys` (62%), `NYSIISCode` (71.7%), `DefaultScorer` (75%), `DamerauLevenshteinOSADistanceRunes` (80%).
- **Rationale:** The 100% public API target is stated in both the skill and CLAUDE.md Constraints. While the `verify-coverage-floors.sh` Floor 3 check does not enforce statement-level 100% (it only checks that the function is exercised at all — see finding 17), the standard requires 100% and the test-analyst agent gate should catch this on milestone reviews.

---

### [Important] Coverage floor script Floor 3 enforces non-zero coverage, not 100% statement coverage for public API

- **File:** `/Users/johnny/Development/fuzzymatch/scripts/verify-coverage-floors.sh`, line 14–17
- **Phase introduced:** Phase 1 (`01-foundation-infrastructure`)
- **Issue:** The script's Floor 3 semantics are explicitly documented as weaker than the SKILL's stated target:
  ```
  # Floor #3 semantics
  # 100% public-API coverage is enforced by the EXISTENCE of an exercising
  # test, not by requiring 100.0% statement coverage on the symbol's body.
  ```
  The `go-testing-standards` SKILL.md states "100% on public API surface". The script's interpretation allows `DoubleMetaphoneKeys` at 62% to pass Floor 3, meaning 18 exported functions with 62–95% statement coverage are not caught by any automated CI gate.
- **Standard:** `go-testing-standards` SKILL.md §Coverage Targets ("100% on public API surface")
- **Action:** Discuss-phase needed — either (a) update `verify-coverage-floors.sh` to enforce 100% statement coverage on exported functions (stricter, aligns with skill text), or (b) update the `go-testing-standards` SKILL to acknowledge that "100%" means "fully exercised, non-zero" (weaker, aligns with current script). The script's current behaviour is intentionally documented; the skill may need clarification of intent.
- **Rationale:** The current enforcement gap means 18 public API functions have uncovered branches in production code and no automated gate will fail until the misalignment is resolved.

---

### [Important] Orphaned TODO without real issue number — partial_ratio.go:148

- **File:** `/Users/johnny/Development/fuzzymatch/partial_ratio.go`, line 148
- **Phase introduced:** Phase 6 (`06-token-based-algorithms`)
- **Issue:** The grep check for orphaned TODOs finds:
  ```
  partial_ratio.go:148://  TODO(#TBD): implement sliding-window DP per Bachmann RapidFuzz
  partial_ratio.go:153://  DP implementation; this TODO will be updated with the issue
  ```
  The pattern `#TBD` does not match `(#[0-9]+)`. The CLAUDE.md standard requires every TODO to reference a real GitHub issue number.
- **Standard:** CLAUDE.md §GitHub Issues ("Every TODO must reference a GitHub issue: `// TODO(#42): ...`"); Check item 12 in the automated quality gate
- **Action:** Code fix — create a GitHub issue for the sliding-window DP optimisation and replace `#TBD` with the real issue number. The comment already describes the work (spec line 612, O(|s|·|l|) variant deferred to v1.x) so the issue content is ready.
- **Rationale:** The orphaned-TODO check (`grep -rn "TODO\|FIXME\|HACK\|BUG" --include="*.go" . | grep -v "_test.go" | grep -v "(#[0-9]"`) is part of the project's automated quality gate (item 12 in CLAUDE.md Checks). An unreferenced TODO cannot be tracked or prioritised.

---

## Improvement Findings

### [Improvement] gosec not pre-installed — required manual installation during gate run

- **File:** global / CI configuration
- **Phase introduced:** Phase 1 (`01-foundation-infrastructure`)
- **Issue:** `gosec` was not present on the developer machine at the start of this gate run. The command `gosec ./...` failed with `zsh: command not found: gosec`. It was installed with `go install github.com/securego/gosec/v2/cmd/gosec@latest` before the scan could proceed. The `security.yml` CI workflow presumably installs gosec via its action, but the local developer workflow (`make security`) only invokes `govulncheck` and does not install or invoke gosec locally (Makefile `security` target calls `$(GOVULNCHECK) ./...` only).
- **Standard:** CLAUDE.md "Checks" item 10 ("gosec ./..."); `go-coding-standards` SKILL.md §Quality Gate
- **Action:** Skill clarification — update the `make security` target in `Makefile` to invoke gosec in addition to govulncheck, with the same tolerant "if not installed, print install instructions" pattern already used for govulncheck. Alternatively document that gosec is a CI-only gate and is not expected locally.
- **Rationale:** gosec found 3 real findings (mra.go G602, see Important finding 7) that are now blocked in CI via `security.yml`. Developers running `make check` locally will not see these findings without local gosec installation.

---

### [Improvement] markdownlint-cli2 not installed

- **File:** global / CI configuration
- **Phase introduced:** Phase 1 (`01-foundation-infrastructure`)
- **Issue:** `markdownlint-cli2` is not available locally. The command `which markdownlint-cli2` returns exit 1. The tool is listed as spec-locked in CLAUDE.md "Recommended Stack" and is expected by the CI `ci.yml` workflow.
- **Standard:** CLAUDE.md Recommended Stack ("markdownlint-cli2 v0.22.1 — Markdown linting for README, docs/, CHANGELOG"); CLAUDE.md "Checks" item 11
- **Action:** Improvement — document the `npm install -g markdownlint-cli2@0.22.1` local install step in `CONTRIBUTING.md` or a `docs/dev-setup.md` file. The tool cannot be invoked via `go install` so it requires a separate Node-based install step not currently documented for local developer setup.
- **Rationale:** Without `markdownlint-cli2`, markdown quality is only enforced in CI, not locally. Given the project's extensive `docs/` directory (requirements.md, algorithms.md, etc.), markdown lint drift is likely to accumulate between PR cycles.

---

### [Improvement] scripts/verify-llms-sync.sh absent — llms sync enforced by test only

- **File:** `/Users/johnny/Development/fuzzymatch/scripts/` (file does not exist)
- **Phase introduced:** Phase 1 (`01-foundation-infrastructure`)
- **Issue:** The automated quality gate specifies `bash scripts/verify-llms-sync.sh` (CLAUDE.md Checks item 11, referenced in Makefile as `verify-llms-sync`). The script does not exist in `scripts/`. The Makefile does not have a `verify-llms-sync` target. The llms sync check IS enforced programmatically by `ai_friendly_test.go` (which runs as part of `go test ./...` and passes cleanly), but the standalone script form expected by the gate spec is absent.
- **Standard:** CLAUDE.md §Makefile Targets ("make verify-llms-sync or equivalent"); `documentation-standards` SKILL.md §AI-Friendly Documentation
- **Action:** Improvement — either (a) create `scripts/verify-llms-sync.sh` as a thin wrapper that runs `go test -run TestAIFriendly ./...` (consistent with how the test already enforces the invariant), or (b) add a `verify-llms-sync` Makefile target that invokes the same test. The underlying enforcement by `ai_friendly_test.go` is correct and complete; only the script/target form is missing.
- **Rationale:** The missing script means the gate cannot be invoked standalone (`bash scripts/verify-llms-sync.sh`) as documented. The enforcement is present but through a different mechanism than the spec describes.

---

## Summary

| Check | Command | Exit Code | Outcome |
|---|---|---|---|
| fmt-check (gofmt + goimports) | `make fmt-check` | 0 | PASS |
| go vet | `go vet ./...` | 0 | PASS |
| golangci-lint | `golangci-lint run ./...` | **1** | **FAIL** — 27 issues |
| root tests | `go test -race -count=1 ./...` | 0 | PASS (59s) |
| BDD tests | `cd tests/bdd && go test -race -count=1 ./...` | 0 | PASS (12s) |
| go mod tidy (root) | `go mod tidy && git diff go.sum` | **1** | **FAIL** — 3 stale go.sum entries |
| go mod tidy (BDD) | `cd tests/bdd && go mod tidy && git diff` | 0 | PASS |
| coverage | `go test -race -coverprofile=coverage.out ./...` | 0 | 90.2% |
| coverage floors | `bash scripts/verify-coverage-floors.sh` | **1** | **FAIL** — 90.2% < 95.0% |
| govulncheck | `govulncheck ./...` | 0 | PASS — no vulnerabilities |
| gosec | `gosec ./...` | 0 | 4 issues (all G602 in mra.go, LOW severity) |
| markdownlint-cli2 | — | — | NOT INSTALLED |
| verify-no-runtime-deps | `bash scripts/verify-no-runtime-deps.sh` | 0 | PASS (golang.org/x/text allowlisted) |
| verify-license-headers | `bash scripts/verify-license-headers.sh` | 0 | PASS (165 files) |
| verify-coverage-floors | `bash scripts/verify-coverage-floors.sh` | **1** | **FAIL** (same as coverage check above) |
| verify-llms-sync | — | — | Script absent; `TestAIFriendly` passes |
| verify-determinism | `go test -run TestGolden_ ./...` | 0 | PASS |
| No orphaned TODOs | grep check | — | 1 violation: `partial_ratio.go:148 TODO(#TBD)` |
| No testify in root | grep check | — | PASS (only comment references, no imports) |
| No GPL/LGPL code | grep check | — | PASS (all references are negating: "provenance: none") |
| No log.Fatal/os.Exit | grep check | — | PASS (0 matches in non-test library code) |
| No fmt.Println/Printf | grep check | — | PASS (only in examples/ standalone programs) |
| No import . | grep check | — | PASS |
| No init() functions | grep check | — | PASS |
| No type name stutter | grep check | — | PASS |
