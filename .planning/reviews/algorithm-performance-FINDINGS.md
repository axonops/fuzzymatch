---
status: issues_found
agent: algorithm-performance-reviewer
scope: entire codebase (phases 1-8)
reviewed: 2026-05-17T10:00:00Z
finding_counts:
  critical: 6
  important: 13
  improvement: 8
  total: 27
---

# Algorithm Performance Review — Phases 1–8 Comprehensive Findings

Platform: darwin/arm64 (Apple M2), Go 1.26.3.
Benchmark command: `go test -bench=. -benchmem -count=3 -run=^$ ./...`
Baseline: `bench.txt` (1296 lines, count=10 per benchmark, darwin/arm64 Apple M2).

---

## Critical Findings

### [Critical] Scorer allocation budget breach: 12/34 allocs vs §14.2 ≤ 8
- **File:** `/Users/johnny/Development/fuzzymatch/scorer.go`, `/Users/johnny/Development/fuzzymatch/scorer_bench_test.go`
- **Phase introduced:** Phase 8
- **Issue:** `DefaultScorer().Score("abc","abcd")` = 12 allocs/op. `DefaultScorer().Score(scorerAMedium, scorerBMedium)` = 34 allocs/op. Spec §14.2 mandates ≤ 8 allocs/op for `DefaultScorer().Score` on ASCII inputs ≤ 50 chars. §20 acceptance criteria lists "All section-14 budgets met" as a v1.0 ship gate.
- **Standard:** `docs/requirements.md` §14.2, §20; performance-standards.md "Scorer Budgets"
- **Action:** Discuss-phase needed — the §14.2 budget of ≤ 8 is mathematically impossible given the summed per-algorithm §14.1 budgets for the 6-algorithm DefaultScorer (summed floor = 16). Options: (A) revise §14.2 to ≤ 16 for Short + ≤ 40 for Medium, (B) fix DoubleMetaphone Builder→[4]byte to save 4 allocs bringing Short to exactly 8. See 08-PERFORMANCE-REVIEW.md for full analysis.
- **Rationale:** Acceptance criteria cannot be declared met with a documented breach; benchstat CI is blind to Scorer regressions until bench.txt is updated.
- **Suggested fix:** Implement the [4]byte replacement in `DoubleMetaphoneKeys` (saves 4 allocs) + revise §14.2 budget to ≤ 16 for ASCII Short. See 08-PERFORMANCE-REVIEW.md PERF-03 for the exact patch pattern.

---

### [Critical] Scorer benchmarks absent from bench.txt — benchstat CI blind
- **File:** `/Users/johnny/Development/fuzzymatch/bench.txt`, `/Users/johnny/Development/fuzzymatch/scorer_bench_test.go`
- **Phase introduced:** Phase 8
- **Issue:** `bench.txt` (1296 lines) contains zero `BenchmarkDefaultScorer_*` entries. Six scorer benchmarks shipped in Phase 8 (`BenchmarkDefaultScorer_Score_ASCII_Short`, `_ASCII_Medium`, `_ASCII_Long`, `_Unicode_Short`, `_ScoreAll_ASCII_Short`, `_Match_ASCII_Short`) but `scripts/update-bench-txt.sh` was never run. Any future commit that doubles Scorer allocation count or wall time produces no benchstat signal. This is an unconditional regression-detection gap for the entire Scorer layer.
- **Standard:** performance-standards.md "Updating bench.txt"; `docs/requirements.md` §14.4
- **Action:** Code fix — run `go test -bench=BenchmarkDefaultScorer -benchmem -count=10 -run=^$ ./...` on the self-hosted runner, append results to `bench.txt`, commit before v1.0 tag.
- **Rationale:** The benchstat CI job (which catches >10% regressions) has no baseline rows to compare against. Any regression to the Scorer introduced after Phase 8 is invisible to CI.
- **Suggested fix:** `go test -bench=BenchmarkDefaultScorer -benchmem -count=10 -run=^$ ./... >> bench.txt` on the self-hosted benchmark runner; verify `benchstat bench.txt` exits 0.

---

### [Critical] DamerauLevenshteinFull ASCII Short/Medium: 1 alloc vs §14.1 0-alloc budget
- **File:** `/Users/johnny/Development/fuzzymatch/damerau_full.go`
- **Phase introduced:** Phase 2
- **Issue:** `BenchmarkDamerauLevenshteinFullScore_ASCII_Short` = 1 alloc/op (128 B/op), `BenchmarkDamerauLevenshteinFullScore_ASCII_Medium` = 1 alloc/op (21760 B/op). Spec §14.1 specifies "0 allocations" for Damerau-Levenshtein Full on ASCII ≤ 50 char inputs. The implementation uses a full O(m·n) heap-allocated DP table for all inputs, including ASCII Short/Medium, instead of the two-row + stack optimisation. The file godoc explicitly documents this as "v1.x follow-up" but the §14.1 budget has no such exception clause.
- **Standard:** `docs/requirements.md` §14.1 "Damerau-Levenshtein Full: < 3 µs per call, 0 allocations"; performance-standards.md "Two-Row DP Optimisation"
- **Action:** Discuss-phase needed — either (A) implement the two-row + auxiliary anchor-table optimisation for ASCII Short/Medium inputs to achieve 0 allocs, or (B) amend §14.1 to document the O(m·n) space requirement as an accepted exception for the Lowrance-Wagner formulation. The wall-time budget (< 3 µs) passes (64 ns at Short, 8.2 µs at Medium — note Medium exceeds 3 µs budget too, see finding below).
- **Rationale:** A 1-alloc/op number in bench.txt for a budget of 0 allocs is a documented known gap that blocks §20 acceptance criteria.
- **Suggested fix:** Implement the stack-allocated scratch buffer pattern for the Short path (inputs ≤ maxStackInputLen): allocate a `[(maxStackInputLen+2)*(maxStackInputLen+2)]int` stack array as the DP table when both m and n fit within the threshold, falling back to `make([]int, size)` for larger inputs.

