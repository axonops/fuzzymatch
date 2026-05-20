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

// scan.go declares the Phase 9 Layer-3 collection-scan surface of the
// three-layer fuzzymatch architecture: Item, Config, Warning, Kind,
// Check, DefaultConfig. It consumes the Phase 8 *fuzzymatch.Scorer for
// per-pair similarity and orchestrates within/cross-group passes with
// token-bucket optimisation, deterministic sort, and suppression
// composition.
//
// SPEC OVERRIDE (Phase 9):
//
//   - Warning.Scores is map[fuzzymatch.AlgoID]float64 (NOT
//     map[string]float64 as in docs/requirements.md §12.1 line 1337).
//     Extends the Phase 8 ScoreAll override at §8.3 + §8.6
//     (08-CONTEXT.md §1) for the same typed-enum-keys rationale. See
//     09-CONTEXT.md §1 D-01.
//   - Kind type renamed from WarningKind (NOT scan.WarnKind to avoid
//     accidental symmetry with root's WarnKind). See 09-CONTEXT.md §1
//     D-02. Spec §12.1 renamed in lockstep.
//
// api-ergonomics-reviewer signed off on both overrides in plan 09-01's
// PR, mirroring the Phase 7 Plan 07-02 and Phase 8 Plan 08-03 precedent
// of recording the reviewer verdict in the PR description.
//
// Design notes (per 09-CONTEXT.md + 09-RESEARCH.md):
//
//   - Plan 09-01 lands the foundation skeleton: Item / Config /
//     Warning struct declarations, the three sentinel errors, the
//     DefaultConfig opinionated helper, and the Check stub that returns
//     (nil, ErrNilScorer) when cfg.Scorer is nil. Plan 09-02 adds the
//     full validation pipeline. Plan 09-03 lands the Check body with
//     the naive within-group + cross-group passes and the SCAN-04
//     identical-name suppression default (inline at the cross-group
//     emission site). Plan 09-04 wires in the token-bucket optimisation
//     (bucketThreshold = 50, private const). Plan 09-05 lands the full
//     suppression composition (SilenceLint + SuppressedPairs + Rule 3)
//     in scan/suppress.go and routes both naive and bucket emission
//     paths through the isSuppressed predicate; the Plan-09-03 inline
//     identical-name check is migrated into Rule 3 for unified
//     semantics. Plan 09-06 adds the deterministic sort and in-line
//     completeness assertion.
//
//   - Validation pipeline order (P1..P4) is LOCKED (09-CONTEXT.md §2):
//     nil-Scorer fail-fast → Config field validation → Items validation
//     (D-03 + D-06 collect-all via errors.Join) → SuppressedPairs
//     validation (D-05 collect-all).
//
//   - The Scorer's normalisation options are accessed via the
//     Scorer.NormalisationOptions() public method introduced in Plan
//     09-01 (resolves 09-RESEARCH.md Open Question 1). Plan 09-04
//     consumes the accessor when building token buckets; Plan 09-05
//     consumes it when canonicalising SuppressedPairs (Pitfall 4).
//
//   - In-line completeness assertion (added in Plan 09-06) panics with
//     fuzzymatch.ErrInternalInvariantViolated (Phase 8.5 Gap 5) when
//     two adjacent post-sort warnings share the (Kind, NameA, NameB,
//     GroupA, GroupB) sort key — that would only occur via a library
//     bug, never via caller input, because D-06 rejects duplicate
//     (Name, Group) at validation time.
//
//   - No goroutines, no channels, no mutexes. Pure-function library.

package scan

import (
	"fmt"
	"math"
	"sort"

	"github.com/axonops/fuzzymatch"
)

// Item is one named thing the scanner compares. Consumers construct
// items from whatever schema or vocabulary they care about.
//
// Concurrency: Item is a plain data struct with no methods and no
// internal synchronisation. Consumers are responsible for not mutating
// an Item that scan.Check is currently reading.
type Item struct {
	// Name is the value being compared. Required, non-empty — Check
	// returns ErrInvalidItem (wrapped with the offending index) when
	// Name is the empty string.
	Name string

	// Group scopes the comparison. Items with the same Group are
	// compared against each other in the within-group pass. Items with
	// different Group values are compared in the cross-group pass when
	// Config.CompareAcrossGroups is true. The empty string is a valid
	// Group (a single global group).
	Group string

	// SilenceLint, when true, suppresses any warning involving this
	// Item. The suppression is one-sided: setting the flag on either
	// side of a pair silences the pair.
	SilenceLint bool

	// Tag is opaque consumer data. The library does not interpret it;
	// it appears unchanged in Warning.TagA / Warning.TagB. Use it to
	// carry source-file line numbers, schema paths, or any other
	// context useful in diagnostics.
	//
	// Security note: Tag is never stringified inside error messages or
	// panic values. Only the int index of an offending Item appears in
	// validation errors. Consumers may safely store sensitive data in
	// Tag without risking leakage through the error surface.
	Tag any
}

