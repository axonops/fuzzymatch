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

// props_test.go contains testing/quick property tests for Phase 2's six
// character-based algorithms. Each algorithm is covered by:
//
//   - TestProp_XxxScore_RangeBounds   ([0,1] for any input)
//   - TestProp_XxxScore_Identity      (Score(x,x) == 1.0 for non-empty x)
//   - TestProp_XxxScore_Symmetric     (Score(a,b) == Score(b,a))
//   - TestProp_XxxDistance_TriangleInequality (for DP algorithms that form a
//     metric; NOT for Jaro/JW which are not metrics)
//   - TestProp_XxxScore_NoNaN
//   - TestProp_XxxScore_NoInf
//   - TestProp_XxxScore_NoNegativeZero
//
// Wave 2 plans (02-02 through 02-06) append their algorithm's property tests
// to this SAME file. The file-level godoc comment is owned by Wave 1 (this
// plan).
//
// Stdlib `testing` and `testing/quick` only — no testify in root tests, per
// .claude/skills/go-coding-standards.

package fuzzymatch_test

import (
	"math"
	"testing"
	"testing/quick"

	"github.com/axonops/fuzzymatch"
)

// ---------------------------------------------------------------------------
// Levenshtein property tests
// ---------------------------------------------------------------------------

// TestProp_LevenshteinScore_RangeBounds asserts the score is in [0.0, 1.0] for
// any pair of strings. This is the DET-04 range-bounds invariant.
func TestProp_LevenshteinScore_RangeBounds(t *testing.T) {
	f := func(a, b string) bool {
		s := fuzzymatch.LevenshteinScore(a, b)
		return s >= 0.0 && s <= 1.0
	}
	if err := quick.Check(f, nil); err != nil {
		t.Errorf("LevenshteinScore out of [0,1]: %v", err)
	}
}

// TestProp_LevenshteinScore_Identity asserts Score(x, x) == 1.0 for any non-
// empty string x. Both-empty is a special case (also 1.0) tested separately.
func TestProp_LevenshteinScore_Identity(t *testing.T) {
	f := func(x string) bool {
		if x == "" {
			return true // both-empty special case — score is 1.0; covered elsewhere
		}
		return fuzzymatch.LevenshteinScore(x, x) == 1.0
	}
	if err := quick.Check(f, nil); err != nil {
		t.Errorf("LevenshteinScore identity violated: %v", err)
	}
}

// TestProp_LevenshteinScore_Symmetric asserts Score(a,b) == Score(b,a) for any
// a, b. This is the symmetry invariant — the edit distance D(a,b) == D(b,a)
// and max(len) is also symmetric, so the score is symmetric.
func TestProp_LevenshteinScore_Symmetric(t *testing.T) {
	f := func(a, b string) bool {
		return fuzzymatch.LevenshteinScore(a, b) == fuzzymatch.LevenshteinScore(b, a)
	}
	if err := quick.Check(f, nil); err != nil {
		t.Errorf("LevenshteinScore not symmetric: %v", err)
	}
}

// TestProp_LevenshteinDistance_TriangleInequality asserts the triangle
// inequality for the Levenshtein edit distance: D(a,c) <= D(a,b) + D(b,c).
// Levenshtein distance is a metric; the triangle inequality holds for all
// inputs, including empty strings and non-ASCII bytes.
func TestProp_LevenshteinDistance_TriangleInequality(t *testing.T) {
	f := func(a, b, c string) bool {
		dAC := fuzzymatch.LevenshteinDistance(a, c)
		dAB := fuzzymatch.LevenshteinDistance(a, b)
		dBC := fuzzymatch.LevenshteinDistance(b, c)
		return dAC <= dAB+dBC
	}
	if err := quick.Check(f, nil); err != nil {
		t.Errorf("LevenshteinDistance triangle inequality violated: %v", err)
	}
}

// TestProp_LevenshteinScore_NoNaN asserts LevenshteinScore never returns NaN.
// The both-empty guard (if maxLen == 0 { return 1.0 }) prevents 0/0 = NaN.
func TestProp_LevenshteinScore_NoNaN(t *testing.T) {
	f := func(a, b string) bool {
		return !math.IsNaN(fuzzymatch.LevenshteinScore(a, b))
	}
	if err := quick.Check(f, nil); err != nil {
		t.Errorf("LevenshteinScore produced NaN: %v", err)
	}
}

// TestProp_LevenshteinScore_NoInf asserts LevenshteinScore never returns ±Inf.
// The score formula 1 - d/maxLen yields a finite float for all finite inputs.
func TestProp_LevenshteinScore_NoInf(t *testing.T) {
	f := func(a, b string) bool {
		return !math.IsInf(fuzzymatch.LevenshteinScore(a, b), 0)
	}
	if err := quick.Check(f, nil); err != nil {
		t.Errorf("LevenshteinScore produced Inf: %v", err)
	}
}

// TestProp_LevenshteinScore_NoNegativeZero asserts that when the score is 0.0
// it is positive zero (+0.0), not negative zero (-0.0). The formula
// 1.0 - 1.0 is +0.0 in IEEE-754; direct 0.0 from the both-empty guard is also
// +0.0. math.Signbit detects -0.0.
func TestProp_LevenshteinScore_NoNegativeZero(t *testing.T) {
	f := func(a, b string) bool {
		s := fuzzymatch.LevenshteinScore(a, b)
		return s != 0.0 || !math.Signbit(s)
	}
	if err := quick.Check(f, nil); err != nil {
		t.Errorf("LevenshteinScore produced -0.0: %v", err)
	}
}
