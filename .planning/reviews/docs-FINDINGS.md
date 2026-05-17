---
status: issues_found
agent: docs-writer
scope: all documentation + godoc (phases 1-8)
reviewed: 2026-05-17T00:00:00Z
finding_counts:
  critical: 5
  important: 14
  improvement: 9
  total: 28
---

# Documentation Review Findings — fuzzymatch (Phases 1–8)

---

### [Critical] docs/requirements.md §7.5.1 contradicts implemented Ratcliff-Obershelp asymmetry

- **File:** `/Users/johnny/Development/fuzzymatch/docs/requirements.md` (line 713)
- **Phase introduced:** Phase 4 (algorithm spec) / Phase 4 (OQ-1 resolution locked 2026-05-14)
- **Issue:** Requirements §7.5.1 states "Mathematical invariants: identity, **symmetry**, range bounds." The implementation is intentionally asymmetric by OQ-1 resolution (locked 2026-05-14) to match Python `difflib.SequenceMatcher(autojunk=False).ratio()`. The symmetry claim in the spec is false for the shipped code. Additionally, §15 line 1204 lists `PropAlgorithmScore_Symmetric` as excluding "Monge-Elkan and asymmetric Tversky" but does NOT exclude Ratcliff-Obershelp, making the property-test exclusion invisible to a reader of the spec.
- **Standard:** documentation-standards SKILL.md "docs/ structure" — algorithm documentation must correctly state mathematical invariants; algorithm-correctness-standards SKILL.md §"Mathematical Invariants / Symmetric algorithms"
- **Action:** Code fix (update docs/requirements.md §7.5.1 to remove "symmetry" from the invariant list and add "asymmetric by OQ-1 design decision — see ratcliff_obershelp.go"). Also update §15 line 1204 to exclude `AlgoRatcliffObershelp` from the symmetric property list.
- **Rationale:** A consumer or reviewer reading §7.5.1 will incorrectly conclude `RatcliffObershelpScore(a,b) == RatcliffObershelpScore(b,a)`. The implementation explicitly rejects this and the prop tests intentionally omit the symmetry check. The spec contradiction will mislead future algorithm-correctness reviewers and consumer integrators.

---

### [Critical] docs/requirements.md §7.1.4 specifies `HammingDistance` returns `(int, error)` but implementation returns `int`

- **File:** `/Users/johnny/Development/fuzzymatch/docs/requirements.md` (lines 362–363)
- **Phase introduced:** Phase 2 (Hamming implementation)
- **Issue:** The requirements spec at §7.1.4 specifies `HammingDistance(a, b string) (int, error)` returning `ErrHammingLengthMismatch` on unequal lengths, and `HammingDistanceRunes(a, b string) (int, error)`. The shipped implementation (`hamming.go`) has `HammingDistance(a, b string) int` (no error return) and a documented "return max(len(a), len(b)) on mismatch" policy instead. `ErrHammingLengthMismatch` is never declared — errors.go only references it in a comment as "deferred". The same discrepancy appears in `llms-full.txt` (no Hamming function signatures documented separately, gap in coverage), but `llms.txt` correctly reflects the implemented signature `int`. The requirements spec is stale and contradicts the code.
- **Standard:** documentation-standards SKILL.md "docs/ structure" — spec must match implementation; algorithm-correctness-standards SKILL.md §"Score Normalisation / Edge cases"
- **Action:** Code fix — update docs/requirements.md §7.1.4 to match the locked implementation: `HammingDistance(a, b string) int` (silent max-length policy, no error return); remove the `ErrHammingLengthMismatch` reference from that section. Remove the parenthetical in errors.go line 31 referencing this deferred sentinel. The `api-ergonomics-reviewer` locked the silent-zero policy and this is well-documented in the implementation godoc — the spec just has not caught up.
- **Rationale:** Any future contributor or reviewer reading §7.1.4 will expect a `(int, error)` return and will be confused by the mismatch. The algorithm-correctness-reviewer gate will flag "signature does not match spec" on the next algorithm PR.

---

### [Critical] `monge_elkan.go` and `llms-full.txt` reference non-existent `ErrInvalidAlgoID` sentinel

