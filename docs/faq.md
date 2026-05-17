# Frequently Asked Questions

This document covers common questions about fuzzymatch's scope, its
inclusions and exclusions, and the reasoning behind some of the more
opinionated design choices. The entries below are mandated by
requirement DX-06; subsequent phases extend this list.

## What is Validate and when should I use it?

`fuzzymatch.Validate(a, b string) []Warning` is the consumer-facing
input-quality diagnostic. It is a pure function that inspects the two
input strings and returns a sorted slice of `Warning` values describing
problematic-but-non-fatal input shapes — empty input, unequal-length
input for Hamming, no tokens after normalisation for token-tier
algorithms, all-non-ASCII input for ASCII-only algorithms, and
pathologically large input. `Validate` never panics and returns `nil`
(not an empty slice) when no warning applies.

Use it on code paths that audit input quality — typically log lines,
telemetry, or pre-flight checks ahead of a `Scorer.Score` call. The
algorithms themselves are tolerant: they accept any string and return
a sensible value (the lenient comparison-data contract documented in
`docs/requirements.md` §6.A). `Validate` is the optional companion
that says whether the value will be meaningful.

The recommended idiom is "Validate-then-Score":

```go
warnings := fuzzymatch.Validate(a, b)
for _, w := range warnings {
    log.Printf("input quality: %s (%s): %s", w.Kind, w.Algorithm, w.Detail)
}
score := fuzzymatch.DefaultScorer().Score(a, b)
```

