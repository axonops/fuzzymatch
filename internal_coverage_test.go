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

// internal_coverage_test.go is the `go test ./...`-time gate for the
// Floor-3 exported-coverage invariant (Phase 8.5 Q12a + Plan 11).
// `make coverage` already invokes scripts/verify-coverage-floors.sh,
// but a developer running plain `go test ./...` would not see Floor 3
// fail — this meta-test closes that gap so the invariant is enforced
// at the same gate that all other invariants live behind.
//
// Mechanism: generate a coverage profile against the root package
// only (the verify-exported-coverage helper at scripts/cmd/
// verify-exported-coverage/main.go is the canonical Plan-11 path
// for Floor 3 enforcement). The test then invokes the helper as a
// subprocess, forwarding its exit code.
//
// Build-tag guard
// ---------------
// The test is gated by the `coveragefloor` build tag so that
// `go test ./...` does not recursively self-invoke `go test ...
// -coverprofile=...` — that would re-trigger this very test and
// cause an infinite loop. The intended invocation is:
//
//	go test -tags coveragefloor -run TestInternalCoverageFloor ./...
//
// CI's `make check` target wires this into the standard pre-commit
// pipeline; the daily nightly run also includes it.

//go:build coveragefloor

package fuzzymatch_test

import (
	"bytes"
	"os"
	"os/exec"
	"path/filepath"
	"testing"
)

// TestInternalCoverageFloor enforces the Floor-3 exported-coverage
// invariant at `go test` time by invoking the canonical Plan-11
// verify-exported-coverage helper as a subprocess.
//
// On success: the helper exits 0 and the test passes silently.
// On failure: the helper's stderr (the list of offending exported
// symbols) is surfaced via t.Fatalf so a single CI run shows every
// offender without re-running.
func TestInternalCoverageFloor(t *testing.T) {
	// Resolve the repo root so the subprocess invocation works
	// regardless of the test's runtime cwd. testing.T does not give
	// us the working directory directly, but the test binary's cwd
	// is the package directory by Go convention; we walk one level
	// from the package up to the repo root if needed.
	repoRoot, err := os.Getwd()
	if err != nil {
		t.Fatalf("os.Getwd: %v", err)
	}

	// Generate a coverage profile against the root package only.
	// Using a t.TempDir-backed file lets t.Cleanup remove it
	// automatically when the test exits.
	tmpDir := t.TempDir()
	profilePath := filepath.Join(tmpDir, "cover.out")

	// `go test -coverprofile=<path> -covermode=atomic .` against the
	// root package — NOT `./...` — to avoid the recursive-self-test
	// loop that the build tag also defends against. The atomic mode
	// matches the canonical `make coverage` recipe so the helper
	// sees the same profile format it sees in CI.
	cov := exec.Command("go", "test", "-covermode=atomic", "-coverprofile="+profilePath, ".") // #nosec G204 — fixed arg list, no user input
	cov.Dir = repoRoot
	var covStderr bytes.Buffer
	cov.Stderr = &covStderr
	if err := cov.Run(); err != nil {
		t.Fatalf("go test -coverprofile: %v\nstderr:\n%s", err, covStderr.String())
	}

	// Invoke the Plan-11 helper via `go run` against the generated
	// profile. The helper's exit code is the test's exit signal.
	helper := exec.Command("go", "run", "./scripts/cmd/verify-exported-coverage", profilePath) // #nosec G204 — fixed arg list
	helper.Dir = repoRoot
	var helperStdout, helperStderr bytes.Buffer
	helper.Stdout = &helperStdout
	helper.Stderr = &helperStderr
	if err := helper.Run(); err != nil {
		t.Fatalf("verify-exported-coverage failed:\nstdout:\n%s\nstderr:\n%s\nerror: %v",
			helperStdout.String(), helperStderr.String(), err)
	}
}
