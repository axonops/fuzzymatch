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

// scan_internal_test.go covers the unexported helpers that back the
// Plan 09-04 token-bucket optimisation. Black-box tests in
// scan_test.go cover the public Check surface; these tests pin the
// internal helpers' contracts directly so a refactor that breaks one
// of the optimisation primitives fails before the bucket-equivalence
// property test does.
//
// Covered helpers:
//
//   - bucketThreshold (private const): pinned to a non-zero positive
//     int so a refactor that accidentally zeroes the constant or
//     flips its sign is caught immediately. The exact value is
//     empirically validated by BenchmarkScanCheck_BucketVsNaive_GroupSize
//     (Task 3 + manual checkpoint Task 4).
//
//   - tokeniseAll: one tokens slice per input Item, Normalise applied
//     before Tokenise (per Open Question 2 resolution), determinism
//     across consecutive calls.
//
//   - buildBucket + bucketCandidates: bucket build dedupes repeated
//     tokens per item; candidate enumeration excludes the source index
//     and returns a sorted slice (deterministic output per
//     .claude/skills/determinism-standards).
//
// Stdlib testing only — no testify in root tests per CLAUDE.md.

package scan

import (
	"errors"
	"fmt"
	"reflect"
	"sort"
	"strings"
	"testing"

	"github.com/axonops/fuzzymatch"
)

// Test_bucketThreshold_PositiveNonZero pins the empirically-validated
// private constant to a sensible range. The starting hypothesis is 50
// (per docs/requirements.md §12.5 and 09-CONTEXT.md §5 D-08); the
// constant may be revised by Task 4's manual checkpoint based on the
// BenchmarkScanCheck_BucketVsNaive_GroupSize sweep. Either way the
// value must be a positive int greater than 1 (single-item groups
// cannot meaningfully bucket) and less than the empirical sweep
// ceiling of 1000 (any value above that is a misconfiguration —
// scan.Check should never refuse to bucket in practice).
func Test_bucketThreshold_PositiveNonZero(t *testing.T) {
	t.Parallel()
	if bucketThreshold <= 1 {
		t.Fatalf("bucketThreshold = %d; want > 1 (zero/one would disable bucketing or never trigger)", bucketThreshold)
	}
	if bucketThreshold >= 1000 {
		t.Fatalf("bucketThreshold = %d; want < 1000 (above the empirical sweep ceiling)", bucketThreshold)
	}
}

// Test_tokeniseAll_OncePerItem confirms tokeniseAll returns exactly
// one tokens slice per input Item, in input order. The token content
// matches Tokenise(Normalise(name, normOpts), DefaultTokeniseOptions())
// when applyNormalisation == true, and Tokenise(name, ...) when false.
// Pitfall 7: tokenisation runs at most once per Item — the helper is
// the documented place that promise is implemented.
func Test_tokeniseAll_OncePerItem(t *testing.T) {
	t.Parallel()

	items := []Item{
		{Name: "user_id", Group: "g1"},
		{Name: "userId", Group: "g1"},
		{Name: "user_name", Group: "g2"},
	}

	// Pull the production normalisation options from a DefaultScorer
	// so the test exercises the same pipeline scan.Check uses.
	s := fuzzymatch.DefaultScorer()
	normOpts, applied := s.NormalisationOptions()

	tokens := tokeniseAll(items, normOpts, applied)
	if len(tokens) != len(items) {
		t.Fatalf("len(tokens) = %d; want %d", len(tokens), len(items))
	}

	// Re-derive each item's tokens via the public pipeline and assert
	// equality. This is the structural invariant — tokeniseAll is a
	// pure composition of Normalise + Tokenise.
	for i, item := range items {
		var want []string
		if applied {
			want = fuzzymatch.Tokenise(
				fuzzymatch.Normalise(item.Name, normOpts),
				fuzzymatch.DefaultTokeniseOptions(),
			)
		} else {
			want = fuzzymatch.Tokenise(item.Name, fuzzymatch.DefaultTokeniseOptions())
		}
		if !reflect.DeepEqual(tokens[i], want) {
			t.Errorf("tokens[%d] = %v; want %v", i, tokens[i], want)
		}
	}
}

