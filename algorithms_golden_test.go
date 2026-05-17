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
		"_staging/cosine.json",
		"_staging/damerau_full.json",
		"_staging/damerau_osa.json",
		"_staging/double_metaphone.json",
		"_staging/hamming.json",
		"_staging/jaro.json",
		"_staging/jarowinkler.json",
		"_staging/lcsstr.json",
		"_staging/levenshtein.json",
		"_staging/monge_elkan.json",
		"_staging/mra.json",
		"_staging/nysiis.json",
		"_staging/partial_ratio.json",
		"_staging/qgram_jaccard.json",
		"_staging/ratcliff_obershelp.json",
		"_staging/sorensen_dice.json",
		"_staging/soundex.json",
		"_staging/strcmp95.json",
		"_staging/swg.json",
		"_staging/token_jaccard.json",
		"_staging/token_set_ratio.json",
		"_staging/token_sort_ratio.json",
		"_staging/tversky.json",
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
//     example; n=2; 3/7 ≈ 0.4286)
//   - QGramJaccard_abcd_abxy          (RV-J4; single-shared bigram;
//     n=2; 1/5 = 0.2)
//   - QGramJaccard_both_empty         (n=2; both-empty convention; 1.0)
//   - QGramJaccard_cafe_runes         (RV-J5-Runes; rune path;
//     n=2; 2/4 = 0.5)
//   - QGramJaccard_identical          (RV-J2; identity short-circuit;
//     n=2; 1.0)
//   - QGramJaccard_n_too_large        (RV-J6; n > min length; both-empty
//     convention; 1.0)
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
//     2·1/(4+4) = 0.25)
//   - SorensenDice_abcdef_bcdefg      (RV-D2; high-overlap analogue;
//     n=2; 2·4/(5+5) = 0.8)
//   - SorensenDice_both_empty         (n=2; both-empty convention; 1.0)
//   - SorensenDice_cafe_runes         (rune-path canary; n=2;
//     2·2/(3+3) = 4/6 ≈ 0.6667)
//   - SorensenDice_identical          (RV-D4; identity short-circuit;
//     n=2; 1.0)
//   - SorensenDice_night_nacht        (RV-D1; load-bearing canonical
//     NLP-textbook bigram pair; n=2;
//     2·1/(4+4) = 0.25)
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

// buildCosineStagingEntries returns the nine Cosine entries used by
// TestGolden_Cosine_Staging. ExpectedScore is computed from the current
// implementation so the staging file stays in sync with actual output.
// Nine entries (sorted by Name in the test) per CONTEXT.md §1 LOCKED
// (algorithms.json entries spanning ASCII + Unicode at n ∈ {2, 3, 4})
// and RESEARCH.md §2.3 slate:
//
//   - Cosine_ascii_n2_irrational    ("abc"/"abcd"/n=2/byte; RV-C1;
//     2/sqrt(6) ≈ 0.8164965809277261)
//   - Cosine_ascii_n3_large_intersection ("abcdefgh"/"abcdefgi"/n=3/byte;
//     RV-C2; 5/6 = 0.8333333333333334)
//   - Cosine_ascii_n4_exact         ("abcde"/"abcdf"/n=4/byte; RV-C4;
//     1/(sqrt(2)·sqrt(2)) =
//     0.49999999999999989 — see
//     cosine_test.go RV-C4 derivation
//     block; RESEARCH.md "Pitfall 2")
//   - Cosine_both_empty             (""/""/n=2/byte; both-empty
//     convention; 1.0)
//   - Cosine_identical              ("hello"/"hello"/n=2/byte; identity
//     short-circuit; 1.0)
//   - Cosine_one_empty              (""/"abc"/n=2/byte; one-empty
//     convention; 0.0)
//   - Cosine_orthogonal             ("abc"/"xyz"/n=2/byte; empty
//     intersection → cos=0)
//   - Cosine_unicode_n2_runes       ("café"/"cafe"/n=2/rune; RV-C3;
//     2/3 = 0.6666666666666666)
//   - Cosine_unicode_n3_runes       ("héllo"/"hello"/n=3/rune;
//     1/3 = 0.3333333333333333 —
//     rune-trigrams: ["hél","éll","llo"]
//     vs ["hel","ell","llo"];
//     intersection sorted = ["llo"];
//     dot=1; ‖A‖²=‖B‖²=3;
//     cos = 1/(sqrt(3)·sqrt(3)) = 1/3)
//
// Plan 05-05 owns the merge into testdata/golden/algorithms.json — this
// plan only writes the staging file. The unicode entries are rune-path;
// all others are byte-path.
//
// LOAD-BEARING per CONTEXT.md §1: this slate is the cross-platform
// float-determinism gate. Any single-byte drift in algorithms.json
// (after plan 05-05 merge) on linux/amd64 vs linux/arm64 vs darwin/arm64
// vs windows/amd64 fails `make verify-determinism` HARD.
func buildCosineStagingEntries(t *testing.T) []goldenAlgorithmEntry {
	t.Helper()
	return []goldenAlgorithmEntry{
		{
			Name:          "Cosine_ascii_n2_irrational",
			Algorithm:     "Cosine",
			A:             "abc",
			B:             "abcd",
			ExpectedScore: fuzzymatch.CosineScore("abc", "abcd", 2),
		},
		{
			Name:          "Cosine_ascii_n3_large_intersection",
			Algorithm:     "Cosine",
			A:             "abcdefgh",
			B:             "abcdefgi",
			ExpectedScore: fuzzymatch.CosineScore("abcdefgh", "abcdefgi", 3),
		},
		{
			Name:          "Cosine_ascii_n4_exact",
			Algorithm:     "Cosine",
			A:             "abcde",
			B:             "abcdf",
			ExpectedScore: fuzzymatch.CosineScore("abcde", "abcdf", 4),
		},
		{
			Name:          "Cosine_both_empty",
			Algorithm:     "Cosine",
			A:             "",
			B:             "",
			ExpectedScore: fuzzymatch.CosineScore("", "", 2),
		},
		{
			Name:          "Cosine_identical",
			Algorithm:     "Cosine",
			A:             "hello",
			B:             "hello",
			ExpectedScore: fuzzymatch.CosineScore("hello", "hello", 2),
		},
		{
			Name:          "Cosine_one_empty",
			Algorithm:     "Cosine",
			A:             "",
			B:             "abc",
			ExpectedScore: fuzzymatch.CosineScore("", "abc", 2),
		},
		{
			Name:          "Cosine_orthogonal",
			Algorithm:     "Cosine",
			A:             "abc",
			B:             "xyz",
			ExpectedScore: fuzzymatch.CosineScore("abc", "xyz", 2),
		},
		{
			Name:          "Cosine_unicode_n2_runes",
			Algorithm:     "Cosine",
			A:             "café",
			B:             "cafe",
			ExpectedScore: fuzzymatch.CosineScoreRunes("café", "cafe", 2),
		},
		{
			Name:          "Cosine_unicode_n3_runes",
			Algorithm:     "Cosine",
			A:             "héllo",
			B:             "hello",
			ExpectedScore: fuzzymatch.CosineScoreRunes("héllo", "hello", 3),
		},
	}
}

