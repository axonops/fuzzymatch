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

// strcmp95_test.go pins the public-API contract of strcmp95.go: identity,
// both-empty, one-empty, canonical reference vectors from Winkler 1994 +
// Census Bureau strcmp95.c canonical surnames (MARTHA/MARHTA, DWAYNE/DUANE,
// DIXON/DICKSONX), symmetry, the similar-character table invariants
// (PITFALLS §14 / Pitfall 1), the Strcmp95 ≥ JaroWinkler hierarchy property
// on the canonical pairs, the long-string adjustment trigger conditions
// (Pitfall 5), and the runtime allocation gate.
//
// Stdlib `testing` only — no testify in root tests, per
// .claude/skills/go-coding-standards.

package fuzzymatch_test

import (
	"testing"

	"github.com/axonops/fuzzymatch"
)

// TestStrcmp95_BothEmpty asserts Strcmp95Score("", "") = 1.0 — the
// canonical both-empty identity convention.
func TestStrcmp95_BothEmpty(t *testing.T) {
	if got := fuzzymatch.Strcmp95Score("", ""); got != 1.0 {
		t.Errorf("Strcmp95Score(\"\", \"\") = %g; want 1.0", got)
	}
}

// TestStrcmp95_OneEmpty asserts Strcmp95Score(x, "") == 0.0 and
// Strcmp95Score("", x) == 0.0 — the one-empty silent-zero convention.
func TestStrcmp95_OneEmpty(t *testing.T) {
	tests := []struct {
		a, b string
	}{
		{"", "abc"},
		{"abc", ""},
		{"", "x"},
		{"x", ""},
		{"", "MARTHA"},
		{"MARTHA", ""},
	}
	for _, tt := range tests {
		t.Run(tt.a+"_"+tt.b, func(t *testing.T) {
			if got := fuzzymatch.Strcmp95Score(tt.a, tt.b); got != 0.0 {
				t.Errorf("Strcmp95Score(%q, %q) = %g; want 0.0", tt.a, tt.b, got)
			}
		})
	}
}

// TestStrcmp95_Identical asserts Strcmp95Score(x, x) == 1.0 for non-empty
// x — the identity short-circuit at function entry.
func TestStrcmp95_Identical(t *testing.T) {
	tests := []string{
		"a",
		"abc",
		"MARTHA",
		"http_request",
		"DWAYNE",
		"the_quick_brown_fox",
	}
	for _, s := range tests {
		t.Run(s, func(t *testing.T) {
			if got := fuzzymatch.Strcmp95Score(s, s); got != 1.0 {
				t.Errorf("Strcmp95Score(%q, %q) = %g; want 1.0", s, s, got)
			}
		})
	}
}

// TestStrcmp95_ReferenceVectors_CensusBureau pins canonical Winkler 1994 /
// Census Bureau strcmp95.c surnames. The expected values come from our
// implementation's deterministic output — the strcmp95.c canonical algorithm
// produces scores that DIFFER from JaroWinkler precisely where the similar-
// character table fires (DWAYNE/DUANE, DIXON/DICKSONX) and equals JaroWinkler
// where no similar pairs are present (MARTHA/MARHTA, modulo the long-string
// adjustment).
//
// RESEARCH.md Pitfall 1 warning sign #2: Strcmp95Score == JaroWinklerScore on
// EVERY input means the similar-character table is NOT firing. The DWAYNE
// pair has W/U which is in the table (entry {'W', 'U', 0.3}); we assert
// Strcmp95Score strictly exceeds JaroWinklerScore for that pair.
//
// Tolerance: 1e-3 — wide enough to absorb algorithm-implementation variance
// across published strcmp95.c forks (richmilne, OpenRefine, Census Bureau
// originals) while tight enough to catch silent score drift.
func TestStrcmp95_ReferenceVectors_CensusBureau(t *testing.T) {
	tests := []struct {
		a, b string
		want float64
	}{
		// MARTHA/MARHTA: no similar pair fires (T/H, H/T not similar).
		// The long-string adjustment fires (min=6>4, m=6, prefix=3,
		// 2·m=12>=6+3=9) and additively contributes a small lift above
		// JaroWinkler's 0.9611.
		{"MARTHA", "MARHTA", 0.9676},
		// DWAYNE/DUANE: D=D matched; W/U is in the similar table (entry
		// {'W','U',0.3}). The similar-character credit fires AND the
		// long-string adjustment fires.
		{"DWAYNE", "DUANE", 0.8925},
		// DIXON/DICKSONX: D=D, I=I matched; the unmatched (X / C, K, S,
		// O, N, X) layout exercises the similar pass. C/K is in the table
		// (entry {'C','K',0.3}).
		{"DIXON", "DICKSONX", 0.8517},
	}
	const tol = 1e-3
	for _, tt := range tests {
		t.Run(tt.a+"_"+tt.b, func(t *testing.T) {
			got := fuzzymatch.Strcmp95Score(tt.a, tt.b)
			if d := got - tt.want; d > tol || d < -tol {
				t.Errorf("Strcmp95Score(%q, %q) = %g; want %g (tol %g)",
					tt.a, tt.b, got, tt.want, tol)
			}
		})
	}
}

