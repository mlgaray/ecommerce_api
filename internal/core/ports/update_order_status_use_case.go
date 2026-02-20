package ports

import (
	"context"

	"github.com/mlgaray/ecommerce_api/internal/core/models"
)

// UpdateOrderStatusUseCase orchestrates the order status update flow.
type UpdateOrderStatusUseCase interface {
	Execute(ctx context.Context, shopID int, orderID int, newStatus models.OrderStatus) (*models.Order, error)
}
