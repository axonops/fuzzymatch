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

// scan_steps.go contains the godog step definitions for the Phase 9
// collection-scan sub-package (scan.Check + suppression composition).
// The ScanContext struct holds state between steps within a scenario;
// each scenario instantiates a fresh ScanContext via InitScanSteps
// (called from algorithms_steps.go's InitializeScenario alongside
// InitScorerSteps / InitValidateSteps / InitNormalisationSteps /
// InitDeterminismSteps).
//
// All scan scenarios live under the @scan Gherkin tag in
// features/scan.feature; all suppression scenarios live under the
// @suppression tag in features/suppression.feature.
//
// goleak.VerifyTestMain in tests/bdd/bdd_test.go gates the implicit
// "no goroutine leak" invariant — scan is a pure-function library so
// no goroutines are expected, but goleak catches regressions.
//
// testify is permitted in tests/bdd/ per CLAUDE.md test-dependency
// allowlist; the steps below follow the existing scorer_steps.go /
// validate_steps.go convention of returning fmt.Errorf assertions
// rather than relying on testify, for consistency.

package steps

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"math"
	"sort"
	"strconv"
	"strings"

	"github.com/cucumber/godog"
	messages "github.com/cucumber/messages/go/v21"

	"github.com/axonops/fuzzymatch"
	"github.com/axonops/fuzzymatch/scan"
)

// ScanContext holds state between BDD steps within a scan or
// suppression scenario. Each scenario instantiates a fresh ScanContext
// to ensure isolation (steps within one scenario share state; steps
// across scenarios do not).
//
// scorer is the *fuzzymatch.Scorer under test for the scenario; it is
// initialised by the "Given I construct" steps and consumed by the
// scan config and by the cross-group threshold-boost arithmetic step.
//
// items is the []scan.Item under test for the scenario; populated by
// the "the scan items" table step (with "(empty)" sentinel support for
// the D-03 empty-Name validation scenario).
//
// cfg is the scan.Config under test; initialised by "the scan config
// is the default scan config" or mutated by the various
// "I set ... to ..." steps.
//
// warnings is the []scan.Warning result of scan.Check (the When step
// that drives invocation), consumed by every Then assertion step.
//
// secondWarnings backs the determinism scenario: "I invoke scan.Check
// twice" runs Check a second time and stores the result here so the
// byte-identical assertion step can compare both results.
//
// lastErr is the error result of scan.Check; consumed by the
// errors.Is-style assertion steps (matches scan.ErrInvalidItem,
// scan.ErrInvalidConfig, scan.ErrNilScorer, etc.).
type ScanContext struct {
	scorer         *fuzzymatch.Scorer
	items          []scan.Item
	cfg            scan.Config
	warnings       []scan.Warning
	secondWarnings []scan.Warning
	lastErr        error
}

// iConstructTheDefaultScorerForScan constructs DefaultScorer and stores
// it in the ScanContext. Backs scenarios under @scan and @suppression.
// Step regex: `^I construct the default Scorer for scan$`
func (sc *ScanContext) iConstructTheDefaultScorerForScan() error {
	sc.scorer = fuzzymatch.DefaultScorer()
	return nil
}

// theScanItems parses a godog table of scan.Item rows. Columns: name,
// group, silence_lint. The "(empty)" sentinel in the name column
// yields "" — used by the D-03 empty-Name validation scenario. The
// silence_lint column accepts "true" / "false"; defaults to false on
// any other value to keep happy-path tables compact.
// Step regex: `^the scan items$`
func (sc *ScanContext) theScanItems(table *godog.Table) error {
	items, err := parseScanItems(table)
	if err != nil {
		return fmt.Errorf("parseScanItems: %w", err)
	}
	sc.items = items
	return nil
}

// theScanConfigIsTheDefaultScanConfig initialises sc.cfg via
// scan.DefaultConfig(sc.scorer). Requires that a prior step has set
// sc.scorer (the "Given I construct the default Scorer for scan" step).
// Step regex: `^the scan config is the default scan config$`
func (sc *ScanContext) theScanConfigIsTheDefaultScanConfig() error {
	if sc.scorer == nil {
		return fmt.Errorf("scorer not constructed; precede this step with 'I construct the default Scorer for scan'")
	}
	sc.cfg = scan.DefaultConfig(sc.scorer)
	return nil
}

