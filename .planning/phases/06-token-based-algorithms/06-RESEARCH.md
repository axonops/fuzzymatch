# Phase 6: Token-based Algorithms - Research

**Researched:** 2026-05-15
**Domain:** Token-based string-similarity algorithms (LCS-Indel-formula ratios + per-token-max composition + token-set Jaccard)
**Confidence:** HIGH

## Summary

Phase 6 ships five token-based similarity algorithms — Monge-Elkan, Token Sort Ratio, Token Set Ratio, Partial Ratio, and Token Jaccard — plus a shared `token_indel.go` infrastructure file holding the unexported `lcsLen` + `indelRatio` kernel that the three Indel-based ratios consume. CONTEXT.md has already locked all major decisions (cross-validation source = RapidFuzz; foundation file mirrors Phase 5's `q_gram.go`; Monge-Elkan inner-metric strict allow-list; symmetric variant in dispatch; DoS-vector godoc format). This research verifies those locked decisions against primary sources, documents the precise LCS DP recurrence and Indel-formula equivalence, captures the *actual* RapidFuzz reference implementation (Python source code, MIT licensed) for TokenSort/TokenSet/PartialRatio so the planner has byte-faithful behavioural specifications, and surfaces three high-impact landmines that the planner must address: (1) RapidFuzz tokenises via Python `str.split()` (whitespace-only, no lowercasing) — divergent from fuzzymatch's `Tokenise` which is camelCase/snake/kebab/dot-aware AND lowercasing; (2) RapidFuzz's `partial_ratio` is NOT a naive sliding-window — it has a three-segment iteration pattern (left tail, middle full-window, right tail) plus a `s1_char_set` early-skip optimisation; (3) RapidFuzz's `token_set_ratio` returns 0.0 (not 1.0) when either token-set is empty, mirroring fuzzywuzzy bug-for-bug per RapidFuzz issue #110.

**Primary recommendation:** Plan exactly the structure CONTEXT.md prescribes (foundation file + 5 algorithms + finalisation), but the planner MUST resolve the tokeniser-divergence question before plan 06-01: either (a) the cross-validation generator pre-tokenises with `str.split()`-equivalent input on both sides so RapidFuzz sees pre-joined sorted tokens (the "tokens-prepared upstream" path — recommended); or (b) document that fuzzymatch's TokenSort/TokenSet/PartialRatio adopt a *different* tokenisation (the project's `Tokenise` with camelCase splits) and accept that cross-validation runs on a controlled corpus where the tokeniser divergence doesn't affect output (i.e. inputs with whitespace boundaries only). Path (a) is structurally simpler and matches the RapidFuzz contract literally — it requires the cross-validation script to call `rapidfuzz.fuzz.token_sort_ratio(a, b)` directly, and the Go test asserts our algorithm produces the same score when run on the same `(a, b)` inputs WHERE both inputs use whitespace-only word boundaries. Path (b) requires writing custom Python that mimics the project's Tokenise rules — fragile and a documentation burden.

## Architectural Responsibility Map

| Capability | Primary Tier | Secondary Tier | Rationale |
|------------|-------------|----------------|-----------|
| LCS-subsequence DP (`lcsLen`) | Algorithm internals (`token_indel.go`) | — | Foundation kernel; mirrors `q_gram.go`'s extractor role for the q-gram tier |
| Indel-formula normalisation (`indelRatio`) | Algorithm internals (`token_indel.go`) | — | Single-source kernel consumed by TokenSort, TokenSet, PartialRatio |
| Tokenisation | Library primitive (`tokenise.go` from Phase 1) | — | Already implemented; consumed verbatim |
| Token-Sort sorted-join | Algorithm wrapper (`token_sort_ratio.go`) | — | Calls `Tokenise`, sorts, joins, calls `indelRatio` |
| Token-Set three-way max | Algorithm wrapper (`token_set_ratio.go`) | — | Calls `Tokenise`, set-difference, three `indelRatio` calls, max |
| Sliding-window over longer string | Algorithm wrapper (`partial_ratio.go`) | — | Bytes path AND runes path, both calling `indelRatio` / `indelRatioRunes` |
| Per-token-max average (Monge-Elkan) | Algorithm wrapper (`monge_elkan.go`) | AlgoID dispatch table (Phase 1 reserved slot) | Calls inner-metric via dispatch lookup |
| Inner-metric allow-list validation | Algorithm wrapper (`monge_elkan.go`) | Phase 8 Scorer (later) | Direct-call panics; Phase 8 Scorer will return `ErrInvalidAlgoID` |
| AlgoID dispatch wiring | Per-algorithm `dispatch_<algo>.go` files | Phase 1 dispatch table | Pattern established Phase 2; slots already reserved (algoid.go lines 129-154) |
| Cross-validation corpus generation | Developer toolchain (`scripts/`) | Python+rapidfuzz | NOT a runtime dep; pinned-version script + committed JSON corpus |
| Cross-validation assertion | Test (`*_cross_validation_test.go`) | — | Loads JSON, asserts byte-stable score match within ε=1e-9 |

## Standard Stack

### Core (project-internal, already in repo)

| Component | Path | Purpose | Why Standard |
|-----------|------|---------|--------------|
| `Tokenise(s, opts)` | `tokenise.go` (Phase 1) | Returns `[]string` of camelCase/snake/kebab/dot-split tokens, lowercased per `DefaultTokeniseOptions{}` | Single source of truth for the project's tokenisation contract — Monge-Elkan / TokenJaccard / TokenSortRatio / TokenSetRatio all consume this |
| `AlgoID` dispatch table | `algoid.go` (Phase 1) | Array-backed `[numAlgorithms]func(a,b string) float64` | Zero-alloc dispatch; Monge-Elkan reads from it for inner-metric resolution |
| `extractQGrams` pattern | `q_gram.go` (Phase 5) | Unexported helpers + `export_test.go` re-export | Direct template for `token_indel.go` per CONTEXT §2 LOCKED |
| `dispatch[AlgoX] = ...` idiom | `dispatch_<algo>.go` files (Phase 2 onwards) | Package-level side-effect via `var _ = func() bool { dispatch[AlgoX] = XScore; return true }()` | Avoids `init()` per DET-13 |
| `props_test.go` accumulator | `props_test.go` (Phase 2 onwards) | Single file accumulating per-algorithm property tests | Each phase appends; Phase 6 adds ~5×6=30 new property tests |
| `algorithms.json` golden | `testdata/golden/` | Pinned scores for cross-platform determinism gate | Phase 6 finalisation extends with 5 new algorithms |
| `_staging/<algo>.json` | `testdata/golden/_staging/` | Per-plan staging golden files merged in finalisation | Pattern proven over 14 algorithm staging files |
| `<algo>.feature` BDD | `tests/bdd/features/` | Godog scenarios per algorithm | Pattern proven over 14 algorithms |
| `tests/bdd/steps/algorithms_steps.go` | tests/bdd/steps/ | Step-registration accumulator | Each Phase 6 algorithm appends a `step.Step` registration |
| `example_test.go` | root | Runnable godoc examples | One per public function; Phase 6 adds 7 (MongeElkanScore, MongeElkanScoreSymmetric, TokenSortRatioScore, TokenSetRatioScore, PartialRatioScore, PartialRatioScoreRunes, TokenJaccardScore) |
| `llms.txt` / `llms-full.txt` | root | AI-friendly doc index | Per-plan sync (NOT deferred) — caught Phase 4 mid-flight, reinforced Phase 5 |
| `bench.txt` | root | benchstat regression baseline | Phase 6 finalisation full-replaces with new bench numbers |

### Cross-validation toolchain (developer-only, NOT runtime)

| Component | Version | Purpose | Why Recommended |
|-----------|---------|---------|-----------------|
| **`rapidfuzz` (PyPI)** | **3.14.5** (released 2026-04-07; verify with `pip index versions rapidfuzz` at planning time) | Python reference implementation of TokenSort/TokenSet/PartialRatio for cross-validation corpus | Spec mandates RapidFuzz cross-validation per CONTEXT §1 LOCKED. MIT-licensed. Documents the exact Indel formula `1 - distance/(len1+len2)` which is mathematically equivalent to the spec's `2·LCS/(len1+len2)` (proof in §3 below). [VERIFIED: PyPI registry, https://pypi.org/pypi/rapidfuzz/json — version 3.14.5 confirmed 2026-04-07] |
| **Python 3.7+** | 3.12 (matches Phase 4 RO precedent) | Generator script runtime | Stdlib `json` with `sort_keys=False` requires 3.7+ for dict-insertion-order stability — same gate as Phase 4 RO generator |

**Installation (developer machine only):**
```bash
python3 -m pip install --user rapidfuzz==3.14.5
```

**Version verification before script runs:** the generator script header MUST declare `RAPIDFUZZ_VERSION = "3.14.5"` and assert `rapidfuzz.__version__ == RAPIDFUZZ_VERSION`. Mirrors Phase 3's biopython-version assertion and Phase 4's Python-version assertion. [CITED: scripts/gen-ratcliff-obershelp-cross-validation.py (Phase 4 precedent)]

### Alternatives Considered (and rejected — see CONTEXT.md)

| Instead of | Could Use | Why Rejected (per CONTEXT) |
|------------|-----------|----------------------------|
| RapidFuzz Python | `fuzzywuzzy` Python | fuzzywuzzy's pure-Python and C-extension paths produce DIFFERENT scores for the same library call — non-canonical. Spec mandates RapidFuzz |
| RapidFuzz Python | Hand-derived vectors only | Rejected for Indel-based ratios because corpus needs 20-30 vectors per algorithm to exercise three-way max + sliding window; reviewers cannot hand-derive 20+ LCS computations |
| RapidFuzz Python | Vendored RapidFuzz test vectors verbatim | Rejected — freezes corpus to a snapshot that can't be extended with project-specific edge cases |
| RapidFuzz Python (for Monge-Elkan) | RapidFuzz `process.fuzz` ME variant | Rejected — RapidFuzz's ME default inner may not match ours; would conflate "ME formula correctness" with "inner-metric fidelity" |
| Levenshtein-based Indel variant | LCS-based Indel (spec) | Rejected — spec explicitly mandates `2·LCS/(|a|+|b|)`. Levenshtein-based available via `1 - LevenshteinScore` composition for consumers |

## Architecture Patterns

### System Architecture Diagram

