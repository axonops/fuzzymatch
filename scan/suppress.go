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

// suppress.go declares the Plan 09-05 suppression composition predicate
// — the inner-loop gate scan.Check consults before emitting a Warning.
// Three suppression rules compose via short-circuit OR; the cheapest
// rule fires first.
//
// Rule order (locked, lowest-cost first):
//
//   - Rule 1 — per-item SilenceLint (two bool reads). One-sided
//     semantics: setting the flag on EITHER side of a candidate pair
//     silences the pair.
//
//   - Rule 2 — SuppressedPairs canonical-pair lookup. The consumer-
//     supplied [][2]string of name pairs is normalised at Check entry
//     (via fuzzymatch.Normalise + the Scorer's NormalisationOptions),
//     then stored in a map[pairKey]struct{} keyed by the lex-sorted
//     canonical pair. The inner-loop predicate constructs the candidate
//     pair's canonical key from the pre-Normalised names and tests for
//     membership. Order-independent by construction (canonicalPair
//     lex-sorts).
//
//   - Rule 3 — cross-group identical-name default (SCAN-04). When
//     cfg.CompareIdenticalAcrossGroups == false (the DefaultConfig
//     default), pairs in the cross-group pass whose Normalised names
//     coincide are suppressed. Within-group pairs are NEVER suppressed
//     by Rule 3 — within-group identical-name candidates are unreachable
//     anyway, because D-06 rejects duplicate (Name, Group) at the
//     validation gate, and within-group items always share Group.
//
// Plan-09-03's inline cross-group identical-name check at the
// cross-group emission sites has been MIGRATED into isSuppressed
// Rule 3 by Plan 09-05 — both the naive and bucket cross-group
// emission paths now route through the same predicate. The migration
// is a behaviour-preserving refactor: the SCAN-04 default still
// fires, just from one centralised location.
//
// Pitfall 4 (09-RESEARCH.md): SuppressedPairs entries are normalised
// ONCE at Check entry. Consumers may pass raw forms ("USER_ID",
// "user_id") and the lookup still succeeds against items whose
// Normalised forms coincide. Without this normalisation step, raw-vs-
// Normalised mismatches would silently disable Rule 2.
//
// Determinism (per .claude/skills/determinism-standards):
//
//   - The suppressedPairs map is READ via direct lookup
//     (sc.suppressedPairs[...]) inside isSuppressed; it is NEVER
//     iterated on any output path. Iteration would expose Go's map-
//     iteration randomisation to the warning order, violating the
//     determinism contract.
//   - canonicalPair is a pure function with stable byte-compare
//     semantics; the same inputs produce the same output regardless of
//     call order, platform, or build.
//
// Concurrency: every function in this file is a pure predicate or
// pure constructor — no I/O, no goroutines, no package-global state
// (the suppressionCtx is built fresh per Check invocation). Safe for
// concurrent use across Check calls.

package scan

import (
	"github.com/axonops/fuzzymatch"
)

// pairKey is the lex-sorted canonical form of a name pair used as the
// map key in suppressionCtx.suppressedPairs. Lo is the lexicographically
// smaller of the two names; Hi is the larger. canonicalPair enforces
// Lo <= Hi for every constructed pairKey.
//
// The struct is comparable (both fields are strings) so it can be used
// as a map key without a custom Hash or Equals function.
//
// Package-private — pairKey is consumed by isSuppressed and
// buildSuppressionCtx exclusively. Not part of the public API surface.
type pairKey struct {
	// Lo is the lex-smaller of the two normalised names.
	Lo string
	// Hi is the lex-larger of the two normalised names. When the two
	// inputs are equal, Lo == Hi (the D-05 self-pair case — silently
	// kept in the suppression set because Check never emits a self-
	// warning).
	Hi string
}

// suppressionCtx carries the per-Check suppression state into the
// inner-loop predicate. It is built once at Check entry by
// buildSuppressionCtx and read-only for the rest of the Check
// invocation.
//
// Fields are package-private because the struct is consumed by
// isSuppressed alone. Not part of the public API surface.
//
// Concurrency: the struct is read-only after buildSuppressionCtx
// returns. Multiple concurrent isSuppressed calls on the same
// suppressionCtx are safe because Go's map read is concurrency-safe
// when no concurrent writes occur.
type suppressionCtx struct {
	// suppressedPairs is the canonical-pair set built from
	// cfg.SuppressedPairs at Check entry. Keys are lex-sorted canonical
	// pairs of the NORMALISED names; values are struct{} (set semantics
	// — only key presence is checked). A nil-vs-empty map distinction
	// is irrelevant: empty SuppressedPairs produces an empty (but
	// non-nil) map so direct lookup never nil-derefs.
	suppressedPairs map[pairKey]struct{}

	// compareIdenticalAcrossGroups mirrors Config.CompareIdenticalAcrossGroups.
	// When false (DefaultConfig), Rule 3 fires for cross-group pairs
	// with byte-identical normalised names. When true, Rule 3 is
	// effectively disabled — the consumer wants cross-group identical-
	// name pairs to surface.
	compareIdenticalAcrossGroups bool
}

// canonicalPair returns the lex-sorted (Lo, Hi) pairKey for a and b.
// Inputs whose byte ordering is a <= b are returned as pairKey{a, b};
// otherwise the swap pairKey{b, a} preserves the Lo <= Hi invariant.
//
// Order-independence is the load-bearing property: canonicalPair(x, y)
// and canonicalPair(y, x) return identical pairKey values, so a Rule 2
// lookup catches either argument permutation. Equal inputs collapse
// to pairKey{a, a} (D-05 self-pair).
//
// Pure: no allocations beyond the inline pairKey value-type
// construction. Safe for concurrent use.
func canonicalPair(a, b string) pairKey {
	if a <= b {
		return pairKey{Lo: a, Hi: b}
	}
	return pairKey{Lo: b, Hi: a}
}

