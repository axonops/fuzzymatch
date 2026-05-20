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

// scan.go declares the Phase 9 Layer-3 collection-scan surface of the
// three-layer fuzzymatch architecture: Item, Config, Warning, Kind,
// Check, DefaultConfig. It consumes the Phase 8 *fuzzymatch.Scorer for
// per-pair similarity and orchestrates within/cross-group passes with
// token-bucket optimisation, deterministic sort, and suppression
// composition.
//
// SPEC OVERRIDE (Phase 9):
//
//   - Warning.Scores is map[fuzzymatch.AlgoID]float64 (NOT
//     map[string]float64 as in docs/requirements.md §12.1 line 1337).
//     Extends the Phase 8 ScoreAll override at §8.3 + §8.6
//     (08-CONTEXT.md §1) for the same typed-enum-keys rationale. See
//     09-CONTEXT.md §1 D-01.
//   - Kind type renamed from WarningKind (NOT scan.WarnKind to avoid
//     accidental symmetry with root's WarnKind). See 09-CONTEXT.md §1
//     D-02. Spec §12.1 renamed in lockstep.
//
// api-ergonomics-reviewer signed off on both overrides in plan 09-01's
// PR, mirroring the Phase 7 Plan 07-02 and Phase 8 Plan 08-03 precedent
// of recording the reviewer verdict in the PR description.
//
// Design notes (per 09-CONTEXT.md + 09-RESEARCH.md):
//
//   - Plan 09-01 lands the foundation skeleton: Item / Config /
//     Warning struct declarations, the three sentinel errors, the
//     DefaultConfig opinionated helper, and the Check stub that returns
//     (nil, ErrNilScorer) when cfg.Scorer is nil. Plan 09-02 adds the
//     full validation pipeline; Plan 09-03 lands the Check body for the
//     within-group and naive cross-group passes; Plan 09-04 wires in the
//     token-bucket optimisation (bucketThreshold = 50, private const);
//     Plan 09-05 adds suppression composition; Plan 09-06 adds the
//     deterministic sort and in-line completeness assertion.
//
//   - Validation pipeline order (P1..P4) is LOCKED (09-CONTEXT.md §2):
//     nil-Scorer fail-fast → Config field validation → Items validation
//     (D-03 + D-06 collect-all via errors.Join) → SuppressedPairs
//     validation (D-05 collect-all).
//
//   - The Scorer's normalisation options are accessed via the new
//     Scorer.NormalisationOptions() public method introduced in Plan
//     09-01 (resolves 09-RESEARCH.md Open Question 1). Plans 09-04 +
//     09-05 will consume this accessor when building token buckets and
//     canonicalising SuppressedPairs.
//
//   - In-line completeness assertion (added in Plan 09-06) panics with
//     fuzzymatch.ErrInternalInvariantViolated (Phase 8.5 Gap 5) when
//     two adjacent post-sort warnings share the (Kind, NameA, NameB,
//     GroupA, GroupB) sort key — that would only occur via a library
//     bug, never via caller input, because D-06 rejects duplicate
//     (Name, Group) at validation time.
//
//   - No goroutines, no channels, no mutexes. Pure-function library.

package scan

import (
	"github.com/axonops/fuzzymatch"
)

// Item is one named thing the scanner compares. Consumers construct
// items from whatever schema or vocabulary they care about.
//
// Concurrency: Item is a plain data struct with no methods and no
// internal synchronisation. Consumers are responsible for not mutating
// an Item that scan.Check is currently reading.
type Item struct {
	// Name is the value being compared. Required, non-empty — Check
	// returns ErrInvalidItem (wrapped with the offending index) when
	// Name is the empty string.
	Name string

	// Group scopes the comparison. Items with the same Group are
	// compared against each other in the within-group pass. Items with
	// different Group values are compared in the cross-group pass when
	// Config.CompareAcrossGroups is true. The empty string is a valid
	// Group (a single global group).
	Group string

	// SilenceLint, when true, suppresses any warning involving this
	// Item. The suppression is one-sided: setting the flag on either
	// side of a pair silences the pair.
	SilenceLint bool

	// Tag is opaque consumer data. The library does not interpret it;
	// it appears unchanged in Warning.TagA / Warning.TagB. Use it to
	// carry source-file line numbers, schema paths, or any other
	// context useful in diagnostics.
	//
	// Security note: Tag is never stringified inside error messages or
	// panic values. Only the int index of an offending Item appears in
	// validation errors. Consumers may safely store sensitive data in
	// Tag without risking leakage through the error surface.
	Tag any
}

