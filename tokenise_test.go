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

// tokenise_test.go pins the public-API contract of Tokenise: the
// non-nil-empty-slice contract on empty input, the canonical splits for
// snake_case / camelCase / PascalCase / kebab-case / dot-case, the
// consecutive-uppercase trailing-boundary rule, mixed-script handling,
// numeric-attachment behaviour, the no-empty-tokens invariant, and
// property tests for order stability, panic-freedom, valid UTF-8 output,
// the token-count rune bound, and reconstructibility on ASCII inputs.
//
// Stdlib `testing` only — no testify in root tests, per
// .claude/skills/go-testing-standards.

package fuzzymatch_test

import (
	"math/rand"
	"reflect"
	"strings"
	"testing"
	"testing/quick"
	"unicode/utf8"

	"github.com/axonops/fuzzymatch"
)

// ---------------------------------------------------------------------
// Unit tests
// ---------------------------------------------------------------------

// TestTokenise_Empty asserts the documented empty-input contract:
// Tokenise("", any-opts) returns a non-nil slice of length zero.
// Distinguishing nil from empty is part of the public contract because
// downstream consumers (Phase 6 token algorithms) may want to iterate
// the slice unconditionally.
func TestTokenise_Empty(t *testing.T) {
	cases := []struct {
		name string
		opts fuzzymatch.TokeniseOptions
	}{
		{"default", fuzzymatch.DefaultTokeniseOptions()},
		{"zero", fuzzymatch.TokeniseOptions{}},
		{"only-separators-default", fuzzymatch.DefaultTokeniseOptions()},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			got := fuzzymatch.Tokenise("", c.opts)
			if got == nil {
				t.Errorf("Tokenise(\"\", %s) returned nil; want non-nil empty slice", c.name)
			}
			if len(got) != 0 {
				t.Errorf("Tokenise(\"\", %s) = %q; want []", c.name, got)
			}
		})
	}
}

// TestTokenise_WhitespaceOnly asserts that input consisting only of
// whitespace (which appears in the default SeparatorChars) returns a
// non-nil empty slice. This is a stronger statement than Empty —
// whitespace-only input is non-empty as a string but tokenises to
// nothing because every rune is discarded as a separator.
func TestTokenise_WhitespaceOnly(t *testing.T) {
	opts := fuzzymatch.DefaultTokeniseOptions()
	inputs := []string{" ", "  ", "\t", "\n", "\r", " \t\n\r ", "   \t\n  "}
	for _, in := range inputs {
		t.Run(strings.ReplaceAll(in, "\n", "\\n"), func(t *testing.T) {
			got := fuzzymatch.Tokenise(in, opts)
			if got == nil {
				t.Errorf("Tokenise(%q, default) returned nil; want non-nil empty slice", in)
			}
			if len(got) != 0 {
				t.Errorf("Tokenise(%q, default) = %q; want []", in, got)
			}
		})
	}
}

// TestTokenise_SingleChar locks the minimum non-empty case: a single
// character input returns a one-element slice containing that character
// (lowercased under default opts).
func TestTokenise_SingleChar(t *testing.T) {
	opts := fuzzymatch.DefaultTokeniseOptions()
	cases := []struct {
		in   string
		want []string
	}{
		{"a", []string{"a"}},
		{"A", []string{"a"}},
		{"Z", []string{"z"}},
		{"5", []string{"5"}},
	}
	for _, c := range cases {
		t.Run(c.in, func(t *testing.T) {
			got := fuzzymatch.Tokenise(c.in, opts)
			if !reflect.DeepEqual(got, c.want) {
				t.Errorf("Tokenise(%q, default) = %q; want %q", c.in, got, c.want)
			}
		})
	}
}

// TestTokenise_AlreadyTokenised covers the common "consumer pre-split
// the input" path. Spaces are in the default SeparatorChars so each
// space-separated word becomes a token.
func TestTokenise_AlreadyTokenised(t *testing.T) {
	opts := fuzzymatch.DefaultTokeniseOptions()
	cases := []struct {
		in   string
		want []string
	}{
		{"foo bar baz", []string{"foo", "bar", "baz"}},
		{"  foo  bar  ", []string{"foo", "bar"}},
		{"a b c d", []string{"a", "b", "c", "d"}},
	}
	for _, c := range cases {
		t.Run(c.in, func(t *testing.T) {
			got := fuzzymatch.Tokenise(c.in, opts)
			if !reflect.DeepEqual(got, c.want) {
				t.Errorf("Tokenise(%q, default) = %q; want %q", c.in, got, c.want)
			}
		})
	}
}

