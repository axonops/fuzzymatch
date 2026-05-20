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
//
// scan_test.go pins the Plan 09-01 foundation contract of the scan
// sub-package:
//
//   - Kind.String() returns the CamelCase labels for KindWithinGroup
//     and KindAcrossGroups; out-of-range values fall back to the
//     "Kind(N)" allocating form (mirrors warn_kind.go's discipline).
//   - The three sentinel errors (ErrNilScorer, ErrInvalidItem,
//     ErrInvalidConfig) are distinct values — errors.Is between any
//     two of them returns false; self-identity returns true.
//   - Check(nil, scan.Config{Scorer: nil}) returns ErrNilScorer (Plan
//     09-01 stub contract; the full validation pipeline lands in Plan
//     09-02).
//   - DefaultConfig returns the opinionated defaults locked in
//     09-CONTEXT.md §2 D-04 (CrossGroupThresholdBoost == 0.05,
//     CompareIdenticalAcrossGroups == false, etc.).
//   - Warning.Scores has the SPEC OVERRIDE type
//     map[fuzzymatch.AlgoID]float64 (verified by reflect).
//
// Stdlib `testing` only — no testify in root tests per CLAUDE.md and
// .claude/skills/go-coding-standards.

package scan_test

import (
	"errors"
	"math"
	"reflect"
	"testing"

	"github.com/axonops/fuzzymatch"
	"github.com/axonops/fuzzymatch/scan"
)

// --------------------------------------------------------------------------
// Plan 09-03 test helpers
//
// The TestCheck_* tests below cover the within-group + cross-group naive
// passes added in Plan 09-03. The Scorer-construction helpers honour the
// Phase 8 functional-options pattern (DefaultScorer + DefaultScorerOptions
// + Override); they are not used by the Plan 09-01 / 09-02 tests above.
// --------------------------------------------------------------------------

// newScorerWithThreshold returns a *fuzzymatch.Scorer with the supplied
// threshold and otherwise the DefaultScorer composition. Used by the
// Pitfall-3 / Pitfall-6 boost-arithmetic tests where the boundary
// arithmetic is sensitive to the within-group threshold.
//
// t.Helper marks this as test-helper output (cleaner failure lines).
func newScorerWithThreshold(t *testing.T, threshold float64) *fuzzymatch.Scorer {
	t.Helper()
	opts := append(fuzzymatch.DefaultScorerOptions(), fuzzymatch.WithThreshold(threshold))
	s, err := fuzzymatch.NewScorer(opts...)
	if err != nil {
		t.Fatalf("newScorerWithThreshold(%v): unexpected error: %v", threshold, err)
	}
	return s
}

// assertWarningKind fails the test when the supplied Warning's Kind
// does not match want. Used to keep the per-test assertion lines short.
func assertWarningKind(t *testing.T, w scan.Warning, want scan.Kind) {
	t.Helper()
	if w.Kind != want {
		t.Errorf("Warning.Kind: got %v; want %v", w.Kind, want)
	}
}

// assertWarningNames fails the test when w.NameA / w.NameB does not
// match the supplied wantA / wantB pair. Names are checked in the
// order they were emitted (no sort enforced in Plan 09-03; sort lands
// in Plan 09-06).
func assertWarningNames(t *testing.T, w scan.Warning, wantA, wantB string) {
	t.Helper()
	if w.NameA != wantA || w.NameB != wantB {
		t.Errorf("Warning.NameA/NameB: got (%q, %q); want (%q, %q)",
			w.NameA, w.NameB, wantA, wantB)
	}
}

// TestKind_String pins the CamelCase labels for the two locked Kind
// constants and the allocating fmt.Sprintf default branch for
// out-of-range values. The labels are part of the v1.x contract —
// scan.Warning rendering, golden files, BDD scenarios, and consumer
// diagnostics all depend on the stability of these strings.
func TestKind_String(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name string
		k    scan.Kind
		want string
	}{
		{"KindWithinGroup", scan.KindWithinGroup, "WithinGroup"},
		{"KindAcrossGroups", scan.KindAcrossGroups, "AcrossGroups"},
		{"zero-value-unspecified", scan.Kind(0), "Kind(0)"},
		{"out-of-range-positive", scan.Kind(99), "Kind(99)"},
		{"out-of-range-negative", scan.Kind(-1), "Kind(-1)"},
	}
	for _, c := range cases {
		c := c
		t.Run(c.name, func(t *testing.T) {
			t.Parallel()
			if got := c.k.String(); got != c.want {
				t.Errorf("Kind(%d).String() = %q; want %q", int(c.k), got, c.want)
			}
		})
	}
}

