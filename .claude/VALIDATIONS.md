# Validation Strategy

This document defines the validation strategy for the e-commerce API, ensuring proper separation between HTTP validations and business validations according to Hexagonal Architecture principles.

## Validation Layers

### 1. HTTP Validations (Infrastructure Layer)

**Location**: `internal/infraestructure/adapters/http/contracts/`

**Responsibility**: Validate that the HTTP request is well-formed and meets HTTP-specific requirements.

**What to validate**:
- ✅ Required HTTP fields (name, description, category_id, shop_id, etc.)
- ✅ JSON format and parsing errors
- ✅ HTTP parameter formats (IDs must be integers, emails must be valid format, etc.)
- ✅ File uploads (size limits, MIME types, file quantity)
- ✅ Request structure (multipart form data, JSON body)

**What NOT to validate**:
- ❌ Business rules (price > 0, stock validations, promotional price rules)
- ❌ Domain constraints (minimum stock logic, promotional price comparisons)
- ❌ Complex business logic

**Error Types**: Use `httpErrors.BadRequestError` from `internal/infraestructure/adapters/http/errors/`

### 2. Business Validations (Domain Layer)

**Location**: `internal/core/models/` (in domain model methods like `Validate()`)

**Responsibility**: Validate that domain objects comply with business rules and domain invariants.

**What to validate**:
- ✅ Price must be positive
- ✅ Stock cannot be negative
- ✅ Minimum stock business rules
- ✅ Promotional price rules
- ✅ Any business constraint or invariant

**What NOT to validate**:
- ❌ HTTP-specific concerns (required fields in request)
- ❌ Data format issues (JSON parsing, multipart form)
- ❌ File upload validations

**Error Types**: Use domain errors from `internal/core/errors/`:
- `ValidationError` - Business rule validation failed
- `BusinessRuleError` - Business constraint violated

## Validation Flow

```
HTTP Request
    ↓
┌─────────────────────────────────────────┐
│  Handler (Infrastructure Layer)         │
│  - Parse multipart form                 │
│  - Build request object                 │
└─────────────────┬───────────────────────┘
                  ↓
┌─────────────────────────────────────────┐
│  Contract.Validate() (Infrastructure)   │
│  - HTTP Validations                     │
│  - Required fields                      │
│  - File validations (size, type)        │
└─────────────────┬───────────────────────┘
                  ↓
┌─────────────────────────────────────────┐
│  UseCase (Application Layer)            │
│  - Coordinator only                     │
│  - NO validations here                  │
└─────────────────┬───────────────────────┘
                  ↓
┌─────────────────────────────────────────┐
│  Service (Application Layer)            │
│  - Calls Model.Validate()               │
└─────────────────┬───────────────────────┘
                  ↓
┌─────────────────────────────────────────┐
│  Model.Validate() (Domain Layer)        │
│  - Business Validations                 │
│  - Domain rules                         │
└─────────────────────────────────────────┘
```

## Implementation Examples

### HTTP Validation (Contract)

