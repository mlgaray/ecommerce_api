package coupon

import (
	"context"

	"github.com/mlgaray/ecommerce_api/internal/core/models"
	"github.com/mlgaray/ecommerce_api/internal/core/ports"
)

// GetCouponByIDUseCase orchestrates retrieving a coupon by ID.
type GetCouponByIDUseCase struct {
	couponService ports.CouponService
}

func NewGetCouponByIDUseCase(couponService ports.CouponService) ports.GetCouponByIDUseCase {
	return &GetCouponByIDUseCase{couponService: couponService}
}

func (uc *GetCouponByIDUseCase) Execute(ctx context.Context, couponID, shopID int) (*models.Coupon, error) {
	return uc.couponService.GetByID(ctx, couponID, shopID)
}