// theScanConfigHasANilScorer sets cfg.Scorer to nil for the
// ErrNilScorer error-path scenario.
// Step regex: `^the scan config has a nil Scorer$`
func (sc *ScanContext) theScanConfigHasANilScorer() error {
	sc.cfg = scan.Config{Scorer: nil}
	return nil
}

// iSetCompareAcrossGroupsTo flips Config.CompareAcrossGroups.
// Step regex: `^I set CompareAcrossGroups to (true|false)$`
func (sc *ScanContext) iSetCompareAcrossGroupsTo(value string) error {
	b, err := strconv.ParseBool(value)
	if err != nil {
		return fmt.Errorf("parse CompareAcrossGroups bool %q: %w", value, err)
	}
	sc.cfg.CompareAcrossGroups = b
	return nil
}

// iSetCompareIdenticalAcrossGroupsTo flips Config.CompareIdenticalAcrossGroups.
// Step regex: `^I set CompareIdenticalAcrossGroups to (true|false)$`
func (sc *ScanContext) iSetCompareIdenticalAcrossGroupsTo(value string) error {
	b, err := strconv.ParseBool(value)
	if err != nil {
		return fmt.Errorf("parse CompareIdenticalAcrossGroups bool %q: %w", value, err)
	}
	sc.cfg.CompareIdenticalAcrossGroups = b
	return nil
}

// theScanConfigHasCrossGroupThresholdBoostSetToNaN injects NaN into
// cfg.CrossGroupThresholdBoost for the D-04 ErrInvalidConfig scenario.
// Step regex: `^the scan config has CrossGroupThresholdBoost set to NaN$`
func (sc *ScanContext) theScanConfigHasCrossGroupThresholdBoostSetToNaN() error {
	sc.cfg.CrossGroupThresholdBoost = math.NaN()
	return nil
}

// theScanConfigHasCrossGroupThresholdBoostSetTo sets a specific float
// value on cfg.CrossGroupThresholdBoost. Used by the boost arithmetic
// scenario to set a known value before assertion.
// Step regex: `^the scan config has CrossGroupThresholdBoost set to (-?\d+\.?\d*)$`
func (sc *ScanContext) theScanConfigHasCrossGroupThresholdBoostSetTo(value float64) error {
	sc.cfg.CrossGroupThresholdBoost = value
	return nil
}

// iSetSuppressedPairs appends rows of a godog table into
// cfg.SuppressedPairs. Columns: a, b. Strings flow through unchanged
// (Check canonicalises via the Scorer's normalisation options).
// Step regex: `^I set suppressed pairs$`
func (sc *ScanContext) iSetSuppressedPairs(table *godog.Table) error {
	if len(table.Rows) < 2 {
		return fmt.Errorf("suppressed-pairs table needs a header plus at least one data row")
	}
	header := table.Rows[0]
	idxA, idxB := -1, -1
	for i, cell := range header.Cells {
		switch strings.TrimSpace(cell.Value) {
		case "a":
			idxA = i
		case "b":
			idxB = i
		}
	}
	if idxA < 0 || idxB < 0 {
		return fmt.Errorf("suppressed-pairs table must have columns 'a' and 'b'; got header %v", header.Cells)
	}
	for ri, row := range table.Rows[1:] {
		// Defensive guard against malformed (ragged) tables. godog
		// normally rejects these at parse time, but the explicit check
		// mirrors parseScanItemRow's discipline and surfaces the failure
		// clearly. Closes code-reviewer MEDIUM M1 finding on Plan 09-07.
		need := idxA
		if idxB > need {
			need = idxB
		}
		if need >= len(row.Cells) {
			return fmt.Errorf("suppressed-pairs row %d has %d cells; need at least %d", ri+1, len(row.Cells), need+1)
		}
		a := decodeEmptySentinel(row.Cells[idxA].Value)
		b := decodeEmptySentinel(row.Cells[idxB].Value)
		sc.cfg.SuppressedPairs = append(sc.cfg.SuppressedPairs, [2]string{a, b})
	}
	return nil
}

// iInvokeScanCheck calls scan.Check(items, cfg) and stores the result.
// Step regex: `^I invoke scan\.Check$`
func (sc *ScanContext) iInvokeScanCheck() error {
	sc.warnings, sc.lastErr = scan.Check(sc.items, sc.cfg)
	return nil
}