// TestStrcmp95_SimilarCharTableFires is the load-bearing regression test for
// RESEARCH.md Pitfall 1 warning sign #2: Strcmp95Score must DIFFER from
// JaroWinklerScore on inputs where the similar-character table fires. The
// DWAYNE/DUANE pair has the canonical W/U similar pair from Winkler 1994 §3.
//
// If this test passes but every input produces Strcmp95 == JaroWinkler, the
// similar-character table is broken (missing pairs, wrong lookup, etc.).
func TestStrcmp95_SimilarCharTableFires(t *testing.T) {
	tests := []struct {
		a, b string
	}{
		// W/U pair (Winkler 1994 §3).
		{"DWAYNE", "DUANE"},
		// C/K pair (Winkler 1994 §3).
		{"DIXON", "DICKSONX"},
	}
	for _, tt := range tests {
		t.Run(tt.a+"_"+tt.b, func(t *testing.T) {
			s := fuzzymatch.Strcmp95Score(tt.a, tt.b)
			jw := fuzzymatch.JaroWinklerScore(tt.a, tt.b)
			if s <= jw {
				t.Errorf("Strcmp95Score(%q, %q) = %g; expected to STRICTLY EXCEED JaroWinklerScore = %g (similar-character table should fire) — RESEARCH.md Pitfall 1 warning sign #2",
					tt.a, tt.b, s, jw)
			}
		})
	}
}

// TestStrcmp95_LongStringAdjustment_Triggers pins the load-bearing
// long-string adjustment trigger (RESEARCH.md Pitfall 5):
//
//   - HAMINGTON/HAMMINGTON: min=9>4, m≈8, prefix=3, conditions all hold →
//     adjustment fires; Strcmp95 > JaroWinkler.
//   - AB/AC: min=2, fails min>4 condition → adjustment does NOT fire;
//     Strcmp95 == JaroWinkler (both ≈ 0.6667).
//
// The test asserts BOTH the firing case and the not-firing case to lock the
// gate condition itself, not just one side of the boundary.
func TestStrcmp95_LongStringAdjustment_Triggers(t *testing.T) {
	t.Run("HAMINGTON_HAMMINGTON_adjustment_fires", func(t *testing.T) {
		s := fuzzymatch.Strcmp95Score("HAMINGTON", "HAMMINGTON")
		jw := fuzzymatch.JaroWinklerScore("HAMINGTON", "HAMMINGTON")
		if s <= jw {
			t.Errorf("Strcmp95Score(HAMINGTON, HAMMINGTON) = %g; expected to exceed JaroWinklerScore = %g (long-string adjustment should fire)", s, jw)
		}
	})
	t.Run("AB_AC_adjustment_does_not_fire", func(t *testing.T) {
		// min(2,2)=2, fails min>4 condition. No similar pair (B/C not in table).
		// Strcmp95 should equal JaroWinkler exactly.
		s := fuzzymatch.Strcmp95Score("AB", "AC")
		jw := fuzzymatch.JaroWinklerScore("AB", "AC")
		if s != jw {
			t.Errorf("Strcmp95Score(AB, AC) = %g; expected exact equality with JaroWinklerScore = %g (no adjustment should fire on short input)", s, jw)
		}
	})
}

// TestStrcmp95_AtLeastJaroWinkler_OnReferenceVectors is a hand-pinned spot
// check of the Strcmp95 ≥ JaroWinkler hierarchy invariant on the canonical
// reference set. The full testing/quick property version lives in
// props_test.go as TestProp_Strcmp95Score_AtLeastJaroWinkler.
func TestStrcmp95_AtLeastJaroWinkler_OnReferenceVectors(t *testing.T) {
	pairs := []struct{ a, b string }{
		{"MARTHA", "MARHTA"},
		{"DWAYNE", "DUANE"},
		{"DIXON", "DICKSONX"},
		{"HAMINGTON", "HAMMINGTON"},
		{"AB", "AC"},
		{"abc", "abc"},
		{"", ""},
		{"abcXYZabc", "abc"},
	}
	for _, p := range pairs {
		t.Run(p.a+"_"+p.b, func(t *testing.T) {
			s := fuzzymatch.Strcmp95Score(p.a, p.b)
			jw := fuzzymatch.JaroWinklerScore(p.a, p.b)
			if s+1e-12 < jw {
				t.Errorf("Strcmp95Score(%q, %q) = %g < JaroWinklerScore = %g — hierarchy invariant violated", p.a, p.b, s, jw)
			}
		})
	}
}

