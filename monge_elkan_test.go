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

// monge_elkan_test.go pins the public-API contract of monge_elkan.go:
// the six hand-derived reference vectors RV-ME1..RV-ME6 per CONTEXT.md
// §1b LOCKED (each carries its per-token-max derivation in the test
// comment so a reviewer can re-derive the expected value in under a
// minute), identity, both-empty (STANDARD catalogue convention),
// one-empty, the symmetric variant pin, the KEYSTONE asymmetry gate
// (RV-ME6 — load-bearing direction-sensitivity regression detector),
// dispatch registration with the LOCKED CONTEXT.md §4 defaults
// (symmetric variant + AlgoJaroWinkler default inner +
// DefaultNormalisationOptions), and the exhaustive panic test walking
// all 9 NON-permitted AlgoIDs per RESEARCH.md Pitfall 4.
//
// Stdlib `testing` only — no testify in root tests, per
// .claude/skills/go-coding-standards.

package fuzzymatch_test

import (
	"math"
	"strings"
	"testing"

	"github.com/axonops/fuzzymatch"
)

// mongeElkanEpsilon is the float-comparison tolerance for irrational
// expected values (e.g. 0.9125 from the JaroWinkler-driven max-mean).
// Phase 2/3/4/5/6 convention is 1e-9; the Monge-Elkan asymmetric formula
// is a sum of float64 max operations and a single division so the
// derivation is exact at IEEE-754 precision but values from the inner
// metric (JaroWinkler) carry their own rounding. 1e-9 is well above the
// observed accumulated error on these inputs.
const mongeElkanEpsilon = 1e-9

