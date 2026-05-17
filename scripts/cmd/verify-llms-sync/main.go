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

// verify-llms-sync is the CI-facing AI-friendly-docs drift gate that
// implements Phase 8.5 Q13. It walks the root fuzzymatch package via
// the shared scripts/internal/astwalk helper, enumerates every exported
// top-level symbol (functions, methods, types, vars, consts), and
// asserts each one appears verbatim in llms.txt (strictly) and in
// llms-full.txt (advisory-until-Plan-17).
//
// Mechanism: companion to the in-process gate
// `TestAIFriendly_LLMSTxtReferencesEveryExportedSymbol` in
// ai_friendly_test.go. That unit test runs inside `go test`; this
// helper runs as a standalone CI step before `make check`, surfacing
// drift earlier and without needing the full test pipeline. Both gates
// share the AST walk via scripts/internal/astwalk so they cannot
// disagree about the set of exported symbols.
//
// llms-full.txt strictness toggle: the Phase 8.5 Plan 13 SUMMARY records
// that the 6 doc-residue surfaces (including llms-full.txt fill-in) are
// deferred to Plan 17. Until Plan 17 lands the canonical algorithm-tier
// fill-in, llms-full.txt has known drift gaps (32 Phase 1–3 symbols not
// indexed in the per-phase sections that start at Phase 4). The strict
// gate is therefore scoped to llms.txt only by default; llms-full.txt
// drift is reported as a WARNING. Flip `-strict-llms-full` to true once
// Plan 17 lands to promote the warning to a failure.
//
// Usage:
//
//	go run ./scripts/cmd/verify-llms-sync
//
// Optional flags:
//
//	-llms-txt           path to llms.txt        (default: "llms.txt")
//	-llms-full          path to llms-full.txt   (default: "llms-full.txt")
//	-package            package name to scope   (default: "fuzzymatch")
//	-dir                package directory       (default: ".")
//	-strict-llms-full   promote llms-full drift from warning to failure
//	                    (default: false — flip to true after Plan 17 lands)
//
// Exit codes:
//
//	0 — every exported symbol appears in llms.txt (strict) AND either:
//	    - every exported symbol appears in llms-full.txt; OR
//	    - llms-full.txt drift exists but -strict-llms-full is false.
//	1 — at least one exported symbol is missing from llms.txt (always),
//	    or missing from llms-full.txt with -strict-llms-full=true.
//	2 — script invocation error (missing files, AST parse failure).
//
// Threat model (Phase 8.5 T-08.5-30 / T-08.5-32 adjacent): the helper
// trusts the working tree it is invoked against. CI invocation guarantees
// the tree matches the tagged / PR'd commit; local invocations inherit
// the developer's working tree state.
package main

import (
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"sort"
	"strings"

	"github.com/axonops/fuzzymatch/scripts/internal/astwalk"
)

// internalAllowlist mirrors the allowlist in ai_friendly_test.go: a
// map of exported identifier name → rationale for being absent from the
// AI-friendly docs. Phase 8.5 Plan 15a emptied the allowlist (writeGoldenFile
// was unexported). The map is retained empty so any future post-1.0
// exported symbol that legitimately needs to be absent has an obvious
// place to land alongside a rationale.
var internalAllowlist = map[string]string{}

// rootPackageName is the Go package name we scope the AST walk to.
const rootPackageName = "fuzzymatch"

func main() {
	flag.Usage = func() {
		fmt.Fprintf(os.Stderr,
			"usage: %s [flags]\n\n"+
				"Verifies Phase 8.5 Q13 llms.txt / llms-full.txt sync: every exported\n"+
				"symbol in the root fuzzymatch package must appear verbatim in BOTH\n"+
				"llms.txt and llms-full.txt.\n\nFlags:\n",
			os.Args[0])
		flag.PrintDefaults()
	}
	llmsTxtPath := flag.String("llms-txt", "llms.txt", "path to llms.txt")
	llmsFullPath := flag.String("llms-full", "llms-full.txt", "path to llms-full.txt")
	pkgDir := flag.String("dir", ".", "package directory to walk")
	pkgName := flag.String("package", rootPackageName, "package name to scope")
	strictFull := flag.Bool("strict-llms-full", false,
		"promote llms-full.txt drift from warning to failure (flip to true after Plan 17 lands)")
	flag.Parse()

	if err := run(*pkgDir, *pkgName, *llmsTxtPath, *llmsFullPath, *strictFull, os.Stdout, os.Stderr); err != nil {
		fmt.Fprintln(os.Stderr, err) //nolint:errcheck // best-effort stderr emission; exit code is the canonical signal
		var dve *driftViolationErr
		if errors.As(err, &dve) {
			os.Exit(1)
		}
		os.Exit(2)
	}
}

// driftViolationErr signals one or more llms-drift violations. It is
// distinguished from other errors so main() can map to exit code 1
// (drift) vs 2 (invocation).
type driftViolationErr struct {
	count int
	msg   string
}

func (e *driftViolationErr) Error() string { return e.msg }

