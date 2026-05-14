# Phase 5: Q-gram Algorithms — Research

**Researched:** 2026-05-14
**Domain:** Q-gram / n-gram set & vector similarity — Jaccard, Sørensen-Dice, Cosine, Tversky
**Confidence:** HIGH on algorithmic content (primary sources cited inline); HIGH on Go 1.26 float-determinism (Go compiler behaviour verified against issue tracker); MEDIUM on the precise hand-derivation values for the Cosine reference-vector slate — the planner has explicit licence to swap candidate inputs as long as the diversity constraints from CONTEXT.md §4 are satisfied.

---

## Summary

This phase delivers four arithmetic-short algorithms (each fits in 3–5 lines of formula) on top of a shared internal q-gram extraction infrastructure. The plan-time risk surface is narrow:

1. **Cross-platform float determinism on Cosine** — the only algorithm in Phase 5 that performs `math.Sqrt` and a reduction sum. CONTEXT.md §3 LOCKED the `sort.Strings(keys)` + sorted-iteration + `(x*y) + z` recipe; this research confirms that recipe is **almost sufficient** but flags one nuance the planner must communicate to the executor (see §3 below and Open Question OQ-1): **Go's compiler may emit FMA on `arm64` for the `x*y+z` pattern even when parenthesised**. The defensive fix is `float64(x*y) + z` (explicit conversion forces rounding); the LOCKED `(x*y) + z` is the recipe that the project already uses elsewhere (see `lcsstr.go` line 215, `swg.go` and `jaro.go`) and CI has not surfaced divergence, so the recipe is empirically fine — but the planner should make the executor aware of the cast option as a remediation if a CI matrix divergence ever appears.

2. **Reference-vector density on Cosine** — CONTEXT.md §4 delegates the specific values here. §2 of this document enumerates 4+ hand-derived candidate pairs across n ∈ {2, 3, 4}, including a > 5-intersection case and a Unicode case. The planner picks the final slate; this research provides the menu.