- **File:** `/Users/johnny/Development/fuzzymatch/monge_elkan.go` (lines 162, 342); `/Users/johnny/Development/fuzzymatch/llms-full.txt` (line 652)
- **Phase introduced:** Phase 6 (Monge-Elkan)
- **Issue:** Two godoc comments in `monge_elkan.go` state that `WithMongeElkanAlgorithm` returns `ErrInvalidAlgoID`. The actual implementation in `scorer_options.go` (line 431, 435) returns `ErrInvalidAlgorithm`. `ErrInvalidAlgoID` is never declared anywhere in the codebase. `llms-full.txt` line 652 repeats the incorrect name. Any consumer reading the godoc will discriminate via `errors.Is(err, ErrInvalidAlgoID)` — which will never match because the sentinel does not exist; the actual returned sentinel is `ErrInvalidAlgorithm`.
- **Standard:** documentation-standards SKILL.md §"Godoc" — every exported function's godoc must accurately describe its return values
- **Action:** Code fix — in `monge_elkan.go` lines 162 and 342, replace `ErrInvalidAlgoID` with `ErrInvalidAlgorithm`. In `llms-full.txt` line 652, replace `ErrInvalidAlgoID` with `ErrInvalidAlgorithm`.
- **Rationale:** A consumer whose code contains `errors.Is(err, fuzzymatch.ErrInvalidAlgoID)` will receive a compilation error because the symbol does not exist. This is a directly misleading godoc promise.

---

### [Critical] `docs/algorithms.md` is entirely scaffold — all 23 algorithms still say "Primary source: TBD" and "Status: planned" despite all being shipped

- **File:** `/Users/johnny/Development/fuzzymatch/docs/algorithms.md`
- **Phase introduced:** Phase 1 (scaffold created); applicable to all phases 2–8 (each phase shipped algorithms without updating this document)
- **Issue:** All 23 algorithm entries in `docs/algorithms.md` have `Primary source: TBD — filled in by the implementing phase` and `Status: planned (Phase N)`. Phases 2 through 8 have all shipped. None of the per-algorithm sections contain the mandatory fields from the documentation-standards skill: Algorithm name (canonical form), Description (1–3 paragraphs in prose), Mathematical formulation, Complexity, Score normalisation rule, Mathematical invariants, Edge cases, Reference vectors, Intended use cases, Comparable references. The document is a 239-line scaffold with no substantive per-algorithm content.
- **Standard:** documentation-standards SKILL.md §"Algorithm Documentation" — every field listed is mandatory; system-level CLAUDE.md "Algorithm Documentation" section repeats the requirement verbatim
- **Action:** Code fix — fill in all 23 algorithm entries with the mandatory fields. The source material exists in `docs/requirements.md` §7, the implementation file godoc blocks, and the test files. This is the highest-priority single documentation task remaining.
- **Rationale:** `docs/algorithms.md` is the per-algorithm detail document linked from the README catalogue table. Every link in the README (`[docs/algorithms.md#levenshtein]` etc.) lands on a stub. Consumers, contributors, and the algorithm-correctness-reviewer all expect full detail here.

---

### [Critical] README Quick Start example is Phase-1 primitives only — stale relative to phases 2–8 shipping

- **File:** `/Users/johnny/Development/fuzzymatch/README.md` (lines 96–118)
- **Phase introduced:** Phase 1 (Quick Start written for Phase-1 scope); Phases 2–8 did not update it
- **Issue:** The Quick Start section explicitly says "the full algorithm-driven quick start is added with Phase 2" and shows only `Normalise` and `Tokenise`. All 23 algorithms and the full Scorer are now shipped. The README prominently tells consumers that `LevenshteinScore` "lands in Phase 2" — which is past. The documentation-standards skill requires the README Quick Start to contain "a complete, working program" including `DefaultScorer + custom Scorer + scan`. The current example does not demonstrate a single algorithm function or the Scorer.
- **Standard:** documentation-standards SKILL.md §"README" — "First code block shows a complete, working program (the 'headline example' verified by `readme_shop_front_test.go`)"; system-level CLAUDE.md §"README Structure" item 8 — "Quick Start with runnable example (DefaultScorer + custom Scorer + scan)"
- **Action:** Code fix — replace the Phase-1-only Quick Start with a multi-block example: one block showing `LevenshteinScore` + `JaroWinklerScore`, one block showing `DefaultScorer().Score()`, one block showing `NewScorer` custom composition. Remove the "Phase 2" note.
- **Rationale:** Any first-time consumer reading the README sees a Quick Start that shows primitives, not the algorithms. The readme_shop_front_test.go meta-test should be verifying the headline example compiles and produces documented output — with a Phase-1-only example that test is not exercising the primary API surface.

---

### [Important] `doc.go` platform list omits `darwin/amd64` — lists only 4 of 5 platforms

