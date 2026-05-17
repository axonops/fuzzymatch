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

// scorer_fuzz_test.go runs native Go fuzzing against the three
// Scorer methods that take data inputs: Score, ScoreAll, and Match.
// All harnesses target the package-level DefaultScorer() instance
// (the production canonical six-algorithm composition at threshold
// 0.85) so the fuzz surface mirrors the most-used real-world Scorer.
//
// Properties (per harness, per input pair):
//
//  1. Never panics. Any panic from the Scorer (or any of its six
//     constituent algorithms) propagates to the harness as a fuzz
//     crash. The Q2 data-vs-parameter framework (docs/requirements.md
//     §6.A) explicitly forbids data inputs from triggering a panic —
//     all parameter validation runs at NewScorer construction.
//  2. Score returns a value in [0.0, 1.0]; never NaN; never +/-Inf.
//     DefaultScorer uses WithNormaliseWeights(true) (the default), so
//     the [0, 1] range guarantee holds unconditionally.
//  3. ScoreAll returns a map whose every value is finite in [0, 1]
//     AND whose key set is exactly len(DefaultScorer.Algorithms()).
//     The deterministic-key-set property closes a gap a per-algorithm
//     fuzzer cannot catch: a Scorer that silently drops an algorithm
//     for some adversarial input would be caught here.
//  4. Match returns (Score >= Threshold) — the boolean is a strict
//     function of the float, so any drift between the two methods is
//     a correctness violation.
//
// Seed corpora cover empty inputs, ASCII short, Unicode multi-byte,
// invalid UTF-8, mismatched-length pairs, and a long-input case.
// The on-disk fuzz corpus lives under testdata/fuzz/FuzzScorer_*; CI's
// nightly fuzz job runs these for 5 minutes each.
//
// Threat model: T-08.5-24 (D - DoS via fuzz-discovered panic).
//
// Stdlib testing only.

package fuzzymatch_test

import (
	"math"
	"strings"
	"testing"

	"github.com/axonops/fuzzymatch"
)

// scorerFuzzSeeds is the shared seed-corpus for the three FuzzScorer_*
// harnesses. Each entry is a (a, b) pair covering one input-shape
// class:
//
//   - empty/empty (degenerate)
//   - one-empty (degenerate)
//   - identity (both equal)
//   - ASCII short (typical identifier comparison)
//   - Unicode multi-byte (NFC equivalence path)
//   - invalid UTF-8 (resilience path)
//   - mismatched length (Hamming silent-max path)
//   - long input (boundary-of-budget path)
var scorerFuzzSeeds = []struct{ a, b string }{
	{"", ""},
	{"", "abc"},
	{"abc", ""},
	{"user_id", "user_id"},   // identity
	{"user_id", "userId"},    // snake_case vs camelCase
	{"kitten", "sitting"},    // Wagner-Fischer canonical pair
	{"Schmidt", "Schmit"},    // phonetic equivalence canary
	{"café", "cafe"},         // Latin-supplement
	{"Привет", "привет"},     // Cyrillic, case-difference
	{"\xff\xfe", "abc"},      // invalid UTF-8
	{"hello", "hello world"}, // length mismatch
	{strings.Repeat("a", 200), strings.Repeat("ab", 100)}, // long
}

// FuzzScorer_Score asserts the DefaultScorer's Score method is
// panic-free and returns a finite value in [0.0, 1.0] for arbitrary
// input pairs.
func FuzzScorer_Score(f *testing.F) {
	for _, seed := range scorerFuzzSeeds {
		f.Add(seed.a, seed.b)
	}
	s := fuzzymatch.DefaultScorer()
	f.Fuzz(func(t *testing.T, a, b string) {
		got := s.Score(a, b)
		if math.IsNaN(got) {
			t.Fatalf("Scorer.Score(%q, %q) = NaN", a, b)
		}
		if math.IsInf(got, 0) {
			t.Fatalf("Scorer.Score(%q, %q) = Inf", a, b)
		}
		if got < 0.0 || got > 1.0 {
			t.Fatalf("Scorer.Score(%q, %q) = %g; want in [0,1]", a, b, got)
		}
	})
}

// FuzzScorer_ScoreAll asserts the DefaultScorer's ScoreAll method
// returns a map whose every value is finite in [0.0, 1.0] AND whose
// key set has exactly len(DefaultScorer.Algorithms()) entries — the
// algorithm set is deterministic and must not depend on input.
func FuzzScorer_ScoreAll(f *testing.F) {
	for _, seed := range scorerFuzzSeeds {
		f.Add(seed.a, seed.b)
	}
	s := fuzzymatch.DefaultScorer()
	expectedKeyCount := len(s.Algorithms())
	f.Fuzz(func(t *testing.T, a, b string) {
		got := s.ScoreAll(a, b)
		if len(got) != expectedKeyCount {
			t.Fatalf("Scorer.ScoreAll(%q, %q) returned %d keys; want %d (deterministic key set)",
				a, b, len(got), expectedKeyCount)
		}
		for id, v := range got {
			if math.IsNaN(v) {
				t.Fatalf("Scorer.ScoreAll(%q, %q)[%s] = NaN", a, b, id)
			}
			if math.IsInf(v, 0) {
				t.Fatalf("Scorer.ScoreAll(%q, %q)[%s] = Inf", a, b, id)
			}
			if v < 0.0 || v > 1.0 {
				t.Fatalf("Scorer.ScoreAll(%q, %q)[%s] = %g; want in [0,1]",
					a, b, id, v)
			}
		}
	})
}

// FuzzScorer_Match asserts Match(a, b) == (Score(a, b) >= Threshold())
// for arbitrary inputs. Match is a thin wrapper around Score; any
// drift between the two is a correctness regression.
func FuzzScorer_Match(f *testing.F) {
	for _, seed := range scorerFuzzSeeds {
		f.Add(seed.a, seed.b)
	}
	s := fuzzymatch.DefaultScorer()
	threshold := s.Threshold()
	f.Fuzz(func(t *testing.T, a, b string) {
		score := s.Score(a, b)
		match := s.Match(a, b)
		want := score >= threshold
		if match != want {
			t.Fatalf("Scorer.Match(%q, %q) = %v; Score = %g, Threshold = %g; want Match == (Score >= Threshold) == %v",
				a, b, match, score, threshold, want)
		}
	})
}