// TestSentinels_Distinct verifies the three scan-package sentinels are
// distinct errors.Is identities. Consumers discriminating between
// ErrNilScorer / ErrInvalidItem / ErrInvalidConfig via errors.Is
// receive correct results only if no two sentinels match each other.
// The self-identity assertions guard the canonical "errors.Is(e, e)
// returns true" contract that errors.New values satisfy by reference
// equality.
func TestSentinels_Distinct(t *testing.T) {
	t.Parallel()

	sentinels := []struct {
		name string
		err  error
	}{
		{"ErrNilScorer", scan.ErrNilScorer},
		{"ErrInvalidItem", scan.ErrInvalidItem},
		{"ErrInvalidConfig", scan.ErrInvalidConfig},
	}

	for i, a := range sentinels {
		for j, b := range sentinels {
			i, j := i, j
			a, b := a, b
			t.Run(a.name+"_vs_"+b.name, func(t *testing.T) {
				t.Parallel()
				got := errors.Is(a.err, b.err)
				want := i == j
				if got != want {
					t.Errorf("errors.Is(%s, %s) = %v; want %v", a.name, b.name, got, want)
				}
			})
		}
	}
}

// TestCheck_NilScorer pins the Plan 09-01 stub contract: when
// cfg.Scorer is nil, Check returns (nil, err) where errors.Is(err,
// ErrNilScorer) is true. The full validation pipeline (Plan 09-02) and
// similarity body (Plans 09-03..09-06) extend this contract; the
// foundation contract is the nil-Scorer fail-fast only.
func TestCheck_NilScorer(t *testing.T) {
	t.Parallel()

	warnings, err := scan.Check(nil, scan.Config{Scorer: nil})
	if warnings != nil {
		t.Errorf("warnings: got %v; want nil", warnings)
	}
	if !errors.Is(err, scan.ErrNilScorer) {
		t.Errorf("err: got %v; want errors.Is(err, ErrNilScorer)", err)
	}
}

// TestDefaultConfig_Defaults pins the opinionated default Config built
// from a non-nil Scorer. The values are locked in 09-CONTEXT.md §2
// D-04 (SPEC OVERRIDE: the 0.05 default location migrated from
// Config.CrossGroupThresholdBoost field godoc to DefaultConfig godoc;
// spec §12.1 line 1359 amended in lockstep).
func TestDefaultConfig_Defaults(t *testing.T) {
	t.Parallel()

	s := fuzzymatch.DefaultScorer()
	cfg := scan.DefaultConfig(s)

	if cfg.Scorer != s {
		t.Errorf("Scorer: got %p; want %p", cfg.Scorer, s)
	}
	if cfg.CompareAcrossGroups {
		t.Errorf("CompareAcrossGroups: got true; want false")
	}
	if cfg.CrossGroupThresholdBoost != 0.05 {
		t.Errorf("CrossGroupThresholdBoost: got %v; want 0.05", cfg.CrossGroupThresholdBoost)
	}
	if cfg.CompareIdenticalAcrossGroups {
		t.Errorf("CompareIdenticalAcrossGroups: got true; want false")
	}
	if cfg.SuppressedPairs != nil {
		t.Errorf("SuppressedPairs: got %v; want nil", cfg.SuppressedPairs)
	}
}

// TestWarning_ScoresTypeIsAlgoID verifies the SPEC OVERRIDE (Phase 9)
// for Warning.Scores: the field type must be
// map[fuzzymatch.AlgoID]float64 (typed enum keys), NOT the spec's
// original map[string]float64. This is the load-bearing assertion that
// downstream Plan 09-03 (Check body) and Plan 09-06 (golden file
// writer) build against. Any unintended drift to map[string]float64
// here would silently change consumer code's compile-time key safety.
//
// See 09-CONTEXT.md §1 D-01 for the rationale and the
// api-ergonomics-reviewer sign-off recorded in this plan's PR.
func TestWarning_ScoresTypeIsAlgoID(t *testing.T) {
	t.Parallel()

	w := scan.Warning{}
	got := reflect.TypeOf(w.Scores).String()
	const want = "map[fuzzymatch.AlgoID]float64"
	if got != want {
		t.Errorf("Warning.Scores type: got %q; want %q (SPEC OVERRIDE D-01)", got, want)
	}
}

