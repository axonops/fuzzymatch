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

// example_test.go contains runnable godoc examples for the Phase 2
// character-based algorithms. Each ExampleXxx function appears on
// pkg.go.dev alongside the function it documents. Wave 2 plans append their
// ExampleXxx functions to this SAME file.
//
// The Output: blocks are verified byte-for-byte by `go test -run ExampleXxx
// ./...` — any drift in score computation causes a test failure.

package fuzzymatch_test

import (
	"fmt"

	"github.com/axonops/fuzzymatch"
)

// ExampleLevenshteinScore demonstrates the Levenshtein similarity on the
// canonical Wagner-Fischer 1974 reference pair. The score is
// 1 - 3/7 ≈ 0.5714 (distance 3, max length 7).
func ExampleLevenshteinScore() {
	fmt.Printf("%.4f\n", fuzzymatch.LevenshteinScore("kitten", "sitting"))
	// Output:
	// 0.5714
}

// ExampleHammingScore demonstrates the Hamming similarity. The first call uses
// the canonical Hamming 1950 reference pair (equal-length, score ≈ 0.5714).
// The second call demonstrates the LOCKED unequal-length policy: when len(a)
// != len(b), HammingDistance returns max(len(a), len(b)) (NOT zero) and
// HammingScore therefore returns 0.0 silently — no error, no panic. The "zero"
// is in the score normalisation 1 - max/max, not in the underlying distance.
func ExampleHammingScore() {
	// Equal-length: 3 mismatches in 7 positions → 1 - 3/7 ≈ 0.5714.
	fmt.Printf("%.4f\n", fuzzymatch.HammingScore("karolin", "kathrin"))
	// Unequal-length: HammingDistance("abc","ab") = max(3,2) = 3,
	// HammingScore = 1 - 3/max(3,2) = 1 - 3/3 = 0.0 (zero score, not zero distance).
	fmt.Printf("%.4f\n", fuzzymatch.HammingScore("abc", "ab"))
	// Output:
	// 0.5714
	// 0.0000
}

// ExampleJaroScore demonstrates the Jaro similarity on the canonical Winkler
// 1990 reference pair. The score is (6/6 + 6/6 + 5/6) / 3 ≈ 0.9444.
func ExampleJaroScore() {
	fmt.Printf("%.4f\n", fuzzymatch.JaroScore("MARTHA", "MARHTA"))
	// Output:
	// 0.9444
}

// ExampleDamerauLevenshteinOSAScore demonstrates the Damerau-Levenshtein OSA
// similarity. The first call shows the canonical transposition pair "ab"/"ba"
// (distance 1, score 0.5). The second call shows the discriminating vector
// "ca"/"abc" that distinguishes OSA from Full DL — OSA returns 0.0 (distance 3)
// while Full DL would return 0.3333 (distance 2) for the same pair.
func ExampleDamerauLevenshteinOSAScore() {
	fmt.Printf("%.4f\n", fuzzymatch.DamerauLevenshteinOSAScore("ab", "ba"))
	fmt.Printf("%.4f\n", fuzzymatch.DamerauLevenshteinOSAScore("ca", "abc"))
	// Output:
	// 0.5000
	// 0.0000
}

// ExampleDamerauLevenshteinFullScore demonstrates the Damerau-Levenshtein Full
// (Lowrance-Wagner 1975) similarity. The "ca"/"abc" pair demonstrates DL-Full's
// divergence from DL-OSA: DL-OSA returns 0.0000 (distance 3) for the same pair,
// while DL-Full returns 0.3333 (distance 2) because Full DL permits unrestricted
// transpositions with subsequent editing.
func ExampleDamerauLevenshteinFullScore() {
	// The "ca"/"abc" pair demonstrates DL-Full's divergence from DL-OSA:
	// DL-OSA returns 0.0000 (distance 3); DL-Full returns 0.3333 (distance 2).
	fmt.Printf("%.4f\n", fuzzymatch.DamerauLevenshteinFullScore("ca", "abc"))
	fmt.Printf("%.4f\n", fuzzymatch.DamerauLevenshteinFullScore("ab", "ba"))
	// Output:
	// 0.3333
	// 0.5000
}

