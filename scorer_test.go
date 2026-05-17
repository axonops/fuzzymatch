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
//
// scorer_test.go pins the external (consumer-perspective) contract of
// the Phase 8 plan 08-02 Scorer surface:
//
//   - NewScorer error paths (ErrMissingThreshold FIRST, ErrEmptyScorer,
//     and the option-layer errors propagated through the constructor).
//   - Score happy paths (identity, AlgoID-sorted reduction with a
//     single algorithm, deterministic-across-calls invariant).
//   - Match boundary behaviour (boundary-inclusive >= threshold,
//     below-threshold false).
//   - WithoutNormalisation semantic (raw inputs reach the per-algorithm
//     dispatch; identifier-style "XMLParser" vs "xml_parser" scores
//     differ between normalisation modes).
//
// Stdlib `testing` only — no testify in root tests, per
// .claude/skills/go-coding-standards.

package fuzzymatch_test

import (
	"errors"
	"testing"

	"github.com/axonops/fuzzymatch"
)

// TestNewScorer_MissingThreshold pins the LOCKED validation-pipeline
// ordering (CONTEXT.md §2 + 08-RESEARCH.md Pitfall 3): the missing-
// threshold check fires FIRST. Both an empty option list AND a list
// containing algorithms-but-no-threshold must surface
// ErrMissingThreshold (not ErrEmptyScorer for the empty case, and not
// any other error for the algorithms-no-threshold case).
func TestNewScorer_MissingThreshold(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name string
		opts []fuzzymatch.ScorerOption
	}{
		{
			name: "no options at all",
			opts: nil,
		},
		{
			name: "algorithm but no threshold",
			opts: []fuzzymatch.ScorerOption{
				fuzzymatch.WithAlgorithm(fuzzymatch.AlgoLevenshtein, 1.0),
			},
		},
	}
	for _, c := range cases {
		c := c
		t.Run(c.name, func(t *testing.T) {
			t.Parallel()
			s, err := fuzzymatch.NewScorer(c.opts...)
			if s != nil {
				t.Errorf("got non-nil *Scorer; want nil on error")
			}
			if !errors.Is(err, fuzzymatch.ErrMissingThreshold) {
				t.Errorf("err: got %v, want errors.Is(err, ErrMissingThreshold)", err)
			}
		})
	}
}

// TestNewScorer_EmptyScorer pins the second gate in the validation
// pipeline: with WithThreshold present but no algorithms, the
// constructor returns ErrEmptyScorer (distinct from
// ErrMissingThreshold).
func TestNewScorer_EmptyScorer(t *testing.T) {
	t.Parallel()
	s, err := fuzzymatch.NewScorer(fuzzymatch.WithThreshold(0.5))
	if s != nil {
		t.Errorf("got non-nil *Scorer; want nil on error")
	}
	if !errors.Is(err, fuzzymatch.ErrEmptyScorer) {
		t.Errorf("err: got %v, want errors.Is(err, ErrEmptyScorer)", err)
	}
}

// TestNewScorer_InvalidWeightPropagates confirms that an option-layer
// error (in this case ErrInvalidWeight from WithAlgorithm's weight ≤ 0
// gate, already covered in plan 08-01's unit tests) propagates through
// NewScorer's first-error short-circuit. The user sees the FIRST
// malformed option's error, not ErrMissingThreshold (even though that
// gate also applies — the option short-circuit happens before the
// post-application validation).
func TestNewScorer_InvalidWeightPropagates(t *testing.T) {
	t.Parallel()
	s, err := fuzzymatch.NewScorer(
		fuzzymatch.WithAlgorithm(fuzzymatch.AlgoLevenshtein, -0.5),
		fuzzymatch.WithThreshold(0.5),
	)
	if s != nil {
		t.Errorf("got non-nil *Scorer; want nil on error")
	}
	if !errors.Is(err, fuzzymatch.ErrInvalidWeight) {
		t.Errorf("err: got %v, want errors.Is(err, ErrInvalidWeight)", err)
	}
}

