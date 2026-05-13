# fuzzymatch — canonical Makefile.
#
# All 19 targets enumerated in CLAUDE.md "Makefile Targets". `make check` is
# the aggregate local pre-PR gate; CI mirrors it from .github/workflows/ci.yml.
# Targets that depend on artefacts produced by later plans (01-03 release
# pipeline, 01-04 determinism infra) are tolerant no-ops printing a "pending"
# message until those plans land.

SHELL := /usr/bin/env bash

GO          ?= go
GOFMT       ?= gofmt
GOIMPORTS   ?= goimports
GOLANGCILINT?= golangci-lint
GOVULNCHECK ?= govulncheck
BENCHSTAT   ?= benchstat
GORELEASER  ?= goreleaser

MODULE_PATH      := github.com/axonops/fuzzymatch
BDD_DIR          := tests/bdd
COVERAGE_FILE    := coverage.out
BENCH_FILE       := bench.txt
BENCH_NEW_FILE   := bench.txt.new
COVERAGE_FLOOR   := 95.0

.PHONY: check test test-bdd test-fuzz lint vet fmt fmt-check bench bench-compare \
	coverage coverage-check tidy tidy-check security verify-deps-allowlist \
	verify-determinism verify-license-headers release-check clean

# `check` — the canonical pre-PR aggregate gate. CI runs the same target.
check: fmt-check vet lint verify-license-headers verify-deps-allowlist tidy-check security test coverage coverage-check
	@echo "OK: make check passed."

# ---- test targets ---------------------------------------------------------

test:
	$(GO) test -race -shuffle=on -count=1 ./...

test-bdd:
	cd $(BDD_DIR) && $(GO) test -race -count=1 ./...

# Discovers fuzzers via `go test -list 'Fuzz.*'`. If no fuzzers exist, prints
# a friendly note and exits 0 — `make check` does not depend on this target.
test-fuzz:
	@found=0; \
	for pkg in $$($(GO) list ./...); do \
	  if $(GO) test -list 'Fuzz.*' "$$pkg" 2>/dev/null | grep -q '^Fuzz'; then \
	    found=1; \
	    $(GO) test -run='^$$' -fuzz='Fuzz.*' -fuzztime=60s "$$pkg" || exit 1; \
	  fi; \
	done; \
	if [ "$$found" -eq 0 ]; then \
	  echo "no fuzzers found; skipping (pending implementation)."; \
	fi

# ---- lint / vet / fmt -----------------------------------------------------

lint:
	$(GOLANGCILINT) run ./...
	cd $(BDD_DIR) && $(GOLANGCILINT) run ./...

vet:
	$(GO) vet ./...
	cd $(BDD_DIR) && $(GO) vet ./...

fmt:
	$(GOFMT) -s -w .
	$(GOIMPORTS) -local $(MODULE_PATH) -w .

# Asserts the tree is fmt-clean. Any diff from gofmt -s or goimports fails.
fmt-check:
	@out=$$($(GOFMT) -s -l . 2>&1); \
	if [ -n "$$out" ]; then \
	  echo "fmt-check: gofmt -s would reformat:"; \
	  echo "$$out"; \
	  exit 1; \
	fi
	@out=$$($(GOIMPORTS) -local $(MODULE_PATH) -l . 2>&1); \
	if [ -n "$$out" ]; then \
	  echo "fmt-check: goimports would reformat:"; \
	  echo "$$out"; \
	  exit 1; \
	fi
	@echo "OK: fmt-check passed."

# ---- benchmarks -----------------------------------------------------------

# Runs `go test -bench=. -benchmem -count=10`. Tolerant if no benchmarks
# exist yet (Phase 2+ adds them) — prints message and exits 0.
bench:
	@found=0; \
	for pkg in $$($(GO) list ./...); do \
	  if $(GO) test -list 'Benchmark.*' "$$pkg" 2>/dev/null | grep -q '^Benchmark'; then \
	    found=1; break; \
	  fi; \
	done; \
	if [ "$$found" -eq 0 ]; then \
	  echo "no benchmarks found; skipping (pending Phase 2)."; \
	else \
	  $(GO) test -bench=. -benchmem -count=10 ./... | tee $(BENCH_NEW_FILE); \
	fi

