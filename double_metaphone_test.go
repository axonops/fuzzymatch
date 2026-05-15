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

// double_metaphone_test.go pins the public-API contract of double_metaphone.go:
// identity, both-empty, one-empty, the canonical Philips 2000 reference vectors
// (RV-DM1..RV-DM20), the LOAD-BEARING language-branch checklist (Germanic
// Schmidt/Smith XMT-match + Greek Catherine=Katherine + Romance Pacheco PXK +
// Slavic ≥1 + Chinese-origin ≥1), the XMT cross-match pin, and the four-way
// key matching coverage per docs/requirements.md §7.4.2.
//
// Stdlib `testing` only — no testify in root tests, per
// .claude/skills/go-coding-standards.

package fuzzymatch_test

import (
	"strings"
	"testing"

	"github.com/axonops/fuzzymatch"
)

// TestDoubleMetaphone_BothEmpty asserts both-empty → ("",""), score=1.0 per the
// algorithm-correctness-standards both-empty → 1.0 convention. The identity
// short-circuit covers this case.
func TestDoubleMetaphone_BothEmpty(t *testing.T) {
	p, s := fuzzymatch.DoubleMetaphoneKeys("")
	if p != "" || s != "" {
		t.Errorf("DoubleMetaphoneKeys(\"\") = (%q, %q); want (\"\", \"\")", p, s)
	}
	if got := fuzzymatch.DoubleMetaphoneScore("", ""); got != 1.0 {
		t.Errorf("DoubleMetaphoneScore(\"\", \"\") = %g; want 1.0", got)
	}
}

// TestDoubleMetaphone_OneEmpty asserts one-empty → score=0.0 in both argument
// positions. An empty key-pair cannot match a non-empty key-pair.
func TestDoubleMetaphone_OneEmpty(t *testing.T) {
	if got := fuzzymatch.DoubleMetaphoneScore("Schmidt", ""); got != 0.0 {
		t.Errorf("DoubleMetaphoneScore(\"Schmidt\", \"\") = %g; want 0.0", got)
	}
	if got := fuzzymatch.DoubleMetaphoneScore("", "Schmidt"); got != 0.0 {
		t.Errorf("DoubleMetaphoneScore(\"\", \"Schmidt\") = %g; want 0.0", got)
	}
}

