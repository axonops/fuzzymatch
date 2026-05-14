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

// damerau_osa_bench_test.go runs allocation-aware benchmarks for
// DamerauLevenshteinOSAScore at four input sizes. b.ReportAllocs() on
// every benchmark gates allocation regressions in bench.txt via benchstat.
//
// Performance budgets (PERF-01) per .claude/skills/performance-standards:
//
//   - ASCII <= 64 bytes (stack path): target < 2 µs/op, 0 allocs/op
//     Stack buffer = (maxStackInputLen+1)*3 = 195 ints = 1560 bytes.
//   - ASCII > 64 bytes (heap path): proportional, 3 allocs/op
//     (three make([]int, n+1) heap allocations)
//   - Unicode short (rune path): target < 2 µs/op, 2 allocs/op
//     (two []rune conversions are expected and documented — Pattern 8)

package fuzzymatch_test

import (
	"strings"
	"testing"

	"github.com/axonops/fuzzymatch"
)

// BenchmarkDamerauLevenshteinOSAScore_ASCII_Short exercises the zero-alloc
// fast path for a short ASCII pair (2 and 2 bytes — the canonical transposition
// pair from Boytsov 2011). Target: 0 allocs/op.
func BenchmarkDamerauLevenshteinOSAScore_ASCII_Short(b *testing.B) {
	b.ReportAllocs()
	b.ResetTimer()
	var sink float64
	for i := 0; i < b.N; i++ {
		sink = fuzzymatch.DamerauLevenshteinOSAScore("ab", "ba")
	}
	if sink < 0 {
		b.Fatal("sink unexpectedly negative — compiler folded the benchmark away")
	}
}

// BenchmarkDamerauLevenshteinOSAScore_ASCII_Medium exercises the stack-buffer
// fast path at 50-char inputs (still within maxStackInputLen=64). Target: 0 allocs/op.
func BenchmarkDamerauLevenshteinOSAScore_ASCII_Medium(b *testing.B) {
	b.ReportAllocs()
	b.ResetTimer()
	var sink float64
	for i := 0; i < b.N; i++ {
		sink = fuzzymatch.DamerauLevenshteinOSAScore(a50, b50)
	}
	if sink < 0 {
		b.Fatal("sink unexpectedly negative — compiler folded the benchmark away")
	}
}

// BenchmarkDamerauLevenshteinOSAScore_ASCII_Long exercises the heap path
// (inputs > 64 bytes). Three make([]int, n+1) allocations are expected.
// This benchmark is informational for relative throughput.
func BenchmarkDamerauLevenshteinOSAScore_ASCII_Long(b *testing.B) {
	// Two 500-char ASCII strings.
	aLong := strings.Repeat("abcde", 100)
	bLong := strings.Repeat("bcdef", 100)
	b.ReportAllocs()
	b.ResetTimer()
	var sink float64
	for i := 0; i < b.N; i++ {
		sink = fuzzymatch.DamerauLevenshteinOSAScore(aLong, bLong)
	}
	if sink < 0 {
		b.Fatal("sink unexpectedly negative — compiler folded the benchmark away")
	}
}

// BenchmarkDamerauLevenshteinOSAScore_Unicode_Short exercises the rune path on
// a short multi-byte UTF-8 pair. Expects 2 allocs/op (the two []rune
// conversions are documented and expected per Pattern 8 in the phase patterns).
func BenchmarkDamerauLevenshteinOSAScore_Unicode_Short(b *testing.B) {
	b.ReportAllocs()
	b.ResetTimer()
	var sink float64
	for i := 0; i < b.N; i++ {
		sink = fuzzymatch.DamerauLevenshteinOSAScoreRunes("café", "cafe")
	}
	if sink < 0 {
		b.Fatal("sink unexpectedly negative — compiler folded the benchmark away")
	}
}
