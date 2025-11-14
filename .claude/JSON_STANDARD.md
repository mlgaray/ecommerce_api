# JSON Standard

## Rule: ALWAYS use `jsonb`, NEVER use `json`

PostgreSQL provides two JSON types: `json` (text-based) and `jsonb` (binary-optimized).

**Use `jsonb` exclusively** in all new code. The `json` type exists only for backward compatibility with PostgreSQL versions before 9.4 (2014).

---

## ✅ Correct Usage (JSONB)

### SQL Queries

```sql
-- Use jsonb_agg() for aggregating rows into JSON arrays
SELECT jsonb_agg(
    jsonb_build_object(
        'id', product.id,
        'name', product.name
    )
) FROM products;

-- Use jsonb_build_object() for constructing JSON objects
SELECT jsonb_build_object(
    'user_id', users.id,
    'email', users.email
);

-- Cast empty arrays to jsonb
COALESCE(subquery, '[]'::jsonb)

-- Cast empty objects to jsonb
COALESCE(subquery, '{}'::jsonb)
```

### Column Definitions

```sql
CREATE TABLE products (
    id SERIAL PRIMARY KEY,
    metadata jsonb,              -- ✅ Use jsonb
    configuration jsonb NOT NULL -- ✅ Use jsonb
);
```

### Stored Functions

```sql
CREATE OR REPLACE FUNCTION get_user_data(p_user_id INT)
RETURNS jsonb AS $$              -- ✅ Return jsonb
BEGIN
    RETURN jsonb_build_object(   -- ✅ Use jsonb functions
        'id', user_id,
        'data', metadata
    );
END;
$$ LANGUAGE plpgsql;
```

---

## ❌ Incorrect Usage (JSON)

```sql
-- ❌ DON'T use json_agg()
SELECT json_agg(...) FROM products;

-- ❌ DON'T use json_build_object()
SELECT json_build_object(...);

-- ❌ DON'T cast to json
COALESCE(subquery, '[]'::json)

-- ❌ DON'T use json column type
CREATE TABLE products (
    metadata json  -- ❌ Use jsonb instead
);
```

---

## 📊 Why JSONB?

| Feature | `json` (text) | `jsonb` (binary) |
|---------|---------------|------------------|
| **Storage** | Plain text | Binary compressed |
| **Read speed** | Slow | **Fast** ⚡ |
| **Write speed** | Fast | Slightly slower (compression overhead) |
| **Disk space** | Larger | **Smaller** 💾 |
| **Indexing** | ❌ Not supported | ✅ **GIN indexes** 🚀 |
| **Ordering** | Preserves key order | Reorders keys |
| **Operations** | Limited | **Rich operators** (`->`, `->>`, `@>`, etc.) |
| **Modification** | Replace entire value | **Update specific fields** |

---

## 🔧 Go Code (No Changes Needed)

**Good news**: Go code remains unchanged when using `jsonb` instead of `json`.

```go
// PostgreSQL returns jsonb → Go receives []byte → Unmarshal to struct
var data []byte
err := db.QueryRow("SELECT metadata FROM products WHERE id = $1", id).Scan(&data)

// Unmarshal works the same
var product Product
json.Unmarshal(data, &product)

// Marshal and insert
data, _ := json.Marshal(product)
db.Exec("INSERT INTO products (metadata) VALUES ($1)", data)
```

The PostgreSQL driver (`lib/pq`) handles the conversion automatically.

---

## 🎯 Migration Strategy

When refactoring existing code:

1. **Find all occurrences**:
   ```bash
   grep -r "json_agg" internal/
   grep -r "json_build_object" internal/
   grep -r "'\\[\\]'::json" internal/
   ```

2. **Replace**:
   - `json_agg` → `jsonb_agg`
   - `json_build_object` → `jsonb_build_object`
   - `'[]'::json` → `'[]'::jsonb`
   - `'{}'::json` → `'{}'::jsonb`

3. **Test**: Run unit and integration tests to verify no regressions

---

## 📚 References

- [PostgreSQL JSON Types](https://www.postgresql.org/docs/current/datatype-json.html)
- [PostgreSQL JSON Functions](https://www.postgresql.org/docs/current/functions-json.html)
- [JSONB Performance](https://www.postgresql.org/docs/current/datatype-json.html#JSON-INDEXING)

---

## ✅ Checklist for New Code

Before committing SQL code, verify:

- [ ] Uses `jsonb_agg()` instead of `json_agg()`
- [ ] Uses `jsonb_build_object()` instead of `json_build_object()`
- [ ] Casts empty values to `jsonb`: `'[]'::jsonb`, `'{}'::jsonb`
- [ ] Column types use `jsonb` instead of `json`
- [ ] Function returns use `RETURNS jsonb` instead of `RETURNS json`

---

**Last Updated**: 2025-11-09
**Status**: Active standard for all PostgreSQL queries
