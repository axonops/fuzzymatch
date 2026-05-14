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

// swg_bench_test.go runs allocation-aware benchmarks for SmithWatermanGotohScore
// at six input sizes. b.ReportAllocs() on every benchmark gates allocation
// regressions in bench.txt via benchstat.
//
// Performance budgets per .claude/skills/performance-standards:
//
//   - ASCII <= 64 bytes (stack path):       target < 2 µs/op, 0 allocs/op
//   - ASCII Medium (50 bytes, stack path):  target 0 allocs/op
//   - ASCII Long (> 64 bytes, heap path):   6 allocs/op (six float64 row slices)
//   - Unicode Short (rune path):            8 allocs/op MINIMUM (2 []rune + 6
//     row slices) — there is no stack fast path for the rune path, so 8 is the
//     floor for ANY rune input size, not just short.
//   - WithParams ASCII Short:               same as ASCII Short with custom params
//   - RawScore ASCII Short:                 exercises the unclamped path
//
// The four Short benches (Short, Medium, WithParams_Short, RawScore_Short)
// are 0-alloc gates; the Long and Unicode_Short benches are informational
// (their alloc counts are documented and expected).

package fuzzymatch_test

import (
	"strings"
	"testing"

	"github.com/axonops/fuzzymatch"
)

// swgA50 and swgB50 are 50-byte ASCII strings used for the ASCII_Medium
// benchmark. Both fit within maxStackInputLen=64 so the stack buffer is used.
//
// These intentionally do NOT collide with levenshtein_bench_test.go's a50/b50
// — Go allows package-level constants with distinct names to coexist; we use
// swg-prefixed names to keep grep/codeflow readable.
const (
	swgA50 = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWX"
	swgB50 = "bcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXY"
)

// BenchmarkSmithWatermanGotohScore_ASCII_Short exercises the zero-alloc fast
// path on a short ASCII pair (well within maxStackInputLen=64).
// Target: < 2 µs/op, 0 allocs/op.
func BenchmarkSmithWatermanGotohScore_ASCII_Short(b *testing.B) {
	b.ReportAllocs()
	b.ResetTimer()
	var sink float64
	for i := 0; i < b.N; i++ {
		sink = fuzzymatch.SmithWatermanGotohScore("kitten", "sitting")
	}
	if sink < 0 {
		b.Fatal("sink unexpectedly negative — compiler folded the benchmark away")
	}
}

// BenchmarkSmithWatermanGotohScore_ASCII_Medium exercises the stack-buffer
// fast path at 50-char inputs (within maxStackInputLen=64). Target: 0 allocs/op.
func BenchmarkSmithWatermanGotohScore_ASCII_Medium(b *testing.B) {
	b.ReportAllocs()
	b.ResetTimer()
	var sink float64
	for i := 0; i < b.N; i++ {
		sink = fuzzymatch.SmithWatermanGotohScore(swgA50, swgB50)
	}
	if sink < 0 {
		b.Fatal("sink unexpectedly negative — compiler folded the benchmark away")
	}
}

// BenchmarkSmithWatermanGotohScore_ASCII_Long exercises the heap path (inputs
// > 64 bytes). Six make([]float64, n+1) allocations are expected (one per
// rolling row). This benchmark is informational for relative throughput; the
// alloc count (6) is expected.
func BenchmarkSmithWatermanGotohScore_ASCII_Long(b *testing.B) {
	// Two 500-char ASCII strings (heap path).
	aLong := strings.Repeat("abcde", 100)
	bLong := strings.Repeat("bcdef", 100)
	b.ReportAllocs()
	b.ResetTimer()
	var sink float64
	for i := 0; i < b.N; i++ {
		sink = fuzzymatch.SmithWatermanGotohScore(aLong, bLong)
	}
	if sink < 0 {
		b.Fatal("sink unexpectedly negative — compiler folded the benchmark away")
	}
}

// BenchmarkSmithWatermanGotohScore_Unicode_Short exercises the rune path on a
// short multi-byte UTF-8 pair. Expects 8 allocs/op as a floor (2 []rune + 6 row
// slices) — the rune path has no stack fast path, so 8 is the minimum at any
// rune input size, not just "short". This benchmark is informational.
func BenchmarkSmithWatermanGotohScore_Unicode_Short(b *testing.B) {
	b.ReportAllocs()
	b.ResetTimer()
	var sink float64
	for i := 0; i < b.N; i++ {
		sink = fuzzymatch.SmithWatermanGotohScoreRunes("café", "cafe")
	}
	if sink < 0 {
		b.Fatal("sink unexpectedly negative — compiler folded the benchmark away")
	}
}

// BenchmarkSmithWatermanGotohScore_WithParams_ASCII_Short exercises the
// custom-params entry point on the ASCII Short pair. Target: 0 allocs/op.
// params is declared OUTSIDE the loop (per the locked Phase 2 pattern) so the
// struct construction does not enter the timed region.
func BenchmarkSmithWatermanGotohScore_WithParams_ASCII_Short(b *testing.B) {
	params := fuzzymatch.NewSWGParams()
	params.Match = 2.0
	b.ReportAllocs()
	b.ResetTimer()
	var sink float64
	for i := 0; i < b.N; i++ {
		sink = fuzzymatch.SmithWatermanGotohScoreWithParams("kitten", "sitting", params)
	}
	if sink < 0 {
		b.Fatal("sink unexpectedly negative — compiler folded the benchmark away")
	}
}

// BenchmarkSmithWatermanGotohRawScore_ASCII_Short exercises the unclamped
// raw path on the ASCII Short pair. Target: 0 allocs/op (no clamp/divide).
func BenchmarkSmithWatermanGotohRawScore_ASCII_Short(b *testing.B) {
	b.ReportAllocs()
	b.ResetTimer()
	var sink float64
	for i := 0; i < b.N; i++ {
		sink = fuzzymatch.SmithWatermanGotohRawScore("kitten", "sitting")
	}
	if sink < 0 {
		b.Fatal("sink unexpectedly negative — compiler folded the benchmark away")
	}
}
