# TODO: Search and Filters System

## Objetivo

Implementar un sistema flexible de búsqueda y filtros para productos que sea:
- RESTful y estándar industria
- Escalable (agregar filtros sin código nuevo)
- Performante con índices estratégicos
- Client-friendly (query params combinables)

---

## Prioridad

**ALTA** - GetAll es crítico porque:
- Alto tráfico (homepage, catálogo)
- Involucra muchas filas
- Mayor impacto en experiencia de usuario

**GetByID**: NO prioritario (ya es rápido, 1 producto)

---

## Arquitectura Recomendada

### ✅ Endpoint único flexible (Opción recomendada)

```
GET /shops/:shop_id/products?
    search=laptop              # Full-text search
    &category_id=5             # Filtrar por categoría
    &is_active=true            # Solo activos
    &is_highlighted=true       # Destacados
    &is_promotional=true       # En promoción
    &min_price=100             # Rango precio
    &max_price=1000
    &sort=price                # Ordenar: price, name, created_at
    &order=asc                 # asc/desc
    &limit=20                  # Paginación
    &cursor=123
```

**Ventajas:**
- RESTful y estándar (Amazon, MercadoLibre, Stripe, etc.)
- Flexible - combinar filtros sin código nuevo
- Un endpoint, un handler, un servicio
- Client-friendly

### ❌ NO hacer: Endpoint por filtro

```
GET /products/active           ❌
GET /products/by-category/:id  ❌
GET /products/search?q=laptop  ❌
```

**Problemas:**
- Código duplicado
- No escalable
- No se pueden combinar filtros
- Explosión de rutas

---

## Implementación

### 1. Estructura de Filtros

```go
// internal/core/models/product_filters.go
package models

type ProductFilters struct {
    ShopID          int
    Search          *string   // nil = no aplicar
    CategoryID      *int
    IsActive        *bool
    IsHighlighted   *bool
    IsPromotional   *bool
    MinPrice        *float64
    MaxPrice        *float64
    SortBy          string    // "price", "name", "created_at"
    SortOrder       string    // "asc", "desc"
    Limit           int
    Cursor          int
}

// Validaciones y defaults
func (f *ProductFilters) Validate() error {
    if f.ShopID <= 0 {
        return errors.New("shop_id is required")
    }

    if f.Limit <= 0 || f.Limit > 100 {
        f.Limit = 20  // Default
    }

    validSorts := map[string]bool{"price": true, "name": true, "created_at": true}
    if f.SortBy == "" {
        f.SortBy = "created_at"  // Default
    } else if !validSorts[f.SortBy] {
        return errors.New("invalid sort field")
    }

    if f.SortOrder != "asc" && f.SortOrder != "desc" {
        f.SortOrder = "desc"  // Default
    }

    return nil
}
```

### 2. Query Builder en Repository

