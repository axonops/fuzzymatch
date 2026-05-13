# Phase 1: Foundation & Infrastructure - Context

**Gathered:** 2026-05-13
**Status:** Ready for planning

<domain>
## Phase Boundary

Phase 1 lands the entire safety scaffolding **before any algorithm is implemented**. Deliverables:

1. **Module bootstrap** — `go.mod` (Go 1.26.3 directive, `golang.org/x/text` as the sole runtime dep), Apache-2.0 LICENSE + NOTICE, per-file Apache-2.0 headers, root directory scaffolding, `tests/bdd/go.mod` sub-module with `replace` directive (godog v0.15.0 + goleak v1.3.0 + testify v1.10.0).
2. **Quality gates** — golangci-lint v2.12.2 (`version: "2"` config), `go vet`, `go test -race -shuffle=on -count=1`, govulncheck, gosec v2.25.0 + SARIF, CodeQL `github/codeql-action@v4`, markdownlint-cli2 v0.22.1, Dependabot (gomod root + tests/bdd + github-actions), Makefile (`check`, `lint`, `vet`, `test`, `bench`, `bench-compare`, `coverage`, `tidy-check`, `verify-deps-allowlist`, `verify-determinism`, `verify-license-headers`, `security`, `clean`).
3. **Release pipeline** — GoReleaser v2.15.4 (`~> v2` pin) producing `checksums.txt` + Syft SPDX-JSON SBOM, Cosign v3.0.1 keyless `sign-blob --bundle` via GitHub OIDC, `actions/attest-build-provenance@v2`, `wagoid/commitlint-github-action` on PR titles + commits, `contributor-assistant/github-action` (CLA Assistant), `release.yml` triggered exclusively by tag push.
4. **Determinism infrastructure** — cross-platform CI matrix (linux/amd64, linux/arm64, darwin/amd64, darwin/arm64, windows/amd64), `testdata/golden/` directory, golden-file harness with `-update` flag, `make verify-deps-allowlist` (asserts the only `require` line in root `go.mod` is `golang.org/x/text`), `make verify-determinism` (diffs committed golden files against current marshalling).
5. **Foundation primitives** —
   - `AlgoID` typed `int` enum with `iota`-based constants, `String()` method, package-level `AlgoIDs() []AlgoID` accessor, dispatch table backing (no `Algorithm` interface — avoids hot-path boxing per `.claude/skills/go-coding-standards/SKILL.md`).
   - `Normalise(s string, opts NormalisationOptions) string` — struct-of-options API. `NormalisationOptions` fields: `Lowercase bool`, `StripSeparators bool`, `SeparatorChars string`, `SplitCamelCase bool`, `NFC bool`, `StripDiacritics bool`. `DefaultNormalisationOptions()` returns `{Lowercase: true, StripSeparators: true, SeparatorChars: "_-.:/", SplitCamelCase: true, NFC: true, StripDiacritics: false}`. Unicode NFC/NFD via `golang.org/x/text/unicode/norm`; ASCII fast path; diacritic stripping opt-in.
   - `Tokenise(s string, opts TokeniseOptions) []string` — splits camelCase / snake_case / PascalCase / kebab-case / dot-case with stable ordering. `TokeniseOptions` struct shape TBD by `api-ergonomics-reviewer`.
   - Flat sentinel error set via `errors.New` — `ErrInvalidInput`, `ErrInvalidConfiguration`, `ErrInvalidAlgorithm`, `ErrEmptyInput` (full list TBD by `api-ergonomics-reviewer`); wrap via `fmt.Errorf("%w", ErrX)`.
6. **DX scaffolding** — README (mask-style polish + 5-line quickstart placeholder), CHANGELOG (Keep-a-Changelog seed), `llms.txt` + `llms-full.txt` + sync meta-test, `docs/algorithms.md` / `scorer.md` / `scan.md` / `extending.md` / `tuning.md` / `performance.md` / `faq.md` scaffolds, CODEOWNERS, issue templates (bug, feature, algorithm-proposal), PR template, SECURITY.md, CODE_OF_CONDUCT.md, CONTRIBUTING.md.

**Out of Phase 1 scope (explicitly):** any string-similarity algorithm code (lands in Phase 2+); the `Scorer` (Phase 8); the `scan` sub-package (Phase 9); `Extract` (Phase 10); algorithm-specific golden file entries (each algorithm phase extends `algorithms.json`).

