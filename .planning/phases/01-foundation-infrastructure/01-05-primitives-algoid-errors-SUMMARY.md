---
phase: 01-foundation-infrastructure
plan: 05
subsystem: api
tags: [algoid, dispatch-table, sentinel-errors, errors.Is, iota-enum, testing/quick]

# Dependency graph
requires:
  - phase: 01-foundation-infrastructure
    provides: root go.mod freeze + Apache-2.0 header convention + verifier (plan 01-01); make check / lint / fmt-check / coverage-check pipeline (plan 01-02); coverage-floor gate + canonicalMarshal infrastructure (plan 01-04)
provides:
  - algoid.go — `type AlgoID int` + 23 iota-backed constants in spec catalogue order (docs/requirements.md §7)
  - `(AlgoID).String() string` — switch-based, zero-allocation in-range hot path, fmt.Sprintf("AlgoID(%d)") fallback for out-of-range
  - `AlgoIDs() []AlgoID` — fresh slice literal per call, deterministic order, no map iteration
  - unexported `numAlgorithms = int(AlgoRatcliffObershelp) + 1` (compile-time 23) sizing the dispatch array
  - unexported `dispatch [numAlgorithms]func(a, b string) float64` — Phase 1 ships an all-nil skeleton; Phase 2+ algorithm plans populate entries
  - errors.go — 4 flat sentinel errors via errors.New (ErrInvalidInput, ErrInvalidConfiguration, ErrInvalidAlgorithm, ErrEmptyInput) with the `fuzzymatch: ` prefix, lowercase, no trailing punctuation
  - 17 unit/property tests across algoid_test.go (10 incl. benchmark) and errors_test.go (6) pinning the public-API contract
  - export_test.go extensions — NumAlgorithmsForTest / DispatchLenForTest / DispatchEntryNilForTest re-exports keep the dispatch skeleton linker-live without polluting the public API
affects: 01-06-primitives-normalise (errors.go is the canonical sentinel vocabulary), 01-07-primitives-tokenise (same), 01-08-dx-docs (godoc anchors and llms.txt enumerate the AlgoID constants), 02+ algorithm phases (each algorithm registers itself into dispatch[AlgoX] at package load), 08-scorer (consumes dispatch via package-internal helper + ErrInvalidAlgorithm), 10-extract (same)

# Tech tracking
tech-stack:
  added: []  # No runtime deps added — only first production Go code in the root package
  patterns:
    - "Typed AlgoID enum + iota-backed dense-from-zero contiguous block: type AlgoID int, AlgoLevenshtein = iota, 22 implicit-iota follow-ups, numAlgorithms = int(AlgoRatcliffObershelp) + 1 sizes the dispatch array — compile-time correctness."
    - "Switch-based String() for stringly-typed enums: every AlgoID gets an explicit case, default branch handles out-of-range via fmt.Sprintf. 0 allocs/op in-range; intentionally allocating on the error path. //nolint:gocyclo with godoc rationale because gocyclo's 10-branch threshold does not fit a 23-case dispatch."
    - "Fresh-slice-literal accessor for the deterministic catalogue: AlgoIDs() returns a slice literal of all 23 constants — no map iteration, no init-time table, allocation per call so callers can mutate freely."
    - "Test-only re-export of internal scaffolding to keep unused-linter quiet: NumAlgorithmsForTest constant + DispatchLenForTest/DispatchEntryNilForTest function helpers in export_test.go let black-box tests pin the dispatch shape without exposing it to consumers."
    - "Flat sentinel errors via errors.New + table-driven discrimination tests: sentinelCases() in errors_test.go is the single source of truth; wrap-identity, distinct-message, package-prefix, lowercase-no-trailing-punctuation, pairwise-distinct, and not-nil checks all iterate it. Adding a new sentinel adds one row and inherits every check."
    - "Property-test discipline for enum determinism: TestAlgoID_String_StableAcrossCalls uses testing/quick to assert two successive String() calls return the same value for ANY AlgoID value (including out-of-range). The two invocations are stored in named locals so staticcheck SA4000 (identical-expressions-across-operator) stays clean."

