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

// verify-exported-coverage is the AST-based exported-coverage gate that
// implements Phase 8.5 Q12a (locked 2026-05-17). It walks the root
// fuzzymatch package via go/parser.ParseDir, enumerates every exported
// top-level symbol (functions, types, vars, consts), cross-references
// the function set against the coverage profile produced by `go test
// -coverprofile=<path>`, and exits non-zero when:
//
//   - Any exported function has zero coverage rows in the profile
//     (Floor 3 baseline: existence-at-least-one-test); OR
//   - Any exported function has statement coverage < 90.0% (Floor 3
//     tightened semantics per Q12a).
//
// For non-function exported symbols (types, vars, consts) the helper
// applies a lighter check: it scans the package's *_test.go files (via
// go/parser, NOT via grep) and confirms each symbol appears in at least
// one test-file identifier reference. Test files are walked once and
// the union of referenced exported symbol names is cached. This avoids
// shelling out to `grep` and keeps the gate AST-precise.
//
// The helper is shelled out from scripts/verify-coverage-floors.sh as
// the Floor 3 implementation. It is ALSO consumable by the
// verify-llms-sync target (planned for Plan 14 / Q13) — both gates
// share the same "what is exported in this package?" question.
//
// Usage:
//
//	go run ./scripts/cmd/verify-exported-coverage <coverage-profile-path>
//
// Default profile path (when omitted) is "coverage.out".
//
// Exit codes:
//
//	0 — all exported funcs >= 90.0% coverage, all exported
//	    types/vars/consts referenced in at least one *_test.go.
//	1 — at least one exported func is uncovered or under-covered, or
//	    at least one exported non-func is unreferenced. Offenders are
//	    printed to stderr, one per line.
//	2 — script invocation error (malformed profile, missing files, etc.).
//
// Threat model (Phase 8.5 T-08.5-22): the helper trusts the coverage
// profile format produced by `go tool cover`. Malformed input is
// rejected with a parse error and exit code 2 rather than panicking;
// CI guarantees a well-formed profile from `go test -coverprofile`.
package main

import (
	"bufio"
	"errors"
	"flag"
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"io"
	"os"
	"os/exec"
	"sort"
	"strconv"
	"strings"

	"github.com/axonops/fuzzymatch/scripts/internal/astwalk"
)

// floorPercent is the per-exported-function statement-coverage floor
// per Q12a. Locked at 90.0; do not bump without updating
// docs/requirements.md §15.8.
const floorPercent = 90.0

// rootPackageDir is the directory containing the root fuzzymatch
// package (relative to the helper's CWD when invoked from the repo
// root). The Makefile / verify-coverage-floors.sh CHDIR into the repo
// root before invoking the helper.
const rootPackageDir = "."

// rootPackageName is the Go package name we accept. parser.ParseDir
// may surface multiple synthetic packages (e.g. "fuzzymatch_test" from
// the same directory); we only care about the production package.
const rootPackageName = "fuzzymatch"

// excludeTestsBDD is the prefix of the BDD sub-module which has its
// own go.mod and is therefore parsed separately. The Q12a AST walk
// scopes only the root module's .go files.
const excludeTestsBDD = "tests/bdd"

// main parses CLI args, runs the gate, and translates outcomes to
// shell-friendly exit codes.
func main() {
	flag.Usage = func() {
		fmt.Fprintf(os.Stderr,
			"usage: %s <coverage-profile-path>\n\n"+
				"Verifies Phase 8.5 Q12a Floor 3: every exported function in the\n"+
				"root fuzzymatch package has >= %.1f%% statement coverage AND every\n"+
				"exported type/var/const is referenced in at least one *_test.go.\n",
			os.Args[0], floorPercent)
	}
	flag.Parse()

	profile := "coverage.out"
	if flag.NArg() > 0 {
		profile = flag.Arg(0)
	}

	if err := run(profile, os.Stdout, os.Stderr); err != nil {
		fmt.Fprintln(os.Stderr, err) //nolint:errcheck // best-effort stderr emission; exit code is the canonical signal
		// Exit 2 for invocation errors (missing files, malformed
		// profile); exit 1 for floor violations (run returns a
		// floorViolationErr in that case). Use errors.As (not type
		// assertion) so wrapped sentinels still classify correctly.
		var fve *floorViolationErr
		if errors.As(err, &fve) {
			os.Exit(1)
		}
		os.Exit(2)
	}
}

// floorViolationErr signals one or more Floor-3 violations. It is
// distinguished from other errors so main() can map to exit code 1
// (floor) vs 2 (invocation).
type floorViolationErr struct {
	count int
	msg   string
}

func (e *floorViolationErr) Error() string { return e.msg }