// TestScorer_Score_Identity verifies the identity case for a single-
// algorithm Scorer: Score(x, x) for a non-empty x returns 1.0 exactly,
// because LevenshteinScore(x, x) = 1.0 and the normalised single-entry
// weight is 1.0, so the reduction is acc = 0.0 + (1.0 * 1.0) = 1.0
// (byte-exact).
func TestScorer_Score_Identity(t *testing.T) {
	t.Parallel()
	s, err := fuzzymatch.NewScorer(
		fuzzymatch.WithAlgorithm(fuzzymatch.AlgoLevenshtein, 1.0),
		fuzzymatch.WithThreshold(0.5),
	)
	if err != nil {
		t.Fatalf("NewScorer: %v", err)
	}
	if got := s.Score("kitten", "kitten"); got != 1.0 {
		t.Errorf("Score(kitten, kitten) = %g; want 1.0 exactly", got)
	}
}

// TestScorer_Score_AsciiPair verifies a 1-algorithm Scorer reduces to
// the underlying algorithm's score multiplied by the normalised single
// weight (1.0). The composite of one normalised entry must equal the
// raw LevenshteinScore byte-for-byte, because acc = 0 + (1.0 *
// LevenshteinScore) = LevenshteinScore. Crucially this asserts that
// Scorer.Score does NOT modify the score; the normalisation step
// applies to the inputs (which here are ASCII-lowercase already, so
// Normalise is a near-no-op for this test corpus).
func TestScorer_Score_AsciiPair(t *testing.T) {
	t.Parallel()
	s, err := fuzzymatch.NewScorer(
		fuzzymatch.WithAlgorithm(fuzzymatch.AlgoLevenshtein, 1.0),
		fuzzymatch.WithThreshold(0.5),
	)
	if err != nil {
		t.Fatalf("NewScorer: %v", err)
	}
	// The inputs are pure-ASCII-lowercase with no separators or
	// camelCase boundaries — DefaultNormalisationOptions(){Lowercase: true,
	// StripSeparators: true, SeparatorChars: "_-.:/", SplitCamelCase: true}
	// is a no-op on these strings. The composite Score therefore equals
	// LevenshteinScore byte-for-byte.
	want := fuzzymatch.LevenshteinScore("kitten", "sitting")
	got := s.Score("kitten", "sitting")
	if got != want {
		t.Errorf("Score(kitten, sitting) = %g; want %g (single-algo composite must equal LevenshteinScore)", got, want)
	}
	// Sanity check: score is in [0, 1].
	if got < 0.0 || got > 1.0 {
		t.Errorf("Score out of [0, 1]: %g", got)
	}
}

// TestScorer_Score_DeterministicAcrossCalls verifies determinism
// invariant SCORER-04: the same Scorer called 1000 times with the same
// inputs produces byte-identical float64 results. This is a per-run /
// per-process determinism gate; the cross-platform byte-identity is
// gated by the plan 08-04 golden file.
//
// Uses a 2-algorithm Scorer to exercise the reduction loop with a
// non-trivial accumulation (acc starts at 0, then acc1 = 0 + (w0*s0),
// then acc2 = acc1 + (w1*s1)).
func TestScorer_Score_DeterministicAcrossCalls(t *testing.T) {
	t.Parallel()
	s, err := fuzzymatch.NewScorer(
		fuzzymatch.WithAlgorithm(fuzzymatch.AlgoLevenshtein, 0.6),
		fuzzymatch.WithAlgorithm(fuzzymatch.AlgoJaroWinkler, 0.4),
		fuzzymatch.WithThreshold(0.5),
	)
	if err != nil {
		t.Fatalf("NewScorer: %v", err)
	}
	want := s.Score("kitten", "sitting")
	for i := 0; i < 1000; i++ {
		got := s.Score("kitten", "sitting")
		if got != want {
			t.Fatalf("iteration %d: got %g, want %g (determinism breach)", i, got, want)
		}
	}
}

