package ports

import (
	"context"

	"github.com/mlgaray/ecommerce_api/internal/core/models"
)

// UpdateOrderUseCase orchestrates the full order update flow.
// Validates editability, items, delivery method, totals, and domain rules
// before delegating persistence to the service.
type UpdateOrderUseCase interface {
	Execute(ctx context.Context, shopID int, orderID int, updatedData *models.Order) (*models.Order, error)
}
