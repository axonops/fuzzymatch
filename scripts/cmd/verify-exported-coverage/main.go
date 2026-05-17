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
		fmt.Fprintln(os.Stderr, err)
		// Exit 2 for invocation errors (missing files, malformed
		// profile); exit 1 for floor violations (run returns a
		// floorViolationErr in that case).
		if _, ok := err.(*floorViolationErr); ok {
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
func run(profile string, out, errOut io.Writer) error {
	if _, err := os.Stat(profile); err != nil {
		return fmt.Errorf("verify-exported-coverage: coverage profile not found: %s", profile)
	}

	exportedFuncs, exportedNonFuncs, err := collectExported(rootPackageDir)
	if err != nil {
		return fmt.Errorf("verify-exported-coverage: AST walk failed: %w", err)
	}
	if len(exportedFuncs) == 0 && len(exportedNonFuncs) == 0 {
		fmt.Fprintln(out,
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
		fmt.Fprintf(errOut,
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

	fmt.Fprintf(out,
		"OK: verify-exported-coverage — %d exported func(s) >= %.1f%%; %d exported type/var/const all referenced in *_test.go.\n",
		len(exportedFuncs), floorPercent, len(exportedNonFuncs))
	return nil
}

// collectExported walks dir via go/parser.ParseDir and returns two
// sorted slices: exportedFuncs (top-level FuncDecls; methods excluded
// because their coverage is rolled up to their receiver type and the
// per-method floor is intentionally NOT enforced separately) and
// exportedNonFuncs (TypeSpec / ValueSpec names — types, vars, consts).
//
// Test files (*_test.go) and the tests/bdd subtree are excluded from
// the walk. The helper expects to be invoked from the repo root.
func collectExported(dir string) (funcs, nonFuncs []string, err error) {
	fset := token.NewFileSet()
	filter := func(info os.FileInfo) bool {
		name := info.Name()
		if strings.HasSuffix(name, "_test.go") {
			return false
		}
		return true
	}
	// parser.SkipObjectResolution speeds the parse for AST-only walks
	// (we never resolve cross-file identifier links).
	pkgs, err := parser.ParseDir(fset, dir, filter, parser.SkipObjectResolution)
	if err != nil {
		return nil, nil, fmt.Errorf("parser.ParseDir(%s): %w", dir, err)
	}

	funcSet := make(map[string]struct{})
	nonFuncSet := make(map[string]struct{})

	for _, pkg := range pkgs {
		if pkg.Name != rootPackageName {
			continue
		}
		for path, f := range pkg.Files {
			// Defence in depth: skip the BDD subtree even if
			// ParseDir surfaces it via symlink. ParseDir does not
			// recurse so this guard is belt-and-braces.
			if strings.Contains(path, excludeTestsBDD) {
				continue
			}
			collectExportedFromFile(f, funcSet, nonFuncSet)
		}
	}

	funcs = sortedKeys(funcSet)
	nonFuncs = sortedKeys(nonFuncSet)
	return funcs, nonFuncs, nil
}

// collectExportedFromFile walks the top-level Decls of a single
// *ast.File and partitions exported names into funcs (FuncDecl with no
// receiver — i.e. package-level functions, NOT methods) vs non-funcs
// (TypeSpec, ValueSpec).
func collectExportedFromFile(f *ast.File, funcs, nonFuncs map[string]struct{}) {
	for _, decl := range f.Decls {
		switch d := decl.(type) {
		case *ast.FuncDecl:
			// Package-level functions only (Recv == nil).
			// Methods are reachable via their receiver type;
			// `go tool cover -func` reports them with the
			// "(Recv).Name" composite key which the Floor-3a
			// gate intentionally does NOT walk per-method (a
			// per-method floor would over-constrain method
			// receivers that legitimately have rare paths).
			if d.Recv != nil {
				continue
			}
			if d.Name != nil && d.Name.IsExported() {
				funcs[d.Name.Name] = struct{}{}
			}
		case *ast.GenDecl:
			for _, spec := range d.Specs {
				collectExportedFromSpec(spec, nonFuncs)
			}
		}
	}
}

// collectExportedFromSpec records exported identifier names from a
// single declaration spec (TypeSpec or ValueSpec). Grouped declarations
// like `const ( A AlgoID = iota; B; C )` produce one ValueSpec per
// named row, each contributing every name in spec.Names.
func collectExportedFromSpec(spec ast.Spec, nonFuncs map[string]struct{}) {
	switch s := spec.(type) {
	case *ast.TypeSpec:
		if s.Name != nil && s.Name.IsExported() {
			nonFuncs[s.Name.Name] = struct{}{}
		}
	case *ast.ValueSpec:
		for _, name := range s.Names {
			if name != nil && name.IsExported() {
				nonFuncs[name.Name] = struct{}{}
			}
		}
	}
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
	cmd := exec.Command("go", "tool", "cover", "-func="+profile)
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
func collectTestRefs(dir string) (map[string]struct{}, error) {
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
