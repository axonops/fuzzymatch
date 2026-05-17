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

// swg.go implements the Smith-Waterman-Gotoh local-alignment similarity with
// affine gap penalty for the fuzzymatch catalogue.
//
// Sources:
//   - Smith, T. F. & Waterman, M. S. (1981). "Identification of common
//     molecular subsequences." J. Mol. Biol. 147:195-197 (local-alignment
//     formulation).
//   - Gotoh, O. (1982). "An improved algorithm for matching biological
//     sequences." J. Mol. Biol. 162:705-708 (affine-gap O(mn) reduction).
//   - Flouri, T. et al. (2015). "Are all global alignment algorithms and
//     implementations correct?" biorxiv 031500 — documents the Gotoh 1982
//     initialisation erratum and the corrected formulation transcribed here.
//
// Gotoh 1982 contains a known erratum in the affine-gap initialisation step
// (the global-alignment border setup that textbook treatments often blur into
// local alignment); this implementation uses the corrected formulation per
// Flouri et al. 2015: every border cell of M, Ix, Iy initialises to 0 for
// LOCAL alignment (NOT -Inf, NOT the global-alignment gap-open ladder).
// Five of ten implementations audited by Flouri et al. reproduced the bug;
// the gap-split canary scenario in tests/bdd/features/swg.feature and the
// GapSplitInvariance property test in props_test.go gate against regressions.
//
// Recurrence (corrected per Flouri et al. 2015; for LOCAL alignment every
// border cell of M, Ix, Iy initialises to 0). Three matrices: M (match/
// mismatch ending), Ix (gap in a / insertion in b), Iy (gap in b / insertion
// in a). For i = 1..m, j = 1..n, with s(a[i-1], b[j-1]) = Match if equal else
// Mismatch:
//
//	M[i,j]  = max( 0,
//	               M[i-1,j-1]  + s(a[i-1], b[j-1]),
//	               Ix[i-1,j-1] + s(a[i-1], b[j-1]),
//	               Iy[i-1,j-1] + s(a[i-1], b[j-1]) )
//
//	Ix[i,j] = max( 0,
//	               M[i-1,j]    + GapOpen,
//	               Ix[i-1,j]   + GapExtend )
//
//	Iy[i,j] = max( 0,
//	               M[i,j-1]    + GapOpen,
//	               Iy[i,j-1]   + GapExtend )
//
// Best raw score: bestRaw = max over all (i, j) of M[i, j], tracked during the
// fill (no post-pass scan). Normalised: clamp(bestRaw / min(len(a), len(b)), 0, 1).
//
// Implementation discipline (inherits Phase 2):
//
//   - ASCII fast path operates on bytes directly when the shorter dimension
//     n <= maxStackInputLen && isASCII(a) && isASCII(b); a stack-allocated
//     [(maxStackInputLen+1)*6]float64 buffer (3120 bytes) holds the six rolling
//     rows (prevM, currM, prevIx, currIx, prevIy, currIy).
//     (maxStackInputLen is defined in levenshtein.go — do NOT redeclare.)
//   - Heap path: six make([]float64, n+1) calls; 6 allocs on ASCII Long.
//     There is NO stack fast path for the rune path: smithWatermanGotohRawRunes
//     unconditionally heap-allocates the six rows regardless of input length,
//     so the rune path's allocation count is 8 (two []rune + six rows) as a
//     MINIMUM for any rune input — not an achievable target for "short" runes.
//   - NO init()-time table builds (per docs/requirements.md §5(12) and
//     .claude/skills/determinism-standards): no var-level side effects.
//   - NO map iteration on output paths (DET-03).
//   - NO transcendental float operations (DET-06) on the SCORING hot path:
//     only +, -, *, /, max-style if-comparison, and float64() conversion
//     inside swgDPRaw / swgDPRawRunes / swgClampNormalise. The forbidden
//     stdlib intrinsics enumerated in determinism-standards §13.3 are not
//     referenced anywhere on the scoring path. (math.IsNaN / math.IsInf
//     appear in (SWGParams).validate(), which is a parameter-checking
//     surface, not a scoring path — these functions inspect bit patterns
//     and are deterministic across platforms per IEEE-754 §6.3.)
//   - NO goroutines, channels, or mutexes.
//   - Rune variants allocate two []rune slices — documented per Phase 2
//     Pattern 8.
//   - 0-alloc budget applies only to the byte path on ASCII inputs whose
//     shorter dimension <= maxStackInputLen.