// TestStrcmp95_Symmetric_OnReferenceVectors pins the byte-path symmetry
// invariant on the canonical reference set. The full testing/quick property
// version lives in props_test.go as TestProp_Strcmp95Score_Symmetric.
func TestStrcmp95_Symmetric_OnReferenceVectors(t *testing.T) {
	pairs := []struct{ a, b string }{
		{"MARTHA", "MARHTA"},
		{"DWAYNE", "DUANE"},
		{"DIXON", "DICKSONX"},
		{"HAMINGTON", "HAMMINGTON"},
		{"kitten", "sitting"},
		{"abc", "xyz"},
	}
	for _, p := range pairs {
		t.Run(p.a+"_"+p.b, func(t *testing.T) {
			ab := fuzzymatch.Strcmp95Score(p.a, p.b)
			ba := fuzzymatch.Strcmp95Score(p.b, p.a)
			if ab != ba {
				t.Errorf("Strcmp95Score symmetric failed: (%q,%q)=%g != (%q,%q)=%g",
					p.a, p.b, ab, p.b, p.a, ba)
			}
		})
	}
}

// TestStrcmp95_TableInvariants asserts the similar-character table is well-
// formed: exactly 36 entries, every entry has similarity 0.3 (the canonical
// Winkler 1994 weight), no duplicate pair (counting (a,b) and (b,a) as the
// same pair).
//
// This is the load-bearing regression test for RESEARCH.md Pitfall 1
// (transcription typos): if any pair is missing, swapped, duplicated, or
// has the wrong similarity value, this test catches it before the score
// drift propagates into downstream consumers.
func TestStrcmp95_TableInvariants(t *testing.T) {
	const wantLen = 36
	const wantSim = 0.3
	if got := fuzzymatch.Strcmp95SimilarCharsLenForTest; got != wantLen {
		t.Errorf("strcmp95SimilarChars has %d entries; want %d (Winkler 1994 TR-2 §3)", got, wantLen)
	}
	if got := fuzzymatch.Strcmp95SimilarCreditForTest; got != wantSim {
		t.Errorf("strcmp95SimilarCredit = %g; want %g (Winkler 1994 canonical weight)", got, wantSim)
	}

	// Pair-canonicalisation key: (min(a,b), max(a,b)) to treat (a,b) and
	// (b,a) as the same pair. Duplicate detection is via a map of these
	// canonical keys.
	type pairKey struct {
		lo, hi byte
	}
	seen := map[pairKey]int{}
	for i := 0; i < fuzzymatch.Strcmp95SimilarCharsLenForTest; i++ {
		a, b, sim := fuzzymatch.Strcmp95SimilarCharsEntryForTest(i)
		if sim != wantSim {
			t.Errorf("entry %d (%c, %c) has sim = %g; want %g", i, a, b, sim, wantSim)
		}
		lo, hi := a, b
		if hi < lo {
			lo, hi = hi, lo
		}
		key := pairKey{lo: lo, hi: hi}
		if prev, ok := seen[key]; ok {
			t.Errorf("duplicate pair (%c, %c) at indices %d and %d", a, b, prev, i)
		}
		seen[key] = i
	}
}

// TestStrcmp95Score_ZeroAllocs_ASCII_Short pins the 0-alloc budget for the
// ASCII fast path at runtime (not just bench time). The MARTHA/MARHTA pair
// is 6 and 6 bytes — well within maxJaroStackLen = 256, so the match-flag
// arrays AND the similar-pair consumption arena stack-allocate.
func TestStrcmp95Score_ZeroAllocs_ASCII_Short(t *testing.T) {
	// Quick warmup to let escape analysis settle (first-call init artefacts).
	_ = fuzzymatch.Strcmp95Score("MARTHA", "MARHTA")

	allocs := testing.AllocsPerRun(100, func() {
		_ = fuzzymatch.Strcmp95Score("MARTHA", "MARHTA")
	})
	if allocs > 0 {
		t.Errorf("Strcmp95Score ASCII short: %.1f allocs/op; want 0 (stack buffers not escaping?)", allocs)
	}
}

// TestStrcmp95Score_ZeroAllocs_ASCII_Medium pins the 0-alloc budget at
// ~50 bytes — still within maxJaroStackLen = 256.
func TestStrcmp95Score_ZeroAllocs_ASCII_Medium(t *testing.T) {
	const a50 = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWX"
	const b50 = "bcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXY"
	_ = fuzzymatch.Strcmp95Score(a50, b50)

	allocs := testing.AllocsPerRun(100, func() {
		_ = fuzzymatch.Strcmp95Score(a50, b50)
	})
	if allocs > 0 {
		t.Errorf("Strcmp95Score ASCII medium: %.1f allocs/op; want 0 (stack buffers not escaping?)", allocs)
	}
}

// TestStrcmp95_LowerCaseEqualsUpperCase asserts the case-folding behaviour of
// the similar-character lookup: lower-case ASCII letters fold to upper-case
// before consulting the Winkler 1994 table, so 'dwayne'/'duane' produces the
// same Strcmp95 result as 'DWAYNE'/'DUANE'. This is the documented
// case-insensitive similar-character pass behaviour from the Census Bureau
// strcmp95.c reference.
func TestStrcmp95_LowerCaseEqualsUpperCase(t *testing.T) {
	upper := fuzzymatch.Strcmp95Score("DWAYNE", "DUANE")
	lower := fuzzymatch.Strcmp95Score("dwayne", "duane")
	if upper != lower {
		t.Errorf("Strcmp95Score case-folding mismatch: upper=%g lower=%g (similar table should fold ASCII case)", upper, lower)
	}
}
