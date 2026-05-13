#!/usr/bin/env bash
#
# verify-no-runtime-deps.sh
#
# Exits 0 if the root go.mod's resolved runtime dependency graph contains
# exactly one non-stdlib runtime module (golang.org/x/text). Exits 1 with a
# diff printed to stderr otherwise.
#
# Strategy
# --------
# `go list -m all` returns every module in the build graph — including
# indirect modules that Go resolves during module-graph computation but
# never actually compiles into the binary (e.g. golang.org/x/mod,
# golang.org/x/sync, golang.org/x/tools when x/text references them in
# its own go.mod). Those indirect entries have no h1: source hash in
# go.sum, only a /go.mod hash; they are not part of the runtime
# dependency surface.
#
# This script therefore filters `go list -m -mod=mod all` to non-indirect
# modules, asserting the resulting set equals the allowlist exactly:
#   - github.com/axonops/fuzzymatch (the main module — always present)
#   - golang.org/x/text             (the single curated runtime dep)
#
# Any other non-indirect module fails the script with a diff. Adding a
# third non-indirect dep requires explicit user approval AND
# algorithm-licensing-reviewer sign-off per CLAUDE.md.
#
# Why not enforce on the full build list?
# ---------------------------------------
# `go list -m all` exposes indirect modules driven entirely by Go's
# module-graph laziness (MVS predecessors of x/text's own toolchain
# requirements). Failing CI on those would yield false positives without
# any supply-chain meaning — none of those packages' code ever enters
# the compiled artefact.
#
# Usage
# -----
#   bash scripts/verify-no-runtime-deps.sh
#
# Exit codes
# ----------
#   0 — root go.mod's non-indirect dep set == [fuzzymatch, x/text]
#   1 — extraneous non-indirect modules present (diff on stderr)
#   2 — script invocation error (caught by `set -e` / `set -u`)

set -euo pipefail

readonly ALLOWED=(
    "github.com/axonops/fuzzymatch"
    "golang.org/x/text"
)

# Collect non-indirect modules from the build list. The template prints
# one module path per line iff the module's .Indirect field is false.
# Read line-by-line for bash 3.2 compatibility (macOS default).
declare -a actual=()
while IFS= read -r path; do
    [ -z "$path" ] && continue
    actual+=("$path")
done < <(
    go list -m -mod=mod -f '{{if not .Indirect}}{{.Path}}{{end}}' all \
        2>/dev/null \
        | awk 'NF' \
        | LC_ALL=C sort -u
)

# Determine extras: anything in `actual` that does not appear in `ALLOWED`
# (or as a sub-path of an `ALLOWED` entry — defensive against future Go
# module reporting changes that might split out sub-packages).
declare -a extras=()
for mod in "${actual[@]}"; do
    ok=0
    for a in "${ALLOWED[@]}"; do
        if [[ "$mod" == "$a" ]]; then
            ok=1
            break
        fi
        if [[ "$mod" == "$a/"* ]]; then
            ok=1
            break
        fi
    done
    if (( ok == 0 )); then
        extras+=("$mod")
    fi
done

if (( ${#extras[@]} > 0 )); then
    {
        echo "verify-no-runtime-deps: FAIL — unexpected non-indirect runtime modules in root go.mod:"
        for m in "${extras[@]}"; do
            echo "  + $m"
        done
        echo
        echo "Allowed (per CLAUDE.md runtime-dep allowlist):"
        for a in "${ALLOWED[@]}"; do
            echo "  - $a"
        done
        echo
        echo "Adding a new runtime dep requires explicit user approval AND"
        echo "algorithm-licensing-reviewer sign-off."
    } >&2
    exit 1
fi

# Sanity check: the allowlist entries we expect to see MUST actually be in
# the build list. Catches accidental removal of x/text (e.g. someone
# deletes the blank import in doc.go and runs `go mod tidy`).
declare -a missing=()
for a in "${ALLOWED[@]}"; do
    found=0
    for mod in "${actual[@]}"; do
        if [[ "$mod" == "$a" ]]; then
            found=1
            break
        fi
    done
    if (( found == 0 )); then
        missing+=("$a")
    fi
done

if (( ${#missing[@]} > 0 )); then
    {
        echo "verify-no-runtime-deps: FAIL — expected allowlisted modules absent from build list:"
        for m in "${missing[@]}"; do
            echo "  - $m"
        done
        echo
        echo "If a curated runtime dep was intentionally removed, update the"
        echo "ALLOWED array in this script in the same commit."
    } >&2
    exit 1
fi

echo "OK: root go.mod allowlist clean (${#actual[@]} non-indirect modules: ${actual[*]})"
exit 0
