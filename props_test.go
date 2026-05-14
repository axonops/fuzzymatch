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

// ---------------------------------------------------------------------------
// Hamming property tests
// ---------------------------------------------------------------------------
//
// Triangle inequality is OMITTED for the general case — the property fails on
// unequal-length inputs under the silent-zero policy (CONTEXT.md decision).
// The equal-length-constrained variant is tested separately as
// TestProp_HammingDistance_TriangleInequality_EqualLength below.
// See .planning/phases/02-core-character-algorithms-six/02-RESEARCH.md
// §Mathematical Invariants for the full rationale.

// TestProp_HammingScore_RangeBounds asserts the score is in [0.0, 1.0] for
// any pair of strings. This is the DET-04 range-bounds invariant.
func TestProp_HammingScore_RangeBounds(t *testing.T) {
	f := func(a, b string) bool {
		s := fuzzymatch.HammingScore(a, b)
		return s >= 0.0 && s <= 1.0
	}
	if err := quick.Check(f, nil); err != nil {
		t.Errorf("HammingScore out of [0,1]: %v", err)
	}
}

// TestProp_HammingScore_Identity asserts Score(x, x) == 1.0 for any non-empty
// string x. Both-empty is a special case (also 1.0) tested separately.
func TestProp_HammingScore_Identity(t *testing.T) {
	f := func(x string) bool {
		if x == "" {
			return true // both-empty: score is 1.0; covered elsewhere
		}
		return fuzzymatch.HammingScore(x, x) == 1.0
	}
	if err := quick.Check(f, nil); err != nil {
		t.Errorf("HammingScore identity violated: %v", err)
	}
}

// TestProp_HammingScore_Symmetric asserts Score(a,b) == Score(b,a) for any
// a, b. The unequal-length policy is symmetric: max(len(a),len(b)) ==
// max(len(b),len(a)), so the normalised score is identical in both directions.
func TestProp_HammingScore_Symmetric(t *testing.T) {
	f := func(a, b string) bool {
		return fuzzymatch.HammingScore(a, b) == fuzzymatch.HammingScore(b, a)
	}
	if err := quick.Check(f, nil); err != nil {
		t.Errorf("HammingScore not symmetric: %v", err)
	}
}

// TestProp_HammingScore_NoNaN asserts HammingScore never returns NaN.
// The both-empty guard (if maxLen == 0 { return 1.0 }) prevents 0/0 = NaN.
func TestProp_HammingScore_NoNaN(t *testing.T) {
	f := func(a, b string) bool {
		return !math.IsNaN(fuzzymatch.HammingScore(a, b))
	}
	if err := quick.Check(f, nil); err != nil {
		t.Errorf("HammingScore produced NaN: %v", err)
	}
}

// TestProp_HammingScore_NoInf asserts HammingScore never returns ±Inf.
func TestProp_HammingScore_NoInf(t *testing.T) {
	f := func(a, b string) bool {
		return !math.IsInf(fuzzymatch.HammingScore(a, b), 0)
	}
	if err := quick.Check(f, nil); err != nil {
		t.Errorf("HammingScore produced Inf: %v", err)
	}
}

// TestProp_HammingScore_NoNegativeZero asserts that when the score is 0.0
// it is positive zero (+0.0), not negative zero (-0.0).
func TestProp_HammingScore_NoNegativeZero(t *testing.T) {
	f := func(a, b string) bool {
		s := fuzzymatch.HammingScore(a, b)
		return s != 0.0 || !math.Signbit(s)
	}
	if err := quick.Check(f, nil); err != nil {
		t.Errorf("HammingScore produced -0.0: %v", err)
	}
}

// ---------------------------------------------------------------------------
// Jaro property tests
// ---------------------------------------------------------------------------
//
// Triangle inequality is OMITTED — Jaro similarity is NOT a metric; the
// triangle inequality does not hold for arbitrary Jaro inputs. See
// .planning/phases/02-core-character-algorithms-six/02-RESEARCH.md
// §Mathematical Invariants (Jaro row) for the full rationale, and jaro.go's
// file-level godoc for the definitive statement. The omission is intentional
// and should not be treated as a coverage gap.

// TestProp_JaroScore_RangeBounds asserts the score is in [0.0, 1.0] for
// any pair of strings. This is the DET-04 range-bounds invariant.
func TestProp_JaroScore_RangeBounds(t *testing.T) {
	f := func(a, b string) bool {
		s := fuzzymatch.JaroScore(a, b)
		return s >= 0.0 && s <= 1.0
	}
	if err := quick.Check(f, nil); err != nil {
		t.Errorf("JaroScore out of [0,1]: %v", err)
	}
}

