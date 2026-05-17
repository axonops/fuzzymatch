// Copyright 2026 AxonOps Limited
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

// scorer.go declares the Phase 8 composite weighted Scorer — the
// Layer 2 surface of the three-layer fuzzymatch architecture. It
// consumes the functional-options surface from scorer_options.go
// (plan 08-01), validates and freezes the resulting *Scorer at
// construction time, and exposes the Score / Match methods that
// downstream consumers use to compute composite similarity.
//
// Plan 08-02 lands NewScorer + Score + Match. The remaining surface
// (ScoreAll, Threshold accessor, Algorithms accessor, DefaultScorer,
// DefaultScorerOptions) lands in plan 08-03. Golden file + BDD + docs
// finalisation lands in plan 08-04.
//
// Design notes (per .planning/phases/08-composite-scorer/08-CONTEXT.md
// and 08-RESEARCH.md):
//
//   - Validation pipeline order is LOCKED (CONTEXT.md §2 + Pitfall 3):
//     missing-threshold FIRST → empty-algorithms → defensive AlgoID
//     bounds (option layer already validates; NewScorer re-checks for
//     defence-in-depth) → weight normalisation.
//
//   - Duplicate WithAlgorithm(SameAlgoID, _) is resolved by
//     last-write-wins per Pitfall 4: NewScorer dedups via a
//     map[AlgoID]scorerEntry pass that overwrites earlier entries
//     with later ones. The post-dedup slice is then iterated in
//     canonical AlgoIDs() order so the resulting algorithmsAlgoIDSorted
//     field is AlgoID-ascending. The reduction loop's iteration order
//     is therefore deterministic across runs, processes, and platforms.
//
//   - Weight normalisation: when WithNormaliseWeights(true) (default),
//     each entry's weight is divided by the left-to-right sum. The
//     division is the only float arithmetic at construction time; the
//     resulting weights are stable across platforms because IEEE-754
//     division is correctly rounded.
//
//   - Float-determinism reduction (CONTEXT.md §5 LOCKED): the inner
//     loop of Score uses the canonical Phase 5 Cosine pattern (see
//     cosine.go:341-344): explicit `(entry.weight * score)`
//     parenthesisation, left-to-right accumulation
//     (acc = acc + (entry.weight * score)), AlgoID-sorted iteration
//     order. No math.Pow, math.Log, math.Exp, math.FMA, or parallel
//     reductions. Per the FMA-fusion caveat documented at
//     cosine.go:288-297, parentheses do NOT defeat FMA on arm64; the
//     empirical observation is that score × weight products lie in
//     [0, 1] and stay below the byte-diff threshold of the
//     cross-platform golden gate. The fallback remediation (explicit
//     double float64 cast) is documented in cosine.go and applies
//     verbatim to this loop if matrix divergence ever appears.
//
//   - Concurrency: the *Scorer is immutable after NewScorer returns.
//     All fields are set once and never written again, so Score / Match
//     are safe for concurrent use from multiple goroutines without any
//     synchronisation. No sync.Mutex, no atomic ops, no goroutines.
//
//   - No init(); no goroutines; no third-party imports. The only stdlib
//     dependency is "sort" (not even strictly necessary at this plan
//     boundary — AlgoIDs() returns the canonical order — but reserved
//     for plan 08-03's Algorithms() accessor which sorts a fresh slice
//     copy). Phase 8.5 adds "fmt" for the typed-panic wrapping at
//     DefaultScorer (ErrInternalInvariantViolated per Gap 5).

package fuzzymatch

import "fmt"