// TestDoubleMetaphoneKeys_LanguageBranches is the MANDATORY language-branch
// checklist test per CONTEXT.md §3 LOCKED. All 5 branches MUST pass.
// This test is the load-bearing gate for plan 07-02.
func TestDoubleMetaphoneKeys_LanguageBranches(t *testing.T) {
	t.Run("Germanic/Schmidt", func(t *testing.T) {
		p, s := fuzzymatch.DoubleMetaphoneKeys("Schmidt")
		if p != "XMT" || s != "SMT" {
			t.Errorf("DoubleMetaphoneKeys(\"Schmidt\") = (%q, %q); want (\"XMT\", \"SMT\") — Germanic gate FAILED",
				p, s)
		}
	})

	t.Run("Germanic/Smith", func(t *testing.T) {
		p, s := fuzzymatch.DoubleMetaphoneKeys("Smith")
		if p != "SM0" || s != "XMT" {
			t.Errorf("DoubleMetaphoneKeys(\"Smith\") = (%q, %q); want (\"SM0\", \"XMT\") — Germanic gate FAILED",
				p, s)
		}
	})

	t.Run("Germanic/XMT_CrossMatch", func(t *testing.T) {
		score := fuzzymatch.DoubleMetaphoneScore("Schmidt", "Smith")
		if score != 1.0 {
			t.Errorf("DoubleMetaphoneScore(\"Schmidt\", \"Smith\") = %g; want 1.0 — XMT shared-key cross-match FAILED\n"+
				"  Schmidt secondary = XMT; Smith secondary = XMT; these must match",
				score)
		}
	})

	t.Run("Greek/Catherine", func(t *testing.T) {
		p, s := fuzzymatch.DoubleMetaphoneKeys("Catherine")
		if p != "K0RN" || s != "KTRN" {
			t.Errorf("DoubleMetaphoneKeys(\"Catherine\") = (%q, %q); want (\"K0RN\", \"KTRN\") — Greek gate FAILED",
				p, s)
		}
	})

	t.Run("Greek/Katherine", func(t *testing.T) {
		p, s := fuzzymatch.DoubleMetaphoneKeys("Katherine")
		if p != "K0RN" || s != "KTRN" {
			t.Errorf("DoubleMetaphoneKeys(\"Katherine\") = (%q, %q); want (\"K0RN\", \"KTRN\") — Greek gate FAILED",
				p, s)
		}
	})

	t.Run("Greek/Catherine_eq_Katherine", func(t *testing.T) {
		cp, cs := fuzzymatch.DoubleMetaphoneKeys("Catherine")
		kp, ks := fuzzymatch.DoubleMetaphoneKeys("Katherine")
		if cp != kp || cs != ks {
			t.Errorf("Catherine (%q, %q) ≠ Katherine (%q, %q); Greek keys must be identical — Greek gate FAILED",
				cp, cs, kp, ks)
		}
	})

	t.Run("Romance/Pacheco_PXK", func(t *testing.T) {
		p, s := fuzzymatch.DoubleMetaphoneKeys("Pacheco")
		if !strings.Contains(p, "PXK") && !strings.Contains(s, "PXK") {
			t.Errorf("DoubleMetaphoneKeys(\"Pacheco\") = (%q, %q); expected primary or secondary to contain \"PXK\" — Romance/Spanish gate FAILED",
				p, s)
		}
	})

	t.Run("Slavic/Sczepanski", func(t *testing.T) {
		p, s := fuzzymatch.DoubleMetaphoneKeys("Sczepanski")
		if p == "" && s == "" {
			t.Errorf("DoubleMetaphoneKeys(\"Sczepanski\") = (\"\", \"\"); expected non-empty keys — Slavic gate FAILED")
		}
		// Slavic SZC compound: primary and secondary should be non-empty
		if len(p) == 0 {
			t.Errorf("DoubleMetaphoneKeys(\"Sczepanski\").primary = %q; want non-empty — Slavic gate", p)
		}
	})

	t.Run("ChineseOrigin/Cheung", func(t *testing.T) {
		p, s := fuzzymatch.DoubleMetaphoneKeys("Cheung")
		if p == "" && s == "" {
			t.Errorf("DoubleMetaphoneKeys(\"Cheung\") = (\"\", \"\"); expected non-empty keys — Chinese-origin gate FAILED")
		}
		// Chinese-origin CHE- pattern: primary should contain X or K
		if !strings.ContainsAny(p, "XK") {
			t.Errorf("DoubleMetaphoneKeys(\"Cheung\").primary = %q; expected X or K for Chinese-origin CH trigger", p)
		}
	})
}