// iInvokeScanCheckTwice calls scan.Check twice and stores both results.
// Used by the determinism scenario.
// Step regex: `^I invoke scan\.Check twice$`
func (sc *ScanContext) iInvokeScanCheckTwice() error {
	first, err1 := scan.Check(sc.items, sc.cfg)
	second, err2 := scan.Check(sc.items, sc.cfg)
	sc.warnings = first
	sc.secondWarnings = second
	if err1 != nil {
		return fmt.Errorf("first scan.Check returned error: %w", err1)
	}
	if err2 != nil {
		return fmt.Errorf("second scan.Check returned error: %w", err2)
	}
	return nil
}

// scanCheckReturnsNoError asserts the stored lastErr is nil.
// Step regex: `^scan\.Check returns no error$`
func (sc *ScanContext) scanCheckReturnsNoError() error {
	if sc.lastErr != nil {
		return fmt.Errorf("expected nil error; got: %w", sc.lastErr)
	}
	return nil
}

// scanCheckReturnsAnErrorMatching asserts the stored lastErr matches
// the named sentinel via errors.Is.
// Step regex: `^scan\.Check returns an error matching scan\.(ErrInvalidItem|ErrInvalidConfig|ErrNilScorer)$`
func (sc *ScanContext) scanCheckReturnsAnErrorMatching(sentinel string) error {
	if sc.lastErr == nil {
		return fmt.Errorf("expected error matching scan.%s; got nil", sentinel)
	}
	var want error
	switch sentinel {
	case "ErrInvalidItem":
		want = scan.ErrInvalidItem
	case "ErrInvalidConfig":
		want = scan.ErrInvalidConfig
	case "ErrNilScorer":
		want = scan.ErrNilScorer
	default:
		return fmt.Errorf("unknown sentinel scan.%s", sentinel)
	}
	if !errors.Is(sc.lastErr, want) {
		return fmt.Errorf("error does not match scan.%s: %w", sentinel, sc.lastErr)
	}
	return nil
}

// theErrorMentionsIndexAndIndex asserts the error message string
// contains the two named indices. Used by the D-03 / D-06 multi-error
// scenarios where errors.Join wraps one entry per offending index.
// Step regex: `^the error mentions index (\d+) and index (\d+)$`
func (sc *ScanContext) theErrorMentionsIndexAndIndex(i, j int) error {
	if sc.lastErr == nil {
		return fmt.Errorf("expected non-nil error; got nil")
	}
	msg := sc.lastErr.Error()
	needles := []string{
		fmt.Sprintf("index %d", i),
		fmt.Sprintf("index %d", j),
	}
	for _, n := range needles {
		if !strings.Contains(msg, n) {
			return fmt.Errorf("error message %q does not mention %q", msg, n)
		}
	}
	return nil
}

// theScanWarningsIncludeAPairWithKindAndNames asserts at least one
// warning has the named Kind and the unordered name pair.
// Step regex: `^the scan warnings include a (WithinGroup|AcrossGroups) pair with names "([^"]*)" and "([^"]*)"$`
func (sc *ScanContext) theScanWarningsIncludeAPairWithKindAndNames(kindName, nameA, nameB string) error {
	want, err := parseScanKind(kindName)
	if err != nil {
		return err
	}
	for _, w := range sc.warnings {
		if w.Kind != want {
			continue
		}
		if (w.NameA == nameA && w.NameB == nameB) || (w.NameA == nameB && w.NameB == nameA) {
			return nil
		}
	}
	return fmt.Errorf(
		"no %s warning with names {%q, %q}; got %d warnings: %s",
		kindName, nameA, nameB, len(sc.warnings), renderScanWarnings(sc.warnings),
	)
}

// theScanWarningsListHasEntries asserts the total warning count.
// Step regex: `^the scan warnings list has (\d+) entries$`
func (sc *ScanContext) theScanWarningsListHasEntries(n int) error {
	if len(sc.warnings) != n {
		return fmt.Errorf("warning count = %d; want %d; got: %s", len(sc.warnings), n, renderScanWarnings(sc.warnings))
	}
	return nil
}

// theScanWarningsListHasKindEntries asserts the count of a given Kind.
// Step regex: `^the scan warnings list has (\d+) (WithinGroup|AcrossGroups) entries$`
func (sc *ScanContext) theScanWarningsListHasKindEntries(n int, kindName string) error {
	want, err := parseScanKind(kindName)
	if err != nil {
		return err
	}
	got := 0
	for _, w := range sc.warnings {
		if w.Kind == want {
			got++
		}
	}
	if got != n {
		return fmt.Errorf("%s warning count = %d; want %d; all warnings: %s", kindName, got, n, renderScanWarnings(sc.warnings))
	}
	return nil
}

