# ------------------------------
# Root Makefile
# ------------------------------

# Basic toolchain variables
GO           ?= go
SHELL        := /usr/bin/env bash
DOCKER_REGISTRY        := shaekhhasan
CUR          := $(abspath .)

TOOLS_DIR    := $(CUR)/.tools/bin
export GOBIN := $(TOOLS_DIR)

# Tool versions
AIR_VERSION          ?= latest
GOLANGCI_VERSION     ?= v2.5.0
SWAGGER_VERSION      ?= latest

# Convenience paths to binaries
AIR                  := $(TOOLS_DIR)/air
GOLANGCI_LINT        := $(TOOLS_DIR)/golangci-lint
MIGRATE              := $(TOOLS_DIR)/migrate

# All service
SERVICES             := gateway-service auth-service account-service
MODULE_DIRS          := $(foreach s,$(SERVICES),$(if $(wildcard $(s)/go.mod),$(s),))

# Help output
.PHONY: help
help:
	@echo "bankops-core — Makefile"
	@echo
	@echo "Common helpers:"
	@echo "  make bootstrap         - install Air, golangci-lint, golang-migrate, swagger into ./.tools/bin"
	@echo "  make tidy              - run 'go mod tidy' for existing modules"
	@echo "  make lint              - run golangci-lint for existing modules"
	@echo "  make test              - run unit + integration tests with race detector"
	@echo "  make docs              - generate swagger docs"
	@echo "  make docker-build      - docker build all services"
	@echo "  make docker-push       - docker push all services"
	@echo
	@echo "Auth service helpers:"
	@echo "  make run-auth          - run auth-service"
	@echo "  make air-auth          - run auth-service with Air"
	@echo "  make proto-auth        - generate proto buf to all services"
	@echo "  make proto-auth-only   - generate proto buf for auth service"
	@echo
	@echo "Gateway service helpers:"
	@echo "  make run-gateway       - run gateway-service"
	@echo
	@echo "Account service helpers:"
	@echo "  make run-account        - run account-service"
	@echo "  make air-account        - run account-service with Air"
	@echo "  make proto-gateway      - generate proto buf to all services"
	@echo "  make proto-account-only - generate proto buf for account service"



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

$(SWAGGER):
	@mkdir -p $(TOOLS_DIR)
	@echo "→ installing swagger ($(SWAGGER_VERSION))"
	@$(GO) install github.com/swaggo/swag/cmd/swag@$(SWAGGER_VERSION)

docs:
	@echo "→ running auth-service"
	@cd gateway-service && swag init -g cmd/gatewaysvc/main.go --output api/docs

# go tidy for all modules
.PHONY: tidy
tidy:
	@set -e; \
	for d in $(MODULE_DIRS); do \
	  echo "→ go mod tidy in $$d"; \
	  (cd $$d && $(GO) mod tidy); \
	done

.PHONY: docker-build docker-push
docker-build:
	@set -e; \
	for d in $(MODULE_DIRS); do \
	  echo "→ docker build $$d"; \
	  (cd $$d && docker build -t $(DOCKER_REGISTRY)/bankapp-core:$$d-v1.0.0 .); \
	done

docker-push:
	@set -e; \
	for d in $(MODULE_DIRS); do \
	  echo "→ docker push $$d"; \
	  (cd $$d && docker push $(DOCKER_REGISTRY)/bankapp-core:$$d-v1.0.0); \
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

# run/air for auth-service
.PHONY: air-auth run-auth proto-auth proto-auth-all
air-auth: $(AIR)
	@cd auth-service && $(AIR) -c .air.toml

run-auth:
	@echo "→ running auth-service"
	@cd auth-service && $(GO) run ./cmd/authsvc

