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

// ratcliff_obershelp_test.go pins the public-API contract of
// ratcliff_obershelp.go: identity, both-empty, one-empty, the canonical
// Dr. Dobb's 1988 reference vectors (WIKIMEDIA/WIKIMANIA,
// GESTALT/GESTALT_PATTERN_MATCHING), a numerical-regression pin OUTSIDE the
// cross-validation corpus (WR-03 closure — plan 04-04 ships the corpus),
// the asymmetric-by-design pin (OQ-1 RESOLUTION LOCKED 2026-05-14 — RO is
// intentionally asymmetric to preserve difflib parity), byte-vs-rune
// equivalence on ASCII, and rune multi-byte handling.
//
// Stdlib `testing` only — no testify in root tests, per
// .claude/skills/go-coding-standards.

package fuzzymatch_test

import (
	"encoding/json"
	"math"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/axonops/fuzzymatch"
)

// roCrossValidationEntry is one entry in the committed
// testdata/cross-validation/ratcliff-obershelp/vectors.json corpus produced
// by scripts/gen-ratcliff-obershelp-cross-validation.py. Each entry pins
// our RatcliffObershelpScore byte-path output against Python stdlib
// difflib.SequenceMatcher(autojunk=False).ratio() within 1e-9.
type roCrossValidationEntry struct {
	Name         string  `json:"name"`
	A            string  `json:"a"`
	B            string  `json:"b"`
	DifflibRatio float64 `json:"difflib_ratio"`
}

// roCrossValidationCorpus is the top-level shape of the corpus JSON file
// — see scripts/gen-ratcliff-obershelp-cross-validation.py for the generator
// schema. The Version field is a forward-compatibility integer (currently 1).
// PythonVersion records the sys.version_info string of the Python that
// produced the corpus (e.g. "3.12.4") so cross-validation failures can be
// triaged against an outdated generator.
type roCrossValidationCorpus struct {
	Version       int                      `json:"version"`
	PythonVersion string                   `json:"python_version"`
	Entries       []roCrossValidationEntry `json:"entries"`
}

// TestRatcliffObershelp_BothEmpty pins the both-empty convention:
// RatcliffObershelpScore returns 1.0 (matches difflib.SequenceMatcher
// (autojunk=False, a="", b="").ratio() = 1.0).
func TestRatcliffObershelp_BothEmpty(t *testing.T) {
	if got := fuzzymatch.RatcliffObershelpScore("", ""); got != 1.0 {
		t.Errorf("RatcliffObershelpScore(\"\", \"\") = %g; want 1.0", got)
	}
	if got := fuzzymatch.RatcliffObershelpScoreRunes("", ""); got != 1.0 {
		t.Errorf("RatcliffObershelpScoreRunes(\"\", \"\") = %g; want 1.0", got)
	}
}

// TestRatcliffObershelp_OneEmpty pins the one-empty convention: 0.0 on both
// directions (matches difflib).
func TestRatcliffObershelp_OneEmpty(t *testing.T) {
	tests := []struct{ a, b string }{
		{"", "abc"},
		{"abc", ""},
		{"", "x"},
		{"x", ""},
	}
	for _, tt := range tests {
		t.Run(tt.a+"_"+tt.b, func(t *testing.T) {
			if got := fuzzymatch.RatcliffObershelpScore(tt.a, tt.b); got != 0.0 {
				t.Errorf("RatcliffObershelpScore(%q, %q) = %g; want 0.0", tt.a, tt.b, got)
			}
			if got := fuzzymatch.RatcliffObershelpScoreRunes(tt.a, tt.b); got != 0.0 {
				t.Errorf("RatcliffObershelpScoreRunes(%q, %q) = %g; want 0.0", tt.a, tt.b, got)
			}
		})
	}
}

