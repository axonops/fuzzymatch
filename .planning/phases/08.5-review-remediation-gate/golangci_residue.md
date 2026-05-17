# golangci-lint Residue (Phase 8.5 Plan 15b)

This document records every `//nolint:<linter>` annotation introduced (or kept) during Plan 15b's lint sweep, with a one-line justification AND a flag for whether to revisit at v1.x.

The Plan 18 verification gate reads this file's row count to assess the "Improvement-tier residue" surface.

## Sweep Result

- **Before sweep:** 63 issues across 10 linter classes (gocritic, gocyclo, gofmt, gofumpt, gosec, misspell, staticcheck, errcheck, errorlint, unparam).
- **After sweep:** **0 issues** (`golangci-lint run ./...` exits 0).
- **British English `misspell` sweep:** clean — no American spellings remain in `*.go` files (`grep -rE 'color\b|behavior\b|normalize\b|serialize\b|artifact\b|analyze\b' $(git ls-files '*.go')` returns nothing).
- **Test gates:** `go test -race -count=1 ./...` and `cd tests/bdd && go test ./...` both green; `go vet ./...` clean.

## Residue Table

Every entry below is a `//nolint:<linter>` annotation. Each has a one-line justification AND a revisit-at-v1.x flag. The file feeds Plan 18's verification report under "Improvement-tier residue".

| File | Line | Linter / Rule | Justification | Revisit at v1.x? |
|------|------|---------------|---------------|------------------|
| `algoid.go` | 213 | `gocyclo` | one switch case per AlgoID is intentional (pre-existing — Plan 13) — every new algorithm forces a case branch; this is the canonical Go pattern for typed enums | No — preserved as part of the v1.x stability contract |
| `double_metaphone.go` | 123 | `gocyclo` | `dmSlgCheck` scans for 4 alternative SlavoGermanic patterns; flat early-return shape | No |
| `double_metaphone.go` | 191 | `gocyclo` | `dmPrep` is the canonical Philips 2000 pre-prep scan; per-character branches are inherent to the documented rules | No |
| `double_metaphone.go` | 258 | `gocyclo` | `DoubleMetaphoneKeys` — CC of 275 reflects the canonical Philips 2000 state machine (single switch on current char with ~200 lookahead/lookbehind branches); structurally unavoidable | No — this is the algorithm |
| `double_metaphone.go` | 916 | `gocyclo` | `DoubleMetaphoneScore` — exhaustive primary-vs-secondary tier matching | No |
| `mra.go` | 135 | `gocyclo` | `MRACode` — canonical NBS-943 three-step encoding | No |
| `mra.go` | 219 | `gosec` G602 | false positive — `step2Len > 6` early-return above guarantees indices ≥ 4 and < 64; gosec's bounds-prover doesn't trace the negation | No |
| `mra.go` | 220 | `gosec` G602 | same false positive (step2Len-2) | No |
| `mra.go` | 221 | `gosec` G602 | same false positive (step2Len-1) | No |
| `mra.go` | 266 | `gocyclo` | `MRACompare` — canonical NBS-943 six-step comparison | No |
| `nysiis.go` | 116 | `gocyclo` | `NYSIISCode` — CC of 78 reflects the canonical Taft 1970 branch table (~20 distinct transformation rules) | No |
| `scorer_options.go` | 454 | `gocyclo` | `WithTverskyAlgorithm` — Q2 strict-parameter framework with 4 distinct sentinel returns | No |
| `soundex.go` | 155 | `gocyclo` | `SoundexCode` — Russell-Knuth 6-rule encoder | No |
| `swg.go` | 374 | `gocyclo` | `swgDPRaw` (pre-existing — Plan 3) — SWG three-matrix kernel; extraction would obscure the recurrence | No |
| `swg.go` | 206 | `gocyclo` | `(SWGParams).validate` — per-field invariant blocks (Match/Mismatch/GapOpen/GapExtend) | No |
| `swg.go` | 472 | `gocyclo` | `swgDPRawRunes` (pre-existing — Plan 3) — rune-surface mirror of `swgDPRaw` | No |
| `tversky.go` | 257 | `gocyclo` | `TverskyScore` — Q2 strict-parameter framework | No |
| `tversky.go` | 313 | `gocyclo` | `TverskyScoreRunes` — rune-surface mirror | No |
| `validate.go` | 152 | `gocyclo` | `Validate` — 5 cross-cutting WarnKind classifications | No |
| `scorer_golden_test.go` | 287 | `misspell` | "abreviation" is a deliberate misspelling test fixture (compares correct vs misspelled) | No |
| `scripts/cmd/verify-exported-coverage/main.go` | 115 | `errcheck` | best-effort stderr write; exit code is the canonical signal | No |
| `scripts/cmd/verify-exported-coverage/main.go` | 143 | `gocyclo` | `run` — canonical pipeline shape (stat → walk → load → enumerate → check → emit) | No |
| `scripts/cmd/verify-exported-coverage/main.go` | 154 | `errcheck` | best-effort writer emission | No |
| `scripts/cmd/verify-exported-coverage/main.go` | 194 | `errcheck` | best-effort writer emission | No |
| `scripts/cmd/verify-exported-coverage/main.go` | 204 | `errcheck` | best-effort writer emission | No |
| `scripts/cmd/verify-exported-coverage/main.go` | 235 | `staticcheck` SA1019 | `parser.ParseDir` is deprecated in Go 1.25 — we keep it deliberately because this is internal AST tooling, we do not inspect build tags, and `golang.org/x/tools` is not in the root module's zero-runtime-dep allowlist | Yes — re-evaluate once `golang.org/x/tools` is whitelisted (currently forbidden by spec §1) |
| `scripts/cmd/verify-exported-coverage/main.go` | 322 | `gosec` G204 | `profile` is operator-supplied via the `-profile` flag and is treated as a relative path; command path is the literal "go" | No |
| `scripts/cmd/verify-exported-coverage/main.go` | 388 | `gocyclo` | `collectTestRefs` — AST walk with per-decl-type and per-identifier inspection | No |