// Config controls Check behaviour. Construct directly with a non-nil
// Scorer, or use DefaultConfig for the opinionated default.
//
// The zero value of Config is invalid: Scorer is nil, so Check returns
// ErrNilScorer. CrossGroupThresholdBoost zero-value is 0.0 (no boost).
// The opinionated default 0.05 lives in DefaultConfig (SPEC OVERRIDE
// (Phase 9): default-0.05 location migrated to DefaultConfig per
// 09-CONTEXT.md §2 D-04; spec §12.1 line 1359 was amended in lockstep).
//
// Concurrency: Config is a plain data struct. Consumers passing the
// same Config to multiple concurrent Check invocations must not mutate
// it between calls.
type Config struct {
	// Scorer is required. It governs the similarity computation and the
	// emission threshold (warnings are emitted when Scorer.Match returns
	// true, with the cross-group boost applied where relevant).
	//
	// Check returns ErrNilScorer when Scorer is nil.
	Scorer *fuzzymatch.Scorer

	// CompareAcrossGroups enables the cross-group pass: items with
	// different Group values are compared against each other. When
	// false (default), only same-group pairs are compared.
	CompareAcrossGroups bool

	// CrossGroupThresholdBoost is added to the Scorer's threshold when
	// evaluating cross-group pairs. The cross-group pass is inherently
	// noisier than the within-group pass; a small positive value
	// (typically 0.02–0.05) reduces false positives without disabling
	// the pass.
	//
	// Range: [0.0, 1.0]. NaN, ±Inf, or out-of-range values are
	// rejected with ErrInvalidConfig at Check entry (D-04). The
	// zero-value is 0.0 (no boost); the opinionated default 0.05
	// lives in DefaultConfig.
	//
	// If scorer.Threshold() + CrossGroupThresholdBoost exceeds 1.0,
	// the effective cross-group threshold is clamped to 1.0 (only
	// byte-identical matches pass; combined with
	// CompareIdenticalAcrossGroups=false this disables cross-group
	// emission, which is documented behaviour).
	CrossGroupThresholdBoost float64

	// CompareIdenticalAcrossGroups, when false (the recommended
	// default), suppresses cross-group warnings for pairs whose names
	// are byte-identical after normalisation. Operators legitimately
	// reuse the same name (e.g. user_id) across groups; surfacing every
	// such pair would drown real similar-but-not-equal signals.
	CompareIdenticalAcrossGroups bool

	// SuppressedPairs is a list of name pairs that should not produce a
	// warning, in addition to per-Item SilenceLint flags. Pairs are
	// order-independent and canonicalised at Check entry using the
	// Scorer's normalisation options.
	//
	// Range: every entry's two strings must be non-empty.
	// ErrInvalidConfig is returned at Check entry for any entry where
	// one or both strings are empty (D-05); offending indices are
	// collected via errors.Join.
	//
	// Self-pairs (a == b after normalisation) are silently kept — they
	// are harmless because Check never emits a self-warning under D-06
	// (duplicate (Name, Group) is rejected at validation) and the i<j
	// pair iteration discipline.
	//
	// Caveat: the suppression map is keyed by the canonical
	// (lexicographically-sorted, normalised) pair. A self-pair entry
	// {"foo","foo"} whose normalised form coincides with the canonical
	// key of a DISTINCT-name candidate pair (e.g. {"foo","FOO"} when
	// the Scorer's normalisation lowercases) will also suppress that
	// distinct pair. This is the inevitable consequence of canonical-
	// pair semantics — flagged here so consumers building suppression
	// lists programmatically know to expect it. If precise control is
	// needed, omit self-pair entries from SuppressedPairs.
	//
	// Build cost is O(N) in len(SuppressedPairs); per-candidate lookup
	// is O(1) via an internal canonical-pair map built once at Check
	// entry.
	SuppressedPairs [][2]string
}