---

### [Critical] DamerauLevenshteinFull ASCII Medium wall-time: 8.2 µs vs §14.1 < 3 µs budget
- **File:** `/Users/johnny/Development/fuzzymatch/damerau_full.go`
- **Phase introduced:** Phase 2
- **Issue:** `BenchmarkDamerauLevenshteinFullScore_ASCII_Medium` (50-char inputs) = 8218 ns/op in bench.txt. Spec §14.1 states "Damerau-Levenshtein Full: < 3 µs per call, 0 allocations" for ASCII ≤ 50 chars. The full O(m·n) DP table allocation for every call (including 50-char inputs) drives both the alloc count and the wall-time breach. A 50×50 table requires (52)×(52)×8 = ~21 KB — matching the 21760 B/op in bench.txt.
- **Standard:** `docs/requirements.md` §14.1; performance-standards.md "Two-Row DP Optimisation"
- **Action:** Code fix — the Lowrance-Wagner formulation cannot reduce to a standard two-row DP because the transposition term `D[l-1,k-1]` references arbitrary prior rows. However the `da[256]int` auxiliary table already lives on the stack; the full DP table itself is the bottleneck. A column-at-a-time compression or a condensed sparse representation for the transposition anchor rows is the approach. Alternatively, amend §14.1 to document the DL-Full medium budget as "< 10 µs per call" (matching the actual 8.2 µs) with an explicit note that O(m·n) space is required by the Lowrance-Wagner algorithm.
- **Rationale:** The spec budget of < 3 µs for DL-Full at Medium is physically unachievable without changing the algorithm itself (OSA achieves it via three-row rolling DP, but Full requires the full table or a column-sparse variant). The budget was likely copied from OSA without accounting for the structural difference.
- **Suggested fix:** Amend §14.1 to read "Damerau-Levenshtein Full: < 10 µs per call at ASCII ≤ 50 chars (O(m·n) space is structurally required by the Lowrance-Wagner algorithm; 0 allocations is achievable at Short ≤ 10 chars via stack-resident DP table)." Update `damerau_full.go` godoc accordingly.

---

### [Critical] DoubleMetaphone Score: 6 allocs vs §14.1 ≤ 2-alloc budget
- **File:** `/Users/johnny/Development/fuzzymatch/double_metaphone.go`
- **Phase introduced:** Phase 7
- **Issue:** `BenchmarkDoubleMetaphoneScore_ASCII_Short` = 6 allocs/op (48 B/op). Spec §14.1 states "Double Metaphone: < 2 µs per call, ≤ 2 allocations." The 6 allocs come from: 2× `dmPrep` → `string(stackBuf[:n])` (2 allocs despite stack buffer), 2× `p.String()` from `strings.Builder`, 2× `alt.String()` from `strings.Builder`. The `strings.Builder.String()` method always heap-allocates the returned string even when the builder's internal buffer is small (4 bytes max for DM keys).
- **Standard:** `docs/requirements.md` §14.1 "Double Metaphone: < 2 µs per call, ≤ 2 allocations"; performance-standards.md "Per-Algorithm Budgets"
- **Action:** Code fix — replace `strings.Builder` primary/secondary key accumulators in `DoubleMetaphoneKeys` with `[dmMaxLen]byte` arrays + length counters. Return `string(pBuf[:pLen]), string(altBuf[:altLen])` — still 2 allocs for the returned strings, but eliminates the 4 Builder.String() allocs. Total drops from 6 → 2. This change has zero correctness risk (keys are bounded at 4 bytes by `dmMaxLen`).
- **Rationale:** This is a 3× budget overshoot and is directly fixable with a `[4]byte` pattern. The fix also saves 4 allocs in the DefaultScorer path, which is required to approach the §14.2 budget (see Scorer finding).
- **Suggested fix:** In `double_metaphone.go`, replace `var primary, secondary strings.Builder` with `var pBuf, altBuf [dmMaxLen]byte` and `var pLen, altLen int`; replace `dmAdd` to write into the byte arrays; return `string(pBuf[:pLen]), string(altBuf[:altLen])`.

---