// Test_tokeniseAll_DeterministicAcrossRuns pins the determinism of the
// helper — two consecutive calls on the same input must produce
// identical tokens. This is a precondition for the bucket dispatch
// being deterministic (PropCheck_DeterministicAcrossRuns) under the
// production code path.
func Test_tokeniseAll_DeterministicAcrossRuns(t *testing.T) {
	t.Parallel()

	items := []Item{
		{Name: "user_id", Group: "g1"},
		{Name: "userId", Group: "g1"},
		{Name: "user-name", Group: "g2"},
		{Name: "user.id", Group: "g2"},
	}
	s := fuzzymatch.DefaultScorer()
	normOpts, applied := s.NormalisationOptions()

	a := tokeniseAll(items, normOpts, applied)
	b := tokeniseAll(items, normOpts, applied)
	if !reflect.DeepEqual(a, b) {
		t.Errorf("tokeniseAll non-deterministic:\n  call 1: %v\n  call 2: %v", a, b)
	}
}

// Test_buildBucket_DedupesRepeatedTokens guards against the obvious
// bug of pushing the same idx into bucket[token] more than once when
// an item's token list contains repeated tokens (e.g. "id_id" tokenises
// to ["id", "id"]). The bucket value slice should contain idx at most
// once per token key.
func Test_buildBucket_DedupesRepeatedTokens(t *testing.T) {
	t.Parallel()

	// Synthetic tokens with deliberate repetition.
	tokens := [][]string{
		{"id", "id", "user"}, // idx 0
		{"user", "name"},     // idx 1
		{"id"},               // idx 2
	}
	indices := []int{0, 1, 2}

	bucket := buildBucket(indices, tokens)

	// Helper to check whether s is sorted and unique.
	mustSortedUnique := func(t *testing.T, key string, slice []int) {
		t.Helper()
		if !sort.IntsAreSorted(slice) {
			t.Errorf("bucket[%q] = %v; expected sorted", key, slice)
		}
		for i := 1; i < len(slice); i++ {
			if slice[i] == slice[i-1] {
				t.Errorf("bucket[%q] = %v; duplicate entry %d", key, slice, slice[i])
			}
		}
	}

	// bucket["id"] should be {0, 2} — idx 0 appears once even though its
	// tokens list contains "id" twice; idx 2 contains "id" once.
	got := bucket["id"]
	mustSortedUnique(t, "id", got)
	if !reflect.DeepEqual(got, []int{0, 2}) {
		t.Errorf("bucket[\"id\"] = %v; want [0 2]", got)
	}

	// bucket["user"] should be {0, 1}.
	got = bucket["user"]
	mustSortedUnique(t, "user", got)
	if !reflect.DeepEqual(got, []int{0, 1}) {
		t.Errorf("bucket[\"user\"] = %v; want [0 1]", got)
	}

	// bucket["name"] should be {1}.
	got = bucket["name"]
	if !reflect.DeepEqual(got, []int{1}) {
		t.Errorf("bucket[\"name\"] = %v; want [1]", got)
	}
}

// Test_bucketCandidates_ExcludesSelfAndSorts verifies that
// bucketCandidates returns a sorted slice that does not contain the
// source index itself (Pitfall 8 — no self-pairs) and that the
// candidate set is the union of bucket[token] for each token in
// tokens[srcIdx], minus srcIdx itself.
func Test_bucketCandidates_ExcludesSelfAndSorts(t *testing.T) {
	t.Parallel()

	tokens := [][]string{
		{"user", "id"},   // idx 0
		{"user", "name"}, // idx 1
		{"id", "value"},  // idx 2
		{"unrelated"},    // idx 3
	}
	indices := []int{0, 1, 2, 3}
	bucket := buildBucket(indices, tokens)

	// idx 0's tokens: "user" → {0, 1}; "id" → {0, 2}. Candidates =
	// {0, 1, 2} \ {0} = {1, 2}.
	got := bucketCandidates(0, bucket, tokens)
	if !sort.IntsAreSorted(got) {
		t.Errorf("bucketCandidates(0): got %v; expected sorted ascending", got)
	}
	want := []int{1, 2}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("bucketCandidates(0): got %v; want %v", got, want)
	}

	// idx 3's tokens: "unrelated" → {3}. Candidates = {3} \ {3} = {}.
	got = bucketCandidates(3, bucket, tokens)
	if len(got) != 0 {
		t.Errorf("bucketCandidates(3): got %v; want empty (no shared tokens)", got)
	}
}

