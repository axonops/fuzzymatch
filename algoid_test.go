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

// algoid_test.go pins the public-API contract of algoid.go: the 23
// constants are stable, dense from zero, distinct in value and in
// String() form, and String() never returns the empty string or a
// whitespace-containing label. The harness also exercises the
// internal dispatch-table skeleton (via export_test.go's
// NumAlgorithmsForTest / DispatchLenForTest / DispatchEntryNilForTest)
// to assert it's sized for 23 entries with every slot nil at the
// Phase 1 state.
//
// Stdlib `testing` only — no testify in root tests, per
// .claude/skills/go-coding-standards.

package fuzzymatch_test

import (
	"strings"
	"testing"
	"testing/quick"
	"unicode"

	"github.com/axonops/fuzzymatch"
)

// TestAlgoIDs_Count_Is23 pins the catalogue size at exactly 23 (the
// v1.x contract from docs/requirements.md §7). Adding or removing an
// AlgoID is a major-version-bump event.
func TestAlgoIDs_Count_Is23(t *testing.T) {
	const want = 23
	got := len(fuzzymatch.AlgoIDs())
	if got != want {
		t.Errorf("len(AlgoIDs()) = %d; want %d (the spec catalogue is locked at 23 — adding an algorithm requires a major version bump)", got, want)
	}
}

// TestAlgoIDs_DeterministicOrder proves AlgoIDs() returns the same
// slice contents in the same order on every call. The slice is freshly
// allocated each time (so identity differs) but the elements MUST be
// byte-identical.
//
// This is the runtime gate against any future refactor that might
// build AlgoIDs() from map iteration (forbidden by DET-03 in
// .claude/skills/determinism-standards).
func TestAlgoIDs_DeterministicOrder(t *testing.T) {
	const iterations = 100
	baseline := fuzzymatch.AlgoIDs()
	for i := 0; i < iterations; i++ {
		got := fuzzymatch.AlgoIDs()
		if len(got) != len(baseline) {
			t.Fatalf("iteration %d: len(AlgoIDs()) = %d; want %d", i, len(got), len(baseline))
		}
		for j := range got {
			if got[j] != baseline[j] {
				t.Fatalf("iteration %d, index %d: AlgoIDs()[%d] = %v; baseline = %v (map-iteration leak?)", i, j, j, got[j], baseline[j])
			}
		}
	}
}

// TestAlgoIDs_Distinct asserts every entry in the catalogue is unique.
// Two AlgoIDs with the same integer value would collide on dispatch
// and break the round-trip String() invariant.
func TestAlgoIDs_Distinct(t *testing.T) {
	ids := fuzzymatch.AlgoIDs()
	// Use a map internally — map contents check, not iteration order
	// (test code internals are exempt from the no-map-iteration rule).
	seen := make(map[fuzzymatch.AlgoID]int, len(ids))
	for i, id := range ids {
		if prev, dup := seen[id]; dup {
			t.Errorf("AlgoIDs()[%d] = %v duplicates AlgoIDs()[%d]", i, id, prev)
		}
		seen[id] = i
	}
	if len(seen) != len(ids) {
		t.Errorf("distinct AlgoID count = %d; want %d", len(seen), len(ids))
	}
}

// TestAlgoIDs_DenseFromZero asserts the catalogue is a contiguous
// block starting at AlgoLevenshtein = 0. This pins the iota backing:
// consumers can rely on AlgoID(i) being the (i+1)-th catalogue entry
// for any 0 <= i < 23.
//
// Failure modes this catches: a future PR sneaking in `AlgoID = iota +
// 1` (shifts everything by one), or a `_ AlgoID = iota` blank entry
// that gaps the block, or a manually-assigned value that breaks the
// run.
func TestAlgoIDs_DenseFromZero(t *testing.T) {
	ids := fuzzymatch.AlgoIDs()
	if len(ids) == 0 || ids[0] != 0 {
		t.Fatalf("AlgoIDs()[0] = %v; want 0 (catalogue must start at zero per iota backing)", ids[0])
	}
	for i, id := range ids {
		if int(id) != i {
			t.Errorf("AlgoIDs()[%d] = %d (int); want %d (catalogue must be dense — no gaps, no offsets)", i, int(id), i)
		}
	}
}

