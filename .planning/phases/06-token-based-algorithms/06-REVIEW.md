---
phase: 06-token-based-algorithms
reviewed: 2026-05-15T00:00:00Z
depth: standard
files_reviewed: 56
files_reviewed_list:
  - CONTRIBUTING.md
  - Makefile
  - algoid_test.go
  - algorithms_golden_test.go
  - bench.txt
  - cross_algorithm_consistency_test.go
  - dispatch_monge_elkan.go
  - dispatch_partial_ratio.go
  - dispatch_token_jaccard.go
  - dispatch_token_set_ratio.go
  - dispatch_token_sort_ratio.go
  - docs/cross-validation.md
  - example_test.go
  - examples/identifier-similarity/main.go
  - examples/identifier-similarity/main_test.go
  - export_test.go
  - llms-full.txt
  - llms.txt
  - monge_elkan.go
  - monge_elkan_bench_test.go
  - monge_elkan_fuzz_test.go
  - monge_elkan_test.go
  - partial_ratio.go
  - partial_ratio_bench_test.go
  - partial_ratio_fuzz_test.go
  - partial_ratio_test.go
  - props_test.go
  - scripts/gen-token-ratio-cross-validation.py
  - testdata/cross-validation/token-ratios/vectors.json
  - testdata/golden/_staging/monge_elkan.json
  - testdata/golden/_staging/partial_ratio.json
  - testdata/golden/_staging/token_jaccard.json
  - testdata/golden/_staging/token_set_ratio.json
  - testdata/golden/_staging/token_sort_ratio.json
  - testdata/golden/algorithms.json
  - tests/bdd/features/monge_elkan.feature
  - tests/bdd/features/partial_ratio.feature
  - tests/bdd/features/token_jaccard.feature
  - tests/bdd/features/token_set_ratio.feature
  - tests/bdd/features/token_sort_ratio.feature
  - tests/bdd/steps/algorithms_steps.go
  - token_indel.go
  - token_indel_test.go
  - token_jaccard.go
  - token_jaccard_bench_test.go
  - token_jaccard_fuzz_test.go
  - token_jaccard_test.go
  - token_ratio_cross_validation_test.go
  - token_set_ratio.go
  - token_set_ratio_bench_test.go
  - token_set_ratio_fuzz_test.go
  - token_set_ratio_test.go
  - token_sort_ratio.go
  - token_sort_ratio_bench_test.go
  - token_sort_ratio_fuzz_test.go
  - token_sort_ratio_test.go
findings:
  critical: 0
  blocker: 0
  warning: 5
  info: 6
  total: 11
status: issues_found
---

# Phase 6: Code Review Report — Token-Based Algorithms

**Reviewed:** 2026-05-15
**Depth:** standard
**Files Reviewed:** 56
**Status:** issues_found

## Summary

Phase 6 lands five token-tier algorithms — `TokenSortRatio`, `TokenSetRatio`, `PartialRatio`
(byte + rune), `TokenJaccard`, and the parameter-rich `MongeElkan` (asymmetric + symmetric
surfaces with a 14-entry inner-metric allow-list). The implementations are correct against
their hand-derived reference vectors and cross-validated against RapidFuzz 3.14.5 for the
four Indel-based ratios. Determinism discipline (DET-03 / DET-06) is honoured throughout —
no map iteration on output paths, no transcendental floats, explicit left-to-right
parenthesisation, identity short-circuits before allocating side-effects. The shared
`token_indel.go` kernel (LCS-subsequence + Indel formula) is a faithful Wagner-Fischer
1974 transcription with a documented PITFALL-6 regression gate against the lcsstr.go
substring kernel.

No `BLOCKER` / Critical defects were found — the code is functionally sound, security-clean
(no injection vectors, no hardcoded secrets, no eval/exec, no unsafe operations), and
honours the project's Apache-2.0 file-header / fresh-transcription discipline.

