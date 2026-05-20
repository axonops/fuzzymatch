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
	"math"
	"sync"
	"testing"
	"testing/quick"

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

// TestScorer_NormalisationOptions_Default pins the accessor contract
// for the default Scorer: DefaultScorer() applies normalisation with
// DefaultNormalisationOptions, so the accessor returns
// (DefaultNormalisationOptions(), true). Used by the scan sub-package
// (09-RESEARCH.md Open Question 1 resolution) to canonicalise
// SuppressedPairs and build token buckets using the same normalisation
// pipeline the Scorer uses for scoring.
func TestScorer_NormalisationOptions_Default(t *testing.T) {
	t.Parallel()

	s := fuzzymatch.DefaultScorer()
	opts, applied := s.NormalisationOptions()

	if !applied {
		t.Errorf("applied: got false; want true (DefaultScorer applies normalisation)")
	}
	want := fuzzymatch.DefaultNormalisationOptions()
	if opts != want {
		t.Errorf("opts: got %+v; want %+v (DefaultNormalisationOptions)", opts, want)
	}
}

// TestScorer_NormalisationOptions_WithoutNormalisation verifies the
// accessor reports applied == false when the Scorer was constructed
// with WithoutNormalisation. The returned opts value is the previously
// stored NormalisationOptions (i.e. DefaultNormalisationOptions in
// this case because no explicit WithNormalisation was applied) —
// callers seeing applied == false should ignore opts and pass raw
// inputs downstream, mirroring what the Scorer itself does on the
// WithoutNormalisation path. This contract is documented in the
// accessor's godoc.
func TestScorer_NormalisationOptions_WithoutNormalisation(t *testing.T) {
	t.Parallel()

	s, err := fuzzymatch.NewScorer(
		fuzzymatch.WithAlgorithm(fuzzymatch.AlgoLevenshtein, 1.0),
		fuzzymatch.WithThreshold(0.85),
		fuzzymatch.WithoutNormalisation(),
	)
	if err != nil {
		t.Fatalf("NewScorer: %v", err)
	}
	_, applied := s.NormalisationOptions()

	if applied {
		t.Errorf("applied: got true; want false (WithoutNormalisation)")
	}
	// opts is intentionally not asserted: scorer_options.go's
	// WithoutNormalisation only flips applyNorm=false and leaves the
	// previously-stored normOpts value as-is (per scorer_options.go
	// lines 262–266). Callers seeing applied=false should ignore opts.
}

// TestScorer_NormalisationOptions_WithoutNormalisation_AfterCustom
// pins the "later option wins" semantics: WithNormalisation then
// WithoutNormalisation leaves applied=false but the previously stored
// opts struct survives (per scorer_options.go lines 262–266 "applyNorm
// becomes false but the previously-stored normOpts value is
// intentionally not cleared"). The accessor surfaces that exact state.
func TestScorer_NormalisationOptions_WithoutNormalisation_AfterCustom(t *testing.T) {
	t.Parallel()

	custom := fuzzymatch.NormalisationOptions{
		Lowercase:       true,
		StripSeparators: false,
		SeparatorChars:  "_",
		SplitCamelCase:  false,
		NFC:             true,
		StripDiacritics: false,
	}
	s, err := fuzzymatch.NewScorer(
		fuzzymatch.WithAlgorithm(fuzzymatch.AlgoLevenshtein, 1.0),
		fuzzymatch.WithThreshold(0.85),
		fuzzymatch.WithNormalisation(custom),
		fuzzymatch.WithoutNormalisation(),
	)
	if err != nil {
		t.Fatalf("NewScorer: %v", err)
	}
	opts, applied := s.NormalisationOptions()

	if applied {
		t.Errorf("applied: got true; want false (later WithoutNormalisation wins)")
	}
	if opts != custom {
		t.Errorf("opts: got %+v; want previously-stored custom %+v (WithoutNormalisation does not clear normOpts)", opts, custom)
	}
}

