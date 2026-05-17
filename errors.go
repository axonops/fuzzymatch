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

// Package-level sentinel errors for fuzzymatch. All errors are wrappable
// via fmt.Errorf("...: %w", ErrX) and discoverable by errors.Is /
// errors.As — never via string matching. See docs/requirements.md §6
// (canonical sentinel set) and §6.A (data-vs-parameter validation
// framework + panic policy).
//
// Error message convention (per .claude/skills/go-coding-standards):
//
//   - Every message starts with the "fuzzymatch: " package prefix so
//     wrappers like fmt.Errorf("scorer: %w", err) produce readable
//     compositions ("scorer: fuzzymatch: invalid algorithm identifier").
//   - The text after the prefix is lowercase and carries no trailing
//     punctuation ('.', '!', or '?') so concatenation flows cleanly.
//   - Each sentinel is a flat errors.New value, not a typed struct;
//     richer per-item context can be added in a later phase if scan
//     or extract needs it.
//
// The Phase 8.5 v1.0 sentinel contract (per docs/requirements.md §6):
// ErrInvalidAlgoID, ErrInvalidInnerAlgo, ErrInvalidQGramSize,
// ErrInvalidTverskyParam, ErrEmptyScorer, ErrInvalidWeight,
// ErrInvalidThreshold, ErrMissingThreshold, ErrInternalInvariantViolated.

package fuzzymatch

import "errors"

// ErrInvalidQGramSize indicates a q-gram-based algorithm option was
// constructed with n < 1 — q-gram extraction requires a positive window
// length and the formulas (Jaccard, Dice, Cosine, Tversky) are undefined
// for empty q-gram sets produced by a sub-unit window.
//
// Common causes: passing n = 0 to WithQGramJaccardAlgorithm /
// WithSorensenDiceAlgorithm / WithCosineAlgorithm / WithTverskyAlgorithm
// in the mistaken belief that n = 0 selects a sensible default; passing
// a computed window size that the caller failed to clamp to ≥ 1.
//
// Resolution: pass n ≥ 1. The dispatch default for the q-gram tier is
// n = 3 (the trigram convention from Ukkonen 1992). Direct algorithm
// calls (QGramJaccardScore, SorensenDiceScore, CosineScore, TverskyScore)
// panic with a value wrapping this sentinel — per docs/requirements.md
// §6.A direct calls fail loudly on programmer error and the Scorer
// option layer returns the typed error.
//
// Example:
//
//	_, err := fuzzymatch.NewScorer(
//	    fuzzymatch.WithQGramJaccardAlgorithm(1.0, 3),
//	    fuzzymatch.WithThreshold(0.85),
//	)
//	if errors.Is(err, fuzzymatch.ErrInvalidQGramSize) {
//	    // diagnostic
//	}
var ErrInvalidQGramSize = errors.New("fuzzymatch: invalid q-gram size")

// ErrInvalidTverskyParam indicates a Tversky algorithm option was
// constructed with an invalid α/β parameter pair: α < 0, β < 0, α + β
// ≤ 0, or any NaN/Inf value. The Tversky formula requires non-negative
// weights with at least one strictly positive so the denominator does
// not collapse to zero.
//
// Common causes: passing α = β = 0 in the mistaken belief that a zero
// pair selects a sensible default; passing a NaN computed by an upstream
// arithmetic step; passing a negative tuning parameter inadvertently.
//
// Resolution: pass α ≥ 0 and β ≥ 0 with α + β > 0; the canonical
// Jaccard-equivalent is α = β = 1.0 and the prototype-matching default
// is α = 1.0, β = 0.0. Direct calls to TverskyScore / TverskyScoreRunes
// panic with a value wrapping this sentinel on the same inputs.
//
// Example:
//
//	_, err := fuzzymatch.NewScorer(
//	    fuzzymatch.WithTverskyAlgorithm(1.0, 1.0, 1.0, 3),
//	    fuzzymatch.WithThreshold(0.85),
//	)
//	if errors.Is(err, fuzzymatch.ErrInvalidTverskyParam) {
//	    // diagnostic
//	}
var ErrInvalidTverskyParam = errors.New("fuzzymatch: invalid tversky parameter")