However, several `WARNING`-tier findings touch maintainability and the integrity of the
golden-file regression mechanism. The most consequential is the absence of
compute-from-implementation staging-file builders for the Phase 6 algorithms (`WR-01`):
the merged `testdata/golden/algorithms.json` consumes static, hand-written JSON staging
files instead of the established `build*StagingEntries(t)` -> `assertGoldenStaging` pattern
that every Phase 2-5 algorithm uses. This breaks the `-update` workflow and removes the
self-consistency guarantee that elsewhere protects against silent score drift. Several
other warnings touch a load-bearing Pitfall-3 test that doesn't actually exercise
Regions 1/3, a defensive nil-pointer hazard in Monge-Elkan dispatch lookup, and an
unstable timestamp shape in the cross-validation corpus.

The Info-tier items are stylistic — duplicate gates, unused parameters, comment-vs-code
divergence in one test name, and one dead branch in `joinSectAndDiff`.

## Warnings

### WR-01: Phase 6 staging-file builders missing — golden-file -update workflow broken

**File:** `algorithms_golden_test.go:162-207`, `testdata/golden/_staging/{monge_elkan,partial_ratio,token_jaccard,token_set_ratio,token_sort_ratio}.json`
**Issue:** Every Phase 2-5 algorithm has a paired `build<Algo>StagingEntries(t)` builder
plus a `TestGolden_<Algo>_Staging(t)` test that writes the staging JSON file from
`fuzzymatch.<Algo>Score(...)` computed at test time (see lines 216-275 for the Levenshtein
exemplar). The Phase 6 staging files (`_staging/monge_elkan.json`,
`_staging/partial_ratio.json`, `_staging/token_jaccard.json`,
`_staging/token_set_ratio.json`, `_staging/token_sort_ratio.json`) are referenced by
`TestGolden_Algorithms_Merge` at lines 172-181 but **no Phase 6 builder functions exist**.
The staging files are hand-typed JSON whose `expected_score` values can drift silently
from the algorithm implementation: `go test -run TestGolden_ -update ./...` will not
regenerate them, so a future implementation tweak that changes the actual score will only
trip the per-algorithm unit test (which has independent hand-derived expected values) —
never the staging-file gate. The merge test (`assertGolden`) will continue passing because
it byte-compares the marshalled blob, but the upstream input is no longer linked to the
implementation. This is a regression of the established pattern's self-consistency
guarantee.

Additionally, the staging files carry non-schema fields (`derivation_note`,
`deviation_note` in `partial_ratio.json` and `token_set_ratio.json`) that JSON-unmarshal
into the merged file silently drops them. They have no functional effect today but are
inconsistent with every other staging file in the repository and will confuse a future
contributor running `-update`.

**Fix:** Add five `buildMongeElkanStagingEntries(t)`, `buildPartialRatioStagingEntries(t)`,
`buildTokenJaccardStagingEntries(t)`, `buildTokenSetRatioStagingEntries(t)`,
`buildTokenSortRatioStagingEntries(t)` helpers that compute `ExpectedScore` from the live
implementation (matching the Phase 2-5 pattern at lines 228-260), plus the corresponding
`TestGolden_<Algo>_Staging` writers. Drop the `derivation_note`/`deviation_note` fields
from the staging JSON — derivation notes belong in test code comments, not the canonical
golden artefact.

```go
func buildMongeElkanStagingEntries(t *testing.T) []goldenAlgorithmEntry {
    t.Helper()
    opts := fuzzymatch.DefaultNormalisationOptions()
    return []goldenAlgorithmEntry{
        {Name: "MongeElkan_RV-ME1_user_create_vs_usr_creating_symmetric",
            Algorithm: "MongeElkan", A: "user create", B: "usr creating",
            ExpectedScore: fuzzymatch.MongeElkanScoreSymmetric("user create", "usr creating", fuzzymatch.AlgoJaroWinkler, opts)},
        // ... remaining entries
    }
}
func TestGolden_MongeElkan_Staging(t *testing.T) {
    entries := buildMongeElkanStagingEntries(t)
    sort.Slice(entries, func(i, j int) bool { return entries[i].Name < entries[j].Name })
    file := goldenAlgorithmsFile{Version: 1, Entries: entries}
    assertGoldenStaging(t, "_staging/monge_elkan.json", file)
}
```

### WR-02: PartialRatio Pitfall-3 keystone tests don't actually exercise Region 1 / Region 3