// TestMongeElkanScore covers the six hand-derived reference vectors
// RV-ME1..RV-ME6 per CONTEXT.md §1b LOCKED, plus the conventional
// short-circuit cases (identity, both-empty, one-empty). Each row's
// derivation is reproduced in the test comment so a reviewer can
// re-derive the expected value from Monge & Elkan 1996 §3 in under a
// minute. RV-ME6 (the asymmetry KEYSTONE) is the load-bearing
// regression gate for direction-sensitivity.
func TestMongeElkanScore(t *testing.T) {
	tests := []struct {
		name       string
		a, b       string
		inner      fuzzymatch.AlgoID
		want       float64
		exact      bool // exact equality (rational) vs. epsilon (irrational)
		derivation string
	}{
		{
			// RV-ME1 — canonical asymmetric example with JaroWinkler inner.
			// tokens(A) = ["user","create"]; tokens(B) = ["usr","creating"]
			// JW(user, usr)      = 0.9333… (3-of-3 prefix bonus)
			// JW(user, creating) = 0.4167 (low overlap)
			// JW(create, usr)    = 0.5
			// JW(create, creating) = 0.8917 (5-of-5 prefix; "creat" common)
			// max_inner(user,  *) = 0.9333…
			// max_inner(create,*) = 0.8917
			// ME(A, B) = (0.9333… + 0.8917) / 2 = 0.91249999…
			name:       "RV-ME1_user_create_vs_usr_creating_jw",
			a:          "user create",
			b:          "usr creating",
			inner:      fuzzymatch.AlgoJaroWinkler,
			want:       0.9125,
			exact:      false,
			derivation: "tokens(A)=[user,create]; tokens(B)=[usr,creating]; max(JW(user,usr)=0.9333, JW(user,creating)=0.4167)=0.9333; max(JW(create,usr)=0.5, JW(create,creating)=0.8917)=0.8917; ME=(0.9333+0.8917)/2=0.9125",
		},
		{
			// RV-ME2 — identity short-circuit (a == b, returns 1.0
			// without invoking Tokenise or the inner metric).
			name:       "RV-ME2_identity_alpha",
			a:          "alpha beta",
			b:          "alpha beta",
			inner:      fuzzymatch.AlgoJaroWinkler,
			want:       1.0,
			exact:      true,
			derivation: "a == b identity short-circuit; ME(x, x, sim) = 1.0 for any sim",
		},
		{
			// RV-ME3 — disjoint token sets with JaroWinkler inner.
			// tokens(A) = ["alpha","beta"]; tokens(B) = ["gamma","delta"]
			// JW(alpha, gamma) = 0.6;        JW(alpha, delta) = 0.6
			// JW(beta,  gamma) = 0.4833;     JW(beta,  delta) = 0.7833
			// max_inner(alpha, *) = 0.6
			// max_inner(beta,  *) = 0.7833
			// ME(A, B) = (0.6 + 0.7833) / 2 = 0.69166…
			name:       "RV-ME3_disjoint_greek_jw",
			a:          "alpha beta",
			b:          "gamma delta",
			inner:      fuzzymatch.AlgoJaroWinkler,
			want:       0.6916666666666667,
			exact:      false,
			derivation: "tokens(A)=[alpha,beta]; tokens(B)=[gamma,delta]; max(JW(alpha,gamma)=0.6, JW(alpha,delta)=0.6)=0.6; max(JW(beta,gamma)=0.4833, JW(beta,delta)=0.7833)=0.7833; ME=(0.6+0.7833)/2=0.6917",
		},
		{
			// RV-ME4 — partial-overlap, token-count asymmetry (|tA| < |tB|).
			// tokens(A) = ["alpha"]; tokens(B) = ["alpha","beta","gamma"]
			// max_inner(alpha, *) = max(Lev(alpha,alpha)=1, Lev(alpha,beta)=0.2,
			//                           Lev(alpha,gamma)=0.2) = 1.0
			// ME(A, B) = 1.0 / 1 = 1.0
			//
			// This row forces |tA| ≠ |tB| so direction matters; paired
			// with RV-ME6 below the asymmetry is visible.
			name:       "RV-ME4_subset_alpha_levenshtein",
			a:          "alpha",
			b:          "alpha beta gamma",
			inner:      fuzzymatch.AlgoLevenshtein,
			want:       1.0,
			exact:      true,
			derivation: "tokens(A)=[alpha]; tokens(B)=[alpha,beta,gamma]; max(Lev(alpha,alpha)=1, Lev(alpha,beta)=0.2, Lev(alpha,gamma)=0.2)=1.0; ME=1.0/1=1.0",
		},
		{
			// RV-ME5 — Unicode tokens with Levenshtein inner. Tokenise
			// lowercases via Lowercase=true; the single-token input
			// stays single-token after tokenisation.
			// tokens(A) = ["café"]; tokens(B) = ["cafe"]
			// Lev(café, cafe) = 0.6 (1 substitution in 5 bytes — the
			//                       é→e replaces 2 bytes; max len = 5,
			//                       distance = 2; score = 1 - 2/5 = 0.6)
			// ME(A, B) = 0.6 / 1 = 0.6
			name:       "RV-ME5_unicode_café_vs_cafe_levenshtein",
			a:          "café",
			b:          "cafe",
			inner:      fuzzymatch.AlgoLevenshtein,
			want:       0.6,
			exact:      false,
			derivation: "tokens(A)=[café]; tokens(B)=[cafe]; Lev(café, cafe)=0.6 (byte path, distance=2, maxLen=5); ME=0.6/1=0.6",
		},
		{
			// RV-ME6 — KEYSTONE asymmetry gate (RV-ME-asym fixture).
			// Pair with RV-ME4 above: same (a, b) tokens, swapped
			// direction. The expected scores differ:
			//
			// RV-ME4: MongeElkanScore("alpha", "alpha beta gamma", Lev) = 1.0
			// RV-ME6: MongeElkanScore("alpha beta gamma", "alpha", Lev) = ?
			//
			// tokens(A) = ["alpha","beta","gamma"]; tokens(B) = ["alpha"]
			// max_inner(alpha, *) = Lev(alpha, alpha) = 1.0
			// max_inner(beta,  *) = Lev(beta,  alpha) = 0.2
			// max_inner(gamma, *) = Lev(gamma, alpha) = 0.2
			// ME(B, A) = (1.0 + 0.2 + 0.2) / 3 = 0.466666…
			//
			// 1.0 ≠ 0.4666 — the input swap with the same inner
			// produces a direction-sensitive score. This is the
			// load-bearing regression gate; without it a silent
			// direction-swap inside the implementation would still
			// pass RangeBounds + Identity invariants.
			name:       "RV-ME6_asymmetry_keystone_levenshtein",
			a:          "alpha beta gamma",
			b:          "alpha",
			inner:      fuzzymatch.AlgoLevenshtein,
			want:       0.4666666666666666,
			exact:      false,
			derivation: "tokens(A)=[alpha,beta,gamma]; tokens(B)=[alpha]; max(Lev(alpha,alpha)=1)=1.0; max(Lev(beta,alpha)=0.2)=0.2; max(Lev(gamma,alpha)=0.2)=0.2; ME=(1.0+0.2+0.2)/3=0.4666… — RV-ME-asym keystone, RV-ME4's mirror",
		},
		{
			// Conventional: both-empty STANDARD (mirrors TokenSortRatio /
			// TokenJaccard — distinct from TokenSetRatio's RapidFuzz #110
			// deviation). Covered by a == b identity short-circuit.
			name:       "both_empty_standard",
			a:          "",
			b:          "",
			inner:      fuzzymatch.AlgoJaroWinkler,
			want:       1.0,
			exact:      true,
			derivation: "both inputs empty; a == b identity short-circuit fires",
		},
		{
			// Conventional: one-empty.
			name:       "one_empty_a",
			a:          "",
			b:          "hello world",
			inner:      fuzzymatch.AlgoJaroWinkler,
			want:       0.0,
			exact:      true,
			derivation: "tokens(A)=[]; tokens(B)=[hello,world]; one-Tokenised-empty guard fires",
		},
		{
			// Conventional: one-empty (other direction).
			name:       "one_empty_b",
			a:          "hello world",
			b:          "",
			inner:      fuzzymatch.AlgoJaroWinkler,
			want:       0.0,
			exact:      true,
			derivation: "tokens(A)=[hello,world]; tokens(B)=[]; one-Tokenised-empty guard fires",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := fuzzymatch.MongeElkanScore(tt.a, tt.b, tt.inner, fuzzymatch.DefaultNormalisationOptions())
			if tt.exact {
				if got != tt.want {
					t.Errorf("MongeElkanScore(%q, %q, %s) = %.17g; want %.17g (exact)\nderivation: %s",
						tt.a, tt.b, tt.inner, got, tt.want, tt.derivation)
				}
			} else {
				if math.Abs(got-tt.want) > mongeElkanEpsilon {
					t.Errorf("MongeElkanScore(%q, %q, %s) = %.17g; want %.17g ± %g\nderivation: %s",
						tt.a, tt.b, tt.inner, got, tt.want, mongeElkanEpsilon, tt.derivation)
				}
			}
			if got < 0.0 || got > 1.0 {
				t.Errorf("MongeElkanScore(%q, %q, %s) = %g; want in [0, 1]", tt.a, tt.b, tt.inner, got)
			}
		})
	}
}

