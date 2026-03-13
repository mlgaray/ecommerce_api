package requests

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/mlgaray/ecommerce_api/internal/core/models"
	httpErrors "github.com/mlgaray/ecommerce_api/internal/infraestructure/adapters/http/errors"
)

// =============================================================================
// Test Helpers
// =============================================================================

func intPtr(v int) *int       { return &v }
func strPtr(v string) *string { return &v }

func newValidCreateCouponRequest() *CreateCouponRequest {
	return &CreateCouponRequest{
		couponRequestBase: couponRequestBase{
			Code:     "SUMMER20",
			Type:     "percentage",
			Value:    20,
			IsActive: true,
		},
	}
}

func assertBadRequestError(t *testing.T, err error, expectedMessage string) {
	t.Helper()
	var badReq *httpErrors.BadRequestError
	require.ErrorAs(t, err, &badReq)
	assert.Equal(t, expectedMessage, badReq.Message)
}

// =============================================================================
// CreateCouponRequest.Validate
// =============================================================================

func TestCreateCouponRequest_Validate(t *testing.T) {
	t.Run("when all fields valid then returns no error", func(t *testing.T) {
		req := newValidCreateCouponRequest()
		assert.NoError(t, req.Validate())
	})

	t.Run("when code is empty then returns error", func(t *testing.T) {
		req := newValidCreateCouponRequest()
		req.Code = ""
		assertBadRequestError(t, req.Validate(), "coupon_code_is_required")
	})

	t.Run("when code is whitespace then returns error", func(t *testing.T) {
		req := newValidCreateCouponRequest()
		req.Code = "   "
		assertBadRequestError(t, req.Validate(), "coupon_code_is_required")
	})

	t.Run("when code exceeds max length then returns error", func(t *testing.T) {
		req := newValidCreateCouponRequest()
		req.Code = strings.Repeat("A", models.MaxCouponCodeLength+1)
		assertBadRequestError(t, req.Validate(), "coupon_code_exceeds_max_length")
	})

	t.Run("when type is invalid then returns error", func(t *testing.T) {
		req := newValidCreateCouponRequest()
		req.Type = "invalid"
		assertBadRequestError(t, req.Validate(), "coupon_type_must_be_percentage_or_fixed")
	})

	t.Run("when type is fixed then passes", func(t *testing.T) {
		req := newValidCreateCouponRequest()
		req.Type = "fixed"
		assert.NoError(t, req.Validate())
	})

	t.Run("when value is zero then returns error", func(t *testing.T) {
		req := newValidCreateCouponRequest()
		req.Value = 0
		assertBadRequestError(t, req.Validate(), "coupon_value_must_be_positive")
	})

	t.Run("when value is negative then returns error", func(t *testing.T) {
		req := newValidCreateCouponRequest()
		req.Value = -5
		assertBadRequestError(t, req.Validate(), "coupon_value_must_be_positive")
	})

	t.Run("when min_order_amount is negative then returns error", func(t *testing.T) {
		req := newValidCreateCouponRequest()
		req.MinOrderAmount = -1
		assertBadRequestError(t, req.Validate(), "min_order_amount_cannot_be_negative")
	})

	t.Run("when min_order_amount is zero then passes", func(t *testing.T) {
		req := newValidCreateCouponRequest()
		req.MinOrderAmount = 0
		assert.NoError(t, req.Validate())
	})

	t.Run("when usage_limit is zero then returns error", func(t *testing.T) {
		req := newValidCreateCouponRequest()
		req.UsageLimit = intPtr(0)
		assertBadRequestError(t, req.Validate(), "usage_limit_must_be_positive")
	})

	t.Run("when usage_limit is negative then returns error", func(t *testing.T) {
		req := newValidCreateCouponRequest()
		req.UsageLimit = intPtr(-1)
		assertBadRequestError(t, req.Validate(), "usage_limit_must_be_positive")
	})

	t.Run("when usage_limit is nil then passes", func(t *testing.T) {
		req := newValidCreateCouponRequest()
		req.UsageLimit = nil
		assert.NoError(t, req.Validate())
	})

	t.Run("when max_uses_per_phone is zero then returns error", func(t *testing.T) {
		req := newValidCreateCouponRequest()
		req.MaxUsesPerPhone = intPtr(0)
		assertBadRequestError(t, req.Validate(), "max_uses_per_phone_must_be_positive")
	})

	t.Run("when max_uses_per_phone is nil then passes", func(t *testing.T) {
		req := newValidCreateCouponRequest()
		req.MaxUsesPerPhone = nil
		assert.NoError(t, req.Validate())
	})

	t.Run("when starts_at is invalid format then returns error", func(t *testing.T) {
		req := newValidCreateCouponRequest()
		req.StartsAt = strPtr("2024-01-01")
		assertBadRequestError(t, req.Validate(), "invalid_starts_at_format")
	})

	t.Run("when starts_at is valid RFC3339 then passes", func(t *testing.T) {
		req := newValidCreateCouponRequest()
		req.StartsAt = strPtr("2024-01-01T00:00:00Z")
		assert.NoError(t, req.Validate())
	})

	t.Run("when starts_at is nil then passes", func(t *testing.T) {
		req := newValidCreateCouponRequest()
		req.StartsAt = nil
		assert.NoError(t, req.Validate())
	})

	t.Run("when expires_at is invalid format then returns error", func(t *testing.T) {
		req := newValidCreateCouponRequest()
		req.ExpiresAt = strPtr("not-a-date")
		assertBadRequestError(t, req.Validate(), "invalid_expires_at_format")
	})

	t.Run("when expires_at is valid RFC3339 then passes", func(t *testing.T) {
		req := newValidCreateCouponRequest()
		req.ExpiresAt = strPtr("2024-12-31T23:59:59Z")
		assert.NoError(t, req.Validate())
	})
}

