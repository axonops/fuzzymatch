---
status: issues_found
agent: algorithm-licensing-reviewer
scope: entire codebase (phases 1-8)
reviewed: 2026-05-17T08:30:00Z
finding_counts:
  critical: 0
  important: 6
  improvement: 9
  total: 15
---

# Algorithm Licensing Review — Phases 1-8 Gate Review

## Executive Summary

The fuzzymatch codebase is in **good** licensing shape overall. No critical
findings — there are no GPL/LGPL-derived files, no patent-encumbered
algorithms shipped, the runtime-deps allowlist holds, every `.go` file
carries the Apache-2.0 header (CI verifies this: 165 files all pass), and
the Metaphone 3 exclusion is properly documented in `docs/faq.md` and
`docs/requirements.md` §4 with the patent number cited in `algoid.go`.

The findings cluster around **completeness and uniformity** of source-
origin attribution rather than substantive licence risk:

1. The NOTICE file does not yet enumerate the algorithm academic sources
   or the studied Go implementations as the
   `algorithm-licensing-standards` skill explicitly requires (it defers
   this to file-level statements, which themselves are inconsistent
   between Phase 2/3 files and Phase 4+ files).
2. Eight Phase 2 / Phase 3 algorithm files (Levenshtein, Damerau OSA,
   Damerau Full, Hamming, Jaro, Jaro-Winkler, LCSStr, SWG) lack the
   explicit "Source-Origin Statement" block (with the
   GPL/LGPL-provenance and code-copied-verbatim attestations) that
   Phase 4+ files include.
3. `docs/algorithms.md` has "Primary source: TBD — filled in by the
   implementing phase" for **all 23 algorithms** despite all 23 being
   shipped. This is the public-facing per-algorithm reference and
   should now carry the inline citations.
4. The token-tier algorithms (TokenSort, TokenSet, PartialRatio) and the
   Strcmp95 algorithm openly describe their implementations as "fresh
   transcription from RapidFuzz's ... Python source structure" /
   "structurally transcribed from RapidFuzz's MIT-licensed
   `fuzz_py.py::_partial_ratio_impl`". This is borderline against the
   skill's "Translate code line-by-line ... is forbidden" rule, but the
   wording and the open admission suggest the intent matches the
   allowed "performance-pattern inspiration" path. Worth a discuss-phase
   reconciliation to update the skill's wording or document the
   exception.

The Phase 9 (scan sub-package) gate is **GREEN** with these caveats: the
8 Source-Origin-Statement gaps should be backfilled before v1.0 ships,
and the NOTICE file enrichment is a small, high-leverage doc change.

---

## Detailed Findings

### [Important] NOTICE file does not list academic sources or studied Go implementations
- **File:** /Users/johnny/Development/fuzzymatch/NOTICE
- **Phase introduced:** Phase 1 (foundation)
- **Issue:** The skill text in `.claude/skills/algorithm-licensing-standards/SKILL.md` §"In the NOTICE file" specifies the NOTICE file MUST enumerate the academic sources (Levenshtein 1965, Damerau 1964, Jaro 1989, Winkler 1990, …) and the studied permissively-licensed Go ports. The current NOTICE states "This product does not currently incorporate any third-party source code" and defers all attribution to file headers — but the skill's exemplar NOTICE explicitly distinguishes academic-source acknowledgement (always required for shipped algorithms) from copyright-notice attribution (only when code is incorporated). Both axes are missing today: 23 algorithm academic sources are absent from NOTICE; reference Go ports (agnivade/levenshtein, xrash/smetrics, adrg/strutil, hbollon/go-edlib, jellyfish, oubiwann/metaphone, RapidFuzz, biopython, Python difflib) are absent from NOTICE.
- **Standard:** `.claude/skills/algorithm-licensing-standards/SKILL.md` §"Attribution Format" → "In the NOTICE file" (lines 88-113 of the skill).
- **Action:** Code fix (NOTICE update).
- **Rationale:** Apache-2.0 distribution discipline. Consumers reading NOTICE expect to see the academic-source attribution one place, not 23 places. Especially important when the algorithm-licensing-reviewer skill is the explicit reviewer authority — the skill's own exemplar should be reflected in the artefact.
- **Suggested fix:** Update NOTICE with two sections: (a) "Academic sources for similarity algorithms implemented in this library:" with the 22-25 primary-source citations (one per algorithm; q-gram tier shares Jaccard 1912; SWG cites both Smith-Waterman 1981 and Gotoh 1982); (b) "Reference Go implementations studied for cross-validation of test vectors (no code copied):" listing agnivade/levenshtein (MIT), xrash/smetrics (MIT), adrg/strutil (MIT), hbollon/go-edlib (MIT), RapidFuzz (MIT), oubiwann/metaphone (BSD-3), jellyfish (BSD-2), biopython (BSD-3), Python difflib (PSF). The SWI-Prolog public-domain double_metaphone.c mirror, U.S. Census Bureau strcmp95.c (public domain), and OpenRefine Strcmp95.java (Apache-2.0) also belong here.

