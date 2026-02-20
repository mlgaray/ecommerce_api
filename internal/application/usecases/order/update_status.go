package order

import (
	"context"

	"github.com/mlgaray/ecommerce_api/internal/core/models"
	"github.com/mlgaray/ecommerce_api/internal/core/ports"
)

// UpdateOrderStatusUseCase orchestrates the order status update flow.
// 1. Fetches current order (404 if not found)
// 2. Validates transition + persists with optimistic lock (400/409)
// 3. Re-fetches order for complete response
type UpdateOrderStatusUseCase struct {
	orderService ports.OrderService
}

func NewUpdateOrderStatusUseCase(orderService ports.OrderService) ports.UpdateOrderStatusUseCase {
	return &UpdateOrderStatusUseCase{
		orderService: orderService,
	}
}

func (uc *UpdateOrderStatusUseCase) Execute(ctx context.Context, shopID int, orderID int, newStatus models.OrderStatus) (*models.Order, error) {
	order, err := uc.orderService.GetByID(ctx, shopID, orderID)
	if err != nil {
		return nil, err
	}

	if err := uc.orderService.UpdateStatus(ctx, shopID, order, newStatus); err != nil {
		return nil, err
	}

	return uc.orderService.GetByID(ctx, shopID, orderID)
}
