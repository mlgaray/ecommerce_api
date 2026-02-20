package ports

import (
	"context"

	"github.com/mlgaray/ecommerce_api/internal/core/models"
)

// GetOrderByIDUseCase orchestrates retrieval of a single order with full details.
type GetOrderByIDUseCase interface {
	Execute(ctx context.Context, shopID int, orderID int) (*models.Order, error)
}
