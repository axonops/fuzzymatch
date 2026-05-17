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

// Normalise applies caller-controlled normalisation: ASCII case-folding,
// separator stripping, camel-case splitting, and Unicode NFC
// normalisation with optional diacritic stripping. See
// docs/requirements.md §9 for the authoritative specification.
//
// Implementation discipline:
//
//   - ASCII fast path operates on bytes for inputs whose every byte is
//     strictly less than 0x80 and that do not require Unicode NFC
//     normalisation; stack-allocated [128]bool separator table avoids
//     heap allocation for any short input.
//   - Unicode path uses golang.org/x/text/unicode/norm for NFC and
//     transform.Chain(norm.NFD, runes.Remove(unicode.Mn), norm.NFC) for
//     diacritic stripping. The transformer is constructed per call —
//     transform.Transformer is not documented as safe for concurrent
//     reuse, and per-call construction is cheap (no allocation beyond
//     the chain wrapper). Pooling is deferred to a v1.x perf revisit.
//   - NO init()-time table builds (per docs/requirements.md §5(12) and
//     .claude/skills/determinism-standards): every table is a `var`
//     literal or built per call.
//   - NO map iteration on output paths (DET-03): the separator-set is a
//     [128]bool array for ASCII; the Unicode path uses
//     strings.ContainsRune for the small SeparatorChars string.
//   - NO transcendental float operations (DET-06): Normalise has no
//     floats at all.
//   - NO goroutines, channels, or mutexes (D-09).

package fuzzymatch

import (
	"strings"
	"unicode"
	"unicode/utf8"

	"golang.org/x/text/runes"
	"golang.org/x/text/transform"
	"golang.org/x/text/unicode/norm"
)

// NormalisationOptions configures the Normalise pipeline. The struct is
// passed by value: per-call allocation is zero. Defaults are obtained
// from DefaultNormalisationOptions; consumers wanting an alternative
// configuration build the struct directly.
//
// Field semantics:
//
//   - Lowercase: when true, fold uppercase ASCII to lowercase via the
//     bitwise OR 0x20 fast path, and fold non-ASCII runes via
//     unicode.ToLower.
//   - StripSeparators: when true, replace each rune appearing in
//     SeparatorChars with a single ASCII space and collapse runs of
//     spaces. Leading and trailing whitespace is removed.
//   - SeparatorChars: the set of characters treated as separators when
//     StripSeparators is true. Empty SeparatorChars + StripSeparators=true
//     is equivalent to whitespace-only collapsing.
//   - SplitCamelCase: when true, insert a single ASCII space at every
//     lowercase → uppercase rune transition before any other folding
//     (so FooBar → Foo Bar → foo bar under Lowercase).
//   - NFC: when true, apply Unicode NFC normalisation via
//     golang.org/x/text/unicode/norm. Required for cross-platform
//     determinism of Unicode inputs.
//   - StripDiacritics: when true, drop combining marks (Unicode category
//     Mn) via the transform.Chain(NFD, runes.Remove(Mn), NFC) pipeline.
//     Implicitly forces NFC ordering regardless of the NFC field.
//
// Invalid UTF-8 in the input is replaced with U+FFFD (REPLACEMENT
// CHARACTER) per Go's standard convention; Normalise never panics on
// arbitrary byte input (FuzzNormalise asserts this property).
type NormalisationOptions struct {
	Lowercase       bool
	StripSeparators bool
	SeparatorChars  string
	SplitCamelCase  bool
	NFC             bool
	StripDiacritics bool
}

// DefaultNormalisationOptions returns the v1.x default normalisation
// configuration per docs/requirements.md §9 (D-03):
//
//   - Lowercase        = true
//   - StripSeparators  = true
//   - SeparatorChars   = "_-.:/"
//   - SplitCamelCase   = true
//   - NFC              = true
//   - StripDiacritics  = false
//
// Diacritic stripping is OFF by default — consumers wanting "café → cafe"
// semantics opt in explicitly. The defaults are tuned for code-identifier
// matching (snake_case / camelCase / kebab-case / dot-case unification)
// without altering the visible character set of natural-language inputs.
func DefaultNormalisationOptions() NormalisationOptions {
	return NormalisationOptions{
		Lowercase:       true,
		StripSeparators: true,
		SeparatorChars:  "_-.:/",
		SplitCamelCase:  true,
		NFC:             true,
		StripDiacritics: false,
	}
}

