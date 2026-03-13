package requests

import (
	"strconv"
	"strings"
	"time"

	"github.com/mlgaray/ecommerce_api/internal/core/models"
	"github.com/mlgaray/ecommerce_api/internal/core/pagination"
	httpErrors "github.com/mlgaray/ecommerce_api/internal/infraestructure/adapters/http/errors"
)

// =============================================================================
// Shared coupon request fields (DRY: Create and Update have identical fields)
// =============================================================================

// couponRequestBase contains shared fields for Create and Update coupon requests.
type couponRequestBase struct {
	Code            string  `json:"code"`
	Type            string  `json:"type"`
	Value           float64 `json:"value"`
	MinOrderAmount  float64 `json:"min_order_amount,omitempty"`
	UsageLimit      *int    `json:"usage_limit,omitempty"`
	MaxUsesPerPhone *int    `json:"max_uses_per_phone,omitempty"`
	StartsAt        *string `json:"starts_at,omitempty"`
	ExpiresAt       *string `json:"expires_at,omitempty"`
	IsActive        bool    `json:"is_active"`
}

// validate validates HTTP input for coupon create/update.
func (r *couponRequestBase) validate() error {
	if strings.TrimSpace(r.Code) == "" {
		return &httpErrors.BadRequestError{Message: "coupon_code_is_required"}
	}
	if len(r.Code) > models.MaxCouponCodeLength {
		return &httpErrors.BadRequestError{Message: "coupon_code_exceeds_max_length"}
	}
	if r.Type != "percentage" && r.Type != "fixed" {
		return &httpErrors.BadRequestError{Message: "coupon_type_must_be_percentage_or_fixed"}
	}
	if err := validateNumericFields(r.Value, r.MinOrderAmount, r.UsageLimit, r.MaxUsesPerPhone); err != nil {
		return err
	}
	return validateDateFields(r.StartsAt, r.ExpiresAt)
}

// validateNumericFields validates value, min order amount, and optional limit fields.
func validateNumericFields(value float64, minOrderAmount float64, usageLimit, maxUsesPerPhone *int) error {
	if value <= 0 {
		return &httpErrors.BadRequestError{Message: "coupon_value_must_be_positive"}
	}
	if minOrderAmount < 0 {
		return &httpErrors.BadRequestError{Message: "min_order_amount_cannot_be_negative"}
	}
	if usageLimit != nil && *usageLimit <= 0 {
		return &httpErrors.BadRequestError{Message: "usage_limit_must_be_positive"}
	}
	if maxUsesPerPhone != nil && *maxUsesPerPhone <= 0 {
		return &httpErrors.BadRequestError{Message: "max_uses_per_phone_must_be_positive"}
	}
	return nil
}

// validateDateFields validates optional starts_at and expires_at RFC3339 date strings.
func validateDateFields(startsAt, expiresAt *string) error {
	if startsAt != nil {
		if _, err := time.Parse(time.RFC3339, *startsAt); err != nil {
			return &httpErrors.BadRequestError{Message: "invalid_starts_at_format"}
		}
	}
	if expiresAt != nil {
		if _, err := time.Parse(time.RFC3339, *expiresAt); err != nil {
			return &httpErrors.BadRequestError{Message: "invalid_expires_at_format"}
		}
	}
	return nil
}

// toModel converts the base request to a domain Coupon model.
func (r *couponRequestBase) toModel() *models.Coupon {
	coupon := &models.Coupon{
		Code:            strings.ToUpper(strings.TrimSpace(r.Code)),
		Type:            models.CouponType(r.Type),
		Value:           r.Value,
		MinOrderAmount:  r.MinOrderAmount,
		UsageLimit:      r.UsageLimit,
		MaxUsesPerPhone: r.MaxUsesPerPhone,
		IsActive:        r.IsActive,
	}

	if r.StartsAt != nil {
		t, _ := time.Parse(time.RFC3339, *r.StartsAt)
		coupon.StartsAt = &t
	}
	if r.ExpiresAt != nil {
		t, _ := time.Parse(time.RFC3339, *r.ExpiresAt)
		coupon.ExpiresAt = &t
	}

	return coupon
}

// =============================================================================
// Create
// =============================================================================

// CreateCouponRequest represents the HTTP request for creating a coupon.
type CreateCouponRequest struct {
	couponRequestBase
}

