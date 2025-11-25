package postgresql

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"time"

	jsoniter "github.com/json-iterator/go"
	"github.com/lib/pq"

	"github.com/mlgaray/ecommerce_api/internal/core/errors"
	"github.com/mlgaray/ecommerce_api/internal/core/models"
	"github.com/mlgaray/ecommerce_api/internal/infraestructure/adapters/logs"
)

var json = jsoniter.ConfigCompatibleWithStandardLibrary

type ProductRepository struct {
	db *sql.DB
}

// Product repository log field constants
const (
	ProductRepositoryField          = "product_repository"
	ProductCreateFunctionField      = "create"
	ProductGetByIDFunctionField     = "get_by_id"
	ProductUpdateFunctionField      = "update"
	ProductUnmarshallSubFuncField   = "unmarshall"
	MarshalVariantsSubFuncField     = "marshal_variants"
	MarshalImagesSubFuncField       = "marshal_images"
	CallStoredProcedureSubFuncField = "call_stored_procedure"
)

// Product repository log message constants
const (
	LogFailedInsertProduct       = "Failed to insert product"
	LogFailedInsertProductImage  = "Failed to insert product image"
	LogFailedInsertVariant       = "Failed to insert variant"
	LogFailedInsertVariantOption = "Failed to insert variant option"
	LogFailedCreateProductSP     = "Failed to create product via stored procedure"
	LogFailedUpdateProductSP     = "Failed to update product via stored procedure"
	LogFailedMarshalVariants     = "Failed to marshal variants for stored procedure"
	LogFailedMarshalImages       = "Failed to marshal images for stored procedure"
	failedReadProductsByShop     = "Failed to read products by shop"
	failedReadProductByID        = "Failed to read product by ID"
	productNotFoundMessage       = "Product not found"
)

func NewProductRepository(dataBaseConnection DataBaseConnection) *ProductRepository {
	return &ProductRepository{
		db: dataBaseConnection.Connect(),
	}
}

func (r *ProductRepository) GetByID(ctx context.Context, productID int) (*models.Product, error) {
	query := `
		SELECT
			p.id, p.name, p.description, p.price, p.stock, COALESCE(p.minimum_stock, 0),
			p.is_active, p.is_highlighted, p.is_promotional, COALESCE(p.promotional_price, 0),
			p.created_at,
			c.id, c.name, COALESCE(c.description, ''),
			COALESCE(
				(SELECT jsonb_agg(
					jsonb_build_object(
						'id', pi2.id,
						'url', pi2.url
					) ORDER BY pi2.id
				)
				FROM product_images pi2
				WHERE pi2.product_id = p.id),
				'[]'::jsonb
			) AS images,
			COALESCE(
				(SELECT jsonb_agg(
					jsonb_build_object(
						'id', pv2.id,
						'name', pv2.name,
						'order', pv2."order",
						'selection_type', pv2.selection_type,
						'max_selections', pv2.max_selections,
						'options', (
							SELECT COALESCE(jsonb_agg(
								jsonb_build_object(
									'id', vo.id,
									'name', vo.name,
									'price', vo.price,
									'order', vo."order"
								) ORDER BY vo."order"
							), '[]'::jsonb)
							FROM variant_options vo
							WHERE vo.variant_id = pv2.id
						)
					) ORDER BY pv2."order"
				)
				FROM product_variants pv2
				WHERE pv2.product_id = p.id),
				'[]'::jsonb
			) AS variants
		FROM products p
		INNER JOIN categories c ON p.category_id = c.id
		WHERE p.id = $1`

	product := &models.Product{
		Category: &models.Category{},
	}

	var imagesJSON, variantsJSON []byte

	err := r.db.QueryRowContext(ctx, query, productID).Scan(
		&product.ID,
		&product.Name,
		&product.Description,
		&product.Price,
		&product.Stock,
		&product.MinimumStock,
		&product.IsActive,
		&product.IsHighlighted,
		&product.IsPromotional,
		&product.PromotionalPrice,
		&product.CreatedAt,
		&product.Category.ID,
		&product.Category.Name,
		&product.Category.Description,
		&imagesJSON,
		&variantsJSON,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			logs.WithFields(map[string]interface{}{
				"file":       ProductRepositoryField,
				"function":   ProductGetByIDFunctionField,
				"product_id": productID,
			}).Warn(productNotFoundMessage)
			return nil, &errors.RecordNotFoundError{Message: errors.ProductNotFound}
		}

		logs.WithFields(map[string]interface{}{
			"file":       ProductRepositoryField,
			"function":   ProductGetByIDFunctionField,
			"sub_func":   ScanField,
			"product_id": productID,
			"error":      err.Error(),
		}).Error(failedReadProductByID)
		return nil, fmt.Errorf("database operation failed")
	}

	// Parse images JSON
	if err := json.Unmarshal(imagesJSON, &product.Images); err != nil {
		logs.WithFields(map[string]interface{}{
			"file":       ProductRepositoryField,
			"function":   ProductGetByIDFunctionField,
			"sub_func":   UnmarshallField,
			"product_id": product.ID,
			"error":      err.Error(),
		}).Error("Failed to unmarshal product images")
		return nil, fmt.Errorf("database operation failed")
	}

	// Parse variants JSON
	if err := json.Unmarshal(variantsJSON, &product.Variants); err != nil {
		logs.WithFields(map[string]interface{}{
			"file":       ProductRepositoryField,
			"function":   ProductGetByIDFunctionField,
			"sub_func":   UnmarshallField,
			"product_id": product.ID,
			"error":      err.Error(),
		}).Error("Failed to unmarshal product variants")
		return nil, fmt.Errorf("database operation failed")
	}

	return product, nil
}

