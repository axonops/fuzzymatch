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

// partial_ratio_bench_test.go runs allocation-aware benchmarks for the
// two Partial Ratio public surfaces at multiple input sizes plus the
// LOCKED Pathological_LongShortMismatch fixture per 06-CONTEXT.md §5.
// b.ReportAllocs() on every benchmark gates allocation regressions in
// bench.txt via benchstat (Phase 6 finalisation in plan 06-06).
//
// Performance budget per RESEARCH.md §4.1 + .claude/skills/performance-
// standards (inherited from Phase 2/3/4/5):
//
//   - ASCII Short  (~10 chars):     low single-digit µs/op
//   - ASCII Medium (~50 chars):     mid single-digit µs/op
//   - ASCII Long   (~200 chars):    proportional to |s|·|l|·|s|
//   - Unicode Short (rune path):    rune charSet allocates one map
//
// The LOCKED Pathological_LongShortMismatch fixture (10-char shorter
// vs 10000-char longer) is the DoS-vector benchmark per 06-CONTEXT.md
// §5 LOCKED. The s1_char_set early-skip is load-bearing for this
// fixture: without it, ~10,000 indelRatio calls are made, each ~100
// LCS DP cell updates. With the char-set early-skip, ~95% of those
// calls are pruned when the shared alphabet is small. Target: < 100 µs
// per call on developer hardware.
//
// `var sink` outside the loop + a `sink < 0` gate after the loop
// prevents compiler dead-code elimination (locked Phase 2 pattern).

package fuzzymatch_test

import (
	"strings"
	"testing"

	"github.com/axonops/fuzzymatch"
)

// partialRatioA50 / B50 are 50-character ASCII strings used by the
// medium-length benchmark. shorter is "hello world" (11 chars);
// longer is constructed to embed the shorter string at the middle.
const (
	partialRatioA50 = "hello world"
	partialRatioB50 = "the quick brown hello world fox jumps over the lazy"
)

// BenchmarkPartialRatioScore_ASCII_Short exercises the byte path on
// the canonical RapidFuzz `("YANKEES", "NEW YORK YANKEES")` reference
// pair (Region 2 middle wins → 1.0).
func BenchmarkPartialRatioScore_ASCII_Short(b *testing.B) {
	b.ReportAllocs()
	b.ResetTimer()
	var sink float64
	for i := 0; i < b.N; i++ {
		sink = fuzzymatch.PartialRatioScore("YANKEES", "NEW YORK YANKEES")
	}
	if sink < 0 {
		b.Fatal("sink unexpectedly negative — compiler folded the benchmark away")
	}
}

// BenchmarkPartialRatioScore_ASCII_Medium exercises the byte path on
// the 50-char fixture above. Region 2 middle wins because "hello world"
// appears verbatim in the middle of the longer string.
func BenchmarkPartialRatioScore_ASCII_Medium(b *testing.B) {
	b.ReportAllocs()
	b.ResetTimer()
	var sink float64
	for i := 0; i < b.N; i++ {
		sink = fuzzymatch.PartialRatioScore(partialRatioA50, partialRatioB50)
	}
	if sink < 0 {
		b.Fatal("sink unexpectedly negative — compiler folded the benchmark away")
	}
}

// BenchmarkPartialRatioScore_ASCII_Long exercises the byte path on
// ~200-char inputs. shorter is "alpha beta gamma" (16 chars);
// longer is constructed via strings.Repeat to embed the shorter string
// at the END so Region 2 catches it at i = n - m.
func BenchmarkPartialRatioScore_ASCII_Long(b *testing.B) {
	aLong := "alpha beta gamma" // 16 chars
	// Construct a ~200-char longer string ending in "alpha beta gamma".
	// Region 2 catches the perfect match at i = n - m.
	bLong := strings.Repeat("xyzqrs ", 26) + "alpha beta gamma" // 7*26 + 16 = 198 chars
	b.ReportAllocs()
	b.ResetTimer()
	var sink float64
	for i := 0; i < b.N; i++ {
		sink = fuzzymatch.PartialRatioScore(aLong, bLong)
	}
	if sink < 0 {
		b.Fatal("sink unexpectedly negative — compiler folded the benchmark away")
	}
}