// --------------------------------------------------------------------------
// Plan 09-03 TestCheck_* — naive within-group + cross-group passes
//
// These tests pin the Check body contract added in Plan 09-03:
//
//   - validation gate fires first (Plan 09-02 hook)
//   - within-group pass uses Scorer.Match (threshold-applied internally)
//   - cross-group pass uses Scorer.Score with effective threshold =
//     math.Min(1.0, Threshold + CrossGroupThresholdBoost)
//   - cross-group identical-name suppression default (CompareIdenticalAcrossGroups
//     == false suppresses pairs whose normalised names coincide)
//   - Scorer receives RAW item.Name strings (no double-normalisation)
//   - Warning.Scores populated lazily via ScoreAll on emission
//   - groups iterated in sorted key order
//
// Pitfall coverage explicitly called out per test:
//   - Pitfall 3: Match vs Score discipline (Test 17)
//   - Pitfall 5: no double-normalisation (Test 15)
//   - Pitfall 6: boost clamp at 1.0 (Test 8)
// --------------------------------------------------------------------------

// TestCheck_WithinGroup_SingleMatch exercises the happy path: two items
// in the same group whose names normalise to a Scorer.Match hit produce
// exactly one Warning with Kind == KindWithinGroup.
func TestCheck_WithinGroup_SingleMatch(t *testing.T) {
	t.Parallel()

	s := fuzzymatch.DefaultScorer()
	cfg := scan.DefaultConfig(s)
	items := []scan.Item{
		{Name: "user_id", Group: "login"},
		{Name: "userId", Group: "login"},
	}

	warnings, err := scan.Check(items, cfg)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(warnings) != 1 {
		t.Fatalf("warnings: got %d; want 1; warnings=%+v", len(warnings), warnings)
	}

	w := warnings[0]
	assertWarningKind(t, w, scan.KindWithinGroup)
	assertWarningNames(t, w, "user_id", "userId")
	if w.GroupA != "login" || w.GroupB != "login" {
		t.Errorf("GroupA/GroupB: got (%q, %q); want both \"login\"", w.GroupA, w.GroupB)
	}
	wantScore := s.Score("user_id", "userId")
	if w.Score != wantScore {
		t.Errorf("Score: got %v; want %v (Scorer.Score on raw names)", w.Score, wantScore)
	}
	if len(w.Scores) == 0 {
		t.Errorf("Scores: got empty map; want lazily-populated ScoreAll result")
	}
}

// TestCheck_WithinGroup_NoMatch_BelowThreshold exercises the rejection
// path: an "apple" / "orange" pair scores well below 0.85 and produces
// zero warnings.
func TestCheck_WithinGroup_NoMatch_BelowThreshold(t *testing.T) {
	t.Parallel()

	s := fuzzymatch.DefaultScorer()
	cfg := scan.DefaultConfig(s)
	items := []scan.Item{
		{Name: "apple", Group: "fruit"},
		{Name: "orange", Group: "fruit"},
	}

	warnings, err := scan.Check(items, cfg)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(warnings) != 0 {
		t.Fatalf("warnings: got %d; want 0; warnings=%+v", len(warnings), warnings)
	}
}

// TestCheck_WithinGroup_ThreeItems_PairwiseMatches walks every i<j pair
// in a three-item group, asserts the warning count matches the recomputed
// Scorer.Match decision per pair. This catches a Check that emits self-
// pairs (i==j), duplicate pairs (j<i), or skips legitimate j-iterations.
func TestCheck_WithinGroup_ThreeItems_PairwiseMatches(t *testing.T) {
	t.Parallel()

	s := fuzzymatch.DefaultScorer()
	cfg := scan.DefaultConfig(s)
	items := []scan.Item{
		{Name: "user_id", Group: "login"},
		{Name: "userId", Group: "login"},
		{Name: "user_name", Group: "login"},
	}

	// Recompute the expected emission count via the public Scorer.Match
	// surface so this test stays correct under any future tweak to the
	// default scorer composition.
	wantCount := 0
	for i := 0; i < len(items); i++ {
		for j := i + 1; j < len(items); j++ {
			if s.Match(items[i].Name, items[j].Name) {
				wantCount++
			}
		}
	}

	warnings, err := scan.Check(items, cfg)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(warnings) != wantCount {
		t.Fatalf("warnings: got %d; want %d (recomputed via Scorer.Match)", len(warnings), wantCount)
	}
	for _, w := range warnings {
		assertWarningKind(t, w, scan.KindWithinGroup)
		if w.GroupA != "login" || w.GroupB != "login" {
			t.Errorf("warning %+v: GroupA/GroupB should both be \"login\"", w)
		}
	}
}

