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

// suppress_test.go pins the Plan 09-05 internal contract of the
// suppression composition predicate. Internal tests live in package
// scan to exercise the unexported symbols (isSuppressed, canonicalPair,
// suppressionCtx, buildSuppressionCtx) directly.
//
// Coverage outline:
//
//   - canonicalPair lexicographic ordering (Lo / Hi by string compare).
//   - isSuppressed Rule 1 (SilenceLint) one-sided semantics.
//   - isSuppressed Rule 2 (SuppressedPairs canonical-pair lookup) with
//     order-independence verified by argument swap.
//   - isSuppressed Rule 3 (cross-group identical-name default) firing
//     only for KindAcrossGroups, gated by ctx.compareIdenticalAcrossGroups.
//   - buildSuppressionCtx normalisation pipeline (raw pairs canonicalise
//     to the normalised form when applyNormalisation == true).
//
// Stdlib testing only — no testify in root tests per CLAUDE.md.

package scan

import (
	"testing"

	"github.com/axonops/fuzzymatch"
)

// suppressItem is a small constructor that produces an Item with the
// supplied name, group, and SilenceLint flag. The other Item fields
// (Tag) are zero-valued — irrelevant to the suppression predicate.
func suppressItem(name, group string, silenced bool) Item {
	return Item{Name: name, Group: group, SilenceLint: silenced}
}

// --- canonicalPair tests --------------------------------------------------

// Test_canonicalPair_LexOrder verifies the canonicalPair helper sorts
// its inputs lexicographically into the (Lo, Hi) struct. Order
// independence is the load-bearing property — both call orders must
// produce the same pairKey value so the map lookup in Rule 2 catches
// either argument order.
func Test_canonicalPair_LexOrder(t *testing.T) {
	t.Parallel()

	k1 := canonicalPair("user_id", "userid")
	k2 := canonicalPair("userid", "user_id")

	if k1 != k2 {
		t.Errorf("canonicalPair order-dependence: k1=%+v k2=%+v", k1, k2)
	}
	if k1.Lo != "user_id" || k1.Hi != "userid" {
		t.Errorf("canonicalPair Lo/Hi: got (%q, %q); want (%q, %q)",
			k1.Lo, k1.Hi, "user_id", "userid")
	}
}

// Test_canonicalPair_EqualStrings verifies canonicalPair handles the
// degenerate equal-string case (a == b). Lo and Hi both equal the
// shared value. This is the D-05 self-pair case that's silently kept
// in the suppression map (Check never emits self-warnings so the entry
// is harmless).
func Test_canonicalPair_EqualStrings(t *testing.T) {
	t.Parallel()

	k := canonicalPair("abc", "abc")
	if k.Lo != "abc" || k.Hi != "abc" {
		t.Errorf("canonicalPair equal: got (%q, %q); want (\"abc\", \"abc\")",
			k.Lo, k.Hi)
	}
}

// Test_canonicalPair_EmptyVsNonEmpty pins the byte-compare semantics:
// the empty string is lexicographically smaller than any non-empty
// string. This case isn't reachable in production (D-03 rejects empty
// Item.Name and D-05 rejects empty SuppressedPairs entries) but the
// helper is total and the test documents the boundary.
func Test_canonicalPair_EmptyVsNonEmpty(t *testing.T) {
	t.Parallel()

	k := canonicalPair("", "x")
	if k.Lo != "" || k.Hi != "x" {
		t.Errorf("canonicalPair empty vs nonempty: got (%q, %q); want (\"\", \"x\")",
			k.Lo, k.Hi)
	}
}

// --- isSuppressed Rule 1: SilenceLint ------------------------------------