// Warning is one detected similar-name pair. The deterministic output
// sort by (Kind, NameA, NameB, GroupA, GroupB) — including the
// lexicographic canonicalisation of NameA/NameB so NameA is the
// raw-byte-lex smaller of the pair (with TagA/TagB and GroupA/GroupB
// swapping in lockstep when the names are flipped) — is applied by
// Plan 09-06's sort step. Consumers reading post-09-06 Check output get
// the fully canonicalised form: for any emitted Warning, NameA <= NameB
// under string-lex comparison.
//
// SPEC OVERRIDE (Phase 9): Scores is map[fuzzymatch.AlgoID]float64
// (typed enum keys), NOT map[string]float64 as docs/requirements.md
// §12.1 line 1337 originally specified. Extends the Phase 8 ScoreAll
// override at §8.3 + §8.6 for the same typed-enum-keys rationale: the
// rest of the library exposes AlgoID and gives consumers compile-time
// key safety. Use AlgoID.String() for snake_case / CamelCase display.
// See 09-CONTEXT.md §1 D-01 and the api-ergonomics-reviewer sign-off
// recorded in plan 09-01's PR. Spec §12.1 was amended in lockstep with
// this declaration.
//
// Concurrency: Warning is a plain data struct. The map in Scores is
// freshly allocated by Check for each emitted Warning; consumers may
// mutate it freely without affecting other Warning values.
type Warning struct {
	// Kind classifies the pair scope (within-group vs cross-group).
	// SPEC OVERRIDE (Phase 9): type renamed from WarningKind to Kind
	// per 09-CONTEXT.md §1 D-02; spec §12.1 was amended in lockstep.
	Kind Kind

	// NameA, NameB are the raw item names (NOT normalised). Plan 09-06's
	// sort step canonicalises the pair so NameA is the raw-byte-lex
	// smaller of the two strings. For any post-09-06 Warning, NameA <=
	// NameB under Go's native string comparison. When the canonicaliser
	// swaps NameA and NameB, TagA/TagB and GroupA/GroupB swap in
	// lockstep so each (NameA, GroupA, TagA) triple still describes the
	// same source Item.
	NameA, NameB string

	// GroupA, GroupB are the raw group values from the corresponding
	// Items. For KindWithinGroup warnings GroupA == GroupB; for
	// KindAcrossGroups warnings GroupA != GroupB.
	GroupA, GroupB string

	// TagA, TagB are the opaque Tag values from the corresponding
	// Items. The library does not interpret them.
	TagA, TagB any

	// Score is the composite score returned by Scorer.Score for this
	// pair. In [0.0, 1.0] under the default WithNormaliseWeights(true)
	// Scorer construction; the consumer takes responsibility for the
	// range otherwise.
	Score float64

	// Scores carries the per-algorithm breakdown from
	// Scorer.ScoreAll(NameA, NameB).
	//
	// SPEC OVERRIDE (Phase 9): map[fuzzymatch.AlgoID]float64 (typed
	// enum keys) per 09-CONTEXT.md §1 D-01. Extends the Phase 8
	// ScoreAll override at §8.3 + §8.6. Iteration order is
	// non-deterministic per Go map semantics; consumers requiring
	// stable order MUST sort the keys themselves (typically via
	// fuzzymatch.AlgoIDs() then key-lookup).
	Scores map[fuzzymatch.AlgoID]float64
}

// DefaultConfig returns an opinionated Config bound to the supplied
// Scorer. The defaults are:
//
//   - Scorer:                       s (the argument)
//   - CompareAcrossGroups:          false (within-group only)
//   - CrossGroupThresholdBoost:     0.05 — applied only when the
//     consumer subsequently sets CompareAcrossGroups = true
//   - CompareIdenticalAcrossGroups: false (suppress identical names
//     across groups by default)
//   - SuppressedPairs:              nil
//
// SPEC OVERRIDE (Phase 9): the 0.05 default lives here, NOT as the
// zero-value of Config.CrossGroupThresholdBoost; see 09-CONTEXT.md §2
// D-04. Spec §12.1 line 1359 was amended in lockstep — the spec's
// "Default: 0.05" sentence was migrated from the field godoc to this
// helper godoc.
//
// Mirrors the Phase 8 functional-options-with-helper pattern
// (DefaultScorer / DefaultScorerOptions): the zero-value of the Config
// struct remains a valid (minimal) configuration; the opinionated
// helper bakes in the experience-tuned values.
//
// Typical usage:
//
//	s := fuzzymatch.DefaultScorer()
//	cfg := scan.DefaultConfig(s)
//	cfg.CompareAcrossGroups = true // opt in
//	warnings, err := scan.Check(items, cfg)
//
// Concurrency: DefaultConfig has no side effects and is safe for
// concurrent use. A fresh Config value is returned on every call.
func DefaultConfig(s *fuzzymatch.Scorer) Config {
	return Config{
		Scorer:                       s,
		CompareAcrossGroups:          false,
		CrossGroupThresholdBoost:     0.05,
		CompareIdenticalAcrossGroups: false,
		SuppressedPairs:              nil,
	}
}

