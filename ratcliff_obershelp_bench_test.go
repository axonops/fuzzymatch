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

// ratcliff_obershelp_bench_test.go runs allocation-aware benchmarks for the
// two Ratcliff-Obershelp public surfaces at multiple input sizes.
// b.ReportAllocs() on every benchmark gates allocation regressions in
// bench.txt via benchstat.
//
// Performance budgets (Q8d, docs/requirements.md §14.1 — revised
// 2026-05 to match implementation reality):
//
//   - ASCII Short  (~10 chars):  ≤ 4 allocs/op (recursive decomposition
//                                allocates 2 rolling rows per recursion
//                                level; depth ≈ 1-2 on short inputs)
//   - ASCII Medium (~50 chars):  informational (recursion depth grows
//                                with the number of matched substrings)
//   - ASCII Long   (~500 chars): informational — long-input row in §14.1
//                                permits ≥ 100 allocs as recursion
//                                depth scales with input size
//   - Unicode Short (rune path): adds 2 allocs for []rune conversion
//
// Q8d note: tighter Short budget (≤ 2) was judged too complex for marginal
// gain given the recursive structure (no ASCII fast path is structurally
// available — each recursive call works on different substring boundaries).
// See Q7c scope note on RatcliffObershelpScore godoc for the long-input
// fallback rationale.
//
// Allocation behaviour notes:
//
// Each recursive call to roFindLongestMatch / roFindLongestMatchRunes
// allocates two rolling-row slices ([]int of size lb+1). The recursion
// depth scales with the number of matched substrings discovered. For
// short ASCII inputs (~10 chars), expect a handful of allocations per
// call. For longer inputs and the rune-path variant, allocations scale
// with recursion depth.
//
// The byte path on a substring-input slices the input string directly
// (slice-header only — no copy), so recursion into a[aHi:] / b[bHi:]
// adds no allocation beyond the two rolling rows per level.
//
// `var sink` outside the loop + `if sink < 0 { b.Fatal(...) }` after the
// loop prevents compiler dead-code elimination (locked Phase 2 pattern).

package fuzzymatch_test

import (
	"strings"
	"testing"

	"github.com/axonops/fuzzymatch"
)

// roA50 and roB50 are 50-byte ASCII strings used for the ASCII_Medium
// benchmarks. Distinct from the lcsstr 50-byte constants to keep grep
// readable.
const (
	roA50 = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWX"
	roB50 = "bcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXY"
)

// BenchmarkRatcliffObershelpScore_ASCII_Short benchmarks the byte path on a
// short ASCII pair (the canonical Dr. Dobb's 1988 WIKIMEDIA/WIKIMANIA pair).
// Q8d budget: ≤ 4 allocs/op (recursion depth ≈ 1-2 on short inputs).
// Recursion depth depends on the matched substrings discovered.
func BenchmarkRatcliffObershelpScore_ASCII_Short(b *testing.B) {
	b.ReportAllocs()
	b.ResetTimer()
	var sink float64
	for i := 0; i < b.N; i++ {
		sink = fuzzymatch.RatcliffObershelpScore("WIKIMEDIA", "WIKIMANIA")
	}
	if sink < 0 {
		b.Fatal("sink unexpectedly negative — compiler folded the benchmark away")
	}
}

// BenchmarkRatcliffObershelpScore_ASCII_Medium benchmarks the byte path on
// a 50-char ASCII pair. Recursion may go deeper as more partial matches
// are uncovered. Informational.
func BenchmarkRatcliffObershelpScore_ASCII_Medium(b *testing.B) {
	b.ReportAllocs()
	b.ResetTimer()
	var sink float64
	for i := 0; i < b.N; i++ {
		sink = fuzzymatch.RatcliffObershelpScore(roA50, roB50)
	}
	if sink < 0 {
		b.Fatal("sink unexpectedly negative — compiler folded the benchmark away")
	}
}

// BenchmarkRatcliffObershelpScore_ASCII_Long benchmarks the byte path on
// 500-char ASCII inputs (well above any plausible stack-fast-path
// threshold). This benchmark establishes the regression baseline for
// long-input pathological cases per the threat model.
func BenchmarkRatcliffObershelpScore_ASCII_Long(b *testing.B) {
	aLong := strings.Repeat("abcde", 100)
	bLong := strings.Repeat("bcdef", 100)
	b.ReportAllocs()
	b.ResetTimer()
	var sink float64
	for i := 0; i < b.N; i++ {
		sink = fuzzymatch.RatcliffObershelpScore(aLong, bLong)
	}
	if sink < 0 {
		b.Fatal("sink unexpectedly negative — compiler folded the benchmark away")
	}
}

// BenchmarkRatcliffObershelpScore_Unicode_Short exercises the byte path on
// a multi-byte UTF-8 pair. The byte path operates on UTF-8 bytes directly,
// so this measures the cost of byte-comparison through the recursion on
// multi-byte input. Informational.
func BenchmarkRatcliffObershelpScore_Unicode_Short(b *testing.B) {
	b.ReportAllocs()
	b.ResetTimer()
	var sink float64
	for i := 0; i < b.N; i++ {
		sink = fuzzymatch.RatcliffObershelpScore("café", "cafe")
	}
	if sink < 0 {
		b.Fatal("sink unexpectedly negative — compiler folded the benchmark away")
	}
}

// BenchmarkRatcliffObershelpScoreRunes_Unicode_Short exercises the rune
// path. Expects 2 additional allocations beyond the byte path for the
// `[]rune(a)` / `[]rune(b)` conversions, plus the rolling-row slices per
// recursion level. Informational.
func BenchmarkRatcliffObershelpScoreRunes_Unicode_Short(b *testing.B) {
	b.ReportAllocs()
	b.ResetTimer()
	var sink float64
	for i := 0; i < b.N; i++ {
		sink = fuzzymatch.RatcliffObershelpScoreRunes("café", "cafe")
	}
	if sink < 0 {
		b.Fatal("sink unexpectedly negative — compiler folded the benchmark away")
	}
}
