# Phase 04 Deferred Items

Out-of-scope discoveries logged during execution per gsd-executor.md scope-boundary rule.

## Resolved

- **strcmp95.go fails `gofmt -s`** — pre-existing from plan 04-01 (commit 7fb6319). RESOLVED in plan 04-05 finalisation: `gofmt -s -w strcmp95.go` applied; the comment-continuation line under "Strcmp95Score — Jaro-Winkler + ..." was reformatted to the canonical gofmt -s indentation. `make fmt-check` exits 0.

## Outstanding (review-flagged, intentionally deferred)

The following items were flagged by `04-REVIEW.md` and explicitly deferred — low actual risk, more invasive to fix than the phase close-out window justifies. Track in the next-milestone backlog rather than carrying them across phases as un-tracked TODOs.

- **IN-04 — Strcmp95 `j > 1.0` defensive clamp could mask an overshoot bug.** `strcmp95.go` clamps the pre-prefix-boost Jaro-like value to `1.0` to absorb similar-credit overshoots. A future similar-credit refactor could produce `j = 1.5` and the clamp would silently pull it back to `1.0`, hiding the bug. Recommended fix: add a property test asserting the PRE-CLAMP `j` is in `[0, 1 + ε]` via an `export_test.go` hook returning the pre-clamp value. Needs a new export hook surface + a new property test; the existing `Strcmp95Score ≥ JaroWinklerScore` invariant covers the lower bound but not the upper. Revisit when the Scorer surface (Phase 8+) lands — natural place to revisit clamp discipline.
- **IN-06 — `examples/identifier-similarity/main_test.go` uses fragile stdout-redirect.** Two robustness gaps: (a) if `os.Pipe()` fails before `defer` is registered, `os.Stdout` is not restored; (b) `w.Close()` is unchecked. Tests pass on every reviewed platform (darwin/arm64, linux/amd64/arm64, windows/amd64); the pattern is fragile but the failure modes are vanishingly rare on `go test` happy paths. Recommended fix: helper using `os.CreateTemp` for the capture target (no pipe-blocking semantics). Revisit when the example is next extended (likely when Scorer/Scan ship).

Both items are advisory per `04-REVIEW.md` (no BLOCKER severity); deferring them does not affect any phase-completion gate. Promote either to a new phase plan or a stand-alone GitHub issue when the surrounding surfaces are next touched.
