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

// WarnKind enumerates the diagnostic categories Validate emits, and the
// companion AlgoIDAny sentinel scopes a Warning to "any / cross-cutting"
// rather than to a specific algorithm in the AlgoIDs() catalogue.
//
// The constants follow the same CamelCase String() naming convention as
// AlgoID (see algoid.go) — the consumer-facing label for WarnEmptyInput
// is "EmptyInput", not "WarnEmptyInput" — so Warning rendering composes
// cleanly with AlgoID.String() in human-readable diagnostics.
//
// See docs/requirements.md §11.5 for the canonical specification and
// .planning/REQUIREMENTS.md VALIDATE-03 / VALIDATE-04 for the
// requirement-tracker entries.
//
// Implementation discipline:
//
//   - No init() — String() is a switch, WarnKinds() returns a slice
//     literal. Mirrors the AlgoID pattern.
//   - No map iteration on output paths (per
//     .claude/skills/determinism-standards).
//   - Constants start at iota + 1 so the zero value of WarnKind is
//     "unknown / unset" — defensive against accidental zero-initialisation
//     emitting a misleading WarnEmptyInput.

package fuzzymatch

import "fmt"

// WarnKind classifies the kind of input-quality concern Validate
// reports. WarnKind is a plain int (not int32 / int64 / struct) so it
// is trivially comparable and serialisable. The zero value is reserved
// as "unspecified" — every documented WarnKind starts at 1 (iota + 1).
//
// Values are stable across patch releases. The integer values themselves
// are PART of the v1.x contract: consumers may persist them, compare
// them, and rely on WarnEmptyInput evaluating to 1. Future additions
// append to the END of the const block — existing WarnKind values never
// shift (per docs/requirements.md §11.5 "Forward compatibility").
//
// Use WarnKinds() to enumerate the valid set; use String() to obtain
// the canonical CamelCase label.
type WarnKind int

// The 5 WarnKind constants documented in docs/requirements.md §11.5.
// Iota starts at 1 so the zero value remains "unset" — useful for
// detecting accidentally zero-initialised Warning values in consumer
// code.
const (
	// WarnEmptyInput indicates that at least one of the two inputs to
	// Validate was the empty string. Affects every algorithm: identity
	// short-circuits, Hamming returns trivial values, phonetic algorithms
	// produce empty keys that match each other. Emitted with
	// Algorithm == AlgoIDAny when surfaced by Validate.
	WarnEmptyInput WarnKind = iota + 1

	// WarnUnequalLength indicates the two inputs differ in length, which
	// the Hamming family (AlgoHamming) handles by silent-max policy —
	// the score collapses to 0.0 without any error. Consumers wanting
	// to detect that case explicitly check for this Kind. Emitted with
	// Algorithm == AlgoHamming.
	WarnUnequalLength

	// WarnNoTokensAfterNormalise indicates that, after the default
	// Normalise + Tokenise pipeline, one or both inputs produce an empty
	// token list. Affects the token-tier algorithms (TokenSortRatio,
	// TokenSetRatio, PartialRatio, TokenJaccard, MongeElkan) which would
	// compare empty token sets. Emitted once per affected algorithm.
	WarnNoTokensAfterNormalise

	// WarnAllNonASCIIDropped indicates that, after ASCII-only filtering,
	// one or both inputs collapse to empty. Affects the ASCII-only
	// algorithms (Strcmp95, Soundex, NYSIIS, MRA — and per §7.4 the
	// Double Metaphone fallback). Emitted once per affected algorithm.
	WarnAllNonASCIIDropped

	// WarnPathologicallyLargeInput indicates input size that risks
	// triggering O(m·n) DP-table cost on the quadratic algorithms
	// (DamerauLevenshteinFull, SmithWatermanGotoh, RatcliffObershelp,
	// MongeElkan, PartialRatio). Threshold is a documented per-algorithm
	// constant; see docs/algorithms.md. Consumers should gate at the
	// call site (e.g. reject inputs over a project-specific size) rather
	// than relying on this warning post-hoc.
	WarnPathologicallyLargeInput
)

// AlgoIDAny is the sentinel AlgoID value used to scope a Warning that is
// not algorithm-specific (e.g. WarnEmptyInput, which applies to every
// algorithm in the catalogue). It is declared OUTSIDE the AlgoLevenshtein..
// AlgoRatcliffObershelp iota block so it never participates in the
// dispatch table or in AlgoIDs() — Validate is the only public consumer.
//
// AlgoIDAny.String() returns "Any". Consumers comparing a Warning's
// Algorithm field to AlgoIDAny use plain equality:
//
//	if w.Algorithm == fuzzymatch.AlgoIDAny { /* cross-cutting */ }
//
// The value (-2) is chosen deliberately: AlgoID(-1) is used in tests as
// the canonical "out-of-range invalid AlgoID" sentinel (see
// scorer_options_test.go, algoid_test.go), and reserving -1 for that
// role keeps the negative-value contract clear: -1 means "invalid",
// AlgoIDAny (= -2) means "any / cross-cutting".
const AlgoIDAny AlgoID = -2

// String returns the canonical CamelCase label for k. For WarnEmptyInput
// the label is "EmptyInput" — the constant prefix is dropped to match
// the AlgoID.String() convention (where AlgoLevenshtein → "Levenshtein",
// not "AlgoLevenshtein").
//
// For an unknown WarnKind value (the zero value, or any future value
// declared after this method is compiled), String returns the fallback
// "WarnKind(N)" via fmt.Sprintf — intentionally allocating because the
// path is for diagnostic and error output, not a hot dispatch path.
//
// String never allocates on the in-range path: every case returns a
// compile-time string constant.
func (k WarnKind) String() string {
	switch k {
	case WarnEmptyInput:
		return "EmptyInput"
	case WarnUnequalLength:
		return "UnequalLength"
	case WarnNoTokensAfterNormalise:
		return "NoTokensAfterNormalise"
	case WarnAllNonASCIIDropped:
		return "AllNonASCIIDropped"
	case WarnPathologicallyLargeInput:
		return "PathologicallyLargeInput"
	default:
		return fmt.Sprintf("WarnKind(%d)", int(k))
	}
}

// WarnKinds returns the full set of 5 WarnKind constants in their
// declared order. The returned slice is freshly allocated on every call
// so the caller may mutate, sort, or filter it without affecting other
// callers. The order is deterministic and identical across runs,
// processes, and platforms — there is no map iteration on this path
// (per .claude/skills/determinism-standards).
//
// Useful for tests, documentation generation, and consumer discovery
// (e.g. iterating every Kind to surface a one-line description in a
// CLI help screen).
func WarnKinds() []WarnKind {
	return []WarnKind{
		WarnEmptyInput,
		WarnUnequalLength,
		WarnNoTokensAfterNormalise,
		WarnAllNonASCIIDropped,
		WarnPathologicallyLargeInput,
	}
}
