# Stack Research — fuzzymatch

**Domain:** Pure-Go string-similarity library shipping academic-source algorithms with zero runtime dependencies, cross-platform determinism, BDD scenarios isolated in a sub-module, native fuzz tests, and Apache-2.0 release plumbing modelled on `axonops/mask`.
**Researched:** 2026-05-13
**Overall confidence:** HIGH — every version is verified against an authoritative source (official release notes, pkg.go.dev, GitHub releases) within the last 90 days; the architectural choices are spec-locked in `docs/requirements.md`.

---

## Spec-Locked vs. Open Decisions

Before recommending versions, here is what the spec already mandates and what is still open. The roadmap should treat spec-locked items as constraints, not options.

| Item | Status | Source |
|------|--------|--------|
| Go 1.26+ minimum | **Spec-locked** | `docs/requirements.md` §1, CLAUDE.md "Constraints" |
| Zero runtime deps (root `go.mod` has zero non-stdlib `require` lines) | **Spec-locked** | §1, §5(1), Acceptance Criteria |
| No cgo anywhere | **Spec-locked** | §1, §5(2), Acceptance Criteria |
| Apache-2.0 throughout, AxonOps file header on every `.go` | **Spec-locked** | §1, §5(9) |
| Test-only deps (godog, goleak, testify) isolated in `tests/bdd/go.mod` | **Spec-locked** | §1, §15.11 |
| **No testify in root tests** (stdlib `testing` only — stricter than mask) | **Spec-locked** | §15.11, go-coding-standards skill |
| Property tests via `testing/quick` (stdlib) | **Spec-locked** | §15.3 |
| Native Go fuzz (`go test -fuzz`) | **Spec-locked** | §15.4 |
| Cross-platform CI matrix (linux/amd64, linux/arm64, darwin/amd64, darwin/arm64, windows/amd64) | **Spec-locked** | §13.3, §17.1 |
| goreleaser + cosign keyless + GitHub OIDC attestations + Sigstore | **Spec-locked** | §17.1 (release.yml) |
| `golangci-lint` + `go vet` + `markdownlint-cli2` + `govulncheck` + `gosec` + CodeQL | **Spec-locked** | §17.1 |
| CI-only releases — no local tagging | **Spec-locked** | CLAUDE.md "Releases", §17.2 |
| Dependabot | **Spec-locked** | §17.4 |
| benchstat regression > 10% fails CI | **Spec-locked** | §6(6), §14.4 |
| **Patch version of Go 1.26** (1.26.0 vs 1.26.3) | Open | Spec says "1.26+"; recommendation below |
| **golangci-lint v1 vs v2 configuration** | Open | Spec says "matching axonops/mask"; recommendation: v2 with migration |
| **goreleaser v1 vs v2** | Open | Spec doesn't pin; recommendation: v2 |
| **CLA vs DCO** | Spec-leaning-CLA | §17.1 mentions CLA; mask uses CLA Assistant. Recommendation: mirror mask. |
| **Conventional-commit lint tool** (Go-native vs Node `commitlint`) | Open | Recommendation: `wagoid/commitlint-github-action` (mask's pattern) |

---

## Recommended Stack

### Core Language and Toolchain

| Technology | Version | Purpose | Why Recommended |
|------------|---------|---------|-----------------|
| **Go** | **1.26.3** (`go.mod` directive: `go 1.26.3`; `toolchain` line omitted or pinned to `go1.26.3`) | Compiler, stdlib, `testing/quick`, `go test -fuzz`, `go test -race` | Spec-locked at 1.26+ to match `axonops/audit` and `axonops/mask`. 1.26.3 (released 2026-05-07) is the current security-patch release and includes fixes to `crypto/x509`, `html/template`, `net`, `net/http`, the compiler, linker, and runtime. The previously-experimental Green Tea GC is on by default in 1.26 — beneficial for the short-string allocation patterns this library uses. `testing/quick` (stdlib) covers all property-test needs without external libraries; native fuzz (since 1.18) covers all fuzz-test needs. |
| **CGO_ENABLED** | `0` (enforced in CI for all builds) | Pure-Go builds | Spec-locked. `CGO_ENABLED=0 go build ./...` is the canonical verification that no transitive dep slips in cgo. Build matrix runs with `CGO_ENABLED=0` on every platform. |

**Confidence: HIGH** (Go 1.26.3 release notes verified; ≤ 1 week old).

### Test Stack — Root Module (`go.mod`, zero non-stdlib `require` lines)

| Technology | Version | Purpose | Why Recommended |
|------------|---------|---------|-----------------|
| **`testing`** (stdlib) | bundled with Go 1.26.3 | Unit tests, table-driven tests | Spec-locked: no testify in root. The verbosity trade-off is accepted per §15.11. Use `t.Errorf`/`t.Fatalf` with helper functions in `internal_test_helpers.go` if patterns repeat. |
| **`testing/quick`** (stdlib) | bundled with Go 1.26.3 | Property-based tests (`PropAlgorithmScore_RangeBounds`, `PropAlgorithmScore_Identity`, `PropAlgorithmScore_Symmetric`, `PropEditDistance_TriangleInequality`, etc.) | Spec-locked. Stdlib; zero deps. Sufficient for the mathematical invariants the library guarantees. |
| **`go test -fuzz`** (native, stdlib) | bundled with Go 1.26.3 | Coverage-guided fuzz tests (one `Fuzz*` per public function — 26 fuzzers total) | Spec-locked. Since Go 1.18, native fuzzing is the standard; corpus checks into `testdata/fuzz/`. CI runs 60s per fuzzer per build; nightly runs 5 min per fuzzer. |
| **`go test -race -shuffle=on -count=1`** | bundled with Go 1.26.3 | Race detector, deterministic shuffling, no result caching | Spec-locked in §17.1. `-shuffle=on` surfaces test-order dependencies (important because the library promises pure functions). `-count=1` defeats result caching for accurate signal in CI. |

**Confidence: HIGH** (all stdlib, all spec-locked).

### Test Stack — BDD Sub-Module (`tests/bdd/go.mod`, isolated)

| Technology | Version | Purpose | Why Recommended |
|------------|---------|---------|-----------------|
| **`github.com/cucumber/godog`** | **v0.15.0** (latest stable; verify with `go list -m -versions github.com/cucumber/godog` before pinning) | Gherkin/Cucumber BDD scenarios per algorithm and per Scorer/scan pattern | Spec-locked in §15.6. Official Cucumber framework for Go. Active maintenance; integrates with `go test`. Lives only in `tests/bdd/go.mod` so consumers never transitively depend on it. Feature files in `tests/bdd/features/{algorithms,scorer,normalisation,determinism,scan,suppression}.feature`. |
| **`go.uber.org/goleak`** | **v1.3.0** (current stable) | Goroutine leak detection via `goleak.VerifyTestMain(m)` in BDD `TestMain` | Spec-locked in §15.9. The library is pure-function with no goroutines; goleak catches regressions if anyone introduces background work. Note: goleak supports only the two most recent minor Go versions — Go 1.26 satisfies this. |
| **`github.com/stretchr/testify`** | **v1.10.0** (current stable; v1 is the supported track — no breaking changes accepted) | Assertion sugar (`assert.Equal`, `require.NoError`) inside `tests/bdd/steps/` step definitions | Permitted ONLY in `tests/bdd/`. Root tests forbid it. `testify` is the most widely-used Go assertion library and is the path of least resistance for BDD step definitions where assertion verbosity hurts readability. |

**Confidence: HIGH** (godog and goleak in active use across Go ecosystem; testify v1.10.0 is the current stable on the v1 line per stretchr/testify maintainers' statement that v1 is the supported track).

### Build & Quality Tooling

| Tool | Version | Purpose | Notes |
|------|---------|---------|-------|
| **`golangci-lint`** | **v2.12.2** (released 2026-05-06) | Aggregator linter — runs `gofmt`/`gofumpt`/`goimports` (now under `formatters:`), `govet`, `errcheck`, `staticcheck`, `revive`, `gocyclo`, `gosec` (optionally), etc. | Use **v2 configuration** (`version: "2"` at top of `.golangci.yml`). The v2 config replaces `enable-all`/`disable-all` with `linters.default: standard|fast|all|none` and moves formatters out of `linters:` into a dedicated `formatters:` section. Run `golangci-lint migrate` to convert legacy configs. Mask's `.golangci.yml` is the structural template; adapt to v2 layout. CI uses `golangci/golangci-lint-action@v8` (or the matching current major). |
| **`go vet`** | bundled with Go 1.26.3 | Static analysis | Run separately from golangci-lint to keep the signal cleanly attributed; some teams disable `govet` in golangci-lint and run `go vet ./...` as its own CI step. |
| **`gofmt -s`** | bundled with Go 1.26.3 | Canonical formatting | Run via `make fmt-check` as a CI gate; under golangci-lint v2 this lives in the `formatters:` block. |
| **`goimports`** | `golang.org/x/tools/cmd/goimports@latest` (pin a specific commit in CI) | Import grouping (stdlib → `github.com/axonops/fuzzymatch/...`) | Under golangci-lint v2, configured in `formatters:`. |
| **`govulncheck`** | **`golang.org/x/vuln/cmd/govulncheck@latest`** (latest release published 2026-04-22; pin a specific tag in CI for reproducibility) | Vulnerability scanning against Go's vulnerability DB | Spec-locked in §17.1 (vulncheck step) and §17.1 (security.yml weekly). Note: known issue #77670 was a transient bug after Go 1.26.0; resolved in current releases. Run on PR + weekly. Fails CI on any HIGH/CRITICAL. |
| **`gosec`** | **`github.com/securego/gosec/v2@v2.25.0`** (released 2026-03-19) | Security linter — 50+ rules mapped to OWASP Top 10 / CWE; SARIF output to GitHub Security tab | Spec-locked in §17.1 (security.yml). Use the official `securego/gosec` GitHub Action. SARIF upload to GitHub code-scanning UI is the standard pattern. |
| **`benchstat`** | **`golang.org/x/perf/cmd/benchstat@latest`** (modern rewrite by Austin Clements, 2023) | A/B benchmark comparison with confidence intervals | Spec-locked in §14.4. CI compares `bench.txt` (committed) against current run; > 10% regression on any benchmark fails. Use self-hosted runner shared with mask/audit for hardware consistency; fall back to `ubuntu-latest` informationally if unavailable. |
| **`markdownlint-cli2`** | **v0.22.1** (current stable) | Markdown linting for README, docs/, CHANGELOG | Spec-locked in §17.1. Configure via `.markdownlint-cli2.yaml` (YAML supported). Mask uses this; mirror its rule set. |
| **CodeQL** | **`github/codeql-action@v4`** (v3 deprecated end-of-2026) with bundle `2.25.3+` (Go 1.26 supported as of bundle 2.24.2, Feb 2026) | Semantic security analysis on push to `main` + weekly schedule | Spec-locked in §17.1 (codeql.yml). Use `init` + `analyze` actions; build mode `autobuild` works for pure Go. |
| **Dependabot** | GitHub-native (no version) | Daily checks for `gomod` (root + `tests/bdd`) and `github-actions` ecosystems | Spec-locked in §17.4. Group PRs: indirect / direct / test-only / actions. Auto-merge patch updates that pass CI. |

**Confidence:** golangci-lint v2.12.2 HIGH (release page verified 2026-05-06); govulncheck HIGH (latest published 2026-04-22); gosec v2.25.0 HIGH (release 2026-03-19); markdownlint-cli2 v0.22.1 HIGH; benchstat HIGH; CodeQL v4 / Go 1.26 support HIGH (GitHub changelog 2026-02-24).

### Release Stack

| Tool | Version | Purpose | Notes |
|------|---------|---------|-------|
| **GoReleaser** | **v2.15.4** (released 2026-04-21; pin via `version: "~> v2"` in the action) | Release automation: cross-compile, archive, checksum, GitHub Release publishing, SBOM hooking, signing hooking | Spec-locked in §17.1 (release.yml). For a **library** (no binaries to ship), GoReleaser still runs to: (1) generate `checksums.txt` for the source tarball, (2) trigger cosign signing on the checksum, (3) publish the GitHub Release with the auto-extracted CHANGELOG section. If the library acquires a CLI in future, GoReleaser handles cross-compile then too. Configuration in `.goreleaser.yml` (or `.goreleaser.yaml`). |
| **`goreleaser/goreleaser-action`** | **v7** (current major; defaults to `version: "~> v2"`) | GitHub Action wrapper for GoReleaser | Pin `distribution: goreleaser` and `version: "~> v2"` (do not use `latest`). |
| **Sigstore Cosign** | **v3.0.1** (current stable; `COSIGN_EXPERIMENTAL` env var removed — keyless is now default) | Keyless signing of `checksums.txt` via `cosign sign-blob --yes` + GitHub OIDC | Spec-locked in §17.1. Sign in `release.yml` via `cosign/sign-blob` with `--oidc-issuer=https://token.actions.githubusercontent.com`. Signature + certificate uploaded as release assets. Note: `--bundle` is now required in v3 (moved from optional). |
| **`sigstore/cosign-installer`** | latest action major (verify before pinning) | Installs cosign in CI | Provides the `cosign` binary; verified against released checksums automatically. |
| **GitHub Artifact Attestation** | GitHub-native (`actions/attest-build-provenance@v2` or current major) | OIDC build-provenance attestation per release | Spec-locked in §17.1 ("GitHub Actions OIDC build provenance attestation"). Native GitHub feature; no extra deps. |
| **Syft** | latest stable (current ~v1.42; uses GoReleaser's built-in SBOM hook) | SPDX-JSON SBOM generation for the source archive | Apache-2.0 licensed, by Anchore. GoReleaser's `sboms:` block invokes `syft` by default. SBOM published as a release asset; signed by cosign alongside checksums. |
| **`actions/checkout`** | **v6** (current major; uses Node 24 runtime per the Sept 2025 deprecation of Node 20) | Repo checkout in CI | Use `fetch-depth: 0` in the release job (GoReleaser needs full history for changelog extraction). |
| **`actions/setup-go`** | **v6** (current major; Node 24 runtime) | Install Go toolchain | Pin `go-version-file: go.mod` (preferred) or `go-version: 'stable'` (looser). Caches modules and build cache automatically. |
| **`wagoid/commitlint-github-action`** | latest stable | Conventional-commit linting on PR titles + commits | Configure via `.commitlintrc.yml` at repo root with `@commitlint/config-conventional` rules. Run on `pull_request` events. Alternative Go-native: `github.com/conventionalcommit/commitlint` (CLI) — viable but adds a Node-free toolchain at the cost of a slightly less polished GitHub Action experience. Recommendation: use `wagoid/commitlint-github-action` to match the mainstream ecosystem; the Node runtime is in CI only, not in the library. |
| **`contributor-assistant/github-action`** (CLA Assistant Lite) | latest stable | CLA signing workflow via PR comment | Mask uses CLA Assistant; mirror that. Signatures stored in repo (decentralised). Allowlist Dependabot + bots. **Decision flag:** mask uses CLA exclusively (not DCO). Recommendation: same for fuzzymatch. If the user prefers DCO instead, the recommendation switches to `apps/dco` (GitHub App, simpler, no signing files). |

**Confidence:** GoReleaser v2.15.4 HIGH (release 2026-04-21); goreleaser-action v7 HIGH; cosign v3.0.1 HIGH; syft Apache-2.0 HIGH; checkout/setup-go v6 HIGH (Node 24 transition verified Sept 2025 deprecation notice).

### Docs Stack

| Technology | Version | Purpose | Why Recommended |
|------------|---------|---------|-----------------|
| **godoc / pkg.go.dev** | n/a — generated from source | Canonical Go API reference | Spec-locked in §16.4. Every exported symbol has a godoc starting with the symbol name. Every algorithm file cites its primary source at the top. Runnable examples in `example_test.go` appear on pkg.go.dev. CI pings the pkg.go.dev proxy on release for prompt indexing. |
| **Runnable examples (`example_test.go`)** | bundled with Go | godoc examples that compile + execute + verify expected output | Spec-locked in §16.5. One example per algorithm + Scorer + Normalise + Tokenise. Meta-test `readme_shop_front_test.go` ensures README quick-start matches an example. |
| **`llms.txt` + `llms-full.txt`** | Markdown, no specific version (specification is an evolving informal convention proposed by Jeremy Howard / Answer.AI) | AI-assistant-friendly documentation index | Spec-locked in §16.3 and mirrored from mask. `llms.txt` is a concise index; `llms-full.txt` embeds full content. Verified in sync with code via `ai_friendly_test.go` (parses `go/ast` and asserts every exported symbol is listed). Note: the format is **not yet a finalised standard** — the project follows it because it is the emerging convention and mask already uses it. Track the spec at https://llmstxt.org/ for changes. |
| **README** | Markdown | Project front door | Spec-locked structure in §16.1: logo, badges (CI, Go Reference, Go Report Card, License, Status), TOC, status (pre-release orange), overview, key features, why, quick start, algorithm catalogue table, Scorer composition, thread safety, configuration, tuning link, API reference link, AI assistants pointer, contributing, security, licence. Mirror mask layout including the emoji section headers. |
| **CHANGELOG** | Markdown — **Keep a Changelog** format | Per-release change record | Spec-locked in §16.6 (acceptance criterion: "CHANGELOG following Keep-a-Changelog"). `release.yml` extracts the relevant section for the GitHub Release description. |
| **`docs/*.md`** | Markdown, project-specific | algorithms.md, scorer.md, scan.md, tuning.md, extending.md, performance.md, faq.md | Spec-locked in §16.2. `docs/requirements.md` is the authoritative spec; the others are user-facing. |

**Confidence:** godoc/pkg.go.dev HIGH; Keep-a-Changelog HIGH (established convention); llms.txt MEDIUM (convention is real but evolving — verify llmstxt.org spec at release time).

---

## Installation Recipes

### Developer machine (one-time)

```bash
# Go toolchain
# (use whichever installer is canonical for the platform — go.dev/dl, asdf, mise, brew, etc.)
go version  # expect: go1.26.3

# Quality tools (versioned installs)
go install github.com/golangci/golangci-lint/v2/cmd/golangci-lint@v2.12.2
go install golang.org/x/vuln/cmd/govulncheck@latest
go install github.com/securego/gosec/v2/cmd/gosec@v2.25.0
go install golang.org/x/perf/cmd/benchstat@latest

# Docs lint
npm install -g markdownlint-cli2@0.22.1
```

### Repo bootstrap (one-time when scaffolding Phase 1)

```bash
# Root module
cd fuzzymatch
go mod init github.com/axonops/fuzzymatch
# Note: this creates go.mod with `go 1.26.3` and zero require lines.
# Verify with: ! grep -q '^require' go.mod

# BDD sub-module
mkdir -p tests/bdd
cd tests/bdd
go mod init github.com/axonops/fuzzymatch/tests/bdd
go get github.com/cucumber/godog@latest
go get go.uber.org/goleak@latest
go get github.com/stretchr/testify@latest
```

The BDD sub-module's `go.mod` will have a `replace github.com/axonops/fuzzymatch => ../..` directive so step definitions can import the library under test directly. This keeps consumers off the BDD deps because the directive is local-only.

### CI verification (every PR)

```bash
# Zero-deps verification (the canonical structural check)
scripts/verify-no-runtime-deps.sh
# Implementation: grep -q '^require' go.mod && exit 1 || exit 0
# (Or use 'go list -m all | grep -v ^github.com/axonops/fuzzymatch' returning only stdlib markers.)

# No-cgo verification
CGO_ENABLED=0 go build ./...

# Full quality gate
make check
```

---

## Alternatives Considered

| Recommended | Alternative | When to Use Alternative |
|-------------|-------------|-------------------------|
| **`testing` (stdlib) in root** | `stretchr/testify` in root | Never for this project — spec-locked. Testify is permitted only in `tests/bdd/`. Trade-off accepted in §15.11. |
| **`testing/quick` (stdlib)** | `pgregory.net/rapid`, `leanovate/gopter` | If property tests need shrinking or richer generators. `testing/quick` is sufficient for the invariants this library checks (symmetry, identity, range bounds, triangle inequality). Adding rapid/gopter would require a test-only dep in `tests/bdd/`; not justified by current scope. Revisit at v1.x if property tests grow in complexity. |
| **Native Go fuzz** | `dvyukov/go-fuzz`, ClusterFuzzLite, OSS-Fuzz | Native fuzz is sufficient and stdlib. OSS-Fuzz integration may be valuable post-v1.0.0 for continuous fuzzing at scale — flag for later phase. |
| **`golangci-lint` v2** | `golangci-lint` v1.x | v1 is end-of-line. v2 is the current major and includes the `linters.default`/`formatters` split. New projects should start on v2. |
| **GoReleaser v2** | GoReleaser v1.26 (last v1) | v1.26 is the final v1 release. v2 is the same minus deprecated options; new projects should start on v2 with `version: "~> v2"` pinning. |
| **Cosign keyless via Fulcio + Rekor** | Self-managed keys (KMS, hardware) | Keyless is the modern default (since `COSIGN_EXPERIMENTAL` removal in v3). Self-managed keys add operational burden and offer no integrity benefit for an OSS library. Use keyless. |
| **CLA Assistant Lite (mask's pattern)** | DCO (`apps/dco` GitHub App) | DCO is simpler operationally (a Signed-off-by trailer), CLA is heavier but matches mask. Recommendation: mirror mask (CLA). Flag this for explicit user confirmation — switching to DCO is a one-line decision but it diverges from mask. |
| **`wagoid/commitlint-github-action` (Node)** | `conventionalcommit/commitlint` (Go-native CLI) | Go-native avoids Node in CI but the Action is less polished. Recommendation: Node wagoid action — Node runs in CI only, not in the library. |
| **Self-hosted benchmark runner (mask/audit shared)** | `ubuntu-latest` runner | Shared self-hosted runner gives stable hardware → meaningful `benchstat` numbers. Spec-locked in §14.4. If unavailable, fall back to `ubuntu-latest` informationally with regression detection skipped. |
| **`actions/setup-go@v6`** | `actions/setup-go@v5` | v5 still works but uses Node 20 runtime which is deprecated. v6 uses Node 24. New projects start on v6. |
| **markdownlint-cli2** | `markdownlint-cli` (v1) | v2 is faster and config-driven; mask uses it; mirror. |

---

## What NOT to Use (and Why)

| Avoid | Why | Use Instead |
|-------|-----|-------------|
| **cgo / any C bindings** | Spec-locked prohibition (§1, §5(2)). Breaks portability, complicates cross-compile, blocks pure-Go consumers. Even a transitive cgo dep would break the no-cgo guarantee. | Pure-Go stdlib only. Verify with `CGO_ENABLED=0 go build ./...`. |
| **Any non-stdlib runtime dependency in root `go.mod`** | Spec-locked (§1, §17.1 verify-no-runtime-deps.sh). Any external runtime dep introduces supply-chain risk, transitive cgo risk, licence-compatibility risk, and version-conflict risk for consumers. | Implement everything from primary academic sources. ~1,400 lines of pure Go covers all 23 algorithms per `docs/prior-art-research.md`. |
| **`testify` in root `_test.go` files** | Stricter than mask (§15.11). Even test-only deps in the root module pollute downstream consumer modules under certain `go test` invocations and complicate the "zero runtime deps" story. | Stdlib `testing` in root. `testify` permitted only in `tests/bdd/`. Verbosity accepted. |
| **`init()` functions doing non-trivial work** | Determinism risk (init order is platform-sensitive); package-load side effects break the pure-function guarantee. Spec-locked (§5(12), §13.5). | Initialise tables (e.g. Strcmp95 similar-character matrix) via `var x = ...` literal expressions. |
| **Map iteration on any output path** | Go map iteration is randomised. Any function that builds output from map iteration produces non-deterministic ordering, breaking the determinism guarantee (§13.4, determinism-standards skill). | Extract keys, sort them, iterate the sorted slice. `Warning` slices use `sort.SliceStable` with a complete sort key. |
| **`math.Pow`, `math.Log`, `math.Exp`, `math.FMA`** in algorithm hot paths | Float-determinism risk across platforms (linux/arm64 vs windows/amd64). Spec-locked (§13.3, determinism-standards skill). `math.Sqrt` is IEEE-754-correctly-rounded and IS allowed; transcendentals are not. | Use addition, subtraction, multiplication, division, and `math.Sqrt`/`math.Abs`/`math.Min`/`math.Max` only. Flag any other `math.X` call for determinism-reviewer. |
| **Parallel/reduction sum operations** | Order of float addition affects the result (associativity violation). Reduces determinism across architectures with different vector widths. | Sum slices left-to-right with a single loop; never `sync/atomic` floats or chunked parallel reductions. |
| **Local `git tag` invocations** | Spec-locked (§17.2, CLAUDE.md "Releases"). Bypasses CI's reproducibility, signing, attestation. devops agent flags this as BLOCKING. | Releases happen exclusively via CI on tag push from a release PR merge or a GitHub Release UI action. Maintainers never run `goreleaser release` locally. |
| **`--no-verify` on commits** | Bypasses pre-commit hooks, conventional-commit lint, sign-off. CLAUDE.md "Releases" implicitly forbids. | Fix the hook failure properly; never skip. |
| **Patent-encumbered algorithms** (Metaphone 3 / U.S. Patent 7440941) | Apache-2.0 hygiene; AxonOps declines patent encumbrance regardless of enforcement (§4, algorithm-licensing-standards skill). | Double Metaphone + NYSIIS cover the phonetic use cases. Patent screen every algorithm via algorithm-licensing-reviewer BEFORE implementation. |
| **GPL/LGPL-derived code** | Licence incompatibility with Apache-2.0 distribution. Spec-locked (§5(9)). | Fresh implementation from primary academic sources. MIT/BSD reference Go implementations may be CONSULTED for reference-vector cross-validation but never copied. |
| **`init()`-time table builds** | Determinism: init order is platform-dependent. | `var table = buildTable()` or `var table = [...]int{...}` literal. |
| **`golangci-lint` v1 config** | v1 is end-of-line; v2 is the current major. | `version: "2"` in `.golangci.yml`; run `golangci-lint migrate` if porting. |
| **`goreleaser-action@v6` or lower** | Node 20 deprecation; mismatch with goreleaser v2. | `goreleaser/goreleaser-action@v7` with `version: "~> v2"`. |
| **`actions/checkout@v4` or `actions/setup-go@v5`** | Node 20 deprecation. | `@v6` for both. |
| **Goroutines in library code** | Spec-locked: pure-function library, no concurrency primitives, no channels. `goleak` in BDD will catch regressions. | Synchronous code only. |
| **`Algorithm` interface for dispatch** | Interface allocations on hot paths violate per-algorithm allocation budgets. Spec-locked in go-coding-standards. | Typed `AlgoID` enum + switch dispatch. |
| **Hidden allocations on short ASCII inputs** | Spec-locked performance budgets (§14.1). String-to-byte conversion, slice growth, map writes all allocate. | Stack-allocated buffers for ≤ 50-char ASCII; ASCII fast paths in every algorithm where applicable; `-benchmem` to verify. |

---

## Stack Patterns by Variant

### When implementing a character-based algorithm (Levenshtein, Damerau-Levenshtein, Hamming, Jaro, Jaro-Winkler, Strcmp95, Smith-Waterman-Gotoh, LCSStr)

- **Test deps:** stdlib `testing` only in root.
- **Property tests:** `testing/quick` covering identity, symmetry, range bounds, triangle inequality (for distance-based).
- **Fuzz:** one `Fuzz*` per public function, seed corpus from literature reference vectors.
- **ASCII fast path:** `[]byte` operations + `unicode.IsASCII` gate; rune variant only if input contains non-ASCII.
- **DP optimisation:** two-row DP for O(mn) algorithms; stack-allocated `[100]int` buffer for short inputs.
- **Determinism:** no `math.X` beyond Sqrt/Abs/Min/Max; sum reductions left-to-right.
- **Benchmark:** `_bench_test.go` at 10/50/200/500 chars; commit to `bench.txt` per release.

### When implementing a q-gram algorithm (Q-Gram Jaccard, Sørensen-Dice, Cosine, Tversky)

- **Shared q-gram extraction** in `q_gram.go` to avoid duplication.
- **Map usage:** internally fine for frequency counting; never iterate maps on output paths.
- **Allocations:** ≤ 4 per call per §14.1 budget; small-input fast path can reduce to 0 in v1.x.

### When implementing a token-based algorithm (Monge-Elkan, Token Sort/Set/Partial Ratio, Token Jaccard)

- **Tokeniser:** shared `tokenise.go` handling snake_case, camelCase, kebab-case, dot-case.
- **Inner-metric dispatch** for Monge-Elkan via `AlgoID` enum + switch — no `Metric` interface.

### When implementing a phonetic algorithm (Soundex, Double Metaphone, NYSIIS, MRA)

- **Fresh implementation from primary source** (Knuth 1973 for Soundex, Philips 2000 for Double Metaphone, Taft 1970 for NYSIIS, NBS Tech Note 943 for MRA).
- **Patent screen first** via algorithm-licensing-reviewer.
- **Tables via `var`, not `init()`** (§13.5).

### When designing the Scorer / scan layer

- Immutable after construction; pure functions; safe for concurrent use without locks.
- `ScoreAll` returns a `map[string]float64` — map contents deterministic, iteration order is not (§13.4 documented).
- `scan.Check` output sorted by `(Kind, NameA, NameB, GroupA, GroupB)` with `sort.SliceStable`.
- Token-bucket optimisation in scan: property test proves equivalence to naive O(N²).

---

## Version Compatibility

| Component A | Compatible With | Notes |
|-------------|-----------------|-------|
| **Go 1.26.3** | golangci-lint v2.12.2 | golangci-lint v2 supports Go 1.22+; 1.26 fully supported. |
| **Go 1.26.3** | goleak v1.3.0 | goleak supports the two most recent minor Go versions; 1.26 and 1.25 are both supported. |
| **Go 1.26.3** | CodeQL bundle 2.25.3 | CodeQL Go 1.26 support landed in bundle 2.24.2 (Feb 2026). |
| **Go 1.26.3** | govulncheck (latest) | Resolved post-1.26 issue #77670 confirmed fixed in current releases. |
| **GoReleaser v2.15.4** | goreleaser-action v7 | v7 defaults to `"~> v2"`; older v6 of the action also worked but is superseded. |
| **GoReleaser v2.15.4** | Cosign v3.0.1 | Sign verification action requires GoReleaser v2.13.0+; v2.15.4 satisfies. |
| **Cosign v3.0.1** | GitHub OIDC (`token.actions.githubusercontent.com`) | Keyless default in v3; `--bundle` now required. |
| **godog v0.15.0** | Go 1.26.3 | Compatible; lives in `tests/bdd/go.mod`. |
| **testify v1.10.0** | Go 1.26.3 | Supports Go 1.19+. v1 is the supported track. |
| **markdownlint-cli2 v0.22.1** | Node.js | Requires Node 20+ in CI runner (provided by `actions/setup-node@v4`). |
| **`actions/checkout@v6`** | Node 24 runner | Requires runner version ≥ v2.327.1. GitHub-hosted runners auto-upgraded. |

---

## Phase-Specific Stack Notes for the Roadmap

| Phase (from §19) | Stack additions needed |
|------------------|------------------------|
| **Phase 1 — Bootstrap** | All quality + release tooling scaffolding lands here: `.golangci.yml` (v2), `.goreleaser.yml` (v2), Makefile, CI workflows (ci, nightly, release, security, codeql), Dependabot, CODEOWNERS, issue/PR templates, CLA workflow, commitlint workflow, `scripts/verify-no-runtime-deps.sh`, `scripts/verify-license-headers.sh`, `scripts/verify-llms-sync.sh`. **No production Go code yet.** |
| **Phase 2 — Core algorithms (v0.1.0 / v0.2.0)** | `testing` + `testing/quick` + native fuzz exercised for the first time. `bench.txt` first committed. Cross-platform determinism golden file first populated. |
| **Phase 3 — Q-gram and token-based (v0.3.0)** | Shared `q_gram.go` extraction; map-iteration discipline tested heavily. |
| **Phase 4 — Phonetic (v0.4.0)** | algorithm-licensing-reviewer is critical — patent screen Double Metaphone (Philips 2000), NYSIIS, MRA before implementation. |
| **Phase 5 — Scorer (v0.5.0)** | Functional-options pattern lands. BDD scenarios in `tests/bdd/features/scorer.feature` first written — godog + goleak + testify enter the BDD sub-module. |
| **Phase 6 — Scan sub-package (v0.6.0)** | Property test `PropCheck_BucketEquivalentToNaive`; deterministic sort key verified. |
| **Phase 7 — Integration shakedown** | Downstream consumer (axonops/audit #853) imports `fuzzymatch` — confirms zero-runtime-deps in practice. |
| **Phase 8 — v1.0.0 stable** | First signed release via cosign keyless; SBOM published; OIDC attestation attached. |

---

## Sources

### Authoritative (Official Docs / Release Pages)

- [Go 1.26 Release Notes](https://go.dev/doc/go1.26) — version baseline, Green Tea GC default
- [Go 1.26 is released (blog)](https://go.dev/blog/go1.26) — 1.26.0 release announcement
- [Go release history](https://go.dev/doc/devel/release) — 1.26.3 patch release info
- [Microsoft Go 1.26.2-1 / 1.25.9-1 builds](https://devblogs.microsoft.com/go/go-1-26-2-1-and-1-25-9-1-microsoft-builds-now-available/) — patch cadence
- [golangci-lint Configuration File](https://golangci-lint.run/docs/configuration/file/) — v2 config format
- [golangci-lint Migration Guide](https://golangci-lint.run/docs/product/migration-guide/) — v1 → v2 migration
- [golangci-lint Changelog](https://golangci-lint.run/docs/product/changelog/) — v2.12.2 release 2026-05-06
- [Welcome to golangci-lint v2](https://ldez.github.io/blog/2025/03/23/golangci-lint-v2/) — v2 design rationale
- [Announcing GoReleaser v2](https://goreleaser.com/blog/goreleaser-v2/) — v2 release rationale
- [GoReleaser releases](https://github.com/goreleaser/goreleaser/releases) — v2.15.4 (2026-04-21)
- [GoReleaser SBOMs](https://goreleaser.com/customization/sbom/) — syft integration
- [GoReleaser GitHub Actions](https://goreleaser.com/customization/ci/actions/) — action-v7 + `~> v2`
- [GoReleaser supply-chain example](https://github.com/goreleaser/example-supply-chain) — canonical cosign + SBOM + attestation pattern
- [Sigstore cosign](https://github.com/sigstore/cosign) — v3 keyless default
- [Cosign CHANGELOG](https://github.com/sigstore/cosign/blob/main/CHANGELOG.md) — v3.0.1 details
- [govulncheck on pkg.go.dev](https://pkg.go.dev/golang.org/x/vuln/cmd/govulncheck) — latest release 2026-04-22
- [golang/vuln releases](https://github.com/golang/vuln/releases)
- [gosec releases](https://github.com/securego/gosec/releases) — v2.25.0 (2026-03-19)
- [godog on pkg.go.dev](https://pkg.go.dev/github.com/cucumber/godog) — official Cucumber for Go
- [godog repo](https://github.com/cucumber/godog)
- [goleak repo](https://github.com/uber-go/goleak) — v1.3.0; supports two most recent Go minors
- [testify repo](https://github.com/stretchr/testify) — v1.10.0 current stable; v1 supported track
- [benchstat](https://pkg.go.dev/golang.org/x/perf/cmd/benchstat) — modern A/B comparison
- [markdownlint-cli2](https://github.com/DavidAnson/markdownlint-cli2) — v0.22.1
- [actions/checkout](https://github.com/actions/checkout/releases) — v6 current
- [actions/setup-go](https://github.com/actions/setup-go/releases) — v6 current
- [goreleaser-action](https://github.com/goreleaser/goreleaser-action) — v7 current
- [CodeQL Action repo](https://github.com/github/codeql-action) — v4 current; v3 deprecating end of 2026
- [CodeQL Go 1.26 support changelog](https://github.blog/changelog/2026-02-24-codeql-adds-go-1-26-and-kotlin-2-3-10-support-and-improves-query-accuracy/)
- [Anchore syft](https://github.com/anchore/syft) — Apache-2.0; SBOM standard
- [llms.txt spec](https://llmstxt.org/) — evolving convention
- [wagoid/commitlint-github-action](https://github.com/wagoid/commitlint-github-action)
- [contributor-assistant/github-action (CLA Assistant Lite)](https://github.com/contributor-assistant/github-action)
- [Go Fuzzing tutorial](https://go.dev/doc/tutorial/fuzz) — native fuzz baseline
- [Go Vulnerability Management](https://go.dev/doc/security/vuln/)

### Project-Internal (Confidence: HIGH — spec-locked)

- `docs/requirements.md` §1, §4, §5, §13, §14, §15, §17 — the authoritative spec
- `docs/prior-art-research.md` — Go ecosystem survey
- `CLAUDE.md` — constraints, release discipline, agent gates
- `.planning/PROJECT.md` — high-level context and constraints
- `.claude/skills/go-coding-standards/SKILL.md` — error patterns, naming, dependency rules
- `.claude/skills/determinism-standards/SKILL.md` — no map iteration, float stability, golden files
- `.claude/skills/research-guidance/SKILL.md` — what to research, what is settled

### Confirmatory (cross-referenced, MEDIUM-HIGH)

- [axonops/mask](https://github.com/axonops/mask) — structural template: Go 1.26+, CLA via CLA Assistant, signed commits enforced, `.goreleaser.yml` + `.golangci.yml` + Makefile + `llms.txt` + `llms-full.txt` + `tests/bdd/` + `docs/` directory pattern

---

*Stack research for: fuzzymatch (pure-Go string-similarity library)*
*Researched: 2026-05-13*
