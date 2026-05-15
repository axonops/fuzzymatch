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

// monge_elkan_bench_test.go runs allocation-aware benchmarks for both
// MongeElkanScore (asymmetric) and MongeElkanScoreSymmetric (mean of
// the two directions — 2x baseline cost). b.ReportAllocs() on every
// benchmark gates allocation regressions in bench.txt via benchstat.
//
// Performance budget per .claude/skills/performance-standards
// ("Monge-Elkan: < 10 µs (dominated by inner-metric × token-count²)").
// Token counts and inner-metric complexity dominate the timing —
// MongeElkan_Pathological_1000Tokens (the DoS-T3 fixture per
// 06-CONTEXT.md §5 LOCKED) provides the descriptive 1000×1000-token
// timing envelope; it is NOT a regression gate (no allocation ceiling
// is asserted) but it pins the worst-case timing for the DoS notice in
// monge_elkan.go's three-part godoc block.
//
// `var sink` outside the loop + a `sink < 0` gate after the loop
// prevents compiler dead-code elimination (locked Phase 2 pattern).

package fuzzymatch_test

import (
	"strings"
	"testing"

	"github.com/axonops/fuzzymatch"
)

// mongeElkanA50 / B50 are ~50-character ASCII strings used by the
// medium-length benchmark. Constructed as space-separated identifiers
// to exercise the per-token-max reduction with realistic token shapes.
const (
	mongeElkanA50 = "alpha beta gamma delta epsilon zeta eta theta"
	mongeElkanB50 = "beta gamma delta epsilon zeta eta theta iota"
)

// BenchmarkMongeElkanScore_ASCII_Short exercises the canonical
// asymmetric RV-ME1 pair with JaroWinkler inner. The 2x2 inner-metric
// matrix is the smallest non-trivial Monge-Elkan workload.
func BenchmarkMongeElkanScore_ASCII_Short(b *testing.B) {
	opts := fuzzymatch.DefaultNormalisationOptions()
	b.ReportAllocs()
	b.ResetTimer()
	var sink float64
	for i := 0; i < b.N; i++ {
		sink = fuzzymatch.MongeElkanScore("user create", "usr creating", fuzzymatch.AlgoJaroWinkler, opts)
	}
	if sink < 0 {
		b.Fatal("sink unexpectedly negative — compiler folded the benchmark away")
	}
}

// BenchmarkMongeElkanScore_ASCII_Medium exercises ~50-char inputs
// (8 tokens per side) with JaroWinkler inner. 8×8 = 64 inner-metric
// comparisons per call; the dominant cost is the JaroWinkler O(min(m,n))
// per comparison.
func BenchmarkMongeElkanScore_ASCII_Medium(b *testing.B) {
	opts := fuzzymatch.DefaultNormalisationOptions()
	b.ReportAllocs()
	b.ResetTimer()
	var sink float64
	for i := 0; i < b.N; i++ {
		sink = fuzzymatch.MongeElkanScore(mongeElkanA50, mongeElkanB50, fuzzymatch.AlgoJaroWinkler, opts)
	}
	if sink < 0 {
		b.Fatal("sink unexpectedly negative — compiler folded the benchmark away")
	}
}

// BenchmarkMongeElkanScore_ASCII_Long exercises ~200-char inputs (40
// tokens per side) with JaroWinkler inner. 40×40 = 1600 inner-metric
// comparisons per call — the typical mid-tier workload.
func BenchmarkMongeElkanScore_ASCII_Long(b *testing.B) {
	stems := []string{
		"ab", "bc", "cd", "de", "ef", "fg", "gh", "hi", "ij", "jk",
		"kl", "lm", "mn", "no", "op", "pq", "qr", "rs", "st", "tu",
		"uv", "vw", "wx", "xy", "yz", "za", "ba", "cb", "dc", "ed",
		"fe", "gf", "hg", "ih", "ji", "kj", "lk", "ml", "nm", "on",
	}
	aLong := strings.Join(stems, " ")
	bLong := strings.Join(stems[1:], " ") + " xy"
	opts := fuzzymatch.DefaultNormalisationOptions()
	b.ReportAllocs()
	b.ResetTimer()
	var sink float64
	for i := 0; i < b.N; i++ {
		sink = fuzzymatch.MongeElkanScore(aLong, bLong, fuzzymatch.AlgoJaroWinkler, opts)
	}
	if sink < 0 {
		b.Fatal("sink unexpectedly negative — compiler folded the benchmark away")
	}
}

