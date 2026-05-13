---
plan_id: 01-04
phase: 01-foundation-infrastructure
plan: 04
type: execute
wave: 4
depends_on:
  - 01-03
autonomous: true
objective: >
  Land the determinism + benchmark infrastructure: cross-platform CI matrix
  (linux/amd64+arm64, darwin/amd64+arm64, windows/amd64), runtime-deps
  allowlist script, golden-file canonical-form helper (which LOCKS the v1.x
  format), benchstat scaffolding with informational CI per D-09, and the
  Makefile glue for `make verify-deps-allowlist`, `make verify-determinism`,
  `make bench`, `make bench-compare`. Every subsequent algorithm/Scorer/scan
  /Extract phase inherits this scaffolding.
files_modified:
  - .github/workflows/ci.yml
  - scripts/verify-no-runtime-deps.sh
  - scripts/verify-coverage-floors.sh
  - testdata/golden/.gitkeep
  - golden_test.go
  - golden_canonical.go
  - golden_canonical_test.go
  - bench.txt
  - Makefile
requirements:
  - DET-01
  - DET-03
  - PERF-04
  - PERF-06
  - CI-05
  - CI-06
  - TEST-07
must_haves:
  truths:
    - "CI ci.yml runs the test matrix across linux/amd64, linux/arm64, darwin/amd64, darwin/arm64, windows/amd64 with `CGO_ENABLED=0`"
    - "Every platform in the matrix independently diffs `testdata/golden/*.json` against the committed file (D-14)"
    - "`scripts/verify-no-runtime-deps.sh` asserts `go list -m all` minus stdlib is exactly `golang.org/x/text` and nothing else"
    - "Golden-file canonical format is LOCKED: sorted-struct + `json.MarshalIndent(v, \"\", \"  \")` + `\"\\n\"` line endings + UTF-8 no BOM (D-12)"
    - "`golden_canonical.go`'s canonical-form helper is exercised by a dedicated unit test asserting CRLF-free / BOM-free / sorted output"
    - "`bench.txt` exists (placeholder); `make bench` writes `bench.txt.new`; `make bench-compare` runs benchstat and emits diff; CI `make bench-compare` is INFORMATIONAL only (`continue-on-error: true`) per D-09"
    - "Coverage floors (≥95% overall, ≥90% per file, 100% public API) enforced by `scripts/verify-coverage-floors.sh`"
    - "No `init()`-time table builds (convention enforced by determinism-reviewer; plan documents the rule)"
    - "No map iteration on output paths (DET-03; first exercised in plan 01-06)"
  artifacts:
    - path: .github/workflows/ci.yml
      provides: 5-platform matrix with CGO_ENABLED=0 and informational benchstat
      contains: "matrix:"
    - path: scripts/verify-no-runtime-deps.sh
      provides: Runtime-deps allowlist gate (root go.mod = x/text only)
    - path: scripts/verify-coverage-floors.sh
      provides: Coverage floor enforcement (95/90/100)
    - path: golden_test.go
      provides: `var updateGolden = flag.Bool("update", false, ...)` flag and the harness loader; `TestGolden_*` table-driven shell
    - path: golden_canonical.go
      provides: `canonicalMarshal` helper that produces the LOCKED v1.x byte format
    - path: golden_canonical_test.go
      provides: Unit test asserting canonicalMarshal produces sorted, BOM-free, LF-terminated UTF-8 JSON
    - path: bench.txt
      provides: Empty placeholder; populated by user's `make bench` + manual commit
  key_links:
    - from: Makefile
      to: scripts/verify-no-runtime-deps.sh
      via: `verify-deps-allowlist` target invokes the script
      pattern: "verify-no-runtime-deps.sh"
    - from: .github/workflows/ci.yml
      to: testdata/golden/
      via: matrix steps run `make verify-determinism` which runs `TestGolden_*`
      pattern: "verify-determinism"
    - from: golden_test.go
      to: golden_canonical.go
      via: TestGolden_* uses canonicalMarshal to produce the bytes diffed against testdata/golden/*.json
      pattern: "canonicalMarshal"
---

<objective>
Land the determinism + benchmark + dep-allowlist infrastructure so:
  1. CI proves byte-identical golden output across 5 platforms on every PR.
  2. Root go.mod can NEVER pick up a non-x/text runtime dep (CI gate).
  3. The golden-file canonical-form helper LOCKS the v1.x output format.
  4. Coverage floors are enforceable.
  5. Benchstat infrastructure exists locally; CI runs it informationally
     per D-09 until a self-hosted runner is available.

Purpose: Every algorithm/Scorer/scan/Extract plan from Phase 2 onwards
inherits this scaffolding. The format LOCK in `golden_canonical.go` is a
v1.x stability contract.

Output:
  - 5-platform matrix in ci.yml
  - scripts/verify-no-runtime-deps.sh
  - scripts/verify-coverage-floors.sh
  - golden_test.go (harness, no real test cases yet)
  - golden_canonical.go (the LOCKED format)
  - golden_canonical_test.go (unit test on the helper)
  - bench.txt (placeholder)
  - Makefile updates filling in the pending targets
</objective>

<execution_context>
@$HOME/.claude/get-shit-done/workflows/execute-plan.md
@$HOME/.claude/get-shit-done/templates/summary.md
</execution_context>

<context>
@.planning/PROJECT.md
@.planning/REQUIREMENTS.md
@.planning/phases/01-foundation-infrastructure/01-CONTEXT.md
@.planning/research/STACK.md
@.planning/research/PITFALLS.md
@CLAUDE.md
@.claude/skills/determinism-standards/SKILL.md
@.claude/skills/performance-standards/SKILL.md
@.claude/skills/go-coding-standards/SKILL.md
@.claude/skills/go-testing-standards/SKILL.md
@.github/workflows/ci.yml
@Makefile
@go.mod
@doc.go

<interfaces>
Locked decisions inherited from CONTEXT.md:
- D-12: Golden-file canonical format = sorted-struct + `json.MarshalIndent(v, "", "  ")` + `"\n"` line endings + UTF-8 no BOM. LOCKED for v1.x.
- D-13: `testdata/golden/` + `-update` flag for regen. CI runs without `-update`; any divergence fails.
- D-14: Each platform in CI matrix diffs independently against the single committed file. Per-platform diff IS the cross-platform check.
- D-09: `bench.txt` local-driven + committed; CI runs `make bench-compare` informationally only (`continue-on-error: true`). PERF-04's CI-fail gate (>10% regression at p<0.05) is RELAXED for Phase 1.

CI matrix shape (D-14):
  matrix:
    os: [ubuntu-latest, ubuntu-24.04-arm, macos-latest, macos-13, windows-latest]
    # ubuntu-latest        → linux/amd64
    # ubuntu-24.04-arm     → linux/arm64 (GitHub-hosted Arm Linux runner GA mid-2025)
    # macos-latest         → darwin/arm64 (M-series)
    # macos-13             → darwin/amd64 (Intel; macOS 14+ no longer ships amd64 GHR images)
    # windows-latest       → windows/amd64

  Note: if `ubuntu-24.04-arm` is unavailable to the org at execution time,
  fall back to QEMU emulation via `uraimo/run-on-arch-action@v3` for the
  arm64 step. Document the fallback in the SUMMARY.

Canonical-marshal helper signature (lives in `golden_canonical.go`, internal
but exported for testdata-rewriting purposes — package fuzzymatch):

  // canonicalMarshal serialises v to the LOCKED v1.x golden-file canonical
  // form: sorted-struct ordering enforced by the caller (the function does
  // not sort maps; callers MUST pass an already-sorted slice-of-struct),
  // json.MarshalIndent(v, "", "  "), trailing `"\n"`, UTF-8 no BOM. The
  // returned bytes are stable across linux/darwin/windows and across patch
  // versions of Go (uses only encoding/json's deterministic emitters).
  //
  // Concrete name (per api-ergonomics-reviewer at execution time): the
  // function name may be `CanonicalMarshal` (exported) or `canonicalMarshal`
  // (unexported with `_test.go`-only export shim). The default proposal:
  // unexported `canonicalMarshal` + an exported `WriteGoldenFile(path string,
  // v any) error` wrapper that the `-update` flag invokes.
  func canonicalMarshal(v any) ([]byte, error)

Golden-test flag and helper (lives in `golden_test.go`):

  // updateGolden, if set, rewrites every testdata/golden/*.json file from
  // the current code output instead of asserting equality. Used as
  // `go test -run TestGolden_ -update`. CI runs without this flag.
  var updateGolden = flag.Bool("update", false, "rewrite testdata/golden files")

  // assertGolden marshals v via canonicalMarshal and compares it to the
  // bytes at testdata/golden/<filename>. On -update, rewrites the file
  // instead.
  func assertGolden(t *testing.T, filename string, v any) { ... }

TestGolden_* shell tests in plan 01-04 exist but do not exercise any real
data — they are placeholders. Plan 01-06 populates the first real
`TestGolden_Normalisation` against `testdata/golden/normalisation.json`.
</interfaces>
</context>

<tasks>

<task type="auto">
  <name>Task 1: Write verify-no-runtime-deps.sh and verify-coverage-floors.sh; wire into Makefile</name>
  <files>scripts/verify-no-runtime-deps.sh, scripts/verify-coverage-floors.sh, Makefile</files>
  <read_first>
    - .planning/phases/01-foundation-infrastructure/01-CONTEXT.md (D-07 — go.mod freeze; CI-05 — verify-deps-allowlist; TEST-07 — coverage floors 95/90/100)
    - .claude/skills/go-coding-standards/SKILL.md (Dependencies — zero non-stdlib runtime deps except x/text)
    - .claude/skills/go-testing-standards/SKILL.md (Coverage target ≥95% overall, ≥90% per file, 100% public API)
    - go.mod (the freeze baseline — verify against this)
    - Makefile (from plan 01-02 — `verify-deps-allowlist` and `coverage-check` targets currently print "pending plan 01-04")
    - scripts/verify-license-headers.sh (plan 01-01 — pattern reference for bash conventions)
  </read_first>
  <action>
    Write two shell scripts and update the Makefile to invoke them definitively
    (no longer print "pending plan 01-04").

    1. `scripts/verify-no-runtime-deps.sh`:
       - Shebang `#!/usr/bin/env bash`
       - `set -euo pipefail`
       - Header comment explaining: "Exits 0 if root go.mod's resolved runtime
         dependency graph contains exactly one non-stdlib module
         (golang.org/x/text). Exits 1 with a diff otherwise. Verified via
         `go list -m -mod=mod all` filtered to non-stdlib modules."
       - Implementation:
         ```
         allowed=("github.com/axonops/fuzzymatch" "golang.org/x/text")
         # Get all modules in the build list
         mapfile -t actual < <(go list -m -mod=mod all 2>/dev/null | awk '{print $1}')
         # Filter to exclude allowed
         bad=()
         for mod in "${actual[@]}"; do
           ok=0
           for a in "${allowed[@]}"; do
             [[ "$mod" == "$a" ]] && ok=1 && break
             [[ "$mod" == "$a/"* ]] && ok=1 && break  # sub-packages of allowed
           done
           [[ $ok -eq 0 ]] && bad+=("$mod")
         done
         if (( ${#bad[@]} > 0 )); then
           echo "FAIL: unexpected runtime dependencies in root go.mod:" >&2
           printf '  %s\n' "${bad[@]}" >&2
           echo "" >&2
           echo "Allowed list: ${allowed[*]}" >&2
           exit 1
         fi
         echo "OK: root go.mod allowlist clean (${#actual[@]} modules: only fuzzymatch + golang.org/x/text)"
         ```
       - Note: `golang.org/x/text` is the parent module; sub-packages
         (`golang.org/x/text/unicode/norm`, etc.) are reported by `go list -m`
         under the parent module name and so already pass the equality check.
         The `$a/*` glob handles future cases where Go's module tooling
         reports a sub-path (defensive).
       - `chmod +x scripts/verify-no-runtime-deps.sh`.

    2. `scripts/verify-coverage-floors.sh`:
       - Shebang `#!/usr/bin/env bash`
       - `set -euo pipefail`
       - Usage: expects `coverage.out` (the profile) to exist at the repo root
         (produced by `make coverage`).
       - Three checks:
         a. Overall floor ≥ 95.0%:
            Parse `go tool cover -func=coverage.out` last line (`total:`),
            extract the percentage, compare against 95.0 using `awk` or `bc`.
         b. Per-file floor ≥ 90.0%:
            Iterate every file in `go tool cover -func=coverage.out` output
            (filenames in column 1; per-function rows aggregate via filename
            grouping). For each file: assert ≥ 90.0%. Files with no statements
            (e.g. `doc.go`) are exempt (skip filenames whose function rows
            all show 0.0% / 0 statements).
         c. Public API floor = 100%:
            Run `go doc -all .` and extract the list of exported identifiers
            (matching `^(func|type|var|const) [A-Z]`). For each exported
            symbol, assert there exists at least one coverage row showing
            non-zero coverage for that symbol. If any exported symbol has
            0% coverage, fail.
            Note: 100% public-API coverage is enforced by the EXISTENCE of
            an exercising test, not necessarily a coverage of 100.0% on
            the symbol's body. Document this in the script comment block.
       - Output format on failure: `FAIL: <metric> = <observed>% < <required>%`
         to stderr, followed by the offending filenames or symbol names.
       - `chmod +x scripts/verify-coverage-floors.sh`.

    3. Update `Makefile`:
       - `verify-deps-allowlist:` — drop the "pending plan 01-04" fallback;
         invoke `bash scripts/verify-no-runtime-deps.sh` definitively.
       - `coverage-check:` — invoke `bash scripts/verify-coverage-floors.sh
         coverage.out` (passing the profile path explicitly).
       - The `check` aggregate now runs these definitively.

    Concrete identifiers:
      - File `scripts/verify-no-runtime-deps.sh`, allowlist
        `["github.com/axonops/fuzzymatch", "golang.org/x/text"]`.
      - File `scripts/verify-coverage-floors.sh`, floors 95.0 / 90.0 / 100.0.
      - Coverage profile path: `coverage.out`.
  </action>
  <verify>
    <automated>test -x scripts/verify-no-runtime-deps.sh &amp;&amp; bash scripts/verify-no-runtime-deps.sh &amp;&amp; test -x scripts/verify-coverage-floors.sh &amp;&amp; make verify-deps-allowlist &amp;&amp; make coverage &amp;&amp; make coverage-check</automated>
  </verify>
  <acceptance_criteria>
    - `scripts/verify-no-runtime-deps.sh` exists, is executable, exits 0 on the current tree
    - Removing `golang.org/x/text` from the allowlist (or temporarily adding a fake dep) causes the script to exit 1 with a diff printed to stderr (executor should perform this validation interactively and restore)
    - `scripts/verify-coverage-floors.sh` exists and is executable
    - `scripts/verify-coverage-floors.sh` exits 0 when there are no .go files with bodies (doc.go has no statements; the script exits 0 cleanly OR documents the no-statements case as "PASS: no measurable files yet" exit 0)
    - `make verify-deps-allowlist` exits 0
    - `make coverage` exits 0 (produces coverage.out — even if mostly empty)
    - `make coverage-check` exits 0 on the current tree
    - The Makefile's `verify-deps-allowlist` target no longer contains "pending plan 01-04"
  </acceptance_criteria>
  <done>
    Both verification scripts exist and are executable; Makefile invokes them
    definitively (no more pending messages); the root module's x/text-only
    allowlist is enforced in CI and locally.
  </done>
</task>

<task type="auto">
  <name>Task 2: Write golden_canonical.go + golden_canonical_test.go (the LOCKED format)</name>
  <files>golden_canonical.go, golden_canonical_test.go, golden_test.go, testdata/golden/.gitkeep</files>
  <read_first>
    - .planning/phases/01-foundation-infrastructure/01-CONTEXT.md (D-12 — sorted-struct + json.MarshalIndent + LF + UTF-8 no BOM, LOCKED for v1.x; D-13 — `-update` flag; D-14 — per-platform diff)
    - .claude/skills/determinism-standards/SKILL.md (Golden Files section; Float Stability section)
    - .claude/skills/go-coding-standards/SKILL.md (Files — every .go gets the AxonOps Apache-2.0 header; doc on every exported symbol)
    - .claude/skills/go-testing-standards/SKILL.md (Meta-tests section — golden tests live under standard `go test`)
    - doc.go (plan 01-01 — example of the file-header form)
    - go.mod (the module path)
  </read_first>
  <action>
    Create three Go files locking the golden-file canonical format and the
    test-harness flag, plus a unit test on the canonical helper. Every file
    starts with the AxonOps Apache-2.0 header.

    1. `golden_canonical.go` (`package fuzzymatch`):
       - File header (Apache-2.0).
       - File-level block comment documenting that this file LOCKS the v1.x
         golden-file canonical format and that ANY change to its output bytes
         requires a major version bump per `docs/requirements.md` §11.2.
       - Exported (or test-only-exported) helper:
         `func canonicalMarshal(v any) ([]byte, error)` (or `CanonicalMarshal`
         — final exported-ness decided by api-ergonomics-reviewer at
         execution time. The default proposal: unexported with an exported
         `WriteGoldenFile(path string, v any) error` wrapper).
         Behaviour:
           a. Run `json.MarshalIndent(v, "", "  ")` (two-space indent — D-12).
           b. Append `"\n"` (single LF — D-12).
           c. Return bytes. NO BOM, NO CRLF, NO `, ` swap to `,\n` (the
              standard MarshalIndent already emits `,\n` on a new line; the
              comma + indent placement is encoding/json's responsibility and
              has been stable since Go 1.0).
           d. UTF-8 output is implicit (`encoding/json` produces UTF-8).
         Implementation note: this helper does NOT sort maps. Callers must
         pass an already-sorted slice-of-struct or a struct with explicit
         field order. Document this loud in the godoc.
       - Public wrapper `WriteGoldenFile(path string, v any) error`:
           a. Compute bytes via canonicalMarshal.
           b. Write to `path` via `os.WriteFile(path, bytes, 0o644)`.
           c. Return any error.
           Used only by `-update` flag — not invoked in production code.
         The wrapper is exported so that test helpers in `golden_test.go` can
         call it; it is NOT exported as part of the library's consumer API
         (the godoc explicitly states "intended for test maintenance only").
       - NO `init()` function. NO map iteration. NO `math.X` beyond `Sqrt`/`Abs`/`Min`/`Max` (and this file uses none).

    2. `golden_canonical_test.go` (`package fuzzymatch_test`):
       - File header (Apache-2.0).
       - Unit tests on canonicalMarshal:
         a. `TestCanonicalMarshal_ProducesSortedStructOutput` — pass a
            struct with explicitly-ordered fields, assert the output JSON
            has the keys in the declared struct order (json.MarshalIndent
            preserves struct field declaration order — this is the contract
            we depend on).
         b. `TestCanonicalMarshal_TrailingNewline` — assert the output ends
            with exactly one `'\n'` byte and no `'\r'`.
         c. `TestCanonicalMarshal_NoBOM` — assert the output's first 3 bytes
            are NOT `\xef\xbb\xbf` (UTF-8 BOM).
         d. `TestCanonicalMarshal_TwoSpaceIndent` — assert the output uses
            exactly two-space indentation (`grep -E '^  [^ ]'` matches; no
            tab characters; no four-space sequences).
         e. `TestCanonicalMarshal_StableAcrossCalls` — call canonicalMarshal
            twice on the same input, assert bytes are identical (determinism
            within a single process).
         f. `TestCanonicalMarshal_NoTrailingWhitespace` — assert no line
            ends with `' '` or `'\t'` followed by `'\n'`.

    3. `golden_test.go` (`package fuzzymatch_test`):
       - File header (Apache-2.0).
       - `flag.Bool("update", false, "rewrite testdata/golden files")` package
         variable named `updateGolden`.
       - `assertGolden(t *testing.T, filename string, v any)` helper:
         a. Compute bytes via fuzzymatch.canonicalMarshal (or whatever the
            exported wrapper is named per api-ergonomics-reviewer).
         b. If `*updateGolden`: call `fuzzymatch.WriteGoldenFile(filename, v)`
            and `t.Logf("updated %s", filename)`; return.
         c. Else: `want, err := os.ReadFile(filename)`. If err: `t.Fatalf`.
            Diff `got` vs `want` byte-by-byte. On mismatch print a unified
            diff fragment (line-oriented) to the test log and `t.Fail()`.
       - Placeholder test `TestGolden_Bootstrap`:
         a. Creates a tiny in-memory struct, marshals it via canonicalMarshal,
            asserts the output is non-empty. This test exists to demonstrate
            the harness works and is exercised by `make verify-determinism`
            from this plan forward. It does NOT diff against any committed
            file (because none exists yet — plan 01-06 lands the first one).

    4. `testdata/golden/.gitkeep` — already exists from plan 01-01. Confirm
       it's tracked.

    Concrete identifiers:
      - File `golden_canonical.go`, helper name `canonicalMarshal`
        (unexported) + `WriteGoldenFile` (exported wrapper).
      - File `golden_canonical_test.go`, 6 unit tests as listed.
      - File `golden_test.go`, flag name `updateGolden`, helper
        `assertGolden`, placeholder test `TestGolden_Bootstrap`.
      - Canonical-form contract: two-space indent, single trailing LF, no
        BOM, no CRLF, deterministic byte-for-byte across runs.

    Note (per CLAUDE.md Design Principle 13): exact naming of
    `canonicalMarshal` / `WriteGoldenFile` / `assertGolden` may be adjusted
    by api-ergonomics-reviewer at execution time. The structure (two-space
    indent, single trailing LF, no BOM, deterministic) is the v1.x LOCK and
    is non-negotiable.
  </action>
  <verify>
    <automated>go build ./... &amp;&amp; go vet ./... &amp;&amp; go test -race -run 'TestCanonicalMarshal_|TestGolden_Bootstrap' ./... &amp;&amp; make verify-license-headers</automated>
  </verify>
  <acceptance_criteria>
    - `golden_canonical.go` exists with the AxonOps Apache-2.0 header
    - `golden_canonical.go` declares `package fuzzymatch`
    - `golden_canonical.go` contains a `canonicalMarshal` (or equivalent) function returning `([]byte, error)`
    - `golden_canonical.go` invokes `json.MarshalIndent(v, "", "  ")` (two-space indent)
    - `golden_canonical.go` appends a single `"\n"` to the output
    - `golden_canonical.go` has NO `init()` function
    - `golden_canonical.go` performs no map iteration on output
    - `golden_canonical_test.go` exists with 6 unit tests covering: sorted-struct output, trailing newline, no BOM, two-space indent, stability across calls, no trailing whitespace
    - `golden_test.go` declares the `updateGolden` flag
    - `golden_test.go` declares the `assertGolden` helper
    - `golden_test.go` contains `TestGolden_Bootstrap` exercising the harness
    - `go build ./...` exits 0
    - `go test ./...` exits 0 (the new tests pass)
    - `make verify-license-headers` exits 0 (all new .go files carry the header)
    - Running `go test -run TestCanonicalMarshal_NoBOM ./...` exits 0
    - Producing a JSON output via canonicalMarshal on a simple `map[string]int{"a":1, "b":2}` results in bytes ending in `}\n` (single LF) with no `\r` and no BOM
  </acceptance_criteria>
  <done>
    Golden-file canonical format is LOCKED in `golden_canonical.go`; six unit
    tests verify the format; the `-update` flag and `assertGolden` helper
    exist in `golden_test.go`; placeholder `TestGolden_Bootstrap` exercises
    the harness. The v1.x format contract is now in code.
  </done>
</task>

<task type="auto">
  <name>Task 3: Expand ci.yml to 5-platform matrix; create bench.txt placeholder; wire bench/bench-compare targets definitively</name>
  <files>.github/workflows/ci.yml, bench.txt, Makefile</files>
  <read_first>
    - .planning/phases/01-foundation-infrastructure/01-CONTEXT.md (D-09 — bench informational only; D-14 — per-platform diff)
    - .planning/research/STACK.md (actions/checkout@v6, setup-go@v6; benchstat; cross-platform matrix in CI)
    - .planning/research/PITFALLS.md (cross-platform float determinism risk — informs the matrix coverage)
    - .claude/skills/determinism-standards/SKILL.md (Cross-Platform CI Matrix section)
    - .claude/skills/performance-standards/SKILL.md (Benchstat Regression Detection section — informational fallback)
    - .github/workflows/ci.yml (plan 01-02 — current single-platform shape)
    - Makefile (plan 01-02 + Task 1 update — current `bench` and `bench-compare` targets)
  </read_first>
  <action>
    Three changes:

    1. EXPAND `.github/workflows/ci.yml` matrix on the `quality` job to 5
       platforms:
       ```
       strategy:
         fail-fast: false
         matrix:
           include:
             - { os: ubuntu-latest,      arch: amd64, label: "linux-amd64" }
             - { os: ubuntu-24.04-arm,   arch: arm64, label: "linux-arm64" }
             - { os: macos-13,           arch: amd64, label: "darwin-amd64" }
             - { os: macos-latest,       arch: arm64, label: "darwin-arm64" }
             - { os: windows-latest,     arch: amd64, label: "windows-amd64" }
       runs-on: ${{ matrix.os }}
       env:
         CGO_ENABLED: "0"
       ```
       Notes:
         - `ubuntu-24.04-arm` is GitHub-hosted Arm64 Linux (GA mid-2025);
           if unavailable to the org at execution time, replace with the
           QEMU fallback documented under <interfaces>. Record the chosen
           runner in the SUMMARY.
         - `macos-13` is the last Intel macOS GHR image; needed for
           darwin/amd64 coverage. If `macos-13` is decommissioned at
           execution time, fall back to the latest Intel-available
           macos-N tag. Record in SUMMARY.
         - `CGO_ENABLED: "0"` is a workflow-level env to enforce the no-cgo
           guarantee on every matrix entry.
       After matrix expansion, each entry runs the SAME steps from plan
       01-02 (checkout → setup-go → install tools → `make check`). Add one
       additional step BEFORE `make check`:
         `name: verify determinism`
         `run: make verify-determinism`
       This runs `TestGolden_*` independently per platform, comparing the
       current code's canonicalMarshal output against the committed
       `testdata/golden/*.json` files. Any divergence on ANY platform fails
       that platform's job (D-14).

       Add an INFORMATIONAL bench job (separate from the matrix):
         `name: bench-compare-informational`
         `runs-on: ubuntu-latest`
         `continue-on-error: true`  (per D-09 — informational only)
         steps:
           a. checkout@v6, setup-go@v6, install benchstat
           b. `make bench` (writes bench.txt.new)
           c. `make bench-compare` (runs benchstat against committed
              bench.txt; prints diff to the workflow log)
         The job posts its output to the PR check summary but does NOT block
         merge. When a self-hosted runner becomes available (Deferred Item
         in CONTEXT.md), this becomes a blocking gate.

    2. Create `bench.txt` placeholder at the repo root:
       - First line: `# benchstat baseline for github.com/axonops/fuzzymatch`
       - Second line: `# Populated by `make bench` from a local workstation per CONTRIBUTING (D-09)`
       - Third line: `# Empty until the first algorithm benchmark lands in Phase 2`
       - File is otherwise empty (benchstat tolerates empty input gracefully).

    3. Update Makefile `bench` and `bench-compare` targets:
       - `bench:` no longer needs the no-op fallback — keep the same shape,
         since `go test -bench=. -benchmem -count=10 ./...` produces no
         output rows when there are no benchmarks (Go reports "PASS" lines
         only; benchstat reads empty input cleanly).
       - `bench-compare:` — drop the "pending plan 01-04" fallback. Invoke
         `benchstat bench.txt bench.txt.new` directly; if benchstat is not
         installed locally, print the install hint and exit 0 (tolerant
         locally; CI installs benchstat).
       - Update the inline comment block at the top of the Makefile noting
         that `bench-compare` is INFORMATIONAL per D-09 until the
         self-hosted-runner deferred item is resolved.

    Concrete identifiers:
      - File `.github/workflows/ci.yml`
      - Matrix list: ubuntu-latest, ubuntu-24.04-arm, macos-13, macos-latest, windows-latest
      - Env `CGO_ENABLED: "0"` (workflow-level)
      - Determinism step name `verify determinism`
      - Bench job name `bench-compare-informational`
      - Bench job `continue-on-error: true`
      - File `bench.txt` (header comments only)
  </action>
  <verify>
    <automated>python3 -c "import yaml; cfg=yaml.safe_load(open('.github/workflows/ci.yml')); jobs=cfg['jobs']; quality=jobs['quality']; m=quality['strategy']['matrix']; entries=m.get('include', []); names=[e.get('label','') for e in entries]; assert 'linux-amd64' in names and 'linux-arm64' in names and 'darwin-amd64' in names and 'darwin-arm64' in names and 'windows-amd64' in names, f'missing platforms: {names}'; print('matrix OK:', names)" &amp;&amp; grep -q 'CGO_ENABLED' .github/workflows/ci.yml &amp;&amp; grep -q 'verify-determinism\|make verify-determinism' .github/workflows/ci.yml &amp;&amp; test -f bench.txt &amp;&amp; make bench-compare</automated>
  </verify>
  <acceptance_criteria>
    - `.github/workflows/ci.yml` has a 5-entry matrix covering linux-amd64, linux-arm64, darwin-amd64, darwin-arm64, windows-amd64
    - The matrix `include:` array contains exactly 5 entries with the expected `os:` values
    - `CGO_ENABLED: "0"` appears at the workflow or job env level
    - Every matrix job runs `make verify-determinism` before `make check`
    - There is a separate `bench-compare-informational` job with `continue-on-error: true`
    - `bench.txt` exists at the repo root with at least the header comment line
    - `make bench-compare` exits 0 (no-op tolerant for empty bench.txt vs bench.txt.new)
    - `python3 -c "import yaml; yaml.safe_load(open('.github/workflows/ci.yml'))"` exits 0 (valid YAML)
    - `actionlint` (if installed) reports no errors on ci.yml
    - `make check` still exits 0 locally
    - Removing the trailing LF from a hypothetical committed `testdata/golden/<file>.json` and re-running `make verify-determinism` would produce a non-zero exit (the harness asserts byte-equality including the trailing LF — verifiable when plan 01-06 lands real golden data)
  </acceptance_criteria>
  <done>
    CI matrix covers all 5 mandated platforms with CGO_ENABLED=0; each
    platform independently runs `make verify-determinism`; the
    bench-compare-informational job runs on PRs but does not block;
    bench.txt is committed as the empty baseline; Makefile invokes
    bench-compare definitively.
  </done>
</task>

</tasks>

<threat_model>
## Trust Boundaries

| Boundary | Description |
|----------|-------------|
| Cross-platform CI matrix → golden-file equality | Each platform diffs independently against the single committed file. A platform-dependent float/byte divergence would fail CI on that platform, surfacing the issue. |
| Future code change → golden-file format | The format LOCK in `golden_canonical.go` makes inadvertent format drift (e.g. tab indentation instead of two-space) a visible CI failure on the canonical-marshal unit tests. |
| Dependency proxy → root go.mod | `scripts/verify-no-runtime-deps.sh` is the gate; any PR that introduces a non-x/text runtime dep fails CI, even from Dependabot. |
| Local benchmark runner → CI benchstat job | CI benchstat is INFORMATIONAL (D-09); no hardware-noise-induced false-positive can block PRs. When a self-hosted runner exists, the gate becomes blocking (deferred). |

## STRIDE Threat Register

| Threat ID | Category | Component | Disposition | Mitigation Plan |
|-----------|----------|-----------|-------------|-----------------|
| T-01-04-01 | Tampering | Golden-file format drift across patch versions | mitigate | `golden_canonical_test.go` asserts the byte-level contract (two-space indent, single LF, no BOM, no CRLF, no trailing whitespace); any change to the helper that breaks the contract fails the unit tests; `make verify-determinism` re-checks across all 5 platforms. |
| T-01-04-02 | Tampering | Cross-platform output divergence (line endings, BOM, FMA float emissions) | mitigate | 5-platform CI matrix with CGO_ENABLED=0; per-platform `make verify-determinism` runs `TestGolden_*` against the same committed file; any divergence on ANY platform fails that platform's job. |
| T-01-04-03 | Tampering | Non-x/text runtime dep slipping into root go.mod | mitigate | `scripts/verify-no-runtime-deps.sh` runs in `make check` locally and on every PR via CI; allowlist is `[fuzzymatch, golang.org/x/text]` only. |
| T-01-04-04 | Information Disclosure | Coverage regression hiding untested code paths | mitigate | `scripts/verify-coverage-floors.sh` enforces 95.0/90.0/100.0 floors; runs in `make check` and CI. |
| T-01-04-05 | Repudiation | Bench-regression noise blocking PRs | accept | Per D-09, bench-compare CI job is `continue-on-error: true` until a self-hosted runner exists. Trade-off explicitly documented; revisited when shared infrastructure is available. |
| T-01-04-06 | Tampering | A malicious commit subverting the canonicalMarshal helper to silently emit non-canonical bytes | mitigate | 6 unit tests on canonicalMarshal validate the byte-level contract (sorted-struct output, trailing newline, no BOM, two-space indent, stability, no trailing whitespace). Any subversion that produces different bytes fails these tests. |
| T-01-04-07 | Denial of Service | Pathological test input slowing the matrix | accept | `make verify-determinism` only runs `TestGolden_*` (small, fixed input set). Algorithm-level pathological-input concerns belong to algorithm plans (Phase 2+). |
</threat_model>

<verification>
1. `make verify-deps-allowlist` exits 0.
2. `make coverage-check` exits 0.
3. `make verify-determinism` exits 0 (TestGolden_Bootstrap runs).
4. `make bench-compare` exits 0 (empty bench.txt vs empty bench.txt.new).
5. `go test -race -run 'TestCanonicalMarshal_|TestGolden_' ./...` exits 0.
6. `.github/workflows/ci.yml` matrix lists all 5 platforms and sets CGO_ENABLED=0.
7. `make check` exits 0 locally on a clean tree.
</verification>

<success_criteria>
- 5-platform CI matrix with CGO_ENABLED=0 enforced.
- `scripts/verify-no-runtime-deps.sh` is the runtime-deps allowlist gate.
- `scripts/verify-coverage-floors.sh` enforces 95/90/100 floors.
- Golden-file canonical format is LOCKED in code (golden_canonical.go)
  and verified by 6 unit tests.
- `-update` flag and `assertGolden` helper exist for future plans.
- `bench.txt` placeholder is committed; `make bench-compare` runs locally
  and informationally in CI.
- Plan 01-04's contract makes plan 01-06's golden file `normalisation.json`
  cleanly droppable into the harness.
</success_criteria>

<output>
After completion, create
`.planning/phases/01-foundation-infrastructure/01-04-determinism-infra-SUMMARY.md`
recording:
  - The exact runner names selected for each matrix entry (e.g.
    `ubuntu-24.04-arm` vs the QEMU fallback if unavailable).
  - The exact canonical-marshal helper name (per api-ergonomics-reviewer
    decision at execution time).
  - Any deviations from D-12's byte contract (none expected; record nothing
    if none).
  - Whether `bench.txt` was populated by an initial local run (likely
    empty until Phase 2's first benchmark).
</output>
