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

// golden_canonical_test.go pins the LOCKED v1.x golden-file canonical
// byte contract:
//
//   - struct-field declaration order is preserved on output
//   - exactly one trailing "\n" (LF) terminator; no CRLF
//   - no UTF-8 byte-order mark
//   - two-space indentation throughout
//   - stable across calls (no time, no randomness, no global state)
//   - no trailing whitespace on any line
//
// Any change that breaks any of these is a v1.x stability violation per
// docs/requirements.md §11.2.

package fuzzymatch_test

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/axonops/fuzzymatch"
)

// canonicalMarshalSample is a small fixture that exercises every byte
// property the contract pins. Field declaration order is intentional:
// Name first (single string), Tokens second (slice), Counts third
// (nested struct) — the JSON output must mirror this order exactly.
type canonicalMarshalSample struct {
	Name   string                       `json:"name"`
	Tokens []string                     `json:"tokens"`
	Counts canonicalMarshalSampleCounts `json:"counts"`
}

type canonicalMarshalSampleCounts struct {
	Alpha int `json:"alpha"`
	Beta  int `json:"beta"`
}

func newCanonicalMarshalSample() canonicalMarshalSample {
	return canonicalMarshalSample{
		Name:   "foo",
		Tokens: []string{"a", "b", "c"},
		Counts: canonicalMarshalSampleCounts{Alpha: 1, Beta: 2},
	}
}

func TestCanonicalMarshal_ProducesSortedStructOutput(t *testing.T) {
	t.Helper()
	got, err := fuzzymatch.CanonicalMarshalForTest(newCanonicalMarshalSample())
	if err != nil {
		t.Fatalf("canonicalMarshal returned error: %v", err)
	}
	// The output must contain "name" before "tokens" before "counts".
	idxName := bytes.Index(got, []byte(`"name"`))
	idxTokens := bytes.Index(got, []byte(`"tokens"`))
	idxCounts := bytes.Index(got, []byte(`"counts"`))
	if idxName < 0 || idxTokens < 0 || idxCounts < 0 {
		t.Fatalf("canonicalMarshal missing one of the expected fields; output:\n%s", got)
	}
	if idxName >= idxTokens || idxTokens >= idxCounts {
		t.Errorf("canonicalMarshal field order broken: name=%d tokens=%d counts=%d; want strictly ascending", idxName, idxTokens, idxCounts)
	}
	// Inside the nested counts struct, alpha must precede beta.
	idxAlpha := bytes.Index(got, []byte(`"alpha"`))
	idxBeta := bytes.Index(got, []byte(`"beta"`))
	if idxAlpha < 0 || idxBeta < 0 {
		t.Fatalf("canonicalMarshal missing nested fields; output:\n%s", got)
	}
	if idxAlpha >= idxBeta {
		t.Errorf("canonicalMarshal nested field order broken: alpha=%d beta=%d", idxAlpha, idxBeta)
	}
}

func TestCanonicalMarshal_TrailingNewline(t *testing.T) {
	t.Helper()
	got, err := fuzzymatch.CanonicalMarshalForTest(newCanonicalMarshalSample())
	if err != nil {
		t.Fatalf("canonicalMarshal returned error: %v", err)
	}
	if len(got) == 0 {
		t.Fatal("canonicalMarshal produced empty output")
	}
	if got[len(got)-1] != '\n' {
		t.Errorf("canonicalMarshal did not end with LF; last byte = 0x%02x", got[len(got)-1])
	}
	// No CRLF anywhere.
	if bytes.Contains(got, []byte("\r")) {
		t.Errorf("canonicalMarshal output contains a CR byte; got:\n%s", got)
	}
	// Exactly one trailing LF — the byte before the final LF must not
	// itself be LF (no blank-line-at-EOF).
	if len(got) >= 2 && got[len(got)-2] == '\n' {
		t.Errorf("canonicalMarshal output ends with an extra blank line; tail = %q", got[len(got)-3:])
	}
}

