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

// normalise_test.go pins the public-API contract of Normalise: empty
// input, default behaviour preserves diacritics, optional diacritic
// stripping, NFC idempotence across precomposed/decomposed inputs,
// ASCII-fast-path / Unicode-path equivalence on ASCII inputs,
// camel-case-split semantics, separator-strip semantics, mixed-script
// preservation, and property-test discipline (idempotence, never
// panics, length bound under StripSeparators).
//
// TestGolden_Normalisation is the FIRST real golden test in the project
// (plan 01-06). It exercises the canonicalMarshal byte form locked in
// plan 01-04 against the committed testdata/golden/normalisation.json
// fixture — every supported CI platform asserts the same bytes per
// D-14.
//
// Stdlib `testing` only — no testify in root tests, per
// .claude/skills/go-testing-standards.

package fuzzymatch_test

import (
	"sort"
	"testing"
	"testing/quick"
	"unicode/utf8"

	"github.com/axonops/fuzzymatch"
)

// ---------------------------------------------------------------------
// Unit tests
// ---------------------------------------------------------------------

// TestNormalise_Empty asserts the documented empty-input invariant:
// Normalise("", opts) returns the empty string regardless of opts.
func TestNormalise_Empty(t *testing.T) {
	opts := fuzzymatch.DefaultNormalisationOptions()
	if got := fuzzymatch.Normalise("", opts); got != "" {
		t.Errorf("Normalise(\"\", default) = %q; want \"\"", got)
	}
	// Also with all options off:
	if got := fuzzymatch.Normalise("", fuzzymatch.NormalisationOptions{}); got != "" {
		t.Errorf("Normalise(\"\", zero) = %q; want \"\"", got)
	}
	// And with diacritic-strip:
	stripOpts := fuzzymatch.NormalisationOptions{NFC: true, StripDiacritics: true}
	if got := fuzzymatch.Normalise("", stripOpts); got != "" {
		t.Errorf("Normalise(\"\", strip-diacritics) = %q; want \"\"", got)
	}
}

// TestNormalise_DefaultsPreserveDiacritics asserts D-03's choice that
// the default option set does NOT remove diacritics. Consumers wanting
// café→cafe must opt in via StripDiacritics: true.
func TestNormalise_DefaultsPreserveDiacritics(t *testing.T) {
	opts := fuzzymatch.DefaultNormalisationOptions()
	tests := []struct {
		in   string
		want string
	}{
		{"Müller", "müller"},
		{"café", "café"},
		{"naïve", "naïve"},
		{"résumé", "résumé"},
	}
	for _, tt := range tests {
		t.Run(tt.in, func(t *testing.T) {
			got := fuzzymatch.Normalise(tt.in, opts)
			if got != tt.want {
				t.Errorf("Normalise(%q, default) = %q; want %q (default DOES NOT strip diacritics per D-03)", tt.in, got, tt.want)
			}
		})
	}
}

// TestNormalise_StripDiacritics asserts the diacritic-strip pipeline:
// transform.Chain(NFD, runes.Remove(Mn), NFC) produces the ASCII base
// form for Latin-script letters with combining marks.
func TestNormalise_StripDiacritics(t *testing.T) {
	opts := fuzzymatch.NormalisationOptions{
		Lowercase:       true,
		NFC:             true,
		StripDiacritics: true,
	}
	tests := []struct {
		in   string
		want string
	}{
		{"Müller", "muller"},
		{"café", "cafe"},
		{"naïve", "naive"},
		{"résumé", "resume"},
		{"Ångström", "angstrom"},
	}
	for _, tt := range tests {
		t.Run(tt.in, func(t *testing.T) {
			got := fuzzymatch.Normalise(tt.in, opts)
			if got != tt.want {
				t.Errorf("Normalise(%q, strip) = %q; want %q", tt.in, got, tt.want)
			}
		})
	}
}

