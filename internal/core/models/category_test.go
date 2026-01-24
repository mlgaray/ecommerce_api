package models

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/mlgaray/ecommerce_api/internal/core/errors"
)

// =============================================================================
// Category.GetID Tests
// =============================================================================

func TestCategory_GetID(t *testing.T) {
	t.Run("when category has ID then returns ID", func(t *testing.T) {
		// Arrange
		category := &Category{ID: 42, Name: "Test Category"}

		// Act
		result := category.GetID()

		// Assert
		assert.Equal(t, 42, result)
	})

	t.Run("when category has zero ID then returns zero", func(t *testing.T) {
		// Arrange
		category := &Category{ID: 0, Name: "Test Category"}

		// Act
		result := category.GetID()

		// Assert
		assert.Equal(t, 0, result)
	})
}

// =============================================================================
// Category.GetSortValue Tests
// =============================================================================

func TestCategory_GetSortValue(t *testing.T) {
	t.Run("when sort by name then returns name", func(t *testing.T) {
		// Arrange
		category := &Category{ID: 1, Name: "Electronics", CreatedAt: time.Now()}

		// Act
		result := category.GetSortValue(CategorySortByName)

		// Assert
		assert.Equal(t, "Electronics", result)
	})

	t.Run("when sort by created_at then returns created_at", func(t *testing.T) {
		// Arrange
		createdAt := time.Date(2024, 1, 15, 10, 0, 0, 0, time.UTC)
		category := &Category{ID: 1, Name: "Electronics", CreatedAt: createdAt}

		// Act
		result := category.GetSortValue(CategorySortByCreatedAt)

		// Assert
		assert.Equal(t, createdAt, result)
	})

	t.Run("when sort by unknown field then returns ID", func(t *testing.T) {
		// Arrange
		category := &Category{ID: 42, Name: "Electronics", CreatedAt: time.Now()}

		// Act
		result := category.GetSortValue("unknown_field")

		// Assert
		assert.Equal(t, 42, result)
	})

	t.Run("when sort by empty string then returns ID", func(t *testing.T) {
		// Arrange
		category := &Category{ID: 42, Name: "Electronics", CreatedAt: time.Now()}

		// Act
		result := category.GetSortValue("")

		// Assert
		assert.Equal(t, 42, result)
	})
}

// =============================================================================
// CategoryFilters.Validated Tests
// =============================================================================

