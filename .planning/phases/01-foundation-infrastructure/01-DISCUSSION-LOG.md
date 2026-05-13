# Phase 1: Foundation & Infrastructure - Discussion Log

> **Audit trail only.** Do not use as input to planning, research, or execution agents.
> Decisions are captured in CONTEXT.md — this log preserves the alternatives considered.

**Date:** 2026-05-13
**Phase:** 1-Foundation & Infrastructure
**Areas discussed:** Foundation API shape, Plan slicing strategy, Governance + runner choices, Golden-file & determinism scaffolding

---

## Foundation API shape

### Question 1: How should AlgoID be backed?

| Option | Description | Selected |
|--------|-------------|----------|
| Typed int const + table | `type AlgoID int` with iota constants; dispatch via [N]func switch or array indexed by AlgoID. Cheapest hot-path dispatch (no boxing, branch-predictable). Adding an algorithm = code change. | ✓ |
| Typed string + map | `type AlgoID string` with named constants; dispatch via `map[AlgoID]func`. More flexible but map lookup and iteration-determinism risk. | |
| Generated via go:generate | AlgoID enum + dispatch table generated from a YAML/TOML manifest. Eliminates manual sync but adds build-time codegen. | |

**User's choice:** Typed int const + table
**Notes:** Matches docs/requirements.md §6 and avoids hot-path interface boxing called out in go-coding-standards.

---

### Question 2: What should Normalise do by default?

| Option | Description | Selected |
|--------|-------------|----------|
| Aggressive default | `Normalise(s)` does: lowercase + strip separators + NFC + strip diacritics. Single function, opinionated. | |
| Two variants | `Normalise` (preserves diacritics) + `NormaliseAggressive` (strips). | |
| Options-based | `Normalise(s, opts ...NormaliseOption)` with `WithCaseFold`, `WithDiacriticStrip`, etc. (variadic functional options). | |

**User's choice:** Free-text — "Struct-of-options, as the spec specifies in §9 — not variadic functional options (allocates per call). NormalisationOptions struct with bool/string fields: Lowercase, StripSeparators, SeparatorChars, SplitCamelCase, NFC, StripDiacritics. DefaultNormalisationOptions() returns conservative defaults: Lowercase=true, StripSeparators=true, SeparatorChars=\"_-.:/\", SplitCamelCase=true, NFC=true, StripDiacritics=false. Consumer can opt into diacritic stripping for audit-taxonomy use cases. Matches requirements.md §9 with the NFC and StripDiacritics fields as small additive enhancements."
**Notes:** User rejected all three options and named a fourth — struct-of-options per docs/requirements.md §9. Per-call allocation was the explicit driver for rejecting variadic functional options. NFC and StripDiacritics are explicitly additive to spec §9.

---

### Question 3: What shape should Tokenise return?

| Option | Description | Selected |
|--------|-------------|----------|
| []string slice | `Tokenise(s, opts TokeniseOptions) []string`. Simplest, idiomatic, predictable allocation. Matches spec. | ✓ |
| iter.Seq[string] | `iter.Seq[string]` (Go 1.23+). Zero-alloc if not materialised; but token algorithms materialise anyway. | |
| Both — []string + Iter | `Tokenise` returns `[]string`; `TokeniseIter` returns `iter.Seq[string]`. Doubles surface. | |

**User's choice:** []string slice

---

### Question 4: How structured should the error hierarchy be?

| Option | Description | Selected |
|--------|-------------|----------|
| Flat sentinels | Package-level `var ErrInvalidInput`, `ErrInvalidConfiguration`, etc. via `errors.New`. Wrap via `fmt.Errorf("%w", ErrX)`. | ✓ |
| Typed error structs | `type InvalidInputError struct {...}` with sentinels only for category checks. | |
| Hybrid — sentinels + one typed | Flat sentinels everywhere except a `ScannerError` typed struct for Phase 9 motivation. Defer the typed struct. | |

**User's choice:** Flat sentinels
**Notes:** Composes with errors.Is/As trivially. Typed error structs deferred unless Phase 9 (scan) needs richer context.

---

### Continuation check: More questions about Foundation API shape, or move to next?

| Option | Selected |
|--------|----------|
| Next area | ✓ |
| More questions | |

---

## Plan slicing strategy

### Question 1: How should Phase 1's 38 requirements be sliced into plans?

| Option | Description | Selected |
|--------|-------------|----------|
| Granular (~7 plans) | Module bootstrap → quality gates → release pipeline → determinism infra → primitives-algoid-errors → primitives-normalise → primitives-tokenise. | ✓ |
| Coarse (3 plans) | Infrastructure → primitives → DX docs. | |
| Mega (1 plan) | Single 01-PLAN.md covering everything. | |

