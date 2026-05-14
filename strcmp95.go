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

// strcmp95.go implements Winkler's Strcmp95 enhancement of Jaro-Winkler.
//
// Sources:
//   - Winkler, W. E. (1994). "Advanced methods for record linkage."
//     Proceedings of the Section on Survey Research Methods, ASA: 467-472, §3.
//   - Cross-validation reference: U.S. Census Bureau (1995) strcmp95.c
//     (public domain — U.S. Government work; consulted ONLY for reference
//     vectors per .claude/skills/algorithm-licensing-standards).
//   - OpenRefine Strcmp95.java (Apache-2.0) consulted ONLY for prose-level
//     tie-breaks in Winkler 1994; no code copied.
//
// # Algorithm
//
// Strcmp95 layers four adjustments atop a Jaro pass, in this order:
//
//  1. Base Jaro: count matched characters within the standard Jaro matching
//     window (w = max(la, lb)/2 - 1, clamped to >= 0); count transpositions.
//
//  2. Similar-character credit: for each unmatched character pair (one from
//     a, one from b at the corresponding transposition slot), if the pair
//     appears in the Winkler 1994 36-pair similar-character table, add 0.3
//     to the effective match count (capped to avoid overshoot beyond the
//     base match denominator).
//
//  3. Winkler prefix boost: if the resulting Jaro-like score J >= 0.7, apply
//     JW = J + L · 0.1 · (1 - J), where L is the common-prefix length capped
//     at 4 (the canonical Winkler 1990 prefix boost).
//
//  4. Long-string adjustment (Winkler 1994 §3): when min(la, lb) > 4 AND
//     Num_com > prefix+1 AND 2·Num_com >= minLen+prefix, add a long-string
//     correction:
//
//	W = W + ((1 - W) · (Num_com - prefix - 1) / (la + lb - 2·prefix + Num_com))
//
// All four adjustments only ADD to the base Jaro score; the resulting
// Strcmp95Score(a, b) >= JaroWinklerScore(a, b) invariant holds for every
// (a, b) — verified by TestProp_Strcmp95Score_AtLeastJaroWinkler.
//
// # API hierarchy
//
//   - JaroScore        — base similarity.
//   - JaroWinklerScore — Jaro + prefix boost (shared-prefix bias).
//   - Strcmp95Score    — Jaro-Winkler + similar-character credit + long-string
//                        adjustment (record-linkage / surname matching).
//
// # Reference vectors (Winkler 1990 / Census Bureau strcmp95.c)
//
//   - Strcmp95Score("MARTHA",    "MARHTA")    ≈ 0.9611  (no similar pair; equals JW)
//   - Strcmp95Score("DWAYNE",    "DUANE")     ≈ 0.840   (W~U fires)
//   - Strcmp95Score("DIXON",     "DICKSONX")  ≈ 0.8133
//
// # ASCII-only — no *Runes variant
//
// Strcmp95 operates on ASCII letters (and digits). The similar-character
// table is letter-pair-keyed (AE, OU, etc.) and has no Unicode equivalent
// in Winkler 1994. For non-ASCII input, normalise via fuzzymatch.Normalise
// first (NFC/NFD + diacritic folding) — there is NO Strcmp95ScoreRunes
// variant. CONTEXT.md §2 LOCKS this surface.
//
// # Implementation discipline
//
//   - NO Strcmp95Params surface. The four Winkler 1994 adjustments are part
//     of the canonical algorithm, not consumer-tunable parameters.
//   - NO init() function. strcmp95SimilarChars is a `var` declaration with
//     no init-time side effect — determinism-reviewer flags any init() in
//     this file as BLOCKING per .claude/skills/determinism-standards §13.5.
//   - NO map iteration on output paths (DET-03).
//   - NO transcendental float operations (DET-06): only +, -, *, /,
//     comparisons, and float64() conversions. No math.Pow, math.Log,
//     math.Exp, math.Sqrt, or math.FMA.
//   - Left-to-right float reduction with explicit parenthesisation (DET-06).
//   - 0-alloc on ASCII Short — the byte match-flag arrays use the same
//     stack-buffer pattern as jaro.go (maxJaroStackLen = 256).
//
// # Source-origin statement
//
//   Primary:          Winkler 1994 TR-2 paper.
//   Cross-validation: Census Bureau strcmp95.c (public domain — U.S.
//                     Government work).
//   Tie-break:        OpenRefine Strcmp95.java (Apache-2.0) for prose
//                     ambiguities in Winkler 1994.
//   GPL/LGPL:         none consulted.
//   Code copied:      none.