// Test_forceNaivePath_TogglesPathConsistently uses the test-only
// forceNaivePath hook to confirm the dispatch decision is honoured.
// When forceNaivePath == true the bucket path is suppressed even on a
// large group (>bucketThreshold); both paths produce identical
// warning sets (the SCAN-02 load-bearing equivalence — proved more
// rigorously by PropCheck_BucketEquivalentToNaive but smoke-tested
// here for fast feedback during development).
func Test_forceNaivePath_TogglesPathConsistently(t *testing.T) {
	// NOT t.Parallel(): this test mutates the package-private
	// forceNaivePath atomic flag. Even though atomic.Bool prevents
	// data races, a concurrent test that observes the flipped flag
	// would see inconsistent dispatch behaviour. Serial execution
	// keeps the toggle visible only to this test's goroutine.

	// Build a group large enough to exercise the bucket path under
	// production dispatch.
	const groupSize = bucketThreshold + 5
	items := make([]Item, groupSize)
	for i := 0; i < groupSize; i++ {
		// Generate moderately-similar identifier names so a few pairs
		// cross the 0.85 threshold (e.g. user_id_0 vs userId0).
		switch i % 3 {
		case 0:
			items[i] = Item{Name: fmt.Sprintf("user_id_%d", i), Group: "g"}
		case 1:
			items[i] = Item{Name: fmt.Sprintf("userId%d", i), Group: "g"}
		default:
			items[i] = Item{Name: fmt.Sprintf("user_name_%d", i), Group: "g"}
		}
	}
	cfg := DefaultConfig(fuzzymatch.DefaultScorer())

	// Run with production dispatch (bucket path active).
	forceNaivePath.Store(false)
	prodWarnings, err := Check(items, cfg)
	if err != nil {
		t.Fatalf("Check (bucket path): unexpected error: %v", err)
	}

	// Run with forced naive path.
	forceNaivePath.Store(true)
	defer forceNaivePath.Store(false)
	naiveWarnings, err := Check(items, cfg)
	if err != nil {
		t.Fatalf("Check (forced naive): unexpected error: %v", err)
	}

	if len(prodWarnings) != len(naiveWarnings) {
		t.Fatalf("bucket vs naive warning count mismatch: bucket=%d, naive=%d", len(prodWarnings), len(naiveWarnings))
	}

	// Sort both slices by the canonical sort key and compare. The
	// post-Plan-09-06 Check will apply this sort internally, but Plan
	// 09-04 returns insertion-order warnings — the test sorts locally so
	// the comparison is stable.
	sortWarningsLocal(prodWarnings)
	sortWarningsLocal(naiveWarnings)
	for i := range prodWarnings {
		if !warningCoreEqual(prodWarnings[i], naiveWarnings[i]) {
			t.Errorf("warnings[%d] differ between paths:\n  bucket: %+v\n  naive:  %+v", i, prodWarnings[i], naiveWarnings[i])
		}
	}
}