// TestAlgoID_String_NotEmpty_ForEveryConstant exercises String() for
// every AlgoID in the catalogue and asserts the result is non-empty
// and contains no whitespace. The canonical labels are
// CamelCase-without-spaces by contract.
func TestAlgoID_String_NotEmpty_ForEveryConstant(t *testing.T) {
	for _, id := range fuzzymatch.AlgoIDs() {
		t.Run(id.String(), func(t *testing.T) {
			s := id.String()
			if s == "" {
				t.Errorf("AlgoID(%d).String() = empty", int(id))
			}
			for _, r := range s {
				if unicode.IsSpace(r) {
					t.Errorf("AlgoID(%d).String() = %q contains whitespace at rune %q", int(id), s, r)
					return
				}
			}
		})
	}
}

// TestAlgoID_String_OutOfRange exercises the fallback branch for an
// AlgoID outside the declared catalogue. The contract is the
// fmt.Sprintf("AlgoID(%d)", int(id)) form.
func TestAlgoID_String_OutOfRange(t *testing.T) {
	tests := []struct {
		id   fuzzymatch.AlgoID
		want string
	}{
		{fuzzymatch.AlgoID(999), "AlgoID(999)"},
		{fuzzymatch.AlgoID(-1), "AlgoID(-1)"},
		{fuzzymatch.AlgoID(100), "AlgoID(100)"},
	}
	for _, tt := range tests {
		t.Run(tt.want, func(t *testing.T) {
			got := tt.id.String()
			if got != tt.want {
				t.Errorf("AlgoID(%d).String() = %q; want %q", int(tt.id), got, tt.want)
			}
		})
	}
}

// TestAlgoID_String_StableAcrossCalls is a property test via
// testing/quick: for ANY AlgoID value (including out-of-range), two
// successive calls to String() return the same string. This is the
// determinism gate against any future refactor introducing
// time-dependent or random behaviour.
//
// The two String() invocations are stored in separate local variables
// to make it explicit (and to satisfy staticcheck's SA4000 rule, which
// flags identical sub-expressions across an operator — here the intent
// is to call the method twice and compare the results).
func TestAlgoID_String_StableAcrossCalls(t *testing.T) {
	f := func(id fuzzymatch.AlgoID) bool {
		first := id.String()
		second := id.String()
		return first == second
	}
	if err := quick.Check(f, nil); err != nil {
		t.Errorf("AlgoID.String() not stable across calls: %v", err)
	}
}

// TestAlgoID_RoundTrip asserts every AlgoID in the catalogue maps to
// a UNIQUE String() label. Two AlgoIDs sharing a label would break
// any downstream label-based round-trip (e.g. golden-file emission).
//
// A map is used internally; the assertion checks the COUNT (not the
// iteration order), so this conforms to the no-map-iteration-on-output
// rule (output here is a single int, not a slice or map).
func TestAlgoID_RoundTrip(t *testing.T) {
	ids := fuzzymatch.AlgoIDs()
	labels := make(map[string]fuzzymatch.AlgoID, len(ids))
	for _, id := range ids {
		label := id.String()
		if prev, dup := labels[label]; dup {
			t.Errorf("AlgoID %v.String() = %q duplicates AlgoID %v.String()", id, label, prev)
		}
		labels[label] = id
	}
	if len(labels) != len(ids) {
		t.Errorf("distinct label count = %d; want %d (catalogue String() labels must be unique)", len(labels), len(ids))
	}
}

// TestAlgoID_String_NoAlgoPrefix asserts canonical labels do NOT carry
// the "Algo" prefix that the constant names use — String() returns the
// algorithm name itself ("Levenshtein", not "AlgoLevenshtein").
// This pins the documented contract from algoid.go's String() godoc.
func TestAlgoID_String_NoAlgoPrefix(t *testing.T) {
	for _, id := range fuzzymatch.AlgoIDs() {
		s := id.String()
		if strings.HasPrefix(s, "Algo") {
			t.Errorf("AlgoID(%d).String() = %q must not carry the Algo prefix", int(id), s)
		}
	}
}