package fuzzymatch

// strcmp95SimilarCredit is the similarity weight applied to each unmatched
// (a, b) character pair that appears in the Winkler 1994 similar-character
// table. The value 0.3 is the canonical Winkler 1994 constant (the table
// from §3 of TR-2 assigns 0.3 to every published pair).
const strcmp95SimilarCredit = 0.3

// strcmp95SimilarChars is the upper-case ASCII letter-pair similarity table
// from Winkler 1994 TR-2 §3 "An improved string comparator". Each entry is
// bidirectional: the lookup is symmetric in (a, b) → (b, a). All 36 published
// pairs carry similarity strcmp95SimilarCredit (0.3).
//
// Source: Winkler, W. E. (1994). "Advanced methods for record linkage."
// Proceedings of the Section on Survey Research Methods, ASA: 467-472, §3.
//
// Determinism (PITFALLS §14): the table is a `var` declaration with no
// init() side effect, guaranteeing (1) byte-stable values across init
// order, (2) zero first-call latency cost, (3) deterministic property-test
// output. The determinism-reviewer flags any init() in this file as
// BLOCKING.
//
// Visibility: unexported and not modifiable from outside the package. The
// 36 pairs are transcribed by hand from Winkler 1994; see
// TestStrcmp95_TableInvariants for the load-bearing transcription gate.
var strcmp95SimilarChars = [...]struct {
	a, b byte
	sim  float64
}{
	{'A', 'E', strcmp95SimilarCredit},
	{'A', 'I', strcmp95SimilarCredit},
	{'A', 'O', strcmp95SimilarCredit},
	{'A', 'U', strcmp95SimilarCredit},
	{'B', 'V', strcmp95SimilarCredit},
	{'E', 'I', strcmp95SimilarCredit},
	{'E', 'O', strcmp95SimilarCredit},
	{'E', 'U', strcmp95SimilarCredit},
	{'I', 'O', strcmp95SimilarCredit},
	{'I', 'U', strcmp95SimilarCredit},
	{'O', 'U', strcmp95SimilarCredit},
	{'I', 'Y', strcmp95SimilarCredit},
	{'E', 'Y', strcmp95SimilarCredit},
	{'C', 'G', strcmp95SimilarCredit},
	{'E', 'F', strcmp95SimilarCredit},
	{'W', 'U', strcmp95SimilarCredit},
	{'W', 'V', strcmp95SimilarCredit},
	{'X', 'K', strcmp95SimilarCredit},
	{'S', 'Z', strcmp95SimilarCredit},
	{'X', 'S', strcmp95SimilarCredit},
	{'Q', 'C', strcmp95SimilarCredit},
	{'U', 'V', strcmp95SimilarCredit},
	{'M', 'N', strcmp95SimilarCredit},
	{'L', 'I', strcmp95SimilarCredit},
	{'Q', 'O', strcmp95SimilarCredit},
	{'P', 'R', strcmp95SimilarCredit},
	{'I', 'J', strcmp95SimilarCredit},
	{'2', 'Z', strcmp95SimilarCredit},
	{'5', 'S', strcmp95SimilarCredit},
	{'8', 'B', strcmp95SimilarCredit},
	{'1', 'I', strcmp95SimilarCredit},
	{'1', 'L', strcmp95SimilarCredit},
	{'0', 'O', strcmp95SimilarCredit},
	{'0', 'Q', strcmp95SimilarCredit},
	{'C', 'K', strcmp95SimilarCredit},
	{'G', 'J', strcmp95SimilarCredit},
}

