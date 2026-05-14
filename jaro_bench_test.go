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

// jaro_bench_test.go runs allocation-aware benchmarks for JaroScore at four
// input sizes. b.ReportAllocs() on every benchmark gates allocation regressions
// in bench.txt via benchstat.
//
// Performance budgets (PERF-01, PERF-02 per docs/requirements.md §14):
//
//   - ASCII <= 256 bytes (stack path): 0 allocs/op
//     (the [256]bool match-flag arrays stay on the stack for ASCII inputs
//     up to maxJaroStackLen=256)
//   - ASCII > 256 bytes (heap path): 2 allocs/op (two make([]bool,...) calls)
//   - Unicode (rune path): 2 allocs/op (the two []rune conversions are
//     expected and documented per Pattern 8 in the phase patterns)

package fuzzymatch_test

import (
	"strings"
	"testing"

	"github.com/axonops/fuzzymatch"
)

// BenchmarkJaroScore_ASCII_Short exercises the zero-alloc fast path for a
// short ASCII pair (MARTHA / MARHTA — 6 bytes each). These are the Winkler
// 1990 canonical reference pair and fit well within maxJaroStackLen=256.
// Target: 0 allocs/op, 0 B/op.
func BenchmarkJaroScore_ASCII_Short(b *testing.B) {
	b.ReportAllocs()
	b.ResetTimer()
	var sink float64
	for i := 0; i < b.N; i++ {
		sink = fuzzymatch.JaroScore("MARTHA", "MARHTA")
	}
	if sink < 0 {
		b.Fatal("sink unexpectedly negative — compiler folded the benchmark away")
	}
}

// BenchmarkJaroScore_ASCII_Medium exercises the zero-alloc stack path at 50-
// char inputs (still within maxJaroStackLen=256). Target: 0 allocs/op, 0 B/op.
func BenchmarkJaroScore_ASCII_Medium(b *testing.B) {
	const aM = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWX"
	const bM = "bcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXY"
	b.ReportAllocs()
	b.ResetTimer()
	var sink float64
	for i := 0; i < b.N; i++ {
		sink = fuzzymatch.JaroScore(aM, bM)
	}
	if sink < 0 {
		b.Fatal("sink unexpectedly negative — compiler folded the benchmark away")
	}
}

// BenchmarkJaroScore_ASCII_Long exercises the heap path for inputs exceeding
// maxJaroStackLen=256. Two make([]bool, n) allocations are expected. This
// benchmark is informational for relative throughput; the alloc count (2) is
// expected and documented.
func BenchmarkJaroScore_ASCII_Long(b *testing.B) {
	// Two 300-byte ASCII strings — exceeds maxJaroStackLen=256.
	aLong := strings.Repeat("abcde", 60)
	bLong := strings.Repeat("bcdef", 60)
	b.ReportAllocs()
	b.ResetTimer()
	var sink float64
	for i := 0; i < b.N; i++ {
		sink = fuzzymatch.JaroScore(aLong, bLong)
	}
	if sink < 0 {
		b.Fatal("sink unexpectedly negative — compiler folded the benchmark away")
	}
}

// BenchmarkJaroScore_Unicode_Short exercises the rune path on a short multi-
// byte UTF-8 pair. Expects 2 allocs/op (the two []rune conversions are
// documented and expected per Pattern 8).
func BenchmarkJaroScore_Unicode_Short(b *testing.B) {
	b.ReportAllocs()
	b.ResetTimer()
	var sink float64
	for i := 0; i < b.N; i++ {
		sink = fuzzymatch.JaroScoreRunes("café", "cafe")
	}
	if sink < 0 {
		b.Fatal("sink unexpectedly negative — compiler folded the benchmark away")
	}
}