// TestMongeElkanScore_AsymmetryDirectionSensitive is the LOAD-BEARING
// regression gate for RV-ME6 / RV-ME-asym: the same inputs in the two
// argument orders MUST produce different scores. A silent direction
// swap (or accidental symmetrisation) inside MongeElkanScore would
// cause both calls to return the same value — both would still be in
// [0, 1] and still pass RangeBounds + Identity — but this test catches
// the regression by asserting a MINIMUM separation.
//
// Mirrors the Tversky α≠β asymmetry test pattern.
func TestMongeElkanScore_AsymmetryDirectionSensitive(t *testing.T) {
	const a = "alpha beta gamma"
	const b = "alpha"
	inner := fuzzymatch.AlgoLevenshtein
	opts := fuzzymatch.DefaultNormalisationOptions()
	fwd := fuzzymatch.MongeElkanScore(a, b, inner, opts) // ME(B, A) per RV-ME6 (a here is the multi-token side)
	rev := fuzzymatch.MongeElkanScore(b, a, inner, opts) // ME(A, B) per RV-ME4
	delta := math.Abs(fwd - rev)
	// The actual difference is 1.0 - 0.4666… ≈ 0.5333; the threshold
	// 0.1 gates against silent direction-swap regressions while leaving
	// IEEE-754 rounding headroom.
	const minDelta = 0.1
	if delta <= minDelta {
		t.Errorf("MongeElkanScore asymmetry gate FAILED: fwd=%g (a→b), rev=%g (b→a), |Δ|=%g; want > %g (the input swap with fixed inner should produce direction-sensitive scores per RV-ME6 KEYSTONE)",
			fwd, rev, delta, minDelta)
	}
}