// Scorer is the immutable composite weighted scorer. A Scorer is
// constructed via NewScorer(opts ...ScorerOption) — once constructed,
// all fields are read-only and the methods (Score, Match, and
// plan-08-03's ScoreAll / Threshold / Algorithms) are safe for
// concurrent use from any number of goroutines without external
// synchronisation.
//
// The zero-value Scorer is NOT usable. Always obtain a *Scorer from
// NewScorer or DefaultScorer (the latter lands in plan 08-03).
//
// The exported field set is intentionally empty: every field is
// unexported so the v1.x API surface preserves the freedom to evolve
// internals (e.g. add allocation-pooling, ScoreAll fast paths) without
// breaking consumers. Accessors land in plan 08-03 (Threshold(),
// Algorithms()).
type Scorer struct {
	// algorithmsAlgoIDSorted is the final, dedup'd, AlgoID-ascending
	// slice of weighted algorithm entries. Built once in NewScorer by
	// (a) iterating cfg.entries to dedup via a map[AlgoID]scorerEntry
	// (last-write-wins), (b) materialising in canonical AlgoIDs() order,
	// and (c) optionally normalising weights to sum-to-1. The slice is
	// read-only after NewScorer returns.
	//
	// The reduction loop in Score iterates this slice directly — no
	// per-call sort, no map iteration. AlgoID-ascending order makes the
	// reduction order deterministic across runs, processes, and
	// platforms (DET-04 / CONTEXT.md §5 LOCKED).
	algorithmsAlgoIDSorted []scorerEntry

	// threshold is the match boundary stored from WithThreshold. Match
	// returns Score(a, b) >= threshold (boundary inclusive). Threshold
	// is required at construction time per CONTEXT.md §2 LOCKED —
	// NewScorer returns ErrMissingThreshold if no WithThreshold option
	// was applied.
	threshold float64

	// applyNormalisation gates Score's pre-comparison Normalise calls.
	// When true (default), Score normalises both inputs once at the
	// Scorer boundary before dispatching to each algorithm. When false
	// (consumer set WithoutNormalisation), Score passes raw inputs to
	// every algorithm; token-based algorithms still tokenise the raw
	// input via their internal Tokenise call.
	applyNormalisation bool

	// normaliseOpts is the NormalisationOptions passed to Normalise when
	// applyNormalisation is true. Stored by value (not by reference) so
	// the consumer cannot mutate it through an aliased struct after
	// NewScorer returns.
	normaliseOpts NormalisationOptions
}

