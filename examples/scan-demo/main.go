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

// Package main demonstrates the Phase 9 scan sub-package public API
// across three sections, each exercising a different combination of
// Config flags:
//
//   - Section 1 — within-group only with the default Scorer and a
//     stock DefaultConfig (CompareAcrossGroups stays false). This is
//     the simplest possible scan: every item in a single group is
//     compared against every other item, suppression is off, and the
//     Scorer's baked-in 0.85 threshold governs emission.
//
//   - Section 2 — within + cross-group with DefaultConfig
//     (CrossGroupThresholdBoost = 0.05, CompareIdenticalAcrossGroups
//     = false). The default identical-name suppression silences
//     pairs whose Names coincide across distinct Groups so the
//     consumer is not drowned in operator-reused-name noise.
//
//   - Section 3 — all three suppression modes composed in a single
//     scan: per-item SilenceLint (Rule 1), SuppressedPairs canonical
//     pair (Rule 2), and cross-group identical-name default (Rule
//     3). The rules compose via OR — any rule firing suppresses the
//     pair pre-emission.
//
// The output is deterministic and byte-stable across runs and
// platforms: warnings are sorted by (Kind, NameA, NameB, GroupA,
// GroupB) inside scan.Check, and the rendering below uses
// fixed-width columns + %.4f score formatting.
//
// Run with:
//
//	go run ./examples/scan-demo/
package main

import (
	"fmt"
	"os"

	"github.com/axonops/fuzzymatch"
	"github.com/axonops/fuzzymatch/scan"
)

// defaultS is the opinionated six-algorithm Phase 8 default Scorer.
// Constructed once at package init time; immutable; safe for
// concurrent use. Shared by all three demonstration sections so the
// Section 2 / Section 3 cross-group boost is comparable to the
// Section 1 within-group threshold.
var defaultS = fuzzymatch.DefaultScorer()

// printWarning renders one Warning as a single deterministic line.
// The column widths are chosen so the longest realistic Name fits
// without truncation; %.4f score formatting keeps float rendering
// platform-stable.
//
// Format columns:
//
//	Kind          (12 chars) — "WithinGroup" / "AcrossGroups"
//	NameA / NameB (18 chars each) — raw item Names
//	GroupA/GroupB (10 chars each) — raw item Groups
//	Score         (8 chars)  — %.4f composite
func printWarning(w scan.Warning) {
	fmt.Printf("  %-12s %-18s %-18s %-10s %-10s %.4f\n",
		w.Kind, w.NameA, w.NameB, w.GroupA, w.GroupB, w.Score)
}

// printHeader prints the column-header line + separator above the
// warnings table for one section. Same column widths as printWarning.
func printHeader() {
	fmt.Printf("  %-12s %-18s %-18s %-10s %-10s %s\n",
		"Kind", "NameA", "NameB", "GroupA", "GroupB", "Score")
	fmt.Println("  " + // 2-space indent to align with the rows
		"------------ " +
		"------------------ " +
		"------------------ " +
		"---------- " +
		"---------- " +
		"------")
}

// runWithinGroupOnly demonstrates the simplest scan: a single group,
// no cross-group pass, no suppression. The Scorer's baked-in 0.85
// threshold governs emission; only pairs above it surface as
// warnings.
func runWithinGroupOnly() {
	fmt.Println("Section 1: within-group only (default Scorer + DefaultConfig)")
	fmt.Println()

	items := []scan.Item{
		{Name: "user_id", Group: "login"},
		{Name: "userId", Group: "login"},
		{Name: "user_name", Group: "login"},
		{Name: "lastSeen", Group: "login"},
	}
	cfg := scan.DefaultConfig(defaultS)
	// CompareAcrossGroups stays false — within-group only.

	warnings, err := scan.Check(items, cfg)
	if err != nil {
		fmt.Fprintf(os.Stderr, "scan-demo Section 1: %v\n", err)
		os.Exit(1)
	}

	printHeader()
	for _, w := range warnings {
		printWarning(w)
	}
	fmt.Printf("  (%d warning(s))\n", len(warnings))
}

