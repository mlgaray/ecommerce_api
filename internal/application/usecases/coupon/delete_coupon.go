package coupon

import (
	"context"

	"github.com/mlgaray/ecommerce_api/internal/core/ports"
)

// DeleteCouponUseCase orchestrates deleting a coupon.
type DeleteCouponUseCase struct {
	couponService ports.CouponService
}

func NewDeleteCouponUseCase(couponService ports.CouponService) ports.DeleteCouponUseCase {
	return &DeleteCouponUseCase{couponService: couponService}
}

func (uc *DeleteCouponUseCase) Execute(ctx context.Context, couponID, shopID int) error {
	return uc.couponService.Delete(ctx, couponID, shopID)
}
