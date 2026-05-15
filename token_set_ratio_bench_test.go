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

// token_set_ratio_bench_test.go runs allocation-aware benchmarks for
// the TokenSetRatioScore public surface at multiple input sizes,
// PLUS the LOCKED pathological-input fixture per 06-CONTEXT.md §5
// (BenchmarkTokenSetRatio_Pathological_AsymmetricSetCardinalities).
// b.ReportAllocs() on every benchmark gates allocation regressions in
// bench.txt via benchstat (Phase 6 finalisation in plan 06-06).
//
// Performance budget per .claude/skills/performance-standards
// (inherited from Phase 2/3/4/5):
//
//   - Token Sort / Set / Jaccard Ratio: < 5 µs per call, ≤ 4
//     allocations on ASCII Short. TokenSetRatio's three-way max
//     construction has higher allocation overhead than TokenSortRatio
//     due to the dedup-set + three joined-string + multi-DP-row
//     pattern: realistic ceiling on ASCII Short is ≤ 10 allocs/op
//     (two Tokenise outputs + 2 sets + 3 dedup-tracking sets + 3
//     joined strings + 2 combined strings + DP rows when applicable).
//   - Pathological_AsymmetricSetCardinalities budget: < 50 µs per call
//     per CONTEXT §5 LOCKED. Numbers commit to bench.txt in plan
//     06-06.
//
// `var sink` outside the loop + a `sink < 0` gate after the loop
// prevents compiler dead-code elimination (locked Phase 2 pattern).

package fuzzymatch_test

import (
	"strings"
	"testing"

	"github.com/axonops/fuzzymatch"
)

// tokenSetA50 / B50 are 50-character ASCII strings used by the
// medium-length benchmark. Six-token inputs (typical identifier-like
// content) on each side with the intersection being four tokens and
// each diff being two tokens — exercises the three-way max code path
// rather than the subset short-circuit.
const (
	tokenSetA50 = "alpha beta gamma delta epsilon zeta eta theta iota"
	tokenSetB50 = "alpha beta gamma delta lambda mu nu xi omicron"
)

// BenchmarkTokenSetRatioScore_ASCII_Short exercises a small-input
// three-way-max case where r3 dominates (combined-vs-combined wins).
// The inputs deliberately bypass the subset short-circuit so the
// indelRatio calls all run.
func BenchmarkTokenSetRatioScore_ASCII_Short(b *testing.B) {
	b.ReportAllocs()
	b.ResetTimer()
	var sink float64
	for i := 0; i < b.N; i++ {
		sink = fuzzymatch.TokenSetRatioScore("hello world", "world peace")
	}
	if sink < 0 {
		b.Fatal("sink unexpectedly negative — compiler folded the benchmark away")
	}
}

// BenchmarkTokenSetRatioScore_ASCII_Medium exercises 50-char inputs
// with nine tokens per side, four shared and five diff (two diff_ab,
// five diff_ba). The combined strings exceed maxStackInputLen (= 64)
// so DP rows may allocate on the heap.
func BenchmarkTokenSetRatioScore_ASCII_Medium(b *testing.B) {
	b.ReportAllocs()
	b.ResetTimer()
	var sink float64
	for i := 0; i < b.N; i++ {
		sink = fuzzymatch.TokenSetRatioScore(tokenSetA50, tokenSetB50)
	}
	if sink < 0 {
		b.Fatal("sink unexpectedly negative — compiler folded the benchmark away")
	}
}

// BenchmarkTokenSetRatioScore_ASCII_Long exercises ~200-char inputs
// (20 tokens per side with a small intersection).
func BenchmarkTokenSetRatioScore_ASCII_Long(b *testing.B) {
	aLong := strings.Repeat("alpha beta gamma delta ", 10)   // 230 bytes, 40 tokens dedup→4
	bLong := strings.Repeat("alpha beta delta omicron ", 10) // similar
	b.ReportAllocs()
	b.ResetTimer()
	var sink float64
	for i := 0; i < b.N; i++ {
		sink = fuzzymatch.TokenSetRatioScore(aLong, bLong)
	}
	if sink < 0 {
		b.Fatal("sink unexpectedly negative — compiler folded the benchmark away")
	}
}

// BenchmarkTokenSetRatioScore_Unicode_Short exercises a multi-byte
// UTF-8 pair. Tokenise allocates a []rune internally per side; the
// LCS DP still operates on byte slices of the joined strings.
func BenchmarkTokenSetRatioScore_Unicode_Short(b *testing.B) {
	b.ReportAllocs()
	b.ResetTimer()
	var sink float64
	for i := 0; i < b.N; i++ {
		sink = fuzzymatch.TokenSetRatioScore("café société", "société café")
	}
	if sink < 0 {
		b.Fatal("sink unexpectedly negative — compiler folded the benchmark away")
	}
}

// BenchmarkTokenSetRatio_Pathological_AsymmetricSetCardinalities is
// the LOCKED DoS-vector fixture per 06-CONTEXT.md §5 LOCKED — a
// 5-token vs 100-token input pair with a 2-token shared core, driving
// the combined-vs-combined indelRatio call to perform ~10^4 DP cell
// updates.
//
// The expected runtime budget on developer hardware is < 50 µs / ≤ 30
// allocs per call. Numbers baseline into bench.txt during plan 06-06
// finalisation; benchstat regression detection > 10% fails CI for
// future PRs.
//
// Consumers in untrusted-input contexts (HTTP request body, file
// uploads, user-submitted identifiers) MUST pre-validate token-count
// ceilings before calling TokenSetRatioScore — see the algorithm's
// DoS-notice godoc block for the recommended pattern.
//
// The input shape: side A = 5 tokens (2 shared with B, 3 unique to
// A); side B = 100 tokens (2 shared with A, 98 unique to B). The
// asymmetry maximises the diff_ba cardinality, which dominates
// combined2to1's length (~ 100·6 chars = 600 bytes including
// separators) and drives the third indelRatio call's O(N²) cost.
func BenchmarkTokenSetRatio_Pathological_AsymmetricSetCardinalities(b *testing.B) {
	// Build the pathological inputs once outside the timed loop.
	aTokens := []string{"shared1", "shared2", "uniqueA1", "uniqueA2", "uniqueA3"}
	bTokens := make([]string, 0, 100)
	bTokens = append(bTokens, "shared1", "shared2")
	// 98 unique tokens for side B; each is a 6-char identifier so the
	// combined string lands around 600 bytes.
	for i := 0; i < 98; i++ {
		bTokens = append(bTokens, "uniqB"+stringFromIndex(i))
	}
	a := strings.Join(aTokens, " ")
	bInput := strings.Join(bTokens, " ")

	b.ReportAllocs()
	b.ResetTimer()
	var sink float64
	for i := 0; i < b.N; i++ {
		sink = fuzzymatch.TokenSetRatioScore(a, bInput)
	}
	if sink < 0 {
		b.Fatal("sink unexpectedly negative — compiler folded the benchmark away")
	}
}

// stringFromIndex returns a 2-3 character ASCII suffix derived from
// an integer index, used to build unique tokens for the pathological
// benchmark fixture. The implementation is decimal-conversion-free
// (avoids the strconv.Itoa allocation in benchmark setup) — the
// closed-form chr(a)+chr(b) over a fixed alphabet is enough for the
// 0..99 index range used here.
func stringFromIndex(i int) string {
	const alpha = "abcdefghijklmnopqrstuvwxyz"
	return string(alpha[i/26%26]) + string(alpha[i%26])
}
