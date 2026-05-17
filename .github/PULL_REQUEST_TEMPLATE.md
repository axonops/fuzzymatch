<!--
fuzzymatch PR template.

Fill out every applicable section. The "Algorithm-specific section"
applies ONLY when the PR adds or modifies a similarity algorithm —
delete it otherwise.
-->

## Summary

<!-- 1-3 sentences. What does this PR change, and why? -->

## Related issue

<!-- Use "Fixes #123" to auto-close on merge, or "Refs #123" otherwise. -->

Fixes #

## Type of change

- [ ] Bug fix (non-breaking)
- [ ] New feature (non-breaking)
- [ ] Breaking change (API surface, public types, or scoring semantics)
- [ ] Documentation update
- [ ] Refactor (no behavioural change)
- [ ] Performance improvement
- [ ] Test or benchmark addition
- [ ] Chore (build, CI, release, dependency)

## Algorithm-specific section

<!--
Required only for algorithm additions or scoring-formula changes.
Delete this entire block for non-algorithm PRs.
-->

### Source Origin Statement

```text
Source origin:
- Primary source: <full citation: author, year, title, journal/conference, page range, DOI>
- Studied for reference vectors: <list each implementation with its licence>
- No code copied from any source: confirmed
- No GPL/LGPL references consulted: confirmed
```

### Reference vectors

<!--
Cite at least 3 input → expected-output pairs from the primary source.
These become unit tests in algorithm_test.go.
-->

| Input A | Input B | Expected score | Citation |
|---------|---------|----------------|----------|
|         |         |                |          |

### Mathematical invariants verified by property tests

- [ ] Identity (score(x, x) == 1.0 for similarity; 0 for distance)
- [ ] Range bounds (score ∈ [0.0, 1.0])
- [ ] Symmetry (score(a, b) == score(b, a)) — note exceptions
- [ ] Triangle inequality (for distance-based algorithms only)

## Test plan

- [ ] `go test -race -shuffle=on -count=1 ./...` passes locally
- [ ] `make check` exits 0
- [ ] `make bench-compare` run for algorithm code; regressions > 10% explained below
- [ ] BDD scenarios added/updated in `tests/bdd/features/` if behaviour changed

<!-- If a regression > 10% is intentional, explain here. -->

## CHANGELOG

- [ ] `CHANGELOG.md` has an entry under `## [Unreleased]`

## Review gates

These are auto-invoked by the verifier protocol — list is here for
contributor awareness:

- [ ] **algorithm-correctness-reviewer** (algorithm PRs)
- [ ] **algorithm-performance-reviewer** (algorithm or hot-path PRs)
- [ ] **determinism-reviewer** (any change touching Normalise,
      Tokenise, golden files, or float operations)
- [ ] **api-ergonomics-reviewer** (any public API change)
- [ ] **code-reviewer** (all PRs)