- **File:** `/Users/johnny/Development/fuzzymatch/doc.go` (line 40–41)
- **Phase introduced:** Phase 1
- **Issue:** The package godoc says "Output is byte-identical across linux/amd64, linux/arm64, darwin/arm64, and windows/amd64." This omits `darwin/amd64`. The spec (docs/requirements.md §13.3 line 1184), CLAUDE.md, README.md, and llms-full.txt (line 1303) all list five platforms including `darwin/amd64`. Only `doc.go` is missing one entry.
- **Standard:** documentation-standards SKILL.md §"Godoc" — package overview must be accurate; docs must be internally consistent
- **Action:** Code fix — add `darwin/amd64` to the platform list in `doc.go` line 40: "linux/amd64, linux/arm64, darwin/amd64, darwin/arm64, and windows/amd64"
- **Rationale:** Minor inaccuracy but since this appears in the published godoc on pkg.go.dev it directly misleads developers on Apple Intel machines who may wonder whether the library is tested on their platform.

---

### [Important] `algorithm-correctness-standards` SKILL.md does not list Ratcliff-Obershelp as an asymmetric exception in the symmetry section

- **File:** `/Users/johnny/Development/fuzzymatch/.claude/skills/algorithm-correctness-standards/SKILL.md` (line 119)
- **Phase introduced:** Phase 1 (skill written); Phase 4 (OQ-1 resolution locked)
- **Issue:** Line 119 states "All algorithms in this library are symmetric **except** Monge-Elkan (when used asymmetrically) and Tversky (when α ≠ β)." Ratcliff-Obershelp is not listed. However the OQ-1 resolution (locked 2026-05-14) makes Ratcliff-Obershelp intentionally asymmetric. The skill is used by the algorithm-correctness-reviewer as the gating standard — future reviewers following the skill will incorrectly require a symmetry property test for Ratcliff-Obershelp.
- **Standard:** No governing standard — this is a skill accuracy issue
- **Action:** Skill clarification — add Ratcliff-Obershelp to the exception list: "All algorithms in this library are symmetric **except** Monge-Elkan (when used asymmetrically), Tversky (when α ≠ β), and Ratcliff-Obershelp (intentionally asymmetric per OQ-1 RESOLUTION LOCKED 2026-05-14 — see ratcliff_obershelp.go)."
- **Rationale:** The skill is the authoritative reference for the algorithm-correctness-reviewer. An incorrect skill produces incorrect review decisions.

---

### [Important] README is missing required structural elements from the spec and CLAUDE.md

- **File:** `/Users/johnny/Development/fuzzymatch/README.md`
- **Phase introduced:** Phase 1
- **Issue:** Comparing the current README sections against `docs/requirements.md §16.1` and `system-level CLAUDE.md §"README Structure"`, the following required elements are absent: (1) a logo image — requirements §16.1 item 1 specifies `.github/images/logo-readme.png`; the `.github/images/` directory does not exist; (2) a "Quick links" section (requirements item 4 — links to Quick Start, Features, Algorithms, Docs, API Reference); (3) an emoji-headed section "🔍 Overview" (requirements item 7 — currently the README has "What this is" without an emoji); (4) "✨ Key Features" is present but uses no emoji header; (5) no dedicated "🛠 Scorer composition" section (requirements item 12 / system-level CLAUDE.md item 10); (6) no "🎯 Tuning Guidance" section with link to tuning.md (requirements item 15 / system-level CLAUDE.md item 13) — tuning.md is mentioned in Configuration but there is no dedicated section; (7) a CI badge is missing — CLAUDE.md "Docs Stack" specifies badges "(CI, Go Reference, Go Report Card, License, Status pre-release)"; only 4 badges are present (Go Reference, Go Report Card, License, Status); CI badge is absent.
- **Standard:** documentation-standards SKILL.md §"README"; system-level CLAUDE.md §"README Structure"; docs/requirements.md §16.1
- **Action:** Code fix — add missing structural elements. For the CI badge, add the GitHub Actions badge for the `ci.yml` workflow. The logo image can be deferred until artwork is ready but the placeholder reference in requirements.md should match the actual state (either remove the requirement note or add a placeholder). Add the Quick Links section, emoji headings, dedicated Scorer composition section, and Tuning Guidance section.
- **Rationale:** The README is the library's public front door. The documented structure mirrors `axonops/mask` style — deviating from it makes the project feel inconsistent with its own stated intent.

---

### [Important] README `scorer.md` reference says "once Phase 8 lands" — Phase 8 has shipped

