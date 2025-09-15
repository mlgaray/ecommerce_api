package ports

import (
	"context"

	"github.com/mlgaray/ecommerce_api/internal/core/models"
)

type ShopRepository interface {
	Create(ctx context.Context, shop *models.Shop) (*models.Shop, error)
}
