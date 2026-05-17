---
status: issues_found
agent: determinism-reviewer
scope: entire codebase (phases 1-8)
reviewed: 2026-05-17T00:00:00+00:00
finding_counts:
  critical: 0
  important: 5
  improvement: 12
  total: 17
---

# Determinism Review — fuzzymatch (Phases 1 through 8)

## Scope

Comprehensive cross-codebase determinism audit of the fuzzymatch root
package as it exists after Phase 8 (composite Scorer) closes. Builds on
`.planning/phases/08-composite-scorer/08-DETERMINISM-REVIEW.md` (which
covered the Phase 8 surface only) and expands to every algorithm, the
dispatch tables, Normalise, Tokenise, q-gram helpers, and the four
golden files. Cross-references the `determinism-standards` skill.

## Method

- Source walk: every production `.go` file in the package root (66 files).
- Grep audit for `init()`, `math.*`, `range map`, `sort.Slice*`,
  `time.Now`, `rand.*`, `os.Getenv`, `runtime.GOOS|GOARCH`, `sync.*`,
  `chan `, `go func`, `slices.Sum`.
- Golden-file inspection for timestamps and non-deterministic content.
- `go test -run TestProp.*Deterministic ./...` — PASS (0.5 s).
- `go test -run TestGolden ./...` — PASS (0.6 s).
- `go test -run TestProp.*(NoNaN|NoInf|Symmetric|Range|Identity) ./...` — PASS (2.9 s).

## Top-Line

**No CRITICAL findings.** No cross-platform byte-identity break has been
introduced anywhere in Phases 1 through 8. The library's determinism
contract — no `init()`, no `math.X` beyond `math.Sqrt`, no map iteration
on output paths, no time/rand/env sniffing, no concurrency primitives,
no parallel reduction — is intact and enforced by passing property +
golden-file tests on the local platform.

The five IMPORTANT findings flag patterns that are CORRECT today but
either (a) deviate from the locked DET-06 contract style without being
wrong, (b) leak documented-as-non-deterministic information through
backdoors that are not yet exercised, or (c) are conditional on the
cross-platform CI matrix running before each release. Twelve
IMPROVEMENT findings are defence-in-depth / style.

The cross-platform CI matrix (linux/{amd64,arm64}, darwin/{amd64,arm64},
windows/amd64) is wired into `.github/workflows/ci.yml` and runs
`make verify-determinism` (which executes `go test -run TestGolden_
./...`) on every PR. The matrix is the load-bearing empirical gate; the
local-only review here cannot certify cross-platform byte-identity by
itself.

---

## Verification Checklist Results

| Check | Result |
|---|---|
| No `init()` doing non-trivial work | PASS — zero `func init()` in the package |
| No map iteration on output paths | PASS — every `range map` reduces to a scalar or feeds an explicitly-sorted slice |
| No `math.Pow/Log/Exp/FMA` | PASS — only `math.Sqrt` (Cosine), `math.IsNaN`/`math.IsInf` (test code) |
| Float reduction left-to-right | PASS — Scorer reduction, Cosine dot-product, weight normalisation all explicit `(x*y)+z` parens |
| No `slices.Sum` / parallel reduction | PASS — none used |
| NaN/Inf/-0 guarded on every public path | PASS — every `/maxLen` and `/m` has an explicit `== 0` guard preceding it |
| Sort-key completeness | MOSTLY PASS — see DET-W1 (Algorithms_Merge uses sort.Slice not SliceStable with single-key Name) |
| Golden file format | PASS — single trailing LF, no BOM, no timestamps in scorer/algorithms/normalisation; only static literal in phonetic |
| Cross-platform CI matrix wired in | PASS — `.github/workflows/ci.yml` runs `make verify-determinism` on 5 platforms |
| AlgoID enum order stability | PASS — `AlgoIDs()` returns a hand-written slice literal; algoid_test.go pins count=23 and stable order |
| Tokenise stable output ordering | PASS — left-to-right rune walk, no map, no goroutines |
| Normalise NFC/NFD selection consistency | PASS — switch in `applyUnicodeTransformer` is total over `(StripDiacritics, NFC)` cross-product |
| Per-algorithm: no hidden RNG / time.Now / env sniffing | PASS — zero occurrences |
| Concurrent-use determinism | PASS — Scorer is immutable after construction; no globals are written after package init |