- **File:** `/Users/johnny/Development/fuzzymatch/README.md` (lines 198, 224)
- **Phase introduced:** Phase 8 (text added before Phase 8 shipped; never updated)
- **Issue:** Line 198: "see [`docs/scorer.md`] for the `Scorer` API once Phase 8 lands." Line 224: "`docs/scorer.md` — `Scorer` configuration and tuning (Phase 8)." The Scorer shipped in Phase 8. These lines should be updated to present tense. Additional stale Phase references: line 96 (Quick Start note about Phase 2), line 57 ("Phase 8" parenthetical on the Scorer feature), line 58 ("Phase 9" on scan), line 204 ("from Phase 2" algorithm function note), line 206 ("Phase 8" Scorer immutability), line 123 ("once it lands (Phase 2+)").
- **Standard:** documentation-standards SKILL.md §"README" — docs must reflect current state
- **Action:** Code fix — audit all Phase-N references and remove or update those that are now past history. Phrases like "once Phase 8 lands" should become present tense; "Phase 2+" should be removed where the algorithms are shipped; "(Phase 8)" labels on shipped features should be dropped.
- **Rationale:** Forward-looking scaffolding language left in a shipped product creates consumer confusion about what is actually available.

---

### [Important] `example_test.go` is missing required examples for Normalise, Tokenise, Scorer, and ScoreAll

- **File:** `/Users/johnny/Development/fuzzymatch/example_test.go`
- **Phase introduced:** Phase 8 (Scorer examples were the responsibility of Phase 8)
- **Issue:** `docs/requirements.md §16.5` specifies the following `Example*` functions must exist in `example_test.go`: `ExampleScorer` (DefaultScorer with a typical pair), `ExampleNewScorer` (custom Scorer composition), `ExampleScorer_ScoreAll` (per-algorithm breakdown), `ExampleScorer_Match` (threshold matching), `ExampleNormalise`, and `ExampleTokenise`. None of these six functions are present in `example_test.go`. The 39 examples that exist cover all 23 algorithm score functions and phonetic encoder functions, but the Scorer layer and the two Phase-1 primitive functions have no godoc examples visible on pkg.go.dev.
- **Standard:** documentation-standards SKILL.md §"Godoc" — "Include runnable `Example` functions in `example_test.go` for major use cases"; docs/requirements.md §16.5
- **Action:** Code fix — add `ExampleNormalise`, `ExampleTokenise`, `ExampleScorer` (using DefaultScorer), `ExampleNewScorer` (custom 2-algorithm composition with threshold), `ExampleScorer_ScoreAll`, and `ExampleScorer_Match` to `example_test.go`. All must have `// Output:` blocks so they are verified by `go test`.
- **Rationale:** Missing pkg.go.dev examples for the Scorer mean the primary consumer-facing API layer has no runnable documentation. This is the most common first point of contact for new users.

---

### [Important] CHANGELOG is a stub — no entries for any of the 23 implemented algorithms, Scorer, or cross-validation infrastructure

- **File:** `/Users/johnny/Development/fuzzymatch/CHANGELOG.md`
- **Phase introduced:** Phases 2–8 (no CHANGELOG entries added when features shipped)
- **Issue:** The CHANGELOG has a single `[Unreleased]` section with five bootstrap-only bullet points from Phase 1. All 23 algorithms (Phases 2–7), the composite Scorer (Phase 8), all property tests, fuzz tests, BDD scenarios, the cross-validation infrastructure (three corpora), and the examples directory have shipped with zero CHANGELOG entries. Keep a Changelog format requires added/changed/deprecated/removed/fixed/security subsections per release or per unreleased feature group. The current state violates `CONTRIBUTING.md §CHANGELOG` (line 27: "A CHANGELOG entry has been added under `## [Unreleased]`") on every one of the ~40 feature commits.
- **Standard:** documentation-standards SKILL.md §"docs/ structure" references Keep a Changelog; CONTRIBUTING.md §"Pre-PR Checklist" item 4; docs/requirements.md §16.6
- **Action:** Code fix — back-fill CHANGELOG entries for all shipped features. Minimum: one bullet per algorithm, one for the cross-validation corpora, one for the Scorer, one for the examples. Group by phase for clarity.
- **Rationale:** A consumer evaluating the library's history to understand what changed between notional releases will find an empty record. The goreleaser release workflow extracts the CHANGELOG section for the GitHub release description — a stub will produce an empty release body.

---