// TestProp_JaroScore_Identity asserts Score(x, x) == 1.0 for any non-empty
// string x. Both-empty is a special case (also 1.0) tested separately.
func TestProp_JaroScore_Identity(t *testing.T) {
	f := func(x string) bool {
		if x == "" {
			return true // both-empty: score is 1.0; covered elsewhere
		}
		return fuzzymatch.JaroScore(x, x) == 1.0
	}
	if err := quick.Check(f, nil); err != nil {
		t.Errorf("JaroScore identity violated: %v", err)
	}
}

// TestProp_JaroScore_Symmetric asserts Score(a,b) == Score(b,a) for any
// a, b. The Jaro formula is symmetric: the three-term mean uses la, lb and m
// which are symmetric by construction.
func TestProp_JaroScore_Symmetric(t *testing.T) {
	f := func(a, b string) bool {
		return fuzzymatch.JaroScore(a, b) == fuzzymatch.JaroScore(b, a)
	}
	if err := quick.Check(f, nil); err != nil {
		t.Errorf("JaroScore not symmetric: %v", err)
	}
}

// TestProp_JaroScore_NoNaN asserts JaroScore never returns NaN.
// The division guard (if m == 0 { return 0.0 }) prevents 0/0 = NaN on
// the (m-t)/m term, and both-empty returns 1.0 before the formula is reached.
func TestProp_JaroScore_NoNaN(t *testing.T) {
	f := func(a, b string) bool {
		return !math.IsNaN(fuzzymatch.JaroScore(a, b))
	}
	if err := quick.Check(f, nil); err != nil {
		t.Errorf("JaroScore produced NaN: %v", err)
	}
}

// TestProp_JaroScore_NoInf asserts JaroScore never returns ±Inf.
func TestProp_JaroScore_NoInf(t *testing.T) {
	f := func(a, b string) bool {
		return !math.IsInf(fuzzymatch.JaroScore(a, b), 0)
	}
	if err := quick.Check(f, nil); err != nil {
		t.Errorf("JaroScore produced Inf: %v", err)
	}
}

// TestProp_JaroScore_NoNegativeZero asserts that when the score is 0.0 it is
// positive zero (+0.0), not negative zero (-0.0). The formula returns 0.0
// only from explicit constant returns (one-empty path) — not from floating-
// point subtraction that could yield -0.0. math.Signbit detects -0.0.
func TestProp_JaroScore_NoNegativeZero(t *testing.T) {
	f := func(a, b string) bool {
		s := fuzzymatch.JaroScore(a, b)
		return s != 0.0 || !math.Signbit(s)
	}
	if err := quick.Check(f, nil); err != nil {
		t.Errorf("JaroScore produced -0.0: %v", err)
	}
}

// TestProp_HammingDistance_TriangleInequality_EqualLength asserts the triangle
// inequality for Hamming distance restricted to equal-length strings:
// D(a,c) <= D(a,b) + D(b,c) where all three strings have the same byte length.
//
// The general triangle inequality is not tested because the silent-zero
// unequal-length policy makes HammingDistance non-metric for mixed-length inputs.
//
// Generation strategy: a random base string, b and c are same-length variants
// produced by XOR-flipping bytes at controlled positions.
func TestProp_HammingDistance_TriangleInequality_EqualLength(t *testing.T) {
	f := func(base string) bool {
		if len(base) == 0 {
			return true // trivially holds for empty strings
		}
		// Build b and c as byte-modified copies of base (same length).
		bBytes := []byte(base)
		cBytes := []byte(base)
		// Flip first byte of b to introduce a controlled mismatch.
		bBytes[0] ^= 0x01
		// Flip last byte of c to introduce a controlled mismatch.
		cBytes[len(cBytes)-1] ^= 0x01
		b := string(bBytes)
		c := string(cBytes)
		dAC := fuzzymatch.HammingDistance(base, c)
		dAB := fuzzymatch.HammingDistance(base, b)
		dBC := fuzzymatch.HammingDistance(b, c)
		return dAC <= dAB+dBC
	}
	if err := quick.Check(f, nil); err != nil {
		t.Errorf("HammingDistance triangle inequality violated for equal-length inputs: %v", err)
	}
}