proto-auth-only:
	 @protoc \
	 --proto_path=auth-service/api/proto auth-service/api/proto/*.proto \
	 --go_out=./auth-service/api \
	 --go_opt=paths=import \
	 --go-grpc_out=./auth-service/api \
	 --go-grpc_opt=paths=import

# Generate proto files for all services
proto-auth:
	 @protoc \
	 --proto_path=auth-service/api/proto \
	 auth-service/api/proto/*.proto \
	 --go_out=./auth-service/api \
	 --go_opt=paths=import \
	 --go-grpc_out=./auth-service/api \
	 --go-grpc_opt=paths=import
	@protoc \
		--proto_path=auth-service/api/proto \
		auth-service/api/proto/*.proto \
		--go_out=./gateway-service/api \
		--go_opt=paths=import \
		--go-grpc_out=./gateway-service/api \
		--go-grpc_opt=paths=import

.PHONY: coverage-auth
coverage-auth:
	@echo "→ coverage of auth-service"
	@cd auth-service && $(GO) test -cover ./...

.PHONY: docker-build-auth docker-push-auth
docker-build-auth:
	@echo "→ running docker build for auth-service"
	@cd auth-service && docker build -t $(DOCKER_REGISTRY)/bankapp-core:auth-service-v1.0.0 .

docker-push-auth:
	@echo "→ running docker push for auth-service"
	@cd auth-service && docker push $(DOCKER_REGISTRY)/bankapp-core:auth-service-v1.0.0

# run/air for gateway-service
.PHONY: air-gateway run-gateway
run-gateway:
	@echo "→ running gateway-service"
	@cd gateway-service && $(GO) run ./cmd/gatewaysvc

air-gateway: $(AIR)
	@cd gateway-service && $(AIR) -c .air.toml

test-gateway:
	@echo "→ testing gateway-service"
	@cd gateway-service && $(GO) test ./...

.PHONY: docker-build-gateway docker-push-gateway
docker-build-gateway:
	@echo "→ running docker build for gateway-service"
	@cd gateway-service && docker build -t $(DOCKER_REGISTRY)/bankapp-core:gateway-service-v1.0.0 .

docker-push-gateway:
	@echo "→ running docker push for gateway-service"
	@cd gateway-service && docker push $(DOCKER_REGISTRY)/bankapp-core:gateway-service-v1.0.0

# run/air for account-service
.PHONY: air-account run-account proto-account proto-account-all
run-account:
	@echo "→ running account-service"
	@cd account-service && $(GO) run ./cmd/accountsvc

air-account: $(AIR)
	@cd account-service && $(AIR) -c .air.toml

#test-account:
#	@echo "→ testing account-service"
#	@cd account-service && $(GO) test ./...

test-account:
	@echo "→ testing account-service"
	@cd account-service && cd internal && cd app && $(GO) test ./...
	@cd account-service && cd internal && cd grpc && cd account_handler && $(GO) test ./...

coverage-account:
	@echo "→ coverage of account-service"
	@cd account-service && $(GO) test -cover ./...

#coverage-account:
#	@echo "→ coverage of account-service business/service domain"
#	@cd account-service && cd internal && cd app && $(GO) test -cover ./...
#	@echo "→ coverage of account-service handler domain"
#	@cd account-service && cd internal && cd grpc && cd account_handler && $(GO) test -cover ./...


proto-account-only:
	@protoc \
     		--proto_path=account-service/api/proto \
     		--go_out=./account-service/api \
     		--go_opt=paths=import \
     		--go-grpc_out=./account-service/api \
     		--go-grpc_opt=paths=import \
     		account-service/api/proto/common/*.proto \
     		account-service/api/proto/account/*.proto \
     		account-service/api/proto/transaction/*.proto \
     		account-service/api/proto/customer/*.proto \
     		account-service/api/proto/*.proto

# Generate proto files for all services
proto-account:
	@protoc \
     		--proto_path=account-service/api/proto \
     		--go_out=./account-service/api \
     		--go_opt=paths=import \
     		--go-grpc_out=./account-service/api \
     		--go-grpc_opt=paths=import \
     		account-service/api/proto/common/*.proto \
     		account-service/api/proto/account/*.proto \
     		account-service/api/proto/customer/*.proto \
     		account-service/api/proto/transaction/*.proto \
     		account-service/api/proto/*.proto
	@protoc \
     		--proto_path=account-service/api/proto \
     		--go_out=./gateway-service/api \
     		--go_opt=paths=import \
     		--go-grpc_out=./gateway-service/api \
     		--go-grpc_opt=paths=import \
     		account-service/api/proto/common/*.proto \
     		account-service/api/proto/account/*.proto \
     		account-service/api/proto/transaction/*.proto \
     		account-service/api/proto/customer/*.proto \
     		account-service/api/proto/*.proto

.PHONY: docker-build-account docker-push-account
docker-build-account:
	@echo "→ running docker build for account-service"
	@cd account-service && docker build -t $(DOCKER_REGISTRY)/bankapp-core:account-service-v1.0.0 .

docker-push-account:
	@echo "→ running docker push for account-service"
	@cd account-service && docker push $(DOCKER_REGISTRY)/bankapp-core:account-service-v1.0.0