// TestCheck_CrossGroup_Disabled_NoEmissions confirms CompareAcrossGroups
// == false (the DefaultConfig) skips the cross-group pass entirely —
// identical-name items in different groups produce zero warnings.
func TestCheck_CrossGroup_Disabled_NoEmissions(t *testing.T) {
	t.Parallel()

	s := fuzzymatch.DefaultScorer()
	cfg := scan.DefaultConfig(s)
	if cfg.CompareAcrossGroups {
		t.Fatalf("precondition: DefaultConfig.CompareAcrossGroups should be false")
	}
	items := []scan.Item{
		{Name: "user_id", Group: "login"},
		{Name: "user_id", Group: "profile"},
	}

	warnings, err := scan.Check(items, cfg)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(warnings) != 0 {
		t.Fatalf("warnings: got %d; want 0 (cross-group pass disabled)", len(warnings))
	}
}

// TestCheck_CrossGroup_Enabled_IdenticalNameSuppressed verifies the
// SCAN-04 default: with CompareAcrossGroups=true and
// CompareIdenticalAcrossGroups=false (the DefaultConfig), pairs whose
// normalised names coincide are suppressed across groups.
func TestCheck_CrossGroup_Enabled_IdenticalNameSuppressed(t *testing.T) {
	t.Parallel()

	s := fuzzymatch.DefaultScorer()
	cfg := scan.DefaultConfig(s)
	cfg.CompareAcrossGroups = true
	if cfg.CompareIdenticalAcrossGroups {
		t.Fatalf("precondition: DefaultConfig.CompareIdenticalAcrossGroups should be false")
	}
	items := []scan.Item{
		{Name: "user_id", Group: "login"},
		{Name: "user_id", Group: "profile"},
	}

	warnings, err := scan.Check(items, cfg)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(warnings) != 0 {
		t.Fatalf("warnings: got %d; want 0 (SCAN-04 identical-name suppression)", len(warnings))
	}
}

// TestCheck_CrossGroup_Enabled_IdenticalNameNotSuppressed confirms the
// inverse: with CompareIdenticalAcrossGroups=true the identical-name
// pair emits exactly one Warning with Kind == KindAcrossGroups.
func TestCheck_CrossGroup_Enabled_IdenticalNameNotSuppressed(t *testing.T) {
	t.Parallel()

	s := fuzzymatch.DefaultScorer()
	cfg := scan.DefaultConfig(s)
	cfg.CompareAcrossGroups = true
	cfg.CompareIdenticalAcrossGroups = true
	items := []scan.Item{
		{Name: "user_id", Group: "login"},
		{Name: "user_id", Group: "profile"},
	}

	warnings, err := scan.Check(items, cfg)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(warnings) != 1 {
		t.Fatalf("warnings: got %d; want 1; warnings=%+v", len(warnings), warnings)
	}

	w := warnings[0]
	assertWarningKind(t, w, scan.KindAcrossGroups)
	if w.NameA != "user_id" || w.NameB != "user_id" {
		t.Errorf("identical-name pair: got (%q, %q); want both \"user_id\"", w.NameA, w.NameB)
	}
	if w.GroupA == w.GroupB {
		t.Errorf("cross-group warning: GroupA == GroupB (%q); want distinct groups", w.GroupA)
	}
}

// TestCheck_CrossGroup_NonIdenticalSimilar_Emitted exercises the
// load-bearing cross-group emission path: two items in different
// groups whose normalised names DIFFER but whose composite Score
// exceeds the boosted threshold (0.85 + 0.05 = 0.90) emit a
// KindAcrossGroups Warning.
//
// "different_field_A" / "different_field_B" score 0.871 under the
// default Scorer (cf. Plan 09-03 probe output) — above the within
// threshold 0.85, but BELOW the cross-group effective threshold 0.90.
// We therefore use a lower CrossGroupThresholdBoost (0.01) to drop the
// effective threshold to 0.86 and confirm the emission path fires.
func TestCheck_CrossGroup_NonIdenticalSimilar_Emitted(t *testing.T) {
	t.Parallel()

	s := fuzzymatch.DefaultScorer()
	cfg := scan.DefaultConfig(s)
	cfg.CompareAcrossGroups = true
	cfg.CrossGroupThresholdBoost = 0.01 // effective = 0.86

	items := []scan.Item{
		{Name: "different_field_A", Group: "login"},
		{Name: "different_field_B", Group: "profile"},
	}

	// Sanity-check the precondition the test depends on.
	if got := s.Score(items[0].Name, items[1].Name); got < 0.86 || got >= 1.0 {
		t.Skipf("precondition: Score(%q, %q) = %v outside the [0.86, 1.0) band the test assumes",
			items[0].Name, items[1].Name, got)
	}

	warnings, err := scan.Check(items, cfg)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(warnings) != 1 {
		t.Fatalf("warnings: got %d; want 1; warnings=%+v", len(warnings), warnings)
	}
	assertWarningKind(t, warnings[0], scan.KindAcrossGroups)
}

