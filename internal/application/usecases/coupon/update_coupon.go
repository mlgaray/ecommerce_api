package coupon

import (
	"context"

	"github.com/mlgaray/ecommerce_api/internal/core/models"
	"github.com/mlgaray/ecommerce_api/internal/core/ports"
)

// UpdateCouponUseCase orchestrates updating an existing coupon.
type UpdateCouponUseCase struct {
	couponService ports.CouponService
}

func NewUpdateCouponUseCase(couponService ports.CouponService) ports.UpdateCouponUseCase {
	return &UpdateCouponUseCase{couponService: couponService}
}

func (uc *UpdateCouponUseCase) Execute(ctx context.Context, couponID, shopID int, coupon *models.Coupon) (*models.Coupon, error) {
	if err := uc.couponService.Update(ctx, couponID, shopID, coupon); err != nil {
		return nil, err
	}
	// Re-fetch to return complete data including usage_count
	return uc.couponService.GetByID(ctx, couponID, shopID)
}
