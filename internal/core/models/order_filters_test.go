package models

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/mlgaray/ecommerce_api/internal/core/errors"
)

// =============================================================================
// OrderFilters.WithTimezone Tests
// =============================================================================

func TestOrderFilters_WithTimezone(t *testing.T) {
	t.Run("when both dates set then re-interprets in given timezone", func(t *testing.T) {
		// Arrange
		dateFrom := time.Date(2024, 6, 15, 10, 30, 0, 0, time.UTC)
		dateTo := time.Date(2024, 6, 20, 14, 45, 0, 0, time.UTC)
		loc, err := time.LoadLocation("America/Buenos_Aires")
		require.NoError(t, err)

		filters := OrderFilters{
			DateFrom: &dateFrom,
			DateTo:   &dateTo,
		}

		// Act
		result := filters.WithTimezone(loc)

		// Assert
		expectedFrom := time.Date(2024, 6, 15, 0, 0, 0, 0, loc)
		expectedTo := time.Date(2024, 6, 20, 23, 59, 59, 999999000, loc)
		assert.Equal(t, expectedFrom, *result.DateFrom)
		assert.Equal(t, expectedTo, *result.DateTo)
	})

	t.Run("when only DateFrom set then re-interprets only DateFrom", func(t *testing.T) {
		// Arrange
		dateFrom := time.Date(2024, 3, 10, 8, 0, 0, 0, time.UTC)
		loc, err := time.LoadLocation("America/Buenos_Aires")
		require.NoError(t, err)

		filters := OrderFilters{
			DateFrom: &dateFrom,
			DateTo:   nil,
		}

		// Act
		result := filters.WithTimezone(loc)

		// Assert
		expectedFrom := time.Date(2024, 3, 10, 0, 0, 0, 0, loc)
		assert.Equal(t, expectedFrom, *result.DateFrom)
		assert.Nil(t, result.DateTo)
	})

	t.Run("when only DateTo set then re-interprets only DateTo", func(t *testing.T) {
		// Arrange
		dateTo := time.Date(2024, 12, 25, 18, 0, 0, 0, time.UTC)
		loc, err := time.LoadLocation("America/Buenos_Aires")
		require.NoError(t, err)

		filters := OrderFilters{
			DateFrom: nil,
			DateTo:   &dateTo,
		}

		// Act
		result := filters.WithTimezone(loc)

		// Assert
		assert.Nil(t, result.DateFrom)
		expectedTo := time.Date(2024, 12, 25, 23, 59, 59, 999999000, loc)
		assert.Equal(t, expectedTo, *result.DateTo)
	})

	t.Run("when no dates set then returns unchanged", func(t *testing.T) {
		// Arrange
		loc, err := time.LoadLocation("America/Buenos_Aires")
		require.NoError(t, err)

		filters := OrderFilters{
			DateFrom: nil,
			DateTo:   nil,
		}

		// Act
		result := filters.WithTimezone(loc)

		// Assert
		assert.Nil(t, result.DateFrom)
		assert.Nil(t, result.DateTo)
	})

	t.Run("when DateTo set then end of day is 23:59:59.999999 in timezone", func(t *testing.T) {
		// Arrange
		dateTo := time.Date(2024, 8, 1, 12, 0, 0, 0, time.UTC)
		loc, err := time.LoadLocation("Europe/London")
		require.NoError(t, err)

		filters := OrderFilters{
			DateTo: &dateTo,
		}

		// Act
		result := filters.WithTimezone(loc)

		// Assert
		assert.Equal(t, 23, result.DateTo.Hour())
		assert.Equal(t, 59, result.DateTo.Minute())
		assert.Equal(t, 59, result.DateTo.Second())
		assert.Equal(t, 999999000, result.DateTo.Nanosecond())
		assert.Equal(t, loc, result.DateTo.Location())
	})

	t.Run("when called then original struct is not modified", func(t *testing.T) {
		// Arrange
		dateFrom := time.Date(2024, 6, 15, 10, 30, 0, 0, time.UTC)
		dateTo := time.Date(2024, 6, 20, 14, 45, 0, 0, time.UTC)
		loc, err := time.LoadLocation("America/Buenos_Aires")
		require.NoError(t, err)

		originalDateFrom := dateFrom
		originalDateTo := dateTo

		filters := OrderFilters{
			DateFrom: &dateFrom,
			DateTo:   &dateTo,
		}

		// Act
		_ = filters.WithTimezone(loc)

		// Assert
		assert.Equal(t, originalDateFrom, *filters.DateFrom)
		assert.Equal(t, originalDateTo, *filters.DateTo)
		assert.Equal(t, time.UTC, filters.DateFrom.Location())
		assert.Equal(t, time.UTC, filters.DateTo.Location())
	})
}

