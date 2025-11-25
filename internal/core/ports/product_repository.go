package ports

import (
	"context"

	"github.com/mlgaray/ecommerce_api/internal/core/models"
)

type ProductRepository interface {
	Create(ctx context.Context, product *models.Product, shopID int) (*models.Product, error)
	GetAllByShopIDWithFilters(ctx context.Context, filters models.ProductFilters) ([]*models.Product, error)
	// CountByShopIDWithFilters returns total count of products matching filters (for pagination "X of Y")
	CountByShopIDWithFilters(ctx context.Context, filters models.ProductFilters) (int, error)
	GetByID(ctx context.Context, productID int) (*models.Product, error)
	Update(ctx context.Context, productID int, product *models.Product) error
}