package fuzzymatch

import (
	"fmt"
	"math"
)

// SWGParams holds the affine-gap parameters for Smith-Waterman-Gotoh local
// alignment. All fields are exported; SWGParams is a value type (no pointer
// receivers required). Conventionally: Match >= 0, Mismatch <= 0,
// GapOpen <= GapExtend <= 0 (extending an existing gap is cheaper than
// opening a new one).
//
// Callers may pass nonsense values (e.g. Match < 0, GapOpen > 0, NaN, +Inf):
// the algorithm still produces a deterministic — though potentially
// meaningless — score. No validation is performed in the *Score / *RawScore
// entry points; consumers who want a strict-parameter gate (per the Q2
// data-vs-parameter framework in docs/requirements.md §6.A) call
// (SWGParams).Validate() after mutation, which panics with a typed-error
// value wrapping ErrInvalidSWGParam when the invariants are violated.
type SWGParams struct {
	// Match is the reward for a matching position. Should be >= 0.
	// Default 1.0 (per NewSWGParams).
	Match float64

	// Mismatch is the penalty for a mismatching position. Should be <= 0.
	// Default -1.0 (per NewSWGParams).
	Mismatch float64

	// GapOpen is the penalty for opening a new gap. Should be <= 0 and
	// conventionally <= GapExtend (more expensive than extending).
	// Default -1.5 (per NewSWGParams).
	GapOpen float64

	// GapExtend is the penalty for extending an existing gap. Should be <= 0
	// and conventionally >= GapOpen (cheaper than opening).
	// Default -0.5 (per NewSWGParams).
	GapExtend float64
}

// NewSWGParams returns SWGParams populated with the documented defaults
// (Match=1.0, Mismatch=-1.0, GapOpen=-1.5, GapExtend=-0.5). Callers may
// override individual fields after construction:
//
//	params := NewSWGParams()
//	params.Match = 2.0
//	score := SmithWatermanGotohScoreWithParams(a, b, params)
//
// The returned value is a fresh copy; callers can mutate freely without
// affecting subsequent NewSWGParams() invocations.
//
// Validation (Phase 8.5 Gap 7):
//
// The returned defaults are guaranteed to satisfy (SWGParams).Validate()
// — they pass Match ≥ 0, Mismatch ≤ 0, GapOpen ≤ 0, GapExtend ≤ 0 and
// every field is finite. Consumers MUTATING the returned struct should
// call Validate() before passing it to SmithWatermanGotohScoreWithParams
// when the mutated values come from external configuration or upstream
// arithmetic that might introduce NaN / Inf / sign-flipped values. A
// defence-in-depth self-test runs inside NewSWGParams to catch the
// hypothetical case where the locked defaults are tampered with at
// build time (e.g. via a -ldflags injection); the self-test panics
// with a typed-error value wrapping ErrInternalInvariantViolated.
func NewSWGParams() SWGParams {
	p := SWGParams{
		Match:     1.0,
		Mismatch:  -1.0,
		GapOpen:   -1.5,
		GapExtend: -0.5,
	}
	// Defence-in-depth: validate the baked-in defaults. The locked
	// defaults always pass; this self-test fires only if the constants
	// above have been tampered with (a programmer error / build-time
	// injection), which is an internal-invariant violation rather than
	// a caller error — hence ErrInternalInvariantViolated instead of
	// ErrInvalidSWGParam.
	if err := p.validate(); err != nil {
		panic(fmt.Errorf("%w: NewSWGParams default constants violate invariants: %w", ErrInternalInvariantViolated, err))
	}
	return p
}

