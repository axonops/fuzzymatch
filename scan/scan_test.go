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
//
// scan_test.go pins the Plan 09-01 foundation contract of the scan
// sub-package:
//
//   - Kind.String() returns the CamelCase labels for KindWithinGroup
//     and KindAcrossGroups; out-of-range values fall back to the
//     "Kind(N)" allocating form (mirrors warn_kind.go's discipline).
//   - The three sentinel errors (ErrNilScorer, ErrInvalidItem,
//     ErrInvalidConfig) are distinct values — errors.Is between any
//     two of them returns false; self-identity returns true.
//   - Check(nil, scan.Config{Scorer: nil}) returns ErrNilScorer (Plan
//     09-01 stub contract; the full validation pipeline lands in Plan
//     09-02).
//   - DefaultConfig returns the opinionated defaults locked in
//     09-CONTEXT.md §2 D-04 (CrossGroupThresholdBoost == 0.05,
//     CompareIdenticalAcrossGroups == false, etc.).
//   - Warning.Scores has the SPEC OVERRIDE type
//     map[fuzzymatch.AlgoID]float64 (verified by reflect).
//
// Stdlib `testing` only — no testify in root tests per CLAUDE.md and
// .claude/skills/go-coding-standards.

package scan_test

import (
	"errors"
	"reflect"
	"testing"

	"github.com/axonops/fuzzymatch"
	"github.com/axonops/fuzzymatch/scan"
)

// TestKind_String pins the CamelCase labels for the two locked Kind
// constants and the allocating fmt.Sprintf default branch for
// out-of-range values. The labels are part of the v1.x contract —
// scan.Warning rendering, golden files, BDD scenarios, and consumer
// diagnostics all depend on the stability of these strings.
func TestKind_String(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name string
		k    scan.Kind
		want string
	}{
		{"KindWithinGroup", scan.KindWithinGroup, "WithinGroup"},
		{"KindAcrossGroups", scan.KindAcrossGroups, "AcrossGroups"},
		{"zero-value-unspecified", scan.Kind(0), "Kind(0)"},
		{"out-of-range-positive", scan.Kind(99), "Kind(99)"},
		{"out-of-range-negative", scan.Kind(-1), "Kind(-1)"},
	}
	for _, c := range cases {
		c := c
		t.Run(c.name, func(t *testing.T) {
			t.Parallel()
			if got := c.k.String(); got != c.want {
				t.Errorf("Kind(%d).String() = %q; want %q", int(c.k), got, c.want)
			}
		})
	}
}

// TestSentinels_Distinct verifies the three scan-package sentinels are
// distinct errors.Is identities. Consumers discriminating between
// ErrNilScorer / ErrInvalidItem / ErrInvalidConfig via errors.Is
// receive correct results only if no two sentinels match each other.
// The self-identity assertions guard the canonical "errors.Is(e, e)
// returns true" contract that errors.New values satisfy by reference
// equality.
func TestSentinels_Distinct(t *testing.T) {
	t.Parallel()

	sentinels := []struct {
		name string
		err  error
	}{
		{"ErrNilScorer", scan.ErrNilScorer},
		{"ErrInvalidItem", scan.ErrInvalidItem},
		{"ErrInvalidConfig", scan.ErrInvalidConfig},
	}

	for i, a := range sentinels {
		for j, b := range sentinels {
			i, j := i, j
			a, b := a, b
			t.Run(a.name+"_vs_"+b.name, func(t *testing.T) {
				t.Parallel()
				got := errors.Is(a.err, b.err)
				want := i == j
				if got != want {
					t.Errorf("errors.Is(%s, %s) = %v; want %v", a.name, b.name, got, want)
				}
			})
		}
	}
}

// TestCheck_NilScorer pins the Plan 09-01 stub contract: when
// cfg.Scorer is nil, Check returns (nil, err) where errors.Is(err,
// ErrNilScorer) is true. The full validation pipeline (Plan 09-02) and
// similarity body (Plans 09-03..09-06) extend this contract; the
// foundation contract is the nil-Scorer fail-fast only.
func TestCheck_NilScorer(t *testing.T) {
	t.Parallel()

	warnings, err := scan.Check(nil, scan.Config{Scorer: nil})
	if warnings != nil {
		t.Errorf("warnings: got %v; want nil", warnings)
	}
	if !errors.Is(err, scan.ErrNilScorer) {
		t.Errorf("err: got %v; want errors.Is(err, ErrNilScorer)", err)
	}
}

// TestDefaultConfig_Defaults pins the opinionated default Config built
// from a non-nil Scorer. The values are locked in 09-CONTEXT.md §2
// D-04 (SPEC OVERRIDE: the 0.05 default location migrated from
// Config.CrossGroupThresholdBoost field godoc to DefaultConfig godoc;
// spec §12.1 line 1359 amended in lockstep).
func TestDefaultConfig_Defaults(t *testing.T) {
	t.Parallel()

	s := fuzzymatch.DefaultScorer()
	cfg := scan.DefaultConfig(s)

	if cfg.Scorer != s {
		t.Errorf("Scorer: got %p; want %p", cfg.Scorer, s)
	}
	if cfg.CompareAcrossGroups {
		t.Errorf("CompareAcrossGroups: got true; want false")
	}
	if cfg.CrossGroupThresholdBoost != 0.05 {
		t.Errorf("CrossGroupThresholdBoost: got %v; want 0.05", cfg.CrossGroupThresholdBoost)
	}
	if cfg.CompareIdenticalAcrossGroups {
		t.Errorf("CompareIdenticalAcrossGroups: got true; want false")
	}
	if cfg.SuppressedPairs != nil {
		t.Errorf("SuppressedPairs: got %v; want nil", cfg.SuppressedPairs)
	}
}

// TestWarning_ScoresTypeIsAlgoID verifies the SPEC OVERRIDE (Phase 9)
// for Warning.Scores: the field type must be
// map[fuzzymatch.AlgoID]float64 (typed enum keys), NOT the spec's
// original map[string]float64. This is the load-bearing assertion that
// downstream Plan 09-03 (Check body) and Plan 09-06 (golden file
// writer) build against. Any unintended drift to map[string]float64
// here would silently change consumer code's compile-time key safety.
//
// See 09-CONTEXT.md §1 D-01 for the rationale and the
// api-ergonomics-reviewer sign-off recorded in this plan's PR.
func TestWarning_ScoresTypeIsAlgoID(t *testing.T) {
	t.Parallel()

	w := scan.Warning{}
	got := reflect.TypeOf(w.Scores).String()
	const want = "map[fuzzymatch.AlgoID]float64"
	if got != want {
		t.Errorf("Warning.Scores type: got %q; want %q (SPEC OVERRIDE D-01)", got, want)
	}
}
