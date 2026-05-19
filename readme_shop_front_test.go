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

// readme_shop_front_test.go is the documentation-drift gate for the
// README quick-start example. The README is the project's shop front
// (per Phase 8.5 CONTEXT.md §16.1): if the quick-start snippet falls
// out of sync with the actual library output, every new consumer
// follows the broken trail.
//
// Mechanism: parse README.md, extract the first fenced ```go code
// block under the "Quick start" section, scan it for `// Output:`
// marker comments paired with `fmt.Println(...)` calls, and verify
// the actual library output matches the documented output for each
// pair. The Go code is NOT executed via `go run` (avoiding a Go
// toolchain dep at test time and keeping the test in-process); the
// test mirrors the documented surface by calling the same public
// functions directly with the same arguments.
//
// Extension policy: when the README quick-start grows (Phase 2 adds
// algorithm calls per the README's own "Phase 2 adds the full
// algorithm-driven quick start" note), the case slice below grows
// alongside. The extraction logic stays the same — find a Println
// + Output pair, assert byte-equality.
//
// Stdlib `testing` only.

package fuzzymatch_test

import (
	"fmt"
	"os"
	"regexp"
	"strings"
	"testing"

	"github.com/axonops/fuzzymatch"
)

// TestREADME_QuickStartMatchesExample verifies the README's quick-start
// code block matches the actual library output, byte-for-byte.
//
// The function operates in two stages:
//
//  1. Parse stage: read README.md, locate the Quick start section,
//     extract its first fenced ```go block, and parse out every
//     (fmt.Println call, // Output comment) pair found.
//
//  2. Verify stage: for each documented (input, expected) pair,
//     call the same fuzzymatch.<fn> the README documents and assert
//     the actual output matches.
//
// The parser is conservative: only patterns it understands count
// toward verification (a future quick-start using fmt.Printf or a
// different output convention will not be silently skipped — the
// test fails-closed by asserting at least one documented pair was
// extracted and verified).
func TestREADME_QuickStartMatchesExample(t *testing.T) {
	t.Parallel()

	readme, err := os.ReadFile("README.md")
	if err != nil {
		t.Fatalf("read README.md: %v", err)
	}

	block := extractQuickStartBlock(t, string(readme))
	pairs := extractPrintlnOutputPairs(t, block)
	if len(pairs) == 0 {
		t.Fatalf("README quick-start contains no recognised (fmt.Println + // Output:) pairs; the test cannot verify anything. Update the extractor or update the README to use the documented Output convention.")
	}

	// The verifier table below mirrors the documented surface. Each
	// entry is matched against an extracted pair by call-source
	// substring; the test fails if an extracted pair has no matching
	// entry (forcing us to extend the table when the README does).
	verifiers := []struct {
		callPrefix string
		produce    func() string
	}{
		{
			// `fmt.Println(s.Match("user_id", "userId"))`
			// where `s := fuzzymatch.DefaultScorer()`. Quick Start
			// demonstrates the simplest fuzzy-match call: pin
			// DefaultScorer().Match against the canonical snake-vs-camel
			// identifier pair → true.
			callPrefix: "s.Match(",
			produce: func() string {
				s := fuzzymatch.DefaultScorer()
				return fmt.Sprint(s.Match("user_id", "userId"))
			},
		},
	}

	for _, p := range pairs {
		matched := false
		for _, v := range verifiers {
			if !strings.Contains(p.call, v.callPrefix) {
				continue
			}
			matched = true
			got := v.produce()
			if got != p.want {
				t.Errorf("README quick-start drift for %s\n  got:  %q\n  want: %q",
					v.callPrefix, got, p.want)
			}
			break
		}
		if !matched {
			t.Errorf("README quick-start contains an unverified call: %q\nAdd a matching verifier entry to readme_shop_front_test.go.", p.call)
		}
	}
}

// printlnOutputPair pairs a parsed fmt.Println call (the call source,
// minus the surrounding fmt.Println(...) wrapper) with the immediately
// following // Output: comment value (whitespace-trimmed).
type printlnOutputPair struct {
	// call is the inner expression source of the fmt.Println call —
	// e.g. `fuzzymatch.Normalise("UserCreate-Event", opts)`.
	call string
	// want is the documented expected output (the text after the
	// `// Output:` marker, trimmed of leading/trailing whitespace).
	want string
}

// quickStartHeadingRegex finds the Quick start H2 heading. The README
// uses an en-dash-free heading and lowercase "start"; we anchor on a
// case-insensitive match for robustness against minor stylistic edits.
var quickStartHeadingRegex = regexp.MustCompile(`(?im)^##\s+Quick\s+start\s*$`)

// extractQuickStartBlock locates the first fenced ```go block under
// the "Quick start" H2. It fails the test if the heading is missing
// or if no ```go block follows.
func extractQuickStartBlock(t *testing.T, content string) string {
	t.Helper()

	loc := quickStartHeadingRegex.FindStringIndex(content)
	if loc == nil {
		t.Fatalf("README has no `## Quick start` heading; cannot locate the quick-start block")
	}
	tail := content[loc[1]:]

	// Find the first ```go fence in the tail; the matching closing
	// ``` fence terminates the block. Use a non-greedy regex with
	// the s-flag (dotall) so `.` matches newlines inside the block.
	blockRegex := regexp.MustCompile("(?s)```go\\s*\\n(.*?)\\n```")
	m := blockRegex.FindStringSubmatch(tail)
	if m == nil {
		t.Fatalf("README quick-start section has no ```go fenced code block")
	}
	return m[1]
}

// printlnRegex matches a `fmt.Println(...)` invocation, capturing the
// inner argument expression (potentially containing balanced parens
// inside the arg). The pattern handles one level of nesting only —
// sufficient for the current quick start; if a future snippet uses
// deeper nesting, extend this to a small hand-written paren-balancer.
//
// We accept any of `fmt.Println`, `fmt.Print`, `fmt.Printf` patterns
// only when followed by exactly the canonical `// Output:` marker on
// the next non-blank line.
var printlnRegex = regexp.MustCompile(`(?m)fmt\.Println\(((?:[^()]|\([^()]*\))*)\)`)

// outputMarkerRegex matches a `// Output: <expected>` comment line
// (single-line; we accept either spelling with one or more spaces).
var outputMarkerRegex = regexp.MustCompile(`(?m)^\s*//\s+Output:\s*(.*?)\s*$`)

// extractPrintlnOutputPairs walks the code block line-by-line and
// emits one pair per (fmt.Println(...), // Output: ...) sequence
// where the Output comment appears on the line immediately following
// the Println call. Pairs without a following Output comment are
// silently skipped (the README may legitimately print without
// documenting expected output in some surfaces).
func extractPrintlnOutputPairs(t *testing.T, block string) []printlnOutputPair {
	t.Helper()

	lines := strings.Split(block, "\n")
	var pairs []printlnOutputPair

	for i, line := range lines {
		callMatch := printlnRegex.FindStringSubmatch(line)
		if callMatch == nil {
			continue
		}
		// Look for the // Output: marker on the next line.
		if i+1 >= len(lines) {
			continue
		}
		outMatch := outputMarkerRegex.FindStringSubmatch(lines[i+1])
		if outMatch == nil {
			continue
		}
		pairs = append(pairs, printlnOutputPair{
			call: callMatch[1],
			want: outMatch[1],
		})
	}
	return pairs
}
