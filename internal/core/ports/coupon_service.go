package ports

import (
	"context"

	"github.com/mlgaray/ecommerce_api/internal/core/models"
)

// CouponService contains business logic for coupon operations.
type CouponService interface {
	// Create creates a new coupon. Validates domain rules before persisting.
	Create(ctx context.Context, coupon *models.Coupon, shopID int) (*models.Coupon, error)

	// Update updates an existing coupon. Validates domain rules before persisting.
	Update(ctx context.Context, couponID, shopID int, coupon *models.Coupon) error

	// Delete deletes a coupon by ID.
	Delete(ctx context.Context, couponID, shopID int) error

	// GetByID retrieves a coupon by ID.
	GetByID(ctx context.Context, couponID, shopID int) (*models.Coupon, error)

	// GetAllByShopIDWithFilters retrieves coupons with filters.
	GetAllByShopIDWithFilters(ctx context.Context, shopID int, filters models.CouponFilters) ([]*models.Coupon, error)

	// CountByShopIDWithFilters returns total count.
	CountByShopIDWithFilters(ctx context.Context, shopID int, filters models.CouponFilters) (int, error)

	// ValidateCoupon validates a coupon code for a given store, phone, and subtotal.
	// Returns the validated coupon with DiscountAmount calculated, or an error if invalid.
	ValidateCoupon(ctx context.Context, couponCode string, shopID int, phone string, subtotal float64) (*models.Coupon, error)
}
