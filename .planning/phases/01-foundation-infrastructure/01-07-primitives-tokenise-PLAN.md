---
plan_id: 01-07
phase: 01-foundation-infrastructure
plan: 07
type: execute
wave: 7
depends_on:
  - 01-06
autonomous: true
objective: >
  Land `Tokenise(s string, opts TokeniseOptions) []string` that splits
  camelCase / snake_case / PascalCase / kebab-case / dot-case with stable
  ordering. Property tests verify order stability under permutation of
  equal-rank tokens, panic-free on arbitrary input, and reconstructibility
  under documented option combinations. Tokenise is the second consumer of
  the no-map-iteration discipline and the second primitive Phase 6 token
  algorithms (Phase 6) will compose against.
files_modified:
  - tokenise.go
  - tokenise_test.go
  - tokenise_bench_test.go
  - tokenise_fuzz_test.go
  - testdata/fuzz/FuzzTokenise/seed-001
requirements:
  - FOUND-04
must_haves:
  truths:
    - "`Tokenise` returns `[]string` (D-04)"
    - "Tokenise splits on camelCase, snake_case, PascalCase, kebab-case, dot-case"
    - "Token order is stable across runs (no map iteration leak)"
    - "Empty / pure-whitespace input returns `[]string{}` (NOT nil — documented contract)"
    - "Tokenise contains NO `init()`-time table builds"
    - "Tokenise contains NO map iteration on output paths"
    - "FuzzTokenise asserts panic-free on arbitrary input including invalid UTF-8"
    - "100% public-API coverage on tokenise.go"
  artifacts:
    - path: tokenise.go
      provides: TokeniseOptions struct, DefaultTokeniseOptions(), Tokenise function
      contains: "func Tokenise"
    - path: tokenise_test.go
      provides: Unit + property tests
    - path: tokenise_bench_test.go
      provides: ASCII-short / ASCII-long / Unicode benchmarks
    - path: tokenise_fuzz_test.go
      provides: FuzzTokenise — panic-free property
    - path: testdata/fuzz/FuzzTokenise/seed-001
      provides: Hand-crafted seed corpus
  key_links:
    - from: tokenise.go
      to: Phase 6 token algorithms (future)
      via: tokens are consumed by MongeElkan, TokenSortRatio, TokenSetRatio, PartialRatio, TokenJaccard
      pattern: "Tokenise"
    - from: tokenise_test.go
      to: testing/quick
      via: property tests for order stability
      pattern: "testing/quick"
---

<objective>
Land the Tokenise primitive — the second consumer-facing API and the
foundation that Phase 6's five token-based algorithms compose against.

Purpose: Establish the splitter contract and the order-stability discipline
NOW so Phase 6 algorithms can be reasoned about without re-deciding token
shape.

Output:
  - tokenise.go (production code)
  - tokenise_test.go (unit + property tests)
  - tokenise_bench_test.go (benchmarks)
  - tokenise_fuzz_test.go (panic-free fuzz)
  - testdata/fuzz/FuzzTokenise/seed-001 (hand-crafted seed)
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
@.claude/skills/performance-standards/SKILL.md
@.claude/skills/algorithm-correctness-standards/SKILL.md
@docs/requirements.md
@normalise.go
@normalise_test.go
@algoid.go
@errors.go

<interfaces>
Locked decisions:
- D-04: `Tokenise(s string, opts TokeniseOptions) []string` — slice return,
  predictable allocation, easy to pass to set/map operations the downstream
  token algorithms need.
- API ergonomics (per CLAUDE.md Design Principle 13): exact field names of
  TokeniseOptions are api-ergonomics-reviewer's call. Default proposal:

  type TokeniseOptions struct {
      // Lowercase: apply ASCII lowercase to each token before returning.
      // For consumers that want Unicode-aware lowercasing, wrap Normalise
      // first or use the lowercasing inside Normalise (per docs/requirements.md §10).
      Lowercase bool

      // SplitCamelCase: insert split boundaries at lowercase->uppercase
      // transitions ("FooBar" → ["foo", "bar"]).
      SplitCamelCase bool

      // SplitConsecutiveUpper: split at the LAST uppercase of a run when
      // followed by lowercase ("XMLHTTPRequest" → ["xml", "http", "request"]).
      // Without this, the same input yields ["xmlhttp", "request"].
      SplitConsecutiveUpper bool

      // SeparatorChars: characters treated as token separators (split AND
      // discarded). Default "_-.:/ \t\n\r".
      SeparatorChars string
  }

  func DefaultTokeniseOptions() TokeniseOptions {
      return TokeniseOptions{
          Lowercase:             true,
          SplitCamelCase:        true,
          SplitConsecutiveUpper: true,
          SeparatorChars:        "_-.:/ \t\n\r",
      }
  }

