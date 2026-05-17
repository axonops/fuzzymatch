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

// scorer_internal_test.go pins the unexported invariants of the Scorer
// type that the external scorer_test.go (package fuzzymatch_test)
// cannot observe through the public API alone:
//
//   - Duplicate WithAlgorithm(SameAlgoID, _) calls collapse to ONE
//     entry by last-write-wins (08-RESEARCH.md Pitfall 4).
//   - WithNormaliseWeights(true) (the default) produces post-NewScorer
//     weights that sum to exactly 1.0 left-to-right reduction.
//   - WithNormaliseWeights(false) preserves raw consumer-supplied
//     weights.
//   - The internal algorithmsAlgoIDSorted slice is AlgoID-ascending
//     regardless of the option order the consumer supplied.
//
// Living in package fuzzymatch (no _test suffix) is the conventional Go
// pattern for exposing unexported state to internal tests; the file's
// _test.go suffix ensures it never ships in the public artefact.
//
// Stdlib `testing` only — no testify in root tests, per
// .claude/skills/go-coding-standards.

package fuzzymatch

import (
	"errors"
	"testing"
)

// TestScorer_LastWriteWins verifies the duplicate-AlgoID dedup
// semantics from 08-RESEARCH.md Pitfall 4: when the consumer applies
// two WithAlgorithm(SameID, _) options, only the LATER weight survives
// after NewScorer.
//
// With WithNormaliseWeights default-true and only one surviving entry,
// the post-normalisation weight is 1.0 (the single entry's weight
// divided by itself).
func TestScorer_LastWriteWins(t *testing.T) {
	s, err := NewScorer(
		WithAlgorithm(AlgoLevenshtein, 0.3),
		WithAlgorithm(AlgoLevenshtein, 0.7),
		WithThreshold(0.5),
	)
	if err != nil {
		t.Fatalf("NewScorer returned err: %v", err)
	}
	if got := len(s.algorithmsAlgoIDSorted); got != 1 {
		t.Fatalf("entries after dedup: got %d, want 1 (last-write-wins)", got)
	}
	if got := s.algorithmsAlgoIDSorted[0].id; got != AlgoLevenshtein {
		t.Errorf("entry id: got %v, want AlgoLevenshtein", got)
	}
	// Single entry with normalisation → weight = w / w = 1.0 exactly.
	if got := s.algorithmsAlgoIDSorted[0].weight; got != 1.0 {
		t.Errorf("single-entry normalised weight: got %g, want 1.0", got)
	}
}

// TestScorer_WeightNormalisation_SumsToOne verifies the SCORER-03
// sum-to-1 invariant. A 2-algorithm Scorer with raw weights 1.0 and 3.0
// normalises to (0.25, 0.75). The sum, computed left-to-right
// (per CONTEXT.md §5 LOCKED determinism reduction), equals exactly 1.0
// because the two normalised weights are dyadic-friendly fractions of
// the sum (1/4 + 3/4 = 1.0 in IEEE-754 — 0.25 and 0.75 are both exactly
// representable).
func TestScorer_WeightNormalisation_SumsToOne(t *testing.T) {
	s, err := NewScorer(
		WithAlgorithm(AlgoLevenshtein, 1.0),
		WithAlgorithm(AlgoJaroWinkler, 3.0),
		WithThreshold(0.5),
	)
	if err != nil {
		t.Fatalf("NewScorer returned err: %v", err)
	}
	if got := len(s.algorithmsAlgoIDSorted); got != 2 {
		t.Fatalf("entries: got %d, want 2", got)
	}
	// Both 0.25 and 0.75 are exactly representable in IEEE-754, so the
	// sum is byte-exactly 1.0. No tolerance needed — assert exact ==.
	sum := s.algorithmsAlgoIDSorted[0].weight + s.algorithmsAlgoIDSorted[1].weight
	if sum != 1.0 {
		t.Errorf(
			"normalised weight sum: got %g (entries: %g + %g), want exactly 1.0",
			sum,
			s.algorithmsAlgoIDSorted[0].weight,
			s.algorithmsAlgoIDSorted[1].weight,
		)
	}
	// Spot-check the individual values for sanity.
	if got := s.algorithmsAlgoIDSorted[0].weight; got != 0.25 {
		t.Errorf("entry[0] weight: got %g, want 0.25", got)
	}
	if got := s.algorithmsAlgoIDSorted[1].weight; got != 0.75 {
		t.Errorf("entry[1] weight: got %g, want 0.75", got)
	}
}