// TestGolden_Cosine_Staging produces
// testdata/golden/_staging/cosine.json for plan 05-05's merge step.
// Entries are sorted alphabetically by Name. Nine entries cover
// CONTEXT.md §1 LOCKED ASCII + Unicode at n ∈ {2, 3, 4}: ascii_n2_
// irrational (RV-C1), ascii_n3_large_intersection (RV-C2),
// ascii_n4_exact (RV-C4 — IEEE-754 1-ULP shortfall from 0.5; see
// cosine_test.go derivation), both_empty, identical, one_empty,
// orthogonal, unicode_n2_runes (RV-C3), unicode_n3_runes (1/3 from
// "héllo"/"hello").
//
// LOAD-BEARING per CONTEXT.md §1 — this is the cross-platform
// float-determinism gate. Plan 05-05 owns the merge into
// testdata/golden/algorithms.json; this plan only writes the staging
// file. Do NOT update TestGolden_Algorithms_Merge's stagingFiles slice
// here — that's plan 05-05's responsibility.
//
// Run with `-update` to create or refresh the staging file.
// Re-running without `-update` must exit 0 (file is byte-stable).
func TestGolden_Cosine_Staging(t *testing.T) {
	entries := buildCosineStagingEntries(t)
	sort.Slice(entries, func(i, j int) bool {
		return entries[i].Name < entries[j].Name
	})
	file := goldenAlgorithmsFile{Version: 1, Entries: entries}
	assertGoldenStaging(t, "_staging/cosine.json", file)
}

// buildTverskyStagingEntries returns the eight Tversky entries used by
// TestGolden_Tversky_Staging. ExpectedScore is computed from the current
// implementation so the staging file stays in sync with actual output.
// Eight entries (sorted by Name in the test):
//
//   - Tversky_abcd_abce_dice_eq        (RV-T4; α=β=0.5 → Sørensen-Dice;
//     n=2; 2/3 ≈ 0.6667)
//   - Tversky_abcd_abce_jaccard_eq     (RV-T3; α=β=1.0 → Jaccard;
//     n=2; 2/4 = 0.5)
//   - Tversky_abcd_abcdef_asym         (RV-T1; LOAD-BEARING asymmetry
//     gate first half; α=0.8, β=0.2;
//     n=2; 0.8823529411764706)
//   - Tversky_abcdef_abcd_asym_swap    (RV-T2; LOAD-BEARING asymmetry
//     gate second half — input swap
//     of RV-T1 with same (α, β);
//     n=2; 0.6521739130434783)
//   - Tversky_both_empty               (n=2; α=β=0.5; both-empty
//     convention; 1.0)
//   - Tversky_cafe_runes               (rune-path canary; n=2;
//     α=β=0.5; 2/3 ≈ 0.6667)
//   - Tversky_identical                ("hello"/"hello"/n=2;
//     α=0.8/β=0.2; identity
//     short-circuit; 1.0)
//   - Tversky_one_empty                (n=2; α=β=0.5; one-empty
//     convention; 0.0)
//
// The two asymmetry-pair rows (Tversky_abcd_abcdef_asym at the head
// of the alphabetical sort and Tversky_abcdef_abcd_asym_swap two rows
// later) form the LOAD-BEARING asymmetry-direction-sensitivity fixture
// at the golden-file level: same (α, β), inputs swapped, scores
// differ. The adjacency-by-prefix is intentional — both names start
// with "Tversky_abcd" so they sort together; the cross-check
// degeneracy rows (dice_eq, jaccard_eq) are interleaved by the
// alphabetical sort but are unrelated to the asymmetry gate. This
// fixture is the third-layer regression detector, alongside the
// unit-test (TestTversky_AsymmetryDirectionSensitive) and BDD
// (tversky.feature asymmetry scenario) gates.
//
// Plan 05-05 owns the merge into testdata/golden/algorithms.json — this
// plan only writes the staging file. The cafe_runes entry is the
// rune-path canary; all others are byte-path.
func buildTverskyStagingEntries(t *testing.T) []goldenAlgorithmEntry {
	t.Helper()
	return []goldenAlgorithmEntry{
		{
			Name:          "Tversky_abcd_abce_dice_eq",
			Algorithm:     "Tversky",
			A:             "abcd",
			B:             "abce",
			ExpectedScore: fuzzymatch.TverskyScore("abcd", "abce", 2, 0.5, 0.5),
		},
		{
			Name:          "Tversky_abcd_abce_jaccard_eq",
			Algorithm:     "Tversky",
			A:             "abcd",
			B:             "abce",
			ExpectedScore: fuzzymatch.TverskyScore("abcd", "abce", 2, 1.0, 1.0),
		},
		{
			Name:          "Tversky_abcd_abcdef_asym",
			Algorithm:     "Tversky",
			A:             "abcd",
			B:             "abcdef",
			ExpectedScore: fuzzymatch.TverskyScore("abcd", "abcdef", 2, 0.8, 0.2),
		},
		{
			Name:          "Tversky_abcdef_abcd_asym_swap",
			Algorithm:     "Tversky",
			A:             "abcdef",
			B:             "abcd",
			ExpectedScore: fuzzymatch.TverskyScore("abcdef", "abcd", 2, 0.8, 0.2),
		},
		{
			Name:          "Tversky_both_empty",
			Algorithm:     "Tversky",
			A:             "",
			B:             "",
			ExpectedScore: fuzzymatch.TverskyScore("", "", 2, 0.5, 0.5),
		},
		{
			Name:          "Tversky_cafe_runes",
			Algorithm:     "Tversky",
			A:             "café",
			B:             "cafe",
			ExpectedScore: fuzzymatch.TverskyScoreRunes("café", "cafe", 2, 0.5, 0.5),
		},
		{
			Name:          "Tversky_identical",
			Algorithm:     "Tversky",
			A:             "hello",
			B:             "hello",
			ExpectedScore: fuzzymatch.TverskyScore("hello", "hello", 2, 0.8, 0.2),
		},
		{
			Name:          "Tversky_one_empty",
			Algorithm:     "Tversky",
			A:             "",
			B:             "abc",
			ExpectedScore: fuzzymatch.TverskyScore("", "abc", 2, 0.5, 0.5),
		},
	}
}

