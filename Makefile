# Copyright © 2026 envaar
# SPDX-License-Identifier: Apache-2.0

SHELL := /bin/sh

GO ?= go
GORELEASER_BIN ?= $(shell if [ -x /snap/goreleaser/current/goreleaser ]; then printf '%s' /snap/goreleaser/current/goreleaser; else printf '%s' goreleaser; fi)
BIN_DIR ?= bin
BIN := $(BIN_DIR)/vaar
GOCACHE ?= /tmp/go-build

export GOCACHE

.DEFAULT_GOAL := help

.PHONY: help test vet lint bench build snapshot release clean

help: ## Show the available maintainer commands.
	@printf '%s\n' "Available targets:" \
		"  make test     - run the Go test suite" \
		"  make vet      - run go vet" \
		"  make lint     - run gofmt and vaar lint in a clean temp mirror" \
		"  make bench    - run the lint smoke benchmark" \
		"  make build    - build bin/vaar" \
		"  make snapshot - build snapshot release artifacts" \
		"  make release  - build real release artifacts from a git tag" \
		"  make clean    - remove local build output"

test: ## Run the Go test suite.
	$(GO) test ./...

vet: ## Run go vet.
	$(GO) vet ./...

lint: ## Check formatting and run the CLI lint smoke test.
	@if [ -n "$$(gofmt -l .)" ]; then \
		printf '%s\n' "gofmt found unformatted Go files:"; \
		gofmt -l .; \
		exit 1; \
	fi
	@tmpdir="$$(mktemp -d)"; \
	trap 'rm -rf "$$tmpdir"' EXIT INT TERM; \
	rsync -a --exclude='.git' --exclude='bin' --exclude='dist' --exclude='.env' --exclude='examples' ./ "$$tmpdir"/; \
	cd "$$tmpdir" && $(GO) run ./cmd/vaar lint

bench: ## Run the lint smoke benchmark.
	$(GO) test ./internal/lint -run '^$$' -bench '^BenchmarkRunnerSmoke$$' -benchmem

build: ## Build the CLI binary into bin/vaar.
	@mkdir -p $(BIN_DIR)
	$(GO) build -buildvcs=false -o $(BIN) ./cmd/vaar

snapshot: ## Build snapshot release artifacts with GoReleaser.
	@goreleaser_bin="$(GORELEASER_BIN)"; \
	if [ -z "$$goreleaser_bin" ]; then \
		printf '%s\n' "GoReleaser is not available. Set GORELEASER_BIN or install goreleaser."; \
		exit 1; \
	fi; \
	"$$goreleaser_bin" release --snapshot --clean

release: ## Build real release artifacts with GoReleaser from a git tag.
	@goreleaser_bin="$(GORELEASER_BIN)"; \
	if [ -z "$$goreleaser_bin" ]; then \
		printf '%s\n' "GoReleaser is not available. Set GORELEASER_BIN or install goreleaser."; \
		exit 1; \
	fi; \
	tag="$$(git describe --tags --exact-match 2>/dev/null || true)"; \
	if [ -z "$$tag" ]; then \
		printf '%s\n' "make release must be run from an exact version tag, such as v0.1.0."; \
		exit 1; \
	fi; \
	"$$goreleaser_bin" release --clean

clean: ## Remove local build artifacts.
	rm -rf $(BIN_DIR) dist