```go
// internal/infraestructure/adapters/repositories/postgresql/product_repository.go

func (r *ProductRepository) GetAllWithFilters(ctx context.Context, filters models.ProductFilters) ([]*models.Product, error) {
    // Base query con JSONB aggregation
    baseQuery := `
        SELECT
            p.id, p.name, p.description, p.price, p.stock, p.minimum_stock,
            p.is_active, p.is_highlighted, p.is_promotional, p.promotional_price,
            p.category_id, p.shop_id, p.created_at,

            -- Aggregate images
            COALESCE(
                jsonb_agg(DISTINCT jsonb_build_object(
                    'id', pi.id,
                    'url', pi.url
                )) FILTER (WHERE pi.id IS NOT NULL),
                '[]'::jsonb
            ) as images,

            -- Aggregate variants with options
            COALESCE(
                jsonb_agg(DISTINCT jsonb_build_object(
                    'id', pv.id,
                    'name', pv.name,
                    'order', pv."order",
                    'selection_type', pv.selection_type,
                    'max_selections', pv.max_selections,
                    'options', (
                        SELECT COALESCE(jsonb_agg(jsonb_build_object(
                            'id', vo.id,
                            'name', vo.name,
                            'price', vo.price,
                            'order', vo."order"
                        )), '[]'::jsonb)
                        FROM variant_options vo
                        WHERE vo.variant_id = pv.id
                    )
                )) FILTER (WHERE pv.id IS NOT NULL),
                '[]'::jsonb
            ) as variants

        FROM products p
        LEFT JOIN product_images pi ON p.id = pi.product_id
        LEFT JOIN product_variants pv ON p.id = pv.product_id
        WHERE p.shop_id = $1
    `

    // Build dynamic WHERE conditions
    conditions := []string{}
    args := []interface{}{filters.ShopID}
    argPos := 2

    if filters.CategoryID != nil {
        conditions = append(conditions, fmt.Sprintf("p.category_id = $%d", argPos))
        args = append(args, *filters.CategoryID)
        argPos++
    }

    if filters.IsActive != nil {
        conditions = append(conditions, fmt.Sprintf("p.is_active = $%d", argPos))
        args = append(args, *filters.IsActive)
        argPos++
    }

    if filters.IsHighlighted != nil {
        conditions = append(conditions, fmt.Sprintf("p.is_highlighted = $%d", argPos))
        args = append(args, *filters.IsHighlighted)
        argPos++
    }

    if filters.IsPromotional != nil {
        conditions = append(conditions, fmt.Sprintf("p.is_promotional = $%d", argPos))
        args = append(args, *filters.IsPromotional)
        argPos++
    }

    if filters.MinPrice != nil {
        conditions = append(conditions, fmt.Sprintf("p.price >= $%d", argPos))
        args = append(args, *filters.MinPrice)
        argPos++
    }

    if filters.MaxPrice != nil {
        conditions = append(conditions, fmt.Sprintf("p.price <= $%d", argPos))
        args = append(args, *filters.MaxPrice)
        argPos++
    }

    if filters.Search != nil && *filters.Search != "" {
        // Full-text search (usar índice GIN)
        conditions = append(conditions, fmt.Sprintf(
            "to_tsvector('spanish', p.name || ' ' || COALESCE(p.description, '')) @@ plainto_tsquery('spanish', $%d)",
            argPos,
        ))
        args = append(args, *filters.Search)
        argPos++
    }

    // Cursor-based pagination
    if filters.Cursor > 0 {
        conditions = append(conditions, fmt.Sprintf("p.id > $%d", argPos))
        args = append(args, filters.Cursor)
        argPos++
    }

    // Append all conditions
    if len(conditions) > 0 {
        baseQuery += " AND " + strings.Join(conditions, " AND ")
    }

    // GROUP BY (necesario por aggregations)
    baseQuery += " GROUP BY p.id"

    // ORDER BY dinámico (sanitizado en Validate())
    baseQuery += fmt.Sprintf(" ORDER BY p.%s %s", filters.SortBy, filters.SortOrder)

    // LIMIT
    baseQuery += fmt.Sprintf(" LIMIT $%d", argPos)
    args = append(args, filters.Limit)

    // Execute query
    rows, err := r.db.QueryContext(ctx, baseQuery, args...)
    if err != nil {
        logs.WithFields(map[string]interface{}{
            "file":     ProductRepositoryField,
            "function": "get_all_with_filters",
            "error":    err.Error(),
        }).Error("Failed to query products with filters")
        return nil, fmt.Errorf("database operation failed")
    }
    defer rows.Close()

    // Scan results (similar a GetAllByShopID actual)
    var products []*models.Product
    for rows.Next() {
        var product models.Product
        var imagesJSON, variantsJSON []byte

        err := rows.Scan(
            &product.ID, &product.Name, &product.Description,
            &product.Price, &product.Stock, &product.MinimumStock,
            &product.IsActive, &product.IsHighlighted, &product.IsPromotional,
            &product.PromotionalPrice, &product.Category.ID, &product.ShopID,
            &product.CreatedAt, &imagesJSON, &variantsJSON,
        )

        if err != nil {
            return nil, fmt.Errorf("failed to scan product: %w", err)
        }

        // Unmarshal JSONB
        if err := json.Unmarshal(imagesJSON, &product.Images); err != nil {
            return nil, fmt.Errorf("failed to unmarshal images: %w", err)
        }

        if err := json.Unmarshal(variantsJSON, &product.Variants); err != nil {
            return nil, fmt.Errorf("failed to unmarshal variants: %w", err)
        }

        products = append(products, &product)
    }

    return products, nil
}
```

### 3. Handler con Query Params

