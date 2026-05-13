---
plan_id: 01-05
phase: 01-foundation-infrastructure
plan: 05
type: execute
wave: 5
depends_on:
  - 01-04
autonomous: true
objective: >
  Land the first production Go code: `AlgoID` typed enum (23 constants with
  `iota` backing, `String()` method, `AlgoIDs()` accessor, dispatch-array
  skeleton sized for 23 with all entries nil), plus the flat sentinel error
  set (`ErrInvalidInput`, `ErrInvalidConfiguration`, `ErrInvalidAlgorithm`,
  `ErrEmptyInput`) composing with `errors.Is`/`errors.As`. Public-API 100%
  coverage via property tests on `String()` round-trip and `AlgoIDs()`
  invariants; sentinel errors property-tested for wrap-and-Is identity.
files_modified:
  - algoid.go
  - algoid_test.go
  - errors.go
  - errors_test.go
requirements:
  - FOUND-02
  - FOUND-05
must_haves:
  truths:
    - "`AlgoID` is a typed `int` (`type AlgoID int`)"
    - "23 `iota`-based constants covering the algorithms in docs/requirements.md §7"
    - "`String()` method uses a switch — NO `init()`-time table build; NO `[]string` indexed by AlgoID built in init (D-01 + §13.5)"
    - "`AlgoIDs() []AlgoID` returns a deterministic slice — no map iteration on the return path (DET-03)"
    - "Dispatch array `[N]func(a, b string) float64` is package-level `var`, sized for 23 with all entries nil (Phase 2+ populates)"
    - "Sentinel errors are package-level `var`s created via `errors.New` with the `fuzzymatch:` prefix"
    - "Sentinel errors compose with `errors.Is(fmt.Errorf(\"...: %w\", ErrX), ErrX)` — verified by table-driven test"
    - "100% coverage on public API surface (AlgoID, AlgoIDs, String, every sentinel)"
  artifacts:
    - path: algoid.go
      provides: AlgoID type, 23 iota constants, String() switch, AlgoIDs() accessor, dispatch array skeleton
      contains: "type AlgoID int"
    - path: algoid_test.go
      provides: Unit + property tests for AlgoID
    - path: errors.go
      provides: Sentinel error variables
      contains: "var ErrInvalidInput"
    - path: errors_test.go
      provides: errors.Is wrap-identity tests for each sentinel
  key_links:
    - from: algoid.go
      to: docs/requirements.md §7
      via: 23 constants name the algorithms in §7's order
    - from: errors.go
      to: errors.Is/errors.As consumers
      via: sentinels constructed by errors.New
      pattern: "errors.New"
---

<objective>
Land the first production Go code in fuzzymatch: `algoid.go` (AlgoID enum +
dispatch skeleton) and `errors.go` (sentinel set).

Purpose: Establish the typed dispatch baseline that Monge-Elkan (Phase 6),
the Scorer (Phase 8), and Extract (Phase 10) depend on. Establish the
sentinel error vocabulary the rest of the library composes against.

Output:
  - algoid.go + algoid_test.go
  - errors.go + errors_test.go
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
@.claude/skills/go-coding-standards/SKILL.md
@.claude/skills/go-testing-standards/SKILL.md
@.claude/skills/determinism-standards/SKILL.md
@docs/requirements.md
@doc.go
@golden_canonical.go

<interfaces>
Locked decisions from CONTEXT.md:
- D-01: `type AlgoID int` + `iota`-backed unexported-then-exported constants
  + dispatch via fixed-size array (NO `Algorithm` interface — avoids hot-path
  boxing per `.claude/skills/go-coding-standards/SKILL.md`).
- D-05: Flat sentinel errors via `errors.New`; package-level `var ErrInvalidInput
  = errors.New("fuzzymatch: invalid input")` style; wrap via
  `fmt.Errorf("%w", ErrX)`. Composes with `errors.Is`/`errors.As`.

