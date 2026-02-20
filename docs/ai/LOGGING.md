# Logging Patterns

This project uses **structured logging** with consistent field naming across all layers.

## Log Structure

All logs follow this format:

```go
logs.WithFields(map[string]interface{}{
    "file":     "repository_name",     // Constant: UserRepositoryField
    "function": "function_name",       // Constant: GetByEmailFunctionField
    "sub_func": "specific_operation",  // Optional: "db.QueryContext", "rows.Scan"
    // ... contextual data (user_id, email, etc)
    "error":    err.Error(),           // Always last if present
}).Error("Human-readable message")
```

## Field Naming Conventions

### Constants with `Field` Suffix

All log field constants MUST have `Field` suffix:

```go
const (
    UserRepositoryField           = "user_repository"
    UserCreateFunctionField       = "create"
    UserGetByEmailFunctionField   = "get_by_email"
    ScanUserWithRolesSubFuncField = "scan_user_with_roles"
)
```

### Sub-Function Specificity

Use **specific method names**, not generic labels:

```go
// ✅ GOOD: Specific operation
"sub_func": "db.QueryContext"
"sub_func": "tx.QueryContext"
"sub_func": "rows.Next"
"sub_func": "rows.Scan"
"sub_func": "json.Unmarshal"
"sub_func": "json.Marshal"
"sub_func": "strconv.Atoi"

// ❌ BAD: Too generic
"sub_func": "db"
"sub_func": "query"
"sub_func": "scan"
```

## Patterns by Layer

### Repository Layer (Technical Only)

Focus on **database operations** and technical details:

```go
const (
    UserRepositoryField           = "user_repository"
    UserCreateFunctionField       = "create"
    UserGetByEmailFunctionField   = "get_by_email"
)

func (r *UserRepository) GetByEmail(ctx context.Context, email string) (*models.User, error) {
    row := r.db.QueryRowContext(ctx, query, email)

    err := row.Scan(&user.ID, &user.Email, &user.Password)
    if err == sql.ErrNoRows {
        logs.WithFields(map[string]interface{}{
            "file":     UserRepositoryField,
            "function": UserGetByEmailFunctionField,
            "email":    email,
        }).Warn("User not found")
        return nil, &errors.RecordNotFoundError{Message: errors.UserNotFound}
    }

    if err != nil {
        logs.WithFields(map[string]interface{}{
            "file":     UserRepositoryField,
            "function": UserGetByEmailFunctionField,
            "sub_func": "rows.Scan",
            "email":    email,
            "error":    err.Error(),
        }).Error("Database scan failed")
        return nil, fmt.Errorf("failed to scan user")
    }

    return user, nil
}
```

### Service Layer (Business Context)

Focus on **business operations**:

```go
const (
    UserServiceField                     = "user_service"
    ValidateCredentialsFunctionField     = "validate_credentials"
    ComparePasswordSubFuncField          = "compare_password"
)

func (s *UserService) ValidateCredentials(ctx context.Context, user *models.User, password string) (*models.User, error) {
    err := s.authService.ComparePassword(ctx, user.Password, password)
    if err != nil {
        logs.WithFields(map[string]interface{}{
            "file":     UserServiceField,
            "function": ValidateCredentialsFunctionField,
            "sub_func": ComparePasswordSubFuncField,
            "error":    err.Error(),
        }).Error("Error comparing passwords")
        return nil, &errors.AuthenticationError{Message: errors.InvalidUserCredentials}
    }

    return user, nil
}
```

### Handler Layer (HTTP Context)

Focus on **HTTP operations**:

```go
const (
    ProductHandlerField           = "product_handler"
    GetByIDFunctionField          = "get_by_id"
    ParseProductIDSubFuncField    = "parse_product_id"
)

func (h *ProductHandler) GetByID(w http.ResponseWriter, r *http.Request) {
    ctx := r.Context()

    productID, err := h.parseProductID(r)
    if err != nil {
        httpErrors.HandleError(w, err)
        return
    }

    product, err := h.getByID.Execute(ctx, productID)
    if err != nil {
        logs.WithFields(map[string]interface{}{
            "file":       ProductHandlerField,
            "function":   GetByIDFunctionField,
            "product_id": productID,
            "error":      err.Error(),
        }).Error("Error retrieving product")
        httpErrors.HandleError(w, err)
        return
    }

    w.Header().Set("Content-Type", "application/json")
    w.WriteHeader(http.StatusOK)
    if err := json.NewEncoder(w).Encode(product); err != nil {
        logs.WithFields(map[string]interface{}{
            "file":     ProductHandlerField,
            "function": GetByIDFunctionField,
            "sub_func": "json.Encode",
            "error":    err.Error(),
        }).Error("Error encoding response")
    }
}
```

