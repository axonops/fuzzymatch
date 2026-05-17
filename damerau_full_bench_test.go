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

// damerau_full_bench_test.go runs allocation-aware benchmarks for
// DamerauLevenshteinFullScore at four input sizes. b.ReportAllocs() on
// every benchmark gates allocation regressions in bench.txt via benchstat.
//
// Performance budgets (Q7b / Q8a, docs/requirements.md §14.1 — revised
// 2026-05 to match implementation reality):
//
//   - ASCII Short  (≤ 4 chars):  ≤ 1 alloc/op  (the flat DP slice)
//   - ASCII Medium (~50 chars):  ≤ 1 alloc/op, < 8.2 µs wall (Q7b)
//   - ASCII Long   (~500 chars): informational — O(m·n) DP table dominates
//   - Unicode Short (rune path): ≤ 3 allocs/op (DP slice + 2 []rune)
//
// Q7b/Q8a budget notes:
//
//   - The DL-Full v1.0 implementation unconditionally heap-allocates the
//     full (m+2)×(n+2) DP table (1 alloc, byte-count O(m·n)·sizeof(int)).
//     There is no ASCII fast path or stack-buffer short-circuit —
//     Lowrance-Wagner 1975 requires simultaneous access to the entire DP
//     matrix for the transposition look-back (see Q7c scope note on
//     DamerauLevenshteinFullDistance godoc).
//   - The earlier 0-alloc target documented in v0.x was unachievable: it
//     would require a ~34 KB stack frame at the 64-byte input ceiling,
//     judged too fragile against Go's escape-analysis quirks.
//   - The Q11e un-skipped TestDamerauLevenshteinFullScore_ShortAllocBudget_ASCII
//     enforces the ≤ 1 alloc Short budget at test time; this file
//     documents the same budget plus the timing target.
//
// These benchmarks document the current allocation profile and gate
// regressions via benchstat.

package fuzzymatch_test

import (
	"strings"
	"testing"

	"github.com/axonops/fuzzymatch"
)

// BenchmarkDamerauLevenshteinFullScore_ASCII_Short exercises DL-Full on a
// short ASCII pair (2 and 2 bytes — the canonical transposition pair).
// Q8a budget: ≤ 1 alloc/op (the flat DP slice). Allocates the full (4×4)
// DP table — 16 ints on heap.
func BenchmarkDamerauLevenshteinFullScore_ASCII_Short(b *testing.B) {
	b.ReportAllocs()
	b.ResetTimer()
	var sink float64
	for i := 0; i < b.N; i++ {
		sink = fuzzymatch.DamerauLevenshteinFullScore("ab", "ba")
	}
	if sink < 0 {
		b.Fatal("sink unexpectedly negative — compiler folded the benchmark away")
	}
}

// BenchmarkDamerauLevenshteinFullScore_ASCII_Medium exercises DL-Full on
// 50-char ASCII inputs (within maxStackInputLen=64). Q7b budget: ≤ 1
// alloc/op, < 8.2 µs wall (revised from < 3 µs to match implementation
// reality — the full DP table at 52×52 dominates wall-time). Allocates
// the full (52×52) DP table — 2704 ints on heap.
func BenchmarkDamerauLevenshteinFullScore_ASCII_Medium(b *testing.B) {
	b.ReportAllocs()
	b.ResetTimer()
	var sink float64
	for i := 0; i < b.N; i++ {
		sink = fuzzymatch.DamerauLevenshteinFullScore(a50, b50)
	}
	if sink < 0 {
		b.Fatal("sink unexpectedly negative — compiler folded the benchmark away")
	}
}

// BenchmarkDamerauLevenshteinFullScore_ASCII_Long exercises DL-Full on
// 500-char ASCII inputs. v1.0: allocates the full (502×502) DP table
// — ~250K ints on heap. This benchmark is informational for relative throughput.
func BenchmarkDamerauLevenshteinFullScore_ASCII_Long(b *testing.B) {
	aLong := strings.Repeat("abcde", 100)
	bLong := strings.Repeat("bcdef", 100)
	b.ReportAllocs()
	b.ResetTimer()
	var sink float64
	for i := 0; i < b.N; i++ {
		sink = fuzzymatch.DamerauLevenshteinFullScore(aLong, bLong)
	}
	if sink < 0 {
		b.Fatal("sink unexpectedly negative — compiler folded the benchmark away")
	}
}

// BenchmarkDamerauLevenshteinFullScore_Unicode_Short exercises the rune path
// on a short multi-byte UTF-8 pair. v1.0: 2 allocs for []rune + 1 alloc for
// the DP table (total 3 allocs).
func BenchmarkDamerauLevenshteinFullScore_Unicode_Short(b *testing.B) {
	b.ReportAllocs()
	b.ResetTimer()
	var sink float64
	for i := 0; i < b.N; i++ {
		sink = fuzzymatch.DamerauLevenshteinFullScoreRunes("café", "cafe")
	}
	if sink < 0 {
		b.Fatal("sink unexpectedly negative — compiler folded the benchmark away")
	}
}
