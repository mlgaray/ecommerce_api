package ports

import (
	"context"

	"github.com/mlgaray/ecommerce_api/internal/core/models"
)

type GetByIDUseCase interface {
	Execute(ctx context.Context, productID int) (*models.Product, error)
}
