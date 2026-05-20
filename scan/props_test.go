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

// props_test.go declares the property tests that gate Plan 09-04's
// token-bucket optimisation:
//
//   - PropCheck_BucketEquivalentToNaive (SCAN-02 load-bearing):
//     for any randomly-generated []Item where at least one group's
//     size exceeds bucketThreshold, scan.Check with the production
//     dispatch (bucket path active) produces a warning set identical
//     to scan.Check with the forced naive path (toggled via the
//     test-only forceNaivePath hook exposed by export_test.go).
//
//   - PropCheck_DeterministicAcrossRuns: two consecutive Check
//     invocations on the same []Item produce identical warning sets
//     (after the canonical Plan-09-06 sort, applied locally here).
//
//   - PropCheck_NoSelfWarnings: no warning has (NameA == NameB AND
//     GroupA == GroupB AND TagA == TagB) — i.e. Check never emits an
//     (i, i) pair (Pitfall 8).
//
// All three properties use testing/quick with a custom generator over
// []scan.Item; the generator produces inputs that exercise both the
// naive (small groups) and bucket (large groups) dispatch paths, and
// guarantees no duplicate (Name, Group) pairs (D-06) so the
// validateCheck gate accepts every generated input.
//
// Stdlib testing + testing/quick only — no third-party assertion or
// property frameworks in root tests per CLAUDE.md.

package scan_test

import (
	"fmt"
	"math"
	"math/rand"
	"reflect"
	"sort"
	"testing"
	"testing/quick"

	"github.com/axonops/fuzzymatch"
	"github.com/axonops/fuzzymatch/scan"
)

// ----------------------------------------------------------------------
// Generator: random []scan.Item with controlled group-size distribution.
// ----------------------------------------------------------------------

// itemSliceGen is the testing/quick generator that produces a random
// []scan.Item. The generator:
//
//   - Produces 50–250 items.
//   - Partitions them across 3–8 distinct group names drawn from a
//     fixed alphabet (so the result is reproducible per quick.Check
//     seed; group names themselves do not need to be diverse — we only
//     care that group SIZES vary).
//   - At least one group is forced to exceed bucketThreshold so the
//     bucket dispatch path is exercised on every generated input.
//   - Item names are drawn from a base bag of identifier-style names
//     with small mutations (suffix digits, case flips, underscore-vs-
//     dash) so a non-trivial fraction of pairs crosses the 0.85
//     similarity threshold.
//   - Guarantees no duplicate (Name, Group) pairs (D-06 acceptance):
//     each generated Name carries a unique disambiguating suffix per
//     group so validateCheck never rejects the input.
//   - Always sets SilenceLint == false and Tag == nil (per-item
//     suppression is Plan 09-05's scope; the bucket property test
//     focuses on the bucket vs naive equivalence in isolation).
type itemSliceGen []scan.Item

