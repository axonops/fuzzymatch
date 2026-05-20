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
// validate_test.go pins the Plan 09-02 validation pipeline (P1..P4) for
// the scan sub-package's pre-flight surface. The 22 Test* functions
// cover every locked behaviour declared in 09-CONTEXT.md §2:
//
//   - P1 (fail-fast): cfg.Scorer == nil returns ErrNilScorer directly
//     (sentinel identity, not wrap).
//   - P2 (fail-fast): cfg.CrossGroupThresholdBoost NaN / ±Inf / < 0 / > 1
//     returns ErrInvalidConfig wrapped with the offending value.
//   - P3 (collect-all via errors.Join): every empty Item.Name (D-03) and
//     every duplicate (Name, Group) (D-06) wrapped into an
//     ErrInvalidItem chain.
//   - P4 (collect-all via errors.Join): every empty SuppressedPairs
//     entry (D-05) wrapped into an ErrInvalidConfig chain. Self-pairs
//     are silently kept.
//   - Fail-fast ordering between phases (P1 → P2 → P3 → P4); collect-all
//     within a phase.
//
// Stdlib `testing` only — no testify in root tests per CLAUDE.md and
// .claude/skills/go-coding-standards. This file lives in `package scan`
// (internal) because validateCheck is unexported.

package scan

import (
	"errors"
	"math"
	"strings"
	"testing"

	"github.com/axonops/fuzzymatch"
)

// newGoodScorer returns the opinionated DefaultScorer for tests that
// need a valid non-nil Scorer pointer. The Scorer is never invoked by
// validateCheck (validation is pure-data); the pointer's non-nil-ness
// is what P1 checks.
func newGoodScorer() *fuzzymatch.Scorer {
	return fuzzymatch.DefaultScorer()
}

// newGoodConfig returns a valid Config bound to the supplied Scorer
// with the opinionated defaults from scan.DefaultConfig.
func newGoodConfig(s *fuzzymatch.Scorer) Config {
	return DefaultConfig(s)
}

// countWrapped walks err via the Go 1.20+ Unwrap() []error interface
// and returns how many leaves match errors.Is(_, target). Used by the
// "collected N offending indices" assertions.
func countWrapped(err error, target error) int {
	if err == nil {
		return 0
	}
	type multi interface{ Unwrap() []error }
	if u, ok := err.(multi); ok {
		count := 0
		for _, w := range u.Unwrap() {
			if errors.Is(w, target) {
				count++
			}
		}
		return count
	}
	// Single non-joined wrapped error: count the chain via errors.Is on
	// the whole error.
	if errors.Is(err, target) {
		return 1
	}
	return 0
}

// --- P1: nil Scorer fail-fast ---------------------------------------------

// TestValidate_NilScorer_ReturnsErrNilScorer pins P1: when cfg.Scorer
// is nil, validateCheck returns the ErrNilScorer sentinel directly
// (not a wrapped value). The sentinel-identity comparison is the
// contract documented in scan/errors.go.
func TestValidate_NilScorer_ReturnsErrNilScorer(t *testing.T) {
	t.Parallel()

	cfg := Config{Scorer: nil}
	err := validateCheck(nil, cfg)
	if err != ErrNilScorer { //nolint:errorlint // sentinel-identity contract: validateCheck returns ErrNilScorer directly, not wrapped
		t.Fatalf("want ErrNilScorer (sentinel identity), got %v", err)
	}
}

// --- P2: CrossGroupThresholdBoost field validation -------------------------

// TestValidate_NaNBoost_ReturnsErrInvalidConfig pins P2 D-04: NaN boost
// is rejected with ErrInvalidConfig.
func TestValidate_NaNBoost_ReturnsErrInvalidConfig(t *testing.T) {
	t.Parallel()

	cfg := newGoodConfig(newGoodScorer())
	cfg.CrossGroupThresholdBoost = math.NaN()

	err := validateCheck(nil, cfg)
	if !errors.Is(err, ErrInvalidConfig) {
		t.Fatalf("want errors.Is(ErrInvalidConfig), got %v", err)
	}
	if !strings.Contains(err.Error(), "NaN") {
		t.Errorf("want error message to contain %q; got %q", "NaN", err.Error())
	}
}

// TestValidate_PosInfBoost_ReturnsErrInvalidConfig pins P2 D-04: +Inf
// boost is rejected.
func TestValidate_PosInfBoost_ReturnsErrInvalidConfig(t *testing.T) {
	t.Parallel()

	cfg := newGoodConfig(newGoodScorer())
	cfg.CrossGroupThresholdBoost = math.Inf(+1)

	err := validateCheck(nil, cfg)
	if !errors.Is(err, ErrInvalidConfig) {
		t.Fatalf("want errors.Is(ErrInvalidConfig), got %v", err)
	}
}

