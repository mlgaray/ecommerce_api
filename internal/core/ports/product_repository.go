package ports

import (
	"context"

	"github.com/mlgaray/ecommerce_api/internal/core/models"
)

type ProductRepository interface {
	Create(ctx context.Context, product *models.Product, shopID int) (*models.Product, error)
	GetAllByShopID(ctx context.Context, shopID, limit, cursor int) ([]*models.Product, error)
	GetByID(ctx context.Context, productID int) (*models.Product, error)
	Update(ctx context.Context, productID int, product *models.Product) error
}
