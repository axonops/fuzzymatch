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

// token_sort_ratio.go implements the Token Sort Ratio similarity for the
// fuzzymatch catalogue. Token Sort Ratio is the simplest Indel-formula
// consumer of the shared LCS-subsequence kernel in token_indel.go
// (Wagner-Fischer 1974 J. ACM 21(1):168-173).
//
// Sources:
//
//   - Engineering provenance: RapidFuzz documentation,
//     https://rapidfuzz.github.io/RapidFuzz/Usage/fuzz.html#token-sort-ratio
//     — the canonical modern reference for the algorithm (per
//     algorithm-correctness-standards "For algorithms without an academic
//     primary source ... cite the canonical modern reference"). RapidFuzz
//     in turn descends from SeatGeek's fuzzywuzzy (2014 — superseded by
//     RapidFuzz which fixed several scoring inconsistencies).
//   - Underlying DP source: Wagner, R. A., & Fischer, M. J. (1974). "The
//     string-to-string correction problem." Journal of the ACM 21(1):
//     168-173 — the LCS-subsequence dynamic-programming recurrence used
//     by the indelRatio kernel.
//   - Indel-formula equivalence: see 06-RESEARCH.md Pattern 3 for the
//     proof that 2·LCS / (|a|+|b|) equals 1 - IndelDistance / (|a|+|b|)
//     where IndelDistance is the Levenshtein distance restricted to
//     insertions and deletions (the RapidFuzz "Indel" similarity).
//
// Algorithm — TokenSortRatioScore(a, b):
//
//   1. Identity short-circuit: if a == b, return 1.0 immediately. This
//      avoids the Tokenise allocation on identical inputs (matching
//      qgram_jaccard.go's IN-04 closure pattern).
//   2. Tokenise both sides using DefaultTokeniseOptions() (Lowercase: true,
//      SplitCamelCase: true, SplitConsecutiveUpper: true, SeparatorChars:
//      "_-.:/ \t\n\r"). See OQ-1 RESOLUTION below for the divergence note.
//   3. Both-Tokenised-empty guard: return 1.0 (vacuous match).
//   4. One-Tokenised-empty guard: return 0.0.
//   5. Sort each token slice in-place via sort.Strings (byte-lex ascending).
//   6. Join each side with a single ASCII space.
//   7. Return indelRatio over the byte slices of the two joined strings.
//
// Conventions (mirror the Phase 2/3/4/5 short-circuit pattern):
//
//   - both-empty       → 1.0 (covered by a == b identity short-circuit)
//   - identical        → 1.0 (a == b short-circuit)
//   - one-empty        → 0.0
//   - token-reordered  → 1.0 (sorted-joined strings are identical)
//
// OQ-1 RESOLUTION (tokeniser-divergence handling — LOCKED in plan
// 06-01):
//
// RapidFuzz tokenises via Python `str.split()` — whitespace-only,
// case-preserving. fuzzymatch's `Tokenise(s, DefaultTokeniseOptions())`
// is camelCase / snake_case / kebab-case / dot-case aware AND
// lowercasing. For inputs without identifier-style boundaries (pure
// whitespace-separated lowercase ASCII text), the two tokenisations
// agree and the scores match. For inputs with mixed identifier styles
// — e.g. "userID" vs "user_id" — the project tokenisation produces
// semantically richer splits ([user, id] vs [user, i, d] under
// RapidFuzz's str.split which leaves "userID" as one token).
//
// The cross-validation corpus at
// testdata/cross-validation/token-ratios/vectors.json is restricted to
// whitespace-only lowercase ASCII inputs so cross-validation against
// RapidFuzz 3.14.5 is byte-stable. Consumers wanting RapidFuzz parity
// across identifier-style inputs should build a TokeniseOptions struct
// with SplitCamelCase: false, Lowercase: false, SeparatorChars: " \t\n\r"
// and tokenise themselves before calling lower-level kernels — but this
// is OUT OF SCOPE for v1.0.
//
// Source-origin discipline (per algorithm-licensing-standards):
//
//   - Primary source:        RapidFuzz docs (engineering provenance) +
//                            Wagner & Fischer 1974 (underlying DP).
//   - Cross-validation:      RapidFuzz 3.14.5 via the corpus at
//                            testdata/cross-validation/token-ratios/vectors.json
//                            — every TokenSort entry asserts byte-stable
//                            agreement within epsilon = 1e-9.
//   - Tie-break:             none (Tokenise is deterministic;
//                            sort.Strings is stable byte-lex; indelRatio
//                            is symmetric — see PropTokenSortRatio_Symmetric).
//   - GPL/LGPL provenance:   none.
//   - Code copied verbatim:  none — fresh transcription from RapidFuzz's
//                            _split_sequence + token_sort_ratio Python
//                            source structure only; the Indel kernel is
//                            this project's own Wagner-Fischer 1974
//                            implementation in token_indel.go.
//
// Implementation discipline:
//
//   - NO init()-time table builds (per docs/requirements.md §5(12)).
//   - NO map iteration on output paths (DET-03). Tokenise returns a
//     slice; sort.Strings operates in place; strings.Join is
//     order-preserving.
//   - NO transcendental float operations (DET-06): integer arithmetic
//     in the kernel plus one final division in indelRatio with
//     explicit left-to-right parenthesisation.
//   - NO goroutines, channels, or mutexes.
//   - Identity short-circuit `if a == b { return 1.0 }` BEFORE Tokenise
//     to avoid the `make([]string, 0, 4)` allocation on identical
//     inputs (IN-04 closure pattern from Phase 4).
//   - NO public *Runes variant — Tokenise handles UTF-8 internally per
//     06-CONTEXT.md §6 LOCKED.
//
// Public surface (one function — the dispatched byte-path score):
//
//   - TokenSortRatioScore(a, b string) float64
//
// Registered in dispatch table slot AlgoTokenSortRatio (slot 14 — see
// algoid.go lines 135-139). The dispatch table maps AlgoID to
// (a, b string) float64 with no place for parameters; TokenSortRatio
// has no parameters so the wrapper is the function value directly
// (no closure needed — mirrors dispatch_lcsstr.go).
//
// Worst-case complexity: O((|a|+|b|)^2) time + O(min(|a|,|b|)) space
// for the two-row DP. The sort step is O(t log t) for t = total
// token count (typically t << |a|+|b|); strings.Join is O(|a|+|b|).
// Pure-function library — caller controls input size; the algorithm
// has no input-validation rejection on long input.

