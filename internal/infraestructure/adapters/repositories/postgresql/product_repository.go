package postgresql

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"time"

	"github.com/lib/pq"

	"github.com/mlgaray/ecommerce_api/internal/core/errors"
	"github.com/mlgaray/ecommerce_api/internal/core/models"
	"github.com/mlgaray/ecommerce_api/internal/infraestructure/adapters/logs"
)

type ProductRepository struct {
	db *sql.DB
}

// Product repository log field constants
const (
	ProductRepositoryField             = "product_repository"
	ProductCreateFunctionField         = "create"
	ProductGetAllByShopIDFunctionField = "get_all_by_shop_id"
	ProductGetByIDFunctionField        = "get_by_id"
	ProductUpdateFunctionField         = "update"
	ProductUnmarshallSubFuncField      = "unmarshall"
	MarshalVariantsSubFuncField        = "marshal_variants"
	MarshalImagesSubFuncField          = "marshal_images"
	CallStoredProcedureSubFuncField    = "call_stored_procedure"
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

func (r *ProductRepository) GetAllByShopID(ctx context.Context, shopID, limit, cursor int) ([]*models.Product, error) {
	// Default limit if not specified
	if limit <= 0 {
		limit = 20
	}
	// Max limit to prevent abuse
	if limit > 100 {
		limit = 100
	}

	var query string
	var args []interface{}

	if cursor > 0 {
		// Cursor-based pagination
		query = `
			SELECT
				p.id, p.name, p.description, p.price, p.stock, COALESCE(p.minimum_stock, 0),
				p.is_active, p.is_highlighted, p.is_promotional, COALESCE(p.promotional_price, 0),
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
			WHERE p.shop_id = $1 AND p.id < $2
			ORDER BY p.id DESC
			LIMIT $3`
		args = []interface{}{shopID, cursor, limit}
	} else {
		// First page
		query = `
			SELECT
				p.id, p.name, p.description, p.price, p.stock, COALESCE(p.minimum_stock, 0),
				p.is_active, p.is_highlighted, p.is_promotional, COALESCE(p.promotional_price, 0),
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
			WHERE p.shop_id = $1
			ORDER BY p.id DESC
			LIMIT $2`
		args = []interface{}{shopID, limit}
	}

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		logs.WithFields(map[string]interface{}{
			"file":     ProductRepositoryField,
			"function": ProductGetAllByShopIDFunctionField,
			"sub_func": BeginTransactionField,
			"shop_id":  shopID,
			"error":    err.Error(),
		}).Error(failedReadProductsByShop)
		return nil, fmt.Errorf("database operation failed")
	}
	defer rows.Close()

	products := make([]*models.Product, 0)

	for rows.Next() {
		product := &models.Product{
			Category: &models.Category{},
		}

		var imagesJSON, variantsJSON []byte

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
			&product.Category.ID,
			&product.Category.Name,
			&product.Category.Description,
			&imagesJSON,
			&variantsJSON,
		)
		if err != nil {
			logs.WithFields(map[string]interface{}{
				"file":     ProductRepositoryField,
				"function": ProductGetAllByShopIDFunctionField,
				"sub_func": ScanField,
				"shop_id":  shopID,
				"error":    err.Error(),
			}).Error("Failed to scan product row")
			return nil, fmt.Errorf("database operation failed")
		}

		// Parse images JSON
		if err := json.Unmarshal(imagesJSON, &product.Images); err != nil {
			return nil, fmt.Errorf("database operation failed")
		}

		// Parse variants JSON
		if err := json.Unmarshal(variantsJSON, &product.Variants); err != nil {
			logs.WithFields(map[string]interface{}{
				"file":       ProductRepositoryField,
				"function":   ProductGetAllByShopIDFunctionField,
				"sub_func":   UnmarshallField,
				"product_id": product.ID,
				"error":      err.Error(),
			}).Error("Failed to unmarshal product variants")
			return nil, fmt.Errorf("database operation failed")
		}

		products = append(products, product)
	}

	if err := rows.Err(); err != nil {
		logs.WithFields(map[string]interface{}{
			"file":     ProductRepositoryField,
			"function": ProductGetAllByShopIDFunctionField,
			"sub_func": NextField,
			"shop_id":  shopID,
			"error":    err.Error(),
		}).Error("Error iterating product rows")
		return nil, fmt.Errorf("database operation failed")
	}

	return products, nil
}

func (r *ProductRepository) GetByID(ctx context.Context, productID int) (*models.Product, error) {
	query := `
		SELECT
			p.id, p.name, p.description, p.price, p.stock, COALESCE(p.minimum_stock, 0),
			p.is_active, p.is_highlighted, p.is_promotional, COALESCE(p.promotional_price, 0),
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