// ExampleJaroWinklerScore demonstrates the Jaro-Winkler similarity on the
// canonical Winkler 1990 reference pair. The underlying Jaro score is 0.9444
// (MARTHA / MARHTA share a 3-char common prefix "MAR"); JW adds the prefix
// bonus: 0.9444 + 3 * 0.1 * (1 - 0.9444) ≈ 0.9611.
func ExampleJaroWinklerScore() {
	fmt.Printf("%.4f\n", fuzzymatch.JaroWinklerScore("MARTHA", "MARHTA"))
	// Output:
	// 0.9611
}

// ExampleSmithWatermanGotohScore demonstrates the SWG local-alignment
// similarity on a substring-containment pair. The shorter input is fully
// contained in the longer; the local alignment finds the full match, so the
// normalised score clamp(raw / min(len), 0, 1) = clamp(12.0 / 12, 0, 1) =
// 1.0000.
func ExampleSmithWatermanGotohScore() {
	fmt.Printf("%.4f\n", fuzzymatch.SmithWatermanGotohScore("http_request", "http_request_header_fields"))
	// Output:
	// 1.0000
}

// ExampleSmithWatermanGotohRawScore demonstrates the unclamped raw alignment
// score. For the same substring-containment pair, the raw score equals
// Match × min(len) = 1.0 × 12 = 12.0 (twelve match positions, no gap
// penalty). Contrast with the normalised *Score variant which clamps to 1.0.
func ExampleSmithWatermanGotohRawScore() {
	fmt.Printf("%.1f\n", fuzzymatch.SmithWatermanGotohRawScore("http_request", "http_request_header_fields"))
	// Output:
	// 12.0
}

// ExampleStrcmp95Score demonstrates Winkler's Strcmp95 similarity on the
// canonical Winkler 1990 reference pair. Strcmp95 layers four adjustments
// atop Jaro: similar-character credit (Winkler 1994 §3 table), Winkler
// prefix boost, and the long-string adjustment. MARTHA/MARHTA contains no
// similar-character pair (T/H is not in the table) but DOES trigger the
// long-string adjustment, so the score lifts above JaroWinkler's 0.9611.
func ExampleStrcmp95Score() {
	fmt.Printf("%.4f\n", fuzzymatch.Strcmp95Score("MARTHA", "MARHTA"))
	// Output:
	// 0.9676
}

// ExampleLongestCommonSubstring demonstrates the substring-returning surface
// of LCSStr. The canonical schema-similarity pair returns the full identifier
// "http_request" — the longest contiguous segment shared by both inputs.
func ExampleLongestCommonSubstring() {
	fmt.Println(fuzzymatch.LongestCommonSubstring("http_request", "http_request_header_fields"))
	// Output:
	// http_request
}

// ExampleLongestCommonSubstringRunes demonstrates the rune-aware variant on
// a multi-byte UTF-8 pair. The shared rune-substring is "caf" (3 runes); the
// byte path would return "caf" as well for these particular inputs but the
// rune variant guarantees rune-boundary alignment for any input.
func ExampleLongestCommonSubstringRunes() {
	fmt.Println(fuzzymatch.LongestCommonSubstringRunes("café", "cafe"))
	// Output:
	// caf
}

// ExampleLCSStrScore demonstrates the Sørensen-Dice-normalised LCSStr
// similarity score on the schema-similarity pair. The LCS is "http_request"
// of length 12; the score is 2·12 / (12 + 26) = 24/38 ≈ 0.6316.
func ExampleLCSStrScore() {
	fmt.Printf("%.4f\n", fuzzymatch.LCSStrScore("http_request", "http_request_header_fields"))
	// Output:
	// 0.6316
}

// ExampleLCSStrScoreRunes demonstrates the rune-path score variant on
// café/cafe. The rune LCS is "caf" of length 3; the score is 2·3 / (4 + 4) =
// 6/8 = 0.7500.
func ExampleLCSStrScoreRunes() {
	fmt.Printf("%.4f\n", fuzzymatch.LCSStrScoreRunes("café", "cafe"))
	// Output:
	// 0.7500
}