// ErrInvalidAlgoID indicates an AlgoID parameter does not match any
// registered algorithm in the dispatch table. Returned by the Phase 8
// Scorer option layer (WithAlgorithm, WithMongeElkanAlgorithm, …) when
// called with an out-of-range AlgoID (e.g. AlgoID(999)), a negative
// AlgoID, or a catalogue AlgoID whose dispatch entry has not yet been
// populated.
//
// Common causes: casting an int from external configuration to AlgoID
// without bounds-checking; using a constant from a future minor release
// that the running version does not yet ship; passing AlgoMongeElkan
// to WithMongeElkanAlgorithm (trivial-recursion guard — use
// WithAlgorithm with a different inner instead).
//
// Resolution: call AlgoIDs() to discover the valid set rather than
// guessing; for Monge-Elkan inner metrics, prefer one of the 18
// permitted inner AlgoIDs documented in monge_elkan.go. Direct calls
// to MongeElkanScore (symmetric default) / MongeElkanScoreAsymmetric
// (directional) panic with a value wrapping ErrInvalidInnerAlgo (the
// inner-specific sentinel) on a non-permitted inner.
//
// Example:
//
//	_, err := fuzzymatch.NewScorer(
//	    fuzzymatch.WithAlgorithm(fuzzymatch.AlgoLevenshtein, 1.0),
//	    fuzzymatch.WithThreshold(0.85),
//	)
//	if errors.Is(err, fuzzymatch.ErrInvalidAlgoID) {
//	    // diagnostic
//	}
var ErrInvalidAlgoID = errors.New("fuzzymatch: invalid algorithm identifier")

// ErrInvalidInnerAlgo indicates a Monge-Elkan composite was constructed
// with an inner AlgoID outside the 18-entry permitted set documented in
// monge_elkan.go. The Monge-Elkan formula composes a per-token inner
// similarity; the inner metric must itself be a per-pair character-,
// q-gram-, gestalt-, or phonetic-tier algorithm — token-tier algorithms
// (TokenSortRatio, TokenSetRatio, PartialRatio, TokenJaccard) and the
// MongeElkan self-reference are rejected.
//
// Common causes: passing AlgoMongeElkan to WithMongeElkanAlgorithm
// (self-recursion); passing a token-tier AlgoID as the inner metric
// in the mistaken belief that ME composes arbitrary algorithms.
//
// Resolution: pick one of the 18 permitted inner AlgoIDs (9 character-
// tier + 4 q-gram tier + 1 gestalt + 4 phonetic-tier). The default
// inner is AlgoJaroWinkler — pass that via WithAlgorithm(AlgoMongeElkan,
// w) for the typical case. Direct MongeElkanScore (symmetric default)
// / MongeElkanScoreAsymmetric (directional) calls with a non-permitted
// inner panic with a value wrapping this sentinel.
//
// Example:
//
//	defer func() {
//	    if r := recover(); r != nil {
//	        if err, ok := r.(error); ok && errors.Is(err, fuzzymatch.ErrInvalidInnerAlgo) {
//	            // programmer error — pick a permitted inner
//	        }
//	    }
//	}()
//	_ = fuzzymatch.MongeElkanScore("a b", "c d", fuzzymatch.AlgoMongeElkan)
var ErrInvalidInnerAlgo = errors.New("fuzzymatch: invalid inner algorithm for Monge-Elkan composite")

// ErrEmptyScorer indicates NewScorer was called without any algorithm
// option — the option slice contained zero WithAlgorithm /
// With*Algorithm entries by the time validation ran. A Scorer with no
// algorithms has no meaningful composite to compute.
//
// Pass at least one WithAlgorithm option (or use DefaultScorer() for
// the opinionated six-algorithm composition).
//
// Returned by NewScorer (Phase 8) after the missing-threshold check
// passes and the option-validation pipeline finds cfg.entries empty.
//
// Discriminate via errors.Is(err, fuzzymatch.ErrEmptyScorer); never
// match the error message string.
var ErrEmptyScorer = errors.New("fuzzymatch: scorer has no algorithms (pass at least one WithAlgorithm option or use DefaultScorer)")

// ErrInvalidWeight indicates an algorithm weight passed to a Phase 8
// Scorer With*Algorithm option was ≤ 0. Weights must be strictly
// positive so that the auto-normalisation step (sum-to-1) has a
// positive divisor and the composite score remains in [0.0, 1.0].
//
// Returned by every Phase 8 With*Algorithm option at
// option-application time when the supplied weight fails the
// strict-positive constraint.
//
// Discriminate via errors.Is(err, fuzzymatch.ErrInvalidWeight); never
// match the error message string.
var ErrInvalidWeight = errors.New("fuzzymatch: invalid algorithm weight (must be > 0)")