</domain>

<decisions>
## Implementation Decisions

### Foundation API shape

- **D-01: `AlgoID` is a typed `int` constant + dispatch array.** `type AlgoID int` with `iota`-based unexported-then-exported constants (concrete naming per `api-ergonomics-reviewer`). Dispatch via a fixed-size array `[N]func(a, b string) float64` indexed by `AlgoID`, so `Scorer.Score` and `Extract` perform a zero-allocation array index — no interface boxing on the hot path per `.claude/skills/go-coding-standards/SKILL.md`. Adding a new algorithm requires a code change (acceptable for a curated catalogue; spec `docs/requirements.md` §6).
- **D-02: `Normalise` uses a struct-of-options, not variadic functional options.** `Normalise(s string, opts NormalisationOptions) string` matches `docs/requirements.md` §9. The struct is passed by value so per-call allocation is zero. Fields: `Lowercase bool`, `StripSeparators bool`, `SeparatorChars string`, `SplitCamelCase bool`, `NFC bool`, `StripDiacritics bool`. `NFC` and `StripDiacritics` are additive fields extending the spec to cover the audit-event-taxonomy use case.
- **D-03: `DefaultNormalisationOptions()` returns conservative defaults that preserve diacritics.** Defaults: `Lowercase=true`, `StripSeparators=true`, `SeparatorChars="_-.:/"`, `SplitCamelCase=true`, `NFC=true`, `StripDiacritics=false`. Consumers who want `café → cafe` semantics opt in explicitly via `WithStripDiacritics()`-style override (exact override API per `api-ergonomics-reviewer`).
- **D-04: `Tokenise` returns `[]string`.** `Tokenise(s string, opts TokeniseOptions) []string`. Slice return is the idiomatic Go shape, predictable allocation, easy to pass to set/map operations the downstream token algorithms (Phase 6) need. `iter.Seq[string]` rejected because consumers materialise anyway, eliminating any savings.
- **D-05: Error hierarchy is flat sentinels via `errors.New`.** Package-level `var ErrInvalidInput = errors.New("fuzzymatch: invalid input")` style; wrap with `fmt.Errorf("%w", ErrX)`. Composes with `errors.Is` / `errors.As` trivially. Typed error structs deferred unless Phase 9 (scan) needs richer per-item context.

### Plan slicing strategy

