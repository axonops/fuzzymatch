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

// token_jaccard_bench_test.go runs allocation-aware benchmarks for
// TokenJaccardScore at multiple input sizes. b.ReportAllocs() on every
// benchmark gates allocation regressions in bench.txt via benchstat.
//
// Performance budget per .claude/skills/performance-standards
// ("Token Sort/Set/Jaccard Ratio: < 5 µs, ≤ 4 allocations" — baseline)
// with a realistic ceiling accommodation per Phase 5's q-gram
// budget pattern:
//
//   - ASCII Short  (~10 chars):  baseline ~ 4 allocs/op (2 Tokenise + 2 maps)
//   - ASCII Medium (~50 chars):  growth ~ 6 allocs/op (Tokenise output capacity)
//   - ASCII Long   (~200 chars): growth ~ 8 allocs/op (token slice + map growth)
//   - Unicode Short:             same shape as ASCII Short (Tokenise is
//                                UTF-8-aware; the rune semantic is
//                                preserved at the tokenisation layer —
//                                there is no separate rune-path).
//
// No pathological fixture: TokenJaccard is O(|tokens|) set construction
// + O(min(|setA|, |setB|)) intersection scan — linear in input size with
// no DoS-class behaviour. Distinct from TokenSetRatio (which has the
// asymmetric-cardinality LCS DP cost driver per 06-CONTEXT.md §5) and
// PartialRatio (long-vs-short mismatch O(|s|·|l|·max)).
//
// `var sink` outside the loop + a `sink < 0` gate after the loop
// prevents compiler dead-code elimination (locked Phase 2 pattern).

package fuzzymatch_test

import (
	"strings"
	"testing"

	"github.com/axonops/fuzzymatch"
)

// tokenJaccardA50 / B50 are ~50-character ASCII strings used by the
// medium-length benchmark. Constructed as space-separated identifiers
// that produce token sets with significant but not total overlap.
const (
	tokenJaccardA50 = "alpha beta gamma delta epsilon zeta eta theta"
	tokenJaccardB50 = "beta gamma delta epsilon zeta eta theta iota"
)

// BenchmarkTokenJaccardScore_ASCII_Short exercises the canonical
// partial-overlap RV-TJ1 pair. Expected ~4 allocs/op (2 Tokenise outputs
// + 2 maps).
func BenchmarkTokenJaccardScore_ASCII_Short(b *testing.B) {
	b.ReportAllocs()
	b.ResetTimer()
	var sink float64
	for i := 0; i < b.N; i++ {
		sink = fuzzymatch.TokenJaccardScore("a b c", "b c d")
	}
	if sink < 0 {
		b.Fatal("sink unexpectedly negative — compiler folded the benchmark away")
	}
}

// BenchmarkTokenJaccardScore_ASCII_Medium exercises ~50-char
// identifier-style inputs with 7 shared tokens (one in/out on each
// side). Expected ~6-10 allocs/op with map growth and per-token string
// allocations from Tokenise's lowercase fold.
func BenchmarkTokenJaccardScore_ASCII_Medium(b *testing.B) {
	b.ReportAllocs()
	b.ResetTimer()
	var sink float64
	for i := 0; i < b.N; i++ {
		sink = fuzzymatch.TokenJaccardScore(tokenJaccardA50, tokenJaccardB50)
	}
	if sink < 0 {
		b.Fatal("sink unexpectedly negative — compiler folded the benchmark away")
	}
}

// BenchmarkTokenJaccardScore_ASCII_Long exercises ~200-char inputs (40
// shared token-pairs constructed by overlapping shifts). Map growth and
// per-token-string allocations dominate. The set-Jaccard scan stays
// O(min(|setA|, |setB|)) by construction.
func BenchmarkTokenJaccardScore_ASCII_Long(b *testing.B) {
	// Build two strings of ~200 chars where most tokens are shared.
	// 40 distinct word stems, ~5 chars each + space = ~200 chars per side.
	stems := []string{
		"ab", "bc", "cd", "de", "ef", "fg", "gh", "hi", "ij", "jk",
		"kl", "lm", "mn", "no", "op", "pq", "qr", "rs", "st", "tu",
		"uv", "vw", "wx", "xy", "yz", "za", "ba", "cb", "dc", "ed",
		"fe", "gf", "hg", "ih", "ji", "kj", "lk", "ml", "nm", "on",
	}
	aLong := strings.Join(stems, " ")            // 40 tokens
	bLong := strings.Join(stems[1:], " ") + " x" // 40 tokens, last differs
	b.ReportAllocs()
	b.ResetTimer()
	var sink float64
	for i := 0; i < b.N; i++ {
		sink = fuzzymatch.TokenJaccardScore(aLong, bLong)
	}
	if sink < 0 {
		b.Fatal("sink unexpectedly negative — compiler folded the benchmark away")
	}
}

// BenchmarkTokenJaccardScore_Unicode_Short exercises a multi-byte UTF-8
// token pair. Tokenise is UTF-8-aware; the rune semantic is preserved
// at the tokenisation layer so map[string]struct{} keys carry the full
// UTF-8 byte sequence. There is no separate rune-path variant for
// TokenJaccard.
func BenchmarkTokenJaccardScore_Unicode_Short(b *testing.B) {
	b.ReportAllocs()
	b.ResetTimer()
	var sink float64
	for i := 0; i < b.N; i++ {
		sink = fuzzymatch.TokenJaccardScore("café münchen", "münchen wien")
	}
	if sink < 0 {
		b.Fatal("sink unexpectedly negative — compiler folded the benchmark away")
	}
}