// Check compares every pair of items per the Config and returns
// warnings for pairs where the Scorer reports a similar match.
//
// Within-group pass: for every group, every i<j pair is evaluated via
// cfg.Scorer.Match(itemA.Name, itemB.Name). Match returns true when the
// composite Score is at or above the Scorer's threshold (the boundary
// is inclusive — see scorer.go Match godoc).
//
// Cross-group pass: when cfg.CompareAcrossGroups == true, every pair
// spanning two distinct groups is evaluated via cfg.Scorer.Score
// against the *effective cross-group threshold*:
//
//	effectiveThreshold = math.Min(1.0, Scorer.Threshold() + cfg.CrossGroupThresholdBoost)
//
// The math.Min clamp pins the effective threshold at 1.0 even when the
// arithmetic sum would exceed it (e.g. Threshold 0.85 + Boost 0.20 would
// arithmetically yield 1.05 — the clamp makes that 1.0, meaning only
// byte-identical-post-normalise pairs reach the threshold). Cross-group
// emission uses Scorer.Score (not Scorer.Match) because Match applies
// the within-group threshold; only Score exposes the raw composite that
// can be compared against the boosted threshold.
//
// Suppression composition (Plan 09-05): three rules compose via OR
// via the isSuppressed predicate; the cheapest rule fires first.
//
//   - Rule 1 — per-item SilenceLint: when either Item.SilenceLint is
//     true, the pair is suppressed. One-sided semantics.
//   - Rule 2 — SuppressedPairs canonical-pair lookup: the consumer-
//     supplied [][2]string is normalised once at Check entry (via the
//     Scorer's NormalisationOptions) and stored as a canonical-pair
//     set; lookups are order-independent (Pitfall 4 mitigation —
//     consumers may pass raw forms regardless of case or separators).
//   - Rule 3 — cross-group identical-name default (SCAN-04): when
//     cfg.CompareIdenticalAcrossGroups == false (the DefaultConfig
//     default), pairs in the cross-group pass whose normalised names
//     coincide are suppressed. Operators legitimately reuse the same
//     name across groups; surfacing every such pair would drown signal.
//
// The Plan-09-03 inline cross-group identical-name check has been
// MIGRATED into Rule 3 — both naive and bucket cross-group emission
// paths now route through isSuppressed. The migration is behaviour-
// preserving; pairs that are SIMILAR but not byte-identical post-
// normalise are unaffected.
//
// Name normalisation policy: Check reads the Scorer's normalisation
// options via Scorer.NormalisationOptions() ONCE per invocation and
// pre-computes a parallel `normalisedNames` array used only for the
// identical-name suppression check above. The RAW item.Name strings are
// passed to Scorer.Score / Scorer.Match / Scorer.ScoreAll — the Scorer
// re-normalises internally (scorer.go:356-360). Check never
// double-normalises (Pitfall 5 — 09-RESEARCH.md lines 488-498).
//
// Warning population: Warning.Scores is populated via
// Scorer.ScoreAll(NameA, NameB) only on emission (lazy population). The
// per-algorithm breakdown is paid for once per emitted Warning, not once
// per candidate pair.
//
// Validation gate: Check's first step is validateCheck(items, cfg).
// Failures propagate unmodified (sentinel-identity ErrNilScorer for
// nil Scorer; errors.Join-wrapped ErrInvalidItem / ErrInvalidConfig for
// the collect-all phases). See validate.go for the locked P1..P4 order.
//
// Determinism: groups are iterated in sorted-key order; items within a
// group are iterated in slice-index order (groupIndices preserves
// insertion order via append). Before returning, Check
// lex-canonicalises every Warning (raw-byte-lex on (NameA, NameB) so
// NameA <= NameB; TagA/TagB and GroupA/GroupB swap in lockstep) and
// then applies sort.SliceStable on the 5-tuple sort key
// (Kind, NameA, NameB, GroupA, GroupB). Map iteration on the
// groupIndices map is confined to building the sortedGroups slice (no
// map iteration on output paths).
//
// In-line completeness assertion (Plan 09-06; SCAN-05 defence-in-depth):
// after sorting, Check scans adjacent warnings linearly; any pair
// sharing the full 5-tuple sort key triggers
// panic(fmt.Errorf("...: %w", fuzzymatch.ErrInternalInvariantViolated)).
// The assertion is unreachable under valid input because D-06's
// validateCheck rejects duplicate (Name, Group) at the door — but it
// exists as the documented invariant gate per the Phase 8.5 Gap 5
// typed-panic convention. Consumers discriminate via
// errors.Is(recovered, fuzzymatch.ErrInternalInvariantViolated); the
// panic value indicates a library bug, not user error. The panic
// message includes the duplicate index + Kind + Names + Groups; Tag
// values are intentionally omitted (T-09-06-02 — Tag is consumer-
// supplied opaque data that may carry sensitive context).
//
// Plan-stage scope:
//
//   - Plan 09-03: naive O(N²) within-group + O(N×M) cross-group
//     passes; SCAN-04 identical-name suppression default; validateCheck
//     wired as P0.
//   - Plan 09-04: bucket dispatch lands as a parallel branch alongside
//     the naive passes from 09-03. For each group, when len(idx) >
//     bucketThreshold (and the test-only forceNaivePath flag is false)
//     the bucket path runs; otherwise the naive nested loop from Plan
//     09-03 runs. The PropCheck_BucketEquivalentToNaive property test
//     (scan/props_test.go) proves the two paths produce identical
//     warning sets for any randomly-generated input — SCAN-02 load-
//     bearing gate.
//   - Plan 09-05 (this plan): full suppression composition (SCAN-03)
//     via the isSuppressed predicate. Three rules — per-item SilenceLint,
//     SuppressedPairs canonical-pair lookup, and the SCAN-04 cross-
//     group identical-name default (migrated from Plan-09-03's inline
//     check into Rule 3) — compose via short-circuit OR. The predicate
//     is called pre-emission on both the naive and bucket emission
//     paths so SCAN-02 bucket-vs-naive equivalence holds under
//     suppression.
//   - Plan 09-06: deterministic sort + in-line completeness assertion.
//
// Dispatch decision (Plan 09-04):
//
//   - Within-group: per group, if len(idx) > bucketThreshold &&
//     !forceNaivePath, use the bucket path. Otherwise use the naive
//     nested loop.
//   - Cross-group: per group-pair (gi, gj), if len(idxA) + len(idxB) >
//     bucketThreshold && !forceNaivePath, use the bucket path
//     (per-pair bucket built over idxA + idxB combined). Otherwise
//     use the naive nested loop.
//
// Worst-case complexity is unchanged from naive — an adversarial
// input where every item shares a single token reduces to the same
// nested loop. Expected complexity on realistic identifier-style
// workloads (snake_case vs camelCase, etc.) drops sharply because
// most non-matching pairs share no token and are eliminated without
// paying Scorer.Score. The PERF-05 ≤ 2s / 10,000-items budget is met
// by this pruning combined with the Phase 8.5 Q8b Tokenise ASCII fast
// path.
//
// Concurrency: Check is a pure function with no goroutines, channels,
// or mutexes. Safe for concurrent invocation on disjoint inputs; the
// Scorer is immutable post-NewScorer (Phase 8), so concurrent Check
// invocations sharing a Scorer are also safe. The fresh Scores map in
// each emitted Warning is freshly allocated by Scorer.ScoreAll —
// consumers may mutate it freely.
//
// Returns:
//
//   - (warnings, nil) on success (warnings is nil for the < 2 items
//     early-exit case and for the no-emissions happy path; consumers
//     using len(warnings) treat both nil and empty identically)
//   - (nil, ErrNilScorer) when cfg.Scorer is nil
//   - (nil, err) where errors.Is(err, ErrInvalidConfig) is true when
//     Config field validation fails
//   - (nil, err) where errors.Is(err, ErrInvalidItem) is true when
//     items[] validation fails (errors.Join-wrapped, one entry per
//     offending index)
//   - (nil, err) where errors.Is(err, ErrInvalidConfig) is true when
//     SuppressedPairs validation fails (errors.Join-wrapped)
func Check(items []Item, cfg Config) ([]Warning, error) { //nolint:gocyclo // locked pipeline (validation → normalise-once → group-build → within-group naive → cross-group naive with SCAN-04 suppression); each branch is a documented gate that cannot be folded into a sub-helper without splitting the Pitfall 3/5/6 + SCAN-04 contract across files. Plan 09-04 will introduce bucket dispatch as a parallel branch — separating now would require unwinding the split.
	// P0 — Validation gate. validateCheck runs the locked P1..P4
	// pipeline (nil-Scorer → Config fields → Items[] → SuppressedPairs).
	// Failures propagate unmodified. See validate.go.
	if err := validateCheck(items, cfg); err != nil {
		return nil, err
	}

	// Early exit on degenerate inputs. validateCheck already accepted
	// a zero- or one-item slice as structurally valid; no pairs exist
	// to compare. The nil return is consistent with the no-emissions
	// happy path below.
	if len(items) < 2 {
		return nil, nil
	}

	// Read normalisation policy from the Scorer (Plan 09-01 accessor).
	// applyNormalisation == true means the Scorer applies Normalise
	// internally; we mirror that here only to pre-compute the
	// parallel `normalisedNames` array for the SCAN-04 identical-name
	// suppression check. Critically, the RAW item.Name strings flow
	// through to Scorer.Match / Scorer.Score / Scorer.ScoreAll — the
	// Scorer re-normalises (Pitfall 5).
	normOpts, applyNormalisation := cfg.Scorer.NormalisationOptions()

	// Parallel `normalisedNames` array. Index i holds the normalised
	// (or raw, when applyNormalisation == false) form of items[i].Name.
	// Used for the cross-group identical-name suppression check.
	normalisedNames := make([]string, len(items))
	for i, item := range items {
		if applyNormalisation {
			normalisedNames[i] = fuzzymatch.Normalise(item.Name, normOpts)
		} else {
			normalisedNames[i] = item.Name
		}
	}

	// Plan 09-04: pre-compute tokenised names once per Check
	// invocation (Pitfall 7: tokenise at most once per Item). The
	// tokens slice mirrors items by index; bucket builds read from it
	// directly. Per Open Question 2 (09-RESEARCH.md), tokeniseAll
	// Normalises (when applyNormalisation == true) then Tokenises with
	// DefaultTokeniseOptions, so the bucket keys mirror what the
	// Scorer sees at scoring time.
	tokenisedNames := tokeniseAll(items, normOpts, applyNormalisation)

	// Plan 09-05: build the suppression context once at Check entry.
	// buildSuppressionCtx canonicalises every SuppressedPairs entry
	// using the Scorer's normalisation options (Pitfall 4 — without
	// this step, raw consumer-supplied pairs would never match the
	// Normalised candidate pairs in the inner loop). The returned ctx
	// is read-only for the rest of this Check invocation and passes
	// directly into isSuppressed at every emission site.
	suppressCtx := buildSuppressionCtx(
		cfg.SuppressedPairs,
		normOpts,
		applyNormalisation,
		cfg.CompareIdenticalAcrossGroups,
	)

	// Group items by their Group value. The slice values preserve
	// insertion order via append, which mirrors items[] slice-index
	// order — important for deterministic warning emission within a
	// group.
	groupIndices := make(map[string][]int, len(items))
	for i, item := range items {
		groupIndices[item.Group] = append(groupIndices[item.Group], i)
	}

	// Materialise the sorted group-key slice. This is the ONLY place
	// the groupIndices map is iterated; downstream code walks the
	// sortedGroups slice. No map iteration on output paths (per
	// .claude/skills/determinism-standards).
	sortedGroups := make([]string, 0, len(groupIndices))
	for g := range groupIndices {
		sortedGroups = append(sortedGroups, g)
	}
	sort.Strings(sortedGroups)

	// Warnings accumulator. nil-vs-empty-slice semantics are equivalent
	// for len()-based consumers; we leave this as a default-zero slice
	// so the no-emissions happy path naturally returns a non-nil but
	// empty slice. (The early-exit case above returns nil for the
	// degenerate < 2 items input; both representations are valid per
	// the godoc.)
	warnings := make([]Warning, 0)

	// Within-group pass. For each group (in sorted-key order), dispatch
	// to the bucket path when len(idx) > bucketThreshold (and the
	// test-only forceNaivePath flag is false), otherwise to the naive
	// nested loop from Plan 09-03. Both paths produce identical
	// warning sets per the SCAN-02 property test
	// PropCheck_BucketEquivalentToNaive (scan/props_test.go). Compute
	// Score once per pair and compare against Scorer.Threshold()
	// directly — preserves Match's inclusive >= semantics
	// (scorer.go:399) and avoids the triple-Scorer-call pattern
	// (Match-internally-calls-Score + explicit Score + ScoreAll) that
	// inflated the per-emission cost ~3× (convergent reviewer finding
	// on Plan 09-03).
	withinThreshold := cfg.Scorer.Threshold()
	for _, group := range sortedGroups {
		idx := groupIndices[group]
		if !forceNaivePath.Load() && len(idx) > bucketThreshold {
			// Bucket path. Build the bucket once per group, then walk
			// each source index's candidate set. The j > i filter
			// (== idx-position-aware dedup) is replaced by direct index
			// comparison (j > i on the raw item index) because the
			// bucket dispatch enumerates candidates by raw item index,
			// not by group-local position.
			bucket := buildBucket(idx, tokenisedNames)
			for _, i := range idx {
				candidates := bucketCandidates(i, bucket, tokenisedNames)
				for _, j := range candidates {
					if j <= i {
						continue // emit each unordered pair exactly once
					}
					a, b := items[i], items[j]
					warnings = tryEmit(warnings, a, b, KindWithinGroup,
						normalisedNames[i], normalisedNames[j],
						withinThreshold, suppressCtx, cfg.Scorer)
				}
			}
		} else {
			// Naive path (Plan 09-03 fallback for small groups + the
			// test-only forceNaivePath toggle).
			for i := 0; i < len(idx); i++ {
				for j := i + 1; j < len(idx); j++ {
					a, b := items[idx[i]], items[idx[j]]
					warnings = tryEmit(warnings, a, b, KindWithinGroup,
						normalisedNames[idx[i]], normalisedNames[idx[j]],
						withinThreshold, suppressCtx, cfg.Scorer)
				}
			}
		}
	}

	// Cross-group pass. Active only when CompareAcrossGroups == true.
	// Walks sorted group-key pairs (gi < gj) to ensure deterministic
	// emission ordering; within each group pair, dispatches to bucket
	// or naive based on the combined group sizes (Plan 09-04). Both
	// paths apply the SCAN-04 identical-name suppression default
	// identically.
	if cfg.CompareAcrossGroups {
		// Effective threshold = within-group threshold + boost,
		// CLAMPED to 1.0 (Pitfall 6 — boosts can never push the
		// effective threshold above 1.0, the score range upper bound).
		effectiveThreshold := math.Min(1.0, cfg.Scorer.Threshold()+cfg.CrossGroupThresholdBoost)

		// Cross-group bucket optimisation (phase-end remediation):
		// build ONE per-group tokenBucket per group at Check entry
		// (perGroupBuckets[gi] is the bucket over groupIndices[gi]).
		// For each (gi, gj) pair we then enumerate candidates from
		// bucket-gj for each source i in idxA — the candidate set is
		// naturally pre-filtered to gj members by construction, so
		// no separate group-membership filter is needed. The bucket
		// build cost is amortised once per group across all (gi, gj)
		// pairs that reference it. This replaces the prior per-pair
		// union+inB rebuild that accumulated ~1.3B allocs and ~189s
		// wall-clock on the 10k items / 500 groups workload (spec
		// §12.6) — the per-pair cost is now dominated by the
		// inner-loop Scorer.Score calls, not bucket maintenance.
		//
		// SCAN-02 correctness: bucket equivalence is preserved by
		// construction. For any source i in idxA seeking candidates
		// in idxB, the per-group bucket-gj contains exactly idxB's
		// items keyed by their tokens. A pair (i, j) with j ∈ idxB
		// is reachable from bucketCandidates(i, bucket-gj) iff i and
		// j share at least one token — identical to the per-pair
		// bucket built over (idxA ∪ idxB) and filtered to idxB.
		//
		// Build gate: skip the per-group buckets for very small
		// inputs (≤ bucketThreshold total items) where naive
		// cross-group is faster than the bucket-build overhead.
		// forceNaivePath (test-only) also forces naive everywhere
		// to keep PropCheck_BucketEquivalentToNaive deterministic.
		var perGroupBuckets []map[string][]int
		useGlobalBucket := !forceNaivePath.Load() && len(items) > bucketThreshold
		if useGlobalBucket {
			perGroupBuckets = make([]map[string][]int, len(sortedGroups))
			for gi, g := range sortedGroups {
				perGroupBuckets[gi] = buildBucket(groupIndices[g], tokenisedNames)
			}
		}

		for gi := 0; gi < len(sortedGroups); gi++ {
			for gj := gi + 1; gj < len(sortedGroups); gj++ {
				idxA := groupIndices[sortedGroups[gi]]
				idxB := groupIndices[sortedGroups[gj]]
				if useGlobalBucket {
					// Per-group-bucket path. For each i in idxA,
					// enumerate candidates from bucket-gj. The
					// candidate set is naturally restricted to gj
					// members; no group-filter is needed.
					bucketB := perGroupBuckets[gj]
					for _, i := range idxA {
						candidates := bucketCandidates(i, bucketB, tokenisedNames)
						for _, j := range candidates {
							a, b := items[i], items[j]
							warnings = tryEmit(warnings, a, b, KindAcrossGroups,
								normalisedNames[i], normalisedNames[j],
								effectiveThreshold, suppressCtx, cfg.Scorer)
						}
					}
				} else {
					// Naive path (Plan 09-03 fallback for small group
					// pairs + the test-only forceNaivePath toggle).
					for _, i := range idxA {
						for _, j := range idxB {
							a, b := items[i], items[j]
							warnings = tryEmit(warnings, a, b, KindAcrossGroups,
								normalisedNames[i], normalisedNames[j],
								effectiveThreshold, suppressCtx, cfg.Scorer)
						}
					}
				}
			}
		}
	}

	// Plan 09-06: lex-canonicalise every Warning so NameA <= NameB on
	// raw-byte lex compare. Groups + Tags swap in lockstep so the
	// (Name, Group, Tag) attribution of each item is preserved.
	for i := range warnings {
		if warnings[i].NameA > warnings[i].NameB {
			warnings[i].NameA, warnings[i].NameB = warnings[i].NameB, warnings[i].NameA
			warnings[i].GroupA, warnings[i].GroupB = warnings[i].GroupB, warnings[i].GroupA
			warnings[i].TagA, warnings[i].TagB = warnings[i].TagB, warnings[i].TagA
		}
	}

	// Plan 09-06: deterministic sort by the 5-tuple
	// (Kind, NameA, NameB, GroupA, GroupB). sort.SliceStable preserves
	// relative input order on ties — but D-06 guarantees no two
	// Warnings can share the full 5-tuple key, so stability is a
	// belt-and-braces choice. Every field participates in the
	// comparator so the sort is a strict total order on valid input.
	sort.SliceStable(warnings, func(i, j int) bool {
		if warnings[i].Kind != warnings[j].Kind {
			return warnings[i].Kind < warnings[j].Kind
		}
		if warnings[i].NameA != warnings[j].NameA {
			return warnings[i].NameA < warnings[j].NameA
		}
		if warnings[i].NameB != warnings[j].NameB {
			return warnings[i].NameB < warnings[j].NameB
		}
		if warnings[i].GroupA != warnings[j].GroupA {
			return warnings[i].GroupA < warnings[j].GroupA
		}
		return warnings[i].GroupB < warnings[j].GroupB
	})

	// Plan 09-06: in-line completeness assertion. Linear scan of
	// adjacent sorted warnings; any pair sharing the full sort key
	// panics with fuzzymatch.ErrInternalInvariantViolated (Phase 8.5
	// Gap 5). Unreachable on valid input because D-06 rejects
	// duplicate (Name, Group) at validation — defence-in-depth gate
	// per the Phase 9 SCAN-05 requirement.
	assertSortKeyComplete(warnings)

	return warnings, nil
}