// TestMongeElkanScoreSymmetric pins the symmetric variant:
//
//   - identity → 1.0
//   - both-empty → 1.0 (STANDARD)
//   - one-empty → 0.0
//   - explicit `MongeElkanScoreSymmetric(a, b, ...) ==
//     (MongeElkanScore(a, b, ...) + MongeElkanScore(b, a, ...)) / 2.0`
//   - symmetry pin: `MongeElkanScoreSymmetric(a, b) ==
//     MongeElkanScoreSymmetric(b, a)` for the RV-ME-asym pair
func TestMongeElkanScoreSymmetric(t *testing.T) {
	opts := fuzzymatch.DefaultNormalisationOptions()
	// Identity.
	if got := fuzzymatch.MongeElkanScoreSymmetric("alpha", "alpha", fuzzymatch.AlgoJaroWinkler, opts); got != 1.0 {
		t.Errorf("MongeElkanScoreSymmetric identity violated: got %v, want 1.0", got)
	}
	// Both-empty.
	if got := fuzzymatch.MongeElkanScoreSymmetric("", "", fuzzymatch.AlgoJaroWinkler, opts); got != 1.0 {
		t.Errorf("MongeElkanScoreSymmetric both-empty: got %v, want 1.0", got)
	}
	// One-empty (both directions average to 0.0 since both directions are 0.0).
	if got := fuzzymatch.MongeElkanScoreSymmetric("hello", "", fuzzymatch.AlgoJaroWinkler, opts); got != 0.0 {
		t.Errorf("MongeElkanScoreSymmetric one-empty (a-non-empty): got %v, want 0.0", got)
	}
	if got := fuzzymatch.MongeElkanScoreSymmetric("", "hello", fuzzymatch.AlgoJaroWinkler, opts); got != 0.0 {
		t.Errorf("MongeElkanScoreSymmetric one-empty (b-non-empty): got %v, want 0.0", got)
	}
	// Explicit construction pin: symmetric is the mean of the two
	// asymmetric directions.
	const a = "alpha beta gamma"
	const b = "alpha"
	inner := fuzzymatch.AlgoLevenshtein
	asymAB := fuzzymatch.MongeElkanScore(a, b, inner, opts)
	asymBA := fuzzymatch.MongeElkanScore(b, a, inner, opts)
	sym := fuzzymatch.MongeElkanScoreSymmetric(a, b, inner, opts)
	want := (asymAB + asymBA) / 2.0
	if sym != want {
		t.Errorf("MongeElkanScoreSymmetric construction violated: got %v, want (%v + %v)/2 = %v", sym, asymAB, asymBA, want)
	}
	// Symmetry pin: swap (a, b) → same score (the load-bearing
	// arithmetic-mean order-independence property).
	symSwapped := fuzzymatch.MongeElkanScoreSymmetric(b, a, inner, opts)
	if sym != symSwapped {
		t.Errorf("MongeElkanScoreSymmetric not symmetric: ME_sym(a,b)=%v != ME_sym(b,a)=%v", sym, symSwapped)
	}
}