// TestTokenise_CamelCase pins the canonical camelCase splits: a
// lowercase letter followed by an uppercase letter triggers a boundary
// at the uppercase letter.
func TestTokenise_CamelCase(t *testing.T) {
	opts := fuzzymatch.DefaultTokeniseOptions()
	cases := []struct {
		in   string
		want []string
	}{
		{"FooBar", []string{"foo", "bar"}},
		{"fooBar", []string{"foo", "bar"}},
		{"parseJSON5", []string{"parse", "json5"}},
		{"IOError", []string{"io", "error"}},
		{"myVariableName", []string{"my", "variable", "name"}},
		{"httpRequestBody", []string{"http", "request", "body"}},
		{"aB", []string{"a", "b"}},
	}
	for _, c := range cases {
		t.Run(c.in, func(t *testing.T) {
			got := fuzzymatch.Tokenise(c.in, opts)
			if !reflect.DeepEqual(got, c.want) {
				t.Errorf("Tokenise(%q, default) = %q; want %q", c.in, got, c.want)
			}
		})
	}
}

// TestTokenise_SnakeCase pins the underscore-as-separator behaviour:
// underscores split AND are discarded. Leading/trailing/consecutive
// underscores never produce empty tokens.
func TestTokenise_SnakeCase(t *testing.T) {
	opts := fuzzymatch.DefaultTokeniseOptions()
	cases := []struct {
		in   string
		want []string
	}{
		{"foo_bar", []string{"foo", "bar"}},
		{"FOO_BAR_BAZ", []string{"foo", "bar", "baz"}},
		{"__leading__", []string{"leading"}},
		{"trailing__", []string{"trailing"}},
		{"a_b_c_d_e", []string{"a", "b", "c", "d", "e"}},
	}
	for _, c := range cases {
		t.Run(c.in, func(t *testing.T) {
			got := fuzzymatch.Tokenise(c.in, opts)
			if !reflect.DeepEqual(got, c.want) {
				t.Errorf("Tokenise(%q, default) = %q; want %q", c.in, got, c.want)
			}
		})
	}
}

// TestTokenise_KebabCase pins the hyphen-as-separator behaviour.
func TestTokenise_KebabCase(t *testing.T) {
	opts := fuzzymatch.DefaultTokeniseOptions()
	cases := []struct {
		in   string
		want []string
	}{
		{"foo-bar", []string{"foo", "bar"}},
		{"foo-bar-baz", []string{"foo", "bar", "baz"}},
		{"-leading-", []string{"leading"}},
	}
	for _, c := range cases {
		t.Run(c.in, func(t *testing.T) {
			got := fuzzymatch.Tokenise(c.in, opts)
			if !reflect.DeepEqual(got, c.want) {
				t.Errorf("Tokenise(%q, default) = %q; want %q", c.in, got, c.want)
			}
		})
	}
}

// TestTokenise_DotCase pins the dot-as-separator behaviour. Dot-case is
// the dominant shape for Cassandra column paths and similar
// hierarchical identifiers.
func TestTokenise_DotCase(t *testing.T) {
	opts := fuzzymatch.DefaultTokeniseOptions()
	cases := []struct {
		in   string
		want []string
	}{
		{"foo.bar", []string{"foo", "bar"}},
		{"foo.bar.baz", []string{"foo", "bar", "baz"}},
		{"a.b.c", []string{"a", "b", "c"}},
		{".leading.", []string{"leading"}},
	}
	for _, c := range cases {
		t.Run(c.in, func(t *testing.T) {
			got := fuzzymatch.Tokenise(c.in, opts)
			if !reflect.DeepEqual(got, c.want) {
				t.Errorf("Tokenise(%q, default) = %q; want %q", c.in, got, c.want)
			}
		})
	}
}

