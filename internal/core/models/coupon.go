package models

import (
	"math"
	"strings"
	"time"

	"github.com/mlgaray/ecommerce_api/internal/core/errors"
)

// CouponType represents the type of discount a coupon applies
type CouponType string

const (
	CouponTypePercentage CouponType = "percentage"
	CouponTypeFixed      CouponType = "fixed"
)

// MaxCouponCodeLength is the maximum allowed length for a coupon code.
const MaxCouponCodeLength = 50

// Coupon represents a discount coupon for a store
type Coupon struct {
	ID              int        `json:"id,omitempty"`
	ShopID          int        `json:"shop_id,omitempty"`
	Code            string     `json:"code,omitempty"`
	Type            CouponType `json:"type,omitempty"`
	Value           float64    `json:"value,omitempty"`
	MinOrderAmount  float64    `json:"min_order_amount,omitempty"`
	UsageLimit      *int       `json:"usage_limit,omitempty"`
	MaxUsesPerPhone *int       `json:"max_uses_per_phone,omitempty"`
	StartsAt        *time.Time `json:"starts_at,omitempty"`
	ExpiresAt       *time.Time `json:"expires_at,omitempty"`
	IsActive        bool       `json:"is_active"`
	DiscountAmount  float64    `json:"discount_amount,omitempty"`
	UsageCount      int        `json:"usage_count"`
	CreatedAt       time.Time  `json:"created_at,omitempty"`
	UpdatedAt       time.Time  `json:"updated_at,omitempty"`
}

// GetID implements Identifiable interface for pagination
func (c *Coupon) GetID() int {
	return c.ID
}

// GetSortValue implements Sortable interface for pagination
func (c *Coupon) GetSortValue(sortBy string) interface{} {
	switch sortBy {
	case CouponSortByCode:
		return c.Code
	case SortByCreatedAt:
		return c.CreatedAt
	default:
		return c.ID
	}
}

// Validate validates business rules for the Coupon domain model.
// Called by CouponService before Create and Update operations.
// Normalizes the code (trim + uppercase) as part of validation.
func (c *Coupon) Validate() error {
	c.Code = strings.ToUpper(strings.TrimSpace(c.Code))

	validators := []func() error{
		c.validateCode,
		c.validateType,
		c.validateValue,
		c.validateLimits,
		c.validateDates,
	}

	for _, validate := range validators {
		if err := validate(); err != nil {
			return err
		}
	}

	return nil
}

// validateCode checks that the coupon code is present and within length limits.
func (c *Coupon) validateCode() error {
	if c.Code == "" {
		return &errors.ValidationError{Message: errors.CouponCodeIsRequired}
	}

	if len(c.Code) > MaxCouponCodeLength {
		return &errors.ValidationError{Message: errors.CouponCodeTooLong}
	}

	return nil
}

// validateType checks that the coupon type is a valid enum value.
func (c *Coupon) validateType() error {
	if c.Type != CouponTypePercentage && c.Type != CouponTypeFixed {
		return &errors.ValidationError{Message: errors.CouponTypeInvalid}
	}

	return nil
}

// validateValue checks that the coupon value is positive and within type-specific bounds.
func (c *Coupon) validateValue() error {
	if c.Value <= 0 {
		return &errors.ValidationError{Message: errors.CouponValueMustBePositive}
	}

	if c.Type == CouponTypePercentage && c.Value > 100 {
		return &errors.ValidationError{Message: errors.CouponPercentageExceedsMax}
	}

	if c.MinOrderAmount < 0 {
		return &errors.ValidationError{Message: errors.CouponMinOrderCannotBeNegative}
	}

	return nil
}

// validateLimits checks that usage limit and per-phone limit are positive when set.
func (c *Coupon) validateLimits() error {
	if c.UsageLimit != nil && *c.UsageLimit <= 0 {
		return &errors.ValidationError{Message: errors.CouponUsageLimitMustBePositive}
	}

	if c.MaxUsesPerPhone != nil && *c.MaxUsesPerPhone <= 0 {
		return &errors.ValidationError{Message: errors.CouponPhoneLimitMustBePositive}
	}

	return nil
}

// validateDates checks that expires_at is after starts_at when both are set.
func (c *Coupon) validateDates() error {
	if c.StartsAt != nil && c.ExpiresAt != nil && !c.ExpiresAt.After(*c.StartsAt) {
		return &errors.ValidationError{Message: errors.CouponExpiresBeforeStarts}
	}

	return nil
}

// CalculateDiscount calculates the discount amount for a given subtotal.
// For percentage coupons: rounds to 2 decimal places.
// For fixed coupons: uses the value directly.
// Result is capped at the subtotal (discount cannot exceed order value).
func (c *Coupon) CalculateDiscount(subtotal float64) float64 {
	var discount float64
	switch c.Type {
	case CouponTypePercentage:
		discount = math.Round(subtotal*c.Value/100*100) / 100
	case CouponTypeFixed:
		discount = c.Value
	}

	if discount > subtotal {
		discount = subtotal
	}

	return discount
}

// ============================================================================
// CouponFilters - Query criteria for coupon listing
// ============================================================================

// Valid sort fields for coupons
const (
	CouponSortByCode = "code"
)

// CouponFilters represents search and filter criteria for coupons.
// Note: ShopID is NOT included here - it's a context parameter passed separately.
type CouponFilters struct {
	// Optional search filter (searches by code)
	Search *string

	// Optional active status filter
	IsActive *bool

	// Sorting
	SortBy    string // "code", "created_at" (default: "created_at")
	SortOrder string // "asc", "desc" (default: "desc")

	// Pagination (cursor-based)
	Limit         int         // Number of items per page (default: 20, max: 100)
	LastID        *int        // ID of the last item from previous page (nil = first page)
	LastSortValue interface{} // Value of the sort field from last item
}

// Validated validates business rules and returns a normalized copy of CouponFilters.
// This is an immutable operation - the original struct is not modified.
func (f CouponFilters) Validated() (CouponFilters, error) {
	result := f

	result.normalizeLimit()

	if err := result.validateAndNormalizeSorting(); err != nil {
		return CouponFilters{}, err
	}

	return result, nil
}

// normalizeLimit ensures limit is within valid bounds
func (f *CouponFilters) normalizeLimit() {
	if f.Limit <= 0 {
		f.Limit = DefaultLimit
	}
	if f.Limit > MaxLimit {
		f.Limit = MaxLimit
	}
}

// validateAndNormalizeSorting validates sort field and order, sets defaults
func (f *CouponFilters) validateAndNormalizeSorting() error {
	validSortFields := map[string]bool{
		CouponSortByCode: true,
		SortByCreatedAt:  true,
		"":               true,
	}

	if !validSortFields[f.SortBy] {
		return &errors.ValidationError{
			Message: errors.InvalidSortField,
		}
	}

	if f.SortBy == "" {
		f.SortBy = SortByCreatedAt
	}

	if f.SortOrder != SortOrderAsc && f.SortOrder != SortOrderDesc && f.SortOrder != "" {
		return &errors.ValidationError{
			Message: errors.InvalidSortOrder,
		}
	}

	if f.SortOrder == "" {
		f.SortOrder = SortOrderDesc
	}

	return nil
}