// noWarningInvolves asserts that no emitted warning references the
// named Item by Name. Used by the SilenceLint suppression scenarios.
// Step regex: `^no scan warning involves "([^"]*)"$`
func (sc *ScanContext) noWarningInvolves(name string) error {
	for _, w := range sc.warnings {
		if w.NameA == name || w.NameB == name {
			return fmt.Errorf(
				"warning involves %q despite suppression; got: %s",
				name, renderScanWarnings(sc.warnings),
			)
		}
	}
	return nil
}

// theWarningsAreSortedByTheCanonical5TupleKey walks the warnings slice
// and asserts each adjacent pair respects the
// (Kind, NameA, NameB, GroupA, GroupB) lex ordering. Used by the
// determinism / sort scenario.
// Step regex: `^the scan warnings are sorted by \(Kind, NameA, NameB, GroupA, GroupB\)$`
func (sc *ScanContext) theWarningsAreSortedByTheCanonical5TupleKey() error {
	for i := 1; i < len(sc.warnings); i++ {
		prev := sc.warnings[i-1]
		cur := sc.warnings[i]
		if !scanWarningLess(prev, cur) && !scanWarningEqualKey(prev, cur) {
			return fmt.Errorf(
				"warnings not sorted at index %d: prev=%+v cur=%+v",
				i, prev, cur,
			)
		}
	}
	return nil
}

// theTwoWarningsSlicesAreByteIdentical asserts the two stored result
// slices marshal to byte-identical JSON. Used by the determinism
// scenario; Scores map keys are stringified for stable JSON ordering.
// Step regex: `^the two scan warnings slices are byte-identical$`
func (sc *ScanContext) theTwoWarningsSlicesAreByteIdentical() error {
	a, err := marshalScanWarnings(sc.warnings)
	if err != nil {
		return fmt.Errorf("marshal first warnings: %w", err)
	}
	b, err := marshalScanWarnings(sc.secondWarnings)
	if err != nil {
		return fmt.Errorf("marshal second warnings: %w", err)
	}
	if !bytes.Equal(a, b) {
		return fmt.Errorf("warnings differ between Check invocations:\nfirst:  %s\nsecond: %s", a, b)
	}
	return nil
}

// theEffectiveCrossGroupThresholdEquals computes the documented boost
// arithmetic (min(1.0, Scorer.Threshold() + boost)) and asserts the
// expected value matches the hardcoded canonical value from the
// Gherkin step. Documents the boost arithmetic invariant at BDD
// level — the runtime threshold gate itself is exercised by
// scan_test.go unit tests.
//
// Fixes the bdd-scenario-reviewer BLOCKING + code-reviewer HIGH
// finding on Plan 09-07: previously the step computed `got` and
// `want` from the identical expression, making the assertion a
// tautology that could never fail.
//
// Step regex: `^the effective cross-group threshold equals ([0-9]+\.?[0-9]*)$`
func (sc *ScanContext) theEffectiveCrossGroupThresholdEquals(want float64) error {
	if sc.scorer == nil {
		return fmt.Errorf("scorer not constructed")
	}
	got := math.Min(1.0, sc.scorer.Threshold()+sc.cfg.CrossGroupThresholdBoost)
	if math.Abs(got-want) > 1e-9 {
		return fmt.Errorf("effective cross-group threshold: got %v want %v (Threshold=%v Boost=%v)",
			got, want, sc.scorer.Threshold(), sc.cfg.CrossGroupThresholdBoost)
	}
	return nil
}

// theWarningsMatchTheNaiveReferenceOutput re-runs scan.Check on the
// same input (which is structurally identical since Check is pure)
// and asserts the two slices match. For the small-group BDD-smoke
// scenario the group size is ≤ bucketThreshold so both runs traverse
// the naive path; this serves as a BDD-granularity smoke test that
// PropCheck_BucketEquivalentToNaive (the load-bearing unit/property
// test) holds at the BDD layer.
// Step regex: `^the scan warnings match the naive-loop reference output$`
func (sc *ScanContext) theWarningsMatchTheNaiveReferenceOutput() error {
	reference, err := scan.Check(sc.items, sc.cfg)
	if err != nil {
		return fmt.Errorf("reference scan.Check returned error: %w", err)
	}
	a, err := marshalScanWarnings(sc.warnings)
	if err != nil {
		return fmt.Errorf("marshal warnings: %w", err)
	}
	b, err := marshalScanWarnings(reference)
	if err != nil {
		return fmt.Errorf("marshal reference: %w", err)
	}
	if !bytes.Equal(a, b) {
		return fmt.Errorf("warnings disagree with naive reference:\nwarnings:  %s\nreference: %s", a, b)
	}
	return nil
}