key-files:
  created:
    - algoid.go (350 lines — type AlgoID, 23 iota constants with per-constant godoc citing primary source, numAlgorithms, String(), AlgoIDs(), dispatch skeleton)
    - algoid_test.go (10 functions — 9 tests + 1 benchmark; covers count, deterministic-order, distinct, dense-from-zero, non-empty/no-whitespace, out-of-range fallback, stable-across-calls property, round-trip-uniqueness, no-Algo-prefix, dispatch sized + all-nil)
    - errors.go (4 sentinel vars — ErrInvalidInput, ErrInvalidConfiguration, ErrInvalidAlgorithm, ErrEmptyInput)
    - errors_test.go (6 functions — wrap-identity, distinct-messages, package-prefix, lowercase-no-trailing-punctuation, pairwise-distinct, not-nil)
  modified:
    - export_test.go (added NumAlgorithmsForTest constant + DispatchLenForTest + DispatchEntryNilForTest helpers so the dispatch skeleton stays linker-live)

key-decisions:
  - "AlgoID backing uses bare iota (start at 0), not iota+1 as in the spec's illustrative code. CONTEXT.md D-01 plus Plan task 2 TestAlgoIDs_DenseFromZero (`AlgoIDs()[0] == 0`) lock this choice. Per CLAUDE.md Design Principle 13 the spec's code blocks are illustrative; api-ergonomics-reviewer's plan-locked decision wins. Rationale: dense-from-zero gives the dispatch array its natural sizing (numAlgorithms = int(AlgoRatcliffObershelp) + 1) without an off-by-one and matches Go's iota idiom; if a 'none/unset' sentinel ever surfaces it gets added at the END of the const block (after numAlgorithms) so existing AlgoID values never shift."
  - "String() returns CamelCase canonical spelling matching the constant name without the Algo prefix (Levenshtein / DamerauLevenshteinOSA / NYSIIS / MRA / LCSStr) — NOT the snake_case shown in the spec's illustrative code. The plan locks this in the action block ('canonical spellings: Levenshtein, DamerauLevenshteinOSA, ...'). Rationale: CamelCase preserves acronym capitalisation (NYSIIS/MRA/LCSStr/OSA) as a single readable token; matches the established naming convention in .claude/skills/go-coding-standards (acronyms stay uppercase); matches Go's stringer convention. A snake_case adapter can be layered in Phase 8 if Scorer.ScoreAll's map keys want that form — not a contract change."
  - "Exactly four sentinels shipped: ErrInvalidInput, ErrInvalidConfiguration, ErrInvalidAlgorithm, ErrEmptyInput — the plan's named-four set. The spec's illustrative code shows six (ErrEmptyScorer, ErrInvalidWeight, ErrInvalidThreshold, ErrInvalidAlgoID, ErrInvalidQGramSize, ErrInvalidTverskyParam); those are deferred to the phases that need them (Phase 8 Scorer / Phase 6 Monge-Elkan etc). Adding a sentinel in a later phase is a non-breaking change — the table-driven tests in errors_test.go inherit every new row from sentinelCases() with zero rewrite."
  - "//nolint:gocyclo applied at String() with explicit godoc rationale: the switch is intentionally one case per AlgoID so a new constant cannot fall through to the AlgoID(N) fallback unnoticed; algorithm-correctness-reviewer flags any future PR that omits a String() case. The lint threshold of 10 branches doesn't fit the canonical Go stringer pattern for a 23-entry enum."
  - "Test-only re-export of internal scaffolding (numAlgorithms + dispatch) via export_test.go (NumAlgorithmsForTest / DispatchLenForTest / DispatchEntryNilForTest). Rationale: numAlgorithms and dispatch are unused outside this file at the Phase 1 state (algorithms register themselves from Phase 2). Without a reference they trip the unused linter. The re-export pattern is the same one plan 01-04 established for canonicalMarshal → CanonicalMarshalForTest — keeps the public API minimal while letting black-box tests pin internal invariants. Helpers return values rather than copy the array because the dispatch array holds function pointers and a copy would be ~184 bytes on 64-bit platforms."
  - "Property-test self-comparison rewritten to two named locals to satisfy staticcheck SA4000 (identical-expressions-across-operator). `id.String() == id.String()` is the standard testing/quick idiom for stability properties but SA4000 flags it; `first := id.String(); second := id.String(); return first == second` is semantically identical and lint-clean."

