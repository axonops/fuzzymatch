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

// main_test.go pins the example program's stdout byte-for-byte across
// runs and platforms. Mirror of the identifier-similarity example's
// stdout golden pattern (os.Pipe redirect + line-by-line diff).
//
// The committed `want` constant captures the exact bytes that
// DefaultScorer + DefaultScorer-MinusDoubleMetaphone produce on the
// example's five input pairs. Any score drift (algorithm internals
// change, Normalise behaviour change, weight-normalisation change)
// will surface as a diff here AND as a diff in
// testdata/golden/scorer-default.json — the two gates corroborate.

package main

import (
	"bytes"
	"io"
	"os"
	"strconv"
	"strings"
	"testing"
)

// want is the committed, byte-stable expected stdout of the
// scorer-composition example. Regenerate by running `go run .` and
// pasting the output here. Any diff requires a deliberate, reviewed
// update — the composite scores are the manifest evidence that
// DefaultScorer and DefaultScorer-MinusDoubleMetaphone produce the
// documented Phase 8 outputs on this corpus.
const want = `Pair (a / b)                         Default     MinusDM        MDef        MMDM
--------------------------------------------------------------------------------
user_id / userId                      1.0000      1.0000        true        true
Smith / Schmidt                       0.3608      0.2330       false       false
customer / customers                  0.7745      0.7294       false       false
XMLParser / xml_parser                0.6739      0.6087       false       false
org_id / organisation_id              0.2911      0.3493       false       false
`

// TestExample_Output captures the example's stdout by redirecting
// os.Stdout to a pipe, running main(), and restoring os.Stdout. It
// compares the captured bytes to the committed `want` constant.
//
// The test is in package main (not main_test) so it can call main()
// directly without spawning a subprocess — this avoids requiring a Go
// toolchain at test time and keeps the test deterministic on all CI
// platforms (mirrors examples/identifier-similarity/main_test.go).
func TestExample_Output(t *testing.T) {
	origStdout := os.Stdout
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatalf("TestExample_Output: os.Pipe: %v", err)
	}
	os.Stdout = w
	defer func() { os.Stdout = origStdout }()

	main()

	if err := w.Close(); err != nil {
		t.Fatalf("TestExample_Output: pipe close: %v", err)
	}

	var buf bytes.Buffer
	if _, err := io.Copy(&buf, r); err != nil {
		t.Fatalf("TestExample_Output: io.Copy from pipe: %v", err)
	}
	if err := r.Close(); err != nil {
		t.Fatalf("TestExample_Output: pipe close (reader): %v", err)
	}

	got := buf.String()
	if got == want {
		return
	}
	gotLines := strings.Split(strings.TrimRight(got, "\n"), "\n")
	wantLines := strings.Split(strings.TrimRight(want, "\n"), "\n")
	if len(gotLines) != len(wantLines) {
		t.Errorf("TestExample_Output: line count mismatch (got %d, want %d).\nFull diff below:\n--- got ---\n%s\n--- want ---\n%s",
			len(gotLines), len(wantLines), got, want)
		return
	}
	var differing []string
	for i := range wantLines {
		if gotLines[i] != wantLines[i] {
			differing = append(differing,
				"  line "+strconv.Itoa(i+1)+":\n    got:  "+gotLines[i]+"\n    want: "+wantLines[i])
		}
	}
	t.Errorf("TestExample_Output: %d line(s) differ. Update the `want` constant if the change is intentional.\n%s",
		len(differing), strings.Join(differing, "\n"))
}
