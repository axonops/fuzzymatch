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

// Package astwalk is the shared AST helper consumed by the two Q13
// internal-tooling helpers — `scripts/cmd/verify-exported-coverage` (the
// Floor 3a/3b gate) and `scripts/cmd/verify-llms-sync` (the llms.txt /
// llms-full.txt drift gate). Both helpers answer the same primitive
// question: "what does the root `fuzzymatch` package export?" and we keep
// the answer in one place so the two gates never disagree about the set
// of exported symbols.
//
// The walker is intentionally minimal: it walks one package directory,
// honours the test-file filter and the `tests/bdd` subtree exclusion,
// and returns three sorted slices — package-level FuncDecls, methods on
// exported receivers, and exported non-functions (TypeSpec / ValueSpec
// names). Methods are surfaced separately so the verify-llms-sync gate
// (which asserts `strings.Contains(llmsTxt, method.Name)`) can verify a
// method's name appears regardless of whether the receiver is named in
// llms.txt verbatim — matching the existing
// `TestAIFriendly_LLMSTxtReferencesEveryExportedSymbol` behaviour in
// `ai_friendly_test.go` (which records method names as bare identifiers).
//
// Determinism: all output slices are sorted lexicographically so callers
// can rely on stable iteration order across CI runs.
package astwalk

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"sort"
	"strings"
)

// excludeTestsBDD is the prefix of the BDD sub-module which has its own
// go.mod and is therefore parsed separately. The walk scopes only the
// root module's .go files.
const excludeTestsBDD = "tests/bdd"

// Result is the structured output of a CollectExported call.
//
//   - Funcs: package-level FuncDecls with no receiver (top-level
//     functions). Methods are recorded separately because the two
//     consumers treat them differently (coverage-floor enforcement vs
//     llms.txt-presence enforcement).
//   - Methods: FuncDecls with a non-nil receiver whose name is exported.
//   - NonFuncs: TypeSpec / ValueSpec names — types, vars, consts.
//
// All three slices are sorted ascending.
type Result struct {
	Funcs    []string
	Methods  []string
	NonFuncs []string
}

// AllNames returns the union of Funcs ∪ Methods ∪ NonFuncs, sorted
// ascending and deduplicated. This is the canonical "exported surface"
// set that verify-llms-sync asserts against the llms.txt / llms-full.txt
// content.
func (r Result) AllNames() []string {
	seen := make(map[string]struct{}, len(r.Funcs)+len(r.Methods)+len(r.NonFuncs))
	for _, n := range r.Funcs {
		seen[n] = struct{}{}
	}
	for _, n := range r.Methods {
		seen[n] = struct{}{}
	}
	for _, n := range r.NonFuncs {
		seen[n] = struct{}{}
	}
	out := make([]string, 0, len(seen))
	for name := range seen {
		out = append(out, name)
	}
	sort.Strings(out)
	return out
}

// CollectExported walks dir via go/parser.ParseDir and returns the
// exported-symbol Result for packages whose Name matches packageName.
// Test files (*_test.go) and the tests/bdd subtree are excluded from
// the walk.
//
// The caller is expected to invoke the helper from the repo root (or
// from a path whose relative `dir` resolves to the target package).
func CollectExported(dir, packageName string) (Result, error) {
	fset := token.NewFileSet()
	filter := func(info os.FileInfo) bool {
		return !strings.HasSuffix(info.Name(), "_test.go")
	}
	// parser.SkipObjectResolution speeds the parse for AST-only walks
	// (we never resolve cross-file identifier links). parser.ParseDir
	// is deprecated as of Go 1.25 in favour of golang.org/x/tools/go/
	// packages — we keep it deliberately here because (a) this package
	// is internal tooling not consumed by downstream code, (b) we do
	// not inspect build tags so the deprecation rationale does not
	// apply, and (c) golang.org/x/tools is not part of the root
	// module's zero-runtime-dep allowlist.
	pkgs, err := parser.ParseDir(fset, dir, filter, parser.SkipObjectResolution) //nolint:staticcheck // SA1019: ParseDir suits internal AST tooling; see comment above
	if err != nil {
		return Result{}, fmt.Errorf("parser.ParseDir(%s): %w", dir, err)
	}

	funcSet := make(map[string]struct{})
	methodSet := make(map[string]struct{})
	nonFuncSet := make(map[string]struct{})

	for _, pkg := range pkgs {
		if pkg.Name != packageName {
			continue
		}
		for path, f := range pkg.Files {
			// Defence in depth: skip the BDD subtree even if
			// ParseDir surfaces it via symlink. ParseDir does not
			// recurse so this guard is belt-and-braces.
			if strings.Contains(path, excludeTestsBDD) {
				continue
			}
			collectFromFile(f, funcSet, methodSet, nonFuncSet)
		}
	}

	return Result{
		Funcs:    sortedKeys(funcSet),
		Methods:  sortedKeys(methodSet),
		NonFuncs: sortedKeys(nonFuncSet),
	}, nil
}

// collectFromFile walks the top-level Decls of a single *ast.File and
// partitions exported names into funcs (FuncDecl with no receiver),
// methods (FuncDecl with a non-nil receiver), and non-funcs (TypeSpec,
// ValueSpec).
func collectFromFile(f *ast.File, funcs, methods, nonFuncs map[string]struct{}) {
	for _, decl := range f.Decls {
		switch d := decl.(type) {
		case *ast.FuncDecl:
			if d.Name == nil || !d.Name.IsExported() {
				continue
			}
			if d.Recv != nil {
				methods[d.Name.Name] = struct{}{}
				continue
			}
			funcs[d.Name.Name] = struct{}{}
		case *ast.GenDecl:
			for _, spec := range d.Specs {
				collectFromSpec(spec, nonFuncs)
			}
		}
	}
}

// collectFromSpec records exported identifier names from a single
// declaration spec (TypeSpec or ValueSpec). Grouped declarations like
// `const ( A AlgoID = iota; B; C )` produce one ValueSpec per named row,
// each contributing every name in spec.Names.
func collectFromSpec(spec ast.Spec, nonFuncs map[string]struct{}) {
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

// sortedKeys returns the keys of m sorted lexicographically. Used to
// produce deterministic output across CI runs.
func sortedKeys(m map[string]struct{}) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return keys
}