func (r *ProductRepository) Create(ctx context.Context, product *models.Product, shopID int) (*models.Product, error) {
	startTime := time.Now()

	// 1. Prepare image URLs array
	imageURLs := make([]string, len(product.Images))
	for i, img := range product.Images {
		imageURLs[i] = img.URL
	}

	// 2. Serialize variants to JSON
	variantsJSON, err := json.Marshal(product.Variants)
	if err != nil {
		logs.WithFields(map[string]interface{}{
			"file":     ProductRepositoryField,
			"function": ProductCreateFunctionField,
			"sub_func": MarshalVariantsSubFuncField,
			"error":    err.Error(),
		}).Error(LogFailedMarshalVariants)
		return nil, fmt.Errorf("failed to prepare variants: %w", err)
	}

	// 3. Call stored procedure (single query - all inserts happen in DB)
	var productID int
	queryStart := time.Now()
	err = r.db.QueryRowContext(ctx, `
		SELECT create_product(
			$1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13
		)`,
		product.Name,
		product.Description,
		product.Price,
		product.Stock,
		product.MinimumStock,
		product.IsActive,
		product.IsHighlighted,
		product.IsPromotional,
		product.PromotionalPrice,
		product.Category.ID,
		shopID,
		pq.Array(imageURLs),
		variantsJSON,
	).Scan(&productID)

	if err != nil {
		logs.WithFields(map[string]interface{}{
			"file":         ProductRepositoryField,
			"function":     ProductCreateFunctionField,
			"sub_func":     CallStoredProcedureSubFuncField,
			"product_name": product.Name,
			"shop_id":      shopID,
			"error":        err.Error(),
		}).Error(LogFailedCreateProductSP)

		// Check if it's a PostgreSQL error from the stored procedure
		if pqErr, ok := err.(*pq.Error); ok {
			// RAISE EXCEPTION from stored procedure comes as pq.Error
			// Extract meaningful error message for better debugging
			logs.WithFields(map[string]interface{}{
				"file":       ProductRepositoryField,
				"function":   ProductCreateFunctionField,
				"pg_code":    pqErr.Code,    // PostgreSQL error code
				"pg_message": pqErr.Message, // Error message from RAISE EXCEPTION
				"pg_detail":  pqErr.Detail,  // Additional detail if any
				"pg_hint":    pqErr.Hint,    // Hint if provided
			}).Debug("PostgreSQL error details from stored procedure")

			// Return error with SP context (preserves original message)
			return nil, fmt.Errorf("stored procedure error: %s", pqErr.Message)
		}

		// Not a PostgreSQL error (network, context cancelled, etc.)
		return nil, fmt.Errorf("database operation failed: %w", err)
	}

	logs.WithFields(map[string]interface{}{
		"file":        ProductRepositoryField,
		"function":    ProductCreateFunctionField,
		"sub_func":    CallStoredProcedureSubFuncField,
		"duration_ms": time.Since(queryStart).Milliseconds(),
	}).Debug("Stored procedure executed successfully")

	// 4. Set product ID
	product.ID = productID

	logs.WithFields(map[string]interface{}{
		"file":              ProductRepositoryField,
		"function":          ProductCreateFunctionField,
		"product_id":        productID,
		"total_duration_ms": time.Since(startTime).Milliseconds(),
	}).Info("Product creation completed (stored procedure)")

	return product, nil
}

