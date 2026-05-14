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

// algorithms_golden_test.go pins the score output of all Phase 2 algorithms
// byte-for-byte across the 5-platform CI matrix. It uses the LOCKED canonical
// marshal form established in plan 01-04 (encoding/json MarshalIndent with
// two-space indent, single trailing LF, no BOM, UTF-8).
//
// This file also defines the canonical assertGoldenStaging helper consumed by
// Wave 2 plans (02-02 through 02-06). The helper signature is LOCKED for
// Phase 2: do not change it without coordinated updates to all Wave 2 plans.
//
// Schema of testdata/golden/algorithms.json mirrors testdata/golden/
// normalisation.json: a version field plus an entries array sorted
// alphabetically by Name (deterministic key order per determinism-standards).
//
// Wave 3 (plan 02-07) merges all six per-algorithm staging files into the
// canonical algorithms.json. Until then, algorithms.json contains only
// Levenshtein entries (Wave 1 output), and Wave 2 plans each write to
// testdata/golden/_staging/<algo>.json.

package fuzzymatch_test

import (
	"bytes"
	"os"
	"path/filepath"
	"sort"
	"testing"

	"github.com/axonops/fuzzymatch"
)

// goldenAlgorithmEntry is one (algorithm, a, b, expected_score) case in
// testdata/golden/algorithms.json. Field names are part of the LOCKED JSON
// contract — renaming any field is a major-version-bump event.
type goldenAlgorithmEntry struct {
	Name          string  `json:"name"`
	Algorithm     string  `json:"algorithm"`
	A             string  `json:"a"`
	B             string  `json:"b"`
	ExpectedScore float64 `json:"expected_score"`
}

// goldenAlgorithmsFile wraps entries with a version field. The schema matches
// the normalisation golden file structure (version + entries sorted by Name).
type goldenAlgorithmsFile struct {
	Version int                    `json:"version"`
	Entries []goldenAlgorithmEntry `json:"entries"`
}

// assertGoldenStaging is the Phase-2-canonical write helper for per-algorithm
// staging golden files. Wave 2 plans (02-02 through 02-06) call this helper
// directly — there is NO "if helper exists / else create" branch in any Wave 2
// plan. The signature is LOCKED for Phase 2: do not change the parameter list
// without coordinated updates to all Wave 2 plans.
//
// relPath is the path relative to testdata/golden/, e.g. "_staging/hamming.json".
// v is any value implementing the goldenAlgorithmsFile schema.
//
// Behaviour mirrors assertGolden:
//   - With -update flag: marshal v via CanonicalMarshalForTest and write
//     through WriteGoldenFile to testdata/golden/<relPath>.
//   - Without -update flag: read the existing file and assert byte-equality
//     with the marshalled v.
func assertGoldenStaging(t *testing.T, relPath string, v any) {
	t.Helper()
	got, err := fuzzymatch.CanonicalMarshalForTest(v)
	if err != nil {
		t.Fatalf("assertGoldenStaging: canonicalMarshal: %v", err)
	}
	path := filepath.Join("testdata", "golden", relPath)
	if *updateGolden {
		// Ensure the parent directory exists (e.g. testdata/golden/_staging/).
		if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil { //nolint:gosec // G301: test fixture directory
			t.Fatalf("assertGoldenStaging: MkdirAll(%s): %v", filepath.Dir(path), err)
		}
		if err := fuzzymatch.WriteGoldenFile(path, v); err != nil {
			t.Fatalf("assertGoldenStaging: WriteGoldenFile(%s): %v", path, err)
		}
		t.Logf("assertGoldenStaging: updated %s", path)
		return
	}
	want, err := os.ReadFile(path) //nolint:gosec // path is a fixed test-fixture join, not consumer input
	if err != nil {
		t.Fatalf("assertGoldenStaging: read %s: %v (regenerate with `go test -run TestGolden_ -update ./...`)", path, err)
	}
	if !bytes.Equal(got, want) {
		t.Errorf("assertGoldenStaging: %s mismatch.\n--- got (len=%d) ---\n%s\n--- want (len=%d) ---\n%s\n--- end ---\nRegenerate with `go test -run TestGolden_ -update ./...` after verifying the diff is intentional.",
			path, len(got), truncateForLog(got), len(want), truncateForLog(want))
	}
}

// buildAlgorithmGoldenEntries returns the four Levenshtein entries used by
// both TestGolden_Algorithms (canonical file) and TestGolden_Levenshtein_Staging
// (per-algorithm staging file). ExpectedScore is computed from the current
// implementation so the golden file stays in sync with actual output.
//
// Wave 3 (plan 02-07) replaces this function with a multi-source-file-merging
// form that reads all six _staging/<algo>.json files. For Wave 1, this
// function is the seed.
func buildAlgorithmGoldenEntries(t *testing.T) []goldenAlgorithmEntry {
	t.Helper()
	return []goldenAlgorithmEntry{
		{
			Name:          "Levenshtein_empty_empty",
			Algorithm:     "Levenshtein",
			A:             "",
			B:             "",
			ExpectedScore: fuzzymatch.LevenshteinScore("", ""),
		},
		{
			Name:          "Levenshtein_identical",
			Algorithm:     "Levenshtein",
			A:             "abc",
			B:             "abc",
			ExpectedScore: fuzzymatch.LevenshteinScore("abc", "abc"),
		},
		{
			Name:          "Levenshtein_kitten_sitting",
			Algorithm:     "Levenshtein",
			A:             "kitten",
			B:             "sitting",
			ExpectedScore: fuzzymatch.LevenshteinScore("kitten", "sitting"),
		},
		{
			Name:          "Levenshtein_saturday_sunday",
			Algorithm:     "Levenshtein",
			A:             "saturday",
			B:             "sunday",
			ExpectedScore: fuzzymatch.LevenshteinScore("saturday", "sunday"),
		},
	}
}

// TestGolden_Algorithms pins score outputs across the CI matrix platforms.
// Run with `-update` to rewrite testdata/golden/algorithms.json.
// Re-running without `-update` must exit 0 (file is byte-stable).
func TestGolden_Algorithms(t *testing.T) {
	entries := buildAlgorithmGoldenEntries(t)
	sort.Slice(entries, func(i, j int) bool {
		return entries[i].Name < entries[j].Name
	})
	file := goldenAlgorithmsFile{Version: 1, Entries: entries}
	assertGolden(t, "algorithms.json", file)
}

// TestGolden_Levenshtein_Staging produces testdata/golden/_staging/levenshtein.json
// so that plan 02-07's merge step has uniform inputs across all six algorithms
// (no Levenshtein special case). The staging file contains the same four
// Levenshtein entries as algorithms.json, sorted alphabetically by Name.
//
// Run with `-update` to create or refresh the staging file.
// Re-running without `-update` must exit 0 (file is byte-stable).
func TestGolden_Levenshtein_Staging(t *testing.T) {
	entries := buildAlgorithmGoldenEntries(t) // same four Levenshtein entries
	sort.Slice(entries, func(i, j int) bool {
		return entries[i].Name < entries[j].Name
	})
	file := goldenAlgorithmsFile{Version: 1, Entries: entries}
	assertGoldenStaging(t, "_staging/levenshtein.json", file)
}
