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

// Validate is the consumer-facing diagnostic surface that reports
// problematic-but-non-fatal input shapes BEFORE scoring runs. See
// docs/requirements.md §11.5 for the authoritative specification and
// .planning/REQUIREMENTS.md VALIDATE-01..VALIDATE-05 for the
// requirement-tracker entries.
//
// Implementation discipline:
//
//   - Pure function — no goroutines, no channels, no mutexes, no I/O.
//   - Never panics, never returns an error. Lenient by contract per
//     docs/requirements.md §6.A.
//   - Returns nil (not an empty slice) when no warnings apply, per
//     VALIDATE-01.
//   - Deterministic output ordering: warnings are sorted by
//     (Algorithm, Kind) via sort.SliceStable with a complete sort key.
//     No map iteration on the output path
//     (per .claude/skills/determinism-standards).
//   - Safe for concurrent use — no shared state.
//   - Allocation-light: one slice header, plus one entry per emitted
//     warning. Pre-allocated with a capacity hint of 8 to cover the
//     common case (empty input → 1 warning; non-ASCII input to
//     ASCII-only algos → up to 5 warnings).
//
// Per-WarnKind dispatch (matches docs/requirements.md §11.5):
//
//   - WarnEmptyInput → one cross-cutting Warning with AlgoIDAny scope.
//   - WarnUnequalLength → one Warning scoped to AlgoHamming.
//   - WarnNoTokensAfterNormalise → one Warning per affected token-tier
//     algorithm (AlgoTokenSortRatio, AlgoTokenSetRatio, AlgoPartialRatio,
//     AlgoTokenJaccard, AlgoMongeElkan).
//   - WarnAllNonASCIIDropped → one Warning per affected ASCII-only
//     algorithm (AlgoStrcmp95, AlgoSoundex, AlgoDoubleMetaphone,
//     AlgoNYSIIS, AlgoMRA).
//   - WarnPathologicallyLargeInput → one cross-cutting Warning with
//     AlgoIDAny scope (consumers gate at the call site per §11.5
//     guidance).

package fuzzymatch

import (
	"fmt"
	"sort"
	"unicode/utf8"
)

// validatePathologicalThreshold is the input-size cut-off (in bytes,
// applied to max(len(a), len(b))) above which Validate emits a
// WarnPathologicallyLargeInput. The threshold is intentionally
// generous — 64 KiB per side — because the warning is consumer-DoS
// guidance, not a hard limit. Algorithms still produce a value above
// the threshold; the warning tells the consumer that the O(m·n)
// algorithms (DamerauLevenshteinFull, SmithWatermanGotoh,
// RatcliffObershelp, MongeElkan, PartialRatio) will allocate a DP
// table proportional to len(a) × len(b).
//
// The 65_536-byte ceiling corresponds to a worst-case ~4 GiB DP table
// for the O(m·n) algorithms — well past any reasonable workload.
// Tuning is deferred to docs/algorithms.md per-algorithm guidance.
const validatePathologicalThreshold = 65_536

// Warning carries a single diagnostic emitted by Validate. The struct
// is plain-data: no methods, no pointers, no mutability concerns. A
// Warning may be freely copied, compared (==), and serialised.
//
// Field semantics:
//
//   - Algorithm: the AlgoID the warning applies to, or AlgoIDAny if
//     the warning is cross-cutting (affects every algorithm — e.g.
//     WarnEmptyInput, WarnPathologicallyLargeInput). Consumers
//     distinguish the two cases by checking
//     `w.Algorithm == fuzzymatch.AlgoIDAny`.
//   - Kind: the WarnKind discriminator (see warn_kind.go). The zero
//     value is "unspecified" — Validate never emits a Warning with
//     Kind == 0.
//   - Detail: a stable English (British spelling) human-readable
//     elaboration. Suitable for log lines, telemetry, or audit
//     trails. Stable across patch releases; consumers MUST NOT parse
//     this field — use Kind and Algorithm for programmatic checks.
//
// See docs/requirements.md §11.5 and VALIDATE-02.
type Warning struct {
	Algorithm AlgoID
	Kind      WarnKind
	Detail    string
}