### [Important] `CONTRIBUTING.md` references `docs/requirements.md §11.2` which does not exist

- **File:** `/Users/johnny/Development/fuzzymatch/CONTRIBUTING.md` (line 186–187)
- **Phase introduced:** Phase 1
- **Issue:** CONTRIBUTING.md states "This is requirement REL-07 from [`docs/requirements.md`] §11.2." Section §11 of `docs/requirements.md` is titled "Phonetic Algorithm Integration" and has no §11.2 subsection. The algorithm deprecation policy described in CONTRIBUTING.md corresponds approximately to the note at requirements §13.1 ("Algorithm score stability") or the general text in §5 design principles — there is no §11.2. The linked section does not exist.
- **Standard:** documentation-standards SKILL.md — cross-references must be accurate
- **Action:** Code fix — change the section reference to the correct one. The closest match is `docs/requirements.md §13.1` (Algorithm score stability, lines 1174–1177) or optionally add a dedicated deprecation-policy subsection to requirements.md. Also note that `REL-07` exists only in `.planning/REQUIREMENTS.md`, not in `docs/requirements.md` — the reference to `REL-07` in CONTRIBUTING.md points to a planning artifact ID, not a user-facing requirement identifier.
- **Rationale:** A contributor following the CONTRIBUTING.md link will land on a non-existent section and lose confidence in the document's accuracy.

---

### [Important] `llms-full.txt` Phase 4 section omits Hamming function signatures entirely

- **File:** `/Users/johnny/Development/fuzzymatch/llms-full.txt`
- **Phase introduced:** Phase 2 (Hamming shipped); llms-full.txt Phase 4 section was authored later
- **Issue:** The `llms-full.txt` document has a "Phase 4 algorithm surface" section covering Strcmp95, LCSStr, and Ratcliff-Obershelp. Hamming, Jaro, Jaro-Winkler, Levenshtein, and DL-OSA/Full appear elsewhere but there is no section dedicated to Hamming's exact function signatures with godoc summaries (the format used for every other algorithm). The `llms.txt` concise index correctly lists `HammingDistance`, `HammingDistanceRunes`, `HammingScore`, `HammingScoreRunes` as `func ... int` or `float64` but `llms-full.txt` provides no per-function godoc summary for Hamming.
- **Standard:** documentation-standards SKILL.md §"llms.txt / llms-full.txt" — "llms-full.txt includes full API signatures, all algorithm citations, and worked examples"
- **Action:** Code fix — add a Phase 2 Hamming section to llms-full.txt mirroring the style of other algorithm sections (typed signatures with inline godoc summaries).
- **Rationale:** AI assistants reading llms-full.txt will find Hamming underdocumented relative to every other algorithm.

---

### [Important] `strcmp95.go` godoc block does not state that Strcmp95 is not a metric and does not satisfy the triangle inequality

- **File:** `/Users/johnny/Development/fuzzymatch/strcmp95.go`
- **Phase introduced:** Phase 4
- **Issue:** The algorithm-correctness-standards skill explicitly states: "Jaro, Jaro-Winkler, Strcmp95 do NOT satisfy triangle inequality and are NOT metrics. Document this in their godoc." Jaro (`jaro.go` line 60) documents "Jaro is NOT a metric: the triangle inequality does not hold". Jaro-Winkler (`jarowinkler.go` lines 58–64) documents "Jaro-Winkler is NOT a metric". Strcmp95 has no such statement anywhere in `strcmp95.go` — neither in the file-level comment block nor in the function godocs.
- **Standard:** algorithm-correctness-standards SKILL.md §"Mathematical Invariants / Distance-based algorithms" and §"Symmetric algorithms" — specifically: "Jaro, Jaro-Winkler, Strcmp95 do NOT satisfy triangle inequality and are NOT metrics. Document this in their godoc."
- **Action:** Code fix — add to the `strcmp95.go` file-level comment (after the API hierarchy section): "# Strcmp95 is NOT a metric. Triangle inequality does not hold. Inherits the non-metric property of Jaro-Winkler."
- **Rationale:** The skill mandates this documentation for all three algorithms. Consistency with jaro.go and jarowinkler.go is broken, and a reviewer following the skill would flag this as a BLOCKING gap.

---

### [Important] `docs/faq.md` is missing the mandated "Why no Soft-TFIDF?" question

