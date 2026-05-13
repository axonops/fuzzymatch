# Pitfalls Research

**Domain:** Go string-similarity / fuzzy-matching library (correctness-first, deterministic, zero-dep)
**Researched:** 2026-05-13
**Confidence:** HIGH for algorithmic / determinism pitfalls (verified against primary sources, Go issue tracker, RapidFuzz docs); MEDIUM for some phonetic-variant pitfalls (multiple canonical formulations exist).

This file catalogues the domain-specific traps that fuzzy-matching libraries fall into. Each pitfall is mapped to a roadmap phase so the planner can prevent it instead of fixing it after the fact. Generic SE advice has been excluded — every entry below is specific to this library's design or to fuzzy-matching as a discipline.

---

## Critical Pitfalls

### Pitfall 1: OSA vs Full Damerau-Levenshtein output disagreement

**What goes wrong:**
A consumer expects "Damerau-Levenshtein distance" to be a single value but gets different numbers depending on which variant they call. On `"ca"` / `"abc"`, OSA returns 3 (`ca → a → ab → abc`) while full Damerau-Levenshtein returns 2 (`ca → ac → abc`). A library that ships only one but names it "Damerau-Levenshtein" silently surprises users coming from literature or from `hbollon/go-edlib`.

**Why it happens:**
The OSA restriction ("no substring edited more than once") is a *performance* simplification rarely emphasised in textbook descriptions. Many Wikipedia-derived implementations are OSA-only but advertise themselves as "Damerau-Levenshtein". The full variant requires the Lowrance-Wagner position-table formulation, which is a heavier piece of code.

**How to avoid:**
- Ship both variants with distinct names: `DamerauLevenshteinOSAScore` and `DamerauLevenshteinFullScore`. Distinct `AlgoID` constants. Distinct godoc sections. No alias called "Damerau-Levenshtein" — it is ambiguous.
- Include the canonical `"ca"`/`"abc"` divergence example as a reference vector for *both* algorithms in the unit tests, with a comment pointing at the Boytsov 2011 paper.
- Document in the godoc that OSA does NOT satisfy the triangle inequality (full DL does), so OSA is not a metric. Property tests must NOT assert triangle inequality on OSA.
- The `algorithm-correctness-reviewer` agent must specifically check that the OSA recurrence has the `min(...)` over the transposition case AFTER the standard Levenshtein step, not as a replacement.

**Warning signs:**
- Property test for OSA triangle inequality fails on `("ca", "ac", "abc")`. (This is *expected* — if OSA passes the triangle test, the implementation is actually full DL, which is a different bug.)
- Reference vector test on `"ca"`/`"abc"` returns 2 from OSA. Indicates accidental full-DL implementation.
- Property test for full DL triangle inequality fails. Indicates accidental OSA implementation.

**Phase to address:** Algorithm-implementation phase covering Damerau-Levenshtein (per requirements §7.1.2 and §7.1.3). Two separate issues, two separate PRs, two distinct AlgoIDs from the day the first variant lands.

---

### Pitfall 2: Jaro-Winkler boost threshold omitted, prefix cap wrong, or scale != 0.1

**What goes wrong:**
Three independent constants in Winkler 1990 are routinely mis-implemented:
- `winklerBoostThreshold = 0.7` — the Jaro score below which the prefix bonus is **not** applied. Many implementations skip this gate entirely, giving disproportionate boosts to low-Jaro pairs.
- `winklerMaxPrefix = 4` — the prefix-length cap. Some implementations use 3, some use 5, some use `min(prefix, max(len(a), len(b)) / 2)`.
- `winklerPrefixScale = 0.1` — the scaling factor `p`. Some implementations use 0.25 (which violates the `p · L_max ≤ 0.25` bound when combined with `L_max = 4`).

The Wikipedia article and many tutorial implementations omit the 0.7 boost threshold. Apache Commons Text has implemented the cap correctly (PREFIX_LENGTH_LIMIT = 4) but other implementations get this wrong.

**Why it happens:**
Winkler 1990 is a conference proceedings paper not always read end-to-end. Implementers crib from Wikipedia or a textbook summary that elides the boost threshold.

**How to avoid:**
- Declare all three as unexported package constants with godoc comments quoting Winkler 1990 directly (per `algorithm-correctness-standards.md`).
- Include reference vectors that *exercise* the boost threshold gate: pairs with Jaro < 0.7 must produce score = Jaro (no boost), pairs with Jaro ≥ 0.7 must produce score = Jaro + L·p·(1-Jaro).
- Property test: `JaroWinklerScore(a, b) ≥ JaroScore(a, b)` for any inputs (the boost is always non-negative).
- Property test: `JaroWinklerScore(a, b) ≤ JaroScore(a, b) + 0.4` (since `L·p ≤ 4·0.1 = 0.4` and `(1 - Jaro) ≤ 1`).
- The `algorithm-correctness-reviewer` agent must specifically verify all three constants against Winkler 1990 — not against Wikipedia.

**Warning signs:**
- Reference vector `"DIXON"`/`"DICKSONX"` produces JW ≠ 0.8133. (Jaro is 0.7667, so the boost gate fires; the canonical JW is 0.8133.)
- Property test `JW ≤ Jaro + 0.4` fails → scale or cap is wrong.
- Reference vector for two strings with Jaro < 0.7 returns JW > Jaro → boost gate is missing.

**Phase to address:** Algorithm-implementation phase covering Jaro / Jaro-Winkler / Strcmp95 (requirements §7.1.5–7.1.7). The three are usually implemented in the same PR or adjacent PRs; the boost-threshold pitfall affects Jaro-Winkler and Strcmp95.

---

### Pitfall 3: Gotoh 1982 has a known erratum in the affine gap recurrence

**What goes wrong:**
Gotoh's 1982 paper contains two documented mathematical mistakes — an indexing flip and an initialisation issue — that propagate into implementations cribbed directly from the paper. The initialisation error in particular affects implementations that copy the paper's matrix setup without consulting later corrections. Multiple textbook treatments (Gusfield, Waterman) also reproduce the error. A biorxiv preprint surveyed 31 lecture slides — 8 contained the mistake and 16 had incomplete initialisations.

**Why it happens:**
"Implement Smith-Waterman with Gotoh's affine gap penalty from the 1982 paper" sounds like a clear brief. The paper is a primary source per the project's discipline. But the paper itself has errors that an implementer working only from the primary source will reproduce.

**How to avoid:**
- Cite Gotoh 1982 as primary BUT note in the file's comment block that the implementation also consulted the EMBOSS or biopython reference implementations (both permissively licensed) for cross-validation of the recurrence. State explicitly: "Gotoh 1982 contains known errata in the initialisation step; this implementation uses the corrected formulation per ${reference}."
- Reference vectors must include affine-gap-sensitive inputs (e.g. one string with a single long gap should score equivalently regardless of how the gap is partitioned across cells — this catches the initialisation bug).
- Property test: monotonicity — increasing the gap-open penalty cannot increase the score; increasing the match reward cannot decrease the identity score `Score(x, x) = 1.0`.
- The `algorithm-correctness-reviewer` agent must specifically flag SWG and request the implementer cross-validate against at least one independent implementation (cross-validation only; no code copied).

**Warning signs:**
- Identity test `Score(x, x) = 1.0` fails for some inputs (initialisation bug).
- Score for a pair with one long gap is *better* than the score for the same pair after splitting the gap into two halves with no intervening match (gap-open vs gap-extend confusion).
- Asymmetry: `SWGScore(a, b) ≠ SWGScore(b, a)` (a flipped index in the recurrence).

**Phase to address:** Algorithm-implementation phase covering Smith-Waterman-Gotoh (requirements §7.1.8). This algorithm should be flagged in the roadmap as needing extra cross-validation — not just "primary source + property tests".

---

### Pitfall 4: Soundex H/W handling diverges between American/Census, SQL, and Russell variants

