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

// AlgoID is the typed identifier for the 23 algorithms in the
// fuzzymatch catalogue. Dispatch is array-backed for zero-allocation
// hot-path access — see docs/requirements.md §6.
//
// The constant order matches the spec catalogue order in §7 and is the
// v1.x stability contract. Adding new algorithms appends to the end and
// NEVER reorders existing constants — consumers that pin AlgoID values
// (e.g. by persisting them) must remain stable across minor releases.
//
// No init() function appears in this file (per design principle 5(12)
// and determinism-standards): the String() method is implemented via a
// switch statement, and AlgoIDs() returns a freshly-allocated slice
// literal on each call. No []string indexed by AlgoID is built at
// package load time.

package fuzzymatch

import "fmt"

// AlgoID is the typed enum that identifies each of the 23 algorithms in
// the fuzzymatch catalogue. It is a plain int — not int32, not int64,
// not a struct — so that the package-internal dispatch table can index
// into it directly with zero allocation overhead.
//
// AlgoID values are stable across patch releases. The integer values
// themselves are PART of the v1.x contract: consumers may persist them,
// compare them, and rely on AlgoLevenshtein evaluating to 0 (its iota
// position). Future additions append to the END of the const block —
// existing AlgoID values never shift.
//
// Use AlgoIDs() to enumerate the valid set; use String() to obtain the
// canonical-spelling text label for an AlgoID.
type AlgoID int

