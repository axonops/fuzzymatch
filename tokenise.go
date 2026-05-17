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

// Tokenise splits s into tokens by camelCase, snake_case, PascalCase,
// kebab-case, and dot-case boundaries. See docs/requirements.md §10
// for the authoritative spec.
//
// Token order matches input order (left-to-right byte order). Empty or
// whitespace-only input returns a non-nil empty slice. Consecutive
// separators never produce empty tokens.
//
// Implementation discipline:
//
//   - NO init()-time table builds (per docs/requirements.md §5(12) and
//     .claude/skills/determinism-standards): the [128]bool separator set
//     is built per call from opts.SeparatorChars.
//   - NO map iteration on output paths (DET-03): all internal data
//     structures are slices or arrays.
//   - NO transcendental float operations (DET-06): Tokenise has no
//     floats at all.
//   - NO goroutines, channels, or mutexes (D-09).
//
// Tokenise is the second consumer of the no-map-iteration discipline
// (after Normalise in plan 01-06) and the second public-API primitive
// in the fuzzymatch surface. Phase 6's five token-based algorithms
// (Monge-Elkan, TokenSortRatio, TokenSetRatio, PartialRatio,
// TokenJaccard) compose against Tokenise's output as the canonical
// token shape.

package fuzzymatch

import (
	"strings"
	"unicode"
	"unicode/utf8"
)

// TokeniseOptions configures the Tokenise splitter. The struct is passed
// by value: per-call allocation is zero. Defaults are obtained from
// DefaultTokeniseOptions; consumers wanting an alternative configuration
// build the struct directly.
//
// Field semantics:
//
//   - Lowercase: when true, fold each emitted token to lowercase. ASCII
//     bytes use the bitwise OR 0x20 fast path; non-ASCII runes go via
//     unicode.ToLower. Folding is applied AFTER tokenisation so that
//     uppercase-letter-driven camelCase boundaries are preserved (per
//     docs/requirements.md §10 step 5).
//   - SplitCamelCase: when true, insert a token boundary at every
//     uppercase rune that follows a non-uppercase rune (a lowercase
//     letter OR a digit). "fooBar" -> ["foo", "bar"];
//     "Foo123Bar" -> ["foo123", "bar"] — digits stay attached to the
//     preceding alpha run but a subsequent uppercase letter starts a
//     new token.
//   - SplitConsecutiveUpper: when true (and SplitCamelCase is also true),
//     insert a boundary at the last uppercase of a >=2-rune uppercase
//     run when followed by a lowercase letter: the last uppercase
//     joins the next token. ("XMLParser" -> ["xml", "parser"];
//     "XMLHTTPRequest" -> ["xmlhttp", "request"] — the boundary
//     fires exactly once per run, at the run's trailing edge before
//     the following lowercase letter). When false the consecutive-
//     upper run stays cohesive with its trailing lowercase
//     ("XMLParser" -> ["xmlparser"]).
//   - SeparatorChars: characters treated as token separators. Each
//     occurrence is a split point AND is discarded. Default
//     "_-.:/ \t\n\r" covers snake_case, kebab-case, dot-case, slash-
//     path-case, and ASCII whitespace.
//
// Invalid UTF-8 in the input is replaced with U+FFFD (REPLACEMENT
// CHARACTER) per Go's standard convention; Tokenise never panics on
// arbitrary byte input (FuzzTokenise asserts this property).
type TokeniseOptions struct {
	Lowercase             bool
	SplitCamelCase        bool
	SplitConsecutiveUpper bool
	SeparatorChars        string
}

// DefaultTokeniseOptions returns the v1.x default tokenisation
// configuration per docs/requirements.md §10:
//
//   - Lowercase             = true
//   - SplitCamelCase        = true
//   - SplitConsecutiveUpper = true
//   - SeparatorChars        = "_-.:/ \t\n\r"
//
// The defaults are tuned for code-identifier matching: snake_case,
// camelCase, kebab-case, dot-case, and slash-path-case all split into
// the same lowercase token sequence under these defaults. ASCII
// whitespace is included in SeparatorChars so already-tokenised inputs
// ("foo bar baz") split correctly without requiring a pre-normalisation
// step.
func DefaultTokeniseOptions() TokeniseOptions {
	return TokeniseOptions{
		Lowercase:             true,
		SplitCamelCase:        true,
		SplitConsecutiveUpper: true,
		SeparatorChars:        "_-.:/ \t\n\r",
	}
}

