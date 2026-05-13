---
plan_id: 01-06
phase: 01-foundation-infrastructure
plan: 06
type: execute
wave: 6
depends_on:
  - 01-05
autonomous: true
objective: >
  Land `Normalise(s string, opts NormalisationOptions) string` with ASCII fast
  path, Unicode NFC/NFD pipeline via `golang.org/x/text/unicode/norm`,
  optional diacritic stripping, snake_case/camelCase/PascalCase/kebab-case/
  dot-case splitting. Populate `testdata/golden/normalisation.json` with 20-40
  entries per D-11 exercising every code path. The golden file uses the
  canonical format LOCKED in plan 01-04. Property tests verify idempotence,
  length-bound, and ASCII-input equivalence. The first real `TestGolden_*`
  case in the project lands here.
files_modified:
  - normalise.go
  - normalise_test.go
  - normalise_bench_test.go
  - normalise_fuzz_test.go
  - testdata/golden/normalisation.json
  - testdata/fuzz/FuzzNormalise/seed.txt
requirements:
  - FOUND-03
  - DET-05
  - DET-06
  - TEST-03
must_haves:
  truths:
    - "`Normalise` has an ASCII fast path returning before invoking x/text norm.Transformer when all bytes < 0x80"
    - "Unicode pipeline uses `golang.org/x/text/unicode/norm` (NFC) and optionally `transform.Chain(norm.NFD, runes.Remove(runes.In(unicode.Mn)), norm.NFC)` for diacritic stripping"
    - "`DefaultNormalisationOptions()` returns Lowercase=true, StripSeparators=true, SeparatorChars=\"_-.:/\", SplitCamelCase=true, NFC=true, StripDiacritics=false (D-03)"
    - "Normalise contains NO `math.Pow`, `math.Log`, `math.Exp`, `math.FMA` (DET-06)"
    - "Normalise contains NO `init()`-time table builds; all tables are `var x = literal` (§13.5)"
    - "Normalise contains NO map iteration on output paths (DET-03)"
    - "`testdata/golden/normalisation.json` contains 20-40 pinned (input, options, expected output) entries per D-11"
    - "Golden file uses the canonical format LOCKED in golden_canonical.go from plan 01-04"
    - "TestGolden_Normalisation passes on the local platform AND will pass on all 5 CI matrix platforms (D-14)"
    - "FuzzNormalise asserts panic-free on arbitrary input including invalid UTF-8 (TEST-03)"
  artifacts:
    - path: normalise.go
      provides: NormalisationOptions struct, DefaultNormalisationOptions(), Normalise function with ASCII fast path + Unicode pipeline
      contains: "func Normalise"
    - path: normalise_test.go
      provides: Unit + property tests for Normalise, plus TestGolden_Normalisation
      contains: "TestGolden_Normalisation"
    - path: normalise_bench_test.go
      provides: ASCII-short / ASCII-long / Unicode-short / Unicode-long benchmarks
    - path: normalise_fuzz_test.go
      provides: FuzzNormalise — panic-free property under arbitrary bytes
    - path: testdata/golden/normalisation.json
      provides: 20-40 canonical (input, opts, expected) entries
    - path: testdata/fuzz/FuzzNormalise/seed.txt
      provides: Seed corpus for FuzzNormalise (canonical reference vectors)
  key_links:
    - from: normalise.go
      to: golang.org/x/text/unicode/norm
      via: import golang.org/x/text/unicode/norm
      pattern: "golang.org/x/text/unicode/norm"
    - from: normalise_test.go
      to: testdata/golden/normalisation.json
      via: assertGolden(t, ...) helper from plan 01-04
      pattern: "assertGolden"
    - from: normalise_test.go
      to: golden_canonical.go
      via: TestGolden_Normalisation uses canonicalMarshal to produce diff'd bytes
      pattern: "canonicalMarshal\\|assertGolden"
---

<objective>
Land the Normalise pipeline — the FIRST production-data-shaping API in
fuzzymatch and the FIRST consumer of the locked golden-file format.

