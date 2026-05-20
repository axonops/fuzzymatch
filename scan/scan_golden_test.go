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

// scan_golden_test.go pins the byte-form output of scan.Check across the
// five-platform CI matrix (linux/amd64, linux/arm64, darwin/amd64,
// darwin/arm64, windows/amd64). It is the load-bearing cross-platform
// float-determinism gate for Phase 9 per 09-CONTEXT.md §1 + Plan 09-06
// VALIDATION.md.
//
// The schema follows the LOCKED canonical-form contract established at
// the root package (golden_canonical.go): json.MarshalIndent with prefix
// "" and indent "  ", a single trailing "\n" line terminator, UTF-8 with
// no BOM. The helper is re-implemented locally in this file because the
// root package's writeGoldenFile / canonicalMarshal helpers live behind
// _test.go-only re-exports that are visible only to fuzzymatch_test;
// scan_test (this file's package) is a sibling test package and cannot
// import them. The duplication is a few lines and keeps the contract
// byte-identical to the root harness.
//
// `scores` map keys in the JSON are AlgoID.String() values (e.g.
// "DamerauLevenshteinOSA", NOT integer-form "1") per the same
// convention used in testdata/golden/scorer-default.json. The
// public-API map[AlgoID]float64 is converted to map[string]float64 at
// golden-construction time so the on-disk JSON stays grep-able. Map
// iteration order is randomised by Go but encoding/json sorts map keys
// alphabetically on marshal — the on-disk bytes are stable regardless
// of insertion order here.
//
// _metadata.generated_at is INTENTIONALLY OMITTED — mirroring the
// scorer-default.json / algorithms.json policy. A wall-clock timestamp
// would make every -update regen produce a different file even when
// the underlying warnings are unchanged, breaking the byte-identical
// CI matrix gate.
//
// The -update flag is shared with the root golden tests; the scan
// golden test reuses the existing `go test -run TestGolden_ -update ./...`
// workflow via a local -update flag with the same name. Go's flag
// package allows the same -update flag to be declared in two packages
// because the test binaries are separate. The flag is checked at
// TestMain time (here: at flag.Parse() which runs before tests via
// testing.M).
//
// Run to refresh both the final and staging fixtures:
//
//	go test -run "TestGolden_ScanDefault|TestGolden_ScanDefault_Staging" -update ./scan/...
//
// Re-run without `-update` to verify the gate is green:
//
//	go test -run "TestGolden_ScanDefault|TestGolden_ScanDefault_Staging" ./scan/...

package scan_test

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/axonops/fuzzymatch"
	"github.com/axonops/fuzzymatch/scan"
)

// updateGoldenScan, when set, rewrites the scan golden files from the
// current code output instead of asserting equality. The flag is
// declared at scan_test scope (the root package has its own -update
// flag); both can coexist because Go test binaries are separate per
// package.
var updateGoldenScan = flag.Bool("update", false, "rewrite testdata/golden/scan-*.json fixtures from current code output instead of asserting equality")

// scanGoldenItem mirrors scan.Item but with omitempty serialisation so
// the JSON envelope is grep-able. Tag is intentionally omitted from the
// JSON serialisation entirely — every entry in the corpus uses Tag ==
// nil, and a nil Tag in JSON would be encoded as `null` which is more
// noise than signal.
type scanGoldenItem struct {
	Name        string `json:"name"`
	Group       string `json:"group"`
	SilenceLint bool   `json:"silence_lint,omitempty"`
}

// scanGoldenWarning mirrors scan.Warning's JSON-friendly projection.
// Tag is omitted (per the threat-model T-09-06-02 + Tag-not-stringified
// discipline). Scores uses map[string]float64 (AlgoID.String() keys)
// for grep-ability; encoding/json sorts keys alphabetically on marshal.
type scanGoldenWarning struct {
	Kind   string             `json:"kind"`
	NameA  string             `json:"nameA"`
	NameB  string             `json:"nameB"`
	GroupA string             `json:"groupA"`
	GroupB string             `json:"groupB"`
	Score  float64            `json:"score"`
	Scores map[string]float64 `json:"scores"`
}

// scanGoldenEntry captures one (config-label, items, warnings) triple
// — the corpus entry shape. Config is the human-readable label of the
// scan.Config composition (e.g. "DefaultConfig" or
// "DefaultConfig_CrossEnabled") used downstream by reviewers to grep
// by composition.
type scanGoldenEntry struct {
	Config   string              `json:"config"`
	Items    []scanGoldenItem    `json:"items"`
	Warnings []scanGoldenWarning `json:"warnings"`
}

// scanGoldenMetadata mirrors the scorer-default.json envelope. NO
// generated_at field — see file header for rationale.
type scanGoldenMetadata struct {
	Phase            int    `json:"phase"`
	ScannerSignature string `json:"scanner_signature"`
}

