package models

import (
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/mlgaray/ecommerce_api/internal/core/errors"
)

// =============================================================================
// Coupon.GetID Tests
// =============================================================================

func TestCoupon_GetID(t *testing.T) {
	t.Run("when coupon has ID then returns ID", func(t *testing.T) {
		// Arrange
		coupon := &Coupon{ID: 42, Code: "TEST"}

		// Act
		result := coupon.GetID()

		// Assert
		assert.Equal(t, 42, result)
	})

	t.Run("when coupon has zero ID then returns zero", func(t *testing.T) {
		// Arrange
		coupon := &Coupon{ID: 0, Code: "TEST"}

		// Act
		result := coupon.GetID()

		// Assert
		assert.Equal(t, 0, result)
	})
}

// =============================================================================
// Coupon.GetSortValue Tests
// =============================================================================

func TestCoupon_GetSortValue(t *testing.T) {
	t.Run("when sort by code then returns code", func(t *testing.T) {
		// Arrange
		coupon := &Coupon{ID: 1, Code: "SUMMER20", CreatedAt: time.Now()}

		// Act
		result := coupon.GetSortValue(CouponSortByCode)

		// Assert
		assert.Equal(t, "SUMMER20", result)
	})

	t.Run("when sort by created_at then returns created_at", func(t *testing.T) {
		// Arrange
		createdAt := time.Date(2024, 6, 15, 10, 0, 0, 0, time.UTC)
		coupon := &Coupon{ID: 1, Code: "SUMMER20", CreatedAt: createdAt}

		// Act
		result := coupon.GetSortValue(SortByCreatedAt)

		// Assert
		assert.Equal(t, createdAt, result)
	})

	t.Run("when sort by unknown field then returns ID", func(t *testing.T) {
		// Arrange
		coupon := &Coupon{ID: 42, Code: "SUMMER20"}

		// Act
		result := coupon.GetSortValue("unknown_field")

		// Assert
		assert.Equal(t, 42, result)
	})

	t.Run("when sort by empty string then returns ID", func(t *testing.T) {
		// Arrange
		coupon := &Coupon{ID: 42, Code: "SUMMER20"}

		// Act
		result := coupon.GetSortValue("")

		// Assert
		assert.Equal(t, 42, result)
	})
}

// =============================================================================
// Coupon.Validate Tests
// =============================================================================

