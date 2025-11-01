package postgresql

import (
	"context"
	"database/sql"
	"encoding/json"

	"github.com/mlgaray/ecommerce_api/internal/core/errors"
	"github.com/mlgaray/ecommerce_api/internal/core/models"
	"github.com/mlgaray/ecommerce_api/internal/infraestructure/adapters/logs"
)

type ProductRepository struct {
	db *sql.DB
}

// Product domain-specific log message constants
const (
	// create
	LogFailedInsertProduct       = "Failed to insert product"
	LogFailedInsertProductImage  = "Failed to insert product image"
	LogFailedInsertVariant       = "Failed to insert variant"
	LogFailedInsertVariantOption = "Failed to insert variant option"
	ProductRepositoryField       = "product_repository"
	CreateFunctionField          = "create"
	// getAllByShopIDField
	getAllByShopIDField      = "get_all_by_shop_id"
	failedReadProductsByShop = "Failed to read products by shop"
)

func NewProductRepository(dataBaseConnection DataBaseConnection) *ProductRepository {
	return &ProductRepository{
		db: dataBaseConnection.Connect(),
	}
}

func (r *ProductRepository) Create(ctx context.Context, product *models.Product, shopID int) (*models.Product, error) {
	// Check if we're in a transaction
	if tx, ok := ctx.Value(TxContextKey).(*sql.Tx); ok {
		return r.createWithTx(ctx, tx, product, shopID)
	}

	// Start our own transaction if not already in one
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		logs.WithFields(map[string]interface{}{
			"file":     ProductRepositoryField,
			"function": CreateFunctionField,
			"sub_func": BeginTransactionField,
			"error":    err.Error(),
		}).Error(FailedBeginTransactionLog)
		return nil, &errors.InternalServiceError{Message: errors.DatabaseError}
	}
	defer func() {
		if err := tx.Rollback(); err != nil && err != sql.ErrTxDone {
			logs.WithFields(map[string]interface{}{
				"operation": "rollback_transaction",
				"error":     err.Error(),
			}).Error("Failed to rollback transaction")
		}
	}()

	createdProduct, err := r.createWithTx(ctx, tx, product, shopID)
	if err != nil {
		return nil, err
	}

	// Commit transaction
	if err := tx.Commit(); err != nil {
		logs.WithFields(map[string]interface{}{
			"operation": CommitTransactionField,
			"error":     err.Error(),
		}).Error(FailedCommitTransactionLog)
		return nil, &errors.InternalServiceError{Message: errors.DatabaseError}
	}

	return createdProduct, nil
}

func (r *ProductRepository) createWithTx(ctx context.Context, tx *sql.Tx, product *models.Product, shopID int) (*models.Product, error) {
	// 1. Insert product
	productID, err := r.insertProduct(ctx, tx, product, shopID)
	if err != nil {
		return nil, err
	}
	product.ID = productID

	// 2. Insert product images
	if len(product.Images) > 0 {
		if err := r.insertProductImages(ctx, tx, productID, product.Images); err != nil {
			return nil, err
		}
	}

	// 3. Insert variants and their options
	if len(product.Variants) > 0 {
		if err := r.insertProductVariants(ctx, tx, productID, product.Variants); err != nil {
			return nil, err
		}
	}

	return product, nil
}

func (r *ProductRepository) insertProduct(ctx context.Context, tx *sql.Tx, product *models.Product, shopID int) (int, error) {
	query := `
		INSERT INTO products (name, description, price, stock, is_active, is_highlighted, is_promotional, promotional_price, category_id, shop_id)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
		RETURNING id`

	var id int
	err := tx.QueryRowContext(ctx, query,
		product.Name,
		product.Description,
		product.Price,
		product.Stock,
		product.IsActive,
		product.IsHighlighted,
		product.IsPromotional,
		product.PromotionalPrice,
		product.Category.ID,
		shopID,
	).Scan(&id)

	if err != nil {
		logs.WithFields(map[string]interface{}{
			"operation": "insert_product",
			"shop_id":   shopID,
			"error":     err.Error(),
		}).Error(LogFailedInsertProduct)
		return 0, &errors.InternalServiceError{Message: errors.DatabaseError}
	}

	return id, nil
}

func (r *ProductRepository) insertProductImages(ctx context.Context, tx *sql.Tx, productID int, imageURLs []string) error {
	query := `INSERT INTO product_images (url, product_id) VALUES ($1, $2)`

	for _, imageURL := range imageURLs {
		_, err := tx.ExecContext(ctx, query, imageURL, productID)
		if err != nil {
			logs.WithFields(map[string]interface{}{
				"operation":  "insert_product_image",
				"product_id": productID,
				"image_url":  imageURL,
				"error":      err.Error(),
			}).Error(LogFailedInsertProductImage)
			return &errors.InternalServiceError{Message: errors.DatabaseError}
		}
	}

	return nil
}

func (r *ProductRepository) insertProductVariants(ctx context.Context, tx *sql.Tx, productID int, variants []*models.Variant) error {
	for _, variant := range variants {
		// Insert variant
		variantID, err := r.insertVariant(ctx, tx, productID, variant)
		if err != nil {
			return err
		}
		variant.ID = variantID

		// Insert variant options
		if len(variant.Options) > 0 {
			if err := r.insertVariantOptions(ctx, tx, variantID, variant.Options); err != nil {
				return err
			}
		}
	}

	return nil
}

