# Error Handling

This project follows a **strict separation between Domain errors and HTTP errors** to maintain hexagonal architecture principles.

## Error Architecture

```
┌──────────────────────────────────────────────┐
│  HTTP Layer (Infrastructure)                 │
│  - httpErrors.BadRequestError                │
│  - Maps domain errors → HTTP status codes    │
└──────────────────┬───────────────────────────┘
                   │ uses ↓
┌──────────────────▼───────────────────────────┐
│  Domain Layer (Core)                         │
│  - RecordNotFoundError → 404                 │
│  - DuplicateRecordError → 409                │
│  - ValidationError → 400                     │
│  - AuthenticationError → 401                 │
│  - AuthorizationError → 403                  │
│  - BusinessRuleError → 422                   │
└──────────────────────────────────────────────┘
```

## Domain Errors (`internal/core/errors/`)

### domain_errors.go

Domain errors represent **business concerns**, NOT HTTP concerns:

```go
package errors

// RecordNotFoundError - Resource doesn't exist in domain
type RecordNotFoundError struct {
    Message string
}

// DuplicateRecordError - Constraint violation (unique email, etc)
type DuplicateRecordError struct {
    Message string
}

// ValidationError - Business rule validation failed
type ValidationError struct {
    Message string
}

// AuthenticationError - Identity verification failed
type AuthenticationError struct {
    Message string
}

// AuthorizationError - Permission denied
type AuthorizationError struct {
    Message string
}

// BusinessRuleError - Business constraint violated
type BusinessRuleError struct {
    Message string
}
```

### messages.go

All error messages in **snake_case** format:

```go
package errors

const (
    // User related
    UserNotFound           = "user_not_found"
    UserAlreadyExists      = "user_already_exists"
    InvalidUserCredentials = "invalid_credentials"

    // Product related
    ProductNotFound                = "product_not_found"
    ProductPriceMustBePositive     = "product_price_must_be_positive"
    ProductStockCannotBeNegative   = "product_stock_cannot_be_negative"
    InsufficientStock              = "insufficient_stock"

    // Authentication
    TokenExpired            = "token_expired"
    TokenInvalid            = "token_invalid"
    PasswordsCannotBeEmpty  = "passwords_cannot_be_empty"

    // Validation
    InvalidInput = "invalid_input"
)
```

## HTTP Errors (`internal/infraestructure/adapters/http/errors/`)

### http_errors.go

HTTP errors for **HTTP input validation only**:

```go
package errors

// BadRequestError - HTTP 400: Invalid HTTP input format, missing required fields, etc.
type BadRequestError struct {
    Message string
}

func (e *BadRequestError) Error() string {
    return e.Message
}
```

### handler.go

Maps domain errors to HTTP status codes:

```go
package errors

import (
    domainErrors "github.com/mlgaray/ecommerce_api/internal/core/errors"
)

func HandleError(w http.ResponseWriter, err error) {
    w.Header().Set("Content-Type", "application/json")

    var statusCode int
    var message string

    switch e := err.(type) {
    // HTTP Layer errors
    case *BadRequestError:
        statusCode = http.StatusBadRequest // 400
        message = e.Message

    // Domain errors → HTTP status
    case *domainErrors.RecordNotFoundError:
        statusCode = http.StatusNotFound // 404
        message = e.Message

    case *domainErrors.DuplicateRecordError:
        statusCode = http.StatusConflict // 409
        message = e.Message

    case *domainErrors.ValidationError:
        statusCode = http.StatusBadRequest // 400
        message = e.Message

    case *domainErrors.AuthenticationError:
        statusCode = http.StatusUnauthorized // 401
        message = e.Message

    case *domainErrors.AuthorizationError:
        statusCode = http.StatusForbidden // 403
        message = e.Message

    case *domainErrors.BusinessRuleError:
        statusCode = http.StatusUnprocessableEntity // 422
        message = e.Message

    default:
        // Technical/unexpected errors
        statusCode = http.StatusInternalServerError // 500
        message = "internal_server_error"
        // Log details but don't expose to client
    }

    response := map[string]string{"error": message}
    w.WriteHeader(statusCode)
    json.NewEncoder(w).Encode(response)
}
```

## Usage by Layer

### Handlers (Infrastructure Layer)

