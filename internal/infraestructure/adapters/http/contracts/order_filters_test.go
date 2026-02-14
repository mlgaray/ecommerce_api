package contracts

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/mlgaray/ecommerce_api/internal/application/services"
	httpErrors "github.com/mlgaray/ecommerce_api/internal/infraestructure/adapters/http/errors"
)

// =============================================================================
// NewOrderFiltersRequest Tests
// =============================================================================

func TestNewOrderFiltersRequest(t *testing.T) {
	t.Run("when all params provided then parses correctly", func(t *testing.T) {
		// Arrange
		params := map[string][]string{
			"search":    {"John"},
			"status":    {"pending"},
			"date_from": {"2026-02-10"},
			"date_to":   {"2026-02-13"},
			"tz":        {"America/Buenos_Aires"},
			"sort":      {"total"},
			"order":     {"asc"},
			"limit":     {"20"},
			"cursor":    {"abc123"},
		}

		// Act
		result, err := NewOrderFiltersRequest(params)

		// Assert
		require.NoError(t, err)
		require.NotNil(t, result)
		assert.Equal(t, "John", *result.Search)
		assert.Equal(t, "pending", *result.Status)
		assert.Equal(t, "2026-02-10", *result.DateFrom)
		assert.Equal(t, "2026-02-13", *result.DateTo)
		assert.Equal(t, "America/Buenos_Aires", *result.Timezone)
		assert.Equal(t, "total", result.SortBy)
		assert.Equal(t, "asc", result.SortOrder)
		assert.Equal(t, 20, result.Limit)
		assert.Equal(t, "abc123", result.Cursor)
	})

	t.Run("when no params then returns empty request", func(t *testing.T) {
		// Arrange
		params := map[string][]string{}

		// Act
		result, err := NewOrderFiltersRequest(params)

		// Assert
		require.NoError(t, err)
		require.NotNil(t, result)
		assert.Nil(t, result.Search)
		assert.Nil(t, result.Status)
		assert.Nil(t, result.DateFrom)
		assert.Nil(t, result.DateTo)
		assert.Nil(t, result.Timezone)
		assert.Equal(t, "", result.SortBy)
		assert.Equal(t, "", result.SortOrder)
		assert.Equal(t, 0, result.Limit)
		assert.Equal(t, "", result.Cursor)
	})

	t.Run("when only date_from provided then sets DateFrom", func(t *testing.T) {
		// Arrange
		params := map[string][]string{
			"date_from": {"2026-02-10"},
		}

		// Act
		result, err := NewOrderFiltersRequest(params)

		// Assert
		require.NoError(t, err)
		require.NotNil(t, result.DateFrom)
		assert.Equal(t, "2026-02-10", *result.DateFrom)
		assert.Nil(t, result.DateTo)
	})

	t.Run("when only date_to provided then sets DateTo", func(t *testing.T) {
		// Arrange
		params := map[string][]string{
			"date_to": {"2026-02-13"},
		}

		// Act
		result, err := NewOrderFiltersRequest(params)

		// Assert
		require.NoError(t, err)
		assert.Nil(t, result.DateFrom)
		require.NotNil(t, result.DateTo)
		assert.Equal(t, "2026-02-13", *result.DateTo)
	})

	t.Run("when tz provided then sets Timezone", func(t *testing.T) {
		// Arrange
		params := map[string][]string{
			"tz": {"America/Buenos_Aires"},
		}

		// Act
		result, err := NewOrderFiltersRequest(params)

		// Assert
		require.NoError(t, err)
		require.NotNil(t, result.Timezone)
		assert.Equal(t, "America/Buenos_Aires", *result.Timezone)
	})

	t.Run("when invalid limit then returns error", func(t *testing.T) {
		// Arrange
		params := map[string][]string{
			"limit": {"abc"},
		}

		// Act
		result, err := NewOrderFiltersRequest(params)

		// Assert
		assert.Error(t, err)
		assert.Nil(t, result)
		var badReqErr *httpErrors.BadRequestError
		assert.ErrorAs(t, err, &badReqErr)
		assert.Equal(t, "invalid_limit_format", badReqErr.Message)
	})
}

// =============================================================================
// OrderFiltersRequest.Validate Tests
// =============================================================================