func (r *ProductRepository) insertVariant(ctx context.Context, tx *sql.Tx, productID int, variant *models.Variant) (int, error) {
	query := `
		INSERT INTO product_variants (name, "order", selection_type, max_selections, product_id)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING id`

	var id int
	err := tx.QueryRowContext(ctx, query,
		variant.Name,
		variant.Order,
		variant.SelectionType,
		variant.MaxSelections,
		productID,
	).Scan(&id)

	if err != nil {
		logs.WithFields(map[string]interface{}{
			"operation":    "insert_variant",
			"product_id":   productID,
			"variant_name": variant.Name,
			"error":        err.Error(),
		}).Error(LogFailedInsertVariant)
		return 0, &errors.InternalServiceError{Message: errors.DatabaseError}
	}

	return id, nil
}

func (r *ProductRepository) insertVariantOptions(ctx context.Context, tx *sql.Tx, variantID int, options []*models.Option) error {
	query := `INSERT INTO variant_options (name, price, "order", variant_id) VALUES ($1, $2, $3, $4)`

	for _, option := range options {
		_, err := tx.ExecContext(ctx, query,
			option.Name,
			option.Price,
			option.Order,
			variantID,
		)
		if err != nil {
			logs.WithFields(map[string]interface{}{
				"operation":   "insert_variant_option",
				"variant_id":  variantID,
				"option_name": option.Name,
				"error":       err.Error(),
			}).Error(LogFailedInsertVariantOption)
			return &errors.InternalServiceError{Message: errors.DatabaseError}
		}
	}

	return nil
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
					(SELECT json_agg(pi2.url ORDER BY pi2.id)
					 FROM product_images pi2
					 WHERE pi2.product_id = p.id),
					'[]'
				) AS images,
				COALESCE(
					(SELECT json_agg(
						jsonb_build_object(
							'id', pv2.id,
							'name', pv2.name,
							'order', pv2."order",
							'selection_type', pv2.selection_type,
							'max_selections', pv2.max_selections,
							'options', (
								SELECT COALESCE(json_agg(
									jsonb_build_object(
										'id', vo.id,
										'name', vo.name,
										'price', vo.price,
										'order', vo."order"
									) ORDER BY vo."order"
								), '[]')
								FROM variant_options vo
								WHERE vo.variant_id = pv2.id
							)
						) ORDER BY pv2."order"
					)
					FROM product_variants pv2
					WHERE pv2.product_id = p.id),
					'[]'
				) AS variants
			FROM products p
			INNER JOIN categories c ON p.category_id = c.id
			WHERE p.shop_id = $1 AND p.is_active = true AND p.id < $2
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
					(SELECT json_agg(pi2.url ORDER BY pi2.id)
					 FROM product_images pi2
					 WHERE pi2.product_id = p.id),
					'[]'
				) AS images,
				COALESCE(
					(SELECT json_agg(
						jsonb_build_object(
							'id', pv2.id,
							'name', pv2.name,
							'order', pv2."order",
							'selection_type', pv2.selection_type,
							'max_selections', pv2.max_selections,
							'options', (
								SELECT COALESCE(json_agg(
									jsonb_build_object(
										'id', vo.id,
										'name', vo.name,
										'price', vo.price,
										'order', vo."order"
									) ORDER BY vo."order"
								), '[]')
								FROM variant_options vo
								WHERE vo.variant_id = pv2.id
							)
						) ORDER BY pv2."order"
					)
					FROM product_variants pv2
					WHERE pv2.product_id = p.id),
					'[]'
				) AS variants
			FROM products p
			INNER JOIN categories c ON p.category_id = c.id
			WHERE p.shop_id = $1 AND p.is_active = true
			ORDER BY p.id DESC
			LIMIT $2`
		args = []interface{}{shopID, limit}
	}

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		logs.WithFields(map[string]interface{}{
			"file":     ProductRepositoryField,
			"function": getAllByShopIDField,
			"sub_func": BeginTransactionField,
			"shop_id":  shopID,
			"error":    err.Error(),
		}).Error(failedReadProductsByShop)
		return nil, &errors.InternalServiceError{Message: errors.DatabaseError}
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
				"function": getAllByShopIDField,
				"sub_func": ScanField,
				"shop_id":  shopID,
				"error":    err.Error(),
			}).Error("Failed to scan product row")
			return nil, &errors.InternalServiceError{Message: errors.DatabaseError}
		}

		// Parse images JSON
		if err := json.Unmarshal(imagesJSON, &product.Images); err != nil {
			logs.WithFields(map[string]interface{}{
				"file":       ProductRepositoryField,
				"function":   getAllByShopIDField,
				"sub_func":   UnmarshallField,
				"product_id": product.ID,
				"error":      err.Error(),
			}).Error("Failed to unmarshal product images")
			return nil, &errors.InternalServiceError{Message: errors.DatabaseError}
		}

		// Parse variants JSON
		if err := json.Unmarshal(variantsJSON, &product.Variants); err != nil {
			logs.WithFields(map[string]interface{}{
				"file":       ProductRepositoryField,
				"function":   getAllByShopIDField,
				"sub_func":   UnmarshallField,
				"product_id": product.ID,
				"error":      err.Error(),
			}).Error("Failed to unmarshal product variants")
			return nil, &errors.InternalServiceError{Message: errors.DatabaseError}
		}

		products = append(products, product)
	}

	if err := rows.Err(); err != nil {
		logs.WithFields(map[string]interface{}{
			"file":     ProductRepositoryField,
			"function": getAllByShopIDField,
			"sub_func": NextField,
			"shop_id":  shopID,
			"error":    err.Error(),
		}).Error("Error iterating product rows")
		return nil, &errors.InternalServiceError{Message: errors.DatabaseError}
	}

	return products, nil
}
