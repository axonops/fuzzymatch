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

// Package-level sentinel errors for the scan sub-package. All errors
// are wrappable via fmt.Errorf("...: %w", ErrX) and discoverable by
// errors.Is / errors.As — never via string matching. See
// docs/requirements.md §12.1 for the canonical sentinel set and
// 09-CONTEXT.md §2 for the data-vs-parameter validation framework
// (D-03 / D-04 / D-05 / D-06) applied here.
//
// Error message convention (per .claude/skills/go-coding-standards,
// adapted from the root errors.go file-header):
//
//   - Every message starts with the "scan: " package prefix so
//     wrappers like fmt.Errorf("check failed: %w", err) produce
//     readable compositions ("check failed: scan: invalid Item").
//   - The text after the prefix is lowercase and carries no trailing
//     punctuation ('.', '!', or '?') so concatenation flows cleanly.
//   - Each sentinel is a flat errors.New value, not a typed struct;
//     callers receive index information via the wrapping fmt.Errorf
//     call site (see Check / validateCheck — Plan 09-02).
//   - When validation collects multiple offending indices (D-03,
//     D-05, D-06), the returned error is errors.Join of multiple
//     wrapped values; errors.Is still discriminates because errors.Is
//     walks Unwrap() []error (Go 1.20+).

package scan

import "errors"

// ErrNilScorer indicates Config.Scorer was nil when Check was
// invoked. A nil Scorer has no algorithms, no threshold, and no
// normalisation options — there is nothing for Check to do.
//
// Common causes: forgetting to set Config.Scorer before calling
// Check; passing a freshly-allocated zero-value Config struct;
// dereferencing a Scorer pointer that NewScorer returned alongside a
// non-nil error.
//
// Resolution: set cfg.Scorer to a non-nil *fuzzymatch.Scorer
// (typically from fuzzymatch.DefaultScorer() or
// fuzzymatch.NewScorer(...)). For the opinionated default Config +
// Scorer composition, use
// scan.DefaultConfig(fuzzymatch.DefaultScorer()).
//
// Example:
//
//	_, err := scan.Check(items, scan.Config{Scorer: nil})
//	if errors.Is(err, scan.ErrNilScorer) {
//	    // diagnostic
//	}
var ErrNilScorer = errors.New("scan: Config.Scorer is required")

// ErrInvalidItem indicates one or more entries in the items slice
// passed to Check failed validation. The two locked validation rules
// (per 09-CONTEXT.md §2 D-03 and §3 D-06) are:
//
//   - Item.Name == "" (D-03: empty Name)
//   - Duplicate (Name, Group) of an earlier item in the same slice
//     (D-06: duplicates collide with the sort-key completeness
//     invariant)
//
// Common causes: source schema where Name is optional and the
// consumer forgot to filter empty rows; duplicate identifier rows in
// a Cassandra audit corpus where the same field appears across two
// schema versions.
//
// Resolution: filter empty Names upstream of Check, and deduplicate
// (Name, Group) tuples before calling Check. Each wrapped error
// carries the offending item's slice index in its message so the
// caller can pinpoint the failures without re-walking the slice.
//
// Multi-error semantics (D-03 + D-06): when validation discovers
// multiple invalid items, the returned error is errors.Join of every
// wrapped ErrInvalidItem value (one per offending index, in
// index-ascending order). errors.Is(err, scan.ErrInvalidItem) still
// returns true because errors.Is walks Unwrap() []error (Go 1.20+).
// errors.Is is the canonical discrimination; never match the message
// string.
//
// Example:
//
//	_, err := scan.Check(items, cfg)
//	if errors.Is(err, scan.ErrInvalidItem) {
//	    // one or more items failed validation; the wrapped error
//	    // chain identifies every offending index.
//	}
var ErrInvalidItem = errors.New("scan: invalid item")

// ErrInvalidConfig indicates the Config struct passed to Check failed
// validation. The locked validation rules (per 09-CONTEXT.md §2 D-04
// and D-05) are:
//
//   - Config.CrossGroupThresholdBoost is NaN, ±Inf, < 0, or > 1
//     (D-04: parameter strict-range)
//   - Config.SuppressedPairs contains a [2]string with one or both
//     entries equal to "" (D-05: empty entries collected via
//     errors.Join)
//
// Common causes: computing CrossGroupThresholdBoost from upstream
// arithmetic that produced NaN; deserialising SuppressedPairs from a
// data source where one column is nullable.
//
// Resolution: ensure CrossGroupThresholdBoost is in the closed
// interval [0.0, 1.0] (or use scan.DefaultConfig which bakes 0.05);
// filter empty SuppressedPairs entries before calling Check.
//
// Multi-error semantics (D-05): when SuppressedPairs validation
// discovers multiple offending entries, the returned error is
// errors.Join of every wrapped ErrInvalidConfig value (one per
// offending index, in index-ascending order).
// errors.Is(err, scan.ErrInvalidConfig) still returns true because
// errors.Is walks Unwrap() []error (Go 1.20+).
//
// Example:
//
//	cfg := scan.DefaultConfig(s)
//	cfg.CrossGroupThresholdBoost = math.NaN()
//	_, err := scan.Check(items, cfg)
//	if errors.Is(err, scan.ErrInvalidConfig) {
//	    // diagnostic
//	}
var ErrInvalidConfig = errors.New("scan: invalid config")
