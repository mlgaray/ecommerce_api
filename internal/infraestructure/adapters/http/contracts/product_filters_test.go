package contracts

import (
	"testing"

	"github.com/stretchr/testify/assert"

	httpErrors "github.com/mlgaray/ecommerce_api/internal/infraestructure/adapters/http/errors"
)

// =============================================================================
// NewProductFiltersRequest Tests
// =============================================================================

func TestNewProductFiltersRequest(t *testing.T) {
	t.Run("when no params then returns empty request", func(t *testing.T) {
		// Arrange
		params := map[string][]string{}

		// Act
		result, err := NewProductFiltersRequest(params)

		// Assert
		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Nil(t, result.Search)
		assert.Nil(t, result.CategoryID)
		assert.Equal(t, 0, result.Limit)
	})

	t.Run("when search provided then sets search", func(t *testing.T) {
		// Arrange
		params := map[string][]string{
			"search": {"laptop"},
		}

		// Act
		result, err := NewProductFiltersRequest(params)

		// Assert
		assert.NoError(t, err)
		assert.NotNil(t, result.Search)
		assert.Equal(t, "laptop", *result.Search)
	})

	t.Run("when category_id provided then sets category", func(t *testing.T) {
		// Arrange
		params := map[string][]string{
			"category_id": {"5"},
		}

		// Act
		result, err := NewProductFiltersRequest(params)

		// Assert
		assert.NoError(t, err)
		assert.NotNil(t, result.CategoryID)
		assert.Equal(t, 5, *result.CategoryID)
	})

	t.Run("when invalid category_id then returns error", func(t *testing.T) {
		// Arrange
		params := map[string][]string{
			"category_id": {"abc"},
		}

		// Act
		result, err := NewProductFiltersRequest(params)

		// Assert
		assert.Error(t, err)
		assert.Nil(t, result)
		var badReqErr *httpErrors.BadRequestError
		assert.ErrorAs(t, err, &badReqErr)
		assert.Equal(t, "invalid_category_id_format", badReqErr.Message)
	})

	t.Run("when is_active true provided then sets flag", func(t *testing.T) {
		// Arrange
		params := map[string][]string{
			"is_active": {"true"},
		}

		// Act
		result, err := NewProductFiltersRequest(params)

		// Assert
		assert.NoError(t, err)
		assert.NotNil(t, result.IsActive)
		assert.True(t, *result.IsActive)
	})

	t.Run("when is_active false provided then sets flag", func(t *testing.T) {
		// Arrange
		params := map[string][]string{
			"is_active": {"false"},
		}

		// Act
		result, err := NewProductFiltersRequest(params)

		// Assert
		assert.NoError(t, err)
		assert.NotNil(t, result.IsActive)
		assert.False(t, *result.IsActive)
	})

	t.Run("when invalid is_active then returns error", func(t *testing.T) {
		// Arrange
		params := map[string][]string{
			"is_active": {"maybe"},
		}

		// Act
		result, err := NewProductFiltersRequest(params)

		// Assert
		assert.Error(t, err)
		assert.Nil(t, result)
		var badReqErr *httpErrors.BadRequestError
		assert.ErrorAs(t, err, &badReqErr)
		assert.Equal(t, "invalid_is_active_format", badReqErr.Message)
	})

	t.Run("when is_highlighted provided then sets flag", func(t *testing.T) {
		// Arrange
		params := map[string][]string{
			"is_highlighted": {"true"},
		}

		// Act
		result, err := NewProductFiltersRequest(params)

		// Assert
		assert.NoError(t, err)
		assert.NotNil(t, result.IsHighlighted)
		assert.True(t, *result.IsHighlighted)
	})

	t.Run("when is_promotional provided then sets flag", func(t *testing.T) {
		// Arrange
		params := map[string][]string{
			"is_promotional": {"true"},
		}

		// Act
		result, err := NewProductFiltersRequest(params)

		// Assert
		assert.NoError(t, err)
		assert.NotNil(t, result.IsPromotional)
		assert.True(t, *result.IsPromotional)
	})

	t.Run("when min_price provided then sets price", func(t *testing.T) {
		// Arrange
		params := map[string][]string{
			"min_price": {"100.50"},
		}

		// Act
		result, err := NewProductFiltersRequest(params)

		// Assert
		assert.NoError(t, err)
		assert.NotNil(t, result.MinPrice)
		assert.Equal(t, 100.50, *result.MinPrice)
	})

	t.Run("when invalid min_price then returns error", func(t *testing.T) {
		// Arrange
		params := map[string][]string{
			"min_price": {"abc"},
		}

		// Act
		result, err := NewProductFiltersRequest(params)

		// Assert
		assert.Error(t, err)
		assert.Nil(t, result)
		var badReqErr *httpErrors.BadRequestError
		assert.ErrorAs(t, err, &badReqErr)
		assert.Equal(t, "invalid_min_price_format", badReqErr.Message)
	})

	t.Run("when max_price provided then sets price", func(t *testing.T) {
		// Arrange
		params := map[string][]string{
			"max_price": {"500.99"},
		}

		// Act
		result, err := NewProductFiltersRequest(params)

		// Assert
		assert.NoError(t, err)
		assert.NotNil(t, result.MaxPrice)
		assert.Equal(t, 500.99, *result.MaxPrice)
	})

	t.Run("when invalid max_price then returns error", func(t *testing.T) {
		// Arrange
		params := map[string][]string{
			"max_price": {"xyz"},
		}

		// Act
		result, err := NewProductFiltersRequest(params)

		// Assert
		assert.Error(t, err)
		assert.Nil(t, result)
		var badReqErr *httpErrors.BadRequestError
		assert.ErrorAs(t, err, &badReqErr)
		assert.Equal(t, "invalid_max_price_format", badReqErr.Message)
	})

	t.Run("when limit provided then sets limit", func(t *testing.T) {
		// Arrange
		params := map[string][]string{
			"limit": {"50"},
		}

		// Act
		result, err := NewProductFiltersRequest(params)

		// Assert
		assert.NoError(t, err)
		assert.Equal(t, 50, result.Limit)
	})

	t.Run("when invalid limit then returns error", func(t *testing.T) {
		// Arrange
		params := map[string][]string{
			"limit": {"abc"},
		}

		// Act
		result, err := NewProductFiltersRequest(params)

		// Assert
		assert.Error(t, err)
		assert.Nil(t, result)
		var badReqErr *httpErrors.BadRequestError
		assert.ErrorAs(t, err, &badReqErr)
		assert.Equal(t, "invalid_limit_format", badReqErr.Message)
	})

	t.Run("when sort and order provided then sets them", func(t *testing.T) {
		// Arrange
		params := map[string][]string{
			"sort":  {"price"},
			"order": {"asc"},
		}

		// Act
		result, err := NewProductFiltersRequest(params)

		// Assert
		assert.NoError(t, err)
		assert.Equal(t, "price", result.SortBy)
		assert.Equal(t, "asc", result.SortOrder)
	})

	t.Run("when cursor provided then sets cursor", func(t *testing.T) {
		// Arrange
		params := map[string][]string{
			"cursor": {"eyJsYXN0X2lkIjoxMH0="},
		}

		// Act
		result, err := NewProductFiltersRequest(params)

		// Assert
		assert.NoError(t, err)
		assert.Equal(t, "eyJsYXN0X2lkIjoxMH0=", result.Cursor)
	})

	t.Run("when all params provided then sets all", func(t *testing.T) {
		// Arrange
		params := map[string][]string{
			"search":         {"laptop"},
			"category_id":    {"5"},
			"is_active":      {"true"},
			"is_highlighted": {"false"},
			"is_promotional": {"true"},
			"min_price":      {"100"},
			"max_price":      {"500"},
			"sort":           {"name"},
			"order":          {"desc"},
			"limit":          {"25"},
			"cursor":         {"abc123"},
		}

		// Act
		result, err := NewProductFiltersRequest(params)

		// Assert
		assert.NoError(t, err)
		assert.Equal(t, "laptop", *result.Search)
		assert.Equal(t, 5, *result.CategoryID)
		assert.True(t, *result.IsActive)
		assert.False(t, *result.IsHighlighted)
		assert.True(t, *result.IsPromotional)
		assert.Equal(t, 100.0, *result.MinPrice)
		assert.Equal(t, 500.0, *result.MaxPrice)
		assert.Equal(t, "name", result.SortBy)
		assert.Equal(t, "desc", result.SortOrder)
		assert.Equal(t, 25, result.Limit)
		assert.Equal(t, "abc123", result.Cursor)
	})
}