### [Important] Eight Phase 2/3 algorithm files lack the explicit Source-Origin Statement block
- **File:** /Users/johnny/Development/fuzzymatch/levenshtein.go (and 7 others — damerau_osa.go, damerau_full.go, hamming.go, jaro.go, jarowinkler.go, lcsstr.go, swg.go)
- **Phase introduced:** Phase 2 (core character algorithms) and Phase 3 (SWG)
- **Issue:** The skill text shows every algorithm file's header block should explicitly attest: primary source; studied reference implementations + licences (or "none"); "no GPL/LGPL references consulted"; "no code copied". Phase 4+ files (cosine, double_metaphone, monge_elkan, mra, nysiis, partial_ratio, q_gram, qgram_jaccard, ratcliff_obershelp, sorensen_dice, soundex, strcmp95, token_indel, token_jaccard, token_set_ratio, token_sort_ratio, tversky) all carry this block ("Source-origin discipline" / "Source-Origin Statement" / "Source-origin statement"). Phase 2/3 files do cite the primary source inline ("Source: Levenshtein, V. I. (1965). …") but omit the explicit GPL/LGPL-none and code-copied-none attestations, and do not enumerate any studied Go implementations.
- **Standard:** `.claude/skills/algorithm-licensing-standards/SKILL.md` §"Attribution Format" → "In the algorithm file" + §"Source Origin Statement". The skill's exemplar Levenshtein block shows the expected wording.
- **Action:** Code fix (8 file header updates).
- **Rationale:** Uniformity. The licensing-reviewer's PR-review procedure (in the agent description) specifically looks for these statements; their absence is a BLOCKING finding under the agent's own protocol for PR review of Phase 2/3 algorithms. They were grandfathered in at Phase 2 era before the discipline was tightened; backfilling closes the gap before v1.0.
- **Suggested fix:** Add a "Source-origin discipline" block to each of the 8 files, matching the Phase 4+ style: primary source, cross-validation (or "none — hand-derived reference vectors"), GPL/LGPL provenance: none, code copied verbatim: none, and any MIT/BSD Go ports consulted-for-references-only. For Levenshtein specifically: agnivade/levenshtein (MIT), xrash/smetrics (MIT). For Jaro/Jaro-Winkler: smetrics (MIT). For SWG: biopython/EMBOSS (BSD) — already half-cited as the corpus generator's source. The remaining files (Damerau OSA/Full, Hamming, LCSStr) likely have no studied Go port and the block should explicitly state that.