func TestCanonicalMarshal_NoBOM(t *testing.T) {
	t.Helper()
	got, err := fuzzymatch.CanonicalMarshalForTest(newCanonicalMarshalSample())
	if err != nil {
		t.Fatalf("canonicalMarshal returned error: %v", err)
	}
	if len(got) >= 3 && got[0] == 0xEF && got[1] == 0xBB && got[2] == 0xBF {
		t.Errorf("canonicalMarshal output starts with a UTF-8 BOM; first 3 bytes = %x", got[:3])
	}
}

func TestCanonicalMarshal_TwoSpaceIndent(t *testing.T) {
	t.Helper()
	got, err := fuzzymatch.CanonicalMarshalForTest(newCanonicalMarshalSample())
	if err != nil {
		t.Fatalf("canonicalMarshal returned error: %v", err)
	}
	// No tab characters anywhere.
	if bytes.Contains(got, []byte("\t")) {
		t.Errorf("canonicalMarshal output contains a tab character; got:\n%s", got)
	}
	// Every indented line uses a multiple of two leading spaces. Inspect
	// each line and assert: if it starts with a space, the leading-space
	// run length is divisible by 2.
	lines := bytes.Split(got, []byte("\n"))
	for i, line := range lines {
		if len(line) == 0 {
			continue
		}
		spaces := 0
		for _, c := range line {
			if c != ' ' {
				break
			}
			spaces++
		}
		if spaces > 0 && spaces%2 != 0 {
			t.Errorf("line %d has %d leading spaces (not a multiple of 2): %q", i+1, spaces, line)
		}
	}
	// At least one nested line must start with exactly two leading spaces
	// (the top-level field rows).
	hasTwoSpaceIndent := false
	for _, line := range lines {
		if len(line) >= 3 && line[0] == ' ' && line[1] == ' ' && line[2] != ' ' {
			hasTwoSpaceIndent = true
			break
		}
	}
	if !hasTwoSpaceIndent {
		t.Errorf("canonicalMarshal output has no two-space indent rows; got:\n%s", got)
	}
}

func TestCanonicalMarshal_StableAcrossCalls(t *testing.T) {
	t.Helper()
	v := newCanonicalMarshalSample()
	first, err := fuzzymatch.CanonicalMarshalForTest(v)
	if err != nil {
		t.Fatalf("first canonicalMarshal: %v", err)
	}
	second, err := fuzzymatch.CanonicalMarshalForTest(v)
	if err != nil {
		t.Fatalf("second canonicalMarshal: %v", err)
	}
	if !bytes.Equal(first, second) {
		t.Errorf("canonicalMarshal not stable across calls:\nfirst:\n%s\nsecond:\n%s", first, second)
	}
}

func TestCanonicalMarshal_NoTrailingWhitespace(t *testing.T) {
	t.Helper()
	got, err := fuzzymatch.CanonicalMarshalForTest(newCanonicalMarshalSample())
	if err != nil {
		t.Fatalf("canonicalMarshal returned error: %v", err)
	}
	// Split on LF; verify no line (except a possibly-empty final one)
	// ends with whitespace.
	lines := strings.Split(string(got), "\n")
	for i, line := range lines {
		if line == "" {
			continue
		}
		if strings.HasSuffix(line, " ") || strings.HasSuffix(line, "\t") {
			t.Errorf("line %d has trailing whitespace: %q", i+1, line)
		}
	}
}

func TestCanonicalMarshal_EndsExactlyWithLF(t *testing.T) {
	t.Helper()
	// Additional regression check: a simple input produces bytes ending
	// in `}\n` with no `\r` and no BOM. This is the most common consumer
	// expectation (the audit-event taxonomy use case).
	got, err := fuzzymatch.CanonicalMarshalForTest(map[string]int{"a": 1, "b": 2})
	if err != nil {
		t.Fatalf("canonicalMarshal map: %v", err)
	}
	if len(got) < 2 || got[len(got)-2] != '}' || got[len(got)-1] != '\n' {
		t.Errorf("canonicalMarshal map: expected trailing \"}\\n\"; got %q", got[len(got)-2:])
	}
}