// Validate validates HTTP input for creating a coupon.
func (r *CreateCouponRequest) Validate() error { return r.validate() }

// ToModel converts the HTTP request to a domain model.
func (r *CreateCouponRequest) ToModel() *models.Coupon { return r.toModel() }

// =============================================================================
// Update
// =============================================================================

// UpdateCouponRequest represents the HTTP request for updating a coupon.
// Same fields as create — full replacement semantics.
type UpdateCouponRequest struct {
	couponRequestBase
}

// Validate validates HTTP input for updating a coupon.
func (r *UpdateCouponRequest) Validate() error { return r.validate() }

// ToModel converts the HTTP request to a domain model.
func (r *UpdateCouponRequest) ToModel() *models.Coupon { return r.toModel() }

// =============================================================================
// Validate (Public — storefront)
// =============================================================================

// ValidateCouponRequest represents the HTTP request for validating a coupon.
// Public endpoint — used by customers before placing an order.
type ValidateCouponRequest struct {
	Code     string  `json:"code"`
	Subtotal float64 `json:"subtotal"`
	Phone    string  `json:"phone,omitempty"`
}

// Validate validates HTTP input for coupon validation.
func (r *ValidateCouponRequest) Validate() error {
	if strings.TrimSpace(r.Code) == "" {
		return &httpErrors.BadRequestError{Message: "coupon_code_is_required"}
	}
	if r.Subtotal < 0 {
		return &httpErrors.BadRequestError{Message: "subtotal_cannot_be_negative"}
	}
	return nil
}

// =============================================================================
// Filters
// =============================================================================

// CouponFiltersRequest represents HTTP query parameters for coupon listing.
type CouponFiltersRequest struct {
	Search    *string
	IsActive  *string // "true" or "false" string from query param
	SortBy    string
	SortOrder string
	Limit     int
	Cursor    string
}

// NewCouponFiltersRequest parses query parameters from URL.
func NewCouponFiltersRequest(queryParams map[string][]string) (*CouponFiltersRequest, error) {
	request := &CouponFiltersRequest{}

	if search := getQueryParam(queryParams, "search"); search != "" {
		request.Search = &search
	}

	if isActive := getQueryParam(queryParams, "is_active"); isActive != "" {
		request.IsActive = &isActive
	}

	request.SortBy = getQueryParam(queryParams, "sort")
	request.SortOrder = getQueryParam(queryParams, "order")

	if limitStr := getQueryParam(queryParams, "limit"); limitStr != "" {
		limit, err := strconv.Atoi(limitStr)
		if err != nil {
			return nil, &httpErrors.BadRequestError{Message: "invalid_limit_format"}
		}
		request.Limit = limit
	}

	if cursorStr := getQueryParam(queryParams, "cursor"); cursorStr != "" {
		request.Cursor = cursorStr
	}

	return request, nil
}

// Validate validates HTTP-specific concerns.
func (r *CouponFiltersRequest) Validate() error {
	if r.Search != nil {
		trimmed := strings.TrimSpace(*r.Search)
		if trimmed == "" {
			return &httpErrors.BadRequestError{Message: "search_term_cannot_be_empty"}
		}
		*r.Search = trimmed
	}

	if r.IsActive != nil {
		if _, err := strconv.ParseBool(*r.IsActive); err != nil {
			return &httpErrors.BadRequestError{Message: "is_active_must_be_true_or_false"}
		}
	}

	if r.Limit < 0 {
		return &httpErrors.BadRequestError{Message: "limit_cannot_be_negative"}
	}

	return nil
}

// ToCouponFilters converts HTTP request to domain model.
func (r *CouponFiltersRequest) ToCouponFilters() models.CouponFilters {
	filters := models.CouponFilters{
		Search:    r.Search,
		SortBy:    r.SortBy,
		SortOrder: r.SortOrder,
		Limit:     r.Limit,
	}

	// Convert is_active string to bool pointer
	if r.IsActive != nil {
		isActive := *r.IsActive == "true"
		filters.IsActive = &isActive
	}

	// Decode cursor if present — return error for invalid cursor
	if r.Cursor != "" {
		cursorData, err := pagination.DecodeCursor(r.Cursor)
		if err != nil {
			return filters
		}
		filters.LastID = &cursorData.ID
		filters.LastSortValue = cursorData.SortValue
	}

	return filters
}