```go
// internal/infraestructure/adapters/http/contracts/product_create.go
package contracts

import (
    httpErrors "github.com/mlgaray/ecommerce_api/internal/infraestructure/adapters/http/errors"
)

type ProductCreateRequest struct {
    Product models.Product
    ShopID  int
    Images  []*multipart.FileHeader
}

func (r *ProductCreateRequest) Validate() error {
    // ✅ HTTP validation: required fields
    if strings.TrimSpace(r.Product.Name) == "" {
        return &httpErrors.BadRequestError{Message: "product_name_is_required"}
    }

    if strings.TrimSpace(r.Product.Description) == "" {
        return &httpErrors.BadRequestError{Message: "product_description_is_required"}
    }

    if r.Product.Category == nil || r.Product.Category.ID <= 0 {
        return &httpErrors.BadRequestError{Message: "category_id_is_required"}
    }

    // ✅ HTTP validation: shop_id required
    if r.ShopID <= 0 {
        return &httpErrors.BadRequestError{Message: "shop_id_is_required"}
    }

    // ✅ HTTP validation: images (size, type, quantity)
    if len(r.Images) == 0 {
        return &httpErrors.BadRequestError{Message: "at_least_one_image_is_required"}
    }

    for _, imageHeader := range r.Images {
        // Check file size (max 3MB)
        if imageHeader.Size > 3*1024*1024 {
            return &httpErrors.BadRequestError{Message: "image_size_too_large_max_3mb"}
        }

        // Check MIME type
        file, _ := imageHeader.Open()
        buffer := make([]byte, 512)
        file.Read(buffer)
        mimeType := http.DetectContentType(buffer)

        if !isValidImageType(mimeType) {
            return &httpErrors.BadRequestError{Message: "invalid_image_type_only_jpeg_png_allowed"}
        }
    }

    // ❌ DO NOT validate business rules here
    // Business validations are handled by Product.Validate() in the service layer
    return nil
}
```

### Business Validation (Domain Model)

```go
// internal/core/models/product.go
package models

import "github.com/mlgaray/ecommerce_api/internal/core/errors"

type Product struct {
    ID               int
    Name             string
    Description      string
    Price            float64
    Stock            int
    MinimumStock     int
    IsPromotional    bool
    PromotionalPrice float64
    Category         *Category
}

// Validate validates business rules for the Product domain model
func (p *Product) Validate() error {
    // ✅ Business rule: price must be positive
    if p.Price <= 0 {
        return &errors.ValidationError{
            Message: errors.ProductPriceMustBePositive,
        }
    }

    // ✅ Business rule: stock cannot be negative
    if p.Stock < 0 {
        return &errors.ValidationError{
            Message: errors.ProductStockCannotBeNegative,
        }
    }

    // ✅ Business rule: minimum stock cannot be negative
    if p.MinimumStock < 0 {
        return &errors.ValidationError{
            Message: errors.ProductMinimumStockCannotBeNegative,
        }
    }

    // ✅ Business rule: minimum stock can only exist if there's stock
    if p.MinimumStock > 0 && p.Stock == 0 {
        return &errors.ValidationError{
            Message: errors.MinimumStockRequiresStock,
        }
    }

    // ✅ Business rule: minimum stock cannot be greater than stock
    if p.Stock > 0 && p.MinimumStock > p.Stock {
        return &errors.ValidationError{
            Message: errors.ProductMinimumStockCannotBeGreaterThanStock,
        }
    }

    // ✅ Business rule: if promotional, must have promotional price
    if p.IsPromotional && p.PromotionalPrice <= 0 {
        return &errors.ValidationError{
            Message: errors.PromotionalProductRequiresPromotionalPrice,
        }
    }

    // ✅ Business rule: promotional price must be lower than regular price
    if p.IsPromotional && p.PromotionalPrice >= p.Price {
        return &errors.ValidationError{
            Message: errors.PromotionalPriceMustBeLowerThanRegularPrice,
        }
    }

    return nil
}
```

### Service Layer (Calls Domain Validation)

```go
// internal/application/services/product_service.go
package services

func (s *ProductService) Create(ctx context.Context, product *models.Product, imageBuffers [][]byte, shopID int) (*models.Product, error) {
    // ✅ Call domain validation - this is where business rules are checked
    if err := product.Validate(); err != nil {
        return nil, err
    }

    // Process images and create product
    // ...
    return s.productRepository.Create(ctx, product, shopID)
}

func (s *ProductService) Update(ctx context.Context, productID int, product *models.Product, newImageBuffers [][]byte) error {
    // ✅ Call domain validation - this is where business rules are checked
    if err := product.Validate(); err != nil {
        return err
    }

    // Process images and update product
    // ...
    return s.productRepository.Update(ctx, productID, product)
}
```

### Handler (HTTP Layer)