// TestScorer_NormalisationOptions_WithCustom verifies the accessor
// returns the exact NormalisationOptions struct the consumer supplied
// to WithNormalisation, byte-for-byte. Every documented field is
// covered (Lowercase, StripSeparators, SeparatorChars, SplitCamelCase,
// NFC, StripDiacritics) to catch any future field addition that the
// accessor accidentally drops.
func TestScorer_NormalisationOptions_WithCustom(t *testing.T) {
	t.Parallel()

	custom := fuzzymatch.NormalisationOptions{
		Lowercase:       false,
		StripSeparators: true,
		SeparatorChars:  "|:/",
		SplitCamelCase:  false,
		NFC:             true,
		StripDiacritics: true,
	}
	s, err := fuzzymatch.NewScorer(
		fuzzymatch.WithAlgorithm(fuzzymatch.AlgoLevenshtein, 1.0),
		fuzzymatch.WithThreshold(0.85),
		fuzzymatch.WithNormalisation(custom),
	)
	if err != nil {
		t.Fatalf("NewScorer: %v", err)
	}
	opts, applied := s.NormalisationOptions()

	if !applied {
		t.Errorf("applied: got false; want true (WithNormalisation)")
	}
	if opts != custom {
		t.Errorf("opts: got %+v; want %+v (byte-for-byte equality)", opts, custom)
	}
}

// TestScorer_NormalisationOptions_ByValue verifies the return is
// by-value (NOT by reference). A caller mutating the returned struct
// MUST NOT observe the mutation on a subsequent NormalisationOptions
// call. This pins the immutability contract documented in the
// accessor's godoc — mirrors the Threshold() and Algorithms() patterns
// where Scorer state is never exposed to caller mutation.
func TestScorer_NormalisationOptions_ByValue(t *testing.T) {
	t.Parallel()

	s := fuzzymatch.DefaultScorer()
	first, _ := s.NormalisationOptions()

	// Mutate the returned struct's fields. The mutation is observable
	// in `first` itself but MUST NOT propagate into the Scorer.
	first.Lowercase = !first.Lowercase
	first.StripSeparators = !first.StripSeparators
	first.SeparatorChars = "X"
	first.SplitCamelCase = !first.SplitCamelCase
	first.NFC = !first.NFC
	first.StripDiacritics = !first.StripDiacritics

	second, _ := s.NormalisationOptions()
	want := fuzzymatch.DefaultNormalisationOptions()
	if second != want {
		t.Errorf("second call: got %+v; want %+v (mutation of first return must not affect Scorer)", second, want)
	}
}

