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

// nysiis_bench_test.go benchmarks NYSIISCode and NYSIISScore at three
// representative ASCII sizes. Budget: < 500 ns, 0 allocations per
// performance-standards SKILL.

package fuzzymatch_test

import (
	"testing"

	"github.com/axonops/fuzzymatch"
)

// sink prevents dead-code elimination of benchmark results.
var nysiisSink string
var nysiisScoreSink float64

// BenchmarkNYSIISCode_ASCII_Short benchmarks NYSIISCode on a short name (Brown,
// 5 chars). Budget: < 500 ns, 0 allocs/op.
func BenchmarkNYSIISCode_ASCII_Short(b *testing.B) {
	b.ReportAllocs()
	var result string
	for i := 0; i < b.N; i++ {
		result = fuzzymatch.NYSIISCode("Brown")
	}
	nysiisSink = result
}

// BenchmarkNYSIISCode_ASCII_Medium benchmarks NYSIISCode on a medium name
// (Johnathan, 9 chars — also exercises the truncation path). Budget: < 500 ns.
func BenchmarkNYSIISCode_ASCII_Medium(b *testing.B) {
	b.ReportAllocs()
	var result string
	for i := 0; i < b.N; i++ {
		result = fuzzymatch.NYSIISCode("Johnathan")
	}
	nysiisSink = result
}

// BenchmarkNYSIISCode_ASCII_Long benchmarks NYSIISCode on a long name
// (Konstantinopoulou, 18 chars). Budget: < 500 ns.
func BenchmarkNYSIISCode_ASCII_Long(b *testing.B) {
	b.ReportAllocs()
	var result string
	for i := 0; i < b.N; i++ {
		result = fuzzymatch.NYSIISCode("Konstantinopoulou")
	}
	nysiisSink = result
}

// BenchmarkNYSIISScore_Match benchmarks NYSIISScore for a matching pair
// (Brown/Browne, both encode to BRAN).
func BenchmarkNYSIISScore_Match(b *testing.B) {
	b.ReportAllocs()
	var result float64
	for i := 0; i < b.N; i++ {
		result = fuzzymatch.NYSIISScore("Brown", "Browne")
	}
	nysiisScoreSink = result
}

// BenchmarkNYSIISScore_NoMatch benchmarks NYSIISScore for a non-matching pair.
func BenchmarkNYSIISScore_NoMatch(b *testing.B) {
	b.ReportAllocs()
	var result float64
	for i := 0; i < b.N; i++ {
		result = fuzzymatch.NYSIISScore("Brown", "Robert")
	}
	nysiisScoreSink = result
}