// Tokenise returns the tokens of s per the configured options. Empty or
// whitespace-only input returns a non-nil empty slice. The token order
// matches the byte order of s; consecutive separators do not produce
// empty tokens.
//
// Behaviour summary:
//
//   - If s is empty, an empty non-nil slice is returned.
//   - If s is pure ASCII (every byte < 0x80) AND opts.SeparatorChars is
//     also pure ASCII, the ASCII fast path runs (see "Performance note"
//     below). The fast path produces byte-identical output to the rune
//     path on any ASCII input — verified by
//     TestProp_Tokenise_ASCIIFastPathEquivalent.
//   - Otherwise s is decoded to a []rune (one allocation) and walked
//     left-to-right. Each boundary (separator OR camelCase transition,
//     subject to opts) closes the current token and starts a new one.
//   - Empty tokens (produced by leading separators, trailing separators,
//     or consecutive separators) are filtered out before return.
//   - Lowercasing is applied per-token AFTER boundary detection so that
//     the camelCase decisions see the original-case input.
//
// Output guarantees:
//
//   - Returned slice is never nil (an empty slice is returned for empty
//     or all-separator input).
//   - Every returned token has at least one rune (no empty strings).
//   - Every returned token is valid UTF-8 (invalid sequences in the
//     input become U+FFFD via Go's []rune conversion before splitting).
//   - Order is stable across calls (no map iteration on the output
//     path): two calls with identical (s, opts) return slices with
//     identical contents in identical order.
//
// Tokenise has no error return: malformed input is handled by Unicode
// replacement rather than by failure.
//
// Performance note (Phase 8.5 Q8b, docs/requirements.md §10 and §14.3):
//
// On ASCII input, the returned tokens are substrings of either the
// input string (when opts.Lowercase == false) or a single internal
// scratch buffer (when opts.Lowercase == true). The returned []string
// retains a reference to that backing array until every emitted token
// string is released by the consumer; for very large inputs this is
// a memory-retention consideration. Consumers who need the input
// string to be eligible for garbage collection can copy each token via
// strings.Clone (Go 1.18+).
//
// The achieved allocation budget is:
//
//   - opts.Lowercase == false on ASCII input: 1 allocation for the
//     []string header itself; zero allocations per emitted token
//     (substrings share the input's backing array).
//   - opts.Lowercase == true on ASCII input: 1 allocation for the
//     scratch buffer (independent of token count), plus 1 allocation
//     per emitted token for the canonical string([]byte) conversion.
//     Bringing the per-token alloc down to zero is technically
//     achievable via unsafe.String but is excluded by project policy
//     (correctness-first default; CLAUDE.md Architecture extends the
//     no-cgo / no-non-stdlib spirit to the unsafe package).
//   - Non-ASCII input: falls back to the rune-based path, which
//     allocates a []rune and one string per token (unchanged from
//     pre-Phase-8.5 behaviour).
//
// The token-tier algorithms (TokenSortRatio, TokenSetRatio,
// PartialRatio, TokenJaccard, MongeElkan) consume Tokenise output and
// inherit the fast-path savings on ASCII identifier inputs.
func Tokenise(s string, opts TokeniseOptions) []string {
	if s == "" {
		return []string{}
	}

	// ASCII fast path detection (Phase 8.5 Q8b). The fast path is taken
	// when every byte of s AND every byte of opts.SeparatorChars is
	// below 0x80. A non-ASCII separator string would force the rune
	// path even on an ASCII input because the rune path's
	// strings.ContainsRune check is the only place a multi-byte
	// separator can match; the fast path's [128]bool lookup cannot
	// represent multi-byte separators. Verified by
	// TestProp_Tokenise_ASCIIFastPathEquivalent.
	if isASCIITokenise(s) && !containsNonASCII(opts.SeparatorChars) {
		return tokeniseASCII(s, opts)
	}

	return tokeniseRune(s, opts)
}

