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

// soundex_bench_test.go runs allocation-aware benchmarks for SoundexCode
// and SoundexScore at multiple input sizes. b.ReportAllocs() on every
// benchmark gates allocation regressions via benchstat.
//
// Performance budget per .claude/skills/performance-standards:
//
//	SoundexCode: < 500 ns, 0 allocations (stack-allocated [4]byte result)
//	SoundexScore: < 500 ns, 0 allocations (two SoundexCode calls + compare)
//
// The `var sink string/float64` pattern outside the loop + the
// `sink == ""`/`sink < 0` gate prevents compiler dead-code elimination
// (locked Phase 2 pattern).
//
// No pathological fixtures: Soundex is O(n) on bounded inputs (typical
// name length < 50 chars); no DoS-class behaviour exists.

package fuzzymatch_test

import (
	"testing"

	"github.com/axonops/fuzzymatch"
)

// BenchmarkSoundexCode_ASCII_Short exercises the canonical Robert/Rupert
// reference pair. This is the load-bearing allocation budget benchmark —
// the [4]byte stack buffer must produce 0 allocs/op.
func BenchmarkSoundexCode_ASCII_Short(b *testing.B) {
	b.ReportAllocs()
	b.ResetTimer()
	var sink string
	for i := 0; i < b.N; i++ {
		sink = fuzzymatch.SoundexCode("Robert")
	}
	if sink == "" {
		b.Fatal("sink unexpectedly empty — compiler folded the benchmark away")
	}
}

// BenchmarkSoundexCode_ASCII_Medium exercises a ~20-char name.
func BenchmarkSoundexCode_ASCII_Medium(b *testing.B) {
	b.ReportAllocs()
	b.ResetTimer()
	var sink string
	for i := 0; i < b.N; i++ {
		sink = fuzzymatch.SoundexCode("Przybyszewski")
	}
	if sink == "" {
		b.Fatal("sink unexpectedly empty — compiler folded the benchmark away")
	}
}

// BenchmarkSoundexCode_ASCII_Long exercises a ~50-char input.
func BenchmarkSoundexCode_ASCII_Long(b *testing.B) {
	b.ReportAllocs()
	b.ResetTimer()
	var sink string
	for i := 0; i < b.N; i++ {
		sink = fuzzymatch.SoundexCode("Bartholomew Theophrastus Von Hohenstaufen Jr")
	}
	if sink == "" {
		b.Fatal("sink unexpectedly empty — compiler folded the benchmark away")
	}
}

// BenchmarkSoundexScore_ASCII_Short exercises the canonical Robert/Rupert
// match pair. Both arguments produce R163 → score 1.0.
func BenchmarkSoundexScore_ASCII_Short(b *testing.B) {
	b.ReportAllocs()
	b.ResetTimer()
	var sink float64
	for i := 0; i < b.N; i++ {
		sink = fuzzymatch.SoundexScore("Robert", "Rupert")
	}
	if sink < 0 {
		b.Fatal("sink unexpectedly negative — compiler folded the benchmark away")
	}
}

// BenchmarkSoundexScore_ASCII_Identity exercises the identity short-circuit
// path (a == b → immediate 1.0, no SoundexCode calls).
func BenchmarkSoundexScore_ASCII_Identity(b *testing.B) {
	b.ReportAllocs()
	b.ResetTimer()
	var sink float64
	for i := 0; i < b.N; i++ {
		sink = fuzzymatch.SoundexScore("Robert", "Robert")
	}
	if sink < 0 {
		b.Fatal("sink unexpectedly negative — compiler folded the benchmark away")
	}
}