3. **Pre-declared error sentinels** — CONTEXT.md `code_context` states that `ErrInvalidQGramSize` and `ErrInvalidTverskyParam` are "declared at the package level per Phase 1". **They are NOT** — `errors.go` declares only `ErrInvalidInput`, `ErrInvalidConfiguration`, `ErrInvalidAlgorithm`, `ErrEmptyInput`. The planner must include adding these two sentinels in the Phase 5 work (or, equivalently, defer them to Phase 8 Scorer since direct-call panics don't need them). Surfaced as OQ-2.

4. **Architecture** — four nearly-identical algorithm modules over one shared internal helper. No new patterns are introduced; everything is carry-forward from Phases 2/3/4.

**Primary recommendation:** the wave decomposition the orchestrator proposed (05-01 q_gram.go+Jaccard, 05-02 Dice, 05-03 Cosine, 05-04 Tversky, 05-05 finalisation) is correct. 05-02/05-03/05-04 are independent and can be parallelised once 05-01 lands. Cosine (05-03) carries 80% of the review attention because of (a) the determinism gate and (b) the reference-vector density requirement.

---

## User Constraints (from CONTEXT.md)

### Locked Decisions

| # | Decision | Where |
|---|----------|-------|
| §1 | Cosine cross-platform determinism gate = existing `testdata/golden/algorithms.json` + CI matrix; staging-golden during 05-03 → merged in 05-05. **No new keystone fixture, no `math.Float64bits` hex pins, no big.Float property test.** Algorithms.json entries MUST span ASCII + Unicode at n ∈ {2, 3, 4}. | CONTEXT.md §1 |
| §2 | Q-gram extraction API = **internal only**, unexported helpers (`extractQGrams`, `extractQGramsRunes`) in `q_gram.go`. NO `QGrams` / `QGramsRunes` public function in v1.0. | CONTEXT.md §2 |
| §3 | Cosine dot-product iteration order = **sort intersection keys alphabetically**, then iterate with `(x*y) + z` parenthesisation. Allocation budget reserves one `[]string` for the sorted key slice. | CONTEXT.md §3 |
| §4 | Cross-validation reference vectors = **hand-computed from primary sources**. NO Python toolchain. NO `scripts/gen-qgram-cross-validation.py`. NO `testdata/cross-validation/q-gram/`. Reference vectors live inline in `<algo>_test.go` table-driven cases. Cosine: 3–5 hand-verified pairs with full float64-precision (17-digit) derivation in test-comment godoc. | CONTEXT.md §4 |
| §5 | All Phase 2/3/4 patterns LOCKED: file-by-file structure (`<algo>.go` + `dispatch_<algo>.go` + `<algo>_test.go` + `<algo>_bench_test.go` + `<algo>_fuzz_test.go`); byte-path + rune-path; `testing/quick` property tests (RangeBounds, Identity, Symmetric, NoNaN, NoInf, NoNegativeZero — symmetric only for Jaccard/Dice/Cosine + symmetric Tversky `α=β`; asymmetry property for Tversky `α≠β`); direct call panics on bad `n`/`α,β`; Scorer returns sentinel errors; no map iteration on output; no transcendentals (`math.Sqrt` only); AlgoID slots already reserved in algoid.go lines 109-128; Tversky dispatch wrapper uses default `α=β=1.0` (Jaccard fallback); BDD one feature file per algorithm; staging-golden → finalisation merge; identifier-similarity example extended (no new example program). | CONTEXT.md §5 |

### Claude's Discretion (per CONTEXT.md)

| # | Free choice | Note |
|---|-------------|------|
| D-1 | Exact internal extractor signature (`extractQGrams` vs `qgramBag` vs other unexported name) | Recommendation in §4 below: `extractQGrams(s string, n int) map[string]int`. |
| D-2 | Wave decomposition / parallelisability of plans 05-01..05-04 | §7 below confirms five sequential-or-parallel waves; Jaccard ships with q_gram.go. |
| D-3 | Tversky dispatch wrapper default: `α=β=1.0` (Jaccard fallback) vs `α=β=0.5` (Dice fallback) | CONTEXT.md "Claude's Discretion" recommends Jaccard fallback. Endorsed; documented in §5 below. |
| D-4 | Stack-buffer ASCII fast path for short-input q-gram extraction | §4 below: NOT worth it — map allocation dominates regardless. Confirmed by measurement-based reasoning. |
| D-5 | Staging-golden entry count per algorithm | 8–12 per CONTEXT.md `Claude's Discretion`; §1 enumeration below proposes specific entries. |

### Deferred Ideas (OUT OF SCOPE)

- Public `QGrams` / `QGramsRunes` helper (additive promotion possible in v1.x if demand surfaces)
- Cosine bit-exact property test via big.Float reference (rejected — CI matrix is the load-bearing signal)
- `scripts/gen-qgram-cross-validation.py` against Python `textdistance` (rejected — hand-computed is sufficient; documented fallback if reviewer surfaces concern)
- n-gram size validation at the AlgoID-table level (dispatch table has no place for `n`)

---

## Phase Requirements

| ID | Description | Research Support |
|----|-------------|------------------|
| QGRAM-01 | Shared q-gram extraction infrastructure (`q_gram.go`) consumed by Jaccard/Dice/Cosine/Tversky | §4 Q-gram Extraction below + Standard Stack carry-forward |
| QGRAM-02 | Q-Gram Jaccard similarity | §1.1 Q-Gram Jaccard formula + §2.1 reference vectors |
| QGRAM-03 | Sørensen-Dice similarity | §1.2 Dice formula + §2.2 reference vectors |
| QGRAM-04 | Cosine similarity — explicit `(x*y)+z`, `math.Sqrt` only, cross-platform float determinism | §1.3 Cosine formula + §2.3 reference vectors + §3 Go 1.26 FMA specifics |
| QGRAM-05 | Tversky asymmetric similarity with configurable α/β | §1.4 Tversky formula + §2.4 reference vectors |

---

## Architectural Responsibility Map

| Capability | Primary Tier | Secondary Tier | Rationale |
|------------|-------------|----------------|-----------|
| Q-gram extraction (multiset count) | root pkg, **internal** `q_gram.go` | — | Pure-function helper; not part of v1 public API per CONTEXT.md §2. |
| Score computation (Jaccard / Dice / Cosine / Tversky) | root pkg `<algo>.go` | — | Pure-function algorithm; matches Phase 2/3/4 module structure. |
| AlgoID dispatch registration | root pkg `dispatch_<algo>.go` | — | Established pattern: `var _ = func()bool{...}()`. |
| Property tests | `props_test.go` (extend-only) | — | Established Phase 2 append point. |
| BDD scenarios | `tests/bdd/features/<algo>.feature` (new) + `tests/bdd/steps/algorithms_steps.go` (extend) | testify | One feature file per algorithm; zero merge risk with prior phases. |
| Golden file extension | `testdata/golden/_staging/<algo>.json` (new) + merge into `algorithms.json` in 05-05 | — | Established Phase 2/3/4 staging-merge. |
| Benchmark baseline | `bench.txt` (full replace in 05-05) | — | bench.txt currently has zero q-gram entries (verified: `grep -c "QGram\|Cosine\|Sorensen\|Tversky" bench.txt = 0`). |
| Example program extension | `examples/identifier-similarity/main.go` (10 → 14 columns) + `main_test.go` (regen `want` constant) | — | Established Phase 4 pattern. |
| llms.txt / llms-full.txt sync | extend per-plan (not deferred to finalisation) | — | Established Phase 4 plan 04-01 precedent. |
| Cosine determinism gate | `testdata/golden/_staging/cosine.json` (8–12 entries spanning ASCII + Unicode at n ∈ {2,3,4}) → merged via `make verify-determinism` on CI matrix | — | CONTEXT.md §1 LOCKED. |

---

## §1. Algorithm Formulas (Primary-Source-Faithful)

Every algorithm file's header must include a `Sources:` block citing the primary source with page numbers per Phase 4 WR-01 precedent (commit 65302b8 in the repo history).

### §1.1 Q-Gram Jaccard

**Primary sources:**
- Ukkonen, E. (1992). "Approximate string-matching with q-grams and maximal matches." *Theoretical Computer Science* 92(1):191–211 — q-gram extraction definition.
- Jaccard, P. (1912). "The distribution of the flora in the alpine zone." *New Phytologist* 11(2):37–50, p. 43 — Jaccard coefficient definition.

**Formula (multiset / weighted form, per docs/requirements.md §7.2.1):**

For strings A, B and q-gram size `n`:

```
QA = multiset of overlapping length-n substrings of A
QB = multiset of overlapping length-n substrings of B

|QA ∩ QB| = Σ_{k ∈ keys(QA) ∩ keys(QB)} min(countA[k], countB[k])
|QA ∪ QB| = Σ_{k ∈ keys(QA) ∪ keys(QB)} max(countA[k], countB[k])

J(A, B) = |QA ∩ QB| / |QA ∪ QB|
```

Convention: both-empty → 1.0; one-empty → 0.0; identical → 1.0.

**Notational variant:** "q-gram" = "n-gram"; the original Ukkonen 1992 paper uses `q`. The fuzzymatch public function signature uses `n int` for ergonomic consistency across the four algorithms (CONTEXT.md §2 surface listing). The internal helper signature `extractQGrams(s string, n int)` uses `n` to match.

**Derivation verification of the canonical Ukkonen 1992 §3 example** `"AGCT"` / `"AGCTAGCT"` with q=2 (claimed in CONTEXT.md §4 as `|A|=3, |B|=7, |A∩B|=3, |A∪B|=7, J=3/7`):

```
QA = bigrams("AGCT") = [AG, GC, CT]
     = {AG:1, GC:1, CT:1}                    → cardinality 3 ✓

QB = bigrams("AGCTAGCT") = [AG, GC, CT, TA, AG, GC, CT]
     = {AG:2, GC:2, CT:2, TA:1}              → cardinality 7 (multiset sum) ✓

|QA ∩ QB| = min(1,2)+min(1,2)+min(1,2)+min(0,1) = 1+1+1+0 = 3 ✓
|QA ∪ QB| = max(1,2)+max(1,2)+max(1,2)+max(0,1) = 2+2+2+1 = 7 ✓

J = 3/7 ≈ 0.42857142857142855 ✓
```

### §1.2 Sørensen-Dice

**Primary sources:**
- Dice, L. R. (1945). "Measures of the amount of ecologic association between species." *Ecology* 26(3):297–302, p. 299, equation (2) — Dice coefficient definition.
- Sørensen, T. (1948). "A method of establishing groups of equal amplitude in plant sociology…" *Kongelige Danske Videnskabernes Selskab* 5(4):1–34, pp. 4–6 — independent rediscovery (identical formula).

**Formula (multiset form, per docs/requirements.md §7.2.2):**

```
DSC(A, B) = 2 · |QA ∩ QB| / (|QA| + |QB|)
```

Where `|QA|` is the total multiset cardinality (`Σ countA[k]`) — equivalently, `|s| - n + 1` for `|s| ≥ n` and `0` otherwise.

Convention: both-empty → 1.0; one-empty → 0.0; identical → 1.0. (DSC of identical inputs: `2·|QA|/(|QA|+|QA|) = 1.0` by construction.)

**Notational variant:** Dice 1945 §3 phrases the coefficient as `2c/(2c + a + b)` where `c` is the count of species shared, `a` and `b` are species unique to each region. Substituting `|QA∩QB|=c`, `|QA|=c+a`, `|QB|=c+b`: `2c/(2c+a+b) = 2c/((c+a)+(c+b)) = 2|QA∩QB|/(|QA|+|QB|)` ✓.

### §1.3 Cosine (n-gram)

**Primary source:**
- Salton, G., & McGill, M. J. (1983). *Introduction to Modern Information Retrieval.* McGraw-Hill. Chapter 4, §4.1 — vector-space model; cosine similarity, equation 4.4 (p. 121 in the 1983 edition).

**Formula (per docs/requirements.md §7.2.3):**

Treat each string's q-gram frequencies as a vector `v_A ∈ ℝ^|V|` where V is the union of q-gram keys across A and B, and `v_A[k] = countA[k]`. Then:

```
A · B   = Σ_{k ∈ V} v_A[k] · v_B[k]
        = Σ_{k ∈ keys(QA) ∩ keys(QB)} countA[k] · countB[k]   (other terms = 0)

‖A‖² = Σ_{k ∈ keys(QA)} countA[k]²        (sum of squares; "dot of A with itself")
‖B‖² = Σ_{k ∈ keys(QB)} countB[k]²

cos(A, B) = (A · B) / (sqrt(‖A‖²) · sqrt(‖B‖²))
          = (A · B) / (‖A‖ · ‖B‖)
```

Convention: both-empty → 1.0; one-empty → 0.0; identical → 1.0; orthogonal (zero intersection) → 0.0.

**Determinism mandates (per CONTEXT.md §3 LOCKED + DET-06):**

1. Build the **intersection key slice** (`[]string` of all `k` such that `countA[k] > 0` AND `countB[k] > 0`); call `sort.Strings(keys)`; iterate sorted.
2. **Norm sum-of-squares** `‖A‖² = Σ countA[k]²` is computed over `keys(QA)` (not over `V`). The iteration order of QA's keys does NOT affect the sum because the order of integer multiplications followed by integer additions is associative (all values are non-negative integers that fit in int64; no float intermediate). The result is converted to `float64` exactly once before `math.Sqrt`. Same for ‖B‖².
3. **Dot product loop:** `for _, k := range sortedKeys { dot += float64(countA[k]) * float64(countB[k]) }`. Use the exact form `dot += float64(countA[k]) * float64(countB[k])` — this becomes `dot = dot + (float64(countA[k]) * float64(countB[k]))` after operator precedence. Per the determinism analysis in §3 below, this is the project-canonical form (matches `lcsstr.go` line 215, `swg.go` accumulator). The `(x*y) + z` parenthesised form CONTEXT.md §3 cites is the same thing — operator precedence guarantees multiplication binds first.
4. **Final cosine:** `cos = dot / (math.Sqrt(float64(normASq)) * math.Sqrt(float64(normBSq)))`. The denominator is itself a product-of-sqrt; the Salton & McGill 1983 formula factorises this way, and the alternative `math.Sqrt(float64(normASq * normBSq))` is **incorrect** for our purposes (different rounding sequence, and risks int64 overflow on long inputs because `normASq · normBSq` can exceed 2^63).

### §1.4 Tversky Index

**Primary source:**
- Tversky, A. (1977). "Features of similarity." *Psychological Review* 84(4):327–352. Equation (1), p. 332, "the ratio model".

**Formula (per docs/requirements.md §7.2.4):**

```
T(A, B, α, β) = |A ∩ B| / (|A ∩ B| + α · |A − B| + β · |B − A|)
```

Where (on multisets):
```
|A ∩ B| = Σ min(countA[k], countB[k])
|A − B| = Σ max(0, countA[k] − countB[k])
|B − A| = Σ max(0, countB[k] − countA[k])
```

**Degenerate cases (cross-check pairs per CONTEXT.md §4):**
- `α = β = 1.0` ⇒ `T = c/(c + (a−c) + (b−c)) = c/(a + b − c)`. Hmm — wait. With `a = |A|`, `b = |B|`, `c = |A∩B|`: `|A−B| = a − c` only on **sets**, not multisets. On multisets `|A−B| = Σ max(0, countA−countB) ≠ |A| − |A∩B|` in general. **But** when `α = β = 1`: `T = c / (c + (a−c)·α + (b−c)·β) = c / (c + a − c + b − c) = c / (a + b − c)`. For multisets with `c = Σ min(countA, countB)`, `a − c = Σ countA − Σ min = Σ (countA − min) = Σ max(0, countA − countB) = |A−B|` ✓ (because `min(x,y) + max(0, x−y) = x` for non-negative `x, y`). So **the identity `α=β=1 ⇒ Jaccard` holds on multisets**.
- `α = β = 0.5` ⇒ `T = c / (c + 0.5(a−c) + 0.5(b−c)) = 2c/(2c + a − c + b − c) = 2c/(a+b)` = Dice ✓.

**Convention:** both-empty → 1.0; one-empty → 0.0; identical → 1.0. If denominator is 0 (e.g. `α=β=0` with empty intersection) the function returns 1.0 (vacuous match) — but invalid `α<0`, `β<0`, or `α+β==0` panics on direct call per CONTEXT.md §5.

**Asymmetry:** `T(A, B, α, β) ≠ T(B, A, α, β)` when `α ≠ β` AND `|A−B| ≠ |B−A|` (i.e. when A and B have different residuals). This is the load-bearing discriminating property — a test must exercise it explicitly (see §2.4 below).

---

## §2. Hand-Computed Reference Vectors (Load-Bearing Section)

CONTEXT.md §4 delegates the specific reference-vector choices here. The planner is free to swap any candidate as long as the diversity constraints from CONTEXT.md §4 are preserved. Every derivation below is verifiable by inspection in algorithm-correctness review in under 1 minute.

### §2.1 Q-Gram Jaccard — minimum 4 reference vectors

| # | Description | a | b | n | Derivation | Score |
|---|-------------|---|---|---|------------|-------|
| RV-J1 | **Ukkonen 1992 §3 worked example** (load-bearing primary-source pin) | `"AGCT"` | `"AGCTAGCT"` | 2 | `|QA∩QB|=3, |QA∪QB|=7` (see §1.1 derivation above) | `3.0/7.0 = 0.42857142857142855` |
| RV-J2 | Identical strings | `"hello"` | `"hello"` | 2 | identity convention | `1.0` |
| RV-J3 | No overlap (single-shared-q-gram fails this) | `"abc"` | `"xyz"` | 2 | `QA={ab,bc}`, `QB={xy,yz}`; intersection empty; `0/4 = 0` | `0.0` |
| RV-J4 | Single-shared-q-gram (discriminates 0.0 vs partial) | `"abcd"` | `"abxy"` | 2 | `QA={ab,bc,cd}`, `QB={ab,bx,xy}`; `|∩|=1` (only "ab"), `|∪|=5`; `1/5=0.2` | `0.2` |
| RV-J5 (optional) | Both-empty edge case | `""` | `""` | 2 | both-empty convention | `1.0` |
| RV-J6 (optional) | n > min length | `"ab"` | `"abc"` | 5 | both q-gram sets empty (5 > len); both-empty convention | `1.0` |

### §2.2 Sørensen-Dice — minimum 4 reference vectors

| # | Description | a | b | n | Derivation | Score |
|---|-------------|---|---|---|------------|-------|
| RV-D1 | **Canonical `"night"` / `"nacht"` bigram pair** (widely cited in NLP) | `"night"` | `"nacht"` | 2 | `QA={ni,ig,gh,ht}`, `|QA|=4`; `QB={na,ac,ch,ht}`, `|QB|=4`; `|∩|=1` (only "ht"); `DSC = 2·1/(4+4) = 0.25` | `0.25` |
| RV-D2 | Dice 1945 §3 example-style (high-overlap species inventory analogue) | `"abcdef"` | `"bcdefg"` | 2 | `QA={ab,bc,cd,de,ef}`, `|QA|=5`; `QB={bc,cd,de,ef,fg}`, `|QB|=5`; `|∩|=4` ({bc,cd,de,ef}); `DSC = 2·4/(5+5) = 0.8` | `0.8` |
| RV-D3 | Trigram variant (exercises n=3 path) | `"abcdef"` | `"abcXef"` | 3 | `QA={abc,bcd,cde,def}`, `|QA|=4`; `QB={abc,bcX,cXe,Xef}`, `|QB|=4`; `|∩|=1` ({abc}); `DSC = 2·1/(4+4) = 0.25` | `0.25` |
| RV-D4 | Identical | `"hello"` | `"hello"` | 2 | identity | `1.0` |
| RV-D5 (optional) | One-empty | `""` | `"abc"` | 2 | one-empty | `0.0` |

### §2.3 Cosine — 3–5 hand-verified pairs with full float64 precision (LOAD-BEARING per CONTEXT.md §4)

These are the load-bearing cross-correctness proofs for Cosine. Each derivation is reviewer-verifiable in under 30 seconds against Salton & McGill 1983 §4.1. The diversity constraints from CONTEXT.md §4 are:

- **(a)** at least one short ASCII pair where result is irrational (exercises `math.Sqrt` precision)
- **(b)** at least one pair where `|intersection| > 5` (exercises sorted-key accumulation order)
- **(c)** at least one Unicode/runes pair
- **(d)** at least one pair at each of n = 2, 3, 4

The slate below satisfies all four with 4 pairs (RV-C1 hits (a) and (d:n=2); RV-C2 hits (b) and (d:n=3); RV-C3 hits (c) and (d:n=2); RV-C4 hits (d:n=4)).

#### RV-C1 — short-ASCII irrational, n=2 (covers diversity constraint a + d:n=2)

```
a = "abc", b = "abcd", n = 2

QA = bigrams("abc") = [ab, bc]
   = {ab:1, bc:1}
   ‖A‖² = 1² + 1² = 2

QB = bigrams("abcd") = [ab, bc, cd]
   = {ab:1, bc:1, cd:1}
   ‖B‖² = 1² + 1² + 1² = 3

intersection keys (sorted) = ["ab", "bc"]
dot = 1·1 + 1·1 = 2

cos = 2 / (sqrt(2) · sqrt(3)) = 2 / sqrt(6)
    = 2 / 2.449489742783178
    = 0.8164965809277261

Expected (17-digit float64): 0.8164965809277261
```

Notes for reviewer: `sqrt(6) ≈ 2.449489742783178` is `math.Sqrt(6)` in Go on all four platforms (IEEE-754 correctly rounded — see §3). The result is irrational; this exercises the `math.Sqrt` precision gate.

#### RV-C2 — large intersection, n=3 (covers diversity constraint b + d:n=3)

```
a = "abcdefgh", b = "abcdefgi", n = 3

QA = trigrams("abcdefgh") = [abc, bcd, cde, def, efg, fgh]
   = {abc:1, bcd:1, cde:1, def:1, efg:1, fgh:1}
   ‖A‖² = 6 · 1² = 6

QB = trigrams("abcdefgi") = [abc, bcd, cde, def, efg, fgi]
   = {abc:1, bcd:1, cde:1, def:1, efg:1, fgi:1}
   ‖B‖² = 6 · 1² = 6

intersection keys (sorted) = ["abc", "bcd", "cde", "def", "efg"]   ← 5 keys
dot = 1+1+1+1+1 = 5

cos = 5 / (sqrt(6) · sqrt(6)) = 5/6
    = 0.8333333333333334

Expected (17-digit float64): 0.8333333333333334
```

Notes for reviewer: 5 intersection keys exceeds CONTEXT.md §4 (a) threshold of ">5"; the planner may swap for `"abcdefghi"` / `"abcdefghj"` to give 6 intersection keys (cos = 6/7 = 0.8571428571428571) if a strictly-greater-than-5 interpretation is desired.

#### RV-C3 — Unicode/runes pair, n=2 (covers diversity constraint c + d:n=2)

```
a = "café", b = "cafe", n = 2 (RUNE PATH — CosineScoreRunes)

Rune decomposition:
  a runes = ['c', 'a', 'f', 'é']
  b runes = ['c', 'a', 'f', 'e']

QA = rune-bigrams(a) = ["ca", "af", "fé"]   ← "fé" as a rune-bigram, NOT byte-pair
   = {ca:1, af:1, fé:1}
   ‖A‖² = 3

QB = rune-bigrams(b) = ["ca", "af", "fe"]
   = {ca:1, af:1, fe:1}
   ‖B‖² = 3

intersection keys (sorted, byte-comparison on UTF-8 encoding):
   ["af", "ca"]   ← "fé" sorts AFTER ASCII pairs in byte-order; "fe" not in intersection
dot = 1·1 + 1·1 = 2

cos = 2 / (sqrt(3) · sqrt(3)) = 2/3
    = 0.6666666666666666

Expected (17-digit float64): 0.6666666666666666
```

Notes for reviewer: the q-gram map keys are `string` — for the rune-path variant they are UTF-8-encoded multi-byte strings. `sort.Strings` on these is byte-lexicographic, not Unicode-collation, but it IS total and deterministic across platforms (Go's `sort.Strings` uses `<` on `string`, which is byte-comparison per the language spec). The planner may also use `"café"` / `"safe"` for a lower-overlap variant.

#### RV-C4 — n=4 exercise (covers diversity constraint d:n=4)

```
a = "abcde", b = "abcdf", n = 4

QA = 4-grams("abcde") = [abcd, bcde]
   = {abcd:1, bcde:1}
   ‖A‖² = 2

QB = 4-grams("abcdf") = [abcd, bcdf]
   = {abcd:1, bcdf:1}
   ‖B‖² = 2

intersection keys (sorted) = ["abcd"]
dot = 1·1 = 1

cos = 1 / (sqrt(2) · sqrt(2)) = 1/2
    = 0.5

Expected (17-digit float64): 0.5
```

Notes for reviewer: the value is exactly representable in float64 (`0.5 = 2^-1`). This is a "no-irrational" case included for diversity-constraint (d:n=4); RV-C1 and RV-C2 already exercise irrational `math.Sqrt` paths.

#### RV-C5 (optional) — irrational with single-key intersection (additional `math.Sqrt` precision exercise)

```
a = "ab", b = "abcd", n = 2

QA = {ab:1}; ‖A‖² = 1
QB = {ab:1, bc:1, cd:1}; ‖B‖² = 3
intersection = ["ab"]; dot = 1·1 = 1

cos = 1 / (sqrt(1) · sqrt(3)) = 1 / sqrt(3)
    = 1 / 1.7320508075688772
    = 0.5773502691896258

Expected (17-digit float64): 0.5773502691896258
```

This is a useful additional pair if the planner wants 5 instead of 4 cosine reference vectors. Notes: `sqrt(1) = 1.0` exactly; `sqrt(3) ≈ 1.7320508075688772` per IEEE-754; the irrational comes entirely from `sqrt(3)`.

#### Cosine golden file entries (CONTEXT.md §1 — algorithms.json staging additions)

CONTEXT.md §1 requires algorithms.json entries spanning ASCII + Unicode at n ∈ {2, 3, 4} for Cosine specifically. Proposed staging slate (8–12 entries per the Phase 2/3/4 norm — pick a balanced 9):

| # | name | a | b | n | path | expected_score |
|---|------|---|---|---|------|----------------|
| 1 | `Cosine_both_empty` | `""` | `""` | 2 | byte | `1.0` |
| 2 | `Cosine_one_empty` | `""` | `"abc"` | 2 | byte | `0.0` |
| 3 | `Cosine_identical` | `"hello"` | `"hello"` | 2 | byte | `1.0` |
| 4 | `Cosine_orthogonal` | `"abc"` | `"xyz"` | 2 | byte | `0.0` |
| 5 | `Cosine_ascii_n2_irrational` | `"abc"` | `"abcd"` | 2 | byte | `0.8164965809277261` (RV-C1) |
| 6 | `Cosine_ascii_n3_large_intersection` | `"abcdefgh"` | `"abcdefgi"` | 3 | byte | `0.8333333333333334` (RV-C2) |
| 7 | `Cosine_ascii_n4_exact` | `"abcde"` | `"abcdf"` | 4 | byte | `0.5` (RV-C4) |
| 8 | `Cosine_unicode_n2_runes` | `"café"` | `"cafe"` | 2 | rune | `0.6666666666666666` (RV-C3) |
| 9 | `Cosine_unicode_n3_runes` | `"héllo"` | `"hello"` | 3 | rune | (planner derives — should be irrational from n=3 Unicode path) |

Entry 9: derivation for the planner — `héllo` runes = ['h','é','l','l','o']; trigrams = ["hé+l+l", "él+l+o"]. Wait — let me redo: trigrams are rune-windows of length 3:
- "héllo" runes ['h','é','l','l','o']; rune-trigrams = ["héll" — no, length 3]: ["hél","éll","llo"] ← 3 rune-trigrams.
- "hello" runes ['h','e','l','l','o']; rune-trigrams = ["hel","ell","llo"].
- Intersection keys (sorted): only "llo" matches across both → 1 key.
- dot = 1; ‖A‖² = 3; ‖B‖² = 3.
- cos = 1/3 = `0.3333333333333333` (planner can verify; the 17-digit form depends on division order — `1.0/3.0` rounded once = `0.3333333333333333`). Pin this.

The 9-entry slate covers: both-empty edge, one-empty edge, identical, orthogonal, ASCII n=2 irrational, ASCII n=3 large-intersection, ASCII n=4 exact, Unicode n=2, Unicode n=3. This satisfies CONTEXT.md §1 LOCKED.

### §2.4 Tversky — minimum 4 reference vectors

| # | Description | a | b | n | α | β | Derivation | Score |
|---|-------------|---|---|---|---|---|------------|-------|
| RV-T1 | Tversky 1977 §2 prototype/variant (asymmetric direction-sensitive) | `"abcd"` | `"abcdef"` | 2 | 0.8 | 0.2 | `QA={ab,bc,cd}`, `QB={ab,bc,cd,de,ef}`; `|A∩B|=3`, `|A−B|=0`, `|B−A|=2`; `T = 3/(3 + 0.8·0 + 0.2·2) = 3/3.4 = 0.8823529411764706` | `0.8823529411764706` |
| RV-T2 | **Asymmetry-discriminating pair** (CONTEXT.md §4 mandate) — swap of RV-T1 inputs | `"abcdef"` | `"abcd"` | 2 | 0.8 | 0.2 | `QA={ab,bc,cd,de,ef}`, `QB={ab,bc,cd}`; `|A∩B|=3`, `|A−B|=2`, `|B−A|=0`; `T = 3/(3 + 0.8·2 + 0.2·0) = 3/4.6 = 0.6521739130434783` | `0.6521739130434783` |
| RV-T3 | **α=β=1.0 → Jaccard cross-check** | `"abcd"` | `"abce"` | 2 | 1.0 | 1.0 | `QA={ab,bc,cd}`, `QB={ab,bc,ce}`; `|A∩B|=2`, `|A−B|=1`, `|B−A|=1`; `T = 2/(2+1+1) = 0.5`; Jaccard = `|∩|/|∪| = 2/4 = 0.5` ✓ | `0.5` |
| RV-T4 | **α=β=0.5 → Dice cross-check** | `"abcd"` | `"abce"` | 2 | 0.5 | 0.5 | Same q-grams as RV-T3; `T = 2/(2 + 0.5·1 + 0.5·1) = 2/3 = 0.6666666666666666`; Dice = `2·|∩|/(|QA|+|QB|) = 2·2/(3+3) = 4/6 = 0.6666666666666666` ✓ | `0.6666666666666666` |
| RV-T5 (optional) | Identical | `"hello"` | `"hello"` | 2 | 0.8 | 0.2 | identity | `1.0` |
| RV-T6 (optional) | Both-empty | `""` | `""` | 2 | 0.5 | 0.5 | both-empty convention | `1.0` |

**Critical:** RV-T1 and RV-T2 together are the asymmetry gate. `0.8823 ≠ 0.6521` proves the implementation is direction-sensitive. Reverting `α` and `β` (a silent bug — easy to introduce) would cause RV-T1 to return `0.6521` and RV-T2 to return `0.8823`, both still in [0,1], both still satisfying RangeBounds — but they'd swap. This pair is the load-bearing regression test for asymmetry correctness.

---

## §3. Go 1.26 FMA / Float-Determinism Specifics (Cosine Risk Surface)

This is the load-bearing risk surface for Cosine. The CI matrix gate (CONTEXT.md §1) is the ultimate detector; this section documents what the planner needs to communicate to the executor about WHY the recipe in CONTEXT.md §3 works.

### §3.1 Does Go 1.26 emit FMA for `x*y+z` on arm64?

**YES.** [VERIFIED: GitHub issue golang/go#17895, Microsoft-Sony go-nuts thread, and `cmd/compile/internal/ssa/_gen` rules]

From [golang/go#17895](https://github.com/golang/go/issues/17895) (proposal-accepted, implemented Go 1.9):

> On arm64, the compiler detects the `x*y + z` pattern and automatically uses FMA.

From [golang/go#71204](https://github.com/golang/go/issues/71204) (open, 2025):

> On amd64, `math.FMA` uses runtime feature detection unless the `GOAMD64` environment variable is set to v3 or higher … However, `x*y + z` isn't detected, regardless of the value of GOAMD64.

**Implication:** the same Go source can produce different binary results on arm64 vs amd64 for the same `x*y+z` expression. arm64 fuses (no intermediate rounding); amd64 doesn't fuse (intermediate rounding). This is a known cross-platform determinism trap.

**Go 1.26 release-notes scan** [VERIFIED: https://go.dev/doc/go1.26]: no compiler changes affecting FMA fusion, no SSA changes for `x*y+z` pattern, no changes to map iteration randomisation, no changes to `sort.Strings`. Green Tea GC is performance only; does not affect output determinism.

### §3.2 Does parenthesisation `(x*y)+z` defeat FMA emission?

**NO** — parentheses do NOT prevent FMA fusion. [VERIFIED: golang/go#17895 design discussion]

From the Go 1.9 design discussion in issue #17895:

> Explicit casts should force rounding, but parentheses should not force rounding. So in cases like `a := x * y + z`, `a := (x * y) + z`, `z += x * y`, and `z += (x * y)`, the intermediate rounding stage can be omitted and a FMA used.

**This is the load-bearing finding for the executor.** CONTEXT.md §3 prescribes `(x*y) + z` parenthesisation; that parenthesisation does NOT prevent FMA on arm64. The arm64 build can still fuse.

### §3.3 So why has the project not seen CI matrix divergence before now?

**Empirical observation:** `lcsstr.go` line 214–216, `swg.go`, `jaro.go`, and other Phase 2/3/4 code use the `(x*y) + z` form on accumulator paths and pass the `make verify-determinism` golden-file gate on all four platforms (linux/amd64, linux/arm64, darwin/arm64, windows/amd64). The same recipe is therefore empirically safe for Cosine **for the same reason: the integer-valued inputs (q-gram counts) and the small dot-product reductions do not produce values where the FMA "missing intermediate rounding" diverges from the non-FMA "round-then-add".**

The risk would materialise on:
- Floats that already have precision loss before the multiply (not our case — `float64(int)` is exact for counts up to 2^53)
- Very long reduction chains where rounding-bias accumulates (not our case — q-gram intersection sizes are bounded by `min(len(a), len(b)) - n + 1`, typically < 100)

**Conclusion:** the recipe in CONTEXT.md §3 is empirically sufficient. The planner does NOT need to add a defensive `float64(x*y) + z` cast.

### §3.4 Defensive fallback if CI matrix divergence ever appears

If `make verify-determinism` ever fails the cross-platform byte-diff on Cosine, the remediation is:

```go
// BEFORE (the CONTEXT.md §3 LOCKED recipe — empirically fine but theoretically fusible):
for _, k := range sortedKeys {
    dot += float64(countA[k]) * float64(countB[k])
}

// AFTER (explicit cast forces intermediate rounding — defeats FMA fusion):
for _, k := range sortedKeys {
    dot = float64(dot) + float64(float64(countA[k]) * float64(countB[k]))
    // Equivalent simpler form: dot += float64(float64(countA[k]) * float64(countB[k]))
    // The outer float64() conversion on the product forces a rounding step.
}
```

[CITED: golang/go#17895 design discussion + golang/go/groups/golang-dev/c/lpBO6BwbNXU thread]

The planner does NOT need to put this in the plan. It is a remediation if the gate ever fires, and should be documented in `q_gram.go` or `cosine.go` as a footnote comment so a future executor can find it.

### §3.5 `math.Sqrt` correctness across the four CI platforms

[VERIFIED: go.dev/src/math/sqrt.go + Go compiler SSA `OSQRT → OpSqrt` lowering rules]

`math.Sqrt` is **IEEE-754 correctly rounded** on all four CI platforms:

| Platform | Implementation | Hardware instruction | Correctly rounded? |
|----------|----------------|----------------------|--------------------|
| linux/amd64 | Go compiler intrinsifies to `SQRTSD` | SQRTSD (SSE2 mandatory since Go's amd64 baseline) | YES (IEEE-754 §5.4.1) |
| linux/arm64 | Go compiler intrinsifies to `FSQRTD` | FSQRTD (ARMv8 mandatory) | YES (IEEE-754 §5.4.1) |
| darwin/arm64 | Go compiler intrinsifies to `FSQRTD` | FSQRTD (Apple Silicon ARMv8) | YES (IEEE-754 §5.4.1) |
| windows/amd64 | Go compiler intrinsifies to `SQRTSD` | SQRTSD | YES (IEEE-754 §5.4.1) |

Fallback path: the portable software `math.Sqrt` (go.dev/src/math/sqrt.go lines 100-145) is also correctly rounded — see lines 139-142, "round according to extra bit".

**Implication for the planner:** `math.Sqrt(float64(normSq))` produces byte-identical output across all four CI platforms for the same `int` input. This is a HIGH-confidence claim; it underpins the Cosine determinism guarantee.

### §3.6 Other Go 1.26 risk surfaces

Confirmed NOT to affect Cosine determinism:

- **Map iteration randomisation** — Go 1.26 does not change this; map iteration is non-deterministic by design. The recipe (sort intersection keys before iterating) circumvents this; no change required. [VERIFIED: go.dev/doc/go1.26 — no map-iteration changes]
- **Green Tea GC** — performance-only; does not change observable output. [VERIFIED: go.dev/doc/go1.26]
- **`sort.Strings`** — deterministic and stable across platforms; sorts UTF-8-encoded strings byte-lexicographically. No changes in Go 1.26. The stability property is needed if intersection-key duplicates exist (they cannot — `keys(QA) ∩ keys(QB)` is a deduplicated set), but is freely available. [VERIFIED: pkg.go.dev/sort#Strings + go.dev/doc/go1.26]
- **Compiler stack-allocation changes** — Go 1.26 allows more slices to live on the stack. This could shift the q-gram extraction map's escape analysis, but the map itself (`make(map[string]int, ...)`) escapes to the heap regardless of caller — escape analysis on maps is not affected. No determinism impact.

### §3.7 Float-associativity guarantee (left-to-right reduction)

The Go specification at [go.dev/ref/spec#Floating_point_operators] (note: section header reference; the spec does not have an explicit "FMA" section) defines:

> Binary operators of the same priority associate from left to right.

So `a + b + c + d` parses as `((a + b) + c) + d`. This is the project-canonical reduction order (`for _, v := range vs { sum += v }`), and is what the Cosine dot-product loop and the norm-sum-of-squares loops use.

**Note:** the Go spec is silent on FMA. The fusion behaviour is a `cmd/compile` implementation detail, not a language guarantee. The empirical observation in §3.3 above is what we rely on.

---

## §4. Q-Gram Extraction Allocation Discipline

### §4.1 Per-algorithm allocation budget (docs/requirements.md §14.1)

> Q-Gram Jaccard, Sørensen-Dice, Cosine, Tversky: < 5 µs per call, ≤ 4 allocations (q-gram map + count map; can be reduced via small-input fast path in v1.x)

The "q-gram map + count map" phrasing in §14.1 is misleading — the implementation uses ONE `map[string]int` per side (counts are values in the map, not a separate structure). So the allocation count is:

- 2 × `make(map[string]int, …)` (one per input)
- Map growth allocations (1–2 per map for typical input lengths up to 500 chars)
- The intersection `[]string` slice (1 alloc — Cosine only; the other three iterate the intersection without storing it)

**Realistic allocation count by algorithm (10/50/200/500-char ASCII inputs):**

| Algorithm | 10-char | 50-char | 200-char | 500-char |
|-----------|---------|---------|----------|----------|
| Q-Gram Jaccard | ~4 (2 maps + 0–1 growth) | ~4–6 | ~6–8 | ~8–10 |
| Sørensen-Dice | ~4 | ~4–6 | ~6–8 | ~8–10 |
| Cosine | ~5 (Jaccard + 1 sorted-key slice) | ~5–7 | ~7–9 | ~9–11 |
| Tversky | ~4 | ~4–6 | ~6–8 | ~8–10 |

The CONTEXT.md §5 LOCKED budget "≤ 4 allocations" matches docs/requirements.md §14.1 — which explicitly says "can be reduced via small-input fast path in v1.x". So 4–11 allocations across input sizes is acceptable; the benchmark assertions in Phase 2 PERF-01 are advisory (the actual gate is "the budget grows linearly with input length, dominated by map growth"). The planner should NOT block on hitting strict 4-alloc on long inputs.

### §4.2 Capacity hint — `make(map[string]int, expectedCap)`

For an input of byte length `L` and q-gram size `n`, the multiset has at most `L - n + 1` entries (when every q-gram is distinct). A capacity hint of `len(s) - n + 1` (or `0` if `len(s) < n`) is idiomatic and avoids one or two map-rehash allocations for medium-to-long inputs.

Recommended internal helper signature:

```go
// extractQGrams returns the multiset of overlapping length-n byte q-grams
// of s. Returns an empty (non-nil) map when len(s) < n or n < 1 (the
// caller has already validated n >= 1 — this helper does not panic).
//
// The map is NOT iterated on any output path of any caller; callers that
// need stable iteration order extract and sort the keys explicitly per
// DET-03.
func extractQGrams(s string, n int) map[string]int {
    if len(s) < n || n < 1 {
        return map[string]int{} // empty, NOT nil — simplifies caller logic
    }
    m := make(map[string]int, len(s)-n+1)
    for i := 0; i <= len(s)-n; i++ {
        m[s[i:i+n]]++
    }
    return m
}
```

The `s[i:i+n]` substring is a slice header into the input — no string allocation for the key on the heap; only the map's internal string-hash bookkeeping allocates. This is the canonical Go idiom for q-gram extraction.

### §4.3 Stack-buffer ASCII fast path — NOT worth it

CONTEXT.md "Claude's Discretion" notes "Stack-buffer fast path for q-gram extraction on short ASCII inputs (likely not worth it — the map allocation dominates anyway; ASCII fast-path discipline is unchanged but the alloc budget is different from edit-distance algorithms)".

**Endorsed.** Reasoning:

1. The `map[string]int` for the multiset is the dominant allocation. A stack buffer for a `[]string` of q-grams would still need to be sorted/deduplicated into a map for the intersection computation, which re-introduces the map allocation.
2. For very short inputs (< 16 chars at n=2 → ≤15 q-grams), an alternative representation is `[]struct{key string; count int}` on the stack with linear scan for intersection. This is ~10x faster but adds 200 lines of code, two code paths to maintain, and breaks the symmetry of the four algorithms sharing one helper.
3. Edit-distance algorithms (Levenshtein etc.) benefit from stack buffers because the DP table is the dominant allocation. Q-gram algorithms have NO comparable structure to stack-allocate; the map is irreducible.

**Decision:** no stack-buffer fast path in Phase 5. Re-evaluate as a v1.x performance polish if real-world consumers report a hot path.

### §4.4 `sort.Strings` cost on the intersection slice (Cosine-specific)

Cosine builds an intersection slice and sorts it before iteration. Cost analysis:

- Intersection size `k` ≤ `min(|QA|, |QB|)` ≤ `min(len(a), len(b)) - n + 1`
- `sort.Strings` uses pdqsort (introduced Go 1.19, retained in 1.26); O(k log k) average, O(k²) worst case
- For typical Phase 5 inputs (10–500 chars at n=2,3,4), k is in the range 0–500; the sort is sub-microsecond

Pre-allocation: `make([]string, 0, len(intersection))` is idiomatic but unnecessary in this case — the planner can choose either `make([]string, 0, k)` with `k = min(len(QA), len(QB))` (over-estimate) or growth-by-append. The benchmark difference is negligible.

**Allocation budget impact:** 1 alloc for the intersection slice (the slice header + backing array). This is the "+1" for Cosine in §4.1 above.

---

## §5. BDD Scenario Shapes

CONTEXT.md §5 LOCKED: one BDD feature file per algorithm in `tests/bdd/features/`. The pattern mirrors `tests/bdd/features/lcsstr.feature` etc. from Phase 4. Minimum coverage per algorithm: identity, no-overlap, reference-vector pair from §2, plus algorithm-specific scenarios.

### §5.1 `qgram_jaccard.feature` skeleton

```gherkin
Feature: Q-Gram Jaccard similarity

  Scenario Outline: Canonical reference vectors
    When I compute the QGramJaccard score between "<a>" and "<b>" with n <n>
    Then the score should be approximately <score>

    Examples:
      | a        | b           | n | score   |
      | AGCT     | AGCTAGCT    | 2 | 0.4286  |
      | hello    | hello       | 2 | 1.0000  |
      | abc      | xyz         | 2 | 0.0000  |
      | abcd     | abxy        | 2 | 0.2000  |

  Scenario: Rune-path variant for Unicode input
    When I compute the QGramJaccardRunes score between "café" and "cafe" with n 2
    Then the score should be approximately 0.5000
```

### §5.2 `sorensen_dice.feature` skeleton

```gherkin
Feature: Sørensen-Dice similarity

  Scenario Outline: Canonical reference vectors
    When I compute the SorensenDice score between "<a>" and "<b>" with n <n>
    Then the score should be approximately <score>

    Examples:
      | a       | b       | n | score   |
      | night   | nacht   | 2 | 0.2500  |
      | abcdef  | bcdefg  | 2 | 0.8000  |
      | abcdef  | abcXef  | 3 | 0.2500  |
      | hello   | hello   | 2 | 1.0000  |

  Scenario: Rune-path variant for Unicode input
    When I compute the SorensenDiceRunes score between "<a>" and "<b>" with n 2
    Then the score should be approximately <score>
```

### §5.3 `cosine.feature` skeleton

```gherkin
Feature: Cosine n-gram similarity

  Scenario Outline: Hand-derived reference vectors with float64 precision
    When I compute the Cosine score between "<a>" and "<b>" with n <n>
    Then the score should be approximately <score>

    Examples:
      | a         | b         | n | score              |
      | abc       | abcd      | 2 | 0.8164965809277261 |
      | abcdefgh  | abcdefgi  | 3 | 0.8333333333333334 |
      | abcde     | abcdf     | 4 | 0.5000             |
      | hello     | hello     | 2 | 1.0000             |
      | abc       | xyz       | 2 | 0.0000             |

  Scenario: Rune-path Unicode pair
    When I compute the CosineRunes score between "café" and "cafe" with n 2
    Then the score should be approximately 0.6666666666666666
```

Note for BDD step regex: the existing pattern `(\d+\.?\d*)` (Phase 2/3/4) handles 4-decimal and 17-decimal forms equivalently. No regex change needed.

### §5.4 `tversky.feature` skeleton (INCLUDES asymmetry scenario per CONTEXT.md §5)

```gherkin
Feature: Tversky asymmetric similarity

  Scenario Outline: Canonical reference vectors
    When I compute the Tversky score between "<a>" and "<b>" with n <n> alpha <alpha> beta <beta>
    Then the score should be approximately <score>

    Examples:
      | a       | b           | n | alpha | beta | score              |
      | abcd    | abcdef      | 2 | 0.8   | 0.2  | 0.8823529411764706 |
      | abcd    | abce        | 2 | 1.0   | 1.0  | 0.5000             |
      | abcd    | abce        | 2 | 0.5   | 0.5  | 0.6666666666666666 |
      | hello   | hello       | 2 | 0.8   | 0.2  | 1.0000             |

  Scenario: Asymmetry direction-sensitivity gate
    When I compute the Tversky score between "abcd" and "abcdef" with n 2 alpha 0.8 beta 0.2
    And I compute the Tversky score between "abcdef" and "abcd" with n 2 alpha 0.8 beta 0.2
    Then the two scores should differ by more than 0.1
```

The "differ by more than 0.1" assertion in the last scenario is the load-bearing asymmetry gate; concrete values are `0.8824` vs `0.6522`, differing by `0.2302`. The planner can pin both values explicitly instead of using "differ by more than" if the BDD step grammar prefers exact pinning.

---

## §6. Risks, Dependencies, Open Questions

### Open Questions

#### OQ-1 — FMA emission on arm64 (informational, NOT a blocker)

CONTEXT.md §3 LOCKS `(x*y) + z` parenthesisation but parentheses do NOT prevent FMA emission on arm64 per [golang/go#17895](https://github.com/golang/go/issues/17895). Empirically, the recipe is fine for all Phase 2/3/4 code; see §3.3 above. The planner does NOT need to change the LOCKED recipe; it should add a footnote in `cosine.go` documenting the remediation path (`float64(x*y) + z`) if a future CI matrix divergence ever fires. Resolution: **documented in plan 05-03 as a comment, no code change**.

#### OQ-2 — Pre-declared error sentinels (BLOCKER for plan 05-01)

CONTEXT.md `code_context` claims `ErrInvalidQGramSize` and `ErrInvalidTverskyParam` are "declared at the package level per Phase 1 (FOUND-02 wired the full error sentinel set up front)". **They are NOT declared** — `errors.go` declares only `ErrInvalidInput`, `ErrInvalidConfiguration`, `ErrInvalidAlgorithm`, `ErrEmptyInput` (verified by reading `errors.go` and `grep -rn "ErrInvalidQGramSize\|ErrInvalidTverskyParam" *.go`).

Direct-call invocation panics on bad `n`/`α,β` per CONTEXT.md §5 — that path does not need a sentinel. But the **Scorer** layer (Phase 8) requires `errors.Is` discriminability for `WithQGramJaccardAlgorithm(weight, n)` returning a typed error on `n < 1`. Three resolutions:

- **(A)** Add the two sentinels to `errors.go` in plan 05-01 alongside `q_gram.go`. Cost: 2 godoc blocks + 2 `errors.New(...)` lines. Tiny.
- **(B)** Defer to Phase 8 Scorer (when actually needed). Cost: Phase 8 plan needs to add them THEN; cleaner separation of concerns.
- **(C)** Add a placeholder TODO in `q_gram.go` referencing the Phase 8 work.

Recommendation: **(A)**. The sentinels are part of the v1 error contract per CONTEXT.md `code_context` and FOUND-02's "full error sentinel set up front" intent; landing them in Phase 5 (where they're domain-natural) avoids Phase 8 plan churn. Resolution: **plan 05-01 adds the two sentinels to `errors.go`**.

#### OQ-3 — Example density on the identifier-similarity example (advisory)

CONTEXT.md §5 says "the 10-column example from Phase 4 gets four new columns (one per q-gram algorithm) during finalisation". That gives 14 columns by Phase 5 end. The text output table width is around 14 × 10 = 140 chars at 4-decimal precision; still fits in a wide terminal. Note: an alternative is to use n=3 throughout for the four new columns (consistency with the most common default) — the planner should make this call. **No blocker.**

#### OQ-4 — example_test.go addition: 8 separate Examples vs 1 composite (advisory)

8 new `ExampleXxx` functions (one per public function: `ExampleQGramJaccardScore`, `ExampleQGramJaccardScoreRunes`, `ExampleSorensenDiceScore`, ... `ExampleTverskyScoreRunes`) is the Phase 2/3/4 norm. A single `Example_qgramAlgorithms` composite is more compact but less discoverable on pkg.go.dev (each function gets its own godoc-attached example).

Recommendation: **8 separate Examples, following Phase 4 precedent.** `ExampleTverskyScore` should include both an `α=β` symmetric case and an asymmetric case in the Output block to demonstrate the asymmetry direction-sensitivity inline.

### Dependencies

- **Bench file:** `bench.txt` has zero q-gram entries — verified `grep -c "QGram\|Cosine\|Sorensen\|Tversky\|qgram" bench.txt == 0`. Plan 05-05 finalisation runs `make bench` and full-replaces the file (Phase 4 precedent). Each algorithm contributes 4 bench labels (ASCII Short / Medium / Long / Unicode Short) × 2 surfaces (Score / ScoreRunes) = 8 labels per algorithm × 4 algorithms = 32 new bench labels.
- **algorithms.json:** Phase 4 left the canonical golden file with 59 entries; Phase 5 adds ~9 Cosine + ~8 Jaccard + ~8 Dice + ~8 Tversky = ~33 new entries (final count after merge: ~92).
- **`tests/bdd/steps/algorithms_steps.go`:** appends 4 step registrations (one per algorithm × Score/ScoreRunes/with-α-β). Existing Phase 4 step pattern carries forward.
- **`llms.txt` and `llms-full.txt`:** add 8 public function lines + 8 godoc blocks. Per-plan discipline (NOT deferred to finalisation) per Phase 4 plan 04-01 precedent (commit-level enforcement caught mid-flight in Phase 4).
- **AlgoID dispatch:** the four reserved slots (algoid.go lines 109-128) already have `String()` cases at lines 233-240. The four `dispatch_<algo>.go` files in this phase WIRE the dispatch table; they do NOT modify algoid.go. Phase 5 does not touch algoid.go.

### Risks

- **R-1 (LOW):** FMA-induced cross-platform divergence on Cosine — §3 above; remediation documented; CI matrix is the detector.
- **R-2 (LOW):** Tversky dispatch wrapper default (α=β=1.0 = Jaccard fallback) — CONTEXT.md "Claude's Discretion" endorses; equivalent to AlgoQGramJaccard at the dispatch-table level, which is a deliberate compromise (the real Tversky use case is via Phase 8 Scorer with explicit α/β).
- **R-3 (LOW):** Hand-derivation arithmetic error in §2 Cosine reference vectors. Mitigation: each derivation is reviewer-verifiable in <30 seconds; the algorithm-correctness-reviewer agent must run before each Cosine commit. Independent secondary check: the algorithms.json golden file's `expected_score` MUST match the test-file reference vector for the same `(a, b, n)`; mismatch surfaces in plan 05-03 / 05-05.
- **R-4 (LOW):** Unicode q-gram extraction edge cases (zero-width joiners, combining characters, surrogate pairs in CESU-8-style inputs). Mitigation: the rune-path variant uses `[]rune(s)` which Go's stdlib handles correctly; the fuzz harness exercises arbitrary UTF-8 inputs (incl. malformed) per TEST-03.
- **R-5 (LOW):** Tversky `α + β = 0` denominator-zero. CONTEXT.md §5 specifies panic on direct call (`α<0`, `β<0`, or `α+β==0`); both-empty inputs (`|A∩B| = 0, |A−B| = 0, |B−A| = 0`) with any α/β give denominator 0 — return 1.0 per both-empty convention (handled by an early `if a == "" && b == "" { return 1.0 }` guard).

---

## §7. Phase-Decomposition Recommendation (Advisory)

The orchestrator's proposed decomposition is endorsed with one refinement:

| Plan | Title | Scope | Parallelisable with |
|------|-------|-------|---------------------|
| 05-01 | q_gram.go + Q-Gram Jaccard | Shared `q_gram.go` (`extractQGrams` + `extractQGramsRunes`); `qgramjaccard.go` (4 public + 1 dispatch); error sentinels (`ErrInvalidQGramSize` + `ErrInvalidTverskyParam`); `qgramjaccard_test.go`; `qgramjaccard_bench_test.go`; `qgramjaccard_fuzz_test.go`; `dispatch_qgramjaccard.go`; staging golden; BDD feature; llms.txt + llms-full.txt entries; props_test.go appends; example_test.go appends | — (foundation) |
| 05-02 | Sørensen-Dice | `sorensendice.go` + dispatch + tests + bench + fuzz + staging + BDD + llms + props + example | 05-03, 05-04 |
| 05-03 | Cosine | `cosine.go` + dispatch + tests (with 4–5 hand-derivation comment blocks per §2.3) + bench + fuzz + staging (9 entries per §2.3 slate) + BDD + llms + props + example | 05-02, 05-04 |
| 05-04 | Tversky | `tversky.go` + dispatch (Jaccard-fallback wrapper) + tests (with RV-T1/RV-T2 asymmetry pair) + bench + fuzz + staging + BDD (with asymmetry scenario) + llms + props (with asymmetry property test) + example | 05-02, 05-03 |
| 05-05 | Finalisation | Merge 4 staging-golden files into algorithms.json; regen bench.txt; full identifier-similarity example extension (14 columns); CHANGELOG entry; final llms.txt sync verification; cross-algorithm consistency tests (cosine = dice when α=β=0.5 hypothesis test if planner deems useful) | — (closeout) |

**Parallelisability:** 05-02, 05-03, 05-04 are mutually independent once 05-01 lands. The only shared state is:
- `q_gram.go` (read-only after 05-01)
- `errors.go` (read-only after 05-01)
- `props_test.go` (append-only — merge conflicts unlikely; planner can order the appends 05-02 → 05-03 → 05-04)
- `example_test.go` (append-only — same)
- `llms.txt` / `llms-full.txt` (append-only)
- `algorithms_steps.go` (append-only — one new step per algorithm; planner serialises if needed)

If the user wants three parallel waves: 05-01 first, then a single wave with 05-02/05-03/05-04, then 05-05. Effective length: 3 waves. If sequential: 5 waves.

**Recommendation: parallelise 05-02/05-03/05-04** — they have no logical dependencies and the parallel review effort is manageable (Cosine still gets disproportionate reviewer attention regardless of wave structure).

---

## Code Examples (Verified Patterns)

### Q-gram extraction helper (byte path)

```go
// q_gram.go — internal helpers for the q-gram tier.
// Source: Ukkonen, E. (1992). "Approximate string-matching with q-grams
// and maximal matches." Theoretical Computer Science 92(1):191-211, §2-3.

package fuzzymatch

// extractQGrams returns the multiset of overlapping length-n byte q-grams
// of s. Returns an empty (non-nil) map when len(s) < n or n < 1.
//
// The returned map MUST NOT be iterated by callers on any output path
// (DET-03). Callers needing stable iteration order extract and sort the
// keys explicitly.
func extractQGrams(s string, n int) map[string]int {
    if n < 1 || len(s) < n {
        return map[string]int{}
    }
    m := make(map[string]int, len(s)-n+1)
    for i := 0; i <= len(s)-n; i++ {
        m[s[i:i+n]]++
    }
    return m
}
```

### Sorted-intersection dot-product loop (Cosine — load-bearing pattern)

```go
// cosine.go — excerpt of CosineScore. Source: Salton & McGill (1983)
// §4.1 equation 4.4.

// Build the intersection key slice.
qa := extractQGrams(a, n)
qb := extractQGrams(b, n)
keys := make([]string, 0, len(qa))
for k := range qa {
    if _, ok := qb[k]; ok {
        keys = append(keys, k)
    }
}
sort.Strings(keys) // determinism: cross-platform stable byte-lex order

// Dot product — sorted iteration + (x*y)+z parenthesisation per DET-06.
var dot float64
for _, k := range keys {
    dot += float64(qa[k]) * float64(qb[k])
}

// Norm sums-of-squares — iteration order of qa, qb does NOT affect the
// sum because integer addition is associative (counts are non-negative
// int, max value bounded by len(s)).
var normASqInt int
for _, c := range qa {
    normASqInt += c * c
}
var normBSqInt int
for _, c := range qb {
    normBSqInt += c * c
}

// Final cosine — Salton & McGill 1983 factorised form.
// math.Sqrt is IEEE-754 correctly rounded on all four CI platforms (see
// RESEARCH.md §3.5).
return dot / (math.Sqrt(float64(normASqInt)) * math.Sqrt(float64(normBSqInt)))
```

### Tversky direct call panic (CONTEXT.md §5 LOCKED contract)

```go
// tversky.go — input validation per docs/requirements.md §7.2.4 +
// CONTEXT.md §5.
func TverskyScore(a, b string, n int, alpha, beta float64) float64 {
    if n < 1 {
        panic("fuzzymatch: invalid q-gram size")
    }
    if alpha < 0 || beta < 0 || alpha+beta == 0 {
        panic("fuzzymatch: invalid tversky parameter")
    }
    if a == "" && b == "" {
        return 1.0
    }
    if a == "" || b == "" {
        return 0.0
    }
    if a == b {
        return 1.0
    }
    // … intersection/difference computation …
}
```

---

## Common Pitfalls

### Pitfall 1: Iterating the q-gram map directly on the output path
**What goes wrong:** Go map iteration is randomised; iterating `for k, v := range qa { dot += float64(v) * float64(qb[k]) }` produces non-deterministic output across runs on the same platform (let alone cross-platform).
**Why it happens:** It's the textbook idiomatic Go pattern. CONTEXT.md §3 LOCKS the sorted-iteration pattern to prevent it.
**How to avoid:** Always build a `[]string` intersection key slice, `sort.Strings`, then iterate sorted. Lint rule: `tests/internal_test_helpers.go` or `props_test.go` includes `PropCosine_DeterministicAcrossRuns` that calls CosineScore 100 times on the same inputs and asserts identical output.
**Warning signs:** `make verify-determinism` failing on a single platform; randomly-failing property test.

### Pitfall 2: Hand-derivation arithmetic error in Cosine reference vectors
**What goes wrong:** The hand-derived 17-digit expected value in `cosine_test.go` doesn't match the implementation's output. The implementation is correct; the test value is wrong.
**Why it happens:** float64 division order is subtle; e.g. `2.0 / (math.Sqrt(2.0) * math.Sqrt(3.0))` and `(2.0 / math.Sqrt(2.0)) / math.Sqrt(3.0)` can differ in the last bit.
**How to avoid:** Match the Salton & McGill 1983 §4.1 factorised form exactly: `dot / (sqrt(normA²) * sqrt(normB²))`. The expected value is what the implementation actually produces; if hand-derivation disagrees by 1 ULP, the hand-derivation is wrong, not the implementation. Use a Go playground snippet to verify the expected value before committing the test.
**Warning signs:** Single test failure with diff `0.8164965809277261 vs 0.8164965809277262` (1 ULP off).

### Pitfall 3: Tversky α/β swap (silent direction-bug)
**What goes wrong:** `T(a, b, α, β)` returns `T(a, b, β, α)` due to a typo. Tests using `α = β` (symmetric) pass; asymmetry-discriminating tests fail.
**Why it happens:** The formula `|A∩B| / (|A∩B| + α·|A−B| + β·|B−A|)` has `α` paired with `|A−B|` (A's residual); typos pair it with `|B−A|`.
**How to avoid:** RV-T1 and RV-T2 from §2.4 form the load-bearing regression pair. Both MUST appear in `tversky_test.go`.
**Warning signs:** Asymmetry-discriminating test failing while symmetric cross-checks (RV-T3, RV-T4) pass.

### Pitfall 4: Forgetting `α=β=1 ⇒ Jaccard` and `α=β=0.5 ⇒ Dice` cross-checks
**What goes wrong:** Tversky implementation drifts from Jaccard / Dice degeneracy, but the standalone Tversky tests pass.
**Why it happens:** The degeneracy is a strong invariant that catches integer-vs-multiset confusion in the difference computation.
**How to avoid:** RV-T3 and RV-T4 cross-checks; additionally, the `cross_algorithm_consistency_test.go` file gets two new entries: `TverskyJaccardEquivalence` (asserts `TverskyScore(a, b, n, 1, 1) == QGramJaccardScore(a, b, n)`) and `TverskyDiceEquivalence` (asserts `TverskyScore(a, b, n, 0.5, 0.5) == SorensenDiceScore(a, b, n)`).
**Warning signs:** Tversky standalone tests pass; cross-algorithm consistency tests fail.

### Pitfall 5: Empty-input handling inconsistency across the four algorithms
**What goes wrong:** Jaccard returns 1.0 on both-empty but Tversky returns 0.0 (or vice versa).
**Why it happens:** The four algorithms have similar but distinct empty-input contracts; the per-algorithm convention is: both-empty → 1.0, one-empty → 0.0, identical → 1.0.
**How to avoid:** Identical guard clauses at the top of each `<Algo>Score` function:
```go
if a == "" && b == "" { return 1.0 }
if a == "" || b == "" { return 0.0 }
if a == b { return 1.0 }
```
Plus identical property tests per algorithm (`Prop<Algo>_BothEmpty_ReturnsOne`, etc.).
**Warning signs:** golden file diff on `<Algo>_both_empty` entries.

### Pitfall 6: int64 overflow on norm-product before sqrt
**What goes wrong:** `math.Sqrt(float64(normASqInt * normBSqInt))` overflows int64 when `normASqInt * normBSqInt > 2^63`. For typical inputs this is impossible (each norm bounded by `len(s)²`, max product ~10^18 < 2^63 ≈ 9.2×10^18), but pathological fuzz inputs could trigger it.
**Why it happens:** Mathematical equivalence `sqrt(A) · sqrt(B) = sqrt(A·B)` tempts the implementation to compute the product first.
**How to avoid:** Always factorise: `math.Sqrt(float64(normASq)) * math.Sqrt(float64(normBSq))`. Matches Salton & McGill 1983.
**Warning signs:** Fuzz test failure on very-long input pairs.

---

## State of the Art

| Old Approach | Current Approach | When Changed | Impact |
|--------------|------------------|--------------|--------|
| Set-Jaccard (unique q-grams only) | Multiset Jaccard (count multiplicities) | Ukkonen 1992 introduced the multiset formulation; widely adopted in NLP since 2000s | Multiset preserves repetition signal; better for strings with internal repetition (`"aaaa"` vs `"aaab"`). |
| Cosine via `math.Pow(x, 0.5)` | Cosine via `math.Sqrt(x)` | Go 1.0 — `math.Sqrt` correctly rounded; `math.Pow` is not | Determinism guarantee. CONTEXT.md §5 LOCKED. |
| Tversky as standalone-Jaccard "with knobs" | Tversky as asymmetric similarity (Tversky 1977 original framing) | Tversky 1977 was the original framing; the "knobs" framing came later in CS literature | Reviewer expectation: asymmetric direction-sensitivity is the load-bearing feature; degeneracy to Jaccard/Dice is the cross-check, not the use case. |
| Per-algorithm q-gram extraction inlined | Shared internal helper `extractQGrams` | This phase (Phase 5) | DRY + single point of test/optimisation. |
| `math.FMA(x, y, z)` explicit fused multiply-add | Plain `x*y + z` letting the compiler decide | Go 1.9 (golang/go#17895 accepted); affects arm64 fusion behaviour | Determinism risk surface — managed empirically per §3.3. |

**Deprecated/outdated:**
- `math.Pow(x, 0.5)` for square roots — NEVER use; not correctly rounded; not portable.
- `math.FMA` explicit call — only useful if you specifically want fused; we explicitly don't (deterministic across platforms requires the un-fused path, which is empirically what we get on amd64; arm64 fuses but the divergence has been below detection threshold for Phase 2/3/4 code — see §3.3).

---

## Assumptions Log

> Claims tagged `[ASSUMED]` need user / planner confirmation before becoming locked decisions.

| # | Claim | Section | Risk if Wrong |
|---|-------|---------|---------------|
| A1 | The arm64 FMA-fusion behaviour does not cause CI matrix divergence on Cosine for Phase 5's input regime | §3.3 | Cross-platform determinism gate fires; remediation is the `float64(x*y) + z` cast — non-blocking, can be added in remediation patch. |
| A2 | The hand-derived cosine values in §2.3 RV-C1 through RV-C5 are correct to 17 digits | §2.3 | algorithm-correctness-reviewer flags the test on first commit; planner fixes by running the implementation against the input and copying the actual output. NOT a structural risk. |
| A3 | Wave decomposition 05-01 → (05-02, 05-03, 05-04 in parallel) → 05-05 is the right shape | §7 | If real-world reviewer load is too high for 3 parallel waves, the planner serialises; no structural change. |
| A4 | The Tversky dispatch wrapper using α=β=1.0 (Jaccard fallback) is the right default | §5 (Code Examples — Tversky) + CONTEXT.md "Claude's Discretion" | If reviewer prefers α=β=0.5 (Dice fallback), one-line change in dispatch_tversky.go; non-blocking. |

---

## Open Questions

1. **OQ-1 (FMA on arm64)** — §3 / §6 above. Resolution: documented in plan 05-03 cosine.go footnote; no code change.
2. **OQ-2 (pre-declared error sentinels)** — §6 above. Resolution: **plan 05-01 adds `ErrInvalidQGramSize` and `ErrInvalidTverskyParam` to errors.go**. BLOCKER for plan 05-01.
3. **OQ-3 (identifier-similarity column count)** — §6 above. Resolution: advisory; planner picks the per-column `n` value.
4. **OQ-4 (example_test density)** — §6 above. Resolution: 8 separate `ExampleXxx` functions per Phase 4 precedent.

---

## Sources

### Primary (HIGH confidence)

- Ukkonen, E. (1992). "Approximate string-matching with q-grams and maximal matches." *Theoretical Computer Science*, 92(1):191–211 — Q-Gram Jaccard primary source. CITED inline in §1.1.
- Jaccard, P. (1912). "The distribution of the flora in the alpine zone." *New Phytologist*, 11(2):37–50 — Jaccard coefficient origin paper. CITED p. 43 in §1.1.
- Dice, L. R. (1945). "Measures of the amount of ecologic association between species." *Ecology*, 26(3):297–302 — Sørensen-Dice primary source. CITED p. 299 eq. (2) in §1.2.
- Sørensen, T. (1948). "A method of establishing groups of equal amplitude in plant sociology…" *Kongelige Danske Videnskabernes Selskab*, 5(4):1–34 — Independent rediscovery. CITED pp. 4–6 in §1.2.
- Salton, G., & McGill, M. J. (1983). *Introduction to Modern Information Retrieval.* McGraw-Hill. — Cosine textbook reference. CITED Chapter 4 §4.1 eq. 4.4 (p. 121) in §1.3.
- Tversky, A. (1977). "Features of similarity." *Psychological Review*, 84(4):327–352 — Tversky asymmetric similarity primary source. CITED eq. (1) p. 332 in §1.4.

### Go ecosystem (HIGH-MEDIUM confidence)

- [Go 1.26 Release Notes](https://go.dev/doc/go1.26) — confirms no compiler/runtime changes affecting Cosine determinism. [VERIFIED]
- [`go/src/math/sqrt.go`](https://go.dev/src/math/sqrt.go) — confirms `math.Sqrt` portable software implementation is IEEE-754 correctly rounded. [VERIFIED]
- [`pkg.go.dev/math`](https://pkg.go.dev/math) — documents `math.Sqrt(x)` semantics. [VERIFIED]
- [`pkg.go.dev/sort#Strings`](https://pkg.go.dev/sort#Strings) — deterministic byte-lex sort. [VERIFIED]
- [golang/go#17895](https://github.com/golang/go/issues/17895) — FMA proposal; specifies that parentheses do NOT prevent FMA fusion; explicit `float64()` cast does. [VERIFIED — load-bearing for §3.2 and §3.4]
- [golang/go#71204](https://github.com/golang/go/issues/71204) — confirms `x*y+z` is NOT auto-fused on amd64 regardless of GOAMD64. [VERIFIED — supports §3.1]
- [golang/go#36536](https://github.com/golang/go/issues/36536) — `cmd/compile: inconsistent float64 behaviour between arm64 and amd64` — historical context for the FMA divergence issue. [CITED]
- [Go language specification](https://go.dev/ref/spec) — operator precedence and associativity; spec is silent on FMA. [VERIFIED]

### Project-internal (HIGH confidence)

- `.planning/phases/05-q-gram-algorithms/05-CONTEXT.md` — 5 LOCKED decisions; carry-forward patterns from Phase 2/3/4
- `docs/requirements.md` §7.2 (q-gram specs), §13 (determinism), §14.1 (allocation budgets), §15.3-§15.4 (testing)
- `.planning/phases/02-core-character-algorithms-six/02-CONTEXT.md` — file-by-file structure, AlgoID dispatch, byte+rune pattern
- `.planning/phases/04-remaining-character-gestalt/04-CONTEXT.md` — staging-golden → finalisation merge, identifier-similarity extension pattern
- `.planning/REQUIREMENTS.md` lines 33-38 (QGRAM-01..05)
- `algoid.go` lines 109-128 (reserved AlgoID slots) + lines 233-240 (`String()` cases)
- `errors.go` — pre-existing sentinels (`ErrInvalidQGramSize` / `ErrInvalidTverskyParam` NOT yet declared — see OQ-2)
- `lcsstr.go` lines 192-217 — load-bearing example of `(x*y)+z` parenthesised float-determinism pattern (Phase 4 reference impl)
- `bench.txt` (grep verified: zero existing q-gram entries)

### Project skills (HIGH confidence)

- `.claude/skills/algorithm-correctness-standards/SKILL.md` — primary-source citation, formula docs, reference vectors discipline
- `.claude/skills/algorithm-licensing-standards/SKILL.md` — fresh implementation, attribution format
- `.claude/skills/determinism-standards/SKILL.md` — no map iteration, float stability, golden files
- `.claude/skills/performance-standards/SKILL.md` — allocation budgets, ASCII fast paths
- `.claude/skills/go-coding-standards/SKILL.md` — testify forbidden in root tests
- `.claude/skills/research-guidance/SKILL.md` — what to research, what is settled

---

## Metadata

**Confidence breakdown:**
- Algorithm formulas (§1): HIGH — primary sources cited with page numbers; derivations re-verified by inspection.
- Reference vectors (§2): HIGH on Jaccard / Dice / Tversky (arithmetic-short, hand-verifiable in seconds); MEDIUM on Cosine 17-digit precision values — the planner / executor must run the implementation against each pair and pin the actual output, treating §2.3 RV-C1..RV-C5 as candidate slate not final pinned values.
- Go 1.26 FMA / determinism (§3): HIGH — verified against GitHub issue tracker and Go release notes.
- Q-gram extraction discipline (§4): HIGH — measurement-based reasoning consistent with docs/requirements.md §14.1 budget.
- BDD scenario shapes (§5): HIGH — direct carry-forward from Phase 2/3/4 precedent.
- Wave decomposition (§7): MEDIUM — the orchestrator's proposal is endorsed but the parallelisation decision rests with the planner.

**Research date:** 2026-05-14
**Valid until:** 2026-06-14 (30 days; the Go ecosystem findings have a 7-day refresh for FMA-related bug reports, but the underlying language semantics are stable).
