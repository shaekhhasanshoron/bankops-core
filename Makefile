# ------------------------------
# Root Makefile
# ------------------------------

# Basic toolchain variables
GO           ?= go
SHELL        := /usr/bin/env bash
CUR          := $(abspath .)

TOOLS_DIR    := $(CUR)/.tools/bin
export GOBIN := $(TOOLS_DIR)

# Tool versions
AIR_VERSION          ?= latest
GOLANGCI_VERSION     ?= v2.5.0
MIGRATE_VERSION      ?= v4.17.1

# Convenience paths to binaries
AIR                  := $(TOOLS_DIR)/air
GOLANGCI_LINT        := $(TOOLS_DIR)/golangci-lint
MIGRATE              := $(TOOLS_DIR)/migrate

# All service
SERVICES             := api-gateway auth-service account-service transaction-service event-tracking-service
MODULE_DIRS          := $(foreach s,$(SERVICES),$(if $(wildcard $(s)/go.mod),$(s),))

# Help output
.PHONY: help
help:
	@echo "bankops-core — Makefile"
	@echo
	@echo "Common helpers:"
	@echo "  make bootstrap         - install Air, golangci-lint, golang-migrate into ./.tools/bin"
	@echo "  make tidy              - run 'go mod tidy' for existing modules"
	@echo "  make lint              - run golangci-lint for existing modules"
	@echo "  make test              - run unit + integration tests with race detector"
	@echo
	@echo "Auth Service helpers:"
	@echo "  make run-auth          - run auth-service"
	@echo "  make air-auth          - run auth-service with Air"
	@echo "  make proto-gen-auth    - generate proto buf"


# 8) Install dev tools locally (Air, golangci-lint, migrate)
.PHONY: bootstrap
bootstrap: $(AIR) $(GOLANGCI_LINT)
	@echo "✔ tools ready in $(TOOLS_DIR)"

$(AIR):
	@mkdir -p $(TOOLS_DIR)
	@echo "→ installing Air ($(AIR_VERSION))"
	@$(GO) install github.com/air-verse/air@$(AIR_VERSION)

$(GOLANGCI_LINT):
	@mkdir -p $(TOOLS_DIR)
	@echo "→ installing golangci-lint ($(GOLANGCI_VERSION))"
	@$(GO) install github.com/golangci/golangci-lint/v2/cmd/golangci-lint@$(GOLANGCI_VERSION)


# go tidy for all modules
.PHONY: tidy
tidy:
	@set -e; \
	for d in $(MODULE_DIRS); do \
	  echo "→ go mod tidy in $$d"; \
	  (cd $$d && $(GO) mod tidy); \
	done

# Lint all modules
.PHONY: lint lint-fix
lint: $(GOLANGCI_LINT)
	@set -e; \
	for d in $(MODULE_DIRS); do \
	  echo "→ linting $$d"; \
	  (cd $$d && $(GOLANGCI_LINT) run ./...); \
	done

lint-fix: $(GOLANGCI_LINT)
	@set -e; \
	for d in $(MODULE_DIRS); do \
	  echo "→ linting $$d"; \
	  (cd $$d && $(GOLANGCI_LINT) run ./... --fix); \
	done

# Test all modules with race detector
.PHONY: test
test:
	@set -e; \
	for d in $(MODULE_DIRS); do \
	  echo "→ testing $$d"; \
	  (cd $$d && $(GO) test ./... -race -count=1 -timeout=5m); \
	done

# run/air/migrate for auth-service
.PHONY: air-auth run-auth proto-gen-auth
air-auth: $(AIR)
	@cd auth-service && $(AIR) -c .air.toml

run-auth:
	@echo "→ running auth-service"
	@cd auth-service && $(GO) run ./cmd/authsvc

proto-gen-auth:
	 @protoc \
	 --proto_path=auth-service/api/proto auth-service/api/proto/*.proto \
	 --go_out=./auth-service/api/proto \
	 --go_opt=paths=source_relative \
	 --go-grpc_out=./auth-service/api/proto \
	 --go-grpc_opt=paths=source_relative