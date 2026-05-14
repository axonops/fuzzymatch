# Phase 04 Deferred Items

Out-of-scope discoveries logged during execution per gsd-executor.md scope-boundary rule.

## Resolved

- **strcmp95.go fails `gofmt -s`** — pre-existing from plan 04-01 (commit 7fb6319). RESOLVED in plan 04-05 finalisation: `gofmt -s -w strcmp95.go` applied; the comment-continuation line under "Strcmp95Score — Jaro-Winkler + ..." was reformatted to the canonical gofmt -s indentation. `make fmt-check` exits 0.

No outstanding items.
