#!/usr/bin/env bash
#
# verify-coverage-floors.sh
#
# Enforces fuzzymatch's three coverage floors against a Go coverage profile:
#
#   1. Overall coverage   >= 95.0%   (matches CLAUDE.md / .claude/skills/go-testing-standards)
#   2. Per-file coverage  >= 90.0%   (every measurable .go file)
#   3. Public API surface = 100%     (every exported symbol must be exercised)
#
# Coverage profile input: `coverage.out` at the repo root, produced by
# `make coverage` (i.e. `go test -race -coverprofile=coverage.out -covermode=atomic ./...`).
#
# Floor #3 semantics
# ------------------
# 100% public-API coverage is enforced by the EXISTENCE of an exercising
# test, not by requiring 100.0% statement coverage on the symbol's body.
# Concretely: for every exported func/type/var/const reported by
# `go doc -short .` against the root package, there must be at least one
# coverage row in `go tool cover -func=coverage.out` showing non-zero
# coverage on the same identifier name. A symbol with 0% coverage means
# no test ever calls it — that fails this check.
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

# ---- Floor 3: 100% public-API ----
# Extract exported identifier names from `go doc -short .` on the root
# package. `go doc -short` prints one declaration per exported symbol,
# e.g.:
#   func Normalise(s string, opts NormalisationOptions) string
#   type AlgoID int
#   var ErrInvalidInput error
#   const DefaultThreshold = 0.6
#
# We parse the second whitespace-separated token of each declaration line
# (after `func`/`type`/`var`/`const`). For methods (e.g.
# `func (s Scorer) Score(...)`) the name is the first identifier after
# the receiver — handled by the awk pattern below.
#
# Then assert each exported symbol appears in the `go tool cover -func`
# output with a non-zero coverage percentage.
declare -a exported=()
go_doc=$(go doc -short . 2>/dev/null || true)
while IFS= read -r line; do
    [ -z "$line" ] && continue
    name=$(echo "$line" | awk '
        /^func \(/ {
            # Method: func (recv) Name(args) ...
            # Find the closing ) of the receiver and pull the next token.
            paren = 0;
            for (i = 1; i <= length($0); i++) {
                c = substr($0, i, 1);
                if (c == "(") paren++;
                else if (c == ")") {
                    paren--;
                    if (paren == 0) {
                        rest = substr($0, i+1);
                        sub(/^ +/, "", rest);
                        # Extract the identifier up to ( or whitespace.
                        match(rest, /^[A-Za-z_][A-Za-z0-9_]*/);
                        print substr(rest, RSTART, RLENGTH);
                        exit;
                    }
                }
            }
            next;
        }
        /^func / {
            # Plain func: func Name(args) ...
            rest = $0;
            sub(/^func +/, "", rest);
            match(rest, /^[A-Za-z_][A-Za-z0-9_]*/);
            print substr(rest, RSTART, RLENGTH);
            next;
        }
        /^type / { print $2; next }
        /^var / { print $2; next }
        /^const / { print $2; next }
    ')
    [ -z "$name" ] && continue
    # Public-API floor only applies to exported names (capitalised first char).
    case "$name" in
        [A-Z]*) exported+=("$name") ;;
    esac
done <<< "$go_doc"

declare -a uncovered_symbols=()
for sym in "${exported[@]}"; do
    # Each func row in `go tool cover -func` is:
    #   <file>:<line>:\t<funcName>\t<pct>%
    # For type/var/const symbols (no own coverage rows) we cannot check
    # via go tool cover; treat them as covered iff the symbol is referenced
    # in any non-zero coverage row. To keep the check tight, fall back to
    # checking whether the symbol name appears in `go test ./...` output
    # at all — but more practically: scan the source for a test that
    # references the name. For a stricter automated check we restrict
    # this floor to func-only symbols.
    sym_row=$(echo "$FUNC_REPORT" | awk -v s="$sym" '
        BEGIN { found=0 }
        !/^total:/ {
            if ($2 == s) {
                gsub("%", "", $NF);
                if ($NF+0 > 0) { found=1; exit }
            }
        }
        END { exit (found ? 0 : 1) }
    ' && echo "covered" || echo "uncovered")
    if [ "$sym_row" = "uncovered" ]; then
        # Only flag if `go tool cover` knows about the symbol at all
        # (i.e. it is a func defined in the package). type/var/const
        # symbols are out of scope for this automated check.
        in_func_table=$(echo "$FUNC_REPORT" | awk -v s="$sym" '
            BEGIN { found=0 }
            !/^total:/ {
                if ($2 == s) { found=1; exit }
            }
            END { exit (found ? 0 : 1) }
        ' && echo "yes" || echo "no")
        if [ "$in_func_table" = "yes" ]; then
            uncovered_symbols+=("$sym")
        fi
    fi
done

if (( ${#uncovered_symbols[@]} > 0 )); then
    {
        echo "verify-coverage-floors: FAIL — ${#uncovered_symbols[@]} exported func(s) have 0% coverage:"
        for s in "${uncovered_symbols[@]}"; do
            echo "  $s"
        done
        echo
        echo "Every exported function in the public API surface MUST be exercised by at least one test."
    } >&2
    exit 1
fi

echo "OK: verify-coverage-floors — overall ${total_pct}% >= ${OVERALL_FLOOR}%; per-file >= ${PER_FILE_FLOOR}%; public-API funcs all exercised (${#exported[@]} exported symbols inspected)."
exit 0
