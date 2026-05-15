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

// sorensen_dice_bench_test.go runs allocation-aware benchmarks for the
// two Sørensen-Dice public surfaces at multiple input sizes.
// b.ReportAllocs() on every benchmark gates allocation regressions in
// bench.txt via benchstat (Phase 5 finalisation in plan 05-05).
//
// Performance budget per RESEARCH.md §4.1 + .claude/skills/performance-
// standards (inherited from plan 05-01 — Sørensen-Dice consumes the
// same shared q-gram extractor):
//
//   - ASCII Short  (~5 chars):    ≤ 4 allocs/op (two map allocations)
//   - ASCII Medium (~50 chars):   ≤ 6 allocs/op (two maps + map growth)
//   - ASCII Long   (~200 chars):  ≤ 8 allocs/op (two maps + more growth)
//   - Unicode Short (rune path):  ≤ 6 allocs/op (two maps + 2 []rune)
//
// No stack-buffer fast path per RESEARCH.md §4.3 — the map allocation
// dominates regardless. The 4-alloc ideal in CONTEXT.md §5 is the
// canonical-source budget; the realistic ceiling grows with input
// length per RESEARCH.md §4.1.
//
// `var sink` outside the loop + a `sink < 0` gate after the loop
// prevents compiler dead-code elimination (locked Phase 2 pattern).

package fuzzymatch_test

import (
	"strings"
	"testing"

	"github.com/axonops/fuzzymatch"
)

// sorensenDiceA50 / B50 are 50-character ASCII strings used by the
// medium-length benchmark. Constructed via overlapping shifts of the
// alphabet so the bigram intersection is non-trivial (most bigrams
// shared, a handful divergent).
const (
	sorensenDiceA50 = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWX"
	sorensenDiceB50 = "bcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXY"
)

// BenchmarkSorensenDiceScore_ASCII_Short exercises the byte path on the
// canonical NLP-textbook bigram pair (RV-D1; n=2). Expected ≤ 4
// allocs/op (two extractQGrams maps; capacity hint avoids growth
// allocations on short inputs).
func BenchmarkSorensenDiceScore_ASCII_Short(b *testing.B) {
	b.ReportAllocs()
	b.ResetTimer()
	var sink float64
	for i := 0; i < b.N; i++ {
		sink = fuzzymatch.SorensenDiceScore("night", "nacht", 2)
	}
	if sink < 0 {
		b.Fatal("sink unexpectedly negative — compiler folded the benchmark away")
	}
}

// BenchmarkSorensenDiceScore_ASCII_Medium exercises the byte path on
// 50-char inputs (n=3 trigrams). Expected ≤ 6 allocs/op.
func BenchmarkSorensenDiceScore_ASCII_Medium(b *testing.B) {
	b.ReportAllocs()
	b.ResetTimer()
	var sink float64
	for i := 0; i < b.N; i++ {
		sink = fuzzymatch.SorensenDiceScore(sorensenDiceA50, sorensenDiceB50, 3)
	}
	if sink < 0 {
		b.Fatal("sink unexpectedly negative — compiler folded the benchmark away")
	}
}

// BenchmarkSorensenDiceScore_ASCII_Long exercises the byte path on
// ~200-char inputs (n=3 trigrams). Map growth dominates; expected
// ≤ 8 allocs/op.
func BenchmarkSorensenDiceScore_ASCII_Long(b *testing.B) {
	aLong := strings.Repeat("abcde", 40) // 200 chars
	bLong := strings.Repeat("bcdef", 40) // 200 chars
	b.ReportAllocs()
	b.ResetTimer()
	var sink float64
	for i := 0; i < b.N; i++ {
		sink = fuzzymatch.SorensenDiceScore(aLong, bLong, 3)
	}
	if sink < 0 {
		b.Fatal("sink unexpectedly negative — compiler folded the benchmark away")
	}
}

// BenchmarkSorensenDiceScoreRunes_Unicode_Short exercises the rune
// path on the canonical café/cafe pair (n=2). Expected ≤ 6 allocs/op:
// two []rune slice allocations + two extractQGramsRunes maps. Plus one
// rune-bigram string allocation per distinct key (small for short
// inputs).
func BenchmarkSorensenDiceScoreRunes_Unicode_Short(b *testing.B) {
	b.ReportAllocs()
	b.ResetTimer()
	var sink float64
	for i := 0; i < b.N; i++ {
		sink = fuzzymatch.SorensenDiceScoreRunes("café", "cafe", 2)
	}
	if sink < 0 {
		b.Fatal("sink unexpectedly negative — compiler folded the benchmark away")
	}
}
