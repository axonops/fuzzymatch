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

// dispatch_partial_ratio.go registers PartialRatioScore (the byte-path
// surface) into the dispatch table at package load time. This file
// MUST be the sole writer to dispatch[AlgoPartialRatio] (slot 16 —
// see algoid.go AlgoPartialRatio).
//
// Only the BYTE-PATH PartialRatioScore is registered. PartialRatioScoreRunes
// is public but NOT dispatched — the dispatch table signature is
// (a, b string) float64 (byte path); the rune surface is direct-call-only.
// This matches the LCSStr convention (lcsstr.go lines 79-88) where
// LongestCommonSubstringRunes / LCSStrScoreRunes are public but not
// registered in dispatch[AlgoLCSStr].
//
// PartialRatioScore has no parameters beyond (a, b string), so the
// dispatch wrapper is the function value directly — NO closure is
// needed (mirrors dispatch_lcsstr.go, dispatch_token_sort_ratio.go,
// dispatch_token_set_ratio.go). This contrasts with the q-gram
// dispatchers in plans 05-01 through 05-04, which bind a default n
// parameter via a closure.
//
// See algoid.go for the dispatch array declaration and its design
// rationale. The `var _ = func() bool { ... }()` idiom is the
// canonical Phase-2-onward form for package-level side effects without
// init() (per determinism-standards §13.5 and docs/requirements.md
// §5(12)).

package fuzzymatch

// _ ensures dispatch[AlgoPartialRatio] is populated before any call to
// the Scorer (Phase 8) or Extract (Phase 10) that reads the dispatch
// table. No closure: PartialRatioScore's signature already matches the
// dispatch table type. PartialRatioScoreRunes is NOT registered — see
// the file-header godoc for the dispatch-byte-path-only convention.
var _ = func() bool {
	dispatch[AlgoPartialRatio] = PartialRatioScore
	return true
}()
