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

// errors_test.go pins the contract of the package-level sentinel errors:
//
//   - errors.Is wrap-identity: fmt.Errorf("%w", ErrX) is discoverable
//     as ErrX via errors.Is (the canonical Go discrimination pattern).
//   - Each sentinel carries a distinct message — no two sentinels alias
//     to the same Error() string.
//   - Each message starts with the "fuzzymatch: " package prefix per
//     .claude/skills/go-coding-standards.
//   - Each message body (after the prefix) is lowercase and carries no
//     trailing punctuation per the same standard.
//   - Each sentinel is distinguishable from every other via errors.Is
//     (pairwise distinct as values).
//
// Stdlib testing only — no testify in root.

package fuzzymatch_test

import (
	"errors"
	"fmt"
	"strings"
	"testing"
	"unicode"

	"github.com/axonops/fuzzymatch"
)

// sentinelCases enumerates every sentinel exported from errors.go so
// the table-driven tests below stay in sync with the source as new
// sentinels are added (later phases extend this list).
//
// Order is not load-bearing — the tests iterate the slice.
func sentinelCases() []struct {
	name string
	err  error
} {
	return []struct {
		name string
		err  error
	}{
		{"ErrInvalidAlgoID", fuzzymatch.ErrInvalidAlgoID},
		{"ErrInvalidInnerAlgo", fuzzymatch.ErrInvalidInnerAlgo},
		{"ErrInvalidQGramSize", fuzzymatch.ErrInvalidQGramSize},
		{"ErrInvalidTverskyParam", fuzzymatch.ErrInvalidTverskyParam},
		// Phase 8 sentinels — Scorer construction surface.
		{"ErrEmptyScorer", fuzzymatch.ErrEmptyScorer},
		{"ErrInvalidWeight", fuzzymatch.ErrInvalidWeight},
		{"ErrInvalidThreshold", fuzzymatch.ErrInvalidThreshold},
		{"ErrMissingThreshold", fuzzymatch.ErrMissingThreshold},
		// Phase 8.5 sentinels — typed-panic value for library bugs.
		{"ErrInternalInvariantViolated", fuzzymatch.ErrInternalInvariantViolated},
	}
}

// TestSentinels_Identity asserts errors.Is(sentinel, sentinel) is true
// for each Phase 8 Scorer sentinel and that the canonical message
// strings match the contract pinned in CONTEXT.md §2 / PATTERNS.md.
// This is the Phase 8 plan 08-01 acceptance gate for the four new
// Scorer-construction sentinels.
func TestSentinels_Identity(t *testing.T) {
	cases := []struct {
		name    string
		err     error
		message string
	}{
		{
			name:    "ErrEmptyScorer",
			err:     fuzzymatch.ErrEmptyScorer,
			message: "fuzzymatch: scorer has no algorithms (pass at least one WithAlgorithm option or use DefaultScorer)",
		},
		{
			name:    "ErrInvalidWeight",
			err:     fuzzymatch.ErrInvalidWeight,
			message: "fuzzymatch: invalid algorithm weight (must be > 0)",
		},
		{
			name:    "ErrInvalidThreshold",
			err:     fuzzymatch.ErrInvalidThreshold,
			message: "fuzzymatch: invalid threshold (must be in [0.0, 1.0])",
		},
		{
			name:    "ErrMissingThreshold",
			err:     fuzzymatch.ErrMissingThreshold,
			message: "fuzzymatch: scorer threshold required (pass WithThreshold or use DefaultScorer)",
		},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			if c.err == nil {
				t.Fatalf("%s is nil — Phase 8 sentinel missing from errors.go", c.name)
			}
			if !errors.Is(c.err, c.err) {
				t.Errorf("errors.Is(%s, %s) = false; want true (sentinel identity)", c.name, c.name)
			}
			if got := c.err.Error(); got != c.message {
				t.Errorf("%s.Error() = %q; want %q", c.name, got, c.message)
			}
			const prefix = "fuzzymatch: "
			if !strings.HasPrefix(c.err.Error(), prefix) {
				t.Errorf("%s.Error() = %q; must begin with %q", c.name, c.err.Error(), prefix)
			}
		})
	}
}

// TestSentinels_WrapIdentity asserts each sentinel survives one level
// of fmt.Errorf("%w") wrapping discoverable by errors.Is. This is the
// canonical Go pattern for sentinel discrimination across layers.
func TestSentinels_WrapIdentity(t *testing.T) {
	for _, c := range sentinelCases() {
		t.Run(c.name, func(t *testing.T) {
			if c.err == nil {
				t.Fatalf("%s: sentinel is nil — exported var missing from errors.go", c.name)
			}
			wrapped := fmt.Errorf("scorer: %w", c.err)
			if !errors.Is(wrapped, c.err) {
				t.Errorf("errors.Is(fmt.Errorf(\"scorer: %%w\", %s), %s) = false; want true", c.name, c.name)
			}
		})
	}
}