// TestScorer_WeightNormalisation_DisabledRawPreserved verifies that
// WithNormaliseWeights(false) leaves raw consumer-supplied weights
// untouched. A 2-algorithm Scorer with raw weights 1.0 and 3.0 should
// retain those exact values when normalisation is disabled.
func TestScorer_WeightNormalisation_DisabledRawPreserved(t *testing.T) {
	s, err := NewScorer(
		WithAlgorithm(AlgoLevenshtein, 1.0),
		WithAlgorithm(AlgoJaroWinkler, 3.0),
		WithNormaliseWeights(false),
		WithThreshold(0.5),
	)
	if err != nil {
		t.Fatalf("NewScorer returned err: %v", err)
	}
	if got := len(s.algorithmsAlgoIDSorted); got != 2 {
		t.Fatalf("entries: got %d, want 2", got)
	}
	if got := s.algorithmsAlgoIDSorted[0].weight; got != 1.0 {
		t.Errorf("entry[0] raw weight: got %g, want 1.0", got)
	}
	if got := s.algorithmsAlgoIDSorted[1].weight; got != 3.0 {
		t.Errorf("entry[1] raw weight: got %g, want 3.0", got)
	}
}

// TestScorer_EntriesSorted_AlgoIDAscending verifies that NewScorer's
// dedup-and-sort step produces an AlgoID-ascending slice regardless of
// the order in which the consumer supplied the WithAlgorithm options.
// This invariant is load-bearing for the float-determinism reduction
// loop (CONTEXT.md §5): the reduction order must be platform-stable, so
// the slice must be sorted before Score iterates it.
//
// Algorithm picks (chosen for AlgoID stability across phases):
//
//   - AlgoLevenshtein            iota 0  (Phase 2)
//   - AlgoJaro                   iota 4  (Phase 2)
//   - AlgoTokenJaccard           iota 17 (Phase 6)
//
// Supplied in reverse AlgoID order; expected ascending after NewScorer.
func TestScorer_EntriesSorted_AlgoIDAscending(t *testing.T) {
	s, err := NewScorer(
		// Reverse AlgoID order: TokenJaccard (17), Jaro (4), Levenshtein (0).
		WithAlgorithm(AlgoTokenJaccard, 1.0),
		WithAlgorithm(AlgoJaro, 1.0),
		WithAlgorithm(AlgoLevenshtein, 1.0),
		WithThreshold(0.5),
	)
	if err != nil {
		t.Fatalf("NewScorer returned err: %v", err)
	}
	if got := len(s.algorithmsAlgoIDSorted); got != 3 {
		t.Fatalf("entries: got %d, want 3", got)
	}
	// Expected ascending order: Levenshtein, Jaro, TokenJaccard.
	want := []AlgoID{AlgoLevenshtein, AlgoJaro, AlgoTokenJaccard}
	for i, w := range want {
		if got := s.algorithmsAlgoIDSorted[i].id; got != w {
			t.Errorf("entry[%d].id: got %v, want %v", i, got, w)
		}
	}
	// Cross-check: int(AlgoID) is strictly ascending.
	for i := 1; i < len(s.algorithmsAlgoIDSorted); i++ {
		prev := int(s.algorithmsAlgoIDSorted[i-1].id)
		curr := int(s.algorithmsAlgoIDSorted[i].id)
		if prev >= curr {
			t.Errorf(
				"entries[%d..%d] not strictly ascending: got int(%v)=%d >= int(%v)=%d",
				i-1, i,
				s.algorithmsAlgoIDSorted[i-1].id, prev,
				s.algorithmsAlgoIDSorted[i].id, curr,
			)
		}
	}
}

// TestNewScorer_RejectsAlgoIDOutOfRange covers the dispatch-validation
// failure path in NewScorer (scorer.go line 233): when a consumer
// constructs a WithAlgorithm option with an AlgoID whose dispatch[id]
// is nil (either negative, beyond numAlgorithms, or a gap in the
// dispatch table), NewScorer returns ErrInvalidAlgoID. The default
// public surface cannot produce such an option — this test reaches the
// branch by constructing an option from inside the package with a
// manually-set AlgoID outside the canonical range.
func TestNewScorer_RejectsAlgoIDOutOfRange(t *testing.T) {
	bad := func(cfg *scorerConfig) error {
		cfg.entries = append(cfg.entries, scorerEntry{
			id:     AlgoID(numAlgorithms + 100), // beyond dispatch table
			weight: 1.0,
		})
		return nil
	}
	_, err := NewScorer(bad, WithThreshold(0.5))
	if err == nil {
		t.Fatal("NewScorer with out-of-range AlgoID: want ErrInvalidAlgoID, got nil")
	}
	if !errorsIsLocal(err, ErrInvalidAlgoID) {
		t.Errorf("NewScorer with out-of-range AlgoID: want ErrInvalidAlgoID, got %v", err)
	}
}

