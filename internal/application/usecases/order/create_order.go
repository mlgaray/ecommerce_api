package order

import (
	"context"

	"github.com/mlgaray/ecommerce_api/internal/core/models"
	"github.com/mlgaray/ecommerce_api/internal/core/ports"
)

// CreateOrderUseCase orchestrates creating a new order.
// Uses StoreService for validation, OrderService for persistence,
// and OrderEventNotifier for real-time notifications.
type CreateOrderUseCase struct {
	storeService ports.StoreService
	orderService ports.OrderService
	notifier     ports.OrderEventNotifier
}

func NewCreateOrderUseCase(
	storeService ports.StoreService,
	orderService ports.OrderService,
	notifier ports.OrderEventNotifier,
) ports.CreateOrderUseCase {
	return &CreateOrderUseCase{
		storeService: storeService,
		orderService: orderService,
		notifier:     notifier,
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
	createdOrder, err := uc.orderService.Create(ctx, order, store)
	if err != nil {
		return nil, err
	}

	// 5. Notify listeners (fire-and-forget)
	uc.notifier.NotifyNewOrder(ctx, createdOrder)

	return createdOrder, nil
}
