.PHONY: help build run test lint clean

# Variables
BINARY_NAME=polling-app
GO=go
GOFLAGS=-v
LDFLAGS=-ldflags "-s -w"

help: ## Display this help screen
	@grep -h -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-30s\033[0m %s\n", $$1, $$2}'

build: ## Build the application
	$(GO) build $(GOFLAGS) $(LDFLAGS) -o $(BINARY_NAME) ./cmd

run: build ## Build and run the application
	./$(BINARY_NAME)

test: ## Run tests
	$(GO) test $(GOFLAGS) -race -coverprofile=coverage.out ./...
	$(GO) tool cover -html=coverage.out -o coverage.html

lint: ## Run linters (golangci-lint)
	@which golangci-lint > /dev/null || (echo "Installing golangci-lint..." && go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest)
	golangci-lint run ./...

fmt: ## Format code
	$(GO) fmt ./...

clean: ## Remove build artifacts and temporary files
	$(GO) clean
	rm -f $(BINARY_NAME)
	rm -f coverage.out coverage.html

.DEFAULT_GOAL := help
