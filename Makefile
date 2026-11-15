# SPDX-License-Identifier: Apache-2.0 OR MIT
#
# `make check` runs every local gate CI runs without a server; `make
# conformance` adds the live ones and needs a running crypto-server
# (CRYPTO_SERVICE_URL, CRYPTO_SERVICE_API_KEY, CRYPTO_SERVICE_JWT_SECRET);
# `make regression` starts the pinned server itself and runs both. The
# analysers are fetched as Go modules at pinned versions by `go run`;
# nothing is installed globally.

STATICCHECK := honnef.co/go/tools/cmd/staticcheck@v0.8.2
GOCYCLO     := github.com/fzipp/gocyclo/cmd/gocyclo@v0.6.0
GOCOGNIT    := github.com/uudashr/gocognit/cmd/gocognit@v1.3.0
GOVULNCHECK := golang.org/x/vuln/cmd/govulncheck@v1.9.0

.PHONY: check versions fmt-check fmt vet staticcheck complexity vulncheck build test api-check conformance live examples readme bench regression docs docs-lint fuzz demo

check: versions fmt-check vet staticcheck complexity vulncheck build test

# Every file that restates the version agrees with version.go, and the
# server pin is a full commit SHA (scripts/verify-release-versions).
versions:
	scripts/verify-release-versions

fmt-check:
	@out="$$(gofmt -l .)"; if [ -n "$$out" ]; then echo "gofmt needed:"; echo "$$out"; exit 1; fi

fmt:
	gofmt -w .

vet:
	go vet ./...

staticcheck:
	go run $(STATICCHECK) ./...

# Known vulnerabilities reachable from this module: the module is
# stdlib-only, so this covers the Go standard library in use. Needs
# vuln.go.dev, or VULNDB=file:///path for a local copy.
VULNDB ?=

vulncheck:
	go run $(GOVULNCHECK) $(if $(VULNDB),-db $(VULNDB)) ./...

# Cyclomatic <= 10 and cognitive <= 15 per function; size_test.go holds
# 60 lines per function and 500 per file, halstead_test.go a Halstead
# difficulty of 30.
complexity:
	go run $(GOCYCLO) -over 10 .
	go run $(GOCOGNIT) -over 15 .

build:
	go build ./...

# Race-enabled tests of the library and internal/kit; fails below
# 100.0% of statements (scripts/coverage.sh). The live tests skip here.
test:
	scripts/coverage.sh

# gorelease against the last release tag (scripts/api-check.sh).
api-check:
	scripts/api-check.sh

# The conformance kit, the typed live tests, every example and every
# README code block against a running server. Fails, rather than skips,
# when the server variables are missing.
conformance: live examples readme

live:
	CRYPTO_SERVICE_REQUIRE_LIVE=1 go test -race -count=1 -v -run 'TestConformance|TestLive' .

# Builds every program under examples/ into bin/ (gitignored) and runs
# each against CRYPTO_SERVICE_URL; each exits non-zero when the server
# does not answer as it expects.
EXAMPLES := $(notdir $(patsubst %/,%,$(dir $(wildcard examples/*/main.go))))

examples:
	@test -n "$$CRYPTO_SERVICE_URL" || { echo "examples: CRYPTO_SERVICE_URL is unset" >&2; exit 1; }
	mkdir -p bin
	go build -o bin/ ./examples/...
	@set -e; for name in $(EXAMPLES); do echo "example: $$name"; bin/$$name; done

# Extracts every fenced code block of README.md and runs it
# (scripts/readme-snippets.sh).
readme:
	scripts/readme-snippets.sh

# testing.B benchmarks under the shared scenario names: encode_request,
# decode_response, decode_problem (no network) and round_trip_hash
# (CRYPTO_SERVICE_URL, skipped when unset). Output to bin/bench.txt,
# benchstat input. CI smoke-runs it with BENCHCOUNT=1 BENCHTIME=1x.
BENCHCOUNT ?= 6
BENCHTIME ?= 1s

bench:
	mkdir -p bin
	go test -run '^$$' -bench . -benchmem -count $(BENCHCOUNT) -benchtime $(BENCHTIME) . | tee bin/bench.txt

# The full regression run CI's conformance job makes: starts the pinned
# crypto-server (CRYPTO_SERVICE_SERVER_DIR, a built checkout of the
# commit in ci/server.ref) on a free loopback port, runs make check,
# make conformance and a benchmark smoke against it, and stops it.
regression:
	scripts/regression.sh

# Each fuzz target for FUZZTIME (plain `go test` replays the seeds and
# testdata/fuzz). A crash leaves its input in testdata/fuzz; commit it
# as a regression with the fix.
FUZZTIME ?= 60s

fuzz:
	@set -e; targets="$$(go test -list '^Fuzz' . | grep '^Fuzz')"; \
	test -n "$$targets"; \
	for target in $$targets; do \
		echo "fuzz: $$target for $(FUZZTIME)"; \
		go test -run='^$$' -fuzz="^$$target$$" -fuzztime=$(FUZZTIME) . ; \
	done

# The user manual (mkdocs.yml), built strictly into site/ (gitignored),
# then checked for the author link on every page. Needs MkDocs Material
# from ci/docs-requirements.txt on PATH.
docs:
	mkdocs build --strict
	scripts/check-site-footer.sh site

# The Docs lint workflow: markdownlint, codespell and REUSE.
docs-lint:
	npx --yes markdownlint-cli2@0.23.3 "**/*.md"
	codespell
	reuse lint

# The README demo, .github/demo.gif, rendered by vhs from
# .github/demo.tape against a server already running at
# CRYPTO_SERVICE_URL.
demo:
	@test -n "$$CRYPTO_SERVICE_URL" || { echo "demo: start a server and set CRYPTO_SERVICE_URL" >&2; exit 1; }
	vhs .github/demo.tape