// Generate satisfies quick.Generator. rng is testing/quick's seeded
// random source; size is testing/quick's size hint, ignored here in
// favour of the generator's own bounded range to keep per-Check
// duration manageable (each call runs Check twice, once per path).
func (itemSliceGen) Generate(rng *rand.Rand, _ int) reflect.Value {
	// fatGroupTarget tracks the current bucketThreshold via the
	// test-only accessor so the generator never silently stops
	// exercising the bucket dispatch path if bucketThreshold is
	// raised in a future empirical-validation pass (D-08). Closes
	// the algorithm-correctness-reviewer MEDIUM finding on Plan
	// 09-04 (durability concern: previously hard-coded at 75
	// against a 50 hypothesis).
	fatGroupTarget := scan.BucketThreshold() + 25

	// Ensure n always exceeds fatGroupTarget by at least one item so
	// remaining groups receive at least one item each (n in
	// [fatGroupTarget+8, fatGroupTarget+208]).
	n := fatGroupTarget + 8 + rng.Intn(201)

	// 3–8 groups.
	groupCount := 3 + rng.Intn(6)
	groupNames := make([]string, groupCount)
	for i := 0; i < groupCount; i++ {
		groupNames[i] = fmt.Sprintf("group_%d", i)
	}

	// Assign each item to a group such that group 0 is "fat" (exceeds
	// bucketThreshold to exercise the bucket dispatch) and the rest
	// share the remaining items roughly evenly. The exact split does
	// not matter for the property — only that at least one group has
	// size > bucketThreshold.
	groupAssign := make([]int, n)
	for i := 0; i < n; i++ {
		if i < fatGroupTarget {
			groupAssign[i] = 0
		} else {
			groupAssign[i] = 1 + rng.Intn(groupCount-1)
		}
	}

	// Shuffle the group assignment so the items aren't laid out in
	// group-blocks (otherwise the slice-index order trivially matches
	// the group iteration order — we want some real interleaving to
	// catch order-sensitive bugs).
	rng.Shuffle(len(groupAssign), func(i, j int) {
		groupAssign[i], groupAssign[j] = groupAssign[j], groupAssign[i]
	})

	// Identifier-style base names. Mutations append a per-(group, idx)
	// suffix so (Name, Group) is unique within a group.
	bases := []string{
		"user_id", "userId", "user_name", "userName",
		"customer_id", "customerId", "account_id", "accountId",
		"order_total", "orderTotal", "email_address", "emailAddress",
	}

	items := make([]scan.Item, n)
	perGroupCounter := make(map[int]int, groupCount)
	for i := 0; i < n; i++ {
		g := groupAssign[i]
		base := bases[rng.Intn(len(bases))]
		// Disambiguator: append "_<groupIdx>_<perGroupIdx>" so every
		// (Name, Group) pair is unique. Group-local counters keep the
		// suffixes small and predictable.
		counter := perGroupCounter[g]
		perGroupCounter[g] = counter + 1
		name := fmt.Sprintf("%s_%d_%d", base, g, counter)
		items[i] = scan.Item{
			Name:  name,
			Group: groupNames[g],
			// Randomly populate SilenceLint on a ~10% subset so the
			// bucket-vs-naive equivalence property exercises Rule 1
			// (per-item SilenceLint) suppression. Closes the
			// algorithm-correctness-reviewer HIGH finding INV-6 on
			// Plan 09-05 — SCAN-02 equivalence under SUPPRESSED
			// emission was previously untested.
			SilenceLint: rng.Intn(10) == 0,
			Tag:         nil,
		}
	}
	return reflect.ValueOf(itemSliceGen(items))
}

// generateSuppressedPairs builds a random [][2]string suitable for
// cfg.SuppressedPairs from the items already generated. About 30% of
// entries are drawn from real item names (so suppression fires on
// real candidate pairs); 70% are arbitrary non-overlapping names
// (exercising the map's negative-lookup path). Returned slice has
// 0–10 entries.
//
// Closes the algorithm-correctness-reviewer HIGH finding INV-6 on
// Plan 09-05 — SCAN-02 equivalence under SuppressedPairs
// suppression was previously untested.
func generateSuppressedPairs(items []scan.Item, rng *rand.Rand) [][2]string {
	n := rng.Intn(11) // 0–10 entries
	if n == 0 || len(items) < 2 {
		return nil
	}
	pairs := make([][2]string, 0, n)
	for i := 0; i < n; i++ {
		if rng.Intn(10) < 3 {
			// 30% chance: draw two real item names.
			a := items[rng.Intn(len(items))].Name
			b := items[rng.Intn(len(items))].Name
			pairs = append(pairs, [2]string{a, b})
		} else {
			// 70% chance: arbitrary names unlikely to match.
			pairs = append(pairs, [2]string{
				fmt.Sprintf("noop_a_%d", i),
				fmt.Sprintf("noop_b_%d", i),
			})
		}
	}
	return pairs
}

// ----------------------------------------------------------------------
// PropCheck_BucketEquivalentToNaive — SCAN-02 load-bearing
// ----------------------------------------------------------------------