// =============================================================================
// CreateCouponRequest.ToModel
// =============================================================================

func TestCreateCouponRequest_ToModel(t *testing.T) {
	t.Run("when valid request then maps all fields correctly", func(t *testing.T) {
		req := &CreateCouponRequest{
			couponRequestBase: couponRequestBase{
				Code:            " summer20 ",
				Type:            "percentage",
				Value:           20,
				MinOrderAmount:  50,
				UsageLimit:      intPtr(100),
				MaxUsesPerPhone: intPtr(3),
				StartsAt:        strPtr("2024-01-01T00:00:00Z"),
				ExpiresAt:       strPtr("2024-12-31T23:59:59Z"),
				IsActive:        true,
			},
		}

		model := req.ToModel()

		assert.Equal(t, "SUMMER20", model.Code) // trimmed + uppercased
		assert.Equal(t, models.CouponTypePercentage, model.Type)
		assert.Equal(t, 20.0, model.Value)
		assert.Equal(t, 50.0, model.MinOrderAmount)
		assert.Equal(t, 100, *model.UsageLimit)
		assert.Equal(t, 3, *model.MaxUsesPerPhone)
		assert.True(t, model.IsActive)
		require.NotNil(t, model.StartsAt)
		require.NotNil(t, model.ExpiresAt)
		assert.Equal(t, 2024, model.StartsAt.Year())
		assert.Equal(t, 2024, model.ExpiresAt.Year())
	})

	t.Run("when optional date fields nil then model dates are nil", func(t *testing.T) {
		req := newValidCreateCouponRequest()

		model := req.ToModel()

		assert.Nil(t, model.StartsAt)
		assert.Nil(t, model.ExpiresAt)
	})
}

// =============================================================================
// UpdateCouponRequest
// =============================================================================

func TestUpdateCouponRequest_Validate(t *testing.T) {
	t.Run("when valid then returns no error", func(t *testing.T) {
		req := &UpdateCouponRequest{
			couponRequestBase: couponRequestBase{
				Code:  "UPDATE10",
				Type:  "fixed",
				Value: 10,
			},
		}
		assert.NoError(t, req.Validate())
	})
}

func TestUpdateCouponRequest_ToModel(t *testing.T) {
	t.Run("when valid then maps correctly", func(t *testing.T) {
		req := &UpdateCouponRequest{
			couponRequestBase: couponRequestBase{
				Code:  "update10",
				Type:  "fixed",
				Value: 10,
			},
		}

		model := req.ToModel()

		assert.Equal(t, "UPDATE10", model.Code)
		assert.Equal(t, models.CouponTypeFixed, model.Type)
	})
}

// =============================================================================
// ValidateCouponRequest.Validate
// =============================================================================

func TestValidateCouponRequest_Validate(t *testing.T) {
	t.Run("when all valid then passes", func(t *testing.T) {
		req := &ValidateCouponRequest{Code: "SUMMER20", Subtotal: 100}
		assert.NoError(t, req.Validate())
	})

	t.Run("when code empty then returns error", func(t *testing.T) {
		req := &ValidateCouponRequest{Code: "", Subtotal: 100}
		assertBadRequestError(t, req.Validate(), "coupon_code_is_required")
	})

	t.Run("when code whitespace then returns error", func(t *testing.T) {
		req := &ValidateCouponRequest{Code: "   ", Subtotal: 100}
		assertBadRequestError(t, req.Validate(), "coupon_code_is_required")
	})

	t.Run("when subtotal negative then returns error", func(t *testing.T) {
		req := &ValidateCouponRequest{Code: "ABC", Subtotal: -1}
		assertBadRequestError(t, req.Validate(), "subtotal_cannot_be_negative")
	})

	t.Run("when subtotal zero then passes", func(t *testing.T) {
		req := &ValidateCouponRequest{Code: "ABC", Subtotal: 0}
		assert.NoError(t, req.Validate())
	})
}

