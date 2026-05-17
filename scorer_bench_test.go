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

// scorer_bench_test.go runs allocation-aware benchmarks for the Phase 8
// composite Scorer surface at multiple input sizes. b.ReportAllocs() on
// every benchmark gates allocation regressions in bench.txt via
// benchstat (Phase 8 finalisation).
//
// Performance budget per .claude/skills/performance-standards "Scorer
// Budgets" + 08-VALIDATION.md (Q8c, docs/requirements.md §14.2 — revised
// 2026-05 to match implementation reality):
//
//   - ASCII Short  (~5 chars):    < 30 µs wall, ≤ 8 allocs/op
//   - ASCII Medium (~30-50 chars):< 30 µs wall, ≤ 20 allocs/op (Q8c —
//                                 revised from 8: the DefaultScorer
//                                 composes 11 algorithms and each
//                                 q-gram-tier inner score allocates 2
//                                 maps, so 20 reflects the structural
//                                 floor of the composite)
//   - ASCII Long   (~500 chars):  informational (no budget — long
//                                 inputs exceed Phase 8 budget by
//                                 design; included for benchstat
//                                 trending)
//   - Unicode Short (multi-byte): informational
//
// Q8c rationale: the post-Q7a DoubleMetaphone optimisation (Plan 07
// Cluster 5) preserves the Short ≤ 8 floor; the Medium budget rises to
// ≤ 20 because the q-gram-tier algorithms (Jaccard, Sørensen-Dice,
// Cosine, Tversky) each allocate 2 map[string]int per call, and the
// token-tier (Token Sort/Set/Partial Ratio, Token Jaccard) adds further
// tokenisation allocations on identifier-style inputs.
//
// DefaultScorer is constructed ONCE before b.ResetTimer() so construction
// cost is not part of the per-op measurement. Score, ScoreAll, and Match
// each have a dedicated benchmark on the same short input pair so the
// three methods' relative cost is visible in a single benchstat report.
//
// `var sink` outside the loop + a `sink < 0` gate (or analogous bool/
// map sink) after the loop prevents compiler dead-code elimination
// (locked Phase 2 pattern, carry-forward from cosine_bench_test.go).
//
// If allocs/op > 8 on ASCII Short/Medium OR µs/op > 30 on ASCII Medium,
// the result is escalated to algorithm-performance-reviewer per
// VALIDATION.md "Manual-Only Verifications" #4. The budget is enforced
// manually via reviewer sign-off, not as a hard test failure (the
// benchmark is allocation-aware but does not embed a t.Fatal threshold
// — that would couple the test to a specific runner hardware).

package fuzzymatch_test

import (
	"strings"
	"testing"

	"github.com/axonops/fuzzymatch"
)

// scorerA50 / scorerB50 are ~30-char ASCII identifier-style pairs used
// by the medium-length benchmark. The pair exercises the most
// realistic Phase 8 consumer workload: snake_case-vs-camelCase
// identifier comparison (the axonops/audit use case per CONTEXT.md
// §Specific Ideas). Length stays at or below the 50-char ASCII budget
// for the < 30 µs / ≤ 8 allocs ceiling.
const (
	scorerAMedium = "customer_billing_history_2024_v2"
	scorerBMedium = "customerBillingHistory2024V2"
)

// BenchmarkDefaultScorer_Score_ASCII_Short exercises DefaultScorer.Score
// on the canonical 3-4 char pair. Expected < 30 µs wall + ≤ 8 allocs/op
// per performance-standards "Scorer Budgets". DefaultScorer is
// constructed once outside the timer.
func BenchmarkDefaultScorer_Score_ASCII_Short(b *testing.B) {
	s := fuzzymatch.DefaultScorer()
	b.ReportAllocs()
	b.ResetTimer()
	var sink float64
	for i := 0; i < b.N; i++ {
		sink = s.Score("abc", "abcd")
	}
	if sink < 0 {
		b.Fatal("sink unexpectedly negative — compiler folded the benchmark away")
	}
}