// TestMongeElkan_PanicsOnNonPermittedInner is the load-bearing
// exhaustive panic test per RESEARCH.md Pitfall 4 + CONTEXT.md §3
// LOCKED. Walks all 9 NON-permitted AlgoIDs and asserts each panics
// with the documented message format. When Phase 7 lands and adds the
// 4 phonetic AlgoIDs to permittedMongeElkanInner, this fixture's
// `rejected` slice shrinks from 9 to 5 entries; the test structure is
// unchanged.
//
// Also runs a representative-subset sanity check that 5 permitted
// AlgoIDs (one per tier — char/q-gram/gestalt + 2 character) return a
// value in [0, 1] without panic. The exhaustive permitted-list
// coverage lives in TestMongeElkanScore above (RV-ME1..RV-ME6 cover
// JaroWinkler + Levenshtein); this sanity check pins the OTHER 12
// dispatch slots fire correctly.
func TestMongeElkan_PanicsOnNonPermittedInner(t *testing.T) {
	rejected := []fuzzymatch.AlgoID{
		fuzzymatch.AlgoMongeElkan,     // self-recursion (RESEARCH.md Pitfall 4)
		fuzzymatch.AlgoTokenSortRatio, // token-on-token meaningless
		fuzzymatch.AlgoTokenSetRatio,  // token-on-token meaningless
		fuzzymatch.AlgoPartialRatio,   // token-on-token meaningless
		fuzzymatch.AlgoTokenJaccard,   // token-on-token meaningless
		// AlgoDoubleMetaphone: permitted as of plan 07-02 (removed from rejected)
		// AlgoNYSIIS: permitted as of plan 07-03 (removed from rejected)
		fuzzymatch.AlgoMRA, // Phase 7 reserved — 07-04 removes
	}
	for _, inner := range rejected {
		t.Run("rejected_"+inner.String(), func(t *testing.T) {
			func() {
				defer func() {
					r := recover()
					if r == nil {
						t.Errorf("MongeElkanScore(\"a b\", \"c d\", %s, opts) did not panic", inner)
						return
					}
					msg, ok := r.(string)
					if !ok {
						t.Errorf("panic value type = %T (%v); want string", r, r)
						return
					}
					// Verify the message contains BOTH the AlgoID name
					// AND the documented prefix; the exact format is
					// pinned by TestMongeElkan_PanicMessageFormat below.
					if !strings.Contains(msg, "fuzzymatch: AlgoID "+inner.String()) {
						t.Errorf("panic message %q does not start with documented prefix containing %s", msg, inner)
					}
					if !strings.Contains(msg, "not permitted as Monge-Elkan inner metric") {
						t.Errorf("panic message %q does not contain documented suffix", msg)
					}
				}()
				_ = fuzzymatch.MongeElkanScore("a b", "c d", inner, fuzzymatch.DefaultNormalisationOptions())
			}()
			// Also verify the symmetric variant surfaces the panic
			// (the panic fires on the FIRST MongeElkanScore call
			// inside MongeElkanScoreSymmetric).
			func() {
				defer func() {
					r := recover()
					if r == nil {
						t.Errorf("MongeElkanScoreSymmetric(\"a b\", \"c d\", %s, opts) did not panic", inner)
					}
				}()
				_ = fuzzymatch.MongeElkanScoreSymmetric("a b", "c d", inner, fuzzymatch.DefaultNormalisationOptions())
			}()
		})
	}
	// Representative-subset sanity check for the permitted side: pick
	// one AlgoID per tier so a regression that loosens or tightens the
	// allow-list surfaces here. The exhaustive permitted-list coverage
	// is provided by TestMongeElkanScore's RV-ME1..RV-ME6 (covering
	// JaroWinkler + Levenshtein) plus this loop covering the other 12
	// permitted entries.
	permittedSanity := []fuzzymatch.AlgoID{
		fuzzymatch.AlgoLevenshtein,
		fuzzymatch.AlgoDamerauLevenshteinOSA,
		fuzzymatch.AlgoDamerauLevenshteinFull,
		fuzzymatch.AlgoHamming,
		fuzzymatch.AlgoJaro,
		fuzzymatch.AlgoJaroWinkler,
		fuzzymatch.AlgoStrcmp95,
		fuzzymatch.AlgoSmithWatermanGotoh,
		fuzzymatch.AlgoLCSStr,
		fuzzymatch.AlgoQGramJaccard,
		fuzzymatch.AlgoSorensenDice,
		fuzzymatch.AlgoCosine,
		fuzzymatch.AlgoTversky,
		fuzzymatch.AlgoRatcliffObershelp,
		fuzzymatch.AlgoSoundex,         // plan 07-01 — phonetic tier addition
		fuzzymatch.AlgoDoubleMetaphone, // plan 07-02 — phonetic tier addition
		fuzzymatch.AlgoNYSIIS,          // plan 07-03 — phonetic tier addition
	}
	for _, inner := range permittedSanity {
		t.Run("permitted_"+inner.String(), func(t *testing.T) {
			got := fuzzymatch.MongeElkanScore("alpha beta", "alpha gamma", inner, fuzzymatch.DefaultNormalisationOptions())
			if got < 0.0 || got > 1.0 || math.IsNaN(got) || math.IsInf(got, 0) {
				t.Errorf("MongeElkanScore(\"alpha beta\", \"alpha gamma\", %s, opts) = %g; want finite value in [0, 1]", inner, got)
			}
		})
	}
}