- **File:** `/Users/johnny/Development/fuzzymatch/docs/faq.md`
- **Phase introduced:** Phase 1 (faq.md created with DX-06 mandated entries); subsequent phases did not add
- **Issue:** `docs/requirements.md §16.2` specifies faq.md must cover "why not Needleman-Wunsch", "why not Soft-TFIDF", "why not ML embeddings", "why was Metaphone 3 excluded". The current faq.md has Needleman-Wunsch, Metaphone 3, ML/embeddings, plus bonus questions (phonetic-as-binary, generics, x/text). The "Why no Soft-TFIDF?" entry is absent. Soft-TFIDF is explicitly out of scope per requirements §4 ("requires a consumer-supplied corpus frequency table, which conflicts with the library's stateless pure-function design").
- **Standard:** documentation-standards SKILL.md §"docs/ structure" — faq.md must address the listed exclusions; docs/requirements.md §16.2
- **Action:** Code fix — add a "Why no Soft-TFIDF?" FAQ entry citing the stateless pure-function design conflict documented in requirements §4.
- **Rationale:** Consumers from the IR/NLP community will ask about Soft-TFIDF. Its absence without explanation will prompt repeated identical questions.

---

### [Important] `prior-art-research.md` contains American English spellings in a British English project

- **File:** `/Users/johnny/Development/fuzzymatch/docs/prior-art-research.md` (lines 163, 324, 330)
- **Phase introduced:** Phase 1
- **Issue:** The file uses "normalization" (three occurrences: lines 163, 324, 330) rather than "normalisation". The documentation-standards skill specifies British English throughout (normalisation, behaviour, organisation, etc.). All other docs in the project use British English consistently.
- **Standard:** documentation-standards SKILL.md §"Style" — "British English (colour, behaviour, organisation, normalisation, optimisation)"
- **Action:** Code fix — replace all three "normalization" instances with "normalisation" in `docs/prior-art-research.md`.
- **Rationale:** Consistency in spelling conventions across all docs.

---

### [Important] `docs/extending.md` — the composing-phonetic-algorithms section references `docs/requirements.md §11` but that section covers Phonetic Integration specs, not consumer-facing composition patterns

- **File:** `/Users/johnny/Development/fuzzymatch/docs/extending.md` (line 33)
- **Phase introduced:** Phase 7
- **Issue:** The "Composing phonetic algorithms with edit distance" section says "See `docs/requirements.md` §11." Section 11 is the internal spec for phonetic algorithm integration (encoding rules, variant discipline, cross-validation). It is not a consumer-facing guide for composing phonetic codes with edit distance. The consumer-useful content is actually in the function godocs (`SoundexCode`, `DoubleMetaphoneKeys`, `NYSIISCode`, `MRACode`) and in the relevant implementation files. Pointing at the spec section is a developer-facing reference, not a user-guide reference.
- **Standard:** documentation-standards SKILL.md §"docs/ structure" — extending.md is a consumer guide, not a spec cross-reference
- **Action:** Code fix — replace the §11 cross-reference with a concrete code example showing `LevenshteinScore(SoundexCode(a), SoundexCode(b))` pattern, plus a note pointing at the function godocs for the encoder surfaces.
- **Rationale:** The extending.md section is currently a TBD stub with an unhelpful reference. A consumer wanting to compose phonetic codes with edit distance gets no actionable content.

---

### [Improvement] `docs/algorithms.md` H2 anchor IDs use camelCase (e.g. `#sorensendice`) but README links use inconsistent casing

- **File:** `/Users/johnny/Development/fuzzymatch/README.md` (lines 143–145); `/Users/johnny/Development/fuzzymatch/docs/algorithms.md` (H2 headings)
- **Phase introduced:** Phase 1
- **Issue:** The README algorithm catalogue table links use the H2 header text directly as anchors: `#qgramjaccard`, `#sorensendice`, `#cosine`, `#tversky`, etc. The corresponding H2 headings in algorithms.md are `## QGramJaccard`, `## SorensenDice`, `## Cosine`, `## Tversky` — which produce anchors `#qgramjaccard` (GitHub Markdown lowercases). This is consistent for the two-word entries but inconsistent with the file-level comment that says "H2 anchors below match the algorithm's canonical spelling (e.g. `#levenshtein`, `#damerau-levenshtein-osa`)". The README table links for `#damerau-levenshtein-osa` and `#damerau-levenshtein-full` use hyphens (correct for the heading text) while single-word algorithms use no separator. This is actually correct behaviour. However the README link for `#qgramjaccard` will not resolve if the H2 heading is changed to `## Q-Gram Jaccard` (the canonical form per CLAUDE.md). When algorithms.md is fully filled in (Critical finding above), the headings should be verified against all README links.
- **Standard:** No governing standard — this is a preemptive consistency flag
- **Action:** Discuss-phase needed — when algorithms.md is populated, verify all deep-link anchors in the README match the H2 headings. Consider using the canonical hyphenated form "Q-Gram Jaccard" for the H2 to match the spec terminology.
- **Rationale:** Broken anchor links are a common consequence of renaming headings during content population.