// Config controls Check behaviour. Construct directly with a non-nil
// Scorer, or use DefaultConfig for the opinionated default.
//
// The zero value of Config is invalid: Scorer is nil, so Check returns
// ErrNilScorer. CrossGroupThresholdBoost zero-value is 0.0 (no boost).
// The opinionated default 0.05 lives in DefaultConfig (SPEC OVERRIDE
// (Phase 9): default-0.05 location migrated to DefaultConfig per
// 09-CONTEXT.md §2 D-04; spec §12.1 line 1359 was amended in lockstep).
//
// Concurrency: Config is a plain data struct. Consumers passing the
// same Config to multiple concurrent Check invocations must not mutate
// it between calls.
type Config struct {
	// Scorer is required. It governs the similarity computation and the
	// emission threshold (warnings are emitted when Scorer.Match returns
	// true, with the cross-group boost applied where relevant).
	//
	// Check returns ErrNilScorer when Scorer is nil.
	Scorer *fuzzymatch.Scorer

	// CompareAcrossGroups enables the cross-group pass: items with
	// different Group values are compared against each other. When
	// false (default), only same-group pairs are compared.
	CompareAcrossGroups bool

	// CrossGroupThresholdBoost is added to the Scorer's threshold when
	// evaluating cross-group pairs. The cross-group pass is inherently
	// noisier than the within-group pass; a small positive value
	// (typically 0.02–0.05) reduces false positives without disabling
	// the pass.
	//
	// Range: [0.0, 1.0]. NaN, ±Inf, or out-of-range values are
	// rejected with ErrInvalidConfig at Check entry (D-04). The
	// zero-value is 0.0 (no boost); the opinionated default 0.05
	// lives in DefaultConfig.
	//
	// If scorer.Threshold() + CrossGroupThresholdBoost exceeds 1.0,
	// the effective cross-group threshold is clamped to 1.0 (only
	// byte-identical matches pass; combined with
	// CompareIdenticalAcrossGroups=false this disables cross-group
	// emission, which is documented behaviour).
	CrossGroupThresholdBoost float64

	// CompareIdenticalAcrossGroups, when false (the recommended
	// default), suppresses cross-group warnings for pairs whose names
	// are byte-identical after normalisation. Operators legitimately
	// reuse the same name (e.g. user_id) across groups; surfacing every
	// such pair would drown real similar-but-not-equal signals.
	CompareIdenticalAcrossGroups bool

	// SuppressedPairs is a list of name pairs that should not produce a
	// warning, in addition to per-Item SilenceLint flags. Pairs are
	// order-independent and canonicalised at Check entry using the
	// Scorer's normalisation options.
	//
	// Range: every entry's two strings must be non-empty.
	// ErrInvalidConfig is returned at Check entry for any entry where
	// one or both strings are empty (D-05); offending indices are
	// collected via errors.Join.
	//
	// Self-pairs (a == b after normalisation) are silently kept — they
	// are harmless because Check never emits a self-warning.
	//
	// Build cost is O(N) in len(SuppressedPairs); per-candidate lookup
	// is O(1) via an internal canonical-pair map built once at Check
	// entry.
	SuppressedPairs [][2]string
}

// Warning is one detected similar-name pair. NameA is the
// lexicographically smaller of the pair (after normalisation) so output
// ordering is deterministic.
//
// SPEC OVERRIDE (Phase 9): Scores is map[fuzzymatch.AlgoID]float64
// (typed enum keys), NOT map[string]float64 as docs/requirements.md
// §12.1 line 1337 originally specified. Extends the Phase 8 ScoreAll
// override at §8.3 + §8.6 for the same typed-enum-keys rationale: the
// rest of the library exposes AlgoID and gives consumers compile-time
// key safety. Use AlgoID.String() for snake_case / CamelCase display.
// See 09-CONTEXT.md §1 D-01 and the api-ergonomics-reviewer sign-off
// recorded in plan 09-01's PR. Spec §12.1 was amended in lockstep with
// this declaration.
//
// Concurrency: Warning is a plain data struct. The map in Scores is
// freshly allocated by Check for each emitted Warning; consumers may
// mutate it freely without affecting other Warning values.
type Warning struct {
	// Kind classifies the pair scope (within-group vs cross-group).
	// SPEC OVERRIDE (Phase 9): type renamed from WarningKind to Kind
	// per 09-CONTEXT.md §1 D-02; spec §12.1 was amended in lockstep.
	Kind Kind

	// NameA, NameB are the raw item names (NOT normalised). NameA is
	// the lexicographically smaller after normalisation so ordering is
	// stable; NameB is the lexicographically larger.
	NameA, NameB string

	// GroupA, GroupB are the raw group values from the corresponding
	// Items. For KindWithinGroup warnings GroupA == GroupB; for
	// KindAcrossGroups warnings GroupA != GroupB.
	GroupA, GroupB string

	// TagA, TagB are the opaque Tag values from the corresponding
	// Items. The library does not interpret them.
	TagA, TagB any

	// Score is the composite score returned by Scorer.Score for this
	// pair. In [0.0, 1.0] under the default WithNormaliseWeights(true)
	// Scorer construction; the consumer takes responsibility for the
	// range otherwise.
	Score float64

	// Scores carries the per-algorithm breakdown from
	// Scorer.ScoreAll(NameA, NameB).
	//
	// SPEC OVERRIDE (Phase 9): map[fuzzymatch.AlgoID]float64 (typed
	// enum keys) per 09-CONTEXT.md §1 D-01. Extends the Phase 8
	// ScoreAll override at §8.3 + §8.6. Iteration order is
	// non-deterministic per Go map semantics; consumers requiring
	// stable order MUST sort the keys themselves (typically via
	// fuzzymatch.AlgoIDs() then key-lookup).
	Scores map[fuzzymatch.AlgoID]float64
}

