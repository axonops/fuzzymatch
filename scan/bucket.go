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

// bucket.go declares the Phase 9 Plan 09-04 token-bucket optimisation
// that prunes the O(N²) candidate set inside scan.Check. The
// optimisation is a textbook inverted-index pattern: map every token
// to the slice of item indices that contain it; for each source item,
// the candidate set is the union of the buckets keyed by that source
// item's tokens (minus the source itself). Pairs that share no token
// are eliminated without paying the Scorer.Score cost.
//
// Worst-case complexity is unchanged from naive (an adversarial input
// where every item shares a single token reduces to the same nested
// loop), but expected complexity on realistic identifier-style
// workloads drops sharply because most non-matching pairs share no
// token. The PERF-05 < 2s budget for 10,000 items is met by this
// pruning combined with the Phase 8.5 Q8b Tokenise ASCII fast path.
//
// Correctness guarantee — SCAN-02:
//
// PropCheck_BucketEquivalentToNaive (scan/props_test.go) proves that
// scan.Check's bucket dispatch produces a warning set identical to the
// forced-naive dispatch for any randomly-generated []scan.Item whose
// fattest group exceeds bucketThreshold. The property test is the
// SCAN-02 load-bearing gate; a refactor that breaks bucket-vs-naive
// equivalence fails it before landing in production.
//
// No map iteration on output paths (per
// .claude/skills/determinism-standards):
//
//   - tokenBucket is read via direct lookup (bucket[token]) inside
//     bucketCandidates; the function never iterates the map.
//   - bucketCandidates dedupes via a seen[int]struct{} map then writes
//     the deduped indices into a slice and sorts that slice before
//     return. Callers consume the sorted slice, not the seen map.
//   - The map iteration discipline observed by Plan 09-03's group
//     iteration (sortedGroups built from groupIndices, then walked)
//     carries forward: the bucket dispatch reads bucketCandidates'
//     sorted slice in deterministic order, so warning emission order
//     remains deterministic-by-construction with respect to slice-index
//     order.
//
// Per Open Question 2 resolution (09-RESEARCH.md): the bucket consumes
// pre-Normalised, pre-Tokenised names. tokeniseAll applies
// fuzzymatch.Normalise(item.Name, normOpts) (when the Scorer is
// constructed with normalisation enabled) THEN
// fuzzymatch.Tokenise(normalised, fuzzymatch.DefaultTokeniseOptions())
// to derive the per-Item token set. Items whose normalised forms
// produce identical token sets land in identical buckets — exactly
// what the Scorer itself sees at scoring time. Pitfall 5 (no
// double-normalisation) is preserved because the Scorer re-normalises
// internally during Score; tokeniseAll's output is used ONLY for
// bucket lookup, never as input to Scorer.Score.

package scan

import (
	"sort"
	"sync/atomic"

	"github.com/axonops/fuzzymatch"
)

// bucketThreshold is the group-size cutoff at which the token-bucket
// optimisation overtakes the naive nested-loop comparison.
// Empirically validated on darwin/arm64 (Apple M2) by
// BenchmarkScanCheck_BucketVsNaive_GroupSize at 2026-05-20 / Plan
// 09-04. The starting hypothesis from docs/requirements.md §12.5 was
// 50; the benchmark sweep over group sizes 10 / 25 / 50 / 75 / 100 /
// 200 (10,000 items per variant) confirmed the bucket path overtakes
// naive immediately at the first dispatch-eligible variant
// (GroupSize75), delivering a ~9× speedup that holds across larger
// groups. The crossover is empirically within ±10 of the hypothesis,
// so the constant remains 50. Per-group-size numbers (count=3):
//
//	GroupSize75   naive ≈ 1593 ms / bucket ≈ 161 ms — bucket 9.9× faster
//	GroupSize100  naive ≈ 1850 ms / bucket ≈ 211 ms — bucket 8.8× faster
//	GroupSize200  naive ≈ 3729 ms / bucket ≈ 419 ms — bucket 8.9× faster
//
// PERF-05 budget (10,000 items / 500 groups): 362 ms per Check —
// 5.5× under the 2 s spec. Full numbers recorded in bench.txt and
// in .planning/phases/09-collection-scan-sub-package/09-04-SUMMARY.md.
//
// Per 09-CONTEXT.md §5 D-08 (LOCKED): this is a package-private
// constant, NOT a Config.BucketThreshold field. Promotion to a public
// field is a non-breaking minor-version addition that can land in
// v1.x if consumers surface a real tuning need; for v1.0 the YAGNI
// case (user-memory feedback_yagni_for_tuning_knobs) wins.
const bucketThreshold = 50

