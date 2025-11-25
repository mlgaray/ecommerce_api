package models

import (
	"github.com/mlgaray/ecommerce_api/internal/core/errors"
)

// ProductFilters represents search and filter criteria for products
// This is a domain model with business validation rules
type ProductFilters struct {
	// Required filters
	ShopID int // Required: shop context for multi-tenancy

	// Optional search filter
	Search *string // nil = no search applied

	// Optional category filter
	CategoryID *int // nil = no category filter

	// Optional boolean filters
	IsActive      *bool // nil = no filter on active status
	IsHighlighted *bool // nil = no filter on highlighted status
	IsPromotional *bool // nil = no filter on promotional status

	// Optional price range filters
	MinPrice *float64 // nil = no minimum price filter
	MaxPrice *float64 // nil = no maximum price filter

	// Sorting
	SortBy    string // "price", "name", "created_at" (default: "created_at")
	SortOrder string // "asc", "desc" (default: "desc")

	// Pagination (cursor-based)
	Limit         int         // Number of items per page (default: 20, max: 100)
	LastID        *int        // ID of the last item from previous page (nil = first page)
	LastSortValue interface{} // Value of the sort field from last item (nil = first page or sorting by ID)
}

// Validate validates business rules for ProductFilters
// This ensures domain invariants are maintained
func (f *ProductFilters) Validate() error {
	// Business rule: shop_id is required (multi-tenancy)
	if f.ShopID <= 0 {
		return &errors.ValidationError{
			Message: errors.ShopIDIsRequired,
		}
	}

	// Business rule: limit must be positive and within reasonable bounds
	if f.Limit <= 0 {
		f.Limit = 20 // Default
	}
	if f.Limit > 100 {
		f.Limit = 100 // Max to prevent abuse
	}

	// Business rule: price range validations
	if f.MinPrice != nil && *f.MinPrice < 0 {
		return &errors.ValidationError{
			Message: errors.MinPriceCannotBeNegative,
		}
	}

	if f.MaxPrice != nil && *f.MaxPrice < 0 {
		return &errors.ValidationError{
			Message: errors.MaxPriceCannotBeNegative,
		}
	}

	if f.MinPrice != nil && f.MaxPrice != nil && *f.MinPrice > *f.MaxPrice {
		return &errors.ValidationError{
			Message: errors.MinPriceCannotBeGreaterThanMaxPrice,
		}
	}

	// Business rule: validate sort field (prevent SQL injection)
	validSortFields := map[string]bool{
		"price":      true,
		"name":       true,
		"created_at": true,
		"":           true, // Empty = use default
	}

	if !validSortFields[f.SortBy] {
		return &errors.ValidationError{
			Message: errors.InvalidSortField,
		}
	}

	// Set default sort field
	if f.SortBy == "" {
		f.SortBy = "created_at"
	}

	// Business rule: validate sort order
	if f.SortOrder != "asc" && f.SortOrder != "desc" && f.SortOrder != "" {
		return &errors.ValidationError{
			Message: errors.InvalidSortOrder,
		}
	}

	// Set default sort order
	if f.SortOrder == "" {
		f.SortOrder = "desc"
	}

	return nil
}
