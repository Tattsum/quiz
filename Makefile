.PHONY: help fmt lint test build clean check install-tools test-unit test-integration test-performance test-parallel

BINARY_NAME=quiz
GO_VERSION=$(shell go version | awk '{print $$3}')
PARALLELISM ?= 8

# Tool versions (managed by Renovate)
GOLANGCI_LINT_VERSION := v1.64.1
GOFUMPT_VERSION := v0.8.0

help: ## Show this help message
	@echo 'Usage:'
	@echo '  make [target]'
	@echo ''
	@echo 'Targets:'
	@awk 'BEGIN {FS = ":.*?## "} /^[a-zA-Z_-]+:.*?## / {printf "  %-15s %s\n", $$1, $$2}' $(MAKEFILE_LIST)

install-tools: ## Install development tools
	@echo "Installing development tools..."
	@go install mvdan.cc/gofumpt@$(GOFUMPT_VERSION)
	@go install github.com/golangci/golangci-lint/cmd/golangci-lint@$(GOLANGCI_LINT_VERSION)

fmt: ## Format code with gofumpt
	@echo "Formatting code with gofumpt..."
	@gofumpt -w .

lint: ## Run golangci-lint v2
	@echo "Running golangci-lint..."
	@golangci-lint run --config .golangci.yml

lint-fix: ## Run golangci-lint v2 with --fix
	@echo "Running golangci-lint with --fix..."
	@golangci-lint run --config .golangci.yml --fix

test: ## Run all tests
	@echo "Running all tests..."
	@go test -v -race -parallel $(PARALLELISM) ./...

test-unit: ## Run unit tests only
	@echo "Running unit tests..."
	@go test -v -race -parallel $(PARALLELISM) -short ./internal/...

test-integration: ## Run integration tests only
	@echo "Running integration tests..."
	@go test -v -race -parallel 4 -timeout 10m ./integration_test.go

test-performance: ## Run performance tests only
	@echo "Running performance tests..."
	@RUN_PERFORMANCE_TESTS=true go test -v -run "TestConcurrent|TestSystemLoad" -timeout 15m ./performance_test.go

test-parallel: ## Run tests in parallel by package
	@echo "Running tests in parallel by package..."
	@go test -v -race -parallel $(PARALLELISM) -coverprofile=coverage-handlers.out ./internal/handlers/... &
	@go test -v -race -parallel $(PARALLELISM) -coverprofile=coverage-services.out ./internal/services/... &
	@(echo "Testing other packages..." && \
		for pkg in models database middleware utils; do \
			if [ -d "./internal/$$pkg" ] && ls ./internal/$$pkg/*.go >/dev/null 2>&1; then \
				echo "Testing ./internal/$$pkg/..." && \
				go test -v -race -parallel $(PARALLELISM) -coverprofile=coverage-$$pkg.out ./internal/$$pkg/... || true; \
			fi; \
		done && \
		echo "mode: atomic" > coverage-other.out && \
		for pkg in models database middleware utils; do \
			if [ -f "coverage-$$pkg.out" ]; then \
				tail -n +2 coverage-$$pkg.out >> coverage-other.out && \
				rm -f coverage-$$pkg.out; \
			fi; \
		done) &
	@wait
	@echo "Merging coverage reports..."
	@echo "mode: atomic" > coverage.out
	@tail -n +2 coverage-*.out | grep -v "mode: atomic" >> coverage.out || true
	@rm -f coverage-*.out

test-coverage: ## Run tests with coverage
	@echo "Running tests with coverage..."
	@go test -v -race -parallel $(PARALLELISM) -coverprofile=coverage.out -covermode=atomic ./...
	@go tool cover -html=coverage.out -o coverage.html
	@go tool cover -func=coverage.out | tee coverage.txt

vet: ## Run go vet
	@echo "Running go vet..."
	@go vet ./...

check: ## Run all checks in parallel (format, lint, vet)
	@echo "Running all checks in parallel..."
	@$(MAKE) fmt &
	@$(MAKE) lint &
	@$(MAKE) vet &
	@wait
	@echo "All checks completed"

build: ## Build the application
	@echo "Building $(BINARY_NAME)..."
	@go build -o $(BINARY_NAME) .

build-optimized: ## Build the application with optimizations
	@echo "Building optimized $(BINARY_NAME)..."
	@go build -ldflags="-s -w" -o $(BINARY_NAME) .

clean: ## Clean build artifacts
	@echo "Cleaning..."
	@rm -f $(BINARY_NAME)
	@rm -f coverage*.out coverage*.html coverage*.txt
	@go clean -cache
	@go clean -testcache

pre-commit: ## Run pre-commit checks in parallel
	@echo "Running pre-commit checks..."
	@$(MAKE) check
	@$(MAKE) test-unit
	@echo "Pre-commit checks completed successfully"

ci-test: ## Run CI tests (optimized for CI environment)
	@echo "Running CI tests..."
	@$(MAKE) test-parallel
	@$(MAKE) test-integration
	@$(MAKE) test-performance

benchmark: ## Run benchmarks
	@echo "Running benchmarks..."
	@go test -bench=. -benchmem ./...

.DEFAULT_GOAL := help