// =============================================================================
// OrderFilters.Validated Tests
// =============================================================================

func TestOrderFilters_Validated(t *testing.T) {
	t.Run("when all defaults then returns normalized filters", func(t *testing.T) {
		// Arrange
		filters := OrderFilters{}

		// Act
		result, err := filters.Validated()

		// Assert
		assert.NoError(t, err)
		assert.Equal(t, DefaultLimit, result.Limit)
		assert.Equal(t, SortByCreatedAt, result.SortBy)
		assert.Equal(t, SortOrderDesc, result.SortOrder)
	})

	t.Run("when valid filters then preserves values", func(t *testing.T) {
		// Arrange
		search := "order-123"
		status := string(OrderStatusPending)
		dateFrom := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
		dateTo := time.Date(2024, 12, 31, 23, 59, 59, 0, time.UTC)
		timezone := "America/Buenos_Aires"
		lastID := 10

		filters := OrderFilters{
			Search:    &search,
			Status:    &status,
			DateFrom:  &dateFrom,
			DateTo:    &dateTo,
			Timezone:  &timezone,
			SortBy:    OrderSortByTotal,
			SortOrder: SortOrderAsc,
			Limit:     50,
			LastID:    &lastID,
		}

		// Act
		result, err := filters.Validated()

		// Assert
		assert.NoError(t, err)
		assert.Equal(t, &search, result.Search)
		assert.Equal(t, &status, result.Status)
		assert.Equal(t, &dateFrom, result.DateFrom)
		assert.Equal(t, &dateTo, result.DateTo)
		assert.Equal(t, &timezone, result.Timezone)
		assert.Equal(t, OrderSortByTotal, result.SortBy)
		assert.Equal(t, SortOrderAsc, result.SortOrder)
		assert.Equal(t, 50, result.Limit)
		assert.Equal(t, &lastID, result.LastID)
	})

	t.Run("when original is not modified then maintains immutability", func(t *testing.T) {
		// Arrange
		filters := OrderFilters{
			SortBy:    "",
			SortOrder: "",
			Limit:     0,
		}

		// Act
		result, err := filters.Validated()

		// Assert
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
// OrderFilters Status Validation Tests
// =============================================================================

func TestOrderFilters_Validated_Status(t *testing.T) {
	t.Run("when status is pending then succeeds", func(t *testing.T) {
		// Arrange
		status := string(OrderStatusPending)
		filters := OrderFilters{Status: &status}

		// Act
		result, err := filters.Validated()

		// Assert
		assert.NoError(t, err)
		assert.Equal(t, string(OrderStatusPending), *result.Status)
	})

	t.Run("when status is confirmed then succeeds", func(t *testing.T) {
		// Arrange
		status := string(OrderStatusConfirmed)
		filters := OrderFilters{Status: &status}

		// Act
		result, err := filters.Validated()

		// Assert
		assert.NoError(t, err)
		assert.Equal(t, string(OrderStatusConfirmed), *result.Status)
	})

	t.Run("when status is completed then succeeds", func(t *testing.T) {
		// Arrange
		status := string(OrderStatusCompleted)
		filters := OrderFilters{Status: &status}

		// Act
		result, err := filters.Validated()

		// Assert
		assert.NoError(t, err)
		assert.Equal(t, string(OrderStatusCompleted), *result.Status)
	})

	t.Run("when status is cancelled then succeeds", func(t *testing.T) {
		// Arrange
		status := string(OrderStatusCancelled)
		filters := OrderFilters{Status: &status}

		// Act
		result, err := filters.Validated()

		// Assert
		assert.NoError(t, err)
		assert.Equal(t, string(OrderStatusCancelled), *result.Status)
	})

	t.Run("when status is nil then succeeds", func(t *testing.T) {
		// Arrange
		filters := OrderFilters{Status: nil}

		// Act
		result, err := filters.Validated()

		// Assert
		assert.NoError(t, err)
		assert.Nil(t, result.Status)
	})

	t.Run("when status is invalid then returns validation error", func(t *testing.T) {
		// Arrange
		status := "shipped"
		filters := OrderFilters{Status: &status}

		// Act
		result, err := filters.Validated()

		// Assert
		assert.Error(t, err)
		var validationErr *errors.ValidationError
		assert.ErrorAs(t, err, &validationErr)
		assert.Equal(t, errors.InvalidOrderStatus, validationErr.Message)
		assert.Equal(t, OrderFilters{}, result)
	})

	t.Run("when status is empty string then returns validation error", func(t *testing.T) {
		// Arrange
		status := ""
		filters := OrderFilters{Status: &status}

		// Act
		result, err := filters.Validated()

		// Assert
		assert.Error(t, err)
		var validationErr *errors.ValidationError
		assert.ErrorAs(t, err, &validationErr)
		assert.Equal(t, errors.InvalidOrderStatus, validationErr.Message)
		assert.Equal(t, OrderFilters{}, result)
	})

	t.Run("when status is uppercase then returns validation error", func(t *testing.T) {
		// Arrange
		status := "PENDING"
		filters := OrderFilters{Status: &status}

		// Act
		result, err := filters.Validated()

		// Assert
		assert.Error(t, err)
		var validationErr *errors.ValidationError
		assert.ErrorAs(t, err, &validationErr)
		assert.Equal(t, errors.InvalidOrderStatus, validationErr.Message)
		assert.Equal(t, OrderFilters{}, result)
	})
}

// =============================================================================
// OrderFilters Date Range Validation Tests
// =============================================================================

func TestOrderFilters_Validated_DateRange(t *testing.T) {
	t.Run("when date_from is after date_to then returns validation error", func(t *testing.T) {
		// Arrange
		dateFrom := time.Date(2024, 6, 20, 0, 0, 0, 0, time.UTC)
		dateTo := time.Date(2024, 6, 10, 0, 0, 0, 0, time.UTC)
		filters := OrderFilters{
			DateFrom: &dateFrom,
			DateTo:   &dateTo,
		}

		// Act
		result, err := filters.Validated()

		// Assert
		assert.Error(t, err)
		var validationErr *errors.ValidationError
		assert.ErrorAs(t, err, &validationErr)
		assert.Equal(t, errors.DateFromCannotBeAfterDateTo, validationErr.Message)
		assert.Equal(t, OrderFilters{}, result)
	})

	t.Run("when date_from is before date_to then succeeds", func(t *testing.T) {
		// Arrange
		dateFrom := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
		dateTo := time.Date(2024, 12, 31, 0, 0, 0, 0, time.UTC)
		filters := OrderFilters{
			DateFrom: &dateFrom,
			DateTo:   &dateTo,
		}

		// Act
		result, err := filters.Validated()

		// Assert
		assert.NoError(t, err)
		assert.Equal(t, dateFrom, *result.DateFrom)
		assert.Equal(t, dateTo, *result.DateTo)
	})

	t.Run("when date_from equals date_to then succeeds", func(t *testing.T) {
		// Arrange
		date := time.Date(2024, 6, 15, 0, 0, 0, 0, time.UTC)
		filters := OrderFilters{
			DateFrom: &date,
			DateTo:   &date,
		}

		// Act
		result, err := filters.Validated()

		// Assert
		assert.NoError(t, err)
		assert.Equal(t, date, *result.DateFrom)
		assert.Equal(t, date, *result.DateTo)
	})

	t.Run("when only date_from set then succeeds", func(t *testing.T) {
		// Arrange
		dateFrom := time.Date(2024, 6, 15, 0, 0, 0, 0, time.UTC)
		filters := OrderFilters{
			DateFrom: &dateFrom,
			DateTo:   nil,
		}

		// Act
		result, err := filters.Validated()

		// Assert
		assert.NoError(t, err)
		assert.Equal(t, dateFrom, *result.DateFrom)
		assert.Nil(t, result.DateTo)
	})

	t.Run("when only date_to set then succeeds", func(t *testing.T) {
		// Arrange
		dateTo := time.Date(2024, 6, 15, 0, 0, 0, 0, time.UTC)
		filters := OrderFilters{
			DateFrom: nil,
			DateTo:   &dateTo,
		}

		// Act
		result, err := filters.Validated()

		// Assert
		assert.NoError(t, err)
		assert.Nil(t, result.DateFrom)
		assert.Equal(t, dateTo, *result.DateTo)
	})

	t.Run("when no dates set then succeeds", func(t *testing.T) {
		// Arrange
		filters := OrderFilters{
			DateFrom: nil,
			DateTo:   nil,
		}

		// Act
		result, err := filters.Validated()

		// Assert
		assert.NoError(t, err)
		assert.Nil(t, result.DateFrom)
		assert.Nil(t, result.DateTo)
	})
}

// =============================================================================
// OrderFilters Limit Normalization Tests
// =============================================================================

func TestOrderFilters_Validated_Limit(t *testing.T) {
	t.Run("when limit is zero then uses default", func(t *testing.T) {
		// Arrange
		filters := OrderFilters{Limit: 0}

		// Act
		result, err := filters.Validated()

		// Assert
		assert.NoError(t, err)
		assert.Equal(t, DefaultLimit, result.Limit)
	})

	t.Run("when limit is negative then uses default", func(t *testing.T) {
		// Arrange
		filters := OrderFilters{Limit: -10}

		// Act
		result, err := filters.Validated()

		// Assert
		assert.NoError(t, err)
		assert.Equal(t, DefaultLimit, result.Limit)
	})

	t.Run("when limit exceeds max then uses max", func(t *testing.T) {
		// Arrange
		filters := OrderFilters{Limit: 200}

		// Act
		result, err := filters.Validated()

		// Assert
		assert.NoError(t, err)
		assert.Equal(t, MaxLimit, result.Limit)
	})

	t.Run("when limit equals max then uses max", func(t *testing.T) {
		// Arrange
		filters := OrderFilters{Limit: MaxLimit}

		// Act
		result, err := filters.Validated()

		// Assert
		assert.NoError(t, err)
		assert.Equal(t, MaxLimit, result.Limit)
	})

	t.Run("when limit is within bounds then uses provided value", func(t *testing.T) {
		// Arrange
		filters := OrderFilters{Limit: 50}

		// Act
		result, err := filters.Validated()

		// Assert
		assert.NoError(t, err)
		assert.Equal(t, 50, result.Limit)
	})

	t.Run("when limit is one then uses one", func(t *testing.T) {
		// Arrange
		filters := OrderFilters{Limit: 1}

		// Act
		result, err := filters.Validated()

		// Assert
		assert.NoError(t, err)
		assert.Equal(t, 1, result.Limit)
	})
}

// =============================================================================
// OrderFilters Sort Field Validation Tests
// =============================================================================

func TestOrderFilters_Validated_SortBy(t *testing.T) {
	t.Run("when sort_by is empty then uses default created_at", func(t *testing.T) {
		// Arrange
		filters := OrderFilters{SortBy: ""}

		// Act
		result, err := filters.Validated()

		// Assert
		assert.NoError(t, err)
		assert.Equal(t, SortByCreatedAt, result.SortBy)
	})

	t.Run("when sort_by is created_at then uses created_at", func(t *testing.T) {
		// Arrange
		filters := OrderFilters{SortBy: SortByCreatedAt}

		// Act
		result, err := filters.Validated()

		// Assert
		assert.NoError(t, err)
		assert.Equal(t, SortByCreatedAt, result.SortBy)
	})

	t.Run("when sort_by is total then uses total", func(t *testing.T) {
		// Arrange
		filters := OrderFilters{SortBy: OrderSortByTotal}

		// Act
		result, err := filters.Validated()

		// Assert
		assert.NoError(t, err)
		assert.Equal(t, OrderSortByTotal, result.SortBy)
	})

	t.Run("when sort_by is order_number then uses order_number", func(t *testing.T) {
		// Arrange
		filters := OrderFilters{SortBy: OrderSortByOrderNumber}

		// Act
		result, err := filters.Validated()

		// Assert
		assert.NoError(t, err)
		assert.Equal(t, OrderSortByOrderNumber, result.SortBy)
	})

	t.Run("when sort_by is invalid then returns validation error", func(t *testing.T) {
		// Arrange
		filters := OrderFilters{SortBy: "invalid_field"}

		// Act
		result, err := filters.Validated()

		// Assert
		assert.Error(t, err)
		var validationErr *errors.ValidationError
		assert.ErrorAs(t, err, &validationErr)
		assert.Equal(t, errors.InvalidSortField, validationErr.Message)
		assert.Equal(t, OrderFilters{}, result)
	})

	t.Run("when sort_by is sql injection attempt then returns validation error", func(t *testing.T) {
		// Arrange
		filters := OrderFilters{SortBy: "total; DROP TABLE orders;--"}

		// Act
		result, err := filters.Validated()

		// Assert
		assert.Error(t, err)
		var validationErr *errors.ValidationError
		assert.ErrorAs(t, err, &validationErr)
		assert.Equal(t, errors.InvalidSortField, validationErr.Message)
		assert.Equal(t, OrderFilters{}, result)
	})
}

// =============================================================================
// OrderFilters Sort Order Validation Tests
// =============================================================================

func TestOrderFilters_Validated_SortOrder(t *testing.T) {
	t.Run("when sort_order is empty then uses default desc", func(t *testing.T) {
		// Arrange
		filters := OrderFilters{SortOrder: ""}

		// Act
		result, err := filters.Validated()

		// Assert
		assert.NoError(t, err)
		assert.Equal(t, SortOrderDesc, result.SortOrder)
	})

	t.Run("when sort_order is asc then uses asc", func(t *testing.T) {
		// Arrange
		filters := OrderFilters{SortOrder: SortOrderAsc}

		// Act
		result, err := filters.Validated()

		// Assert
		assert.NoError(t, err)
		assert.Equal(t, SortOrderAsc, result.SortOrder)
	})

	t.Run("when sort_order is desc then uses desc", func(t *testing.T) {
		// Arrange
		filters := OrderFilters{SortOrder: SortOrderDesc}

		// Act
		result, err := filters.Validated()

		// Assert
		assert.NoError(t, err)
		assert.Equal(t, SortOrderDesc, result.SortOrder)
	})

	t.Run("when sort_order is invalid then returns validation error", func(t *testing.T) {
		// Arrange
		filters := OrderFilters{SortOrder: "random"}

		// Act
		result, err := filters.Validated()

		// Assert
		assert.Error(t, err)
		var validationErr *errors.ValidationError
		assert.ErrorAs(t, err, &validationErr)
		assert.Equal(t, errors.InvalidSortOrder, validationErr.Message)
		assert.Equal(t, OrderFilters{}, result)
	})

	t.Run("when sort_order is uppercase ASC then returns validation error", func(t *testing.T) {
		// Arrange
		filters := OrderFilters{SortOrder: "ASC"}

		// Act
		result, err := filters.Validated()

		// Assert
		assert.Error(t, err)
		var validationErr *errors.ValidationError
		assert.ErrorAs(t, err, &validationErr)
		assert.Equal(t, errors.InvalidSortOrder, validationErr.Message)
		assert.Equal(t, OrderFilters{}, result)
	})

	t.Run("when sort_order is uppercase DESC then returns validation error", func(t *testing.T) {
		// Arrange
		filters := OrderFilters{SortOrder: "DESC"}

		// Act
		result, err := filters.Validated()

		// Assert
		assert.Error(t, err)
		var validationErr *errors.ValidationError
		assert.ErrorAs(t, err, &validationErr)
		assert.Equal(t, errors.InvalidSortOrder, validationErr.Message)
		assert.Equal(t, OrderFilters{}, result)
	})
}