// run is the testable entry point. It writes the OK / FAIL summary to
// out and detailed offender enumeration to errOut, returning a
// floorViolationErr on Floor-3 violations or a plain error on
// invocation problems (missing files, parse failures).
func run(profile string, out, errOut io.Writer) error { //nolint:gocyclo // canonical pipeline: stat profile → walk AST → load coverage → enumerate test refs → check Floor 3a + 3b → emit report; each stage is one branch
	if _, err := os.Stat(profile); err != nil {
		return fmt.Errorf("verify-exported-coverage: coverage profile not found: %s", profile)
	}

	exportedFuncs, exportedNonFuncs, err := collectExported(rootPackageDir)
	if err != nil {
		return fmt.Errorf("verify-exported-coverage: AST walk failed: %w", err)
	}
	if len(exportedFuncs) == 0 && len(exportedNonFuncs) == 0 {
		fmt.Fprintln(out, //nolint:errcheck // best-effort write to writer; caller already controls the writer choice
			"OK: verify-exported-coverage — no exported symbols in the root package (bootstrap state).")
		return nil
	}

	funcPercents, err := loadFuncCoverage(profile)
	if err != nil {
		return fmt.Errorf("verify-exported-coverage: %w", err)
	}

	testRefs, err := collectTestRefs(rootPackageDir)
	if err != nil {
		return fmt.Errorf("verify-exported-coverage: test-file scan failed: %w", err)
	}

	var offenders []string

	// Floor 3a: every exported func has a coverage row AND >= 90.0%.
	for _, fn := range exportedFuncs {
		pct, ok := funcPercents[fn]
		if !ok {
			offenders = append(offenders, fmt.Sprintf("  %s — no coverage row (no test exercises it)", fn))
			continue
		}
		if pct < floorPercent {
			offenders = append(offenders, fmt.Sprintf("  %s — %.1f%% < %.1f%%", fn, pct, floorPercent))
		}
	}

	// Floor 3b: every exported non-func is referenced in at least one
	// _test.go file's AST.
	for _, sym := range exportedNonFuncs {
		if _, ok := testRefs[sym]; !ok {
			offenders = append(offenders, fmt.Sprintf("  %s — exported type/var/const not referenced in any *_test.go", sym))
		}
	}

	if len(offenders) > 0 {
		sort.Strings(offenders)
		fmt.Fprintf(errOut, //nolint:errcheck // best-effort write to writer; failure is non-actionable in CI context
			"verify-exported-coverage: FAIL — %d exported symbol(s) below Floor 3:\n%s\n\n"+
				"Floor 3 (Phase 8.5 Q12a): every exported function in the root package\n"+
				"must have >= %.1f%% statement coverage; every exported type/var/const\n"+
				"must be referenced in at least one *_test.go file.\n",
			len(offenders), strings.Join(offenders, "\n"), floorPercent)
		return &floorViolationErr{
			count: len(offenders),
			msg:   fmt.Sprintf("%d Floor-3 violation(s)", len(offenders)),
		}
	}

	fmt.Fprintf(out, //nolint:errcheck // best-effort write to writer; failure is non-actionable in CI context
		"OK: verify-exported-coverage — %d exported func(s) >= %.1f%%; %d exported type/var/const all referenced in *_test.go.\n",
		len(exportedFuncs), floorPercent, len(exportedNonFuncs))
	return nil
}

// collectExported delegates to the shared astwalk.CollectExported
// helper and returns two sorted slices for this gate's purposes:
// exportedFuncs (top-level FuncDecls; methods excluded because their
// coverage is rolled up to their receiver type and the per-method
// floor is intentionally NOT enforced separately) and exportedNonFuncs
// (TypeSpec / ValueSpec names — types, vars, consts).
//
// Methods returned by the shared helper are intentionally DROPPED here
// — see the FuncDecl branch in the AST walk for the per-method-floor
// rationale. The verify-llms-sync helper consumes the same shared
// helper but keeps the Methods slice.
//
// Test files (*_test.go) and the tests/bdd subtree are excluded from
// the walk by the shared helper. The caller expects to be invoked from
// the repo root.
func collectExported(dir string) (funcs, nonFuncs []string, err error) {
	res, err := astwalk.CollectExported(dir, rootPackageName)
	if err != nil {
		return nil, nil, err
	}
	return res.Funcs, res.NonFuncs, nil
}