// TestScorer_Match_ThresholdInclusive verifies the >= boundary of
// Match: when Score(a, b) == threshold, Match returns true. We
// construct a single-algorithm Scorer where the identity input yields
// exactly 1.0 and set threshold to exactly 1.0; Match must accept.
//
// This is the boundary-inclusive case explicitly named in the LOCKED
// behaviour for Match (CONTEXT.md §2 + plan 08-02 acceptance criteria).
func TestScorer_Match_ThresholdInclusive(t *testing.T) {
	t.Parallel()
	s, err := fuzzymatch.NewScorer(
		fuzzymatch.WithAlgorithm(fuzzymatch.AlgoLevenshtein, 1.0),
		fuzzymatch.WithThreshold(1.0),
	)
	if err != nil {
		t.Fatalf("NewScorer: %v", err)
	}
	// Score("kitten", "kitten") == 1.0; threshold == 1.0 → Match true.
	if !s.Match("kitten", "kitten") {
		t.Errorf("Match(kitten, kitten) with threshold 1.0: got false; want true (boundary inclusive)")
	}
}

// TestScorer_Match_BelowThreshold verifies the complementary case:
// when Score(a, b) < threshold, Match returns false. We construct a
// Scorer whose underlying LevenshteinScore on a dissimilar pair is
// strictly less than 1.0, then set the threshold to exactly 1.0 to
// force a false return.
func TestScorer_Match_BelowThreshold(t *testing.T) {
	t.Parallel()
	s, err := fuzzymatch.NewScorer(
		fuzzymatch.WithAlgorithm(fuzzymatch.AlgoLevenshtein, 1.0),
		fuzzymatch.WithThreshold(1.0),
	)
	if err != nil {
		t.Fatalf("NewScorer: %v", err)
	}
	// LevenshteinScore("kitten", "sitting") < 1.0; threshold == 1.0 →
	// Match false.
	if s.Match("kitten", "sitting") {
		t.Errorf("Match(kitten, sitting) with threshold 1.0: got true; want false")
	}
}

// TestScorer_WithoutNormalisation verifies that the WithoutNormalisation
// option toggles applyNormalisation off, so Scorer.Score passes raw
// input bytes to every registered algorithm (no Normalise call). For
// inputs that differ in case + separators but are normalisation-
// equivalent ("XMLParser" vs "xml_parser"), the without-normalisation
// Scorer's Levenshtein score is strictly lower than the with-
// normalisation Scorer's score — confirming the gate is actually
// applied at Score time.
//
// Also asserts the identity invariant (Score(x, x) = 1.0) holds
// regardless of normalisation, because LevenshteinScore(x, x) = 1.0 on
// any input.
func TestScorer_WithoutNormalisation(t *testing.T) {
	t.Parallel()
	withNorm, err := fuzzymatch.NewScorer(
		fuzzymatch.WithAlgorithm(fuzzymatch.AlgoLevenshtein, 1.0),
		fuzzymatch.WithThreshold(0.5),
	)
	if err != nil {
		t.Fatalf("NewScorer (with-norm): %v", err)
	}
	withoutNorm, err := fuzzymatch.NewScorer(
		fuzzymatch.WithAlgorithm(fuzzymatch.AlgoLevenshtein, 1.0),
		fuzzymatch.WithoutNormalisation(),
		fuzzymatch.WithThreshold(0.5),
	)
	if err != nil {
		t.Fatalf("NewScorer (without-norm): %v", err)
	}

	const a, b = "XMLParser", "xml_parser"
	with := withNorm.Score(a, b)
	without := withoutNorm.Score(a, b)
	if !(without < with) {
		t.Errorf(
			"raw-bytes score must be < normalised score for identifier-style input: with=%g, without=%g (a=%q b=%q)",
			with, without, a, b,
		)
	}
	// Both must lie in [0, 1].
	if with < 0.0 || with > 1.0 {
		t.Errorf("with-norm score out of [0, 1]: %g", with)
	}
	if without < 0.0 || without > 1.0 {
		t.Errorf("without-norm score out of [0, 1]: %g", without)
	}

	// Identity invariant: Score(x, x) = 1.0 regardless of normalisation.
	if got := withoutNorm.Score("hello", "hello"); got != 1.0 {
		t.Errorf("WithoutNormalisation: Score(hello, hello) = %g; want 1.0", got)
	}
}

