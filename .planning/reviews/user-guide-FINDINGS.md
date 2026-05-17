---
status: issues_found
agent: user-guide-reviewer
scope: consumer-facing docs (phases 1-8)
reviewed: 2026-05-17T06:36:00Z
finding_counts:
  critical: 7
  important: 22
  improvement: 18
  total: 47
---

## Summary

The library has shipped Phase 8 (Scorer + DefaultScorer + 23 algorithms + 39 godoc examples). The consumer-facing documentation has NOT kept pace.

A new user landing on the README today sees a "Quick Start" framed around Phase 1 primitives (Normalise / Tokenise) with no mention that 23 algorithms and a working `DefaultScorer()` are already available. The single Quick Start program in README is also factually wrong: `Tokenise("XMLHttpRequest", ...)` actually returns `[xml http request]`, not the documented `[xmlhttp request]`. The README never shows a `go get` command, never shows `LevenshteinScore`, never shows `DefaultScorer().Score(...)`, never points at the three runnable `examples/` directories, and never demonstrates the error-handling sentinels.

`docs/algorithms.md` is the most visible regression: 100% scaffold with `TBD` primary sources for every one of the 23 algorithms. README's catalogue table deep-links into this file. Every link goes to a stub.

`docs/scorer.md` is the strongest piece of new documentation but contains: (a) an internal contradiction ("All four methods" describing a five-method table), (b) an aspirational claim that `ErrInvalidThreshold` rejects NaN that does not match the code (CR-01 — confirmed against `scorer_options.go:257-266`), (c) inconsistent error-handling across its code samples, (d) `sort.Slice` and `fmt.Printf` in examples with no imports shown, (e) no warning that the "default minus algorithm" pattern relies on options last-write-wins for the threshold, and (f) no callout that `WithoutNormalisation` does NOT actually pass raw bytes to token-based algorithms (Tokenise still lowercases internally).

`docs/tuning.md` is the most polished consumer doc but does not link to any concrete calibration script or example — the "100-500 pair calibration loop" has no starter code.

`docs/scan.md` and `docs/extending.md` are scaffolds with TBD bodies — fine for pre-Phase-9, but the README points at them as if they were ready.

The `examples/` directory has three carefully-crafted runnable programs that are essentially invisible: the README does not mention `examples/` even once. `docs/scorer.md` is the only consumer-facing doc that links to them.

TTHW measurement:
- README Quick Start: 6 lines of actual code (including 2 imports and 2 var statements), prints `[xml http request]` — but README claims `[xmlhttp request]`. **First contact with the library yields a contradiction.**
- No `go get` command anywhere in any consumer-facing doc.
- TTHW for `LevenshteinScore("kitten", "sitting")`: not present in any consumer-facing prose. Only in `example_test.go` (which the README never links to as the canonical entry point).
- TTHW for `DefaultScorer().Score("user_id", "userId")`: present in `docs/scorer.md` Quickstart only.
- Headline example error handling: README example uses `Normalise` (which cannot fail). `DefaultScorer().Score(...)` would be a cleaner zero-error headline; not shown in README.

Severity tags below are organisational only. All findings should be addressed before v1.0.

---

## Critical

### [Critical] README Quick Start example output is wrong
- **File:** /Users/johnny/Development/fuzzymatch/README.md (lines 107-117)
- **Phase introduced:** Phase 1
- **Issue:** The Quick Start prints `// Output: [xmlhttp request]` for `Tokenise("XMLHttpRequest", DefaultTokeniseOptions())`. The actual output is `[xml http request]` (verified by running the program). `DefaultTokeniseOptions().SplitConsecutiveUpper = true` splits "XMLHttp" into "xml" + "http" before splitting on the case transition into "Request".
- **Standard:** documentation-standards.md "Quick Start: copy-paste and it compiles and runs"; "Every code example in documentation must compile (verified by `documentation_test.go`)"
- **Action:** Code fix (the doc) — either update the comment to `[xml http request]` or change the input to something that actually produces a meaningful pair like `[xmlhttp request]` (e.g. by tweaking options or using a different camelCase input).
- **Rationale:** First contact with the library produces a "the docs lied to me" moment for any user who copy-pastes and runs.
- **Suggested fix:** Pin the example to `// Output: [xml http request]` and add a meta-test that runs the README's first code block (the standards skill calls this `readme_shop_front_test.go`).

### [Critical] docs/scorer.md claims `ErrInvalidThreshold` rejects NaN; code does not (CR-01)
- **File:** /Users/johnny/Development/fuzzymatch/docs/scorer.md (line 283); /Users/johnny/Development/fuzzymatch/scorer_options.go (lines 257-266)
- **Phase introduced:** Phase 8
- **Issue:** Documentation row reads `ErrInvalidThreshold | Returned when WithThreshold receives a value outside [0.0, 1.0], or a NaN.` The implementation only checks `t < 0.0 || t > 1.0`, both of which are `false` for `math.NaN()`. So `WithThreshold(math.NaN())` succeeds, the Scorer freezes with `threshold = NaN`, and `Match(a, b)` (which evaluates `score >= NaN`, always false) silently never matches anything. This is the exact CR-01 finding from `.planning/phases/08-composite-scorer/08-REVIEW.md`.
- **Standard:** "Does the documentation match the actual code behaviour?" — user-guide-reviewer key question.
- **Action:** Code fix (fix the code per CR-01: add `math.IsNaN(t)` check). Documentation is correct and should remain.
- **Rationale:** Silent no-match malfunctions are the worst class of consumer bug — the Scorer looks like it's "just very strict" until the consumer realises nothing is ever matching.

