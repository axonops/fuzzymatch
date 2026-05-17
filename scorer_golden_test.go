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

// scorer_golden_test.go pins the byte-form output of the Phase 8 composite
// Scorer surface (DefaultScorer + four variant Scorer compositions) across
// the five-platform CI matrix (linux/amd64, linux/arm64, darwin/amd64,
// darwin/arm64, windows/amd64). It is the load-bearing cross-platform
// float-determinism gate for Phase 8 per CONTEXT.md §6 + VALIDATION.md
// "Manual-Only Verifications" #10.
//
// The schema follows the LOCKED canonical-form contract from
// golden_canonical.go (json.MarshalIndent with two-space indent, single
// trailing LF, no BOM, UTF-8). The struct definitions are declared in
// this file (not shared with goldenAlgorithmEntry from
// algorithms_golden_test.go) because the Scorer surface produces a
// composite score + per-algorithm breakdown + scorer-config name, not
// the simple (algorithm, a, b, expected_score) tuple of the algorithms
// golden.
//
// `scoreAll` map keys in the JSON are AlgoID.String() values (e.g.
// "DamerauLevenshteinOSA", NOT integer-form "1") per RESEARCH.md
// Pitfall 6. The public API returns map[AlgoID]float64 (typed enum
// keys) but the JSON serialisation must be human-readable AND
// byte-stable across runs — converting via AlgoID.String() at golden-
// construction time pins the readable form into the on-disk fixture.
//
// The map is built by iterating AlgoIDs() (canonical iota order from
// algoid.go:282-308) so the insertion order is deterministic. The on-
// disk JSON serialises with alphabetical key order (encoding/json sorts
// map keys on marshal), which is also deterministic — so either
// build-order or alphabetical, the on-disk bytes are stable.
//
// `_metadata.generated_at` is INTENTIONALLY OMITTED from the schema per
// RESEARCH.md Open Question 1: algorithms.json has no timestamp, so
// scorer-default.json follows the same pattern. Including a wall-clock
// timestamp would make every -update regen produce a different file
// even when the underlying scores are unchanged — that breaks the
// byte-identical CI matrix gate.

package fuzzymatch_test

import (
	"testing"

	"github.com/axonops/fuzzymatch"
)

// scorerGoldenEntry is one (a, b, score, match, scoreAll, scorer_config)
// tuple in testdata/golden/scorer-default.json. Field names are part of
// the LOCKED JSON contract — renaming any field is a major-version-bump
// event per CONTEXT.md §6 + docs/requirements.md §11.2.
//
// ScoreAll uses map[string]float64 (AlgoID.String() keys) for human-
// readable on-disk form. The public Scorer.ScoreAll API returns
// map[AlgoID]float64; the conversion happens inside
// buildScorerGoldenEntries via AlgoID.String() so the on-disk fixture
// is grep-able without the consumer having to decode integer enum
// values.
type scorerGoldenEntry struct {
	A          string             `json:"a"`
	B          string             `json:"b"`
	Score      float64            `json:"score"`
	Match      bool               `json:"match"`
	ScoreAll   map[string]float64 `json:"scoreAll"`
	ScorerConf string             `json:"scorer_config"`
}

// scorerGoldenMetadata is the fixed-shape preamble of every
// scorer-default.json entry. NO `generated_at` field — see file header
// for rationale (algorithms.json parity + byte-identical CI gate).
type scorerGoldenMetadata struct {
	Phase           int    `json:"phase"`
	ScorerSignature string `json:"scorer_signature"`
}

// scorerGoldenFile is the top-level JSON envelope: metadata then a
// sorted slice of entries. The slice ordering is the insertion order
// in buildScorerGoldenEntries (deterministic by construction), NOT a
// post-build sort.
type scorerGoldenFile struct {
	Metadata scorerGoldenMetadata `json:"_metadata"`
	Entries  []scorerGoldenEntry  `json:"entries"`
}

