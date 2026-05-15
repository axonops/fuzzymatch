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
	"math/rand"
	"reflect"
	"strings"
	"testing"
	"testing/quick"

	"github.com/axonops/fuzzymatch"
)

// randShortASCII returns a random printable-ASCII string of length in [0, maxLen].
// Used by property tests that require in-domain inputs (e.g. WR-04: OSA triangle
// inequality holds only for short ASCII triples). Using a custom generator via
// quick.Config.Values guarantees every drawn triple actually exercises the
// property, replacing the prior return-true filter idiom that could silently
// pass when all generated inputs were out-of-domain.
func randShortASCII(r *rand.Rand, maxLen int) string {
	n := r.Intn(maxLen + 1) // 0..maxLen inclusive
	if n == 0 {
		return ""
	}
	const printableLow, printableHigh = byte(0x20), byte(0x7e)
	b := make([]byte, n)
	for i := range b {
		b[i] = printableLow + byte(r.Intn(int(printableHigh-printableLow)+1))
	}
	return string(b)
}

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
	// WR-04 companion fix: track that a meaningful number of non-empty bases
	// were actually exercised. quick.Check defaults to 100 iterations; the
	// previous "return true on empty" was correct but invisible to readers
	// inspecting coverage.
	exercised := 0
	f := func(base string) bool {
		if len(base) == 0 {
			return true // trivially holds for empty strings
		}
		exercised++
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
	// At least one non-empty base must be drawn for the property to have been
	// exercised. testing/quick's default generator overwhelmingly produces
	// non-empty strings; this guard makes the coverage gate explicit.
	if exercised == 0 {
		t.Fatalf("HammingDistance triangle inequality: 0 non-empty bases drawn — generator misconfigured")
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
	// WR-04: Use a custom quick.Config.Values generator that always emits
	// in-domain inputs (printable ASCII, length ≤ 6). The previous return-true
	// filter could silently pass when every generated triple was out-of-domain.
	// A counter records how many triples were actually exercised; the test
	// fails if zero were drawn (defence against a misconfigured generator).
	const maxLen = 6
	exercised := 0
	f := func(a, b, c string) bool {
		exercised++
		dAC := fuzzymatch.DamerauLevenshteinOSADistance(a, c)
		dAB := fuzzymatch.DamerauLevenshteinOSADistance(a, b)
		dBC := fuzzymatch.DamerauLevenshteinOSADistance(b, c)
		return dAC <= dAB+dBC
	}
	cfg := &quick.Config{
		Values: func(args []reflect.Value, r *rand.Rand) {
			args[0] = reflect.ValueOf(randShortASCII(r, maxLen))
			args[1] = reflect.ValueOf(randShortASCII(r, maxLen))
			args[2] = reflect.ValueOf(randShortASCII(r, maxLen))
		},
	}
	if err := quick.Check(f, cfg); err != nil {
		t.Errorf("DamerauLevenshteinOSADistance triangle inequality violated (constrained short-ASCII): %v", err)
	}
	// Guard against silent-pass: quick.Check defaults to 100 iterations. If
	// zero triples were actually exercised, the test would have passed without
	// running the property — which is the WR-04 defect being closed.
	if exercised == 0 {
		t.Fatalf("DamerauLevenshteinOSADistance triangle inequality: 0 triples exercised — generator misconfigured")
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

// ---------------------------------------------------------------------------
// Smith-Waterman-Gotoh property tests (plan 03-01)
// ---------------------------------------------------------------------------
//
// Standard Phase 2 invariants (range bounds, identity, symmetric, NoNaN,
// NoInf, NoNegativeZero) plus three SWG-specific canaries per PITFALLS.md
// §3 (GapSplitInvariance, RawNeverExceedsMatchTimesMinLen,
// MonotonicWithMatchReward). Triangle inequality is OMITTED — SWG is not a
// metric over the full string space.

// TestProp_SmithWatermanGotohScore_RangeBounds asserts the score is in
// [0.0, 1.0] for any pair of strings. DET-04 range-bounds invariant.
func TestProp_SmithWatermanGotohScore_RangeBounds(t *testing.T) {
	f := func(a, b string) bool {
		s := fuzzymatch.SmithWatermanGotohScore(a, b)
		return s >= 0.0 && s <= 1.0
	}
	if err := quick.Check(f, nil); err != nil {
		t.Errorf("SmithWatermanGotohScore out of [0,1]: %v", err)
	}
}

// TestProp_SmithWatermanGotohScore_Identity asserts Score(x, x) == 1.0 for any
// non-empty string x. Both-empty is also 1.0 but covered by the unit tests.
func TestProp_SmithWatermanGotohScore_Identity(t *testing.T) {
	f := func(x string) bool {
		if x == "" {
			return true // both-empty special case — covered elsewhere
		}
		return fuzzymatch.SmithWatermanGotohScore(x, x) == 1.0
	}
	if err := quick.Check(f, nil); err != nil {
		t.Errorf("SmithWatermanGotohScore identity violated: %v", err)
	}
}

// TestProp_SmithWatermanGotohScore_Symmetric asserts Score(a, b) == Score(b, a)
// for any a, b. SWG's recurrence is symmetric in the two inputs.
func TestProp_SmithWatermanGotohScore_Symmetric(t *testing.T) {
	f := func(a, b string) bool {
		return fuzzymatch.SmithWatermanGotohScore(a, b) == fuzzymatch.SmithWatermanGotohScore(b, a)
	}
	if err := quick.Check(f, nil); err != nil {
		t.Errorf("SmithWatermanGotohScore not symmetric: %v", err)
	}
}

// TestProp_SmithWatermanGotohScore_NoNaN asserts the score is never NaN.
// The min-length zero guard (both-empty → 1.0, one-empty → 0.0) prevents
// 0/0 = NaN; the kernel uses only +/-/*/comparisons so no other NaN paths
// exist.
func TestProp_SmithWatermanGotohScore_NoNaN(t *testing.T) {
	f := func(a, b string) bool {
		return !math.IsNaN(fuzzymatch.SmithWatermanGotohScore(a, b))
	}
	if err := quick.Check(f, nil); err != nil {
		t.Errorf("SmithWatermanGotohScore produced NaN: %v", err)
	}
}

// TestProp_SmithWatermanGotohScore_NoInf asserts the score never returns
// ±Inf. The DP kernel sums finitely many finite quantities; the clamp/divide
// step is bounded by definition.
func TestProp_SmithWatermanGotohScore_NoInf(t *testing.T) {
	f := func(a, b string) bool {
		return !math.IsInf(fuzzymatch.SmithWatermanGotohScore(a, b), 0)
	}
	if err := quick.Check(f, nil); err != nil {
		t.Errorf("SmithWatermanGotohScore produced Inf: %v", err)
	}
}

// TestProp_SmithWatermanGotohScore_NoNegativeZero asserts that when the score
// is 0.0 it is positive zero (+0.0), not negative zero (-0.0). math.Signbit
// detects -0.0.
func TestProp_SmithWatermanGotohScore_NoNegativeZero(t *testing.T) {
	f := func(a, b string) bool {
		s := fuzzymatch.SmithWatermanGotohScore(a, b)
		return s != 0.0 || !math.Signbit(s)
	}
	if err := quick.Check(f, nil); err != nil {
		t.Errorf("SmithWatermanGotohScore produced -0.0: %v", err)
	}
}

// TestProp_SmithWatermanGotoh_GapSplitInvariance is the load-bearing
// Gotoh-erratum canary (PITFALLS.md §3 warning sign #2). Splitting a single
// long gap into a shorter gap (by removing some gap characters) must NOT
// produce a higher score than the original case — affine-gap penalties
// can only accrue, never improve, as gap length grows. Equivalently: shorter
// gaps score >= longer gaps for the same target string.
//
// Hand-curated triples (per CONTEXT.md §5 — NOT testing/quick); each row is
// (longGap, shortGap, target) where we assert Score(longGap, target) <=
// Score(shortGap, target).
func TestProp_SmithWatermanGotoh_GapSplitInvariance(t *testing.T) {
	triples := []struct {
		longGap, shortGap, target string
	}{
		{"abc________def", "abc____def", "abcdef"},
		{"abc____def", "abc__def", "abcdef"},
		{"aXXXXXXXXb", "aXXXXb", "ab"},
		{"aXXXXb", "aXXb", "ab"},
		{"xyz_____abc_____pqr", "xyz_abc_pqr", "xyzabcpqr"},
	}
	for _, tr := range triples {
		long := fuzzymatch.SmithWatermanGotohScore(tr.longGap, tr.target)
		short := fuzzymatch.SmithWatermanGotohScore(tr.shortGap, tr.target)
		if long > short {
			t.Errorf("GapSplitInvariance: Score(%q, %q)=%g > Score(%q, %q)=%g (longer gap improved score — PITFALLS §3 #2)",
				tr.longGap, tr.target, long, tr.shortGap, tr.target, short)
		}
	}
}

// TestProp_SmithWatermanGotoh_RawNeverExceedsMatchTimesMinLen asserts the
// upper bound: RawScore(a, b) <= Match * min(len(a), len(b)). The best local
// alignment has at most min(len) matching positions; with zero or more gap
// penalties, the raw score cannot exceed Match * min(len). Skips degenerate
// cases where min == 0.
func TestProp_SmithWatermanGotoh_RawNeverExceedsMatchTimesMinLen(t *testing.T) {
	params := fuzzymatch.NewSWGParams()
	f := func(a, b string) bool {
		la, lb := len(a), len(b)
		minLen := lb
		if la < lb {
			minLen = la
		}
		if minLen == 0 {
			return true // degenerate
		}
		raw := fuzzymatch.SmithWatermanGotohRawScore(a, b)
		upper := params.Match * float64(minLen)
		return raw <= upper
	}
	if err := quick.Check(f, nil); err != nil {
		t.Errorf("RawScore exceeded Match * min(len): %v", err)
	}
}

// TestProp_SmithWatermanGotoh_MonotonicWithMatchReward asserts that
// increasing only the Match parameter (keeping Mismatch, GapOpen, GapExtend
// fixed) cannot decrease the raw alignment score. Higher Match reward can
// only improve (or hold) the best local alignment.
//
// The property is checked across a randomised (baseMatch, delta) pair from
// testing/quick rather than the single +1.0 delta from the default — this
// protects the test from silent coverage shrinkage if NewSWGParams's
// defaults ever drift, and exercises the monotonicity across the full
// non-negative Match domain (not just the [1.0, 2.0] sliver).
func TestProp_SmithWatermanGotoh_MonotonicWithMatchReward(t *testing.T) {
	f := func(a, b string, baseMatch, delta float64) bool {
		if len(a) == 0 || len(b) == 0 {
			return true // degenerate: raw is 0 regardless
		}
		// Constrain randomised params to the documented Match domain: non-NaN,
		// non-Inf, non-negative. Out-of-domain inputs return true (vacuously
		// satisfied) so testing/quick's float64 generator does not flood the
		// signal with degenerate cases.
		if math.IsNaN(baseMatch) || math.IsInf(baseMatch, 0) || baseMatch < 0 {
			return true
		}
		if math.IsNaN(delta) || math.IsInf(delta, 0) || delta < 0 {
			return true
		}
		// Bound the magnitudes so high-Match doesn't drive RawScore past
		// float64 precision (testing/quick's default float64 range includes
		// values like 1e+300 which would overflow into Inf inside the DP).
		if baseMatch > 1e6 || delta > 1e6 {
			return true
		}
		baseParams := fuzzymatch.NewSWGParams()
		baseParams.Match = baseMatch
		highParams := baseParams
		highParams.Match = baseMatch + delta
		base := fuzzymatch.SmithWatermanGotohRawScoreWithParams(a, b, baseParams)
		high := fuzzymatch.SmithWatermanGotohRawScoreWithParams(a, b, highParams)
		return high >= base
	}
	if err := quick.Check(f, nil); err != nil {
		t.Errorf("MonotonicWithMatchReward violated: %v", err)
	}
}

// ---------------------------------------------------------------------------
// Rune-path symmetry property tests (WR-03)
// ---------------------------------------------------------------------------
//
// Each *Runes function is a separate code path from its byte-path sibling
// (e.g. jaroRunes vs jaroBytes, levenshteinDistanceRuneSlices vs levenshteinDP).
// The byte-path symmetry property tests above do not cover regressions in the
// rune kernels. These property tests close that gap by quick.Check-ing the
// symmetry invariant Score(a, b) == Score(b, a) on the rune variants.
//
// testing/quick's default string generator routinely produces multi-byte UTF-8
// (occasionally including invalid sequences), which is exactly the input class
// the rune paths exist to handle.

func TestProp_LevenshteinScoreRunes_Symmetric(t *testing.T) {
	f := func(a, b string) bool {
		return fuzzymatch.LevenshteinScoreRunes(a, b) == fuzzymatch.LevenshteinScoreRunes(b, a)
	}
	if err := quick.Check(f, nil); err != nil {
		t.Errorf("LevenshteinScoreRunes not symmetric: %v", err)
	}
}

func TestProp_HammingScoreRunes_Symmetric(t *testing.T) {
	f := func(a, b string) bool {
		return fuzzymatch.HammingScoreRunes(a, b) == fuzzymatch.HammingScoreRunes(b, a)
	}
	if err := quick.Check(f, nil); err != nil {
		t.Errorf("HammingScoreRunes not symmetric: %v", err)
	}
}

func TestProp_JaroScoreRunes_Symmetric(t *testing.T) {
	f := func(a, b string) bool {
		return fuzzymatch.JaroScoreRunes(a, b) == fuzzymatch.JaroScoreRunes(b, a)
	}
	if err := quick.Check(f, nil); err != nil {
		t.Errorf("JaroScoreRunes not symmetric: %v", err)
	}
}

func TestProp_JaroWinklerScoreRunes_Symmetric(t *testing.T) {
	f := func(a, b string) bool {
		return fuzzymatch.JaroWinklerScoreRunes(a, b) == fuzzymatch.JaroWinklerScoreRunes(b, a)
	}
	if err := quick.Check(f, nil); err != nil {
		t.Errorf("JaroWinklerScoreRunes not symmetric: %v", err)
	}
}

func TestProp_DamerauLevenshteinOSAScoreRunes_Symmetric(t *testing.T) {
	f := func(a, b string) bool {
		return fuzzymatch.DamerauLevenshteinOSAScoreRunes(a, b) == fuzzymatch.DamerauLevenshteinOSAScoreRunes(b, a)
	}
	if err := quick.Check(f, nil); err != nil {
		t.Errorf("DamerauLevenshteinOSAScoreRunes not symmetric: %v", err)
	}
}

func TestProp_DamerauLevenshteinFullScoreRunes_Symmetric(t *testing.T) {
	f := func(a, b string) bool {
		return fuzzymatch.DamerauLevenshteinFullScoreRunes(a, b) == fuzzymatch.DamerauLevenshteinFullScoreRunes(b, a)
	}
	if err := quick.Check(f, nil); err != nil {
		t.Errorf("DamerauLevenshteinFullScoreRunes not symmetric: %v", err)
	}
}

func TestProp_SmithWatermanGotohScoreRunes_Symmetric(t *testing.T) {
	f := func(a, b string) bool {
		return fuzzymatch.SmithWatermanGotohScoreRunes(a, b) == fuzzymatch.SmithWatermanGotohScoreRunes(b, a)
	}
	if err := quick.Check(f, nil); err != nil {
		t.Errorf("SmithWatermanGotohScoreRunes not symmetric: %v", err)
	}
}

// ---------------------------------------------------------------------------
// Strcmp95 property tests (plan 04-01)
// ---------------------------------------------------------------------------
//
// Standard Phase 2 invariants (range bounds, identity, symmetric, NoNaN, NoInf,
// NoNegativeZero) plus two algorithm-specific properties:
//
//   - AtLeastJaroWinkler — Strcmp95 = Jaro-Winkler + (similar-char credit) +
//     (long-string adjustment); both adjustments only ADD, so
//     Strcmp95Score(a, b) >= JaroWinklerScore(a, b) for every (a, b).
//   - DeterministicAcrossRuns — same input produces byte-identical output
//     across 1000 sequential calls (PITFALLS §14 closure: confirms the
//     `var`-level similar-character table is not mutated mid-process).
//
// There is no *Runes variant per CONTEXT.md §2 — no rune-path symmetry test.

// TestProp_Strcmp95Score_RangeBounds asserts the score is in [0.0, 1.0] for
// any pair of strings. DET-04 range-bounds invariant.
func TestProp_Strcmp95Score_RangeBounds(t *testing.T) {
	f := func(a, b string) bool {
		s := fuzzymatch.Strcmp95Score(a, b)
		return s >= 0.0 && s <= 1.0
	}
	if err := quick.Check(f, nil); err != nil {
		t.Errorf("Strcmp95Score out of [0,1]: %v", err)
	}
}

// TestProp_Strcmp95Score_Identity asserts Score(x, x) == 1.0 for any non-empty
// string x. Both-empty is also 1.0 but covered by the unit tests.
func TestProp_Strcmp95Score_Identity(t *testing.T) {
	f := func(x string) bool {
		if x == "" {
			return true // both-empty special case — covered elsewhere
		}
		return fuzzymatch.Strcmp95Score(x, x) == 1.0
	}
	if err := quick.Check(f, nil); err != nil {
		t.Errorf("Strcmp95Score identity violated: %v", err)
	}
}

// TestProp_Strcmp95Score_Symmetric asserts Score(a, b) == Score(b, a) for any
// a, b. Strcmp95's match-flag derivation, transposition pass, and
// similar-character pass are all symmetric in argument order — the recurrence
// is structurally symmetric.
func TestProp_Strcmp95Score_Symmetric(t *testing.T) {
	f := func(a, b string) bool {
		return fuzzymatch.Strcmp95Score(a, b) == fuzzymatch.Strcmp95Score(b, a)
	}
	if err := quick.Check(f, nil); err != nil {
		t.Errorf("Strcmp95Score not symmetric: %v", err)
	}
}

// TestProp_Strcmp95Score_NoNaN asserts the score is never NaN. The empty-
// input guard (both-empty → 1.0, one-empty → 0.0) and the m == 0 guard
// (returns 0.0 before any division) together prevent 0/0 = NaN; the
// algorithm uses only +/-/*/comparisons so no other NaN paths exist.
func TestProp_Strcmp95Score_NoNaN(t *testing.T) {
	f := func(a, b string) bool {
		return !math.IsNaN(fuzzymatch.Strcmp95Score(a, b))
	}
	if err := quick.Check(f, nil); err != nil {
		t.Errorf("Strcmp95Score produced NaN: %v", err)
	}
}

// TestProp_Strcmp95Score_NoInf asserts the score never returns ±Inf. The
// algorithm sums finitely many finite quantities; the clamp at function exit
// caps the final value at 1.0.
func TestProp_Strcmp95Score_NoInf(t *testing.T) {
	f := func(a, b string) bool {
		return !math.IsInf(fuzzymatch.Strcmp95Score(a, b), 0)
	}
	if err := quick.Check(f, nil); err != nil {
		t.Errorf("Strcmp95Score produced Inf: %v", err)
	}
}

// TestProp_Strcmp95Score_NoNegativeZero asserts that when the score is 0.0 it
// is positive zero (+0.0), not negative zero (-0.0). math.Signbit detects
// -0.0.
func TestProp_Strcmp95Score_NoNegativeZero(t *testing.T) {
	f := func(a, b string) bool {
		s := fuzzymatch.Strcmp95Score(a, b)
		return s != 0.0 || !math.Signbit(s)
	}
	if err := quick.Check(f, nil); err != nil {
		t.Errorf("Strcmp95Score produced -0.0: %v", err)
	}
}

// TestProp_Strcmp95Score_AtLeastJaroWinkler is the load-bearing hierarchy
// invariant: Strcmp95Score(a, b) >= JaroWinklerScore(a, b) for every (a, b).
// The Winkler 1994 adjustments (similar-character credit, long-string
// adjustment) only ADD to the base Jaro/Jaro-Winkler score; they never
// subtract. Violation indicates either the similar-character credit pass
// decreases the Jaro denominator/numerator in the wrong direction, OR the
// long-string adjustment formula has a sign error.
//
// RESEARCH.md Pitfall 1 warning sign #3 closure.
//
// Uses an absolute-precision epsilon of 1e-12 to absorb ULP-level float
// imprecision in the parallel kernel runs — the algorithm is float-stable
// per DET-06, but downstream floating-point reductions in the Jaro and
// Strcmp95 paths can differ by sub-ULP amounts without violating the
// hierarchy invariant.
func TestProp_Strcmp95Score_AtLeastJaroWinkler(t *testing.T) {
	f := func(a, b string) bool {
		s := fuzzymatch.Strcmp95Score(a, b)
		jw := fuzzymatch.JaroWinklerScore(a, b)
		return s+1e-12 >= jw
	}
	if err := quick.Check(f, nil); err != nil {
		t.Errorf("Strcmp95Score < JaroWinklerScore (hierarchy violated): %v", err)
	}
}

// TestProp_Strcmp95Score_DeterministicAcrossRuns asserts that the same input
// produces byte-identical output across 1000 sequential calls. This is the
// PITFALLS §14 closure: confirms the `var`-level similar-character table is
// not mutated mid-process AND that no init-order non-determinism affects the
// score.
//
// The pair MARTHA/MARHTA is chosen because it exercises the long-string
// adjustment AND the prefix boost without firing the similar-character pass
// — a representative non-trivial input.
func TestProp_Strcmp95Score_DeterministicAcrossRuns(t *testing.T) {
	const a, b = "MARTHA", "MARHTA"
	want := fuzzymatch.Strcmp95Score(a, b)
	for i := 0; i < 1000; i++ {
		if got := fuzzymatch.Strcmp95Score(a, b); got != want {
			t.Fatalf("Strcmp95Score non-deterministic at iteration %d: got %g; want %g", i, got, want)
		}
	}
}

// ---------------------------------------------------------------------------
// LCSStr property tests (plan 04-02)
//
// Standard Phase 2 invariants (RangeBounds, Identity, Symmetric, NoNaN, NoInf,
// NoNegativeZero) for both byte and rune surfaces, PLUS three LCSStr-specific
// invariants:
//   - IsSubstringOfBoth — the returned substring is a substring of both a and b
//   - LengthMatchesScore — score == 2·len(LongestCommonSubstring(a,b))/(la+lb)
//   - LeftmostTieBreak — hand-curated tie-break inputs return leftmost-in-a
//
// LCSStr is not a metric (it's a similarity), so no triangle inequality.
// ---------------------------------------------------------------------------

// TestProp_LCSStrScore_RangeBounds asserts the score is in [0.0, 1.0] for any
// pair of strings. DET-04 range-bounds invariant.
func TestProp_LCSStrScore_RangeBounds(t *testing.T) {
	f := func(a, b string) bool {
		s := fuzzymatch.LCSStrScore(a, b)
		return s >= 0.0 && s <= 1.0
	}
	if err := quick.Check(f, nil); err != nil {
		t.Errorf("LCSStrScore out of [0,1]: %v", err)
	}
}

// TestProp_LCSStrScore_Identity asserts Score(x, x) == 1.0 for any non-empty
// string x. Both-empty is also 1.0 by convention but covered by unit tests.
func TestProp_LCSStrScore_Identity(t *testing.T) {
	f := func(x string) bool {
		if x == "" {
			return true // both-empty handled in unit tests
		}
		return fuzzymatch.LCSStrScore(x, x) == 1.0
	}
	if err := quick.Check(f, nil); err != nil {
		t.Errorf("LCSStrScore identity violated: %v", err)
	}
}

// TestProp_LCSStrScore_Symmetric asserts Score(a, b) == Score(b, a) for any
// (a, b). The substring relation is symmetric and the normalisation
// 2·len(lcs)/(la+lb) is symmetric in (a, b).
func TestProp_LCSStrScore_Symmetric(t *testing.T) {
	f := func(a, b string) bool {
		return fuzzymatch.LCSStrScore(a, b) == fuzzymatch.LCSStrScore(b, a)
	}
	if err := quick.Check(f, nil); err != nil {
		t.Errorf("LCSStrScore not symmetric: %v", err)
	}
}

// TestProp_LCSStrScore_NoNaN asserts the score is never NaN. The empty-input
// guard prevents the 0/0 division; the algorithm uses only +, *, /, and
// comparisons so no other NaN paths exist.
func TestProp_LCSStrScore_NoNaN(t *testing.T) {
	f := func(a, b string) bool {
		return !math.IsNaN(fuzzymatch.LCSStrScore(a, b))
	}
	if err := quick.Check(f, nil); err != nil {
		t.Errorf("LCSStrScore produced NaN: %v", err)
	}
}

// TestProp_LCSStrScore_NoInf asserts the score never returns ±Inf. The
// numerator and denominator are finite integers (max len(a)+len(b) ~= 256 for
// quick.Check default inputs); their quotient is finite.
func TestProp_LCSStrScore_NoInf(t *testing.T) {
	f := func(a, b string) bool {
		return !math.IsInf(fuzzymatch.LCSStrScore(a, b), 0)
	}
	if err := quick.Check(f, nil); err != nil {
		t.Errorf("LCSStrScore produced Inf: %v", err)
	}
}

// TestProp_LCSStrScore_NoNegativeZero asserts that when the score is 0.0 it
// is positive zero (+0.0), not negative zero (-0.0).
func TestProp_LCSStrScore_NoNegativeZero(t *testing.T) {
	f := func(a, b string) bool {
		s := fuzzymatch.LCSStrScore(a, b)
		return s != 0.0 || !math.Signbit(s)
	}
	if err := quick.Check(f, nil); err != nil {
		t.Errorf("LCSStrScore produced -0.0: %v", err)
	}
}

// TestProp_LCSStrScoreRunes_RangeBounds: rune path range bounds.
func TestProp_LCSStrScoreRunes_RangeBounds(t *testing.T) {
	f := func(a, b string) bool {
		s := fuzzymatch.LCSStrScoreRunes(a, b)
		return s >= 0.0 && s <= 1.0
	}
	if err := quick.Check(f, nil); err != nil {
		t.Errorf("LCSStrScoreRunes out of [0,1]: %v", err)
	}
}

// TestProp_LCSStrScoreRunes_Identity: rune path identity.
func TestProp_LCSStrScoreRunes_Identity(t *testing.T) {
	f := func(x string) bool {
		if x == "" {
			return true
		}
		return fuzzymatch.LCSStrScoreRunes(x, x) == 1.0
	}
	if err := quick.Check(f, nil); err != nil {
		t.Errorf("LCSStrScoreRunes identity violated: %v", err)
	}
}

// TestProp_LCSStrScoreRunes_Symmetric: rune path symmetry.
func TestProp_LCSStrScoreRunes_Symmetric(t *testing.T) {
	f := func(a, b string) bool {
		return fuzzymatch.LCSStrScoreRunes(a, b) == fuzzymatch.LCSStrScoreRunes(b, a)
	}
	if err := quick.Check(f, nil); err != nil {
		t.Errorf("LCSStrScoreRunes not symmetric: %v", err)
	}
}

// TestProp_LCSStrScoreRunes_NoNaN: rune path NaN guard.
func TestProp_LCSStrScoreRunes_NoNaN(t *testing.T) {
	f := func(a, b string) bool {
		return !math.IsNaN(fuzzymatch.LCSStrScoreRunes(a, b))
	}
	if err := quick.Check(f, nil); err != nil {
		t.Errorf("LCSStrScoreRunes produced NaN: %v", err)
	}
}

// TestProp_LCSStrScoreRunes_NoInf: rune path Inf guard.
func TestProp_LCSStrScoreRunes_NoInf(t *testing.T) {
	f := func(a, b string) bool {
		return !math.IsInf(fuzzymatch.LCSStrScoreRunes(a, b), 0)
	}
	if err := quick.Check(f, nil); err != nil {
		t.Errorf("LCSStrScoreRunes produced Inf: %v", err)
	}
}

// TestProp_LCSStrScoreRunes_NoNegativeZero: rune path -0.0 guard.
func TestProp_LCSStrScoreRunes_NoNegativeZero(t *testing.T) {
	f := func(a, b string) bool {
		s := fuzzymatch.LCSStrScoreRunes(a, b)
		return s != 0.0 || !math.Signbit(s)
	}
	if err := quick.Check(f, nil); err != nil {
		t.Errorf("LCSStrScoreRunes produced -0.0: %v", err)
	}
}

// TestProp_LongestCommonSubstring_IsSubstringOfBoth asserts that the returned
// substring is actually a substring of BOTH a and b. The empty-string return
// case is allowed (both-empty AND no-overlap legitimately return "").
//
// LCSStr-specific structural invariant: confirms the DP kernel correctly
// reports a contiguous shared segment, not e.g. a non-contiguous LCS or a
// substring of only one input.
func TestProp_LongestCommonSubstring_IsSubstringOfBoth(t *testing.T) {
	f := func(a, b string) bool {
		got := fuzzymatch.LongestCommonSubstring(a, b)
		if got == "" {
			return true // both-empty or no-overlap; both legitimate
		}
		return strings.Contains(a, got) && strings.Contains(b, got)
	}
	if err := quick.Check(f, nil); err != nil {
		t.Errorf("LongestCommonSubstring not a substring of both inputs: %v", err)
	}
}

// TestProp_LongestCommonSubstring_LengthMatchesScore asserts the relationship
// between the substring-returning surface and the score-returning surface:
//
//	|2·len(LongestCommonSubstring(a,b))/(la+lb) - LCSStrScore(a,b)| < 1e-9
//
// Special case: both-empty inputs return substring "" with score 1.0; the
// formula 2·0/0 would NaN, so we handle that case explicitly.
//
// This invariant prevents a subtle bug where the substring path and the
// length-only path could compute different LCS lengths (e.g. one uses strict
// `>` and the other `>=`).
func TestProp_LongestCommonSubstring_LengthMatchesScore(t *testing.T) {
	const eps = 1e-9
	f := func(a, b string) bool {
		got := fuzzymatch.LongestCommonSubstring(a, b)
		score := fuzzymatch.LCSStrScore(a, b)
		la, lb := len(a), len(b)
		if la == 0 && lb == 0 {
			return got == "" && score == 1.0
		}
		expected := 2.0 * float64(len(got)) / float64(la+lb)
		return math.Abs(score-expected) < eps
	}
	if err := quick.Check(f, nil); err != nil {
		t.Errorf("len(LongestCommonSubstring) does not match LCSStrScore numerator: %v", err)
	}
}

// TestProp_LongestCommonSubstring_LeftmostTieBreak is the LOAD-BEARING
// regression test for RESEARCH.md Pitfall 4. Hand-curated tied-candidate
// inputs (multiple longest common substrings of equal length exist in `a`);
// the strict-`>` max-update guarantees the LEFTMOST occurrence in `a` is
// returned. Any drift to `>=` flips the tie-break and breaks these
// assertions.
//
// Each row asserts: (1) LongestCommonSubstring(a, b) == wantSub; (2) the
// substring starts at index 0 (or wantStart) in `a` — verified via
// strings.Index, which returns the leftmost match index by definition.
func TestProp_LongestCommonSubstring_LeftmostTieBreak(t *testing.T) {
	tests := []struct {
		a, b      string
		wantSub   string
		wantStart int
	}{
		{"abcXYZabc", "abc", "abc", 0},
		{"xy_abc_xy_abc", "abc", "abc", 3},
		{"aaa", "aa", "aa", 0},
		{"foo_bar_foo", "foo", "foo", 0},
		{"mississippi", "issi", "issi", 1},
		// Three tied length-2 windows in `a`:
		{"abXabYabZ", "ab", "ab", 0},
		// Tied length-1 matches: "ab" and "ba" share both 'a' and 'b' as
		// length-1 matches; the DP iterates outer=a then inner=b, so the
		// first match RECORDED is at (i=1, j=1) → a[0]='a' matching b[0]='a'
		// → maxEnd=1, substring "a". Leftmost-in-`a`-by-ENDING-INDEX.
		{"ab", "ba", "a", 0},
	}
	for _, tt := range tests {
		got := fuzzymatch.LongestCommonSubstring(tt.a, tt.b)
		if got != tt.wantSub {
			t.Errorf("LongestCommonSubstring(%q, %q) = %q; want %q (leftmost-in-a)",
				tt.a, tt.b, got, tt.wantSub)
			continue
		}
		if idx := strings.Index(tt.a, got); idx != tt.wantStart {
			t.Errorf("LongestCommonSubstring(%q, %q) found %q at index %d; want leftmost %d",
				tt.a, tt.b, got, idx, tt.wantStart)
		}
	}
}

// ---------------------------------------------------------------------------
// Ratcliff-Obershelp property tests (plan 04-03)
//
// FIVE standard invariants (RangeBounds, Identity, NoNaN, NoInf, NoNegativeZero)
// for both byte and rune surfaces — TEN property tests total — PLUS one
// algorithm-specific table-driven test
// (TestRatcliffObershelpScore_AtLeastLevenshtein_OnSubstringContainment).
//
// NB: TestProp_RatcliffObershelpScore_Symmetric is INTENTIONALLY OMITTED per
//     OQ-1 resolution (LOCKED 2026-05-14). Ratcliff-Obershelp is asymmetric
//     by design to preserve byte-for-byte difflib equivalence. See
//     ratcliff_obershelp.go's godoc and CONTEXT.md §4 for the rationale.
//     TestRatcliffObershelp_AsymmetryPin in ratcliff_obershelp_test.go is
//     the load-bearing regression guard.
//
// Ratcliff-Obershelp is not a metric (it's a similarity), so no triangle
// inequality.
// ---------------------------------------------------------------------------

// TestProp_RatcliffObershelpScore_RangeBounds asserts the score is in
// [0.0, 1.0] for any pair of strings. DET-04 range-bounds invariant.
func TestProp_RatcliffObershelpScore_RangeBounds(t *testing.T) {
	f := func(a, b string) bool {
		s := fuzzymatch.RatcliffObershelpScore(a, b)
		return s >= 0.0 && s <= 1.0
	}
	if err := quick.Check(f, nil); err != nil {
		t.Errorf("RatcliffObershelpScore out of [0,1]: %v", err)
	}
}

// TestProp_RatcliffObershelpScore_Identity asserts Score(x, x) == 1.0 for
// any non-empty string x. Both-empty is also 1.0 by convention but covered
// by unit tests.
func TestProp_RatcliffObershelpScore_Identity(t *testing.T) {
	f := func(x string) bool {
		if x == "" {
			return true // both-empty handled in unit tests
		}
		return fuzzymatch.RatcliffObershelpScore(x, x) == 1.0
	}
	if err := quick.Check(f, nil); err != nil {
		t.Errorf("RatcliffObershelpScore identity violated: %v", err)
	}
}

// TestProp_RatcliffObershelpScore_NoNaN asserts the score is never NaN.
// The empty-input guard prevents the 0/0 division; the algorithm uses only
// +, *, /, and comparisons so no other NaN paths exist.
func TestProp_RatcliffObershelpScore_NoNaN(t *testing.T) {
	f := func(a, b string) bool {
		return !math.IsNaN(fuzzymatch.RatcliffObershelpScore(a, b))
	}
	if err := quick.Check(f, nil); err != nil {
		t.Errorf("RatcliffObershelpScore produced NaN: %v", err)
	}
}

// TestProp_RatcliffObershelpScore_NoInf asserts the score never returns
// ±Inf. The numerator and denominator are finite integers (max
// len(a)+len(b) ~= 256 for quick.Check default inputs); their quotient is
// finite.
func TestProp_RatcliffObershelpScore_NoInf(t *testing.T) {
	f := func(a, b string) bool {
		return !math.IsInf(fuzzymatch.RatcliffObershelpScore(a, b), 0)
	}
	if err := quick.Check(f, nil); err != nil {
		t.Errorf("RatcliffObershelpScore produced Inf: %v", err)
	}
}

// TestProp_RatcliffObershelpScore_NoNegativeZero asserts that when the
// score is 0.0 it is positive zero (+0.0), not negative zero (-0.0).
func TestProp_RatcliffObershelpScore_NoNegativeZero(t *testing.T) {
	f := func(a, b string) bool {
		s := fuzzymatch.RatcliffObershelpScore(a, b)
		return s != 0.0 || !math.Signbit(s)
	}
	if err := quick.Check(f, nil); err != nil {
		t.Errorf("RatcliffObershelpScore produced -0.0: %v", err)
	}
}

// TestProp_RatcliffObershelpScoreRunes_RangeBounds: rune path range bounds.
func TestProp_RatcliffObershelpScoreRunes_RangeBounds(t *testing.T) {
	f := func(a, b string) bool {
		s := fuzzymatch.RatcliffObershelpScoreRunes(a, b)
		return s >= 0.0 && s <= 1.0
	}
	if err := quick.Check(f, nil); err != nil {
		t.Errorf("RatcliffObershelpScoreRunes out of [0,1]: %v", err)
	}
}

// TestProp_RatcliffObershelpScoreRunes_Identity: rune path identity.
func TestProp_RatcliffObershelpScoreRunes_Identity(t *testing.T) {
	f := func(x string) bool {
		if x == "" {
			return true
		}
		return fuzzymatch.RatcliffObershelpScoreRunes(x, x) == 1.0
	}
	if err := quick.Check(f, nil); err != nil {
		t.Errorf("RatcliffObershelpScoreRunes identity violated: %v", err)
	}
}

// TestProp_RatcliffObershelpScoreRunes_NoNaN: rune path NaN guard.
func TestProp_RatcliffObershelpScoreRunes_NoNaN(t *testing.T) {
	f := func(a, b string) bool {
		return !math.IsNaN(fuzzymatch.RatcliffObershelpScoreRunes(a, b))
	}
	if err := quick.Check(f, nil); err != nil {
		t.Errorf("RatcliffObershelpScoreRunes produced NaN: %v", err)
	}
}

// TestProp_RatcliffObershelpScoreRunes_NoInf: rune path Inf guard.
func TestProp_RatcliffObershelpScoreRunes_NoInf(t *testing.T) {
	f := func(a, b string) bool {
		return !math.IsInf(fuzzymatch.RatcliffObershelpScoreRunes(a, b), 0)
	}
	if err := quick.Check(f, nil); err != nil {
		t.Errorf("RatcliffObershelpScoreRunes produced Inf: %v", err)
	}
}

// TestProp_RatcliffObershelpScoreRunes_NoNegativeZero: rune path -0.0 guard.
func TestProp_RatcliffObershelpScoreRunes_NoNegativeZero(t *testing.T) {
	f := func(a, b string) bool {
		s := fuzzymatch.RatcliffObershelpScoreRunes(a, b)
		return s != 0.0 || !math.Signbit(s)
	}
	if err := quick.Check(f, nil); err != nil {
		t.Errorf("RatcliffObershelpScoreRunes produced -0.0: %v", err)
	}
}

// ---------------------------------------------------------------------------
// Q-Gram Jaccard property tests (plan 05-01)
// ---------------------------------------------------------------------------

// qgramJaccardN coerces an arbitrary int into the [1, 5] inclusive range
// used by the Q-Gram Jaccard property tests. Negative and zero n values
// are mapped into the valid range; the n parameter would otherwise
// panic per CONTEXT.md §5 LOCKED, but that contract is unit-tested
// separately by TestQGramJaccard_PanicsOnInvalidN — the property tests
// exercise the [0, 1] score-range invariants.
func qgramJaccardN(n int) int {
	if n < 0 {
		n = -n
	}
	return (n % 5) + 1
}

// TestProp_QGramJaccardScore_RangeBounds asserts the byte-path score
// stays in [0.0, 1.0] for any (a, b, n) triple. Inline NaN/Inf guards
// document the joint invariant; dedicated _NoNaN / _NoInf tests below
// retest each guard in isolation for documentation clarity.
func TestProp_QGramJaccardScore_RangeBounds(t *testing.T) {
	f := func(a, b string, n int) bool {
		s := fuzzymatch.QGramJaccardScore(a, b, qgramJaccardN(n))
		return s >= 0.0 && s <= 1.0 && !math.IsNaN(s) && !math.IsInf(s, 0)
	}
	if err := quick.Check(f, nil); err != nil {
		t.Errorf("QGramJaccardScore out of [0,1] or non-finite: %v", err)
	}
}

// TestProp_QGramJaccardScore_Identity asserts Score(x, x, n) == 1.0
// EXACTLY for any non-empty x and any n >= 1 — the identity short-circuit
// fires before extraction and the result is the literal 1.0.
func TestProp_QGramJaccardScore_Identity(t *testing.T) {
	f := func(x string, n int) bool {
		if x == "" {
			return true // both-empty handled by unit tests
		}
		return fuzzymatch.QGramJaccardScore(x, x, qgramJaccardN(n)) == 1.0
	}
	if err := quick.Check(f, nil); err != nil {
		t.Errorf("QGramJaccardScore identity violated: %v", err)
	}
}

// TestProp_QGramJaccardScore_Symmetric asserts Score(a, b, n) == Score(b, a, n)
// EXACTLY (not within tolerance) — set-Jaccard is symmetric and the
// integer-derived single division produces bit-identical output.
func TestProp_QGramJaccardScore_Symmetric(t *testing.T) {
	f := func(a, b string, n int) bool {
		nn := qgramJaccardN(n)
		return fuzzymatch.QGramJaccardScore(a, b, nn) == fuzzymatch.QGramJaccardScore(b, a, nn)
	}
	if err := quick.Check(f, nil); err != nil {
		t.Errorf("QGramJaccardScore not symmetric: %v", err)
	}
}

// TestProp_QGramJaccardScore_NoNaN asserts the byte-path score never
// returns NaN. The both-empty + one-empty + identity short-circuits
// gate away the only potential 0/0 paths; the explicit
// jaccardFromQGramMaps len-check provides the secondary guard.
func TestProp_QGramJaccardScore_NoNaN(t *testing.T) {
	f := func(a, b string, n int) bool {
		return !math.IsNaN(fuzzymatch.QGramJaccardScore(a, b, qgramJaccardN(n)))
	}
	if err := quick.Check(f, nil); err != nil {
		t.Errorf("QGramJaccardScore produced NaN: %v", err)
	}
}

// TestProp_QGramJaccardScore_NoInf asserts the byte-path score never
// returns ±Inf. Numerator and denominator are bounded integers fitting
// in float64 (counts up to 2^53 are exact); the single division never
// overflows.
func TestProp_QGramJaccardScore_NoInf(t *testing.T) {
	f := func(a, b string, n int) bool {
		return !math.IsInf(fuzzymatch.QGramJaccardScore(a, b, qgramJaccardN(n)), 0)
	}
	if err := quick.Check(f, nil); err != nil {
		t.Errorf("QGramJaccardScore produced Inf: %v", err)
	}
}

// TestProp_QGramJaccardScore_NoNegativeZero asserts that when the byte-path
// score is 0.0 it is positive zero, not negative zero. The intersection
// cardinality is a non-negative integer; float64(0) / float64(positive)
// is +0.0 in IEEE-754.
func TestProp_QGramJaccardScore_NoNegativeZero(t *testing.T) {
	f := func(a, b string, n int) bool {
		s := fuzzymatch.QGramJaccardScore(a, b, qgramJaccardN(n))
		return s != 0.0 || !math.Signbit(s)
	}
	if err := quick.Check(f, nil); err != nil {
		t.Errorf("QGramJaccardScore produced -0.0: %v", err)
	}
}

// TestProp_QGramJaccardScoreRunes_RangeBounds: rune-path mirror of
// _RangeBounds.
func TestProp_QGramJaccardScoreRunes_RangeBounds(t *testing.T) {
	f := func(a, b string, n int) bool {
		s := fuzzymatch.QGramJaccardScoreRunes(a, b, qgramJaccardN(n))
		return s >= 0.0 && s <= 1.0 && !math.IsNaN(s) && !math.IsInf(s, 0)
	}
	if err := quick.Check(f, nil); err != nil {
		t.Errorf("QGramJaccardScoreRunes out of [0,1] or non-finite: %v", err)
	}
}

// TestProp_QGramJaccardScoreRunes_Identity: rune-path identity.
func TestProp_QGramJaccardScoreRunes_Identity(t *testing.T) {
	f := func(x string, n int) bool {
		if x == "" {
			return true
		}
		return fuzzymatch.QGramJaccardScoreRunes(x, x, qgramJaccardN(n)) == 1.0
	}
	if err := quick.Check(f, nil); err != nil {
		t.Errorf("QGramJaccardScoreRunes identity violated: %v", err)
	}
}

// TestProp_QGramJaccardScoreRunes_Symmetric: rune-path symmetry.
func TestProp_QGramJaccardScoreRunes_Symmetric(t *testing.T) {
	f := func(a, b string, n int) bool {
		nn := qgramJaccardN(n)
		return fuzzymatch.QGramJaccardScoreRunes(a, b, nn) == fuzzymatch.QGramJaccardScoreRunes(b, a, nn)
	}
	if err := quick.Check(f, nil); err != nil {
		t.Errorf("QGramJaccardScoreRunes not symmetric: %v", err)
	}
}

// TestProp_QGramJaccardScoreRunes_NoNaN: rune-path NaN guard.
func TestProp_QGramJaccardScoreRunes_NoNaN(t *testing.T) {
	f := func(a, b string, n int) bool {
		return !math.IsNaN(fuzzymatch.QGramJaccardScoreRunes(a, b, qgramJaccardN(n)))
	}
	if err := quick.Check(f, nil); err != nil {
		t.Errorf("QGramJaccardScoreRunes produced NaN: %v", err)
	}
}

// TestProp_QGramJaccardScoreRunes_NoInf: rune-path Inf guard.
func TestProp_QGramJaccardScoreRunes_NoInf(t *testing.T) {
	f := func(a, b string, n int) bool {
		return !math.IsInf(fuzzymatch.QGramJaccardScoreRunes(a, b, qgramJaccardN(n)), 0)
	}
	if err := quick.Check(f, nil); err != nil {
		t.Errorf("QGramJaccardScoreRunes produced Inf: %v", err)
	}
}

// TestProp_QGramJaccardScoreRunes_NoNegativeZero: rune-path -0.0 guard.
func TestProp_QGramJaccardScoreRunes_NoNegativeZero(t *testing.T) {
	f := func(a, b string, n int) bool {
		s := fuzzymatch.QGramJaccardScoreRunes(a, b, qgramJaccardN(n))
		return s != 0.0 || !math.Signbit(s)
	}
	if err := quick.Check(f, nil); err != nil {
		t.Errorf("QGramJaccardScoreRunes produced -0.0: %v", err)
	}
}

// TestProp_QGramJaccardScore_DeterministicAcrossRuns asserts that 1000
// sequential calls on the same (a, b, n) input produce byte-identical
// output. PITFALLS §14 closure carried forward from Phase 4 — guards
// against any future regression that might re-introduce map-iteration
// order dependence on the output path.
//
// Compares via math.Float64bits to detect bit-level differences (e.g.
// +0.0 vs -0.0, signalling vs quiet NaN — even though the algorithm
// emits neither).
func TestProp_QGramJaccardScore_DeterministicAcrossRuns(t *testing.T) {
	const iterations = 1000
	const a = "AGCT"
	const b = "AGCTAGCT"
	const n = 2
	baseline := fuzzymatch.QGramJaccardScore(a, b, n)
	baselineBits := math.Float64bits(baseline)
	for i := 0; i < iterations; i++ {
		got := fuzzymatch.QGramJaccardScore(a, b, n)
		if math.Float64bits(got) != baselineBits {
			t.Fatalf("iteration %d: QGramJaccardScore(%q,%q,%d) = %.17g (bits=%x); baseline = %.17g (bits=%x)",
				i, a, b, n, got, math.Float64bits(got), baseline, baselineBits)
		}
	}
	// Mirror gate on the rune surface.
	baselineR := fuzzymatch.QGramJaccardScoreRunes("café", "cafe", 2)
	baselineRBits := math.Float64bits(baselineR)
	for i := 0; i < iterations; i++ {
		got := fuzzymatch.QGramJaccardScoreRunes("café", "cafe", 2)
		if math.Float64bits(got) != baselineRBits {
			t.Fatalf("iteration %d (rune): got %.17g (bits=%x); baseline = %.17g (bits=%x)",
				i, got, math.Float64bits(got), baselineR, baselineRBits)
		}
	}
}

// ---------------------------------------------------------------------------
// Sørensen-Dice property tests (plan 05-02)
// ---------------------------------------------------------------------------

// sorensenDiceN coerces an arbitrary int into the [1, 5] inclusive range
// used by the Sørensen-Dice property tests. Negative and zero n values
// are mapped into the valid range; the n parameter would otherwise
// panic per CONTEXT.md §5 LOCKED, but that contract is unit-tested
// separately by TestSorensenDice_PanicsOnInvalidN — the property tests
// exercise the [0, 1] score-range invariants.
func sorensenDiceN(n int) int {
	if n < 0 {
		n = -n
	}
	return (n % 5) + 1
}

// TestProp_SorensenDiceScore_RangeBounds asserts the byte-path score
// stays in [0.0, 1.0] for any (a, b, n) triple. Inline NaN/Inf guards
// document the joint invariant; dedicated _NoNaN / _NoInf tests below
// retest each guard in isolation for documentation clarity.
func TestProp_SorensenDiceScore_RangeBounds(t *testing.T) {
	f := func(a, b string, n int) bool {
		s := fuzzymatch.SorensenDiceScore(a, b, sorensenDiceN(n))
		return s >= 0.0 && s <= 1.0 && !math.IsNaN(s) && !math.IsInf(s, 0)
	}
	if err := quick.Check(f, nil); err != nil {
		t.Errorf("SorensenDiceScore out of [0,1] or non-finite: %v", err)
	}
}

// TestProp_SorensenDiceScore_Identity asserts Score(x, x, n) == 1.0
// EXACTLY for any non-empty x and any n >= 1 — the identity short-circuit
// fires before extraction and the result is the literal 1.0.
func TestProp_SorensenDiceScore_Identity(t *testing.T) {
	f := func(x string, n int) bool {
		if x == "" {
			return true // both-empty handled by unit tests
		}
		return fuzzymatch.SorensenDiceScore(x, x, sorensenDiceN(n)) == 1.0
	}
	if err := quick.Check(f, nil); err != nil {
		t.Errorf("SorensenDiceScore identity violated: %v", err)
	}
}

// TestProp_SorensenDiceScore_Symmetric asserts Score(a, b, n) == Score(b, a, n)
// EXACTLY (not within tolerance) — Sørensen-Dice is symmetric and the
// integer-derived single multiplication-then-division produces
// bit-identical output.
func TestProp_SorensenDiceScore_Symmetric(t *testing.T) {
	f := func(a, b string, n int) bool {
		nn := sorensenDiceN(n)
		return fuzzymatch.SorensenDiceScore(a, b, nn) == fuzzymatch.SorensenDiceScore(b, a, nn)
	}
	if err := quick.Check(f, nil); err != nil {
		t.Errorf("SorensenDiceScore not symmetric: %v", err)
	}
}

// TestProp_SorensenDiceScore_NoNaN asserts the byte-path score never
// returns NaN. The both-empty + one-empty + identity short-circuits
// gate away the only potential 0/0 paths; the explicit
// diceFromQGramMaps len-check provides the secondary guard.
func TestProp_SorensenDiceScore_NoNaN(t *testing.T) {
	f := func(a, b string, n int) bool {
		return !math.IsNaN(fuzzymatch.SorensenDiceScore(a, b, sorensenDiceN(n)))
	}
	if err := quick.Check(f, nil); err != nil {
		t.Errorf("SorensenDiceScore produced NaN: %v", err)
	}
}

// TestProp_SorensenDiceScore_NoInf asserts the byte-path score never
// returns ±Inf. Numerator and denominator are bounded integers fitting
// in float64 (counts up to 2^53 are exact); the single
// multiplication-then-division never overflows.
func TestProp_SorensenDiceScore_NoInf(t *testing.T) {
	f := func(a, b string, n int) bool {
		return !math.IsInf(fuzzymatch.SorensenDiceScore(a, b, sorensenDiceN(n)), 0)
	}
	if err := quick.Check(f, nil); err != nil {
		t.Errorf("SorensenDiceScore produced Inf: %v", err)
	}
}

// TestProp_SorensenDiceScore_NoNegativeZero asserts that when the
// byte-path score is 0.0 it is positive zero, not negative zero. The
// intersection cardinality is a non-negative integer; 2·0/(positive)
// is +0.0 in IEEE-754.
func TestProp_SorensenDiceScore_NoNegativeZero(t *testing.T) {
	f := func(a, b string, n int) bool {
		s := fuzzymatch.SorensenDiceScore(a, b, sorensenDiceN(n))
		return s != 0.0 || !math.Signbit(s)
	}
	if err := quick.Check(f, nil); err != nil {
		t.Errorf("SorensenDiceScore produced -0.0: %v", err)
	}
}

// TestProp_SorensenDiceScoreRunes_RangeBounds: rune-path mirror of
// _RangeBounds.
func TestProp_SorensenDiceScoreRunes_RangeBounds(t *testing.T) {
	f := func(a, b string, n int) bool {
		s := fuzzymatch.SorensenDiceScoreRunes(a, b, sorensenDiceN(n))
		return s >= 0.0 && s <= 1.0 && !math.IsNaN(s) && !math.IsInf(s, 0)
	}
	if err := quick.Check(f, nil); err != nil {
		t.Errorf("SorensenDiceScoreRunes out of [0,1] or non-finite: %v", err)
	}
}

// TestProp_SorensenDiceScoreRunes_Identity: rune-path identity.
func TestProp_SorensenDiceScoreRunes_Identity(t *testing.T) {
	f := func(x string, n int) bool {
		if x == "" {
			return true
		}
		return fuzzymatch.SorensenDiceScoreRunes(x, x, sorensenDiceN(n)) == 1.0
	}
	if err := quick.Check(f, nil); err != nil {
		t.Errorf("SorensenDiceScoreRunes identity violated: %v", err)
	}
}

// TestProp_SorensenDiceScoreRunes_Symmetric: rune-path symmetry.
func TestProp_SorensenDiceScoreRunes_Symmetric(t *testing.T) {
	f := func(a, b string, n int) bool {
		nn := sorensenDiceN(n)
		return fuzzymatch.SorensenDiceScoreRunes(a, b, nn) == fuzzymatch.SorensenDiceScoreRunes(b, a, nn)
	}
	if err := quick.Check(f, nil); err != nil {
		t.Errorf("SorensenDiceScoreRunes not symmetric: %v", err)
	}
}

// TestProp_SorensenDiceScoreRunes_NoNaN: rune-path NaN guard.
func TestProp_SorensenDiceScoreRunes_NoNaN(t *testing.T) {
	f := func(a, b string, n int) bool {
		return !math.IsNaN(fuzzymatch.SorensenDiceScoreRunes(a, b, sorensenDiceN(n)))
	}
	if err := quick.Check(f, nil); err != nil {
		t.Errorf("SorensenDiceScoreRunes produced NaN: %v", err)
	}
}

// TestProp_SorensenDiceScoreRunes_NoInf: rune-path Inf guard.
func TestProp_SorensenDiceScoreRunes_NoInf(t *testing.T) {
	f := func(a, b string, n int) bool {
		return !math.IsInf(fuzzymatch.SorensenDiceScoreRunes(a, b, sorensenDiceN(n)), 0)
	}
	if err := quick.Check(f, nil); err != nil {
		t.Errorf("SorensenDiceScoreRunes produced Inf: %v", err)
	}
}

// TestProp_SorensenDiceScoreRunes_NoNegativeZero: rune-path -0.0 guard.
func TestProp_SorensenDiceScoreRunes_NoNegativeZero(t *testing.T) {
	f := func(a, b string, n int) bool {
		s := fuzzymatch.SorensenDiceScoreRunes(a, b, sorensenDiceN(n))
		return s != 0.0 || !math.Signbit(s)
	}
	if err := quick.Check(f, nil); err != nil {
		t.Errorf("SorensenDiceScoreRunes produced -0.0: %v", err)
	}
}

// TestProp_SorensenDiceScore_DeterministicAcrossRuns asserts that 1000
// sequential calls on the same (a, b, n) input produce byte-identical
// output. PITFALLS §14 closure carried forward from plan 05-01 — guards
// against any future regression that might re-introduce map-iteration
// order dependence on the output path.
//
// Compares via math.Float64bits to detect bit-level differences (e.g.
// +0.0 vs -0.0, signalling vs quiet NaN — even though the algorithm
// emits neither). Uses the load-bearing RV-D1 night/nacht/n=2 pair on
// the byte surface, mirrored on the café/cafe/n=2 rune surface.
func TestProp_SorensenDiceScore_DeterministicAcrossRuns(t *testing.T) {
	const iterations = 1000
	const a = "night"
	const b = "nacht"
	const n = 2
	baseline := fuzzymatch.SorensenDiceScore(a, b, n)
	baselineBits := math.Float64bits(baseline)
	for i := 0; i < iterations; i++ {
		got := fuzzymatch.SorensenDiceScore(a, b, n)
		if math.Float64bits(got) != baselineBits {
			t.Fatalf("iteration %d: SorensenDiceScore(%q,%q,%d) = %.17g (bits=%x); baseline = %.17g (bits=%x)",
				i, a, b, n, got, math.Float64bits(got), baseline, baselineBits)
		}
	}
	// Mirror gate on the rune surface.
	baselineR := fuzzymatch.SorensenDiceScoreRunes("café", "cafe", 2)
	baselineRBits := math.Float64bits(baselineR)
	for i := 0; i < iterations; i++ {
		got := fuzzymatch.SorensenDiceScoreRunes("café", "cafe", 2)
		if math.Float64bits(got) != baselineRBits {
			t.Fatalf("iteration %d (rune): got %.17g (bits=%x); baseline = %.17g (bits=%x)",
				i, got, math.Float64bits(got), baselineR, baselineRBits)
		}
	}
}

// ---------------------------------------------------------------------------
// Cosine property tests (plan 05-03)
// ---------------------------------------------------------------------------

// cosineN coerces an arbitrary int into the [1, 5] inclusive range used
// by the Cosine property tests. Negative and zero n values are mapped
// into the valid range; the n parameter would otherwise panic per
// CONTEXT.md §5 LOCKED, but that contract is unit-tested separately by
// TestCosine_PanicsOnInvalidN — the property tests exercise the [0, 1]
// score-range invariants.
func cosineN(n int) int {
	if n < 0 {
		n = -n
	}
	return (n % 5) + 1
}

// TestProp_CosineScore_RangeBounds asserts the byte-path score stays in
// [0.0, 1.0] for any (a, b, n) triple. Inline NaN/Inf guards document
// the joint invariant; dedicated _NoNaN / _NoInf tests below retest
// each guard in isolation for documentation clarity.
func TestProp_CosineScore_RangeBounds(t *testing.T) {
	f := func(a, b string, n int) bool {
		s := fuzzymatch.CosineScore(a, b, cosineN(n))
		return s >= 0.0 && s <= 1.0 && !math.IsNaN(s) && !math.IsInf(s, 0)
	}
	if err := quick.Check(f, nil); err != nil {
		t.Errorf("CosineScore out of [0,1] or non-finite: %v", err)
	}
}

// TestProp_CosineScore_Identity asserts Score(x, x, n) == 1.0 EXACTLY
// for any non-empty x and any n >= 1 — the identity short-circuit
// fires before extraction and the result is the literal 1.0.
func TestProp_CosineScore_Identity(t *testing.T) {
	f := func(x string, n int) bool {
		if x == "" {
			return true // both-empty handled by unit tests
		}
		return fuzzymatch.CosineScore(x, x, cosineN(n)) == 1.0
	}
	if err := quick.Check(f, nil); err != nil {
		t.Errorf("CosineScore identity violated: %v", err)
	}
}

// TestProp_CosineScore_Symmetric asserts Score(a, b, n) == Score(b, a, n)
// EXACTLY (not within tolerance) — Cosine is symmetric and the
// sorted-key iteration is canonical regardless of input argument
// order, producing bit-identical float64 output.
func TestProp_CosineScore_Symmetric(t *testing.T) {
	f := func(a, b string, n int) bool {
		nn := cosineN(n)
		return fuzzymatch.CosineScore(a, b, nn) == fuzzymatch.CosineScore(b, a, nn)
	}
	if err := quick.Check(f, nil); err != nil {
		t.Errorf("CosineScore not symmetric: %v", err)
	}
}

// TestProp_CosineScore_NoNaN asserts the byte-path score never returns
// NaN. The both-empty + one-empty + identity short-circuits gate away
// the only potential 0/0 paths; the explicit cosineFromQGramMaps
// len-check provides the secondary guard.
func TestProp_CosineScore_NoNaN(t *testing.T) {
	f := func(a, b string, n int) bool {
		return !math.IsNaN(fuzzymatch.CosineScore(a, b, cosineN(n)))
	}
	if err := quick.Check(f, nil); err != nil {
		t.Errorf("CosineScore produced NaN: %v", err)
	}
}

// TestProp_CosineScore_NoInf asserts the byte-path score never returns
// ±Inf. The numerator is bounded by min(‖A‖², ‖B‖²) ≤ ‖A‖·‖B‖ (Cauchy-
// Schwarz); the denominator is the product of two non-negative norms
// that fit in float64 for any input where len(a)+len(b) < 2^53. The
// final division never overflows.
func TestProp_CosineScore_NoInf(t *testing.T) {
	f := func(a, b string, n int) bool {
		return !math.IsInf(fuzzymatch.CosineScore(a, b, cosineN(n)), 0)
	}
	if err := quick.Check(f, nil); err != nil {
		t.Errorf("CosineScore produced Inf: %v", err)
	}
}

// TestProp_CosineScore_NoNegativeZero asserts that when the byte-path
// score is 0.0 it is positive zero, not negative zero. The dot-product
// reduction sums non-negative integer-derived float64 products; the
// final division `0.0 / (positive)` is +0.0 in IEEE-754.
func TestProp_CosineScore_NoNegativeZero(t *testing.T) {
	f := func(a, b string, n int) bool {
		s := fuzzymatch.CosineScore(a, b, cosineN(n))
		return s != 0.0 || !math.Signbit(s)
	}
	if err := quick.Check(f, nil); err != nil {
		t.Errorf("CosineScore produced -0.0: %v", err)
	}
}

// TestProp_CosineScoreRunes_RangeBounds: rune-path mirror of
// _RangeBounds.
func TestProp_CosineScoreRunes_RangeBounds(t *testing.T) {
	f := func(a, b string, n int) bool {
		s := fuzzymatch.CosineScoreRunes(a, b, cosineN(n))
		return s >= 0.0 && s <= 1.0 && !math.IsNaN(s) && !math.IsInf(s, 0)
	}
	if err := quick.Check(f, nil); err != nil {
		t.Errorf("CosineScoreRunes out of [0,1] or non-finite: %v", err)
	}
}

// TestProp_CosineScoreRunes_Identity: rune-path identity.
func TestProp_CosineScoreRunes_Identity(t *testing.T) {
	f := func(x string, n int) bool {
		if x == "" {
			return true
		}
		return fuzzymatch.CosineScoreRunes(x, x, cosineN(n)) == 1.0
	}
	if err := quick.Check(f, nil); err != nil {
		t.Errorf("CosineScoreRunes identity violated: %v", err)
	}
}

// TestProp_CosineScoreRunes_Symmetric: rune-path symmetry.
func TestProp_CosineScoreRunes_Symmetric(t *testing.T) {
	f := func(a, b string, n int) bool {
		nn := cosineN(n)
		return fuzzymatch.CosineScoreRunes(a, b, nn) == fuzzymatch.CosineScoreRunes(b, a, nn)
	}
	if err := quick.Check(f, nil); err != nil {
		t.Errorf("CosineScoreRunes not symmetric: %v", err)
	}
}

// TestProp_CosineScoreRunes_NoNaN: rune-path NaN guard.
func TestProp_CosineScoreRunes_NoNaN(t *testing.T) {
	f := func(a, b string, n int) bool {
		return !math.IsNaN(fuzzymatch.CosineScoreRunes(a, b, cosineN(n)))
	}
	if err := quick.Check(f, nil); err != nil {
		t.Errorf("CosineScoreRunes produced NaN: %v", err)
	}
}

// TestProp_CosineScoreRunes_NoInf: rune-path Inf guard.
func TestProp_CosineScoreRunes_NoInf(t *testing.T) {
	f := func(a, b string, n int) bool {
		return !math.IsInf(fuzzymatch.CosineScoreRunes(a, b, cosineN(n)), 0)
	}
	if err := quick.Check(f, nil); err != nil {
		t.Errorf("CosineScoreRunes produced Inf: %v", err)
	}
}

// TestProp_CosineScoreRunes_NoNegativeZero: rune-path -0.0 guard.
func TestProp_CosineScoreRunes_NoNegativeZero(t *testing.T) {
	f := func(a, b string, n int) bool {
		s := fuzzymatch.CosineScoreRunes(a, b, cosineN(n))
		return s != 0.0 || !math.Signbit(s)
	}
	if err := quick.Check(f, nil); err != nil {
		t.Errorf("CosineScoreRunes produced -0.0: %v", err)
	}
}

// TestProp_CosineScore_DeterministicAcrossRuns asserts that 1000
// sequential calls on the same (a, b, n) input produce byte-identical
// output. Cosine-specific load-bearing test: this catches any future
// regression that re-introduces map-iteration order dependence on the
// dot-product reduction (CONTEXT.md §3 LOCKED — sort.Strings on the
// intersection key slice is the determinism gate).
//
// Compares via math.Float64bits to detect bit-level differences (e.g.
// +0.0 vs -0.0, signalling vs quiet NaN — even though the algorithm
// emits neither). Uses RV-C2 (5-key intersection) on the byte surface
// — the same pair as TestCosine_SortedKeyIteration in cosine_test.go.
// The rune surface gate uses RV-C3 (café/cafe/n=2) for cross-surface
// determinism coverage.
func TestProp_CosineScore_DeterministicAcrossRuns(t *testing.T) {
	const iterations = 1000
	const a = "abcdefgh"
	const b = "abcdefgi"
	const n = 3
	baseline := fuzzymatch.CosineScore(a, b, n)
	baselineBits := math.Float64bits(baseline)
	for i := 0; i < iterations; i++ {
		got := fuzzymatch.CosineScore(a, b, n)
		if math.Float64bits(got) != baselineBits {
			t.Fatalf("iteration %d: CosineScore(%q,%q,%d) = %.17g (bits=%x); baseline = %.17g (bits=%x)",
				i, a, b, n, got, math.Float64bits(got), baseline, baselineBits)
		}
	}
	// Mirror gate on the rune surface.
	baselineR := fuzzymatch.CosineScoreRunes("café", "cafe", 2)
	baselineRBits := math.Float64bits(baselineR)
	for i := 0; i < iterations; i++ {
		got := fuzzymatch.CosineScoreRunes("café", "cafe", 2)
		if math.Float64bits(got) != baselineRBits {
			t.Fatalf("iteration %d (rune): got %.17g (bits=%x); baseline = %.17g (bits=%x)",
				i, got, math.Float64bits(got), baselineR, baselineRBits)
		}
	}
}

// ---------------------------------------------------------------------------
// Tversky property tests (plan 05-04)
// ---------------------------------------------------------------------------
//
// Tversky carries the largest property-test surface in the q-gram tier
// because the (α, β) parameter pair adds two new failure modes beyond
// the standard six invariants:
//
//   - Asymmetry-conditional:   when α ≠ β AND |A−B| ≠ |B−A|, the
//     scores must differ on input swap. Plain symmetry is FALSE for
//     Tversky in general, so the standard _Symmetric property test
//     does NOT apply — instead we have _SymmetricWhenAlphaEqBeta
//     (asserts symmetry only when α = β) AND
//     _AsymmetricWhenAlphaNeqBeta (asserts asymmetry only when the
//     residuals also differ).
//
//   - Parameter-swap symmetry: T(a, b, α, β) = T(b, a, β, α) ALWAYS.
//     This is the algebraic identity that pins the asymmetry as a
//     consequence of α ≠ β rather than a one-sided coding error. If
//     the implementation silently swapped α and β internally, this
//     property would still hold (vacuously); the asymmetry-conditional
//     property is what catches that bug.
//
//   - Bit-exact algebraic cross-checks:
//       * T(a, b, n, 1.0, 1.0) == QGramJaccardScore(a, b, n)
//       * T(a, b, n, 0.5, 0.5) == SorensenDiceScore(a, b, n)
//     These hold bit-for-bit on the same multisets — pinned via
//     math.Float64bits comparison rather than tolerance.

// tverskyN coerces an arbitrary int into the [1, 5] inclusive range
// used by the Tversky property tests. Negative and zero n values
// are mapped into the valid range; the n parameter would otherwise
// panic per CONTEXT.md §5 LOCKED, but that contract is unit-tested
// separately by TestTversky_PanicsOnInvalidN — the property tests
// exercise the [0, 1] score-range invariants.
func tverskyN(n int) int {
	if n < 0 {
		n = -n
	}
	return (n % 5) + 1
}

// tverskyAlpha coerces an arbitrary float64 into the [0, 1] inclusive
// range. The squashing function `|x| / (|x| + 1)` maps R → [0, 1)
// monotonically; we then clamp to [0, 1]. NaN inputs are mapped to
// 0.5 (a safe interior point); ±Inf inputs squash to 1.0. Used jointly
// with tverskyBetaWithMin to ensure α + β > 0 in the property bodies
// (avoiding the documented panic path).
func tverskyAlpha(a float64) float64 {
	if math.IsNaN(a) {
		return 0.5
	}
	if math.IsInf(a, 0) {
		return 1.0
	}
	a = math.Abs(a)
	return a / (a + 1.0)
}

// tverskyBetaWithMin coerces an arbitrary float64 into [0, 1] using
// the same squashing function, but if the resulting (α, β) pair would
// have α + β == 0, β is forced to 1.0 to satisfy the α + β > 0
// invariant the public-API gate enforces.
func tverskyBetaWithMin(b float64, alpha float64) float64 {
	bb := tverskyAlpha(b)
	if alpha == 0.0 && bb == 0.0 {
		return 1.0
	}
	return bb
}

// TestProp_TverskyScore_RangeBounds asserts the byte-path score stays
// in [0.0, 1.0] for any (a, b, n, α, β) — with α and β coerced into
// the valid range. Inline NaN/Inf guards document the joint invariant;
// dedicated _NoNaN / _NoInf tests below retest each guard in isolation.
func TestProp_TverskyScore_RangeBounds(t *testing.T) {
	f := func(a, b string, n int, alpha, beta float64) bool {
		al := tverskyAlpha(alpha)
		be := tverskyBetaWithMin(beta, al)
		s := fuzzymatch.TverskyScore(a, b, tverskyN(n), al, be)
		return s >= 0.0 && s <= 1.0 && !math.IsNaN(s) && !math.IsInf(s, 0)
	}
	if err := quick.Check(f, nil); err != nil {
		t.Errorf("TverskyScore out of [0,1] or non-finite: %v", err)
	}
}

// TestProp_TverskyScore_Identity asserts Score(x, x, n, α, β) == 1.0
// EXACTLY for any non-empty x, any n >= 1, and any valid (α, β) — the
// identity short-circuit fires before extraction and the result is the
// literal 1.0 (regardless of α, β).
func TestProp_TverskyScore_Identity(t *testing.T) {
	f := func(x string, n int, alpha, beta float64) bool {
		if x == "" {
			return true // both-empty handled by unit tests
		}
		al := tverskyAlpha(alpha)
		be := tverskyBetaWithMin(beta, al)
		return fuzzymatch.TverskyScore(x, x, tverskyN(n), al, be) == 1.0
	}
	if err := quick.Check(f, nil); err != nil {
		t.Errorf("TverskyScore identity violated: %v", err)
	}
}

// TestProp_TverskyScore_NoNaN asserts the byte-path score never
// returns NaN. The both-empty + one-empty + identity short-circuits
// gate away the only potential 0/0 paths; the explicit
// tverskyFromQGramMaps len-check provides the secondary guard.
func TestProp_TverskyScore_NoNaN(t *testing.T) {
	f := func(a, b string, n int, alpha, beta float64) bool {
		al := tverskyAlpha(alpha)
		be := tverskyBetaWithMin(beta, al)
		return !math.IsNaN(fuzzymatch.TverskyScore(a, b, tverskyN(n), al, be))
	}
	if err := quick.Check(f, nil); err != nil {
		t.Errorf("TverskyScore produced NaN: %v", err)
	}
}

// TestProp_TverskyScore_NoInf asserts the byte-path score never
// returns ±Inf. Numerator and denominator are bounded — counts fit in
// float64 (≤ 2^53), α and β are coerced into [0, 1] — so the single
// multiplication + addition + division never overflows.
func TestProp_TverskyScore_NoInf(t *testing.T) {
	f := func(a, b string, n int, alpha, beta float64) bool {
		al := tverskyAlpha(alpha)
		be := tverskyBetaWithMin(beta, al)
		return !math.IsInf(fuzzymatch.TverskyScore(a, b, tverskyN(n), al, be), 0)
	}
	if err := quick.Check(f, nil); err != nil {
		t.Errorf("TverskyScore produced Inf: %v", err)
	}
}

// TestProp_TverskyScore_NoNegativeZero asserts that when the byte-path
// score is 0.0 it is positive zero, not negative zero. The intersection
// cardinality is a non-negative integer; float64(0) / float64(positive)
// is +0.0 in IEEE-754.
func TestProp_TverskyScore_NoNegativeZero(t *testing.T) {
	f := func(a, b string, n int, alpha, beta float64) bool {
		al := tverskyAlpha(alpha)
		be := tverskyBetaWithMin(beta, al)
		s := fuzzymatch.TverskyScore(a, b, tverskyN(n), al, be)
		return s != 0.0 || !math.Signbit(s)
	}
	if err := quick.Check(f, nil); err != nil {
		t.Errorf("TverskyScore produced -0.0: %v", err)
	}
}

// TestProp_TverskyScoreRunes_RangeBounds: rune-path mirror of
// _RangeBounds.
func TestProp_TverskyScoreRunes_RangeBounds(t *testing.T) {
	f := func(a, b string, n int, alpha, beta float64) bool {
		al := tverskyAlpha(alpha)
		be := tverskyBetaWithMin(beta, al)
		s := fuzzymatch.TverskyScoreRunes(a, b, tverskyN(n), al, be)
		return s >= 0.0 && s <= 1.0 && !math.IsNaN(s) && !math.IsInf(s, 0)
	}
	if err := quick.Check(f, nil); err != nil {
		t.Errorf("TverskyScoreRunes out of [0,1] or non-finite: %v", err)
	}
}

// TestProp_TverskyScoreRunes_Identity: rune-path identity.
func TestProp_TverskyScoreRunes_Identity(t *testing.T) {
	f := func(x string, n int, alpha, beta float64) bool {
		if x == "" {
			return true
		}
		al := tverskyAlpha(alpha)
		be := tverskyBetaWithMin(beta, al)
		return fuzzymatch.TverskyScoreRunes(x, x, tverskyN(n), al, be) == 1.0
	}
	if err := quick.Check(f, nil); err != nil {
		t.Errorf("TverskyScoreRunes identity violated: %v", err)
	}
}

// TestProp_TverskyScoreRunes_NoNaN: rune-path NaN guard.
func TestProp_TverskyScoreRunes_NoNaN(t *testing.T) {
	f := func(a, b string, n int, alpha, beta float64) bool {
		al := tverskyAlpha(alpha)
		be := tverskyBetaWithMin(beta, al)
		return !math.IsNaN(fuzzymatch.TverskyScoreRunes(a, b, tverskyN(n), al, be))
	}
	if err := quick.Check(f, nil); err != nil {
		t.Errorf("TverskyScoreRunes produced NaN: %v", err)
	}
}

// TestProp_TverskyScoreRunes_NoInf: rune-path Inf guard.
func TestProp_TverskyScoreRunes_NoInf(t *testing.T) {
	f := func(a, b string, n int, alpha, beta float64) bool {
		al := tverskyAlpha(alpha)
		be := tverskyBetaWithMin(beta, al)
		return !math.IsInf(fuzzymatch.TverskyScoreRunes(a, b, tverskyN(n), al, be), 0)
	}
	if err := quick.Check(f, nil); err != nil {
		t.Errorf("TverskyScoreRunes produced Inf: %v", err)
	}
}

// TestProp_TverskyScoreRunes_NoNegativeZero: rune-path -0.0 guard.
func TestProp_TverskyScoreRunes_NoNegativeZero(t *testing.T) {
	f := func(a, b string, n int, alpha, beta float64) bool {
		al := tverskyAlpha(alpha)
		be := tverskyBetaWithMin(beta, al)
		s := fuzzymatch.TverskyScoreRunes(a, b, tverskyN(n), al, be)
		return s != 0.0 || !math.Signbit(s)
	}
	if err := quick.Check(f, nil); err != nil {
		t.Errorf("TverskyScoreRunes produced -0.0: %v", err)
	}
}

// TestProp_TverskyScore_SymmetricWhenAlphaEqBeta asserts T(a, b, n, α,
// α) == T(b, a, n, α, α) bit-for-bit. When α = β the function is
// symmetric in (a, b) — this is the corollary of the parameter-swap
// symmetry algebraic identity. Quick.Check over arbitrary (a, b, n, α);
// β is bound to α inside the closure.
func TestProp_TverskyScore_SymmetricWhenAlphaEqBeta(t *testing.T) {
	f := func(a, b string, n int, alpha float64) bool {
		al := tverskyAlpha(alpha)
		if al == 0.0 {
			al = 0.5 // satisfy α + β > 0 with both at 0.5
		}
		nn := tverskyN(n)
		fwd := fuzzymatch.TverskyScore(a, b, nn, al, al)
		rev := fuzzymatch.TverskyScore(b, a, nn, al, al)
		return math.Float64bits(fwd) == math.Float64bits(rev)
	}
	if err := quick.Check(f, nil); err != nil {
		t.Errorf("TverskyScore α=β symmetry violated: %v", err)
	}
}

// TestProp_TverskyScoreRunes_SymmetricWhenAlphaEqBeta: rune-path
// mirror of _SymmetricWhenAlphaEqBeta.
func TestProp_TverskyScoreRunes_SymmetricWhenAlphaEqBeta(t *testing.T) {
	f := func(a, b string, n int, alpha float64) bool {
		al := tverskyAlpha(alpha)
		if al == 0.0 {
			al = 0.5
		}
		nn := tverskyN(n)
		fwd := fuzzymatch.TverskyScoreRunes(a, b, nn, al, al)
		rev := fuzzymatch.TverskyScoreRunes(b, a, nn, al, al)
		return math.Float64bits(fwd) == math.Float64bits(rev)
	}
	if err := quick.Check(f, nil); err != nil {
		t.Errorf("TverskyScoreRunes α=β symmetry violated: %v", err)
	}
}

// TestProp_TverskyScore_AsymmetricWhenAlphaNeqBeta is the asymmetry-
// conditional property test. With fixed α=0.8, β=0.2 (the LOAD-BEARING
// asymmetric configuration), the implication is:
//
//	IF the multiset residuals differ (|A−B| ≠ |B−A|) on the q-gram
//	extraction THEN swapping inputs MUST produce a different score
//	(T(a, b, ...) ≠ T(b, a, ...)).
//
// Detecting whether residuals differ from inside the property body
// without re-implementing extraction: the multiset total cardinality
// equals len(s)−n+1 for any non-empty s with len(s) ≥ n. The residual
// imbalance |A−B| vs |B−A| is asymmetric iff totalA ≠ totalB iff
// len(a) ≠ len(b) (when both ≥ n; when one is shorter than n, the
// extraction returns an empty multiset and the symmetric branch fires).
//
// This gives a clean implication: when len(a) ≠ len(b) AND both ≥ n
// AND the multisets have at least one shared key (otherwise both
// scores collapse to 0/(α·totalA + β·totalB) which is symmetric in
// (a, b) only when α=β — which is NOT our case here, so even orthogonal
// pairs exercise the asymmetry).
//
// Edge case: when intersection = 0, T = 0 / (α·totalA + β·totalB) = 0
// for both directions; this is structurally symmetric. So the
// implication captures only the partial-overlap, length-mismatched
// case.
func TestProp_TverskyScore_AsymmetricWhenAlphaNeqBeta(t *testing.T) {
	const alpha = 0.8
	const beta = 0.2
	f := func(a, b string, n int) bool {
		nn := tverskyN(n)
		// Skip identity / one-empty (short-circuit branches).
		if a == b || a == "" || b == "" {
			return true
		}
		fwd := fuzzymatch.TverskyScore(a, b, nn, alpha, beta)
		rev := fuzzymatch.TverskyScore(b, a, nn, alpha, beta)
		// Premise: residuals must differ. Approximate: len(a) != len(b)
		// AND both at least n bytes long (so extraction is non-empty)
		// AND fwd > 0 (so there's some shared multiset content; if
		// fwd == 0 it's an orthogonal pair and rev is also 0 — the
		// 0/(α·tA + β·tB) = 0 collapse).
		if len(a) == len(b) || len(a) < nn || len(b) < nn || fwd == 0.0 {
			// Premise false → vacuous truth.
			return true
		}
		// Premise holds (residuals likely differ AND there's overlap).
		// Conclusion: scores must differ.
		return fwd != rev
	}
	if err := quick.Check(f, nil); err != nil {
		t.Errorf("TverskyScore asymmetry-conditional property violated: %v", err)
	}
	// Spot-check on the canonical RV-T1/RV-T2 pair to confirm the
	// property body actively detects asymmetric inputs (a degenerate
	// `return true` would otherwise slip past quick.Check).
	rvT1 := fuzzymatch.TverskyScore("abcd", "abcdef", 2, alpha, beta)
	rvT2 := fuzzymatch.TverskyScore("abcdef", "abcd", 2, alpha, beta)
	if rvT1 == rvT2 {
		t.Errorf("asymmetry spot-check failed: T(abcd,abcdef,2,%g,%g)=%.17g equals T(abcdef,abcd,...)=%.17g — RV-T1/RV-T2 should differ",
			alpha, beta, rvT1, rvT2)
	}
}

// TestProp_TverskyScoreRunes_AsymmetricWhenAlphaNeqBeta: rune-path
// mirror of _AsymmetricWhenAlphaNeqBeta. The premise uses rune count
// (utf8.RuneCountInString equivalent via len([]rune(...))) instead of
// byte length, since the rune extractor's multiset total is
// runeCount−n+1.
func TestProp_TverskyScoreRunes_AsymmetricWhenAlphaNeqBeta(t *testing.T) {
	const alpha = 0.8
	const beta = 0.2
	f := func(a, b string, n int) bool {
		nn := tverskyN(n)
		if a == b || a == "" || b == "" {
			return true
		}
		fwd := fuzzymatch.TverskyScoreRunes(a, b, nn, alpha, beta)
		rev := fuzzymatch.TverskyScoreRunes(b, a, nn, alpha, beta)
		// Rune-count premise. []rune(s) decodes UTF-8; invalid bytes
		// become U+FFFD. The premise is "rune length differs AND both
		// at least n runes long AND scores show overlap" (mirroring
		// the byte-path premise).
		ra := len([]rune(a))
		rb := len([]rune(b))
		if ra == rb || ra < nn || rb < nn || fwd == 0.0 {
			return true
		}
		return fwd != rev
	}
	if err := quick.Check(f, nil); err != nil {
		t.Errorf("TverskyScoreRunes asymmetry-conditional property violated: %v", err)
	}
	// Spot-check via the rune surface on the same RV-T1/RV-T2 pair
	// (ASCII inputs; rune and byte paths produce equivalent multisets).
	rvT1 := fuzzymatch.TverskyScoreRunes("abcd", "abcdef", 2, alpha, beta)
	rvT2 := fuzzymatch.TverskyScoreRunes("abcdef", "abcd", 2, alpha, beta)
	if rvT1 == rvT2 {
		t.Errorf("rune asymmetry spot-check failed: T(abcd,abcdef)=%.17g equals T(abcdef,abcd)=%.17g", rvT1, rvT2)
	}
}

// TestProp_TverskyScore_ParameterSwapSymmetry pins the Tversky 1977 §2
// algebraic identity T(a, b, n, α, β) = T(b, a, n, β, α) bit-for-bit
// for ANY valid (a, b, n, α, β). This is the load-bearing property
// that proves the asymmetry is a consequence of α ≠ β rather than a
// one-sided coding error; if the implementation silently swapped α and
// β internally, this property would still hold (vacuously true for
// the swapped state), but combined with the asymmetry-discriminating
// unit test it forms the parameter-order regression detector.
func TestProp_TverskyScore_ParameterSwapSymmetry(t *testing.T) {
	f := func(a, b string, n int, alpha, beta float64) bool {
		nn := tverskyN(n)
		al := tverskyAlpha(alpha)
		be := tverskyBetaWithMin(beta, al)
		fwd := fuzzymatch.TverskyScore(a, b, nn, al, be)
		rev := fuzzymatch.TverskyScore(b, a, nn, be, al)
		return math.Float64bits(fwd) == math.Float64bits(rev)
	}
	if err := quick.Check(f, nil); err != nil {
		t.Errorf("TverskyScore parameter-swap symmetry violated: %v", err)
	}
}

// TestProp_TverskyScoreRunes_ParameterSwapSymmetry: rune-path mirror.
func TestProp_TverskyScoreRunes_ParameterSwapSymmetry(t *testing.T) {
	f := func(a, b string, n int, alpha, beta float64) bool {
		nn := tverskyN(n)
		al := tverskyAlpha(alpha)
		be := tverskyBetaWithMin(beta, al)
		fwd := fuzzymatch.TverskyScoreRunes(a, b, nn, al, be)
		rev := fuzzymatch.TverskyScoreRunes(b, a, nn, be, al)
		return math.Float64bits(fwd) == math.Float64bits(rev)
	}
	if err := quick.Check(f, nil); err != nil {
		t.Errorf("TverskyScoreRunes parameter-swap symmetry violated: %v", err)
	}
}

// TestProp_TverskyScore_JaccardCrossCheck asserts the algebraic
// identity T(a, b, n, 1.0, 1.0) == QGramJaccardScore(a, b, n)
// bit-for-bit (via math.Float64bits comparison) for any (a, b, n).
// Tversky 1977 §2 with α=β=1 reduces to the Jaccard coefficient on
// multisets — pinned at the property level (in addition to the unit-
// test TestTversky_JaccardCrossCheck) so a regression on either side
// of the equivalence surfaces immediately.
func TestProp_TverskyScore_JaccardCrossCheck(t *testing.T) {
	f := func(a, b string, n int) bool {
		nn := tverskyN(n)
		tv := fuzzymatch.TverskyScore(a, b, nn, 1.0, 1.0)
		jc := fuzzymatch.QGramJaccardScore(a, b, nn)
		return math.Float64bits(tv) == math.Float64bits(jc)
	}
	if err := quick.Check(f, nil); err != nil {
		t.Errorf("TverskyScore(α=β=1) ≠ QGramJaccardScore: %v", err)
	}
}

// TestProp_TverskyScoreRunes_JaccardCrossCheck: rune-path mirror.
func TestProp_TverskyScoreRunes_JaccardCrossCheck(t *testing.T) {
	f := func(a, b string, n int) bool {
		nn := tverskyN(n)
		tv := fuzzymatch.TverskyScoreRunes(a, b, nn, 1.0, 1.0)
		jc := fuzzymatch.QGramJaccardScoreRunes(a, b, nn)
		return math.Float64bits(tv) == math.Float64bits(jc)
	}
	if err := quick.Check(f, nil); err != nil {
		t.Errorf("TverskyScoreRunes(α=β=1) ≠ QGramJaccardScoreRunes: %v", err)
	}
}

// TestProp_TverskyScore_DiceCrossCheck asserts the algebraic identity
// T(a, b, n, 0.5, 0.5) == SorensenDiceScore(a, b, n) bit-for-bit. With
// α=β=0.5 the Tversky denominator collapses to (totalA + totalB)/2,
// matching the Dice coefficient exactly.
func TestProp_TverskyScore_DiceCrossCheck(t *testing.T) {
	f := func(a, b string, n int) bool {
		nn := tverskyN(n)
		tv := fuzzymatch.TverskyScore(a, b, nn, 0.5, 0.5)
		dc := fuzzymatch.SorensenDiceScore(a, b, nn)
		return math.Float64bits(tv) == math.Float64bits(dc)
	}
	if err := quick.Check(f, nil); err != nil {
		t.Errorf("TverskyScore(α=β=0.5) ≠ SorensenDiceScore: %v", err)
	}
}

// TestProp_TverskyScoreRunes_DiceCrossCheck: rune-path mirror.
func TestProp_TverskyScoreRunes_DiceCrossCheck(t *testing.T) {
	f := func(a, b string, n int) bool {
		nn := tverskyN(n)
		tv := fuzzymatch.TverskyScoreRunes(a, b, nn, 0.5, 0.5)
		dc := fuzzymatch.SorensenDiceScoreRunes(a, b, nn)
		return math.Float64bits(tv) == math.Float64bits(dc)
	}
	if err := quick.Check(f, nil); err != nil {
		t.Errorf("TverskyScoreRunes(α=β=0.5) ≠ SorensenDiceScoreRunes: %v", err)
	}
}

// TestProp_TverskyScore_DeterministicAcrossRuns asserts that 1000
// sequential calls on the same (a, b, n, α, β) input produce
// byte-identical output. PITFALLS §14 closure carried forward from
// plans 05-01/05-02/05-03 — guards against any future regression that
// might re-introduce map-iteration order dependence on the output
// path.
//
// Compares via math.Float64bits to detect bit-level differences (e.g.
// +0.0 vs -0.0, signalling vs quiet NaN — even though the algorithm
// emits neither). Uses the load-bearing RV-T1 input pair on the byte
// surface, mirrored on the café/cafe rune surface.
func TestProp_TverskyScore_DeterministicAcrossRuns(t *testing.T) {
	const iterations = 1000
	const a = "abcd"
	const b = "abcdef"
	const n = 2
	const alpha = 0.8
	const beta = 0.2
	baseline := fuzzymatch.TverskyScore(a, b, n, alpha, beta)
	baselineBits := math.Float64bits(baseline)
	for i := 0; i < iterations; i++ {
		got := fuzzymatch.TverskyScore(a, b, n, alpha, beta)
		if math.Float64bits(got) != baselineBits {
			t.Fatalf("iteration %d: TverskyScore(%q,%q,%d,%g,%g) = %.17g (bits=%x); baseline = %.17g (bits=%x)",
				i, a, b, n, alpha, beta, got, math.Float64bits(got), baseline, baselineBits)
		}
	}
	// Mirror gate on the rune surface.
	baselineR := fuzzymatch.TverskyScoreRunes("café", "cafe", 2, 0.5, 0.5)
	baselineRBits := math.Float64bits(baselineR)
	for i := 0; i < iterations; i++ {
		got := fuzzymatch.TverskyScoreRunes("café", "cafe", 2, 0.5, 0.5)
		if math.Float64bits(got) != baselineRBits {
			t.Fatalf("iteration %d (rune): got %.17g (bits=%x); baseline = %.17g (bits=%x)",
				i, got, math.Float64bits(got), baselineR, baselineRBits)
		}
	}
}

// TestRatcliffObershelpScore_AtLeastLevenshtein_OnSubstringContainment checks the
// "generally" property from RESEARCH.md: on substring-containment inputs
// the Ratcliff-Obershelp score is typically ≥ the Levenshtein score
// because RO finds the contiguous match and ignores the deletion cost,
// while Levenshtein pays for every dropped character.
//
// The symbol name deliberately omits the `TestProp_` prefix used elsewhere
// in this file: this is a table-driven test over hand-curated inputs, not a
// testing/quick property over all strings (review IN-05). RESEARCH.md
// explicitly notes the property is "generally" true, not universal.
// Quick.Check would produce degenerate inputs (e.g. random non-overlapping
// strings) where the property either trivially holds or accidentally fails
// for reasons unrelated to the algorithm. Hand-curated substring-containment
// cases directly exercise the intended use case.
func TestRatcliffObershelpScore_AtLeastLevenshtein_OnSubstringContainment(t *testing.T) {
	tests := []struct {
		a, b string
	}{
		{"http_request", "http_request_header_fields"},
		{"abc", "xyzabcdef"},
		{"kitten", "the_kitten_purrs"},
		{"WIKIMEDIA", "WIKIMANIA"},
		{"abcdef", "xyzabcdefuvw"},
		{"user_id", "user_id_v2"},
	}
	for _, tt := range tests {
		t.Run(tt.a+"_"+tt.b, func(t *testing.T) {
			ro := fuzzymatch.RatcliffObershelpScore(tt.a, tt.b)
			lev := fuzzymatch.LevenshteinScore(tt.a, tt.b)
			if ro < lev {
				t.Errorf("RatcliffObershelpScore(%q,%q) = %g < LevenshteinScore = %g (RO should be >= Lev on substring-containment inputs)",
					tt.a, tt.b, ro, lev)
			}
		})
	}
}

// ---------------------------------------------------------------------------
// TokenSortRatio property tests (plan 06-01)
// ---------------------------------------------------------------------------

// TestProp_TokenSortRatioScore_RangeBounds asserts the score stays in
// [0.0, 1.0] for any (a, b) pair. Joint NaN/Inf gate documents the
// composite invariant; dedicated _NoNaN / _NoInf tests below retest
// each guard in isolation.
func TestProp_TokenSortRatioScore_RangeBounds(t *testing.T) {
	f := func(a, b string) bool {
		s := fuzzymatch.TokenSortRatioScore(a, b)
		return s >= 0.0 && s <= 1.0 && !math.IsNaN(s) && !math.IsInf(s, 0)
	}
	if err := quick.Check(f, nil); err != nil {
		t.Errorf("TokenSortRatioScore out of [0,1] or non-finite: %v", err)
	}
}

// TestProp_TokenSortRatioScore_Identity asserts Score(x, x) == 1.0
// EXACTLY for any string x — the identity short-circuit fires before
// Tokenise and the result is the literal 1.0. This includes all
// strings: empty, all-separator, mixed UTF-8, identifier-style. The
// property is stronger than the Q-Gram Jaccard identity (which skips
// the empty case) because TokenSortRatio's short-circuit covers
// every input.
func TestProp_TokenSortRatioScore_Identity(t *testing.T) {
	f := func(x string) bool {
		return fuzzymatch.TokenSortRatioScore(x, x) == 1.0
	}
	if err := quick.Check(f, nil); err != nil {
		t.Errorf("TokenSortRatioScore identity violated: %v", err)
	}
}

// TestProp_TokenSortRatioScore_Symmetric asserts Score(a, b) ==
// Score(b, a) EXACTLY (bit-for-bit). Tokenise is deterministic;
// sort.Strings is stable byte-lex; strings.Join is order-preserving;
// indelRatio is symmetric — every composition step preserves
// symmetry, so the IEEE-754 division produces identical output
// regardless of argument order.
func TestProp_TokenSortRatioScore_Symmetric(t *testing.T) {
	f := func(a, b string) bool {
		return fuzzymatch.TokenSortRatioScore(a, b) == fuzzymatch.TokenSortRatioScore(b, a)
	}
	if err := quick.Check(f, nil); err != nil {
		t.Errorf("TokenSortRatioScore not symmetric: %v", err)
	}
}

// TestProp_TokenSortRatioScore_NoNaN asserts the score never returns
// NaN. The identity / both-empty / one-empty guards gate away the only
// potential 0/0 paths; the indelRatio sum-check provides the
// secondary guard.
func TestProp_TokenSortRatioScore_NoNaN(t *testing.T) {
	f := func(a, b string) bool {
		return !math.IsNaN(fuzzymatch.TokenSortRatioScore(a, b))
	}
	if err := quick.Check(f, nil); err != nil {
		t.Errorf("TokenSortRatioScore produced NaN: %v", err)
	}
}

// TestProp_TokenSortRatioScore_NoInf asserts the score never returns
// ±Inf. Numerator and denominator are bounded integers fitting in
// float64 (counts up to 2^53 are exact for typical input sizes); the
// single division never overflows.
func TestProp_TokenSortRatioScore_NoInf(t *testing.T) {
	f := func(a, b string) bool {
		return !math.IsInf(fuzzymatch.TokenSortRatioScore(a, b), 0)
	}
	if err := quick.Check(f, nil); err != nil {
		t.Errorf("TokenSortRatioScore produced Inf: %v", err)
	}
}

// TestProp_TokenSortRatioScore_NoNegativeZero asserts that when the
// score is 0.0 it is positive zero, not negative zero. The
// numerator (2 · lcsLen) is a non-negative integer; float64(0) /
// float64(positive) is +0.0 in IEEE-754.
func TestProp_TokenSortRatioScore_NoNegativeZero(t *testing.T) {
	f := func(a, b string) bool {
		s := fuzzymatch.TokenSortRatioScore(a, b)
		return s != 0.0 || !math.Signbit(s)
	}
	if err := quick.Check(f, nil); err != nil {
		t.Errorf("TokenSortRatioScore produced -0.0: %v", err)
	}
}

// ---------------------------------------------------------------------------
// TokenSetRatio property tests (plan 06-02)
// ---------------------------------------------------------------------------

// TestProp_TokenSetRatioScore_RangeBounds asserts the score stays in
// [0.0, 1.0] for any (a, b) pair. Joint NaN/Inf gate documents the
// composite invariant; dedicated _NoNaN / _NoInf tests below retest
// each guard in isolation.
func TestProp_TokenSetRatioScore_RangeBounds(t *testing.T) {
	f := func(a, b string) bool {
		s := fuzzymatch.TokenSetRatioScore(a, b)
		return s >= 0.0 && s <= 1.0 && !math.IsNaN(s) && !math.IsInf(s, 0)
	}
	if err := quick.Check(f, nil); err != nil {
		t.Errorf("TokenSetRatioScore out of [0,1] or non-finite: %v", err)
	}
}

// TestProp_TokenSetRatioScore_Identity asserts Score(x, x) == 1.0 for
// any NON-EMPTY string x. The empty case is GUARDED: when x == "" the
// LOCKED RapidFuzz issue #110 DEVIATION fires first and returns 0.0,
// NOT 1.0 (the empty-input gate runs before the identity
// short-circuit per the LOCKED deviation). Per RESEARCH.md Pitfall 2,
// this is the documented exception from the catalogue's
// unconditional Score(x, x) == 1.0 property.
//
// Pure-separator strings (e.g. " ", "___") that Tokenise to an empty
// slice ALSO return 0.0 via the post-Tokenise empty-set gate. The
// guard `if x == "" return true` only skips the literal-empty case;
// the property is then false for pure-separator strings too. We
// therefore guard on `len(Tokenise(x, opts)) == 0` to skip all
// post-Tokenise-empty inputs — the deviation is documented elsewhere
// (algorithm godoc, BDD scenarios, staging-golden, unit tests, and
// this docstring).
func TestProp_TokenSetRatioScore_Identity(t *testing.T) {
	opts := fuzzymatch.DefaultTokeniseOptions()
	f := func(x string) bool {
		// Guard the LOCKED DEVIATION: when Tokenise(x) is empty
		// (literal-empty x or pure-separator x), the function
		// returns 0.0 not 1.0. The identity property is
		// vacuously true for those inputs per RESEARCH.md Pitfall
		// 2.
		if len(fuzzymatch.Tokenise(x, opts)) == 0 {
			return true
		}
		return fuzzymatch.TokenSetRatioScore(x, x) == 1.0
	}
	if err := quick.Check(f, nil); err != nil {
		t.Errorf("TokenSetRatioScore identity violated: %v", err)
	}
}

// TestProp_TokenSetRatioScore_Symmetric asserts Score(a, b) ==
// Score(b, a) EXACTLY (bit-for-bit). Tokenise is deterministic; set
// construction is order-independent; the three-way max operator is
// order-insensitive; r1 and r2 swap when (a, b) → (b, a) but
// max(r1, r2, r3) is invariant; r3 is symmetric in its argument order
// via indelRatio's own symmetry.
func TestProp_TokenSetRatioScore_Symmetric(t *testing.T) {
	f := func(a, b string) bool {
		return fuzzymatch.TokenSetRatioScore(a, b) == fuzzymatch.TokenSetRatioScore(b, a)
	}
	if err := quick.Check(f, nil); err != nil {
		t.Errorf("TokenSetRatioScore not symmetric: %v", err)
	}
}

// TestProp_TokenSetRatioScore_NoNaN asserts the score never returns
// NaN. The identity / empty-set / subset short-circuit guards gate
// away every 0/0 path; the indelRatio sum-check provides the secondary
// guard.
func TestProp_TokenSetRatioScore_NoNaN(t *testing.T) {
	f := func(a, b string) bool {
		return !math.IsNaN(fuzzymatch.TokenSetRatioScore(a, b))
	}
	if err := quick.Check(f, nil); err != nil {
		t.Errorf("TokenSetRatioScore produced NaN: %v", err)
	}
}

// TestProp_TokenSetRatioScore_NoInf asserts the score never returns
// ±Inf. Numerator and denominator are bounded integers fitting in
// float64; the three divisions never overflow.
func TestProp_TokenSetRatioScore_NoInf(t *testing.T) {
	f := func(a, b string) bool {
		return !math.IsInf(fuzzymatch.TokenSetRatioScore(a, b), 0)
	}
	if err := quick.Check(f, nil); err != nil {
		t.Errorf("TokenSetRatioScore produced Inf: %v", err)
	}
}

// TestProp_TokenSetRatioScore_NoNegativeZero asserts that when the
// score is 0.0 it is positive zero, not negative zero. The numerator
// (2 · lcsLen) is a non-negative integer; float64(0) / float64(positive)
// is +0.0 in IEEE-754. The empty-set DEVIATION returns the literal
// 0.0 (positive zero by Go-language guarantee).
func TestProp_TokenSetRatioScore_NoNegativeZero(t *testing.T) {
	f := func(a, b string) bool {
		s := fuzzymatch.TokenSetRatioScore(a, b)
		return s != 0.0 || !math.Signbit(s)
	}
	if err := quick.Check(f, nil); err != nil {
		t.Errorf("TokenSetRatioScore produced -0.0: %v", err)
	}
}

// ---------------------------------------------------------------------------
// PartialRatio property tests (plan 06-03) — byte path
// ---------------------------------------------------------------------------

// TestProp_PartialRatioScore_RangeBounds asserts the score stays in
// [0.0, 1.0] for any (a, b) pair. Joint NaN/Inf gate documents the
// composite invariant; dedicated _NoNaN / _NoInf tests below retest
// each guard in isolation.
func TestProp_PartialRatioScore_RangeBounds(t *testing.T) {
	f := func(a, b string) bool {
		s := fuzzymatch.PartialRatioScore(a, b)
		return s >= 0.0 && s <= 1.0 && !math.IsNaN(s) && !math.IsInf(s, 0)
	}
	if err := quick.Check(f, nil); err != nil {
		t.Errorf("PartialRatioScore out of [0,1] or non-finite: %v", err)
	}
}

// TestProp_PartialRatioScore_Identity asserts Score(x, x) == 1.0
// EXACTLY for any string x — the identity short-circuit fires before
// any byte slicing or charSet construction and the result is the
// literal 1.0. This includes all strings: empty, all-separator, mixed
// UTF-8 (the byte path treats each byte independently — invalid UTF-8
// at the byte level is still byte-identical at identity check).
//
// PartialRatio does NOT inherit TokenSetRatio's RapidFuzz issue #110
// deviation — both-empty returns 1.0 (caught by the identity gate)
// per the catalogue's standard convention.
func TestProp_PartialRatioScore_Identity(t *testing.T) {
	f := func(x string) bool {
		return fuzzymatch.PartialRatioScore(x, x) == 1.0
	}
	if err := quick.Check(f, nil); err != nil {
		t.Errorf("PartialRatioScore identity violated: %v", err)
	}
}

// TestProp_PartialRatioScore_Symmetric asserts Score(a, b) == Score(b, a)
// EXACTLY (bit-for-bit). The shorter-longer swap is internal to the
// algorithm; indelRatio is symmetric over its argument order; the
// three-region iteration is symmetric in the (shorter, longer) pair
// because regions 1 and 3 cover the two tails of the alignment and
// region 2 covers the middle.
func TestProp_PartialRatioScore_Symmetric(t *testing.T) {
	f := func(a, b string) bool {
		return fuzzymatch.PartialRatioScore(a, b) == fuzzymatch.PartialRatioScore(b, a)
	}
	if err := quick.Check(f, nil); err != nil {
		t.Errorf("PartialRatioScore not symmetric: %v", err)
	}
}

// TestProp_PartialRatioScore_NoNaN asserts the score never returns NaN.
// The identity / both-empty / one-empty guards gate away the only
// potential 0/0 paths; the indelRatio sum-check provides the
// secondary guard.
func TestProp_PartialRatioScore_NoNaN(t *testing.T) {
	f := func(a, b string) bool {
		return !math.IsNaN(fuzzymatch.PartialRatioScore(a, b))
	}
	if err := quick.Check(f, nil); err != nil {
		t.Errorf("PartialRatioScore produced NaN: %v", err)
	}
}

// TestProp_PartialRatioScore_NoInf asserts the score never returns ±Inf.
// Numerator and denominator are bounded integers fitting in float64;
// the single division in indelRatio never overflows.
func TestProp_PartialRatioScore_NoInf(t *testing.T) {
	f := func(a, b string) bool {
		return !math.IsInf(fuzzymatch.PartialRatioScore(a, b), 0)
	}
	if err := quick.Check(f, nil); err != nil {
		t.Errorf("PartialRatioScore produced Inf: %v", err)
	}
}

// TestProp_PartialRatioScore_NoNegativeZero asserts that when the
// score is 0.0 it is positive zero, not negative zero. The numerator
// (2 · lcsLen) is a non-negative integer; float64(0) / float64(positive)
// is +0.0 in IEEE-754. The disjoint case returns the literal 0.0
// (positive zero by Go-language guarantee).
func TestProp_PartialRatioScore_NoNegativeZero(t *testing.T) {
	f := func(a, b string) bool {
		s := fuzzymatch.PartialRatioScore(a, b)
		return s != 0.0 || !math.Signbit(s)
	}
	if err := quick.Check(f, nil); err != nil {
		t.Errorf("PartialRatioScore produced -0.0: %v", err)
	}
}

// ---------------------------------------------------------------------------
// PartialRatio property tests (plan 06-03) — rune path
// ---------------------------------------------------------------------------

// TestProp_PartialRatioScoreRunes_RangeBounds asserts the rune-path
// score stays in [0.0, 1.0] for any (a, b) pair.
func TestProp_PartialRatioScoreRunes_RangeBounds(t *testing.T) {
	f := func(a, b string) bool {
		s := fuzzymatch.PartialRatioScoreRunes(a, b)
		return s >= 0.0 && s <= 1.0 && !math.IsNaN(s) && !math.IsInf(s, 0)
	}
	if err := quick.Check(f, nil); err != nil {
		t.Errorf("PartialRatioScoreRunes out of [0,1] or non-finite: %v", err)
	}
}

// TestProp_PartialRatioScoreRunes_Identity asserts Score(x, x) == 1.0
// EXACTLY for any string x — the identity short-circuit fires BEFORE
// the `[]rune(x)` conversion (saves 2 heap allocations on identical
// inputs; same pattern as LongestCommonSubstringRunes).
func TestProp_PartialRatioScoreRunes_Identity(t *testing.T) {
	f := func(x string) bool {
		return fuzzymatch.PartialRatioScoreRunes(x, x) == 1.0
	}
	if err := quick.Check(f, nil); err != nil {
		t.Errorf("PartialRatioScoreRunes identity violated: %v", err)
	}
}

// TestProp_PartialRatioScoreRunes_Symmetric asserts the rune-path
// score is symmetric across argument order. Same reasoning as the
// byte-path symmetry property: the shorter-longer swap is internal
// and indelRatioRunes is symmetric.
func TestProp_PartialRatioScoreRunes_Symmetric(t *testing.T) {
	f := func(a, b string) bool {
		return fuzzymatch.PartialRatioScoreRunes(a, b) == fuzzymatch.PartialRatioScoreRunes(b, a)
	}
	if err := quick.Check(f, nil); err != nil {
		t.Errorf("PartialRatioScoreRunes not symmetric: %v", err)
	}
}

// TestProp_PartialRatioScoreRunes_NoNaN asserts the rune-path score
// never returns NaN.
func TestProp_PartialRatioScoreRunes_NoNaN(t *testing.T) {
	f := func(a, b string) bool {
		return !math.IsNaN(fuzzymatch.PartialRatioScoreRunes(a, b))
	}
	if err := quick.Check(f, nil); err != nil {
		t.Errorf("PartialRatioScoreRunes produced NaN: %v", err)
	}
}

// TestProp_PartialRatioScoreRunes_NoInf asserts the rune-path score
// never returns ±Inf.
func TestProp_PartialRatioScoreRunes_NoInf(t *testing.T) {
	f := func(a, b string) bool {
		return !math.IsInf(fuzzymatch.PartialRatioScoreRunes(a, b), 0)
	}
	if err := quick.Check(f, nil); err != nil {
		t.Errorf("PartialRatioScoreRunes produced Inf: %v", err)
	}
}

// TestProp_PartialRatioScoreRunes_NoNegativeZero asserts the rune-path
// score does not produce negative zero on disjoint inputs.
func TestProp_PartialRatioScoreRunes_NoNegativeZero(t *testing.T) {
	f := func(a, b string) bool {
		s := fuzzymatch.PartialRatioScoreRunes(a, b)
		return s != 0.0 || !math.Signbit(s)
	}
	if err := quick.Check(f, nil); err != nil {
		t.Errorf("PartialRatioScoreRunes produced -0.0: %v", err)
	}
}

// ---------------------------------------------------------------------------
// TokenJaccard property tests (plan 06-04)
// ---------------------------------------------------------------------------

// TestProp_TokenJaccardScore_RangeBounds asserts the score stays in
// [0.0, 1.0] for any (a, b) pair. Joint NaN/Inf gate documents the
// composite invariant; dedicated _NoNaN / _NoInf tests below retest
// each guard in isolation.
func TestProp_TokenJaccardScore_RangeBounds(t *testing.T) {
	f := func(a, b string) bool {
		s := fuzzymatch.TokenJaccardScore(a, b)
		return s >= 0.0 && s <= 1.0 && !math.IsNaN(s) && !math.IsInf(s, 0)
	}
	if err := quick.Check(f, nil); err != nil {
		t.Errorf("TokenJaccardScore out of [0,1] or non-finite: %v", err)
	}
}

// TestProp_TokenJaccardScore_Identity asserts Score(x, x) == 1.0
// EXACTLY for any string x — the identity short-circuit fires before
// Tokenise and the result is the literal 1.0. This is stronger than the
// Q-Gram Jaccard identity (which skips the empty case) because
// TokenJaccard's short-circuit covers every input INCLUDING the empty
// case (per the LOCKED both-empty STANDARD catalogue convention —
// TokenJaccard does NOT deviate like TokenSetRatio).
func TestProp_TokenJaccardScore_Identity(t *testing.T) {
	f := func(x string) bool {
		return fuzzymatch.TokenJaccardScore(x, x) == 1.0
	}
	if err := quick.Check(f, nil); err != nil {
		t.Errorf("TokenJaccardScore identity violated: %v", err)
	}
}

// TestProp_TokenJaccardScore_Symmetric asserts Score(a, b) ==
// Score(b, a) EXACTLY (bit-for-bit). Tokenise is deterministic; set
// construction via map[string]struct{} is order-independent; the
// integer-counter intersection cardinality is invariant under argument
// swap; the single division on integer-derived float64 values
// produces identical output regardless of argument order.
func TestProp_TokenJaccardScore_Symmetric(t *testing.T) {
	f := func(a, b string) bool {
		return fuzzymatch.TokenJaccardScore(a, b) == fuzzymatch.TokenJaccardScore(b, a)
	}
	if err := quick.Check(f, nil); err != nil {
		t.Errorf("TokenJaccardScore not symmetric: %v", err)
	}
}

// TestProp_TokenJaccardScore_NoNaN asserts the score never returns
// NaN. The identity / both-empty / one-empty guards gate away the only
// potential 0/0 paths; the union==0 defensive guard provides the
// secondary guard.
func TestProp_TokenJaccardScore_NoNaN(t *testing.T) {
	f := func(a, b string) bool {
		return !math.IsNaN(fuzzymatch.TokenJaccardScore(a, b))
	}
	if err := quick.Check(f, nil); err != nil {
		t.Errorf("TokenJaccardScore produced NaN: %v", err)
	}
}

// TestProp_TokenJaccardScore_NoInf asserts the score never returns
// ±Inf. Numerator and denominator are bounded integer set
// cardinalities fitting in float64; the single division never
// overflows.
func TestProp_TokenJaccardScore_NoInf(t *testing.T) {
	f := func(a, b string) bool {
		return !math.IsInf(fuzzymatch.TokenJaccardScore(a, b), 0)
	}
	if err := quick.Check(f, nil); err != nil {
		t.Errorf("TokenJaccardScore produced Inf: %v", err)
	}
}

// TestProp_TokenJaccardScore_NoNegativeZero asserts that when the
// score is 0.0 it is positive zero, not negative zero. The numerator
// (intersection cardinality) is a non-negative integer; float64(0) /
// float64(positive) is +0.0 in IEEE-754.
func TestProp_TokenJaccardScore_NoNegativeZero(t *testing.T) {
	f := func(a, b string) bool {
		s := fuzzymatch.TokenJaccardScore(a, b)
		return s != 0.0 || !math.Signbit(s)
	}
	if err := quick.Check(f, nil); err != nil {
		t.Errorf("TokenJaccardScore produced -0.0: %v", err)
	}
}

// --- Monge-Elkan property tests (plan 06-05)
//
// Two surfaces share the same property suite, with the asymmetric
// direct call (MongeElkanScore) carrying its own asymmetry-conditional
// property test mirroring Tversky α≠β. The symmetric variant
// (MongeElkanScoreSymmetric) participates in the standard symmetric
// set without exemption — the dispatch wrapper binds the symmetric
// variant per CONTEXT.md §4 LOCKED so AlgoMongeElkan also passes the
// dispatch-level symmetric property test (the standard
// PropAlgorithmScore_Symmetric set is per-AlgoID; AlgoMongeElkan is
// listed alongside all other symmetric algorithms).
//
// Inner-metric coercion via fuzzCoerceMongeElkanInner (defined in
// monge_elkan_fuzz_test.go) maps arbitrary int generators into the
// 14-entry permitted allow-list so the documented panic path is
// never exercised by the property harness (the panic contract is
// unit-tested separately by TestMongeElkan_PanicsOnNonPermittedInner).

// TestProp_MongeElkanScore_RangeBounds asserts the asymmetric
// direct-call score stays in [0, 1] for any (a, b, inner).
func TestProp_MongeElkanScore_RangeBounds(t *testing.T) {
	f := func(a, b string, innRaw int) bool {
		inner := fuzzCoerceMongeElkanInner(innRaw)
		s := fuzzymatch.MongeElkanScore(a, b, inner, fuzzymatch.DefaultNormalisationOptions())
		return s >= 0.0 && s <= 1.0 && !math.IsNaN(s) && !math.IsInf(s, 0)
	}
	if err := quick.Check(f, nil); err != nil {
		t.Errorf("MongeElkanScore out of [0, 1] or non-finite: %v", err)
	}
}

// TestProp_MongeElkanScore_NoNaN asserts the asymmetric score never
// returns NaN. The all-empty / one-empty / identity short-circuits
// gate away the only potential 0/0 paths.
func TestProp_MongeElkanScore_NoNaN(t *testing.T) {
	f := func(a, b string, innRaw int) bool {
		inner := fuzzCoerceMongeElkanInner(innRaw)
		return !math.IsNaN(fuzzymatch.MongeElkanScore(a, b, inner, fuzzymatch.DefaultNormalisationOptions()))
	}
	if err := quick.Check(f, nil); err != nil {
		t.Errorf("MongeElkanScore produced NaN: %v", err)
	}
}

// TestProp_MongeElkanScore_NoInf asserts the asymmetric score never
// returns ±Inf. The inner metric returns [0, 1]; |tA| ≥ 1 in the
// non-short-circuit path; the single division never overflows.
func TestProp_MongeElkanScore_NoInf(t *testing.T) {
	f := func(a, b string, innRaw int) bool {
		inner := fuzzCoerceMongeElkanInner(innRaw)
		return !math.IsInf(fuzzymatch.MongeElkanScore(a, b, inner, fuzzymatch.DefaultNormalisationOptions()), 0)
	}
	if err := quick.Check(f, nil); err != nil {
		t.Errorf("MongeElkanScore produced Inf: %v", err)
	}
}

// TestProp_MongeElkanScore_NoNegativeZero asserts that when the
// asymmetric score is 0.0 it is positive zero, not negative zero. The
// numerator is a non-negative accumulator; the denominator is a
// strictly-positive int; float64(0) / float64(positive) is +0.0 in
// IEEE-754.
func TestProp_MongeElkanScore_NoNegativeZero(t *testing.T) {
	f := func(a, b string, innRaw int) bool {
		inner := fuzzCoerceMongeElkanInner(innRaw)
		s := fuzzymatch.MongeElkanScore(a, b, inner, fuzzymatch.DefaultNormalisationOptions())
		return s != 0.0 || !math.Signbit(s)
	}
	if err := quick.Check(f, nil); err != nil {
		t.Errorf("MongeElkanScore produced -0.0: %v", err)
	}
}

// TestProp_MongeElkanScore_AsymmetricWhenTokenCountAsymmetric is the
// asymmetry-conditional property test (mirrors
// TestProp_TverskyScore_AsymmetricWhenAlphaNeqBeta). With fixed
// inner = AlgoLevenshtein (which produces non-zero scores on
// partially-overlapping tokens), the implication is:
//
//	IF |tokens(a)| ≠ |tokens(b)| AND fwd > 0 AND fwd != 1.0
//	THEN MongeElkanScore(a, b, inner) ≠ MongeElkanScore(b, a, inner).
//
// The fwd != 1.0 guard excludes the corner case where the per-token
// max-mean happens to collapse to 1.0 in both directions (which can
// happen when one side is a strict singleton subset of the other AND
// the lone token is identical-to-self). The fwd > 0 guard excludes
// fully-orthogonal pairs where both directions collapse to 0.0.
//
// Detecting whether token-counts differ from inside the property body
// without re-implementing Tokenise: len(strings.Fields(s)) is a coarse
// approximation for whitespace-only inputs. For mixed identifier-style
// inputs the project Tokenise produces semantically richer splits, but
// the property holds whenever the TOKEN counts differ — the
// approximation under-counts conservative cases where Tokenise splits
// camelCase, which means the property STILL holds for those (the
// premise is just looser than strictly required).
//
// Spot-check on the canonical RV-ME6 / RV-ME4 pair confirms the
// property body actively detects asymmetric inputs (a degenerate
// `return true` would otherwise slip past quick.Check).
func TestProp_MongeElkanScore_AsymmetricWhenTokenCountAsymmetric(t *testing.T) {
	inner := fuzzymatch.AlgoLevenshtein
	opts := fuzzymatch.DefaultNormalisationOptions()
	f := func(a, b string) bool {
		if a == b || a == "" || b == "" {
			return true // short-circuit branches
		}
		fwd := fuzzymatch.MongeElkanScore(a, b, inner, opts)
		rev := fuzzymatch.MongeElkanScore(b, a, inner, opts)
		// Token-count premise — approximate via strings.Fields (a
		// whitespace split). The project Tokenise can split FURTHER on
		// identifier boundaries, so this under-estimates the token-count
		// imbalance; the property still holds whenever token counts
		// differ.
		aTokens := len(strings.Fields(a))
		bTokens := len(strings.Fields(b))
		// Premise: token counts differ AND fwd is a partial-overlap
		// score (0 < fwd < 1). Both bounds are needed: fwd == 0 means
		// orthogonal pairs (rev also 0); fwd == 1 means a subset-with-
		// identical-tokens case where both directions trivially equal 1.
		if aTokens == bTokens || fwd == 0.0 || fwd == 1.0 {
			return true // vacuous truth
		}
		// Premise holds — scores must differ.
		return fwd != rev
	}
	if err := quick.Check(f, nil); err != nil {
		t.Errorf("MongeElkanScore asymmetry-conditional property violated: %v", err)
	}
	// Spot-check on the RV-ME4 / RV-ME6 canonical pair.
	rvME4 := fuzzymatch.MongeElkanScore("alpha", "alpha beta gamma", inner, opts)
	rvME6 := fuzzymatch.MongeElkanScore("alpha beta gamma", "alpha", inner, opts)
	if rvME4 == rvME6 {
		t.Errorf("asymmetry spot-check failed: ME(alpha, alpha beta gamma, Lev)=%g equals ME(alpha beta gamma, alpha, Lev)=%g — RV-ME4/RV-ME6 should differ (the load-bearing direction-sensitivity gate)",
			rvME4, rvME6)
	}
}

// TestProp_MongeElkanScoreSymmetric_RangeBounds asserts the symmetric
// variant's score stays in [0, 1] for any (a, b, inner).
func TestProp_MongeElkanScoreSymmetric_RangeBounds(t *testing.T) {
	f := func(a, b string, innRaw int) bool {
		inner := fuzzCoerceMongeElkanInner(innRaw)
		s := fuzzymatch.MongeElkanScoreSymmetric(a, b, inner, fuzzymatch.DefaultNormalisationOptions())
		return s >= 0.0 && s <= 1.0 && !math.IsNaN(s) && !math.IsInf(s, 0)
	}
	if err := quick.Check(f, nil); err != nil {
		t.Errorf("MongeElkanScoreSymmetric out of [0, 1] or non-finite: %v", err)
	}
}

// TestProp_MongeElkanScoreSymmetric_Identity asserts that for any
// non-empty x and any permitted inner, the symmetric variant returns
// 1.0 EXACTLY (the identity short-circuit inside MongeElkanScore
// fires before any inner-metric work; the symmetric variant averages
// 1.0 + 1.0 / 2 = 1.0).
func TestProp_MongeElkanScoreSymmetric_Identity(t *testing.T) {
	f := func(x string, innRaw int) bool {
		inner := fuzzCoerceMongeElkanInner(innRaw)
		return fuzzymatch.MongeElkanScoreSymmetric(x, x, inner, fuzzymatch.DefaultNormalisationOptions()) == 1.0
	}
	if err := quick.Check(f, nil); err != nil {
		t.Errorf("MongeElkanScoreSymmetric identity violated: %v", err)
	}
}

// TestProp_MongeElkanScoreSymmetric_Symmetric is the load-bearing
// symmetric-variant property: MongeElkanScoreSymmetric(a, b, inner)
// must equal MongeElkanScoreSymmetric(b, a, inner) for all inputs,
// EXACTLY (bit-for-bit). The construction (ME(A,B) + ME(B,A))/2 is
// invariant under argument swap (the sum of two terms swapped is the
// same sum), and the divide-by-2 on a sum is exact in IEEE-754.
func TestProp_MongeElkanScoreSymmetric_Symmetric(t *testing.T) {
	f := func(a, b string, innRaw int) bool {
		inner := fuzzCoerceMongeElkanInner(innRaw)
		opts := fuzzymatch.DefaultNormalisationOptions()
		return fuzzymatch.MongeElkanScoreSymmetric(a, b, inner, opts) == fuzzymatch.MongeElkanScoreSymmetric(b, a, inner, opts)
	}
	if err := quick.Check(f, nil); err != nil {
		t.Errorf("MongeElkanScoreSymmetric not symmetric: %v", err)
	}
}

// TestProp_MongeElkanScoreSymmetric_NoNaN asserts the symmetric
// variant never returns NaN.
func TestProp_MongeElkanScoreSymmetric_NoNaN(t *testing.T) {
	f := func(a, b string, innRaw int) bool {
		inner := fuzzCoerceMongeElkanInner(innRaw)
		return !math.IsNaN(fuzzymatch.MongeElkanScoreSymmetric(a, b, inner, fuzzymatch.DefaultNormalisationOptions()))
	}
	if err := quick.Check(f, nil); err != nil {
		t.Errorf("MongeElkanScoreSymmetric produced NaN: %v", err)
	}
}

// TestProp_MongeElkanScoreSymmetric_NoInf asserts the symmetric variant
// never returns ±Inf.
func TestProp_MongeElkanScoreSymmetric_NoInf(t *testing.T) {
	f := func(a, b string, innRaw int) bool {
		inner := fuzzCoerceMongeElkanInner(innRaw)
		return !math.IsInf(fuzzymatch.MongeElkanScoreSymmetric(a, b, inner, fuzzymatch.DefaultNormalisationOptions()), 0)
	}
	if err := quick.Check(f, nil); err != nil {
		t.Errorf("MongeElkanScoreSymmetric produced Inf: %v", err)
	}
}

// TestProp_MongeElkanScoreSymmetric_NoNegativeZero asserts that when
// the symmetric score is 0.0 it is positive zero, not negative zero.
func TestProp_MongeElkanScoreSymmetric_NoNegativeZero(t *testing.T) {
	f := func(a, b string, innRaw int) bool {
		inner := fuzzCoerceMongeElkanInner(innRaw)
		s := fuzzymatch.MongeElkanScoreSymmetric(a, b, inner, fuzzymatch.DefaultNormalisationOptions())
		return s != 0.0 || !math.Signbit(s)
	}
	if err := quick.Check(f, nil); err != nil {
		t.Errorf("MongeElkanScoreSymmetric produced -0.0: %v", err)
	}
}

// ---------------------------------------------------------------------------
// Soundex (Phase 7 plan 07-01)
//
// Soundex is SYMMETRIC: SoundexScore(a,b) == SoundexScore(b,a) always, because
// code-equality is symmetric. The Symmetric invariant IS included (unlike
// RatcliffObershelp which is asymmetric by design).
//
// PropSoundex_CodeCharset asserts the output matches ^([A-Z][0-9]{3})?$
// (empty string or exactly 4 chars: letter + 3 digits).
// ---------------------------------------------------------------------------

// TestProp_SoundexScore_RangeBounds asserts the score is in [0.0, 1.0]
// for any pair of strings. DET-04 range-bounds invariant.
func TestProp_SoundexScore_RangeBounds(t *testing.T) {
	f := func(a, b string) bool {
		s := fuzzymatch.SoundexScore(a, b)
		return s >= 0.0 && s <= 1.0
	}
	if err := quick.Check(f, nil); err != nil {
		t.Errorf("SoundexScore out of [0,1]: %v", err)
	}
}

// TestProp_SoundexScore_Identity asserts Score(x, x) == 1.0 for any string x.
func TestProp_SoundexScore_Identity(t *testing.T) {
	f := func(x string) bool {
		return fuzzymatch.SoundexScore(x, x) == 1.0
	}
	if err := quick.Check(f, nil); err != nil {
		t.Errorf("SoundexScore identity violated: %v", err)
	}
}

// TestProp_SoundexScore_Symmetric asserts Score(a,b) == Score(b,a).
// Soundex is symmetric because code-equality is symmetric.
func TestProp_SoundexScore_Symmetric(t *testing.T) {
	f := func(a, b string) bool {
		return fuzzymatch.SoundexScore(a, b) == fuzzymatch.SoundexScore(b, a)
	}
	if err := quick.Check(f, nil); err != nil {
		t.Errorf("SoundexScore symmetry violated: %v", err)
	}
}

// TestProp_SoundexScore_NoNaN asserts the score is never NaN.
func TestProp_SoundexScore_NoNaN(t *testing.T) {
	f := func(a, b string) bool {
		return !math.IsNaN(fuzzymatch.SoundexScore(a, b))
	}
	if err := quick.Check(f, nil); err != nil {
		t.Errorf("SoundexScore produced NaN: %v", err)
	}
}

// TestProp_SoundexScore_NoInf asserts the score never returns ±Inf.
func TestProp_SoundexScore_NoInf(t *testing.T) {
	f := func(a, b string) bool {
		return !math.IsInf(fuzzymatch.SoundexScore(a, b), 0)
	}
	if err := quick.Check(f, nil); err != nil {
		t.Errorf("SoundexScore produced Inf: %v", err)
	}
}

// TestProp_SoundexCode_Charset asserts that SoundexCode output matches
// ^([A-Z][0-9]{3})?$ — either empty string (for empty/non-ASCII input)
// or exactly 4 characters (1 uppercase letter + 3 digits).
func TestProp_SoundexCode_Charset(t *testing.T) {
	f := func(s string) bool {
		code := fuzzymatch.SoundexCode(s)
		if code == "" {
			return true // empty input or all-non-ASCII
		}
		if len(code) != 4 {
			return false
		}
		if code[0] < 'A' || code[0] > 'Z' {
			return false
		}
		for i := 1; i < 4; i++ {
			if code[i] < '0' || code[i] > '9' {
				return false
			}
		}
		return true
	}
	if err := quick.Check(f, nil); err != nil {
		t.Errorf("SoundexCode charset invariant violated: %v", err)
	}
}