### [Critical] docs/algorithms.md is 100% scaffold with TBD primary sources for all 23 algorithms after Phases 2-7 shipped
- **File:** /Users/johnny/Development/fuzzymatch/docs/algorithms.md (every algorithm section, lines 25-238)
- **Phase introduced:** Phase 1 (scaffold); should have been filled in over Phases 2-7
- **Issue:** Every algorithm entry says `Primary source: TBD — filled in by the implementing phase`. The README's catalogue table deep-links into this file (`docs/algorithms.md#levenshtein`, `docs/algorithms.md#damerau-levenshtein-osa`, etc.). All 23 links resolve to a 5-line stub. Meanwhile the actual primary source is cited in inline Go file comments and in `llms.txt` (where it is up to date) — so a sophisticated reader can find the citation, but the consumer-targeted doc has been overlooked.
- **Standard:** documentation-standards.md "Algorithm Documentation" mandates per-algorithm: name, category, primary academic source citation, description, mathematical formulation, complexity, score normalisation rule, mathematical invariants, edge cases, reference vectors with citation, intended use cases, comparable references.
- **Action:** Code fix (the docs). This is a major write-up but the source content already exists in inline Go godocs.
- **Rationale:** Without this, the README catalogue table is essentially a teaser — clicking through gives no information beyond the AlgoID enum value. New users can't make algorithm-selection decisions.

### [Critical] README Quick Start does not show the v0.x→v1.0 headline use case (algorithm score or DefaultScorer)
- **File:** /Users/johnny/Development/fuzzymatch/README.md (lines 94-117)
- **Phase introduced:** Phase 1 (stale); should have been updated at every phase boundary
- **Issue:** The README "Quick Start" still says `Phase 1 (foundation) ships Normalise and Tokenise primitives... Algorithm functions (e.g. LevenshteinScore) land in Phase 2. The example below uses the Phase-1 primitives`. Phase 8 has shipped. The example uses Normalise and Tokenise — neither of which is the headline use case. A new user sees lowercase-and-tokenise as the library's purpose, not similarity scoring.
- **Standard:** documentation-standards.md "First code block shows a complete, working program (the 'headline example')"; user-guide-reviewer "Can a developer compute a similarity score in under 2 minutes?"
- **Action:** Code fix (the docs). Replace Quick Start with a `DefaultScorer().Score(...)` example, OR a `LevenshteinScore(...)` example, ideally both.
- **Rationale:** The headline example IS the library's first impression. Today it's misleading.
- **Suggested fix:**
  ```go
  package main

  import (
      "fmt"
      "github.com/axonops/fuzzymatch"
  )

  func main() {
      score := fuzzymatch.LevenshteinScore("kitten", "sitting")
      fmt.Printf("%.4f\n", score) // 0.5714

      // Or use the opinionated DefaultScorer:
      s := fuzzymatch.DefaultScorer()
      fmt.Println(s.Match("user_id", "userId")) // true
  }
  ```

### [Critical] No `go get github.com/axonops/fuzzymatch` anywhere in consumer-facing docs
- **File:** /Users/johnny/Development/fuzzymatch/README.md; /Users/johnny/Development/fuzzymatch/docs/scorer.md; /Users/johnny/Development/fuzzymatch/llms.txt
- **Phase introduced:** Phase 1
- **Issue:** README never shows the install command. Bare imports of `github.com/axonops/fuzzymatch` appear in examples, but no `go get` invocation. The standard "install line" — the single most important sentence in any Go library README — is missing.
- **Standard:** documentation-standards.md "Quick Start: copy-paste and it compiles and runs". Implicitly: install + use.
- **Action:** Code fix (the docs).
- **Rationale:** TTHW failure. Step zero of "I found this library" is missing.

### [Critical] README never mentions the `examples/` directory
- **File:** /Users/johnny/Development/fuzzymatch/README.md
- **Phase introduced:** Phase 7 (when `examples/` was first introduced) or earlier
- **Issue:** Three carefully-built runnable programs live under `examples/`:
  - `examples/identifier-similarity/` (23-algorithm side-by-side + Default Scorer Score/Match columns)
  - `examples/phonetic-keys/` (SoundexCode, DoubleMetaphoneKeys, NYSIISCode, MRACode, MRACompare)
  - `examples/scorer-composition/` (DefaultScorer vs Default-minus-DoubleMetaphone)
  Each has a `main_test.go` pinning byte-stable expected output. None of these are linked from README. Only `docs/scorer.md` lines 336-339 mentions two of them. A new user has no way to discover that there are runnable demos.
- **Standard:** documentation-standards.md "Include runnable Example functions in `example_test.go` for major use cases"; the spirit applies to `examples/` directory programs too.
- **Action:** Code fix (the docs).
- **Rationale:** The runnable examples are the fastest path to confidence — and they're invisible.

### [Critical] README "Layer 2" illustrative code is syntactically misleading
- **File:** /Users/johnny/Development/fuzzymatch/README.md (line 82)
- **Phase introduced:** Phase 1
- **Issue:** The three-layers diagram shows `Layer 2: Scorer    NewScorer().Score(a, b)`. The actual API is `NewScorer(opts ...ScorerOption) (*Scorer, error)` — it MUST take options (at minimum `WithThreshold`) and returns an error. The diagram suggests a zero-argument call that yields a `*Scorer` directly. This is a documentation error a careful reader will trip over.
- **Standard:** "API ergonomics — the surface shape is what's shown"
- **Action:** Code fix (the docs).
- **Suggested fix:** `Layer 2: Scorer    DefaultScorer().Score(a, b)` (this composes correctly).

---

## Important