**File:** `partial_ratio_test.go:55-194` (especially the `region_3_right_tail_wins_pitfall_3_keystone`
and `region_1_left_tail_wins_pitfall_3_keystone` rows at lines 96-110), and
`partial_ratio_test.go:305-330` (`TestPartialRatioScore_Pitfall3_Keystones`).
**Issue:** The two keystone fixtures `("abc", "ab")` and `("abc", "bc")` are documented in
the test name and inline comment as proving Region 1 left-tail and Region 3 right-tail
matter — but they don't. With `m = 2`, `n = 3`, Region 2 alone iterates `i = 0` and
`i = 1`, which yields substrings `longer[0:2] = "ab"` and `longer[1:3] = "bc"`. Region 2
catches the perfect match in BOTH cases (test commentary lines 87-90 and 104-106 actually
admit this: "Region 2 catches" / "Region 2 at i=0"). A naive single-region implementation
that DROPPED Regions 1 and 3 entirely would still pass these two scenarios.

The extended fixtures in `TestPartialRatioScore_Pitfall3_Keystones`
(`abc_a_left_tail_only` and `abc_c_right_tail_only`, lines 316-317) where `m = 1` are
similarly Region-2-dominant — Region 1's loop body is `for i := 1; i < m; i++` which is
empty for `m == 1`, and Region 2's `for i := 0; i <= n-m; i++` covers all three positions
including the matching one. The plan SUMMARY language ("Pitfall-3 keystone") is therefore
aspirational; the actual regression detector is absent.

**Fix:** Add one keystone fixture where the perfect match LIVES in Region 1 or Region 3
EXCLUSIVELY. The natural shape is an equal-length input that the equal-length symmetric
tie-break + Region 1/3 catches but Region 2 cannot. For example,
`("ab", "ba")` with `m == n == 2`: Region 2 evaluates only `i = 0` (substr `"ba"` vs
shorter `"ab"`, ratio 0.5). Without the equal-length tie-break second pass, the answer
collapses to 0.5; with it, Region 1's `i=1 → substr "b"` plus Region 3's `i=1 → substr
"a"` combined with the role-swapped second pass give the correct higher value. A clearer
fixture: `("X", "..X..")` where m=1, n=5 — Region 2 catches X at i=2; Region 1/3 don't
matter. To genuinely exercise Region 1, force a partial substring where the matching
window EXCEEDS the m-length but the longest contiguous shared prefix is shorter; this is
hard to construct without contrived inputs. The cleaner remediation is to remove the
"Pitfall 3 keystone" claim from these tests and replace it with a comment honestly
describing what is being tested (Region 2 catching m=n-1 alignments at both edges), plus
ONE pin-fixture for the equal-length re-run loop at `partial_ratio.go:317-321` which IS
load-bearing and currently has no direct coverage.

### WR-03: Monge-Elkan dispatch lookup has no defensive nil check — Phase 7 risk

**File:** `monge_elkan.go:410, 420`
**Issue:** `MongeElkanScore` resolves the inner-metric function via
`innerFn := dispatch[inner]` and then calls `innerFn(tokA, tokB)` without a nil check. The
function depends on the invariant that every AlgoID in `permittedMongeElkanInner` has a
registered dispatch slot. Today the invariant holds (all 14 permitted AlgoIDs are
registered by phase-load time), and the file-header godoc explicitly cites the
`var _ = func() bool { ... }()` package-init mechanism as the guarantee. But the godoc
also says (at lines 271-294) that Phase 7 planners will ADD 4 phonetic entries to
`permittedMongeElkanInner`. If the Phase 7 dispatch wiring is not also added in the same
PR (or is delayed across plans), the runtime behaviour is a nil-function-call panic deep
inside the inner loop — much harder to triage than a documented panic. The `nolint`
boundary cost here is small; the safety upside is clear.

**Fix:** Add a defensive nil check immediately after the dispatch lookup. Treating
"permitted-but-not-registered" as a programmer-error panic mirrors the existing
direct-call panic discipline.

```go
innerFn := dispatch[inner]
if innerFn == nil {
    panic("fuzzymatch: AlgoID " + inner.String() + " permitted as Monge-Elkan inner but dispatch slot not registered")
}
```