// ErrInvalidThreshold indicates a WithThreshold value was outside the
// closed interval [0.0, 1.0]. Thresholds are compared against composite
// scores which the Scorer guarantees fall in [0.0, 1.0] under default
// weight normalisation; values outside the interval are non-sensical.
//
// Returned by WithThreshold at option-application time and surfaced
// through NewScorer.
//
// Discriminate via errors.Is(err, fuzzymatch.ErrInvalidThreshold);
// never match the error message string.
var ErrInvalidThreshold = errors.New("fuzzymatch: invalid threshold (must be in [0.0, 1.0])")

// ErrMissingThreshold indicates NewScorer was called without
// WithThreshold. The threshold is a calibration parameter with no
// universally-safe default, so the library refuses to guess. Pass
// WithThreshold(t) with t ∈ [0.0, 1.0] during construction, or use
// DefaultScorer() for the opinionated default composition that bakes
// 0.85 in.
//
// Returned by NewScorer (Phase 8) when no WithThreshold option is
// present in the variadic opts slice. This check fires FIRST in the
// validation pipeline so the error is unambiguous when a user forgets
// the threshold alongside another option problem.
//
// Discriminate via errors.Is(err, fuzzymatch.ErrMissingThreshold);
// never match the error message string.
var ErrMissingThreshold = errors.New("fuzzymatch: scorer threshold required (pass WithThreshold or use DefaultScorer)")

// ErrInternalInvariantViolated indicates a library invariant that
// should never fire in correct usage was violated — seeing it as a
// recovered panic value is unambiguous evidence of a library bug.
//
// Common causes: NEVER caller-supplied input. The single current panic
// site is DefaultScorer() when its baked-in option pipeline fails
// validation (which would indicate the locked default composition has
// drifted out of sync with NewScorer's validation pipeline). Future
// "this should be impossible" guards may wrap the same sentinel.
//
// Resolution: file a bug report against
// https://github.com/axonops/fuzzymatch with the panic stack trace
// and a minimal reproducer. This sentinel MUST NOT be used to wrap
// caller-supplied parameter errors; parameter errors use the dedicated
// parameter sentinels (ErrInvalidAlgoID, ErrInvalidWeight, etc.).
//
// Example:
//
//	defer func() {
//	    if r := recover(); r != nil {
//	        if err, ok := r.(error); ok && errors.Is(err, fuzzymatch.ErrInternalInvariantViolated) {
//	            // library bug — collect stack + file issue
//	        }
//	    }
//	}()
//	s := fuzzymatch.DefaultScorer()
//	_ = s.Score("a", "b")
var ErrInternalInvariantViolated = errors.New("fuzzymatch: internal invariant violated (library bug — please file an issue)")

// ErrInvalidSWGParam indicates a Smith-Waterman-Gotoh parameter struct
// (SWGParams) violates the documented sign-and-finite invariants:
// Match must be a finite, non-negative float; Mismatch, GapOpen, and
// GapExtend must be finite, non-positive floats; NaN and ±Inf are
// rejected in every field. The Tversky-style strict-parameter framework
// (docs/requirements.md §6.A) applies — invalid construction is a
// programmer error and panics with a typed-error value wrapping this
// sentinel so consumers can discriminate via errors.Is on a recovered
// panic value.
//
// Common causes: copying a partially-initialised SWGParams across a
// boundary; computing GapOpen / GapExtend from upstream arithmetic that
// produced NaN or Inf; flipping Match negative by accident; mutating
// the SWGParams returned by NewSWGParams and forgetting to call
// Validate() before passing it to SmithWatermanGotohScoreWithParams.
//
// Resolution: pass params satisfying Match ≥ 0, Mismatch ≤ 0,
// GapOpen ≤ 0, GapExtend ≤ 0, all finite. The canonical defaults from
// NewSWGParams (Match=1.0, Mismatch=-1.0, GapOpen=-1.5, GapExtend=-0.5)
// always pass this gate; consumers mutating the struct after
// construction should call params.Validate() to re-assert before use.
//
// Example:
//
//	defer func() {
//	    if r := recover(); r != nil {
//	        if err, ok := r.(error); ok && errors.Is(err, fuzzymatch.ErrInvalidSWGParam) {
//	            // programmer error — log and re-panic, or substitute defaults
//	        }
//	    }
//	}()
//	params := fuzzymatch.NewSWGParams()
//	params.Match = math.NaN() // consumer error
//	params.Validate()         // panics with a typed error wrapping ErrInvalidSWGParam
var ErrInvalidSWGParam = errors.New("fuzzymatch: invalid Smith-Waterman-Gotoh parameters")
