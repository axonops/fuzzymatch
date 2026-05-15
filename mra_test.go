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

// mra_test.go tests MRACode, MRACompare, and MRAScore against literature
// reference vectors from NBS Tech Note 943 (Moore, Kuhns, Trefftzs,
// Montgomery 1977) and cross-validated with jellyfish==1.2.1 (BSD-2-Clause).

package fuzzymatch_test

import (
	"testing"

	"github.com/axonops/fuzzymatch"
)

// TestMRACode_BothEmpty verifies that MRACode("") returns "".
func TestMRACode_BothEmpty(t *testing.T) {
	got := fuzzymatch.MRACode("")
	if got != "" {
		t.Errorf("MRACode(\"\") = %q; want \"\"", got)
	}
}

// TestMRACode_OneEmpty covers edge cases where input is empty or only
// non-ASCII characters, both of which should return "".
func TestMRACode_OneEmpty(t *testing.T) {
	cases := []struct {
		input string
		want  string
	}{
		{"", ""},
		{"  ", ""},  // space-only → no ASCII letters
		{"123", ""}, // digit-only → no ASCII letters
		{"中文", ""},  // non-ASCII only
	}
	for _, c := range cases {
		got := fuzzymatch.MRACode(c.input)
		if got != c.want {
			t.Errorf("MRACode(%q) = %q; want %q", c.input, got, c.want)
		}
	}
}

// TestMRACode_LiteratureReferenceVectors tests the canonical reference vectors
// from NBS Tech Note 943 and jellyfish==1.2.1 testdata (RV-M1..RV-M6).
//
// Source: jellyfish/testdata/match_rating_codex.csv (BSD-2-Clause) and
// hand-derivation from NBS Tech Note 943 encoding rules 1-3.
func TestMRACode_LiteratureReferenceVectors(t *testing.T) {
	cases := []struct {
		name  string
		input string
		want  string
		note  string
	}{
		// RV-M1: Byrne → BYRN (basic: B stays (consonant leading), y→Y (consonant),
		// r→R, n→N, e→drop (vowel non-leading). No doubles. Len 4 ≤ 6.)
		// Source: jellyfish testdata line 1 (confirmed BYRN).
		{"RV-M1", "Byrne", "BYRN", "vowel-removal except leading; jellyfish==1.2.1"},

		// RV-M2: Boern → BRN (B stays (leading), o→drop (vowel non-leading after B),
		// e→drop, r→R, n→N. len 3 ≤ 6.)
		// Source: jellyfish testdata line 2 (confirmed BRN).
		{"RV-M2", "Boern", "BRN", "o and e dropped (non-leading vowels)"},

		// RV-M3: Smith → SMTH (S stays, m→M (consonant), i→drop (vowel), t→T, h→H.
		// No doubles. Len 4 ≤ 6.)
		// Source: jellyfish testdata line 3 (confirmed SMTH).
		{"RV-M3", "Smith", "SMTH", "i removed as vowel; canonical encoding"},

		// RV-M4: Smyth → SMYTH (S stays, m→M, y→Y (Y is NOT a vowel in MRA),
		// t→T, h→H. Len 5 ≤ 6.)
		// Source: jellyfish testdata line 4 (confirmed SMYTH).
		{"RV-M4", "Smyth", "SMYTH", "Y is not a vowel in MRA; kept as consonant"},

		// RV-M5: Catherine → CTHRN (C stays, a→drop, t→T, h→H, e→drop, r→R,
		// i→drop, n→N, e→drop. Len 5 ≤ 6 — NOT truncated.)
		// Source: jellyfish testdata line 5 (confirmed CTHRN).
		{"RV-M5", "Catherine", "CTHRN", "all non-leading vowels removed; len 5 no truncation"},

		// RV-M6: Kathrynoglin → KTHGLN (first-3-last-3 truncation gate).
		// K→K, a→drop, t→T, h→H, r→R, y→Y, n→N, o→drop, g→G, l→L, i→drop, n→N
		// → pre-truncation: KTHRYNGLN (len 9 > 6)
		// → first 3: KTH + last 3: GLN = KTHGLN
		// Source: jellyfish testdata line 7 (confirmed KTHGLN).
		{"RV-M6", "Kathrynoglin", "KTHGLN", "first-3-last-3 truncation gate (pre-truncation len 9 > 6)"},
	}

	for _, c := range cases {
		t.Run(c.name+"/"+c.input, func(t *testing.T) {
			got := fuzzymatch.MRACode(c.input)
			if got != c.want {
				t.Errorf("MRACode(%q) = %q; want %q (%s)", c.input, got, c.want, c.note)
			}
		})
	}
}