// TestValidate_NegInfBoost_ReturnsErrInvalidConfig pins P2 D-04: -Inf
// boost is rejected.
func TestValidate_NegInfBoost_ReturnsErrInvalidConfig(t *testing.T) {
	t.Parallel()

	cfg := newGoodConfig(newGoodScorer())
	cfg.CrossGroupThresholdBoost = math.Inf(-1)

	err := validateCheck(nil, cfg)
	if !errors.Is(err, ErrInvalidConfig) {
		t.Fatalf("want errors.Is(ErrInvalidConfig), got %v", err)
	}
}

// TestValidate_NegativeBoost_ReturnsErrInvalidConfig pins P2 D-04: any
// boost < 0 is rejected (strictly outside the [0, 1] closed range).
func TestValidate_NegativeBoost_ReturnsErrInvalidConfig(t *testing.T) {
	t.Parallel()

	cfg := newGoodConfig(newGoodScorer())
	cfg.CrossGroupThresholdBoost = -0.1

	err := validateCheck(nil, cfg)
	if !errors.Is(err, ErrInvalidConfig) {
		t.Fatalf("want errors.Is(ErrInvalidConfig), got %v", err)
	}
}

// TestValidate_OverflowBoost_ReturnsErrInvalidConfig pins P2 D-04: any
// boost > 1 is rejected.
func TestValidate_OverflowBoost_ReturnsErrInvalidConfig(t *testing.T) {
	t.Parallel()

	cfg := newGoodConfig(newGoodScorer())
	cfg.CrossGroupThresholdBoost = 1.5

	err := validateCheck(nil, cfg)
	if !errors.Is(err, ErrInvalidConfig) {
		t.Fatalf("want errors.Is(ErrInvalidConfig), got %v", err)
	}
}

// TestValidate_ZeroBoost_OK pins P2 D-04: boost = 0.0 is the lower
// closed-interval bound and must be accepted.
func TestValidate_ZeroBoost_OK(t *testing.T) {
	t.Parallel()

	cfg := newGoodConfig(newGoodScorer())
	cfg.CrossGroupThresholdBoost = 0.0

	err := validateCheck(nil, cfg)
	if err != nil {
		t.Fatalf("boost=0.0 should be valid, got %v", err)
	}
}

// TestValidate_UnitBoost_OK pins P2 D-04: boost = 1.0 is the upper
// closed-interval bound and must be accepted.
func TestValidate_UnitBoost_OK(t *testing.T) {
	t.Parallel()

	cfg := newGoodConfig(newGoodScorer())
	cfg.CrossGroupThresholdBoost = 1.0

	err := validateCheck(nil, cfg)
	if err != nil {
		t.Fatalf("boost=1.0 should be valid, got %v", err)
	}
}

// --- P3: Items[] validation -----------------------------------------------

// TestValidate_SingleEmptyName_CollectsViaJoin pins P3 D-03: a single
// empty-Name item produces an ErrInvalidItem wrap whose message
// includes the offending index.
func TestValidate_SingleEmptyName_CollectsViaJoin(t *testing.T) {
	t.Parallel()

	items := []Item{
		{Name: "a"},
		{Name: "b"},
		{Name: "c"},
		{Name: ""},
		{Name: "e"},
	}
	cfg := newGoodConfig(newGoodScorer())

	err := validateCheck(items, cfg)
	if !errors.Is(err, ErrInvalidItem) {
		t.Fatalf("want errors.Is(ErrInvalidItem), got %v", err)
	}
	if !strings.Contains(err.Error(), "index 3") {
		t.Errorf("want error message to contain %q; got %q", "index 3", err.Error())
	}
}