## Field Order

Always maintain this order:

1. `file` - Repository/Service/Handler name
2. `function` - Main function name
3. `sub_func` - Specific operation (optional)
4. **Contextual data** - Business/domain data (user_id, email, product_id, etc.)
5. `error` - **Always last** if present

```go
// ✅ GOOD: Correct order
logs.WithFields(map[string]interface{}{
    "file":       ProductRepositoryField,
    "function":   ProductGetByIDFunctionField,
    "sub_func":   "rows.Scan",
    "product_id": productID,
    "shop_id":    shopID,
    "error":      err.Error(),
}).Error("Failed to scan product")

// ❌ BAD: Wrong order
logs.WithFields(map[string]interface{}{
    "error":      err.Error(),  // error not last
    "product_id": productID,
    "function":   ProductGetByIDFunctionField,
    "file":       ProductRepositoryField,
}).Error("Failed to scan product")
```

## Message Format

- **snake_case** for error field values
- **Human-readable** for log messages
- **Concise** but descriptive

```go
// ✅ GOOD
logs.WithFields(map[string]interface{}{
    "file":  UserRepositoryField,
    "error": "user_not_found",
}).Error("User not found in database")

// ❌ BAD
logs.WithFields(map[string]interface{}{
    "file":  "user repository",  // should be constant
    "error": "User Not Found",   // should be snake_case
}).Error("error")  // too generic
```

## Examples by Scenario

### Database Query Error

```go
logs.WithFields(map[string]interface{}{
    "file":     ProductRepositoryField,
    "function": ProductGetAllByShopIDFunctionField,
    "sub_func": "db.QueryContext",
    "shop_id":  shopID,
    "limit":    limit,
    "cursor":   cursor,
    "error":    err.Error(),
}).Error("Failed to query products by shop")
```

### Validation Error

```go
logs.WithFields(map[string]interface{}{
    "file":       ProductHandlerField,
    "function":   CreateProductFunctionField,
    "sub_func":   "request.Validate",
    "shop_id":    request.ShopID,
    "error":      err.Error(),
}).Error("Product validation failed")
```

### PostgreSQL Constraint Violation

```go
logs.WithFields(map[string]interface{}{
    "file":       UserRepositoryField,
    "function":   UserCreateFunctionField,
    "constraint": pqErr.Constraint,
    "email":      email,
}).Error("User with email already exists")
```

### JSON Encoding Error

```go
logs.WithFields(map[string]interface{}{
    "file":     AuthHandlerField,
    "function": SignInFunctionField,
    "sub_func": "json.Marshal",
    "error":    err.Error(),
}).Error("Failed to encode response")
```

## Context-Aware Logging (Advanced)

For request-scoped logging with request IDs:

```go
logs.FromContext(ctx).WithFields(map[string]interface{}{
    "endpoint":   "/api/products",
    "method":     "GET",
    "request_id": requestID,
    "user_id":    userID,
}).Info("Request received")
```

## Log Levels

- **Error**: Operation failed, needs investigation
- **Warn**: Unexpected but handled (e.g., user not found)
- **Info**: Significant business events (e.g., user created)
- **Debug**: Development debugging (avoid in production)

```go
// Error: Something went wrong
logs.WithFields(...).Error("Database query failed")

// Warn: Expected but noteworthy
logs.WithFields(...).Warn("User not found")

// Info: Business event
logs.WithFields(...).Info("User successfully created")
```

## Anti-Patterns to Avoid

```go
// ❌ DON'T: Use string literals
logs.WithFields(map[string]interface{}{
    "file": "product_handler",  // Use constant: ProductHandlerField
})

// ❌ DON'T: Generic sub_func
logs.WithFields(map[string]interface{}{
    "sub_func": "db",  // Use specific: "db.QueryContext"
})

// ❌ DON'T: Error not last
logs.WithFields(map[string]interface{}{
    "error":    err.Error(),
    "user_id":  userID,  // error should be last
})

// ❌ DON'T: Missing Field suffix
const (
    UserRepository = "user_repository"  // Should be: UserRepositoryField
)
```
