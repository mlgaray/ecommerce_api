package order

import (
	"context"

	"github.com/mlgaray/ecommerce_api/internal/core/errors"
	"github.com/mlgaray/ecommerce_api/internal/core/models"
	"github.com/mlgaray/ecommerce_api/internal/core/ports"
)

// RemoveOrderCouponUseCase removes the coupon from an order.
// 1. Fetches order by ID and shopID (404 if not found)
// 2. Validates editability (422 if completed/cancelled)
// 3. If no coupon → returns order as-is (no-op)
// 4. Clears coupon, discount, and recalculates total
// 5. Persists via service (SP handles coupon_usages cleanup)
// 6. Re-fetches and returns updated order
type RemoveOrderCouponUseCase struct {
	orderService ports.OrderService
}

func NewRemoveOrderCouponUseCase(orderService ports.OrderService) ports.RemoveOrderCouponUseCase {
	return &RemoveOrderCouponUseCase{
		orderService: orderService,
	}
}

func (uc *RemoveOrderCouponUseCase) Execute(ctx context.Context, orderID, shopID int) (*models.Order, error) {
	// 1. Fetch existing order (returns 404 if not found)
	order, err := uc.orderService.GetByID(ctx, shopID, orderID)
	if err != nil {
		return nil, err
	}

	// 2. Validate editability (only pending/confirmed orders can be edited)
	if !order.IsEditable() {
		return nil, &errors.BusinessRuleError{Message: errors.OrderCannotBeEdited}
	}

	// 3. Remove coupon via domain model (no-op if no coupon)
	if !order.RemoveCoupon() {
		return order, nil
	}

	// 4. Persist changes (SP handles coupon_usages cleanup when coupon_id is NULL)
	if err := uc.orderService.Update(ctx, shopID, order); err != nil {
		return nil, err
	}

	// 6. Re-fetch for complete response (consistent with Update pattern)
	return uc.orderService.GetByID(ctx, shopID, orderID)
}