Alternatively, add a `TestMongeElkan_DispatchInvariant` to `monge_elkan_test.go` that walks
`permittedMongeElkanInner` and asserts each entry's `dispatch[id]` is non-nil at test
time — fail-fast at CI rather than at hot-path execution. (The walk uses a map but the
output is a single boolean per slot, which satisfies DET-03.)

### WR-04: Cross-validation corpus regenerated_at timestamp not skipped by Go loader

**File:** `token_ratio_cross_validation_test.go:131-155`, `testdata/cross-validation/token-ratios/vectors.json:6`
**Issue:** The corpus JSON carries an ISO-format `regenerated_at` timestamp
(`2026-05-15T10:21:25.669278+00:00`). The Go loader checks `rapidfuzz_version` and
`python_version` for non-empty + Python >= 3.7, but never asserts byte-stability on the
file as a whole — so the test "passes" even if a developer regenerates the corpus and
commits it with a new timestamp without re-running cross-validation. The timestamp is
informational (no test consumes its value), but its mere presence means the corpus JSON
is structurally non-deterministic across regenerations. A reviewer doing a code-archive
diff against an older commit cannot quickly tell whether `vectors.json` changed semantically
or merely had its timestamp refreshed.

This isn't strictly a regression-introducing bug, but it does undermine the "committed JSON
is the verification fixture" promise from `docs/cross-validation.md:8-9`. The simpler
Phase 4 ratcliff-obershelp corpus has the same shape and inherits the same problem; this
isn't unique to Phase 6 but is freshly added by Phase 6.

**Fix:** Either drop the `regenerated_at` field from the generator output (lose audit
trail; gain byte-stability), or add a CI step that diffs the corpus against the committed
version excluding the timestamp field — `jq 'del(._metadata.regenerated_at)'` over both
sides before `diff`. The generator script is the cleaner fix; the timestamp's documented
purpose ("informational for triage") is served equally well by the git commit history.

```python
out = {
    "version": 1,
    "_metadata": {
        "rapidfuzz_version": RAPIDFUZZ_VERSION,
        "python_version": f"{sys.version_info.major}.{sys.version_info.minor}.{sys.version_info.micro}",
        # regenerated_at removed — byte-stability over audit-trail trade-off
    },
    "entries": entries,
}
```

### WR-05: Monge-Elkan dispatch entry passes through `dispatch[inner]` rather than the direct algorithm function — creates surprising parameterisation coupling

**File:** `dispatch_monge_elkan.go:68-73`, `monge_elkan.go:410-426`
**Issue:** The Monge-Elkan dispatch slot binds `MongeElkanScoreSymmetric` with
`AlgoJaroWinkler` as the inner. But inside `MongeElkanScore`, the inner-metric resolution
goes through the `dispatch[inner]` table — meaning when Phase 5's q-gram or Tversky
algorithms are used as Monge-Elkan inners, they execute with the DISPATCH-DEFAULT
parameters (`n=3`, `α=β=1.0` for Tversky). This is documented at the consumer site, but
it creates surprising coupling: a single change to the q-gram dispatch defaults (say,
moving from trigram to bigram in some future plan) would silently shift every
Monge-Elkan score that uses a q-gram inner. The asymmetric direct-call surface
(`MongeElkanScore(a, b, AlgoTversky, opts)`) silently locks in α=β=1.0 even though
the surface-level API claims Tversky semantics — the user can't pass α=0.8/β=0.2 through
the Monge-Elkan path even when they know Tversky's behaviour is asymmetric.

The current dispatch-default fallback may be the only practical choice given the
fixed `(a, b string) float64` dispatch signature, but the silent coupling is undocumented
on the Monge-Elkan side. A reviewer or downstream consumer reading `MongeElkanScore`'s
godoc would not know that passing `AlgoTversky` gives them a Jaccard-equivalent inner
(α=β=1.0) until they trace `dispatch_tversky.go`.

**Fix:** Add a paragraph to `MongeElkanScore`'s godoc immediately after the allow-list
section explaining that q-gram and Tversky inners use the dispatch-table defaults
(`n=3`, `α=β=1.0`) and that consumers needing custom q-gram or Tversky parameters
should compose `MongeElkanScore` with a `*Score` call manually rather than passing the
AlgoID. Reference the Phase 8 Scorer issue `WithMongeElkanAlgorithm` as the proper
plumbing for that use case. No code change required if the surface is documented.

