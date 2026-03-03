package requests

import (
	"net/url"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	httpErrors "github.com/mlgaray/ecommerce_api/internal/infraestructure/adapters/http/errors"
)

// =============================================================================
// NewMetricsFiltersRequest
// =============================================================================

func TestNewMetricsFiltersRequest(t *testing.T) {
	t.Run("when all params provided then parses correctly", func(t *testing.T) {
		// Arrange
		params := url.Values{
			"date_from": {"2024-01-01"},
			"date_to":   {"2024-01-31"},
			"tz":        {"America/New_York"},
			"limit":     {"20"},
			"sort_by":   {"revenue"},
		}

		// Act
		req, err := NewMetricsFiltersRequest(params)

		// Assert
		assert.NoError(t, err)
		assert.NotNil(t, req.DateFrom)
		assert.Equal(t, "2024-01-01", *req.DateFrom)
		assert.NotNil(t, req.DateTo)
		assert.Equal(t, "2024-01-31", *req.DateTo)
		assert.NotNil(t, req.Timezone)
		assert.Equal(t, "America/New_York", *req.Timezone)
		assert.Equal(t, 20, req.Limit)
		assert.Equal(t, "revenue", req.SortBy)
	})

	t.Run("when no params then returns defaults", func(t *testing.T) {
		// Arrange
		params := url.Values{}

		// Act
		req, err := NewMetricsFiltersRequest(params)

		// Assert
		assert.NoError(t, err)
		assert.Nil(t, req.DateFrom)
		assert.Nil(t, req.DateTo)
		assert.Nil(t, req.Timezone)
		assert.Equal(t, 0, req.Limit)
		assert.Equal(t, "", req.SortBy)
	})

	t.Run("when limit is not numeric then returns error", func(t *testing.T) {
		// Arrange
		params := url.Values{"limit": {"abc"}}

		// Act
		req, err := NewMetricsFiltersRequest(params)

		// Assert
		assert.Error(t, err)
		assert.Nil(t, req)
		badReqErr, ok := err.(*httpErrors.BadRequestError)
		assert.True(t, ok)
		assert.Equal(t, "invalid_limit_format", badReqErr.Message)
	})

	t.Run("when partial params then only those are set", func(t *testing.T) {
		// Arrange
		params := url.Values{
			"date_from": {"2024-06-01"},
			"sort_by":   {"quantity"},
		}

		// Act
		req, err := NewMetricsFiltersRequest(params)

		// Assert
		assert.NoError(t, err)
		assert.NotNil(t, req.DateFrom)
		assert.Equal(t, "2024-06-01", *req.DateFrom)
		assert.Nil(t, req.DateTo)
		assert.Nil(t, req.Timezone)
		assert.Equal(t, 0, req.Limit)
		assert.Equal(t, "quantity", req.SortBy)
	})
}

// =============================================================================
// MetricsFiltersRequest.Validate
// =============================================================================

func TestMetricsFiltersRequest_Validate(t *testing.T) {
	t.Run("when all fields valid then returns nil", func(t *testing.T) {
		// Arrange
		dateFrom := "2024-01-01"
		dateTo := "2024-01-31"
		req := &MetricsFiltersRequest{
			DateFrom: &dateFrom,
			DateTo:   &dateTo,
			Limit:    10,
		}

		// Act
		err := req.Validate()

		// Assert
		assert.NoError(t, err)
	})

	t.Run("when no dates then returns nil", func(t *testing.T) {
		// Arrange
		req := &MetricsFiltersRequest{}

		// Act
		err := req.Validate()

		// Assert
		assert.NoError(t, err)
	})

	t.Run("when date_from is invalid format then returns error", func(t *testing.T) {
		// Arrange
		dateFrom := "not-a-date"
		req := &MetricsFiltersRequest{DateFrom: &dateFrom}

		// Act
		err := req.Validate()

		// Assert
		assert.Error(t, err)
		badReqErr, ok := err.(*httpErrors.BadRequestError)
		assert.True(t, ok)
		assert.Equal(t, "invalid_date_from_format", badReqErr.Message)
	})

	t.Run("when date_to is invalid format then returns error", func(t *testing.T) {
		// Arrange
		dateTo := "bad-date"
		req := &MetricsFiltersRequest{DateTo: &dateTo}

		// Act
		err := req.Validate()

		// Assert
		assert.Error(t, err)
		badReqErr, ok := err.(*httpErrors.BadRequestError)
		assert.True(t, ok)
		assert.Equal(t, "invalid_date_to_format", badReqErr.Message)
	})

	t.Run("when limit is negative then returns error", func(t *testing.T) {
		// Arrange
		req := &MetricsFiltersRequest{Limit: -5}

		// Act
		err := req.Validate()

		// Assert
		assert.Error(t, err)
		badReqErr, ok := err.(*httpErrors.BadRequestError)
		assert.True(t, ok)
		assert.Equal(t, "limit_cannot_be_negative", badReqErr.Message)
	})

	t.Run("when limit is zero then returns nil", func(t *testing.T) {
		// Arrange
		req := &MetricsFiltersRequest{Limit: 0}

		// Act
		err := req.Validate()

		// Assert
		assert.NoError(t, err)
	})

	t.Run("when timezone is invalid then still validates without error", func(t *testing.T) {
		// Arrange — loadTimezone falls back to UTC, Validate uses UTC for parsing
		dateFrom := "2024-01-01"
		tz := "Invalid/Zone"
		req := &MetricsFiltersRequest{DateFrom: &dateFrom, Timezone: &tz}

		// Act
		err := req.Validate()

		// Assert
		assert.NoError(t, err) // fallback to UTC, no error
	})
}

