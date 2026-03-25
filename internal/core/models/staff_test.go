package models

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/mlgaray/ecommerce_api/internal/core/errors"
)

// =============================================================================
// GetID Tests (Identifiable Interface)
// =============================================================================

func TestStaff_GetID(t *testing.T) {
	t.Run("returns staff ID", func(t *testing.T) {
		staff := &Staff{ID: 42}
		assert.Equal(t, 42, staff.GetID())
	})

	t.Run("returns zero when ID not set", func(t *testing.T) {
		staff := &Staff{}
		assert.Equal(t, 0, staff.GetID())
	})
}

// =============================================================================
// GetSortValue Tests (Sortable Interface)
// =============================================================================

func TestStaff_GetSortValue(t *testing.T) {
	now := time.Now()
	staff := &Staff{
		ID:        1,
		User:      &User{Name: "John Doe"},
		CreatedAt: now,
	}

	t.Run("returns user name when sorting by name", func(t *testing.T) {
		value := staff.GetSortValue(StaffSortByName)
		assert.Equal(t, "John Doe", value)
	})

	t.Run("returns created_at when sorting by created_at", func(t *testing.T) {
		value := staff.GetSortValue(SortByCreatedAt)
		assert.Equal(t, now, value)
	})

	t.Run("returns ID for unknown sort field", func(t *testing.T) {
		value := staff.GetSortValue("unknown_field")
		assert.Equal(t, 1, value)
	})

	t.Run("returns ID for empty sort field", func(t *testing.T) {
		value := staff.GetSortValue("")
		assert.Equal(t, 1, value)
	})

	t.Run("returns empty string when user is nil and sorting by name", func(t *testing.T) {
		staffNoUser := &Staff{ID: 2, User: nil, CreatedAt: now}
		value := staffNoUser.GetSortValue(StaffSortByName)
		assert.Equal(t, "", value)
	})
}

// =============================================================================
// StaffFilters.Validated Tests
// =============================================================================