### [Important] No runnable godoc example for the Scorer
- **File:** /Users/johnny/Development/fuzzymatch/example_test.go (39 example functions but none for Scorer/NewScorer/DefaultScorer)
- **Phase introduced:** Phase 8
- **Issue:** `example_test.go` has 39 `Example*` functions for individual algorithms (verified by `grep '^func Example'`), but ZERO for any of: `Scorer`, `NewScorer`, `DefaultScorer`, `DefaultScorerOptions`, `WithAlgorithm`, `WithThreshold`, `Scorer.Score`, `Scorer.Match`, `Scorer.ScoreAll`, `Scorer.Algorithms`, `Scorer.Threshold`. pkg.go.dev will not show a runnable, output-verified example for the central Scorer API.
- **Standard:** documentation-standards.md "Include runnable `Example` functions in `example_test.go` for major use cases".
- **Action:** Code fix.
- **Rationale:** The runnable example is the canonical "this is exactly what you type" surface. Its absence forces consumers to learn the Scorer API from prose only.

### [Important] No runnable godoc examples for Normalise or Tokenise
- **File:** /Users/johnny/Development/fuzzymatch/example_test.go; /Users/johnny/Development/fuzzymatch/normalise_test.go; /Users/johnny/Development/fuzzymatch/tokenise_test.go
- **Phase introduced:** Phase 1
- **Issue:** No `ExampleNormalise` or `ExampleTokenise` anywhere in the repository. The README Quick Start uses both, but pkg.go.dev shows no runnable example for either function. Compare against the 39 algorithm examples.
- **Standard:** documentation-standards.md "runnable examples for every algorithm category" — Normalise/Tokenise are not algorithms but they are core to the public API.
- **Action:** Code fix.

### [Important] No documented `errors.Is` usage example
- **File:** /Users/johnny/Development/fuzzymatch/example_test.go; /Users/johnny/Development/fuzzymatch/docs/scorer.md
- **Phase introduced:** Phase 8 (the Scorer error sentinels are the first place a consumer encounters errors)
- **Issue:** `docs/scorer.md` line 275 says "discriminate via `errors.Is`, never by matching the error message string" — but no example of how. The doc lists 4 sentinels in a table; the consumer is left to know to write `if errors.Is(err, fuzzymatch.ErrMissingThreshold) { ... }`. `errors.go` godoc has the same prose but no example. A runnable godoc example would close the gap.
- **Standard:** user-guide-reviewer key question: "Are error sentinels documented with `errors.Is` usage example?"
- **Action:** Code fix.
- **Suggested fix:** Add `ExampleNewScorer_errors` in `example_test.go` showing the `errors.Is` discrimination pattern for at least `ErrMissingThreshold` and `ErrInvalidThreshold`.

### [Important] docs/scorer.md "All four methods" contradicts its five-row method-reference table (IN-01)
- **File:** /Users/johnny/Development/fuzzymatch/docs/scorer.md (lines 95-107, 258-259)
- **Phase introduced:** Phase 8
- **Issue:** Table at lines 97-103 lists FIVE methods: Score, Match, ScoreAll, Threshold, Algorithms. Immediately following text at line 105: "All four methods are pure functions". Repeated at line 258-259: "All four methods (`Score`, `Match`, `ScoreAll`, `Threshold`, `Algorithms`)" — calls five methods "four". Reader confidence damaged on first encounter.
- **Standard:** Documentation accuracy / self-consistency.
- **Action:** Code fix (the docs).
- **Suggested fix:** Change "four" → "five" both places, or restructure as "All Scorer methods".

### [Important] docs/scorer.md error-handling code samples are inconsistent
- **File:** /Users/johnny/Development/fuzzymatch/docs/scorer.md
- **Phase introduced:** Phase 8
- **Issue:** Quickstart at line 27-32 calls `DefaultScorer()` directly with no error handling (correct — it cannot fail). Custom example at lines 43-51 shows `if err != nil { return fmt.Errorf(...) }` (with imports/context unstated). Lines 67-72 (`opts := append(...)`) drops the `err` check entirely. Lines 87-92 (parameterised options) drops the `err` check. Lines 152-156 drops the `err` check with bare `_`. Inconsistent across the same document — readers can't tell what the recommended idiom is.
- **Standard:** documentation-standards.md "Every code example in documentation must compile"; consumer ergonomics.
- **Action:** Code fix (the docs). Standardise on the pattern from lines 43-51 throughout (the named `s, err := …` plus an `if err != nil` block), OR explicitly comment on each spot why ignoring the error is safe here.

### [Important] Code examples in docs/scorer.md use unimported packages
- **File:** /Users/johnny/Development/fuzzymatch/docs/scorer.md (lines 227-237)
- **Phase introduced:** Phase 8
- **Issue:** The ScoreAll example uses `sort.Slice` and `fmt.Printf` without `import "sort"` or `import "fmt"`. The reader has to infer that "of course you need those imports" — but the scorer.md Quickstart at lines 22-24 DOES show the import; the pattern is inconsistent. The CR's "every code example must compile" standard requires either the example shows imports or the snippet is unambiguously incomplete (e.g. inside a `// ...` block).
- **Standard:** documentation-standards.md "Every code example in documentation must compile (verified by `documentation_test.go`)"
- **Action:** Code fix (the docs).

### [Important] `WithoutNormalisation` semantic surprise not flagged
- **File:** /Users/johnny/Development/fuzzymatch/docs/scorer.md (lines 168-180)
- **Phase introduced:** Phase 8
- **Issue:** "With `WithoutNormalisation`, the raw input bytes are passed to every algorithm" — but token-based algorithms (Monge-Elkan, Token*, PartialRatio) internally call `Tokenise(s, DefaultTokeniseOptions())`, and `DefaultTokeniseOptions().Lowercase = true` (verified in `tokenise.go:107`). So `WithoutNormalisation` does NOT mean "no character changes" — it means "no Unicode normalisation pipeline, but tokens are still lowercased on the way to set/sort algorithms". This is a footgun. A consumer comparing case-sensitive identifiers will get unexpected lowercase-after-tokenise behaviour from any token algorithm.
- **Standard:** user-guide-reviewer key question: "Are common pitfalls called out (ScoreAll iteration non-determinism, WithoutNormalisation surprises, threshold tuning)?"
- **Action:** Code fix (the docs).
- **Suggested fix:** Add a callout block: "Note: token-based algorithms still call `Tokenise(DefaultTokeniseOptions())` internally regardless of `WithoutNormalisation`. The default `Tokenise` lowercases. To get truly byte-identical input through every algorithm, either pre-tokenise upstream or avoid the token-based algorithms in the Scorer composition."