// BenchmarkDefaultScorer_Score_ASCII_Medium exercises DefaultScorer.Score
// on ~30-char identifier-style pairs (≤ 50 chars budget threshold).
// Q8c budget: < 30 µs wall + ≤ 20 allocs/op (revised from 8 to match the
// q-gram + token-tier composite floor; see file godoc).
func BenchmarkDefaultScorer_Score_ASCII_Medium(b *testing.B) {
	s := fuzzymatch.DefaultScorer()
	b.ReportAllocs()
	b.ResetTimer()
	var sink float64
	for i := 0; i < b.N; i++ {
		sink = s.Score(scorerAMedium, scorerBMedium)
	}
	if sink < 0 {
		b.Fatal("sink unexpectedly negative — compiler folded the benchmark away")
	}
}

// BenchmarkDefaultScorer_Score_ASCII_Long exercises DefaultScorer.Score
// on ~500-char inputs. INFORMATIONAL only — Phase 8 budget targets
// ASCII ≤ 50 chars; long inputs exceed the budget by design. Included
// for benchstat trending so a performance regression on long inputs is
// still visible in CI history.
func BenchmarkDefaultScorer_Score_ASCII_Long(b *testing.B) {
	aLong := strings.Repeat("ab", 250) // 500 chars
	bLong := strings.Repeat("ba", 250) // 500 chars
	s := fuzzymatch.DefaultScorer()
	b.ReportAllocs()
	b.ResetTimer()
	var sink float64
	for i := 0; i < b.N; i++ {
		sink = s.Score(aLong, bLong)
	}
	if sink < 0 {
		b.Fatal("sink unexpectedly negative — compiler folded the benchmark away")
	}
}

// BenchmarkDefaultScorer_Score_Unicode_Short exercises DefaultScorer.Score
// on a multi-byte UTF-8 short pair (rune path implicitly engaged by
// algorithms that use the rune surface). INFORMATIONAL only — Unicode
// budget is not pinned for Phase 8.
func BenchmarkDefaultScorer_Score_Unicode_Short(b *testing.B) {
	s := fuzzymatch.DefaultScorer()
	b.ReportAllocs()
	b.ResetTimer()
	var sink float64
	for i := 0; i < b.N; i++ {
		sink = s.Score("café", "cafë")
	}
	if sink < 0 {
		b.Fatal("sink unexpectedly negative — compiler folded the benchmark away")
	}
}

// BenchmarkDefaultScorer_ScoreAll_ASCII_Short exercises ScoreAll on
// the same short pair as Score. ScoreAll adds 1 allocation for the
// result map per spec §8.6 (compared to Score's reduce-to-scalar).
// INFORMATIONAL — the < 30 µs / ≤ 8 allocs budget applies to Score;
// ScoreAll's relative cost vs Score is what reviewers track.
func BenchmarkDefaultScorer_ScoreAll_ASCII_Short(b *testing.B) {
	s := fuzzymatch.DefaultScorer()
	b.ReportAllocs()
	b.ResetTimer()
	var sink map[fuzzymatch.AlgoID]float64
	for i := 0; i < b.N; i++ {
		sink = s.ScoreAll("abc", "abcd")
	}
	if sink == nil {
		b.Fatal("sink unexpectedly nil — compiler folded the benchmark away")
	}
}

// BenchmarkDefaultScorer_Match_ASCII_Short exercises Match on the same
// short pair as Score. Match delegates to Score + threshold comparison,
// so the cost should be near-identical to Score itself. The bool sink
// uses an int counter (incremented when match is true) plus a `sink <
// -1` gate to defeat compiler dead-code elimination without conditional
// branches inside the timed loop.
func BenchmarkDefaultScorer_Match_ASCII_Short(b *testing.B) {
	s := fuzzymatch.DefaultScorer()
	b.ReportAllocs()
	b.ResetTimer()
	var sink int
	for i := 0; i < b.N; i++ {
		if s.Match("abc", "abcd") {
			sink++
		}
	}
	if sink < -1 {
		b.Fatal("sink unexpectedly negative — compiler folded the benchmark away")
	}
}