// =============================================================================
// ProductFiltersRequest.Validate Tests
// =============================================================================

func TestProductFiltersRequest_Validate(t *testing.T) {
	t.Run("when no filters then no error", func(t *testing.T) {
		// Arrange
		request := &ProductFiltersRequest{}

		// Act
		err := request.Validate()

		// Assert
		assert.NoError(t, err)
	})

	t.Run("when search is whitespace only then returns error", func(t *testing.T) {
		// Arrange
		search := "   "
		request := &ProductFiltersRequest{Search: &search}

		// Act
		err := request.Validate()

		// Assert
		assert.Error(t, err)
		var badReqErr *httpErrors.BadRequestError
		assert.ErrorAs(t, err, &badReqErr)
		assert.Equal(t, "search_term_cannot_be_empty", badReqErr.Message)
	})

	t.Run("when search has content then trims and passes", func(t *testing.T) {
		// Arrange
		search := "  laptop  "
		request := &ProductFiltersRequest{Search: &search}

		// Act
		err := request.Validate()

		// Assert
		assert.NoError(t, err)
		assert.Equal(t, "laptop", *request.Search)
	})

	t.Run("when limit is negative then returns error", func(t *testing.T) {
		// Arrange
		request := &ProductFiltersRequest{Limit: -5}

		// Act
		err := request.Validate()

		// Assert
		assert.Error(t, err)
		var badReqErr *httpErrors.BadRequestError
		assert.ErrorAs(t, err, &badReqErr)
		assert.Equal(t, "limit_cannot_be_negative", badReqErr.Message)
	})

	t.Run("when limit is zero then no error", func(t *testing.T) {
		// Arrange
		request := &ProductFiltersRequest{Limit: 0}

		// Act
		err := request.Validate()

		// Assert
		assert.NoError(t, err)
	})

	t.Run("when min_price is negative then returns error", func(t *testing.T) {
		// Arrange
		minPrice := -10.0
		request := &ProductFiltersRequest{MinPrice: &minPrice}

		// Act
		err := request.Validate()

		// Assert
		assert.Error(t, err)
		var badReqErr *httpErrors.BadRequestError
		assert.ErrorAs(t, err, &badReqErr)
		assert.Equal(t, "min_price_cannot_be_negative", badReqErr.Message)
	})

	t.Run("when max_price is negative then returns error", func(t *testing.T) {
		// Arrange
		maxPrice := -10.0
		request := &ProductFiltersRequest{MaxPrice: &maxPrice}

		// Act
		err := request.Validate()

		// Assert
		assert.Error(t, err)
		var badReqErr *httpErrors.BadRequestError
		assert.ErrorAs(t, err, &badReqErr)
		assert.Equal(t, "max_price_cannot_be_negative", badReqErr.Message)
	})

	t.Run("when min_price is zero then no error", func(t *testing.T) {
		// Arrange
		minPrice := 0.0
		request := &ProductFiltersRequest{MinPrice: &minPrice}

		// Act
		err := request.Validate()

		// Assert
		assert.NoError(t, err)
	})

	t.Run("when all valid then no error", func(t *testing.T) {
		// Arrange
		search := "laptop"
		minPrice := 100.0
		maxPrice := 500.0
		request := &ProductFiltersRequest{
			Search:   &search,
			MinPrice: &minPrice,
			MaxPrice: &maxPrice,
			Limit:    20,
		}

		// Act
		err := request.Validate()

		// Assert
		assert.NoError(t, err)
	})
}