// runWithinAndAcrossGroups demonstrates the default cross-group pass
// with the opinionated boost (0.05) and identical-name suppression
// (CompareIdenticalAcrossGroups = false). Two items share the same
// raw Name ("user_id") across distinct groups — without Rule 3, this
// would emit a noisy KindAcrossGroups warning every time an operator
// reused a common identifier across schemas. With Rule 3 active (the
// DefaultConfig default), the cross-group pair is silently
// suppressed. Items with similar Names within the same group still
// surface as KindWithinGroup warnings — Rule 3 only fires on the
// cross-group identical-name pre-emission check.
func runWithinAndAcrossGroups() {
	fmt.Println("Section 2: within + cross-group (DefaultConfig, boost=0.05, identical-cross=false)")
	fmt.Println()

	items := []scan.Item{
		{Name: "user_id", Group: "login"},
		{Name: "userId", Group: "login"},    // within-group similar — surfaces
		{Name: "user_id", Group: "profile"}, // cross-group identical — Rule 3 silences
	}
	cfg := scan.DefaultConfig(defaultS)
	cfg.CompareAcrossGroups = true
	// CompareIdenticalAcrossGroups stays false — Rule 3 active.

	warnings, err := scan.Check(items, cfg)
	if err != nil {
		fmt.Fprintf(os.Stderr, "scan-demo Section 2: %v\n", err)
		os.Exit(1)
	}

	printHeader()
	for _, w := range warnings {
		printWarning(w)
	}
	fmt.Printf("  (%d warning(s))\n", len(warnings))
}

// runWithSuppression demonstrates all three suppression rules
// composed. Each pair below would otherwise emit a warning; each is
// silenced by a different rule:
//
//   - items[0] / items[1]: Rule 1 — items[0] carries SilenceLint=true,
//     so the within-group pair is suppressed.
//   - items[2] / items[3]: Rule 2 — the consumer-supplied
//     SuppressedPairs entry catches this within-group pair via the
//     Scorer's normalisation pipeline (case + separator).
//   - items[4] / items[5]: Rule 3 — identical raw Name "request_id"
//     across distinct Groups, with the default
//     CompareIdenticalAcrossGroups=false silencing it.
//
// All three pairs sit above the within-group threshold; the
// "0 warning(s)" output is the suppression composition firing, not
// the threshold.
func runWithSuppression() {
	fmt.Println("Section 3: suppression composition (SilenceLint + SuppressedPairs + cross-group identical default)")
	fmt.Println()

	items := []scan.Item{
		// Rule 1 — SilenceLint on items[0] suppresses the
		// within-group pair (items[0], items[1]).
		{Name: "user_id", Group: "login", SilenceLint: true},
		{Name: "userId", Group: "login"},

		// Rule 2 — SuppressedPairs entry below catches this
		// within-group pair after canonicalisation.
		{Name: "user_name", Group: "profile"},
		{Name: "userName", Group: "profile"},

		// Rule 3 — cross-group identical-Name default suppresses
		// this pair when CompareAcrossGroups=true and
		// CompareIdenticalAcrossGroups=false (the DefaultConfig
		// default).
		{Name: "request_id", Group: "audit"},
		{Name: "request_id", Group: "metrics"},
	}
	cfg := scan.DefaultConfig(defaultS)
	cfg.CompareAcrossGroups = true // enable Rule 3 reachability
	cfg.SuppressedPairs = [][2]string{{"user_name", "userName"}}

	warnings, err := scan.Check(items, cfg)
	if err != nil {
		fmt.Fprintf(os.Stderr, "scan-demo Section 3: %v\n", err)
		os.Exit(1)
	}

	printHeader()
	for _, w := range warnings {
		printWarning(w)
	}
	fmt.Printf("  (%d warning(s))\n", len(warnings))
}

func main() {
	runWithinGroupOnly()
	fmt.Println()
	runWithinAndAcrossGroups()
	fmt.Println()
	runWithSuppression()
}
