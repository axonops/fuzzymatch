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

// tokenise_bench_test.go runs allocation-aware benchmarks for the
// Tokenise splitter covering three ASCII input sizes (short / medium /
// long), one Unicode mixed-script case, one PascalCase shape, and the
// most-common DefaultOptions path. b.ReportAllocs() on every benchmark
// gates allocation regressions in bench.txt via benchstat (D-09 makes
// benchstat informational in CI; the developer workflow runs `make
// bench` locally and commits bench.txt when intentional).
//
// Performance budgets per .claude/skills/performance-standards (recorded
// here for reference; not asserted at runtime — bench.txt + benchstat is
// the enforcement mechanism):
//
//   - ASCII <= 50 chars: target < 500 ns/op, <= 2 allocs/op
//   - ASCII <= 500 chars: target proportional
//   - Unicode <= 50 runes: target < 2 µs/op
//
// The benchmarks iterate over a fixed input rather than per-call fresh
// allocation so the compiler can't fold the call into a constant.

package fuzzymatch_test

import (
	"strings"
	"testing"

	"github.com/axonops/fuzzymatch"
)

// tokeniseASCIIShort is 10 bytes — the typical short-identifier shape.
const tokeniseASCIIShort = "FooBar_Baz"

// tokeniseASCIIMedium is 49 bytes — at the upper end of "short" per
// performance-standards (target < 500 ns/op).
const tokeniseASCIIMedium = "FooBar_Baz.Qux/Quux-corgeGraultGarply_Waldo.Fred5"

// BenchmarkTokenise_ASCII_Short exercises the byte-level fast path on
// a 10-byte input under DefaultTokeniseOptions. Target: < 500 ns/op,
// <= 2 allocs/op (the []rune conversion + the result slice + the
// per-token lowercase buffer).
func BenchmarkTokenise_ASCII_Short(b *testing.B) {
	opts := fuzzymatch.DefaultTokeniseOptions()
	b.ReportAllocs()
	b.ResetTimer()
	var sink []string
	for i := 0; i < b.N; i++ {
		sink = fuzzymatch.Tokenise(tokeniseASCIIShort, opts)
	}
	if len(sink) == 0 {
		b.Fatal("sink unexpectedly empty — compiler folded the benchmark away")
	}
}

// BenchmarkTokenise_ASCII_Medium exercises the fast path with a
// 49-byte input. Establishes the linear-in-input-length behaviour at
// the "short" budget upper bound.
func BenchmarkTokenise_ASCII_Medium(b *testing.B) {
	opts := fuzzymatch.DefaultTokeniseOptions()
	b.ReportAllocs()
	b.ResetTimer()
	var sink []string
	for i := 0; i < b.N; i++ {
		sink = fuzzymatch.Tokenise(tokeniseASCIIMedium, opts)
	}
	if len(sink) == 0 {
		b.Fatal("sink unexpectedly empty")
	}
}

// BenchmarkTokenise_ASCII_Long exercises the fast path with a
// 500-byte compound identifier. Establishes scaling for benchstat
// regression detection across multi-token long inputs.
func BenchmarkTokenise_ASCII_Long(b *testing.B) {
	input := strings.Repeat("FooBar_Baz", 50) // 500 bytes
	opts := fuzzymatch.DefaultTokeniseOptions()
	b.ReportAllocs()
	b.ResetTimer()
	var sink []string
	for i := 0; i < b.N; i++ {
		sink = fuzzymatch.Tokenise(input, opts)
	}
	if len(sink) == 0 {
		b.Fatal("sink unexpectedly empty")
	}
}

// BenchmarkTokenise_Unicode_Short exercises the rune-level path with
// a 10-rune mixed-script input including a Latin->Cyrillic boundary
// (which the camelCase rule fires on).
func BenchmarkTokenise_Unicode_Short(b *testing.B) {
	input := "userПриветBaz"
	opts := fuzzymatch.DefaultTokeniseOptions()
	b.ReportAllocs()
	b.ResetTimer()
	var sink []string
	for i := 0; i < b.N; i++ {
		sink = fuzzymatch.Tokenise(input, opts)
	}
	if len(sink) == 0 {
		b.Fatal("sink unexpectedly empty")
	}
}