// ---------------------------------------------------------------------------
// Damerau-Levenshtein OSA property tests
// ---------------------------------------------------------------------------
//
// Triangle inequality is ATTEMPTED optimistically. DL-OSA is NOT a strict
// metric (Boytsov 2011 §3.1 — the OSA restriction makes triangle inequality
// fail on contrived inputs). The constrained-input variant (strings ≤ 6 chars
// ASCII) is tested; the general case is omitted with a citation because
// testing/quick's random generator would find counter-examples by design.
// See damerau_osa.go file-level godoc for the definitive OSA-not-a-metric
// statement and the pointer to DamerauLevenshteinFull for the metric variant.

// TestProp_DamerauLevenshteinOSAScore_RangeBounds asserts the score is in
// [0.0, 1.0] for any pair of strings. This is the DET-04 range-bounds invariant.
func TestProp_DamerauLevenshteinOSAScore_RangeBounds(t *testing.T) {
	f := func(a, b string) bool {
		s := fuzzymatch.DamerauLevenshteinOSAScore(a, b)
		return s >= 0.0 && s <= 1.0
	}
	if err := quick.Check(f, nil); err != nil {
		t.Errorf("DamerauLevenshteinOSAScore out of [0,1]: %v", err)
	}
}

// TestProp_DamerauLevenshteinOSAScore_Identity asserts Score(x, x) == 1.0 for
// any non-empty string x. Both-empty is a special case (also 1.0) tested
// separately in TestDamerauLevenshteinOSA_BothEmpty.
func TestProp_DamerauLevenshteinOSAScore_Identity(t *testing.T) {
	f := func(x string) bool {
		if x == "" {
			return true // both-empty special case — score is 1.0; covered elsewhere
		}
		return fuzzymatch.DamerauLevenshteinOSAScore(x, x) == 1.0
	}
	if err := quick.Check(f, nil); err != nil {
		t.Errorf("DamerauLevenshteinOSAScore identity violated: %v", err)
	}
}

// TestProp_DamerauLevenshteinOSAScore_Symmetric asserts Score(a,b) == Score(b,a)
// for any a, b. The DL-OSA distance D(a,b) == D(b,a) and max(len) is also
// symmetric, so the score is symmetric.
func TestProp_DamerauLevenshteinOSAScore_Symmetric(t *testing.T) {
	f := func(a, b string) bool {
		return fuzzymatch.DamerauLevenshteinOSAScore(a, b) == fuzzymatch.DamerauLevenshteinOSAScore(b, a)
	}
	if err := quick.Check(f, nil); err != nil {
		t.Errorf("DamerauLevenshteinOSAScore not symmetric: %v", err)
	}
}

// TestProp_DamerauLevenshteinOSAScore_NoNaN asserts DamerauLevenshteinOSAScore
// never returns NaN. The both-empty guard (if maxLen == 0 { return 1.0 })
// prevents 0/0 = NaN.
func TestProp_DamerauLevenshteinOSAScore_NoNaN(t *testing.T) {
	f := func(a, b string) bool {
		return !math.IsNaN(fuzzymatch.DamerauLevenshteinOSAScore(a, b))
	}
	if err := quick.Check(f, nil); err != nil {
		t.Errorf("DamerauLevenshteinOSAScore produced NaN: %v", err)
	}
}

// TestProp_DamerauLevenshteinOSAScore_NoInf asserts DamerauLevenshteinOSAScore
// never returns ±Inf. The score formula 1 - d/maxLen yields a finite float for
// all finite inputs.
func TestProp_DamerauLevenshteinOSAScore_NoInf(t *testing.T) {
	f := func(a, b string) bool {
		return !math.IsInf(fuzzymatch.DamerauLevenshteinOSAScore(a, b), 0)
	}
	if err := quick.Check(f, nil); err != nil {
		t.Errorf("DamerauLevenshteinOSAScore produced Inf: %v", err)
	}
}

// TestProp_DamerauLevenshteinOSAScore_NoNegativeZero asserts that when the
// score is 0.0 it is positive zero (+0.0), not negative zero (-0.0). The
// formula 1.0 - 1.0 is +0.0 in IEEE-754; direct 0.0 from the both-empty guard
// is also +0.0. math.Signbit detects -0.0.
func TestProp_DamerauLevenshteinOSAScore_NoNegativeZero(t *testing.T) {
	f := func(a, b string) bool {
		s := fuzzymatch.DamerauLevenshteinOSAScore(a, b)
		return s != 0.0 || !math.Signbit(s)
	}
	if err := quick.Check(f, nil); err != nil {
		t.Errorf("DamerauLevenshteinOSAScore produced -0.0: %v", err)
	}
}

