package order

import (
	"context"

	"github.com/mlgaray/ecommerce_api/internal/core/errors"
	"github.com/mlgaray/ecommerce_api/internal/core/models"
	"github.com/mlgaray/ecommerce_api/internal/core/ports"
)

// UpdateOrderUseCase orchestrates the full order update flow.
// 1. Fetches current order (404 if not found)
// 2. Validates editability (422 if completed/cancelled)
// 3. Validates item data against store products (including stock)
// 4. Validates delivery method and shipping cost
// 5. Applies updates and calculates/validates totals
// 6. Validates domain rules
// 7. Persists changes
// 8. Re-fetches order for complete response
type UpdateOrderUseCase struct {
	orderService ports.OrderService
	storeService ports.StoreService
}

func NewUpdateOrderUseCase(
	orderService ports.OrderService,
	storeService ports.StoreService,
) ports.UpdateOrderUseCase {
	return &UpdateOrderUseCase{
		orderService: orderService,
		storeService: storeService,
	}
}

func (uc *UpdateOrderUseCase) Execute(ctx context.Context, shopID int, orderID int, updatedData *models.Order) (*models.Order, error) {
	// 1. Fetch existing order (returns 404 if not found)
	order, err := uc.orderService.GetByID(ctx, shopID, orderID)
	if err != nil {
		return nil, err
	}

	// 2. Validate editability (only pending/confirmed orders can be edited)
	if !order.IsEditable() {
		return nil, &errors.BusinessRuleError{Message: errors.OrderCannotBeEdited}
	}

	// 3. Validate order items against store products (including stock)
	if err := uc.storeService.ValidateOrderItems(ctx, updatedData.Items, order.Store.ID); err != nil {
		return nil, err
	}

	// 4. Validate delivery method and shipping cost (if delivery method provided)
	if updatedData.DeliveryMethod != nil {
		store, err := uc.storeService.GetBySlug(ctx, order.Store.Slug)
		if err != nil {
			return nil, err
		}
		if err := uc.storeService.ValidateDeliveryMethod(store, updatedData.DeliveryMethod, updatedData.ShippingCost); err != nil {
			return nil, err
		}
	}

	// 5. Apply mutable fields from updatedData to existing order
	// Preserve immutable fields: ID, OrderNumber, Status, Store, CreatedAt
	order.Customer = updatedData.Customer
	order.PaymentMethod = updatedData.PaymentMethod
	order.DeliveryMethod = updatedData.DeliveryMethod
	order.Items = updatedData.Items
	order.ShippingCost = updatedData.ShippingCost
	order.Subtotal = updatedData.Subtotal
	order.Total = updatedData.Total

	// 6. Calculate and validate totals (backend is source of truth)
	if err := uc.orderService.CalculateAndValidateTotals(order); err != nil {
		return nil, err
	}

	// 7. Validate domain rules
	if err := order.Validate(); err != nil {
		return nil, err
	}

	// 8. Persist (service receives already-validated data)
	if err := uc.orderService.Update(ctx, shopID, order); err != nil {
		return nil, err
	}

	// 9. Re-fetch for complete response (consistent with UpdateStatus pattern)
	return uc.orderService.GetByID(ctx, shopID, orderID)
}