// BenchmarkMongeElkanScore_Unicode_Short exercises a multi-byte UTF-8
// token pair with the byte-path JaroWinkler inner. Tokenise is
// UTF-8-aware; the rune semantic is preserved at the tokenisation
// layer.
func BenchmarkMongeElkanScore_Unicode_Short(b *testing.B) {
	opts := fuzzymatch.DefaultNormalisationOptions()
	b.ReportAllocs()
	b.ResetTimer()
	var sink float64
	for i := 0; i < b.N; i++ {
		sink = fuzzymatch.MongeElkanScore("café münchen", "münchen wien", fuzzymatch.AlgoJaroWinkler, opts)
	}
	if sink < 0 {
		b.Fatal("sink unexpectedly negative — compiler folded the benchmark away")
	}
}

// BenchmarkMongeElkanScoreSymmetric_ASCII_Short exercises the SYMMETRIC
// variant on the canonical RV-ME1 pair. Expected ~2x the asymmetric
// cost (the symmetric variant calls MongeElkanScore in both
// directions). benchstat (over BenchmarkMongeElkanScore_ASCII_Short)
// quantifies the overhead.
func BenchmarkMongeElkanScoreSymmetric_ASCII_Short(b *testing.B) {
	opts := fuzzymatch.DefaultNormalisationOptions()
	b.ReportAllocs()
	b.ResetTimer()
	var sink float64
	for i := 0; i < b.N; i++ {
		sink = fuzzymatch.MongeElkanScoreSymmetric("user create", "usr creating", fuzzymatch.AlgoJaroWinkler, opts)
	}
	if sink < 0 {
		b.Fatal("sink unexpectedly negative — compiler folded the benchmark away")
	}
}

// BenchmarkMongeElkan_Pathological_1000Tokens is the DoS-T3 fixture per
// 06-CONTEXT.md §5 LOCKED. 1000 tokens per side → ~10^6 inner-metric
// comparisons per call. With JaroWinkler inner (the default), each
// comparison is O(token_length) — total cost approximates 10^7
// character operations.
//
// Informational, not a regression gate: no allocation or timing
// ceiling is asserted. The bench number documents the worst-case
// timing for the DoS notice in monge_elkan.go's three-part godoc block.
// In untrusted-input contexts consumers must pre-validate token-count
// ceilings before calling — documented in the godoc DoS notice.
func BenchmarkMongeElkan_Pathological_1000Tokens(b *testing.B) {
	// Build two 1000-token strings with stable per-token shape so the
	// allocator behaviour is predictable. The trailing trim removes the
	// dangling space so Tokenise produces exactly 1000 tokens per side
	// (strings.Repeat ends with a separator).
	manyA := strings.TrimSpace(strings.Repeat("alpha beta gamma delta ", 250))
	manyB := strings.TrimSpace(strings.Repeat("epsilon zeta eta theta ", 250))
	opts := fuzzymatch.DefaultNormalisationOptions()
	b.ReportAllocs()
	b.ResetTimer()
	var sink float64
	for i := 0; i < b.N; i++ {
		sink = fuzzymatch.MongeElkanScoreSymmetric(manyA, manyB, fuzzymatch.AlgoJaroWinkler, opts)
	}
	if sink < 0 {
		b.Fatal("sink unexpectedly negative — compiler folded the benchmark away")
	}
}