func (r *ProductRepository) GetAllByShopIDWithFilters(ctx context.Context, filters models.ProductFilters) ([]*models.Product, error) {
	startTime := time.Now()

	// Lightweight query - NO variants/options (for real-time search performance)
	// Variants come as empty array [] for listing views (full details in GetByID)
	baseQuery := `
		SELECT
			p.id, p.name, p.description, p.price, p.stock, COALESCE(p.minimum_stock, 0),
			p.is_active, p.is_highlighted, p.is_promotional, COALESCE(p.promotional_price, 0),
			p.created_at,
			c.id, c.name, COALESCE(c.description, ''),
			COALESCE(
				(SELECT jsonb_agg(
					jsonb_build_object(
						'id', pi2.id,
						'url', pi2.url
					) ORDER BY pi2.id
				)
				FROM product_images pi2
				WHERE pi2.product_id = p.id),
				'[]'::jsonb
			) AS images
		FROM products p
		INNER JOIN categories c ON p.category_id = c.id
		WHERE p.shop_id = $1`

	// Build dynamic WHERE conditions
	conditions := []string{}
	args := []interface{}{filters.ShopID}
	argPos := 2

	// Category filter
	if filters.CategoryID != nil {
		conditions = append(conditions, fmt.Sprintf("p.category_id = $%d", argPos))
		args = append(args, *filters.CategoryID)
		argPos++
	}

	// Active status filter
	if filters.IsActive != nil {
		conditions = append(conditions, fmt.Sprintf("p.is_active = $%d", argPos))
		args = append(args, *filters.IsActive)
		argPos++
	}

	// Highlighted filter
	if filters.IsHighlighted != nil {
		conditions = append(conditions, fmt.Sprintf("p.is_highlighted = $%d", argPos))
		args = append(args, *filters.IsHighlighted)
		argPos++
	}

	// Promotional filter
	if filters.IsPromotional != nil {
		conditions = append(conditions, fmt.Sprintf("p.is_promotional = $%d", argPos))
		args = append(args, *filters.IsPromotional)
		argPos++
	}

	// Minimum price filter
	if filters.MinPrice != nil {
		conditions = append(conditions, fmt.Sprintf("p.price >= $%d", argPos))
		args = append(args, *filters.MinPrice)
		argPos++
	}

	// Maximum price filter
	if filters.MaxPrice != nil {
		conditions = append(conditions, fmt.Sprintf("p.price <= $%d", argPos))
		args = append(args, *filters.MaxPrice)
		argPos++
	}

	// Search: High-performance approach using trigram indexes (pg_trgm)
	// Requires migration 000012_add_trigram_indexes to be applied
	//
	// Strategy:
	// 1. ILIKE with trigram GIN index (idx_products_name_trgm) on name field only
	//    - Fast partial matches: "4" finds "Product 4", "Product 400"
	//    - Performance: O(log n) with index (not O(n) like standard ILIKE)
	//    - Name-only search provides better precision and fewer false positives
	//
	// Performance on 500k products:
	// - ILIKE without index: ~5000ms (full scan)
	// - ILIKE with pg_trgm (name only): ~30-40ms (index scan, fewer results)
	if filters.Search != nil && *filters.Search != "" {
		searchTerm := *filters.Search

		// ILIKE with trigram index: Fast partial matching
		// Searches only in name (more precise, better UX for real-time search)
		conditions = append(conditions, fmt.Sprintf(
			"p.name ILIKE $%d",
			argPos,
		))
		args = append(args, "%"+searchTerm+"%")
		argPos++
	}

	// Keyset pagination: use LastID and LastSortValue for cursor-based pagination
	// These values come from the decoded cursor (decoded in Use Case layer)
	if filters.LastID != nil {
		// Build keyset condition: (sort_field, id) comparison
		// This ensures stable pagination even when sort_field has duplicate values
		sortField := fmt.Sprintf("p.%s", filters.SortBy)

		if filters.SortOrder == "desc" {
			// DESC: (sort_field < last_value) OR (sort_field = last_value AND id < last_id)
			if filters.LastSortValue != nil {
				conditions = append(conditions, fmt.Sprintf(
					"(%s < $%d OR (%s = $%d AND p.id < $%d))",
					sortField, argPos, sortField, argPos, argPos+1,
				))
				args = append(args, filters.LastSortValue, *filters.LastID)
				argPos += 2
			} else {
				// Sorting by ID only
				conditions = append(conditions, fmt.Sprintf("p.id < $%d", argPos))
				args = append(args, *filters.LastID)
				argPos++
			}
		} else {
			// ASC: (sort_field > last_value) OR (sort_field = last_value AND id > last_id)
			if filters.LastSortValue != nil {
				conditions = append(conditions, fmt.Sprintf(
					"(%s > $%d OR (%s = $%d AND p.id > $%d))",
					sortField, argPos, sortField, argPos, argPos+1,
				))
				args = append(args, filters.LastSortValue, *filters.LastID)
				argPos += 2
			} else {
				// Sorting by ID only
				conditions = append(conditions, fmt.Sprintf("p.id > $%d", argPos))
				args = append(args, *filters.LastID)
				argPos++
			}
		}
	}

	// Append all WHERE conditions
	if len(conditions) > 0 {
		baseQuery += " AND " + strings.Join(conditions, " AND ")
	}

	// ORDER BY with validated and sanitized fields
	// Always include id as secondary sort for stable pagination
	baseQuery += fmt.Sprintf(" ORDER BY p.%s %s, p.id %s",
		filters.SortBy,
		strings.ToUpper(filters.SortOrder),
		strings.ToUpper(filters.SortOrder))

	// LIMIT - solicitar limit + 1 para detectar si hay más páginas
	baseQuery += fmt.Sprintf(" LIMIT $%d", argPos)
	args = append(args, filters.Limit+1)

	// Execute query
	logs.WithFields(map[string]interface{}{
		"file":          ProductRepositoryField,
		"function":      "get_all_by_shop_id_with_filters",
		"shop_id":       filters.ShopID,
		"sort_by":       filters.SortBy,
		"sort_order":    filters.SortOrder,
		"has_search":    filters.Search != nil,
		"has_category":  filters.CategoryID != nil,
		"has_price_min": filters.MinPrice != nil,
		"has_price_max": filters.MaxPrice != nil,
		"limit":         filters.Limit,
		"last_id":       filters.LastID,
	}).Debug("Executing lightweight product search query (no variants)")

	rows, err := r.db.QueryContext(ctx, baseQuery, args...)
	if err != nil {
		logs.WithFields(map[string]interface{}{
			"file":        ProductRepositoryField,
			"function":    "get_all_by_shop_id_with_filters",
			"duration_ms": time.Since(startTime).Milliseconds(),
			"error":       err.Error(),
		}).Error("Failed to query products with filters")
		return nil, fmt.Errorf("database operation failed")
	}
	defer rows.Close()

	// Scan results (lightweight - no variants for performance)
	products := make([]*models.Product, 0, filters.Limit)

	for rows.Next() {
		product := &models.Product{
			Category: &models.Category{},
			Variants: []*models.Variant{}, // Empty array for listing view
		}

		var imagesJSON []byte

		err := rows.Scan(
			&product.ID,
			&product.Name,
			&product.Description,
			&product.Price,
			&product.Stock,
			&product.MinimumStock,
			&product.IsActive,
			&product.IsHighlighted,
			&product.IsPromotional,
			&product.PromotionalPrice,
			&product.CreatedAt,
			&product.Category.ID,
			&product.Category.Name,
			&product.Category.Description,
			&imagesJSON,
		)

		if err != nil {
			logs.WithFields(map[string]interface{}{
				"file":     ProductRepositoryField,
				"function": "get_all_by_shop_id_with_filters",
				"sub_func": ScanField,
				"error":    err.Error(),
			}).Error("Failed to scan product row")
			return nil, fmt.Errorf("database operation failed")
		}

		// Parse images JSON
		if err := json.Unmarshal(imagesJSON, &product.Images); err != nil {
			logs.WithFields(map[string]interface{}{
				"file":       ProductRepositoryField,
				"function":   "get_all_by_shop_id_with_filters",
				"sub_func":   UnmarshallField,
				"product_id": product.ID,
				"error":      err.Error(),
			}).Error("Failed to unmarshal product images")
			return nil, fmt.Errorf("database operation failed")
		}

		// Note: Variants intentionally omitted for performance
		// Use GetByID for full product details including variants

		products = append(products, product)
	}

	if err := rows.Err(); err != nil {
		logs.WithFields(map[string]interface{}{
			"file":     ProductRepositoryField,
			"function": "get_all_by_shop_id_with_filters",
			"sub_func": NextField,
			"error":    err.Error(),
		}).Error("Error iterating product rows")
		return nil, fmt.Errorf("database operation failed")
	}

	queryDuration := time.Since(startTime).Milliseconds()

	// Performance monitoring
	logs.WithFields(map[string]interface{}{
		"file":         ProductRepositoryField,
		"function":     "get_all_by_shop_id_with_filters",
		"duration_ms":  queryDuration,
		"result_count": len(products),
		"shop_id":      filters.ShopID,
		"has_search":   filters.Search != nil,
	}).Debug("Products query with filters completed (lightweight - no variants)")

	// Alert if query is slow (target < 50ms for real-time search)
	if queryDuration > 50 {
		logs.WithFields(map[string]interface{}{
			"file":        ProductRepositoryField,
			"function":    "get_all_by_shop_id_with_filters",
			"duration_ms": queryDuration,
			"shop_id":     filters.ShopID,
		}).Warn("Slow query detected - target is <50ms for real-time search")
	}

	return products, nil
}