func TestStaffFilters_Validated(t *testing.T) {
	t.Run("when all defaults then returns normalized filters", func(t *testing.T) {
		filters := StaffFilters{}

		result, err := filters.Validated()

		assert.NoError(t, err)
		assert.Equal(t, DefaultLimit, result.Limit)
		assert.Equal(t, SortByCreatedAt, result.SortBy)
		assert.Equal(t, SortOrderDesc, result.SortOrder)
	})

	t.Run("when valid filters then preserves values", func(t *testing.T) {
		search := "john"
		isActive := true
		lastID := 10
		filters := StaffFilters{
			Search:    &search,
			IsActive:  &isActive,
			SortBy:    StaffSortByName,
			SortOrder: SortOrderAsc,
			Limit:     50,
			LastID:    &lastID,
		}

		result, err := filters.Validated()

		assert.NoError(t, err)
		assert.Equal(t, &search, result.Search)
		assert.Equal(t, &isActive, result.IsActive)
		assert.Equal(t, StaffSortByName, result.SortBy)
		assert.Equal(t, SortOrderAsc, result.SortOrder)
		assert.Equal(t, 50, result.Limit)
		assert.Equal(t, &lastID, result.LastID)
	})

	t.Run("when invalid sort_by then returns validation error", func(t *testing.T) {
		filters := StaffFilters{SortBy: "invalid_field"}

		result, err := filters.Validated()

		assert.Error(t, err)
		var validationErr *errors.ValidationError
		assert.ErrorAs(t, err, &validationErr)
		assert.Equal(t, errors.InvalidSortField, validationErr.Message)
		assert.Equal(t, StaffFilters{}, result)
	})

	t.Run("when invalid sort_order then returns validation error", func(t *testing.T) {
		filters := StaffFilters{SortOrder: "invalid"}

		result, err := filters.Validated()

		assert.Error(t, err)
		var validationErr *errors.ValidationError
		assert.ErrorAs(t, err, &validationErr)
		assert.Equal(t, errors.InvalidSortOrder, validationErr.Message)
		assert.Equal(t, StaffFilters{}, result)
	})

	t.Run("when original is not modified then maintains immutability", func(t *testing.T) {
		filters := StaffFilters{
			SortBy:    "",
			SortOrder: "",
			Limit:     0,
		}

		result, err := filters.Validated()

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
// StaffFilters Limit Normalization Tests
// =============================================================================

func TestStaffFilters_Validated_Limit(t *testing.T) {
	t.Run("when limit is zero then uses default", func(t *testing.T) {
		filters := StaffFilters{Limit: 0}

		result, err := filters.Validated()

		assert.NoError(t, err)
		assert.Equal(t, DefaultLimit, result.Limit)
	})

	t.Run("when limit is negative then uses default", func(t *testing.T) {
		filters := StaffFilters{Limit: -10}

		result, err := filters.Validated()

		assert.NoError(t, err)
		assert.Equal(t, DefaultLimit, result.Limit)
	})

	t.Run("when limit exceeds max then uses max", func(t *testing.T) {
		filters := StaffFilters{Limit: 200}

		result, err := filters.Validated()

		assert.NoError(t, err)
		assert.Equal(t, MaxLimit, result.Limit)
	})

	t.Run("when limit equals max then uses max", func(t *testing.T) {
		filters := StaffFilters{Limit: MaxLimit}

		result, err := filters.Validated()

		assert.NoError(t, err)
		assert.Equal(t, MaxLimit, result.Limit)
	})

	t.Run("when limit is within bounds then uses provided value", func(t *testing.T) {
		filters := StaffFilters{Limit: 50}

		result, err := filters.Validated()

		assert.NoError(t, err)
		assert.Equal(t, 50, result.Limit)
	})

	t.Run("when limit is one then uses one", func(t *testing.T) {
		filters := StaffFilters{Limit: 1}

		result, err := filters.Validated()

		assert.NoError(t, err)
		assert.Equal(t, 1, result.Limit)
	})
}

// =============================================================================
// StaffFilters Sort Field Validation Tests
// =============================================================================

func TestStaffFilters_Validated_SortBy(t *testing.T) {
	t.Run("when sort_by is empty then uses default", func(t *testing.T) {
		filters := StaffFilters{SortBy: ""}

		result, err := filters.Validated()

		assert.NoError(t, err)
		assert.Equal(t, SortByCreatedAt, result.SortBy)
	})

	t.Run("when sort_by is name then uses name", func(t *testing.T) {
		filters := StaffFilters{SortBy: StaffSortByName}

		result, err := filters.Validated()

		assert.NoError(t, err)
		assert.Equal(t, StaffSortByName, result.SortBy)
	})

	t.Run("when sort_by is created_at then uses created_at", func(t *testing.T) {
		filters := StaffFilters{SortBy: SortByCreatedAt}

		result, err := filters.Validated()

		assert.NoError(t, err)
		assert.Equal(t, SortByCreatedAt, result.SortBy)
	})

	t.Run("when sort_by is invalid then returns validation error", func(t *testing.T) {
		filters := StaffFilters{SortBy: "invalid_field"}

		result, err := filters.Validated()

		assert.Error(t, err)
		var validationErr *errors.ValidationError
		assert.ErrorAs(t, err, &validationErr)
		assert.Equal(t, errors.InvalidSortField, validationErr.Message)
		assert.Equal(t, StaffFilters{}, result)
	})

	t.Run("when sort_by is sql injection attempt then returns validation error", func(t *testing.T) {
		filters := StaffFilters{SortBy: "name; DROP TABLE staff;--"}

		result, err := filters.Validated()

		assert.Error(t, err)
		var validationErr *errors.ValidationError
		assert.ErrorAs(t, err, &validationErr)
		assert.Equal(t, errors.InvalidSortField, validationErr.Message)
		assert.Equal(t, StaffFilters{}, result)
	})
}

// =============================================================================
// StaffFilters Sort Order Validation Tests
// =============================================================================

func TestStaffFilters_Validated_SortOrder(t *testing.T) {
	t.Run("when sort_order is empty then uses default desc", func(t *testing.T) {
		filters := StaffFilters{SortOrder: ""}

		result, err := filters.Validated()

		assert.NoError(t, err)
		assert.Equal(t, SortOrderDesc, result.SortOrder)
	})

	t.Run("when sort_order is asc then uses asc", func(t *testing.T) {
		filters := StaffFilters{SortOrder: SortOrderAsc}

		result, err := filters.Validated()

		assert.NoError(t, err)
		assert.Equal(t, SortOrderAsc, result.SortOrder)
	})

	t.Run("when sort_order is desc then uses desc", func(t *testing.T) {
		filters := StaffFilters{SortOrder: SortOrderDesc}

		result, err := filters.Validated()

		assert.NoError(t, err)
		assert.Equal(t, SortOrderDesc, result.SortOrder)
	})

	t.Run("when sort_order is invalid then returns validation error", func(t *testing.T) {
		filters := StaffFilters{SortOrder: "invalid"}

		result, err := filters.Validated()

		assert.Error(t, err)
		var validationErr *errors.ValidationError
		assert.ErrorAs(t, err, &validationErr)
		assert.Equal(t, errors.InvalidSortOrder, validationErr.Message)
		assert.Equal(t, StaffFilters{}, result)
	})

	t.Run("when sort_order is uppercase ASC then returns validation error", func(t *testing.T) {
		filters := StaffFilters{SortOrder: "ASC"}

		result, err := filters.Validated()

		assert.Error(t, err)
		var validationErr *errors.ValidationError
		assert.ErrorAs(t, err, &validationErr)
		assert.Equal(t, errors.InvalidSortOrder, validationErr.Message)
		assert.Equal(t, StaffFilters{}, result)
	})

	t.Run("when sort_order is uppercase DESC then returns validation error", func(t *testing.T) {
		filters := StaffFilters{SortOrder: "DESC"}

		result, err := filters.Validated()

		assert.Error(t, err)
		var validationErr *errors.ValidationError
		assert.ErrorAs(t, err, &validationErr)
		assert.Equal(t, errors.InvalidSortOrder, validationErr.Message)
		assert.Equal(t, StaffFilters{}, result)
	})
}
