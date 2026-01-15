.PHONY: help build run test test-integration lint clean ent ent-gen install-atlas migrate-new migrate-apply migrate-status migrate-validate migrate-rollback migrate-reset migrate-ci migrate-hash db-up db-down db-shell seed install-hooks

# Variables
BINARY_NAME=polling-app
GO=go
GOFLAGS=-v
LDFLAGS=-ldflags "-s -w"
ATLAS_VERSION=latest
ENT_DIR=internal/ent/schema
PATH:=$(HOME)/bin:$(PATH)
export PATH

# Database configuration (override in CI/CD)
DB_HOST?=localhost
DB_PORT?=5432
DB_USER?=polling
DB_PASSWORD?=polling
DB_NAME?=polling_app
DB_URL=postgres://$(DB_USER):$(DB_PASSWORD)@$(DB_HOST):$(DB_PORT)/$(DB_NAME)?sslmode=disable

# Atlas dev database configuration
# For migrate-new: Atlas needs a clean dev database to compute diffs
# Option 1 (default): Use docker - atlas manages its own container
# Option 2: Use a separate postgres database - set ATLAS_DEV_URL env var
ATLAS_DEV_URL?=docker://postgres/16/test?search_path=public

help: ## Display this help screen
	@grep -h -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-30s\033[0m %s\n", $$1, $$2}'

build: ## Build the application
	$(GO) build $(GOFLAGS) $(LDFLAGS) -o $(BINARY_NAME) ./cmd/server

run: ## Run the application with go run
	$(GO) run ./cmd/server

test: ## Run unit tests (no Docker required)
	$(GO) test $(GOFLAGS) -race -coverprofile=coverage.out ./...
	$(GO) tool cover -html=coverage.out -o coverage.html

test-integration: ## Run integration tests (requires Docker)
	$(GO) test $(GOFLAGS) -race -tags=integration -count=1 -coverprofile=coverage-integration.out ./...
	$(GO) tool cover -html=coverage-integration.out -o coverage-integration.html

lint: ## Run linters (golangci-lint)
	@which golangci-lint > /dev/null || (echo "Installing golangci-lint..." && go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest)
	golangci-lint run ./...

fmt: ## Format code
	$(GO) fmt ./...

clean: ## Remove build artifacts and temporary files
	$(GO) clean
	rm -f $(BINARY_NAME)
	rm -f coverage.out coverage.html coverage-integration.out coverage-integration.html

ent-gen: ## Generate Ent code from schema
	$(GO) run -mod=mod entgo.io/ent/cmd/ent generate ./internal/ent/schema

install-atlas: ## Install Atlas CLI to ~/bin
	@echo "Installing Atlas CLI to ~/bin..."
	@mkdir -p ~/bin
	@curl -sSfL https://atlasbinaries.com/atlas/atlas-linux-amd64-latest -o ~/bin/atlas
	@chmod +x ~/bin/atlas
	@echo "Atlas installed successfully! Version:"
	@~/bin/atlas version

# Internal target to ensure atlas is installed (auto-installs if missing)
.ensure-atlas:
	@if ! which atlas > /dev/null 2>&1; then \
		echo "Atlas not found. Installing automatically..."; \
		mkdir -p ~/bin; \
		curl -sSfL https://atlasbinaries.com/atlas/atlas-linux-amd64-latest -o ~/bin/atlas; \
		chmod +x ~/bin/atlas; \
		echo "Atlas installed successfully!"; \
	fi

migrate-new: .ensure-atlas ## Create a new Atlas migration (usage: make migrate-new name=migration_name)
	@if [ -z "$(name)" ]; then \
		echo "Error: name parameter required. Usage: make migrate-new name=migration_name"; \
		exit 1; \
	fi
	@echo "Creating new migration: $(name)"
	@echo "Using Ent schema as source..."
	@echo "Dev database: $(ATLAS_DEV_URL)"
	atlas migrate diff --dir "file://internal/migrate/migrations" "$(name)" --to "ent://internal/ent/schema" --dev-url "$(ATLAS_DEV_URL)"

migrate-apply: .ensure-atlas ## Apply pending migrations
	atlas migrate apply --dir "file://internal/migrate/migrations" --url "$(DB_URL)"

migrate-status: .ensure-atlas ## Show migration status
	atlas migrate status --dir "file://internal/migrate/migrations" --url "$(DB_URL)"

migrate-validate: .ensure-atlas ## Validate migrations
	atlas migrate validate --dir "file://internal/migrate/migrations"

migrate-hash: .ensure-atlas ## Recalculate migration directory hash (after manually removing migrations)
	@echo "Recalculating migration directory hash..."
	atlas migrate hash --dir "file://internal/migrate/migrations"
	@echo "✓ Hash updated in atlas.sum"

migrate-rollback: .ensure-atlas ## Rollback to a specific version (usage: make migrate-rollback version=VERSION or use migrate-reset for full reset)
	@if [ -z "$(version)" ]; then \
		echo "Error: version parameter required. Usage: make migrate-rollback version=VERSION"; \
		echo "Use 'make migrate-status' to see available versions"; \
		echo "Or use 'make migrate-reset' to reset the entire database"; \
		exit 1; \
	fi
	@echo "Rolling back to version $(version)..."
	atlas schema apply --url "$(DB_URL)" --to "file://internal/migrate/migrations?version=$(version)" --auto-approve

migrate-reset: .ensure-atlas ## Reset database - drop all tables and reapply migrations (WARNING: deletes all data)
	@echo "WARNING: This will delete all data in the database!"
	@read -p "Type 'yes' to confirm: " confirm; \
	if [ "$$confirm" = "yes" ]; then \
		atlas schema clean --url "$(DB_URL)" --auto-approve; \
		PGPASSWORD=$(DB_PASSWORD) psql -h $(DB_HOST) -p $(DB_PORT) -U $(DB_USER) -d $(DB_NAME) -c "CREATE SCHEMA IF NOT EXISTS public;" 2>/dev/null || true; \
		atlas migrate apply --dir "file://internal/migrate/migrations" --url "$(DB_URL)"; \
		echo "Database reset and migrations applied"; \
	else \
		echo "Cancelled"; \
	fi

db-up: ## Start PostgreSQL database
	docker-compose up -d postgres
	@echo "Database started. Waiting for it to be ready..."
	@sleep 2

db-down: ## Stop PostgreSQL database
	docker-compose down

db-shell: ## Connect to PostgreSQL database shell
	docker-compose exec postgres psql -U polling -d polling_app

seed: ## Seed database with demo data
	$(GO) run ./cmd/seed

install-hooks: ## Install git pre-commit hooks
	@cp scripts/pre-commit .git/hooks/pre-commit
	@chmod +x .git/hooks/pre-commit
	@echo "Pre-commit hook installed!"

.DEFAULT_GOAL := help