// loadFuncCoverage invokes `go tool cover -func=<profile>` and parses
// the per-function coverage report. Each row matches the format:
//
//	<file>:<line>:\t<funcName>\t<percent>%
//
// The trailing row is `total:\t(statements)\t<percent>%` and is
// ignored. Method rows have funcName == "(Recv).Name" which is split
// on '.' — the method name (Name) is the key we record, matching how
// `go doc` emits the symbol.
//
// Funcs with multiple coverage rows (e.g. due to build-tag variants)
// are reduced to the MINIMUM percent — Floor 3a is the worst-case
// guarantee.
func loadFuncCoverage(profile string) (map[string]float64, error) {
	// gosec G204: command path is literal ("go"); profile is operator-
	// supplied via the -profile flag and is treated as a relative path
	// to a file the operator owns. The exec.Command invocation is the
	// canonical way to call `go tool cover -func=<profile>` and the
	// flag is whitelisted in CI scripts only — no untrusted input.
	cmd := exec.Command("go", "tool", "cover", "-func="+profile) //nolint:gosec // G204: profile is operator-supplied path; literal cmd "go"
	stdout, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("go tool cover -func=%s: %w", profile, err)
	}

	percents := make(map[string]float64)
	scanner := bufio.NewScanner(strings.NewReader(string(stdout)))
	for scanner.Scan() {
		line := scanner.Text()
		if strings.HasPrefix(line, "total:") {
			continue
		}
		// `go tool cover -func` produces tab-padded rows where the
		// padding is multiple tabs at each column boundary. Splitting
		// on any-whitespace runs (strings.Fields) collapses the
		// padding correctly. Expected three logical fields:
		//   [0] = file:line:           (the leading "file.go:NNN:" anchor)
		//   [1] = funcName             (the function or method name)
		//   [2] = "<percent>%"         (the trailing percent)
		fields := strings.Fields(line)
		if len(fields) < 3 {
			continue
		}
		funcField := fields[1]
		pctField := strings.TrimSuffix(fields[2], "%")
		pct, err := strconv.ParseFloat(pctField, 64)
		if err != nil {
			// Malformed row — skip rather than fail; CI guarantees
			// well-formed profiles. Logging here would produce false
			// signal during local dev.
			continue
		}

		// Method rows: "(Recv).Name" → extract Name.
		name := funcField
		if dot := strings.LastIndex(name, "."); dot >= 0 {
			name = name[dot+1:]
		}

		// Reduce duplicates to MIN — Floor 3a is worst-case.
		if existing, ok := percents[name]; !ok || pct < existing {
			percents[name] = pct
		}
	}
	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("scan go tool cover output: %w", err)
	}

	return percents, nil
}

// collectTestRefs walks every *_test.go file under dir via
// go/parser.ParseDir and returns the set of identifier names referenced
// anywhere in any test file's AST. This is the AST-precise alternative
// to grep for the Floor 3b check on exported non-func symbols.
//
// We use `parser.ParseFile` on each test file individually rather than
// the synthetic "fuzzymatch_test" pseudo-package returned by ParseDir,
// because test files in both the `fuzzymatch` and `fuzzymatch_test`
// packages must be inspected and ParseDir's package boundary is not
// the right primitive here.
func collectTestRefs(dir string) (map[string]struct{}, error) { //nolint:gocyclo // AST walk over *_test.go files with per-decl-type and per-identifier inspection; the branch count mirrors the AST node-type cases
	refs := make(map[string]struct{})

	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, fmt.Errorf("read dir %s: %w", dir, err)
	}

	fset := token.NewFileSet()
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		name := entry.Name()
		if !strings.HasSuffix(name, "_test.go") {
			continue
		}
		path := dir + string(os.PathSeparator) + name
		// Skip BDD test files (separate go.mod).
		if strings.Contains(path, excludeTestsBDD) {
			continue
		}
		f, err := parser.ParseFile(fset, path, nil, parser.SkipObjectResolution)
		if err != nil {
			return nil, fmt.Errorf("parse %s: %w", path, err)
		}
		ast.Inspect(f, func(n ast.Node) bool {
			if id, ok := n.(*ast.Ident); ok && id.IsExported() {
				refs[id.Name] = struct{}{}
			}
			// SelectorExpr (pkg.Name) — we want the .Sel name. The
			// Ident case above ALSO walks selectors when the
			// inspector descends, but SelectorExpr is the standard
			// way callers write `fuzzymatch.LevenshteinScore`.
			if sel, ok := n.(*ast.SelectorExpr); ok && sel.Sel != nil && sel.Sel.IsExported() {
				refs[sel.Sel.Name] = struct{}{}
			}
			return true
		})
	}

	return refs, nil
}

// sortedKeys returns the keys of m sorted lexicographically. Used to
// produce deterministic output (offender lists, summary counts) so the
// helper is reproducible across CI runs.
func sortedKeys(m map[string]struct{}) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return keys
}