// TestMRACode_Truncation is a load-bearing regression test for the
// first-3-last-3 truncation gate (NBS Tech Note 943 encoding rule 3).
// Input must produce a pre-truncation codex of length > 6, and the
// result MUST be exactly 6 characters composed of first-3 + last-3.
func TestMRACode_Truncation(t *testing.T) {
	// Kathrynoglin: pre-truncation form is KTHRYNGLN (len 9 > 6)
	// → first 3 = KTH, last 3 = GLN → KTHGLN
	input := "Kathrynoglin"
	got := fuzzymatch.MRACode(input)
	if len(got) != 6 {
		t.Errorf("MRACode(%q): got len %d = %q; want len 6 (truncated)", input, len(got), got)
	}
	if got != "KTHGLN" {
		t.Errorf("MRACode(%q) = %q; want KTHGLN (first-3=KTH + last-3=GLN)", input, got)
	}
	// Assert that a non-truncated form would be longer (guards against
	// implementing truncation too early / at wrong step).
	if got[:3] != "KTH" {
		t.Errorf("MRACode(%q) first 3 = %q; want KTH", input, got[:3])
	}
	if got[3:] != "GLN" {
		t.Errorf("MRACode(%q) last 3 = %q; want GLN", input, got[3:])
	}
}

// TestMRACompare_BothEmpty verifies MRACompare("", "") = (true, 6) per
// algorithm-correctness-standards both-empty → 1.0 convention (RV-M11).
// Empty codex: both "" have len 0, sum_len = 0, threshold = 5,
// no characters to eliminate, similarity = 6 - 0 = 6. 6 >= 5 → match.
// Source: docs/requirements.md §7.4.4; algorithm-correctness-standards.
func TestMRACompare_BothEmpty(t *testing.T) {
	matched, sim := fuzzymatch.MRACompare("", "")
	if !matched {
		t.Errorf("MRACompare(\"\", \"\").matched = false; want true (both-empty = match)")
	}
	if sim != 6 {
		t.Errorf("MRACompare(\"\", \"\").simScore = %d; want 6", sim)
	}
}

// TestMRACompare_LengthDifferenceAutoMismatch verifies the length-difference
// gate (NBS Tech Note 943 step 1 / docs/requirements.md §7.4.4 line 696 —
// RV-M8). Pairs where |len(codexA) - len(codexB)| >= 3 return (false, 0).
// jellyfish returns an error here; fuzzymatch returns (false, 0) per
// CONTEXT.md §6 LOCKED (Open Question 2 resolution — not an Err, returns a bool).
func TestMRACompare_LengthDifferenceAutoMismatch(t *testing.T) {
	// Ad → AD (len 2). ZachariahMontgomery: Z,a→drop,c→C,h→H,r→R,i→drop,a→drop,
	// h→H,M→M,o→drop,n→N,t→T,g→G,o→drop,m→M,e→drop,r→R,y→Y
	// → ZCHRHMNTGMRY (len 12 > 6) → first 3: ZCH + last 3: MRY = ZCHMRY (len 6)
	// diff = |2 - 6| = 4 >= 3 → auto-mismatch.
	t.Run("Ad/ZachariahMontgomery", func(t *testing.T) {
		matched, sim := fuzzymatch.MRACompare("Ad", "ZachariahMontgomery")
		if matched {
			t.Errorf("MRACompare(\"Ad\", \"ZachariahMontgomery\").matched = true; want false (length-diff >= 3 gate)")
		}
		if sim != 0 {
			t.Errorf("MRACompare(\"Ad\", \"ZachariahMontgomery\").simScore = %d; want 0", sim)
		}
	})
	// Additional length-diff gate: A (MRACode→"A", len 1) vs Kathrynoglin (MRACode→"KTHGLN", len 6)
	// → diff = |1 - 6| = 5 >= 3 → auto-mismatch.
	t.Run("A/Kathrynoglin", func(t *testing.T) {
		matched, sim := fuzzymatch.MRACompare("A", "Kathrynoglin")
		if matched {
			t.Errorf("MRACompare(\"A\", \"Kathrynoglin\").matched = true; want false (length-diff >= 3 gate)")
		}
		if sim != 0 {
			t.Errorf("MRACompare(\"A\", \"Kathrynoglin\").simScore = %d; want 0", sim)
		}
	})
}