// TestGolden_Tversky_Staging produces
// testdata/golden/_staging/tversky.json for plan 05-05's merge step.
// Entries are sorted alphabetically by Name. Eight entries cover:
// abcd_abce_dice_eq (RV-T4 Dice-equivalent), abcd_abce_jaccard_eq
// (RV-T3 Jaccard-equivalent), abcd_abcdef_asym (RV-T1 — first half of
// LOAD-BEARING asymmetry gate), abcdef_abcd_asym_swap (RV-T2 — input
// swap of RV-T1; the two asymmetry rows are adjacent in the sort for
// reviewer clarity), both_empty, cafe_runes (rune-path canary),
// identical, one_empty.
//
// LOAD-BEARING per CONTEXT.md §5: the asymmetry-pair rows
// (Tversky_abcd_abcdef_asym and Tversky_abcdef_abcd_asym_swap) form
// the asymmetry-direction-sensitivity gate at the golden-file level.
// A silent α/β swap inside TverskyScore would cause both rows to
// flip values (RV-T1 → RV-T2's score and vice versa) and the staging
// file regeneration would surface the regression as a byte-diff.
//
// Plan 05-05 owns the canonical algorithms.json merge; this plan only
// writes the staging file. Do NOT update TestGolden_Algorithms_Merge's
// stagingFiles slice here — that's plan 05-05's responsibility.
//
// Run with `-update` to create or refresh the staging file.
// Re-running without `-update` must exit 0 (file is byte-stable).
func TestGolden_Tversky_Staging(t *testing.T) {
	entries := buildTverskyStagingEntries(t)
	sort.Slice(entries, func(i, j int) bool {
		return entries[i].Name < entries[j].Name
	})
	file := goldenAlgorithmsFile{Version: 1, Entries: entries}
	assertGoldenStaging(t, "_staging/tversky.json", file)
}

// -----------------------------------------------------------------------------
// Phase 08.5 Plan 09 (Q13) — staging-test fill-in for Phase 6 and Phase 7
// algorithms that previously had committed staging JSON files but no
// producing Go-side test. The nine functions below close the catalogue at
// 23 TestGolden_*_Staging entries (matching the 23-algorithm catalogue in
// CLAUDE.md and algoid.go).
//
// Each function follows the mechanical template established by the 14
// pre-existing TestGolden_*_Staging functions above:
//
//   - a build<Algo>StagingEntries(t) constructor returning a slice of
//     goldenAlgorithmEntry whose ExpectedScore is computed from the current
//     implementation (so the staging file stays in sync with actual output);
//   - the test sorts by Name and calls assertGoldenStaging on the
//     per-algorithm staging path.
//
// Test inputs are drawn from each algorithm's existing reference-vector
// unit tests (the canonical worked examples cited from the primary source)
// PLUS the standard catalogue edge cases (both-empty, one-empty, identity).
// The set is hand-picked (no random sampling, no time, no map iteration)
// so the staging files are deterministic across runs and platforms per
// determinism-standards §13.4.
//
// Code-returning algorithms (DoubleMetaphone, NYSIIS, Soundex, MRA) are
// scored via their *Score wrappers (which compare codes and return the
// canonical 0.0 / 1.0 binary), so the standard float64 ExpectedScore
// shape applies — no framework extension is required (verified against
// dispatch_*.go for each of the four).

// buildDoubleMetaphoneStagingEntries returns ten DoubleMetaphone entries
// drawn from double_metaphone_test.go's worked-example slate (Philips 2000)
// plus the catalogue edge cases. DoubleMetaphoneScore wraps the two-code
// comparison into the 0.0 / 1.0 binary so the entries fit the standard
// goldenAlgorithmEntry shape (per the dispatch wrapper in
// dispatch_double_metaphone.go).
func buildDoubleMetaphoneStagingEntries(t *testing.T) []goldenAlgorithmEntry {
	t.Helper()
	return []goldenAlgorithmEntry{
		{
			Name:          "DoubleMetaphone_Catherine_Katherine_exact",
			Algorithm:     "DoubleMetaphone",
			A:             "Catherine",
			B:             "Katherine",
			ExpectedScore: fuzzymatch.DoubleMetaphoneScore("Catherine", "Katherine"),
		},
		{
			Name:          "DoubleMetaphone_Cheung_Chen_match",
			Algorithm:     "DoubleMetaphone",
			A:             "Cheung",
			B:             "Chen",
			ExpectedScore: fuzzymatch.DoubleMetaphoneScore("Cheung", "Chen"),
		},
		{
			Name:          "DoubleMetaphone_Cheung_XNK_match",
			Algorithm:     "DoubleMetaphone",
			A:             "Cheung",
			B:             "Cheung",
			ExpectedScore: fuzzymatch.DoubleMetaphoneScore("Cheung", "Cheung"),
		},
		{
			Name:          "DoubleMetaphone_Schmidt_Mueller_nomatch",
			Algorithm:     "DoubleMetaphone",
			A:             "Schmidt",
			B:             "Mueller",
			ExpectedScore: fuzzymatch.DoubleMetaphoneScore("Schmidt", "Mueller"),
		},
		{
			Name:          "DoubleMetaphone_Schmidt_Smith_XMT_match",
			Algorithm:     "DoubleMetaphone",
			A:             "Schmidt",
			B:             "Smith",
			ExpectedScore: fuzzymatch.DoubleMetaphoneScore("Schmidt", "Smith"),
		},
		{
			Name:          "DoubleMetaphone_Smith_Garcia_nomatch",
			Algorithm:     "DoubleMetaphone",
			A:             "Smith",
			B:             "Garcia",
			ExpectedScore: fuzzymatch.DoubleMetaphoneScore("Smith", "Garcia"),
		},
		{
			Name:          "DoubleMetaphone_both_empty",
			Algorithm:     "DoubleMetaphone",
			A:             "",
			B:             "",
			ExpectedScore: fuzzymatch.DoubleMetaphoneScore("", ""),
		},
		{
			Name:          "DoubleMetaphone_identical_Schmidt",
			Algorithm:     "DoubleMetaphone",
			A:             "Schmidt",
			B:             "Schmidt",
			ExpectedScore: fuzzymatch.DoubleMetaphoneScore("Schmidt", "Schmidt"),
		},
		{
			Name:          "DoubleMetaphone_one_empty_a",
			Algorithm:     "DoubleMetaphone",
			A:             "",
			B:             "Schmidt",
			ExpectedScore: fuzzymatch.DoubleMetaphoneScore("", "Schmidt"),
		},
		{
			Name:          "DoubleMetaphone_one_empty_b",
			Algorithm:     "DoubleMetaphone",
			A:             "Schmidt",
			B:             "",
			ExpectedScore: fuzzymatch.DoubleMetaphoneScore("Schmidt", ""),
		},
	}
}

// TestGolden_DoubleMetaphone_Staging produces
// testdata/golden/_staging/double_metaphone.json. Entries are sorted
// alphabetically by Name. Ten entries cover: Catherine/Katherine,
// Cheung/Chen, Cheung/Cheung, Schmidt/Mueller, Schmidt/Smith,
// Smith/Garcia (canonical Philips 2000 worked-example pairs), the
// catalogue edge cases (both-empty, identity, two one-empty variants).
//
// Run with `-update` to create or refresh the staging file.
// Re-running without `-update` must exit 0 (file is byte-stable).
func TestGolden_DoubleMetaphone_Staging(t *testing.T) {
	entries := buildDoubleMetaphoneStagingEntries(t)
	sort.Slice(entries, func(i, j int) bool {
		return entries[i].Name < entries[j].Name
	})
	file := goldenAlgorithmsFile{Version: 1, Entries: entries}
	assertGoldenStaging(t, "_staging/double_metaphone.json", file)
}