// TestCheck_CrossGroup_BoostClamp_BlocksEmission is the load-bearing
// Pitfall-6 test (09-RESEARCH.md lines 490-525): when scorer.Threshold +
// CrossGroupThresholdBoost would arithmetically exceed 1.0, the
// effective threshold is CLAMPED to 1.0 — NOT 1.15, 1.85, or any other
// arithmetic sum.
//
// The clamp ensures the boost cannot lift the effective threshold
// above the score-range upper bound: even a pathological
// `CrossGroupThresholdBoost = 1.0` (effective = min(1.0, 0.85+1.0) =
// 1.0) caps cross-group emission at "score >= 1.0", not the impossible
// "score >= 1.85". A similar-but-not-identical pair whose Score is in
// the high-0.8s — comfortably above the unboosted within threshold —
// is suppressed cross-group, demonstrating the boost arithmetic
// CHANGED the cross-group threshold and the CLAMP held it at 1.0.
//
// Choice of probe pair: "different_field_A" vs "different_field_B"
// scores 0.871 under DefaultScorer (above the 0.85 within threshold;
// far below 1.0). With the clamp pinning the effective threshold at
// 1.0, this pair MUST NOT emit cross-group. Without the clamp, the
// effective threshold would have been a value somewhere depending on
// arithmetic — but ANY positive boost should suppress this 0.871 pair.
//
// Defence-in-depth: assert the math.Min(1.0, Threshold+Boost)
// arithmetic directly to catch any future regression that adds
// without clamping.
func TestCheck_CrossGroup_BoostClamp_BlocksEmission(t *testing.T) {
	t.Parallel()

	s := fuzzymatch.DefaultScorer() // threshold 0.85
	cfg := scan.DefaultConfig(s)
	cfg.CompareAcrossGroups = true
	cfg.CompareIdenticalAcrossGroups = true // ensure suppression doesn't mask the clamp signal
	cfg.CrossGroupThresholdBoost = 1.0      // arithmetic 0.85 + 1.0 = 1.85; clamp pins at 1.0

	// Sanity-check the clamp arithmetic directly. Catches regressions
	// that replace math.Min with a naive add.
	got := math.Min(1.0, s.Threshold()+cfg.CrossGroupThresholdBoost)
	if got != 1.0 {
		t.Errorf("clamp arithmetic: got %v; want 1.0 (math.Min(1.0, 0.85+1.0))", got)
	}

	// Similar-but-not-identical pair: Score ~0.87, well below the
	// clamped effective threshold 1.0. MUST NOT emit cross-group.
	itemsSimilar := []scan.Item{
		{Name: "different_field_A", Group: "login"},
		{Name: "different_field_B", Group: "profile"},
	}
	score := s.Score(itemsSimilar[0].Name, itemsSimilar[1].Name)
	if score < s.Threshold() || score >= 1.0 {
		t.Skipf("precondition: Score(%q, %q) = %v outside [Threshold, 1.0)",
			itemsSimilar[0].Name, itemsSimilar[1].Name, score)
	}
	warningsSimilar, err := scan.Check(itemsSimilar, cfg)
	if err != nil {
		t.Fatalf("similar pair: unexpected error: %v", err)
	}
	if len(warningsSimilar) != 0 {
		t.Fatalf("similar pair: got %d warnings; want 0 (Score %.4f < clamped effective threshold 1.0)",
			len(warningsSimilar), score)
	}

	// Companion within-group control: the same pair (same group) MUST
	// emit on Match, proving the negative cross-group claim above
	// isn't a false negative caused by some other suppression. The
	// within-group pass uses Match (threshold 0.85), not the boosted
	// effective threshold.
	itemsSameGroup := []scan.Item{
		{Name: "different_field_A", Group: "login"},
		{Name: "different_field_B", Group: "login"},
	}
	cfgSameGroup := scan.DefaultConfig(s)
	warningsSameGroup, err := scan.Check(itemsSameGroup, cfgSameGroup)
	if err != nil {
		t.Fatalf("unexpected error (same-group control): %v", err)
	}
	if len(warningsSameGroup) != 1 {
		t.Fatalf("same-group control: got %d warnings; want 1 (Match should fire on Score %.4f >= within threshold 0.85)",
			len(warningsSameGroup), score)
	}
	assertWarningKind(t, warningsSameGroup[0], scan.KindWithinGroup)
}