// scanGoldenFile is the top-level JSON envelope: metadata then a
// deterministically-ordered slice of entries.
type scanGoldenFile struct {
	Metadata scanGoldenMetadata `json:"_metadata"`
	Entries  []scanGoldenEntry  `json:"entries"`
}

// canonicalMarshalScan serialises v with the LOCKED v1.x canonical
// byte form: json.MarshalIndent with two-space indent plus a single
// trailing "\n". Mirrors the root package's canonicalMarshal byte-for-
// byte (different file, same contract).
func canonicalMarshalScan(v any) ([]byte, error) {
	body, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		return nil, fmt.Errorf("scan: canonicalMarshalScan: %w", err)
	}
	out := make([]byte, len(body)+1)
	copy(out, body)
	out[len(body)] = '\n'
	return out, nil
}

// writeGoldenFileScan writes v via canonicalMarshalScan to the supplied
// path. Mirrors the root package's writeGoldenFile.
func writeGoldenFileScan(t *testing.T, path string, v any) {
	t.Helper()
	body, err := canonicalMarshalScan(v)
	if err != nil {
		t.Fatalf("writeGoldenFileScan: %v", err)
	}
	// Ensure the parent directory exists (the _staging file lives under
	// testdata/golden/_staging/ which is committed but the helper is
	// defensive against a freshly-cloned checkout where the directory
	// was somehow purged).
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil { //nolint:gosec // G301: test fixture directory
		t.Fatalf("writeGoldenFileScan: MkdirAll(%s): %v", filepath.Dir(path), err)
	}
	if err := os.WriteFile(path, body, 0o644); err != nil { //nolint:gosec // G306: test fixture, world-readable by design
		t.Fatalf("writeGoldenFileScan: WriteFile(%s): %v", path, err)
	}
	t.Logf("writeGoldenFileScan: wrote %s", path)
}

// scoreAllAsStringKeysScan converts the public-API map[AlgoID]float64
// returned by Scorer.ScoreAll (and re-exposed in scan.Warning.Scores)
// into a map[string]float64 keyed by AlgoID.String(). Mirrors the root
// package's scoreAllAsStringKeys helper.
func scoreAllAsStringKeysScan(m map[fuzzymatch.AlgoID]float64) map[string]float64 {
	if m == nil {
		return nil
	}
	out := make(map[string]float64, len(m))
	for id, score := range m {
		out[id.String()] = score
	}
	return out
}

// projectWarning converts a scan.Warning into its JSON-friendly
// scanGoldenWarning projection. Tag values are intentionally NOT
// projected (T-09-06-02 mitigation; consistent with the on-disk
// schema's omission of Tag fields).
func projectWarning(w scan.Warning) scanGoldenWarning {
	return scanGoldenWarning{
		Kind:   w.Kind.String(),
		NameA:  w.NameA,
		NameB:  w.NameB,
		GroupA: w.GroupA,
		GroupB: w.GroupB,
		Score:  w.Score,
		Scores: scoreAllAsStringKeysScan(w.Scores),
	}
}

// projectItem converts a scan.Item into its JSON-friendly form. Tag is
// intentionally omitted.
func projectItem(it scan.Item) scanGoldenItem {
	return scanGoldenItem{
		Name:        it.Name,
		Group:       it.Group,
		SilenceLint: it.SilenceLint,
	}
}

// runScanEntry executes scan.Check on the supplied items + cfg and
// returns a fully-populated scanGoldenEntry. Both items and warnings
// are converted to their JSON-friendly projections at this layer; the
// production Check output is preserved verbatim (Plan 09-06 sort +
// canonicalisation already applied inside Check).
func runScanEntry(t *testing.T, configLabel string, items []scan.Item, cfg scan.Config) scanGoldenEntry {
	t.Helper()
	warnings, err := scan.Check(items, cfg)
	if err != nil {
		t.Fatalf("runScanEntry(%s): %v", configLabel, err)
	}
	projItems := make([]scanGoldenItem, len(items))
	for i, it := range items {
		projItems[i] = projectItem(it)
	}
	projWarnings := make([]scanGoldenWarning, len(warnings))
	for i, w := range warnings {
		projWarnings[i] = projectWarning(w)
	}
	return scanGoldenEntry{
		Config:   configLabel,
		Items:    projItems,
		Warnings: projWarnings,
	}
}