// PropCheck_BucketEquivalentToNaive proves that for every randomly
// generated []Item with at least one group exceeding bucketThreshold,
// scan.Check's bucket path produces a warning set identical to the
// forced-naive path. This is the SCAN-02 load-bearing gate: any future
// refactor that breaks bucket-vs-naive equivalence fails this property
// before it lands in production.
//
// The test exposes the test-only forceNaivePath hook via
// scan.ForceNaivePath (declared in scan/export_test.go) to flip the
// dispatch path. The hook is global package state so the test ensures
// it restores forceNaivePath = false on exit.
//
// MaxCount: 50 — each invocation runs Check twice on a ~50–250-item
// input, so the wall-clock budget per property invocation is generous
// without overshooting the typical `go test` runtime.
//
// NOT t.Parallel(): this test mutates the package-private
// forceNaivePath atomic flag. Even though atomic.Bool prevents data
// races, a concurrent test that observes the flipped flag would see
// inconsistent dispatch behaviour. Serial execution keeps the toggle
// visible only to this test's goroutine.
func TestPropCheck_BucketEquivalentToNaive(t *testing.T) {
	// Restore the hook on exit so a panicking property doesn't leak
	// forceNaivePath = true into a subsequent test.
	defer scan.SetForceNaivePath(false)

	// Deterministic per-property RNG seed for SuppressedPairs +
	// CompareIdenticalAcrossGroups generation. quick.Check passes its
	// own seed via the generator's rng, but the cfg-shape randomisation
	// is independent of input generation, so we use a separate seed
	// here. Stable across runs for reproducibility.
	cfgRng := rand.New(rand.NewSource(0xCF13C0FF13))

	prop := func(gen itemSliceGen) bool {
		items := []scan.Item(gen)
		cfg := scan.DefaultConfig(fuzzymatch.DefaultScorer())
		cfg.CompareAcrossGroups = true // exercise both passes
		// Randomise CompareIdenticalAcrossGroups so Rule 3
		// (cross-group identical-name) is exercised both enabled and
		// disabled. Closes algorithm-correctness INV-6 partial.
		cfg.CompareIdenticalAcrossGroups = cfgRng.Intn(2) == 0
		// Randomise SuppressedPairs so Rule 2 (canonical-pair
		// suppression) is exercised on both naive and bucket paths.
		cfg.SuppressedPairs = generateSuppressedPairs(items, cfgRng)

		// Production (bucket) path.
		scan.SetForceNaivePath(false)
		prodWarnings, err := scan.Check(items, cfg)
		if err != nil {
			t.Logf("Check (bucket path) error: %v", err)
			return false
		}

		// Forced-naive path.
		scan.SetForceNaivePath(true)
		naiveWarnings, err := scan.Check(items, cfg)
		if err != nil {
			t.Logf("Check (forced naive) error: %v", err)
			return false
		}

		// Reset for the next iteration.
		scan.SetForceNaivePath(false)

		if !warningSetsEqual(prodWarnings, naiveWarnings) {
			t.Logf("bucket vs naive divergence on %d items / %d groups (CompareIdenticalAcrossGroups=%v, SuppressedPairs=%d): bucket=%d, naive=%d",
				len(items), distinctGroupCount(items),
				cfg.CompareIdenticalAcrossGroups, len(cfg.SuppressedPairs),
				len(prodWarnings), len(naiveWarnings))
			return false
		}
		return true
	}

	if err := quick.Check(prop, &quick.Config{MaxCount: 50}); err != nil {
		t.Errorf("PropCheck_BucketEquivalentToNaive: %v", err)
	}
}

// ----------------------------------------------------------------------
// PropCheck_DeterministicAcrossRuns
// ----------------------------------------------------------------------

// PropCheck_DeterministicAcrossRuns proves that two consecutive
// scan.Check invocations on the same []Item produce identical (after
// canonical sort) warning slices. This exercises the production code
// path — bucket dispatch active where the group exceeds threshold,
// naive otherwise.
func TestPropCheck_DeterministicAcrossRuns(t *testing.T) {
	t.Parallel()

	prop := func(gen itemSliceGen) bool {
		items := []scan.Item(gen)
		cfg := scan.DefaultConfig(fuzzymatch.DefaultScorer())
		cfg.CompareAcrossGroups = true

		w1, err := scan.Check(items, cfg)
		if err != nil {
			t.Logf("Check (run 1) error: %v", err)
			return false
		}
		w2, err := scan.Check(items, cfg)
		if err != nil {
			t.Logf("Check (run 2) error: %v", err)
			return false
		}
		return warningSetsEqual(w1, w2)
	}

	if err := quick.Check(prop, &quick.Config{MaxCount: 100}); err != nil {
		t.Errorf("PropCheck_DeterministicAcrossRuns: %v", err)
	}
}