// TestValidate_MultipleEmptyNames_CollectsAllViaJoin pins P3 D-03: every
// empty-Name index is collected via errors.Join. errors.Is still
// discriminates because Unwrap() []error is walked by the stdlib.
func TestValidate_MultipleEmptyNames_CollectsAllViaJoin(t *testing.T) {
	t.Parallel()

	items := []Item{
		{Name: "zero"},
		{Name: ""},
		{Name: "two"},
		{Name: "three"},
		{Name: ""},
		{Name: "five"},
		{Name: "six"},
		{Name: ""},
	}
	cfg := newGoodConfig(newGoodScorer())

	err := validateCheck(items, cfg)
	if !errors.Is(err, ErrInvalidItem) {
		t.Fatalf("want errors.Is(ErrInvalidItem), got %v", err)
	}

	if got, want := countWrapped(err, ErrInvalidItem), 3; got != want {
		t.Errorf("want %d wrapped ErrInvalidItem leaves, got %d (err=%v)", want, got, err)
	}

	for _, idx := range []string{"index 1", "index 4", "index 7"} {
		if !strings.Contains(err.Error(), idx) {
			t.Errorf("want error message to mention %q; got %q", idx, err.Error())
		}
	}
}

// TestValidate_DuplicateNameGroup_CollectsViaJoin pins P3 D-06: a
// duplicate (Name, Group) entry is rejected with the first-seen index
// referenced in the message.
func TestValidate_DuplicateNameGroup_CollectsViaJoin(t *testing.T) {
	t.Parallel()

	items := []Item{
		{Name: "user_id", Group: "login"},
		{Name: "email", Group: "login"},
		{Name: "user_id", Group: "login"}, // duplicate of index 0
	}
	cfg := newGoodConfig(newGoodScorer())

	err := validateCheck(items, cfg)
	if !errors.Is(err, ErrInvalidItem) {
		t.Fatalf("want errors.Is(ErrInvalidItem), got %v", err)
	}
	if !strings.Contains(err.Error(), "index 2") {
		t.Errorf("want error message to mention %q; got %q", "index 2", err.Error())
	}
	if !strings.Contains(err.Error(), "of index 0") {
		t.Errorf("want error message to mention %q; got %q", "of index 0", err.Error())
	}
}

// TestValidate_DuplicatePlusEmptyName_CollectsBoth pins the mixed
// D-03 + D-06 path: both kinds of error appear in the joined output.
func TestValidate_DuplicatePlusEmptyName_CollectsBoth(t *testing.T) {
	t.Parallel()

	items := []Item{
		{Name: "alpha", Group: "g"},
		{Name: "", Group: "g"},        // D-03 empty Name at index 1
		{Name: "alpha", Group: "g"},   // D-06 duplicate of index 0
	}
	cfg := newGoodConfig(newGoodScorer())

	err := validateCheck(items, cfg)
	if !errors.Is(err, ErrInvalidItem) {
		t.Fatalf("want errors.Is(ErrInvalidItem), got %v", err)
	}

	if got, want := countWrapped(err, ErrInvalidItem), 2; got != want {
		t.Errorf("want %d wrapped ErrInvalidItem leaves, got %d (err=%v)", want, got, err)
	}

	if !strings.Contains(err.Error(), "index 1") {
		t.Errorf("want error message to mention empty-Name index %q; got %q", "index 1", err.Error())
	}
	if !strings.Contains(err.Error(), "index 2") {
		t.Errorf("want error message to mention duplicate index %q; got %q", "index 2", err.Error())
	}
	if !strings.Contains(err.Error(), "of index 0") {
		t.Errorf("want error message to mention first-seen %q; got %q", "of index 0", err.Error())
	}
}

// TestValidate_DifferentGroups_NoDuplicate pins P3 D-06: the (Name,
// Group) tuple is the dedup key. Same Name with different Group is
// legal.
func TestValidate_DifferentGroups_NoDuplicate(t *testing.T) {
	t.Parallel()

	items := []Item{
		{Name: "user_id", Group: "login"},
		{Name: "user_id", Group: "profile"},
	}
	cfg := newGoodConfig(newGoodScorer())

	err := validateCheck(items, cfg)
	if err != nil {
		t.Fatalf("same Name across different Groups should be valid, got %v", err)
	}
}

// --- P4: SuppressedPairs validation ---------------------------------------

// TestValidate_EmptySuppressedPair_CollectsViaJoin pins P4 D-05: a
// single empty-side pair is rejected with the offending index.
func TestValidate_EmptySuppressedPair_CollectsViaJoin(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name  string
		pairs [][2]string
		want  string
	}{
		{
			name:  "empty-left-side",
			pairs: [][2]string{{"a", "b"}, {"", "x"}},
			want:  "SuppressedPairs[1]",
		},
		{
			name:  "empty-right-side",
			pairs: [][2]string{{"a", "b"}, {"x", ""}},
			want:  "SuppressedPairs[1]",
		},
		{
			name:  "both-sides-empty",
			pairs: [][2]string{{"", ""}},
			want:  "SuppressedPairs[0]",
		},
	}

	for _, c := range cases {
		c := c
		t.Run(c.name, func(t *testing.T) {
			t.Parallel()
			cfg := newGoodConfig(newGoodScorer())
			cfg.SuppressedPairs = c.pairs

			err := validateCheck(nil, cfg)
			if !errors.Is(err, ErrInvalidConfig) {
				t.Fatalf("want errors.Is(ErrInvalidConfig), got %v", err)
			}
			if !strings.Contains(err.Error(), c.want) {
				t.Errorf("want error message to mention %q; got %q", c.want, err.Error())
			}
		})
	}
}