func TestOrderFiltersRequest_Validate(t *testing.T) {
	t.Run("when valid dates then no error", func(t *testing.T) {
		// Arrange
		dateFrom := "2026-02-10"
		dateTo := "2026-02-13"
		request := &OrderFiltersRequest{
			DateFrom: &dateFrom,
			DateTo:   &dateTo,
		}

		// Act
		err := request.Validate()

		// Assert
		assert.NoError(t, err)
	})

	t.Run("when invalid date_from format then returns error", func(t *testing.T) {
		// Arrange
		dateFrom := "10-02-2026"
		request := &OrderFiltersRequest{
			DateFrom: &dateFrom,
		}

		// Act
		err := request.Validate()

		// Assert
		assert.Error(t, err)
		var badReqErr *httpErrors.BadRequestError
		assert.ErrorAs(t, err, &badReqErr)
		assert.Equal(t, "invalid_date_from_format", badReqErr.Message)
	})

	t.Run("when invalid date_to format then returns error", func(t *testing.T) {
		// Arrange
		dateTo := "13/02/2026"
		request := &OrderFiltersRequest{
			DateTo: &dateTo,
		}

		// Act
		err := request.Validate()

		// Assert
		assert.Error(t, err)
		var badReqErr *httpErrors.BadRequestError
		assert.ErrorAs(t, err, &badReqErr)
		assert.Equal(t, "invalid_date_to_format", badReqErr.Message)
	})

	t.Run("when empty search term then returns error", func(t *testing.T) {
		// Arrange
		search := "   "
		request := &OrderFiltersRequest{
			Search: &search,
		}

		// Act
		err := request.Validate()

		// Assert
		assert.Error(t, err)
		var badReqErr *httpErrors.BadRequestError
		assert.ErrorAs(t, err, &badReqErr)
		assert.Equal(t, "search_term_cannot_be_empty", badReqErr.Message)
	})

	t.Run("when negative limit then returns error", func(t *testing.T) {
		// Arrange
		request := &OrderFiltersRequest{
			Limit: -5,
		}

		// Act
		err := request.Validate()

		// Assert
		assert.Error(t, err)
		var badReqErr *httpErrors.BadRequestError
		assert.ErrorAs(t, err, &badReqErr)
		assert.Equal(t, "limit_cannot_be_negative", badReqErr.Message)
	})

	t.Run("when all valid then no error", func(t *testing.T) {
		// Arrange
		search := "John"
		status := "pending"
		dateFrom := "2026-02-10"
		dateTo := "2026-02-13"
		request := &OrderFiltersRequest{
			Search:   &search,
			Status:   &status,
			DateFrom: &dateFrom,
			DateTo:   &dateTo,
			Limit:    20,
		}

		// Act
		err := request.Validate()

		// Assert
		assert.NoError(t, err)
	})
}

// =============================================================================
// OrderFiltersRequest.ToOrderFilters Tests
// =============================================================================

func TestOrderFiltersRequest_ToOrderFilters(t *testing.T) {
	t.Run("when dates provided then parses to time.Time in UTC", func(t *testing.T) {
		// Arrange
		dateFrom := "2026-02-10"
		request := &OrderFiltersRequest{
			DateFrom: &dateFrom,
		}

		// Act
		result := request.ToOrderFilters()

		// Assert
		require.NotNil(t, result.DateFrom)
		expected := time.Date(2026, 2, 10, 0, 0, 0, 0, time.UTC)
		assert.Equal(t, expected, *result.DateFrom)
	})

	t.Run("when date_to provided then sets end of day", func(t *testing.T) {
		// Arrange
		dateTo := "2026-02-10"
		request := &OrderFiltersRequest{
			DateTo: &dateTo,
		}

		// Act
		result := request.ToOrderFilters()

		// Assert
		require.NotNil(t, result.DateTo)
		// End of day = start of day + 24h - 1 microsecond = 23:59:59.999999
		expected := time.Date(2026, 2, 10, 23, 59, 59, 999999000, time.UTC)
		assert.Equal(t, expected, *result.DateTo)
	})

	t.Run("when tz provided then passes through to filters Timezone", func(t *testing.T) {
		// Arrange
		tz := "America/Buenos_Aires"
		request := &OrderFiltersRequest{
			Timezone: &tz,
		}

		// Act
		result := request.ToOrderFilters()

		// Assert
		require.NotNil(t, result.Timezone)
		assert.Equal(t, "America/Buenos_Aires", *result.Timezone)
	})

	t.Run("when cursor provided then decodes LastID and LastSortValue", func(t *testing.T) {
		// Arrange
		cursorData := services.CursorData{
			ID:        42,
			SortValue: "2026-02-10T15:00:00Z",
		}
		cursor, err := services.EncodeCursor(cursorData)
		require.NoError(t, err)

		request := &OrderFiltersRequest{
			Cursor: cursor,
		}

		// Act
		result := request.ToOrderFilters()

		// Assert
		require.NotNil(t, result.LastID)
		assert.Equal(t, 42, *result.LastID)
		assert.NotNil(t, result.LastSortValue)
	})

	t.Run("when invalid cursor then treats as first page", func(t *testing.T) {
		// Arrange
		request := &OrderFiltersRequest{
			Cursor: "invalid_cursor_not_base64!!!",
		}

		// Act
		result := request.ToOrderFilters()

		// Assert
		assert.Nil(t, result.LastID)
		assert.Nil(t, result.LastSortValue)
	})

	t.Run("when no dates then DateFrom and DateTo are nil", func(t *testing.T) {
		// Arrange
		request := &OrderFiltersRequest{}

		// Act
		result := request.ToOrderFilters()

		// Assert
		assert.Nil(t, result.DateFrom)
		assert.Nil(t, result.DateTo)
	})
}