```go
import httpErrors "github.com/mlgaray/ecommerce_api/internal/infraestructure/adapters/http/errors"

func (h *ProductHandler) GetByID(w http.ResponseWriter, r *http.Request) {
    // HTTP validation
    productID, err := parseProductID(r)
    if err != nil {
        httpErrors.HandleError(w, &httpErrors.BadRequestError{Message: "invalid_product_id_format"})
        return
    }

    // Business logic
    product, err := h.getByID.Execute(ctx, productID)
    if err != nil {
        httpErrors.HandleError(w, err) // Maps domain error → HTTP status
        return
    }

    json.NewEncoder(w).Encode(product)
}
```

### Contracts (Infrastructure Layer - HTTP Validation)

```go
import httpErrors "github.com/mlgaray/ecommerce_api/internal/infraestructure/adapters/http/errors"

func (r *SignInRequest) Validate() error {
    // HTTP validation: format, required fields
    if strings.TrimSpace(r.Email) == "" {
        return &httpErrors.BadRequestError{Message: "email_is_required"}
    }

    if !emailRegex.MatchString(r.Email) {
        return &httpErrors.BadRequestError{Message: "invalid_email_format"}
    }

    return nil
}
```

### Services (Application Layer)

```go
import "github.com/mlgaray/ecommerce_api/internal/core/errors"

func (s *AuthService) ComparePassword(ctx context.Context, hashedPassword, password string) error {
    // Domain validation
    if hashedPassword == "" || password == "" {
        return &errors.ValidationError{Message: errors.PasswordsCannotBeEmpty}
    }

    if hashedPassword != password {
        return &errors.AuthenticationError{Message: errors.InvalidUserCredentials}
    }

    return nil
}
```

### Repositories (Infrastructure Layer)

```go
import "github.com/mlgaray/ecommerce_api/internal/core/errors"

func (r *UserRepository) GetByEmail(ctx context.Context, email string) (*models.User, error) {
    err := row.Scan(&user.ID, &user.Email, &user.Password)
    if err == sql.ErrNoRows {
        return nil, &errors.RecordNotFoundError{Message: errors.UserNotFound}
    }

    if err != nil {
        // Technical error - use fmt.Errorf
        return nil, fmt.Errorf("database operation failed: %w", err)
    }

    return user, nil
}

func (r *UserRepository) handlePostgreSQLError(err error, email string) error {
    if pqErr, ok := err.(*pq.Error); ok {
        if pqErr.Code == "23505" && pqErr.Constraint == "users_email_key" {
            return &errors.DuplicateRecordError{Message: errors.UserAlreadyExists}
        }
    }
    // Technical error
    return fmt.Errorf("failed to create user: %w", err)
}
```

### Rich Domain Models

```go
import "github.com/mlgaray/ecommerce_api/internal/core/errors"

func (p *Product) Validate() error {
    if p.Price <= 0 {
        return &errors.ValidationError{Message: errors.ProductPriceMustBePositive}
    }

    if p.Stock < 0 {
        return &errors.ValidationError{Message: errors.ProductStockCannotBeNegative}
    }

    return nil
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
```

## Validation Layers

### HTTP Validation (Infrastructure)

Validates HTTP-specific concerns:
- Required HTTP fields
- JSON format
- Email format
- URL format
- File upload format

Uses: `httpErrors.BadRequestError`

### Domain Validation (Core)

Validates business rules:
- Price must be positive
- Stock cannot be negative
- Promotional price < regular price
- Sufficient stock for order

Uses: `errors.ValidationError` or `errors.BusinessRuleError`

## Error Flow Example

```
1. HTTP Request arrives
2. Handler validates HTTP format → httpErrors.BadRequestError
3. Contract validates HTTP input → httpErrors.BadRequestError
4. Use Case calls Service
5. Service validates business rules → errors.ValidationError
6. Repository queries database
7. Repository returns → errors.RecordNotFoundError
8. Error bubbles up to Handler
9. httpErrors.HandleError() maps to HTTP 404
10. Client receives {"error": "user_not_found"}
```

## Rules

✅ **DO:**
- Use `httpErrors.BadRequestError` for HTTP validation in handlers/contracts
- Use domain errors (`ValidationError`, `AuthenticationError`, etc.) in services/repositories
- Use `fmt.Errorf()` for technical/unexpected errors
- Keep error messages in snake_case
- Centralize messages in `errors/messages.go`

❌ **DON'T:**
- Return HTTP errors from services or repositories
- Return domain errors from HTTP validation in contracts
- Expose technical error details to clients
- Use string literals for error messages
- Mix HTTP concerns with domain concerns