// buildMongeElkanStagingEntries returns ten Monge-Elkan entries scored via
// MongeElkanScore — the symmetric default bound to dispatch[AlgoMongeElkan]
// per CONTEXT.md §4 LOCKED with AlgoJaroWinkler as the default inner.
// Phase 8.5 Q3 rename: the unsuffixed MongeElkanScore is the symmetric
// default; the previously-inert NormalisationOptions parameter was removed.
// The symmetric default participates in the standard
// PropAlgorithmScore_Symmetric property test set, so the staging entries
// match the dispatched output (not the directional surface).
//
// Reference vectors are drawn from monge_elkan_test.go (RV-ME1 through
// RV-ME6) plus the standard catalogue edge cases.
func buildMongeElkanStagingEntries(t *testing.T) []goldenAlgorithmEntry {
	t.Helper()
	return []goldenAlgorithmEntry{
		{
			Name:          "MongeElkan_RV-ME1_user_create_vs_usr_creating_symmetric",
			Algorithm:     "MongeElkan",
			A:             "user create",
			B:             "usr creating",
			ExpectedScore: fuzzymatch.MongeElkanScore("user create", "usr creating", fuzzymatch.AlgoJaroWinkler),
		},
		{
			Name:          "MongeElkan_RV-ME2_identity",
			Algorithm:     "MongeElkan",
			A:             "alpha beta",
			B:             "alpha beta",
			ExpectedScore: fuzzymatch.MongeElkanScore("alpha beta", "alpha beta", fuzzymatch.AlgoJaroWinkler),
		},
		{
			Name:          "MongeElkan_RV-ME3_disjoint_greek_symmetric",
			Algorithm:     "MongeElkan",
			A:             "alpha beta",
			B:             "gamma delta",
			ExpectedScore: fuzzymatch.MongeElkanScore("alpha beta", "gamma delta", fuzzymatch.AlgoJaroWinkler),
		},
		{
			Name:          "MongeElkan_RV-ME4_RV-ME6_asymmetric_symmetric_average",
			Algorithm:     "MongeElkan",
			A:             "alpha",
			B:             "alpha beta gamma",
			ExpectedScore: fuzzymatch.MongeElkanScore("alpha", "alpha beta gamma", fuzzymatch.AlgoJaroWinkler),
		},
		{
			Name:          "MongeElkan_RV-ME6_RV-ME4_swapped_symmetric_same_average",
			Algorithm:     "MongeElkan",
			A:             "alpha beta gamma",
			B:             "alpha",
			ExpectedScore: fuzzymatch.MongeElkanScore("alpha beta gamma", "alpha", fuzzymatch.AlgoJaroWinkler),
		},
		{
			Name:          "MongeElkan_both_empty_standard",
			Algorithm:     "MongeElkan",
			A:             "",
			B:             "",
			ExpectedScore: fuzzymatch.MongeElkanScore("", "", fuzzymatch.AlgoJaroWinkler),
		},
		{
			Name:          "MongeElkan_one_empty_a",
			Algorithm:     "MongeElkan",
			A:             "",
			B:             "hello world",
			ExpectedScore: fuzzymatch.MongeElkanScore("", "hello world", fuzzymatch.AlgoJaroWinkler),
		},
		{
			Name:          "MongeElkan_one_empty_b",
			Algorithm:     "MongeElkan",
			A:             "hello world",
			B:             "",
			ExpectedScore: fuzzymatch.MongeElkanScore("hello world", "", fuzzymatch.AlgoJaroWinkler),
		},
		{
			Name:          "MongeElkan_pure_separators_both_empty_tokens",
			Algorithm:     "MongeElkan",
			A:             "___",
			B:             "...",
			ExpectedScore: fuzzymatch.MongeElkanScore("___", "...", fuzzymatch.AlgoJaroWinkler),
		},
		{
			Name:          "MongeElkan_token_reorder_symmetric",
			Algorithm:     "MongeElkan",
			A:             "alpha beta gamma",
			B:             "gamma alpha beta",
			ExpectedScore: fuzzymatch.MongeElkanScore("alpha beta gamma", "gamma alpha beta", fuzzymatch.AlgoJaroWinkler),
		},
	}
}

// TestGolden_MongeElkan_Staging produces
// testdata/golden/_staging/monge_elkan.json. Entries are sorted
// alphabetically by Name. Ten entries cover RV-ME1 through RV-ME6
// (Monge & Elkan 1996 §3 reference vectors — including the symmetric
// average of the asymmetry-pair RV-ME4 / RV-ME6) and the standard
// catalogue edge cases (both-empty, identity via RV-ME2, two one-empty
// variants, pure-separator inputs, token reorder).
//
// All entries score via MongeElkanScore (the symmetric default —
// post Phase 8.5 Q3 rename) with the LOCKED default inner
// AlgoJaroWinkler per dispatch_monge_elkan.go.
//
// Run with `-update` to create or refresh the staging file.
// Re-running without `-update` must exit 0 (file is byte-stable).
func TestGolden_MongeElkan_Staging(t *testing.T) {
	entries := buildMongeElkanStagingEntries(t)
	sort.Slice(entries, func(i, j int) bool {
		return entries[i].Name < entries[j].Name
	})
	file := goldenAlgorithmsFile{Version: 1, Entries: entries}
	assertGoldenStaging(t, "_staging/monge_elkan.json", file)
}

// buildMRAStagingEntries returns ten MRA (Match Rating Approach) entries
// drawn from mra_test.go's reference vectors (NBS Tech Note 943, Lynch
// & Arends 1977) plus the catalogue edge cases. MRAScore returns the
// canonical 0.0 / 1.0 binary outcome from MRACompare's threshold rule;
// the goldenAlgorithmEntry float64 shape applies directly.
func buildMRAStagingEntries(t *testing.T) []goldenAlgorithmEntry {
	t.Helper()
	return []goldenAlgorithmEntry{
		{
			Name:          "MRA_Ad_ZachariahMontgomery_nomatch",
			Algorithm:     "MRA",
			A:             "Ad",
			B:             "ZachariahMontgomery",
			ExpectedScore: fuzzymatch.MRAScore("Ad", "ZachariahMontgomery"),
		},
		{
			Name:          "MRA_Byrne_Boern_match",
			Algorithm:     "MRA",
			A:             "Byrne",
			B:             "Boern",
			ExpectedScore: fuzzymatch.MRAScore("Byrne", "Boern"),
		},
		{
			Name:          "MRA_Catherine_Katherine_match",
			Algorithm:     "MRA",
			A:             "Catherine",
			B:             "Katherine",
			ExpectedScore: fuzzymatch.MRAScore("Catherine", "Katherine"),
		},
		{
			Name:          "MRA_Smith_Jones_nomatch",
			Algorithm:     "MRA",
			A:             "Smith",
			B:             "Jones",
			ExpectedScore: fuzzymatch.MRAScore("Smith", "Jones"),
		},
		{
			Name:          "MRA_Smith_Smyth_match",
			Algorithm:     "MRA",
			A:             "Smith",
			B:             "Smyth",
			ExpectedScore: fuzzymatch.MRAScore("Smith", "Smyth"),
		},
		{
			Name:          "MRA_Smith_both",
			Algorithm:     "MRA",
			A:             "Smith",
			B:             "Smith",
			ExpectedScore: fuzzymatch.MRAScore("Smith", "Smith"),
		},
		{
			Name:          "MRA_William_Willyam_match",
			Algorithm:     "MRA",
			A:             "William",
			B:             "Willyam",
			ExpectedScore: fuzzymatch.MRAScore("William", "Willyam"),
		},
		{
			Name:          "MRA_both_empty",
			Algorithm:     "MRA",
			A:             "",
			B:             "",
			ExpectedScore: fuzzymatch.MRAScore("", ""),
		},
		{
			Name:          "MRA_one_empty_a",
			Algorithm:     "MRA",
			A:             "",
			B:             "Smith",
			ExpectedScore: fuzzymatch.MRAScore("", "Smith"),
		},
		{
			Name:          "MRA_one_empty_b",
			Algorithm:     "MRA",
			A:             "Smith",
			B:             "",
			ExpectedScore: fuzzymatch.MRAScore("Smith", ""),
		},
	}
}