// TestRatcliffObershelp_Identical pins RatcliffObershelpScore(x, x) == 1.0
// for non-empty x. The identity short-circuit fires BEFORE the DP, and on
// the *Runes variant BEFORE the []rune allocation (IN-04 closure).
func TestRatcliffObershelp_Identical(t *testing.T) {
	tests := []string{"abc", "user_id", "x", "WIKIMEDIA", "GESTALT", "café"}
	for _, s := range tests {
		t.Run(s, func(t *testing.T) {
			if got := fuzzymatch.RatcliffObershelpScore(s, s); got != 1.0 {
				t.Errorf("RatcliffObershelpScore(%q, %q) = %g; want 1.0", s, s, got)
			}
			if got := fuzzymatch.RatcliffObershelpScoreRunes(s, s); got != 1.0 {
				t.Errorf("RatcliffObershelpScoreRunes(%q, %q) = %g; want 1.0", s, s, got)
			}
		})
	}
}

// TestRatcliffObershelp_DrDobbs1988_ReferenceVectors pins the canonical
// Dr. Dobb's 1988 paper-cited pairs. Expected values come from running
// Python `difflib.SequenceMatcher(autojunk=False, a=a, b=b).ratio()` once
// and pasting the values into this test as float64 literals. Tolerance is
// 1e-9 (matches Phase 3 epsilon convention).
//
// Pairs:
//   - WIKIMEDIA / WIKIMANIA  → 0.7777777777777778  (canonical Dr. Dobb's)
//   - GESTALT / GESTALT_PATTERN_MATCHING → 0.45161290322580644
//
// The full cross-validation corpus (plan 04-04) covers more pairs; this
// test pins the two paper-cited canonical pairs locally so a regression
// is caught immediately even before the corpus runs.
func TestRatcliffObershelp_DrDobbs1988_ReferenceVectors(t *testing.T) {
	tests := []struct {
		a, b string
		want float64
	}{
		{"WIKIMEDIA", "WIKIMANIA", 0.7777777777777778},
		{"GESTALT", "GESTALT_PATTERN_MATCHING", 0.45161290322580644},
	}
	const eps = 1e-9
	for _, tt := range tests {
		t.Run(tt.a+"_"+tt.b, func(t *testing.T) {
			got := fuzzymatch.RatcliffObershelpScore(tt.a, tt.b)
			if math.Abs(got-tt.want) > eps {
				t.Errorf("RatcliffObershelpScore(%q, %q) = %.17g; want ≈ %.17g (Δ=%g, eps=%g)",
					tt.a, tt.b, got, tt.want, math.Abs(got-tt.want), eps)
			}
		})
	}
}

// TestRatcliffObershelp_PinnedDrDobbsValue is the Phase 3 WR-03 closure:
// a numerical-regression pin OUTSIDE the cross-validation corpus. Even if
// the cross-validation corpus (plan 04-04) is accidentally accepted with a
// stale value, this test catches the regression directly.
//
// The canonical Dr. Dobb's 1988 WIKIMEDIA/WIKIMANIA pair has the
// well-known ratio 0.7777777777777778 (= 14/18 from M=7 matched chars
// across la=lb=9; 2·7/18 = 14/18). This value is verifiable by hand and
// reproducible with the one-line Python:
//
//	difflib.SequenceMatcher(autojunk=False, a="WIKIMEDIA", b="WIKIMANIA").ratio()
//
// → 0.7777777777777778.
func TestRatcliffObershelp_PinnedDrDobbsValue(t *testing.T) {
	const want = 0.7777777777777778
	const eps = 1e-9
	got := fuzzymatch.RatcliffObershelpScore("WIKIMEDIA", "WIKIMANIA")
	if math.Abs(got-want) > eps {
		t.Errorf("RatcliffObershelpScore(\"WIKIMEDIA\", \"WIKIMANIA\") = %.17g; want ≈ %.17g (Δ=%g, eps=%g)",
			got, want, math.Abs(got-want), eps)
	}
}