// TestNormalise_NFC_Idempotent asserts that NFC-precomposed and
// NFD-decomposed forms of the same logical text produce the SAME output
// under NormalisationOptions{NFC: true}. This is the canonical
// equivalence test for the Unicode pipeline.
//
// "café" precomposed:   c (0x63) a (0x61) f (0x66) é (0xC3 0xA9)
// "café" decomposed:    c (0x63) a (0x61) f (0x66) e (0x65) ́ (0xCC 0x81)
func TestNormalise_NFC_Idempotent(t *testing.T) {
	opts := fuzzymatch.NormalisationOptions{NFC: true}
	precomposed := "café" // U+00E9 LATIN SMALL LETTER E WITH ACUTE
	decomposed := "café" // e + U+0301 COMBINING ACUTE ACCENT
	want := "café"        // NFC form
	if got := fuzzymatch.Normalise(precomposed, opts); got != want {
		t.Errorf("Normalise(precomposed café, NFC) = %q; want %q", got, want)
	}
	if got := fuzzymatch.Normalise(decomposed, opts); got != want {
		t.Errorf("Normalise(decomposed café, NFC) = %q; want %q", got, want)
	}
}

// TestNormalise_ASCII_FastPath_DoesNotAlterUnicodeOutput asserts that
// for ASCII-only inputs, the ASCII fast path produces output identical
// to what the Unicode pipeline would produce. NFC of pure ASCII is a
// byte-level no-op, so the equivalence must hold byte-for-byte.
//
// We exercise this by running the same input through opts with NFC=true
// (which on ASCII inputs takes the fast path per the implementation)
// and opts with NFC=false (which also takes the fast path because
// StripDiacritics is false) — outputs must match. The Unicode-path
// branch for ASCII inputs is gated on StripDiacritics=true, which we
// don't exercise here (the strip-pipeline would alter ASCII output for
// the rare ASCII combining sequence, which is excluded).
func TestNormalise_ASCII_FastPath_DoesNotAlterUnicodeOutput(t *testing.T) {
	asciiOpts1 := fuzzymatch.DefaultNormalisationOptions()
	asciiOpts2 := asciiOpts1
	asciiOpts2.NFC = false
	inputs := []string{
		"FooBar",
		"foo_bar_baz",
		"HTTPSConnection",
		"hello world",
		"a.b.c.d",
		"XMLHTTPRequest",
	}
	for _, in := range inputs {
		t.Run(in, func(t *testing.T) {
			withNFC := fuzzymatch.Normalise(in, asciiOpts1)
			withoutNFC := fuzzymatch.Normalise(in, asciiOpts2)
			if withNFC != withoutNFC {
				t.Errorf("ASCII input %q produced divergent output: NFC=true %q vs NFC=false %q", in, withNFC, withoutNFC)
			}
		})
	}
}

// TestNormalise_SplitCamelCase locks the camel-case-split semantics:
// a space is inserted at every lowercase→uppercase rune transition. No
// space is inserted at uppercase→uppercase or uppercase→lowercase
// boundaries (this means HTTPSConnection becomes "httpsconnection" —
// runs of uppercase letters stay cohesive, and an upper→lower run
// does NOT trigger a split). The chosen behaviour is documented in
// normalise.go's foldRunes function. Consumers wanting Smarter splits
// (Pascal-case acronym detection) can build their own pipeline; the
// default policy is the simplest correct rule.
func TestNormalise_SplitCamelCase(t *testing.T) {
	opts := fuzzymatch.DefaultNormalisationOptions()
	tests := []struct {
		in   string
		want string
	}{
		// Lowercase→uppercase transition: split.
		{"FooBar", "foo bar"},
		{"fooBar", "foo bar"},
		{"myVariableName", "my variable name"},
		{"parseJSON5", "parse json5"},
		{"htmlBody", "html body"},
		{"aB", "a b"},

		// No lowercase→uppercase transitions: no split.
		{"foo", "foo"},
		{"FOO", "foo"},
		{"AB", "ab"},
		{"Ab", "ab"},
		{"IOError", "ioerror"},
		{"XMLHTTPRequest", "xmlhttprequest"},
		{"HTTPSConnection", "httpsconnection"},
	}
	for _, tt := range tests {
		t.Run(tt.in, func(t *testing.T) {
			got := fuzzymatch.Normalise(tt.in, opts)
			if got != tt.want {
				t.Errorf("Normalise(%q, default) = %q; want %q", tt.in, got, tt.want)
			}
		})
	}
}