```go
// internal/infraestructure/adapters/http/product_handler.go
package http

func (p *ProductHandler) Create(w http.ResponseWriter, r *http.Request) {
    ctx := r.Context()

    // 1. Parse multipart form (HTTP concern)
    if err := r.ParseMultipartForm(13 << 20); err != nil {
        httpErrors.HandleError(w, &httpErrors.BadRequestError{Message: "error_parsing_multipart_form"})
        return
    }

    // 2. Build request (HTTP concern)
    request, err := p.buildProductCreateRequest(r)
    if err != nil {
        httpErrors.HandleError(w, err)
        return
    }

    // 3. Validate HTTP request (HTTP concern)
    if err := request.Validate(); err != nil {
        httpErrors.HandleError(w, err)
        return
    }

    // 4. Convert images (HTTP concern)
    imageBuffers, err := request.ToImageBuffers()
    if err != nil {
        httpErrors.HandleError(w, &httpErrors.BadRequestError{Message: err.Error()})
        return
    }

    // 5. Call service (business logic happens here via Product.Validate())
    createdProduct, err := p.createProduct.Execute(ctx, &request.Product, imageBuffers, request.ShopID)
    if err != nil {
        httpErrors.HandleError(w, err) // Maps domain errors to HTTP status
        return
    }

    // 6. Return response
    w.Header().Set("Content-Type", "application/json")
    w.WriteHeader(http.StatusCreated)
    json.NewEncoder(w).Encode(createdProduct)
}
```

## Error Message Naming Convention

All error messages use **snake_case** format and are defined in `internal/core/errors/messages.go`:

```go
// HTTP validation error messages (used with httpErrors.BadRequestError)
"product_name_is_required"
"product_description_is_required"
"category_id_is_required"
"shop_id_is_required"
"at_least_one_image_is_required"
"image_size_too_large_max_3mb"
"invalid_image_type_only_jpeg_png_allowed"

// Business validation error messages (used with domain errors)
"product_price_must_be_positive"
"product_stock_cannot_be_negative"
"product_minimum_stock_cannot_be_negative"
"minimum_stock_requires_stock"
"product_minimum_stock_cannot_be_greater_than_stock"
"promotional_product_requires_promotional_price"
"promotional_price_must_be_lower_than_regular_price"
```

## Testing Strategy

### HTTP Validation Tests (Integration Tests)

Located in: `tests/integration/features/`

Example scenarios:
```gherkin
# HTTP Validations (Infrastructure layer - Contracts)
Scenario: Create product with empty name
    Given I have product data with empty name
    When I send a create product request
    Then the response status should be 400
    And the user should receive an error message "product_name_is_required"

Scenario: Create product with oversized image
    Given I have product data with oversized image
    When I send a create product request
    Then the response status should be 400
    And the user should receive an error message "image_size_too_large_max_3mb"
```

### Business Validation Tests (Integration Tests)

```gherkin
# Business Validations (Domain layer - Product.Validate())
Scenario: Create product with negative price
    Given I have product data with negative price
    When I send a create product request
    Then the response status should be 400
    And the user should receive an error message "product_price_must_be_positive"

Scenario: Create product with minimum stock but no stock
    Given I have product data with minimum stock but no stock
    When I send a create product request
    Then the response status should be 400
    And the user should receive an error message "minimum_stock_requires_stock"
```

### Unit Tests (Domain Models)

