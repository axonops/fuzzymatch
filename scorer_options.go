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

// scorer_options.go declares the Phase 8 functional-options surface for
// the composite weighted Scorer: the public ScorerOption type, the
// unexported scorerConfig accumulator, and the With* option functions
// that consumers pass to NewScorer (introduced in plan 08-02).
//
// Design notes (per .planning/phases/08-composite-scorer/08-CONTEXT.md
// and 08-RESEARCH.md):
//
//   - ScorerOption is `func(*scorerConfig) error`. Each option validates
//     its inputs at option-application time and returns a sentinel
//     error from errors.go on failure. The caller (NewScorer in plan
//     08-02) short-circuits on the first non-nil error so the user
//     sees the first malformed option, not a cascading list.
//
//   - scorerConfig is a slice-of-entries (NOT a map). The slice keeps
//     option-application order so NewScorer can implement last-write-
//     wins on duplicate AlgoIDs by deduplicating to a map after the
//     full option pass. WithoutAlgorithm removes ALL entries matching
//     the target AlgoID (not just the most recent) so that subsequent
//     dedup sees zero of them — this matches Pitfall 4 from
//     08-RESEARCH.md.
//
//   - The 6 non-parameterised options (Task 2) live above the 6
//     parameterised algorithm options (Task 3) for readability. Every
//     option function has a godoc block opening with the function name.
//
//   - No init(), no goroutines, no third-party imports. The file uses
//     only package-internal references (AlgoID, NormalisationOptions,
//     ErrInvalidWeight, ErrInvalidThreshold, ErrInvalidAlgoID, the
//     dispatch[] array, numAlgorithms).

package fuzzymatch

// ScorerOption configures a Scorer at construction time. Options are
// consumed by NewScorer (plan 08-02): each option is applied to an
// internal scorerConfig in order, accumulating state until NewScorer
// validates the result and freezes the *Scorer.
//
// Every option function returned by the With* / Without* constructors
// in this file validates its inputs at option-application time and
// returns one of the documented sentinel errors on failure (see
// errors.go). NewScorer short-circuits on the first non-nil error.
//
// A ScorerOption is a plain function value and may be passed,
// appended into option slices, or composed via
// `append(DefaultScorerOptions(), ...)` patterns. Options are NOT
// reusable in the sense that an option may close over consumer-
// supplied parameters (n, alpha, beta, etc.); reusing the same option
// value across multiple NewScorer calls is safe because the captured
// parameters are stored by value.
type ScorerOption func(*scorerConfig) error

// scorerEntry is the unexported representation of a single weighted
// algorithm inside a scorerConfig. Each With*Algorithm option appends a
// scorerEntry to cfg.entries; NewScorer (plan 08-02) deduplicates and
// sorts the slice into the immutable *Scorer field.
//
// id is the typed AlgoID for the algorithm (used by NewScorer's
// canonical-sort step and by the public Algorithms() accessor).
// weight is the consumer-supplied weight before normalisation.
// scoreFn is the closure that produces a [0.0, 1.0] similarity score
// for the (a, b) pair — either dispatch[id] for non-parameterised
// algorithms or a closure capturing the consumer's parameters
// (q-gram n, Tversky α/β, Monge-Elkan inner, SWGParams).
type scorerEntry struct {
	id      AlgoID
	weight  float64
	scoreFn func(a, b string) float64
}