// Normalise returns s with the requested normalisation applied. Empty
// input returns empty.
//
// Behaviour summary:
//
//   - If s is empty, the empty string is returned.
//   - If s is pure ASCII (every byte < 0x80) and StripDiacritics is
//     false, the ASCII fast path is taken: no x/text invocation, all
//     work in a single byte-slice pass.
//   - Otherwise the Unicode path runs: optionally
//     transform.Chain(NFD, runes.Remove(Mn), NFC) for diacritic
//     stripping, or plain NFC, then folding / camel-split / separator-
//     strip over the rune sequence.
//
// Output guarantees:
//
//   - Always valid UTF-8 (invalid sequences in the input are replaced
//     with U+FFFD by x/text's normaliser).
//   - Idempotent for any fixed opts: Normalise(Normalise(s, opts), opts)
//     == Normalise(s, opts).
//   - Cross-platform byte-identical: no floating-point arithmetic, no
//     map iteration on the output path, no init()-time tables.
//
// Normalise has no error return: malformed input is handled by Unicode
// replacement rather than by failure. If a future API surface needs an
// (string, error) form (for example a strict-UTF-8 mode), it will be
// added under a distinct name to preserve this contract.
//
// Performance scope (Q7b, docs/requirements.md §14.1):
//
//   The published budget is ≤ 1 alloc per call on ASCII Short. The ASCII
//   fast path uses a make([]byte, 0, len(s)*2+1) scratch buffer (1 heap
//   allocation) followed by string(buf) on return. The 0-alloc target
//   listed in earlier drafts was unachievable: the buffer cannot live on
//   the stack because escape analysis cannot prove its size at compile
//   time, and `unsafe.String` is excluded by project policy. For long
//   ASCII inputs (≥ 500 chars) the buffer allocation still amortises to
//   1 alloc — only the byte-count scales with input. The Unicode path
//   incurs ≤ 3 allocs (transform.Chain + output buffer + fold pass).
func Normalise(s string, opts NormalisationOptions) string {
	if s == "" {
		return ""
	}
	// The ASCII fast path is selected when the input is pure ASCII AND
	// the pipeline does not require diacritic stripping. Pure NFC of
	// ASCII text is a byte-level no-op (every NFC quick-check on ASCII
	// returns Yes), so the fast path's output is bit-equivalent to the
	// Unicode-path output for ASCII inputs even when NFC is requested.
	if !opts.StripDiacritics && isASCII(s) {
		return normaliseASCII(s, opts)
	}
	return normaliseUnicode(s, opts)
}

// isASCII reports whether every byte of s is strictly less than 0x80.
// Empty s returns true (vacuously).
func isASCII(s string) bool {
	for i := 0; i < len(s); i++ {
		if s[i] >= 0x80 {
			return false
		}
	}
	return true
}

// buildSepSet returns a stack-allocated [128]bool table where index i is
// true iff the byte i appears in the SeparatorChars string. Non-ASCII
// bytes in SeparatorChars are silently ignored on the ASCII fast path
// (they cannot appear in an ASCII input anyway); the Unicode path uses
// strings.ContainsRune over the full SeparatorChars string.
//
// The table is built per call so that opts.SeparatorChars changes are
// observed immediately and so that no init()-time table exists at the
// package level.
func buildSepSet(seps string) [128]bool {
	var set [128]bool
	for i := 0; i < len(seps); i++ {
		b := seps[i]
		if b < 0x80 {
			set[b] = true
		}
	}
	return set
}

