# Layer Responsibilities

Clear separation of concerns across architectural layers.

## Use Cases (`internal/application/usecases/`)

**Role**: Orchestrators/Coordinators - **NO business logic**

Use cases are **thin pass-through layers** that:
- Coordinate calls to one or more services
- Should NOT contain business logic
- Act as entry points for specific business operations
- Are essentially "traffic directors"

### Good Example

```go
func (uc *GetAllByShopIDUseCase) Execute(ctx context.Context, shopID, limit, cursor int) ([]*models.Product, int, bool, error) {
    // Just coordinate - delegate to service
    return uc.productService.GetAllByShopID(ctx, shopID, limit, cursor)
}
```

### Bad Example (DON'T DO THIS)

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

## Services (`internal/application/services/`)

**Role**: Business Logic Layer

Services contain:
- **ALL business logic** (calculations, transformations, validations)
- Domain rules and constraints
- Data aggregation and processing
- Calls to repositories for data access
- **Delegates cross-cutting concerns** (like pagination) to specialized services

### Example

```go
func (s *ProductService) GetAllByShopID(ctx context.Context, shopID, limit, cursor int) ([]*models.Product, int, bool, error) {
    // Get products from repository
    products, err := s.productRepository.GetAllByShopID(ctx, shopID, limit, cursor)
    if err != nil {
        return nil, 0, false, err
    }

    // ✅ Delegate pagination logic to specialized service (reusable)
    nextCursor, hasMore := s.paginationService.BuildCursorPagination(products, limit)

    return products, nextCursor, hasMore, nil
}
```

## Shared Services Pattern

**PaginationService** - Generic service following hexagonal architecture

### 1. Port Interface (core/ports/pagination_service.go)

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

### 2. Implementation (application/services/pagination_service.go)

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

### 3. Dependency Injection (main.go)

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

### 4. Usage in Services

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

func (s *ProductService) GetAllByShopID(ctx context.Context, shopID, limit, cursor int) ([]*models.Product, int, bool, error) {
    products, err := s.productRepository.GetAllByShopID(ctx, shopID, limit, cursor)
    if err != nil {
        return nil, 0, false, err
    }

    // ✅ Call through interface
    nextCursor, hasMore := s.paginationService.BuildCursorPagination(products, limit)

    return products, nextCursor, hasMore, nil
}
```

### 5. Model Implementation

```go
// models/product.go
func (p *Product) GetID() int { return p.ID }

// models/category.go (when needed)
func (c *Category) GetID() int { return c.ID }
```

### Benefits

- ✅ **Loose coupling** - Services depend on interfaces, not implementations
- ✅ **Type-safe** - Generics provide compile-time type checking
- ✅ **Testable** - Easy to mock the interface
- ✅ **Reusable** - Single implementation for all `Identifiable` types
- ✅ **Hexagonal Architecture** - Follows ports and adapters pattern
- ✅ **Consistent pagination** across all entities

**Pattern:** Generic interface in `core/ports/`, generic implementation in `application/services/`, bound via dependency injection.

## Repositories (`internal/infraestructure/adapters/repositories/`)

**Role**: Data Access Layer

Repositories:
- Execute database queries
- Handle transactions
- Map database results to domain models
- NO business logic (only data access logic)
- Return domain errors (RecordNotFoundError, DuplicateRecordError)

### Example

```go
func (r *ProductRepository) GetByID(ctx context.Context, productID int) (*models.Product, error) {
    // Detect transaction from context
    if tx, ok := ctx.Value(TxContextKey).(*sql.Tx); ok {
        return r.getByIDWithTx(ctx, tx, productID)
    }
    return r.getByIDWithDB(ctx, productID)
}