// scorerConfig accumulates option state during the variadic
// NewScorer(opts ...ScorerOption) call. It is deliberately unexported:
// the only legal consumer is NewScorer (plan 08-02). The shape is
// frozen by plan 08-01 so that plan 08-02 can rely on the field set
// without churn.
//
// Field semantics:
//
//   - entries: appended-in-order by every With*Algorithm option; the
//     slice may contain duplicate AlgoIDs (resolved by NewScorer via
//     last-write-wins) and may contain entries later removed by
//     WithoutAlgorithm. NewScorer dedups and sorts into the immutable
//     *Scorer slice.
//
//   - threshold: the value supplied to WithThreshold. Zero by default
//     (Go zero value); the thresholdSet flag distinguishes "never set"
//     from "explicitly set to 0.0" so NewScorer can return
//     ErrMissingThreshold for the former.
//
//   - thresholdSet: true if WithThreshold was applied successfully at
//     least once. NewScorer (plan 08-02) checks this FIRST (before
//     any other validation) and returns ErrMissingThreshold if false.
//
//   - normaliseWeights: controls the sum-to-1 auto-normalisation step
//     at NewScorer time. WithNormaliseWeights(true) (default) divides
//     each entry's weight by the sum of all weights; (false) leaves
//     raw weights and waives the [0, 1] composite guarantee.
//
//   - applyNorm: when true, NewScorer's resulting Scorer.Score pre-
//     normalises BOTH inputs via Normalise(s, normOpts) before
//     dispatching to each algorithm. WithoutNormalisation() sets this
//     to false; WithNormalisation(opts) sets it to true AND stores
//     opts.
//
//   - normOpts: the NormalisationOptions used when applyNorm == true.
//     Stored by value (per .claude/skills/determinism-standards/SKILL.md
//     no-shared-mutable-state rule).
type scorerConfig struct {
	entries          []scorerEntry
	threshold        float64
	thresholdSet     bool
	normaliseWeights bool
	applyNorm        bool
	normOpts         NormalisationOptions
}

// --- Non-parameterised options ---------------------------------------------

// WithAlgorithm returns a ScorerOption that registers algo with the
// given weight using the algorithm's dispatch-table-registered default
// parameters (e.g. n = 3 for q-gram algorithms, α = β = 1.0 for
// Tversky, JaroWinkler inner for Monge-Elkan, NewSWGParams() defaults
// for Smith-Waterman-Gotoh).
//
// weight must be strictly positive — the option returns
// ErrInvalidWeight at option-application time if weight ≤ 0. algo must
// be a valid AlgoID with a populated dispatch entry (the 23 catalogue
// AlgoIDs after Phase 7 all qualify) — the option returns
// ErrInvalidAlgoID if int(algo) >= numAlgorithms or
// dispatch[algo] == nil.
//
// Consumers wanting non-default parameters call the corresponding
// parameterised option function (WithQGramJaccardAlgorithm,
// WithCosineAlgorithm, WithTverskyAlgorithm, WithMongeElkanAlgorithm,
// WithSmithWatermanGotohAlgorithm) instead.
func WithAlgorithm(algo AlgoID, weight float64) ScorerOption {
	return func(cfg *scorerConfig) error {
		if weight <= 0 {
			return ErrInvalidWeight
		}
		if int(algo) < 0 || int(algo) >= numAlgorithms || dispatch[algo] == nil {
			return ErrInvalidAlgoID
		}
		cfg.entries = append(cfg.entries, scorerEntry{
			id:      algo,
			weight:  weight,
			scoreFn: dispatch[algo],
		})
		return nil
	}
}

// WithoutAlgorithm returns a ScorerOption that removes every previously
// accumulated entry matching id from the in-flight option slice. If id
// is not currently present, WithoutAlgorithm returns nil without error
// — the no-op-on-absent semantic enables composition patterns like
//
//	opts := append(fuzzymatch.DefaultScorerOptions(),
//	    fuzzymatch.WithoutAlgorithm(fuzzymatch.AlgoDoubleMetaphone),
//	    fuzzymatch.WithThreshold(0.80),
//	)
//	s, _ := fuzzymatch.NewScorer(opts...)
//
// even when AlgoDoubleMetaphone is not part of the DefaultScorerOptions
// composition (e.g. a future minor release tweaks the default mix).
//
// All matching entries are removed — if the user inadvertently passed
// WithAlgorithm(AlgoLevenshtein, w) twice and then WithoutAlgorithm
// (AlgoLevenshtein), zero Levenshtein entries remain.
func WithoutAlgorithm(id AlgoID) ScorerOption {
	return func(cfg *scorerConfig) error {
		// Linear scan-and-compact. The option slice is typically small
		// (default Scorer has 6 entries) so the O(n) cost is
		// negligible. Iterate in reverse so removals do not shift
		// later indices.
		filtered := cfg.entries[:0]
		for _, e := range cfg.entries {
			if e.id != id {
				filtered = append(filtered, e)
			}
		}
		cfg.entries = filtered
		return nil
	}
}