// isASCIITokenise reports whether every byte of s is strictly less than
// 0x80. Empty s returns true (vacuously). Mirrors isASCII in
// normalise.go but is named distinctly so each primitive's ASCII gate
// is locally readable (the two helpers are intentionally not shared so
// tokenise.go has no internal dependency on normalise.go, per
// .claude/skills/go-coding-standards).
//
// The byte loop is the fast pattern (one comparison per byte, no
// rune-decoding cost); a `range s` form would decode every byte to
// a rune which is materially slower for the strict-ASCII detection
// case. Profiling on Phase 1's Normalise fast path confirmed this.
func isASCIITokenise(s string) bool {
	for i := 0; i < len(s); i++ {
		if s[i] >= 0x80 {
			return false
		}
	}
	return true
}

// tokeniseASCII applies the byte-level Tokenise fast path. Pre-condition:
// every byte of s is < 0x80 AND opts.SeparatorChars contains only ASCII
// bytes. Output is byte-identical to tokeniseRune on any input that
// satisfies the pre-condition — verified by
// TestProp_Tokenise_ASCIIFastPathEquivalent.
//
// Two sub-modes, branched on opts.Lowercase:
//
//   - opts.Lowercase == false: emits tokens as s[lo:hi] substrings.
//     Every emitted token shares the input string's backing array
//     (zero-copy). One allocation total for the []string header.
//   - opts.Lowercase == true: copies s into a fresh scratch buffer
//     with A..Z folded to a..z via bitwise OR 0x20, then emits tokens
//     as substrings of the scratch buffer via string(scratch[lo:hi]).
//     One allocation for the scratch buffer, plus one allocation per
//     emitted token (the string conversion is structurally unavoidable
//     without unsafe.String, which is excluded by project policy).
//
// The boundary-detection logic mirrors tokeniseRune byte-for-byte; the
// only differences are (a) operating on bytes rather than runes,
// (b) using isUpperASCII / isLowerASCII (from normalise.go) rather
// than unicode.IsUpper / unicode.IsLower, and (c) the lowercase fold
// is applied to the scratch buffer before boundary detection so that
// substring slicing on the buffer produces lowercase tokens directly.
//
// CRITICAL: boundary detection MUST run against the pre-fold bytes
// (or operate on the original case info). The rune path defers
// lowercasing until appendToken — boundary decisions see original
// case. The ASCII fast path mirrors this by examining the SOURCE
// bytes (s) for boundary decisions even when emitting from the
// scratch buffer; the boundary indices in s match the buffer 1:1
// because the lowercase fold is a per-byte identity-preserving
// transform on length.
func tokeniseASCII(s string, opts TokeniseOptions) []string {
	sepASCII := buildTokeniseSepSet(opts.SeparatorChars)
	tokens := make([]string, 0, 4)

	// Build the scratch buffer once (lowercase mode only). The buffer
	// is byte-equal to s except that A..Z bytes are folded to a..z.
	// Length matches len(s) exactly; index space is shared with s.
	var scratch []byte
	if opts.Lowercase {
		scratch = make([]byte, len(s))
		for i := 0; i < len(s); i++ {
			b := s[i]
			if isUpperASCII(b) {
				scratch[i] = b | 0x20
			} else {
				scratch[i] = b
			}
		}
	}

	// emitToken appends s[lo:hi] (no-lowercase) or string(scratch[lo:hi])
	// (lowercase) to tokens. Empty ranges are silently dropped.
	emitToken := func(lo, hi int) {
		if lo >= hi {
			return
		}
		if opts.Lowercase {
			tokens = append(tokens, string(scratch[lo:hi]))
		} else {
			tokens = append(tokens, s[lo:hi])
		}
	}

	start := 0
	for i := 0; i < len(s); i++ {
		b := s[i]

		// Separator: close any pending token and skip the separator byte.
		if b < 0x80 && sepASCII[b] {
			emitToken(start, i)
			start = i + 1
			continue
		}

		// camelCase boundary detection. The byte-level analogue of
		// isCamelBoundary / isUpperRunTrailing — boundary decisions
		// are made against ORIGINAL-case bytes in s (matching the
		// rune path's "lowercase after boundary detection" semantic).
		if isCamelBoundaryASCII(s, i, start, opts) {
			emitToken(start, i)
			start = i
		}
	}

	// Flush any pending final token.
	emitToken(start, len(s))

	return tokens
}