// TestSentinels_DistinctMessages asserts no two sentinels share the
// same Error() string. Two sentinels with the same message would be
// indistinguishable from a log-grep perspective even though errors.Is
// would still distinguish them.
func TestSentinels_DistinctMessages(t *testing.T) {
	cases := sentinelCases()
	// Map check, internal — output isn't a slice/map, so the
	// no-map-iteration rule doesn't apply (test internals are exempt).
	seen := make(map[string]string, len(cases))
	for _, c := range cases {
		msg := c.err.Error()
		if prev, dup := seen[msg]; dup {
			t.Errorf("%s.Error() = %q duplicates %s.Error()", c.name, msg, prev)
		}
		seen[msg] = c.name
	}
	if len(seen) != len(cases) {
		t.Errorf("distinct message count = %d; want %d", len(seen), len(cases))
	}
}

// TestSentinels_StartWithPackagePrefix asserts every sentinel message
// begins with the "fuzzymatch: " prefix so that wrapper composition
// produces readable error chains.
func TestSentinels_StartWithPackagePrefix(t *testing.T) {
	const prefix = "fuzzymatch: "
	for _, c := range sentinelCases() {
		t.Run(c.name, func(t *testing.T) {
			msg := c.err.Error()
			if !strings.HasPrefix(msg, prefix) {
				t.Errorf("%s.Error() = %q; must begin with %q", c.name, msg, prefix)
			}
		})
	}
}

// TestSentinels_LowercaseAndNoTrailingPunctuation asserts every
// sentinel's body text (after the "fuzzymatch: " prefix) starts with a
// lowercase rune and carries no trailing punctuation ('.', '!', or
// '?'). This matches the Go convention codified in
// .claude/skills/go-coding-standards/SKILL.md ("Error strings:
// lowercase, no punctuation") which constrains the FIRST-rune
// capitalisation only — embedded identifier names (e.g.
// "WithAlgorithm", "DefaultScorer") inside parenthesised remediation
// hints are permitted because they refer to public API symbols and
// disambiguate the actionable next step for the consumer.
func TestSentinels_LowercaseAndNoTrailingPunctuation(t *testing.T) {
	const prefix = "fuzzymatch: "
	for _, c := range sentinelCases() {
		t.Run(c.name, func(t *testing.T) {
			msg := c.err.Error()
			body := strings.TrimPrefix(msg, prefix)
			if body == msg {
				// Already failed in TestSentinels_StartWithPackagePrefix; nothing more to do here.
				return
			}
			// Trailing punctuation check: '.', '!', '?' all disallowed.
			if last := body[len(body)-1]; last == '.' || last == '!' || last == '?' {
				t.Errorf("%s.Error() = %q has trailing punctuation %q", c.name, msg, string(last))
			}
			// First-rune lowercase gate. Per Effective Go's error-string
			// convention, the message is concatenable into other
			// contexts — capitalising the FIRST word would produce
			// awkward compositions ("scorer: Invalid weight"). Embedded
			// identifier names later in the message are referencing
			// public Go symbols and are permitted (and required for
			// disambiguation of remediation hints).
			for _, r := range body {
				if unicode.IsUpper(r) {
					t.Errorf("%s.Error() body %q starts with uppercase rune %q (only embedded identifier names may be capitalised)", c.name, body, string(r))
				}
				// Only inspect the first rune.
				break
			}
		})
	}
}

// TestSentinels_AreDistinctAsValues asserts pairwise distinctness via
// errors.Is. This catches the regression where a future refactor
// might accidentally alias two sentinels to the same errors.New value
// (e.g. `ErrInvalidInnerAlgo = ErrInvalidAlgoID`).
func TestSentinels_AreDistinctAsValues(t *testing.T) {
	cases := sentinelCases()
	for i, a := range cases {
		for j, b := range cases {
			if i == j {
				continue
			}
			if errors.Is(a.err, b.err) {
				t.Errorf("errors.Is(%s, %s) = true; sentinels must be pairwise distinct", a.name, b.name)
			}
		}
	}
}

// TestSentinels_NotNil asserts every sentinel is initialised (non-nil).
// This is a belt-and-braces gate against a future refactor accidentally
// declaring a sentinel without assigning it (`var ErrX error` instead
// of `var ErrX = errors.New(...)`).
func TestSentinels_NotNil(t *testing.T) {
	for _, c := range sentinelCases() {
		t.Run(c.name, func(t *testing.T) {
			if c.err == nil {
				t.Errorf("%s is nil — must be initialised via errors.New", c.name)
			}
		})
	}
}