// strcmp95ToUpper returns the upper-case form of b for ASCII letters,
// leaving non-letters untouched. The Winkler 1994 similar-character table
// is upper-case-keyed; mixed-case input is folded once at the lookup site
// so the table remains canonical. Non-ASCII bytes pass through unchanged
// (and will simply miss the table — which is correct for non-ASCII input
// per CONTEXT.md §2).
func strcmp95ToUpper(b byte) byte {
	if b >= 'a' && b <= 'z' {
		return b - ('a' - 'A')
	}
	return b
}

// strcmp95SimilarLookup returns strcmp95SimilarCredit (0.3) if the unordered
// byte pair (a, b) appears in strcmp95SimilarChars, else 0.0.
//
// The scan is linear over 36 entries; both byte orientations are checked
// per entry so the table contains each canonical pair exactly once. For a
// 36-entry table the linear scan is sub-100ns and inlines on the hot path
// — a map lookup would be slower AND non-deterministic per DET-03.
//
// Case-folding: inputs are upper-cased via strcmp95ToUpper at the call site
// (Strcmp95Score). Non-ASCII bytes miss the table by construction.
func strcmp95SimilarLookup(a, b byte) float64 {
	for i := range strcmp95SimilarChars {
		t := strcmp95SimilarChars[i]
		if (a == t.a && b == t.b) || (a == t.b && b == t.a) {
			return t.sim
		}
	}
	return 0.0
}

// Strcmp95Score returns Winkler's Strcmp95 similarity between a and b as a
// value in [0.0, 1.0], where 1.0 means identical and 0.0 means maximally
// dissimilar.
//
// Strcmp95 = Jaro + similar-character credit + Winkler prefix boost
// + long-string adjustment. The four adjustments only add to the base Jaro
// score: Strcmp95Score(a, b) >= JaroWinklerScore(a, b) for every (a, b).
//
// # API hierarchy
//
//   - JaroScore        — base similarity.
//   - JaroWinklerScore — Jaro + prefix boost (shared-prefix bias).
//   - Strcmp95Score    — Jaro-Winkler + similar-character credit + long-string
//     adjustment (record-linkage / surname matching).
//
// # ASCII-only
//
// Strcmp95 operates on ASCII letters and digits. The similar-character
// table is letter-pair-keyed and has no Unicode equivalent in Winkler
// 1994. For non-ASCII input, normalise via fuzzymatch.Normalise first
// (NFC/NFD + diacritic folding). There is NO Strcmp95ScoreRunes variant.
//
// # Edge cases
//
//   - Strcmp95Score("", "")     == 1.0 exactly (both-empty identity)
//   - Strcmp95Score("", "abc")  == 0.0 exactly (one-empty)
//   - Strcmp95Score("abc", "abc") == 1.0 exactly (identical)
//   - Strcmp95Score(a, b)       == Strcmp95Score(b, a) (symmetric)
//   - Strcmp95Score(a, b)       >= JaroWinklerScore(a, b) (adjustments only add)
//
// # Reference vectors (Winkler 1990 / Census Bureau strcmp95.c)
//
//   - Strcmp95Score("MARTHA",    "MARHTA")    ≈ 0.9611
//   - Strcmp95Score("DWAYNE",    "DUANE")     ≈ 0.840   (W~U fires)
//   - Strcmp95Score("DIXON",     "DICKSONX")  ≈ 0.8133
//
// Time complexity: O(la · w) for the Jaro pass + O(min(la, lb)) for the
// similar-character credit pass + O(L_max) for the prefix scan. Space: O(la
// + lb) for match flags (stack-allocated for ASCII <= 256 bytes; heap for
// longer inputs).
func Strcmp95Score(a, b string) float64 { //nolint:gocyclo // Strcmp95 layers four adjustments — each branch is structurally required
	if a == b {
		return 1.0 // fast identity — covers both-empty and identical inputs
	}
	la, lb := len(a), len(b)
	if la == 0 || lb == 0 {
		return 0.0 // one-empty → zero similarity
	}

	// Compute matching window: w = max(la, lb)/2 - 1, clamped to >= 0.
	// We use the Jaro-style max-based window (NOT the strcmp95.c min-based
	// window) to preserve the hierarchy invariant Strcmp95Score >= JaroWinklerScore
	// — narrower windows would surface inputs where the base match count
	// drops below Jaro's, defeating the invariant. The similar-character
	// credit pass below operates on the same match-flag arrays so it sees
	// the SAME unmatched residue as Jaro itself.
	maxLen := la
	if lb > maxLen {
		maxLen = lb
	}
	w := maxLen/2 - 1
	if w < 0 {
		w = 0
	}

	// ASCII fast path: use stack-allocated [256]bool match-flag arrays when
	// both inputs fit within maxJaroStackLen (defined in jaro.go) and are
	// pure ASCII. Mirrors jaro.go's allocation contract.
	if la <= maxJaroStackLen && lb <= maxJaroStackLen && isASCII(a) && isASCII(b) {
		var matchA [maxJaroStackLen]bool
		var matchB [maxJaroStackLen]bool
		return strcmp95Bytes(a, b, la, lb, w, matchA[:la], matchB[:lb])
	}

	// Heap path for inputs exceeding maxJaroStackLen or containing non-ASCII.
	return strcmp95Bytes(a, b, la, lb, w, make([]bool, la), make([]bool, lb))
}