// BenchmarkTokenise_PascalCase exercises the PascalCase shape — the
// pattern most relevant to identifier-matching consumers (Phase 6
// token algorithms).
func BenchmarkTokenise_PascalCase(b *testing.B) {
	input := "UserCreateEventHandlerRegistry"
	opts := fuzzymatch.DefaultTokeniseOptions()
	b.ReportAllocs()
	b.ResetTimer()
	var sink []string
	for i := 0; i < b.N; i++ {
		sink = fuzzymatch.Tokenise(input, opts)
	}
	if len(sink) == 0 {
		b.Fatal("sink unexpectedly empty")
	}
}

// BenchmarkTokenise_DefaultOptions measures the most common path —
// DefaultTokeniseOptions on a typical camelCase identifier — because
// real-world consumers will hit it overwhelmingly often. Target:
// < 500 ns/op, <= 2 allocs/op.
func BenchmarkTokenise_DefaultOptions(b *testing.B) {
	input := "userServiceImpl"
	opts := fuzzymatch.DefaultTokeniseOptions()
	b.ReportAllocs()
	b.ResetTimer()
	var sink []string
	for i := 0; i < b.N; i++ {
		sink = fuzzymatch.Tokenise(input, opts)
	}
	if len(sink) == 0 {
		b.Fatal("sink unexpectedly empty")
	}
}

// ---------------------------------------------------------------------
// Phase 8.5 Q8b — ASCII fast path allocation-budget benchmarks
// ---------------------------------------------------------------------
//
// The benchmarks below exist to prove the Q8b alloc-count budget:
//
//   - opts.Lowercase = false on ASCII input: substrings of the input
//     string with zero per-token allocation. The only alloc is the
//     []string header itself (1 alloc total, independent of input
//     length).
//   - opts.Lowercase = true on ASCII input: one scratch-buffer alloc
//     plus one per emitted token (the string([]byte) conversion is
//     structurally unavoidable without unsafe.String, which is
//     excluded by project policy).
//
// Three sizes (Short / Medium / Long) cover the spectrum of identifier
// lengths in real consumer workloads. The companion BenchmarkTokenise_
// Unicode_FastPath_Long benchmark documents the rune-path fallback
// shape on the same input size for a side-by-side bench.txt
// comparison via benchstat.

// BenchmarkTokenise_ASCII_FastPath_Short — opts.Lowercase=false,
// ~20-byte camelCase identifier. Expected: 1 alloc (the []string
// header); tokens are zero-copy substrings of the input.
func BenchmarkTokenise_ASCII_FastPath_Short(b *testing.B) {
	input := "fooBarBazQuxQuuxCorge" // 21 bytes, 6 tokens
	opts := fuzzymatch.TokeniseOptions{
		Lowercase:             false,
		SplitCamelCase:        true,
		SplitConsecutiveUpper: true,
		SeparatorChars:        fuzzymatch.DefaultTokeniseOptions().SeparatorChars,
	}
	b.ReportAllocs()
	b.ResetTimer()
	var sink []string
	for i := 0; i < b.N; i++ {
		sink = fuzzymatch.Tokenise(input, opts)
	}
	if len(sink) == 0 {
		b.Fatal("sink unexpectedly empty")
	}
}

// BenchmarkTokenise_ASCII_FastPath_Medium — opts.Lowercase=false,
// ~80-byte mixed-style identifier. Expected: 1 alloc (the []string
// header); tokens are zero-copy substrings.
func BenchmarkTokenise_ASCII_FastPath_Medium(b *testing.B) {
	input := "fooBar_baz.qux/quuxCorge-graultGarply_waldoFredPlughXyzzyThudHelloWorldSpamEggs" // ~80 bytes
	opts := fuzzymatch.TokeniseOptions{
		Lowercase:             false,
		SplitCamelCase:        true,
		SplitConsecutiveUpper: true,
		SeparatorChars:        fuzzymatch.DefaultTokeniseOptions().SeparatorChars,
	}
	b.ReportAllocs()
	b.ResetTimer()
	var sink []string
	for i := 0; i < b.N; i++ {
		sink = fuzzymatch.Tokenise(input, opts)
	}
	if len(sink) == 0 {
		b.Fatal("sink unexpectedly empty")
	}
}