// ExampleRatcliffObershelpScore demonstrates the Ratcliff-Obershelp
// gestalt-pattern-matching similarity on the canonical Dr. Dobb's Journal
// 1988 reference pair WIKIMEDIA/WIKIMANIA. The recursive longest-common-
// substring decomposition matches 7 characters (W, I, K, I, M, the
// L-shared prefix); score = 2·7/(9+9) = 14/18 ≈ 0.7778. This matches
// Python difflib.SequenceMatcher(autojunk=False).ratio() byte-for-byte.
func ExampleRatcliffObershelpScore() {
	fmt.Printf("%.4f\n", fuzzymatch.RatcliffObershelpScore("WIKIMEDIA", "WIKIMANIA"))
	// Output:
	// 0.7778
}

// ExampleRatcliffObershelpScoreRunes demonstrates the rune-path variant on
// café/cafe. The rune-level matched substring is "caf" (3 runes); score =
// 2·3/(4+4) = 0.7500. The byte path would treat "café" as 5 bytes (due to
// UTF-8 encoding of é) and produce a different score; the rune variant
// guarantees rune-boundary alignment for any input.
func ExampleRatcliffObershelpScoreRunes() {
	fmt.Printf("%.4f\n", fuzzymatch.RatcliffObershelpScoreRunes("café", "cafe"))
	// Output:
	// 0.7500
}

// ExampleQGramJaccardScore demonstrates the Q-Gram Jaccard similarity on
// the canonical Ukkonen 1992 §3 worked-example pair. The bigram multisets
// are |QA|=3 ("AGCT") and |QB|=7 ("AGCTAGCT"); intersection = 3, union = 7,
// J = 3/7 ≈ 0.4286. This is the load-bearing primary-source reference
// vector for the q-gram tier.
func ExampleQGramJaccardScore() {
	fmt.Printf("%.4f\n", fuzzymatch.QGramJaccardScore("AGCT", "AGCTAGCT", 2))
	// Output:
	// 0.4286
}

// ExampleQGramJaccardScoreRunes demonstrates the rune-path variant on the
// café/cafe pair. The rune-bigram multisets are QA={"ca","af","fé"} and
// QB={"ca","af","fe"}; intersection = 2, union = 4, J = 2/4 = 0.5000. The
// byte path would split "é" mid-codepoint and produce a different score;
// the rune variant guarantees rune-boundary alignment.
func ExampleQGramJaccardScoreRunes() {
	fmt.Printf("%.4f\n", fuzzymatch.QGramJaccardScoreRunes("café", "cafe", 2))
	// Output:
	// 0.5000
}

// ExampleSorensenDiceScore demonstrates the Sørensen-Dice coefficient on
// the canonical NLP-textbook bigram pair. The bigram multisets are
// QA={"ni","ig","gh","ht"} and QB={"na","ac","ch","ht"}; intersection = 1
// (only "ht" shared); DSC = 2·1/(4+4) = 0.2500. Sørensen-Dice weights
// the intersection more heavily than Jaccard (DSC = 2·|∩|/(|QA|+|QB|)
// vs J = |∩|/|∪|), making it the canonical default for many fuzzy
// name-matching workloads.
func ExampleSorensenDiceScore() {
	fmt.Printf("%.4f\n", fuzzymatch.SorensenDiceScore("night", "nacht", 2))
	// Output:
	// 0.2500
}

// ExampleSorensenDiceScoreRunes demonstrates the rune-path variant on
// the café/cafe pair. The rune-bigram multisets are QA={"ca","af","fé"}
// and QB={"ca","af","fe"}; intersection = 2 ("ca" + "af");
// DSC = 2·2/(3+3) = 4/6 ≈ 0.6667. The byte path would split "é"
// mid-codepoint and produce a different score; the rune variant
// guarantees rune-boundary alignment.
func ExampleSorensenDiceScoreRunes() {
	fmt.Printf("%.4f\n", fuzzymatch.SorensenDiceScoreRunes("café", "cafe", 2))
	// Output:
	// 0.6667
}