// TestScorer_Threshold_ReturnsStoredValue verifies the plain-accessor
// contract for Threshold(): the value passed to WithThreshold during
// NewScorer is returned verbatim. No mutation, no transformation, no
// hidden state.
func TestScorer_Threshold_ReturnsStoredValue(t *testing.T) {
	t.Parallel()
	s, err := fuzzymatch.NewScorer(
		fuzzymatch.WithAlgorithm(fuzzymatch.AlgoLevenshtein, 1.0),
		fuzzymatch.WithThreshold(0.73),
	)
	if err != nil {
		t.Fatalf("NewScorer: %v", err)
	}
	if got := s.Threshold(); got != 0.73 {
		t.Errorf("Threshold() = %g; want 0.73 exactly", got)
	}
}

// TestScorer_Algorithms_FreshSlice verifies that Algorithms() returns a
// fresh slice on every call. Mutating the returned slice MUST NOT
// affect the internal Scorer state, and a subsequent Algorithms() call
// MUST return unmodified data.
func TestScorer_Algorithms_FreshSlice(t *testing.T) {
	t.Parallel()
	s, err := fuzzymatch.NewScorer(
		fuzzymatch.WithAlgorithm(fuzzymatch.AlgoLevenshtein, 1.0),
		fuzzymatch.WithAlgorithm(fuzzymatch.AlgoJaroWinkler, 1.0),
		fuzzymatch.WithThreshold(0.5),
	)
	if err != nil {
		t.Fatalf("NewScorer: %v", err)
	}
	first := s.Algorithms()
	if len(first) != 2 {
		t.Fatalf("first call: len = %d; want 2", len(first))
	}
	// Mutate the first slice: change weight and ID.
	originalID := first[0].ID
	originalWeight := first[0].Weight
	first[0].Weight = 999.0
	first[0].ID = fuzzymatch.AlgoCosine // wrong ID

	second := s.Algorithms()
	if len(second) != 2 {
		t.Fatalf("second call: len = %d; want 2", len(second))
	}
	if second[0].ID != originalID {
		t.Errorf("second call's entry[0].ID = %v; want %v (fresh slice required)", second[0].ID, originalID)
	}
	if second[0].Weight != originalWeight {
		t.Errorf("second call's entry[0].Weight = %g; want %g (fresh slice required)", second[0].Weight, originalWeight)
	}
}