// Test_isSuppressed_Rule1_SilenceLintLeft verifies Rule 1 fires when the
// LEFT item carries SilenceLint=true; the right item is clean.
func Test_isSuppressed_Rule1_SilenceLintLeft(t *testing.T) {
	t.Parallel()

	ctx := suppressionCtx{}
	a := suppressItem("foo", "g", true)
	b := suppressItem("bar", "g", false)
	if !isSuppressed(a, b, KindWithinGroup, "foo", "bar", ctx) {
		t.Errorf("Rule 1 (left SilenceLint): expected true")
	}
}

// Test_isSuppressed_Rule1_SilenceLintRight verifies Rule 1 fires when
// the RIGHT item carries SilenceLint=true; the left item is clean.
// Demonstrates the one-sided suppression semantics: either side
// silences the pair.
func Test_isSuppressed_Rule1_SilenceLintRight(t *testing.T) {
	t.Parallel()

	ctx := suppressionCtx{}
	a := suppressItem("foo", "g", false)
	b := suppressItem("bar", "g", true)
	if !isSuppressed(a, b, KindWithinGroup, "foo", "bar", ctx) {
		t.Errorf("Rule 1 (right SilenceLint): expected true")
	}
}

// Test_isSuppressed_Rule1_BothSilenced verifies Rule 1 fires when BOTH
// items carry SilenceLint=true. Trivial case; included for symmetry
// with the one-sided tests above.
func Test_isSuppressed_Rule1_BothSilenced(t *testing.T) {
	t.Parallel()

	ctx := suppressionCtx{}
	a := suppressItem("foo", "g", true)
	b := suppressItem("bar", "g", true)
	if !isSuppressed(a, b, KindWithinGroup, "foo", "bar", ctx) {
		t.Errorf("Rule 1 (both SilenceLint): expected true")
	}
}

// Test_isSuppressed_Rule1_NeitherSilenced verifies Rule 1 does NOT fire
// when neither item carries SilenceLint. Falls through to Rule 2 / Rule
// 3 (both inactive in this ctx) and returns false.
func Test_isSuppressed_Rule1_NeitherSilenced(t *testing.T) {
	t.Parallel()

	ctx := suppressionCtx{compareIdenticalAcrossGroups: true} // disable Rule 3
	a := suppressItem("foo", "g", false)
	b := suppressItem("bar", "g", false)
	if isSuppressed(a, b, KindWithinGroup, "foo", "bar", ctx) {
		t.Errorf("Rule 1 (neither): expected false")
	}
}

// --- isSuppressed Rule 2: SuppressedPairs lookup --------------------------

// Test_isSuppressed_Rule2_PairPresent verifies Rule 2 fires when the
// canonical pair (lex-sorted) of the normalised names is present in
// suppressedPairs. The compareIdenticalAcrossGroups flag is set so
// Rule 3 is inactive — isolates Rule 2.
func Test_isSuppressed_Rule2_PairPresent(t *testing.T) {
	t.Parallel()

	ctx := suppressionCtx{
		suppressedPairs: map[pairKey]struct{}{
			{Lo: "a_norm", Hi: "b_norm"}: {},
		},
		compareIdenticalAcrossGroups: true, // disable Rule 3
	}
	a := suppressItem("aRaw", "g", false)
	b := suppressItem("bRaw", "g", false)
	if !isSuppressed(a, b, KindWithinGroup, "a_norm", "b_norm", ctx) {
		t.Errorf("Rule 2 (pair present): expected true")
	}
}

// Test_isSuppressed_Rule2_PairPresentReversed verifies Rule 2 still
// fires when the normalised names are passed in REVERSED lex order
// (b before a). canonicalPair absorbs order; the lookup catches either
// argument permutation.
func Test_isSuppressed_Rule2_PairPresentReversed(t *testing.T) {
	t.Parallel()

	ctx := suppressionCtx{
		suppressedPairs: map[pairKey]struct{}{
			{Lo: "a_norm", Hi: "b_norm"}: {},
		},
		compareIdenticalAcrossGroups: true,
	}
	a := suppressItem("aRaw", "g", false)
	b := suppressItem("bRaw", "g", false)
	// Pass b_norm first, a_norm second — canonicalPair sorts.
	if !isSuppressed(a, b, KindWithinGroup, "b_norm", "a_norm", ctx) {
		t.Errorf("Rule 2 (pair present, reversed args): expected true")
	}
}