### [Important] `WithTverskyAlgorithm` constraint `α + β > 0` is not documented in scorer.md or tuning.md (CR-02 surface)
- **File:** /Users/johnny/Development/fuzzymatch/docs/scorer.md (line 89 example uses α=β=0.5); /Users/johnny/Development/fuzzymatch/docs/tuning.md; /Users/johnny/Development/fuzzymatch/scorer_options.go (line 381-399)
- **Phase introduced:** Phase 8
- **Issue:** `docs/scorer.md` line 89 shows `WithTverskyAlgorithm(0.3, 0.5, 0.5, 3)` as the example. The α + β > 0 constraint is enforced at TverskyScore call time (panic in `tversky.go:241`) but the option layer accepts `(0, 0)`. A consumer who tries `WithTverskyAlgorithm(0.3, 0.0, 0.0, 3)` (perhaps to disable both asymmetric terms while keeping intersection-only) will see no error at NewScorer time and a panic on the first Score call. The docs nowhere warn that "α and β must both be ≥ 0 AND at least one must be > 0". This is CR-02 from Phase 8 review.
- **Standard:** "fail loudly at construction" — Phase 8 contract. Documentation should match.
- **Action:** Code fix (the code per CR-02). Then update docs to document the constraint.

### [Important] Phonetic encoder API surface (SoundexCode, DoubleMetaphoneKeys, NYSIISCode, MRACode, MRACompare) is absent from consumer docs
- **File:** /Users/johnny/Development/fuzzymatch/docs/algorithms.md; /Users/johnny/Development/fuzzymatch/docs/extending.md; /Users/johnny/Development/fuzzymatch/docs/scorer.md
- **Phase introduced:** Phase 7
- **Issue:** The phonetic algorithms expose both a binary score (`SoundexScore`, etc.) AND a raw encoder (`SoundexCode`, `DoubleMetaphoneKeys`, `NYSIISCode`, `MRACode`) plus `MRACompare(a, b) (bool, int)`. The encoder is the right tool for "compose phonetic codes with edit distance" patterns (documented in faq.md "Why phonetic-as-binary in the Scorer?" and gestured-at in docs/extending.md "TBD" stub). But the raw encoder API is documented ONLY in `docs/requirements.md` §7.20 (an internal spec) and in `examples/phonetic-keys/`. The README catalogue doesn't mention encoders exist. A consumer reading `docs/algorithms.md#soundex` (a TBD stub) has no idea they can call `SoundexCode("Robert") → "R163"`.
- **Standard:** documentation-standards.md "Each algorithm must be documented in `docs/algorithms.md` with... intended use cases"; user-guide-reviewer "Can they extend the library?"
- **Action:** Code fix (the docs).

### [Important] README Algorithm Catalogue links all resolve to TBD scaffolds
- **File:** /Users/johnny/Development/fuzzymatch/README.md (lines 127-171)
- **Phase introduced:** Phase 1 (links); Phase 2-7 (should have filled in)
- **Issue:** Every catalogue table entry deep-links into `docs/algorithms.md#<algo>`. Every target is a 5-line TBD stub. The links are functionally useless for learning what an algorithm does — they only confirm the algorithm exists.
- **Standard:** documentation-standards.md "Algorithm catalogue table... with: algorithm name, primary source citation, intended use, link to per-algorithm detail in `docs/algorithms.md`".
- **Action:** Code fix (fills in `docs/algorithms.md`, see Critical finding above).

### [Important] CHANGELOG.md has no entries for Phases 1-8
- **File:** /Users/johnny/Development/fuzzymatch/CHANGELOG.md
- **Phase introduced:** Phase 1
- **Issue:** The CHANGELOG `## [Unreleased]` block lists only the project bootstrap (repository structure, licence, CLAUDE.md, requirements doc, GSD). It does NOT list:
  - Phase 1: Normalise, Tokenise, AlgoID, errors, golden fixtures.
  - Phase 2: 6 character-based algorithms + Ratcliff-Obershelp.
  - Phase 3: Smith-Waterman-Gotoh.
  - Phase 4: 4 q-gram algorithms.
  - Phase 5: ASCII fast paths and Cosine.
  - Phase 6: 5 token-based algorithms.
  - Phase 7: 4 phonetic algorithms.
  - Phase 8: Scorer, NewScorer, DefaultScorer, ScoreAll, etc.
- **Standard:** documentation-standards.md "CHANGELOG following Keep-a-Changelog"; `release.yml` extracts the section for GitHub release notes.
- **Action:** Code fix.

### [Important] `docs/scan.md` is a scaffold but `docs/scan.md` is linked from README as a real doc
- **File:** /Users/johnny/Development/fuzzymatch/docs/scan.md; /Users/johnny/Development/fuzzymatch/README.md (line 225)
- **Phase introduced:** Phase 1 (scaffold); should remain as scaffold or be hidden until Phase 9
- **Issue:** `README.md` lists `docs/scan.md — `scan` sub-package consumer guide (Phase 9).` Clicking through gives a "TBD. See docs/requirements.md §12.X" stub for every section. The "(Phase 9)" annotation is helpful but the link target is essentially "we plan to write this". README should make clearer that this is aspirational, or hide the link until Phase 9 ships.
- **Standard:** Consumer-trust — links should go somewhere useful.
- **Action:** Code fix (the docs). Either gate the README link behind Phase 9 completion, or change the stub doc to say "Not yet shipped — track progress in [issue link]".

