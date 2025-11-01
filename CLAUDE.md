# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Architecture Overview

This is a Go ecommerce API implementing **Hexagonal Architecture** (Ports and Adapters pattern) with three distinct layers:

- **Core Layer** (`internal/core/`): Business entities, models, and port interfaces (no external dependencies)
  - `ports/`: All port interfaces (repositories, services, use cases, technical utilities)
  - `models/`: Domain entities and value objects
- **Application Layer** (`internal/application/`): Services and use cases implementing business logic
- **Infrastructure Layer** (`internal/infraestructure/`): Adapters for external concerns (HTTP, database, JWT, logging)

### Key Architectural Patterns

- **Ports and Adapters**: Core defines ports (interfaces), infrastructure provides adapters (implementations)
- **Dependency Injection**: Uses Uber FX with `fx.Annotate` for interface binding in `main.go`
- **Repository Pattern**: Interfaces in `core/ports/`, PostgreSQL implementations in `infraestructure/adapters/repositories/`
- **Transaction Management**: Context-based with `TxContextKey` - repositories support both transactional and direct DB operations
- **Use Cases**: Specific business operations like `SignInUseCase`, `SignUpUseCase` coordinate multiple services

### Dependency Flow (CRITICAL)

**ALWAYS follow the Dependency Rule: dependencies point INWARD toward the core**

```
┌─────────────────────────────────────────┐
│   Infrastructure Layer (outermost)      │  ✅ CAN depend on: Application, Core
│   - HTTP handlers                       │  ❌ CANNOT depend on: Nothing (outermost)
│   - Database adapters                   │
│   - External services                   │
└──────────────┬──────────────────────────┘
               │ depends on ↓
┌──────────────▼──────────────────────────┐
│   Application Layer (middle)            │  ✅ CAN depend on: Core
│   - Services (business logic)           │  ❌ CANNOT depend on: Infrastructure
│   - Use Cases (coordinators)            │
└──────────────┬──────────────────────────┘
               │ depends on ↓
┌──────────────▼──────────────────────────┐
│   Core Layer (innermost)                │  ✅ CAN depend on: Nothing (pure domain)
│   - Models/Entities                     │  ❌ CANNOT depend on: Application, Infrastructure
│   - Port interfaces                     │
│   - Domain logic                        │
└─────────────────────────────────────────┘
```