// TestGolden_MRA_Staging produces testdata/golden/_staging/mra.json.
// Entries are sorted alphabetically by Name. Ten entries cover the
// canonical NBS Tech Note 943 / Lynch & Arends 1977 reference pairs
// (Smith/Smyth, Byrne/Boern, William/Willyam, Catherine/Katherine), the
// Smith/Jones discriminating non-match, the Ad/ZachariahMontgomery
// length-gate non-match (|len(A)-len(B)| > 3 → fails), Smith identity,
// and the standard catalogue edge cases.
//
// Run with `-update` to create or refresh the staging file.
// Re-running without `-update` must exit 0 (file is byte-stable).
func TestGolden_MRA_Staging(t *testing.T) {
	entries := buildMRAStagingEntries(t)
	sort.Slice(entries, func(i, j int) bool {
		return entries[i].Name < entries[j].Name
	})
	file := goldenAlgorithmsFile{Version: 1, Entries: entries}
	assertGoldenStaging(t, "_staging/mra.json", file)
}

// buildNYSIISStagingEntries returns ten NYSIIS entries drawn from
// nysiis_test.go's reference vectors (Taft 1970, NY State Special Report
// No. 1) plus the catalogue edge cases. NYSIISScore wraps the code-
// comparison into the canonical 0.0 / 1.0 binary so the standard
// goldenAlgorithmEntry shape applies.
func buildNYSIISStagingEntries(t *testing.T) []goldenAlgorithmEntry {
	t.Helper()
	return []goldenAlgorithmEntry{
		{
			Name:          "NYSIIS_Brown_Browne_match",
			Algorithm:     "NYSIIS",
			A:             "Brown",
			B:             "Browne",
			ExpectedScore: fuzzymatch.NYSIISScore("Brown", "Browne"),
		},
		{
			Name:          "NYSIIS_Brown_Robert_nomatch",
			Algorithm:     "NYSIIS",
			A:             "Brown",
			B:             "Robert",
			ExpectedScore: fuzzymatch.NYSIISScore("Brown", "Robert"),
		},
		{
			Name:          "NYSIIS_Brown_both",
			Algorithm:     "NYSIIS",
			A:             "Brown",
			B:             "Brown",
			ExpectedScore: fuzzymatch.NYSIISScore("Brown", "Brown"),
		},
		{
			Name:          "NYSIIS_Catherine_Katherine_match",
			Algorithm:     "NYSIIS",
			A:             "Catherine",
			B:             "Katherine",
			ExpectedScore: fuzzymatch.NYSIISScore("Catherine", "Katherine"),
		},
		{
			Name:          "NYSIIS_John_Jon_match",
			Algorithm:     "NYSIIS",
			A:             "John",
			B:             "Jon",
			ExpectedScore: fuzzymatch.NYSIISScore("John", "Jon"),
		},
		{
			Name:          "NYSIIS_Robert_both",
			Algorithm:     "NYSIIS",
			A:             "Robert",
			B:             "Robert",
			ExpectedScore: fuzzymatch.NYSIISScore("Robert", "Robert"),
		},
		{
			Name:          "NYSIIS_Smith_Jones_nomatch",
			Algorithm:     "NYSIIS",
			A:             "Smith",
			B:             "Jones",
			ExpectedScore: fuzzymatch.NYSIISScore("Smith", "Jones"),
		},
		{
			Name:          "NYSIIS_both_empty",
			Algorithm:     "NYSIIS",
			A:             "",
			B:             "",
			ExpectedScore: fuzzymatch.NYSIISScore("", ""),
		},
		{
			Name:          "NYSIIS_one_empty_a",
			Algorithm:     "NYSIIS",
			A:             "",
			B:             "Brown",
			ExpectedScore: fuzzymatch.NYSIISScore("", "Brown"),
		},
		{
			Name:          "NYSIIS_one_empty_b",
			Algorithm:     "NYSIIS",
			A:             "Brown",
			B:             "",
			ExpectedScore: fuzzymatch.NYSIISScore("Brown", ""),
		},
	}
}

// TestGolden_NYSIIS_Staging produces testdata/golden/_staging/nysiis.json.
// Entries are sorted alphabetically by Name. Ten entries cover the Taft
// 1970 reference pairs (Brown/Browne, Catherine/Katherine, John/Jon),
// identity (Brown/Brown, Robert/Robert), the Smith/Jones and Brown/Robert
// non-matches, and the standard catalogue edge cases.
//
// Run with `-update` to create or refresh the staging file.
// Re-running without `-update` must exit 0 (file is byte-stable).
func TestGolden_NYSIIS_Staging(t *testing.T) {
	entries := buildNYSIISStagingEntries(t)
	sort.Slice(entries, func(i, j int) bool {
		return entries[i].Name < entries[j].Name
	})
	file := goldenAlgorithmsFile{Version: 1, Entries: entries}
	assertGoldenStaging(t, "_staging/nysiis.json", file)
}