patterns-established:
  - "Dense-from-zero typed enum: type AlgoID int + AlgoX AlgoID = iota + 22 implicit follow-ups + numAlgorithms compile-time-derived from the last constant. Future enums in fuzzymatch (if any — for example, a hypothetical TokenisationMode in Phase 7) follow the same shape."
  - "Switch-based stringer with allocating-default fallback: one case per enum constant, fmt.Sprintf only on the default branch, //nolint:gocyclo with godoc-justified rationale. This is the canonical pattern for any future stringly-typed enum in the library."
  - "Sentinel-discrimination test harness: sentinelCases() returns a single slice of {name, err} entries; six independent property tests iterate the same slice. New sentinels added in later phases (ErrInvalidThreshold in Phase 8, etc.) inherit every check by adding one row."
  - "Phase 1 + 2 hand-off contract for algorithm dispatch: numAlgorithms sizes a [N]func(a, b string) float64 array; algorithm files in Phase 2 onwards register their score function via `dispatch[AlgoLevenshtein] = levenshteinScore` at package-load time (NOT in init() — by direct assignment in a var ... = func at package scope, or by being a package-level var with the appropriate type signature). The lookup helper that Scorer (Phase 8) and Extract (Phase 10) will use gates on `if int(id) >= numAlgorithms || dispatch[id] == nil { return ErrInvalidAlgorithm }`."

requirements-completed:
  - FOUND-02
  - FOUND-05

# Metrics
duration: ~20min
completed: 2026-05-13
---

# Phase 1 Plan 5: Primitives — AlgoID + Errors Summary

**First production Go code: typed AlgoID enum (23 constants, iota-backed, switch-stringed, zero-alloc hot path), nil-initialised dispatch skeleton sized at compile time, and the four flat sentinel errors that compose with errors.Is / errors.As across the library's wrapping surface.**

## Performance

- **Duration:** ~20 min
- **Tasks:** 3 of 3
- **Files created:** 4 (algoid.go, algoid_test.go, errors.go, errors_test.go)
- **Files modified:** 1 (export_test.go — re-exported numAlgorithms + dispatch helpers)

## Accomplishments

- **`algoid.go`** — the first ~350 lines of production Go in fuzzymatch. Declares `type AlgoID int` (a plain int per CONTEXT.md D-01, NOT int32/int64/struct), 23 iota-backed exported constants in the spec catalogue order from `docs/requirements.md` §7 with per-constant godoc citing the originating algorithm (Levenshtein 1965 → Ratcliff & Metzener 1988), the unexported `numAlgorithms = int(AlgoRatcliffObershelp) + 1` (compile-time 23) sizing the dispatch array, `String()` via a switch with one explicit case per constant returning canonical CamelCase labels ("Levenshtein", "DamerauLevenshteinOSA", "NYSIIS", "MRA", "LCSStr") plus a `fmt.Sprintf("AlgoID(%d)", int(id))` fallback for out-of-range values, `AlgoIDs()` returning a freshly-allocated slice literal enumerating all 23 constants in declared order (NO map iteration, NO loop), and the all-nil `var dispatch [numAlgorithms]func(a, b string) float64` skeleton ready for Phase 2+ algorithm registration.