// assertSortKeyComplete is the in-line defence-in-depth gate for
// SCAN-05's sort-key uniqueness invariant. Plan 09-06 calls it after
// sorting the warnings slice; any two adjacent warnings sharing the
// full (Kind, NameA, NameB, GroupA, GroupB) sort key trigger a panic
// wrapping fuzzymatch.ErrInternalInvariantViolated (Phase 8.5 Gap 5
// typed-panic convention).
//
// Under correct usage the panic is unreachable: D-06's validateCheck
// rejects items with duplicate (Name, Group), so two distinct items
// can never produce two warnings with the same NameA/NameB/GroupA/
// GroupB tuple. The assertion exists to catch any future library bug
// (a refactor of the bucket dispatch, a regression in the
// canonicalisation step, etc.) that could violate the invariant — at
// which point the panic surfaces immediately rather than silently
// emitting a non-deterministic slice.
//
// Panic message format (locked):
//
//	scan: duplicate sort key at index <i> (Kind=<k> NameA=<a> NameB=<b>
//	GroupA=<gA> GroupB=<gB>) — Items[] validation should have prevented
//	this: fuzzymatch: internal invariant violated (...)
//
// Tag values are intentionally OMITTED (T-09-06-02 mitigation — Tag is
// consumer-supplied opaque data that may carry sensitive context).
// Consumers discriminate via errors.Is on the recovered panic value:
//
//	defer func() {
//	    if r := recover(); r != nil {
//	        if err, ok := r.(error); ok && errors.Is(err, fuzzymatch.ErrInternalInvariantViolated) {
//	            // library bug — collect stack + file issue
//	        }
//	    }
//	}()
//
// tryEmit applies the suppression + threshold gate and appends a
// Warning to the in-flight slice when the pair passes. Collapses
// the four emission sites (within naive, within bucket, cross
// naive, cross bucket) into a single helper — the code-reviewer NIT
// from Plan 09-06 (Check body length + duplicated bucket-vs-naive
// emit body). All four sites now share one source-of-truth for the
// emit semantics: suppression check first, then ScoreAndAll, then
// threshold gate, then Warning construction.
//
// Returns the (possibly-extended) warnings slice so callers can
// chain `warnings = tryEmit(...)` per emission attempt without
// pointer indirection.
func tryEmit(
	warnings []Warning,
	a, b Item,
	kind Kind,
	normalisedA, normalisedB string,
	threshold float64,
	suppressCtx suppressionCtx,
	scorer *fuzzymatch.Scorer,
) []Warning {
	if isSuppressed(a, b, kind, normalisedA, normalisedB, suppressCtx) {
		return warnings
	}
	score, breakdown := scorer.ScoreAndAll(a.Name, b.Name)
	if score < threshold {
		return warnings
	}
	return append(warnings, Warning{
		Kind:   kind,
		NameA:  a.Name,
		NameB:  b.Name,
		GroupA: a.Group,
		GroupB: b.Group,
		TagA:   a.Tag,
		TagB:   b.Tag,
		Score:  score,
		Scores: breakdown,
	})
}

func assertSortKeyComplete(warnings []Warning) {
	for i := 1; i < len(warnings); i++ {
		prev := warnings[i-1]
		cur := warnings[i]
		if prev.Kind == cur.Kind &&
			prev.NameA == cur.NameA && prev.NameB == cur.NameB &&
			prev.GroupA == cur.GroupA && prev.GroupB == cur.GroupB {
			panic(fmt.Errorf(
				"scan: duplicate sort key at index %d (Kind=%s NameA=%q NameB=%q GroupA=%q GroupB=%q) — Items[] validation should have prevented this: %w",
				i, cur.Kind, cur.NameA, cur.NameB, cur.GroupA, cur.GroupB,
				fuzzymatch.ErrInternalInvariantViolated,
			))
		}
	}
}
