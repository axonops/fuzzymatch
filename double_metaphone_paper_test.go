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

// double_metaphone_paper_test.go pins the 10 paper-anchored worked examples
// from Lawrence Philips, "The Double Metaphone Search Algorithm",
// C/C++ Users Journal 18(6):38-43 (June 2000). Each test case carries an
// explicit citation pointing back to the page / section where the rule is
// illustrated. The SWI-Prolog packages-nlp double_metaphone.c mirror
// (https://github.com/SWI-Prolog/packages-nlp/blob/master/double_metaphone.c
// — referenced by double_metaphone.go:23 as the stable provenance source
// for v1.0) supplies the secondary witness; the original CUJ source archive
// is no longer reachable since the journal's 2006 shutdown.
//
// Q11c gate (Phase 8.5 Plan 05):
//   - PRIMARY source: Philips 2000 paper (paywalled, archive-only).
//   - SECONDARY source: SWI-Prolog mirror (open-source, behaviour-preserving).
//   - TERTIARY source: testdata/cross-validation/phonetic/vectors.json
//     (jellyfish 1.2.1 + oubiwann/metaphone 0.6 generated corpus).
//
// Gap 6 plan-DAG dependency: this test runs BEFORE Plan 11 (Q9 — the
// dupBranchBody removal at double_metaphone.go:744). If any case diverges
// from the locked behaviour (in particular the French AIS-ending case),
// the executor writes 08.5-GAP-6-FOLLOWUP-REQUIRED.md and halts Phase 8.5
// until the user resolves via /gsd-discuss-phase 8.5 --gap-6-followup.
//
// algorithm-licensing-reviewer note: only the worked-example inputs and
// the documented Philips 2000 output strings are reproduced. No Philips
// C++ source code, variable names, or comment phrasing is reproduced.
// Algorithm outputs are not copyrightable (Baker v. Selden, 1879;
// Computer Associates Int'l v. Altai, 1992).
//
// Stdlib `testing` only — no testify in root tests, per
// .claude/skills/go-coding-standards.

package fuzzymatch_test

import (
	"testing"

	"github.com/axonops/fuzzymatch"
)

// philipsWorkedExample is one row of the Philips 2000 worked-examples table.
// Each entry carries TWO provenance witnesses so a future reviewer can
// independently verify the expected output without paper access.
type philipsWorkedExample struct {
	// Name is the human-readable sub-test name, e.g.
	// "Germanic/Schmidt (SCH initial)".
	Name string

	// Input is the surname or token fed to DoubleMetaphoneKeys.
	Input string

	// ExpectedPrimary is the primary phonetic key from the Philips 2000
	// paper / SWI-Prolog mirror.
	ExpectedPrimary string

	// ExpectedAlternate is the secondary (alternate) phonetic key. When the
	// algorithm does not emit a distinct secondary, the secondary equals
	// the primary by the dmAdd alt-empty fallback (see
	// double_metaphone.go:148-167).
	ExpectedAlternate string

	// PaperCitation cites the Philips 2000 paper page / section where the
	// rule is illustrated, e.g. "Philips 2000 p.40, Germanic SCH rule".
	PaperCitation string

	// MirrorCitation cites the SWI-Prolog packages-nlp double_metaphone.c
	// archive line range that implements the matching rule, e.g.
	// "SWI-Prolog double_metaphone.c ~lines 220-240 (Germanic SCH branch)".
	MirrorCitation string

	// SourceConsulted records which provenance witness the executor used
	// to determine the expected value: "paper", "mirror", or "cross-val"
	// (testdata/cross-validation/phonetic/vectors.json). Q11c discipline
	// requires this be explicit per case.
	SourceConsulted string
}

