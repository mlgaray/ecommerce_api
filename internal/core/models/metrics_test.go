package models

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

// =============================================================================
// MetricsFilters.Validated - Limit
// =============================================================================

func TestMetricsFilters_Validated(t *testing.T) {
	t.Run("when limit is 0 then defaults to 10", func(t *testing.T) {
		// Arrange
		f := MetricsFilters{Limit: 0}

		// Act
		result := f.Validated()

		// Assert
		assert.Equal(t, 10, result.Limit)
	})

	t.Run("when limit is negative then defaults to 10", func(t *testing.T) {
		// Arrange
		f := MetricsFilters{Limit: -5}

		// Act
		result := f.Validated()

		// Assert
		assert.Equal(t, 10, result.Limit)
	})

	t.Run("when limit exceeds 50 then caps to 50", func(t *testing.T) {
		// Arrange
		f := MetricsFilters{Limit: 51}

		// Act
		result := f.Validated()

		// Assert
		assert.Equal(t, 50, result.Limit)
	})

	t.Run("when limit is 50 then keeps 50", func(t *testing.T) {
		// Arrange
		f := MetricsFilters{Limit: 50}

		// Act
		result := f.Validated()

		// Assert
		assert.Equal(t, 50, result.Limit)
	})

	t.Run("when limit is 1 then keeps 1", func(t *testing.T) {
		// Arrange
		f := MetricsFilters{Limit: 1}

		// Act
		result := f.Validated()

		// Assert
		assert.Equal(t, 1, result.Limit)
	})

	t.Run("when limit is 25 then keeps 25", func(t *testing.T) {
		// Arrange
		f := MetricsFilters{Limit: 25}

		// Act
		result := f.Validated()

		// Assert
		assert.Equal(t, 25, result.Limit)
	})

	// =============================================================================
	// MetricsFilters.Validated - SortBy
	// =============================================================================

	t.Run("when sortBy is empty then defaults to quantity", func(t *testing.T) {
		// Arrange
		f := MetricsFilters{SortBy: ""}

		// Act
		result := f.Validated()

		// Assert
		assert.Equal(t, "quantity", result.SortBy)
	})

	t.Run("when sortBy is revenue then keeps revenue", func(t *testing.T) {
		// Arrange
		f := MetricsFilters{SortBy: "revenue"}

		// Act
		result := f.Validated()

		// Assert
		assert.Equal(t, "revenue", result.SortBy)
	})

	t.Run("when sortBy is invalid then defaults to quantity", func(t *testing.T) {
		// Arrange
		f := MetricsFilters{SortBy: "name"}

		// Act
		result := f.Validated()

		// Assert
		assert.Equal(t, "quantity", result.SortBy)
	})

	// =============================================================================
	// MetricsFilters.Validated - Immutability
	// =============================================================================

	t.Run("when validated then original is not modified", func(t *testing.T) {
		// Arrange
		f := MetricsFilters{Limit: 0, SortBy: "invalid"}

		// Act
		result := f.Validated()

		// Assert
		assert.Equal(t, 0, f.Limit)          // original unchanged
		assert.Equal(t, "invalid", f.SortBy) // original unchanged
		assert.Equal(t, 10, result.Limit)
		assert.Equal(t, "quantity", result.SortBy)
	})

	// =============================================================================
	// MetricsFilters.Validated - Passthrough
	// =============================================================================

	t.Run("when dates and timezone set then passes through unchanged", func(t *testing.T) {
		// Arrange
		dateFrom := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
		dateTo := time.Date(2024, 12, 31, 0, 0, 0, 0, time.UTC)
		tz := "America/New_York"
		f := MetricsFilters{
			DateFrom: &dateFrom,
			DateTo:   &dateTo,
			Timezone: &tz,
			Limit:    10,
			SortBy:   "revenue",
		}

		// Act
		result := f.Validated()

		// Assert
		assert.Equal(t, &dateFrom, result.DateFrom)
		assert.Equal(t, &dateTo, result.DateTo)
		assert.Equal(t, &tz, result.Timezone)
	})
}