// Test_isSuppressed_Rule2_PairAbsent verifies Rule 2 does NOT fire when
// the canonical pair is missing from suppressedPairs. Falls through to
// Rule 3 (inactive here) and returns false.
func Test_isSuppressed_Rule2_PairAbsent(t *testing.T) {
	t.Parallel()

	ctx := suppressionCtx{
		suppressedPairs:              map[pairKey]struct{}{{Lo: "x", Hi: "y"}: {}},
		compareIdenticalAcrossGroups: true,
	}
	a := suppressItem("aRaw", "g", false)
	b := suppressItem("bRaw", "g", false)
	if isSuppressed(a, b, KindWithinGroup, "a_norm", "b_norm", ctx) {
		t.Errorf("Rule 2 (pair absent): expected false")
	}
}

// --- isSuppressed Rule 3: cross-group identical-name default --------------

// Test_isSuppressed_Rule3_CrossGroupIdenticalDefault verifies Rule 3
// fires on a cross-group pair whose normalised names match, when
// compareIdenticalAcrossGroups is false (the DefaultConfig default).
func Test_isSuppressed_Rule3_CrossGroupIdenticalDefault(t *testing.T) {
	t.Parallel()

	ctx := suppressionCtx{compareIdenticalAcrossGroups: false}
	a := suppressItem("user_id", "login", false)
	b := suppressItem("user_id", "profile", false)
	if !isSuppressed(a, b, KindAcrossGroups, "x", "x", ctx) {
		t.Errorf("Rule 3 (cross-group identical, default): expected true")
	}
}

// Test_isSuppressed_Rule3_CrossGroupIdenticalDisabled verifies Rule 3
// does NOT fire when compareIdenticalAcrossGroups is true (consumer
// flipped the toggle). Confirms the toggle behaviour.
func Test_isSuppressed_Rule3_CrossGroupIdenticalDisabled(t *testing.T) {
	t.Parallel()

	ctx := suppressionCtx{compareIdenticalAcrossGroups: true}
	a := suppressItem("user_id", "login", false)
	b := suppressItem("user_id", "profile", false)
	if isSuppressed(a, b, KindAcrossGroups, "x", "x", ctx) {
		t.Errorf("Rule 3 (cross-group identical, disabled): expected false")
	}
}

// Test_isSuppressed_Rule3_NotAppliedToWithinGroup verifies Rule 3 only
// fires for KindAcrossGroups — a within-group pair with identical
// normalised names is NOT suppressed (within-group identical names
// would be a D-06 duplicate, so this case is unreachable in production,
// but the predicate must be defensive).
func Test_isSuppressed_Rule3_NotAppliedToWithinGroup(t *testing.T) {
	t.Parallel()

	ctx := suppressionCtx{compareIdenticalAcrossGroups: false}
	a := suppressItem("user_id", "login", false)
	b := suppressItem("user_id", "login", false)
	if isSuppressed(a, b, KindWithinGroup, "x", "x", ctx) {
		t.Errorf("Rule 3 (within-group identical): expected false (Rule 3 must only fire for KindAcrossGroups)")
	}
}

// Test_isSuppressed_NoRulesFire_ReturnsFalse verifies the predicate
// returns false on clean inputs (no SilenceLint, no SuppressedPairs
// match, no Rule 3 trigger). Sanity test for the default-deny posture.
func Test_isSuppressed_NoRulesFire_ReturnsFalse(t *testing.T) {
	t.Parallel()

	ctx := suppressionCtx{
		suppressedPairs:              map[pairKey]struct{}{},
		compareIdenticalAcrossGroups: true,
	}
	a := suppressItem("foo", "g1", false)
	b := suppressItem("bar", "g2", false)
	if isSuppressed(a, b, KindAcrossGroups, "foo", "bar", ctx) {
		t.Errorf("clean inputs: expected false")
	}
}