// isCamelBoundaryASCII is the byte-level analogue of isCamelBoundary.
// Pre-condition: s[i] is NOT a separator (the caller has already
// handled that case) AND every byte of s is < 0x80. Returns true iff
// s[i] is the START of a new token under the camelCase /
// consecutive-uppercase rules in opts.
//
// Boundary rules (mirror isCamelBoundary byte-for-byte; the rune
// path's unicode.IsUpper / unicode.IsLower checks reduce to the ASCII
// helpers isUpperASCII / isLowerASCII when the input is pure ASCII):
//
//   - Rule 1 (non-uppercase -> uppercase): split before s[i] when
//     SplitCamelCase is set, s[i-1] is NOT uppercase, and s[i] IS
//     uppercase. Lowercase letters AND digits AND any other ASCII
//     non-letter both qualify as "non-uppercase" so that "Foo123Bar"
//     splits at the 3->B transition.
//   - Rule 2 (consecutive-uppercase trailing): split before s[i]
//     when SplitCamelCase and SplitConsecutiveUpper are both set,
//     s[i] is uppercase, s[i+1] exists and is lowercase, AND s[i-1]
//     is uppercase. The boundary fires once per >=2-uppercase run
//     at its trailing edge before the following lowercase letter.
//
// i must be > start (we don't emit empty tokens at the head of a
// run); the caller asserts this implicitly because i > start is part
// of every rule's precondition.
func isCamelBoundaryASCII(s string, i, start int, opts TokeniseOptions) bool {
	if !opts.SplitCamelCase || i == 0 || i <= start {
		return false
	}
	prev := s[i-1]
	cur := s[i]
	// Rule 1: non-uppercase -> uppercase.
	if !isUpperASCII(prev) && isUpperASCII(cur) {
		return true
	}
	// Rule 2: consecutive-uppercase trailing.
	if opts.SplitConsecutiveUpper && isUpperRunTrailingASCII(s, i, prev, cur) {
		return true
	}
	return false
}

// isUpperRunTrailingASCII is the byte-level analogue of
// isUpperRunTrailing. Reports whether s[i] is the trailing-uppercase
// boundary of a >=2-byte uppercase run followed by a lowercase byte.
// All three preconditions must hold:
//
//   - s[i] (cur) is uppercase
//   - s[i+1] exists and is lowercase
//   - s[i-1] (prev) is uppercase
func isUpperRunTrailingASCII(s string, i int, prev, cur byte) bool {
	if i+1 >= len(s) {
		return false
	}
	return isUpperASCII(cur) && isLowerASCII(s[i+1]) && isUpperASCII(prev)
}

