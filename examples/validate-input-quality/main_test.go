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

// main_test.go pins the validate-input-quality example program's
// stdout byte-for-byte across runs and platforms. Mirror of the
// scorer-composition example's stdout golden pattern (os.Pipe
// redirect + line-by-line diff).
//
// The committed `want` constant captures the exact bytes that the
// example produces on its four documented input pairs. Any drift in
// the Validate output shape, the WarnKind ordering, the AlgoID
// scope, or the DefaultScorer score will surface as a diff here.
//
// The test is in package main (not main_test) so it can call main()
// directly without spawning a subprocess.

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
// validate-input-quality example. Regenerate by running `go run .`
// and pasting the output here. Any diff requires a deliberate,
// reviewed update.
const want = `Pair: "user_id" / ""
  warnings: 7
    - EmptyInput (Any)
    - UnequalLength (Hamming)
    - NoTokensAfterNormalise (MongeElkan)
    - NoTokensAfterNormalise (TokenSortRatio)
    - NoTokensAfterNormalise (TokenSetRatio)
    - NoTokensAfterNormalise (PartialRatio)
    - NoTokensAfterNormalise (TokenJaccard)
  score:    0.0000

Pair: "abc" / "abcd"
  warnings: 1
    - UnequalLength (Hamming)
  score:    0.4764

Pair: "中文" / "日本語"
  warnings: 6
    - UnequalLength (Hamming)
    - AllNonASCIIDropped (Strcmp95)
    - AllNonASCIIDropped (Soundex)
    - AllNonASCIIDropped (DoubleMetaphone)
    - AllNonASCIIDropped (NYSIIS)
    - AllNonASCIIDropped (MRA)
  score:    0.0895

Pair: "user_id" / "userId"
  warnings: 1
    - UnequalLength (Hamming)
  score:    1.0000

`

// TestExample_Output captures the example's stdout by redirecting
// os.Stdout to a pipe, running main(), and restoring os.Stdout. It
// compares the captured bytes to the committed `want` constant.
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
