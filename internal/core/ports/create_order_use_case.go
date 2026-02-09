package ports

import (
	"context"

	"github.com/mlgaray/ecommerce_api/internal/core/models"
)

// CreateOrderUseCase orchestrates creating a new order.
type CreateOrderUseCase interface {
	Execute(ctx context.Context, order *models.Order, storeSlug string) (*models.Order, error)
}