// scanItemColumns captures the column indices for a scan-items table.
// idxSilence is -1 when the silence_lint column is omitted (treated as
// "all rows have SilenceLint=false").
type scanItemColumns struct {
	idxName, idxGroup, idxSilence int
}

// resolveScanItemColumns walks the header row and returns the column
// indices for name (required), group (required), silence_lint
// (optional). Returns an error when either required column is missing.
// Pulled out of parseScanItems to keep that function's cyclomatic
// complexity within golangci-lint's gocyclo threshold (10).
func resolveScanItemColumns(headerCells []*messages.PickleTableCell) (scanItemColumns, error) {
	cols := scanItemColumns{idxName: -1, idxGroup: -1, idxSilence: -1}
	for i, cell := range headerCells {
		switch strings.TrimSpace(cell.Value) {
		case "name":
			cols.idxName = i
		case "group":
			cols.idxGroup = i
		case "silence_lint":
			cols.idxSilence = i
		}
	}
	if cols.idxName < 0 || cols.idxGroup < 0 {
		return cols, fmt.Errorf("scan-items table must have columns 'name' and 'group'; got header %v", headerCells)
	}
	return cols, nil
}

// parseScanItemRow converts one godog row into a scan.Item using the
// resolved column indices. Pulled out of parseScanItems to keep that
// function's cyclomatic complexity within golangci-lint's gocyclo
// threshold (10).
func parseScanItemRow(rowIdx int, cells []*messages.PickleTableCell, cols scanItemColumns) (scan.Item, error) {
	name := decodeEmptySentinel(cells[cols.idxName].Value)
	group := decodeEmptySentinel(cells[cols.idxGroup].Value)
	silence, err := parseSilenceLint(rowIdx, cells, cols.idxSilence)
	if err != nil {
		return scan.Item{}, err
	}
	return scan.Item{
		Name:        name,
		Group:       group,
		SilenceLint: silence,
	}, nil
}

// parseSilenceLint extracts the silence_lint boolean from a row when
// the column index is present. Empty / absent values default to false
// so happy-path tables stay compact.
func parseSilenceLint(rowIdx int, cells []*messages.PickleTableCell, idx int) (bool, error) {
	if idx < 0 || idx >= len(cells) {
		return false, nil
	}
	raw := strings.TrimSpace(cells[idx].Value)
	if raw == "" {
		return false, nil
	}
	b, err := strconv.ParseBool(raw)
	if err != nil {
		return false, fmt.Errorf("row %d: parse silence_lint %q: %w", rowIdx, raw, err)
	}
	return b, nil
}

// parseScanItems converts a godog.Table into a []scan.Item. Columns:
// name (required), group (required), silence_lint (optional; defaults
// to false). The "(empty)" sentinel in a name or group column yields
// the empty string — used by the D-03 empty-Name validation scenario.
//
// Backs the "the scan items" Given step in scan.feature and
// suppression.feature.
func parseScanItems(table *godog.Table) ([]scan.Item, error) {
	if table == nil || len(table.Rows) == 0 {
		return nil, fmt.Errorf("scan-items table needs a header row")
	}
	// Header-only table (e.g. empty-items-slice scenario) returns an
	// empty []scan.Item without error. scan.Check accepts zero-item
	// input per documented contract — exercising that path requires
	// the BDD to express "no items" explicitly.
	if len(table.Rows) == 1 {
		if _, err := resolveScanItemColumns(table.Rows[0].Cells); err != nil {
			return nil, err
		}
		return nil, nil
	}
	cols, err := resolveScanItemColumns(table.Rows[0].Cells)
	if err != nil {
		return nil, err
	}
	items := make([]scan.Item, 0, len(table.Rows)-1)
	for ri, row := range table.Rows[1:] {
		item, err := parseScanItemRow(ri, row.Cells, cols)
		if err != nil {
			return nil, err
		}
		items = append(items, item)
	}
	return items, nil
}