func (r *ProductRepository) getByIDWithDB(ctx context.Context, productID int) (*models.Product, error) {
    var product models.Product
    err := r.db.QueryRowContext(ctx, query, productID).Scan(&product.ID, &product.Name)

    if err == sql.ErrNoRows {
        return nil, &errors.RecordNotFoundError{Message: errors.ProductNotFound}
    }

    if err != nil {
        // Technical error
        return nil, fmt.Errorf("database operation failed: %w", err)
    }

    return &product, nil
}
```

## Handlers (`internal/infraestructure/adapters/http/`)

**Role**: HTTP Layer

Handlers:
- Parse HTTP requests
- Validate HTTP input (format, required fields)
- Call use cases
- **Build HTTP responses from primitive data**
- NO business logic

### Important Pattern - Response Construction

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

### Why?

- ✅ Use cases and services stay HTTP-agnostic
- ✅ Response DTOs (`contracts/`) belong in infrastructure layer
- ✅ Handler has full control over HTTP response format
- ✅ Same use case can be called from different protocols (HTTP, gRPC, CLI)

## Rich Domain Models (`internal/core/models/`)

**Role**: Domain Entities with Behavior

Models should be **Rich**, not Anemic:
- Contain domain logic
- Validate business rules
- Encapsulate behavior
- Protect invariants

### Example

```go
package models

import "github.com/mlgaray/ecommerce_api/internal/core/errors"

type Product struct {
    ID              int
    Name            string
    Price           float64
    Stock           int
    MinimumStock    int
    IsPromotional   bool
    PromotionalPrice float64
    IsActive        bool
}

// Validate validates business rules for the Product domain model
func (p *Product) Validate() error {
    if p.Price <= 0 {
        return &errors.ValidationError{Message: errors.ProductPriceMustBePositive}
    }

    if p.Stock < 0 {
        return &errors.ValidationError{Message: errors.ProductStockCannotBeNegative}
    }

    if p.IsPromotional && p.PromotionalPrice <= 0 {
        return &errors.ValidationError{Message: errors.PromotionalProductRequiresPromotionalPrice}
    }

    if p.IsPromotional && p.PromotionalPrice >= p.Price {
        return &errors.ValidationError{Message: errors.PromotionalPriceMustBeLowerThanRegularPrice}
    }

    return nil
}

// Business logic methods
func (p *Product) CanBeSold() bool {
    return p.IsActive && p.Stock > 0
}

func (p *Product) IsLowStock() bool {
    return p.Stock <= p.MinimumStock
}

func (p *Product) GetEffectivePrice() float64 {
    if p.IsPromotional {
        return p.PromotionalPrice
    }
    return p.Price
}

func (p *Product) DecrementStock(quantity int) error {
    if quantity <= 0 {
        return &errors.ValidationError{Message: errors.QuantityMustBePositive}
    }

    if p.Stock < quantity {
        return &errors.BusinessRuleError{Message: errors.InsufficientStock}
    }

    p.Stock -= quantity
    return nil
}

func (p *Product) IncrementStock(quantity int) error {
    if quantity <= 0 {
        return &errors.ValidationError{Message: errors.QuantityMustBePositive}
    }

    p.Stock += quantity
    return nil
}

// GetID implements Identifiable interface for pagination
func (p *Product) GetID() int { return p.ID }
```

## Summary Table

| Layer | Location | Responsibilities | Can Depend On | Cannot Depend On |
|-------|----------|------------------|---------------|------------------|
| **Handlers** | `infraestructure/adapters/http/` | HTTP I/O, validation, response building | Use Cases, Domain Models | Nothing (outermost) |
| **Use Cases** | `application/usecases/` | Orchestration ONLY | Services, Domain Models | Infrastructure |
| **Services** | `application/services/` | Business logic, domain rules | Repositories, Domain Models | Infrastructure |
| **Repositories** | `infraestructure/adapters/repositories/` | Data access, queries | Domain Models | Nothing (outermost) |
| **Models** | `core/models/` | Domain entities, business behavior | Domain Errors | Application, Infrastructure |
| **Ports** | `core/ports/` | Interface definitions | Domain Models | Application, Infrastructure |

## Data Flow

```
HTTP Request
    ↓
Handler (validates HTTP, parses request)
    ↓
Use Case (coordinates)
    ↓
Service (business logic)
    ↓
Repository (data access)
    ↓
Database
```

## Return Flow

```
Database
    ↓
Repository (domain models + domain errors)
    ↓
Service (domain models + domain errors)
    ↓
Use Case (domain models + domain errors)
    ↓
Handler (maps to HTTP response + HTTP status)
    ↓
HTTP Response
```