// TestScorer_Algorithms_SortedAscending verifies that the returned
// slice is in AlgoID-ascending order regardless of the option-
// application order the consumer supplied. The internal slice is
// AlgoID-sorted at NewScorer time (per scorer_internal_test.go's
// invariant test); Algorithms() preserves that order.
func TestScorer_Algorithms_SortedAscending(t *testing.T) {
	t.Parallel()
	// Supplied in reverse AlgoID order:
	//   TokenJaccard (17), JaroWinkler (5), Levenshtein (0).
	s, err := fuzzymatch.NewScorer(
		fuzzymatch.WithAlgorithm(fuzzymatch.AlgoTokenJaccard, 1.0),
		fuzzymatch.WithAlgorithm(fuzzymatch.AlgoJaroWinkler, 1.0),
		fuzzymatch.WithAlgorithm(fuzzymatch.AlgoLevenshtein, 1.0),
		fuzzymatch.WithThreshold(0.5),
	)
	if err != nil {
		t.Fatalf("NewScorer: %v", err)
	}
	algos := s.Algorithms()
	if len(algos) != 3 {
		t.Fatalf("len = %d; want 3", len(algos))
	}
	want := []fuzzymatch.AlgoID{
		fuzzymatch.AlgoLevenshtein,
		fuzzymatch.AlgoJaroWinkler,
		fuzzymatch.AlgoTokenJaccard,
	}
	for i, w := range want {
		if algos[i].ID != w {
			t.Errorf("algos[%d].ID = %v; want %v", i, algos[i].ID, w)
		}
	}
	// Strict-ascending integer cross-check.
	for i := 1; i < len(algos); i++ {
		if int(algos[i-1].ID) >= int(algos[i].ID) {
			t.Errorf(
				"algos[%d..%d] not strictly ascending: int(%v)=%d >= int(%v)=%d",
				i-1, i,
				algos[i-1].ID, int(algos[i-1].ID),
				algos[i].ID, int(algos[i].ID),
			)
		}
	}
}

// TestScorer_Algorithms_PostNormalisationWeights verifies that the
// returned ScorerAlgorithm.Weight is the POST-normalisation weight that
// the Scorer actually uses during Score's reduction. A 2-algorithm
// Scorer with raw weights 1.0 and 3.0 normalises to (0.25, 0.75); both
// values are dyadic-friendly fractions and exactly representable in
// IEEE-754, so the comparison is exact ==.
func TestScorer_Algorithms_PostNormalisationWeights(t *testing.T) {
	t.Parallel()
	s, err := fuzzymatch.NewScorer(
		fuzzymatch.WithAlgorithm(fuzzymatch.AlgoLevenshtein, 1.0),
		fuzzymatch.WithAlgorithm(fuzzymatch.AlgoJaroWinkler, 3.0),
		fuzzymatch.WithThreshold(0.5),
	)
	if err != nil {
		t.Fatalf("NewScorer: %v", err)
	}
	algos := s.Algorithms()
	if len(algos) != 2 {
		t.Fatalf("len = %d; want 2", len(algos))
	}
	if algos[0].Weight != 0.25 {
		t.Errorf("algos[0].Weight = %g; want 0.25 (post-normalisation)", algos[0].Weight)
	}
	if algos[1].Weight != 0.75 {
		t.Errorf("algos[1].Weight = %g; want 0.75 (post-normalisation)", algos[1].Weight)
	}
}

// TestScorer_ScoreAll_Keys verifies that ScoreAll returns a map whose
// keyset is EXACTLY the configured algorithm set. A 3-algorithm Scorer
// produces a map of length 3 containing exactly those AlgoIDs as keys.
func TestScorer_ScoreAll_Keys(t *testing.T) {
	t.Parallel()
	s, err := fuzzymatch.NewScorer(
		fuzzymatch.WithAlgorithm(fuzzymatch.AlgoLevenshtein, 1.0),
		fuzzymatch.WithAlgorithm(fuzzymatch.AlgoJaroWinkler, 1.0),
		fuzzymatch.WithAlgorithm(fuzzymatch.AlgoTokenJaccard, 1.0),
		fuzzymatch.WithThreshold(0.5),
	)
	if err != nil {
		t.Fatalf("NewScorer: %v", err)
	}
	got := s.ScoreAll("kitten", "sitting")
	if len(got) != 3 {
		t.Fatalf("len(ScoreAll) = %d; want 3", len(got))
	}
	wantKeys := []fuzzymatch.AlgoID{
		fuzzymatch.AlgoLevenshtein,
		fuzzymatch.AlgoJaroWinkler,
		fuzzymatch.AlgoTokenJaccard,
	}
	for _, k := range wantKeys {
		if _, ok := got[k]; !ok {
			t.Errorf("ScoreAll missing key %v", k)
		}
	}
}

