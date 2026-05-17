---
status: issues_found
agent: security-reviewer
scope: entire codebase (phases 1-8)
reviewed: 2026-05-17T08:30:00Z
finding_counts:
  critical: 2
  important: 11
  improvement: 18
  total: 31
---

# Security Review — fuzzymatch (Phases 1–8)

Reviewer: security-reviewer
Date: 2026-05-17
Scope: every algorithm (23), Normalise, Tokenise, Scorer, supply-chain, workflows.

The Phase 8-only review at
`.planning/phases/08-composite-scorer/08-SECURITY-REVIEW.md` enumerated
12 findings (SEC-01 to SEC-12) covering the Scorer surface. This whole-
codebase review re-asserts the still-unfixed Phase 8 findings (verified
against current source) and extends coverage to the algorithm tier,
Normalise / Tokenise, supply-chain (`go.mod` / `go.sum` / workflow
actions), and the catalogue's algorithmic-complexity profile.

Severity organisational only — every finding is surfaced regardless of
perceived priority per the security-reviewer directive.

The most notable whole-codebase gaps:

- **CRITICAL:** Two consumer-input panics from the Scorer surface are
  still present in source (SEC-01 / SEC-02). Verified at
  `scorer_options.go:381-399` (Tversky α=β=0 accepted at construction;
  panics at Score) and `scorer_options.go:425-446` (Monge-Elkan
  non-allowlisted inner accepted at construction; panics at Score).

- **IMPORTANT:** Four direct-call algorithm panics on programmer error
  remain part of the public surface — five q-gram functions
  (`CosineScore`, `CosineScoreRunes`, `QGramJaccardScore`,
  `QGramJaccardScoreRunes`, `SorensenDiceScore`, `SorensenDiceScoreRunes`,
  `TverskyScore`, `TverskyScoreRunes`) panic on `n < 1`; the two Tversky
  functions also panic on bad α/β; `MongeElkanScore` and
  `MongeElkanScoreSymmetric` panic on a non-allowlisted inner. These are
  documented as direct-call panic-on-programmer-error, but they still
  represent panics-from-consumer-input on the public surface and are
  reachable from arbitrary consumer code. Documented; not gated.

- **IMPORTANT:** No memory-bound on the worst-case DP allocations.
  `DamerauLevenshteinFullDistance` allocates a flat `(m+2)*(n+2)` int
  slice. On adversarial input pair (m = n = 50,000) this is 2.5×10^9
  ints ≈ 20 GB on a 64-bit Go runtime — Go panics with `runtime error:
  makeslice: cap out of range` (recoverable but disruptive), and on
  smaller-but-still-large pairs (e.g. m = n = 10,000) the algorithm
  silently allocates 800 MB and burns CPU for seconds. None of the DP
  algorithms (Levenshtein, OSA, Full, SWG, LCSStr, Ratcliff-Obershelp)
  has an input-size guard. `docs/performance.md` is a 60-line scaffold.

