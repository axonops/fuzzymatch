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

// mra_bench_test.go benchmarks MRACode, MRACompare, and MRAScore at three
// representative ASCII sizes. Allocation budgets per performance-standards SKILL:
//   - MRACode: < 500 ns, 0 allocs/op (stack-allocated buffers)
//   - MRACompare: < 500 ns, ≤ 2 allocs/op (the two intermediate codex strings)
//   - MRAScore: < 500 ns, ≤ 2 allocs/op (wraps MRACompare)

package fuzzymatch_test

import (
	"testing"

	"github.com/axonops/fuzzymatch"
)

// Sink variables prevent dead-code elimination of benchmark results.
var mraCodeSink string
var mraMatchedSink bool
var mraSimSink int
var mraScoreSink float64

// BenchmarkMRACode_ASCII_Short benchmarks MRACode on a short name (Byrne, 5 chars).
// Budget: < 500 ns, 0 allocs/op (all stack-allocated buffers).
func BenchmarkMRACode_ASCII_Short(b *testing.B) {
	b.ReportAllocs()
	var result string
	for i := 0; i < b.N; i++ {
		result = fuzzymatch.MRACode("Byrne")
	}
	mraCodeSink = result
}

// BenchmarkMRACode_ASCII_Medium benchmarks MRACode on a medium name
// (Catherine, 9 chars — also exercises the vowel-deletion path heavily).
// Budget: < 500 ns, 0 allocs/op.
func BenchmarkMRACode_ASCII_Medium(b *testing.B) {
	b.ReportAllocs()
	var result string
	for i := 0; i < b.N; i++ {
		result = fuzzymatch.MRACode("Catherine")
	}
	mraCodeSink = result
}

// BenchmarkMRACode_ASCII_Long benchmarks MRACode on a long name
// (Kathrynoglin, 12 chars — exercises the first-3-last-3 truncation path).
// Budget: < 500 ns, 0 allocs/op.
func BenchmarkMRACode_ASCII_Long(b *testing.B) {
	b.ReportAllocs()
	var result string
	for i := 0; i < b.N; i++ {
		result = fuzzymatch.MRACode("Kathrynoglin")
	}
	mraCodeSink = result
}

// BenchmarkMRACompare_ASCII_Short benchmarks MRACompare on a short pair
// (Smith/Smyth — a typical matching pair, codex lens 4 and 5).
// Budget: < 500 ns, ≤ 2 allocs/op.
func BenchmarkMRACompare_ASCII_Short(b *testing.B) {
	b.ReportAllocs()
	var matched bool
	var sim int
	for i := 0; i < b.N; i++ {
		matched, sim = fuzzymatch.MRACompare("Smith", "Smyth")
	}
	mraMatchedSink = matched
	mraSimSink = sim
}

// BenchmarkMRACompare_ASCII_Medium benchmarks MRACompare on a medium pair
// (William/Willyam). Budget: < 500 ns, ≤ 2 allocs/op.
func BenchmarkMRACompare_ASCII_Medium(b *testing.B) {
	b.ReportAllocs()
	var matched bool
	var sim int
	for i := 0; i < b.N; i++ {
		matched, sim = fuzzymatch.MRACompare("William", "Willyam")
	}
	mraMatchedSink = matched
	mraSimSink = sim
}

// BenchmarkMRACompare_ASCII_Long benchmarks MRACompare on a long pair
// (Catherine/Katherine — two 6-char codexes from long inputs).
// Budget: < 500 ns, ≤ 2 allocs/op.
func BenchmarkMRACompare_ASCII_Long(b *testing.B) {
	b.ReportAllocs()
	var matched bool
	var sim int
	for i := 0; i < b.N; i++ {
		matched, sim = fuzzymatch.MRACompare("Catherine", "Katherine")
	}
	mraMatchedSink = matched
	mraSimSink = sim
}

// BenchmarkMRACompare_Pathological_LengthDifferenceShortcut benchmarks the
// early-exit path triggered when |len(codexA) - len(codexB)| >= 3 (NBS Tech
// Note 943 step 1 auto-mismatch). This path avoids all comparison work and
// should be cheaper than a full comparison.
// Budget: < 500 ns, ≤ 2 allocs/op.
func BenchmarkMRACompare_Pathological_LengthDifferenceShortcut(b *testing.B) {
	b.ReportAllocs()
	var matched bool
	var sim int
	for i := 0; i < b.N; i++ {
		// "Ad" → "AD" (len 2), "Kathrynoglin" → "KTHGLN" (len 6): diff=4 >= 3.
		matched, sim = fuzzymatch.MRACompare("Ad", "Kathrynoglin")
	}
	mraMatchedSink = matched
	mraSimSink = sim
}

// BenchmarkMRAScore_ASCII_Short benchmarks MRAScore on the Smith/Smyth pair.
// Budget: < 500 ns, ≤ 2 allocs/op.
func BenchmarkMRAScore_ASCII_Short(b *testing.B) {
	b.ReportAllocs()
	var result float64
	for i := 0; i < b.N; i++ {
		result = fuzzymatch.MRAScore("Smith", "Smyth")
	}
	mraScoreSink = result
}

// BenchmarkMRAScore_Match benchmarks MRAScore for a matching pair.
func BenchmarkMRAScore_Match(b *testing.B) {
	b.ReportAllocs()
	var result float64
	for i := 0; i < b.N; i++ {
		result = fuzzymatch.MRAScore("Byrne", "Boern")
	}
	mraScoreSink = result
}

// BenchmarkMRAScore_NoMatch benchmarks MRAScore for a non-matching pair.
func BenchmarkMRAScore_NoMatch(b *testing.B) {
	b.ReportAllocs()
	var result float64
	for i := 0; i < b.N; i++ {
		result = fuzzymatch.MRAScore("Ad", "Kathrynoglin")
	}
	mraScoreSink = result
}