```go
// internal/infraestructure/adapters/http/product_handler.go

func (h *ProductHandler) GetAll(w http.ResponseWriter, r *http.Request) {
    shopID, _ := strconv.Atoi(mux.Vars(r)["shop_id"])

    // Parse query params
    filters := models.ProductFilters{
        ShopID: shopID,
    }

    // Optional filters
    if search := r.URL.Query().Get("search"); search != "" {
        filters.Search = &search
    }

    if catID := r.URL.Query().Get("category_id"); catID != "" {
        id, _ := strconv.Atoi(catID)
        filters.CategoryID = &id
    }

    if isActive := r.URL.Query().Get("is_active"); isActive != "" {
        active := isActive == "true"
        filters.IsActive = &active
    }

    if isHighlighted := r.URL.Query().Get("is_highlighted"); isHighlighted != "" {
        highlighted := isHighlighted == "true"
        filters.IsHighlighted = &highlighted
    }

    if isPromotional := r.URL.Query().Get("is_promotional"); isPromotional != "" {
        promotional := isPromotional == "true"
        filters.IsPromotional = &promotional
    }

    if minPrice := r.URL.Query().Get("min_price"); minPrice != "" {
        price, _ := strconv.ParseFloat(minPrice, 64)
        filters.MinPrice = &price
    }

    if maxPrice := r.URL.Query().Get("max_price"); maxPrice != "" {
        price, _ := strconv.ParseFloat(maxPrice, 64)
        filters.MaxPrice = &price
    }

    filters.SortBy = r.URL.Query().Get("sort")
    filters.SortOrder = r.URL.Query().Get("order")
    filters.Limit, _ = strconv.Atoi(r.URL.Query().Get("limit"))
    filters.Cursor, _ = strconv.Atoi(r.URL.Query().Get("cursor"))

    // Validate
    if err := filters.Validate(); err != nil {
        // Return 400 Bad Request
        return
    }

    // Call service
    products, nextCursor, hasMore, err := h.productService.GetAllWithFilters(ctx, filters)
    // ...
}
```

### 4. Interfaces actualizadas

```go
// internal/core/ports/product_repository.go
type ProductRepository interface {
    Create(...)
    CreateWithStoredProcedure(...)
    GetAllByShopID(...)
    GetAllWithFilters(ctx context.Context, filters models.ProductFilters) ([]*models.Product, error)
    GetByID(...)
    Update(...)
    UpdateWithStoredProcedure(...)
}

// internal/core/ports/product_service.go
type ProductService interface {
    Create(...)
    CreateWithStoredProcedure(...)
    GetAllByShopID(...)
    GetAllWithFilters(ctx context.Context, filters models.ProductFilters) ([]*models.Product, int, bool, error)
    GetByID(...)
    Update(...)
    UpdateWithStoredProcedure(...)
}
```

---

## Optimizaciones de Performance

### Índices estratégicos (CRÍTICO)

```sql
-- Índice compuesto para filtros comunes
CREATE INDEX idx_products_shop_active ON products(shop_id, is_active);

-- Índice por categoría
CREATE INDEX idx_products_category ON products(category_id);

-- Índice para rango de precios
CREATE INDEX idx_products_price ON products(price);

-- Índice para highlighted/promotional
CREATE INDEX idx_products_flags ON products(shop_id, is_highlighted, is_promotional)
WHERE is_active = true;

-- Índice GIN para full-text search (español)
CREATE INDEX idx_products_search_spanish
ON products
USING GIN (to_tsvector('spanish', name || ' ' || COALESCE(description, '')));

-- Índice para ordenamiento por fecha
CREATE INDEX idx_products_created_at ON products(created_at DESC);
```

### ¿Por qué NO usar Stored Procedure para GetAll?

❌ **Read-heavy, no transacciones** (SP brillan en writes)
❌ **Queries dinámicos** (WHERE clauses variables → más fácil en Go)
❌ **Full-text search** (Postgres tiene tsquery/GIN nativos)
❌ **Caching posible** (reads se cachean, writes no)

**Mejor approach:**
✅ Query builder dinámico en Go
✅ Índices estratégicos
✅ JSONB aggregation (ya lo tenemos)
✅ Cursor-based pagination (ya lo tenemos)

### Caching (opcional, fase posterior)

```go
// Redis cache para queries frecuentes
cacheKey := fmt.Sprintf("products:shop:%d:filters:%s", shopID, filtersHash)
```

---

## Casos de Uso - Ejemplos

### Homepage - Productos activos destacados
```
GET /shops/1/products?is_active=true&is_highlighted=true&limit=10&sort=created_at&order=desc
```

