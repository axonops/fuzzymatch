# Algorithm Catalogue

This document is the per-algorithm reference for fuzzymatch's 23-algorithm
catalogue. The authoritative formal specification — formulae, edge-case
contracts, complexity, primary-source citations, and acceptance reference
vectors — lives in [`docs/requirements.md`](requirements.md) §7. This
document expands on each algorithm with prose-friendly description, intended
use cases, and worked examples once the implementing phase lands.

Phase 1 ships the scaffold only. Phase 2 fills in the character-based
algorithms (§7.1); Phase 3 isolates Smith-Waterman-Gotoh due to the Gotoh
1982 erratum requiring EMBOSS / biopython cross-validation; Phase 4 fills
in q-gram / n-gram algorithms (§7.2); Phase 5 fills in token-based
algorithms (§7.3); Phase 7 fills in the phonetic algorithms (§7.4) under
specialised licence-screen discipline; Phase 2 closes with Ratcliff-
Obershelp (§7.5). Each algorithm's **Status** line is updated to "implemented
in vX.Y.Z" as the corresponding plan lands.

Cross-reference: every algorithm constant is enumerated in
[`algoid.go`](../algoid.go), and the dispatch table is sized to the
catalogue at compile time. The H2 anchors below match the algorithm's
canonical spelling (e.g. `#levenshtein`, `#damerau-levenshtein-osa`) so
the README catalogue table can deep-link.

## Levenshtein

- **Category:** character-based, edit distance
- **AlgoID constant:** `AlgoLevenshtein`
- **Primary source:** TBD — filled in by the implementing phase
  (see `docs/requirements.md` §7.1.1).
- **Status:** planned (Phase 2).
- **Cross-reference:** `docs/requirements.md` §7.1.1.

## Damerau-Levenshtein OSA

- **Category:** character-based, edit distance with transposition (Optimal
  String Alignment restriction)
- **AlgoID constant:** `AlgoDamerauLevenshteinOSA`
- **Primary source:** TBD — filled in by the implementing phase
  (see `docs/requirements.md` §7.1.2).
- **Status:** planned (Phase 2).
- **Cross-reference:** `docs/requirements.md` §7.1.2.

## Damerau-Levenshtein Full

- **Category:** character-based, edit distance with unrestricted
  transposition (Lowrance-Wagner formulation)
- **AlgoID constant:** `AlgoDamerauLevenshteinFull`
- **Primary source:** TBD — filled in by the implementing phase
  (see `docs/requirements.md` §7.1.3).
- **Status:** planned (Phase 2).
- **Cross-reference:** `docs/requirements.md` §7.1.3.

## Hamming

- **Category:** character-based, equal-length distance
- **AlgoID constant:** `AlgoHamming`
- **Primary source:** TBD — filled in by the implementing phase
  (see `docs/requirements.md` §7.1.4).
- **Status:** planned (Phase 2).
- **Cross-reference:** `docs/requirements.md` §7.1.4.

## Jaro

- **Category:** character-based, name-matching with positional tolerance
- **AlgoID constant:** `AlgoJaro`
- **Primary source:** TBD — filled in by the implementing phase
  (see `docs/requirements.md` §7.1.5).
- **Status:** planned (Phase 2).
- **Cross-reference:** `docs/requirements.md` §7.1.5.

## Jaro-Winkler

- **Category:** character-based, name-matching with prefix bonus
- **AlgoID constant:** `AlgoJaroWinkler`
- **Primary source:** TBD — filled in by the implementing phase
  (see `docs/requirements.md` §7.1.6).
- **Status:** planned (Phase 2).
- **Cross-reference:** `docs/requirements.md` §7.1.6.

## Strcmp95

- **Category:** character-based, refined Jaro-Winkler with similar-
  character credit and long-string bonus
- **AlgoID constant:** `AlgoStrcmp95`
- **Primary source:** TBD — filled in by the implementing phase
  (see `docs/requirements.md` §7.1.7).
- **Status:** planned (Phase 2).
- **Cross-reference:** `docs/requirements.md` §7.1.7.

## Smith-Waterman-Gotoh

- **Category:** character-based, local sequence alignment with affine gap
  penalty
- **AlgoID constant:** `AlgoSmithWatermanGotoh`
- **Primary source:** TBD — filled in by the implementing phase
  (see `docs/requirements.md` §7.1.8).
- **Status:** planned (Phase 3 — isolated due to the documented Gotoh
  1982 erratum requiring EMBOSS / biopython cross-validation).
- **Cross-reference:** `docs/requirements.md` §7.1.8.

## LCSStr

- **Category:** character-based, longest common substring
- **AlgoID constant:** `AlgoLCSStr`
- **Primary source:** TBD — filled in by the implementing phase
  (see `docs/requirements.md` §7.1.9).
- **Status:** planned (Phase 2).
- **Cross-reference:** `docs/requirements.md` §7.1.9.

## Q-Gram Jaccard

- **Category:** q-gram, set similarity (Jaccard index over q-gram sets)
- **AlgoID constant:** `AlgoQGramJaccard`
- **Primary source:** TBD — filled in by the implementing phase
  (see `docs/requirements.md` §7.2.1).