// buildPartialRatioStagingEntries returns ten PartialRatio entries drawn
// from partial_ratio_test.go's reference vectors (RapidFuzz canonical
// alignment pairs + 06-RESEARCH.md Pitfall 3 KEYSTONE three-region
// regression gate) plus the catalogue edge cases. PartialRatioScore is
// the byte-path surface bound to dispatch[AlgoPartialRatio]; the rune-
// path PartialRatioScoreRunes is exercised by separate unit tests.
func buildPartialRatioStagingEntries(t *testing.T) []goldenAlgorithmEntry {
	t.Helper()
	return []goldenAlgorithmEntry{
		{
			Name:          "PartialRatio_both_empty",
			Algorithm:     "PartialRatio",
			A:             "",
			B:             "",
			ExpectedScore: fuzzymatch.PartialRatioScore("", ""),
		},
		{
			Name:          "PartialRatio_disjoint_no_overlap",
			Algorithm:     "PartialRatio",
			A:             "abc",
			B:             "xyzzz",
			ExpectedScore: fuzzymatch.PartialRatioScore("abc", "xyzzz"),
		},
		{
			Name:          "PartialRatio_identity",
			Algorithm:     "PartialRatio",
			A:             "abc",
			B:             "abc",
			ExpectedScore: fuzzymatch.PartialRatioScore("abc", "abc"),
		},
		{
			Name:          "PartialRatio_one_empty_a",
			Algorithm:     "PartialRatio",
			A:             "",
			B:             "hello",
			ExpectedScore: fuzzymatch.PartialRatioScore("", "hello"),
		},
		{
			Name:          "PartialRatio_one_empty_b",
			Algorithm:     "PartialRatio",
			A:             "hello",
			B:             "",
			ExpectedScore: fuzzymatch.PartialRatioScore("hello", ""),
		},
		{
			Name:          "PartialRatio_partial_overlap",
			Algorithm:     "PartialRatio",
			A:             "abcd",
			B:             "xabcy",
			ExpectedScore: fuzzymatch.PartialRatioScore("abcd", "xabcy"),
		},
		{
			Name:          "PartialRatio_region_1_left_tail_pitfall_3",
			Algorithm:     "PartialRatio",
			A:             "abc",
			B:             "ab",
			ExpectedScore: fuzzymatch.PartialRatioScore("abc", "ab"),
		},
		{
			Name:          "PartialRatio_region_2_middle_wins",
			Algorithm:     "PartialRatio",
			A:             "YANKEES",
			B:             "NEW YORK YANKEES",
			ExpectedScore: fuzzymatch.PartialRatioScore("YANKEES", "NEW YORK YANKEES"),
		},
		{
			Name:          "PartialRatio_region_3_right_tail_pitfall_3",
			Algorithm:     "PartialRatio",
			A:             "abc",
			B:             "bc",
			ExpectedScore: fuzzymatch.PartialRatioScore("abc", "bc"),
		},
		{
			Name:          "PartialRatio_subset_short_at_end",
			Algorithm:     "PartialRatio",
			A:             "world",
			B:             "hello world",
			ExpectedScore: fuzzymatch.PartialRatioScore("world", "hello world"),
		},
	}
}

// TestGolden_PartialRatio_Staging produces
// testdata/golden/_staging/partial_ratio.json. Entries are sorted
// alphabetically by Name. Ten entries cover: identity, both-empty,
// two one-empty variants, the RapidFuzz canonical YANKEES /
// NEW YORK YANKEES Region 2 win, the 06-RESEARCH.md Pitfall 3 KEYSTONE
// Region 1 (left-tail) and Region 3 (right-tail) regression gates,
// partial-overlap with shifted alignment, subset-at-end, and the
// disjoint-char-set short-circuit pair.
//
// Run with `-update` to create or refresh the staging file.
// Re-running without `-update` must exit 0 (file is byte-stable).
func TestGolden_PartialRatio_Staging(t *testing.T) {
	entries := buildPartialRatioStagingEntries(t)
	sort.Slice(entries, func(i, j int) bool {
		return entries[i].Name < entries[j].Name
	})
	file := goldenAlgorithmsFile{Version: 1, Entries: entries}
	assertGoldenStaging(t, "_staging/partial_ratio.json", file)
}

// buildSoundexStagingEntries returns ten Soundex entries drawn from
// soundex_test.go's reference vectors (Knuth 1973 §6.4 worked examples
// plus the Russell/Odell 1918 patent pairs) and the catalogue edge cases.
// SoundexScore wraps the four-character-code comparison into the
// canonical 0.0 / 1.0 binary so the standard goldenAlgorithmEntry shape
// applies.
func buildSoundexStagingEntries(t *testing.T) []goldenAlgorithmEntry {
	t.Helper()
	return []goldenAlgorithmEntry{
		{
			Name:          "Soundex_Ashcraft_Ashcroft_match",
			Algorithm:     "Soundex",
			A:             "Ashcraft",
			B:             "Ashcroft",
			ExpectedScore: fuzzymatch.SoundexScore("Ashcraft", "Ashcroft"),
		},
		{
			Name:          "Soundex_Catherine_Katherine_nomatch",
			Algorithm:     "Soundex",
			A:             "Catherine",
			B:             "Katherine",
			ExpectedScore: fuzzymatch.SoundexScore("Catherine", "Katherine"),
		},
		{
			Name:          "Soundex_Robert_Rupert_match",
			Algorithm:     "Soundex",
			A:             "Robert",
			B:             "Rupert",
			ExpectedScore: fuzzymatch.SoundexScore("Robert", "Rupert"),
		},
		{
			Name:          "Soundex_Robert_Smith_nomatch",
			Algorithm:     "Soundex",
			A:             "Robert",
			B:             "Smith",
			ExpectedScore: fuzzymatch.SoundexScore("Robert", "Smith"),
		},
		{
			Name:          "Soundex_Smith_Jones_nomatch",
			Algorithm:     "Soundex",
			A:             "Smith",
			B:             "Jones",
			ExpectedScore: fuzzymatch.SoundexScore("Smith", "Jones"),
		},
		{
			Name:          "Soundex_Tymczak_self_match",
			Algorithm:     "Soundex",
			A:             "Tymczak",
			B:             "Tymczak",
			ExpectedScore: fuzzymatch.SoundexScore("Tymczak", "Tymczak"),
		},
		{
			Name:          "Soundex_both_empty",
			Algorithm:     "Soundex",
			A:             "",
			B:             "",
			ExpectedScore: fuzzymatch.SoundexScore("", ""),
		},
		{
			Name:          "Soundex_identical_Robert",
			Algorithm:     "Soundex",
			A:             "Robert",
			B:             "Robert",
			ExpectedScore: fuzzymatch.SoundexScore("Robert", "Robert"),
		},
		{
			Name:          "Soundex_one_empty_a",
			Algorithm:     "Soundex",
			A:             "",
			B:             "Robert",
			ExpectedScore: fuzzymatch.SoundexScore("", "Robert"),
		},
		{
			Name:          "Soundex_one_empty_b",
			Algorithm:     "Soundex",
			A:             "Robert",
			B:             "",
			ExpectedScore: fuzzymatch.SoundexScore("Robert", ""),
		},
	}
}

// TestGolden_Soundex_Staging produces testdata/golden/_staging/soundex.json.
// Entries are sorted alphabetically by Name. Ten entries cover Knuth 1973
// §6.4 worked-example pairs (Ashcraft/Ashcroft, Tymczak, Robert/Rupert),
// the Catherine/Katherine non-match (a load-bearing differentiator from
// NYSIIS — Soundex codes C635 vs K635 differ at index 0, so the binary
// drops to 0), Robert/Smith and Smith/Jones non-matches, identity, and
// the catalogue edge cases.
//
// Run with `-update` to create or refresh the staging file.
// Re-running without `-update` must exit 0 (file is byte-stable).
func TestGolden_Soundex_Staging(t *testing.T) {
	entries := buildSoundexStagingEntries(t)
	sort.Slice(entries, func(i, j int) bool {
		return entries[i].Name < entries[j].Name
	})
	file := goldenAlgorithmsFile{Version: 1, Entries: entries}
	assertGoldenStaging(t, "_staging/soundex.json", file)
}

