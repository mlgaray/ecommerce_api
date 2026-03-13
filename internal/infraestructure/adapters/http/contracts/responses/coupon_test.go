package responses

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/mlgaray/ecommerce_api/internal/core/models"
)

// =============================================================================
// CouponResponseFromModel
// =============================================================================

func TestCouponResponseFromModel(t *testing.T) {
	t.Run("when nil then returns nil", func(t *testing.T) {
		result := CouponResponseFromModel(nil)
		assert.Nil(t, result)
	})

	t.Run("when full model then maps all fields", func(t *testing.T) {
		usageLimit := 100
		maxUses := 3
		startsAt := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
		expiresAt := time.Date(2024, 12, 31, 23, 59, 59, 0, time.UTC)
		now := time.Now()

		coupon := &models.Coupon{
			ID:              1,
			Code:            "SUMMER20",
			Type:            models.CouponTypePercentage,
			Value:           20,
			MinOrderAmount:  50,
			UsageLimit:      &usageLimit,
			MaxUsesPerPhone: &maxUses,
			StartsAt:        &startsAt,
			ExpiresAt:       &expiresAt,
			IsActive:        true,
			UsageCount:      42,
			CreatedAt:       now,
			UpdatedAt:       now,
		}

		resp := CouponResponseFromModel(coupon)

		require.NotNil(t, resp)
		assert.Equal(t, 1, resp.ID)
		assert.Equal(t, "SUMMER20", resp.Code)
		assert.Equal(t, "percentage", resp.Type)
		assert.Equal(t, 20.0, resp.Value)
		assert.Equal(t, 50.0, resp.MinOrderAmount)
		assert.Equal(t, 100, *resp.UsageLimit)
		assert.Equal(t, 3, *resp.MaxUsesPerPhone)
		assert.Equal(t, startsAt, *resp.StartsAt)
		assert.Equal(t, expiresAt, *resp.ExpiresAt)
		assert.True(t, resp.IsActive)
		assert.Equal(t, 42, resp.UsageCount)
		assert.Equal(t, now, resp.CreatedAt)
		assert.Equal(t, now, resp.UpdatedAt)
	})

	t.Run("when optional fields nil then omits them", func(t *testing.T) {
		coupon := &models.Coupon{
			ID:   2,
			Code: "FIXED10",
			Type: models.CouponTypeFixed,
		}

		resp := CouponResponseFromModel(coupon)

		require.NotNil(t, resp)
		assert.Nil(t, resp.UsageLimit)
		assert.Nil(t, resp.MaxUsesPerPhone)
		assert.Nil(t, resp.StartsAt)
		assert.Nil(t, resp.ExpiresAt)
	})
}

// =============================================================================
// CouponResponsesFromModels
// =============================================================================

func TestCouponResponsesFromModels(t *testing.T) {
	t.Run("when nil then returns nil", func(t *testing.T) {
		result := CouponResponsesFromModels(nil)
		assert.Nil(t, result)
	})

	t.Run("when empty slice then returns empty slice", func(t *testing.T) {
		result := CouponResponsesFromModels([]*models.Coupon{})
		assert.NotNil(t, result)
		assert.Empty(t, result)
	})

	t.Run("when multiple coupons then maps all", func(t *testing.T) {
		coupons := []*models.Coupon{
			{ID: 1, Code: "A"},
			{ID: 2, Code: "B"},
		}

		result := CouponResponsesFromModels(coupons)

		assert.Len(t, result, 2)
		assert.Equal(t, "A", result[0].Code)
		assert.Equal(t, "B", result[1].Code)
	})
}

// =============================================================================
// NewListCouponsResponse
// =============================================================================

func TestNewListCouponsResponse(t *testing.T) {
	t.Run("when coupons with pagination then maps all fields", func(t *testing.T) {
		totalCount := 50
		coupons := []*models.Coupon{{ID: 1, Code: "A"}}

		resp := NewListCouponsResponse(coupons, "next_cursor_abc", true, &totalCount)

		assert.Len(t, resp.Coupons, 1)
		assert.Equal(t, "next_cursor_abc", resp.NextCursor)
		assert.True(t, resp.HasMore)
		assert.Equal(t, 50, *resp.TotalCount)
	})

	t.Run("when no next cursor then next_cursor is empty", func(t *testing.T) {
		resp := NewListCouponsResponse([]*models.Coupon{{ID: 1}}, "", false, nil)

		assert.Empty(t, resp.NextCursor)
		assert.False(t, resp.HasMore)
		assert.Nil(t, resp.TotalCount)
	})

	t.Run("when nil coupons then coupons is nil", func(t *testing.T) {
		resp := NewListCouponsResponse(nil, "", false, nil)

		assert.Nil(t, resp.Coupons)
	})
}