**What goes wrong:**
Four common Soundex variants exist:
- **American/Census 1880**: H and W are *not* separators — two consonants of the same group separated by H or W collapse to one digit. Vowels *are* separators.
- **SQL Server / MySQL**: H and W are treated like vowels — they act as separators.
- **Russell variant**: A/E/H/I/O/U/W/Y mapped to digit 9, then 9s removed.
- **NARA**: similar to American but adds extra rules for prefixes (Van, De).

`"Tymczak"` → `T522` in American Soundex but `T520` in SQL Soundex. A library that doesn't pick a variant *and document it* will get bug reports from both camps.

**Why it happens:**
"Soundex" is treated as one algorithm in common parlance, but it has been re-formulated multiple times across US Census decades. Implementers crib from Wikipedia's general description (which conflates variants) and from Knuth TAOCP Vol. 3 §6.4 (which describes a specific variant) without checking that the test data matches.

**How to avoid:**
- Pick one variant explicitly (the spec already commits to the Russell/Odell 1918+1922 algorithm canonicalised in Knuth TAOCP Vol. 3 §6.4 — requirements §7.4.1). State the variant by name and citation in the godoc.
- Include `"Tymczak"` in reference vectors with the chosen variant's expected output (`T522` for Knuth/Census). This single test case rules out the wrong-variant pitfall on its own.
- Document in the godoc that other Soundex variants exist and that consumers expecting SQL Soundex output should normalise differently or use NYSIIS.
- The `algorithm-correctness-reviewer` agent verifies that the variant name appears in the file's block comment.

**Warning signs:**
- Reference vector test `Soundex("Ashcraft") = Soundex("Ashcroft")` fails → likely an H/W-as-separator implementation.
- Reference vector test `Soundex("Tymczak") = "T522"` fails (returning `T520`) → SQL-variant implementation.

**Phase to address:** Algorithm-implementation phase covering Soundex (requirements §7.4.1). Trivial to prevent if the reference vectors are picked from Knuth before coding starts.

---

### Pitfall 5: NYSIIS iteration bound and Double Metaphone Spanish/Slavic/Greek rule branches

**What goes wrong:**
- **NYSIIS** has multiple published formulations (original 1970 Taft report, modified NYSIIS-1991, "wonderland" variants). The original truncates the code at 6 characters; modified variants keep more. The iteration over the input also differs — some variants iterate to fixed point, others apply each rule once.
- **Double Metaphone** has a primary key and a secondary key. The rule table is large (~200 conditional branches) covering Germanic, Slavic, Romance, Greek, and Chinese-origin English names. The most common bug is missing one of the language-origin branches — e.g. failing to emit the secondary `XMT` for `"Schmidt"`, which then doesn't match `"Smith"` (`SM0`/`XMT`). The full algorithm is ~400 lines of conditional logic with no clean mathematical structure.

**Why it happens:**
- NYSIIS: Taft 1970 is hard to obtain (NY State Special Report No. 1). Wikipedia summarises it; implementers code from the summary.
- Double Metaphone: Philips 2000 article + Philips' public-domain C reference is the primary source, but it is ~400 lines of conditionals. Implementers shortcut by porting from `CalypsoSys/godoublemetaphone` (forbidden by the fresh-implementation rule).

**How to avoid:**
- NYSIIS: commit to the canonical 1970 algorithm with 6-character truncation in the requirements doc (already done — §7.4.3). Include reference vectors from the Taft 1970 paper *and* from the New York State criminal-justice testing corpus where available. If Taft is hard to obtain, cite Knuth or a secondary review article that documents the algorithm.
- Double Metaphone: implementer must work from Philips' 2000 article *and* the public-domain C reference (which is the canonical implementation). Cross-validation against `dlclark/metaphone3` or `CalypsoSys/godoublemetaphone` is permitted *for reference vectors only*. The implementation file's block comment must list every reference consulted with its licence.
- Reference vectors must cover at least one input per language-origin branch (Germanic name, Slavic name, Romance name, Greek name) — verifies the multilingual rule table is wired up.
- The `algorithm-correctness-reviewer` agent flags Double Metaphone for extra scrutiny: rule-table count check, multilingual reference vector coverage, no code copied from the MIT Go ports.

**Warning signs:**
- `DoubleMetaphoneKeys("Schmidt")` does not return `("XMT", "SMT")` → Slavic branch missing.
- `DoubleMetaphoneKeys("Pacheco")` does not contain the Spanish `PXK` variant → Romance branch missing.
- Reference vector pair that should match (`"Catherine"`/`"Katherine"`) returns 0.0 → either keys don't match or primary/secondary cross-matching is wrong.
- NYSIIS code length > 6 for any input → truncation step missing.

**Phase to address:** Algorithm-implementation phase covering phonetic algorithms (requirements §7.4). Allocate extra time for Double Metaphone — the spec already estimates 400 lines (`docs/prior-art-research.md`); plan accordingly.

---

### Pitfall 6: Token-based ratios — disagreement about which "ratio" to use

**What goes wrong:**
`fuzzywuzzy`'s `ratio()` was originally based on Python's `difflib.SequenceMatcher`, which is Ratcliff-Obershelp. `RapidFuzz`'s `ratio()` always uses Indel similarity (`2·LCS/(|a|+|b|)`), regardless of fallback path. `fuzzywuzzy` later added a C extension that uses Indel, giving *different results between pure-Python and C-accelerated runs of the same library*. Consumers cross-validating against either library will see drift.

The requirements doc commits to Indel ratio (LCS-based) for Token Sort, Token Set, and Partial Ratio (§7.3.2–7.3.4). But "ratio" is genuinely ambiguous in the community.

**Why it happens:**
The fuzzywuzzy/thefuzz/RapidFuzz ecosystem has a documented inconsistency. RapidFuzz's `api_differences.md` explicitly calls it out. Implementers reading fuzzywuzzy docs and then cross-validating against RapidFuzz hit drift; implementers reading RapidFuzz and using fuzzywuzzy for ground truth hit drift.

**How to avoid:**
- Document explicitly in `TokenSortRatioScore`'s godoc: "Indel ratio (`2·LCS/(|a|+|b|)`) per RapidFuzz semantics. NOT Ratcliff-Obershelp `difflib.ratio()`." Same for Token Set and Partial Ratio.
- Reference vectors cross-validated against RapidFuzz, *not* fuzzywuzzy. State this in the file's block comment.
- The Ratcliff-Obershelp algorithm (`AlgoRatcliffObershelp`) is separately available for consumers who want the `difflib.ratio()` behaviour — point them at it from the Token Sort/Set godoc.
- The `api-ergonomics-reviewer` agent should consider whether the name "Ratio" carries enough specificity. (E.g. `TokenSortIndelRatioScore` is more explicit but uglier; the agent decides.)

**Warning signs:**
- Cross-validation against `seatgeek/thefuzz` pure-Python path returns different scores → that path uses difflib.
- Cross-validation against `rapidfuzz/RapidFuzz` matches → correct.
- Reference vector `TokenSortRatio("New York Mets", "Mets New York") = 1.0` should hold for any sensible implementation (post-sort identity).

**Phase to address:** Algorithm-implementation phase covering token-based algorithms (requirements §7.3). Document the Indel-vs-difflib decision in the per-algorithm godoc and link from FAQ.

---

### Pitfall 7: Byte-vs-rune indexing causes UTF-8 panics or silent wrong answers

**What goes wrong:**
A naïve `s[i:i+1]` on a string containing multi-byte UTF-8 returns a fragment of a code point. Levenshtein over byte slices treats `é` (U+00E9, two UTF-8 bytes 0xC3 0xA9) as two "characters", so `Levenshtein("café", "cafe") = 2` byte-level (correct for the byte interpretation) and `1` rune-level (correct for the human-reader interpretation). Q-gram extraction over bytes splits multi-byte code points across q-grams, producing garbage n-grams. Phonetic algorithms (Soundex etc.) panic or silently mis-encode if they don't validate ASCII first.

The requirements doc handles this by exposing both byte- and rune-level variants for character-based algorithms (§7.1) and committing q-gram and token algorithms to byte-level by default with a rune-level variant (§7.2 preamble). But the trap is in the *implementation* — silent partial reads of multi-byte code points.