// TestTokenise_PascalCase pins the PascalCase shape: the camelCase
// boundary rule fires the same way; the leading uppercase doesn't get
// special treatment.
func TestTokenise_PascalCase(t *testing.T) {
	opts := fuzzymatch.DefaultTokeniseOptions()
	cases := []struct {
		in   string
		want []string
	}{
		{"FooBarBaz", []string{"foo", "bar", "baz"}},
		{"UserCreateEvent", []string{"user", "create", "event"}},
		{"ABCDef", []string{"abc", "def"}},
	}
	for _, c := range cases {
		t.Run(c.in, func(t *testing.T) {
			got := fuzzymatch.Tokenise(c.in, opts)
			if !reflect.DeepEqual(got, c.want) {
				t.Errorf("Tokenise(%q, default) = %q; want %q", c.in, got, c.want)
			}
		})
	}
}

// TestTokenise_ConsecutiveUpper locks the trailing-uppercase-boundary
// rule. With SplitConsecutiveUpper=true (the default), the last
// uppercase of a >=2-rune upper run when followed by a lowercase letter
// joins the next token: XMLParser -> ["xml", "parser"]. Earlier within-
// run upper-upper transitions do NOT split — XMLHTTPRequest produces
// ["xmlhttp", "request"] not ["xml", "http", "request"], because only
// the trailing P->R boundary has a following lowercase.
func TestTokenise_ConsecutiveUpper(t *testing.T) {
	defaults := fuzzymatch.DefaultTokeniseOptions()
	noTrailing := defaults
	noTrailing.SplitConsecutiveUpper = false
	cases := []struct {
		name string
		in   string
		opts fuzzymatch.TokeniseOptions
		want []string
	}{
		// SplitConsecutiveUpper=true (default).
		{"XMLParser-default", "XMLParser", defaults, []string{"xml", "parser"}},
		{"XMLHTTPRequest-default", "XMLHTTPRequest", defaults, []string{"xmlhttp", "request"}},
		{"IPv4Address-default", "IPv4Address", defaults, []string{"i", "pv4", "address"}},

		// SplitConsecutiveUpper=false: the trailing rule is disabled, so
		// the entire upper run sticks with its following lowercase chunk.
		{"XMLParser-no-trailing", "XMLParser", noTrailing, []string{"xmlparser"}},
		{"XMLHTTPRequest-no-trailing", "XMLHTTPRequest", noTrailing, []string{"xmlhttprequest"}},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			got := fuzzymatch.Tokenise(c.in, c.opts)
			if !reflect.DeepEqual(got, c.want) {
				t.Errorf("Tokenise(%q, %s) = %q; want %q", c.in, c.name, got, c.want)
			}
		})
	}
}

// TestTokenise_MixedScript locks the mixed-script behaviour. Cyrillic
// letters DO have case (unicode.IsLower / unicode.IsUpper return true
// for them), so a Latin lowercase followed by a Cyrillic uppercase
// triggers a camelCase boundary at the script transition. CJK and
// Arabic have no case so they pass through as single tokens.
func TestTokenise_MixedScript(t *testing.T) {
	opts := fuzzymatch.DefaultTokeniseOptions()
	cases := []struct {
		name string
		in   string
		want []string
	}{
		{"user-then-Privet", "userПривет", []string{"user", "привет"}},
		{"foo-Privet-Baz", "fooПриветBaz", []string{"foo", "привет", "baz"}},
		{"arabic-passthrough", "مرحبا", []string{"مرحبا"}},
		{"cjk-passthrough", "你好", []string{"你好"}},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			got := fuzzymatch.Tokenise(c.in, opts)
			if !reflect.DeepEqual(got, c.want) {
				t.Errorf("Tokenise(%q, default) = %q; want %q", c.in, got, c.want)
			}
		})
	}
}

// TestTokenise_Numeric pins the numeric-attachment behaviour: digits
// stay attached to the preceding alphabetic run, but a digit followed
// by an uppercase letter IS a boundary (rule 1's non-uppercase->
// uppercase fires on digit prev too). So Foo123Bar splits cleanly at
// the 3->B transition; parseJSON5 keeps the trailing 5 attached to
// json.
func TestTokenise_Numeric(t *testing.T) {
	opts := fuzzymatch.DefaultTokeniseOptions()
	cases := []struct {
		in   string
		want []string
	}{
		{"Foo123Bar", []string{"foo123", "bar"}},
		{"parse5JSON", []string{"parse5", "json"}},
		{"HTTP_REQUEST_V2", []string{"http", "request", "v2"}},
		{"abc123", []string{"abc123"}},
		{"123abc", []string{"123abc"}},
	}
	for _, c := range cases {
		t.Run(c.in, func(t *testing.T) {
			got := fuzzymatch.Tokenise(c.in, opts)
			if !reflect.DeepEqual(got, c.want) {
				t.Errorf("Tokenise(%q, default) = %q; want %q", c.in, got, c.want)
			}
		})
	}
}