// buildScanGolden assembles the canonical scan corpus used by both
// TestGolden_ScanDefault and TestGolden_ScanDefault_Staging.
//
// Five entries cover the major code paths:
//
//  1. DefaultConfig — within-group only baseline.
//  2. DefaultConfig_CrossEnabled — within + cross with the SCAN-04
//     identical-cross suppression default firing.
//  3. DefaultConfig_CrossEnabled_AllowIdentical — within + cross with
//     identical-cross unblocked (more warnings emit).
//  4. DefaultConfig_WithSuppressedPair — SuppressedPairs silences a
//     canonical pair.
//  5. DefaultConfig_WithSilenceLint — a per-item SilenceLint silences
//     every warning involving that item.
//
// All five entries share the same items[] shape so reviewers can spot
// the cross-config divergences without re-reading the items in each
// entry — but Entry 5 swaps in one item with SilenceLint=true.
func buildScanGolden(t *testing.T) scanGoldenFile {
	t.Helper()
	s := fuzzymatch.DefaultScorer()

	// Canonical items: 3 in "login" group, 2 in "profile" group.
	// user_id / userId match within-group; user_id / user_id match
	// cross-group (suppressed by SCAN-04 default).
	baseItems := []scan.Item{
		{Name: "user_id", Group: "login"},
		{Name: "userId", Group: "login"},
		{Name: "user_name", Group: "login"},
		{Name: "user_id", Group: "profile"},
		{Name: "userId", Group: "profile"},
	}

	// Entry 1: within-only baseline.
	entry1 := runScanEntry(t, "DefaultConfig", baseItems, scan.DefaultConfig(s))

	// Entry 2: cross enabled, identical-cross suppressed (the
	// SCAN-04 default).
	cfg2 := scan.DefaultConfig(s)
	cfg2.CompareAcrossGroups = true
	entry2 := runScanEntry(t, "DefaultConfig_CrossEnabled", baseItems, cfg2)

	// Entry 3: cross enabled, identical-cross unblocked.
	cfg3 := scan.DefaultConfig(s)
	cfg3.CompareAcrossGroups = true
	cfg3.CompareIdenticalAcrossGroups = true
	entry3 := runScanEntry(t, "DefaultConfig_CrossEnabled_AllowIdentical", baseItems, cfg3)

	// Entry 4: SuppressedPairs silences the user_id / userId canonical
	// pair (which would otherwise be the load-bearing within-group
	// warning).
	cfg4 := scan.DefaultConfig(s)
	cfg4.SuppressedPairs = [][2]string{{"user_id", "userId"}}
	entry4 := runScanEntry(t, "DefaultConfig_WithSuppressedPair", baseItems, cfg4)

	// Entry 5: SilenceLint=true on items[0] (login/user_id) silences
	// every warning involving that item — one-sided suppression.
	silencedItems := []scan.Item{
		{Name: "user_id", Group: "login", SilenceLint: true},
		{Name: "userId", Group: "login"},
		{Name: "user_name", Group: "login"},
		{Name: "user_id", Group: "profile"},
		{Name: "userId", Group: "profile"},
	}
	entry5 := runScanEntry(t, "DefaultConfig_WithSilenceLint", silencedItems, scan.DefaultConfig(s))

	// Entry 6: 60-item single group — exercises the bucket dispatch
	// path (60 > bucketThreshold=50). Closes the determinism +
	// test-analyst gap on Plan 09-06 where the original 5-entry
	// corpus exercised only the naive path. Items mirror the
	// canonical user_id/userId family with disambiguating suffixes
	// so every pair scores at the within-group threshold.
	bucketItems := make([]scan.Item, 0, 60)
	for i := 0; i < 30; i++ {
		bucketItems = append(bucketItems,
			scan.Item{Name: "user_id_" + bucketGoldenSuffix(i), Group: "login"},
			scan.Item{Name: "userId_" + bucketGoldenSuffix(i), Group: "login"},
		)
	}
	entry6 := runScanEntry(t, "DefaultConfig_BucketDispatch_60Items", bucketItems, scan.DefaultConfig(s))

	return scanGoldenFile{
		Metadata: scanGoldenMetadata{
			Phase:            9,
			ScannerSignature: "DefaultConfig-2026-05-20",
		},
		Entries: []scanGoldenEntry{entry1, entry2, entry3, entry4, entry5, entry6},
	}
}

// bucketGoldenSuffix returns a 2-digit zero-padded suffix for the
// bucket-dispatch golden entry's item names. Keeps generated names
// lexicographically stable across the 60-item corpus.
func bucketGoldenSuffix(i int) string {
	const digits = "0123456789"
	return string([]byte{digits[(i/10)%10], digits[i%10]})
}