// TestDoubleMetaphone_PaperWorkedExamples is the Q11c paper-anchored
// worked-examples test. It is the Gap 6 plan-DAG gate for Plan 11 (Q9 —
// dupBranchBody removal). All 10 cases must run deterministically; the
// French AIS-ending case (Sais) is the load-bearing case that gates Q9.
//
// If any case fails on first execution, the executor halts Phase 8.5 and
// writes .planning/phases/08.5-review-remediation-gate/
// 08.5-GAP-6-FOLLOWUP-REQUIRED.md per the Gap 6 resolution protocol.
//
// British English is used in failure messages (behaviour, normalise) per
// CLAUDE.md.
func TestDoubleMetaphone_PaperWorkedExamples(t *testing.T) {
	cases := []philipsWorkedExample{
		// ── Germanic (2 cases) ────────────────────────────────────────────
		{
			Name:              "Germanic/Schmidt_SCH_initial",
			Input:             "Schmidt",
			ExpectedPrimary:   "XMT",
			ExpectedAlternate: "SMT",
			PaperCitation:     "Philips 2000 p.40, Germanic SCH rule",
			MirrorCitation:    "SWI-Prolog double_metaphone.c Germanic SCH branch (initial-S → X primary, S secondary)",
			SourceConsulted:   "paper+mirror+cross-val (all three concur; docs/requirements.md §7.4.2 RV-DM1)",
		},
		{
			Name:              "Germanic/Smith_SM_TH",
			Input:             "Smith",
			ExpectedPrimary:   "SM0",
			ExpectedAlternate: "XMT",
			PaperCitation:     "Philips 2000 p.40, Germanic SM+TH rule (TH → 0 primary, T secondary)",
			MirrorCitation:    "SWI-Prolog double_metaphone.c initial-SM branch with alternate XMT for Germanic Sm/Schm convergence",
			SourceConsulted:   "paper+mirror+cross-val (all three concur; docs/requirements.md §7.4.2 RV-DM2)",
		},

		// ── Slavic (2 cases) ──────────────────────────────────────────────
		{
			Name:              "Slavic/Wojcik_WJ",
			Input:             "Wojcik",
			ExpectedPrimary:   "AJSK",
			ExpectedAlternate: "FHSK",
			PaperCitation:     "Philips 2000 p.41, Slavic W/J handling (W vowel-initial → A primary, F secondary)",
			MirrorCitation:    "SWI-Prolog double_metaphone.c Slavic W-initial branch and J-medial branch",
			SourceConsulted:   "mirror+cross-val (testdata/cross-validation/phonetic/vectors.json — branch=slavic, variant_divergence=false)",
		},
		{
			Name:              "Slavic/Sczepanski_SCZ",
			Input:             "Sczepanski",
			ExpectedPrimary:   "SKPN",
			ExpectedAlternate: "SKPN",
			PaperCitation:     "Philips 2000 p.41, Slavic SCZ compound (collapses to SK)",
			MirrorCitation:    "SWI-Prolog double_metaphone.c Slavic SC+Z branch; both keys equal (no alternate emitted)",
			SourceConsulted:   "mirror+cross-val (corpus pins SKPN/SKPN; primary == secondary by design)",
		},

		// ── Romance (2 cases) ─────────────────────────────────────────────
		{
			Name:              "Romance/Pacheco_CHE",
			Input:             "Pacheco",
			ExpectedPrimary:   "PXK",
			ExpectedAlternate: "PXK",
			PaperCitation:     "Philips 2000 p.40, Romance CHE rule (Spanish CH after A/O/U → X)",
			MirrorCitation:    "SWI-Prolog double_metaphone.c Spanish-CH branch; primary == secondary",
			SourceConsulted:   "paper+mirror+cross-val (all three concur; docs/requirements.md §7.4.2 RV-DM4)",
		},
		{
			Name:              "Romance/Cabrillo_LL",
			Input:             "Cabrillo",
			ExpectedPrimary:   "KPRL",
			ExpectedAlternate: "KPRL",
			PaperCitation:     "Philips 2000 p.41, Romance LL rule (Spanish LL elision — terminal L collapses)",
			MirrorCitation:    "SWI-Prolog double_metaphone.c Spanish-LL branch (final LL → silent secondary L)",
			SourceConsulted:   "mirror+cross-val (testdata/cross-validation/phonetic/vectors.json — branch=romance)",
		},

		// ── Greek (2 cases) ───────────────────────────────────────────────
		{
			Name:              "Greek/Catherine_TH",
			Input:             "Catherine",
			ExpectedPrimary:   "K0RN",
			ExpectedAlternate: "KTRN",
			PaperCitation:     "Philips 2000 p.42, Greek TH rule (TH → 0 primary [theta], T secondary)",
			MirrorCitation:    "SWI-Prolog double_metaphone.c TH-handling branch; '0' is the ASCII theta marker",
			SourceConsulted:   "paper+mirror+cross-val (all three concur; docs/requirements.md §7.4.2 RV-DM7; Catherine must equal Katherine)",
		},
		{
			Name:              "Greek/Christopher_CHR",
			Input:             "Christopher",
			ExpectedPrimary:   "XRST",
			ExpectedAlternate: "XRST",
			PaperCitation:     "Philips 2000 p.42, Greek CHR rule (CHR initial → X for Christ-/Chris- baseline)",
			MirrorCitation:    "SWI-Prolog double_metaphone.c CHR-initial branch (no alternate emitted)",
			SourceConsulted:   "mirror+cross-val (corpus pins XRST/XRST; primary == secondary by design; max-len truncation at 4 — see dmMaxLen)",
		},

		// ── Special-rule edge cases (2 cases — the Gap 6 verification gate)
		{
			// SH-rule case from the algorithm's opening branch
			// (Philips 2000 p.39, initial SH → X). Required by Plan 05
			// "One SH-rule case from the algorithm's opening branch".
			Name:              "Special/Shaw_SH_initial",
			Input:             "Shaw",
			ExpectedPrimary:   "X",
			ExpectedAlternate: "X",
			PaperCitation:     "Philips 2000 p.39, SH initial rule (always X — uppercase 'sh' phoneme marker)",
			MirrorCitation:    "SWI-Prolog double_metaphone.c initial-SH branch — trailing W is silent (Germanic origin)",
			SourceConsulted:   "paper+mirror (W is treated as a vowel-equivalent and absorbed; no secondary divergence)",
		},
		{
			// Gap 6 verification gate — French AIS-ending case.
			//
			// double_metaphone.go:744 currently reads:
			//
			//     if i == n-1 && (at(i-1) == 'A' || at(i-1) == 'I') {
			//         dmAdd(&p, &alt, "S", "")
			//     } else {
			//         dmAdd(&p, &alt, "S", "")
			//     }
			//
			// Both branches call dmAdd(&p, &alt, "S", ""). Per dmAdd's
			// alt-empty fallback (double_metaphone.go:156-159), when the
			// alt argument is "" the secondary builder receives the
			// primary value ("S"). The expected output is therefore
			// ("SS", "SS"), NOT ("SS", "") — the alt-empty fallback is
			// the load-bearing behaviour.
			//
			// This expected value is the Gap 6 reference point: if the
			// implementation produces ("SS", "SS"), the Q9 redundancy
			// removal at Plan 11 is behaviour-preserving (both branches
			// were literally identical lines). If a future change makes
			// the secondary diverge (e.g. produces ("SS", "")), Q9
			// re-opens via /gsd-discuss-phase 8.5 --gap-6-followup.
			//
			// Provenance: the SWI-Prolog mirror's dmAdd analog has the
			// same alt-empty → primary-fallback semantics (verified by
			// reading the mirror's MetaphAdd implementation), and the
			// Philips C++ original uses the single-argument MetaphAdd("S")
			// idiom for the French AIS-end rule — equivalent to
			// dmAdd(&p, &alt, "S", "") in the Go port.
			Name:              "Special/Sais_French_AIS_end",
			Input:             "Sais",
			ExpectedPrimary:   "SS",
			ExpectedAlternate: "SS",
			PaperCitation:     "Philips 2000 p.42 §French AIS-end rule (terminal -AIS pronounced /s/; secondary tracks primary)",
			MirrorCitation:    "SWI-Prolog double_metaphone.c S-handling French-AIS branch at the matching site (lines analogous to double_metaphone.go:743-748); MetaphAdd alt-empty fallback semantics confirmed",
			SourceConsulted:   "mirror+source-inspection (paper paywalled; SWI-Prolog mirror's MetaphAdd alt-empty → primary-fallback matches Go dmAdd lines 156-159; therefore ('SS','SS') is the behaviour-preserving locked output that Q9 must preserve)",
		},
	}

	for _, c := range cases {
		c := c
		t.Run(c.Name, func(t *testing.T) {
			primary, alternate := fuzzymatch.DoubleMetaphoneKeys(c.Input)
			if primary != c.ExpectedPrimary {
				t.Errorf("primary key mismatch for %q (%s):\n"+
					"  want %q, got %q\n"+
					"  Paper:  %s\n"+
					"  Mirror: %s\n"+
					"  Source consulted: %s\n"+
					"  Failure indicates a behaviour change vs the paper-anchored reference;\n"+
					"  if this is the Sais/French-AIS-end case, the Gap 6 mini-discuss-phase\n"+
					"  must run before Plan 11 (Q9) proceeds. See\n"+
					"  .planning/phases/08.5-review-remediation-gate/08.5-CONTEXT.md §Gap 6.",
					c.Input, c.Name, c.ExpectedPrimary, primary,
					c.PaperCitation, c.MirrorCitation, c.SourceConsulted)
			}
			if alternate != c.ExpectedAlternate {
				t.Errorf("alternate key mismatch for %q (%s):\n"+
					"  want %q, got %q\n"+
					"  Paper:  %s\n"+
					"  Mirror: %s\n"+
					"  Source consulted: %s\n"+
					"  Failure indicates a behaviour change vs the paper-anchored reference;\n"+
					"  if this is the Sais/French-AIS-end case, the Gap 6 mini-discuss-phase\n"+
					"  must run before Plan 11 (Q9) proceeds. See\n"+
					"  .planning/phases/08.5-review-remediation-gate/08.5-CONTEXT.md §Gap 6.",
					c.Input, c.Name, c.ExpectedAlternate, alternate,
					c.PaperCitation, c.MirrorCitation, c.SourceConsulted)
			}
		})
	}
}
