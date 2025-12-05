package models

import (
	"github.com/mlgaray/ecommerce_api/internal/core/errors"
)

// ProductFilters represents search and filter criteria for products.
// Note: ShopID is NOT included here - it's a context parameter passed separately
// in method signatures (e.g., GetAllByShopIDWithFilters(ctx, shopID, filters)).
type ProductFilters struct {
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

// Validated validates business rules and returns a normalized copy of ProductFilters.
// This is an immutable operation - the original struct is not modified.
// Returns the validated/normalized copy and any validation error.
func (f ProductFilters) Validated() (ProductFilters, error) {
	result := f // Create a copy

	result.normalizeLimit()

	if err := result.validatePriceRange(); err != nil {
		return ProductFilters{}, err
	}

	if err := result.validateAndNormalizeSorting(); err != nil {
		return ProductFilters{}, err
	}

	return result, nil
}

// normalizeLimit ensures limit is within valid bounds
func (f *ProductFilters) normalizeLimit() {
	if f.Limit <= 0 {
		f.Limit = DefaultLimit
	}
	if f.Limit > MaxLimit {
		f.Limit = MaxLimit
	}
}

// validatePriceRange validates price range filters
func (f *ProductFilters) validatePriceRange() error {
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

	return nil
}

// validateAndNormalizeSorting validates sort field and order, sets defaults
func (f *ProductFilters) validateAndNormalizeSorting() error {
	// Validate sort field (prevent SQL injection)
	validSortFields := map[string]bool{
		SortByPrice:     true,
		SortByName:      true,
		SortByCreatedAt: true,
		"":              true, // Empty = use default
	}

	if !validSortFields[f.SortBy] {
		return &errors.ValidationError{
			Message: errors.InvalidSortField,
		}
	}

	// Set default sort field
	if f.SortBy == "" {
		f.SortBy = SortByCreatedAt
	}

	// Validate sort order
	if f.SortOrder != SortOrderAsc && f.SortOrder != SortOrderDesc && f.SortOrder != "" {
		return &errors.ValidationError{
			Message: errors.InvalidSortOrder,
		}
	}

	// Set default sort order
	if f.SortOrder == "" {
		f.SortOrder = SortOrderDesc
	}

	return nil
}
