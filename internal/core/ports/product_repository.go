package ports

import (
	"context"

	"github.com/mlgaray/ecommerce_api/internal/core/models"
)

type ProductRepository interface {
	Create(ctx context.Context, product *models.Product, shopID int) (*models.Product, error)
}