// NewScorer constructs a composite weighted Scorer from the supplied
// options. The variadic opts are applied in order; the first option
// that returns a non-nil error short-circuits the constructor (the
// consumer sees the first malformed option, not a cascading list).
//
// After every option has been applied successfully, NewScorer runs the
// validation pipeline in the order LOCKED by CONTEXT.md §2 (mirroring
// 08-RESEARCH.md Pitfall 3):
//
//  1. Missing-threshold (ErrMissingThreshold) — fires FIRST so a user
//     who forgets WithThreshold AND has another option problem still
//     sees a clear "you forgot the threshold" message.
//  2. Empty-algorithms (ErrEmptyScorer) — at least one WithAlgorithm
//     (or any With*Algorithm) option must have been applied.
//  3. Defensive per-entry AlgoID bounds + dispatch nil-check
//     (ErrInvalidAlgoID) — the option layer already gates on these
//     conditions; NewScorer re-validates for defence-in-depth so the
//     final *Scorer's contract is bulletproof regardless of how the
//     option slice was built.
//
// After validation, NewScorer dedups duplicate AlgoIDs by last-write-
// wins (Pitfall 4 / 08-RESEARCH.md "Pattern 1") and sorts the result
// into AlgoID-ascending order via the canonical AlgoIDs() iteration.
// If WithNormaliseWeights(true) (the default), each entry's weight is
// divided by the left-to-right sum of all weights so the resulting
// Scorer.Score result lies in [0.0, 1.0] for any input. When
// WithNormaliseWeights(false), raw weights are preserved (the [0, 1]
// composite guarantee is then waived; the consumer takes responsibility
// for the weight semantics).
//
// The returned *Scorer is immutable after NewScorer returns. All
// fields are set once and never written again; Score / Match (and
// plan-08-03's ScoreAll / Threshold / Algorithms) are safe for
// concurrent use without external synchronisation. No mutex, no atomic
// ops, no goroutines.
//
// On any error, NewScorer returns (nil, err) and never allocates a
// partially-constructed *Scorer.
//
// Typical usage:
//
//	s, err := fuzzymatch.NewScorer(
//	    fuzzymatch.WithAlgorithm(fuzzymatch.AlgoLevenshtein, 0.6),
//	    fuzzymatch.WithAlgorithm(fuzzymatch.AlgoJaroWinkler, 0.4),
//	    fuzzymatch.WithThreshold(0.75),
//	)
//	if err != nil {
//	    return fmt.Errorf("build scorer: %w", err)
//	}
//
// For the opinionated default 6-algorithm composition with a baked-in
// threshold of 0.85, use DefaultScorer() (lands in plan 08-03) which
// cannot fail.
func NewScorer(opts ...ScorerOption) (*Scorer, error) { //nolint:gocyclo // 9-step LOCKED validation pipeline (CONTEXT.md §2 + 08-RESEARCH.md Pitfall 3); each conditional is a documented sentinel gate that cannot be folded into a sub-helper without splitting the locked-order contract across files
	// Step 1 — initialise the accumulator with the documented defaults
	// BEFORE applying options. Options that don't touch normalisation
	// (WithAlgorithm, WithThreshold, …) leave applyNorm = true and
	// normaliseWeights = true so the Scorer's behaviour matches the
	// "sensible defaults" half of spec §9.4. Setting these defaults
	// here (not at the scorerConfig zero-value level) keeps
	// scorerConfig itself a plain value type with no init-time work.
	cfg := scorerConfig{
		normaliseWeights: true,
		applyNorm:        true,
		normOpts:         DefaultNormalisationOptions(),
	}

	// Step 2 — apply options in supplied order. Short-circuit on the
	// first non-nil error so the consumer sees the first malformed
	// option (not a cascading list). Each option validates its inputs
	// at option-application time and returns the appropriate sentinel
	// from errors.go.
	for _, opt := range opts {
		if err := opt(&cfg); err != nil {
			return nil, err
		}
	}

	// Step 3 — missing-threshold check FIRST per CONTEXT.md §2 +
	// 08-RESEARCH.md Pitfall 3. A consumer who calls NewScorer with no
	// arguments at all, or with algorithms but no WithThreshold, gets
	// ErrMissingThreshold (not ErrEmptyScorer) — this disambiguates the
	// "you forgot the threshold" diagnostic from "you forgot algorithms."
	if !cfg.thresholdSet {
		return nil, ErrMissingThreshold
	}

	// Step 4 — empty-algorithms check. Distinct from missing-threshold:
	// a consumer who passes WithThreshold but no algorithms gets
	// ErrEmptyScorer.
	if len(cfg.entries) == 0 {
		return nil, ErrEmptyScorer
	}

	// Step 5 — defensive per-entry AlgoID bounds + dispatch nil-check.
	// The option layer (scorer_options.go) already gates on these
	// conditions, so this loop is defence-in-depth: it ensures the
	// final *Scorer's invariants hold regardless of how the option
	// slice was assembled (e.g. a future caller building options via
	// reflection or a config-file decoder). The dispatch nil-check
	// guarantees the reduction loop in Score never panics on a nil
	// function-pointer dereference.
	for _, e := range cfg.entries {
		if int(e.id) < 0 || int(e.id) >= numAlgorithms || dispatch[e.id] == nil {
			return nil, ErrInvalidAlgoID
		}
	}

	// Step 6 — dedup via map[AlgoID]scorerEntry. Iterating cfg.entries
	// in order and assigning seen[e.id] = e gives last-write-wins
	// semantics per 08-RESEARCH.md Pitfall 4: a later
	// WithAlgorithm(SameID, w) overwrites the earlier weight (and
	// scoreFn — parameterised algorithm options can also re-register
	// the same AlgoID with different parameters; the LAST option wins).
	seen := make(map[AlgoID]scorerEntry, len(cfg.entries))
	for _, e := range cfg.entries {
		seen[e.id] = e
	}

	// Step 7 — materialise the deduped entries into an AlgoID-ascending
	// slice. Iterating AlgoIDs() (canonical iota order from
	// algoid.go:282-308) and skipping missing entries produces a slice
	// whose order is determined by the AlgoID enum, not by the option
	// application order. This is the load-bearing invariant for the
	// float-determinism reduction loop in Score (CONTEXT.md §5 LOCKED).
	sorted := make([]scorerEntry, 0, len(seen))
	for _, id := range AlgoIDs() {
		if e, ok := seen[id]; ok {
			sorted = append(sorted, e)
		}
	}

	// Step 8 — weight auto-normalisation. When cfg.normaliseWeights is
	// true (the default), divide every entry's weight by the left-to-
	// right sum so the post-normalisation weights sum to 1.0 and the
	// composite Score result is guaranteed to lie in [0.0, 1.0]. The
	// sum is computed as `sum = sum + sorted[i].weight` (explicit left-
	// to-right per DET-06) so the divisor is deterministic across
	// platforms.
	//
	// The defensive sum == 0 check should be unreachable: the option
	// layer rejects weight <= 0 at option-application time, so any
	// surviving entry has weight > 0 and the sum of N > 0 entries is
	// > 0. Kept as a paranoid guard so the division never produces
	// NaN/Inf — returning ErrInvalidWeight surfaces the contract
	// violation as a typed error rather than silently propagating a
	// poisoned float.
	if cfg.normaliseWeights {
		var sum float64
		for i := range sorted {
			// Explicit `sum = sum + …` per DET-06 / CONTEXT.md §5 — the
			// left-to-right additive accumulation is the locked
			// determinism pattern from cosine.go:343. The +=
			// shorthand is observationally equivalent but the
			// explicit form is the contract.
			sum = sum + sorted[i].weight //nolint:gocritic // DET-06 locked left-to-right additive accumulation pattern; explicit form is the contract per CONTEXT.md §5
		}
		if sum == 0 {
			return nil, ErrInvalidWeight
		}
		for i := range sorted {
			// Explicit `x = x / sum` for symmetry with the sum
			// reduction above. The /= shorthand is observationally
			// equivalent but the explicit-arithmetic form keeps the
			// determinism-discipline visible at the read site.
			sorted[i].weight = sorted[i].weight / sum //nolint:gocritic // explicit-arithmetic form per DET-06 / CONTEXT.md §5 locked discipline; symmetric with the sum reduction above
		}
	}

	// Step 9 — freeze the Scorer. All fields are set once here and
	// never written again; the struct is immutable and concurrent-safe
	// for the lifetime of the returned pointer.
	return &Scorer{
		algorithmsAlgoIDSorted: sorted,
		threshold:              cfg.threshold,
		applyNormalisation:     cfg.applyNorm,
		normaliseOpts:          cfg.normOpts,
	}, nil
}