### [Important] `docs/extending.md` is a scaffold but linked as a real doc
- **File:** /Users/johnny/Development/fuzzymatch/docs/extending.md
- **Phase introduced:** Phase 1
- **Issue:** All four substantive sections (`## Composing domain-specific Scorers`, `## Composing phonetic algorithms with edit distance`, `## Custom inner metric for Monge-Elkan`, `## Custom algorithms outside the catalogue`) begin with "TBD." or are stubs. The README links to this doc. The user-guide-reviewer key question "Can they extend the library?" cannot be answered from current `docs/extending.md`.
- **Standard:** documentation-standards.md "docs/extending.md — building domain-specific Scorers, composing phonetic algorithms with edit distance, custom inner metrics for Monge-Elkan".
- **Action:** Code fix (the docs).

### [Important] `docs/performance.md` is a scaffold but linked as a real doc
- **File:** /Users/johnny/Development/fuzzymatch/docs/performance.md
- **Phase introduced:** Phase 1
- **Issue:** Every section is TBD. "Benchmark methodology" is TBD. "Per-algorithm budgets" is TBD. "Scorer budgets" is TBD. The committed `bench.txt` file exists with full data — but the doc points at `docs/requirements.md` §14 for the real numbers, which is an internal spec.
- **Standard:** documentation-standards.md "docs/performance.md — benchmark numbers, optimisation notes, profiling tips".
- **Action:** Code fix (the docs).

### [Important] No documented worked example for the "scan/Item/Warning" loop in `docs/scan.md`
- **File:** /Users/johnny/Development/fuzzymatch/docs/scan.md
- **Phase introduced:** Phase 9 (not yet shipped — but the scaffold should at least show what the API will look like)
- **Issue:** User-guide-reviewer key question 6: "Can they use the scan sub-package? A worked example from a slice of `Item` to a list of `Warning` in under 10 lines." Currently scan.md is all TBD pointing at requirements.md. A scaffold that previews the API surface would orient consumers planning their Phase 9 adoption.
- **Standard:** user-guide-reviewer.
- **Action:** Code fix (the docs) — when Phase 9 lands. Pre-Phase-9, downgrade to a clear "coming in Phase 9" banner.

### [Important] `docs/scorer.md` Threshold rationale is stated but no quick recommendation table
- **File:** /Users/johnny/Development/fuzzymatch/docs/scorer.md (lines 110-130); /Users/johnny/Development/fuzzymatch/docs/tuning.md (lines 41-58)
- **Phase introduced:** Phase 8
- **Issue:** docs/scorer.md tells you the threshold is mandatory and gives the rationale, but does NOT include a quick lookup table of "for X data, start at threshold Y". `docs/tuning.md` has the canonical calibration loop but it's a 6-step process. A "if you don't have labelled data, start at 0.85 for identifier matching, 0.80 for name matching, 0.70 for free text" table at the top of `docs/scorer.md` § Threshold would close the gap for the user who just wants a sane starting point.
- **Standard:** user-guide-reviewer key question 5: "Can they tune thresholds?... At minimum: starting point recommendations".
- **Action:** Code fix (the docs).

### [Important] Tuning guidance has no runnable calibration script
- **File:** /Users/johnny/Development/fuzzymatch/docs/tuning.md
- **Phase introduced:** Phase 8
- **Issue:** docs/tuning.md describes a 6-step calibration loop but offers no starter Go code. Compare to the high-quality `examples/scorer-composition/`. A `examples/threshold-calibration/` runnable program would close the loop: given a CSV of labelled pairs and a Scorer, print the precision/recall/F1 at every threshold in 0.05 increments. Today the consumer has to write this from scratch.
- **Standard:** user-guide-reviewer "Can they tune thresholds? ... and how to calibrate against a domain corpus."
- **Action:** Code fix (add an example directory) or doc-only fix (paste a starter snippet in tuning.md).

### [Important] `docs/scorer.md` ScoreAll deterministic-content note uses `sort.Slice` but the more idiomatic API is `Scorer.Algorithms()`
- **File:** /Users/johnny/Development/fuzzymatch/docs/scorer.md (lines 227-237)
- **Phase introduced:** Phase 8
- **Issue:** The "Consumers that need stable ordering" example builds a fresh `[]AlgoID`, runs `sort.Slice(...)`, then iterates. But `Scorer.Algorithms()` already returns the algorithm set in AlgoID-ascending order; a more idiomatic example would be:
  ```go
  scores := s.ScoreAll(a, b)
  for _, alg := range s.Algorithms() {
      fmt.Printf("%s: %.4f\n", alg.ID, scores[alg.ID])
  }
  ```
- **Standard:** API ergonomics — show the canonical idiom.
- **Action:** Code fix (the docs).

### [Important] `Match` boundary semantic (inclusive `>=`) is documented but the off-by-one case is not highlighted
- **File:** /Users/johnny/Development/fuzzymatch/docs/scorer.md (line 100, 388 of scorer.go)
- **Phase introduced:** Phase 8
- **Issue:** "True when `Score(a, b) >= Threshold()` (boundary inclusive)" — correct but easy to miss. Consumers calibrating with `WithThreshold(0.85)` and getting score 0.84999999... will see Match=false; consumers getting 0.8500000... will see Match=true. The `>=` choice is a meaningful design decision worth one extra sentence: "If you want strict-greater-than semantics, choose your threshold one ulp lower (or — equivalently — set the threshold to the lowest acceptable score, not the upper bound of unacceptable scores)."
- **Standard:** Documentation clarity for the boundary case.
- **Action:** Code fix (the docs).