Required algorithm constants (23 — names per `docs/requirements.md` §7 in
order; concrete identifiers per api-ergonomics-reviewer at execution time —
the proposed default names below mirror the spec and the
.claude/agents/algorithm-correctness-reviewer file pattern conventions in
go-coding-standards SKILL.md: acronyms stay uppercase, no stutter):

  Character-based (9):
    AlgoLevenshtein, AlgoDamerauLevenshteinOSA, AlgoDamerauLevenshteinFull,
    AlgoHamming, AlgoJaro, AlgoJaroWinkler, AlgoStrcmp95,
    AlgoSmithWatermanGotoh, AlgoLCSStr

  Q-gram (4):
    AlgoQGramJaccard, AlgoSorensenDice, AlgoCosine, AlgoTversky

  Token-based (5):
    AlgoMongeElkan, AlgoTokenSortRatio, AlgoTokenSetRatio,
    AlgoPartialRatio, AlgoTokenJaccard

  Phonetic (4):
    AlgoSoundex, AlgoDoubleMetaphone, AlgoNYSIIS, AlgoMRA

  Gestalt (1):
    AlgoRatcliffObershelp

Total: 9 + 4 + 5 + 4 + 1 = 23 ✓

The first constant uses `iota` (e.g. `AlgoLevenshtein AlgoID = iota`); the
remaining 22 follow with no explicit value. No "sentinel" zero value — Algo IDs
start at zero with AlgoLevenshtein; consumers checking `if id == 0` would
match the first algorithm not an unset value. If a "none" sentinel is later
needed it gets added at the END of the enum (after the dispatch-table size
constant) so existing AlgoIDs do not shift.

`numAlgorithms` constant: an unexported constant defined as
`AlgoRatcliffObershelp + 1` (i.e. one past the last) — gives the dispatch
array its size: `var dispatch = [numAlgorithms]func(a, b string) float64{}`.

