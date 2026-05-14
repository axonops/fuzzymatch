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

// levenshtein_bench_test.go runs allocation-aware benchmarks for
// LevenshteinScore at four input sizes. b.ReportAllocs() on every benchmark
// gates allocation regressions in bench.txt via benchstat.
//
// Performance budgets per .claude/skills/performance-standards:
//
//   - ASCII <= 64 bytes (stack path):  target < 1 µs/op, 0 allocs/op
//   - ASCII > 64 bytes (heap path):    proportional, <= 2 allocs/op
//   - Unicode short (rune path):       target < 2 µs/op, 2 allocs/op
//     (the two []rune slice conversions are expected and documented)

package fuzzymatch_test

import (
	"strings"
	"testing"

	"github.com/axonops/fuzzymatch"
)

// a50 and b50 are 50-byte ASCII strings used for the ASCII_Medium benchmark.
// Both fit within maxStackInputLen=64 so the stack buffer is used (0 allocs).
const (
	a50 = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWX"
	b50 = "bcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXY"
)

// BenchmarkLevenshteinScore_ASCII_Short exercises the zero-alloc fast path for
// a short ASCII pair (6 and 7 bytes — well within maxStackInputLen=64).
// Target: < 1 µs/op, 0 allocs/op. The Wagner-Fischer 1974 canonical pair.
func BenchmarkLevenshteinScore_ASCII_Short(b *testing.B) {
	b.ReportAllocs()
	b.ResetTimer()
	var sink float64
	for i := 0; i < b.N; i++ {
		sink = fuzzymatch.LevenshteinScore("kitten", "sitting")
	}
	if sink < 0 {
		b.Fatal("sink unexpectedly negative — compiler folded the benchmark away")
	}
}

// BenchmarkLevenshteinScore_ASCII_Medium exercises the stack-buffer fast path
// at 50-char inputs (still within maxStackInputLen=64). Target: 0 allocs/op.
func BenchmarkLevenshteinScore_ASCII_Medium(b *testing.B) {
	b.ReportAllocs()
	b.ResetTimer()
	var sink float64
	for i := 0; i < b.N; i++ {
		sink = fuzzymatch.LevenshteinScore(a50, b50)
	}
	if sink < 0 {
		b.Fatal("sink unexpectedly negative — compiler folded the benchmark away")
	}
}

// BenchmarkLevenshteinScore_ASCII_Long exercises the heap path (inputs > 64
// bytes). Two make([]int, n+1) allocations are expected. This benchmark is
// informational for relative throughput; the alloc count (2) is expected.
func BenchmarkLevenshteinScore_ASCII_Long(b *testing.B) {
	// Two 500-char ASCII strings.
	aLong := strings.Repeat("abcde", 100)
	bLong := strings.Repeat("bcdef", 100)
	b.ReportAllocs()
	b.ResetTimer()
	var sink float64
	for i := 0; i < b.N; i++ {
		sink = fuzzymatch.LevenshteinScore(aLong, bLong)
	}
	if sink < 0 {
		b.Fatal("sink unexpectedly negative — compiler folded the benchmark away")
	}
}

// BenchmarkLevenshteinScore_Unicode_Short exercises the rune path on a short
// multi-byte UTF-8 pair. Expects 2 allocs/op (the two []rune conversions are
// documented and expected per Pattern 8 in the phase patterns).
func BenchmarkLevenshteinScore_Unicode_Short(b *testing.B) {
	b.ReportAllocs()
	b.ResetTimer()
	var sink float64
	for i := 0; i < b.N; i++ {
		sink = fuzzymatch.LevenshteinScoreRunes("café", "cafe")
	}
	if sink < 0 {
		b.Fatal("sink unexpectedly negative — compiler folded the benchmark away")
	}
}
