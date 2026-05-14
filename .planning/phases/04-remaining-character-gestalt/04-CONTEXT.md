---
phase: 04-remaining-character-gestalt
phase_number: 4
date: 2026-05-14
spec_loaded: false
prior_decisions_consulted: [01-CONTEXT.md, 02-CONTEXT.md, 03-CONTEXT.md, 02-VERIFICATION.md, 03-VERIFICATION.md, PITFALLS.md]
---

# Phase 4: Remaining Character & Gestalt — Context

**Gathered:** 2026-05-14
**Status:** Ready for planning

<domain>
## Phase Boundary

**What this phase delivers:** the three remaining character-and-gestalt
similarity algorithms that complete the non-DP-q-gram, non-token, non-phonetic
catalogue tier:

1. **Strcmp95** (CHAR-07) — Winkler 1994's enhancement of Jaro-Winkler with
   the similar-character table + long-string adjustment + AS/I-S/RS-RB
   letter-pair adjustments. Full Winkler 1994 spec — all four adjustments
   stacked atop Jaro.
2. **LCSStr** (CHAR-09) — longest common substring similarity. Returns both
   the substring (`LongestCommonSubstring`) and a normalised score
   (`LCSStrScore`) with byte + rune variants — 4 public functions, wider
   surface than other character algorithms.
3. **Ratcliff-Obershelp** (GESTALT-01) — Dr. Dobb's Journal 1988 recursive
   longest-common-substring matching. The load-bearing **Python
   `difflib.SequenceMatcher.ratio()` equivalent** — exists specifically so
   consumers wanting that semantic have a clean path, distinct from the
   Indel-based token ratios coming in Phase 6.

Each algorithm is a fresh implementation from its primary source, with
the discipline already locked in Phases 2 & 3 inherited intact (see
`<carry_forward>` below).

**Out of scope:** any q-gram / token / phonetic algorithm; the Scorer
composite (Phase 8); the scan sub-package (Phase 9); the Extract API
(Phase 10).
</domain>

<canonical_refs>
## Canonical References

**Downstream agents MUST read these before planning or implementing.**

| Reference | Why |
|-----------|-----|
| `docs/requirements.md` §7.1.7 — Strcmp95 | Public API names locked here, default behaviour, score normalisation, complexity, edge cases |
| `docs/requirements.md` §7.1.9 — LCSStr | Public API names locked at 4 functions (LongestCommonSubstring + Runes + LCSStrScore + Runes); normalisation formula `2·len(lcs)/(len(a)+len(b))` is SPEC-PINNED — not a gray area |
| `docs/requirements.md` §7.1.24 — Ratcliff-Obershelp | Public API names locked here; "difflib-equivalent" is the binding contract |
| `docs/requirements.md` §14 — Performance budgets | Inherits Phase 2/3's ASCII-fast-path + two-row DP discipline |
| `docs/requirements.md` §15 — Test strategy | Unit + property + fuzz + benchmark + BDD per algorithm |
| `.planning/research/PITFALLS.md` §6 — Token ratio vs difflib | The whole reason Ratcliff-Obershelp is a separate algorithm: difflib.ratio() ≠ RapidFuzz Indel ratio. Document the distinction in RatcliffObershelp's godoc; point Token Sort / Set / Partial users here. |
| `.planning/research/PITFALLS.md` §14 — Tables via var, not init() | Strcmp95's similar-character table is the load-bearing test case for this rule. |
| `.planning/research/PITFALLS.md` §7 — Byte-vs-rune indexing | LCSStr is byte+rune; Strcmp95 is byte-only (no *Runes by design — see §2). |
| `.planning/research/PITFALLS.md` §9 — Float determinism | Strcmp95 sums Jaro components + adjustments; left-to-right reduction only; no FMA / no math.X transcendentals. |
| `.claude/skills/algorithm-correctness-standards/SKILL.md` | Primary-source citation, formula docs, mathematical-invariant property tests |
| `.claude/skills/algorithm-licensing-standards/SKILL.md` | Fresh-impl rule. Census Bureau strcmp95.c (public domain, U.S. government work) used for cross-validation vectors only, NOT for table values — source is Winkler 1994 paper. OpenRefine Strcmp95.java (Apache-2.0) consulted only for tie-breaking ambiguities. Python difflib (PSF licence) used for reference-vector cross-validation of Ratcliff-Obershelp, NOT for code copying. |
| `.claude/skills/performance-standards/SKILL.md` | Two-row DP requirement for LCSStr; ASCII fast paths; benchstat regression |
| `.claude/skills/determinism-standards/SKILL.md` | Float-stability rules (no `math.Pow`, no `math.FMA`), no map iteration on output paths, no init() |
| `.claude/skills/go-testing-standards/SKILL.md` | Coverage floors (>= 90% per-file, >= 95% overall), property test conventions |
| `.claude/skills/fuzzymatch-review-protocol/SKILL.md` | Phase-end review gates |
| `.planning/phases/02-core-character-algorithms-six/02-CONTEXT.md` | Inherited file pattern (`<algo>.go` + `dispatch_<algo>.go` + `<algo>_test.go` + `<algo>_bench_test.go` + `<algo>_fuzz_test.go` + staging golden + BDD feature) |
| `.planning/phases/02-core-character-algorithms-six/02-PATTERNS.md` | Pattern 3 (ASCII fast-path gate LOCKED): applies to LCSStr identically |
| `.planning/phases/02-core-character-algorithms-six/02-07-finalisation-SUMMARY.md` | Cross-algorithm consistency test + bench.txt baseline pattern — Phase 4 extends both |
| `.planning/phases/03-smith-waterman-gotoh/03-CONTEXT.md` | Cross-validation evidence path (§1) and decomposition shape (§6) Phase 4 mirrors for Ratcliff-Obershelp |
| `.planning/phases/03-smith-waterman-gotoh/03-02-swg-cross-validation-SUMMARY.md` | Reference implementation of the Python-generator-plus-committed-corpus pattern Phase 4 replicates for Ratcliff-Obershelp |
| `algoid.go` (existing) | `AlgoStrcmp95` (slot 5), `AlgoLCSStr` (slot 8), `AlgoRatcliffObershelp` (last slot, 22 in 0-indexed) already declared |
| `jaro.go`, `jarowinkler.go` (existing) | Strcmp95 builds atop Jaro's match-flag arrays; Strcmp95 may either call jaroCore directly or replicate the match-flag loop. Planner picks. |
| `levenshtein.go`, `damerau_full.go`, `swg.go` (existing) | DP kernel patterns — LCSStr's DP differs but the file structure, godoc shape, header, and ASCII fast-path gate match |
| `props_test.go` (existing, extend-only) | Property test conventions — Phase 4 appends three blocks (Strcmp95 / LCSStr / Ratcliff-Obershelp) |
| `tests/bdd/steps/algorithms_steps.go` (existing, extend-only) | BDD step bindings — Phase 4 appends three sets |
| `testdata/golden/algorithms.json` (existing) | Phase 4 adds three sets of entries via the staging-merge pattern locked in Phase 2 |
</canonical_refs>

