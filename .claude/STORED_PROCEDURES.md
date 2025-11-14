# Stored Procedures - Error Handling Guide

## Exception Handling in PostgreSQL Stored Procedures

### In PL/pgSQL (Database):

```sql
CREATE OR REPLACE FUNCTION create_product(...) RETURNS INTEGER AS $$
BEGIN
    -- Business logic here
    INSERT INTO products ...

EXCEPTION
    WHEN OTHERS THEN
        RAISE EXCEPTION 'Error creating product: %', SQLERRM;
END;
$$ LANGUAGE plpgsql;
```

**What happens:**
- `EXCEPTION WHEN OTHERS` catches any error during execution
- `RAISE EXCEPTION` sends error back to Go
- `SQLERRM` contains the original error message
- Transaction is automatically rolled back

---

## Error Handling in Go

### How PostgreSQL Errors Arrive in Go:

When a stored procedure raises an exception, Go receives it as a `*pq.Error`:

```go
type pq.Error struct {
    Code     string  // PostgreSQL error code (e.g., "P0001" for RAISE EXCEPTION)
    Message  string  // Text from RAISE EXCEPTION
    Detail   string  // Optional detail
    Hint     string  // Optional hint
    Position string  // Position in query if applicable
    ...
}
```

### Current Implementation Pattern:

```go
err = r.db.QueryRowContext(ctx, `SELECT create_product(...)`, params...).Scan(&productID)

if err != nil {
    // Log original error
    logs.WithFields(map[string]interface{}{
        "error": err.Error(),
    }).Error("Failed to create product via stored procedure")

    // Type assertion to check if it's a PostgreSQL error
    if pqErr, ok := err.(*pq.Error); ok {
        // Extract PostgreSQL error details
        logs.WithFields(map[string]interface{}{
            "pg_code":    pqErr.Code,        // "P0001" for user-defined exception
            "pg_message": pqErr.Message,     // "Error creating product: ..."
            "pg_detail":  pqErr.Detail,
            "pg_hint":    pqErr.Hint,
        }).Debug("PostgreSQL error details from stored procedure")

        // Return error preserving the SP's message
        return nil, fmt.Errorf("stored procedure error: %s", pqErr.Message)
    }

    // Not a PostgreSQL error (network, context cancelled, etc.)
    return nil, fmt.Errorf("database operation failed: %w", err)
}
```

---

## Error Flow Example

### Scenario: Invalid JSON causes type casting error

**1. PostgreSQL stored procedure:**
```sql
-- This fails if JSON is malformed
(v_variant->>'order')::INTEGER
```

**2. PL/pgSQL catches it:**
```sql
EXCEPTION
    WHEN OTHERS THEN
        RAISE EXCEPTION 'Error creating product: %', SQLERRM;
        -- Message becomes: "Error creating product: invalid input syntax for type integer"
```

**3. Go receives `*pq.Error`:**
```go
pqErr.Code = "P0001"  // User-defined exception
pqErr.Message = "Error creating product: invalid input syntax for type integer"
```

**4. Logged as:**
```json
{
  "level": "error",
  "error": "pq: Error creating product: invalid input syntax for type integer",
  "pg_code": "P0001",
  "pg_message": "Error creating product: invalid input syntax for type integer"
}
```

**5. Returned to service layer:**
```go
return nil, fmt.Errorf("stored procedure error: Error creating product: invalid input syntax for type integer")
```

---

## Common PostgreSQL Error Codes

| Code | Name | Description |
|------|------|-------------|
| `P0001` | RAISE_EXCEPTION | User-defined exception (our RAISE EXCEPTION) |
| `23505` | UNIQUE_VIOLATION | Duplicate key violates unique constraint |
| `23503` | FOREIGN_KEY_VIOLATION | Foreign key constraint violation |
| `23502` | NOT_NULL_VIOLATION | NOT NULL constraint violation |
| `22P02` | INVALID_TEXT_REPRESENTATION | Invalid input syntax (e.g., casting error) |
| `42P01` | UNDEFINED_TABLE | Table does not exist |

**Reference:** https://www.postgresql.org/docs/current/errcodes-appendix.html

---

## Advanced Error Handling (Optional Enhancement)

For more granular error handling, you can create custom error types:

```go
// Define domain errors
var (
    ErrInvalidData     = errors.New("invalid data provided")
    ErrConstraintViolation = errors.New("database constraint violation")
    ErrNotFound        = errors.New("record not found")
)

// Helper function to classify PostgreSQL errors
func classifyPQError(pqErr *pq.Error) error {
    switch pqErr.Code {
    case "23505": // UNIQUE_VIOLATION
        return fmt.Errorf("%w: %s", ErrConstraintViolation, pqErr.Message)
    case "23503": // FOREIGN_KEY_VIOLATION
        return fmt.Errorf("%w: %s", ErrConstraintViolation, pqErr.Message)
    case "22P02": // INVALID_TEXT_REPRESENTATION
        return fmt.Errorf("%w: %s", ErrInvalidData, pqErr.Message)
    case "P0001": // RAISE_EXCEPTION (custom)
        // Parse message to determine type
        if strings.Contains(pqErr.Message, "not found") {
            return fmt.Errorf("%w: %s", ErrNotFound, pqErr.Message)
        }
        return fmt.Errorf("stored procedure error: %s", pqErr.Message)
    default:
        return fmt.Errorf("database error [%s]: %s", pqErr.Code, pqErr.Message)
    }
}
```

**Usage:**
```go
if pqErr, ok := err.(*pq.Error); ok {
    return nil, classifyPQError(pqErr)
}
```

---

## Best Practices

### ✅ Do:
- Always log the full error for debugging (`err.Error()`)
- Extract and log PostgreSQL details (`pqErr.Code`, `pqErr.Message`)
- Preserve meaningful error messages from stored procedures
- Use type assertion to detect `*pq.Error`
- Wrap errors with context using `fmt.Errorf("context: %w", err)`

### ❌ Don't:
- Return generic "database operation failed" and lose context
- Expose internal database details to end users
- Ignore error codes (they're valuable for classification)
- Swallow errors silently

---

## Testing Error Handling

```go
func TestCreateWithStoredProcedure_InvalidData(t *testing.T) {
    // Mock invalid data that will cause SP to fail
    product := &models.Product{
        Name: "", // Empty name should fail validation
        // ...
    }

    _, err := repo.CreateWithStoredProcedure(ctx, product, shopID)

    // Assert error contains SP context
    assert.Error(t, err)
    assert.Contains(t, err.Error(), "stored procedure error")
}
```

---

## Summary

**Flow:**
1. Stored procedure encounters error
2. `EXCEPTION WHEN OTHERS` catches it
3. `RAISE EXCEPTION` sends it to Go
4. Go receives as `*pq.Error`
5. Type assertion extracts details
6. Log structured data + return meaningful error
7. Service layer can handle or propagate

**Key benefit:** Preserve error context from database to application layer for better debugging and error reporting.