```go
// internal/core/models/product_test.go
func TestProduct_Validate(t *testing.T) {
    t.Run("when price is negative then returns validation error", func(t *testing.T) {
        // Arrange
        product := &Product{
            Name:        "Test Product",
            Description: "Test Description",
            Price:       -10.00, // Invalid
            Stock:       10,
            Category:    &Category{ID: 1},
        }

        // Act
        err := product.Validate()

        // Assert
        assert.Error(t, err)
        assert.Equal(t, errors.ProductPriceMustBePositive, err.(*errors.ValidationError).Message)
    })

    t.Run("when promotional price is not lower than regular price then returns validation error", func(t *testing.T) {
        // Arrange
        product := &Product{
            Name:             "Test Product",
            Description:      "Test Description",
            Price:            100.00,
            IsPromotional:    true,
            PromotionalPrice: 120.00, // Higher than regular price
            Stock:            10,
            Category:         &Category{ID: 1},
        }

        // Act
        err := product.Validate()

        // Assert
        assert.Error(t, err)
        assert.Equal(t, errors.PromotionalPriceMustBeLowerThanRegularPrice, err.(*errors.ValidationError).Message)
    })
}
```

## Rules and Best Practices

### ✅ DO

1. **Separate HTTP and Business validations clearly**
   - HTTP validations in contracts
   - Business validations in domain models

2. **Call Model.Validate() in services**
   - Always call `product.Validate()` before persisting
   - This ensures business rules are always enforced

3. **Use appropriate error types**
   - `httpErrors.BadRequestError` for HTTP validations
   - `errors.ValidationError` for business validations
   - `errors.BusinessRuleError` for complex business constraints

4. **Centralize error messages**
   - Define all messages in `internal/core/errors/messages.go`
   - Use constants, never string literals

5. **Document validation layers in tests**
   - Add comments in feature files indicating which layer validates what
   ```gherkin
   # HTTP Validations (Infrastructure layer)
   Scenario: ...

   # Business Validations (Domain layer)
   Scenario: ...
   ```

### ❌ DON'T

1. **Don't mix validation concerns**
   - Never put business logic in HTTP contracts
   - Never put HTTP concerns in domain models

2. **Don't duplicate validations**
   - If it's a business rule, it belongs ONLY in the domain model
   - If it's an HTTP requirement, it belongs ONLY in the contract

3. **Don't skip domain validation**
   - Always call `Model.Validate()` in services
   - Never assume data is valid just because HTTP validation passed

4. **Don't validate in use cases**
   - Use cases are coordinators only
   - They should NOT contain validation logic

5. **Don't return HTTP errors from domain layer**
   - Domain models should never know about HTTP
   - Use domain errors (`ValidationError`, `BusinessRuleError`)

## Decision Tree

When adding a new validation, ask:

```
Is this about the HTTP request format/structure?
├─ YES → Add to Contract.Validate() (HTTP Layer)
│         Use httpErrors.BadRequestError
│
└─ NO → Is this a business rule or domain constraint?
         ├─ YES → Add to Model.Validate() (Domain Layer)
         │         Use errors.ValidationError or errors.BusinessRuleError
         │
         └─ NO → Reconsider if validation is needed
```

## Related Documentation

- [Error Handling](./ERRORS.md) - Complete error handling strategy
- [Architecture](./ARCHITECTURE.md) - Hexagonal Architecture principles
- [Layer Responsibilities](./LAYERS.md) - Responsibilities of each layer
- [Testing Strategy](./TESTING.md) - How to test validations

## Examples in Codebase

**HTTP Validations**:
- `internal/infraestructure/adapters/http/contracts/product_create.go`
- `internal/infraestructure/adapters/http/contracts/product_update.go`
- `internal/infraestructure/adapters/http/contracts/sign_in.go`
- `internal/infraestructure/adapters/http/contracts/sign_up.go`

**Business Validations**:
- `internal/core/models/product.go` - `Product.Validate()`
- `internal/core/models/user.go` - User domain validations

**Service Layer Validation Calls**:
- `internal/application/services/product_service.go` - Lines 26, 77
- `internal/application/services/auth_service.go`

**Integration Tests**:
- `tests/integration/features/create_product.feature`
- `tests/integration/features/update_product.feature`

---

**Remember**: The key principle is **Separation of Concerns**. HTTP validations ensure the request is well-formed, while business validations ensure domain invariants are maintained. This separation makes the code more maintainable, testable, and aligned with Hexagonal Architecture principles.
