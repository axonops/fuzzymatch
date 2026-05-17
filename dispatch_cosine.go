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

// dispatch_cosine.go registers CosineScore into the dispatch table at
// package load time. This file MUST be the sole writer to
// dispatch[AlgoCosine] (see algoid.go for the slot map).
//
// The dispatch table maps AlgoID to (a, b string) float64 — a fixed
// signature with no place for the q-gram size n. Per CONTEXT.md
// "Claude's Discretion" Deferred §4 ("Specific n overrides happen via
// the Phase 8 Scorer option layer"), the dispatch wrapper binds a
// default n = 3 (the canonical trigram value). Consumers needing a
// non-default n call CosineScore directly or, in Phase 8, use
// WithCosineAlgorithm(weight, n) on the Scorer.
//
// Only CosineScore is dispatched — CosineScoreRunes is public but not
// wired into the dispatch table (the dispatch signature is the
// byte-path one).
//
// See algoid.go for the dispatch array declaration and its design
// rationale.

package fuzzymatch

// init registers the Cosine dispatch entry. Q14b option A (Phase 8.5
// Plan 15a) — explicit init replaces the var _ = func() bool {...}()
// pattern per the determinism-standards SKILL (pure-write into a
// pre-allocated slot; no IO, no time, no goroutines, no ordering
// dependency on other init functions). Default n = 3 trigram per
// CONTEXT.md Deferred §4.
func init() {
	dispatch[AlgoCosine] = func(a, b string) float64 {
		return CosineScore(a, b, 3)
	}
}