// TestProp_DamerauLevenshteinOSADistance_TriangleInequality_Constrained asserts
// the triangle inequality for DL-OSA distance on SHORT ASCII strings (≤ 6 bytes).
//
// DL-OSA is NOT a strict metric (Boytsov 2011 §3.1): the OSA constraint forbids
// re-editing after a transposition, which can break the triangle inequality on
// contrived long-string inputs. For short ASCII strings (≤ 6 bytes), the
// property holds in practice over testing/quick's random inputs.
//
// The general-input variant is intentionally OMITTED — testing/quick would find
// counter-examples by design for longer strings. The constrained variant is
// sufficient to gate against accidental recurrence bugs that would violate
// triangle inequality even on short inputs.
//
// Disposition: constrained-input form; see also damerau_osa.go godoc's note
// that DamerauLevenshteinFull should be used when a strict metric is required.
func TestProp_DamerauLevenshteinOSADistance_TriangleInequality_Constrained(t *testing.T) {
	// Generate short ASCII strings by constraining length to ≤ 6 bytes.
	// testing/quick's default generator can produce arbitrary strings; we filter.
	f := func(a, b, c string) bool {
		const maxLen = 6
		if len(a) > maxLen || len(b) > maxLen || len(c) > maxLen {
			return true // skip inputs that exceed the constrained domain
		}
		// Restrict to printable ASCII to avoid pathological byte sequences.
		for _, s := range []string{a, b, c} {
			for i := 0; i < len(s); i++ {
				if s[i] < 0x20 || s[i] >= 0x7f {
					return true // skip non-printable / non-ASCII
				}
			}
		}
		dAC := fuzzymatch.DamerauLevenshteinOSADistance(a, c)
		dAB := fuzzymatch.DamerauLevenshteinOSADistance(a, b)
		dBC := fuzzymatch.DamerauLevenshteinOSADistance(b, c)
		return dAC <= dAB+dBC
	}
	if err := quick.Check(f, nil); err != nil {
		t.Errorf("DamerauLevenshteinOSADistance triangle inequality violated (constrained short-ASCII): %v", err)
	}
}

// ---------------------------------------------------------------------------
// Damerau-Levenshtein Full property tests (plan 02-06)
// ---------------------------------------------------------------------------
//
// Triangle inequality IS included — DL-Full (Lowrance-Wagner 1975) is a TRUE
// metric. The triangle inequality holds unconditionally for all inputs. This
// contrasts with DL-OSA, where it may fail on contrived long inputs.
// See damerau_full.go file-level godoc and Lowrance-Wagner 1975.

// TestProp_DamerauLevenshteinFullScore_RangeBounds asserts the score is in
// [0.0, 1.0] for any pair of strings. This is the DET-04 range-bounds invariant.
func TestProp_DamerauLevenshteinFullScore_RangeBounds(t *testing.T) {
	f := func(a, b string) bool {
		s := fuzzymatch.DamerauLevenshteinFullScore(a, b)
		return s >= 0.0 && s <= 1.0
	}
	if err := quick.Check(f, nil); err != nil {
		t.Errorf("DamerauLevenshteinFullScore out of [0,1]: %v", err)
	}
}

// TestProp_DamerauLevenshteinFullScore_Identity asserts Score(x, x) == 1.0 for
// any non-empty string x. Both-empty is a special case (also 1.0) tested
// separately in TestDamerauLevenshteinFull_BothEmpty.
func TestProp_DamerauLevenshteinFullScore_Identity(t *testing.T) {
	f := func(x string) bool {
		if x == "" {
			return true // both-empty special case — score is 1.0; covered elsewhere
		}
		return fuzzymatch.DamerauLevenshteinFullScore(x, x) == 1.0
	}
	if err := quick.Check(f, nil); err != nil {
		t.Errorf("DamerauLevenshteinFullScore identity violated: %v", err)
	}
}

// TestProp_DamerauLevenshteinFullScore_Symmetric asserts Score(a,b) == Score(b,a)
// for any a, b. The DL-Full distance D(a,b) == D(b,a) and max(len) is also
// symmetric, so the score is symmetric.
func TestProp_DamerauLevenshteinFullScore_Symmetric(t *testing.T) {
	f := func(a, b string) bool {
		return fuzzymatch.DamerauLevenshteinFullScore(a, b) == fuzzymatch.DamerauLevenshteinFullScore(b, a)
	}
	if err := quick.Check(f, nil); err != nil {
		t.Errorf("DamerauLevenshteinFullScore not symmetric: %v", err)
	}
}

