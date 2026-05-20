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

// validate.go declares the internal pre-flight validation pipeline that
// scan.Check invokes as its first step (wired in Plan 09-03). The
// pipeline implements the Phase 8.5 Q2 "parameters strict, comparison
// data lenient" framework adapted to the scan sub-package: input
// configuration errors fail fast, while the items slice and the
// SuppressedPairs list collect every offending entry via errors.Join
// so callers fix a whole batch in one round-trip.
//
// Pipeline order (LOCKED per 09-CONTEXT.md §2):
//
//   - P1 — nil-Scorer fail-fast. Cheapest check; returns ErrNilScorer
//     directly (sentinel identity, never wrapped).
//   - P2 — Config field validation, fail-fast. Currently only
//     CrossGroupThresholdBoost (D-04: NaN / ±Inf / < 0 / > 1).
//   - P3 — Items[] collect-all via errors.Join. Two rules: D-03 empty
//     Item.Name, D-06 duplicate (Name, Group).
//   - P4 — SuppressedPairs collect-all via errors.Join. One rule:
//     D-05 empty string on either side of the pair. Self-pairs
//     (p[0] == p[1]) are silently kept (D-05 explicit semantics).
//
// Fail-fast BETWEEN phases (P1 returns before P2 runs, etc.);
// collect-all WITHIN each multi-error phase (P3 walks every Item; P4
// walks every SuppressedPairs entry).
//
// Implementation discipline (mirrors scorer.go:130-200 NewScorer):
//
//   - Pure function — no goroutines, no I/O, no package-global state.
//   - No map iteration on output paths (per
//     .claude/skills/determinism-standards). The duplicate-detection
//     map seen[itemKey]int is read via direct lookup; the for-loop
//     walks items in slice order, so emitted errors land in
//     ascending-index order by construction. No post-sort step needed.
//   - errors.Join is the Go 1.20+ multi-error type. errors.Is on the
//     joined value walks Unwrap() []error so callers can still
//     discriminate against the wrapped sentinel.
//   - Item.Tag is never stringified in any error message — only the
//     int index appears, mitigating T-09-02-02 (consumer-supplied Tag
//     leakage through the error surface).

package scan

import (
	"errors"
	"fmt"
	"math"
)

// itemKey is the dedup key for D-06 duplicate-detection. The pair
// (Name, Group) is compared with == — both fields are plain strings so
// equality is byte-identical without any normalisation. Validation
// uses the raw values from Item; the canonicalisation that Check
// performs later (Plan 09-03 onwards) acts on a separate normalised
// view of the same data.
type itemKey struct {
	Name  string
	Group string
}

// validateCheck runs the locked P1..P4 pre-flight pipeline against the
// items slice and Config struct. Returns nil on valid input.
//
// Phase order is fail-fast (the first phase to produce a non-nil error
// returns immediately); within a multi-error phase (P3, P4) every
// offending entry is collected via errors.Join.
//
// Return values:
//
//   - P1: ErrNilScorer (sentinel identity, never wrapped).
//   - P2: errors.Is(err, ErrInvalidConfig) is true; the offending
//     value is included in the message.
//   - P3: errors.Is(err, ErrInvalidItem) is true; the joined value
//     wraps one fmt.Errorf per offending index.
//   - P4: errors.Is(err, ErrInvalidConfig) is true; the joined value
//     wraps one fmt.Errorf per offending SuppressedPairs index.
//
// Consumers discriminate via errors.Is — never via string matching.
func validateCheck(items []Item, cfg Config) error {
	// P1 — nil-Scorer fail-fast. Return the sentinel directly (no wrap)
	// so consumers can compare with `err == scan.ErrNilScorer` or with
	// errors.Is — both work. The sentinel is the cheapest possible
	// signal and reflects the foundation contract pinned in Plan 09-01.
	if cfg.Scorer == nil {
		return ErrNilScorer
	}

	// P2 — Config field validation, fail-fast. Currently only
	// CrossGroupThresholdBoost; future Config fields land here in
	// declaration order, each with its own fail-fast check.
	if err := validateConfigFields(cfg); err != nil {
		return err
	}

	// P3 — Items[] collect-all. Walks every item, accumulating
	// ErrInvalidItem wraps for D-03 and D-06 violations.
	if err := validateItems(items); err != nil {
		return err
	}

	// P4 — SuppressedPairs collect-all. Walks every pair, accumulating
	// ErrInvalidConfig wraps for D-05 violations. Reached only when
	// P1..P3 all pass; combined with the SuppressedPairs build cost in
	// Plan 09-05 the consumer pays O(N) once at Check entry.
	if err := validateSuppressedPairs(cfg.SuppressedPairs); err != nil {
		return err
	}

	return nil
}