- **Status:** planned (Phase 4).
- **Cross-reference:** `docs/requirements.md` §7.2.1.

## Sørensen-Dice

- **Category:** q-gram, set similarity (Dice coefficient over q-gram sets)
- **AlgoID constant:** `AlgoSorensenDice`
- **Primary source:** TBD — filled in by the implementing phase
  (see `docs/requirements.md` §7.2.2).
- **Status:** planned (Phase 4).
- **Cross-reference:** `docs/requirements.md` §7.2.2.

## Cosine

- **Category:** q-gram, vector similarity (cosine of q-gram frequency vectors)
- **AlgoID constant:** `AlgoCosine`
- **Primary source:** TBD — filled in by the implementing phase
  (see `docs/requirements.md` §7.2.3).
- **Status:** planned (Phase 4).
- **Cross-reference:** `docs/requirements.md` §7.2.3.

## Tversky

- **Category:** q-gram, asymmetric set similarity (Tversky index)
- **AlgoID constant:** `AlgoTversky`
- **Primary source:** TBD — filled in by the implementing phase
  (see `docs/requirements.md` §7.2.4).
- **Status:** planned (Phase 4).
- **Cross-reference:** `docs/requirements.md` §7.2.4.

## Monge-Elkan

- **Category:** token-based, hybrid (uses an inner character-based metric)
- **AlgoID constant:** `AlgoMongeElkan`
- **Primary source:** TBD — filled in by the implementing phase
  (see `docs/requirements.md` §7.3.1).
- **Status:** planned (Phase 6).
- **Cross-reference:** `docs/requirements.md` §7.3.1.

## Token Sort Ratio

- **Category:** token-based, sort-and-compare
- **AlgoID constant:** `AlgoTokenSortRatio`
- **Primary source:** TBD — filled in by the implementing phase
  (see `docs/requirements.md` §7.3.2).
- **Status:** planned (Phase 6).
- **Cross-reference:** `docs/requirements.md` §7.3.2.

## Token Set Ratio

- **Category:** token-based, set-and-compare
- **AlgoID constant:** `AlgoTokenSetRatio`
- **Primary source:** TBD — filled in by the implementing phase
  (see `docs/requirements.md` §7.3.3).
- **Status:** planned (Phase 6).
- **Cross-reference:** `docs/requirements.md` §7.3.3.

## Partial Ratio

- **Category:** token-based, sliding-window
- **AlgoID constant:** `AlgoPartialRatio`
- **Primary source:** TBD — filled in by the implementing phase
  (see `docs/requirements.md` §7.3.4).
- **Status:** planned (Phase 6).
- **Cross-reference:** `docs/requirements.md` §7.3.4.

## Token Jaccard

- **Category:** token-based, set similarity (Jaccard index over token sets)
- **AlgoID constant:** `AlgoTokenJaccard`
- **Primary source:** TBD — filled in by the implementing phase
  (see `docs/requirements.md` §7.3.5).
- **Status:** planned (Phase 6).
- **Cross-reference:** `docs/requirements.md` §7.3.5.

## Soundex

- **Category:** phonetic, English (Russell/Knuth)
- **AlgoID constant:** `AlgoSoundex`
- **Primary source:** TBD — filled in by the implementing phase
  (see `docs/requirements.md` §7.4.1).
- **Status:** planned (Phase 7).
- **Cross-reference:** `docs/requirements.md` §7.4.1.

## Double Metaphone

- **Category:** phonetic, multi-language tolerant (Philips 2000)
- **AlgoID constant:** `AlgoDoubleMetaphone`
- **Primary source:** TBD — filled in by the implementing phase
  (see `docs/requirements.md` §7.4.2).
- **Status:** planned (Phase 7). Metaphone 3 (USP 7,440,941) is
  explicitly EXCLUDED; see `docs/faq.md` for the patent screen.
- **Cross-reference:** `docs/requirements.md` §7.4.2.

## NYSIIS

- **Category:** phonetic, English (Taft 1970)
- **AlgoID constant:** `AlgoNYSIIS`
- **Primary source:** TBD — filled in by the implementing phase
  (see `docs/requirements.md` §7.4.3). Taft 1970 (NY State Special
  Report No. 1) is the canonical citation; sourcing is acknowledged
  as a Phase-7 readiness risk.
- **Status:** planned (Phase 7).
- **Cross-reference:** `docs/requirements.md` §7.4.3.

## MRA

- **Category:** phonetic, English Match Rating Approach (NBS TN 943)
- **AlgoID constant:** `AlgoMRA`
- **Primary source:** TBD — filled in by the implementing phase
  (see `docs/requirements.md` §7.4.4).
- **Status:** planned (Phase 7).
- **Cross-reference:** `docs/requirements.md` §7.4.4.

## Ratcliff-Obershelp

- **Category:** gestalt, recursive longest-common-substring (Ratcliff-Metzener)
- **AlgoID constant:** `AlgoRatcliffObershelp`
- **Primary source:** TBD — filled in by the implementing phase
  (see `docs/requirements.md` §7.5.1).
- **Status:** planned (Phase 2 closing).
- **Cross-reference:** `docs/requirements.md` §7.5.1.