```
                     ┌─────────────────────────────────────────────────┐
                     │         Phase 6 Public API Surface (7 fns)       │
                     └─────────────────────────────────────────────────┘
                                            │
        ┌──────────────────┬────────────────┼─────────────────┬──────────────────┐
        │                  │                │                 │                  │
        ▼                  ▼                ▼                 ▼                  ▼
  MongeElkanScore   TokenSortRatio   TokenSetRatio   PartialRatioScore    TokenJaccardScore
  MongeElkan        Score            Score           PartialRatioScore-
  ScoreSymmetric                                     Runes
        │                  │                │                 │                  │
        │                  │                │                 │                  │
        ▼                  ▼                ▼                 ▼                  ▼
   Tokenise            Tokenise         Tokenise          (no Tokenise         Tokenise
   (Phase 1)           (Phase 1)        (Phase 1)         — char-level         (Phase 1)
        │                  │                │              sliding window)         │
        │                  │                │                 │                  │
        │                  │  sorted-join   │  set-diff        │  identity-        │  set-Jaccard
        │                  │  +join " "     │  3-substr build │  short-circuit     │  via
        │                  │                │                 │                    │  intersect/union
        │                  │                │                 │                    │  cardinalities
        ▼                  ▼                ▼                 ▼                  ▼
  AlgoID dispatch   token_indel.go   token_indel.go   token_indel.go     (no kernel —
  table lookup           │                │                 │              direct count
  for inner metric       │                │                 │              over sorted
        │                ▼                ▼                 ▼              key sets)
        │           lcsLen + indelRatio (shared kernel)
        │
        ▼
  permittedMongeElkanInner allow-list validation
  → calls dispatch[inner](tokenA, tokenB) per (i, j)
  → max over j; average over i
  → ME(A,B) [+ ME(B,A) for symmetric variant]
```

**File layout (planner: 12 new `.go` files + 5 BDD features + 5 staging-golden files + 1 cross-validation script + 1 corpus JSON):**

```
fuzzymatch/
├── token_indel.go                       # Foundation kernel (lcsLen, indelRatio, lcsLenRunes, indelRatioRunes)
├── token_indel_test.go                  # Kernel tests via export_test.go re-exports
├── monge_elkan.go                       # MongeElkanScore + MongeElkanScoreSymmetric + permittedMongeElkanInner map
├── monge_elkan_test.go
├── monge_elkan_bench_test.go            # ASCII Short/Medium/Long + Unicode + Pathological_1000Tokens
├── monge_elkan_fuzz_test.go             # FuzzMongeElkanScore (asymmetric) + FuzzMongeElkanScoreSymmetric
├── dispatch_monge_elkan.go              # Wires AlgoMongeElkan → MongeElkanScoreSymmetric(a, b, AlgoJaroWinkler, DefaultNormalisationOptions)
├── token_sort_ratio.go                  # TokenSortRatioScore
├── token_sort_ratio_test.go
├── token_sort_ratio_bench_test.go       # Standard benches
├── token_sort_ratio_fuzz_test.go        # FuzzTokenSortRatioScore
├── dispatch_token_sort_ratio.go
├── token_set_ratio.go                   # TokenSetRatioScore (three-way max)
├── token_set_ratio_test.go
├── token_set_ratio_bench_test.go        # + Pathological_AsymmetricSetCardinalities
├── token_set_ratio_fuzz_test.go
├── dispatch_token_set_ratio.go
├── partial_ratio.go                     # PartialRatioScore + PartialRatioScoreRunes
├── partial_ratio_test.go
├── partial_ratio_bench_test.go          # + Pathological_LongShortMismatch (both byte AND rune surfaces)
├── partial_ratio_fuzz_test.go
├── dispatch_partial_ratio.go            # Wires byte-path PartialRatioScore (signature is (a, b string) float64)
├── token_jaccard.go                     # TokenJaccardScore (set-Jaccard over Tokenise output)
├── token_jaccard_test.go
├── token_jaccard_bench_test.go
├── token_jaccard_fuzz_test.go
├── dispatch_token_jaccard.go
├── token_ratio_cross_validation_test.go # OR per-algorithm cross-validation files (planner choice)
├── tests/bdd/features/
│   ├── monge_elkan.feature
│   ├── token_sort_ratio.feature
│   ├── token_set_ratio.feature
│   ├── partial_ratio.feature           # Covers BOTH byte and rune surfaces
│   └── token_jaccard.feature
├── testdata/cross-validation/token-ratios/
│   └── vectors.json                     # Generated by gen-token-ratio-cross-validation.py; ~80-120 entries
├── testdata/golden/_staging/
│   ├── monge_elkan.json
│   ├── token_sort_ratio.json
│   ├── token_set_ratio.json
│   ├── partial_ratio.json
│   └── token_jaccard.json
└── scripts/
    └── gen-token-ratio-cross-validation.py
```

### Pattern 1: Foundation kernel mirroring Phase 5 q_gram.go

**What:** Unexported helpers in `token_indel.go` re-exported via `export_test.go` for kernel-level testing. Consumers (TokenSort, TokenSet, PartialRatio) test composition only — never duplicate kernel-level cases.

**When to use:** Whenever multiple sibling algorithms share a core computation. The Phase 5 `q_gram.go`/`extractQGrams` pattern is the direct template per CONTEXT §2 LOCKED.