// buildTokenJaccardStagingEntries returns ten TokenJaccard entries drawn
// from token_jaccard_test.go's reference vectors (RV-TJ1 through RV-TJ6)
// plus the catalogue edge cases. TokenJaccardScore is the surface bound
// to dispatch[AlgoTokenJaccard]; the set-vs-multiset KEYSTONE entry
// (RV-TJ3) is the load-bearing regression gate that distinguishes this
// algorithm from a multiset-based variant.
func buildTokenJaccardStagingEntries(t *testing.T) []goldenAlgorithmEntry {
	t.Helper()
	return []goldenAlgorithmEntry{
		{
			Name:          "TokenJaccard_RV-TJ1_partial_overlap",
			Algorithm:     "TokenJaccard",
			A:             "a b c",
			B:             "b c d",
			ExpectedScore: fuzzymatch.TokenJaccardScore("a b c", "b c d"),
		},
		{
			Name:          "TokenJaccard_RV-TJ2_subset",
			Algorithm:     "TokenJaccard",
			A:             "a b",
			B:             "a b c",
			ExpectedScore: fuzzymatch.TokenJaccardScore("a b", "a b c"),
		},
		{
			Name:          "TokenJaccard_RV-TJ3_set_vs_multiset_keystone",
			Algorithm:     "TokenJaccard",
			A:             "a a b",
			B:             "a b",
			ExpectedScore: fuzzymatch.TokenJaccardScore("a a b", "a b"),
		},
		{
			Name:          "TokenJaccard_RV-TJ4_disjoint",
			Algorithm:     "TokenJaccard",
			A:             "a b c",
			B:             "x y z",
			ExpectedScore: fuzzymatch.TokenJaccardScore("a b c", "x y z"),
		},
		{
			Name:          "TokenJaccard_RV-TJ5_identity",
			Algorithm:     "TokenJaccard",
			A:             "a b c",
			B:             "a b c",
			ExpectedScore: fuzzymatch.TokenJaccardScore("a b c", "a b c"),
		},
		{
			Name:          "TokenJaccard_RV-TJ6_partial_overlap_greek",
			Algorithm:     "TokenJaccard",
			A:             "alpha beta gamma delta",
			B:             "alpha beta epsilon zeta",
			ExpectedScore: fuzzymatch.TokenJaccardScore("alpha beta gamma delta", "alpha beta epsilon zeta"),
		},
		{
			Name:          "TokenJaccard_both_empty",
			Algorithm:     "TokenJaccard",
			A:             "",
			B:             "",
			ExpectedScore: fuzzymatch.TokenJaccardScore("", ""),
		},
		{
			Name:          "TokenJaccard_one_empty_a",
			Algorithm:     "TokenJaccard",
			A:             "",
			B:             "hello",
			ExpectedScore: fuzzymatch.TokenJaccardScore("", "hello"),
		},
		{
			Name:          "TokenJaccard_one_empty_b",
			Algorithm:     "TokenJaccard",
			A:             "hello",
			B:             "",
			ExpectedScore: fuzzymatch.TokenJaccardScore("hello", ""),
		},
		{
			Name:          "TokenJaccard_token_reorder",
			Algorithm:     "TokenJaccard",
			A:             "alpha beta gamma",
			B:             "gamma alpha beta",
			ExpectedScore: fuzzymatch.TokenJaccardScore("alpha beta gamma", "gamma alpha beta"),
		},
	}
}

// TestGolden_TokenJaccard_Staging produces
// testdata/golden/_staging/token_jaccard.json. Entries are sorted
// alphabetically by Name. Ten entries cover RV-TJ1 (partial overlap),
// RV-TJ2 (subset), RV-TJ3 (set-vs-multiset KEYSTONE — the load-bearing
// regression gate distinguishing the set-based variant from a multiset
// variant), RV-TJ4 (disjoint), RV-TJ5 (identity), RV-TJ6 (partial overlap
// with Greek tokens), token reorder, and the standard catalogue edge
// cases.
//
// Run with `-update` to create or refresh the staging file.
// Re-running without `-update` must exit 0 (file is byte-stable).
func TestGolden_TokenJaccard_Staging(t *testing.T) {
	entries := buildTokenJaccardStagingEntries(t)
	sort.Slice(entries, func(i, j int) bool {
		return entries[i].Name < entries[j].Name
	})
	file := goldenAlgorithmsFile{Version: 1, Entries: entries}
	assertGoldenStaging(t, "_staging/token_jaccard.json", file)
}

// buildTokenSetRatioStagingEntries returns twelve TokenSetRatio entries
// drawn from token_set_ratio_test.go's reference vectors (RapidFuzz
// canonical three-way-max alignment plus the RapidFuzz issue #110
// empty-input deviation) and the catalogue edge cases. TokenSetRatioScore
// is the surface bound to dispatch[AlgoTokenSetRatio].
//
// The both-empty and both-pure-separator entries record the RapidFuzz
// issue #110 deviation (returns 0.0, NOT 1.0) — the empty-input gate
// fires BEFORE the identity short-circuit by design, mirroring RapidFuzz's
// reference implementation per CONTEXT.md §3 LOCKED.
func buildTokenSetRatioStagingEntries(t *testing.T) []goldenAlgorithmEntry {
	t.Helper()
	return []goldenAlgorithmEntry{
		{
			Name:          "TokenSetRatio_both_empty_strings_deviation",
			Algorithm:     "TokenSetRatio",
			A:             "",
			B:             "",
			ExpectedScore: fuzzymatch.TokenSetRatioScore("", ""),
		},
		{
			Name:          "TokenSetRatio_both_pure_separator_deviation",
			Algorithm:     "TokenSetRatio",
			A:             " ",
			B:             "  ",
			ExpectedScore: fuzzymatch.TokenSetRatioScore(" ", "  "),
		},
		{
			Name:          "TokenSetRatio_dedup_set_equal",
			Algorithm:     "TokenSetRatio",
			A:             "alpha beta",
			B:             "beta alpha alpha beta",
			ExpectedScore: fuzzymatch.TokenSetRatioScore("alpha beta", "beta alpha alpha beta"),
		},
		{
			Name:          "TokenSetRatio_disjoint",
			Algorithm:     "TokenSetRatio",
			A:             "abc def",
			B:             "xyz qrs",
			ExpectedScore: fuzzymatch.TokenSetRatioScore("abc def", "xyz qrs"),
		},
		{
			Name:          "TokenSetRatio_identity",
			Algorithm:     "TokenSetRatio",
			A:             "hello world",
			B:             "hello world",
			ExpectedScore: fuzzymatch.TokenSetRatioScore("hello world", "hello world"),
		},
		{
			Name:          "TokenSetRatio_low_overlap_singletons",
			Algorithm:     "TokenSetRatio",
			A:             "hello",
			B:             "world",
			ExpectedScore: fuzzymatch.TokenSetRatioScore("hello", "world"),
		},
		{
			Name:          "TokenSetRatio_one_empty_a",
			Algorithm:     "TokenSetRatio",
			A:             "",
			B:             "hello world",
			ExpectedScore: fuzzymatch.TokenSetRatioScore("", "hello world"),
		},
		{
			Name:          "TokenSetRatio_one_empty_b",
			Algorithm:     "TokenSetRatio",
			A:             "hello world",
			B:             "",
			ExpectedScore: fuzzymatch.TokenSetRatioScore("hello world", ""),
		},
		{
			Name:          "TokenSetRatio_subset_a_in_b",
			Algorithm:     "TokenSetRatio",
			A:             "alpha beta",
			B:             "alpha beta gamma",
			ExpectedScore: fuzzymatch.TokenSetRatioScore("alpha beta", "alpha beta gamma"),
		},
		{
			Name:          "TokenSetRatio_subset_b_in_a",
			Algorithm:     "TokenSetRatio",
			A:             "alpha beta gamma",
			B:             "alpha beta",
			ExpectedScore: fuzzymatch.TokenSetRatioScore("alpha beta gamma", "alpha beta"),
		},
		{
			Name:          "TokenSetRatio_three_way_max_combined_wins",
			Algorithm:     "TokenSetRatio",
			A:             "hello world",
			B:             "world peace",
			ExpectedScore: fuzzymatch.TokenSetRatioScore("hello world", "world peace"),
		},
		{
			Name:          "TokenSetRatio_unicode_reorder",
			Algorithm:     "TokenSetRatio",
			A:             "café société",
			B:             "société café",
			ExpectedScore: fuzzymatch.TokenSetRatioScore("café société", "société café"),
		},
	}
}

