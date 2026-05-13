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

// normalise_fuzz_test.go runs native Go fuzzing against Normalise. The
// fuzzer asserts two properties for ANY input — valid UTF-8 strings,
// invalid UTF-8 byte sequences, lone surrogates, embedded NULs, very
// long inputs:
//
//   1. Normalise never panics (TEST-03 + .claude/skills/
//      algorithm-correctness-standards Unicode Handling).
//   2. The output is always valid UTF-8 (x/text replaces invalid
//      sequences with U+FFFD per Go's standard convention).
//
// Options are encoded as a uint8 bitfield so the fuzz engine can
// explore the 2^5 = 32 option combinations cheaply. SeparatorChars is
// fixed at DefaultNormalisationOptions().SeparatorChars rather than
// fuzzed — varying it adds combinatorial explosion without exercising
// new code paths (the path differences are gated on the option bits,
// not on the separator string content).
//
// Seed corpus lives in testdata/fuzz/FuzzNormalise/ as files named
// seed-001 .. seed-005, each in Go's `go test fuzz v1` format. The
// corpus exercises the known pathological shapes documented in
// .planning/research/PITFALLS.md.

package fuzzymatch_test

import (
	"testing"
	"unicode/utf8"

	"github.com/axonops/fuzzymatch"
)

// FuzzNormalise is the panic-free / valid-UTF-8 fuzz harness for the
// Normalise pipeline. The argument types — (string, uint8) — match the
// seed files in testdata/fuzz/FuzzNormalise/.
//
// CI's nightly fuzz job (per Makefile `test-fuzz`) runs each fuzzer for
// 60 seconds; the developer workflow runs `go test -fuzz=FuzzNormalise
// -fuzztime=5s ./...` as a quick smoke test before commits that touch
// normalise.go.
func FuzzNormalise(f *testing.F) {
	// Programmatic seed entries — these compose with the on-disk seed
	// files in testdata/fuzz/FuzzNormalise/ (the on-disk files cover
	// the known pathological patterns; these in-file seeds cover the
	// happy-path canonical reference vectors).
	for _, s := range []string{
		"",
		"café",
		"FooBar",
		"naïve",
		"\xff\xfe",     // invalid UTF-8 — high bytes without continuation
		"\xed\xa0\x80", // lone surrogate (encoded as 3-byte UTF-8)
		"Müller",
		"a\x00b",          // embedded NUL
		"  \t\n  ",        // whitespace-only
		"Привет 你好 مرحبا", // mixed script
	} {
		for bits := uint8(0); bits < 32; bits++ {
			f.Add(s, bits)
		}
	}

	f.Fuzz(func(t *testing.T, s string, optBits uint8) {
		opts := fuzzymatch.NormalisationOptions{
			Lowercase:       optBits&1 != 0,
			StripSeparators: optBits&2 != 0,
			SeparatorChars:  fuzzymatch.DefaultNormalisationOptions().SeparatorChars,
			SplitCamelCase:  optBits&4 != 0,
			NFC:             optBits&8 != 0,
			StripDiacritics: optBits&16 != 0,
		}
		// Property 1: must not panic. (Implicit — any panic from the
		// pipeline propagates to the fuzz harness and is reported as a
		// crash.)
		got := fuzzymatch.Normalise(s, opts)

		// Property 2: output is valid UTF-8. x/text's normaliser
		// substitutes U+FFFD for malformed sequences in its input, so
		// even when s is invalid UTF-8 the output must validate. The
		// ASCII fast path is bounded to byte values < 0x80 (so its
		// output is trivially valid UTF-8); the Unicode path goes
		// through transform.String which respects the U+FFFD
		// substitution rule.
		if !utf8.ValidString(got) {
			t.Errorf("Normalise produced invalid UTF-8 from input %q (bits=%d): out=%q", s, optBits, got)
		}
	})
}