// DefaultConfig returns an opinionated Config bound to the supplied
// Scorer. The defaults are:
//
//   - Scorer:                       s (the argument)
//   - CompareAcrossGroups:          false (within-group only)
//   - CrossGroupThresholdBoost:     0.05 — applied only when the
//     consumer subsequently sets CompareAcrossGroups = true
//   - CompareIdenticalAcrossGroups: false (suppress identical names
//     across groups by default)
//   - SuppressedPairs:              nil
//
// SPEC OVERRIDE (Phase 9): the 0.05 default lives here, NOT as the
// zero-value of Config.CrossGroupThresholdBoost; see 09-CONTEXT.md §2
// D-04. Spec §12.1 line 1359 was amended in lockstep — the spec's
// "Default: 0.05" sentence was migrated from the field godoc to this
// helper godoc.
//
// Mirrors the Phase 8 functional-options-with-helper pattern
// (DefaultScorer / DefaultScorerOptions): the zero-value of the Config
// struct remains a valid (minimal) configuration; the opinionated
// helper bakes in the experience-tuned values.
//
// Typical usage:
//
//	s := fuzzymatch.DefaultScorer()
//	cfg := scan.DefaultConfig(s)
//	cfg.CompareAcrossGroups = true // opt in
//	warnings, err := scan.Check(items, cfg)
//
// Concurrency: DefaultConfig has no side effects and is safe for
// concurrent use. A fresh Config value is returned on every call.
func DefaultConfig(s *fuzzymatch.Scorer) Config {
	return Config{
		Scorer:                       s,
		CompareAcrossGroups:          false,
		CrossGroupThresholdBoost:     0.05,
		CompareIdenticalAcrossGroups: false,
		SuppressedPairs:              nil,
	}
}

// Check compares every pair of items per the Config and returns
// warnings for pairs where the Scorer's Match returns true (with the
// cross-group threshold boost applied where relevant).
//
// Output is sorted deterministically by
// (Kind, NameA, NameB, GroupA, GroupB) and is byte-identical across
// runs for the same input and Scorer configuration.
//
// Check is a pure function: it never reads files, environment
// variables, or any package-global state. Safe for concurrent
// invocation on disjoint inputs.
//
// Returns ErrNilScorer if cfg.Scorer is nil. The full validation
// pipeline (ErrInvalidItem for empty Name / duplicate (Name, Group);
// ErrInvalidConfig for out-of-range CrossGroupThresholdBoost / empty
// SuppressedPairs entries) lands in Plan 09-02. The full similarity
// body (within-group + cross-group + bucket + suppression + sort)
// lands across Plans 09-03 through 09-06. Empty items slice returns
// an empty Warning slice and no error.
//
// Plan 09-01 status: this is a foundation stub. It returns
// (nil, ErrNilScorer) when cfg.Scorer is nil, and (nil, nil)
// otherwise. The empty body is intentional scaffolding so subsequent
// plans have a callable Check function to extend without a
// half-finished algorithm landing in the meantime.
func Check(items []Item, cfg Config) ([]Warning, error) {
	if cfg.Scorer == nil {
		return nil, ErrNilScorer
	}
	// Plan 09-02 lands the full validation pipeline (D-03 + D-04 +
	// D-05 + D-06). Plan 09-03 lands the Check body. Until then the
	// stub returns (nil, nil) on a non-nil Scorer — sufficient for
	// Plan 09-01's smoke tests and Plan 09-02's validation-only tests.
	_ = items
	return nil, nil
}