// scoreAllAsStringKeys converts the public-API map[AlgoID]float64
// returned by Scorer.ScoreAll into a map[string]float64 keyed by
// AlgoID.String(). The conversion is deterministic — Go map iteration
// order is randomised but encoding/json sorts map keys alphabetically
// on marshal, so the on-disk bytes are stable across runs regardless of
// the insertion order here.
//
// Helper exists to keep buildScorerGoldenEntries readable; the
// conversion is one line at every call site.
func scoreAllAsStringKeys(m map[fuzzymatch.AlgoID]float64) map[string]float64 {
	out := make(map[string]float64, len(m))
	for id, score := range m {
		out[id.String()] = score
	}
	return out
}

// makeScorerGoldenEntry runs s.Score / s.Match / s.ScoreAll on the
// (a, b) pair and assembles a scorerGoldenEntry tagged with the
// scorer-config name. The config name is the human-readable label
// (e.g. "DefaultScorer", "DefaultScorer-WithoutNormalisation") that
// downstream reviewers use to disambiguate which Scorer composition
// produced each entry.
func makeScorerGoldenEntry(s *fuzzymatch.Scorer, a, b, configName string) scorerGoldenEntry {
	return scorerGoldenEntry{
		A:          a,
		B:          b,
		Score:      s.Score(a, b),
		Match:      s.Match(a, b),
		ScoreAll:   scoreAllAsStringKeys(s.ScoreAll(a, b)),
		ScorerConf: configName,
	}
}