// TestRatcliffObershelp_AsymmetryPin is the LOAD-BEARING regression test
// for the OQ-1 resolution (LOCKED 2026-05-14). RatcliffObershelpScore is
// INTENTIONALLY asymmetric in argument order — this mirrors Python's
// difflib.SequenceMatcher.ratio() per CPython bpo-37004.
//
// The canonical asymmetric pair "tide"/"diet": with autojunk=False,
// difflib.SequenceMatcher(a="tide", b="diet").ratio() = 0.25 while
// difflib.SequenceMatcher(a="diet", b="tide").ratio() = 0.5. The fwd != rev
// inequality is the locked contract.
//
// If a future refactor accidentally introduces symmetry (e.g. by sorting
// inputs by length internally), this test fails. The cross-algorithm
// consistency test landing in plan 04-05 adds an inverse-form guard for
// completeness.
func TestRatcliffObershelp_AsymmetryPin(t *testing.T) {
	fwd := fuzzymatch.RatcliffObershelpScore("tide", "diet")
	rev := fuzzymatch.RatcliffObershelpScore("diet", "tide")
	if fwd == rev {
		t.Errorf("RatcliffObershelp expected asymmetric: Score(\"tide\",\"diet\")=%g == Score(\"diet\",\"tide\")=%g; want fwd != rev (OQ-1 resolution)",
			fwd, rev)
	}
	// Pin the exact difflib(autojunk=False) values for additional
	// regression detection.
	const eps = 1e-9
	const wantFwd, wantRev = 0.25, 0.5
	if math.Abs(fwd-wantFwd) > eps {
		t.Errorf("RatcliffObershelpScore(\"tide\",\"diet\") = %.17g; want ≈ %g", fwd, wantFwd)
	}
	if math.Abs(rev-wantRev) > eps {
		t.Errorf("RatcliffObershelpScore(\"diet\",\"tide\") = %.17g; want ≈ %g", rev, wantRev)
	}
}

// TestRatcliffObershelp_ByteVsRune_Equivalence asserts the byte and rune
// paths produce identical output on pure-ASCII inputs (rune count == byte
// count, so the two recursive decompositions coincide).
func TestRatcliffObershelp_ByteVsRune_Equivalence(t *testing.T) {
	tests := [][2]string{
		{"WIKIMEDIA", "WIKIMANIA"},
		{"GESTALT", "GESTALT_PATTERN_MATCHING"},
		{"kitten", "sitting"},
		{"abc", "abc"},
		{"", ""},
		{"", "abc"},
		{"abc", ""},
		{"abc", "xyz"},
		{"tide", "diet"},
		{"http_request", "http_request_header_fields"},
	}
	for _, tt := range tests {
		a, b := tt[0], tt[1]
		t.Run(a+"_"+b, func(t *testing.T) {
			byteScore := fuzzymatch.RatcliffObershelpScore(a, b)
			runeScore := fuzzymatch.RatcliffObershelpScoreRunes(a, b)
			if byteScore != runeScore {
				t.Errorf("byte vs rune score mismatch on ASCII: byte=%g rune=%g for (%q, %q)",
					byteScore, runeScore, a, b)
			}
		})
	}
}