func (r *ProductRepository) CountByShopIDWithFilters(ctx context.Context, filters models.ProductFilters) (int, error) {
	startTime := time.Now()

	// Build COUNT query with same filters as GetAllByShopIDWithFilters
	// But WITHOUT cursor, ORDER BY, or LIMIT (we want total count)
	countQuery := "SELECT COUNT(*) FROM products p WHERE p.shop_id = $1"
	args := []interface{}{filters.ShopID}
	argPos := 2

	var conditions []string

	// Category filter
	if filters.CategoryID != nil {
		conditions = append(conditions, fmt.Sprintf("p.category_id = $%d", argPos))
		args = append(args, *filters.CategoryID)
		argPos++
	}

	// Active status filter
	if filters.IsActive != nil {
		conditions = append(conditions, fmt.Sprintf("p.is_active = $%d", argPos))
		args = append(args, *filters.IsActive)
		argPos++
	}

	// Highlighted filter
	if filters.IsHighlighted != nil {
		conditions = append(conditions, fmt.Sprintf("p.is_highlighted = $%d", argPos))
		args = append(args, *filters.IsHighlighted)
		argPos++
	}

	// Promotional filter
	if filters.IsPromotional != nil {
		conditions = append(conditions, fmt.Sprintf("p.is_promotional = $%d", argPos))
		args = append(args, *filters.IsPromotional)
		argPos++
	}

	// Minimum price filter
	if filters.MinPrice != nil {
		conditions = append(conditions, fmt.Sprintf("p.price >= $%d", argPos))
		args = append(args, *filters.MinPrice)
		argPos++
	}

	// Maximum price filter
	if filters.MaxPrice != nil {
		conditions = append(conditions, fmt.Sprintf("p.price <= $%d", argPos))
		args = append(args, *filters.MaxPrice)
		argPos++
	}

	// Search filter (same as GetAllByShopIDWithFilters - uses trigram index)
	if filters.Search != nil && *filters.Search != "" {
		searchTerm := *filters.Search
		conditions = append(conditions, fmt.Sprintf("p.name ILIKE $%d", argPos))
		args = append(args, "%"+searchTerm+"%")
	}

	// Append all WHERE conditions
	if len(conditions) > 0 {
		countQuery += " AND " + strings.Join(conditions, " AND ")
	}

	// Execute COUNT query
	logs.WithFields(map[string]interface{}{
		"file":          ProductRepositoryField,
		"function":      "count_by_shop_id_with_filters",
		"shop_id":       filters.ShopID,
		"has_search":    filters.Search != nil,
		"has_category":  filters.CategoryID != nil,
		"has_price_min": filters.MinPrice != nil,
		"has_price_max": filters.MaxPrice != nil,
	}).Debug("Executing COUNT query for pagination")

	var totalCount int
	err := r.db.QueryRowContext(ctx, countQuery, args...).Scan(&totalCount)
	if err != nil {
		logs.WithFields(map[string]interface{}{
			"file":        ProductRepositoryField,
			"function":    "count_by_shop_id_with_filters",
			"duration_ms": time.Since(startTime).Milliseconds(),
			"error":       err.Error(),
		}).Error("Failed to count products with filters")
		return 0, fmt.Errorf("database operation failed")
	}

	countDuration := time.Since(startTime).Milliseconds()

	logs.WithFields(map[string]interface{}{
		"file":        ProductRepositoryField,
		"function":    "count_by_shop_id_with_filters",
		"duration_ms": countDuration,
		"total_count": totalCount,
		"shop_id":     filters.ShopID,
		"has_search":  filters.Search != nil,
	}).Debug("COUNT query completed")

	// Alert if COUNT is slow (target < 100ms even with search)
	if countDuration > 100 {
		logs.WithFields(map[string]interface{}{
			"file":        ProductRepositoryField,
			"function":    "count_by_shop_id_with_filters",
			"duration_ms": countDuration,
			"shop_id":     filters.ShopID,
		}).Warn("Slow COUNT query detected - target is <100ms")
	}

	return totalCount, nil
}