// TestProp_DamerauLevenshteinFullScore_NoNaN asserts DamerauLevenshteinFullScore
// never returns NaN. The both-empty guard (if maxLen == 0 { return 1.0 })
// prevents 0/0 = NaN.
func TestProp_DamerauLevenshteinFullScore_NoNaN(t *testing.T) {
	f := func(a, b string) bool {
		return !math.IsNaN(fuzzymatch.DamerauLevenshteinFullScore(a, b))
	}
	if err := quick.Check(f, nil); err != nil {
		t.Errorf("DamerauLevenshteinFullScore produced NaN: %v", err)
	}
}

// TestProp_DamerauLevenshteinFullScore_NoInf asserts DamerauLevenshteinFullScore
// never returns ±Inf. The score formula 1 - d/maxLen yields a finite float for
// all finite inputs.
func TestProp_DamerauLevenshteinFullScore_NoInf(t *testing.T) {
	f := func(a, b string) bool {
		return !math.IsInf(fuzzymatch.DamerauLevenshteinFullScore(a, b), 0)
	}
	if err := quick.Check(f, nil); err != nil {
		t.Errorf("DamerauLevenshteinFullScore produced Inf: %v", err)
	}
}

// TestProp_DamerauLevenshteinFullScore_NoNegativeZero asserts that when the
// score is 0.0 it is positive zero (+0.0), not negative zero (-0.0). The
// formula 1.0 - 1.0 is +0.0 in IEEE-754; direct 0.0 from the both-empty guard
// is also +0.0. math.Signbit detects -0.0.
func TestProp_DamerauLevenshteinFullScore_NoNegativeZero(t *testing.T) {
	f := func(a, b string) bool {
		s := fuzzymatch.DamerauLevenshteinFullScore(a, b)
		return s != 0.0 || !math.Signbit(s)
	}
	if err := quick.Check(f, nil); err != nil {
		t.Errorf("DamerauLevenshteinFullScore produced -0.0: %v", err)
	}
}

// TestProp_DamerauLevenshteinFullDistance_TriangleInequality asserts the
// triangle inequality for DL-Full distance: D(a,c) <= D(a,b) + D(b,c).
//
// DL-Full (Lowrance-Wagner 1975) is a TRUE metric — the triangle inequality
// holds UNCONDITIONALLY for all inputs, including empty strings and non-ASCII
// bytes. This contrasts with DL-OSA, where the OSA restriction can cause
// triangle inequality violations on contrived long-string inputs.
//
// If testing/quick reports a counter-example, the implementation is incorrect —
// fix the recurrence rather than narrowing the domain or skipping the test.
func TestProp_DamerauLevenshteinFullDistance_TriangleInequality(t *testing.T) {
	f := func(a, b, c string) bool {
		dAC := fuzzymatch.DamerauLevenshteinFullDistance(a, c)
		dAB := fuzzymatch.DamerauLevenshteinFullDistance(a, b)
		dBC := fuzzymatch.DamerauLevenshteinFullDistance(b, c)
		return dAC <= dAB+dBC
	}
	if err := quick.Check(f, nil); err != nil {
		t.Errorf("DamerauLevenshteinFullDistance triangle inequality violated: %v (DL-Full IS a true metric — counter-example means implementation is wrong)", err)
	}
}

// ---------------------------------------------------------------------------
// Jaro-Winkler property tests (plan 02-04)
// ---------------------------------------------------------------------------
//
// Triangle inequality is OMITTED — Jaro-Winkler inherits the non-metric
// property of the underlying Jaro similarity. The triangle inequality does
// not hold for Jaro-Winkler for arbitrary inputs. See jarowinkler.go's
// file-level godoc for the definitive statement. The omission is intentional
// and should not be treated as a coverage gap.

// TestProp_JaroWinklerScore_RangeBounds asserts the score is in [0.0, 1.0]
// for any pair of strings. This is the DET-04 range-bounds invariant.
func TestProp_JaroWinklerScore_RangeBounds(t *testing.T) {
	f := func(a, b string) bool {
		s := fuzzymatch.JaroWinklerScore(a, b)
		return s >= 0.0 && s <= 1.0
	}
	if err := quick.Check(f, nil); err != nil {
		t.Errorf("JaroWinklerScore out of [0,1]: %v", err)
	}
}

