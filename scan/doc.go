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

// Package scan is the Layer-3 collection-scan sub-package of the
// fuzzymatch library. It composes a Phase 8 *fuzzymatch.Scorer with a
// pre-flight validation pass, within/cross-group similarity passes, a
// token-bucket optimisation for large groups, suppression composition,
// and a deterministic output sort.
//
// The scan sub-package is optional. Consumers wanting only algorithm
// functions or only the Scorer never import scan and incur no cost from
// its existence — the root package has no dependency on this sub-package.
//
// Typical usage:
//
//	items := []scan.Item{
//	    {Name: "user_id",   Group: "login"},
//	    {Name: "userId",    Group: "login"},
//	    {Name: "user_name", Group: "profile"},
//	}
//	cfg := scan.DefaultConfig(fuzzymatch.DefaultScorer())
//	cfg.CompareAcrossGroups = true
//	warnings, err := scan.Check(items, cfg)
//	if err != nil {
//	    // handle ErrNilScorer / ErrInvalidItem / ErrInvalidConfig
//	}
//	for _, w := range warnings {
//	    // ... consume the deterministic []Warning ...
//	}
//
// Concurrency: Check is a pure function. No goroutines, no channels,
// no mutexes. Safe for concurrent invocation on disjoint inputs from
// any number of goroutines without external synchronisation.
//
// See docs/requirements.md §12 for the authoritative specification and
// docs/scan.md for the consumer-facing guide.
package scan