func TestCoupon_Validate(t *testing.T) {
	t.Run("when valid percentage coupon then returns nil", func(t *testing.T) {
		// Arrange
		coupon := &Coupon{
			Code:  "SUMMER20",
			Type:  CouponTypePercentage,
			Value: 20,
		}

		// Act
		err := coupon.Validate()

		// Assert
		assert.NoError(t, err)
	})

	t.Run("when valid fixed coupon then returns nil", func(t *testing.T) {
		// Arrange
		coupon := &Coupon{
			Code:  "FLAT500",
			Type:  CouponTypeFixed,
			Value: 500,
		}

		// Act
		err := coupon.Validate()

		// Assert
		assert.NoError(t, err)
	})

	t.Run("when valid coupon with dates then returns nil", func(t *testing.T) {
		// Arrange
		startsAt := time.Now()
		expiresAt := startsAt.Add(24 * time.Hour)
		coupon := &Coupon{
			Code:      "PROMO",
			Type:      CouponTypePercentage,
			Value:     10,
			StartsAt:  &startsAt,
			ExpiresAt: &expiresAt,
		}

		// Act
		err := coupon.Validate()

		// Assert
		assert.NoError(t, err)
	})

	t.Run("when code is empty then returns validation error", func(t *testing.T) {
		// Arrange
		coupon := &Coupon{
			Code:  "",
			Type:  CouponTypePercentage,
			Value: 10,
		}

		// Act
		err := coupon.Validate()

		// Assert
		var validationErr *errors.ValidationError
		assert.ErrorAs(t, err, &validationErr)
		assert.Equal(t, errors.CouponCodeIsRequired, validationErr.Message)
	})

	t.Run("when code is whitespace only then returns validation error", func(t *testing.T) {
		// Arrange
		coupon := &Coupon{
			Code:  "   ",
			Type:  CouponTypePercentage,
			Value: 10,
		}

		// Act
		err := coupon.Validate()

		// Assert
		var validationErr *errors.ValidationError
		assert.ErrorAs(t, err, &validationErr)
		assert.Equal(t, errors.CouponCodeIsRequired, validationErr.Message)
	})

	t.Run("when type is invalid then returns validation error", func(t *testing.T) {
		// Arrange
		coupon := &Coupon{
			Code:  "TEST",
			Type:  "invalid_type",
			Value: 10,
		}

		// Act
		err := coupon.Validate()

		// Assert
		var validationErr *errors.ValidationError
		assert.ErrorAs(t, err, &validationErr)
		assert.Equal(t, errors.CouponTypeInvalid, validationErr.Message)
	})

	t.Run("when type is empty then returns validation error", func(t *testing.T) {
		// Arrange
		coupon := &Coupon{
			Code:  "TEST",
			Type:  "",
			Value: 10,
		}

		// Act
		err := coupon.Validate()

		// Assert
		var validationErr *errors.ValidationError
		assert.ErrorAs(t, err, &validationErr)
		assert.Equal(t, errors.CouponTypeInvalid, validationErr.Message)
	})

	t.Run("when value is zero then returns validation error", func(t *testing.T) {
		// Arrange
		coupon := &Coupon{
			Code:  "TEST",
			Type:  CouponTypePercentage,
			Value: 0,
		}

		// Act
		err := coupon.Validate()

		// Assert
		var validationErr *errors.ValidationError
		assert.ErrorAs(t, err, &validationErr)
		assert.Equal(t, errors.CouponValueMustBePositive, validationErr.Message)
	})

	t.Run("when value is negative then returns validation error", func(t *testing.T) {
		// Arrange
		coupon := &Coupon{
			Code:  "TEST",
			Type:  CouponTypeFixed,
			Value: -10,
		}

		// Act
		err := coupon.Validate()

		// Assert
		var validationErr *errors.ValidationError
		assert.ErrorAs(t, err, &validationErr)
		assert.Equal(t, errors.CouponValueMustBePositive, validationErr.Message)
	})

	t.Run("when percentage exceeds 100 then returns validation error", func(t *testing.T) {
		// Arrange
		coupon := &Coupon{
			Code:  "TEST",
			Type:  CouponTypePercentage,
			Value: 101,
		}

		// Act
		err := coupon.Validate()

		// Assert
		var validationErr *errors.ValidationError
		assert.ErrorAs(t, err, &validationErr)
		assert.Equal(t, errors.CouponPercentageExceedsMax, validationErr.Message)
	})

	t.Run("when percentage is exactly 100 then returns nil", func(t *testing.T) {
		// Arrange
		coupon := &Coupon{
			Code:  "FREE",
			Type:  CouponTypePercentage,
			Value: 100,
		}

		// Act
		err := coupon.Validate()

		// Assert
		assert.NoError(t, err)
	})

	t.Run("when fixed value exceeds 100 then returns nil", func(t *testing.T) {
		// Arrange — fixed coupons have no max value constraint
		coupon := &Coupon{
			Code:  "FLAT500",
			Type:  CouponTypeFixed,
			Value: 500,
		}

		// Act
		err := coupon.Validate()

		// Assert
		assert.NoError(t, err)
	})

	t.Run("when min_order_amount is negative then returns validation error", func(t *testing.T) {
		// Arrange
		coupon := &Coupon{
			Code:           "TEST",
			Type:           CouponTypeFixed,
			Value:          100,
			MinOrderAmount: -1,
		}

		// Act
		err := coupon.Validate()

		// Assert
		var validationErr *errors.ValidationError
		assert.ErrorAs(t, err, &validationErr)
		assert.Equal(t, errors.CouponMinOrderCannotBeNegative, validationErr.Message)
	})

	t.Run("when min_order_amount is zero then returns nil", func(t *testing.T) {
		// Arrange
		coupon := &Coupon{
			Code:           "TEST",
			Type:           CouponTypeFixed,
			Value:          100,
			MinOrderAmount: 0,
		}

		// Act
		err := coupon.Validate()

		// Assert
		assert.NoError(t, err)
	})

	t.Run("when usage_limit is zero then returns validation error", func(t *testing.T) {
		// Arrange
		zero := 0
		coupon := &Coupon{
			Code:       "TEST",
			Type:       CouponTypeFixed,
			Value:      100,
			UsageLimit: &zero,
		}

		// Act
		err := coupon.Validate()

		// Assert
		var validationErr *errors.ValidationError
		assert.ErrorAs(t, err, &validationErr)
		assert.Equal(t, errors.CouponUsageLimitMustBePositive, validationErr.Message)
	})

	t.Run("when usage_limit is negative then returns validation error", func(t *testing.T) {
		// Arrange
		neg := -5
		coupon := &Coupon{
			Code:       "TEST",
			Type:       CouponTypeFixed,
			Value:      100,
			UsageLimit: &neg,
		}

		// Act
		err := coupon.Validate()

		// Assert
		var validationErr *errors.ValidationError
		assert.ErrorAs(t, err, &validationErr)
		assert.Equal(t, errors.CouponUsageLimitMustBePositive, validationErr.Message)
	})

	t.Run("when usage_limit is nil then returns nil", func(t *testing.T) {
		// Arrange
		coupon := &Coupon{
			Code:       "TEST",
			Type:       CouponTypeFixed,
			Value:      100,
			UsageLimit: nil,
		}

		// Act
		err := coupon.Validate()

		// Assert
		assert.NoError(t, err)
	})

	t.Run("when max_uses_per_phone is zero then returns validation error", func(t *testing.T) {
		// Arrange
		zero := 0
		coupon := &Coupon{
			Code:            "TEST",
			Type:            CouponTypeFixed,
			Value:           100,
			MaxUsesPerPhone: &zero,
		}

		// Act
		err := coupon.Validate()

		// Assert
		var validationErr *errors.ValidationError
		assert.ErrorAs(t, err, &validationErr)
		assert.Equal(t, errors.CouponPhoneLimitMustBePositive, validationErr.Message)
	})

	t.Run("when max_uses_per_phone is negative then returns validation error", func(t *testing.T) {
		// Arrange
		neg := -1
		coupon := &Coupon{
			Code:            "TEST",
			Type:            CouponTypeFixed,
			Value:           100,
			MaxUsesPerPhone: &neg,
		}

		// Act
		err := coupon.Validate()

		// Assert
		var validationErr *errors.ValidationError
		assert.ErrorAs(t, err, &validationErr)
		assert.Equal(t, errors.CouponPhoneLimitMustBePositive, validationErr.Message)
	})

	t.Run("when expires_at equals starts_at then returns validation error", func(t *testing.T) {
		// Arrange — zero-duration window is not useful
		sameTime := time.Now()
		coupon := &Coupon{
			Code:      "TEST",
			Type:      CouponTypePercentage,
			Value:     10,
			StartsAt:  &sameTime,
			ExpiresAt: &sameTime,
		}

		// Act
		err := coupon.Validate()

		// Assert
		var validationErr *errors.ValidationError
		assert.ErrorAs(t, err, &validationErr)
		assert.Equal(t, errors.CouponExpiresBeforeStarts, validationErr.Message)
	})

	t.Run("when expires_at is before starts_at then returns validation error", func(t *testing.T) {
		// Arrange
		startsAt := time.Now().Add(24 * time.Hour)
		expiresAt := time.Now().Add(-24 * time.Hour)
		coupon := &Coupon{
			Code:      "TEST",
			Type:      CouponTypePercentage,
			Value:     10,
			StartsAt:  &startsAt,
			ExpiresAt: &expiresAt,
		}

		// Act
		err := coupon.Validate()

		// Assert
		var validationErr *errors.ValidationError
		assert.ErrorAs(t, err, &validationErr)
		assert.Equal(t, errors.CouponExpiresBeforeStarts, validationErr.Message)
	})

	t.Run("when only starts_at is set then returns nil", func(t *testing.T) {
		// Arrange
		startsAt := time.Now()
		coupon := &Coupon{
			Code:     "TEST",
			Type:     CouponTypePercentage,
			Value:    10,
			StartsAt: &startsAt,
		}

		// Act
		err := coupon.Validate()

		// Assert
		assert.NoError(t, err)
	})

	t.Run("when only expires_at is set then returns nil", func(t *testing.T) {
		// Arrange
		expiresAt := time.Now().Add(24 * time.Hour)
		coupon := &Coupon{
			Code:      "TEST",
			Type:      CouponTypePercentage,
			Value:     10,
			ExpiresAt: &expiresAt,
		}

		// Act
		err := coupon.Validate()

		// Assert
		assert.NoError(t, err)
	})

	t.Run("when code exceeds max length then returns validation error", func(t *testing.T) {
		// Arrange
		coupon := &Coupon{
			Code:  strings.Repeat("A", MaxCouponCodeLength+1),
			Type:  CouponTypePercentage,
			Value: 10,
		}

		// Act
		err := coupon.Validate()

		// Assert
		var validationErr *errors.ValidationError
		assert.ErrorAs(t, err, &validationErr)
		assert.Equal(t, errors.CouponCodeTooLong, validationErr.Message)
	})

	t.Run("when code has spaces and lowercase then normalizes", func(t *testing.T) {
		// Arrange
		coupon := &Coupon{
			Code:  "  summer20  ",
			Type:  CouponTypePercentage,
			Value: 10,
		}

		// Act
		err := coupon.Validate()

		// Assert
		assert.NoError(t, err)
		assert.Equal(t, "SUMMER20", coupon.Code)
	})
}

