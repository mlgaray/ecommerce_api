package responses

import (
	"time"

	"github.com/mlgaray/ecommerce_api/internal/core/models"
)

// CouponResponse represents a coupon in HTTP responses.
type CouponResponse struct {
	ID              int        `json:"id"`
	Code            string     `json:"code"`
	Type            string     `json:"type"`
	Value           float64    `json:"value"`
	MinOrderAmount  float64    `json:"min_order_amount"`
	UsageLimit      *int       `json:"usage_limit,omitempty"`
	MaxUsesPerPhone *int       `json:"max_uses_per_phone,omitempty"`
	StartsAt        *time.Time `json:"starts_at,omitempty"`
	ExpiresAt       *time.Time `json:"expires_at,omitempty"`
	IsActive        bool       `json:"is_active"`
	UsageCount      int        `json:"usage_count"`
	CreatedAt       time.Time  `json:"created_at"`
	UpdatedAt       time.Time  `json:"updated_at"`
}

// CouponResponseFromModel converts a domain Coupon to a CouponResponse.
func CouponResponseFromModel(c *models.Coupon) *CouponResponse {
	if c == nil {
		return nil
	}

	return &CouponResponse{
		ID:              c.ID,
		Code:            c.Code,
		Type:            string(c.Type),
		Value:           c.Value,
		MinOrderAmount:  c.MinOrderAmount,
		UsageLimit:      c.UsageLimit,
		MaxUsesPerPhone: c.MaxUsesPerPhone,
		StartsAt:        c.StartsAt,
		ExpiresAt:       c.ExpiresAt,
		IsActive:        c.IsActive,
		UsageCount:      c.UsageCount,
		CreatedAt:       c.CreatedAt,
		UpdatedAt:       c.UpdatedAt,
	}
}

// CouponResponsesFromModels converts a slice of domain Coupons to CouponResponses.
func CouponResponsesFromModels(coupons []*models.Coupon) []*CouponResponse {
	if coupons == nil {
		return nil
	}

	responses := make([]*CouponResponse, len(coupons))
	for i, c := range coupons {
		responses[i] = CouponResponseFromModel(c)
	}
	return responses
}

// CreateCouponResponse represents the response for coupon creation.
type CreateCouponResponse struct {
	Coupon *CouponResponse `json:"coupon"`
}

// GetCouponResponse represents the response for coupon detail retrieval.
type GetCouponResponse struct {
	Coupon *CouponResponse `json:"coupon"`
}

// UpdateCouponResponse represents the response for coupon update.
type UpdateCouponResponse struct {
	Coupon *CouponResponse `json:"coupon"`
}

// ListCouponsResponse represents the HTTP response for list coupons.
type ListCouponsResponse struct {
	Coupons    []*CouponResponse `json:"coupons"`
	NextCursor string            `json:"next_cursor,omitempty"`
	HasMore    bool              `json:"has_more"`
	TotalCount *int              `json:"total_count,omitempty"`
}

// NewListCouponsResponse creates a ListCouponsResponse from domain models.
func NewListCouponsResponse(
	coupons []*models.Coupon,
	nextCursor string,
	hasMore bool,
	totalCount *int,
) *ListCouponsResponse {
	return &ListCouponsResponse{
		Coupons:    CouponResponsesFromModels(coupons),
		NextCursor: nextCursor,
		HasMore:    hasMore,
		TotalCount: totalCount,
	}
}

// ValidateCouponResponse represents the response for a valid coupon.
// Errors follow the standard API error pattern via HandleError.
type ValidateCouponResponse struct {
	CouponType     string  `json:"coupon_type"`
	CouponValue    float64 `json:"coupon_value"`
	DiscountAmount float64 `json:"discount_amount"`
}

// OrderCouponResponse represents coupon data within an order response.
// Lightweight — only the fields relevant to the order context.
type OrderCouponResponse struct {
	Code           string  `json:"code"`
	Type           string  `json:"type"`
	Value          float64 `json:"value"`
	DiscountAmount float64 `json:"discount_amount"`
	MinOrderAmount float64 `json:"min_order_amount,omitempty"`
}
