# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Architecture Overview

This is a Go ecommerce API implementing **Hexagonal Architecture** (Ports and Adapters pattern) with three distinct layers:

- **Core Layer** (`internal/core/`): Business entities, models, and port interfaces (no external dependencies)
- **Application Layer** (`internal/application/`): Services and use cases implementing business logic  
- **Infrastructure Layer** (`internal/infraestructure/`): Adapters for external concerns (HTTP, database, JWT, logging)

### Key Architectural Patterns

- **Ports and Adapters**: Core defines ports (interfaces), infrastructure provides adapters (implementations)
- **Dependency Injection**: Uses Uber FX with `fx.Annotate` for interface binding in `main.go`
- **Repository Pattern**: Interfaces in `core/ports/`, PostgreSQL implementations in `infraestructure/adapters/repositories/`
- **Transaction Management**: Context-based with `TxContextKey` - repositories support both transactional and direct DB operations
- **Use Cases**: Specific business operations like `SignInUseCase`, `SignUpUseCase` coordinate multiple services

### Request Flow
```
HTTP Request → Middleware → Handler → UseCase → Service → Repository → Database
```

## Development Commands

### Database Operations
```bash
# Create and run migrations
make migrate-create MIGRATION_NAME="create_table_products"
make migrate-up
make migrate-down

# Seeds (test data)
make migrate-create-seeds MIGRATION_NAME="seed_users"
make migrate-up-seeds
```

### Code Quality
```bash
# Format code and run linter (uses golangci-lint with --out-format=tab for clickable errors)
make code-quality

# Individual commands
make fmt        # gofmt + goimports
make lint       # golangci-lint run
make lint-fix   # golangci-lint run --fix
```

### Testing
```bash
# Unit tests
go test ./...

# Integration tests (uses Cucumber/Godog with BDD approach)
cd tests/integration && go test

# Generate mocks (uses mockery with .mockery.yml config)
make generate-mocks
```

### Environment Setup
```bash
# Create .env.develop file with database configuration
make create-env-file
```

## Testing Strategy

### Unit Tests
- Uses `testify/assert` and generated mocks
- Tests located alongside source files (`*_test.go`)
- Mock generation automated with Mockery
- Pattern: Arrange-Act-Assert with descriptive test names

### Integration Tests
- BDD approach using Cucumber/Godog in `tests/integration/`
- Feature files define scenarios, step definitions in `steps/`
- Uses sqlmock for database mocking
- Test context shared between steps for scenario state

## Key Technologies

- **Web**: Gorilla Mux routing, rs/cors middleware
- **Database**: PostgreSQL with lib/pq driver, golang-migrate
- **Auth**: JWT tokens with golang-jwt/jwt/v5
- **Logging**: Logrus with Lumberjack file rotation
- **DI**: Uber FX framework
- **Testing**: Testify, Mockery, Cucumber/Godog

## Logging Patterns

### Repository Layer (Technical Only)
```go
logs.WithFields(map[string]interface{}{
    "operation": "get_user_by_email_tx",
    "error": err.Error(),
}).Error("Database query failed")
```

### Handler/Service Layer (Business Context)
```go
logs.FromContext(ctx).WithFields(map[string]interface{}{
    "email": request.Email,
    "endpoint": "/auth/signin",
    "request_id": requestID,
}).Error("Sign in attempt failed")
```

## Database Patterns

### Repository Methods
Most repositories support both direct DB and transaction contexts:

```go
// Detects transaction from context automatically
if tx, ok := ctx.Value(TxContextKey).(*sql.Tx); ok {
    return s.methodWithTx(ctx, tx, params)
}
return s.methodWithDB(ctx, params)
```

### Error Handling
- Custom error types in `core/errors/`
- PostgreSQL-specific error handling for constraints
- HTTP status mapping in error handlers

## Configuration

### Environment Variables (.env.develop)
- `DB_USER`, `DB_PASSWORD`, `DB_HOST`, `DB_PORT`, `DB_NAME`
- `ENVIRONMENT` (develop/production)

### Linting Configuration (.golangci.yml)
- Comprehensive linter set enabled
- Email regex validation for contracts
- Excludes `w.Write` error checking for HTTP responses

## Dependencies Management

All interface definitions are in `core/ports/`. When adding new dependencies:

1. Define interface in appropriate `*_repository.go` or `*_service.go` port file
2. Implement in infrastructure layer
3. Register binding in `main.go` using `fx.Annotate`
4. Generate mocks with `make generate-mocks`

## Project Structure Notes

- `entities/` vs `models/`: Entities are pure domain objects, models include database-specific fields
- `contracts/`: HTTP request/response DTOs with validation methods
- `usecases/`: High-level business operations that coordinate services
- `services/`: Lower-level business logic operations
- `handlers/`: Thin HTTP adapters, minimal business logic