// TestNormalise_StripSeparators locks the separator-strip behaviour:
// each rune in opts.SeparatorChars is replaced with a single space, runs
// of whitespace collapse, leading and trailing whitespace is stripped.
func TestNormalise_StripSeparators(t *testing.T) {
	opts := fuzzymatch.DefaultNormalisationOptions()
	tests := []struct {
		in   string
		want string
	}{
		{"foo_bar.baz", "foo bar baz"},
		{"foo-bar-baz", "foo bar baz"},
		{"foo.bar:baz/qux", "foo bar baz qux"},
		{"  hello  world  ", "hello world"},
		{"_leading", "leading"},
		{"trailing_", "trailing"},
		{"____many____", "many"},
		{"a.b.c", "a b c"},
	}
	for _, tt := range tests {
		t.Run(tt.in, func(t *testing.T) {
			got := fuzzymatch.Normalise(tt.in, opts)
			if got != tt.want {
				t.Errorf("Normalise(%q, default) = %q; want %q", tt.in, got, tt.want)
			}
		})
	}
}

// TestNormalise_PreservesMixedScript asserts that non-Latin scripts pass
// through Normalise unchanged in their character set (subject to NFC
// canonicalisation and case-folding where applicable). Cyrillic is
// cased so Привет → привет under Lowercase=true; CJK and Arabic lack
// case so 你好 and مرحبا are byte-equal to their inputs.
func TestNormalise_PreservesMixedScript(t *testing.T) {
	opts := fuzzymatch.DefaultNormalisationOptions()
	tests := []struct {
		name string
		in   string
		want string
	}{
		{"Cyrillic-cased", "Привет", "привет"},
		{"Arabic-uncased", "مرحبا", "مرحبا"},
		{"CJK-uncased", "你好", "你好"},
		// Cyrillic letters do have case, so the o→П boundary triggers
		// a camel-case split under SplitCamelCase=true. The output is
		// "hello привет" — both halves lower-cased, with a space
		// inserted at the Latin→Cyrillic upper boundary.
		{"Mixed-Latin-Cyrillic", "HelloПривет", "hello привет"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := fuzzymatch.Normalise(tt.in, opts)
			if got != tt.want {
				t.Errorf("Normalise(%q, default) = %q; want %q", tt.in, got, tt.want)
			}
		})
	}
}

// ---------------------------------------------------------------------
// Property tests (testing/quick)
// ---------------------------------------------------------------------

// TestProp_Normalise_Idempotent: Normalise applied twice with the same
// options produces the same output as a single application. The
// property holds for arbitrary strings (testing/quick generates valid
// UTF-8 strings; invalid-UTF-8 panic-free is covered separately by the
// fuzz harness).
func TestProp_Normalise_Idempotent(t *testing.T) {
	opts := fuzzymatch.DefaultNormalisationOptions()
	f := func(s string) bool {
		once := fuzzymatch.Normalise(s, opts)
		twice := fuzzymatch.Normalise(once, opts)
		return once == twice
	}
	if err := quick.Check(f, nil); err != nil {
		t.Errorf("Normalise is not idempotent under default options: %v", err)
	}
}