// decodeEmptySentinel maps the "(empty)" string literal to "" so
// Gherkin tables can express empty-name rows for the D-03 validation
// scenario. Every other input passes through unchanged.
func decodeEmptySentinel(s string) string {
	if s == "(empty)" {
		return ""
	}
	return s
}

// parseScanKind maps a Gherkin step label ("WithinGroup",
// "AcrossGroups") to the corresponding scan.Kind constant.
func parseScanKind(s string) (scan.Kind, error) {
	switch s {
	case "WithinGroup":
		return scan.KindWithinGroup, nil
	case "AcrossGroups":
		return scan.KindAcrossGroups, nil
	default:
		return 0, fmt.Errorf("unknown scan.Kind label: %q", s)
	}
}

// scanWarningLess returns true when prev sorts strictly before cur on
// the canonical (Kind, NameA, NameB, GroupA, GroupB) 5-tuple.
func scanWarningLess(prev, cur scan.Warning) bool {
	if prev.Kind != cur.Kind {
		return prev.Kind < cur.Kind
	}
	if prev.NameA != cur.NameA {
		return prev.NameA < cur.NameA
	}
	if prev.NameB != cur.NameB {
		return prev.NameB < cur.NameB
	}
	if prev.GroupA != cur.GroupA {
		return prev.GroupA < cur.GroupA
	}
	return prev.GroupB < cur.GroupB
}

// scanWarningEqualKey returns true when prev and cur share the full
// (Kind, NameA, NameB, GroupA, GroupB) sort key. D-06 guarantees this
// can never happen on valid input, but the predicate is needed by the
// adjacency walk so the "sorted" check accepts equal-keyed adjacent
// entries when they exist.
func scanWarningEqualKey(prev, cur scan.Warning) bool {
	return prev.Kind == cur.Kind &&
		prev.NameA == cur.NameA &&
		prev.NameB == cur.NameB &&
		prev.GroupA == cur.GroupA &&
		prev.GroupB == cur.GroupB
}

// renderScanWarnings produces a stable, human-readable rendering of a
// scan.Warning slice for error messages. Sorted-by-construction
// (Check sorts before returning) so the output is deterministic.
func renderScanWarnings(ws []scan.Warning) string {
	if len(ws) == 0 {
		return "<empty>"
	}
	var b strings.Builder
	b.WriteString("[")
	for i, w := range ws {
		if i > 0 {
			b.WriteString(", ")
		}
		fmt.Fprintf(&b, "{%s, %q/%q, %q/%q}", w.Kind, w.NameA, w.NameB, w.GroupA, w.GroupB)
	}
	b.WriteString("]")
	return b.String()
}

// scanWarningJSON is a serialisation shape for scan.Warning that
// renders Scores with stringified AlgoID keys so json.Marshal output
// is deterministic across runs. Mirrors the convention used by
// scan/scan_golden_test.go for the cross-platform golden file.
type scanWarningJSON struct {
	Kind   string             `json:"kind"`
	NameA  string             `json:"name_a"`
	NameB  string             `json:"name_b"`
	GroupA string             `json:"group_a"`
	GroupB string             `json:"group_b"`
	Score  float64            `json:"score"`
	Scores map[string]float64 `json:"scores,omitempty"`
}

// marshalScanWarnings renders a []scan.Warning to JSON bytes with
// deterministic Scores key ordering. AlgoID keys are stringified
// (CamelCase form per AlgoID.String()) and the resulting
// map[string]float64 is JSON-marshalled by the stdlib which sorts
// string keys lexicographically — producing byte-identical output
// across runs.
//
// Tag fields are intentionally omitted: scan.Warning.Tag is opaque
// consumer data of type `any`, not safely round-trippable through
// json.Marshal in the general case, and not used by any BDD
// assertion.
func marshalScanWarnings(ws []scan.Warning) ([]byte, error) {
	out := make([]scanWarningJSON, 0, len(ws))
	for _, w := range ws {
		var scores map[string]float64
		if w.Scores != nil {
			scores = make(map[string]float64, len(w.Scores))
			keys := make([]fuzzymatch.AlgoID, 0, len(w.Scores))
			for k := range w.Scores {
				keys = append(keys, k)
			}
			sort.Slice(keys, func(i, j int) bool { return keys[i].String() < keys[j].String() })
			for _, k := range keys {
				scores[k.String()] = w.Scores[k]
			}
		}
		out = append(out, scanWarningJSON{
			Kind:   w.Kind.String(),
			NameA:  w.NameA,
			NameB:  w.NameB,
			GroupA: w.GroupA,
			GroupB: w.GroupB,
			Score:  w.Score,
			Scores: scores,
		})
	}
	return json.Marshal(out)
}