// TestValidate_MultipleEmptySuppressedPairs_CollectsAll pins P4 D-05:
// every offending SuppressedPairs index is collected via errors.Join.
func TestValidate_MultipleEmptySuppressedPairs_CollectsAll(t *testing.T) {
	t.Parallel()

	cfg := newGoodConfig(newGoodScorer())
	cfg.SuppressedPairs = [][2]string{
		{"a", "b"},
		{"", "x"},   // index 1
		{"c", "d"},
		{"x", ""},   // index 3
		{"e", "f"},
		{"", ""},    // index 5
	}

	err := validateCheck(nil, cfg)
	if !errors.Is(err, ErrInvalidConfig) {
		t.Fatalf("want errors.Is(ErrInvalidConfig), got %v", err)
	}

	if got, want := countWrapped(err, ErrInvalidConfig), 3; got != want {
		t.Errorf("want %d wrapped ErrInvalidConfig leaves, got %d (err=%v)", want, got, err)
	}

	for _, idx := range []string{"SuppressedPairs[1]", "SuppressedPairs[3]", "SuppressedPairs[5]"} {
		if !strings.Contains(err.Error(), idx) {
			t.Errorf("want error message to mention %q; got %q", idx, err.Error())
		}
	}
}

// TestValidate_SelfPairInSuppressed_OK pins D-05's explicit semantics:
// self-pairs (a == b) are valid and silently kept. They are harmless
// because Check never emits a self-warning.
func TestValidate_SelfPairInSuppressed_OK(t *testing.T) {
	t.Parallel()

	cfg := newGoodConfig(newGoodScorer())
	cfg.SuppressedPairs = [][2]string{{"abc", "abc"}}

	err := validateCheck(nil, cfg)
	if err != nil {
		t.Fatalf("self-pair must be silently kept (D-05), got %v", err)
	}
}

// --- Fail-fast ordering between phases ------------------------------------

// TestValidate_FailFastOrder_NilScorerBeforeItemsCheck pins the
// pipeline order P1 → P3: a nil Scorer combined with an invalid Item
// still returns ErrNilScorer. Items validation never runs.
func TestValidate_FailFastOrder_NilScorerBeforeItemsCheck(t *testing.T) {
	t.Parallel()

	items := []Item{{Name: ""}}
	cfg := Config{Scorer: nil}

	err := validateCheck(items, cfg)
	if err != ErrNilScorer { //nolint:errorlint // sentinel-identity contract: P1 fires first and returns ErrNilScorer directly
		t.Fatalf("P1 should fire first; want ErrNilScorer (sentinel identity), got %v", err)
	}
	if errors.Is(err, ErrInvalidItem) {
		t.Errorf("Items validation must not run when Scorer is nil; got err=%v", err)
	}
}

// TestValidate_FailFastOrder_BoostBeforeItemsCheck pins the pipeline
// order P2 → P3: an invalid boost combined with an invalid Item
// returns the boost error. Items validation never runs.
func TestValidate_FailFastOrder_BoostBeforeItemsCheck(t *testing.T) {
	t.Parallel()

	items := []Item{{Name: ""}}
	cfg := newGoodConfig(newGoodScorer())
	cfg.CrossGroupThresholdBoost = math.NaN()

	err := validateCheck(items, cfg)
	if !errors.Is(err, ErrInvalidConfig) {
		t.Fatalf("P2 should fire before P3; want errors.Is(ErrInvalidConfig), got %v", err)
	}
	if errors.Is(err, ErrInvalidItem) {
		t.Errorf("Items validation must not run when boost is invalid; got err=%v", err)
	}
}