Empty / whitespace-only input contract: returns `[]string{}` (non-nil empty
slice). Document this loud in the godoc and assert in tests.

Order discipline:
- Tokenise walks the input left-to-right and emits tokens in their input
  order. The function's contract: token[i] occurs in the input before
  token[i+1] in input-byte order. Verified by property test.
- Internal data structures: no maps for splitting (an explicit byte-table
  lookup is fine — a `[128]bool` ASCII-separator-set). Unicode separators
  checked via `strings.ContainsRune(opts.SeparatorChars, r)`.

Edge cases (per algorithm-correctness-standards Unicode Handling):
- Empty string → []string{} (NOT nil)
- Whitespace-only → []string{}
- Single character "a" → ["a"]
- Already-tokenised "foo bar baz" → ["foo", "bar", "baz"]
- camelCase "FooBar" → ["foo", "bar"] (when SplitCamelCase=true)
- snake_case "foo_bar_baz" → ["foo", "bar", "baz"]
- PascalCase "FooBarBaz" → ["foo", "bar", "baz"]
- kebab-case "foo-bar-baz" → ["foo", "bar", "baz"]
- dot.case "foo.bar.baz" → ["foo", "bar", "baz"]
- Consecutive uppercase: "XMLHTTPRequest" → ["xml", "http", "request"]
  (when SplitConsecutiveUpper=true) OR ["xmlhttp", "request"] (when off —
  document the default)
- Mixed scripts: "userПривет" → ["user", "Привет"] (no case split on
  Cyrillic; user remains an ASCII token, Привет remains a Cyrillic token)
- Empty-after-split: "__foo__" → ["foo"] (consecutive separators don't
  produce empty tokens)
- Numeric runs: "Foo123Bar" → ["foo123", "bar"]? OR ["foo", "123", "bar"]?
  Default proposal: keep digits attached to the preceding alpha run; only
  the camelCase/snake_case/etc boundaries split. Document the choice
  explicitly in the godoc.

Performance target (per performance-standards):
- ASCII input ≤ 50 chars: < 500 ns, ≤ 2 allocations (one for the result
  slice, one for the underlying string-builder backing).
- Stack-allocated boundary-index buffer for inputs ≤ 64 bytes:
  `var boundaries [65]int`.
</interfaces>
</context>

<tasks>