// TestCanonicalMarshal_RejectsUnmarshalableValue covers the encoding/json
// error path in canonicalMarshal. A channel value cannot be marshalled,
// so json.MarshalIndent returns a *json.UnsupportedTypeError; the helper
// must wrap it with the "fuzzymatch: canonicalMarshal: …" prefix.
func TestCanonicalMarshal_RejectsUnmarshalableValue(t *testing.T) {
	t.Helper()
	_, err := fuzzymatch.CanonicalMarshalForTest(make(chan int))
	if err == nil {
		t.Fatal("canonicalMarshal: expected error for unmarshalable value, got nil")
	}
	if !strings.Contains(err.Error(), "fuzzymatch: canonicalMarshal:") {
		t.Errorf("canonicalMarshal: error message %q is missing the package prefix", err.Error())
	}
}

// TestWriteGoldenFile_RoundTrip covers WriteGoldenFile's happy path: the
// file written to disk must contain exactly the bytes canonicalMarshal
// would have returned. Sandboxed in t.TempDir() so no real golden file
// is touched.
func TestWriteGoldenFile_RoundTrip(t *testing.T) {
	t.Helper()
	payload := newCanonicalMarshalSample()
	tmp := t.TempDir()
	path := filepath.Join(tmp, "round-trip.json")
	if err := fuzzymatch.WriteGoldenFile(path, payload); err != nil {
		t.Fatalf("WriteGoldenFile: %v", err)
	}
	written, err := os.ReadFile(path) //nolint:gosec // path is t.TempDir-derived
	if err != nil {
		t.Fatalf("read back: %v", err)
	}
	want, err := fuzzymatch.CanonicalMarshalForTest(payload)
	if err != nil {
		t.Fatalf("canonicalMarshal: %v", err)
	}
	if !bytes.Equal(written, want) {
		t.Errorf("WriteGoldenFile drift from canonicalMarshal:\nwritten:\n%s\ncanonical:\n%s", written, want)
	}
}

// TestWriteGoldenFile_RejectsUnmarshalableValue covers WriteGoldenFile's
// error path when canonicalMarshal fails. The helper must surface the
// wrapped error without ever calling os.WriteFile.
func TestWriteGoldenFile_RejectsUnmarshalableValue(t *testing.T) {
	t.Helper()
	tmp := t.TempDir()
	path := filepath.Join(tmp, "should-not-exist.json")
	err := fuzzymatch.WriteGoldenFile(path, make(chan int))
	if err == nil {
		t.Fatal("WriteGoldenFile: expected error for unmarshalable value, got nil")
	}
	if !strings.Contains(err.Error(), "fuzzymatch: canonicalMarshal:") {
		t.Errorf("WriteGoldenFile: error message %q should wrap canonicalMarshal's error", err.Error())
	}
	// The file must not have been created.
	if _, statErr := os.Stat(path); statErr == nil {
		t.Errorf("WriteGoldenFile: file was created despite marshal failure at %s", path)
	}
}

// TestWriteGoldenFile_WriteFailureSurfacesError covers WriteGoldenFile's
// os.WriteFile error branch. We point it at a path inside a non-existent
// directory; os.WriteFile cannot create the parent and must return an
// error that the helper wraps under "fuzzymatch: WriteGoldenFile:".
func TestWriteGoldenFile_WriteFailureSurfacesError(t *testing.T) {
	t.Helper()
	tmp := t.TempDir()
	badPath := filepath.Join(tmp, "no-such-dir", "child.json")
	err := fuzzymatch.WriteGoldenFile(badPath, newCanonicalMarshalSample())
	if err == nil {
		t.Fatal("WriteGoldenFile: expected error writing to non-existent directory, got nil")
	}
	if !strings.Contains(err.Error(), "fuzzymatch: WriteGoldenFile:") {
		t.Errorf("WriteGoldenFile: error message %q is missing the wrapper prefix", err.Error())
	}
}