// sortWarningsLocal applies the Plan 09-06 canonical sort key (Kind,
// NameA, NameB, GroupA, GroupB) to the supplied slice in place. The
// test helper is local to this file (not exported) because Check itself
// does not yet sort — Plan 09-06 lands the production sort.
func sortWarningsLocal(ws []Warning) {
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

// Test_completenessAssertion_PanicsOnConstructedDuplicate confirms the
// in-line completeness assertion (Plan 09-06; SCAN-05 defence-in-depth)
// panics with a value wrapping fuzzymatch.ErrInternalInvariantViolated
// when two adjacent warnings share the full (Kind, NameA, NameB,
// GroupA, GroupB) sort key. The duplicate is constructed manually here
// because D-06's input validation gate makes the panic unreachable via
// the public Check path.
func Test_completenessAssertion_PanicsOnConstructedDuplicate(t *testing.T) {
	t.Parallel()

	defer func() {
		r := recover()
		if r == nil {
			t.Fatal("expected panic on duplicate sort key; got none")
		}
		err, ok := r.(error)
		if !ok {
			t.Fatalf("recovered panic value: got %T (%v); want error", r, r)
		}
		if !errors.Is(err, fuzzymatch.ErrInternalInvariantViolated) {
			t.Fatalf("recovered err: got %v; want errors.Is(err, ErrInternalInvariantViolated)", err)
		}
	}()

	// Construct a []Warning slice with two entries sharing the full
	// sort key — D-06 prevents this from happening under valid Check
	// usage, but the assertion is the documented defence-in-depth gate
	// for any future library bug that could violate the invariant.
	ws := []Warning{
		{Kind: KindWithinGroup, NameA: "a", NameB: "b", GroupA: "g", GroupB: "g"},
		{Kind: KindWithinGroup, NameA: "a", NameB: "b", GroupA: "g", GroupB: "g"},
	}
	assertSortKeyComplete(ws)
}

// Test_completenessAssertion_NoPanicOnDistinctKeys confirms the
// assertion is a no-op on a slice with strictly increasing sort keys.
// This smoke-tests the linear scan's correctness in the (overwhelmingly
// common) zero-duplicate case.
func Test_completenessAssertion_NoPanicOnDistinctKeys(t *testing.T) {
	t.Parallel()

	ws := []Warning{
		{Kind: KindWithinGroup, NameA: "a", NameB: "b", GroupA: "g", GroupB: "g"},
		{Kind: KindWithinGroup, NameA: "a", NameB: "c", GroupA: "g", GroupB: "g"},
		{Kind: KindAcrossGroups, NameA: "a", NameB: "b", GroupA: "g", GroupB: "h"},
	}
	// Must not panic.
	assertSortKeyComplete(ws)
}

// Test_completenessAssertion_PanicMessageIncludesContext checks the
// panic message embeds the duplicate-sort-key context (Kind, names,
// groups, index) so a maintainer who recovers the panic in a debug
// session can locate the offending warning without a stack walk.
// Tag values are NOT included (T-09-06-02 mitigation).
func Test_completenessAssertion_PanicMessageIncludesContext(t *testing.T) {
	t.Parallel()

	defer func() {
		r := recover()
		if r == nil {
			t.Fatal("expected panic; got none")
		}
		err, ok := r.(error)
		if !ok {
			t.Fatalf("recovered panic: got %T; want error", r)
		}
		msg := err.Error()
		// Each of these substrings must appear in the panic message so
		// the maintainer can identify the offending warning. Tag is
		// intentionally omitted (T-09-06-02).
		for _, sub := range []string{"duplicate sort key", "WithinGroup", "\"alpha\"", "\"beta\"", "\"g1\"", "\"g2\""} {
			if !strings.Contains(msg, sub) {
				t.Errorf("panic message %q missing %q", msg, sub)
			}
		}
		if strings.Contains(msg, "secret-tag-value") {
			t.Errorf("panic message %q leaks Tag value (T-09-06-02)", msg)
		}
	}()

	ws := []Warning{
		{
			Kind: KindWithinGroup, NameA: "alpha", NameB: "beta",
			GroupA: "g1", GroupB: "g2",
			TagA: "secret-tag-value", TagB: nil,
		},
		{
			Kind: KindWithinGroup, NameA: "alpha", NameB: "beta",
			GroupA: "g1", GroupB: "g2",
			TagA: nil, TagB: "secret-tag-value",
		},
	}
	assertSortKeyComplete(ws)
}

// warningCoreEqual compares two Warning values for equality on the
// fields that bucket-vs-naive dispatch must preserve. The Scores map is
// compared by content (not by reference); Tag is compared by reflect
// deep equality so any consumer-supplied type works.
func warningCoreEqual(a, b Warning) bool {
	if a.Kind != b.Kind || a.NameA != b.NameA || a.NameB != b.NameB ||
		a.GroupA != b.GroupA || a.GroupB != b.GroupB || a.Score != b.Score {
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
