---
plan_id: 01-08
phase: 01-foundation-infrastructure
plan: 08
type: execute
wave: 8
depends_on:
  - 01-07
autonomous: true
objective: >
  Land the DX scaffolding so the project ships against a complete v0.x-ready
  contributor and consumer experience: README polish, CHANGELOG seed,
  llms.txt / llms-full.txt with sync meta-test, docs/* scaffolds (algorithms
  / scorer / scan / extending / tuning / performance / faq), CODEOWNERS,
  issue templates, PR template, SECURITY.md, CODE_OF_CONDUCT.md,
  CONTRIBUTING.md (covering make check, bench-compare local-driven story,
  signed commits, conventional commits, release-via-CI-only, and the
  algorithm deprecation policy from REL-07). The ai_friendly_test.go meta-
  test asserts every exported root-package symbol appears in llms.txt.
files_modified:
  - README.md
  - CHANGELOG.md
  - llms.txt
  - llms-full.txt
  - ai_friendly_test.go
  - makefile_targets_test.go
  - docs/algorithms.md
  - docs/scorer.md
  - docs/scan.md
  - docs/extending.md
  - docs/tuning.md
  - docs/performance.md
  - docs/faq.md
  - SECURITY.md
  - CODE_OF_CONDUCT.md
  - CONTRIBUTING.md
  - .github/CODEOWNERS
  - .github/ISSUE_TEMPLATE/bug.yml
  - .github/ISSUE_TEMPLATE/feature.yml
  - .github/ISSUE_TEMPLATE/algorithm-proposal.yml
  - .github/PULL_REQUEST_TEMPLATE.md
requirements:
  - DX-01
  - DX-03
  - DX-04
  - DX-06
  - DX-07
  - TEST-06
  - REL-06
  - REL-07
must_haves:
  truths:
    - "README mirrors mask-style polish: logo+badges + TOC + status + overview + algorithm catalogue + thread-safety note + AI-assistants pointer + contributing + security + licence"
    - "CHANGELOG follows Keep-a-Changelog with `## [Unreleased]` and a placeholder for the first release"
    - "`llms.txt` and `llms-full.txt` exist; `ai_friendly_test.go` asserts every exported root-package symbol is referenced in llms.txt (parsed via go/ast)"
    - "Docs scaffolds exist for algorithms.md, scorer.md, scan.md, extending.md, tuning.md, performance.md, faq.md — each with TBD markers signed off by future phases"
    - "`docs/faq.md` contains entries for: Why no Needleman-Wunsch?, Why no Metaphone 3?, Why no embeddings?, Why phonetic-as-binary in the Scorer?, Why aren't algorithm functions generic?, Why x/text but no other deps? (DX-06)"
    - "CODEOWNERS routes review to maintainers (initial state); issue templates (bug, feature, algorithm-proposal) and PR template exist"
    - "SECURITY.md documents the vulnerability disclosure path; CODE_OF_CONDUCT.md is Contributor Covenant 2.1"
    - "CONTRIBUTING.md covers: `make check` as the local pre-PR gate; `make bench-compare` local-driven story (D-09); signed commits if/when required; conventional commits; release-via-CI-only; algorithm deprecation policy (REL-07)"
    - "Every Makefile target in the canonical list is documented in CONTRIBUTING (per makefile_targets_test.go meta-test) OR carries a `## suppress:` comment (none expected — all 19 targets are documented)"
  artifacts:
    - path: README.md
      provides: Project front door with mask-style polish
    - path: llms.txt
      provides: Concise AI-friendly summary
    - path: llms-full.txt
      provides: Detailed AI-friendly reference
    - path: ai_friendly_test.go
      provides: Meta-test asserting llms.txt sync with `go/ast`-parsed exported symbols
    - path: makefile_targets_test.go
      provides: Meta-test asserting every Makefile target is documented in CONTRIBUTING
    - path: docs/algorithms.md
      provides: Per-algorithm scaffold (23 H2 sections — content TBD)
    - path: docs/scorer.md
      provides: Scorer scaffold (Phase 8 fills in)
    - path: docs/scan.md
      provides: Scan scaffold (Phase 9 fills in)
    - path: docs/extending.md
      provides: Extending guide scaffold
    - path: docs/tuning.md
      provides: Tuning guide scaffold
    - path: docs/performance.md
      provides: Performance numbers scaffold (Phase 2+ fills in)
    - path: docs/faq.md
      provides: 6+ FAQ entries (DX-06)
    - path: SECURITY.md
      provides: Vulnerability disclosure path
    - path: CODE_OF_CONDUCT.md
      provides: Contributor Covenant 2.1
    - path: CONTRIBUTING.md
      provides: Pre-PR checklist, make check, bench-compare story, deprecation policy
    - path: .github/CODEOWNERS
      provides: Review routing
    - path: .github/ISSUE_TEMPLATE/bug.yml
      provides: Bug-report template
    - path: .github/ISSUE_TEMPLATE/feature.yml
      provides: Feature-request template
    - path: .github/ISSUE_TEMPLATE/algorithm-proposal.yml
      provides: Algorithm-proposal template (drives algorithm-licensing-reviewer screen)
    - path: .github/PULL_REQUEST_TEMPLATE.md
      provides: PR template with Source Origin Statement section for algorithm PRs
  key_links:
    - from: ai_friendly_test.go
      to: llms.txt
      via: meta-test parses go/ast and asserts every exported root-package symbol is referenced in llms.txt
      pattern: "go/ast"
    - from: makefile_targets_test.go
      to: CONTRIBUTING.md + Makefile
      via: meta-test asserts target coverage bi-directionally
      pattern: "Makefile.*CONTRIBUTING"
    - from: README.md
      to: docs/algorithms.md
      via: algorithm catalogue table links to docs anchors
      pattern: "docs/algorithms.md"
    - from: CONTRIBUTING.md
      to: REL-07 algorithm deprecation policy
      via: explicit section documenting "within a major version, algorithms can be added but not removed; scoring changes require minor bump"
      pattern: "deprecation"
---

<objective>
Land the consumer-facing and contributor-facing documentation so the project
is publicly-discoverable, AI-friendly, contributor-friendly, and policy-
documented (deprecation rules, conventional commits, release-via-CI-only).

Purpose: README and CHANGELOG already exist (pre-plan); this plan polishes
README to mask-style standards and fills in everything else.

Output:
  - Mask-polished README (badges, TOC, full algorithm catalogue table)
  - CHANGELOG (already exists — no changes here unless required)
  - llms.txt + llms-full.txt + ai_friendly_test.go meta-test
  - 7 docs/* scaffold files
  - CODEOWNERS + 3 issue templates + PR template
  - SECURITY + CODE_OF_CONDUCT + CONTRIBUTING
  - makefile_targets_test.go meta-test
</objective>

<execution_context>
@$HOME/.claude/get-shit-done/workflows/execute-plan.md
@$HOME/.claude/get-shit-done/templates/summary.md
</execution_context>

<context>
@.planning/PROJECT.md
@.planning/REQUIREMENTS.md
@.planning/phases/01-foundation-infrastructure/01-CONTEXT.md
@CLAUDE.md
@.claude/skills/documentation-standards/SKILL.md
@.claude/skills/algorithm-licensing-standards/SKILL.md
@.claude/skills/commit-standards/SKILL.md
@.claude/skills/fuzzymatch-review-protocol/SKILL.md
@.claude/skills/go-testing-standards/SKILL.md
@README.md
@CHANGELOG.md
@LICENSE
@NOTICE
@docs/requirements.md
@Makefile
@algoid.go
@errors.go
@normalise.go
@tokenise.go
@golden_canonical.go

<interfaces>
Mask-style structural template (per CONTEXT.md established-patterns):
  - Logo + badges block (CI, Go Reference, Go Report Card, License, Status)
  - Table of contents (auto-generated by GitHub from headings, but include
    explicit anchors for top-level sections)
  - Status (pre-release orange — already in README)
  - Overview ("What this is" — already in README, polish for mask consistency)
  - Key features bulleted list
  - Why this library exists (vs alternatives)
  - Quick Start (5-line placeholder until Phase 2's first algorithm —
    leave a TODO marker pointing at Phase 2)
  - Algorithm catalogue table (23 rows, grouped by category — already
    partially in README, replace with full mask-style table linking each
    algorithm to its docs/algorithms.md anchor)
  - Thread safety note
  - Configuration link (Scorer.md once it exists)
  - Tuning link (tuning.md once it exists)
  - API reference (pkg.go.dev/github.com/axonops/fuzzymatch)
  - "For AI Assistants" pointer to llms.txt / llms-full.txt (replace the
    current "When... will live" placeholder)
  - Contributing
  - Security
  - Licence

The README emoji headers (per mask): "⚠ Status", "🤖 For AI Assistants" etc.
already present in README.md — preserve.

ai_friendly_test.go meta-test logic:
  1. Use `go/parser.ParseDir` + `go/ast` to walk every .go file in the root
     package, collecting names of exported types, functions, methods, vars,
     constants.
  2. Read llms.txt.
  3. For each exported symbol name, assert `strings.Contains(llmsTxt,
     symbolName)`.
  4. Log all missing symbols. Fail test if any are missing.
  Excluded from the check: identifiers that are obviously internal-only
  (e.g. test helper exports inside `golden_test.go` like `WriteGoldenFile`
  — if it's truly internal-only, mark it via a `// +build testonly` build
  tag or document as an explicit allowlist in ai_friendly_test.go).

makefile_targets_test.go meta-test logic:
  1. Parse Makefile, extract all `^[a-z][a-z-]*:` target names.
  2. Read CONTRIBUTING.md.
  3. For each target name, assert it appears in CONTRIBUTING.md OR carries
     a `## suppress: <reason>` comment in the Makefile.
  4. Reverse direction: every target mentioned in CONTRIBUTING.md must
     exist in Makefile.

docs/algorithms.md scaffold: H1 "Algorithm Catalogue" + 23 H2 sections, one
per algorithm in the spec order, each with placeholders:
  - **Algorithm**: <name>
  - **Category**: <character/q-gram/token/phonetic/gestalt>
  - **Primary source**: TBD (filled in by the implementing phase)
  - **Status**: not implemented (replaced with "implemented in vX.Y.Z" when
    Phase 2+ lands)
  - **AlgoID constant**: `Algo<Name>`
  Each H2 has an anchor matching the algorithm canonical spelling for
  cross-linking from README.

docs/faq.md DX-06 entries (mandatory):
  - Why no Needleman-Wunsch? — superseded by Smith-Waterman-Gotoh
  - Why no Metaphone 3? — patent encumbrance (USP 7,440,941); cite
    docs/requirements.md §4 and algorithm-licensing-standards SKILL.md
  - Why no embeddings? — out of scope for a pure-function library
  - Why phonetic-as-binary in the Scorer? — phonetic algorithms produce
    discrete codes, not continuous similarity; binary 1.0/0.0 is the
    canonical normalisation per docs/requirements.md §7.20-§7.23
  - Why aren't algorithm functions generic? — Go generics introduce
    indirection costs incompatible with our zero-alloc fast-path targets;
    string is the lowest-common-denominator and the AlgoID dispatch is
    array-backed
  - Why x/text but no other deps? — Unicode NFC/NFD is table-stakes for
    audit-taxonomy use cases; inlining a maintained Unicode normalisation
    implementation would dwarf the rest of the library and require ongoing
    Unicode database updates; x/text is Go-team maintained, supply-chain
    auditable, narrowest possible scope

CONTRIBUTING.md sections:
  - Welcome / scope
  - Pre-PR checklist:
    1. `make check` green locally
    2. Sign the CLA (link to CLA.md once it exists — flag CLA.md as
       follow-up if not present)
    3. Conventional commit message (link to commit-standards skill;
       reminder about no AI attribution)
    4. Run `make bench-compare` if touching algorithm code; commit
       updated bench.txt only after reading the D-09 rationale
    5. Update CHANGELOG.md under `## [Unreleased]`
    6. PR template fields filled
  - Local development setup (Go 1.26.3+, install golangci-lint v2.12.2,
    govulncheck, goimports — links to STACK.md / docs/extending.md as
    references)
  - Make targets reference (every target with a one-line description)
  - Algorithm deprecation policy (REL-07):
    "Within a major version, algorithms may be ADDED but never REMOVED.
    Score-changing edits to an existing algorithm require a minor version
    bump and a CHANGELOG entry. Bug fixes that change scores are minor;
    intentional algorithm-formula changes are minor; algorithm removals
    are major."
  - Release process (CI-only — links to plan 01-03's release.yml):
    "Releases happen via CI on tag push only. Maintainers never run
    `git tag` or `goreleaser release` locally. See release.yml for the
    automated pipeline. Tags follow semver (vMAJOR.MINOR.PATCH); v0.x.y
    indicates pre-release."
  - Algorithm contribution flow (subset of CLAUDE.md "Workflow — Agent Gates"):
    1. Open an issue using the algorithm-proposal template
    2. algorithm-licensing-reviewer screens for patent encumbrance
    3. algorithm-correctness-reviewer verifies primary-source citation
       BEFORE implementation
    4. Implement against the citation; cross-validate with the literature
       reference vectors
    5. All 7 review gates pass before merge

SECURITY.md sections (per documentation-standards SKILL.md and mask):
  - Reporting a vulnerability: email security@axonops.com (or equivalent;
    record actual address used in SUMMARY)
  - Disclosure timeline: 90 days standard
  - Supported versions table (initially: v0.x — pre-release, no SLA)
  - Public security tooling reference: govulncheck weekly, gosec on every
    PR, CodeQL on every PR, Cosign signature verification command

CODE_OF_CONDUCT.md: Contributor Covenant 2.1 — verbatim from
https://www.contributor-covenant.org/version/2/1/code_of_conduct.txt with
the contact email substituted.

CODEOWNERS — initial routing: `* @<maintainer-github-username>`. The exact
maintainer username is filled in at execution time by the user; placeholder
documented in SUMMARY.

Issue templates (YAML form for GitHub native UI):
  - bug.yml: title prefix `bug:`, fields for reproduction steps,
    expected/actual behaviour, Go version, OS/arch, fuzzymatch version
  - feature.yml: title prefix `feat:`, fields for use case, proposed API
    shape, alternatives considered
  - algorithm-proposal.yml: title prefix `algo:`, fields for algorithm
    name, primary academic source citation, patent screen (yes/no/unknown),
    expected reference vectors, related existing algorithms in the catalogue

PR template (markdown): Source Origin Statement section template (per
algorithm-licensing-standards SKILL.md), summary, related issue, test plan,
CHANGELOG entry, reviewer checklist (5 named gates from CLAUDE.md).
</interfaces>
</context>

<tasks>

<task type="auto">
  <name>Task 1: Update README (mask polish), create llms.txt + llms-full.txt + ai_friendly_test.go</name>
  <files>README.md, llms.txt, llms-full.txt, ai_friendly_test.go</files>
  <read_first>
    - README.md (current state — preserve emoji headers + status block)
    - .planning/PROJECT.md ("What This Is", "Core Value", "Constraints", "Key Decisions")
    - .planning/phases/01-foundation-infrastructure/01-CONTEXT.md (mask-style polish established-pattern)
    - .claude/skills/documentation-standards/SKILL.md (README section)
    - docs/requirements.md §7 (algorithm catalogue — 23 entries with primary-source citations)
    - algoid.go (the exported AlgoID constants — every constant must be referenced in llms.txt)
    - errors.go (the exported sentinels — every sentinel referenced in llms.txt)
    - normalise.go (Normalise, NormalisationOptions, DefaultNormalisationOptions — referenced)
    - tokenise.go (Tokenise, TokeniseOptions, DefaultTokeniseOptions — referenced)
    - golden_canonical.go (WriteGoldenFile if exported — referenced)
    - https://llmstxt.org/ (llms.txt convention reference; track that the spec is evolving)
  </read_first>
  <action>
    1. Polish `README.md` to mask-style. Replace the "Documentation" section's
       "(forthcoming)" markers with active links to the docs files that land
       in this plan. Replace the "🤖 For AI Assistants" placeholder with:
       ```
       This repository ships `llms.txt` (concise index) and `llms-full.txt`
       (full API reference + algorithm citations) at the repo root. AI
       assistants and code generators should consult these first.
       ```
       Add a full algorithm catalogue table (23 rows, grouped by category)
       with columns: Name | Category | Primary source citation (compact
       form) | docs/algorithms.md anchor. Each row links to the
       per-algorithm H2 anchor in docs/algorithms.md.
       Add a "Configuration" section with a one-paragraph teaser about
       NormalisationOptions and TokeniseOptions, with code snippets:
       ```go
       opts := fuzzymatch.DefaultNormalisationOptions()
       opts.StripDiacritics = true
       n := fuzzymatch.Normalise("café", opts)  // "cafe"
       ```
       Add a "Thread Safety" section: every function in the root package is
       pure; the Scorer (Phase 8) is immutable after construction.

    2. Create `llms.txt`:
       ```
       # fuzzymatch — github.com/axonops/fuzzymatch

       A pure-Go library for fuzzy string matching: 23 string-similarity
       algorithms, a weighted Scorer (Phase 8), a collection-scan
       sub-package (Phase 9), and a one-to-many Extract API (Phase 10).
       Stdlib-only with one curated runtime dep (golang.org/x/text).
       Apache-2.0. Go 1.26+.

       ## Public API (root package: fuzzymatch)

       - type AlgoID int (23 constants — see llms-full.txt)
       - func AlgoIDs() []AlgoID
       - func (AlgoID) String() string
       - type NormalisationOptions struct
       - func DefaultNormalisationOptions() NormalisationOptions
       - func Normalise(s string, opts NormalisationOptions) string
       - type TokeniseOptions struct
       - func DefaultTokeniseOptions() TokeniseOptions
       - func Tokenise(s string, opts TokeniseOptions) []string
       - var ErrInvalidInput error
       - var ErrInvalidConfiguration error
       - var ErrInvalidAlgorithm error
       - var ErrEmptyInput error

       ## Out of v1.0 scope

       - Needleman-Wunsch, Soft-TFIDF, Metaphone 3 (patent), embeddings,
         goroutine-parallel Scan, persistent caching.

       ## Documentation

       - README.md — project front door
       - docs/algorithms.md — per-algorithm detail (23 sections)
       - docs/scorer.md — Scorer composition (Phase 8)
       - docs/scan.md — collection scan (Phase 9)
       - docs/extending.md — custom algorithms
       - docs/tuning.md — threshold calibration
       - docs/performance.md — benchmark numbers
       - docs/faq.md — common questions and exclusions

       ## For agents

       - .claude/agents/ — domain review gates
       - .claude/skills/ — reusable standards
       - CLAUDE.md — project conventions
       ```

    3. Create `llms-full.txt`:
       Longer document mirroring llms.txt but with:
       - Full per-symbol signature for every exported root-package identifier
       - Per-AlgoID-constant: name, canonical String() output, primary-source
         citation (one line each)
       - Per-error: full message text
       - Per-option-struct: every field with type and default
       - Sample programs (5-line quick start placeholder for Phase 2)
       - References to docs/requirements.md sections per topic

    4. Create `ai_friendly_test.go` (`package fuzzymatch_test`):
       - AxonOps Apache-2.0 header.
       - Imports: `testing`, `os`, `strings`, `go/parser`, `go/ast`,
         `go/token`.
       - Test `TestAIFriendly_LLMSTxtReferencesEveryExportedSymbol`:
         a. Parse the root package via `parser.ParseDir(token.NewFileSet(),
            ".", nil, 0)`.
         b. Walk the AST, collect every exported identifier name (functions,
            types, methods, vars, consts). Filter to exported names
            (uppercase first letter; skip `_test.go` declarations).
         c. Allowlist internal-only exports: any symbol in
            `internalAllowlist = []string{"WriteGoldenFile"}` (if
            WriteGoldenFile is exported but is test-maintenance-only).
            Document the allowlist with a one-line rationale.
         d. Read `llms.txt`. For each exported symbol NOT in the allowlist,
            assert `strings.Contains(llmsTxtContent, symbolName)`.
         e. Log the full missing-list before failing.

    Concrete identifiers:
      - File `llms.txt` (concise)
      - File `llms-full.txt` (full reference)
      - File `ai_friendly_test.go`, test `TestAIFriendly_LLMSTxtReferencesEveryExportedSymbol`
      - Allowlist: `WriteGoldenFile` if exported (otherwise empty)
  </action>
  <verify>
    <automated>test -f llms.txt &amp;&amp; test -f llms-full.txt &amp;&amp; go test -race -run 'TestAIFriendly_' ./... &amp;&amp; markdownlint-cli2 README.md docs/*.md 2&gt;/dev/null || markdownlint-cli2 README.md</automated>
  </verify>
  <acceptance_criteria>
    - `README.md` no longer contains the literal string "(forthcoming)" — all docs links resolve
    - `README.md` algorithm-catalogue table has 23 rows
    - `README.md` 🤖 For AI Assistants section references llms.txt and llms-full.txt
    - `llms.txt` exists at repo root
    - `llms.txt` references every exported identifier from algoid.go, errors.go, normalise.go, tokenise.go (verifiable by grep for each name)
    - `llms-full.txt` exists at repo root and is longer than llms.txt
    - `ai_friendly_test.go` exists with the AxonOps Apache-2.0 header
    - `go test -race -count=1 -run TestAIFriendly_LLMSTxtReferencesEveryExportedSymbol ./...` exits 0
    - `markdownlint-cli2 README.md` exits 0
    - `make verify-license-headers` exits 0
  </acceptance_criteria>
  <done>
    README is mask-polished; llms.txt + llms-full.txt are committed; the
    sync meta-test enforces that every exported symbol is referenced.
  </done>
</task>

<task type="auto">
  <name>Task 2: Create docs/ scaffolds + makefile_targets_test.go meta-test</name>
  <files>docs/algorithms.md, docs/scorer.md, docs/scan.md, docs/extending.md, docs/tuning.md, docs/performance.md, docs/faq.md, makefile_targets_test.go</files>
  <read_first>
    - .planning/REQUIREMENTS.md (DX-04 listing the doc files; DX-06 listing the FAQ entries)
    - .claude/skills/documentation-standards/SKILL.md (docs/ structure section; algorithm documentation requirements)
    - .claude/skills/algorithm-licensing-standards/SKILL.md (Metaphone 3 precedent — referenced from FAQ)
    - .claude/skills/go-testing-standards/SKILL.md (Meta-tests section — makefile_targets_test.go pattern)
    - docs/requirements.md §7 (23 algorithms in canonical order — for docs/algorithms.md H2 scaffolds)
    - docs/requirements.md §4 (out-of-scope list — for FAQ rationale)
    - Makefile (just created — target list for the meta-test)
    - llms.txt (just created — referenced from docs/faq.md as the LLM-friendly version)
  </read_first>
  <action>
    1. Create `docs/algorithms.md`:
       - H1 "Algorithm Catalogue"
       - One-paragraph introduction citing docs/requirements.md §7 as the
         authoritative spec for per-algorithm formulas and edge cases.
       - 23 H2 sections in spec order. Each H2's anchor is the algorithm's
         canonical spelling (e.g. `## Levenshtein`). Each contains:
         - **Category**: <character/q-gram/token/phonetic/gestalt>
         - **AlgoID constant**: `Algo<Name>`
         - **Primary source**: TBD — filled in by the implementing phase
         - **Status**: planned (no implementation yet — pre-Phase 2)
         - **Cross-reference**: link to docs/requirements.md §7.X
       The 23 H2s are placeholders. Phase 2+ fills in the source-citation
       and the description prose.

    2. Create `docs/scorer.md`:
       - H1 "Scorer (Phase 8)"
       - One-paragraph TBD note: "The Scorer is the second layer of the
         three-layer architecture. It lands in Phase 8. See
         docs/requirements.md §8 for the spec until this document is
         filled in."
       - H2 scaffolds: "Construction", "Defaults", "Composition", "Threshold",
         "ScoreAll", "Match", "Thread Safety" — each with TBD body.

    3. Create `docs/scan.md`, `docs/extending.md`, `docs/tuning.md`,
       `docs/performance.md` — same pattern: H1 + one-paragraph TBD +
       H2 scaffolds with TBD bodies.

    4. Create `docs/faq.md` with DX-06's 6 mandatory entries:
       - H2 "Why no Needleman-Wunsch?"
         Body: Cite docs/requirements.md §4. Smith-Waterman-Gotoh's local-
         alignment variant covers our use cases; Needleman-Wunsch's
         global-alignment specialisation is rarely the right semantic for
         identifier-similarity work.
       - H2 "Why no Metaphone 3?"
         Body: U.S. Patent 7,440,941 (Lawrence Philips 2009). AxonOps
         declines patent-encumbered algorithms regardless of enforcement
         posture or alternative-implementation availability. Cite
         algorithm-licensing-standards SKILL.md and docs/requirements.md §4.
         Double Metaphone (Philips 2000, patent-free) and NYSIIS cover the
         phonetic use cases.
       - H2 "Why no embeddings / ML?"
         Body: Pure-function library, stdlib-mostly. ML / embeddings
         require model storage, runtime dependencies (torch / onnxruntime
         / transformers), and consumer-side stateful caching. Out of scope.
       - H2 "Why phonetic-as-binary in the Scorer?"
         Body: Phonetic algorithms produce discrete encoded codes (Soundex
         → "T522"); two inputs either share a code or they don't. The
         canonical normalisation per docs/requirements.md §7.20-§7.23 is
         1.0 for code match, 0.0 otherwise. Continuous-similarity
         alternatives (Levenshtein over phonetic codes) are available but
         not the default.
       - H2 "Why aren't algorithm functions generic?"
         Body: Go generics introduce indirection via type-parameter dispatch
         that breaks our zero-allocation fast-path budget on character-based
         algorithms. `string` is the lowest-common-denominator type;
         dispatch via the typed `AlgoID` enum (array-backed, zero-alloc)
         per docs/requirements.md §6. Phase 6's Monge-Elkan accepts an
         inner-metric AlgoID parameter for parametric use cases.
       - H2 "Why golang.org/x/text but no other deps?"
         Body: Unicode NFC/NFD + diacritic stripping is table-stakes for
         the audit-event-taxonomy consumer (Müller vs Mueller). Inlining
         a maintained Unicode normalisation implementation would require
         ongoing Unicode database updates and dwarf the rest of the
         library. `x/text` is Go-team maintained, supply-chain auditable,
         and added to the allowlist on 2026-05-13 per the documented
         decision in CONTEXT.md. The allowlist is locked at one entry;
         future additions require user approval + algorithm-licensing-
         reviewer sign-off.

    5. Create `makefile_targets_test.go` (`package fuzzymatch_test`):
       - AxonOps Apache-2.0 header.
       - Imports: `testing`, `os`, `regexp`, `strings`.
       - Test `TestMakefile_TargetsDocumentedInContributing`:
         a. Read Makefile, extract all `^[a-z][a-z-]*:` target names (filter
            to canonical target names — strip targets that are obviously
            internal like `_phony_helper`).
         b. Read CONTRIBUTING.md.
         c. For each target name, assert it appears in CONTRIBUTING.md OR
            in the Makefile near a `## suppress:` comment.
         d. Reverse: for each target mentioned in CONTRIBUTING (extracted
            via a regex matching `\b<target>:` or backticked `<target>`),
            assert it exists in Makefile.
         e. Log all missing entries before failing.
       - Test `TestMakefile_HasCanonicalTargets`:
         a. Hardcoded list of the 19 canonical targets from CLAUDE.md.
         b. For each, grep the Makefile for `^<target>:` line. Fail if any
            is missing.

    Concrete identifiers:
      - 7 docs/*.md scaffolds
      - 23 H2 sections in docs/algorithms.md
      - 6 H2 sections in docs/faq.md (DX-06)
      - File `makefile_targets_test.go`, 2 test functions
  </action>
  <verify>
    <automated>for f in docs/algorithms.md docs/scorer.md docs/scan.md docs/extending.md docs/tuning.md docs/performance.md docs/faq.md; do test -f "$f" || { echo "missing $f"; exit 1; }; done &amp;&amp; grep -c '^## ' docs/algorithms.md | xargs -I{} test {} -ge 23 &amp;&amp; grep -c '^## ' docs/faq.md | xargs -I{} test {} -ge 6 &amp;&amp; go test -race -run 'TestMakefile_' ./... &amp;&amp; markdownlint-cli2 docs/*.md</automated>
  </verify>
  <acceptance_criteria>
    - All 7 docs/*.md files exist
    - `docs/algorithms.md` has at least 23 H2 sections
    - `docs/faq.md` has at least 6 H2 sections, including the mandatory DX-06 entries (Needleman-Wunsch, Metaphone 3, embeddings, phonetic-as-binary, generics, x/text)
    - `makefile_targets_test.go` exists with the AxonOps Apache-2.0 header
    - `go test -race -count=1 -run TestMakefile_ ./...` exits 0
    - `markdownlint-cli2 docs/*.md` exits 0
    - Every Makefile target appears in CONTRIBUTING.md (verified by `TestMakefile_TargetsDocumentedInContributing` once CONTRIBUTING.md is written in Task 3)
  </acceptance_criteria>
  <done>
    Seven docs/* scaffolds are committed; 23 algorithm H2 sections placeholder
    every catalogue entry; DX-06's 6 FAQ entries are present; the makefile-
    targets meta-test enforces bi-directional documentation.
  </done>
</task>

<task type="auto">
  <name>Task 3: Create CODEOWNERS, issue templates, PR template, SECURITY, CODE_OF_CONDUCT, CONTRIBUTING</name>
  <files>SECURITY.md, CODE_OF_CONDUCT.md, CONTRIBUTING.md, .github/CODEOWNERS, .github/ISSUE_TEMPLATE/bug.yml, .github/ISSUE_TEMPLATE/feature.yml, .github/ISSUE_TEMPLATE/algorithm-proposal.yml, .github/PULL_REQUEST_TEMPLATE.md
  <read_first>
    - .claude/skills/documentation-standards/SKILL.md (CONTRIBUTING + algorithm doc rules)
    - .claude/skills/algorithm-licensing-standards/SKILL.md (Source Origin Statement pattern for PR template)
    - .claude/skills/commit-standards/SKILL.md (Conventional-commit format for CONTRIBUTING reference)
    - .claude/skills/fuzzymatch-review-protocol/SKILL.md (Review gates listing — referenced from PR template checklist)
    - CLAUDE.md ("Workflow — Agent Gates"; "Releases — CI Only"; "Branching & Commits"; "GitHub Issues Are the Source of Truth")
    - https://www.contributor-covenant.org/version/2/1/code_of_conduct.txt (CoC text)
    - Makefile (target list — referenced from CONTRIBUTING; meta-test from Task 2 enforces parity)
    - README.md (just polished — CONTRIBUTING links back to it)
    - LICENSE + NOTICE (existing — referenced from SECURITY.md)
  </read_first>
  <action>
    Land the contributor + security + policy documentation, CODEOWNERS,
    and GitHub UI templates.

    1. `SECURITY.md`:
       - H1 "Security Policy"
       - H2 "Supported Versions":
         Table — v0.x.y rows marked "pre-release / no SLA"; v1.x.y rows
         (added when v1.0.0 ships) marked "supported".
       - H2 "Reporting a Vulnerability":
         Email `security@axonops.com` (placeholder — the user confirms the
         real address at execution time and records in SUMMARY). Request
         that reporters do NOT open public issues for vulnerabilities.
       - H2 "Disclosure Timeline":
         90-day coordinated disclosure standard. Acknowledgment within 2
         business days; initial assessment within 7; fix or workaround
         within 30; public disclosure 90 days after initial report unless
         negotiated otherwise.
       - H2 "Verification":
         For released versions (v1.0.0+), consumers verify the release
         signature via cosign:
         ```
         cosign verify-blob \
           --bundle checksums.txt.bundle \
           --certificate-identity-regexp 'https://github.com/axonops/fuzzymatch/.+' \
           --certificate-oidc-issuer https://token.actions.githubusercontent.com \
           checksums.txt
         ```
         Cite plan 01-03's release.yml as the canonical signing workflow.
       - H2 "Security Tooling":
         List the CI pipeline: govulncheck weekly + on every PR, gosec
         with SARIF upload to GitHub Security tab, CodeQL weekly + on
         every PR. Cite the workflows by name.

    2. `CODE_OF_CONDUCT.md`:
       - Use the Contributor Covenant 2.1 text verbatim from
         https://www.contributor-covenant.org/version/2/1/code_of_conduct.txt.
       - Substitute the contact email placeholder `[INSERT CONTACT METHOD]`
         with `conduct@axonops.com` (or the user's confirmed address —
         record in SUMMARY).
       - At the bottom add: "This Code of Conduct is adapted from the
         Contributor Covenant, version 2.1, available at
         https://www.contributor-covenant.org/version/2/1/code_of_conduct.html".

    3. `CONTRIBUTING.md`:
       - H1 "Contributing to fuzzymatch"
       - H2 "Welcome / Scope":
         One paragraph: pre-release, external contributions welcome via
         issues during v0.x; PRs accepted after v1.0.0 freeze. Reference
         CLAUDE.md as the canonical workflow source.
       - H2 "Pre-PR Checklist":
         Numbered list — `make check` green, conventional commit, sign-off
         via CLA Assistant (link to plan 01-03's cla.yml), CHANGELOG entry
         under `[Unreleased]`, `make bench-compare` run locally if
         touching algorithm code, PR template fields filled.
       - H2 "Local Development Setup":
         Required tools at versions (Go 1.26.3, golangci-lint v2.12.2,
         govulncheck, goimports, optional benchstat). Reference
         `.planning/research/STACK.md`.
       - H2 "Make Targets":
         Bulleted list of all 19 canonical targets with one-line
         descriptions. This list is the source of truth for the
         makefile_targets_test.go meta-test (Task 2).
       - H2 "Conventional Commits":
         Brief explanation. Link to `.claude/skills/commit-standards/SKILL.md`.
         Explicit "No AI attribution in commit messages" rule.
       - H2 "Algorithm Contribution Flow":
         Numbered list — open issue using algorithm-proposal template;
         algorithm-licensing-reviewer screens; algorithm-correctness-
         reviewer verifies primary-source citation; implement against
         citation; cross-validate with literature reference vectors;
         all review gates pass. Cite CLAUDE.md "Workflow — Agent Gates".
       - H2 "Algorithm Deprecation Policy" (REL-07):
         Quoted block per <interfaces>:
         "Within a major version, algorithms may be ADDED but never
         REMOVED. Score-changing edits to an existing algorithm require
         a minor version bump and a CHANGELOG entry. Bug fixes that
         change scores are minor; intentional algorithm-formula changes
         are minor; algorithm removals are major (v2.x.y or later)."
         Cite docs/requirements.md §11.2.
       - H2 "Release Process":
         Three paragraphs: tag push triggers automated release; no local
         `git tag` or `goreleaser release`; maintainers ONLY use the GitHub
         tag-push or release-creation UI. Cite plan 01-03's release.yml.
       - H2 "Benchmark Discipline" (D-09):
         Paragraph: `bench.txt` is committed at the repo root and updated
         locally via `make bench`. CI runs `make bench-compare`
         informationally (does not block) until a self-hosted runner
         exists. Contributors touching algorithm code MUST run
         `make bench-compare` locally and explain any regression > 10%
         in the PR description.

    4. `.github/CODEOWNERS`:
       Single line: `* @<maintainer-github-username>` — placeholder. The
       user fills in the actual maintainer username at execution time.
       Documentation in the file header: "CODEOWNERS routes review.
       Expand to per-area routing as the project grows."

    5. `.github/ISSUE_TEMPLATE/bug.yml`:
       YAML form per GitHub's issue-form schema. Fields:
       - title prefix `bug: `
       - description summary (textarea, required)
       - reproduction steps (textarea, required)
       - expected behaviour (input, required)
       - actual behaviour (input, required)
       - Go version (input, required)
       - OS/arch (dropdown: linux/amd64, linux/arm64, darwin/amd64,
         darwin/arm64, windows/amd64, other)
       - fuzzymatch version or commit SHA (input, required)
       - additional context (textarea, optional)
       - labels: `[bug, needs-triage]`

    6. `.github/ISSUE_TEMPLATE/feature.yml`:
       Fields:
       - title prefix `feat: `
       - use-case description (textarea, required)
       - proposed API shape (textarea, optional — flag will go to
         api-ergonomics-reviewer)
       - alternatives considered (textarea, optional)
       - related issues / discussions (input, optional)
       - labels: `[enhancement, needs-triage]`

    7. `.github/ISSUE_TEMPLATE/algorithm-proposal.yml`:
       Fields:
       - title prefix `algo: `
       - algorithm name (input, required)
       - primary academic source citation (textarea, required — full
         citation: author, year, title, journal/conference, page range)
       - patent screen status (dropdown: yes (encumbered), no (clear),
         unknown — required; "unknown" triggers algorithm-licensing-
         reviewer screening)
       - existing Go implementations studied (textarea, optional —
         reference-vector cross-validation only, no code derivation)
       - expected reference vectors (textarea, required — at least 3
         input → expected-output pairs with primary-source attribution)
       - related existing algorithms in our catalogue (input, optional)
       - rationale for inclusion (textarea, required — why this algorithm
         is worth adding given the curated-catalogue discipline)
       - labels: `[algorithm-proposal, needs-licence-screen, needs-triage]`

    8. `.github/PULL_REQUEST_TEMPLATE.md`:
       Markdown sections:
       - Summary (1-3 sentences)
       - Related issue: `Fixes #` or `Refs #`
       - Type of change: checkboxes (bug fix, new feature, breaking
         change, docs, refactor, perf, test, chore)
       - Algorithm-specific section (only filled for algorithm PRs):
         Source Origin Statement per algorithm-licensing-standards:
         ```
         Source origin:
         - Primary source: [citation]
         - Studied for reference vectors: [list, with licences]
         - No code copied from any source: confirmed
         - No GPL/LGPL references consulted: confirmed
         ```
       - Test plan: bulleted checklist (`go test -race`, `make check`,
         benchstat if algo code, BDD scenarios)
       - CHANGELOG entry: confirmation that `[Unreleased]` updated
       - Reviewer checklist: the 5 named review gates from CLAUDE.md
         (algorithm-correctness, algorithm-performance, determinism,
         api-ergonomics, code-reviewer) — auto-invoked by the verifier
         protocol.

    Concrete identifiers:
      - 8 files
      - CODEOWNERS placeholder maintainer username (record actual at
        execution time)
      - Issue templates use GitHub's YAML form schema
      - PR template carries the Source Origin Statement section
  </action>
  <verify>
    <automated>for f in SECURITY.md CODE_OF_CONDUCT.md CONTRIBUTING.md .github/CODEOWNERS .github/ISSUE_TEMPLATE/bug.yml .github/ISSUE_TEMPLATE/feature.yml .github/ISSUE_TEMPLATE/algorithm-proposal.yml .github/PULL_REQUEST_TEMPLATE.md; do test -f "$f" || { echo "missing $f"; exit 1; }; done &amp;&amp; for f in .github/ISSUE_TEMPLATE/*.yml; do python3 -c "import yaml; yaml.safe_load(open('$f'))" || exit 1; done &amp;&amp; grep -q 'deprecation' CONTRIBUTING.md &amp;&amp; grep -q 'make check' CONTRIBUTING.md &amp;&amp; grep -q 'bench-compare' CONTRIBUTING.md &amp;&amp; grep -q 'Contributor Covenant' CODE_OF_CONDUCT.md &amp;&amp; grep -q 'cosign' SECURITY.md &amp;&amp; go test -race -run 'TestMakefile_TargetsDocumentedInContributing' ./... &amp;&amp; markdownlint-cli2 SECURITY.md CODE_OF_CONDUCT.md CONTRIBUTING.md</automated>
  </verify>
  <acceptance_criteria>
    - All 8 files exist
    - `SECURITY.md` has H2 sections: Supported Versions, Reporting a Vulnerability, Disclosure Timeline, Verification, Security Tooling
    - `SECURITY.md` references the `cosign verify-blob` command
    - `CODE_OF_CONDUCT.md` is Contributor Covenant 2.1 text with substituted contact email
    - `CONTRIBUTING.md` references `make check`, `make bench-compare`, conventional commits, algorithm deprecation policy, release-via-CI-only
    - `CONTRIBUTING.md` documents every Makefile target (verified by makefile_targets_test.go from Task 2)
    - `.github/CODEOWNERS` exists with at least one routing rule
    - `.github/ISSUE_TEMPLATE/bug.yml`, `feature.yml`, `algorithm-proposal.yml` are valid YAML
    - Each issue template uses GitHub's issue-form schema (`name:`, `description:`, `body:` keys)
    - `.github/PULL_REQUEST_TEMPLATE.md` contains the Source Origin Statement section template
    - `markdownlint-cli2 SECURITY.md CODE_OF_CONDUCT.md CONTRIBUTING.md` exits 0
    - `go test -race -run TestMakefile_TargetsDocumentedInContributing ./...` exits 0 (the meta-test from Task 2 passes against the new CONTRIBUTING.md)
    - `make check` exits 0
  </acceptance_criteria>
  <done>
    SECURITY, COC, CONTRIBUTING, CODEOWNERS, 3 issue templates, and PR
    template are committed; CONTRIBUTING documents the deprecation policy
    (REL-07) and every Makefile target; the meta-test enforces parity.
  </done>
</task>

</tasks>

<threat_model>
## Trust Boundaries

| Boundary | Description |
|----------|-------------|
| External contributor → repo | Issue templates and PR template constrain submissions to a reviewable shape; CODEOWNERS routes reviews. |
| AI-assistant code generator → repo | `llms.txt` and `llms-full.txt` provide a stable, public-API-accurate reference; the sync meta-test enforces fidelity. |
| Reporter of vulnerability → maintainer | SECURITY.md establishes the disclosure channel and timeline. |
| Documentation drift → user trust | The makefile_targets_test.go and ai_friendly_test.go meta-tests catch documentation drift in CI. |

## STRIDE Threat Register

| Threat ID | Category | Component | Disposition | Mitigation Plan |
|-----------|----------|-----------|-------------|-----------------|
| T-01-08-01 | Information Disclosure | Outdated llms.txt mis-leads AI assistants into generating non-existent API calls | mitigate | `ai_friendly_test.go` parses the root package AST and asserts every exported symbol is referenced in llms.txt. CI runs the meta-test on every PR; drift fails CI. |
| T-01-08-02 | Tampering | Makefile targets undocumented in CONTRIBUTING — contributors invoke wrong commands | mitigate | `makefile_targets_test.go` enforces bi-directional documentation coverage; PRs adding new targets MUST update CONTRIBUTING. |
| T-01-08-03 | Repudiation | Vulnerability reported via public issue (no private channel) | mitigate | SECURITY.md publishes the private email channel; CODE_OF_CONDUCT references it; issue templates do NOT include a "security" type (forcing private channel use). |
| T-01-08-04 | Spoofing | Patent-encumbered algorithm proposal slipping through review | mitigate | algorithm-proposal.yml template's "patent screen status" field forces explicit declaration; the `needs-licence-screen` label triggers algorithm-licensing-reviewer per CLAUDE.md Workflow — Agent Gates. |
| T-01-08-05 | Tampering | Algorithm deprecation breaking consumers | mitigate | REL-07 deprecation policy documented in CONTRIBUTING.md: within-major-version, algorithms can be added but not removed; scoring changes require minor bump. Reviewers and release process enforce. |
| T-01-08-06 | Elevation of Privilege | Compromised CODEOWNERS routing | accept | CODEOWNERS is plain text in the repo; routing changes go through normal PR review like any other change. Initial state is single-maintainer placeholder; user expands as project grows. |
| T-01-08-07 | Repudiation | Source Origin Statement omitted for algorithm PR | mitigate | PR template's algorithm-specific section makes the statement explicit and required; algorithm-licensing-reviewer verifies during PR review. |
| T-01-08-08 | Information Disclosure | CoC contact email becomes stale | accept | CoC contact email is a static text; user updates manually when the address changes. Low maintenance burden. |
</threat_model>

<verification>
1. `make check` exits 0.
2. `markdownlint-cli2 "**/*.md"` exits 0 (excluding .planning and .claude
   per plan 01-02's config).