<task type="auto">
  <name>Task 1: Write tokenise.go (TokeniseOptions, DefaultTokeniseOptions, Tokenise)</name>
  <files>tokenise.go</files>
  <read_first>
    - .planning/phases/01-foundation-infrastructure/01-CONTEXT.md (D-04)
    - docs/requirements.md §10 (Tokenise spec)
    - .claude/skills/determinism-standards/SKILL.md (Init() rule; No-Map-Iteration)
    - .claude/skills/performance-standards/SKILL.md (Normalisation Budgets — Tokenise targets)
    - .claude/skills/algorithm-correctness-standards/SKILL.md (Unicode Handling)
    - .claude/skills/go-coding-standards/SKILL.md (Files header, API Design)
    - normalise.go (plan 01-06 — reference pattern: ASCII fast path, no init, x/text usage if needed)
    - algoid.go (plan 01-05 — file structure reference)
  </read_first>
  <action>
    Create `tokenise.go` (`package fuzzymatch`).

    1. AxonOps Apache-2.0 file header.

    2. File-level block comment citing the spec:
       ```
       // Tokenise splits s into tokens by camelCase, snake_case, PascalCase,
       // kebab-case, and dot-case boundaries. See docs/requirements.md §10
       // for the authoritative spec.
       //
       // Token order matches input order (left-to-right). Empty / whitespace-
       // only input returns a non-nil empty slice. NO init()-time tables;
       // NO map iteration on output paths.
       ```

    3. Imports: `unicode`, `unicode/utf8`, `strings`. NO x/text usage —
       Tokenise operates on the (assumed-already-normalised) input.

    4. Public types and helpers:
       ```
       type TokeniseOptions struct {
           Lowercase             bool
           SplitCamelCase        bool
           SplitConsecutiveUpper bool
           SeparatorChars        string
       }

       func DefaultTokeniseOptions() TokeniseOptions {
           return TokeniseOptions{
               Lowercase:             true,
               SplitCamelCase:        true,
               SplitConsecutiveUpper: true,
               SeparatorChars:        "_-.:/ \t\n\r",
           }
       }

       // Tokenise returns the tokens of s per the configured options. Empty
       // or whitespace-only input returns a non-nil empty slice. The token
       // order matches the byte order of s.
       func Tokenise(s string, opts TokeniseOptions) []string {
           if s == "" {
               return []string{}  // explicit non-nil empty
           }
           // Build a [128]bool separator set for ASCII fast lookup.
           // Non-ASCII separator chars (rare) are checked via strings.ContainsRune.
           var sepASCII [128]bool
           for i := 0; i < len(opts.SeparatorChars); i++ {
               if c := opts.SeparatorChars[i]; c < 128 {
                   sepASCII[c] = true
               }
           }
           // Walk the input; emit tokens at boundaries.
           tokens := make([]string, 0, 4) // typical short identifier has 1-4 tokens
           start := 0
           runes := []rune(s) // single allocation; needed for camelCase logic on multi-byte
           // ... (loop logic as detailed below) ...
           return tokens
       }
       ```

       Loop logic (per <interfaces>):
       For each rune in runes:
         a. If rune is a separator (ASCII fast: sepASCII[byte] if rune < 128;
            else strings.ContainsRune(opts.SeparatorChars, r)):
            - Emit token from `runes[start:i]` if non-empty (after applying
              Lowercase if opts.Lowercase).
            - Set `start = i+1` (skip the separator).
         b. Else if opts.SplitCamelCase:
            - If runes[i] is uppercase AND runes[i-1] is lowercase:
              Emit `runes[start:i]`, set `start = i`.
            - If opts.SplitConsecutiveUpper AND
              runes[i] is uppercase AND runes[i+1] is lowercase AND
              i > start AND runes[i-1] is uppercase:
              Emit `runes[start:i]`, set `start = i`.
              (This catches "XMLHttp" → split at the X→H boundary's
              successor: "XML" + "Http". For "XMLHTTPRequest", the last X
              before the lowercase R is the boundary: "XMLHTTP" + "Request".
              Adjust the canonical break point if api-ergonomics-reviewer
              picks a different convention.)
       After the loop: emit the final token from `runes[start:]` if non-empty.

       Apply Lowercase per token if opts.Lowercase. For ASCII tokens use
       byte-level bitwise lowercase; for non-ASCII tokens use
       `strings.ToLower` (preserves Unicode case-folding semantics).

    5. NO `init()` function. NO map literals at package scope used for output.

    6. NO map iteration on any return path.

    Concrete identifiers (default proposal — api-ergonomics-reviewer at
    execution time):
      - File `tokenise.go`, package `fuzzymatch`
      - Type `TokeniseOptions` with 4 fields
      - `DefaultTokeniseOptions() TokeniseOptions`
      - `Tokenise(s string, opts TokeniseOptions) []string`
      - Internal helpers: `[128]bool` sepASCII table built per-call

    Note (Design Principle 13): the exact field names, the
    SplitConsecutiveUpper semantics (whether "XMLHTTPRequest" produces
    ["xmlhttp", "request"] or ["xml", "http", "request"]) are
    api-ergonomics-reviewer's call. The default above is the proposed
    starting point. Whichever choice lands, document it loud in the godoc
    and pin in property tests.
  </action>
  <verify>
    <automated>go build ./... &amp;&amp; go vet ./... &amp;&amp; make verify-license-headers &amp;&amp; make lint &amp;&amp; ! grep -E '^func init\(\)' tokenise.go &amp;&amp; ! grep -E 'math\.(Pow|Log|Exp|FMA)' tokenise.go &amp;&amp; grep -q 'func Tokenise' tokenise.go &amp;&amp; grep -q 'TokeniseOptions' tokenise.go</automated>
  </verify>
  <acceptance_criteria>
    - `tokenise.go` exists with the AxonOps Apache-2.0 header
    - Declares `type TokeniseOptions struct { ... }` with 4 fields
    - Declares `DefaultTokeniseOptions() TokeniseOptions` returning the proposed defaults
    - Declares `Tokenise(s string, opts TokeniseOptions) []string`
    - `Tokenise("", DefaultTokeniseOptions())` returns a non-nil empty slice (`len(result) == 0 && result != nil`)
    - `Tokenise("FooBar", DefaultTokeniseOptions())` returns `[]string{"foo", "bar"}`
    - `Tokenise("foo_bar_baz", DefaultTokeniseOptions())` returns `[]string{"foo", "bar", "baz"}`
    - `Tokenise("foo.bar.baz", DefaultTokeniseOptions())` returns `[]string{"foo", "bar", "baz"}`
    - `Tokenise("foo-bar-baz", DefaultTokeniseOptions())` returns `[]string{"foo", "bar", "baz"}`
    - Contains NO `func init()`
    - Contains NO `math.Pow`, `math.Log`, `math.Exp`, `math.FMA`
    - `go build ./...` exits 0
    - `make verify-license-headers` exits 0
    - `make lint` exits 0
    - `make verify-deps-allowlist` exits 0 (no new runtime deps)
  </acceptance_criteria>
  <done>
    Tokenise compiles, vets, lints cleanly; respects D-04 (slice return);
    no init-time work; emits stable order; the empty/whitespace contract is
    explicit.
  </done>