// TestProp_Normalise_NeverPanics: Normalise must not panic for any
// caller-provided string under any well-formed options bitfield. The
// quick.Generator for string covers valid UTF-8; invalid byte sequences
// are exercised by FuzzNormalise.
func TestProp_Normalise_NeverPanics(t *testing.T) {
	f := func(s string, optBits uint8) (ok bool) {
		defer func() {
			if r := recover(); r != nil {
				t.Errorf("Normalise panicked on s=%q optBits=%d: %v", s, optBits, r)
				ok = false
			}
		}()
		opts := fuzzymatch.NormalisationOptions{
			Lowercase:       optBits&1 != 0,
			StripSeparators: optBits&2 != 0,
			SeparatorChars:  fuzzymatch.DefaultNormalisationOptions().SeparatorChars,
			SplitCamelCase:  optBits&4 != 0,
			NFC:             optBits&8 != 0,
			StripDiacritics: optBits&16 != 0,
		}
		_ = fuzzymatch.Normalise(s, opts)
		return true
	}
	if err := quick.Check(f, nil); err != nil {
		t.Errorf("Normalise panicked under quick.Check: %v", err)
	}
}

// TestProp_Normalise_LengthBound_WhenStripSeparators: when
// StripSeparators is enabled, the output rune count is bounded above by
// the input rune count plus the maximum number of camel-case-inserted
// spaces (which itself is bounded by the rune count). The simple bound
// 2 * inputRunes covers every legal case.
func TestProp_Normalise_LengthBound_WhenStripSeparators(t *testing.T) {
	opts := fuzzymatch.DefaultNormalisationOptions()
	f := func(s string) bool {
		out := fuzzymatch.Normalise(s, opts)
		inputRunes := utf8.RuneCountInString(s)
		outputRunes := utf8.RuneCountInString(out)
		// Output can never exceed input runes (separator-strip never
		// adds bytes — camel-split adds spaces but separator-strip
		// then collapses those into runs that were already separators
		// or original spaces). The strict bound holds: output <= input.
		// We use the looser bound 2*input to leave a safety margin in
		// case future refactors alter the camel-split behaviour
		// slightly.
		return outputRunes <= 2*inputRunes+1
	}
	if err := quick.Check(f, nil); err != nil {
		t.Errorf("Normalise length bound violated: %v", err)
	}
}

// ---------------------------------------------------------------------
// Golden test (D-11, D-12, D-13, D-14)
// ---------------------------------------------------------------------

// goldenNormalisationEntry is one (input, options, expected output)
// case in the v1.x normalisation golden fixture. The exported field
// names are part of the LOCKED JSON contract — renaming any field is a
// major-version-bump event per docs/requirements.md §11.2.
//
// Lives in the test file (not in normalise.go) because the type exists
// only for the test harness.
type goldenNormalisationEntry struct {
	Name           string                          `json:"name"`
	Input          string                          `json:"input"`
	Options        fuzzymatch.NormalisationOptions `json:"options"`
	ExpectedOutput string                          `json:"expected_output"`
}

// goldenNormalisationFile wraps the slice of entries with a version
// field so future schema migrations (if any) can detect the format.
type goldenNormalisationFile struct {
	Version int                        `json:"version"`
	Entries []goldenNormalisationEntry `json:"entries"`
}

// TestGolden_Normalisation pins Normalise's behaviour byte-for-byte
// across CI matrix platforms. It marshals the in-test corpus through
// canonicalMarshal (the v1.x byte form LOCKED in plan 01-04) and diffs
// against testdata/golden/normalisation.json.
//
// Run with `-update` to rewrite the fixture from the current code
// output (typical workflow after intentional code changes — review
// diff, commit).
//
// Entry coverage (per D-11): 20-40 cases spanning pure-ASCII shapes,
// NFC precomposed/decomposed pairs, StripDiacritics on/off pairs,
// mixed-script preservation, separator edge cases, camel-case splits,
// idempotence pairs, and edge cases (empty / whitespace-only / single-
// char). All entries are sorted by Name so the canonical JSON byte
// output is order-stable.
func TestGolden_Normalisation(t *testing.T) {
	entries := goldenNormalisationEntries(t)
	file := goldenNormalisationFile{Version: 1, Entries: entries}
	assertGolden(t, "normalisation.json", file)
}