// --- buildSuppressionCtx tests --------------------------------------------

// Test_buildSuppressionCtx_NormalisesEntries verifies the raw
// SuppressedPairs entries are normalised at build time when
// applyNormalisation is true. The lookup must succeed against the
// NORMALISED canonical key. Pitfall 4 in 09-RESEARCH.md.
func Test_buildSuppressionCtx_NormalisesEntries(t *testing.T) {
	t.Parallel()

	// Raw pair has mixed case + underscore; both sides normalise to
	// "user id" (DefaultNormalisationOptions lowercases + collapses
	// separators).
	pairs := [][2]string{{"USER_ID", "user_id"}}
	normOpts := fuzzymatch.DefaultNormalisationOptions()
	ctx := buildSuppressionCtx(pairs, normOpts, true, false)

	// The canonical key in the map must be the normalised form's
	// canonical pair.
	wantKey := canonicalPair(
		fuzzymatch.Normalise("USER_ID", normOpts),
		fuzzymatch.Normalise("user_id", normOpts),
	)
	if _, ok := ctx.suppressedPairs[wantKey]; !ok {
		t.Errorf("buildSuppressionCtx: normalised canonical key %+v missing from map %+v",
			wantKey, ctx.suppressedPairs)
	}
}

// Test_buildSuppressionCtx_NoNormalisation verifies that when
// applyNormalisation is false (Scorer constructed WithoutNormalisation),
// SuppressedPairs entries are stored AS-IS — the raw input becomes the
// canonical key directly.
func Test_buildSuppressionCtx_NoNormalisation(t *testing.T) {
	t.Parallel()

	pairs := [][2]string{{"USER_ID", "user_id"}}
	normOpts := fuzzymatch.DefaultNormalisationOptions() // irrelevant when applyNormalisation==false
	ctx := buildSuppressionCtx(pairs, normOpts, false, false)

	wantKey := canonicalPair("USER_ID", "user_id")
	if _, ok := ctx.suppressedPairs[wantKey]; !ok {
		t.Errorf("buildSuppressionCtx (no normalisation): raw canonical key %+v missing", wantKey)
	}
}

// Test_buildSuppressionCtx_EmptyInput verifies the helper returns a
// usable (non-nil-map) ctx on empty input, so subsequent lookups don't
// nil-deref.
func Test_buildSuppressionCtx_EmptyInput(t *testing.T) {
	t.Parallel()

	ctx := buildSuppressionCtx(nil, fuzzymatch.DefaultNormalisationOptions(), true, false)
	if ctx.suppressedPairs == nil {
		t.Errorf("buildSuppressionCtx(nil): map is nil; lookups would nil-deref")
	}
	// Lookup against the empty map must return zero / false.
	if _, ok := ctx.suppressedPairs[canonicalPair("a", "b")]; ok {
		t.Errorf("empty ctx: unexpected hit on canonical key (a,b)")
	}
}

// Test_buildSuppressionCtx_FlagPassthrough verifies the
// compareIdenticalAcrossGroups argument propagates verbatim into the
// returned ctx — the field is read by isSuppressed Rule 3.
func Test_buildSuppressionCtx_FlagPassthrough(t *testing.T) {
	t.Parallel()

	ctxTrue := buildSuppressionCtx(nil, fuzzymatch.DefaultNormalisationOptions(), true, true)
	if !ctxTrue.compareIdenticalAcrossGroups {
		t.Errorf("compareIdenticalAcrossGroups: expected true after passthrough")
	}
	ctxFalse := buildSuppressionCtx(nil, fuzzymatch.DefaultNormalisationOptions(), true, false)
	if ctxFalse.compareIdenticalAcrossGroups {
		t.Errorf("compareIdenticalAcrossGroups: expected false after passthrough")
	}
}
