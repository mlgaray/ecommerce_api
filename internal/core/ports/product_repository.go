package ports

import (
	"context"

	"github.com/mlgaray/ecommerce_api/internal/core/models"
)

type ProductRepository interface {
	Create(ctx context.Context, product *models.Product, shopID int) (*models.Product, error)
	// GetAllByShopIDWithFilters retrieves products with filters.
	// ShopID is a context parameter (not a filter), passed separately.
	GetAllByShopIDWithFilters(ctx context.Context, shopID int, filters models.ProductFilters) ([]*models.Product, error)
	// CountByShopIDWithFilters returns total count of products matching filters (for pagination "X of Y")
	// ShopID is a context parameter (not a filter), passed separately.
	CountByShopIDWithFilters(ctx context.Context, shopID int, filters models.ProductFilters) (int, error)
	GetByID(ctx context.Context, productID int) (*models.Product, error)
	// Update updates product and returns storage_refs of deleted images for cleanup.
	// Validates that the product belongs to the specified shop.
	// Returns RecordNotFoundError if product doesn't exist or doesn't belong to shop.
	Update(ctx context.Context, productID int, product *models.Product, shopID int) (deletedRefs []string, err error)

	// Delete deletes a product by ID.
	// Validates that the product belongs to the specified shop via WHERE clause.
	// Returns array of storage_refs of deleted images for cleanup in external storage.
	// Returns RecordNotFoundError if product doesn't exist or doesn't belong to shop.
	// Related entities (images, variants, variant_options) are cascade-deleted automatically.
	Delete(ctx context.Context, productID, shopID int) (deletedRefs []string, err error)

	// GetByIDAndShopID retrieves a product by ID and validates shop ownership.
	// Used by public store endpoints where shop ID comes from slug lookup.
	// Returns RecordNotFoundError if product doesn't exist or doesn't belong to shop.
	GetByIDAndShopID(ctx context.Context, productID, shopID int) (*models.Product, error)

	// GetByIDsAndShopID retrieves multiple products by IDs and validates shop ownership.
	// Returns a map of productID -> Product for O(1) lookup during validation.
	// Used by StoreService.ValidateOrderItems for batch validation.
	// Products not found are simply not included in the returned map.
	GetByIDsAndShopID(ctx context.Context, ids []int, shopID int) (map[int]*models.Product, error)
}
