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

// dispatch_token_set_ratio.go registers TokenSetRatioScore into the
// dispatch table at package load time. This file MUST be the sole
// writer to dispatch[AlgoTokenSetRatio] (slot 15 — see algoid.go for
// the slot map).
//
// TokenSetRatioScore has no parameters beyond (a, b string), so the
// dispatch wrapper is the function value directly — NO closure is
// needed (mirrors dispatch_lcsstr.go, dispatch_token_sort_ratio.go,
// dispatch_levenshtein.go, etc.). This contrasts with the q-gram
// dispatchers in plans 05-01 through 05-04, which bind a default n
// parameter via a closure (q-gram algorithms carry the n parameter;
// the dispatch signature does not).
//
// See algoid.go for the dispatch array declaration and its design
// rationale. The `var _ = func() bool { ... }()` idiom is the
// canonical Phase-2-onward form for package-level side effects without
// init() (per determinism-standards §13.5 and docs/requirements.md
// §5(12)).

package fuzzymatch

// _ ensures dispatch[AlgoTokenSetRatio] is populated before any call
// to the Scorer (Phase 8) or Extract (Phase 10) that reads the
// dispatch table. No closure: TokenSetRatioScore's signature already
// matches the dispatch table type.
var _ = func() bool {
	dispatch[AlgoTokenSetRatio] = TokenSetRatioScore
	return true
}()