// TestMongeElkan_PanicMessageFormat pins the EXACT panic message string
// format per CONTEXT.md §3 LOCKED. A regression in the message text
// (e.g. accidental rename, capitalisation drift) surfaces here.
func TestMongeElkan_PanicMessageFormat(t *testing.T) {
	const wantPrefix = "fuzzymatch: AlgoID "
	const wantSuffix = " not permitted as Monge-Elkan inner metric"
	cases := []struct {
		inner fuzzymatch.AlgoID
		name  string
	}{
		{fuzzymatch.AlgoMongeElkan, "MongeElkan"},
		{fuzzymatch.AlgoTokenSortRatio, "TokenSortRatio"},
		{fuzzymatch.AlgoMRA, "MRA"}, // AlgoNYSIIS now permitted (plan 07-03); MRA still non-permitted until 07-04
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			func() {
				defer func() {
					r := recover()
					if r == nil {
						t.Fatalf("did not panic for inner=%s", c.inner)
					}
					msg, ok := r.(string)
					if !ok {
						t.Fatalf("panic value type = %T (%v); want string", r, r)
					}
					want := wantPrefix + c.name + wantSuffix
					if msg != want {
						t.Errorf("panic message format drift:\n  got:  %q\n  want: %q", msg, want)
					}
				}()
				_ = fuzzymatch.MongeElkanScore("a b", "c d", c.inner, fuzzymatch.DefaultNormalisationOptions())
			}()
		})
	}
}

// TestMongeElkanScore_DispatchRegistration verifies the LOCKED CONTEXT.md
// §4 dispatch defaults: the AlgoMongeElkan dispatch slot binds the
// TestMongeElkanScore_BinaryInner_Soundex asserts the binary-inner-composition
// behaviour of MongeElkanScore when AlgoSoundex is used as the inner metric
// (per CONTEXT.md §4 LOCKED). Three sub-cases lock the contract:
//
//   - one_matches: "alpha beta" vs "alpha gamma" → Soundex("alpha")=="alpha"
//     identity gives 1.0 for first token; "beta" vs "gamma" codes differ →
//     0.0 for second token. MongeElkanScore = (1.0 + 0.0) / 2 = 0.5.
//
//   - both_match: "alpha beta" vs "alpha beta" → identity short-circuit
//     fires in SoundexScore; both tokens match. MongeElkanScore = 1.0.
//
//   - neither: "alpha" vs "gamma" → codes differ → 0.0.
//
// This test locks the binary-inner-composition behaviour against silent
// regression (e.g. a change to per-token-max accumulation logic that breaks
// ME over discrete-valued inners).
func TestMongeElkanScore_BinaryInner_Soundex(t *testing.T) {
	opts := fuzzymatch.DefaultNormalisationOptions()

	t.Run("one_matches", func(t *testing.T) {
		// "alpha beta" vs "alpha gamma": alpha matches (identity), beta vs gamma
		// have different Soundex codes (B300 vs G565 — wait: beta=B300, gamma=G650).
		// MongeElkanScore = max(SoundexScore("alpha","alpha"), SoundexScore("alpha","gamma"))
		//                 + max(SoundexScore("beta","alpha"), SoundexScore("beta","gamma"))
		// = max(1.0, 0.0) + max(0.0, 0.0) ) / 2 = 0.5
		got := fuzzymatch.MongeElkanScore("alpha beta", "alpha gamma", fuzzymatch.AlgoSoundex, opts)
		if got != 0.5 {
			t.Errorf("MongeElkanScore(\"alpha beta\", \"alpha gamma\", AlgoSoundex, opts) = %g; want 0.5 (one token matches)", got)
		}
	})

	t.Run("both_match", func(t *testing.T) {
		got := fuzzymatch.MongeElkanScore("alpha beta", "alpha beta", fuzzymatch.AlgoSoundex, opts)
		if got != 1.0 {
			t.Errorf("MongeElkanScore(\"alpha beta\", \"alpha beta\", AlgoSoundex, opts) = %g; want 1.0 (both tokens match)", got)
		}
	})

	t.Run("neither", func(t *testing.T) {
		// "alpha" vs "gamma": A416 vs G650 — codes differ → 0.0.
		got := fuzzymatch.MongeElkanScore("alpha", "gamma", fuzzymatch.AlgoSoundex, opts)
		if got != 0.0 {
			t.Errorf("MongeElkanScore(\"alpha\", \"gamma\", AlgoSoundex, opts) = %g; want 0.0 (no token match)", got)
		}
	})
}

