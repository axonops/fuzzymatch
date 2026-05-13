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

// makefile_targets_test.go is the Makefile <-> CONTRIBUTING documentation-
// drift gate. It enforces bi-directional coverage:
//
//   1. Every target declared in the canonical Makefile target list (the
//      19 targets enumerated in CLAUDE.md "Makefile Targets") MUST exist
//      in the Makefile.
//   2. Every Makefile target MUST be documented in CONTRIBUTING.md (the
//      "Make Targets" section) OR carry a `## suppress: <reason>`
//      comment line immediately above its rule.
//   3. Reverse: every target name appearing as a backtick-quoted token
//      in CONTRIBUTING.md MUST exist in the Makefile (catches typos).
//
// Together these rules ensure that contributors invoking commands from
// CONTRIBUTING.md never hit a "no rule to make target" error and that
// new Makefile targets cannot land without a CONTRIBUTING update.

package fuzzymatch_test

import (
	"bufio"
	"os"
	"regexp"
	"sort"
	"strings"
	"testing"
)

// canonicalMakefileTargets is the locked list of targets from CLAUDE.md
// "Makefile Targets". Adding a new target requires:
//   1. Adding it to the Makefile.
//   2. Adding it to CONTRIBUTING.md's "Make Targets" section.
//   3. Adding it to this list.
//
// The fmt-check vs fmt distinction is intentional: fmt mutates, fmt-check
// reports. Both are user-facing targets.
var canonicalMakefileTargets = []string{
	"check",
	"test",
	"test-bdd",
	"test-fuzz",
	"lint",
	"vet",
	"fmt",
	"fmt-check",
	"bench",
	"bench-compare",
	"coverage",
	"tidy",
	"tidy-check",
	"security",
	"verify-deps-allowlist",
	"verify-determinism",
	"verify-license-headers",
	"release-check",
	"clean",
}

// targetRuleRe matches a Makefile target rule header. Captures the
// target name. Restricted to lowercase + hyphen so we don't pick up
// variable assignments like SHELL := /usr/bin/env bash.
//
// Pattern: line starts with a target name (lowercase letters / digits /
// hyphens), followed by a colon, optionally followed by dependencies.
// We also exclude `.PHONY:` because its dependency list is the
// canonical target enumeration, not a target definition.
var targetRuleRe = regexp.MustCompile(`^([a-z][a-z0-9-]*[a-z0-9]):(?:[[:space:]]|$)`)

// TestMakefile_HasCanonicalTargets enforces rule (1): every canonical
// target name must exist in the Makefile as a rule definition.
func TestMakefile_HasCanonicalTargets(t *testing.T) {
	t.Parallel()

	makefileContent, err := os.ReadFile("Makefile")
	if err != nil {
		t.Fatalf("read Makefile: %v", err)
	}

	declared := parseMakefileTargets(string(makefileContent))

	var missing []string
	for _, want := range canonicalMakefileTargets {
		if _, ok := declared[want]; !ok {
			missing = append(missing, want)
		}
	}

	if len(missing) > 0 {
		sort.Strings(missing)
		t.Errorf(
			"Makefile is missing %d canonical target(s):\n  %s\n\n"+
				"These targets are mandated by CLAUDE.md \"Makefile Targets\". "+
				"Either add the rule to the Makefile, or update CLAUDE.md and "+
				"canonicalMakefileTargets here in lockstep.",
			len(missing),
			strings.Join(missing, "\n  "),
		)
	}
}