### Búsqueda de usuario
```
GET /shops/1/products?search=laptop&is_active=true&sort=price&order=asc
```

### Filtros combinados (catálogo con sidebar)
```
GET /shops/1/products?category_id=5&min_price=100&max_price=500&is_active=true&sort=price&order=asc&limit=20
```

### Página de promociones
```
GET /shops/1/products?is_promotional=true&is_active=true&sort=promotional_price&order=asc
```

### Paginación
```
GET /shops/1/products?is_active=true&limit=20&cursor=0       # Primera página
GET /shops/1/products?is_active=true&limit=20&cursor=123     # Siguiente página
```

---

## Testing

### Unit tests - Query Builder

```go
func TestBuildFilterQuery(t *testing.T) {
    tests := []struct {
        name     string
        filters  models.ProductFilters
        wantSQL  string
        wantArgs []interface{}
    }{
        {
            name: "Solo shop_id",
            filters: models.ProductFilters{ShopID: 1, Limit: 20},
            // Validar query generado
        },
        {
            name: "Con búsqueda",
            filters: models.ProductFilters{
                ShopID: 1,
                Search: stringPtr("laptop"),
                Limit: 20,
            },
            // Validar query con full-text search
        },
        // ... más casos
    }
}
```

### Integration tests

```go
func TestGetAllWithFilters_Integration(t *testing.T) {
    // Setup: Crear productos de prueba con diferentes atributos

    // Test: Filtrar por categoría
    products, err := repo.GetAllWithFilters(ctx, models.ProductFilters{
        ShopID:     1,
        CategoryID: intPtr(5),
    })
    assert.NoError(t, err)
    for _, p := range products {
        assert.Equal(t, 5, p.Category.ID)
    }

    // Test: Búsqueda full-text
    // Test: Rango de precios
    // Test: Combinación de filtros
}
```

---

## Migración desde GetAllByShopID

### Opción A: Deprecar gradualmente (RECOMENDADO)

```go
// Mantener endpoint viejo por compatibilidad
GET /shops/:shop_id/products  → GetAllByShopID (sin filtros)

// Nuevo endpoint con filtros
GET /shops/:shop_id/products?...  → GetAllWithFilters
```

**Implementación:**
```go
func (h *ProductHandler) GetAll(w http.ResponseWriter, r *http.Request) {
    // Si tiene query params, usar nueva implementación
    if len(r.URL.Query()) > 0 {
        h.GetAllWithFilters(w, r)
        return
    }

    // Sin query params, usar implementación vieja (backward compatible)
    h.GetAllByShopIDLegacy(w, r)
}
```

### Opción B: Reemplazar completamente

Cambiar `GetAllByShopID` a `GetAllWithFilters` directamente.

---

## Plan de Implementación

### Sprint 1: Base (1-2 días)
- [ ] Crear `ProductFilters` struct
- [ ] Implementar query builder dinámico
- [ ] Migrar `GetAllByShopID` a `GetAllWithFilters`
- [ ] Tests unitarios

### Sprint 2: Filtros básicos (1 día)
- [ ] Implementar: `is_active`, `category_id`, `is_highlighted`
- [ ] Handler con query params
- [ ] Tests integración

### Sprint 3: Búsqueda (1-2 días)
- [ ] Crear índice GIN para full-text
- [ ] Implementar search con `to_tsvector`
- [ ] Tests con casos reales

### Sprint 4: Filtros avanzados (1 día)
- [ ] Rango de precios (`min_price`, `max_price`)
- [ ] Ordenamiento dinámico (`sort`, `order`)
- [ ] Más tests

### Sprint 5: Performance (1 día)
- [ ] Crear índices estratégicos
- [ ] EXPLAIN ANALYZE de queries
- [ ] Benchmarks

---

## Referencias

- **Amazon**: `?k=laptop&rh=n:123&s=price-asc-rank`
- **MercadoLibre**: `?q=laptop&category=123&sort=price&order=asc`
- **Stripe API**: Query params para filtros (industry standard)
- **PostgreSQL Full-Text Search**: https://www.postgresql.org/docs/current/textsearch.html

---

## Notas

- Priorizar índices compuestos por uso real (monitorear queries lentas)
- Considerar caching solo si hay queries muy frecuentes (Redis)
- GetByID probablemente NO necesita optimización (ya rápido)
- Stored procedures NO benefician reads con filtros dinámicos