// TestTokenise_NoEmptyTokens locks the no-empty-tokens invariant
// across the pathological-separator shapes: consecutive separators,
// leading/trailing separators, mixed separators all collapse without
// producing zero-width tokens.
func TestTokenise_NoEmptyTokens(t *testing.T) {
	opts := fuzzymatch.DefaultTokeniseOptions()
	cases := []struct {
		in   string
		want []string
	}{
		{"__foo__bar__", []string{"foo", "bar"}},
		{"--foo--bar--", []string{"foo", "bar"}},
		{"..foo..bar..", []string{"foo", "bar"}},
		{"_-._foo_-._bar_-._", []string{"foo", "bar"}},
		{"___", []string{}}, // separator-only input -> empty slice
	}
	for _, c := range cases {
		t.Run(c.in, func(t *testing.T) {
			got := fuzzymatch.Tokenise(c.in, opts)
			if got == nil {
				t.Errorf("Tokenise(%q, default) returned nil; want non-nil slice", c.in)
			}
			if !reflect.DeepEqual(got, c.want) {
				t.Errorf("Tokenise(%q, default) = %q; want %q", c.in, got, c.want)
			}
			for i, tok := range got {
				if tok == "" {
					t.Errorf("Tokenise(%q, default) emitted empty token at index %d", c.in, i)
				}
			}
		})
	}
}

// ---------------------------------------------------------------------
// Property tests (testing/quick)
// ---------------------------------------------------------------------

// TestProp_Tokenise_OrderStable asserts that calling Tokenise twice on
// the same (s, opts) returns byte-identical slices in identical order.
// This rules out any future regression introducing map iteration on
// the output path (DET-03). The check is run for arbitrary input
// strings and arbitrary option bitfields.
func TestProp_Tokenise_OrderStable(t *testing.T) {
	f := func(s string, optBits uint8) bool {
		opts := tokeniseOptsFromBits(optBits)
		first := fuzzymatch.Tokenise(s, opts)
		second := fuzzymatch.Tokenise(s, opts)
		return reflect.DeepEqual(first, second)
	}
	if err := quick.Check(f, nil); err != nil {
		t.Errorf("Tokenise order is not stable across calls: %v", err)
	}
}

// TestProp_Tokenise_NeverPanics asserts that Tokenise never panics for
// any (s, optBits) over the 2^4 option bitfield. testing/quick's string
// generator covers valid UTF-8; invalid UTF-8 panic-freedom is covered
// separately by FuzzTokenise (which generates arbitrary byte slices
// directly).
func TestProp_Tokenise_NeverPanics(t *testing.T) {
	f := func(s string, optBits uint8) (ok bool) {
		defer func() {
			if r := recover(); r != nil {
				t.Errorf("Tokenise panicked on s=%q optBits=%d: %v", s, optBits, r)
				ok = false
			}
		}()
		opts := tokeniseOptsFromBits(optBits)
		_ = fuzzymatch.Tokenise(s, opts)
		return true
	}
	if err := quick.Check(f, nil); err != nil {
		t.Errorf("Tokenise panicked under quick.Check: %v", err)
	}
}

// TestProp_Tokenise_NoEmptyTokens asserts that under default options
// no token in the returned slice is empty. The implementation filters
// out zero-width tokens at every boundary; this property pins that
// invariant against future refactors.
func TestProp_Tokenise_NoEmptyTokens(t *testing.T) {
	opts := fuzzymatch.DefaultTokeniseOptions()
	f := func(s string) bool {
		tokens := fuzzymatch.Tokenise(s, opts)
		for _, tok := range tokens {
			if tok == "" {
				return false
			}
		}
		return true
	}
	if err := quick.Check(f, nil); err != nil {
		t.Errorf("Tokenise emitted an empty token: %v", err)
	}
}

// TestProp_Tokenise_AllOutputUTF8 asserts that every returned token is
// valid UTF-8. This is trivially true on the Go side because Tokenise
// operates on []rune (which is valid Unicode by construction); the
// property pins the guarantee against any future byte-level refactor.
func TestProp_Tokenise_AllOutputUTF8(t *testing.T) {
	f := func(s string, optBits uint8) bool {
		opts := tokeniseOptsFromBits(optBits)
		tokens := fuzzymatch.Tokenise(s, opts)
		for _, tok := range tokens {
			if !utf8.ValidString(tok) {
				return false
			}
		}
		return true
	}
	if err := quick.Check(f, nil); err != nil {
		t.Errorf("Tokenise produced invalid UTF-8: %v", err)
	}
}