// Validate checks the affine-gap parameter invariants documented on
// SWGParams: Match must be a finite, non-negative float; Mismatch,
// GapOpen, and GapExtend must be finite, non-positive floats. NaN and
// ±Inf are rejected on every field.
//
// Validate panics with a typed-error value wrapping ErrInvalidSWGParam
// when any invariant is violated — the Q2 strict-parameter framework
// from docs/requirements.md §6.A applies. Consumers discriminate via
// errors.Is on a recovered panic value:
//
//	defer func() {
//	    if r := recover(); r != nil {
//	        if err, ok := r.(error); ok && errors.Is(err, fuzzymatch.ErrInvalidSWGParam) {
//	            // programmer error — log and re-panic, or substitute defaults
//	        }
//	    }
//	}()
//	params := fuzzymatch.NewSWGParams()
//	params.Match = math.NaN() // consumer mutation
//	params.Validate()         // panics with a typed error wrapping ErrInvalidSWGParam
//
// Callers who prefer a non-panicking surface inspect SWGParams's fields
// directly — the per-field constraints are simple enough to assert
// inline.
func (p SWGParams) Validate() {
	if err := p.validate(); err != nil {
		panic(fmt.Errorf("%w: %w", ErrInvalidSWGParam, err))
	}
}

// validate returns a non-nil error describing the first invariant
// violation found, or nil if every field satisfies the documented
// constraints. Used by NewSWGParams (with the defence-in-depth
// ErrInternalInvariantViolated wrap) and Validate (with the
// ErrInvalidSWGParam wrap) so the two surfaces share a single
// invariant definition.
func (p SWGParams) validate() error { //nolint:gocyclo // 4 per-field invariant blocks (Match/Mismatch/GapOpen/GapExtend) each splitting NaN-or-Inf vs sign — the linear structure is the readable form for "first violation wins"
	if math.IsNaN(p.Match) || math.IsInf(p.Match, 0) {
		return fmt.Errorf("field Match must be finite (got %v)", p.Match)
	}
	if p.Match < 0 {
		return fmt.Errorf("field Match must be >= 0 (got %v)", p.Match)
	}
	if math.IsNaN(p.Mismatch) || math.IsInf(p.Mismatch, 0) {
		return fmt.Errorf("field Mismatch must be finite (got %v)", p.Mismatch)
	}
	if p.Mismatch > 0 {
		return fmt.Errorf("field Mismatch must be <= 0 (got %v)", p.Mismatch)
	}
	if math.IsNaN(p.GapOpen) || math.IsInf(p.GapOpen, 0) {
		return fmt.Errorf("field GapOpen must be finite (got %v)", p.GapOpen)
	}
	if p.GapOpen > 0 {
		return fmt.Errorf("field GapOpen must be <= 0 (got %v)", p.GapOpen)
	}
	if math.IsNaN(p.GapExtend) || math.IsInf(p.GapExtend, 0) {
		return fmt.Errorf("field GapExtend must be finite (got %v)", p.GapExtend)
	}
	if p.GapExtend > 0 {
		return fmt.Errorf("field GapExtend must be <= 0 (got %v)", p.GapExtend)
	}
	return nil
}

