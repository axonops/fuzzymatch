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

// strcmp95_bench_test.go runs allocation-aware benchmarks for Strcmp95Score
// at three input sizes. b.ReportAllocs() on every benchmark gates allocation
// regressions in bench.txt via benchstat.
//
// Performance budgets per .claude/skills/performance-standards:
//
//   - ASCII Short (≤ 64 bytes; canonical MARTHA/MARHTA): target < 2 µs/op,
//     0 allocs/op (match-flag arrays + similar-pair consumption arena all
//     stack-allocate under maxJaroStackLen = 256).
//   - ASCII Medium (~50 bytes):                          target 0 allocs/op.
//   - ASCII Long (> maxJaroStackLen = 256 bytes):        2 allocs/op
//     (one make([]bool, la) + one make([]bool, lb); the similar-pair
//     consumption arena also heap-allocates on the > 256 path). Long is
//     informational — the alloc count is expected, not gated.
//
// NB: Strcmp95 is byte-only by CONTEXT.md §2 — there is NO Unicode_Short
// benchmark and no *Runes variant. There is NO WithParams variant (no
// Strcmp95Params surface). There is NO Raw variant (no unclamped surface).

package fuzzymatch_test

import (
	"strings"
	"testing"

	"github.com/axonops/fuzzymatch"
)

// strcmp95A50 and strcmp95B50 are 50-byte ASCII strings used for the
// ASCII_Medium benchmark. Both fit within maxJaroStackLen = 256 so the stack
// buffer is used and 0 allocs/op is achievable.
//
// These intentionally do NOT collide with other algorithms' a50/b50 in this
// package — Go allows package-level constants with distinct names to coexist;
// strcmp95-prefixed names keep grep/codeflow readable.
const (
	strcmp95A50 = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWX"
	strcmp95B50 = "bcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXY"
)

// BenchmarkStrcmp95Score_ASCII_Short exercises the zero-alloc fast path on
// the canonical Winkler 1990 reference pair (MARTHA/MARHTA — 6 and 6 bytes,
// well within maxJaroStackLen = 256). Target: < 2 µs/op, 0 allocs/op.
func BenchmarkStrcmp95Score_ASCII_Short(b *testing.B) {
	b.ReportAllocs()
	b.ResetTimer()
	var sink float64
	for i := 0; i < b.N; i++ {
		sink = fuzzymatch.Strcmp95Score("MARTHA", "MARHTA")
	}
	if sink < 0 {
		b.Fatal("sink unexpectedly negative — compiler folded the benchmark away")
	}
}

// BenchmarkStrcmp95Score_ASCII_Medium exercises the stack-buffer fast path at
// 50-char inputs (within maxJaroStackLen = 256). Target: 0 allocs/op.
func BenchmarkStrcmp95Score_ASCII_Medium(b *testing.B) {
	b.ReportAllocs()
	b.ResetTimer()
	var sink float64
	for i := 0; i < b.N; i++ {
		sink = fuzzymatch.Strcmp95Score(strcmp95A50, strcmp95B50)
	}
	if sink < 0 {
		b.Fatal("sink unexpectedly negative — compiler folded the benchmark away")
	}
}

// BenchmarkStrcmp95Score_ASCII_Long exercises the heap path (inputs > 256
// bytes). Three allocations are expected (two []bool match-flag arrays plus
// one []bool similar-pair consumption arena). This benchmark is informational
// for relative throughput; the alloc count (3) is expected.
func BenchmarkStrcmp95Score_ASCII_Long(b *testing.B) {
	// Two ~500-char ASCII strings (heap path).
	aLong := strings.Repeat("abcde", 100)
	bLong := strings.Repeat("bcdef", 100)
	b.ReportAllocs()
	b.ResetTimer()
	var sink float64
	for i := 0; i < b.N; i++ {
		sink = fuzzymatch.Strcmp95Score(aLong, bLong)
	}
	if sink < 0 {
		b.Fatal("sink unexpectedly negative — compiler folded the benchmark away")
	}
}
