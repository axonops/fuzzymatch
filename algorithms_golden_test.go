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
// Wave 3 (plan 02-07) owns the merge step: TestGolden_Algorithms_Merge reads
// all six per-algorithm staging files from testdata/golden/_staging/ and
// produces the canonical testdata/golden/algorithms.json. The per-algorithm
// TestGolden_<algo>_Staging helpers (Wave 1 + Wave 2) remain for the
// per-algorithm determinism gate and as the audit trail of each algorithm's
// contribution.

package fuzzymatch_test

import (
	"bytes"
	"encoding/json"
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

// TestGolden_Algorithms_Merge reads all six per-algorithm staging files from
// testdata/golden/_staging/ and merges them into the canonical
// testdata/golden/algorithms.json via CanonicalMarshalForTest.
//
// Wave 3 of phase 02 (plan 02-07) owns the merge step — Wave 2 plans each
// wrote one staging file to avoid algorithms.json merge conflicts during
// parallel execution. This function supersedes the Wave 1 TestGolden_Algorithms
// stub (which contained Levenshtein-only entries) and is the canonical merge
// gate going forward.
//
// Run with `-update` to rewrite testdata/golden/algorithms.json.
// Re-running without `-update` must exit 0 (file is byte-stable).
func TestGolden_Algorithms_Merge(t *testing.T) {
	stagingFiles := []string{
		"_staging/damerau_full.json",
		"_staging/damerau_osa.json",
		"_staging/hamming.json",
		"_staging/jaro.json",
		"_staging/jarowinkler.json",
		"_staging/lcsstr.json",
		"_staging/levenshtein.json",
		"_staging/ratcliff_obershelp.json",
		"_staging/strcmp95.json",
		"_staging/swg.json",
	}
	var allEntries []goldenAlgorithmEntry
	for _, f := range stagingFiles {
		raw, err := os.ReadFile(filepath.Join("testdata", "golden", f)) //nolint:gosec // path is a fixed test-fixture join
		if err != nil {
			t.Fatalf("TestGolden_Algorithms_Merge: read %s: %v", f, err)
		}
		var staged goldenAlgorithmsFile
		if err := json.Unmarshal(raw, &staged); err != nil {
			t.Fatalf("TestGolden_Algorithms_Merge: parse %s: %v", f, err)
		}
		allEntries = append(allEntries, staged.Entries...)
	}
	sort.Slice(allEntries, func(i, j int) bool {
		return allEntries[i].Name < allEntries[j].Name
	})
	// Sanity check: no duplicate Names across staging files.
	for i := 1; i < len(allEntries); i++ {
		if allEntries[i].Name == allEntries[i-1].Name {
			t.Fatalf("TestGolden_Algorithms_Merge: duplicate entry Name across staging files: %q", allEntries[i].Name)
		}
	}
	file := goldenAlgorithmsFile{Version: 1, Entries: allEntries}
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

// buildHammingStagingEntries returns the four Hamming entries used by
// TestGolden_Hamming_Staging. ExpectedScore is computed from the current
// implementation so the staging file stays in sync with actual output.
func buildHammingStagingEntries(t *testing.T) []goldenAlgorithmEntry {
	t.Helper()
	return []goldenAlgorithmEntry{
		{
			Name:          "Hamming_empty_empty",
			Algorithm:     "Hamming",
			A:             "",
			B:             "",
			ExpectedScore: fuzzymatch.HammingScore("", ""),
		},
		{
			Name:          "Hamming_identical",
			Algorithm:     "Hamming",
			A:             "abc",
			B:             "abc",
			ExpectedScore: fuzzymatch.HammingScore("abc", "abc"),
		},
		{
			Name:          "Hamming_karolin_kathrin",
			Algorithm:     "Hamming",
			A:             "karolin",
			B:             "kathrin",
			ExpectedScore: fuzzymatch.HammingScore("karolin", "kathrin"),
		},
		{
			Name:          "Hamming_unequal_length",
			Algorithm:     "Hamming",
			A:             "abc",
			B:             "ab",
			ExpectedScore: fuzzymatch.HammingScore("abc", "ab"),
		},
	}
}

// TestGolden_Hamming_Staging produces testdata/golden/_staging/hamming.json
// for plan 02-07's merge step. Entries are sorted alphabetically by Name.
// Contains the four canonical Hamming cases: empty_empty, identical,
// karolin_kathrin, and unequal_length (the locked silent-zero case).
//
// Run with `-update` to create or refresh the staging file.
// Re-running without `-update` must exit 0 (file is byte-stable).
func TestGolden_Hamming_Staging(t *testing.T) {
	entries := buildHammingStagingEntries(t)
	sort.Slice(entries, func(i, j int) bool {
		return entries[i].Name < entries[j].Name
	})
	file := goldenAlgorithmsFile{Version: 1, Entries: entries}
	assertGoldenStaging(t, "_staging/hamming.json", file)
}

// buildJaroStagingEntries returns the six Jaro entries used by
// TestGolden_Jaro_Staging. ExpectedScore is computed from the current
// implementation so the staging file stays in sync with actual output.
//
// Six entries (sorted by Name in the test): DIXON_DICKSONX, empty_empty,
// identical, JELLYFISH_SMELLYFISH, MARTHA_MARHTA, one_empty.
// These cover the canonical Jaro 1989 / Winkler 1990 reference vectors plus
// the edge-case identity and empty-string conventions.
func buildJaroStagingEntries(t *testing.T) []goldenAlgorithmEntry {
	t.Helper()
	return []goldenAlgorithmEntry{
		{
			Name:          "Jaro_DIXON_DICKSONX",
			Algorithm:     "Jaro",
			A:             "DIXON",
			B:             "DICKSONX",
			ExpectedScore: fuzzymatch.JaroScore("DIXON", "DICKSONX"),
		},
		{
			Name:          "Jaro_empty_empty",
			Algorithm:     "Jaro",
			A:             "",
			B:             "",
			ExpectedScore: fuzzymatch.JaroScore("", ""),
		},
		{
			Name:          "Jaro_identical",
			Algorithm:     "Jaro",
			A:             "ABC",
			B:             "ABC",
			ExpectedScore: fuzzymatch.JaroScore("ABC", "ABC"),
		},
		{
			Name:          "Jaro_JELLYFISH_SMELLYFISH",
			Algorithm:     "Jaro",
			A:             "JELLYFISH",
			B:             "SMELLYFISH",
			ExpectedScore: fuzzymatch.JaroScore("JELLYFISH", "SMELLYFISH"),
		},
		{
			Name:          "Jaro_MARTHA_MARHTA",
			Algorithm:     "Jaro",
			A:             "MARTHA",
			B:             "MARHTA",
			ExpectedScore: fuzzymatch.JaroScore("MARTHA", "MARHTA"),
		},
		{
			Name:          "Jaro_one_empty",
			Algorithm:     "Jaro",
			A:             "",
			B:             "ABC",
			ExpectedScore: fuzzymatch.JaroScore("", "ABC"),
		},
	}
}

// TestGolden_Jaro_Staging produces testdata/golden/_staging/jaro.json for
// plan 02-07's merge step. Entries are sorted alphabetically by Name.
// Contains the six canonical Jaro cases: DIXON_DICKSONX, empty_empty,
// identical, JELLYFISH_SMELLYFISH, MARTHA_MARHTA, one_empty.
//
// Run with `-update` to create or refresh the staging file.
// Re-running without `-update` must exit 0 (file is byte-stable).
func TestGolden_Jaro_Staging(t *testing.T) {
	entries := buildJaroStagingEntries(t)
	sort.Slice(entries, func(i, j int) bool {
		return entries[i].Name < entries[j].Name
	})
	file := goldenAlgorithmsFile{Version: 1, Entries: entries}
	assertGoldenStaging(t, "_staging/jaro.json", file)
}

// buildDamerauOSAStagingEntries returns the five DamerauLevenshteinOSA entries
// used by TestGolden_DamerauLevenshteinOSA_Staging. ExpectedScore is computed
// from the current implementation so the staging file stays in sync with actual
// output.
//
// Five entries (sorted by Name in the test):
//   - DamerauLevenshteinOSA_ab_ba       (transposition; OSA distance 1, score 0.5)
//   - DamerauLevenshteinOSA_ca_abc      (discriminating vector; distance 3, score 0.0)
//   - DamerauLevenshteinOSA_empty_empty (both-empty identity; score 1.0)
//   - DamerauLevenshteinOSA_identical   (abc/abc; score 1.0)
//   - DamerauLevenshteinOSA_one_empty   (""/abc; score 0.0)
//
// The ca/abc entry is the locked discriminating-vector gate: if this entry
// ever shows expected_score != 0.0, the recurrence has drifted toward Full DL
// semantics. Wave 3 plan 02-07 will diff this against the DL-Full staging file
// to verify the OSA-vs-Full divergence at the cross-algorithm level.
func buildDamerauOSAStagingEntries(t *testing.T) []goldenAlgorithmEntry {
	t.Helper()
	return []goldenAlgorithmEntry{
		{
			Name:          "DamerauLevenshteinOSA_ab_ba",
			Algorithm:     "DamerauLevenshteinOSA",
			A:             "ab",
			B:             "ba",
			ExpectedScore: fuzzymatch.DamerauLevenshteinOSAScore("ab", "ba"),
		},
		{
			Name:          "DamerauLevenshteinOSA_ca_abc",
			Algorithm:     "DamerauLevenshteinOSA",
			A:             "ca",
			B:             "abc",
			ExpectedScore: fuzzymatch.DamerauLevenshteinOSAScore("ca", "abc"),
		},
		{
			Name:          "DamerauLevenshteinOSA_empty_empty",
			Algorithm:     "DamerauLevenshteinOSA",
			A:             "",
			B:             "",
			ExpectedScore: fuzzymatch.DamerauLevenshteinOSAScore("", ""),
		},
		{
			Name:          "DamerauLevenshteinOSA_identical",
			Algorithm:     "DamerauLevenshteinOSA",
			A:             "abc",
			B:             "abc",
			ExpectedScore: fuzzymatch.DamerauLevenshteinOSAScore("abc", "abc"),
		},
		{
			Name:          "DamerauLevenshteinOSA_one_empty",
			Algorithm:     "DamerauLevenshteinOSA",
			A:             "",
			B:             "abc",
			ExpectedScore: fuzzymatch.DamerauLevenshteinOSAScore("", "abc"),
		},
	}
}

// TestGolden_DamerauLevenshteinOSA_Staging produces
// testdata/golden/_staging/damerau_osa.json for plan 02-07's merge step.
// Entries are sorted alphabetically by Name.
//
// The five entries cover: ab_ba (transposition, distance 1), ca_abc
// (discriminating vector, distance 3, score 0.0), empty_empty, identical,
// one_empty.
//
// Run with `-update` to create or refresh the staging file.
// Re-running without `-update` must exit 0 (file is byte-stable).
func TestGolden_DamerauLevenshteinOSA_Staging(t *testing.T) {
	entries := buildDamerauOSAStagingEntries(t)
	sort.Slice(entries, func(i, j int) bool {
		return entries[i].Name < entries[j].Name
	})
	file := goldenAlgorithmsFile{Version: 1, Entries: entries}
	assertGoldenStaging(t, "_staging/damerau_osa.json", file)
}

// buildDamerauFullStagingEntries returns the five DamerauLevenshteinFull entries
// used by TestGolden_DamerauLevenshteinFull_Staging. ExpectedScore is computed
// from the current implementation so the staging file stays in sync with actual
// output.
//
// Five entries (sorted by Name in the test):
//   - DamerauLevenshteinFull_ab_ba       (transposition; Full distance 1, score 0.5)
//   - DamerauLevenshteinFull_ca_abc      (discriminating vector; distance 2, score ≈0.3333)
//   - DamerauLevenshteinFull_empty_empty (both-empty identity; score 1.0)
//   - DamerauLevenshteinFull_identical   (abc/abc; score 1.0)
//   - DamerauLevenshteinFull_one_empty   (""/abc; score 0.0)
//
// The ca/abc entry is the locked discriminating-vector gate: if this entry
// ever shows expected_score == 0.0, the recurrence has drifted toward OSA
// semantics. Wave 3 plan 02-07 will diff this against the DL-OSA staging file
// to verify the Full-vs-OSA divergence at the cross-algorithm level.
func buildDamerauFullStagingEntries(t *testing.T) []goldenAlgorithmEntry {
	t.Helper()
	return []goldenAlgorithmEntry{
		{
			Name:          "DamerauLevenshteinFull_ab_ba",
			Algorithm:     "DamerauLevenshteinFull",
			A:             "ab",
			B:             "ba",
			ExpectedScore: fuzzymatch.DamerauLevenshteinFullScore("ab", "ba"),
		},
		{
			Name:          "DamerauLevenshteinFull_ca_abc",
			Algorithm:     "DamerauLevenshteinFull",
			A:             "ca",
			B:             "abc",
			ExpectedScore: fuzzymatch.DamerauLevenshteinFullScore("ca", "abc"),
		},
		{
			Name:          "DamerauLevenshteinFull_empty_empty",
			Algorithm:     "DamerauLevenshteinFull",
			A:             "",
			B:             "",
			ExpectedScore: fuzzymatch.DamerauLevenshteinFullScore("", ""),
		},
		{
			Name:          "DamerauLevenshteinFull_identical",
			Algorithm:     "DamerauLevenshteinFull",
			A:             "abc",
			B:             "abc",
			ExpectedScore: fuzzymatch.DamerauLevenshteinFullScore("abc", "abc"),
		},
		{
			Name:          "DamerauLevenshteinFull_one_empty",
			Algorithm:     "DamerauLevenshteinFull",
			A:             "",
			B:             "abc",
			ExpectedScore: fuzzymatch.DamerauLevenshteinFullScore("", "abc"),
		},
	}
}

// TestGolden_DamerauLevenshteinFull_Staging produces
// testdata/golden/_staging/damerau_full.json for plan 02-07's merge step.
// Entries are sorted alphabetically by Name.
//
// The five entries cover: ab_ba (transposition, distance 1, score 0.5),
// ca_abc (discriminating vector, distance 2, score ≈0.3333 — DIFFERENT from
// the DL-OSA staging file where ca_abc has score 0.0), empty_empty, identical,
// one_empty.
//
// Run with `-update` to create or refresh the staging file.
// Re-running without `-update` must exit 0 (file is byte-stable).
func TestGolden_DamerauLevenshteinFull_Staging(t *testing.T) {
	entries := buildDamerauFullStagingEntries(t)
	sort.Slice(entries, func(i, j int) bool {
		return entries[i].Name < entries[j].Name
	})
	file := goldenAlgorithmsFile{Version: 1, Entries: entries}
	assertGoldenStaging(t, "_staging/damerau_full.json", file)
}

// buildJaroWinklerStagingEntries returns the eight JaroWinkler entries used by
// TestGolden_JaroWinkler_Staging. ExpectedScore is computed from the current
// implementation so the staging file stays in sync with actual output.
//
// Eight entries (sorted by Name in the test):
//   - JaroWinkler_below_threshold  (abc/xyz — Jaro = 0.0, below gate; JW == J)
//   - JaroWinkler_DIXON_DICKSONX   (Winkler 1990 reference; JW ≈ 0.8133)
//   - JaroWinkler_DWAYNE_DUANE     (Winkler 1990 reference; JW ≈ 0.8400)
//   - JaroWinkler_empty_empty      (both-empty identity; score 1.0)
//   - JaroWinkler_identical        (ABC/ABC; score 1.0)
//   - JaroWinkler_MARTHA_MARHTA    (Winkler 1990 reference; JW ≈ 0.9611)
//   - JaroWinkler_one_empty        (""/ABC; score 0.0)
//   - JaroWinkler_prefix_cap       (TESTABCD/TESTABCE — 7-char prefix capped at 4)
func buildJaroWinklerStagingEntries(t *testing.T) []goldenAlgorithmEntry {
	t.Helper()
	return []goldenAlgorithmEntry{
		{
			Name:          "JaroWinkler_below_threshold",
			Algorithm:     "JaroWinkler",
			A:             "abc",
			B:             "xyz",
			ExpectedScore: fuzzymatch.JaroWinklerScore("abc", "xyz"),
		},
		{
			Name:          "JaroWinkler_DIXON_DICKSONX",
			Algorithm:     "JaroWinkler",
			A:             "DIXON",
			B:             "DICKSONX",
			ExpectedScore: fuzzymatch.JaroWinklerScore("DIXON", "DICKSONX"),
		},
		{
			Name:          "JaroWinkler_DWAYNE_DUANE",
			Algorithm:     "JaroWinkler",
			A:             "DWAYNE",
			B:             "DUANE",
			ExpectedScore: fuzzymatch.JaroWinklerScore("DWAYNE", "DUANE"),
		},
		{
			Name:          "JaroWinkler_empty_empty",
			Algorithm:     "JaroWinkler",
			A:             "",
			B:             "",
			ExpectedScore: fuzzymatch.JaroWinklerScore("", ""),
		},
		{
			Name:          "JaroWinkler_identical",
			Algorithm:     "JaroWinkler",
			A:             "ABC",
			B:             "ABC",
			ExpectedScore: fuzzymatch.JaroWinklerScore("ABC", "ABC"),
		},
		{
			Name:          "JaroWinkler_MARTHA_MARHTA",
			Algorithm:     "JaroWinkler",
			A:             "MARTHA",
			B:             "MARHTA",
			ExpectedScore: fuzzymatch.JaroWinklerScore("MARTHA", "MARHTA"),
		},
		{
			Name:          "JaroWinkler_one_empty",
			Algorithm:     "JaroWinkler",
			A:             "",
			B:             "ABC",
			ExpectedScore: fuzzymatch.JaroWinklerScore("", "ABC"),
		},
		{
			Name:          "JaroWinkler_prefix_cap",
			Algorithm:     "JaroWinkler",
			A:             "TESTABCD",
			B:             "TESTABCE",
			ExpectedScore: fuzzymatch.JaroWinklerScore("TESTABCD", "TESTABCE"),
		},
	}
}

// TestGolden_JaroWinkler_Staging produces testdata/golden/_staging/jarowinkler.json
// for plan 02-07's merge step. Entries are sorted alphabetically by Name.
//
// Eight entries covering: below_threshold (gate test), DIXON_DICKSONX,
// DWAYNE_DUANE, empty_empty, identical, MARTHA_MARHTA (canonical Winkler 1990
// reference vectors), one_empty, and prefix_cap (L=4 cap verification).
//
// Run with `-update` to create or refresh the staging file.
// Re-running without `-update` must exit 0 (file is byte-stable).
func TestGolden_JaroWinkler_Staging(t *testing.T) {
	entries := buildJaroWinklerStagingEntries(t)
	sort.Slice(entries, func(i, j int) bool {
		return entries[i].Name < entries[j].Name
	})
	file := goldenAlgorithmsFile{Version: 1, Entries: entries}
	assertGoldenStaging(t, "_staging/jarowinkler.json", file)
}

// buildSWGStagingEntries returns the Smith-Waterman-Gotoh entries used by
// TestGolden_SmithWatermanGotoh_Staging. ExpectedScore is computed from the
// current implementation so the staging file stays in sync with actual
// output. Six entries cover: both-empty (1.0), identical (1.0), one-empty
// (0.0), substring-containment (1.0), no-overlap (0.0), and the
// one-long-gap-canary (PITFALLS.md §3 Gotoh-erratum gate).
//
// Plan 03-03 owns the merge into testdata/golden/algorithms.json — this
// plan only writes the staging file.
func buildSWGStagingEntries(t *testing.T) []goldenAlgorithmEntry {
	t.Helper()
	return []goldenAlgorithmEntry{
		{
			Name:          "SmithWatermanGotoh_both_empty",
			Algorithm:     "SmithWatermanGotoh",
			A:             "",
			B:             "",
			ExpectedScore: fuzzymatch.SmithWatermanGotohScore("", ""),
		},
		{
			Name:          "SmithWatermanGotoh_identical",
			Algorithm:     "SmithWatermanGotoh",
			A:             "abc",
			B:             "abc",
			ExpectedScore: fuzzymatch.SmithWatermanGotohScore("abc", "abc"),
		},
		{
			Name:          "SmithWatermanGotoh_no_overlap",
			Algorithm:     "SmithWatermanGotoh",
			A:             "qqqq",
			B:             "zzzz",
			ExpectedScore: fuzzymatch.SmithWatermanGotohScore("qqqq", "zzzz"),
		},
		{
			Name:          "SmithWatermanGotoh_one_empty",
			Algorithm:     "SmithWatermanGotoh",
			A:             "abc",
			B:             "",
			ExpectedScore: fuzzymatch.SmithWatermanGotohScore("abc", ""),
		},
		{
			Name:          "SmithWatermanGotoh_one_long_gap_canary",
			Algorithm:     "SmithWatermanGotoh",
			A:             "abc________def",
			B:             "abcdef",
			ExpectedScore: fuzzymatch.SmithWatermanGotohScore("abc________def", "abcdef"),
		},
		{
			Name:          "SmithWatermanGotoh_two_substring",
			Algorithm:     "SmithWatermanGotoh",
			A:             "http_request",
			B:             "http_request_header_fields",
			ExpectedScore: fuzzymatch.SmithWatermanGotohScore("http_request", "http_request_header_fields"),
		},
	}
}

// TestGolden_SmithWatermanGotoh_Staging produces
// testdata/golden/_staging/swg.json for plan 03-03's merge step. Entries are
// sorted alphabetically by Name. Six entries covering: both-empty, identical,
// no-overlap, one-empty, one-long-gap-canary (Gotoh-erratum gate), and the
// canonical substring-containment pair.
//
// Run with `-update` to create or refresh the staging file.
// Re-running without `-update` must exit 0 (file is byte-stable).
func TestGolden_SmithWatermanGotoh_Staging(t *testing.T) {
	entries := buildSWGStagingEntries(t)
	sort.Slice(entries, func(i, j int) bool {
		return entries[i].Name < entries[j].Name
	})
	file := goldenAlgorithmsFile{Version: 1, Entries: entries}
	assertGoldenStaging(t, "_staging/swg.json", file)
}

// buildStrcmp95StagingEntries returns the Strcmp95 entries used by
// TestGolden_Strcmp95_Staging. ExpectedScore is computed from the current
// implementation so the staging file stays in sync with actual output.
// Seven entries cover: both-empty (1.0), identical (1.0), one-empty (0.0),
// Winkler 1990 / Census Bureau canonical surnames (MARTHA/MARHTA,
// DWAYNE/DUANE, DIXON/DICKSONX) — the latter two trigger the similar-
// character table per RESEARCH.md Pitfall 1 — AND the long-string adjustment
// trigger pair (HAMINGTON/HAMMINGTON).
//
// Plan 04-05 owns the merge into testdata/golden/algorithms.json — this
// plan only writes the staging file.
func buildStrcmp95StagingEntries(t *testing.T) []goldenAlgorithmEntry {
	t.Helper()
	return []goldenAlgorithmEntry{
		{
			Name:          "Strcmp95_DIXON_DICKSONX",
			Algorithm:     "Strcmp95",
			A:             "DIXON",
			B:             "DICKSONX",
			ExpectedScore: fuzzymatch.Strcmp95Score("DIXON", "DICKSONX"),
		},
		{
			Name:          "Strcmp95_DWAYNE_DUANE",
			Algorithm:     "Strcmp95",
			A:             "DWAYNE",
			B:             "DUANE",
			ExpectedScore: fuzzymatch.Strcmp95Score("DWAYNE", "DUANE"),
		},
		{
			Name:          "Strcmp95_HAMINGTON_HAMMINGTON",
			Algorithm:     "Strcmp95",
			A:             "HAMINGTON",
			B:             "HAMMINGTON",
			ExpectedScore: fuzzymatch.Strcmp95Score("HAMINGTON", "HAMMINGTON"),
		},
		{
			Name:          "Strcmp95_MARTHA_MARHTA",
			Algorithm:     "Strcmp95",
			A:             "MARTHA",
			B:             "MARHTA",
			ExpectedScore: fuzzymatch.Strcmp95Score("MARTHA", "MARHTA"),
		},
		{
			Name:          "Strcmp95_both_empty",
			Algorithm:     "Strcmp95",
			A:             "",
			B:             "",
			ExpectedScore: fuzzymatch.Strcmp95Score("", ""),
		},
		{
			Name:          "Strcmp95_identical",
			Algorithm:     "Strcmp95",
			A:             "abc",
			B:             "abc",
			ExpectedScore: fuzzymatch.Strcmp95Score("abc", "abc"),
		},
		{
			Name:          "Strcmp95_one_empty",
			Algorithm:     "Strcmp95",
			A:             "abc",
			B:             "",
			ExpectedScore: fuzzymatch.Strcmp95Score("abc", ""),
		},
	}
}

// TestGolden_Strcmp95_Staging produces testdata/golden/_staging/strcmp95.json
// for plan 04-05's merge step. Entries are sorted alphabetically by Name.
// Seven entries cover: both-empty, identical, one-empty, the three Winkler
// 1990 / Census Bureau canonical surnames (MARTHA/MARHTA, DWAYNE/DUANE,
// DIXON/DICKSONX), and the long-string adjustment trigger pair
// (HAMINGTON/HAMMINGTON).
//
// Plan 04-05 owns the canonical algorithms.json merge; this plan only writes
// the staging file. Do NOT update TestGolden_Algorithms_Merge's stagingFiles
// slice here — that's plan 04-05's responsibility.
//
// Run with `-update` to create or refresh the staging file.
// Re-running without `-update` must exit 0 (file is byte-stable).
func TestGolden_Strcmp95_Staging(t *testing.T) {
	entries := buildStrcmp95StagingEntries(t)
	sort.Slice(entries, func(i, j int) bool {
		return entries[i].Name < entries[j].Name
	})
	file := goldenAlgorithmsFile{Version: 1, Entries: entries}
	assertGoldenStaging(t, "_staging/strcmp95.json", file)
}

// buildLCSStrStagingEntries returns the LCSStr entries used by
// TestGolden_LCSStr_Staging. ExpectedScore is computed from the current
// implementation so the staging file stays in sync with actual output. Seven
// byte-path entries cover: both-empty (1.0), identical (1.0), one-empty (0.0),
// canonical Wagner-Fischer 1974 reference vectors (kitten/sitting,
// http_request/http_request_header_fields), no-overlap (0.0 — RESEARCH.md
// Pitfall 6 disambiguation), AND the leftmost-tie-break load-bearing pin
// (abcXYZabc/abc — RESEARCH.md Pitfall 4).
//
// All entries record LCSStrScore on byte-string inputs (the Phase 2 staging-
// golden convention — rune-path coverage stays in unit tests; the golden
// records the dispatched byte-path score for cross-platform determinism).
//
// Plan 04-05 owns the merge into testdata/golden/algorithms.json — this plan
// only writes the staging file.
func buildLCSStrStagingEntries(t *testing.T) []goldenAlgorithmEntry {
	t.Helper()
	return []goldenAlgorithmEntry{
		{
			Name:          "LCSStr_both_empty",
			Algorithm:     "LCSStr",
			A:             "",
			B:             "",
			ExpectedScore: fuzzymatch.LCSStrScore("", ""),
		},
		{
			Name:          "LCSStr_http_request_containment",
			Algorithm:     "LCSStr",
			A:             "http_request",
			B:             "http_request_header_fields",
			ExpectedScore: fuzzymatch.LCSStrScore("http_request", "http_request_header_fields"),
		},
		{
			Name:          "LCSStr_identical",
			Algorithm:     "LCSStr",
			A:             "abc",
			B:             "abc",
			ExpectedScore: fuzzymatch.LCSStrScore("abc", "abc"),
		},
		{
			Name:          "LCSStr_kitten_sitting",
			Algorithm:     "LCSStr",
			A:             "kitten",
			B:             "sitting",
			ExpectedScore: fuzzymatch.LCSStrScore("kitten", "sitting"),
		},
		{
			Name:          "LCSStr_leftmost_tie_break",
			Algorithm:     "LCSStr",
			A:             "abcXYZabc",
			B:             "abc",
			ExpectedScore: fuzzymatch.LCSStrScore("abcXYZabc", "abc"),
		},
		{
			Name:          "LCSStr_no_overlap",
			Algorithm:     "LCSStr",
			A:             "abc",
			B:             "xyz",
			ExpectedScore: fuzzymatch.LCSStrScore("abc", "xyz"),
		},
		{
			Name:          "LCSStr_one_empty",
			Algorithm:     "LCSStr",
			A:             "abc",
			B:             "",
			ExpectedScore: fuzzymatch.LCSStrScore("abc", ""),
		},
	}
}

// TestGolden_LCSStr_Staging produces testdata/golden/_staging/lcsstr.json for
// plan 04-05's merge step. Entries are sorted alphabetically by Name. Seven
// entries cover: both-empty, http_request substring containment, identical,
// kitten/sitting, leftmost-tie-break (the load-bearing strict-`>` pin),
// no-overlap (Pitfall 6 disambiguation), and one-empty.
//
// Plan 04-05 owns the canonical algorithms.json merge; this plan only writes
// the staging file. Do NOT update TestGolden_Algorithms_Merge's stagingFiles
// slice here — that's plan 04-05's responsibility.
//
// Run with `-update` to create or refresh the staging file.
// Re-running without `-update` must exit 0 (file is byte-stable).
func TestGolden_LCSStr_Staging(t *testing.T) {
	entries := buildLCSStrStagingEntries(t)
	sort.Slice(entries, func(i, j int) bool {
		return entries[i].Name < entries[j].Name
	})
	file := goldenAlgorithmsFile{Version: 1, Entries: entries}
	assertGoldenStaging(t, "_staging/lcsstr.json", file)
}

// buildRatcliffObershelpStagingEntries returns the Ratcliff-Obershelp entries
// used by TestGolden_RatcliffObershelp_Staging. ExpectedScore is computed
// from the current implementation so the staging file stays in sync with
// actual output. Seven entries cover: both-empty (1.0), identical (1.0),
// one-empty (0.0), substring-middle (abcdef inside xyzabcdefuvw),
// asymmetric tide/diet (ONE direction — the cross-algorithm asymmetry-pin
// test in plan 04-05 verifies fwd != rev), Dr. Dobb's 1988 canonical
// reference pair WIKIMEDIA/WIKIMANIA, and the GESTALT/GESTALT_PATTERN_MATCHING
// paper-cited pair.
//
// Plan 04-05 owns the merge into testdata/golden/algorithms.json — this
// plan only writes the staging file.
func buildRatcliffObershelpStagingEntries(t *testing.T) []goldenAlgorithmEntry {
	t.Helper()
	return []goldenAlgorithmEntry{
		{
			Name:          "RatcliffObershelp_both_empty",
			Algorithm:     "RatcliffObershelp",
			A:             "",
			B:             "",
			ExpectedScore: fuzzymatch.RatcliffObershelpScore("", ""),
		},
		{
			Name:          "RatcliffObershelp_gestalt_paper",
			Algorithm:     "RatcliffObershelp",
			A:             "GESTALT",
			B:             "GESTALT_PATTERN_MATCHING",
			ExpectedScore: fuzzymatch.RatcliffObershelpScore("GESTALT", "GESTALT_PATTERN_MATCHING"),
		},
		{
			Name:          "RatcliffObershelp_identical",
			Algorithm:     "RatcliffObershelp",
			A:             "abc",
			B:             "abc",
			ExpectedScore: fuzzymatch.RatcliffObershelpScore("abc", "abc"),
		},
		{
			Name:          "RatcliffObershelp_one_empty",
			Algorithm:     "RatcliffObershelp",
			A:             "abc",
			B:             "",
			ExpectedScore: fuzzymatch.RatcliffObershelpScore("abc", ""),
		},
		{
			Name:          "RatcliffObershelp_substring_middle",
			Algorithm:     "RatcliffObershelp",
			A:             "abcdef",
			B:             "xyzabcdefuvw",
			ExpectedScore: fuzzymatch.RatcliffObershelpScore("abcdef", "xyzabcdefuvw"),
		},
		{
			Name:          "RatcliffObershelp_tide_diet_asymmetric",
			Algorithm:     "RatcliffObershelp",
			A:             "tide",
			B:             "diet",
			ExpectedScore: fuzzymatch.RatcliffObershelpScore("tide", "diet"),
		},
		{
			Name:          "RatcliffObershelp_wikimedia_wikimania",
			Algorithm:     "RatcliffObershelp",
			A:             "WIKIMEDIA",
			B:             "WIKIMANIA",
			ExpectedScore: fuzzymatch.RatcliffObershelpScore("WIKIMEDIA", "WIKIMANIA"),
		},
	}
}

// TestGolden_RatcliffObershelp_Staging produces
// testdata/golden/_staging/ratcliff_obershelp.json for plan 04-05's merge
// step. Entries are sorted alphabetically by Name. Seven entries cover:
// both-empty, GESTALT (Dr. Dobb's paper-cited), identical, one-empty,
// substring middle, asymmetric tide/diet pair (ONE direction; asymmetry
// itself is verified by TestRatcliffObershelp_AsymmetryPin and by the
// cross-algorithm consistency test in plan 04-05), and the canonical
// WIKIMEDIA/WIKIMANIA pair.
//
// Plan 04-05 owns the canonical algorithms.json merge; this plan only
// writes the staging file. Do NOT update TestGolden_Algorithms_Merge's
// stagingFiles slice here — that's plan 04-05's responsibility.
//
// Run with `-update` to create or refresh the staging file.
// Re-running without `-update` must exit 0 (file is byte-stable).
func TestGolden_RatcliffObershelp_Staging(t *testing.T) {
	entries := buildRatcliffObershelpStagingEntries(t)
	sort.Slice(entries, func(i, j int) bool {
		return entries[i].Name < entries[j].Name
	})
	file := goldenAlgorithmsFile{Version: 1, Entries: entries}
	assertGoldenStaging(t, "_staging/ratcliff_obershelp.json", file)
}

// buildQGramJaccardStagingEntries returns the eight Q-Gram Jaccard
// entries used by TestGolden_QGramJaccard_Staging. ExpectedScore is
// computed from the current implementation so the staging file stays
// in sync with actual output. Eight entries (sorted by Name in the test):
//
//   - QGramJaccard_AGCT_AGCTAGCT      (RV-J1; Ukkonen 1992 §3 worked
//                                      example; n=2; 3/7 ≈ 0.4286)
//   - QGramJaccard_abcd_abxy          (RV-J4; single-shared bigram;
//                                      n=2; 1/5 = 0.2)
//   - QGramJaccard_both_empty         (n=2; both-empty convention; 1.0)
//   - QGramJaccard_cafe_runes         (RV-J5-Runes; rune path;
//                                      n=2; 2/4 = 0.5)
//   - QGramJaccard_identical          (RV-J2; identity short-circuit;
//                                      n=2; 1.0)
//   - QGramJaccard_n_too_large        (RV-J6; n > min length; both-empty
//                                      convention; 1.0)
//   - QGramJaccard_no_overlap         (RV-J3; n=2; 0/4 = 0.0)
//   - QGramJaccard_one_empty          (n=2; one-empty convention; 0.0)
//
// Plan 05-05 owns the merge into testdata/golden/algorithms.json — this
// plan only writes the staging file. The cafe_runes entry is the
// rune-path canary; all others are byte-path.
func buildQGramJaccardStagingEntries(t *testing.T) []goldenAlgorithmEntry {
	t.Helper()
	return []goldenAlgorithmEntry{
		{
			Name:          "QGramJaccard_AGCT_AGCTAGCT",
			Algorithm:     "QGramJaccard",
			A:             "AGCT",
			B:             "AGCTAGCT",
			ExpectedScore: fuzzymatch.QGramJaccardScore("AGCT", "AGCTAGCT", 2),
		},
		{
			Name:          "QGramJaccard_abcd_abxy",
			Algorithm:     "QGramJaccard",
			A:             "abcd",
			B:             "abxy",
			ExpectedScore: fuzzymatch.QGramJaccardScore("abcd", "abxy", 2),
		},
		{
			Name:          "QGramJaccard_both_empty",
			Algorithm:     "QGramJaccard",
			A:             "",
			B:             "",
			ExpectedScore: fuzzymatch.QGramJaccardScore("", "", 2),
		},
		{
			Name:          "QGramJaccard_cafe_runes",
			Algorithm:     "QGramJaccard",
			A:             "café",
			B:             "cafe",
			ExpectedScore: fuzzymatch.QGramJaccardScoreRunes("café", "cafe", 2),
		},
		{
			Name:          "QGramJaccard_identical",
			Algorithm:     "QGramJaccard",
			A:             "hello",
			B:             "hello",
			ExpectedScore: fuzzymatch.QGramJaccardScore("hello", "hello", 2),
		},
		{
			Name:          "QGramJaccard_n_too_large",
			Algorithm:     "QGramJaccard",
			A:             "ab",
			B:             "abc",
			ExpectedScore: fuzzymatch.QGramJaccardScore("ab", "abc", 5),
		},
		{
			Name:          "QGramJaccard_no_overlap",
			Algorithm:     "QGramJaccard",
			A:             "abc",
			B:             "xyz",
			ExpectedScore: fuzzymatch.QGramJaccardScore("abc", "xyz", 2),
		},
		{
			Name:          "QGramJaccard_one_empty",
			Algorithm:     "QGramJaccard",
			A:             "",
			B:             "abc",
			ExpectedScore: fuzzymatch.QGramJaccardScore("", "abc", 2),
		},
	}
}

// TestGolden_QGramJaccard_Staging produces
// testdata/golden/_staging/qgram_jaccard.json for plan 05-05's merge
// step. Entries are sorted alphabetically by Name. Eight entries cover:
// AGCT_AGCTAGCT (RV-J1 Ukkonen 1992 §3 worked example), abcd_abxy
// (RV-J4 single-shared bigram), both-empty, cafe_runes (RV-J5-Runes
// rune-path canary), identical, n_too_large, no_overlap (RV-J3),
// one_empty.
//
// Plan 05-05 owns the canonical algorithms.json merge; this plan only
// writes the staging file. Do NOT update TestGolden_Algorithms_Merge's
// stagingFiles slice here — that's plan 05-05's responsibility.
//
// Run with `-update` to create or refresh the staging file.
// Re-running without `-update` must exit 0 (file is byte-stable).
func TestGolden_QGramJaccard_Staging(t *testing.T) {
	entries := buildQGramJaccardStagingEntries(t)
	sort.Slice(entries, func(i, j int) bool {
		return entries[i].Name < entries[j].Name
	})
	file := goldenAlgorithmsFile{Version: 1, Entries: entries}
	assertGoldenStaging(t, "_staging/qgram_jaccard.json", file)
}

// buildSorensenDiceStagingEntries returns the eight Sørensen-Dice
// entries used by TestGolden_SorensenDice_Staging. ExpectedScore is
// computed from the current implementation so the staging file stays
// in sync with actual output. Eight entries (sorted by Name in the test):
//
//   - SorensenDice_abcdef_abcXef_n3   (RV-D3; trigram variant; n=3;
//                                      2·1/(4+4) = 0.25)
//   - SorensenDice_abcdef_bcdefg      (RV-D2; high-overlap analogue;
//                                      n=2; 2·4/(5+5) = 0.8)
//   - SorensenDice_both_empty         (n=2; both-empty convention; 1.0)
//   - SorensenDice_cafe_runes         (rune-path canary; n=2;
//                                      2·2/(3+3) = 4/6 ≈ 0.6667)
//   - SorensenDice_identical          (RV-D4; identity short-circuit;
//                                      n=2; 1.0)
//   - SorensenDice_night_nacht        (RV-D1; load-bearing canonical
//                                      NLP-textbook bigram pair; n=2;
//                                      2·1/(4+4) = 0.25)
//   - SorensenDice_no_overlap         (n=2; 2·0/(2+2) = 0.0)
//   - SorensenDice_one_empty          (n=2; one-empty convention; 0.0)
//
// Plan 05-05 owns the merge into testdata/golden/algorithms.json — this
// plan only writes the staging file. The cafe_runes entry is the
// rune-path canary; all others are byte-path.
func buildSorensenDiceStagingEntries(t *testing.T) []goldenAlgorithmEntry {
	t.Helper()
	return []goldenAlgorithmEntry{
		{
			Name:          "SorensenDice_abcdef_abcXef_n3",
			Algorithm:     "SorensenDice",
			A:             "abcdef",
			B:             "abcXef",
			ExpectedScore: fuzzymatch.SorensenDiceScore("abcdef", "abcXef", 3),
		},
		{
			Name:          "SorensenDice_abcdef_bcdefg",
			Algorithm:     "SorensenDice",
			A:             "abcdef",
			B:             "bcdefg",
			ExpectedScore: fuzzymatch.SorensenDiceScore("abcdef", "bcdefg", 2),
		},
		{
			Name:          "SorensenDice_both_empty",
			Algorithm:     "SorensenDice",
			A:             "",
			B:             "",
			ExpectedScore: fuzzymatch.SorensenDiceScore("", "", 2),
		},
		{
			Name:          "SorensenDice_cafe_runes",
			Algorithm:     "SorensenDice",
			A:             "café",
			B:             "cafe",
			ExpectedScore: fuzzymatch.SorensenDiceScoreRunes("café", "cafe", 2),
		},
		{
			Name:          "SorensenDice_identical",
			Algorithm:     "SorensenDice",
			A:             "hello",
			B:             "hello",
			ExpectedScore: fuzzymatch.SorensenDiceScore("hello", "hello", 2),
		},
		{
			Name:          "SorensenDice_night_nacht",
			Algorithm:     "SorensenDice",
			A:             "night",
			B:             "nacht",
			ExpectedScore: fuzzymatch.SorensenDiceScore("night", "nacht", 2),
		},
		{
			Name:          "SorensenDice_no_overlap",
			Algorithm:     "SorensenDice",
			A:             "abc",
			B:             "xyz",
			ExpectedScore: fuzzymatch.SorensenDiceScore("abc", "xyz", 2),
		},
		{
			Name:          "SorensenDice_one_empty",
			Algorithm:     "SorensenDice",
			A:             "",
			B:             "abc",
			ExpectedScore: fuzzymatch.SorensenDiceScore("", "abc", 2),
		},
	}
}

// TestGolden_SorensenDice_Staging produces
// testdata/golden/_staging/sorensen_dice.json for plan 05-05's merge
// step. Entries are sorted alphabetically by Name. Eight entries cover:
// abcdef_abcXef_n3 (RV-D3 trigram variant), abcdef_bcdefg (RV-D2
// high-overlap analogue), both_empty, cafe_runes (rune-path canary),
// identical (RV-D4), night_nacht (RV-D1 load-bearing canonical
// NLP-textbook bigram pair), no_overlap, one_empty.
//
// Plan 05-05 owns the canonical algorithms.json merge; this plan only
// writes the staging file. Do NOT update TestGolden_Algorithms_Merge's
// stagingFiles slice here — that's plan 05-05's responsibility.
//
// Run with `-update` to create or refresh the staging file.
// Re-running without `-update` must exit 0 (file is byte-stable).
func TestGolden_SorensenDice_Staging(t *testing.T) {
	entries := buildSorensenDiceStagingEntries(t)
	sort.Slice(entries, func(i, j int) bool {
		return entries[i].Name < entries[j].Name
	})
	file := goldenAlgorithmsFile{Version: 1, Entries: entries}
	assertGoldenStaging(t, "_staging/sorensen_dice.json", file)
}