Example doc addition:

```
// PARAMETER WARNING for q-gram and Tversky inners: the inner metric is resolved via
// dispatch[inner], which binds DISPATCH DEFAULTS for parameterised algorithms — n=3 for
// QGramJaccard/SorensenDice/Cosine and n=3, α=β=1.0 for Tversky. There is no surface
// here to override those parameters. Consumers needing a non-default n or asymmetric
// Tversky weights inside Monge-Elkan should call the target algorithm directly rather
// than routing through MongeElkanScore's inner-dispatch.
```

## Info

### IN-01: Duplicate gate in PartialRatioScore — step 2 is dead code

**File:** `partial_ratio.go:281-289`
**Issue:** Steps 1 and 2 in `PartialRatioScore`:

```go
if a == b {
    return 1.0
}
if len(a) == 0 && len(b) == 0 {
    return 1.0
}
```

Step 2 is mathematically unreachable: when `len(a) == 0 && len(b) == 0`, then `a == b == ""`
and step 1 returns 1.0. The code comment correctly acknowledges this ("this branch is
effectively dead because step 1 fires first when a == '' == b"). The same pattern is
duplicated in `PartialRatioScoreRunes` at lines 448-453 and in `MongeElkanScore` at
lines 386-388/398-400 (though the latter is non-redundant because the `a == b`
short-circuit fires before tokenisation, and the `len(tokensA) == 0 && len(tokensB) == 0`
guard catches the pure-separator case where raw strings differ but tokenise to empty).

For `PartialRatio` specifically, since there's no Tokenise call, the `len()==0`-both
branch is provably dead. Keeping the dead code "for clarity / parity" makes the function
6 lines longer and gives a code reviewer a false signal that the function defends against
a state that cannot reach it.

**Fix:** Either remove the dead `len(a) == 0 && len(b) == 0` branch from
`PartialRatioScore` and `PartialRatioScoreRunes`, or convert it to a documentary comment.
The current state is correct but adds maintenance overhead.

### IN-02: `partial_ratio_test.go` test name and scenario contradict each other

**File:** `partial_ratio_test.go:96-110`
**Issue:** The test rows for the two Pitfall-3 keystone fixtures are named
`region_3_right_tail_wins_pitfall_3_keystone` and `region_1_left_tail_wins_pitfall_3_keystone`,
but the inline `derivation` strings on both rows state that Region 2 actually wins:

```
"Region 2 catches \"bc\" of \"abc\" at i=1 → indelRatio(\"bc\",\"bc\")=1.0
 (Pitfall-3 keystone — naive single-loop would miss this if the regions are mis-implemented)"
```

The test author understood the situation but the name remains misleading. A grep for
"region_3_right_tail" will surface this test as a Region-3 coverage signal, which it is
not (see WR-02). This is the same defect as WR-02 but called out separately because the
fix is a one-line rename, not a redesign.

**Fix:** Rename the two test rows to `region_2_catches_right_tail_m_equals_n_minus_1`
and `region_2_catches_left_tail_m_equals_n_minus_1` — or, paired with WR-02's remediation,
to `pitfall_3_keystone_naive_single_region_would_pass_anyway`. The truthful name leaves
no future contributor surprised by the absence of true Region-1/3 coverage.

### IN-03: `joinSectAndDiff` `diff == ""` branch is unreachable in production paths

**File:** `token_set_ratio.go:362-380`
**Issue:** `joinSectAndDiff(sect, diff)` has three branches. The `diff == ""` branch
returns `sect`. But the call sites in `tokenSetThreeWayMax` only execute after the subset
short-circuit (`if len(intersectKeys) > 0 && (len(diffABKeys) == 0 || len(diffBAKeys) == 0)`)
has been checked at line 318 and did NOT fire. The short-circuit fires when intersection
is non-empty AND at least one diff is empty. The combined-string code path only runs
when (a) intersection is empty (then `sect == ""` branch handles it correctly), OR (b)
both diffs are non-empty (then the `default` branch handles it). The remaining case —
non-empty intersection AND empty diff_ab AND non-empty diff_ba (or symmetric) — is
exactly the case the short-circuit covers. So `diff == ""` is unreachable on the
production path.

The godoc on `joinSectAndDiff` correctly states this: "that case never arises in practice
because the empty-set gate and subset short-circuit fire first". The branch is defensive
but contributes no observable behaviour. Coverage would mark it as uncovered unless a
direct unit test exercises it.

**Fix:** Either remove the `case diff == "":` branch (and let the unreachable case fall
through to `default`, which would produce `sect + " "` — also OK semantically since it
would only be reached on the impossible production path), or add an explicit panic
asserting the invariant:

```go
case diff == "":
    panic("fuzzymatch: joinSectAndDiff called with empty diff — invariant violation, subset short-circuit should have fired")
```

The panic gives clear feedback if someone restructures the caller and accidentally
breaks the invariant. Status quo is acceptable but coverage-wasteful.

### IN-04: `opts` parameter in `MongeElkanScore` is unused — `_ = opts` discards but does not document

**File:** `monge_elkan.go:392`
**Issue:** `MongeElkanScore` accepts a `NormalisationOptions` parameter and immediately
discards it with `_ = opts`. The godoc explains why (forward-compatibility with the Phase
8 Scorer option), but a future contributor running `go vet` or `golangci-lint` with strict
unused-parameter detection would have to read the godoc to understand the intent. More
subtly, the function NAME doesn't hint that opts is unused — the parameter list still
takes it. A consumer passing custom `NormalisationOptions` expects them to take effect;
the actual behaviour is `DefaultTokeniseOptions()` baked into the body at lines 393-394.

**Fix:** This is documented at the surface; the `_ = opts` is the standard Go idiom for
"intentionally unused". Two cleaner alternatives:

1. Drop the parameter from the public signature — but this is a breaking API change.
2. Rename to `_opts` or add a `//nolint:unused-parameter` directive with a link to the
   Phase 8 issue, which makes the discard explicit to future readers.

The current state is acceptable. Consider tagging the line with a `// TODO(#<issue>): wire
opts through to Tokenise once Phase 8 lands` comment so the linter signal is preserved
when the parameter starts being used.

### IN-05: `setIntersectionCardinality` smaller-set selection is asymmetric in a documentation-only way

**File:** `token_jaccard.go:261-273`
**Issue:** `setIntersectionCardinality` chooses the smaller set with
`if len(setB) < len(setA)` — i.e., when sizes are equal, `setA` becomes the smaller. This
is deterministic per-input (no map iteration of the choice itself), so DET-03 is
satisfied, and the integer-counter result is invariant under which side is iterated. But
the comment ("Walking the smaller side keeps the lookup count to len(min(setA, setB))")
doesn't make the equal-length tiebreak explicit. A reviewer comparing this to
`tokenSetThreeWayMax`'s explicit symmetric handling might wonder if the asymmetry has
correctness implications.

**Fix:** Update the comment to "When |A| == |B|, A is treated as the smaller side; this
is value-preserving because the integer-counter intersection cardinality is invariant
under argument swap." No code change.

### IN-06: `partial_ratio.go` "TODO(#TBD)" carries no issue number

**File:** `partial_ratio.go:147-154`
**Issue:** CLAUDE.md mandates that every `TODO` references a GitHub issue:
`// TODO(#42): add Ukkonen banding optimisation`. The Phase 6 PartialRatio file carries:

```
// TODO(#TBD): implement sliding-window DP per Bachmann RapidFuzz
// docs — spec line 612 explicitly defers the O(|s|·|l|) sliding-
// window variant to v1.x.
```

The `(#TBD)` placeholder makes the TODO unscannable by an issue-tracker integration
and means the deferred-work catalogue is incomplete. The same TODO is repeated in
`llms-full.txt:488-493`.

**Fix:** Open a GitHub issue tracking the v1.x sliding-window DP optimisation; replace
both `#TBD` instances with the assigned issue number. Until the issue is created, the
TODO violates the CLAUDE.md `TODO` discipline.

---

_Reviewed: 2026-05-15_
_Reviewer: Claude (gsd-code-reviewer)_
_Depth: standard_
