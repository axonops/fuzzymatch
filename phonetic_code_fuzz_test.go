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

// phonetic_code_fuzz_test.go runs native Go fuzzing against the four
// phonetic-code public surfaces:
//
//   - DoubleMetaphoneKeys (Philips 2000) — returns (primary, secondary)
//   - NYSIISCode          (Taft 1970)    — returns the NYSIIS code
//   - SoundexCode         (Russell 1918) — returns the 4-char Soundex code
//   - MRACode             (NBS TN 943)   — returns the MRA code
//
// The byte-path *Score surfaces have their own per-algorithm fuzz
// harnesses (double_metaphone_fuzz_test.go etc.); this file targets
// the encoder surfaces directly so we exercise the rule tables
// without the score-comparison wrapper.
//
// Properties (per encoder, per input):
//
//  1. Never panics on arbitrary input including invalid UTF-8,
//     embedded NULs, very long strings, and arbitrary high-byte
//     sequences. Phonetic encoders run their own rule-table state
//     machines and have historically been a source of panic bugs
//     (out-of-bounds reads on edge cases) in third-party Go
//     phonetic libraries — this fuzz harness defends against
//     regression.
//
// The harness does NOT assert specific output strings — phonetic
// algorithms are deterministic but their outputs vary across the
// algorithm family and the harness's job is panic-free arbitrary
// input, not algorithm-specific correctness. Algorithm-specific
// correctness is pinned by the per-algorithm _test.go files plus
// double_metaphone_paper_test.go (Q11c paper-anchored worked
// examples).
//
// Threat model: T-08.5-24 (D - DoS via fuzz-discovered panic).

package fuzzymatch_test

import (
	"testing"

	"github.com/axonops/fuzzymatch"
)

// FuzzPhoneticCodes asserts panic-freedom for the four phonetic
// encoder surfaces on arbitrary single-string input. The harness
// recovers any panic with a t.Fatalf so the offending input is
// surfaced as a fuzz crash with the exact panic value.
func FuzzPhoneticCodes(f *testing.F) {
	for _, s := range []string{
		"Schmidt",                    // DM canonical (German -> SHMT/XMT)
		"Schmit",                     // shortened variant
		"Knight",                     // silent K
		"Pfister",                    // PF cluster
		"Caesar",                     // C-E rule
		"O'Brien",                    // apostrophe
		"García",                     // diacritic
		"",                           // empty
		"a",                          // single char
		"\x00",                       // NUL
		"café",                       // Latin-supplement
		"Привет",                     // Cyrillic (encoders typically strip non-ASCII)
		"\xff\xfe",                   // invalid UTF-8
		"123",                        // digits (no letters)
		"!@#",                        // punctuation only
		"AbCdEfGhIjKlMnOpQrStUvWxYz", // every ASCII letter
		"AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA", // long run of same letter
	} {
		f.Add(s)
	}
	f.Fuzz(func(t *testing.T, s string) {
		// DoubleMetaphoneKeys
		func() {
			defer func() {
				if r := recover(); r != nil {
					t.Fatalf("DoubleMetaphoneKeys(%q) panicked: %v", s, r)
				}
			}()
			_, _ = fuzzymatch.DoubleMetaphoneKeys(s)
		}()

		// NYSIISCode
		func() {
			defer func() {
				if r := recover(); r != nil {
					t.Fatalf("NYSIISCode(%q) panicked: %v", s, r)
				}
			}()
			_ = fuzzymatch.NYSIISCode(s)
		}()

		// SoundexCode
		func() {
			defer func() {
				if r := recover(); r != nil {
					t.Fatalf("SoundexCode(%q) panicked: %v", s, r)
				}
			}()
			_ = fuzzymatch.SoundexCode(s)
		}()

		// MRACode
		func() {
			defer func() {
				if r := recover(); r != nil {
					t.Fatalf("MRACode(%q) panicked: %v", s, r)
				}
			}()
			_ = fuzzymatch.MRACode(s)
		}()
	})
}
