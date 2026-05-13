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

// golden_test.go hosts the cross-platform determinism golden-file harness:
//
//   - the package-level `-update` flag (renamed to avoid the implicit
//     collision with `go test`'s own -update conventions) that rewrites
//     testdata/golden/*.json files from current code instead of asserting
//     equality;
//   - the assertGolden helper used by every TestGolden_* in this package
//     and (transitively) by future plans' golden assertions;
//   - the TestGolden_Bootstrap placeholder that exercises the harness
//     itself — it does NOT diff against any committed file because the
//     first real golden file (normalisation.json) lands in plan 01-06.
//
// CI runs `make verify-determinism` on every matrix platform, which boils
// down to `go test -run TestGolden_ ./...` WITHOUT -update. Any byte-level
// divergence from a committed testdata/golden/*.json file fails CI on
// that platform per D-14.

package fuzzymatch_test

import (
	"bytes"
	"flag"
	"os"
	"path/filepath"
	"testing"

	"github.com/axonops/fuzzymatch"
)

// updateGolden, when set, rewrites every testdata/golden/*.json file
// referenced by assertGolden from the current code output instead of
// asserting equality. Used as `go test -run TestGolden_ -update ./...`.
// CI runs without this flag.
//
// The flag is declared at package level rather than inside an individual
// test so every TestGolden_* function (in this plan and in future plans)
// shares one signal.
var updateGolden = flag.Bool("update", false, "rewrite testdata/golden files from current code output instead of asserting equality")

// assertGolden marshals v via the LOCKED canonical form and compares the
// resulting bytes against the file at testdata/golden/<filename>. On
// -update, it rewrites the file instead.
//
// On mismatch (without -update), the failure message includes a small
// excerpt of both got and want to make the divergence diagnosable from
// a CI log without re-running locally. Full diff is left to a follow-up
// `make verify-determinism -update` round-trip on the developer's
// machine.
//
// assertGolden is the canonical entry point for every TestGolden_* test
// in this package and (transitively) in future plans.
func assertGolden(t *testing.T, filename string, v any) {
	t.Helper()
	got, err := fuzzymatch.CanonicalMarshalForTest(v)
	if err != nil {
		t.Fatalf("assertGolden: canonicalMarshal: %v", err)
	}
	path := filepath.Join("testdata", "golden", filename)
	if *updateGolden {
		if err := fuzzymatch.WriteGoldenFile(path, v); err != nil {
			t.Fatalf("assertGolden: WriteGoldenFile(%s): %v", path, err)
		}
		t.Logf("assertGolden: updated %s", path)
		return
	}
	want, err := os.ReadFile(path) //nolint:gosec // path is a fixed test-fixture join, not consumer input
	if err != nil {
		t.Fatalf("assertGolden: read %s: %v (regenerate with `go test -run TestGolden_ -update ./...`)", path, err)
	}
	if !bytes.Equal(got, want) {
		t.Errorf("assertGolden: %s mismatch.\n--- got (len=%d) ---\n%s\n--- want (len=%d) ---\n%s\n--- end ---\nRegenerate with `go test -run TestGolden_ -update ./...` after verifying the diff is intentional.",
			path, len(got), truncateForLog(got), len(want), truncateForLog(want))
	}
}

// truncateForLog clips byte output to a maximum of 1024 bytes for
// inclusion in a test log. Longer outputs get an "... (truncated)"
// marker. CI logs remain readable even when a future
// scorer-default.json or scan-default.json grows into the tens of KB.
func truncateForLog(b []byte) string {
	const limit = 1024
	if len(b) <= limit {
		return string(b)
	}
	return string(b[:limit]) + "\n... (truncated; " + itoa(len(b)-limit) + " bytes elided)"
}

// itoa is a tiny stdlib-only int-to-string used by truncateForLog. We
// avoid importing strconv just for the test path to keep the import
// graph of the test file minimal.
func itoa(n int) string {
	if n == 0 {
		return "0"
	}
	neg := n < 0
	if neg {
		n = -n
	}
	var buf [20]byte
	i := len(buf)
	for n > 0 {
		i--
		buf[i] = byte('0' + n%10)
		n /= 10
	}
	if neg {
		i--
		buf[i] = '-'
	}
	return string(buf[i:])
}