---

### [Improvement] README `Table of contents` uses non-emoji section names that diverge from the emoji section headers required by spec

- **File:** `/Users/johnny/Development/fuzzymatch/README.md` (lines 14–28)
- **Phase introduced:** Phase 1
- **Issue:** The Table of Contents links `[Status](#-status)`, `[What this is](#what-this-is)`, etc. The corresponding sections use `## ⚠ Status` (has emoji) and `## What this is` (no emoji). The spec requires emoji section headers (🔍 Overview, ✨ Key Features, ❓ Why fuzzymatch?, 🚀 Quick Start, 🛠 Scorer composition, 🧵 Thread Safety, etc.). The TOC links work for existing sections but will need updating when the missing sections are added. Also, `[What this is]` does not match the spec's "🔍 Overview" heading; `[Key features]` does not match "✨ Key Features"; `[Why this library exists]` does not match "❓ Why fuzzymatch?".
- **Standard:** docs/requirements.md §16.1 items 7–20; system-level CLAUDE.md §"README Structure"
- **Action:** Code fix (coordinate with the README restructuring from the Critical finding above) — standardise section headings to use the emoji + title from the spec; update TOC links accordingly.
- **Rationale:** The spec mandates emoji headers mirroring axonops/mask. The current README uses them inconsistently (only ⚠ Status and 🤖 For AI Assistants have emojis).

---

### [Improvement] `scorer.md` "Method Reference" table lists five methods but the prose says "All four methods"

- **File:** `/Users/johnny/Development/fuzzymatch/docs/scorer.md` (line 107)
- **Phase introduced:** Phase 8
- **Issue:** The method reference table (lines 98–104) lists five methods: `Score`, `Match`, `ScoreAll`, `Threshold`, and `Algorithms`. The prose immediately after (line 107) says "All four methods are pure functions on the `*Scorer` receiver." The count is incorrect: there are five, not four.
- **Standard:** No governing standard — accuracy of documentation prose
- **Action:** Code fix — change "All four methods" to "All five methods" on line 107.
- **Rationale:** Small but factually wrong — confusing to a consumer counting methods against the table.

---

### [Improvement] `scorer.md` "Concurrency" section says "four methods" (`Score`, `Match`, `ScoreAll`, `Threshold`, `Algorithms`) — same off-by-one as above

- **File:** `/Users/johnny/Development/fuzzymatch/docs/scorer.md` (line 258)
- **Phase introduced:** Phase 8
- **Issue:** The concurrency section says "All four methods (`Score`, `Match`, `ScoreAll`, `Threshold`, `Algorithms`) are safe to call concurrently." It then lists five methods. Same off-by-one count as the Method Reference section.
- **Standard:** No governing standard — accuracy of documentation prose
- **Action:** Code fix — change "All four methods" to "All five methods" on line 258.
- **Rationale:** Same as the finding above; these two sentences were written at different times and the count was not updated when the fifth method was added.

---

### [Improvement] `docs/cross-validation.md` is undocumented in `docs/requirements.md §16.2` — missing from the docs/ structure spec

- **File:** `/Users/johnny/Development/fuzzymatch/docs/cross-validation.md`; `/Users/johnny/Development/fuzzymatch/docs/requirements.md` §16.2 (line 1430–1438)
- **Phase introduced:** Phase 6 (cross-validation doc created)
- **Issue:** `docs/requirements.md §16.2` enumerates the expected docs/ files: algorithms.md, scorer.md, scan.md, tuning.md, extending.md, performance.md, faq.md. It does not mention `cross-validation.md`, which was added in Phase 6 as a reference document for the cross-validation corpora. The file exists and is linked from CONTRIBUTING.md. The spec does not anticipate it.
- **Standard:** documentation-standards SKILL.md §"docs/ structure" — should enumerate all docs/ files
- **Action:** Code fix — add `docs/cross-validation.md` to the §16.2 list in requirements.md with a one-line description: "cross-validation.md — corpora, pinned reference implementations, regeneration protocol, and variant-divergence tagging mechanism."
- **Rationale:** The requirements.md is the authoritative spec for the repository layout. An unlisted file creates ambiguity about whether it is intentional.

