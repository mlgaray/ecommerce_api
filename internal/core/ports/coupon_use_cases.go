package ports

import (
	"context"

	"github.com/mlgaray/ecommerce_api/internal/core/models"
)

// CreateCouponUseCase creates a new coupon for a shop.
type CreateCouponUseCase interface {
	Execute(ctx context.Context, coupon *models.Coupon, shopID int) (*models.Coupon, error)
}

// UpdateCouponUseCase updates an existing coupon and returns the updated coupon with usage_count.
type UpdateCouponUseCase interface {
	Execute(ctx context.Context, couponID, shopID int, coupon *models.Coupon) (*models.Coupon, error)
}

// DeleteCouponUseCase deletes a coupon by ID.
type DeleteCouponUseCase interface {
	Execute(ctx context.Context, couponID, shopID int) error
}

// GetCouponByIDUseCase retrieves a coupon by ID.
type GetCouponByIDUseCase interface {
	Execute(ctx context.Context, couponID, shopID int) (*models.Coupon, error)
}

// GetAllCouponsByShopIDUseCase retrieves coupons with filters and pagination.
type GetAllCouponsByShopIDUseCase interface {
	Execute(ctx context.Context, shopID int, filters models.CouponFilters) ([]*models.Coupon, string, bool, *int, error)
}

// ValidateStoreCouponUseCase validates a coupon for a store (public storefront).
// Resolves store slug → shop ID, then delegates to CouponService.ValidateCoupon.
type ValidateStoreCouponUseCase interface {
	Execute(ctx context.Context, storeSlug string, code string, phone string, subtotal float64) (*models.Coupon, error)
}