// TestDispatch_SizedForCatalogue asserts the internal dispatch array
// is sized for exactly the number of AlgoIDs. Exercised via the
// export_test.go re-exports so the test stays in the black-box package.
func TestDispatch_SizedForCatalogue(t *testing.T) {
	wantLen := len(fuzzymatch.AlgoIDs())
	if wantLen != fuzzymatch.NumAlgorithmsForTest {
		t.Errorf("NumAlgorithmsForTest = %d; want %d (size constant must match AlgoIDs() length)", fuzzymatch.NumAlgorithmsForTest, wantLen)
	}
	gotLen := fuzzymatch.DispatchLenForTest()
	if gotLen != wantLen {
		t.Errorf("DispatchLenForTest() = %d; want %d (dispatch array must be sized for the catalogue)", gotLen, wantLen)
	}
}

// TestDispatch_LevenshteinRegistered asserts that dispatch[AlgoLevenshtein]
// (slot 0) is non-nil after Phase 2 plan 02-01 registers LevenshteinScore.
// Wave 2 plans (02-02 through 02-06) further update this test as each
// algorithm populates its slot.
func TestDispatch_LevenshteinRegistered(t *testing.T) {
	if fuzzymatch.DispatchEntryNilForTest(int(fuzzymatch.AlgoLevenshtein)) {
		t.Errorf("dispatch[AlgoLevenshtein] (%d) is nil — dispatch_levenshtein.go must register LevenshteinScore at package load time",
			int(fuzzymatch.AlgoLevenshtein))
	}
}

// TestDispatch_HammingRegistered asserts that dispatch[AlgoHamming]
// (slot 3) is non-nil after Phase 2 plan 02-02 registers HammingScore.
func TestDispatch_HammingRegistered(t *testing.T) {
	if fuzzymatch.DispatchEntryNilForTest(int(fuzzymatch.AlgoHamming)) {
		t.Errorf("dispatch[AlgoHamming] (%d) is nil — dispatch_hamming.go must register HammingScore at package load time",
			int(fuzzymatch.AlgoHamming))
	}
}

// TestDispatch_JaroRegistered asserts that dispatch[AlgoJaro]
// (slot 4) is non-nil after Phase 2 plan 02-03 registers JaroScore.
func TestDispatch_JaroRegistered(t *testing.T) {
	if fuzzymatch.DispatchEntryNilForTest(int(fuzzymatch.AlgoJaro)) {
		t.Errorf("dispatch[AlgoJaro] (%d) is nil — dispatch_jaro.go must register JaroScore at package load time",
			int(fuzzymatch.AlgoJaro))
	}
}

// TestDispatch_DamerauLevenshteinOSARegistered asserts that
// dispatch[AlgoDamerauLevenshteinOSA] (slot 1) is non-nil after Phase 2 plan
// 02-05 registers DamerauLevenshteinOSAScore.
func TestDispatch_DamerauLevenshteinOSARegistered(t *testing.T) {
	if fuzzymatch.DispatchEntryNilForTest(int(fuzzymatch.AlgoDamerauLevenshteinOSA)) {
		t.Errorf("dispatch[AlgoDamerauLevenshteinOSA] (%d) is nil — dispatch_damerau_osa.go must register DamerauLevenshteinOSAScore at package load time",
			int(fuzzymatch.AlgoDamerauLevenshteinOSA))
	}
}

// TestDispatch_DamerauLevenshteinFullRegistered asserts that
// dispatch[AlgoDamerauLevenshteinFull] (slot 2) is non-nil after Phase 2 plan
// 02-06 registers DamerauLevenshteinFullScore.
func TestDispatch_DamerauLevenshteinFullRegistered(t *testing.T) {
	if fuzzymatch.DispatchEntryNilForTest(int(fuzzymatch.AlgoDamerauLevenshteinFull)) {
		t.Errorf("dispatch[AlgoDamerauLevenshteinFull] (%d) is nil — dispatch_damerau_full.go must register DamerauLevenshteinFullScore at package load time",
			int(fuzzymatch.AlgoDamerauLevenshteinFull))
	}
}

