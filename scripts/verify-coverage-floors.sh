#!/usr/bin/env bash
#
# verify-coverage-floors.sh
#
# Enforces fuzzymatch's three coverage floors against a Go coverage profile:
#
#   1. Overall coverage   >= 95.0%   (matches CLAUDE.md / .claude/skills/go-testing-standards)
#   2. Per-file coverage  >= 90.0%   (every measurable .go file)
#   3. Per-exported-func  >= 90.0%   (every exported function in the root package)
#      Plus: every exported type/var/const must be referenced in at
#      least one *_test.go (AST-based detection).
#
# Coverage profile input: `coverage.out` at the repo root, produced by
# `make coverage` (i.e. `go test -race -coverprofile=coverage.out -covermode=atomic ./...`).
#
# Floor #3 semantics (Phase 8.5 Q12a LOCKED 2026-05-17)
# ------------------
# Floor 3 was tightened from "exists-at-least-one-test" to ">= 90.0%
# statement coverage per exported function". Detection moves from
# `go doc -short` parsing to an AST-based Go helper at
# scripts/cmd/verify-exported-coverage/main.go which:
#   - walks the root package via go/parser.ParseDir;
#   - cross-references exported FuncDecls against `go tool cover -func`;
#   - for non-func symbols, scans every *_test.go for at least one
#     identifier reference (Floor 3b lighter check).
# The helper is shelled out from this script; we forward its exit code.
#
# Tolerance for the bootstrap state
# ---------------------------------
# When the coverage profile contains no profiled statements (the tree has
# no test files yet, the Phase 1 bootstrap state), the script exits 0
# with a "no measurable files yet" note. Phase 2's first algorithm plan
# brings real code + tests; from that point onward the floors are
# enforced unconditionally.
#
# Usage
# -----
#   bash scripts/verify-coverage-floors.sh [coverage-profile-path]
#
# Default coverage-profile-path is `coverage.out`.
#
# Exit codes
# ----------
#   0 — all floors pass (or no measurable files yet)
#   1 — at least one floor failed (offenders printed to stderr)
#   2 — script invocation error (caught by `set -e` / `set -u`)

set -euo pipefail

readonly OVERALL_FLOOR=95.0
readonly PER_FILE_FLOOR=90.0
readonly PUBLIC_API_FLOOR=100.0  # enforced as "non-zero on every exported symbol"

PROFILE="${1:-coverage.out}"

if [ ! -f "$PROFILE" ]; then
    echo "verify-coverage-floors: FAIL — coverage profile not found: $PROFILE" >&2
    echo "  (run 'make coverage' first to produce $PROFILE)" >&2
    exit 1
fi

# Count profiled lines: every line in the profile except the leading
# `mode:` header is a (file, range, count) tuple. A profile with zero
# tuples means no .go files contributed any measurable statements yet.
profiled_lines=$(awk 'BEGIN{n=0} !/^mode:/{n++} END{print n}' "$PROFILE")

if [ "$profiled_lines" -eq 0 ]; then
    echo "OK: verify-coverage-floors — no measurable files yet (profile empty); pending Phase 2."
    exit 0
fi

# Parse the func-mode coverage table once. Output format:
#   <file>:<line>:\t<func>\t<percent>%
# Trailing row: `total:\t(statements)\t<percent>%`
FUNC_REPORT=$(go tool cover -func="$PROFILE" 2>/dev/null || true)
if [ -z "$FUNC_REPORT" ]; then
    echo "verify-coverage-floors: FAIL — go tool cover produced no output" >&2
    echo "  profile may be malformed or all files have zero statements" >&2
    exit 1
fi

# ---- Floor 1: overall coverage ----
total_pct=$(echo "$FUNC_REPORT" | awk '/^total:/ {gsub("%", "", $NF); print $NF; exit}')
if [ -z "$total_pct" ]; then
    echo "verify-coverage-floors: FAIL — could not parse total coverage from go tool cover" >&2
    exit 1
fi

overall_ok=$(awk -v t="$total_pct" -v f="$OVERALL_FLOOR" 'BEGIN { print (t+0 >= f+0) ? "1" : "0" }')
if [ "$overall_ok" != "1" ]; then
    echo "verify-coverage-floors: FAIL — overall coverage ${total_pct}% < ${OVERALL_FLOOR}%" >&2
    exit 1
fi