// SmithWatermanGotohScore returns the Smith-Waterman-Gotoh local-alignment
// similarity between a and b as a value in [0.0, 1.0] using the documented
// default parameters (Match=1.0, Mismatch=-1.0, GapOpen=-1.5, GapExtend=-0.5).
//
// The returned score is CLAMPED: if the underlying alignment score is
// negative (e.g. two unrelated strings dominated by mismatch/gap penalties)
// the clamp returns 0.0; if it exceeds min(len(a), len(b)) (custom params
// with Match > 1.0) the clamp returns 1.0. Use SmithWatermanGotohRawScore
// for the unclamped raw alignment score.
//
// Normalisation:
//
//	score = clamp(best_local_score / min(len(a), len(b)), 0.0, 1.0)
//
// Edge cases:
//   - SmithWatermanGotohScore("", "") == 1.0 exactly (both-empty identity)
//   - SmithWatermanGotohScore("", "abc") == 0.0 exactly (one-empty)
//   - SmithWatermanGotohScore(x, x) == 1.0 for any non-empty x (identity)
//   - SmithWatermanGotohScore(a, b) == SmithWatermanGotohScore(b, a) (symmetric)
//
// This function operates on bytes. For multi-byte UTF-8 inputs, use
// SmithWatermanGotohScoreRunes to obtain the rune-aware similarity.
//
// Worst-case time: O(m·n) where m = len(a), n = len(b).
// Space: O(min(m,n)) — three-matrix two-row DP (six rolling rows), no full
// m×n table allocated.
func SmithWatermanGotohScore(a, b string) float64 {
	return SmithWatermanGotohScoreWithParams(a, b, NewSWGParams())
}

// SmithWatermanGotohScoreWithParams returns the Smith-Waterman-Gotoh
// local-alignment similarity between a and b in [0.0, 1.0] using the
// supplied affine-gap params. Score normalisation and clamp semantics are
// identical to SmithWatermanGotohScore.
//
// No validation is performed on params: nonsense values (e.g. GapOpen > 0,
// NaN, +Inf) produce a deterministic-but-meaningless result. See SWGParams
// for the documented sign convention.
func SmithWatermanGotohScoreWithParams(a, b string, params SWGParams) float64 {
	if a == b {
		return 1.0 // identity short-circuit (covers both-empty and identical inputs)
	}
	la, lb := len(a), len(b)
	if la == 0 || lb == 0 {
		return 0.0
	}
	// Ensure b is the shorter dimension so the inner-loop dimension is minimal.
	if la < lb {
		a, b = b, a
		la, lb = lb, la
	}
	raw := smithWatermanGotohRawByte(a, b, la, lb, params)
	return swgClampNormalise(raw, lb)
}

// SmithWatermanGotohScoreRunes returns the Smith-Waterman-Gotoh
// local-alignment similarity treating a and b as sequences of Unicode code
// points (runes) rather than bytes. The score is in [0.0, 1.0], where 1.0
// means identical and 0.0 means maximally dissimilar.
//
// Normalisation uses the rune count: clamp(best_local_score / min(runeLen(a),
// runeLen(b)), 0, 1).
//
// The rune variant allocates two []rune slices. For ASCII inputs, prefer
// SmithWatermanGotohScore (zero allocations on inputs ≤ 64 bytes).
//
// There is intentionally no SmithWatermanGotohScoreRunesWithParams: the v1.0
// public surface pairs the *WithParams variants only with the byte path.
// Consumers needing custom params on Unicode-aware input should normalise
// their input to ASCII via Normalise first (folding diacritics) and then
// call SmithWatermanGotohScoreWithParams.
func SmithWatermanGotohScoreRunes(a, b string) float64 {
	if a == b {
		return 1.0 // fast identity — saves two []rune allocations (IN-02 pattern)
	}
	ra := []rune(a) // 1 alloc
	rb := []rune(b) // 1 alloc
	la, lb := len(ra), len(rb)
	if la == 0 || lb == 0 {
		return 0.0
	}
	if la < lb {
		ra, rb = rb, ra
		la, lb = lb, la
	}
	raw := smithWatermanGotohRawRunes(ra, rb, la, lb, NewSWGParams())
	return swgClampNormalise(raw, lb)
}