// The 23 algorithm identifiers, in the spec catalogue order from
// docs/requirements.md §7. Each constant cites the originating
// algorithm briefly; the full primary-source citation lives at the top
// of the corresponding implementation file (added from Phase 2 onwards).
const (
	// AlgoLevenshtein identifies the Levenshtein edit-distance
	// similarity (Levenshtein 1965 — "Binary codes capable of correcting
	// deletions, insertions, and reversals").
	AlgoLevenshtein AlgoID = iota

	// AlgoDamerauLevenshteinOSA identifies the Optimal String Alignment
	// variant of Damerau-Levenshtein, where adjacent transpositions are
	// counted as one edit but each substring is edited at most once
	// (Boytsov 2011 — "Indexing methods for approximate dictionary
	// searching: comparative analysis").
	AlgoDamerauLevenshteinOSA

	// AlgoDamerauLevenshteinFull identifies the unrestricted
	// Damerau-Levenshtein distance, where any pair of adjacent
	// transpositions counts as a single edit (Damerau 1964 — "A
	// technique for computer detection and correction of spelling
	// errors").
	AlgoDamerauLevenshteinFull

	// AlgoHamming identifies the Hamming distance — number of differing
	// positions between two equal-length strings (Hamming 1950 — "Error
	// detecting and error correcting codes"). Defined only for equal-
	// length inputs.
	AlgoHamming

	// AlgoJaro identifies the Jaro similarity (Jaro 1989 — "Advances in
	// record-linkage methodology as applied to matching the 1985 census
	// of Tampa, Florida").
	AlgoJaro

	// AlgoJaroWinkler identifies the Jaro-Winkler similarity, which
	// boosts the Jaro score for strings sharing a common prefix
	// (Winkler 1990 — "String comparator metrics and enhanced decision
	// rules for the Fellegi-Sunter model of record linkage").
	AlgoJaroWinkler

	// AlgoStrcmp95 identifies Winkler's Strcmp95 enhancement of
	// Jaro-Winkler, which adds similar-character credit for common
	// typewriter substitutions (Winkler & Thibaudeau 1991 — "An
	// application of the Fellegi-Sunter model of record linkage to the
	// 1990 U.S. decennial census").
	AlgoStrcmp95

	// AlgoSmithWatermanGotoh identifies the Smith-Waterman algorithm
	// with affine gap penalties per Gotoh's improvement (Smith & Waterman
	// 1981; Gotoh 1982 — "An improved algorithm for matching biological
	// sequences"). Subject to the documented Gotoh erratum — see Phase 3
	// for cross-validation discipline.
	AlgoSmithWatermanGotoh

	// AlgoLCSStr identifies the Longest Common Substring similarity
	// (Hunt & Szymanski 1977 — "A fast algorithm for computing longest
	// common subsequences" — substring variant).
	AlgoLCSStr

	// AlgoQGramJaccard identifies the Jaccard similarity over q-gram
	// sets (Ukkonen 1992 — "Approximate string-matching with q-grams
	// and maximal matches").
	AlgoQGramJaccard

	// AlgoSorensenDice identifies the Sørensen-Dice coefficient over
	// q-gram sets (Sørensen 1948; Dice 1945 — independent rediscoveries
	// of the same coefficient).
	AlgoSorensenDice

	// AlgoCosine identifies the cosine similarity over q-gram frequency
	// vectors (Salton & McGill 1983 — "Introduction to Modern
	// Information Retrieval").
	AlgoCosine

	// AlgoTversky identifies the Tversky index over q-gram sets — a
	// generalisation of Jaccard and Dice with configurable asymmetric
	// weights (Tversky 1977 — "Features of similarity").
	AlgoTversky

	// AlgoMongeElkan identifies the Monge-Elkan token-level similarity,
	// which composes any character-level inner metric across token pairs
	// (Monge & Elkan 1996 — "The field matching problem: algorithms
	// and applications").
	AlgoMongeElkan

	// AlgoTokenSortRatio identifies the token-sort ratio: tokenise both
	// sides, sort the token sets, then apply the character-level inner
	// metric to the joined strings (Cohen, Ravikumar & Fienberg 2003 —
	// SecondString library reference).
	AlgoTokenSortRatio

	// AlgoTokenSetRatio identifies the token-set ratio: tokenise both
	// sides, compute intersection / difference / difference subsets,
	// then apply the inner metric to the best-matching reconstruction
	// (Cohen et al. 2003 — SecondString library reference).
	AlgoTokenSetRatio

	// AlgoPartialRatio identifies the partial ratio: the inner metric
	// applied to the best-aligned substring of the longer input
	// (RapidFuzz / FuzzyWuzzy heritage — Cohen et al. 2003 lineage).
	AlgoPartialRatio

	// AlgoTokenJaccard identifies the Jaccard similarity over token sets
	// (Cohen et al. 2003 — SecondString library reference).
	AlgoTokenJaccard

	// AlgoSoundex identifies the Soundex phonetic code (Russell 1918 —
	// U.S. Patent 1,261,167; the American Soundex variant per Knuth
	// 1973 — TAOCP Vol. 3, §6).
	AlgoSoundex

	// AlgoDoubleMetaphone identifies Philips's Double Metaphone phonetic
	// code (Philips 2000 — "The Double Metaphone search algorithm").
	// Note: Metaphone 3 is explicitly EXCLUDED from this catalogue per
	// docs/requirements.md §4 (U.S. Patent 7440941).
	AlgoDoubleMetaphone

	// AlgoNYSIIS identifies the New York State Identification and
	// Intelligence System phonetic code (Taft 1970 — NY State Special
	// Report No. 1).
	AlgoNYSIIS

	// AlgoMRA identifies the Match Rating Approach (Moore 1977 — NBS
	// Technical Note 943, "Accessing Individual Records from Personal
	// Data Files Using Non-unique Identifiers").
	AlgoMRA

	// AlgoRatcliffObershelp identifies the Ratcliff-Obershelp pattern-
	// matching similarity (Ratcliff & Metzener 1988 — "Pattern matching:
	// the Gestalt approach", Dr. Dobb's Journal).
	AlgoRatcliffObershelp
)

// numAlgorithms is the count of declared AlgoID constants. It sizes the
// internal dispatch array and is exposed (only) to package-internal
// validation paths — consumers wanting the algorithm count call
// len(AlgoIDs()).
const numAlgorithms = int(AlgoRatcliffObershelp) + 1

