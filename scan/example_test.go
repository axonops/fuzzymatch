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

package scan_test

import (
	"fmt"

	"github.com/axonops/fuzzymatch"
	"github.com/axonops/fuzzymatch/scan"
)

// ExampleKind_String demonstrates the CamelCase String() form returned
// by the two Kind constants — the canonical labels appear unchanged in
// scan.Warning rendering, JSON output, and BDD scenarios.
func ExampleKind_String() {
	fmt.Println(scan.KindWithinGroup.String())
	fmt.Println(scan.KindAcrossGroups.String())
	// Output:
	// WithinGroup
	// AcrossGroups
}

// ExampleDefaultConfig demonstrates the opinionated default Config
// constructed from a Scorer. CrossGroupThresholdBoost is baked at 0.05
// per 09-CONTEXT.md §2 D-04 (SPEC OVERRIDE: the default location
// migrated from Config.CrossGroupThresholdBoost's field godoc to this
// helper).
func ExampleDefaultConfig() {
	cfg := scan.DefaultConfig(fuzzymatch.DefaultScorer())
	fmt.Printf("%.2f\n", cfg.CrossGroupThresholdBoost)
	// Output: 0.05
}

// ExampleCheck_withSuppression demonstrates the three suppression rules
// of Plan 09-05 in a single program. Each item exercises one rule:
//
//   - items[0] / items[1]: Rule 1 (per-item SilenceLint) — items[0]
//     carries SilenceLint=true, so the within-group pair with items[1]
//     is suppressed regardless of similarity.
//   - items[2] / items[3]: Rule 2 (SuppressedPairs canonical-pair
//     lookup) — the consumer-supplied SuppressedPairs entry
//     ["user_name", "userName"] catches this pair before emission.
//     Canonicalisation absorbs argument order.
//   - items[4] / items[5]: Rule 3 (cross-group identical-name default) —
//     two items with identical raw Name "request_id" in distinct
//     groups would otherwise emit a KindAcrossGroups warning under
//     CompareAcrossGroups=true, but CompareIdenticalAcrossGroups=false
//     (the DefaultConfig default) suppresses it.
//
// All three suppressed pairs sit above the within-group similarity
// threshold; the "0 warnings" output is the suppression composition
// firing, not the threshold.
func ExampleCheck_withSuppression() {
	items := []scan.Item{
		// Rule 1 — SilenceLint on items[0] suppresses the within-group
		// pair (items[0], items[1]).
		{Name: "user_id", Group: "login", SilenceLint: true},
		{Name: "userId", Group: "login"},

		// Rule 2 — SuppressedPairs entry below catches this within-group
		// pair.
		{Name: "user_name", Group: "profile"},
		{Name: "userName", Group: "profile"},

		// Rule 3 — cross-group identical-name default suppresses this
		// pair when CompareAcrossGroups is true.
		{Name: "request_id", Group: "audit"},
		{Name: "request_id", Group: "metrics"},
	}
	cfg := scan.DefaultConfig(fuzzymatch.DefaultScorer())
	cfg.CompareAcrossGroups = true // enable cross-group so Rule 3 is reachable
	cfg.SuppressedPairs = [][2]string{{"user_name", "userName"}}
	warnings, err := scan.Check(items, cfg)
	if err != nil {
		panic(err)
	}
	fmt.Printf("%d warnings\n", len(warnings))
	// Output: 0 warnings
}
