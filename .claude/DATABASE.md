# Database Patterns

This project uses PostgreSQL with golang-migrate for schema management and transaction support.

## Database Technology

- **Database**: PostgreSQL
- **Driver**: lib/pq
- **Migrations**: golang-migrate
- **Connection**: database/sql with connection pooling

## Repository Pattern

### Dual-Context Support

Most repositories support both **direct DB operations** and **transaction contexts**:

```go
func (r *ProductRepository) GetByID(ctx context.Context, productID int) (*models.Product, error) {
    // Automatically detect transaction from context
    if tx, ok := ctx.Value(TxContextKey).(*sql.Tx); ok {
        return r.getByIDWithTx(ctx, tx, productID)
    }
    return r.getByIDWithDB(ctx, productID)
}

// Direct DB operation
func (r *ProductRepository) getByIDWithDB(ctx context.Context, productID int) (*models.Product, error) {
    row := r.db.QueryRowContext(ctx, query, productID)
    // ... scan logic
}

// Transaction operation
func (r *ProductRepository) getByIDWithTx(ctx context.Context, tx *sql.Tx, productID int) (*models.Product, error) {
    row := tx.QueryRowContext(ctx, query, productID)
    // ... scan logic
}
```

### Transaction Management

**Context-Based with TxContextKey**:

```go
const TxContextKey = "tx"

// Start transaction
tx, err := db.BeginTx(ctx, nil)
if err != nil {
    return err
}
defer tx.Rollback() // Rollback if not committed

// Add transaction to context
ctx = context.WithValue(ctx, TxContextKey, tx)

// Repository methods automatically use transaction
user, err := userRepo.Create(ctx, user)
if err != nil {
    return err  // Rollback happens in defer
}

shop, err := shopRepo.Create(ctx, shop)
if err != nil {
    return err  // Rollback happens in defer
}

// Commit if all operations succeed
return tx.Commit()
```

## Migrations

### Structure

```
database/
  migrations/
    000001_create_table_users.up.sql
    000001_create_table_users.down.sql
    000002_create_table_roles.up.sql
    000002_create_table_roles.down.sql
  migrations/seeds/
    000001_insert_into_users.up.sql
    000001_insert_into_users.down.sql
```

### Migration Commands

```bash
# Create new migration
make migrate-create MIGRATION_NAME="create_table_products"

# Run migrations (up)
make migrate-up

# Rollback migrations (down)
make migrate-down

# Check migration status
migrate -path database/migrations -database "postgresql://..." version
```

### Migration Example

**Up Migration** (`000007_create_table_products.up.sql`):
```sql
CREATE TABLE IF NOT EXISTS products (
    id SERIAL PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    description TEXT,
    price DECIMAL(10, 2) NOT NULL CHECK (price >= 0),
    stock INTEGER NOT NULL DEFAULT 0 CHECK (stock >= 0),
    minimum_stock INTEGER NOT NULL DEFAULT 0 CHECK (minimum_stock >= 0),
    is_promotional BOOLEAN NOT NULL DEFAULT FALSE,
    promotional_price DECIMAL(10, 2) CHECK (promotional_price >= 0),
    is_active BOOLEAN NOT NULL DEFAULT TRUE,
    shop_id INTEGER NOT NULL REFERENCES shops(id) ON DELETE CASCADE,
    category_id INTEGER NOT NULL REFERENCES categories(id),
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

-- Indexes
CREATE INDEX idx_products_shop_id ON products(shop_id);
CREATE INDEX idx_products_category_id ON products(category_id);
CREATE INDEX idx_products_is_active ON products(is_active);

-- Check constraints
ALTER TABLE products ADD CONSTRAINT check_promotional_price
    CHECK (NOT is_promotional OR (is_promotional AND promotional_price IS NOT NULL AND promotional_price < price));
```

**Down Migration** (`000007_create_table_products.down.sql`):
```sql
DROP TABLE IF EXISTS products CASCADE;
```

### Seed Data

Use for development/testing data:

```bash
# Create seed migration
make migrate-create-seeds MIGRATION_NAME="seed_users"

# Run seeds
make migrate-up-seeds
```

**Seed Example** (`000001_insert_into_users.up.sql`):
```sql
INSERT INTO users (name, last_name, email, password, phone, created_at, updated_at)
VALUES
    ('Admin', 'User', 'admin@example.com', 'hashed_password', '+1234567890', CURRENT_TIMESTAMP, CURRENT_TIMESTAMP),
    ('Test', 'User', 'test@example.com', 'hashed_password', '+0987654321', CURRENT_TIMESTAMP, CURRENT_TIMESTAMP)
ON CONFLICT (email) DO NOTHING;
```

## Error Handling

### PostgreSQL-Specific Errors

Handle PostgreSQL constraint violations and map to domain errors:

```go
import "github.com/lib/pq"

func (r *UserRepository) handlePostgreSQLError(err error, email string) error {
    pqErr, ok := err.(*pq.Error)
    if !ok {
        // Not a PostgreSQL error
        return fmt.Errorf("database operation failed: %w", err)
    }

    // Unique constraint violation
    if pqErr.Code == "23505" && pqErr.Constraint == "users_email_key" {
        logs.WithFields(map[string]interface{}{
            "file":       UserRepositoryField,
            "function":   UserCreateFunctionField,
            "constraint": pqErr.Constraint,
            "email":      email,
        }).Error("User with email already exists")

        return &errors.DuplicateRecordError{Message: errors.UserAlreadyExists}
    }

    // Foreign key violation
    if pqErr.Code == "23503" {
        logs.WithFields(map[string]interface{}{
            "file":       UserRepositoryField,
            "function":   UserCreateFunctionField,
            "constraint": pqErr.Constraint,
        }).Error("Foreign key constraint violation")

        return &errors.ValidationError{Message: errors.InvalidInput}
    }

    // Other database errors
    return fmt.Errorf("database operation failed: %w", err)
}
```

