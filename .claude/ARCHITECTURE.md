# Architecture Overview

This is a Go ecommerce API implementing **Hexagonal Architecture** (Ports and Adapters pattern) with three distinct layers:

- **Core Layer** (`internal/core/`): Business entities, models, and port interfaces (no external dependencies)
  - `ports/`: All port interfaces (repositories, services, use cases, technical utilities)
  - `models/`: Domain entities and value objects
  - `errors/`: Domain error types
- **Application Layer** (`internal/application/`): Services and use cases implementing business logic
- **Infrastructure Layer** (`internal/infraestructure/`): Adapters for external concerns (HTTP, database, JWT, logging)

## Key Architectural Patterns

- **Ports and Adapters**: Core defines ports (interfaces), infrastructure provides adapters (implementations)
- **Dependency Injection**: Uses Uber FX with `fx.Annotate` for interface binding in `main.go`
- **Repository Pattern**: Interfaces in `core/ports/`, PostgreSQL implementations in `infraestructure/adapters/repositories/`
- **Transaction Management**: Context-based with `TxContextKey` - repositories support both transactional and direct DB operations
- **Use Cases**: Specific business operations like `SignInUseCase`, `SignUpUseCase` coordinate multiple services

## Dependency Flow (CRITICAL)

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

## Request Flow

```
HTTP Request → Middleware → Handler → UseCase → Service → Repository → Database
```

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

## Project Structure Notes

- `entities/` vs `models/`: Entities are pure domain objects, models include database-specific fields
- `contracts/`: HTTP request/response DTOs with validation methods
- `usecases/`: **Coordinators only** - orchestrate service calls, NO business logic
- `services/`: **Business logic layer** - all domain rules, calculations, transformations
- `handlers/`: Thin HTTP adapters, minimal business logic