// ----------------------------------------------------------------------
// PropCheck_NoSelfWarnings
// ----------------------------------------------------------------------

// PropCheck_NoSelfWarnings proves that scan.Check never emits a
// Warning whose (NameA, GroupA) equals (NameB, GroupB) — i.e. no
// (i, i) self-pairs (Pitfall 8). The validateCheck gate rejects
// duplicate (Name, Group) tuples (D-06), so a self-warning would
// only be possible via a bug in the bucket candidate enumeration.
func TestPropCheck_NoSelfWarnings(t *testing.T) {
	t.Parallel()

	prop := func(gen itemSliceGen) bool {
		items := []scan.Item(gen)
		cfg := scan.DefaultConfig(fuzzymatch.DefaultScorer())
		cfg.CompareAcrossGroups = true

		warnings, err := scan.Check(items, cfg)
		if err != nil {
			t.Logf("Check error: %v", err)
			return false
		}
		for i, w := range warnings {
			if w.NameA == w.NameB && w.GroupA == w.GroupB {
				t.Logf("self-warning at index %d: %+v", i, w)
				return false
			}
		}
		return true
	}

	if err := quick.Check(prop, &quick.Config{MaxCount: 100}); err != nil {
		t.Errorf("PropCheck_NoSelfWarnings: %v", err)
	}
}

// ----------------------------------------------------------------------
// PropCheck_NoNaN — DET-04 surface
// ----------------------------------------------------------------------

// PropCheck_NoNaN proves that scan.Check never emits a Warning whose
// composite Score is NaN nor whose per-algorithm Scores entries contain
// NaN. Closes DET-04 (NaN/Inf/-0 handling): scan does no float
// arithmetic of its own — every Score / Scores value flows through
// Scorer.Score / Scorer.ScoreAll, which Phase 8.5 hardened against NaN
// drift via the FMA-defeating double-cast (scorer.go float-determinism
// gate). This property test pins the contract on the scan boundary so
// any future regression in the Scorer or the algorithm catalogue is
// caught at the scan-output surface.
func TestPropCheck_NoNaN(t *testing.T) {
	t.Parallel()

	prop := func(gen itemSliceGen) bool {
		items := []scan.Item(gen)
		cfg := scan.DefaultConfig(fuzzymatch.DefaultScorer())
		cfg.CompareAcrossGroups = true

		warnings, err := scan.Check(items, cfg)
		if err != nil {
			t.Logf("Check error: %v", err)
			return false
		}
		for i, w := range warnings {
			if math.IsNaN(w.Score) {
				t.Logf("Warning[%d].Score is NaN: %+v", i, w)
				return false
			}
			for id, v := range w.Scores {
				if math.IsNaN(v) {
					t.Logf("Warning[%d].Scores[%v] is NaN: %+v", i, id, w)
					return false
				}
			}
		}
		return true
	}

	if err := quick.Check(prop, &quick.Config{MaxCount: 100}); err != nil {
		t.Errorf("PropCheck_NoNaN: %v", err)
	}
}

// ----------------------------------------------------------------------
// PropCheck_NoInf — DET-04 surface
// ----------------------------------------------------------------------