// =============================================================================
// Coupon.CalculateDiscount Tests
// =============================================================================

func TestCoupon_CalculateDiscount(t *testing.T) {
	t.Run("when percentage coupon then calculates percentage of subtotal", func(t *testing.T) {
		// Arrange
		coupon := &Coupon{Type: CouponTypePercentage, Value: 10}

		// Act
		result := coupon.CalculateDiscount(1000)

		// Assert
		assert.Equal(t, 100.0, result)
	})

	t.Run("when fixed coupon then returns fixed value", func(t *testing.T) {
		// Arrange
		coupon := &Coupon{Type: CouponTypeFixed, Value: 250}

		// Act
		result := coupon.CalculateDiscount(1000)

		// Assert
		assert.Equal(t, 250.0, result)
	})

	t.Run("when fixed discount exceeds subtotal then caps at subtotal", func(t *testing.T) {
		// Arrange
		coupon := &Coupon{Type: CouponTypeFixed, Value: 500}

		// Act
		result := coupon.CalculateDiscount(200)

		// Assert
		assert.Equal(t, 200.0, result)
	})

	t.Run("when 100 percent coupon then discount equals subtotal", func(t *testing.T) {
		// Arrange
		coupon := &Coupon{Type: CouponTypePercentage, Value: 100}

		// Act
		result := coupon.CalculateDiscount(350)

		// Assert
		assert.Equal(t, 350.0, result)
	})

	t.Run("when subtotal is zero then discount is zero", func(t *testing.T) {
		// Arrange
		coupon := &Coupon{Type: CouponTypePercentage, Value: 20}

		// Act
		result := coupon.CalculateDiscount(0)

		// Assert
		assert.Equal(t, 0.0, result)
	})
}

