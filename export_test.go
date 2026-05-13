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

// export_test.go uses the build-tag-free `_test.go` suffix to re-export
// selected unexported symbols to the external (black-box) test package
// without polluting the public API surface. This is the canonical Go
// pattern for testing package internals from package fuzzymatch_test.
//
// Anything added here is visible only to tests; consumers never see it.

package fuzzymatch

// CanonicalMarshalForTest exposes the unexported canonicalMarshal helper
// to the external (black-box) fuzzymatch_test package so that
// golden_canonical_test.go can assert the locked v1.x byte contract
// without dragging canonicalMarshal into the public API.
//
// Do not use this in production code — it does not exist outside of
// _test.go compilation. Use WriteGoldenFile (the public test-maintenance
// wrapper) instead.
var CanonicalMarshalForTest = canonicalMarshal

// NumAlgorithmsForTest re-exports the unexported numAlgorithms constant
// to the external test package. Test code asserts the dispatch array
// is sized for exactly 23 entries; consumers never see this symbol.
const NumAlgorithmsForTest = numAlgorithms

// DispatchLenForTest returns the length of the unexported dispatch
// array. Test code uses this to assert (a) the array is sized for
// numAlgorithms entries, and (b) every entry is nil at the Phase 1
// state (algorithms register themselves from Phase 2 onwards). The
// function rather than a direct re-export is used to avoid copying
// the array (which contains function pointers).
func DispatchLenForTest() int { return len(dispatch) }

// DispatchEntryNilForTest reports whether the dispatch entry at the
// given index is currently nil. Phase 1 expects every entry to be
// nil; future phases populate entries as they implement algorithms.
//
// Out-of-range indices return false (the entry doesn't exist).
func DispatchEntryNilForTest(i int) bool {
	if i < 0 || i >= len(dispatch) {
		return false
	}
	return dispatch[i] == nil
}
