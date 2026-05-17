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
//     copy).

package fuzzymatch

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
//     (ErrInvalidAlgorithm) — the option layer already gates on these
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
func NewScorer(opts ...ScorerOption) (*Scorer, error) {
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
			return nil, ErrInvalidAlgorithm
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
			sum = sum + sorted[i].weight
		}
		if sum == 0 {
			return nil, ErrInvalidWeight
		}
		for i := range sorted {
			sorted[i].weight = sorted[i].weight / sum
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
