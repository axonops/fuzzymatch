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

// token_sort_ratio_bench_test.go runs allocation-aware benchmarks for
// the TokenSortRatioScore public surface at multiple input sizes.
// b.ReportAllocs() on every benchmark gates allocation regressions in
// bench.txt via benchstat (Phase 6 finalisation in plan 06-06).
//
// Performance budget per .claude/skills/performance-standards
// (inherited from Phase 2/3/4/5):
//
//   - Token Sort / Set / Jaccard Ratio: < 5 µs per call, ≤ 4
//     allocations on ASCII Short (one [string,4] slice per Tokenise
//     call × 2 sides + 2 joined-string allocs from strings.Join).
//   - Longer inputs see additional allocs from Tokenise's []rune
//     conversion (~1 per side) and any DP-row heap allocations when
//     min(joined lengths) > maxStackInputLen (= 64).
//
// `var sink` outside the loop + a `sink < 0` gate after the loop
// prevents compiler dead-code elimination (locked Phase 2 pattern).

package fuzzymatch_test

import (
	"strings"
	"testing"

	"github.com/axonops/fuzzymatch"
)

// tokenSortA50 / B50 are 50-character ASCII strings used by the
// medium-length benchmark. Five-token inputs (typical identifier-like
// content) on both sides with one token reordered to exercise the
// sort-and-LCS path rather than the identity short-circuit.
const (
	tokenSortA50 = "alpha beta gamma delta epsilon zeta eta theta iota"
	tokenSortB50 = "epsilon delta alpha gamma beta zeta eta theta iota"
)

// BenchmarkTokenSortRatioScore_ASCII_Short exercises the canonical
// "fuzzy wuzzy" reorder pair — five tokens on each side, sorted-joined
// strings identical. The result is 1.0 via the indelRatio kernel
// (identical joined strings → LCS = full length → ratio 1.0). Expected
// ≤ 4 allocs/op: 2 Tokenise output slices + 2 strings.Join output strings.
func BenchmarkTokenSortRatioScore_ASCII_Short(b *testing.B) {
	b.ReportAllocs()
	b.ResetTimer()
	var sink float64
	for i := 0; i < b.N; i++ {
		sink = fuzzymatch.TokenSortRatioScore("fuzzy wuzzy was a bear", "wuzzy fuzzy was a bear")
	}
	if sink < 0 {
		b.Fatal("sink unexpectedly negative — compiler folded the benchmark away")
	}
}

// BenchmarkTokenSortRatioScore_ASCII_Medium exercises 50-char inputs
// with nine tokens per side reordered. The joined strings have
// min-length around 50 bytes — within the stack-buffer threshold of
// maxStackInputLen (= 64), so the DP rows stay on the stack.
func BenchmarkTokenSortRatioScore_ASCII_Medium(b *testing.B) {
	b.ReportAllocs()
	b.ResetTimer()
	var sink float64
	for i := 0; i < b.N; i++ {
		sink = fuzzymatch.TokenSortRatioScore(tokenSortA50, tokenSortB50)
	}
	if sink < 0 {
		b.Fatal("sink unexpectedly negative — compiler folded the benchmark away")
	}
}

// BenchmarkTokenSortRatioScore_ASCII_Long exercises ~200-char inputs
// (20 tokens per side, slight rotation). The joined strings exceed
// maxStackInputLen (= 64) so DP rows allocate on the heap (2 allocs);
// total ≤ ~8 allocs/op including Tokenise + Join.
func BenchmarkTokenSortRatioScore_ASCII_Long(b *testing.B) {
	aLong := strings.Repeat("alpha beta gamma delta ", 10) // 230 bytes, 40 tokens
	bLong := strings.Repeat("delta alpha gamma beta ", 10) // mirror reorder
	b.ReportAllocs()
	b.ResetTimer()
	var sink float64
	for i := 0; i < b.N; i++ {
		sink = fuzzymatch.TokenSortRatioScore(aLong, bLong)
	}
	if sink < 0 {
		b.Fatal("sink unexpectedly negative — compiler folded the benchmark away")
	}
}

// BenchmarkTokenSortRatioScore_Unicode_Short exercises a multi-byte
// UTF-8 pair. Tokenise allocates a []rune internally per side; the
// LCS DP still operates on byte slices of the joined strings.
func BenchmarkTokenSortRatioScore_Unicode_Short(b *testing.B) {
	b.ReportAllocs()
	b.ResetTimer()
	var sink float64
	for i := 0; i < b.N; i++ {
		sink = fuzzymatch.TokenSortRatioScore("café société", "société café")
	}
	if sink < 0 {
		b.Fatal("sink unexpectedly negative — compiler folded the benchmark away")
	}
}