// Score returns the composite similarity score in [0.0, 1.0] for the
// pair (a, b). The score is the weight-normalised sum of every
// registered algorithm's per-pair score evaluated against the
// (optionally normalised) inputs.
//
// Pre-normalisation policy (CONTEXT.md §3 LOCKED): when the Scorer was
// constructed with normalisation enabled (the default, or
// WithNormalisation(opts)), Score applies Normalise(s, normaliseOpts)
// to BOTH a and b ONCE at the Scorer boundary before dispatching to
// each algorithm. Token-based algorithms (Monge-Elkan, Token*,
// PartialRatio) continue to call Tokenise internally on the already-
// normalised string — Phase 8 does not modify any token-based
// algorithm's internals. When the Scorer was constructed with
// WithoutNormalisation(), Score passes the raw inputs to every
// algorithm.
//
// Determinism guarantee: Score returns the same float64 to the last
// bit on every call with the same Scorer configuration and the same
// (a, b) inputs, across runs, processes, and CI-matrix platforms
// (linux/amd64, linux/arm64, darwin/amd64, darwin/arm64,
// windows/amd64). This is enforced by AlgoID-sorted internal iteration
// (the slice was sorted at NewScorer time; iteration order does NOT
// depend on option-application order) and by the explicit
// (entry.weight * score) parenthesisation in the reduction loop,
// matching the Phase 5 Cosine precedent at cosine.go:341-344. Per the
// FMA-fusion caveat documented at cosine.go:288-297, parentheses do
// not defeat FMA on arm64; the empirical observation is that score *
// weight products in [0, 1] stay below the byte-diff threshold of the
// cross-platform golden gate.
//
// Range guarantee: with default weight normalisation
// (WithNormaliseWeights(true), the default) and underlying algorithms
// that satisfy score ∈ [0.0, 1.0] (all 23 catalogue algorithms do),
// the returned value lies in [0.0, 1.0]. When the Scorer was
// constructed with WithNormaliseWeights(false), raw weights are
// preserved and the [0, 1] range guarantee is waived — the caller
// takes responsibility for the weight semantics.
//
// Concurrency: Score is safe for concurrent use from any number of
// goroutines on the same *Scorer without external synchronisation. The
// Scorer is immutable after NewScorer returns; this method does no
// writes to the receiver's state.
func (s *Scorer) Score(a, b string) float64 {
	// Pre-normalisation boundary (CONTEXT.md §3 LOCKED). Single
	// conditional, two Normalise calls when active — avoids the
	// function-call overhead when WithoutNormalisation was applied.
	na, nb := a, b
	if s.applyNormalisation {
		na = Normalise(a, s.normaliseOpts)
		nb = Normalise(b, s.normaliseOpts)
	}

	// Float-determinism reduction loop. The canonical Phase 5 pattern
	// from cosine.go:343, lifted to the Scorer's per-algorithm
	// composite: outer float64(...) wrap on the multiplication
	// product, left-to-right accumulation
	// (acc = float64(entry.weight*score) + acc), AlgoID-sorted
	// iteration order (the slice was sorted at NewScorer time).
	// DET-06 explicit parens plus Q11b FMA-defence — see
	// docs/requirements.md §14.4 and CONTEXT.md §5 LOCKED.
	var acc float64
	for _, entry := range s.algorithmsAlgoIDSorted {
		score := entry.scoreFn(na, nb)
		// FMA-defeating double-cast (Q11b LOCKED —
		// docs/requirements.md §14.4). The outer float64(...) cast on
		// the multiplication product forces an IEEE-754 round-to-
		// nearest-even at float64 precision, which the Go compiler
		// treats as a rounding fence and therefore cannot fuse into
		// FMA (fused multiply-add). On arm64, where the FPU emits
		// FMA by default for `(a*b)+c` patterns, this defence is
		// load-bearing for cross-platform byte-identical output on
		// testdata/golden/scorer-default.json. golang/go#17895
		// documents why parenthesisation alone is not sufficient.
		// Mirror site: cosine.go:343 (dot-product reduction). The
		// reduction is left-to-right, AlgoID-sorted, and uses only +
		// and * — no transcendentals, no FMA.
		acc = float64(entry.weight*score) + acc //nolint:gocritic // DET-06 + Q11b FMA-defence locked pattern mirroring cosine.go:343 / docs/requirements.md §14.4
	}
	return acc
}

