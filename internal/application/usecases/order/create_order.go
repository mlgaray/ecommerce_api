package order

import (
	"context"

	"github.com/mlgaray/ecommerce_api/internal/core/models"
	"github.com/mlgaray/ecommerce_api/internal/core/ports"
)

// CreateOrderUseCase orchestrates creating a new order.
// Uses StoreService for validation and OrderService for persistence.
type CreateOrderUseCase struct {
	storeService ports.StoreService
	orderService ports.OrderService
}

func NewCreateOrderUseCase(
	storeService ports.StoreService,
	orderService ports.OrderService,
) ports.CreateOrderUseCase {
	return &CreateOrderUseCase{
		storeService: storeService,
		orderService: orderService,
	}
}

// Execute creates a new order for a store.
// Orchestrates:
// 1. Get store by slug
// 2. Validate order items (product data, stock)
// 3. Validate delivery method and shipping cost
// 4. Persist order
func (uc *CreateOrderUseCase) Execute(ctx context.Context, order *models.Order, storeSlug string) (*models.Order, error) {
	// 1. Get store by slug
	store, err := uc.storeService.GetBySlug(ctx, storeSlug)
	if err != nil {
		return nil, err
	}

	// 2. Validate order items (product data matches DB, stock available)
	if err := uc.storeService.ValidateOrderItems(ctx, order.Items, store.ID); err != nil {
		return nil, err
	}

	// 3. Validate delivery method and shipping cost
	if err := uc.storeService.ValidateDeliveryMethod(store, order.DeliveryMethod, order.ShippingCost); err != nil {
		return nil, err
	}

	// 4. Persist order
	return uc.orderService.Create(ctx, order, store)
}