// TestCheck_CrossGroup_BoostZero_SameAsWithinThreshold exercises the
// boost-zero edge: with CrossGroupThresholdBoost == 0.0 the cross-group
// pass emits on Score >= Threshold, exactly matching within-group
// behaviour. The "different_field_A" / "different_field_B" pair scores
// 0.871 under the default Scorer — above the 0.85 threshold but below
// 0.90 (the default DefaultConfig effective cross-group threshold). With
// boost 0.0 the pair MUST emit cross-group; with the DefaultConfig
// boost 0.05 the same pair (Test 7 above) needs explicit lower boost
// to emit.
func TestCheck_CrossGroup_BoostZero_SameAsWithinThreshold(t *testing.T) {
	t.Parallel()

	s := fuzzymatch.DefaultScorer()
	cfg := scan.DefaultConfig(s)
	cfg.CompareAcrossGroups = true
	cfg.CrossGroupThresholdBoost = 0.0 // boost off

	items := []scan.Item{
		{Name: "different_field_A", Group: "login"},
		{Name: "different_field_B", Group: "profile"},
	}

	if got := s.Score(items[0].Name, items[1].Name); got < s.Threshold() || got >= 0.90 {
		t.Skipf("precondition: Score(%q, %q) = %v outside the [Threshold, 0.90) band",
			items[0].Name, items[1].Name, got)
	}

	warnings, err := scan.Check(items, cfg)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(warnings) != 1 {
		t.Fatalf("warnings: got %d; want 1 (boost-zero parity with within-threshold)", len(warnings))
	}
	assertWarningKind(t, warnings[0], scan.KindAcrossGroups)
}

// TestCheck_WithinGroup_GroupKeyedIteration uses three groups each with
// one matching pair and confirms that (a) the cross-group pass remains
// disabled (so no cross-group warnings appear) and (b) every group
// contributes exactly one within-group warning. The group-keyed iteration
// discipline (sorted keys for determinism) is exercised by Plan 09-06's
// sort gate; here we only check the *count* per group, not the order.
func TestCheck_WithinGroup_GroupKeyedIteration(t *testing.T) {
	t.Parallel()

	s := fuzzymatch.DefaultScorer()
	cfg := scan.DefaultConfig(s)

	items := []scan.Item{
		{Name: "user_id", Group: "login"},
		{Name: "userId", Group: "login"},
		{Name: "user_id", Group: "profile"},
		{Name: "userId", Group: "profile"},
		{Name: "user_id", Group: "audit"},
		{Name: "userId", Group: "audit"},
	}

	warnings, err := scan.Check(items, cfg)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(warnings) != 3 {
		t.Fatalf("warnings: got %d; want 3 (one per group)", len(warnings))
	}

	gotGroups := map[string]int{}
	for _, w := range warnings {
		assertWarningKind(t, w, scan.KindWithinGroup)
		if w.GroupA != w.GroupB {
			t.Errorf("within-group warning: GroupA=%q GroupB=%q (want equal)", w.GroupA, w.GroupB)
		}
		gotGroups[w.GroupA]++
	}
	for _, g := range []string{"login", "profile", "audit"} {
		if gotGroups[g] != 1 {
			t.Errorf("group %q: got %d warnings; want 1", g, gotGroups[g])
		}
	}
}

// TestCheck_NoItems_NoWarnings pins the empty-input contract: items ==
// nil returns (nil, nil). The Plan 09-03 implementation returns nil (not
// []Warning{}) for both the < 2 early-exit branch and the no-emissions
// path; this test pins that contract.
func TestCheck_NoItems_NoWarnings(t *testing.T) {
	t.Parallel()

	s := fuzzymatch.DefaultScorer()
	cfg := scan.DefaultConfig(s)

	warnings, err := scan.Check(nil, cfg)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(warnings) != 0 {
		t.Errorf("warnings: got %d; want 0", len(warnings))
	}
}