// PropCheck_NoInf proves that scan.Check never emits a Warning whose
// composite Score is +Inf or -Inf nor whose per-algorithm Scores
// entries contain ±Inf. Same DET-04 contract as PropCheck_NoNaN:
// composite + per-algorithm scores are in [0.0, 1.0] under
// DefaultScorer; any infinity would indicate a Scorer regression that
// scan must surface immediately.
func TestPropCheck_NoInf(t *testing.T) {
	t.Parallel()

	prop := func(gen itemSliceGen) bool {
		items := []scan.Item(gen)
		cfg := scan.DefaultConfig(fuzzymatch.DefaultScorer())
		cfg.CompareAcrossGroups = true

		warnings, err := scan.Check(items, cfg)
		if err != nil {
			t.Logf("Check error: %v", err)
			return false
		}
		for i, w := range warnings {
			if math.IsInf(w.Score, 0) {
				t.Logf("Warning[%d].Score is Inf: %+v", i, w)
				return false
			}
			for id, v := range w.Scores {
				if math.IsInf(v, 0) {
					t.Logf("Warning[%d].Scores[%v] is Inf: %+v", i, id, w)
					return false
				}
			}
		}
		return true
	}

	if err := quick.Check(prop, &quick.Config{MaxCount: 100}); err != nil {
		t.Errorf("PropCheck_NoInf: %v", err)
	}
}

// ----------------------------------------------------------------------
// Helpers
// ----------------------------------------------------------------------

// warningSetsEqual compares two []scan.Warning slices for equality as
// sets (order-independent). Both slices are sorted in place via the
// canonical (Kind, NameA, NameB, GroupA, GroupB) key and then
// element-wise compared on the core fields (Kind / Names / Groups /
// Tags / Score / Scores).
//
// Score equality is exact (no tolerance) — the Scorer is deterministic
// across calls per Phase 8's invariants.
func warningSetsEqual(a, b []scan.Warning) bool {
	if len(a) != len(b) {
		return false
	}
	aSorted := append([]scan.Warning(nil), a...)
	bSorted := append([]scan.Warning(nil), b...)
	sortWarnings(aSorted)
	sortWarnings(bSorted)
	for i := range aSorted {
		if !warningEqualCore(aSorted[i], bSorted[i]) {
			return false
		}
	}
	return true
}

// sortWarnings applies the Plan 09-06 canonical sort key to ws in
// place. Local helper because Check itself does not yet sort (Plan
// 09-06 lands the production sort).
func sortWarnings(ws []scan.Warning) {
	sort.SliceStable(ws, func(i, j int) bool {
		if ws[i].Kind != ws[j].Kind {
			return ws[i].Kind < ws[j].Kind
		}
		if ws[i].NameA != ws[j].NameA {
			return ws[i].NameA < ws[j].NameA
		}
		if ws[i].NameB != ws[j].NameB {
			return ws[i].NameB < ws[j].NameB
		}
		if ws[i].GroupA != ws[j].GroupA {
			return ws[i].GroupA < ws[j].GroupA
		}
		return ws[i].GroupB < ws[j].GroupB
	})
}

// warningEqualCore compares two Warning values on the fields that
// bucket-vs-naive equivalence must preserve. Scores map equality is
// exact (Go map deep-equal); Tag is compared via reflect.DeepEqual so
// any consumer-supplied type works (Tag is typed any in the public
// API).
func warningEqualCore(a, b scan.Warning) bool {
	if a.Kind != b.Kind || a.NameA != b.NameA || a.NameB != b.NameB ||
		a.GroupA != b.GroupA || a.GroupB != b.GroupB {
		return false
	}
	if !floatExactlyEqual(a.Score, b.Score) {
		return false
	}
	if !reflect.DeepEqual(a.TagA, b.TagA) || !reflect.DeepEqual(a.TagB, b.TagB) {
		return false
	}
	if !reflect.DeepEqual(a.Scores, b.Scores) {
		return false
	}
	return true
}

// floatExactlyEqual handles the NaN-NaN case (NaN != NaN in IEEE-754,
// but for our purposes two NaN scores should compare equal). We do not
// expect NaN in practice — Scorer.Score under DefaultScorer is finite
// — but the helper documents the intent.
func floatExactlyEqual(a, b float64) bool {
	if math.IsNaN(a) && math.IsNaN(b) {
		return true
	}
	return a == b
}

// distinctGroupCount returns the number of distinct Group values in
// items — diagnostic helper for property-test failure logging.
func distinctGroupCount(items []scan.Item) int {
	seen := make(map[string]struct{}, len(items))
	for _, it := range items {
		seen[it.Group] = struct{}{}
	}
	return len(seen)
}
