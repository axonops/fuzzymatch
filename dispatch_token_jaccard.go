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

// dispatch_token_jaccard.go registers TokenJaccardScore into the dispatch
// table at package load time. This file MUST be the sole writer to
// dispatch[AlgoTokenJaccard] (slot 17 — see algoid.go for the slot map).
//
// TokenJaccard is parameter-free (no n, no α/β), so the dispatch wrapper
// is the function value directly — no closure needed. Mirrors
// dispatch_lcsstr.go and dispatch_token_sort_ratio.go (the canonical
// no-params direct-dispatch pattern).
//
// See algoid.go for the dispatch array declaration and its design
// rationale.

package fuzzymatch

// init registers the TokenJaccard dispatch entry. Q14b option A
// (Phase 8.5 Plan 15a) — explicit init replaces the var _ = func() bool
// {...}() pattern per the determinism-standards SKILL (pure-write into a
// pre-allocated slot; no IO, no time, no goroutines, no ordering
// dependency on other init functions).
func init() {
	dispatch[AlgoTokenJaccard] = TokenJaccardScore
}