---

### [Improvement] `CHANGELOG.md` uses `### Notes` subsection which is not a Keep-a-Changelog convention

- **File:** `/Users/johnny/Development/fuzzymatch/CHANGELOG.md` (line 17)
- **Phase introduced:** Phase 1
- **Issue:** Keep a Changelog 1.1.0 (the format linked in the CHANGELOG) defines exactly six subsection types under each release: Added, Changed, Deprecated, Removed, Fixed, Security. The current `[Unreleased]` section includes `### Notes` — which is not a Keep-a-Changelog convention. The Note content ("This project is pre-release. The API is not stable until `v1.0.0`.") belongs in the `[Unreleased]` section preamble or as a notice at the top of the file, not as a `### Notes` subsection.
- **Standard:** docs/requirements.md §16.6 "CHANGELOG following Keep-a-Changelog format"
- **Action:** Code fix — move the pre-release note to a blockquote or freeform paragraph before the `### Added` subsection; remove the non-standard `### Notes` header.
- **Rationale:** Keep-a-Changelog tooling and goreleaser's changelog extractor may not handle non-standard subsection names gracefully.

---

### [Improvement] `docs/performance.md` "ASCII fast paths" section incorrectly omits Damerau-Levenshtein from the list

- **File:** `/Users/johnny/Development/fuzzymatch/docs/performance.md` (lines 47–52)
- **Phase introduced:** Phase 8 (performance.md scaffold written; not updated when DL algorithms shipped)
- **Issue:** The ASCII fast paths section lists "Levenshtein, Jaro, Jaro-Winkler, Hamming, etc." — a partially-illustrative list. All character-based algorithms (DL-OSA, DL-Full, SWG, LCSStr, Strcmp95) have ASCII fast paths documented in their implementation files. The "etc." placeholder means the section is inaccurate as a reference for consumers choosing algorithms by allocation profile. This section should either reference the per-algorithm benchmarks in `bench.txt` or enumerate all algorithms with ASCII fast paths.
- **Standard:** documentation-standards SKILL.md §"docs/ structure" — performance.md covers "ASCII fast paths" accurately
- **Action:** Code fix (deferred — can wait for the full performance.md population pass): at minimum, change "etc." to "and all character-based algorithms" to avoid implied completeness.
- **Rationale:** Minor accuracy gap; low impact while performance.md is still a scaffold but should be fixed before v1.0.0.

---

### [Improvement] `algoid.go` constant godocs for token-tier algorithms cite "Cohen, Ravikumar & Fienberg 2003 — SecondString" but `llms-full.txt` and README cite "SeatGeek fuzzywuzzy / RapidFuzz"

- **File:** `/Users/johnny/Development/fuzzymatch/algoid.go` (lines 134–147); `/Users/johnny/Development/fuzzymatch/llms.txt` (lines 152–156); `/Users/johnny/Development/fuzzymatch/llms-full.txt` (line 83–84)
- **Phase introduced:** Phase 1 (algoid.go constants) / Phase 6 (llms.txt updated)
- **Issue:** The AlgoID godoc comments in algoid.go cite "Cohen, Ravikumar & Fienberg 2003 — SecondString library reference" for TokenSortRatio, TokenSetRatio, PartialRatio, and TokenJaccard. The README algorithm catalogue table and llms.txt cite "SeatGeek fuzzywuzzy / RapidFuzz" for the first three and "Jaccard 1912" for TokenJaccard. The algorithm-correctness-standards skill and the implementation files (token_sort_ratio.go etc.) use the RapidFuzz lineage. The SecondString citation in algoid.go is inconsistent with the rest of the project's source attribution.
- **Standard:** algorithm-correctness-standards SKILL.md §"Primary Source Citation" — citation must be consistent across all references to the same algorithm
- **Action:** Code fix — update the AlgoID constant godocs in algoid.go for AlgoTokenSortRatio, AlgoTokenSetRatio, AlgoPartialRatio to cite "SeatGeek fuzzywuzzy (2014); modern reference RapidFuzz (Bachmann, M., 2020–)" rather than SecondString. For AlgoTokenJaccard, the citation should be "Jaccard, P. (1912)" to match the rest of the project.
- **Rationale:** Inconsistent citations make it unclear which source governs the algorithm's reference vectors and constants. The algorithm-correctness-reviewer uses citations to gate correctness — an inconsistent source in algoid.go will surface as a reviewer question.