// tokeniseRune is the original pre-Phase-8.5 rune-based Tokenise
// implementation, retained as an internal entry point so that
// TestProp_Tokenise_ASCIIFastPathEquivalent can compare the ASCII fast
// path against the rune path on ASCII inputs. Tokenise() above
// dispatches to this function on any non-ASCII input or non-ASCII
// SeparatorChars.
//
// Behaviour is identical to the pre-Phase-8.5 Tokenise: decode to
// []rune (one allocation), walk left-to-right, emit tokens at
// separator and camelCase boundaries, filter zero-width tokens,
// optionally lowercase per token.
func tokeniseRune(s string, opts TokeniseOptions) []string {
	if s == "" {
		return []string{}
	}

	// Build the ASCII separator table per call so that opts.SeparatorChars
	// changes are observed immediately and so that no init()-time table
	// exists at the package level. Non-ASCII separator chars are checked
	// via strings.ContainsRune in the loop body — separator strings
	// containing multibyte runes are supported but the ASCII fast lookup
	// is the common case.
	sepASCII := buildTokeniseSepSet(opts.SeparatorChars)
	hasNonASCIISep := containsNonASCII(opts.SeparatorChars)

	// Decode s to runes in a single allocation. Tokenise must operate on
	// runes (not bytes) so that camelCase boundary detection works on
	// Unicode case transitions (e.g. Cyrillic Latin-mixed identifiers).
	// Invalid UTF-8 bytes in s become U+FFFD per Go's []rune conversion.
	runes := []rune(s)
	if len(runes) == 0 {
		return []string{}
	}

	// Pre-size the result slice for the common case (short identifier
	// with 1..4 tokens). Larger inputs grow the slice naturally.
	tokens := make([]string, 0, 4)

	start := 0
	for i := 0; i < len(runes); i++ {
		r := runes[i]

		// Separator: close any pending token and skip the separator rune.
		if isSeparator(r, sepASCII, opts.SeparatorChars, hasNonASCIISep) {
			if i > start {
				tokens = appendToken(tokens, runes[start:i], opts.Lowercase)
			}
			start = i + 1
			continue
		}

		// camelCase boundary detection: lowercase->uppercase OR the
		// trailing-uppercase-of-a-run boundary (when
		// SplitConsecutiveUpper is enabled). Extracted into a helper so
		// the loop body stays under the gocyclo budget.
		if isCamelBoundary(runes, i, start, opts) {
			tokens = appendToken(tokens, runes[start:i], opts.Lowercase)
			start = i
		}
	}

	// Flush any pending final token.
	if start < len(runes) {
		tokens = appendToken(tokens, runes[start:], opts.Lowercase)
	}

	return tokens
}

// isCamelBoundary reports whether the rune at runes[i] is the START of a
// new token under the camelCase or consecutive-uppercase rules in opts.
// Pre-conditions: runes[i] is NOT a separator (the caller has already
// handled that case), and i is a valid index into runes.
//
// Boundary rules (per docs/requirements.md §10 step 3):
//
//   - non-uppercase -> uppercase: split before runes[i] when
//     SplitCamelCase is set, runes[i-1] is NOT uppercase, and runes[i]
//     IS uppercase. Lowercase letters AND digits both qualify as
//     "non-uppercase" so that "Foo123Bar" splits at the 3->B
//     transition into ["foo123", "bar"] — keeping digits attached to
//     the preceding alphabetic run but recognising the digit->upper
//     transition as a token boundary. "fooBar" -> ["foo", "bar"].
//     The i > start guard prevents firing immediately after a
//     separator (where runes[i-1] would be the discarded separator).
//   - consecutive-uppercase trailing: split before runes[i] when
//     SplitCamelCase and SplitConsecutiveUpper are both set, runes[i]
//     is uppercase, runes[i+1] is lowercase, AND runes[i-1] is also
//     uppercase ("XMLParser" -> ["XML", "Parser"];
//     "XMLHTTPRequest" -> ["XMLHTTP", "Request"] because only the
//     P->R transition satisfies the trailing rule; earlier within-run
//     transitions don't have a following lowercase).
//
// i must be > start (we don't emit empty tokens at the head of a run);
// the caller asserts this implicitly because i > start is part of every
// rule's precondition.
func isCamelBoundary(runes []rune, i, start int, opts TokeniseOptions) bool {
	if !opts.SplitCamelCase || i == 0 || i <= start {
		return false
	}
	prev := runes[i-1]
	cur := runes[i]
	// Rule 1: non-uppercase -> uppercase (covers lowercase letters AND
	// digits; digits attach to the preceding alpha run but the
	// transition to uppercase is still a boundary).
	if !unicode.IsUpper(prev) && unicode.IsUpper(cur) {
		return true
	}
	// Rule 2: consecutive-uppercase trailing boundary.
	if opts.SplitConsecutiveUpper && isUpperRunTrailing(runes, i, prev, cur) {
		return true
	}
	return false
}