// ExampleCosineScore demonstrates the Cosine similarity over q-gram
// frequency vectors on the load-bearing RV-C1 hand-derivation pair
// (Salton & McGill 1983 §4.1 eq. 4.4 p.121). The bigram multisets are
// QA={"ab":1,"bc":1} (‖A‖²=2) and QB={"ab":1,"bc":1,"cd":1} (‖B‖²=3);
// intersection (sorted) = ["ab","bc"]; dot = 2;
// cos = 2/(sqrt(2)·sqrt(3)) ≈ 0.8164965809277259 (the IEEE-754 actual
// from the factorised form is 1 ULP from the rational limit
// 2/sqrt(6) = 0.8164965809277261 — see cosine_test.go RV-C1 derivation
// for the full precision discussion). Printed to 16 digits to show the
// load-bearing precision.
func ExampleCosineScore() {
	fmt.Printf("%.16f\n", fuzzymatch.CosineScore("abc", "abcd", 2))
	// Output:
	// 0.8164965809277259
}

// ExampleCosineScoreRunes demonstrates the rune-path variant on the
// café/cafe pair (RV-C3). The rune-bigram multisets are
// QA={"ca","af","fé"} (‖A‖²=3) and QB={"ca","af","fe"} (‖B‖²=3);
// intersection (sorted byte-lex) = ["af","ca"]; dot = 2;
// cos = 2/(sqrt(3)·sqrt(3)) ≈ 0.6666666666666667 (the IEEE-754 actual
// from the factorised form is 1 ULP from the rational limit
// 2/3 = 0.6666666666666666 — same precision discussion as RV-C1). The
// byte path would split "é" mid-codepoint and yield a different score.
func ExampleCosineScoreRunes() {
	fmt.Printf("%.16f\n", fuzzymatch.CosineScoreRunes("café", "cafe", 2))
	// Output:
	// 0.6666666666666667
}

// ExampleTverskyScore demonstrates the Tversky asymmetric similarity
// over q-gram multisets. The example is structured per RESEARCH.md
// OQ-4 recommendation: BOTH a symmetric case (α=β=1.0 → Q-Gram Jaccard
// degeneracy) AND an asymmetric case (α=0.8, β=0.2 → input-swap
// produces different scores) appear in the Output block, illustrating
// the direction-sensitivity property inline.
//
// Symmetric case (α=β=1.0):
//   - QA = bigrams("abcd") = {ab:1, bc:1, cd:1}
//   - QB = bigrams("abce") = {ab:1, bc:1, ce:1}
//   - |A∩B| = 2; |A−B| = 1; |B−A| = 1
//   - T = 2/(2 + 1·1 + 1·1) = 2/4 = 0.5 (also equals
//     QGramJaccardScore("abcd", "abce", 2))
//
// Asymmetric case (α=0.8, β=0.2 — RV-T1 / RV-T2 from RESEARCH.md §2.4):
//   - TverskyScore("abcd", "abcdef", 2, 0.8, 0.2) = 0.8823529411764706
//     (RV-T1; |A∩B|=3, |A−B|=0, |B−A|=2 → 3/3.4)
//   - TverskyScore("abcdef", "abcd", 2, 0.8, 0.2) = 0.6521739130434783
//     (RV-T2; |A∩B|=3, |A−B|=2, |B−A|=0 → 3/4.6)
//   - The two scores differ because α weighs the FIRST argument's
//     residuals (|A−B|) more heavily than β weighs the SECOND
//     (|B−A|); swapping the inputs flips which residual carries the
//     larger weight.
func ExampleTverskyScore() {
	// Symmetric case: α=β=1.0 reduces to Q-Gram Jaccard.
	fmt.Printf("%.4f\n", fuzzymatch.TverskyScore("abcd", "abce", 2, 1.0, 1.0))
	// Asymmetric case: α≠β; swapping inputs produces different scores.
	fmt.Printf("%.4f\n", fuzzymatch.TverskyScore("abcd", "abcdef", 2, 0.8, 0.2))
	fmt.Printf("%.4f\n", fuzzymatch.TverskyScore("abcdef", "abcd", 2, 0.8, 0.2))
	// Output:
	// 0.5000
	// 0.8824
	// 0.6522
}

