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

// export_test.go exposes package-private symbols to the external
// scan_test test package for property tests that need to manipulate
// internal state.
//
// This file is compiled only during `go test` (no _test.go is included
// in regular builds), so the test-only hooks are unreachable from
// consumer code in production — T-09-04-02 mitigation per the threat
// model in 09-04-PLAN.md.
//
// Symbols exposed:
//
//   - SetForceNaivePath: setter for the package-private forceNaivePath
//     flag (declared in scan/bucket.go). The property test
//     PropCheck_BucketEquivalentToNaive flips this flag to compare the
//     bucket and naive dispatch paths on the same input.
//
//   - BucketThreshold: read-only accessor for the package-private
//     bucketThreshold constant. Used by benchmarks that need to size
//     fat groups relative to the dispatch threshold.

package scan

// SetForceNaivePath toggles the package-private forceNaivePath flag.
// When true, scan.Check suppresses the bucket dispatch and falls back
// to the naive nested-loop pass even on large groups. Used by
// PropCheck_BucketEquivalentToNaive to compare bucket vs naive output
// on the same input.
//
// Test-only. The setter is package-scope (callable only from scan_test
// via this export_test.go file), so consumer code cannot reach it in
// production.
func SetForceNaivePath(v bool) {
	forceNaivePath = v
}

// BucketThreshold returns the empirically-validated bucketThreshold
// private constant. Read-only accessor for benchmarks and tests that
// need to size inputs relative to the dispatch cutoff.
func BucketThreshold() int {
	return bucketThreshold
}
