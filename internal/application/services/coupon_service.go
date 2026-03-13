package services

import (
	"context"
	"strings"
	"time"

	"github.com/mlgaray/ecommerce_api/internal/core/errors"
	"github.com/mlgaray/ecommerce_api/internal/core/models"
	"github.com/mlgaray/ecommerce_api/internal/core/ports"
)

// CouponService contains business logic for coupon operations.
type CouponService struct {
	couponRepository ports.CouponRepository
}

func NewCouponService(couponRepository ports.CouponRepository) *CouponService {
	return &CouponService{
		couponRepository: couponRepository,
	}
}

// Create validates and persists a new coupon.
// Validate() normalizes code (trim + uppercase) automatically.
func (s *CouponService) Create(ctx context.Context, coupon *models.Coupon, shopID int) (*models.Coupon, error) {
	coupon.ShopID = shopID

	if err := coupon.Validate(); err != nil {
		return nil, err
	}

	return s.couponRepository.Create(ctx, coupon)
}

// Update validates and persists coupon changes.
// Validate() normalizes code (trim + uppercase) automatically.
func (s *CouponService) Update(ctx context.Context, couponID, shopID int, coupon *models.Coupon) error {
	coupon.ID = couponID
	coupon.ShopID = shopID

	if err := coupon.Validate(); err != nil {
		return err
	}

	return s.couponRepository.Update(ctx, coupon)
}

// Delete deletes a coupon by ID.
func (s *CouponService) Delete(ctx context.Context, couponID, shopID int) error {
	return s.couponRepository.Delete(ctx, couponID, shopID)
}

// GetByID retrieves a coupon by ID.
func (s *CouponService) GetByID(ctx context.Context, couponID, shopID int) (*models.Coupon, error) {
	return s.couponRepository.GetByID(ctx, couponID, shopID)
}

// GetAllByShopIDWithFilters retrieves coupons with filters.
func (s *CouponService) GetAllByShopIDWithFilters(ctx context.Context, shopID int, filters models.CouponFilters) ([]*models.Coupon, error) {
	return s.couponRepository.GetAllByShopIDWithFilters(ctx, shopID, filters)
}

// CountByShopIDWithFilters returns total count.
func (s *CouponService) CountByShopIDWithFilters(ctx context.Context, shopID int, filters models.CouponFilters) (int, error) {
	return s.couponRepository.CountByShopIDWithFilters(ctx, shopID, filters)
}

// ValidateCoupon validates a coupon code for a given store, phone, and subtotal.
// Returns the validated coupon with DiscountAmount calculated.
func (s *CouponService) ValidateCoupon(ctx context.Context, couponCode string, shopID int, phone string, subtotal float64) (*models.Coupon, error) {
	// 1. Normalize and get coupon by code and shop
	couponCode = strings.ToUpper(strings.TrimSpace(couponCode))
	coupon, err := s.couponRepository.GetByCodeAndShopID(ctx, couponCode, shopID)
	if err != nil {
		return nil, err
	}

	// 2. Validate coupon is active
	if !coupon.IsActive {
		return nil, &errors.BusinessRuleError{Message: errors.CouponNotActive}
	}

	// 3. Validate temporal validity (starts_at / expires_at)
	if err := s.validateTemporalValidity(coupon, time.Now()); err != nil {
		return nil, err
	}

	// 4. Validate minimum order amount
	if coupon.MinOrderAmount > 0 && subtotal < coupon.MinOrderAmount {
		return nil, &errors.BusinessRuleError{Message: errors.CouponMinOrderNotMet}
	}

	// 5. Validate usage limits (global and per-phone)
	if err := s.validateUsageLimits(ctx, coupon, phone); err != nil {
		return nil, err
	}

	// 6. Calculate discount amount
	coupon.DiscountAmount = coupon.CalculateDiscount(subtotal)

	return coupon, nil
}

// validateTemporalValidity checks that the coupon is within its valid date range.
func (s *CouponService) validateTemporalValidity(coupon *models.Coupon, now time.Time) error {
	if coupon.ExpiresAt != nil && now.After(*coupon.ExpiresAt) {
		return &errors.BusinessRuleError{Message: errors.CouponExpired}
	}
	if coupon.StartsAt != nil && now.Before(*coupon.StartsAt) {
		return &errors.BusinessRuleError{Message: errors.CouponNotYetStarted}
	}
	return nil
}

// validateUsageLimits checks global usage limit and per-phone usage limit.
func (s *CouponService) validateUsageLimits(ctx context.Context, coupon *models.Coupon, phone string) error {
	if coupon.UsageLimit != nil {
		count, err := s.couponRepository.CountUsages(ctx, coupon.ID)
		if err != nil {
			return err
		}
		if count >= *coupon.UsageLimit {
			return &errors.BusinessRuleError{Message: errors.CouponUsageLimitReached}
		}
	}

	if coupon.MaxUsesPerPhone != nil && phone != "" {
		count, err := s.couponRepository.CountUsagesByPhone(ctx, coupon.ID, phone)
		if err != nil {
			return err
		}
		if count >= *coupon.MaxUsesPerPhone {
			return &errors.BusinessRuleError{Message: errors.CouponPhoneLimitReached}
		}
	}

	return nil
}