// BenchmarkPartialRatioScoreRunes_ASCII_Short exercises the rune-path
// surface on the canonical ("YANKEES", "NEW YORK YANKEES") pair. The
// rune path allocates two []rune slices plus one map[rune]struct{}.
func BenchmarkPartialRatioScoreRunes_ASCII_Short(b *testing.B) {
	b.ReportAllocs()
	b.ResetTimer()
	var sink float64
	for i := 0; i < b.N; i++ {
		sink = fuzzymatch.PartialRatioScoreRunes("YANKEES", "NEW YORK YANKEES")
	}
	if sink < 0 {
		b.Fatal("sink unexpectedly negative — compiler folded the benchmark away")
	}
}

// BenchmarkPartialRatioScoreRunes_Unicode_Short exercises the rune-path
// surface on the canonical Unicode keystone ("café", "caf") pair —
// the rune path correctly handles the 3-rune subsequence of the
// 4-rune input.
func BenchmarkPartialRatioScoreRunes_Unicode_Short(b *testing.B) {
	b.ReportAllocs()
	b.ResetTimer()
	var sink float64
	for i := 0; i < b.N; i++ {
		sink = fuzzymatch.PartialRatioScoreRunes("café", "caf")
	}
	if sink < 0 {
		b.Fatal("sink unexpectedly negative — compiler folded the benchmark away")
	}
}

// BenchmarkPartialRatio_Pathological_LongShortMismatch_Bytes is the
// LOCKED DoS-vector fixture per 06-CONTEXT.md §5 LOCKED.
//
// Fixture: shorter = 10-char ASCII string; longer = 10000-char ASCII
// string with the shorter string embedded once at the end. The
// s1_char_set early-skip prunes ~95% of the ~10000 Region 2 alignments
// because the longer string is constructed from a small alphabet
// disjoint from the shorter's alphabet except for one shared character.
//
// Target: < 100 µs per call on developer hardware. The actual measured
// timing is recorded in plan SUMMARY.md and committed to bench.txt in
// plan 06-06 finalisation. benchstat regression detection > 10% fails
// CI per docs/requirements.md §6(6).
//
// Without the char-set early-skip, this benchmark would take O(10000)
// indelRatio calls, each O(100) DP cells = O(10^6) operations per
// call. With the early-skip, the inner loop is mostly a single
// charSet lookup per alignment.
//
// Per CONTEXT §5: the fixture inputs are precomputed OUTSIDE the
// timed loop so allocations / strings.Repeat don't pollute the
// benchmark numbers.
func BenchmarkPartialRatio_Pathological_LongShortMismatch_Bytes(b *testing.B) {
	// shorter has 10 chars with a distinct alphabet (a..j).
	shorter := "abcdefghij"
	// longer has 10000 chars from a 3-char alphabet (x,y,z) disjoint
	// from shorter EXCEPT for the trailing 'j' which guarantees ONE
	// alignment matches at the right edge (so best > 0; not a degenerate
	// all-zero case).
	longer := strings.Repeat("xyz", 3333) + "j" // 9999 + 1 = 10000 chars
	b.ReportAllocs()
	b.ResetTimer()
	var sink float64
	for i := 0; i < b.N; i++ {
		sink = fuzzymatch.PartialRatioScore(shorter, longer)
	}
	if sink < 0 {
		b.Fatal("sink unexpectedly negative — compiler folded the benchmark away")
	}
}

// BenchmarkPartialRatio_Pathological_LongShortMismatch_Runes is the
// rune-path companion to the bytes-path pathological benchmark above.
// Same input shape (10 runes vs 10000 runes); the rune path
// additionally allocates two []rune slices and one map[rune]struct{}.
func BenchmarkPartialRatio_Pathological_LongShortMismatch_Runes(b *testing.B) {
	shorter := "abcdefghij"
	longer := strings.Repeat("xyz", 3333) + "j"
	b.ReportAllocs()
	b.ResetTimer()
	var sink float64
	for i := 0; i < b.N; i++ {
		sink = fuzzymatch.PartialRatioScoreRunes(shorter, longer)
	}
	if sink < 0 {
		b.Fatal("sink unexpectedly negative — compiler folded the benchmark away")
	}
}