// Match returns true when the composite Score(a, b) is at or above the
// threshold supplied to WithThreshold during NewScorer. The comparison
// is `>=` so the boundary is inclusive — a Scorer with threshold 0.85
// matches inputs whose composite score is exactly 0.85.
//
// Match is a thin wrapper around Score; the same determinism and
// concurrency guarantees apply.
func (s *Scorer) Match(a, b string) bool {
	return s.Score(a, b) >= s.threshold
}

// ScorerAlgorithm describes a single weighted algorithm in a Scorer's
// configured set. It is returned by Scorer.Algorithms() as a fresh
// slice on every call so consumers can introspect, log, or display the
// Scorer's composition without coupling to unexported internals.
//
// ID is the typed AlgoID enum value for the algorithm; use ID.String()
// for the canonical snake-free CamelCase display name. Weight is the
// POST-normalisation weight that the Scorer actually uses during
// Score's reduction loop — i.e. the value AFTER NewScorer applied the
// sum-to-1 step (when WithNormaliseWeights(true), the default). When
// WithNormaliseWeights(false) was applied, Weight reflects the raw
// consumer-supplied weight unchanged.
//
// ScorerAlgorithm has no behavioural methods; it is a pure data holder
// exported so the Scorer's composition is introspectable. Consumers
// must not rely on ScorerAlgorithm having a stable memory address
// across Algorithms() calls — fresh slices imply fresh element backing
// storage on every invocation.
type ScorerAlgorithm struct {
	// ID is the typed AlgoID enum value for the algorithm. Use
	// ID.String() to obtain the canonical CamelCase display name (e.g.
	// AlgoLevenshtein → "Levenshtein", AlgoDamerauLevenshteinOSA →
	// "DamerauLevenshteinOSA").
	ID AlgoID

	// Weight is the post-normalisation weight that the Scorer actually
	// uses during Score's reduction loop. When WithNormaliseWeights(true)
	// (the default), the weights of all ScorerAlgorithm entries returned
	// by Algorithms() sum to exactly 1.0 (within IEEE-754 representation
	// — see scorer_internal_test.go TestScorer_WeightNormalisation_
	// SumsToOne for the dyadic-friendly case). When
	// WithNormaliseWeights(false), Weight is the raw consumer-supplied
	// weight unchanged.
	Weight float64
}