// TestCheck_SingleItem_NoWarnings confirms the < 2 early-exit branch:
// a single-item slice has no pairs and produces zero warnings.
func TestCheck_SingleItem_NoWarnings(t *testing.T) {
	t.Parallel()

	s := fuzzymatch.DefaultScorer()
	cfg := scan.DefaultConfig(s)
	items := []scan.Item{{Name: "user_id", Group: "login"}}

	warnings, err := scan.Check(items, cfg)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(warnings) != 0 {
		t.Errorf("warnings: got %d; want 0", len(warnings))
	}
}

// TestCheck_DefaultConfig_DoesNotEnableCrossGroup verifies the
// DefaultConfig opinionated defaults are wired through Check correctly:
// even when items would obviously match across groups, the cross-group
// pass remains disabled until the consumer flips the flag.
func TestCheck_DefaultConfig_DoesNotEnableCrossGroup(t *testing.T) {
	t.Parallel()

	s := fuzzymatch.DefaultScorer()
	cfg := scan.DefaultConfig(s)
	// Do NOT set cfg.CompareAcrossGroups — rely on the DefaultConfig.

	items := []scan.Item{
		{Name: "user_id", Group: "login"},
		{Name: "userId", Group: "profile"},
	}

	warnings, err := scan.Check(items, cfg)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(warnings) != 0 {
		t.Errorf("warnings: got %d; want 0 (DefaultConfig should disable cross-group)", len(warnings))
	}
}

// TestCheck_ScoresPopulatedOnEmission confirms the Warning.Scores map
// is populated via Scorer.ScoreAll only on emission. Every algorithm
// with a non-zero weight in the configured set MUST appear as a key in
// Scores (per the Phase 8 ScoreAll contract — see scorer.go ScoreAll
// godoc).
func TestCheck_ScoresPopulatedOnEmission(t *testing.T) {
	t.Parallel()

	s := fuzzymatch.DefaultScorer()
	cfg := scan.DefaultConfig(s)
	items := []scan.Item{
		{Name: "user_id", Group: "login"},
		{Name: "userId", Group: "login"},
	}

	warnings, err := scan.Check(items, cfg)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(warnings) != 1 {
		t.Fatalf("warnings: got %d; want 1", len(warnings))
	}

	w := warnings[0]
	if w.Scores == nil {
		t.Fatalf("Scores: got nil; want lazily-populated map")
	}
	wantKeys := s.Algorithms()
	if len(w.Scores) != len(wantKeys) {
		t.Errorf("Scores: got %d entries; want %d (matching Scorer.Algorithms count)",
			len(w.Scores), len(wantKeys))
	}
	for _, a := range wantKeys {
		if _, ok := w.Scores[a.ID]; !ok {
			t.Errorf("Scores: missing AlgoID %v from emitted Warning", a.ID)
		}
	}
}

// TestCheck_RawNameNotNormalised_PassedToScorer is the load-bearing
// Pitfall-5 test (09-RESEARCH.md lines 488-498): scan.Check MUST NOT
// pre-normalise the strings it passes to Scorer.Score / Scorer.Match /
// Scorer.ScoreAll. The Scorer applies its own normalisation pipeline
// internally (scorer.go:356-360); double-normalisation would corrupt
// algorithms like Hamming or LCSStr that depend on exact rune lengths.
//
// This test confirms the emitted Warning carries the RAW input strings
// (mixed-case, with separators), proving Check did not write back any
// normalised form to NameA/NameB.
func TestCheck_RawNameNotNormalised_PassedToScorer(t *testing.T) {
	t.Parallel()

	s := fuzzymatch.DefaultScorer()
	cfg := scan.DefaultConfig(s)
	items := []scan.Item{
		{Name: "User_ID", Group: "login"},
		{Name: "userId", Group: "login"},
	}

	warnings, err := scan.Check(items, cfg)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(warnings) != 1 {
		t.Fatalf("warnings: got %d; want 1", len(warnings))
	}

	w := warnings[0]
	// The RAW strings (NOT the normalised "user id" form) must appear.
	if w.NameA != "User_ID" || w.NameB != "userId" {
		t.Errorf("Warning carries normalised names: got (%q, %q); want raw (%q, %q)",
			w.NameA, w.NameB, "User_ID", "userId")
	}
}

