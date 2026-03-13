package order

import (
	"context"

	"github.com/mlgaray/ecommerce_api/internal/core/models"
	"github.com/mlgaray/ecommerce_api/internal/core/ports"
)

// CreateOrderUseCase orchestrates creating a new order.
// Uses StoreService for validation, CouponService for coupon validation,
// OrderService for persistence, and OrderEventNotifier for real-time notifications.
type CreateOrderUseCase struct {
	storeService  ports.StoreService
	couponService ports.CouponService
	orderService  ports.OrderService
	notifier      ports.OrderEventNotifier
}

func NewCreateOrderUseCase(
	storeService ports.StoreService,
	couponService ports.CouponService,
	orderService ports.OrderService,
	notifier ports.OrderEventNotifier,
) ports.CreateOrderUseCase {
	return &CreateOrderUseCase{
		storeService:  storeService,
		couponService: couponService,
		orderService:  orderService,
		notifier:      notifier,
	}
}

// Execute creates a new order for a store.
// Orchestrates:
// 1. Get store by slug
// 2. Validate order items (product data, stock)
// 3. Validate delivery method and shipping cost
// 4. Validate and apply coupon (if present)
// 5. Persist order (SP handles order + coupon_usage atomically)
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

	// 4. Validate and apply coupon (if present)
	if order.Coupon != nil && order.Coupon.Code != "" {
		coupon, err := uc.couponService.ValidateCoupon(ctx, order.Coupon.Code, store.ID, order.Customer.Phone, order.Subtotal)
		if err != nil {
			return nil, err
		}
		order.Coupon = coupon
		order.Discount = coupon.DiscountAmount
		order.Total = order.Subtotal - order.Discount + order.ShippingCost
	}

	// 5. Persist order (SP persists order + coupon_usage atomically)
	createdOrder, err := uc.orderService.Create(ctx, order, store)
	if err != nil {
		return nil, err
	}

	// 6. Notify listeners (fire-and-forget)
	uc.notifier.NotifyNewOrder(ctx, createdOrder)

	return createdOrder, nil
}
