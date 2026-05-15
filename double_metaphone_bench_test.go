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

// double_metaphone_bench_test.go runs allocation-aware benchmarks for
// DoubleMetaphoneKeys and DoubleMetaphoneScore at multiple input sizes.
// b.ReportAllocs() on every benchmark gates allocation regressions via
// benchstat.
//
// Performance budget per .claude/skills/performance-standards:
//
//	DoubleMetaphoneKeys: < 2 µs, ≤ 2 allocations
//	DoubleMetaphoneScore: < 2 µs, ≤ 2 allocations (plus short-circuit path)
//
// No pathological fixtures: Double Metaphone is O(n) on bounded inputs
// (typical name length < 50 chars); worst-case < 2µs.
//
// The `var sinkP, sinkS string` / `var sink float64` outside the loop +
// gate-after-loop pattern prevents compiler dead-code elimination
// (locked Phase 2 pattern).

package fuzzymatch_test

import (
	"testing"

	"github.com/axonops/fuzzymatch"
)

// BenchmarkDoubleMetaphoneKeys_ASCII_Short exercises the canonical Schmidt
// reference vector — the load-bearing allocation budget benchmark.
func BenchmarkDoubleMetaphoneKeys_ASCII_Short(b *testing.B) {
	b.ReportAllocs()
	b.ResetTimer()
	var sinkP, sinkS string
	for i := 0; i < b.N; i++ {
		sinkP, sinkS = fuzzymatch.DoubleMetaphoneKeys("Schmidt")
	}
	if sinkP == "" && sinkS == "" {
		b.Fatal("sinkP and sinkS unexpectedly empty — compiler folded the benchmark away")
	}
}

// BenchmarkDoubleMetaphoneKeys_ASCII_Medium exercises a ~20-char name.
func BenchmarkDoubleMetaphoneKeys_ASCII_Medium(b *testing.B) {
	b.ReportAllocs()
	b.ResetTimer()
	var sinkP, sinkS string
	for i := 0; i < b.N; i++ {
		sinkP, sinkS = fuzzymatch.DoubleMetaphoneKeys("Bartholomew")
	}
	if sinkP == "" && sinkS == "" {
		b.Fatal("sinkP and sinkS unexpectedly empty — compiler folded the benchmark away")
	}
}

// BenchmarkDoubleMetaphoneKeys_ASCII_Long exercises a ~50-char input.
func BenchmarkDoubleMetaphoneKeys_ASCII_Long(b *testing.B) {
	b.ReportAllocs()
	b.ResetTimer()
	var sinkP, sinkS string
	for i := 0; i < b.N; i++ {
		sinkP, sinkS = fuzzymatch.DoubleMetaphoneKeys("Bartholomew Theophrastus Von Hohenstaufen")
	}
	if sinkP == "" && sinkS == "" {
		b.Fatal("sinkP and sinkS unexpectedly empty — compiler folded the benchmark away")
	}
}

// BenchmarkDoubleMetaphoneScore_ASCII_Short exercises the Schmidt/Smith
// XMT cross-match pair.
func BenchmarkDoubleMetaphoneScore_ASCII_Short(b *testing.B) {
	b.ReportAllocs()
	b.ResetTimer()
	var sink float64
	for i := 0; i < b.N; i++ {
		sink = fuzzymatch.DoubleMetaphoneScore("Schmidt", "Smith")
	}
	if sink < 0 {
		b.Fatal("sink unexpectedly negative — compiler folded the benchmark away")
	}
}

// BenchmarkDoubleMetaphoneScore_Identity exercises the identity short-circuit
// path (a == b → immediate 1.0, no DoubleMetaphoneKeys calls).
func BenchmarkDoubleMetaphoneScore_Identity(b *testing.B) {
	b.ReportAllocs()
	b.ResetTimer()
	var sink float64
	for i := 0; i < b.N; i++ {
		sink = fuzzymatch.DoubleMetaphoneScore("Schmidt", "Schmidt")
	}
	if sink < 0 {
		b.Fatal("sink unexpectedly negative — compiler folded the benchmark away")
	}
}