**Why it happens:**
Go strings are byte arrays. `len(s)` returns bytes, not runes. `for i := 0; i < len(s); i++` iterates bytes. Implementers writing in "byte mode" forget that the byte-mode score is well-defined only when the *interpretation* of "the strings are sequences of bytes" is what the user actually wanted.

**How to avoid:**
- Every algorithm has both `XxxScore` (byte-level) and `XxxScoreRunes` (rune-level) where the distinction matters (per requirements §7 per-algorithm). Document in godoc which is which and what the consequence is for non-ASCII input.
- Fuzz tests on every algorithm with arbitrary `[]byte` corpus — including malformed UTF-8 sequences (lone surrogates, overlong encodings, truncated multi-byte sequences). The function must not panic.
- For phonetic algorithms: document explicitly that non-ASCII characters are dropped (Soundex, NYSIIS, MRA) or handled per the algorithm's rule (Double Metaphone may pass certain non-ASCII through; verify against the reference).
- The `security-reviewer` agent specifically checks for panic safety on malformed UTF-8 — this is a DoS vector if exposed to user input.

**Warning signs:**
- Fuzz test panics with "index out of range" or "invalid UTF-8 byte sequence" → an algorithm is doing unbounded byte indexing.
- Test on `"café"` returns a number that doesn't match either the byte interpretation OR the rune interpretation → mixed indexing.
- Q-gram test with `n=3` on `"café"` returns fewer n-grams than expected → byte/rune mismatch.

**Phase to address:** Every algorithm-implementation phase. Each algorithm needs:
1. A `_test.go` test with `"café"`, `"漢字"`, and `"🎯emoji"` inputs.
2. A `_fuzz_test.go` corpus that includes malformed UTF-8 (per `algorithm-correctness-standards.md` §"Unicode Handling").
3. A fuzz target run for >= 60s in `make test-fuzz` CI.

---

### Pitfall 8: Map iteration leaks into output order — silent non-determinism

