# Database Migrations

This document explains the migration structure and workflow for this project.

## Directory Structure

```
database/migrations/
├── 000001-000010_*.sql     # Tables (core schema)
├── seeds/                   # Initial data
│   └── 000001-00000N_*.sql
└── functions/              # Stored procedures & functions
    └── 000001-00000N_*.sql
```

## Migration Types

### 1. Tables (Root Directory)
**Location**: `database/migrations/*.sql`
**Naming**: `{version}_create_table_{table_name}.{up|down}.sql`
**Purpose**: Core database schema (tables, indexes, constraints)
**Commands**:
```bash
make migrate-up         # Apply table migrations
make migrate-down       # Rollback table migrations
```

**Examples**:
- `000001_create_table_users.up.sql`
- `000002_create_table_products.up.sql`

---

### 2. Seeds (Initial Data)
**Location**: `database/migrations/seeds/*.sql`
**Naming**: `{version}_insert_into_{table_name}.{up|down}.sql`
**Purpose**: Initial/test data for development and testing
**Commands**:
```bash
make migrate-up-seeds       # Insert seed data
make migrate-down-seeds     # Remove seed data
make migrate-seeds-version  # Check seeds version
```

**Tracking Table**: `schema_seeds` (separate from main migrations)

**Examples**:
- `000001_insert_into_users.up.sql`
- `000002_insert_into_roles.up.sql`

---

### 3. Functions (Stored Procedures)
**Location**: `database/migrations/functions/*.sql`
**Naming**: `{version}_create_function_{function_name}.{up|down}.sql`
**Purpose**: PostgreSQL functions and stored procedures
**Commands**:
```bash
make migrate-up-functions       # Create functions
make migrate-down-functions     # Drop functions
make migrate-functions-version  # Check functions version
```

**Tracking Table**: `schema_functions` (separate from main migrations)

**Examples**:
- `000001_create_function_create_product_with_relations.up.sql`
- `000002_create_function_calculate_order_total.up.sql`

---

## Migration Workflow

### Creating New Migrations

**Tables**:
```bash
make migrate-create MIGRATION_NAME="create_table_orders"
# Creates: 000011_create_table_orders.up.sql + .down.sql
```

**Seeds**:
```bash
make migrate-create-seeds MIGRATION_NAME="insert_into_orders"
# Creates: seeds/000011_insert_into_orders.up.sql + .down.sql
```

**Functions** (manual creation):
```bash
# Create manually in functions/ directory
# Numbering starts at 000001 independently
touch database/migrations/functions/000002_create_function_myfunction.up.sql
touch database/migrations/functions/000002_create_function_myfunction.down.sql
```

### Running Migrations

**Development Setup** (fresh database):
```bash
make migrate-up              # 1. Create tables
make migrate-up-seeds        # 2. Insert seed data
make migrate-up-functions    # 3. Create functions
```

**Production Deployment**:
```bash
make migrate-up              # Tables only (no seeds)
make migrate-up-functions    # Functions
```

---

## Why Separate Directories?

### ✅ Benefits

1. **Independent Versioning**: Each type has its own numbering (all start at 000001)
2. **Clear Separation**: Easy to find tables vs seeds vs functions
3. **Selective Application**:
   - Production: tables + functions (NO seeds)
   - Development: tables + functions + seeds
4. **Independent Rollback**: Rollback functions without affecting tables
5. **Better Git History**: Changes to functions don't conflict with table changes

### 📊 Tracking Tables

Each migration type uses a separate tracking table:

| Directory | Tracking Table | Purpose |
|-----------|---------------|---------|
| `./` | `schema_migrations` | Tables structure |
| `seeds/` | `schema_seeds` | Development data |
| `functions/` | `schema_functions` | Stored procedures |

This prevents version conflicts and allows independent management.

---

## Best Practices

### DO ✅
- Keep table migrations simple (one table per migration)
- Document complex functions with comments
- Test `.down.sql` rollback scripts
- Use descriptive migration names
- Add functions for performance-critical operations

### DON'T ❌
- Don't mix table changes with seed data
- Don't commit sensitive data in seeds
- Don't create cross-dependencies between types
- Don't skip version numbers
- Don't modify existing migrations (create new ones)

---

## Examples

### Table Migration
```sql
-- 000011_create_table_orders.up.sql
CREATE TABLE orders (
    id SERIAL PRIMARY KEY,
    user_id INTEGER NOT NULL REFERENCES users(id),
    total DECIMAL(10,2) NOT NULL,
    created_at TIMESTAMP DEFAULT NOW()
);
```

### Seed Migration
```sql
-- seeds/000011_insert_into_test_orders.up.sql
INSERT INTO orders (user_id, total) VALUES
(1, 100.00),
(2, 200.00);
```

### Function Migration
```sql
-- functions/000002_create_function_calculate_total.up.sql
CREATE OR REPLACE FUNCTION calculate_order_total(p_order_id INT)
RETURNS DECIMAL AS $$
BEGIN
    RETURN (SELECT SUM(price * quantity)
            FROM order_items
            WHERE order_id = p_order_id);
END;
$$ LANGUAGE plpgsql;
```

---

## Troubleshooting

**Check current version**:
```bash
make migrate-version            # Tables
make migrate-seeds-version      # Seeds
make migrate-functions-version  # Functions
```

**Force specific version** (use with caution):
```bash
migrate -path database/migrations/functions/ \
        -database "postgresql://..." \
        force 1
```

**Rollback one step**:
```bash
make migrate-down              # Last table migration
make migrate-down-functions    # Last function migration
```

---

**Last Updated**: 2025-11-09
**golang-migrate Version**: Compatible with v4.x