// =============================================================================
// CouponFilters.Validated Tests
// =============================================================================

func TestCouponFilters_Validated(t *testing.T) {
	t.Run("when all defaults then returns normalized filters", func(t *testing.T) {
		// Arrange
		filters := CouponFilters{}

		// Act
		result, err := filters.Validated()

		// Assert
		assert.NoError(t, err)
		assert.Equal(t, DefaultLimit, result.Limit)
		assert.Equal(t, SortByCreatedAt, result.SortBy)
		assert.Equal(t, SortOrderDesc, result.SortOrder)
	})

	t.Run("when valid filters then preserves values", func(t *testing.T) {
		// Arrange
		search := "SUMMER"
		isActive := true
		lastID := 10
		filters := CouponFilters{
			Search:    &search,
			IsActive:  &isActive,
			SortBy:    CouponSortByCode,
			SortOrder: SortOrderAsc,
			Limit:     50,
			LastID:    &lastID,
		}

		// Act
		result, err := filters.Validated()

		// Assert
		assert.NoError(t, err)
		assert.Equal(t, &search, result.Search)
		assert.Equal(t, &isActive, result.IsActive)
		assert.Equal(t, CouponSortByCode, result.SortBy)
		assert.Equal(t, SortOrderAsc, result.SortOrder)
		assert.Equal(t, 50, result.Limit)
		assert.Equal(t, &lastID, result.LastID)
	})

	t.Run("when original is not modified then maintains immutability", func(t *testing.T) {
		// Arrange
		filters := CouponFilters{
			SortBy:    "",
			SortOrder: "",
			Limit:     0,
		}

		// Act
		result, err := filters.Validated()

		// Assert
		assert.NoError(t, err)
		// Original should be unchanged
		assert.Equal(t, "", filters.SortBy)
		assert.Equal(t, "", filters.SortOrder)
		assert.Equal(t, 0, filters.Limit)
		// Result should have defaults
		assert.Equal(t, SortByCreatedAt, result.SortBy)
		assert.Equal(t, SortOrderDesc, result.SortOrder)
		assert.Equal(t, DefaultLimit, result.Limit)
	})
}