// TestScorer_ScoreAll_ValuesMatchPerAlgoCalls verifies that ScoreAll's
// map values are byte-identical to the per-algorithm score function
// called on the same (pre-normalised) inputs. We use a single-
// Levenshtein Scorer with WithoutNormalisation so the comparison is
// trivially against LevenshteinScore on raw inputs.
func TestScorer_ScoreAll_ValuesMatchPerAlgoCalls(t *testing.T) {
	t.Parallel()
	s, err := fuzzymatch.NewScorer(
		fuzzymatch.WithAlgorithm(fuzzymatch.AlgoLevenshtein, 1.0),
		fuzzymatch.WithoutNormalisation(),
		fuzzymatch.WithThreshold(0.5),
	)
	if err != nil {
		t.Fatalf("NewScorer: %v", err)
	}
	got := s.ScoreAll("kitten", "sitting")
	want := fuzzymatch.LevenshteinScore("kitten", "sitting")
	if got[fuzzymatch.AlgoLevenshtein] != want {
		t.Errorf(
			"ScoreAll[AlgoLevenshtein] = %g; want %g (must equal raw LevenshteinScore on raw inputs)",
			got[fuzzymatch.AlgoLevenshtein], want,
		)
	}
}

// TestScorer_ScoreAll_FreshMap verifies that ScoreAll allocates a fresh
// map on every call (per spec §8.6). Mutating the returned map MUST
// NOT affect subsequent ScoreAll calls.
func TestScorer_ScoreAll_FreshMap(t *testing.T) {
	t.Parallel()
	s, err := fuzzymatch.NewScorer(
		fuzzymatch.WithAlgorithm(fuzzymatch.AlgoLevenshtein, 1.0),
		fuzzymatch.WithThreshold(0.5),
	)
	if err != nil {
		t.Fatalf("NewScorer: %v", err)
	}
	m1 := s.ScoreAll("kitten", "sitting")
	m1[fuzzymatch.AlgoLevenshtein] = 999.0
	m2 := s.ScoreAll("kitten", "sitting")
	if m2[fuzzymatch.AlgoLevenshtein] == 999.0 {
		t.Errorf(
			"second ScoreAll returned mutated value 999.0; want a fresh map (got %g)",
			m2[fuzzymatch.AlgoLevenshtein],
		)
	}
}

// TestDefaultScorer_Composition verifies the spec §8.5 / CONTEXT.md §6
// canonical composition for DefaultScorer(): exactly six algorithms
// (DamerauLevenshteinOSA, JaroWinkler, TokenJaccard, QGramJaccard,
// SorensenDice, DoubleMetaphone) and threshold exactly 0.85.
func TestDefaultScorer_Composition(t *testing.T) {
	t.Parallel()
	s := fuzzymatch.DefaultScorer()
	if s == nil {
		t.Fatal("DefaultScorer returned nil")
	}
	if got := s.Threshold(); got != 0.85 {
		t.Errorf("Threshold = %g; want 0.85", got)
	}
	algos := s.Algorithms()
	if len(algos) != 6 {
		t.Fatalf("len(Algorithms) = %d; want 6", len(algos))
	}
	wantSet := map[fuzzymatch.AlgoID]bool{
		fuzzymatch.AlgoDamerauLevenshteinOSA: true,
		fuzzymatch.AlgoJaroWinkler:           true,
		fuzzymatch.AlgoQGramJaccard:          true,
		fuzzymatch.AlgoSorensenDice:          true,
		fuzzymatch.AlgoTokenJaccard:          true,
		fuzzymatch.AlgoDoubleMetaphone:       true,
	}
	got := make(map[fuzzymatch.AlgoID]bool, 6)
	for _, a := range algos {
		got[a.ID] = true
	}
	for id := range wantSet {
		if !got[id] {
			t.Errorf("DefaultScorer composition missing %v", id)
		}
	}
	for id := range got {
		if !wantSet[id] {
			t.Errorf("DefaultScorer composition contains unexpected %v", id)
		}
	}
}