### [Critical] Normalise ASCII Short: 1 alloc vs §14.3 0-alloc budget
- **File:** `/Users/johnny/Development/fuzzymatch/normalise.go`
- **Phase introduced:** Phase 1
- **Issue:** `BenchmarkNormalise_ASCII_Short` = 1 alloc/op (16 B/op); `BenchmarkNormalise_DefaultOptions_Short` = 1 alloc/op (24 B/op). Spec §14.3 states "Normalise ASCII input ≤ 50 chars: < 200 ns, 0 allocations (stack buffer)." The `string(buf)` conversion at `normalise.go:249` unconditionally allocates a heap string because Go string values must be heap-backed when returned from a function — the stack-resident `buf` backing array cannot be referenced by the returned string (its lifetime would not outlive the function call). The `make([]byte, 0, len(s)*2+1)` does NOT escape (confirmed by escape analysis), but the return conversion does.
- **Standard:** `docs/requirements.md` §14.3; performance-standards.md "Per-Algorithm Budgets"
- **Action:** Discuss-phase needed — the 0-alloc claim in §14.3 is structurally unachievable without `unsafe.String` (which is forbidden by go-coding-standards). Options: (A) amend §14.3 to "≤ 1 allocation (output string heap escape)" for ASCII inputs; (B) document in the normalise.go godoc that the stack buffer eliminates growth allocs but the output string itself requires 1 allocation. No code change required.
- **Rationale:** The committed bench.txt baseline already shows 1 alloc, which is inconsistent with the spec's stated 0-alloc target. Leaving this discrepancy unresolved confuses future algorithm-performance reviews.
- **Suggested fix:** Amend `docs/requirements.md` §14.3 to read: "`Normalise` ASCII input ≤ 50 chars: < 200 ns, ≤ 1 allocation (output string; stack buffer avoids growth allocs but the returned string is heap-backed)." Update `normalise.go` godoc accordingly.

---

## Important Findings

### [Important] Q-Gram algorithms (Jaccard, Sørensen-Dice, Cosine, Tversky): 6–7 allocs at Medium vs §14.1 ≤ 4-alloc budget
- **File:** `/Users/johnny/Development/fuzzymatch/q_gram.go`, `/Users/johnny/Development/fuzzymatch/qgram_jaccard.go`, `/Users/johnny/Development/fuzzymatch/sorensen_dice.go`, `/Users/johnny/Development/fuzzymatch/cosine.go`, `/Users/johnny/Development/fuzzymatch/tversky.go`
- **Phase introduced:** Phase 5
- **Issue:** `BenchmarkQGramJaccardScore_ASCII_Medium` = 6 allocs/op; `BenchmarkSorensenDiceScore_ASCII_Medium` = 6 allocs/op; `BenchmarkTverskyScore_ASCII_Medium` = 6 allocs/op; `BenchmarkCosineScore_ASCII_Medium` = 7 allocs/op. Spec §14.1 budget is "≤ 4 allocations" for these algorithms. The 2–3 extra allocs above budget at Medium are caused by map rehashing: `extractQGrams` sizes the initial map at `len(s)-n+1` but when multiple windows hash to the same bucket, the map grows beyond the initial capacity, triggering additional internal rehash allocs. For a 50-char string with n=3, this yields ~47 trigrams, and the map grows 1–2 times during filling.
- **Standard:** `docs/requirements.md` §14.1 "Q-Gram Jaccard, Sørensen-Dice, Cosine, Tversky: < 5 µs per call, ≤ 4 allocations"
- **Action:** Code fix (optional for v0.x, required for v1.0). Options: (A) use a larger capacity multiplier (e.g. `len(s)-n+1)*2`) to reduce rehash probability; (B) implement a small-input fast path using a fixed-size `[64][3]byte` sorted array for trigrams when `len(s) <= 20` (avoids map allocation entirely); (C) amend §14.1 to "≤ 6 allocations" reflecting the observed behaviour.
- **Rationale:** The 6-alloc Medium result is documented in bench.txt as the baseline, meaning CI will accept it as the new floor. However, it breaches the spec budget, which future reviewers will flag again.
- **Suggested fix:** In `q_gram.go`, change the map initial capacity to `(len(s)-n+1) + (len(s)-n+1)/4` (add 25% headroom) to reduce rehash probability without over-allocating for short inputs.

---

### [Important] Token Sort Ratio: 14 allocs at ASCII Short vs §14.1 ≤ 4-alloc budget
- **File:** `/Users/johnny/Development/fuzzymatch/token_sort_ratio.go`, `/Users/johnny/Development/fuzzymatch/tokenise.go`
- **Phase introduced:** Phase 6
- **Issue:** `BenchmarkTokenSortRatioScore_ASCII_Short` = 14 allocs/op (474 B/op) in bench.txt. Spec §14.1 states "Token Sort Ratio, Token Set Ratio, Token Jaccard: < 5 µs per call, ≤ 4 allocations." The 14-alloc count is driven by `Tokenise` producing multiple heap strings from rune-to-string conversion per token, plus the intermediate sorted string construction. For a short input like "hello world" (2 tokens), this breaks down as: 2 Tokenise output slices + per-token strings + sorted join + Levenshtein intermediate.
- **Standard:** `docs/requirements.md` §14.1 "Token Sort Ratio … ≤ 4 allocations"; performance-standards.md "Per-Algorithm Budgets"
- **Action:** Code fix — the Tokenise function at `/Users/johnny/Development/fuzzymatch/tokenise.go` allocates per-token strings via `string(rs)` conversion. A pooled token buffer or a `strings.Builder`-based join avoiding intermediate per-token strings would reduce alloc count. However, given the inherent cost of sorting and joining tokens, 4 allocs is a very aggressive target; a realistic budget is ≤ 8 allocs for short inputs.
- **Rationale:** The bench.txt baseline shows 14 allocs, which is 3.5× the spec budget. This is tracked in bench.txt without having been flagged, meaning CI cannot catch a further regression relative to the spec target.
- **Suggested fix:** Amend §14.1 to "Token Sort Ratio, Token Set Ratio, Token Jaccard: < 5 µs per call, ≤ 10 allocations (token string allocations are proportional to token count)" until a Tokenise pooling optimisation is implemented.

---