### Common PostgreSQL Error Codes

- `23505` - Unique violation
- `23503` - Foreign key violation
- `23502` - Not null violation
- `23514` - Check constraint violation

### No Rows Found

```go
err := row.Scan(&user.ID, &user.Email)
if err == sql.ErrNoRows {
    return nil, &errors.RecordNotFoundError{Message: errors.UserNotFound}
}

if err != nil {
    // Technical error
    return nil, fmt.Errorf("database scan failed: %w", err)
}
```

## Query Patterns

### Simple Query

```go
func (r *ProductRepository) GetByID(ctx context.Context, productID int) (*models.Product, error) {
    query := `
        SELECT id, name, price, stock
        FROM products
        WHERE id = $1
    `

    var product models.Product
    err := r.db.QueryRowContext(ctx, query, productID).Scan(
        &product.ID,
        &product.Name,
        &product.Price,
        &product.Stock,
    )

    if err == sql.ErrNoRows {
        return nil, &errors.RecordNotFoundError{Message: errors.ProductNotFound}
    }

    if err != nil {
        return nil, fmt.Errorf("database operation failed: %w", err)
    }

    return &product, nil
}
```

### Query with Multiple Rows

```go
func (r *ProductRepository) GetAllByShopID(ctx context.Context, shopID, limit, cursor int) ([]*models.Product, error) {
    query := `
        SELECT id, name, price, stock
        FROM products
        WHERE shop_id = $1 AND id > $2
        ORDER BY id ASC
        LIMIT $3
    `

    rows, err := r.db.QueryContext(ctx, query, shopID, cursor, limit)
    if err != nil {
        return nil, fmt.Errorf("database query failed: %w", err)
    }
    defer rows.Close()

    var products []*models.Product
    for rows.Next() {
        var product models.Product
        err := rows.Scan(&product.ID, &product.Name, &product.Price, &product.Stock)
        if err != nil {
            return nil, fmt.Errorf("database scan failed: %w", err)
        }
        products = append(products, &product)
    }

    if err := rows.Err(); err != nil {
        return nil, fmt.Errorf("database rows error: %w", err)
    }

    return products, nil
}
```

### Insert with RETURNING

```go
func (r *ProductRepository) Create(ctx context.Context, product *models.Product) (*models.Product, error) {
    query := `
        INSERT INTO products (name, price, stock, shop_id)
        VALUES ($1, $2, $3, $4)
        RETURNING id, created_at, updated_at
    `

    err := r.db.QueryRowContext(ctx, query,
        product.Name,
        product.Price,
        product.Stock,
        product.ShopID,
    ).Scan(&product.ID, &product.CreatedAt, &product.UpdatedAt)

    if err != nil {
        return nil, r.handlePostgreSQLError(err, product.Name)
    }

    return product, nil
}
```

### Update

```go
func (r *ProductRepository) Update(ctx context.Context, productID int, product *models.Product) error {
    query := `
        UPDATE products
        SET name = $1, price = $2, stock = $3, updated_at = CURRENT_TIMESTAMP
        WHERE id = $4
    `

    result, err := r.db.ExecContext(ctx, query,
        product.Name,
        product.Price,
        product.Stock,
        productID,
    )

    if err != nil {
        return fmt.Errorf("database update failed: %w", err)
    }

    rowsAffected, err := result.RowsAffected()
    if err != nil {
        return fmt.Errorf("failed to get rows affected: %w", err)
    }

    if rowsAffected == 0 {
        return &errors.RecordNotFoundError{Message: errors.ProductNotFound}
    }

    return nil
}
```

## JSON Aggregation

For complex relationships, use PostgreSQL JSON functions:

```go
query := `
    SELECT
        p.id, p.name, p.price,
        COALESCE(
            json_agg(
                json_build_object('id', pi.id, 'url', pi.url, 'order', pi."order")
                ORDER BY pi."order"
            ) FILTER (WHERE pi.id IS NOT NULL),
            '[]'
        ) as images
    FROM products p
    LEFT JOIN product_images pi ON p.id = pi.product_id
    WHERE p.id = $1
    GROUP BY p.id
`

var imagesJSON []byte
err := row.Scan(&product.ID, &product.Name, &product.Price, &imagesJSON)

// Unmarshal JSON
if err := json.Unmarshal(imagesJSON, &product.Images); err != nil {
    return nil, fmt.Errorf("failed to unmarshal images: %w", err)
}
```

## Connection Pool Configuration

```go
db, err := sql.Open("postgres", connectionString)
if err != nil {
    return nil, err
}

// Configure connection pool
db.SetMaxOpenConns(25)
db.SetMaxIdleConns(5)
db.SetConnMaxLifetime(5 * time.Minute)

// Verify connection
if err := db.Ping(); err != nil {
    return nil, err
}
```

## Best Practices

✅ **DO:**
- Use parameterized queries ($1, $2, etc.) to prevent SQL injection
- Close rows with `defer rows.Close()`
- Check `rows.Err()` after iteration
- Use transactions for multi-table operations
- Handle `sql.ErrNoRows` separately from other errors
- Use `COALESCE` for nullable JSON aggregations
- Add appropriate indexes for query performance

❌ **DON'T:**
- Build queries with string concatenation
- Forget to close resources (rows, tx)
- Ignore `RowsAffected` for UPDATE/DELETE
- Use SELECT * (specify columns explicitly)
- Leave transactions uncommitted
- Expose PostgreSQL errors directly to clients
