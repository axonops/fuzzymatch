---
status: issues_found
agent: code-reviewer
scope: entire codebase (phases 1-8)
reviewed: 2026-05-17T06:34:34Z
finding_counts:
  critical: 3
  important: 18
  improvement: 24
  total: 45
---

# Comprehensive code review — fuzzymatch (Phases 1–8)

Findings are organised by severity; severity tags are organisational only. Every refactoring opportunity is surfaced regardless of perceived priority.

---

## CRITICAL — bugs that produce wrong results, panic paths, or contract violations

### [Critical] `HammingDistance` signature deviates from spec — silent wrong answer on length mismatch
- **File:** `/Users/johnny/Development/fuzzymatch/hamming.go:69-90`
- **Phase introduced:** Phase 2
- **Issue:** `HammingDistance(a, b string) int` (no error return). On unequal-length inputs it silently returns `max(len(a), len(b))`. `docs/requirements.md` §7 line 362 specifies the signature `HammingDistance(a, b string) (int, error)` returning `ErrHammingLengthMismatch`. `.claude/skills/algorithm-correctness-standards/SKILL.md` line 104 reiterates this. `.planning/research/PITFALLS.md` line 610 documents the rule: "Functions returning `int` distance with no error for length mismatch in Hamming → Silent wrong answer when caller passes unequal-length strings → `HammingDistance` returns `(int, error)` with `ErrHammingLengthMismatch`".
- **Standard:** algorithm-correctness-standards §"Edge cases"; docs/requirements.md §7
- **Action:** Discuss-phase needed — either (a) update spec and skills to reflect "silent return max(len)" decision and document the deviation, or (b) introduce `ErrHammingLengthMismatch` in errors.go and break the signature to `(int, error)`. Option (b) is a breaking API change.
- **Rationale:** Three documents (requirements.md, skill, pitfalls) agree that the current implementation is wrong. Either the spec is wrong or the code is. Until reconciled, downstream consumers may silently misuse Hamming on unequal inputs.

### [Critical] `MongeElkanScore` asymmetry property test has a documented failure in `bench.txt.new`
- **File:** `/Users/johnny/Development/fuzzymatch/bench.txt.new` (test failure record)
- **Phase introduced:** Phase 6 (Monge-Elkan)
- **Issue:** A committed `bench.txt.new` file at the repo root captures `--- FAIL: TestProp_MongeElkanScore_AsymmetricWhenTokenCountAsymmetric` with a multi-rune random input. Either (a) the test premise (`strings.Fields` token count proxy) is buggy and produces false positives because `Tokenise`'s separator set does not include Unicode whitespace like ` ` (figure space, U+2007), or (b) Monge-Elkan really produces equal scores on inputs where `strings.Fields` reports different token counts but `Tokenise` reports the same — which would be a property-test correctness gap.
- **Standard:** go-testing-standards §"Property tests"
- **Action:** Investigate, then either fix the test premise to gate on `len(Tokenise(...))` instead of `len(strings.Fields(...))` OR delete the leftover `bench.txt.new` file with an issue-linked CHANGELOG entry. Do NOT leave a failing test record committed to the working tree.
- **Rationale:** A persistent test-failure artefact in the repo undermines confidence and creates "is this a known issue?" ambiguity for future reviewers.

