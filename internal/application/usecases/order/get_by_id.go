package order

import (
	"context"

	"github.com/mlgaray/ecommerce_api/internal/core/models"
	"github.com/mlgaray/ecommerce_api/internal/core/ports"
)

// GetOrderByIDUseCase orchestrates retrieval of a single order with full details.
// Pure delegation to service layer (no extra business logic for reads).
type GetOrderByIDUseCase struct {
	orderService ports.OrderService
}

func NewGetOrderByIDUseCase(orderService ports.OrderService) ports.GetOrderByIDUseCase {
	return &GetOrderByIDUseCase{
		orderService: orderService,
	}
}

func (uc *GetOrderByIDUseCase) Execute(ctx context.Context, shopID int, orderID int) (*models.Order, error) {
	return uc.orderService.GetByID(ctx, shopID, orderID)
}
