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

migrate-force:
	@echo "Running migrations..."
	migrate -path database/migrations/seeds/ -database "postgresql://$(DB_USER):$(DB_PASSWORD)@$(DB_HOST):$(MIGRATE_DB_PORT)/$(DB_NAME)" force $(V)

.PHONY: migrate-force

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
	@echo "Running seeds migrations..."
	migrate -path database/migrations/seeds/ -database "postgresql://$(DB_USER):$(DB_PASSWORD)@$(DB_HOST):$(MIGRATE_DB_PORT)/$(DB_NAME)?x-migrations-table=schema_seeds" force 3

.PHONY: migrate-force-seeds

migrate-seeds-version:
	@echo "Running seeds migrations..."
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

.PHONY: fmt lint lint-fix code-quality