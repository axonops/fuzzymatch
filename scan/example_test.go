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

// example_test.go carries the runnable godoc Example functions for
// the scan sub-package. Each Example compiles + runs as part of
// `go test ./scan/...`; the `// Output:` comment gates expected
// stdout against actual stdout, so any drift surfaces as a test
// failure.
//
// Required surface (DX-02 — one Example per demonstrable public
// symbol):
//
//   - ExampleKind_String          — Kind constants + String() form
//   - ExampleDefaultConfig        — opinionated config factory
//   - ExampleCheck                — within-group happy path
//   - ExampleCheck_acrossGroups   — cross-group pass with the
//                                   identical-Name default
//   - ExampleCheck_suppression    — all three suppression rules
//                                   composed
//
// Naming convention: variant suffix is lowercase after the
// underscore per https://pkg.go.dev/testing#hdr-Examples.

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

// ExampleCheck demonstrates the within-group happy path: four items
// in a single group, the default Scorer + DefaultConfig, no
// cross-group pass, no suppression. The "user_id" / "userId" pair
// normalises to the same form under the default Scorer's
// normalisation pipeline, scoring 1.0 — well above the 0.85
// threshold — and surfaces as a KindWithinGroup warning.
func ExampleCheck() {
	items := []scan.Item{
		{Name: "user_id", Group: "login"},
		{Name: "userId", Group: "login"},
		{Name: "user_name", Group: "login"},
		{Name: "lastSeen", Group: "login"},
	}
	cfg := scan.DefaultConfig(fuzzymatch.DefaultScorer())
	warnings, err := scan.Check(items, cfg)
	if err != nil {
		panic(err)
	}
	for _, w := range warnings {
		fmt.Printf("%s %s/%s (%s/%s) %.4f\n",
			w.Kind, w.NameA, w.NameB, w.GroupA, w.GroupB, w.Score)
	}
	// Output:
	// WithinGroup userId/user_id (login/login) 1.0000
}

// ExampleCheck_acrossGroups demonstrates the cross-group pass with
// the default identical-Name suppression (Rule 3) active. The
// "user_id" Name appears in two distinct Groups; the
// CompareIdenticalAcrossGroups=false default silences that
// cross-group pair so operators reusing common identifiers across
// schemas are not drowned in noise.
//
// The within-group pass still surfaces the "user_id" / "userId" pair
// (login group), independent of the cross-group setting.
func ExampleCheck_acrossGroups() {
	items := []scan.Item{
		{Name: "user_id", Group: "login"},
		{Name: "userId", Group: "login"},    // within-group similar — surfaces
		{Name: "user_id", Group: "profile"}, // cross-group identical — Rule 3 silences
	}
	cfg := scan.DefaultConfig(fuzzymatch.DefaultScorer())
	cfg.CompareAcrossGroups = true
	// CompareIdenticalAcrossGroups stays false — Rule 3 active.

	warnings, err := scan.Check(items, cfg)
	if err != nil {
		panic(err)
	}
	for _, w := range warnings {
		fmt.Printf("%s %s/%s (%s/%s) %.4f\n",
			w.Kind, w.NameA, w.NameB, w.GroupA, w.GroupB, w.Score)
	}
	// Output:
	// WithinGroup userId/user_id (login/login) 1.0000
}

// ExampleCheck_suppression demonstrates the three suppression rules
// of Plan 09-05 in a single program. Each item exercises one rule:
//
//   - items[0] / items[1]: Rule 1 (per-item SilenceLint) — items[0]
//     carries SilenceLint=true, so the within-group pair with
//     items[1] is suppressed regardless of similarity.
//   - items[2] / items[3]: Rule 2 (SuppressedPairs canonical-pair
//     lookup) — the consumer-supplied SuppressedPairs entry
//     ["user_name", "userName"] catches this pair before emission.
//     Canonicalisation absorbs argument order.
//   - items[4] / items[5]: Rule 3 (cross-group identical-name default)
//     — two items with identical raw Name "request_id" in distinct
//     groups would otherwise emit a KindAcrossGroups warning under
//     CompareAcrossGroups=true, but CompareIdenticalAcrossGroups=false
//     (the DefaultConfig default) suppresses it.
//
// All three suppressed pairs sit above the within-group similarity
// threshold; the "0 warnings" output is the suppression composition
// firing, not the threshold.
func ExampleCheck_suppression() {
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