// TestMakefile_TargetsDocumentedInContributing enforces rule (2):
// every Makefile target must be documented in CONTRIBUTING.md OR carry
// a `## suppress: <reason>` comment on the line before its rule. It
// also enforces rule (3): every target mentioned in CONTRIBUTING.md
// (as `<target>` or in a `make <target>` command) must exist in the
// Makefile.
func TestMakefile_TargetsDocumentedInContributing(t *testing.T) {
	t.Parallel()

	makefileContent, err := os.ReadFile("Makefile")
	if err != nil {
		t.Fatalf("read Makefile: %v", err)
	}
	contribContent, err := os.ReadFile("CONTRIBUTING.md")
	if err != nil {
		t.Fatalf("read CONTRIBUTING.md: %v", err)
	}

	declared := parseMakefileTargets(string(makefileContent))
	suppressed := parseSuppressedTargets(string(makefileContent))
	contribStr := string(contribContent)

	// Rule (2): each declared target appears in CONTRIBUTING or is suppressed.
	var undocumented []string
	for name := range declared {
		if _, ok := suppressed[name]; ok {
			continue
		}
		if !targetMentionedInContributing(contribStr, name) {
			undocumented = append(undocumented, name)
		}
	}
	if len(undocumented) > 0 {
		sort.Strings(undocumented)
		t.Errorf(
			"%d Makefile target(s) are not documented in CONTRIBUTING.md:\n  %s\n\n"+
				"Document them in CONTRIBUTING.md \"Make Targets\" section, "+
				"or add a `## suppress: <reason>` comment line in the Makefile "+
				"immediately before the target's rule.",
			len(undocumented),
			strings.Join(undocumented, "\n  "),
		)
	}

	// Rule (3): every backticked target name in CONTRIBUTING must exist
	// in the Makefile. This catches typos like writing "make benchcompare"
	// when the actual target is "bench-compare".
	mentioned := extractTargetMentionsFromContributing(contribStr)
	var unknown []string
	for _, name := range mentioned {
		if _, ok := declared[name]; !ok {
			unknown = append(unknown, name)
		}
	}
	if len(unknown) > 0 {
		sort.Strings(unknown)
		t.Errorf(
			"CONTRIBUTING.md mentions %d `make <target>` invocation(s) "+
				"whose target does not exist in the Makefile:\n  %s\n\n"+
				"Either fix the typo in CONTRIBUTING.md or add the target to the Makefile.",
			len(unknown),
			strings.Join(unknown, "\n  "),
		)
	}
}

// parseMakefileTargets returns the set of target names declared as
// `^<name>:` rule headers in the Makefile content. Excludes the
// `.PHONY:` directive (whose dependency list is just an enumeration,
// not a target body).
func parseMakefileTargets(content string) map[string]struct{} {
	out := make(map[string]struct{})
	scanner := bufio.NewScanner(strings.NewReader(content))
	for scanner.Scan() {
		line := scanner.Text()
		// Skip .PHONY (the directive enumerates targets; the rule itself
		// is not a target).
		if strings.HasPrefix(line, ".PHONY:") {
			continue
		}
		// Skip variable assignments (`X := …`, `X ?= …`, `X = …`). The
		// target regex restricts to lowercase already, so UPPERCASE names
		// like SHELL, GO, MODULE_PATH naturally don't match, but we skip
		// any assignment line just to be safe.
		if strings.Contains(line, ":=") || strings.Contains(line, "?=") {
			continue
		}
		match := targetRuleRe.FindStringSubmatch(line)
		if match != nil {
			out[match[1]] = struct{}{}
		}
	}
	return out
}

// parseSuppressedTargets returns the set of target names whose rule is
// preceded by a `## suppress: <reason>` comment line. Suppressing a
// target opts it out of the CONTRIBUTING-documentation requirement.
func parseSuppressedTargets(content string) map[string]struct{} {
	out := make(map[string]struct{})
	lines := strings.Split(content, "\n")
	for i := 1; i < len(lines); i++ {
		prev := strings.TrimSpace(lines[i-1])
		if !strings.HasPrefix(prev, "## suppress:") {
			continue
		}
		match := targetRuleRe.FindStringSubmatch(lines[i])
		if match != nil {
			out[match[1]] = struct{}{}
		}
	}
	return out
}

// targetMentionedInContributing reports whether the given target name
// appears in CONTRIBUTING.md in one of the recognised forms:
//
//   - `<name>` (backtick-quoted plain target)
//   - `make <name>` (backtick-quoted full invocation)
//   - section heading or bullet referring to the target verbatim
//
// We do a forgiving substring check anchored by the target's distinctive
// hyphenated form. Two-word targets like "bench-compare" are unique
// enough that a substring match is reliable.
func targetMentionedInContributing(content, target string) bool {
	// Common mentions in docs: `target`, `make target`, **target**, list
	// bullet "- `target`".
	patterns := []string{
		"`" + target + "`",
		"`make " + target + "`",
		"** " + target + " **",
		"**" + target + "**",
	}
	for _, p := range patterns {
		if strings.Contains(content, p) {
			return true
		}
	}
	return false
}

// extractTargetMentionsFromContributing returns the unique target names
// mentioned via `make <name>` invocations in CONTRIBUTING.md. Used by
// rule (3) to catch typos.
func extractTargetMentionsFromContributing(content string) []string {
	re := regexp.MustCompile("`make ([a-z][a-z0-9-]*[a-z0-9])`")
	matches := re.FindAllStringSubmatch(content, -1)
	seen := make(map[string]struct{})
	for _, m := range matches {
		seen[m[1]] = struct{}{}
	}
	out := make([]string, 0, len(seen))
	for name := range seen {
		out = append(out, name)
	}
	sort.Strings(out)
	return out
}