// validateConfigFields implements P2 — fail-fast field validation for
// the Config struct. Returns nil when every field is in range.
//
// Currently checks CrossGroupThresholdBoost (D-04): NaN, ±Inf, and
// values outside the closed interval [0.0, 1.0] are all rejected. The
// closed-interval bounds (0.0 and 1.0) are accepted — useful for
// disabling the boost (0.0) or clamping cross-group emission to
// byte-identical pairs combined with CompareIdenticalAcrossGroups
// (1.0).
//
// math.IsInf(b, 0) covers both +Inf and -Inf in a single call. The
// ordered evaluation (NaN check → ±Inf check → range bounds) is
// deliberate: NaN compares false against every range bound, so a
// naive `b < 0 || b > 1` check would silently miss NaN. The NaN-first
// gate is the canonical Go float-validation idiom.
func validateConfigFields(cfg Config) error {
	b := cfg.CrossGroupThresholdBoost
	if math.IsNaN(b) || math.IsInf(b, 0) || b < 0.0 || b > 1.0 {
		return fmt.Errorf(
			"%w: CrossGroupThresholdBoost=%v is invalid (must be in [0.0, 1.0], finite, non-NaN)",
			ErrInvalidConfig, b,
		)
	}
	return nil
}

// validateItems implements P3 — collect-all Items[] validation. Walks
// items in slice order, accumulating one ErrInvalidItem wrap per
// offending index, and returns errors.Join(errs...) (which is nil
// when errs is empty).
//
// Two validation rules:
//
//   - D-03: Item.Name == "" — the empty string is meaningless for
//     similarity scoring; every algorithm short-circuits to its
//     empty-vs-* contract and would emit warnings that say nothing
//     useful.
//   - D-06: duplicate (Name, Group) — the in-line completeness
//     assertion in Plan 09-06 panics on a duplicate sort key, and
//     the sort key derives from (Kind, NameA, NameB, GroupA, GroupB).
//     Two items with the same (Name, Group) at different slice indices
//     would produce two warnings with identical sort keys; rejecting
//     them at the door turns that runtime panic into a clean
//     ErrInvalidItem the consumer can fix upstream.
//
// Ordering: items[] is walked in slice order, so the emitted errors
// land in ascending-index order by construction — no post-sort step
// needed. The seen[itemKey]int map is read via direct lookup
// (`seen[k]`) only; the loop never iterates the map, so no map-iter
// non-determinism reaches the output path (per
// .claude/skills/determinism-standards).
//
// D-03 fires before D-06 within a single item — an empty Name skips
// the duplicate check (the `continue` after the D-03 emission) so an
// empty-Name item never registers in seen[].
//
// Cost: O(N) time, O(N) space for the seen map. The map is bounded by
// len(items) and discarded at function exit; T-09-02-01 (DoS via
// huge items) is accepted per CONTEXT.md threat model (consumer
// responsibility per spec §12.6).
//
// Item.Tag is never referenced in any error message — only the int
// index appears (T-09-02-02 mitigation).
func validateItems(items []Item) error {
	var errs []error
	seen := make(map[itemKey]int, len(items))

	for i, item := range items {
		// D-03 fires first; an empty Name skips the duplicate check
		// so the empty value never lands in seen[].
		if item.Name == "" {
			errs = append(errs, fmt.Errorf(
				"%w: index %d: empty name",
				ErrInvalidItem, i,
			))
			continue
		}

		// D-06: duplicate (Name, Group) detection. The first-seen index
		// is reported so the consumer can correlate the two rows.
		k := itemKey{Name: item.Name, Group: item.Group}
		if first, ok := seen[k]; ok {
			errs = append(errs, fmt.Errorf(
				"%w: index %d: duplicate (Name, Group) of index %d",
				ErrInvalidItem, i, first,
			))
			continue
		}
		seen[k] = i
	}

	// errors.Join(errs...) returns nil when errs is empty, so the
	// happy path returns nil naturally — no explicit len-check needed.
	return errors.Join(errs...)
}

// validateSuppressedPairs implements P4 — collect-all SuppressedPairs
// validation. Walks pairs in slice order, accumulating one
// ErrInvalidConfig wrap per offending index.
//
// Single rule (D-05): either side of the [2]string equal to "" makes
// the entry meaningless — Check could neither store the entry in the
// suppression map (Plan 09-05 builds the map at Check entry) nor
// compare it against any candidate pair (D-03 already rejects empty
// Item.Name).
//
// Self-pair semantics (D-05 explicit): pairs where p[0] == p[1] are
// silently kept. They are harmless because Check never emits a
// self-warning (a Warning's NameA and NameB always come from two
// distinct Items, and D-06 guarantees no two Items share a sort key).
// Allowing self-pairs lets consumers programmatically build the
// SuppressedPairs list without filtering for the rare self-equality
// case.
//
// Ordering: pairs[] is walked in slice order, so emitted errors land
// in ascending-index order by construction — no post-sort step
// needed.
//
// Cost: O(N) time, O(N) space (one wrap per offending entry). No map
// is allocated.
func validateSuppressedPairs(pairs [][2]string) error {
	var errs []error
	for i, p := range pairs {
		if p[0] == "" || p[1] == "" {
			errs = append(errs, fmt.Errorf(
				"%w: SuppressedPairs[%d]: empty string in pair",
				ErrInvalidConfig, i,
			))
		}
	}
	return errors.Join(errs...)
}