Pre-existing `//nolint` annotations from earlier plans (not introduced by Plan 15b but still present in the repo):

| File | Line | Linter | Plan | Note |
|------|------|--------|------|------|
| `scorer_test.go` | 825, 870 | `gocritic` | Plan 11 | DET-06 locked left-to-right additive accumulation pattern |

## Items Fixed Inline (not residue)

Resolved without `//nolint`:

- `gofumpt`/`gofmt`: 25 files reformatted via `gofumpt -w`.
- `gocritic` ifElseChain at `double_metaphone.go:739`: refactored to a switch statement.
- `staticcheck` QF1001 (De Morgan): 7 sites refactored from `!((cond1) || (cond2))` to `(!cond1 && !cond2)` form.
- `staticcheck` SA4004 at `errors_test.go:209`: refactored `for-range/break` → `utf8.DecodeRuneInString`.
- `staticcheck` ST1005 at `swg.go:214, 217`: error strings prefixed with "field " so they start lowercase.
- `staticcheck` S1008 at `scripts/cmd/verify-exported-coverage/main.go:219`: simplified `if !strings.HasSuffix(...) return false; return true` to a single negated return.
- `errorlint` at `scripts/cmd/verify-exported-coverage/main.go:119`: replaced `type assertion on error` with `errors.As`.
- `unparam` at `scorer_options_internal_test.go:130`: dropped the always-zero `i` parameter from `probeScoreFnInvoke`; all 6 call sites updated.
- `unparam` at `soundex.go:283`: renamed `runeAt(string, int) (rune, int)` to `runeSizeAt(string, int) int` since no caller used the rune; 6 call sites updated.

## v1.x Revisit List

Only one entry warrants future re-evaluation:

1. `parser.ParseDir` deprecation (Go 1.25, `scripts/cmd/verify-exported-coverage/main.go:235`) — currently kept because `golang.org/x/tools` is outside the spec's zero-runtime-dep allowlist (§1). Re-evaluate when/if the allowlist is extended.
