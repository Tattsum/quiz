.PHONY: help fmt lint test build clean check install-tools

BINARY_NAME=quiz
GO_VERSION=$(shell go version | awk '{print $$3}')

# Tool versions (managed by Renovate)
GOLANGCI_LINT_VERSION := v1.64.8
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

test: ## Run tests
	@echo "Running tests..."
	@go test ./...

vet: ## Run go vet
	@echo "Running go vet..."
	@go vet ./...

check: fmt lint vet test ## Run all checks (format, lint, vet, test)

build: ## Build the application
	@echo "Building $(BINARY_NAME)..."
	@go build -o $(BINARY_NAME) .

clean: ## Clean build artifacts
	@echo "Cleaning..."
	@rm -f $(BINARY_NAME)
	@go clean

pre-commit: check ## Run pre-commit checks
	@echo "Pre-commit checks completed successfully"

.DEFAULT_GOAL := help