// SmithWatermanGotohRawScore returns the UNCLAMPED raw Smith-Waterman-Gotoh
// local-alignment score between a and b using the documented default
// parameters. The returned value may be negative (two unrelated strings) or
// greater than min(len(a), len(b)) for custom high-Match params; it is NOT
// normalised to [0, 1].
//
// Use SmithWatermanGotohScore for the normalised [0, 1] similarity. The Raw*
// surface is intended for advanced consumers (bioinformatics, schema-
// similarity research) that need absolute alignment quality unaffected by
// the normalisation choice.
//
// Edge cases:
//   - SmithWatermanGotohRawScore("", "") == 0.0 (both-empty: no positions to score)
//   - SmithWatermanGotohRawScore("", "abc") == 0.0 (one-empty)
//   - SmithWatermanGotohRawScore(x, x) == Match * float64(len(x)) for non-empty x
//   - SmithWatermanGotohRawScore(a, b) == SmithWatermanGotohRawScore(b, a) (symmetric)
//
// This function operates on bytes. For multi-byte UTF-8 inputs, use
// SmithWatermanGotohRawScoreRunes.
func SmithWatermanGotohRawScore(a, b string) float64 {
	return SmithWatermanGotohRawScoreWithParams(a, b, NewSWGParams())
}

// SmithWatermanGotohRawScoreWithParams returns the UNCLAMPED raw local-
// alignment score between a and b using the supplied affine-gap params. The
// returned value may be negative or exceed min(len(a), len(b)); it is NOT
// normalised to [0, 1].
//
// Identity-with-non-empty returns params.Match * float64(len(a)) (every
// position matches with no gaps). Identity-with-both-empty returns 0.0.
func SmithWatermanGotohRawScoreWithParams(a, b string, params SWGParams) float64 {
	if a == b {
		if len(a) == 0 {
			return 0.0 // both-empty: no positions to score
		}
		return params.Match * float64(len(a)) // identity: every position matches, no gaps
	}
	la, lb := len(a), len(b)
	if la == 0 || lb == 0 {
		return 0.0
	}
	if la < lb {
		a, b = b, a
		la, lb = lb, la
	}
	return smithWatermanGotohRawByte(a, b, la, lb, params)
}

// SmithWatermanGotohRawScoreRunes returns the UNCLAMPED raw Smith-Waterman-
// Gotoh local-alignment score treating a and b as sequences of Unicode code
// points (runes) rather than bytes.
//
// Identity-with-non-empty returns Match * float64(len([]rune(a))) (every
// rune position matches). Identity-with-both-empty returns 0.0.
//
// The rune variant allocates two []rune slices. For ASCII inputs, prefer
// SmithWatermanGotohRawScore.
//
// There is intentionally no SmithWatermanGotohRawScoreRunesWithParams: as
// with SmithWatermanGotohScoreRunes, the *WithParams variants are paired
// only with the byte path. Normalise to ASCII first if you need custom
// params on Unicode-aware input.
func SmithWatermanGotohRawScoreRunes(a, b string) float64 {
	if a == b {
		// Identity short-circuit. Logic mirrors the byte-path identity branch
		// in SmithWatermanGotohRawScoreWithParams above (line ~256), but uses
		// the rune count rather than byte count so multi-byte UTF-8 inputs
		// (e.g. "café") score by character, not by byte. The duplication is
		// deliberate: factoring the rune-path through the byte-path entry
		// would force a []rune conversion before the identity test, defeating
		// the short-circuit's purpose.
		if a == "" {
			return 0.0 // both-empty
		}
		// Identity: every rune matches with no gaps. Use rune count, not byte count.
		return NewSWGParams().Match * float64(len([]rune(a)))
	}
	ra := []rune(a) // 1 alloc
	rb := []rune(b) // 1 alloc
	la, lb := len(ra), len(rb)
	if la == 0 || lb == 0 {
		return 0.0
	}
	if la < lb {
		ra, rb = rb, ra
		la, lb = lb, la
	}
	return smithWatermanGotohRawRunes(ra, rb, la, lb, NewSWGParams())
}

// swgClampNormalise returns clamp(raw / float64(minLen), 0, 1). minLen MUST
// be > 0 (callers guard the empty-input case before calling).
//
// Inlined per file-level discipline: no math.Min / math.Max (transcendental-
// adjacent on some platforms via intrinsic emission). Only +/-/*/comparisons.
func swgClampNormalise(raw float64, minLen int) float64 {
	n := raw / float64(minLen)
	if n < 0 {
		return 0.0
	}
	if n > 1 {
		return 1.0
	}
	return n
}