- **D-06: Phase 1 ships as 8 plans, executed strictly sequentially.** No wave parallelism in Phase 1 — CI/release/quality-gate plans (02/03/04) touch overlapping workflow files and would conflict if parallelised. The plan list (final filenames per `gsd-planner` `{padded_phase}-{NN}-PLAN.md` convention):
  1. **`01-01-module-bootstrap`** — `go.mod` (Go 1.26.3, `golang.org/x/text` only), `tests/bdd/go.mod` sub-module + `replace` directive, LICENSE, NOTICE, Apache-2.0 file-header verifier (`scripts/verify-license-headers.sh`), directory scaffolding.
  2. **`01-02-quality-gates`** — `.golangci.yml` (v2 config), `make check/lint/vet/fmt-check/test/tidy-check/security/coverage`, govulncheck workflow, gosec workflow with SARIF upload, CodeQL workflow, `.markdownlint-cli2.yaml`, Dependabot config (gomod root + tests/bdd + github-actions, daily, grouped indirect/direct/test-only/actions).
  3. **`01-03-release-pipeline`** — `.goreleaser.yml` (v2 schema, library-only — `checksums.txt` + source tarball, no binaries), `release.yml` (tag-push trigger, cosign keyless `sign-blob --bundle`, Syft SBOM, `actions/attest-build-provenance@v2`), `wagoid/commitlint-github-action` workflow, CLA Assistant workflow, `.commitlintrc.yml`.
  4. **`01-04-determinism-infra`** — cross-platform CI matrix workflow (`ci.yml` running on linux/amd64+arm64, darwin/amd64+arm64, windows/amd64 with `CGO_ENABLED=0`), `testdata/golden/` directory, `scripts/verify-no-runtime-deps.sh` (root `go.mod` `require` block must be `golang.org/x/text` only), `make verify-deps-allowlist`, `make verify-determinism`, empty `bench.txt` placeholder, `make bench`, `make bench-compare` (calls `benchstat` against committed `bench.txt`).
  5. **`01-05-primitives-algoid-errors`** — `algoid.go` (`AlgoID` typed int, `iota` constants, `String()` method, `AlgoIDs()` accessor, dispatch array skeleton sized for the 23 algorithms with all entries nil for now), `errors.go` (sentinel set), unit + property tests for `String()` round-trip determinism.
  6. **`01-06-primitives-normalise`** — `normalise.go` (`NormalisationOptions` struct, `DefaultNormalisationOptions()`, `Normalise` with ASCII fast path, Unicode NFC/NFD pipeline via `golang.org/x/text/unicode/norm`, diacritic stripping when `StripDiacritics=true`), unit tests with literature-citation comments, property tests (idempotence under `Normalise(Normalise(s)) == Normalise(s)`, length-bound invariant), populated `testdata/golden/normalisation.json` (20–40 entries — see D-09).
  7. **`01-07-primitives-tokenise`** — `tokenise.go` (`TokeniseOptions` struct, `Tokenise` splitter), unit tests, property tests (order stability under permutation of equal-rank tokens; reconstructibility when applicable).
  8. **`01-08-dx-docs`** — README (mask-style polish, 5-line quickstart placeholder pending Phase 2's first algorithm), CHANGELOG (Keep-a-Changelog seed), `llms.txt` + `llms-full.txt` + `ai_friendly_test.go` sync meta-test parsing `go/ast`, `docs/algorithms.md` / `scorer.md` / `scan.md` / `extending.md` / `tuning.md` / `performance.md` / `faq.md` (scaffolds — TBD entries marked), `SECURITY.md`, `CODE_OF_CONDUCT.md`, `CONTRIBUTING.md`, `.github/CODEOWNERS`, `.github/ISSUE_TEMPLATE/{bug,feature,algorithm-proposal}.yml`, `.github/PULL_REQUEST_TEMPLATE.md`.
- **D-07: `go.mod` is frozen after `01-01-module-bootstrap`.** Only `golang.org/x/text` appears as a `require`. `make verify-deps-allowlist` and `make tidy-check` are CI gates that fail any subsequent change to the root `go.mod` `require` block. Extending the allowlist requires explicit user approval and `algorithm-licensing-reviewer` sign-off per CLAUDE.md.

### Governance + runner choices

- **D-08: CLA Assistant via `contributor-assistant/github-action`.** Mirrors `axonops/mask`. Signatures stored decentralised in-repo. Dependabot + bots allowlisted. Mask consistency lowers contributor cognitive load.
- **D-09: GitHub-hosted ubuntu-latest CI runners only — no self-hosted.** Acknowledged compromise. **PERF-04 spec gate is RELAXED for Phase 1:** the requirement "benchstat regression >10% at p<0.05 fails CI" is deferred until a self-hosted runner is available. Phase 1 ships the local-workstation workflow (`make bench`, `make bench-compare`) and CI runs `make bench-compare` informationally (prints diff in PR, does not fail). `bench.txt` is committed to the repo from the user's workstation. README + CONTRIBUTING document "run `make bench-compare` locally before opening a PR that touches algorithm code." Revisit when shared infrastructure exists. Tracked in Deferred Ideas.
- **D-10: Mask-style branch protection and conventional-commit lint.** `wagoid/commitlint-github-action` with `@commitlint/config-conventional` rules on PR titles and all commits. `main` requires: PR (no direct push), all CI checks green (build, test, lint, vet, vulncheck, gosec, codeql, markdownlint, tidy-check, verify-deps-allowlist, verify-determinism, coverage ≥ 95%), linear history. Commit signing not required (mask doesn't require it; defer unless mask changes). CODEOWNERS routes review (maintainer-only initially; expand as project grows).

### Golden-file & determinism scaffolding

- **D-11: Phase 1 lands `normalisation.json` only — real and populated (20–40 entries).** Algorithm phases extend `algorithms.json` incrementally as each algorithm lands; Phase 8 adds `scorer-default.json`; Phase 9 adds `scan-default.json`; Phase 10 adds `extract-default.json`. The `normalisation.json` corpus exercises: pure-ASCII fast path (`"FooBar" → "foo bar"`), NFC/NFD divergent inputs (`"café"` pre-composed vs decomposed → identical normalised output), diacritic stripping ON+OFF (`"Müller"` with `StripDiacritics=true` → `"muller"`, with `false` → `"müller"`), snake_case/camelCase/PascalCase/kebab-case/dot-case splits, separator-strip edge cases, mixed-script Unicode (e.g., Cyrillic, Arabic — preserved), idempotence pairs, and at least one entry exercising every `NormalisationOptions` field combination touched in Phase 1 code.
- **D-12: Sorted-struct + `json.MarshalIndent` canonical-form format.** Golden file structure: a Go struct with sorted-by-key entries (alphabetic ordering enforced at the struct level, not at marshal time, so the canonical form is deterministic and human-readable). Encoded via `json.MarshalIndent(v, "", "  ")` with `"\n"` line endings (no trailing CRLF on Windows), UTF-8 no BOM. **The format locks across all 5 golden files for v1.x.** Every subsequent algorithm/Scorer/scan/Extract phase MUST emit the same canonical form.
- **D-13: `testdata/golden/` + `go test ./... -update` regen flag.** Golden files live at `testdata/golden/normalisation.json` (and later `algorithms.json` / `scorer-default.json` / `scan-default.json` / `extract-default.json`). A `-update` flag (declared via `flag.Bool("update", false, ...)` in a `golden_test.go`) regenerates them from current code. CI runs `go test ./...` without `-update` so any divergence between committed and current fails. README + CONTRIBUTING document "run `go test ./... -update` after intentional changes; review the diff before committing."
- **D-14: Each platform in the CI matrix diffs against the committed file independently.** `TestGolden_Normalisation` marshals current outputs in the canonical form and diffs byte-for-byte against `testdata/golden/normalisation.json`. **Any divergence on ANY platform fails CI on that platform.** On failure, the test prints a unified diff to the test log so the divergence is immediately diagnosable in the PR. No cross-platform consolidation job — the per-platform diff against a single committed file IS the cross-platform check.

### Claude's Discretion

- Exact concrete names for `AlgoID` constants, `NormalisationOptions` field accessors (if any helper methods exist), `TokeniseOptions` field set, sentinel error names beyond the spec list, Makefile target naming, and CI job names are delegated to `api-ergonomics-reviewer` and `user-guide-reviewer` per CLAUDE.md Design Principle 13.
- Specific entries in `normalisation.json` (the 20–40 pinned input/output pairs) are picked by `gsd-planner` / `test-writer` from the spec's normalisation examples, the audit-taxonomy use cases, and Unicode normalisation reference vectors. Property tests verify the corpus exercises every code path.
- The exact set of sentinel errors beyond the four named (`ErrInvalidInput`, `ErrInvalidConfiguration`, `ErrInvalidAlgorithm`, `ErrEmptyInput`) — additions are flagged to `api-ergonomics-reviewer` during planning.

</decisions>

<canonical_refs>
## Canonical References

**Downstream agents MUST read these before planning or implementing.**

### Authoritative spec
- `docs/requirements.md` — the deep technical contract (1812 lines). Read in full before planning. Specific Phase-1-relevant sections:
  - §1 — module, Go 1.26+, zero-runtime-deps, no cgo, Apache-2.0
  - §5 — design principles (no cgo, no goroutines, no `init()`-time table builds, no map iteration on output, no transcendentals on output paths, fresh-implementation discipline)
  - §6 — AlgoID dispatch (typed enum, table-backed, String() method, AlgoIDs() accessor)
  - §6.4 — sentinel error hierarchy
  - §9 — Normalise pipeline (struct-of-options shape)
  - §10 — Tokenise (camelCase/snake_case/PascalCase/kebab-case splitting)
  - §11 — determinism (cross-platform byte-identical output, no map iteration, NaN/Inf/-0 handling, golden files, no transcendentals)
  - §14 — performance budgets (PERF-04 relaxed for Phase 1 per D-09)
  - §15 — testing discipline (property tests, fuzz, BDD, coverage targets)
  - §16 — CI quality gates
  - §17 — release pipeline (CI-only releases, cosign keyless, Syft SBOM, OIDC attestations)
  - §18 — DX (README, godoc, llms.txt, docs/* scaffolds)
- `.planning/REQUIREMENTS.md` — 93 v1 requirements with Phase-1 traceability (38 requirements map to Phase 1).
- `.planning/PROJECT.md` — high-level scope, Core Value, Key Decisions table.
- `.planning/ROADMAP.md` Phase 1 — phase goal, dependencies, success criteria, requirements list.

### Project standards (`.claude/skills/`)
- `.claude/skills/go-coding-standards/SKILL.md` — error patterns, naming, dependency rules, **no hot-path interface boxing**, no testify in root.
- `.claude/skills/go-testing-standards/SKILL.md` — unit / property / fuzz / benchmark / BDD discipline, coverage ≥95% / ≥90% / 100%-public-API.
- `.claude/skills/determinism-standards/SKILL.md` — no map iteration on output paths, float stability, NaN/Inf/-0, golden file format, cross-platform matrix.
- `.claude/skills/performance-standards/SKILL.md` — per-algorithm budgets, ASCII fast paths, two-row DP, allocation discipline, benchstat regression (relaxed for Phase 1 per D-09).
- `.claude/skills/algorithm-licensing-standards/SKILL.md` — patent screen, no GPL/LGPL derivation, fresh-implementation discipline, attribution format. Relevant for Phase 1 because dep-allowlist additions are gated on `algorithm-licensing-reviewer` sign-off.
- `.claude/skills/commit-standards/SKILL.md` — conventional commits with issue references, no AI attribution, sign-off rules.
- `.claude/skills/documentation-standards/SKILL.md` — godoc, README, algorithm docs, llms.txt format.
- `.claude/skills/issue-standards/SKILL.md` — labels, required sections.
- `.claude/skills/algorithm-correctness-standards/SKILL.md` — primary-source citation, formula docs, reference vectors, mathematical invariants. Phase 1 reads this to set up the conventions; Phase 2 first exercises them.
- `.claude/skills/research-guidance/SKILL.md`, `.claude/skills/fuzzymatch-review-protocol/SKILL.md`.

### Research outputs
- `.planning/research/STACK.md` — version-pinned recommendations for every tool (Go 1.26.3, golangci-lint v2.12.2, GoReleaser v2.15.4, Cosign v3.0.1, godog v0.15.0, goleak v1.3.0, testify v1.10.0, gosec v2.25.0, markdownlint-cli2 v0.22.1, benchstat, `actions/checkout@v6`, `actions/setup-go@v6`, `goreleaser/goreleaser-action@v7`).
- `.planning/research/ARCHITECTURE.md` — high-level architecture choices.
- `.planning/research/PITFALLS.md` — 20 inventoried correctness pitfalls (9 infrastructure-gated, addressed by Phase 1; the rest by Phases 2–10).
- `.planning/research/FEATURES.md` — feature inventory and gaps.
- `.planning/research/SUMMARY.md` — research synthesis.

### Project conventions
- `CLAUDE.md` (project root) — Design Principle 13 (`api-ergonomics-reviewer` + `user-guide-reviewer` veto over surface shape), release discipline (CI only, no local tagging), commit hygiene.
- `BOOTSTRAP.md` — bundle-extraction guide (informational).
- `gsd-agent-skills.json` — agent + skill mapping.

</canonical_refs>

<code_context>
## Existing Code Insights

### Reusable Assets

**None — Phase 1 is greenfield Go code.** The repo currently contains only documentation (`README.md`, `CHANGELOG.md`, `CLAUDE.md`, `BOOTSTRAP.md`, `LICENSE`, `NOTICE`, `docs/requirements.md`, `docs/prior-art-research.md`), planning artifacts (`.planning/`), and review tooling (`.claude/agents/`, `.claude/skills/`, `gsd-agent-skills.json`). No `.go` files exist yet — Phase 1 lands the first Go code in the repository.

### Established Patterns

- **Mask consistency** — `axonops/mask` is the reference AxonOps Go library. Phase 1 mirrors its `.golangci.yml`, `.goreleaser.yml`, Makefile, `llms.txt` / `llms-full.txt`, `tests/bdd/` sub-module pattern, CLA Assistant workflow, commitlint config, README structure (logo + badges + TOC + status + overview + key features + why + quick start + algorithm catalogue + thread safety + configuration + tuning link + API ref link + AI assistants pointer + contributing + security + licence), and emoji section headers. Divergences from mask need rationale.
- **Apache-2.0 file headers** — every `.go` file gets the AxonOps Apache-2.0 header. `scripts/verify-license-headers.sh` is a CI gate.
- **`docs/requirements.md` illustrative code is non-binding** — code blocks there are illustrative, not authoritative. `api-ergonomics-reviewer` and `user-guide-reviewer` have veto over final surface shape per CLAUDE.md Design Principle 13.

### Integration Points

- **Downstream consumer:** `axonops/audit` (issue #853) will consume `github.com/axonops/fuzzymatch` and `github.com/axonops/fuzzymatch/scan` (and `Extract`) end-to-end in Phase 11. Phase 1's primitives (`AlgoID`, `Normalise`, `Tokenise`, sentinel errors) are the foundation that audit's identifier-similarity checks will eventually consume.
- **Sub-package boundary:** Phase 9 introduces `github.com/axonops/fuzzymatch/scan`. Phase 1 must leave the directory structure ready (root package only; `scan/` is created in Phase 9). The root package and `scan/` share no internal state — `scan/` will consume the public API of the root package.
- **Test sub-module:** `tests/bdd/go.mod` with `replace github.com/axonops/fuzzymatch => ../../` so BDD scenarios test the local module. Phase 1 ships the sub-module skeleton; actual feature files for Normalise + Tokenise may land in plans 06–07 or be deferred to Phase 2.

</code_context>

<specifics>
## Specific Ideas

- **Mirror `axonops/mask` deliberately.** Every divergence needs a documented reason. Mask is the load-bearing structural template.
- **Spec §9 wins on Normalise.** The struct-of-options shape (not variadic functional options) is explicit per-call-allocation rationale. Adding `NFC` and `StripDiacritics` to the spec §9 field set is the only additive change — defaults stay conservative (StripDiacritics=false).
- **Audit-event taxonomy use case shapes the Unicode defaults.** Consumers comparing field names that may contain diacritics (`Müller` vs `Mueller`) opt into `StripDiacritics=true` explicitly. The library does not normalise away diacritics by default — that's an explicit consumer choice.
- **`bench.txt` is local-driven and committed.** User runs `make bench-compare` on their workstation before any PR that touches algorithm code; `bench.txt` is the committed baseline. CI prints the diff informationally. PERF-04's CI-fail gate is deferred — explicit, documented compromise.
- **8 plans, strict sequential.** The user wants atomic plan commits (mask-style), narrow blast radius, easy review, and easy revert. CI/release/quality-gate plans (02/03/04) touch overlapping workflow files; parallelising is not worth the merge-conflict risk.
- **`go.mod` brittle-by-design.** Once `01-01-module-bootstrap` adds `golang.org/x/text`, no plan in Phase 1 (or any later phase without explicit dep-allowlist approval) may modify `go.mod`. `make verify-deps-allowlist` and `make tidy-check` are the enforcing gates.
- **Golden-file format locks for v1.x.** The sorted-struct + `json.MarshalIndent` choice is not a Phase 1 implementation detail — it's a v1.x stability contract. Every subsequent golden file (`algorithms.json`, `scorer-default.json`, `scan-default.json`, `extract-default.json`) inherits this format.

</specifics>

<deferred>
## Deferred Ideas

- **Self-hosted bench runner shared with `axonops/mask` + `axonops/audit`** — enables PERF-04's CI-fail gate for benchstat regressions (`>10%` at `p<0.05`). Deferred until shared infrastructure exists. Tracked here so a future phase can pick it up and re-enable the spec gate. Owner: user. Revisit trigger: when a shared self-hosted runner becomes available.
- **Commit signing on `main`** — mask doesn't require this today, so neither does fuzzymatch. Revisit if mask adopts.
- **Generated `AlgoID` enum + dispatch table** — code generation from a YAML/TOML manifest would eliminate manual sync between the constant list and the dispatch array as algorithms are added. Deferred — manual sync is fine for a curated 23-algorithm catalogue, and adding codegen complexity to Phase 1 is unjustified. Revisit if the catalogue grows beyond ~30 or third-party algorithm registration becomes a real consumer ask.
- **Distributed DX docs (docs land inside each algorithm plan)** — considered for Phase 1 slicing; rejected in favour of a single 08-dx-docs plan that writes against actual landed code. If contributors consistently forget to update docs alongside code in Phases 2+, revisit distributed-docs slicing.
- **Front-loaded DX skeleton (docs first, code fills in)** — considered; rejected because skeleton-then-fill risks doc drift. Revisit only if a strong author-experience reason emerges.

</deferred>

---

*Phase: 1-Foundation-Infrastructure*
*Context gathered: 2026-05-13*