// =============================================================================
// MetricsFiltersRequest.ToMetricsFilters
// =============================================================================

func TestMetricsFiltersRequest_ToMetricsFilters(t *testing.T) {
	t.Run("when dates with timezone then parses in that timezone", func(t *testing.T) {
		// Arrange
		dateFrom := "2024-06-01"
		dateTo := "2024-06-30"
		tz := "America/New_York"
		req := &MetricsFiltersRequest{
			DateFrom: &dateFrom,
			DateTo:   &dateTo,
			Timezone: &tz,
			Limit:    15,
			SortBy:   "revenue",
		}

		// Act
		filters := req.ToMetricsFilters()

		// Assert
		nyLoc, _ := time.LoadLocation("America/New_York")
		expectedFrom := time.Date(2024, 6, 1, 0, 0, 0, 0, nyLoc)
		assert.Equal(t, expectedFrom, *filters.DateFrom)
		assert.Equal(t, 15, filters.Limit)
		assert.Equal(t, "revenue", filters.SortBy)
		assert.Equal(t, &tz, filters.Timezone)
	})

	t.Run("when date_to provided then adds one day for exclusive upper bound", func(t *testing.T) {
		// Arrange
		dateTo := "2024-06-30"
		req := &MetricsFiltersRequest{DateTo: &dateTo}

		// Act
		filters := req.ToMetricsFilters()

		// Assert — critical: date_to should be July 1st (exclusive upper bound)
		expectedTo := time.Date(2024, 7, 1, 0, 0, 0, 0, time.UTC)
		assert.Equal(t, expectedTo, *filters.DateTo)
	})

	t.Run("when no dates then returns nil date pointers", func(t *testing.T) {
		// Arrange
		req := &MetricsFiltersRequest{}

		// Act
		filters := req.ToMetricsFilters()

		// Assert
		assert.Nil(t, filters.DateFrom)
		assert.Nil(t, filters.DateTo)
	})

	t.Run("when timezone is nil then uses UTC", func(t *testing.T) {
		// Arrange
		dateFrom := "2024-06-01"
		req := &MetricsFiltersRequest{DateFrom: &dateFrom}

		// Act
		filters := req.ToMetricsFilters()

		// Assert
		expectedFrom := time.Date(2024, 6, 1, 0, 0, 0, 0, time.UTC)
		assert.Equal(t, expectedFrom, *filters.DateFrom)
	})

	t.Run("when limit and sortBy set then passes through", func(t *testing.T) {
		// Arrange
		req := &MetricsFiltersRequest{Limit: 25, SortBy: "quantity"}

		// Act
		filters := req.ToMetricsFilters()

		// Assert
		assert.Equal(t, 25, filters.Limit)
		assert.Equal(t, "quantity", filters.SortBy)
	})
}

// =============================================================================
// MetricsFiltersRequest.loadTimezone
// =============================================================================

func TestMetricsFiltersRequest_loadTimezone(t *testing.T) {
	t.Run("when timezone is nil then returns UTC", func(t *testing.T) {
		// Arrange
		req := &MetricsFiltersRequest{Timezone: nil}

		// Act
		loc := req.loadTimezone()

		// Assert
		assert.Equal(t, time.UTC, loc)
	})

	t.Run("when timezone is empty then returns UTC", func(t *testing.T) {
		// Arrange
		tz := ""
		req := &MetricsFiltersRequest{Timezone: &tz}

		// Act
		loc := req.loadTimezone()

		// Assert
		assert.Equal(t, time.UTC, loc)
	})

	t.Run("when timezone is valid then returns location", func(t *testing.T) {
		// Arrange
		tz := "America/New_York"
		req := &MetricsFiltersRequest{Timezone: &tz}

		// Act
		loc := req.loadTimezone()

		// Assert
		expectedLoc, _ := time.LoadLocation("America/New_York")
		assert.Equal(t, expectedLoc, loc)
	})

	t.Run("when timezone is invalid then returns UTC", func(t *testing.T) {
		// Arrange
		tz := "Invalid/Zone"
		req := &MetricsFiltersRequest{Timezone: &tz}

		// Act
		loc := req.loadTimezone()

		// Assert
		assert.Equal(t, time.UTC, loc)
	})
}