Purpose: Every algorithm phase from Phase 2 onwards calls Normalise (directly
or via the Scorer's normalisation step). Getting the ASCII fast path, the
Unicode pipeline, the option defaults, and the cross-platform determinism
right NOW prevents 23 algorithm-phase rewrites later.

Output:
  - normalise.go (production code)
  - normalise_test.go (unit + property + golden)
  - normalise_bench_test.go (benchmarks)
  - normalise_fuzz_test.go (panic-free fuzz)
  - testdata/golden/normalisation.json (20-40 entries — first real golden)
  - testdata/fuzz/FuzzNormalise/seed.txt (seed corpus)
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
@.claude/skills/go-coding-standards/SKILL.md
@.claude/skills/go-testing-standards/SKILL.md
@.claude/skills/determinism-standards/SKILL.md
@.claude/skills/performance-standards/SKILL.md
@.claude/skills/algorithm-correctness-standards/SKILL.md
@docs/requirements.md
@doc.go
@golden_canonical.go
@golden_test.go
@algoid.go
@errors.go

<interfaces>
Locked decisions:
- D-02: `Normalise(s string, opts NormalisationOptions) string` — struct-of-options.
- D-03: defaults — Lowercase=true, StripSeparators=true, SeparatorChars="_-.:/",
  SplitCamelCase=true, NFC=true, StripDiacritics=false.
- D-11: golden file `normalisation.json` populated with 20-40 entries.
- D-12: canonical format LOCKED — use the canonicalMarshal helper from plan
  01-04.
- D-13: `-update` flag regenerates from current code.
- D-14: cross-platform diff per platform against the single committed file.

NormalisationOptions struct shape (concrete field names per
api-ergonomics-reviewer at execution time; defaults below mirror the spec):

  type NormalisationOptions struct {
      Lowercase        bool
      StripSeparators  bool
      SeparatorChars   string  // characters treated as separators when StripSeparators is true; replaced with a single space
      SplitCamelCase   bool   // insert a space between lowercase->uppercase transitions
      NFC              bool   // run Unicode NFC normalisation
      StripDiacritics  bool   // remove combining marks (requires NFC=true implicitly; helper handles)
  }

  func DefaultNormalisationOptions() NormalisationOptions {
      return NormalisationOptions{
          Lowercase:       true,
          StripSeparators: true,
          SeparatorChars:  "_-.:/",
          SplitCamelCase:  true,
          NFC:             true,
          StripDiacritics: false,
      }
  }

ASCII fast path (per performance-standards):

  func isASCII(s string) bool {
      for i := 0; i < len(s); i++ {
          if s[i] >= 0x80 { return false }
      }
      return true
  }

  When isASCII(s) AND !opts.StripDiacritics AND !opts.NFC:
    Skip x/text invocation entirely. Apply Lowercase (bitwise OR 0x20 for
    'A'..'Z'), StripSeparators (replace bytes in opts.SeparatorChars with
    space, then collapse runs of spaces; whitespace already a separator),
    SplitCamelCase (lowercase->uppercase byte-boundary insertion) all in a
    single byte-slice pass. Use a stack-allocated buffer for inputs ≤ 64
    bytes per performance-standards "Stack-Allocated Buffer" section.

Unicode pipeline (when !isASCII OR StripDiacritics):
  - If StripDiacritics:
      norm.NFD → runes.Remove(runes.In(unicode.Mn)) → norm.NFC
    Build the transformer via `transform.Chain(...)`. The transformer
    instance is created PER CALL (cheap — x/text transformers are zero-alloc
    to construct) OR cached as a package-level `var` (preferred — avoids the
    per-call construction cost but requires goroutine-safety guarantees from
    x/text. Per x/text docs, transform.Transformer is NOT safe for
    concurrent use of a single instance. Solution: create a package-level
    `sync.Pool` of transformers. But D-09's "no goroutines, no channels"
    constraint and §5(12) "no init()-time table builds" require care.
    Per-call construction is acceptable for v1 and matches mask's simpler
    patterns. Document the trade-off in the godoc and revisit in v1.x if
    benchmark shows it dominates.).
  - Else if NFC: just norm.NFC.
  - Else: pass through without x/text.

After Unicode pipeline, apply Lowercase / StripSeparators / SplitCamelCase
on the resulting rune sequence (using `unicode.ToLower` for non-ASCII
runes; ASCII fast path within this loop uses the bitwise OR).

CRITICAL constraints (per determinism-standards + go-coding-standards):
- NO `math.X` calls beyond Sqrt/Abs/Min/Max — Normalise doesn't need
  floats anyway, but the constraint is recorded.
- NO `init()` function. Any pre-built table is `var x = literal` or
  `var x = buildX()` where `buildX` is a pure function.
- NO map iteration on output. Internal maps (e.g. a `map[rune]bool` for
  the separator set) are fine but the iteration must not leak into output
  ordering.
- NO goroutines, channels, or mutexes.

Golden-file structure for `testdata/golden/normalisation.json` — proposed
shape (per api-ergonomics-reviewer at execution time):

  // GoldenNormalisationEntry: one input/output case. Used by
  // TestGolden_Normalisation to build the canonical-form JSON.
  type goldenNormalisationEntry struct {
      Name            string                 `json:"name"`
      Input           string                 `json:"input"`
      Options         NormalisationOptions   `json:"options"`
      ExpectedOutput  string                 `json:"expected_output"`
  }

  type goldenNormalisationFile struct {
      Version int                          `json:"version"`  // = 1 for v1.x
      Entries []goldenNormalisationEntry   `json:"entries"`
  }

  Entries are sorted by Name (alphabetic) so the slice order is
  deterministic and the canonical-form byte output is stable.

20-40 entries per D-11 covering:
  - 5-10 pure-ASCII cases (FooBar, foo_bar, foo.bar.baz, hello-world-here,
    HTTPSConnection)
  - 5-10 NFC/NFD-divergent cases (café precomposed → café normalised;
    café decomposed (e + ́) → café normalised; both produce the same NFC form)
  - 3-5 StripDiacritics ON+OFF pairs (Müller, naïve, résumé)
  - 3-5 mixed-script preservation cases (Привет, مرحبا, 你好) — these
    should NOT be stripped or altered beyond NFC
  - 3-5 separator-edge cases (empty, single-char, multiple separators,
    leading/trailing separator)
  - 3-5 camelCase splits (XMLHTTPRequest, htmlBody, IOError, parseJSON5)
  - 3-5 idempotence pairs (apply Normalise twice; result identical to one
    application)
</interfaces>
</context>

<tasks>

<task type="auto">
  <name>Task 1: Write normalise.go (NormalisationOptions, DefaultNormalisationOptions, Normalise)</name>
  <files>normalise.go</files>
  <read_first>
    - .planning/phases/01-foundation-infrastructure/01-CONTEXT.md (D-02, D-03)
    - docs/requirements.md §9 (Normalise pipeline)
    - .claude/skills/determinism-standards/SKILL.md (Init() rule; No-Map-Iteration; Float Stability — Normalise has no floats so DET-06 is satisfied trivially)
    - .claude/skills/performance-standards/SKILL.md (ASCII Fast Path; Stack-Allocated Buffer)
    - .claude/skills/algorithm-correctness-standards/SKILL.md (Score Normalisation — N/A here; Unicode Handling)
    - .claude/skills/go-coding-standards/SKILL.md (API Design; Files — Apache-2.0 header)
    - algoid.go (plan 01-05 — file structure reference)
    - errors.go (plan 01-05 — file structure reference)
    - go.mod (confirm `golang.org/x/text` is present)
  </read_first>
  <action>
    Create `normalise.go` (`package fuzzymatch`).

    1. AxonOps Apache-2.0 file header.

    2. File-level block comment citing the spec:
       ```
       // Normalise applies caller-controlled normalisation: ASCII case-folding,
       // separator stripping, camel-case splitting, and Unicode NFC
       // normalisation with optional diacritic stripping. See
       // docs/requirements.md §9 for the authoritative spec.
       //
       // Implementation: ASCII fast path operates on bytes for inputs whose
       // every byte is < 0x80; Unicode path uses golang.org/x/text/unicode/norm
       // for NFC and transform.Chain(NFD, runes.Remove(Mn), NFC) for diacritic
       // stripping. NO init()-time tables; NO map iteration on output paths;
       // NO transcendental float ops (DET-06).
       ```

    3. Imports: `unicode`, `unicode/utf8`, `strings`,
       `golang.org/x/text/transform`, `golang.org/x/text/unicode/norm`,
       `golang.org/x/text/unicode/runenames` (if needed),
       `golang.org/x/text/runes`. Use `goimports -local
       github.com/axonops/fuzzymatch` to enforce import grouping.

    4. Public types and helpers:
       ```
       // NormalisationOptions configures the Normalise pipeline. Per-call
       // allocation is zero because the struct is passed by value.
       type NormalisationOptions struct {
           Lowercase        bool
           StripSeparators  bool
           SeparatorChars   string
           SplitCamelCase   bool
           NFC              bool
           StripDiacritics  bool
       }

       // DefaultNormalisationOptions returns the conservative default set
       // matching docs/requirements.md §9: Lowercase, separator-strip,
       // camel-split, NFC enabled; diacritic stripping DISABLED. Consumers
       // wanting "café → cafe" semantics opt in explicitly.
       func DefaultNormalisationOptions() NormalisationOptions {
           return NormalisationOptions{
               Lowercase:       true,
               StripSeparators: true,
               SeparatorChars:  "_-.:/",
               SplitCamelCase:  true,
               NFC:             true,
               StripDiacritics: false,
           }
       }

       // Normalise returns s with the requested normalisation applied. Empty
       // input returns empty. Inputs that are pure ASCII (no byte >= 0x80)
       // and do not require Unicode NFC follow a stack-allocated fast path
       // with zero allocations for inputs <= 64 bytes (see
       // .claude/skills/performance-standards/SKILL.md). Non-ASCII inputs or
       // inputs with NFC/StripDiacritics enabled go through the
       // golang.org/x/text/unicode/norm pipeline.
       func Normalise(s string, opts NormalisationOptions) string {
           // ... implementation per <interfaces>
       }
       ```

    5. Internal helpers (all unexported, all pure functions, no init()):
       - `func isASCII(s string) bool`
       - `func toLowerASCIIInPlace(buf []byte) []byte`
       - `func stripSeparatorsASCII(buf []byte, sepSet [128]bool) []byte`
       - `func splitCamelCaseASCII(buf []byte) []byte`
       - `func buildSepSet(s string) [128]bool` — builds a stack-allocated
         128-byte boolean table for ASCII separator lookup (each byte of
         opts.SeparatorChars with value < 128 sets the corresponding index
         to true). Non-ASCII separator chars handled in the Unicode path.
       - `func normaliseUnicode(s string, opts NormalisationOptions) string`
         — the slow path using x/text.

    6. The Unicode-path implementation:
       - Build the transformer:
         - If StripDiacritics: `transform.Chain(norm.NFD,
           runes.Remove(runes.In(unicode.Mn)), norm.NFC)`
         - Else if NFC: `norm.NFC`
         - Else: no transformer (input passes through)
       - Apply transformer: `result, _, _ := transform.String(t, s)`
       - Apply remaining options (Lowercase / StripSeparators /
         SplitCamelCase) on the resulting rune sequence using `unicode.ToLower`
         for non-ASCII runes and a per-rune separator check (the SeparatorChars
         set may contain multi-byte UTF-8 chars; iterate via
         `range result` to get runes correctly).

    7. NO `init()` function. All tables (`buildSepSet` etc.) are built
       per-call from the options struct — passing the struct by value, the
       table lives on the stack and is freed at function return.

    8. NO map iteration on output. The internal sepSet is a `[128]bool`
       array; the Unicode-path separator check uses
       `strings.ContainsRune(opts.SeparatorChars, r)` which is O(len(SepChars))
       — acceptable for the typical short separator strings.

    9. Edge cases (per algorithm-correctness-standards Unicode Handling):
       - Empty input → empty output
       - Invalid UTF-8 in input: never panic. The x/text norm.Transformer
         tolerates invalid UTF-8 by substituting `utf8.RuneError` for
         malformed sequences. Normalise propagates this — invalid UTF-8 in
         produces a result with RuneError-replaced segments out. Documented
         in the godoc as "invalid UTF-8 is replaced with U+FFFD per Go's
         standard behaviour".

    Concrete identifiers (per api-ergonomics-reviewer at execution time;
    defaults below):
      - File `normalise.go`, package `fuzzymatch`
      - Type `NormalisationOptions` with 6 bool/string fields
      - `DefaultNormalisationOptions() NormalisationOptions`
      - `Normalise(s string, opts NormalisationOptions) string`
      - Internal helpers: isASCII, toLowerASCIIInPlace,
        stripSeparatorsASCII, splitCamelCaseASCII, buildSepSet,
        normaliseUnicode

    Note (Design Principle 13): the exact field names of NormalisationOptions
    (e.g. `Lowercase` vs `ToLower`, `StripSeparators` vs `RemoveSeparators`,
    `SeparatorChars` vs `Separators`, etc.) and the exact function name
    (`Normalise` vs `NormaliseString` vs `Canonicalise`) are
    api-ergonomics-reviewer's call. The default proposal mirrors
    docs/requirements.md §9.
  </action>
  <verify>
    <automated>go build ./... &amp;&amp; go vet ./... &amp;&amp; make verify-license-headers &amp;&amp; make lint &amp;&amp; ! grep -E 'math\.(Pow|Log|Exp|FMA)' normalise.go &amp;&amp; ! grep -E '^func init\(\)' normalise.go &amp;&amp; grep -q 'golang.org/x/text/unicode/norm' normalise.go</automated>
  </verify>
  <acceptance_criteria>
    - `normalise.go` exists with the AxonOps Apache-2.0 header
    - Declares `type NormalisationOptions struct { ... }` with 6 fields
    - Declares `func DefaultNormalisationOptions() NormalisationOptions` returning the D-03 defaults exactly
    - Declares `func Normalise(s string, opts NormalisationOptions) string`
    - Imports `golang.org/x/text/unicode/norm` and `golang.org/x/text/transform`
    - Contains NO `func init()`
    - Contains NO `math.Pow`, `math.Log`, `math.Exp`, `math.FMA`
    - Contains NO `range mapVar` pattern on any return path (internal map iteration in tests is fine — this is production code)
    - Empty-input behaviour: `Normalise("", DefaultNormalisationOptions()) == ""`
    - `Normalise("café", DefaultNormalisationOptions())` preserves diacritics (returns "café")
    - `Normalise("café", NormalisationOptions{Lowercase: true, NFC: true, StripDiacritics: true})` returns "cafe"
    - `Normalise("FooBar", DefaultNormalisationOptions())` returns "foo bar" (camelCase split + lowercase + sep-strip is not applied since FooBar has no SeparatorChars match)
    - `make verify-license-headers` exits 0
    - `make lint` exits 0
    - `make verify-deps-allowlist` exits 0 (the import of x/text/unicode/norm is allowed; root go.mod did not change)
  </acceptance_criteria>
  <done>
    Normalise compiles and lints cleanly, follows the ASCII fast path / Unicode
    pipeline split, has zero init-time work, and respects D-02 / D-03 / DET-06.
  </done>
</task>

<task type="auto">
  <name>Task 2: Write normalise_test.go (unit + property tests + TestGolden_Normalisation) and populate normalisation.json</name>
  <files>normalise_test.go, testdata/golden/normalisation.json</files>
  <read_first>
    - normalise.go (just created)
    - golden_test.go + golden_canonical.go (plan 01-04)
    - .planning/phases/01-foundation-infrastructure/01-CONTEXT.md (D-11 — 20-40 golden entries; D-12 — canonical format LOCKED; D-13 — `-update` flag; D-14 — per-platform diff)
    - .claude/skills/go-testing-standards/SKILL.md (Stdlib testing only; Property tests via testing/quick; Float comparisons N/A here since no floats)
    - .claude/skills/determinism-standards/SKILL.md (Golden Files section)
    - docs/requirements.md §9 (Normalise pipeline reference vectors — citation source for golden entries)
    - .planning/research/PITFALLS.md (Unicode normalisation pitfalls, mixed-script preservation)
  </read_first>
  <action>
    Create `normalise_test.go` (`package fuzzymatch_test`) AND populate
    `testdata/golden/normalisation.json`. The test file is written FIRST
    with the helper struct definitions; the JSON file is then GENERATED by
    running `go test -run TestGolden_Normalisation -update` once and
    committing the result. The executor verifies the resulting JSON
    bytes match D-12's canonical form before committing.

    1. `normalise_test.go`:
       - AxonOps Apache-2.0 file header.
       - Package: `fuzzymatch_test`.
       - Imports: `testing`, `testing/quick`, `unicode/utf8`, `strings`,
         `github.com/axonops/fuzzymatch`.
       - Unit tests:
         a. `TestNormalise_Empty` — `Normalise("", DefaultNormalisationOptions()) == ""`
         b. `TestNormalise_DefaultsPreserveDiacritics` —
            `Normalise("Müller", DefaultNormalisationOptions())` returns
            `"müller"` (lowercased, diacritic preserved).
         c. `TestNormalise_StripDiacritics` — with `StripDiacritics: true`,
            `Müller → müller → muller`. Use a small table for cafe / Mueller /
            naive / resume / café variants.
         d. `TestNormalise_NFC_Idempotent` — both NFC-precomposed `"café"`
            and NFD-decomposed `"café"` produce identical output under
            `NormalisationOptions{NFC: true}`.
         e. `TestNormalise_ASCII_FastPath_DoesNotAlterUnicodeOutput` —
            ASCII-only input produces identical output whether or not
            NFC is enabled (the ASCII fast path is bit-equivalent to the
            Unicode path for ASCII-only input).
         f. `TestNormalise_SplitCamelCase` — table covering FooBar →
            "foo bar", XMLHTTPRequest → "xmlhttp request" (typical edge
            case for consecutive uppercase runs — document the chosen
            behaviour: split only at lowercase→uppercase transitions, NOT
            at uppercase→lowercase. Or document the alternative if
            api-ergonomics-reviewer picks differently.).
         g. `TestNormalise_StripSeparators` — table covering "foo_bar.baz"
            → "foo bar baz" (under-multipath collapses) etc.
         h. `TestNormalise_PreservesMixedScript` — Cyrillic Привет, Arabic
            مرحبا, CJK 你好 all return unchanged under NFC+lowercase (Cyrillic
            does have case so `Привет → привет` is expected; CJK has no
            case so output equals input).
       - Property tests (testing/quick):
         a. `TestProp_Normalise_Idempotent` — for arbitrary string s,
            `Normalise(Normalise(s, opts), opts) == Normalise(s, opts)` for
            the default options.
         b. `TestProp_Normalise_NeverPanics` — for arbitrary string s
            (including invalid UTF-8 via `quick.Value` arbitrary bytes),
            `Normalise(s, opts)` does not panic. Use defer-recover sentinel.
         c. `TestProp_Normalise_LengthBound_WhenStripSeparators` — for
            arbitrary s with StripSeparators=true,
            `utf8.RuneCountInString(Normalise(s, opts)) <= utf8.RuneCountInString(s) + 1`
            (the "+1" accounts for camelCase split potentially inserting one
            space — adjust the bound after empirical observation).
       - Golden test:
         ```
         func TestGolden_Normalisation(t *testing.T) {
             entries := goldenNormalisationEntries()  // helper in this file
             file := goldenNormalisationFile{Version: 1, Entries: entries}
             assertGolden(t, "testdata/golden/normalisation.json", file)
         }

         // goldenNormalisationEntries returns the 20-40 cases that pin
         // Normalise behaviour across the v1.x lifetime. Entries MUST be
         // sorted by Name (alphabetic) so the JSON byte output is stable.
         func goldenNormalisationEntries() []goldenNormalisationEntry {
             // ... 25-30 entries ...
             // After construction:
             sort.Slice(entries, func(i, j int) bool { return entries[i].Name < entries[j].Name })
             return entries
         }
         ```

         The entries cover (per D-11):
         - PureASCII_FooBar
         - PureASCII_snake_case
         - PureASCII_dot.case
         - PureASCII_kebab-case
         - PureASCII_HTTPSConnection
         - PureASCII_XMLHTTPRequest
         - PureASCII_parseJSON5
         - NFC_NFD_Cafe_Precomposed
         - NFC_NFD_Cafe_Decomposed
         - NFC_Muller_PreserveDiacritic
         - NFC_Muller_StripDiacritic
         - NFC_Resume_Diacritic
         - NFC_Naive_Diacritic
         - NFC_Resume_StripDiacritic
         - MixedScript_Cyrillic_Privet
         - MixedScript_Arabic_Marhaba
         - MixedScript_CJK_NiHao
         - Idempotence_DoubleApply_Hello
         - Idempotence_DoubleApply_Cafe
         - Sep_Empty
         - Sep_SingleSeparator
         - Sep_MultipleSeparators
         - Sep_LeadingTrailing
         - Camel_Simple
         - Camel_ConsecutiveUpper
         - Empty_String
         - Edge_Whitespace_Only
         - Edge_Single_Char
         - Edge_Long_Mixed_60_chars

         (≈ 29 entries — within D-11's 20-40 bound.)

       - The helper struct types `goldenNormalisationEntry` and
         `goldenNormalisationFile` are declared in this test file (NOT in
         normalise.go — they exist only for the test harness).

    2. Run `go test -run TestGolden_Normalisation -update ./...` ONCE to
       generate `testdata/golden/normalisation.json` from the current code.

    3. Verify the resulting JSON satisfies D-12's canonical form:
       a. `head -c 3 testdata/golden/normalisation.json | xxd` — first 3
          bytes are NOT `ef bb bf` (no BOM).
       b. `tail -c 1 testdata/golden/normalisation.json | xxd` — last byte
          is `0a` (single LF, no CRLF).
       c. `grep -P '^\t' testdata/golden/normalisation.json` returns
          nothing (no tab indentation).
       d. `grep -P '  [^ ]' testdata/golden/normalisation.json | head -5`
          shows two-space indented lines (canonical-form indent).
       e. The JSON parses cleanly: `python3 -m json.tool testdata/golden/normalisation.json > /dev/null`.

    4. Re-run `go test -run TestGolden_Normalisation ./...` WITHOUT `-update`.
       MUST exit 0 (diff-clean). If it doesn't, the canonical form is
       drifting — investigate before committing.

    Concrete identifiers:
      - File `normalise_test.go`, package `fuzzymatch_test`
      - 8 unit tests + 3 property tests + 1 golden test
      - Golden helper structs `goldenNormalisationEntry`, `goldenNormalisationFile`
      - Golden file path `testdata/golden/normalisation.json`
      - Entry count ~29 (within 20-40 per D-11)
      - Entries sorted by Name (alphabetic) before marshalling
  </action>
  <verify>
    <automated>go test -race -shuffle=on -count=1 -run 'TestNormalise_|TestProp_Normalise_|TestGolden_Normalisation' ./... &amp;&amp; test -f testdata/golden/normalisation.json &amp;&amp; python3 -m json.tool testdata/golden/normalisation.json &gt; /dev/null &amp;&amp; python3 -c "import json; d=json.load(open('testdata/golden/normalisation.json')); assert d['version']==1; assert 20 &lt;= len(d['entries']) &lt;= 40, f'entry count {len(d[\"entries\"])} outside D-11 [20,40]'; print(f'OK: {len(d[\"entries\"])} entries')" &amp;&amp; test "$(tail -c 1 testdata/golden/normalisation.json | od -An -c | tr -d ' ')" = '\n' &amp;&amp; ! head -c 3 testdata/golden/normalisation.json | grep -Fq $'\xef\xbb\xbf'</automated>
  </verify>
  <acceptance_criteria>
    - `normalise_test.go` exists with the AxonOps Apache-2.0 header
    - Declares `package fuzzymatch_test`
    - Contains all 8 unit tests, 3 property tests, and `TestGolden_Normalisation`
    - `testdata/golden/normalisation.json` exists
    - Entry count is between 20 and 40 (inclusive) — verified by python3 json parsing
    - Top-level JSON has a `version: 1` field and an `entries: [...]` array
    - Entries are sorted by Name (alphabetic) — verifiable by parsing and asserting `sorted(names) == names`
    - File ends with exactly one `\n` byte (single LF)
    - File does NOT start with a BOM (first 3 bytes are not `ef bb bf`)
    - File contains no tab characters in indentation
    - File contains no CRLF line endings
    - `go test -run TestGolden_Normalisation ./...` (without `-update`) exits 0 (diff-clean)
    - `go test -run 'TestProp_Normalise_' ./...` exits 0 (property tests pass)
    - `make verify-determinism` exits 0
    - `make verify-license-headers` exits 0
  </acceptance_criteria>
  <done>
    Normalise has full unit + property test coverage; TestGolden_Normalisation
    is the FIRST real golden test in the project; the canonical-form contract
    from plan 01-04 is exercised; the file passes diff-clean on the local
    platform AND is the contract every other CI matrix platform will be
    asserted against.
  </done>
</task>

<task type="auto">
  <name>Task 3: Write normalise_bench_test.go and normalise_fuzz_test.go + seed corpus</name>
  <files>normalise_bench_test.go, normalise_fuzz_test.go, testdata/fuzz/FuzzNormalise/seed.txt</files>
  <read_first>
    - normalise.go (just created)
    - normalise_test.go (just created)
    - .claude/skills/performance-standards/SKILL.md (Benchmark File Structure section — ASCII-short/medium/long + Unicode-short/long)
    - .claude/skills/go-testing-standards/SKILL.md (Fuzz Tests section; benchmark structure)
    - .planning/research/PITFALLS.md (malformed UTF-8 — fuzz target)
  </read_first>
  <action>
    Create the benchmark file, fuzz file, and the seed corpus.

    1. `normalise_bench_test.go`:
       - AxonOps Apache-2.0 file header.
       - Package: `fuzzymatch_test`.
       - Imports: `testing`, `github.com/axonops/fuzzymatch`.
       - Benchmarks (`b.ReportAllocs()` on every one):
         a. `BenchmarkNormalise_ASCII_Short` — 10-char ASCII input
         b. `BenchmarkNormalise_ASCII_Medium` — 50-char ASCII input
         c. `BenchmarkNormalise_ASCII_Long` — 500-char ASCII input
         d. `BenchmarkNormalise_Unicode_Short` — 10-rune mixed input
         e. `BenchmarkNormalise_Unicode_Long` — 500-rune mixed input
         f. `BenchmarkNormalise_StripDiacritics_Short` — diacritic-rich
            short input with StripDiacritics enabled
         g. `BenchmarkNormalise_DefaultOptions_Short` — using
            `DefaultNormalisationOptions()` to exercise the most-common path
       - Targets per performance-standards: ASCII ≤ 50 chars: < 200 ns, 0
         allocations. The benchmarks DO NOT assert these targets at runtime
         (allocation assertions live in benchstat regression detection); they
         just produce numbers. The targets are documented in the file's
         block comment.

    2. `normalise_fuzz_test.go`:
       - AxonOps Apache-2.0 file header.
       - Package: `fuzzymatch_test`.
       - Imports: `testing`, `unicode/utf8`, `github.com/axonops/fuzzymatch`.
       - `FuzzNormalise`:
         ```
         func FuzzNormalise(f *testing.F) {
             // Seed corpus from canonical reference vectors.
             f.Add("", fuzzymatch.DefaultNormalisationOptions()) // can't pass struct directly to Fuzz; instead seed encoded options as bits
             // Seed strategy: encode options as a uint8 bitfield.
             // Bits: 0=Lowercase, 1=StripSeparators, 2=SplitCamelCase, 3=NFC, 4=StripDiacritics.
             // SeparatorChars hardcoded to DefaultNormalisationOptions().SeparatorChars.
             for _, s := range []string{"", "café", "FooBar", "naïve", "\xff\xfe", "\xed\xa0\x80", "Müller"} {
                 for bits := uint8(0); bits < 32; bits++ {
                     f.Add(s, bits)
                 }
             }
             f.Fuzz(func(t *testing.T, s string, optBits uint8) {
                 opts := fuzzymatch.NormalisationOptions{
                     Lowercase:       optBits&1 != 0,
                     StripSeparators: optBits&2 != 0,
                     SeparatorChars:  fuzzymatch.DefaultNormalisationOptions().SeparatorChars,
                     SplitCamelCase:  optBits&4 != 0,
                     NFC:             optBits&8 != 0,
                     StripDiacritics: optBits&16 != 0,
                 }
                 // Must not panic on arbitrary bytes (including invalid UTF-8).
                 got := fuzzymatch.Normalise(s, opts)
                 // Sanity: the output must be valid UTF-8 (x/text norm
                 // replaces invalid bytes with U+FFFD, so this should hold).
                 if !utf8.ValidString(got) {
                     t.Errorf("Normalise produced invalid UTF-8 from input %q", s)
                 }
             })
         }
         ```

    3. `testdata/fuzz/FuzzNormalise/seed.txt`:
       The native Go fuzz tooling auto-creates the per-corpus directory
       layout under `testdata/fuzz/FuzzNormalise/`. Manually seeded entries
       belong in `testdata/fuzz/FuzzNormalise/seed-XXX` files, one per
       case, in Go's fuzz file format (`go test fuzz v1\n<typed args>`).
       For plan 01-06, write 5 hand-crafted seed entries covering:
         - Empty string + all-bits-zero options
         - "café" + DefaultNormalisationOptions encoded
         - Invalid UTF-8 byte `\xff\xfe` + various options
         - Lone surrogate `\xed\xa0\x80` + various options
         - Long mixed-script "Müller Привет 你好" + StripDiacritics on
       Each seed lives in its own file (`testdata/fuzz/FuzzNormalise/seed-001`
       through `seed-005`). The file format is Go's fuzz corpus format:
         ```
         go test fuzz v1
         string("<input>")
         byte(<optBits>)
         ```

    4. Run `go test -fuzz=FuzzNormalise -fuzztime=10s ./...` locally and
       confirm no crashes. The 10-second smoke test is enough for plan
       01-06; CI's nightly fuzz runs longer.

    Concrete identifiers:
      - File `normalise_bench_test.go` — 7 benchmarks
      - File `normalise_fuzz_test.go` — `FuzzNormalise` fuzz harness
      - Directory `testdata/fuzz/FuzzNormalise/` with 5 seed files
      - Fuzz target: panic-free + output is valid UTF-8
  </action>
  <verify>
    <automated>go test -bench=BenchmarkNormalise -benchmem -count=1 -run=^$ ./... &amp;&amp; go test -fuzz=FuzzNormalise -fuzztime=5s -run=^$ ./... &amp;&amp; test -d testdata/fuzz/FuzzNormalise &amp;&amp; ls testdata/fuzz/FuzzNormalise/seed-* | wc -l | grep -qE '^\s*[5-9]'</automated>
  </verify>
  <acceptance_criteria>
    - `normalise_bench_test.go` exists with 7 benchmarks
    - Each benchmark calls `b.ReportAllocs()`
    - `go test -bench=BenchmarkNormalise -benchmem -count=1 -run=^$ ./...` exits 0
    - `normalise_fuzz_test.go` exists with `func FuzzNormalise(f *testing.F)`
    - `testdata/fuzz/FuzzNormalise/seed-001` through `seed-005` exist (5 hand-crafted seeds)
    - `go test -fuzz=FuzzNormalise -fuzztime=5s -run=^$ ./...` exits 0 (no panics in 5 seconds of fuzzing)
    - Benchmark output for `BenchmarkNormalise_ASCII_Short` shows 0 allocs/op (verifiable from `b.ReportAllocs` output) — if it shows >0, document in SUMMARY and flag for algorithm-performance-reviewer
    - `make verify-license-headers` exits 0
    - `make check` exits 0 (everything green)
  </acceptance_criteria>
  <done>
    Normalise has full benchmark + fuzz coverage; the seed corpus exercises
    canonical reference vectors and the malformed-UTF-8 panic-free property
    (TEST-03); the file-header convention holds.
  </done>
</task>

</tasks>

<threat_model>
## Trust Boundaries

| Boundary | Description |
|----------|-------------|
| Consumer-supplied input → Normalise | Arbitrary strings (potentially malformed UTF-8, embedded nulls, very long inputs) cross into Normalise; the panic-free fuzz property is the gate. |
| Normalise → consumer (output) | Output is guaranteed valid UTF-8 (per the fuzz assertion); consumers downstream of Normalise can assume valid UTF-8. |
| Normalise → x/text transformer | x/text is the curated runtime dep; its NFC/NFD implementation is the reference for Unicode canonical equivalence. Trust delegated. |
| Golden file → cross-platform diff | The single committed `testdata/golden/normalisation.json` is the v1.x stability contract; any platform-dependent divergence fails CI on that platform. |

## STRIDE Threat Register

| Threat ID | Category | Component | Disposition | Mitigation Plan |
|-----------|----------|-----------|-------------|-----------------|
| T-01-06-01 | Denial of Service | Pathological Unicode input slowing Normalise | mitigate | Performance budget < 200 ns for ASCII ≤ 50 chars and benchmark coverage detects regressions. Long Unicode inputs are O(n) per x/text norm. No quadratic paths. |
| T-01-06-02 | Tampering | Cross-platform Unicode divergence (line-ending, BOM, normalised form drift) | mitigate | TestGolden_Normalisation runs per-platform; D-12's canonical form (two-space indent, single LF, no BOM) is byte-asserted; canonicalMarshal unit tests (plan 01-04) cover the format-helper invariants. |
| T-01-06-03 | Information Disclosure | Mixed-script preservation regression (e.g. Cyrillic stripped accidentally) | mitigate | Mixed-script golden entries pin Cyrillic, Arabic, CJK preservation behaviour; regression would visibly fail the golden diff. |
| T-01-06-04 | Spoofing | StripDiacritics=true producing inconsistent output for visually-similar inputs (e.g. café vs cafe vs `cafe`) | mitigate | Golden entries pin the exact expected output for café/Müller/naïve/résumé under both StripDiacritics ON and OFF; regression fails the golden diff. |
| T-01-06-05 | Tampering | Normalise panic on malformed UTF-8 | mitigate | FuzzNormalise asserts panic-free across arbitrary bytes including invalid UTF-8 sequences and lone surrogates (per `.claude/skills/algorithm-correctness-standards/SKILL.md` Unicode Handling). Hand-crafted seed corpus covers known pitfall patterns. |
| T-01-06-06 | Repudiation | Output silently containing transcendental float artifact | accept | Normalise uses no floats; DET-06 satisfied trivially. Determinism-reviewer verifies at execution time. |
| T-01-06-07 | Tampering | A future change introducing `init()`-time table build, breaking determinism on Windows where init order can differ | mitigate | The file's structure (no init function, tables built per-call from opts) is asserted by `! grep -E '^func init\(\)' normalise.go` in the verify command; determinism-reviewer checks at execution time. |
</threat_model>

<verification>
1. `go build ./...` and `go vet ./...` exit 0.
2. `make check` exits 0.
3. `go test -race -shuffle=on -count=1 -run 'TestNormalise_|TestProp_Normalise_|TestGolden_Normalisation' ./...` exits 0.
4. `go test -fuzz=FuzzNormalise -fuzztime=5s -run=^$ ./...` exits 0.
5. `go test -bench=BenchmarkNormalise -benchmem -count=1 -run=^$ ./...` exits 0 with documented allocation counts.
6. `testdata/golden/normalisation.json` exists, parses as JSON, contains
   20-40 entries (per D-11), ends with single LF, has no BOM.
7. `make verify-determinism` exits 0.
8. `make verify-deps-allowlist` exits 0 (still only x/text in the runtime
   allowlist).
</verification>

<success_criteria>
- `Normalise` implements the spec §9 pipeline with ASCII fast path and
  Unicode (x/text NFC/NFD + diacritic strip) path.
- `NormalisationOptions` struct with D-03 defaults.
- 20-40 golden entries pin the v1.x behaviour across 5 platforms.
- Property tests verify idempotence, length-bound, and panic-free.
- FuzzNormalise covers malformed UTF-8 panic-free property.
- Benchmark coverage exists for ASCII-short / ASCII-long / Unicode paths.
- No init-time work; no map iteration on output; no transcendental floats.
- `make check` is green.
</success_criteria>

<output>
After completion, create
`.planning/phases/01-foundation-infrastructure/01-06-primitives-normalise-SUMMARY.md`
recording:
  - The final NormalisationOptions field set (per api-ergonomics-reviewer
    at execution time; record any drift from the default).
  - The exact entry count and category breakdown in
    `testdata/golden/normalisation.json`.
  - Benchmark numbers for the 7 benchmarks (allocations + ns/op) — to be
    added to bench.txt by the user's `make bench` workstation run.
  - Whether any property test caught a real bug during development
    (informational).
  - Cross-platform divergences observed when CI runs the full matrix (none
    expected; recorded for completeness).
</output>