// TestDispatch_JaroWinklerRegistered asserts that dispatch[AlgoJaroWinkler]
// (slot 5) is non-nil after Phase 2 plan 02-04 registers JaroWinklerScore.
func TestDispatch_JaroWinklerRegistered(t *testing.T) {
	if fuzzymatch.DispatchEntryNilForTest(int(fuzzymatch.AlgoJaroWinkler)) {
		t.Errorf("dispatch[AlgoJaroWinkler] (%d) is nil — dispatch_jarowinkler.go must register JaroWinklerScore at package load time",
			int(fuzzymatch.AlgoJaroWinkler))
	}
}

// TestDispatch_SmithWatermanGotohRegistered asserts that
// dispatch[AlgoSmithWatermanGotoh] (slot 7) is non-nil after Phase 3 plan
// 03-01 registers SmithWatermanGotohScore.
func TestDispatch_SmithWatermanGotohRegistered(t *testing.T) {
	if fuzzymatch.DispatchEntryNilForTest(int(fuzzymatch.AlgoSmithWatermanGotoh)) {
		t.Errorf("dispatch[AlgoSmithWatermanGotoh] (%d) is nil — dispatch_swg.go must register SmithWatermanGotohScore at package load time",
			int(fuzzymatch.AlgoSmithWatermanGotoh))
	}
}

// TestDispatch_Strcmp95Registered asserts that dispatch[AlgoStrcmp95]
// (slot 6) is non-nil after Phase 4 plan 04-01 registers Strcmp95Score.
func TestDispatch_Strcmp95Registered(t *testing.T) {
	if fuzzymatch.DispatchEntryNilForTest(int(fuzzymatch.AlgoStrcmp95)) {
		t.Errorf("dispatch[AlgoStrcmp95] (%d) is nil — dispatch_strcmp95.go must register Strcmp95Score at package load time",
			int(fuzzymatch.AlgoStrcmp95))
	}
}

// TestDispatch_LCSStrRegistered asserts that dispatch[AlgoLCSStr] (slot 8)
// is non-nil after Phase 4 plan 04-02 registers LCSStrScore. Only the
// score-returning byte path is dispatched — LongestCommonSubstring*,
// LCSStrScoreRunes are public but not dispatched (the dispatch table maps
// AlgoID to (a, b string) float64).
func TestDispatch_LCSStrRegistered(t *testing.T) {
	if fuzzymatch.DispatchEntryNilForTest(int(fuzzymatch.AlgoLCSStr)) {
		t.Errorf("dispatch[AlgoLCSStr] (%d) is nil — dispatch_lcsstr.go must register LCSStrScore at package load time",
			int(fuzzymatch.AlgoLCSStr))
	}
}

// TestDispatch_RatcliffObershelpRegistered asserts that
// dispatch[AlgoRatcliffObershelp] (slot 22 — the LAST slot, numAlgorithms-1)
// is non-nil after Phase 4 plan 04-03 registers RatcliffObershelpScore.
// Only the byte-path score is dispatched — RatcliffObershelpScoreRunes is
// public but not dispatched (the dispatch table maps AlgoID to
// (a, b string) float64).
func TestDispatch_RatcliffObershelpRegistered(t *testing.T) {
	if fuzzymatch.DispatchEntryNilForTest(int(fuzzymatch.AlgoRatcliffObershelp)) {
		t.Errorf("dispatch[AlgoRatcliffObershelp] (%d) is nil — dispatch_ratcliff_obershelp.go must register RatcliffObershelpScore at package load time",
			int(fuzzymatch.AlgoRatcliffObershelp))
	}
}