</task>

<task type="auto">
  <name>Task 2: Write tokenise_test.go (unit + property tests)</name>
  <files>tokenise_test.go</files>
  <read_first>
    - tokenise.go (just created)
    - .claude/skills/go-testing-standards/SKILL.md (Stdlib testing only; property tests via testing/quick)
    - .claude/skills/determinism-standards/SKILL.md (order stability; no map iteration)
    - normalise_test.go (plan 01-06 — pattern reference: structure, property test idioms)
  </read_first>
  <action>
    Create `tokenise_test.go` (`package fuzzymatch_test`).

    1. AxonOps Apache-2.0 file header.

    2. Unit tests (table-driven where applicable):
       a. `TestTokenise_Empty` — `Tokenise("", DefaultTokeniseOptions())`
          returns a non-nil slice of length 0.
       b. `TestTokenise_WhitespaceOnly` — `Tokenise("   \t\n", ...)` returns
          a non-nil empty slice.
       c. `TestTokenise_SingleChar` — `Tokenise("a", ...)` returns `["a"]`.
       d. `TestTokenise_AlreadyTokenised` — `Tokenise("foo bar baz", ...)`
          returns `["foo", "bar", "baz"]`.
       e. `TestTokenise_CamelCase` — table covering FooBar, fooBar,
          parseJSON5, IOError.
       f. `TestTokenise_SnakeCase` — table covering foo_bar, FOO_BAR_BAZ,
          __leading__, trailing__.
       g. `TestTokenise_KebabCase` — table covering foo-bar, foo-bar-baz.
       h. `TestTokenise_DotCase` — table covering foo.bar, foo.bar.baz.
       i. `TestTokenise_PascalCase` — table covering FooBarBaz, ABCDef.
       j. `TestTokenise_ConsecutiveUpper` — XMLHTTPRequest, IPv4Address.
          The expected output reflects whichever SplitConsecutiveUpper
          convention lands per api-ergonomics-reviewer.
       k. `TestTokenise_MixedScript` — userПривет, fooПриветBaz. Cyrillic
          tokens preserved; ASCII tokens split correctly around them.
       l. `TestTokenise_Numeric` — Foo123Bar, parse5JSON. Document the
          chosen behaviour: digits stay attached to the preceding alpha run
          (default proposal).
       m. `TestTokenise_NoEmptyTokens` — `Tokenise("__foo__bar__", ...)`
          returns `["foo", "bar"]` (no empty strings).

    3. Property tests:
       a. `TestProp_Tokenise_OrderStable` — for arbitrary string s and
          arbitrary opts, calling `Tokenise(s, opts)` twice returns slices
          with identical contents in identical order (rules out map-iteration
          leak).
       b. `TestProp_Tokenise_NeverPanics` — for arbitrary string s
          (including invalid UTF-8 via quick.Value byte slices), Tokenise
          does not panic. Use defer-recover.
       c. `TestProp_Tokenise_NoEmptyTokens` — for arbitrary string s and
          DefaultTokeniseOptions, no token in the returned slice is empty
          (`len(t) == 0`).
       d. `TestProp_Tokenise_AllOutputUTF8` — every returned token is valid
          UTF-8 (`utf8.ValidString(t)` is true).
       e. `TestProp_Tokenise_TokenCount_LessOrEqualInputRunes` — for
          arbitrary s, `len(Tokenise(s, opts)) <= utf8.RuneCountInString(s)`
          (each token has at least one rune; tokens are non-overlapping).
       f. `TestProp_Tokenise_ReconstructibleASCII` — for ASCII input
          containing only [a-z A-Z _ - . space], joining the tokens with
          a single space and applying lowercase produces a string that is
          a subsequence of the lowercased input (informal correctness
          property — concrete bound depends on the chosen CamelCase
          semantics; pin loosely here, tighten when api-ergonomics-reviewer
          fixes the semantics).

    Concrete identifiers:
      - 13 unit tests + 6 property tests
      - Stdlib testing + testing/quick only (no testify)
  </action>
  <verify>
    <automated>go test -race -shuffle=on -count=1 -run 'TestTokenise_|TestProp_Tokenise_' ./... &amp;&amp; make verify-license-headers &amp;&amp; make lint &amp;&amp; make coverage &amp;&amp; make coverage-check</automated>
  </verify>
  <acceptance_criteria>
    - `tokenise_test.go` exists with the AxonOps Apache-2.0 header
    - Declares `package fuzzymatch_test`
    - Contains all 13 unit tests + 6 property tests as listed
    - `go test -race -shuffle=on -count=1 -run TestTokenise_ ./...` exits 0
    - `go test -race -shuffle=on -count=1 -run TestProp_Tokenise_ ./...` exits 0
    - Coverage on tokenise.go is ≥ 90% per file (verified via `go tool cover -func`); 100% on the public API surface (Tokenise, DefaultTokeniseOptions, TokeniseOptions)
    - `make coverage-check` exits 0
    - `make verify-license-headers` exits 0
  </acceptance_criteria>
  <done>
    Tokenise has full unit + property test coverage; order stability,
    non-empty output, valid UTF-8, panic-free all verified; the
    chosen splitting semantics are pinned in tests.
  </done>