// TestProp_Tokenise_TokenCount_LessOrEqualInputRunes asserts that the
// returned slice has at most as many tokens as the input has runes.
// Each token has at least one rune (NoEmptyTokens), tokens are
// non-overlapping, and separator runes are dropped — so the upper
// bound on token count is the input rune count. This pins the
// invariant against any future rule that might over-split.
func TestProp_Tokenise_TokenCount_LessOrEqualInputRunes(t *testing.T) {
	f := func(s string, optBits uint8) bool {
		opts := tokeniseOptsFromBits(optBits)
		tokens := fuzzymatch.Tokenise(s, opts)
		return len(tokens) <= utf8.RuneCountInString(s)
	}
	if err := quick.Check(f, nil); err != nil {
		t.Errorf("Tokenise produced more tokens than input runes: %v", err)
	}
}

// TestProp_Tokenise_ReconstructibleASCII asserts a loose
// reconstructibility invariant: for ASCII input restricted to
// alphabetic characters (no digits, no separators, no case mixing) the
// joined-on-space token output is a subsequence of the lowercased
// input. The strict "joining tokens recovers the input" property does
// not hold because the camelCase rule discards no characters but
// inserts no separator either — the join() form differs by inserted
// spaces. The looser property "every output character appears in the
// lowercased input in order" pins the no-character-loss guarantee.
func TestProp_Tokenise_ReconstructibleASCII(t *testing.T) {
	opts := fuzzymatch.DefaultTokeniseOptions()
	f := func(input asciiAlpha) bool {
		s := string(input)
		tokens := fuzzymatch.Tokenise(s, opts)
		joined := strings.Join(tokens, "")
		// The joined form is the lowercased input minus the
		// separators (there are none in asciiAlpha) and minus any
		// inserted boundary characters (none — Tokenise never
		// inserts characters, only splits).
		want := strings.ToLower(s)
		return joined == want
	}
	if err := quick.Check(f, &quick.Config{MaxCount: 200}); err != nil {
		t.Errorf("Tokenise lost characters on ASCII alpha input: %v", err)
	}
}

// ---------------------------------------------------------------------
// Property-test helpers
// ---------------------------------------------------------------------

// tokeniseOptsFromBits maps a uint8 to a TokeniseOptions value. Three
// bool fields fit in 3 bits; SeparatorChars is fixed at the default to
// avoid combinatorial noise (varying it doesn't exercise new code
// paths — the byte path is gated on the [128]bool table built from
// SeparatorChars; the Unicode path uses strings.ContainsRune over the
// string).
func tokeniseOptsFromBits(bits uint8) fuzzymatch.TokeniseOptions {
	return fuzzymatch.TokeniseOptions{
		Lowercase:             bits&1 != 0,
		SplitCamelCase:        bits&2 != 0,
		SplitConsecutiveUpper: bits&4 != 0,
		SeparatorChars:        fuzzymatch.DefaultTokeniseOptions().SeparatorChars,
	}
}

// asciiAlpha is a quick.Generator-compatible type producing ASCII-only
// alphabetic strings (no digits, no separators, no whitespace) of
// arbitrary length up to ~50 runes. It implements quick.Generator so
// TestProp_Tokenise_ReconstructibleASCII can stay within a small input
// shape where the join-recovery property holds.
type asciiAlpha string

// Generate implements quick.Generator for asciiAlpha. It draws a length
// in [0, 50] and fills with random ASCII alphabetic runes (a-z and
// A-Z). The empty string is a valid output (length 0).
func (asciiAlpha) Generate(r *rand.Rand, _ int) reflect.Value {
	n := r.Intn(51)
	buf := make([]byte, n)
	for i := range buf {
		// 52 letters total: 26 lowercase + 26 uppercase. The first 26
		// values produce a lowercase letter; the next 26 produce an
		// uppercase letter.
		v := r.Intn(52)
		if v < 26 {
			buf[i] = byte('a' + v)
		} else {
			buf[i] = byte('A' + (v - 26))
		}
	}
	return reflect.ValueOf(asciiAlpha(buf))
}
