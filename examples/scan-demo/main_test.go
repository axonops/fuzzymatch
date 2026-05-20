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

// main_test.go pins the scan-demo program's stdout byte-for-byte
// across runs and platforms. Mirror of the scorer-composition
// example's stdout golden pattern (os.Pipe redirect + line-by-line
// diff with strconv.Itoa line numbering).
//
// The committed `want` constant captures the exact bytes scan.Check
// produces on the three demonstration sections (within-group only,
// within+cross-group with DefaultConfig, all three suppression
// modes). Any drift in any of the following sources surfaces here:
//
//   - Scorer score values (algorithm internals change, Normalise
//     change, weight-normalisation change).
//   - scan.Check sort key — (Kind, NameA, NameB, GroupA, GroupB).
//   - Suppression composition predicate ordering or semantics.
//   - Cross-group threshold boost arithmetic.
//   - DefaultConfig field defaults.
//
// The same drift would ALSO surface in
// testdata/golden/scan-default.json (Plan 09-06's cross-platform
// determinism corpus). The two gates corroborate.
//
// Regeneration workflow: write `const want = ""` placeholder, run
// `go test`, copy the printed "got" lines into want, re-run — must
// pass. Every update must be deliberate and reviewed: the score
// values are evidence that scan.Check + DefaultScorer produce the
// documented Phase 9 outputs on this corpus.

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
// scan-demo example. Regenerate by running `go run .` from
// examples/scan-demo/ and pasting the output here. Any diff
// requires a deliberate, reviewed update.
const want = `Section 1: within-group only (default Scorer + DefaultConfig)

  Kind         NameA              NameB              GroupA     GroupB     Score
  ------------ ------------------ ------------------ ---------- ---------- ------
  WithinGroup  userId             user_id            login      login      1.0000
  (1 warning(s))

Section 2: within + cross-group (DefaultConfig, boost=0.05, identical-cross=false)

  Kind         NameA              NameB              GroupA     GroupB     Score
  ------------ ------------------ ------------------ ---------- ---------- ------
  WithinGroup  userId             user_id            login      login      1.0000
  (1 warning(s))

Section 3: suppression composition (SilenceLint + SuppressedPairs + cross-group identical default)

  Kind         NameA              NameB              GroupA     GroupB     Score
  ------------ ------------------ ------------------ ---------- ---------- ------
  (0 warning(s))
`

// TestExample_Output captures the example's stdout by redirecting
// os.Stdout to a pipe, running main(), and restoring os.Stdout. It
// compares the captured bytes to the committed `want` constant.
//
// The test is in package main (not main_test) so it can call main()
// directly without spawning a subprocess — this avoids requiring a Go
// toolchain at test time and keeps the test deterministic on all CI
// platforms (mirrors examples/scorer-composition/main_test.go).
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