</task>

<task type="auto">
  <name>Task 3: Write tokenise_bench_test.go and tokenise_fuzz_test.go + seed corpus</name>
  <files>tokenise_bench_test.go, tokenise_fuzz_test.go, testdata/fuzz/FuzzTokenise/seed-001</files>
  <read_first>
    - tokenise.go (just created)
    - tokenise_test.go (just created)
    - normalise_bench_test.go + normalise_fuzz_test.go (plan 01-06 — pattern reference)
    - .claude/skills/performance-standards/SKILL.md (Tokenise target: ASCII ≤ 50 chars < 500 ns, ≤ 2 allocations)
  </read_first>
  <action>
    1. `tokenise_bench_test.go`:
       - AxonOps Apache-2.0 header. Package `fuzzymatch_test`.
       - Benchmarks (all with `b.ReportAllocs()`):
         a. `BenchmarkTokenise_ASCII_Short` — 10-char identifier
         b. `BenchmarkTokenise_ASCII_Medium` — 50-char identifier
         c. `BenchmarkTokenise_ASCII_Long` — 500-char compound identifier
         d. `BenchmarkTokenise_Unicode_Short` — mixed-script 10-rune input
         e. `BenchmarkTokenise_PascalCase` — typical Pascal-case identifier
         f. `BenchmarkTokenise_DefaultOptions` — the most common path

    2. `tokenise_fuzz_test.go`:
       - AxonOps Apache-2.0 header. Package `fuzzymatch_test`.
       - `FuzzTokenise`:
         ```
         func FuzzTokenise(f *testing.F) {
             // Seed corpus: canonical reference inputs + edge cases.
             for _, s := range []string{
                 "", " ", "foo", "FooBar", "snake_case", "kebab-case",
                 "dot.case", "XMLHTTPRequest", "userПривет",
                 "\xff\xfe", "\xed\xa0\x80",
                 "Foo123Bar", "__leading", "trailing__",
             } {
                 for bits := uint8(0); bits < 16; bits++ {
                     f.Add(s, bits)
                 }
             }
             f.Fuzz(func(t *testing.T, s string, optBits uint8) {
                 opts := fuzzymatch.TokeniseOptions{
                     Lowercase:             optBits&1 != 0,
                     SplitCamelCase:        optBits&2 != 0,
                     SplitConsecutiveUpper: optBits&4 != 0,
                     SeparatorChars:        fuzzymatch.DefaultTokeniseOptions().SeparatorChars,
                 }
                 got := fuzzymatch.Tokenise(s, opts)
                 // Never returns nil for any non-error case.
                 if got == nil {
                     t.Errorf("Tokenise returned nil; expected non-nil slice for input %q", s)
                 }
                 // Each token is valid UTF-8 and non-empty.
                 for i, tok := range got {
                     if tok == "" {
                         t.Errorf("Tokenise returned empty token at index %d for input %q", i, s)
                     }
                     if !utf8.ValidString(tok) {
                         t.Errorf("Tokenise returned invalid UTF-8 token at index %d for input %q", i, s)
                     }
                 }
             })
         }
         ```

    3. `testdata/fuzz/FuzzTokenise/seed-001`:
       Hand-crafted seed in Go's fuzz file format:
       ```
       go test fuzz v1
       string("FooBar_baz.qux")
       byte(0x0f)
       ```
       Plus 4 additional seed files (seed-002 through seed-005) covering
       invalid UTF-8, lone surrogate, all-separators, and mixed-script
       cases — same pattern as plan 01-06.

    4. Run `go test -fuzz=FuzzTokenise -fuzztime=10s ./...` locally to
       confirm no crashes.

    Concrete identifiers:
      - 6 benchmarks
      - `FuzzTokenise` harness
      - 5 hand-crafted seed files in `testdata/fuzz/FuzzTokenise/`
  </action>
  <verify>
    <automated>go test -bench=BenchmarkTokenise -benchmem -count=1 -run=^$ ./... &amp;&amp; go test -fuzz=FuzzTokenise -fuzztime=5s -run=^$ ./... &amp;&amp; test -d testdata/fuzz/FuzzTokenise &amp;&amp; ls testdata/fuzz/FuzzTokenise/seed-* | wc -l | grep -qE '^\s*[5-9]'</automated>
  </verify>
  <acceptance_criteria>
    - `tokenise_bench_test.go` exists with 6 benchmarks
    - Each benchmark calls `b.ReportAllocs()`
    - `go test -bench=BenchmarkTokenise -benchmem -count=1 -run=^$ ./...` exits 0
    - `tokenise_fuzz_test.go` exists with `func FuzzTokenise(f *testing.F)`
    - `testdata/fuzz/FuzzTokenise/seed-001` through `seed-005` exist
    - `go test -fuzz=FuzzTokenise -fuzztime=5s -run=^$ ./...` exits 0
    - Benchmark output for `BenchmarkTokenise_ASCII_Short` shows ≤ 2 allocs/op (per performance-standards target); record numbers in SUMMARY
    - `make check` exits 0
  </acceptance_criteria>
  <done>
    Tokenise has full benchmark + fuzz coverage; the seed corpus exercises
    canonical inputs and known pitfall patterns; the file-header convention
    holds.
  </done>