// isUpperRunTrailing reports whether runes[i] is the trailing-uppercase
// boundary of a >=2-rune uppercase run followed by a lowercase rune.
// All three preconditions must hold:
//
//   - runes[i] (cur) is uppercase
//   - runes[i+1] exists and is lowercase
//   - runes[i-1] (prev) is uppercase
//
// Extracted from isCamelBoundary so each function stays within the
// gocyclo complexity budget. Operates purely on the supplied rune
// neighbours; no fresh data structures.
func isUpperRunTrailing(runes []rune, i int, prev, cur rune) bool {
	if i+1 >= len(runes) {
		return false
	}
	return unicode.IsUpper(cur) && unicode.IsLower(runes[i+1]) && unicode.IsUpper(prev)
}

// buildTokeniseSepSet returns a stack-allocated [128]bool table where
// index i is true iff the byte i appears in seps. Non-ASCII bytes in
// seps are silently ignored (matched separately via strings.ContainsRune
// in the loop body).
//
// The table is built per call so that opts.SeparatorChars changes are
// observed immediately and no init()-time table exists at the package
// level. Matches the Normalise package's buildSepSet pattern from plan
// 01-06; named distinctly so each primitive's table-builder is locally
// readable.
func buildTokeniseSepSet(seps string) [128]bool {
	var set [128]bool
	for i := 0; i < len(seps); i++ {
		b := seps[i]
		if b < 0x80 {
			set[b] = true
		}
	}
	return set
}

// containsNonASCII reports whether s contains any byte >= 0x80, i.e.
// whether the Unicode-rune separator membership check is needed in
// addition to the ASCII fast-path lookup. Empty s returns false.
func containsNonASCII(s string) bool {
	for i := 0; i < len(s); i++ {
		if s[i] >= 0x80 {
			return true
		}
	}
	return false
}

// isSeparator reports whether r should be treated as a token-boundary
// separator. The ASCII fast-path (sepASCII lookup) is used for runes <
// 0x80; for non-ASCII runes, the slow path falls back to
// strings.ContainsRune over the full SeparatorChars string. The slow
// path is skipped entirely when SeparatorChars is pure-ASCII (the
// hasNonASCIISep precomputed flag), saving an O(len(seps)) scan per
// non-ASCII input rune in the common case.
func isSeparator(r rune, sepASCII [128]bool, seps string, hasNonASCIISep bool) bool {
	if r < 0x80 {
		return sepASCII[r]
	}
	if !hasNonASCIISep {
		return false
	}
	return strings.ContainsRune(seps, r)
}

// appendToken folds the rune slice into a string, optionally lowercases
// it, and appends to dst. Empty rune slices are silently dropped so
// callers don't need to guard against zero-width tokens (consecutive
// separators flow through here as empty slices and disappear).
//
// The lowercase fold is rune-aware: ASCII runes take the bitwise OR
// 0x20 fast path, non-ASCII runes delegate to unicode.ToLower. The
// pattern matches normalise.go's lowerRune helper.
func appendToken(dst []string, rs []rune, lowercase bool) []string {
	if len(rs) == 0 {
		return dst
	}
	if !lowercase {
		return append(dst, string(rs))
	}
	// Inline lowercase: build a byte buffer rune-by-rune. Allocating a
	// fresh buffer per token is the simplest correct path; profiling
	// can revisit if benchmarks show this is a bottleneck.
	buf := make([]byte, 0, len(rs)*utf8.UTFMax)
	for _, r := range rs {
		buf = utf8.AppendRune(buf, lowerRuneToken(r))
	}
	return append(dst, string(buf))
}

// lowerRuneToken folds r to lowercase. ASCII A..Z take the bitwise OR
// fast path; non-ASCII delegate to unicode.ToLower. Runes outside the
// cased range pass through unchanged. Identical in shape to
// normalise.go's lowerRune; duplicated here so tokenise.go has no
// internal dependency on normalise.go (each primitive is independently
// reviewable per .claude/skills/go-coding-standards).
func lowerRuneToken(r rune) rune {
	if r < 0x80 {
		if r >= 'A' && r <= 'Z' {
			return r | 0x20
		}
		return r
	}
	return unicode.ToLower(r)
}