package fuzzymatch

import (
	"sort"
	"strings"
)

// TokenSortRatioScore returns the Token Sort Ratio similarity between a
// and b in [0.0, 1.0]: tokenise both sides using
// DefaultTokeniseOptions(), sort each token slice byte-lex ascending,
// join with a single space, then apply the Indel formula
// 2·LCS / (|joinedA|+|joinedB|) (Wagner & Fischer 1974
// LCS-subsequence DP over bytes, RapidFuzz-canonical normalisation).
//
// For programmatic input-quality checks before scoring (including
// WarnNoTokensAfterNormalise scoped to AlgoTokenSortRatio),
// see [fuzzymatch.Validate].
//
// Conventions (mirror Q-Gram Jaccard / LCSStr):
//
//   - TokenSortRatioScore("",        "")               == 1.0  (both-empty / identity)
//   - TokenSortRatioScore("hello",   "hello")          == 1.0  (identity)
//   - TokenSortRatioScore("hello",   "")               == 0.0  (one-empty)
//   - TokenSortRatioScore("",        "hello")          == 0.0  (one-empty)
//   - TokenSortRatioScore("alpha beta",        "beta alpha")        == 1.0
//   - TokenSortRatioScore("fuzzy wuzzy was a bear", "wuzzy fuzzy was a bear") == 1.0
//
// Symmetric across argument order — TokenSortRatioScore(a, b) ==
// TokenSortRatioScore(b, a) — because Tokenise is deterministic,
// sort.Strings is stable byte-lex, and indelRatio is symmetric in its
// argument order. The byte-equality is exact (no float tolerance
// needed); see TestProp_TokenSortRatioScore_Symmetric for the
// quick.Check property.
//
// Tokeniser divergence from RapidFuzz (OQ-1 RESOLUTION LOCKED — see
// file-header godoc): fuzzymatch's Tokenise is identifier-aware
// (camelCase / snake_case / etc.), unlike RapidFuzz's
// whitespace-only Python str.split. For whitespace-only lowercase
// ASCII inputs the two agree; for identifier-style inputs the project
// tokenisation produces semantically richer splits.
//
// Reference vector (cross-validated against RapidFuzz 3.14.5):
//
//	TokenSortRatioScore("fuzzy wuzzy was a bear", "wuzzy fuzzy was a bear") = 1.0
//
// Worst-case time: O((|a|+|b|)^2) for the LCS DP; sort and join are
// dominated by the DP. Allocation budget: 1 sort-by-side + 1 join-by-side
// + 2 DP rows (zero allocs when min(joined lengths) <= 64 bytes; ASCII
// short-input fast path).
//
// This function operates on bytes (the joined sorted strings are
// compared byte-by-byte by the LCS-subsequence DP). For multi-byte
// UTF-8 token contents, Tokenise still produces well-formed UTF-8
// tokens; the byte-level Indel kernel operates on those bytes
// directly. There is no rune-path variant: Tokenise is UTF-8-aware so
// the rune semantic is already preserved at the tokenisation layer.
func TokenSortRatioScore(a, b string) float64 {
	if a == b {
		return 1.0 // identity short-circuit — avoids Tokenise allocations
	}
	tokensA := Tokenise(a, DefaultTokeniseOptions())
	tokensB := Tokenise(b, DefaultTokeniseOptions())
	// Both-Tokenised-empty: vacuous match. This branch fires for inputs
	// that are pure separators on both sides (e.g. " ___ " vs "...") —
	// the a == b short-circuit above doesn't catch these because the
	// raw strings differ.
	if len(tokensA) == 0 && len(tokensB) == 0 {
		return 1.0
	}
	// One-Tokenised-empty: 0.0. RapidFuzz returns 0 in this case
	// (matching the empty-vs-non-empty convention for Indel ratios).
	if len(tokensA) == 0 || len(tokensB) == 0 {
		return 0.0
	}
	sort.Strings(tokensA)
	sort.Strings(tokensB)
	joinedA := strings.Join(tokensA, " ")
	joinedB := strings.Join(tokensB, " ")
	return indelRatio([]byte(joinedA), []byte(joinedB))
}