// buildScorerGoldenEntries returns the full 22-26 entry corpus that
// scorer-default.json captures. The corpus splits into two halves per
// CONTEXT.md §6:
//
//  1. 7 identifier-similarity rows reusing the pairs from
//     examples/identifier-similarity/main.go (the canonical
//     cross-validation corpus shared with the algorithms.json golden).
//  2. 11 Scorer-specific rows covering threshold edges, both-empty /
//     one-empty / identity, Unicode-NFC normalisation, phonetic-only
//     divergent, WithoutNormalisation variant, WithoutAlgorithm
//     variant, custom single-algorithm Scorer, and a raw-weights
//     Scorer that exercises WithNormaliseWeights(false).
//
// All five Scorer compositions are constructed once at the top of the
// function; every entry's makeScorerGoldenEntry call reuses them. The
// composition labels are passed verbatim to scorer_config so reviewers
// can grep the JSON by composition (e.g.
// `jq '.entries[] | select(.scorer_config == "DefaultScorer-MinusDoubleMetaphone")'`).
func buildScorerGoldenEntries(t *testing.T) []scorerGoldenEntry {
	t.Helper()

	defaultS := fuzzymatch.DefaultScorer()

	// Variant 1: DefaultScorer minus pre-comparison Normalisation. The
	// XMLParser vs xml_parser pair scores much lower under this Scorer
	// because the byte-form differences (case + underscore) are not
	// erased before the algorithms run.
	withoutNormS, err := fuzzymatch.NewScorer(append(
		fuzzymatch.DefaultScorerOptions(),
		fuzzymatch.WithoutNormalisation(),
	)...)
	if err != nil {
		t.Fatalf("buildScorerGoldenEntries: NewScorer(WithoutNormalisation) failed: %v", err)
	}

	// Variant 2: DefaultScorer minus DoubleMetaphone — the canonical
	// "default minus phonetic" use case (CONTEXT.md §Specific Ideas).
	// Phonetic-divergent pairs (Smith / Schmidt) score lower under
	// this Scorer because DM's binary phonetic match is removed.
	minusDMS, err := fuzzymatch.NewScorer(append(
		fuzzymatch.DefaultScorerOptions(),
		fuzzymatch.WithoutAlgorithm(fuzzymatch.AlgoDoubleMetaphone),
	)...)
	if err != nil {
		t.Fatalf("buildScorerGoldenEntries: NewScorer(WithoutAlgorithm DM) failed: %v", err)
	}

	// Variant 3: a single-algorithm Scorer with a 0.5 threshold — the
	// minimum viable composition. kitten / sitting is the canonical
	// Levenshtein pair (3 edits / 7 max-len = 4/7 ≈ 0.571 score, just
	// above the 0.5 threshold).
	levOnlyS, err := fuzzymatch.NewScorer(
		fuzzymatch.WithAlgorithm(fuzzymatch.AlgoLevenshtein, 1.0),
		fuzzymatch.WithThreshold(0.5),
	)
	if err != nil {
		t.Fatalf("buildScorerGoldenEntries: NewScorer(Levenshtein-only) failed: %v", err)
	}

	// Variant 4: raw weights (WithNormaliseWeights(false)). The
	// composite score may exceed 1.0 — the entry pins the actual no-
	// clamp value to confirm the documented behaviour from
	// CONTEXT.md §1 / scorer.go's WithNormaliseWeights godoc.
	rawWeightsS, err := fuzzymatch.NewScorer(
		fuzzymatch.WithAlgorithm(fuzzymatch.AlgoLevenshtein, 1.0),
		fuzzymatch.WithAlgorithm(fuzzymatch.AlgoJaroWinkler, 3.0),
		fuzzymatch.WithThreshold(0.5),
		fuzzymatch.WithNormaliseWeights(false),
	)
	if err != nil {
		t.Fatalf("buildScorerGoldenEntries: NewScorer(raw weights) failed: %v", err)
	}

	// Identifier-similarity reuse corpus — same 7 pairs as
	// examples/identifier-similarity/main.go. The order matches the
	// example program's `pairs` slice (CONTEXT.md §6).
	identifierPairs := [][2]string{
		{"user_id", "userId"},
		{"created_at", "creationTimestamp"},
		{"status", "state"},
		{"email", "e_mail"},
		{"org_id", "organisation_id"},
		{"latitude", "longitude"},
		{"is_deleted", "is_active"},
	}

	entries := make([]scorerGoldenEntry, 0, 22)
	for _, p := range identifierPairs {
		entries = append(entries, makeScorerGoldenEntry(defaultS, p[0], p[1], "DefaultScorer"))
	}

	// Scorer-specific mandatory rows per CONTEXT.md §6.
	//
	// Row 8: identity — two identical strings score exactly 1.0
	// regardless of algorithm composition (every algorithm returns
	// 1.0 for a == b).
	entries = append(entries, makeScorerGoldenEntry(defaultS, "hello", "hello", "DefaultScorer"))

	// Row 9: both-empty. The behaviour is algorithm-defined; the
	// Scorer composite reflects the weighted average. We pin the
	// actual value (Phase 1+ documents that both-empty is the
	// vacuous-identity case for distance-based algorithms).
	entries = append(entries, makeScorerGoldenEntry(defaultS, "", "", "DefaultScorer"))

	// Row 10: one-empty. Composite score under DefaultScorer is
	// dominated by the zero-match contributions from
	// edit-distance algorithms.
	entries = append(entries, makeScorerGoldenEntry(defaultS, "", "hello", "DefaultScorer"))

	// Row 11: Unicode-NFC. café / cafe under default normalisation
	// (which strips diacritics) scores high; the entry pins the
	// actual composite to detect any future Unicode-normalisation
	// regression.
	entries = append(entries, makeScorerGoldenEntry(defaultS, "café", "cafe", "DefaultScorer"))

	// Row 12: phonetic-only divergent. Smith / Schmidt — DM full
	// match (binary 1.0), DL-OSA low (significant edit distance).
	// The composite reflects the weighted balance.
	entries = append(entries, makeScorerGoldenEntry(defaultS, "Smith", "Schmidt", "DefaultScorer"))

	// Row 13: configs vs config — chosen for likely just-above-
	// threshold behaviour. The actual score is pinned at -update
	// time; threshold-edge entries are the canonical Phase 8
	// determinism test case.
	entries = append(entries, makeScorerGoldenEntry(defaultS, "config", "configs", "DefaultScorer"))

	// Row 14: completely-different short strings. Composite far
	// below 0.85 threshold; match: false. Pins the "no-match
	// behaviour" floor.
	entries = append(entries, makeScorerGoldenEntry(defaultS, "abc", "xyz", "DefaultScorer"))

	// Row 15: WithoutNormalisation variant. XMLParser / xml_parser
	// scores high under DefaultScorer (Normalise strips case and
	// underscore-vs-camelCase differences); under
	// WithoutNormalisation the raw bytes differ and the score
	// reflects the raw-byte distance.
	entries = append(entries, makeScorerGoldenEntry(withoutNormS, "XMLParser", "xml_parser", "DefaultScorer-WithoutNormalisation"))

	// Row 16: WithoutAlgorithm variant. Smith / Schmidt under
	// DefaultScorer-minus-DoubleMetaphone scores lower than under
	// the full DefaultScorer (the phonetic signal is removed).
	entries = append(entries, makeScorerGoldenEntry(minusDMS, "Smith", "Schmidt", "DefaultScorer-MinusDoubleMetaphone"))

	// Row 17: custom single-algorithm Scorer. Levenshtein-only with
	// threshold 0.5 on the canonical kitten / sitting pair.
	entries = append(entries, makeScorerGoldenEntry(levOnlyS, "kitten", "sitting", "Levenshtein-Only-Threshold-0.5"))

	// Row 18: raw weights (no auto-normalisation). The composite
	// may exceed 1.0; the entry pins the actual value to confirm
	// the no-clamp behaviour.
	entries = append(entries, makeScorerGoldenEntry(rawWeightsS, "user_id", "userId", "Raw-Weights-Lev-1-JW-3-NoNorm"))

	// Row 19: just-below-threshold candidate. The default-0.85
	// threshold is tuned for the 6-algorithm default mix; some
	// pairs land near the boundary. This entry pins the actual
	// composite so future algorithm-internal changes that shift
	// the boundary surface a deliberate -update review.
	entries = append(entries, makeScorerGoldenEntry(defaultS, "abbreviation", "abreviation", "DefaultScorer"))

	// Row 20: phonetic divergent under DefaultScorer-minus-DM.
	// Same Smith / Schmidt pair as row 12 + row 16, but under the
	// raw-weight Scorer for completeness — exercises the third
	// scorer variant on a phonetic-divergent pair.
	entries = append(entries, makeScorerGoldenEntry(rawWeightsS, "Smith", "Schmidt", "Raw-Weights-Lev-1-JW-3-NoNorm"))

	// Row 21: Unicode under WithoutNormalisation. café / cafe with
	// the diacritic-stripping disabled — the bytes differ (UTF-8
	// 'é' is 0xC3 0xA9 vs ASCII 'e') and the composite reflects
	// the raw-byte distance.
	entries = append(entries, makeScorerGoldenEntry(withoutNormS, "café", "cafe", "DefaultScorer-WithoutNormalisation"))

	// Row 22: Levenshtein-only on an identifier pair. The same
	// user_id / userId pair as row 1, but scored by a single-
	// algorithm Scorer — useful for cross-checking that the
	// composite formula reduces to the algorithm's raw output
	// when only one algorithm is configured.
	entries = append(entries, makeScorerGoldenEntry(levOnlyS, "user_id", "userId", "Levenshtein-Only-Threshold-0.5"))

	return entries
}

// TestGolden_ScorerDefault is the cross-platform float-determinism gate
// for the Phase 8 composite Scorer. It constructs DefaultScorer (and
// four variant Scorers), runs Score / Match / ScoreAll over a 22-26
// entry corpus per CONTEXT.md §6, and asserts the canonical-form JSON
// bytes match testdata/golden/scorer-default.json exactly.
//
// Run with `-update` to regenerate the golden file:
//
//	go test -run TestGolden_ScorerDefault -update ./...
//
// Re-running without `-update` MUST exit 0 on every platform in the CI
// matrix. Any byte-level divergence (e.g. FMA fusion on arm64, math.X
// transcendental drift, sort-key instability) fails CI on that
// platform per VALIDATION.md.
func TestGolden_ScorerDefault(t *testing.T) {
	entries := buildScorerGoldenEntries(t)
	assertGolden(t, "scorer-default.json", scorerGoldenFile{
		Metadata: scorerGoldenMetadata{
			Phase:           8,
			ScorerSignature: "DefaultScorer-2026-05-16",
		},
		Entries: entries,
	})
}
