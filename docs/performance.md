# Performance

This document covers fuzzymatch's per-algorithm performance budgets,
benchmark methodology, and the cost/accuracy trade-off across the
catalogue. The authoritative budgets live in
[`docs/requirements.md`](requirements.md) §14; this document expands on
them with concrete benchmark numbers as each phase lands.

This document is a scaffold. Benchmark numbers are populated in Phase
2+ as algorithms ship; the Phase-1 `Normalise` and `Tokenise`
benchmarks live in `normalise_bench_test.go` and
`tokenise_bench_test.go` respectively, with results captured in the
committed `bench.txt`.

See also:

- [`docs/requirements.md`](requirements.md) §14 — performance budgets
  (per-algorithm, Scorer, Normalisation).
- [`docs/tuning.md`](tuning.md) — accuracy/cost trade-off in algorithm
  selection.
- [`bench.txt`](../bench.txt) — committed benchstat baseline.
- [`CONTRIBUTING.md`](../CONTRIBUTING.md) — local `make bench` /
  `make bench-compare` workflow.

## Benchmark methodology

TBD. `go test -bench=. -benchmem -count=10` per algorithm at the size
classes specified in `docs/requirements.md` §14.4. CI runs
`make bench-compare` informationally until a self-hosted bench runner
is available (D-09).

## Per-algorithm budgets

TBD. See `docs/requirements.md` §14.1 for the per-algorithm time and
allocation budgets. Each algorithm's implementation file enforces its
budget through dedicated benchmarks.

## Scorer budgets

TBD. See `docs/requirements.md` §14.2.

## Normalisation budgets

TBD. See `docs/requirements.md` §14.3. The Phase-1 `Normalise`
benchmark results live in `bench.txt`.

## ASCII fast paths

Several algorithms (Levenshtein, Jaro, Jaro-Winkler, Hamming, etc.)
have ASCII fast paths that operate on `[]byte` for inputs whose every
byte is `< 0x80`. The fast path avoids `[]rune` conversion and uses
stack-allocated buffers for inputs under a per-algorithm threshold.
Documented per algorithm in `docs/algorithms.md`.

## Benchstat regression detection

TBD. `make bench-compare` runs `benchstat $(BENCH_FILE) $(BENCH_NEW_FILE)`
and reports any > 10% regression. The regression gate is local-developer
discipline until a self-hosted CI runner is available (D-09; see
[`CONTRIBUTING.md`](../CONTRIBUTING.md) §Benchmark discipline).
