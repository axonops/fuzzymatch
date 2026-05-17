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

// dispatch_monge_elkan.go registers Monge-Elkan into the dispatch table
// at package load time. This file MUST be the sole writer to
// dispatch[AlgoMongeElkan] (slot 13 — see algoid.go for the slot map).
//
// The dispatch table maps AlgoID to (a, b string) float64 — a fixed
// signature with no place for the inner AlgoID parameter. Per
// CONTEXT.md §4 LOCKED + Phase 8.5 Q3 symmetric-by-default rename, the
// dispatch wrapper binds:
//
//   - The SYMMETRIC default MongeElkanScore (post-rename unsuffixed name)
//     — so AlgoMongeElkan participates in the standard
//     PropAlgorithmScore_Symmetric property test set without exemption.
//   - AlgoJaroWinkler as the default inner — the standard reference for
//     fuzzy-name matching in the Monge & Elkan 1996 paper's empirical
//     evaluation and the most widely-known inner-metric choice across
//     SecondString / py_stringmatching / RapidFuzz lineages.
//
// The directional surface (MongeElkanScoreAsymmetric) is reachable via
// the public API but NOT via the dispatch table — direction-sensitive
// scoring is an advanced use case that requires the caller to be aware
// of which argument's tokens drive the per-token-max reduction. The
// dispatch wrapper provides the symmetric "best-of-both-directions"
// reduction as the safer default for Scorer / Extract integrations.
//
// Phase 8.5 Q3 — opts removal: the v0.x dispatch wrapper forwarded
// DefaultNormalisationOptions() to a NormalisationOptions parameter that
// was inert inside the body (Tokenise carries its own
// DefaultTokeniseOptions()). The parameter has been removed and the
// dispatch wrapper now passes only (a, b, inner).
//
// Phase 7 forward-compatibility: AlgoSoundex / AlgoDoubleMetaphone /
// AlgoNYSIIS / AlgoMRA were ADDED to permittedMongeElkanInner (in
// monge_elkan.go); the dispatch wrapper itself is UNCHANGED — the
// JaroWinkler default is preserved.
//
// See algoid.go for the dispatch array declaration and its design
// rationale.

package fuzzymatch

// init registers the MongeElkan dispatch entry. Q14b option A (Phase 8.5
// Plan 15a) — explicit init replaces the var _ = func() bool {...}()
// pattern per the determinism-standards SKILL (pure-write into a
// pre-allocated slot; no IO, no time, no goroutines, no ordering
// dependency on other init functions). Per CONTEXT.md §4 LOCKED +
// Phase 8.5 Q3 symmetric-by-default rename:
//   - the SYMMETRIC default MongeElkanScore is dispatched (so
//     AlgoMongeElkan participates in the standard symmetric property-
//     test set);
//   - the default inner is AlgoJaroWinkler.
//
// See the file-level godoc above for the rationale and the
// forward-compatibility notes for Phase 7's phonetic-tier additions.
func init() {
	dispatch[AlgoMongeElkan] = func(a, b string) float64 {
		return MongeElkanScore(a, b, AlgoJaroWinkler)
	}
}
