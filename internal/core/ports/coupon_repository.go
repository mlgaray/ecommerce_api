package ports

import (
	"context"

	"github.com/mlgaray/ecommerce_api/internal/core/models"
)

// CouponRepository handles coupon persistence operations.
type CouponRepository interface {
	// Create creates a new coupon in the database.
	// Returns DuplicateRecordError if code+shop_id already exists.
	Create(ctx context.Context, coupon *models.Coupon) (*models.Coupon, error)

	// Update updates an existing coupon.
	// Returns RecordNotFoundError if not found.
	Update(ctx context.Context, coupon *models.Coupon) error

	// Delete deletes a coupon by ID and shop ID.
	// Returns RecordNotFoundError if not found.
	Delete(ctx context.Context, couponID, shopID int) error

	// GetByID retrieves a coupon by ID and shop ID.
	// Returns RecordNotFoundError if not found.
	GetByID(ctx context.Context, couponID, shopID int) (*models.Coupon, error)

	// GetByCodeAndShopID retrieves a coupon by code and shop ID.
	// Returns RecordNotFoundError if not found.
	GetByCodeAndShopID(ctx context.Context, code string, shopID int) (*models.Coupon, error)

	// GetAllByShopIDWithFilters retrieves coupons with filters and pagination.
	// Returns coupons with LIMIT+1 strategy for cursor-based pagination.
	GetAllByShopIDWithFilters(ctx context.Context, shopID int, filters models.CouponFilters) ([]*models.Coupon, error)

	// CountByShopIDWithFilters returns total count of coupons matching filters.
	CountByShopIDWithFilters(ctx context.Context, shopID int, filters models.CouponFilters) (int, error)

	// CountUsages returns the total number of times a coupon has been used.
	CountUsages(ctx context.Context, couponID int) (int, error)

	// CountUsagesByPhone returns the number of times a coupon has been used by a specific phone.
	CountUsagesByPhone(ctx context.Context, couponID int, phone string) (int, error)
}