// WithNormalisation returns a ScorerOption that enables pre-comparison
// normalisation in the resulting Scorer's Score / ScoreAll / Match
// methods. The Scorer applies Normalise(s, opts) to BOTH inputs once
// at the Scorer boundary; token-based algorithms (Monge-Elkan, Token*,
// Partial Ratio) continue to call Tokenise internally on the already-
// normalised string.
//
// The zero-value NormalisationOptions{} (all fields false) makes the
// underlying Normalise a no-op pass-through and is semantically
// equivalent to WithoutNormalisation(); the latter is the sugared form.
//
// Passing this option AFTER WithoutNormalisation in the same option
// slice re-enables normalisation with the supplied opts (later option
// wins).
func WithNormalisation(opts NormalisationOptions) ScorerOption {
	return func(cfg *scorerConfig) error {
		cfg.applyNorm = true
		cfg.normOpts = opts
		return nil
	}
}

// WithoutNormalisation returns a ScorerOption that disables pre-
// comparison normalisation in the resulting Scorer. With this option
// applied, Scorer.Score passes the raw inputs (no Normalise call) to
// every registered algorithm; token-based algorithms still tokenise
// the raw input via their internal Tokenise(s, DefaultTokeniseOptions())
// call.
//
// Passing this option AFTER WithNormalisation in the same option slice
// disables normalisation (later option wins); applyNorm becomes false
// but the previously-stored normOpts value is intentionally not
// cleared (cheap; harmless; allows a subsequent WithNormalisation to
// inspect-and-reuse if it wishes).
func WithoutNormalisation() ScorerOption {
	return func(cfg *scorerConfig) error {
		cfg.applyNorm = false
		return nil
	}
}

// WithThreshold returns a ScorerOption that sets the Scorer's match
// threshold to t. t must lie in the closed interval [0.0, 1.0]; values
// outside the interval produce ErrInvalidThreshold at option-application
// time.
//
// WithThreshold is MANDATORY for NewScorer: if no WithThreshold option
// is present in the variadic opts slice, NewScorer returns
// ErrMissingThreshold (see errors.go and CONTEXT.md §2 for the
// rationale — the threshold is a calibration parameter with no
// universally-safe default). DefaultScorer() bakes 0.85 in via this
// option so casual consumers using the default are unaffected.
//
// Passing this option more than once in the same option slice keeps
// the LAST value (later option wins) — a rare pattern but matches the
// general functional-options convention.
func WithThreshold(t float64) ScorerOption {
	return func(cfg *scorerConfig) error {
		if t < 0.0 || t > 1.0 {
			return ErrInvalidThreshold
		}
		cfg.threshold = t
		cfg.thresholdSet = true
		return nil
	}
}

// WithNormaliseWeights returns a ScorerOption that controls weight
// auto-normalisation at NewScorer time. When normalise == true
// (default), NewScorer divides every entry's weight by the sum of all
// weights so the composite Scorer.Score result is guaranteed to lie in
// [0.0, 1.0]. When normalise == false, raw weights are preserved and
// the [0, 1] composite guarantee is waived (consumers may want this
// when the weight values themselves carry semantic meaning, e.g.
// log-odds composites).
//
// Most consumers should leave the default true. Setting false is an
// advanced pattern documented in docs/tuning.md (plan 08-04).
func WithNormaliseWeights(normalise bool) ScorerOption {
	return func(cfg *scorerConfig) error {
		cfg.normaliseWeights = normalise
		return nil
	}
}

// --- Parameterised algorithm options ---------------------------------------