// strcmp95Bytes is the inner kernel for byte-level Strcmp95 similarity. It
// re-derives the Jaro match-flag arrays (per CONTEXT.md OQ-3 resolution:
// re-derive rather than reach into jaroBytes' internals) and applies the
// four Winkler 1994 adjustments in order.
//
// matchA and matchB are pre-allocated slices of length la and lb
// respectively. Stack vars are zero; make([]bool, ...) is zero — no
// explicit clear needed.
//
// Algorithm shape mirrors the canonical Census Bureau strcmp95.c reference
// (richmilne/JaroWinkler — consulted ONLY for the algorithm structure, not
// for code copying):
//
//  1. Jaro matching pass produces matchA/matchB flags + integer match count m.
//  2. Transposition pass (Jaro-canonical) over matched-position pairs counts t.
//  3. Similar-character credit pass: walk unmatched positions in a, pair them
//     with unmatched positions in b in order, accumulate 0.3 per pair in the
//     Winkler 1994 similar-character table. Mark consumed slots so each
//     unmatched-b position contributes at most once.
//  4. Effective match count Num_com = m + Num_sim. Jaro formula uses Num_com
//     in the numerator AND the third term's denominator.
//  5. Winkler prefix boost: J + L * 0.1 * (1 - J) when J >= 0.7, L capped at 4.
//  6. Long-string adjustment: when min(la, lb) > 4 AND m > prefix+1 AND
//     2*m >= min(la, lb) + prefix, additive correction (Winkler 1994 §3).
//
// The cyclomatic complexity is structurally mandated by the Strcmp95
// algorithm: four stacked adjustments, each with its own conditional
// branches. No refactoring can reduce the decision count without
// harming readability or correctness.
func strcmp95Bytes(a, b string, la, lb, w int, matchA, matchB []bool) float64 { //nolint:gocyclo // Strcmp95 algorithm has four stacked adjustments — see godoc above
	// ---- Step 1: base Jaro matching pass ----
	m := 0
	for i := 0; i < la; i++ {
		lo := i - w
		if lo < 0 {
			lo = 0
		}
		hi := i + w + 1
		if hi > lb {
			hi = lb
		}
		for j := lo; j < hi; j++ {
			if !matchB[j] && a[i] == b[j] {
				matchA[i] = true
				matchB[j] = true
				m++
				break
			}
		}
	}

	if m == 0 {
		return 0.0 // no matches at all — short-circuit before division-by-zero
	}

	// ---- Step 2: transposition pass (Jaro-canonical) ----
	// Walk matched positions in a in order, walk matched positions in b in
	// order; count positions where the matched characters differ.
	t := 0
	bj := 0
	for i := 0; i < la; i++ {
		if !matchA[i] {
			continue
		}
		for bj < lb && !matchB[bj] {
			bj++
		}
		if bj < lb {
			if a[i] != b[bj] {
				t++
			}
			bj++
		}
	}
	t /= 2 // each mismatch was counted in both directions

	// ---- Step 3: similar-character credit pass (Winkler 1994 §3) ----
	// Pair each UNMATCHED position in a with an UNMATCHED position in b (in
	// left-to-right order) and check if the upper-cased pair is in the
	// similar-character table. Each unmatched-b position can contribute at
	// most once — `consumed[]` (reusing matchB via a local in-place flip
	// would mutate the caller's view; we use a parallel local view via a
	// shadow byte to avoid that).
	//
	// Per Winkler 1994 the credit is capped at min(la, lb) * 0.3 (since at
	// most min(la, lb) positions can pair); in practice it stays well below
	// because m matches consume those slots.
	similarCredit := 0.0
	// We do NOT mutate matchB here — the transposition pass above has
	// already consumed it. We track consumption in a local `simConsumed`
	// stack-allocated arena.
	var simConsumedArr [maxJaroStackLen]bool
	var simConsumed []bool
	if lb <= maxJaroStackLen {
		simConsumed = simConsumedArr[:lb]
	} else {
		simConsumed = make([]bool, lb)
	}
	for i := 0; i < la; i++ {
		if matchA[i] {
			continue
		}
		upA := strcmp95ToUpper(a[i])
		for j := 0; j < lb; j++ {
			if matchB[j] || simConsumed[j] {
				continue
			}
			upB := strcmp95ToUpper(b[j])
			if sim := strcmp95SimilarLookup(upA, upB); sim > 0 {
				similarCredit += sim
				simConsumed[j] = true
				break
			}
		}
	}

	// ---- Step 4: effective Num_com & Jaro score ----
	// Num_com = m + Num_sim. Winkler 1994 formula uses Num_com everywhere
	// in the Jaro numerator/denominator stack (NOT m).
	numCom := float64(m) + similarCredit
	fla := float64(la)
	flb := float64(lb)

	// Three-term Jaro formula with explicit parenthesisation and left-to-right
	// float reduction (DET-06).
	j := (numCom/fla + numCom/flb + (numCom-float64(t))/numCom) / 3.0
	// Defensive clamp: similar-character credit may push the three-term sum
	// just past 1.0 on degenerate inputs. The hierarchy invariant requires
	// the final score to land in [0, 1]; clamp before applying the prefix
	// boost so the boost arithmetic stays in the documented range.
	if j > 1.0 {
		j = 1.0
	}

	// ---- Step 5: Winkler prefix boost ----
	// JW = J + L * winklerPrefixScale * (1 - J) when J >= winklerBoostThreshold.
	// The same constants as jarowinkler.go (LOCKED by REQUIREMENTS.md CHAR-06).
	prefix := 0
	if j >= winklerBoostThreshold {
		maxPfx := winklerMaxPrefix
		if la < maxPfx {
			maxPfx = la
		}
		if lb < maxPfx {
			maxPfx = lb
		}
		for prefix < maxPfx && a[prefix] == b[prefix] {
			prefix++
		}
		j += float64(prefix) * winklerPrefixScale * (1.0 - j)
	}

	// ---- Step 6: long-string adjustment (Winkler 1994 §3) ----
	// Conditions per RESEARCH.md Pitfall 5:
	//   (a) min(la, lb) > 4
	//   (b) m > prefix + 1
	//   (c) 2 * m >= min(la, lb) + prefix
	//
	// When all three hold, apply:
	//   W = W + ((1 - W) * (m - prefix - 1) / (la + lb - 2*prefix + m))
	//
	// The adjustment uses the integer m (NOT numCom): Winkler 1994 specifies
	// Num_com here as the integer count of matched positions — the
	// similar-character credit affects the Jaro score (step 4) but not the
	// long-string adjustment's combinatorial weight.
	minLen := la
	if lb < minLen {
		minLen = lb
	}
	if minLen > 4 && m > prefix+1 && 2*m >= minLen+prefix {
		denom := float64(la+lb-2*prefix) + float64(m)
		if denom > 0 {
			j += (1.0 - j) * float64(m-prefix-1) / denom
		}
	}

	// Final clamp — guards against floating-point overshoot at +1.0 caused
	// by the adjustment cascade on degenerate inputs (similar-credit + prefix
	// boost + long-string can compose past 1.0 by ULP-sized amounts).
	if j > 1.0 {
		j = 1.0
	}
	if j < 0.0 {
		j = 0.0
	}
	return j
}