// BenchmarkTokenise_ASCII_FastPath_Long — opts.Lowercase=false,
// ~500-byte compound identifier. Expected: 1 alloc (the []string
// header); tokens are zero-copy substrings even at long input size,
// because the scratch buffer is skipped on the no-lowercase branch.
func BenchmarkTokenise_ASCII_FastPath_Long(b *testing.B) {
	input := strings.Repeat("fooBarBaz_quxQuuxCorge.", 23) + "garplyWaldo" // ~530 bytes
	opts := fuzzymatch.TokeniseOptions{
		Lowercase:             false,
		SplitCamelCase:        true,
		SplitConsecutiveUpper: true,
		SeparatorChars:        fuzzymatch.DefaultTokeniseOptions().SeparatorChars,
	}
	b.ReportAllocs()
	b.ResetTimer()
	var sink []string
	for i := 0; i < b.N; i++ {
		sink = fuzzymatch.Tokenise(input, opts)
	}
	if len(sink) == 0 {
		b.Fatal("sink unexpectedly empty")
	}
}

// BenchmarkTokenise_ASCII_FastPath_LowercaseLong — opts.Lowercase=true,
// ~500-byte compound identifier. Documents the lowercase branch
// allocation shape: one scratch-buffer alloc for the lowercased copy
// of the input plus one alloc per emitted token (string conversion).
// This is the upper bound of the v1.0 budget per the m11 LOCKED
// decision (2026-05-17) excluding unsafe.String.
func BenchmarkTokenise_ASCII_FastPath_LowercaseLong(b *testing.B) {
	input := strings.Repeat("fooBarBaz_quxQuuxCorge.", 23) + "garplyWaldo"
	opts := fuzzymatch.TokeniseOptions{
		Lowercase:             true,
		SplitCamelCase:        true,
		SplitConsecutiveUpper: true,
		SeparatorChars:        fuzzymatch.DefaultTokeniseOptions().SeparatorChars,
	}
	b.ReportAllocs()
	b.ResetTimer()
	var sink []string
	for i := 0; i < b.N; i++ {
		sink = fuzzymatch.Tokenise(input, opts)
	}
	if len(sink) == 0 {
		b.Fatal("sink unexpectedly empty")
	}
}

// BenchmarkTokenise_Unicode_FastPath_Long is the rune-path counterpart
// to BenchmarkTokenise_ASCII_FastPath_Long: same nominal byte length,
// same boundary shape, but one Cyrillic byte inserted to force the
// rune path. Used for a benchstat side-by-side comparison documenting
// the alloc-count delta between paths.
func BenchmarkTokenise_Unicode_FastPath_Long(b *testing.B) {
	// One non-ASCII rune at the front forces the rune path even
	// though the rest of the input is ASCII-identical to the
	// ASCII_FastPath_Long shape.
	input := "Приfoo" + strings.Repeat("BarBaz_quxQuuxCorge.foo", 22) + "garplyWaldo"
	opts := fuzzymatch.TokeniseOptions{
		Lowercase:             false,
		SplitCamelCase:        true,
		SplitConsecutiveUpper: true,
		SeparatorChars:        fuzzymatch.DefaultTokeniseOptions().SeparatorChars,
	}
	b.ReportAllocs()
	b.ResetTimer()
	var sink []string
	for i := 0; i < b.N; i++ {
		sink = fuzzymatch.Tokenise(input, opts)
	}
	if len(sink) == 0 {
		b.Fatal("sink unexpectedly empty")
	}
}
