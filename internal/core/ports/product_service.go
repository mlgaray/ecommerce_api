package ports

import (
	"context"

	"github.com/mlgaray/ecommerce_api/internal/core/models"
)

type ProductService interface {
	Create(ctx context.Context, product *models.Product, imageBuffers [][]byte, shopID int) (*models.Product, error)
	GetAllByShopID(ctx context.Context, shopID, limit, cursor int) ([]*models.Product, int, bool, error)
	GetByID(ctx context.Context, productID int) (*models.Product, error)
	Update(ctx context.Context, productID int, product *models.Product, newImageBuffers [][]byte) error
}