// Validate inspects the two inputs and returns warnings describing
// problematic-but-non-fatal input shapes. Returns nil if no warnings
// apply (per VALIDATE-01 — NOT an empty slice; consumers checking
// `len(warnings) > 0` see the same answer either way, but the nil-vs-
// empty-slice distinction is part of the v1.x contract).
//
// Validate is pure: no goroutines, no I/O, no shared state. Safe for
// concurrent use from any number of goroutines. Never panics, never
// returns an error.
//
// Output is deterministically sorted by (Algorithm, Kind) — two
// successive calls with identical (a, b) return slices that are
// byte-identical (per the determinism contract;
// TestValidate_DeterministicOrdering is the regression gate).
//
// Per-WarnKind rules (matches docs/requirements.md §11.5):
//
//   - WarnEmptyInput: emitted if a == "" or b == "" (or both). Scoped
//     to AlgoIDAny — affects every algorithm. Identity short-circuits,
//     Hamming returns trivial values, phonetic algorithms produce
//     empty keys that match each other.
//   - WarnUnequalLength: emitted if len(a) != len(b). Scoped to
//     AlgoHamming — documents the silent-max policy (the score
//     collapses to 0.0 without any error).
//   - WarnNoTokensAfterNormalise: emitted if either input produces
//     an empty token list under DefaultTokeniseOptions. Scoped per
//     affected token-tier algorithm (TokenSortRatio, TokenSetRatio,
//     PartialRatio, TokenJaccard, MongeElkan).
//   - WarnAllNonASCIIDropped: emitted if either input contains
//     characters but is entirely non-ASCII. Scoped per affected
//     ASCII-only algorithm (Strcmp95, Soundex, DoubleMetaphone,
//     NYSIIS, MRA).
//   - WarnPathologicallyLargeInput: emitted if max(len(a), len(b))
//     exceeds validatePathologicalThreshold (64 KiB). Scoped to
//     AlgoIDAny — consumers gate at the call site rather than
//     relying on this warning post-hoc.
//
// The validate-then-score pattern is the recommended idiom for code
// paths that audit input quality:
//
//	warnings := fuzzymatch.Validate(a, b)
//	if len(warnings) > 0 {
//	    for _, w := range warnings {
//	        log.Printf("input quality: %s (%s): %s",
//	            w.Kind, w.Algorithm, w.Detail)
//	    }
//	}
//	score := fuzzymatch.DefaultScorer().Score(a, b)
//
// See docs/requirements.md §11.5 and VALIDATE-01 + VALIDATE-02 +
// VALIDATE-05.
func Validate(a, b string) []Warning { //nolint:gocyclo // 5 cross-cutting WarnKind classifications dispatched in series; each branch is a distinct documented warning category with its own fanout policy
	// Pre-allocate with capacity 8: covers the common cases without
	// growing the backing array. Most calls emit zero or one warning;
	// a non-ASCII-only input emits up to 5 (one per ASCII-only algo);
	// a token-empty input emits up to 5 (one per token-tier algo).
	warnings := make([]Warning, 0, 8)

	// --- Cross-cutting checks ---

	// WarnEmptyInput: any empty input affects every algorithm.
	if a == "" || b == "" {
		warnings = append(warnings, Warning{
			Algorithm: AlgoIDAny,
			Kind:      WarnEmptyInput,
			Detail:    validateEmptyInputDetail(a, b),
		})
	}

	// WarnPathologicallyLargeInput: input over the documented threshold
	// risks O(m·n) DP-table allocation on the quadratic algorithms.
	// Cross-cutting because the threshold applies to any algorithm with
	// quadratic complexity; consumers gate at the call site.
	if len(a) > validatePathologicalThreshold || len(b) > validatePathologicalThreshold {
		warnings = append(warnings, Warning{
			Algorithm: AlgoIDAny,
			Kind:      WarnPathologicallyLargeInput,
			Detail: fmt.Sprintf(
				"input length max(len(a)=%d, len(b)=%d) exceeds threshold %d — O(m·n) algorithms (DamerauLevenshteinFull, SmithWatermanGotoh, RatcliffObershelp, MongeElkan, PartialRatio) will allocate proportionally",
				len(a), len(b), validatePathologicalThreshold,
			),
		})
	}

	// --- Per-algorithm checks ---

	// WarnUnequalLength: Hamming silent-max policy.
	if len(a) != len(b) {
		warnings = append(warnings, Warning{
			Algorithm: AlgoHamming,
			Kind:      WarnUnequalLength,
			Detail: fmt.Sprintf(
				"len(a)=%d, len(b)=%d — Hamming requires equal-length inputs and silently scores 0.0 on unequal-length input",
				len(a), len(b),
			),
		})
	}

	// WarnNoTokensAfterNormalise: token-tier algorithms compare empty
	// token sets when either input tokenises to nothing. Single
	// Tokenise() call per side covers all five token-tier algorithms.
	aTokens := Tokenise(a, DefaultTokeniseOptions())
	bTokens := Tokenise(b, DefaultTokeniseOptions())
	if (len(aTokens) == 0 || len(bTokens) == 0) && (a != "" || b != "") {
		// Suppress for the all-empty case — WarnEmptyInput already
		// covers it. Emit one Warning per affected token-tier algorithm
		// so consumers filtering by Algorithm see the issue against
		// their algorithm of interest.
		detail := fmt.Sprintf(
			"Tokenise produces empty token list (len(aTokens)=%d, len(bTokens)=%d) — token-tier algorithms compare empty sets",
			len(aTokens), len(bTokens),
		)
		for _, algo := range tokenTierAlgosForValidate {
			warnings = append(warnings, Warning{
				Algorithm: algo,
				Kind:      WarnNoTokensAfterNormalise,
				Detail:    detail,
			})
		}
	}

	// WarnAllNonASCIIDropped: ASCII-only algorithms (Strcmp95, Soundex,
	// DoubleMetaphone, NYSIIS, MRA) collapse non-ASCII-only input to
	// empty after ASCII filtering. Trigger when at least one side
	// contains characters but every rune is non-ASCII.
	if hasOnlyNonASCII(a) || hasOnlyNonASCII(b) {
		detail := fmt.Sprintf(
			"input contains only non-ASCII runes (a all-non-ASCII=%t, b all-non-ASCII=%t) — ASCII-only algorithms drop every rune and compare empty codes",
			hasOnlyNonASCII(a), hasOnlyNonASCII(b),
		)
		for _, algo := range asciiOnlyAlgosForValidate {
			warnings = append(warnings, Warning{
				Algorithm: algo,
				Kind:      WarnAllNonASCIIDropped,
				Detail:    detail,
			})
		}
	}

	// Per VALIDATE-01: return nil (not an empty slice) when nothing
	// applied.
	if len(warnings) == 0 {
		return nil
	}

	// Deterministic ordering: sort by (Algorithm, Kind). Use
	// sort.SliceStable (not sort.Slice) so two warnings with the same
	// (Algorithm, Kind) preserve insertion order — only the per-
	// algorithm-loop iteration order remains, which itself is
	// deterministic because the algo lists are sorted slices declared
	// below.
	sort.SliceStable(warnings, func(i, j int) bool {
		if warnings[i].Algorithm != warnings[j].Algorithm {
			return warnings[i].Algorithm < warnings[j].Algorithm
		}
		return warnings[i].Kind < warnings[j].Kind
	})

	return warnings
}