**Rules:**
- ✅ **Infrastructure → Application → Core** (valid dependency direction)
- ❌ **Core → Application** (NEVER - core must be independent)
- ❌ **Core → Infrastructure** (NEVER - core doesn't know about outer layers)
- ❌ **Application → Infrastructure** (NEVER - application depends on ports/interfaces only)

**Example violations to AVOID:**
```go
// ❌ BAD: Core importing from Application
package ports
import "github.com/mlgaray/ecommerce_api/internal/application/services"

// ✅ GOOD: Application importing from Core
package services
import "github.com/mlgaray/ecommerce_api/internal/core/ports"
```

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

1. Define interface in `core/ports/<name>.go`
2. Use `package ports` in the file
3. Implement in application or infrastructure layer
4. Register binding in `main.go` using `fx.Annotate`
5. Generate mocks with `make generate-mocks`

**Import example**:
```go
import "github.com/mlgaray/ecommerce_api/internal/core/ports"

type MyService struct {
    productRepo   ports.ProductRepository
    pagination    ports.PaginationService[...]
}
```

**Note**: While you may see both domain interfaces (like `ProductRepository`) and technical utilities (like `PaginationService`) in the same `ports/` directory, this follows Go's convention of keeping packages flat and simple. Avoid creating subdirectories within ports unless absolutely necessary.

## Layer Responsibilities

### Use Cases (`internal/application/usecases/`)
**Role**: Orchestrators/Coordinators - **NO business logic**

Use cases are **thin pass-through layers** that:
- Coordinate calls to one or more services
- Should NOT contain business logic
- Act as entry points for specific business operations
- Are essentially "traffic directors"

**Good Example:**
```go
func (uc *GetAllByShopIDUseCase) Execute(ctx context.Context, shopID, limit, cursor int) (*ports.PaginatedProducts, error) {
    // Just coordinate - delegate to service
    return uc.productService.GetAllByShopID(ctx, shopID, limit, cursor)
}
```

**Bad Example (DON'T DO THIS):**
```go
func (uc *GetAllByShopIDUseCase) Execute(ctx context.Context, shopID int) (*ports.PaginatedProducts, error) {
    products, err := uc.repository.GetAll(ctx, shopID)
    // ❌ Business logic in use case (should be in service)
    var nextCursor int
    if len(products) == limit {
        nextCursor = products[len(products)-1].ID
    }
    return &ports.PaginatedProducts{...}, nil
}
```

### Services (`internal/application/services/`)
**Role**: Business Logic Layer

Services contain:
- **ALL business logic** (calculations, transformations, validations)
- Domain rules and constraints
- Data aggregation and processing
- Calls to repositories for data access
- **Delegates cross-cutting concerns** (like pagination) to specialized services

**Example with PaginationService:**
```go
func (s *ProductService) GetAllByShopID(ctx context.Context, shopID, limit, cursor int) (*ports.PaginatedProducts, error) {
    products, err := s.productRepository.GetAllByShopID(ctx, shopID, limit, cursor)
    if err != nil {
        return nil, err
    }

    // ✅ Delegate pagination logic to specialized service (reusable)
    return s.paginationService.BuildProductsPagination(products, limit), nil
}
```

### Shared Services Pattern

**PaginationService** - Generic service following hexagonal architecture:

**1. Port Interface (core/ports/pagination_service.go):**
```go
package ports

// Identifiable represents any entity that has an ID field
// Defined in core layer - no dependencies on outer layers
type Identifiable interface {
    GetID() int
}

// Generic interface for cursor-based pagination
type PaginationService[T Identifiable] interface {
    BuildCursorPagination(items []T, limit int) (nextCursor int, hasMore bool)
}
```

**2. Implementation (application/services/pagination_service.go):**
```go
package services

import "github.com/mlgaray/ecommerce_api/internal/core/ports"

// Generic service implementation
// ✅ Application layer CAN depend on core layer (inward dependency)
type PaginationService[T ports.Identifiable] struct{}

func NewPaginationService[T ports.Identifiable]() *PaginationService[T] {
    return &PaginationService[T]{}
}

func (p *PaginationService[T]) BuildCursorPagination(
    items []T,
    limit int,
) (nextCursor int, hasMore bool) {
    if len(items) > 0 && len(items) == limit {
        nextCursor = items[len(items)-1].GetID()
        hasMore = true
    }
    return nextCursor, hasMore
}
```

**3. Dependency Injection (main.go):**
```go
import (
    "github.com/mlgaray/ecommerce_api/internal/core/models"
    // ... other imports
)

// PAGINATION - Bind implementation to interface
fx.Annotate(
    services.NewPaginationService[*models.Product],
    fx.As(new(ports.PaginationService[*models.Product])),
),

// For categories (when needed):
// fx.Annotate(
//     services.NewPaginationService[*models.Category],
//     fx.As(new(ports.PaginationService[*models.Category])),
// ),
```

**4. Usage in Services:**
```go
type ProductService struct {
    productRepository ports.ProductRepository
    paginationService ports.PaginationService[*models.Product]  // ✅ Interface, not concrete type
}

func NewProductService(
    productRepository ports.ProductRepository,
    paginationService ports.PaginationService[*models.Product],  // ✅ Depend on abstraction
) *ProductService {
    return &ProductService{
        productRepository: productRepository,
        paginationService: paginationService,
    }
}

func (s *ProductService) GetAllByShopID(ctx context.Context, shopID, limit, cursor int) (*ports.PaginatedProducts, error) {
    products, err := s.productRepository.GetAllByShopID(ctx, shopID, limit, cursor)
    if err != nil {
        return nil, err
    }

    // ✅ Call through interface
    nextCursor, hasMore := s.paginationService.BuildCursorPagination(products, limit)

    return &ports.PaginatedProducts{
        Products:   products,
        NextCursor: nextCursor,
        HasMore:    hasMore,
    }, nil
}
```

**5. Model Implementation:**
```go
// models/product.go
func (p *Product) GetID() int { return p.ID }

// models/category.go (when needed)
func (c *Category) GetID() int { return c.ID }
```

**Benefits:**
- ✅ **Loose coupling** - Services depend on interfaces, not implementations
- ✅ **Type-safe** - Generics provide compile-time type checking
- ✅ **Testable** - Easy to mock the interface
- ✅ **Reusable** - Single implementation for all `Identifiable` types
- ✅ **Hexagonal Architecture** - Follows ports and adapters pattern
- ✅ **Consistent pagination** across all entities

**Pattern:** Generic interface in `core/ports/`, generic implementation in `application/services/`, bound via dependency injection.

### Repositories (`internal/infraestructure/adapters/repositories/`)
**Role**: Data Access Layer

Repositories:
- Execute database queries
- Handle transactions
- Map database results to domain models
- NO business logic (only data access logic)

### Handlers (`internal/infraestructure/adapters/http/`)
**Role**: HTTP Layer

Handlers:
- Parse HTTP requests
- Validate input
- Call use cases
- **Build HTTP responses from primitive data**
- NO business logic

**Important Pattern - Response Construction**:
Use cases and services return **primitive data** (arrays, ints, bools), NOT response DTOs. The handler is responsible for constructing the HTTP response:

```go
// ✅ GOOD: Use case returns primitives
func (uc *GetAllByShopIDUseCase) Execute(ctx context.Context, shopID, limit, cursor int) ([]*models.Product, int, bool, error) {
    return uc.productService.GetAllByShopID(ctx, shopID, limit, cursor)
}

// ✅ GOOD: Handler constructs response
func (h *ProductHandler) GetAllByShopID(w http.ResponseWriter, r *http.Request) {
    products, nextCursor, hasMore, err := h.getAllByShopID.Execute(ctx, shopID, limit, cursor)

    // Handler builds the response DTO
    response := contracts.PaginatedProductsResponse{
        Products:   products,
        NextCursor: nextCursor,
        HasMore:    hasMore,
    }

    json.NewEncoder(w).Encode(response)
}
```

**Why?**
- ✅ Use cases and services stay HTTP-agnostic
- ✅ Response DTOs (`contracts/`) belong in infrastructure layer
- ✅ Handler has full control over HTTP response format
- ✅ Same use case can be called from different protocols (HTTP, gRPC, CLI)

## Project Structure Notes

- `entities/` vs `models/`: Entities are pure domain objects, models include database-specific fields
- `contracts/`: HTTP request/response DTOs with validation methods
- `usecases/`: **Coordinators only** - orchestrate service calls, NO business logic
- `services/`: **Business logic layer** - all domain rules, calculations, transformations
- `handlers/`: Thin HTTP adapters, minimal business logic