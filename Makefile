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
SERVICES             := gateway-service auth-service account-service transaction-service
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
	@echo "  make test              - generate test for all services"
	@echo "  make coverage          - generate test coverage for all services"
	@echo "  make docs              - generate swagger docs"
	@echo "  make docker-build      - docker build all services"
	@echo "  make docker-push       - docker push all services"
	@echo "  make compose-up        - docker-compose up all services"
	@echo "  make compose-down      - docker-compose down all services"
	@echo "  make proto             - generate proto files for all services"
	@echo
	@echo "Auth service helpers:"
	@echo "  make run-auth          - run auth service"
	@echo "  make run-auth-dev      - run auth service (dev-mode)"
	@echo "  make air-auth          - run auth-service with Air"
	@echo "  make proto-auth        - generate proto buf to all services"
	@echo
	@echo "Gateway service helpers:"
	@echo "  make run-gateway       - run gateway-service"
	@echo "  make run-gateway-dev   - run gateway-service (dev-mode)"
	@echo
	@echo "Account service helpers:"
	@echo "  make run-account        - run account service"
	@echo "  make run-account-dev    - run account service (dev-mode)"
	@echo "  make air-account        - run account-service with Air"
	@echo "  make proto-account      - generate proto buf to all services"
	@echo
	@echo "Transaction service helpers:"
	@echo "  make run-tx        - run transaction service"
	@echo "  make run-tx-dev    - run transaction service (dev-mode)"
	@echo "  make air-tx        - run transaction service with Air"
	@echo "  make proto-tx      - generate proto buf to all services"



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

compose-up:
	@echo "→ running service on docker"
	@cd deployment && cd docker && docker-compose up

compose-down:
	@echo "→ running service on docker"
	@cd deployment && cd docker && docker-compose down

.PHONY: proto coverage test
proto: proto-account proto-auth proto-tx

coverage: coverage-gateway coverage-account coverage-auth coverage-tx

test: test-gateway test-account test-auth test-tx

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
.PHONY: air-auth run-auth proto-auth proto-auth-all run-auth-dev
air-auth: $(AIR)
	@cd auth-service && $(AIR) -c .air.toml

run-auth:
	@echo "→ running auth-service"
	@cd auth-service && $(GO) run ./cmd/authsvc

run-auth-dev:
	@echo "→ running auth-service (dev-mode)"
	@cd auth-service && AUTH_ENV=dev $(GO) run ./cmd/authsvc

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

.PHONY: test-auth coverage-auth
test-auth:
	@echo "→ test of auth-service"
	@cd auth-service && cd internal && cd app && $(GO) test ./...

coverage-auth:
	@echo "→ coverage of auth-service"
	@cd auth-service && cd internal && cd app && $(GO) test -cover ./...

.PHONY: docker-build-auth docker-push-auth
docker-build-auth:
	@echo "→ running docker build for auth-service"
	@cd auth-service && docker build -t $(DOCKER_REGISTRY)/bankapp-core:auth-service-v1.0.0 .

docker-push-auth:
	@echo "→ running docker push for auth-service"
	@cd auth-service && docker push $(DOCKER_REGISTRY)/bankapp-core:auth-service-v1.0.0

# run/air for gateway-service
.PHONY: air-gateway run-gateway run-gateway-dev
run-gateway:
	@echo "→ running gateway-service"
	@cd gateway-service && $(GO) run ./cmd/gatewaysvc

run-gateway-dev:
	@echo "→ running gateway-service (dev-mode)"
	@cd gateway-service && GATEWAY_ENV=dev $(GO) run ./cmd/gatewaysvc

air-gateway: $(AIR)
	@cd gateway-service && $(AIR) -c .air.toml

.PHONY: coverage-auth test-gateway
test-gateway:
	@echo "→ testing of gateway-service"
	@cd gateway-service && $(GO) test ./...

coverage-gateway:
	@echo "→ coverage of gateway-service"
	@cd gateway-service && cd internal && cd http && cd handlers && $(GO) test -cover ./...

.PHONY: docker-build-gateway docker-push-gateway
docker-build-gateway:
	@echo "→ running docker build for gateway-service"
	@cd gateway-service && docker build -t $(DOCKER_REGISTRY)/bankapp-core:gateway-service-v1.0.0 .

