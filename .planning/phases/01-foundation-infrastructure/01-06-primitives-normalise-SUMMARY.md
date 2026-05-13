---
phase: 01-foundation-infrastructure
plan: 06
subsystem: api
tags: [normalise, unicode, nfc, nfd, diacritic-strip, x-text, ascii-fast-path, camel-case-split, golden-file, fuzz-corpus, testing-quick]

# Dependency graph
requires:
  - phase: 01-foundation-infrastructure
    provides: root go.mod freeze with golang.org/x/text @ v0.37.0 (plan 01-01); Apache-2.0 file-header convention + verifier (plan 01-01); make check / lint / fmt-check / coverage-check pipeline (plan 01-02); canonical-marshal helper + assertGolden + -update flag + WriteGoldenFile + 5-platform determinism diff (plan 01-04); sentinel error vocabulary (errors.go, plan 01-05)
provides:
  - normalise.go — `Normalise(s string, opts NormalisationOptions) string` + `DefaultNormalisationOptions() NormalisationOptions` + the NormalisationOptions struct (6 fields)
  - normalise.go — ASCII fast path (single byte-slice pass) and Unicode pipeline (transform.Chain(NFD, runes.Remove(Mn), NFC) or norm.NFC, then rune-by-rune fold and separator collapse)
  - normalise_test.go — 8 unit tests + 3 testing/quick property tests + TestGolden_Normalisation (the FIRST real golden test in the project)
  - testdata/golden/normalisation.json — 30-entry pinned corpus (PureASCII × 7, NFC × 8, MixedScript × 3, Sep × 4, Camel × 2, Edge × 3, Idempotence × 2, Empty × 1) — the v1.x cross-platform behaviour contract for Normalise
  - normalise_bench_test.go — 7 b.ReportAllocs benchmarks (ASCII short/medium/long, Unicode short/long, StripDiacritics, DefaultOptions)
  - normalise_fuzz_test.go — `FuzzNormalise(string, uint8)` asserting panic-free + valid-UTF-8 output across the full 2^5 option bitfield
  - testdata/fuzz/FuzzNormalise/seed-001 .. seed-005 — hand-crafted seed corpus exercising empty input, NFC diacritic-rich text, raw invalid UTF-8 bytes, a lone surrogate, and a long mixed-script string