// =============================================================================
// NewCouponFiltersRequest
// =============================================================================

func TestNewCouponFiltersRequest(t *testing.T) {
	t.Run("when all params then parses correctly", func(t *testing.T) {
		params := map[string][]string{
			"search":    {"summer"},
			"is_active": {"true"},
			"sort":      {"code"},
			"order":     {"asc"},
			"limit":     {"25"},
			"cursor":    {"abc123"},
		}

		req, err := NewCouponFiltersRequest(params)

		require.NoError(t, err)
		assert.Equal(t, "summer", *req.Search)
		assert.Equal(t, "true", *req.IsActive)
		assert.Equal(t, "code", req.SortBy)
		assert.Equal(t, "asc", req.SortOrder)
		assert.Equal(t, 25, req.Limit)
		assert.Equal(t, "abc123", req.Cursor)
	})

	t.Run("when no params then returns defaults", func(t *testing.T) {
		req, err := NewCouponFiltersRequest(map[string][]string{})

		require.NoError(t, err)
		assert.Nil(t, req.Search)
		assert.Nil(t, req.IsActive)
		assert.Equal(t, 0, req.Limit)
		assert.Empty(t, req.Cursor)
	})

	t.Run("when invalid limit then returns error", func(t *testing.T) {
		params := map[string][]string{"limit": {"abc"}}

		_, err := NewCouponFiltersRequest(params)

		assertBadRequestError(t, err, "invalid_limit_format")
	})
}

// =============================================================================
// CouponFiltersRequest.Validate
// =============================================================================

func TestCouponFiltersRequest_Validate(t *testing.T) {
	t.Run("when all valid then passes", func(t *testing.T) {
		req := &CouponFiltersRequest{SortBy: "code", Limit: 10}
		assert.NoError(t, req.Validate())
	})

	t.Run("when search is empty string then returns error", func(t *testing.T) {
		search := "   "
		req := &CouponFiltersRequest{Search: &search}
		assertBadRequestError(t, req.Validate(), "search_term_cannot_be_empty")
	})

	t.Run("when is_active is invalid then returns error", func(t *testing.T) {
		isActive := "maybe"
		req := &CouponFiltersRequest{IsActive: &isActive}
		assertBadRequestError(t, req.Validate(), "is_active_must_be_true_or_false")
	})

	t.Run("when negative limit then returns error", func(t *testing.T) {
		req := &CouponFiltersRequest{Limit: -1}
		assertBadRequestError(t, req.Validate(), "limit_cannot_be_negative")
	})
}

// =============================================================================
// CouponFiltersRequest.ToCouponFilters
// =============================================================================

func TestCouponFiltersRequest_ToCouponFilters(t *testing.T) {
	t.Run("when is_active true then converts to bool pointer", func(t *testing.T) {
		isActive := "true"
		req := &CouponFiltersRequest{IsActive: &isActive}

		filters := req.ToCouponFilters()

		require.NotNil(t, filters.IsActive)
		assert.True(t, *filters.IsActive)
	})

	t.Run("when is_active false then converts to false", func(t *testing.T) {
		isActive := "false"
		req := &CouponFiltersRequest{IsActive: &isActive}

		filters := req.ToCouponFilters()

		require.NotNil(t, filters.IsActive)
		assert.False(t, *filters.IsActive)
	})

	t.Run("when no cursor then LastID is nil", func(t *testing.T) {
		req := &CouponFiltersRequest{}

		filters := req.ToCouponFilters()

		assert.Nil(t, filters.LastID)
		assert.Nil(t, filters.LastSortValue)
	})

	t.Run("when invalid cursor then treats as first page", func(t *testing.T) {
		req := &CouponFiltersRequest{Cursor: "invalid-cursor"}

		filters := req.ToCouponFilters()

		assert.Nil(t, filters.LastID)
	})

	t.Run("when search provided then maps to filters", func(t *testing.T) {
		search := "summer"
		req := &CouponFiltersRequest{Search: &search, SortBy: "code", SortOrder: "asc"}

		filters := req.ToCouponFilters()

		require.NotNil(t, filters.Search)
		assert.Equal(t, "summer", *filters.Search)
		assert.Equal(t, "code", filters.SortBy)
		assert.Equal(t, "asc", filters.SortOrder)
	})
}
