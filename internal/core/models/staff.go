package models

import (
	"time"

	"github.com/mlgaray/ecommerce_api/internal/core/errors"
)

// Valid sort fields for staff
const (
	StaffSortByName = "name"
)

type Staff struct {
	ID        int       `json:"id,omitempty"`
	ShopID    int       `json:"shop_id,omitempty"`
	IsActive  bool      `json:"is_active"`
	User      *User     `json:"user,omitempty"`
	Roles     []*Role   `json:"roles,omitempty"`
	CreatedAt time.Time `json:"created_at,omitempty"`
}

// GetID implements Identifiable interface for pagination
func (s *Staff) GetID() int {
	return s.ID
}

// GetSortValue implements Sortable interface for pagination
func (s *Staff) GetSortValue(sortBy string) interface{} {
	switch sortBy {
	case StaffSortByName:
		if s.User != nil {
			return s.User.Name
		}
		return ""
	case SortByCreatedAt:
		return s.CreatedAt
	default:
		return s.ID
	}
}

// ============================================================================
// StaffFilters - Query criteria for staff listing
// ============================================================================

// StaffFilters represents search and filter criteria for staff.
type StaffFilters struct {
	// Optional search filter (searches by user name and email)
	Search *string

	// Optional active status filter
	IsActive *bool

	// Sorting
	SortBy    string // "name", "created_at" (default: "created_at")
	SortOrder string // "asc", "desc" (default: "desc")

	// Pagination (cursor-based)
	Limit         int         // Number of items per page (default: 20, max: 100)
	LastID        *int        // ID of the last item from previous page (nil = first page)
	LastSortValue interface{} // Value of the sort field from last item
}

// Validated validates business rules and returns a normalized copy of StaffFilters.
func (f StaffFilters) Validated() (StaffFilters, error) {
	result := f

	result.normalizeLimit()

	if err := result.validateAndNormalizeSorting(); err != nil {
		return StaffFilters{}, err
	}

	return result, nil
}

func (f *StaffFilters) normalizeLimit() {
	if f.Limit <= 0 {
		f.Limit = DefaultLimit
	}
	if f.Limit > MaxLimit {
		f.Limit = MaxLimit
	}
}

func (f *StaffFilters) validateAndNormalizeSorting() error {
	validSortFields := map[string]bool{
		StaffSortByName: true,
		SortByCreatedAt: true,
		"":              true,
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
