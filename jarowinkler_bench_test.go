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

// jarowinkler_bench_test.go benchmarks JaroWinklerScore at the allocation
// budget checkpoints required by PERF-01 and PERF-02.
//
// Target: 0 B/op, 0 allocs/op on ASCII Short (PERF-01). JaroWinklerScore
// adds only a constant-bounded prefix loop over JaroScore; the allocation
// profile is identical to Jaro for ASCII inputs.
//
// Stdlib `testing` only — no testify in root tests.

package fuzzymatch_test

import (
	"strings"
	"testing"

	"github.com/axonops/fuzzymatch"
)

// BenchmarkJaroWinklerScore_ASCII_Short benchmarks the canonical Winkler 1990
// reference pair (MARTHA / MARHTA — 6 bytes each). This is the PERF-01 gate:
// must report 0 B/op, 0 allocs/op due to the ASCII fast path in JaroScore.
func BenchmarkJaroWinklerScore_ASCII_Short(b *testing.B) {
	b.ReportAllocs()
	b.ResetTimer()
	var sink float64
	for i := 0; i < b.N; i++ {
		sink = fuzzymatch.JaroWinklerScore("MARTHA", "MARHTA")
	}
	if sink < 0 {
		b.Fatal("sink unexpectedly negative")
	}
}

// BenchmarkJaroWinklerScore_ASCII_Medium benchmarks a 50-character ASCII
// identifier pair. Both inputs are within maxJaroStackLen (256 bytes) so
// the ASCII fast path is used — 0 allocs expected.
func BenchmarkJaroWinklerScore_ASCII_Medium(b *testing.B) {
	a := strings.Repeat("abcdefghij", 5)    // 50 bytes
	bStr := strings.Repeat("abcdeXghij", 5) // 50 bytes, differs at position 5 of each repeat
	b.ReportAllocs()
	b.ResetTimer()
	var sink float64
	for i := 0; i < b.N; i++ {
		sink = fuzzymatch.JaroWinklerScore(a, bStr)
	}
	if sink < 0 {
		b.Fatal("sink unexpectedly negative")
	}
}

// BenchmarkJaroWinklerScore_ASCII_Long benchmarks a 300-character ASCII pair.
// Inputs exceed maxJaroStackLen (256 bytes); JaroScore uses the heap path
// (make([]bool, n)). JaroWinklerScore contributes 0 extra allocs.
func BenchmarkJaroWinklerScore_ASCII_Long(b *testing.B) {
	base := "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	a := base
	for len(a) < 300 {
		a += "x"
	}
	bStr := base[1:] + "Z"
	for len(bStr) < 300 {
		bStr += "y"
	}
	b.ReportAllocs()
	b.ResetTimer()
	var sink float64
	for i := 0; i < b.N; i++ {
		sink = fuzzymatch.JaroWinklerScore(a, bStr)
	}
	if sink < 0 {
		b.Fatal("sink unexpectedly negative")
	}
}

// BenchmarkJaroWinklerScore_Unicode_Short benchmarks the rune-aware path on a
// short multi-byte UTF-8 pair. JaroWinklerScoreRunes allocates exactly two
// []rune slices (shared between the Jaro kernel and the prefix scan after the
// IN-01 consolidation). On short inputs the compiler stack-allocates the two
// slices via escape analysis, so this benchmark reports 0 B/op, 0 allocs/op.
func BenchmarkJaroWinklerScore_Unicode_Short(b *testing.B) {
	b.ReportAllocs()
	b.ResetTimer()
	var sink float64
	for i := 0; i < b.N; i++ {
		sink = fuzzymatch.JaroWinklerScoreRunes("café", "cafè")
	}
	if sink < 0 {
		b.Fatal("sink unexpectedly negative")
	}
}
