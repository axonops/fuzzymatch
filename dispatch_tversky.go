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

// dispatch_tversky.go registers TverskyScore into the dispatch table at
// package load time. This file MUST be the sole writer to
// dispatch[AlgoTversky] (see algoid.go for the slot map).
//
// The dispatch table maps AlgoID to (a, b string) float64 — a fixed
// signature with no place for the q-gram size n NOR for the Tversky
// weights α and β. Per CONTEXT.md "Claude's Discretion" the dispatch
// wrapper binds n=3 (the canonical trigram value) AND α=β=1.0 (the
// Jaccard-equivalent weights). This is a deliberate compromise:
//
//   - α=β=1.0 produces output IDENTICAL to QGramJaccardScore on the
//     same q-gram multisets (Tversky 1977 §2 — the Jaccard degeneracy
//     is the cleanest algebraic special case).
//   - The real Tversky use case — direction-sensitive asymmetric
//     scoring with α ≠ β — lands in Phase 8 via the Scorer option
//     WithTverskyAlgorithm(weight, alpha, beta) which forwards the
//     user-supplied α and β to TverskyScore directly.
//
// Why Jaccard fallback rather than Dice fallback (α=β=0.5)? Both are
// algebraic-equivalence wrappers; Jaccard is the more widely-known
// reference algorithm and the equivalence is verified by RV-T3 in
// tversky_test.go (TverskyScore("abcd", "abce", 2, 1.0, 1.0) ==
// QGramJaccardScore("abcd", "abce", 2) bit-for-bit). Choosing Jaccard
// also means the dispatch-table output for AlgoTversky and
// AlgoQGramJaccard are identical for the default n=3 case — a
// reviewer-friendly property that makes the fallback-equivalence
// auditable from the dispatch table alone.
//
// Only TverskyScore is dispatched — TverskyScoreRunes is public but
// not wired into the dispatch table (the dispatch signature is the
// byte-path one).
//
// See algoid.go for the dispatch array declaration and its design
// rationale. The var _ = func() bool { ... }() idiom is the canonical
// Phase-2-onward form for package-level side effects without init()
// (per determinism-standards §13.5 and docs/requirements.md §5(12)).

package fuzzymatch

// _ ensures dispatch[AlgoTversky] is populated before any call to the
// Scorer (Phase 8) or Extract (Phase 10) that reads the dispatch
// table. Default n=3 trigram + α=β=1.0 (Jaccard-equivalent) per
// CONTEXT.md "Claude's Discretion" — see the file-level godoc above
// for the rationale and the algebraic equivalence proof.
var _ = func() bool {
	dispatch[AlgoTversky] = func(a, b string) float64 {
		return TverskyScore(a, b, 3, 1.0, 1.0)
	}
	return true
}()