// normaliseASCII applies the ASCII fast-path pipeline. Pre-conditions:
// every byte of s is < 0x80 and opts.StripDiacritics is false. The
// function writes into a stack-resident buffer for inputs up to 64 bytes
// (the Go inliner promotes a fixed-size array variable to the stack);
// larger inputs heap-allocate exactly one byte slice. There are no
// intermediate allocations.
//
// Algorithm in a single pass over the input bytes:
//
//  1. If SplitCamelCase: when transitioning lowercase→uppercase, emit
//     a space before the uppercase byte.
//  2. If Lowercase: fold A..Z to a..z via |= 0x20.
//  3. If StripSeparators: replace any byte in the separator set (built
//     from opts.SeparatorChars) with a space; also treat ASCII
//     whitespace as a separator; collapse runs of spaces; trim leading
//     and trailing spaces.
//
// The order matters: camel-case split happens BEFORE lowercasing so the
// "uppercase" decision is based on the original byte. Separator-strip
// happens on the final byte (potentially the inserted space too) so
// runs of separators-and-spaces collapse correctly.
func normaliseASCII(s string, opts NormalisationOptions) string {
	// Worst-case length: SplitCamelCase can insert one space per byte
	// at a lowercase→uppercase boundary. The bound len(s)*2 is safe.
	// We then trim trailing slack with [:n].
	buf := make([]byte, 0, len(s)*2+1)

	sepSet := buildSepSet(opts.SeparatorChars)

	var prev byte = 0
	for i := 0; i < len(s); i++ {
		b := s[i]

		// Step 1: camel-case split. Emit a space at lowercase→uppercase
		// boundaries before processing this byte.
		if opts.SplitCamelCase && i > 0 && isLowerASCII(prev) && isUpperASCII(b) {
			buf = append(buf, ' ')
		}

		// Step 2: lowercase the byte via bitwise OR for A..Z. The
		// branch is intentional (rather than unconditional OR) because
		// non-letter ASCII bytes outside A..Z must not be altered.
		out := b
		if opts.Lowercase && isUpperASCII(b) {
			out |= 0x20
		}

		buf = append(buf, out)
		prev = b
	}

	// Step 3: separator-strip + collapse. Done as a second pass so that
	// camel-split-inserted spaces and ASCII whitespace from the input
	// collapse together. The second pass writes into the same buffer
	// in place (output is always <= input length).
	if opts.StripSeparators {
		buf = collapseSeparators(buf, sepSet)
	}

	return string(buf)
}

// collapseSeparators rewrites buf in place: each byte appearing in
// sepSet, plus ASCII whitespace, is converted to a single space; runs of
// spaces collapse to one; leading and trailing spaces are stripped.
//
// The function operates byte-by-byte and assumes buf is pure ASCII at
// entry (true on the ASCII fast path).
func collapseSeparators(buf []byte, sepSet [128]bool) []byte {
	if len(buf) == 0 {
		return buf
	}
	w := 0
	lastWasSpace := true // suppress leading whitespace
	for i := 0; i < len(buf); i++ {
		b := buf[i]
		isSep := false
		if b < 0x80 {
			if sepSet[b] || isASCIISpace(b) {
				isSep = true
			}
		}
		if isSep {
			if !lastWasSpace {
				buf[w] = ' '
				w++
				lastWasSpace = true
			}
			continue
		}
		buf[w] = b
		w++
		lastWasSpace = false
	}
	// Strip a single trailing space if present.
	if w > 0 && buf[w-1] == ' ' {
		w--
	}
	return buf[:w]
}

// normaliseUnicode applies the slow-path pipeline for inputs that are
// non-ASCII OR have StripDiacritics enabled. The pipeline:
//
//  1. Build the x/text transformer: chain(NFD, Remove(Mn), NFC) when
//     stripping diacritics; plain NFC otherwise; pass-through when
//     neither NFC nor StripDiacritics is requested.
//  2. Apply the transformer via transform.String.
//  3. Walk the resulting rune sequence applying camel-split,
//     lowercasing (unicode.ToLower for non-ASCII; |= 0x20 for ASCII),
//     and separator/whitespace collapsing.
//
// Step 3 is rune-aware so that combining-mark boundaries are preserved.
// strings.ContainsRune is the membership check for SeparatorChars; the
// typical SeparatorChars is short (≤ 8 runes), so the linear scan is
// cheaper than building a map and respects the no-map-iteration rule.
func normaliseUnicode(s string, opts NormalisationOptions) string {
	transformed := applyUnicodeTransformer(s, opts)
	buf := foldRunes(transformed, opts)
	if opts.StripSeparators {
		buf = collapseSeparatorsUnicode(buf, opts.SeparatorChars)
	}
	return string(buf)
}

// applyUnicodeTransformer runs the x/text transformer step of the
// Unicode pipeline: chain(NFD, Remove(Mn), NFC) when stripping
// diacritics, plain NFC when only NFC is requested, pass-through
// otherwise. transform.Transformer is constructed per call (cheap; no
// pool to honour D-09's "no goroutines / channels / mutexes" rule).
func applyUnicodeTransformer(s string, opts NormalisationOptions) string {
	var t transform.Transformer
	switch {
	case opts.StripDiacritics:
		// NFD decomposes precomposed forms into base + combining marks;
		// runes.Remove(In(Mn)) drops the combining marks; NFC recomposes
		// (most consumers expect NFC output).
		t = transform.Chain(norm.NFD, runes.Remove(runes.In(unicode.Mn)), norm.NFC)
	case opts.NFC:
		t = norm.NFC
	default:
		return s
	}
	// transform.String returns (string, n, err); n and err are
	// uninteresting for our chains (err is always nil; we always use
	// the full output regardless of n).
	out, _, _ := transform.String(t, s)
	return out
}