### [Important] docs/algorithms.md primary-source citations are all "TBD"
- **File:** /Users/johnny/Development/fuzzymatch/docs/algorithms.md (lines 29, 39, 49, 58, 67, 76, 86, 96, 106, 115, 124, 133, 142, 151, 160, 169, 178, 187, 196, 205, 215, 226, 235)
- **Phase introduced:** Phase 1 (scaffolded as a stub) — should have been backfilled phase-by-phase
- **Issue:** Every algorithm section in docs/algorithms.md has `- **Primary source:** TBD — filled in by the implementing phase`. All 23 algorithms have shipped. The skill's §"Attribution Format" → "In docs/algorithms.md" says "Per-algorithm documentation lists the primary source and any studied references explicitly." That requirement is currently unmet for the public docs.
- **Standard:** `.claude/skills/algorithm-licensing-standards/SKILL.md` §"Attribution Format" → "In docs/algorithms.md".
- **Action:** Code fix (docs/algorithms.md backfill).
- **Rationale:** This is the consumer-facing per-algorithm reference; "TBD" defeats its purpose. The data exists in each algorithm file's godoc header — just needs to be transcribed.
- **Suggested fix:** Replace "TBD — filled in by the implementing phase" with the inline citation from each algorithm's `.go` file. Also update Status from "planned (Phase N)" to "implemented in vX.Y.Z" per the line-19 schema.

### [Important] Per-algorithm "Status:" still shows "planned (Phase N)" for all 23 algorithms in docs/algorithms.md
- **File:** /Users/johnny/Development/fuzzymatch/docs/algorithms.md (every `- **Status:** planned …` line)
- **Phase introduced:** Phase 1
- **Issue:** Adjacent to the TBD primary-source finding above. The introductory paragraph (line 17-19) says "Each algorithm's **Status** line is updated to 'implemented in vX.Y.Z' as the corresponding plan lands." That mechanical update has not happened for any of the 23 algorithms.
- **Standard:** `docs/algorithms.md` self-declared protocol (line 17-19).
- **Action:** Code fix (one-liner per algorithm).
- **Rationale:** Misleading public documentation: a reader of docs/algorithms.md would conclude no algorithm is implemented. The CHANGELOG and the actual `.go` files contradict this.
- **Suggested fix:** Rewrite each "Status" line to reflect actual implementation status (e.g. "implemented in v0.2.0").

### [Important] partial_ratio.go "structurally transcribed from RapidFuzz" claim warrants explicit reconciliation against the skill
- **File:** /Users/johnny/Development/fuzzymatch/partial_ratio.go:40-44, 185-189, 310 (and similar phrasing in token_sort_ratio.go:93-97, token_set_ratio.go:150-154)
- **Phase introduced:** Phase 6 (token-based algorithms)
- **Issue:** The file headers state "The three-region iteration pattern + s1_char_set early-skip pattern were transcribed structurally from RapidFuzz's MIT-licensed `fuzz_py.py::_partial_ratio_impl` … No code was copied — the implementation is a fresh Go transcription from the Python source's logical structure only." The variable name `s1_char_set` is identical to RapidFuzz's; the "Region 1 / Region 2 / Region 3" terminology mirrors the upstream comments. The skill at §"Fresh-Implementation Discipline" explicitly forbids "Translate code line-by-line from another implementation into Go" and "Derive variable names, comment phrasing, or structural decisions from another implementation". "Structural transcription" sits in the grey zone — possibly acceptable under the "performance-pattern inspiration" carve-out, but the file's own wording makes the carve-out application non-obvious.
- **Standard:** `.claude/skills/algorithm-licensing-standards/SKILL.md` §"Fresh-Implementation Discipline" (lines 53-71). The forbidden list says "Derive variable names, comment phrasing, or structural decisions from another implementation". The variable `s1_char_set` and the "Region 1/2/3" naming both qualify under a strict reading.
- **Action:** Discuss-phase needed.
- **Rationale:** Either (a) the skill should be tightened/clarified to permit "documented structural inspiration with variable-name preservation when the variable is a load-bearing semantic marker", or (b) the implementation should rename internal variables and comments to be Go-native before v1.0 ships. The MIT-licensed source means even verbatim copying would not be a licence violation — but it would violate the fresh-implementation discipline that AxonOps applies as a hygiene measure (stricter than licence law requires).
- **Suggested fix:** Either rename `s1_char_set` → `shorterByteSet`, `charSet` (already used in places); rename "Region 1/2/3" terminology to "left tail / centred / right tail" (which is already the prose phrasing); OR add a paragraph to the skill explicitly approving this pattern with the partial_ratio.go and token_*ratio.go cases as the named precedent. The latter is the lower-effort path and matches the codebase's actual practice.

