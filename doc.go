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

// Package fuzzymatch is a standalone Go library for fuzzy string matching.
//
// fuzzymatch ships a pluggable catalogue of 23 string-similarity algorithms,
// a weighted composite Scorer, and an optional collection-scan sub-package.
// It is stdlib-only with a single curated runtime dependency — golang.org/x/text,
// used for Unicode NFC/NFD normalisation in [Normalise]. There is no cgo
// anywhere, and the library makes no use of goroutines, channels, or
// background work — every public function is pure and deterministic.
//
// The library has three layers:
//
//   - Algorithm functions: 23 string-similarity algorithms exposed as
//     standalone public functions. Consumers wanting just one algorithm
//     import this package and call directly.
//   - Scorer: composes any subset of the algorithms into a weighted
//     composite score with configurable threshold and normalisation.
//     Immutable after construction; safe for concurrent use.
//   - Scan sub-package (github.com/axonops/fuzzymatch/scan): a turnkey
//     collection-scan layer over the Scorer. Optional — the root package
//     has no dependency on scan.
//
// Every algorithm is implemented from its primary academic source with the
// citation inline at the top of the implementation file, mathematical
// invariants verified by property tests, canonical reference vectors drawn
// from the originating paper, and behaviour pinned by BDD scenarios.
// Output is byte-identical across linux/amd64, linux/arm64, darwin/arm64,
// and windows/amd64.
//
// The authoritative technical specification lives at docs/requirements.md
// in the repository; this godoc is a thin entry point. For per-algorithm
// detail see docs/algorithms.md; for the Scorer see docs/scorer.md; for the
// scan sub-package see docs/scan.md.
//
// Patent and licence hygiene: this library is Apache 2.0 and incorporates no
// GPL/LGPL-derived code. Patent-encumbered algorithms (notably Metaphone 3)
// are excluded by design.
package fuzzymatch

// Blank import: pin golang.org/x/text as a direct runtime dependency at the
// module's freeze point (plan 01-01). Subsequent plans wire concrete uses of
// golang.org/x/text/unicode/norm into the public Normalise API; until then
// this blank import keeps the require line direct (not // indirect) and
// makes `go mod tidy` idempotent. See .planning/phases/01-foundation-infrastructure/
// 01-01-module-bootstrap-PLAN.md and .planning/research/STACK.md for rationale.
import _ "golang.org/x/text/unicode/norm"