// TestGolden_TokenSetRatio_Staging produces
// testdata/golden/_staging/token_set_ratio.json. Entries are sorted
// alphabetically by Name. Twelve entries cover the RapidFuzz canonical
// three-way max alignment (intersection / r1 / r2 / r3 combined), the
// RapidFuzz issue #110 empty-input deviation (returns 0.0, NOT 1.0) for
// both-empty-strings and both-pure-separator variants, dedup_set_equal
// (subset short-circuit when sets coincide), disjoint, identity, low-
// overlap singletons, two subset variants, two one-empty variants, and
// the unicode reorder pair.
//
// Run with `-update` to create or refresh the staging file.
// Re-running without `-update` must exit 0 (file is byte-stable).
func TestGolden_TokenSetRatio_Staging(t *testing.T) {
	entries := buildTokenSetRatioStagingEntries(t)
	sort.Slice(entries, func(i, j int) bool {
		return entries[i].Name < entries[j].Name
	})
	file := goldenAlgorithmsFile{Version: 1, Entries: entries}
	assertGoldenStaging(t, "_staging/token_set_ratio.json", file)
}

// buildTokenSortRatioStagingEntries returns ten TokenSortRatio entries
// drawn from token_sort_ratio_test.go's reference vectors (RapidFuzz
// canonical reorder-invariance pairs and the FuzzyWuzzy "fuzzy wuzzy was
// a bear" reorder canonical) plus the catalogue edge cases.
// TokenSortRatioScore is the surface bound to dispatch[AlgoTokenSortRatio].
func buildTokenSortRatioStagingEntries(t *testing.T) []goldenAlgorithmEntry {
	t.Helper()
	return []goldenAlgorithmEntry{
		{
			Name:          "TokenSortRatio_both_empty",
			Algorithm:     "TokenSortRatio",
			A:             "",
			B:             "",
			ExpectedScore: fuzzymatch.TokenSortRatioScore("", ""),
		},
		{
			Name:          "TokenSortRatio_disjoint",
			Algorithm:     "TokenSortRatio",
			A:             "abc",
			B:             "xyz",
			ExpectedScore: fuzzymatch.TokenSortRatioScore("abc", "xyz"),
		},
		{
			Name:          "TokenSortRatio_identity",
			Algorithm:     "TokenSortRatio",
			A:             "hello world",
			B:             "hello world",
			ExpectedScore: fuzzymatch.TokenSortRatioScore("hello world", "hello world"),
		},
		{
			Name:          "TokenSortRatio_low_overlap",
			Algorithm:     "TokenSortRatio",
			A:             "hello",
			B:             "world",
			ExpectedScore: fuzzymatch.TokenSortRatioScore("hello", "world"),
		},
		{
			Name:          "TokenSortRatio_one_empty_a",
			Algorithm:     "TokenSortRatio",
			A:             "",
			B:             "hello world",
			ExpectedScore: fuzzymatch.TokenSortRatioScore("", "hello world"),
		},
		{
			Name:          "TokenSortRatio_one_empty_b",
			Algorithm:     "TokenSortRatio",
			A:             "hello world",
			B:             "",
			ExpectedScore: fuzzymatch.TokenSortRatioScore("hello world", ""),
		},
		{
			Name:          "TokenSortRatio_subset_mid",
			Algorithm:     "TokenSortRatio",
			A:             "alpha beta",
			B:             "alpha beta gamma",
			ExpectedScore: fuzzymatch.TokenSortRatioScore("alpha beta", "alpha beta gamma"),
		},
		{
			Name:          "TokenSortRatio_subset_short",
			Algorithm:     "TokenSortRatio",
			A:             "alpha",
			B:             "alpha beta",
			ExpectedScore: fuzzymatch.TokenSortRatioScore("alpha", "alpha beta"),
		},
		{
			Name:          "TokenSortRatio_token_reorder_canonical",
			Algorithm:     "TokenSortRatio",
			A:             "fuzzy wuzzy was a bear",
			B:             "wuzzy fuzzy was a bear",
			ExpectedScore: fuzzymatch.TokenSortRatioScore("fuzzy wuzzy was a bear", "wuzzy fuzzy was a bear"),
		},
		{
			Name:          "TokenSortRatio_token_reorder_two",
			Algorithm:     "TokenSortRatio",
			A:             "alpha beta",
			B:             "beta alpha",
			ExpectedScore: fuzzymatch.TokenSortRatioScore("alpha beta", "beta alpha"),
		},
	}
}

// TestGolden_TokenSortRatio_Staging produces
// testdata/golden/_staging/token_sort_ratio.json. Entries are sorted
// alphabetically by Name. Ten entries cover identity, both-empty, two
// one-empty variants, two token-reorder pairs (two-token swap +
// FuzzyWuzzy "fuzzy wuzzy was a bear" canonical), two subset pairs
// (short and mid), disjoint, and low overlap singletons.
//
// Run with `-update` to create or refresh the staging file.
// Re-running without `-update` must exit 0 (file is byte-stable).
func TestGolden_TokenSortRatio_Staging(t *testing.T) {
	entries := buildTokenSortRatioStagingEntries(t)
	sort.Slice(entries, func(i, j int) bool {
		return entries[i].Name < entries[j].Name
	})
	file := goldenAlgorithmsFile{Version: 1, Entries: entries}
	assertGoldenStaging(t, "_staging/token_sort_ratio.json", file)
}