// ExampleTverskyScoreRunes demonstrates the rune-path variant on the
// café/cafe pair with α=β=0.5 (the Sørensen-Dice degeneracy).
//
//   - QA = rune-bigrams("café") = {"ca":1, "af":1, "fé":1}
//   - QB = rune-bigrams("cafe") = {"ca":1, "af":1, "fe":1}
//   - |A∩B| = 2 (ca + af); |A−B| = 1 (fé); |B−A| = 1 (fe)
//   - T = 2/(2 + 0.5·1 + 0.5·1) = 2/3 ≈ 0.6667 (also equals
//     SorensenDiceScoreRunes("café", "cafe", 2))
//
// The byte path would split "é" mid-codepoint and yield a different
// score; the rune variant guarantees rune-boundary alignment.
func ExampleTverskyScoreRunes() {
	fmt.Printf("%.4f\n", fuzzymatch.TverskyScoreRunes("café", "cafe", 2, 0.5, 0.5))
	// Output:
	// 0.6667
}

// ExampleTokenSortRatioScore demonstrates the Token Sort Ratio
// similarity on the canonical RapidFuzz-documentation reorder pair.
// Both sides tokenise to {fuzzy, wuzzy, was, a, bear}; sorting and
// joining gives identical strings "a bear fuzzy was wuzzy" on each
// side; the Indel-formula reduces to 2·n/(2n) = 1.0 (identity over
// the sorted-joined representation).
//
// Token Sort Ratio is the simplest of the three Indel-based ratios in
// the catalogue: TokenSetRatio (plan 06-02) extends it with a
// three-way max over intersection / diff-A / diff-B token reconstructions;
// PartialRatio (plan 06-03) drops tokenisation and applies the Indel
// formula over the best-aligned substring of the longer input.
func ExampleTokenSortRatioScore() {
	fmt.Printf("%.4f\n", fuzzymatch.TokenSortRatioScore("fuzzy wuzzy was a bear", "wuzzy fuzzy was a bear"))
	// Output:
	// 1.0000
}

// ExampleTokenSetRatioScore demonstrates the Token Set Ratio
// three-way max construction. Token Set Ratio differs from Token
// Sort Ratio in three ways: (a) it deduplicates tokens (set
// semantics, not multiset); (b) when the intersection is non-empty
// AND one token set is a subset of the other, it short-circuits to
// 1.0; (c) when neither set is a subset, it takes the max of three
// Indel ratios over the sorted intersection, intersection+diff_ab,
// and intersection+diff_ba string forms.
//
// The example below — "hello world" vs "world peace" — exercises the
// non-subset three-way max case where the third branch
// (intersection+diff_ab vs intersection+diff_ba) strictly dominates
// the first two. The intersection is {"world"}; diff_ab = {"hello"};
// diff_ba = {"peace"}. The third Indel ratio computes
// 2·LCS("world hello", "world peace") / (11+11) = 14/22 = 7/11.
//
// DEVIATION (locked in plan 06-02): TokenSetRatioScore returns 0.0
// (NOT 1.0) when either tokenised input is empty — bug-for-bug
// compatibility with RapidFuzz issue #110. Other tokenised
// algorithms in the catalogue (TokenJaccard, MongeElkan) follow the
// standard both-empty → 1.0 convention.
func ExampleTokenSetRatioScore() {
	fmt.Printf("%.4f\n", fuzzymatch.TokenSetRatioScore("hello world", "world peace"))
	// Output:
	// 0.6364
}

// ExamplePartialRatioScore demonstrates the Partial Ratio similarity
// (byte path) on the canonical RapidFuzz reference pair where the
// shorter string "YANKEES" matches the right-edge alignment of the
// longer string "NEW YORK YANKEES" in its Region 2 middle-window
// iteration — the substring "YANKEES" of "NEW YORK YANKEES" at
// position 9 has length 7 and is byte-identical to the shorter input,
// so indelRatio = 2·7/(7+7) = 1.0.
//
// Partial Ratio differs from Token Sort Ratio / Token Set Ratio in
// three ways: (a) it operates at the character level (NO tokenisation
// — no whitespace splitting, no lowercasing, no camelCase awareness);
// (b) it iterates over THREE regions of the longer string (left tail,
// middle, right tail) per the RapidFuzz reference implementation; and
// (c) the score is the maximum Indel-formula similarity across all
// alignments (NOT a single composition).
//
// The two RapidFuzz issue-110-style empty-set deviations of
// TokenSetRatio do NOT apply to PartialRatio — PartialRatio follows
// the catalogue's standard both-empty → 1.0 convention.
func ExamplePartialRatioScore() {
	fmt.Printf("%.4f\n", fuzzymatch.PartialRatioScore("YANKEES", "NEW YORK YANKEES"))
	// Output:
	// 1.0000
}