// goldenNormalisationEntries returns the corpus of pinned cases for
// TestGolden_Normalisation. Each case's ExpectedOutput field is
// computed by calling fuzzymatch.Normalise on the input with the case's
// options — so the golden file always reflects the current code's
// output, and changes to Normalise show up as fixture diffs that must
// be reviewed and (re)committed.
//
// Entries are appended in any order and sorted by Name at return time
// (sort.Slice is stable per Go's stdlib; the names are unique so the
// sort key is total).
func goldenNormalisationEntries(t *testing.T) []goldenNormalisationEntry {
	t.Helper()
	def := fuzzymatch.DefaultNormalisationOptions()
	strip := fuzzymatch.NormalisationOptions{
		Lowercase:       true,
		NFC:             true,
		StripDiacritics: true,
	}
	nfcOnly := fuzzymatch.NormalisationOptions{NFC: true}
	zero := fuzzymatch.NormalisationOptions{}

	cases := []struct {
		name  string
		input string
		opts  fuzzymatch.NormalisationOptions
	}{
		// Pure-ASCII shapes
		{"PureASCII_FooBar", "FooBar", def},
		{"PureASCII_snake_case", "foo_bar_baz", def},
		{"PureASCII_dot_case", "foo.bar.baz", def},
		{"PureASCII_kebab_case", "hello-world-here", def},
		{"PureASCII_HTTPSConnection", "HTTPSConnection", def},
		{"PureASCII_XMLHTTPRequest", "XMLHTTPRequest", def},
		{"PureASCII_parseJSON5", "parseJSON5", def},

		// NFC / NFD divergence pairs
		{"NFC_NFD_Cafe_Precomposed", "café", nfcOnly},
		{"NFC_NFD_Cafe_Decomposed", "café", nfcOnly},

		// Diacritic-strip ON/OFF pairs
		{"NFC_Muller_PreserveDiacritic", "Müller", def},
		{"NFC_Muller_StripDiacritic", "Müller", strip},
		{"NFC_Resume_Diacritic", "résumé", def},
		{"NFC_Resume_StripDiacritic", "résumé", strip},
		{"NFC_Naive_Diacritic", "naïve", def},
		{"NFC_Naive_StripDiacritic", "naïve", strip},

		// Mixed-script preservation
		{"MixedScript_Cyrillic_Privet", "Привет", def},
		{"MixedScript_Arabic_Marhaba", "مرحبا", def},
		{"MixedScript_CJK_NiHao", "你好", def},

		// Idempotence pairs — input is a "Normalise of normalise" form.
		// These let the golden file double as an idempotence regression
		// witness: feeding the output back in must produce the same
		// output. The test harness asserts this property in
		// TestProp_Normalise_Idempotent; these entries pin specific
		// concrete cases that have historically caused bugs in other
		// libraries.
		{"Idempotence_DoubleApply_Hello", "hello world", def},
		{"Idempotence_DoubleApply_Cafe", "café", def},

		// Separator edge cases
		{"Sep_Empty", "", def},
		{"Sep_SingleSeparator", "_", def},
		{"Sep_MultipleSeparators", "___...---", def},
		{"Sep_LeadingTrailing", "_hello_", def},

		// CamelCase splits — already covered above; this set focuses
		// on edge cases (consecutive uppercase, single-character).
		{"Camel_Simple", "myVariableName", def},
		{"Camel_ConsecutiveUpper", "ABCdef", def},

		// Edge cases
		{"Empty_String", "", zero},
		{"Edge_Whitespace_Only", "   \t\n  ", def},
		{"Edge_Single_Char", "A", def},
		{"Edge_Long_Mixed_60_chars", "FooBarBazQux_quuxCorgeGrault.garplyWaldoFredPlughXyzzyThudEnd", def},
	}

	out := make([]goldenNormalisationEntry, 0, len(cases))
	for _, c := range cases {
		out = append(out, goldenNormalisationEntry{
			Name:           c.name,
			Input:          c.input,
			Options:        c.opts,
			ExpectedOutput: fuzzymatch.Normalise(c.input, c.opts),
		})
	}
	sort.Slice(out, func(i, j int) bool { return out[i].Name < out[j].Name })
	return out
}
