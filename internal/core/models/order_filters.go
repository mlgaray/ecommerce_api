package models

import (
	"time"

	"github.com/mlgaray/ecommerce_api/internal/core/errors"
)

// Sort field constants for order queries
const (
	OrderSortByTotal       = "total"
	OrderSortByOrderNumber = "order_number"
	// SortByCreatedAt is reused from constants.go
)

// OrderFilters represents search and filter criteria for orders.
// Note: ShopID is NOT included here - it's a context parameter passed separately
// in method signatures (e.g., GetAllByShopIDWithFilters(ctx, shopID, filters)).
type OrderFilters struct {
	// Optional search filter (ILIKE on order_number and customer_name)
	Search *string // nil = no search applied

	// Optional status filter
	Status *string // nil = no status filter; valid: "pending", "confirmed", "completed", "cancelled"

	// Optional date range filters
	DateFrom *time.Time // nil = no start date filter (inclusive)
	DateTo   *time.Time // nil = no end date filter (inclusive, end of day)

	// Optional IANA timezone for date interpretation (e.g., "America/Buenos_Aires")
	// Sent by frontend to avoid extra DB lookup. Falls back to shop's timezone if nil.
	Timezone *string

	// Sorting
	SortBy    string // "created_at", "total", "order_number" (default: "created_at")
	SortOrder string // "asc", "desc" (default: "desc")

	// Pagination (cursor-based)
	Limit         int         // Number of items per page (default: 20, max: 100)
	LastID        *int        // ID of the last item from previous page (nil = first page)
	LastSortValue interface{} // Value of the sort field from last item (nil = first page)
}

// Validated validates business rules and returns a normalized copy of OrderFilters.
// This is an immutable operation - the original struct is not modified.
// Returns the validated/normalized copy and any validation error.
func (f OrderFilters) Validated() (OrderFilters, error) {
	result := f // Create a copy

	result.normalizeLimit()

	if err := result.validateDateRange(); err != nil {
		return OrderFilters{}, err
	}

	if err := result.validateStatus(); err != nil {
		return OrderFilters{}, err
	}

	if err := result.validateAndNormalizeSorting(); err != nil {
		return OrderFilters{}, err
	}

	return result, nil
}

// normalizeLimit ensures limit is within valid bounds
func (f *OrderFilters) normalizeLimit() {
	if f.Limit <= 0 {
		f.Limit = DefaultLimit
	}
	if f.Limit > MaxLimit {
		f.Limit = MaxLimit
	}
}

// validateDateRange validates that date_from is not after date_to
func (f *OrderFilters) validateDateRange() error {
	if f.DateFrom != nil && f.DateTo != nil {
		if f.DateFrom.After(*f.DateTo) {
			return &errors.ValidationError{
				Message: errors.DateFromCannotBeAfterDateTo,
			}
		}
	}
	return nil
}

// validateStatus validates that status is a valid order status
func (f *OrderFilters) validateStatus() error {
	if f.Status != nil {
		validStatuses := map[string]bool{
			string(OrderStatusPending):   true,
			string(OrderStatusConfirmed): true,
			string(OrderStatusCompleted): true,
			string(OrderStatusCancelled): true,
		}

		if !validStatuses[*f.Status] {
			return &errors.ValidationError{
				Message: errors.InvalidOrderStatus,
			}
		}
	}
	return nil
}

// WithTimezone re-interprets date filters in the given timezone.
// Extracts the date (year, month, day) and reconstructs it at the start/end of day
// in the shop's timezone. This ensures "Feb 10" means the full day in the shop's
// local time, not in UTC.
// Immutable operation - returns a new copy.
func (f OrderFilters) WithTimezone(loc *time.Location) OrderFilters {
	result := f

	if result.DateFrom != nil {
		y, m, d := result.DateFrom.Date()
		dateFrom := time.Date(y, m, d, 0, 0, 0, 0, loc)
		result.DateFrom = &dateFrom
	}

	if result.DateTo != nil {
		y, m, d := result.DateTo.Date()
		dateTo := time.Date(y, m, d, 23, 59, 59, 999999000, loc)
		result.DateTo = &dateTo
	}

	return result
}

// validateAndNormalizeSorting validates sort field and order, sets defaults
func (f *OrderFilters) validateAndNormalizeSorting() error {
	// Validate sort field (prevent SQL injection)
	validSortFields := map[string]bool{
		SortByCreatedAt:        true,
		OrderSortByTotal:       true,
		OrderSortByOrderNumber: true,
		"":                     true, // Empty = use default
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