**Example (planner specification — exact signatures are Claude's discretion per CONTEXT):**
```go
// In token_indel.go (unexported):
func lcsLen(a, b []byte) int             // LCS-subsequence length, two-row DP
func indelRatio(a, b []byte) float64     // 2·lcsLen(a,b) / (len(a)+len(b))
func lcsLenRunes(a, b []rune) int
func indelRatioRunes(a, b []rune) float64

// In export_test.go (test-only re-export):
var LCSLenForTest = lcsLen
var IndelRatioForTest = indelRatio
var LCSLenRunesForTest = lcsLenRunes
var IndelRatioRunesForTest = indelRatioRunes
```

[CITED: q_gram.go (Phase 5) lines 88-149 — the precedent for unexported-helper-with-export_test-re-export]

### Pattern 2: LCS-subsequence DP (Wagner-Fischer 1974)

**What:** Standard O(mn) two-row DP for LCS-subsequence length. NOT the same as `lcsstr.go`'s LCS-substring DP (which resets to 0 on mismatch).

**Recurrence (Wagner & Fischer 1974, J. ACM 21(1):168-173):**
```
LCS[0, j] = 0       (boundary)
LCS[i, 0] = 0       (boundary)
LCS[i, j] = LCS[i-1, j-1] + 1                  if a[i-1] == b[j-1]
            max(LCS[i-1, j], LCS[i, j-1])      otherwise
```

**Two-row optimisation (per Phase 2/3/4 discipline):**
```go
// Use min(m, n) as the row length: always make the inner loop over
// the shorter string to minimise space.
prev := make([]int, n+1)  // or stack-allocated [65]int when n <= 64
curr := make([]int, n+1)
for i := 1; i <= m; i++ {
    for j := 1; j <= n; j++ {
        if a[i-1] == b[j-1] {
            curr[j] = prev[j-1] + 1
        } else {
            // explicit if/else, NOT builtin max() — project canonical
            // pattern for determinism-reviewer auditability (per
            // qgram_jaccard.go lines 240-246)
            if prev[j] > curr[j-1] {
                curr[j] = prev[j]
            } else {
                curr[j] = curr[j-1]
            }
        }
    }
    prev, curr = curr, prev
}
return prev[n]
```

**Stack buffer threshold:** `[64]int` array used when `min(m, n) <= 50`. CONTEXT §2 sets the buffer at `[64]int` but allows planner to tune via benchstat; the stack-allocated buffer is the load-bearing optimisation for PartialRatio's sliding-window which calls `lcsLen` repeatedly. [CITED: performance-standards SKILL.md "Stack-Allocated Buffer" pattern]

**Edge cases:**
- Both empty: `lcsLen("", "") = 0`. The wrapper `indelRatio` MUST treat `len(a)+len(b) == 0` → return `1.0` (both-empty identity convention per algorithm-correctness-standards).
- One empty: `lcsLen(s, "") = 0`. `indelRatio` returns `2·0/(len(s)+0) = 0.0` — matches the "one-empty → 0.0" convention.
- Identical: `lcsLen(s, s) = len(s)`. `indelRatio` returns `2·len(s)/(2·len(s)) = 1.0` — matches identity.

[VERIFIED: Wagner & Fischer 1974, J. ACM 21(1):168-173 — the canonical DP recurrence; cross-checked at https://en.wikipedia.org/wiki/Longest_common_subsequence]

### Pattern 3: Indel formula equivalence

**Claim:** RapidFuzz's `normalized_similarity = 1 - distance/(len(s1)+len(s2))` (where Indel distance counts insertions and deletions only — substitutions cost 2) is mathematically identical to the spec's `2·LCS/(|a|+|b|)`.

**Proof (no hand-waving):**
```
indel_distance(a, b) = (|a| - LCS) + (|b| - LCS)    -- the only edits available are
                    = |a| + |b| - 2·LCS              -- ins (gives a→a chars not in LCS)
                                                     -- and del (drops b chars not in LCS)

normalized_similarity = 1 - distance/(|a|+|b|)
                      = (|a|+|b| - distance)/(|a|+|b|)
                      = (|a|+|b| - (|a|+|b| - 2·LCS))/(|a|+|b|)
                      = 2·LCS/(|a|+|b|)  ∎
```

[VERIFIED: derived from RapidFuzz Indel docs (https://rapidfuzz.github.io/RapidFuzz/Usage/distance/Indel.html) + standard LCS-Indel relation cross-checked at https://en.wikipedia.org/wiki/Longest_common_subsequence]

**Edge case handled by RapidFuzz's `_norm_distance`:** when `lensum == 0`, returns 100 (i.e. `1.0` after divide-by-100). [VERIFIED: rapidfuzz/fuzz_py.py line 28-30, https://github.com/rapidfuzz/RapidFuzz/blob/main/src/rapidfuzz/fuzz_py.py]

### Pattern 4: TokenSortRatio composition

**RapidFuzz reference implementation (MIT, copy-LLM-friendly for understanding only — fresh implementation per project rules):**
```python
# rapidfuzz/fuzz_py.py: token_sort_ratio
def token_sort_ratio(s1, s2, *, processor=None, score_cutoff=None):
    sorted_s1 = " ".join(sorted(s1.split()))   # Python str.split() = whitespace-only, no empty
    sorted_s2 = " ".join(sorted(s2.split()))
    return ratio(sorted_s1, sorted_s2)         # ratio = Indel-based similarity
```

**Critical observation — RapidFuzz tokenisation:**
- Uses Python `str.split()` — whitespace-only boundaries (`" "`, `"\t"`, `"\n"` etc.); NO punctuation splitting.
- NO lowercasing (default `processor=None` since RapidFuzz 3.0).
- Empty string produces empty token list `[]` — `_join_splitted_sequence([])` returns `""` (handled in helper).

**fuzzymatch divergence:** the project's `Tokenise(s, DefaultTokeniseOptions())` uses camelCase / snake_case / kebab-case / dot-case splits AND lowercases. This is a documented project choice (Phase 1 FOUND-04). For TokenSortRatio:
1. The cross-validation corpus MUST consist of inputs where fuzzymatch's `Tokenise` output and Python `str.split()` output produce IDENTICAL token lists. This means: pure whitespace-separated lowercase ASCII identifiers (e.g. `"alpha beta gamma"`, `"the quick brown fox"`).
2. Any input with camelCase, snake_case, hyphens, or mixed case will produce DIFFERENT scores between the project's TokenSortRatio and RapidFuzz's `token_sort_ratio` — the divergence is by design but MUST NOT appear in the cross-validation corpus.
3. The algorithm's godoc MUST document this divergence: "Uses fuzzymatch's `Tokenise` (camelCase-aware), NOT Python `str.split()`. For inputs without identifier-style boundaries, behaviour matches RapidFuzz; for inputs with mixed identifier styles, the project's tokenisation produces semantically richer splits."

[VERIFIED: rapidfuzz/fuzz_py.py lines 396-399, https://github.com/rapidfuzz/RapidFuzz/blob/main/src/rapidfuzz/fuzz_py.py]

### Pattern 5: TokenSetRatio three-way max (verbatim from RapidFuzz reference)

**RapidFuzz reference (MIT):**
```python
# rapidfuzz/fuzz_py.py: token_set_ratio (~line 408 onward)
tokens_a = set(s1.split())
tokens_b = set(s2.split())

if not tokens_a or not tokens_b:
    return 0  # bug-for-bug compat with fuzzywuzzy issue #110

intersect = tokens_a.intersection(tokens_b)
diff_ab = tokens_a.difference(tokens_b)
diff_ba = tokens_b.difference(tokens_a)

# Subset short-circuit: one sentence is part of the other
if intersect and (not diff_ab or not diff_ba):
    return 100

diff_ab_joined = " ".join(sorted(diff_ab))
diff_ba_joined = " ".join(sorted(diff_ba))

ab_len = len(diff_ab_joined)
ba_len = len(diff_ba_joined)
sect_len = len(" ".join(intersect))   # NOTE: intersect is a SET — iteration order varies

# string-construction lengths (sect_ab = "intersection diff_ab" with single space if both nonempty)
sect_ab_len = sect_len + (1 if sect_len != 0 else 0) + ab_len
sect_ba_len = sect_len + (1 if sect_len != 0 else 0) + ba_len

# Three ratios:
# 1. ratio(diff_ab_joined, diff_ba_joined) = ratio of just-the-differences   <-- via indel
# 2. ratio(intersection, intersection+diff_ab) = analytical (length-based, no DP)
# 3. ratio(intersection, intersection+diff_ba) = analytical (length-based, no DP)

# Detail: sect_ab_dist = (1 if sect_len != 0 else 0) + ab_len  -- the # of chars that
# differ between "intersection" and "intersection diff_ab" is exactly ab_len + 1 (for the space)
# so the Indel ratio reduces to (1 - dist/(sect_len + sect_ab_len))

return max(
    indelRatio(diff_ab_joined, diff_ba_joined),
    analyticalRatio(sect_len, sect_ab_len, ab_len),
    analyticalRatio(sect_len, sect_ba_len, ba_len),
)
```

**Critical landmines:**

1. **Empty-token-set returns 0.0, NOT 1.0.** Both fuzzywuzzy and RapidFuzz return `0` (not `100`) when either input has no tokens. This is bug-for-bug compatibility per RapidFuzz issue #110. The project should match this for cross-validation alignment — diverging here breaks the cross-validation gate. Document the intentional deviation from the both-empty-→-1.0 convention with a clear godoc note. [VERIFIED: fuzz_py.py line 467, https://github.com/rapidfuzz/RapidFuzz/issues/110]

2. **Subset short-circuit at the top.** When `intersect` is non-empty AND one of `diff_ab` / `diff_ba` is empty (one set is a subset of the other), the function returns 100 (i.e. 1.0) WITHOUT computing any of the three ratios. Implementation MUST handle this short-circuit explicitly.

3. **Set iteration order randomness for `sect_len`.** RapidFuzz computes `sect_len = len(" ".join(intersect))` where `intersect` is an unordered Python set — the LENGTH is deterministic (independent of order), but if we ever use the joined intersection STRING (we don't — only its length), order matters. The project's implementation MUST NOT iterate the intersection set on output paths (DET-03); compute `sect_len` from sorted-key iteration or a length-only sum to satisfy DET-03 even though RapidFuzz gets away with the set-order randomness because only length is consumed.

4. **Three-way max construction order in tests.** At least one test pair MUST exercise the case where the max is NOT the intersection-only ratio — i.e. the `diff_ab` vs `diff_ba` comparison wins. CONTEXT §6 specifies this; planner MUST verify the bench / unit test fixtures cover it.

[VERIFIED: rapidfuzz/fuzz_py.py lines 446-509]

### Pattern 6: PartialRatio sliding-window (verbatim from RapidFuzz reference)

**The RapidFuzz `_partial_ratio_impl` is NOT a naive sliding window.** It has THREE iteration regions plus an `s1_char_set` early-skip:

```python
# rapidfuzz/fuzz_py.py: _partial_ratio_impl(s1, s2, score_cutoff)
# precondition: len(s1) <= len(s2)

s1_char_set = set(s1)
len1 = len(s1)
len2 = len(s2)

# Region 1: left-tail (substrings shorter than len1, from position 0)
for i in range(1, len1):
    substr_last = s2[i - 1]
    if substr_last not in s1_char_set:
        continue
    # compute ratio(s1, s2[:i])
    ...

# Region 2: middle (full-length windows starting at positions 0..len2-len1)
for i in range(len2 - len1):
    substr_last = s2[i + len1 - 1]
    if substr_last not in s1_char_set:
        continue
    # compute ratio(s1, s2[i : i + len1])
    ...

# Region 3: right-tail (substrings shorter than len1, from end)
for i in range(len2 - len1, len2):
    substr_first = s2[i]
    if substr_first not in s1_char_set:
        continue
    # compute ratio(s1, s2[i:])
    ...

return max ratio over all three regions
```

**Why three regions?** RapidFuzz searches alignments where the shorter string overlaps the longer string PARTIALLY at the start (left-tail) and end (right-tail), in addition to fully-aligned middle positions. This is the "best partial" semantic — the shorter string can hang off either end.

**Early-skip optimisation:** `s1_char_set` lets RapidFuzz skip alignments where the LAST (or first, in region 3) character of the candidate substring doesn't appear in `s1` at all — that alignment cannot improve the running max. The project MUST replicate this for budget compliance — without it, `BenchmarkPartialRatio_Pathological_LongShortMismatch` (10-char vs 10000-char) blows past the < 10 µs budget.

**Spec-deferred sliding-window DP optimisation:** spec line 612 explicitly defers the O(nm) sliding-window DP (vs the naive O(nm·(n-m)) repeated indelRatio calls) to v1.x. The planner MUST add a `// TODO(#<issue>): implement sliding-window DP per Bachmann RapidFuzz docs` referencing a future GitHub issue. The straightforward implementation in Phase 6 is "loop over alignments; call `lcsLen` repeatedly" with the `s1_char_set` early-skip.

**Both byte AND rune surfaces ship in Phase 6 per spec line 609-610:**
- `PartialRatioScore(a, b string) float64` — bytes
- `PartialRatioScoreRunes(a, b string) float64` — runes (calls `lcsLenRunes` / `indelRatioRunes`)

**Why no tokenisation in PartialRatio?** PartialRatio is a CHARACTER-LEVEL sliding-window over the longer string — not a token-level operation. Hence no `Tokenise` call. The "longer / shorter" distinction is purely a length comparison. [VERIFIED: rapidfuzz/fuzz_py.py lines 122-186]

### Pattern 7: Monge-Elkan composition

**Asymmetric formula (Monge & Elkan 1996):**
```
ME(A, B, sim) = (1 / |A|) · Σ_{i=1..|A|} max_{j=1..|B|} sim(A[i], B[j])

where:
  A = Tokenise(a, DefaultTokeniseOptions())   # ordered token list
  B = Tokenise(b, DefaultTokeniseOptions())
  sim = inner-metric function dispatched via AlgoID
```

**Symmetric formula (project choice — arithmetic mean per CONTEXT §4 LOCKED):**
```
ME_sym(A, B, sim) = (ME(A, B, sim) + ME(B, A, sim)) / 2.0
```

**Edge cases:**
- Both inputs produce empty token lists (e.g. `Tokenise("") = []`): return `1.0` (both-empty identity).
- One produces empty tokens, the other does not: return `0.0` (one-empty convention).
- Asymmetric `MongeElkanScore("a b", "a")` is well-defined: `|A|=2, |B|=1; max_j sim("a", *) over B={"a"} = sim("a","a")=1.0; max_j sim("b", *) over B={"a"} = sim("b","a")<1; ME = (1.0 + low)/2`.

**Inner-metric dispatch via AlgoID allow-list:**
```go
var permittedMongeElkanInner = map[AlgoID]bool{
    AlgoLevenshtein: true,
    AlgoDamerauLevenshteinOSA: true,
    AlgoDamerauLevenshteinFull: true,
    AlgoHamming: true,
    AlgoJaro: true,
    AlgoJaroWinkler: true,
    AlgoStrcmp95: true,
    AlgoSmithWatermanGotoh: true,
    AlgoLCSStr: true,
    AlgoQGramJaccard: true,
    AlgoSorensenDice: true,
    AlgoCosine: true,
    AlgoTversky: true,
    // EXPLICITLY NOT in map: AlgoMongeElkan (self-recursion),
    // AlgoTokenSortRatio, AlgoTokenSetRatio, AlgoPartialRatio,
    // AlgoTokenJaccard (token-on-token meaningless), and Phase 7
    // phonetic algorithms (added in Phase 7).
}
```

**On invalid inner:** `panic("fuzzymatch: AlgoID <X> not permitted as Monge-Elkan inner metric")` per CONTEXT §3 LOCKED. Phase 8 Scorer's `WithMongeElkanAlgorithm` returns `ErrInvalidAlgoID` instead.

**Inner metric resolution:** since `dispatch[inner]` is `func(a, b string) float64`, Monge-Elkan calls `dispatch[inner](tokenA, tokenB)` for each `(i, j)` pair. Per-token allocation cost: zero (tokens are already strings from `Tokenise`).

**Dispatch wrapper (`dispatch_monge_elkan.go`):** wires `AlgoMongeElkan → MongeElkanScoreSymmetric(a, b, AlgoJaroWinkler, DefaultNormalisationOptions)` per CONTEXT §4 LOCKED. The dispatch table signature `func(a, b string) float64` has no slot for inner metric or symmetric flag — the wrapper picks the spec defaults (Jaro-Winkler inner, symmetric variant).

**`NormalisationOptions` parameter:** the project's `Normalise` (Phase 1) accepts `NormalisationOptions` — Monge-Elkan's signature is likely `MongeElkanScore(a, b string, inner AlgoID, opts NormalisationOptions) float64`. Planner confirms exact signature shape with api-ergonomics-reviewer. The dispatch wrapper passes `DefaultNormalisationOptions()`.

[VERIFIED: Monge-Elkan formula from "An Algorithm to Calculate the p-Value of the Monge-Elkan Distance" (Ryšavý & Železný 2025), https://ida.fel.cvut.cz/zelezny/pubs/jcb-author-version.pdf — formula reproduces the 1996 original; cross-checked at https://www.gabormelli.com/RKB/Monge-Elkan_Distance_Function]

### Pattern 8: TokenJaccard

**Formula (Jaccard 1912 — set-Jaccard, NOT multiset):**
```
J(A, B) = |A ∩ B| / |A ∪ B|

where A = set(Tokenise(a)), B = set(Tokenise(b))
```

**Edge cases:**
- Both produce empty token sets: return `1.0` (both-empty identity convention).
- One empty: return `0.0`.
- Identical: `1.0` (covered by `a == b` short-circuit).

**Implementation:**
```go
func TokenJaccardScore(a, b string) float64 {
    if a == b {
        return 1.0  // identity short-circuit
    }
    tokensA := Tokenise(a, DefaultTokeniseOptions())
    tokensB := Tokenise(b, DefaultTokeniseOptions())
    if len(tokensA) == 0 && len(tokensB) == 0 {
        return 1.0
    }
    if len(tokensA) == 0 || len(tokensB) == 0 {
        return 0.0
    }
    // Build sets via map[string]struct{} or map[string]bool
    setA := make(map[string]struct{}, len(tokensA))
    for _, t := range tokensA {
        setA[t] = struct{}{}
    }
    setB := make(map[string]struct{}, len(tokensB))
    for _, t := range tokensB {
        setB[t] = struct{}{}
    }
    // Intersection: walk smaller side, lookup against larger side.
    // Union = |A| + |B| - |intersection|.
    // INTEGER counters; map iteration order does NOT affect output (DET-03).
    small, large := setA, setB
    if len(setB) < len(setA) {
        small, large = setB, setA
    }
    var intersection int
    for k := range small {
        if _, ok := large[k]; ok {
            intersection++
        }
    }
    union := len(setA) + len(setB) - intersection
    return float64(intersection) / float64(union)
}
```

**Note distinct from Phase 5 Q-Gram Jaccard:** Q-Gram Jaccard uses MULTISET (counts repetitions); Token Jaccard uses SET (no repetitions). This is a deliberate semantic difference — token presence is a binary signal in this context.

**TokenJaccard does NOT have a Runes variant.** `Tokenise` already handles Unicode (via `[]rune` decode internally per `tokenise.go` line 161). Token equality is string-equality of UTF-8-encoded runes; no separate byte-vs-rune surface needed. Mirrors the spec — TokenJaccard ships ONE public function. [CITED: CONTEXT §6 "Tokenised algorithms ... single surface"]

### Pattern 9: Cross-validation generator script

**Mirrors Phase 4's `gen-ratcliff-obershelp-cross-validation.py` structurally; uses `rapidfuzz` instead of stdlib `difflib`.**

```python
#!/usr/bin/env python3
"""scripts/gen-token-ratio-cross-validation.py"""

import json
import os
import sys
from datetime import datetime, timezone

RAPIDFUZZ_VERSION = "3.14.5"
RAPIDFUZZ_INSTALLED_AT = "2026-MM-DDTHH:MM:SSZ"  # filled at script-generation time

import rapidfuzz
assert rapidfuzz.__version__ == RAPIDFUZZ_VERSION, (
    f"rapidfuzz version mismatch: installed {rapidfuzz.__version__}, "
    f"script pinned to {RAPIDFUZZ_VERSION}"
)

from rapidfuzz import fuzz

# CASES: 20-30 entries per algorithm × 4 surfaces (TokenSort, TokenSet,
# PartialRatio bytes, PartialRatio runes) = 80-120 entries total.
# Each entry produces 4 outputs (one per surface).
CASES = [
    # Identity cases
    {"name": "identity_short", "a": "hello", "b": "hello"},
    {"name": "both_empty", "a": "", "b": ""},
    {"name": "one_empty_a", "a": "", "b": "hello world"},
    {"name": "one_empty_b", "a": "hello world", "b": ""},

    # Basic token-sort behaviour
    {"name": "tokens_reordered", "a": "fuzzy wuzzy was a bear",
     "b": "wuzzy fuzzy was a bear"},

    # Token-set asymmetric
    {"name": "subset_a_in_b", "a": "fuzzy bear",
     "b": "fuzzy bear lived in the woods"},

    # Partial-ratio sliding window
    {"name": "partial_short_in_long", "a": "YANKEES",
     "b": "NEW YORK YANKEES"},
    {"name": "partial_disjoint", "a": "abc", "b": "xyzzzz"},

    # Unicode (rune-path validation only — bytes path computes byte LCS
    # which differs)
    {"name": "unicode_accents", "a": "café société",
     "b": "cafe societe"},

    # Token-set three-way max edge: case where diff-ab vs diff-ba wins
    {"name": "tokenset_diff_dominates",
     "a": "alpha beta gamma delta", "b": "alpha beta epsilon zeta"},

    # Pathological-length pair (small)
    {"name": "long_short_mismatch",
     "a": "x", "b": "y" * 1000 + "x" + "y" * 1000},

    # ... 20-25 more entries spanning all four locked categories
]

entries = []
for case in CASES:
    entry = {
        "name": case["name"],
        "a": case["a"],
        "b": case["b"],
        "token_sort_ratio": fuzz.token_sort_ratio(case["a"], case["b"]) / 100.0,
        "token_set_ratio": fuzz.token_set_ratio(case["a"], case["b"]) / 100.0,
        "partial_ratio_bytes": fuzz.partial_ratio(case["a"], case["b"]) / 100.0,
        # For partial_ratio_runes: the cross-validation runs on Python str
        # which IS unicode-aware; fuzzymatch's PartialRatioScoreRunes
        # operates on runes — so the rune path matches Python's str path
        # for these inputs.
        "partial_ratio_runes": fuzz.partial_ratio(case["a"], case["b"]) / 100.0,
    }
    entries.append(entry)

corpus = {
    "version": 1,
    "rapidfuzz_version": RAPIDFUZZ_VERSION,
    "rapidfuzz_installed_at": RAPIDFUZZ_INSTALLED_AT,
    "python_version": sys.version.split()[0],
    "regenerated_at": datetime.now(timezone.utc).isoformat(),
    "entries": entries,
}

with open("testdata/cross-validation/token-ratios/vectors.json", "w") as f:
    json.dump(corpus, f, indent=2, sort_keys=False)
    f.write("\n")
```

**Critical constraint — tokeniser-divergence avoidance:** the `CASES` corpus MUST consist of inputs where the project's `Tokenise(s, DefaultTokeniseOptions())` produces THE SAME token list as Python's `s.split()`. This means:
- Use whitespace-only separators (`" "`).
- Use lowercase ASCII OR document the case-sensitivity divergence (project lowercases; RapidFuzz preserves case).
- AVOID camelCase, snake_case, kebab-case, dot-case in the corpus inputs.

If any corpus input violates this, the cross-validation test will fail with "expected RapidFuzz score X, got fuzzymatch score Y" because of the tokeniser divergence — NOT because of a real algorithm bug.

**Lowercasing reconciliation:** the Project's `DefaultTokeniseOptions{}.Lowercase = true`; RapidFuzz's `processor=None` (no lowercase). Either: (1) use only-already-lowercase corpus inputs, OR (2) the cross-validation script calls `.lower()` on all inputs before passing to RapidFuzz. Path (2) is cleaner — add `entry["a"] = case["a"].lower()` and `entry["b"] = case["b"].lower()` to ensure both implementations see the same input.

[CITED: scripts/gen-ratcliff-obershelp-cross-validation.py (Phase 4 precedent for structure); rapidfuzz/fuzz_py.py for behavioural verification]

### Pattern 10: Cross-validation Go test

**Mirrors `TestRatcliffObershelp_CrossValidation` (Phase 4):**
```go
// In token_ratio_cross_validation_test.go (or per-algorithm files)
type tokenRatioEntry struct {
    Name              string  `json:"name"`
    A                 string  `json:"a"`
    B                 string  `json:"b"`
    TokenSortRatio    float64 `json:"token_sort_ratio"`
    TokenSetRatio     float64 `json:"token_set_ratio"`
    PartialRatioBytes float64 `json:"partial_ratio_bytes"`
    PartialRatioRunes float64 `json:"partial_ratio_runes"`
}

type tokenRatioCorpus struct {
    Version          int               `json:"version"`
    RapidFuzzVersion string            `json:"rapidfuzz_version"`
    PythonVersion    string            `json:"python_version"`
    Entries          []tokenRatioEntry `json:"entries"`
}

func TestTokenRatios_CrossValidation(t *testing.T) {
    const epsilon = 1e-9
    path := filepath.Join("testdata", "cross-validation", "token-ratios", "vectors.json")
    raw, err := os.ReadFile(path)
    if err != nil {
        t.Fatalf("read %s: %v (regenerate with `make regen-token-ratio-cross-validation`)", path, err)
    }
    var c tokenRatioCorpus
    if err := json.Unmarshal(raw, &c); err != nil {
        t.Fatalf("parse %s: %v", path, err)
    }
    if c.Version != 1 {
        t.Fatalf("unsupported corpus version %d", c.Version)
    }
    if c.RapidFuzzVersion == "" {
        t.Fatalf("corpus missing rapidfuzz_version field — regenerate")
    }
    for _, e := range c.Entries {
        e := e
        t.Run(e.Name+"/token_sort", func(t *testing.T) {
            got := fuzzymatch.TokenSortRatioScore(e.A, e.B)
            if math.Abs(got-e.TokenSortRatio) > epsilon {
                t.Errorf("TokenSortRatioScore(%q,%q) = %.17g; rapidfuzz = %.17g (rapidfuzz_version=%s)",
                    e.A, e.B, got, e.TokenSortRatio, c.RapidFuzzVersion)
            }
        })
        // ... TokenSet, PartialRatio bytes, PartialRatio runes sub-tests
    }
}
```

[CITED: ratcliff_obershelp_test.go lines 283-356 — exact pattern]

### Anti-Patterns to Avoid

- **Don't iterate Python `set` order semantics in the project's TokenSet.** RapidFuzz gets away with `len(" ".join(intersect))` because only the length matters, but the project MUST NOT iterate the intersection map on output paths (DET-03). Compute `sect_len` from sorted-key iteration or a count-only sum.
- **Don't add testify in any `_test.go` in the root module.** Stricter than mask. (`go-coding-standards` skill.)
- **Don't put `permittedMongeElkanInner` in `init()`.** Use `var = map[...]bool{...}` literal (DET-13 / Phase 5 §5).
- **Don't reuse `lcsstr.go`'s LCS-substring DP.** That's the contiguous-match variant; LCS-subsequence (Indel) allows gaps. They are different DPs.
- **Don't naive-loop PartialRatio without the `s1_char_set` early-skip.** Pathological 10-vs-10000-char benchmarks will blow past the < 10 µs budget.
- **Don't promote `lcsLen` / `indelRatio` to public surface in v1.** Reserved for v1.x (see CONTEXT.md Deferred). Promoting an unexported helper is non-breaking; the inverse is not.
- **Don't compute Cosine-style `math.Pow` or `math.Exp` anywhere.** Phase 6 algorithms compute via integer arithmetic + one division. NO transcendentals.

## Don't Hand-Roll

| Problem | Don't Build | Use Instead | Why |
|---------|-------------|-------------|-----|
| LCS DP recurrence | Custom variant | Wagner-Fischer 1974 two-row DP | The canonical formulation is unambiguous and 100% understood; deviating courts subtle off-by-one bugs. Inline the recurrence per Pattern 2. |
| Token splitting | Custom regex | `Tokenise(s, DefaultTokeniseOptions())` from Phase 1 | The project's tokeniser is the single source of truth; bypassing it diverges from `WithMongeElkanAlgorithm` semantics in Phase 8 |
| AlgoID dispatch | New dispatch system | Existing `dispatch[AlgoID]` array (Phase 1) | Zero-alloc; already tested across 14 algorithms |
| Cross-validation reference | Hand-derived 80-vector corpus | RapidFuzz Python (CONTEXT §1 LOCKED) | Hand-derivation of 80 LCS computations is reviewer-hostile |
| Indel-formula `2·LCS/(m+n)` | Custom formulation | Pattern 3 (proven equivalent to RapidFuzz's `1 - dist/(m+n)`) | Ambiguity-free; both forms produce the same float64 |

**Key insight:** Phase 6's complexity is in the *composition* (TokenSet's three-way max, PartialRatio's three-region iteration, Monge-Elkan's allow-list dispatch), not in the *kernel*. The kernel is one well-understood DP; the cost is in getting the composition to exactly match RapidFuzz's reference behaviour. Read RapidFuzz's `fuzz_py.py` Python source as the behavioural specification — copy structure, NOT code (per algorithm-licensing-standards fresh-implementation discipline).

## Common Pitfalls

### Pitfall 1: Tokeniser Divergence between fuzzymatch and RapidFuzz

**What goes wrong:** Cross-validation corpus contains an input like `"fooBar"` or `"foo_bar"`; RapidFuzz tokenises as `["fooBar"]` (single whitespace-bounded token); fuzzymatch tokenises as `["foo", "bar"]` (camelCase / snake_case split). TokenSort/TokenSet scores diverge sharply — cross-validation test fails — reviewer assumes the algorithm is wrong, not the corpus.

**Why it happens:** The project's `Tokenise` is camelCase / snake / kebab / dot aware AND lowercasing; RapidFuzz's `_split_sequence` is `seq.split()` (Python whitespace-only, case-preserving).

**How to avoid:**
1. Restrict the cross-validation corpus to inputs where `Tokenise(s, DefaultTokeniseOptions())` and `s.split()` produce equivalent token lists — i.e. whitespace-separated lowercase ASCII text only.
2. Have the generator script call `.lower()` on inputs before passing to RapidFuzz to handle the case mismatch.
3. Document the divergence prominently in the algorithm's godoc and in CONTRIBUTING / `docs/cross-validation.md`.
4. Add a meta-test that asserts every entry in `vectors.json` has `Tokenise(entry.a, DefaultTokeniseOptions())` matching `strings.Fields(strings.ToLower(entry.a))` — this catches accidental corpus drift.

**Warning signs:** Cross-validation test fails on a single entry whose input contains `_`, `-`, `.`, or any uppercase letter. Diff shows score `0.something` vs `1.0` (or vice versa). The divergence is non-subtle — different token lists produce wildly different scores.

[VERIFIED: rapidfuzz/fuzz_py.py `_split_sequence` lines 27-43; tokenise.go DefaultTokeniseOptions line 105-112]

### Pitfall 2: Empty-token-set returns 0.0, not 1.0, in TokenSetRatio

**What goes wrong:** Project implements TokenSetRatio with the standard `both_empty → 1.0` convention from algorithm-correctness-standards. Cross-validation against RapidFuzz fails on the both-empty entry (`""`, `""`) because RapidFuzz returns 0, not 1.

**Why it happens:** RapidFuzz's `token_set_ratio` returns 0 when either token set is empty — bug-for-bug compatibility with fuzzywuzzy per RapidFuzz issue #110. This is a deliberate divergence from the both-empty-identity convention used by every other algorithm in the catalogue.

**How to avoid:** Document the deviation in `token_set_ratio.go` godoc explicitly:
```
// Edge case (deviation from the catalogue convention):
//
//   When EITHER input produces zero tokens, TokenSetRatioScore returns
//   0.0 — NOT 1.0. This matches RapidFuzz's behaviour (rapidfuzz issue
//   #110) which itself mirrors fuzzywuzzy. Other tokenised algorithms
//   in the catalogue (TokenJaccard, MongeElkan) follow the standard
//   both-empty → 1.0 convention; TokenSetRatio is the documented
//   exception. The deviation is necessary because the algorithm's
//   three-way construction has no meaningful interpretation when there
//   are no tokens to intersect.
```

**Warning signs:** Property test `TestProp_TokenSetRatioScore_Identity` fails on `""` input (because identity says both-empty should be 1.0, but TokenSet returns 0.0). The Identity property test for TokenSetRatio MUST guard `if x == "" { return true }` to skip the empty case (mirrors how other algorithms handle their edge cases).

[VERIFIED: rapidfuzz/fuzz_py.py line 467 + comment "in FuzzyWuzzy this returns 0. For sake of compatibility return 0 here as well"; https://github.com/rapidfuzz/RapidFuzz/issues/110]

### Pitfall 3: PartialRatio sliding-window ignored regions

**What goes wrong:** Implementer reads "sliding window" and writes a naive single-loop:
```go
// WRONG — only covers Region 2
for i := 0; i <= len(longer)-len(shorter); i++ {
    score := indelRatio(shorter, longer[i:i+len(shorter)])
    if score > best { best = score }
}
```
This MISSES Regions 1 and 3 — alignments where the shorter string hangs off the start or end of the longer string. Cross-validation fails on inputs like `("abc", "bc")` or `("abc", "ab")` where the optimal alignment is partial.

**Why it happens:** The Wikipedia / textbook description of "partial ratio" sometimes omits the tail regions. RapidFuzz's reference implementation has all three.

**How to avoid:** Implement all three regions per Pattern 6. Add an explicit unit test:
```go
{"a": "abc", "b": "bc", "expected": 1.0},  // Region 3 (right tail) wins
{"a": "abc", "b": "ab", "expected": 1.0},  // Region 1 (left tail) wins
```
Both should return 1.0 (perfect partial match). If they return < 1.0, Regions 1/3 are missing.

**Warning signs:** Cross-validation fails specifically on entries where `len(a) < len(b)` and the matching substring is at the beginning or end of `b`. The middle region passes but tail entries fail.

[VERIFIED: rapidfuzz/fuzz_py.py `_partial_ratio_impl` lines 122-186 — three explicit `for` loops]

### Pitfall 4: Monge-Elkan inner-metric self-reference infinite recursion

**What goes wrong:** Implementer forgets to exclude `AlgoMongeElkan` from `permittedMongeElkanInner`. A user calls `MongeElkanScore(a, b, AlgoMongeElkan)` — Monge-Elkan tokenises, then for each pair calls Monge-Elkan again, which tokenises (single tokens that often produce one-token lists), recurses, eventually stack-overflows.

**Why it happens:** Spec line 744 explicitly notes this case but is easy to overlook. Phase 7 will ADD entries to `permittedMongeElkanInner` (phonetic algorithms) — the test that asserts `AlgoMongeElkan` is absent from the map MUST stay green even after Phase 7 mutations.

**How to avoid:**
1. Hard-code the exclusion: in the map literal, add a comment `// AlgoMongeElkan: explicitly excluded — self-recursion`.
2. Add a unit test:
   ```go
   func TestMongeElkan_SelfReferenceRejected(t *testing.T) {
       defer func() {
           if r := recover(); r == nil {
               t.Fatal("MongeElkanScore(*, *, AlgoMongeElkan) did not panic")
           }
       }()
       _ = fuzzymatch.MongeElkanScore("a b", "c d", fuzzymatch.AlgoMongeElkan, fuzzymatch.DefaultNormalisationOptions())
   }
   ```
3. Add an exhaustive panic test that walks ALL 23 AlgoIDs and asserts the documented panic message for the 10 that are NOT in the allow-list (Phase 6: 13 permitted, 10 rejected; Phase 7: 17 permitted, 6 rejected).

**Warning signs:** A `runtime: goroutine stack exceeds <limit>` error from running the algorithm with `AlgoMongeElkan` as inner. Test suite hangs.

[VERIFIED: docs/requirements.md line 744 (per CONTEXT.md citation); CONTEXT §3]

### Pitfall 5: TokenSet's `sect_len` map-iteration determinism

**What goes wrong:** Project copies RapidFuzz's `len(" ".join(intersect))` literally:
```go
// WRONG — iterates a map for output construction
intersect := buildIntersectionSet(setA, setB)
joined := strings.Join(toSlice(intersect), " ")  // slice from map = random order
sectLen := len(joined)  // length is deterministic, but if anything else uses joined, broken
```
The LENGTH of the joined string is deterministic regardless of order, but if the implementation EVER touches `joined` (e.g. in a debug log, a test fixture, an alternate code path), DET-03 is violated.

**Why it happens:** RapidFuzz gets away with this because Python's `set` iteration is ALSO randomised (insertion-order dependent in CPython 3.7+, but not guaranteed across versions / impls), and only the `len` is consumed. Go's stricter determinism rules require explicit ordering.

**How to avoid:** Compute `sect_len` via a sum over sorted keys, OR via a length-summation loop that doesn't construct a string:
```go
// CORRECT — output is an integer count, no string is built
var sectLen int
for k := range intersection {
    sectLen += len(k)
}
if len(intersection) > 0 {
    sectLen += len(intersection) - 1  // separators between tokens
}
// integer addition is associative; iteration order doesn't affect output (DET-03 satisfied)
```

**Warning signs:** `golangci-lint` flags map iteration in `token_set_ratio.go` on output paths. The cross-platform determinism CI matrix shows random byte differences in `algorithms.json` for TokenSetRatio entries.

[CITED: determinism-standards SKILL.md "No-Map-Iteration Rule"; DET-03; rapidfuzz/fuzz_py.py line 480]

### Pitfall 6: `lcsLen` vs `lcsstr.go`'s LCS-substring confusion

**What goes wrong:** Implementer references `lcsstr.go`'s DP and adapts it. `lcsstr.go` resets to 0 on mismatch (substring = contiguous); `token_indel.go`'s `lcsLen` takes max on mismatch (subsequence = with gaps). Off-by-one fundamental: `lcsstr.go("abc", "axc")` returns 1 (substring = "a" or "c"); `lcsLen("abc", "axc")` returns 2 (subsequence = "ac"). All Indel formulas break — every TokenSort score is wrong.

**Why it happens:** Both files contain "LCS" in name. Reviewer reads the existing `lcsstr.go` and assumes it's the right helper.

**How to avoid:**
1. The `token_indel.go` file's godoc EXPLICITLY contrasts with `lcsstr.go`:
   ```
   // lcsLen computes the length of the longest common SUBSEQUENCE of a
   // and b — characters can appear non-contiguously in a or b but must
   // appear in the same relative order. This is the "Indel" semantic of
   // RapidFuzz's `Indel` distance metric.
   //
   // Distinct from lcsstr.go's LongestCommonSubstring, which requires
   // CONTIGUOUS matches. Example: lcsLen("abc", "axc") = 2 ("ac");
   // LongestCommonSubstring("abc", "axc") = 1 ("a" or "c").
   ```
2. Add a regression test in `token_indel_test.go` that pins the divergence:
   ```go
   func TestLCSLen_DistinctFromLCSStr(t *testing.T) {
       // Subsequence: "ac" (a, then c skipping x)
       if got := fuzzymatch.LCSLenForTest([]byte("abc"), []byte("axc")); got != 2 {
           t.Errorf("lcsLen('abc','axc') = %d, want 2 (subsequence)", got)
       }
       // Substring (for contrast): "a" or "c", length 1
       if got, _ := fuzzymatch.LongestCommonSubstring("abc", "axc"); got != "a" && got != "c" {
           t.Errorf("LongestCommonSubstring('abc','axc') = %q, want 'a' or 'c' (substring)", got)
       }
   }
   ```

**Warning signs:** TokenSortRatio cross-validation fails uniformly with low scores (~50% of expected). The kernel is computing substring instead of subsequence.

## Code Examples

### Example 1: lcsLen two-row DP (Pattern 2 — verified)
```go
// lcsLen returns the length of the longest common SUBSEQUENCE of a and b.
// O(|a|·|b|) time; O(min(|a|,|b|)) space via two-row DP.
//
// Source: Wagner & Fischer 1974, J. ACM 21(1):168-173 — the canonical
// Wagner-Fischer DP recurrence for LCS.
func lcsLen(a, b []byte) int {
    // Always make the inner loop over the shorter string.
    if len(b) < len(a) {
        a, b = b, a
    }
    m, n := len(a), len(b)
    if m == 0 {
        return 0
    }
    var prevArr, currArr [65]int  // stack-allocated when m <= 64
    var prev, curr []int
    if m <= 64 {
        prev = prevArr[:m+1]
        curr = currArr[:m+1]
    } else {
        prev = make([]int, m+1)
        curr = make([]int, m+1)
    }
    for j := 1; j <= n; j++ {
        for i := 1; i <= m; i++ {
            if a[i-1] == b[j-1] {
                curr[i] = prev[i-1] + 1
            } else {
                if prev[i] > curr[i-1] {
                    curr[i] = prev[i]
                } else {
                    curr[i] = curr[i-1]
                }
            }
        }
        prev, curr = curr, prev
    }
    return prev[m]
}
```
[SOURCE: Wagner-Fischer 1974 + project Phase 2 two-row DP discipline]

### Example 2: indelRatio with both-empty / one-empty handling
```go
// indelRatio returns 2·lcsLen(a,b) / (len(a)+len(b)) in [0.0, 1.0].
// Both-empty returns 1.0 (identity convention); one-empty returns 0.0.
func indelRatio(a, b []byte) float64 {
    sum := len(a) + len(b)
    if sum == 0 {
        return 1.0  // both-empty identity
    }
    if len(a) == 0 || len(b) == 0 {
        return 0.0  // one-empty
    }
    lcs := lcsLen(a, b)
    return 2.0 * float64(lcs) / float64(sum)
}
```

### Example 3: TokenSortRatio composition
```go
// TokenSortRatioScore tokenises both sides, sorts the token lists,
// joins with space, and applies the Indel-formula similarity ratio
// (`2·LCS(joined_a, joined_b) / (len(joined_a) + len(joined_b))`).
//
// See package godoc for the divergence from RapidFuzz token_sort_ratio:
// fuzzymatch uses Tokenise (camelCase-aware, lowercased) instead of
// Python str.split() (whitespace-only, case-preserving).
func TokenSortRatioScore(a, b string) float64 {
    if a == b {
        return 1.0  // identity short-circuit
    }
    tokensA := Tokenise(a, DefaultTokeniseOptions())
    tokensB := Tokenise(b, DefaultTokeniseOptions())
    if len(tokensA) == 0 && len(tokensB) == 0 {
        return 1.0
    }
    if len(tokensA) == 0 || len(tokensB) == 0 {
        return 0.0
    }
    sort.Strings(tokensA)
    sort.Strings(tokensB)
    joinedA := strings.Join(tokensA, " ")
    joinedB := strings.Join(tokensB, " ")
    return indelRatio([]byte(joinedA), []byte(joinedB))
}
```

### Example 4: PartialRatio with three-region iteration + char-set early-skip
```go
// PartialRatioScore returns the maximum Indel-formula similarity ratio
// over alignments of the shorter string against the longer string. Uses
// the three-region iteration (left tail, middle, right tail) and
// char-set early-skip per RapidFuzz's _partial_ratio_impl reference.
func PartialRatioScore(a, b string) float64 {
    if a == b { return 1.0 }
    if len(a) == 0 && len(b) == 0 { return 1.0 }
    if len(a) == 0 || len(b) == 0 { return 0.0 }

    var shorter, longer []byte
    if len(a) <= len(b) {
        shorter, longer = []byte(a), []byte(b)
    } else {
        shorter, longer = []byte(b), []byte(a)
    }
    m, n := len(shorter), len(longer)

    // Build s1_char_set for early-skip optimisation
    var charSet [256]bool
    for _, ch := range shorter {
        charSet[ch] = true
    }

    best := 0.0

    // Region 1: left-tail (substrings shorter than m, ending at positions 1..m-1)
    for i := 1; i < m; i++ {
        if !charSet[longer[i-1]] { continue }
        if r := indelRatio(shorter, longer[:i]); r > best { best = r }
    }

    // Region 2: middle (full m-length windows starting at 0..n-m)
    for i := 0; i <= n-m; i++ {
        if !charSet[longer[i+m-1]] { continue }
        if r := indelRatio(shorter, longer[i:i+m]); r > best { best = r }
        if best == 1.0 { return 1.0 }  // early exit
    }

    // Region 3: right-tail (substrings shorter than m, starting at n-m..n-1)
    if n > m {
        for i := n - m; i < n; i++ {
            if !charSet[longer[i]] { continue }
            if r := indelRatio(shorter, longer[i:]); r > best { best = r }
        }
    }

    return best
}
```
[SOURCE: structural transcription from rapidfuzz/fuzz_py.py `_partial_ratio_impl`; fresh Go implementation — no code copied]

### Example 5: Monge-Elkan asymmetric + symmetric
```go
// MongeElkanScore returns the asymmetric Monge-Elkan similarity
// (Monge & Elkan 1996, KDD'96 §3): the average over tokens(a) of the
// maximum inner-metric similarity against any token in tokens(b).
//
// inner MUST be one of the AlgoIDs in permittedMongeElkanInner; passing
// an excluded AlgoID panics with a documented message (per CONTEXT.md
// §3 LOCKED). Phase 8's WithMongeElkanAlgorithm returns ErrInvalidAlgoID
// at construction time instead.
//
// Asymmetric: MongeElkanScore(a, b, *) ≠ MongeElkanScore(b, a, *) in
// general. Use MongeElkanScoreSymmetric for the symmetric variant
// (arithmetic mean of both directions).
func MongeElkanScore(a, b string, inner AlgoID, opts NormalisationOptions) float64 {
    if !permittedMongeElkanInner[inner] {
        panic("fuzzymatch: AlgoID " + inner.String() + " not permitted as Monge-Elkan inner metric")
    }
    tokensA := Tokenise(a, DefaultTokeniseOptions())
    tokensB := Tokenise(b, DefaultTokeniseOptions())
    if len(tokensA) == 0 && len(tokensB) == 0 {
        return 1.0
    }
    if len(tokensA) == 0 || len(tokensB) == 0 {
        return 0.0
    }
    innerFn := dispatch[inner]  // safe — allow-list gated
    var sumOfMax float64
    for _, tokA := range tokensA {
        var maxSim float64
        for _, tokB := range tokensB {
            s := innerFn(tokA, tokB)
            if s > maxSim {
                maxSim = s
            }
        }
        sumOfMax += maxSim
    }
    return sumOfMax / float64(len(tokensA))
}

// MongeElkanScoreSymmetric returns (MongeElkanScore(a,b) + MongeElkanScore(b,a)) / 2.
func MongeElkanScoreSymmetric(a, b string, inner AlgoID, opts NormalisationOptions) float64 {
    return (MongeElkanScore(a, b, inner, opts) + MongeElkanScore(b, a, inner, opts)) / 2.0
}
```

## State of the Art

| Old Approach | Current Approach | When Changed | Impact |
|--------------|------------------|--------------|--------|
| fuzzywuzzy Python (Levenshtein-based ratio with Python-vs-C inconsistencies) | RapidFuzz with Indel metric (LCS-based, single canonical formula) | RapidFuzz 1.0+ (~2020) | Token ratios now have a single unambiguous reference implementation; spec mandates RapidFuzz |
| Naive O(nm) sliding-window for PartialRatio | RapidFuzz's three-region iteration + char-set early-skip | RapidFuzz 0.x | ~50× speedup on long-vs-short inputs; project replicates the early-skip in Phase 6 to meet < 10 µs budget |
| RapidFuzz with `processor=utils.default_process` (pre-processed by default) | RapidFuzz with `processor=None` (no preprocessing) | RapidFuzz 3.0 (2023) | Project's cross-validation script must explicitly call `.lower()` on inputs to reconcile with fuzzymatch's lowercasing |
| LCS-Indel via Wagner-Fischer DP | Hyyrös bit-parallel LCS algorithm (RapidFuzz internal) | Hyyrö 2004 | Performance optimisation; project sticks with classical Wagner-Fischer DP — bit-parallel is overkill for the < 5 µs Phase 6 budget |
| Monge-Elkan with default-Jaro-Winkler inner | Same — Monge-Elkan 1996 + Winkler 1990 default inner | Spec line 567 LOCKED | No change |

**Deprecated/outdated:**
- **fuzzywuzzy**: superseded by RapidFuzz; original maintainer recommends migration. Project explicitly does NOT cross-validate against fuzzywuzzy.
- **`utils.default_process`**: still available in RapidFuzz 3.x but no longer applied by default. Project explicitly does NOT call it.

## Assumptions Log

| # | Claim | Section | Risk if Wrong |
|---|-------|---------|---------------|
| A1 | The cross-validation corpus will use ONLY whitespace-separated lowercase ASCII inputs (avoiding tokeniser divergence) | Pattern 4 / Pitfall 1 | If corpus accidentally includes camelCase or snake_case, every TokenSort/TokenSet test fails — but the failure mode is loud and easy to fix |
| A2 | The cross-validation script will lowercase inputs before passing to RapidFuzz | Pattern 9 | If not lowercased, Token ratios fail on any input with uppercase letters because fuzzymatch lowercases via Tokenise but RapidFuzz preserves case |
| A3 | TokenSetRatio will adopt RapidFuzz's empty-token-set-returns-0.0 behaviour (deviating from the catalogue's both-empty-→-1.0 convention) | Pitfall 2 | If we adopt 1.0 instead, cross-validation fails on the both-empty entry; if we adopt 0.0 (recommended), the Identity property test must be guarded for empty input |
| A4 | The dispatch wrapper for AlgoMongeElkan will use Symmetric variant + AlgoJaroWinkler inner + DefaultNormalisationOptions per CONTEXT §4 | Pattern 7 | Locked in CONTEXT — no risk |
| A5 | Pinning RapidFuzz to v3.14.5 (current stable as of 2026-04-07) | Standard Stack | If a newer version drops before plan 06-01 starts, planner picks the newer version; cross-validation runs against the pinned version regardless |
| A6 | Cross-validation epsilon `1e-9` matches Phase 3 / Phase 4 conventions | Pattern 10 | Standard convention — no risk |
| A7 | PartialRatio's three-region pattern is structurally identical to RapidFuzz's `_partial_ratio_impl` | Pattern 6 / Example 4 | The three-region iteration is verifiable via RapidFuzz's MIT-licensed source; structurally direct transcription |
| A8 | Monge-Elkan inner = Jaro-Winkler default per spec line 567 | Pattern 7 | Locked in CONTEXT §4 — no risk |
| A9 | The asymmetric Monge-Elkan variant's property test follows the Tversky α≠β pattern from Phase 5 | Standard Stack | Established pattern; planner follows precedent |

## Open Questions

1. **Should TokenSortRatio / TokenSetRatio use a separate "RapidFuzz-compatible" tokeniser internally for cross-validation purposes, or accept the Tokenise divergence and constrain the corpus?**
   - What we know: fuzzymatch's Tokenise is camelCase-aware AND lowercases; RapidFuzz's `_split_sequence` is whitespace-only.
   - What's unclear: does the planner WANT to ship a separate "RapidFuzz-compatible mode" (e.g. `TokenSortRatioScore` uses Tokenise; `TokenSortRatioScoreRapidFuzzCompat` uses whitespace-only)? Or accept the divergence?
   - Recommendation: **Accept the divergence.** Document it prominently in the algorithm's godoc. The cross-validation corpus is constrained to inputs where the tokenisations match. Producing a separate "compat mode" is API surface bloat and api-ergonomics-reviewer will flag it. The project's `Tokenise` is the documented project tokeniser; the algorithm's contract is "this algorithm uses the project's Tokenise" — RapidFuzz is the formula reference, not the tokeniser reference.

2. **Should the cross-validation corpus split into per-algorithm files (`token_sort_vectors.json`, `token_set_vectors.json`, etc.) or one combined file (`vectors.json` with all four scores per entry)?**
   - What we know: Phase 3 SWG used one file; Phase 4 RO used one file.
   - What's unclear: maintainability vs file-bloat trade-off.
   - Recommendation: **Single combined file.** Each entry produces 4 scores (TokenSort, TokenSet, PartialRatio bytes, PartialRatio runes); per-algorithm sub-tests via `t.Run` give per-algorithm failure isolation. Mirrors Phase 3/4 structure.

3. **Is it acceptable for the cross-validation corpus to include `partial_ratio_runes` entries that match `partial_ratio_bytes` for ASCII inputs (and only diverge for Unicode inputs)?**
   - What we know: For ASCII inputs, byte and rune paths produce identical scores.
   - What's unclear: whether the corpus should label rune entries explicitly OR whether the test should derive the rune assertion from the byte assertion for ASCII.
   - Recommendation: **Always include both** — duplicates the assertion for ASCII (cheap), but exercises the rune path's separate code path. Tests catch regressions in either path independently.

4. **Should `permittedMongeElkanInner` be expanded BEYOND CONTEXT §3's locked list to include `AlgoRatcliffObershelp` (Phase 4 algorithm not yet enumerated)?**
   - What we know: CONTEXT §3 lists 13 inner metrics: 9 character + 4 q-gram. Ratcliff-Obershelp shipped in Phase 4 but is NOT in the list.
   - What's unclear: whether Ratcliff-Obershelp is intentionally excluded.
   - Recommendation: **Include `AlgoRatcliffObershelp` (14 entries total).** Ratcliff-Obershelp is a character-tier algorithm that fits the Monge-Elkan inner-metric semantics. Excluding it would be arbitrary. Verify with planner / project owner whether CONTEXT §3 should be amended to include it; if not, document the rationale for exclusion in `monge_elkan.go`.

## Environment Availability

| Dependency | Required By | Available | Version | Fallback |
|------------|------------|-----------|---------|----------|
| Go 1.26+ | Library compilation | ✓ | (project standard) | — |
| Python 3.7+ | `make regen-token-ratio-cross-validation` (developer-only) | likely ✓ on developer machines | 3.12 (Phase 4 precedent) | Document Python install in CONTRIBUTING |
| `rapidfuzz` Python package | `make regen-token-ratio-cross-validation` (developer-only) | NOT installed by default | 3.14.5 | `pip install --user rapidfuzz==3.14.5`; cross-validation test still passes WITHOUT rapidfuzz installed (uses committed JSON corpus) |

**Missing dependencies with no fallback:** None — all runtime dependencies are stdlib + `golang.org/x/text` (project allowlist).

**Missing dependencies with fallback:** rapidfuzz Python package — only needed when REGENERATING the corpus. Routine PR builds and CI tests load the committed `vectors.json` without needing Python or rapidfuzz.

## Validation Architecture

### Test Framework
| Property | Value |
|----------|-------|
| Framework | Go stdlib `testing` + `testing/quick` (root); godog + goleak + testify (`tests/bdd/` only) |
| Config file | none (Go convention) |
| Quick run command | `go test -race -shuffle=on -count=1 ./...` |
| Full suite command | `make check` (fmt-check + vet + lint + verify-license-headers + verify-deps-allowlist + tidy-check + security + test + coverage + coverage-check) |

### Phase Requirements → Test Map
| Req ID | Behavior | Test Type | Automated Command | File Exists? |
|--------|----------|-----------|-------------------|-------------|
| TOKEN-01 | MongeElkanScore + MongeElkanScoreSymmetric + permittedMongeElkanInner allow-list | unit + property + fuzz + bench + BDD | `go test -run TestMongeElkan -race ./...` | ❌ Wave |
| TOKEN-02 | TokenSortRatioScore matches RapidFuzz | unit + property + fuzz + bench + BDD + cross-validation | `go test -run TestTokenSortRatio -race ./... && go test -run TestTokenRatios_CrossValidation -race ./...` | ❌ Wave |
| TOKEN-03 | TokenSetRatioScore three-way max + bug-for-bug empty-set behaviour | unit + property + fuzz + bench + BDD + cross-validation | `go test -run TestTokenSetRatio -race ./...` | ❌ Wave |
| TOKEN-04 | PartialRatioScore + PartialRatioScoreRunes three-region iteration | unit + property + fuzz + bench + BDD + cross-validation | `go test -run TestPartialRatio -race ./...` | ❌ Wave |
| TOKEN-05 | TokenJaccardScore set-Jaccard over Tokenise output | unit + property + fuzz + bench + BDD | `go test -run TestTokenJaccard -race ./...` | ❌ Wave |

**Cross-cutting verifications:**

| Verification | Test | Command |
|--------------|------|---------|
| Cross-platform determinism | `algorithms.json` golden diff on CI matrix | `make verify-determinism` |
| llms.txt sync | `ai_friendly_test.go` walks `go/ast` and asserts every exported symbol has an entry | `go test -run TestLLMsTxt ./...` |
| Coverage floors (≥95% overall, ≥90% per file, 100% public API) | `verify-coverage-floors.sh` | `make coverage-check` |
| License header on every .go | `verify-license-headers.sh` | `make verify-license-headers` |
| Runtime deps allowlist | `verify-no-runtime-deps.sh` | `make verify-deps-allowlist` |

### Sampling Rate
- **Per task commit:** `go test -race -shuffle=on -count=1 ./<files-touched>` then `make fmt-check && make lint`.
- **Per wave merge:** `make test` (full root + bdd) + `go test -run TestTokenRatios_CrossValidation -race ./...`.
- **Phase gate:** `make check` green + cross-validation green + bench.txt regenerated and committed before `/gsd-verify-work`.

### Wave 0 Gaps
- [ ] `token_indel.go` + `token_indel_test.go` — covers foundation kernel for TOKEN-02/03/04 (Wave 1, plan 06-01)
- [ ] `monge_elkan.go` + companion files — covers TOKEN-01 (Wave 5, plan 06-05 — most complex)
- [ ] `token_sort_ratio.go` + companion files — covers TOKEN-02 (Wave 1 alongside foundation)
- [ ] `token_set_ratio.go` + companion files — covers TOKEN-03 (Wave 2-4)
- [ ] `partial_ratio.go` + companion files — covers TOKEN-04 (Wave 2-4)
- [ ] `token_jaccard.go` + companion files — covers TOKEN-05 (Wave 2-4)
- [ ] `scripts/gen-token-ratio-cross-validation.py` + `testdata/cross-validation/token-ratios/vectors.json` — cross-validation corpus
- [ ] `Makefile` target `regen-token-ratio-cross-validation` — developer regeneration entry point
- [ ] `tests/bdd/features/{monge_elkan,token_sort_ratio,token_set_ratio,partial_ratio,token_jaccard}.feature` — BDD scenarios
- [ ] `props_test.go` appendage — ~30 new property tests (5 algorithms × ~6 invariants each)
- [ ] `example_test.go` appendage — 7 new `ExampleXxx` runnable godoc examples
- [ ] `llms.txt` + `llms-full.txt` per-plan sync — every new exported symbol gets an entry IN THE SAME PLAN that adds it
- [ ] `bench.txt` finalisation regeneration — includes 5 algorithm benchmarks + 3 pathological fixtures
- [ ] `testdata/golden/algorithms.json` finalisation merge — adds 5 algorithm entries
- [ ] `examples/identifier-similarity/` extension — 5 new columns

## Project Constraints (from CLAUDE.md)

| Constraint | Phase 6 Implication |
|------------|---------------------|
| Stdlib + `golang.org/x/text` only on runtime | Cross-validation rapidfuzz is developer-only — never appears in root `go.mod` |
| No cgo | All implementations are pure Go |
| Apache-2.0 throughout | Every new `.go` file gets the AxonOps Apache-2.0 header |
| No testify in root | `_test.go` files in root use stdlib `testing` only; testify confined to `tests/bdd/steps/` |
| Property tests via `testing/quick` | All Phase 6 property tests use `testing/quick` |
| Native Go fuzz | One `Fuzz*` per public function (7 total: MongeElkanScore, MongeElkanScoreSymmetric, TokenSortRatioScore, TokenSetRatioScore, PartialRatioScore, PartialRatioScoreRunes, TokenJaccardScore) |
| Cross-platform CI matrix (linux/amd64, linux/arm64, darwin/amd64, darwin/arm64, windows/amd64) | `algorithms.json` golden file diffs byte-identically across all 5 platforms |
| Releases CI-only — no local `git tag` | Phase 6 plans MUST NOT include any release-tagging steps |
| Conventional commits with issue refs, no AI attribution | Every commit message includes `(#<issue>)` and never mentions AI / LLM / Claude |
| One logical change per commit | Each plan's wave structure produces atomic commits |
| `algorithm-correctness-reviewer` agent gate | Every algorithm PR cites primary source + formula + reference vectors |
| `algorithm-licensing-reviewer` agent gate | Every algorithm PR carries Source-Origin Statement; rapidfuzz consultation noted |
| `algorithm-performance-reviewer` agent gate | Every algorithm has bench file + meets per-algorithm budget |
| `determinism-reviewer` agent gate | No map iteration on output paths; no transcendental floats; cross-platform CI green |
| `api-ergonomics-reviewer` veto authority over function names | The exact public function names (TokenSortRatioScore, MongeElkanScore, etc.) are subject to api-ergonomics-reviewer's final word — RESEARCH.md uses spec-default names but planner consults reviewer |

## Sources

### Primary (HIGH confidence)
- **Wagner, R. A., & Fischer, M. J. (1974).** "The string-to-string correction problem." *Journal of the ACM* 21(1):168-173 — LCS DP recurrence canonical source [VERIFIED via project's existing `lcsstr.go` citation + Wikipedia LCS cross-reference]
- **Monge, A. E., Elkan, C. P. (1996).** "The field matching problem: algorithms and applications." *KDD'96 Proceedings*: 267-270 — Monge-Elkan formula primary source [VERIFIED via Ryšavý & Železný 2025 paper which reproduces the formula https://ida.fel.cvut.cz/zelezny/pubs/jcb-author-version.pdf]
- **Jaccard, P. (1912).** "The distribution of the flora in the alpine zone." *New Phytologist* 11(2):37-50 — set-Jaccard origin paper (already cited in Phase 5's q_gram_jaccard.go)
- **RapidFuzz documentation (current — v3.14.5):**
  - https://rapidfuzz.github.io/RapidFuzz/Usage/distance/Indel.html — Indel formula `1 - distance/(len1+len2)` [VERIFIED 2026-05-15]
  - https://rapidfuzz.github.io/RapidFuzz/Usage/fuzz.html — ratio / partial_ratio / token_sort_ratio / token_set_ratio definitions and examples [VERIFIED 2026-05-15]
  - https://rapidfuzz.github.io/RapidFuzz/Usage/distance/LCSseq.html — LCS-subsequence reference [VERIFIED 2026-05-15]
- **RapidFuzz source code** (MIT, https://github.com/rapidfuzz/RapidFuzz/blob/main/src/rapidfuzz/fuzz_py.py):
  - `_split_sequence` — tokenisation rule (Python `seq.split()`, whitespace-only) [VERIFIED 2026-05-15]
  - `_partial_ratio_impl` — three-region iteration + char-set early-skip [VERIFIED 2026-05-15]
  - `token_set_ratio` — three-way max construction + bug-for-bug empty-set [VERIFIED 2026-05-15]
  - `token_sort_ratio` — sorted-join then ratio [VERIFIED 2026-05-15]
- **PyPI registry:** https://pypi.org/pypi/rapidfuzz/json — confirmed v3.14.5 released 2026-04-07 [VERIFIED 2026-05-15]
- **RapidFuzz issue #110:** https://github.com/rapidfuzz/RapidFuzz/issues/110 — bug-for-bug compat with fuzzywuzzy on empty token set [CITED]

### Project-internal (HIGH confidence — spec-locked)
- `.planning/phases/06-token-based-algorithms/06-CONTEXT.md` — locked decisions §1, §1b, §2, §3, §4, §5, §6
- `.planning/REQUIREMENTS.md` — TOKEN-01..TOKEN-05 traceability
- `.planning/ROADMAP.md` — Phase 6 goal + success criteria
- `docs/requirements.md` §7.3 (Token-based algorithms), §10 (Tokenise), §13.3-13.6 (determinism), §14.1 (allocation budgets), §15.3-15.4 (property + fuzz tests), line 567 (Monge-Elkan default inner), line 571 (symmetric variant), line 609-610 (PartialRatio byte+rune), line 612 (sliding-window DP deferred), line 744 (ErrInvalidAlgoID for ME self-reference)
- `q_gram.go` (Phase 5) — direct template for `token_indel.go` shared-foundation pattern
- `qgram_jaccard.go` (Phase 5) — pattern for explicit-if-else min-max (DET-06 auditability)
- `lcsstr.go` (Phase 4) — DISTINCT from `token_indel.go`'s `lcsLen` — substring vs subsequence (Pitfall 6)
- `tokenise.go` (Phase 1) — `Tokenise` consumed verbatim
- `algoid.go` (Phase 1) — AlgoID slots 13-17 reserved
- `dispatch_lcsstr.go` (Phase 4) — pattern for `var _ = func() bool { ... }()` dispatch wiring
- `ratcliff_obershelp_test.go` lines 283-356 (Phase 4) — exact pattern for cross-validation test
- `scripts/gen-ratcliff-obershelp-cross-validation.py` (Phase 4) — exact pattern for generator script
- `props_test.go` (Phase 2-5) — appendable property-test accumulator
- `example_test.go` (Phase 2-5) — append-here pattern for examples
- `tests/bdd/steps/algorithms_steps.go` (Phase 2-5) — append step.Step registration
- `Makefile` lines 197-228 — pattern for `regen-*-cross-validation` target

### Confirmatory (MEDIUM-HIGH confidence — multiple sources agree)
- Wikipedia LCS-subsequence article (https://en.wikipedia.org/wiki/Longest_common_subsequence) — Indel-LCS relation `m + n - 2·LCS` [cross-checked in 2 sources]
- "An Algorithm to Calculate the p-Value of the Monge-Elkan Distance" (Ryšavý & Železný 2025) — Monge-Elkan formula reproduction [academic peer-reviewed]
- ChairNerd (SeatGeek 2014) blog — historical context on fuzzywuzzy → RapidFuzz lineage [original SeatGeek source]
- comparator R package documentation — symmetric Monge-Elkan variants (max, geometric mean, arithmetic mean) [project chooses arithmetic mean per CONTEXT §4]

### Tertiary (LOW confidence — single source or marketing)
- None — all behavioural claims for the four RapidFuzz functions are verified directly against the MIT-licensed `fuzz_py.py` source.

## Metadata

**Confidence breakdown:**
- Standard stack: HIGH — RapidFuzz version verified against PyPI; project-internal patterns inherited from Phases 2-5
- LCS DP recurrence: HIGH — Wagner-Fischer 1974 is uncontroversial canonical
- Indel formula equivalence: HIGH — proven mathematically (Pattern 3); cross-checked against RapidFuzz docs
- TokenSort/TokenSet/PartialRatio behaviour: HIGH — verified against MIT-licensed source code, not just docs
- Monge-Elkan formula: HIGH — formula is well-documented across multiple academic sources
- Tokeniser-divergence pitfall: HIGH — verified by reading both `_split_sequence` and `tokenise.go`
- DoS-vector godoc format: HIGH — locked in CONTEXT §5
- Cross-validation script structure: HIGH — directly mirrors Phase 4 RO precedent (4 months old, working)
- Wave decomposition recommendations: MEDIUM — CONTEXT.md gives strong hints; planner has discretion

**Research date:** 2026-05-15
**Valid until:** 2026-06-14 (30 days for stable; rapidfuzz version may bump but cross-validation script's pin protects against drift)

## RESEARCH COMPLETE
