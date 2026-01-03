package ports

import (
	"context"

	"github.com/mlgaray/ecommerce_api/internal/core/models"
)

type GetShopByIDUseCase interface {
	Execute(ctx context.Context, shopID int) (*models.Shop, error)
}
