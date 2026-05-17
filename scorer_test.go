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