// smithWatermanGotohRawByte performs the ASCII fast-path gate then dispatches
// to swgDPRaw with either a stack-allocated or heap-allocated row set.
// Caller has ensured la >= lb > 0.
func smithWatermanGotohRawByte(a, b string, la, lb int, params SWGParams) float64 {
	if lb <= maxStackInputLen && isASCII(a) && isASCII(b) {
		// Stack-allocated buffer: 3120 bytes = (maxStackInputLen+1) * 6 * 8.
		// Escape analysis confirmed buf does not escape (the six slices point
		// into buf; swgDPRaw treats them as length-(lb+1) []float64 windows).
		var buf [(maxStackInputLen + 1) * 6]float64
		n1 := lb + 1
		return swgDPRaw(a, b, la, lb, params,
			buf[0*n1:1*n1], buf[1*n1:2*n1],
			buf[2*n1:3*n1], buf[3*n1:4*n1],
			buf[4*n1:5*n1], buf[5*n1:6*n1])
	}
	// Heap path: 6 allocations of float64 slices.
	return swgDPRaw(a, b, la, lb, params,
		make([]float64, lb+1), make([]float64, lb+1),
		make([]float64, lb+1), make([]float64, lb+1),
		make([]float64, lb+1), make([]float64, lb+1))
}

// swgDPRaw computes the raw Smith-Waterman-Gotoh local alignment score.
// Caller has ensured len(a) >= len(b) > 0 (m, n) and supplied six row buffers
// each of length n+1.
//
// The kernel uses three two-row buffer pairs:
//   - prevM, currM  — match/mismatch ending matrix M
//   - prevIx, currIx — gap-in-a (insertion in b) matrix Ix
//   - prevIy, currIy — gap-in-b (insertion in a) matrix Iy
//
// Each cell at [i][j] depends only on [i-1][j-1], [i-1][j], and [i][j-1], so
// the full m×n tables reduce to two rolling rows per matrix. Six rolling
// rows total.
//
// Local-alignment correctness gate: every border cell of M, Ix, Iy
// initialises to 0 (corrected per Flouri et al. 2015). For LOCAL alignment
// this is the right initialisation; the Gotoh 1982 erratum primarily
// affects the global-alignment border setup but textbook treatments often
// propagate the wrong border into local implementations.
//
// Best raw score tracked during the fill (no post-pass scan); only M's max
// is tracked because Ix/Iy contribute to bestRaw only via their feed into M.
func swgDPRaw(a, b string, m, n int, params SWGParams, //nolint:gocyclo // SWG three-matrix kernel — match/Ix/Iy recurrence + per-cell max-with-0 + running bestRaw; extraction would obscure the recurrence; see godoc above
	prevM, currM, prevIx, currIx, prevIy, currIy []float64,
) float64 {
	// Local-alignment zero-init: every border cell is 0 (Flouri et al. 2015).
	for j := 0; j <= n; j++ {
		prevM[j] = 0
		prevIx[j] = 0
		prevIy[j] = 0
	}
	bestRaw := 0.0
	for i := 1; i <= m; i++ {
		// Border at column 0 also initialises to 0 for local alignment.
		currM[0] = 0
		currIx[0] = 0
		currIy[0] = 0
		ai := a[i-1]
		for j := 1; j <= n; j++ {
			var sij float64
			if ai == b[j-1] {
				sij = params.Match
			} else {
				sij = params.Mismatch
			}

			// M[i,j]: best of (start fresh = 0) / (extend from M) /
			// (close gap on either side then match).
			m1 := prevM[j-1] + sij
			m2 := prevIx[j-1] + sij
			m3 := prevIy[j-1] + sij
			mij := 0.0
			if m1 > mij {
				mij = m1
			}
			if m2 > mij {
				mij = m2
			}
			if m3 > mij {
				mij = m3
			}
			currM[j] = mij

			// Ix[i,j]: open new gap in a from M, or extend existing Ix.
			x1 := prevM[j] + params.GapOpen
			x2 := prevIx[j] + params.GapExtend
			xij := 0.0
			if x1 > xij {
				xij = x1
			}
			if x2 > xij {
				xij = x2
			}
			currIx[j] = xij

			// Iy[i,j]: open new gap in b from M, or extend existing Iy.
			y1 := currM[j-1] + params.GapOpen
			y2 := currIy[j-1] + params.GapExtend
			yij := 0.0
			if y1 > yij {
				yij = y1
			}
			if y2 > yij {
				yij = y2
			}
			currIy[j] = yij

			if mij > bestRaw {
				bestRaw = mij
			}
		}
		// Swap row buffers for the next outer iteration.
		prevM, currM = currM, prevM
		prevIx, currIx = currIx, prevIx
		prevIy, currIy = currIy, prevIy
	}
	return bestRaw
}

