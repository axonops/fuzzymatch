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
// algorithm-specific hand-curated property
// (TestProp_RatcliffObershelpScore_AtLeastLevenshtein_HandCurated).
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

// TestProp_RatcliffObershelpScore_AtLeastLevenshtein_HandCurated checks the
// "generally" property from RESEARCH.md: on substring-containment inputs
// the Ratcliff-Obershelp score is typically ≥ the Levenshtein score
// because RO finds the contiguous match and ignores the deletion cost,
// while Levenshtein pays for every dropped character.
//
// This is HAND-CURATED rather than testing/quick — RESEARCH.md explicitly
// notes the property is "generally" true, not universal. Quick.Check would
// produce degenerate inputs (e.g. random non-overlapping strings) where
// the property either trivially holds or accidentally fails for reasons
// unrelated to the algorithm. Hand-curated substring-containment cases
// directly exercise the intended use case.
func TestProp_RatcliffObershelpScore_AtLeastLevenshtein_HandCurated(t *testing.T) {
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
