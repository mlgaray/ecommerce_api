package ports

import (
	"context"

	"github.com/mlgaray/ecommerce_api/internal/core/models"
)

type GetAllByShopIDUseCase interface {
	Execute(ctx context.Context, shopID, limit, cursor int) ([]*models.Product, int, bool, error)
}
