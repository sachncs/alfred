SHELL := /bin/bash
.DEFAULT_GOAL := help

GO         ?= go
GOFLAGS    ?= -trimpath
BIN_DIR    ?= bin
APP_NAME   ?= alfred
PKG        := ./...
COVERAGE   := coverage.out

.PHONY: help
help: ## Show this help.
	@awk 'BEGIN {FS = ":.*##"; printf "Usage:\n  make \033[36m<target>\033[0m\n\nTargets:\n"} \
	  /^[a-zA-Z_\.\-]+:.*?##/ { printf "  \033[36m%-15s\033[0m %s\n", $$1, $$2 }' $(MAKEFILE_LIST)

.PHONY: build
build: ## Compile the application into ./bin/alfred
	@mkdir -p $(BIN_DIR)
	$(GO) build $(GOFLAGS) -o $(BIN_DIR)/$(APP_NAME) ./cmd/$(APP_NAME)
	@echo "Built $(BIN_DIR)/$(APP_NAME)"

.PHONY: test
test: ## Run all package tests
	$(GO) test $(GOFLAGS) -race -count=1 $(PKG)
	@echo "All tests passed"

.PHONY: test-cover
test-cover: ## Run all package tests with coverage
	$(GO) test $(GOFLAGS) -race -count=1 -coverprofile=$(COVERAGE) $(PKG)
	$(GO) tool cover -func=$(COVERAGE) | tail -1
	@rm -f $(COVERAGE)

.PHONY: vet
vet: ## Run go vet
	$(GO) vet $(PKG)
	@echo "go vet clean"

.PHONY: lint
lint: ## Run golangci-lint
	golangci-lint run $(PKG)
	@echo "lint clean"

.PHONY: run
run: build ## Run the application
	./$(BIN_DIR)/$(APP_NAME)

.PHONY: clean
clean: ## Remove build artifacts
	rm -rf $(BIN_DIR) $(COVERAGE)
	$(GO) clean -testcache

.PHONY: tidy
tidy: ## Tidy go.mod and go.sum
	$(GO) mod tidy

.PHONY: fmt
fmt: ## Run gofmt
	$(GO) fmt $(PKG)