</task>

</tasks>

<threat_model>
## Trust Boundaries

| Boundary | Description |
|----------|-------------|
| Consumer-supplied input → Tokenise | Arbitrary strings cross into Tokenise; panic-free fuzz property is the gate. |
| Tokenise → consumer (output slice) | Output is guaranteed `[]string{}` (non-nil), every token is valid UTF-8, every token is non-empty. Consumers downstream can rely on these invariants. |
| Tokenise → Phase 6 token algorithms | Phase 6's five token algorithms (Monge-Elkan, TokenSort/Set/Partial Ratio, TokenJaccard) consume Tokenise's output as the canonical token shape; any semantic drift in Tokenise (e.g. changing camelCase split semantics) requires a major version bump. |

## STRIDE Threat Register

| Threat ID | Category | Component | Disposition | Mitigation Plan |
|-----------|----------|-----------|-------------|-----------------|
| T-01-07-01 | Tampering | Order stability regression (map iteration leak) | mitigate | `TestProp_Tokenise_OrderStable` runs Tokenise twice on the same input and asserts byte-identical token slices; any future change introducing map iteration on the return path fails this test. |
| T-01-07-02 | Denial of Service | Pathological input slowing Tokenise (e.g. all-separators 1MB string) | mitigate | Tokenise is O(n) in input rune count; performance benchmarks for short / medium / long inputs detect quadratic regressions. |
| T-01-07-03 | Information Disclosure | Empty / nil ambiguity in return contract | mitigate | The non-nil empty-slice contract is asserted in `TestTokenise_Empty`, `TestTokenise_WhitespaceOnly`, and `TestProp_Tokenise_*` (the fuzz harness checks `got != nil`). |
| T-01-07-04 | Tampering | Invalid UTF-8 producing partial-rune tokens | mitigate | `TestProp_Tokenise_AllOutputUTF8` asserts every returned token is valid UTF-8; Tokenise operates via `[]rune(s)` which substitutes RuneError for invalid sequences. |
| T-01-07-05 | Repudiation | Panic on lone surrogate or invalid UTF-8 byte sequence | mitigate | FuzzTokenise exercises arbitrary byte input including invalid UTF-8 patterns; seed corpus covers known pitfalls. |
| T-01-07-06 | Tampering | A future change introducing init()-time table build, breaking determinism on Windows where init order can differ | mitigate | The file's structure (no init function, [128]bool sepASCII table built per-call) is asserted by `! grep -E '^func init\(\)' tokenise.go`; determinism-reviewer enforces at execution time. |
</threat_model>