// buildSuppressionCtx constructs the per-Check suppression context from
// the consumer's SuppressedPairs list, the Scorer's normalisation
// state, and the CompareIdenticalAcrossGroups toggle. Called once at
// Check entry; the returned suppressionCtx is read-only for the rest
// of the Check invocation.
//
// Pipeline:
//
//  1. Allocate the suppressedPairs map with a capacity hint of
//     len(pairs) — buildSuppressionCtx writes exactly one entry per
//     valid pair, so the map never rehashes during construction.
//
//  2. For each pair, normalise both sides (when applyNormalisation is
//     true) using fuzzymatch.Normalise with the supplied normOpts.
//     When applyNormalisation is false (Scorer constructed via
//     WithoutNormalisation), the raw strings are stored as-is.
//
//  3. Build the canonical key via canonicalPair on the normalised
//     forms; insert into the map.
//
// Self-pair entries (a == b after normalisation) are stored as
// pairKey{Lo: a, Hi: a} — silently kept per D-05. Check never emits a
// self-warning (the inner loop's i < j discipline + D-06's duplicate
// rejection guarantee no (i, i) pairs), so the entry is harmless.
//
// Pitfall 4 mitigation: the normalisation step is the SOLE reason this
// helper exists. Without it, raw consumer-supplied SuppressedPairs
// entries would never match against Normalised candidate pairs in the
// inner loop. The implementation deliberately mirrors the Scorer's
// Normalise pipeline (Pitfall 5 — no double-normalisation downstream).
//
// Cost: O(N) where N is len(pairs). The Normalise step is amortised
// against the canonical-pair build; subsequent lookups in the inner
// loop are O(1) per candidate.
//
// Safe for concurrent use: buildSuppressionCtx produces a fresh map on
// every call; the input pairs slice is read-only.
func buildSuppressionCtx(
	pairs [][2]string,
	normOpts fuzzymatch.NormalisationOptions,
	applyNormalisation bool,
	compareIdenticalAcrossGroups bool,
) suppressionCtx {
	// Allocate at len(pairs) — Go's map growth is amortised O(1) but
	// a correct capacity hint avoids the first rehash. nil-vs-zero is
	// equivalent for the empty case; we return a non-nil empty map so
	// callers can read via map index without checking for nil first.
	out := make(map[pairKey]struct{}, len(pairs))
	for _, p := range pairs {
		a, b := p[0], p[1]
		if applyNormalisation {
			a = fuzzymatch.Normalise(a, normOpts)
			b = fuzzymatch.Normalise(b, normOpts)
		}
		out[canonicalPair(a, b)] = struct{}{}
	}
	return suppressionCtx{
		suppressedPairs:              out,
		compareIdenticalAcrossGroups: compareIdenticalAcrossGroups,
	}
}

// isSuppressed reports whether a candidate pair (a, b) — already known
// to satisfy the Scorer's similarity threshold — should be suppressed
// from emission. Returns true on first rule firing; short-circuit
// evaluation in declaration order.
//
// Rule order (cheapest first):
//
//	Rule 1: a.SilenceLint || b.SilenceLint  → suppress
//	Rule 2: canonical pair in suppressedPairs → suppress
//	Rule 3: kind == KindAcrossGroups && na == nb && !compareIdenticalAcrossGroups → suppress
//
// Parameters:
//
//   - a, b: the candidate items (read for SilenceLint only; the names
//     used for Rule 2 lookup are normalisedNameA / normalisedNameB,
//     NOT a.Name / b.Name).
//   - kind: the candidate's classification (KindWithinGroup or
//     KindAcrossGroups). Rule 3 fires only for KindAcrossGroups.
//   - normalisedNameA, normalisedNameB: pre-Normalised names supplied
//     by Check (the parallel normalisedNames slice built at Check
//     entry). Passing pre-Normalised forms avoids re-Normalising in
//     the hot path.
//   - sc: the per-Check suppression context built by
//     buildSuppressionCtx. Read-only.
//
// Pure: no I/O, no allocations beyond the canonical-pair value-type
// construction inside Rule 2. Safe for concurrent use on shared
// suppressionCtx.
func isSuppressed(a, b Item, kind Kind, normalisedNameA, normalisedNameB string, sc suppressionCtx) bool {
	// Rule 1 — per-item SilenceLint (one-sided semantics).
	if a.SilenceLint || b.SilenceLint {
		return true
	}

	// Rule 2 — SuppressedPairs canonical-pair lookup. canonicalPair
	// lex-sorts so the lookup is order-independent regardless of the
	// caller's argument order.
	if _, ok := sc.suppressedPairs[canonicalPair(normalisedNameA, normalisedNameB)]; ok {
		return true
	}

	// Rule 3 — cross-group identical-name default (SCAN-04). Migrated
	// from Plan-09-03's inline check at the cross-group emission sites
	// into this predicate for unified semantics. Within-group pairs
	// are NEVER suppressed by Rule 3 — within-group identical-name
	// candidates are unreachable anyway (D-06 rejects duplicate (Name,
	// Group), and within-group items always share Group).
	if kind == KindAcrossGroups && normalisedNameA == normalisedNameB && !sc.compareIdenticalAcrossGroups {
		return true
	}

	return false
}
