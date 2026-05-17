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

// dispatch_mra.go wires MRAScore into the dispatch table at slot
// AlgoMRA (26 — see algoid.go).
//
// NOTE: MRACode and MRACompare are public but NOT dispatched. The dispatch
// table maps AlgoID to (a, b string) float64; MRACompare returns (bool, int)
// (the catalogue's only non-float64 return shape per CONTEXT.md §6 LOCKED).
// Consumers wanting the raw 0-6 NBS similarity counter call MRACompare directly.
// Consumers wanting binary 0.0/1.0 (e.g. Scorer, MongeElkan inner) call the
// dispatched MRAScore.

package fuzzymatch

// init registers the MRA dispatch entry at AlgoMRA (26). Q14b option A
// (Phase 8.5 Plan 15a) — explicit init replaces the var _ = func() bool
// {...}() pattern per the determinism-standards SKILL (pure-write into a
// pre-allocated slot; no IO, no time, no goroutines, no ordering
// dependency on other init functions). This runs before any test or
// caller can invoke the dispatch table, ensuring that MongeElkanScore /
// MongeElkanScoreAsymmetric and Scorer dispatch paths see the registered
// function.
func init() {
	dispatch[AlgoMRA] = MRAScore
}