// =============================================================================
// CouponFilters Limit Normalization Tests
// =============================================================================

func TestCouponFilters_Validated_Limit(t *testing.T) {
	t.Run("when limit is zero then uses default", func(t *testing.T) {
		// Arrange
		filters := CouponFilters{Limit: 0}

		// Act
		result, err := filters.Validated()

		// Assert
		assert.NoError(t, err)
		assert.Equal(t, DefaultLimit, result.Limit)
	})

	t.Run("when limit is negative then uses default", func(t *testing.T) {
		// Arrange
		filters := CouponFilters{Limit: -10}

		// Act
		result, err := filters.Validated()

		// Assert
		assert.NoError(t, err)
		assert.Equal(t, DefaultLimit, result.Limit)
	})

	t.Run("when limit exceeds max then uses max", func(t *testing.T) {
		// Arrange
		filters := CouponFilters{Limit: 200}

		// Act
		result, err := filters.Validated()

		// Assert
		assert.NoError(t, err)
		assert.Equal(t, MaxLimit, result.Limit)
	})

	t.Run("when limit is within bounds then uses provided value", func(t *testing.T) {
		// Arrange
		filters := CouponFilters{Limit: 50}

		// Act
		result, err := filters.Validated()

		// Assert
		assert.NoError(t, err)
		assert.Equal(t, 50, result.Limit)
	})
}

// =============================================================================
// CouponFilters Sort Validation Tests
// =============================================================================

func TestCouponFilters_Validated_SortBy(t *testing.T) {
	t.Run("when sort_by is empty then uses default", func(t *testing.T) {
		// Arrange
		filters := CouponFilters{SortBy: ""}

		// Act
		result, err := filters.Validated()

		// Assert
		assert.NoError(t, err)
		assert.Equal(t, SortByCreatedAt, result.SortBy)
	})

	t.Run("when sort_by is code then uses code", func(t *testing.T) {
		// Arrange
		filters := CouponFilters{SortBy: CouponSortByCode}

		// Act
		result, err := filters.Validated()

		// Assert
		assert.NoError(t, err)
		assert.Equal(t, CouponSortByCode, result.SortBy)
	})

	t.Run("when sort_by is invalid then returns validation error", func(t *testing.T) {
		// Arrange
		filters := CouponFilters{SortBy: "invalid_field"}

		// Act
		result, err := filters.Validated()

		// Assert
		assert.Error(t, err)
		var validationErr *errors.ValidationError
		assert.ErrorAs(t, err, &validationErr)
		assert.Equal(t, errors.InvalidSortField, validationErr.Message)
		assert.Equal(t, CouponFilters{}, result)
	})
}

func TestCouponFilters_Validated_SortOrder(t *testing.T) {
	t.Run("when sort_order is empty then uses default desc", func(t *testing.T) {
		// Arrange
		filters := CouponFilters{SortOrder: ""}

		// Act
		result, err := filters.Validated()

		// Assert
		assert.NoError(t, err)
		assert.Equal(t, SortOrderDesc, result.SortOrder)
	})

	t.Run("when sort_order is asc then uses asc", func(t *testing.T) {
		// Arrange
		filters := CouponFilters{SortOrder: SortOrderAsc}

		// Act
		result, err := filters.Validated()

		// Assert
		assert.NoError(t, err)
		assert.Equal(t, SortOrderAsc, result.SortOrder)
	})

	t.Run("when sort_order is invalid then returns validation error", func(t *testing.T) {
		// Arrange
		filters := CouponFilters{SortOrder: "invalid"}

		// Act
		result, err := filters.Validated()

		// Assert
		assert.Error(t, err)
		var validationErr *errors.ValidationError
		assert.ErrorAs(t, err, &validationErr)
		assert.Equal(t, errors.InvalidSortOrder, validationErr.Message)
		assert.Equal(t, CouponFilters{}, result)
	})
}