// forceNaivePath is a package-private test-only flag. When true,
// scan.Check suppresses the bucket dispatch and falls back to the
// naive nested-loop pass even on groups exceeding bucketThreshold. It
// is exposed to the external scan_test package via the
// SetForceNaivePath() helper declared in export_test.go so the
// PropCheck_BucketEquivalentToNaive property test can compare the
// bucket and naive paths on the same input.
//
// Threat model T-09-04-02 mitigation: forceNaivePath is unreachable
// from consumer code in production. The test file export_test.go is
// compiled only by `go test`; no regular build path can flip the flag.
// The default value (false, the zero value) preserves the production
// dispatch.
//
// Concurrency: stored as an atomic.Bool so concurrent Check
// invocations (including parallel tests under -race -shuffle=on) can
// safely read the flag while a different test goroutine sets it. The
// production code path performs an atomic Load() on every dispatch
// decision; the test code path performs an atomic Store() via the
// SetForceNaivePath helper. Tests that flip the flag are responsible
// for restoring it to false on exit (the property test uses defer).
var forceNaivePath atomic.Bool

// tokenBucket is the inverted-index data structure that maps a token
// to the slice of item indices containing that token. The slice for
// each token is sorted ascending and de-duplicated; buildBucket
// enforces both invariants.
//
// The map is consumed via direct lookup (bucket[token]) inside
// bucketCandidates — never iterated on the output path. Iteration over
// tokenBucket would expose Go's map-iteration randomisation to the
// warning order, violating the determinism guarantee
// (.claude/skills/determinism-standards).
type tokenBucket map[string][]int

// tokeniseAll returns one tokens slice per item in items, in input
// order. For each item:
//
//   - If applyNormalisation == true, the Item.Name is first Normalised
//     via fuzzymatch.Normalise(name, normOpts), then Tokenised via
//     fuzzymatch.Tokenise(normalised, DefaultTokeniseOptions()).
//   - If applyNormalisation == false, the Item.Name is Tokenised
//     directly (no Normalise step), mirroring the
//     WithoutNormalisation() Scorer construction.
//
// Per Open Question 2 resolution (09-RESEARCH.md) the Normalise step
// uses the Scorer's normalisation options (NormalisationOptions
// returned by Scorer.NormalisationOptions()) so the bucket keys mirror
// what the Scorer sees at scoring time. The Tokenise step uses the
// default tokenise options because the bucket needs a fixed
// identifier-friendly split — case-sensitive bucket keys would defeat
// the point on snake_case-vs-camelCase pairs.
//
// Pitfall 7 (09-RESEARCH.md): each Item.Name is tokenised at most
// ONCE per Check invocation. Check calls tokeniseAll once at entry
// and reuses the returned slice for both the within-group bucket
// build and the cross-group bucket build.
//
// Cost: O(N · L) where N is len(items) and L is the average name
// length. The Tokenise ASCII fast path (Phase 8.5 Q8b) makes this
// effectively zero-allocation for short ASCII identifiers and is
// load-bearing for the PERF-05 < 2s budget.
//
// Safe for concurrent use only if the input slice is read-only. The
// returned [][]string is a fresh slice owned by the caller.
func tokeniseAll(items []Item, normOpts fuzzymatch.NormalisationOptions, applyNormalisation bool) [][]string {
	out := make([][]string, len(items))
	tokOpts := fuzzymatch.DefaultTokeniseOptions()
	for i, item := range items {
		var src string
		if applyNormalisation {
			src = fuzzymatch.Normalise(item.Name, normOpts)
		} else {
			src = item.Name
		}
		out[i] = fuzzymatch.Tokenise(src, tokOpts)
	}
	return out
}

