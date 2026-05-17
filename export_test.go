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

// export_test.go uses the build-tag-free `_test.go` suffix to re-export
// selected unexported symbols to the external (black-box) test package
// without polluting the public API surface. This is the canonical Go
// pattern for testing package internals from package fuzzymatch_test.
//
// Anything added here is visible only to tests; consumers never see it.

package fuzzymatch

// CanonicalMarshalForTest exposes the unexported canonicalMarshal helper
// to the external (black-box) fuzzymatch_test package so that
// golden_canonical_test.go can assert the locked v1.x byte contract
// without dragging canonicalMarshal into the public API.
//
// Do not use this in production code — it does not exist outside of
// _test.go compilation. Use WriteGoldenFile (the test-only re-export of
// the unexported writeGoldenFile helper) for the -update workflow.
var CanonicalMarshalForTest = canonicalMarshal

// WriteGoldenFile is the test-only re-export of the unexported
// writeGoldenFile helper from golden_canonical.go. The Q14b mechanical
// refactor (Phase 8.5 Plan 15a) unexported the production-side helper —
// no consumer should ever invoke it — while keeping the symbol visible
// to _test.go callers (algorithms_golden_test.go, golden_canonical_test.go,
// golden_test.go) under the original capitalised name. The build-tag
// _test.go file suffix excludes this re-export from the public artefact.
var WriteGoldenFile = writeGoldenFile

// NumAlgorithmsForTest re-exports the unexported numAlgorithms constant
// to the external test package. Test code asserts the dispatch array
// is sized for exactly 23 entries; consumers never see this symbol.
const NumAlgorithmsForTest = numAlgorithms

// DispatchLenForTest returns the length of the unexported dispatch
// array. Test code uses this to assert (a) the array is sized for
// numAlgorithms entries, and (b) every entry is nil at the Phase 1
// state (algorithms register themselves from Phase 2 onwards). The
// function rather than a direct re-export is used to avoid copying
// the array (which contains function pointers).
func DispatchLenForTest() int { return len(dispatch) }

// DispatchEntryNilForTest reports whether the dispatch entry at the
// given index is currently nil. Phase 1 expects every entry to be
// nil; future phases populate entries as they implement algorithms.
//
// Out-of-range indices return false (the entry doesn't exist).
func DispatchEntryNilForTest(i int) bool {
	if i < 0 || i >= len(dispatch) {
		return false
	}
	return dispatch[i] == nil
}

// DispatchInvokeForTest invokes the dispatch entry at index i with the
// given (a, b string) arguments and returns the resulting score. This is
// used by per-algorithm dispatch tests to exercise the dispatched closure
// body — Phase 5 dispatch wrappers (q-gram tier) are closures that bind a
// default n parameter, and the closure body must be exercised at least
// once to satisfy the per-file 90% coverage floor.
//
// Out-of-range indices and nil dispatch entries return 0.0 — callers
// should pre-check via DispatchEntryNilForTest if they need to disambiguate.
func DispatchInvokeForTest(i int, a, b string) float64 {
	if i < 0 || i >= len(dispatch) || dispatch[i] == nil {
		return 0.0
	}
	return dispatch[i](a, b)
}

// WinklerPrefixScaleForTest re-exports the unexported winklerPrefixScale
// constant to the external test package. Test code asserts the constant is
// exactly 0.1 (Winkler 1990 p. 357) against accidental drift.
const WinklerPrefixScaleForTest = winklerPrefixScale

// WinklerMaxPrefixForTest re-exports the unexported winklerMaxPrefix constant
// to the external test package. Test code asserts the constant is exactly 4
// (Winkler 1990 p. 357 — the L_max cap) against accidental drift.
const WinklerMaxPrefixForTest = winklerMaxPrefix

// WinklerBoostThresholdForTest re-exports the unexported winklerBoostThreshold
// constant to the external test package. Test code asserts the constant is
// exactly 0.7 (Winkler 1990 p. 357 — the boost gate) against accidental drift.
const WinklerBoostThresholdForTest = winklerBoostThreshold

// Strcmp95SimilarCharsLenForTest re-exports the count of entries in the
// strcmp95SimilarChars table to the external test package. Used by
// TestStrcmp95_TableInvariants to assert the canonical Winkler 1994 TR-2
// 36-pair count against transcription drift (RESEARCH.md Pitfall 1).
const Strcmp95SimilarCharsLenForTest = len(strcmp95SimilarChars)

