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

test-integration-fast: ## Run integration tests with optimized settings
	@echo "Running fast integration tests..."
	@go test -v -race -parallel 8 -timeout 5m ./integration_test.go

test-integration-parallel: ## Run integration tests by type in parallel
	@echo "Running parallel integration tests..."
	@INTEGRATION_TEST_TYPE=flow-tests go test -v -race -parallel 4 -run "TestIntegrationQuizFlow|TestIntegrationParticipantFlow" -timeout 3m ./integration_test.go &
	@INTEGRATION_TEST_TYPE=session-tests go test -v -race -parallel 4 -run "TestIntegrationSessionManagement" -timeout 2m ./integration_test.go &
	@INTEGRATION_TEST_TYPE=concurrent-tests go test -v -race -parallel 2 -run "TestIntegrationConcurrentAnswers" -timeout 3m ./integration_test.go &
	@wait

test-performance: ## Run performance tests only
	@echo "Running performance tests..."
	@RUN_PERFORMANCE_TESTS=true go test -v -run "TestConcurrent|TestSystemLoad" -timeout 15m ./performance_test.go

test-parallel: ## Run tests in parallel by package
	@echo "Running tests in parallel by package..."
	@echo "Running handlers tests..."
	@go test -v -race -parallel $(PARALLELISM) -coverprofile=coverage-handlers.out ./internal/handlers/...
	@echo "Running services tests..."
	@go test -v -race -parallel $(PARALLELISM) -coverprofile=coverage-services.out ./internal/services/...
	@echo "Merging coverage reports..."
	@echo "mode: atomic" > coverage.out
	@if [ -f coverage-handlers.out ] && [ -s coverage-handlers.out ]; then \
		echo "Merging handlers coverage..." && \
		tail -n +2 coverage-handlers.out >> coverage.out; \
	fi
	@if [ -f coverage-services.out ] && [ -s coverage-services.out ]; then \
		echo "Merging services coverage..." && \
		tail -n +2 coverage-services.out >> coverage.out; \
	fi
	@echo "Coverage merge complete."
	@rm -f coverage-handlers.out coverage-services.out

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