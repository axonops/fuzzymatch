# Scan sub-package

The `scan` sub-package at `github.com/axonops/fuzzymatch/scan` is the
turnkey collection-scan layer over the `Scorer`. It answers the question
"which pairs in this collection are similar?" with a deterministic,
suppression-aware, group-aware API.

This document is a scaffold; the full consumer-facing reference is
populated in a follow-up plan. See
[`docs/requirements.md`](requirements.md) §12 for the authoritative
spec — that section pins the public API surface (`Item`, `Warning`,
`Config`, `Check`), within-group vs cross-group passes, suppression
composition, determinism guarantees, token-bucket optimisation, and
the per-scan performance budget.

## Public API

TBD. See `docs/requirements.md` §12.1 for the `Item` / `Warning` /
`Config` / `Check` shapes.

## Within-group vs cross-group passes

TBD. See `docs/requirements.md` §12.2 for the two-pass semantics.

## Suppression composition

TBD. See `docs/requirements.md` §12.3 for suppression rule composition
and order-of-evaluation.

## Determinism

TBD. See `docs/requirements.md` §12.4 for the deterministic sort key
`(Kind, NameA, NameB, GroupA, GroupB)` and the no-map-iteration
discipline.

## Token-bucket optimisation

TBD. See `docs/requirements.md` §12.5 for the token-bucket pre-filter
and the property-test invariant proving equivalence to the naive O(N²)
implementation.

## Performance

TBD. See `docs/requirements.md` §12.6 for the per-scan time and
allocation budgets.

## Repository layout

TBD. See `docs/requirements.md` §12.7 for the `scan/` directory layout
including `scan.go`, `bucket.go`, suppression rule files, and the
internal sort key.

## Thread Safety

The constructed `scan.Config` is immutable; `scan.Check` is safe for
concurrent invocation on disjoint inputs. See the Thread safety note
in [`README.md`](../README.md).
