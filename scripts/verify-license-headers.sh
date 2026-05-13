#!/usr/bin/env bash
#
# verify-license-headers.sh
#
# Exits 0 if every .go file tracked under the repository carries the AxonOps
# Apache-2.0 header within its first 25 lines; exits non-zero with a list of
# offending files printed to stderr otherwise.
#
# This script is invoked from the Makefile target `verify-license-headers`
# and from the corresponding CI workflow. It is idempotent, has no side
# effects beyond writing to stdout/stderr, and is safe to re-run.
#
# Header signature: the literal substring
#   Licensed under the Apache License, Version 2.0
# which appears in the canonical AxonOps Apache-2.0 file header defined in
# .planning/phases/01-foundation-infrastructure/01-01-module-bootstrap-PLAN.md.
#
# Discovery mode:
#   - Prefers `git ls-files '*.go'` so only tracked files are inspected
#     (untracked scratch files never cause spurious failures).
#   - Falls back to `find . -name '*.go'` (excluding the .git directory) when
#     the working tree is not inside a git repo, e.g. a freshly-extracted
#     source tarball in CI.
#
# Exit codes:
#   0 — every tracked .go file carries the header
#   1 — at least one tracked .go file is missing the header (offending paths
#       printed to stderr, one per line)
#   2 — script invocation error (caught by `set -e` / `set -u`)

set -euo pipefail

HEADER_NEEDLE='Licensed under the Apache License, Version 2.0'
SEARCH_LINES=25

# Discover .go files. Prefer git; fall back to find.
declare -a go_files=()
if git rev-parse --is-inside-work-tree >/dev/null 2>&1; then
    # `git ls-files` is null-safe; use -z + read for filenames with spaces.
    while IFS= read -r -d '' path; do
        go_files+=("$path")
    done < <(git ls-files -z -- '*.go')
else
    while IFS= read -r -d '' path; do
        go_files+=("$path")
    done < <(find . -type f -name '*.go' -not -path './.git/*' -print0)
fi

if [ "${#go_files[@]}" -eq 0 ]; then
    echo "OK: 0 .go files to check." >&2
    exit 0
fi

declare -a offenders=()
for f in "${go_files[@]}"; do
    if ! head -n "$SEARCH_LINES" -- "$f" | grep -qF -- "$HEADER_NEEDLE"; then
        offenders+=("$f")
    fi
done

if [ "${#offenders[@]}" -gt 0 ]; then
    {
        echo "verify-license-headers: ${#offenders[@]} .go file(s) missing the Apache-2.0 header:"
        for f in "${offenders[@]}"; do
            echo "  $f"
        done
    } >&2
    exit 1
fi

echo "OK: ${#go_files[@]} .go files carry the Apache-2.0 header."
exit 0