**What goes wrong:**
Q-gram extraction uses `map[string]int` for multiset counts. The `scan` sub-package uses maps for token buckets. If *any* function builds output from `for k, v := range someMap`, the output order varies across runs. Tests pass locally (Go's map iteration randomisation happens at runtime, so deterministic test runs do exist by chance), then break in CI on a different platform or under `-shuffle=on`.

The `determinism-standards.md` skill calls this out specifically as "The No-Map-Iteration Rule". But it remains the single most common determinism bug.

**Why it happens:**
Maps are the idiomatic Go data structure for set/multiset operations. The shortcut from "I have a map" to "I'll iterate it to build my output" is one keystroke. The bug is invisible in unit tests that compare unordered sets (`reflect.DeepEqual` on a map handles unordered comparison).

**How to avoid:**
- Property test on every score function and the scan output: `PropAlgorithm_DeterministicAcrossRuns` calling the function 1000 times with the same input and asserting bit-identical output (per requirements §13.6 and `determinism-standards.md`).
- Golden file test in `testdata/golden/` regenerated only on deliberate score changes (per requirements §13.5).
- Cross-platform CI matrix running the golden test on linux/amd64, linux/arm64, darwin/arm64, darwin/amd64, windows/amd64 — catches any platform-specific iteration-order accidents (per requirements §13.3).
- `make verify-determinism` target that consumers can run locally before submitting.
- The `determinism-reviewer` agent reviews every PR for map iteration in output paths.

**Warning signs:**
- `PropAlgorithm_DeterministicAcrossRuns` fails for any algorithm.
- Golden file test fails on one platform but passes on others.
- `go test -shuffle=on -count=10` produces different output than `go test -count=1`.
- BDD scenario "scan output is deterministic" fails intermittently.

**Phase to address:** Foundation phase (CI matrix, golden file infrastructure, property test harness). Every subsequent algorithm phase inherits the test infrastructure.

---

### Pitfall 9: Float reductions order-of-summation gives different results across architectures

**What goes wrong:**
Cosine similarity, Monge-Elkan, and the Scorer composite all sum sequences of floats. `Σ x_i` summed left-to-right does not equal `Σ x_i` summed right-to-left or via Kahan summation when intermediate rounding differs. The result is bit-identical on one platform, off by 1 ULP on another. Cross-platform CI catches it, but only after a regression has landed.

Compiler-level FMA (fused multiply-add) optimisation can also produce non-bit-identical results across `GOARCH` values. On arm64 the compiler detects `x*y + z` and emits FMA; on amd64 without `GOAMD64=v3`, FMA is not emitted; with `GOAMD64=v3`, FMA *is* emitted by `math.FMA` but still not for pattern-detected `x*y + z`. Result: a Cosine implementation that computes `dot / (sqrt(normA) * sqrt(normB))` gets one float64 on arm64 and a different float64 on amd64 unless the implementer is careful.

**Why it happens:**
"Float64 is IEEE-754, results are deterministic." This is correct *per operation* but incorrect *per expression* because compiler choices (FMA, reassociation, intrinsics) and loop order (sum reduction direction, parallel reduction) vary across architectures.

**How to avoid:**
- Per `determinism-standards.md`: "Sum a slice always left-to-right, never via parallel reduction." No goroutine-based reductions anywhere.
- For Cosine and any algorithm computing `x*y + z` patterns in a hot path: write as `(x*y) + z` with explicit parentheses, or use `float64(x*y) + z` to force intermediate rounding (per the Go issue #17895 guidance). The `algorithm-performance-reviewer` and `determinism-reviewer` agents both check.
- Avoid `math.FMA` — even where it would be faster, the cross-platform behaviour is not bit-identical.
- Avoid `math.Pow`, `math.Exp`, `math.Log` — all transcendentals are implementation-specific.
- `math.Sqrt` IS safe (Go documents it as IEEE-754 correctly rounded on all supported platforms).
- Cross-platform CI matrix verifies bit-identity of golden file output.

**Warning signs:**
- Golden file test fails on arm64 but passes on amd64 (or vice versa).
- `PropAlgorithm_NoNaN` or `PropAlgorithm_NoInf` fails on one platform but not others.
- Score for an identity test (`Score(x, x)`) is 0.9999999999999998 instead of 1.0 → likely a sqrt or sum-order issue producing a `-0` or `1 - ε` instead of 1.0.

**Phase to address:** Algorithm phases that introduce floating-point sums (Jaro, Jaro-Winkler, Strcmp95, SWG, Cosine, Monge-Elkan, the Scorer composite). The cross-platform CI matrix is needed by Phase 2 at the latest.

---

### Pitfall 10: NaN, ±Inf, and negative zero leak from division-by-zero in normalisation

**What goes wrong:**
`1 - distance/maxLen` with both inputs empty → `0/0` → NaN. `dot/(sqrt(normA)*sqrt(normB))` with one zero-vector → 0/0 → NaN. `(m/lenA + m/lenB + (m-t/2)/m)/3` in Jaro with `m=0` → 0/0 → NaN. A library that doesn't guard every normalisation path produces NaN for empty-string inputs.

NaN propagates: `NaN < threshold` is false, `NaN > threshold` is false, `NaN == NaN` is false. A Scorer comparing `NaN >= 0.8` returns false, so the warning isn't emitted. A consumer trying to debug "why didn't this match?" gets no signal.

Negative zero `-0.0` is a separate trap: `score := 1.0 - 1.0` can produce `+0.0` or `-0.0` depending on the path. Output should always be `+0.0`.

**Why it happens:**
Edge cases (both-empty, one-empty, all-same-character, all-different) are easy to forget. The implementer tests the happy path. The fuzz test catches it, but only if it's wired up.

**How to avoid:**
- Every normalisation has an explicit edge-case guard (per `determinism-standards.md` §"NaN / Inf / Negative Zero"). The standard pattern:
  ```go
  if maxLen == 0 {
      return 1.0 // both empty → 1.0 by convention, documented in godoc
  }
  return 1.0 - float64(distance)/float64(maxLen)
  ```
- Property tests `PropAlgorithm_NoNaN` and `PropAlgorithm_NoInf` for every algorithm (per requirements §13.6).
- Property test `PropAlgorithm_NoNegativeZero` — assert `score != -0.0` for all inputs (or canonicalise by `if score == 0.0 { score = 0.0 }`, which converts `-0.0` to `+0.0`).
- Edge-case unit tests: every algorithm has `_test.go` cases for both-empty, one-empty, identical, and "no shared characters" inputs with expected outputs documented per requirements §7.

**Warning signs:**
- `PropAlgorithm_NoNaN` fails for any algorithm.
- `Score("", "") != 1.0` for any algorithm (the project convention per requirements §7 is 1.0 for both-empty).
- `Score("abc", "") != 0.0` for any distance-based algorithm.
- A `Scorer.Match` returns `false` for inputs that should be similar, and the underlying score is NaN.

**Phase to address:** Every algorithm-implementation phase. The property test harness must include `PropAlgorithm_NoNaN` and `PropAlgorithm_NoInf` from the first algorithm.

---

### Pitfall 11: Sort key incomplete on scan output — ties produce platform-dependent order

**What goes wrong:**
`scan.Check` returns `[]Warning`. The warnings are sorted by `(Kind, NameA, NameB, GroupA, GroupB)` per requirements §12.4. If two warnings have all five fields equal, they are duplicates and should never both appear; but if the suppression logic has a bug, duplicates *do* appear. Their relative order is then non-deterministic (or stable only in the sense of `sort.SliceStable`, which only preserves *input* order — and the input order is itself non-deterministic if it came from a map).

**Why it happens:**
"This sort key is complete" is an *invariant*, not a fact. A bug elsewhere (duplicate emission) breaks the invariant.

**How to avoid:**
- Assert sort-key completeness inside `scan.Check` before returning — assert that the warnings list contains no duplicates (per the standard's "document why ties don't occur and assert it in a test" guidance, but tightened to an in-line assertion in this case).
- Property test: `PropScan_DeterministicAcrossRuns` runs `scan.Check` on a fixed corpus 1000 times and asserts bit-identical output (per requirements §13.6).
- Golden file `testdata/golden/scan-default.json` verifies output stability across platforms and patch versions.
- The `determinism-reviewer` agent flags any `sort.SliceStable` call as needing a completeness justification.

**Warning signs:**
- Scan golden file test fails intermittently or on one platform.
- A duplicate warning appears in the output (same NameA, NameB, Kind, group fields).

**Phase to address:** Scan sub-package phase. The completeness assertion is part of the scan implementation, not retrofittable.

---

### Pitfall 12: Algorithmic complexity attacks — Monge-Elkan, Partial Ratio, Token Set Ratio

**What goes wrong:**
- **Monge-Elkan** is O(|tokens(a)| · |tokens(b)| · cost(inner)). A user passing two strings with 1000 tokens each runs the inner metric 1,000,000 times. If the inner metric is Smith-Waterman-Gotoh (O(m·n) on long tokens), the worst case is multi-second per call.
- **Partial Ratio** is O(n·m·(n-m)) in the naïve implementation (sliding-window Levenshtein). A user passing a 10-character query against a 10,000-character target runs ~100,000,000 cell evaluations.
- **Token Set Ratio** computes the intersection, then runs three Indel ratios on potentially-long strings — three O(n·m) DPs per call.

Any of these called in a hot loop with adversarial input is a DoS vector.

**Why it happens:**
The algorithms are correct, fast on typical input, slow on adversarial input. The library does not impose input-length limits — consumers may not know the worst-case complexity.

**How to avoid:**
- Document worst-case complexity in every algorithm's godoc per requirements §7.
- The `security-reviewer` agent specifically checks for DoS vectors and recommends input-length advisories in the godoc ("Monge-Elkan is O(|A|·|B|·cost(inner)). For inputs above a few hundred tokens, prefer Token Jaccard.").
- Benchmark suite includes worst-case inputs (long strings, many tokens) — `algorithm-performance-reviewer` enforces budgets per requirements §14.1.
- A "fast rejection" optimisation: Scorer's threshold can short-circuit downstream metrics if accumulated weighted score cannot possibly reach the threshold even with perfect remaining scores. (This is a v1.x optimisation per requirements §8.6 — not a v0.1.0 blocker.)
- Document in `docs/faq.md` and `docs/performance.md` that consumers should validate input length before passing to expensive algorithms.

**Warning signs:**
- Benchmark `BenchmarkMongeElkanScore_Long` runs in seconds, not microseconds.
- Fuzz test times out (Go's fuzz harness has a per-input timeout).
- `make bench-compare` flags a regression on `BenchmarkPartialRatioScore_Long`.

**Phase to address:** Algorithm-implementation phases for Monge-Elkan, Partial Ratio, Token Set Ratio. Documentation phase for the input-length advisory in `docs/performance.md`.

---

### Pitfall 13: Full DP table allocated when two rows would suffice

**What goes wrong:**
Levenshtein, Damerau-Levenshtein OSA, Hamming, LCS, Cosine all need only the previous row to compute the next row. A naïve implementation allocates `[m+1][n+1]int`, which for two 50-character strings is 51*51 = 2601 ints = 20,808 bytes on the heap. Times millions of calls per second in a Cassandra-like consumer, you allocate gigabytes per second.

The performance budget (requirements §14.1, `performance-standards.md`) requires *zero allocations* on the ASCII fast path for Levenshtein, Damerau-Levenshtein OSA, Hamming, Jaro, Jaro-Winkler, Strcmp95. A full DP table makes this impossible.

**Why it happens:**
Wikipedia's pseudocode for Levenshtein shows a 2-D table. Implementers translate directly. The two-row optimisation requires understanding the recurrence dependency graph — implementers may not realise it's safe.

**How to avoid:**
- Per `performance-standards.md` §"Two-Row DP Optimisation": use the two-row pattern from the first commit, never the full table. Use `min(m, n)` as the row length (inner loop over the shorter string).
- For inputs ≤ 64 bytes, use stack-allocated `[65]int` arrays (Go's escape analysis keeps them on the stack — verified by `go test -bench -gcflags='-m'`).
- Heap-allocate slices only for the long-input fallback path (`> 64` bytes).
- Allocation budget enforced in benchmark: `b.ReportAllocs()` on every benchmark, target `0 allocs/op` on the ASCII fast path.
- `make bench-compare` with benchstat flags any 0 → 1 alloc/op regression as BLOCKING (per `performance-standards.md`).
- The `algorithm-performance-reviewer` agent verifies the two-row pattern at PR review.

**Warning signs:**
- `b.ReportAllocs()` on `BenchmarkLevenshteinScore_ASCII_Short` shows `> 0 allocs/op`.
- `go test -bench -gcflags='-m'` reports the DP buffer as escaping to the heap.
- Memory profile (`-memprofile`) attributes allocations to a `make([]int, ...)` call inside the hot loop.

**Phase to address:** Algorithm-implementation phases for every O(m·n) algorithm (Levenshtein, both DL variants, Hamming, Jaro, Strcmp95, SWG, LCSStr, Cosine). The benchmark + budget enforcement is part of CI from Phase 2 onward.

---

### Pitfall 14: Cached vs constructed scoring tables silently break score stability

**What goes wrong:**
Strcmp95 needs a similar-character similarity matrix (Winkler 1994). Double Metaphone needs a rule table. If these are constructed lazily in `init()` or inside the first call, two issues arise:
1. **Init order** is technically deterministic in Go but fragile — changes elsewhere can shift it.
2. **First-call latency** is non-zero, surprising callers who benchmark.
3. **Modification after init** (e.g. a test mutating the matrix) breaks score stability for the rest of the process.

**Why it happens:**
"Compute on first use" feels idiomatic in many languages. In Go's no-init() discipline (`determinism-standards.md`, `go-coding-standards.md`), tables must be `var` declarations.

**How to avoid:**
- Per requirements §7.1.7 and `determinism-standards.md`: similarity-character tables and rule tables are `var` declarations, not built in `init()`.
- Tables are unexported and not modifiable from outside the package.
- Property test: call `Strcmp95Score` repeatedly with the same input and verify identical output (covered by `PropAlgorithm_DeterministicAcrossRuns`).
- The `determinism-reviewer` agent flags any `init()` function with non-trivial work.

**Warning signs:**
- An `init()` function exists in an algorithm file → BLOCKING.
- A package-level `var` of type `map[...]...` is mutated anywhere → likely a stability bug.
- First-call benchmark is noticeably slower than subsequent calls.

**Phase to address:** Algorithm-implementation phases for Strcmp95, Double Metaphone, NYSIIS (where rule tables are used).

---

### Pitfall 15: Score change in a patch release silently breaks consumer suppression lists

**What goes wrong:**
A consumer builds a denylist keyed on `(NameA, NameB, observedScore)` to suppress known false positives. In v0.1.0, `LevenshteinScore("user_id", "user_uid") = 0.7142857142857143`. In v0.1.1, a maintainer "fixes" the normalisation to use a different formula and the score becomes `0.7857142857142857`. The denylist no longer matches the new score. Every consumer's suppression silently breaks.

**Why it happens:**
"It's just a normalisation tweak." But score values are part of the API surface — consumers depend on them being stable.

**How to avoid:**
- Score-changing edits require a minor version bump and a CHANGELOG entry (per the research-guidance pitfall #5 already in `.claude/skills/research-guidance/SKILL.md`).
- Golden file `testdata/golden/algorithms.json` captures `Score` output on a fixed corpus. Any change to a golden file is a deliberate act with PR justification (per requirements §13 / `determinism-standards.md` §"Golden Files").
- The `release-check` workflow verifies the golden file has not changed since the last patch tag.
- Documented in `docs/requirements.md` §13 (Determinism Guarantee) — consumers know that scores are stable.

**Warning signs:**
- A PR modifies `testdata/golden/algorithms.json` without a CHANGELOG entry and version bump → BLOCKING.
- Score for any reference vector changes without a documented rationale.

**Phase to address:** Foundation phase establishes the golden file infrastructure and the CI check. Every subsequent phase that adds an algorithm extends the golden corpus.

---

### Pitfall 16: GPL/LGPL contamination via reference implementations

**What goes wrong:**
An implementer reading `kompendium/go-edlib` (which is MIT) gets a clean reference for Levenshtein. The same implementer reading a GPL-licensed Java reference for Strcmp95 contaminates the implementation even if no code is copied — derivation from GPL source is itself a contamination concern.

The project is Apache-2.0 and `algorithm-licensing-standards.md` is explicit: no GPL/LGPL reference may be consulted.

**Why it happens:**
GPL/LGPL implementations exist on GitHub. They're often well-tested. The licence is two clicks away in the README — easy to miss in a hurry.

**How to avoid:**
- Every algorithm PR has a "Source origin" statement (per `algorithm-licensing-standards.md` §"Source Origin Statement") listing:
  - Primary academic source consulted
  - Any reference implementations studied, with their licences (MIT/BSD/Apache only)
  - Explicit "No GPL/LGPL references consulted"
  - Explicit "No code copied"
- The `algorithm-licensing-reviewer` agent checks the licence of every cited reference before approval.
- The file's block comment names the reference implementations studied (per the format in `algorithm-licensing-standards.md`).

**Warning signs:**
- A reference URL in a file's block comment points at a repository with `LICENSE: GPL-3.0` or `LICENSE: LGPL-*`.
- A PR description omits the Source origin statement.
- A reference implementation's licence is unclear or "see source".

**Phase to address:** Foundation phase (the licensing agent and standards are already in place). Every algorithm PR onwards.

---

### Pitfall 17: Patent gap on a newly-added algorithm

**What goes wrong:**
A maintainer post-v1.0 wants to add a new algorithm. They find it on Wikipedia, find a Go reference implementation, look at the licence ("MIT, great"), and implement. Six months later, a patent surfaces — the algorithm was patented but unenforced, and the project is now distributing patent-encumbered code.

The Metaphone 3 precedent (requirements §4, `algorithm-licensing-standards.md` §"The Metaphone 3 Precedent") explicitly excludes this — but only for that one algorithm. Every new algorithm needs the screen.

**Why it happens:**
Licence != patent. MIT-licensed code can implement patented algorithms.

**How to avoid:**
- Every new algorithm issue is labelled `needs-licence-screen` at creation.
- The `algorithm-licensing-reviewer` agent searches Google Patents / uspto.gov for the algorithm name + primary author's name before implementation begins.
- Result recorded as a comment: CLEAR / CLEAR (academic patent expired) / ENCUMBERED / UNCERTAIN.
- Implementation cannot start until the issue carries the `licence-clear` label.
- The initial 23-algorithm catalogue (requirements §7) has been screened and documented as clear (with Metaphone 3 explicitly excluded).

**Warning signs:**
- Issue lacks `licence-clear` label but a draft PR exists → BLOCKING.
- An algorithm proposal cites a primary source from a corporate author with industry affiliation (Microsoft Research, Google Research) without a patent search.
- A primary source's acknowledgments section mentions "patent application" or "patent pending".

**Phase to address:** Every algorithm-implementation phase (already-cleared algorithms for v0.1.0; new screens for any post-v1.0 additions).

---

### Pitfall 18: Local tagging accidentally cuts a release

**What goes wrong:**
A maintainer runs `git tag v0.2.0 && git push --tags` locally to mark a milestone. The release workflow triggers on `push:tags:v*`. GoReleaser runs, publishes to GitHub Releases, pushes module documentation to pkg.go.dev. The release was not reviewed, not approved, not built from a clean CI environment — but it's now public and immutable.

**Why it happens:**
`git tag` is muscle memory. Many Go libraries do release this way. The discipline that releases happen only through CI on tag push from an approved workflow is documented in CLAUDE.md and `BOOTSTRAP.md` but easily forgotten.

**How to avoid:**
- Per CLAUDE.md §"Releases — CI Only": no `git tag` on local machines, no `goreleaser release` on local machines.
- The release workflow is triggered by pushing a tag *via the CI release job*, not by local tag push.
- The `devops` agent flags any local tagging pattern in workflows or documentation as BLOCKING.
- Branch protection on `main` requires CI to pass before merge — but tag push bypasses branch protection. Workaround: the release workflow's first step verifies the tag came from a specific CI workflow (e.g. by checking that the tag was created by the `release-bot` identity, not by a personal account).
- Alternatively: tag the release on `main` only via a GitHub Release UI invocation that runs through a workflow_dispatch trigger.

**Warning signs:**
- A tag exists on `main` that was not produced by the release workflow.
- `git tag --list` on a local machine shows tags the user did not deliberately create.
- The release workflow runs but produces a stale `bench.txt` (because the local tag predated the bench update).

**Phase to address:** Foundation phase (release workflow, branch protection, devops agent already in place). Documented in `BOOTSTRAP.md`.

---

### Pitfall 19: Benchmark regressions sneak in without benchstat statistical validity

**What goes wrong:**
A PR adds a refactor. CI's benchmark step runs each benchmark once and compares to `bench.txt`. The new number is 8% slower — but the noise floor on a single run is ±15%. The PR merges. Over six PRs, the algorithm is 60% slower. By the time anyone notices, the regression is across many commits.

`benchstat` exists specifically for this — it runs the same benchmark `N` times (typically 10) and compares distributions with a t-test at `p < 0.05`. Without `benchstat`, single-run comparisons are statistically meaningless.

**Why it happens:**
Running benchmarks 10 times in CI is slow. Maintainers feel pressure to keep CI under 5 minutes. The shortcut is one run.

**How to avoid:**
- Per `performance-standards.md` §"Benchstat Regression Detection": CI runs benchmarks 10 times via `go test -bench -count=10`, saves to `new.txt`, runs `benchstat bench.txt new.txt`.
- If any benchmark shows > 10% regression at `p < 0.05`, the build fails with the `benchstat` output.
- The benchmark CI job runs on a labelled self-hosted runner (shared with `axonops/mask` and `axonops/audit` for hardware consistency). If unavailable, falls back to `ubuntu-latest` informationally with a CI annotation explaining the regression detection is skipped.
- `bench.txt` is committed to the repository and updated on every tagged release.
- The `algorithm-performance-reviewer` agent flags missing benchmark coverage on new algorithms.

**Warning signs:**
- A PR claims "no performance impact" with a single-run benchmark.
- `bench.txt` has not been updated since the last tagged release but new algorithms have been added.
- `make bench-compare` shows "n=1" — benchmarks were run only once.

**Phase to address:** Foundation phase (CI infrastructure, self-hosted runner setup, `bench.txt` baseline). Subsequent algorithm phases inherit the regression detection.

---

### Pitfall 20: Coverage thresholds reported but not actually failing the build

**What goes wrong:**
The README displays a 95% coverage badge. The CI workflow runs `go test -coverprofile=cov.out`. But CI doesn't *enforce* the 95% threshold — it just generates the badge. A PR drops coverage to 85% because the maintainer forgot to add tests for a new edge case. The badge updates to "85%" but CI is green. Six PRs later, coverage is at 70%.

The coverage targets (≥ 95% overall, ≥ 90% per file, 100% on public API per `go-testing-standards.md`) are aspirations only if not enforced.

**Why it happens:**
"Coverage badge" is wired to a third-party service (codecov, coveralls) that *reports* coverage but doesn't fail the CI step. The CI step that generates `cov.out` exits 0 regardless.

**How to avoid:**
- A `make coverage-check` target that runs `go tool cover -func=cov.out`, parses the total, and exits non-zero if below threshold.
- CI workflow step: `make coverage-check` (not `make coverage` alone).
- Per-file enforcement: a script that parses `cov.out` line-by-line and fails if any non-generated file is below 90%.
- Public API enforcement: a script that uses `go doc` to list all exported symbols and verifies each is referenced from at least one test (or run with `coverage.out` and require 100% on the exported names).
- The `go-quality` agent verifies coverage enforcement is wired in.

**Warning signs:**
- README shows a coverage badge but no `make coverage-check` target exists.
- A PR drops coverage but CI is green.
- `go tool cover -func=cov.out | grep -v '100%'` returns non-trivial output on the public API.

**Phase to address:** Foundation phase (CI workflows, Makefile targets, coverage enforcement). Every subsequent phase increases the corpus that contributes to coverage.

---

## Technical Debt Patterns

Shortcuts that seem reasonable but create long-term problems.

| Shortcut | Immediate Benefit | Long-term Cost | When Acceptable |
|----------|-------------------|----------------|-----------------|
| Implement only OSA, name it "Damerau-Levenshtein" | One less algorithm to ship | Consumers cite mismatching values vs. literature; renaming later is a breaking change | Never — distinct names from day one |
| Skip the boost-threshold gate in Jaro-Winkler | 5 fewer lines of code | Scores diverge from canonical Winkler reference vectors; suppresion lists drift | Never |
| Use `init()` to build the Strcmp95 similarity table | "Lazy"-feels-idiomatic | Init order fragility; first-call latency spike; mutation risk | Never per project standards |
| Use `map[string]int` directly in q-gram output | Faster initial implementation | Map iteration leaks into output; non-determinism across runs | Internal use only; output must sort |
| `sort.SliceStable` without verifying complete sort key | "Stable sort handles ties" | Ties get input-order dependence; input may itself be non-deterministic | Only with assertion that ties don't occur |
| Full m×n DP table instead of two-row | Easier to translate from Wikipedia | Allocates ~20KB per call; 0 → many allocs/op on ASCII fast path | Only for algorithms inherently needing the full matrix (Lowrance-Wagner full DL, SWG) |
| Skip the patent screen on a "well-known" algorithm | Faster issue triage | Patent surfaces later; algorithm must be removed | Never — every new algorithm screened |
| Cross-validate against `fuzzywuzzy` rather than RapidFuzz | Familiar library | fuzzywuzzy's pure-Python and C paths disagree; cross-validation is ambiguous | Never — use RapidFuzz |
| Benchmark once and compare to `bench.txt` | Faster CI | Noise floor swamps small regressions; cumulative drift | Only as informational signal, not a gate |
| Skip ASCII fast path "for simplicity" | Smaller code | 10× slowdown on the common case | Only for algorithms that must use runes (phonetics on ASCII letters only is OK) |

## Integration Gotchas

Common mistakes when integrating fuzzymatch into a consumer.

| Integration | Common Mistake | Correct Approach |
|-------------|----------------|------------------|
| Consumer caches `Scorer` and mutates it | Tries to add/remove algorithms after construction → race | Scorer is immutable after `NewScorer`; build a new one per configuration |
| Consumer compares scores with `==` | Floating-point equality breaks on edge cases | Use `math.Abs(a - b) < epsilon` for "near equal" tests; for golden tests use byte-identical comparison via JSON serialisation |
| Consumer calls `ScoreAll` then sorts the map by key | `ScoreAll` returns a map — iteration order is non-deterministic | Sort keys explicitly, or use `Scorer.Algorithms()` (returns AlgoID-ordered slice) and look up |
| Consumer passes `[]byte` containing multibyte UTF-8 to byte-level algorithm | Treats multibyte codepoints as multiple characters | Use the `*Runes` variant for Unicode input |
| Consumer builds suppression denylist on `(NameA, NameB, score)` triple | Score changes break the denylist on patch release | Suppress on `(NameA, NameB)` pair only; scores are stable but the pair is the identity |
| Consumer expects `Soundex("Smith") == "S530"` | Uses SQL-variant Soundex from another tool | Library uses Knuth/Census Soundex; document the variant explicitly |
| Consumer expects token-sort ratio to match Python `difflib.ratio()` | difflib uses Ratcliff-Obershelp; library uses Indel | Use `RatcliffObershelpScore` if difflib semantics are needed |

## Performance Traps

Patterns that work at small scale but fail as usage grows.

| Trap | Symptoms | Prevention | When It Breaks |
|------|----------|------------|----------------|
| Full DP table allocation | `> 0 allocs/op` on `BenchmarkLevenshteinScore_ASCII_Short`; GC pressure under load | Two-row DP from day one; stack-allocated arrays for short input | Immediately on any caller running >100k comparisons/sec |
| Cosine similarity with `math.Pow` for norms | Slow + non-deterministic across architectures | `math.Sqrt` of sum-of-squares; no `Pow` | Immediately on cross-arch CI; gradually on hot-path callers |
| Monge-Elkan on long token lists | O(n·m·cost(inner)) explodes | Document complexity in godoc; recommend Token Jaccard for >100-token inputs | Inputs >~500 tokens per side |
| Partial Ratio naive sliding window | O(n·m·(n-m)) | Sliding-window optimisation (v1.x); v1.0 documents complexity | Long-target inputs (>1000 chars) |
| Scorer composite without early termination | All 6 algorithms run even when score cannot reach threshold | Threshold short-circuit when remaining-weight scores cannot achieve threshold (v1.x optimisation) | Hot-path callers with high threshold (e.g. 0.9) |
| Q-gram extraction allocating per-call | Allocates `map[string]int` per comparison | `sync.Pool` for q-gram maps (only if profile shows it's a bottleneck — `sync.Pool` complicates testing); or accept the 1–2 alloc budget per `performance-standards.md` | High-frequency callers with short inputs |
| Scan sub-package within-group then cross-group passes | Quadratic when groups are large | Token-bucket optimisation (requirements §12.5) verified equivalent to naive O(N²) | Collections >1000 items |
| `unicode.ToLower` on ASCII bytes | ~10× slower than `c |= 0x20` for ASCII | ASCII fast path via bitwise OR (`performance-standards.md` §"Lowercasing ASCII") | Normalisation hot path |
| Compiling regex per call in phonetic algorithms | Regex compile dominates the call cost | Pre-compile as `var` declaration (no `init()`) per `performance-standards.md` §"Pre-Compilation" | Any high-frequency caller |
| Goroutine-based parallel scan | Non-determinism via map iteration in goroutines | No goroutines in library (per `go-coding-standards.md` §"Concurrency") | Always — concurrency lives in consumers |

## Security Mistakes

Domain-specific security issues beyond general web security.

| Mistake | Risk | Prevention |
|---------|------|------------|
| Panic on malformed UTF-8 | DoS via user-supplied input | Fuzz every algorithm with malformed UTF-8 corpus; `security-reviewer` agent enforces panic safety |
| Unbounded input length | DoS via 100MB strings — `O(m·n)` becomes catastrophic | Document complexity in godoc; recommend consumer-side length validation; per `docs/performance.md` |
| Adversarial Monge-Elkan input | DoS via 10,000-token strings → 10^8 inner-metric calls | Document worst-case complexity; recommend Token Jaccard for high-cardinality |
| `unsafe` to elide bounds checks | Memory safety regression for marginal perf | Forbidden per `performance-standards.md` §"When Optimisation is NOT Worth It" |
| SIMD intrinsics (`asm` files) | Determinism regression across architectures | Forbidden without determinism-reviewer sign-off; benchmarks must show clear win |
| Regex with user-supplied pattern | ReDoS in regex-based phonetic encoder | No `regexp` in algorithm implementations; phonetic rules are explicit conditionals |
| Logging the input strings | Information disclosure in audit logs | Library has no logging; consumers responsible for log hygiene |
| Pattern caching with user-provided cache key | Denial via cache pollution | No caching in library (pure functions); consumers responsible |

## UX Pitfalls

Common user experience mistakes in this domain.

| Pitfall | User Impact | Better Approach |
|---------|-------------|-----------------|
| Returning raw integer distance without normalisation | User can't compare distances across algorithms (Levenshtein distance 3 vs Jaro 0.8 — apples and oranges) | Always expose `XxxScore` in [0.0, 1.0]; expose `XxxDistance` separately for consumers who genuinely want it |
| Embedding threshold inside metric | Caller can't compute the underlying score | Threshold lives on the Scorer (`Match` returns bool, `Score` returns the raw score) |
| Single `Compare(a, b, algorithm string)` function | Stringly-typed — typos in algorithm name fail at runtime, not compile time | Typed `AlgoID` enum per `go-coding-standards.md` §"API Design" |
| Functions returning `int` distance with no error for length mismatch in Hamming | Silent wrong answer when caller passes unequal-length strings | `HammingDistance` returns `(int, error)` with `ErrHammingLengthMismatch`; `HammingScore` documents 0.0 on mismatch |
| Configuration via struct literal | Adding a field is a breaking change | Functional options pattern per `go-coding-standards.md` §"API Design" |
| `ScoreAll` returning a map keyed on `AlgoID.String()` | Stringly-typed lookup; keys vary across versions | Keyed on `AlgoID` directly; or returns `[]ScoreEntry{Algo AlgoID, Score float64}` sorted by AlgoID |
| Default Scorer's algorithm composition undocumented | Consumers can't reason about what `DefaultScorer()` actually computes | Document the 6 default algorithms and weights in godoc (per requirements §8.5) |
| Errors that don't compose with `errors.Is` / `errors.As` | Consumers can't pattern-match on error types | Sentinel errors at package level; `errors.Is` only per `go-coding-standards.md` §"Errors" |
| Different normalisation rules for similar algorithms | Inconsistent scores (Levenshtein normalises by max length, Cosine by formula — but Hamming by `len(a)` not `max`) | Document each algorithm's normalisation rule in godoc and in `docs/algorithms.md`; `algorithm-correctness-reviewer` verifies consistency |

## "Looks Done But Isn't" Checklist

Things that appear complete but are missing critical pieces.

- [ ] **Algorithm implementation:** Has the primary source citation block, but the formula in the comment doesn't match the code → verify formula matches recurrence; `algorithm-correctness-reviewer` enforces.
- [ ] **Algorithm implementation:** Unit tests pass with reference vectors, but property tests for symmetry / identity / range are missing → verify all four invariants present; `algorithm-correctness-reviewer` enforces.
- [ ] **Algorithm implementation:** Property tests exist but use a small `testing/quick` corpus → verify `MaxCount` is at least 1000; default 100 is too low.
- [ ] **Algorithm implementation:** Fuzz test exists but corpus directory is empty → verify `testdata/fuzz/Fuzz<Algorithm>/` contains malformed UTF-8 corpus.
- [ ] **Algorithm implementation:** Has byte-level variant but no rune-level variant for an algorithm that needs Unicode → check requirements §7 per-algorithm spec; verify `XxxScoreRunes` is exposed if applicable.
- [ ] **Algorithm implementation:** Has `XxxScoreRunes` but it's an alias for `XxxScore` → verify `[]rune` conversion is actually used; test on multi-byte input gives different result from byte-level.
- [ ] **Phonetic algorithm:** `XxxCode` returns the encoded form, `XxxScore` returns boolean → verify both are exposed; some implementations skip the `Code` function.
- [ ] **Scorer:** Accepts algorithm list but does not reject invalid algorithm IDs at construction → verify `NewScorer` returns `ErrInvalidAlgoID` for unknown IDs.
- [ ] **Scorer:** Sums weights but does not validate non-negative → verify `NewScorer` returns `ErrInvalidWeight` for any weight ≤ 0.
- [ ] **Scorer:** Returns score but does not validate threshold in [0,1] → verify `NewScorer` returns `ErrInvalidThreshold`.
- [ ] **Scorer:** Methods present but `Scorer` is mutable → verify all fields unexported; no setter methods; safe for concurrent use.
- [ ] **Scan:** Returns warnings but order is not deterministic → verify sort key is `(Kind, NameA, NameB, GroupA, GroupB)` and is complete.
- [ ] **Scan:** Suppression list applied but emitting duplicates → verify in-line assertion that no duplicates remain in output.
- [ ] **CI:** Benchmark runs once → verify `-count=10` and `benchstat` against `bench.txt`.
- [ ] **CI:** Cross-platform matrix exists but golden file test not in matrix → verify `make verify-determinism` runs on all of linux/amd64, linux/arm64, darwin/arm64, darwin/amd64, windows/amd64.
- [ ] **CI:** Coverage badge displays but threshold not enforced → verify `make coverage-check` fails on <95% overall.
- [ ] **Release:** Workflow exists but local tagging not blocked → verify release workflow checks tag source; `devops` agent enforces no local tagging in docs/workflows.
- [ ] **Release:** `bench.txt` not updated since last tag → block release until `update-bench-txt.sh` re-runs.
- [ ] **Documentation:** README has algorithm list but no per-algorithm citations → verify `docs/algorithms.md` has primary source + reference vectors per algorithm.
- [ ] **Documentation:** `llms.txt` exists but doesn't link to canonical examples → verify `llms-full.txt` includes godoc-extracted symbol definitions.
- [ ] **Patent screen:** Issue has `licence-clear` label but no comment from `algorithm-licensing-reviewer` → verify the screen result is recorded as a comment.

## Recovery Strategies

When pitfalls occur despite prevention, how to recover.

| Pitfall | Recovery Cost | Recovery Steps |
|---------|---------------|----------------|
| OSA/Full DL mislabeled | LOW (pre-v1.0) → HIGH (post-v1.0, requires major version bump) | Rename in v0.x; deprecate alias + dual-name for v1.x compat |
| Jaro-Winkler constant wrong | MEDIUM | Fix constant; bump minor version; CHANGELOG; update golden files with rationale; notify consumers via release notes |
| Gotoh implementation bug | HIGH | Fix recurrence; bump minor version; update reference vectors; regenerate golden files; CHANGELOG entry explains the corrected formulation |
| Soundex wrong variant | LOW (if caught pre-v1.0) → MEDIUM (post-v1.0) | Document the implemented variant; if changing the variant, add a v2 function and deprecate the v1 (don't change v1's behaviour) |
| Map iteration leaks into output | LOW | Add explicit sort; regenerate golden file; add property test; backport to last patch release |
| Float reduction order varies | MEDIUM | Pin reduction order in code (`(a+b)+c`, not `a+(b+c)`); add property test; regenerate golden files; cross-platform CI rerun |
| NaN/Inf leakage | LOW | Add explicit edge-case guards; add property test; regenerate golden files |
| Score change in patch release | HIGH (breaks consumers' suppression lists) | Yank the patch release; ship a `vX.Y.Z+1` reverting the score change; document in CHANGELOG; rebuild as a minor bump if change is intentional |
| GPL contamination in a reference | HIGH (rewrite required) | Identify which functions were derived; rewrite from primary source with different reviewer; document the rewrite in CHANGELOG; legal review |
| Patent surface on shipped algorithm | CRITICAL | Remove algorithm; major version bump; deprecation notice; offer alternative algorithm |
| Local-tag accidental release | MEDIUM | Yank the release on pkg.go.dev (if possible); push a patch tag from CI; document in CHANGELOG |
| Coverage regression | LOW | Add tests for missing branches; enforce coverage gate in CI |
| Benchmark regression | LOW–MEDIUM | Revert PR; investigate with `pprof`; fix; re-benchmark with `benchstat` 10-run |

## Pitfall-to-Phase Mapping

How roadmap phases should address these pitfalls. Phase names below are illustrative — the actual roadmap may restructure per CLAUDE.md "§19 phasing is the default but the roadmapper may restructure".

| # | Pitfall | Prevention Phase | Verification |
|---|---------|------------------|--------------|
| 1 | OSA vs Full DL mislabel | Algorithm phase: Damerau-Levenshtein | Reference vector `"ca"`/`"abc"` divergence test; distinct AlgoIDs |
| 2 | Jaro-Winkler constants | Algorithm phase: Jaro family | Three reference vectors covering boost-gate; property tests on bounds |
| 3 | Gotoh erratum | Algorithm phase: SWG (flag for extra scrutiny) | Cross-validation against EMBOSS / biopython for reference vectors; identity + symmetry property tests |
| 4 | Soundex variant | Algorithm phase: Soundex | `"Tymczak" = "T522"` reference vector; variant name in file's block comment |
| 5 | NYSIIS / Double Metaphone branches | Algorithm phase: Phonetic | Multi-language reference vectors; `algorithm-correctness-reviewer` rule-table coverage check |
| 6 | Indel vs Ratcliff-Obershelp ambiguity | Algorithm phase: Token-based | Reference vectors cross-validated against RapidFuzz; godoc cites Indel formula explicitly |
| 7 | Byte vs rune indexing | Foundation phase (test harness) + every algorithm phase | Fuzz corpus with malformed UTF-8; multi-byte test cases per algorithm |
| 8 | Map iteration → output | Foundation phase (CI matrix, property test harness) | `PropAlgorithm_DeterministicAcrossRuns`; golden file CI on 5 platforms |
| 9 | Float reduction order | Foundation phase (golden infra) + algorithms using float sums (Jaro, Cosine, SWG, Scorer) | Cross-platform CI matrix; pin FMA-free expression formulation |
| 10 | NaN/Inf/−0 leakage | Foundation phase (property test harness) | `PropAlgorithm_NoNaN`, `PropAlgorithm_NoInf`, `PropAlgorithm_NoNegativeZero` |
| 11 | Scan sort key incomplete | Scan sub-package phase | In-line assertion; property test; golden file |
| 12 | Algorithmic complexity attacks | Algorithm phases for Monge-Elkan, Partial Ratio, Token Set Ratio | Benchmark on long inputs; security-reviewer flag; godoc complexity statement |
| 13 | Full DP table allocation | Every algorithm phase with O(m·n) algorithms | `b.ReportAllocs()` shows `0 allocs/op` on ASCII fast path; benchstat detects regressions |
| 14 | Cached tables in `init()` | Algorithm phases for Strcmp95, Double Metaphone, NYSIIS | `var` declaration only; determinism-reviewer flags `init()` |
| 15 | Score change in patch release | Foundation phase (golden file infra) + every release | Golden file diff requires CHANGELOG + version bump; CI verifies on `release-check` |
| 16 | GPL/LGPL contamination | Every algorithm phase | "Source origin" statement in PR; algorithm-licensing-reviewer signs off |
| 17 | Patent gap on new algorithm | Every new-algorithm issue post-v0.1 | `needs-licence-screen` label workflow; `licence-clear` gate before implementation |
| 18 | Local tagging release accident | Foundation phase (CI/release infra) | Release workflow verifies tag source; devops agent enforces |
| 19 | Benchmark regressions without benchstat | Foundation phase (CI infra, self-hosted runner) | `make bench-compare` with `-count=10`; > 10% at p<0.05 fails build |
| 20 | Coverage thresholds not enforced | Foundation phase (CI infra) | `make coverage-check` exits non-zero below 95% overall, 90% per-file |

**Phase ordering implication:** Pitfalls 7, 8, 9, 10, 13, 15, 18, 19, 20 are *test infrastructure / CI* pitfalls that must be addressed in the foundation phase, *before* any algorithm is implemented. Without that infrastructure, every subsequent algorithm phase carries hidden risk.

Pitfalls 1, 2, 3, 4, 5, 6, 14 are *per-algorithm correctness* pitfalls handled at the algorithm's implementation phase by the `algorithm-correctness-reviewer` agent.

Pitfalls 11, 12 are *integration* pitfalls handled at the Scorer and scan sub-package phases.

Pitfalls 16, 17 are *governance* pitfalls — the licensing/patent agents and the issue-writer agent enforce them at issue triage, before implementation.

---

## Sources

- `docs/requirements.md` §4 (Out of Scope: Metaphone 3), §7 (per-algorithm specifications including OSA/Full DL divergence, Jaro-Winkler constants, Soundex variant, Double Metaphone reference vectors), §11 (Phonetic Algorithm Integration), §13 (Determinism Guarantees), §14 (Performance Budgets), §15 (Testing Strategy), §17 (CI/CD Requirements)
- `docs/prior-art-research.md` (Go ecosystem survey identifying fuzzywuzzy/RapidFuzz Indel-vs-difflib trap, Wikipedia-derived implementation pitfalls)
- `.claude/skills/algorithm-correctness-standards/SKILL.md` (primary-source citation, mathematical invariants, edge cases, byte-vs-rune)
- `.claude/skills/algorithm-licensing-standards/SKILL.md` (patent screen, GPL/LGPL avoidance, source-origin statement, Metaphone 3 precedent)
- `.claude/skills/determinism-standards/SKILL.md` (no-map-iteration rule, float stability, NaN/Inf/-0, golden files, cross-platform matrix)
- `.claude/skills/performance-standards/SKILL.md` (two-row DP, ASCII fast path, stack allocation, benchstat, allocation budgets)
- `.claude/skills/go-coding-standards/SKILL.md` (no `init()`, sentinel errors, AlgoID enum)
- `.claude/skills/research-guidance/SKILL.md` (existing pitfall numbering — extended by this research)
- Damerau-Levenshtein Wikipedia + RDocumentation `OSA` reference — confirms OSA vs Full divergence on `"ca"`/`"abc"`
- Apache Commons Text `JaroWinklerDistance.java` — canonical `PREFIX_LENGTH_LIMIT = 4` constant
- biorxiv "Are all global alignment algorithms and implementations correct?" (Flouri et al.) — Gotoh 1982 erratum, prevalence in textbooks
- RapidFuzz `api_differences.md` — Indel vs difflib divergence between fuzzywuzzy implementations
- Go issue #17895 (FMA reproducibility), #71204 (GOAMD64=v3 FMA detection) — float cross-platform determinism
- Soundex Wikipedia / Rosetta Code — Census-vs-SQL-vs-Russell variant pitfalls (`"Tymczak"` divergence)

---
*Pitfalls research for: Go string-similarity / fuzzy-matching library (fuzzymatch)*
*Researched: 2026-05-13*