### [Important] Token Set Ratio: 9 allocs at ASCII Short vs §14.1 ≤ 4-alloc budget
- **File:** `/Users/johnny/Development/fuzzymatch/token_set_ratio.go`
- **Phase introduced:** Phase 6
- **Issue:** `BenchmarkTokenSetRatioScore_ASCII_Short` = 9 allocs/op (245 B/op). Same budget violation as Token Sort Ratio. Set construction (two `map[string]struct{}` + intersection/difference logic) adds 2 extra map allocs on top of the base Tokenise cost.
- **Standard:** `docs/requirements.md` §14.1 "Token Set Ratio … ≤ 4 allocations"
- **Action:** Code fix (same family as Token Sort Ratio). Alternatively amend budget.
- **Rationale:** 9 allocs is 2.25× the 4-alloc budget. Bench.txt records it as the baseline.
- **Suggested fix:** Same as Token Sort Ratio — revise §14.1 budget upward to reflect the inherent token-string allocation cost.

---

### [Important] Tokenise: 4 allocs at ASCII Short vs §14.3 ≤ 2-alloc budget
- **File:** `/Users/johnny/Development/fuzzymatch/tokenise.go`
- **Phase introduced:** Phase 6 (tokenise)
- **Issue:** `BenchmarkTokenise_ASCII_Short` = 4 allocs/op (73 B/op). Spec §14.3 states "Tokenise ASCII input ≤ 50 chars: < 500 ns, ≤ 2 allocations (token slice + storage)." The actual 4 allocs exceed the budget by 2. The extra allocs come from per-token `string(rs)` conversions where `rs` is a rune sub-slice — each token's rune run is converted to a heap string.
- **Standard:** `docs/requirements.md` §14.3 "Tokenise ASCII input ≤ 50 chars: < 500 ns, ≤ 2 allocations"
- **Action:** Code fix — on the ASCII fast path, token boundaries can be computed as byte offsets into the original string, and each token can be returned as a slice of the input string (zero-copy, no per-token alloc). This requires the Tokenise function to operate on bytes directly for ASCII inputs. Alternatively amend §14.3 to ≤ 4 allocs.
- **Rationale:** Tokenise is on the hot path for all 5 token-based algorithms. Each call allocates per-token strings. For ASCII inputs, a substring-based approach avoids these allocs.
- **Suggested fix:** In `tokenise.go`, add an ASCII fast path that identifies token boundaries as byte-index ranges and returns substrings of the original input string (using `s[lo:hi]`) rather than creating new strings from rune slices.

---

### [Important] Soundex/NYSIIS/MRA Code functions: 1 alloc vs §14.1 0-alloc budget
- **File:** `/Users/johnny/Development/fuzzymatch/soundex.go`, `/Users/johnny/Development/fuzzymatch/nysiis.go`, `/Users/johnny/Development/fuzzymatch/mra.go`
- **Phase introduced:** Phase 7
- **Issue:** `BenchmarkSoundexCode_ASCII_Short` = 1 alloc/op (4 B/op); `BenchmarkNYSIISCode_ASCII_Short` = 1 alloc/op (4 B/op); `BenchmarkMRACode_ASCII_Short` = 1 alloc/op (4 B/op). Spec §14.1 states "Soundex, NYSIIS, MRA: < 500 ns per call, 0 allocations (stack-allocated code buffer)." All three functions use stack-allocated intermediate buffers (`[64]byte`, `[6]byte` etc.) but return the result via `string(result[:n])` or `string(result[:])` which necessarily allocates a heap-backed string.
- **Standard:** `docs/requirements.md` §14.1 "Soundex, NYSIIS, MRA: < 500 ns per call, 0 allocations (stack-allocated code buffer)"
- **Action:** Discuss-phase needed — same structural situation as Normalise. The `string(buf)` conversion is unavoidable without `unsafe.String`. The spec's "0 allocations" claim is incorrect. Options: (A) amend §14.1 to "≤ 1 allocation (output string)"; (B) note that `MRAScore` and `SoundexScore` (which call Code twice and compare strings) budget at 2 allocs for the two Code calls — which matches bench.txt values.
- **Rationale:** `BenchmarkSoundexScore_ASCII_Short` = 2 allocs/op (matching 2 SoundexCode calls), `BenchmarkMRAScore_ASCII_Short` = 2 allocs/op, `BenchmarkNYSIISScore_Match` = 2 allocs/op. These are consistent with 1 alloc per Code call. The spec says 0 allocs but achieves 1.
- **Suggested fix:** Amend §14.1 to "Soundex, NYSIIS, MRA: < 500 ns per call, ≤ 1 allocation per Code call (output string; intermediate buffers are stack-allocated)."

---

### [Important] MRA Compare/Score: 2 allocs — budget says 0 (from §14.1 phonetic 0-alloc claim)
- **File:** `/Users/johnny/Development/fuzzymatch/mra.go`
- **Phase introduced:** Phase 7
- **Issue:** `BenchmarkMRACompare_ASCII_Short` = 2 allocs/op (16 B/op); `BenchmarkMRAScore_ASCII_Short` = 2 allocs/op (16 B/op); `BenchmarkMRAScore_Match` = 2 allocs/op; `BenchmarkMRAScore_NoMatch` = 2 allocs/op. `MRACompare` calls `MRACode(a)` and `MRACode(b)`, each of which allocates 1 string. There is no alloc-free path for MRACompare since both Code results must be held simultaneously for comparison.
- **Standard:** `docs/requirements.md` §14.1 "Soundex, NYSIIS, MRA: < 500 ns per call, 0 allocations"
- **Action:** Same as above — amend §14.1 to match the structural reality.
- **Rationale:** MRACompare necessarily calls MRACode twice. If MRACode costs 1 alloc each, MRACompare minimum is 2 allocs. Budget needs updating.
- **Suggested fix:** Include in the §14.1 amendment: "MRACompare: ≤ 2 allocations (two MRACode calls)."