// Threshold returns the match boundary stored at construction time
// (the value passed to WithThreshold during NewScorer, or 0.85 when
// the Scorer was constructed via DefaultScorer). The returned value is
// in [0.0, 1.0] — NewScorer's validation pipeline rejected any
// out-of-range threshold at construction.
//
// Threshold is a plain accessor with no side effects; it is safe for
// concurrent use from any number of goroutines without external
// synchronisation.
func (s *Scorer) Threshold() float64 {
	return s.threshold
}

// Algorithms returns the configured weighted algorithm set as a fresh
// slice of ScorerAlgorithm values. The slice is freshly allocated on
// every call so consumers may freely mutate, sort, or filter it
// without affecting subsequent calls or the Scorer's internal state.
//
// The returned slice is in AlgoID-ascending order — the same iteration
// order Score uses for its float-determinism reduction loop. Each
// element's Weight is the POST-normalisation weight (consumers see
// what the Score computation actually uses; raw weights are not
// exposed through this surface).
//
// Algorithms is safe for concurrent use from any number of goroutines
// without external synchronisation. The Scorer is immutable after
// NewScorer returns; this method only reads from the receiver's state
// to build the fresh return slice.
func (s *Scorer) Algorithms() []ScorerAlgorithm {
	out := make([]ScorerAlgorithm, 0, len(s.algorithmsAlgoIDSorted))
	for _, e := range s.algorithmsAlgoIDSorted {
		out = append(out, ScorerAlgorithm{ID: e.id, Weight: e.weight})
	}
	return out
}

