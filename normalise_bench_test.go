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

// normalise_bench_test.go runs allocation-aware benchmarks for the
// Normalise pipeline, covering the ASCII fast path at three input sizes
// and the Unicode pipeline at two input sizes, plus the strip-
// diacritics and default-options shapes. b.ReportAllocs() on every
// benchmark gates allocation regressions in bench.txt via benchstat
// (D-09 makes benchstat informational in CI; the developer workflow
// runs `make bench` locally and commits bench.txt when intentional).
//
// Performance budgets per .claude/skills/performance-standards (Q7b,
// docs/requirements.md §14.1 — revised 2026-05 to match implementation
// reality; recorded here for reference; not asserted at runtime —
// bench.txt + benchstat is the enforcement mechanism):
//
//   - ASCII <= 50 chars:   target < 200 ns/op, <= 1 alloc/op (the
//                          make([]byte, 0, cap) scratch buffer + the
//                          string(buf) conversion on return; the
//                          buffer cannot live on the stack because
//                          escape analysis cannot prove its size at
//                          compile time, and `unsafe.String` is
//                          excluded by project policy)
//   - ASCII <= 500 chars:  target proportional, <= 1 alloc/op
//   - Unicode <= 50 runes: target < 2 µs/op, <= 3 allocs/op
//   - Unicode <= 500 runes: target proportional
//
// Q7b note: the 0-alloc target in earlier drafts was unachievable for
// the reason above. See Q7c scope note on Normalise godoc.
//
// The benchmarks iterate over a fixed input rather than a per-call
// fresh allocation so the compiler can't fold the call into a constant.

package fuzzymatch_test

import (
	"strings"
	"testing"

	"github.com/axonops/fuzzymatch"
)

// asciiShort is 10 bytes — the typical short-identifier shape.
const asciiShort = "FooBar_Baz"

// asciiMedium is 50 bytes — the upper end of "short" per
// performance-standards.
const asciiMedium = "FooBar_Baz.Qux/Quux-corgeGraultGarply_Waldo.Fred5"

// BenchmarkNormalise_ASCII_Short exercises the byte-level fast path
// for a 10-byte input under DefaultNormalisationOptions. Q7b target:
// < 200 ns/op, <= 1 alloc/op (the scratch buffer + string conversion;
// see file godoc).
func BenchmarkNormalise_ASCII_Short(b *testing.B) {
	opts := fuzzymatch.DefaultNormalisationOptions()
	b.ReportAllocs()
	b.ResetTimer()
	var sink string
	for i := 0; i < b.N; i++ {
		sink = fuzzymatch.Normalise(asciiShort, opts)
	}
	if sink == "" {
		b.Fatal("sink unexpectedly empty — compiler folded the benchmark away")
	}
}

// BenchmarkNormalise_ASCII_Medium exercises the fast path with a
// 50-byte input. The buffer is still small enough to be stack-eligible
// per the Go escape analysis; allocation budget remains 1 (the output
// string conversion).
func BenchmarkNormalise_ASCII_Medium(b *testing.B) {
	opts := fuzzymatch.DefaultNormalisationOptions()
	b.ReportAllocs()
	b.ResetTimer()
	var sink string
	for i := 0; i < b.N; i++ {
		sink = fuzzymatch.Normalise(asciiMedium, opts)
	}
	if sink == "" {
		b.Fatal("sink unexpectedly empty")
	}
}

// BenchmarkNormalise_ASCII_Long exercises the fast path with a
// 500-byte input. The make([]byte, 0, len(s)*2+1) capacity hint allocates
// once; expected allocations: 2 (the buffer plus the output string
// conversion).
func BenchmarkNormalise_ASCII_Long(b *testing.B) {
	input := strings.Repeat("FooBar_Baz", 50) // 500 bytes
	opts := fuzzymatch.DefaultNormalisationOptions()
	b.ReportAllocs()
	b.ResetTimer()
	var sink string
	for i := 0; i < b.N; i++ {
		sink = fuzzymatch.Normalise(input, opts)
	}
	if sink == "" {
		b.Fatal("sink unexpectedly empty")
	}
}

// BenchmarkNormalise_Unicode_Short exercises the slow path with a
// 10-rune mixed-script input. transform.String allocates the chain plus
// the output buffer; the fold-runes second pass allocates the byte
// buffer.
func BenchmarkNormalise_Unicode_Short(b *testing.B) {
	input := "Müller café"
	opts := fuzzymatch.DefaultNormalisationOptions()
	b.ReportAllocs()
	b.ResetTimer()
	var sink string
	for i := 0; i < b.N; i++ {
		sink = fuzzymatch.Normalise(input, opts)
	}
	if sink == "" {
		b.Fatal("sink unexpectedly empty")
	}
}

// BenchmarkNormalise_Unicode_Long exercises the slow path with a
// 500-byte mixed-script input. Establishes scaling behaviour for
// benchstat regression detection.
func BenchmarkNormalise_Unicode_Long(b *testing.B) {
	input := strings.Repeat("Müller café Привет 你好 مرحبا ", 12) // ~500 bytes
	opts := fuzzymatch.DefaultNormalisationOptions()
	b.ReportAllocs()
	b.ResetTimer()
	var sink string
	for i := 0; i < b.N; i++ {
		sink = fuzzymatch.Normalise(input, opts)
	}
	if sink == "" {
		b.Fatal("sink unexpectedly empty")
	}
}

// BenchmarkNormalise_StripDiacritics_Short exercises the
// transform.Chain(NFD, Remove(Mn), NFC) pipeline with a diacritic-rich
// short input. This is the most allocation-heavy Normalise shape — the
// chain transformer plus the fold pass plus the strip pass.
func BenchmarkNormalise_StripDiacritics_Short(b *testing.B) {
	input := "café Müller naïve résumé"
	opts := fuzzymatch.NormalisationOptions{
		Lowercase:       true,
		StripSeparators: true,
		SeparatorChars:  fuzzymatch.DefaultNormalisationOptions().SeparatorChars,
		SplitCamelCase:  true,
		NFC:             true,
		StripDiacritics: true,
	}
	b.ReportAllocs()
	b.ResetTimer()
	var sink string
	for i := 0; i < b.N; i++ {
		sink = fuzzymatch.Normalise(input, opts)
	}
	if sink == "" {
		b.Fatal("sink unexpectedly empty")
	}
}

// BenchmarkNormalise_DefaultOptions_Short measures the most common
// path — DefaultNormalisationOptions() on a short ASCII identifier —
// because real-world consumers will hit it overwhelmingly often.
// Target: < 200 ns/op, <= 1 alloc/op.
func BenchmarkNormalise_DefaultOptions_Short(b *testing.B) {
	input := "userServiceImpl"
	opts := fuzzymatch.DefaultNormalisationOptions()
	b.ReportAllocs()
	b.ResetTimer()
	var sink string
	for i := 0; i < b.N; i++ {
		sink = fuzzymatch.Normalise(input, opts)
	}
	if sink == "" {
		b.Fatal("sink unexpectedly empty")
	}
}
