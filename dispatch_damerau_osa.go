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

// dispatch_damerau_osa.go registers DamerauLevenshteinOSAScore into the
// dispatch table at package load time. This file MUST be the sole writer to
// dispatch[AlgoDamerauLevenshteinOSA].
//
// See algoid.go for the dispatch array declaration and its design rationale.
// The var _ = func() bool { ... }() idiom is the Phase-2-canonical form for
// package-level side effects without init() (per determinism-standards §13.5
// and docs/requirements.md §5(12)); this file copies the pattern from
// dispatch_levenshtein.go, changing only the AlgoXxx and XxxScore identifiers.

package fuzzymatch

// _ ensures dispatch[AlgoDamerauLevenshteinOSA] is populated before any call
// to the Scorer (Phase 8) or Extract (Phase 10) that reads the dispatch table.
// The var _ = func()bool{...}() idiom is the canonical way to run
// package-level side effects without init() (per determinism-standards §13.5).
var _ = func() bool {
	dispatch[AlgoDamerauLevenshteinOSA] = DamerauLevenshteinOSAScore
	return true
}()