**User's choice:** Granular (~7 plans)

---

### Question 2: Where do the DX docs (README, CHANGELOG, llms.txt, docs/*) live in the slicing?

| Option | Description | Selected |
|--------|-------------|----------|
| Final plan 08-dx-docs | Add 08-dx-docs as the last Phase 1 plan. Lets primitives land first, then docs are written against actual code. | ✓ |
| Distributed | DX docs live inside each relevant plan. | |
| First plan 01-dx-skeleton | Land docs skeleton FIRST so subsequent plans fill them in. | |

**User's choice:** Final plan 08-dx-docs
**Notes:** Slicing therefore becomes 8 plans, not 7.

---

### Question 3: How should the executor wave-parallelise these 8 plans?

| Option | Description | Selected |
|--------|-------------|----------|
| Strict sequential | 01 → 02 → 03 → 04 → 05 → 06 → 07 → 08. No parallelism. | ✓ |
| Two parallel waves | Wave 1: 01. Wave 2: [02, 03, 04, 05]. Wave 3: [06, 07]. Wave 4: 08. | |
| Auto | Let gsd-planner derive wave structure from declared file_modified lists. | |

**User's choice:** Strict sequential
**Notes:** Phase 1 lays foundation; CI/release/quality-gate plans (02/03/04) touch overlapping workflow files. Phases 2+ may parallelise per-algorithm plans freely.

---

### Question 4: Should each plan commit its own go.mod / go.sum changes, or batch?

| Option | Description | Selected |
|--------|-------------|----------|
| Per-plan commits | Each plan that touches go.mod commits the mod change in that plan's commit. | |
| Bootstrap-only, frozen after | Only 01-module-bootstrap touches go.mod (with x/text already added). Subsequent plans add internal packages only. `make verify-deps-allowlist` + `make tidy-check` fail any subsequent go.mod change. | ✓ |

**User's choice:** Bootstrap-only, frozen after
**Notes:** Keeps the allowlist invariant brittle-by-design.

---

### Continuation check: More questions about Plan slicing strategy, or move to next?

| Option | Selected |
|--------|----------|
| Next area | ✓ |
| More questions | |

---

## Governance + runner choices

### Question 1: CLA Assistant or DCO for contributor sign-off?

| Option | Description | Selected |
|--------|-------------|----------|
| CLA Assistant | Mirror axonops/mask exactly: contributor-assistant/github-action with signatures stored in-repo (decentralised). | ✓ |
| DCO | apps/dco GitHub App. Lightest weight (Signed-off-by trailer per commit). | |

**User's choice:** CLA Assistant
**Notes:** Mask consistency lowers contributor cognitive load.

---

### Question 2: How should the benchstat regression check handle runner availability?

| Option | Description | Selected |
|--------|-------------|----------|
| Self-hosted + ubuntu fallback | Primary: shared self-hosted runner; fallback: ubuntu-latest informational only. | |
| Ubuntu-only | Always run benchstat on ubuntu-latest. Heterogeneous hardware → noisy. | |
| Self-hosted required | Self-hosted is a hard prerequisite; otherwise gate disabled. | |

**User's choice:** Free-text — "For now, I think we're just going to use the GitHub runners. I haven't got a single runner I can share. I'll deal with this later, but just for as long as I can run them on my workstation locally is the most important thing. And then I capture the benchmarks there. I'll sort out the CI one at a later date."
**Notes:** GitHub-hosted ubuntu-latest only. User runs benchmarks locally on their workstation; `bench.txt` is captured from there. PERF-04's `>10% regression fails CI` gate is RELAXED for Phase 1. Tracked in Deferred Ideas.

---

### Question 3: How should bench.txt be managed given no CI gating?

| Option | Description | Selected |
|--------|-------------|----------|
| Committed, local-driven | `bench.txt` is committed and updated manually from the user's workstation. CI runs `make bench-compare` informationally (no fail). | ✓ |
| Empty placeholder for Phase 1 | Phase 1 commits an empty `bench.txt`; first real baseline lands in Phase 2. | |
| No file in Phase 1 | Skip `bench.txt` entirely until Phase 2. | |

**User's choice:** Committed, local-driven
**Notes:** Phase 1 commits an empty/placeholder `bench.txt` with a comment explaining the local-driven workflow; Phase 2 (Levenshtein) writes the first real baseline.

---

### Question 4: Anything to add about contributor conventions — conventional-commit lint, sign-off, branch protection?