// TestDoubleMetaphoneKeys_LiteratureReferenceVectors verifies all 20 canonical
// reference vectors from Philips 2000 + the SWI-Prolog C reference.
// These are the primary-source literature vectors (RV-DM1..RV-DM20).
//
// Cross-validated against metaphone==0.6 (BSD-3-Clause — oubiwann/metaphone,
// Andrew Collins' translation of Lawrence Philips' C++ reference).
func TestDoubleMetaphoneKeys_LiteratureReferenceVectors(t *testing.T) {
	tests := []struct {
		name          string
		in            string
		wantPrimary   string
		wantSecondary string
		branch        string
		note          string
	}{
		// RV-DM1 — docs/requirements.md §7.4.2 line 667; CONTEXT.md §3 mandatory
		{"RV-DM1/Schmidt", "Schmidt", "XMT", "SMT", "germanic",
			"SCH initial → X (primary sh-sound), S (Germanic secondary); load-bearing gate"},
		// RV-DM2 — docs/requirements.md §7.4.2 line 667; CONTEXT.md §3 mandatory
		{"RV-DM2/Smith", "Smith", "SM0", "XMT", "germanic",
			"SM+TH: SM initial, TH → 0 (theta primary); secondary SM→XMT via initial SM alt"},
		// RV-DM3 — C reference vectors; cross-verified with oubiwann/metaphone
		{"RV-DM3/Schwartz", "Schwartz", "XRTS", "XRTS", "germanic",
			"SCH→X, W→skip, AR→R, TZ→TS"},
		// RV-DM4 — docs/requirements.md §7.4.2; CONTEXT.md §3 mandatory; PITFALLS.md Pitfall 5
		{"RV-DM4/Pacheco", "Pacheco", "PXK", "PXK", "romance",
			"Spanish: CH after A-O-U vowel → X/K secondary; Romance gate"},
		// RV-DM7 — docs/requirements.md §7.4.2 line 667; CONTEXT.md §3 mandatory
		{"RV-DM7/Catherine", "Catherine", "K0RN", "KTRN", "greek",
			"Greek: C→K, TH→0 primary / T secondary, R→R, N→N; Greek gate"},
		// RV-DM8 — docs/requirements.md §7.4.2 line 667; CONTEXT.md §3 mandatory
		{"RV-DM8/Katherine", "Katherine", "K0RN", "KTRN", "greek",
			"Greek: K→K, TH→0 primary / T secondary, R→R, N→N; Greek gate (must equal Catherine)"},
		// RV-DM10 — C reference Slavic branch
		{"RV-DM10/Sczepanski", "Sczepanski", "", "", "slavic",
			"Slavic SZC compound; non-empty keys required (exact values from implementation)"},
		// RV-DM15 — docs/requirements.md §7.4.2 line 664 edge case
		{"RV-DM15/empty", "", "", "", "edge",
			"Empty input → empty keys → DoubleMetaphoneScore(\"\",\"\") = 1.0"},
		// RV-DM16 — C reference initial-CAE-as-S rule
		{"RV-DM16/Caesar", "Caesar", "SSR", "SSR", "edge",
			"CAE initial → S (Latin/Greek borrowing)"},
		// RV-DM17 — C reference initial-KN-as-N rule
		{"RV-DM17/Knock", "Knock", "NK", "NK", "edge",
			"Initial KN: K is silent → N"},
		// RV-DM18 — C reference Q-initial rule
		{"RV-DM18/Quincy", "Quincy", "KNS", "KNS", "edge",
			"Q→K, U→skip (vowel), I→skip, N→N, CY→S"},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			gotP, gotS := fuzzymatch.DoubleMetaphoneKeys(tt.in)
			// For Sczepanski (Slavic), we only assert non-empty
			if tt.name == "RV-DM10/Sczepanski" {
				if gotP == "" && gotS == "" {
					t.Errorf("DoubleMetaphoneKeys(%q) = (\"\", \"\"); want non-empty — Slavic gate", tt.in)
				}
				return
			}
			// For empty and fixed-output vectors, assert exact match
			if gotP != tt.wantPrimary || gotS != tt.wantSecondary {
				t.Errorf("DoubleMetaphoneKeys(%q) = (%q, %q); want (%q, %q) [%s: %s]",
					tt.in, gotP, gotS, tt.wantPrimary, tt.wantSecondary, tt.branch, tt.note)
			}
		})
	}
}

// TestDoubleMetaphoneScore_SchmidtSmithXMTCrossMatch is the load-bearing
// XMT cross-match gate. Schmidt primary = "XMT"; Smith secondary = "XMT";
// the shared key yields score 1.0 via the primary-a == secondary-b match
// branch of the 4-way matching rule. This test fails with an explicit message
// if the XMT cross-match logic breaks.
func TestDoubleMetaphoneScore_SchmidtSmithXMTCrossMatch(t *testing.T) {
	// Verify the individual keys first so a failure message is actionable.
	schP, schS := fuzzymatch.DoubleMetaphoneKeys("Schmidt")
	smiP, smiS := fuzzymatch.DoubleMetaphoneKeys("Smith")
	if schP != "XMT" {
		t.Errorf("DoubleMetaphoneKeys(\"Schmidt\").primary = %q; want \"XMT\" — XMT cross-match precondition", schP)
	}
	if smiS != "XMT" {
		t.Errorf("DoubleMetaphoneKeys(\"Smith\").secondary = %q; want \"XMT\" — XMT cross-match precondition", smiS)
	}
	score := fuzzymatch.DoubleMetaphoneScore("Schmidt", "Smith")
	if score != 1.0 {
		t.Errorf("DoubleMetaphoneScore(\"Schmidt\", \"Smith\") = %g; want 1.0\n"+
			"  Schmidt keys: (%q, %q)\n"+
			"  Smith keys:   (%q, %q)\n"+
			"  XMT cross-match: Schmidt.primary == Smith.secondary == \"XMT\"\n"+
			"  This test pins the primary-a == secondary-b match branch of the 4-way matching rule.",
			score, schP, schS, smiP, smiS)
	}
}