---

### [Important] DamerauLevenshteinOSA Unicode Short: 3 allocs — exceeds documented 0-alloc scope boundary
- **File:** `/Users/johnny/Development/fuzzymatch/damerau_osa.go`
- **Phase introduced:** Phase 2
- **Issue:** `BenchmarkDamerauLevenshteinOSAScore_Unicode_Short` = 3 allocs/op (144 B/op). This is expected (2 `[]rune` allocs + 3 DP rows on the rune path), but bench.txt records 3 allocs when the Unicode path uses `make([]int, n+1)` × 3 (three-row DP). The rune path allocates 2 `[]rune` slices + 3 DP row slices = 5 allocations, but the identity short-circuit and the rune-count-based row sizing reduces the effective count to 3 for short equal inputs. The 0-alloc budget applies only to ASCII; Unicode is documented as not covered.
- **Standard:** performance-standards.md "Per-Call Allocations" (0-alloc budget for ASCII ≤ 50 chars)
- **Action:** No code fix needed. Improvement — add a godoc clarification to `DamerauLevenshteinOSAScoreRunes` stating the 3-alloc expected allocation count for the Unicode path at Short inputs.
- **Rationale:** The bench.txt number is correct and expected. Documenting it explicitly prevents future reviewers from flagging it as a regression.

---

### [Important] LCSStr ASCII Long: 2 allocs — heap path expected but not documented
- **File:** `/Users/johnny/Development/fuzzymatch/lcsstr.go`
- **Phase introduced:** Phase 4
- **Issue:** `BenchmarkLCSStrScore_ASCII_Long` = 2 allocs/op. The file godoc says "Heap path: two make([]int, n+1) calls; 2 allocs on ASCII Long." This is correct and expected. However the performance-standards.md skill states "LCSStr: < 2 µs, 0 allocations" — this applies only to ASCII ≤ 50 chars (Short/Medium). The Long path (>64 chars) correctly falls back to heap allocation and is not covered by the 0-alloc budget.
- **Standard:** performance-standards.md "Per-Call Allocations"; `docs/requirements.md` §14.1 "LCSStr: < 2 µs per call, 0 allocations"
- **Action:** No code fix needed. The bench.txt numbers are correct. The spec §14.1 should clarify that "0 allocations" applies only to inputs ≤ maxStackInputLen (64 chars). Add a note: "For inputs > 64 bytes: 2 allocations (heap path)."
- **Rationale:** The `bench.txt` shows 2 allocs for Long across all count=10 runs, so CI accepts this as baseline. Clarifying the scope of the 0-alloc claim prevents confusion.

---

### [Important] Jaro/JaroWinkler/Strcmp95 ASCII Long: 2–3 allocs — heap path triggers above maxJaroStackLen
- **File:** `/Users/johnny/Development/fuzzymatch/jaro.go`, `/Users/johnny/Development/fuzzymatch/jarowinkler.go`, `/Users/johnny/Development/fuzzymatch/strcmp95.go`
- **Phase introduced:** Phase 2
- **Issue:** `BenchmarkJaroScore_ASCII_Long` = 2 allocs/op (640 B/op); `BenchmarkJaroWinklerScore_ASCII_Long` = 2 allocs/op; `BenchmarkStrcmp95Score_ASCII_Long` = 3 allocs/op (1536 B/op). The `maxJaroStackLen = 256` threshold means inputs > 256 bytes use `make([]bool, la)` and `make([]bool, lb)` (heap). The 500-char "Long" benchmark exceeds this threshold and triggers 2 allocs. Strcmp95 adds a third `make([]bool, lb)` for the `simConsumed` slice. The spec §14.1 budget of "0 allocations" applies only to ASCII ≤ 50 chars (within the stack threshold).
- **Standard:** `docs/requirements.md` §14.1 "Jaro, Jaro-Winkler, Strcmp95: < 1 µs per call, 0 allocations" (applies to ≤ 50 chars)
- **Action:** No code fix needed for v1.0. The 0-alloc budget is only specified for inputs ≤ 50 chars, and all Short/Medium benchmarks correctly show 0 allocs. The Long allocs are expected and should be documented in the respective bench file comments.
- **Rationale:** The bench.txt entries for Long are consistent across count=10 runs, confirming the heap path is stable.

---

### [Important] Levenshtein ASCII Long/Unicode Short: 2 allocs — heap path expected but bench.txt is the only documentation
- **File:** `/Users/johnny/Development/fuzzymatch/levenshtein.go`
- **Phase introduced:** Phase 2
- **Issue:** `BenchmarkLevenshteinScore_ASCII_Long` = 2 allocs/op (8192 B/op); `BenchmarkLevenshteinScore_Unicode_Short` = 2 allocs/op (96 B/op). Both are correct and expected (heap two-row DP for Long; 2×`[]rune` for Unicode Short). No budget violation. However the `levenshtein_bench_test.go` comments for the Long benchmark do not document the expected 2-alloc count.
- **Standard:** performance-standards.md "Benchmark File Structure"
- **Action:** Improvement — add `// Expected: 2 allocs/op (heap path for inputs > 64 bytes)` comment to `BenchmarkLevenshteinScore_ASCII_Long` and `BenchmarkLevenshteinScore_Unicode_Short`.
- **Rationale:** Without explicit documentation of the expected alloc count in the benchmark, future reviewers cannot distinguish "expected 2 allocs" from "regression from 0 to 2 allocs."

