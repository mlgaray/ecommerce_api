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
// 5. Validates and applies coupon (if present)
// 6. Applies updates and calculates/validates totals
// 7. Validates domain rules
// 8. Persists changes (SP handles order + coupon_usage atomically)
// 9. Re-fetches order for complete response
type UpdateOrderUseCase struct {
	orderService  ports.OrderService
	storeService  ports.StoreService
	couponService ports.CouponService
}

func NewUpdateOrderUseCase(
	orderService ports.OrderService,
	storeService ports.StoreService,
	couponService ports.CouponService,
) ports.UpdateOrderUseCase {
	return &UpdateOrderUseCase{
		orderService:  orderService,
		storeService:  storeService,
		couponService: couponService,
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
	if err := uc.validateDelivery(ctx, order, updatedData); err != nil {
		return nil, err
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

	// 6. Handle coupon logic
	if err := uc.applyCouponLogic(ctx, order, updatedData); err != nil {
		return nil, err
	}

	// 7. Calculate and validate totals (backend is source of truth)
	if err := uc.orderService.CalculateAndValidateTotals(order); err != nil {
		return nil, err
	}

	// 8. Validate domain rules
	if err := order.Validate(); err != nil {
		return nil, err
	}

	// 9. Persist (SP handles order + coupon_usage atomically)
	if err := uc.orderService.Update(ctx, shopID, order); err != nil {
		return nil, err
	}

	// 10. Re-fetch for complete response (consistent with UpdateStatus pattern)
	return uc.orderService.GetByID(ctx, shopID, orderID)
}

// validateDelivery validates delivery method and shipping cost against store configuration.
func (uc *UpdateOrderUseCase) validateDelivery(ctx context.Context, order *models.Order, updatedData *models.Order) error {
	if updatedData.DeliveryMethod == nil {
		return nil
	}
	store, err := uc.storeService.GetBySlug(ctx, order.Store.Slug)
	if err != nil {
		return err
	}
	return uc.storeService.ValidateDeliveryMethod(store, updatedData.DeliveryMethod, updatedData.ShippingCost)
}

// applyCouponLogic handles coupon application during order update (2 cases):
// a) coupon_code provided → apply new coupon (full validation)
// b) no coupon_code → preserve existing coupon (recalculate discount if subtotal changed)
// Note: coupon removal is handled by DELETE /orders/{id}/coupon (separate endpoint)
func (uc *UpdateOrderUseCase) applyCouponLogic(ctx context.Context, order *models.Order, updatedData *models.Order) error {
	if updatedData.Coupon != nil && updatedData.Coupon.Code != "" {
		// Case a: New coupon requested → full validation
		phone := ""
		if updatedData.Customer != nil {
			phone = updatedData.Customer.Phone
		}
		coupon, err := uc.couponService.ValidateCoupon(ctx, updatedData.Coupon.Code, order.Store.ID, phone, order.Subtotal)
		if err != nil {
			return err
		}
		order.Coupon = coupon
		order.Discount = coupon.DiscountAmount
		order.Total = order.Subtotal - order.Discount + order.ShippingCost
	} else if order.Coupon != nil {
		// Case b: Preserve existing coupon, recalculate discount for new subtotal
		order.Discount = order.Coupon.CalculateDiscount(order.Subtotal)
		order.Coupon.DiscountAmount = order.Discount
		order.Total = order.Subtotal - order.Discount + order.ShippingCost
	}
	return nil
}
