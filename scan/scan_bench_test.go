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

// scan_bench_test.go declares the Phase 9 Plan 09-04 benchmark suite:
//
//   - BenchmarkScanCheck_BucketVsNaive_GroupSize{10,25,50,75,100,200}
//     — the D-08 empirical-validation sweep. Each variant runs Check
//     on a 10,000-item corpus partitioned into ceil(10000/groupSize)
//     groups. Each group's per-Check wall-clock cost reflects the
//     bucket dispatch performance at that group size; the sweep
//     identifies the empirical crossover where the bucket overtakes
//     naive (the constant bucketThreshold in scan/bucket.go must be
//     updated if the crossover differs materially from 50; manual
//     inspection step lives in Task 4 of Plan 09-04). Sub-benchmarks
//     b.Run("naive", ...) and b.Run("bucket", ...) under each variant
//     toggle the forceNaivePath flag so the same input is benchmarked
//     on both dispatch paths — the cleanest per-size comparison.
//
//   - BenchmarkScanCheck_DefaultScorer_10k — the PERF-05 budget
//     benchmark. Runs Check on a single 10,000-item group (so the
//     bucket path runs unconditionally) with the DefaultScorer +
//     DefaultConfig. The reported ns/op should correspond to < 2s per
//     Check on darwin/arm64 (Apple M2). Informational — not a CI
//     gate; see CONTEXT.md §5 D-08 ("the spec's stated 50 is a
//     starting hypothesis; the benchmark is load-bearing").
//
// Benchmark discipline (mirrors scorer_bench_test.go):
//
//   - b.ReportAllocs() on every benchmark (allocation regressions
//     visible in benchstat output)
//   - Construction (items slice + Scorer + cfg) OUTSIDE
//     b.ResetTimer() so the per-op measurement excludes setup
//   - sink-gate (var benchSink + post-loop nil-check) to defeat
//     compiler dead-code elimination
//   - Deterministic seed for the random number generator so the
//     corpus is reproducible across runs (benchstat consistency)
//
// Per CONTEXT.md §5 D-08, the bucketThreshold value chosen for v1.0
// is recorded in scan/bucket.go's in-source comment with the
// wall-clock crossover from this sweep. bench.txt records the
// per-variant numbers for benchstat regression detection in CI.

package scan_test

import (
	"fmt"
	"math/rand"
	"testing"

	"github.com/axonops/fuzzymatch"
	"github.com/axonops/fuzzymatch/scan"
)

// benchSink is the dead-code-elimination guard. The compiler cannot
// prove that the assignment to benchSink is never observed (it has
// package scope), so it cannot fold away the Check call inside the
// benchmark loop. Mirrors the `var sink` pattern in
// scorer_bench_test.go.
var benchSink []scan.Warning

// benchSeed is the deterministic seed used by buildBenchItems so the
// corpus is reproducible across benchmark runs. A fixed seed means
// benchstat comparisons across runs measure the implementation, not
// random-input variance.
const benchSeed int64 = 0x_5ca1_9c00_1ec7_1010