// TestRatcliffObershelp_RuneMultiByte verifies the rune path handles multi-
// byte UTF-8 correctly. "café" is 4 runes / 5 bytes; "cafe" is 4 runes /
// 4 bytes. The rune-path Ratcliff-Obershelp finds the rune-level matched
// substring "caf" (3 runes); score = 2·3/(4+4) = 0.75 (matches Python
// difflib.SequenceMatcher(autojunk=False, a="café", b="cafe").ratio()).
//
// The byte path operates on bytes, so it normalises by 5+4=9 bytes and
// counts matched bytes — the result differs from the rune path. We pin
// the rune-path value here and assert the byte-path value differs.
func TestRatcliffObershelp_RuneMultiByte(t *testing.T) {
	const eps = 1e-9
	// Rune path: matches Python difflib(autojunk=False).ratio().
	const wantRune = 0.75
	gotRune := fuzzymatch.RatcliffObershelpScoreRunes("café", "cafe")
	if math.Abs(gotRune-wantRune) > eps {
		t.Errorf("RatcliffObershelpScoreRunes(\"café\", \"cafe\") = %.17g; want ≈ %g", gotRune, wantRune)
	}
	// Byte path differs because "café" is 5 bytes (c, a, f, 0xc3, 0xa9)
	// while "cafe" is 4 bytes (c, a, f, e). The 3 matched bytes c/a/f
	// over (5+4) bytes = 6/9 = 0.666… (verified via Python on the
	// equivalent byte-string).
	gotByte := fuzzymatch.RatcliffObershelpScore("café", "cafe")
	if gotByte == gotRune {
		t.Errorf("byte and rune paths unexpectedly identical on multi-byte input: byte=%g rune=%g — rune path must operate on code points, not bytes",
			gotByte, gotRune)
	}
	const wantByte = 6.0 / 9.0
	if math.Abs(gotByte-wantByte) > eps {
		t.Errorf("RatcliffObershelpScore(\"café\", \"cafe\") = %.17g; want ≈ %g (3 matched bytes over 9 total)",
			gotByte, wantByte)
	}
}

// TestRatcliffObershelp_RuneIdentity_ShortCircuit verifies the IN-04
// closure: RatcliffObershelpScoreRunes returns 1.0 on identical inputs
// WITHOUT allocating []rune(a) / []rune(b). The identity short-circuit
// `if a == b { return 1.0 }` must fire BEFORE the []rune allocation.
func TestRatcliffObershelp_RuneIdentity_ShortCircuit(t *testing.T) {
	const s = "café_WIKIMEDIA"
	if got := fuzzymatch.RatcliffObershelpScoreRunes(s, s); got != 1.0 {
		t.Errorf("RatcliffObershelpScoreRunes(%q, %q) = %g; want 1.0", s, s, got)
	}
	allocs := testing.AllocsPerRun(100, func() {
		_ = fuzzymatch.RatcliffObershelpScoreRunes(s, s)
	})
	if allocs > 0 {
		t.Errorf("RatcliffObershelpScoreRunes identity short-circuit: %.1f allocs/op; want 0 (no []rune alloc)", allocs)
	}
}

// TestRatcliffObershelp_CrossValidation asserts agreement between our
// RatcliffObershelp byte-path implementation and the Python
// difflib.SequenceMatcher(autojunk=False).ratio() reference corpus
// committed at testdata/cross-validation/ratcliff-obershelp/vectors.json.
//
// Tolerance: |our_score - difflib_ratio| <= 1e-9 per entry (matches the
// Phase 3 SWG cross-validation epsilon convention).
//
// The corpus is regenerated by `make regen-ratcliff-obershelp-cross-validation`
// (developer-only); CI does NOT require Python at test time. If this test
// fails after a corpus regeneration, EITHER our Ratcliff-Obershelp kernel
// drifted from the Ratcliff & Metzener 1988 / difflib semantics, OR the
// Python version emitted different scores (the python_version from the
// corpus header is included in the failure message for triage), OR autojunk
// was accidentally enabled in the generator (the autojunk_sensitive entry
// is the keystone gate for this).
//
// Per-entry sub-tests via t.Run so individual entry failures are visible
// without truncation. The autojunk_sensitive sub-test is load-bearing per
// VALIDATION.md row 04-04-04 — its failure alone proves the autojunk=False
// contract is broken and forces the algorithm-correctness-reviewer to
// block the PR.
func TestRatcliffObershelp_CrossValidation(t *testing.T) {
	const epsilon = 1e-9
	path := filepath.Join("testdata", "cross-validation", "ratcliff-obershelp", "vectors.json")
	raw, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("TestRatcliffObershelp_CrossValidation: read %s: %v (regenerate with `make regen-ratcliff-obershelp-cross-validation`)", path, err)
	}
	var c roCrossValidationCorpus
	if err := json.Unmarshal(raw, &c); err != nil {
		t.Fatalf("TestRatcliffObershelp_CrossValidation: parse %s: %v", path, err)
	}
	if c.Version != 1 {
		t.Fatalf("TestRatcliffObershelp_CrossValidation: unsupported corpus version %d (want 1)", c.Version)
	}
	if len(c.Entries) == 0 {
		t.Fatalf("TestRatcliffObershelp_CrossValidation: empty corpus")
	}
	for _, e := range c.Entries {
		e := e // local copy for the closure
		t.Run(e.Name, func(t *testing.T) {
			got := fuzzymatch.RatcliffObershelpScore(e.A, e.B)
			delta := math.Abs(got - e.DifflibRatio)
			if delta > epsilon {
				t.Errorf("RatcliffObershelpScore(%q, %q) = %.17g; difflib_ratio = %.17g (delta %.2e, tol %g, python %s)",
					e.A, e.B, got, e.DifflibRatio, delta, epsilon, c.PythonVersion)
			}
		})
	}
}