### [Important] docs/prior-art-research.md references dlclark/metaphone3 multiple times without flagging the patent exclusion
- **File:** /Users/johnny/Development/fuzzymatch/docs/prior-art-research.md:90, 112, 212, 253, 360, 406
- **Phase introduced:** Phase 1 (research artefact)
- **Issue:** This is a historical research/survey document that lists `dlclark/metaphone3` (MIT) as one of the candidate Go phonetic implementations — including a positive description ("Most accurate; handles `EncodeVowels`, `EncodeExact` opts"). The same document discusses Metaphone 3 as "Most accurate" without noting that it has been explicitly EXCLUDED per the patent screen. A future contributor reading this document in isolation might conclude Metaphone 3 is on the roadmap.
- **Standard:** `.claude/skills/algorithm-licensing-standards/SKILL.md` §"The Metaphone 3 Precedent" (lines 132-138) — the precedent is meant to be visible at every decision point.
- **Action:** Code fix (docs edit).
- **Rationale:** Prevents accidental re-introduction. The exclusion is documented in docs/faq.md and docs/requirements.md §4, but a contributor working from the prior-art-research survey might miss it.
- **Suggested fix:** Add a single-sentence callout near each `dlclark/metaphone3` mention in docs/prior-art-research.md: "NOTE: Metaphone 3 is EXCLUDED from this catalogue per docs/requirements.md §4 (U.S. Patent 7,440,941). See docs/faq.md for the patent screen rationale." Or add a top-of-document banner noting that the document is a historical survey and the canonical exclusion list lives in §4 of the requirements doc.

### [Improvement] NOTICE file mentions a hypothetical future update protocol but doesn't note the standard's exemplar
- **File:** /Users/johnny/Development/fuzzymatch/NOTICE (paragraph starting "When third-party source IS incorporated...")
- **Phase introduced:** Phase 1
- **Issue:** The current NOTICE has a forward-looking paragraph anticipating future third-party incorporation, but the standard's exemplar NOTICE shows that the academic-source listing is NOT contingent on incorporation — it is the standard format for an Apache-2.0 library that re-implements published algorithms. The two interpretations of NOTICE are at odds: (a) NOTICE attributes only INCORPORATED third-party code (current file); (b) NOTICE acknowledges academic sources and consulted references whether or not code was incorporated (skill exemplar).
- **Standard:** `.claude/skills/algorithm-licensing-standards/SKILL.md` §"In the NOTICE file".
- **Action:** Skill clarification OR code fix — depending on which interpretation the project owner prefers.
- **Rationale:** The current NOTICE is internally consistent and Apache-2.0-compliant in the narrow legal sense (NOTICE-required attribution is only for incorporated code under Apache-2.0). The skill's expansive interpretation is stricter — it treats academic-source acknowledgement as an editorial obligation distinct from the legal NOTICE function. Worth pinning down before v1.0 documentation freezes.
- **Suggested fix:** If the project takes the legal-minimum interpretation, update the skill text to reflect that NOTICE is only for INCORPORATED code and move the academic-source listing to docs/algorithms.md (already partially scoped under finding 3 above). If the project takes the expansive interpretation, update NOTICE per finding 1 above.