// =============================================================================
// ProductFiltersRequest.ToProductFilters Tests
// =============================================================================

func TestProductFiltersRequest_ToProductFilters(t *testing.T) {
	t.Run("when no params then returns defaults", func(t *testing.T) {
		// Arrange
		request := &ProductFiltersRequest{}

		// Act
		result := request.ToProductFilters()

		// Assert
		assert.Nil(t, result.Search)
		assert.Nil(t, result.CategoryID)
		assert.Nil(t, result.LastID)
		assert.Nil(t, result.LastSortValue)
	})

	t.Run("when all params then maps all", func(t *testing.T) {
		// Arrange
		search := "laptop"
		categoryID := 5
		isActive := true
		minPrice := 100.0
		maxPrice := 500.0
		request := &ProductFiltersRequest{
			Search:     &search,
			CategoryID: &categoryID,
			IsActive:   &isActive,
			MinPrice:   &minPrice,
			MaxPrice:   &maxPrice,
			SortBy:     "price",
			SortOrder:  "asc",
			Limit:      25,
		}

		// Act
		result := request.ToProductFilters()

		// Assert
		assert.Equal(t, "laptop", *result.Search)
		assert.Equal(t, 5, *result.CategoryID)
		assert.True(t, *result.IsActive)
		assert.Equal(t, 100.0, *result.MinPrice)
		assert.Equal(t, 500.0, *result.MaxPrice)
		assert.Equal(t, "price", result.SortBy)
		assert.Equal(t, "asc", result.SortOrder)
		assert.Equal(t, 25, result.Limit)
	})

	t.Run("when invalid cursor then treats as first page", func(t *testing.T) {
		// Arrange
		request := &ProductFiltersRequest{
			Cursor: "invalid_cursor_not_base64",
		}

		// Act
		result := request.ToProductFilters()

		// Assert
		assert.Nil(t, result.LastID)
		assert.Nil(t, result.LastSortValue)
	})

	t.Run("when no cursor then LastID is nil", func(t *testing.T) {
		// Arrange
		request := &ProductFiltersRequest{
			Cursor: "",
		}

		// Act
		result := request.ToProductFilters()

		// Assert
		assert.Nil(t, result.LastID)
		assert.Nil(t, result.LastSortValue)
	})
}

// =============================================================================
// getQueryParam Tests
// =============================================================================

func TestGetQueryParam(t *testing.T) {
	t.Run("when key exists then returns value", func(t *testing.T) {
		// Arrange
		params := map[string][]string{
			"key": {"value"},
		}

		// Act
		result := getQueryParam(params, "key")

		// Assert
		assert.Equal(t, "value", result)
	})

	t.Run("when key not exists then returns empty", func(t *testing.T) {
		// Arrange
		params := map[string][]string{}

		// Act
		result := getQueryParam(params, "key")

		// Assert
		assert.Equal(t, "", result)
	})

	t.Run("when key exists with empty array then returns empty", func(t *testing.T) {
		// Arrange
		params := map[string][]string{
			"key": {},
		}

		// Act
		result := getQueryParam(params, "key")

		// Assert
		assert.Equal(t, "", result)
	})

	t.Run("when key has multiple values then returns first", func(t *testing.T) {
		// Arrange
		params := map[string][]string{
			"key": {"first", "second", "third"},
		}

		// Act
		result := getQueryParam(params, "key")

		// Assert
		assert.Equal(t, "first", result)
	})
}