---

### [Important] Ratcliff-Obershelp Short: 4 allocs — no alloc budget in spec, but above character-algorithm norm
- **File:** `/Users/johnny/Development/fuzzymatch/ratcliff_obershelp.go`
- **Phase introduced:** Phase 4
- **Issue:** `BenchmarkRatcliffObershelpScore_ASCII_Short` = 4 allocs/op (256 B/op). Spec §14.1 gives only "Ratcliff-Obershelp: < 5 µs per call for short inputs" — no allocation budget. However the performance-standards.md skill states that character-based algorithms "should have 0 allocations per call" for ASCII ≤ 50 chars. The 4 allocs at Short come from `roFindLongestMatch` calling `make([]int, lb+1)` × 2 (prev/curr rows) at each level of the recursion. For a short ~10-char input with ~3 levels of recursion, this produces 3×2 = 6 allocs (bench shows 4, implying identity short-circuits prune some branches).
- **Standard:** performance-standards.md "Per-Call Allocations" (character-based algorithms, ASCII ≤ 50 chars)
- **Action:** Discuss-phase needed — Ratcliff-Obershelp is classified as "Gestalt" (not strictly "character-based" in the DP sense) and the spec lacks an explicit alloc budget. The 4 allocs are inherent to the recursive LCS-decomposition unless a stack-allocated buffer pool is added for short inputs. Options: (A) add an explicit allocation budget to §14.1 for Ratcliff-Obershelp (≤ 4 allocs for short inputs); (B) implement a stack-allocated `[maxStackInputLen+1]int` × 2 buffer and pass it down the recursion.
- **Rationale:** The bench.txt records 4 allocs/op at Short consistently. Without a spec budget, CI cannot detect regressions to, say, 8 allocs/op.
- **Suggested fix:** Add `Ratcliff-Obershelp: < 5 µs per call, ≤ 4 allocations for short inputs (recursive LCS decomposition allocates 2 DP rows per recursion level)` to §14.1.

---

### [Important] Ratcliff-Obershelp Long: 200 allocs, 433 KB — no documentation of growth scaling
- **File:** `/Users/johnny/Development/fuzzymatch/ratcliff_obershelp.go`
- **Phase introduced:** Phase 4
- **Issue:** `BenchmarkRatcliffObershelpScore_ASCII_Long` (500-char inputs) = 200 allocs/op (433857 B/op) in bench.txt. This is the O(N² · M) worst-case allocation explosion from the recursive LCS decomposition: each level of recursion allocates 2 DP rows of size `lb`, and recursion depth is O(min(la, lb)) ≈ 250 levels for equal-length strings. The 200-alloc and ~434 KB/op numbers are expected given the algorithm's complexity but are not documented anywhere in the codebase as the expected Long baseline.
- **Standard:** performance-standards.md "Benchmark Coverage" (all benchmarks use `b.ReportAllocs()`)
- **Action:** Improvement — add a comment to `BenchmarkRatcliffObershelpScore_ASCII_Long` stating "INFORMATIONAL: O(N²·M) recursion allocates proportionally to input length; ~200 allocs/op at 500 chars is expected." Also add a `Ratcliff-Obershelp: DoS notice` section to the ratcliff_obershelp.go file header alongside Monge-Elkan and Partial Ratio.
- **Rationale:** A consumer calling `RatcliffObershelpScore` on untrusted long inputs faces unbounded memory allocation. The DoS notice is required by the "Worst-case complexity documentation for DoS-prone algorithms" in the review scope.

---

### [Important] Missing `BenchmarkAlgoID_String` in *_bench_test.go — only in algoid_test.go
- **File:** `/Users/johnny/Development/fuzzymatch/algoid_test.go`
- **Phase introduced:** Phase 1
- **Issue:** `BenchmarkAlgoID_String` is defined in `algoid_test.go` (a non-bench test file) rather than in a dedicated `algoid_bench_test.go` file. The performance-standards.md skill states "Every algorithm has `<algorithm>_bench_test.go`." The AlgoID String method is not an algorithm per se, but it is performance-sensitive (dispatch table lookup). Having the benchmark in the regular test file conflates unit and performance concerns and prevents per-file performance tooling.
- **Standard:** performance-standards.md "Benchmark File Structure"
- **Action:** Improvement — move `BenchmarkAlgoID_String` into a new `algoid_bench_test.go` file. Keep the test in `algoid_test.go` as is.
- **Rationale:** Minor structural issue. Does not affect CI or benchstat.

---

## Improvement Findings

### [Improvement] DamerauLevenshteinFull: O(m·n) space vs O(n) possible via auxiliary anchor compression
- **File:** `/Users/johnny/Development/fuzzymatch/damerau_full.go`
- **Phase introduced:** Phase 2
- **Issue:** The current Lowrance-Wagner implementation allocates a full `(m+2)×(n+2)` table. The file godoc documents a "two-row + auxiliary-anchor-table" optimisation as a v1.x follow-up. While the transposition term does reference arbitrary prior rows, the number of distinct anchor rows is bounded by the alphabet size (256 for ASCII), meaning a sparse representation storing only the last row per character (rather than all m rows) can reduce space to O(n + |alphabet|). This is the standard optimisation for the Lowrance-Wagner algorithm in production string-matching libraries.
- **Standard:** performance-standards.md "Two-Row DP Optimisation"; `docs/requirements.md` §14.1
- **Action:** Improvement — implement the O(n + 256) space optimisation as a v1.x task. Track in a GitHub issue.
- **Suggested fix:** The `da[256]int` array (already stack-allocated) combined with a single-row rolling DP that re-builds the transposition term from `da` alone can reduce the 21760 B/op Medium allocation to ~512 B/op.