func TestCategoryFilters_Validated(t *testing.T) {
	t.Run("when all defaults then returns normalized filters", func(t *testing.T) {
		// Arrange
		filters := CategoryFilters{}

		// Act
		result, err := filters.Validated()

		// Assert
		assert.NoError(t, err)
		assert.Equal(t, DefaultLimit, result.Limit)
		assert.Equal(t, CategorySortByCreatedAt, result.SortBy)
		assert.Equal(t, SortOrderDesc, result.SortOrder)
	})

	t.Run("when valid filters then preserves values", func(t *testing.T) {
		// Arrange
		search := "test"
		lastID := 10
		filters := CategoryFilters{
			Search:    &search,
			SortBy:    CategorySortByName,
			SortOrder: SortOrderAsc,
			Limit:     50,
			LastID:    &lastID,
		}

		// Act
		result, err := filters.Validated()

		// Assert
		assert.NoError(t, err)
		assert.Equal(t, &search, result.Search)
		assert.Equal(t, CategorySortByName, result.SortBy)
		assert.Equal(t, SortOrderAsc, result.SortOrder)
		assert.Equal(t, 50, result.Limit)
		assert.Equal(t, &lastID, result.LastID)
	})

	t.Run("when original is not modified then maintains immutability", func(t *testing.T) {
		// Arrange
		filters := CategoryFilters{
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
		assert.Equal(t, CategorySortByCreatedAt, result.SortBy)
		assert.Equal(t, SortOrderDesc, result.SortOrder)
		assert.Equal(t, DefaultLimit, result.Limit)
	})
}

// =============================================================================
// CategoryFilters Limit Normalization Tests
// =============================================================================

func TestCategoryFilters_Validated_Limit(t *testing.T) {
	t.Run("when limit is zero then uses default", func(t *testing.T) {
		// Arrange
		filters := CategoryFilters{Limit: 0}

		// Act
		result, err := filters.Validated()

		// Assert
		assert.NoError(t, err)
		assert.Equal(t, DefaultLimit, result.Limit)
	})

	t.Run("when limit is negative then uses default", func(t *testing.T) {
		// Arrange
		filters := CategoryFilters{Limit: -10}

		// Act
		result, err := filters.Validated()

		// Assert
		assert.NoError(t, err)
		assert.Equal(t, DefaultLimit, result.Limit)
	})

	t.Run("when limit exceeds max then uses max", func(t *testing.T) {
		// Arrange
		filters := CategoryFilters{Limit: 200}

		// Act
		result, err := filters.Validated()

		// Assert
		assert.NoError(t, err)
		assert.Equal(t, MaxLimit, result.Limit)
	})

	t.Run("when limit equals max then uses max", func(t *testing.T) {
		// Arrange
		filters := CategoryFilters{Limit: MaxLimit}

		// Act
		result, err := filters.Validated()

		// Assert
		assert.NoError(t, err)
		assert.Equal(t, MaxLimit, result.Limit)
	})

	t.Run("when limit is within bounds then uses provided value", func(t *testing.T) {
		// Arrange
		filters := CategoryFilters{Limit: 50}

		// Act
		result, err := filters.Validated()

		// Assert
		assert.NoError(t, err)
		assert.Equal(t, 50, result.Limit)
	})

	t.Run("when limit is one then uses one", func(t *testing.T) {
		// Arrange
		filters := CategoryFilters{Limit: 1}

		// Act
		result, err := filters.Validated()

		// Assert
		assert.NoError(t, err)
		assert.Equal(t, 1, result.Limit)
	})
}

// =============================================================================
// CategoryFilters Sort Field Validation Tests
// =============================================================================

func TestCategoryFilters_Validated_SortBy(t *testing.T) {
	t.Run("when sort_by is empty then uses default", func(t *testing.T) {
		// Arrange
		filters := CategoryFilters{SortBy: ""}

		// Act
		result, err := filters.Validated()

		// Assert
		assert.NoError(t, err)
		assert.Equal(t, CategorySortByCreatedAt, result.SortBy)
	})

	t.Run("when sort_by is name then uses name", func(t *testing.T) {
		// Arrange
		filters := CategoryFilters{SortBy: CategorySortByName}

		// Act
		result, err := filters.Validated()

		// Assert
		assert.NoError(t, err)
		assert.Equal(t, CategorySortByName, result.SortBy)
	})

	t.Run("when sort_by is created_at then uses created_at", func(t *testing.T) {
		// Arrange
		filters := CategoryFilters{SortBy: CategorySortByCreatedAt}

		// Act
		result, err := filters.Validated()

		// Assert
		assert.NoError(t, err)
		assert.Equal(t, CategorySortByCreatedAt, result.SortBy)
	})

	t.Run("when sort_by is invalid then returns validation error", func(t *testing.T) {
		// Arrange
		filters := CategoryFilters{SortBy: "invalid_field"}

		// Act
		result, err := filters.Validated()

		// Assert
		assert.Error(t, err)
		var validationErr *errors.ValidationError
		assert.ErrorAs(t, err, &validationErr)
		assert.Equal(t, errors.InvalidSortField, validationErr.Message)
		assert.Equal(t, CategoryFilters{}, result)
	})

	t.Run("when sort_by is price then returns validation error", func(t *testing.T) {
		// Arrange - price is valid for products but not categories
		filters := CategoryFilters{SortBy: "price"}

		// Act
		result, err := filters.Validated()

		// Assert
		assert.Error(t, err)
		var validationErr *errors.ValidationError
		assert.ErrorAs(t, err, &validationErr)
		assert.Equal(t, errors.InvalidSortField, validationErr.Message)
		assert.Equal(t, CategoryFilters{}, result)
	})
}

// =============================================================================
// CategoryFilters Sort Order Validation Tests
// =============================================================================

func TestCategoryFilters_Validated_SortOrder(t *testing.T) {
	t.Run("when sort_order is empty then uses default desc", func(t *testing.T) {
		// Arrange
		filters := CategoryFilters{SortOrder: ""}

		// Act
		result, err := filters.Validated()

		// Assert
		assert.NoError(t, err)
		assert.Equal(t, SortOrderDesc, result.SortOrder)
	})

	t.Run("when sort_order is asc then uses asc", func(t *testing.T) {
		// Arrange
		filters := CategoryFilters{SortOrder: SortOrderAsc}

		// Act
		result, err := filters.Validated()

		// Assert
		assert.NoError(t, err)
		assert.Equal(t, SortOrderAsc, result.SortOrder)
	})

	t.Run("when sort_order is desc then uses desc", func(t *testing.T) {
		// Arrange
		filters := CategoryFilters{SortOrder: SortOrderDesc}

		// Act
		result, err := filters.Validated()

		// Assert
		assert.NoError(t, err)
		assert.Equal(t, SortOrderDesc, result.SortOrder)
	})

	t.Run("when sort_order is invalid then returns validation error", func(t *testing.T) {
		// Arrange
		filters := CategoryFilters{SortOrder: "invalid"}

		// Act
		result, err := filters.Validated()

		// Assert
		assert.Error(t, err)
		var validationErr *errors.ValidationError
		assert.ErrorAs(t, err, &validationErr)
		assert.Equal(t, errors.InvalidSortOrder, validationErr.Message)
		assert.Equal(t, CategoryFilters{}, result)
	})

	t.Run("when sort_order is uppercase ASC then returns validation error", func(t *testing.T) {
		// Arrange
		filters := CategoryFilters{SortOrder: "ASC"}

		// Act
		result, err := filters.Validated()

		// Assert
		assert.Error(t, err)
		var validationErr *errors.ValidationError
		assert.ErrorAs(t, err, &validationErr)
		assert.Equal(t, errors.InvalidSortOrder, validationErr.Message)
		assert.Equal(t, CategoryFilters{}, result)
	})
}