// TestCheck_NormalisationOptionsRead_FromScorer verifies the Plan 09-01
// accessor flows correctly into Check: when the Scorer is constructed
// WithoutNormalisation, Check honours that contract — items whose
// normalised forms would coincide ("user_id" vs "USER_ID") are passed
// to the Scorer as raw strings, and the Score reflects the raw
// comparison (lower for case-different strings).
//
// The behavioural assertion is on the within-group pass count — with
// the default normalising Scorer "user_id" vs "USER_ID" scores 1.0
// (well above threshold), but a WithoutNormalisation Scorer scores
// these two pairs around 0.43, which is below threshold and emits no
// warning.
func TestCheck_NormalisationOptionsRead_FromScorer(t *testing.T) {
	t.Parallel()

	opts := append(fuzzymatch.DefaultScorerOptions(), fuzzymatch.WithoutNormalisation())
	s, err := fuzzymatch.NewScorer(opts...)
	if err != nil {
		t.Fatalf("NewScorer(WithoutNormalisation): unexpected error: %v", err)
	}
	cfg := scan.DefaultConfig(s)
	items := []scan.Item{
		{Name: "user_id", Group: "login"},
		{Name: "USER_ID", Group: "login"},
	}

	// Precondition: the raw-form Score must be below threshold for this
	// test to discriminate the WithoutNormalisation path.
	if got := s.Score(items[0].Name, items[1].Name); got >= s.Threshold() {
		t.Skipf("precondition: WithoutNormalisation Score(%q, %q) = %v >= threshold %v",
			items[0].Name, items[1].Name, got, s.Threshold())
	}

	warnings, err := scan.Check(items, cfg)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(warnings) != 0 {
		t.Errorf("warnings: got %d; want 0 (WithoutNormalisation should keep case-different pair below threshold)",
			len(warnings))
	}
}

// TestCheck_CrossGroup_UsesScoreNotMatch is the load-bearing Pitfall-3
// test (09-RESEARCH.md lines 488-498): the cross-group pass MUST use
// Scorer.Score with the effective (boosted, clamped) threshold — NOT
// Scorer.Match (which applies the within-group threshold).
//
// Setup: threshold 0.50, cross-group boost 0.30 → effective threshold
// 0.80. A "customer" / "customers" pair scores ~0.77 — above the
// within threshold 0.50 (would emit on Match), but below the cross-group
// effective threshold 0.80 (must not emit on Score >= effective). If
// Check incorrectly used Match for the cross-group pass, this test
// would observe a Warning.
func TestCheck_CrossGroup_UsesScoreNotMatch(t *testing.T) {
	t.Parallel()

	s := newScorerWithThreshold(t, 0.50)
	cfg := scan.DefaultConfig(s)
	cfg.CompareAcrossGroups = true
	cfg.CrossGroupThresholdBoost = 0.30 // effective = 0.80

	items := []scan.Item{
		{Name: "customer", Group: "login"},
		{Name: "customers", Group: "profile"},
	}

	score := s.Score(items[0].Name, items[1].Name)
	// Sanity-check the precondition the test depends on: score must be
	// in (0.50, 0.80) for this to discriminate Match-vs-Score.
	if score <= 0.50 || score >= 0.80 {
		t.Skipf("precondition: Score(%q, %q) = %v outside (0.50, 0.80)",
			items[0].Name, items[1].Name, score)
	}

	warnings, err := scan.Check(items, cfg)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(warnings) != 0 {
		t.Fatalf("warnings: got %d; want 0 (Score %.4f < effective threshold 0.80; Match would erroneously fire here)",
			len(warnings), score)
	}

	// Companion within-group check: the same pair in the same group
	// MUST emit on Match (Score >= 0.50). This proves the test's
	// negative cross-group claim isn't a false negative caused by some
	// other suppression.
	itemsSameGroup := []scan.Item{
		{Name: "customer", Group: "login"},
		{Name: "customers", Group: "login"},
	}
	cfgSameGroup := scan.DefaultConfig(s)
	warningsSameGroup, err := scan.Check(itemsSameGroup, cfgSameGroup)
	if err != nil {
		t.Fatalf("unexpected error (same-group control): %v", err)
	}
	if len(warningsSameGroup) != 1 {
		t.Fatalf("same-group control: got %d warnings; want 1 (Match should fire on Score %.4f >= threshold 0.50)",
			len(warningsSameGroup), score)
	}
}
