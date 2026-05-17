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
// errors.As — never via string matching. See docs/requirements.md §6.4.
//
// Error message convention (per .claude/skills/go-coding-standards):
//
//   - Every message starts with the "fuzzymatch: " package prefix so
//     wrappers like fmt.Errorf("scorer: %w", err) produce readable
//     compositions ("scorer: fuzzymatch: invalid input").
//   - The text after the prefix is lowercase and carries no trailing
//     punctuation ('.', '!', or '?') so concatenation flows cleanly.
//   - Each sentinel is a flat errors.New value, not a typed struct;
//     richer per-item context can be added in Phase 9 if scan needs it.
//
// The four sentinels named here are the v1.x contract; additional
// sentinels (e.g. ErrInvalidThreshold for the Scorer, or
// ErrHammingLengthMismatch for the Hamming algorithm) land alongside
// the features that introduce them in later phases.

package fuzzymatch

import "errors"

// ErrInvalidInput indicates a caller-provided string fails an
// algorithm's documented input constraints — most commonly invalid
// UTF-8 on a rune-aware API, or a non-comparable embedded NUL byte
// where the algorithm rejects it.
//
// Most algorithms accept arbitrary bytes and do NOT return this error;
// the exceptions document their constraints in their own godoc.
//
// Discriminate via errors.Is(err, fuzzymatch.ErrInvalidInput); never
// match the error message string.
var ErrInvalidInput = errors.New("fuzzymatch: invalid input")

// ErrInvalidConfiguration indicates a Scorer or Extract option set is
// internally inconsistent — for example, a negative weight, a threshold
// outside [0.0, 1.0], an empty algorithm list, or a normalisation
// option combination the library forbids.
//
// Returned by the option-applying constructors in Phase 8 (Scorer) and
// Phase 10 (Extract). See docs/requirements.md §8.
var ErrInvalidConfiguration = errors.New("fuzzymatch: invalid configuration")

// ErrInvalidQGramSize indicates a q-gram-based algorithm option was
// constructed with n < 1 — q-gram extraction requires a positive window
// length.
//
// Returned by the Phase 8 Scorer options (e.g. WithQGramJaccardAlgorithm,
// WithSorensenDiceAlgorithm, WithCosineAlgorithm, WithTverskyAlgorithm)
// when their n parameter is < 1. Direct algorithm calls
// (QGramJaccardScore, SorensenDiceScore, CosineScore, TverskyScore)
// instead panic with a message containing the text of this sentinel —
// per CONTEXT.md §5 LOCKED, direct calls fail loudly on programmer
// error and the Scorer returns the typed error.
//
// Discriminate via errors.Is(err, fuzzymatch.ErrInvalidQGramSize); never
// match the error message string.
var ErrInvalidQGramSize = errors.New("fuzzymatch: invalid q-gram size")

// ErrInvalidTverskyParam indicates a Tversky algorithm option was
// constructed with an invalid α/β parameter pair: either α < 0, β < 0,
// or α + β == 0. The Tversky formula requires non-negative weights with
// at least one strictly positive so the denominator does not collapse.
//
// Returned by the Phase 8 Scorer option WithTverskyAlgorithm when the
// supplied (alpha, beta) violate any of the three constraints. Direct
// calls to TverskyScore / TverskyScoreRunes panic with a message
// containing the text of this sentinel instead — per CONTEXT.md §5
// LOCKED, direct calls fail loudly on programmer error.
//
// Discriminate via errors.Is(err, fuzzymatch.ErrInvalidTverskyParam);
// never match the error message string.
var ErrInvalidTverskyParam = errors.New("fuzzymatch: invalid tversky parameter")

// ErrInvalidAlgorithm indicates an AlgoID parameter does not match any
// registered algorithm in the dispatch table. Returned from the
// package-internal dispatch helpers (Phase 8+) when called with an
// out-of-range AlgoID (e.g. AlgoID(999)) or with a catalogue AlgoID
// whose dispatch entry has not yet been populated.
//
// Consumers should call AlgoIDs() to discover the valid set rather
// than guessing.
var ErrInvalidAlgorithm = errors.New("fuzzymatch: invalid algorithm")

// ErrEmptyInput indicates BOTH input strings are empty at the boundary
// of an API that has no defined empty-empty behaviour. Individual
// algorithm score functions handle empty inputs per their per-algorithm
// specification in docs/requirements.md §7 and do NOT return this
// error — only higher-level APIs (Scorer, Extract) that require
// non-degenerate input may surface it.
var ErrEmptyInput = errors.New("fuzzymatch: empty input")

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