| Option | Description | Selected |
|--------|-------------|----------|
| Standard mask-style | wagoid commitlint, mask-style branch protection (require PR, all CI green, linear history, no direct pushes), CODEOWNERS, no commit signing. | ✓ |
| Looser — lint warning, not gate | commitlint warns only; merges not blocked. | |
| Stricter — add commit signing | Require signed commits on main. | |

**User's choice:** Standard mask-style

---

### Continuation check: More questions about Governance + runner choices, or move to next?

| Option | Selected |
|--------|----------|
| Next area | ✓ |
| More questions | |

---

## Golden-file & determinism scaffolding

### Question 1: What does Phase 1's placeholder golden-file test pin?

| Option | Description | Selected |
|--------|-------------|----------|
| normalisation.json only | Phase 1 lands `normalisation.json` real with ~20-40 pinned `Normalise(input) → output` entries. | ✓ |
| Synthetic placeholder | `algorithms.json` with one synthetic dummy entry. | |
| Both — normalisation real + algorithms stub | normalisation.json real AND algorithms.json with placeholder stub. | |

**User's choice:** normalisation.json only
**Notes:** Algorithm phases extend `algorithms.json` incrementally as each algorithm lands.

---

### Question 2: What format should the golden files use?

| Option | Description | Selected |
|--------|-------------|----------|
| Pretty JSON, stable key order | `json.MarshalIndent` with sorted keys at every level; UTF-8 no BOM. | |
| JSON Lines | One JSON object per line; streamable but worse for human reading. | |
| Sorted struct + JSON | Go struct with sorted entries (alphabetic by key), encoded via standard `json.MarshalIndent`. Same diff behaviour as pretty JSON but simpler generator. | ✓ |

**User's choice:** Sorted struct + JSON
**Notes:** Locks across all 5 golden files for v1.x.

---

### Question 3: Where do golden files live and how are they regenerated?

| Option | Description | Selected |
|--------|-------------|----------|
| testdata/golden/ + go test -update | Files at `testdata/golden/`. A `-update` flag regenerates them. CI runs without `-update` to diff. Standard Go idiom. | ✓ |
| testdata/golden/ + dedicated cmd | Files at testdata/golden/. A `cmd/goldengen/main.go` regenerates them. | |
| testdata/golden/ + Makefile only | Files at testdata/golden/. `make golden-update` runs a script. | |

**User's choice:** testdata/golden/ + go test -update

---

### Question 4: How should the cross-platform CI matrix exercise the golden test?

| Option | Description | Selected |
|--------|-------------|----------|
| Diff against committed | Each OS/arch runs the test without `-update`; diff against committed file; ANY divergence fails CI on that platform. | ✓ |
| Cross-platform comparison job | Each platform uploads artifact; final consolidation job diffs all. More CI complexity, no extra signal. | |
| Diff committed + tee log | Same as option A but with unified-diff printed on failure. Folded into option A as an implementation detail. | |

**User's choice:** Diff against committed
**Notes:** Implementation MUST also print a unified diff to the test log on failure for diagnosability (option C folded in).

---

### Final check: We've discussed all four areas. Which gray areas remain unclear?

| Option | Selected |
|--------|----------|
| I'm ready for context | ✓ |
| Explore more gray areas | |

---

## Claude's Discretion

- Exact concrete names for `AlgoID` constants, `NormalisationOptions` helper methods (if any), `TokeniseOptions` field set, sentinel error names beyond the spec list, Makefile target naming, and CI job names — delegated to `api-ergonomics-reviewer` and `user-guide-reviewer` per CLAUDE.md Design Principle 13.
- Specific 20–40 entries in `normalisation.json` — picked by `gsd-planner` / `test-writer` from spec normalisation examples, audit-taxonomy use cases, and Unicode normalisation reference vectors.
- Sentinel error set beyond `ErrInvalidInput`, `ErrInvalidConfiguration`, `ErrInvalidAlgorithm`, `ErrEmptyInput` — flagged to `api-ergonomics-reviewer` during planning.

## Deferred Ideas

- **Self-hosted bench runner shared with axonops/mask + axonops/audit** — enables PERF-04's CI-fail gate for benchstat regressions. Deferred until shared infrastructure exists.
- **Commit signing on `main`** — mask doesn't require this today; revisit if mask adopts.
- **Generated `AlgoID` enum + dispatch table** — code generation from a YAML/TOML manifest. Manual sync is fine for a curated 23-algorithm catalogue; revisit if catalogue grows beyond ~30 or third-party algorithm registration becomes a real ask.
- **Distributed DX docs (docs land inside each algorithm plan)** — rejected for Phase 1; revisit if contributors forget to update docs alongside code in later phases.
- **Front-loaded DX skeleton (docs first, code fills in)** — rejected (drift risk).
