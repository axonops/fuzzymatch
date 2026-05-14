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

// lcsstr_bench_test.go runs allocation-aware benchmarks for the four LCSStr
// public surfaces at multiple input sizes. b.ReportAllocs() on every
// benchmark gates allocation regressions in bench.txt via benchstat.
//
// Performance budgets per .claude/skills/performance-standards (inherited
// from Phase 2):
//
//   - ASCII Short (stack path, ≤ 64 bytes):  0 allocs/op (LCSStrScore)
//   - ASCII Medium (~50 bytes, stack path):  0 allocs/op (LCSStrScore)
//   - ASCII Long (> 64 bytes, heap path):    2 allocs/op (two rolling rows)
//   - Unicode Short (rune path):             4 allocs/op (2 []rune + 2 rows)
//
// LongestCommonSubstring is similarly budgeted but additionally returns a
// substring slice header — the underlying []byte storage is shared with the
// input string so there is no copy allocation on the byte path.
//
// `var sink` outside the loop + a `len(sink) < 0` (or `sink < 0`) gate after
// the loop prevents compiler dead-code elimination (locked Phase 2 pattern).

package fuzzymatch_test

import (
	"strings"
	"testing"

	"github.com/axonops/fuzzymatch"
)

// lcsstrA50 and lcsstrB50 are 50-byte ASCII strings used for the ASCII_Medium
// benchmarks. Both fit within maxStackInputLen=64 so the stack buffer is used.
// Prefixed to keep grep/codeflow readable.
const (
	lcsstrA50 = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWX"
	lcsstrB50 = "bcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXY"
)

// BenchmarkLCSStrScore_ASCII_Short exercises the zero-alloc fast path on a
// short ASCII pair (well within maxStackInputLen=64). Target: 0 allocs/op.
func BenchmarkLCSStrScore_ASCII_Short(b *testing.B) {
	b.ReportAllocs()
	b.ResetTimer()
	var sink float64
	for i := 0; i < b.N; i++ {
		sink = fuzzymatch.LCSStrScore("kitten", "sitting")
	}
	if sink < 0 {
		b.Fatal("sink unexpectedly negative — compiler folded the benchmark away")
	}
}

// BenchmarkLCSStrScore_ASCII_Medium exercises the stack-buffer fast path at
// 50-char inputs (within maxStackInputLen=64). Target: 0 allocs/op.
func BenchmarkLCSStrScore_ASCII_Medium(b *testing.B) {
	b.ReportAllocs()
	b.ResetTimer()
	var sink float64
	for i := 0; i < b.N; i++ {
		sink = fuzzymatch.LCSStrScore(lcsstrA50, lcsstrB50)
	}
	if sink < 0 {
		b.Fatal("sink unexpectedly negative — compiler folded the benchmark away")
	}
}

// BenchmarkLCSStrScore_ASCII_Long exercises the heap path (inputs > 64
// bytes). Two make([]int, n+1) allocations are expected (one per rolling
// row). This benchmark is informational for relative throughput; the alloc
// count (2) is expected.
func BenchmarkLCSStrScore_ASCII_Long(b *testing.B) {
	aLong := strings.Repeat("abcde", 100)
	bLong := strings.Repeat("bcdef", 100)
	b.ReportAllocs()
	b.ResetTimer()
	var sink float64
	for i := 0; i < b.N; i++ {
		sink = fuzzymatch.LCSStrScore(aLong, bLong)
	}
	if sink < 0 {
		b.Fatal("sink unexpectedly negative — compiler folded the benchmark away")
	}
}

// BenchmarkLCSStrScore_Unicode_Short exercises the byte path on a multi-byte
// UTF-8 pair. The isASCII gate gates this OFF the stack-fast path even when
// the byte length fits — so 2 allocs/op are expected for the rolling rows.
// This benchmark is informational.
func BenchmarkLCSStrScore_Unicode_Short(b *testing.B) {
	b.ReportAllocs()
	b.ResetTimer()
	var sink float64
	for i := 0; i < b.N; i++ {
		sink = fuzzymatch.LCSStrScore("café", "cafe")
	}
	if sink < 0 {
		b.Fatal("sink unexpectedly negative — compiler folded the benchmark away")
	}
}

