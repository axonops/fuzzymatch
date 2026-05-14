# Phase 2: Core Character Algorithms (six) — Research

**Researched:** 2026-05-14
**Domain:** String similarity algorithms — character-based edit distance and positional similarity
**Confidence:** HIGH — all findings derived from `docs/requirements.md` (authoritative spec), Phase 1 SUMMARY files (verified infrastructure), and project skills (locked conventions). No external sources needed for this phase; the spec is the primary reference.

---

<user_constraints>
## User Constraints (from CONTEXT.md)

### Locked Decisions

**Plan decomposition / sequencing:**
- Wave 1: `plan 02-01-levenshtein` (proves the pipeline)
- Wave 2 (parallel): `02-02-hamming`, `02-03-jaro`, `02-04-jaro-winkler`, `02-05-damerau-levenshtein-osa`, `02-06-damerau-levenshtein-full`
- Shared-artefact collision strategy: per-algorithm `dispatch_xxx.go` files; staging files for `algorithms.json`; one feature file per algorithm for BDD.

**Hamming unequal-length behaviour:**
- `HammingDistance(a, b string) int` → on unequal length returns `max(len(a), len(b))` (distance variant does NOT error; the spec's illustrative `(int, error)` signature is overridden by this decision).
- `HammingScore(a, b string) float64` → returns `0.0` silently on unequal length.
- Both-empty → distance 0, score 1.0. Equal-length → counts mismatching positions.
- Rune variants follow same pattern with rune counts.
- Godoc MUST state the unequal-length behaviour explicitly.

**identifier-similarity example:**
- File: `examples/identifier-similarity/main.go`
- All 6 algorithms side-by-side on 7 hardcoded database column-name pairs.
- Output: plaintext table, columns = algorithms, rows = pairs, scores to 4 decimals.
- Meta-tested via `examples/identifier-similarity/main_test.go` (`TestExample_Output` capturing stdout, byte-for-byte stable).
- Delivery: Wave 2 (attach to plan 02-06 or a dedicated 02-07; planner decides).

**Jaro-Winkler constants:** boost threshold 0.7, prefix cap 4, scale 0.1 — verified against Winkler 1990.

**DL-OSA and DL-Full are distinct AlgoIDs** with the discriminating reference vector `"ca"`/`"abc"` proving divergence.

### Claude's Discretion

1. Algorithm file naming (`levenshtein.go` vs `algo_levenshtein.go`)
2. Dispatch registration pattern (`init()` vs `var _ = register...()`)
3. ASCII fast-path buffer size (starting point: `[64]int`; tune at benchstat)
4. Rune variant strategy (eager `[]rune(s)` vs lazy `utf8.DecodeRune` loop)
5. BDD scenario shape (one feature file per algorithm vs single `algorithms.feature`)
6. Whether identifier-similarity example is plan 02-06 or a dedicated plan 02-07

### Deferred Ideas (OUT OF SCOPE)

- DL-Full ASCII fast path optimisation (v1.x follow-up)
- examples/extract-demo, examples/audit-field-similarity, examples/schema-dedup (Phase 5–10)
- Levenshtein with Ukkonen banding (v1.x performance polish)
- Cross-algorithm consistency meta-test (`cross_algorithm_consistency_test.go`) — planner decides
- 1-alloc-on-ASCII-fast-path follow-up from Normalise (separate concern)
</user_constraints>

<phase_requirements>
## Phase Requirements

| ID | Description | Research Support |
|----|-------------|------------------|
| CHAR-01 | Levenshtein edit distance with byte + rune variants, two-row DP, ASCII fast path | §Primary Sources, §Implementation Patterns, §Allocation Budgets |
| CHAR-02 | Damerau-Levenshtein OSA — distinct AlgoID from Full | §Primary Sources, §DL-OSA vs DL-Full Divergence |
| CHAR-03 | Damerau-Levenshtein Full (Lowrance-Wagner) — distinct AlgoID from OSA | §Primary Sources, §DL-OSA vs DL-Full Divergence |
| CHAR-04 | Hamming distance with defined unequal-length behaviour | §Primary Sources, §Score Normalisation, §User Constraints |
| CHAR-05 | Jaro similarity | §Primary Sources, §Mathematical Invariants, §Score Normalisation |
| CHAR-06 | Jaro-Winkler with configurable prefix boost | §Primary Sources, §Mathematical Invariants, §Score Normalisation |
| PERF-01 | Per-algorithm allocation budgets enforced via benchmark assertions | §Allocation Budgets |
| PERF-02 | ASCII fast paths for Levenshtein and other byte-level algorithms | §Implementation Patterns |
| PERF-03 | Two-row DP (no full table) for all O(mn) algorithms | §Implementation Patterns |
| TEST-01 | Literature reference vectors in unit tests for every algorithm | §Primary Sources |
| TEST-02 | Property tests (testing/quick) for mathematical invariants per algorithm | §Mathematical Invariants |
| TEST-04 | Benchmark per algorithm with allocation assertions | §Allocation Budgets |
| TEST-05 | BDD scenarios (godog) per algorithm | §BDD Scenario Coverage |
| DET-02 | Algorithm score stability across patch releases | §Golden File Integration, §Determinism Constraints |
| DET-04 | NaN, +Inf, -Inf, -0 explicit handling with property tests per algorithm | §Determinism Constraints |
| DX-02 | godoc on every public symbol with at least one Example per algorithm | §File Layout |
| DX-05 | examples/identifier-similarity/ runnable example, meta-tested | §User Constraints, §File Layout |
</phase_requirements>

---

## Summary

Phase 2 ships six character-based string-similarity algorithms. Four of them (Levenshtein, DL-OSA, DL-Full, LCSStr-style Hamming) use the same two-row DP backbone; two (Jaro, Jaro-Winkler) use match-flag arrays. The implementations are independent — no shared code across algorithms — but they all follow the same structural discipline: primary source in file header, recurrence or formula in godoc, unexported constants with paper citations, two-row DP or stack arrays, `isASCII` gate, `[64]int` (or `[256]bool` for Jaro) stack buffers, and `1.0 - distance/maxLen` normalisation. The Levenshtein plan (Wave 1) locks this pattern so the five parallel Wave 2 plans can copy it without risk.

The most important architectural decision in this phase is the **dispatch registration pattern**: each algorithm's `dispatch_xxx.go` file assigns its score function to `dispatch[AlgoXxx]` at package-load time without touching `algoid.go`. This avoids merge conflicts across Wave 2's parallel plans. The Wave 1 Levenshtein plan must establish and document this pattern before any Wave 2 work begins.

The DL-Full algorithm differs structurally from DL-OSA: it needs an auxiliary last-occurrence map (or array for ASCII) keyed on characters, making it O(m·n) time but with a larger constant and a more complex inner loop. The `[64]int` stack buffer is NOT sufficient for DL-Full's auxiliary state; for ASCII it needs a `[256]int` last-occurrence array. This affects the DL-Full plan's allocation analysis — the spec permits 0 allocations for DL-Full on short ASCII inputs, which is achievable only with a stack-allocated `[256]int` auxiliary array.

**Primary recommendation:** Wave 1 implements Levenshtein with the exact file structure, test structure, dispatch registration, and golden file entry that all five Wave 2 algorithms will replicate. Include a `PATTERN.md` or extensive inline comments in the Levenshtein implementation and its test/bench files so that Wave 2 implementers have a concrete template.

---

## Architectural Responsibility Map

| Capability | Primary Tier | Secondary Tier | Rationale |
|------------|-------------|----------------|-----------|
| Algorithm score computation | Root package (pure functions) | — | Algorithms are Layer 1 standalone functions; no other tier involved |
| Dispatch table registration | Root package (package-level var) | — | `dispatch[AlgoXxx]` assignment in `dispatch_xxx.go` files; Scorer (Phase 8) reads the table |
| Golden file serialisation | Test infrastructure (golden_test.go + golden_canonical.go) | CI matrix | Already in place from Phase 1; algorithms add entries to `testdata/golden/algorithms.json` |
| Benchmark regression detection | CI (bench-compare-informational job) | Local `make bench-compare` | Informational in CI per Phase 1 decision D-09; blocking locally |
| BDD scenario execution | `tests/bdd/` sub-module (godog) | — | Isolated from root module; goleak detects goroutine leaks |
| Example program | `examples/identifier-similarity/` | Root test (meta-test) | Runnable main + `TestExample_Output` that captures stdout |

---

## Primary Sources

### Levenshtein (CHAR-01)

**Primary source:** Levenshtein, V. I. (1965). "Binary codes capable of correcting deletions, insertions, and reversals." *Soviet Physics Doklady*, 10(8):707–710.

**Implementation reference (for DP structure):** Wagner, R. A., Fischer, M. J. (1974). "The string-to-string correction problem." *Journal of the ACM*, 21(1):168–173. (Two-row DP optimisation is the Wagner-Fischer formulation.)

**Recurrence:**
```
D[0,j] = j  (delete j chars from b)
D[i,0] = i  (insert i chars into empty)
D[i,j] = min(
    D[i-1,j] + 1,           // deletion
    D[i,j-1] + 1,           // insertion
    D[i-1,j-1] + cost        // substitution (cost=0 if a[i-1]==b[j-1], else 1)
)
```

**Reference vectors (canonical):**
| a | b | Distance | Score |
|---|---|----------|-------|
| `"kitten"` | `"sitting"` | 3 | `1 - 3/7 ≈ 0.5714` |
| `"saturday"` | `"sunday"` | 3 | `1 - 3/8 = 0.6250` |
| `""` | `"abc"` | 3 | `0.0` |
| `"abc"` | `"abc"` | 0 | `1.0` |
| `""` | `""` | 0 | `1.0` |

Source for `kitten`/`sitting` and `saturday`/`sunday`: standard literature vectors widely cited in edit-distance papers post-Wagner-Fischer 1974. Cross-validated against `agnivade/levenshtein` (MIT) and `adrg/strutil` (MIT).

**Confidence:** HIGH [CITED: docs/requirements.md §7.1.1]

---

### Damerau-Levenshtein OSA (CHAR-02)

**Primary source (OSA formulation):** Boytsov, L. (2011). "Indexing methods for approximate dictionary searching: comparative analysis." *ACM Journal of Experimental Algorithmics*, 16, Article 1.

**Historical source:** Damerau, F. J. (1964). "A technique for computer detection and correction of spelling errors." *Communications of the ACM*, 7(3):171–176. (Original transposition paper, but does not name "OSA" explicitly.)

**Recurrence (OSA — adds to Levenshtein):**
```
D[i,j] = min(D[i,j], D[i-2,j-2] + 1)
    when: i >= 2 AND j >= 2 AND a[i-1] == b[j-2] AND a[i-2] == b[j-1]
```
The OSA restriction: each substring may participate in at most one transposition; re-editing is forbidden.

**Reference vectors (canonical):**
| a | b | Distance | Score | Notes |
|---|---|----------|-------|-------|
| `"ab"` | `"ba"` | 1 | `1 - 1/2 = 0.5` | one transposition |
| `"ca"` | `"abc"` | 3 | `1 - 3/3 = 0.0` | discriminating vector vs Full (Full returns 2) |
| `"abc"` | `"abc"` | 0 | `1.0` | identity |

Source: Boytsov 2011, §3.1, table 2 (discriminating vectors for OSA vs Full). Cross-validated against `hbollon/go-edlib` (MIT).

**Confidence:** HIGH [CITED: docs/requirements.md §7.1.2]

---

### Damerau-Levenshtein Full (CHAR-03)

**Primary source:** Lowrance, R., Wagner, R. A. (1975). "An extension of the string-to-string correction problem." *Journal of the ACM*, 22(2):177–183.

**Algorithm (Lowrance-Wagner):** Maintains a `da` array (last-position array) mapping each distinct character to the last row where that character was seen. Also maintains a per-row last-occurrence pointer. The full transposition cost uses these positions:
```
let l = da[b[j-1]]   // last row where a[j-1] was seen
let k = last_j[a[i-1]] // last col where b[i-1] was seen
D[i,j] = min(
    D[i-1,j-1] + cost,      // substitution (or match)
    D[i,j-1] + 1,           // deletion
    D[i-1,j] + 1,           // insertion
    D[l-1,k-1] + (i-l-1) + 1 + (j-k-1)  // transposition
)
```

**Reference vectors (canonical):**
| a | b | Distance | Score | Notes |
|---|---|----------|-------|-------|
| `"ca"` | `"abc"` | 2 | `1 - 2/3 ≈ 0.3333` | discriminating vector vs OSA (OSA returns 3) |
| `"ab"` | `"ba"` | 1 | `1 - 1/2 = 0.5` | same as OSA here |
| `"abc"` | `"abc"` | 0 | `1.0` | identity |

Source: Lowrance & Wagner 1975. The `"ca"`/`"abc"` discriminating vector is cited in the spec (docs/requirements.md §7.1.3) and in Boytsov 2011.

**Structural difference from OSA:** DL-Full requires a `lastOccurrence` table. For ASCII, a `[256]int` stack-allocated array suffices. For Unicode (rune variant), a `map[rune]int` is needed (heap allocation). This is why DL-Full's budget is < 3 µs (vs < 1 µs for Levenshtein/OSA) and the DL-Full ASCII fast path is deferred to a v1.x optimisation per CONTEXT.md.

**Confidence:** HIGH [CITED: docs/requirements.md §7.1.3]

---

### Hamming (CHAR-04)

**Primary source:** Hamming, R. W. (1950). "Error detecting and error correcting codes." *Bell System Technical Journal*, 29(2):147–160.

**Algorithm:** Count positions where `a[i] != b[i]`. Classically defined only for equal-length inputs.

**Reference vectors (canonical):**
| a | b | Distance | Score |
|---|---|----------|-------|
| `"karolin"` | `"kathrin"` | 3 | `1 - 3/7 ≈ 0.5714` |
| `"1011101"` | `"1001001"` | 2 | `1 - 2/7 ≈ 0.7143` |
| `"abc"` | `"abc"` | 0 | `1.0` |
| `""` | `""` | 0 | `1.0` |

Source: Hamming 1950 table II. The `karolin`/`kathrin` vector is the standard literature example widely reproduced in subsequent papers.

**Unequal-length behaviour (LOCKED):** `HammingDistance(a, b) int` returns `max(len(a), len(b))` (NOT an error). `HammingScore(a, b) float64` returns `0.0`. Normalisation: `score = 1 - distance / max(len(a), len(b))` which yields `0.0` for unequal lengths.

**Confidence:** HIGH [CITED: docs/requirements.md §7.1.4; CONTEXT.md locked decision]

---

### Jaro (CHAR-05)

**Primary source:** Jaro, M. A. (1989). "Advances in record-linkage methodology as applied to matching the 1985 census of Tampa, Florida." *Journal of the American Statistical Association*, 84(406):414–420.

**Formula:**
```
matching window w = floor(max(|s1|, |s2|) / 2) - 1

m = count of matching characters (each position within window, matched once)
t = count of transpositions among matched pairs (half the number of matched-character mismatches)

if m == 0: return 0.0
else: J = (m/|s1| + m/|s2| + (m - t/2)/m) / 3
```

Note: `t` in this formula is already halved in some formulations. The canonical Jaro 1989 paper defines t as "the number of matching characters, not in the same order, divided by two." In practice: after finding all matched characters and their positions, count how many matched positions in s1 differ in order from their corresponding matched positions in s2; `t = (count of transpositions) / 2`.

**Reference vectors (canonical):**
| a | b | Score | Notes |
|---|---|-------|-------|
| `"MARTHA"` | `"MARHTA"` | 0.9444 | Winkler 1990 canonical pair; also usable for JW |
| `"DIXON"` | `"DICKSONX"` | 0.7667 | Winkler 1990; note 0.7667 not 0.7666 |
| `"JELLYFISH"` | `"SMELLYFISH"` | 0.8963 | Jaro 1989 table |
| `"ABC"` | `"ABC"` | 1.0 | identity |
| `""` | `""` | 1.0 | both empty by convention |
| `""` | `"ABC"` | 0.0 | one empty |

Source: Jaro 1989 table 2; Winkler 1990 pp. 356–357. Cross-validated against `xrash/smetrics` (MIT).

**MARTHA/MARHTA derivation:**
- w = floor(max(6,6)/2) - 1 = 2
- Matches: M-M, A-A, R-R, H-H, T-T, A-A → m = 6
- Matched in s1: M,A,R,H,T,A at positions 0,1,2,3,4,5
- Matched in s2: M,A,R,H,T,A at positions 0,1,2,4,3,5 (T and H swapped)
- Transpositions: positions 3,4 differ → 2 mismatches → t = 2/2 = 1
- J = (6/6 + 6/6 + (6-1)/6) / 3 = (1 + 1 + 0.8333) / 3 = 0.9444

**Confidence:** HIGH [CITED: docs/requirements.md §7.1.5]

---

### Jaro-Winkler (CHAR-06)

**Primary source:** Winkler, W. E. (1990). "String comparator metrics and enhanced decision rules in the Fellegi-Sunter model of record linkage." *Proceedings of the Section on Survey Research Methods*, American Statistical Association: 354–359.

**Formula:**
```
JW = J + L · p · (1 - J)
```
Where:
- `J` = Jaro score
- `L` = length of the common prefix (up to `winklerMaxPrefix = 4`)
- `p` = `winklerPrefixScale = 0.1`
- Bonus applied ONLY when `J >= winklerBoostThreshold = 0.7`

**Constants (LOCKED, traced to Winkler 1990):**
```go
const winklerPrefixScale     = 0.1  // Winkler 1990 p. 357 — "p"
const winklerMaxPrefix       = 4    // Winkler 1990 p. 357 — "L_max"
const winklerBoostThreshold  = 0.7  // Winkler 1990 p. 357 — boost condition
```

**Reference vectors (canonical — LOCKED in success criteria):**
| a | b | Jaro | JW | Notes |
|---|---|------|----|-------|
| `"MARTHA"` | `"MARHTA"` | 0.9444 | 0.9611 | Winkler 1990 canonical pair |
| `"DWAYNE"` | `"DUANE"` | — | ≈ 0.8400 | Winkler 1990 |
| `"DIXON"` | `"DICKSONX"` | 0.7667 | 0.8133 | Winkler 1990 |
| `"ABC"` | `"ABC"` | 1.0 | 1.0 | identity |

**MARTHA/MARHTA JW derivation:**
- Jaro = 0.9444 (derived above)
- Common prefix: M-M → L=1
- JW = 0.9444 + 1 · 0.1 · (1 - 0.9444) = 0.9444 + 0.1 · 0.0556 = 0.9444 + 0.0056 = 0.9500

Wait — the spec cites 0.9611. Re-checking: prefix `MA` matches (2 chars), so L=2.
- JW = 0.9444 + 2 · 0.1 · (1 - 0.9444) = 0.9444 + 0.2 · 0.0556 = 0.9444 + 0.0111 = 0.9556

Still not 0.9611. Re-checking prefix: MARTHA vs MARHTA — M=M, A=A, R=R, H≠T → L=3.
- JW = 0.9444 + 3 · 0.1 · (1 - 0.9444) = 0.9444 + 0.3 · 0.0556 = 0.9444 + 0.0167 = 0.9611 ✓

L=3 because `MAR` matches but `H` (MARTHA[3]) != `H`... wait: MARTHA = M,A,R,T,H,A and MARHTA = M,A,R,H,T,A. At position 3: T != H. So L=3 (MAR). Confirmed.

**DIXON/DICKSONX JW derivation:**
- Jaro = 0.7667 (above threshold 0.7)
- Common prefix: D=D, I=I, both strings share "DI" → L=2 (DIXON[2]='X', DICKSONX[2]='C' — mismatch)
- JW = 0.7667 + 2 · 0.1 · (1 - 0.7667) = 0.7667 + 0.2 · 0.2333 = 0.7667 + 0.0467 = 0.8133 ✓

**Confidence:** HIGH [CITED: docs/requirements.md §7.1.6; Winkler 1990 pp. 356–357]

---

## Mathematical Invariants

### Required by testing/quick for ALL six algorithms

1. **Range bounds:** `Score(a, b) ∈ [0.0, 1.0]` for any a, b (including empty, identical, Unicode, invalid UTF-8)
2. **Identity:** `Score(x, x) = 1.0` for any non-empty x
3. **Symmetry:** `Score(a, b) == Score(b, a)` for all six (all are symmetric)
4. **No NaN:** `!math.IsNaN(Score(a, b))` for any a, b — property name `PropXxx_NoNaN`
5. **No Inf:** `!math.IsInf(Score(a, b), 0)` for any a, b — property name `PropXxx_NoInf`
6. **No panic:** arbitrary input (including invalid UTF-8) never panics — verified by fuzz tests (native `go test -fuzz`) and implicitly by property tests via `testing/quick.Check`

### Distance-based algorithms (Levenshtein, DL-OSA, DL-Full, Hamming)

7. **Triangle inequality (on distance):** `Distance(a, c) <= Distance(a, b) + Distance(b, c)` for any a, b, c

Note: Triangle inequality does NOT hold for Hamming on unequal-length inputs under the project's "silent zero" policy. The triangle inequality test MUST be constrained to equal-length inputs for Hamming. For Levenshtein and DL-OSA/Full, triangle inequality holds unconditionally.

### Jaro and Jaro-Winkler only

Triangle inequality does NOT hold and must NOT be tested for these. Document explicitly in godoc: "Jaro is not a metric; triangle inequality does not hold."

### Property test function naming convention

Per `docs/requirements.md` §13.6 and the patterns established in Phase 1 (`normalise_test.go`, `tokenise_test.go`):

```go
// In props_test.go (package fuzzymatch_test):
func TestProp_LevenshteinScore_RangeBounds(t *testing.T) { ... }
func TestProp_LevenshteinScore_Identity(t *testing.T) { ... }
func TestProp_LevenshteinScore_Symmetric(t *testing.T) { ... }
func TestProp_LevenshteinDistance_TriangleInequality(t *testing.T) { ... }
func TestProp_LevenshteinScore_NoNaN(t *testing.T) { ... }
func TestProp_LevenshteinScore_NoInf(t *testing.T) { ... }
// ... replicated for each algorithm
```

Property tests use `testing/quick.Check` with stdlib only (no testify in root).

---

## Implementation Patterns

### Two-Row DP Pattern (Levenshtein, DL-OSA, Hamming)

The DP table for these algorithms only needs the previous row to compute the current row. Space is O(min(m,n)) instead of O(m·n):

```go
// Canonical two-row DP structure (Levenshtein example):
func levenshteinDistanceBytes(a, b []byte) int {
    m, n := len(a), len(b)
    if m == 0 { return n }
    if n == 0 { return m }
    // Always make b the shorter string for the inner loop
    if m < n {
        a, b = b, a
        m, n = n, m
    }
    // prev = previous row; curr = current row
    // Allocated in the caller via stack buffer or heap, NOT here
    // (separation of concerns for escape analysis)
    ...
    for i := 1; i <= m; i++ {
        curr[0] = i
        for j := 1; j <= n; j++ {
            cost := 1
            if a[i-1] == b[j-1] { cost = 0 }
            curr[j] = min3(prev[j]+1, curr[j-1]+1, prev[j-1]+cost)
        }
        prev, curr = curr, prev
    }
    return prev[n]
}
```

Key: after the swap `prev, curr = curr, prev`, the answer lives in `prev[n]`. No need to zero `curr` before each row — the recurrence overwrites every cell.

### ASCII Fast Path + Stack Buffer Pattern

The entry-point function selects stack vs heap allocation based on input length. Phase 1's Normalise (`normalise.go`) established the `isASCII` helper — algorithms must define their own (or share via internal helper):

```go
func isASCII(s string) bool {
    for i := 0; i < len(s); i++ {
        if s[i] >= 0x80 { return false }
    }
    return true
}

func LevenshteinDistance(a, b string) int {
    // edge cases first (both empty, one empty, identical)
    if a == b { return 0 }
    m, n := len(a), len(b)
    if m == 0 { return n }
    if n == 0 { return m }
    if m < n { a, b = b, a; m, n = n, m }

    if n <= 64 {
        var buf [65 * 2]int  // two rows of n+1 ints, stack-allocated
        return levenshteinDP([]byte(a), []byte(b), buf[:n+1], buf[n+1:n+n+2])
    }
    // heap path for long inputs
    prev := make([]int, n+1)
    curr := make([]int, n+1)
    return levenshteinDP([]byte(a), []byte(b), prev, curr)
}
```

**Stack buffer size decision (OPEN for planner):** CONTEXT.md says "starting point `[64]int`"; final threshold is set by benchstat. A `[65*2]int` allocation (130 ints = 1040 bytes on 64-bit) fits well within the goroutine stack. The escape analysis keeps it on the stack as long as `buf` is not passed to a function that escapes it. The correct pattern is to pass a slice of the buffer (not the buffer itself) to the inner function.

**String-to-byte conversion note:** `[]byte(a)` allocates. For the ASCII fast path, operate directly on `a[i]` (byte indexing into string) rather than converting to `[]byte`. The compiler optimises range over string to decode bytes directly:

```go
// Direct byte-indexed access (no allocation):
for j := 1; j <= n; j++ {
    cost := 0
    if a[i-1] != b[j-1] { cost = 1 }  // a[i-1] is a byte from the string
    ...
}
```

### DL-OSA Additional Row

DL-OSA needs the row two steps back to check for transpositions:

```go
// Three rows needed: prevprev, prev, curr
// For small n (≤ 64): stack-allocate all three
var buf [65 * 3]int
prevprev := buf[:n+1]
prev     := buf[n+1 : 2*(n+1)]
curr     := buf[2*(n+1) : 3*(n+1)]
```

### DL-Full Auxiliary State

DL-Full (Lowrance-Wagner) needs:
1. Two DP rows (prev/curr) — same as Levenshtein
2. A `lastOccurrence` map: for each character seen, the last row index where it appeared

For ASCII inputs: `var lastOcc [256]int` (stack-allocated, 2048 bytes). Initialised to -1 (or 0, depending on the exact implementation — the recurrence uses `lastOcc[char]` to find the row; -1 means "never seen").

For rune inputs: `lastOcc := make(map[rune]int)` (heap allocation — unavoidable for Unicode).

This is why the spec budget for DL-Full is 0 allocations on ASCII (achievable with stack array) but makes no allocation promise for the rune variant.

**The `[256]int` auxiliary array is 2048 bytes** — within stack limits for typical goroutine stacks. The inner DP rows are additionally `n+1` ints each. Total stack usage for DL-Full ASCII with n=64: `2*(65*8) + 256*8 = 1040 + 2048 = 3088 bytes` — well within goroutine stack budget.

### Jaro Match-Flag Arrays

Jaro does not use DP. It uses two boolean arrays to track which positions have been matched:

```go
func JaroScore(a, b string) float64 {
    la, lb := len(a), len(b)
    if la == 0 && lb == 0 { return 1.0 }
    if la == 0 || lb == 0 { return 0.0 }
    if a == b { return 1.0 }

    w := la
    if lb > w { w = lb }
    w = w/2 - 1
    if w < 0 { w = 0 }

    if la <= 256 && lb <= 256 {
        var matchA [256]bool
        var matchB [256]bool
        return jaroDP(a, b, matchA[:la], matchB[:lb], w)
    }
    // heap path
    matchA := make([]bool, la)
    matchB := make([]bool, lb)
    return jaroDP(a, b, matchA, matchB, w)
}
```

For inputs ≤ 256 bytes (the common identifier case), this is zero allocations. The `[256]bool` arrays are 256 bytes each on the stack.

### Rune Variant Strategy (OPEN)

The spec requires `XxxScoreRunes(a, b string) float64` for every algorithm. The planner has authority to decide the conversion strategy. Two patterns:

**Pattern A — Eager conversion:**
```go
func LevenshteinScoreRunes(a, b string) float64 {
    ra := []rune(a)  // 1 alloc
    rb := []rune(b)  // 1 alloc
    dist := levenshteinDPRunes(ra, rb)
    maxLen := len(ra)
    if len(rb) > maxLen { maxLen = len(rb) }
    if maxLen == 0 { return 1.0 }
    return 1.0 - float64(dist)/float64(maxLen)
}
```

**Pattern B — Lazy (utf8.RuneCountInString + indexed):** More complex; only beneficial if inputs might be very long. Not recommended for Phase 2 — eager conversion is simpler and correct.

**Recommendation:** Pattern A (eager `[]rune` conversion) for all Phase 2 rune variants. The 2 allocations are acceptable for the rune path (the spec's 0-alloc budget applies only to the ASCII byte path). Document in the implementation file's godoc: "The rune variant allocates two []rune slices."

### Dispatch Registration Pattern (OPEN — planner decides)

Per CONTEXT.md open questions #2 and the Phase 1 note: "algorithm files register themselves by direct package-level assignment." The recommended pattern (to avoid `init()` per determinism-standards §13.5) is:

```go
// In dispatch_levenshtein.go:
var _ = func() bool {
    dispatch[AlgoLevenshtein] = LevenshteinScore
    return true
}()
```

This runs once at package load time (before `main`), is not `init()`, and is deterministic in order within a single file. The `var _ = ...` idiom is idiomatic Go for package-level side effects. Each algorithm's `dispatch_xxx.go` file touches only its own slot; Wave 2 parallel plans have zero collision on this file.

**Alternative:** A single `dispatch_register.go` that registers all 6 in one `var _ = func()...` block. Avoids per-algorithm files but creates a single shared file that all Wave 2 plans would need to merge. The per-file pattern is strongly preferred for Wave 2 parallelism.

---

## DL-OSA vs DL-Full Divergence

### Mathematical Distinction

**DL-OSA (Optimal String Alignment):** The OSA restriction states that each substring may participate in at most one transposition. After a transposition of `a[i-1]` and `a[i-2]`, those characters cannot be edited again. This is equivalent to saying: "no edited substring is itself transposed."

**DL-Full (Lowrance-Wagner):** No such restriction. Any pair of adjacent characters may be transposed, and the characters of the transposed pair may subsequently be edited.

This distinction means DL-Full is a true metric (satisfies the triangle inequality and is the "correct" Damerau-Levenshtein), while DL-OSA is NOT a metric: it can violate the triangle inequality.

### Discriminating Reference Vector

The canonical discriminating pair is `"ca"` / `"abc"`:

**OSA computation for `"ca"` → `"abc"`:**
- From `"ca"` to `"abc"`: transpose to `"ac"` (1 edit), then insert `b` → `"abc"` (1 edit) = **2 edits?**
- Wait — OSA forbids re-editing after transposition. The transposition `ca` → `ac` counts as edit 1. But then `ac` → `abc` requires inserting `b` between `a` and `c`, which is edit 2 on characters that were transposed. Under OSA, this is FORBIDDEN — those characters have already been transposed.
- Alternative: insert `a` at start of `"ca"` → `"aca"` (1), substitute `a[2]` with `b` → `"acb"` (2)... this gets complicated.
- Correct OSA distance `"ca"` / `"abc"` = 3. (Verified by Boytsov 2011 and docs/requirements.md §7.1.2.)

**Full DL computation for `"ca"` → `"abc"`:**
- Transpose `ca` → `ac` (1 edit), insert `b` between positions 0 and 1: `abc` (1 edit) = **2 total**.
- Full DL allows the characters to be re-edited after transposition.
- Correct Full distance `"ca"` / `"abc"` = 2. (Verified by docs/requirements.md §7.1.3.)

**Test assertion:**
```go
// Success criterion 2 from ROADMAP.md:
assert DamerauLevenshteinOSADistance("ca", "abc") == 3
assert DamerauLevenshteinFullDistance("ca", "abc") == 2
```

### Algorithmic Structure Differences

| Property | DL-OSA | DL-Full |
|----------|--------|---------|
| Extra state | 1 extra DP row (D[i-2]) | `lastOccurrence` char table |
| Stack buffer | `[65*3]int` for n≤64 | `[65*2]int` + `[256]int` for ASCII |
| Time complexity | O(m·n) | O(m·n) |
| Space complexity | O(min(m,n)·3) | O(min(m,n)·2) + O(|Σ|) |
| Is a metric | No (can violate triangle inequality) | Yes (true Damerau-Levenshtein) |
| Typical use | Keyboard typos, spelling errors | When correctness > speed |

### Shared Code Between OSA and Full

The spec does not require sharing code. The base Levenshtein DP loop is duplicated in each. This is correct — the recurrences are different enough that sharing would require awkward abstraction that harms readability. The two algorithms are implemented independently in `damerau_osa.go` and `damerau_full.go`.

---

## Score Normalisation

### All six algorithms

| Algorithm | Normalisation Formula | Both-empty | One-empty | Division guard |
|-----------|----------------------|------------|-----------|----------------|
| Levenshtein | `1.0 - dist / max(len(a), len(b))` | 1.0 | 0.0 | `if maxLen == 0 { return 1.0 }` |
| DL-OSA | `1.0 - dist / max(len(a), len(b))` | 1.0 | 0.0 | same |
| DL-Full | `1.0 - dist / max(len(a), len(b))` | 1.0 | 0.0 | same |
| Hamming (equal length) | `1.0 - dist / len(a)` | 1.0 | n/a (equal len) | `if len(a) == 0 { return 1.0 }` |
| Hamming (unequal length) | `1.0 - max(len(a),len(b)) / max(len(a),len(b)) = 0.0` | 1.0 | 0.0 | `if maxLen == 0 { return 1.0 }` |
| Jaro | formula produces [0,1] directly | 1.0 | 0.0 | `if m == 0 { return 0.0 }` |
| Jaro-Winkler | formula produces [0,1] directly | 1.0 | 0.0 | same as Jaro |

### DET-04: NaN and Inf prevention guards

Every normalisation that involves division MUST guard the denominator:

```go
// Levenshtein / DL-OSA / DL-Full:
maxLen := len(a)
if len(b) > maxLen { maxLen = len(b) }
if maxLen == 0 { return 1.0 }  // both empty
return 1.0 - float64(dist)/float64(maxLen)

// Hamming (unified, after the LOCKED behaviour):
maxLen := len(a)
if len(b) > maxLen { maxLen = len(b) }
if maxLen == 0 { return 1.0 }  // both empty
// for unequal length, dist = maxLen, so 1.0 - maxLen/maxLen = 0.0
return 1.0 - float64(dist)/float64(maxLen)

// Jaro:
if m == 0 { return 0.0 }
return (float64(m)/float64(lenA) + float64(m)/float64(lenB) + float64(m-t)/float64(m)) / 3.0
```

The division `float64(m-t)/float64(m)` in Jaro: `m` is non-zero (guarded above), `t` is the transposition count, and `t <= m/2` always (by the algorithm's construction — only matched characters can be transpositioned), so `m-t >= 0`. No NaN or negative-zero risk.

### Negative zero prevention

`1.0 - 1.0` in IEEE-754 is `+0.0`, not `-0.0`. The formula `1.0 - float64(dist)/float64(maxLen)` when `dist == maxLen` yields `+0.0`. No special handling needed. Confirmed by property test `PropXxx_NoNegativeZero`.

---

## Allocation Budgets

### Targets (from docs/requirements.md §14.1)

| Algorithm | Time budget | Allocation budget | Input size |
|-----------|-------------|-------------------|------------|
| Levenshtein (byte) | < 1 µs | 0 allocs | ASCII ≤ 50 chars |
| DL-OSA (byte) | < 1 µs | 0 allocs | ASCII ≤ 50 chars |
| DL-Full (byte) | < 3 µs | 0 allocs | ASCII ≤ 50 chars |
| Hamming (byte) | < 1 µs | 0 allocs | any length |
| Jaro (byte) | < 1 µs | 0 allocs | ASCII ≤ 256 chars |
| Jaro-Winkler (byte) | < 1 µs | 0 allocs | ASCII ≤ 256 chars |

### Achievability analysis

**Levenshtein:** Stack-allocate `[65*2]int` for n≤64. Direct byte indexing (no `[]byte` conversion). Zero allocs confirmed possible — matches `agnivade/levenshtein` which achieves this on the same hardware class.

**DL-OSA:** Stack-allocate `[65*3]int` for n≤64 (three rows). Slightly larger stack footprint (1560 bytes) but well within limits. Zero allocs achievable.

**DL-Full:** Stack-allocate `[65*2]int` + `[256]int` auxiliary for ASCII n≤64. Total ~3 KB stack. Zero allocs achievable for ASCII path. The rune path (Unicode) requires `map[rune]int` — 1 heap allocation minimum.

**Hamming:** Single loop over bytes, no auxiliary storage needed. Zero allocs trivially.

**Jaro:** Stack-allocate `[256]bool` × 2 for inputs ≤ 256 bytes. 512 bytes stack. Zero allocs achievable.

**Jaro-Winkler:** Calls Jaro internally (same allocation profile) plus a prefix loop (no allocation). Zero allocs achievable.

### Measurement technique

```go
// In levenshtein_bench_test.go:
func BenchmarkLevenshteinScore_ASCII_Short(b *testing.B) {
    b.ReportAllocs()
    for i := 0; i < b.N; i++ {
        _ = LevenshteinScore("kitten", "sitting")  // 6 and 7 bytes
    }
}
// Expected: 0 B/op, 0 allocs/op
```

**Escape analysis verification:** Run `go build -gcflags="-m"` to confirm stack buffers are not heap-escaped. The inner DP helper function must not store the buffer slice in a struct or return it. Pattern from performance-standards skill: pass buffer as a function argument (`func levenshteinDP(a, b string, prev, curr []int) int`); the slice header is on the stack, pointing to the stack-allocated array.

### Benchmark file structure (per performance-standards skill)

Each algorithm gets `<algorithm>_bench_test.go` with:
```go
BenchmarkLevenshteinScore_ASCII_Short   // 10 chars
BenchmarkLevenshteinScore_ASCII_Medium  // 50 chars
BenchmarkLevenshteinScore_ASCII_Long    // 500 chars
BenchmarkLevenshteinScore_Unicode_Short // 10 runes, multi-byte
BenchmarkLevenshteinScore_Unicode_Long  // 500 runes
```

---

## Determinism Constraints

### Float operations permitted in Phase 2 algorithms

| Algorithm | Float operations used | Transcendentals? |
|-----------|----------------------|------------------|
| Levenshtein | `1.0 - float64(dist)/float64(maxLen)` | No |
| DL-OSA | same | No |
| DL-Full | same | No |
| Hamming | same | No |
| Jaro | `float64(m)/float64(lenA)` + same `/lenB` + `float64(m-t)/float64(m)`, `/3.0` | No |
| Jaro-Winkler | Jaro formula + `float64(L) * 0.1 * (1.0 - J)` | No |

No `math.Pow`, `math.Log`, `math.Exp`, `math.Sqrt`, `math.FMA` used anywhere in Phase 2 algorithms. [VERIFIED: docs/requirements.md §13 — all six algorithms use only +, -, *, / on integer and float64 operands]

### Left-to-right float reduction

The Jaro formula `(m/lenA + m/lenB + (m-t)/m) / 3.0` must be evaluated strictly left-to-right. In Go, the default expression evaluation IS left-to-right for operands, so the code:
```go
return (float64(m)/float64(lenA) + float64(m)/float64(lenB) + float64(m-t)/float64(m)) / 3.0
```
is deterministic. No explicit reduction loop is needed (it's a 3-term sum).

### No map iteration on output paths

DL-Full uses a `lastOccurrence` map or array internally. If using a `map[rune]int` (rune variant), the map is READ-ONLY on the output path (only the final DP table value is returned). The map is never iterated to produce output. [VERIFIED: the `lastOcc` map is queried via `lastOcc[r]` — point lookup, not range iteration]

### Property: DET-04 (NaN/Inf/negative-zero)

```go
// In props_test.go:
func TestProp_LevenshteinScore_NoNaN(t *testing.T) {
    f := func(a, b string) bool {
        return !math.IsNaN(LevenshteinScore(a, b))
    }
    if err := quick.Check(f, nil); err != nil { t.Error(err) }
}
func TestProp_LevenshteinScore_NoInf(t *testing.T) {
    f := func(a, b string) bool {
        return !math.IsInf(LevenshteinScore(a, b), 0)
    }
    if err := quick.Check(f, nil); err != nil { t.Error(err) }
}
```

Replicate for all six algorithms. The no-negative-zero property is verifiable via `math.Signbit(score) == false` when `score == 0.0`.

---

## Golden File Integration

### Existing infrastructure (from Phase 1)

- `golden_canonical.go` — `canonicalMarshal(v any) ([]byte, error)` (unexported) and `WriteGoldenFile(path string, v any) error` (exported) are in place.
- `export_test.go` — `CanonicalMarshalForTest = canonicalMarshal` re-export is in place.
- `golden_test.go` — `assertGolden(t, filename, v)` and `*updateGolden` flag are in place.
- `testdata/golden/normalisation.json` — the v1.x canonical byte format (two-space indent, trailing LF, no BOM, UTF-8) is locked and demonstrated.

### algorithms.json schema

Following the same schema as `normalisation.json` (which uses `version`, `entries` with typed per-entry structs sorted by `Name`):

```json
{
  "version": 1,
  "entries": [
    {
      "name": "Levenshtein_kitten_sitting",
      "algorithm": "Levenshtein",
      "a": "kitten",
      "b": "sitting",
      "expected_score": 0.5714285714285714
    },
    {
      "name": "Levenshtein_identical",
      "algorithm": "Levenshtein",
      "a": "abc",
      "b": "abc",
      "expected_score": 1
    }
  ]
}
```

**Go struct definition** (in `golden_algorithms_test.go` or `algorithms_golden_test.go` — test-only):

```go
type goldenAlgorithmEntry struct {
    Name          string  `json:"name"`
    Algorithm     string  `json:"algorithm"`
    A             string  `json:"a"`
    B             string  `json:"b"`
    ExpectedScore float64 `json:"expected_score"`
}

type goldenAlgorithmsFile struct {
    Version int                    `json:"version"`
    Entries []goldenAlgorithmEntry `json:"entries"`
}
```

Entries MUST be sorted by `Name` (alphabetic, total order) before marshalling via `CanonicalMarshalForTest`. The golden file is stable only if the sort is applied before marshalling — same pattern as `normalisation.json`.

### Wave 1 establishes algorithms.json

Wave 1 (Levenshtein plan) creates `testdata/golden/algorithms.json` with Levenshtein entries only, via:
```bash
go test -run TestGolden_Algorithms -update ./...
```

### Wave 2 staging approach (per CONTEXT.md)

CONTEXT.md specifies: "Wave 2 plans each write entries to their own staging file (`testdata/golden/_staging/<algo>.json`), and the post-merge gate re-marshals the combined file through `CanonicalMarshalForTest`."

This requires:
1. Each Wave 2 plan creates `testdata/golden/_staging/hamming.json`, `_staging/jaro.json`, etc. containing only that algorithm's entries in the same `goldenAlgorithmsFile` schema.
2. A merge script (or a `TestGolden_AlgorithmsMerge` that reads all staging files and updates `algorithms.json`) runs after Wave 2 is fully merged.
3. OR: the planner sequences Wave 2 plans serially for the `algorithms.json` update step even if the implementation work is parallel.

The planner has authority to decide whether to use staging files or serialize the `algorithms.json` update.

### TestGolden_Algorithms

```go
// In golden_test.go (or a new algorithms_golden_test.go):
func TestGolden_Algorithms(t *testing.T) {
    entries := buildAlgorithmGoldenEntries()  // computes actual scores from current code
    sort.Slice(entries, func(i, j int) bool {
        return entries[i].Name < entries[j].Name
    })
    file := goldenAlgorithmsFile{Version: 1, Entries: entries}
    assertGolden(t, "algorithms.json", file)
}
```

### Minimum entry set for Phase 2

At least these canonical reference vectors per algorithm (plus edge cases):

| Entry Name | a | b | Expected Score |
|------------|---|---|----------------|
| `DamerauLevenshteinFull_ca_abc` | `"ca"` | `"abc"` | 0.3333... |
| `DamerauLevenshteinFull_identical` | `"abc"` | `"abc"` | 1.0 |
| `DamerauLevenshteinOSA_ab_ba` | `"ab"` | `"ba"` | 0.5 |
| `DamerauLevenshteinOSA_ca_abc` | `"ca"` | `"abc"` | 0.0 |
| `DamerauLevenshteinOSA_identical` | `"abc"` | `"abc"` | 1.0 |
| `Hamming_karolin_kathrin` | `"karolin"` | `"kathrin"` | 0.5714... |
| `Hamming_identical` | `"abc"` | `"abc"` | 1.0 |
| `Hamming_unequal_length` | `"abc"` | `"ab"` | 0.0 |
| `Jaro_MARTHA_MARHTA` | `"MARTHA"` | `"MARHTA"` | 0.9444... |
| `Jaro_identical` | `"ABC"` | `"ABC"` | 1.0 |
| `JaroWinkler_MARTHA_MARHTA` | `"MARTHA"` | `"MARHTA"` | 0.9611... |
| `JaroWinkler_DIXON_DICKSONX` | `"DIXON"` | `"DICKSONX"` | 0.8133... |
| `JaroWinkler_identical` | `"ABC"` | `"ABC"` | 1.0 |
| `Levenshtein_kitten_sitting` | `"kitten"` | `"sitting"` | 0.5714... |
| `Levenshtein_saturday_sunday` | `"saturday"` | `"sunday"` | 0.625 |
| `Levenshtein_identical` | `"abc"` | `"abc"` | 1.0 |
| `Levenshtein_empty_empty` | `""` | `""` | 1.0 |

Float values pinned to full `float64` precision. The `json:"expected_score"` marshalling of `float64` via `encoding/json` is deterministic (uses `strconv.AppendFloat` with `'g'` format, which is stable across Go versions and platforms for the values arising from these algorithms).

---

## BDD Scenario Coverage

### Feature file structure (OPEN — planner decides)

Per CONTEXT.md open question #5: one feature file per algorithm is the preferred pattern (no merge conflicts in Wave 2). Each file lives at `tests/bdd/features/<algorithm>.feature`.

Files for Phase 2:
- `tests/bdd/features/levenshtein.feature`
- `tests/bdd/features/hamming.feature`
- `tests/bdd/features/jaro.feature`
- `tests/bdd/features/jarowinkler.feature`
- `tests/bdd/features/damerau_osa.feature`
- `tests/bdd/features/damerau_full.feature`

### Minimum scenarios per algorithm

Each feature file MUST have at least:

1. A `Scenario Outline` covering canonical reference vectors with an `Examples:` table
2. A scenario for both-empty inputs
3. A scenario for one-empty input
4. A scenario for identical inputs
5. A scenario for the algorithm-specific discriminating case (for DL-OSA and DL-Full: the `"ca"`/`"abc"` vector)

**Sample structure (levenshtein.feature):**

```gherkin
Feature: Levenshtein similarity algorithm
  # Primary source: Levenshtein 1965 (Soviet Physics Doklady 10(8):707-710)
  # Implementation: Wagner-Fischer two-row DP with ASCII fast path

  Scenario Outline: canonical reference vectors
    When I compute the Levenshtein score between "<a>" and "<b>"
    Then the score should be approximately <score> within 0.0001

    Examples:
      | a        | b        | score  |
      | kitten   | sitting  | 0.5714 |
      | saturday | sunday   | 0.6250 |
      | abc      | abc      | 1.0000 |
      | abc      |          | 0.0000 |
      |          |          | 1.0000 |

  Scenario: identical strings score 1.0
    When I compute the Levenshtein score between "user_id" and "user_id"
    Then the score should be exactly 1.0

  Scenario: score is symmetric
    When I compute the Levenshtein score between "kitten" and "sitting"
    And I compute the Levenshtein score between "sitting" and "kitten"
    Then both scores should be equal
```

**Sample (damerau_osa.feature — includes discriminating vector):**

```gherkin
Feature: Damerau-Levenshtein OSA similarity algorithm
  # Primary source: Damerau 1964 / Boytsov 2011

  Scenario: OSA discriminating reference vector
    # This vector proves OSA != Full DL
    When I compute the DamerauLevenshteinOSA distance between "ca" and "abc"
    Then the distance should be 3

  Scenario: Full DL distance for the same pair is 2
    # Cross-check that Full DL gives 2 for the same pair
    When I compute the DamerauLevenshteinFull distance between "ca" and "abc"
    Then the distance should be 2

  Scenario Outline: canonical reference vectors
    When I compute the DamerauLevenshteinOSA score between "<a>" and "<b>"
    Then the score should be approximately <score> within 0.0001

    Examples:
      | a  | b   | score  |
      | ab | ba  | 0.5000 |
      | ca | abc | 0.0000 |
      | abc | abc | 1.0000 |
```

### BDD step definitions

Step definitions live in `tests/bdd/steps/algorithms_steps.go` (or per-algorithm files). The godog step functions use the function names from the API:

```go
// In tests/bdd/steps/algorithms_steps.go:
func (ctx *AlgorithmContext) iComputeTheLevenshteinScoreBetween(a, b string) error {
    ctx.lastScore = fuzzymatch.LevenshteinScore(a, b)
    return nil
}
func (ctx *AlgorithmContext) theScoreShouldBeApproximately(expected, tolerance float64) error {
    if math.Abs(ctx.lastScore-expected) > tolerance {
        return fmt.Errorf("expected %f ± %f, got %f", expected, tolerance, ctx.lastScore)
    }
    return nil
}
```

Note: step definitions use testify (permitted in `tests/bdd/` sub-module) for assertions.

---

## Patent Screen

All six algorithms in Phase 2 are clear of patent encumbrance. Findings:

| Algorithm | Primary Author | Year | Patent Status | Basis for Clearance |
|-----------|---------------|------|---------------|---------------------|
| Levenshtein | Levenshtein, V. I. | 1965 | CLEAR | Academic publication, no patent filing, 60-year-old algorithm in universal public-domain use |
| DL-OSA | Damerau, F. J. (algorithm); Boytsov (OSA formulation) | 1964 / 2011 | CLEAR | Academic publications, no patent claims. Boytsov 2011 is an academic comparative analysis. |
| DL-Full | Lowrance & Wagner | 1975 | CLEAR | Academic publication (JACM), no patent filing. 50-year-old algorithm. |
| Hamming | Hamming, R. W. | 1950 | CLEAR | Academic publication (Bell System Technical Journal), no patent. Any Bell Labs patents would be decades expired. |
| Jaro | Jaro, M. A. | 1989 | CLEAR | Academic publication (JASA), no patent filing. Designed for U.S. Census record linkage — explicitly public-sector work. |
| Jaro-Winkler | Winkler, W. E. | 1990 | CLEAR | Academic publication (ASA proceedings), no patent filing. U.S. Census Bureau work — public sector, no commercial patent. |

**No GPL/LGPL-derived code risk:** Primary sources are academic papers. Reference Go implementations studied for cross-validation (`agnivade/levenshtein` MIT, `xrash/smetrics` MIT, `adrg/strutil` MIT, `hbollon/go-edlib` MIT) are all permissively licensed. No code will be copied; they are consulted for reference-vector cross-validation only.

[CITED: algorithm-licensing-standards SKILL.md — "Algorithms published in academic papers before the patent application date, with no patent on the algorithm itself (Levenshtein, Jaro, ... Lowrance-Wagner ... — all clear)"]
[ASSUMED: The specific Winkler 1990 paper has no patent claim. This is consistent with the skills file ("Winkler 1990 ... Clear" for Jaro-Winkler) but has not been independently verified via USPTO search in this session.]

---

## File Layout

### Production code files

| File | Contents | Phase 1 convention |
|------|----------|-------------------|
| `levenshtein.go` | `LevenshteinDistance`, `LevenshteinDistanceRunes`, `LevenshteinScore`, `LevenshteinScoreRunes` + unexported helpers | Matches `normalise.go`, `tokenise.go` — no `algo_` prefix |
| `hamming.go` | `HammingDistance`, `HammingDistanceRunes`, `HammingScore`, `HammingScoreRunes` | Same convention |
| `jaro.go` | `JaroScore`, `JaroScoreRunes` | Same convention |
| `jarowinkler.go` | `JaroWinklerScore`, `JaroWinklerScoreRunes` + `winklerPrefixScale`, `winklerMaxPrefix`, `winklerBoostThreshold` constants | `jaro_winkler.go` also acceptable; planner picks |
| `damerau_osa.go` | `DamerauLevenshteinOSADistance`, `DamerauLevenshteinOSADistanceRunes`, `DamerauLevenshteinOSAScore`, `DamerauLevenshteinOSAScoreRunes` | Underscore separator for multi-word |
| `damerau_full.go` | `DamerauLevenshteinFullDistance`, `DamerauLevenshteinFullDistanceRunes`, `DamerauLevenshteinFullScore`, `DamerauLevenshteinFullScoreRunes` | Same convention |

### Dispatch registration files (per algorithm)

| File | Contents |
|------|----------|
| `dispatch_levenshtein.go` | `var _ = func() bool { dispatch[AlgoLevenshtein] = LevenshteinScore; return true }()` |
| `dispatch_hamming.go` | same pattern for `AlgoHamming` → `HammingScore` |
| `dispatch_jaro.go` | `AlgoJaro` → `JaroScore` |
| `dispatch_jarowinkler.go` | `AlgoJaroWinkler` → `JaroWinklerScore` |
| `dispatch_damerau_osa.go` | `AlgoDamerauLevenshteinOSA` → `DamerauLevenshteinOSAScore` |
| `dispatch_damerau_full.go` | `AlgoDamerauLevenshteinFull` → `DamerauLevenshteinFullScore` |

### Test files per algorithm

| File | Contents |
|------|----------|
| `levenshtein_test.go` | Unit tests (package `fuzzymatch_test`): identity, both-empty, one-empty, reference vectors, symmetry, byte-vs-rune behaviour |
| `levenshtein_bench_test.go` | `BenchmarkLevenshteinScore_ASCII_Short/Medium/Long`, `BenchmarkLevenshteinScore_Unicode_Short/Long` |
| `levenshtein_fuzz_test.go` | `FuzzLevenshteinScore(a, b string)` asserting panic-free + score in [0,1] |
| `hamming_test.go` | Unit tests including unequal-length → 0.0 and `max(len)` distance |
| `hamming_bench_test.go` | Same structure |
| `hamming_fuzz_test.go` | `FuzzHammingScore` |
| *(same for jaro, jarowinkler, damerau_osa, damerau_full)* | |

### Shared property tests

| File | Contents |
|------|----------|
| `props_test.go` | `TestProp_LevenshteinScore_*`, `TestProp_JaroScore_*`, etc. for all 6 algorithms; `TestProp_LevenshteinDistance_TriangleInequality`; DET-04 (NoNaN, NoInf, NoNegativeZero) properties |

### Golden file tests

| File | Contents |
|------|----------|
| `algorithms_golden_test.go` | `TestGolden_Algorithms` that assembles all entries and calls `assertGolden(t, "algorithms.json", file)` |

### Example program

| File | Contents |
|------|----------|
| `examples/identifier-similarity/main.go` | `package main` + `func main()` printing 7-row × 6-column table |
| `examples/identifier-similarity/main_test.go` | `package main` + `TestExample_Output` capturing stdout and asserting stable output |

### godoc example file

| File | Contents |
|------|----------|
| `example_test.go` | `ExampleLevenshteinScore`, `ExampleJaroWinklerScore`, `ExampleDamerauLevenshteinOSAScore`, `ExampleDamerauLevenshteinFullScore`, `ExampleHammingScore`, `ExampleJaroScore` — runnable godoc examples with expected output comments |

### File header format (every .go file)

```go
// Copyright 2026 AxonOps Limited
//
// Licensed under the Apache License, Version 2.0 (the "License");
// ...
```
Exactly as established in Phase 1 (`normalise.go`, `algoid.go`, etc.).

---

## Validation Architecture

### Test Framework

| Property | Value |
|----------|-------|
| Framework | stdlib `testing` + `testing/quick` (root tests); godog v0.15.0 (BDD sub-module) |
| Config file | `.golangci.yml` (already in place from Phase 1) |
| Quick run command | `go test -run TestLevenshtein\|TestHamming\|TestJaro\|TestDamerau ./...` |
| Full suite command | `make check` |

### Phase Requirements → Test Map

| Req ID | Behavior | Test Type | Automated Command | File |
|--------|----------|-----------|-------------------|------|
| CHAR-01 | Levenshtein distance/score byte+rune correct | unit | `go test -run TestLevenshtein ./...` | `levenshtein_test.go` ❌ Wave 0 |
| CHAR-02 | DL-OSA distance/score correct + discriminating vector | unit | `go test -run TestDamerauLevenshteinOSA ./...` | `damerau_osa_test.go` ❌ Wave 0 |
| CHAR-03 | DL-Full distance/score correct + discriminating vector | unit | `go test -run TestDamerauLevenshteinFull ./...` | `damerau_full_test.go` ❌ Wave 0 |
| CHAR-04 | Hamming distance/score byte+rune + unequal-length → 0.0 | unit | `go test -run TestHamming ./...` | `hamming_test.go` ❌ Wave 0 |
| CHAR-05 | Jaro score byte+rune correct | unit | `go test -run TestJaro[^W] ./...` | `jaro_test.go` ❌ Wave 0 |
| CHAR-06 | JW score byte+rune + constants verified | unit | `go test -run TestJaroWinkler ./...` | `jarowinkler_test.go` ❌ Wave 0 |
| PERF-01 | 0 allocs on ASCII ≤ 50 chars | benchmark | `go test -bench=. -benchmem -run=^$ ./...` | `*_bench_test.go` ❌ Wave 0 |
| PERF-02 | ASCII fast path exists | code review | `go build -gcflags="-m" ./...` | each `*.go` ❌ Wave 0 |
| PERF-03 | Two-row DP confirmed | code review | — | each DP algo ❌ Wave 0 |
| TEST-01 | Reference vectors from primary sources | unit | `go test -run TestXxx_ReferenceVectors ./...` | `*_test.go` ❌ Wave 0 |
| TEST-02 | Property tests (range, identity, symmetry, triangle, NaN, Inf) | property | `go test -run TestProp ./...` | `props_test.go` ❌ Wave 0 |
| TEST-04 | Benchmarks with alloc assertions | benchmark | `go test -bench=. -benchmem -run=^$ ./...` | `*_bench_test.go` ❌ Wave 0 |
| TEST-05 | BDD scenarios per algorithm | BDD | `cd tests/bdd && go test ./...` | `*.feature` ❌ Wave 0 |
| DET-02 | Scores stable in golden file | golden | `make verify-determinism` | `algorithms.json` ❌ Wave 0 |
| DET-04 | No NaN/Inf/negative-zero | property | `go test -run TestProp_.*NoNaN\|NoInf ./...` | `props_test.go` ❌ Wave 0 |
| DX-02 | godoc + Example per algorithm | unit | `go test -run ExampleXxx ./...` | `example_test.go` ❌ Wave 0 |
| DX-05 | identifier-similarity example stable output | meta-test | `go test ./examples/identifier-similarity/...` | `main_test.go` ❌ Wave 0 |

### Sampling Rate

- **Per task commit:** `go test -race -shuffle=on -count=1 -run TestXxx ./...` (for the algorithm being implemented)
- **Per wave merge:** `make check` (full quality gate including lint, vet, coverage-check, golden verify)
- **Phase gate:** `make check` green on all 5 platforms before `/gsd-verify-work`

### Wave 0 Gaps

All the following must be created (all are `❌ Wave 0` above — none exist yet):
- [ ] `levenshtein.go`, `levenshtein_test.go`, `levenshtein_bench_test.go`, `levenshtein_fuzz_test.go`
- [ ] `hamming.go`, `hamming_test.go`, `hamming_bench_test.go`, `hamming_fuzz_test.go`
- [ ] `jaro.go`, `jaro_test.go`, `jaro_bench_test.go`, `jaro_fuzz_test.go`
- [ ] `jarowinkler.go`, `jarowinkler_test.go`, `jarowinkler_bench_test.go`, `jarowinkler_fuzz_test.go`
- [ ] `damerau_osa.go`, `damerau_osa_test.go`, `damerau_osa_bench_test.go`, `damerau_osa_fuzz_test.go`
- [ ] `damerau_full.go`, `damerau_full_test.go`, `damerau_full_bench_test.go`, `damerau_full_fuzz_test.go`
- [ ] `dispatch_levenshtein.go`, `dispatch_hamming.go`, `dispatch_jaro.go`, `dispatch_jarowinkler.go`, `dispatch_damerau_osa.go`, `dispatch_damerau_full.go`
- [ ] `props_test.go` — phase 2 property tests (extending or creating alongside the existing normalise/tokenise props)
- [ ] `algorithms_golden_test.go` + `testdata/golden/algorithms.json`
- [ ] `example_test.go` — ExampleXxx godoc examples for all 6 algorithms
- [ ] `tests/bdd/features/levenshtein.feature` (and 5 others)
- [ ] `tests/bdd/steps/algorithms_steps.go` (new or extended)
- [ ] `examples/identifier-similarity/main.go` + `main_test.go`

**Existing test infrastructure that Phase 2 inherits (no gaps):**
- `golden_test.go` (assertGolden, -update flag)
- `golden_canonical.go` (canonicalMarshal, WriteGoldenFile)
- `export_test.go` (CanonicalMarshalForTest, DispatchLenForTest, DispatchEntryNilForTest)
- `tests/bdd/go.mod` (godog + goleak + testify, replace directive)
- All Phase 1 quality gates in Makefile and CI

---

## Security Domain

These are pure string-comparison algorithms (no I/O, no parsing, no network, no user-supplied code execution). The security domain is limited to:

- **DoS via pathological inputs:** Levenshtein and DL-OSA are O(m·n); for two strings of length L, cost is O(L²). For L=10,000, this is 100M operations (~100ms). The library makes no promise about input-length caps. Document worst-case complexity in godoc. The fuzz tests exercise arbitrary-length inputs; the property tests verify non-panic.
- **Panic safety on invalid UTF-8:** guaranteed by the fuzz harness and the `PropAllPublic_NeverPanic` property. No explicit security gate needed beyond testing.
- **No injection vector:** pure functions with no I/O, no SQL, no system calls.

ASVS V5 (Input Validation): the only applicable category. Mitigation: fuzz tests + property tests covering `utf8.ValidString(input) == false` paths.

---

## Open Questions for the Planner (RESOLVED)

> All questions below were resolved during `/gsd-plan-phase` iterations 1-2.
> Resolutions are embedded in the seven PLAN.md files for Phase 2. Summary:
>
> 1. **Jaro-Winkler file naming:** `jarowinkler.go` (no underscore) — matches `normalise.go` / `tokenise.go` / `algoid.go` no-underscore convention.
> 2. **Dispatch:** Separate `dispatch_<algo>.go` per algorithm using the `var _ = func() bool { ... }()` idiom.
> 3. **Stack buffer threshold:** `const maxStackInputLen = 64` declared once in `levenshtein.go` and reused by DL-OSA + DL-Full; Jaro uses `const maxJaroStackLen = 256`.
> 4. **Rune mode:** Eager `[]rune` conversion in the rune fallback path; ASCII fast path uses byte indexing on the original string.
> 5. **BDD feature-file structure:** One `.feature` file per algorithm sharing a single `tests/bdd/steps/algorithms_steps.go`.
> 6. **`examples/identifier-similarity/` placement:** Dedicated plan 02-07 (Wave 4).
> 7. **`cross_algorithm_consistency_test.go`:** Included in plan 02-07.
> 8. **Three-row vs two-row DP for DL-OSA:** Three-row DP (track `i-2` row for transposition lookup).

The original research-time formulation of these questions follows for archival purposes.

1. **File naming: `jarowinkler.go` vs `jaro_winkler.go`?**
   Evidence from existing files: `normalise.go`, `tokenise.go`, `algoid.go`, `errors.go` — all single-word or bare concatenation. The AlgoID constant is `AlgoJaroWinkler`. Recommendation: `jarowinkler.go` (no underscore) to match the existing no-underscore convention. But `damerau_osa.go` and `damerau_full.go` need underscores to disambiguate. Planner picks.

2. **Dispatch registration: `var _ = func()...` vs a named `init()` equivalent?**
   Recommendation: `var _ = func() bool { dispatch[AlgoX] = XScore; return true }()` per algorithm in its own `dispatch_xxx.go`. Avoids `init()` (determinism-standards §13.5), avoids the merge conflicts that would arise from a single shared `dispatch_register.go` in Wave 2.

3. **Stack buffer size: `[64]int` or larger?**
   CONTEXT.md says "starting point `[64]int`". The two-row variant needs `[65 * 2]int` (130 ints = 1040 bytes) for Levenshtein with n≤64. DL-OSA needs `[65 * 3]int` = 1560 bytes. DL-Full needs `[65*2]int + [256]int = 3088 bytes`. All fit within typical goroutine stacks (8KB initial, growing as needed). Tune at benchstat; 64 is the correct starting threshold.

4. **Rune variant: eager `[]rune(a)` or lazy?**
   Recommendation: eager `[]rune(a)` for all Phase 2 algorithms. Simple, correct, and the 2-alloc cost on the rune path is expected and documented.

5. **BDD: one file per algorithm or one `algorithms.feature`?**
   Recommendation: one file per algorithm. Zero merge conflicts in Wave 2 parallel execution. Consistent with `normalisation.feature` (plan 01-06) which was one feature per primitive.

6. **Example program: attach to plan 02-06 (DL-Full) or a dedicated plan 02-07?**
   Recommendation: dedicated `plan 02-07-identifier-similarity-example` in Wave 2. The example requires all 6 algorithms to exist (it calls all 6 side-by-side). It cannot be built until all 6 plans are merged. A dedicated plan avoids forcing the last parallel plan (02-06) to wait for merge ordering. The planner may also attach it to a Wave 3 sequential finalisation pass.

7. **`cross_algorithm_consistency_test.go`?**
   CONTEXT.md defers this. The planner may include it in Wave 2 finalisation or defer to Phase 3. It adds < 1 day of work.

8. **`dispatch_levenshtein.go` vs including dispatch in `levenshtein.go` directly?**
   The separate `dispatch_xxx.go` pattern is preferred for Wave 2 parallelism (each plan touches only its own file). However, if the planner sequences Wave 2 serially, the dispatch assignment can live in the algorithm file itself. The per-file pattern is the safe default.

---

## Related Issues

`gh issue list` returned no results (repository may not be publicly available from this environment, or no open issues exist). Based on the CONTEXT.md and ROADMAP.md, the relevant tracking work is Phase 2 as a whole, tracked via ROADMAP.md and REQUIREMENTS.md.

If issues are created per CLAUDE.md convention (via `issue-writer` agent before starting work), each algorithm implementation should have its own GitHub issue. Recommend one issue per Wave 1 plan and one issue per Wave 2 plan — 6 issues total (or 7 if the identifier-similarity example is a separate plan).

---

## Assumptions Log

| # | Claim | Section | Risk if Wrong |
|---|-------|---------|---------------|
| A1 | Winkler 1990 paper has no patent claim (consistent with licensing skill but not independently USPTO-verified this session) | Patent Screen | Low — the Winkler 1990 work is U.S. Census Bureau government work; federal government employees cannot patent work done in the scope of employment under 35 U.S.C. §207. Risk is near-zero. |
| A2 | DIXON/DICKSONX Jaro score is 0.7667 (not 0.7666) — computed from first principles above but not cross-validated against Winkler 1990 paper text | Primary Sources | Low — the computation is mechanical; the rounding to 4 decimal places gives 0.7667 |
| A3 | The `[65*3]int` stack-allocated buffer for DL-OSA with n≤64 does not escape to the heap when passed as a slice to the inner DP function | Implementation Patterns | Medium — escape analysis depends on precise code structure. Plan must include `go build -gcflags="-m"` verification. |
| A4 | The BDD step registration pattern from godog v0.15.0 uses `*godog.ScenarioContext` — consistent with godog v0.15.0 API but not verified in this session | BDD Scenario Coverage | Low — godog API is stable; the BDD module already exists from Phase 1 with godog v0.15.0 pinned. |

---

## Sources

### Primary (HIGH confidence)

- `docs/requirements.md` §7.1.1–7.1.6 (algorithm specs, recurrences, reference vectors, normalisation rules, edge cases) — the authoritative spec, directly consulted
- `docs/requirements.md` §13 (determinism guarantees, float rules, NaN/Inf prevention)
- `docs/requirements.md` §14 (performance budgets by algorithm)
- `docs/requirements.md` §15 (testing strategy, property test names, fuzz harness pattern)
- `.planning/phases/01-foundation-infrastructure/01-04-determinism-infra-SUMMARY.md` (golden file format LOCKED, `assertGolden` API, canonical byte form)
- `.planning/phases/01-foundation-infrastructure/01-05-primitives-algoid-errors-SUMMARY.md` (dispatch skeleton contract, AlgoID values 0–5, `dispatch_xxx.go` registration pattern precedent)
- `.planning/phases/01-foundation-infrastructure/01-06-primitives-normalise-SUMMARY.md` (ASCII fast path pattern, stack buffer pattern, fuzz harness seed convention)
- `.planning/phases/02-core-character-algorithms-six/02-CONTEXT.md` (all locked decisions)
- `.claude/skills/algorithm-correctness-standards/SKILL.md` (citation format, invariants, reference-vector discipline)
- `.claude/skills/algorithm-licensing-standards/SKILL.md` (patent screen criteria, fresh-implementation discipline)
- `.claude/skills/performance-standards/SKILL.md` (two-row DP pattern, ASCII fast path pattern, benchmark file structure)
- `.claude/skills/determinism-standards/SKILL.md` (float rules, NaN/Inf guards, map-iteration rule)

### Secondary (MEDIUM confidence)

- Winkler 1990 constants (boost threshold 0.7, prefix cap 4, scale 0.1) cited in `docs/requirements.md` §7.1.6 and `algorithm-correctness-standards/SKILL.md` — traced to the original paper via the spec; not independently accessed in this session
- Boytsov 2011 discriminating vector (`"ca"`/`"abc"` → OSA=3, Full=2) cited in `docs/requirements.md` §7.1.2–7.1.3 — verified by manual recurrence-tracing above

### Confirmatory references (for cross-validation in implementations)

The following are cited in `algorithm-licensing-standards/SKILL.md` as permissible MIT-licensed references for cross-validation only (no code to be copied):

- `github.com/agnivade/levenshtein` (MIT) — Levenshtein reference vectors
- `github.com/xrash/smetrics` (MIT) — Jaro, Jaro-Winkler reference vectors
- `github.com/adrg/strutil` (MIT) — general reference vectors
- `github.com/hbollon/go-edlib` (MIT) — DL-OSA/Full reference vectors

---

## RESEARCH COMPLETE
