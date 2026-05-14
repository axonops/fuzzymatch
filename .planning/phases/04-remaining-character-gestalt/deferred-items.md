# Phase 04 Deferred Items

Out-of-scope discoveries logged during execution per gsd-executor.md scope-boundary rule.

## Discovered during Plan 04-04 (Ratcliff-Obershelp cross-validation)

- **strcmp95.go fails `gofmt -s`** — pre-existing from plan 04-01 (commit 7fb6319). Not touched by plan 04-04. Should be auto-fixed by `make fmt` in a small follow-up `style(04)` commit OR resolved in plan 04-05 finalisation when the file is touched for other reasons. Detection: `make fmt-check` reports `strcmp95.go` needs reformat.
