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

// dispatch_lcsstr.go registers LCSStrScore into the dispatch table at
// package load time. This file MUST be the sole writer to
// dispatch[AlgoLCSStr] (slot 8 — see algoid.go for the slot map).
//
// Only LCSStrScore is dispatched — the dispatch table maps AlgoID to
// (a, b string) float64, so the string-returning surfaces
// LongestCommonSubstring / LongestCommonSubstringRunes and the rune-path
// variant LCSStrScoreRunes are public but not dispatched.
//
// See algoid.go for the dispatch array declaration and its design rationale.
// The var _ = func() bool { ... }() idiom is the Phase-2-canonical form for
// package-level side effects without init() (per determinism-standards §13.5
// and docs/requirements.md §5(12)).

package fuzzymatch

// _ ensures dispatch[AlgoLCSStr] is populated before any call to the Scorer
// (Phase 8) or Extract (Phase 10) that reads the dispatch table. The
// var _ = func()bool{...}() idiom is the canonical way to run package-level
// side effects without init() (per determinism-standards §13.5).
var _ = func() bool {
	dispatch[AlgoLCSStr] = LCSStrScore
	return true
}()