func (r *ProductRepository) Update(ctx context.Context, productID int, product *models.Product) error {
	startTime := time.Now()

	// Serialize images to JSONB
	// Format: [{"id": 1, "url": "..."}, {"url": "new_image"}]
	imagesJSON, err := json.Marshal(product.Images)
	if err != nil {
		logs.WithFields(map[string]interface{}{
			"file":       ProductRepositoryField,
			"function":   ProductUpdateFunctionField,
			"sub_func":   MarshalImagesSubFuncField,
			"product_id": productID,
			"error":      err.Error(),
		}).Error(LogFailedMarshalImages)
		return fmt.Errorf("database operation failed")
	}

	// Serialize variants to JSONB
	// Format: [{"id": 1, "name": "...", "options": [...]}, {"name": "new", ...}]
	variantsJSON, err := json.Marshal(product.Variants)
	if err != nil {
		logs.WithFields(map[string]interface{}{
			"file":       ProductRepositoryField,
			"function":   ProductUpdateFunctionField,
			"sub_func":   MarshalVariantsSubFuncField,
			"product_id": productID,
			"error":      err.Error(),
		}).Error(LogFailedMarshalVariants)
		return fmt.Errorf("database operation failed")
	}

	logs.WithFields(map[string]interface{}{
		"file":          ProductRepositoryField,
		"function":      ProductUpdateFunctionField,
		"product_id":    productID,
		"image_count":   len(product.Images),
		"variant_count": len(product.Variants),
		"duration_ms":   time.Since(startTime).Milliseconds(),
	}).Debug("Data prepared for stored procedure")

	// Call stored procedure (single query does everything)
	spStart := time.Now()
	_, err = r.db.ExecContext(ctx, `
		SELECT update_product($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13)`,
		productID,
		product.Name,
		product.Description,
		product.Price,
		product.Stock,
		product.MinimumStock,
		product.IsActive,
		product.IsHighlighted,
		product.IsPromotional,
		product.PromotionalPrice,
		product.Category.ID,
		imagesJSON,
		variantsJSON,
	)

	if err != nil {
		logs.WithFields(map[string]interface{}{
			"file":       ProductRepositoryField,
			"function":   ProductUpdateFunctionField,
			"sub_func":   CallStoredProcedureSubFuncField,
			"product_id": productID,
			"error":      err.Error(),
		}).Error(LogFailedUpdateProductSP)

		// Check if it's a PostgreSQL error from the stored procedure
		if pqErr, ok := err.(*pq.Error); ok {
			// RAISE EXCEPTION from stored procedure comes as pq.Error
			logs.WithFields(map[string]interface{}{
				"file":       ProductRepositoryField,
				"function":   ProductUpdateFunctionField,
				"pg_code":    pqErr.Code,    // PostgreSQL error code
				"pg_message": pqErr.Message, // Error message from RAISE EXCEPTION
				"pg_detail":  pqErr.Detail,  // Additional detail if any
				"pg_hint":    pqErr.Hint,    // Hint if provided
			}).Debug("PostgreSQL error details from stored procedure")

			// Return error with SP context (preserves original message)
			return fmt.Errorf("stored procedure error: %s", pqErr.Message)
		}

		// Not a PostgreSQL error (network, context cancelled, etc.)
		return fmt.Errorf("database operation failed: %w", err)
	}

	logs.WithFields(map[string]interface{}{
		"file":        ProductRepositoryField,
		"function":    ProductUpdateFunctionField,
		"sub_func":    CallStoredProcedureSubFuncField,
		"product_id":  productID,
		"duration_ms": time.Since(spStart).Milliseconds(),
	}).Debug("Stored procedure executed successfully")

	logs.WithFields(map[string]interface{}{
		"file":              ProductRepositoryField,
		"function":          ProductUpdateFunctionField,
		"product_id":        productID,
		"total_duration_ms": time.Since(startTime).Milliseconds(),
	}).Info("Product update completed (stored procedure)")

	return nil
}