// TestScorer_NormalisationOptions_Concurrent pins the concurrent-safety
// guarantee: NormalisationOptions is a read-only accessor; concurrent
// invocation from many goroutines must not race and must observe
// identical results. The race detector (`go test -race`) catches any
// hidden write to the Scorer's state during this loop.
func TestScorer_NormalisationOptions_Concurrent(t *testing.T) {
	t.Parallel()

	s := fuzzymatch.DefaultScorer()
	want, wantApplied := s.NormalisationOptions()

	var wg sync.WaitGroup
	const goroutines = 100
	wg.Add(goroutines)
	for i := 0; i < goroutines; i++ {
		go func() {
			defer wg.Done()
			got, gotApplied := s.NormalisationOptions()
			if got != want || gotApplied != wantApplied {
				t.Errorf("concurrent read: got (%+v, %v); want (%+v, %v)", got, gotApplied, want, wantApplied)
			}
		}()
	}
	wg.Wait()
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

// TestDefaultScorer_NeverPanics_PropertyTest is the Phase 8.5 Gap 5
// companion property test pinning the contract that DefaultScorer never
// reaches its panic site at scorer.go:592 (the
// ErrInternalInvariantViolated wrap landed by Plan 01 — a defence-in-
// depth assertion that the locked default composition cannot drift out
// of sync with NewScorer's validation pipeline).
//
// The panic path is "dead code by construction": DefaultScorerOptions()
// returns a stable, locked option set that DefaultScorer applies to
// NewScorer. The property under test is that the panic NEVER fires for:
//
//   - The plain DefaultScorerOptions() composition.
//   - Reasonable WithoutAlgorithm subsets of DefaultScorerOptions
//     (removing any one of the six default algorithms).
//   - Reasonable WithoutAlgorithm subsets removing an algorithm NOT in
//     the default composition (silent no-op semantic — Gap 7).
//
// testing/quick drives 100 random selections of which default algorithm
// to drop (or to attempt to drop when absent). Any panic from
// DefaultScorer() or NewScorer(append(DefaultScorerOptions(), ...))
// would fail the test. The locked test name is the load-bearing
// identifier for Phase 8.5 Gap 5's compliance gate (CONTEXT.md).
func TestDefaultScorer_NeverPanics_PropertyTest(t *testing.T) {
	t.Parallel()

	// Helper: assert no panic when invoking the supplied closure.
	mustNotPanic := func(t *testing.T, label string, fn func()) {
		t.Helper()
		defer func() {
			if r := recover(); r != nil {
				t.Errorf("%s panicked: %v", label, r)
			}
		}()
		fn()
	}

	// 1. Bare DefaultScorer() returns a non-nil *Scorer without panic.
	mustNotPanic(t, "DefaultScorer()", func() {
		s := fuzzymatch.DefaultScorer()
		if s == nil {
			t.Errorf("DefaultScorer() returned nil")
		}
	})

	// 2. Property: append(DefaultScorerOptions(), WithoutAlgorithm(any default
	//    AlgoID)) constructs without panic. The closure picks one of the six
	//    default algorithms by index modulo len(defaults) so quick.Check
	//    exercises every removal pattern.
	defaults := []fuzzymatch.AlgoID{
		fuzzymatch.AlgoDamerauLevenshteinOSA,
		fuzzymatch.AlgoJaroWinkler,
		fuzzymatch.AlgoTokenJaccard,
		fuzzymatch.AlgoQGramJaccard,
		fuzzymatch.AlgoSorensenDice,
		fuzzymatch.AlgoDoubleMetaphone,
	}
	propRemoveDefault := func(idx uint8) bool {
		i := int(idx) % len(defaults)
		victim := defaults[i]
		var ok bool
		func() {
			defer func() {
				if r := recover(); r != nil {
					t.Errorf("removal of default %s panicked: %v", victim, r)
				}
			}()
			opts := append(
				fuzzymatch.DefaultScorerOptions(),
				fuzzymatch.WithoutAlgorithm(victim),
			)
			s, err := fuzzymatch.NewScorer(opts...)
			if err != nil {
				t.Errorf("NewScorer after removing %s: %v", victim, err)
				return
			}
			if s == nil {
				t.Errorf("NewScorer after removing %s returned nil", victim)
				return
			}
			ok = true
		}()
		return ok
	}
	if err := quick.Check(propRemoveDefault, &quick.Config{MaxCount: 100}); err != nil {
		t.Errorf("propRemoveDefault: %v", err)
	}

	// 3. Property: append(DefaultScorerOptions(), WithoutAlgorithm(non-default))
	//    is a silent no-op (Gap 7) — construction succeeds without panic and
	//    Algorithms is unchanged. AlgoCosine is intentionally not part of the
	//    default composition.
	propRemoveNonDefault := func() bool {
		var ok bool
		func() {
			defer func() {
				if r := recover(); r != nil {
					t.Errorf("removal of absent AlgoCosine panicked: %v", r)
				}
			}()
			opts := append(
				fuzzymatch.DefaultScorerOptions(),
				fuzzymatch.WithoutAlgorithm(fuzzymatch.AlgoCosine),
			)
			s, err := fuzzymatch.NewScorer(opts...)
			if err != nil {
				t.Errorf("NewScorer with WithoutAlgorithm(AlgoCosine): %v", err)
				return
			}
			if s == nil {
				t.Errorf("NewScorer returned nil")
				return
			}
			ok = true
		}()
		return ok
	}
	if err := quick.Check(propRemoveNonDefault, &quick.Config{MaxCount: 100}); err != nil {
		t.Errorf("propRemoveNonDefault: %v", err)
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
	opts := append(
		fuzzymatch.DefaultScorerOptions(),
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

// ---------------------------------------------------------------------
// Property tests (testing/quick)
// ---------------------------------------------------------------------
//
// These tests verify mathematical invariants that hold across an
// arbitrary input distribution. The standard library's testing/quick
// generates 100 random inputs per check by default (the project
// constraint is stdlib only — no rapid, no gopter). For invariants
// where the default generator is sufficient (string inputs to Score),
// we use the default. For invariants that require structured input
// (random weight vectors with constrained positivity), we drive the
// generator with a wrapper function whose signature is consumed by
// quick.Check.

// TestProp_Scorer_DeterministicAcrossRuns verifies that DefaultScorer()
// is deterministic across CONSTRUCTION events: building two FRESH
// DefaultScorer instances and calling Score(a, b) on each returns
// byte-identical float64. This is stronger than "deterministic across
// CALLS on the same Scorer instance" (which scorer_test.go's plan
// 08-02 TestScorer_Score_DeterministicAcrossCalls already proves);
// here we re-build the Scorer between invocations to ensure no
// construction-side state seeps into the result.
func TestProp_Scorer_DeterministicAcrossRuns(t *testing.T) {
	t.Parallel()
	f := func(a, b string) bool {
		// Construct two fresh Scorers; both should produce byte-identical
		// floats on the same input pair.
		s1 := fuzzymatch.DefaultScorer()
		s2 := fuzzymatch.DefaultScorer()
		return s1.Score(a, b) == s2.Score(a, b)
	}
	if err := quick.Check(f, nil); err != nil {
		t.Errorf("PropScorer_DeterministicAcrossRuns: %v", err)
	}
}

// TestProp_Scorer_WeightSumOne verifies the SCORER-03 invariant: when
// WithNormaliseWeights(true) (the default), the post-normalisation
// weights of every algorithm entry in the Scorer sum to exactly 1.0
// (within a small tolerance for many-algorithm float drift).
//
// We exercise two regimes: (a) fixed scenarios with known dyadic and
// non-dyadic weights to pin the tolerance; (b) a quick.Check pass over
// random positive weight vectors to provide breadth.
func TestProp_Scorer_WeightSumOne(t *testing.T) {
	t.Parallel()

	// Fixed scenarios — explicit weight sets.
	fixed := []struct {
		name    string
		weights []float64
		algos   []fuzzymatch.AlgoID
	}{
		{
			name:    "two dyadic weights (1, 3)",
			weights: []float64{1.0, 3.0},
			algos:   []fuzzymatch.AlgoID{fuzzymatch.AlgoLevenshtein, fuzzymatch.AlgoJaroWinkler},
		},
		{
			name:    "six equal weights (DefaultScorer composition)",
			weights: []float64{1.0, 1.0, 1.0, 1.0, 1.0, 1.0},
			algos: []fuzzymatch.AlgoID{
				fuzzymatch.AlgoDamerauLevenshteinOSA,
				fuzzymatch.AlgoJaroWinkler,
				fuzzymatch.AlgoTokenJaccard,
				fuzzymatch.AlgoQGramJaccard,
				fuzzymatch.AlgoSorensenDice,
				fuzzymatch.AlgoDoubleMetaphone,
			},
		},
		{
			name:    "three non-dyadic weights (0.7, 0.001, 1000)",
			weights: []float64{0.7, 0.001, 1000.0},
			algos: []fuzzymatch.AlgoID{
				fuzzymatch.AlgoLevenshtein,
				fuzzymatch.AlgoJaroWinkler,
				fuzzymatch.AlgoTokenJaccard,
			},
		},
	}
	for _, c := range fixed {
		c := c
		t.Run(c.name, func(t *testing.T) {
			t.Parallel()
			opts := make([]fuzzymatch.ScorerOption, 0, len(c.weights)+1)
			for i, w := range c.weights {
				opts = append(opts, fuzzymatch.WithAlgorithm(c.algos[i], w))
			}
			opts = append(opts, fuzzymatch.WithThreshold(0.5))
			s, err := fuzzymatch.NewScorer(opts...)
			if err != nil {
				t.Fatalf("NewScorer: %v", err)
			}
			algos := s.Algorithms()
			var sum float64
			for _, a := range algos {
				// Explicit `sum = sum + …` per DET-06 / CONTEXT.md §5
				// left-to-right additive accumulation — the same pattern
				// scorer.go's reduction loop uses; explicit form is the
				// determinism contract.
				sum = sum + a.Weight //nolint:gocritic // DET-06 locked left-to-right additive accumulation pattern; explicit form is the contract per CONTEXT.md §5
			}
			if math.Abs(sum-1.0) >= 1e-12 {
				t.Errorf("weight sum: got %g (diff %g); want 1.0 ± 1e-12", sum, sum-1.0)
			}
		})
	}

	// Random-vector quick.Check: random 1-3 positive weights drawn from
	// uint16 (mapped into a reasonable positive float64 range).
	f := func(w0, w1, w2 uint16) bool {
		// Bias the inputs into [0.01, 100.0]: small positive floats, no
		// zeros (zero weight is rejected by ErrInvalidWeight at the
		// option layer). Use three algorithms picked deterministically.
		toPositive := func(u uint16) float64 {
			// Avoid zero by adding 1 in the numerator; range becomes
			// (0, 100].
			//
			// Q12b uint16 overflow fix: widen u to uint32 BEFORE the
			// `+ 1` so the addition cannot wrap. The previous form
			// `float64(u+1)` wrapped to 0 when u == 65535 (the uint16
			// addition saturates inside the uint16 type before the
			// float64 conversion), producing weight == 0 which is
			// rejected by ErrInvalidWeight at the option layer — a
			// flaky failure on the single uint16-max seed.
			return float64(uint32(u)+1) / float64(uint32(1)<<16) * 100.0
		}
		w := []float64{toPositive(w0), toPositive(w1), toPositive(w2)}
		algos := []fuzzymatch.AlgoID{
			fuzzymatch.AlgoLevenshtein,
			fuzzymatch.AlgoJaroWinkler,
			fuzzymatch.AlgoTokenJaccard,
		}
		opts := []fuzzymatch.ScorerOption{
			fuzzymatch.WithAlgorithm(algos[0], w[0]),
			fuzzymatch.WithAlgorithm(algos[1], w[1]),
			fuzzymatch.WithAlgorithm(algos[2], w[2]),
			fuzzymatch.WithThreshold(0.5),
		}
		s, err := fuzzymatch.NewScorer(opts...)
		if err != nil {
			return false // constructor failure on positive weights is itself a bug
		}
		var sum float64
		for _, a := range s.Algorithms() {
			sum = sum + a.Weight //nolint:gocritic // DET-06 locked left-to-right additive accumulation pattern; explicit form is the contract per CONTEXT.md §5
		}
		return math.Abs(sum-1.0) < 1e-12
	}
	if err := quick.Check(f, nil); err != nil {
		t.Errorf("PropScorer_WeightSumOne (random vectors): %v", err)
	}
}

// TestProp_Scorer_ScoreInRange verifies the SCORER-04 invariant: with
// normalised weights AND per-algorithm scores in [0, 1] (every
// algorithm in the catalogue satisfies this), the composite Score is
// in [0.0, 1.0] for any input pair (a, b).
func TestProp_Scorer_ScoreInRange(t *testing.T) {
	t.Parallel()
	s := fuzzymatch.DefaultScorer()
	f := func(a, b string) bool {
		score := s.Score(a, b)
		return score >= 0.0 && score <= 1.0
	}
	if err := quick.Check(f, nil); err != nil {
		t.Errorf("PropScorer_ScoreInRange: %v", err)
	}
}

// TestProp_Scorer_NoNaN_NoInf verifies that Score never produces a
// non-finite value. NaN or Inf would propagate through the threshold
// comparison in Match and through any consumer's downstream
// arithmetic; the contract is that Score returns a finite float64 in
// [0, 1] for every well-formed Scorer.
func TestProp_Scorer_NoNaN_NoInf(t *testing.T) {
	t.Parallel()
	s := fuzzymatch.DefaultScorer()
	f := func(a, b string) bool {
		score := s.Score(a, b)
		return !math.IsNaN(score) && !math.IsInf(score, 0)
	}
	if err := quick.Check(f, nil); err != nil {
		t.Errorf("PropScorer_NoNaN_NoInf: %v", err)
	}
}

// TestScorer_ConcurrentSafety verifies the SCORER-01 contract: a
// *Scorer is safe for concurrent use from any number of goroutines
// without external synchronisation. All Score, ScoreAll, and Match
// methods are exercised — 100 goroutines per method, three rounds.
//
// The test asserts both correctness (every goroutine produces the
// same result as the first) AND, when run under `go test -race`,
// non-existence of a data race. The Scorer struct fields are written
// once in NewScorer and never written again; the methods are
// read-only.
//
// stdlib `sync.WaitGroup` is the synchronisation primitive — no
// errgroup (the root module is stdlib-only; `golang.org/x/sync` is
// not in go.mod).
func TestScorer_ConcurrentSafety(t *testing.T) {
	s := fuzzymatch.DefaultScorer()
	const n = 100
	const a, b = "user_id", "userId"

	t.Run("Score", func(t *testing.T) {
		var wg sync.WaitGroup
		wg.Add(n)
		results := make([]float64, n)
		for i := 0; i < n; i++ {
			go func(idx int) {
				defer wg.Done()
				results[idx] = s.Score(a, b)
			}(i)
		}
		wg.Wait()
		for i, got := range results {
			if got != results[0] {
				t.Errorf("goroutine %d: Score = %g; want %g (first goroutine's result)", i, got, results[0])
			}
		}
	})

	t.Run("ScoreAll", func(t *testing.T) {
		var wg sync.WaitGroup
		wg.Add(n)
		results := make([]map[fuzzymatch.AlgoID]float64, n)
		for i := 0; i < n; i++ {
			go func(idx int) {
				defer wg.Done()
				results[idx] = s.ScoreAll(a, b)
			}(i)
		}
		wg.Wait()
		// All maps must have the same length AND identical per-key
		// values. Comparison is exact (==) because the scores are
		// deterministic per call.
		want := results[0]
		if len(want) != 6 {
			t.Fatalf("first goroutine: len(ScoreAll) = %d; want 6", len(want))
		}
		for i := 0; i < n; i++ {
			got := results[i]
			if len(got) != len(want) {
				t.Errorf("goroutine %d: len(ScoreAll) = %d; want %d", i, len(got), len(want))
				continue
			}
			for id, wantValue := range want {
				gotValue, ok := got[id]
				if !ok {
					t.Errorf("goroutine %d: missing key %v", i, id)
					continue
				}
				if gotValue != wantValue {
					t.Errorf("goroutine %d: ScoreAll[%v] = %g; want %g", i, id, gotValue, wantValue)
				}
			}
		}
	})

	t.Run("Match", func(t *testing.T) {
		var wg sync.WaitGroup
		wg.Add(n)
		results := make([]bool, n)
		for i := 0; i < n; i++ {
			go func(idx int) {
				defer wg.Done()
				results[idx] = s.Match(a, b)
			}(i)
		}
		wg.Wait()
		for i, got := range results {
			if got != results[0] {
				t.Errorf("goroutine %d: Match = %v; want %v", i, got, results[0])
			}
		}
	})
}

// TestScorer_ScoreAndAll_MatchesScoreAndScoreAll verifies the
// combined-method invariant: ScoreAndAll(a, b) returns the same
// composite as Score(a, b) and the same per-algorithm breakdown as
// ScoreAll(a, b), bit-identical. Closes the coverage floor for
// scorer.go after the Phase 9 ScoreAndAll addition.
func TestScorer_ScoreAndAll_MatchesScoreAndScoreAll(t *testing.T) {
	t.Parallel()
	s := fuzzymatch.DefaultScorer()
	pairs := []struct{ a, b string }{
		{"user_id", "userId"},
		{"customer", "customer"},
		{"is_deleted", "is_active"},
		{"", "x"},
		{"abc", ""},
	}
	for _, p := range pairs {
		gotScore, gotBreakdown := s.ScoreAndAll(p.a, p.b)
		wantScore := s.Score(p.a, p.b)
		wantBreakdown := s.ScoreAll(p.a, p.b)
		if gotScore != wantScore {
			t.Errorf("ScoreAndAll(%q,%q) composite: got %v want %v", p.a, p.b, gotScore, wantScore)
		}
		if len(gotBreakdown) != len(wantBreakdown) {
			t.Errorf("ScoreAndAll(%q,%q) map length: got %d want %d", p.a, p.b, len(gotBreakdown), len(wantBreakdown))
		}
		for id, v := range wantBreakdown {
			if gotBreakdown[id] != v {
				t.Errorf("ScoreAndAll(%q,%q) breakdown[%v]: got %v want %v", p.a, p.b, id, gotBreakdown[id], v)
			}
		}
	}
}

// TestScorer_ScoreAndAll_FreshMapPerCall verifies that each call
// returns an independently-allocated map (the breakdown contract
// must mirror ScoreAll's freshly-allocated guarantee).
func TestScorer_ScoreAndAll_FreshMapPerCall(t *testing.T) {
	t.Parallel()
	s := fuzzymatch.DefaultScorer()
	_, m1 := s.ScoreAndAll("user_id", "userId")
	_, m2 := s.ScoreAndAll("user_id", "userId")
	// Same input → same contents.
	if len(m1) != len(m2) {
		t.Fatalf("breakdown len: m1=%d m2=%d", len(m1), len(m2))
	}
	for id, v := range m1 {
		if m2[id] != v {
			t.Errorf("breakdown[%v]: m1=%v m2=%v", id, v, m2[id])
		}
	}
	// Independent allocation: mutating one must not affect the other.
	for id := range m1 {
		m1[id] = -1
		break
	}
	for id, v := range m2 {
		if v == -1 {
			t.Errorf("m2 mutated by m1 write at AlgoID %v: maps share backing storage", id)
		}
	}
}

// TestScorer_ScoreAndAll_WithoutNormalisation verifies the
// pre-normalisation boundary toggle is honoured (matching Score and
// ScoreAll's behaviour under WithoutNormalisation).
func TestScorer_ScoreAndAll_WithoutNormalisation(t *testing.T) {
	t.Parallel()
	s, err := fuzzymatch.NewScorer(
		fuzzymatch.WithThreshold(0.5),
		fuzzymatch.WithAlgorithm(fuzzymatch.AlgoLevenshtein, 1.0),
		fuzzymatch.WithoutNormalisation(),
	)
	if err != nil {
		t.Fatalf("NewScorer: %v", err)
	}
	// Case-sensitive comparison because normalisation is disabled.
	score, breakdown := s.ScoreAndAll("FOO", "foo")
	if score == 1.0 {
		t.Errorf("WithoutNormalisation: case-sensitive compare should NOT score 1.0")
	}
	if breakdown[fuzzymatch.AlgoLevenshtein] != score {
		t.Errorf("single-algorithm composite should equal breakdown value: composite=%v breakdown=%v", score, breakdown[fuzzymatch.AlgoLevenshtein])
	}
}