// WithQGramJaccardAlgorithm returns a ScorerOption that registers the
// Q-Gram Jaccard algorithm (Ukkonen 1992 — see docs/algorithms.md) with
// the given weight and q-gram window length n. Use this option instead
// of WithAlgorithm(AlgoQGramJaccard, weight) when n must differ from
// the dispatch-registered default of 3.
//
// weight must be strictly positive (else ErrInvalidWeight). n must be
// ≥ 1 (else ErrInvalidQGramSize) — q-gram extraction is undefined for
// window length < 1.
//
// The captured n is stored by value inside the closure, so applying
// this option to multiple NewScorer calls is safe.
func WithQGramJaccardAlgorithm(weight float64, n int) ScorerOption {
	return func(cfg *scorerConfig) error {
		if weight <= 0 {
			return ErrInvalidWeight
		}
		if n < 1 {
			return ErrInvalidQGramSize
		}
		cfg.entries = append(cfg.entries, scorerEntry{
			id:      AlgoQGramJaccard,
			weight:  weight,
			scoreFn: func(a, b string) float64 { return QGramJaccardScore(a, b, n) },
		})
		return nil
	}
}

// WithSorensenDiceAlgorithm returns a ScorerOption that registers the
// Sørensen-Dice coefficient (Sørensen 1948 / Dice 1945 — see
// docs/algorithms.md) with the given weight and q-gram window length
// n. Use this instead of WithAlgorithm(AlgoSorensenDice, weight) when
// n must differ from the dispatch default of 3.
//
// weight must be > 0 (else ErrInvalidWeight); n must be ≥ 1 (else
// ErrInvalidQGramSize).
func WithSorensenDiceAlgorithm(weight float64, n int) ScorerOption {
	return func(cfg *scorerConfig) error {
		if weight <= 0 {
			return ErrInvalidWeight
		}
		if n < 1 {
			return ErrInvalidQGramSize
		}
		cfg.entries = append(cfg.entries, scorerEntry{
			id:      AlgoSorensenDice,
			weight:  weight,
			scoreFn: func(a, b string) float64 { return SorensenDiceScore(a, b, n) },
		})
		return nil
	}
}

// WithCosineAlgorithm returns a ScorerOption that registers the Cosine
// n-gram similarity (Salton & McGill 1983 — see docs/algorithms.md)
// with the given weight and q-gram window length n. Use this instead
// of WithAlgorithm(AlgoCosine, weight) when n must differ from the
// dispatch default of 3.
//
// weight must be > 0 (else ErrInvalidWeight); n must be ≥ 1 (else
// ErrInvalidQGramSize).
func WithCosineAlgorithm(weight float64, n int) ScorerOption {
	return func(cfg *scorerConfig) error {
		if weight <= 0 {
			return ErrInvalidWeight
		}
		if n < 1 {
			return ErrInvalidQGramSize
		}
		cfg.entries = append(cfg.entries, scorerEntry{
			id:      AlgoCosine,
			weight:  weight,
			scoreFn: func(a, b string) float64 { return CosineScore(a, b, n) },
		})
		return nil
	}
}

// WithTverskyAlgorithm returns a ScorerOption that registers the
// Tversky index (Tversky 1977 — see docs/algorithms.md) with the given
// weight, asymmetric parameters alpha + beta, and q-gram window length
// n. Use this instead of WithAlgorithm(AlgoTversky, weight) when any
// of alpha, beta, or n must differ from the dispatch defaults
// (alpha = beta = 1.0, n = 3 — Jaccard-equivalent).
//
// weight must be > 0 (else ErrInvalidWeight); n must be ≥ 1 (else
// ErrInvalidQGramSize); alpha and beta must be ≥ 0 (else
// ErrInvalidTverskyParam). The α + β > 0 constraint (which guards
// the Tversky denominator) is enforced at runtime by TverskyScore
// itself; this option does not re-check it because either α or β
// being > 0 is satisfied by the typical use cases (alpha = beta = 1
// for Jaccard-equivalent, alpha = 1, beta = 0 for prototype matching).
func WithTverskyAlgorithm(weight, alpha, beta float64, n int) ScorerOption {
	return func(cfg *scorerConfig) error {
		if weight <= 0 {
			return ErrInvalidWeight
		}
		if n < 1 {
			return ErrInvalidQGramSize
		}
		if alpha < 0 || beta < 0 {
			return ErrInvalidTverskyParam
		}
		cfg.entries = append(cfg.entries, scorerEntry{
			id:      AlgoTversky,
			weight:  weight,
			scoreFn: func(a, b string) float64 { return TverskyScore(a, b, n, alpha, beta) },
		})
		return nil
	}
}

