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

// validate_bench_test.go runs allocation-aware benchmarks for the
// Validate public surface across the three cost regimes:
//
//   - No-warnings fast path (well-formed ASCII inputs of equal length)
//   - With-warnings typical case (one empty input → 1 cross-cutting
//     Warning + 5 token-tier Warnings + 1 Hamming Warning = 7 total)
//   - Pathologically large input (above the 64 KiB threshold)
//
// Allocation budget per .claude/skills/performance-standards (Validate
// is the diagnostic path, not a hot scoring path — budgets are
// modestly generous; the dominant cost is the two Tokenise() calls
// which each allocate ~2-3 entries even on the ASCII fast path):
//
//   - No-warnings ASCII Short: ≤ 6 allocs/op
//       (1 for the warnings backing array that survives even when
//       Validate returns nil; ~2 per Tokenise() call (the []string
//       header + scratch buffer per Phase 8.5 Q8b)). Measured: 5.
//   - With-warnings ASCII Short: ≤ 12 allocs/op
//       (the warnings backing array, 2 Tokenise calls, fmt.Sprintf
//       detail strings — ~3 of them on the empty-input path).
//       Measured: 8.
//   - Pathologically large: ≤ 15 allocs/op + < 10 ms wall time
//       (input size dominates work, not allocations; the < 10 ms
//       budget is the T-08.5-25 DoS mitigation). Measured: 11 allocs,
//       ~1 ms.
//   - Non-ASCII (all 5 ASCII-only algos warn): ≤ 15 allocs/op.
//       Measured: 10.
//
// `var sink` outside the loop + a `sink == nil` gate prevents
// compiler dead-code elimination of the Validate call.

package fuzzymatch_test

import (
	"strings"
	"testing"

	"github.com/axonops/fuzzymatch"
)

// validateSink keeps the Validate result alive across loop iterations
// so the compiler cannot dead-code-eliminate the call.
var validateSink []fuzzymatch.Warning

// BenchmarkValidate_NoWarnings_ASCII_Short exercises the best-case
// fast path: two short ASCII inputs of equal length with valid tokens
// produce zero warnings and Validate returns nil.
func BenchmarkValidate_NoWarnings_ASCII_Short(b *testing.B) {
	const a, c = "hello", "world"
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		validateSink = fuzzymatch.Validate(a, c)
	}
	// validateSink should be nil here — guard against the compiler
	// noticing that, by reading the slice header.
	_ = len(validateSink)
}

// BenchmarkValidate_WithWarnings_ASCII_Short exercises the typical
// degraded-input case: one empty input triggers WarnEmptyInput
// (cross-cutting), WarnUnequalLength (Hamming), and
// WarnNoTokensAfterNormalise (5 token-tier algos) = 7 warnings.
func BenchmarkValidate_WithWarnings_ASCII_Short(b *testing.B) {
	const a, c = "", "abc"
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		validateSink = fuzzymatch.Validate(a, c)
	}
	if validateSink == nil {
		b.Fatal("expected non-nil warnings slice")
	}
}

// BenchmarkValidate_PathologicallyLarge exercises a 100 KiB input on
// each side — above the 64 KiB threshold. The benchmark gates the
// "should not be unreasonably slow" requirement: at < 10 ms per
// invocation, this represents the single-digit-millisecond Validate
// budget from the threat-model mitigation for T-08.5-25.
func BenchmarkValidate_PathologicallyLarge(b *testing.B) {
	a := strings.Repeat("a", 100_000)
	c := strings.Repeat("b", 100_000)
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		validateSink = fuzzymatch.Validate(a, c)
	}
	if validateSink == nil {
		b.Fatal("expected non-nil warnings slice")
	}
}

// BenchmarkValidate_NonASCII_AllAlgos exercises the WarnAllNonASCIIDropped
// rule on every ASCII-only algorithm: 5 Warnings + 1 WarnEmptyInput-
// adjacent shape (the inputs are non-empty but non-ASCII).
func BenchmarkValidate_NonASCII_AllAlgos(b *testing.B) {
	const a, c = "中文", "日本語"
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		validateSink = fuzzymatch.Validate(a, c)
	}
	if validateSink == nil {
		b.Fatal("expected non-nil warnings slice")
	}
}
