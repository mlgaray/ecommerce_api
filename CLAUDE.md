# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## E-commerce API

Go-based e-commerce API implementing **Hexagonal Architecture** (Ports and Adapters pattern) with PostgreSQL, JWT authentication, and comprehensive testing.

## 📚 Documentation

The project documentation is organized into specialized guides:

### Core Concepts

- **[Architecture](./.claude/ARCHITECTURE.md)** - Hexagonal Architecture, Dependency Flow, Ports and Adapters
- **[Layer Responsibilities](./.claude/LAYERS.md)** - Use Cases, Services, Repositories, Handlers, Rich Models
- **[Error Handling](./.claude/ERRORS.md)** - Domain vs HTTP errors, Validation layers, Error flow

### Development Practices

- **[Validations](./.claude/VALIDATIONS.md)** - HTTP vs Business validations, Separation of concerns, Validation flow
- **[Logging](./.claude/LOGGING.md)** - Structured logging patterns, Field naming, Log levels
- **[Testing](./.claude/TESTING.md)** - Unit tests, Integration tests (BDD), Mock generation
- **[Database](./.claude/DATABASE.md)** - Migrations, Transactions, Repository patterns, Query patterns

### Setup & Commands

- **[Development Guide](./.claude/DEVELOPMENT.md)** - Setup, Commands, Workflows, Troubleshooting

## Quick Start

```bash
# Install dependencies
go mod download

# Setup environment
make create-env-file

# Run migrations
make migrate-up

# Run application
go run main.go
```

## Key Architectural Principles

### Dependency Rule

**Dependencies point INWARD toward the core**

```
Infrastructure (HTTP, DB) → Application (Services) → Core (Domain)
```

- ✅ Infrastructure CAN depend on Application and Core
- ✅ Application CAN depend on Core
- ❌ Core CANNOT depend on Application or Infrastructure
- ❌ Application CANNOT depend on Infrastructure

### Layer Separation

- **Use Cases**: Coordinators only, NO business logic
- **Services**: ALL business logic, domain rules, validations
- **Repositories**: Data access only, NO business logic
- **Handlers**: HTTP I/O only, NO business logic
- **Models**: Rich domain models with behavior

### Validation Strategy

- **HTTP Validations** (`contracts/`): Required fields, format, file uploads
- **Business Validations** (`models/`): Domain rules, invariants (via `Model.Validate()`)
- Services call `Model.Validate()` before persisting data

See [Validations](./.claude/VALIDATIONS.md) for complete guide.

### Error Handling

- **Domain Errors** (`core/errors/`): ValidationError, AuthenticationError, RecordNotFoundError
- **HTTP Errors** (`infraestructure/adapters/http/errors/`): BadRequestError
- **Technical Errors**: `fmt.Errorf()` for unexpected errors

See [Error Handling](./.claude/ERRORS.md) for complete guide.

## Tech Stack

- **Language**: Go 1.21+
- **Web**: Gorilla Mux
- **Database**: PostgreSQL with lib/pq
- **Migrations**: golang-migrate
- **Auth**: JWT (golang-jwt/jwt/v5)
- **Logging**: Logrus with Lumberjack
- **DI**: Uber FX
- **Testing**: Testify, Mockery, Cucumber/Godog

## Common Commands

```bash
# Code quality
make code-quality        # Format + lint

# Testing
go test ./...           # Unit tests
cd tests/integration && go test  # Integration tests
make generate-mocks     # Generate mocks

# Database
make migrate-up         # Run migrations
make migrate-down       # Rollback migrations
make migrate-create MIGRATION_NAME="name"  # Create migration

# Development
go run main.go          # Run server
go build -o bin/app     # Build binary
```

See [Development Guide](./.claude/DEVELOPMENT.md) for complete command reference.

## Project Structure

```
.
├── .claude/                    # Documentation
├── internal/
│   ├── core/                   # Domain layer (no dependencies)
│   │   ├── errors/             # Domain errors
│   │   ├── models/             # Rich domain models
│   │   └── ports/              # Interface definitions
│   ├── application/            # Business logic
│   │   ├── services/           # Business logic implementation
│   │   └── usecases/           # Orchestration layer
│   └── infraestructure/        # External adapters
│       └── adapters/
│           ├── http/           # HTTP handlers & contracts
│           ├── repositories/   # Database implementations
│           └── auth/           # JWT service
├── tests/integration/          # BDD integration tests
├── database/migrations/        # SQL migrations
└── main.go                     # DI container & entry point
```

## Contributing

1. Read the [Architecture](./.claude/ARCHITECTURE.md) guide
2. Follow [Layer Responsibilities](./.claude/LAYERS.md)
3. Implement proper [Validations](./.claude/VALIDATIONS.md) (HTTP vs Business)
4. Implement proper [Error Handling](./.claude/ERRORS.md)
5. Use structured [Logging](./.claude/LOGGING.md)
6. Write [Tests](./.claude/TESTING.md)
7. Follow [Database](./.claude/DATABASE.md) patterns
8. Use `make code-quality` before committing

## Need Help?

- Check the specific guide for your topic (see [Documentation](#-documentation) above)
- Review existing code for patterns
- Consult [Development Guide](./.claude/DEVELOPMENT.md) for troubleshooting

---

**Note**: This is a learning project demonstrating Clean Architecture, Hexagonal Architecture, and Domain-Driven Design principles in Go.
