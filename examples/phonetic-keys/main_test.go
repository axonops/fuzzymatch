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

// main_test.go pins the phonetic-keys example program's stdout byte-for-byte
// across runs and platforms. This extends the project-wide cross-platform
// determinism gate (verify-determinism / DET-02) to the runnable example:
// the phonetic codes for all four algorithms (Phase 7 Soundex, Double
// Metaphone, NYSIIS, MRA) on all twelve curated name samples, plus the
// MRACompare examples, must be identical on every platform in the CI matrix.
//
// TestExample_Output runs the example via a direct call to main() with stdout
// redirected to a bytes.Buffer, then compares the captured output to the
// committed `want` constant. The test fails if any byte differs.

package main

import (
	"bytes"
	"io"
	"os"
	"strconv"
	"strings"
	"testing"
)

// want is the committed, byte-stable expected stdout of the phonetic-keys
// example program. Regenerate by running `go run .` and pasting the output
// here. Any diff in this constant requires a deliberate, reviewed update.
const want = `Name         Soundex  DM-pri   DM-sec   NYSIIS   MRA
----------------------------------------------------
Robert       R163     RPRT     RPRT     RABAD    RBRT
Rupert       R163     RPRT     RPRT     RAPAD    RPRT
Tymczak      T522     TMSK     TMXK     TYNCSA   TYMCZK
Schmidt      S530     XMT      SMT      SCNAD    SCHMDT
Smith        S530     SM0      XMT      SNAT     SMTH
Catherine    C365     K0RN     KTRN     CATARA   CTHRN
Katherine    K365     K0RN     KTRN     CATARA   KTHRN
Brown        B650     PRN      PRN      BRAN     BRWN
Browne       B650     PRN      PRN      BRAN     BRWN
Pacheco      P220     PXK      PXK      PACAC    PCHC
Byrne        B650     PRN      PRN      BYRN     BYRN
Boern        B650     PRN      PRN      BARN     BRN

MRACompare examples:
  Byrne vs Boern: matched=true sim=5
  Smith vs Smyth: matched=true sim=5
  Catherine vs Katherine: matched=true sim=5
  Ad vs ZachariahMontgomery: matched=false sim=0
`

// TestExample_Output captures the example's stdout by redirecting os.Stdout
// to a pipe, running main(), and restoring os.Stdout. It then compares the
// captured bytes to the committed `want` constant.
//
// The test is in package main (not main_test) so it can call main() directly
// without spawning a subprocess — this avoids requiring a Go toolchain at
// test time and keeps the test deterministic on all CI platforms.
func TestExample_Output(t *testing.T) {
	// Redirect os.Stdout to a pipe so we can capture main()'s fmt.Printf output.
	// The deferred restore runs even if main() panics — without it, a panic
	// would leave os.Stdout pointing at the closed pipe writer and corrupt
	// all subsequent test output in the binary.
	origStdout := os.Stdout
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatalf("TestExample_Output: os.Pipe: %v", err)
	}
	os.Stdout = w
	defer func() { os.Stdout = origStdout }()

	// Run main() — all fmt.Printf calls land in the pipe writer.
	main()

	// Close the write end so the reader gets EOF.
	w.Close()

	// Drain the read end.
	var buf bytes.Buffer
	if _, err := io.Copy(&buf, r); err != nil {
		t.Fatalf("TestExample_Output: io.Copy from pipe: %v", err)
	}
	r.Close()

	got := buf.String()
	if got == want {
		return
	}
	// Report the diff line-by-line rather than as one wall of text. A code
	// regression touches only the affected row; a column-width change touches
	// the header. The line-by-line form makes failure diagnosis trivial without
	// forcing a re-read of the entire stdout.
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

// TestExample_ColumnWidths pins the table layout (header and separator widths)
// separately from the cell values, so a column-width change is diagnosed
// independently of a code regression. If both drift, the dedicated test
// makes it clear which is which.
func TestExample_ColumnWidths(t *testing.T) {
	wantLines := strings.Split(strings.TrimRight(want, "\n"), "\n")
	if len(wantLines) < 2 {
		t.Fatalf("TestExample_ColumnWidths: want has fewer than 2 lines — table structure broken")
	}
	header := wantLines[0]
	separator := wantLines[1]
	// Header and separator must be the same width (the separator is a row of
	// dashes spanning the table width).
	if len(header) != len(separator) {
		t.Errorf("TestExample_ColumnWidths: header width %d != separator width %d", len(header), len(separator))
	}
	// Separator should be all '-' characters.
	for i, c := range separator {
		if c != '-' {
			t.Errorf("TestExample_ColumnWidths: separator[%d] = %q; want '-'", i, c)
			break
		}
	}
}