// assertScanGolden writes (with -update) or asserts byte-equality
// (without -update) of the canonical-marshalled scanGoldenFile against
// the file at path.
func assertScanGolden(t *testing.T, path string, v scanGoldenFile) {
	t.Helper()
	got, err := canonicalMarshalScan(v)
	if err != nil {
		t.Fatalf("assertScanGolden: %v", err)
	}
	if *updateGoldenScan {
		writeGoldenFileScan(t, path, v)
		return
	}
	want, err := os.ReadFile(path) //nolint:gosec // path is a fixed test-fixture join, not consumer input
	if err != nil {
		t.Fatalf("assertScanGolden: read %s: %v (regenerate with `go test -run TestGolden_Scan -update ./scan/...`)", path, err)
	}
	if !bytes.Equal(got, want) {
		// Truncate large outputs for log readability — five entries
		// with full ScoreAll maps push the golden file into the tens
		// of KB.
		const limit = 1024
		gotExcerpt := got
		wantExcerpt := want
		if len(gotExcerpt) > limit {
			gotExcerpt = gotExcerpt[:limit]
		}
		if len(wantExcerpt) > limit {
			wantExcerpt = wantExcerpt[:limit]
		}
		t.Errorf("assertScanGolden: %s mismatch.\n--- got (len=%d) ---\n%s\n--- want (len=%d) ---\n%s\n--- end ---\nRegenerate with `go test -run TestGolden_Scan -update ./scan/...` after verifying the diff is intentional.",
			path, len(got), gotExcerpt, len(want), wantExcerpt)
	}
}

// TestGolden_ScanDefault is the cross-platform determinism gate for
// the Phase 9 scan sub-package. It constructs the canonical five-entry
// corpus via buildScanGolden, runs scan.Check on each configuration,
// and asserts the canonical-form JSON bytes match
// testdata/golden/scan-default.json exactly.
//
// Run with `-update` to regenerate the golden file:
//
//	go test -run TestGolden_ScanDefault -update ./scan/...
//
// Re-running without `-update` MUST exit 0 on every platform in the CI
// matrix. Any byte-level divergence (Scorer float drift, scan sort
// regression, canonicalisation flip, etc.) fails CI on that platform.
//
// The TestGolden_ prefix is picked up automatically by
// `make verify-determinism` (which runs `go test -run TestGolden_ ./...`).
// No Makefile edit is required.
func TestGolden_ScanDefault(t *testing.T) {
	file := buildScanGolden(t)
	path := filepath.Join("..", "testdata", "golden", "scan-default.json")
	assertScanGolden(t, path, file)
}

// TestGolden_ScanDefault_Staging exercises the staging-file convention
// alongside the final fixture. The staging file lives under
// testdata/golden/_staging/scan.json and is the pre-merge form per the
// existing per-algorithm staging pattern (algorithms_golden_test.go
// Wave 2). For Plan 09-06 the staging file is byte-identical to the
// final fixture — the convention is preserved here for forward
// compatibility with any future merge step that might split the corpus
// across multiple staging files (e.g. one per major code path).
func TestGolden_ScanDefault_Staging(t *testing.T) {
	file := buildScanGolden(t)
	path := filepath.Join("..", "testdata", "golden", "_staging", "scan.json")
	assertScanGolden(t, path, file)
}

// TestGolden_ScanDefault_CorpusInvariants asserts the in-corpus
// invariants that the golden file alone cannot enforce:
//
//   - buildScanGolden is deterministic across consecutive calls
//     (entry count + projected warnings byte-for-byte equal).
//   - Every emitted Warning has NameA <= NameB on raw byte lex
//     (Plan 09-06 canonicalisation invariant).
//
// In-entry sort-order monotonicity is NOT checked here — the projected
// Kind field is the CamelCase string, and lex order on strings
// ("AcrossGroups" < "WithinGroup") disagrees with the int-form Kind
// sort key (KindWithinGroup=1 < KindAcrossGroups=2). The post-sort
// production order is tested directly in TestCheck_SortKey_KindFirst
// in scan_test.go against the integer Kind values; the golden file
// itself is the byte-stable record of that order.
func TestGolden_ScanDefault_CorpusInvariants(t *testing.T) {
	t.Parallel()
	file1 := buildScanGolden(t)
	file2 := buildScanGolden(t)
	if len(file1.Entries) != len(file2.Entries) {
		t.Fatalf("entry count differs across builds: %d vs %d", len(file1.Entries), len(file2.Entries))
	}
	// Byte-for-byte determinism across builds (mirrors the canonical
	// form gate on disk).
	b1, err := canonicalMarshalScan(file1)
	if err != nil {
		t.Fatalf("canonicalMarshalScan(file1): %v", err)
	}
	b2, err := canonicalMarshalScan(file2)
	if err != nil {
		t.Fatalf("canonicalMarshalScan(file2): %v", err)
	}
	if !bytes.Equal(b1, b2) {
		t.Errorf("buildScanGolden non-deterministic: byte outputs differ")
	}
	for ei, e := range file1.Entries {
		// Lex canonicalisation invariant: every warning has NameA <=
		// NameB on raw byte lex compare.
		for wi, w := range e.Warnings {
			if w.NameA > w.NameB {
				t.Errorf("entry[%d].Warning[%d] not lex-canonical: NameA=%q > NameB=%q", ei, wi, w.NameA, w.NameB)
			}
		}
	}
}