// String returns the canonical CamelCase spelling of the algorithm.
// The mapping is stable across patch releases (and is part of the v1.x
// contract).
//
// The returned strings match the constant names without the "Algo"
// prefix, e.g. AlgoLevenshtein → "Levenshtein", AlgoDamerauLevenshteinOSA
// → "DamerauLevenshteinOSA", AlgoLCSStr → "LCSStr", AlgoNYSIIS →
// "NYSIIS".
//
// For an out-of-range AlgoID (e.g. AlgoID(999), or any future value
// declared after this method is compiled), String returns the fallback
// form "AlgoID(N)" via fmt.Sprintf — intentionally allocating because
// the path is for error and debug output, not the hot dispatch path.
//
// String never allocates on the in-range path: every case returns a
// compile-time string constant.
//
// The high cyclomatic complexity of the switch is intentional: every
// AlgoID gets an explicit case so a new constant cannot silently fall
// through to the fallback "AlgoID(N)" branch — adding an algorithm
// forces the author to add a String() case in the same PR, which the
// algorithm-correctness reviewer flags. This is the canonical Go
// idiom for a stringly-typed enum and gocyclo's threshold does not
// apply to the pattern.
func (id AlgoID) String() string { //nolint:gocyclo // one switch case per AlgoID is intentional — see godoc above
	switch id {
	case AlgoLevenshtein:
		return "Levenshtein"
	case AlgoDamerauLevenshteinOSA:
		return "DamerauLevenshteinOSA"
	case AlgoDamerauLevenshteinFull:
		return "DamerauLevenshteinFull"
	case AlgoHamming:
		return "Hamming"
	case AlgoJaro:
		return "Jaro"
	case AlgoJaroWinkler:
		return "JaroWinkler"
	case AlgoStrcmp95:
		return "Strcmp95"
	case AlgoSmithWatermanGotoh:
		return "SmithWatermanGotoh"
	case AlgoLCSStr:
		return "LCSStr"
	case AlgoQGramJaccard:
		return "QGramJaccard"
	case AlgoSorensenDice:
		return "SorensenDice"
	case AlgoCosine:
		return "Cosine"
	case AlgoTversky:
		return "Tversky"
	case AlgoMongeElkan:
		return "MongeElkan"
	case AlgoTokenSortRatio:
		return "TokenSortRatio"
	case AlgoTokenSetRatio:
		return "TokenSetRatio"
	case AlgoPartialRatio:
		return "PartialRatio"
	case AlgoTokenJaccard:
		return "TokenJaccard"
	case AlgoSoundex:
		return "Soundex"
	case AlgoDoubleMetaphone:
		return "DoubleMetaphone"
	case AlgoNYSIIS:
		return "NYSIIS"
	case AlgoMRA:
		return "MRA"
	case AlgoRatcliffObershelp:
		return "RatcliffObershelp"
	default:
		// Fallback for out-of-range values. Intentionally allocates —
		// this branch is reached only by malformed input or future-
		// version drift, never by normal dispatch.
		return fmt.Sprintf("AlgoID(%d)", int(id))
	}
}

// AlgoIDs returns the full set of 23 algorithm identifiers in their
// declared order (the v1.x stable catalogue order from
// docs/requirements.md §7).
//
// The returned slice is freshly allocated on every call so the caller
// may freely mutate, sort, or filter it without affecting other
// callers. The order is deterministic and identical across runs,
// processes, and platforms — there is no map iteration on this path
// (per the no-map-iteration rule in
// .claude/skills/determinism-standards).
//
// Consumers needing the algorithm count call len(AlgoIDs()); consumers
// needing the text label call (AlgoID).String().
func AlgoIDs() []AlgoID {
	return []AlgoID{
		AlgoLevenshtein,
		AlgoDamerauLevenshteinOSA,
		AlgoDamerauLevenshteinFull,
		AlgoHamming,
		AlgoJaro,
		AlgoJaroWinkler,
		AlgoStrcmp95,
		AlgoSmithWatermanGotoh,
		AlgoLCSStr,
		AlgoQGramJaccard,
		AlgoSorensenDice,
		AlgoCosine,
		AlgoTversky,
		AlgoMongeElkan,
		AlgoTokenSortRatio,
		AlgoTokenSetRatio,
		AlgoPartialRatio,
		AlgoTokenJaccard,
		AlgoSoundex,
		AlgoDoubleMetaphone,
		AlgoNYSIIS,
		AlgoMRA,
		AlgoRatcliffObershelp,
	}
}

// dispatch maps each AlgoID to its score function. Entries are nil
// until Phase 2+ plans populate them (each algorithm's implementation
// file registers itself at package load time by direct assignment —
// dispatch[AlgoLevenshtein] = levenshteinScore, etc.).
//
// Consumers MUST NOT access dispatch directly — it is unexported by
// design. The Scorer (Phase 8) and Extract (Phase 10) reach into it via
// package-internal lookup helpers that gate on `int(id) >= numAlgorithms
// || dispatch[id] == nil` and return the appropriate sentinel error on
// out-of-range or unregistered ids.
//
// The array sizing comes from numAlgorithms, which is a compile-time
// constant — adding an AlgoID resizes the array automatically without
// touching this declaration.
var dispatch [numAlgorithms]func(a, b string) float64