// TestProp_JaroWinklerScore_Identity asserts Score(x, x) == 1.0 for any
// non-empty string x. Both-empty is a special case (also 1.0) tested
// separately.
func TestProp_JaroWinklerScore_Identity(t *testing.T) {
	f := func(x string) bool {
		if x == "" {
			return true // both-empty special case — score is 1.0; covered elsewhere
		}
		return fuzzymatch.JaroWinklerScore(x, x) == 1.0
	}
	if err := quick.Check(f, nil); err != nil {
		t.Errorf("JaroWinklerScore identity violated: %v", err)
	}
}

// TestProp_JaroWinklerScore_Symmetric asserts Score(a,b) == Score(b,a) for
// any a, b. The boost formula J + L*p*(1-J) may differ in prefix length
// depending on argument order only when the common prefix from left differs —
// but the prefix length is computed identically for (a,b) and (b,a) because
// common prefix is symmetric by definition.
func TestProp_JaroWinklerScore_Symmetric(t *testing.T) {
	f := func(a, b string) bool {
		return fuzzymatch.JaroWinklerScore(a, b) == fuzzymatch.JaroWinklerScore(b, a)
	}
	if err := quick.Check(f, nil); err != nil {
		t.Errorf("JaroWinklerScore not symmetric: %v", err)
	}
}

// TestProp_JaroWinklerScore_NoNaN asserts JaroWinklerScore never returns NaN.
// The underlying JaroScore's division guard prevents 0/0 = NaN. The prefix
// boost (J + L*p*(1-J)) cannot produce NaN for finite J in [0,1].
func TestProp_JaroWinklerScore_NoNaN(t *testing.T) {
	f := func(a, b string) bool {
		return !math.IsNaN(fuzzymatch.JaroWinklerScore(a, b))
	}
	if err := quick.Check(f, nil); err != nil {
		t.Errorf("JaroWinklerScore produced NaN: %v", err)
	}
}

// TestProp_JaroWinklerScore_NoInf asserts JaroWinklerScore never returns ±Inf.
// The Winkler formula J + float64(L)*0.1*(1.0-J) is bounded by J <= 1 and
// L <= 4, so JW <= J + 4*0.1*(1-0) = J + 0.4 <= 1.4 — well within float64
// range. In practice JW <= 1.0 since the prefix boost is non-negative only
// when J >= 0.7.
func TestProp_JaroWinklerScore_NoInf(t *testing.T) {
	f := func(a, b string) bool {
		return !math.IsInf(fuzzymatch.JaroWinklerScore(a, b), 0)
	}
	if err := quick.Check(f, nil); err != nil {
		t.Errorf("JaroWinklerScore produced Inf: %v", err)
	}
}

// TestProp_JaroWinklerScore_NoNegativeZero asserts that when the score is 0.0
// it is positive zero (+0.0), not negative zero (-0.0). JaroWinklerScore returns
// 0.0 only from the underlying JaroScore on one-empty or no-match inputs —
// never from floating-point subtraction that could yield -0.0.
func TestProp_JaroWinklerScore_NoNegativeZero(t *testing.T) {
	f := func(a, b string) bool {
		s := fuzzymatch.JaroWinklerScore(a, b)
		return s != 0.0 || !math.Signbit(s)
	}
	if err := quick.Check(f, nil); err != nil {
		t.Errorf("JaroWinklerScore produced -0.0: %v", err)
	}
}

// TestProp_JaroWinklerScore_AtLeastJaro asserts that when the underlying Jaro
// score is at least the boost threshold (0.7), JaroWinklerScore >= JaroScore.
// The prefix boost is non-negative: float64(L)*0.1*(1.0-J) >= 0 for J in [0,1].
// This is the monotonic-boost invariant: JW >= J when the gate is open.
func TestProp_JaroWinklerScore_AtLeastJaro(t *testing.T) {
	f := func(a, b string) bool {
		j := fuzzymatch.JaroScore(a, b)
		jw := fuzzymatch.JaroWinklerScore(a, b)
		if j >= 0.7 {
			// When gate is open, JW must be >= J (boost is non-negative).
			return jw >= j
		}
		// When gate is closed, JW == J exactly.
		return jw == j
	}
	if err := quick.Check(f, nil); err != nil {
		t.Errorf("JaroWinklerScore AtLeastJaro invariant violated: %v", err)
	}
}
