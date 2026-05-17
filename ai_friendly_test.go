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

// ai_friendly_test.go is the documentation-drift gate for the AI-friendly
// surface described in docs/requirements.md §18 and the
// documentation-standards skill. It asserts that every exported
// identifier in the root fuzzymatch package is referenced verbatim in
// llms.txt — the concise AI-friendly index that AI assistants and code
// generators consult first.
//
// Mechanism: walk the root package via go/parser.ParseDir, collect every
// exported identifier name (functions, types, methods, vars, consts) via
// go/ast, then assert strings.Contains(llmsTxtContent, symbolName) for
// each. A symbol may be absent from the check only if it appears in the
// internalAllowlist below.
//
// On failure the test prints the full missing-set so a single CI run
// surfaces every drift entry without re-running.

package fuzzymatch_test

import (
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"sort"
	"strings"
	"testing"
)

// internalAllowlist contains exported identifiers whose absence from
// llms.txt is permitted. Each entry includes a one-line rationale.
//
// Phase 8.5 Plan 15a (Q14b mechanical refactor) unexported writeGoldenFile
// and moved the test-only WriteGoldenFile re-export into export_test.go,
// so the symbol is no longer an exported root-package identifier and the
// allowlist no longer needs an entry for it. The map is retained empty
// (rather than removed) so that any future post-1.0 exported symbol that
// legitimately needs to be absent from llms.txt has an obvious place to
// land alongside a rationale.
var internalAllowlist = map[string]string{}

// TestAIFriendly_LLMSTxtReferencesEveryExportedSymbol enforces the
// llms.txt sync invariant. Drift between code and llms.txt fails this
// test; the developer must update llms.txt as part of the same PR that
// changes the exported API.
func TestAIFriendly_LLMSTxtReferencesEveryExportedSymbol(t *testing.T) {
	t.Parallel()

	llmsTxt, err := os.ReadFile("llms.txt")
	if err != nil {
		t.Fatalf("read llms.txt: %v", err)
	}
	llmsTxtContent := string(llmsTxt)

	exported := collectExportedSymbols(t)

	var missing []string
	for _, name := range exported {
		if _, ok := internalAllowlist[name]; ok {
			continue
		}
		if !strings.Contains(llmsTxtContent, name) {
			missing = append(missing, name)
		}
	}

	if len(missing) > 0 {
		sort.Strings(missing)
		t.Errorf(
			"llms.txt is missing %d exported identifier(s):\n  %s\n\n"+
				"Update llms.txt so every exported root-package symbol appears verbatim, OR\n"+
				"add the symbol to internalAllowlist with a rationale if it is intentionally\n"+
				"absent from the AI-friendly surface.",
			len(missing),
			strings.Join(missing, "\n  "),
		)
	}
}

// collectExportedSymbols walks the root package via go/parser.ParseDir
// and returns every exported identifier name (functions, methods, types,
// vars, consts) declared in non-test files. Returned names are sorted
// deterministically for stable test output.
func collectExportedSymbols(t *testing.T) []string {
	t.Helper()

	fset := token.NewFileSet()
	// Filter out _test.go files: their exports are test helpers and never
	// part of the public API surface that llms.txt promises.
	filter := func(info os.FileInfo) bool {
		return !strings.HasSuffix(info.Name(), "_test.go")
	}
	// parser.ParseDir is sufficient for this single-package, no-build-tag
	// root scan; go/packages would add a build-tools dep to test-only
	// code with no benefit at our scope.
	pkgs, err := parser.ParseDir(fset, ".", filter, parser.SkipObjectResolution) //nolint:staticcheck // SA1019: see comment above
	if err != nil {
		t.Fatalf("parse root package: %v", err)
	}

	seen := make(map[string]struct{})
	for _, pkg := range pkgs {
		// We only care about the "fuzzymatch" package, not any
		// hypothetical "fuzzymatch_test" external test package (filtered
		// above) or anything else that might land here.
		if pkg.Name != "fuzzymatch" {
			continue
		}
		for _, f := range pkg.Files {
			collectExportedFromFile(f, seen)
		}
	}

	names := make([]string, 0, len(seen))
	for name := range seen {
		names = append(names, name)
	}
	sort.Strings(names)
	return names
}

// collectExportedFromFile walks the top-level declarations of a single
// .go file and records every exported identifier into seen. Methods on
// exported receiver types are recorded as the method name (we want the
// symbol "String" to appear in llms.txt; the receiver context is
// captured separately by the type entry).
func collectExportedFromFile(f *ast.File, seen map[string]struct{}) {
	for _, decl := range f.Decls {
		switch d := decl.(type) {
		case *ast.FuncDecl:
			// Both top-level functions and methods.
			if d.Name != nil && d.Name.IsExported() {
				seen[d.Name.Name] = struct{}{}
			}
		case *ast.GenDecl:
			// Type, var, const declarations.
			for _, spec := range d.Specs {
				collectExportedFromSpec(spec, seen)
			}
		}
	}
}

// collectExportedFromSpec records exported names from a single
// declaration spec (TypeSpec, ValueSpec). Handles grouped declarations
// like `const ( AlgoLevenshtein AlgoID = iota; AlgoDamerauLevenshteinOSA; ... )`
// by walking the Names slice on each ValueSpec.
func collectExportedFromSpec(spec ast.Spec, seen map[string]struct{}) {
	switch s := spec.(type) {
	case *ast.TypeSpec:
		if s.Name != nil && s.Name.IsExported() {
			seen[s.Name.Name] = struct{}{}
		}
	case *ast.ValueSpec:
		for _, name := range s.Names {
			if name != nil && name.IsExported() {
				seen[name.Name] = struct{}{}
			}
		}
	}
}