docker-push-gateway:
	@echo "→ running docker push for gateway-service"
	@cd gateway-service && docker push $(DOCKER_REGISTRY)/bankapp-core:gateway-service-v1.0.0

# run/air for account-service
.PHONY: air-account run-account proto-account proto-account-all run-account-dev
run-account:
	@echo "→ running account-service"
	@cd account-service && $(GO) run ./cmd/accountsvc

run-account-dev:
	@echo "→ running account-service (dev-mode)"
	@cd account-service && ACCOUNT_ENV=dev $(GO) run ./cmd/accountsvc

air-account: $(AIR)
	@cd account-service && $(AIR) -c .air.toml

test-account:
	@echo "→ testing account-service"
	@cd account-service && cd internal && cd app && $(GO) test ./...
	@#cd account-service && cd internal && cd grpc && cd account_handler && $(GO) test ./...

coverage-account:
	@echo "→ coverage of account-service business/service domain"
	@cd account-service && cd internal && cd app && $(GO) test -cover ./...

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
     		account-service/api/proto/transaction_saga/*.proto \
     		account-service/api/proto/*.proto
	@protoc \
     		--proto_path=account-service/api/proto \
     		--go_out=./gateway-service/api \
     		--go_opt=paths=import \
     		--go-grpc_out=./gateway-service/api \
     		--go-grpc_opt=paths=import \
     		account-service/api/proto/common/*.proto \
     		account-service/api/proto/account/*.proto \
     		account-service/api/proto/transaction_saga/*.proto \
     		account-service/api/proto/customer/*.proto \
     		account-service/api/proto/*.proto
	@protoc \
     		--proto_path=account-service/api/proto \
     		--go_out=./transaction-service/api \
     		--go_opt=paths=import \
     		--go-grpc_out=./transaction-service/api \
     		--go-grpc_opt=paths=import \
     		account-service/api/proto/common/*.proto \
     		account-service/api/proto/account/*.proto \
     		account-service/api/proto/transaction_saga/*.proto \
     		account-service/api/proto/customer/*.proto \
     		account-service/api/proto/*.proto

.PHONY: docker-build-account docker-push-account
docker-build-account:
	@echo "→ running docker build for account-service"
	@cd account-service && docker build -t $(DOCKER_REGISTRY)/bankapp-core:account-service-v1.0.0 .

docker-push-account:
	@echo "→ running docker push for account-service"
	@cd account-service && docker push $(DOCKER_REGISTRY)/bankapp-core:account-service-v1.0.0

.PHONY: proto-tx
proto-tx:
	@protoc \
     		--proto_path=transaction-service/api/proto \
     		--go_out=./transaction-service/api \
     		--go_opt=paths=import \
     		--go-grpc_out=./transaction-service/api \
     		--go-grpc_opt=paths=import \
     		transaction-service/api/proto/tx_common/*.proto \
     		transaction-service/api/proto/*.proto
	@protoc \
     		--proto_path=transaction-service/api/proto \
     		--go_out=./gateway-service/api \
     		--go_opt=paths=import \
     		--go-grpc_out=./gateway-service/api \
     		--go-grpc_opt=paths=import \
     		transaction-service/api/proto/tx_common/*.proto \
     		transaction-service/api/proto/*.proto

.PHONY: air-tx run-tx run-tx-dev
air-tx: $(AIR)
	@cd transaction-service && $(AIR) -c .air.toml

run-tx:
	@echo "→ running transaction-service"
	@cd transaction-service && $(GO) run ./cmd/txsvc

run-tx-dev:
	@echo "→ running transaction-service (dev-mode)"
	@cd transaction-service && TRANSACTION_ENV=dev $(GO) run ./cmd/txsvc

.PHONY: test-tx coverage-tx
test-tx:
	@echo "→ test of transaction-service"
	@cd transaction-service && cd internal && cd app && $(GO) test  ./...

coverage-tx:
	@echo "→ coverage of transaction-service"
	@cd transaction-service && cd internal && cd app && $(GO) test -cover ./...

.PHONY: docker-build-tx docker-push-tx
docker-build-tx:
	@echo "→ running docker build for transaction-service"
	@cd transaction-service && docker build -t $(DOCKER_REGISTRY)/bankapp-core:transaction-service-v1.0.0 .

docker-push-tx:
	@echo "→ running docker push for transaction-service"
	@cd transaction-service && docker push $(DOCKER_REGISTRY)/bankapp-core:transaction-service-v1.0.0