package coupon

import (
	"context"

	"github.com/mlgaray/ecommerce_api/internal/core/models"
	"github.com/mlgaray/ecommerce_api/internal/core/ports"
)

// CreateCouponUseCase orchestrates creating a new coupon.
type CreateCouponUseCase struct {
	couponService ports.CouponService
}

func NewCreateCouponUseCase(couponService ports.CouponService) ports.CreateCouponUseCase {
	return &CreateCouponUseCase{couponService: couponService}
}

func (uc *CreateCouponUseCase) Execute(ctx context.Context, coupon *models.Coupon, shopID int) (*models.Coupon, error) {
	return uc.couponService.Create(ctx, coupon, shopID)
}