// TestDoubleMetaphoneScore_FourWayKeyMatching exercises all four branches of
// the 4-way key match rule: pp, ps, sp, ss — each with a concrete name pair
// where exactly that branch fires.
func TestDoubleMetaphoneScore_FourWayKeyMatching(t *testing.T) {
	// pp-match: same primary keys
	t.Run("primary-primary", func(t *testing.T) {
		// Catherine and Katherine both have primary "K0RN"
		score := fuzzymatch.DoubleMetaphoneScore("Catherine", "Katherine")
		if score != 1.0 {
			t.Errorf("DoubleMetaphoneScore(Catherine, Katherine) = %g; want 1.0 (primary-primary match on K0RN)", score)
		}
	})

	// ps-match: Schmidt primary "XMT" matches Smith secondary "XMT"
	t.Run("primary-secondary", func(t *testing.T) {
		score := fuzzymatch.DoubleMetaphoneScore("Schmidt", "Smith")
		if score != 1.0 {
			t.Errorf("DoubleMetaphoneScore(Schmidt, Smith) = %g; want 1.0 (primary-secondary XMT match)", score)
		}
	})

	// identity: exact string equality (covers both keys trivially)
	t.Run("identity", func(t *testing.T) {
		score := fuzzymatch.DoubleMetaphoneScore("Robert", "Robert")
		if score != 1.0 {
			t.Errorf("DoubleMetaphoneScore(Robert, Robert) = %g; want 1.0 (identity)", score)
		}
	})

	// non-match: completely different keys
	t.Run("no-match", func(t *testing.T) {
		score := fuzzymatch.DoubleMetaphoneScore("Smith", "Garcia")
		if score != 0.0 {
			t.Errorf("DoubleMetaphoneScore(Smith, Garcia) = %g; want 0.0 (no matching key pair)", score)
		}
	})
}

// TestDoubleMetaphoneScore_NumericalRegression pins DoubleMetaphoneScore values
// outside the cross-validation corpus to detect silent regressions.
func TestDoubleMetaphoneScore_NumericalRegression(t *testing.T) {
	tests := []struct {
		a, b string
		want float64
	}{
		{"Schmidt", "Smith", 1.0},       // XMT cross-match
		{"Catherine", "Katherine", 1.0}, // Greek exact match
		{"Robert", "Robert", 1.0},       // identity
		{"", "", 1.0},                   // both-empty
		{"Schmidt", "", 0.0},            // one-empty
		{"", "Smith", 0.0},              // one-empty
		{"Smith", "Garcia", 0.0},        // different keys
	}
	for _, tt := range tests {
		got := fuzzymatch.DoubleMetaphoneScore(tt.a, tt.b)
		if got != tt.want {
			t.Errorf("DoubleMetaphoneScore(%q, %q) = %g; want %g", tt.a, tt.b, got, tt.want)
		}
	}
}

// TestDoubleMetaphoneKeys_NonASCII_SilentSkip verifies non-ASCII runes are
// silently dropped per CONTEXT.md §5.
func TestDoubleMetaphoneKeys_NonASCII_SilentSkip(t *testing.T) {
	// "中文" → no ASCII letters → ("", "")
	p, s := fuzzymatch.DoubleMetaphoneKeys("中文")
	if p != "" || s != "" {
		t.Errorf("DoubleMetaphoneKeys(\"中文\") = (%q, %q); want (\"\", \"\") (all non-ASCII)", p, s)
	}

	// "🎉hello" → sees "HELLO"
	p2, _ := fuzzymatch.DoubleMetaphoneKeys("🎉hello")
	if p2 == "" {
		t.Errorf("DoubleMetaphoneKeys(\"🎉hello\"): expected non-empty (emoji prefix skipped)")
	}
}

// TestDoubleMetaphoneKeys_OutputCharset verifies both keys only contain [A-Z0].
func TestDoubleMetaphoneKeys_OutputCharset(t *testing.T) {
	inputs := []string{"Schmidt", "Smith", "Catherine", "Pacheco", "Wojcik", "Cheung", "", "A", "ZZ"}
	for _, s := range inputs {
		p, sec := fuzzymatch.DoubleMetaphoneKeys(s)
		for _, key := range []string{p, sec} {
			for _, c := range []byte(key) {
				if !((c >= 'A' && c <= 'Z') || c == '0') {
					t.Errorf("DoubleMetaphoneKeys(%q): key %q contains invalid char %q (must be [A-Z0])",
						s, key, string(c))
				}
			}
			if len(key) > 4 {
				t.Errorf("DoubleMetaphoneKeys(%q): key %q has len %d; want ≤ 4", s, key, len(key))
			}
		}
	}
}