// BenchmarkLongestCommonSubstring_ASCII_Short exercises the zero-alloc fast
// path on a short ASCII pair for the substring-returning surface. Target:
// 0 allocs/op (the returned substring is a slice header into the input
// string's backing []byte — no copy).
func BenchmarkLongestCommonSubstring_ASCII_Short(b *testing.B) {
	b.ReportAllocs()
	b.ResetTimer()
	var sink string
	for i := 0; i < b.N; i++ {
		sink = fuzzymatch.LongestCommonSubstring("kitten", "sitting")
	}
	if len(sink) > 100 {
		b.Fatal("sink unexpectedly long — compiler folded the benchmark away")
	}
}

// BenchmarkLongestCommonSubstring_ASCII_Medium: 50-char ASCII inputs, stack
// path. Target: 0 allocs/op.
func BenchmarkLongestCommonSubstring_ASCII_Medium(b *testing.B) {
	b.ReportAllocs()
	b.ResetTimer()
	var sink string
	for i := 0; i < b.N; i++ {
		sink = fuzzymatch.LongestCommonSubstring(lcsstrA50, lcsstrB50)
	}
	if len(sink) > 100 {
		b.Fatal("sink unexpectedly long — compiler folded the benchmark away")
	}
}

// BenchmarkLongestCommonSubstring_ASCII_Long: > 64-byte ASCII inputs, heap
// path. Two allocs expected (rolling rows).
func BenchmarkLongestCommonSubstring_ASCII_Long(b *testing.B) {
	aLong := strings.Repeat("abcde", 100)
	bLong := strings.Repeat("bcdef", 100)
	b.ReportAllocs()
	b.ResetTimer()
	var sink string
	for i := 0; i < b.N; i++ {
		sink = fuzzymatch.LongestCommonSubstring(aLong, bLong)
	}
	if len(sink) > 1000 {
		b.Fatal("sink unexpectedly long — compiler folded the benchmark away")
	}
}

// BenchmarkLongestCommonSubstring_Unicode_Short: multi-byte UTF-8 byte path
// (non-ASCII gate hits, so heap rows used). Informational.
func BenchmarkLongestCommonSubstring_Unicode_Short(b *testing.B) {
	b.ReportAllocs()
	b.ResetTimer()
	var sink string
	for i := 0; i < b.N; i++ {
		sink = fuzzymatch.LongestCommonSubstring("café", "cafe")
	}
	if len(sink) > 100 {
		b.Fatal("sink unexpectedly long — compiler folded the benchmark away")
	}
}

// BenchmarkLCSStrScoreRunes_Unicode_Short exercises the rune path. Expects 4
// allocs/op (2 []rune + 2 rolling rows) — the rune path has no stack fast
// path. Informational.
func BenchmarkLCSStrScoreRunes_Unicode_Short(b *testing.B) {
	b.ReportAllocs()
	b.ResetTimer()
	var sink float64
	for i := 0; i < b.N; i++ {
		sink = fuzzymatch.LCSStrScoreRunes("café", "cafe")
	}
	if sink < 0 {
		b.Fatal("sink unexpectedly negative — compiler folded the benchmark away")
	}
}

// BenchmarkLongestCommonSubstringRunes_Unicode_Short exercises the
// substring-returning rune path. Expects 4 allocs/op for the DP plus an
// additional `string(rune-slice)` conversion at the end — so 5 allocs/op
// total when a non-empty substring is returned. Informational.
func BenchmarkLongestCommonSubstringRunes_Unicode_Short(b *testing.B) {
	b.ReportAllocs()
	b.ResetTimer()
	var sink string
	for i := 0; i < b.N; i++ {
		sink = fuzzymatch.LongestCommonSubstringRunes("café", "cafe")
	}
	if len(sink) > 100 {
		b.Fatal("sink unexpectedly long — compiler folded the benchmark away")
	}
}
