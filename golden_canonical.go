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

// golden_canonical.go LOCKS the v1.x canonical byte form used by every
// fuzzymatch golden-file test. The contract is:
//
//   - encoding/json's MarshalIndent with prefix="" and indent="  " (two
//     spaces) for indentation.
//   - a single trailing "\n" (LF) line terminator. NO CRLF on any platform.
//   - UTF-8 with NO byte-order mark.
//   - the caller passes an already-deterministically-ordered Go value
//     (sorted slices of structs; explicit struct field declaration order).
//     This helper does NOT sort or normalise the input; it is the caller's
//     responsibility to feed it a deterministic value (per the
//     no-map-iteration rule in determinism-standards).
//
// Once a golden file is committed under testdata/golden/, the bytes
// produced by canonicalMarshal on the same input MUST remain identical
// across patch versions of Go, across all five supported platforms
// (linux/amd64, linux/arm64, darwin/amd64, darwin/arm64, windows/amd64),
// and across any future refactor of this file.
//
// ANY change that breaks this byte-for-byte stability requires a major
// version bump per docs/requirements.md §11.2 and a CHANGELOG entry.

package fuzzymatch

import (
	"encoding/json"
	"fmt"
	"os"
)

// canonicalMarshal serialises v to the LOCKED v1.x golden-file byte form:
// json.MarshalIndent with two-space indent plus a single trailing "\n".
//
// The helper performs NO ordering or normalisation of v: callers MUST pass
// an already-deterministically-ordered value (an explicit struct, or a
// pre-sorted slice). Passing a value whose serialisation depends on Go map
// iteration order will produce non-deterministic bytes across runs — see
// the no-map-iteration rule in .claude/skills/determinism-standards.
//
// canonicalMarshal uses only encoding/json's stable emitters (MarshalIndent
// has been byte-stable since Go 1.0). It does NOT touch math.X, does NOT
// invoke any goroutine, does NOT allocate beyond what encoding/json
// already allocates, and is safe for concurrent use.
//
// The returned bytes always end with exactly one "\n" character. The
// output is UTF-8 and contains no byte-order mark.
func canonicalMarshal(v any) ([]byte, error) {
	body, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		return nil, fmt.Errorf("fuzzymatch: canonicalMarshal: %w", err)
	}
	// Append the LF terminator. We allocate a single fresh slice (len(body)+1)
	// rather than appending in place, which keeps the helper free of any
	// accidental aliasing of the caller's storage.
	out := make([]byte, len(body)+1)
	copy(out, body)
	out[len(body)] = '\n'
	return out, nil
}

// writeGoldenFile serialises v via the LOCKED canonical form and writes the
// result to path. It is intended for test maintenance only — production
// code never invokes it. The `-update` flag in fuzzymatch's golden-file
// test harness (see golden_test.go) is the sole expected caller.
//
// The helper is unexported (Q14b mechanical refactor, Phase 8.5 Plan 15a)
// because no production consumer should ever call it; the external test
// package fuzzymatch_test reaches it through the test-only re-export
// var WriteGoldenFile = writeGoldenFile in export_test.go, which keeps
// the symbol visible to _test.go files without polluting the public API
// surface.
//
// The file is written with mode 0o644. Existing content at path is
// overwritten. No backup is taken; the test harness's `-update` workflow
// expects the caller to be operating inside a git checkout where the
// committed file is the source of truth.
//
// writeGoldenFile does not create parent directories — testdata/golden/
// is committed to the repository and is expected to exist before this
// helper is invoked.
func writeGoldenFile(path string, v any) error {
	bytes, err := canonicalMarshal(v)
	if err != nil {
		return err
	}
	// Golden files are committed test fixtures meant to be world-readable
	// alongside the source — 0o644 is intentional. gosec G306 is silenced
	// at the call site for that reason.
	if err := os.WriteFile(path, bytes, 0o644); err != nil { //nolint:gosec // G306: test fixture, world-readable by design
		return fmt.Errorf("fuzzymatch: writeGoldenFile: %w", err)
	}
	return nil
}