// SYMMETRIC variant with AlgoJaroWinkler as the default inner and
// DefaultNormalisationOptions(). A regression here would break Phase 8
// Scorer integration silently.
func TestMongeElkanScore_DispatchRegistration(t *testing.T) {
	got := fuzzymatch.DispatchInvokeForTest(int(fuzzymatch.AlgoMongeElkan), "user create", "usr creating")
	want := fuzzymatch.MongeElkanScoreSymmetric("user create", "usr creating", fuzzymatch.AlgoJaroWinkler, fuzzymatch.DefaultNormalisationOptions())
	if got != want {
		t.Errorf("dispatch[AlgoMongeElkan] does not bind the LOCKED defaults: got %v, want %v (= MongeElkanScoreSymmetric with default inner = AlgoJaroWinkler + DefaultNormalisationOptions per CONTEXT §4 LOCKED)", got, want)
	}
	// Also pin identity on the dispatch surface.
	if got := fuzzymatch.DispatchInvokeForTest(int(fuzzymatch.AlgoMongeElkan), "hello", "hello"); got != 1.0 {
		t.Errorf("dispatch[AlgoMongeElkan](\"hello\",\"hello\") = %v; want 1.0 (identity short-circuit)", got)
	}
}

// TestMongeElkanScore_BinaryInner_DoubleMetaphone asserts the
// binary-inner-composition behaviour of MongeElkanScore when AlgoDoubleMetaphone
// is used as the inner metric (per CONTEXT.md §4 LOCKED). Three sub-cases lock
// the contract:
//
//   - one_matches: ME("alpha beta", "alpha gamma", DM) == 0.5
//     (one of two tokens matches phonetically; the other does not).
//   - both_match: ME("alpha beta", "alpha beta", DM) == 1.0
//     (full token-set match; identity short-circuit fires for each token).
//   - neither: "alpha" vs "gamma" → DM codes differ → 0.0.
//
// This test locks the binary-inner-composition behaviour against silent
// regression (e.g. a change to per-token-max accumulation logic that breaks
// ME over discrete-valued inners).
func TestMongeElkanScore_BinaryInner_DoubleMetaphone(t *testing.T) {
	opts := fuzzymatch.DefaultNormalisationOptions()

	t.Run("one_matches", func(t *testing.T) {
		// "alpha beta" vs "alpha gamma":
		// max(DM("alpha","alpha"), DM("alpha","gamma")) = max(1.0, 0.0) = 1.0
		// max(DM("beta","alpha"), DM("beta","gamma")) = max(0.0, 0.0) = 0.0
		// MongeElkanScore = (1.0 + 0.0) / 2 = 0.5
		got := fuzzymatch.MongeElkanScore("alpha beta", "alpha gamma", fuzzymatch.AlgoDoubleMetaphone, opts)
		if got != 0.5 {
			t.Errorf("MongeElkanScore(\"alpha beta\", \"alpha gamma\", AlgoDoubleMetaphone, opts) = %g; want 0.5 (one token matches)", got)
		}
	})

	t.Run("both_match", func(t *testing.T) {
		got := fuzzymatch.MongeElkanScore("alpha beta", "alpha beta", fuzzymatch.AlgoDoubleMetaphone, opts)
		if got != 1.0 {
			t.Errorf("MongeElkanScore(\"alpha beta\", \"alpha beta\", AlgoDoubleMetaphone, opts) = %g; want 1.0 (both tokens match)", got)
		}
	})

	t.Run("neither", func(t *testing.T) {
		// "alpha" vs "gamma": DM codes differ → 0.0.
		got := fuzzymatch.MongeElkanScore("alpha", "gamma", fuzzymatch.AlgoDoubleMetaphone, opts)
		if got != 0.0 {
			t.Errorf("MongeElkanScore(\"alpha\", \"gamma\", AlgoDoubleMetaphone, opts) = %g; want 0.0 (no token match)", got)
		}
	})
}

