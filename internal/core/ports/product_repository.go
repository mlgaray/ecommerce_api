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
	// Update updates product and returns storage_refs of deleted images for cleanup
	Update(ctx context.Context, productID int, product *models.Product) (deletedRefs []string, err error)
}