### [Improvement] gen-token-ratio-cross-validation.py rapidfuzz pin is in script header; could be cross-referenced from NOTICE
- **File:** /Users/johnny/Development/fuzzymatch/scripts/gen-token-ratio-cross-validation.py
- **Phase introduced:** Phase 6
- **Issue:** The token-tier cross-validation depends on RapidFuzz 3.14.5 (pinned in the script with a runtime assertion). RapidFuzz is MIT-licensed and is not redistributed in the library — only its outputs are committed as a JSON corpus. The relationship is clean. But the NOTICE file does not mention the corpus origin, and consumers regenerating the corpus need to know the pinned version and the licence.
- **Standard:** None directly — this is a hygiene improvement.
- **Action:** Code fix (NOTICE addition).
- **Rationale:** Reproducibility and provenance trail. Combine with finding 1 (NOTICE enrichment).
- **Suggested fix:** Include in NOTICE: "Cross-validation corpora are generated from external implementations pinned in scripts/gen-*-cross-validation.py: RapidFuzz 3.14.5 (MIT), biopython (BSD-3-Clause), Python stdlib difflib (PSF), jellyfish 1.2.1 (BSD-2-Clause), metaphone 0.6 / oubiwann (BSD-3-Clause)."

### [Improvement] BDD test module (tests/bdd/go.sum) pulls godog dependency chain — chain includes Apache-2.0, MIT, BSD components; not summarised anywhere
- **File:** /Users/johnny/Development/fuzzymatch/tests/bdd/go.mod, /Users/johnny/Development/fuzzymatch/tests/bdd/go.sum
- **Phase introduced:** Phase 8 (Scorer phase first uses the BDD module per spec; earlier phases scaffolded it)
- **Issue:** The BDD module pulls godog (Apache-2.0), testify (MIT), goleak (MIT), gherkin/messages (MIT), hashicorp/go-immutable-radix + hashicorp/go-memdb + golang-lru (MPL-2.0 or BSD — needs check), gofrs/uuid, pflag (BSD-3), yaml.v3 (MIT/Apache-2.0). None of these is incorporated into the published library — they live in a separate go.mod. But there is no centralised acknowledgement of the dev-time dependency chain.
- **Standard:** None directly — Apache-2.0 NOTICE obligation does not extend to test-only dependencies that are not redistributed. Hygiene improvement.
- **Action:** Code fix (lightweight NOTICE-DEV.md or section in NOTICE under "Development-time dependencies").
- **Rationale:** Belt and braces. Audit-friendly. Consumers asking "what does fuzzymatch transitively depend on?" should be able to read a single document.
- **Suggested fix:** Either inline a "Development-time dependencies (not redistributed)" subsection in NOTICE, or create CONTRIBUTING.md → "Third-party dev-time dependencies" subsection with the BDD module dependency tree.

### [Improvement] Hashicorp dependencies in BDD go.sum should be licence-spot-checked
- **File:** /Users/johnny/Development/fuzzymatch/tests/bdd/go.sum (hashicorp/go-immutable-radix v1.3.1, hashicorp/go-memdb v1.3.4, hashicorp/golang-lru v0.5.4, hashicorp/go-uuid v1.0.2)
- **Phase introduced:** Phase 8 (transitive via godog)
- **Issue:** Hashicorp moved many of their libraries from MPL-2.0 to BSL/MPL hybrid in 2023-2024. For this project the licensing impact is zero (test-only, not redistributed), but it's worth a one-time confirmation that the pinned versions here are still under MPL-2.0 or earlier permissive terms.
- **Standard:** No direct standard — operational hygiene.
- **Action:** Discuss-phase needed (one-time confirmation, no code change required unless a version is non-permissively licensed).
- **Rationale:** Hashicorp's BSL transition is the highest-profile permissive→restrictive licence change in the Go ecosystem in recent years. A two-minute check now prevents a surprise later.
- **Suggested fix:** `go-immutable-radix` v1.3.1, `go-memdb` v1.3.4, `golang-lru` v0.5.4 are MPL-2.0 per their respective LICENSE files on GitHub at those tags. Document this in a one-line comment in tests/bdd/go.mod or a CONTRIBUTING.md addendum.