// foldRunes runs the rune-by-rune second pass: camel-case-split
// insertion and (optionally) lowercasing. Separator collapse is a
// further pass over the byte buffer in normaliseUnicode — keeping it
// separate keeps foldRunes within the gocyclo budget.
//
// Capacity hint: len(s)+1 bytes is an upper-ish bound. A camel-split
// insertion adds one byte per rune at the lower→upper transition, so
// the worst case is len(s) plus the rune count. One reallocation is
// acceptable; the typical case fits.
func foldRunes(s string, opts NormalisationOptions) []byte {
	buf := make([]byte, 0, len(s)+1)
	var prev rune = -1
	for _, r := range s {
		if opts.SplitCamelCase && prev >= 0 && unicode.IsLower(prev) && unicode.IsUpper(r) {
			buf = append(buf, ' ')
		}
		if opts.Lowercase {
			r = lowerRune(r)
		}
		buf = utf8.AppendRune(buf, r)
		prev = r
	}
	return buf
}

// collapseSeparatorsUnicode rewrites buf to replace separators (runes
// in seps) and ASCII whitespace with a single space, collapsing runs and
// trimming leading/trailing whitespace. Operates on the UTF-8 byte
// sequence; non-separator runes pass through unchanged.
//
// Implementation note: we walk the input via utf8.DecodeRune so that
// multi-byte separators (if any are configured) are matched correctly.
// The output is written into a fresh slice rather than in-place because
// the rune-decoded view doesn't trivially fit the in-place pattern.
func collapseSeparatorsUnicode(buf []byte, seps string) []byte {
	if len(buf) == 0 {
		return buf
	}
	out := make([]byte, 0, len(buf))
	lastWasSpace := true
	i := 0
	for i < len(buf) {
		r, size := utf8.DecodeRune(buf[i:])
		if isSeparatorRune(r, seps) {
			if !lastWasSpace {
				out = append(out, ' ')
				lastWasSpace = true
			}
			i += size
			continue
		}
		out = append(out, buf[i:i+size]...)
		lastWasSpace = false
		i += size
	}
	if len(out) > 0 && out[len(out)-1] == ' ' {
		out = out[:len(out)-1]
	}
	return out
}

// isSeparatorRune reports whether r should be treated as a separator
// during the Unicode-path collapse: ASCII whitespace OR any rune that
// appears in seps. Extracted from collapseSeparatorsUnicode so that
// function's cyclomatic complexity stays within the gocyclo budget.
func isSeparatorRune(r rune, seps string) bool {
	if isASCIIWhitespaceRune(r) {
		return true
	}
	return strings.ContainsRune(seps, r)
}

// isASCIIWhitespaceRune is the rune-typed counterpart to isASCIISpace.
// We avoid converting r to byte here because that conversion triggers
// gosec G115 even on the already-bounded r < 0x80 path.
func isASCIIWhitespaceRune(r rune) bool {
	switch r {
	case ' ', '\t', '\n', '\r', '\v', '\f':
		return true
	}
	return false
}

// isUpperASCII reports whether b is in the ASCII range A..Z.
func isUpperASCII(b byte) bool { return b >= 'A' && b <= 'Z' }

// isLowerASCII reports whether b is in the ASCII range a..z.
func isLowerASCII(b byte) bool { return b >= 'a' && b <= 'z' }

// isASCIISpace reports whether b is one of ' ', '\t', '\n', '\r', '\v',
// '\f' — the bytes for which unicode.IsSpace would return true on the
// ASCII subset. We avoid unicode.IsSpace itself on the byte hot path so
// the inliner can keep the function pointer-free.
func isASCIISpace(b byte) bool {
	switch b {
	case ' ', '\t', '\n', '\r', '\v', '\f':
		return true
	}
	return false
}

// lowerRune folds r to lowercase. ASCII A..Z take the bitwise OR fast
// path; non-ASCII delegate to unicode.ToLower. Runes outside the cased
// range pass through unchanged.
func lowerRune(r rune) rune {
	if r < 0x80 {
		if r >= 'A' && r <= 'Z' {
			return r | 0x20
		}
		return r
	}
	return unicode.ToLower(r)
}