See [`docs/algorithms.md#input-validation-with-fuzzymatchvalidate`](algorithms.md#input-validation-with-fuzzymatchvalidate)
for the full per-`WarnKind` reference, [`docs/best-practices.md`](best-practices.md)
for production patterns, and [`examples/validate-input-quality/`](../examples/validate-input-quality/main.go)
for a runnable end-to-end programme.

## Why no Soft-TFIDF?

Soft-TFIDF (Cohen, Ravikumar & Fienberg 2003) layers token-level
soft-matching on top of TF-IDF weights computed from a domain corpus.
The token similarities are typically Jaro-Winkler or Levenshtein-based;
the corpus weights downweight common tokens ("the", "company", "ltd")
and upweight rare distinguishing tokens ("AxonOps", "Cassandra").

fuzzymatch deliberately omits Soft-TFIDF for two reasons.

First, **statelessness**. The corpus IDF table is consumer-specific,
mutable, and non-trivial to ship — different consumers operate on
different corpora (audit field names, product catalogues, person
names) with completely different token distributions. A library that
accepts an IDF table as a parameter is shipping API surface for a
problem the library cannot solve well in the abstract; consumers in
practice end up computing their own IDF anyway and would not gain
much from the wrapper.

Second, **alternatives**. The Monge-Elkan + Jaro-Winkler-inner
composition (available via `WithMongeElkanAlgorithm(weight,
AlgoJaroWinkler)`) captures the most useful Soft-TFIDF dynamic — soft
token matching — without the corpus dependency. For workloads where
common-token down-weighting is critical, consumers compute IDF in
their own package and pre-multiply the per-token Monge-Elkan
contribution.

See [`docs/requirements.md`](requirements.md) §4 for the formal
out-of-scope statement.

## Why no Needleman-Wunsch?

Needleman-Wunsch (1970) is a **global** sequence-alignment algorithm —
it produces an alignment over the full length of both strings, even
when the optimal answer is "they share a substring but otherwise
differ". For identifier matching, schema deduplication, and
configuration-vocabulary similarity (the use cases fuzzymatch is
designed for), the right semantics are almost always "what is the
best-matching subsequence" — and that is precisely what
Smith-Waterman-Gotoh (1981, with Gotoh's 1982 affine-gap improvement)
delivers. SWG subsumes the rare cases where global alignment is the
right tool, with vastly more useful behaviour on the common cases.

See [`docs/requirements.md`](requirements.md) §4 for the formal
out-of-scope statement.

## Why no Metaphone 3?

Metaphone 3 (Lawrence Philips, 2009) is encumbered by U.S. Patent
7,440,941 — assigned to Philips personally and not (as of the
patent-screen date in `algorithm-licensing-standards`) released under
a royalty-free licence. AxonOps declines patent-encumbered algorithms
**regardless of enforcement posture or the availability of
alternative implementations**: the project's Apache-2.0 licence
hygiene requires that consumers be able to rely on the library
without independent patent review, and Metaphone 3 in any form
breaks that promise.

Double Metaphone (Philips 2000, patent-free) and NYSIIS cover the
phonetic-encoding use cases. Soundex covers the simplest case. The
catalogue is intentionally limited to the patent-free phonetic
algorithms.

See `.claude/skills/algorithm-licensing-standards/SKILL.md` for the
patent-screen discipline and `docs/requirements.md` §4 for the formal
out-of-scope statement. The Metaphone 3 precedent is the canonical
reference for any future patent-screen decision.

## Why no embeddings / ML?

fuzzymatch is a **pure-function library**. Embedding-based matching
(e.g. SBERT, OpenAI embeddings, transformer-derived similarity)
requires:

- Model storage (hundreds of MB to multiple GB per model).
- Runtime dependencies (typically torch / onnxruntime / transformers,
  none of which can be incorporated without breaking the no-cgo and
  zero-non-stdlib-runtime-dep constraints).
- Consumer-side stateful caching to amortise the embedding cost.
- Tokeniser-vocabulary drift across model versions (the same model
  family at different versions does not produce comparable
  embeddings).

None of those are compatible with fuzzymatch's design goals: a
pure-Go, deterministic, allocation-budgeted, supply-chain-auditable
library. Consumers who need embedding-based matching are well-served
by dedicated tools (e.g. pgvector, ChromaDB, Pinecone, faiss);
fuzzymatch sits one layer below in the stack as the classical-NLP
counterpart.

See `docs/requirements.md` §4 for the formal out-of-scope statement.

## Why phonetic-as-binary in the Scorer?

The four phonetic algorithms in the catalogue (Soundex, Double
Metaphone, NYSIIS, MRA) produce **discrete encoded codes** rather
than continuous similarity scores. Soundex encodes "Robert" as
"R163"; two inputs either share that code or they don't.

The canonical normalisation, per `docs/requirements.md` §7.20–§7.23,
is binary: 1.0 if the codes match exactly, 0.0 otherwise. Trying to
produce a continuous score (e.g. Levenshtein over the codes) is
**possible** — and consumers wanting that behaviour can compute it
directly from the encoder output (`SoundexCode`, `DoubleMetaphoneKeys`,
`NYSIISCode`, `MRACode`) — but it is not the default because the
canonical phonetic algorithms are not designed for graded similarity.
The binary form is the well-defined Scorer contribution; the
continuous form is a consumer-side composition with whatever inner
metric the consumer chooses.

See [`docs/extending.md`](extending.md) for the pattern of composing
phonetic codes with edit distance.

## Why aren't algorithm functions generic?

Go generics (introduced in 1.18) seem like a natural fit for
"compare two strings" with the type parameter being the string type.
The catalogue, however, is intentionally non-generic — every
algorithm takes `string`, not `[T constraints.~string]` or `comparable`
or `[]rune`. There are two reasons.

First, **performance**: generics in Go are implemented by GC-shape
stenciling, and for tiny functions like an algorithm's inner loop,
the indirection through the generic dispatch breaks zero-allocation
fast-path budgets that the character-based algorithms rely on (see
`docs/requirements.md` §14.1 and the `performance-standards` skill).
The byte-level fast path with stack-allocated buffers is the
canonical pattern; adding a type parameter forces a heap allocation
on inputs that would otherwise fit on the stack.

Second, **dispatch**: the Scorer (Phase 8) dispatches across 23
algorithms via the typed `AlgoID` enum and an array-backed function
table. The array dispatch is O(1) with zero allocation; the same
shape with a generic `Algorithm[T]` interface would require either an
interface conversion (with its allocation cost) or a runtime
reflection step. The typed enum is intentional — see
`docs/requirements.md` §6 and the `go-coding-standards` skill.

`string` is the lowest-common-denominator type for identifier
matching, name matching, and code-vocabulary similarity. Consumers
with `[]rune` data convert to string at the boundary; consumers
with `[]byte` data can do the same. The cost is one conversion per
call, paid by the caller (who has the original data anyway), not
one conversion per algorithm in the hot loop.

## Why x/text but no other deps?

`golang.org/x/text` is the **only** non-stdlib runtime dependency in
the root `go.mod`. It is used exclusively in `Normalise` for Unicode
NFC/NFD normalisation and diacritic stripping (NFD → Remove(Mn) →
NFC).

Unicode NFC/NFD is **table-stakes** for the audit-event-taxonomy
consumer that motivated this library: matching "Müller" against
"Mueller" against "Muller" requires correct handling of combining
marks and precomposed/decomposed equivalence. Without `x/text`, the
library would have to either:

- Ship its own Unicode normalisation table (hundreds of KB, updated
  with every new Unicode revision, requiring a full Go-team-level
  maintenance burden), or
- Punt on the problem and produce wrong answers for non-ASCII inputs
  (unacceptable for an audit-data library).

`x/text` is **Go-team-maintained**, supply-chain-auditable, narrow in
scope (we use only `unicode/norm` and `runes`), and licence-clean
(BSD-3-Clause, compatible with our Apache-2.0). It was added to the
allowlist on the founding date (CLAUDE.md "Constraints" and the
project CONTEXT.md establish-pattern).

The allowlist is **locked at one entry**. Adding any other runtime
dep requires:

1. Explicit user (project owner) approval.
2. `algorithm-licensing-reviewer` sign-off (patent / licence screen).
3. A formal change to the `make verify-deps-allowlist` script (so
   subsequent PRs cannot quietly slip in additional deps).

See CLAUDE.md "Constraints" and `.claude/skills/algorithm-licensing-
standards/SKILL.md` for the canonical statement.