- **`algoid_test.go`** — 9 unit/property tests plus one benchmark, stdlib testing only (testing + testing/quick + strings + unicode — no testify in root):
  | Test / Benchmark | What it pins |
  | --- | --- |
  | `TestAlgoIDs_Count_Is23` | Catalogue size locked at 23 |
  | `TestAlgoIDs_DeterministicOrder` | 100 iterations of AlgoIDs() return byte-identical contents (no map-iteration leak — DET-03) |
  | `TestAlgoIDs_Distinct` | Every entry unique (no duplicate AlgoID values) |
  | `TestAlgoIDs_DenseFromZero` | `AlgoIDs()[0] == 0` and `int(AlgoIDs()[i]) == i` for all i (catches `iota+1` regression and accidental `_ AlgoID = iota` gaps) |
  | `TestAlgoID_String_NotEmpty_ForEveryConstant` | Every in-range AlgoID returns a non-empty, whitespace-free label (sub-tests named per label for CI diagnosis) |
  | `TestAlgoID_String_OutOfRange` | `AlgoID(999) / AlgoID(-1) / AlgoID(100)` return the `"AlgoID(N)"` fmt.Sprintf form |
  | `TestAlgoID_String_StableAcrossCalls` | testing/quick: ANY AlgoID value's String() is stable across calls |
  | `TestAlgoID_RoundTrip` | Every AlgoID has a unique String() label (no two AlgoIDs collide on the canonical form) |
  | `TestAlgoID_String_NoAlgoPrefix` | Canonical labels do NOT carry the "Algo" prefix |
  | `TestDispatch_SizedForCatalogue` | numAlgorithms == len(AlgoIDs()) AND dispatch array length == len(AlgoIDs()) |
  | `TestDispatch_AllNilAtPhase1` | Every dispatch entry is nil at the Phase 1 state (Phase 2+ updates this test) |
  | `BenchmarkAlgoID_String` | `b.ReportAllocs()`; in-range hot path measures 0 B/op, 0 allocs/op on Apple M2 |

- **`errors.go`** — Four flat sentinel errors via `errors.New`, each with the `"fuzzymatch: "` prefix, lowercase body, no trailing punctuation:
  - `ErrInvalidInput`         = `"fuzzymatch: invalid input"`
  - `ErrInvalidConfiguration` = `"fuzzymatch: invalid configuration"`
  - `ErrInvalidAlgorithm`     = `"fuzzymatch: invalid algorithm"`
  - `ErrEmptyInput`           = `"fuzzymatch: empty input"`
  Each godoc documents the conditions that surface it, points consumers at `errors.Is` for discrimination, and explicitly forbids string matching.

- **`errors_test.go`** — 6 table-driven tests over a shared `sentinelCases()` slice covering wrap-identity (`fmt.Errorf("scorer: %w", ErrX)` discoverable via `errors.Is`), distinct messages, package-prefix `"fuzzymatch: "` on every sentinel, lowercase body with no trailing `.`/`!`/`?`, pairwise distinctness via `errors.Is` (catches a future refactor accidentally aliasing two sentinels), and not-nil. Stdlib testing only (testing + errors + fmt + strings + unicode).

- **`export_test.go` (modified)** — Re-exports `numAlgorithms` as `NumAlgorithmsForTest`, plus two helpers `DispatchLenForTest()` and `DispatchEntryNilForTest(i)` so black-box tests can pin the dispatch shape without copying the array (which holds function pointers).

## Concrete decisions recorded (per `<output>` block)