// TestDefaultScorer_WeightsEqual verifies that each post-normalisation
// weight in DefaultScorer's six-algorithm composition equals 1.0/6.0
// (six equal raw weights → uniform normalised weights). Comparison is
// exact == because the same dividend / divisor pair produces the same
// IEEE-754 quotient on every call.
func TestDefaultScorer_WeightsEqual(t *testing.T) {
	t.Parallel()
	s := fuzzymatch.DefaultScorer()
	algos := s.Algorithms()
	if len(algos) != 6 {
		t.Fatalf("len(Algorithms) = %d; want 6", len(algos))
	}
	want := 1.0 / 6.0
	for i, a := range algos {
		if a.Weight != want {
			t.Errorf("algos[%d].Weight = %g; want %g (1.0/6.0)", i, a.Weight, want)
		}
	}
}

// TestDefaultScorer_NeverFails verifies that DefaultScorer() can be
// invoked repeatedly without panicking and without returning nil. The
// godoc contract is that DefaultScorer cannot fail under normal
// operation; this test exercises the contract under 100 sequential
// calls.
func TestDefaultScorer_NeverFails(t *testing.T) {
	t.Parallel()
	for i := 0; i < 100; i++ {
		func() {
			defer func() {
				if r := recover(); r != nil {
					t.Errorf("iteration %d: DefaultScorer panicked: %v", i, r)
				}
			}()
			s := fuzzymatch.DefaultScorer()
			if s == nil {
				t.Errorf("iteration %d: DefaultScorer returned nil", i)
			}
		}()
	}
}

// TestDefaultScorerOptions_FreshSlice verifies that DefaultScorerOptions
// returns a fresh slice on every call: mutating the returned slice MUST
// NOT affect a subsequent call's contents, and the mutated state MUST
// NOT propagate into DefaultScorer's internal construction either.
func TestDefaultScorerOptions_FreshSlice(t *testing.T) {
	t.Parallel()
	opts1 := fuzzymatch.DefaultScorerOptions()
	if len(opts1) == 0 {
		t.Fatalf("len(opts1) = 0; want > 0")
	}
	// Mutate the first slice (replace the first option with nil).
	opts1[0] = nil

	opts2 := fuzzymatch.DefaultScorerOptions()
	if len(opts2) != len(opts1) {
		t.Errorf("len(opts2) = %d; want %d (same length)", len(opts2), len(opts1))
	}
	if opts2[0] == nil {
		t.Errorf("opts2[0] is nil; want a fresh non-nil ScorerOption (the first slice's mutation must not propagate)")
	}
	// NewScorer should still succeed with a freshly-obtained options
	// slice — the mutation of opts1 must not have corrupted the
	// underlying composition.
	if _, err := fuzzymatch.NewScorer(opts2...); err != nil {
		t.Errorf("NewScorer with fresh DefaultScorerOptions: %v", err)
	}
}

// TestDefaultScorerOptions_ProducesEquivalentScorer verifies that
// `NewScorer(DefaultScorerOptions()...)` produces a Scorer
// behaviourally identical to `DefaultScorer()`. We compare the
// composite Score on a representative identifier-style input pair
// — the float64 results must be byte-identical.
func TestDefaultScorerOptions_ProducesEquivalentScorer(t *testing.T) {
	t.Parallel()
	s1 := fuzzymatch.DefaultScorer()
	s2, err := fuzzymatch.NewScorer(fuzzymatch.DefaultScorerOptions()...)
	if err != nil {
		t.Fatalf("NewScorer(DefaultScorerOptions()...): %v", err)
	}
	const a, b = "user_id", "userId"
	got1 := s1.Score(a, b)
	got2 := s2.Score(a, b)
	if got1 != got2 {
		t.Errorf(
			"DefaultScorer.Score(%q,%q)=%g, NewScorer(DefaultScorerOptions()...).Score(%q,%q)=%g (must be byte-identical)",
			a, b, got1, a, b, got2,
		)
	}
	// Threshold and algorithm count also match.
	if s1.Threshold() != s2.Threshold() {
		t.Errorf("threshold mismatch: s1=%g, s2=%g", s1.Threshold(), s2.Threshold())
	}
	if len(s1.Algorithms()) != len(s2.Algorithms()) {
		t.Errorf("len(Algorithms) mismatch: s1=%d, s2=%d",
			len(s1.Algorithms()), len(s2.Algorithms()))
	}
}

