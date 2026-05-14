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
// Performance notes (PERF-01, damerau_full.go implementation discipline):
//
//   - ASCII all sizes: allocates O(m·n) ints for the full DP table (v1.0).
//     The two-row + auxiliary-table 0-alloc optimisation is a v1.x follow-up.
//   - Unicode short (rune path): 1 alloc for full DP table + 2 allocs for []rune.
//
// These benchmarks document the current allocation profile and will regress
// (appropriately) when the v1.x 0-alloc optimisation is shipped.

package fuzzymatch_test

import (
	"strings"
	"testing"

	"github.com/axonops/fuzzymatch"
)

// BenchmarkDamerauLevenshteinFullScore_ASCII_Short exercises DL-Full on a
// short ASCII pair (2 and 2 bytes — the canonical transposition pair).
// v1.0: allocates the full (4×4) DP table — 16 ints on heap.
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
// 50-char ASCII inputs (within maxStackInputLen=64). v1.0: allocates the
// full (52×52) DP table — 2704 ints on heap.
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