// buildBucket constructs a tokenBucket from the supplied item indices
// and pre-computed token sets. For each idx in indices, every token
// in tokens[idx] becomes a key in the bucket whose value slice
// contains idx exactly once (duplicates are silently filtered — e.g.
// an item whose Name tokenises to ["id", "id", "user"] contributes
// idx to bucket["id"] only once).
//
// The resulting bucket's value slices are sorted ascending and
// de-duplicated. Sorting at build time means bucketCandidates does
// not need to re-sort each lookup result.
//
// Cost: O(N · T · log T) where N is len(indices) and T is the
// average token-set size. The per-token-key sort is bounded by the
// number of items containing that token, not by N. In the typical
// snake_case-vs-camelCase identifier workload, most tokens appear in
// only a handful of items, so the sort cost is negligible.
//
// Implementation note on dedup: tokens within a single Item.Name can
// repeat (e.g. "id_id" → ["id", "id"]). The append-then-sort-then-dedup
// approach is simpler than tracking per-item seen sets, and the dedup
// happens only once per build, not per lookup. The cost is O(T log T)
// per bucket key, dominated by the sort.
func buildBucket(indices []int, tokens [][]string) tokenBucket {
	bucket := make(tokenBucket)
	for _, idx := range indices {
		for _, tok := range tokens[idx] {
			bucket[tok] = append(bucket[tok], idx)
		}
	}
	// Sort + dedupe each value slice. The map is iterated here because
	// the loop body operates on each value slice independently of the
	// others — no output ordering depends on the map's iteration
	// order. (The bucket's contents become deterministic after this
	// loop completes; downstream code reads via direct key lookup.)
	for tok, idxs := range bucket {
		sort.Ints(idxs)
		// In-place dedup: write head walks distinct values; read head
		// scans the (already-sorted) slice. Equivalent to the canonical
		// Go dedup-after-sort idiom.
		w := 0
		for r := 0; r < len(idxs); r++ {
			if w == 0 || idxs[r] != idxs[w-1] {
				idxs[w] = idxs[r]
				w++
			}
		}
		bucket[tok] = idxs[:w]
	}
	return bucket
}

// bucketCandidates returns the deduplicated, sorted slice of item
// indices that share at least one token with items[srcIdx], EXCLUDING
// srcIdx itself (Pitfall 8: no self-pairs).
//
// Algorithm:
//
//  1. Walk tokens[srcIdx]; for each token t, gather bucket[t].
//  2. Deduplicate via a seen[int]struct{} set (map used internally,
//     never iterated on the output path).
//  3. Exclude srcIdx itself.
//  4. Sort the result slice ascending and return.
//
// The output slice is freshly allocated by this function and is
// deterministic across calls (same inputs → byte-identical slice). The
// seen map is contained inside this function — Go's map-iteration
// randomisation never reaches the caller because the seen map's keys
// are written to the result slice and then sorted.
//
// Cost: O(C · log C) where C is the candidate-set size. In the
// typical workload C is a small fraction of group size, so this is
// effectively linear in group size with a small constant.
//
// Safe for concurrent use given a stable bucket (no concurrent
// modification of the passed-in tokenBucket).
func bucketCandidates(srcIdx int, bucket tokenBucket, tokens [][]string) []int {
	// Pre-size seen with a rough estimate of the candidate-set ceiling.
	// Each token in tokens[srcIdx] contributes at most len(bucket[tok])
	// candidates; the typical token's bucket is small but adversarial
	// inputs may make it large. A modest initial capacity reduces map
	// rehashing without pre-allocating wastefully.
	//
	// The 4× multiplier on len(tokens[srcIdx]) is a heuristic — typical
	// identifier corpora have ~2–6 items sharing any given token after
	// normalisation/tokenisation, so 4×token-count over-allocates
	// slightly for the small-bucket common case and under-allocates
	// for adversarial high-collision corpora (in which case the map
	// grows organically). Empirical sweet spot on the AxonOps audit
	// corpus; revisit if profiling shows >10% rehash cost on a real
	// workload. The choice is purely an alloc-amortisation tuning
	// constant; correctness is independent of the initial capacity.
	seen := make(map[int]struct{}, len(tokens[srcIdx])*4)
	for _, tok := range tokens[srcIdx] {
		for _, idx := range bucket[tok] {
			if idx == srcIdx {
				continue // exclude self
			}
			seen[idx] = struct{}{}
		}
	}
	if len(seen) == 0 {
		return nil
	}
	out := make([]int, 0, len(seen))
	for idx := range seen {
		out = append(out, idx)
	}
	// The iteration of the seen map is non-deterministic; the sort
	// below re-establishes deterministic order before return.
	sort.Ints(out)
	return out
}