Sentinel errors (per D-05; concrete names per api-ergonomics-reviewer at
execution time — the default set below is the spec's named four plus an
optional fifth that's discretionary):

  var ErrInvalidInput        = errors.New("fuzzymatch: invalid input")
  var ErrInvalidConfiguration = errors.New("fuzzymatch: invalid configuration")
  var ErrInvalidAlgorithm    = errors.New("fuzzymatch: invalid algorithm")
  var ErrEmptyInput          = errors.New("fuzzymatch: empty input")

Additional sentinels (api-ergonomics-reviewer discretion):
  - ErrInvalidThreshold (for Scorer threshold validation — Phase 8 may add)
  - ErrInvalidWeight (for Scorer weight validation — Phase 8 may add)
  - ErrHammingLengthMismatch (for HammingDistance — Phase 2 may add)
  Plan 01-05 ships the spec's four. Later phases extend as the design
  requires.

Hot-path discipline (per go-coding-standards):
- `String()` MUST NOT allocate on the hot path. Use a switch over the
  AlgoID and return string constants. NO `fmt.Sprintf`, NO map lookup.
- For out-of-range AlgoID: return `fmt.Sprintf("AlgoID(%d)", int(id))` —
  this branch is intentionally allocating because it's the error / debug
  path, not a hot path. Acceptable.
</interfaces>
</context>

<tasks>

<task type="auto">
  <name>Task 1: Write algoid.go (AlgoID type, 23 constants, String, AlgoIDs, dispatch skeleton)</name>
  <files>algoid.go</files>
  <read_first>
    - .planning/phases/01-foundation-infrastructure/01-CONTEXT.md (D-01)
    - docs/requirements.md §6, §7 (Algorithm catalogue — 23 entries in canonical order)
    - .claude/skills/go-coding-standards/SKILL.md (Naming — acronyms stay uppercase, no stutter; API Design — typed AlgoID enum not interface; Files — Apache-2.0 header)
    - .claude/skills/determinism-standards/SKILL.md (Init() and Package Loading — NO `init()`-time table builds)
    - doc.go (plan 01-01 — file-header form)
    - golden_canonical.go (plan 01-04 — pattern reference: file structure, godoc style)
  </read_first>
  <action>
    Create `algoid.go` (package fuzzymatch). Structure:

    1. AxonOps Apache-2.0 file header.

    2. File-level block comment:
       ```
       // AlgoID is the typed identifier for the 23 algorithms in the
       // fuzzymatch catalogue. Dispatch is array-backed for zero-allocation
       // hot-path access — see docs/requirements.md §6.
       //
       // The constant order matches the spec catalogue order in §7 and is
       // the v1.x stability contract. Adding new algorithms appends to the
       // end and never reorders existing constants.
       ```

    3. `package fuzzymatch` clause (lives in doc.go for the package godoc;
       algoid.go's package clause comes after the block comment).

    4. `type AlgoID int` with godoc.

    5. The `iota`-backed const block — all 23 constants in spec order. Use
       a single `const (...)` block. First entry: `AlgoLevenshtein AlgoID = iota`.
       The remaining 22 follow with no explicit value (Go applies `iota`
       implicitly). After the 23rd constant, declare an UNEXPORTED constant
       `numAlgorithms = int(AlgoRatcliffObershelp) + 1`.

       Per CLAUDE.md Design Principle 13: the EXACT constant identifiers
       (`AlgoLevenshtein` vs `LevenshteinAlgo` vs something else) are
       api-ergonomics-reviewer's call. The default proposal above (Algo
       prefix + canonical name from spec, acronyms uppercase per
       go-coding-standards) is the starting point.

    6. Per-constant godoc citing the primary academic source briefly. For
       this plan, brief one-line summaries are sufficient (full citations
       live in each algorithm's implementation file from Phase 2 onwards).
       Example:
         ```
         // AlgoLevenshtein identifies the Levenshtein edit-distance similarity
         // (Levenshtein 1965).
         AlgoLevenshtein AlgoID = iota
         ```

    7. `func (id AlgoID) String() string`:
       - Switch over the AlgoID returning a canonical-spelling string for
         each. Canonical spellings: `"Levenshtein"`, `"DamerauLevenshteinOSA"`,
         `"DamerauLevenshteinFull"`, `"Hamming"`, `"Jaro"`, `"JaroWinkler"`,
         `"Strcmp95"`, `"SmithWatermanGotoh"`, `"LCSStr"`, `"QGramJaccard"`,
         `"SorensenDice"`, `"Cosine"`, `"Tversky"`, `"MongeElkan"`,
         `"TokenSortRatio"`, `"TokenSetRatio"`, `"PartialRatio"`,
         `"TokenJaccard"`, `"Soundex"`, `"DoubleMetaphone"`, `"NYSIIS"`,
         `"MRA"`, `"RatcliffObershelp"`.
         (These canonical spellings are the strings emitted by `String()` — they
         match the constant names without the `Algo` prefix.)
       - Default branch (out-of-range AlgoID): return
         `fmt.Sprintf("AlgoID(%d)", int(id))`. Document that this branch is
         intentionally allocating.

    8. `func AlgoIDs() []AlgoID`:
       - Returns `[]AlgoID{AlgoLevenshtein, AlgoDamerauLevenshteinOSA, ...,
         AlgoRatcliffObershelp}` — a slice literal explicitly enumerating all
         23 constants in their declared order. NO map iteration. NO loop.
       - The slice is a NEW allocation per call so consumers can mutate
         freely; godoc states this contract.

    9. Dispatch array skeleton:
       ```
       // dispatch maps each AlgoID to its score function. Entries are nil
       // until Phase 2+ plans populate them. Consumers MUST NOT call into
       // dispatch directly — use the public algorithm functions
       // (LevenshteinScore etc.) or *Scorer.
       var dispatch [numAlgorithms]func(a, b string) float64
       ```
       The dispatch array is UNEXPORTED (lowercase `dispatch`); the Scorer
       in Phase 8 accesses it via package-internal mechanism.

    10. NO `init()` function. NO map literals at package scope used for
        output. NO `golang.org/x/text` import (this file uses stdlib only).

    Concrete identifiers (default proposal — subject to api-ergonomics-reviewer):
      - Type: `AlgoID`
      - Underlying: `int`
      - 23 constants: AlgoLevenshtein, AlgoDamerauLevenshteinOSA,
        AlgoDamerauLevenshteinFull, AlgoHamming, AlgoJaro, AlgoJaroWinkler,
        AlgoStrcmp95, AlgoSmithWatermanGotoh, AlgoLCSStr, AlgoQGramJaccard,
        AlgoSorensenDice, AlgoCosine, AlgoTversky, AlgoMongeElkan,
        AlgoTokenSortRatio, AlgoTokenSetRatio, AlgoPartialRatio,
        AlgoTokenJaccard, AlgoSoundex, AlgoDoubleMetaphone, AlgoNYSIIS,
        AlgoMRA, AlgoRatcliffObershelp
      - Helper accessor: `AlgoIDs() []AlgoID`
      - String method: `(AlgoID).String() string`
      - Internal: `numAlgorithms` const, `dispatch [numAlgorithms]func(a,b string) float64`
  </action>
  <verify>
    <automated>go build ./... &amp;&amp; go vet ./... &amp;&amp; bash scripts/verify-license-headers.sh &amp;&amp; ! grep -q '^func init' algoid.go &amp;&amp; grep -q 'type AlgoID int' algoid.go &amp;&amp; grep -q 'numAlgorithms\s*=\s*int' algoid.go &amp;&amp; test "$(grep -cE '^\s*Algo[A-Z]' algoid.go)" -ge 23</automated>
  </verify>
  <acceptance_criteria>
    - `algoid.go` exists with the AxonOps Apache-2.0 header
    - Declares `type AlgoID int` (NOT `int32`, `int64`, or a struct)
    - Declares exactly 23 exported `AlgoID` constants in the spec order from §7
    - Declares the unexported `numAlgorithms` constant set to `int(AlgoRatcliffObershelp) + 1`
    - `numAlgorithms` evaluates to 23 (compile-time)
    - `(AlgoID).String()` is implemented via a switch (NO map, NO init-time table)
    - `AlgoIDs() []AlgoID` returns a slice literal of all 23 constants in declared order
    - No `init()` function in the file
    - No map iteration in `String()` or `AlgoIDs()`
    - `var dispatch [numAlgorithms]func(a, b string) float64` exists at package scope
    - All entries in `dispatch` are nil (unset)
    - `go build ./...` exits 0
    - `go vet ./...` exits 0
    - `make verify-license-headers` exits 0
    - `make lint` exits 0 (golangci-lint clean)
    - `gofmt -s -d algoid.go` produces no output (canonical formatting)
  </acceptance_criteria>
  <done>
    `algoid.go` is committed with the 23-constant enum, `String()`,
    `AlgoIDs()`, the unexported `numAlgorithms`, and the nil dispatch array.
    No init-time work. The file builds and vets cleanly.
  </done>
</task>

<task type="auto">
  <name>Task 2: Write algoid_test.go — unit + property tests</name>
  <files>algoid_test.go</files>
  <read_first>
    - algoid.go (just created)
    - .claude/skills/go-testing-standards/SKILL.md (Stdlib testing only in root; property tests via testing/quick; 100% public API coverage)
    - .claude/skills/determinism-standards/SKILL.md (PropAlgorithm_DeterministicAcrossRuns pattern)
    - .planning/phases/01-foundation-infrastructure/01-CONTEXT.md (TEST-06 property-test harness convention)
  </read_first>
  <action>
    Create `algoid_test.go` (package `fuzzymatch_test`) with stdlib `testing`
    only — no testify in root. Tests:

    1. AxonOps Apache-2.0 file header.

    2. `TestAlgoIDs_Count_Is23` — assert `len(AlgoIDs()) == 23`.

    3. `TestAlgoIDs_DeterministicOrder` — call `AlgoIDs()` 100 times; assert
       each call returns a slice with identical contents in identical order
       (proves no map-iteration leak per DET-03). Use a single deep-equal
       check between consecutive calls.

    4. `TestAlgoIDs_Distinct` — assert every element of `AlgoIDs()` is unique
       (no duplicates).

    5. `TestAlgoIDs_DenseFromZero` — assert `AlgoIDs()[0] == 0` and the
       sequence is dense (`AlgoIDs()[i] == AlgoID(i)` for all i). This
       proves the iota backing produces a contiguous block.

    6. `TestAlgoID_String_NotEmpty_ForEveryConstant` — for every AlgoID
       returned by `AlgoIDs()`, assert `String()` returns a non-empty string
       containing no whitespace.

    7. `TestAlgoID_String_OutOfRange` — assert `AlgoID(999).String()` returns
       `"AlgoID(999)"` (the fallback format).

    8. `TestAlgoID_String_StableAcrossCalls` — property test via
       `testing/quick`:
         `f := func(id AlgoID) bool { return id.String() == id.String() }`
         `quick.Check(f, nil)`.
       This proves String() is deterministic for arbitrary input.

    9. `TestAlgoID_RoundTrip` — property test: for every AlgoID returned by
       `AlgoIDs()`, the canonical String() form is unique (no two AlgoIDs
       produce the same String()). Build a map[string]AlgoID and assert
       length 23. (Map IS used here, but only INTERNALLY in the test; the
       assertion is on the count, not on iteration order. Test code is
       exempt from the no-map-iteration-on-output rule — verified internally.)

    10. `BenchmarkAlgoID_String` — minimal benchmark exercising String() to
        ensure it doesn't allocate on the hot path. `b.ReportAllocs()`.
        Document the expected target (0 allocs/op for in-range AlgoIDs).
        This benchmark is empty-input-tolerant — even with zero baseline,
        it exercises the path so future regressions are visible in
        bench.txt.

    Concrete identifiers:
      - Test file `algoid_test.go`
      - Package `fuzzymatch_test`
      - 9 test functions + 1 benchmark
      - Property tests use `testing/quick` (stdlib)
  </action>
  <verify>
    <automated>go test -race -shuffle=on -count=1 -run 'TestAlgoID' ./... &amp;&amp; go test -race -shuffle=on -count=1 -run 'TestAlgoIDs' ./... &amp;&amp; go test -bench=BenchmarkAlgoID_String -benchmem -count=1 -run=^$ ./...</automated>
  </verify>
  <acceptance_criteria>
    - `algoid_test.go` exists with the AxonOps Apache-2.0 header
    - Declares `package fuzzymatch_test`
    - Imports `testing`, `testing/quick` (no testify)
    - 9 test functions present: TestAlgoIDs_Count_Is23, TestAlgoIDs_DeterministicOrder, TestAlgoIDs_Distinct, TestAlgoIDs_DenseFromZero, TestAlgoID_String_NotEmpty_ForEveryConstant, TestAlgoID_String_OutOfRange, TestAlgoID_String_StableAcrossCalls, TestAlgoID_RoundTrip, plus one benchmark BenchmarkAlgoID_String
    - `go test -race -shuffle=on -count=1 ./...` exits 0
    - `go test -bench=BenchmarkAlgoID_String -benchmem -count=1 -run=^$ ./...` exits 0
    - The benchmark output shows 0 allocs/op for the in-range String() case (verified manually by the executor inspecting the output and recording in SUMMARY)
    - Coverage on algoid.go is ≥ 90% per file (verify via `go test -coverprofile=coverage.out ./... && go tool cover -func=coverage.out | grep algoid.go`); since the file is small and every exported symbol is exercised, 100% is expected
    - `make coverage` + `make coverage-check` exits 0 (the new code does not degrade the floor)
  </acceptance_criteria>
  <done>
    algoid.go is fully tested with stdlib testing only; 100% public API
    coverage; determinism property-tested; benchmark proves zero-alloc
    String() on the hot path.
  </done>
</task>

<task type="auto">
  <name>Task 3: Write errors.go and errors_test.go (sentinel set + Is-wrap identity tests)</name>
  <files>errors.go, errors_test.go</files>
  <read_first>
    - .planning/phases/01-foundation-infrastructure/01-CONTEXT.md (D-05 — flat sentinel errors via errors.New)
    - docs/requirements.md §6.4 (sentinel error hierarchy)
    - .claude/skills/go-coding-standards/SKILL.md (Errors section — `fuzzymatch:` prefix, lowercase, no punctuation; errors.Is/errors.As only — never string matching)
    - .claude/skills/go-testing-standards/SKILL.md (Stdlib testing only)
    - doc.go (file-header form)
    - algoid.go (pattern reference)
  </read_first>
  <action>
    Create `errors.go` and `errors_test.go`.

    1. `errors.go` (`package fuzzymatch`):
       - AxonOps Apache-2.0 file header.
       - File-level block comment:
         ```
         // Package-level sentinel errors for fuzzymatch. All errors are
         // wrappable via fmt.Errorf("...: %w", ErrX) and discoverable by
         // errors.Is / errors.As. See docs/requirements.md §6.4.
         ```
       - Imports: only `errors` from stdlib.
       - Four sentinels (D-05 — exact names per api-ergonomics-reviewer at
         execution time; defaults below):
         ```
         // ErrInvalidInput indicates a caller-provided string fails the
         // algorithm's documented input constraints (e.g. invalid UTF-8 on
         // a rune-aware API). Most algorithms accept arbitrary bytes and
         // do not raise this — exceptions document their constraints.
         var ErrInvalidInput = errors.New("fuzzymatch: invalid input")

         // ErrInvalidConfiguration indicates a Scorer or Extract option set
         // is internally inconsistent (e.g. negative weight, threshold
         // outside [0.0, 1.0]). See docs/requirements.md §8.
         var ErrInvalidConfiguration = errors.New("fuzzymatch: invalid configuration")

         // ErrInvalidAlgorithm indicates an AlgoID parameter does not match
         // any registered algorithm in the dispatch table. See AlgoIDs() for
         // the valid set.
         var ErrInvalidAlgorithm = errors.New("fuzzymatch: invalid algorithm")

         // ErrEmptyInput indicates both input strings are empty at the
         // boundary of an API that does not have a defined empty-empty
         // behaviour. Algorithm score functions handle empty inputs per
         // their per-algorithm spec and do NOT return this.
         var ErrEmptyInput = errors.New("fuzzymatch: empty input")
         ```
       - NO `init()` function. NO additional types. NO error types beyond
         the four sentinels. (Phase 8/9 may add typed structs if richer
         per-item context is needed.)

    2. `errors_test.go` (`package fuzzymatch_test`):
       - AxonOps Apache-2.0 file header.
       - Imports: `errors`, `fmt`, `testing`.
       - Table-driven test `TestSentinels_WrapIdentity`:
         ```
         cases := []struct{
           name string
           err  error
         }{
           {"ErrInvalidInput", fuzzymatch.ErrInvalidInput},
           {"ErrInvalidConfiguration", fuzzymatch.ErrInvalidConfiguration},
           {"ErrInvalidAlgorithm", fuzzymatch.ErrInvalidAlgorithm},
           {"ErrEmptyInput", fuzzymatch.ErrEmptyInput},
         }
         for _, c := range cases {
           t.Run(c.name, func(t *testing.T) {
             wrapped := fmt.Errorf("scorer: %w", c.err)
             if !errors.Is(wrapped, c.err) {
               t.Errorf("errors.Is(%q, %v) = false; want true", wrapped, c.err)
             }
           })
         }
         ```
       - `TestSentinels_DistinctMessages`: assert every sentinel has a
         distinct `Error()` string (no two sentinels share text).
       - `TestSentinels_StartWithPackagePrefix`: assert every sentinel's
         `Error()` begins with `"fuzzymatch: "`.
       - `TestSentinels_LowercaseAndNoTrailingPunctuation`: assert every
         sentinel's `Error()` text (after the `"fuzzymatch: "` prefix) is
         lowercase and does NOT end with `'.'`, `'!'`, or `'?'` — per
         go-coding-standards "Error strings: lowercase, no punctuation".
       - `TestSentinels_AreDistinctAsValues`: assert each pair of sentinels
         is distinguishable by `errors.Is`:
         `errors.Is(fuzzymatch.ErrInvalidInput, fuzzymatch.ErrInvalidConfiguration)` is false.

    Concrete identifiers:
      - File `errors.go`, package `fuzzymatch`
      - 4 sentinels: ErrInvalidInput, ErrInvalidConfiguration, ErrInvalidAlgorithm, ErrEmptyInput
      - File `errors_test.go`, package `fuzzymatch_test`
      - 4 test functions

    Note: api-ergonomics-reviewer may, at execution time, propose additional
    sentinels (`ErrInvalidThreshold`, `ErrInvalidWeight`, `ErrHammingLengthMismatch`)
    if it judges they're needed even before Phase 2. Plan 01-05's contract is
    the spec's named four; additions are non-blocking.
  </action>
  <verify>
    <automated>go test -race -shuffle=on -count=1 -run 'TestSentinels_' ./... &amp;&amp; make verify-license-headers &amp;&amp; make lint &amp;&amp; make check</automated>
  </verify>
  <acceptance_criteria>
    - `errors.go` exists with the AxonOps Apache-2.0 header
    - Declares `package fuzzymatch`
    - Imports only `errors` from stdlib (no other imports)
    - Declares 4 (or more, per api-ergonomics-reviewer) package-level `var Err* = errors.New(...)` sentinels
    - Each sentinel message starts with `"fuzzymatch: "`
    - Each sentinel message is lowercase after the prefix
    - Each sentinel message has no trailing punctuation
    - `errors_test.go` exists with the AxonOps Apache-2.0 header
    - 4 test functions covering wrap-identity, distinct messages, package prefix, lowercase-no-trailing-punct, and pairwise distinctness
    - `go test -race -shuffle=on -count=1 -run TestSentinels_ ./...` exits 0
    - `make verify-license-headers` exits 0
    - `make check` exits 0 on the now-complete plan 01-05 state
    - Coverage on errors.go is 100% (every sentinel is referenced by at least one test)
  </acceptance_criteria>
  <done>
    Sentinel errors are declared with the canonical prefix, lowercase
    messages, no trailing punctuation; `errors.Is` wrap-identity is
    table-tested; every sentinel passes the property tests.
  </done>
</task>

</tasks>

<threat_model>
## Trust Boundaries

| Boundary | Description |
|----------|-------------|
| Consumer code → public AlgoID enum | Consumers may pass arbitrary AlgoID values (including out-of-range constructed via `AlgoID(999)`); the dispatch array bounds-check and the sentinel `ErrInvalidAlgorithm` are the gates (validation lives in Phase 8's Scorer, not here). |
| Consumer code → sentinel errors | Consumers should test via `errors.Is`, never via string matching; the sentinel-prefix convention (`fuzzymatch:`) preserves error namespacing across wrapping. |

## STRIDE Threat Register

| Threat ID | Category | Component | Disposition | Mitigation Plan |
|-----------|----------|-----------|-------------|-----------------|
| T-01-05-01 | Tampering | AlgoID enum reordering breaks v1.x stability | mitigate | The constant block is the v1.x stability contract; future additions append to the end. Plan documents this in the file's block comment. Algorithm correctness reviews block reordering PRs. |
| T-01-05-02 | Information Disclosure | Out-of-range AlgoID panicking in dispatch | accept | Dispatch array is unexported; consumers cannot index it directly. Phase 8's Scorer.Score gates AlgoID lookups via `if int(id) >= numAlgorithms || dispatch[id] == nil` returning ErrInvalidAlgorithm. Plan 01-05 establishes the contract; plan 8 enforces it. |
| T-01-05-03 | Spoofing | Sentinel error masquerading via string-only check | mitigate | Tests assert errors.Is identity; documentation-standards mandates errors.Is/errors.As (never string matching). Reviewers enforce. |
| T-01-05-04 | Tampering | init()-time table build creating cross-platform divergence | mitigate | algoid.go has no init(); String() is a switch; AlgoIDs() is a slice literal; verified by the file's structure and by determinism-reviewer at execution time. |
| T-01-05-05 | Tampering | Dispatch array indexed by out-of-range AlgoID | mitigate | Plan 01-05 establishes the array as nil-initialised; plan 01-08's docs document the contract. Phase 8's Scorer enforces bounds-checking; this plan ships only the skeleton — no consumers yet. |
</threat_model>

<verification>
1. `go build ./...` and `go vet ./...` exit 0.
2. `make verify-license-headers` exits 0.
3. `make lint` exits 0.
4. `go test -race -shuffle=on -count=1 ./...` exits 0 (algoid + errors tests pass).
5. `make verify-deps-allowlist` exits 0 (no new runtime deps introduced).
6. `make coverage && make coverage-check` exits 0 (per-file ≥ 90%, public-API 100%).
7. `make check` exits 0.
</verification>

<success_criteria>
- algoid.go declares `AlgoID`, 23 iota constants, `String()`, `AlgoIDs()`,
  `numAlgorithms`, and the nil dispatch array.
- errors.go declares 4 sentinel errors with the `fuzzymatch:` prefix.
- All unit + property tests pass.
- 100% public API coverage on both files.
- No `init()` functions anywhere.
- No map iteration on any output path.
- `make check` is green.
</success_criteria>

<output>
After completion, create
`.planning/phases/01-foundation-infrastructure/01-05-primitives-algoid-errors-SUMMARY.md`
recording:
  - The final exported constant names (per api-ergonomics-reviewer at
    execution time — record any drift from the default proposal).
  - The final exported sentinel error names (including any additions
    beyond the spec's named four).
  - Coverage percentages observed on algoid.go and errors.go.
  - Benchmark output for BenchmarkAlgoID_String (allocations per op).
</output>