- **Final exported constant names (drift from the default proposal in the plan's `<interfaces>` block):** ZERO drift. The 23 constants ship exactly as proposed:
  `AlgoLevenshtein, AlgoDamerauLevenshteinOSA, AlgoDamerauLevenshteinFull, AlgoHamming, AlgoJaro, AlgoJaroWinkler, AlgoStrcmp95, AlgoSmithWatermanGotoh, AlgoLCSStr, AlgoQGramJaccard, AlgoSorensenDice, AlgoCosine, AlgoTversky, AlgoMongeElkan, AlgoTokenSortRatio, AlgoTokenSetRatio, AlgoPartialRatio, AlgoTokenJaccard, AlgoSoundex, AlgoDoubleMetaphone, AlgoNYSIIS, AlgoMRA, AlgoRatcliffObershelp`.
- **Final exported sentinel error names (additions beyond the spec's named four):** ZERO additions. The plan's named-four ship exactly as proposed: `ErrInvalidInput, ErrInvalidConfiguration, ErrInvalidAlgorithm, ErrEmptyInput`. Later-phase additions (`ErrInvalidThreshold`, `ErrInvalidWeight`, `ErrHammingLengthMismatch`, etc.) are deferred to the phases that introduce the features that need them.
- **Coverage percentages observed:** overall 100.0% of statements; per-file 100.0% on algoid.go (both `String` and `AlgoIDs` exercised). errors.go contains only `var` declarations (no functions, no statements counted by `go tool cover`); every sentinel is referenced by the test table so the public-API-funcs floor is satisfied. The verify-coverage-floors script reports `OK: verify-coverage-floors — overall 100.0% >= 95.0%; per-file >= 90.0%; public-API funcs all exercised (6 exported symbols inspected).`
- **BenchmarkAlgoID_String output:** `BenchmarkAlgoID_String-8   418609640   2.674 ns/op   0 B/op   0 allocs/op` on Apple M2 (darwin/arm64, Go 1.26.3). The in-range hot path is zero-allocation. CI matrix platforms will populate `bench.txt` from Phase 2 onwards (PERF-04 informational per D-09).

## Task Commits

Each task committed atomically on `worktree-agent-af2d222815774854b`:

1. **Task 1: Write algoid.go (AlgoID type, 23 constants, String, AlgoIDs, dispatch skeleton)** — `e67be87` (feat)
2. **Task 2: Write algoid_test.go — unit + property tests** — `706df09` (test)
3. **Task 3: Write errors.go and errors_test.go (sentinel set + Is-wrap identity tests)** — `3ac9a64` (feat)

## Plan-Level Verification

| Step | Result |
| ---- | ------ |
| `go build ./...` | PASS |
| `go vet ./...` | PASS |
| `bash scripts/verify-license-headers.sh` | PASS (10 .go files, all carry the Apache-2.0 header) |
| `bash scripts/verify-no-runtime-deps.sh` | PASS (root go.mod allowlist clean: github.com/axonops/fuzzymatch + golang.org/x/text) |
| `golangci-lint run ./...` (incl. tests/bdd) | PASS (0 issues) |
| `go test -race -shuffle=on -count=1 ./...` | PASS (all tests across algoid, errors, golden_canonical, golden_bootstrap) |
| `go test -bench=BenchmarkAlgoID_String -benchmem -count=1 -run=^$ ./...` | PASS (0 B/op, 0 allocs/op in-range) |
| `make coverage && make coverage-check` | PASS (overall 100.0% >= 95.0%; per-file >= 90.0%; public-API funcs all exercised — 6 exported symbols inspected) |
| `make tidy-check` | PASS (no diff to go.mod / go.sum / tests/bdd/go.mod / tests/bdd/go.sum) |
| `make security` | PASS (No vulnerabilities found via govulncheck) |
| `make check` end-to-end | PASS |

## Deviations from Plan

### Auto-fixed Issues

**1. [Rule 1 — Lint] gocyclo flagged String() at complexity 24 (> 10 threshold)**
- **Found during:** Task 1, `golangci-lint run ./...` after first-cut algoid.go.
- **Issue:** golangci-lint's gocyclo linter has a threshold of 10 cyclomatic-complexity branches per function (set in `.golangci.yml` from plan 01-02). The String() switch has 24 branches (23 in-range cases + default) which is intentional — every AlgoID gets an explicit case so a new constant cannot silently fall through to the `AlgoID(N)` fallback, and the algorithm-correctness reviewer will flag any future PR that omits one. The lint threshold simply doesn't fit a 23-entry stringly-typed enum.
- **Fix:** Added a `//nolint:gocyclo` directive at the function signature with a 6-line godoc paragraph above it explaining the rationale ("one switch case per AlgoID is intentional — see godoc above"). The lint suppression is surgical (only that one function) and the rationale is documented inline so a future reviewer understands why it's there.
- **Files modified:** `algoid.go` (added the suppression + the godoc paragraph in the same diff).
- **Verification:** `golangci-lint run ./...` exits 0 after the fix.
- **Committed in:** `e67be87` (Task 1 commit).

**2. [Rule 1 — Lint] unused-linter flagged numAlgorithms and dispatch at Phase 1**
- **Found during:** Task 1, `golangci-lint run ./...` after first-cut algoid.go.
- **Issue:** The plan explicitly defines `numAlgorithms` and `dispatch` as scaffolding for Phase 2+ algorithm registration — they intentionally have no consumers at the Phase 1 state. golangci-lint's `unused` linter (which is part of the `default: standard` set per `.golangci.yml`) flagged both as dead code: `algoid.go:187:7: const numAlgorithms is unused` and `algoid.go:316:5: var dispatch is unused`. Adding fake usages in production code would be wrong; deleting them would invalidate the plan's must-haves and the Phase 2+ contract.
- **Fix:** Extended `export_test.go` (which is `_test.go`-compiled and exists outside the production code path) with three re-exports: `const NumAlgorithmsForTest = numAlgorithms`, `func DispatchLenForTest() int { return len(dispatch) }`, and `func DispatchEntryNilForTest(i int) bool { return i >= 0 && i < len(dispatch) && dispatch[i] == nil }`. This is the same pattern plan 01-04 established for `canonicalMarshal → CanonicalMarshalForTest`. The production symbols stay unexported; the test package can pin the dispatch shape (sized for 23, all-nil at Phase 1) via the re-exports. Two new tests (`TestDispatch_SizedForCatalogue`, `TestDispatch_AllNilAtPhase1`) exercise the re-exports so the production symbols are linker-live.
- **Files modified:** `algoid.go` (no change here); `export_test.go` (added the three re-exports); `algoid_test.go` (added the two tests).
- **Verification:** `golangci-lint run ./...` exits 0; both `TestDispatch_*` tests pass.
- **Committed in:** `e67be87` (re-exports landed with Task 1's algoid.go), `706df09` (the two tests landed with Task 2's algoid_test.go).

**3. [Rule 1 — Lint] staticcheck SA4000 flagged the testing/quick stability check**
- **Found during:** Task 2, `golangci-lint run ./...` after first-cut algoid_test.go.
- **Issue:** First-cut `TestAlgoID_String_StableAcrossCalls` body was `return id.String() == id.String()` — the standard testing/quick idiom for asserting two successive calls return the same value. staticcheck's SA4000 rule ("identical expressions on the left and right side of the '==' operator") flags this as a likely bug because `x == x` typically indicates a typo. Here it's deliberate: we want to invoke the method twice and compare the results, asserting determinism across calls.
- **Fix:** Rewrote the body as two named locals: `first := id.String(); second := id.String(); return first == second`. Semantically identical (same two method invocations; same comparison) but lint-clean. Added a godoc paragraph to the test function explaining why the two locals exist.
- **Files modified:** `algoid_test.go`.
- **Verification:** `golangci-lint run ./...` exits 0; the test still exercises testing/quick over the full int domain.
- **Committed in:** `706df09` (Task 2 commit).

---

**Total deviations:** 3 auto-fixed (all Rule 1 — lint friction with intentional design decisions). No architectural changes. No new sentinel additions beyond the plan's named four. No constant renames or reorderings.

**Impact on plan:** All three fixes are surgical and preserve every must-have. Each is documented inline with the rationale so future reviewers understand why the suppression / refactor was applied. The export_test.go re-export pattern (deviation 2) is the established Phase 1 idiom from plan 01-04 — using it again here keeps the codebase consistent.

## Issues Encountered

- **Spec's illustrative AlgoID code uses `iota + 1`; CONTEXT.md D-01 and the plan's Task 2 `TestAlgoIDs_DenseFromZero` use bare `iota`.** Per CLAUDE.md Design Principle 13 ("Code blocks in the requirements doc are illustrative; the agents have veto authority over the surface shape"), the plan-locked decision wins. Implemented bare `iota`; `TestAlgoIDs_DenseFromZero` passes (`AlgoIDs()[0] == 0`).
- **Spec's illustrative AlgoID String() returns snake_case ("levenshtein", "jaro_winkler"); the plan's Task 1 action block locks CamelCase ("Levenshtein", "JaroWinkler", "NYSIIS", "MRA", "LCSStr").** Plan wins per Design Principle 13. CamelCase preserves acronym capitalisation as a single readable token. A snake_case adapter can be layered in Phase 8 if Scorer.ScoreAll's map keys want that form.
- **Spec's illustrative sentinels enumerate six names (ErrEmptyScorer, ErrInvalidWeight, ErrInvalidThreshold, ErrInvalidAlgoID, ErrInvalidQGramSize, ErrInvalidTverskyParam); the plan and CONTEXT.md D-05 lock four (ErrInvalidInput, ErrInvalidConfiguration, ErrInvalidAlgorithm, ErrEmptyInput).** Plan wins. The plan-locked four ship; the rest land in the phases that introduce the features needing them.
- **`go test -shuffle=on` exposed the test-order independence early.** All 26 tests across the package pass under shuffled order on every run.

## User Setup Required

None — every gate is automated and runs in `make check`.

## Threat Surface Scan

Re-reviewed every file created/modified against the plan's `<threat_model>` register:

- **T-01-05-01 (Tampering — AlgoID reordering breaks v1.x stability):** mitigated. `algoid.go`'s block comment locks the catalogue order in the spec catalogue order; `TestAlgoIDs_DenseFromZero` + `TestAlgoIDs_DeterministicOrder` + `TestAlgoIDs_Distinct` collectively pin the integer-value contract. Algorithm-correctness reviewers will block any future PR that reorders existing constants.
- **T-01-05-02 (Information Disclosure — out-of-range AlgoID panicking in dispatch):** accepted. `dispatch` is unexported (consumers cannot index it directly); Phase 8's Scorer will gate AlgoID lookups via the `if int(id) >= numAlgorithms || dispatch[id] == nil { return ErrInvalidAlgorithm }` pattern documented in this plan's CONTEXT.md / threat model.
- **T-01-05-03 (Spoofing — sentinel error masquerading via string-only check):** mitigated. `errors_test.go` `TestSentinels_WrapIdentity` + `TestSentinels_AreDistinctAsValues` assert `errors.Is` discrimination works after wrapping and that no two sentinels alias. Every sentinel godoc says "Discriminate via `errors.Is(...)`; never match the error message string."
- **T-01-05-04 (Tampering — init()-time table build creating cross-platform divergence):** mitigated. `algoid.go` contains no `init()` function (verified by `grep -c '^func init' algoid.go` returning 0); `String()` is a switch (statically compiled, no map lookup); `AlgoIDs()` is a slice literal (no map iteration). `errors.go` likewise has no `init()`. determinism-reviewer's no-init-time-table rule is satisfied.
- **T-01-05-05 (Tampering — dispatch array indexed by out-of-range AlgoID):** mitigated structurally. `dispatch` is unexported (`var dispatch ...`, lowercase). Phase 8's Scorer is the public consumer; this plan ships the skeleton with all entries nil.

No new threat flags raised. No security-relevant surface introduced beyond what the threat model anticipates.

## Next Plan Readiness

- **Plan 01-06 (Normalise)** inherits `ErrInvalidInput` from `errors.go` (the rune-aware path's documented invalid-UTF-8 disposition can return it wrapped via `fmt.Errorf("normalise: %w", fuzzymatch.ErrInvalidInput)` if `Normalise` later grows an `(string, error)` variant; the default `Normalise(s, opts) string` path returns a value, not an error). The 4 sentinels are the canonical vocabulary that plan 01-06's tests will reference via `errors.Is`.
- **Plan 01-07 (Tokenise)** likewise inherits the sentinel vocabulary.
- **Plan 01-08 (DX docs)** lists every exported symbol from this plan in `llms.txt` / `llms-full.txt`: AlgoID, 23 AlgoX constants, AlgoIDs, ErrInvalidInput, ErrInvalidConfiguration, ErrInvalidAlgorithm, ErrEmptyInput. The `ai_friendly_test.go` meta-test (which parses `go/ast` per `.claude/skills/go-testing-standards`) will verify exhaustiveness from plan 01-08 onwards.
- **Phase 2+ algorithm plans** inherit the dispatch contract: each algorithm's implementation file registers itself by direct package-level assignment, e.g. `var _ = func() bool { dispatch[AlgoLevenshtein] = levenshteinScore; return true }()` or (cleaner) a single internal helper called from the algorithm's file that asserts the slot is nil before writing. The `TestDispatch_AllNilAtPhase1` test will need updating in Phase 2 to assert only the unregistered slots are nil (or be deleted entirely once the catalogue is fully wired in Phase 7).
- **Phase 8 (Scorer)** inherits `ErrInvalidAlgorithm`. The Scorer's internal lookup helper uses `if int(id) >= numAlgorithms || dispatch[id] == nil { return 0, fmt.Errorf("scorer: %w", ErrInvalidAlgorithm) }` per the threat-model mitigation.
- **Phase 10 (Extract)** inherits the same contract via the Scorer.

## Self-Check: PASSED

Files claimed to exist (verified with `[ -f ... ]`):
- `algoid.go` — FOUND
- `algoid_test.go` — FOUND
- `errors.go` — FOUND
- `errors_test.go` — FOUND
- `export_test.go` — FOUND (modified)
- `.planning/phases/01-foundation-infrastructure/01-05-primitives-algoid-errors-SUMMARY.md` — FOUND (this file)

Commits claimed to exist (verified with `git log --oneline`):
- `e67be87` — FOUND (Task 1: feat(01-05): add AlgoID typed enum and dispatch skeleton)
- `706df09` — FOUND (Task 2: test(01-05): pin AlgoID public-API contract and dispatch skeleton)
- `3ac9a64` — FOUND (Task 3: feat(01-05): add sentinel error vocabulary)

Plan-level success criteria (from `<success_criteria>`):
- algoid.go declares AlgoID, 23 iota constants, String(), AlgoIDs(), numAlgorithms, and the nil dispatch array — PASS
- errors.go declares 4 sentinel errors with the fuzzymatch: prefix — PASS
- All unit + property tests pass — PASS
- 100% public API coverage on both files — PASS (algoid.go 100.0% statement coverage; errors.go has no statement-bearing functions; both files' public symbols all exercised, verify-coverage-floors reports 6 exported symbols inspected)
- No init() functions anywhere — PASS (`grep -c '^func init' algoid.go errors.go` returns 0)
- No map iteration on any output path — PASS (AlgoIDs() is a slice literal; String() is a switch; sentinel cases() in tests uses a slice; the maps inside TestSentinels_DistinctMessages / TestAlgoID_RoundTrip / TestAlgoIDs_Distinct are internal-only count checks)
- make check is green — PASS

---
*Phase: 01-foundation-infrastructure*
*Plan: 05 (primitives-algoid-errors)*
*Completed: 2026-05-13*