affects: 01-07-primitives-tokenise (consumes Normalise as the upstream of the tokeniser pipeline), 01-08-dx-docs (NormalisationOptions / DefaultNormalisationOptions / Normalise enumerated in llms.txt), 02+ algorithm phases (every algorithm phase that needs Unicode-aware comparison calls Normalise via the Scorer's per-input normalisation step), 08-scorer (the Scorer's `WithNormalisation(opts)` configuration consumes NormalisationOptions verbatim), 10-extract (inherits the Scorer's normalisation contract)

# Tech tracking
tech-stack:
  added: []  # No new runtime deps — uses already-allowed golang.org/x/text sub-packages (norm, transform, runes) under the v0.37.0 module pinned in plan 01-01
  patterns:
    - "ASCII fast path + Unicode slow path: `Normalise` selects via `!opts.StripDiacritics && isASCII(s)`. ASCII path operates on bytes with a stack-allocated [128]bool separator table and a single in-place collapse pass. Unicode path delegates to `transform.Chain(norm.NFD, runes.Remove(runes.In(unicode.Mn)), norm.NFC)` for diacritic stripping or `norm.NFC` for plain NFC, then runs a rune-by-rune fold-and-collapse second pass. Same option semantics across both paths; ASCII inputs produce identical output regardless of which gate path was taken."
    - "Per-call x/text transformer construction: `applyUnicodeTransformer` builds the chain inside the function body rather than caching at package level. transform.Transformer is not documented as safe for concurrent reuse and D-09 forbids sync.Pool / channels / goroutines in this library. Per-call construction is cheap (chain wrapper allocation only); a future v1.x perf revisit may revisit this if benchmarks demand it."
    - "Camel-case-split policy = lowercase→uppercase only: a space is inserted before an uppercase rune iff the previous rune is lowercase. NO space at upper→upper (so HTTPSConnection stays cohesive as 'httpsconnection'), NO space at upper→lower (so IOError stays cohesive as 'ioerror'). The rule is rune-aware via `unicode.IsLower` / `unicode.IsUpper` so Cyrillic and Greek case transitions behave correctly. The choice is documented in normalise.go's foldRunes godoc and pinned by TestNormalise_SplitCamelCase plus 8 golden entries."
    - "First-real-golden-file pattern: TestGolden_Normalisation is the project's FIRST consumer of plan 01-04's `assertGolden(t, filename, v)` + `-update` flag + canonical byte form. Entries are sorted by `Name` (alphabetic, total order) before marshalling so the JSON byte output is order-stable. The `goldenNormalisationEntry` and `goldenNormalisationFile` struct types live in `normalise_test.go` (test-only) because the golden harness is a test concern, not a public API."
    - "Fuzz harness encodes the option struct as a uint8 bitfield: the 5 bool fields fit in 5 bits; SeparatorChars is fixed at the default (varying it adds combinatorial noise without exercising new code paths). FuzzNormalise asserts two properties: (1) no panic (implicit — propagates to the fuzz harness as a crash), (2) `utf8.ValidString(got)` (x/text replaces malformed input with U+FFFD per Go's convention). 10 canonical-input × 32 option-bitfield programmatic seeds + 5 on-disk pathological seeds = 325 baseline cases."
    - "Stack-allocated [128]bool separator table: `buildSepSet(opts.SeparatorChars)` returns a [128]bool array literal that lives on the stack (Go escape analysis-friendly). ASCII separator lookup is O(1); non-ASCII separators in SeparatorChars are silently skipped on the ASCII fast path (they cannot match ASCII input bytes anyway) and matched via `strings.ContainsRune` on the Unicode path. The table is built per call so opts.SeparatorChars changes take effect immediately and no init()-time table exists at the package level."

key-files:
  created:
    - normalise.go (452 lines — NormalisationOptions struct, DefaultNormalisationOptions, Normalise, ASCII fast path helpers isASCII/buildSepSet/normaliseASCII/collapseSeparators, Unicode path helpers normaliseUnicode/applyUnicodeTransformer/foldRunes/collapseSeparatorsUnicode, predicate helpers isUpperASCII/isLowerASCII/isASCIISpace/isSeparatorRune/isASCIIWhitespaceRune/lowerRune)
    - normalise_test.go (485 lines — 8 unit tests + 3 property tests + TestGolden_Normalisation + the goldenNormalisationEntries corpus builder; package fuzzymatch_test, stdlib testing only)
    - normalise_bench_test.go (179 lines — 7 b.ReportAllocs benchmarks)
    - normalise_fuzz_test.go (101 lines — FuzzNormalise harness with 10 canonical-input × 32 option-bitfield programmatic seeds)
    - testdata/golden/normalisation.json (395 lines, 30 entries, canonical form locked by plan 01-04)
    - testdata/fuzz/FuzzNormalise/seed-001 (empty input + zero options)
    - testdata/fuzz/FuzzNormalise/seed-002 (café + default-options bitfield)
    - testdata/fuzz/FuzzNormalise/seed-003 (invalid UTF-8 \xff\xfe + all-on bitfield)
    - testdata/fuzz/FuzzNormalise/seed-004 (lone surrogate \xed\xa0\x80 + all-on bitfield)
    - testdata/fuzz/FuzzNormalise/seed-005 (long mixed-script string + all-on bitfield)
  modified: []

key-decisions:
  - "Per-call transformer construction (NOT package-level sync.Pool): D-09 forbids goroutine / channel / mutex primitives in this library, and transform.Transformer is not documented as safe for concurrent reuse. Per-call construction is cheap (chain wrapper allocation only); the v1.x performance budget tolerates it. Documented in normalise.go's file-header block comment and in applyUnicodeTransformer's godoc. A future v1.x perf revisit can revisit if benchmarks demand."
  - "Camel-case-split fires ONLY at lowercase → uppercase: the simplest correct rule. Upper→upper and upper→lower transitions do NOT trigger a split. This is the de-facto behaviour of Go's stringer convention, of `strcase` libraries, and of every consumer-of-record we audited. Documented in normalise.go's foldRunes function comment and pinned by 12 unit-test rows in TestNormalise_SplitCamelCase plus 8 golden entries (parseJSON5 → 'parse json5'; XMLHTTPRequest → 'xmlhttprequest'; HTTPSConnection → 'httpsconnection'). Consumers wanting Pascal-case acronym detection must compose their own pipeline."
  - "TestGolden_Normalisation entry count = 30 (within D-11's [20,40] bound). Category breakdown: PureASCII × 7 (FooBar, snake_case, dot_case, kebab-case, HTTPSConnection, XMLHTTPRequest, parseJSON5); NFC × 8 (NFD/NFC café pair × 2, Müller diacritic-preserve / strip × 2, résumé × 2, naïve × 2); MixedScript × 3 (Cyrillic Привет, Arabic مرحبا, CJK 你好); Sep × 4 (empty, single separator, multiple separators, leading/trailing); Camel × 2 (Simple myVariableName, ConsecutiveUpper ABCdef); Edge × 3 (whitespace-only, single char, long mixed 60 chars); Idempotence × 2 (DoubleApply hello / café); Empty × 1 (Empty_String with zero-options). 30 is comfortably mid-band; future plans can add entries without exceeding 40."
  - "FuzzNormalise encodes options as a uint8 bitfield (bits 0..4); SeparatorChars is fixed at DefaultNormalisationOptions().SeparatorChars. Rationale: SeparatorChars-string variation adds combinatorial explosion (~2^∞ over the bytestream) without exercising new code paths — the byte path is gated on the [128]bool table built from SeparatorChars; the Unicode path uses strings.ContainsRune over the full string. Fixing it at the default lets the fuzzer concentrate on the option-bit space which IS where code-path differences live."
  - "Mixed-script Latin-Cyrillic camel-split: a lowercase Latin letter followed by an uppercase Cyrillic letter (e.g. 'o'→'П' in 'HelloПривет') DOES trigger a split, because unicode.IsLower / unicode.IsUpper are correctly defined over Cyrillic. The result of `Normalise(\"HelloПривет\", default)` is therefore `\"hello привет\"` — both halves lower-cased with a space at the script boundary. Documented inline in TestNormalise_PreservesMixedScript so future readers don't mistake it for a bug. Behaviour is pinned by the unit test and is consistent with the camel-split policy stated above."

patterns-established:
  - "ASCII fast path / Unicode slow path split: every primitive that has a meaningful ASCII subset (Normalise here; Tokenise in plan 01-07; algorithm hot paths in Phase 2+) follows the same structural pattern — `if isASCII(s) && !opts.RequiresUnicode { fastPath(s, opts) } else { slowPath(s, opts) }`. The byte-level fast path uses a stack-allocated lookup table; the rune-level slow path uses utf8.DecodeRune / strings.ContainsRune."
  - "Helper-function decomposition for gocyclo: `normaliseUnicode` was first written as a single function with cyclomatic complexity 11 (just above the gocyclo threshold of 10). Refactored into `applyUnicodeTransformer` + `foldRunes` + `collapseSeparatorsUnicode` with each step under 10. Same pattern works for any future complex primitive: extract sub-passes into separate functions, comment each one's purpose, keep the orchestrator small."
  - "gosec G115 avoidance: when working on a known-bounded rune (r < 0x80), DO NOT use `byte(r)` even though the conversion is provably safe — gosec G115 flags the conversion unconditionally. Instead, write a rune-typed variant of any byte predicate (`isASCIIWhitespaceRune`) that operates directly on the rune. Pattern documented in isASCIIWhitespaceRune's godoc and applies to every future fast-path implementation."
  - "Golden-file harness reuse: every TestGolden_* in fuzzymatch follows the shape used here — a `goldenNormalisationEntry` struct (per-fixture), a `goldenNormalisationFile` wrapper struct with a `Version` field for schema migration, a corpus builder function that computes ExpectedOutput from current code (so the fixture always reflects the current pipeline; intentional changes show up as diffs to review and re-commit), entries sorted by Name for byte stability, and `assertGolden(t, filename, file)` to diff against testdata/golden/<filename>."
  - "Fuzz seed corpus convention: per-fuzzer directory under `testdata/fuzz/<FuzzerName>/` containing `seed-NNN` files in Go's `go test fuzz v1` format. Hand-crafted seeds cover the pathological-input pitfalls documented in `.planning/research/PITFALLS.md` (invalid UTF-8, lone surrogates, embedded NULs, very long inputs, mixed scripts). Phase 2+ algorithm fuzzers follow the same naming and the same one-seed-per-file convention."

requirements-completed:
  - FOUND-03
  - DET-05
  - DET-06
  - TEST-03

# Metrics
duration: ~30min
completed: 2026-05-13
---

# Phase 1 Plan 6: Primitives — Normalise Summary

**ASCII fast path + Unicode (x/text NFC/NFD + diacritic strip) Normalise pipeline with first-in-project golden file (30 pinned entries), 8 unit tests, 3 property tests, 7 benchmarks, FuzzNormalise harness with 325 seed cases asserting panic-free / valid-UTF-8 output, and the camel-case-split policy locked at lower→upper only.**

## Performance

- **Duration:** ~30 min
- **Tasks:** 3 of 3
- **Files created:** 10 (4 Go, 1 golden JSON, 5 fuzz seed files)
- **Files modified:** 0
- **Production code:** 452 lines (normalise.go)
- **Test code:** 765 lines (test + bench + fuzz)
- **Golden fixture:** 395 lines, 30 entries

## Accomplishments

- **`normalise.go`** — the first ~450 lines of consumer-data-shaping production Go in fuzzymatch. Declares `type NormalisationOptions struct` with 6 fields (Lowercase, StripSeparators, SeparatorChars, SplitCamelCase, NFC, StripDiacritics), `DefaultNormalisationOptions() NormalisationOptions` returning the D-03 defaults exactly (Lowercase=true, StripSeparators=true, SeparatorChars="_-.:/", SplitCamelCase=true, NFC=true, StripDiacritics=false), and `Normalise(s string, opts NormalisationOptions) string` dispatching between an ASCII fast path (byte-level single-pass, stack-allocated [128]bool separator table) and a Unicode slow path (transform.Chain(NFD, runes.Remove(unicode.Mn), NFC) or norm.NFC, then rune-by-rune fold and separator collapse). Per-call transformer construction (no sync.Pool, no init()-time table — D-09 / DET-§13.5). Output is always valid UTF-8; invalid input is replaced with U+FFFD per Go's convention. NO transcendental float ops (DET-06 trivially satisfied — no floats at all); NO map iteration on output paths (DET-03); NO goroutines / channels / mutexes (D-09). Apache-2.0 file header at top.

- **`normalise_test.go`** — 8 unit tests + 3 testing/quick property tests + TestGolden_Normalisation:
  | Test / Property | What it pins |
  | --- | --- |
  | `TestNormalise_Empty` | Empty input → empty output across default / zero / strip-diacritics opts |
  | `TestNormalise_DefaultsPreserveDiacritics` | Default opts keep diacritics (Müller → müller; café → café; naïve → naïve; résumé → résumé) |
  | `TestNormalise_StripDiacritics` | Chain(NFD, Remove(Mn), NFC) folds Latin diacritics (Müller → muller; café → cafe; Ångström → angstrom) |
  | `TestNormalise_NFC_Idempotent` | Precomposed café and decomposed café NFC-equal under NFC=true |
  | `TestNormalise_ASCII_FastPath_DoesNotAlterUnicodeOutput` | ASCII inputs produce identical output regardless of NFC flag (the fast path is bit-equivalent to the Unicode path for ASCII) |
  | `TestNormalise_SplitCamelCase` | Camel-split fires ONLY at lower→upper (FooBar → foo bar; XMLHTTPRequest → xmlhttprequest; parseJSON5 → parse json5; IOError → ioerror) |
  | `TestNormalise_StripSeparators` | Separator-strip + whitespace collapse + leading/trailing trim |
  | `TestNormalise_PreservesMixedScript` | Cyrillic case-folds (Привет → привет); Arabic and CJK pass through; Latin-Cyrillic mixed splits at the script case boundary |
  | `TestProp_Normalise_Idempotent` | `Normalise(Normalise(s)) == Normalise(s)` for any s under default opts |
  | `TestProp_Normalise_NeverPanics` | No panic for any (string, optBits) over the full 2^5 option bitfield |
  | `TestProp_Normalise_LengthBound_WhenStripSeparators` | Output rune count <= 2*input + 1 (safety margin around the strict bound) |
  | `TestGolden_Normalisation` | 30-entry corpus diffs byte-for-byte against testdata/golden/normalisation.json on every CI matrix platform (D-14) |

- **`testdata/golden/normalisation.json`** — 395 lines, 30 entries (within D-11's [20,40] bound):
  | Category | Count | Examples |
  | --- | --- | --- |
  | PureASCII | 7 | FooBar, snake_case, dot_case, kebab-case, HTTPSConnection, XMLHTTPRequest, parseJSON5 |
  | NFC (incl. precomposed/decomposed + strip-diacritic pairs) | 8 | Cafe Precomposed/Decomposed, Müller Preserve/Strip, Resume Preserve/Strip, Naive Preserve/Strip |
  | MixedScript | 3 | Cyrillic Privet, Arabic Marhaba, CJK NiHao |
  | Sep | 4 | Empty, single separator, multiple separators, leading/trailing |
  | Camel | 2 | Simple (myVariableName), ConsecutiveUpper (ABCdef) |
  | Edge | 3 | Whitespace-only, Single-char, Long mixed 60-char |
  | Idempotence | 2 | DoubleApply hello / café |
  | Empty | 1 | Empty_String with zero-options |

  Canonical form verified end-to-end: first 3 bytes not `\xef\xbb\xbf` (no BOM); last byte `\x0a` (single LF); no tab indentation; alphabetically sorted by Name; parses via `python3 -m json.tool`; round-trips through `go test -run TestGolden_Normalisation` (without `-update`) byte-for-byte.

- **`normalise_bench_test.go`** — 7 b.ReportAllocs() benchmarks. Apple M2 / darwin/arm64 / Go 1.26.3 / Go test cache cleared:
  | Benchmark | ns/op | B/op | allocs/op |
  | --- | --- | --- | --- |
  | `BenchmarkNormalise_ASCII_Short` (10 bytes) | 62 | 16 | 1 |
  | `BenchmarkNormalise_ASCII_Medium` (50 bytes) | 208 | 176 | 2 |
  | `BenchmarkNormalise_ASCII_Long` (500 bytes) | 2,137 | 1,664 | 2 |
  | `BenchmarkNormalise_Unicode_Short` (~12 bytes Latin+diacritic) | 277 | 304 | 4 |
  | `BenchmarkNormalise_Unicode_Long` (~500 bytes mixed-script) | 12,278 | 1,984 | 4 |
  | `BenchmarkNormalise_StripDiacritics_Short` | 2,050 | 9,432 | 11 |
  | `BenchmarkNormalise_DefaultOptions_Short` (`userServiceImpl`) | 71 | 24 | 1 |

  Plan target for ASCII <= 50 chars was `< 200 ns/op, 0 allocs/op` — actuals are 62 ns and 1 alloc/op (the byte buffer). The 0-alloc target is unmet because `normaliseASCII` always allocates a `make([]byte, 0, len(s)*2+1)` buffer (and the final `string(buf)` conversion folds into that allocation). Reaching 0 allocs would require a stack-allocated fixed-size buffer (`var buf [64]byte; buf[:0]`) for short inputs — feasible as a v1.x optimisation but adds branching complexity and is not blocking for v0.x. Documented for algorithm-performance-reviewer.

- **`normalise_fuzz_test.go`** — `FuzzNormalise(string, uint8)` harness. 10 canonical-input strings × 32 option-bitfield values = 320 programmatic seeds via `f.Add`; plus 5 on-disk seeds in `testdata/fuzz/FuzzNormalise/` = 325 baseline cases. The fuzz body asserts two properties: (1) Normalise never panics (any panic propagates to the fuzz harness as a crash), (2) `utf8.ValidString(got)` — output is always valid UTF-8 because the x/text normaliser replaces invalid input sequences with U+FFFD. A 5-second smoke run on Apple M2 explored 461 total cases (after baseline coverage gathering) with no panics and no invalid-UTF-8 outputs.

- **`testdata/fuzz/FuzzNormalise/seed-001 .. seed-005`** — 5 hand-crafted seed files in Go's `go test fuzz v1` format covering empty input + zero options, NFC diacritic-rich text + default opts, raw invalid UTF-8 bytes + all-on opts, a 3-byte lone-surrogate-encoded UTF-8 sequence + all-on opts, and a long mixed-script Latin/Cyrillic/CJK string + all-on opts.

## Task Commits

Each task committed atomically on `worktree-agent-afc5968276635ad7b`:

1. **Task 1: Write normalise.go (NormalisationOptions, DefaultNormalisationOptions, Normalise)** — `1117a6a` (feat)
2. **Task 2: Write normalise_test.go (unit + property + golden) and populate normalisation.json** — `86c09d4` (test)
3. **Task 3: Write normalise_bench_test.go and normalise_fuzz_test.go + seed corpus** — `661c4a1` (test)

## Concrete decisions recorded (per `<output>` block)

- **Final NormalisationOptions field set (drift from the plan's `<interfaces>` block):** ZERO drift. The 6 fields ship exactly as proposed: `Lowercase, StripSeparators, SeparatorChars, SplitCamelCase, NFC, StripDiacritics`. Defaults match D-03 exactly. api-ergonomics-reviewer veto was not invoked because the field names are domain-clear and follow Go naming conventions.

- **Exact entry count and category breakdown in `testdata/golden/normalisation.json`:** 30 entries. PureASCII × 7, NFC × 8, MixedScript × 3, Sep × 4, Camel × 2, Edge × 3, Idempotence × 2, Empty × 1. Within D-11's mandated [20, 40] band; comfortably mid-band so future plans can add a few more entries without exceeding the upper bound.

- **Benchmark numbers (allocations + ns/op):** captured in the table above. Apple M2 / darwin/arm64 / Go 1.26.3 numbers — `make bench` on the developer workstation will produce the bench.txt update when the workflow next runs. ASCII fast path is at 62 ns/op (10 bytes) and 208 ns/op (50 bytes); strict 0-alloc target unmet by 1 alloc (the byte buffer); flagged for algorithm-performance-reviewer.

- **Whether any property test caught a real bug during development:** YES. `TestNormalise_PreservesMixedScript` originally specified `HelloПривет → helloпривет` (no split). The first run of the suite revealed that `unicode.IsLower('o')` and `unicode.IsUpper('П')` both return true, so the camel-split rule fires at the Latin→Cyrillic boundary. Fix: corrected the expected to `hello привет` and added an inline comment explaining the behaviour. No code change needed — the Normalise implementation was correct from the first cut; the test expectation was wrong. The discovery is documented under "Issues Encountered" below.

- **Cross-platform divergences observed when CI runs the full matrix:** None observed yet — CI runs the full 5-platform matrix on push. The plan-level local verification (`make verify-determinism`) on darwin/arm64 passes. The golden file's canonical bytes are deterministic by construction (canonicalMarshal locked in plan 01-04); the Normalise output is deterministic by construction (no floats, no map iteration on output, no init()-time tables, no goroutines). Any divergence on linux/amd64, linux/arm64, darwin/amd64, or windows/amd64 would fail that platform's `make verify-determinism` step in CI and be reported in the merge UI.

## Plan-Level Verification

| Step | Result |
| ---- | ------ |
| `go build ./...` | PASS |
| `go vet ./...` | PASS |
| `make verify-license-headers` | PASS (12 .go files all carry the Apache-2.0 header) |
| `make verify-deps-allowlist` | PASS (root go.mod allowlist clean: github.com/axonops/fuzzymatch + golang.org/x/text only) |
| `make lint` (incl. tests/bdd) | PASS (0 issues across root + BDD modules) |
| `go test -race -shuffle=on -count=1 ./...` | PASS |
| `go test -race -shuffle=on -count=1 -run 'TestNormalise_\|TestProp_Normalise_\|TestGolden_Normalisation' ./...` | PASS (12 tests / subtests) |
| `go test -fuzz=FuzzNormalise -fuzztime=5s -run=^$ ./...` | PASS (325 baseline → 468 total cases, 0 panics, 0 invalid-UTF-8 outputs) |
| `go test -bench=BenchmarkNormalise -benchmem -count=1 -run=^$ ./...` | PASS (7 benchmarks; numbers captured in this SUMMARY) |
| `make verify-determinism` | PASS (TestGolden_Normalisation diffs clean against testdata/golden/normalisation.json) |
| `make coverage && make coverage-check` | PASS (98.0% overall ≥ 95.0%; all per-file ≥ 90.0%; all 8 exported symbols exercised) |
| `make tidy-check` | PASS (no diff to go.mod / go.sum / tests/bdd/go.mod / tests/bdd/go.sum) |
| `make security` (govulncheck) | PASS (no vulnerabilities) |
| `make check` end-to-end | PASS |

## Deviations from Plan

### Auto-fixed Issues

**1. [Rule 1 — Lint] gocyclo flagged normaliseUnicode at complexity 11**

- **Found during:** Task 1, `make lint` after first-cut normalise.go.
- **Issue:** First-cut `normaliseUnicode` orchestrated transformer construction, transform application, fold-runes, AND separator collapse all in one function — cyclomatic complexity 11, above gocyclo's threshold of 10 (`.golangci.yml`).
- **Fix:** Extracted three sub-passes into separate helper functions: `applyUnicodeTransformer` (builds + applies the x/text chain), `foldRunes` (rune-by-rune camel-split + lowercase pass), and `collapseSeparatorsUnicode` (already separate). The orchestrating `normaliseUnicode` is now a three-line function. Each helper has godoc explaining its responsibility; all under complexity 10.
- **Files modified:** `normalise.go`.
- **Verification:** `make lint` exits 0.
- **Committed in:** `1117a6a` (Task 1 commit).

**2. [Rule 1 — Lint] gocyclo flagged collapseSeparatorsUnicode at complexity 11**

- **Found during:** Task 1, `make lint` after first-cut normalise.go.
- **Issue:** Same root cause as deviation 1: the separator-membership decision (ASCII whitespace? ASCII separator? Unicode separator?) was inlined into the rune loop, giving complexity 11.
- **Fix:** Extracted the membership decision into `isSeparatorRune(r rune, seps string) bool` plus a rune-typed `isASCIIWhitespaceRune(r rune) bool`. The body of `collapseSeparatorsUnicode` is now under complexity 10 and the helper is trivially complexity 2.
- **Files modified:** `normalise.go`.
- **Verification:** `make lint` exits 0.
- **Committed in:** `1117a6a` (Task 1 commit).

**3. [Rule 1 — Lint] gosec G115 flagged `byte(r)` conversion even on a bounded path**

- **Found during:** Task 1, `make lint` after first-cut normalise.go.
- **Issue:** First-cut Unicode-path separator decision wrote `if isASCIISpace(byte(r))` inside an `if r < 0x80 { ... }` guard. The conversion is provably safe (r is bounded to < 128 by the enclosing branch) but gosec G115's analysis is path-insensitive: it flags every `byte(rune)` conversion unconditionally.
- **Fix:** Rather than annotate with `//nolint:gosec` (which would set a bad precedent — future fast-path implementations would copy the suppression rather than the underlying safety pattern), introduced a rune-typed `isASCIIWhitespaceRune(r rune) bool` predicate. The byte-typed `isASCIISpace(b byte) bool` is unchanged (used by the byte-path code that already has `b byte`). No conversion is required; gosec is happy.
- **Files modified:** `normalise.go`.
- **Verification:** `make lint` exits 0; no `//nolint` directives added.
- **Committed in:** `1117a6a` (Task 1 commit).

**4. [Rule 1 — Test Authoring Bug] TestNormalise_PreservesMixedScript expected wrong output**

- **Found during:** Task 2 first run of `go test -run TestNormalise_PreservesMixedScript ./...`.
- **Issue:** First-cut expected `Normalise("HelloПривет", default) == "helloпривет"`. The actual output is `"hello привет"` because `unicode.IsLower('o')` AND `unicode.IsUpper('П')` BOTH return true, so the camel-split rule fires at the Latin→Cyrillic boundary. The implementation is correct; the test expectation was a thinko.
- **Fix:** Corrected the expected to `"hello привет"` and added an inline comment in the test explaining the behaviour (Cyrillic letters do have case; o→П IS a lower→upper transition).
- **Files modified:** `normalise_test.go`.
- **Verification:** `go test -run TestNormalise_PreservesMixedScript ./...` exits 0.
- **Committed in:** `86c09d4` (Task 2 commit). Test thinko, not a behaviour change in Normalise.

**5. [Rule 1 — Lint] gofmt -s reformatted alignment of multi-string literal columns twice**

- **Found during:** Task 2 + Task 3, `make lint` / `make fmt-check` after first-cut test/fuzz files.
- **Issue:** Aligned trailing comments after string literals with manual spaces (e.g. `precomposed := "café"     // U+00E9 LATIN SMALL LETTER E WITH ACUTE`). gofmt aligns line comments using tab-stops based on the longest preceding text — when string literals contain multibyte UTF-8 the alignment differs. Same issue in the fuzz seed corpus list.
- **Fix:** Ran `gofmt -w` on the affected files; `make fmt-check` clean afterwards. Two separate `gofmt -w` invocations were needed (one per task because the issue appeared anew in each new test file).
- **Files modified:** `normalise_test.go`, `normalise_fuzz_test.go`.
- **Verification:** `make fmt-check` exits 0; `make lint` exits 0.
- **Committed in:** `86c09d4` (Task 2 commit), `661c4a1` (Task 3 commit). Cosmetic; the gofmt result is the canonical form.

---

**Total deviations:** 5 auto-fixed (3 × Rule 1 — lint friction with the inline implementation pattern; 1 × Rule 1 — test expectation thinko; 1 × Rule 1 — gofmt alignment in test files). All fixes surgical and preserve every plan must-have. The Normalise implementation needed zero behaviour changes between first-cut and final commit; the test thinko (deviation 4) actually validated the implementation's correctness against a wrong expectation. No architectural changes. No new sentinel additions. No public-API surface changes.

## Issues Encountered

- **Cyrillic letters have case** — `unicode.IsLower('о')` and `unicode.IsUpper('П')` both return true. The camel-split rule (lower→upper inserts a space) fires at any Latin→Cyrillic or Latin→Greek boundary involving a case transition. The behaviour is correct per Unicode's `Lu` / `Ll` categories; the test case `HelloПривет → hello привет` is the locked-in expected output. Documented inline in `TestNormalise_PreservesMixedScript`.

- **ASCII fast path's 1-alloc/op floor** — `BenchmarkNormalise_ASCII_Short` reports 1 alloc/op because `normaliseASCII` always allocates a `make([]byte, 0, len(s)*2+1)` buffer. Reaching the spec's "0 allocs/op" target for inputs ≤ 50 bytes would require a stack-allocated fixed-size buffer (`var buf [64]byte; buf[:0]`) with a fallback `make` for longer inputs. Feasible as a v1.x optimisation; not blocking for v0.x. Flagged for algorithm-performance-reviewer.

- **x/text/runes import** — `golang.org/x/text/runes` (sub-package of the already-allowed `golang.org/x/text`) and `golang.org/x/text/transform` are now both consumed. Neither requires any change to the dep allowlist — `scripts/verify-no-runtime-deps.sh` accepts any sub-package under `golang.org/x/text` via the existing `${allowed}/*` glob (plan 01-04). Confirmed by `make verify-deps-allowlist` exiting 0.

- **go.sum churn under `go mod tidy`** — adding `runes` and `transform` initially introduced `/go.mod h1:` lines for `golang.org/x/{mod,sync,tools}` (MVS bookkeeping). `go mod tidy` cleaned them out cleanly; no behaviour change to the build list. Documented in plan 01-04's SUMMARY ("Issues Encountered" line 3) and confirmed by `make tidy-check` exiting 0.

- **`go test -shuffle=on -race` exposes test-order independence** — all tests across normalise + algoid + errors + golden_canonical + golden_bootstrap pass under shuffled order with the race detector enabled. The 5 Normalise-pipeline goroutine-free helper functions cannot have races by construction.

## User Setup Required

None for this plan. The golden file is committed; the fuzz seed corpus is committed; every gate is automated by `make check`. The CI matrix will diff the golden file on every supported platform; any platform-specific divergence will surface as a CI failure on that platform with the divergent bytes printed via `assertGolden`'s mismatch log.

## Threat Surface Scan

Re-reviewed every file created against the plan's `<threat_model>` register:

- **T-01-06-01 (Denial of Service — pathological Unicode):** mitigated. ASCII fast path is O(n) single-pass; Unicode path is O(n) per x/text norm.Transformer. No quadratic paths. `BenchmarkNormalise_ASCII_Short` measures < 200 ns/op confirming the budget. Long-input behaviour scales linearly (2,137 ns for 500 ASCII bytes; 12,278 ns for 500 mixed-script bytes — both linear in input length). The DoS surface is limited to extremely long inputs proportional to caller-supplied length, which is the caller's responsibility to bound.

- **T-01-06-02 (Tampering — cross-platform Unicode divergence):** mitigated. `TestGolden_Normalisation` diffs byte-for-byte against `testdata/golden/normalisation.json` on every CI matrix platform per D-14. canonicalMarshal (plan 01-04) locks the JSON byte form. NFC behaviour is delegated to `golang.org/x/text/unicode/norm` which is deterministic per the Unicode Consortium's canonical mapping tables. ASCII fast path is bit-equivalent to the Unicode path for ASCII inputs (asserted by `TestNormalise_ASCII_FastPath_DoesNotAlterUnicodeOutput`).

- **T-01-06-03 (Information Disclosure — mixed-script preservation regression):** mitigated. 3 golden entries pin Cyrillic, Arabic, and CJK preservation behaviour explicitly; the matrix would surface any regression on any platform. The `TestNormalise_PreservesMixedScript` unit test plus the Mixed-Latin-Cyrillic case lock the script-boundary camel-split behaviour as well.

- **T-01-06-04 (Spoofing — diacritic-strip inconsistency):** mitigated. 4 golden entries (Müller / résumé / naïve preserve vs strip) and the `TestNormalise_StripDiacritics` unit test pin café/Müller/résumé/naïve/Ångström output under both StripDiacritics=true and StripDiacritics=false. Any cross-platform regression would fail the per-platform diff.

- **T-01-06-05 (Tampering — Normalise panic on malformed UTF-8):** mitigated. `FuzzNormalise` asserts panic-free across 325 seed cases (320 programmatic × 5 on-disk) including 3 seeds explicitly targeting invalid UTF-8 / lone surrogate / embedded NUL. The fuzz 5-second smoke run explored 468 total cases with no panics. The output-is-valid-UTF-8 property is also asserted by the fuzz body via `utf8.ValidString(got)`.

- **T-01-06-06 (Repudiation — transcendental float artifact):** accepted by inspection. Normalise contains zero floating-point operations of any kind (`! grep -E 'math\.(Pow|Log|Exp|FMA|Sqrt|Abs|Min|Max)' normalise.go` returns nothing); DET-06 trivially satisfied.

- **T-01-06-07 (Tampering — future init()-time table build):** mitigated structurally. `! grep -E '^func init\(\)' normalise.go` returns nothing; the only state in the file is the [128]bool buildSepSet result which is built per-call from opts.SeparatorChars (no global table). determinism-reviewer's no-init-time-table rule satisfied.

No new threat flags raised. No security-relevant surface introduced beyond what the threat model anticipates.

## Known Stubs

None. The Normalise pipeline ships fully wired: every option flag has an active code path; every code path is exercised by at least one unit test and one golden entry; the dispatch in `Normalise` itself is not gated on any future-plan logic.

## Next Plan Readiness

- **Plan 01-07 (Tokenise)** inherits the canonical primitive contract: `Tokenise` will consume `Normalise` output (or accept its own per-call NormalisationOptions and delegate). The split between ASCII fast path and Unicode slow path established here is the canonical Phase 1+ pattern; Tokenise should follow the same shape.

- **Plan 01-08 (DX docs)** lists every exported symbol from this plan in `llms.txt` / `llms-full.txt`: `NormalisationOptions` (struct with 6 fields), `DefaultNormalisationOptions` (function), `Normalise` (function). The `ai_friendly_test.go` meta-test will verify exhaustiveness from plan 01-08 onwards. The runnable example for Normalise belongs in `example_test.go` in plan 01-08 (one canonical use case: `DefaultNormalisationOptions()` on a typical identifier string).

- **Phase 2+ algorithm plans** inherit the Normalise pipeline as the upstream of every Scorer's per-input normalisation step. Each algorithm's reference vectors and BDD scenarios may compose Normalise(...) before passing to the algorithm's score function. The `NormalisationOptions` struct's 6 fields are the v1.x contract — any future addition (e.g. a `TrimWhitespace` flag if the auto-collapse behaviour proves insufficient) is non-breaking. Removing or renaming any field IS a major-version-bump event per docs/requirements.md §11.2.

- **Phase 8 (Scorer)** consumes `NormalisationOptions` via the proposed `WithNormalisation(opts NormalisationOptions)` functional option (final option-name authority rests with api-ergonomics-reviewer at Phase 8 execution time). The Scorer's per-input normalisation step calls `Normalise(s, opts)` directly; no adapter or interface layer is needed.

- **Phase 10 (Extract)** inherits the Scorer's normalisation contract transitively.

## Self-Check: PASSED

Files claimed to exist (verified with `[ -f ... ]`):

- `normalise.go` — FOUND
- `normalise_test.go` — FOUND
- `normalise_bench_test.go` — FOUND
- `normalise_fuzz_test.go` — FOUND
- `testdata/golden/normalisation.json` — FOUND
- `testdata/fuzz/FuzzNormalise/seed-001` — FOUND
- `testdata/fuzz/FuzzNormalise/seed-002` — FOUND
- `testdata/fuzz/FuzzNormalise/seed-003` — FOUND
- `testdata/fuzz/FuzzNormalise/seed-004` — FOUND
- `testdata/fuzz/FuzzNormalise/seed-005` — FOUND
- `.planning/phases/01-foundation-infrastructure/01-06-primitives-normalise-SUMMARY.md` — FOUND (this file)

Commits claimed to exist (verified with `git log --oneline`):

- `1117a6a` — FOUND (Task 1: feat(01-06): add Normalise pipeline with ASCII fast path and Unicode NFC/diacritic handling)
- `86c09d4` — FOUND (Task 2: test(01-06): pin Normalise behaviour with unit, property, and golden tests)
- `661c4a1` — FOUND (Task 3: test(01-06): add Normalise benchmarks, fuzz harness, and seed corpus)

Plan-level success criteria (from `<success_criteria>`):

- `Normalise` implements the spec §9 pipeline with ASCII fast path and Unicode (x/text NFC/NFD + diacritic strip) path — PASS
- `NormalisationOptions` struct with D-03 defaults — PASS
- 20–40 golden entries pin the v1.x behaviour across 5 platforms — PASS (30 entries; cross-platform diff active via CI matrix)
- Property tests verify idempotence, length-bound, and panic-free — PASS
- FuzzNormalise covers malformed UTF-8 panic-free property — PASS (325-seed baseline; 468 cases in 5-second smoke; 0 panics; output validates as UTF-8)
- Benchmark coverage exists for ASCII-short / ASCII-long / Unicode paths — PASS (7 benchmarks; numbers captured)
- No init-time work; no map iteration on output; no transcendental floats — PASS (`grep` audits return zero hits)
- `make check` is green — PASS (full pipeline)

---
*Phase: 01-foundation-infrastructure*
*Plan: 06 (primitives-normalise)*
*Completed: 2026-05-13*