// TestDefaultScorer_WithoutAlgorithm_Composition verifies the documented
// composition pattern:
//
//	opts := append(DefaultScorerOptions(), WithoutAlgorithm(AlgoDoubleMetaphone))
//	NewScorer(opts...)
//
// produces a Scorer with the six default algorithms minus DoubleMetaphone,
// i.e. exactly five. The threshold from the default composition (0.85)
// survives (no WithThreshold override in the example, so it stays).
func TestDefaultScorer_WithoutAlgorithm_Composition(t *testing.T) {
	t.Parallel()
	opts := append(fuzzymatch.DefaultScorerOptions(),
		fuzzymatch.WithoutAlgorithm(fuzzymatch.AlgoDoubleMetaphone),
	)
	s, err := fuzzymatch.NewScorer(opts...)
	if err != nil {
		t.Fatalf("NewScorer: %v", err)
	}
	algos := s.Algorithms()
	if len(algos) != 5 {
		t.Fatalf("len(Algorithms) = %d; want 5", len(algos))
	}
	for _, a := range algos {
		if a.ID == fuzzymatch.AlgoDoubleMetaphone {
			t.Errorf("DoubleMetaphone still present after WithoutAlgorithm: %v", a.ID)
		}
	}
	// Threshold inherited from DefaultScorerOptions.
	if got := s.Threshold(); got != 0.85 {
		t.Errorf("threshold = %g; want 0.85 (inherited from DefaultScorerOptions)", got)
	}
}

// TestScorer_ScoreAll_PreNormalises verifies that ScoreAll applies the
// same normalisation gate as Score: when normalisation is enabled
// (default), the per-algorithm scores in ScoreAll reflect the algorithm
// invocation on the NORMALISED inputs. We compare two Scorers (with /
// without normalisation) on identifier-style input — the with-norm
// Levenshtein score is strictly greater than the without-norm score.
func TestScorer_ScoreAll_PreNormalises(t *testing.T) {
	t.Parallel()
	withNorm, err := fuzzymatch.NewScorer(
		fuzzymatch.WithAlgorithm(fuzzymatch.AlgoLevenshtein, 1.0),
		fuzzymatch.WithThreshold(0.5),
	)
	if err != nil {
		t.Fatalf("NewScorer (with-norm): %v", err)
	}
	withoutNorm, err := fuzzymatch.NewScorer(
		fuzzymatch.WithAlgorithm(fuzzymatch.AlgoLevenshtein, 1.0),
		fuzzymatch.WithoutNormalisation(),
		fuzzymatch.WithThreshold(0.5),
	)
	if err != nil {
		t.Fatalf("NewScorer (without-norm): %v", err)
	}
	const a, b = "XMLParser", "xml_parser"
	with := withNorm.ScoreAll(a, b)
	without := withoutNorm.ScoreAll(a, b)
	if !(without[fuzzymatch.AlgoLevenshtein] < with[fuzzymatch.AlgoLevenshtein]) {
		t.Errorf(
			"ScoreAll must reflect normalisation gate: with=%g, without=%g (a=%q, b=%q)",
			with[fuzzymatch.AlgoLevenshtein],
			without[fuzzymatch.AlgoLevenshtein],
			a, b,
		)
	}
}