# Compares committed bench.txt against bench.txt.new from `make bench`.
# Plan 01-04 lands the full regression harness; until then this is a thin
# pass-through.
bench-compare:
	@if [ -f $(BENCH_FILE) ] && [ -f $(BENCH_NEW_FILE) ]; then \
	  if command -v $(BENCHSTAT) >/dev/null 2>&1; then \
	    $(BENCHSTAT) $(BENCH_FILE) $(BENCH_NEW_FILE); \
	  else \
	    echo "benchstat not installed; install: go install golang.org/x/perf/cmd/benchstat@latest"; \
	  fi; \
	else \
	  echo "pending: bench.txt or bench.txt.new not present; run 'make bench' first."; \
	fi

# ---- coverage -------------------------------------------------------------

coverage:
	$(GO) test -race -coverprofile=$(COVERAGE_FILE) -covermode=atomic ./...

# Asserts overall coverage >= COVERAGE_FLOOR%. Per-file and public-API
# enforcement lands in plan 01-04 (requires a dedicated parser script).
# Tolerant: if coverage.out is absent (fresh clone) or contains no profiled
# lines (no tests yet), print a pending message and exit 0 — `make check`
# chains `coverage` immediately before `coverage-check`.
coverage-check:
	@if [ ! -f $(COVERAGE_FILE) ]; then \
	  echo "coverage-check: $(COVERAGE_FILE) not present; skipping (run 'make coverage' first)."; \
	  exit 0; \
	fi; \
	profiled_lines=$$(awk 'BEGIN{n=0} !/^mode:/{n++} END{print n}' $(COVERAGE_FILE)); \
	if [ "$$profiled_lines" -eq 0 ]; then \
	  echo "coverage-check: no profiled lines in $(COVERAGE_FILE) (no tests yet); skipping (pending Phase 2)."; \
	  exit 0; \
	fi; \
	total=$$($(GO) tool cover -func=$(COVERAGE_FILE) | awk '/^total:/ {gsub("%", "", $$3); print $$3}'); \
	if [ -z "$$total" ]; then \
	  echo "coverage-check: no coverage data; skipping."; \
	  exit 0; \
	fi; \
	awk -v t="$$total" -v f="$(COVERAGE_FLOOR)" 'BEGIN { if (t+0 < f+0) { printf "coverage-check: total %s%% below floor %s%%\n", t, f; exit 1 } else { printf "OK: coverage %s%% >= %s%%\n", t, f; exit 0 } }'

# ---- modules --------------------------------------------------------------

tidy:
	$(GO) mod tidy
	cd $(BDD_DIR) && $(GO) mod tidy

# Verifies `go mod tidy` is a no-op in both modules; any diff fails CI.
tidy-check:
	$(GO) mod tidy
	cd $(BDD_DIR) && $(GO) mod tidy
	@git diff --exit-code -- go.mod go.sum $(BDD_DIR)/go.mod $(BDD_DIR)/go.sum

# ---- security / vuln ------------------------------------------------------

# Local-developer focus: govulncheck. The definitive gosec scan runs in
# .github/workflows/security.yml with SARIF upload. Tolerant of missing tools.
security:
	@if command -v $(GOVULNCHECK) >/dev/null 2>&1; then \
	  $(GOVULNCHECK) ./...; \
	else \
	  echo "govulncheck not installed; install: go install golang.org/x/vuln/cmd/govulncheck@latest"; \
	fi

# ---- verify-* (CI gates) --------------------------------------------------

# Plan 01-04 lands scripts/verify-no-runtime-deps.sh. Until then this
# target is a tolerant no-op printing a pending message.
verify-deps-allowlist:
	@if [ -x scripts/verify-no-runtime-deps.sh ]; then \
	  bash scripts/verify-no-runtime-deps.sh; \
	else \
	  echo "pending plan 01-04: scripts/verify-no-runtime-deps.sh not yet present."; \
	fi

# Plan 01-04 lands the golden-form determinism harness. Until then this
# target runs `go test -run TestGolden_` which currently matches no tests.
verify-determinism:
	$(GO) test -run TestGolden_ ./...

verify-license-headers:
	bash scripts/verify-license-headers.sh

# Plan 01-03 lands .goreleaser.yml. Until then this is a tolerant no-op.
release-check:
	@if [ -f .goreleaser.yml ]; then \
	  if command -v $(GORELEASER) >/dev/null 2>&1; then \
	    $(GORELEASER) check; \
	  else \
	    echo "goreleaser not installed; install per docs/CONTRIBUTING (plan 01-08)."; \
	  fi; \
	else \
	  echo "pending plan 01-03: .goreleaser.yml not yet present."; \
	fi

# ---- housekeeping ---------------------------------------------------------

clean:
	$(GO) clean ./...
	rm -f $(COVERAGE_FILE) $(BENCH_NEW_FILE)