// TestDispatch_QGramJaccardRegistered asserts that
// dispatch[AlgoQGramJaccard] (slot 9) is non-nil after Phase 5 plan
// 05-01 registers a QGramJaccardScore wrapper. The dispatch wrapper
// binds default n = 3 (the canonical trigram value) per CONTEXT.md
// Deferred §4 — the dispatch signature has no place for the n parameter,
// so n overrides happen via the Phase 8 Scorer option
// WithQGramJaccardAlgorithm(weight, n).
//
// Only QGramJaccardScore is dispatched — QGramJaccardScoreRunes is
// public but not dispatched (the dispatch table maps AlgoID to
// (a, b string) float64).
func TestDispatch_QGramJaccardRegistered(t *testing.T) {
	if fuzzymatch.DispatchEntryNilForTest(int(fuzzymatch.AlgoQGramJaccard)) {
		t.Errorf("dispatch[AlgoQGramJaccard] (%d) is nil — dispatch_qgram_jaccard.go must register a QGramJaccardScore wrapper at package load time",
			int(fuzzymatch.AlgoQGramJaccard))
	}
	// Exercise the closure body to confirm the dispatch wrapper actually
	// invokes QGramJaccardScore with the documented default n = 3 trigram.
	got := fuzzymatch.DispatchInvokeForTest(int(fuzzymatch.AlgoQGramJaccard), "hello", "hello")
	want := fuzzymatch.QGramJaccardScore("hello", "hello", 3)
	if got != want {
		t.Errorf("dispatch[AlgoQGramJaccard](\"hello\",\"hello\") = %v; want %v (= QGramJaccardScore with default n=3 per CONTEXT.md Deferred §4)",
			got, want)
	}
}

// TestDispatch_SorensenDiceRegistered asserts that
// dispatch[AlgoSorensenDice] (slot 10) is non-nil after Phase 5 plan
// 05-02 registers a SorensenDiceScore wrapper. The dispatch wrapper
// binds default n = 3 (the canonical trigram value) per CONTEXT.md
// Deferred §4 — the dispatch signature has no place for the n parameter,
// so n overrides happen via the Phase 8 Scorer option
// WithSorensenDiceAlgorithm(weight, n).
//
// Only SorensenDiceScore is dispatched — SorensenDiceScoreRunes is
// public but not dispatched (the dispatch table maps AlgoID to
// (a, b string) float64).
func TestDispatch_SorensenDiceRegistered(t *testing.T) {
	if fuzzymatch.DispatchEntryNilForTest(int(fuzzymatch.AlgoSorensenDice)) {
		t.Errorf("dispatch[AlgoSorensenDice] (%d) is nil — dispatch_sorensen_dice.go must register a SorensenDiceScore wrapper at package load time",
			int(fuzzymatch.AlgoSorensenDice))
	}
	// Exercise the closure body to confirm the dispatch wrapper actually
	// invokes SorensenDiceScore with the documented default n = 3 trigram.
	got := fuzzymatch.DispatchInvokeForTest(int(fuzzymatch.AlgoSorensenDice), "hello", "hello")
	want := fuzzymatch.SorensenDiceScore("hello", "hello", 3)
	if got != want {
		t.Errorf("dispatch[AlgoSorensenDice](\"hello\",\"hello\") = %v; want %v (= SorensenDiceScore with default n=3 per CONTEXT.md Deferred §4)",
			got, want)
	}
}

// TestDispatch_CosineRegistered asserts that dispatch[AlgoCosine]
// (slot 11) is non-nil after Phase 5 plan 05-03 registers a CosineScore
// wrapper. The dispatch wrapper binds default n = 3 (the canonical
// trigram value) per CONTEXT.md Deferred §4 — the dispatch signature
// has no place for the n parameter, so n overrides happen via the
// Phase 8 Scorer option WithCosineAlgorithm(weight, n).
//
// Only CosineScore is dispatched — CosineScoreRunes is public but not
// dispatched (the dispatch table maps AlgoID to (a, b string) float64).
func TestDispatch_CosineRegistered(t *testing.T) {
	if fuzzymatch.DispatchEntryNilForTest(int(fuzzymatch.AlgoCosine)) {
		t.Errorf("dispatch[AlgoCosine] (%d) is nil — dispatch_cosine.go must register a CosineScore wrapper at package load time",
			int(fuzzymatch.AlgoCosine))
	}
	// Exercise the closure body to confirm the dispatch wrapper actually
	// invokes CosineScore with the documented default n = 3 trigram.
	got := fuzzymatch.DispatchInvokeForTest(int(fuzzymatch.AlgoCosine), "hello", "hello")
	want := fuzzymatch.CosineScore("hello", "hello", 3)
	if got != want {
		t.Errorf("dispatch[AlgoCosine](\"hello\",\"hello\") = %v; want %v (= CosineScore with default n=3 per CONTEXT.md Deferred §4)",
			got, want)
	}
}