<code_context>
## Existing Code Insights

**Phase 1 + Phase 2 + Phase 3 outputs that Phase 4 must compose with:**

### Reusable assets

- `algoid.go` — `AlgoStrcmp95` (slot 5), `AlgoLCSStr` (slot 8),
  `AlgoRatcliffObershelp` (last slot) already declared. Phase 4 populates
  `dispatch[...]` via the var-init pattern in `dispatch_strcmp95.go`,
  `dispatch_lcsstr.go`, `dispatch_ratcliff_obershelp.go` (NO `init()`
  functions per §5.12).
- `errors.go` — `ErrInvalidInput`, `ErrInvalidConfiguration` available but
  not used: Phase 4 algorithm `*Score` functions handle every edge case
  deterministically (per §5.11).
- `normalise.go` / `tokenise.go` — not used directly by these algorithms;
  callers pre-normalise input. Strcmp95's godoc points consumers at
  `fuzzymatch.Normalise` for Unicode input (it's an ASCII-only algorithm).
- `golden_canonical.go` — `CanonicalMarshalForTest` is the only legal
  marshaller for golden files; staging files go through it via
  `testdata/golden/_staging/{strcmp95,lcsstr,ratcliff_obershelp}.json` and
  then plan 04-05 finalisation merges into the canonical
  `testdata/golden/algorithms.json`.
- `jaro.go` — Strcmp95 builds atop Jaro's algorithmic core. Planner decides
  whether Strcmp95 calls an internal Jaro helper (DRY but couples Strcmp95
  to Jaro's internals) or re-derives the match-flag arrays (more code, less
  coupling). Recommendation deferred to planner with code review.
- `cross_algorithm_consistency_test.go` (existing) — Phase 4 plan 04-05
  extends it with three new consistency tests:
  - Strcmp95 ≥ JaroWinkler (Strcmp95 only adds adjustments on top, never
    subtracts) on at least one input pair.
  - LCSStr ≥ Levenshtein on a substring-containment input (Levenshtein
    pays the deletion cost; LCSStr ignores it).
  - RatcliffObershelp pinned against difflib on at least one pair where
    it visibly differs from both Levenshtein and Jaro-Winkler.
- `bench.txt` (existing) — Phase 4 appends benchmark rows:
  - `BenchmarkStrcmp95Score_{ASCII_Short,Medium,Long}` (no Unicode_Short:
    Strcmp95 is byte-only by design)
  - `BenchmarkLCSStr{Score,LongestCommonSubstring}_{ASCII_Short,Medium,Long,Unicode_Short}`
  - `BenchmarkRatcliffObershelpScore_{ASCII_Short,Medium,Long,Unicode_Short}`
- `examples/identifier-similarity/main.go` (existing) — Phase 4 plan 04-05
  extends the table 7-row × 7-column → 7-row × 10-column (adds Strcmp95,
  LCSStr, Ratcliff-Obershelp columns; the LongestCommonSubstring string
  value is NOT included in the example table — score-only).
- `tests/bdd/steps/algorithms_steps.go` — extend-only append pattern;
  three new method blocks (per algorithm).
- `tests/bdd/features/` — three new feature files:
  `strcmp95.feature`, `lcsstr.feature`, `ratcliff_obershelp.feature`
  (one feature per algorithm; matches Phase 2/3 pattern).

### New cross-validation tooling (NOT runtime code)

- `scripts/gen-ratcliff-obershelp-cross-validation.py` — Python script using
  `difflib.SequenceMatcher(autojunk=False)` to produce reference vectors.
  Runs at developer-discretion via `make regen-ratcliff-obershelp-cross-validation`.
  Output: `testdata/cross-validation/ratcliff-obershelp/vectors.json`.
- `testdata/cross-validation/ratcliff-obershelp/vectors.json` — committed
  JSON corpus, ~15-18 entries covering all four mandatory categories
  (see <decisions> §1).
- `TestRatcliffObershelp_CrossValidation` in
  `ratcliff_obershelp_test.go` — reads the JSON, asserts
  `|our_score - difflib_score| <= 1e-9` for every entry. No Python at test time.
- Makefile: new `regen-ratcliff-obershelp-cross-validation` target
  (developer-only) + `TestRatcliffObershelp_CrossValidation` runs via
  default `go test ./...`.

### Established patterns (LOCKED — inherited from Phase 2 & 3)

- File-by-file structure: `<algo>.go` + `dispatch_<algo>.go` + `<algo>_test.go`
  + `<algo>_bench_test.go` + `<algo>_fuzz_test.go` + per-algo BDD feature
  + staging golden + on-disk fuzz seed.
- Apache-2.0 header on every `.go` file (`scripts/verify-license-headers.sh`).
- Stdlib testing only in root (testify allowed in `tests/bdd/`).
- AlgoID dispatch via `var _ = func() bool { dispatch[AlgoX] = XScore; return true }()`.
- Coverage floors: ≥ 95% overall, ≥ 90% per file, 100% on public API.
- llms.txt sync — every exported symbol listed; meta-test verifies.
- BDD score regex `(\d+\.?\d*)` accepts integer-form (per IN-03 cleanup).
- ASCII fast path: `if n <= maxStackInputLen && isASCII(a) && isASCII(b)` —
  applies to LCSStr's DP kernel (Strcmp95 and Ratcliff-Obershelp don't
  use DP buffers of this shape, so the gate doesn't apply directly).
</code_context>

<decisions>
## Implementation Decisions

### §1. Ratcliff-Obershelp cross-validation strategy — LOCKED

**Tool: Python `difflib.SequenceMatcher(autojunk=False)`** (stdlib — no
external Python deps required for the generator).

- PSF licence on stdlib — used for reference-vector cross-validation only,
  NOT for code copying, per `algorithm-licensing-standards`.
- `autojunk=False` is REQUIRED: difflib's default `autojunk=True` is a
  performance heuristic that drops "junk" characters when len(b) ≥ 200 and a
  character appears in ≥1% of positions. This is NOT the Ratcliff-Obershelp
  algorithm — it's a difflib speed optimisation. Disabling autojunk gives
  the true algorithm; the corpus MUST be generated with autojunk=False.

**Evidence location: committed JSON + generator script.**

- `scripts/gen-ratcliff-obershelp-cross-validation.py` — Python script,
  reads a fixed list of test cases (inline in the script), invokes
  `difflib.SequenceMatcher(autojunk=False, a=a, b=b).ratio()`, writes
  `testdata/cross-validation/ratcliff-obershelp/vectors.json`.
- Schema: `{"version": 1, "python_version": "<x.y.z>", "entries":
  [{"name": "<id>", "a": "<str>", "b": "<str>", "difflib_ratio": 0.X}]}`.
- 15–18 entries minimum. MUST cover all four mandatory categories:
  1. **Standard edge cases** — identity, both-empty, one-empty, no-overlap.
  2. **Dr. Dobb's 1988 paper examples** — canonical primary-source vectors
     (e.g. `"WIKIMEDIA"` / `"WIKIMANIA"`, `"GESTALT"` /
     `"GESTALT_PATTERN_MATCHING"`).
  3. **autojunk-sensitive case** — one synthetic 200+ char string that
     would trigger difflib's autojunk heuristic if it were enabled. Proves
     `autojunk=False` is correctly disabled in our cross-validation.
  4. **Substring + partial-match + unicode** — mid-similarity cases
     exercising the recursive longest-common-substring search; one
     multi-byte UTF-8 pair (e.g. `café` / `cafe`) for rune-path behaviour.

**Gate: default `go test ./...` includes `TestRatcliffObershelp_CrossValidation`.**

- Test reads the committed JSON, no Python at test time.
- Comparison tolerance: `|our_score - difflib_ratio| <= 1e-9` (matches
  Phase 3 epsilon convention).
- Regeneration is a separate Makefile target
  `make regen-ratcliff-obershelp-cross-validation` (developer-only); the
  verification target runs in `make check`.

**Python version assertion:** the generator script asserts
`sys.version_info >= (3, 7)` (insertion-order-preserving dicts for byte-stable
output) and pins the running `python_version` into the corpus's metadata.
Same pattern as Phase 3's biopython version guard (IN-07 closure).

### §2. Strcmp95 — full Winkler 1994 spec, ASCII-only byte path — LOCKED

**Public surface: one function.**

```go
// Strcmp95Score returns Winkler's Strcmp95 similarity between a and b
// in [0.0, 1.0]. Strcmp95 = Jaro + prefix boost + similar-character
// substitution credit + long-string (≥ 5 chars) adjustment + AS/I-S/RS-RB
// letter-pair similar-position adjustments per Winkler 1994.
//
// Strcmp95 operates on ASCII letters. For non-ASCII input, normalise
// to ASCII via fuzzymatch.Normalise first (NFC/NFD + diacritic folding).
// Non-letter and non-ASCII characters contribute Mismatch only via the
// underlying Jaro pass.
func Strcmp95Score(a, b string) float64
```

- **No `Strcmp95ScoreRunes` variant.** The similar-character table is
  letter-pair-keyed (AE, OU, etc.) — no Unicode equivalent exists in
  Winkler 1994. A rune-path variant would behave identically to the byte
  path for ASCII-letter inputs and ignore non-ASCII pairs, providing
  negligible semantic gain for added public surface. Godoc explicitly
  directs Unicode users to `Normalise` first.
- **No `Strcmp95Params`.** The four Winkler 1994 adjustments are part of
  the canonical algorithm, not consumer-tunable parameters. Phase 3's
  `SWGParams` was justified by genuine gap-penalty tuning needs; Strcmp95's
  adjustments are spec-defined. Phase 8's Scorer composition handles
  algorithm weight tuning at the composite level — not Strcmp95's surface.
- **All four adjustments stacked:** prefix boost (already in Jaro-Winkler),
  similar-character substitution credit, long-string (s ≥ 5 chars)
  adjustment, AS/I-S/RS-RB letter-pair similar-position adjustments. Matches
  Census Bureau strcmp95.c reference vectors byte-for-byte.

**Similar-character table sourcing:**

- Table values transcribed by hand from **Winkler 1994 TR-2** (the published
  paper) into a `var`-level table:
  ```go
  // strcmp95SimilarChars is the upper-case ASCII letter-pair similarity
  // table from Winkler 1994 TR-2 §3 "An improved string comparator".
  // Each entry encodes a bidirectional similarity (lookup symmetric in
  // (a, b) → (b, a)). Values are 0.3 across the published list.
  var strcmp95SimilarChars = [...]struct {
      a, b byte
      sim  float64
  }{
      {'A', 'E', 0.3}, {'A', 'I', 0.3}, /* ... 36 entries from the paper ... */
  }
  ```
- Census Bureau `strcmp95.c` consulted ONLY for canonical reference vectors
  in unit tests (the algorithm's expected output on a set of canonical
  surnames). NOT for table values — source is the Winkler 1994 paper.
- OpenRefine `Strcmp95.java` consulted ONLY for tie-breaking ambiguities
  in the prose of Winkler 1994 (e.g. ordering of adjustment application).
  No structural copying.
- Table is **unexported** and **not modifiable from outside the package**.
- Property test: call `Strcmp95Score` repeatedly with the same input and
  verify identical output (PITFALLS §14 closure — DeterministicAcrossRuns).
- `determinism-reviewer` agent flags any `init()` function in
  `strcmp95.go` as BLOCKING (the table MUST be a `var` literal).

### §3. LCSStr — formula SPEC-PINNED, 4 public functions, leftmost-tie-break — LOCKED

**Public surface: four functions (spec-pinned at `docs/requirements.md §7.1.9`).**

```go
func LongestCommonSubstring(a, b string) string
func LongestCommonSubstringRunes(a, b string) string
func LCSStrScore(a, b string) float64
func LCSStrScoreRunes(a, b string) float64
```

- The substring-returning variants are deliberate consumer-facing utility
  for the schema-similarity use case (callers want to know **what** the
  common segment is, not just its length).
- Score normalisation: `score = 2 · len(lcs) / (len(a) + len(b))`
  (Sørensen-Dice form). SPEC-PINNED — not a gray area.
- Edge cases (per spec): identical → 1.0; both empty → 1.0 (by convention);
  one empty → 0.0; no shared characters → 0.0.

**Tie-break for `LongestCommonSubstring` when multiple equal-length matches exist:**

- Return the **leftmost occurrence in `a`**. Deterministic.
- Document in godoc: "When multiple longest common substrings of equal
  length exist, the leftmost in a is returned. This is the standard
  textbook tie-break — the DP recurrence's natural left-to-right
  iteration."
- Property test: `TestProp_LongestCommonSubstring_LeftmostTieBreak` constructs
  inputs with deliberately-tied candidates and asserts the leftmost wins.

**DP:** two-row rolling buffer + max tracking. Recurrence per spec:
`D[i,j] = D[i-1,j-1] + 1 if a[i-1] = b[j-1], else 0`. Track max value and
ending position. Stack-allocated `[(maxStackInputLen+1)*2]int` buffer for
ASCII Short. ASCII fast-path gate identical to Phase 2/3 pattern.

**Allocation budget (matches Phase 2 PERF-01 contract):**
- 0 B/op, 0 allocs/op on ASCII Short (≤ 64 bytes).
- Heap path on ASCII Long (2 allocs for the two rolling rows).
- Rune path: 2 allocs for `[]rune(a)` and `[]rune(b)`; two more for the
  rolling rows; runtime alloc gate via `testing.AllocsPerRun`.

### §4. Ratcliff-Obershelp — recursive LCSubstr, no params, byte + rune — LOCKED

**Public surface: two functions.**

```go
func RatcliffObershelpScore(a, b string) float64
func RatcliffObershelpScoreRunes(a, b string) float64
```

- Standard surface (no Raw* variant — Ratcliff-Obershelp's score is
  already in [0, 1] by construction; no clamp logic to expose).
- No params (Ratcliff-Obershelp has no algorithm-level tunables).

**Algorithm (Ratcliff & Metzener 1988):**

1. Find the longest common substring of `a` and `b`.
2. Recursively apply step 1 to the left-of-substring portions and
   right-of-substring portions.
3. Sum the matched-character count across all recursion levels.
4. Score = `2 · matched / (len(a) + len(b))`.

**Implementation notes:**

- Recursion is bounded by string length — no stack overflow risk for
  reasonable inputs. Iterative stack-based traversal acceptable as long
  as it's byte-stable across platforms.
- The "longest common substring" inner step shares logic with LCSStr —
  the planner may either reuse `LCSStr`'s internal helper or implement
  the substring-only variant inline. The planner picks based on whether
  the abstraction earns its keep.
- **difflib parity:** `RatcliffObershelpScore` MUST match
  `difflib.SequenceMatcher(autojunk=False, a=a, b=b).ratio()` byte-for-byte
  within 1e-9 tolerance on the committed corpus. The autojunk=False
  qualifier is critical (see §1).

**Godoc directive (consumer hand-off from Phase 6 — PITFALLS §6):**

```go
// RatcliffObershelpScore is the difflib-equivalent. If you want
// "fuzzy string matching that behaves like Python's difflib.ratio()",
// use this.
//
// If you want the RapidFuzz "ratio()" semantics — the Indel formula
// 2·LCS/(|a|+|b|) used by Token Sort Ratio / Token Set Ratio /
// Partial Ratio — use those functions in Phase 6 instead.
```

This text fixes the chronic difflib-vs-Indel confusion that Pitfall #6
documents.

### §5. Property tests — INHERIT FROM PHASE 2 + ALGORITHM-SPECIFIC EXTENSIONS

**Standard Phase 2 invariants per algorithm:**

- `TestProp_<Algo>Score_RangeBounds` — `[0,1]` for all inputs
- `TestProp_<Algo>Score_Identity` — `Score(x, x) == 1.0` for non-empty x
  (Strcmp95 identity returns 1.0 by definition; LCSStr returns 1.0 because
  `2·n/(n+n) = 1`; Ratcliff-Obershelp returns 1.0 because the full string
  is the LCS, so `2·n/(n+n) = 1`)
- `TestProp_<Algo>Score_Symmetric` — `Score(a, b) == Score(b, a)` for byte path
- `TestProp_<Algo>ScoreRunes_Symmetric` — rune path (where rune variant exists;
  Strcmp95 has none)
- `TestProp_<Algo>Score_NoNaN`, `_NoInf`, `_NoNegativeZero` — DET-04

**No triangle inequality** for any of the three (none is a metric — all are
similarities derived from non-distance constructions).

**Algorithm-specific extensions:**

- **Strcmp95:** `TestProp_Strcmp95Score_AtLeastJaroWinkler` — Strcmp95 only
  adds adjustments on top of Jaro-Winkler; the result MUST satisfy
  `Strcmp95Score(a, b) >= JaroWinklerScore(a, b)` for inputs where the
  adjustments fire (or `==` where they don't).
- **Strcmp95:** `TestProp_Strcmp95Score_DeterministicAcrossRuns` — same input
  produces byte-identical output across 1000 sequential calls (Pitfall #14
  closure: confirms the `var`-level table is not mutated mid-process).
- **LCSStr:** `TestProp_LongestCommonSubstring_IsSubstringOfBoth` — the
  returned substring is a substring of both `a` and `b` (`strings.Contains`
  on byte path; rune-slice check on rune path).
- **LCSStr:** `TestProp_LongestCommonSubstring_LengthMatchesScore` — the
  returned substring's length matches the numerator in the score:
  `2 · len(LongestCommonSubstring(a, b)) / (len(a) + len(b)) == LCSStrScore(a, b)`
  within 1e-9.
- **LCSStr:** `TestProp_LongestCommonSubstring_LeftmostTieBreak` — for
  hand-curated inputs with deliberately-tied substring candidates, the
  returned substring is the leftmost.
- **Ratcliff-Obershelp:** `TestProp_RatcliffObershelpScore_AtLeastLevenshtein`
  — on substring-containment inputs, `RatcliffObershelpScore` is generally
  ≥ `LevenshteinScore` because RO finds the substring and ignores deletion
  costs. (Hand-curated rather than universal property — needs case
  filtering to avoid degenerate inputs.)

### §6. Plan decomposition — 5 sequential plans, no parallelism — LOCKED

**Structure: 5 plans, strict `depends_on` chain, no parallel waves.**

The shared-file dependencies (props_test.go, algoid_test.go, example_test.go,
cross_algorithm_consistency_test.go, llms.txt, golden, BDD steps) make
intra-phase parallelism a merge tax that dominates any speed gain.
Phase 2 ran 7 plans sequentially; Phase 3 ran 3 plans sequentially.
Phase 4 follows suit.

**Plan boundaries (planner refines):**

1. **`04-01-strcmp95`** — `strcmp95.go` + `dispatch_strcmp95.go` +
   `strcmp95_test.go` + `strcmp95_bench_test.go` + `strcmp95_fuzz_test.go`
   + Census Bureau reference vectors in unit tests + property tests +
   BDD `strcmp95.feature` + staging golden.

2. **`04-02-lcsstr`** — `lcsstr.go` + `dispatch_lcsstr.go` +
   `lcsstr_test.go` (with 4 public functions: LongestCommonSubstring +
   Runes + LCSStrScore + Runes) + bench + fuzz + property tests
   (including leftmost-tie-break and IsSubstringOfBoth) + BDD
   `lcsstr.feature` + staging golden.

3. **`04-03-ratcliff-obershelp`** — `ratcliff_obershelp.go` +
   `dispatch_ratcliff_obershelp.go` + `ratcliff_obershelp_test.go`
   (algorithm tests + Dr. Dobb's 1988 paper-vector pins; NO cross-validation
   yet — that's plan 04-04) + bench + fuzz + property tests + BDD
   `ratcliff_obershelp.feature` + staging golden. Godoc directive (§4)
   pointing consumers at the difflib-equivalent contract.

4. **`04-04-ratcliff-obershelp-cross-validation`** —
   `scripts/gen-ratcliff-obershelp-cross-validation.py` (Python stdlib only;
   no external deps; autojunk=False; version assertion) +
   `testdata/cross-validation/ratcliff-obershelp/vectors.json` (~15-18
   entries covering all four mandatory categories from §1) +
   `TestRatcliffObershelp_CrossValidation` appended to
   `ratcliff_obershelp_test.go` + Makefile target.

5. **`04-05-finalisation`** — merge `_staging/strcmp95.json`,
   `_staging/lcsstr.json`, `_staging/ratcliff_obershelp.json` into
   `testdata/golden/algorithms.json`. Extend `cross_algorithm_consistency_test.go`
   with three new tests (Strcmp95 ≥ JaroWinkler, LCSStr ≥ Levenshtein on
   substring containment, RatcliffObershelp pinned against difflib).
   Extend `examples/identifier-similarity/main.go` to add three columns
   (7 → 10 columns); update `want` constant and `TestExample_ColumnWidths`.
   Update `bench.txt` (full-replace via `make bench`).
   Verify `llms.txt` SWG entries are extended for the three new algorithms
   + their public surfaces (Strcmp95: 1 func; LCSStr: 4 funcs;
   Ratcliff-Obershelp: 2 funcs = 7 total new entries).
   No update to `docs/requirements.md` §7.1.7/§7.1.9/§7.1.24 needed —
   the existing spec language is honoured verbatim.

### §7. Inherited Phase 2 + 3 patterns (NOT decided here — read prior CONTEXT.md files)

| Pattern | Source | Applies to Phase 4 |
|---------|--------|-------------------|
| AlgoID dispatch via `var _ = func() bool {...}()` (no init() functions) | 02-CONTEXT.md §dispatch | `dispatch_{strcmp95,lcsstr,ratcliff_obershelp}.go` |
| `maxStackInputLen = 64` + `isASCII(a) && isASCII(b)` gate | 02-CONTEXT.md §performance, 02-PATTERNS.md Pattern 3 | LCSStr stack-buffer path. Strcmp95 has no DP buffer; Ratcliff-Obershelp recursion uses no large stack buffer. |
| Per-algorithm staging golden → canonical merge in finalisation plan | 02-CONTEXT.md §golden | `testdata/golden/_staging/{strcmp95,lcsstr,ratcliff_obershelp}.json` → `algorithms.json` |
| Per-algorithm BDD feature file | 02-CONTEXT.md §BDD | `tests/bdd/features/{strcmp95,lcsstr,ratcliff_obershelp}.feature` |
| Per-algorithm fuzz harness with seed-001 entry | 02-CONTEXT.md §fuzz | `*_fuzz_test.go` + `testdata/fuzz/Fuzz*/seed-001` (with 03 IN-06 lesson: don't drift formatting — gofmt enforces) |
| props_test.go / example_test.go / algoid_test.go / cross_algorithm_consistency_test.go / tests/bdd/steps/algorithms_steps.go are extend-only append points | 02-CONTEXT.md §shared-files | Three appends for Phase 4 |
| Stdlib testing only in root (no testify; testify allowed in tests/bdd) | 02-CONTEXT.md §test-stack | All `*_test.go` files use stdlib `testing` only |
| Identity short-circuit `if a == b { return 1.0 }` on `*Runes` entries | 03 IN-04 cleanup | `LCSStrScoreRunes`, `RatcliffObershelpScoreRunes` |
| BDD score regex `(\d+\.?\d*)` accepts integer-form | 03 IN-03 cleanup | Phase 4 feature files can use `0` and `1` |
| Cross-platform CI matrix verifies byte-identical output | Phase 1 verification | `testdata/golden/algorithms.json` Phase 4 entries must be byte-stable |
| Coverage floors ≥ 95% overall, ≥ 90% per file | go-testing-standards skill | All Phase 4 source files |
| llms.txt is the AI-friendly catalogue; meta-test asserts every exported symbol is listed | Phase 1 + Phase 2 + Phase 3 IN-01 lesson | Add 7 entries (Strcmp95Score + LongestCommonSubstring + Runes + LCSStrScore + Runes + RatcliffObershelpScore + Runes) |
| Python version assertion in generator scripts | 03 IN-07 closure | `gen-ratcliff-obershelp-cross-validation.py` asserts `sys.version_info >= (3, 7)` |
| Full surface exercised by fuzz harness, not just default-params byte path | 03 WR-02 closure | LCSStr fuzz exercises all 4 functions; Ratcliff-Obershelp fuzz exercises both functions; Strcmp95 has only 1 function so this is automatic |
| Numerical regression pin alongside cross-validation corpus, not just smoke gate | 03 WR-03 closure | Ratcliff-Obershelp unit test pins one exact value from Dr. Dobb's 1988 in `ratcliff_obershelp_test.go` directly, NOT solely via the cross-validation corpus |
| Test redirecting os.Stdout uses defer-restore | 03 WR-04 closure | No new example tests in Phase 4 (the identifier-similarity example test was set up in Phase 2 and updated in Phase 3); finalisation plan only updates the `want` constant |

### Claude's Discretion

The planner may decide:

- Whether Strcmp95 calls an internal Jaro helper or re-derives the
  match-flag arrays. Either is acceptable provided the Census Bureau
  reference vectors pass. Trade-off is DRY vs algorithm independence.
- Whether Ratcliff-Obershelp's recursion uses the language-native call
  stack or an explicit stack-based iterative implementation. Either is
  acceptable provided byte-stable output across platforms.
- Whether Ratcliff-Obershelp's "find longest common substring" inner step
  reuses LCSStr's internal helper or implements the substring search
  inline. Either is acceptable provided correctness.
- The exact bench label conventions (matching Phase 2/3 prefix-numbering).
- Whether to add a `ratcliff_obershelp.go` or `ratcliffobershelp.go`
  filename (Phase 3 used the underscore form `swg.go` — short; Ratcliff
  is wordier, so an underscore-separated filename is fine and may be
  more readable than `ratcliffobershelp.go`).

</decisions>

<specifics>
## Specific Ideas

- **difflib's autojunk=False is the gate.** This is the single most important
  detail in the entire phase. Any drift in autojunk handling silently changes
  every score in the corpus. The Python script MUST explicitly pass
  `autojunk=False`, the corpus MUST include the 200+-char autojunk-sensitive
  test case to prove it, and the godoc on `RatcliffObershelpScore` MUST
  state "behaves like difflib.SequenceMatcher(autojunk=False, a=a, b=b).ratio()".

- **The Strcmp95-vs-Jaro-Winkler-vs-Jaro hierarchy.** A consumer reading
  the package docs should immediately understand:
  - `JaroScore` — the base similarity.
  - `JaroWinklerScore` — Jaro + prefix boost (for shared-prefix bias).
  - `Strcmp95Score` — Jaro-Winkler + similar-character credit + adjustments
    (for record-linkage / surname matching).
  The Strcmp95 godoc opens with this layering so readers don't ask "why
  do we have three similar algorithms?"

- **LCSStr's `LongestCommonSubstring` returning a string is non-standard.**
  Most similarity-library APIs return only a score. Exposing the substring
  is justified by the schema-similarity use case ("which substring is
  driving the match?"). Document this as a v1.0 commitment; future
  versions cannot remove the function without a major version bump.

- **Ratcliff-Obershelp is THE difflib-equivalent.** This is so important
  the godoc opens with "If you want fuzzy string matching that behaves
  like Python's difflib.ratio(), use this." All downstream documentation
  (FAQ.md, scorer.md, docs/algorithms.md) should reinforce this. Phase 6's
  TokenSortRatio / TokenSetRatio / PartialRatio godoc explicitly directs
  users wanting difflib semantics here.

</specifics>

<deferred>
## Deferred Ideas

Items acknowledged and carried forward, OUT of scope for Phase 4:

- **EMBOSS `water` second cross-validation source for Strcmp95.** Census
  Bureau strcmp95.c is the canonical reference; EMBOSS is for sequence
  alignment, not Strcmp95. Not relevant for Phase 4; mentioning here for
  completeness — no action.

- **A `Strcmp95Params` API.** Not in scope for v1.0 (per §2). If a
  consumer-driven use case for tuning Strcmp95's adjustments surfaces
  post-v1.0, a `Strcmp95Params` struct mirroring `SWGParams` is the path —
  but the canonical algorithm has no parameters.

- **`LongestCommonSubstrings()` returning all tied-longest matches.** Not
  in scope for v1.0; the spec commits to one substring. Future v1.x scope
  question if consumers ask for it.

- **CI installation of Python 3 for re-verification of the Ratcliff-Obershelp
  corpus.** The committed JSON is the verification fixture; regeneration is
  developer-discretion. If we later want CI to re-verify the JSON-vs-difflib
  agreement (e.g. after a Python stdlib update), add a GitHub Actions
  workflow then. (Note: Python is already in CI runners by default; no
  install step needed if we choose to add re-verification later.)

- **Public API freeze for `LongestCommonSubstring` tie-break semantics.**
  Leftmost-in-`a` is documented and property-tested; changing it post-v1.0
  is a breaking change.

</deferred>

<follow_ups>
## Phase 5+ tracking items surfaced during Phase 4 discussion

- **Phase 6's TokenSortRatio / TokenSetRatio / PartialRatio godoc** must
  point users wanting `difflib.ratio()` semantics at `RatcliffObershelpScore`.
  Phase 4 ships the algorithm; Phase 6 ships the cross-reference. The
  `bdd-scenario-reviewer` should verify this hand-off when Phase 6 lands.

- **api-ergonomics-reviewer should flag the LCSStr 4-function surface
  during Phase 4 PR review** — exposing the underlying substring (not just
  the score) is intentional but unusual. Confirm naming symmetry between
  `LongestCommonSubstring`/`LongestCommonSubstringRunes` and
  `LCSStrScore`/`LCSStrScoreRunes`.

- **algorithm-correctness-reviewer gate** for Ratcliff-Obershelp:
  cross-validation JSON must be present + green before the impl can land.
  Plan-phase locks this into the plan check (same as Phase 3 SWG).

- **The Strcmp95 ≥ JaroWinkler property test** (§5) is the algorithm-level
  hierarchy test. If the property fails for any input, that's a hard
  signal Strcmp95 is computing adjustments wrong. Worth promoting to a
  pre-finalisation gate.

</follow_ups>

<carry_forward>
## From Phase 2 + 3 (inherited automatically — no need to re-decide)

| Decision | Source | Applies to Phase 4 |
|----------|--------|-------------------|
| AlgoID dispatch via `var _ = func() bool {...}()` (no init() functions) | 02-CONTEXT.md §dispatch | All three dispatch_*.go files |
| `maxStackInputLen = 64` + `isASCII(a) && isASCII(b)` gate | 02-CONTEXT.md §performance, 02-PATTERNS.md Pattern 3 (locked) | `lcsstr.go` stack-buffer path |
| Per-algorithm staging golden → canonical merge in finalisation plan | 02-CONTEXT.md §golden | `testdata/golden/_staging/{strcmp95,lcsstr,ratcliff_obershelp}.json` |
| Per-algorithm BDD feature file | 02-CONTEXT.md §BDD | `tests/bdd/features/*.feature` (three new) |
| Per-algorithm fuzz harness with seed-001 entry | 02-CONTEXT.md §fuzz | `*_fuzz_test.go` (three new) + on-disk seed |
| props_test.go / example_test.go / algoid_test.go / cross_algorithm_consistency_test.go / tests/bdd/steps/algorithms_steps.go are extend-only append points | 02-CONTEXT.md §shared-files | Three appends per file in Phase 4 |
| Stdlib testing only in root | 02-CONTEXT.md §test-stack | All `*_test.go` files |
| Identity short-circuit `if a == b { return 1.0 }` on `*Runes` entries | 03 IN-04 closure | `LCSStrScoreRunes`, `RatcliffObershelpScoreRunes` |
| BDD score regex `(\d+\.?\d*)` accepts integer-form | 03 IN-03 closure | Phase 4 feature files |
| Cross-platform CI matrix verifies byte-identical output (DET-02 lock) | Phase 1 verification | `algorithms.json` Phase 4 entries |
| Coverage floors ≥ 95% overall, ≥ 90% per file | go-testing-standards skill | All Phase 4 files |
| llms.txt sync — meta-test asserts every exported symbol listed | Phase 1 + 2 + 3 | 7 new entries |
| Cross-validation: committed JSON + Python script + Makefile regen target | Phase 3 SWG pattern | Plan 04-04 replicates the structure for difflib |
| Python version assertion in generator scripts | 03 IN-07 closure | gen-ratcliff-obershelp-cross-validation.py asserts Python >= 3.7 |
| Fuzz harness exercises full public surface | 03 WR-02 closure | LCSStr fuzz body exercises all 4 functions; Ratcliff-Obershelp fuzz exercises both |
| Numerical regression pins alongside cross-validation corpus | 03 WR-03 closure | Ratcliff-Obershelp unit tests pin Dr. Dobb's 1988 vectors directly, not solely via corpus |
| Cross-validation evidence is the algorithm-correctness-reviewer gate | 03 SWG pattern | Plan 04-04 must land green before plan 04-05 finalisation merges goldens |

</carry_forward>

---

*Phase: 04-remaining-character-gestalt*
*Context gathered: 2026-05-14*
