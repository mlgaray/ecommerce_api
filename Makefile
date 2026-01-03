# Define the path where you want to generate your mocks
SRC_DIR := .

# Install mockery if not installed
install-mockery:
	@echo "Installing mockery..."
	@go install github.com/vektra/mockery/v2/...@latest

# Clean old mocks
clean-mocks:
	@echo "Cleaning old mocks..."
	@rm -rf $(MOCK_DIR)

.PHONY: clean-mocks

# Generate mocks for all interfaces found drivers the project
generate-mocks: clean-mocks
	@echo "Generating mocks..."
	@find $(SRC_DIR) -name '*.go' | xargs grep -l 'type .* interface' | \
	while read -r file; do \
		mockery --dir "$$(dirname "$$file")" --all; \
	done

# Combined target to install mockery and generate mocks
setup-mocks: install-mockery generate-mocks

# Default target
all: setup-mocks

.PHONY: install-mockery clean-mocks generate-mocks setup-mocks all

# Define the name of the env file
ENV_FILE=.env.develop

# Define the contents of the env file
ENV_CONTENTS := ENVIRONMENT=develop\n\
DB_USER=\n\
DB_PASSWORD=\n\
DB_HOST=\n\
DB_PORT=\n\
DB_NAME=\n

# Rule to create the env file if it does not exist
create-env-file:
	@echo "Checking for $(ENV_FILE)..."
	@if [ ! -f $(ENV_FILE) ]; then \
		echo "Creating $(ENV_FILE)..."; \
		echo -e "$(ENV_CONTENTS)" > $(ENV_FILE); \
		echo "$(ENV_FILE) created successfully."; \
	else \
		echo "$(ENV_FILE) already exists. No changes made."; \
	fi

.PHONY: create-env-file



# Incluir el archivo .env.develop y definir el comando migrate
include $(ENV_FILE)
# Exportar variables específicas (compatible con Windows)
export DB_USER DB_PASSWORD DB_HOST DB_PORT DB_NAME MIGRATE_DB_PORT


migrate-create:
	migrate create -ext sql --dir database/migrations -seq $(MIGRATION_NAME)

.PHONY: migrate-create # example: make migrate-create MIGRATION_NAME="create_table_products"

migrate-create-seeds:
	@echo "Running seeds migrations..."
	migrate create -ext sql --dir database/migrations/seeds -seq $(MIGRATION_NAME)

.PHONY: migrate-create-seeds # example: make migrate-create MIGRATION_NAME="create_table_products"

migrate-up:
	@echo "Running migrations..."
	migrate -path database/migrations/ -database "postgresql://$(DB_USER):$(DB_PASSWORD)@$(DB_HOST):$(MIGRATE_DB_PORT)/$(DB_NAME)" -verbose up

.PHONY: migrate-up

migrate-down:
	@echo "Running migrations..."
	migrate -path database/migrations/ -database "postgresql://$(DB_USER):$(DB_PASSWORD)@$(DB_HOST):$(MIGRATE_DB_PORT)/$(DB_NAME)"  -verbose down

.PHONY: migrate-down

# Definir la meta para migrate
migrate-up-seeds:
	@echo "Running seeds migrations..."
	migrate -path database/migrations/seeds/ -database "postgresql://$(DB_USER):$(DB_PASSWORD)@$(DB_HOST):$(MIGRATE_DB_PORT)/$(DB_NAME)?x-migrations-table=schema_seeds" -verbose up

.PHONY: migrate-up-seeds

migrate-down-seeds:
	@echo "Running seeds migrations..."
	migrate -path database/migrations/seeds/ -database "postgresql://$(DB_USER):$(DB_PASSWORD)@$(DB_HOST):$(MIGRATE_DB_PORT)/$(DB_NAME)?x-migrations-table=schema_seeds" -verbose down

.PHONY: migrate-down-seeds

migrate-force-seeds:
	@echo "Forcing seeds migrations to version $(V)..."
	migrate -path database/migrations/seeds/ -database "postgresql://$(DB_USER):$(DB_PASSWORD)@$(DB_HOST):$(MIGRATE_DB_PORT)/$(DB_NAME)?x-migrations-table=schema_seeds" force 10

