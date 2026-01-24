package models

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/mlgaray/ecommerce_api/internal/core/errors"
)

// =============================================================================
// ProductFilters.Validated Tests
// =============================================================================

func TestProductFilters_Validated(t *testing.T) {
	t.Run("when all defaults then returns normalized filters", func(t *testing.T) {
		// Arrange
		filters := ProductFilters{}

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
		search := "laptop"
		categoryID := 5
		isActive := true
		minPrice := 100.0
		maxPrice := 500.0
		lastID := 10
		filters := ProductFilters{
			Search:     &search,
			CategoryID: &categoryID,
			IsActive:   &isActive,
			MinPrice:   &minPrice,
			MaxPrice:   &maxPrice,
			SortBy:     SortByPrice,
			SortOrder:  SortOrderAsc,
			Limit:      50,
			LastID:     &lastID,
		}

		// Act
		result, err := filters.Validated()

		// Assert
		assert.NoError(t, err)
		assert.Equal(t, &search, result.Search)
		assert.Equal(t, &categoryID, result.CategoryID)
		assert.Equal(t, &isActive, result.IsActive)
		assert.Equal(t, &minPrice, result.MinPrice)
		assert.Equal(t, &maxPrice, result.MaxPrice)
		assert.Equal(t, SortByPrice, result.SortBy)
		assert.Equal(t, SortOrderAsc, result.SortOrder)
		assert.Equal(t, 50, result.Limit)
		assert.Equal(t, &lastID, result.LastID)
	})

	t.Run("when original is not modified then maintains immutability", func(t *testing.T) {
		// Arrange
		filters := ProductFilters{
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
// ProductFilters Limit Normalization Tests
// =============================================================================

func TestProductFilters_Validated_Limit(t *testing.T) {
	t.Run("when limit is zero then uses default", func(t *testing.T) {
		// Arrange
		filters := ProductFilters{Limit: 0}

		// Act
		result, err := filters.Validated()

		// Assert
		assert.NoError(t, err)
		assert.Equal(t, DefaultLimit, result.Limit)
	})

	t.Run("when limit is negative then uses default", func(t *testing.T) {
		// Arrange
		filters := ProductFilters{Limit: -10}

		// Act
		result, err := filters.Validated()

		// Assert
		assert.NoError(t, err)
		assert.Equal(t, DefaultLimit, result.Limit)
	})

	t.Run("when limit exceeds max then uses max", func(t *testing.T) {
		// Arrange
		filters := ProductFilters{Limit: 200}

		// Act
		result, err := filters.Validated()

		// Assert
		assert.NoError(t, err)
		assert.Equal(t, MaxLimit, result.Limit)
	})

	t.Run("when limit equals max then uses max", func(t *testing.T) {
		// Arrange
		filters := ProductFilters{Limit: MaxLimit}

		// Act
		result, err := filters.Validated()

		// Assert
		assert.NoError(t, err)
		assert.Equal(t, MaxLimit, result.Limit)
	})

	t.Run("when limit is within bounds then uses provided value", func(t *testing.T) {
		// Arrange
		filters := ProductFilters{Limit: 50}

		// Act
		result, err := filters.Validated()

		// Assert
		assert.NoError(t, err)
		assert.Equal(t, 50, result.Limit)
	})

	t.Run("when limit is one then uses one", func(t *testing.T) {
		// Arrange
		filters := ProductFilters{Limit: 1}

		// Act
		result, err := filters.Validated()

		// Assert
		assert.NoError(t, err)
		assert.Equal(t, 1, result.Limit)
	})
}

// =============================================================================
// ProductFilters Price Range Validation Tests
// =============================================================================

func TestProductFilters_Validated_PriceRange(t *testing.T) {
	t.Run("when min_price is nil then no error", func(t *testing.T) {
		// Arrange
		filters := ProductFilters{MinPrice: nil}

		// Act
		result, err := filters.Validated()

		// Assert
		assert.NoError(t, err)
		assert.Nil(t, result.MinPrice)
	})

	t.Run("when max_price is nil then no error", func(t *testing.T) {
		// Arrange
		filters := ProductFilters{MaxPrice: nil}

		// Act
		result, err := filters.Validated()

		// Assert
		assert.NoError(t, err)
		assert.Nil(t, result.MaxPrice)
	})

	t.Run("when min_price is zero then no error", func(t *testing.T) {
		// Arrange
		minPrice := 0.0
		filters := ProductFilters{MinPrice: &minPrice}

		// Act
		result, err := filters.Validated()

		// Assert
		assert.NoError(t, err)
		assert.Equal(t, 0.0, *result.MinPrice)
	})

	t.Run("when max_price is zero then no error", func(t *testing.T) {
		// Arrange
		maxPrice := 0.0
		filters := ProductFilters{MaxPrice: &maxPrice}

		// Act
		result, err := filters.Validated()

		// Assert
		assert.NoError(t, err)
		assert.Equal(t, 0.0, *result.MaxPrice)
	})

	t.Run("when min_price is positive then no error", func(t *testing.T) {
		// Arrange
		minPrice := 100.50
		filters := ProductFilters{MinPrice: &minPrice}

		// Act
		result, err := filters.Validated()

		// Assert
		assert.NoError(t, err)
		assert.Equal(t, 100.50, *result.MinPrice)
	})

	t.Run("when max_price is positive then no error", func(t *testing.T) {
		// Arrange
		maxPrice := 500.99
		filters := ProductFilters{MaxPrice: &maxPrice}

		// Act
		result, err := filters.Validated()

		// Assert
		assert.NoError(t, err)
		assert.Equal(t, 500.99, *result.MaxPrice)
	})

	t.Run("when min_price is negative then returns validation error", func(t *testing.T) {
		// Arrange
		minPrice := -10.0
		filters := ProductFilters{MinPrice: &minPrice}

		// Act
		result, err := filters.Validated()

		// Assert
		assert.Error(t, err)
		var validationErr *errors.ValidationError
		assert.ErrorAs(t, err, &validationErr)
		assert.Equal(t, errors.MinPriceCannotBeNegative, validationErr.Message)
		assert.Equal(t, ProductFilters{}, result)
	})

	t.Run("when max_price is negative then returns validation error", func(t *testing.T) {
		// Arrange
		maxPrice := -10.0
		filters := ProductFilters{MaxPrice: &maxPrice}

		// Act
		result, err := filters.Validated()

		// Assert
		assert.Error(t, err)
		var validationErr *errors.ValidationError
		assert.ErrorAs(t, err, &validationErr)
		assert.Equal(t, errors.MaxPriceCannotBeNegative, validationErr.Message)
		assert.Equal(t, ProductFilters{}, result)
	})

	t.Run("when min_price greater than max_price then returns validation error", func(t *testing.T) {
		// Arrange
		minPrice := 500.0
		maxPrice := 100.0
		filters := ProductFilters{MinPrice: &minPrice, MaxPrice: &maxPrice}

		// Act
		result, err := filters.Validated()

		// Assert
		assert.Error(t, err)
		var validationErr *errors.ValidationError
		assert.ErrorAs(t, err, &validationErr)
		assert.Equal(t, errors.MinPriceCannotBeGreaterThanMaxPrice, validationErr.Message)
		assert.Equal(t, ProductFilters{}, result)
	})

	t.Run("when min_price equals max_price then no error", func(t *testing.T) {
		// Arrange
		price := 100.0
		filters := ProductFilters{MinPrice: &price, MaxPrice: &price}

		// Act
		result, err := filters.Validated()

		// Assert
		assert.NoError(t, err)
		assert.Equal(t, 100.0, *result.MinPrice)
		assert.Equal(t, 100.0, *result.MaxPrice)
	})

	t.Run("when min_price less than max_price then no error", func(t *testing.T) {
		// Arrange
		minPrice := 50.0
		maxPrice := 200.0
		filters := ProductFilters{MinPrice: &minPrice, MaxPrice: &maxPrice}

		// Act
		result, err := filters.Validated()

		// Assert
		assert.NoError(t, err)
		assert.Equal(t, 50.0, *result.MinPrice)
		assert.Equal(t, 200.0, *result.MaxPrice)
	})
}

// =============================================================================
// ProductFilters Sort Field Validation Tests
// =============================================================================

func TestProductFilters_Validated_SortBy(t *testing.T) {
	t.Run("when sort_by is empty then uses default", func(t *testing.T) {
		// Arrange
		filters := ProductFilters{SortBy: ""}

		// Act
		result, err := filters.Validated()

		// Assert
		assert.NoError(t, err)
		assert.Equal(t, SortByCreatedAt, result.SortBy)
	})

	t.Run("when sort_by is price then uses price", func(t *testing.T) {
		// Arrange
		filters := ProductFilters{SortBy: SortByPrice}

		// Act
		result, err := filters.Validated()

		// Assert
		assert.NoError(t, err)
		assert.Equal(t, SortByPrice, result.SortBy)
	})

	t.Run("when sort_by is name then uses name", func(t *testing.T) {
		// Arrange
		filters := ProductFilters{SortBy: SortByName}

		// Act
		result, err := filters.Validated()

		// Assert
		assert.NoError(t, err)
		assert.Equal(t, SortByName, result.SortBy)
	})

	t.Run("when sort_by is created_at then uses created_at", func(t *testing.T) {
		// Arrange
		filters := ProductFilters{SortBy: SortByCreatedAt}

		// Act
		result, err := filters.Validated()

		// Assert
		assert.NoError(t, err)
		assert.Equal(t, SortByCreatedAt, result.SortBy)
	})

	t.Run("when sort_by is invalid then returns validation error", func(t *testing.T) {
		// Arrange
		filters := ProductFilters{SortBy: "invalid_field"}

		// Act
		result, err := filters.Validated()

		// Assert
		assert.Error(t, err)
		var validationErr *errors.ValidationError
		assert.ErrorAs(t, err, &validationErr)
		assert.Equal(t, errors.InvalidSortField, validationErr.Message)
		assert.Equal(t, ProductFilters{}, result)
	})

	t.Run("when sort_by is sql injection attempt then returns validation error", func(t *testing.T) {
		// Arrange
		filters := ProductFilters{SortBy: "name; DROP TABLE products;--"}

		// Act
		result, err := filters.Validated()

		// Assert
		assert.Error(t, err)
		var validationErr *errors.ValidationError
		assert.ErrorAs(t, err, &validationErr)
		assert.Equal(t, errors.InvalidSortField, validationErr.Message)
		assert.Equal(t, ProductFilters{}, result)
	})
}

// =============================================================================
// ProductFilters Sort Order Validation Tests
// =============================================================================

func TestProductFilters_Validated_SortOrder(t *testing.T) {
	t.Run("when sort_order is empty then uses default desc", func(t *testing.T) {
		// Arrange
		filters := ProductFilters{SortOrder: ""}

		// Act
		result, err := filters.Validated()

		// Assert
		assert.NoError(t, err)
		assert.Equal(t, SortOrderDesc, result.SortOrder)
	})

	t.Run("when sort_order is asc then uses asc", func(t *testing.T) {
		// Arrange
		filters := ProductFilters{SortOrder: SortOrderAsc}

		// Act
		result, err := filters.Validated()

		// Assert
		assert.NoError(t, err)
		assert.Equal(t, SortOrderAsc, result.SortOrder)
	})

	t.Run("when sort_order is desc then uses desc", func(t *testing.T) {
		// Arrange
		filters := ProductFilters{SortOrder: SortOrderDesc}

		// Act
		result, err := filters.Validated()

		// Assert
		assert.NoError(t, err)
		assert.Equal(t, SortOrderDesc, result.SortOrder)
	})

	t.Run("when sort_order is invalid then returns validation error", func(t *testing.T) {
		// Arrange
		filters := ProductFilters{SortOrder: "invalid"}

		// Act
		result, err := filters.Validated()

		// Assert
		assert.Error(t, err)
		var validationErr *errors.ValidationError
		assert.ErrorAs(t, err, &validationErr)
		assert.Equal(t, errors.InvalidSortOrder, validationErr.Message)
		assert.Equal(t, ProductFilters{}, result)
	})

	t.Run("when sort_order is uppercase ASC then returns validation error", func(t *testing.T) {
		// Arrange
		filters := ProductFilters{SortOrder: "ASC"}

		// Act
		result, err := filters.Validated()

		// Assert
		assert.Error(t, err)
		var validationErr *errors.ValidationError
		assert.ErrorAs(t, err, &validationErr)
		assert.Equal(t, errors.InvalidSortOrder, validationErr.Message)
		assert.Equal(t, ProductFilters{}, result)
	})

	t.Run("when sort_order is uppercase DESC then returns validation error", func(t *testing.T) {
		// Arrange
		filters := ProductFilters{SortOrder: "DESC"}

		// Act
		result, err := filters.Validated()

		// Assert
		assert.Error(t, err)
		var validationErr *errors.ValidationError
		assert.ErrorAs(t, err, &validationErr)
		assert.Equal(t, errors.InvalidSortOrder, validationErr.Message)
		assert.Equal(t, ProductFilters{}, result)
	})
}

// =============================================================================
// ProductFilters Boolean Filters Tests
// =============================================================================

func TestProductFilters_Validated_BooleanFilters(t *testing.T) {
	t.Run("when is_active is nil then preserves nil", func(t *testing.T) {
		// Arrange
		filters := ProductFilters{IsActive: nil}

		// Act
		result, err := filters.Validated()

		// Assert
		assert.NoError(t, err)
		assert.Nil(t, result.IsActive)
	})

	t.Run("when is_active is true then preserves true", func(t *testing.T) {
		// Arrange
		isActive := true
		filters := ProductFilters{IsActive: &isActive}

		// Act
		result, err := filters.Validated()

		// Assert
		assert.NoError(t, err)
		assert.NotNil(t, result.IsActive)
		assert.True(t, *result.IsActive)
	})

	t.Run("when is_active is false then preserves false", func(t *testing.T) {
		// Arrange
		isActive := false
		filters := ProductFilters{IsActive: &isActive}

		// Act
		result, err := filters.Validated()

		// Assert
		assert.NoError(t, err)
		assert.NotNil(t, result.IsActive)
		assert.False(t, *result.IsActive)
	})

	t.Run("when is_highlighted is true then preserves true", func(t *testing.T) {
		// Arrange
		isHighlighted := true
		filters := ProductFilters{IsHighlighted: &isHighlighted}

		// Act
		result, err := filters.Validated()

		// Assert
		assert.NoError(t, err)
		assert.NotNil(t, result.IsHighlighted)
		assert.True(t, *result.IsHighlighted)
	})

	t.Run("when is_promotional is true then preserves true", func(t *testing.T) {
		// Arrange
		isPromotional := true
		filters := ProductFilters{IsPromotional: &isPromotional}

		// Act
		result, err := filters.Validated()

		// Assert
		assert.NoError(t, err)
		assert.NotNil(t, result.IsPromotional)
		assert.True(t, *result.IsPromotional)
	})

	t.Run("when all boolean filters are set then preserves all", func(t *testing.T) {
		// Arrange
		isActive := true
		isHighlighted := false
		isPromotional := true
		filters := ProductFilters{
			IsActive:      &isActive,
			IsHighlighted: &isHighlighted,
			IsPromotional: &isPromotional,
		}

		// Act
		result, err := filters.Validated()

		// Assert
		assert.NoError(t, err)
		assert.True(t, *result.IsActive)
		assert.False(t, *result.IsHighlighted)
		assert.True(t, *result.IsPromotional)
	})
}

// =============================================================================
// ProductFilters Category Filter Tests
// =============================================================================

func TestProductFilters_Validated_CategoryID(t *testing.T) {
	t.Run("when category_id is nil then preserves nil", func(t *testing.T) {
		// Arrange
		filters := ProductFilters{CategoryID: nil}

		// Act
		result, err := filters.Validated()

		// Assert
		assert.NoError(t, err)
		assert.Nil(t, result.CategoryID)
	})

	t.Run("when category_id is positive then preserves value", func(t *testing.T) {
		// Arrange
		categoryID := 42
		filters := ProductFilters{CategoryID: &categoryID}

		// Act
		result, err := filters.Validated()

		// Assert
		assert.NoError(t, err)
		assert.Equal(t, 42, *result.CategoryID)
	})

	t.Run("when category_id is zero then preserves zero", func(t *testing.T) {
		// Arrange
		categoryID := 0
		filters := ProductFilters{CategoryID: &categoryID}

		// Act
		result, err := filters.Validated()

		// Assert
		assert.NoError(t, err)
		assert.Equal(t, 0, *result.CategoryID)
	})
}
