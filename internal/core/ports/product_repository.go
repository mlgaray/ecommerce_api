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
}