- **IMPORTANT:** `go list -m all` reports indirect modules `x/mod`,
  `x/sync`, `x/tools` in `go.sum` (via `x/text`'s own go.mod). These do
  not enter the compiled binary but are part of the supply-chain trust
  set; `scripts/verify-no-runtime-deps.sh` filters them out by design.
  Their `/go.mod` hashes in `go.sum` are still load-bearing.

- **IMPORTANT:** Five third-party action and tool dependencies pin to
  `@latest` (DavidAnson/markdownlint-cli2-action, govulncheck@latest in
  two places, goimports@latest, benchstat@latest, anchore/sbom-action@v0
  uses major-tag floating). Each of these is a supply-chain attack
  surface that a malicious upstream tag could exploit during a
  release-build window.

- **IMPORTANT:** Scorer fuzz harness still does not exist (re-asserting
  SEC-05). Whole-codebase verification: `ls scorer_fuzz_test.go` returns
  ENOENT.

- **IMPORTANT:** Ratcliff-Obershelp uses unbounded Go call-stack
  recursion (`roMatchedLength` and `roMatchedLengthRunes` —
  `ratcliff_obershelp.go:200-210` and `:279`). The depth is O(min(la,
  lb)); for multi-megabyte adversarial input the stack can grow into
  the hundreds of megabytes. Per the security-reviewer focus area, "no
  algorithm has unbounded recursion" — this is the only one that does.

The remaining findings are improvements / defence-in-depth.

---

## CRITICAL — panic from consumer input

### CR-01: WithTverskyAlgorithm permits α=β=0 → panic on first Score call

- **File:** `/Users/johnny/Development/fuzzymatch/scorer_options.go:381-399`
- **Phase introduced:** Phase 8
- **Issue:** Option layer validates `alpha < 0 || beta < 0` but does not
  validate the `α + β > 0` constraint. `TverskyScore` panics at
  `tversky.go:241-242` on the same condition. A consumer who passes
  `WithTverskyAlgorithm(1.0, 0, 0, 3)` constructs a Scorer successfully
  but every non-identical `Score`, `Match`, or `ScoreAll` call panics.
  The identity short-circuit at `tversky.go:231` (`if a == b { return
  1.0 }`) hides the panic in trivial test inputs, making the defect
  latent.
- **Standard:** `.claude/skills/go-coding-standards/SKILL.md` — panic
  discipline; `.claude/skills/fuzzymatch-review-protocol/SKILL.md` —
  consumer-input panic forbidden.
- **Action:** Code fix.
- **Rationale:** A panic that surfaces only on non-identical input
  classes as a latent denial-of-service: a Scorer that survives unit-
  test smoke (identical-pair calls) but kills the goroutine on the first
  real query. `ErrInvalidTverskyParam` already exists in `errors.go:88`
  for this case.
- **Suggested fix:** Mirror the direct-call panic check at the option
  layer:
  ```go
  if alpha < 0 || beta < 0 || (alpha == 0 && beta == 0) {
      return ErrInvalidTverskyParam
  }
  ```
  Add `TestWithTverskyAlgorithm_RejectsBothZero` unit test plus a BDD
  scenario.

(Re-asserts Phase 8 SEC-01 / 08-REVIEW.md CR-02 — defect remains.)

---

### CR-02: WithMongeElkanAlgorithm permits non-allowlisted inner → panic on first Score call

- **File:** `/Users/johnny/Development/fuzzymatch/scorer_options.go:425-446`
- **Phase introduced:** Phase 8
- **Issue:** Option layer checks dispatch-table bounds and the trivial
  self-recursion case (`inner == AlgoMongeElkan`) but does NOT consult
  `permittedMongeElkanInner` (`monge_elkan.go:291-317`). Four AlgoIDs
  pass the option's bounds check but trigger the
  `MongeElkanScoreSymmetric` panic at `monge_elkan.go:382` on the first
  call: `AlgoTokenSortRatio`, `AlgoTokenSetRatio`, `AlgoPartialRatio`,
  `AlgoTokenJaccard`. As with CR-01 the identity short-circuit at
  `monge_elkan.go:387` hides the panic on trivial inputs, making the
  defect latent.
- **Standard:** `.claude/skills/go-coding-standards/SKILL.md` — panic
  discipline; `.claude/skills/fuzzymatch-review-protocol/SKILL.md`.
- **Action:** Code fix.
- **Rationale:** Same class as CR-01 — latent panic-from-consumer-input
  reachable through the public Scorer surface, with the worst-case
  failure mode being a goroutine crash mid-query.
- **Suggested fix:**
  ```go
  if !permittedMongeElkanInner[inner] {
      return ErrInvalidAlgorithm
  }
  ```
  inserted after the existing bounds + self-recursion check. Export the
  allow-list as `permittedMongeElkanInner` is already package-scoped at
  `monge_elkan.go:291` — no new export required. Add unit test
  iterating the 4 currently-rejected AlgoIDs and asserting the option
  returns `ErrInvalidAlgorithm`.

(Re-asserts Phase 8 SEC-02 / 08-REVIEW.md IN-03 — defect remains.)

---

## IMPORTANT — DoS, info leakage, recursion-depth unbounded, supply-chain

### IM-01: WithThreshold accepts NaN → every Match silently returns false

- **File:** `/Users/johnny/Development/fuzzymatch/scorer_options.go:257-266`
- **Phase introduced:** Phase 8
- **Issue:** `WithThreshold(t)` validates `t < 0.0 || t > 1.0`; both
  comparisons evaluate `false` for `t = math.NaN()`. The NaN threshold
  is frozen into the Scorer and `Match(a, b)` returns
  `s.Score(a, b) >= s.threshold`. `x >= NaN` is always false, so the
  Scorer silently matches nothing — a denial-of-service via wrong-answer
  with no error, no log, no warning.
- **Standard:** `.claude/skills/determinism-standards/SKILL.md` §13.3 —
  NaN handling; `.claude/skills/fuzzymatch-review-protocol/SKILL.md`.
- **Action:** Code fix.
- **Rationale:** A silent-wrong-answer is worse than a panic — every
  downstream decision is corrupted invisibly. A consumer who JSON-
  decodes a threshold from a YAML/JSON config or arithmetic computes one
  from a `math.Sqrt(-1)` mistake gets a non-functional Scorer with no
  signal.
- **Suggested fix:**
  ```go
  import "math"
  if math.IsNaN(t) || t < 0.0 || t > 1.0 {
      return ErrInvalidThreshold
  }
  ```
  plus `TestWithThreshold_RejectsNaN`.

(Re-asserts Phase 8 SEC-03 / 08-REVIEW.md CR-01 — defect remains.)

---

### IM-02: WithAlgorithm + parameterised options accept NaN/+Inf weight → Score returns NaN

- **File:** `/Users/johnny/Development/fuzzymatch/scorer_options.go:150-165`
  (WithAlgorithm) plus the same pattern in every parameterised With*
  option at lines 300-315, 325-340, 350-365, 381-399, 425-446, 465-479
- **Phase introduced:** Phase 8
- **Issue:** The weight gate `if weight <= 0` evaluates `false` for both
  `math.NaN()` and `math.Inf(+1)`. The poisoned weight propagates
  through auto-normalisation: `weight = NaN` produces `sum = NaN`, the
  defensive `sum == 0` check at `scorer.go:284` returns `false` for NaN
  (NaN compares unequal to any value including itself), normalised
  weights become `NaN/NaN = NaN`, every `Score` call returns NaN, every
  `Match` returns `false`, every `ScoreAll` populates the result map
  with NaN values. `weight = +Inf` produces the same NaN-poison chain.
  The `TestProp_Scorer_NoNaN_NoInf` property test (`scorer_test.go:877`)
  exercises `DefaultScorer()` only — it does not exercise the NaN/Inf-
  weight construction path, so this defect is not caught.
- **Standard:** `.claude/skills/determinism-standards/SKILL.md` §13.3.
- **Action:** Code fix.
- **Rationale:** Silent-wrong-answer denial-of-service, same severity
  class as IM-01.
- **Suggested fix:** Tighten every parameterised With* option's weight
  gate:
  ```go
  if math.IsNaN(weight) || math.IsInf(weight, 0) || weight <= 0 {
      return ErrInvalidWeight
  }
  ```
  AND defence-in-depth tighten `scorer.go:284` to:
  ```go
  if sum <= 0 || math.IsNaN(sum) || math.IsInf(sum, 0) {
      return nil, ErrInvalidWeight
  }
  ```

(Re-asserts Phase 8 SEC-04 — defect remains.)

---

### IM-03: No Scorer-level fuzz harness covering aggregate dispatch

- **File:** missing — `/Users/johnny/Development/fuzzymatch/scorer_fuzz_test.go`
  does not exist (verified)
- **Phase introduced:** Phase 8
- **Issue:** Every catalogue algorithm has a dedicated `Fuzz*` harness
  (26 fuzzers total at `testdata/fuzz/Fuzz*`), and `FuzzNormalise` +
  `FuzzTokenise` cover those primitives. But there is no fuzz harness
  exercising `DefaultScorer().Score(a, b)`, `NewScorer(...)`,
  `Match`, or `ScoreAll` end-to-end. This matters because:
  (1) the Scorer chains Normalise + six algorithms in a single call —
  inter-algorithm interaction failures are not fuzz-covered;
  (2) the CR-01 / CR-02 panics surface only at Score-time, a class a
  Scorer-level fuzz harness would catch automatically;
  (3) the "Scorer is panic-free" claim is currently verified by
  induction over individual algorithms; a fuzz harness makes it directly
  testable.
- **Standard:** `.claude/skills/go-testing-standards/SKILL.md` — every
  public function gets a fuzz harness; `.claude/skills/fuzzymatch-review-
  protocol/SKILL.md`.
- **Action:** Code fix.
- **Rationale:** This is the gate that would have caught CR-01 and CR-02
  in CI.
- **Suggested fix:** Add `scorer_fuzz_test.go` with three harnesses:
  - `FuzzScorer_DefaultScorer_NeverPanics` — exercises `DefaultScorer`'s
    Score / Match / ScoreAll on arbitrary `(a, b)` pairs.
  - `FuzzScorer_NewScorer_NeverPanics` — exercises `NewScorer` with
    arbitrary `(threshold, weight, algoIdx)` triples to catch
    construction-time panics.
  - `FuzzScorer_AllInnerMetricsForME_NeverPanics` — iterates every
    AlgoID as ME's inner across `WithMongeElkanAlgorithm`.

Seed the corpus with embedded NUL, lone surrogates, multi-MB inputs, RTL
marks, zero-width joiners. Wire into `make test-fuzz` and CI's nightly
fuzz workflow.

(Re-asserts Phase 8 SEC-05 — gap remains.)

---

### IM-04: Ratcliff-Obershelp uses unbounded Go call-stack recursion

- **File:** `/Users/johnny/Development/fuzzymatch/ratcliff_obershelp.go:200-210`
  (byte path), `:271-280` (rune path)
- **Phase introduced:** Phase 4
- **Issue:** `roMatchedLength` and `roMatchedLengthRunes` recurse via
  ```go
  return n + roMatchedLength(a[:aLo], b[:bLo]) +
             roMatchedLength(a[aHi:], b[bHi:])
  ```
  with recursion depth bounded by O(min(la, lb)) per the file header
  comment at `:81-83`. On pathological multi-megabyte inputs (e.g. an
  all-'a' string with strategic differences forcing recursion to maximum
  depth at every level), the Go goroutine stack can grow into hundreds
  of megabytes. While Go's growable stacks make a literal stack-overflow
  panic unlikely under default limits, the security-reviewer focus area
  explicitly states "Ratcliff-Obershelp's recursive longest-common-
  substring decomposition must use iterative or bounded-depth recursion".
- **Standard:** Security-reviewer focus area (this prompt) — explicit
  callout. `.claude/skills/fuzzymatch-review-protocol/SKILL.md`.
- **Action:** Code fix (or document-and-bound).
- **Rationale:** Violates the only "no unbounded recursion" rule in the
  threat model. While not a real-world DoS today, a future
  fuzz-generated input could trigger pathological stack growth.
- **Suggested fix:** Convert `roMatchedLength` (and its rune sibling) to
  an iterative form with an explicit work-queue. Pattern:
  ```go
  type roWork struct { aLo, aHi, bLo, bHi int }
  func roMatchedLength(a, b string) int {
      stack := []roWork{{0, len(a), 0, len(b)}}
      var total int
      for len(stack) > 0 {
          w := stack[len(stack)-1]
          stack = stack[:len(stack)-1]
          as, bs := a[w.aLo:w.aHi], b[w.bLo:w.bHi]
          if len(as) == 0 || len(bs) == 0 { continue }
          aLo, aHi, bLo, bHi, n := roFindLongestMatch(as, bs)
          if n == 0 { continue }
          total += n
          stack = append(stack,
              roWork{w.aLo, w.aLo + aLo, w.bLo, w.bLo + bLo},
              roWork{w.aLo + aHi, w.aHi, w.bLo + bHi, w.bHi})
      }
      return total
  }
  ```
  Alternative (lighter touch): document the recursion-depth bound,
  add a fuzz-corpus seed designed to maximise depth, gate wall-time at
  N seconds.

(Re-asserts Phase 8 SEC-07 — concern remains.)

---

### IM-05: Damerau-Levenshtein Full allocates O(m·n) DP table — adversarial pair triggers GB-class allocation

- **File:** `/Users/johnny/Development/fuzzymatch/damerau_full.go:238-239`
  (byte path), `:346-348` (rune path)
- **Phase introduced:** Phase 2
- **Issue:** `damerauFullDP` allocates a flat `(m+2)*(n+2)` int slice
  unconditionally. On 64-bit Go this is 8 bytes per int, so an
  adversarial pair with m = n = 10,000 allocates ≈ 800 MB before
  computing anything; m = n = 50,000 attempts ≈ 20 GB and triggers
  `runtime error: makeslice: cap out of range` (recoverable but
  disruptive). Every other DP algorithm (Levenshtein, OSA, SWG, LCSStr)
  uses O(min(m,n)) two-row DP — Full is the outlier because the
  Lowrance-Wagner transposition lookup needs the historical row
  position. The file godoc even mentions a "v1.x two-row optimisation
  plan" (`damerau_full.go:113-114`) but it has not landed.
- **Standard:** `.claude/skills/performance-standards/SKILL.md` —
  allocation budgets; `.claude/skills/fuzzymatch-review-protocol/SKILL.md`
  — algorithmic-complexity DoS.
- **Action:** Code fix or document-and-warn.
- **Rationale:** Every DP-based algorithm carries O(mn) time complexity
  but most carry only O(min(m,n)) space. Full is uniquely vulnerable to
  pathological-input memory exhaustion. Calling Full from inside a
  custom Scorer composition (`WithAlgorithm(AlgoDamerauLevenshteinFull,
  1.0)`) amplifies the attack: a single `Score` call on a malicious
  pair burns a worker thread's heap allocation budget on the DP table
  alone.
- **Suggested fix (defence-in-depth, no algorithmic change required):**
  Add a `if int64(m)*int64(n) > MaxDPCells { ... }` guard at the head of
  `damerauFullDP` (and its rune-slice sibling). Return distance =
  max(m, n) on overflow — equivalent to "no common substring found"
  semantically and matching the convention of returning the worst case
  on degenerate input. Surface a typed sentinel error at the rune
  surface, or document the soft-bound in godoc. Alternative: implement
  the spec-deferred two-row optimisation (file godoc lines 113-114).

---

### IM-06: DP-based algorithms have no input-size ceiling — composite Scorer multiplies the latency hit

- **File:** All DP files:
  - `/Users/johnny/Development/fuzzymatch/levenshtein.go:115-119`
  - `/Users/johnny/Development/fuzzymatch/damerau_osa.go:125-127`
  - `/Users/johnny/Development/fuzzymatch/damerau_full.go:238-239`
  - `/Users/johnny/Development/fuzzymatch/swg.go:347-350`
  - `/Users/johnny/Development/fuzzymatch/lcsstr.go:150`
  - `/Users/johnny/Development/fuzzymatch/ratcliff_obershelp.go:237`
  - `/Users/johnny/Development/fuzzymatch/jaro.go` (O(la·w))
  - `/Users/johnny/Development/fuzzymatch/strcmp95.go`
  - `/Users/johnny/Development/fuzzymatch/monge_elkan.go:418-426`
    (O(|tA|·|tB|·cost(inner)))
  - `/Users/johnny/Development/fuzzymatch/partial_ratio.go:336-353`
    (O(|s|·|l|·max(|s|,|l|)))
- **Phase introduced:** Phases 2-6
- **Issue:** Every super-linear algorithm documents its complexity in
  its file godoc and several (Monge-Elkan, Partial Ratio) have an
  explicit "DoS notice" section. But there is no `docs/performance.md`
  guidance for consumers, no `WithMaxInputBytes` Scorer option, and no
  upper-bound enforcement anywhere in the library. A
  `DefaultScorer().Score(100KB-input, 100KB-input)` call burns at least
  six algorithms' worst-case work on the same pair; if the composition
  includes Full DL the cost is O(10^10) cells.
- **Standard:** `.claude/skills/fuzzymatch-review-protocol/SKILL.md` —
  "the `docs/performance.md` discusses how to bound input size before
  invoking an algorithm".
- **Action:** Docs + (optional) code fix.
- **Rationale:** This is the largest pre-1.0 documentation gap. Adopters
  in untrusted-input contexts (HTTP API surface, file-upload pipeline)
  need clear guidance.
- **Suggested fix:**
  1. Populate `docs/performance.md` with a per-algorithm complexity +
     recommended-input-ceiling table (the 60-line scaffold is currently
     all TBD per Phase 8 SEC-06).
  2. Add a `## DoS / Resource Bounds` section to `docs/scorer.md` cross-
     referencing each algorithm's complexity docstring.
  3. Consider a `WithMaxInputBytes(n int) ScorerOption` that returns
     `ErrInputTooLarge` from `Score` when either input exceeds `n`
     bytes. Defence-in-depth; not BLOCKING; depends on api-ergonomics-
     reviewer sign-off.

(Re-asserts Phase 8 SEC-06 — gap remains.)

---

### IM-07: Direct-call algorithm panics on programmer error are part of the public API surface

- **Files:**
  - `/Users/johnny/Development/fuzzymatch/cosine.go:197, 237`
    (`CosineScore`, `CosineScoreRunes` — panic on `n < 1`)
  - `/Users/johnny/Development/fuzzymatch/qgram_jaccard.go:146, 184`
    (panic on `n < 1`)
  - `/Users/johnny/Development/fuzzymatch/sorensen_dice.go:160, 198`
    (panic on `n < 1`)
  - `/Users/johnny/Development/fuzzymatch/tversky.go:235, 242, 281, 284`
    (panic on `n < 1` AND on bad α/β)
  - `/Users/johnny/Development/fuzzymatch/monge_elkan.go:382`
    (`MongeElkanScore` and `MongeElkanScoreSymmetric` panic on
    non-allowlisted inner)
- **Phase introduced:** Phases 5-6
- **Issue:** The library's contract is "direct-call panic on programmer
  error; Scorer surface returns typed sentinels". This is documented
  consistently across the q-gram and Monge-Elkan files. However it
  remains true that every panic listed is a panic-from-consumer-input on
  the public API — a consumer who passes `n = 0` to `CosineScore`
  receives a panic, not an error. The security-reviewer focus area is
  "every public function MUST NOT panic on arbitrary input"; the library
  exempts `n < 1` and bad α/β under the "programmer error" interpretation,
  but this exemption is not actually surfaced in the public-facing
  README or `docs/algorithms.md`.
- **Standard:** Security-reviewer focus area; `.claude/skills/go-coding-
  standards/SKILL.md` — error-vs-panic discipline.
- **Action:** Discuss-phase needed (and document if accepted).
- **Rationale:** The Scorer layer is gated; the direct surface is not.
  This is a deliberate-but-undocumented decision. Either:
  (a) Convert direct-call panics to typed errors at the public surface
      (breaking API change for v1.0); OR
  (b) Promote the "panics on programmer error" discipline to a
      first-class section of `docs/algorithms.md`, `README.md`, and
      every file's package-godoc, including a checklist of which
      inputs panic which function.
- **Suggested fix:** Option (b) — document the panic surface
  exhaustively. The user-guide-reviewer agent should review for
  completeness; the api-ergonomics-reviewer agent should rule on
  whether option (a) is desired for v1.0.

---

### IM-08: Five third-party action and tool dependencies pin to floating tags (`@latest` / `@v0`)

- **File:** `/Users/johnny/Development/fuzzymatch/.github/workflows/ci.yml:70`
  (`govulncheck@latest`), `:71` (`goimports@latest`), `:111`
  (`benchstat@latest`), `:129`
  (`DavidAnson/markdownlint-cli2-action@latest-stable`);
  `/Users/johnny/Development/fuzzymatch/.github/workflows/security.yml:33`
  (`govulncheck@latest`); `/Users/johnny/Development/fuzzymatch/.github/workflows/release.yml:63`
  (`anchore/sbom-action/download-syft@v0` — major-tag floating)
- **Phase introduced:** Phase 1
- **Issue:** Floating tags resolve at workflow run-time, exposing each
  build to whatever version the upstream maintainer (or a hostile-tag-
  push attacker) most recently published. For the release workflow this
  is especially load-bearing because the cosign-installed binary and the
  Syft SBOM generator BOTH gate the supply-chain trust chain — a
  compromised upstream release tag would propagate signatures and SBOMs
  that consumers transitively trust.
- **Standard:** `.claude/skills/fuzzymatch-review-protocol/SKILL.md` —
  supply-chain dependency pinning; CLAUDE.md "supply-chain integrity".
- **Action:** Code fix.
- **Rationale:** Mainstream Go security practice is to pin every action
  to a SHA or a specific semver, especially in `release.yml`. The
  `securego/gosec@v2.25.0` and `sigstore/cosign-installer@v3` already
  pin specifically; the inconsistency suggests an oversight rather than
  a deliberate policy.
- **Suggested fix:**
  - Pin `govulncheck`, `goimports`, `benchstat` to specific semver tags
    (Go modules support this — `@v1.1.4` style).
  - Pin `DavidAnson/markdownlint-cli2-action` to a specific semver tag
    (`@v18` or current).
  - Pin `anchore/sbom-action` to a SHA or specific semver tag
    (`@v0.17.4`, not `@v0`).
  - Add `dependabot.yml` group for action updates so future bumps land
    via PR review.

---

### IM-09: `go.sum` ships indirect-module hashes for `x/mod`, `x/sync`, `x/tools` — supply-chain surface beyond `x/text`

- **File:** `/Users/johnny/Development/fuzzymatch/go.sum`
- **Phase introduced:** Phase 1
- **Issue:** `go.sum` records hashes for `golang.org/x/mod`,
  `golang.org/x/sync`, `golang.org/x/tools` (only `/go.mod` hashes — no
  `h1:` source hashes since these don't enter the binary). These come
  from `golang.org/x/text`'s own `go.mod` requirements. They DO factor
  into Go's checksum database verification at `go mod download` time,
  so a compromised module-proxy serving a wrong-hash version of any of
  these three would break the build (which is the intended detection
  mechanism). However, the supply-chain trust set is broader than
  `scripts/verify-no-runtime-deps.sh` suggests — the script's design
  explicitly filters these out, which is correct for runtime-dep
  enforcement but should not be confused with "no supply-chain risk".
- **Standard:** `.claude/skills/fuzzymatch-review-protocol/SKILL.md` —
  supply-chain integrity.
- **Action:** Docs.
- **Rationale:** Informational — the script's design is correct, but the
  documentation framing "zero runtime dependencies" can mislead
  consumers into believing the trust set is `{stdlib, x/text}` when in
  practice it includes the transitive go.mod-graph closure of x/text.
- **Suggested fix:** Add a paragraph to `docs/performance.md` or a new
  `docs/supply-chain.md` clarifying which modules are in the trust set
  (those with hashes in `go.sum`) vs which actually compile into the
  artefact (the allowlist).

---

### IM-10: Cosign keyless trust chain anchors on GitHub OIDC — single point of trust

- **File:** `/Users/johnny/Development/fuzzymatch/.github/workflows/release.yml:75-82`
- **Phase introduced:** Phase 1
- **Issue:** The release workflow signs `checksums.txt` via
  `cosign sign-blob --bundle ... --oidc-issuer
  https://token.actions.githubusercontent.com`. The signature's
  authenticity rests on three trust assumptions:
  (1) GitHub's OIDC issuer at `token.actions.githubusercontent.com` is
      not compromised;
  (2) The Fulcio CA's signing certificates have not been compromised;
  (3) The Rekor transparency log has not been tampered with.
  Cosign keyless is the modern best practice and Sigstore's recommended
  default (since v3.0.1 removed the `COSIGN_EXPERIMENTAL` flag), but
  consumers verifying a release should be aware that the workflow's
  trust chain is purely OIDC + Sigstore, with no offline-key fallback.
- **Standard:** `.claude/skills/fuzzymatch-review-protocol/SKILL.md` —
  supply-chain.
- **Action:** Docs.
- **Rationale:** Informational — the design follows current Sigstore
  best practice, but consumers in regulated industries may need an
  offline-keys path. This is a v2.x decision, not a v1.0 BLOCKING.
- **Suggested fix:** Add a `SECURITY.md` section "Trust chain and
  verification" listing the cosign verify command and the trust
  assumptions. (`SECURITY.md` exists but does not yet cover this.)

---

### IM-11: Allocation amplification — composite Scorer hits every algorithm's allocation path

- **File:** `/Users/johnny/Development/fuzzymatch/scorer.go:368-381`
  (reduction loop)
- **Phase introduced:** Phase 8
- **Issue:** A single `s.Score(a, b)` call on `DefaultScorer()`
  dispatches to six algorithms (DamerauLevenshteinOSA, JaroWinkler,
  TokenJaccard, QGramJaccard, SorensenDice, DoubleMetaphone). Each
  algorithm runs its allocation path independently — there is no
  per-input caching across the dispatch list. For 100KB-vs-100KB input
  the six algorithms together allocate:
  - DamerauLevenshteinOSA: 3 × ~100K ints ≈ 2.4 MB (three-row DP).
  - JaroWinkler: 2 × [256]bool stack-allocated (negligible).
  - TokenJaccard: 2 tokenisations (~rune-count maps).
  - QGramJaccard: 2 map[string]int with capacity (len(s)-n+1).
  - SorensenDice: same as QGramJaccard.
  - DoubleMetaphone: two strings.Builder, bounded to 4 chars each
    (negligible).
  Total: ~5 MB heap pressure per `Score` call. ScoreAll has the same
  per-algorithm allocation set. Under 1000-call-per-second adversarial
  load, the Go runtime's GC handles this fine — but the lack of any
  cross-algorithm input-reuse means future optimisations (e.g. cache
  tokenisation across algorithms) are blocked by the dispatch-table
  abstraction.
- **Standard:** `.claude/skills/performance-standards/SKILL.md` —
  allocation discipline; security-reviewer focus area on resource
  amplification.
- **Action:** Discuss-phase needed.
- **Rationale:** Defence-in-depth — not a real-world DoS today, but a
  Score-level allocation-bound option would constrain adversarial
  worst-case heap pressure. The composite Scorer is the natural place
  to amortise repeated work across algorithms.
- **Suggested fix:** Future plan — explore a per-Score allocation
  budget option `WithMaxScoreAllocBytes(n int) ScorerOption` or a
  cross-algorithm tokenisation cache in `Scorer` (built lazily, scoped
  to the single `Score` call).

---

## Improvement — defence-in-depth, documentation gaps, error tightening

### IMP-01: q-gram map capacity hint can request `len(s)+1` slots — adversarial empty-s + huge-n attack

- **File:** `/Users/johnny/Development/fuzzymatch/q_gram.go:111`, `:141`
- **Phase introduced:** Phase 5
- **Issue:** `extractQGrams` runs `make(map[string]int, len(s)-n+1)`.
  When `len(s) >= n` and both are large, the capacity hint is correct
  and the map grows once. When `len(s) < n` the function returns
  immediately (line 105). However, the variable expression
  `len(s)-n+1` is unsigned-safe by virtue of the early return — Go's
  `make` with a negative hint panics, but the early-return guards it.
  Defence-in-depth check: ensure the early-return is BEFORE the make,
  which it is (line 105 ≪ 111). No defect today.
- **Standard:** `.claude/skills/go-coding-standards/SKILL.md` — slice/map
  capacity safety.
- **Action:** Verified clean; no action.
- **Rationale:** Spot-check confirms the guard.

---

### IMP-02: Partial Ratio rune-path charSet sized to `len(shorter)` — adversarial all-distinct-runes input

- **File:** `/Users/johnny/Development/fuzzymatch/partial_ratio.go:492`
- **Phase introduced:** Phase 6
- **Issue:** `charSet := make(map[rune]struct{}, m)` where m =
  `len(shorter)`. On a 100K-rune adversarial input with every rune
  distinct, the map grows to 100K entries. The byte-path equivalent
  (line 341) uses a stack-allocated `[256]bool` — bounded by definition.
  Rune-path map is bounded by `m`, which is the smaller-input length, so
  the worst case is still O(min(la, lb)) entries. Documented in the
  file header (Allocation budget section, `:271-276`).
- **Standard:** `.claude/skills/performance-standards/SKILL.md`.
- **Action:** No action (documented and bounded).
- **Rationale:** Spot-check confirms the bound.

---

### IMP-03: Damerau-Levenshtein Full rune path allocates unbounded `map[rune]int` da-table

- **File:** `/Users/johnny/Development/fuzzymatch/damerau_full.go:362`
- **Phase introduced:** Phase 2
- **Issue:** `da := make(map[rune]int)` with no capacity hint. Grows
  unbounded as rune characters are seen in `ra`. For an adversarial
  input where every rune is distinct (e.g. random Unicode), the map
  grows to `len(ra)` entries. Document at file header notes this is
  "unavoidable for Unicode" (`:330`). Bounded by `len(ra)` which is the
  larger dimension after the swap.
- **Standard:** `.claude/skills/performance-standards/SKILL.md`.
- **Action:** Improvement.
- **Rationale:** Add the capacity hint
  `make(map[rune]int, len(ra))` — same as `partial_ratio.go:492`. Avoids
  the 4-5 rehash allocations on medium-to-long inputs.

---

### IMP-04: Normalise's per-call `transform.Transformer` chain construction

- **File:** `/Users/johnny/Development/fuzzymatch/normalise.go:319-336`
- **Phase introduced:** Phase 1
- **Issue:** `applyUnicodeTransformer` constructs the chain
  `transform.Chain(norm.NFD, runes.Remove(runes.In(unicode.Mn)),
  norm.NFC)` on every call to Normalise (when `StripDiacritics` is
  true). The file header (`:29-31`) explicitly justifies this:
  "transform.Transformer is not documented as safe for concurrent
  reuse, and per-call construction is cheap; pooling is deferred to a
  v1.x perf revisit". Adversarial-input angle: under 1000-call-per-
  second adversarial load, each call performs the chain construction
  even when the chain is identical across calls. This is a known
  performance footprint, not a security defect, but worth noting for
  the v1.x perf revisit.
- **Standard:** `.claude/skills/performance-standards/SKILL.md`.
- **Action:** No action (documented and bounded).
- **Rationale:** Documented design decision.

---

### IMP-05: SWG accepts NaN / +Inf params and produces "deterministic-but-meaningless" output

- **File:** `/Users/johnny/Development/fuzzymatch/swg.go:171-173`
- **Phase introduced:** Phase 3
- **Issue:** `SmithWatermanGotohScoreWithParams` documents (lines
  171-173): "No validation is performed on params: nonsense values
  (e.g. GapOpen > 0, NaN, +Inf) produce a deterministic-but-meaningless
  result." A consumer passing `SWGParams{Match: math.NaN()}` to
  `WithSmithWatermanGotohAlgorithm` from the Scorer surface gets a
  NaN-polluted Scorer — same class as IM-02.
- **Standard:** `.claude/skills/determinism-standards/SKILL.md` §13.3 —
  NaN handling.
- **Action:** Code fix.
- **Rationale:** Same NaN-poison class as IM-02 but on the SWG params
  surface. The Scorer option `WithSmithWatermanGotohAlgorithm`
  (`scorer_options.go:465-479`) does not validate `params`.
- **Suggested fix:** Add a `SWGParams.Validate()` method returning
  `ErrInvalidConfiguration` on NaN/Inf or invalid sign convention. Call
  it from `WithSmithWatermanGotohAlgorithm`. Direct-call surfaces
  retain the documented "GIGO" behaviour.

---

### IMP-06: `errors.go` sentinels are flat `errors.New` — no per-call context

- **File:** `/Users/johnny/Development/fuzzymatch/errors.go:48-162`
- **Phase introduced:** Phase 1
- **Issue:** Every sentinel is a flat `errors.New("fuzzymatch: ...")`
  value. `NewScorer` and the With* options return the sentinel
  verbatim with no wrapping. Spot-check confirms no error message
  embeds user input (good — no info-leakage risk). However, a consumer
  receiving `ErrInvalidWeight` from `NewScorer(opts...)` has no way to
  tell WHICH option failed without the option-index in the wrapper.
- **Standard:** `.claude/skills/go-coding-standards/SKILL.md` — error
  wrapping conventions.
- **Action:** Improvement.
- **Rationale:** Error-message tightening, not a security defect.
  Useful for debuggability under adversarial input.
- **Suggested fix:** Wrap option errors with the option index:
  ```go
  for i, opt := range opts {
      if err := opt(&cfg); err != nil {
          return nil, fmt.Errorf("fuzzymatch: option[%d]: %w", i, err)
      }
  }
  ```
  Verify the wrapped error still satisfies `errors.Is(err, ErrInvalidWeight)`.

---

### IMP-07: `nil` option in `NewScorer` panics with cryptic nil-function dereference

- **File:** `/Users/johnny/Development/fuzzymatch/scorer.go:199-203`
- **Phase introduced:** Phase 8
- **Issue:** A literal `nil` in the variadic `opts` slice (or an option
  set via `opts[i] = nil` after construction of the slice) panics with
  "runtime error: invalid memory address or nil pointer dereference"
  rather than returning a typed error. Consumer error rather than a real
  attack surface, but a defence-in-depth improvement.
- **Standard:** `.claude/skills/go-coding-standards/SKILL.md` — defensive
  coding.
- **Action:** Code fix.
- **Rationale:** Same as Phase 8 SEC-09 — typed error preferred to
  cryptic panic.
- **Suggested fix:**
  ```go
  for i, opt := range opts {
      if opt == nil {
          return nil, fmt.Errorf("fuzzymatch: nil option at index %d: %w",
              i, ErrInvalidConfiguration)
      }
      if err := opt(&cfg); err != nil { ... }
  }
  ```

(Re-asserts Phase 8 SEC-09 — gap remains.)

---

### IMP-08: Scorer's `applyNormalisation` interacts with token-based algorithms — documented but worth re-asserting

- **File:** `/Users/johnny/Development/fuzzymatch/scorer.go:312-321`
- **Phase introduced:** Phase 8
- **Issue:** When the Scorer is constructed with
  `WithNormalisation(opts)`, Score applies `Normalise(s, normaliseOpts)`
  to both inputs at the boundary. Token-based algorithms (Monge-Elkan,
  Token*, PartialRatio) then call `Tokenise` internally on the already-
  normalised string. This means the Scorer's normalisation is applied
  TWICE-ish to token-based algorithms (Normalise's collapse + Tokenise's
  re-split). Documented (lines 314-321), but consumers may be surprised
  that `Normalise(s)` + `Tokenise(s)` is not the same as
  `Tokenise(Normalise(s))`. This is a correctness-discipline question
  rather than a security one, but the security-reviewer angle is:
  consumers tuning a Scorer for adversarial input may believe they have
  control over the normalisation pipeline when they actually only
  control its first half.
- **Standard:** `.claude/skills/documentation-standards/SKILL.md`.
- **Action:** Docs.
- **Rationale:** Improve the docstring with a small Mermaid-style
  example showing the pipeline.

---

### IMP-09: Token-based algorithms internally call `Tokenise(s, DefaultTokeniseOptions())` — no consumer control

- **File:** `/Users/johnny/Development/fuzzymatch/monge_elkan.go:394-395`,
  `/Users/johnny/Development/fuzzymatch/token_jaccard.go`,
  `/Users/johnny/Development/fuzzymatch/token_sort_ratio.go`,
  `/Users/johnny/Development/fuzzymatch/token_set_ratio.go`
- **Phase introduced:** Phase 6
- **Issue:** Every token-based algorithm hard-codes
  `DefaultTokeniseOptions()`. Consumers wanting a non-default tokeniser
  (e.g. custom `SeparatorChars`) cannot pass one through. The Scorer
  options `WithMongeElkanAlgorithm` accept a `NormalisationOptions`
  parameter but ignore it (`monge_elkan.go:393`). This is a feature gap,
  not a security defect — but a consumer who Normalise's with one
  separator set and assumes Tokenise will use the same is surprised.
- **Standard:** `.claude/skills/fuzzymatch-review-protocol/SKILL.md` —
  documentation discipline.
- **Action:** Docs (+ future API expansion).
- **Rationale:** Documentation gap. Surface the hard-coded
  `DefaultTokeniseOptions()` choice in each algorithm's godoc.

---

### IMP-10: Empty `SeparatorChars` in `NormalisationOptions` silently degrades to "whitespace-only collapse"

- **File:** `/Users/johnny/Development/fuzzymatch/normalise.go:68`
- **Phase introduced:** Phase 1
- **Issue:** `SeparatorChars: ""` combined with `StripSeparators: true`
  is documented (`:68-69`) as "equivalent to whitespace-only collapsing".
  An adversarial config could exploit this — a consumer who programmatically
  builds `NormalisationOptions` from a JSON-decoded config where
  `SeparatorChars` is missing gets a different normalisation than
  intended. Not a security defect; defence-in-depth surfacing.
- **Standard:** `.claude/skills/documentation-standards/SKILL.md`.
- **Action:** No action (already documented).
- **Rationale:** Spot-check confirms the documentation; mention in any
  future "configuring from JSON" doc.

---

### IMP-11: Phonetic algorithms accept (and silently drop) non-ASCII runes — no signal

- **File:** `/Users/johnny/Development/fuzzymatch/soundex.go:56-61`,
  `/Users/johnny/Development/fuzzymatch/double_metaphone.go:69-74`,
  `/Users/johnny/Development/fuzzymatch/nysiis.go:75-80`,
  `/Users/johnny/Development/fuzzymatch/mra.go:53-58`
- **Phase introduced:** Phase 7
- **Issue:** All four phonetic algorithms document that non-ASCII runes
  are silently dropped before encoding. This is the correct behaviour
  for phonetic codes that are by definition ASCII-letter-keyed. From a
  security angle, an adversarial input with all-non-ASCII content would
  produce an empty phonetic code, which on the Scorer surface produces
  a score of 1.0 for any two such inputs (vacuous-match convention).
  This could be weaponised by an attacker who controls input strings —
  feed two unrelated multi-byte UTF-8 strings to a Scorer with phonetic
  weights and get a vacuous-match score that contributes to a false-
  positive composite. Documented but the security implication is not
  surfaced.
- **Standard:** Security-reviewer focus area; `.claude/skills/fuzzymatch-
  review-protocol/SKILL.md`.
- **Action:** Docs.
- **Rationale:** Defence-in-depth — surface the adversarial-input angle
  in each phonetic file's godoc and in `docs/scorer.md`.
- **Suggested fix:** Add a paragraph: "Phonetic algorithms over
  predominantly non-ASCII input produce empty codes and a vacuous-match
  1.0 score. Consumers feeding untrusted multi-byte input to a Scorer
  with phonetic weights should pre-validate that at least one
  ASCII-letter rune is present, or compose with `Normalise +
  StripDiacritics` first."

---

### IMP-12: Phonetic algorithms operate on ASCII letters by `>='A' && <='Z'` byte arithmetic — embedded NUL bytes treated as non-letters

- **File:** `/Users/johnny/Development/fuzzymatch/soundex.go:111-126`
  (soundexGroup); similar pattern in double_metaphone, nysiis, mra
- **Phase introduced:** Phase 7
- **Issue:** Embedded NUL bytes (0x00) in input strings are dropped
  before phonetic encoding (they are not ASCII letters). This is
  correct behaviour but the `_test.go` files do not include an embedded-
  NUL canonical reference vector. Defence-in-depth.
- **Standard:** Security-reviewer focus area — embedded NUL handling.
- **Action:** Tests.
- **Rationale:** Add an embedded-NUL test row to each phonetic algorithm
  asserting the documented "silently drop" behaviour.

---

### IMP-13: Phonetic similarity timing — score lookup performs at-most-4-character byte compare

- **File:** `/Users/johnny/Development/fuzzymatch/double_metaphone.go`,
  `/Users/johnny/Development/fuzzymatch/soundex.go`,
  `/Users/johnny/Development/fuzzymatch/nysiis.go`,
  `/Users/johnny/Development/fuzzymatch/mra.go`
- **Phase introduced:** Phase 7
- **Issue:** Per the prompt: "side-channel timing on phonetic
  comparisons". Phonetic algorithms produce a fixed-or-bounded-length
  code (≤ 4 chars for Soundex/DM, ≤ 6 for NYSIIS/MRA). The final
  score comparison is a byte-equality check. Timing is constant-time in
  the secret-handling sense by virtue of the bounded code length —
  there is no early-exit on byte mismatch in a comparison loop that
  would leak position information. Verified by reading
  `DoubleMetaphoneScore` (one byte equality) and `SoundexScore`.
  Spot-check passed.
- **Standard:** Security-reviewer focus area — timing leakage.
- **Action:** No action.
- **Rationale:** The library does not handle secrets; even so, the
  phonetic-score comparison surface is not a timing-leak hazard.

---

### IMP-14: No fuzz-corpus seeds with multi-MB inputs

- **File:** All `*_fuzz_test.go` files
- **Phase introduced:** Phases 2-8
- **Issue:** Inspection of `damerau_full_fuzz_test.go` (representative)
  shows seed entries with at most ~10-character inputs. Per the
  security-reviewer focus area: "Fuzz tests include very long inputs
  (multi-KB) without timeout". Native Go fuzz generates inputs of varied
  lengths automatically, so this is not a strict gap — but adding a
  hand-curated multi-MB seed would guarantee the worst-case allocation
  paths are exercised on every nightly fuzz run.
- **Standard:** Security-reviewer focus area;
  `.claude/skills/go-testing-standards/SKILL.md`.
- **Action:** Tests.
- **Rationale:** Defence-in-depth — guarantee the pathological-input
  path is fuzz-exercised.
- **Suggested fix:** Add a 64KB+ seed to each fuzz corpus
  (`testdata/fuzz/Fuzz*/`).

---

### IMP-15: Make-script generators are Python — supply-chain footprint for cross-validation corpora

- **File:** `/Users/johnny/Development/fuzzymatch/scripts/gen-swg-cross-validation.py`,
  `gen-ratcliff-obershelp-cross-validation.py`,
  `gen-token-ratio-cross-validation.py`,
  `gen-phonetic-cross-validation.py`
- **Phase introduced:** Phases 3-7
- **Issue:** The cross-validation corpora at
  `testdata/cross-validation/*/vectors.json` are committed to the repo
  and consumed by the Go test suite at run-time. They are generated by
  Python scripts that import third-party packages (e.g.
  `rapidfuzz==3.14.5`, `jellyfish==1.2.1`). These Python deps are NOT
  part of the runtime or test trust set — but a CI run that regenerated
  the corpora from a compromised PyPI mirror would inject adversarial
  reference vectors that would then pin "wrong" expected outputs in the
  test suite.
- **Standard:** Security-reviewer focus area — code provenance;
  `.claude/skills/fuzzymatch-review-protocol/SKILL.md`.
- **Action:** Docs.
- **Rationale:** The corpora are committed once and reviewed once; this
  is informational. Document that the corpus regeneration is a
  privileged operation requiring algorithm-correctness-reviewer
  sign-off.
- **Suggested fix:** Add a `CORPUS_REGENERATION.md` documenting the
  trust assumptions of the Python regeneration scripts.

---

### IMP-16: Scorer reduction does not guard against NaN scores from individual algorithms

- **File:** `/Users/johnny/Development/fuzzymatch/scorer.go:368-381`
- **Phase introduced:** Phase 8
- **Issue:** The reduction loop accumulates `acc = acc +
  (entry.weight * score)` without checking that `score` is finite.
  Every algorithm's property tests assert score ∈ [0, 1] — but a future
  algorithm regression that returned NaN under a corner case would
  silently poison the composite Score and every downstream Match /
  ScoreAll. Defence-in-depth.
- **Standard:** `.claude/skills/determinism-standards/SKILL.md` —
  NaN handling.
- **Action:** Code fix (optional).
- **Rationale:** Same as Phase 8 SEC-08.
- **Suggested fix:** Add a property test
  `TestProp_Scorer_NoNaN_NoInf_AllSingleAlgoCompositions` that iterates
  every AlgoID as a single-algorithm Scorer and asserts finite output
  on random pairs. This catches per-algorithm regressions at the Scorer
  surface.

---

### IMP-17: No regex usage — verified clean

- **File:** library-wide (verified by `grep regexp`)
- **Phase introduced:** N/A
- **Issue:** Per the security-reviewer focus area: "Where regular
  expressions are used (rare — possibly in tokenisation or
  normalisation), verify: ...". Verification: no `regexp` import anywhere
  in non-test code (`grep -n "regexp\|MustCompile" /Users/.../*.go |
  grep -v _test.go` returned empty). Tokenise and Normalise use
  hand-coded byte/rune loops with constant-time `[128]bool` tables.
- **Standard:** Security-reviewer focus area.
- **Action:** No action.
- **Rationale:** Verified clean.

---

### IMP-18: `DefaultScorer()` panic on internal inconsistency — verified bounded

- **File:** `/Users/johnny/Development/fuzzymatch/scorer.go:586-592`
- **Phase introduced:** Phase 8
- **Issue:** Documented panic on `NewScorer(DefaultScorerOptions()...)`
  failure. Verified unreachable under any consumer input pattern: the
  six AlgoIDs in `DefaultScorerOptions` are package-load-time bound,
  the threshold 0.85 is a literal in [0, 1], the weights are 1.0
  literals.
- **Standard:** Security-reviewer focus area — panic discipline.
- **Action:** No action.
- **Rationale:** Spot-check confirms boundedness (also confirmed in
  Phase 8 SEC-10).

---

## Verification of focus-area requirements (whole codebase)

| Focus area | Status | Notes |
|------------|--------|-------|
| Every algorithm documents complexity in godoc | YES | Verified across 23 files. Aggregate composite complexity in `docs/scorer.md` not documented — IM-06. |
| Super-linear complexity algorithms documented with warnings | PARTIAL | ME / PartialRatio have explicit "DoS notice" sections; others document complexity but not DoS — IM-06. |
| Fuzz tests include multi-KB inputs | PARTIAL | Per-algorithm fuzz exists for all 26 surfaces; multi-MB hand-curated seeds missing — IMP-14. |
| `docs/performance.md` discusses input-size bounding | NO | 60-line scaffold all TBD — IM-06. |
| No unbounded recursion | NO | Ratcliff-Obershelp uses Go-call-stack recursion bounded only by O(min(la, lb)) — IM-04. |
| Public functions never panic on arbitrary input | NO | CR-01 (Tversky α=β=0); CR-02 (ME non-allowlisted inner); IM-07 (direct-call algorithm panics on programmer error). |
| `PropAlgorithm_NeverPanics` property test exists | PARTIAL | `TestProp_Normalise_NeverPanics` + `TestProp_Tokenise_NeverPanics` exist; no Scorer-level harness — IM-03. |
| Error messages do not embed user input | YES | Verified across `errors.go` and Scorer surfaces. |
| No timing-based information leakage | YES | No secrets handled; phonetic surface has bounded comparison — IMP-13. |
| Zero runtime dependencies | YES (curated) | Single curated `x/text` dep; `verify-no-runtime-deps.sh` clean. `go.sum` still records `x/mod`/`x/sync`/`x/tools` `/go.mod` hashes — IM-09. |
| Test deps isolated in `tests/bdd/go.mod` | YES | Verified. |
| `govulncheck ./...` clean | UNKNOWN | Runs in CI on every PR + weekly; status not verified in this offline review. |
| No GPL/LGPL-derived code | YES | algorithm-licensing-reviewer enforced; provenance statements in every algorithm file. |
| Regex safety | N/A | No regex usage — IMP-17. |
| Invalid UTF-8 graceful | YES | Normalise replaces with U+FFFD; algorithm rune paths use Go's `[]rune` conversion which replaces invalid sequences. |
| Embedded NULs | PARTIAL | Phonetic algorithms drop them silently; no explicit test row — IMP-12. |

---

## Summary

**Must fix before v1.0 (CRITICAL):**
1. CR-01 — `WithTverskyAlgorithm` α=β=0 panic
2. CR-02 — `WithMongeElkanAlgorithm` non-allowlisted inner panic

**Should fix before v1.0 (IMPORTANT — DoS / silent-wrong-answer / supply-chain):**
3. IM-01 — `WithThreshold` NaN handling
4. IM-02 — `WithAlgorithm` NaN/Inf weight handling
5. IM-03 — Scorer-level fuzz harness
6. IM-04 — Ratcliff-Obershelp iterative or bounded-recursion
7. IM-05 — Damerau-Levenshtein Full O(m·n) DP table guard
8. IM-06 — Populate `docs/performance.md`; document Scorer-aggregate complexity
9. IM-07 — Discuss-phase: convert direct-call panics OR exhaustively document
10. IM-08 — Pin floating action / tool tags
11. IM-09 — Document `go.sum` trust set beyond `x/text`
12. IM-10 — `SECURITY.md` cosign trust-chain section
13. IM-11 — Discuss-phase: Scorer-level allocation-bound option

**Defence-in-depth (IMPROVEMENT):**
14. IMP-03 — Damerau-Full rune-path da-map capacity hint
15. IMP-05 — SWG params NaN/Inf validation at Scorer surface
16. IMP-06 — Wrap option errors with index
17. IMP-07 — `nil` ScorerOption guard
18. IMP-08 — Document Normalise + Tokenise pipeline interaction
19. IMP-09 — Surface hard-coded `DefaultTokeniseOptions` in algorithm docs
20. IMP-11 — Phonetic algorithm vacuous-match adversarial-input warning
21. IMP-12 — Embedded NUL test rows in phonetic suites
22. IMP-14 — Multi-MB fuzz-corpus seeds
23. IMP-15 — Document Python regeneration script trust assumptions
24. IMP-16 — Per-AlgoID single-Scorer NaN/Inf property test

**Verified clean (no action):**
25. IMP-01 — q-gram map capacity hint
26. IMP-02 — Partial Ratio rune-path charSet bound
27. IMP-04 — Normalise per-call transformer construction
28. IMP-10 — Normalise empty `SeparatorChars` documented
29. IMP-13 — Phonetic timing surface
30. IMP-17 — No regex usage
31. IMP-18 — `DefaultScorer()` panic bound

---

_Reviewed: 2026-05-17_
_Reviewer: security-reviewer_
_Scope: entire codebase (Phases 1-8)_
