# Algorithm Catalogue

This document is the per-algorithm reference for fuzzymatch's 23-algorithm
catalogue. The authoritative formal specification — formulae, edge-case
contracts, complexity, primary-source citations, and acceptance reference
vectors — lives in [`docs/requirements.md`](requirements.md) §7. This
document expands on each algorithm with prose-friendly descriptions, intended
use cases, allocation budgets, the `WarnKind` values each algorithm can
emit through [`fuzzymatch.Validate`](#input-validation-with-fuzzymatchvalidate),
and the library-wide panic-surface enumeration.

Every algorithm constant is enumerated in [`algoid.go`](../algoid.go), and
the dispatch table is sized to the catalogue at compile time. The H2
anchors below use hyphenated lowercase casing (`#levenshtein`,
`#damerau-levenshtein-osa`, `#sorensen-dice`) so the README catalogue
table and per-algorithm godoc cross-references can deep-link without
casing skew.

## Table of Contents

- Per-algorithm reference
  - [Levenshtein](#levenshtein)
  - [Damerau-Levenshtein OSA](#damerau-levenshtein-osa)
  - [Damerau-Levenshtein Full](#damerau-levenshtein-full)
  - [Hamming](#hamming)
  - [Jaro](#jaro)
  - [Jaro-Winkler](#jaro-winkler)
  - [Strcmp95](#strcmp95)
  - [Smith-Waterman-Gotoh](#smith-waterman-gotoh)
  - [LCSStr](#lcsstr)
  - [Q-Gram Jaccard](#q-gram-jaccard)
  - [Sørensen-Dice](#sorensen-dice)
  - [Cosine](#cosine)
  - [Tversky](#tversky)
  - [Monge-Elkan](#monge-elkan)
  - [Token Sort Ratio](#token-sort-ratio)
  - [Token Set Ratio](#token-set-ratio)
  - [Partial Ratio](#partial-ratio)
  - [Token Jaccard](#token-jaccard)
  - [Soundex](#soundex)
  - [Double Metaphone](#double-metaphone)
  - [NYSIIS](#nysiis)
  - [MRA](#mra)
  - [Ratcliff-Obershelp](#ratcliff-obershelp)
- Library-wide reference
  - [Input validation with fuzzymatch.Validate](#input-validation-with-fuzzymatchvalidate)
  - [Panic surface](#panic-surface)
  - [Performance characteristics](#performance-characteristics)

---

## Levenshtein

- **Category:** character-based, edit distance.
- **AlgoID constant:** `AlgoLevenshtein`.
- **Primary source:** Levenshtein, V. I. (1965). "Binary codes capable of
  correcting deletions, insertions, and reversals." *Soviet Physics
  Doklady*, 10(8):707–710. Two-row DP optimisation per Wagner, R. A.,
  Fischer, M. J. (1974). "The string-to-string correction problem."
  *Journal of the ACM*, 21(1):168–173.
- **Cross-reference:** `docs/requirements.md` §7.1.1.

#### Function signatures

```go
func LevenshteinDistance(a, b string) int
func LevenshteinDistanceRunes(a, b string) int
func LevenshteinScore(a, b string) float64
func LevenshteinScoreRunes(a, b string) float64
```

**Recurrence** (0-indexed, `cost = 0` if `a[i] == b[j]`, else 1):

```text
D[0, j] = j                                  (insert j characters)
D[i, 0] = i                                  (delete i characters)
D[i, j] = min(
            D[i-1, j  ] + 1,                 (deletion)
            D[i,   j-1] + 1,                 (insertion)
            D[i-1, j-1] + cost,              (substitution)
          )
```

**Score normalisation:** `score = 1.0 - distance / max(len(a), len(b))`.
Both-empty → score `1.0`; one-empty → score `0.0` exactly; identical
non-empty inputs → score `1.0`.

#### Reference vectors

| `a`        | `b`        | Distance | Score                |
| ---------- | ---------- | -------- | -------------------- |
| `kitten`   | `sitting`  | 3        | `1 - 3/7 ≈ 0.5714`   |
| `saturday` | `sunday`   | 3        | `1 - 3/8 = 0.625`    |
| `""`       | `abc`      | 3        | `0.0`                |
| `abc`      | `abc`      | 0        | `1.0`                |

**WarnKind values emitted by [`Validate`](#input-validation-with-fuzzymatchvalidate)**

- `WarnEmptyInput` (cross-cutting via `AlgoIDAny`)
- `WarnPathologicallyLargeInput` (cross-cutting via `AlgoIDAny`)

#### Performance characteristics

Per `docs/requirements.md` §14.1; see also the
[Performance characteristics](#performance-characteristics) section
below:

- ASCII Short (≤ 8 chars): 0 allocations (stack-allocated two-row DP).
- ASCII Medium (≤ 50 chars): 0 allocations when the shorter dimension fits
  the 64-byte stack buffer; otherwise 2 allocations (heap-allocated DP rows).
- ASCII Long (≤ 500 chars): 2 allocations (Q7c heap-fallback scope note —
  the stack-buffer optimisation does not apply beyond `maxStackInputLen`).
- Rune variant adds 2 allocations (`[]rune(a)` + `[]rune(b)`).

**Limitations:** sensitive to position. A single inserted character at the
start of a long string drops the score sharply because every position
beyond the insertion is misaligned. Prefer Jaro-Winkler or Partial Ratio
when prefix or substring matching matters more than position-exact edit
distance.

**Intended use:** primary edit-distance metric. Best general-purpose single
algorithm for short identifier-style strings with single-character typos.

---

## Damerau-Levenshtein OSA

- **Category:** character-based, edit distance with transposition (Optimal
  String Alignment restriction).
- **AlgoID constant:** `AlgoDamerauLevenshteinOSA`.
- **Primary source:** Damerau, F. J. (1964). "A technique for computer
  detection and correction of spelling errors." *Communications of the
  ACM*, 7(3):171–176. OSA variant: Boytsov, L. (2011). "Indexing methods
  for approximate dictionary searching: comparative analysis." *ACM Journal
  of Experimental Algorithmics*, 16, Article 1.
- **Cross-reference:** `docs/requirements.md` §7.1.2.

#### Function signatures

```go
func DamerauLevenshteinOSADistance(a, b string) int
func DamerauLevenshteinOSADistanceRunes(a, b string) int
func DamerauLevenshteinOSAScore(a, b string) float64
func DamerauLevenshteinOSAScoreRunes(a, b string) float64
```

**Recurrence:** Levenshtein recurrence plus, when `i ≥ 2`, `j ≥ 2`,
`a[i-1] = b[j-2]`, and `a[i-2] = b[j-1]`:

```text
D[i, j] = min(D[i, j], D[i-2, j-2] + 1)      (transposition)
```

**OSA restriction:** each substring may participate in at most one
transposition; substrings cannot be re-edited after a transposition. This
restriction makes the OSA variant faster than the full Damerau-Levenshtein
but produces slightly different (sometimes larger) distances on inputs
where an already-transposed substring would otherwise be edited again
(e.g. `ca` / `abc` → OSA distance 3, full DL distance 2).

**Score normalisation:** identical to Levenshtein.

#### Reference vectors

| `a`    | `b`     | OSA distance | Score             |
| ------ | ------- | ------------ | ----------------- |
| `ab`   | `ba`    | 1            | `1 - 1/2 = 0.5`   |
| `ca`   | `abc`   | 3            | `0.0`             |
| `abcd` | `acbd`  | 1            | `0.75`            |

#### WarnKind values emitted

- `WarnEmptyInput` (cross-cutting via `AlgoIDAny`)
- `WarnPathologicallyLargeInput` (cross-cutting via `AlgoIDAny`)

#### Performance characteristics

- ASCII Short (≤ 8 chars): 0 allocations (stack-allocated three-row DP).
- ASCII Medium / Long: 3 allocations (heap-allocated DP rows when input
  exceeds 64 bytes per Q7c scope note).
- Rune variant: 2 additional allocations.

**Intended use:** primary choice for identifier typo detection — handles
keyboard-adjacent character swaps (`creatd_at` vs `created_at`) at one
edit instead of two.

---

## Damerau-Levenshtein Full

- **Category:** character-based, edit distance with unrestricted
  transposition (Lowrance-Wagner formulation).
- **AlgoID constant:** `AlgoDamerauLevenshteinFull`.
- **Primary source:** Lowrance, R., Wagner, R. A. (1975). "An extension of
  the string-to-string correction problem." *Journal of the ACM*,
  22(2):177–183.
- **Cross-reference:** `docs/requirements.md` §7.1.3.

#### Function signatures

```go
func DamerauLevenshteinFullDistance(a, b string) int
func DamerauLevenshteinFullDistanceRunes(a, b string) int
func DamerauLevenshteinFullScore(a, b string) float64
func DamerauLevenshteinFullScoreRunes(a, b string) float64
```

**Algorithm:** maintains a `last seen` table indexed by alphabet character.
Substrings may be re-edited after a transposition — the structural
difference from the OSA variant. Mathematically the "correct"
Damerau-Levenshtein.

**Score normalisation:** identical to Levenshtein.

#### Reference vectors

| `a`    | `b`     | Full DL distance | Score          |
| ------ | ------- | ---------------- | -------------- |
| `ca`   | `abc`   | 2                | `1 - 2/3 ≈ 0.333` |
| `ab`   | `ba`    | 1                | `0.5`          |

#### WarnKind values emitted

- `WarnEmptyInput` (cross-cutting via `AlgoIDAny`)
- `WarnPathologicallyLargeInput` (cross-cutting via `AlgoIDAny`)

#### Performance characteristics

- ASCII Short (≤ 8 chars): ≤ 1 allocation (Phase 8.5 Q8a — full O(m·n) DP
  table allocates structurally; the stack-buffer optimisation that would
  achieve 0 allocations requires a 34 KB stack frame and was judged too
  fragile).
- ASCII Medium (≤ 50 chars): wall time ~8.2 µs (Phase 8.5 Q7b — the
  Lowrance-Wagner formulation structurally requires the full O(m·n) DP
  table; the budget was revised upward from < 3 µs to match the
  implementation reality).
- Heap fallback for Long inputs documented per Q7c.

**Limitations:** the full DP table is O(m·n) memory regardless of input
shape (unlike OSA which can use two-row DP). For inputs larger than a few
hundred characters per side, prefer Levenshtein or DL-OSA. The
`WarnPathologicallyLargeInput` warning fires at 64 KiB per side as a
consumer-DoS hint — values above this are still produced but consumers
should gate at the call site for adversarial input.

**Intended use:** when correctness across pathological transposition cases
matters more than constant-factor speed.

---

## Hamming

- **Category:** character-based, equal-length distance.
- **AlgoID constant:** `AlgoHamming`.
- **Primary source:** Hamming, R. W. (1950). "Error detecting and error
  correcting codes." *Bell System Technical Journal*, 29(2):147–160.
- **Cross-reference:** `docs/requirements.md` §7.1.4.

#### Function signatures

```go
func HammingDistance(a, b string) int
func HammingDistanceRunes(a, b string) int
func HammingScore(a, b string) float64
func HammingScoreRunes(a, b string) float64
```

**Length-mismatch policy — silent-max (Phase 8.5 Q1, LOCKED):**

- `HammingDistance` returns `max(len(a), len(b))` on unequal-length input.
  Spec catches up to the shipped code; the earlier `(int, error)` shape
  with an `ErrUnequalLength` sentinel is removed.
- `HammingScore` returns `0.0` on unequal-length input.
- Consumers wanting to detect unequal-length inputs explicitly call
  [`fuzzymatch.Validate(a, b)`](#input-validation-with-fuzzymatchvalidate)
  and check for `WarnUnequalLength` (the only algorithm-scoped warning in
  the catalogue).

**Score normalisation:** `score = 1.0 - distance / len(a)` when lengths
match; `0.0` when lengths differ.

#### Reference vectors

| `a`         | `b`         | Distance | Score                 |
| ----------- | ----------- | -------- | --------------------- |
| `karolin`   | `kathrin`   | 3        | `1 - 3/7 ≈ 0.5714`    |
| `1011101`   | `1001001`   | 2        | `1 - 2/7 ≈ 0.7143`    |
| `ab`        | `abc`       | 3 (max)  | `0.0`                 |

#### WarnKind values emitted

- `WarnEmptyInput` (cross-cutting via `AlgoIDAny`)
- **`WarnUnequalLength`** — scoped to `AlgoHamming`; documents the
  silent-max policy explicitly to the caller.

#### Performance characteristics

- ASCII / Unicode Short / Medium: 0 allocations on the byte path
  (`O(n)` time, `O(1)` space).
- Rune variant: 2 allocations (`[]rune` conversions).

**Intended use:** fixed-width codes (8-character audit IDs, hex hashes,
equal-length fingerprints). Not useful for general identifier comparison.

---

## Jaro

- **Category:** character-based, name-matching with positional tolerance.
- **AlgoID constant:** `AlgoJaro`.
- **Primary source:** Jaro, M. A. (1989). "Advances in record-linkage
  methodology as applied to matching the 1985 census of Tampa, Florida."
  *Journal of the American Statistical Association*, 84(406):414–420.
- **Cross-reference:** `docs/requirements.md` §7.1.5.

#### Function signatures

```go
func JaroScore(a, b string) float64
func JaroScoreRunes(a, b string) float64
```

**Formula:** if `m = 0`, return `0.0`. Otherwise

```text
J = (m/|a| + m/|b| + (m - t/2)/m) / 3
```

where `m` is the count of matching characters within the positional
window `max(len(a), len(b))/2 - 1` and `t` is the count of transpositions
among matched pairs.

**Score normalisation:** the formula itself produces a value in `[0.0, 1.0]`.

#### Reference vectors

| `a`           | `b`           | Jaro score |
| ------------- | ------------- | ---------- |
| `MARTHA`      | `MARHTA`      | `0.9444`   |
| `DIXON`       | `DICKSONX`    | `0.7667`   |
| `JELLYFISH`   | `SMELLYFISH`  | `0.8963`   |

#### WarnKind values emitted

- `WarnEmptyInput` (cross-cutting via `AlgoIDAny`)
- `WarnPathologicallyLargeInput` (cross-cutting via `AlgoIDAny`)

#### Performance characteristics

- Stack-allocated `[256]bool` match-flag arrays for inputs under 256
  characters (zero heap allocation on the byte path).
- ASCII Long (≥ 256 chars): 2 allocations (heap-allocated match-flag
  slices per Q7c scope note).
- Rune variant: 2 allocations (`[]rune` conversions).

**Limitations:** triangle inequality does NOT hold (Jaro is not a metric).
Composite use must avoid algorithms-that-presume-metric distance.

**Intended use:** record-linkage and name matching where positional
tolerance and transposition matter more than substitution.

---

## Jaro-Winkler

- **Category:** character-based, name-matching with prefix bonus.
- **AlgoID constant:** `AlgoJaroWinkler`.
- **Primary source:** Winkler, W. E. (1990). "String comparator metrics
  and enhanced decision rules in the Fellegi-Sunter model of record
  linkage." *Proceedings of the Section on Survey Research Methods*,
  American Statistical Association: 354–359.
- **Cross-reference:** `docs/requirements.md` §7.1.6.

#### Function signatures

```go
func JaroWinklerScore(a, b string) float64
func JaroWinklerScoreRunes(a, b string) float64
```

**Formula:** `JW = J + L · p · (1 - J)` where `J` is the Jaro score, `L`
is the length of the common prefix capped at 4 characters, and `p` is
the prefix scale (canonical value `0.1`). The bonus applies only when
`J ≥ 0.7` (Winkler's canonical boost threshold).

**Constants** (declared as unexported package consts in `jarowinkler.go`):

- `winklerPrefixScale = 0.1`
- `winklerMaxPrefix = 4`
- `winklerBoostThreshold = 0.7`

#### Reference vectors

| `a`        | `b`        | Jaro     | JW score  |
| ---------- | ---------- | -------- | --------- |
| `MARTHA`   | `MARHTA`   | `0.9444` | `0.9611`  |
| `DWAYNE`   | `DUANE`    | —        | `≈ 0.840` |
| `DIXON`    | `DICKSONX` | —        | `≈ 0.813` |

#### WarnKind values emitted

- `WarnEmptyInput` (cross-cutting via `AlgoIDAny`)
- `WarnPathologicallyLargeInput` (cross-cutting via `AlgoIDAny`)

**Performance characteristics:** identical to Jaro (the Winkler prefix
bonus is O(`winklerMaxPrefix`) post-Jaro work — negligible).

**Intended use:** prefix-aligned identifier families, personal names.

---

## Strcmp95

- **Category:** character-based, refined Jaro-Winkler with similar-
  character credit and long-string bonus. **ASCII-only by design.**
- **AlgoID constant:** `AlgoStrcmp95`.
- **Primary source:** Winkler, W. E. (1994). "Advanced methods for record
  linkage." *Proceedings of the Section on Survey Research Methods*,
  American Statistical Association: 467–472. Reference SAS/C
  implementation: U.S. Census Bureau (1995), `strcmp95.c`
  (public domain).
- **Cross-reference:** `docs/requirements.md` §7.1.7.

#### Function signature

```go
func Strcmp95Score(a, b string) float64
```

**No Runes variant.** Strcmp95 is ASCII-only by design — the
similar-character table is built from ASCII letter confusions per
Winkler 1994 / Census Bureau `strcmp95.c`. For Unicode inputs,
normalise via [`fuzzymatch.Normalise`](../normalise.go) first; the
diacritic-stripping and case-folding steps fold most Unicode input down
to comparable ASCII. Phase 8.5 Q5 explicitly locks this absence.

#### Refinements over Jaro-Winkler

1. **Similar-character matching** — letters considered partially-matching
   when commonly confused, e.g. `A`/`E`, `O`/`0`. The similar-character
   table is an unexported package-level `[][]byte`, derived from the
   Winkler 1994 paper.
2. **Long-string bonus** — further boosts scores for long strings that
   share substantial common characters.

#### WarnKind values emitted

- `WarnEmptyInput` (cross-cutting via `AlgoIDAny`)
- **`WarnAllNonASCIIDropped`** — scoped to `AlgoStrcmp95`; input is
  entirely non-ASCII or collapses to empty after the ASCII-only path.

#### Performance characteristics

- ASCII Short / Medium: 0 allocations on the byte path.
- ASCII Long: 2 allocations (heap-allocated match-flag slices per Q7c).

**Intended use:** record linkage where Jaro-Winkler's basic prefix bonus
is insufficient — typically census-style name matching and survey data
deduplication.

---

## Smith-Waterman-Gotoh

- **Category:** character-based, local sequence alignment with affine
  gap penalty.
- **AlgoID constant:** `AlgoSmithWatermanGotoh`.
- **Primary sources:**
  - Smith, T. F., Waterman, M. S. (1981). "Identification of common
    molecular subsequences." *Journal of Molecular Biology*,
    147(1):195–197.
  - Gotoh, O. (1982). "An improved algorithm for matching biological
    sequences." *Journal of Molecular Biology*, 162(3):705–708.
  - Flouri, T. et al. (2015). "Are all global alignment algorithms and
    implementations correct?" *bioRxiv* 031500 — documents the Gotoh
    1982 initialisation erratum and the corrected formulation
    transcribed in the implementation.
- **Cross-reference:** `docs/requirements.md` §7.1.8.

#### Function signatures

```go
type SWGParams struct {
    Match    float64
    Mismatch float64
    GapOpen  float64
    GapExtend float64
}

func NewSWGParams() SWGParams              // populated with documented defaults
func (p SWGParams) Validate()              // panics on invariant violation (see Panic surface)

func SmithWatermanGotohScore(a, b string) float64
func SmithWatermanGotohScoreRunes(a, b string) float64
func SmithWatermanGotohScoreWithParams(a, b string, params SWGParams) float64

func SmithWatermanGotohRawScore(a, b string) float64
func SmithWatermanGotohRawScoreRunes(a, b string) float64
func SmithWatermanGotohRawScoreWithParams(a, b string, params SWGParams) float64
```

**Default parameters** (from `NewSWGParams`):

- `Match = 1.0`
- `Mismatch = -1.0`
- `GapOpen = -1.5`
- `GapExtend = -0.5`

There is intentionally no exported `SWGDefaultParams` package-level value
— callers construct fresh values via `NewSWGParams()` to avoid a
"is this read-only?" footgun.

#### Raw vs normalised surface

- The `*RawScore*` variants return the UNCLAMPED raw alignment score,
  which may be negative (two unrelated strings dominated by mismatch/gap
  penalties) or exceed `min(len(a), len(b))` when custom params produce
  `Match > 1.0`.
- The `*Score*` variants apply `clamp(raw / min(len(a), len(b)), 0, 1)`.

**Gotoh-erratum guard:** the implementation uses the corrected Flouri et
al. 2015 formulation where every border cell of M, Ix, Iy initialises to
0 for local alignment (NOT −∞, NOT the global-alignment gap-open ladder).
Cross-validation against biopython's `Bio.Align.PairwiseAligner` (mode =
`"local"`) at `testdata/cross-validation/swg/vectors.json` is the
load-bearing acceptance gate. The `one_long_gap_canary` entry pins
biopython-normalised = 0.5 — splitting a single long gap into two halves
must NOT improve the score.

#### WarnKind values emitted

- `WarnEmptyInput` (cross-cutting via `AlgoIDAny`)
- `WarnPathologicallyLargeInput` (cross-cutting via `AlgoIDAny`)

#### Performance characteristics

- ASCII Short (≤ 8 chars): 0 allocations. Stack-allocated
  `[(maxStackInputLen+1)*6]float64` buffer (3120 bytes) for the six
  rolling rows of the M / Ix / Iy DP matrices.
- ASCII Medium (≤ 50 chars): < 10 µs wall time (Phase 8.5 Q7b — the
  budget was revised upward from < 5 µs).
- ASCII Long: heap-allocated DP rows per Q7c.

**Limitations:** triangle inequality does not hold (SWG is not a metric
over the full string space). Parameter validation in `*Score` / `*RawScore`
functions is intentionally absent — nonsense params (e.g. `GapOpen > 0`,
NaN, +Inf) produce deterministic-but-meaningless results with no errors
and no panics. The `SWGParams` godoc documents expected ranges:
`Match >= 0`, `Mismatch <= 0`, `GapOpen <= GapExtend <= 0`. Callers who
mutate `SWGParams` after construction should call `params.Validate()`
to surface invalid combinations as typed-error panics (see
[Panic surface](#panic-surface)).

**Intended use:** detecting that one name is a substring or near-substring
of another (`http_request` vs `http_request_header_fields`), or that two
names share a long common middle section despite different prefixes /
suffixes.

---

## LCSStr

- **Category:** character-based, longest common substring.
- **AlgoID constant:** `AlgoLCSStr`.
- **Primary source:** Wagner, R. A., Fischer, M. J. (1974). "The
  string-to-string correction problem." *Journal of the ACM*,
  21(1):168–173. Standard dynamic-programming formulation.
- **Cross-reference:** `docs/requirements.md` §7.1.9.

#### Function signatures

```go
func LongestCommonSubstring(a, b string) string
func LongestCommonSubstringRunes(a, b string) string
func LCSStrScore(a, b string) float64
func LCSStrScoreRunes(a, b string) float64
```

**Recurrence:**

```text
D[i, j] = D[i-1, j-1] + 1   if a[i-1] = b[j-1]
        = 0                  otherwise
```

Track maximum value and ending position.

**Score normalisation:** `score = 2 · len(lcs) / (len(a) + len(b))`.
Identical → `1.0`; both-empty → `1.0` (by convention); one-empty → `0.0`;
no shared characters → `0.0`.

#### WarnKind values emitted

- `WarnEmptyInput` (cross-cutting via `AlgoIDAny`)
- `WarnPathologicallyLargeInput` (cross-cutting via `AlgoIDAny`)

#### Performance characteristics

- ASCII Short: 0 allocations (rolling 1-D buffer plus max tracking).
- ASCII Long: 2 allocations (heap-allocated DP rows per Q7c).

**Intended use:** detecting names with a long common middle
(`my_request_id` vs `your_request_handle` shares `_request_`).

---

## Q-Gram Jaccard

- **Category:** q-gram, set similarity (Jaccard index over q-gram
  multisets).
- **AlgoID constant:** `AlgoQGramJaccard`.
- **Primary sources:**
  - Ukkonen, E. (1992). "Approximate string-matching with q-grams and
    maximal matches." *Theoretical Computer Science*, 92(1):191–211.
  - Jaccard, P. (1912). "The distribution of the flora in the alpine
    zone." *New Phytologist*, 11(2):37–50.
- **Cross-reference:** `docs/requirements.md` §7.2.1.

#### Function signatures

```go
func QGramJaccardScore(a, b string, n int) float64
func QGramJaccardScoreRunes(a, b string, n int) float64
```

**Formula:** extract overlapping character n-grams from both strings,
compute `|A ∩ B| / |A ∪ B|` on the multisets.

**Parameter discipline (Phase 8.5 Q2):**

- `n < 1` panics with `ErrInvalidQGramSize` on direct calls.
- `WithQGramJaccardAlgorithm(weight, n)` returns `ErrInvalidQGramSize` at
  Scorer construction time on the same input.

**Score normalisation:** the Jaccard index is naturally in `[0.0, 1.0]`.
Both-empty multisets → `1.0` by convention; one-empty → `0.0`; identical
→ `1.0`.

#### WarnKind values emitted

- `WarnEmptyInput` (cross-cutting via `AlgoIDAny`)
- `WarnPathologicallyLargeInput` (cross-cutting via `AlgoIDAny`)

#### Performance characteristics

- ASCII Short / Medium: ≤ 4 allocations (q-gram map + count map; Phase
  8.5 Q7d adds a 25 % capacity headroom hint to the internal
  `extractQGrams` map to reduce hash-map growth).

**Limitations:** when `n > min(len(a), len(b))`, at least one of the
q-gram sets is empty; the result is `0.0` (or `1.0` if both empty). No
triangle inequality.

**Intended use:** trigram-level partial-match detection, abbreviation
handling.

---

## Sørensen-Dice

- **Category:** q-gram, set similarity (Dice coefficient over q-gram
  multisets).
- **AlgoID constant:** `AlgoSorensenDice`.
- **Primary sources:**
  - Dice, L. R. (1945). "Measures of the amount of ecologic association
    between species." *Ecology*, 26(3):297–302.
  - Sørensen, T. (1948). "A method of establishing groups of equal
    amplitude in plant sociology based on similarity of species and its
    application to analyses of the vegetation on Danish commons."
    *Kongelige Danske Videnskabernes Selskab*, 5(4):1–34.
- **Cross-reference:** `docs/requirements.md` §7.2.2.

#### Function signatures

```go
func SorensenDiceScore(a, b string, n int) float64
func SorensenDiceScoreRunes(a, b string, n int) float64
```

**Formula:** `DSC = 2|A ∩ B| / (|A| + |B|)` on q-gram multisets.

**Parameter discipline:** identical to Q-Gram Jaccard
(`n < 1` panics with `ErrInvalidQGramSize` on direct calls; Scorer
returns the same sentinel as a typed error).

**Default `n` via Scorer (LOCKED):** `n = 2` (bigrams) when called via
Scorer with no explicit `n`. Bigrams are the canonical Sørensen-Dice
parameterisation in fuzzy-search and DNA sequence similarity
applications.

#### WarnKind values emitted

- `WarnEmptyInput` (cross-cutting via `AlgoIDAny`)
- `WarnPathologicallyLargeInput` (cross-cutting via `AlgoIDAny`)

**Performance characteristics:** identical to Q-Gram Jaccard.

**Intended use:** slightly more permissive than Jaccard for
partially-overlapping sets.

---

## Cosine

- **Category:** q-gram, vector similarity (cosine of q-gram frequency
  vectors).
- **AlgoID constant:** `AlgoCosine`.
- **Primary source:** Salton, G., McGill, M. J. (1983). *Introduction to
  Modern Information Retrieval*. McGraw-Hill. (Vector-space model and
  cosine similarity for IR are textbook standard.)
- **Cross-reference:** `docs/requirements.md` §7.2.3.

#### Function signatures

```go
func CosineScore(a, b string, n int) float64
func CosineScoreRunes(a, b string, n int) float64
```

**Formula:** `cos(A, B) = (A · B) / (‖A‖ · ‖B‖)` on q-gram frequency
vectors.

**Parameter discipline:** identical to Q-Gram Jaccard.

**Determinism:** the dot-product reduction uses an FMA-defeating
double-cast (`dot = float64(qa[k] * qb[k]) + dot`) at `cosine.go:343` —
this forces intermediate rounding and prevents compiler-emitted fused
multiply-add producing platform-divergent results on arm64. See
`docs/requirements.md` §14.4 and Phase 8.5 Q11b.

#### WarnKind values emitted

- `WarnEmptyInput` (cross-cutting via `AlgoIDAny`)
- `WarnPathologicallyLargeInput` (cross-cutting via `AlgoIDAny`)

**Performance characteristics:** identical to Q-Gram Jaccard.

**Intended use:** complements Jaccard and Dice by being length-asymmetry
tolerant (a short string matching a long string with proportional q-gram
frequency scores well).

---

## Tversky

- **Category:** q-gram, asymmetric set similarity (Tversky index).
- **AlgoID constant:** `AlgoTversky`.
- **Primary source:** Tversky, A. (1977). "Features of similarity."
  *Psychological Review*, 84(4):327–352.
- **Cross-reference:** `docs/requirements.md` §7.2.4.

#### Function signatures

```go
func TverskyScore(a, b string, n int, alpha, beta float64) float64
func TverskyScoreRunes(a, b string, n int, alpha, beta float64) float64
```

**Formula:**

```text
T(A, B) = |A ∩ B| / (|A ∩ B| + α·|A − B| + β·|B − A|)
```

With `α = β = 1` reduces to Jaccard; with `α = β = 0.5` reduces to
Sørensen-Dice.

**Parameter discipline (Phase 8.5 Q2 — strict-parameter framework, all
guards locked):**

- `n < 1` panics with a typed-error wrapping `ErrInvalidQGramSize`.
- `α` or `β` is NaN or ±Inf panics with `ErrInvalidTverskyParam`.
- `α < 0`, `β < 0`, or `α + β ≤ 0` panics with `ErrInvalidTverskyParam`.
  The `α + β > 0` guard CLOSES the panic-at-Score-time escape: pre-Q2,
  `WithTverskyAlgorithm(0, 0)` constructed successfully then panicked on
  the first Score call.
- `WithTverskyAlgorithm(weight, α, β, n)` returns the same sentinels as
  typed errors at construction time.

**Symmetry:** holds when `α = β`. Asymmetric otherwise (this is
property-tested; `PropAlgorithmScore_Symmetric` excludes Tversky when
`α ≠ β`).

#### WarnKind values emitted

- `WarnEmptyInput` (cross-cutting via `AlgoIDAny`)
- `WarnPathologicallyLargeInput` (cross-cutting via `AlgoIDAny`)

**Performance characteristics:** identical to Q-Gram Jaccard.

**Intended use:** when one string is intentionally treated as a
"prototype" and the other as a "variant" (e.g. a canonical schema name
vs a candidate match). For symmetric use, `α = β` is required to keep
the Scorer composite well-defined.

---

## Monge-Elkan

- **Category:** token-based, hybrid (uses an inner character-based metric).
- **AlgoID constant:** `AlgoMongeElkan`.
- **Primary source:** Monge, A. E., Elkan, C. P. (1996). "The field
  matching problem: algorithms and applications." *Proceedings of the
  Second International Conference on Knowledge Discovery and Data
  Mining*: 267–270.
- **Cross-reference:** `docs/requirements.md` §7.3.1.

**Function signatures (Phase 8.5 Q3 — symmetric default rename, inert
opts removed):**

```go
// Symmetric variant (default; the Scorer-facing surface).
// Returns the average of the two directional Monge-Elkan scores:
//     (ME(a,b) + ME(b,a)) / 2.
func MongeElkanScore(a, b string, inner AlgoID) float64

// Asymmetric / directional variant.
func MongeElkanScoreAsymmetric(a, b string, inner AlgoID) float64
```

**Formula (asymmetric direction):**

```text
ME(A, B) = (1/|A|) · Σ_{a ∈ A} max_{b ∈ B} sim_inner(a, b)
```

Symmetric: `(ME(A,B) + ME(B,A)) / 2`.

**Default inner metric:** Jaro-Winkler (per the original paper). The
Scorer's `WithMongeElkanAlgorithm(weight, inner)` accepts any AlgoID in
the `permittedMongeElkanInner` allow-list (character, q-gram, phonetic,
and gestalt tiers). Token-tier AlgoIDs (`AlgoTokenSortRatio`,
`AlgoTokenSetRatio`, `AlgoPartialRatio`, `AlgoTokenJaccard`) and the
`AlgoMongeElkan` self-reference are rejected.

**Parameter discipline (Phase 8.5 Q4 follow-up):**

- Direct callers passing an invalid inner AlgoID receive a panic with a
  typed error wrapping `ErrInvalidInnerAlgo`.
- `WithMongeElkanAlgorithm(weight, inner)` returns `ErrInvalidInnerAlgo`
  as a typed error at construction time.

**No Runes variant.** Operates on the byte output of
[`Tokenise`](../tokenise.go), which is itself rune-aware (Phase 8.5 Q5).

#### WarnKind values emitted

- `WarnEmptyInput` (cross-cutting via `AlgoIDAny`)
- **`WarnNoTokensAfterNormalise`** — scoped to `AlgoMongeElkan` when
  `Tokenise` of one or both inputs produces an empty token list under
  `DefaultNormalisationOptions`.
- `WarnPathologicallyLargeInput` (cross-cutting via `AlgoIDAny`)

#### Performance characteristics

- Wall time dominated by the inner metric × token-count squared:
  O(|A| · |B| · cost(inner)).
- Symmetric variant doubles the constant factor over the asymmetric form.
- Medium budget: < 10 µs per call.

**Intended use:** identifier families where token-level matching matters
more than character-level matching, with the inner metric handling
intra-token similarity (e.g. `user_create_event` vs `usr_creating_evt`
— tokens align but each pair has its own similarity).

---

## Token Sort Ratio

- **Category:** token-based, sort-and-compare.
- **AlgoID constant:** `AlgoTokenSortRatio`.
- **Primary source:** SeatGeek (2014). *fuzzywuzzy* Python library,
  `fuzz.token_sort_ratio` implementation. Canonical modern reference:
  RapidFuzz documentation (Bachmann, M., 2020–present),
  <https://rapidfuzz.github.io/RapidFuzz/>. No formal academic source
  exists; this is a practical engineering pattern.
- **Cross-reference:** `docs/requirements.md` §7.3.2.

#### Function signature

```go
func TokenSortRatioScore(a, b string, opts NormalisationOptions) float64
```

**Procedure:** tokenise both strings, sort the tokens, rejoin with a
single space, then compute an Indel-based ratio
`2·LCS / (|a| + |b|)` (Longest Common Subsequence formulation).

**No Runes variant.** Operates on the byte output of
[`Tokenise`](../tokenise.go), which is itself rune-aware (Phase 8.5 Q5).

#### WarnKind values emitted

- `WarnEmptyInput` (cross-cutting via `AlgoIDAny`)
- **`WarnNoTokensAfterNormalise`** — scoped to `AlgoTokenSortRatio`.
- `WarnPathologicallyLargeInput` (cross-cutting via `AlgoIDAny`)

#### Performance characteristics

- ASCII Short / Medium: ≤ 4 allocations (token slice + sort scratch +
  Indel DP rows).

**Intended use:** comparing strings that should be equal up to token
reordering (`UserCreateEvent` vs `CreateUserEvent`).

---

## Token Set Ratio

- **Category:** token-based, set-and-compare.
- **AlgoID constant:** `AlgoTokenSetRatio`.
- **Primary source:** as Token Sort Ratio (SeatGeek `fuzzywuzzy`; modern
  reference RapidFuzz).
- **Cross-reference:** `docs/requirements.md` §7.3.3.

#### Function signature

```go
func TokenSetRatioScore(a, b string, opts NormalisationOptions) float64
```

**Procedure:** tokenise both strings, compute three sub-strings — the
intersection sorted-and-joined; intersection + difference-from-a
sorted-and-joined; intersection + difference-from-b sorted-and-joined —
then compute the maximum Indel ratio among the three pairwise
comparisons.

**Documented RapidFuzz #110 deviation:** when both inputs tokenise to the
empty set, the Indel ratio returns `1.0` (both-empty → identity by
convention). RapidFuzz issue #110 documents the equivalent behaviour;
this library matches RapidFuzz's convention exactly. The `Validate`
function's `WarnNoTokensAfterNormalise` warning is the diagnostic
surface for "your token set is empty — the `1.0` value is the
convention, not a meaningful match."

**No Runes variant.**

#### WarnKind values emitted

- `WarnEmptyInput` (cross-cutting via `AlgoIDAny`)
- **`WarnNoTokensAfterNormalise`** — scoped to `AlgoTokenSetRatio`.
- `WarnPathologicallyLargeInput` (cross-cutting via `AlgoIDAny`)

#### Performance characteristics

- ASCII Short / Medium: ≤ 4 allocations.

**Intended use:** comparing strings with substantially different token
counts but a meaningful shared core (`http_request` vs
`http_request_body_payload`).

---

## Partial Ratio

- **Category:** token-based, sliding-window.
- **AlgoID constant:** `AlgoPartialRatio`.
- **Primary source:** as Token Sort Ratio (SeatGeek `fuzzywuzzy`; modern
  reference RapidFuzz `fuzz.partial_ratio`).
- **Cross-reference:** `docs/requirements.md` §7.3.4.

#### Function signature

```go
func PartialRatioScore(a, b string) float64
```

**Procedure:** slide the shorter string across the longer string, compute
the Indel ratio at each window position, return the maximum.

**No Runes variant** (Phase 8.5 Q5 removed `PartialRatioScoreRunes`). The
token-tier byte-level Indel kernel produces correct results on Unicode
inputs because each token from `Tokenise` is already a complete
code-point sequence.

#### WarnKind values emitted

- `WarnEmptyInput` (cross-cutting via `AlgoIDAny`)
- **`WarnNoTokensAfterNormalise`** — scoped to `AlgoPartialRatio`.
- `WarnPathologicallyLargeInput` (cross-cutting via `AlgoIDAny`)

#### Performance characteristics

- ASCII Short / Medium: ≤ 4 allocations; wall time scales with the
  length difference `(n - m)` × per-window-DP cost.
- O(m·n) per window position; O(n·m·(n-m)) overall in the v1 baseline
  implementation. A sliding-window optimisation is deferred to v1.x.

**Intended use:** detecting that one string contains a near-perfect match
of the other as a substring (e.g. `request_id` matching anywhere inside
`http_request_id_v2`).

---

## Token Jaccard

- **Category:** token-based, set similarity (Jaccard index over token
  sets).
- **AlgoID constant:** `AlgoTokenJaccard`.
- **Primary source:** Jaccard, P. (1912). (As Q-Gram Jaccard, applied to
  word tokens rather than character n-grams.)
- **Cross-reference:** `docs/requirements.md` §7.3.5.

#### Function signature

```go
func TokenJaccardScore(a, b string, opts NormalisationOptions) float64
```

**Procedure:** tokenise both strings; compute Jaccard index on the token
sets.

**No Runes variant.**

#### WarnKind values emitted

- `WarnEmptyInput` (cross-cutting via `AlgoIDAny`)
- **`WarnNoTokensAfterNormalise`** — scoped to `AlgoTokenJaccard`.
- `WarnPathologicallyLargeInput` (cross-cutting via `AlgoIDAny`)

#### Performance characteristics

- ASCII Short / Medium: ≤ 4 allocations.

**Intended use:** camelCase / snake_case equivalence at the word level —
particularly useful for identifier-style names where word order varies
but the word set is the same or near-same.

---

## Soundex

- **Category:** phonetic, English (Russell / Knuth). **ASCII-only.**
- **AlgoID constant:** `AlgoSoundex`.
- **Primary source:** Russell, R. C., Odell, M. K. (1918, 1922). U.S.
  Patents 1261167 and 1435663. Canonical algorithm description: Knuth,
  D. E. (1973). *The Art of Computer Programming, Volume 3: Sorting and
  Searching*, Section 6.4.
- **Cross-reference:** `docs/requirements.md` §7.4.1.

#### Function signatures

```go
func SoundexCode(s string) string
func SoundexScore(a, b string) float64
```

**Encoding rules** (summary; full table per Knuth 1973 §6.4):

- Retain the first letter.
- Map remaining letters to digit groups:
  - B / F / P / V → 1
  - C / G / J / K / Q / S / X / Z → 2
  - D / T → 3
  - L → 4
  - M / N → 5
  - R → 6
- Vowels and H / W are dropped after the first letter.
- Consecutive same-group letters are collapsed.
- Truncate or zero-pad to 4 characters.

**No Runes variant.** ASCII-only by design.

**Score normalisation:** binary `0.0` / `1.0` (Soundex codes either match
exactly or they don't).

#### Reference vectors

| Input      | Soundex code |
| ---------- | ------------ |
| `Robert`   | `R163`       |
| `Rupert`   | `R163`       |
| `Rubin`    | `R150`       |
| `Ashcraft` | `A261`       |
| `Ashcroft` | `A261`       |

#### WarnKind values emitted

- `WarnEmptyInput` (cross-cutting via `AlgoIDAny`)
- **`WarnAllNonASCIIDropped`** — scoped to `AlgoSoundex` when the input
  is entirely non-ASCII or collapses to empty after the ASCII-only path.
- `WarnPathologicallyLargeInput` (cross-cutting via `AlgoIDAny`)

#### Performance characteristics

- `SoundexCode` ASCII Short: ≤ 1 allocation (Phase 8.5 Q7b — the
  `string(stackBuf[:n])` return is structurally unavoidable without
  `unsafe.String`, which is forbidden by `go-coding-standards`).
- `SoundexScore` ASCII Short: ≤ 2 allocations (two `SoundexCode` calls
  plus a boolean equality check).

**Intended use:** rough pronunciation-equivalent matching for English
names. Pre-filter, not primary metric.

---

## Double Metaphone

- **Category:** phonetic, multi-language tolerant (Philips 2000).
  **ASCII-only.**
- **AlgoID constant:** `AlgoDoubleMetaphone`.
- **Primary source:** Philips, L. (2000). "The double-metaphone search
  algorithm." *C/C++ Users Journal*, 18(6):38–43. Reference
  implementation: original C code by Lawrence Philips, public domain.
- **Cross-reference:** `docs/requirements.md` §7.4.2.

#### Function signatures

```go
func DoubleMetaphoneKeys(s string) (primary, secondary string)
func DoubleMetaphoneScore(a, b string) float64
```

**Score normalisation:** binary `0.0` / `1.0`. Returns `1.0` if any
pairwise combination of `(a.primary, a.secondary)` matches any pairwise
combination of `(b.primary, b.secondary)`.

**Patent screen:** Metaphone 3 (U.S. Patent 7,440,941) is explicitly
EXCLUDED from this library; see `docs/faq.md` for the rationale. Double
Metaphone itself is public-domain.

**No Runes variant.** ASCII-only by design.

#### Reference vectors

| Input       | Keys           |
| ----------- | -------------- |
| `Smith`     | `(SM0, XMT)`   |
| `Schmidt`   | `(XMT, SMT)`   |
| `Catherine` | `(K0RN, KTRN)` |
| `Katherine` | `(K0RN, KTRN)` |

`Smith` / `Schmidt` share `XMT` → `DoubleMetaphoneScore = 1.0`.

#### WarnKind values emitted

- `WarnEmptyInput` (cross-cutting via `AlgoIDAny`)
- **`WarnAllNonASCIIDropped`** — scoped to `AlgoDoubleMetaphone`.
- `WarnPathologicallyLargeInput` (cross-cutting via `AlgoIDAny`)

#### Performance characteristics

- ASCII Short: ≤ 2 allocations (Phase 8.5 Q7a — the primary/secondary
  key accumulators moved from `strings.Builder` to `[dmMaxLen]byte` plus
  length counters; the residual 2 allocations are the primary +
  secondary return-string heap escapes).
- Wall time: ~30 % faster than the Phase 7 baseline; bytes-per-op down
  33 % on the Schmidt benchmark.

**Intended use:** robust pronunciation-equivalent matching for
English-language names of various ethnic origins.

---

## NYSIIS

- **Category:** phonetic, English (Taft 1970). **ASCII-only.**
- **AlgoID constant:** `AlgoNYSIIS`.
- **Primary source:** Taft, R. L. (1970). *Name search techniques*. New
  York State Identification and Intelligence System, Special Report
  No. 1. Albany, NY.
- **Cross-reference:** `docs/requirements.md` §7.4.3.

#### Function signatures

```go
func NYSIISCode(s string) string
func NYSIISScore(a, b string) float64
```

**Score normalisation:** binary `0.0` / `1.0`.

**No Runes variant.** ASCII-only by design.

#### Reference vectors

| Input    | NYSIIS code |
| -------- | ----------- |
| `Robert` | `RABAD`     |
| `Brown`  | `BRAN`      |
| `Browne` | `BRAN`      |

#### WarnKind values emitted

- `WarnEmptyInput` (cross-cutting via `AlgoIDAny`)
- **`WarnAllNonASCIIDropped`** — scoped to `AlgoNYSIIS`.
- `WarnPathologicallyLargeInput` (cross-cutting via `AlgoIDAny`)

#### Performance characteristics

- `NYSIISCode` ASCII Short: ≤ 1 allocation (same `string(stackBuf[:n])`
  constraint as Soundex).
- `NYSIISScore` ASCII Short: ≤ 2 allocations.

**Intended use:** higher-accuracy alternative to Soundex for English
names, particularly useful for surname matching.

---

## MRA

- **Category:** phonetic, English Match Rating Approach (NBS TN 943).
  **ASCII-only.**
- **AlgoID constant:** `AlgoMRA`.
- **Primary source:** Moore, G. B., Kuhns, J. L., Trefftzs, J. L.,
  Montgomery, C. A. (1977). *Accessing individual records from personal
  data files using non-unique identifiers*. National Bureau of Standards
  (later NIST), Technical Note 943.
- **Cross-reference:** `docs/requirements.md` §7.4.4.

#### Function signatures

```go
func MRACode(s string) string
func MRACompare(a, b string) (matched bool, simScore int)
func MRAScore(a, b string) float64
```

**`MRACompare` tuple convention:** returns `(matched, simScore)` where
`simScore` is the integer 0–6 raw similarity score and `matched` is the
result of comparing `simScore` against the canonical MRA length-dependent
threshold. The non-idiomatic tuple shape is retained for compatibility
with the literature; the convention is documented inline in the godoc
and is locked for v1.0.

**Score normalisation:** binary `0.0` / `1.0` (the underlying score is
0–6 integer; binary at the threshold).

**No Runes variant.** ASCII-only by design.

#### WarnKind values emitted

- `WarnEmptyInput` (cross-cutting via `AlgoIDAny`)
- **`WarnAllNonASCIIDropped`** — scoped to `AlgoMRA`.
- `WarnPathologicallyLargeInput` (cross-cutting via `AlgoIDAny`)

#### Performance characteristics

- `MRACode` ASCII Short: ≤ 1 allocation.
- `MRACompare` ASCII Short: ≤ 2 allocations.
- `MRAScore` ASCII Short: ≤ 2 allocations.

**Intended use:** highly-tuned phonetic match for English-language
surnames in record-linkage contexts.

---

## Ratcliff-Obershelp

- **Category:** gestalt, recursive longest-common-substring
  (Ratcliff-Metzener).
- **AlgoID constant:** `AlgoRatcliffObershelp`.
- **Primary source:** Ratcliff, J. W., Metzener, D. E. (1988). "Pattern
  matching: the gestalt approach." *Dr. Dobb's Journal*, 13(7):46–51.
- **Cross-reference:** `docs/requirements.md` §7.5.1.

#### Function signatures

```go
func RatcliffObershelpScore(a, b string) float64
func RatcliffObershelpScoreRunes(a, b string) float64
```

**Formula:** `score = 2·M / (|a| + |b|)` where `M` is the sum of lengths
of all matched substrings found by recursive longest-common-substring
decomposition.

**Asymmetric by design (LOCKED OQ-1 2026-05-14, re-confirmed Phase 8.5
Q6a 2026-05-17):**

`RatcliffObershelpScore(a, b)` is **not** guaranteed to equal
`RatcliffObershelpScore(b, a)`. The leftmost-tie-break rule on the
longest-common-substring split produces a directional decomposition that
matches Python `difflib.SequenceMatcher.ratio()` byte-for-byte. This is
intentional — the library ships true difflib equivalence, not a
symmetrised approximation. Consumers wanting a symmetric variant compute
`(RatcliffObershelpScore(a, b) + RatcliffObershelpScore(b, a)) / 2` at
the call site.

The general property-test set (`PropAlgorithmScore_Symmetric`) excludes
`AlgoRatcliffObershelp`. The exception is documented in
[`.claude/skills/algorithm-correctness-standards/SKILL.md`](../.claude/skills/algorithm-correctness-standards/SKILL.md)
§ "Symmetric algorithms".

**Score normalisation:** `[0.0, 1.0]`. Both-empty → `1.0`;
one-empty → `0.0`.

**Cross-validation:** against Python `difflib(autojunk=False)` on
canonical inputs (load-bearing acceptance test).

#### WarnKind values emitted

- `WarnEmptyInput` (cross-cutting via `AlgoIDAny`)
- `WarnPathologicallyLargeInput` (cross-cutting via `AlgoIDAny`)

#### Performance characteristics

- ASCII Short: ≤ 4 allocations (Phase 8.5 Q8d — the `roFindLongestMatch`
  recursive decomposition allocates per recursion level; the stack-buffer
  pool that would achieve a tighter budget was judged too complex for
  the marginal gain).
- Average complexity O(n·m); worst case O(n²·m) or O(n·m²). Documented
  as such; not used in tight loops without consideration of input size.

**Intended use:** general-purpose human-perceived-similarity scoring.
Often considered the closest mechanical metric to human similarity
judgement for arbitrary strings.

---

## Input validation with fuzzymatch.Validate

`fuzzymatch.Validate(a, b string) []Warning` is the consumer-facing
diagnostic surface that reports problematic-but-non-fatal input shapes
**before** scoring runs. It is the recommended companion to any Scorer
or direct algorithm call in code paths where the inputs originate from
untrusted sources (user submissions, scraped data, parsed configuration).

### Why use Validate

The comparison-data leniency contract (`docs/requirements.md` §6.A) means
algorithms always produce a value on any input — no panics, no errors,
even on garbage bytes. `Validate` is the diagnostic that tells consumers
whether the value they got is meaningful. For example:

- Two empty inputs score `1.0` from every algorithm (both empty →
  identity by convention). Without `Validate`, this looks like a perfect
  match.
- An input that is entirely non-ASCII collapses to an empty Soundex code,
  so two unrelated non-ASCII inputs both encode to the empty key and
  score `1.0`. `Validate` surfaces this as `WarnAllNonASCIIDropped`.
- A 1 MB string passed to Damerau-Levenshtein Full allocates a 1 TB DP
  table. The algorithm still produces a value (eventually), but
  `Validate` flags this with `WarnPathologicallyLargeInput` so consumers
  can gate at the call site.

### Public API

```go
// Validate inspects the two inputs and returns warnings describing
// problematic-but-non-fatal input shapes. Returns nil if no warnings
// apply. Safe for concurrent use; never panics; never returns an error.
func Validate(a, b string) []Warning

// Warning carries one diagnostic emitted by Validate.
type Warning struct {
    Algorithm AlgoID    // affected algorithm, or AlgoIDAny for cross-cutting warnings
    Kind      WarnKind  // discriminator
    Detail    string    // British-English human-readable elaboration
}

// WarnKind discriminator (declared via iota+1 so zero value is "unset").
type WarnKind int

const (
    WarnEmptyInput               WarnKind = iota + 1
    WarnUnequalLength            // Hamming family — silent-max policy applies
    WarnNoTokensAfterNormalise   // token-tier algorithms — empty token sets
    WarnAllNonASCIIDropped       // ASCII-only algorithms — input collapses
    WarnPathologicallyLargeInput // O(m·n) DP risk
)

// String returns the CamelCase form ("EmptyInput", "UnequalLength", …)
// matching the AlgoID.String() naming convention (Phase 8.5 Q6b).
func (k WarnKind) String() string

// WarnKinds returns all defined WarnKind values in iota order, for
// deterministic enumeration (test helpers, doc generators).
func WarnKinds() []WarnKind

// AlgoIDAny is the cross-cutting sentinel scope for warnings that
// affect every algorithm (WarnEmptyInput, WarnPathologicallyLargeInput).
// It is declared as -2 so it does not collide with the iota AlgoID
// constants (which start at 0) or with -1 (reserved by existing
// out-of-range-invalid AlgoID tests).
const AlgoIDAny AlgoID = -2
```

### Recommended usage pattern

```go
warnings := fuzzymatch.Validate(a, b)
if len(warnings) > 0 {
    for _, w := range warnings {
        log.Printf("input quality warning: %s (%s): %s",
            w.Kind, w.Algorithm, w.Detail)
    }
}
score := fuzzymatch.DefaultScorer().Score(a, b)
```

The validate-then-score pattern is the recommended idiom for code paths
that audit input quality. Consumers who do not run `Validate` still get
the lenient algorithm contract (no panic, no error) but lose visibility
into degraded-input scores.

### Per-WarnKind semantics

- **`WarnEmptyInput`** — `a == ""` or `b == ""` (or both). Affects every
  algorithm; emitted once with `Algorithm = AlgoIDAny`.
- **`WarnUnequalLength`** — `len(a) != len(b)`. Emitted with
  `Algorithm = AlgoHamming`; documents the silent-max policy.
- **`WarnNoTokensAfterNormalise`** — after applying
  `DefaultNormalisationOptions` and `Tokenise`, one or both inputs
  produce an empty token list. Emitted once per token-tier algorithm:
  `AlgoMongeElkan`, `AlgoTokenSortRatio`, `AlgoTokenSetRatio`,
  `AlgoPartialRatio`, `AlgoTokenJaccard`.
- **`WarnAllNonASCIIDropped`** — input contains characters but is
  entirely non-ASCII (or becomes empty after ASCII-only normalisation).
  Emitted once per ASCII-only algorithm: `AlgoStrcmp95`, `AlgoSoundex`,
  `AlgoDoubleMetaphone`, `AlgoNYSIIS`, `AlgoMRA`.
- **`WarnPathologicallyLargeInput`** — `max(len(a), len(b)) >
  validatePathologicalThreshold` (currently 65 536 bytes / 64 KiB).
  Emitted once with `Algorithm = AlgoIDAny`; the threshold is
  intentionally generous because the warning is consumer-DoS guidance,
  not a hard limit.

### Determinism and concurrency

`Validate` is pure: no goroutines, no channels, no mutexes, no I/O. It
is safe for concurrent use from any number of goroutines. Output is
sorted by `(Algorithm, Kind)` via `sort.SliceStable` with a complete
sort key, so two calls with the same inputs always return byte-identical
slices.

### Forward compatibility

The `WarnKind` enum may grow in v1.x with additional constants.
Consumers must treat unrecognised values as ignorable (the `Detail`
field carries human-readable context). The existing constants are
stable across patch versions.

### Cross-references

`Validate` is documented across the six required surfaces per
`.claude/skills/documentation-standards/SKILL.md` §
"Consumer-facing validation and diagnostics features":

1. README Quick Start / Common Patterns.
2. This `docs/algorithms.md` dedicated section (and per-algorithm
   `WarnKind` subsections above).
3. Per-algorithm godoc cross-references to `fuzzymatch.Validate` (every
   primary algorithm `.go` file).
4. `llms.txt` + `llms-full.txt` inventory.
5. User-guide section (`docs/best-practices.md`).
6. Runnable `examples/validate-input-quality/` program.

---

## Panic surface

The fuzzymatch library follows the data-vs-parameter framework
(`docs/requirements.md` §6.A): **comparison-data inputs** (the strings
being compared) are lenient — algorithms never panic on string
content — while **parameter inputs** (q-gram size `n`, Tversky `α` / `β`,
Monge-Elkan `inner` AlgoID, Scorer threshold / weight, SWG params) are
strict and surface programmer errors via typed-error panic values.

Every panic value is a wrapped `error` so consumers can discriminate
via `errors.Is` on a recovered panic:

```go
defer func() {
    if r := recover(); r != nil {
        if err, ok := r.(error); ok && errors.Is(err, fuzzymatch.ErrInvalidQGramSize) {
            // programmer error — log, swap defaults, or re-panic
        }
    }
}()
score := fuzzymatch.QGramJaccardScore("a", "b", 0)  // panics with ErrInvalidQGramSize
```

### Public functions that may panic

| Function                                  | Sentinel(s) wrapped by the panic                          | Trigger                                                                                          |
| ----------------------------------------- | ---------------------------------------------------------- | ------------------------------------------------------------------------------------------------ |
| `QGramJaccardScore` / `*Runes`            | `ErrInvalidQGramSize`                                      | `n < 1`                                                                                          |
| `SorensenDiceScore` / `*Runes`            | `ErrInvalidQGramSize`                                      | `n < 1`                                                                                          |
| `CosineScore` / `*Runes`                  | `ErrInvalidQGramSize`                                      | `n < 1`                                                                                          |
| `TverskyScore` / `*Runes`                 | `ErrInvalidQGramSize` or `ErrInvalidTverskyParam`          | `n < 1`; or `α` / `β` is NaN / ±Inf; or `α < 0`, `β < 0`, `α + β ≤ 0`                            |
| `MongeElkanScore` / `MongeElkanScoreAsymmetric` | `ErrInvalidInnerAlgo`                              | `inner` is the zero AlgoID, out-of-range, `AlgoMongeElkan` (self-reference), or a token-tier AlgoID |
| `(SWGParams).Validate()`                  | `ErrInvalidSWGParam`                                       | Mutated `SWGParams` violates the documented `Match >= 0`, `Mismatch <= 0`, `GapOpen <= GapExtend <= 0` invariants. Direct callers who construct `SWGParams` via `NewSWGParams` and never mutate the value never need to call `Validate()`. |
| `NewSWGParams`                            | `ErrInternalInvariantViolated`                             | Library-internal self-test on the default constants fails. Never fires in correct usage; surfaces only if a build-time `-ldflags` injection corrupts the defaults. |
| `DefaultScorer()`                         | `ErrInternalInvariantViolated`                             | Internal construction of the default-Scorer composition fails. Never fires in correct usage; surfaces only on a library bug. Consumers seeing this should file an issue. |

### Sentinels NEVER returned as panic values

The following sentinels are typed-error return values (Scorer option
functions), not panic values:

- `ErrEmptyScorer` — `NewScorer` called with no algorithms.
- `ErrInvalidWeight` — option function passed `weight ≤ 0`.
- `ErrInvalidThreshold` — `WithThreshold` passed NaN or value outside
  `[0.0, 1.0]` (±Inf rejected by the range check itself).
- `ErrInvalidAlgoID` — option function passed an out-of-range AlgoID
  (renamed from `ErrInvalidAlgorithm` in Phase 8.5 — Gap 4 resolution).

The Scorer construction path is the recommended entry point for any code
that may receive parameters from untrusted configuration (CLI flags,
environment variables, YAML). Direct algorithm calls are appropriate
when parameters are hard-coded at the call site.

### What does NOT panic

- Any algorithm `Score` / `Distance` / `Code` / `Keys` function called
  with arbitrary string inputs. The library never panics on string
  content, including invalid UTF-8, embedded NULs, lone surrogates, and
  very long strings.
- `Scorer.Score`, `Scorer.ScoreAll`, `Scorer.Match` after successful
  construction. The Scorer is immutable after construction; all
  parameter validation has already happened.
- `fuzzymatch.Validate(a, b)` — never panics, never returns an error, by
  contract per §6.A.

---

## Performance characteristics

Per-algorithm performance budgets are pinned in `docs/requirements.md`
§14.1 (per-algorithm) and §14.2 (Scorer composite). The benchmark
suite (`*_bench_test.go` per algorithm; output committed to
`bench.txt`) exercises Short (≤ 8 chars), Medium (≤ 50 chars), and Long
(≤ 500 chars) regimes; CI runs `benchstat` against the last tagged
release and fails the build on a > 10 % regression on any benchmark.

### Cross-algorithm allocation budget ceiling (Phase 8.5 Q14a)

**Each algorithm call allocates at most 5 MB / call for inputs up to
1 MB.** Cross-algorithm caching of tokenisation results — where two
token-tier algorithms in the same Scorer would share a single
`Tokenise(a)` / `Tokenise(b)` result — is deferred to v1.x post-1.0
(GitHub issue [#2](https://github.com/axonops/fuzzymatch/issues/2)). The
v1.0 ceiling lets consumers reason about worst-case memory usage at
call sites that feed adversarial inputs; the
`WarnPathologicallyLargeInput` warning fires above 64 KiB per side as a
gate hint at the call site.

See `docs/requirements.md` §14 for the per-input-size budget table and
§14.4 for the benchstat regression-detection methodology.

### ASCII fast paths

Every character-based algorithm exposes both byte and rune variants:

- `XxxScore(a, b string) float64` — byte-level, optimised for ASCII
  input.
- `XxxScoreRunes(a, b string) float64` — rune-level, correct for
  multi-byte UTF-8.

For ASCII-only input ≤ 64 bytes per side, the byte variants achieve 0
allocations via stack-allocated DP buffers. For longer ASCII input or
non-ASCII input on the byte path, the implementation falls back to
heap-allocated DP rows (Phase 8.5 Q7c documented this scope across
Levenshtein, DL-OSA Unicode-Short, Jaro, Jaro-Winkler, Strcmp95,
LCSStr — each algorithm's godoc carries the relevant scope note).

The token tier (Monge-Elkan, Token Sort / Set / Partial Ratio, Token
Jaccard) and the phonetic tier (Soundex, Double Metaphone, NYSIIS, MRA)
ship only the byte-string entry points. `Tokenise` is itself rune-aware
(Phase 8.5 Q8b adds an ASCII fast path that returns zero-copy
substrings when `opts.Lowercase == false`); phonetic algorithms are
ASCII-only by design and emit `WarnAllNonASCIIDropped` when their input
collapses to empty after the ASCII-only path.

### Determinism

Algorithm scores are byte-identical across the supported platform
matrix (`linux/amd64`, `linux/arm64`, `darwin/amd64`, `darwin/arm64`,
`windows/amd64`). This is verified by a golden-file test in CI. Key
defences:

- **FMA-defeating double-cast** at `cosine.go:343` and `scorer.go:380`
  (Phase 8.5 Q11b) prevents compiler-emitted fused multiply-add
  producing platform-divergent results on arm64.
- **No map iteration on output paths** (`docs/requirements.md` §13.4).
  Cosine intersection iterates over sorted keys; Scorer `ScoreAll`
  returns a map (contents deterministic; iteration order documented as
  consumer responsibility per §13.4).
- **No transcendental math operations** in algorithm hot paths
  (`docs/requirements.md` §13.3). Only `+`, `-`, `*`, `/`,
  `math.Sqrt`, `math.Abs`, `math.Min`, `math.Max` are used.
- **Left-to-right sum reductions only** — never `sync/atomic` floats or
  chunked parallel reductions.

See `docs/requirements.md` §13 and the
[`determinism-standards`](../.claude/skills/determinism-standards/SKILL.md)
skill for the full discipline.
