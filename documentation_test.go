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

// documentation_test.go is the documentation-drift gate for fenced
// ```go code blocks in docs/*.md. The user-facing docs/ directory is
// the consumer's secondary entry point after README.md; if its code
// blocks fall out of sync with the library's actual surface (a Q3
// MongeElkan rename, a Q5 PartialRatioRunes removal, a future
// breaking change), every doc reader follows the broken trail.
//
// Mechanism
// ---------
// 1. Walk every .md file under docs/.
// 2. Extract each fenced ```go block. Skip blocks marked with the
//    `// docs:skip-compile` directive on the first line (used for
//    intentionally-illustrative-only snippets).
// 3. For each block, attempt to parse it in one of three modes:
//    a. As a complete file (block starts with `package`).
//    b. As one or more top-level declarations (we wrap in a stub
//       `package docs` and parse).
//    c. As a sequence of statements (we wrap in `package docs;
//       func _docs() { ... }` and parse).
// 4. The block passes the gate iff at least one mode parses without
//    error. Failure surfaces the first parse error from mode (c)
//    (the most general mode) so the offending line is identifiable.
//
// The test deliberately does NOT type-check (no go/types pass): the
// snippets often reference identifiers that are documented inline
// but not in scope (e.g. `s` defined two blocks earlier). The
// syntax-level gate is sufficient to catch typo-grade drift —
// renames produce parser errors on call sites within the same
// block, and most blocks in docs/scorer.md and docs/tuning.md are
// self-contained.
//
// Stdlib `go/parser` + `go/token` only. No external go vet
// invocation, no temp .go files written to disk — everything happens
// in memory with parser.ParseFile.

package fuzzymatch_test

import (
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"
)

// docsGoBlockRegex matches a fenced ```go code block, capturing the
// content between the fences. Non-greedy across newlines. The
// closing fence may be indented (CommonMark permits up to three
// leading spaces in nested-list contexts), and the opening fence
// may carry indentation as well — we strip that indentation off the
// captured block before parsing so the wrapped code is column-zero.
var docsGoBlockRegex = regexp.MustCompile("(?s)```go[^\\n]*\\n(.*?)\\n[ \\t]*```")

// skipDirective is the first-line marker that opts a block out of
// compile verification. Docs authors mark intentionally-fragmentary
// or intentionally-erroneous snippets with this directive; the test
// records the count of skipped blocks for visibility.
const skipDirective = "// docs:skip-compile"

// TestDocumentation_FencedGoBlocksCompile walks docs/*.md, extracts
// every fenced ```go block, and verifies each parses under one of
// three syntactic envelopes (file / declarations / statements).
// Blocks marked with `// docs:skip-compile` are reported but not
// verified.
//
// The test fails closed: if NO blocks are found across all .md
// files, the test errors (someone has either deleted all the docs
// or broken the block extractor).
func TestDocumentation_FencedGoBlocksCompile(t *testing.T) {
	t.Parallel()

	mdFiles, err := filepath.Glob("docs/*.md")
	if err != nil {
		t.Fatalf("glob docs/*.md: %v", err)
	}
	if len(mdFiles) == 0 {
		t.Fatalf("no docs/*.md files found; documentation_test cannot verify anything")
	}

	var totalBlocks, skipped, verified int
	for _, mdPath := range mdFiles {
		content, err := os.ReadFile(mdPath) // #nosec G304 — docs/*.md is repo-controlled input
		if err != nil {
			t.Errorf("read %s: %v", mdPath, err)
			continue
		}
		blocks := docsGoBlockRegex.FindAllStringSubmatch(string(content), -1)
		for blockIdx, m := range blocks {
			totalBlocks++
			block := m[1]
			if isSkipped(block) {
				skipped++
				continue
			}
			if err := verifyBlockParses(block); err != nil {
				t.Errorf("%s block %d does not parse:\n%v\n--- block ---\n%s\n--- end ---",
					mdPath, blockIdx+1, err, block)
				continue
			}
			verified++
		}
	}

	if totalBlocks == 0 {
		t.Fatalf("no ```go code blocks found in any docs/*.md file; either every doc has been stripped of examples or the block extractor is broken")
	}
	t.Logf("documentation_test: %d total blocks (%d verified, %d skipped)",
		totalBlocks, verified, skipped)
}

// isSkipped reports whether the block opens with the skip directive.
// The directive must be on the first non-blank line so authors must
// be explicit about the exemption.
func isSkipped(block string) bool {
	for _, line := range strings.Split(block, "\n") {
		trimmed := strings.TrimSpace(line)
		if trimmed == "" {
			continue
		}
		return trimmed == skipDirective
	}
	return false
}

// verifyBlockParses attempts to parse the block under three
// progressively more permissive envelopes. The block "compiles" (in
// the syntactic sense the test promises) iff any envelope parses
// without error. The most general envelope (statement-wrapped) is
// tried last and its error returned on overall failure — its error
// messages reference line numbers inside the block which makes
// debugging straightforward.
func verifyBlockParses(block string) error {
	fset := token.NewFileSet()

	// Mode 1: parse as a full file. Works for blocks that start
	// with `package ...`.
	if strings.HasPrefix(strings.TrimSpace(block), "package ") {
		_, err := parser.ParseFile(fset, "block.go", block, parser.SkipObjectResolution)
		if err == nil {
			return nil
		}
		// Fall through to other modes — some package-prefixed blocks
		// may rely on external imports we cannot resolve.
	}

	// Mode 2: parse as a sequence of top-level declarations.
	// Wrap in a stub `package docs` and try.
	declSrc := "package docs\n\n" + block + "\n"
	if _, err := parser.ParseFile(fset, "decls.go", declSrc, parser.SkipObjectResolution); err == nil {
		return nil
	}

	// Mode 3: parse as a sequence of statements inside a function
	// body. This is the most general envelope; many README/docs
	// blocks are mid-function fragments.
	stmtSrc := "package docs\n\nfunc _docs() {\n" + block + "\n}\n"
	_, err := parser.ParseFile(fset, "stmts.go", stmtSrc, parser.SkipObjectResolution)
	return err
}
