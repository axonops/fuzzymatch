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

// cosine_bench_test.go runs allocation-aware benchmarks for the two
// Cosine public surfaces at multiple input sizes. b.ReportAllocs() on
// every benchmark gates allocation regressions in bench.txt via
// benchstat (Phase 5 finalisation in plan 05-05).
//
// Performance budget per RESEARCH.md §4.1 + .claude/skills/performance-
// standards (Cosine = Sørensen-Dice + 1 sorted-key slice; the
// sort.Strings call itself does not allocate beyond the existing slice
// — pdqsort is in-place):
//
//   - ASCII Short  (~5 chars):    ≤ 5 allocs/op (two maps + 1 []string)
//   - ASCII Medium (~50 chars):   ≤ 7 allocs/op (two maps + map growth + 1 []string)
//   - ASCII Long   (~200 chars):  ≤ 9 allocs/op (two maps + more growth + 1 []string)
//   - Unicode Short (rune path):  ≤ 7 allocs/op (two maps + 2 []rune + 1 []string)
//
// No stack-buffer fast path per RESEARCH.md §4.3 — the map allocation
// dominates regardless. Cosine's "+1" relative to Sørensen-Dice / Q-Gram
// Jaccard comes from the intersection-key []string slice that
// sort.Strings sorts in place (CONTEXT.md §3 LOCKED).
//
// `var sink` outside the loop + a `sink < 0` gate after the loop
// prevents compiler dead-code elimination (locked Phase 2 pattern).

package fuzzymatch_test

import (
	"strings"
	"testing"

	"github.com/axonops/fuzzymatch"
)

// cosineA50 / B50 are 50-character ASCII strings used by the
// medium-length benchmark. Constructed via overlapping shifts of the
// alphabet so the trigram intersection is non-trivial (most trigrams
// shared, a handful divergent — exercises the sorted-key dot-product
// loop on a meaningful intersection size).
const (
	cosineA50 = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWX"
	cosineB50 = "bcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXY"
)

// BenchmarkCosineScore_ASCII_Short exercises the byte path on the
// canonical RV-C1 hand-derivation pair (n=2). Expected ≤ 5 allocs/op
// (two extractQGrams maps + 1 intersection-key []string slice;
// capacity hint avoids growth allocations on short inputs).
func BenchmarkCosineScore_ASCII_Short(b *testing.B) {
	b.ReportAllocs()
	b.ResetTimer()
	var sink float64
	for i := 0; i < b.N; i++ {
		sink = fuzzymatch.CosineScore("abc", "abcd", 2)
	}
	if sink < 0 {
		b.Fatal("sink unexpectedly negative — compiler folded the benchmark away")
	}
}

// BenchmarkCosineScore_ASCII_Medium exercises the byte path on
// 50-char inputs (n=3 trigrams). Expected ≤ 7 allocs/op.
func BenchmarkCosineScore_ASCII_Medium(b *testing.B) {
	b.ReportAllocs()
	b.ResetTimer()
	var sink float64
	for i := 0; i < b.N; i++ {
		sink = fuzzymatch.CosineScore(cosineA50, cosineB50, 3)
	}
	if sink < 0 {
		b.Fatal("sink unexpectedly negative — compiler folded the benchmark away")
	}
}

// BenchmarkCosineScore_ASCII_Long exercises the byte path on ~200-char
// inputs (n=3 trigrams). Map growth dominates; expected ≤ 9 allocs/op.
func BenchmarkCosineScore_ASCII_Long(b *testing.B) {
	aLong := strings.Repeat("abcde", 40) // 200 chars
	bLong := strings.Repeat("bcdef", 40) // 200 chars
	b.ReportAllocs()
	b.ResetTimer()
	var sink float64
	for i := 0; i < b.N; i++ {
		sink = fuzzymatch.CosineScore(aLong, bLong, 3)
	}
	if sink < 0 {
		b.Fatal("sink unexpectedly negative — compiler folded the benchmark away")
	}
}

// BenchmarkCosineScoreRunes_Unicode_Short exercises the rune path on
// the canonical RV-C3 café/cafe pair (n=2). Expected ≤ 7 allocs/op:
// two []rune slice allocations + two extractQGramsRunes maps + 1
// intersection-key []string slice. Plus one rune-bigram string
// allocation per distinct key (small for short inputs).
func BenchmarkCosineScoreRunes_Unicode_Short(b *testing.B) {
	b.ReportAllocs()
	b.ResetTimer()
	var sink float64
	for i := 0; i < b.N; i++ {
		sink = fuzzymatch.CosineScoreRunes("café", "cafe", 2)
	}
	if sink < 0 {
		b.Fatal("sink unexpectedly negative — compiler folded the benchmark away")
	}
}