// ExamplePartialRatioScoreRunes demonstrates the rune-path variant on
// a multi-byte UTF-8 input pair. The shorter input "caf" (3 runes,
// 3 bytes) matches the leftmost 3 runes of "café" (4 runes, 5 bytes)
// — the rune path correctly aligns at the rune boundary and computes
// indelRatioRunes([c,a,f], [c,a,f]) = 1.0 in Region 2 at i=0.
//
// The byte-path equivalent would diverge because len([]byte("café"))=5
// while len([]byte("caf"))=3 — the byte path would split "é"
// mid-codepoint and produce a different score. For ASCII-only inputs
// the two paths agree; for any input containing multi-byte UTF-8
// sequences callers should prefer the rune path.
func ExamplePartialRatioScoreRunes() {
	fmt.Printf("%.4f\n", fuzzymatch.PartialRatioScoreRunes("café", "caf"))
	// Output:
	// 1.0000
}

// ExampleTokenJaccardScore demonstrates the TokenJaccard similarity on a
// canonical partial-overlap pair. Tokenise produces {alpha, beta, gamma}
// on the left and {beta, gamma, delta} on the right. The set
// intersection is {beta, gamma} (cardinality 2) and the union is {alpha,
// beta, gamma, delta} (cardinality 4), so J = 2/4 = 0.5.
//
// TokenJaccard uses SET semantics on the deduplicated token list —
// distinct from Phase 5's Q-Gram Jaccard (QGramJaccardScore) which uses
// MULTISET semantics over q-gram counts. RV-TJ3 ("a a b" vs "a b" → 1.0)
// is the keystone regression gate for this distinction: TokenJaccard
// collapses the duplicate "a" to a single set member; Q-Gram Jaccard at
// any q would count duplicate q-grams.
//
// Unlike TokenSetRatio (whose LOCKED RapidFuzz issue #110 deviation
// returns 0.0 for both-empty inputs), TokenJaccard follows the STANDARD
// catalogue both-empty → 1.0 convention.
func ExampleTokenJaccardScore() {
	fmt.Printf("%.4f\n", fuzzymatch.TokenJaccardScore("alpha beta gamma", "beta gamma delta"))
	// Output:
	// 0.5000
}

// ExampleMongeElkanScore demonstrates the ASYMMETRIC Monge-Elkan
// similarity (Monge & Elkan 1996 §3). For each token in A, take the
// maximum inner-metric similarity over every token in B, then average
// the per-token maxima. With a fixed inner metric the function is
// direction-sensitive — the same inputs in swapped order generally
// produce different scores.
//
// Symmetric case (the RV-ME1 canonical example):
//   - tokens(A) = ["user","create"]; tokens(B) = ["usr","creating"]
//   - max(JW(user,usr)=0.9333, JW(user,creating)=0.4167)      = 0.9333
//   - max(JW(create,usr)=0.5,  JW(create,creating)=0.8917)    = 0.8917
//   - ME(A, B) = (0.9333 + 0.8917) / 2                        ≈ 0.9125
//
// Asymmetric case (the RV-ME6 keystone — same tokens, swapped order):
//   - MongeElkanScore("alpha", "alpha beta gamma", AlgoLevenshtein) = 1.0
//     (single A-token matches one of three B-tokens exactly; max=1.0)
//   - MongeElkanScore("alpha beta gamma", "alpha", AlgoLevenshtein) ≈ 0.4667
//     (three A-tokens, each takes max over the single B-token; the
//     two non-matching tokens drag the mean down)
//   - 1.0 ≠ 0.4667 — the input swap with the same inner produces
//     direction-sensitive scores.
//
// The inner AlgoID MUST be one of the 14 permitted inner metrics (9
// character-tier + 4 q-gram + AlgoRatcliffObershelp). Passing
// AlgoMongeElkan, any token-tier AlgoID, or any phonetic AlgoID panics
// with a documented message — see MongeElkanScore's godoc for the full
// allow-list and the Phase 7 forward-compatibility note.
func ExampleMongeElkanScore() {
	opts := fuzzymatch.DefaultNormalisationOptions()
	// Symmetric-looking case (the RV-ME1 canonical pair).
	fmt.Printf("%.4f\n", fuzzymatch.MongeElkanScore("user create", "usr creating", fuzzymatch.AlgoJaroWinkler, opts))
	// Asymmetric direction-sensitivity (RV-ME4 / RV-ME6 keystone pair).
	fmt.Printf("%.4f\n", fuzzymatch.MongeElkanScore("alpha", "alpha beta gamma", fuzzymatch.AlgoLevenshtein, opts))
	fmt.Printf("%.4f\n", fuzzymatch.MongeElkanScore("alpha beta gamma", "alpha", fuzzymatch.AlgoLevenshtein, opts))
	// Output:
	// 0.9125
	// 1.0000
	// 0.4667
}