// TestMongeElkanScore_BinaryInner_NYSIIS asserts the binary-inner-composition
// behaviour of MongeElkanScore when AlgoNYSIIS is used as the inner metric
// (per CONTEXT.md §4 LOCKED). Three sub-cases lock the contract:
//
//   - one_matches: ME("alpha beta", "alpha gamma", NYSIIS) == 0.5
//     (one of two tokens matches phonetically; the other does not).
//   - both_match: ME("alpha beta", "alpha beta", NYSIIS) == 1.0
//     (full token-set match; identity short-circuit fires for each token).
//   - neither: "alpha" vs "gamma" → NYSIIS codes differ → 0.0.
//
// This test locks the binary-inner-composition behaviour against silent
// regression (e.g. a change to per-token-max accumulation logic that breaks
// ME over discrete-valued inners).
func TestMongeElkanScore_BinaryInner_NYSIIS(t *testing.T) {
	opts := fuzzymatch.DefaultNormalisationOptions()

	t.Run("one_matches", func(t *testing.T) {
		// "alpha beta" vs "alpha gamma":
		// max(NYSIIS("alpha","alpha"), NYSIIS("alpha","gamma")) = max(1.0, 0.0) = 1.0
		// max(NYSIIS("beta","alpha"), NYSIIS("beta","gamma")) = max(0.0, 0.0) = 0.0
		// MongeElkanScore = (1.0 + 0.0) / 2 = 0.5
		got := fuzzymatch.MongeElkanScore("alpha beta", "alpha gamma", fuzzymatch.AlgoNYSIIS, opts)
		if got != 0.5 {
			t.Errorf("MongeElkanScore(\"alpha beta\", \"alpha gamma\", AlgoNYSIIS, opts) = %g; want 0.5 (one token matches)", got)
		}
	})

	t.Run("both_match", func(t *testing.T) {
		got := fuzzymatch.MongeElkanScore("alpha beta", "alpha beta", fuzzymatch.AlgoNYSIIS, opts)
		if got != 1.0 {
			t.Errorf("MongeElkanScore(\"alpha beta\", \"alpha beta\", AlgoNYSIIS, opts) = %g; want 1.0 (both tokens match)", got)
		}
	})

	t.Run("neither", func(t *testing.T) {
		// "alpha" vs "gamma": NYSIIS codes differ → 0.0.
		got := fuzzymatch.MongeElkanScore("alpha", "gamma", fuzzymatch.AlgoNYSIIS, opts)
		if got != 0.0 {
			t.Errorf("MongeElkanScore(\"alpha\", \"gamma\", AlgoNYSIIS, opts) = %g; want 0.0 (no token match)", got)
		}
	})
}