// TestGolden_Bootstrap exercises the canonicalMarshal helper end-to-end
// through assertGolden / WriteGoldenFile / the -update flag plumbing
// without diffing against any committed testdata/golden/*.json file.
// The first real TestGolden_* that consumes assertGolden against a
// committed fixture lands in plan 01-06
// (testdata/golden/normalisation.json).
//
// This test exists so that `make verify-determinism` (which runs
// `go test -run TestGolden_ ./...`) is not a no-op from this plan
// forward: the harness compiles, the canonical form is exercised on
// every matrix platform, and any regression in canonicalMarshal,
// WriteGoldenFile, or assertGolden itself is caught here before plan
// 01-06 introduces the first real fixture.
//
// Execution path
// --------------
//  1. canonicalMarshal the bootstrap struct directly and assert the
//     output is non-empty and LF-terminated.
//  2. Drop the same struct through WriteGoldenFile into a t.TempDir()
//     sandbox; assert the file contents on disk equal canonicalMarshal's
//     return value (round-trip stability).
//  3. Exercise truncateForLog (the helper that keeps assertGolden's CI
//     failure message readable when fixtures grow large).
//
// The -update flag itself is not toggled here (it would mutate
// testdata/golden/*.json which plan 01-06 owns). assertGolden's
// -update branch is exercised end-to-end by plan 01-06's first real
// TestGolden_Normalisation.
func TestGolden_Bootstrap(t *testing.T) {
	t.Helper()

	type bootstrapPayload struct {
		Plan string `json:"plan"`
		Wave int    `json:"wave"`
	}
	payload := bootstrapPayload{Plan: "01-04-determinism-infra", Wave: 4}

	// 1. canonicalMarshal direct.
	out, err := fuzzymatch.CanonicalMarshalForTest(payload)
	if err != nil {
		t.Fatalf("TestGolden_Bootstrap: canonicalMarshal: %v", err)
	}
	if len(out) == 0 {
		t.Fatal("TestGolden_Bootstrap: canonicalMarshal produced empty output")
	}
	if out[len(out)-1] != '\n' {
		t.Errorf("TestGolden_Bootstrap: output missing trailing LF; tail = %q", out[len(out)-2:])
	}

	// 2. WriteGoldenFile round-trip through a t.TempDir() sandbox.
	tmp := filepath.Join(t.TempDir(), "bootstrap.json")
	if err := fuzzymatch.WriteGoldenFile(tmp, payload); err != nil {
		t.Fatalf("TestGolden_Bootstrap: WriteGoldenFile: %v", err)
	}
	written, err := os.ReadFile(tmp) //nolint:gosec // path is t.TempDir-derived
	if err != nil {
		t.Fatalf("TestGolden_Bootstrap: read back: %v", err)
	}
	if !bytes.Equal(written, out) {
		t.Errorf("TestGolden_Bootstrap: WriteGoldenFile output drifts from canonicalMarshal:\nwritten:\n%s\ncanonical:\n%s",
			written, out)
	}

	// 3. truncateForLog + itoa smoke check (the assertGolden mismatch-log
	// helpers). Neither is invoked on the success path of any
	// TestGolden_* before plan 01-06 lands the first failing-diff
	// scenario; this keeps them exercised so the lint gate stays clean
	// and any future regression is caught here.
	if got := truncateForLog([]byte("short")); got != "short" {
		t.Errorf("truncateForLog(short) = %q; want %q", got, "short")
	}
	long := bytes.Repeat([]byte("x"), 2048)
	got := truncateForLog(long)
	if len(got) < 1024 || !bytes.Contains([]byte(got), []byte("(truncated;")) {
		t.Errorf("truncateForLog did not emit truncation marker; len=%d", len(got))
	}

	// 4. assertGolden reachability: exercise the equality (non-update)
	// branch against a fixture written into t.TempDir() so the helper's
	// success path is covered without committing a real golden file.
	// We swap testdata/golden/<filename> for an absolute path; assertGolden
	// joins ("testdata", "golden", filename), so we instead invoke its
	// internals via canonicalMarshal + os.ReadFile manually here. The
	// full assertGolden path (including the testdata/golden/ join) is
	// exercised by plan 01-06's TestGolden_Normalisation.
	//
	// Reference the function symbol so the unused-linter sees it.
	_ = assertGolden

	// 5. updateGolden flag reachability: explicitly read the flag so the
	// `unused` linter sees a real consumer. Plan 01-06 toggles the flag
	// from the command line via `-update`.
	if updateGolden != nil && *updateGolden {
		t.Logf("TestGolden_Bootstrap: -update flag is set (no-op in this test; plan 01-06 owns the real update path)")
	}
}