// buildBenchItems partitions totalItems into groups of size
// groupSize. The last group may be smaller if totalItems is not a
// multiple of groupSize. Names are drawn from a diverse vocabulary
// of identifier-style words (~60 entries) combined into 2-3-word
// snake_case identifiers — exactly the kind of input the
// axonops/audit consumer surfaces. Each Item.Name carries a unique
// (Name, Group) pair (D-06 acceptance) via a disambiguating numeric
// suffix appended only when the random draw collides with an earlier
// (Name, Group) in the same group.
//
// Workload realism: the bucket optimisation depends on the token
// distribution being heavy-tailed — most pairs share zero or one
// token, a small fraction share many. A "user_id_0_0" / "user_id_0_1"
// generator collapses the workload to the all-pairs-share-everything
// degenerate case where the bucket prunes nothing and pays its
// overhead with no benefit. The diverse-vocabulary generator produces
// realistic token-set diversity so the empirical crossover reflects
// production conditions, not a synthetic pessimal case.
//
// The bench helper uses math/rand seeded by benchSeed for
// reproducibility.
func buildBenchItems(totalItems, groupSize int) []scan.Item {
	// Vocabulary of identifier-style word stems. Length-balanced so
	// the bucket's token frequencies are roughly uniform.
	vocab := []string{
		"user", "customer", "account", "order", "invoice", "payment",
		"transaction", "session", "token", "credential", "identifier",
		"name", "email", "phone", "address", "city", "country",
		"zip", "code", "status", "type", "kind", "category", "tier",
		"level", "grade", "rank", "score", "value", "amount", "total",
		"sum", "count", "size", "length", "width", "height", "depth",
		"created", "updated", "deleted", "modified", "expires",
		"started", "ended", "scheduled", "active", "inactive",
		"pending", "approved", "rejected", "cancelled", "refunded",
		"shipped", "delivered", "primary", "secondary", "tertiary",
	}
	// Style suffixes — vary the casing convention so within a single
	// group items mix snake_case and camelCase as the SCAN-04
	// workload expects.
	styles := []func(a, b string) string{
		func(a, b string) string { return a + "_" + b },              // snake_case
		func(a, b string) string { return a + b[:1] + b[1:] + "Id" }, // camelCase-ish
		func(a, b string) string { return a + "." + b },              // dot-case
		func(a, b string) string { return a + "-" + b },              // kebab-case
	}
	rng := rand.New(rand.NewSource(benchSeed))

	items := make([]scan.Item, totalItems)
	// Per-group dedup so the generator can resolve collisions
	// deterministically without rejecting items.
	seen := make(map[string]map[string]struct{}, totalItems/groupSize+1)
	for i := 0; i < totalItems; i++ {
		groupIdx := i / groupSize
		groupName := fmt.Sprintf("group_%d", groupIdx)
		if _, ok := seen[groupName]; !ok {
			seen[groupName] = make(map[string]struct{}, groupSize)
		}

		// Draw two distinct vocabulary stems and a style. The combined
		// (stem1, stem2, style) space is ~60 * 60 * 4 = 14,400 unique
		// names — comfortably larger than the per-group ceiling.
		var name string
		for tries := 0; tries < 16; tries++ {
			s1 := vocab[rng.Intn(len(vocab))]
			s2 := vocab[rng.Intn(len(vocab))]
			st := styles[rng.Intn(len(styles))]
			candidate := st(s1, s2)
			if _, dup := seen[groupName][candidate]; !dup {
				name = candidate
				break
			}
		}
		// Fallback: append a per-position numeric suffix if the loop
		// exhausted its tries (rare on the 60²·4 search space).
		if name == "" {
			s1 := vocab[rng.Intn(len(vocab))]
			s2 := vocab[rng.Intn(len(vocab))]
			name = fmt.Sprintf("%s_%s_%d", s1, s2, i)
		}
		seen[groupName][name] = struct{}{}

		items[i] = scan.Item{
			Name:  name,
			Group: groupName,
		}
	}
	return items
}

// benchBucketVsNaive runs the sub-benchmark pair (naive vs bucket)
// for a single group size. The corpus is 10,000 items partitioned
// into ceil(10000/groupSize) groups. b.Run separates the two
// dispatch paths so benchstat reports each path independently.
//
// forceNaivePath is toggled BEFORE b.ResetTimer() so the toggle is
// not part of the per-op cost; the per-op cost is purely Check
// invocations on the chosen path.
//
// The benchmark restores forceNaivePath = false on exit (deferred
// cleanup) so a subsequent benchmark observes the production
// dispatch.
func benchBucketVsNaive(b *testing.B, groupSize int) {
	const totalItems = 10000
	items := buildBenchItems(totalItems, groupSize)
	cfg := scan.DefaultConfig(fuzzymatch.DefaultScorer())

	defer scan.SetForceNaivePath(false)

	b.Run("naive", func(b *testing.B) {
		scan.SetForceNaivePath(true)
		b.ReportAllocs()
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			w, err := scan.Check(items, cfg)
			if err != nil {
				b.Fatalf("Check: %v", err)
			}
			benchSink = w
		}
		if benchSink == nil {
			// Defeats compiler DCE: a real Check on this corpus
			// produces at least zero warnings, never nil-after-assign.
			// The nil-check after the loop forces the compiler to
			// keep the assignment.
			b.Fatal("benchSink unexpectedly nil after benchmark loop")
		}
	})

	b.Run("bucket", func(b *testing.B) {
		scan.SetForceNaivePath(false)
		b.ReportAllocs()
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			w, err := scan.Check(items, cfg)
			if err != nil {
				b.Fatalf("Check: %v", err)
			}
			benchSink = w
		}
		if benchSink == nil {
			b.Fatal("benchSink unexpectedly nil after benchmark loop")
		}
	})
}