// ExampleMongeElkanScoreSymmetric demonstrates the SYMMETRIC variant —
// the arithmetic mean of MongeElkanScore in the two directions:
//
//	ME_sym(A, B, sim) = (ME(A, B, sim) + ME(B, A, sim)) / 2.0
//
// This is the variant bound to dispatch[AlgoMongeElkan] per
// CONTEXT.md §4 LOCKED — AlgoMongeElkan participates in the standard
// symmetric property-test set without exemption because the symmetric
// average is invariant under argument swap (the sum of two terms
// swapped is the same sum).
//
// For the RV-ME4 / RV-ME6 asymmetric pair: ME(A, B) = 1.0,
// ME(B, A) ≈ 0.4667, so ME_sym = (1.0 + 0.4667) / 2 ≈ 0.7333.
//
// Consumers using AlgoMongeElkan through the Scorer (Phase 8) get the
// symmetric variant transparently; consumers needing genuine
// asymmetric scoring call MongeElkanScore directly or use the Scorer
// option WithMongeElkanAlgorithm(weight, inner).
func ExampleMongeElkanScoreSymmetric() {
	opts := fuzzymatch.DefaultNormalisationOptions()
	fmt.Printf("%.4f\n", fuzzymatch.MongeElkanScoreSymmetric("alpha", "alpha beta gamma", fuzzymatch.AlgoLevenshtein, opts))
	fmt.Printf("%.4f\n", fuzzymatch.MongeElkanScoreSymmetric("alpha beta gamma", "alpha", fuzzymatch.AlgoLevenshtein, opts))
	// Output:
	// 0.7333
	// 0.7333
}

// ExampleSoundexCode demonstrates SoundexCode on the canonical Knuth TAOCP
// Vol. 3 §6.4 reference pair. "Robert" encodes as "R163". "Rupert" also
// encodes as "R163" — illustrating that two phonetically similar names share
// the same Soundex code.
func ExampleSoundexCode() {
	fmt.Println(fuzzymatch.SoundexCode("Robert"))
	fmt.Println(fuzzymatch.SoundexCode("Rupert"))
	// Output:
	// R163
	// R163
}

// ExampleSoundexScore demonstrates SoundexScore on the canonical Robert/Rupert
// pair. Both names encode as R163 → score 1.0. Smith encodes as S530 →
// no match with Robert → score 0.0.
func ExampleSoundexScore() {
	fmt.Printf("%.1f\n", fuzzymatch.SoundexScore("Robert", "Rupert"))
	fmt.Printf("%.1f\n", fuzzymatch.SoundexScore("Robert", "Smith"))
	// Output:
	// 1.0
	// 0.0
}

// ExampleDoubleMetaphoneKeys demonstrates DoubleMetaphoneKeys on the canonical
// Philips 2000 reference pair. "Schmidt" encodes as ("XMT", "SMT") — the
// Germanic SCH initial produces X (primary sh-sound) and S (Germanic secondary).
// "Smith" encodes as ("SM0", "XMT") — SM initial plus TH → theta "0" primary.
// The shared "XMT" key (Schmidt primary == Smith secondary) means these names
// are phonetically similar under Double Metaphone.
func ExampleDoubleMetaphoneKeys() {
	p, s := fuzzymatch.DoubleMetaphoneKeys("Schmidt")
	fmt.Printf("%s %s\n", p, s)
	p2, s2 := fuzzymatch.DoubleMetaphoneKeys("Smith")
	fmt.Printf("%s %s\n", p2, s2)
	// Output:
	// XMT SMT
	// SM0 XMT
}