---

## Findings

### [IMPORTANT] DET-W1: `TestGolden_Algorithms_Merge` uses `sort.Slice` not `sort.SliceStable` with a single-key sort
- **File:** algorithms_golden_test.go:200-202, 222-224, 275-277
- **Phase introduced:** Phase 2 (Wave 3 merge)
- **Issue:** The merge / per-algorithm-staging tests sort entries by `Name` using `sort.Slice` (unstable). The sort key is *only* `Name`. The merge test has a defensive duplicate-name check immediately after the sort, but the per-algorithm staging tests do not. If two entries shared a `Name` (a bug — but not impossible during regen), the on-disk ordering would be non-deterministic across runs, breaking the byte-identity golden.
- **Standard:** `.claude/skills/determinism-standards/SKILL.md` §"Sort Key Completeness" — "When sorting output, use sort.SliceStable AND ensure the key is complete (or document why ties don't occur and assert it in a test)."
- **Action:** Code fix — swap `sort.Slice` for `sort.SliceStable` in algorithms_golden_test.go:200, :222, :275, and every other site in the file. The duplicate-name sanity check in the merge function should be lifted into a shared helper and applied in every staging test too.
- **Rationale:** The current pattern is correct today (no duplicate Names exist by construction), but stable sort defends against a future regression introducing a duplicate. Cost is zero (sort.SliceStable's performance is identical for small N).
- **Suggested fix:**
  ```go
  sort.SliceStable(allEntries, func(i, j int) bool {
      return allEntries[i].Name < allEntries[j].Name
  })
  ```
  Repeat in every `sort.Slice(...by Name...)` call in algorithms_golden_test.go.

### [IMPORTANT] DET-W2: Cosine FMA-fusion remediation is documented but not applied; the `(x*y)+z` dot-product pattern is the LOAD-BEARING cross-platform float-determinism algorithm
- **File:** cosine.go:343
- **Phase introduced:** Phase 5
- **Issue:** The Cosine dot-product reduction is `dot = (float64(qa[k]) * float64(qb[k])) + dot`. Per cosine.go:288-297 own godoc and the explicit warning in scorer.go:55-61, the Go compiler MAY emit a fused multiply-add (FMA) instruction on arm64 for the `(x*y)+z` pattern (per golang/go#17895, parens do NOT defeat FMA fusion). The empirical claim is that the integer-derived `qa[k] * qb[k]` products stay below the ULP threshold where FMA-vs-non-FMA divergence would surface. Today the cross-platform golden gate PASSES on the four-platform matrix run, validating this claim. But the remediation pattern (`dot = float64(float64(qa[k]) * float64(qb[k])) + dot`) is NOT applied — only documented.
- **Standard:** `.claude/skills/determinism-standards/SKILL.md` §"Float Stability" — "No platform-specific intrinsics (math.FMA etc.) — even where they would be faster." The compiler-emitted FMA from a `(x*y)+z` source pattern is the same hazard category.
- **Action:** Discuss-phase needed — option A (status quo, monitor matrix); option B (preemptively apply the double-cast).
- **Rationale:** Latent risk against a future Cosine input where the inner-product magnitude grows to the float64 boundary. The same risk applies to scorer.go:380 (`acc = acc + (entry.weight * score)`) — both sites are the LOAD-BEARING gate. If a future algorithm produces a score whose weighted contribution crosses the ULP threshold, the unmitigated `(weight * score)` would diverge between arm64 (FMA) and amd64 (non-FMA).
- **Suggested fix (Option B, defensive):**
  ```go
  // cosine.go:343
  dot = float64(float64(qa[k]) * float64(qb[k])) + dot
  // scorer.go:380
  acc = acc + float64(float64(entry.weight) * float64(score))
  ```
  Each `float64(...)` cast forces a single rounding step, defeating FMA fusion. Cost: ~1ns per iteration; benign.

### [IMPORTANT] DET-W3: `WithThreshold` does not guard `NaN` — admits non-finite into `*Scorer.threshold` (already CR-01 in 08-REVIEW.md; recorded here from the determinism perspective)
- **File:** scorer_options.go:257-266
- **Phase introduced:** Phase 8
- **Issue:** `WithThreshold(t)` checks `t < 0.0 || t > 1.0`. Both comparisons evaluate to `false` for `NaN`. NaN passes the gate, propagates to `Scorer.threshold`, and `Match` becomes "never match" for every input (IEEE-754: `x >= NaN` is always `false`). The result is deterministic-but-wrong, so the cross-platform golden gate does NOT catch it. But it violates the project-wide NaN/Inf policy.
- **Standard:** `.claude/skills/determinism-standards/SKILL.md` §"NaN / Inf / Negative Zero" — "No algorithm produces NaN, +Inf, or -Inf. Any path that could (e.g. division by zero in normalisation) MUST be guarded with an explicit edge case."
- **Action:** Code fix — already tracked as CR-01 in 08-REVIEW.md. Add `math.IsNaN(t)` to the gate.
- **Rationale:** Defence-in-depth; closes the NaN escape on the Scorer's surface boundary.
- **Suggested fix:**
  ```go
  if math.IsNaN(t) || t < 0.0 || t > 1.0 {
      return ErrInvalidThreshold
  }
  ```

### [IMPORTANT] DET-W4: Scorer reduction trusts dispatched scoreFn return values — no defensive NaN/Inf guard
- **File:** scorer.go:368-382, scorer.go:512-517
- **Phase introduced:** Phase 8
- **Issue:** Both `Score` and `ScoreAll` trust `entry.scoreFn(na, nb)` to return a finite float in `[0, 1]`. All 23 catalogue algorithms satisfy this contract today, verified by per-algorithm `TestProp_<algo>_NoNaN_NoInf` and the Phase-8-level `TestProp_Scorer_NoNaN_NoInf`. The fragility is that a future regression in any algorithm, OR a parameterised custom closure (e.g. `WithSmithWatermanGotohAlgorithm` with pathological `SWGParams` whose godoc warns "nonsense values produce a deterministic-but-meaningless score" but does NOT promise finiteness), could leak NaN into the reduction. NaN poisons `acc` via `acc + (w * NaN) = NaN` and the bug surfaces as "Match always false; ScoreAll returns a map with NaN values" — deterministic, so cross-platform CI doesn't catch it.
- **Standard:** `.claude/skills/determinism-standards/SKILL.md` §"NaN / Inf / Negative Zero" — same as DET-W3 but on the consumer side.
- **Action:** Code fix or skill-clarification — choose between (a) leave as-is (current tests catch every known producer), (b) add a per-iteration `math.IsNaN`/`math.IsInf` check that panics with a programmer-error message, (c) add a `ScoreErr` method that returns the error path. Recommendation: (b) for v1.x; (c) would need api-ergonomics-reviewer sign-off.
- **Rationale:** Latent; not a current bug. Adding the panic guard costs ~1ns per iteration and surfaces algorithm-layer regressions at the earliest possible point.

### [IMPORTANT] DET-W5: `WithoutAlgorithm` in-place compaction `cfg.entries[:0]` aliases the underlying array
- **File:** scorer_options.go:184-199
- **Phase introduced:** Phase 8
- **Issue:** The `filtered := cfg.entries[:0]` pattern reuses the backing array. The algorithm is correct (writes only after reads), but `cfg.entries` retains old-value capacity beyond the new length. If a future maintainer ever returns `cfg.entries` to consumer code, or if a subsequent option capture-by-reference re-reads spare-capacity slots, they'd see stale values from before the removal. For determinism specifically: today the value-typed slice header is never externalised, so no bug surfaces. The fragility is latent.
- **Standard:** No governing skill clause; falls under general "no shared mutable state escapes".
- **Action:** Code fix or accept-with-comment — change to `filtered := make([]scorerEntry, 0, len(cfg.entries))`. Cost: one allocation per `WithoutAlgorithm` at construction time (called rarely, often once).
- **Rationale:** Defence in depth; current pattern is the canonical Go idiom but the godoc comment in `WithoutAlgorithm` (scorer_options.go:184) already documents linear-scan-and-compact semantics, so a fresh allocation matches the documented intent.

---

### [IMPROVEMENT] DET-W6: phonetic-codes.json contains a static `regenerated_at` literal in `_metadata`
- **File:** testdata/golden/phonetic-codes.json:4
- **Phase introduced:** Phase 7
- **Issue:** `"_metadata": { "purpose": "...", "regenerated_at": "2026-05-15T00:00:00+00:00" }`. The string is a static literal (the file is hand-edited, not regenerated from `time.Now()`), so it does NOT break determinism today. But the *appearance* of a timestamp field is a magnet for a future maintainer to convert it to a real-time `time.Now()` value during a regen, which WOULD break the cross-platform byte-identity gate. The other three goldens (algorithms.json, normalisation.json, scorer-default.json) deliberately avoid timestamp fields per 08-DETERMINISM-REVIEW.md DET-07.
- **Standard:** `.claude/skills/determinism-standards/SKILL.md` §"Golden Files" — "Golden files are JSON for human-readability and diff-friendliness. They are regenerated only when the algorithm output legitimately changes (e.g. bug fix)."
- **Action:** Code fix (rename) OR skill-clarification (document the static-literal convention explicitly).
- **Rationale:** Hygiene; matches the scorer-default.json / algorithms.json / normalisation.json pattern. The field is purely informational and the value never actually changes runtime-to-runtime.
- **Suggested fix:** Rename `regenerated_at` to `regenerated_on` or `last_curated` (or drop it entirely; the git log carries the same information). Add a comment in `phonetic_codes_golden_test.go` near the type declaration: "// RegeneratedAt is a STATIC string literal in the golden file, not a runtime-generated timestamp; do NOT replace with time.Now() during regen — that would break the cross-platform byte-identity gate."

### [IMPROVEMENT] DET-W7: scorer_signature is a static literal embedded in test code; drift on DefaultScorer composition change is silent
- **File:** scorer_golden_test.go (search for "DefaultScorer-2026-05-16")
- **Phase introduced:** Phase 8
- **Issue:** Already DET-07 in the Phase 8 standalone review. The string `"DefaultScorer-2026-05-16"` is hard-coded with a construction date. Any future change to DefaultScorer's six-algorithm composition + 0.85 threshold requires hand-updating this constant. Forgetting to update doesn't break determinism (byte-identity still passes if the string is unchanged) but it does mean the metadata lies about which composition the golden represents.
- **Standard:** No governing skill clause; falls under general "golden-file curation discipline".
- **Action:** Skill clarification — add a release-prep checklist item to `docs/scorer.md` or the docs/extending.md release runbook: "If DefaultScorer composition changes, bump scorer_signature in scorer_golden_test.go."
- **Rationale:** Process-level; no code change.

### [IMPROVEMENT] DET-W8: Strcmp95 similar-character table is `[...]struct{...}` literal — good, but the linear scan iteration order is implicit
- **File:** strcmp95.go:133-173 (the table), strcmp95.go:198-206 (the lookup)
- **Phase introduced:** Phase 3 (likely; the file's Phase introduction not explicitly stated but pre-Phase 5)
- **Issue:** The 36-entry similar-character table is declared via `var ... = [...]struct{...}{...}` literal — IDEAL pattern per the no-init rule. The `strcmp95SimilarLookup` function does a linear scan; the iteration order is the declared order of the literal. This is deterministic by construction. The minor friction is that the determinism contract relies implicitly on Go's spec guarantee that array literals iterate in declared order — there's no in-code assertion of this.
- **Standard:** `.claude/skills/determinism-standards/SKILL.md` §"Init() and Package Loading" — already SATISFIED. This is a hygiene note.
- **Action:** No code change; document the dependency. Could be added in commit message or a comment near the `for i := range strcmp95SimilarChars` loop.
- **Rationale:** Defensive documentation only.

### [IMPROVEMENT] DET-W9: `var _ = func() bool { ... }()` package-level dispatch registrations are init() in disguise
- **File:** dispatch_*.go (23 files, one per algorithm)
- **Phase introduced:** Phase 2 onward
- **Issue:** Every dispatch_X.go uses the `var _ = func() bool { dispatch[AlgoX] = X; return true }()` idiom. This runs at package load time, same as `init()`. The skill's "no init() doing non-trivial work" rule is intended to flag non-trivial init — the dispatch registration is trivial (single assignment) and writes to a UNIQUE slot per file, so order between dispatch_X.go files is irrelevant. This is the documented "init-alternative" pattern (see dispatch_double_metaphone.go file header and others). No bug.
- **Standard:** `.claude/skills/determinism-standards/SKILL.md` §"Init() and Package Loading" — "The library has no init() functions doing non-trivial work. Tables that require initialisation ... are declared via `var x = ...` literal expressions, not built in `init()`."
- **Action:** Skill clarification — the rule should explicitly carve out the `var _ = func() bool {...}()` idiom as permitted for trivial dispatch-table population (one assignment, no map iteration, no cross-file dependency). OR change the idiom to `var dispatchCosine = registerDispatch(AlgoCosine, func(a, b string) float64 {...})` returning the input value — same effect, more obviously a `var` literal.
- **Rationale:** The pattern is correct and idiomatic; the skill text just doesn't explicitly call out the exception, creating a documentation gap that a strict reading would flag as a near-violation.

### [IMPROVEMENT] DET-W10: Monge-Elkan reduction uses `+=` while Scorer uses explicit `acc = acc + (x * y)` — consistency gap
- **File:** monge_elkan.go:426 (`sumOfMax += maxSim`)
- **Phase introduced:** Phase 6
- **Issue:** Scorer.Score (scorer.go:380) uses `acc = acc + (entry.weight * score)` — the locked DET-06 pattern. Monge-Elkan's inner reduction uses `sumOfMax += maxSim`. The two are observationally equivalent (compound assignment desugars to `x = x + y`); both have no FMA hazard because there's no multiplication-then-addition fused in this reduction (multiplications happen inside `innerFn`, results land in `s`/`maxSim`, then a pure addition accumulates). No bug.
- **Standard:** `.claude/skills/determinism-standards/SKILL.md` §"Float Stability" — implicit consistency expectation; the DET-06 locked pattern is `x = x + (y * z)`.
- **Action:** Style fix — rewrite as `sumOfMax = sumOfMax + maxSim` for consistency with Scorer / Cosine / weight-normalisation patterns. OR document why Monge-Elkan can use `+=` (no multiply-then-add → no FMA hazard).
- **Rationale:** Reviewer ergonomics; uniform pattern across the codebase makes "find every reduction" greps easier.

### [IMPROVEMENT] DET-W11: JaroWinkler return expression `j + float64(l)*winklerPrefixScale*(1.0-j)` has implicit left-to-right multiplication chain
- **File:** jarowinkler.go:143, jarowinkler.go:184
- **Phase introduced:** Phase 2
- **Issue:** The expression parses as `j + ((float64(l)*winklerPrefixScale)*(1.0-j))` per Go's left-to-right associativity. This is correct and deterministic; multiplication is associative under IEEE-754 round-to-nearest for short chains (three operands) only IF each intermediate result is rounded to float64 — which Go guarantees absent FMA. Subject to the same FMA-fusion caveat as DET-W2. The DET-06 pattern in scorer.go uses fully explicit parens; this site does not.
- **Standard:** `.claude/skills/determinism-standards/SKILL.md` §"Float Stability".
- **Action:** Style fix — add explicit parens: `j + ((float64(l) * winklerPrefixScale) * (1.0 - j))`. Or, for full FMA-defensive form: `j + float64(float64(l)*winklerPrefixScale) * (1.0 - j)`.
- **Rationale:** Hygiene; matches Scorer's `(x*y)+z` form. Today the matrix passes, so no urgent action.

### [IMPROVEMENT] DET-W12: Cosine's `cos = dot / (normA * normB)` uses an unparenthesised division-by-product
- **File:** cosine.go:370
- **Phase introduced:** Phase 5
- **Issue:** `cos := dot / (normA * normB)`. The parens are present around the divisor. The hazard is that `normA * normB` is a multiplication whose ordering matters. As both operands come from `math.Sqrt(...)` calls on int-derived inputs, both are IEEE-754 correctly rounded — `normA * normB` is a single rounded multiplication and deterministic.
- **Standard:** `.claude/skills/determinism-standards/SKILL.md` §"Float Stability".
- **Action:** No code change; already conformant. Recorded for completeness.
- **Rationale:** Confirms the float-determinism contract in the most arithmetic-dense site.

### [IMPROVEMENT] DET-W13: `cosineFromQGramMaps` map iteration of `qa` and `qb` for sum-of-squares uses any-order summation — correct for integers, but the comment could be sharper
- **File:** cosine.go:351-356
- **Phase introduced:** Phase 5
- **Issue:** `for _, c := range qa { sumSquaresA += c * c }` iterates a map (DET-03 candidate). The comment correctly notes "integer addition is exactly associative". This is true and correct: integer multiplication `c * c` and integer addition `+= c*c` produce the same `sumSquaresA` regardless of iteration order, because the operations are over `int` (not `float64`), and on 64-bit Go platforms `int` does not overflow on realistic q-gram counts (`c` is bounded by input length squared; `int64` overflow requires inputs of `2^31` characters).
- **Standard:** `.claude/skills/determinism-standards/SKILL.md` §"The No-Map-Iteration Rule" — output paths must not depend on map iteration. The output here is a scalar `int` — not an output path in the DET-03 sense.
- **Action:** No code change. Optionally tighten the comment to assert int (not float) operands.
- **Rationale:** Reviewer audit trail; helps a future reader confirm at a glance that the loop is safe.

### [IMPROVEMENT] DET-W14: Jaccard / Sørensen-Dice / Tversky map iteration for `totalA` / `totalB` mirrors Cosine — same correctness reasoning
- **File:** qgram_jaccard.go:221-227, sorensen_dice.go:239-246, tversky.go:346-352
- **Phase introduced:** Phase 5
- **Issue:** Three q-gram algorithms iterate `qa` and `qb` maps to sum integer multiset cardinalities. Same int-addition-is-associative reasoning as DET-W13. The intersection iteration walks the smaller-of-(qa, qb) map; the output is a scalar `int` (`intersection`), not an ordered slice. Per DET-03 this is permitted.
- **Standard:** Same as DET-W13.
- **Action:** No code change. Confirmation only.
- **Rationale:** Confirms three independent algorithms apply the same safe pattern.

### [IMPROVEMENT] DET-W15: `Tokenise` uses `[]rune(s)` conversion which silently replaces invalid UTF-8 with U+FFFD — output deterministic but lossy
- **File:** tokenise.go:161
- **Phase introduced:** Phase 1 (Tokenise primitive)
- **Issue:** `runes := []rune(s)` — Go's stdlib converts invalid UTF-8 bytes to U+FFFD individually. The conversion is deterministic across platforms (the Go spec pins it). The conversion *is* documented in tokenise.go's godoc (line 80-82, 158-160). No bug.
- **Standard:** `.claude/skills/determinism-standards/SKILL.md` does not specifically address UTF-8 replacement, but the consequence (byte-identical output for the same input) is preserved.
- **Action:** No code change. Confirmation only.
- **Rationale:** This is the right call (the alternative — error on invalid UTF-8 — would propagate an error type through every algorithm signature, breaking the pure-function contract).

### [IMPROVEMENT] DET-W16: Normalise's `applyUnicodeTransformer` uses `golang.org/x/text/...` whose internal tables come from a separate module — cross-platform determinism depends on that module's stability
- **File:** normalise.go:320-338
- **Phase introduced:** Phase 1 (Normalise primitive)
- **Issue:** The Unicode-path normalisation pipeline depends on `golang.org/x/text/unicode/norm`, `golang.org/x/text/runes`, and `golang.org/x/text/transform`. The dependency is locked in `go.mod`. Different versions of x/text COULD theoretically produce different NFC output for the same input (e.g. if the Unicode standard added a new precomposed form). Cross-platform determinism on a fixed x/text version is guaranteed (the tables are baked into the module); cross-PATCH-VERSION stability of fuzzymatch depends on x/text not changing its NFC output between fuzzymatch's released x/text version and the one consumers `go mod tidy` to. Today x/text is rev-locked in go.mod; Dependabot may bump it.
- **Standard:** `.claude/skills/determinism-standards/SKILL.md` §"Cross-Platform CI Matrix" — gates per-PR; doesn't explicitly cover dependency-version-bump determinism risk.
- **Action:** Skill clarification — add a paragraph to `.claude/skills/determinism-standards/SKILL.md` covering the x/text Unicode-table version risk; recommend pinning x/text via go.mod minor version (the current approach) and re-running the cross-platform matrix on every Dependabot bump.
- **Rationale:** Defensive process documentation. The current Dependabot grouping (test-only vs runtime) already isolates this risk; explicit acknowledgement closes the gap.

### [IMPROVEMENT] DET-W17: `Score` and `ScoreAll` apply Normalise twice — once for each input — per call; no shared cache
- **File:** scorer.go:354-357, scorer.go:502-506
- **Phase introduced:** Phase 8
- **Issue:** Each `Score(a, b)` call invokes `Normalise(a, opts)` and `Normalise(b, opts)` once. Each Normalise call may invoke the x/text transformer (Unicode path) or the byte-pass (ASCII path). The pipeline is deterministic — the same input produces the same output on the same x/text version. No determinism bug. The note here is that the per-call construction of `transform.Chain(...)` inside `applyUnicodeTransformer` (normalise.go:327) re-builds the chain each call — `transform.Transformer` is not documented as safe for concurrent reuse, so per-call construction is the safe choice. The cost is small and the alternative (a `sync.Pool`) would violate D-09 ("no mutexes, no atomics, no pools").
- **Standard:** `.claude/skills/determinism-standards/SKILL.md` §"Mutex-Free" — currently SATISFIED.
- **Action:** No code change. Confirmation only.
- **Rationale:** Re-affirms that the no-pool / no-shared-mutable-state discipline is consistent with the per-call transformer build.

---

## Cross-Platform CI Prerequisite

The cross-platform CI matrix (`.github/workflows/ci.yml`) runs
`make verify-determinism` (which executes `go test -run TestGolden_
./...`) on every PR across:
- linux/amd64 (`ubuntu-latest`)
- linux/arm64 (`ubuntu-24.04-arm`)
- darwin/amd64 (`macos-15-intel`)
- darwin/arm64 (`macos-latest`)
- windows/amd64 (`windows-latest`)

with `CGO_ENABLED=0` workflow-level enforcement. This is the canonical
cross-platform byte-identity gate. Per `.planning/phases/08-composite-
scorer/08-DETERMINISM-REVIEW.md` "Cross-Platform CI Prerequisite", the
Phase 8 Scorer golden test (`TestGolden_ScorerDefault`) runs through
the same `assertGolden` plumbing as the Phase 2-7 algorithm goldens,
so the matrix gate covers all four golden files
(algorithms.json, normalisation.json, phonetic-codes.json,
scorer-default.json) uniformly.

The local-only review here CANNOT certify cross-platform byte-identity
by itself. If a new algorithm or a `(x*y)+z` reduction site is added,
the matrix run on the introducing PR is the only place the FMA-fusion
divergence (DET-W2) would surface.

---

## GO / NO-GO

**GO** on cross-platform determinism for the codebase as it exists after
Phase 8. The library carries the LOCKED determinism contract correctly
across all 23 algorithms, two primitives (Normalise, Tokenise), the
composite Scorer, and the four golden files. No CRITICAL findings on
cross-platform byte-identity.

Five IMPORTANT findings (DET-W1 through DET-W5) flag patterns that
should be hardened before v1.0 but do NOT break current byte-identity.

Twelve IMPROVEMENT findings (DET-W6 through DET-W17) are defence-in-
depth and process. DET-W6 (phonetic-codes.json's static `regenerated_at`
literal) is the most user-visible item; renaming the field reduces a
future-maintainer-error footgun. DET-W9 (the `var _ = func() bool {...}()`
dispatch idiom in 23 dispatch_*.go files) is the most cited pattern;
documenting the carve-out in the skill closes the documentation gap.

**Prerequisite:** the cross-platform CI matrix continues to pass on every
PR. The matrix is the load-bearing empirical gate; this local review
augments it but does not replace it.