# ---- Floor 2: per-file coverage ----
# Aggregate per-file coverage by summing each file's function rows weighted
# by their statement counts. `go tool cover -func` does not directly emit
# per-file totals; we recompute from the raw profile.
#
# Per-line raw profile format:
#   filename:startLine.startCol,endLine.endCol numStatements count
#
# Per-file coverage:
#   sum(numStatements * (count > 0 ? 1 : 0)) / sum(numStatements)
declare -a per_file_offenders=()
while IFS=$'\t' read -r file covered_stmts total_stmts; do
    [ -z "$file" ] && continue
    # Skip files with zero statements (e.g. doc.go) — they have no measurable
    # coverage and the floor does not apply.
    if [ "$total_stmts" = "0" ]; then
        continue
    fi
    # Compute file percentage with awk (avoids bc dependency).
    file_pct=$(awk -v c="$covered_stmts" -v t="$total_stmts" 'BEGIN { printf "%.2f", (c/t)*100.0 }')
    file_ok=$(awk -v p="$file_pct" -v f="$PER_FILE_FLOOR" 'BEGIN { print (p+0 >= f+0) ? "1" : "0" }')
    if [ "$file_ok" != "1" ]; then
        per_file_offenders+=("${file} = ${file_pct}% (< ${PER_FILE_FLOOR}%)")
    fi
done < <(
    awk -F'[: ,.]' '
        !/^mode:/ {
            # file is field 1; numStatements is second-to-last; count is last.
            # Use the whole-line split on space to extract trailing fields.
            n = split($0, parts, " ");
            count = parts[n];
            stmts = parts[n-1];
            # Re-extract filename (everything before the first colon on the line).
            line = $0;
            colon = index(line, ":");
            file = substr(line, 1, colon-1);
            files_total[file] += stmts;
            if (count+0 > 0) files_covered[file] += stmts;
            else files_covered[file] += 0;
        }
        END {
            for (f in files_total) {
                printf "%s\t%d\t%d\n", f, files_covered[f]+0, files_total[f]+0;
            }
        }
    ' "$PROFILE"
)

if (( ${#per_file_offenders[@]} > 0 )); then
    {
        echo "verify-coverage-floors: FAIL — ${#per_file_offenders[@]} file(s) below per-file floor ${PER_FILE_FLOOR}%:"
        for line in "${per_file_offenders[@]}"; do
            echo "  $line"
        done
    } >&2
    exit 1
fi

# ---- Floor 3: >= 90% statement coverage per exported function ----
#
# Phase 8.5 Q12a LOCKED (2026-05-17): Floor 3 tightens from "exists-at-
# least-one-test" to ">= 90.0% statement coverage per exported function".
# Detection switches from `go doc -short` parsing (whitespace-sensitive,
# fragile around methods) to AST-based enumeration via the dedicated
# helper at scripts/cmd/verify-exported-coverage/main.go.
#
# The helper:
#   - walks the root package via go/parser.ParseDir (excludes *_test.go
#     and the tests/bdd/ subtree);
#   - enumerates exported FuncDecls (top-level), TypeSpecs, ValueSpecs;
#   - cross-references the function set against `go tool cover -func`
#     output for the supplied profile;
#   - for non-func symbols, AST-scans every *_test.go for at least one
#     identifier reference (Floor 3b lighter check).
#
# The helper exits 0 on pass, 1 on Floor-3 violations (offender list to
# stderr), 2 on invocation errors. We forward its exit code.

if ! command -v go >/dev/null 2>&1; then
    echo "verify-coverage-floors: FAIL — 'go' not on PATH; cannot invoke verify-exported-coverage helper" >&2
    exit 1
fi

# Resolve the repo root from this script's location so the helper is
# invoked with a stable CWD regardless of where the user runs us from.
REPO_ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"

if ! (
    cd "$REPO_ROOT"
    go run ./scripts/cmd/verify-exported-coverage "$PROFILE"
); then
    # The helper has already printed its own offender enumeration to
    # stderr. We add a one-line trailer for context.
    echo "verify-coverage-floors: Floor 3 failed — see verify-exported-coverage output above." >&2
    exit 1
fi

echo "OK: verify-coverage-floors — overall ${total_pct}% >= ${OVERALL_FLOOR}%; per-file >= ${PER_FILE_FLOOR}%; Floor 3 (AST-based) passed."
exit 0