// Strcmp95SimilarCharsEntryForTest returns the (a, b, sim) entry at index i
// in the strcmp95SimilarChars table. Used by TestStrcmp95_TableInvariants to
// walk the table and assert every entry has the canonical 0.3 weight AND
// that no duplicate pair appears (Pitfall 1 transcription-typo gate).
//
// Returns (0, 0, 0) for out-of-range i.
func Strcmp95SimilarCharsEntryForTest(i int) (a, b byte, sim float64) {
	if i < 0 || i >= len(strcmp95SimilarChars) {
		return 0, 0, 0
	}
	t := strcmp95SimilarChars[i]
	return t.a, t.b, t.sim
}

// Strcmp95SimilarCreditForTest re-exports the canonical Winkler 1994 0.3
// similar-character weight constant to the external test package. Test code
// asserts every table entry's sim value equals this constant.
const Strcmp95SimilarCreditForTest = strcmp95SimilarCredit

// ExtractQGramsForTest re-exports the unexported extractQGrams helper from
// q_gram.go to the external (black-box) test package so q_gram_test.go can
// assert the multiset semantics, capacity hints, and degenerate-input
// behaviour of the byte-path q-gram extractor without dragging the helper
// into the public API.
//
// Plan 05-01 introduces this re-export; plans 05-02 (Sørensen-Dice),
// 05-03 (Cosine), and 05-04 (Tversky) consume the same helper internally
// and can extend the test re-export pattern as needed.
var ExtractQGramsForTest = extractQGrams

// ExtractQGramsRunesForTest re-exports the unexported extractQGramsRunes
// helper for the rune-path q-gram extractor. Pairs with
// ExtractQGramsForTest above.
var ExtractQGramsRunesForTest = extractQGramsRunes

// LCSLenForTest re-exports the unexported lcsLen helper from
// token_indel.go to the external (black-box) test package so
// token_indel_test.go can assert the Wagner-Fischer 1974 LCS-SUBSEQUENCE
// recurrence and pin the PITFALL 6 divergence from `lcsstr.go`'s
// LCS-substring kernel without dragging the helper into the public API.
//
// Plan 06-01 introduces this re-export; plans 06-02 (TokenSetRatio),
// 06-03 (PartialRatio), and 06-05 (MongeElkan, indirectly via the
// PartialRatio inner-metric integration) consume the same kernel
// internally and may extend the test re-export pattern as needed.
var LCSLenForTest = lcsLen

// IndelRatioForTest re-exports the unexported indelRatio helper from
// token_indel.go. The Indel-formula similarity (2·LCS / (|a|+|b|)) is
// the RapidFuzz-canonical normalisation consumed by TokenSortRatio
// (plan 06-01), TokenSetRatio (plan 06-02), and PartialRatio (plan
// 06-03). Pairs with LCSLenForTest above.
var IndelRatioForTest = indelRatio

// LCSLenRunesForTest re-exports the unexported lcsLenRunes helper.
// Used by token_indel_test.go to assert rune-aware LCS-subsequence
// length on multi-byte UTF-8 inputs (e.g. "café"/"cafe" → 3).
//
// The kernel is retained even though Phase 8.5 Q5 (plan 08.5-03)
// removed the former rune-path PartialRatio consumer of
// indelRatioRunes, because token_indel_test.go pins the kernel's
// rune-aware behaviour directly and any future rune-aware consumer
// re-introduced post-v1.0 may compose against it.
var LCSLenRunesForTest = lcsLenRunes

// IndelRatioRunesForTest re-exports the unexported indelRatioRunes
// helper. Pairs with LCSLenRunesForTest above.
var IndelRatioRunesForTest = indelRatioRunes

// MRAThresholdForTest re-exports the unexported mraThreshold function to the
// external test package. Used by PropMRA_ThresholdMonotonic to assert the
// monotonic-decrease property of mraThresholdTable for sumLen ∈ [0, 20], and
// by TestMRAThresholdTable_Clamp to assert the sum>12 clamp behaviour
// (RESEARCH.md Pitfall 7.C — the clamp is often omitted from Wikipedia summaries).
var MRAThresholdForTest = mraThreshold

// TokeniseRuneForTest re-exports the unexported tokeniseRune function — the
// pre-Phase-8.5 rune-based Tokenise implementation retained as an internal
// entry point. Used by TestProp_Tokenise_ASCIIFastPathEquivalent to assert
// that the byte-level ASCII fast path (Phase 8.5 Q8b) produces output
// byte-identical to the rune path on every ASCII input. Load-bearing
// correctness gate: any divergence between paths is by definition a
// fast-path bug.
var TokeniseRuneForTest = tokeniseRune