// TestNewScorer_RejectsZeroWeightSum covers the sum-zero branch in
// NewScorer (scorer.go line 287): when normalisation is on (default)
// and after dedup the surviving weights sum to zero, NewScorer returns
// ErrInvalidWeight. WithAlgorithm rejects zero up front; reaching this
// branch requires installing a zero-weight entry directly into the
// config via a custom option from inside the package.
func TestNewScorer_RejectsZeroWeightSum(t *testing.T) {
	zero := func(cfg *scorerConfig) error {
		cfg.entries = append(cfg.entries, scorerEntry{
			id:     AlgoLevenshtein,
			weight: 0.0,
		})
		return nil
	}
	_, err := NewScorer(zero, WithThreshold(0.5))
	if err == nil {
		t.Fatal("NewScorer with zero weight sum: want ErrInvalidWeight, got nil")
	}
	if !errorsIsLocal(err, ErrInvalidWeight) {
		t.Errorf("NewScorer with zero weight sum: want ErrInvalidWeight, got %v", err)
	}
}

// TestScorerAlgorithm_TypeReference pins ScorerAlgorithm as a value
// type so the AST-based exported-symbol gate (Floor 3, Phase 8.5 Q12a)
// records a test-file identifier reference for it.
func TestScorerAlgorithm_TypeReference(t *testing.T) {
	sa := ScorerAlgorithm{ID: AlgoLevenshtein, Weight: 1.0}
	if sa.ID != AlgoLevenshtein || sa.Weight != 1.0 {
		t.Errorf("ScorerAlgorithm literal init: got %+v, want {AlgoLevenshtein 1.0}", sa)
	}
	s, err := NewScorer(WithAlgorithm(AlgoLevenshtein, 1.0), WithThreshold(0.5))
	if err != nil {
		t.Fatalf("NewScorer setup: %v", err)
	}
	algos := s.Algorithms()
	if len(algos) != 1 || algos[0].ID != AlgoLevenshtein {
		t.Errorf("Algorithms(): got %+v, want [{AlgoLevenshtein 1.0}]", algos)
	}
	// Pin slice element type — explicit form deliberately keeps the
	// ScorerAlgorithm identifier reference visible to the AST gate.
	//nolint:staticcheck // QF1011: explicit type retained so Floor 3's AST walk sees the identifier.
	var _ []ScorerAlgorithm = algos
}

// errorsIsLocal delegates to stdlib errors.Is. It exists to make the
// import-block dependency on `errors` explicit at the call sites
// elsewhere in this file; tests use `errorsIsLocal(err, target)`
// uniformly.
func errorsIsLocal(err, target error) bool {
	return errors.Is(err, target)
}

// TestMustDefaultScorer_PanicsOnNewScorerError covers the panic body
// in mustDefaultScorer (the testable internal helper for DefaultScorer).
// Reaching this branch via the public API is impossible because
// DefaultScorerOptions() always produces valid options that NewScorer
// accepts; this test injects a failing newScorer stub to exercise the
// defence-in-depth typed-panic.
func TestMustDefaultScorer_PanicsOnNewScorerError(t *testing.T) {
	defer func() {
		r := recover()
		if r == nil {
			t.Fatal("mustDefaultScorer with failing newScorer: expected panic, got none")
		}
		err, ok := r.(error)
		if !ok {
			t.Fatalf("mustDefaultScorer panic value is not an error: %T (%v)", r, r)
		}
		if !errorsIsLocal(err, ErrInternalInvariantViolated) {
			t.Errorf("mustDefaultScorer panic error: want errors.Is(_, ErrInternalInvariantViolated), got %v", err)
		}
	}()
	// Inject a newScorer that always fails.
	stubErr := errors.New("simulated NewScorer failure for defence-in-depth test")
	mustDefaultScorer(func(opts ...ScorerOption) (*Scorer, error) {
		return nil, stubErr
	})
	t.Fatal("mustDefaultScorer with failing newScorer: returned normally (should have panicked)")
}