### [Improvement] doc.go's "Patent and licence hygiene" paragraph is two sentences — could be three
- **File:** /Users/johnny/Development/fuzzymatch/doc.go:48-50
- **Phase introduced:** Phase 1
- **Issue:** The packagelevel godoc has: "Patent and licence hygiene: this library is Apache 2.0 and incorporates no GPL/LGPL-derived code. Patent-encumbered algorithms (notably Metaphone 3) are excluded by design." This is correct but could note where to find the per-algorithm provenance trail (the file headers, NOTICE, docs/faq.md).
- **Standard:** None.
- **Action:** Code fix (one-sentence addition).
- **Rationale:** Discoverability. The pkg.go.dev rendering of doc.go is the consumer's first point of contact.
- **Suggested fix:** Append: "Per-algorithm source attribution, studied references, and licence provenance are documented in each implementation file's header block; the consolidated list is in the repository NOTICE file."

### [Improvement] q_gram.go has Source-Origin discipline header but the shared kernel could clarify its inheritance to four callers
- **File:** /Users/johnny/Development/fuzzymatch/q_gram.go:57-70
- **Phase introduced:** Phase 5
- **Issue:** The shared q-gram extraction helper has its own Source-Origin block (Ukkonen 1992). Good. Each consumer (qgram_jaccard.go, sorensen_dice.go, cosine.go, tversky.go) also has its own block citing its own primary source. The relationship is clean. But a one-line cross-reference in q_gram.go ("Consumed by: qgram_jaccard.go, sorensen_dice.go, cosine.go, tversky.go") would make the kernel's downstream surface obvious to a reviewer auditing the q-gram tier as a whole.
- **Standard:** None — uniformity nudge.
- **Action:** Code fix (one-line edit).
- **Rationale:** Aid auditability. Same applies to token_indel.go vis-à-vis its three consumers.
- **Suggested fix:** Add a "Consumed by:" line to each shared-kernel file.

### [Improvement] tests/bdd/steps/*.go don't carry source-origin context but probably should not need to
- **File:** /Users/johnny/Development/fuzzymatch/tests/bdd/steps/algorithms_steps.go, /Users/johnny/Development/fuzzymatch/tests/bdd/steps/scorer_steps.go
- **Phase introduced:** Phase 5 and 8
- **Issue:** Step definition files carry Apache-2.0 headers but no algorithm-source-attribution block. They don't need one — they exercise the public API only — but a one-line statement to that effect in the file header would prevent a future reviewer from flagging them. Pure hygiene.
- **Standard:** None.
- **Action:** No action recommended; documented for completeness.
- **Rationale:** Step definitions don't implement algorithms; they call them. No source-origin attestation is required.

### [Improvement] verify-license-headers.sh checks for the header string but doesn't verify "AxonOps" copyright line specifically
- **File:** /Users/johnny/Development/fuzzymatch/scripts/verify-license-headers.sh:17
- **Phase introduced:** Phase 1
- **Issue:** The script greps for the literal "Licensed under the Apache License, Version 2.0". This is sufficient to catch missing-header files but does not catch a file that has a third-party Apache-2.0 header with a different copyright line. Today this is a hypothetical concern — every file is freshly authored with the AxonOps copyright — but if a future contribution adds a file derived from another Apache-2.0 project, the script wouldn't flag the discrepancy.
- **Standard:** None — operational hardening.
- **Action:** Code fix (script enhancement) — low priority.
- **Rationale:** Defence in depth. The script is the first line of CI defence; tightening it slightly catches a class of provenance regressions.
- **Suggested fix:** Add a second grep for "Copyright 20XX AxonOps Limited" (where 20XX is a 4-digit year) — flag any file matching the Apache-2.0 header but NOT the AxonOps copyright as needing reviewer attention. Could be a separate exit code for soft-warn-but-not-fail to avoid breaking CI on legitimate third-party Apache-2.0 file additions.