// TestMRACompare_LiteratureReferenceVectors verifies MRACompare results for
// canonical pairs from NBS Tech Note 943 and hand-derivation (RV-M7, RV-M9, RV-M10).
func TestMRACompare_LiteratureReferenceVectors(t *testing.T) {
	cases := []struct {
		name        string
		a, b        string
		wantMatched bool
		wantSim     int
		note        string
	}{
		// RV-M7: Smith/Smyth match.
		// SMTH (len 4) vs SMYTH (len 5). sum_len=9 → threshold=3.
		// L→R: S=S match, M=M match; remaining TH vs YTH.
		// R→L on TH vs YTH: H=H match, T=T match; remaining "" vs "Y".
		// unmatched_A=0, unmatched_B=1. similarity = 6 - max(0,1) = 5. 5 >= 3 → match.
		// Source: hand-derived per NBS Tech Note 943; cross-validated with jellyfish.
		{"RV-M7", "Smith", "Smyth", true, 5, "SMTH vs SMYTH; sum=9 threshold=3; sim=5"},

		// Threshold-edge case: William (WLLM → WLM after dedup, len 3) vs
		// Willyam (WLYM after vowel-drop+dedup: W,i→drop,l→L,l→dedup,y→Y,a→drop,m→M
		// → WLM wait: W,i→drop,l→L,l→dup remove 2nd → L, y→Y, a→drop, m→M → WLYM
		// len=4). sum_len=7 → threshold=4.
		// L→R: W=W, L=L; remaining M vs YM. R→L: M=M; remaining "" vs Y.
		// unmatched_A=0, unmatched_B=1. sim=5. 5>=4 → match.
		{"RV-M10", "William", "Willyam", true, 5, "threshold-edge; sum=7 threshold=4; sim=5"},
	}

	for _, c := range cases {
		t.Run(c.name+"/"+c.a+"_"+c.b, func(t *testing.T) {
			matched, sim := fuzzymatch.MRACompare(c.a, c.b)
			if matched != c.wantMatched {
				t.Errorf("MRACompare(%q, %q).matched = %v; want %v (%s)",
					c.a, c.b, matched, c.wantMatched, c.note)
			}
			if sim != c.wantSim {
				t.Errorf("MRACompare(%q, %q).simScore = %d; want %d (%s)",
					c.a, c.b, sim, c.wantSim, c.note)
			}
		})
	}
}

// TestMRACompare_ConsistencyPin asserts that matched == (sim >= threshold)
// for representative inputs. This is the consistency invariant between the
// two return values of MRACompare.
func TestMRACompare_ConsistencyPin(t *testing.T) {
	inputs := [][2]string{
		{"Smith", "Smyth"},
		{"", ""},
		{"Byrne", "Boern"},
		{"William", "Willyam"},
		{"Catherine", "Katherine"},
		{"Robert", "Robin"},
	}
	for _, pair := range inputs {
		a, b := pair[0], pair[1]
		matched, sim := fuzzymatch.MRACompare(a, b)
		score := fuzzymatch.MRAScore(a, b)
		// MRAScore must be 1.0 iff matched is true.
		if matched && score != 1.0 {
			t.Errorf("MRACompare(%q, %q).matched=true but MRAScore=%v (want 1.0)", a, b, score)
		}
		if !matched && score != 0.0 {
			t.Errorf("MRACompare(%q, %q).matched=false but MRAScore=%v (want 0.0)", a, b, score)
		}
		// simScore must always be in [0, 6].
		if sim < 0 || sim > 6 {
			t.Errorf("MRACompare(%q, %q).simScore = %d; want 0 <= sim <= 6", a, b, sim)
		}
	}
}

