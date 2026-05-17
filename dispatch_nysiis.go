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

// dispatch_nysiis.go wires NYSIISScore into the dispatch table at slot
// AlgoNYSIIS (25 — see algoid.go). The registration is performed by a
// package-level init-alternative (`var _ = func() bool {...}()`) to avoid
// init() side effects per docs/requirements.md §5(12).

package fuzzymatch

// _ registers NYSIISScore in the global dispatch table at AlgoNYSIIS (25).
// This runs before any test or caller can invoke the dispatch table, ensuring
// that MongeElkanScore / MongeElkanScoreAsymmetric and Scorer dispatch paths
// see the registered function.
var _ = func() bool {
	dispatch[AlgoNYSIIS] = NYSIISScore
	return true
}()