// TestValidate_FailFastOrder_ItemsBeforeSuppressedPairs pins the
// pipeline order P3 → P4: invalid Items combined with invalid
// SuppressedPairs returns the Items errors only. SuppressedPairs
// validation never runs.
//
// Fail-fast BETWEEN phases; collect-all WITHIN a phase.
func TestValidate_FailFastOrder_ItemsBeforeSuppressedPairs(t *testing.T) {
	t.Parallel()

	items := []Item{{Name: ""}}
	cfg := newGoodConfig(newGoodScorer())
	cfg.SuppressedPairs = [][2]string{{"", "x"}}

	err := validateCheck(items, cfg)
	if !errors.Is(err, ErrInvalidItem) {
		t.Fatalf("P3 should fire before P4; want errors.Is(ErrInvalidItem), got %v", err)
	}
	if errors.Is(err, ErrInvalidConfig) {
		t.Errorf("SuppressedPairs validation must not run when Items are invalid; got err=%v", err)
	}
	if strings.Contains(err.Error(), "SuppressedPairs") {
		t.Errorf("SuppressedPairs validation must not run when Items are invalid; got err=%v", err)
	}
}

// --- Happy paths ----------------------------------------------------------

// TestValidate_EmptyItems_NoError pins the foundation contract: an
// empty items slice with a valid Config is itself valid input (Check
// will return an empty Warning slice). The P1+P2 fail-fasts pass with
// a non-nil Scorer and in-range boost; P3 has nothing to walk; P4
// likewise.
func TestValidate_EmptyItems_NoError(t *testing.T) {
	t.Parallel()

	cfg := newGoodConfig(newGoodScorer())
	err := validateCheck(nil, cfg)
	if err != nil {
		t.Fatalf("empty items + default config should be valid, got %v", err)
	}

	err = validateCheck([]Item{}, cfg)
	if err != nil {
		t.Fatalf("zero-length items slice + default config should be valid, got %v", err)
	}
}

// TestValidate_GoodInput_NoError pins the typical happy-path input: a
// valid Scorer, in-range boost, distinct non-empty Items, non-empty
// SuppressedPairs entries, and a self-pair (which is silently kept).
func TestValidate_GoodInput_NoError(t *testing.T) {
	t.Parallel()

	items := []Item{
		{Name: "user_id", Group: "login"},
		{Name: "user_id", Group: "profile"},
		{Name: "email", Group: "login", Tag: 42},
		{Name: "session_id", Group: "login"},
	}
	cfg := newGoodConfig(newGoodScorer())
	cfg.SuppressedPairs = [][2]string{
		{"foo", "bar"},
		{"abc", "abc"}, // self-pair, silently kept
	}

	err := validateCheck(items, cfg)
	if err != nil {
		t.Fatalf("typical valid input should produce no error, got %v", err)
	}
}

// TestValidate_JoinedError_DiscriminatesAllSentinels pins the
// errors.Join discrimination contract: a P3 joined error containing
// multiple ErrInvalidItem leaves does not falsely match
// ErrInvalidConfig (the P2/P4 sentinel); a P4 joined error containing
// multiple ErrInvalidConfig leaves does not falsely match
// ErrInvalidItem (the P3 sentinel); neither matches ErrNilScorer.
func TestValidate_JoinedError_DiscriminatesAllSentinels(t *testing.T) {
	t.Parallel()

	// P3 joined error
	items := []Item{
		{Name: ""},
		{Name: ""},
	}
	cfg := newGoodConfig(newGoodScorer())
	errP3 := validateCheck(items, cfg)
	if !errors.Is(errP3, ErrInvalidItem) {
		t.Fatalf("P3 path: want errors.Is(ErrInvalidItem), got %v", errP3)
	}
	if errors.Is(errP3, ErrInvalidConfig) {
		t.Errorf("P3 path: ErrInvalidConfig must not match ErrInvalidItem-only chain; got err=%v", errP3)
	}
	if errors.Is(errP3, ErrNilScorer) {
		t.Errorf("P3 path: ErrNilScorer must not match; got err=%v", errP3)
	}

	// P4 joined error
	cfg2 := newGoodConfig(newGoodScorer())
	cfg2.SuppressedPairs = [][2]string{{"", "x"}, {"y", ""}}
	errP4 := validateCheck(nil, cfg2)
	if !errors.Is(errP4, ErrInvalidConfig) {
		t.Fatalf("P4 path: want errors.Is(ErrInvalidConfig), got %v", errP4)
	}
	if errors.Is(errP4, ErrInvalidItem) {
		t.Errorf("P4 path: ErrInvalidItem must not match ErrInvalidConfig-only chain; got err=%v", errP4)
	}
	if errors.Is(errP4, ErrNilScorer) {
		t.Errorf("P4 path: ErrNilScorer must not match; got err=%v", errP4)
	}
}