### [Improvement] No mention of cosign / SBOM / attestation provenance in NOTICE
- **File:** /Users/johnny/Development/fuzzymatch/NOTICE
- **Phase introduced:** Phase 1 (release machinery scaffolded)
- **Issue:** Release artefacts will be signed with cosign keyless via GitHub OIDC + SBOM attached per the spec/STACK.md. NOTICE could include a brief pointer to the verification protocol — useful for downstream consumers running supply-chain hygiene checks.
- **Standard:** None — operational improvement.
- **Action:** Code fix (NOTICE addition) — low priority, can wait until first signed release.
- **Rationale:** Reduces consumer friction.
- **Suggested fix:** Add to NOTICE: "Release artefacts are signed using sigstore cosign (keyless, GitHub OIDC) and accompanied by an SPDX-JSON SBOM. See <release-verification-doc-link> for the verification protocol."

---

## What was verified clean

- **Apache-2.0 file headers:** 165 `.go` files, all carrying the AxonOps Apache-2.0 header. `verify-license-headers.sh` passes.
- **Runtime dependency allowlist:** root `go.mod` has exactly one curated runtime dep (`golang.org/x/text v0.37.0`) and the main module. `verify-no-runtime-deps.sh` passes. No transitive cgo. `CGO_ENABLED=0` builds are clean.
- **GPL/LGPL-derived code:** none. Searched every `.go` file for GPL/LGPL/AGPL/copyleft references; all hits are either "GPL/LGPL provenance: none" attestations (positive) or absent. Phase 4+ algorithm files explicitly attest "no GPL/LGPL references consulted".
- **Metaphone 3 (U.S. Patent 7,440,941):** explicitly EXCLUDED. Documented in docs/faq.md §"Why no Metaphone 3?", docs/requirements.md §4 (Out of Scope), algoid.go:163-164, doc.go:49 (package godoc). No code shipped. No imports of `github.com/dlclark/metaphone3` anywhere. The precedent is preserved.
- **Patent screen for shipped algorithms:** all 23 algorithms covered by published academic papers dated 1912-2000 with no associated patent (Soundex's 1918 patents expired 1935 and are noted in soundex.go). No active patent risk identified.
- **Python generator scripts:** all four (gen-phonetic-cross-validation.py, gen-ratcliff-obershelp-cross-validation.py, gen-swg-cross-validation.py, gen-token-ratio-cross-validation.py) carry the Apache-2.0 header.
- **BDD module isolation:** test-only dependencies (godog Apache-2.0, testify MIT, goleak MIT, cucumber/gherkin MIT, hashicorp libraries MPL-2.0, gofrs/uuid MIT, pflag BSD-3, yaml.v3 MIT/Apache-2.0) all live in tests/bdd/go.mod with a `replace` directive — never reach the root go.mod or the published library surface.
- **No testify in root tests:** verified by inspection — root test files explicitly note "Stdlib testing only — no testify in root tests".
- **Phonetic algorithm Source-Origin Statements:** all 4 phonetic algorithms (Soundex, Double Metaphone, NYSIIS, MRA) have explicit Source-Origin Statement blocks per the skill's strictest standard, including the GPL/LGPL-none + code-copied-none attestations + the patent screen acknowledgement. Phase 7 was implemented under the tightest discipline.
- **Strcmp95 public-domain reference (Census Bureau strcmp95.c):** correctly documented as public domain (U.S. Government work). The OpenRefine Strcmp95.java (Apache-2.0) is correctly cited for prose-level tie-breaks only.

---

## Phase 9 Gate Decision

**GREEN with caveats.** The scan sub-package implementation in Phase 9 can
proceed. The findings above are all backfill / documentation /
uniformity work on already-shipped algorithms — none block new algorithm
implementation in the scan layer because scan does not itself implement
new similarity algorithms.

Recommendation: backfill the 8 Phase 2/3 Source-Origin Statements
(Important finding 2), enrich the NOTICE file (Important finding 1),
and update docs/algorithms.md (Important findings 3 + 4) as a pre-v1.0
documentation sprint, ideally before the v0.6.x integration shakedown
in Phase 7 of the original roadmap (= integration with downstream
consumer axonops/audit).