<verification>
1. `go build ./...` and `go vet ./...` exit 0.
2. `make check` exits 0.
3. `go test -race -shuffle=on -count=1 -run 'TestTokenise_|TestProp_Tokenise_' ./...` exits 0.
4. `go test -fuzz=FuzzTokenise -fuzztime=5s -run=^$ ./...` exits 0.
5. `go test -bench=BenchmarkTokenise -benchmem -count=1 -run=^$ ./...` exits 0.
6. Coverage floors hold (≥ 90% per file, 100% public API).
7. `make verify-deps-allowlist` exits 0 (root go.mod still only x/text).
</verification>

<success_criteria>
- `Tokenise` implements the spec §10 splitter with camelCase/snake_case/
  PascalCase/kebab-case/dot-case support.
- `TokeniseOptions` struct with sensible defaults.
- Non-nil empty-slice contract for empty / whitespace-only input.
- Order stability property-tested.
- FuzzTokenise covers panic-free + non-empty + valid-UTF-8 properties.
- 100% public API coverage on tokenise.go.
- `make check` is green.
</success_criteria>

<output>
After completion, create
`.planning/phases/01-foundation-infrastructure/01-07-primitives-tokenise-SUMMARY.md`
recording:
  - The final TokeniseOptions field set (per api-ergonomics-reviewer
    at execution time).
  - The pinned semantics for SplitConsecutiveUpper (which exact split
    points produce which tokens for XMLHTTPRequest etc.).
  - Benchmark numbers (allocations + ns/op) — to be added to bench.txt by
    the user's `make bench` workstation run.
  - Whether any property test caught a real bug during development
    (informational).
</output>