.PHONY: migrate-force-seeds



migrate-force:
	@echo "Forcing seeds migrations to version $(V)..."
	migrate -path database/migrations/ -database "postgresql://$(DB_USER):$(DB_PASSWORD)@$(DB_HOST):$(MIGRATE_DB_PORT)/$(DB_NAME)" force 7

.PHONY: migrate-force

migrate-seeds-version:
	@echo "Checking seeds migrations version..."
	migrate -path database/migrations/seeds/ -database "postgresql://$(DB_USER):$(DB_PASSWORD)@$(DB_HOST):$(MIGRATE_DB_PORT)/$(DB_NAME)?x-migrations-table=schema_seeds" version

.PHONY: migrate-seeds-version

# Functions migrations (stored procedures)
migrate-up-functions:
	@echo "Running functions migrations..."
	migrate -path database/migrations/functions/ -database "postgresql://$(DB_USER):$(DB_PASSWORD)@$(DB_HOST):$(MIGRATE_DB_PORT)/$(DB_NAME)?x-migrations-table=schema_functions" -verbose up

.PHONY: migrate-up-functions

migrate-down-functions:
	@echo "Running functions migrations..."
	migrate -path database/migrations/functions/ -database "postgresql://$(DB_USER):$(DB_PASSWORD)@$(DB_HOST):$(MIGRATE_DB_PORT)/$(DB_NAME)?x-migrations-table=schema_functions" -verbose down

.PHONY: migrate-down-functions

migrate-functions-version:
	@echo "Checking functions migrations version..."
	migrate -path database/migrations/functions/ -database "postgresql://$(DB_USER):$(DB_PASSWORD)@$(DB_HOST):$(MIGRATE_DB_PORT)/$(DB_NAME)?x-migrations-table=schema_functions" version

.PHONY: migrate-functions-version

# Code formatting and linting
fmt:
	@echo "formatting..."
	@gofmt -s -w .
	@echo "Running goimports..."
	@goimports -w .

lint:
	@echo "Running golangci-lint..."
	@golangci-lint run

lint-fix:
	@echo "Running golangci-lint with auto-fix..."
	@golangci-lint run --fix

# Combined code quality target
code-quality: fmt lint

.PHONY: lint-fix fmt lint code-quality

# Database drop: Drop everything (seeds, functions, tables)
# Order matters: seeds -> functions -> tables (respect dependencies)
drop-db:
	@echo "=========================================="
	@echo "DROPPING DATABASE (All Objects)"
	@echo "=========================================="
	@echo ""
	@echo "Step 1/3: Dropping seeds..."
	@migrate -path database/migrations/seeds/ -database "postgresql://$(DB_USER):$(DB_PASSWORD)@$(DB_HOST):$(MIGRATE_DB_PORT)/$(DB_NAME)?x-migrations-table=schema_seeds" down -all || echo "No seeds to drop or already dropped"
	@echo ""
	@echo "Step 2/3: Dropping functions (stored procedures)..."
	@migrate -path database/migrations/functions/ -database "postgresql://$(DB_USER):$(DB_PASSWORD)@$(DB_HOST):$(MIGRATE_DB_PORT)/$(DB_NAME)?x-migrations-table=schema_functions" down -all || echo "No functions to drop or already dropped"
	@echo ""
	@echo "Step 3/3: Dropping tables and indexes..."
	@migrate -path database/migrations/ -database "postgresql://$(DB_USER):$(DB_PASSWORD)@$(DB_HOST):$(MIGRATE_DB_PORT)/$(DB_NAME)" down -all || echo "No tables to drop or already dropped"
	@echo ""
	@echo "=========================================="
	@echo "DATABASE DROP COMPLETE"
	@echo "=========================================="

.PHONY: drop-db