// TestRatcliffObershelp_CrossValidation_CorpusShape asserts the committed
// corpus contains 15–18 entries spanning all four mandatory CONTEXT.md §1
// categories. Presence is checked by entry-name substring matching:
//
//   - Category 1 (standard edge cases): at least one entry name contains
//     the substring "empty".
//   - Category 2 (Dr. Dobb's 1988 paper examples): at least one entry name
//     equals "wikimedia_wikimania" or "gestalt_paper".
//   - Category 3 (autojunk-sensitive 200+ char keystone): at least one
//     entry name contains "autojunk".
//   - Category 4 (substring / partial / unicode): at least one entry name
//     contains "unicode" or "cafe".
//
// Together with TestRatcliffObershelp_CrossValidation, this shape gate
// prevents a future regression where someone trims the corpus to remove
// the autojunk_sensitive entry (which would silently disarm the
// keystone gate without surfacing a delta).
func TestRatcliffObershelp_CrossValidation_CorpusShape(t *testing.T) {
	path := filepath.Join("testdata", "cross-validation", "ratcliff-obershelp", "vectors.json")
	raw, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("TestRatcliffObershelp_CrossValidation_CorpusShape: read %s: %v", path, err)
	}
	var c roCrossValidationCorpus
	if err := json.Unmarshal(raw, &c); err != nil {
		t.Fatalf("TestRatcliffObershelp_CrossValidation_CorpusShape: parse %s: %v", path, err)
	}
	if n := len(c.Entries); n < 15 || n > 18 {
		t.Fatalf("corpus entry count = %d; want 15..18 (CONTEXT.md §1)", n)
	}
	names := make([]string, len(c.Entries))
	for i, e := range c.Entries {
		names[i] = e.Name
	}
	categories := []struct {
		label string
		ok    func(string) bool
	}{
		{
			label: "Category 1 — standard edge cases (entry name containing \"empty\")",
			ok:    func(n string) bool { return strings.Contains(n, "empty") },
		},
		{
			label: "Category 2 — Dr. Dobb's 1988 paper examples (\"wikimedia_wikimania\" or \"gestalt_paper\")",
			ok:    func(n string) bool { return n == "wikimedia_wikimania" || n == "gestalt_paper" },
		},
		{
			label: "Category 3 — autojunk-sensitive 200+ char keystone (entry name containing \"autojunk\")",
			ok:    func(n string) bool { return strings.Contains(n, "autojunk") },
		},
		{
			label: "Category 4 — substring/partial/unicode (entry name containing \"unicode\" or \"cafe\")",
			ok:    func(n string) bool { return strings.Contains(n, "unicode") || strings.Contains(n, "cafe") },
		},
	}
	for _, cat := range categories {
		found := false
		for _, n := range names {
			if cat.ok(n) {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("missing %s; got entries: %v", cat.label, names)
		}
	}
}
