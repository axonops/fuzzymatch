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

// Package-level sentinel errors for fuzzymatch. All errors are wrappable
// via fmt.Errorf("...: %w", ErrX) and discoverable by errors.Is /
// errors.As — never via string matching. See docs/requirements.md §6.4.
//
// Error message convention (per .claude/skills/go-coding-standards):
//
//   - Every message starts with the "fuzzymatch: " package prefix so
//     wrappers like fmt.Errorf("scorer: %w", err) produce readable
//     compositions ("scorer: fuzzymatch: invalid input").
//   - The text after the prefix is lowercase and carries no trailing
//     punctuation ('.', '!', or '?') so concatenation flows cleanly.
//   - Each sentinel is a flat errors.New value, not a typed struct;
//     richer per-item context can be added in Phase 9 if scan needs it.
//
// The four sentinels named here are the v1.x contract; additional
// sentinels (e.g. ErrInvalidThreshold for the Scorer, or
// ErrHammingLengthMismatch for the Hamming algorithm) land alongside
// the features that introduce them in later phases.

package fuzzymatch

import "errors"

// ErrInvalidInput indicates a caller-provided string fails an
// algorithm's documented input constraints — most commonly invalid
// UTF-8 on a rune-aware API, or a non-comparable embedded NUL byte
// where the algorithm rejects it.
//
// Most algorithms accept arbitrary bytes and do NOT return this error;
// the exceptions document their constraints in their own godoc.
//
// Discriminate via errors.Is(err, fuzzymatch.ErrInvalidInput); never
// match the error message string.
var ErrInvalidInput = errors.New("fuzzymatch: invalid input")

// ErrInvalidConfiguration indicates a Scorer or Extract option set is
// internally inconsistent — for example, a negative weight, a threshold
// outside [0.0, 1.0], an empty algorithm list, or a normalisation
// option combination the library forbids.
//
// Returned by the option-applying constructors in Phase 8 (Scorer) and
// Phase 10 (Extract). See docs/requirements.md §8.
var ErrInvalidConfiguration = errors.New("fuzzymatch: invalid configuration")

// ErrInvalidAlgorithm indicates an AlgoID parameter does not match any
// registered algorithm in the dispatch table. Returned from the
// package-internal dispatch helpers (Phase 8+) when called with an
// out-of-range AlgoID (e.g. AlgoID(999)) or with a catalogue AlgoID
// whose dispatch entry has not yet been populated.
//
// Consumers should call AlgoIDs() to discover the valid set rather
// than guessing.
var ErrInvalidAlgorithm = errors.New("fuzzymatch: invalid algorithm")

// ErrEmptyInput indicates BOTH input strings are empty at the boundary
// of an API that has no defined empty-empty behaviour. Individual
// algorithm score functions handle empty inputs per their per-algorithm
// specification in docs/requirements.md §7 and do NOT return this
// error — only higher-level APIs (Scorer, Extract) that require
// non-degenerate input may surface it.
var ErrEmptyInput = errors.New("fuzzymatch: empty input")