---

### [Improvement] Normalise `normaliseASCII`: two-pass (camel-split + separator-strip) could be one pass
- **File:** `/Users/johnny/Development/fuzzymatch/normalise.go`
- **Phase introduced:** Phase 1
- **Issue:** `normaliseASCII` does a first pass to build `buf` (with camel-split spaces inserted), then a second pass via `collapseSeparators` (separator-strip + space-collapse). Both operate on bytes. A single pass with a small state machine (tracking `prev`, `inSep`, `pendingSpace`) could eliminate the second-pass scan.
- **Standard:** performance-standards.md "When Optimisation is NOT Worth It"
- **Action:** Improvement — benchmark the single-pass variant before implementing to confirm the win (the second pass is O(n) over a buffer already in L1 cache, so the gain may be < 10 ns). File as a v1.x micro-optimisation issue.
- **Rationale:** This would save the second-pass allocation-free scan. The current bench shows 60–70 ns for Short, well within the < 200 ns budget. Low priority.

---

### [Improvement] Q-gram `extractQGrams`: map capacity could use `len(s)` instead of `len(s)-n+1`
- **File:** `/Users/johnny/Development/fuzzymatch/q_gram.go`
- **Phase introduced:** Phase 5
- **Issue:** The capacity hint `len(s)-n+1` is the exact number of windows but underestimates the distinct key count for inputs with repeated patterns (where many windows map to the same key). When the distinct key count < window count (e.g. "aaa...a"), the map stays small; but when key count ≈ window count, the initial capacity is exact and rehashing is avoided. The 6-alloc result at Medium suggests the initial capacity is sometimes underestimated, triggering 2 rehash allocs.
- **Standard:** performance-standards.md "Per-Algorithm Budgets"
- **Action:** Improvement — add a 25% headroom: `make(map[string]int, (len(s)-n+1)*5/4)`. Benchmark to confirm alloc count drops from 6 to 4 at Medium.
- **Suggested fix:** Change `q_gram.go:111` to `m := make(map[string]int, (len(s)-n+1)*5/4)`.

---

### [Improvement] `SmithWatermanGotohScore` Medium wall time at 50 chars is 8.9 µs vs stated "< 5 µs" budget interpretation
- **File:** `/Users/johnny/Development/fuzzymatch/swg.go`
- **Phase introduced:** Phase 3
- **Issue:** `BenchmarkSmithWatermanGotohScore_ASCII_Medium` (50-char inputs) = 8914 ns/op in bench.txt. Spec §14.1 states "Smith-Waterman-Gotoh: < 5 µs per call, 0 allocations (stack buffer for ≤ 50 char inputs)." If the budget is interpreted as "for inputs at exactly 50 chars" then it is breached (8.9 µs vs 5 µs). If interpreted as "for short inputs up to ~20 chars" then the Short benchmark (196 ns) passes comfortably.
- **Standard:** `docs/requirements.md` §14.1 "Smith-Waterman-Gotoh: < 5 µs per call"
- **Action:** Discuss-phase needed — the budget language "≤ 50 char inputs" conflicts with the observed 8.9 µs at exactly 50 chars. SWG has O(mn) complexity; at 50×50 chars, even 6 rolling float64 rows (the Gotoh formulation) requires 3000 float64 operations. Clarify §14.1 to "Smith-Waterman-Gotoh: < 5 µs per call for inputs ≤ 20 chars (short identifier comparison); ≤ 10 µs for 50-char inputs."
- **Rationale:** The current 8.9 µs at 50 chars is a wall-time concern, not an allocation concern (0 allocs at Medium is correct). The < 5 µs budget is likely written for the typical "user_id vs userId" use case (~10 chars) rather than the full 50-char boundary.

---

### [Improvement] `Tokenise` bench file missing explicit expected-alloc comments
- **File:** `/Users/johnny/Development/fuzzymatch/tokenise_bench_test.go`
- **Phase introduced:** Phase 6
- **Issue:** The tokenise bench file does not document the expected alloc counts in benchmark comments. `BenchmarkTokenise_ASCII_Short` = 4 allocs/op and `BenchmarkTokenise_DefaultOptions` = 4 allocs/op; these exceed the §14.3 budget of ≤ 2 allocs and there is no comment explaining why.
- **Standard:** performance-standards.md "Benchmark File Structure"
- **Action:** Improvement — add expected-alloc comments to each tokenise benchmark, and cross-reference the §14.3 budget discussion (or its amended form once resolved).
- **Rationale:** Without comments, reviewers cannot distinguish "known budget miss" from "regression."

---