// smithWatermanGotohRawRunes computes the raw Smith-Waterman-Gotoh local
// alignment score on []rune slices. Caller has ensured la >= lb > 0.
//
// Six row buffers are heap-allocated unconditionally (no rune fast path —
// rune inputs go through the heap to avoid duplicating the buffer-window
// logic). Documented allocation cost: 6 row slices on top of the two []rune
// conversions performed at the *Runes entry point (8 total per call).
func smithWatermanGotohRawRunes(ra, rb []rune, la, lb int, params SWGParams) float64 {
	prevM := make([]float64, lb+1)
	currM := make([]float64, lb+1)
	prevIx := make([]float64, lb+1)
	currIx := make([]float64, lb+1)
	prevIy := make([]float64, lb+1)
	currIy := make([]float64, lb+1)
	return swgDPRawRunes(ra, rb, la, lb, params,
		prevM, currM, prevIx, currIx, prevIy, currIy)
}

// swgDPRawRunes mirrors swgDPRaw but indexes []rune slices instead of string
// bytes. Same correctness gates (zero-init borders for local alignment, max
// tracked during fill, three-matrix two-row form).
func swgDPRawRunes(ra, rb []rune, m, n int, params SWGParams, //nolint:gocyclo // SWG three-matrix kernel mirrors swgDPRaw on []rune; same recurrence complexity; see godoc above
	prevM, currM, prevIx, currIx, prevIy, currIy []float64,
) float64 {
	for j := 0; j <= n; j++ {
		prevM[j] = 0
		prevIx[j] = 0
		prevIy[j] = 0
	}
	bestRaw := 0.0
	for i := 1; i <= m; i++ {
		currM[0] = 0
		currIx[0] = 0
		currIy[0] = 0
		ai := ra[i-1]
		for j := 1; j <= n; j++ {
			var sij float64
			if ai == rb[j-1] {
				sij = params.Match
			} else {
				sij = params.Mismatch
			}

			m1 := prevM[j-1] + sij
			m2 := prevIx[j-1] + sij
			m3 := prevIy[j-1] + sij
			mij := 0.0
			if m1 > mij {
				mij = m1
			}
			if m2 > mij {
				mij = m2
			}
			if m3 > mij {
				mij = m3
			}
			currM[j] = mij

			x1 := prevM[j] + params.GapOpen
			x2 := prevIx[j] + params.GapExtend
			xij := 0.0
			if x1 > xij {
				xij = x1
			}
			if x2 > xij {
				xij = x2
			}
			currIx[j] = xij

			y1 := currM[j-1] + params.GapOpen
			y2 := currIy[j-1] + params.GapExtend
			yij := 0.0
			if y1 > yij {
				yij = y1
			}
			if y2 > yij {
				yij = y2
			}
			currIy[j] = yij

			if mij > bestRaw {
				bestRaw = mij
			}
		}
		prevM, currM = currM, prevM
		prevIx, currIx = currIx, prevIx
		prevIy, currIy = currIy, prevIy
	}
	return bestRaw
}