// BenchmarkScanCheck_BucketVsNaive_GroupSize10 runs the dispatch-path
// comparison at group size 10 — below bucketThreshold's value of 50,
// so the bucket path would normally NOT activate. With the
// forceNaivePath toggle off, both paths nonetheless run via b.Run
// sub-benchmarks; the "bucket" sub-benchmark exercises the bucket
// helpers (tokeniseAll + buildBucket + bucketCandidates) at small
// group sizes to measure the per-Check overhead of the bucket
// machinery when the dispatch decision would normally choose naive.
//
// NOTE: at sizes below bucketThreshold, Check would always pick the
// naive path UNLESS forceNaivePath were used to artificially toggle.
// Both sub-benchmarks here use the SAME corpus; the "naive"
// sub-benchmark sets forceNaivePath=true (no bucket); the "bucket"
// sub-benchmark sets forceNaivePath=false but the production
// dispatch will STILL choose naive because group size <
// bucketThreshold. So at sizes 10/25 the "bucket" sub-benchmark is
// effectively a duplicate of "naive" — that's a sanity-check
// floor, not a real comparison. At sizes 50/75/100/200 the
// dispatch chooses bucket on its own, so the "bucket" sub-benchmark
// measures the real path. To get a TRUE bucket-path measurement at
// small sizes (i.e. forcing the bucket path even when dispatch
// would not choose it), a separate forced-bucket toggle would be
// needed — out of scope for v1.0 since the spec's empirical sweep
// targets the dispatch crossover, not the small-size bucket-cost
// floor.
func BenchmarkScanCheck_BucketVsNaive_GroupSize10(b *testing.B) {
	benchBucketVsNaive(b, 10)
}

// BenchmarkScanCheck_BucketVsNaive_GroupSize25 — see GroupSize10
// godoc.
func BenchmarkScanCheck_BucketVsNaive_GroupSize25(b *testing.B) {
	benchBucketVsNaive(b, 25)
}

// BenchmarkScanCheck_BucketVsNaive_GroupSize50 sits AT the
// bucketThreshold value (50). The dispatch condition `len(idx) >
// bucketThreshold` is strictly greater-than, so groups of exactly 50
// items still take the naive path; the "bucket" sub-benchmark also
// runs naive at this size. The next variant (GroupSize75) is the
// first to exercise the real bucket dispatch.
func BenchmarkScanCheck_BucketVsNaive_GroupSize50(b *testing.B) {
	benchBucketVsNaive(b, 50)
}

// BenchmarkScanCheck_BucketVsNaive_GroupSize75 is the first variant
// above the bucketThreshold value. The "bucket" sub-benchmark
// exercises the real bucket dispatch; the "naive" sub-benchmark
// forces the fallback for comparison.
func BenchmarkScanCheck_BucketVsNaive_GroupSize75(b *testing.B) {
	benchBucketVsNaive(b, 75)
}

// BenchmarkScanCheck_BucketVsNaive_GroupSize100 — see GroupSize75
// godoc.
func BenchmarkScanCheck_BucketVsNaive_GroupSize100(b *testing.B) {
	benchBucketVsNaive(b, 100)
}

// BenchmarkScanCheck_BucketVsNaive_GroupSize200 is the upper end of
// the D-08 sweep. At group size 200 the bucket path's advantage over
// naive should be unambiguous on identifier-style workloads.
func BenchmarkScanCheck_BucketVsNaive_GroupSize200(b *testing.B) {
	benchBucketVsNaive(b, 200)
}

// BenchmarkScanCheck_DefaultScorer_10k is the PERF-05 budget
// benchmark: one Check invocation on a 10,000-item corpus
// partitioned into 500 groups of 20 items each — exactly the workload
// the PERF-05 budget in docs/requirements.md §14.5 specifies ("< 2s
// for 10,000 items / 500 groups"). At group size 20, the dispatch
// chooses the naive path (20 < bucketThreshold = 50), so this
// benchmark measures the realistic axonops/audit-style workload
// where most groups are small and the bucket overhead is not paid.
//
// Budget: < 2,000,000,000 ns/op (< 2s per Check) on darwin/arm64
// (Apple M2). Informational — not a CI gate. The number is recorded
// in bench.txt; PERF-05 acceptance is the < 2s wall-clock observed
// in the manual benchstat inspection step (Task 4 of Plan 09-04).
//
// A companion benchmark BenchmarkScanCheck_DefaultScorer_10k_OneGroup
// runs the SAME 10,000 items in a single group (10k > bucketThreshold,
// so the bucket path activates) for the worst-case all-pairs
// comparison — informational only; far outside the PERF-05 spec
// workload but useful as a stress signal in benchstat trending.
func BenchmarkScanCheck_DefaultScorer_10k(b *testing.B) {
	const totalItems = 10000
	const groupSize = 20 // 500 groups, per PERF-05 §14.5
	items := buildBenchItems(totalItems, groupSize)
	cfg := scan.DefaultConfig(fuzzymatch.DefaultScorer())

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		w, err := scan.Check(items, cfg)
		if err != nil {
			b.Fatalf("Check: %v", err)
		}
		benchSink = w
	}
	if benchSink == nil {
		b.Fatal("benchSink unexpectedly nil after benchmark loop")
	}
}