// TestDispatch_TverskyRegistered asserts that dispatch[AlgoTversky]
// (slot 12) is non-nil after Phase 5 plan 05-04 registers a
// TverskyScore wrapper. The dispatch wrapper binds default n = 3
// (the canonical trigram value) AND α = β = 1.0 (the Jaccard-equivalent
// weights) per CONTEXT.md "Claude's Discretion" — the dispatch
// signature has no place for n, α, or β, so the real Tversky use case
// (asymmetric direction-sensitive scoring with α ≠ β) lands in Phase 8
// via WithTverskyAlgorithm(weight, alpha, beta).
//
// Only TverskyScore is dispatched — TverskyScoreRunes is public but
// not dispatched (the dispatch table maps AlgoID to
// (a, b string) float64).
func TestDispatch_TverskyRegistered(t *testing.T) {
	if fuzzymatch.DispatchEntryNilForTest(int(fuzzymatch.AlgoTversky)) {
		t.Errorf("dispatch[AlgoTversky] (%d) is nil — dispatch_tversky.go must register a TverskyScore wrapper at package load time",
			int(fuzzymatch.AlgoTversky))
	}
	// Exercise the closure body to confirm the dispatch wrapper actually
	// invokes TverskyScore with the documented default n = 3 trigram AND
	// α = β = 1.0 (Jaccard-equivalent fallback per CONTEXT.md "Claude's
	// Discretion").
	got := fuzzymatch.DispatchInvokeForTest(int(fuzzymatch.AlgoTversky), "hello", "hello")
	want := fuzzymatch.TverskyScore("hello", "hello", 3, 1.0, 1.0)
	if got != want {
		t.Errorf("dispatch[AlgoTversky](\"hello\",\"hello\") = %v; want %v (= TverskyScore with default n=3, α=β=1.0 per CONTEXT.md \"Claude's Discretion\")",
			got, want)
	}
}