// ScoreAll returns per-algorithm raw scores for the configured algorithm set as a map[AlgoID]float64.
//
// SPEC OVERRIDE: docs/requirements.md §8.3 specifies map[string]float64; this implementation returns map[AlgoID]float64 because AlgoID is a typed enum that the rest of the library exposes, giving consumers compile-time key safety. Use AlgoID.String() for snake_case display. The spec deviation is documented in CONTEXT.md §1 (Phase 8) and api-ergonomics-reviewer signed off on this override in plan 08-03's PR.
//
// Map iteration order is non-deterministic per Go map semantics. Map CONTENTS are deterministic byte-for-byte (per-algorithm scores are deterministic; see PropScorer_DeterministicAcrossRuns). Consumers requiring stable iteration order MUST sort the keys themselves — typically via fuzzymatch.AlgoIDs() then key-lookup.
//
// A fresh map is allocated on every call (spec §8.6). Hot-path callers wanting to avoid the allocation should use Score(a, b) which returns the composite float without per-algorithm breakdown.
//
// Pre-normalisation policy mirrors Score (CONTEXT.md §3 LOCKED): when
// the Scorer was constructed with normalisation enabled (the default,
// or WithNormalisation(opts)), ScoreAll applies Normalise(s,
// normaliseOpts) to BOTH a and b ONCE before dispatching to each
// algorithm. This ensures the value in result[X] equals the value X
// would contribute to Score's reduction (modulo the multiplication by
// the X-entry's weight). When the Scorer was constructed with
// WithoutNormalisation(), ScoreAll passes raw inputs to every
// algorithm.
//
// Implementation: ScoreAll iterates s.algorithmsAlgoIDSorted (a slice,
// not a map) to populate the result map — per the no-map-iteration
// rule from .claude/skills/determinism-standards, output paths must
// never depend on Go map iteration order. The result map is populated
// in AlgoID-ascending order internally; what becomes non-deterministic
// is only consumer-side range iteration over the returned map.
//
// ScoreAll is safe for concurrent use from any number of goroutines
// without external synchronisation. The Scorer is immutable after
// NewScorer returns; this method does no writes to the receiver's
// state.
func (s *Scorer) ScoreAll(a, b string) map[AlgoID]float64 {
	// Pre-normalisation boundary — identical to Score (CONTEXT.md §3
	// LOCKED). Single conditional, two Normalise calls when active so
	// the per-algorithm values in the returned map match the Score
	// reduction's per-algorithm contributions byte-for-byte.
	na, nb := a, b
	if s.applyNormalisation {
		na = Normalise(a, s.normaliseOpts)
		nb = Normalise(b, s.normaliseOpts)
	}

	// Iterate the AlgoID-sorted SLICE (never the map). Writing into a
	// freshly-allocated map preserves the no-map-iteration rule on the
	// output path: consumers see a map whose contents are deterministic
	// even though Go's range iteration order over a map is randomised.
	out := make(map[AlgoID]float64, len(s.algorithmsAlgoIDSorted))
	for _, entry := range s.algorithmsAlgoIDSorted {
		out[entry.id] = entry.scoreFn(na, nb)
	}
	return out
}