// tokenTierAlgosForValidate is the deterministic list of AlgoIDs that
// consume Tokenise output. Declared as a sorted slice (by AlgoID
// integer value) so the per-algorithm-loop in Validate iterates in a
// stable order even before the final sort.SliceStable runs — making
// the dispatch traceable in debugging.
var tokenTierAlgosForValidate = []AlgoID{
	AlgoMongeElkan,     // 13
	AlgoTokenSortRatio, // 14
	AlgoTokenSetRatio,  // 15
	AlgoPartialRatio,   // 16
	AlgoTokenJaccard,   // 17
}

// asciiOnlyAlgosForValidate is the deterministic list of AlgoIDs that
// drop non-ASCII runes before encoding. Strcmp95 is an ASCII-only
// extension of Jaro-Winkler with a similar-character table; Soundex /
// DoubleMetaphone / NYSIIS / MRA are the phonetic-tier algorithms that
// operate on ASCII letters only (per CONTEXT.md §5 and the per-file
// godoc on each).
var asciiOnlyAlgosForValidate = []AlgoID{
	AlgoStrcmp95,        // 6
	AlgoSoundex,         // 18
	AlgoDoubleMetaphone, // 19
	AlgoNYSIIS,          // 20
	AlgoMRA,             // 21
}

// validateEmptyInputDetail builds a stable Detail string describing
// which side(s) are empty.
func validateEmptyInputDetail(a, b string) string {
	switch {
	case a == "" && b == "":
		return "both inputs are empty — every algorithm short-circuits to its empty-vs-empty contract (typically 1.0 by identity)"
	case a == "":
		return "input a is empty — every algorithm short-circuits to the empty-side contract (typically 0.0 unless both sides are empty)"
	default:
		// b == ""
		return "input b is empty — every algorithm short-circuits to the empty-side contract (typically 0.0 unless both sides are empty)"
	}
}

// hasOnlyNonASCII reports whether s contains at least one rune and
// every rune is non-ASCII (i.e. > 0x7F). Empty strings return false
// (the WarnEmptyInput rule covers them).
//
// Implementation note: walks the string with utf8.DecodeRuneInString
// rather than range-over-string to keep the loop allocation-free; the
// short-circuit on first ASCII rune keeps the common case (pure-ASCII
// input) O(1) cost.
func hasOnlyNonASCII(s string) bool {
	if s == "" {
		return false
	}
	for i := 0; i < len(s); {
		r, size := utf8.DecodeRuneInString(s[i:])
		// utf8.RuneError + size 1 indicates invalid UTF-8. Treat as
		// non-ASCII (it isn't ASCII anyway) and continue.
		if r < 0x80 && r != utf8.RuneError {
			return false
		}
		// Edge case: ASCII NUL or a stray byte that decoded as
		// RuneError with size 1 — both unlikely in practice. Treat
		// non-ASCII-or-error as "not ASCII" and keep scanning.
		i += size
	}
	return true
}