### [Improvement] No `BenchmarkXxx_ASCII_Long` for phonetic Score functions (Soundex, NYSIIS, MRA)
- **File:** `/Users/johnny/Development/fuzzymatch/soundex_bench_test.go`, `/Users/johnny/Development/fuzzymatch/nysiis_bench_test.go`, `/Users/johnny/Development/fuzzymatch/mra_bench_test.go`
- **Phase introduced:** Phase 7
- **Issue:** Phonetic `Code` benchmarks have Short/Medium/Long variants. But `BenchmarkSoundexScore_*` only has `ASCII_Short` and `ASCII_Identity`. `BenchmarkNYSIISScore_*` only has `Match` and `NoMatch`. `BenchmarkMRAScore_*` only has `ASCII_Short`, `Match`, and `NoMatch`. There are no `ASCII_Medium` or `ASCII_Long` Score benchmarks for these three algorithms. The bench structure requirement says "All benchmarks cover Short, Medium, Long."
- **Standard:** performance-standards.md "Benchmark File Structure" (requires `BenchmarkXxxScore_ASCII_Short`, `_ASCII_Medium`, `_ASCII_Long`)
- **Action:** Improvement — add `BenchmarkSoundexScore_ASCII_Medium`, `BenchmarkNYSIISScore_ASCII_Medium`, `BenchmarkMRAScore_ASCII_Medium` score benchmarks to the respective bench files. These are informational (the algorithms have a bounded-length output) but ensure benchstat tracks score cost at different input lengths.
- **Rationale:** The Code functions have Medium/Long benchmarks, but the Score wrappers do not. Score includes two Code calls, so a Medium Score would show 2 allocs (2 Code calls) at all input lengths — confirming the budget.

---

### [Improvement] `DoubleMetaphoneScore_Identity` benchmark name is non-standard
- **File:** `/Users/johnny/Development/fuzzymatch/double_metaphone_bench_test.go`
- **Phase introduced:** Phase 7
- **Issue:** `BenchmarkDoubleMetaphoneScore_Identity` exists but there is no corresponding `BenchmarkSoundexScore_ASCII_Medium`, `BenchmarkNYSIISScore_ASCII_Medium` etc. The `_Identity` name doesn't follow the `_ASCII_Short/_ASCII_Medium/_ASCII_Long/_Unicode_Short` naming convention in performance-standards.md. This causes benchstat comparisons to fail if the benchmark is renamed.
- **Standard:** performance-standards.md "Benchmark File Structure"
- **Action:** Improvement — rename `BenchmarkDoubleMetaphoneScore_Identity` to `BenchmarkDoubleMetaphoneScore_ASCII_Identity` for naming consistency, or add `_ASCII_Short` as the primary "identity case" benchmark.
- **Rationale:** Minor naming inconsistency. Does not affect benchstat unless the rename triggers a benchstat "new benchmark" detection.

---

## Benchstat Regression Summary

The following benchmarks in `bench.txt` show values that breach or approach their §14.1 budgets:

| Benchmark | bench.txt allocs | §14.1 budget | Status |
|---|---|---|---|
| `DamerauLevenshteinFullScore_ASCII_Short` | 1 | 0 | BREACH |
| `DamerauLevenshteinFullScore_ASCII_Medium` | 1 | 0 | BREACH |
| `CosineScore_ASCII_Medium` | 7 | ≤ 4 | BREACH |
| `QGramJaccardScore_ASCII_Medium` | 6 | ≤ 4 | BREACH |
| `SorensenDiceScore_ASCII_Medium` | 6 | ≤ 4 | BREACH |
| `TverskyScore_ASCII_Medium` | 6 | ≤ 4 | BREACH |
| `TokenSortRatioScore_ASCII_Short` | 14 | ≤ 4 | BREACH |
| `TokenSetRatioScore_ASCII_Short` | 9 | ≤ 4 | BREACH |
| `TokenJaccardScore_ASCII_Short` | 2 | ≤ 4 | PASS |
| `DoubleMetaphoneScore_ASCII_Short` | 6 | ≤ 2 | BREACH |
| `DoubleMetaphoneKeys_ASCII_Short` | 3 | ≤ 2 | BREACH |
| `SoundexCode_ASCII_Short` | 1 | 0 | BREACH |
| `NYSIISCode_ASCII_Short` | 1 | 0 | BREACH |
| `MRACode_ASCII_Short` | 1 | 0 | BREACH |
| `Normalise_ASCII_Short` | 1 | 0 | BREACH |
| `Tokenise_ASCII_Short` | 4 | ≤ 2 | BREACH |

No benchstat regressions have been detected relative to `bench.txt` (the bench.txt IS the baseline). All numbers above are the established bench.txt baseline. Regressions from this baseline would be caught by CI.

## GO/NO-GO

**NO-GO for v1.0.0 tagging** on the following conditions:

1. `BenchmarkDefaultScorer_*` entries are absent from `bench.txt` — CI regression detection is blind for the Scorer layer.
2. `docs/requirements.md` §14.2 Scorer budget (≤ 8 allocs) is mathematically inconsistent with the summed §14.1 per-algorithm budgets and must be formally revised before the §20 acceptance criteria can be declared met.
3. `docs/requirements.md` §14.1 states "0 allocations" for DamerauLevenshteinFull but the implementation allocates 1/op for all input sizes including ASCII Short — the spec and implementation are in documented disagreement with no formal waiver.
4. `docs/requirements.md` §14.1 states "0 allocations" for Soundex/NYSIIS/MRA but all three allocate 1/op per Code call — same category of spec/implementation mismatch.
5. `docs/requirements.md` §14.3 states "0 allocations" for Normalise but 1/op is structurally unavoidable — requires spec amendment.

All five are addressable primarily via spec amendments plus the DoubleMetaphone [4]byte fix. No fundamental architectural changes are required to achieve v1.0 readiness.