// InitScanSteps registers all scan + suppression step regexes with the
// supplied godog ScenarioContext. Each scenario gets a fresh
// ScanContext via the Before-scenario hook so cross-scenario state
// never leaks.
//
// Called from algorithms_steps.go's InitializeScenario at the bottom
// of the registration block, alongside InitScorerSteps /
// InitValidateSteps / InitNormalisationSteps / InitDeterminismSteps.
func InitScanSteps(ctx *godog.ScenarioContext) {
	sc := &ScanContext{}

	// Given — input setup.
	ctx.Step(
		`^I construct the default Scorer for scan$`,
		sc.iConstructTheDefaultScorerForScan,
	)
	ctx.Step(
		`^the scan items$`,
		sc.theScanItems,
	)
	ctx.Step(
		`^the scan config is the default scan config$`,
		sc.theScanConfigIsTheDefaultScanConfig,
	)
	ctx.Step(
		`^the scan config has a nil Scorer$`,
		sc.theScanConfigHasANilScorer,
	)
	ctx.Step(
		`^I set CompareAcrossGroups to (true|false)$`,
		sc.iSetCompareAcrossGroupsTo,
	)
	ctx.Step(
		`^I set CompareIdenticalAcrossGroups to (true|false)$`,
		sc.iSetCompareIdenticalAcrossGroupsTo,
	)
	ctx.Step(
		`^the scan config has CrossGroupThresholdBoost set to NaN$`,
		sc.theScanConfigHasCrossGroupThresholdBoostSetToNaN,
	)
	ctx.Step(
		`^the scan config has CrossGroupThresholdBoost set to (-?\d+\.?\d*)$`,
		sc.theScanConfigHasCrossGroupThresholdBoostSetTo,
	)
	ctx.Step(
		`^I set suppressed pairs$`,
		sc.iSetSuppressedPairs,
	)

	// When — Check invocation.
	ctx.Step(
		`^I invoke scan\.Check$`,
		sc.iInvokeScanCheck,
	)
	ctx.Step(
		`^I invoke scan\.Check twice$`,
		sc.iInvokeScanCheckTwice,
	)

	// Then — assertions.
	ctx.Step(
		`^scan\.Check returns no error$`,
		sc.scanCheckReturnsNoError,
	)
	ctx.Step(
		`^scan\.Check returns an error matching scan\.(ErrInvalidItem|ErrInvalidConfig|ErrNilScorer)$`,
		sc.scanCheckReturnsAnErrorMatching,
	)
	ctx.Step(
		`^the error mentions index (\d+) and index (\d+)$`,
		sc.theErrorMentionsIndexAndIndex,
	)
	ctx.Step(
		`^the scan warnings include a (WithinGroup|AcrossGroups) pair with names "([^"]*)" and "([^"]*)"$`,
		sc.theScanWarningsIncludeAPairWithKindAndNames,
	)
	ctx.Step(
		`^the scan warnings list has (\d+) entries$`,
		sc.theScanWarningsListHasEntries,
	)
	ctx.Step(
		`^the scan warnings list has (\d+) (WithinGroup|AcrossGroups) entries$`,
		sc.theScanWarningsListHasKindEntries,
	)
	ctx.Step(
		`^no scan warning involves "([^"]*)"$`,
		sc.noWarningInvolves,
	)
	ctx.Step(
		`^the scan warnings are sorted by \(Kind, NameA, NameB, GroupA, GroupB\)$`,
		sc.theWarningsAreSortedByTheCanonical5TupleKey,
	)
	ctx.Step(
		`^the two scan warnings slices are byte-identical$`,
		sc.theTwoWarningsSlicesAreByteIdentical,
	)
	ctx.Step(
		`^the effective cross-group threshold equals ([0-9]+\.?[0-9]*)$`,
		sc.theEffectiveCrossGroupThresholdEquals,
	)
	ctx.Step(
		`^the scan warnings match the naive-loop reference output$`,
		sc.theWarningsMatchTheNaiveReferenceOutput,
	)
}
