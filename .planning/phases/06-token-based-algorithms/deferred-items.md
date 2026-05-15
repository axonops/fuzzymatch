# Phase 6 Deferred Items

Discovered during plan 06-06 (finalisation) execution. Out-of-scope items for the current plan are logged here per the GSD executor scope-boundary rule.

## Flaky property test: `TestProp_MongeElkanScore_AsymmetricWhenTokenCountAsymmetric`

**File:** `props_test.go:3203-3239`
**Discovered:** plan 06-06 finalisation (running `make bench` which invokes `go test -bench=. -benchmem -count=10 ./...` — this command runs property tests BEFORE benchmarks, so a property-test failure aborts the bench collection).
**Symptom:** The property test failed once on shrunken random Unicode-heavy input (see /tmp/make_bench.log first attempt). Re-running the test 5 times in isolation passes every time. The first attempt at `make bench` aborted with this failure before any benchmark ran.
**Reason for deferral:** The flake is pre-existing in plan 06-05's property test additions, not caused by plan 06-06 changes (which only touch finalisation surfaces: algorithms.json merge, identifier-similarity example, cross-algorithm consistency tests, bench.txt regeneration). Fixing the property test is in-scope for a follow-up plan to plan 06-05.
**Suggested fix:** Tighten the property's premise — `quick.Check` shrinkage can land on highly-shrunken Unicode-heavy inputs where the project's `Tokenise` function (which splits on identifier boundaries and not just whitespace) produces a different token-count partition than `strings.Fields` does. Either (a) replace `strings.Fields` with the project's own `Tokenise` for the premise check, or (b) skip inputs whose token-count under both partitionings is too small.
**Workaround used by plan 06-06:** `bench.txt` was regenerated via `go test -bench=. -benchmem -count=10 -run='^$' ./...` (the `-run='^$'` pattern skips all unit tests, so the flaky property test cannot abort the bench collection). The resulting bench.txt baseline includes all 19 algorithms + 3 LOCKED pathological fixtures.