### [Important] No callout on Hamming length-mismatch silent-zero policy
- **File:** /Users/johnny/Development/fuzzymatch/docs/algorithms.md (Hamming section is TBD)
- **Phase introduced:** Phase 2
- **Issue:** `HammingScore("abc", "ab") = 0.0` silently (no error, no panic — per the LOCKED policy in `hamming.go`). The Scorer-using consumer who feeds Hamming into a Scorer composition with mixed-length identifier pairs will see Hamming dragging the composite score to zero — and won't know why. This is documented inside Go file comments (`hamming.go` file-level godoc) and in `example_test.go ExampleHammingScore`, but NOT in `docs/algorithms.md` (which is TBD) or in `docs/scorer.md` (which is where Scorer consumers look).
- **Standard:** Consumer pitfall surfacing.
- **Action:** Code fix (the docs). Fill `docs/algorithms.md#hamming` with the length-mismatch policy front and centre.

### [Important] `DefaultScorer` composition stability promise is unstated
- **File:** /Users/johnny/Development/fuzzymatch/docs/scorer.md (lines 132-158)
- **Phase introduced:** Phase 8
- **Issue:** The "Default Composition" table lists the six algorithms. But it does not state whether the six-algorithm set is the v1.x contract or whether a future minor release may add/remove algorithms. The threshold `0.85` is documented as "calibrated for this specific mix" — but if the mix changes, what's the migration story? `docs/extending.md` line 5 mentions "the curated 23-algorithm catalogue is the public v1.x contract" but says nothing about the DEFAULT composition.
- **Standard:** API stability documentation.
- **Action:** Code fix (the docs). Add a "Stability" subsection to "Default Composition" explicitly stating "the six-algorithm composition and the 0.85 threshold are part of the v1.x DefaultScorer contract" (or "are subject to minor-version revision; pin a custom Scorer if reproducibility matters").