// ExampleDoubleMetaphoneScore demonstrates DoubleMetaphoneScore on the canonical
// Schmidt/Smith pair. Schmidt primary "XMT" matches Smith secondary "XMT" →
// score 1.0. Schmidt and Garcia have no matching keys → score 0.0.
func ExampleDoubleMetaphoneScore() {
	fmt.Printf("%.1f\n", fuzzymatch.DoubleMetaphoneScore("Schmidt", "Smith"))
	fmt.Printf("%.1f\n", fuzzymatch.DoubleMetaphoneScore("Schmidt", "Garcia"))
	// Output:
	// 1.0
	// 0.0
}

// ExampleNYSIISCode demonstrates NYSIISCode on the canonical Brown/Browne pair
// from Knuth TAOCP Vol. 3 §6.4. Both names encode to "BRAN" because the
// original Taft-1970 6-char truncation rules produce the same code.
func ExampleNYSIISCode() {
	fmt.Println(fuzzymatch.NYSIISCode("Brown"))
	fmt.Println(fuzzymatch.NYSIISCode("Browne"))
	// Output:
	// BRAN
	// BRAN
}

// ExampleNYSIISScore demonstrates NYSIISScore on the canonical Brown/Browne pair.
// Both names encode to "BRAN" → score 1.0. Brown and Robert encode to different
// codes → score 0.0.
func ExampleNYSIISScore() {
	fmt.Printf("%.1f\n", fuzzymatch.NYSIISScore("Brown", "Browne"))
	fmt.Printf("%.1f\n", fuzzymatch.NYSIISScore("Brown", "Robert"))
	// Output:
	// 1.0
	// 0.0
}

// ExampleMRACode demonstrates MRACode on the canonical Byrne reference vector
// from NBS Tech Note 943 (Moore, Kuhns, Trefftzs, Montgomery 1977).
// "Byrne" encodes to "BYRN": B stays (leading), y→Y (consonant), r→R, n→N,
// e is dropped (non-leading vowel). "Kathrynoglin" demonstrates the first-3 +
// last-3 truncation gate (pre-truncation KTHRYNGLN len 9 > 6).
func ExampleMRACode() {
	fmt.Println(fuzzymatch.MRACode("Byrne"))
	fmt.Println(fuzzymatch.MRACode("Kathrynoglin"))
	// Output:
	// BYRN
	// KTHGLN
}

// ExampleMRACompare demonstrates the (bool, int) return shape of MRACompare —
// the only public function in the fuzzymatch catalogue with a non-float64 return.
// The bool is the NBS-943 match decision; the int is the raw 0-6 similarity
// counter. Smith/Smyth match (sim=5, threshold=3 for sum_len=9).
// Both-empty matches with sim=6 (no characters to eliminate; max_unmatched=0).
func ExampleMRACompare() {
	matched, sim := fuzzymatch.MRACompare("Smith", "Smyth")
	fmt.Printf("matched=%v sim=%d\n", matched, sim)
	matched2, sim2 := fuzzymatch.MRACompare("", "")
	fmt.Printf("matched=%v sim=%d\n", matched2, sim2)
	// Output:
	// matched=true sim=5
	// matched=true sim=6
}

// ExampleMRAScore demonstrates MRAScore, the binary 0.0/1.0 dispatch-table
// wrapper around MRACompare. Smith/Smyth match → 1.0. Ad/ZachariahMontgomery
// length-diff >= 3 auto-mismatch → 0.0.
func ExampleMRAScore() {
	fmt.Printf("%.1f\n", fuzzymatch.MRAScore("Smith", "Smyth"))
	fmt.Printf("%.1f\n", fuzzymatch.MRAScore("Ad", "ZachariahMontgomery"))
	// Output:
	// 1.0
	// 0.0
}