// DefaultScorerOptions returns a fresh, mutable slice of ScorerOption
// matching DefaultScorer's composition: six algorithms at equal raw
// weight (DamerauLevenshteinOSA, JaroWinkler, TokenJaccard, QGramJaccard,
// SorensenDice, DoubleMetaphone — per spec §8.5 / CONTEXT.md §6) plus
// WithThreshold(0.85).
//
// Consumers can append additional options or splice in WithoutAlgorithm
// to derive customised Scorers from the default:
//
//	opts := append(fuzzymatch.DefaultScorerOptions(),
//	    fuzzymatch.WithoutAlgorithm(fuzzymatch.AlgoDoubleMetaphone),
//	    fuzzymatch.WithThreshold(0.80),  // override the default
//	)
//	s, err := fuzzymatch.NewScorer(opts...)
//
// The slice is freshly allocated on every call so consumers may mutate
// it without affecting subsequent DefaultScorer() or
// DefaultScorerOptions() calls. The weight values are raw (1.0 each);
// NewScorer's auto-normalisation step divides each by the sum so the
// post-normalisation weights are 1.0/6.0 each.
//
// Safe for concurrent use from any number of goroutines: the function
// constructs and returns a fresh slice on every call with no shared
// state.
func DefaultScorerOptions() []ScorerOption {
	return []ScorerOption{
		WithAlgorithm(AlgoDamerauLevenshteinOSA, 1.0),
		WithAlgorithm(AlgoJaroWinkler, 1.0),
		WithAlgorithm(AlgoTokenJaccard, 1.0),
		WithAlgorithm(AlgoQGramJaccard, 1.0),    // uses default n=3 from dispatch_qgram_jaccard.go
		WithAlgorithm(AlgoSorensenDice, 1.0),    // uses default n=3 from dispatch_sorensen_dice.go
		WithAlgorithm(AlgoDoubleMetaphone, 1.0), //
		WithThreshold(0.85),
	}
}

// DefaultScorer returns the opinionated default Scorer: six algorithms
// at equal weight (per spec §8.5 / CONTEXT.md §6) — DamerauLevenshtein
// OSA, JaroWinkler, TokenJaccard, QGramJaccard, SorensenDice,
// DoubleMetaphone — plus the baked-in threshold 0.85.
//
// DefaultScorer cannot fail under normal operation: the six algorithms
// are guaranteed present in the dispatch table (Phase 7 populated all
// 23 slots), the weights are positive (1.0 each), and the threshold
// lies in [0.0, 1.0]. The implementation panics only on internal
// inconsistency (a programmer error that should be unreachable after
// Phase 7); this protects consumers from silently using a misconfigured
// Scorer in production.
//
// The returned *Scorer is immutable and safe for concurrent use without
// external synchronisation. Consumers wanting "default minus algorithm
// X" use:
//
//	opts := append(fuzzymatch.DefaultScorerOptions(),
//	    fuzzymatch.WithoutAlgorithm(fuzzymatch.AlgoDoubleMetaphone),
//	)
//	s, err := fuzzymatch.NewScorer(opts...)
//
// Consumers wanting a custom composition entirely should pass
// individual WithAlgorithm options to NewScorer directly.
//
// Typical usage:
//
//	s := fuzzymatch.DefaultScorer()
//	if s.Match("user_id", "userId") {
//	    // similar
//	}
func DefaultScorer() *Scorer {
	return mustDefaultScorer(NewScorer)
}

// mustDefaultScorer is the testable internal helper for DefaultScorer.
// It calls newScorer with DefaultScorerOptions and panics with an
// ErrInternalInvariantViolated-wrapped error if construction fails.
//
// Why this exists as a separate function: the panic body is
// defence-in-depth — DefaultScorerOptions() always returns valid
// options, so the err != nil branch is unreachable via the public API.
// Extracting the panic into an unexported helper lets internal tests
// exercise the panic path (by passing a deliberately-failing newScorer
// stub) without exposing a public surface that consumers could misuse.
//
// Phase 8.5 Gap 5: the panic wraps ErrInternalInvariantViolated so
// consumers can discriminate library bugs from caller errors via:
//
//	defer func() {
//	    if r := recover(); r != nil {
//	        if e, ok := r.(error); ok && errors.Is(e, ErrInternalInvariantViolated) { … }
//	    }
//	}()
//
// The double-%w form chains the underlying validation error so
// errors.Is also matches whichever NewScorer sentinel fired.
func mustDefaultScorer(newScorer func(opts ...ScorerOption) (*Scorer, error)) *Scorer {
	s, err := newScorer(DefaultScorerOptions()...)
	if err != nil {
		panic(fmt.Errorf("%w: DefaultScorer construction failed: %w", ErrInternalInvariantViolated, err))
	}
	return s
}