### [Important] Scorer "minus DM" example score table includes a counter-intuitive row not explained
- **File:** /Users/johnny/Development/fuzzymatch/examples/scorer-composition/main_test.go (line 49)
- **Phase introduced:** Phase 8
- **Issue:** The committed `want` constant shows `org_id / organisation_id` with Default=0.2911 and MinusDM=0.3493 — i.e. removing DoubleMetaphone INCREASED the composite score. The narrative in `main.go` says the threshold drops from 0.85 → 0.80 because "removing one of six signals reduces the composite ceiling" — but that's exactly contradicted for this row. A short comment in the example noting "for inputs where DoubleMetaphone returned 0 because phonetic codes disagreed, removing it actually raises the composite" would prevent reader confusion.
- **Standard:** Example clarity.
- **Action:** Code fix (the example or the example's comment block).

### [Important] No documented mapping from AlgoID → display name for ScoreAll output
- **File:** /Users/johnny/Development/fuzzymatch/docs/scorer.md (ScoreAll section)
- **Phase introduced:** Phase 8
- **Issue:** `ScoreAll` returns `map[AlgoID]float64`. The example at lines 234-235 uses `fmt.Printf("%s: %.4f\n", id, scores[id])` — and AlgoID does implement `fmt.Stringer` so this works. But the doc doesn't explicitly note "AlgoID.String() returns the canonical display name (e.g. AlgoLevenshtein → 'Levenshtein')". A consumer who tries `fmt.Printf("%d: ...", id)` would get `0: ...` and be puzzled.
- **Standard:** Documentation precision.
- **Action:** Code fix (the docs).

---

## Improvement

### [Improvement] llms.txt status section is out of date (lists only Phase 1 as shipped)
- **File:** /Users/johnny/Development/fuzzymatch/llms.txt (lines 9-14)
- **Phase introduced:** Phase 1
- **Issue:** "Phase 1 ships the foundation primitives: AlgoID dispatch enum, sentinel errors, Normalise, Tokenise. Phase 2+ adds the 23 algorithms, then Scorer (Phase 8) and scan / Extract (Phases 9–10)." This implies Phase 2 has not shipped, but the Public API section below it lists ALL 23 algorithms + Scorer. Internally inconsistent.
- **Standard:** Doc accuracy.
- **Action:** Code fix (the docs).

### [Improvement] llms.txt Scorer section header still references plan numbers
- **File:** /Users/johnny/Development/fuzzymatch/llms.txt (line 187)
- **Phase introduced:** Phase 8
- **Issue:** "### Scorer construction options (Phase 8 — plan 08-01 lays the option layer; plan 08-02 lands NewScorer + Score + Match; plan 08-03 lands ScoreAll + ..." — plan-internal language leaking into the consumer-facing doc.
- **Standard:** Consumer-facing docs should not reference internal plan numbers.
- **Action:** Code fix (the docs).

### [Improvement] llms.txt and README do not mention `examples/` directories
- **File:** /Users/johnny/Development/fuzzymatch/llms.txt; /Users/johnny/Development/fuzzymatch/README.md
- **Phase introduced:** Phase 7
- **Issue:** Already raised above under Critical; mentioned here as the llms.txt has its own Documentation block that misses the examples.
- **Action:** Code fix.

### [Improvement] Hamming algorithm "name" — README catalogue says "Hamming 1950" but per the locked policy it diverges from strict Hamming for unequal-length inputs
- **File:** /Users/johnny/Development/fuzzymatch/README.md (line 132)
- **Phase introduced:** Phase 2
- **Issue:** Strict Hamming-1950 is defined only for equal-length inputs. The library's `HammingScore` returns 0.0 for unequal-length inputs by policy (LOCKED, see hamming.go). The README catalogue entry "Hamming | AlgoHamming | Hamming 1950" doesn't hint at the divergence. Sophisticated readers will not be misled; novice readers will be.
- **Standard:** Algorithm-correctness documentation.
- **Action:** Code fix (the docs).

### [Improvement] README catalogue "Token Sort Ratio / Set Ratio / Partial Ratio" cite "SeatGeek fuzzywuzzy / RapidFuzz" but the actual correctness baseline is RapidFuzz only
- **File:** /Users/johnny/Development/fuzzymatch/README.md (lines 153-155)
- **Phase introduced:** Phase 6
- **Issue:** `docs/cross-validation.md` explicitly states "We cross-validate against RapidFuzz exclusively; fuzzywuzzy is referenced only as historical context." README citation mentions both as if they were equal sources. Minor but consistent across all three rows.
- **Standard:** Algorithm-correctness — cite the actual baseline.
- **Action:** Code fix (the docs).

### [Improvement] README "Three layers" diagram does not show Layer 1 with a complete signature
- **File:** /Users/johnny/Development/fuzzymatch/README.md (lines 80-84)
- **Phase introduced:** Phase 1
- **Issue:** `Layer 1: Algorithm functions      LevenshteinScore(a, b)` — but the function signature is `LevenshteinScore(a, b string) float64`. Showing the return type would convey "you get a float back, easy" immediately.
- **Action:** Code fix (the docs).

### [Improvement] Algorithm catalogue cross-references `AlgoID` constants but doesn't tell the reader what an AlgoID is
- **File:** /Users/johnny/Development/fuzzymatch/README.md (lines 125-171)
- **Phase introduced:** Phase 1
- **Issue:** The catalogue table column "`AlgoID`" shows constants like `AlgoLevenshtein`. A first-time reader has no context — is this an iota-int enum? Is it the function name? The Scorer composition section later in the README is where these constants are used, but that section is missing entirely.
- **Action:** Code fix (the docs).

### [Improvement] README "Algorithm catalogue" omits per-algorithm "intended use" column
- **File:** /Users/johnny/Development/fuzzymatch/README.md (lines 125-171)
- **Phase introduced:** Phase 1
- **Issue:** The catalogue table has columns: Algorithm | AlgoID | Primary source | Detail. The documentation-standards skill specifies "algorithm name, primary source citation, intended use, link to per-algorithm detail in `docs/algorithms.md`." The "intended use" column is missing. Without it, the reader can't make algorithm-selection decisions from the catalogue alone.
- **Standard:** documentation-standards.md "Algorithm catalogue table".
- **Action:** Code fix (the docs).

### [Improvement] README does not have a "Scorer composition" section as required by documentation-standards
- **File:** /Users/johnny/Development/fuzzymatch/README.md
- **Phase introduced:** Phase 8
- **Issue:** documentation-standards.md (lines 28-29) mandates "Scorer composition section with a typical mixed-algorithm example" and "Scan sub-package introduction with a one-screen example". Neither appears in current README. There's a brief "Configuration" section pointing at Phase-1 primitives only.
- **Standard:** documentation-standards.md README structure.
- **Action:** Code fix.

### [Improvement] README "Configuration" section lists Phase-1 option fields only
- **File:** /Users/johnny/Development/fuzzymatch/README.md (lines 178-198)
- **Phase introduced:** Phase 1
- **Issue:** The Configuration section shows `NormalisationOptions` and `TokeniseOptions` field lists but never mentions `ScorerOption` even though Phase 8 has shipped. Should at least preview the Scorer With* options.
- **Action:** Code fix.

### [Improvement] No "What's new in v0.X" status pointer pointing at CHANGELOG
- **File:** /Users/johnny/Development/fuzzymatch/README.md
- **Phase introduced:** Phase 8
- **Issue:** README links to CHANGELOG only indirectly (via CONTRIBUTING.md's PR-checklist). A "Current state" subsection in README pointing at CHANGELOG would help consumers tracking the pre-release progress.
- **Action:** Code fix (the docs).

### [Improvement] FAQ does not cover "how do I pick a threshold"
- **File:** /Users/johnny/Development/fuzzymatch/docs/faq.md
- **Phase introduced:** Phase 1
- **Issue:** FAQ has good entries for why-no-NW, why-no-Metaphone-3, why-no-embeddings, why-phonetic-binary, why-not-generic, why-x/text. Missing: "What threshold should I start with?" "Why is my Match() always false?" (NaN trap from CR-01 once fixed). "Why does my score drop when I add more algorithms?" (weight normalisation surprise). "Why does WithoutNormalisation not actually remove all character transformations?" (token internal Tokenise call).
- **Standard:** documentation-standards.md "docs/faq.md — common questions including..."
- **Action:** Code fix (the docs).

### [Improvement] `docs/scorer.md` Quickstart code does not show the printed output
- **File:** /Users/johnny/Development/fuzzymatch/docs/scorer.md (lines 21-32)
- **Phase introduced:** Phase 8
- **Issue:** The example calls `s.Match("user_id", "userId")` and just notes "// similar" in a `if` branch. A consumer wants to see the actual return: `true` (because the Normalise pipeline maps both to "user id" which is identical pre-tokenise). Showing `// → true` would close the loop.
- **Action:** Code fix (the docs).

### [Improvement] No documented Score result reproducibility-across-patches promise
- **File:** /Users/johnny/Development/fuzzymatch/README.md; /Users/johnny/Development/fuzzymatch/docs/scorer.md
- **Phase introduced:** Phase 8
- **Issue:** README says "Cross-platform deterministic output — verified byte-identical across linux/amd64, linux/arm64, darwin/amd64, darwin/arm64, windows/amd64." But it does not say whether v0.4.0 → v0.4.1 will produce the same float64 result. Consumers persisting scores need to know the patch-version reproducibility story. The internal spec (requirements.md) defines this but consumers won't read that.
- **Standard:** API stability documentation.
- **Action:** Code fix (the docs).

### [Improvement] No glossary or terminology table
- **File:** /Users/johnny/Development/fuzzymatch/docs/
- **Phase introduced:** Phase 1
- **Issue:** Documentation uses "score" (always [0.0, 1.0] composite), "raw score" (algorithm-specific output), "weight" (consumer-supplied input), "normalised weight" (post auto-normalisation), "threshold" (Match boundary), "algorithm" (one of 23), "composition" (set of weighted algorithms inside a Scorer), "Scorer" (the object), "scan" (Phase 9 sub-package). A short glossary in `docs/` would prevent confusion. Today the consumer has to pick up the terminology from prose, where it's sometimes inconsistent ("metric" vs "algorithm" both appear in some places).
- **Action:** Code fix (add `docs/glossary.md` or a glossary section in README).

### [Improvement] Phonetic example's MRACompare output requires reader to know NBS Tech Note 943's threshold table
- **File:** /Users/johnny/Development/fuzzymatch/examples/phonetic-keys/main.go (lines 77-83)
- **Phase introduced:** Phase 7
- **Issue:** Output is `Byrne vs Boern: matched=true sim=5`. A reader without familiarity with NBS Tech Note 943 sees `sim=5` and has no calibration for whether 5 is high/low. A comment block above the loop explaining "MRA returns matched=true when sim>=threshold per the NBS table; sim is a 0-6 counter" would orient new readers.
- **Action:** Code fix (the example).

### [Improvement] No "common pitfalls" section in any consumer doc
- **File:** /Users/johnny/Development/fuzzymatch/docs/
- **Phase introduced:** Phase 8
- **Issue:** Several footguns identified above (NaN threshold, WithoutNormalisation tokeniser surprise, Tversky α+β=0, Hamming length mismatch, ScoreAll iteration non-determinism). A dedicated "Common Pitfalls" page would gather these — current state is one-pitfall-per-doc, scattered.
- **Action:** Code fix (the docs).

### [Improvement] No "from this library to production" checklist (pin Scorer at package level, etc.)
- **File:** /Users/johnny/Development/fuzzymatch/docs/tuning.md (Pinning a calibrated configuration section)
- **Phase introduced:** Phase 8
- **Issue:** `docs/tuning.md` Pinning section is good but is the only place this guidance lives. A production-readiness checklist would help: "pin Scorer at package level; check `errors.Is` against named sentinels; surface ScoreAll only for tuning, not for production lookups; verify determinism golden file in CI; ...". Today the consumer has to infer this from scattered pieces.
- **Action:** Code fix (the docs).

### [Improvement] CONTRIBUTING.md does not mention runnable godoc examples on pkg.go.dev
- **File:** /Users/johnny/Development/fuzzymatch/CONTRIBUTING.md
- **Phase introduced:** Phase 2
- **Issue:** The contributing doc covers conventional commits, CLA, make-target list — but doesn't emphasise that every public function deserves a runnable godoc example (the 39 existing examples are the precedent). A line in the pre-PR checklist would surface this.
- **Action:** Code fix (the docs).

---

## Notes on Phase-9 / Phase-10 docs

`docs/scan.md` (Phase 9) and the Extract API (Phase 10) are deliberately scaffold. These findings cover the scaffolds AS scaffolds — they should not block Phase 9. When Phase 9 lands, re-run user-guide-reviewer over the actual content.

## File paths (all absolute)

- /Users/johnny/Development/fuzzymatch/README.md
- /Users/johnny/Development/fuzzymatch/CHANGELOG.md
- /Users/johnny/Development/fuzzymatch/CONTRIBUTING.md
- /Users/johnny/Development/fuzzymatch/doc.go
- /Users/johnny/Development/fuzzymatch/example_test.go
- /Users/johnny/Development/fuzzymatch/llms.txt
- /Users/johnny/Development/fuzzymatch/llms-full.txt
- /Users/johnny/Development/fuzzymatch/scorer.go
- /Users/johnny/Development/fuzzymatch/scorer_options.go
- /Users/johnny/Development/fuzzymatch/errors.go
- /Users/johnny/Development/fuzzymatch/hamming.go
- /Users/johnny/Development/fuzzymatch/levenshtein.go
- /Users/johnny/Development/fuzzymatch/normalise.go
- /Users/johnny/Development/fuzzymatch/tokenise.go
- /Users/johnny/Development/fuzzymatch/tversky.go
- /Users/johnny/Development/fuzzymatch/docs/algorithms.md
- /Users/johnny/Development/fuzzymatch/docs/cross-validation.md
- /Users/johnny/Development/fuzzymatch/docs/extending.md
- /Users/johnny/Development/fuzzymatch/docs/faq.md
- /Users/johnny/Development/fuzzymatch/docs/performance.md
- /Users/johnny/Development/fuzzymatch/docs/scan.md
- /Users/johnny/Development/fuzzymatch/docs/scorer.md
- /Users/johnny/Development/fuzzymatch/docs/tuning.md
- /Users/johnny/Development/fuzzymatch/examples/identifier-similarity/main.go
- /Users/johnny/Development/fuzzymatch/examples/phonetic-keys/main.go
- /Users/johnny/Development/fuzzymatch/examples/scorer-composition/main.go
- /Users/johnny/Development/fuzzymatch/examples/scorer-composition/main_test.go
- /Users/johnny/Development/fuzzymatch/.planning/phases/08-composite-scorer/08-REVIEW.md (cross-reference for CR-01)
- /Users/johnny/Development/fuzzymatch/.claude/skills/documentation-standards/SKILL.md (governing standard)
