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

// jarowinkler.go implements the Jaro-Winkler similarity for the fuzzymatch
// catalogue.
//
// Source: Winkler, W. E. (1990). "String comparator metrics and enhanced
// decision rules in the Fellegi-Sunter model of record linkage." Proceedings
// of the Section on Survey Research Methods, American Statistical Association:
// 354-359.
//
// # Formula
//
// Given two strings a and b, compute:
//
//  1. J = JaroScore(a, b)   (delegates to jaro.go)
//  2. If J < winklerBoostThreshold (0.7), return J unchanged.
//  3. L = length of the common prefix of a and b, capped at winklerMaxPrefix (4).
//  4. JW = J + float64(L) * winklerPrefixScale * (1.0 - J)
//
// The prefix boost is additive and non-negative when J >= 0.7, ensuring
// JW >= J for all boosted pairs.
//
// # Constants (Winkler 1990 p. 357, LOCKED for v1.x by REQUIREMENTS.md CHAR-06)
//
//   - winklerPrefixScale    = 0.1  (the "p" factor)
//   - winklerMaxPrefix      = 4    (the "L_max" cap)
//   - winklerBoostThreshold = 0.7  (boost gate)
//
// # Reference Vectors (Winkler 1990 p. 357)
//
// MARTHA / MARHTA:
//   - J = 0.9444…
//   - Prefix: M=M, A=A, R=R, T≠H → L = 3
//   - JW = 0.9444 + 3 * 0.1 * (1 - 0.9444) = 0.9611…
//
// DIXON / DICKSONX:
//   - J = 0.7667…
//   - Prefix: D=D, I=I, X≠C → L = 2
//   - JW = 0.7667 + 2 * 0.1 * (1 - 0.7667) = 0.8133…
//
// DWAYNE / DUANE:
//   - J ≈ 0.8222
//   - Prefix: D=D, W≠U → L = 1
//   - JW ≈ 0.8222 + 1 * 0.1 * (1 - 0.8222) ≈ 0.8400
//
// # Jaro-Winkler is NOT a metric
//
// Jaro-Winkler is NOT a metric (inherits the non-metric property of
// the underlying Jaro similarity). Triangle inequality does not hold.
//
// Because Jaro-Winkler is not a metric, no JaroWinklerDistance function
// is provided — the formula yields a similarity in [0.0, 1.0] directly.
//
// # Implementation discipline
//
//   - JaroWinklerScore delegates to JaroScore (plan 02-03); Jaro-Winkler is
//     a wrapper that adds only a constant-bounded prefix loop.
//   - Prefix loop operates on bytes (same level as JaroScore's byte path);
//     at most winklerMaxPrefix = 4 iterations — O(1) overhead over Jaro.
//   - NO math.Pow, math.Log, math.Exp, math.Sqrt, or math.FMA (DET-06).
//   - NO init() function; dispatch registered via var _ idiom (§13.5).
//   - NO map iteration on output paths (DET-03).
//   - NO goroutines, channels, or mutexes.
//   - Zero allocations on ASCII short inputs — the prefix loop adds no
//     allocations over JaroScore's 0-alloc ASCII fast path.
//   - Float reduction: J + float64(L) * winklerPrefixScale * (1.0 - J)
//     evaluated left-to-right with explicit parenthesisation (DET-06).

package fuzzymatch

// winklerPrefixScale is the prefix-bonus scale factor "p" from
// Winkler 1990 p. 357. The value 0.1 is the canonical default
// and is LOCKED for v1.x by REQUIREMENTS.md CHAR-06 + CONTEXT.md.
const winklerPrefixScale = 0.1

// winklerMaxPrefix is the maximum effective common-prefix length
// ("L_max" in Winkler 1990 p. 357). The value 4 is the canonical
// cap; longer common prefixes saturate at this value.
const winklerMaxPrefix = 4

// winklerBoostThreshold is the underlying-Jaro threshold below
// which the prefix bonus is NOT applied (Winkler 1990 p. 357).
// Pairs with J < 0.7 return JaroScore unchanged.
const winklerBoostThreshold = 0.7

// JaroWinklerScore returns the Jaro-Winkler similarity between a and b as a
// value in [0.0, 1.0], where 1.0 means identical and 0.0 means maximally
// dissimilar.
//
// JaroWinklerScore delegates to JaroScore (Jaro 1989) and then applies the
// Winkler 1990 prefix boost: when the underlying Jaro score is at least
// winklerBoostThreshold (0.7), the score is increased by a bonus proportional
// to the length of the common prefix (capped at winklerMaxPrefix = 4 characters)
// scaled by winklerPrefixScale (0.1).
//
// Edge cases:
//   - JaroWinklerScore("", "") == 1.0 exactly (both-empty identity)
//   - JaroWinklerScore("", "abc") == 0.0 exactly (one-empty)
//   - JaroWinklerScore("abc", "abc") == 1.0 exactly (identical strings)
//   - JaroWinklerScore(a, b) == JaroWinklerScore(b, a) (symmetric)
//
// Reference vectors (Winkler 1990 p. 357):
//   - JaroWinklerScore("MARTHA", "MARHTA") ≈ 0.9611
//   - JaroWinklerScore("DIXON", "DICKSONX") ≈ 0.8133
//   - JaroWinklerScore("DWAYNE", "DUANE") ≈ 0.8400
//
// This function operates on bytes. For multi-byte UTF-8 inputs, use
// JaroWinklerScoreRunes to obtain the rune-aware similarity.
func JaroWinklerScore(a, b string) float64 {
	j := JaroScore(a, b)
	if j < winklerBoostThreshold {
		return j // boost gate: J < 0.7 → no prefix bonus
	}

	// Compute L = common-prefix length on bytes, capped at winklerMaxPrefix.
	// The prefix loop is O(1) — at most winklerMaxPrefix = 4 iterations.
	maxPfx := winklerMaxPrefix
	if len(a) < maxPfx {
		maxPfx = len(a)
	}
	if len(b) < maxPfx {
		maxPfx = len(b)
	}
	l := 0
	for l < maxPfx && a[l] == b[l] {
		l++
	}

	// Jaro-Winkler formula: J + L*p*(1-J)
	// Left-to-right, explicit parenthesisation per DET-06.
	return j + float64(l)*winklerPrefixScale*(1.0-j)
}

// JaroWinklerScoreRunes returns the Jaro-Winkler similarity between a and b,
// treating each string as a sequence of Unicode code points (runes) rather
// than bytes.
//
// This function delegates to JaroScoreRunes for the underlying Jaro
// calculation, then applies the Winkler prefix boost on runes. The rune
// variant allocates two []rune slices (from the underlying JaroScoreRunes
// call) plus an additional []rune conversion to compute the common prefix on
// rune boundaries. For ASCII inputs, prefer JaroWinklerScore (zero
// allocations).
func JaroWinklerScoreRunes(a, b string) float64 {
	j := JaroScoreRunes(a, b)
	if j < winklerBoostThreshold {
		return j // boost gate: J < 0.7 → no prefix bonus
	}

	// Compute L = common-prefix length on runes, capped at winklerMaxPrefix.
	ra := []rune(a) // 1 alloc (for prefix comparison only)
	rb := []rune(b) // 1 alloc

	maxPfx := winklerMaxPrefix
	if len(ra) < maxPfx {
		maxPfx = len(ra)
	}
	if len(rb) < maxPfx {
		maxPfx = len(rb)
	}
	l := 0
	for l < maxPfx && ra[l] == rb[l] {
		l++
	}

	// Left-to-right float reduction per DET-06.
	return j + float64(l)*winklerPrefixScale*(1.0-j)
}
