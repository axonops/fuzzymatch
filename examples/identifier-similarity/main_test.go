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
// runs and platforms. This extends the project-wide cross-platform
// determinism gate (verify-determinism / DET-02) to the runnable example:
// the similarity scores for all six algorithms on all seven identifier
// pairs must be identical on every platform in the CI matrix.
//
// TestExample_Output runs the example via a direct call to main() with
// stdout redirected to a bytes.Buffer, then compares the captured output
// to the committed `want` constant. The test fails if any byte differs —
// a float formatting regression, algorithm drift, or pair reordering will
// be caught immediately.

package main

import (
	"bytes"
	"io"
	"os"
	"testing"
)

// want is the committed, byte-stable expected stdout of the example program.
// Regenerate by running `go run .` and pasting the output here.
// Any diff in this constant requires a deliberate, reviewed update.
const want = `Pair (a / b)                      Levenshtein       DL-OSA      DL-Full      Hamming         Jaro Jaro-Winkler
--------------------------------------------------------------------------------------------------------------
user_id / userId                       0.7143       0.7143       0.7143       0.0000       0.8492       0.9095
created_at / creationTimestamp         0.4118       0.4118       0.4118       0.0000       0.7152       0.8291
status / state                         0.6667       0.6667       0.6667       0.0000       0.8222       0.8933
email / e_mail                         0.8333       0.8333       0.8333       0.0000       0.9444       0.9500
org_id / organisation_id               0.4000       0.4000       0.4000       0.0000       0.6444       0.6444
latitude / longitude                   0.6667       0.6667       0.6667       0.0000       0.7500       0.7750
is_deleted / is_active                 0.4000       0.4000       0.4000       0.0000       0.6185       0.6185
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
	origStdout := os.Stdout
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatalf("TestExample_Output: os.Pipe: %v", err)
	}
	os.Stdout = w

	// Run main() — all fmt.Printf calls land in the pipe writer.
	main()

	// Close the write end so the reader gets EOF.
	w.Close()
	os.Stdout = origStdout

	// Drain the read end.
	var buf bytes.Buffer
	if _, err := io.Copy(&buf, r); err != nil {
		t.Fatalf("TestExample_Output: io.Copy from pipe: %v", err)
	}
	r.Close()

	got := buf.String()
	if got != want {
		t.Errorf("TestExample_Output: stdout mismatch.\n--- got ---\n%s\n--- want ---\n%s\n--- end ---\nUpdate the `want` constant if the change is intentional.", got, want)
	}
}