// TestDispatch_UnregisteredSlotsAreNil asserts that all dispatch slots except
// AlgoLevenshtein (slot 0), AlgoDamerauLevenshteinOSA (slot 1),
// AlgoDamerauLevenshteinFull (slot 2), AlgoHamming (slot 3), AlgoJaro
// (slot 4), AlgoJaroWinkler (slot 5), AlgoStrcmp95 (slot 6 — registered by
// Phase 4 plan 04-01), AlgoSmithWatermanGotoh (slot 7 — registered by Phase
// 3 plan 03-01), AlgoLCSStr (slot 8 — registered by Phase 4 plan 04-02),
// AlgoQGramJaccard (slot 9 — registered by Phase 5 plan 05-01),
// AlgoSorensenDice (slot 10 — registered by Phase 5 plan 05-02),
// AlgoCosine (slot 11 — registered by Phase 5 plan 05-03),
// AlgoTversky (slot 12 — registered by Phase 5 plan 05-04),
// AlgoMongeElkan (slot 13 — registered by Phase 6 plan 06-05; binds
// MongeElkanScore (the post-Phase-8.5-Q3 symmetric default) with
// AlgoJaroWinkler default inner per CONTEXT §4 LOCKED),
// AlgoTokenSortRatio (slot 14 — registered by Phase 6 plan 06-01),
// AlgoTokenSetRatio (slot 15 — registered by Phase 6 plan 06-02),
// AlgoPartialRatio (slot 16 — registered by Phase 6 plan 06-03; the
// byte-path PartialRatioScore is the sole surface after Phase 8.5 Q5
// removed the former rune-variant in plan 08.5-03),
// AlgoTokenJaccard (slot 17 — registered by Phase 6 plan 06-04),
// and AlgoRatcliffObershelp (slot 22 — the LAST slot, registered by
// Phase 4 plan 04-03) are still nil.
//
// Plan 06-05 flips slot 13 (AlgoMongeElkan) to registered.
// Plans 07-01..07-04 flip slots 18..21 (phonetic tier) to registered.
// FINAL Phase 7 state: all 23 slots registered.
func TestDispatch_UnregisteredSlotsAreNil(t *testing.T) {
	// Registered by Wave 1, plan 02-02..02-06, plan 03-01, plan 04-01,
	// plan 04-02, plan 04-03, plan 05-01, plan 05-02, plan 05-03,
	// plan 05-04, plan 06-01, plan 06-02, plan 06-03, plan 06-04, and
	// plan 06-05 respectively; all others nil.
	registered := map[int]bool{
		int(fuzzymatch.AlgoLevenshtein):            true,
		int(fuzzymatch.AlgoDamerauLevenshteinOSA):  true,
		int(fuzzymatch.AlgoDamerauLevenshteinFull): true,
		int(fuzzymatch.AlgoHamming):                true,
		int(fuzzymatch.AlgoJaro):                   true,
		int(fuzzymatch.AlgoJaroWinkler):            true,
		int(fuzzymatch.AlgoStrcmp95):               true,
		int(fuzzymatch.AlgoSmithWatermanGotoh):     true,
		int(fuzzymatch.AlgoLCSStr):                 true,
		int(fuzzymatch.AlgoQGramJaccard):           true,
		int(fuzzymatch.AlgoSorensenDice):           true,
		int(fuzzymatch.AlgoCosine):                 true,
		int(fuzzymatch.AlgoTversky):                true,
		int(fuzzymatch.AlgoMongeElkan):             true,
		int(fuzzymatch.AlgoTokenSortRatio):         true,
		int(fuzzymatch.AlgoTokenSetRatio):          true,
		int(fuzzymatch.AlgoPartialRatio):           true,
		int(fuzzymatch.AlgoTokenJaccard):           true,
		int(fuzzymatch.AlgoSoundex):                true, // registered by Phase 7 plan 07-01
		int(fuzzymatch.AlgoDoubleMetaphone):        true, // registered by Phase 7 plan 07-02
		int(fuzzymatch.AlgoNYSIIS):                 true, // registered by Phase 7 plan 07-03
		int(fuzzymatch.AlgoMRA):                    true, // registered by Phase 7 plan 07-04
		int(fuzzymatch.AlgoRatcliffObershelp):      true,
	}
	for i := 0; i < fuzzymatch.DispatchLenForTest(); i++ {
		isNil := fuzzymatch.DispatchEntryNilForTest(i)
		if registered[i] {
			if isNil {
				t.Errorf("dispatch[%d] is nil; expected non-nil (registered by Wave 1, plan 02-02..02-06, plan 03-01, plan 04-01, plan 04-02, plan 04-03, plan 05-01, plan 05-02, plan 05-03, plan 05-04, plan 06-01, plan 06-02, plan 06-03, plan 06-04, or plan 06-05)", i)
			}
		} else {
			if !isNil {
				t.Errorf("dispatch[%d] is non-nil; expected nil until a later plan registers slot %d", i, i)
			}
		}
	}
}

// BenchmarkAlgoID_String pins zero-allocation behaviour for the
// in-range hot path of String(). `b.ReportAllocs()` is the gate:
// future refactors that introduce a map lookup, fmt.Sprintf, or any
// other allocating pattern surface as a regression in bench.txt.
//
// The benchmark iterates over the catalogue rather than a single
// constant so the optimiser can't fold the switch into a constant
// return.
func BenchmarkAlgoID_String(b *testing.B) {
	ids := fuzzymatch.AlgoIDs()
	b.ReportAllocs()
	b.ResetTimer()
	var sink string
	for i := 0; i < b.N; i++ {
		sink = ids[i%len(ids)].String()
	}
	// Discourage dead-code elimination of sink.
	if sink == "" {
		b.Fatal("sink unexpectedly empty — compiler folded the benchmark away")
	}
}