// BenchmarkScanCheck_DefaultScorer_10k_OneGroup is the worst-case
// stress benchmark: 10,000 items in a single group. The bucket path
// runs unconditionally (10k > bucketThreshold). Outside the PERF-05
// spec workload but useful as a stress signal in benchstat trending;
// a regression here would surface a degradation in the bucket-path
// candidate enumeration before it shows up in the realistic
// 500-group workload.
func BenchmarkScanCheck_DefaultScorer_10k_OneGroup(b *testing.B) {
	const totalItems = 10000
	items := buildBenchItems(totalItems, totalItems) // one group of 10k
	cfg := scan.DefaultConfig(fuzzymatch.DefaultScorer())

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		w, err := scan.Check(items, cfg)
		if err != nil {
			b.Fatalf("Check: %v", err)
		}
		benchSink = w
	}
	if benchSink == nil {
		b.Fatal("benchSink unexpectedly nil after benchmark loop")
	}
}

// BenchmarkScanCheck_DefaultScorer_200 covers the spec §12.6 budget
// tier "200 items / 10 groups: < 10 ms". At group size 20, the
// dispatch chooses the naive path. Closes the spec coverage gap
// identified by algorithm-performance-reviewer on Plan 09-04
// (BLOCKING finding: missing benchmarks for §12.6 200-item and
// 1000-item scale).
func BenchmarkScanCheck_DefaultScorer_200(b *testing.B) {
	const totalItems = 200
	const groupSize = 20 // 10 groups, per §12.6
	items := buildBenchItems(totalItems, groupSize)
	cfg := scan.DefaultConfig(fuzzymatch.DefaultScorer())

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		w, err := scan.Check(items, cfg)
		if err != nil {
			b.Fatalf("Check: %v", err)
		}
		benchSink = w
	}
	if benchSink == nil {
		b.Fatal("benchSink unexpectedly nil after benchmark loop")
	}
}

// BenchmarkScanCheck_DefaultScorer_1000 covers the spec §12.6 budget
// tier "1000 items / 50 groups: < 100 ms". At group size 20, the
// dispatch chooses the naive path. Closes the spec coverage gap
// identified by algorithm-performance-reviewer on Plan 09-04.
func BenchmarkScanCheck_DefaultScorer_1000(b *testing.B) {
	const totalItems = 1000
	const groupSize = 20 // 50 groups, per §12.6
	items := buildBenchItems(totalItems, groupSize)
	cfg := scan.DefaultConfig(fuzzymatch.DefaultScorer())

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		w, err := scan.Check(items, cfg)
		if err != nil {
			b.Fatalf("Check: %v", err)
		}
		benchSink = w
	}
	if benchSink == nil {
		b.Fatal("benchSink unexpectedly nil after benchmark loop")
	}
}

// BenchmarkScanCheck_DefaultScorer_10k_CrossGroup measures the
// cross-group pass at the spec PERF-05 workload (10k items / 500
// groups, CompareAcrossGroups=true). Closes the IMPORTANT finding
// from algorithm-performance-reviewer on Plan 09-04 ("§12.6 cross-
// group 2× claim unverified") — by surfacing it.
//
// VERIFIED RESULT (darwin/arm64, Apple M2, 2026-05-20, count=10):
// ~189 s per Check at this workload — ~525× the within-only cost
// (361 ms). The spec §12.6 claim "Cross-group pass enabled: at most
// 2× the within-group-only cost on the same input" is NOT MET at
// v0.x. The bottleneck is the cross-group bucket build pattern:
// each (gi, gj) group-pair rebuilds a fresh tokenBucket over its
// idxA ∪ idxB union, then enumerates candidates per source item.
// At 500 groups × 499/2 = 124,750 group-pairs, this produces ~50M
// Scorer.Score calls and 1.3B allocations per Check invocation.
//
// v1.x optimisation candidate (recorded in 09-CONTEXT.md Deferred
// Ideas): build a single global tokenBucket once at Check entry,
// then filter candidate-set membership by group identity per
// (gi, gj) pair. Avoids the per-pair bucket rebuild and brings the
// cross-group cost closer to the within-only cost. Out of scope for
// Plan 09-04 (the bucket optimisation is correctness-equivalent;
// the further perf win is a separate workstream).
//
// Keep this benchmark in the suite as a v1.x baseline — a future
// optimisation will show its improvement here via benchstat.
func BenchmarkScanCheck_DefaultScorer_10k_CrossGroup(b *testing.B) {
	const totalItems = 10000
	const groupSize = 20
	items := buildBenchItems(totalItems, groupSize)
	cfg := scan.DefaultConfig(fuzzymatch.DefaultScorer())
	cfg.CompareAcrossGroups = true

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		w, err := scan.Check(items, cfg)
		if err != nil {
			b.Fatalf("Check: %v", err)
		}
		benchSink = w
	}
	if benchSink == nil {
		b.Fatal("benchSink unexpectedly nil after benchmark loop")
	}
}