// WithMongeElkanAlgorithm returns a ScorerOption that registers the
// Monge-Elkan symmetric default (Monge & Elkan 1996 — see
// docs/algorithms.md) with the given weight and inner-metric AlgoID.
// Use this instead of WithAlgorithm(AlgoMongeElkan, weight) when the
// inner metric must differ from the dispatch default of AlgoJaroWinkler.
//
// weight must be > 0 (else ErrInvalidWeight); inner must be a
// permitted Monge-Elkan inner AlgoID with a populated dispatch entry
// (else ErrInvalidAlgoID). The trivial-recursion case inner ==
// AlgoMongeElkan is rejected explicitly here so the consumer sees a
// typed error at construction time instead of a runtime panic from
// MongeElkanScore's allow-list gate.
//
// The full 18-entry inner allow-list is enforced inside
// MongeElkanScore (Phase 6 + Phase 7 locked behaviour) — this option
// performs only the bounds + self-rejection check. Passing an inner
// AlgoID that the underlying ME implementation rejects will panic at
// Score time (programmer error); the panic surfaces via godog's
// recover mechanism in plan 08-04's BDD scenarios.
//
// The captured inner is stored by value inside the closure. Phase 8.5
// Q3 — the previously-inert NormalisationOptions parameter has been
// removed from MongeElkanScore; the closure now passes only (a, b, inner).
func WithMongeElkanAlgorithm(weight float64, inner AlgoID) ScorerOption {
	return func(cfg *scorerConfig) error {
		if weight <= 0 {
			return ErrInvalidWeight
		}
		if int(inner) < 0 || int(inner) >= numAlgorithms || dispatch[inner] == nil {
			return ErrInvalidAlgoID
		}
		if inner == AlgoMongeElkan {
			// Trivial recursion guard — see godoc above.
			return ErrInvalidAlgoID
		}
		cfg.entries = append(cfg.entries, scorerEntry{
			id:     AlgoMongeElkan,
			weight: weight,
			scoreFn: func(a, b string) float64 {
				return MongeElkanScore(a, b, inner)
			},
		})
		return nil
	}
}

// WithSmithWatermanGotohAlgorithm returns a ScorerOption that
// registers the Smith-Waterman-Gotoh local-alignment similarity
// (Smith & Waterman 1981 + Gotoh 1982 — see docs/algorithms.md) with
// the given weight and affine-gap parameters. Use this instead of
// WithAlgorithm(AlgoSmithWatermanGotoh, weight) when the params must
// differ from NewSWGParams() defaults.
//
// weight must be > 0 (else ErrInvalidWeight). params validation is
// the responsibility of SmithWatermanGotohScoreWithParams itself:
// nonsense values produce a deterministic-but-meaningless score (per
// the documented contract on SWGParams) rather than an error. The
// Scorer layer therefore does not pre-validate params.
//
// The captured params struct is stored by value inside the closure
// (SWGParams is a plain value type per Phase 3); the consumer may
// freely mutate their local SWGParams variable after this call
// without affecting the registered closure.
func WithSmithWatermanGotohAlgorithm(weight float64, params SWGParams) ScorerOption {
	return func(cfg *scorerConfig) error {
		if weight <= 0 {
			return ErrInvalidWeight
		}
		cfg.entries = append(cfg.entries, scorerEntry{
			id:     AlgoSmithWatermanGotoh,
			weight: weight,
			scoreFn: func(a, b string) float64 {
				return SmithWatermanGotohScoreWithParams(a, b, params)
			},
		})
		return nil
	}
}