### [Critical] `WithTverskyAlgorithm` does not enforce α+β > 0 — runtime panic escapes Scorer
- **File:** `/Users/johnny/Development/fuzzymatch/scorer_options.go:381-399`
- **Phase introduced:** Phase 8
- **Issue:** The option validates `alpha < 0`, `beta < 0`, `weight <= 0`, `n < 1`, but does NOT check `(alpha == 0 && beta == 0)`. A Scorer constructed with `WithTverskyAlgorithm(weight=1.0, alpha=0, beta=0, n=3)` succeeds at NewScorer time but then panics inside `TverskyScore` (tversky.go:241) when `Score` is first called. The godoc (lines 377-380) claims "α + β > 0 constraint is enforced at runtime by TverskyScore itself; this option does not re-check it" — but the project's documented invariant is that the Scorer option layer returns `ErrInvalidTverskyParam` (typed error), reserving panic for direct calls only (CONTEXT.md §5 LOCKED). The current code violates that invariant.
- **Standard:** go-coding-standards §"API Design" ("No `log.Fatal`, `os.Exit`, or `panic` that escapes the package boundary"); errors.go line 79-88 sentinel godoc.
- **Action:** Code fix — add `if alpha == 0 && beta == 0 { return ErrInvalidTverskyParam }` to `WithTverskyAlgorithm` (and document the Scorer-vs-direct-call divergence in WithTverskyAlgorithm's godoc).
- **Suggested fix:**
  ```go
  if alpha < 0 || beta < 0 {
      return ErrInvalidTverskyParam
  }
  if alpha == 0 && beta == 0 {
      return ErrInvalidTverskyParam
  }
  ```
- **Rationale:** Scorer construction succeeding for a configuration that subsequently panics is exactly the failure mode the typed-error sentinel was designed to prevent.

---

## IMPORTANT — standards violations, missing tests, refactoring needs

### [Important] `WithMongeElkanAlgorithm` only rejects `inner == AlgoMongeElkan`, not the broader allowlist — runtime panic escapes Scorer
- **File:** `/Users/johnny/Development/fuzzymatch/scorer_options.go:425-446`
- **Phase introduced:** Phase 8
- **Issue:** The option validates `dispatch[inner] != nil` and `inner != AlgoMongeElkan`, but does NOT check membership in `permittedMongeElkanInner` (the 18-entry allowlist in monge_elkan.go:291-317). A consumer calling `WithMongeElkanAlgorithm(weight=1.0, inner=AlgoTokenSortRatio)` constructs the Scorer successfully but the underlying `MongeElkanScoreSymmetric` panics at Score-time with "AlgoID TokenSortRatio not permitted as Monge-Elkan inner metric". The godoc acknowledges this (lines 417-419) — "Passing an inner AlgoID that the underlying ME implementation rejects will panic at Score time (programmer error); the panic surfaces via godog's recover mechanism in plan 08-04's BDD scenarios" — but this contradicts the documented design: Scorer option errors should be typed, not panics.
- **Standard:** go-coding-standards §"API Design"; CONTEXT.md §3 LOCKED for Phase 6.
- **Action:** Code fix — check `!permittedMongeElkanInner[inner]` in the option layer and return `ErrInvalidAlgorithm`.
- **Rationale:** Same as the Tversky finding above — Scorer construction should never produce a *Scorer that panics on first use.

### [Important] `damerau_full.go` heap-allocates a full O(m·n) DP table — large alloc on long inputs
- **File:** `/Users/johnny/Development/fuzzymatch/damerau_full.go:226-321`
- **Phase introduced:** Phase 2
- **Issue:** `damerauFullDP` allocates `make([]int, (m+2)*(n+2))` regardless of input length. For two 100-char inputs that's 102×102 = 10404 ints = ~83KB. For 10,000-char inputs that's ~800MB. No upper-bound guard; the algorithm trusts the caller. The file header (lines 70-72) documents this as a v1.x perf follow-up. There is no `#TBD` issue link.
- **Standard:** performance-standards §"Per-Algorithm Budgets"; CLAUDE.md "GitHub Issues Are the Source of Truth"
- **Action:** File a GitHub issue tracking the two-row + auxiliary-anchor-table optimisation; reference the issue number in the file header.
- **Rationale:** Untracked performance regression risk on pathological long inputs; the threat-model says "callers control input size" but a 100MB+ allocation from a single string-similarity call is surprising for a library promising "deterministic, production-grade" behaviour.

### [Important] `roMatchedLength` (Ratcliff-Obershelp) uses unbounded recursion on user input
- **File:** `/Users/johnny/Development/fuzzymatch/ratcliff_obershelp.go:200-210`
- **Phase introduced:** Phase 4
- **Issue:** `roMatchedLength` recurses into left+right segments without depth-limiting. For pathological inputs (e.g. inputs with many short common substrings) recursion depth can reach O(min(la, lb)). Go default goroutine stack grows dynamically, but an attacker-controlled input of e.g. 100,000 bytes could trigger 100,000 stack frames — uses ~6.4MB stack at 64 bytes/frame. The file header (lines 79-84) acknowledges the recursion contract but doesn't bound depth.
- **Standard:** performance-standards §"DoS notice"; security-reviewer scope per CLAUDE.md
- **Action:** Either (a) document the maximum input length the recursion is safe for, (b) convert to an iterative implementation with an explicit work-queue, or (c) add a depth-limit gate that falls back to a clamped score on overflow.
- **Rationale:** The threat model in CLAUDE.md flags "DoS via pathological inputs" as a security-reviewer concern; this recursion is the most prominent unguarded path.

### [Important] `partial_ratio.go` TODO without GitHub issue reference
- **File:** `/Users/johnny/Development/fuzzymatch/partial_ratio.go:148-154`
- **Phase introduced:** Phase 6
- **Issue:** `TODO(#TBD): implement sliding-window DP per Bachmann RapidFuzz docs ... A future GitHub issue will track the sliding-window DP implementation; this TODO will be updated with the issue number once it is created.`
- **Standard:** CLAUDE.md "Every `TODO` must reference a GitHub issue: `// TODO(#42): add Ukkonen banding optimisation`"
- **Action:** Open a GitHub issue and replace `#TBD` with the issue number.
- **Rationale:** Project rule.

### [Important] Inconsistent dispatch slot comments — three are wrong, one is missing
- **File:** `/Users/johnny/Development/fuzzymatch/dispatch_swg.go:17` ("slot 6", should be 7); `/Users/johnny/Development/fuzzymatch/dispatch_soundex.go:17` ("slot 23", actual iota 18); `/Users/johnny/Development/fuzzymatch/dispatch_double_metaphone.go:17` ("slot 24", actual iota 19); `/Users/johnny/Development/fuzzymatch/dispatch_mra.go:15-16` ("slot 26", actual iota 21); `/Users/johnny/Development/fuzzymatch/dispatch_nysiis.go:15` ("slot ... (25 — see algoid.go)", actual iota 20)
- **Phase introduced:** Phases 2, 3, 5, 7
- **Issue:** Multiple dispatch files cite incorrect iota slot numbers. AlgoSoundex is iota 18 (Phase 7 added 4 phonetic entries AFTER position 17 token_jaccard); dispatch_soundex.go claims slot 23. Similarly for double_metaphone (19), NYSIIS (20), MRA (21). dispatch_swg.go says slot 6 but Strcmp95 is at 6 and SmithWatermanGotoh at 7.
- **Standard:** documentation-standards §"godoc accuracy"
- **Action:** Either (a) drop the explicit slot numbers from the comments — they duplicate algoid.go and drift — or (b) audit and correct all of them. Option (a) is preferable.
- **Rationale:** The comments are load-bearing for future reviewers; incorrect numbers undermine trust.

### [Important] `WriteGoldenFile` is a public exported function but exists only for test maintenance
- **File:** `/Users/johnny/Development/fuzzymatch/golden_canonical.go:88-100`
- **Phase introduced:** Phase 1
- **Issue:** `WriteGoldenFile(path string, v any) error` is exported from production code but its godoc says "It is intended for test maintenance only — production code never invokes it." This pollutes the public API surface — consumers see this function on pkg.go.dev and may use it for unintended purposes. The file `golden_canonical.go` is in the production package (not `_test.go`).
- **Standard:** go-coding-standards §"API Design"; api-ergonomics-reviewer scope.
- **Action:** Move `WriteGoldenFile` (and possibly `canonicalMarshal` if it has no production callers) into a `_test.go` file (e.g. `golden_canonical_test.go` or a new test helper). Use `export_test.go` to re-export internal symbols only as needed.
- **Rationale:** Public API hygiene; the `-update` workflow can live in tests.

### [Important] `Tokenise` allocates per-token byte buffer even for pure-ASCII tokens
- **File:** `/Users/johnny/Development/fuzzymatch/tokenise.go:323-338`
- **Phase introduced:** Phase 1
- **Issue:** `appendToken` allocates `make([]byte, 0, len(rs)*utf8.UTFMax)` — 4× over-provisioning for ASCII tokens — then `string(buf)` causes another allocation. Two allocations per token even for ASCII identifiers. The comment acknowledges "profiling can revisit if benchmarks show this is a bottleneck" but does not link an issue.
- **Standard:** performance-standards §"ASCII Fast Path Pattern"
- **Action:** Add an ASCII fast path — when all runes in `rs` are < 0x80, allocate `make([]byte, len(rs))` exactly and use bitwise-OR lowercasing inline.
- **Rationale:** Hot-path performance for token-based algorithms (Monge-Elkan, Token Sort/Set Ratio, Partial Ratio, Token Jaccard) — each one calls Tokenise twice.

### [Important] `extractQGrams` map keys are sub-slices that retain backing storage of the input
- **File:** `/Users/johnny/Development/fuzzymatch/q_gram.go:104-117`
- **Phase introduced:** Phase 5
- **Issue:** `m[s[i:i+n]]++` uses sub-slices into the input string as map keys. Strings are immutable so this is safe for correctness, but the map retains references to the underlying string data — if a consumer extracts a small q-gram map from a multi-MB input then discards the input, the input's backing array stays alive as long as any map key references it. The file header (lines 76-77) acknowledges this but doesn't warn the consumer.
- **Standard:** documentation-standards §"godoc"
- **Action:** Add a godoc note to the algorithm functions (or just the algorithm-correctness-reviewer can decide whether to copy on capture).
- **Rationale:** Retained-reference behaviour is unusual; should be documented even if intentional.

### [Important] `Strcmp95` source citation drift — algoid.go cites Winkler 1991, file cites Winkler 1994
- **File:** `/Users/johnny/Development/fuzzymatch/algoid.go:90-95` vs `/Users/johnny/Development/fuzzymatch/strcmp95.go:18-19`
- **Phase introduced:** Phase 3
- **Issue:** algoid.go's `AlgoStrcmp95` godoc cites "Winkler & Thibaudeau 1991 — An application of the Fellegi-Sunter model of record linkage to the 1990 U.S. decennial census". strcmp95.go's file header cites "Winkler, W. E. (1994). Advanced methods for record linkage. ...§3". Both citations are real Winkler publications but the algorithm-correctness discipline requires a single primary source. The file header is the authoritative source per algorithm-correctness-standards §"Primary Source Citation".
- **Standard:** algorithm-correctness-standards §"Primary Source Citation"
- **Action:** Align — either update algoid.go to cite Winkler 1994 (matching the file) or update strcmp95.go to cite Winkler 1991. The implementation actually transcribes the 1994 TR-2 36-pair table, so 1994 is correct.
- **Rationale:** Consistency between catalogue entry and implementation file.

### [Important] `JaroWinklerScoreRunes` does redundant `runeSlicesEqual` inside `jaroRunes`
- **File:** `/Users/johnny/Development/fuzzymatch/jaro.go:255-257` (called from jarowinkler.go:165)
- **Phase introduced:** Phase 2
- **Issue:** `jaroRunes` performs `if runeSlicesEqual(ra, rb) { return 1.0 }` after `JaroScoreRunes`/`JaroWinklerScoreRunes` have already gated on `a == b` (string equality). Since `[]rune(s)` is deterministic, two strings that differ MUST produce different rune slices UNLESS both contain invalid UTF-8 byte sequences that fold to U+FFFD identically. The check is reachable only for malformed UTF-8 inputs.
- **Standard:** go-coding-standards §"Complexity"
- **Action:** Either document this rationale inline (the check is correct, just non-obvious) or remove it and add a test case for the malformed-UTF-8 path.
- **Rationale:** Code review readability — readers wonder why the check exists.

### [Important] `runeAt` in `soundex.go` re-implements `utf8.DecodeRuneInString` without continuation-byte validation
- **File:** `/Users/johnny/Development/fuzzymatch/soundex.go:271-300`
- **Phase introduced:** Phase 7
- **Issue:** `runeAt` is shared across soundex.go, nysiis.go, mra.go, double_metaphone.go. It decodes UTF-8 length prefixes but does NOT validate that continuation bytes are 10xxxxxx — for malformed UTF-8 where the prefix byte says "2-byte sequence" but the next byte is e.g. 11xxxxxx, the function still returns a 2-byte step but with wrong rune value. Stdlib `utf8.DecodeRuneInString` does this correctly. The comment "avoids an extra import in a small file; the logic is identical to utf8.DecodeRuneInString for the purpose of skip-counting" is inaccurate.
- **Standard:** go-coding-standards §"Dependencies"; algorithm-correctness-standards §"Unicode Handling"
- **Action:** Replace `runeAt` with `utf8.DecodeRuneInString` (add the `unicode/utf8` import). The callers only use the size return; the stdlib function is bytes-equivalent for size on valid input and slightly safer on malformed input.
- **Rationale:** Stdlib already provides the correct primitive; rolling your own is unjustified complexity and slightly wrong.

### [Important] `MRACompare` step 3 L→R inner loop has a stale `matchedA[i]` guard
- **File:** `/Users/johnny/Development/fuzzymatch/mra.go:285-293`
- **Phase introduced:** Phase 7
- **Issue:** The inner loop conditions `if !matchedA[i] && !matchedB[j] && codexA[i] == codexB[j]` — but `matchedA[i]` is checked inside a loop where i is the OUTER index. Once `matchedA[i]` is set, the break fires and the outer loop advances. So the `!matchedA[i]` check is always true on entry (the outer loop has just started this `i`). It's a no-op guard that adds a tiny CPU cost and adds reading confusion.
- **Standard:** go-coding-standards §"Complexity"
- **Action:** Remove the redundant `!matchedA[i]` check inside the inner loop body. Keep only `!matchedB[j] && codexA[i] == codexB[j]`.
- **Rationale:** Tiny correctness-preserving simplification; current code is harder to verify.

### [Important] `MRACompare` step 3 doesn't require `j >= i` — semantic divergence from "L→R common-character removal"
- **File:** `/Users/johnny/Development/fuzzymatch/mra.go:285-293`
- **Phase introduced:** Phase 7
- **Issue:** The NBS Tech Note 943 step 3 description is "process L→R, remove identical characters from both codexes". The current implementation matches char A[i] to ANY unmatched char B[j] (any j, including j < i). This is the standard interpretation but it's not strictly L→R — it's "for each A[i], find the leftmost unmatched B[j]". Compare with jellyfish (which uses the strict-LR variant: `for j >= i`). The implementation matches the most permissive interpretation, which may produce higher similarity scores than the strict variant.
- **Standard:** algorithm-correctness-standards §"Reference Vectors"
- **Action:** Cross-validate against the committed `testdata/cross-validation/phonetic/vectors.json` and document which variant the implementation matches. If cross-validation passes against jellyfish, the current code is correct; otherwise it diverges.
- **Rationale:** The algorithm-correctness-reviewer should sign off explicitly on the interpretation choice.

### [Important] `double_metaphone.go` line 563 has ambiguous operator precedence
- **File:** `/Users/johnny/Development/fuzzymatch/double_metaphone.go:563`
- **Phase introduced:** Phase 7
- **Issue:** `if i == 0 && at(i+4) == 0 || dmContains(v, 0, "SAN") {` — Go evaluates as `(i == 0 && at(i+4) == 0) || dmContains(v, 0, "SAN")` per `&&`-higher-than-`||` precedence. Reading code is non-obvious; a reviewer might mis-read as `i == 0 && (at(i+4) == 0 || dmContains(v, 0, "SAN"))`.
- **Standard:** go-coding-standards §"Complexity"
- **Action:** Add explicit parentheses: `if (i == 0 && at(i+4) == 0) || dmContains(v, 0, "SAN") {`.
- **Rationale:** Readability and review-safety; same logic, clearer code.

### [Important] `scorer.go` file header claims `sort` is imported but it isn't
- **File:** `/Users/johnny/Development/fuzzymatch/scorer.go:69-72`
- **Phase introduced:** Phase 8
- **Issue:** Header comment: "the only stdlib dependency is 'sort' (not even strictly necessary at this plan boundary — AlgoIDs() returns the canonical order — but reserved for plan 08-03's Algorithms() accessor which sorts a fresh slice copy)." But (a) `sort` is NOT actually imported in scorer.go, (b) `Algorithms()` at line 460 does NOT sort — it just iterates the already-sorted slice.
- **Standard:** documentation-standards §"godoc accuracy"
- **Action:** Delete the misleading sentence from the file header.
- **Rationale:** Comment drift.

### [Important] `scorer_options_internal_test.go` godoc misnames package suffix
- **File:** `/Users/johnny/Development/fuzzymatch/scorer_options_internal_test.go:25-28`
- **Phase introduced:** Phase 8
- **Issue:** Comment: "Living in package fuzzymatch (no _test suffix) is the conventional Go pattern for exposing package-internal state to external test files: the build-tag _test.go suffix ensures this file never ships in the public artifact." But the file IS named `_test.go` (suffix `_test.go` excludes from production builds). The comment confuses "package fuzzymatch_test" with "file _test.go" — these are different mechanisms.
- **Standard:** documentation-standards §"godoc accuracy"
- **Action:** Reword to "Living in package fuzzymatch (not fuzzymatch_test) is the conventional Go pattern for exposing package-internal state to external test files via the _test.go file suffix, which excludes the file from production builds."
- **Rationale:** Same mistake in `/Users/johnny/Development/fuzzymatch/scorer_internal_test.go:28-30`.

### [Important] Map iteration in `extractQGrams` consumers — comment claims DET-03 satisfied via integer-counter exit, but Cosine's intersection key build iterates `small` map
- **File:** `/Users/johnny/Development/fuzzymatch/cosine.go:319-323`
- **Phase introduced:** Phase 5
- **Issue:** `for k := range small { if _, ok := large[k]; ok { intersectionKeys = append(intersectionKeys, k) } }` — this iterates a map and builds an ordered slice. The comment on line 311-313 says "the slice content is identical regardless of which side is iterated (intersection is symmetric)". This is TRUE for the SET of keys, but `intersectionKeys` is a slice — its ORDER depends on map iteration order, which is non-deterministic per Go map semantics. Subsequent `sort.Strings(intersectionKeys)` normalises the order, so the final reduction IS deterministic. But the comment claim that "the slice content is identical regardless of which side is iterated" omits the load-bearing role of the sort. This is correct code but partially misleading documentation.
- **Standard:** determinism-standards §"The No-Map-Iteration Rule"
- **Action:** Clarify the comment: "the slice CONTENTS (as a set) are identical regardless of which side is iterated; the slice ORDER is non-deterministic until the subsequent sort.Strings call normalises it before the dot-product reduction."
- **Rationale:** Doc accuracy; the sort is load-bearing per CONTEXT.md §3 LOCKED but the comment understates this.

### [Important] `LongestCommonSubstring` returns a slice header sharing backing storage with input
- **File:** `/Users/johnny/Development/fuzzymatch/lcsstr.go:116-155`
- **Phase introduced:** Phase 3
- **Issue:** `LongestCommonSubstring` returns `a[endI-maxLen : endI]` — a sub-slice of the input. Per Go's string semantics this is safe (strings are immutable), but it keeps the input's backing array alive. The godoc DOES mention this (lines 116-126), which is good. However, this is a public API behaviour that consumers may not expect — typical Go libraries return defensively-copied strings.
- **Standard:** documentation-standards §"godoc"
- **Action:** Consider whether to defensively copy in the byte variant (matching the rune variant which does `string(ra[endI-maxLen : endI])` — itself an allocation). Document the rationale either way.
- **Rationale:** API ergonomics review; the byte variant's "shared backing storage" behaviour is unusual and may surprise consumers retaining results across the lifetime of large inputs.

### [Important] `damerau_full.go` rune path uses `map[rune]int` — potentially non-deterministic if accessed for output
- **File:** `/Users/johnny/Development/fuzzymatch/damerau_full.go:362-399`
- **Phase introduced:** Phase 2
- **Issue:** `da := make(map[rune]int)` is used for the rune-path last-occurrence table. Inside the DP loop only `l := da[rb[j-1]]` (point lookup) and `da[ra[i-1]] = i` (point write) — never range-iterated. So DET-03 is satisfied. BUT: an iteration here would silently break determinism without raising any compile-time error. A future refactor that adds e.g. debug logging via `for k, v := range da` would corrupt the output. Worth a defensive comment.
- **Standard:** determinism-standards §"The No-Map-Iteration Rule"
- **Action:** The existing inline comment at line 361 says "Map LOOKUP only — not iterated to produce output (DET-03)." — this is good. No change needed but flagged for future-proofing.
- **Rationale:** Defensive review.

---

## IMPROVEMENT — minor cleanups, naming, godoc polish

### [Improvement] `algoid.go` `AlgoIDs()` allocates a 23-element slice on every call
- **File:** `/Users/johnny/Development/fuzzymatch/algoid.go:282-308`
- **Phase introduced:** Phase 1
- **Issue:** Each call to `AlgoIDs()` allocates a fresh 23-element slice. The Scorer uses it during `NewScorer` (one allocation per scorer construction), `Algorithms()` doesn't but iterates `algorithmsAlgoIDSorted` directly. For hot-path consumers calling `AlgoIDs()` repeatedly, this is unnecessary.
- **Action:** Consider returning a `[]AlgoID` that's pre-allocated as a package-level `var` and document immutability — but this contradicts the "freshly allocated so the caller may freely mutate" contract. Alternative: keep current behaviour and document the alloc cost.
- **Rationale:** Minor perf; not on the hot path.

### [Improvement] `Tokenise` empty input returns `[]string{}` not nil — inconsistent with idiomatic Go
- **File:** `/Users/johnny/Development/fuzzymatch/tokenise.go:143-146`
- **Phase introduced:** Phase 1
- **Issue:** Empty input returns a non-nil empty slice. Go idiom is usually nil. The godoc explicitly says "Returned slice is never nil" — a deliberate API decision but unusual.
- **Action:** Document the rationale for "never nil" — typical reason is "consumers can safely `range` without nil-check" but Go's range over nil slice is a no-op anyway.
- **Rationale:** API ergonomics; not blocking.

### [Improvement] `Soundex` first-letter scan loop has subtle control flow
- **File:** `/Users/johnny/Development/fuzzymatch/soundex.go:160-181`
- **Phase introduced:** Phase 7
- **Issue:** The for-loop body has multiple paths: ASCII letter sets `found = true`; non-letter increments `i`; non-ASCII increments by rune size. After each iteration `if found { break }`. The structure is correct but hard to follow. Could be simplified to two nested loops or factored into a helper.
- **Action:** Refactor into a helper `findFirstASCIILetter(s string) (b byte, nextIdx int, found bool)` — reduces visual complexity at the call site.
- **Rationale:** Readability.

### [Improvement] `SoundexCode` digit counter naming
- **File:** `/Users/johnny/Development/fuzzymatch/soundex.go:154`
- **Phase introduced:** Phase 7
- **Issue:** `digits := 1` — but `digits` starts at 1 because position 0 is reserved for the first letter, not a digit. The variable counts WRITE POSITIONS in the result, not digit characters. Naming should be `writePos` or `nextOut`.
- **Action:** Rename to `nextOut` or `writePos`.
- **Rationale:** Naming clarity.

### [Improvement] `NYSIISCode` step 7 trailing-A removal happens after step 6 which already converted AY→Y — redundant in practice
- **File:** `/Users/johnny/Development/fuzzymatch/nysiis.go:307-316`
- **Phase introduced:** Phase 7
- **Issue:** Step 6 converts `AY` → `Y` (last 2 chars). Step 7 removes a trailing A. These cannot fire together — after step 6, the trailing char is Y, not A. The two checks are independent (a single trailing A without leading A still hits step 7). Code is correct but the proximity might suggest interplay.
- **Action:** Add a one-line comment to step 7 clarifying "fires when step 6 did NOT fire" — i.e. trailing non-A non-AY-suffixed A.
- **Rationale:** Reviewer aid.

### [Improvement] `damerauOSADP` documentation references "linevels" rather than rows
- **File:** `/Users/johnny/Development/fuzzymatch/damerau_osa.go:200-256`
- **Phase introduced:** Phase 2
- **Issue:** Variables are `prevprev`, `prev`, `curr`. The DP "three-row rolling" terminology in the godoc is clear. No issue, just noting that the rotation `prevprev, prev, curr = prev, curr, prevprev` is non-obvious — the comment "After this: the row we just computed (curr) becomes prev, the previous prev becomes prevprev, and the old prevprev is handed back as the new curr (to be overwritten)" is good and should be preserved.
- **Action:** None; well-documented.
- **Rationale:** N/A.

### [Improvement] `Jaro` constant `maxJaroStackLen = 256` differs from Levenshtein's `maxStackInputLen = 64`
- **File:** `/Users/johnny/Development/fuzzymatch/jaro.go:107` vs `/Users/johnny/Development/fuzzymatch/levenshtein.go:68`
- **Phase introduced:** Phase 2
- **Issue:** Two distinct stack-buffer thresholds. Jaro's match-flag arrays are bool (256 × 1 byte = 256 bytes per side); Levenshtein's DP rows are int (65 × 8 bytes × 2 rows = 1040 bytes). Both fit comfortably on the stack; the divergence in threshold is correct but the rationale should be in a comment at the package-level (e.g. doc.go) so reviewers see the design discipline.
- **Action:** Add a paragraph to doc.go or a new `internal_constants.go` explaining the stack-buffer threshold philosophy.
- **Rationale:** Reviewer aid for future algorithm additions.

### [Improvement] `Tversky` panic message uses lowercase "tversky" without article
- **File:** `/Users/johnny/Development/fuzzymatch/tversky.go:242,284` and `/Users/johnny/Development/fuzzymatch/errors.go:88`
- **Phase introduced:** Phase 5
- **Issue:** `panic("fuzzymatch: invalid tversky parameter")` — "tversky" is a proper noun (Amos Tversky). Style convention is either "Tversky" or document the lowercasing.
- **Action:** Change to "fuzzymatch: invalid Tversky parameter" — but this requires updating the sentinel error message which is part of the v1.x contract. Defer until a major version bump.
- **Rationale:** Cosmetic; flag for future major release.

### [Improvement] `partialRatioRegion2Bytes` early-exit `best == 1.0` comparison
- **File:** `/Users/johnny/Development/fuzzymatch/partial_ratio.go:386-389`
- **Phase introduced:** Phase 6
- **Issue:** `if best == 1.0 { return best, true }` — exact float equality. `indelRatio` can return exactly 1.0 when `2*lcs == la+lb` (e.g. identity), so the comparison is safe. But the comment doesn't explain why exact comparison is OK here.
- **Action:** Add inline comment: "best == 1.0 is exact IEEE-754 because indelRatio returns 2*lcs/(la+lb) which is exact when 2*lcs == la+lb (identity case)."
- **Rationale:** Reviewer aid; defending against future "should this be `best >= 1.0 - epsilon`?" debates.

### [Improvement] `partial_ratio.go` Region 2 / Region 3 overlap at i = n-m
- **File:** `/Users/johnny/Development/fuzzymatch/partial_ratio.go:395-411`
- **Phase introduced:** Phase 6
- **Issue:** When `n > m`, Region 2 evaluates `i = n-m` AND Region 3 also evaluates `i = n-m` (since Region 3 starts at `i = n-m`). The same alignment is computed twice. The file header (line 88-90) acknowledges this: "When n > m there is a single trivial overlap at i = n-m with Region 2 (one redundant indelRatio call, harmless and matches the RapidFuzz reference behaviour)." Acceptable, just noting.
- **Action:** None; documented.
- **Rationale:** N/A.

### [Improvement] `cosine.go` upper clamp comment is excellent but the lower clamp comment is buried
- **File:** `/Users/johnny/Development/fuzzymatch/cosine.go:381-389`
- **Phase introduced:** Phase 5
- **Issue:** The upper-clamp comment (lines 372-381) is detailed (ULP overshoot rationale). The lower-clamp comment (lines 381-384) says "theoretically unreachable but the clamp costs nothing" — fine but less detailed than the upper.
- **Action:** None; cosmetic.
- **Rationale:** Style.

### [Improvement] `monge_elkan.go` allow-list `permittedMongeElkanInner` uses a map for lookup but exhaustive panic test walks all AlgoIDs
- **File:** `/Users/johnny/Development/fuzzymatch/monge_elkan.go:291-317`
- **Phase introduced:** Phase 6
- **Issue:** `var permittedMongeElkanInner = map[AlgoID]bool{...}` — map literal with 18 true entries. For 23 total AlgoIDs, a `[23]bool` array indexed by `int(AlgoID)` would be denser, zero-allocation, and inherently bounds-checked.
- **Action:** Consider switching to `var permittedMongeElkanInner = [numAlgorithms]bool{ AlgoLevenshtein: true, ... }`. The lookup site `if !permittedMongeElkanInner[inner]` would need a bounds check.
- **Rationale:** Minor perf; declarative-data improvement.

### [Improvement] `scorer.go` panic message in `DefaultScorer` includes error text — could simplify to `panic("...")` + err
- **File:** `/Users/johnny/Development/fuzzymatch/scorer.go:589`
- **Phase introduced:** Phase 8
- **Issue:** `panic("fuzzymatch: DefaultScorer construction failed (this is a bug): " + err.Error())` — string concatenation in a panic. Idiomatic Go would use `fmt.Errorf` wrapping or `log.Panicf`.
- **Action:** Consider `panic(fmt.Errorf("fuzzymatch: DefaultScorer construction failed: %w", err))` — panic value is the error chain, recoverable downstream.
- **Rationale:** Idiomatic Go.

### [Improvement] `errors.go` sentinel godoc references "Phase 9" (scan) which doesn't exist yet
- **File:** `/Users/johnny/Development/fuzzymatch/errors.go:27` ("richer per-item context can be added in Phase 9 if scan needs it")
- **Phase introduced:** Phase 1
- **Issue:** Phase 9 is the documented future phase; the comment is forward-looking. No issue, just noting the dependency.
- **Action:** None; documented intent.
- **Rationale:** N/A.

### [Improvement] `tokenise.go` `lowerRuneToken` duplicates `lowerRune` in normalise.go
- **File:** `/Users/johnny/Development/fuzzymatch/tokenise.go:340-354` vs `/Users/johnny/Development/fuzzymatch/normalise.go:441-452`
- **Phase introduced:** Phase 1
- **Issue:** Two identical functions named differently for "no internal dependency on normalise.go" per tokenise.go's comment. The duplication is deliberate but introduces a maintenance burden (if Unicode handling changes, both must update). Compare with the strcmp95 / soundex `runeAt` duplication which IS shared.
- **Action:** Consider factoring into an `internal_ascii.go` file with a single shared `asciiToLowerRune` helper. The "no internal dependency" rule isn't enforced for shared primitives.
- **Rationale:** DRY.

### [Improvement] `algoid.go` `String()` uses a 23-case switch — could be a literal lookup
- **File:** `/Users/johnny/Development/fuzzymatch/algoid.go:213-267`
- **Phase introduced:** Phase 1
- **Issue:** The 23-case switch is correct and clearly documented (the nolint:gocyclo defense is good). An alternative is a package-level `var algoIDStrings = [numAlgorithms]string{...}` with O(1) lookup. The current switch is the canonical Go idiom for stringly-typed enums, but the lookup table would be slightly faster on the hot path (if `String()` is ever called from a hot path, which it shouldn't be).
- **Action:** None; the switch is appropriate for non-hot-path stringification.
- **Rationale:** Reviewer style preference.

### [Improvement] `scorer.go` `Scorer.Score` reduction loop comment block is dense
- **File:** `/Users/johnny/Development/fuzzymatch/scorer.go:349-383`
- **Phase introduced:** Phase 8
- **Issue:** The comment in `Score` (lines 359-380) is ~20 lines of dense determinism rationale for an 11-line function body. The FMA-fusion remediation comment is informative but could be in a separate `// See …` reference rather than inline.
- **Action:** Consider extracting the FMA-fusion documentation into a `determinism.md` doc file or a comment in scorer.go's header (which already has it), reducing the per-call-site verbosity.
- **Rationale:** Readability; the inline density makes the code body harder to scan.

### [Improvement] `mra.go` `mraThreshold` uses a 13-element array with sum>12 clamp at function boundary
- **File:** `/Users/johnny/Development/fuzzymatch/mra.go:95-109`
- **Phase introduced:** Phase 7
- **Issue:** `mraThresholdTable` is sized for `sum in [0, 12]`. For `sum > 12`, the function returns 2 (clamp) without indexing the table. Clean design, but the table size 13 and the clamp boundary 12 are magic numbers; the file comment explains them but a named constant `mraSumLimit = 12` would improve clarity.
- **Action:** Add `const mraSumLimit = 12 // upper bound of the NBS-943 Table A`.
- **Rationale:** Magic-number elimination.

### [Improvement] `normalise.go` `applyUnicodeTransformer` constructs a fresh `transform.Chain` per call
- **File:** `/Users/johnny/Development/fuzzymatch/normalise.go:320-338`
- **Phase introduced:** Phase 1
- **Issue:** The chain is built fresh per call ("transform.Transformer is not documented as safe for concurrent reuse, and per-call construction is cheap"). For Scorer use, this fires twice per `Score(a, b)`. Profiling might justify a sync.Pool of chains, but the comment defers this.
- **Action:** None; documented as v1.x deferral.
- **Rationale:** N/A.

### [Improvement] `lcsstr.go` `LongestCommonSubstring` returns shared backing storage — could be a documented design or could be defensive-copied
- **File:** `/Users/johnny/Development/fuzzymatch/lcsstr.go:116-155`
- **Phase introduced:** Phase 3
- **Issue:** See "Important" finding above on retained-reference behaviour. As an improvement, consider whether the rune variant's freshly-allocated string (line 187) and the byte variant's sub-slice are semantically equivalent for consumers. The mismatch is documented but unexpected.
- **Action:** Decide between (a) make both variants share backing storage (faster), (b) make both defensive-copy (consistent, safe), or (c) document the difference more prominently.
- **Rationale:** API consistency.

### [Improvement] `swg.go` `SmithWatermanGotohRawScoreRunes` has duplicated identity-short-circuit logic
- **File:** `/Users/johnny/Development/fuzzymatch/swg.go:287-313`
- **Phase introduced:** Phase 3
- **Issue:** The identity branch logic (`if a == "" return 0.0; return Match * float64(len([]rune(a)))`) duplicates the byte path's logic in `SmithWatermanGotohRawScoreWithParams`. The comment (lines 290-295) explains the duplication is deliberate to preserve the identity short-circuit's purpose.
- **Action:** None; documented intent.
- **Rationale:** N/A.

### [Improvement] `dispatch_*.go` documentation duplicates the `var _ = func() bool` idiom rationale in every file
- **File:** All 23 `/Users/johnny/Development/fuzzymatch/dispatch_*.go` files
- **Phase introduced:** Phases 2–7
- **Issue:** Each dispatch file has ~10 lines explaining the `var _ = func() bool { ... }()` idiom. This is 230 lines of duplicated boilerplate documentation. A single reference in algoid.go could replace it.
- **Action:** Consider replacing the per-file boilerplate with a one-line `// See algoid.go for the dispatch-registration pattern rationale.` reference, with the full explanation centralised.
- **Rationale:** DRY; reduce maintenance burden.

### [Improvement] `partial_ratio.go` rune charSet uses `map[rune]struct{}` — could be a small slice
- **File:** `/Users/johnny/Development/fuzzymatch/partial_ratio.go:486-495`
- **Phase introduced:** Phase 6
- **Issue:** `make(map[rune]struct{}, m)` for the rune charSet. For typical inputs `m` is small (< 20 runes), so a sorted `[]rune` with binary search OR a linear scan would be faster than a map (no hash overhead) and zero-allocation on stack.
- **Action:** Profile and consider replacing with a stack-allocated `[20]rune` for short inputs with fallback to a map for longer ones.
- **Rationale:** Perf.

### [Improvement] No `examples/` directory contents reviewed but listed in `ls` output
- **File:** `/Users/johnny/Development/fuzzymatch/examples/` (identifier-similarity, phonetic-keys, scorer-composition)
- **Phase introduced:** Phase 8
- **Issue:** Three sub-examples exist but their compilation status, godoc presence, and test coverage was not validated in this review.
- **Action:** Confirm `make check` includes building these examples; check that each has a `main_test.go` that compiles.
- **Rationale:** Public-facing examples need the same quality bar as the library.

---

## Notes on out-of-scope and related findings

- **scan/ sub-package**: empty/non-existent (Phase 9). All scan-related sentinel godoc references are forward-looking. Not reviewed.
- **`tests/bdd/`**: BDD test code not reviewed in this pass — see `08-BDD-REVIEW.md` for prior phase-8 BDD review.
- **Determinism**: Phase 8 has dedicated determinism review (`08-DETERMINISM-REVIEW.md`). No new determinism issues found in this whole-codebase pass beyond the Cosine intersection-keys comment improvement noted above.
- **License headers**: every reviewed `.go` file has the AxonOps Apache 2.0 header. No issues.
- **Test files (`*_test.go`)**: spot-checked; full review is `test-analyst`'s scope.

## Key recommendations for next phase / cycle

1. **Resolve the Hamming `(int, error)` contract** before v1.0 freeze — option (a) update spec to silent return-max, or (b) introduce ErrHammingLengthMismatch and break API.
2. **Clean up `bench.txt.new`** committed to repo root — either fix the failing property test or delete the artefact with CHANGELOG note.
3. **Tighten Scorer option layer** — add α+β > 0 check to WithTverskyAlgorithm and permittedMongeElkanInner check to WithMongeElkanAlgorithm.
4. **Fix the dispatch-file slot-number drift** — three files have wrong iota numbers; remove the explicit numbers entirely.
5. **Move `WriteGoldenFile` out of the public API** — into a `_test.go` file or `internal/golden` package.
