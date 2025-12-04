package models

import (
	"time"

	"github.com/mlgaray/ecommerce_api/internal/core/errors"
)

type Category struct {
	ID          int       `json:"id,omitempty"`
	Name        string    `json:"name,omitempty"`
	Description string    `json:"description,omitempty"`
	Image       *Image    `json:"image,omitempty"`
	CreatedAt   time.Time `json:"created_at,omitempty"`
}

// GetID implements Identifiable interface for pagination
func (c *Category) GetID() int {
	return c.ID
}

// GetSortValue implements Sortable interface for pagination
// Returns the value of the specified sort field for cursor-based pagination
func (c *Category) GetSortValue(sortBy string) interface{} {
	switch sortBy {
	case CategorySortByName:
		return c.Name
	case CategorySortByCreatedAt:
		return c.CreatedAt
	default:
		return c.ID
	}
}

// ============================================================================
// CategoryFilters - Query criteria for category listing
// ============================================================================

// Valid sort fields for categories
const (
	CategorySortByName      = "name"
	CategorySortByCreatedAt = "created_at"
)

// CategoryFilters represents search and filter criteria for categories.
// Note: ShopID is NOT included here - it's a context parameter passed separately
// in method signatures (e.g., GetAllByShopIDWithFilters(ctx, shopID, filters)).
type CategoryFilters struct {
	// Optional search filter
	Search *string // nil = no search applied (searches by name)

	// Sorting
	SortBy    string // "name", "created_at" (default: "created_at")
	SortOrder string // "asc", "desc" (default: "desc")

	// Pagination (cursor-based)
	Limit         int         // Number of items per page (default: 20, max: 100)
	LastID        *int        // ID of the last item from previous page (nil = first page)
	LastSortValue interface{} // Value of the sort field from last item
}

// Validated validates business rules and returns a normalized copy of CategoryFilters.
// This is an immutable operation - the original struct is not modified.
// Returns the validated/normalized copy and any validation error.
func (f CategoryFilters) Validated() (CategoryFilters, error) {
	result := f // Create a copy

	result.normalizeLimit()

	if err := result.validateAndNormalizeSorting(); err != nil {
		return CategoryFilters{}, err
	}

	return result, nil
}

// normalizeLimit ensures limit is within valid bounds
func (f *CategoryFilters) normalizeLimit() {
	if f.Limit <= 0 {
		f.Limit = DefaultLimit
	}
	if f.Limit > MaxLimit {
		f.Limit = MaxLimit
	}
}

// validateAndNormalizeSorting validates sort field and order, sets defaults
func (f *CategoryFilters) validateAndNormalizeSorting() error {
	validSortFields := map[string]bool{
		CategorySortByName:      true,
		CategorySortByCreatedAt: true,
		"":                      true,
	}

	if !validSortFields[f.SortBy] {
		return &errors.ValidationError{
			Message: errors.InvalidSortField,
		}
	}

	if f.SortBy == "" {
		f.SortBy = CategorySortByCreatedAt
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
