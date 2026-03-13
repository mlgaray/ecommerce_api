package coupon

import (
	"context"

	"github.com/mlgaray/ecommerce_api/internal/core/models"
	"github.com/mlgaray/ecommerce_api/internal/core/ports"
)

// ValidateStoreCouponUseCase orchestrates coupon validation for the public storefront.
// Resolves store slug → shop ID, then validates the coupon.
type ValidateStoreCouponUseCase struct {
	storeService  ports.StoreService
	couponService ports.CouponService
}

func NewValidateStoreCouponUseCase(
	storeService ports.StoreService,
	couponService ports.CouponService,
) ports.ValidateStoreCouponUseCase {
	return &ValidateStoreCouponUseCase{
		storeService:  storeService,
		couponService: couponService,
	}
}

func (uc *ValidateStoreCouponUseCase) Execute(ctx context.Context, storeSlug string, code string, phone string, subtotal float64) (*models.Coupon, error) {
	// 1. Resolve store slug → shop ID
	store, err := uc.storeService.GetBySlug(ctx, storeSlug)
	if err != nil {
		return nil, err
	}

	// 2. Validate coupon (active, temporal, min order, usage limits, calculate discount)
	return uc.couponService.ValidateCoupon(ctx, code, store.ID, phone, subtotal)
}