# Database reset and up: Drop everything and recreate from scratch
# Order matters:
#   DOWN: seeds -> functions -> tables (respect dependencies)
#   UP: tables -> functions -> seeds (build dependencies)
reset-and-up-db:
	@echo "=========================================="
	@echo "RESETTING DATABASE (Full Reset)"
	@echo "=========================================="
	@echo ""
	@echo "Step 1/6: Dropping seeds..."
	@migrate -path database/migrations/seeds/ -database "postgresql://$(DB_USER):$(DB_PASSWORD)@$(DB_HOST):$(MIGRATE_DB_PORT)/$(DB_NAME)?x-migrations-table=schema_seeds" down -all || echo "⚠️  No seeds to drop or already dropped"
	@echo ""
	@echo "Step 2/6: Dropping functions (stored procedures)..."
	@migrate -path database/migrations/functions/ -database "postgresql://$(DB_USER):$(DB_PASSWORD)@$(DB_HOST):$(MIGRATE_DB_PORT)/$(DB_NAME)?x-migrations-table=schema_functions" down -all || echo "⚠️  No functions to drop or already dropped"
	@echo ""
	@echo "Step 3/6: Dropping tables and indexes..."
	@migrate -path database/migrations/ -database "postgresql://$(DB_USER):$(DB_PASSWORD)@$(DB_HOST):$(MIGRATE_DB_PORT)/$(DB_NAME)" down -all || echo "⚠️  No tables to drop or already dropped"
	@echo ""
	@echo "Step 4/6: Creating tables and indexes..."
	@migrate -path database/migrations/ -database "postgresql://$(DB_USER):$(DB_PASSWORD)@$(DB_HOST):$(MIGRATE_DB_PORT)/$(DB_NAME)" up
	@echo ""
	@echo "Step 5/6: Creating functions (stored procedures)..."
	@migrate -path database/migrations/functions/ -database "postgresql://$(DB_USER):$(DB_PASSWORD)@$(DB_HOST):$(MIGRATE_DB_PORT)/$(DB_NAME)?x-migrations-table=schema_functions" up
	@echo ""
	@echo "Step 6/6: Loading seeds..."
	@migrate -path database/migrations/seeds/ -database "postgresql://$(DB_USER):$(DB_PASSWORD)@$(DB_HOST):$(MIGRATE_DB_PORT)/$(DB_NAME)?x-migrations-table=schema_seeds" up
	@echo ""
	@echo "=========================================="
	@echo "DATABASE RESET COMPLETE"
	@echo "=========================================="

.PHONY: reset-and-up-db

# Migrate up all: tables -> functions -> seeds (in correct order)
migrate-up-all:
	@echo "=========================================="
	@echo "RUNNING ALL MIGRATIONS UP"
	@echo "=========================================="
	@echo ""
	@echo "Step 1/3: Creating tables and indexes..."
	@migrate -path database/migrations/ -database "postgresql://$(DB_USER):$(DB_PASSWORD)@$(DB_HOST):$(MIGRATE_DB_PORT)/$(DB_NAME)" -verbose up
	@echo ""
	@echo "Step 2/3: Creating functions (stored procedures)..."
	@migrate -path database/migrations/functions/ -database "postgresql://$(DB_USER):$(DB_PASSWORD)@$(DB_HOST):$(MIGRATE_DB_PORT)/$(DB_NAME)?x-migrations-table=schema_functions" -verbose up
	@echo ""
	@echo "Step 3/3: Loading seeds..."
	@migrate -path database/migrations/seeds/ -database "postgresql://$(DB_USER):$(DB_PASSWORD)@$(DB_HOST):$(MIGRATE_DB_PORT)/$(DB_NAME)?x-migrations-table=schema_seeds" -verbose up
	@echo ""
	@echo "=========================================="
	@echo "ALL MIGRATIONS COMPLETE"
	@echo "=========================================="

.PHONY: migrate-up-all

# Quick reset: Reset only seeds (keeps tables and structure)
reset-seeds:
	@echo "🔄 Resetting seeds only..."
	@migrate -path database/migrations/seeds/ -database "postgresql://$(DB_USER):$(DB_PASSWORD)@$(DB_HOST):$(MIGRATE_DB_PORT)/$(DB_NAME)?x-migrations-table=schema_seeds" down -all || echo "⚠️  No seeds to drop"
	@migrate -path database/migrations/seeds/ -database "postgresql://$(DB_USER):$(DB_PASSWORD)@$(DB_HOST):$(MIGRATE_DB_PORT)/$(DB_NAME)?x-migrations-table=schema_seeds" up
	@echo "✅ Seeds reset complete"

.PHONY: reset-seeds