// run is the testable entry point. It writes the OK summary to out and
// detailed drift enumeration to errOut, returning a driftViolationErr on
// drift or a plain error on invocation problems (missing files, parse
// failures).
//
// strictFull promotes llms-full.txt drift from a warning to a failure;
// llms.txt drift is always a failure (matches the existing in-process
// gate semantics).
func run(pkgDir, pkgName, llmsTxtPath, llmsFullPath string, strictFull bool, out, errOut io.Writer) error {
	llmsTxt, err := readFile(llmsTxtPath)
	if err != nil {
		return err
	}
	llmsFull, err := readFile(llmsFullPath)
	if err != nil {
		return err
	}

	res, err := astwalk.CollectExported(pkgDir, pkgName)
	if err != nil {
		return fmt.Errorf("verify-llms-sync: AST walk failed: %w", err)
	}

	exported := res.AllNames()
	if len(exported) == 0 {
		fmt.Fprintln(out, //nolint:errcheck // best-effort write to writer; caller already controls the writer choice
			"OK: verify-llms-sync — no exported symbols in the root package (bootstrap state).")
		return nil
	}

	var missingInLLMS []string
	var missingInLLMSFull []string

	for _, name := range exported {
		if _, ok := internalAllowlist[name]; ok {
			continue
		}
		if !strings.Contains(llmsTxt, name) {
			missingInLLMS = append(missingInLLMS, name)
		}
		if !strings.Contains(llmsFull, name) {
			missingInLLMSFull = append(missingInLLMSFull, name)
		}
	}

	sort.Strings(missingInLLMS)
	sort.Strings(missingInLLMSFull)

	// llms.txt drift is ALWAYS a failure.
	if len(missingInLLMS) > 0 {
		fmt.Fprintln(errOut, "verify-llms-sync: FAIL — AI-friendly docs are out of sync with the exported API surface.") //nolint:errcheck // best-effort
		fmt.Fprintf(errOut, "\n  Missing from %s (%d):\n    %s\n",                                                       //nolint:errcheck // best-effort
			llmsTxtPath, len(missingInLLMS), strings.Join(missingInLLMS, "\n    "))
		if len(missingInLLMSFull) > 0 {
			fmt.Fprintf(errOut, "\n  Also missing from %s (%d):\n    %s\n", //nolint:errcheck // best-effort
				llmsFullPath, len(missingInLLMSFull), strings.Join(missingInLLMSFull, "\n    "))
		}
		fmt.Fprintln(errOut, "\nFix: add each missing symbol to the file(s) above, OR add it to internalAllowlist with a rationale if intentionally absent.") //nolint:errcheck // best-effort
		return &driftViolationErr{
			count: len(missingInLLMS) + len(missingInLLMSFull),
			msg:   fmt.Sprintf("%d llms.txt-drift violation(s)", len(missingInLLMS)),
		}
	}

	// llms-full.txt drift: failure under -strict-llms-full, warning otherwise.
	if len(missingInLLMSFull) > 0 {
		if strictFull {
			fmt.Fprintln(errOut, "verify-llms-sync: FAIL — llms-full.txt is out of sync with the exported API surface (-strict-llms-full=true).") //nolint:errcheck // best-effort
			fmt.Fprintf(errOut, "\n  Missing from %s (%d):\n    %s\n",                                                                            //nolint:errcheck // best-effort
				llmsFullPath, len(missingInLLMSFull), strings.Join(missingInLLMSFull, "\n    "))
			fmt.Fprintln(errOut, "\nFix: add each missing symbol to llms-full.txt, OR add it to internalAllowlist with a rationale.") //nolint:errcheck // best-effort
			return &driftViolationErr{
				count: len(missingInLLMSFull),
				msg:   fmt.Sprintf("%d llms-full.txt-drift violation(s)", len(missingInLLMSFull)),
			}
		}
		// Advisory-only warning to stdout. The exit code stays 0 so CI
		// passes; the message is grep-able by `WARN: verify-llms-sync`.
		fmt.Fprintf(out, //nolint:errcheck // best-effort
			"WARN: verify-llms-sync — %d exported symbol(s) missing from %s (advisory until Plan 17 lands; flip -strict-llms-full=true to promote):\n  %s\n",
			len(missingInLLMSFull), llmsFullPath, strings.Join(missingInLLMSFull, "\n  "))
	}

	fmt.Fprintf(out, //nolint:errcheck // best-effort
		"OK: verify-llms-sync — %d exported symbol(s) all referenced in %s%s.\n",
		len(exported), llmsTxtPath,
		strictOrAdvisorySuffix(strictFull, len(missingInLLMSFull), llmsFullPath))
	return nil
}

// strictOrAdvisorySuffix produces the trailing clause of the OK message
// to clarify whether the llms-full.txt arm was strict (and passed) or
// advisory (and either passed or warned).
func strictOrAdvisorySuffix(strictFull bool, missingFull int, llmsFullPath string) string {
	if strictFull && missingFull == 0 {
		return fmt.Sprintf(" and %s", llmsFullPath)
	}
	if !strictFull && missingFull == 0 {
		return fmt.Sprintf(" (and %s, advisory mode)", llmsFullPath)
	}
	// !strictFull && missingFull > 0 — the warning message already
	// surfaced the gap; the OK line only confirms llms.txt is clean.
	return ""
}

// readFile reads a file and returns its contents as a string, returning
// a wrapped error on failure.
func readFile(path string) (string, error) {
	content, err := os.ReadFile(path)
	if err != nil {
		return "", fmt.Errorf("verify-llms-sync: read %s: %w", path, err)
	}
	return string(content), nil
}
