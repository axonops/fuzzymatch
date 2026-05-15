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

// dispatch_qgram_jaccard.go registers QGramJaccardScore into the dispatch
// table at package load time. This file MUST be the sole writer to
// dispatch[AlgoQGramJaccard] (slot 9 — see algoid.go for the slot map).
//
// The dispatch table maps AlgoID to (a, b string) float64 — a fixed
// signature with no place for the q-gram size n. Per CONTEXT.md
// "Claude's Discretion" Deferred §4 ("Specific n overrides happen via
// the Phase 8 Scorer option layer"), the dispatch wrapper binds a
// default n = 3 (the canonical trigram value). Consumers needing a
// non-default n call QGramJaccardScore directly or, in Phase 8, use
// WithQGramJaccardAlgorithm(weight, n) on the Scorer.
//
// Only QGramJaccardScore is dispatched — QGramJaccardScoreRunes is
// public but not wired into the dispatch table (the dispatch signature
// is the byte-path one).
//
// See algoid.go for the dispatch array declaration and its design
// rationale. The var _ = func() bool { ... }() idiom is the canonical
// Phase-2-onward form for package-level side effects without init()
// (per determinism-standards §13.5 and docs/requirements.md §5(12)).

package fuzzymatch

// _ ensures dispatch[AlgoQGramJaccard] is populated before any call to
// the Scorer (Phase 8) or Extract (Phase 10) that reads the dispatch
// table. Default n = 3 trigram per CONTEXT.md Deferred §4.
var _ = func() bool {
	dispatch[AlgoQGramJaccard] = func(a, b string) float64 {
		return QGramJaccardScore(a, b, 3)
	}
	return true
}()
