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

// hamming_bench_test.go runs allocation-aware benchmarks for HammingScore
// at four input sizes. b.ReportAllocs() gates allocation regressions in
// bench.txt via benchstat.
//
// Performance budget per PERF-01 and PERF-02:
//
//   - ASCII any length:   target 0 allocs/op (Hamming is a single loop — no DP buffer)
//   - Unicode short:      2 allocs/op (two []rune slice conversions, documented)

package fuzzymatch_test

import (
	"strings"
	"testing"

	"github.com/axonops/fuzzymatch"
)

// BenchmarkHammingScore_ASCII_Short exercises the zero-alloc path on the
// Hamming 1950 canonical pair (7 bytes each). Target: 0 allocs/op.
func BenchmarkHammingScore_ASCII_Short(b *testing.B) {
	b.ReportAllocs()
	b.ResetTimer()
	var sink float64
	for i := 0; i < b.N; i++ {
		sink = fuzzymatch.HammingScore("karolin", "kathrin")
	}
	if sink < 0 {
		b.Fatal("sink unexpectedly negative — compiler folded the benchmark away")
	}
}

// BenchmarkHammingScore_ASCII_Medium exercises the zero-alloc path at 50-char
// inputs. Hamming has no DP buffer so 0 allocs is expected at any length.
func BenchmarkHammingScore_ASCII_Medium(b *testing.B) {
	// Two 50-char ASCII strings (equal length — valid Hamming pair).
	const a = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWX"
	const c = "bcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXY"
	b.ReportAllocs()
	b.ResetTimer()
	var sink float64
	for i := 0; i < b.N; i++ {
		sink = fuzzymatch.HammingScore(a, c)
	}
	if sink < 0 {
		b.Fatal("sink unexpectedly negative — compiler folded the benchmark away")
	}
}

// BenchmarkHammingScore_ASCII_Long exercises the zero-alloc path at 500-char
// inputs. Unlike DP algorithms, Hamming needs no buffer at any length.
func BenchmarkHammingScore_ASCII_Long(b *testing.B) {
	// Two 500-char ASCII strings (equal length).
	aLong := strings.Repeat("abcde", 100)
	bLong := strings.Repeat("bcdea", 100)
	b.ReportAllocs()
	b.ResetTimer()
	var sink float64
	for i := 0; i < b.N; i++ {
		sink = fuzzymatch.HammingScore(aLong, bLong)
	}
	if sink < 0 {
		b.Fatal("sink unexpectedly negative — compiler folded the benchmark away")
	}
}

// BenchmarkHammingScore_Unicode_Short exercises the rune path on a short
// multi-byte UTF-8 pair. Expects 2 allocs/op (two []rune conversions per
// Pattern 8 — documented and expected).
func BenchmarkHammingScore_Unicode_Short(b *testing.B) {
	b.ReportAllocs()
	b.ResetTimer()
	var sink float64
	for i := 0; i < b.N; i++ {
		sink = fuzzymatch.HammingScoreRunes("café", "cafè")
	}
	if sink < 0 {
		b.Fatal("sink unexpectedly negative — compiler folded the benchmark away")
	}
}