3. `go test -race -count=1 -run 'TestAIFriendly_|TestMakefile_' ./...` exits 0.
4. All 8 files in Task 3 exist; issue templates are valid YAML.
5. README's algorithm catalogue table has 23 rows.
6. docs/faq.md has at least 6 H2 sections covering the DX-06 mandatory list.
7. CONTRIBUTING.md references make check, make bench-compare, the
   deprecation policy, and conventional commits.
8. `make verify-license-headers` exits 0 (all new .go files carry the header).
9. `make verify-deps-allowlist` exits 0 (root go.mod still only x/text).
10. `make verify-determinism` exits 0 (TestGolden_Normalisation still
    diff-clean).
</verification>

<success_criteria>
- README is mask-polished with the full 23-algorithm catalogue and
  AI-assistant pointer.
- llms.txt and llms-full.txt are committed; ai_friendly_test.go enforces
  sync.
- 7 docs/*.md scaffolds exist; docs/faq.md covers DX-06's 6 mandatory
  entries.
- SECURITY, CODE_OF_CONDUCT, CONTRIBUTING are complete; CONTRIBUTING
  documents the algorithm deprecation policy (REL-07), make check, and
  the bench-compare local-driven story (D-09).
- CODEOWNERS, 3 issue templates, and PR template are committed.
- makefile_targets_test.go meta-test enforces bi-directional Makefile/
  CONTRIBUTING coverage.
- Phase 1's full deliverable set per ROADMAP Success Criteria 1-5 is now
  shippable: clean go build, 5-platform CI matrix passing, make check
  green, Normalise + Tokenise with golden file pinning, signed release
  pipeline ready (a tag push would produce a signed release).
</success_criteria>

<output>
After completion, create
`.planning/phases/01-foundation-infrastructure/01-08-dx-docs-SUMMARY.md`
recording:
  - The chosen contact emails (security@axonops.com / conduct@axonops.com
    or the actual addresses).
  - The maintainer username placed in CODEOWNERS.
  - Whether CLA.md was added in this plan (deferred from plan 01-03 — if
    still absent, record as a follow-up).
  - Coverage of DX-06's 6 FAQ entries (one line each).
  - Any markdownlint waivers added to .markdownlint-cli2.yaml for the new
    docs (none expected).
  - Confirmation that Phase 1's 5 Success Criteria from ROADMAP §Phase 1
    are now satisfied.
</output>