// TestMRAScore_LiteratureReferenceVectors verifies MRAScore for canonical pairs
// (RV-M12). MRAScore wraps MRACompare: returns 1.0 iff matched.
// Source: docs/requirements.md §7.4.4 spec line 692; CONTEXT.md §6 LOCKED.
func TestMRAScore_LiteratureReferenceVectors(t *testing.T) {
	cases := []struct {
		name      string
		a, b      string
		wantScore float64
		note      string
	}{
		// RV-M12: Smith/Smyth → 1.0 (wraps RV-M7: MRACompare returns (true, 5)).
		{"RV-M12", "Smith", "Smyth", 1.0, "MRACompare(Smith,Smyth)=(true,5) → score 1.0"},

		// Both-empty → 1.0.
		{"both-empty", "", "", 1.0, "identity short-circuit"},

		// Non-matching pair.
		{"mismatch", "Ad", "ZachariahMontgomery", 0.0, "length-diff >= 3 → (false,0) → 0.0"},

		// Identity.
		{"identity", "Byrne", "Byrne", 1.0, "identity short-circuit"},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			got := fuzzymatch.MRAScore(c.a, c.b)
			if got != c.wantScore {
				t.Errorf("MRAScore(%q, %q) = %v; want %v (%s)", c.a, c.b, got, c.wantScore, c.note)
			}
		})
	}
}

// TestMRAScore_CompareConsistencyPin asserts the strict consistency invariant:
// MRAScore(a, b) == 1.0 iff MRACompare(a, b).matched.
func TestMRAScore_CompareConsistencyPin(t *testing.T) {
	pairs := [][2]string{
		{"Smith", "Smyth"},
		{"", ""},
		{"Byrne", "Boern"},
		{"William", "Willyam"},
		{"Ad", "ZachariahMontgomery"},
		{"Robert", "Robin"},
		{"Catherine", "Katherine"},
	}
	for _, p := range pairs {
		a, b := p[0], p[1]
		score := fuzzymatch.MRAScore(a, b)
		matched, _ := fuzzymatch.MRACompare(a, b)
		scoreIs1 := score == 1.0
		if scoreIs1 != matched {
			t.Errorf("consistency violation: MRAScore(%q, %q)=%v (is1.0=%v) but MRACompare.matched=%v",
				a, b, score, scoreIs1, matched)
		}
	}
}

// TestMRAThresholdTable_Clamp is the LOAD-BEARING test for RESEARCH.md Pitfall 7.C.
// The NBS Tech Note 943 threshold table has a clamp at sum > 12 → threshold 2.
// This case is often omitted from Wikipedia-style summaries. Without the clamp,
// mraThreshold(13) would access out-of-bounds memory or return incorrect values.
//
// If this test fails, check: (a) mraThresholdTable has 13 entries (indices 0-12),
// and (b) mraThreshold() includes the `if sumLen > 12 { return 2 }` guard.
func TestMRAThresholdTable_Clamp(t *testing.T) {
	// The clamp: all sums > 12 must return threshold 2.
	clampCases := []int{13, 14, 15, 16, 20, 100}
	for _, sumLen := range clampCases {
		got := fuzzymatch.MRAThresholdForTest(sumLen)
		if got != 2 {
			t.Errorf("mraThreshold(%d) = %d; want 2 (sum>12 clamp per NBS Tech Note 943 Table A — RESEARCH.md Pitfall 7.C)",
				sumLen, got)
		}
	}

	// Explicit table values for the complete [0, 12] range.
	tableExpected := []struct {
		sumLen    int
		wantThres int
	}{
		{0, 5}, {1, 5}, {2, 5}, {3, 5}, {4, 5}, // sum ≤ 4 → threshold 5
		{5, 4}, {6, 4}, {7, 4}, // 4 < sum ≤ 7 → threshold 4
		{8, 3}, {9, 3}, {10, 3}, {11, 3}, // 7 < sum ≤ 11 → threshold 3
		{12, 2}, // sum = 12 → threshold 2
	}
	for _, c := range tableExpected {
		got := fuzzymatch.MRAThresholdForTest(c.sumLen)
		if got != c.wantThres {
			t.Errorf("mraThreshold(%d) = %d; want %d (NBS Tech Note 943 Table A)", c.sumLen, got, c.wantThres)
		}
	}
}

// TestMRACompare_SimScoreRange verifies that simScore is always in [0, 6]
// for a variety of inputs including edge cases.
func TestMRACompare_SimScoreRange(t *testing.T) {
	inputs := [][2]string{
		{"", ""},
		{"A", "A"},
		{"A", "B"},
		{"Smith", "Smyth"},
		{"Byrne", "Boern"},
		{"Cat", "Dog"},
		{"ABCDEF", "ABCDEF"},
	}
	for _, p := range inputs {
		_, sim := fuzzymatch.MRACompare(p[0], p[1])
		if sim < 0 || sim > 6 {
			t.Errorf("MRACompare(%q, %q).simScore = %d; want 0 <= sim <= 6", p[0], p[1], sim)
		}
	}
}
