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

// tokenise_fuzz_test.go runs native Go fuzzing against Tokenise. The
// fuzzer asserts four properties for ANY input — valid UTF-8 strings,
// invalid UTF-8 byte sequences, lone surrogates, embedded NULs, very
// long inputs:
//
//   1. Tokenise never panics (TEST-03 + .claude/skills/
//      algorithm-correctness-standards Unicode Handling).
//   2. The returned slice is never nil (the non-nil-empty-slice
//      contract is asserted under all options).
//   3. No token in the returned slice is empty (zero-width-token
//      filter holds across all inputs).
//   4. Every returned token is valid UTF-8 (Go's []rune conversion
//      replaces invalid sequences with U+FFFD).
//
// Options are encoded as a uint8 bitfield (3 bits for the 3 bool
// fields). SeparatorChars is fixed at the default for the same reason
// as FuzzNormalise — varying it adds combinatorial explosion without
// exercising new code paths.
//
// Seed corpus lives in testdata/fuzz/FuzzTokenise/ as files named
// seed-001 .. seed-005, each in Go's `go test fuzz v1` format. The
// corpus exercises canonical reference inputs (FooBar, snake_case)
// and the pathological-input pitfalls (invalid UTF-8, lone surrogates,
// embedded NULs, mixed scripts).

package fuzzymatch_test

import (
	"testing"
	"unicode/utf8"

	"github.com/axonops/fuzzymatch"
)

// FuzzTokenise is the panic-free / non-nil / non-empty-token / valid-
// UTF-8 fuzz harness for the Tokenise splitter. The argument types —
// (string, uint8) — match the seed files in
// testdata/fuzz/FuzzTokenise/.
//
// CI's nightly fuzz job (per Makefile `test-fuzz`) runs each fuzzer for
// 60 seconds; the developer workflow runs `go test -fuzz=FuzzTokenise
// -fuzztime=5s ./...` as a quick smoke test before commits that touch
// tokenise.go.
func FuzzTokenise(f *testing.F) {
	// Programmatic seed entries — these compose with the on-disk seed
	// files in testdata/fuzz/FuzzTokenise/. Inline seeds cover the
	// happy-path canonical reference vectors; on-disk seeds cover the
	// pathological patterns documented in .planning/research/PITFALLS.md.
	for _, s := range []string{
		"",
		" ",
		"foo",
		"FooBar",
		"snake_case",
		"kebab-case",
		"dot.case",
		"XMLHTTPRequest",
		"userПривет",
		"\xff\xfe",     // invalid UTF-8 — high bytes without continuation
		"\xed\xa0\x80", // lone surrogate encoded as 3-byte UTF-8
		"Foo123Bar",
		"__leading",
		"trailing__",
		"a\x00b",   // embedded NUL
		"  \t\n  ", // whitespace-only
	} {
		for bits := uint8(0); bits < 8; bits++ {
			f.Add(s, bits)
		}
	}

	f.Fuzz(func(t *testing.T, s string, optBits uint8) {
		opts := fuzzymatch.TokeniseOptions{
			Lowercase:             optBits&1 != 0,
			SplitCamelCase:        optBits&2 != 0,
			SplitConsecutiveUpper: optBits&4 != 0,
			SeparatorChars:        fuzzymatch.DefaultTokeniseOptions().SeparatorChars,
		}

		// Property 1: must not panic. Implicit — any panic from the
		// pipeline propagates to the fuzz harness and is reported as a
		// crash.
		got := fuzzymatch.Tokenise(s, opts)

		// Property 2: returned slice is never nil. The contract
		// guarantees a non-nil slice even when the result is empty.
		if got == nil {
			t.Errorf("Tokenise returned nil; want non-nil slice (input=%q bits=%d)", s, optBits)
		}

		// Property 3: no token is empty. Consecutive separators and
		// boundary collisions never produce zero-width tokens.
		for i, tok := range got {
			if tok == "" {
				t.Errorf("Tokenise emitted empty token at index %d (input=%q bits=%d)", i, s, optBits)
			}
		}

		// Property 4: every returned token is valid UTF-8. Go's
		// []rune conversion substitutes U+FFFD for malformed input
		// sequences, so even when s is invalid UTF-8 the tokens
		// validate.
		for i, tok := range got {
			if !utf8.ValidString(tok) {
				t.Errorf("Tokenise produced invalid UTF-8 token at index %d (input=%q bits=%d): out=%q", i, s, optBits, tok)
			}
		}
	})
}
